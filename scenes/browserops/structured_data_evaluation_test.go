package browserops

import "testing"

func TestEvaluateBrowserStructuredDataEvidencePassesExpectedFields(t *testing.T) {
	result := EvaluateBrowserStructuredDataEvidence(BrowserStructuredDataEvaluationInput{
		SnapshotText:      "heading Account summary\nCompany: Acme Robotics\nStatus: Active\nARR: $42,000",
		ScreenshotPath:    "data/tmp/browserops/structured-data.png",
		FinalURL:          "https://example.com/accounts/acme",
		TargetURL:         "https://example.com/accounts/acme",
		URLContains:       "accounts/acme",
		RequireScreenshot: true,
		MinSnapshotChars:  20,
		Fields: []BrowserStructuredDataFieldExpectation{
			{Name: "company_name", Required: true, ExpectedText: "Acme Robotics"},
			{Name: "status", Required: true, ExpectedText: "Active"},
			{Name: "arr", Required: true, ExpectedText: "$42,000"},
		},
		ExtractedData: map[string]any{
			"company_name": "Acme Robotics",
			"status":       "Active",
			"arr":          "$42,000",
		},
	})
	if !result.Passed ||
		!result.EvidenceReady ||
		!result.FieldsReady ||
		!result.SnapshotReady ||
		!result.ScreenshotReady ||
		!result.FinalURLReady ||
		!result.URLExpectationReady ||
		result.Score != 1 {
		t.Fatalf("expected structured data evidence to pass, got %#v", result)
	}
	if len(result.FieldResults) != 3 {
		t.Fatalf("expected 3 field results, got %#v", result.FieldResults)
	}
	for _, expected := range []string{"field_ready:company_name", "field_ready:status", "field_ready:arr", "final_url_contains:accounts/acme"} {
		if !containsBrowserFailureReason(result.Evidence, expected) {
			t.Fatalf("expected evidence %q in %#v", expected, result.Evidence)
		}
	}
}

func TestEvaluateBrowserStructuredDataEvidenceFailsMissingFields(t *testing.T) {
	result := EvaluateBrowserStructuredDataEvidence(BrowserStructuredDataEvaluationInput{
		SnapshotText:      "heading Account summary\nCompany: Acme Robotics\nStatus: Pending",
		ScreenshotPath:    "data/tmp/browserops/structured-data.png",
		FinalURL:          "https://example.com/accounts/acme",
		TargetURL:         "https://example.com/accounts/acme",
		RequireScreenshot: true,
		Fields: []BrowserStructuredDataFieldExpectation{
			{Name: "company_name", Required: true, ExpectedText: "Acme Robotics"},
			{Name: "status", Required: true, ExpectedText: "Active"},
			{Name: "arr", Required: true, ExpectedText: "$42,000"},
		},
		ExtractedData: map[string]any{
			"company_name": "Acme Robotics",
			"status":       "Pending",
		},
	})
	if result.Passed || result.FieldsReady {
		t.Fatalf("expected structured data fields to fail, got %#v", result)
	}
	for _, reason := range []string{
		"field_status:expected_text_missing",
		"field_status:extracted_value_mismatch",
		"field_arr:extracted_value_missing",
		"field_arr:expected_text_missing",
	} {
		if !containsBrowserFailureReason(result.FailureReasons, reason) {
			t.Fatalf("expected reason %q in %#v", reason, result.FailureReasons)
		}
	}
}

func TestEvaluateBrowserStructuredDataEvidenceUsesEvidenceBundle(t *testing.T) {
	bundle := BuildBrowserEvidenceBundleFromState(map[string]any{
		"target_url": "https://example.com/accounts/acme",
		"review": map[string]any{
			"snapshot":      "heading Account summary\nCompany: Acme Robotics",
			"final_url":     "https://example.com/accounts/acme",
			"evidence_path": "data:image/png;base64,AAAA",
		},
	})
	result := EvaluateBrowserStructuredDataEvidence(BrowserStructuredDataEvaluationInput{
		Bundle:            bundle,
		RequireScreenshot: true,
		Fields: []BrowserStructuredDataFieldExpectation{
			{Name: "company_name", Required: true, ExpectedText: "Acme Robotics"},
		},
		ExtractedData: map[string]any{"company_name": "Acme Robotics"},
	})
	if !result.Passed || result.EvidenceBundle.PageSnapshot == nil || result.EvidenceBundle.Screenshot == nil {
		t.Fatalf("expected evidence bundle-backed structured data gate to pass, got %#v", result)
	}
}
