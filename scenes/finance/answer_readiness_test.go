package finance

import (
	"testing"

	financialreportmetrics "github.com/wsnacj/agentx-go/scenes/finance/metrics"
)

func TestFinanceReportLookupAnswerReadinessPassedMetrics(t *testing.T) {
	got := FinanceReportLookupAnswerReadiness(FinanceReportLookupPayload{
		GuardStatus: "passed",
		Metrics: &MetricsToolPayload{
			GuardStatus:          "passed",
			RequestedFieldsReady: true,
			Evidence: MetricsEvidence{
				ReportPeriod: "2026-03-31",
				Revenue:      "10亿元",
				NetProfit:    "2亿元",
			},
		},
	})
	if !got.AnswerReady || got.Degraded || got.AllowedSummaryScope != AnswerScopeRequested || got.NextRepairHint != "" {
		t.Fatalf("unexpected readiness: %#v", got)
	}
}

func TestFinanceReportLookupAnswerReadinessRequiresAttributableProfitLabel(t *testing.T) {
	evidence := MetricsEvidence{
		ReportPeriod: "2026-03-31",
		NetProfit:    "-4.96亿元",
		MetricEvidence: map[string]financialreportmetrics.ReportDocumentMetricFieldEvidence{
			"net_profit": {
				Field:           "net_profit",
				Value:           "-4.96亿元",
				Evidence:        "HOLDER_PROFIT：归属于公司权益持有人的净利润或净亏损",
				SelectionReason: "eastmoney_hk_holder_profit",
			},
		},
	}
	got := FinanceReportLookupAnswerReadiness(FinanceReportLookupPayload{
		GuardStatus: "passed",
		Metrics: &MetricsToolPayload{
			GuardStatus:          "passed",
			RequestedFieldsReady: true,
			Evidence:             evidence,
		},
	})
	if !got.AnswerReady || len(got.PresentationRequirements) != 1 ||
		!containsAnyFoldString(got.PresentationRequirements[0], "归属于权益持有人") ||
		!NetProfitRequiresAttributableScope(evidence) {
		t.Fatalf("unexpected presentation requirements: %#v", got)
	}
}

func TestFinanceReportLookupAnswerReadinessPeriodScopeReview(t *testing.T) {
	got := FinanceReportLookupAnswerReadiness(FinanceReportLookupPayload{
		GuardStatus: "needs_review",
		Metrics: &MetricsToolPayload{
			GuardStatus:          "needs_review",
			RequestedFieldsReady: false,
			ReviewRequiredFields: []string{MetricPeriodScopeReviewField},
			Evidence: MetricsEvidence{
				ReportPeriod: "2025-12-31",
				Revenue:      "10亿元",
				NetProfit:    "2亿元",
			},
		},
	})
	if got.AnswerReady || !got.Degraded || got.DegradeReason != AnswerDegradePeriodScopeReview ||
		got.AllowedSummaryScope != AnswerScopeAvailablePeriodOnly ||
		got.NextRepairHint != "fetch_requested_period_source" {
		t.Fatalf("unexpected readiness: %#v", got)
	}
}

func TestFinanceReportLookupAnswerReadinessMissingRequestedFields(t *testing.T) {
	got := FinanceReportLookupAnswerReadiness(FinanceReportLookupPayload{
		GuardStatus: "missing_requested_fields",
		Metrics: &MetricsToolPayload{
			GuardStatus:            "missing_requested_fields",
			RequestedFieldsReady:   false,
			MissingRequestedFields: []string{"operating_cash_flow"},
		},
	})
	if got.AnswerReady || !got.Degraded || got.DegradeReason != AnswerDegradeMissingRequested ||
		got.AllowedSummaryScope != AnswerScopePartialVerifiedMetrics ||
		got.NextRepairHint != "fetch_missing_requested_fields" {
		t.Fatalf("unexpected readiness: %#v", got)
	}
}

func TestFinanceReportLookupAnswerReadinessBriefNotReady(t *testing.T) {
	got := FinanceReportLookupAnswerReadiness(FinanceReportLookupPayload{
		GuardStatus: "needs_review",
		Metrics: &MetricsToolPayload{
			GuardStatus:          "passed",
			RequestedFieldsReady: true,
		},
		Brief: &BriefToolPayload{
			GuardStatus: "needs_review",
			BriefReady:  false,
		},
	})
	if got.AnswerReady || got.DegradeReason != AnswerDegradeBriefNotReady || got.NextRepairHint != "fetch_or_parse_full_report_brief" {
		t.Fatalf("unexpected readiness: %#v", got)
	}
}

func TestFinanceReportLookupAnswerReadinessSourceReturnedNonPDF(t *testing.T) {
	got := FinanceReportLookupAnswerReadiness(FinanceReportLookupPayload{
		GuardStatus: "missing_requested_fields",
		Metrics: &MetricsToolPayload{
			GuardStatus:            "missing_requested_fields",
			RequestedFieldsReady:   false,
			MissingRequestedFields: []string{"revenue", "net_profit"},
			ReviewRequiredFields:   []string{MetricPeriodScopeReviewField},
			Warnings: []string{
				"exchange_a_share_pdf_download_returned_non_pdf:official ir artifact is not a valid PDF after curl fallback (bytes=3872 content_type=text/html; charset=utf-8 artifact_kind=html)",
			},
		},
	})
	if got.AnswerReady ||
		!got.Degraded ||
		got.DegradeReason != AnswerDegradeSourceReturnedNonPDF ||
		got.AllowedSummaryScope != AnswerScopeSourceMetadataOnly ||
		got.NextRepairHint != "stop_without_more_tools_or_use_configured_alternate_source" ||
		!got.StopRecommended ||
		got.StopReason != AnswerDegradeSourceReturnedNonPDF {
		t.Fatalf("unexpected readiness: %#v", got)
	}
}

func TestFinanceReportLookupAnswerReadinessOfficialSourceMissingStops(t *testing.T) {
	got := FinanceReportLookupAnswerReadiness(FinanceReportLookupPayload{
		GuardStatus: "missing_requested_fields",
		Metrics: &MetricsToolPayload{
			AdapterID:              "exchange_a_share_official_disclosure",
			GuardStatus:            "missing_requested_fields",
			RequestedFieldsReady:   false,
			MissingRequestedFields: []string{"revenue", "net_profit"},
			ReviewRequiredFields:   []string{MetricPeriodScopeReviewField, "source_policy"},
			Warnings:               []string{"exchange_a_share_annual_report_pdf_not_found"},
		},
	})
	if got.AnswerReady ||
		!got.Degraded ||
		got.DegradeReason != AnswerDegradeOfficialSourceMissing ||
		got.AllowedSummaryScope != AnswerScopeSourceMetadataOnly ||
		got.NextRepairHint != "stop_without_more_tools_or_use_configured_alternate_source" ||
		!got.StopRecommended ||
		got.StopReason != AnswerDegradeOfficialSourceMissing {
		t.Fatalf("unexpected readiness: %#v", got)
	}
}

func TestFinanceReportLookupAnswerReadinessSourceDownloadBlocked(t *testing.T) {
	got := FinanceReportLookupAnswerReadiness(FinanceReportLookupPayload{
		GuardStatus: "missing_requested_fields",
		Warnings: []string{
			"exchange_a_share_pdf_download_blocked_by_source:official ir artifact is not a valid PDF after curl fallback (block_hint=bot_denied)",
		},
		Metrics: &MetricsToolPayload{
			GuardStatus:          "missing_requested_fields",
			RequestedFieldsReady: false,
		},
	})
	if got.DegradeReason != AnswerDegradeSourceDownloadBlocked ||
		got.AllowedSummaryScope != AnswerScopeSourceMetadataOnly ||
		!got.StopRecommended ||
		got.StopReason != AnswerDegradeSourceDownloadBlocked {
		t.Fatalf("unexpected readiness: %#v", got)
	}
}
