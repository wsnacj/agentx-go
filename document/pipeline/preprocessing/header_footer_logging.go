package preprocessing

import (
	"fmt"
	logger "github.com/wsnacj/agentx-go/document/pipeline/internal/logging"
	"strings"
)

// CleanupStats 清理统计信息
type CleanupStats struct {
	Mode                 string
	TotalPages           int
	HeaderCleanedPages   int
	FooterCleanedPages   int
	SampleRemovedHeaders []string // 采样的被删除页眉内容
	SampleRemovedFooters []string // 采样的被删除页脚内容
}

// logDiscoveredPatterns 记录发现的页眉页脚模式
func logDiscoveredPatterns(method string, headerPatterns, footerPatterns []HeaderFooterPattern) {
	logger.Info("[HeaderFooter] %s 识别结果:", method)

	if len(headerPatterns) > 0 {
		logger.Info("[HeaderFooter] 发现页眉模式 %d 个:", len(headerPatterns))
		for i, pattern := range headerPatterns {
			logger.Info("[HeaderFooter] 页眉模式 %d (行数: %d):", i+1, len(pattern.Lines))
			for j, line := range pattern.Lines {
				// 限制每行日志长度
				displayLine := line
				if len(displayLine) > 80 {
					displayLine = displayLine[:77] + "..."
				}
				logger.Info("[HeaderFooter]   行%d: %s", j+1, displayLine)
			}
		}
	} else {
		logger.Info("[HeaderFooter] 未发现页眉模式")
	}

	if len(footerPatterns) > 0 {
		logger.Info("[HeaderFooter] 发现页脚模式 %d 个:", len(footerPatterns))
		for i, pattern := range footerPatterns {
			logger.Info("[HeaderFooter] 页脚模式 %d (行数: %d):", i+1, len(pattern.Lines))
			for j, line := range pattern.Lines {
				// 限制每行日志长度
				displayLine := line
				if len(displayLine) > 80 {
					displayLine = displayLine[:77] + "..."
				}
				logger.Info("[HeaderFooter]   行%d: %s", j+1, displayLine)
			}
		}
	} else {
		logger.Info("[HeaderFooter] 未发现页脚模式")
	}
}

// cleanPageByPatternsWithLogging 带日志的页面清理函数
func cleanPageByPatternsWithLogging(page string, headerPatterns, footerPatterns []HeaderFooterPattern, pageIndex int, stats *CleanupStats) string {
	lines := strings.Split(page, "\n")
	originalLineCount := len(lines)

	// 验证并确定要移除的页眉行数
	headerLinesToRemove := 0
	var removedHeaderLines []string
	for _, pattern := range headerPatterns {
		if len(pattern.Lines) == 0 {
			continue
		}
		actualRemove := verifyAndCountHeaderLines(lines, pattern)
		if actualRemove > headerLinesToRemove {
			headerLinesToRemove = actualRemove
			if actualRemove > 0 && actualRemove <= len(lines) {
				removedHeaderLines = lines[:actualRemove]
			}
		}
	}

	// 验证并确定要移除的页脚行数
	footerLinesToRemove := 0
	var removedFooterLines []string
	for _, pattern := range footerPatterns {
		if len(pattern.Lines) == 0 {
			continue
		}
		actualRemove := verifyAndCountFooterLines(lines, pattern)
		if actualRemove > footerLinesToRemove {
			footerLinesToRemove = actualRemove
			if actualRemove > 0 && actualRemove <= len(lines) {
				footerStart := len(lines) - actualRemove
				if footerStart >= 0 {
					removedFooterLines = lines[footerStart:]
				}
			}
		}
	}

	// 更新统计信息
	if headerLinesToRemove > 0 {
		stats.HeaderCleanedPages++
		// 采样策略：每10页记录一次，或前5页
		if len(stats.SampleRemovedHeaders) < 5 || pageIndex%10 == 0 {
			headerSample := strings.Join(removedHeaderLines, " | ")
			if len(headerSample) > 150 {
				headerSample = headerSample[:147] + "..."
			}
			stats.SampleRemovedHeaders = append(stats.SampleRemovedHeaders,
				fmt.Sprintf("页%d: %s", pageIndex+1, headerSample))
		}
	}

	if footerLinesToRemove > 0 {
		stats.FooterCleanedPages++
		// 采样策略：每10页记录一次，或前5页
		if len(stats.SampleRemovedFooters) < 5 || pageIndex%10 == 0 {
			footerSample := strings.Join(removedFooterLines, " | ")
			if len(footerSample) > 150 {
				footerSample = footerSample[:147] + "..."
			}
			stats.SampleRemovedFooters = append(stats.SampleRemovedFooters,
				fmt.Sprintf("页%d: %s", pageIndex+1, footerSample))
		}
	}

	// 详细日志：前3页显示清理详情
	if pageIndex < 3 && (headerLinesToRemove > 0 || footerLinesToRemove > 0) {
		logger.Debug("[HeaderFooter] 页%d 清理详情: 原始%d行, 删除页眉%d行, 删除页脚%d行",
			pageIndex+1, originalLineCount, headerLinesToRemove, footerLinesToRemove)
		if len(removedHeaderLines) > 0 {
			logger.Debug("[HeaderFooter] 页%d 删除页眉: %v", pageIndex+1, removedHeaderLines)
		}
		if len(removedFooterLines) > 0 {
			logger.Debug("[HeaderFooter] 页%d 删除页脚: %v", pageIndex+1, removedFooterLines)
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

// logCleanupStats 输出清理统计信息
func logCleanupStats(stats *CleanupStats, totalPages int) {
	stats.TotalPages = totalPages

	logger.Info("[HeaderFooter] === 清理统计 (模式: %s) ===", stats.Mode)
	logger.Info("[HeaderFooter] 总页数: %d", stats.TotalPages)
	logger.Info("[HeaderFooter] 页眉清理: %d页 (%.1f%%)",
		stats.HeaderCleanedPages, float64(stats.HeaderCleanedPages)*100/float64(stats.TotalPages))
	logger.Info("[HeaderFooter] 页脚清理: %d页 (%.1f%%)",
		stats.FooterCleanedPages, float64(stats.FooterCleanedPages)*100/float64(stats.TotalPages))

	// 显示页眉清理样例
	if len(stats.SampleRemovedHeaders) > 0 {
		logger.Info("[HeaderFooter] 页眉清理样例:")
		for _, sample := range stats.SampleRemovedHeaders {
			logger.Info("[HeaderFooter]   %s", sample)
		}
	}

	// 显示页脚清理样例
	if len(stats.SampleRemovedFooters) > 0 {
		logger.Info("[HeaderFooter] 页脚清理样例:")
		for _, sample := range stats.SampleRemovedFooters {
			logger.Info("[HeaderFooter]   %s", sample)
		}
	}

	logger.Info("[HeaderFooter] === 清理统计结束 ===")
}
