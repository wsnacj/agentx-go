package browserops

import "testing"

func TestEvaluateBrowserPageStateEvidencePassesRequiredState(t *testing.T) {
	result := EvaluateBrowserPageStateEvidence(BrowserPageStateEvaluationInput{
		SnapshotText:      "heading Customer profile\nstatus Account active\nbutton Save",
		ScreenshotPath:    "data/tmp/browserops/page-state.png",
		FinalURL:          "https://example.com/customers/42?tab=profile",
		TargetURL:         "https://example.com/customers/42",
		RequiredText:      []string{"Customer profile", "Account active"},
		ForbiddenText:     []string{"Access denied"},
		URLContains:       "customers/42",
		RequireScreenshot: true,
		MinSnapshotChars:  20,
	})
	if !result.Passed ||
		!result.EvidenceReady ||
		!result.SnapshotReady ||
		!result.ScreenshotReady ||
		!result.FinalURLReady ||
		!result.RequiredTextReady ||
		!result.ForbiddenTextReady ||
		!result.URLExpectationReady ||
		result.Score != 1 {
		t.Fatalf("expected page state evidence to pass, got %#v", result)
	}
	for _, expected := range []string{"snapshot_ready", "screenshot_ready", "final_url_ready", "required_text:Customer profile", "forbidden_text_absent:Access denied", "final_url_contains:customers/42"} {
		if !containsBrowserFailureReason(result.Evidence, expected) {
			t.Fatalf("expected evidence %q in %#v", expected, result.Evidence)
		}
	}
}

func TestEvaluateBrowserPageStateEvidenceFailsExpectationMismatches(t *testing.T) {
	result := EvaluateBrowserPageStateEvidence(BrowserPageStateEvaluationInput{
		SnapshotText:      "heading Customer profile\nstatus Access denied",
		ScreenshotPath:    "data/tmp/browserops/page-state.png",
		FinalURL:          "https://other.example.net/login",
		TargetURL:         "https://example.com/customers/42",
		RequiredText:      []string{"Account active"},
		ForbiddenText:     []string{"Access denied"},
		URLContains:       "customers/42",
		RequireScreenshot: true,
	})
	if result.Passed || result.RequiredTextReady || result.ForbiddenTextReady || result.URLExpectationReady {
		t.Fatalf("expected page state expectations to fail, got %#v", result)
	}
	for _, reason := range []string{"required_text_missing:Account active", "forbidden_text_present:Access denied", "final_url_target_mismatch", "final_url_contains_missing:customers/42"} {
		if !containsBrowserFailureReason(result.FailureReasons, reason) {
			t.Fatalf("expected reason %q in %#v", reason, result.FailureReasons)
		}
	}
}

func TestEvaluateBrowserPageStateEvidenceUsesEvidenceBundle(t *testing.T) {
	bundle := BuildBrowserEvidenceBundleFromState(map[string]any{
		"target_url": "https://example.com/status",
		"review": map[string]any{
			"snapshot":      "heading Status\nstatus Healthy",
			"final_url":     "https://example.com/status",
			"evidence_path": "data:image/png;base64,AAAA",
		},
	})
	result := EvaluateBrowserPageStateEvidence(BrowserPageStateEvaluationInput{
		Bundle:            bundle,
		RequiredText:      []string{"Healthy"},
		RequireScreenshot: true,
	})
	if !result.Passed || result.EvidenceBundle.PageSnapshot == nil || result.EvidenceBundle.Screenshot == nil {
		t.Fatalf("expected evidence bundle-backed page state to pass, got %#v", result)
	}
}
