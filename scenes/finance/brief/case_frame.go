package brief

import (
	"regexp"
	"strings"
	"unicode"
)

var briefStockCodePattern = regexp.MustCompile(`(?i)\b(?:sh|sz|bj|hk)?\s*(\d{5,6})\b`)

// BuildBriefCaseInput is a legacy/explicit pack-workflow fallback for
// materializing case inputs after a host has already selected this pack. It is
// not the default natural-language routing path. Normal AgentX short-query turns
// should use finance_report_lookup so the model supplies structured intent and
// adapters verify report identity, source, and extracted facts.
//
// Keep this helper source-neutral. Source resolution and PDF/docparse adapters
// stay in project/plugin layers.
func BuildBriefCaseInput(userMessage string) (map[string]any, bool) {
	message := strings.TrimSpace(userMessage)
	if message == "" || !looksLikeFinancialReportBrief(message) {
		return nil, false
	}
	entity := briefEntityFrame(message)
	name, _ := entity["name"].(string)
	if strings.TrimSpace(name) == "" {
		return nil, false
	}
	return map[string]any{
		"user_message": message,
		"entity":       entity,
		"brief_focus": []any{
			"financial_highlights",
			"business_overview",
			"management_discussion",
			"risk_or_uncertainty",
			"shareholder_return",
		},
		"period_policy":  briefPeriodPolicy(message),
		"source_policy":  "public_web_prefer_official_annual_report_pdf",
		"freshness":      "live",
		"output_style":   briefOutputStyle(message),
		"stop_condition": "guard_passed",
	}, true
}

func looksLikeFinancialReportBrief(message string) bool {
	return containsAnyBrief(message, "财报", "年报", "年度报告", "财务报告", "annual report", "financial report", "20-f", "form 20-f", "10-k", "form 10-k") &&
		containsAnyBrief(message, "简报", "摘要", "总结", "概括", "提炼", "关键信息", "brief", "briefing", "summary", "summarize", "key points")
}

func briefEntityFrame(message string) map[string]any {
	entityName := briefEntityName(message)
	code := briefStockCode(message)
	if entityName == "" && code != "" {
		entityName = code
	}
	if entityName == "" {
		return map[string]any{}
	}
	identifiers := map[string]any{
		"stock_code": "",
		"ticker":     "",
		"exchange":   "",
		"market":     "",
	}
	if code != "" {
		identifiers["stock_code"] = code
	}
	out := map[string]any{"name": entityName}
	out["identifiers"] = identifiers
	return out
}

func briefEntityName(message string) string {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?:找下|查下|查一下|看下|看一下|查询|查找|找到|总结|概括|提炼)([^，。,.；;：:\n]+?)(?:的?\s*(?:最新|20\d{2}\s*年?|[0-9]{2}\s*年)?\s*(?:财报|年报|年度报告|财务报告)|annual report|financial report)`),
		regexp.MustCompile(`([^，。,.；;：:\n]+?)(?:的?\s*(?:最新|20\d{2}\s*年?|[0-9]{2}\s*年)?\s*)?(?:财报|年报|年度报告|财务报告)`),
	}
	for _, pattern := range patterns {
		if match := pattern.FindStringSubmatch(message); len(match) > 1 {
			if cleaned := cleanBriefEntityName(match[1]); cleaned != "" {
				return cleaned
			}
		}
	}
	return ""
}

func cleanBriefEntityName(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ""
	}
	value = regexp.MustCompile(`(?i)[(（]\s*(?:sh|sz|bj|hk)?\s*\d{5,6}\s*[)）]`).ReplaceAllString(value, "")
	value = regexp.MustCompile(`(?i)\b(?:sh|sz|bj|hk)?\s*\d{5,6}\b`).ReplaceAllString(value, "")
	replacers := []string{
		"请", "",
		"帮我", "",
		"去", "",
		"到", "",
		"上", "",
		"从", "",
		"在", "",
		"查一下", "",
		"查下", "",
		"查询", "",
		"查", "",
		"看一下", "",
		"看下", "",
		"最新", "",
		"公司", "",
		"官方", "",
		"官网", "",
		"巨潮资讯", "",
		"巨潮", "",
		"CNINFO", "",
		"cninfo", "",
		"SEC", "",
		"sec", "",
		"20-F", "",
		"20-f", "",
		"Form", "",
		"form", "",
		"10-K", "",
		"10-k", "",
		"的", "",
	}
	for i := 0; i < len(replacers); i += 2 {
		value = strings.ReplaceAll(value, replacers[i], replacers[i+1])
	}
	value = strings.TrimFunc(value, func(r rune) bool {
		return unicode.IsSpace(r) || strings.ContainsRune("，。,.；;：:()（）[]【】<>《》\"'", r)
	})
	switch value {
	case "", "某家", "某个", "某", "他", "它", "其", "该", "该公司", "这家公司":
		return ""
	}
	return value
}

func briefStockCode(message string) string {
	match := briefStockCodePattern.FindStringSubmatch(message)
	if len(match) < 2 {
		return ""
	}
	return strings.TrimSpace(match[1])
}

func briefPeriodPolicy(message string) string {
	if containsAnyBrief(message, "近几年", "最近几年", "过去几年", "趋势", "multi-year", "recent years") {
		return "recent_years"
	}
	if year := regexp.MustCompile(`20\d{2}`).FindString(message); year != "" {
		return "explicit_year:" + year
	}
	if match := regexp.MustCompile(`(?:^|[^\d])(\d{2})\s*年`).FindStringSubmatch(message); len(match) > 1 {
		return "explicit_year:20" + match[1]
	}
	return "latest_annual"
}

func briefOutputStyle(message string) string {
	if containsAnyBrief(message, "一段", "单段", "paragraph", "one paragraph") {
		return "single_paragraph_brief"
	}
	return "concise_brief"
}

func containsAnyBrief(message string, needles ...string) bool {
	lower := strings.ToLower(message)
	for _, needle := range needles {
		needle = strings.TrimSpace(needle)
		if needle == "" {
			continue
		}
		if strings.Contains(lower, strings.ToLower(needle)) {
			return true
		}
	}
	return false
}
