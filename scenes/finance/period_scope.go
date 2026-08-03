package finance

import (
	"regexp"
	"strings"
)

const MetricPeriodScopeReviewField = "period_scope"

type MetricsPeriodScopeRequest struct {
	PeriodScope    string
	ReportKind     string
	TrendRequested bool
	Spec           MetricSpec
}

// MetricsPeriodScopeReviewFields verifies that grounded metric evidence matches
// the structured period/report scope supplied by the model. It intentionally
// consumes tool schema values, not raw natural language.
func MetricsPeriodScopeReviewFields(request MetricsPeriodScopeRequest, evidence MetricsEvidence) []string {
	if metricsPeriodScopeRequiresQuarter(request) {
		if !MetricReportPeriodLooksQuarter(evidence.ReportPeriod) {
			return []string{MetricPeriodScopeReviewField}
		}
		if metricsPeriodScopeRequiresQuarterSeries(request) && !MetricQuarterTrendSeriesReady(request.Spec, evidence.TrendSeries) {
			return []string{MetricPeriodScopeReviewField}
		}
		return nil
	}
	if metricsPeriodScopeRequiresAnnual(request) {
		if !MetricReportPeriodLooksAnnual(evidence.ReportPeriod) {
			return []string{MetricPeriodScopeReviewField}
		}
		if metricsPeriodScopeRequiresAnnualSeries(request) && !TrendSeriesReady(request.Spec, evidence.TrendSeries) {
			return []string{MetricPeriodScopeReviewField}
		}
		return nil
	}
	if metricsPeriodScopeRequiresLatestDisclosedReport(request) {
		if !MetricReportPeriodLooksAnnual(evidence.ReportPeriod) && !MetricReportPeriodLooksQuarter(evidence.ReportPeriod) {
			return []string{MetricPeriodScopeReviewField}
		}
	}
	return nil
}

func metricsPeriodScopeRequiresQuarter(request MetricsPeriodScopeRequest) bool {
	scope := normalizePeriodScopeToken(request.PeriodScope)
	reportKind := normalizePeriodScopeToken(request.ReportKind)
	switch scope {
	case "latest_quarter", "latest_quarterly", "latest_disclosed_quarter", "latest_disclosed_quarters", "recent_quarter", "recent_quarters", "recent_disclosed_quarters", "multi_quarter", "multi_quarters", "quarterly_results":
		return true
	}
	switch reportKind {
	case "quarterly_results", "quarterly_report":
		return true
	}
	return false
}

func metricsPeriodScopeRequiresQuarterSeries(request MetricsPeriodScopeRequest) bool {
	scope := normalizePeriodScopeToken(request.PeriodScope)
	switch scope {
	case "latest_disclosed_quarters", "recent_quarters", "recent_disclosed_quarters", "multi_quarter", "multi_quarters":
		return true
	}
	return request.TrendRequested && metricsPeriodScopeRequiresQuarter(request)
}

func metricsPeriodScopeRequiresAnnual(request MetricsPeriodScopeRequest) bool {
	scope := normalizePeriodScopeToken(request.PeriodScope)
	reportKind := normalizePeriodScopeToken(request.ReportKind)
	switch scope {
	case "latest_annual", "annual", "annual_report", "latest_annual_report", "latest_disclosed_annual", "latest_disclosed_annual_report", "latest_disclosed_annual_three_years", "recent_years", "recent_annual_years", "multi_year", "multi_years":
		return true
	}
	return reportKind == "annual_report"
}

func metricsPeriodScopeRequiresAnnualSeries(request MetricsPeriodScopeRequest) bool {
	scope := normalizePeriodScopeToken(request.PeriodScope)
	switch scope {
	case "latest_disclosed_annual_three_years", "recent_years", "recent_annual_years", "multi_year", "multi_years":
		return true
	}
	return request.TrendRequested && metricsPeriodScopeRequiresAnnual(request)
}

func metricsPeriodScopeRequiresLatestDisclosedReport(request MetricsPeriodScopeRequest) bool {
	scope := normalizePeriodScopeToken(request.PeriodScope)
	switch scope {
	case "latest_disclosed_report", "latest_published_report", "latest_available_report", "latest_report":
		return true
	default:
		return false
	}
}

func normalizePeriodScopeToken(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, "-", "_")
	value = strings.ReplaceAll(value, " ", "_")
	return value
}

func MetricReportPeriodLooksAnnual(period string) bool {
	period = strings.ToLower(strings.TrimSpace(period))
	if period == "" || period == "unknown" {
		return false
	}
	if strings.Contains(period, "单季报") ||
		strings.Contains(period, "一季") ||
		strings.Contains(period, "半年度") ||
		strings.Contains(period, "中报") ||
		strings.Contains(period, "三季") ||
		strings.Contains(period, "quarter") ||
		strings.Contains(period, "interim") ||
		strings.Contains(period, "10-q") ||
		strings.Contains(period, "6-k") ||
		regexp.MustCompile(`\bq[1-4]\b`).MatchString(period) {
		return false
	}
	if strings.Contains(period, "annual") || strings.Contains(period, "年报") {
		return true
	}
	if strings.Contains(period, "10-k") || strings.Contains(period, "20-f") {
		return true
	}
	return regexp.MustCompile(`(?:^|[^\d])(?:19|20)\d{2}-12-31(?:[^\d]|$)`).MatchString(period)
}

func MetricReportPeriodLooksQuarter(period string) bool {
	period = strings.ToLower(strings.TrimSpace(period))
	if period == "" || period == "unknown" {
		return false
	}
	if MetricReportPeriodLooksAnnual(period) {
		return false
	}
	if regexp.MustCompile(`\bq[1-4]\b`).MatchString(period) ||
		strings.Contains(period, "10-q") ||
		strings.Contains(period, "6-k") ||
		strings.Contains(period, "interim") ||
		strings.Contains(period, "单季报") ||
		strings.Contains(period, "一季") ||
		strings.Contains(period, "半年度") ||
		strings.Contains(period, "中报") ||
		strings.Contains(period, "三季") ||
		strings.Contains(period, "quarter") {
		return true
	}
	return regexp.MustCompile(`(?:^|[^\d])(?:19|20)\d{2}-(?:03-31|06-30|09-30)(?:[^\d]|$)`).MatchString(period)
}

func MetricQuarterTrendSeriesReady(spec MetricSpec, series []MetricsTrendSeriesPoint) bool {
	return QuarterTrendSeriesPeriodCount(spec, series) >= 2
}

// QuarterTrendSeriesPeriodCount counts ready, distinct quarterly/interim periods.
// It intentionally does not de-duplicate by calendar year.
func QuarterTrendSeriesPeriodCount(spec MetricSpec, series []MetricsTrendSeriesPoint) int {
	count := 0
	seen := map[string]bool{}
	for _, point := range series {
		if !TrendSeriesPointReady(spec, point) || !MetricReportPeriodLooksQuarter(point.Period) {
			continue
		}
		key := strings.TrimSpace(point.Period)
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		count++
	}
	return count
}
