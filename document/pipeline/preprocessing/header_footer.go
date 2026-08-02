package preprocessing

import (
	"context"
	"fmt"
	"github.com/wsnacj/agentx-go/document/pipeline/configs"
	logger "github.com/wsnacj/agentx-go/document/pipeline/internal/logging"
	"math"
	"strings"
)

// CleanupMode 页眉页脚清理模式
type CleanupMode string

const (
	CleanupNone         CleanupMode = "none"         // 不清理
	CleanupProgrammatic CleanupMode = "programmatic" // 程序清理
	CleanupLLM          CleanupMode = "llm"          // LLM清理
	CleanupAuto         CleanupMode = "auto"         // 自动清理(程序+LLM)
)

const (
	minSupportRatio              = 0.6
	consensusSimilarityThreshold = 0.65
	perLineSimilarityThreshold   = 0.6
	verificationMatchRatio       = 0.7
)

// LLMRequestFunc is the narrow model seam used only by LLM-assisted cleanup.
// The host owns retries, credentials, provider selection and network policy.
type LLMRequestFunc func(ctx context.Context, modelName, prompt string, chunks []string) (string, error)

// RemoveHeaderFooter 根据指定模式移除页眉页脚
func RemoveHeaderFooter(ctx context.Context, pages []string, mode CleanupMode, spec *configs.DocSpec, llmModel string, request LLMRequestFunc) ([]string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if len(pages) == 0 {
		return pages, nil
	}

	switch mode {
	case CleanupNone:
		return pages, nil
	case CleanupProgrammatic:
		return programmaticCleanup(pages, spec.Meta.HeaderLines, spec.Meta.FooterLines)
	case CleanupLLM:
		return llmCleanup(ctx, pages, llmModel, 8, request) // 固定8页采样
	case CleanupAuto:
		return autoCleanup(ctx, pages, spec.Meta.HeaderLines, spec.Meta.FooterLines, llmModel, request)
	default:
		return pages, fmt.Errorf("未知的清理模式: %s", mode)
	}
}

// programmaticCleanup 程序化清理：用程序算法判断真实页眉页脚，然后程序清理
func programmaticCleanup(pages []string, maxHeaderLines, maxFooterLines int) ([]string, error) {
	if maxHeaderLines <= 0 && maxFooterLines <= 0 {
		return pages, nil
	}

	// 固定采样10页，用程序算法分析页眉页脚模式
	samplePages := 10
	if len(pages) < samplePages {
		samplePages = len(pages)
	}

	// 程序算法：通过相似度分析判断真实的页眉页脚内容
	headerPatterns, footerPatterns := analyzeHeaderFooterPatterns(pages[:samplePages], maxHeaderLines, maxFooterLines)

	// 记录识别的模式
	logDiscoveredPatterns("程序算法", headerPatterns, footerPatterns)

	// 程序清理：基于识别的模式清理所有页面
	result := make([]string, len(pages))
	cleanupStats := &CleanupStats{Mode: "programmatic"}
	for i, page := range pages {
		result[i] = cleanPageByPatternsWithLogging(page, headerPatterns, footerPatterns, i, cleanupStats)
	}

	// 输出清理统计
	logCleanupStats(cleanupStats, len(pages))

	return result, nil
}

// HeaderFooterPattern 页眉页脚模式
type HeaderFooterPattern struct {
	Lines    []string // 具体的页眉/页脚文本行
	LineNums []int    // 对应的行号位置
}

// analyzeHeaderFooterPatterns 用程序算法分析页眉页脚模式
func analyzeHeaderFooterPatterns(samplePages []string, maxHeaderLines, maxFooterLines int) ([]HeaderFooterPattern, []HeaderFooterPattern) {
	headerPatterns := deriveHeaderPatterns(samplePages, maxHeaderLines)
	footerPatterns := deriveFooterPatterns(samplePages, maxFooterLines)
	return headerPatterns, footerPatterns
}

func deriveHeaderPatterns(samplePages []string, maxHeaderLines int) []HeaderFooterPattern {
	if maxHeaderLines <= 0 || len(samplePages) == 0 {
		return nil
	}

	candidates := make([][]string, 0, len(samplePages))
	for _, page := range samplePages {
		lines := strings.Split(page, "\n")
		end := maxHeaderLines
		if end > len(lines) {
			end = len(lines)
		}
		candidates = append(candidates, lines[:end])
	}

	consensus, ok := buildSequentialConsensus(candidates, maxHeaderLines)
	if !ok {
		return nil
	}

	pattern := HeaderFooterPattern{
		Lines:    consensus,
		LineNums: make([]int, len(consensus)),
	}
	for i := range pattern.LineNums {
		pattern.LineNums[i] = i
	}

	return []HeaderFooterPattern{pattern}
}

func deriveFooterPatterns(samplePages []string, maxFooterLines int) []HeaderFooterPattern {
	if maxFooterLines <= 0 || len(samplePages) == 0 {
		return nil
	}

	candidates := make([][]string, 0, len(samplePages))
	for _, page := range samplePages {
		lines := strings.Split(page, "\n")
		if len(lines) == 0 {
			candidates = append(candidates, nil)
			continue
		}
		limit := maxFooterLines
		if limit > len(lines) {
			limit = len(lines)
		}
		bottom := make([]string, 0, limit)
		for i := 0; i < limit; i++ {
			bottom = append(bottom, lines[len(lines)-1-i])
		}
		candidates = append(candidates, bottom)
	}

	consensusReversed, ok := buildSequentialConsensus(candidates, maxFooterLines)
	if !ok {
		return nil
	}

	for i, j := 0, len(consensusReversed)-1; i < j; i, j = i+1, j-1 {
		consensusReversed[i], consensusReversed[j] = consensusReversed[j], consensusReversed[i]
	}

	pattern := HeaderFooterPattern{
		Lines:    consensusReversed,
		LineNums: make([]int, len(consensusReversed)),
	}
	for i := range pattern.LineNums {
		pattern.LineNums[i] = -(len(consensusReversed) - i)
	}

	return []HeaderFooterPattern{pattern}
}

func buildSequentialConsensus(candidates [][]string, maxLines int) ([]string, bool) {
	if len(candidates) == 0 || maxLines <= 0 {
		return nil, false
	}

	required := int(math.Ceil(float64(len(candidates)) * minSupportRatio))
	if required < 2 {
		required = 2
	}

	consensus := make([]string, 0, maxLines)
	for pos := 0; pos < maxLines; pos++ {
		bucket := make([]string, 0, len(candidates))
		for _, candidate := range candidates {
			if pos < len(candidate) {
				line := strings.TrimSpace(candidate[pos])
				if line != "" {
					bucket = append(bucket, line)
				}
			}
		}

		if len(bucket) < required {
			break
		}

		line, support := selectConsensusLine(bucket)
		if support < required {
			break
		}
		consensus = append(consensus, line)
	}

	if len(consensus) == 0 {
		return nil, false
	}

	return consensus, true
}

func selectConsensusLine(lines []string) (string, int) {
	bestLine := ""
	bestSupport := 0
	bestScore := 0.0

	for i, base := range lines {
		support := 1
		score := 0.0
		for j, other := range lines {
			if i == j {
				continue
			}
			similarity := stringSimilarity(base, other)
			if similarity >= consensusSimilarityThreshold {
				support++
				score += similarity
			}
		}

		avgScore := 0.0
		if support > 1 {
			avgScore = score / float64(support-1)
		}

		if support > bestSupport || (support == bestSupport && avgScore > bestScore) {
			bestSupport = support
			bestScore = avgScore
			bestLine = base
		}
	}

	return bestLine, bestSupport
}

// stringSimilarity 计算两个字符串的相似度 (简化版Jaccard相似度)
func stringSimilarity(s1, s2 string) float64 {
	if s1 == s2 {
		return 1.0
	}

	words1 := strings.Fields(s1)
	words2 := strings.Fields(s2)

	if len(words1) == 0 && len(words2) == 0 {
		return 1.0
	}

	intersection := 0
	wordSet2 := make(map[string]struct{}, len(words2))
	for _, word := range words2 {
		wordSet2[word] = struct{}{}
	}

	for _, word := range words1 {
		if _, ok := wordSet2[word]; ok {
			intersection++
		}
	}

	union := len(words1) + len(words2) - intersection
	if union == 0 {
		return 0.0
	}

	return float64(intersection) / float64(union)
}

// analyzePatternsWithLLM 用LLM分析页眉页脚模式
func analyzePatternsWithLLM(ctx context.Context, samplePages []string, llmModel string, request LLMRequestFunc) ([]HeaderFooterPattern, []HeaderFooterPattern, error) {
	if request == nil {
		return nil, nil, fmt.Errorf("LLM request adapter is required")
	}
	// 构建LLM提示词
	prompt := buildHeaderFooterAnalysisPrompt(samplePages)

	// 调用LLM
	chunks := []string{prompt}
	response, err := request(ctx, llmModel, "分析文档页眉页脚模式", chunks)
	if err != nil {
		return nil, nil, err
	}

	// 解析LLM响应，提取页眉页脚模式
	rawHeaderPatterns, rawFooterPatterns := parseLLMResponse(response)

	headerPatterns := hydrateHeaderPatterns(samplePages, rawHeaderPatterns)
	footerPatterns := hydrateFooterPatterns(samplePages, rawFooterPatterns)

	if len(rawHeaderPatterns) > 0 && len(headerPatterns) == 0 {
		logger.Debug("[HeaderFooter] LLM识别到页眉行数但样本未达成共识，跳过页眉清理")
	}
	if len(rawFooterPatterns) > 0 && len(footerPatterns) == 0 {
		logger.Debug("[HeaderFooter] LLM识别到页脚行数但样本未达成共识，跳过页脚清理")
	}

	return headerPatterns, footerPatterns, nil
}

// cleanPageByPatterns 基于识别的模式清理单个页面，带内容验证避免误删
func cleanPageByPatterns(page string, headerPatterns, footerPatterns []HeaderFooterPattern) string {
	lines := strings.Split(page, "\n")

	// 验证并确定要移除的页眉行数
	headerLinesToRemove := 0
	for _, pattern := range headerPatterns {
		if len(pattern.Lines) == 0 {
			continue
		}
		actualRemove := verifyAndCountHeaderLines(lines, pattern)
		if actualRemove > headerLinesToRemove {
			headerLinesToRemove = actualRemove
		}
	}

	// 验证并确定要移除的页脚行数
	footerLinesToRemove := 0
	for _, pattern := range footerPatterns {
		if len(pattern.Lines) == 0 {
			continue
		}
		actualRemove := verifyAndCountFooterLines(lines, pattern)
		if actualRemove > footerLinesToRemove {
			footerLinesToRemove = actualRemove
		}
	}

	// 应用清理
	start := headerLinesToRemove
	if start > len(lines) {
		start = len(lines)
	}

	end := len(lines) - footerLinesToRemove
	if end < start {
		end = start
	}

	if start >= end {
		return ""
	}

	return strings.Join(lines[start:end], "\n")
}

// verifyAndCountHeaderLines 验证并计算实际要移除的页眉行数
func verifyAndCountHeaderLines(pageLines []string, pattern HeaderFooterPattern) int {
	if len(pattern.Lines) == 0 || len(pageLines) == 0 {
		return 0
	}

	maxCheckLines := len(pattern.Lines)
	if maxCheckLines > len(pageLines) {
		maxCheckLines = len(pageLines)
	}

	// 低成本验证：检查页面开头几行是否与模式匹配
	matchCount := 0

	for i := 0; i < maxCheckLines; i++ {
		pageLine := strings.TrimSpace(pageLines[i])
		patternLine := strings.TrimSpace(pattern.Lines[i])

		// 简单的内容验证策略
		if pageLine == "" && patternLine == "" {
			matchCount++ // 都是空行，算匹配
		} else if pageLine != "" && patternLine != "" {
			similarity := stringSimilarity(pageLine, patternLine)
			if similarity >= perLineSimilarityThreshold {
				matchCount++
			}
		}
	}

	// 如果匹配度足够高，返回要删除的行数
	if float64(matchCount)/float64(maxCheckLines) >= verificationMatchRatio {
		return maxCheckLines
	}

	return 0 // 不匹配，不删除
}

// verifyAndCountFooterLines 验证并计算实际要移除的页脚行数
func verifyAndCountFooterLines(pageLines []string, pattern HeaderFooterPattern) int {
	if len(pattern.Lines) == 0 || len(pageLines) == 0 {
		return 0
	}

	maxCheckLines := len(pattern.Lines)
	if maxCheckLines > len(pageLines) {
		maxCheckLines = len(pageLines)
	}

	// 低成本验证：检查页面末尾几行是否与模式匹配
	matchCount := 0

	// 从页面末尾开始比较
	pageStart := len(pageLines) - maxCheckLines
	if pageStart < 0 {
		pageStart = 0
		maxCheckLines = len(pageLines)
	}

	for i := 0; i < maxCheckLines; i++ {
		pageLine := strings.TrimSpace(pageLines[pageStart+i])
		patternLine := strings.TrimSpace(pattern.Lines[i])

		// 简单的内容验证策略
		if pageLine == "" && patternLine == "" {
			matchCount++ // 都是空行，算匹配
		} else if pageLine != "" && patternLine != "" {
			similarity := stringSimilarity(pageLine, patternLine)
			if similarity >= perLineSimilarityThreshold {
				matchCount++
			}
		}
	}

	// 如果匹配度足够高，返回要删除的行数
	if float64(matchCount)/float64(maxCheckLines) >= verificationMatchRatio {
		return maxCheckLines
	}

	return 0 // 不匹配，不删除
}

// buildHeaderFooterAnalysisPrompt 构建LLM分析提示词
func buildHeaderFooterAnalysisPrompt(samplePages []string) string {
	var prompt strings.Builder
	prompt.WriteString("请分析以下文档页面，识别页眉和页脚的模式。\n\n")

	for i, page := range samplePages {
		prompt.WriteString(fmt.Sprintf("=== 页面 %d ===\n", i+1))
		prompt.WriteString(page)
		prompt.WriteString("\n\n")
	}

	prompt.WriteString("请识别:\n")
	prompt.WriteString("1. 页眉: 每页顶部重复出现的内容\n")
	prompt.WriteString("2. 页脚: 每页底部重复出现的内容\n")
	prompt.WriteString("3. 返回格式: HEADER_LINES:数字,FOOTER_LINES:数字\n")

	return prompt.String()
}

// parseLLMResponse 解析LLM响应，提取页眉页脚模式
func parseLLMResponse(response string) ([]HeaderFooterPattern, []HeaderFooterPattern) {
	var headerPatterns []HeaderFooterPattern
	var footerPatterns []HeaderFooterPattern

	// 简单解析，寻找HEADER_LINES和FOOTER_LINES
	lines := strings.Split(response, "\n")
	var headerLines, footerLines int

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "HEADER_LINES:") {
			fmt.Sscanf(line, "HEADER_LINES:%d", &headerLines)
		} else if strings.HasPrefix(line, "FOOTER_LINES:") {
			fmt.Sscanf(line, "FOOTER_LINES:%d", &footerLines)
		}
	}

	// 基于行数创建简单的模式
	if headerLines > 0 {
		pattern := HeaderFooterPattern{
			LineNums: make([]int, headerLines),
		}
		for i := 0; i < headerLines; i++ {
			pattern.LineNums[i] = i
		}
		headerPatterns = append(headerPatterns, pattern)
	}

	if footerLines > 0 {
		pattern := HeaderFooterPattern{
			LineNums: make([]int, footerLines),
		}
		for i := 0; i < footerLines; i++ {
			pattern.LineNums[i] = -(footerLines - i)
		}
		footerPatterns = append(footerPatterns, pattern)
	}

	return headerPatterns, footerPatterns
}

func hydrateHeaderPatterns(samplePages []string, patterns []HeaderFooterPattern) []HeaderFooterPattern {
	if len(patterns) == 0 {
		return nil
	}

	var result []HeaderFooterPattern
	for _, pattern := range patterns {
		count := len(pattern.LineNums)
		if count <= 0 {
			continue
		}
		derived := deriveHeaderPatterns(samplePages, count)
		if len(derived) > 0 {
			result = append(result, derived...)
		}
	}

	return result
}

func hydrateFooterPatterns(samplePages []string, patterns []HeaderFooterPattern) []HeaderFooterPattern {
	if len(patterns) == 0 {
		return nil
	}

	var result []HeaderFooterPattern
	for _, pattern := range patterns {
		count := len(pattern.LineNums)
		if count <= 0 {
			continue
		}
		derived := deriveFooterPatterns(samplePages, count)
		if len(derived) > 0 {
			result = append(result, derived...)
		}
	}

	return result
}

// llmCleanup LLM智能清理：用LLM判断真实页眉页脚，然后程序清理
func llmCleanup(ctx context.Context, pages []string, llmModel string, samplePages int, request LLMRequestFunc) ([]string, error) {
	if len(pages) == 0 {
		return pages, nil
	}

	// 采样页面数量限制
	if samplePages > len(pages) {
		samplePages = len(pages)
	}

	// LLM算法：调用LLM分析页眉页脚模式
	headerPatterns, footerPatterns, err := analyzePatternsWithLLM(ctx, pages[:samplePages], llmModel, request)
	if err != nil {
		return pages, fmt.Errorf("LLM分析页眉页脚失败: %w", err)
	}

	// 记录识别的模式
	logDiscoveredPatterns("LLM算法", headerPatterns, footerPatterns)

	// 程序清理：基于LLM识别的模式清理所有页面
	result := make([]string, len(pages))
	cleanupStats := &CleanupStats{Mode: "llm"}
	for i, page := range pages {
		result[i] = cleanPageByPatternsWithLogging(page, headerPatterns, footerPatterns, i, cleanupStats)
	}

	// 输出清理统计
	logCleanupStats(cleanupStats, len(pages))

	return result, nil
}

// autoCleanup 自动清理：先程序判断，如果没判断出来再用LLM判断，最后统一清理
func autoCleanup(ctx context.Context, pages []string, maxHeaderLines, maxFooterLines int, llmModel string, request LLMRequestFunc) ([]string, error) {
	if len(pages) == 0 {
		return pages, nil
	}

	// 固定采样10页用于分析
	samplePages := 10
	if len(pages) < samplePages {
		samplePages = len(pages)
	}

	// 第一步：用程序算法判断页眉页脚模式
	headerPatterns, footerPatterns := analyzeHeaderFooterPatterns(pages[:samplePages], maxHeaderLines, maxFooterLines)

	// 检查程序算法是否成功识别出模式
	hasHeaderPattern := len(headerPatterns) > 0
	hasFooterPattern := len(footerPatterns) > 0

	// 如果程序算法没有识别出模式，使用LLM进行补充判断
	if !hasHeaderPattern || !hasFooterPattern {
		llmHeaderPatterns, llmFooterPatterns, err := analyzePatternsWithLLM(ctx, pages[:8], llmModel, request) // 保持既有固定8页采样语义
		if err == nil {
			// 如果程序算法没找到页眉模式，使用LLM的结果
			if !hasHeaderPattern && len(llmHeaderPatterns) > 0 {
				headerPatterns = llmHeaderPatterns
			}
			// 如果程序算法没找到页脚模式，使用LLM的结果
			if !hasFooterPattern && len(llmFooterPatterns) > 0 {
				footerPatterns = llmFooterPatterns
			}
		}
		// 如果LLM也失败了，继续使用程序算法的结果（可能为空）
	}

	// 记录最终确定的模式
	logDiscoveredPatterns("自动算法(最终)", headerPatterns, footerPatterns)

	// 统一使用程序清理：基于最终确定的模式清理所有页面
	result := make([]string, len(pages))
	cleanupStats := &CleanupStats{Mode: "auto"}
	for i, page := range pages {
		result[i] = cleanPageByPatternsWithLogging(page, headerPatterns, footerPatterns, i, cleanupStats)
	}

	// 输出清理统计
	logCleanupStats(cleanupStats, len(pages))

	return result, nil
}

// ParseCleanupMode 解析清理模式字符串
func ParseCleanupMode(mode string) CleanupMode {
	switch strings.ToLower(mode) {
	case "none", "":
		return CleanupNone
	case "programmatic", "program":
		return CleanupProgrammatic
	case "llm":
		return CleanupLLM
	case "auto", "automatic":
		return CleanupAuto
	default:
		return CleanupNone
	}
}
