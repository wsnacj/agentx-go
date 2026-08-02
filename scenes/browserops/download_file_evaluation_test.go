package browserops

import "testing"

func TestEvaluateBrowserDownloadFileEvidencePassesExpectedArtifact(t *testing.T) {
	result := EvaluateBrowserDownloadFileEvidence(BrowserDownloadFileEvaluationInput{
		TargetURL:           "https://example.com/downloads",
		DownloadURL:         "https://example.com/downloads/report.csv",
		DownloadedPath:      "data/tmp/browserops/report.csv",
		ScreenshotPath:      "data/tmp/browserops/download-page.png",
		FinalURL:            "https://example.com/downloads/report.csv",
		Status:              "downloaded",
		ContentType:         "text/csv; charset=utf-8",
		Bytes:               128,
		ExpectedFilename:    "report.csv",
		ExpectedContentType: "text/csv",
		URLContains:         "report.csv",
		MinBytes:            64,
		RequireFinalURL:     true,
		RequireScreenshot:   true,
	})
	if !result.Passed ||
		!result.EvidenceReady ||
		!result.DownloadedFileReady ||
		!result.ScreenshotReady ||
		!result.FinalURLReady ||
		!result.FilenameReady ||
		!result.ContentTypeReady ||
		!result.ByteSizeReady ||
		result.Score != 1 {
		t.Fatalf("expected download file evidence to pass, got %#v", result)
	}
	for _, expected := range []string{"downloaded_file_ready", "filename_ready:report.csv", "content_type_ready:text/csv", "byte_size_ready:64", "final_url_contains:report.csv"} {
		if !containsBrowserFailureReason(result.Evidence, expected) {
			t.Fatalf("expected evidence %q in %#v", expected, result.Evidence)
		}
	}
}

func TestEvaluateBrowserDownloadFileEvidenceFailsInvalidArtifact(t *testing.T) {
	result := EvaluateBrowserDownloadFileEvidence(BrowserDownloadFileEvaluationInput{
		DownloadedPath:      "data/tmp/browserops/report.tmp",
		Status:              "review_required",
		ContentType:         "text/plain",
		Bytes:               10,
		ExpectedFilename:    "report.csv",
		ExpectedContentType: "text/csv",
		MinBytes:            64,
	})
	if result.Passed || result.DownloadedFileReady || result.FilenameReady || result.ContentTypeReady || result.ByteSizeReady {
		t.Fatalf("expected invalid download file evidence to fail, got %#v", result)
	}
	for _, reason := range []string{
		"download_status_not_ready:review_required",
		"filename_mismatch:report.csv",
		"content_type_mismatch:text/csv",
		"byte_size_too_small:64",
	} {
		if !containsBrowserFailureReason(result.FailureReasons, reason) {
			t.Fatalf("expected reason %q in %#v", reason, result.FailureReasons)
		}
	}
}

func TestEvaluateBrowserDownloadFileEvidenceUsesEvidenceBundle(t *testing.T) {
	bundle := BuildBrowserEvidenceBundleFromState(map[string]any{
		"target_url": "https://example.com/downloads",
		"review": map[string]any{
			"evidence_path": "data:image/png;base64,AAAA",
			"final_url":     "https://example.com/downloads/report.csv",
		},
		"download": map[string]any{
			"path":         "data/tmp/browserops/report.csv",
			"bytes":        128,
			"content_type": "text/csv",
			"status":       "downloaded",
		},
	})
	result := EvaluateBrowserDownloadFileEvidence(BrowserDownloadFileEvaluationInput{
		Bundle:              bundle,
		ExpectedFilename:    "report.csv",
		ExpectedContentType: "text/csv",
		MinBytes:            64,
		RequireScreenshot:   true,
	})
	if !result.Passed || result.EvidenceBundle.DownloadedFile == nil || result.EvidenceBundle.Screenshot == nil {
		t.Fatalf("expected evidence bundle-backed download gate to pass, got %#v", result)
	}
}
