package browserops

import "testing"

func TestEvaluateBrowserVisualEvidenceGatePassesSnapshotScreenshotAndURL(t *testing.T) {
	result := EvaluateBrowserVisualEvidenceGate(BrowserVisualEvidenceEvaluationInput{
		SnapshotText:          "heading Account updated\nbutton Save\nstatus Success",
		ScreenshotPath:        "data/tmp/browserops/success.png",
		FinalURL:              "https://example.com/accounts/123?updated=true",
		TargetURL:             "https://example.com/accounts",
		RequiredSnapshotTerms: []string{"Account updated", "Success"},
		RequireSnapshot:       true,
		RequireScreenshot:     true,
		RequireFinalURL:       true,
		MinSnapshotChars:      20,
	})
	if !result.Passed ||
		!result.VisualEvidenceReady ||
		!result.SnapshotReady ||
		!result.ScreenshotReady ||
		!result.FinalURLReady ||
		result.Score != 1 {
		t.Fatalf("expected visual evidence gate to pass, got %#v", result)
	}
	for _, expected := range []string{"snapshot_ready", "screenshot_ready", "final_url_ready", "snapshot_term:Account updated", "snapshot_term:Success"} {
		if !containsBrowserFailureReason(result.Evidence, expected) {
			t.Fatalf("expected evidence %q in %#v", expected, result.Evidence)
		}
	}
}

func TestEvaluateBrowserVisualEvidenceGateFailsMissingScreenshotAndSnapshotTerm(t *testing.T) {
	result := EvaluateBrowserVisualEvidenceGate(BrowserVisualEvidenceEvaluationInput{
		SnapshotText:          "heading Account updated",
		ScreenshotPath:        "",
		FinalURL:              "https://other.example.com/accounts/123",
		TargetURL:             "https://example.com/accounts",
		RequiredSnapshotTerms: []string{"Success"},
		RequireSnapshot:       true,
		RequireScreenshot:     true,
		RequireFinalURL:       true,
	})
	if result.Passed || result.VisualEvidenceReady {
		t.Fatalf("expected visual evidence gate to fail, got %#v", result)
	}
	for _, reason := range []string{"snapshot_term_missing:Success", "screenshot_missing", "final_url_target_mismatch"} {
		if !containsBrowserFailureReason(result.FailureReasons, reason) {
			t.Fatalf("expected reason %q in %#v", reason, result.FailureReasons)
		}
	}
	if result.Score != 0 {
		t.Fatalf("expected no required visual checks to pass, got %#v", result)
	}
}

func TestEvaluateBrowserVisualEvidenceGateDefaultsToSnapshotAndScreenshot(t *testing.T) {
	result := EvaluateBrowserVisualEvidenceGate(BrowserVisualEvidenceEvaluationInput{
		SnapshotText:   "button Save",
		ScreenshotPath: "data:image/png;base64,AAAA",
	})
	if !result.Passed || !result.SnapshotReady || !result.ScreenshotReady || !result.FinalURLReady {
		t.Fatalf("expected default snapshot/screenshot visual evidence gate to pass, got %#v", result)
	}
}
