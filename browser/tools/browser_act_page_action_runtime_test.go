package tools

import (
	"testing"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

func TestBrowserActPageActionWaitMs(t *testing.T) {
	t.Run("explicit request wins", func(t *testing.T) {
		if got := browserActPageActionWaitMs("https://93.184.216.34", 321, 1500, 250); got != 321 {
			t.Fatalf("browserActPageActionWaitMs(explicit) = %d, want 321", got)
		}
	})

	t.Run("url request uses url wait", func(t *testing.T) {
		if got := browserActPageActionWaitMs("https://93.184.216.34", 0, 1500, 250); got != 1500 {
			t.Fatalf("browserActPageActionWaitMs(url) = %d, want 1500", got)
		}
	})

	t.Run("target request uses target wait", func(t *testing.T) {
		if got := browserActPageActionWaitMs("", 0, 1500, 250); got != 250 {
			t.Fatalf("browserActPageActionWaitMs(target) = %d, want 250", got)
		}
	})
}

func TestBrowserActInteractivePageActionWaitMs(t *testing.T) {
	t.Run("explicit request wins", func(t *testing.T) {
		if got := browserActInteractivePageActionWaitMs("https://93.184.216.34", 432, 1500, browserToolTarget{}); got != 432 {
			t.Fatalf("browserActInteractivePageActionWaitMs(explicit) = %d, want 432", got)
		}
	})

	t.Run("url request uses default wait", func(t *testing.T) {
		if got := browserActInteractivePageActionWaitMs("https://93.184.216.34", 0, 1500, browserToolTarget{}); got != defaultBrowserInteractiveActionWaitMs {
			t.Fatalf("browserActInteractivePageActionWaitMs(url) = %d, want %d", got, defaultBrowserInteractiveActionWaitMs)
		}
	})

	t.Run("explicit target uses tab wait", func(t *testing.T) {
		if got := browserActInteractivePageActionWaitMs("", 0, 1500, browserToolTarget{Explicit: true}); got != defaultBrowserInteractiveActionWaitMs {
			t.Fatalf("browserActInteractivePageActionWaitMs(target) = %d, want %d", got, defaultBrowserInteractiveActionWaitMs)
		}
	})

	t.Run("implicit target uses interactive floor", func(t *testing.T) {
		if got := browserActInteractivePageActionWaitMs("", 0, 1500, browserToolTarget{}); got != defaultBrowserInteractiveActionWaitMs {
			t.Fatalf("browserActInteractivePageActionWaitMs(implicit) = %d, want %d", got, defaultBrowserInteractiveActionWaitMs)
		}
	})
}

func TestBrowserActPageReviewBlockedResultWithPath(t *testing.T) {
	reviewState := browserActPageReviewState{
		Review: browserPendingTargetReviewState{
			Review: &agentxbrowserruntime.BrowserSessionTargetReview{
				Decision: "session_target_popup_review_required",
			},
		},
		TargetID: "popup-target",
	}

	result := browserActPageReviewBlockedResultWithPath(
		"screenshot",
		BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
		"Chromium",
		browserToolTarget{Value: "tab:3", TabIndex: 3},
		reviewState,
		false,
		"pending popup target review blocks action",
		"shots/popup.png",
	)

	if result.Kind != "screenshot" || result.Status != "review_required" || result.Path != "shots/popup.png" {
		t.Fatalf("unexpected blocked result payload: %#v", result)
	}
	if result.TargetID != "popup-target" || result.ReviewDecision != "session_target_popup_review_required" || result.TabIndex != 3 {
		t.Fatalf("unexpected blocked result review state: %#v", result)
	}
}

func TestBrowserActPageReviewBlockedResultWithFields(t *testing.T) {
	reviewState := browserActPageReviewState{
		Review: browserPendingTargetReviewState{
			Review: &agentxbrowserruntime.BrowserSessionTargetReview{
				Decision: "session_target_popup_review_required",
			},
		},
		TargetID: "popup-target",
	}

	result := browserActPageReviewBlockedResultWithFields(
		"type",
		BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
		"Chromium",
		browserToolTarget{Value: "tab:3", TabIndex: 3},
		reviewState,
		false,
		"pending popup target review blocks action",
		"ref://search",
		"#search",
		"agentx",
	)

	if result.Kind != "type" || result.Status != "review_required" || result.Ref != "ref://search" || result.Selector != "#search" || result.Value != "agentx" {
		t.Fatalf("unexpected blocked result fields payload: %#v", result)
	}
	if result.TargetID != "popup-target" || result.ReviewDecision != "session_target_popup_review_required" || result.TabIndex != 3 {
		t.Fatalf("unexpected blocked result fields review state: %#v", result)
	}
}
