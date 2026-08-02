package tools

import "testing"

func TestBrowserActDialogWaitMs(t *testing.T) {
	if got := browserActDialogWaitMs(2500, 1500); got != 2500 {
		t.Fatalf("browserActDialogWaitMs(explicit) = %d, want 2500", got)
	}
	if got := browserActDialogWaitMs(0, 1500); got != defaultBrowserWaitDownloadMs {
		t.Fatalf("browserActDialogWaitMs(default) = %d, want %d", got, defaultBrowserWaitDownloadMs)
	}
	if got := browserActDialogWaitMs(0, 180000); got != 180000 {
		t.Fatalf("browserActDialogWaitMs(max) = %d, want 180000", got)
	}
}

func TestBrowserActDialogReviewState(t *testing.T) {
	decision, ready := browserActDialogReviewState("accept", true)
	if decision != "dialog_review_confirmed" || !ready {
		t.Fatalf("browserActDialogReviewState(accept,true) = (%q,%v)", decision, ready)
	}
	decision, ready = browserActDialogReviewState("dismiss", false)
	if decision != "" || ready {
		t.Fatalf("browserActDialogReviewState(dismiss,false) = (%q,%v)", decision, ready)
	}
}
