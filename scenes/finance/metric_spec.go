package finance

import (
	"regexp"
	"strconv"
	"strings"

	financialreportmetrics "github.com/wsnacj/agentx-go/scenes/finance/metrics"
)

// MetricSpec describes the canonical financial metrics requested by a
// report-metrics task. It is domain-level, not tied to any site or project.
type MetricSpec struct {
	Revenue           bool
	RevenueGrowth     bool
	NetProfit         bool
	NetProfitGrowth   bool
	OperatingCashFlow bool
}

// MetricSpecFromFields converts arbitrary metric labels into a canonical
// MetricSpec.
func MetricSpecFromFields(fields []string) MetricSpec {
	spec := MetricSpec{}
	for _, field := range NormalizeMetricFields(fields) {
		switch field {
		case "revenue":
			spec.Revenue = true
		case "revenue_growth":
			spec.RevenueGrowth = true
		case "net_profit":
			spec.NetProfit = true
		case "net_profit_growth":
			spec.NetProfitGrowth = true
		case "operating_cash_flow":
			spec.OperatingCashFlow = true
		}
	}
	return spec
}

// NormalizeMetricFields normalizes and de-duplicates metric labels while
// preserving first-seen order.
func NormalizeMetricFields(fields []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, field := range fields {
		metric := NormalizeMetricName(field)
		if metric == "" || seen[metric] {
			continue
		}
		seen[metric] = true
		out = append(out, metric)
	}
	return out
}

// NormalizeMetricName maps common bilingual financial metric labels to the
// canonical report-metrics field names.
func NormalizeMetricName(field string) string {
	field = strings.ToLower(strings.TrimSpace(field))
	field = strings.ReplaceAll(field, "-", "_")
	field = strings.ReplaceAll(field, " ", "_")
	switch field {
	case "revenue", "income", "sales", "turnover", "operating_revenue", "total_revenue", "营业收入", "营业总收入", "营收", "收入":
		return "revenue"
	case "revenue_growth", "income_growth", "sales_growth", "revenue_yoy", "yoy_revenue", "营收增长", "收入增长", "营业收入增长":
		return "revenue_growth"
	case "net_profit", "profit", "net_income", "profit_attributable_to_owners", "归母净利润", "净利润", "利润":
		return "net_profit"
	case "net_profit_growth", "profit_growth", "net_income_growth", "profit_yoy", "net_profit_yoy", "净利润增长", "利润增长":
		return "net_profit_growth"
	case "operating_cash_flow", "operating_cashflow", "cash_flow_from_operations", "经营现金流", "经营活动现金流":
		return "operating_cash_flow"
	default:
		return ""
	}
}

func (s MetricSpec) Any() bool {
	return s.Revenue || s.RevenueGrowth || s.NetProfit || s.NetProfitGrowth || s.OperatingCashFlow
}

func (s MetricSpec) Fields() []string {
	out := make([]string, 0, 5)
	if s.Revenue {
		out = append(out, "revenue")
	}
	if s.RevenueGrowth {
		out = append(out, "revenue_growth")
	}
	if s.NetProfit {
		out = append(out, "net_profit")
	}
	if s.NetProfitGrowth {
		out = append(out, "net_profit_growth")
	}
	if s.OperatingCashFlow {
		out = append(out, "operating_cash_flow")
	}
	return out
}

func (s MetricSpec) MissingFields(evidence MetricsEvidence) []string {
	missing := make([]string, 0, 5)
	if s.Revenue && metricValueMissing(evidence.Revenue) {
		missing = append(missing, "revenue")
	}
	if s.RevenueGrowth && metricValueMissing(evidence.RevenueGrowth) {
		missing = append(missing, "revenue_growth")
	}
	if s.NetProfit && metricValueMissing(evidence.NetProfit) {
		missing = append(missing, "net_profit")
	}
	if s.NetProfitGrowth && metricValueMissing(evidence.NetProfitGrowth) {
		missing = append(missing, "net_profit_growth")
	}
	if s.OperatingCashFlow && metricValueMissing(evidence.OperatingCashFlow) {
		missing = append(missing, "operating_cash_flow")
	}
	return missing
}

func (s MetricSpec) Ready(evidence MetricsEvidence) bool {
	return len(s.MissingFields(evidence)) == 0 && len(MetricReviewRequiredFields(s, evidence)) == 0
}

// MetricReviewRequiredFields returns requested fields that are present but not
// yet strong enough for a final answer.
func MetricReviewRequiredFields(spec MetricSpec, evidence MetricsEvidence) []string {
	fields := make([]string, 0, 5)
	if spec.Revenue && (MetricMonetaryValueNeedsReview(evidence.Revenue) || MetricFieldEvidenceNeedsReview(evidence, "revenue")) {
		fields = append(fields, "revenue")
	}
	if spec.RevenueGrowth && (MetricValueNeedsReview(evidence.RevenueGrowth) || MetricFieldEvidenceNeedsReview(evidence, "revenue_growth")) {
		fields = append(fields, "revenue_growth")
	}
	if spec.NetProfit && (MetricMonetaryValueNeedsReview(evidence.NetProfit) || MetricFieldEvidenceNeedsReview(evidence, "net_profit")) {
		fields = append(fields, "net_profit")
	}
	if spec.NetProfitGrowth && (MetricValueNeedsReview(evidence.NetProfitGrowth) || MetricFieldEvidenceNeedsReview(evidence, "net_profit_growth")) {
		fields = append(fields, "net_profit_growth")
	}
	if spec.OperatingCashFlow && (MetricMonetaryValueNeedsReview(evidence.OperatingCashFlow) || MetricFieldEvidenceNeedsReview(evidence, "operating_cash_flow")) {
		fields = append(fields, "operating_cash_flow")
	}
	return fields
}

func MetricFieldEvidenceNeedsReview(evidence MetricsEvidence, field string) bool {
	if len(evidence.MetricEvidence) == 0 {
		return false
	}
	fieldEvidence, ok := evidence.MetricEvidence[field]
	if !ok {
		return false
	}
	return fieldEvidence.ReviewRequired || MetricFieldContractNeedsReview(field, fieldEvidence)
}

func MetricFieldContractNeedsReview(field string, evidence financialreportmetrics.ReportDocumentMetricFieldEvidence) bool {
	if metricValueMissing(evidence.Value) {
		return true
	}
	if metricValueMissing(evidence.Source) {
		return true
	}
	if metricValueMissing(evidence.Period) {
		return true
	}
	valueForReview := MetricFieldValueForReview(evidence)
	switch field {
	case "revenue", "net_profit", "operating_cash_flow":
		return MetricMonetaryValueNeedsReview(valueForReview)
	default:
		return MetricValueNeedsReview(valueForReview)
	}
}

func MetricFieldValueForReview(evidence financialreportmetrics.ReportDocumentMetricFieldEvidence) string {
	parts := []string{strings.TrimSpace(evidence.Value)}
	for _, value := range []string{evidence.Currency, evidence.Unit} {
		value = strings.TrimSpace(value)
		if value == "" || strings.Contains(strings.ToLower(strings.Join(parts, " ")), strings.ToLower(value)) {
			continue
		}
		parts = append(parts, value)
	}
	return strings.TrimSpace(strings.Join(parts, " "))
}

func MetricValueNeedsReview(value string) bool {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return true
	}
	lower := strings.ToLower(trimmed)
	switch lower {
	case "-", "--", "—", "n.a.", "na", "nil", "none", "null":
		return true
	}
	return containsAnyFoldString(trimmed,
		"unknown",
		"not_confirmed",
		"not confirmed",
		"not-confirmed",
		"n/a",
		"tbd",
		"pending",
		"unconfirmed",
		"未确认",
		"待确认",
		"待核实",
		"未核实",
		"无法确认",
		"不确定",
		"暂无",
		"未披露",
		"待披露",
		"估算",
		"估计",
		"estimated",
		"estimate",
		"approx",
		"approximately",
		"about",
		"约",
		"大约",
		"左右",
		"每股",
		"per share",
		"per-share",
		"eps",
	)
}

func MetricMonetaryValueNeedsReview(value string) bool {
	if MetricValueNeedsReview(value) {
		return true
	}
	trimmed := strings.TrimSpace(value)
	if strings.Contains(trimmed, "%") {
		return true
	}
	if MonetaryValueMentionsUnit(trimmed) && !MonetaryValueHasCurrency(trimmed) {
		return true
	}
	return MonetaryValueLooksBare(trimmed)
}

func MonetaryValueMentionsUnit(value string) bool {
	return containsAnyFoldString(value,
		"million",
		"billion",
		"thousand",
		"百万元",
		"千元",
		"万元",
		"亿元",
		"百万",
		"十亿",
	)
}

func MonetaryValueHasCurrency(value string) bool {
	return containsAnyFoldString(value,
		"rmb",
		"cny",
		"hkd",
		"hk$",
		"usd",
		"us$",
		"renminbi",
		"yuan",
		"dollar",
		"人民币",
		"港币",
		"美元",
		"元",
	)
}

func MonetaryValueLooksBare(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return false
	}
	if MonetaryValueHasCurrency(value) || MonetaryValueMentionsUnit(value) {
		return false
	}
	cleaned := regexp.MustCompile(`[\s,，._()（）+-]+`).ReplaceAllString(value, "")
	return cleaned != "" && regexp.MustCompile(`^\d+$`).MatchString(cleaned)
}

func TrendSeriesReady(spec MetricSpec, series []MetricsTrendSeriesPoint) bool {
	return TrendSeriesPeriodCount(spec, series) >= 2 && TrendSeriesContinuous(spec, series)
}

func TrendSeriesPeriodCount(spec MetricSpec, series []MetricsTrendSeriesPoint) int {
	years := map[int]bool{}
	for _, point := range series {
		if !TrendSeriesPointReady(spec, point) {
			continue
		}
		year := TrendSeriesPointYear(point)
		if year <= 0 {
			continue
		}
		years[year] = true
	}
	return len(years)
}

func TrendSeriesContinuous(spec MetricSpec, series []MetricsTrendSeriesPoint) bool {
	years := map[int]bool{}
	minYear := 0
	maxYear := 0
	for _, point := range series {
		if !TrendSeriesPointReady(spec, point) {
			continue
		}
		year := TrendSeriesPointYear(point)
		if year <= 0 {
			continue
		}
		years[year] = true
		if minYear == 0 || year < minYear {
			minYear = year
		}
		if year > maxYear {
			maxYear = year
		}
	}
	if len(years) < 2 {
		return false
	}
	return maxYear-minYear+1 == len(years)
}

func TrendSeriesPointReady(spec MetricSpec, point MetricsTrendSeriesPoint) bool {
	if point.ReviewRequired {
		return false
	}
	if TrendSeriesPointYear(point) <= 0 {
		return false
	}
	if !MetricFieldSourceLooksConcrete(point.Source) {
		return false
	}
	if spec.Revenue && MetricMonetaryValueNeedsReview(point.Revenue) {
		return false
	}
	if spec.NetProfit && MetricMonetaryValueNeedsReview(point.NetProfit) {
		return false
	}
	if spec.OperatingCashFlow && MetricMonetaryValueNeedsReview(point.OperatingCashFlow) {
		return false
	}
	if spec.RevenueGrowth && MetricValueNeedsReview(point.RevenueGrowth) {
		return false
	}
	if spec.NetProfitGrowth && MetricValueNeedsReview(point.NetProfitGrowth) {
		return false
	}
	return spec.Any()
}

func TrendSeriesPointYear(point MetricsTrendSeriesPoint) int {
	match := regexp.MustCompile(`\b(20\d{2}|19\d{2})\b`).FindStringSubmatch(point.Period)
	if len(match) < 2 {
		return 0
	}
	year, _ := strconv.Atoi(match[1])
	return year
}

func MetricFieldSourceLooksConcrete(source string) bool {
	source = strings.TrimSpace(source)
	return strings.HasPrefix(strings.ToLower(source), "http://") || strings.HasPrefix(strings.ToLower(source), "https://")
}

func metricValueMissing(value string) bool {
	return strings.TrimSpace(value) == "" || strings.EqualFold(strings.TrimSpace(value), "unknown")
}
