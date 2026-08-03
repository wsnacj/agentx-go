package finance

import (
	"testing"

	financialreportmetrics "github.com/wsnacj/agentx-go/scenes/finance/metrics"
)

func TestMetricsEvidenceFromDocumentMetricEvidence(t *testing.T) {
	fieldEvidence := map[string]financialreportmetrics.ReportDocumentMetricFieldEvidence{
		"revenue": {Field: "revenue", Value: "100 million RMB"},
	}
	evidence := MetricsEvidenceFromDocumentMetricEvidence(DocumentMetricEvidenceProjectionInput{
		Evidence: financialreportmetrics.ReportDocumentMetricEvidence{
			CompanyName:       "Acme Holdings",
			StockCode:         "ACME",
			SelectionReason:   "selected_by_parser",
			OfficialSource:    "https://example.com/report.pdf",
			ReportPeriod:      "2025-12-31",
			Revenue:           "100 million RMB",
			NetProfit:         "20 million RMB",
			OperatingCashFlow: "30 million RMB",
			ArtifactPath:      "/tmp/acme-2025.pdf",
			FieldEvidence:     fieldEvidence,
		},
		Title: "Acme annual report",
	})
	if evidence.CompanyName != "Acme Holdings" {
		t.Fatalf("unexpected company name: %q", evidence.CompanyName)
	}
	if evidence.StockCode != "ACME" {
		t.Fatalf("unexpected stock code: %q", evidence.StockCode)
	}
	if evidence.SelectionReason != "selected_by_parser" {
		t.Fatalf("unexpected selection reason: %q", evidence.SelectionReason)
	}
	if evidence.PageTitle != "Acme annual report" {
		t.Fatalf("unexpected page title: %q", evidence.PageTitle)
	}
	if evidence.MetricEvidence["revenue"].Value != "100 million RMB" {
		t.Fatalf("metric evidence was not preserved: %+v", evidence.MetricEvidence)
	}
}

func TestMetricsEvidenceFromDocumentMetricEvidenceUsesContextFallbacks(t *testing.T) {
	evidence := MetricsEvidenceFromDocumentMetricEvidence(DocumentMetricEvidenceProjectionInput{
		Evidence: financialreportmetrics.ReportDocumentMetricEvidence{
			ArtifactPath: "/tmp/report.pdf",
		},
		EntityName:              "Fallback Company",
		StockCode:               "000001",
		Ticker:                  "fallback",
		SourceURL:               "https://example.com/fallback.pdf",
		ReportPath:              "/tmp/local.pdf",
		SelectionReasonFallback: "selected_by_project_docparse_annual_report_adapter",
		PageTitleFallback:       "docparse_annual_report_metrics",
	})
	if evidence.CompanyName != "Fallback Company" {
		t.Fatalf("unexpected fallback company: %q", evidence.CompanyName)
	}
	if evidence.StockCode != "000001" {
		t.Fatalf("unexpected fallback stock code: %q", evidence.StockCode)
	}
	if evidence.OfficialSource != "https://example.com/fallback.pdf" {
		t.Fatalf("unexpected fallback source: %q", evidence.OfficialSource)
	}
	if evidence.SelectionReason != "selected_by_project_docparse_annual_report_adapter" {
		t.Fatalf("unexpected fallback selection reason: %q", evidence.SelectionReason)
	}
	if evidence.Revenue != "unknown" || evidence.NetProfit != "unknown" {
		t.Fatalf("expected missing metric fallbacks, got revenue=%q net_profit=%q", evidence.Revenue, evidence.NetProfit)
	}
	if evidence.PageTitle != "report.pdf" {
		t.Fatalf("unexpected artifact page title fallback: %q", evidence.PageTitle)
	}
}

func TestMetricsEvidenceFromDocumentMetricEvidenceUsesExplicitPageTitleFallback(t *testing.T) {
	evidence := MetricsEvidenceFromDocumentMetricEvidence(DocumentMetricEvidenceProjectionInput{
		PageTitleFallback: "docparse_annual_report_metrics",
	})
	if evidence.PageTitle != "docparse_annual_report_metrics" {
		t.Fatalf("unexpected page title fallback: %q", evidence.PageTitle)
	}
}
