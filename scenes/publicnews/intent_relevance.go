package publicnews

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

const (
	SourceRelevanceRuleOK                         = "ok"
	SourceRelevanceRuleTopicSpecificSupportNeeded = "topic_specific_support_needed"
	SourceRelevanceRuleEventCoherenceNeeded       = "event_coherence_needed"
	SourceRelevanceRuleDuplicateEvidenceCopy      = "duplicate_evidence_copy"
)

type eventClaimActuality uint8

const (
	eventClaimActualityUnknown eventClaimActuality = iota
	eventClaimActualityRealized
	eventClaimActualityProspective
)

// SourceRelevanceInput is source-neutral relevance evidence for choosing
// primary/supporting latest-news sources. It deliberately avoids publisher,
// region, provider, or trust policy.
type SourceRelevanceInput struct {
	Primary   LatestNewsLookupSource
	Candidate LatestNewsLookupSource
	Intent    LatestNewsLookupIntent
}

type SourceRelevanceDecision struct {
	Accepted bool     `json:"accepted"`
	RuleID   string   `json:"rule_id,omitempty"`
	Score    int      `json:"score,omitempty"`
	Reasons  []string `json:"reasons,omitempty"`
}

type SourceRelevancePolicy interface {
	EvaluateSupportingSource(SourceRelevanceInput) SourceRelevanceDecision
	ScoreSourceForIntent(LatestNewsLookupSource, LatestNewsLookupIntent) int
}

type defaultSourceRelevancePolicy struct{}

var eventFingerprintStopTerms = []string{
	"最新", "新闻", "消息", "资讯", "进展", "动态", "摘要", "来源", "时间", "影响", "风险", "市场", "核心", "事件", "事实",
	"可能", "预计", "预期", "或将", "有望", "大概率", "潜在",
	"表示", "认为", "宣布", "发布", "报道", "目前", "此前", "相关", "公司", "企业", "集团", "机构", "投资者", "关注", "数据", "信号",
	"人工", "智能", "人工智能", "智能体", "平台", "ai", "agent", "agents",
	"工作组", "指引", "会议", "讲话", "证词", "计划", "更新", "方向", "情况", "内容", "详情",
	"央行", "通胀", "加息", "降息", "黄金", "交易员",
	"人民币", "港元", "美元", "欧元", "日元", "英镑", "亿元", "万元", "千元",
	"latest", "news", "update", "updates", "brief", "summary", "source", "published", "impact", "risk", "market",
	"said", "says", "announced", "reported", "report", "reports", "investors", "company", "group", "today",
	"rmb", "cny", "hkd", "usd", "eur", "jpy", "gbp", "dollar", "dollars",
	"the", "and", "for", "with", "from", "this", "that", "will", "would", "could", "about",
}

var (
	eventChineseDatePattern = regexp.MustCompile(`(?i)(\d{1,2})\s*月\s*(\d{1,2})\s*日`)
	eventISODatePattern     = regexp.MustCompile(`(?i)(?:^|[^0-9])\d{4}[-/.](\d{1,2})[-/.](\d{1,2})(?:[^0-9]|$)`)
	eventQuantityPattern    = regexp.MustCompile(`(?i)([0-9][0-9,]*(?:\.[0-9]+)?)\s*(千|万|亿|k|thousand|m|mn|million|b|bn|billion)?\s*(股|shares?|名员工|员工|employees?|岗位|jobs?)`)
)

var eventRealizedClaimMarkers = []string{
	"已宣布", "正式宣布", "宣布决定", "已决定", "决定维持", "公布决定", "已经公布", "正式公布",
	"已批准", "正式批准", "已通过", "正式通过", "已发布", "正式发布", "已推出", "正式推出", "已完成",
	"has announced", "have announced", "announced its decision", "announced the decision", "officially announced",
	"has decided", "have decided", "decided to", "decision was announced", "decision has been",
	"has approved", "have approved", "officially approved", "has adopted", "have adopted",
	"has issued", "have issued", "has released", "have released", "has launched", "have launched",
	"has completed", "have completed",
}

var eventProspectiveClaimMarkers = []string{
	"预计", "预期", "预测", "概率", "押注", "或将", "可能会", "可能将", "有望", "尚未", "将于", "即将",
	"prediction", "forecast", "probability", "probabilities", "odds of", "expected to", "expects to",
	"likely to", "may decide", "might decide", "could decide", "scheduled to", "upcoming", "ahead of the",
	"awaiting the", "will the",
}

func DefaultSourceRelevancePolicy() SourceRelevancePolicy {
	return defaultSourceRelevancePolicy{}
}

// LatestNewsIntentSubjectAnchors returns the subject identity that evidence
// must mention. Explicit entity mentions win; otherwise the topic is reduced
// by common news-task and facet modifiers such as "latest", "policy", or
// "product release".
func LatestNewsIntentSubjectAnchors(intent LatestNewsLookupIntent) []string {
	out := []string{}
	for _, entity := range intent.EntityMentions {
		out = append(out, latestNewsSubjectAnchorAliases(entity)...)
	}
	if len(out) > 0 {
		return normalizeStringList(out)
	}
	reduced := strings.ToLower(strings.TrimSpace(intent.Topic))
	for _, modifier := range []string{
		"最新产品发布", "产品发布", "货币政策", "利率政策", "监管政策", "监管法规", "市场影响", "风险边界",
		"最新新闻", "最新进展", "最新动态", "新闻简报", "新闻", "资讯", "消息", "最新", "进展", "动态",
		"产品", "发布", "政策", "利率", "监管", "法规", "市场", "影响", "风险", "局势", "事件", "情况", "简报", "摘要",
		"interest rate", "monetary policy", "product launch", "product release", "latest news", "latest update",
	} {
		reduced = strings.ReplaceAll(reduced, modifier, " ")
	}
	englishModifiers := map[string]bool{
		"latest": true, "news": true, "update": true, "updates": true,
		"product": true, "launch": true, "release": true,
		"policy": true, "rate": true, "rates": true, "interest": true, "monetary": true,
		"regulation": true, "regulatory": true, "market": true,
		"impact": true, "risk": true, "risks": true, "situation": true, "event": true, "brief": true, "summary": true,
	}
	parts := strings.FieldsFunc(reduced, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	kept := []string{}
	for _, part := range parts {
		part = strings.ToLower(strings.TrimSpace(part))
		if part != "" && !englishModifiers[part] {
			kept = append(kept, part)
		}
	}
	anchor := normalizeIntentSubjectAnchor(strings.Join(kept, " "))
	if len([]rune(anchor)) < 2 {
		return nil
	}
	return []string{anchor}
}

func latestNewsSubjectAnchorAliases(value string) []string {
	anchor := normalizeIntentSubjectAnchor(value)
	if len([]rune(anchor)) < 2 {
		return nil
	}
	out := []string{anchor}
	for _, suffix := range []string{
		"股份有限公司", "有限责任公司", "控股集团", "有限公司", "控股", "集团", "公司",
		"holdingslimited", "holdinglimited", "grouplimited", "corporation", "incorporated",
		"holdings", "holding", "limited", "group", "corp", "inc", "ltd", "plc",
	} {
		if !strings.HasSuffix(anchor, suffix) {
			continue
		}
		base := strings.TrimSuffix(anchor, suffix)
		if subjectAliasBaseLongEnough(base) {
			out = append(out, base)
		}
		break
	}
	return normalizeStringList(out)
}

func subjectAliasBaseLongEnough(value string) bool {
	runes := []rune(strings.TrimSpace(value))
	if len(runes) < 2 {
		return false
	}
	for _, r := range runes {
		if unicode.In(r, unicode.Han, unicode.Hiragana, unicode.Katakana, unicode.Hangul) {
			return true
		}
	}
	return len(runes) >= 3
}

// TextMatchesLatestNewsIntent requires the selected evidence itself to carry
// the subject identity. Facet-only matches such as an unrelated central bank
// article containing "policy" and "rate" are insufficient.
func TextMatchesLatestNewsIntent(text string, intent LatestNewsLookupIntent) bool {
	haystack := normalizeIntentSubjectAnchor(text)
	if haystack == "" {
		return false
	}
	if anchors := LatestNewsIntentSubjectAnchors(intent); len(anchors) > 0 {
		for _, anchor := range anchors {
			if strings.Contains(haystack, anchor) {
				return true
			}
		}
		return false
	}
	tokens := TopicSpecificIntentTokens(intent)
	if len(tokens) == 0 {
		return true
	}
	for _, token := range tokens {
		if normalized := normalizeIntentSubjectAnchor(token); normalized != "" && strings.Contains(haystack, normalized) {
			return true
		}
	}
	return false
}

func normalizeIntentSubjectAnchor(value string) string {
	b := strings.Builder{}
	for _, r := range strings.ToLower(strings.TrimSpace(value)) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// SupportingSourceRelevantForIntent checks whether a cross-check source is
// specific enough for the structured topic. It intentionally works on generic
// task-frame tokens, not publisher or business-specific rules.
func SupportingSourceRelevantForIntent(primary LatestNewsLookupSource, source LatestNewsLookupSource, intent LatestNewsLookupIntent) bool {
	return DefaultSourceRelevancePolicy().EvaluateSupportingSource(SourceRelevanceInput{
		Primary:   primary,
		Candidate: source,
		Intent:    intent,
	}).Accepted
}

func (defaultSourceRelevancePolicy) EvaluateSupportingSource(input SourceRelevanceInput) SourceRelevanceDecision {
	if duplicateEvidenceCopy(input.Primary, input.Candidate) {
		return SourceRelevanceDecision{
			Accepted: false,
			RuleID:   SourceRelevanceRuleDuplicateEvidenceCopy,
			Reasons:  []string{SourceRelevanceRuleDuplicateEvidenceCopy},
		}
	}
	tokens := TopicSpecificIntentTokens(input.Intent)
	score := 0
	comparable, coherent := eventEvidenceCoherence(input.Primary, input.Candidate, input.Intent)
	if len(tokens) > 0 && TopicSpecificEvidenceScore(input.Primary, tokens) >= 2 {
		score = TopicSpecificEvidenceScore(input.Candidate, tokens)
		if score < 2 && !(comparable && coherent) {
			return SourceRelevanceDecision{
				Accepted: false,
				RuleID:   SourceRelevanceRuleTopicSpecificSupportNeeded,
				Score:    score,
				Reasons:  []string{SourceRelevanceRuleTopicSpecificSupportNeeded},
			}
		}
	}
	if comparable && !coherent {
		return SourceRelevanceDecision{
			Accepted: false,
			RuleID:   SourceRelevanceRuleEventCoherenceNeeded,
			Score:    score,
			Reasons:  []string{SourceRelevanceRuleEventCoherenceNeeded},
		}
	}
	return SourceRelevanceDecision{Accepted: true, RuleID: SourceRelevanceRuleOK, Score: score}
}

func duplicateEvidenceCopy(primary LatestNewsLookupSource, candidate LatestNewsLookupSource) bool {
	primaryUpdate := normalizeEvidenceCopyText(primary.KeyUpdate)
	candidateUpdate := normalizeEvidenceCopyText(candidate.KeyUpdate)
	if len([]rune(primaryUpdate)) >= 16 && primaryUpdate == candidateUpdate {
		return true
	}
	primaryText := normalizeEvidenceCopyText(primary.Text)
	candidateText := normalizeEvidenceCopyText(candidate.Text)
	return len([]rune(primaryText)) >= 80 && primaryText == candidateText
}

func normalizeEvidenceCopyText(value string) string {
	b := strings.Builder{}
	for _, r := range strings.ToLower(strings.TrimSpace(value)) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func (defaultSourceRelevancePolicy) ScoreSourceForIntent(source LatestNewsLookupSource, intent LatestNewsLookupIntent) int {
	tokens := []string{}
	for _, value := range append([]string{intent.Topic}, intent.EntityMentions...) {
		tokens = append(tokens, IntentTokenFragments(value)...)
	}
	tokens = normalizeStringList(tokens)
	if len(tokens) == 0 {
		return 0
	}
	return TopicSpecificSourceScore(source, tokens)
}

func TopicSpecificSourceScore(source LatestNewsLookupSource, tokens []string) int {
	score := TopicSpecificEvidenceScore(source, tokens)
	text := strings.ToLower(source.Text)
	for _, token := range tokens {
		token = strings.ToLower(strings.TrimSpace(token))
		if token != "" && strings.Contains(text, token) {
			score++
		}
	}
	return score
}

// TopicSpecificEvidenceScore only scores the title and selected key update.
// Body-only mentions are useful for ranking, but cannot establish that the
// evidence surfaced to the user actually cross-checks the requested topic.
func TopicSpecificEvidenceScore(source LatestNewsLookupSource, tokens []string) int {
	title := strings.ToLower(firstNonEmpty(source.Title, source.Headline))
	keyUpdate := strings.ToLower(source.KeyUpdate)
	score := 0
	for _, token := range tokens {
		token = strings.ToLower(strings.TrimSpace(token))
		if token == "" {
			continue
		}
		if strings.Contains(title, token) {
			score += 3
		}
		if strings.Contains(keyUpdate, token) {
			score += 2
		}
	}
	return score
}

func eventEvidenceCoherence(primary LatestNewsLookupSource, candidate LatestNewsLookupSource, intent LatestNewsLookupIntent) (bool, bool) {
	if eventClaimActualityConflict(primary, candidate) {
		return true, false
	}
	if eventDateAnchorsConflict(primary.KeyUpdate, candidate.KeyUpdate) {
		return true, false
	}
	if eventQuantityAnchorsConflict(primary.KeyUpdate, candidate.KeyUpdate) {
		return true, false
	}
	intentText := strings.ToLower(strings.Join(append([]string{intent.Topic}, intent.EntityMentions...), " "))
	primaryTokens := eventFingerprintTokens(primary.KeyUpdate, intentText)
	candidateTokens := eventFingerprintTokens(candidate.KeyUpdate, intentText)
	if len(primaryTokens) > 0 && len(candidateTokens) > 0 {
		keyScore := eventFingerprintOverlapScore(primaryTokens, candidateTokens)
		if keyScore >= 2 {
			return true, true
		}
		if keyScore == 0 {
			return true, false
		}
		primaryHeadlines := eventFingerprintTokens(firstNonEmpty(primary.Title, primary.Headline), intentText)
		candidateHeadlines := eventFingerprintTokens(firstNonEmpty(candidate.Title, candidate.Headline), intentText)
		return true, eventFingerprintOverlapScore(primaryHeadlines, candidateHeadlines) >= 2
	}
	primaryTokens = eventFingerprintTokens(firstNonEmpty(primary.Title, primary.Headline), intentText)
	candidateTokens = eventFingerprintTokens(firstNonEmpty(candidate.Title, candidate.Headline), intentText)
	if len(primaryTokens) == 0 || len(candidateTokens) == 0 {
		return false, true
	}
	return true, eventFingerprintOverlapScore(primaryTokens, candidateTokens) >= 2
}

func eventClaimActualityConflict(primary LatestNewsLookupSource, candidate LatestNewsLookupSource) bool {
	primaryActuality := eventSourceClaimActuality(primary)
	candidateActuality := eventSourceClaimActuality(candidate)
	return primaryActuality != eventClaimActualityUnknown &&
		candidateActuality != eventClaimActualityUnknown &&
		primaryActuality != candidateActuality
}

func eventSourceClaimActuality(source LatestNewsLookupSource) eventClaimActuality {
	value := strings.ToLower(strings.Join([]string{
		firstNonEmpty(source.Title, source.Headline),
		source.KeyUpdate,
	}, "\n"))
	// A source reporting a completed decision may still say it was expected.
	// Explicit realization therefore wins over prospective context.
	if containsEventClaimMarker(value, eventRealizedClaimMarkers) {
		return eventClaimActualityRealized
	}
	if containsEventClaimMarker(value, eventProspectiveClaimMarkers) {
		return eventClaimActualityProspective
	}
	return eventClaimActualityUnknown
}

func containsEventClaimMarker(value string, markers []string) bool {
	for _, marker := range markers {
		if strings.Contains(value, marker) {
			return true
		}
	}
	return false
}

func eventDateAnchorsConflict(primary string, candidate string) bool {
	primaryDates := eventDateAnchors(primary)
	candidateDates := eventDateAnchors(candidate)
	if len(primaryDates) == 0 || len(candidateDates) == 0 {
		return false
	}
	for date := range primaryDates {
		if candidateDates[date] {
			return false
		}
	}
	return true
}

func eventDateAnchors(value string) map[string]bool {
	out := map[string]bool{}
	addMatches := func(pattern *regexp.Regexp) {
		for _, match := range pattern.FindAllStringSubmatch(value, -1) {
			if len(match) < 3 {
				continue
			}
			month, monthErr := strconv.Atoi(match[1])
			day, dayErr := strconv.Atoi(match[2])
			if monthErr != nil || dayErr != nil || month < 1 || month > 12 || day < 1 || day > 31 {
				continue
			}
			out[fmt.Sprintf("%02d-%02d", month, day)] = true
		}
	}
	addMatches(eventChineseDatePattern)
	addMatches(eventISODatePattern)
	return out
}

func eventQuantityAnchorsConflict(primary string, candidate string) bool {
	primaryAnchors := eventQuantityAnchors(primary)
	candidateAnchors := eventQuantityAnchors(candidate)
	for category, primaryValues := range primaryAnchors {
		candidateValues := candidateAnchors[category]
		if len(primaryValues) == 0 || len(candidateValues) == 0 {
			continue
		}
		if !eventQuantityValuesOverlap(primaryValues, candidateValues) {
			return true
		}
	}
	return false
}

func eventQuantityAnchors(value string) map[string][]float64 {
	out := map[string][]float64{}
	for _, match := range eventQuantityPattern.FindAllStringSubmatchIndex(value, -1) {
		if len(match) < 8 {
			continue
		}
		numberText := value[match[2]:match[3]]
		multiplierText := ""
		if match[4] >= 0 && match[5] >= 0 {
			multiplierText = value[match[4]:match[5]]
		}
		unitText := value[match[6]:match[7]]
		number, err := strconv.ParseFloat(strings.ReplaceAll(numberText, ",", ""), 64)
		if err != nil || math.IsNaN(number) || math.IsInf(number, 0) || number < 0 {
			continue
		}
		number *= eventQuantityMultiplier(multiplierText)
		category := eventQuantityCategory(unitText, quantityPrefix(value, match[0]), quantitySuffix(value, match[1]))
		if category == "" {
			continue
		}
		duplicate := false
		for _, current := range out[category] {
			if eventQuantityValuesEquivalent(current, number) {
				duplicate = true
				break
			}
		}
		if !duplicate {
			out[category] = append(out[category], number)
		}
	}
	return out
}

func quantityPrefix(value string, byteIndex int) string {
	if byteIndex <= 0 {
		return ""
	}
	runes := []rune(value[:byteIndex])
	if len(runes) > 24 {
		runes = runes[len(runes)-24:]
	}
	return string(runes)
}

func quantitySuffix(value string, byteIndex int) string {
	if byteIndex >= len(value) {
		return ""
	}
	runes := []rune(value[byteIndex:])
	if len(runes) > 16 {
		runes = runes[:16]
	}
	return string(runes)
}

func eventQuantityMultiplier(value string) float64 {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "千", "k", "thousand":
		return 1_000
	case "万":
		return 10_000
	case "m", "mn", "million":
		return 1_000_000
	case "亿":
		return 100_000_000
	case "b", "bn", "billion":
		return 1_000_000_000
	default:
		return 1
	}
}

func eventQuantityCategory(value string, prefix string, suffix string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	switch {
	case value == "股", strings.HasPrefix(value, "share"):
		if shareQuantityIsCapitalTotal(prefix, suffix) {
			return ""
		}
		return "shares"
	case value == "名员工", value == "员工", strings.HasPrefix(value, "employee"), value == "岗位", strings.HasPrefix(value, "job"):
		return "people_or_jobs"
	default:
		return ""
	}
}

func shareQuantityIsCapitalTotal(prefix string, suffix string) bool {
	prefix = strings.ToLower(prefix)
	for _, marker := range []string{
		"已发行股份总数", "已發行股份總數", "股份总数", "股份總數", "总股本", "總股本",
		"发行在外股份", "發行在外股份", "total number of shares", "total issued shares",
		"shares outstanding", "outstanding shares", "issued share capital",
	} {
		if strings.Contains(prefix, marker) {
			return true
		}
	}
	suffix = strings.ToLower(strings.TrimLeft(suffix, " \t:：-"))
	for _, marker := range []string{"已发行股份", "已發行股份", "为已发行股份总数", "為已發行股份總數"} {
		if strings.HasPrefix(suffix, marker) {
			return true
		}
	}
	return false
}

func eventQuantityValuesOverlap(primary []float64, candidate []float64) bool {
	for _, left := range primary {
		for _, right := range candidate {
			if eventQuantityValuesEquivalent(left, right) {
				return true
			}
		}
	}
	return false
}

func eventQuantityValuesEquivalent(left float64, right float64) bool {
	scale := math.Max(math.Abs(left), math.Abs(right))
	tolerance := math.Max(0.5, scale*0.01)
	return math.Abs(left-right) <= tolerance
}

func eventFingerprintOverlapScore(primary map[string]bool, candidate map[string]bool) int {
	score := 0
	for token := range primary {
		if !candidate[token] {
			continue
		}
		weight := 1
		if countCJKRunes(token) >= 3 {
			weight = 2
		} else {
			for _, r := range token {
				if unicode.IsDigit(r) {
					weight = 2
					break
				}
			}
		}
		score += weight
	}
	return score
}

// HeadlineEventEvidenceCoherent reports whether a factual sentence carries
// enough non-generic event anchors from the headline to serve as its key
// update. Entity-only or market-context overlap is intentionally insufficient.
func HeadlineEventEvidenceCoherent(headline string, evidence string, intent LatestNewsLookupIntent) bool {
	return HeadlineEventEvidenceScore(headline, evidence, intent) >= 2
}

// HeadlineEventEvidenceScore ranks factual sentences by their overlap with the
// non-generic event anchors in the headline.
func HeadlineEventEvidenceScore(headline string, evidence string, intent LatestNewsLookupIntent) int {
	intentText := strings.ToLower(strings.Join(append([]string{intent.Topic}, intent.EntityMentions...), " "))
	headlineTokens := eventFingerprintTokens(headline, intentText)
	evidenceTokens := eventFingerprintTokens(evidence, intentText)
	if len(headlineTokens) == 0 || len(evidenceTokens) == 0 {
		return 0
	}
	return eventFingerprintOverlapScore(headlineTokens, evidenceTokens)
}

func eventFingerprintTokens(value string, intentText string) map[string]bool {
	out := map[string]bool{}
	value = trimEventFingerprintLabel(strings.ToLower(strings.TrimSpace(value)))
	for _, contextNoise := range []string{
		"欧元区", "人工智能公司", "人工智能企业",
		"euro area", "artificial intelligence company", "ai company",
	} {
		value = strings.ReplaceAll(value, contextNoise, " ")
	}
	if value == "" {
		return out
	}
	flushCJK := func(segment []rune) {
		for size := 2; size <= 4 && size <= len(segment); size++ {
			for start := 0; start+size <= len(segment); start++ {
				token := string(segment[start : start+size])
				if eventFingerprintTokenUsable(token, intentText) {
					out[token] = true
				}
			}
		}
	}
	flushASCII := func(segment []rune) {
		token := strings.Trim(string(segment), "-._")
		if eventFingerprintTokenUsable(token, intentText) {
			out[token] = true
		}
	}
	cjk := []rune{}
	ascii := []rune{}
	flush := func() {
		flushCJK(cjk)
		flushASCII(ascii)
		cjk = cjk[:0]
		ascii = ascii[:0]
	}
	for _, r := range value {
		switch {
		case unicode.In(r, unicode.Han, unicode.Hiragana, unicode.Katakana, unicode.Hangul):
			if len(ascii) > 0 {
				flushASCII(ascii)
				ascii = ascii[:0]
			}
			cjk = append(cjk, r)
		case unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '.' || r == '_':
			if len(cjk) > 0 {
				flushCJK(cjk)
				cjk = cjk[:0]
			}
			ascii = append(ascii, r)
		default:
			flush()
		}
	}
	flush()
	return out
}

func trimEventFingerprintLabel(value string) string {
	for {
		trimmed := strings.TrimSpace(value)
		matched := false
		for _, prefix := range []string{
			"核心事件：", "核心事件:", "关键事件：", "关键事件:",
			"核心事实：", "核心事实:", "关键事实：", "关键事实:",
			"事件摘要：", "事件摘要:", "core event:", "key event:", "key update:",
		} {
			if strings.HasPrefix(trimmed, prefix) {
				value = strings.TrimSpace(strings.TrimPrefix(trimmed, prefix))
				matched = true
				break
			}
		}
		if !matched {
			return trimmed
		}
	}
}

func eventFingerprintTokenUsable(token string, intentText string) bool {
	token = strings.ToLower(strings.TrimSpace(token))
	if len([]rune(token)) < 2 || strings.Contains(intentText, token) {
		return false
	}
	if eventFingerprintCurrencyUnitToken(token) {
		return false
	}
	for _, stop := range eventFingerprintStopTerms {
		if token == stop || (hasCJK(token) && hasCJK(stop) && strings.Contains(stop, token)) {
			return false
		}
	}
	allDigits := true
	for _, r := range token {
		if !unicode.IsDigit(r) && r != '-' && r != '.' && r != '_' {
			allDigits = false
			break
		}
	}
	return !allDigits
}

func eventFingerprintCurrencyUnitToken(token string) bool {
	for _, currency := range []string{"人民币", "港元", "美元", "欧元", "日元", "英镑"} {
		if strings.Contains(currency, token) || strings.Contains(token, currency) {
			return true
		}
		for _, scale := range []string{"千", "万", "亿"} {
			scaled := scale + currency
			if strings.Contains(scaled, token) || strings.Contains(token, scaled) {
				return true
			}
		}
	}
	return false
}

func TopicSpecificIntentTokens(intent LatestNewsLookupIntent) []string {
	entityTokens := map[string]bool{}
	for _, entity := range intent.EntityMentions {
		for _, token := range IntentTokenFragments(entity) {
			entityTokens[strings.ToLower(strings.TrimSpace(token))] = true
		}
	}
	topic := strings.ToLower(strings.TrimSpace(intent.Topic))
	for entityToken := range entityTokens {
		if entityToken != "" {
			topic = strings.ReplaceAll(topic, entityToken, " ")
		}
	}
	for _, generic := range GenericLatestNewsIntentTerms() {
		topic = strings.ReplaceAll(topic, generic, " ")
	}
	out := []string{}
	for _, token := range IntentTokenFragments(topic) {
		normalized := strings.ToLower(strings.TrimSpace(token))
		if normalized == "" || tokenOverlapsEntity(normalized, entityTokens) || GenericLatestNewsIntentToken(normalized) {
			continue
		}
		out = append(out, normalized)
	}
	return normalizeStringList(out)
}

func IntentTokenFragments(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	tokens := []string{value}
	if hasCJK(value) {
		runes := []rune(value)
		for size := 2; size <= 4 && size <= len(runes); size++ {
			for start := 0; start+size <= len(runes); start++ {
				tokens = append(tokens, string(runes[start:start+size]))
			}
		}
	}
	for _, part := range strings.FieldsFunc(value, func(r rune) bool {
		return r == ' ' || r == '\t' || r == '\n' || r == ',' || r == '，' || r == '/' || r == '、' || r == '(' || r == ')' || r == '（' || r == '）' || r == '-' || r == '_'
	}) {
		if part = strings.TrimSpace(part); len([]rune(part)) >= 2 {
			tokens = append(tokens, part)
		}
	}
	return normalizeStringList(tokens)
}

func tokenOverlapsEntity(token string, entityTokens map[string]bool) bool {
	if entityTokens[token] {
		return true
	}
	for entityToken := range entityTokens {
		if entityToken == "" {
			continue
		}
		if strings.Contains(token, entityToken) || strings.Contains(entityToken, token) {
			return true
		}
	}
	return false
}

func GenericLatestNewsIntentToken(token string) bool {
	token = strings.ToLower(strings.TrimSpace(token))
	if token == "" {
		return true
	}
	for _, generic := range GenericLatestNewsIntentTerms() {
		if token == generic || strings.Contains(token, generic) {
			return true
		}
	}
	return false
}

func GenericLatestNewsIntentTerms() []string {
	return []string{
		"最新", "新闻", "资讯", "消息", "进展", "动态", "情况", "简报", "摘要", "来源", "时间", "影响", "风险", "后续", "市场",
		"latest", "news", "update", "updates", "brief", "summary", "source", "published", "impact", "risk", "risks", "market",
	}
}

func hasCJK(value string) bool {
	for _, r := range value {
		if r >= '\u4e00' && r <= '\u9fff' {
			return true
		}
	}
	return false
}
