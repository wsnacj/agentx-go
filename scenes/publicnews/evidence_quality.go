package publicnews

import (
	"strings"
	"unicode"
)

const (
	EvidenceQualityRuleOK                  = "ok"
	EvidenceQualityRuleEmpty               = "key_update_empty"
	EvidenceQualityRuleEncodingNoise       = "key_update_encoding_noise"
	EvidenceQualityRuleMarkupNoise         = "key_update_markup_noise"
	EvidenceQualityRuleBoilerplateNoise    = "key_update_boilerplate_noise"
	EvidenceQualityRulePromotionalNoise    = "key_update_promotional_noise"
	EvidenceQualityRuleHypotheticalRewrite = "key_update_hypothetical_rewrite"
	EvidenceQualityRuleLowInformation      = "key_update_low_information"
	EvidenceQualityRuleTooShort            = "key_update_too_short"
	EvidenceQualityRuleDateCategoryLine    = "key_update_date_category_line"
	EvidenceQualityRuleHeadlineRestatement = "key_update_headline_restatement"
	EvidenceQualityRuleSourceMetadataLine  = "source_metadata_line"
)

// EvidenceQualityInput is source-neutral evidence text used by extractors,
// guards, and host ranking. It deliberately excludes publisher/provider policy.
type EvidenceQualityInput struct {
	Headline    string
	KeyUpdate   string
	Description string
	Line        string
}

type EvidenceQualityDecision struct {
	Accepted bool     `json:"accepted"`
	Score    int      `json:"score,omitempty"`
	RuleID   string   `json:"rule_id,omitempty"`
	Reasons  []string `json:"reasons,omitempty"`
}

type EvidenceQualityPolicy interface {
	EvaluateKeyUpdate(EvidenceQualityInput) EvidenceQualityDecision
	EvaluateLine(EvidenceQualityInput) EvidenceQualityDecision
	ScoreSpecificity(EvidenceQualityInput) int
}

type defaultEvidenceQualityPolicy struct{}

func DefaultEvidenceQualityPolicy() EvidenceQualityPolicy {
	return defaultEvidenceQualityPolicy{}
}

func EvidenceSpecificityScore(headline string, description string, keyUpdate string) int {
	return DefaultEvidenceQualityPolicy().ScoreSpecificity(EvidenceQualityInput{
		Headline:    headline,
		Description: description,
		KeyUpdate:   keyUpdate,
	})
}

// EvidenceTextLooksNoisy reports generic evidence text noise such as source
// prefixes, quote pages, app-store boilerplate, and promotional CTAs.
func EvidenceTextLooksNoisy(value string) bool {
	lower := strings.ToLower(strings.TrimSpace(value))
	if lower == "" {
		return true
	}
	return looksLikeEvidencePrefixNoise(lower) ||
		EvidenceTextLooksEncodingCorrupt(lower) ||
		looksLikeBoilerplateEvidenceText(lower) ||
		looksLikePromotionalKeyUpdate(lower)
}

func KeyUpdateSufficientForHeadline(headline string, value string) bool {
	return DefaultEvidenceQualityPolicy().EvaluateKeyUpdate(EvidenceQualityInput{
		Headline:  headline,
		KeyUpdate: value,
	}).Accepted
}

func KeyUpdateSufficient(value string) bool {
	return DefaultEvidenceQualityPolicy().EvaluateKeyUpdate(EvidenceQualityInput{
		KeyUpdate: value,
	}).Accepted
}

func EvidenceLineUsable(headline string, line string) bool {
	return DefaultEvidenceQualityPolicy().EvaluateLine(EvidenceQualityInput{
		Headline:  headline,
		Line:      line,
		KeyUpdate: line,
	}).Accepted
}

func (defaultEvidenceQualityPolicy) EvaluateLine(input EvidenceQualityInput) EvidenceQualityDecision {
	line := strings.TrimSpace(input.Line)
	if line == "" {
		return rejectEvidence(EvidenceQualityRuleEmpty)
	}
	if containsAnyFold(line, "发布时间", "published at", "来源 ", "来源:", "来源：", "原标题", "original title") {
		return rejectEvidence(EvidenceQualityRuleSourceMetadataLine)
	}
	if looksLikeAuthorshipMetadataLine(strings.ToLower(line)) {
		return rejectEvidence(EvidenceQualityRuleSourceMetadataLine)
	}
	return DefaultEvidenceQualityPolicy().EvaluateKeyUpdate(EvidenceQualityInput{
		Headline:  input.Headline,
		KeyUpdate: line,
	})
}

func looksLikeAuthorshipMetadataLine(lower string) bool {
	hasAuthor := containsAnyFold(lower, "作者", "撰文", "文 |", "文｜", " by ", "by:")
	hasEditor := containsAnyFold(lower, "编辑", "责编", "责任编辑", "edited by", "editor:")
	return (hasAuthor && hasEditor) || (strings.Contains(lower, "原创") && hasAuthor)
}

func (defaultEvidenceQualityPolicy) EvaluateKeyUpdate(input EvidenceQualityInput) EvidenceQualityDecision {
	value := strings.TrimSpace(input.KeyUpdate)
	if value == "" || strings.EqualFold(value, "unknown") {
		return rejectEvidence(EvidenceQualityRuleEmpty)
	}
	lower := strings.ToLower(value)
	if EvidenceTextLooksEncodingCorrupt(value) {
		return rejectEvidence(EvidenceQualityRuleEncodingNoise)
	}
	if strings.Contains(lower, "<!doctype") ||
		strings.Contains(lower, "<script") ||
		strings.Contains(lower, "</html") ||
		strings.Contains(lower, "window.") {
		return rejectEvidence(EvidenceQualityRuleMarkupNoise)
	}
	if looksLikeEvidencePrefixNoise(lower) || looksLikeBoilerplateEvidenceText(lower) {
		return rejectEvidence(EvidenceQualityRuleBoilerplateNoise)
	}
	if looksLikeDateCategoryLine(value) {
		return rejectEvidence(EvidenceQualityRuleDateCategoryLine)
	}
	if looksLikePromotionalKeyUpdate(lower) {
		return rejectEvidence(EvidenceQualityRulePromotionalNoise)
	}
	if looksLikeHypotheticalRewrite(lower) {
		return rejectEvidence(EvidenceQualityRuleHypotheticalRewrite)
	}
	if keyUpdateRestatesHeadlineWithoutFactSignal(input.Headline, value) {
		return rejectEvidence(EvidenceQualityRuleHeadlineRestatement)
	}
	if looksLikeNarrativeLeadKeyUpdate(lower) || looksLikeCourtesyOnlyKeyUpdate(lower) {
		return rejectEvidence(EvidenceQualityRuleLowInformation)
	}
	hasFactSignal := keyUpdateHasFactSignal(value)
	if looksLikeLowInformationKeyUpdate(lower) && !hasFactSignal {
		return rejectEvidence(EvidenceQualityRuleLowInformation)
	}
	if len([]rune(value)) < 10 {
		return rejectEvidence(EvidenceQualityRuleTooShort)
	}
	if cjkCount := countCJKRunes(value); cjkCount > 0 {
		if cjkCount < 10 {
			return rejectEvidence(EvidenceQualityRuleTooShort)
		}
	} else {
		if meaningfulWordCount(value) < 4 || len([]rune(value)) < 16 {
			return rejectEvidence(EvidenceQualityRuleTooShort)
		}
	}
	return EvidenceQualityDecision{Accepted: true, RuleID: EvidenceQualityRuleOK}
}

func (defaultEvidenceQualityPolicy) ScoreSpecificity(input EvidenceQualityInput) int {
	keyUpdate := strings.TrimSpace(input.KeyUpdate)
	evidenceText := firstNonEmpty(keyUpdate, input.Description)
	if evidenceText == "" {
		return -12
	}
	score := 0
	restatesTitle := keyUpdate != "" && keyUpdateRestatesHeadlineWithoutFactSignal(input.Headline, keyUpdate)
	hasFactSignal := keyUpdateHasFactSignal(evidenceText)
	if restatesTitle {
		score -= 24
	}
	if hasFactSignal {
		score += 8
	} else {
		score -= 14
	}
	if restatesTitle && !hasFactSignal {
		score -= 8
	}
	if len([]rune(evidenceText)) < 20 {
		score -= 4
	} else if len([]rune(evidenceText)) >= 36 {
		score += 2
	}
	return score
}

func rejectEvidence(ruleID string) EvidenceQualityDecision {
	return EvidenceQualityDecision{Accepted: false, RuleID: ruleID, Reasons: []string{ruleID}}
}

func keyUpdateRestatesHeadlineWithoutFactSignal(headline string, value string) bool {
	headline = normalizeHeadlineRestatementText(headline)
	value = normalizeHeadlineRestatementText(value)
	if len([]rune(headline)) < 8 || len([]rune(value)) < 8 {
		return false
	}
	if headline == value {
		return true
	}
	if !strings.Contains(headline, value) && !strings.Contains(value, headline) {
		return false
	}
	longer, shorter := []rune(headline), []rune(value)
	if len(shorter) > len(longer) {
		longer, shorter = shorter, longer
	}
	// Publishers commonly append a short site label to an otherwise identical
	// headline. That suffix must not turn the headline itself into evidence.
	if len(longer)-len(shorter) <= 8 && len(shorter)*4 >= len(longer)*3 {
		return true
	}
	return !keyUpdateHasFactSignal(value)
}

func normalizeHeadlineRestatementText(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	replacer := strings.NewReplacer(
		" ", "", "\t", "", "\n", "",
		"，", "", ",", "", "。", "", ".", "",
		"！", "", "!", "", "？", "", "?", "",
		"：", "", ":", "", "；", "", ";", "",
		"、", "", "-", "", "_", "", "—", "",
		"“", "", "”", "", "\"", "", "'", "",
		"（", "", "）", "", "(", "", ")", "",
		"《", "", "》", "",
	)
	return replacer.Replace(value)
}

func keyUpdateHasFactSignal(value string) bool {
	lower := strings.ToLower(strings.TrimSpace(value))
	if lower == "" {
		return false
	}
	for _, r := range lower {
		if (r >= '0' && r <= '9') || strings.ContainsRune("零一二三四五六七八九十百千万亿", r) {
			return true
		}
	}
	for _, marker := range []string{
		"%", "％", "$", "美元", "亿元", "万", "bp", "basis point",
		"according to", "reported", "said", "announced", "confirmed", "released", "release", "launched", "launch",
		"表示", "称", "宣布", "发布", "确认", "披露", "报道", "据", "消息", "公布",
		"批准", "投票", "通过", "下调", "上调", "增长", "下降", "收跌", "收涨",
		"风险", "波动", "资金流向", "市场预期", "预期变化",
	} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

func looksLikeEvidencePrefixNoise(lower string) bool {
	for _, prefix := range []string{
		"description:", "title:", "image:", "视频简介", "video description",
		"打开网易新闻", "打开新闻", "open in the news app",
	} {
		if strings.HasPrefix(lower, prefix) {
			return true
		}
	}
	return false
}

func looksLikeBoilerplateEvidenceText(lower string) bool {
	if looksLikeChineseSiteNavigationBoilerplate(lower) {
		return true
	}
	for _, marker := range []string{
		"active monthly users",
		"guides and reviews articles",
		"international team authors",
		"years on the market",
		"live prices & charts",
		"discover the institutional",
		"start your day with",
		"join moomoo",
		"online stock and options trading platform",
		"historical data for",
		"buy and sell stocks",
		"commission-free",
		"stock quote",
		"stock quotes",
		"real-time quotes",
		"stock market data",
		"24 hours a day, five days a week",
		"app store",
		"download on the app",
		"requires ios",
		"privacy practices",
		"rate monitor",
		"利率观测器",
		"加息降息预测",
		"economic calendar",
		"财经日历",
		"related news",
		"相关新闻",
		"查看精彩图片",
		"view images in the app",
		"没有更多",
		"no more results",
		"全部内容",
		"all content",
		"加载更多",
		"load more",
		"热门话题",
		"trending topics",
		"热门股票",
		"栏目信息",
		"program information",
		"换一换",
	} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

func looksLikeChineseSiteNavigationBoilerplate(lower string) bool {
	portalSignals := 0
	for _, marker := range []string{
		"今日热门", "本周热门", "本月热门", "热门行业", "主要行业",
		"最新推荐", "最新买入", "最新上调", "精选研报", "用户已上传",
		"更多 >>", "客户端下载", "正在加载，请稍候",
	} {
		if strings.Contains(lower, marker) {
			portalSignals++
		}
	}
	if portalSignals >= 4 {
		return true
	}
	controlSignals := 0
	for _, marker := range []string{
		"桌面版", "手机版", "最新搜", "返回", "放大", "缩小", "收起", "展开", "推荐", "开户",
	} {
		if strings.Contains(lower, marker) {
			controlSignals++
		}
	}
	if controlSignals >= 3 {
		return true
	}
	signals := 0
	for _, marker := range []string{
		"设为首页",
		"加入收藏",
		"当前位置",
		"正文",
		"地址",
		"邮编",
		"歡迎",
		"欢迎",
	} {
		if strings.Contains(lower, marker) {
			signals++
		}
	}
	if signals < 3 {
		return false
	}
	return strings.Contains(lower, "当前位置") ||
		strings.Contains(lower, "设为首页") ||
		strings.Contains(lower, "加入收藏") ||
		(strings.Contains(lower, "地址") && strings.Contains(lower, "邮编"))
}

func looksLikePromotionalKeyUpdate(lower string) bool {
	for _, marker := range []string{
		"sign up",
		"register now",
		"download app",
		"promo code",
		"referral code",
		"open account",
		"claim bonus",
		"analyst report",
		"新人注册",
		"注册首选",
		"开户注册",
		"立即注册",
		"立即开户",
		"下载app",
		"下载 app",
		"优惠码",
		"开户链接",
		"官方注册链接",
		"注册送",
		"首存",
		"币种最多最全",
		"最多最全",
		"炒股就看",
		"金麒麟分析师研报",
		"挖掘潜力主题机会",
		"抄底策略",
		"满仓追",
	} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	if strings.Contains(lower, "我在") &&
		(strings.Contains(lower, "满仓") || strings.Contains(lower, "追进去") || strings.Contains(lower, "亏掉")) {
		return true
	}
	if strings.Contains(lower, "交易所") && strings.Contains(lower, "|") {
		return true
	}
	return false
}

func looksLikeHypotheticalRewrite(lower string) bool {
	paraphraseLead := strings.HasPrefix(lower, "换句话说") ||
		strings.HasPrefix(lower, "也就是说") ||
		strings.HasPrefix(lower, "换言之") ||
		strings.HasPrefix(lower, "in other words") ||
		strings.HasPrefix(lower, "put differently")
	if !paraphraseLead {
		return false
	}
	return containsAnyFold(lower, "如果", "要是", "假如", "倘若", " if ")
}

func looksLikeLowInformationKeyUpdate(lower string) bool {
	for _, marker := range []string{
		"备受市场关注",
		"市场关注",
		"值得关注",
		"最新动向",
		"后续值得关注",
		"有待进一步关注",
		"迎来重大考验",
		"面临重大考验",
		"重大考验",
		"重磅来袭",
		"重大变数",
		"大变数",
		"家事国事天下事",
		"事事关心",
		"并不限于某一具体事件",
		"深层动机",
		"will be closely watched",
		"closely watched",
		"worth watching",
		"faces a major test",
		"major test awaits",
		"major uncertainty",
	} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

func looksLikeNarrativeLeadKeyUpdate(lower string) bool {
	for _, marker := range []string{
		"一个幽灵盘旋",
		"它的名字便是",
		"石油幽灵",
		"历史渊源与深层动机",
	} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

func looksLikeCourtesyOnlyKeyUpdate(lower string) bool {
	courtesy := containsAnyFold(lower,
		"感谢您一直以来给予",
		"感谢您一直以来的支持",
		"感谢广大用户的支持",
		"感谢各位用户的支持",
		"感谢您的理解和支持",
		"thank you for your continued support",
		"we appreciate your continued support",
	)
	if !courtesy {
		return false
	}
	return !containsAnyFold(lower,
		"宣布", "发布", "确认", "披露", "终止", "停止", "关闭", "下线", "上线", "将于", "正式",
		"announced", "confirmed", "will close", "will shut", "will end", "effective ",
	)
}

func EvidenceTextLooksEncodingCorrupt(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return false
	}
	replacementCount := 0
	for _, r := range value {
		if r == '\uFFFD' {
			replacementCount++
		}
	}
	if replacementCount > 0 {
		return true
	}
	lower := strings.ToLower(value)
	for _, marker := range []string{
		"â€", "â€™", "â€œ", "â€˜", "â€“", "â€”",
		"ã€", "ã€‚", "ã€", "ã", "ã", "ã",
		"å¹´", "æœˆ", "æ—¥", "ä¸", "äº", "æ–", "çš„",
		"鈥", "銆", "锛", "涓€", "鍙", "鏂", "閲", "绋", "绠", "绛",
	} {
		if strings.Contains(lower, strings.ToLower(marker)) {
			return true
		}
	}
	return false
}

func countCJKRunes(value string) int {
	count := 0
	for _, r := range value {
		if unicode.In(r, unicode.Han, unicode.Hiragana, unicode.Katakana, unicode.Hangul) {
			count++
		}
	}
	return count
}

func meaningfulWordCount(value string) int {
	count := 0
	inToken := false
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if !inToken {
				count++
				inToken = true
			}
			continue
		}
		inToken = false
	}
	return count
}

func looksLikeDateCategoryLine(value string) bool {
	if looksLikeDateTimeOnly(value) {
		return true
	}
	lower := strings.ToLower(value)
	hasMonth := containsAnyFold(lower,
		"jan", "feb", "mar", "apr", "may", "jun",
		"jul", "aug", "sep", "oct", "nov", "dec",
	)
	if !hasMonth {
		return false
	}
	hasDigit := false
	for _, r := range value {
		if unicode.IsDigit(r) {
			hasDigit = true
			break
		}
	}
	return hasDigit && meaningfulWordCount(value) <= 3
}

func looksLikeDateTimeOnly(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return false
	}
	allowedWords := map[string]bool{
		"am": true, "pm": true, "t": true, "z": true,
		"utc": true, "gmt": true, "cst": true, "est": true, "edt": true,
		"pst": true, "pdt": true, "cet": true, "cest": true, "bst": true,
		"jst": true, "hkt": true,
		"jan": true, "january": true, "feb": true, "february": true,
		"mar": true, "march": true, "apr": true, "april": true,
		"may": true, "jun": true, "june": true, "jul": true, "july": true,
		"aug": true, "august": true, "sep": true, "sept": true, "september": true,
		"oct": true, "october": true, "nov": true, "november": true,
		"dec": true, "december": true,
	}
	hasDigit := false
	hasDateSignal := false
	word := strings.Builder{}
	flushWord := func() bool {
		if word.Len() == 0 {
			return true
		}
		token := strings.ToLower(word.String())
		word.Reset()
		if !allowedWords[token] {
			return false
		}
		hasDateSignal = true
		return true
	}
	for _, r := range value {
		switch {
		case unicode.IsDigit(r):
			if !flushWord() {
				return false
			}
			hasDigit = true
		case r <= unicode.MaxASCII && unicode.IsLetter(r):
			word.WriteRune(r)
		case strings.ContainsRune("年月日时分秒周星期", r):
			if !flushWord() {
				return false
			}
			hasDateSignal = true
		case unicode.IsSpace(r) || strings.ContainsRune("-/:.,+()[]", r):
			if !flushWord() {
				return false
			}
			if strings.ContainsRune("-/:.", r) {
				hasDateSignal = true
			}
		default:
			return false
		}
	}
	return flushWord() && hasDigit && hasDateSignal
}
