package qualityevidence

import (
	"testing"

	"github.com/wsnacj/agentx-go/scenes/docparse/representation"
)

func TestEvaluateReportsBoundedTruncation(t *testing.T) {
	doc, err := representation.FromTextPages("fixture.pdf", []string{"approval rate 82 percent"}, representation.TextSourceOCRX)
	if err != nil {
		t.Fatalf("build representation: %v", err)
	}
	report := Evaluate(Spec{
		FixtureID:   "bounded",
		PrivacySafe: true,
		Pages: []PageExpectation{
			{Page: 1, RequiredPhrases: []string{"approval rate"}},
			{Page: 2, RequiredPhrases: []string{"manual review"}},
		},
	}, doc)
	if report.Status != StatusBounded || report.Passed || !report.Truncated || report.OCRStatus != StatusSuccess {
		t.Fatalf("expected bounded truncation readback, got %#v", report)
	}
}

func TestSkippedDoesNotClaimQualitySuccess(t *testing.T) {
	report := Skipped(Spec{FixtureID: "live_ocr", PrivacySafe: true}, "ocrx_config_missing")
	if report.Status != StatusSkipped || report.Passed || report.OCRStatus != StatusSkipped {
		t.Fatalf("unexpected skipped report: %#v", report)
	}
}
