package finance

import (
	"regexp"
	"strings"
)

// StructuredTextMetricsInput carries adapter-neutral text and context for
// extracting financial metrics from structured report parser output.
type StructuredTextMetricsInput struct {
	Text              string
	Title             string
	SourceURL         string
	EntityName        string
	StockCode         string
	Ticker            string
	UserMessage       string
	RequestedMetrics  []string
	SelectionReason   string
	PageTitleFallback string
}

// ExtractStructuredTextMetricsEvidence extracts metrics from adapter-neutral
// KEY=VALUE text, such as text produced by an official report/PDF parser. It
// intentionally does not know about any website or exchange.
func ExtractStructuredTextMetricsEvidence(input StructuredTextMetricsInput) (MetricsEvidence, bool) {
	fields := ParseStructuredMetricFields(input.Text)
	if len(fields) == 0 {
		return MetricsEvidence{}, false
	}
	evidence := MetricsEvidence{
		CompanyName:       structuredTextCompany(fields, input),
		StockCode:         structuredTextStockCode(fields, input),
		SelectionReason:   firstNonEmptyString(input.SelectionReason, "selected_by_structured_text_metrics_adapter"),
		OfficialSource:    structuredTextOfficialSource(fields, input),
		ReportPeriod:      structuredTextReportPeriod(fields["YEAR"], input),
		Revenue:           structuredTextMetricValue(fields["REVENUE"]),
		RevenueGrowth:     structuredTextMetricValue(fields["REVENUE_GROWTH"]),
		NetProfit:         structuredTextMetricValue(fields["NET_PROFIT"]),
		NetProfitGrowth:   structuredTextMetricValue(fields["NET_PROFIT_GROWTH"]),
		OperatingCashFlow: structuredTextMetricValue(fields["OPERATING_CASH_FLOW"]),
		PageTitle:         firstNonEmptyString(input.Title, input.PageTitleFallback, "structured_text_metrics"),
	}
	if !StructuredTextMetricsHasSignal(evidence, input.RequestedMetrics) {
		return MetricsEvidence{}, false
	}
	return evidence, true
}

// ParseStructuredMetricFields returns upper-case keys from simple KEY=VALUE
// metric text. It accepts semicolon or newline separators and keeps values raw.
func ParseStructuredMetricFields(text string) map[string]string {
	out := map[string]string{}
	text = strings.TrimSpace(strings.ReplaceAll(text, "；", ";"))
	if text == "" {
		return out
	}
	pattern := regexp.MustCompile(`(?im)\b([A-Z][A-Z0-9_]{1,40})\s*=\s*([^;\n\r]+)`)
	for _, match := range pattern.FindAllStringSubmatch(text, -1) {
		key := strings.ToUpper(strings.TrimSpace(match[1]))
		value := strings.TrimSpace(match[2])
		if key == "" || value == "" {
			continue
		}
		out[key] = value
	}
	return out
}

// StructuredTextMetricsHasSignal reports whether evidence contains at least one
// usable requested metric. With no requested metrics it falls back to revenue or
// net profit, matching the public-report short-query default.
func StructuredTextMetricsHasSignal(evidence MetricsEvidence, requestedMetrics []string) bool {
	fields := NormalizeMetricFields(requestedMetrics)
	if len(fields) == 0 {
		return structuredTextMetricUsable(evidence.Revenue) || structuredTextMetricUsable(evidence.NetProfit)
	}
	for _, field := range fields {
		switch field {
		case "revenue":
			if structuredTextMetricUsable(evidence.Revenue) {
				return true
			}
		case "revenue_growth":
			if structuredTextMetricUsable(evidence.RevenueGrowth) {
				return true
			}
		case "net_profit":
			if structuredTextMetricUsable(evidence.NetProfit) {
				return true
			}
		case "net_profit_growth":
			if structuredTextMetricUsable(evidence.NetProfitGrowth) {
				return true
			}
		case "operating_cash_flow":
			if structuredTextMetricUsable(evidence.OperatingCashFlow) {
				return true
			}
		}
	}
	return false
}

func structuredTextCompany(fields map[string]string, input StructuredTextMetricsInput) string {
	return firstNonEmptyString(
		fields["COMPANY_NAME"],
		fields["COMPANY"],
		fields["ISSUER"],
		input.EntityName,
		"unknown",
	)
}

func structuredTextStockCode(fields map[string]string, input StructuredTextMetricsInput) string {
	return firstNonEmptyString(
		fields["STOCK_CODE"],
		fields["TICKER"],
		input.StockCode,
		input.Ticker,
		fields["SECURITY_CODE"],
		"unknown",
	)
}

func structuredTextOfficialSource(fields map[string]string, input StructuredTextMetricsInput) string {
	for _, value := range []string{fields["SOURCE_URL"], input.SourceURL, fields["OFFICIAL_SOURCE"], fields["REPORT_URL"]} {
		value = strings.TrimSpace(value)
		if strings.HasPrefix(strings.ToLower(value), "http://") || strings.HasPrefix(strings.ToLower(value), "https://") {
			return value
		}
	}
	return "unknown"
}

func structuredTextReportPeriod(value string, input StructuredTextMetricsInput) string {
	joined := strings.Join([]string{value, input.UserMessage, input.Title, input.SourceURL, input.Text}, "\n")
	if match := regexp.MustCompile(`(?i)\b(20\d{2})\b`).FindStringSubmatch(joined); len(match) > 1 {
		return strings.TrimSpace(match[1]) + "-12-31"
	}
	return "unknown"
}

func structuredTextMetricValue(value string) string {
	value = strings.TrimSpace(value)
	if !structuredTextMetricUsable(value) {
		return "unknown"
	}
	return value
}

func structuredTextMetricUsable(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return false
	}
	return !containsAnyFoldString(value, "NOT_FOUND", "N/A", "UNKNOWN")
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func containsAnyFoldString(text string, needles ...string) bool {
	lower := strings.ToLower(text)
	for _, needle := range needles {
		if strings.Contains(lower, strings.ToLower(strings.TrimSpace(needle))) {
			return true
		}
	}
	return false
}
