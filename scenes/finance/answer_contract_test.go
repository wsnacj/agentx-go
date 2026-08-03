package finance

import (
	"strings"
	"testing"
)

func TestFinanceReportLookupAnswerContractSourceReturnedNonPDF(t *testing.T) {
	payload := FinanceReportLookupPayload{
		Intent: FinanceReportLookupIntent{
			EntityName:       "瑞可达",
			StockCode:        "688800",
			RequestedMetrics: []string{"revenue", "net_profit"},
		},
		Candidates: &MetricsCandidatesPayload{
			PrimaryURL: "https://www.sse.com.cn/disclosure/listedinfo/announcement/c/new/2026-04-21/688800_20260421_11VH.pdf",
		},
		Metrics: &MetricsToolPayload{
			GuardStatus:            "missing_requested_fields",
			RequestedFieldsReady:   false,
			MissingRequestedFields: []string{"revenue", "net_profit"},
			Warnings: []string{
				"exchange_a_share_pdf_download_returned_non_pdf:official ir artifact is not a valid PDF after curl fallback (content_type=text/html artifact_kind=html)",
			},
			Evidence: MetricsEvidence{
				CompanyName:    "exchange",
				StockCode:      "688800.SH",
				OfficialSource: "https://www.sse.com.cn/disclosure/listedinfo/announcement/c/new/2026-04-21/688800_20260421_11VH.pdf",
			},
		},
	}
	readiness := FinanceReportLookupAnswerReadiness(payload)
	payload.AnswerReady = &readiness

	contract := FinanceReportLookupAnswerContract(payload)
	if contract == nil {
		t.Fatal("expected answer contract")
	}
	if !contract.FinalAnswerRecommended ||
		contract.Reason != AnswerDegradeSourceReturnedNonPDF ||
		contract.AllowedSummaryScope != AnswerScopeSourceMetadataOnly {
		t.Fatalf("unexpected answer contract: %#v", contract)
	}
	for _, want := range []string{"瑞可达", "688800.SH", "不是有效 PDF", "revenue、net_profit"} {
		if !strings.Contains(contract.FinalAnswerDraft, want) {
			t.Fatalf("expected final answer draft to contain %q, got %q", want, contract.FinalAnswerDraft)
		}
	}
	for _, tool := range []string{"pdf", "browser", "search"} {
		if !stringSliceContains(contract.DoNotRetryTools, tool) {
			t.Fatalf("expected do_not_retry_tools to contain %q, got %#v", tool, contract.DoNotRetryTools)
		}
	}
}

func TestFinanceReportLookupAnswerContractOfficialSourceMissing(t *testing.T) {
	payload := FinanceReportLookupPayload{
		Intent: FinanceReportLookupIntent{
			EntityName:       "瑞可达",
			StockCode:        "688800",
			RequestedMetrics: []string{"revenue", "net_profit"},
		},
		Candidates: &MetricsCandidatesPayload{
			PrimaryURL: "https://www.sse.com.cn/disclosure/listedinfo/announcement/",
		},
		Metrics: &MetricsToolPayload{
			AdapterID:              "exchange_a_share_official_disclosure",
			GuardStatus:            "missing_requested_fields",
			RequestedFieldsReady:   false,
			MissingRequestedFields: []string{"revenue", "net_profit"},
			Warnings:               []string{"exchange_a_share_annual_report_pdf_not_found"},
			Evidence: MetricsEvidence{
				CompanyName:    "瑞可达",
				StockCode:      "688800.SH",
				OfficialSource: "https://www.sse.com.cn/disclosure/listedinfo/announcement/",
			},
		},
	}
	readiness := FinanceReportLookupAnswerReadiness(payload)
	payload.AnswerReady = &readiness

	contract := FinanceReportLookupAnswerContract(payload)
	if contract == nil {
		t.Fatal("expected answer contract")
	}
	if !contract.FinalAnswerRecommended ||
		contract.Reason != AnswerDegradeOfficialSourceMissing ||
		contract.AllowedSummaryScope != AnswerScopeSourceMetadataOnly ||
		!strings.Contains(contract.FinalAnswerDraft, "官方来源没有提供可直接解析的报告文件") {
		t.Fatalf("unexpected answer contract: %#v", contract)
	}
	for _, tool := range []string{"browser", "search", "open_page", "find_in_page"} {
		if !stringSliceContains(contract.DoNotRetryTools, tool) {
			t.Fatalf("expected do_not_retry_tools to contain %q, got %#v", tool, contract.DoNotRetryTools)
		}
	}
}

func TestFinanceReportLookupAnswerContractNotEmittedWhenRepairable(t *testing.T) {
	payload := FinanceReportLookupPayload{
		Metrics: &MetricsToolPayload{
			GuardStatus:          "missing_requested_fields",
			RequestedFieldsReady: false,
			Evidence:             MetricsEvidence{CompanyName: "示例公司"},
		},
	}
	readiness := FinanceReportLookupAnswerReadiness(payload)
	payload.AnswerReady = &readiness
	if contract := FinanceReportLookupAnswerContract(payload); contract != nil {
		t.Fatalf("did not expect answer contract for non-terminal repairable readiness: %#v", contract)
	}
}

func stringSliceContains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
