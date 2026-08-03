package finance

import (
	"reflect"
	"testing"

	financialreportmetrics "github.com/wsnacj/agentx-go/scenes/finance/metrics"
)

func TestNormalizeMetricsRequestedOutputsSupportsAliases(t *testing.T) {
	got := NormalizeMetricsRequestedOutputs([]string{
		"财报指标",
		"performance-assessment",
		"投资判断",
		"metrics",
		"unknown",
	})
	want := []string{
		financialreportmetrics.RequestedOutputMetrics,
		financialreportmetrics.RequestedOutputPerformanceAssessment,
		financialreportmetrics.RequestedOutputInvestmentAssessment,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected requested outputs: got %#v want %#v", got, want)
	}
}

func TestMetricsRequestedOutputsForFrameAddsMetricsAndAssessmentOutput(t *testing.T) {
	got := MetricsRequestedOutputsForFrame(MetricsRequestFrame{
		RequestedOutputs: []string{"业绩评估"},
	})
	want := []string{
		financialreportmetrics.RequestedOutputMetrics,
		financialreportmetrics.RequestedOutputPerformanceAssessment,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected projected outputs: got %#v want %#v", got, want)
	}
}

func TestMetricsAssessmentProjectionFromFrame(t *testing.T) {
	frame := MetricsRequestFrame{
		RequestedOutputs:            []string{"metrics"},
		AssessmentKind:              "investment_judgment",
		AssessmentRequiresValuation: false,
	}
	if got := MetricsAssessmentKindForFrame(frame); got != financialreportmetrics.AssessmentKindInvestmentRisk {
		t.Fatalf("unexpected assessment kind %q", got)
	}
	if got := MetricsAssessmentScopeForFrame(frame); got != financialreportmetrics.AssessmentScopeMetricsOnly {
		t.Fatalf("unexpected assessment scope %q", got)
	}
	if !MetricsAssessmentRequiresValuationForFrame(frame) {
		t.Fatalf("expected investment assessment to require valuation")
	}
}
