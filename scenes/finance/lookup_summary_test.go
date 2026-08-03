package finance

import "testing"

func TestFinanceReportLookupSummariesPreferVerifiedCandidateNameOverCodeOnlyEvidence(t *testing.T) {
	summaries := FinanceReportLookupSummaries([]FinanceReportLookupPayload{{
		AdapterID:   "tonghuashun_a_share",
		GuardStatus: "passed",
		Intent: FinanceReportLookupIntent{
			EntityName: "同花顺",
			StockCode:  "300033",
		},
		Candidates: &MetricsCandidatesPayload{
			ResolvedCompany: "同花顺",
			ResolvedCode:    "300033",
		},
		Metrics: &MetricsToolPayload{
			AdapterID:            "tonghuashun_a_share",
			GuardStatus:          "passed",
			RequestedFieldsReady: true,
			Evidence: MetricsEvidence{
				CompanyName:  "300033",
				StockCode:    "300033.SZ",
				ReportPeriod: "2026-03-31",
				Revenue:      "10.53亿元",
				NetProfit:    "2.56亿元",
			},
		},
	}})
	if len(summaries) != 1 {
		t.Fatalf("expected one summary, got %#v", summaries)
	}
	if summaries[0].EntityName != "同花顺" || summaries[0].StockCode != "300033.SZ" {
		t.Fatalf("expected verified subject name with evidence code, got %#v", summaries[0])
	}
}

func TestFinanceReportLookupSummariesPreserveMultiplePayloads(t *testing.T) {
	summaries := FinanceReportLookupSummaries([]FinanceReportLookupPayload{
		{
			Intent: FinanceReportLookupIntent{EntityName: "小鹏"},
			Metrics: &MetricsToolPayload{
				GuardStatus:          "passed",
				RequestedFieldsReady: true,
				Evidence: MetricsEvidence{
					CompanyName:  "小鹏集团-W",
					StockCode:    "09868.HK",
					ReportPeriod: "2025-12-31",
					Revenue:      "408.66亿元",
					NetProfit:    "-57.89亿元",
				},
			},
		},
		{
			Intent: FinanceReportLookupIntent{EntityName: "理想"},
			Metrics: &MetricsToolPayload{
				GuardStatus:          "passed",
				RequestedFieldsReady: true,
				Evidence: MetricsEvidence{
					CompanyName:  "理想汽车-W",
					StockCode:    "02015.HK",
					ReportPeriod: "2025-12-31",
					Revenue:      "1445.00亿元",
					NetProfit:    "80.00亿元",
				},
			},
		},
	})
	if len(summaries) != 2 || summaries[0].EntityName != "小鹏集团-W" || summaries[1].EntityName != "理想汽车-W" {
		t.Fatalf("unexpected summaries: %#v", summaries)
	}
}

func TestFinanceReportLookupSummaryProjectsAssessmentBoundary(t *testing.T) {
	summary, ok := FinanceReportLookupSummaryFromPayload(FinanceReportLookupPayload{
		Intent: FinanceReportLookupIntent{
			EntityName:       "同花顺",
			RequestedOutputs: []string{"investment_assessment"},
			Assessment: map[string]any{
				"kind":               "investment_risk",
				"requires_valuation": true,
			},
		},
		AnswerReady: &FinanceReportAnswerReadiness{
			AnswerReady:          false,
			RequestedFieldsReady: false,
			MissingFields:        []string{"operating_cash_flow"},
		},
		Metrics: &MetricsToolPayload{
			GuardStatus:            "missing_requested_fields",
			RequestedFieldsReady:   false,
			MissingRequestedFields: []string{"operating_cash_flow"},
			Evidence: MetricsEvidence{
				CompanyName:     "同花顺",
				StockCode:       "300033.SZ",
				ReportPeriod:    "2025-12-31",
				Revenue:         "60.29亿元",
				NetProfit:       "32.05亿元",
				RevenueGrowth:   "同比增长44.00%",
				NetProfitGrowth: "同比增长75.79%",
			},
		},
	})
	if !ok {
		t.Fatal("expected summary")
	}
	if summary.Assessment == nil ||
		summary.Assessment.Kind != "investment_risk" ||
		summary.Assessment.Status != FinanceAssessmentStatusPartial ||
		summary.Assessment.CashFlowStatus != FinanceAssessmentCashFlowMissing ||
		!lookupSummaryTestContains(summary.Assessment.MissingInputs, "operating_cash_flow") ||
		!lookupSummaryTestContains(summary.Assessment.MissingInputs, "valuation_metrics") {
		t.Fatalf("unexpected assessment projection in summary: %#v", summary)
	}
}

func lookupSummaryTestContains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
