package finance

import (
	"strings"

	financialreportmetrics "github.com/wsnacj/agentx-go/scenes/finance/metrics"
)

type MetricsRequestFrame struct {
	RequestedOutputs            []string
	AssessmentKind              string
	AssessmentScope             string
	AssessmentRequiresValuation bool
}

func NormalizeMetricsRequestedOutputs(values []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, value := range values {
		output := NormalizeMetricsRequestedOutput(value)
		if output == "" || seen[output] {
			continue
		}
		seen[output] = true
		out = append(out, output)
	}
	return out
}

func NormalizeMetricsRequestedOutput(value string) string {
	value = normalizeMetricsRequestToken(value)
	switch value {
	case "metrics", "metric", "requested_metrics", "financial_metrics", "指标", "财报指标":
		return financialreportmetrics.RequestedOutputMetrics
	case "performance_assessment", "performance", "business_performance", "financial_performance", "业绩评估", "经营评估":
		return financialreportmetrics.RequestedOutputPerformanceAssessment
	case "investment_assessment", "investment", "investment_risk", "investment_judgment", "投资评估", "投资判断":
		return financialreportmetrics.RequestedOutputInvestmentAssessment
	default:
		return ""
	}
}

func NormalizeMetricsAssessmentKind(value string) string {
	value = normalizeMetricsRequestToken(value)
	switch value {
	case "", "none", "no_assessment":
		return financialreportmetrics.AssessmentKindNone
	case financialreportmetrics.AssessmentKindBusinessPerformance, "performance_assessment", "performance", "financial_performance":
		return financialreportmetrics.AssessmentKindBusinessPerformance
	case financialreportmetrics.AssessmentKindInvestmentRisk, "investment_assessment", "investment", "investment_judgment":
		return financialreportmetrics.AssessmentKindInvestmentRisk
	default:
		return financialreportmetrics.AssessmentKindNone
	}
}

func NormalizeMetricsAssessmentScope(value string) string {
	value = normalizeMetricsRequestToken(value)
	switch value {
	case "":
		return ""
	case financialreportmetrics.AssessmentScopeMetricsOnly, "metrics_only", "financial_metrics_only", "report_metrics_only":
		return financialreportmetrics.AssessmentScopeMetricsOnly
	default:
		return financialreportmetrics.AssessmentScopeMetricsOnly
	}
}

func MetricsRequestedOutputsForFrame(frame MetricsRequestFrame) []string {
	outputs := NormalizeMetricsRequestedOutputs(frame.RequestedOutputs)
	if !MetricsOutputRequested(outputs, financialreportmetrics.RequestedOutputMetrics) {
		outputs = append([]string{financialreportmetrics.RequestedOutputMetrics}, outputs...)
	}
	switch MetricsAssessmentKindForFrame(frame) {
	case financialreportmetrics.AssessmentKindBusinessPerformance:
		outputs = appendMetricsOutput(outputs, financialreportmetrics.RequestedOutputPerformanceAssessment)
	case financialreportmetrics.AssessmentKindInvestmentRisk:
		outputs = appendMetricsOutput(outputs, financialreportmetrics.RequestedOutputInvestmentAssessment)
	}
	return outputs
}

func MetricsAssessmentKindForFrame(frame MetricsRequestFrame) string {
	kind := NormalizeMetricsAssessmentKind(frame.AssessmentKind)
	if kind != financialreportmetrics.AssessmentKindNone {
		return kind
	}
	outputs := NormalizeMetricsRequestedOutputs(frame.RequestedOutputs)
	if MetricsOutputRequested(outputs, financialreportmetrics.RequestedOutputInvestmentAssessment) {
		return financialreportmetrics.AssessmentKindInvestmentRisk
	}
	if MetricsOutputRequested(outputs, financialreportmetrics.RequestedOutputPerformanceAssessment) {
		return financialreportmetrics.AssessmentKindBusinessPerformance
	}
	return financialreportmetrics.AssessmentKindNone
}

func MetricsAssessmentScopeForFrame(frame MetricsRequestFrame) string {
	if scope := NormalizeMetricsAssessmentScope(frame.AssessmentScope); scope != "" {
		return scope
	}
	if MetricsAssessmentKindForFrame(frame) != financialreportmetrics.AssessmentKindNone {
		return financialreportmetrics.AssessmentScopeMetricsOnly
	}
	return ""
}

func MetricsAssessmentRequiresValuationForFrame(frame MetricsRequestFrame) bool {
	return frame.AssessmentRequiresValuation ||
		MetricsAssessmentKindForFrame(frame) == financialreportmetrics.AssessmentKindInvestmentRisk
}

func MetricsOutputRequested(outputs []string, target string) bool {
	target = NormalizeMetricsRequestedOutput(target)
	for _, output := range NormalizeMetricsRequestedOutputs(outputs) {
		if output == target {
			return true
		}
	}
	return false
}

func appendMetricsOutput(outputs []string, output string) []string {
	output = NormalizeMetricsRequestedOutput(output)
	if output == "" {
		return outputs
	}
	for _, existing := range outputs {
		if existing == output {
			return outputs
		}
	}
	return append(outputs, output)
}

func normalizeMetricsRequestToken(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, "-", "_")
	value = strings.ReplaceAll(value, " ", "_")
	return value
}
