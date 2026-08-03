package finance

import (
	"reflect"
	"testing"

	financialreportmetrics "github.com/wsnacj/agentx-go/scenes/finance/metrics"
)

func TestMetricSpecFromFieldsNormalizesBilingualAliases(t *testing.T) {
	spec := MetricSpecFromFields([]string{"营收", "revenue", "净利润增长", "operating cashflow"})
	if !spec.Revenue || !spec.NetProfitGrowth || !spec.OperatingCashFlow || spec.NetProfit || spec.RevenueGrowth {
		t.Fatalf("unexpected spec: %+v", spec)
	}
	if got, want := spec.Fields(), []string{"revenue", "net_profit_growth", "operating_cash_flow"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected fields: got %+v want %+v", got, want)
	}
}

func TestMetricSpecMissingAndReviewFields(t *testing.T) {
	spec := MetricSpec{Revenue: true, NetProfit: true, NetProfitGrowth: true}
	evidence := MetricsEvidence{
		Revenue:         "751,766 RMB million",
		NetProfit:       "224,842",
		NetProfitGrowth: "23.1%",
		MetricEvidence: map[string]financialreportmetrics.ReportDocumentMetricFieldEvidence{
			"net_profit": {
				Field:  "net_profit",
				Value:  "224,842",
				Source: "table",
				Period: "2025-12-31",
			},
		},
	}
	if got := spec.MissingFields(evidence); len(got) != 0 {
		t.Fatalf("unexpected missing fields: %+v", got)
	}
	if got := MetricReviewRequiredFields(spec, evidence); !reflect.DeepEqual(got, []string{"net_profit"}) {
		t.Fatalf("unexpected review fields: %+v", got)
	}
}

func TestMetricSpecReadyRequiresConcreteFieldEvidence(t *testing.T) {
	spec := MetricSpec{Revenue: true}
	evidence := MetricsEvidence{
		Revenue: "751,766 RMB million",
		MetricEvidence: map[string]financialreportmetrics.ReportDocumentMetricFieldEvidence{
			"revenue": {
				Field:  "revenue",
				Value:  "751,766",
				Source: "https://example.com/report.pdf",
				Period: "2025-12-31",
				Unit:   "RMB million",
			},
		},
	}
	if !spec.Ready(evidence) {
		t.Fatalf("expected evidence ready")
	}
	evidence.MetricEvidence["revenue"] = financialreportmetrics.ReportDocumentMetricFieldEvidence{
		Field:  "revenue",
		Value:  "751,766",
		Source: "https://example.com/report.pdf",
		Period: "2025-12-31",
	}
	if spec.Ready(evidence) {
		t.Fatalf("expected bare monetary value to require review")
	}
}

func TestTrendSeriesReadyRequiresContinuousConcreteYears(t *testing.T) {
	spec := MetricSpec{Revenue: true, NetProfit: true}
	series := []MetricsTrendSeriesPoint{
		{Period: "2025-12-31", Revenue: "751,766 RMB million", NetProfit: "224,842 RMB million", Source: "https://example.com/2025.pdf"},
		{Period: "2024-12-31", Revenue: "660,257 RMB million", NetProfit: "194,073 RMB million", Source: "https://example.com/2024.pdf"},
	}
	if !TrendSeriesReady(spec, series) {
		t.Fatalf("expected continuous series ready")
	}
	series[1].Period = "2023-12-31"
	if TrendSeriesReady(spec, series) {
		t.Fatalf("expected non-continuous series not ready")
	}
}
