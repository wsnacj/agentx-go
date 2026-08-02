package tools

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBrowserActResolveOptionalArtifactOutputPath(t *testing.T) {
	root := t.TempDir()

	empty, err := browserActResolveOptionalArtifactOutputPath(root, "")
	if err != nil {
		t.Fatalf("browserActResolveOptionalArtifactOutputPath(empty): %v", err)
	}
	if empty.Resolved != "" || empty.Display != "" {
		t.Fatalf("expected empty optional output path, got %#v", empty)
	}

	resolved, err := browserActResolveArtifactOutputPath(root, "downloads/report.pdf")
	if err != nil {
		t.Fatalf("browserActResolveArtifactOutputPath: %v", err)
	}
	if resolved.Display != "downloads/report.pdf" {
		t.Fatalf("unexpected display path: %#v", resolved)
	}
	if _, err := os.Stat(filepath.Dir(resolved.Resolved)); !os.IsNotExist(err) {
		t.Fatalf("output resolution must not mutate the workspace, stat err=%v", err)
	}
}

func TestBrowserActWaitDownloadWaitMs(t *testing.T) {
	if got := browserActWaitDownloadWaitMs(2500, 1500); got != 2500 {
		t.Fatalf("browserActWaitDownloadWaitMs(explicit) = %d, want 2500", got)
	}
	if got := browserActWaitDownloadWaitMs(0, 1500); got != defaultBrowserWaitDownloadMs {
		t.Fatalf("browserActWaitDownloadWaitMs(default) = %d, want %d", got, defaultBrowserWaitDownloadMs)
	}
	if got := browserActWaitDownloadWaitMs(0, 180000); got != 180000 {
		t.Fatalf("browserActWaitDownloadWaitMs(max) = %d, want 180000", got)
	}
}

func TestBrowserActWaitDownloadWaitMsAcceptsTimeoutAlias(t *testing.T) {
	params := map[string]any{"timeout_ms": 30000}
	if got := browserActWaitDownloadWaitMs(firstInt(params, "wait_ms", "timeout_ms"), 1500); got != 30000 {
		t.Fatalf("browserActWaitDownloadWaitMs(timeout_ms alias) = %d, want 30000", got)
	}
}

func TestBrowserActTraceWaitMs(t *testing.T) {
	if got := browserActTraceWaitMs(1200, browserToolTarget{Explicit: true}); got != 1200 {
		t.Fatalf("browserActTraceWaitMs(explicit request) = %d, want 1200", got)
	}
	if got := browserActTraceWaitMs(0, browserToolTarget{Explicit: true}); got != browserTabTargetWaitMs {
		t.Fatalf("browserActTraceWaitMs(explicit target) = %d, want %d", got, browserTabTargetWaitMs)
	}
	if got := browserActTraceWaitMs(0, browserToolTarget{}); got != 0 {
		t.Fatalf("browserActTraceWaitMs(implicit target) = %d, want 0", got)
	}
}

func TestBrowserActReviewBlockedResultWithPath(t *testing.T) {
	result := browserActReviewBlockedResultWithPath(
		"save_pdf",
		BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
		"Chromium",
		browserToolTarget{Value: "tab:2", TabIndex: 2},
		false,
		"save requires review",
		"artifacts/report.pdf",
	)

	if result.Kind != "save_pdf" || result.Status != "review_required" || result.Path != "artifacts/report.pdf" {
		t.Fatalf("unexpected blocked result payload: %#v", result)
	}
	if result.ReviewDecision != "save_pdf_review_required" || result.TabIndex != 2 {
		t.Fatalf("unexpected blocked result review state: %#v", result)
	}
}
