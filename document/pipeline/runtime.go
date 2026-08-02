package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/wsnacj/agentx-go/document/pipeline/configs"
	"github.com/wsnacj/agentx-go/document/pipeline/derive"
	"github.com/wsnacj/agentx-go/document/pipeline/extractors"
	logger "github.com/wsnacj/agentx-go/document/pipeline/internal/logging"
	"github.com/wsnacj/agentx-go/document/pipeline/preprocessing"
	"github.com/wsnacj/agentx-go/document/pipeline/section"
	"github.com/wsnacj/agentx-go/document/pipeline/types"
	u "github.com/wsnacj/agentx-go/document/pipeline/utils"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Options struct {
	ModelName      string
	OutputDir      string // Optional. Empty output dirs are allocated under the system temp directory.
	MaxChunkChars  int    // 默认 12000
	PDFParseMode   *PDFParseMode
	ExtractionMode *DocumentExtractionMode
	ArtifactPolicy ArtifactPolicy
}

type ArtifactPolicy string

const (
	ArtifactPolicyDefault ArtifactPolicy = ""
	ArtifactPolicyFull    ArtifactPolicy = "full"
	ArtifactPolicySummary ArtifactPolicy = "summary"
	ArtifactPolicyNone    ArtifactPolicy = "none"
)

type ParseCachePolicy string

const (
	ParseCachePolicyDefault   ParseCachePolicy = ""
	ParseCachePolicyNone      ParseCachePolicy = "none"
	ParseCachePolicyRead      ParseCachePolicy = "read"
	ParseCachePolicyWrite     ParseCachePolicy = "write"
	ParseCachePolicyReadWrite ParseCachePolicy = "read_write"
)

type ParseCacheOptions struct {
	Policy ParseCachePolicy
	Dir    string
}

type ParseBudget struct {
	TotalTimeout time.Duration
}

type ParseRequest struct {
	DocPath        string
	SpecPath       string
	ModelName      string
	OutputDir      string
	MaxChunkChars  int
	PageLimit      int
	PDFParseMode   *PDFParseMode
	ExtractionMode *DocumentExtractionMode
	ArtifactPolicy ArtifactPolicy
	Cache          ParseCacheOptions
	TraceMode      string
	Budget         ParseBudget
}

// ParseDocument preserves the concise legacy-shaped call on an explicitly
// constructed Runtime.
func (r *Runtime) ParseDocument(docPath, specPath string, opts Options) (*types.DocumentResult, error) {
	return r.Run(context.Background(), ParseRequest{
		DocPath:        docPath,
		SpecPath:       specPath,
		ModelName:      opts.ModelName,
		OutputDir:      opts.OutputDir,
		MaxChunkChars:  opts.MaxChunkChars,
		PDFParseMode:   opts.PDFParseMode,
		ExtractionMode: opts.ExtractionMode,
		ArtifactPolicy: opts.ArtifactPolicy,
	})
}

// Run executes the portable document pipeline with explicit host dependencies.
func (r *Runtime) Run(ctx context.Context, req ParseRequest) (*types.DocumentResult, error) {
	r.observe(ctx, "run", "started", "")
	result, err := r.run(ctx, req)
	if err != nil {
		r.observe(ctx, "run", "failed", err.Error())
		return nil, err
	}
	r.observe(ctx, "run", "completed", "")
	return result, nil
}

func (r *Runtime) run(ctx context.Context, req ParseRequest) (*types.DocumentResult, error) {
	if r == nil {
		return nil, fmt.Errorf("document runtime is nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if req.Budget.TotalTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, req.Budget.TotalTimeout)
		defer cancel()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	docPath := strings.TrimSpace(req.DocPath)
	specPath := strings.TrimSpace(req.SpecPath)
	if docPath == "" {
		return nil, fmt.Errorf("doc path is required")
	}
	if specPath == "" {
		return nil, fmt.Errorf("spec path is required")
	}
	opts := Options{
		ModelName:      strings.TrimSpace(req.ModelName),
		OutputDir:      strings.TrimSpace(req.OutputDir),
		MaxChunkChars:  req.MaxChunkChars,
		PDFParseMode:   req.PDFParseMode,
		ExtractionMode: req.ExtractionMode,
		ArtifactPolicy: req.ArtifactPolicy,
	}
	artifactPolicy, err := normalizeArtifactPolicy(req.ArtifactPolicy)
	if err != nil {
		return nil, err
	}
	cachePolicy, err := normalizeParseCachePolicy(req.Cache.Policy)
	if err != nil {
		return nil, err
	}
	cacheDir := strings.TrimSpace(req.Cache.Dir)
	if cachePolicy != ParseCachePolicyNone && cacheDir == "" {
		return nil, fmt.Errorf("cache dir is required when cache policy is %q", cachePolicy)
	}
	diagnostics := newDocumentDiagnostics()
	totalStart := startTiming("[docparse] ParseDocument 总执行开始")

	if opts.ModelName == "" {
		opts.ModelName = "qwen3-30b-a3b"
	}

	// 阶段1: 加载配置
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	finishStage := startDiagnosticStage(diagnostics, "load_spec")
	stage1Start := startTiming("[docparse] 阶段1: 加载配置")
	spec, err := configs.LoadSpec(specPath)
	if err != nil {
		finishStage("failed", err)
		return nil, err
	}

	// MaxChunkChars优先级：Options传入 > main.yaml配置 > 默认值
	if opts.MaxChunkChars <= 0 {
		opts.MaxChunkChars = spec.Meta.MaxChunkChars
	}

	parseMode := PDFParseSimple // 默认使用简单模式
	if opts.PDFParseMode != nil {
		parseMode = *opts.PDFParseMode
	} else if spec.Meta.PDFParseMode != "" {
		mode, err := parsePDFParseMode(spec.Meta.PDFParseMode)
		if err != nil {
			finishStage("failed", err)
			return nil, err
		}
		parseMode = mode
	}
	extractionMode := DocumentExtractionModeLegacy
	if opts.ExtractionMode != nil {
		mode, err := NormalizeDocumentExtractionMode(string(*opts.ExtractionMode))
		if err != nil {
			finishStage("failed", err)
			return nil, err
		}
		extractionMode = mode
	}
	fingerprint, err := buildParseFingerprint(parseFingerprintInput{
		DocPath:        docPath,
		SpecPath:       firstNonEmptyString(spec.ConfigDir, specPath),
		ModelName:      opts.ModelName,
		PDFParseMode:   parseMode,
		ExtractionMode: extractionMode,
		MaxChunkChars:  opts.MaxChunkChars,
		PageLimit:      req.PageLimit,
	})
	if err != nil {
		finishStage("failed", err)
		return nil, err
	}
	endTiming(stage1Start, "[docparse] 阶段1: 配置加载完成")
	finishStage("completed", nil)

	cacheEntryDir := ""
	if cachePolicy != ParseCachePolicyNone {
		cacheEntryDir = parseCacheEntryDir(cacheDir, fingerprint.CacheKey)
	}
	if parseCachePolicyCanRead(cachePolicy) {
		finishStage = startDiagnosticStage(diagnostics, "cache_lookup")
		cached, hit, err := loadParseCacheResult(cacheEntryDir, fingerprint)
		if err != nil {
			finishStage("failed", err)
			return nil, err
		}
		if hit {
			cached.Cache = parseCacheInfo(cachePolicy, true, false, cacheEntryDir)
			finishStage("completed", nil, "cache_hit")
			endTiming(totalStart, "[docparse] ParseDocument 总执行完成")
			return cached, nil
		}
		finishStage("completed", nil, "cache_miss")
	}

	// 阶段2: 文本提取
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	finishStage = startDiagnosticStage(diagnostics, "extract_text")
	stage2Start := startTiming("[docparse] 阶段2: 文本提取")
	extracted, err := r.loader.Extract(ctx, ExtractRequest{
		Path:           docPath,
		PageLimit:      req.PageLimit,
		PDFParseMode:   parseMode,
		ExtractionMode: extractionMode,
	})
	if err != nil {
		finishStage("failed", err)
		return nil, fmt.Errorf("提取页文本失败: %w", err)
	}
	if extracted == nil {
		err := fmt.Errorf("提取页文本失败: empty extraction result")
		finishStage("failed", err)
		return nil, err
	}
	pages := extracted.Pages
	layout := buildDocumentLayoutFromPDFResponse(extracted.PDFResponse)
	diagnostics.PageCount = len(pages)
	diagnostics.TextQuality = documentTextQuality(pages)
	diagnostics.TextSource = strings.TrimSpace(extracted.TextSource)
	logger.Info("[docparse] 文本页数: %d", len(pages))
	endTiming(stage2Start, "[docparse] 阶段2: 文本提取完成")
	finishStage("completed", nil)

	// 阶段2.5: 页眉页脚清理预处理
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	finishStage = startDiagnosticStage(diagnostics, "cleanup_header_footer")
	stage2_5Start := startTiming("[docparse] 阶段2.5: 页眉页脚清理")
	cleanupMode := preprocessing.ParseCleanupMode(spec.Meta.HeaderFooterCleanup)
	if cleanupMode != preprocessing.CleanupNone {
		cleanedPages, err := preprocessing.RemoveHeaderFooter(ctx, pages, cleanupMode, spec, opts.ModelName, func(callCtx context.Context, modelName, prompt string, chunks []string) (string, error) {
			return r.complete(callCtx, modelName, prompt, chunks, RetryOptions{})
		})
		if err != nil {
			logger.Warn("[docparse] 页眉页脚清理失败: %v，继续使用原始页面", err)
			diagnostics.Warnings = append(diagnostics.Warnings, "header_footer_cleanup_failed")
			diagnostics.Fallbacks = append(diagnostics.Fallbacks, "header_footer_cleanup_original_pages")
			finishStage("failed", err, "fallback_original_pages")
		} else {
			pages = cleanedPages
			logger.Info("[docparse] 页眉页脚清理完成，模式: %s", cleanupMode)
			finishStage("completed", nil)
		}
	} else {
		logger.Info("[docparse] 跳过页眉页脚清理")
		finishStage("skipped", nil)
	}
	endTiming(stage2_5Start, "[docparse] 阶段2.5: 页眉页脚清理完成")

	// 阶段3: 章节切分
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	finishStage = startDiagnosticStage(diagnostics, "split_sections")
	stage3Start := startTiming("[docparse] 阶段3: 章节切分")
	sectionsConfigPath := filepath.Join(spec.ConfigDir, "sections.yaml")
	nodes, err := r.sectioner.Split(ctx, SectionRequest{ConfigPath: sectionsConfigPath, Pages: pages})
	if err != nil {
		finishStage("failed", err)
		return nil, fmt.Errorf("章节切分失败: %w", err)
	}
	flat := u.Flatten(nodes)
	logger.Info("[docparse] 切分节点: %d", len(flat))
	endTiming(stage3Start, "[docparse] 阶段3: 章节切分完成")
	finishStage("completed", nil)

	// 阶段4: 章节匹配
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	finishStage = startDiagnosticStage(diagnostics, "match_sections")
	stage4Start := startTiming("[docparse] 阶段4: 章节匹配")
	chapTexts := map[string]string{}
	keySet := map[string]struct{}{}
	for _, ch := range spec.Chapters {
		keySet[ch.Key] = struct{}{}
		setSectionDiagnostic(diagnostics, types.SectionDiagnostic{
			Key:           ch.Key,
			Status:        "missing",
			MissingReason: "section_not_found",
		})
	}

	// 精确名称匹配：可能出现同一 key 多次命中，这里按"页数最多优先；若相同则取最早出现"的策略去重
	type cand struct {
		node  *section.Node
		order int
	}
	candidates := map[string][]cand{}
	for i, n := range flat {
		if _, ok := keySet[n.Name]; ok {
			candidates[n.Name] = append(candidates[n.Name], cand{node: n, order: i})
		}
	}
	if len(candidates) > 0 {
		picked := []string{}
		for key, arr := range candidates {
			best := arr[0]
			bestPages := len(arr[0].node.Pages)
			for _, c := range arr[1:] {
				pc := len(c.node.Pages)
				if pc > bestPages || (pc == bestPages && c.order < best.order) {
					best = c
					bestPages = pc
				}
			}
			text := u.JoinAndClip(best.node.Pages, opts.MaxChunkChars)
			chapTexts[key] = text
			setSectionDiagnostic(diagnostics, types.SectionDiagnostic{
				Key:            key,
				Status:         "matched",
				MatchedBy:      "rule",
				PageCount:      bestPages,
				CandidatePages: pageRefsForSectionPages(pages, best.node.Pages),
				Confidence:     0.9,
			})
			picked = append(picked, fmt.Sprintf("%s(%d页)", key, bestPages))
			if len(arr) > 1 {
				logger.Debug("[docparse] 章节名称去重: key=%s, 候选=%d, 保留=%d页", key, len(arr), bestPages)
			}
		}
		sort.Strings(picked)
		logger.Info("[docparse] 章节名称直接匹配: %s", strings.Join(picked, ", "))
	}
	endTiming(stage4Start, "[docparse] 阶段4: 章节匹配完成")
	finishStage("completed", nil)

	// 阶段5: LLM章节识别
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	finishStage = startDiagnosticStage(diagnostics, "detect_sections_llm")
	stage5Start := startTiming("[docparse] 阶段5: LLM章节识别")
	{
		missing := missingChapters(spec, chapTexts)
		if len(missing) > 0 {
			// 日志：缺失章节
			var missList []string
			for k := range missing {
				missList = append(missList, k)
			}
			sort.Strings(missList)
			logger.Info("[docparse] LLM兜底-缺失章节: %s", strings.Join(missList, ", "))
			// 仅对缺失章节做 LLM 识别，避免重复处理已命中的章节
			// 1) 生成只包含缺失章节的临时 spec
			var specForDetect configs.DocSpec = *spec
			var filteredChaps []configs.ChapterSpec
			missingSet := map[string]struct{}{}
			for k := range missing {
				missingSet[k] = struct{}{}
			}
			for _, ch := range spec.Chapters {
				if _, ok := missingSet[ch.Key]; ok {
					filteredChaps = append(filteredChaps, ch)
				}
			}
			specForDetect.Chapters = filteredChaps

			// 2) 过滤优先级分组，仅保留缺失章节
			var filteredGroups [][]string
			if len(spec.Meta.DetectPriorityGroups) > 0 {
				for _, g := range spec.Meta.DetectPriorityGroups {
					var keep []string
					for _, k := range g {
						if _, ok := missingSet[k]; ok {
							keep = append(keep, k)
						}
					}
					if len(keep) > 0 {
						filteredGroups = append(filteredGroups, keep)
					}
				}
				if len(filteredGroups) > 0 {
					var grpLogs []string
					for i, g := range filteredGroups {
						grpLogs = append(grpLogs, fmt.Sprintf("g%d=%v", i, g))
					}
					logger.Info("[docparse] LLM兜底-分组: %s", strings.Join(grpLogs, ", "))
				} else {
					logger.Info("[docparse] LLM兜底-分组: 无（使用默认全部一组）")
				}
			}

			var pagesMap map[string][]int
			var derr error
			// 使用新方案（仅缺失章节）
			batch := spec.Meta.DetectBatchPages
			if batch <= 0 {
				batch = 100
			}
			lineClip := spec.Meta.DetectBodyClip
			if lineClip <= 0 {
				lineClip = 120
			}
			detOpt := &DetectOptions{
				Groups:        filteredGroups,
				HeaderLines:   spec.Meta.HeaderLines,
				LineClip:      lineClip,
				BatchPages:    batch,
				MaxConcurrent: spec.Meta.DetectMaxConcurrent,
				RetryOptions: &RetryOptions{
					MaxRetries:        1,
					AttemptTimeout:    time.Duration(spec.Meta.AttemptTimeout) * time.Second,
					TotalTimeout:      time.Duration(spec.Meta.TotalTimeout) * time.Second,
					BackoffBase:       1 * time.Second,
					BackoffMultiplier: 1.8,
					BackoffJitter:     0.2,
				},
			}
			pagesMap, derr = r.LLMDetectChaptersPriority(ctx, pages, opts.ModelName, &specForDetect, detOpt)
			// 兜底：若新方案失败且无优先级配置，则回退旧实现（同样仅缺失章节）
			if derr != nil && len(spec.Meta.DetectPriorityGroups) == 0 {
				pagesMap, derr = r.llmDetectChapters(ctx, pages, opts.ModelName, &specForDetect)
			}
			if derr == nil {
				llmMatches := []string{}
				for key := range missing {
					idxs := pagesMap[key]
					if len(idxs) == 0 {
						continue
					}
					buf := u.TakePages(pages, idxs)
					if len(buf) == 0 {
						continue
					}
					chapTexts[key] = u.JoinAndClip(buf, opts.MaxChunkChars)
					setSectionDiagnostic(diagnostics, types.SectionDiagnostic{
						Key:            key,
						Status:         "matched",
						MatchedBy:      "llm",
						PageCount:      len(buf),
						CandidatePages: append([]int{}, idxs...),
						Confidence:     0.6,
					})
					llmMatches = append(llmMatches, fmt.Sprintf("%s(%d页)", key, len(buf)))
				}
				if len(llmMatches) > 0 {
					logger.Info("[docparse] LLM兜底识别匹配: %s", strings.Join(llmMatches, ", "))
				}
				finishStage("completed", nil)
			} else {
				logger.Warn("[docparse] LLM 备选识别失败: %v", derr)
				diagnostics.Warnings = append(diagnostics.Warnings, "section_llm_detect_failed")
				finishStage("failed", derr)
			}
		} else {
			finishStage("skipped", nil)
		}
	}
	endTiming(stage5Start, "[docparse] 阶段5: LLM章节识别完成")

	// 阶段6: 字段抽取
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	finishStage = startDiagnosticStage(diagnostics, "extract_fields")
	stage6Start := startTiming("[docparse] 阶段6: 字段抽取")
	result := &types.DocumentResult{Chapters: map[string]*types.ChapterResult{}, Diagnostics: diagnostics, Fingerprint: fingerprint}
	if cachePolicy != ParseCachePolicyNone {
		result.Cache = parseCacheInfo(cachePolicy, false, false, cacheEntryDir)
	}
	for _, ch := range spec.Chapters {
		text, ok := chapTexts[ch.Key]
		if !ok || strings.TrimSpace(text) == "" {
			for _, f := range ch.Fields {
				if strings.TrimSpace(f.DerivedFormula) != "" {
					continue
				}
				setFieldDiagnostic(diagnostics, types.FieldDiagnostic{
					Chapter:        ch.Key,
					Field:          f.Key,
					Status:         "missing",
					MissingReason:  "section_missing",
					ReviewRequired: f.Required,
				})
			}
			continue
		}
		cres := &types.ChapterResult{Key: ch.Key, TextSize: len([]rune(text)), Fields: map[string]types.FieldResult{}}

		// 先跑非 LLM 的 extractor（regex/script）
		for _, f := range ch.Fields {
			candidates := []types.FieldCandidate{}
			// 遍历 extractor 链
			for _, ex := range f.Extractors {
				typ := strings.ToLower(strings.TrimSpace(ex.Type))
				if typ == "regex" && ex.Pattern != "" {
					for _, rr := range extractors.RunRegexCandidates(extractors.RegexInput{
						Text:        text,
						Scope:       ex.Scope,
						Pattern:     ex.Pattern,
						HeaderLines: spec.Meta.HeaderLines,
						FooterLines: spec.Meta.FooterLines,
					}) {
						rawValue := rr.Value
						val := any(rr.Value)
						normalizedValue := val
						warnings := []string{}
						// 脚本归一
						if f.Normalize != "" {
							if v2, ok := extractors.ScriptProcess("normalize_number", val); ok {
								val = v2
								normalizedValue = v2
							} else {
								warnings = append(warnings, "normalize_number_failed")
							}
						}
						candidates = append(candidates, types.FieldCandidate{
							Chapter:         ch.Key,
							Value:           val,
							RawValue:        rawValue,
							NormalizedValue: normalizedValue,
							Source:          "regex",
							Extractor:       "regex",
							Confidence:      rr.Confidence,
							Score:           rr.Confidence,
							Evidence:        rr.Snippet,
							PageRefs:        pageRefsForEvidence(pages, rr.Snippet),
							LineNumber:      rr.LineNumber,
							Warnings:        warnings,
						})
					}
				}
				if typ == "table" {
					rowLabels := append([]string{}, ex.RowLabels...)
					rowLabels = append(rowLabels, f.Aliases...)
					for _, tr := range extractors.RunTableCandidates(extractors.TableInput{
						Text:          text,
						FieldKey:      f.Key,
						RowLabels:     rowLabels,
						ColumnLabels:  ex.ColumnLabels,
						ValueColumn:   ex.ValueColumn,
						MaxCandidates: ex.MaxCandidates,
					}) {
						rawValue := tr.Value
						val := any(tr.Value)
						normalizedValue := val
						warnings := []string{}
						if f.Normalize != "" {
							if v2, ok := extractors.ScriptProcess("normalize_number", val); ok {
								val = v2
								normalizedValue = v2
							} else {
								warnings = append(warnings, "normalize_number_failed")
							}
						}
						pageRefs := pageRefsForEvidence(pages, tr.Snippet)
						boundingBoxes, tableCells := tableLayoutEvidenceForResult(layout, pageRefs, tr)
						candidates = append(candidates, types.FieldCandidate{
							Chapter:         ch.Key,
							Value:           val,
							RawValue:        rawValue,
							NormalizedValue: normalizedValue,
							Source:          "table",
							Extractor:       "table",
							Confidence:      tr.Confidence,
							Score:           tr.Confidence,
							Evidence:        tr.Snippet,
							Unit:            tr.Unit,
							Period:          tr.Period,
							PageRefs:        pageRefs,
							BoundingBoxes:   boundingBoxes,
							TableCells:      tableCells,
							RowLabel:        tr.RowLabel,
							ColumnLabel:     tr.ColumnLabel,
							UnitSource:      tr.UnitSource,
							LineNumber:      tr.LineNumber,
							Warnings:        warnings,
						})
					}
				}
				if typ == "script" && ex.Script != "" {
					for i := range candidates {
						if v2, ok := extractors.ScriptProcess(ex.Script, candidates[i].Value); ok {
							candidates[i].Value = v2
							candidates[i].NormalizedValue = v2
							candidates[i].Source = u.Ternary(candidates[i].Source == "", "script", candidates[i].Source+"+script")
						} else {
							candidates[i].Warnings = append(candidates[i].Warnings, "script_"+safeName(ex.Script)+"_failed")
						}
					}
				}
			}
			if fr, ok := selectFieldCandidate(f, ch.Key, candidates); ok {
				cres.Fields[f.Key] = fr
				normalizationWarning := ""
				if hasString(fr.Warnings, "normalize_number_failed") {
					normalizationWarning = "normalize_number_failed"
				}
				setFieldDiagnostic(diagnostics, matchedFieldDiagnostic(ch.Key, f.Key, fr, normalizationWarning, nil))
			}
		}

		// 收集需要 LLM 的字段
		needLLM := []configs.FieldSpec{}
		for _, f := range ch.Fields {
			if _, exists := cres.Fields[f.Key]; exists {
				continue
			}
			// 如果字段有 derived 公式，则跳过 LLM 抽取
			if strings.TrimSpace(f.DerivedFormula) != "" {
				continue
			}
			// 若 extractor 链包含 llm，则纳入
			for _, ex := range f.Extractors {
				if strings.ToLower(ex.Type) == "llm" {
					needLLM = append(needLLM, f)
					break
				}
			}
		}
		if len(needLLM) > 0 {
			// 优先使用每章自定义提示，否则从 PromptFile 中读取 extract 段落；若不存在则回退同目录 prompt.md 全文
			prompt := ch.LLMPrompt
			if strings.TrimSpace(prompt) == "" {
				prompt = loadExtractPromptFromConfig(spec, needLLM)
			}
			llmRetryOptions := &RetryOptions{
				MaxRetries:        1,
				AttemptTimeout:    time.Duration(spec.Meta.AttemptTimeout) * time.Second,
				TotalTimeout:      time.Duration(spec.Meta.TotalTimeout) * time.Second,
				BackoffBase:       1 * time.Second,
				BackoffMultiplier: 1.8,
				BackoffJitter:     0.2,
			}
			// 一次性按章抽取
			m, raw, err := r.runChapterLLM(ctx, opts.ModelName, prompt, text, llmRetryOptions)
			cres.RawLLM = raw
			cres.Prompt = prompt
			llmValues, repairPlan, repairOutcome, repairNeeded := buildLLMRepairPlan(m, needLLM, err)
			diagnostics.Warnings = compactStrings(append(diagnostics.Warnings, repairOutcome.DocWarnings...))
			if repairNeeded {
				repairPrompt := buildLLMRepairPrompt(prompt, raw, repairPlan.Reason, repairPlan.Fields)
				repairRetryOptions := *llmRetryOptions
				repairRetryOptions.MaxRetries = 0
				if repairRetryOptions.AttemptTimeout <= 0 || repairRetryOptions.AttemptTimeout > 30*time.Second {
					repairRetryOptions.AttemptTimeout = 30 * time.Second
				}
				if repairRetryOptions.TotalTimeout <= 0 || repairRetryOptions.TotalTimeout > 60*time.Second {
					repairRetryOptions.TotalTimeout = 60 * time.Second
				}
				repairMap, repairRaw, repairErr := r.runChapterLLM(ctx, opts.ModelName, repairPrompt, text, &repairRetryOptions)
				cres.LLMRepairPrompt = repairPrompt
				cres.LLMRepairRaw = repairRaw
				cres.LLMRepairReason = repairPlan.Reason
				cres.LLMRepairFields = llmRepairFieldKeys(repairPlan.Fields)
				if repairErr != nil {
					markLLMRepairFailure(repairPlan, &repairOutcome)
				} else {
					llmValues = applyLLMRepairValues(llmValues, repairMap, repairPlan, &repairOutcome)
				}
				diagnostics.Warnings = compactStrings(append(diagnostics.Warnings, repairOutcome.DocWarnings...))
			}
			if err != nil && !repairNeeded {
				logger.Warn("[docparse] 章节 %s LLM 抽取失败: %v", ch.Key, err)
				for _, f := range needLLM {
					setFieldDiagnostic(diagnostics, types.FieldDiagnostic{
						Chapter:        ch.Key,
						Field:          f.Key,
						Status:         "missing",
						Source:         "llm",
						MissingReason:  "llm_extract_failed",
						ReviewRequired: f.Required,
					})
				}
			} else {
				// 写入缺失字段
				for _, f := range needLLM {
					rawValue, ok := llmValues[f.Key]
					val := rawValue
					warnings := compactStrings(repairOutcome.FieldWarnings[f.Key])
					if ok && f.Normalize == "number" {
						if v2, ok2 := extractors.ScriptProcess("normalize_number", val); ok2 {
							val = v2
						} else {
							warnings = append(warnings, "normalize_number_failed")
						}
					}
					// 无论是否存在，均落一个条目，缺失值为 nil，便于下游诊断与审计
					if !ok {
						cres.Fields[f.Key] = types.FieldResult{Key: f.Key, Chapter: ch.Key, Value: nil, Source: "llm"}
						missingReason := firstNonEmptyString(repairOutcome.MissingReason[f.Key], "llm_field_missing")
						setFieldDiagnostic(diagnostics, types.FieldDiagnostic{
							Chapter:        ch.Key,
							Field:          f.Key,
							Status:         "missing",
							Source:         "llm",
							MissingReason:  missingReason,
							ReviewRequired: f.Required,
							Warnings:       compactStrings(repairOutcome.FieldWarnings[f.Key]),
						})
					} else {
						candidates := []types.FieldCandidate{{
							Chapter:         ch.Key,
							Value:           val,
							RawValue:        rawValue,
							NormalizedValue: val,
							Source:          "llm",
							Extractor:       "llm",
							Confidence:      0.5,
							Score:           0.5,
							Evidence:        llmFieldEvidence(ch.Key, rawValue),
							PageRefs:        pageRefsForChapter(diagnostics, ch.Key, pages, text),
							Warnings:        warnings,
						}}
						fr, selected := selectFieldCandidate(f, ch.Key, candidates)
						if !selected {
							cres.Fields[f.Key] = types.FieldResult{Key: f.Key, Chapter: ch.Key, Value: nil, Source: "llm"}
							missingReason := firstNonEmptyString(repairOutcome.MissingReason[f.Key], "llm_field_missing")
							setFieldDiagnostic(diagnostics, types.FieldDiagnostic{
								Chapter:        ch.Key,
								Field:          f.Key,
								Status:         "missing",
								Source:         "llm",
								MissingReason:  missingReason,
								ReviewRequired: f.Required,
								Warnings:       compactStrings(repairOutcome.FieldWarnings[f.Key]),
							})
							continue
						}
						cres.Fields[f.Key] = fr
						normalizationWarning := ""
						if hasString(fr.Warnings, "normalize_number_failed") {
							normalizationWarning = "normalize_number_failed"
						}
						setFieldDiagnostic(diagnostics, matchedFieldDiagnostic(ch.Key, f.Key, fr, normalizationWarning, repairOutcome.FieldWarnings[f.Key]))
					}
				}
			}
		}
		for _, f := range ch.Fields {
			if strings.TrimSpace(f.DerivedFormula) != "" {
				continue
			}
			if _, exists := diagnostics.Fields[fieldDiagnosticID(ch.Key, f.Key)]; exists {
				continue
			}
			setFieldDiagnostic(diagnostics, types.FieldDiagnostic{
				Chapter:        ch.Key,
				Field:          f.Key,
				Status:         "missing",
				MissingReason:  "extractor_no_match",
				ReviewRequired: f.Required,
			})
		}

		result.Chapters[ch.Key] = cres
		result.ChapterOrder = append(result.ChapterOrder, ch.Key)
	}
	result.ChapterOrder = u.UniqueKeepOrder(result.ChapterOrder)
	endTiming(stage6Start, "[docparse] 阶段6: 字段抽取完成")
	finishStage("completed", nil)

	// 阶段7: 派生字段计算
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	finishStage = startDiagnosticStage(diagnostics, "derive_fields")
	stage7Start := startTiming("[docparse] 阶段7: 派生字段计算")
	derive.EvaluateDerived(spec, result)
	endTiming(stage7Start, "[docparse] 阶段7: 派生字段计算完成")
	if len(result.DerivedDiagnostics) > 0 {
		finishStage("completed", nil, "derived_diagnostics_present")
	} else {
		finishStage("completed", nil)
	}

	// 阶段8: 全局校验
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	finishStage = startDiagnosticStage(diagnostics, "validate")
	stage8Start := startTiming("[docparse] 阶段8: 全局校验")
	if len(spec.Validations) > 0 {
		result.Validations = runValidations(spec, result)
	}
	endTiming(stage8Start, "[docparse] 阶段8: 全局校验完成")
	finishStage("completed", nil)

	// 阶段9: 保存工件
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	finishStage = startDiagnosticStage(diagnostics, "save_artifacts")
	stage9Start := startTiming("[docparse] 阶段9: 保存工件")
	if artifactPolicy == ArtifactPolicyNone {
		finishStage("skipped", nil, "artifact_policy_none")
		finishDocumentDiagnostics(diagnostics)
		if parseCachePolicyCanWrite(cachePolicy) {
			if err := saveParseCacheResult(cacheEntryDir, result); err != nil {
				return nil, err
			}
			result.Cache = parseCacheInfo(cachePolicy, false, true, cacheEntryDir)
		}
		endTiming(stage9Start, "[docparse] 阶段9: 保存工件完成")
		endTiming(totalStart, "[docparse] ParseDocument 总执行完成")
		return result, nil
	}
	outdir, err := resolveArtifactOutputDir(opts.OutputDir, docPath)
	if err != nil {
		finishStage("failed", err)
		return nil, err
	}
	result.OutputDir = outdir
	// Mark parsing diagnostics before writing artifacts so the manifest carries a stable summary.
	finishDocumentDiagnostics(diagnostics)
	if parseCachePolicyCanWrite(cachePolicy) {
		if err := saveParseCacheResult(cacheEntryDir, result); err != nil {
			return nil, err
		}
		result.Cache = parseCacheInfo(cachePolicy, false, true, cacheEntryDir)
	}
	if err := saveArtifacts(docPath, specPath, spec, pages, nodes, result, outdir, artifactPolicy); err != nil {
		logger.Warn("[docparse] 保存输出失败: %v", err)
		diagnostics.Warnings = append(diagnostics.Warnings, "save_artifacts_failed")
		finishStage("failed", err)
	} else {
		finishStage("completed", nil)
	}
	finishDocumentDiagnostics(diagnostics)
	endTiming(stage9Start, "[docparse] 阶段9: 保存工件完成")

	endTiming(totalStart, "[docparse] ParseDocument 总执行完成")
	return result, nil
}

func normalizeParseCachePolicy(policy ParseCachePolicy) (ParseCachePolicy, error) {
	switch ParseCachePolicy(strings.ToLower(strings.TrimSpace(string(policy)))) {
	case ParseCachePolicyDefault:
		return ParseCachePolicyNone, nil
	case ParseCachePolicyNone:
		return ParseCachePolicyNone, nil
	case ParseCachePolicyRead:
		return ParseCachePolicyRead, nil
	case ParseCachePolicyWrite:
		return ParseCachePolicyWrite, nil
	case ParseCachePolicyReadWrite, "read-write", "readwrite":
		return ParseCachePolicyReadWrite, nil
	default:
		return "", fmt.Errorf("unknown cache policy: %s", policy)
	}
}

func normalizeArtifactPolicy(policy ArtifactPolicy) (ArtifactPolicy, error) {
	switch ArtifactPolicy(strings.ToLower(strings.TrimSpace(string(policy)))) {
	case ArtifactPolicyDefault:
		return ArtifactPolicySummary, nil
	case ArtifactPolicyFull:
		return ArtifactPolicyFull, nil
	case ArtifactPolicySummary:
		return ArtifactPolicySummary, nil
	case ArtifactPolicyNone:
		return ArtifactPolicyNone, nil
	default:
		return "", fmt.Errorf("unknown artifact policy: %s", policy)
	}
}

func resolveArtifactOutputDir(outputDir string, docPath string) (string, error) {
	outputDir = strings.TrimSpace(outputDir)
	if outputDir != "" {
		return outputDir, nil
	}
	name := safeName(filepath.Base(docPath))
	if name == "" {
		name = "document"
	}
	return os.MkdirTemp("", fmt.Sprintf("docparse_%s_*", name))
}

func parsePDFParseMode(val string) (PDFParseMode, error) {
	s := strings.ToLower(strings.TrimSpace(val))
	switch s {
	case "", "normal", "default":
		return PDFParseNormal, nil
	case "simple", "fast":
		return PDFParseSimple, nil
	case "ocr", "force_ocr", "force-ocr":
		return PDFParseForceOCR, nil
	default:
		return PDFParseNormal, fmt.Errorf("未知的 pdf_parse_mode: %s", val)
	}
}

// ---- 工具与辅助 ----

// 根据外部 preamble 构造最终提示词，并追加字段定义和输出 JSON 结构
func buildPromptFromPreamble(preamble string, fields []configs.FieldSpec) string {
	var sb strings.Builder
	if strings.TrimSpace(preamble) == "" {
		sb.WriteString("你是信息抽取助手，请从用户提供的章节文本中，按字段定义输出严格的 JSON 对象。\n")
		sb.WriteString("要求：\n- 仅输出JSON\n- 缺失字段 null\n- 数值换算为元且不含单位与逗号\n- 日期 YYYY 或 YYYY-MM-DD\n- 只输出定义字段\n\n")
	} else {
		sb.WriteString(strings.TrimSpace(preamble))
		if !strings.HasSuffix(preamble, "\n\n") {
			sb.WriteString("\n\n")
		}
	}
	sb.WriteString("输出JSON字段：\n{\n")
	for i, f := range fields {
		comma := ","
		if i == len(fields)-1 {
			comma = ""
		}

		// 构建字段注释，保留原有的字段信息但采用注释格式
		var parts []string
		if f.Description != "" {
			parts = append(parts, f.Description)
		}
		if f.Required {
			parts = append(parts, "必填")
		}
		if f.Type != "" {
			parts = append(parts, "类型:"+f.Type)
		}
		if f.Unit != "" {
			parts = append(parts, "单位:"+f.Unit)
		}
		comment := strings.Join(parts, " | ")
		sb.WriteString(fmt.Sprintf("  \"%s\": null%s  // %s\n", f.Key, comma, comment))
	}
	sb.WriteString("}\n")
	return sb.String()
}

// 从固定的prompt.md文件中按extract_chapter key读取抽取前言；
// 若读取失败，则回退到 loadPreamble 并拼装字段定义。
func loadExtractPromptFromConfig(spec *configs.DocSpec, fields []configs.FieldSpec) string {
	if spec != nil && spec.ConfigDir != "" {
		promptFile := filepath.Join(spec.ConfigDir, "prompt.md")
		prompts, err := readPromptsFromTxt(promptFile)
		if err == nil {
			if pre := prompts["extract_chapter"]; strings.TrimSpace(pre) != "" {
				return buildPromptFromPreamble(pre, fields)
			}
		} else {
			logger.Warn("[docparse] read prompt file failed: %v", err)
		}
	}
	return buildPromptFromPreamble("", fields)
}

// 关键词匹配阶段已移除

// LLM 页眉识别章节页码
func (r *Runtime) llmDetectChapters(ctx context.Context, pages []string, modelName string, spec *configs.DocSpec) (map[string][]int, error) {
	header := func(text string) string {
		lines := strings.Split(text, "\n")
		if len(lines) > spec.Meta.HeaderLines {
			lines = lines[:spec.Meta.HeaderLines]
		}
		return strings.Join(lines, "\n")
	}
	var chunks []string
	const batch = 50
	for i := 0; i < len(pages); i += batch {
		end := i + batch
		if end > len(pages) {
			end = len(pages)
		}
		var b strings.Builder
		for p := i; p < end; p++ {
			fmt.Fprintf(&b, "[PAGE %d]\n%s\n\n", p+1, header(pages[p]))
		}
		chunks = append(chunks, b.String())
	}
	var sb strings.Builder
	sb.WriteString("你将看到若干页的页眉文本，请判断下列章节(以key表示)分别出现在哪些页上。只输出JSON，形如 { \"cover\": [1], \"balance_sheet\": [88] }。如果不确定某章是否存在，给空数组。\n")
	sb.WriteString("章节列表(含关键词)：\n[\n")
	for i, ch := range spec.Chapters {
		fmt.Fprintf(&sb, "  {\"key\": \"%s\", \"keywords\": %q}%s\n", ch.Key, ch.TitleKeywords, map[bool]string{true: "", false: ","}[i == len(spec.Chapters)-1])
	}
	sb.WriteString("]\n")
	prompt := sb.String()
	// 为章节识别定制重试参数：尝试90s超时，最多2次，总时长上限3分钟
	opts := &RetryOptions{
		MaxRetries:        1,                                                     // 共尝试2次（首轮+1次重试）
		AttemptTimeout:    time.Duration(spec.Meta.AttemptTimeout) * time.Second, // 单次尝试超时
		TotalTimeout:      time.Duration(spec.Meta.TotalTimeout) * time.Second,   // 总体超时
		BackoffBase:       1 * time.Second,
		BackoffMultiplier: 1.8,
		BackoffJitter:     0.2,
	}
	logger.Debug("[llmDetectChapters][prompt] : %s", prompt)
	logger.LogChunks("[llmDetectChapters][chunks] : ", chunks)
	resp, err := r.complete(ctx, modelName, prompt, chunks, *opts)
	if err != nil {
		return nil, err
	}
	// 宽容解析 map[string][]int
	if m, perr := extractors.ParseLoosePagesMap(resp); perr == nil {
		return m, nil
	} else {
		return nil, fmt.Errorf("无法解析LLM识别结果: %w", perr)
	}
}

func (r *Runtime) runChapterLLM(ctx context.Context, modelName, prompt, chapterText string, opts *RetryOptions) (map[string]any, string, error) {
	retry := RetryOptions{}
	if opts != nil {
		retry = *opts
	}
	resp, err := r.complete(ctx, modelName, prompt, []string{chapterText}, retry)
	if err != nil {
		return nil, resp, err
	}
	if value, parseErr := extractors.ParseLooseJSONObject(resp); parseErr == nil {
		return value, resp, nil
	}
	var value map[string]any
	if jsonErr := json.Unmarshal([]byte(resp), &value); jsonErr == nil {
		return value, resp, nil
	}
	return nil, resp, fmt.Errorf("invalid json returned by llm")
}

func missingChapters(spec *configs.DocSpec, have map[string]string) map[string]struct{} {
	miss := map[string]struct{}{}
	for _, ch := range spec.Chapters {
		if _, ok := have[ch.Key]; !ok {
			miss[ch.Key] = struct{}{}
		}
	}
	return miss
}
