package finance

import (
	"testing"

	financialreportmetrics "github.com/wsnacj/agentx-go/scenes/finance/metrics"
)

func TestFinanceReportAssessmentProjectionInvestmentRiskMissingCashFlowAndValuation(t *testing.T) {
	payload := FinanceReportLookupPayload{
		Intent: FinanceReportLookupIntent{
			RequestedOutputs: []string{financialreportmetrics.RequestedOutputInvestmentAssessment},
			Assessment: map[string]any{
				"kind":               financialreportmetrics.AssessmentKindInvestmentRisk,
				"scope":              financialreportmetrics.AssessmentScopeMetricsOnly,
				"requires_valuation": true,
			},
		},
		AnswerReady: &FinanceReportAnswerReadiness{
			AnswerReady:          false,
			DegradeReason:        AnswerDegradeMissingRequested,
			AllowedSummaryScope:  AnswerScopePartialVerifiedMetrics,
			RequestedFieldsReady: false,
			MissingFields:        []string{"operating_cash_flow"},
		},
		Metrics: &MetricsToolPayload{
			GuardStatus:            "missing_requested_fields",
			RequestedFieldsReady:   false,
			MissingRequestedFields: []string{"operating_cash_flow"},
			Evidence: MetricsEvidence{
				ReportPeriod:    "2025-12-31",
				OfficialSource:  "https://example.com/report.pdf",
				Revenue:         "60.29亿元",
				RevenueGrowth:   "同比增长44.00%",
				NetProfit:       "32.05亿元",
				NetProfitGrowth: "同比增长75.79%",
			},
		},
	}
	got, ok := FinanceReportAssessmentFromPayload(payload)
	if !ok {
		t.Fatal("expected assessment projection")
	}
	if got.Kind != financialreportmetrics.AssessmentKindInvestmentRisk ||
		got.Status != FinanceAssessmentStatusPartial ||
		got.CashFlowStatus != FinanceAssessmentCashFlowMissing ||
		!assessmentTestContains(got.MissingInputs, "operating_cash_flow") ||
		!assessmentTestContains(got.MissingInputs, "valuation_metrics") ||
		!assessmentTestContainsRef(got.RiskFactors, "cash_flow_unavailable") ||
		!assessmentTestContainsRef(got.RiskFactors, "valuation_unavailable") ||
		got.AdviceBoundary != "partial_assessment_only_missing_inputs_must_be_disclosed" {
		t.Fatalf("unexpected assessment projection: %#v", got)
	}
	if len(got.VerifiedFacts) < 4 || !assessmentTestContainsRef(got.PositiveFactors, "core_metrics_available") {
		t.Fatalf("expected verified metric facts and positive factors, got %#v", got)
	}
}

func TestFinanceReportAssessmentProjectionReadyPerformanceAssessment(t *testing.T) {
	payload := FinanceReportLookupPayload{
		Intent: FinanceReportLookupIntent{
			RequestedOutputs: []string{financialreportmetrics.RequestedOutputPerformanceAssessment},
			Assessment: map[string]any{
				"kind":  financialreportmetrics.AssessmentKindBusinessPerformance,
				"scope": financialreportmetrics.AssessmentScopeMetricsOnly,
			},
		},
		AnswerReady: &FinanceReportAnswerReadiness{
			AnswerReady:          true,
			AllowedSummaryScope:  AnswerScopeRequested,
			RequestedFieldsReady: true,
		},
		Metrics: &MetricsToolPayload{
			GuardStatus:          "passed",
			RequestedFieldsReady: true,
			Evidence: MetricsEvidence{
				ReportPeriod:    "2025-12-31",
				OfficialSource:  "https://example.com/report.pdf",
				Revenue:         "60.29亿元",
				RevenueGrowth:   "同比增长44.00%",
				NetProfit:       "32.05亿元",
				NetProfitGrowth: "同比增长75.79%",
			},
		},
	}
	got, ok := FinanceReportAssessmentFromPayload(payload)
	if !ok {
		t.Fatal("expected assessment projection")
	}
	if got.Status != FinanceAssessmentStatusReady ||
		got.CashFlowStatus != FinanceAssessmentCashFlowNotRequested ||
		got.AdviceBoundary != "facts_support_metrics_level_assessment_only" ||
		len(got.MissingInputs) != 0 {
		t.Fatalf("unexpected ready assessment projection: %#v", got)
	}
}

func TestFinanceReportAssessmentProjectionUsesFieldPeriod(t *testing.T) {
	payload := FinanceReportLookupPayload{
		Intent: FinanceReportLookupIntent{
			RequestedOutputs: []string{financialreportmetrics.RequestedOutputPerformanceAssessment},
		},
		Metrics: &MetricsToolPayload{
			Evidence: MetricsEvidence{
				ReportPeriod:      "2026-03-28 10-Q",
				OfficialSource:    "https://www.sec.gov/example.htm",
				Revenue:           "USD111.184 billion",
				OperatingCashFlow: "USD82.627 billion",
				MetricEvidence: map[string]financialreportmetrics.ReportDocumentMetricFieldEvidence{
					"revenue": {
						Period: "2025-12-28 至 2026-03-28，三个月",
					},
					"operating_cash_flow": {
						Period: "2025-09-28 至 2026-03-28，六个月累计",
					},
				},
			},
		},
	}
	got, ok := FinanceReportAssessmentFromPayload(payload)
	if !ok {
		t.Fatal("expected assessment projection")
	}
	periods := map[string]string{}
	for _, fact := range got.VerifiedFacts {
		periods[fact.Field] = fact.Period
	}
	if periods["revenue"] != "2025-12-28 至 2026-03-28，三个月" ||
		periods["operating_cash_flow"] != "2025-09-28 至 2026-03-28，六个月累计" {
		t.Fatalf("expected field-level assessment periods, got %#v", periods)
	}
}

func assessmentTestContains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func assessmentTestContainsRef(values []FinanceReportAssessmentRef, code string) bool {
	for _, value := range values {
		if value.Code == code {
			return true
		}
	}
	return false
}
