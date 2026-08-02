package browserruntime

import "strings"

// SharedSessionBrowserActionRequiresReview reports whether a browser action
// must stop at the review posture before it can execute.
func SharedSessionBrowserActionRequiresReview(kind string) bool {
	switch strings.TrimSpace(kind) {
	case "download", "wait_download", "save_pdf", "save_html", "upload":
		return true
	default:
		return false
	}
}

// SharedSessionBrowserActionReviewDecision resolves the shared decision token
// for browser action review posture.
func SharedSessionBrowserActionReviewDecision(kind string, force bool) string {
	base := strings.TrimSpace(kind)
	if base == "" {
		return ""
	}
	if force {
		return base + "_review_confirmed"
	}
	return base + "_review_required"
}

// SharedSessionBrowserActionReviewReason returns the shared review explanation
// for browser actions that require explicit acknowledgement.
func SharedSessionBrowserActionReviewReason(kind string, force bool) string {
	switch strings.TrimSpace(kind) {
	case "download", "wait_download":
		if force {
			return "browser_act file download review acknowledged via force=true; compare any downloaded artifact against the task boundary before further use"
		}
		return "browser_act file downloads write remote content into the workspace; rerun with force=true after review"
	case "upload":
		if force {
			return "browser_act file upload review acknowledged via force=true; confirm the selected files are intended for outbound transfer"
		}
		return "browser_act file uploads send local workspace content to the active page; rerun with force=true after review"
	case "save_pdf", "save_html":
		if force {
			return "browser_act page export review acknowledged via force=true; confirm the page artifact is intended to be persisted into the workspace"
		}
		return "browser_act page export writes remote page content into the workspace; rerun with force=true after review"
	case "dialog":
		if force {
			return "browser_act dialog acceptance review acknowledged via force=true; confirm the modal action is intended before continuing"
		}
		return "browser_act dialog accept may confirm a destructive or irreversible page action; rerun with force=true after review"
	default:
		return ""
	}
}
