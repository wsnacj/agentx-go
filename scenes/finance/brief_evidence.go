package finance

import (
	"fmt"
	"strings"
	"unicode/utf8"

	financialreportbrief "github.com/wsnacj/agentx-go/scenes/finance/brief"
)

func NormalizeBriefEvidence(evidence financialreportbrief.BriefEvidence, defaultParserID string) financialreportbrief.BriefEvidence {
	evidence.CompanyName = firstNonEmpty(evidence.CompanyName, "unknown")
	evidence.StockCode = firstNonEmpty(evidence.StockCode, "unknown")
	evidence.ReportPeriod = firstNonEmpty(evidence.ReportPeriod, "unknown")
	evidence.SourceURL = firstNonEmpty(evidence.SourceURL, "unknown")
	evidence.ParserID = firstNonEmpty(evidence.ParserID, defaultParserID)
	evidence.ReviewReasons = nonNilStringSlice(evidence.ReviewReasons)
	evidence.ExtractWarnings = nonNilStringSlice(evidence.ExtractWarnings)
	return evidence
}

func ReviewBriefEvidence(evidence financialreportbrief.BriefEvidence) financialreportbrief.BriefEvidence {
	reasons := append([]string{}, evidence.ReviewReasons...)
	if strings.TrimSpace(evidence.CompanyName) == "" || strings.EqualFold(evidence.CompanyName, "unknown") {
		reasons = appendUniqueString(reasons, "company_name_missing")
	}
	if strings.TrimSpace(evidence.ReportPeriod) == "" || strings.EqualFold(evidence.ReportPeriod, "unknown") {
		reasons = appendUniqueString(reasons, "report_period_missing")
	}
	if strings.TrimSpace(evidence.SourceURL) == "" || strings.EqualFold(evidence.SourceURL, "unknown") {
		reasons = appendUniqueString(reasons, "source_url_missing")
	}
	if strings.TrimSpace(evidence.Brief) == "" {
		reasons = appendUniqueString(reasons, "brief_missing")
	}
	if len(evidence.KeyPoints) < 3 {
		reasons = appendUniqueString(reasons, "key_points_incomplete")
	}
	if !BriefHasCategory(evidence.KeyPoints, "financial") && len(evidence.Metrics) < 2 {
		reasons = appendUniqueString(reasons, "financials_missing")
	}
	if !BriefHasCategory(evidence.KeyPoints, "risk") && !BriefHasCategory(evidence.KeyPoints, "outlook") {
		reasons = appendUniqueString(reasons, "risk_or_outlook_missing")
	}
	for _, point := range evidence.KeyPoints {
		if point.ReviewRequired {
			evidence.ExtractWarnings = appendUniqueString(evidence.ExtractWarnings, "key_point_review_required:"+firstNonEmpty(point.Category, "unknown"))
		}
	}
	for _, metric := range evidence.Metrics {
		if metric.ReviewRequired {
			evidence.ExtractWarnings = appendUniqueString(evidence.ExtractWarnings, "metric_review_required:"+firstNonEmpty(metric.Name, "unknown"))
		}
	}
	evidence.ReviewReasons = reasons
	evidence.ReviewRequired = len(reasons) > 0
	return evidence
}

func BriefHasSignal(evidence financialreportbrief.BriefEvidence) bool {
	return strings.TrimSpace(evidence.Brief) != "" ||
		len(evidence.KeyPoints) > 0 ||
		len(evidence.Metrics) > 0
}

func BuildBriefText(evidence financialreportbrief.BriefEvidence) string {
	company := firstNonEmpty(evidence.CompanyName, "该公司")
	period := firstNonEmpty(evidence.ReportPeriod, "最新报告期")
	parts := []string{}
	skippedCategories := []string{}
	for _, category := range []string{"financial", "business", "management", "risk", "outlook", "shareholder_return", "audit"} {
		for _, point := range evidence.KeyPoints {
			if strings.EqualFold(point.Category, category) && strings.TrimSpace(point.Text) != "" {
				if briefPointUsableInGeneratedText(point) {
					parts = append(parts, point.Title+"："+point.Text)
				} else {
					skippedCategories = appendUniqueString(skippedCategories, briefPointCategoryLabel(category))
				}
				break
			}
		}
	}
	if len(skippedCategories) > 0 {
		parts = append(parts, "补充说明："+strings.Join(skippedCategories, "/")+"已在报告证据中定位，但自动摘录需要翻译或复核，摘要正文未直接引用")
	}
	if len(parts) == 0 {
		return ""
	}
	return trimRunes(fmt.Sprintf("%s %s %s：%s。", company, period, briefDocumentLabel(period), strings.Join(parts, "；")), 900)
}

func briefDocumentLabel(reportPeriod string) string {
	normalized := strings.ToUpper(strings.TrimSpace(reportPeriod))
	switch {
	case strings.Contains(normalized, "10-Q"):
		return "季报简报"
	case strings.Contains(normalized, "6-K"):
		return "财报简报"
	default:
		return "年报简报"
	}
}

func briefPointUsableInGeneratedText(point financialreportbrief.BriefKeyPoint) bool {
	text := compactWhitespace(point.Text)
	if text == "" {
		return false
	}
	category := strings.ToLower(strings.TrimSpace(point.Category))
	if category == "financial" {
		return true
	}
	if point.ReviewRequired {
		return false
	}
	if briefLooksMostlyLatin(text) {
		return false
	}
	return true
}

func briefPointCategoryLabel(category string) string {
	switch strings.ToLower(strings.TrimSpace(category)) {
	case "business":
		return "业务概况"
	case "management":
		return "经营讨论"
	case "risk":
		return "风险信息"
	case "outlook":
		return "未来展望"
	case "shareholder_return":
		return "股东回报"
	case "audit":
		return "审计信息"
	default:
		return "部分关键点"
	}
}

func briefLooksMostlyLatin(value string) bool {
	letters := 0
	latin := 0
	cjk := 0
	for _, r := range value {
		switch {
		case r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z':
			letters++
			latin++
		case r >= '\u4e00' && r <= '\u9fff':
			letters++
			cjk++
		}
	}
	if letters < 24 {
		return false
	}
	return latin >= 24 && latin > cjk*2
}

func ReplaceBriefFinancialKeyPoint(evidence financialreportbrief.BriefEvidence) financialreportbrief.BriefEvidence {
	financial := BriefFinancialKeyPoint(evidence)
	if strings.TrimSpace(financial.Text) == "" {
		return evidence
	}
	replaced := false
	points := make([]financialreportbrief.BriefKeyPoint, 0, len(evidence.KeyPoints)+1)
	for _, point := range evidence.KeyPoints {
		if strings.EqualFold(strings.TrimSpace(point.Category), "financial") {
			if !replaced {
				points = append(points, financial)
				replaced = true
			}
			continue
		}
		points = append(points, point)
	}
	if !replaced {
		points = append([]financialreportbrief.BriefKeyPoint{financial}, points...)
	}
	evidence.KeyPoints = points
	evidence.SourceChapters = BriefSourceChapters(evidence.KeyPoints)
	return evidence
}

func BriefFinancialKeyPoint(evidence financialreportbrief.BriefEvidence) financialreportbrief.BriefKeyPoint {
	parts := []string{}
	sources := []string{}
	for _, metric := range evidence.Metrics {
		if !BriefMetricValueUsableForName(metric.Name, metric.Value) {
			continue
		}
		label := briefMetricLabel(metric)
		switch metric.Name {
		case "revenue":
			parts = append(parts, label+" "+metric.Value)
		case "revenue_growth":
			parts = append(parts, label+" "+metric.Value)
		case "net_profit":
			parts = append(parts, label+" "+metric.Value)
		case "net_profit_growth":
			parts = append(parts, label+" "+metric.Value)
		case "operating_cash_flow":
			parts = append(parts, label+" "+metric.Value)
		}
		if strings.TrimSpace(metric.Source) != "" {
			sources = appendUniqueString(sources, metric.Source)
		}
	}
	if len(parts) == 0 {
		return financialreportbrief.BriefKeyPoint{}
	}
	return financialreportbrief.BriefKeyPoint{
		Category:   "financial",
		Title:      "核心财务表现",
		Text:       strings.Join(parts, "；"),
		Source:     firstNonEmpty(strings.Join(sources, ","), "financial_fields"),
		Confidence: 0.9,
	}
}

func briefMetricLabel(metric financialreportbrief.BriefMetric) string {
	label := ""
	switch strings.ToLower(strings.TrimSpace(metric.Name)) {
	case "revenue":
		label = "收入"
	case "revenue_growth":
		label = "收入增长"
	case "net_profit":
		label = "利润"
	case "net_profit_growth":
		label = "利润增长"
	case "operating_cash_flow":
		label = "经营现金流"
	}
	if period := strings.TrimSpace(metric.Period); label != "" && period != "" {
		return label + "（" + period + "）"
	}
	return label
}

func BriefMetricValue(metrics []financialreportbrief.BriefMetric, name string) string {
	for _, metric := range metrics {
		if strings.EqualFold(strings.TrimSpace(metric.Name), strings.TrimSpace(name)) {
			return strings.TrimSpace(metric.Value)
		}
	}
	return ""
}

func BriefUpsertMetric(metrics []financialreportbrief.BriefMetric, name string, value string, unit string, source string, confidence float64) []financialreportbrief.BriefMetric {
	name = strings.TrimSpace(name)
	value = strings.TrimSpace(value)
	if name == "" || value == "" {
		return metrics
	}
	for index := range metrics {
		if !strings.EqualFold(strings.TrimSpace(metrics[index].Name), name) {
			continue
		}
		metrics[index].Value = value
		metrics[index].Unit = firstNonEmpty(unit, metrics[index].Unit)
		metrics[index].Source = firstNonEmpty(source, metrics[index].Source)
		if confidence > 0 && metrics[index].Confidence < confidence {
			metrics[index].Confidence = confidence
		}
		metrics[index].ReviewRequired = false
		return metrics
	}
	return append(metrics, financialreportbrief.BriefMetric{
		Name:       name,
		Value:      value,
		Unit:       unit,
		Source:     source,
		Confidence: confidence,
	})
}

func BriefRemoveExtractWarning(warnings []string, target string) []string {
	target = strings.TrimSpace(target)
	if target == "" {
		return warnings
	}
	out := make([]string, 0, len(warnings))
	for _, warning := range warnings {
		if strings.TrimSpace(warning) == target {
			continue
		}
		out = append(out, warning)
	}
	return out
}

func BriefSourceChapters(points []financialreportbrief.BriefKeyPoint) []string {
	out := []string{}
	for _, point := range points {
		if strings.TrimSpace(point.Source) != "" {
			out = appendUniqueString(out, strings.TrimSpace(point.Source))
		}
	}
	return out
}

func BriefHasCategory(points []financialreportbrief.BriefKeyPoint, category string) bool {
	for _, point := range points {
		if strings.EqualFold(strings.TrimSpace(point.Category), strings.TrimSpace(category)) && strings.TrimSpace(point.Text) != "" {
			return true
		}
	}
	return false
}

func BriefMetricValueUsable(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return false
	}
	normalized := strings.ToLower(compactWhitespace(value))
	switch normalized {
	case "unknown", "n/a", "na", "nil", "null", "-", "--", "not available", "not applicable":
		return false
	}
	return true
}

func BriefMetricValueUsableForName(name string, value string) bool {
	if !BriefMetricValueUsable(value) {
		return false
	}
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "revenue_growth" || name == "net_profit_growth" {
		return strings.Contains(value, "%") || hasDigit(value)
	}
	return true
}

func compactWhitespace(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}

func trimRunes(value string, limit int) string {
	value = strings.TrimSpace(value)
	if limit <= 0 || utf8.RuneCountInString(value) <= limit {
		return value
	}
	runes := []rune(value)
	return strings.TrimSpace(string(runes[:limit]))
}

func hasDigit(value string) bool {
	for _, r := range value {
		if r >= '0' && r <= '9' {
			return true
		}
	}
	return false
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func appendUniqueString(items []string, values ...string) []string {
	seen := map[string]bool{}
	for _, item := range items {
		seen[item] = true
	}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		items = append(items, value)
	}
	return items
}

func nonNilStringSlice(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}
