package finance

import "testing"

func TestExtractStructuredTextMetricsEvidence(t *testing.T) {
	evidence, ok := ExtractStructuredTextMetricsEvidence(StructuredTextMetricsInput{
		Text:             "COMPANY_NAME=Acme Holdings; STOCK_CODE=ACME; YEAR=2025; REVENUE=100 million USD; NET_PROFIT=20 million USD; SOURCE_URL=https://example.com/report.pdf",
		Title:            "Acme annual report",
		RequestedMetrics: []string{"revenue", "net_profit"},
	})
	if !ok {
		t.Fatalf("expected structured metrics evidence")
	}
	if evidence.CompanyName != "Acme Holdings" {
		t.Fatalf("unexpected company name: %q", evidence.CompanyName)
	}
	if evidence.StockCode != "ACME" {
		t.Fatalf("unexpected stock code: %q", evidence.StockCode)
	}
	if evidence.ReportPeriod != "2025-12-31" {
		t.Fatalf("unexpected report period: %q", evidence.ReportPeriod)
	}
	if evidence.Revenue != "100 million USD" {
		t.Fatalf("unexpected revenue: %q", evidence.Revenue)
	}
	if evidence.NetProfit != "20 million USD" {
		t.Fatalf("unexpected net profit: %q", evidence.NetProfit)
	}
	if evidence.OfficialSource != "https://example.com/report.pdf" {
		t.Fatalf("unexpected official source: %q", evidence.OfficialSource)
	}
}

func TestExtractStructuredTextMetricsEvidenceRejectsNoFields(t *testing.T) {
	if evidence, ok := ExtractStructuredTextMetricsEvidence(StructuredTextMetricsInput{
		Text:             "plain annual-report text without structured metrics",
		RequestedMetrics: []string{"revenue"},
	}); ok {
		t.Fatalf("expected no structured evidence, got %+v", evidence)
	}
}

func TestExtractStructuredTextMetricsEvidenceRejectsMissingRequestedSignal(t *testing.T) {
	if evidence, ok := ExtractStructuredTextMetricsEvidence(StructuredTextMetricsInput{
		Text:             "COMPANY_NAME=Acme Holdings; YEAR=2025; REVENUE=NOT_FOUND; NET_PROFIT=N/A",
		RequestedMetrics: []string{"revenue", "net_profit"},
	}); ok {
		t.Fatalf("expected no signal from unknown metrics, got %+v", evidence)
	}
}

func TestParseStructuredMetricFieldsAcceptsChineseSemicolon(t *testing.T) {
	fields := ParseStructuredMetricFields("REVENUE=100；NET_PROFIT=20")
	if fields["REVENUE"] != "100" {
		t.Fatalf("unexpected revenue field: %q", fields["REVENUE"])
	}
	if fields["NET_PROFIT"] != "20" {
		t.Fatalf("unexpected net profit field: %q", fields["NET_PROFIT"])
	}
}
