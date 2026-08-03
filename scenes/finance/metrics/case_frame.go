package metrics

import (
	"regexp"
	"strings"
	"unicode"
)

var latestMetricsStockCodePattern = regexp.MustCompile(`(?i)\b(?:sh|sz|bj|hk)?\s*(\d{5,6})\b`)

const (
	RequestedOutputMetrics               = "metrics"
	RequestedOutputPerformanceAssessment = "performance_assessment"
	RequestedOutputInvestmentAssessment  = "investment_assessment"

	AssessmentKindNone                = "none"
	AssessmentKindBusinessPerformance = "business_performance"
	AssessmentKindInvestmentRisk      = "investment_risk"
	AssessmentScopeMetricsOnly        = "financial_report_metrics_only"
)

// BuildLatestMetricsCaseInput is a legacy/explicit pack-workflow fallback for
// materializing case inputs after a host has already selected this pack. It is
// not the default natural-language routing path. Normal AgentX short-query turns
// should use the finance_report_lookup tool schema so the model supplies a
// structured task frame and adapters verify facts from public evidence.
//
// Keep this helper source-neutral and do not expand it with project/company/site
// keyword routing; project/plugin adapters still own source resolution and page
// handling.
func BuildLatestMetricsCaseInput(userMessage string) (map[string]any, bool) {
	message := strings.TrimSpace(userMessage)
	if message == "" {
		return nil, false
	}
	entity := latestMetricsEntityFrame(message)
	name, _ := entity["name"].(string)
	if strings.TrimSpace(name) == "" {
		return nil, false
	}
	metrics := latestMetricsRequestedMetrics(message)
	if len(metrics) == 0 {
		return nil, false
	}
	outputs, assessment := latestMetricsRequestedOutputs(message)
	out := map[string]any{
		"user_message":      message,
		"entity":            entity,
		"requested_metrics": metrics,
		"requested_outputs": outputs,
		"assessment":        assessment,
		"period_policy":     latestMetricsPeriodPolicy(message),
		"source_policy":     "public_web_prefer_official_or_accepted_financial_data_source",
		"freshness":         "live",
		"stop_condition":    "guard_passed",
	}
	return out, true
}

func BuildReportMetricsCaseInput(userMessage string) (map[string]any, string, bool) {
	input, ok := BuildLatestMetricsCaseInput(userMessage)
	if !ok {
		return nil, "", false
	}
	if MetricsTrendRequested(userMessage, stringValue(input["period_policy"])) {
		input["period_policy"] = "recent_years"
		input["stop_condition"] = "guard_passed_or_review_required"
		return input, CaseTypeTrend, true
	}
	return input, CaseTypeLatest, true
}

func MetricsTrendRequested(userMessage string, periodScope string) bool {
	scope := strings.ToLower(strings.TrimSpace(periodScope))
	if containsAnyLatestMetrics(scope, "recent_years", "recent-years", "multi_year", "multi-year", "trend", "series", "historical") {
		return true
	}
	return containsAnyLatestMetrics(userMessage,
		"近几年",
		"最近几年",
		"过去几年",
		"近三年",
		"近两年",
		"多年",
		"历年",
		"逐年",
		"趋势",
		"recent years",
		"last few years",
		"multi-year",
		"multi year",
		"over the years",
		"historical trend",
	)
}

func latestMetricsEntityFrame(message string) map[string]any {
	entityName := latestMetricsEntityName(message)
	code := latestMetricsStockCode(message)
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
	if exchange := latestMetricsExchange(message, code); exchange != "" {
		identifiers["exchange"] = exchange
	}
	out := map[string]any{"name": entityName}
	out["identifiers"] = identifiers
	return out
}

func latestMetricsEntityName(message string) string {
	candidates := []string{}
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?:找下|查下|查一下|看下|看一下|查询|查找|找到)([^，。,.；;：:\n]+?)(?:\d{2,4}\s*年|的?(?:营收|收入|营业总收入|净利润|利润)|\b(?:revenue|net profit|profit)\b)`),
		regexp.MustCompile(`(?:找下|查下|查一下|看下|看一下|查询|查找|找到|看看|看)([^，。,.；;：:\n]+?)(?:的?最新|这只股票|这家公司|这家公司|的?财报|的?年报|的?财务报告)`),
		regexp.MustCompile(`(?:提取|看看|核实|总结)([^，。,.；;：:\n]+?)(?:的?最新|这只股票|这家公司|的?财报|的?年报|的?财务报告)`),
		regexp.MustCompile(`(?i)(?:look up|check|find|query|review|analy[sz]e|assess|evaluate)\s+([a-z][a-z0-9&.,'’ -]{1,80}?)(?:'s|’s|\s+latest|\s+annual|\s+financial|\s+revenue|\s+profit|\s+report)`),
	}
	for _, pattern := range patterns {
		if match := pattern.FindStringSubmatch(message); len(match) > 1 {
			candidates = append(candidates, match[1])
		}
	}
	for _, candidate := range candidates {
		if cleaned := cleanLatestMetricsEntityName(candidate); cleaned != "" {
			return cleaned
		}
	}
	return ""
}

func cleanLatestMetricsEntityName(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ""
	}
	value = regexp.MustCompile(`(?i)[(（]\s*(?:sh|sz|bj|hk)?\s*\d{5,6}\s*[)）]`).ReplaceAllString(value, "")
	value = regexp.MustCompile(`(?i)\b(?:sh|sz|bj|hk)?\s*\d{5,6}\b`).ReplaceAllString(value, "")
	value = regexp.MustCompile(`(?i)\b20\d{2}\s*年?`).ReplaceAllString(value, "")
	value = regexp.MustCompile(`(?i)(^|[^\d])\d{2}\s*年`).ReplaceAllString(value, "$1 ")
	value = regexp.MustCompile(`(?:近|最近|过去)?[一二两三四五六七八九十几\d]*年|多年|历年|逐年`).ReplaceAllString(value, "")
	replacers := []string{
		"请", "",
		"帮我", "",
		"去", "",
		"到", "",
		"上", "",
		"从", "",
		"在", "",
		"东方财富", "",
		"SEC", "",
		"sec", "",
		"EDGAR", "",
		"edgar", "",
		"20-F", "",
		"20-f", "",
		"Form", "",
		"form", "",
		"官网", "",
		"官方", "",
		"这只股票", "",
		"这家公司", "",
		"公司", "",
		"的", "",
	}
	for i := 0; i < len(replacers); i += 2 {
		value = strings.ReplaceAll(value, replacers[i], replacers[i+1])
	}
	value = strings.TrimFunc(value, func(r rune) bool {
		return unicode.IsSpace(r) || strings.ContainsRune("，。,.；;：:()（）[]【】<>《》", r)
	})
	switch value {
	case "", "某家", "某个", "某", "他", "它", "其", "ta", "TA", "该公司", "这家公司", "这只股票":
		return ""
	}
	return value
}

func latestMetricsStockCode(message string) string {
	match := latestMetricsStockCodePattern.FindStringSubmatch(message)
	if len(match) < 2 {
		return ""
	}
	return strings.TrimSpace(match[1])
}

func latestMetricsExchange(message string, code string) string {
	lower := strings.ToLower(message)
	switch {
	case strings.Contains(lower, "sh"+code) || strings.Contains(message, code+".SH"):
		return "SH"
	case strings.Contains(lower, "sz"+code) || strings.Contains(message, code+".SZ"):
		return "SZ"
	case strings.Contains(lower, "bj"+code) || strings.Contains(message, code+".BJ"):
		return "BJ"
	case strings.Contains(lower, "hk"+code) || strings.Contains(message, code+".HK"):
		return "HK"
	default:
		return ""
	}
}

func latestMetricsRequestedMetrics(message string) []any {
	hasRevenue := containsAnyLatestMetrics(message, "营收", "收入", "营业总收入", "revenue")
	hasNetProfit := containsAnyLatestMetrics(message, "净利润", "利润", "net profit", "profit")
	hasCashFlow := containsAnyLatestMetrics(message, "经营现金流", "现金流", "operating cash flow")
	hasGrowth := containsAnyLatestMetrics(message, "增长", "同比", "增速", "growth", "yoy")
	hasRichFields := containsAnyLatestMetrics(message, "尽可能丰富", "丰富字段", "更多字段", "全部指标", "all metrics", "as many metrics")
	hasAssessment := latestMetricsPerformanceAssessmentRequested(message) || latestMetricsInvestmentAssessmentRequested(message)
	if (hasRichFields || hasAssessment) && containsAnyLatestMetrics(message, "财报", "年报", "财务报告", "financial report", "annual report", "20-f", "form 20-f", "10-k", "form 10-k") {
		hasRevenue = true
		hasNetProfit = true
		hasGrowth = true
		hasCashFlow = true
	}
	metrics := []string{}
	if hasRevenue {
		metrics = append(metrics, "revenue")
	}
	if hasGrowth && hasRevenue {
		metrics = append(metrics, "revenue_growth")
	}
	if hasNetProfit {
		metrics = append(metrics, "net_profit")
	}
	if hasGrowth && hasNetProfit {
		metrics = append(metrics, "net_profit_growth")
	}
	if hasCashFlow {
		metrics = append(metrics, "operating_cash_flow")
	}
	if len(metrics) == 0 && containsAnyLatestMetrics(message, "财报", "年报", "财务报告", "financial report", "annual report", "20-f", "form 20-f", "10-k", "form 10-k") {
		metrics = append(metrics, "revenue", "net_profit")
	}
	out := make([]any, 0, len(metrics))
	seen := map[string]bool{}
	for _, metric := range metrics {
		if metric == "" || seen[metric] {
			continue
		}
		seen[metric] = true
		out = append(out, metric)
	}
	return out
}

func latestMetricsRequestedOutputs(message string) ([]any, map[string]any) {
	outputs := []string{RequestedOutputMetrics}
	if latestMetricsPerformanceAssessmentRequested(message) {
		outputs = append(outputs, RequestedOutputPerformanceAssessment)
	}
	if latestMetricsInvestmentAssessmentRequested(message) {
		outputs = append(outputs, RequestedOutputInvestmentAssessment)
	}
	kind := AssessmentKindNone
	requiresValuation := false
	if containsStringLatestMetrics(outputs, RequestedOutputInvestmentAssessment) {
		kind = AssessmentKindInvestmentRisk
		requiresValuation = true
	} else if containsStringLatestMetrics(outputs, RequestedOutputPerformanceAssessment) {
		kind = AssessmentKindBusinessPerformance
	}
	normalized := make([]any, 0, len(outputs))
	seen := map[string]bool{}
	for _, output := range outputs {
		output = strings.TrimSpace(output)
		if output == "" || seen[output] {
			continue
		}
		seen[output] = true
		normalized = append(normalized, output)
	}
	return normalized, map[string]any{
		"kind":               kind,
		"scope":              AssessmentScopeMetricsOnly,
		"requires_valuation": requiresValuation,
	}
}

func latestMetricsPerformanceAssessmentRequested(message string) bool {
	return containsAnyLatestMetrics(message,
		"业绩情况",
		"业绩表现",
		"经营情况",
		"经营表现",
		"表现怎么样",
		"表现如何",
		"业绩如何",
		"业绩怎么样",
		"分析一下",
		"给个评估",
		"给个你的评估",
		"帮我评估",
		"怎么看",
		"评价",
		"performance assessment",
		"business performance",
		"financial performance",
		"assess",
		"evaluate",
		"analysis",
		"what do you think",
	)
}

func latestMetricsInvestmentAssessmentRequested(message string) bool {
	return containsAnyLatestMetrics(message,
		"值得投资",
		"投资价值",
		"是否靠谱",
		"靠不靠谱",
		"能不能投",
		"能否投资",
		"适合投资",
		"靠谱不",
		"投资判断",
		"investment",
		"invest",
	)
}

func containsStringLatestMetrics(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func stringValue(value any) string {
	text, _ := value.(string)
	return text
}

func latestMetricsPeriodPolicy(message string) string {
	if containsAnyLatestMetrics(message, "年报", "annual report", "20-f", "form 20-f", "10-k", "form 10-k") {
		return "latest_disclosed_annual"
	}
	return "latest_disclosed_report"
}

func containsAnyLatestMetrics(message string, needles ...string) bool {
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
