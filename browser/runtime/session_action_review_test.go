package browserruntime

import "testing"

func TestSharedSessionBrowserActionRequiresReview(t *testing.T) {
	for _, kind := range []string{"download", "wait_download", "save_pdf", "save_html", "upload"} {
		if !SharedSessionBrowserActionRequiresReview(kind) {
			t.Fatalf("expected %s to require review", kind)
		}
	}
	for _, kind := range []string{"click", "type", "snapshot", ""} {
		if SharedSessionBrowserActionRequiresReview(kind) {
			t.Fatalf("expected %q not to require review", kind)
		}
	}
}

func TestSharedSessionBrowserActionReviewDecision(t *testing.T) {
	if got := SharedSessionBrowserActionReviewDecision(" download ", false); got != "download_review_required" {
		t.Fatalf("expected review required decision, got %q", got)
	}
	if got := SharedSessionBrowserActionReviewDecision("download", true); got != "download_review_confirmed" {
		t.Fatalf("expected review confirmed decision, got %q", got)
	}
	if got := SharedSessionBrowserActionReviewDecision(" ", true); got != "" {
		t.Fatalf("expected blank action decision to stay blank, got %q", got)
	}
}

func TestSharedSessionBrowserActionReviewReason(t *testing.T) {
	if got := SharedSessionBrowserActionReviewReason("download", false); got != "browser_act file downloads write remote content into the workspace; rerun with force=true after review" {
		t.Fatalf("expected download review reason, got %q", got)
	}
	if got := SharedSessionBrowserActionReviewReason("upload", true); got != "browser_act file upload review acknowledged via force=true; confirm the selected files are intended for outbound transfer" {
		t.Fatalf("expected upload acknowledged reason, got %q", got)
	}
	if got := SharedSessionBrowserActionReviewReason("dialog", false); got != "browser_act dialog accept may confirm a destructive or irreversible page action; rerun with force=true after review" {
		t.Fatalf("expected dialog review reason, got %q", got)
	}
	if got := SharedSessionBrowserActionReviewReason("click", false); got != "" {
		t.Fatalf("expected non-review action to have no review reason, got %q", got)
	}
}
