package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/wsnacj/agentx-go/document/pipeline/configs"
	"github.com/wsnacj/agentx-go/document/pipeline/extractors"
	logger "github.com/wsnacj/agentx-go/document/pipeline/internal/logging"
	"math"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// DetectOptions 控制优先级分组与并发等参数
type DetectOptions struct {
	// Groups 按优先级分组的章节 key 列表，groups[0] 为最高优先级。
	// 若为空，将根据常见年报章节推导一个默认分组，并把未出现的章节置于最后一组。
	Groups [][]string

	// HeaderLines 页眉行数（默认使用 spec.Meta.HeaderLines）
	HeaderLines int
	// LineClip 每行最大字符数（默认 120）
	LineClip int
	// BatchPages 每个批次包含的页数（默认 100）
	BatchPages int
	// MaxConcurrent 最大并发批次数（默认 3）
	MaxConcurrent int

	// RetryOptions LLM 重试参数（为空则使用章节识别的默认：90s/1次重试/3m总时长）
	RetryOptions *RetryOptions
}

// DetectChaptersPriority uses priority groups and bounded parallel batches to
// identify 1-based chapter pages with an explicitly supplied model adapter.
func DetectChaptersPriority(ctx context.Context, model Model, pages []string, modelName string, spec *configs.DocSpec, opt *DetectOptions) (map[string][]int, error) {
	if model == nil {
		return nil, fmt.Errorf("model adapter is required")
	}
	runtime := &Runtime{model: model}
	return runtime.detectChaptersPriority(ctx, pages, modelName, spec, opt)
}

func (r *Runtime) detectChaptersPriority(ctx context.Context, pages []string, modelName string, spec *configs.DocSpec, opt *DetectOptions) (map[string][]int, error) {
	if spec == nil {
		return nil, fmt.Errorf("spec is nil")
	}
	if opt == nil {
		opt = &DetectOptions{}
	}
	if opt.HeaderLines <= 0 {
		opt.HeaderLines = spec.Meta.HeaderLines
		if opt.HeaderLines <= 0 {
			opt.HeaderLines = 6
		}
	}
	if opt.LineClip <= 0 {
		opt.LineClip = 120
	}
	if opt.BatchPages <= 0 {
		opt.BatchPages = 100
	}
	if opt.MaxConcurrent <= 0 {
		opt.MaxConcurrent = 3
	}
	if opt.RetryOptions == nil {
		opt.RetryOptions = &RetryOptions{
			MaxRetries:        1,
			AttemptTimeout:    time.Duration(spec.Meta.AttemptTimeout) * time.Second,
			TotalTimeout:      time.Duration(spec.Meta.TotalTimeout) * time.Second,
			BackoffBase:       1 * time.Second,
			BackoffMultiplier: 1.8,
			BackoffJitter:     0.2,
		}
	}

	groups := normalizedGroups(spec, opt.Groups)

	// 剩余页集合（1-based）
	remaining := make(map[int]struct{}, len(pages))
	for i := 1; i <= len(pages); i++ {
		remaining[i] = struct{}{}
	}

	out := map[string][]int{}

	for gi, keys := range groups {
		// 为当前优先级构建需要处理的页索引切片
		pageList := sortedKeys(remaining)
		if len(pageList) == 0 {
			logger.Info("[docparse] priority group %d skipped: no remaining pages", gi)
			continue
		}

		// 按批次切块
		bins := chunkPages(pageList, opt.BatchPages)
		logger.Info("[docparse] priority group %d: keys=%v, bins=%d", gi, keys, len(bins))

		// 预编译 body 匹配正则（每优先级一批），并计算“页中匹配”数量与裁剪
		bodyREs := compileBodyPatterns(spec.Meta.DetectBodyPatterns)
		bodyMax := spec.Meta.DetectBodyMaxMatches
		if bodyMax <= 0 {
			bodyMax = 2
		}
		bodyClip := spec.Meta.DetectBodyClip
		if bodyClip <= 0 {
			bodyClip = opt.LineClip
		}

		// 并发处理每个 bin
		type binResult struct {
			m   map[string][]int
			err error
			idx int
		}
		// 并发 = min(动态bins数, 上限)；上限使用 opt.MaxConcurrent(>0) 否则默认 8
		dynamic := len(bins)
		concCap := opt.MaxConcurrent
		if concCap <= 0 {
			concCap = 8
		}
		maxConc := dynamic
		if maxConc > concCap {
			maxConc = concCap
		}
		if maxConc <= 0 {
			maxConc = 1
		}
		sem := make(chan struct{}, maxConc)
		var wg sync.WaitGroup
		results := make([]binResult, len(bins))

		for bi, bin := range bins {
			wg.Add(1)
			sem <- struct{}{}
			go func(idx int, binPages []int) {
				defer wg.Done()
				defer func() { <-sem }()
				// 构造 chunks：页首 + 页中匹配
				chunks := []string{renderDetectChunk(pages, binPages, opt.HeaderLines, opt.LineClip, bodyREs, bodyMax, bodyClip)}
				prompt := buildDetectPrompt(spec, keys)

				logger.Debug("[LLMDetectChaptersPriority][prompt] : %s", prompt)
				logger.LogChunksDetailed("[LLMDetectChaptersPriority][chunks]", chunks)
				// 调用 LLM
				resp, err := r.complete(ctx, modelName, prompt, chunks, *opt.RetryOptions)
				if err != nil {
					results[idx] = binResult{nil, err, idx}
					return
				}
				// 宽容解析：支持 map[string][]int，或 ranges [[s,e]]，或字符串 "s-e"
				mp, perr := parseRangesOrPagesMap(resp)
				if perr != nil {
					results[idx] = binResult{nil, fmt.Errorf("parse ranges map failed: %w", perr), idx}
					return
				}
				// 将模型输出的 key 归一化为规范 key（允许使用标题关键词/同义写法）
				mp = canonicalizeKeyPagesMapUsingKeywords(mp, keys, spec)
				// 过滤：仅保留当前 keys 且在 binPages 内的页
				filtered := map[string][]int{}
				allowed := makeSet(keys)
				binSet := makeIntSet(binPages)
				for k, arr := range mp {
					if _, ok := allowed[k]; !ok {
						continue
					}
					var keep []int
					for _, p := range arr {
						if _, ok := binSet[p]; ok {
							keep = append(keep, p)
						}
					}
					if len(keep) > 0 {
						filtered[k] = uniqueInts(keep)
					}
				}
				results[idx] = binResult{filtered, nil, idx}
			}(bi, bin)
		}
		wg.Wait()

		// 汇总本优先级结果
		groupPages := map[string]map[int]struct{}{}
		var firstErr error
		for _, r := range results {
			if r.err != nil {
				// 记录第一个错误，但不立即中断，尽量收集其它 bin 结果
				if firstErr == nil {
					firstErr = r.err
				}
				logger.Warn("[docparse] detect bin %d error: %v", r.idx, r.err)
				continue
			}
			for k, arr := range r.m {
				s := groupPages[k]
				if s == nil {
					s = map[int]struct{}{}
				}
				for _, p := range arr {
					s[p] = struct{}{}
				}
				groupPages[k] = s
			}
		}
		if firstErr != nil && len(groupPages) == 0 {
			// 若本组全部失败则返回错误，否则带着部分结果继续
			return nil, firstErr
		}

		// 将本组结果落入 out，并从 remaining 中剔除
		for k, set := range groupPages {
			if len(set) == 0 {
				continue
			}
			// 排序
			var ps []int
			for p := range set {
				ps = append(ps, p)
			}
			sort.Ints(ps)
			out[k] = append(out[k], ps...)
			for _, p := range ps {
				delete(remaining, p)
			}
		}
	}

	// 去重与排序
	for k, arr := range out {
		out[k] = uniqueSorted(arr)
	}
	// 打印最终归类结果（按 key 排序），便于诊断稳定性与覆盖率
	if len(out) == 0 {
		logger.Info("[docparse] priority-detect: no pages matched")
	} else {
		var keys []string
		for k := range out {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			arr := out[k]
			if len(arr) == 0 {
				logger.Info("[docparse] %s:", k)
				continue
			}
			var b strings.Builder
			for i, p := range arr {
				if i > 0 {
					b.WriteByte(' ')
				}
				b.WriteString(strconv.Itoa(p))
			}
			logger.Info("[docparse] %s: %s", k, b.String())
		}
	}
	return out, nil
}

// ---- helpers ----

// normalizedGroups 优先使用 optGroups（调用方传入），否则使用 spec.Meta.DetectPriorityGroups；
// 如仍为空，则默认把所有章节放入一组，顺序与 spec.Chapters 一致。
// 同时会将未在分组中出现的章节追加到最后一组，避免遗漏。
func normalizedGroups(spec *configs.DocSpec, optGroups [][]string) [][]string {
	var groups [][]string
	if len(optGroups) > 0 {
		groups = cloneGroups(optGroups)
	} else if len(spec.Meta.DetectPriorityGroups) > 0 {
		groups = cloneGroups(spec.Meta.DetectPriorityGroups)
	} else if hasAnyPriority(spec) {
		groups = deriveGroupsFromChapterPriority(spec)
	} else {
		// 默认：所有章节一组（通用逻辑，不耦合具体文档类型）
		var all []string
		for _, ch := range spec.Chapters {
			all = append(all, ch.Key)
		}
		groups = [][]string{all}
	}

	// 仅保留 spec.Chapters 中存在的 key，并保持输入顺序；空组移除
	present := map[string]struct{}{}
	for _, ch := range spec.Chapters {
		present[ch.Key] = struct{}{}
	}
	var pruned [][]string
	for _, g := range groups {
		var keep []string
		for _, k := range g {
			if _, ok := present[k]; ok {
				keep = append(keep, k)
			}
		}
		if len(keep) > 0 {
			pruned = append(pruned, keep)
		}
	}
	groups = pruned

	// 追加未覆盖章节：detect_priority_groups 未包含的 key 统一追加在末组
	listed := make(map[string]struct{})
	for _, g := range groups {
		for _, k := range g {
			listed[k] = struct{}{}
		}
	}
	var rest []string
	for _, ch := range spec.Chapters {
		if _, ok := listed[ch.Key]; !ok {
			rest = append(rest, ch.Key)
		}
	}
	if len(rest) > 0 {
		groups = append(groups, rest)
	}
	return groups
}

func hasAnyPriority(spec *configs.DocSpec) bool {
	if spec == nil {
		return false
	}
	for _, ch := range spec.Chapters {
		if ch.Priority > 0 {
			return true
		}
	}
	return false
}

// deriveGroupsFromChapterPriority: 将 chapters 按 Priority 分组；
// Priority 为正整数，越小优先；未配置或 <=0 的章节归为“最低优先级”的一组。
func deriveGroupsFromChapterPriority(spec *configs.DocSpec) [][]string {
	sentinel := math.MaxInt32
	buckets := map[int][]string{}
	for _, ch := range spec.Chapters {
		p := ch.Priority
		if p <= 0 {
			p = sentinel
		}
		buckets[p] = append(buckets[p], ch.Key)
	}
	var ps []int
	for p := range buckets {
		ps = append(ps, p)
	}
	sort.Ints(ps)
	var groups [][]string
	for _, p := range ps {
		groups = append(groups, buckets[p])
	}
	return groups
}

func cloneGroups(gs [][]string) [][]string {
	out := make([][]string, 0, len(gs))
	for _, g := range gs {
		cp := make([]string, len(g))
		copy(cp, g)
		out = append(out, cp)
	}
	return out
}

func clipLine(s string, n int) string {
	if n <= 0 {
		return s
	}
	rs := []rune(strings.TrimSpace(s))
	if len(rs) <= n {
		return string(rs)
	}
	return string(rs[:n])
}

// renderDetectChunk 构建检测用 chunk：包含页首N行与“页中匹配”摘要（可配置正则），并做去重
func renderDetectChunk(pages []string, idxs []int, headerLines int, lineClip int, bodyREs []*regexp.Regexp, bodyMax int, bodyClip int) string {
	header := func(text string) []string {
		lines := strings.Split(text, "\n")
		if headerLines > 0 && len(lines) > headerLines {
			lines = lines[:headerLines]
		}
		for i := range lines {
			lines[i] = clipLine(lines[i], lineClip)
		}
		return lines
	}
	pageInfo := func(full string, hdr []string) []string {
		if len(bodyREs) == 0 || bodyMax <= 0 {
			return nil
		}
		hdrSet := map[string]struct{}{}
		for _, s := range hdr {
			if t := strings.TrimSpace(s); t != "" {
				hdrSet[t] = struct{}{}
			}
		}
		var matches []string
		remain := bodyMax
		for _, re := range bodyREs {
			if remain <= 0 {
				break
			}
			subs := re.FindAllStringSubmatch(full, -1)
			for _, sm := range subs {
				if remain <= 0 {
					break
				}
				pick := ""
				if len(sm) > 1 && strings.TrimSpace(sm[1]) != "" {
					pick = sm[1]
				} else {
					pick = sm[0]
				}
				pick = clipLine(pick, bodyClip)
				pick = strings.TrimSpace(pick)
				if pick == "" {
					continue
				}
				if _, dup := hdrSet[pick]; dup {
					continue
				}
				dup2 := false
				for _, m := range matches {
					if m == pick {
						dup2 = true
						break
					}
				}
				if dup2 {
					continue
				}
				matches = append(matches, pick)
				remain--
			}
		}
		return matches
	}
	var b strings.Builder
	for _, p := range idxs {
		if p < 1 || p > len(pages) {
			continue
		}
		hdr := header(pages[p-1])
		info := pageInfo(pages[p-1], hdr)
		fmt.Fprintf(&b, "[PAGE %d]\n", p)
		if len(hdr) > 0 {
			fmt.Fprintf(&b, "<header_lines count=\"%d\">\n", len(hdr))
			b.WriteString(strings.Join(hdr, "\n"))
			b.WriteString("\n</header_lines>\n")
		}
		if len(info) > 0 {
			fmt.Fprintf(&b, "<body_matches count=\"%d\">\n", len(info))
			for _, s := range info {
				fmt.Fprintf(&b, "<match>%s</match>\n", s)
			}
			b.WriteString("</body_matches>\n")
		}
		b.WriteString("\n")
	}
	return b.String()
}

func compileBodyPatterns(ps []string) []*regexp.Regexp {
	var out []*regexp.Regexp
	for _, p := range ps {
		t := strings.TrimSpace(p)
		if t == "" {
			continue
		}
		if re, err := regexp.Compile(t); err == nil {
			out = append(out, re)
		} else {
			logger.Warn("[docparse] invalid detect_body_pattern: %s", t)
		}
	}
	return out
}

func buildDetectPrompt(spec *configs.DocSpec, keys []string) string {
	// 优先读取配置化的 prompt（PromptFile + DetectPromptKey）；若不可用则回退到通用模板
	base := readDetectPromptFromConfig(spec)
	var sb strings.Builder
	if strings.TrimSpace(base) != "" {
		logger.Debug("[docparse] buildDetectPrompt: using configured prompt from %s[detect_sections]", filepath.Join(spec.ConfigDir, "prompt.md"))
		sb.WriteString(strings.TrimSpace(base))
		if !strings.HasSuffix(base, "\n\n") {
			sb.WriteString("\n\n")
		}
	} else {
		logger.Debug("[docparse] buildDetectPrompt: using default prompt template")
		// 通用模板（不耦合具体业务词），并强制“单页=整数”输出
		sb.WriteString("你将看到若干页的页眉文本（以 [PAGE n] 标注页号），请判断下列章节(以key表示)分别出现在哪些页上。只输出JSON，形如 { \"cover\": [1], \"balance_sheet\": [88] }。如果不确定某章是否存在，给空数组。\n")
	}

	// 构建章节列表(含关键词)，只包含当前批次的keys
	sb.WriteString("章节列表(含关键词)：\n[\n")
	keySet := makeSet(keys)
	validChapters := 0
	for _, ch := range spec.Chapters {
		if _, ok := keySet[ch.Key]; !ok {
			continue
		}
		if validChapters > 0 {
			sb.WriteString(",\n")
		}
		fmt.Fprintf(&sb, "  {\"key\": \"%s\", \"keywords\": %q}", ch.Key, ch.TitleKeywords)
		validChapters++
	}
	sb.WriteString("\n]\n\n")
	return sb.String()
}

func readDetectPromptFromConfig(spec *configs.DocSpec) string {
	if spec == nil || spec.ConfigDir == "" {
		return ""
	}
	promptFile := filepath.Join(spec.ConfigDir, "prompt.md")
	prompts, err := readPromptsFromTxt(promptFile)
	if err != nil {
		logger.Warn("[docparse] read prompt file failed: %v", err)
		return ""
	}
	return prompts["detect_sections"]
}

func chunkPages(pages []int, batch int) [][]int {
	if batch <= 0 {
		batch = 100
	}
	var out [][]int
	for i := 0; i < len(pages); i += batch {
		j := i + batch
		if j > len(pages) {
			j = len(pages)
		}
		cp := make([]int, j-i)
		copy(cp, pages[i:j])
		out = append(out, cp)
	}
	return out
}

func sortedKeys(m map[int]struct{}) []int {
	ps := make([]int, 0, len(m))
	for k := range m {
		ps = append(ps, k)
	}
	sort.Ints(ps)
	return ps
}

func makeSet(arr []string) map[string]struct{} {
	s := make(map[string]struct{}, len(arr))
	for _, v := range arr {
		s[v] = struct{}{}
	}
	return s
}
func makeIntSet(arr []int) map[int]struct{} {
	s := make(map[int]struct{}, len(arr))
	for _, v := range arr {
		s[v] = struct{}{}
	}
	return s
}

func uniqueInts(in []int) []int {
	s := map[int]struct{}{}
	var out []int
	for _, v := range in {
		if _, ok := s[v]; ok {
			continue
		}
		s[v] = struct{}{}
		out = append(out, v)
	}
	return out
}
func uniqueSorted(in []int) []int {
	out := uniqueInts(in)
	sort.Ints(out)
	return out
}

// parseRangesOrPagesMap 解析 LLM 返回的章节-页码映射
// 支持以下格式：
// 1) map[string][]int
// 2) map[string][][]int  // 形如 [[start,end],[start2,end2]]
// 3) map[string][]string // 其中字符串可为 "s-e" 或 "n"
// 4) 宽容对象（带 code fence / 前后噪音），会尝试 ParseLooseJSONObject
func parseRangesOrPagesMap(s string) (map[string][]int, error) {
	// 先尝试 pages map
	var m1 map[string][]int
	if err := json.Unmarshal([]byte(strings.TrimSpace(s)), &m1); err == nil {
		return normalizePagesMap(m1), nil
	}
	// 尝试宽容对象
	obj, err := extractors.ParseLooseJSONObject(s)
	if err != nil {
		return nil, err
	}
	out := map[string][]int{}
	for k, v := range obj {
		switch tv := v.(type) {
		case []any:
			// 可能是 [int,int] 组成的二维数组，或混合字符串区间
			pages := expandMixedArray(tv)
			if len(pages) > 0 {
				out[k] = uniqueSorted(pages)
			}
		case string:
			// 单个字符串也可能是区间
			pages := expandToken(tv)
			if len(pages) > 0 {
				out[k] = uniqueSorted(pages)
			}
		case json.Number:
			if i, err := tv.Int64(); err == nil {
				out[k] = []int{int(i)}
			}
		}
	}
	return out, nil
}

func normalizePagesMap(m map[string][]int) map[string][]int {
	out := map[string][]int{}
	for k, arr := range m {
		out[k] = uniqueSorted(arr)
	}
	return out
}

// 支持连接符：普通连字符-、~、以及 Unicode 短横线/长横线（–, —）
var rangeRe = regexp.MustCompile(`^\s*(\d+)\s*[-~–—]\s*(\d+)\s*$`)

func expandMixedArray(arr []any) []int {
	var out []int
	for _, e := range arr {
		switch t := e.(type) {
		case []any:
			// 情况A：二元数组，视作区间 [a,b]
			if len(t) == 2 {
				a := toIntMaybe(t[0])
				b := toIntMaybe(t[1])
				if a > 0 && b > 0 {
					out = append(out, expandRange(a, b)...)
					continue
				}
			}
			// 情况B：更长的数组或嵌套数组，整体当作一串页码/区间混合，递归展开
			out = append(out, expandMixedArray(t)...)
		case float64:
			out = append(out, int(t))
		case json.Number:
			if i, err := t.Int64(); err == nil {
				out = append(out, int(i))
			} else if f, err := t.Float64(); err == nil {
				out = append(out, int(f))
			}
		case string:
			out = append(out, expandToken(t)...)
		}
	}
	return out
}

func expandToken(tok string) []int {
	tok = strings.TrimSpace(tok)
	if tok == "" {
		return nil
	}
	// 支持逗号分隔的多 token，如 "24, 25, 26" 或 "24-30, 32"
	if strings.Contains(tok, ",") {
		var out []int
		parts := strings.Split(tok, ",")
		for _, p := range parts {
			out = append(out, expandToken(strings.TrimSpace(p))...)
		}
		return out
	}
	if m := rangeRe.FindStringSubmatch(tok); len(m) == 3 {
		a, _ := strconv.Atoi(m[1])
		b, _ := strconv.Atoi(m[2])
		return expandRange(a, b)
	}
	if i, err := strconv.Atoi(tok); err == nil {
		return []int{i}
	}
	return nil
}

func expandRange(a, b int) []int {
	if a <= 0 || b <= 0 {
		return nil
	}
	if a > b {
		a, b = b, a
	}
	out := make([]int, 0, b-a+1)
	for i := a; i <= b; i++ {
		out = append(out, i)
	}
	return out
}

// toIntMaybe 尝试将任意类型转换为正整数，不可转换则返回 0
func toIntMaybe(v any) int {
	switch t := v.(type) {
	case int:
		if t > 0 {
			return t
		}
	case int64:
		if t > 0 {
			return int(t)
		}
	case float64:
		if t > 0 {
			return int(t)
		}
	case json.Number:
		if i, err := t.Int64(); err == nil && i > 0 {
			return int(i)
		}
		if f, err := t.Float64(); err == nil && f > 0 {
			return int(f)
		}
	case string:
		if strings.TrimSpace(t) != "" {
			if i, err := strconv.Atoi(strings.TrimSpace(t)); err == nil && i > 0 {
				return i
			}
		}
	}
	return 0
}

// canonicalizeKeyPagesMapUsingKeywords 尝试将 LLM 返回的键名归一化为规范的章节 key：
// - 允许精确 key 命中；
// - 允许通过 TitleKeywords 的同义词命中（忽略大小写，空格/横线->下划线等简单归一化）；
// - 合并同一规范 key 的页码并去重排序。
func canonicalizeKeyPagesMapUsingKeywords(mp map[string][]int, allowedKeys []string, spec *configs.DocSpec) map[string][]int {
	if len(mp) == 0 || spec == nil {
		return mp
	}
	allowed := makeSet(allowedKeys)
	// 构建规范 key -> true
	// 构建别名 norm(token) -> canonicalKey
	alias := map[string]string{}
	norm := func(s string) string {
		s = strings.TrimSpace(strings.ToLower(s))
		if s == "" {
			return s
		}
		s = strings.ReplaceAll(s, " ", "_")
		s = strings.ReplaceAll(s, "-", "_")
		s = strings.ReplaceAll(s, "：", "")
		s = strings.ReplaceAll(s, ":", "")
		return s
	}
	for _, k := range allowedKeys {
		alias[norm(k)] = k
		// 也接受空格版
		alias[norm(strings.ReplaceAll(k, "_", " "))] = k
	}
	// 根据 spec.Chapters 的 TitleKeywords 构建各 key 的别名
	for _, ch := range spec.Chapters {
		if _, ok := allowed[ch.Key]; !ok {
			continue
		}
		for _, kw := range ch.TitleKeywords {
			if strings.TrimSpace(kw) == "" {
				continue
			}
			alias[norm(kw)] = ch.Key
		}
	}
	out := map[string][]int{}
	for k, arr := range mp {
		if _, ok := allowed[k]; ok {
			out[k] = append(out[k], arr...)
			continue
		}
		if can, ok := alias[norm(k)]; ok {
			out[can] = append(out[can], arr...)
			continue
		}
		// 兜底：原样保留（后续会被 allowed 过滤掉，不影响当前组以外的无关键）
		out[k] = append(out[k], arr...)
	}
	for k, arr := range out {
		out[k] = uniqueSorted(arr)
	}
	return out
}
