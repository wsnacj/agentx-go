package browserruntime

import "testing"

func TestBuildSharedSessionBrowserSessionHandoffSummaryReady(t *testing.T) {
	summary := BuildSharedSessionBrowserSessionHandoffSummary(SharedSessionBrowserSessionHandoffRequest{
		Routes: []SharedSessionBrowserRouteSnapshot{{
			Backend:             "proxy",
			Profile:             "workbench",
			RuntimeTarget:       "node",
			BrowserApp:          "Chromium",
			CurrentTargetID:     "tab-2",
			CurrentTargetSource: "tracked_active_tab",
			Targets: []SharedSessionBrowserRouteTarget{
				{ID: "tab-1", TabIndex: 1, URL: "https://example.com/1"},
				{ID: "tab-2", TabIndex: 2, URL: "https://example.com/2", Title: "Example", Current: true},
			},
		}},
		Runs: []SharedSessionRunInfo{{RunID: "run-1", Status: "running"}},
		Profiles: []SharedSessionBrowserProjectedProfileState{{
			State: SharedSessionBrowserProfileState{
				Backend:       "proxy",
				Profile:       "workbench",
				RuntimeTarget: "node",
				BrowserApp:    "Chromium",
				Status:        "running",
				Running:       true,
				Connected:     true,
			},
			Selected: true,
		}},
		Health: &SharedSessionBrowserHealthSummary{State: "healthy"},
	})

	if summary == nil {
		t.Fatalf("expected handoff summary")
	}
	if summary.State != SharedSessionBrowserSessionHandoffStateReady || summary.NextStepAlias != "continue_current_target" {
		t.Fatalf("expected ready handoff, got %#v", summary)
	}
	if summary.CurrentTarget == nil || summary.CurrentTarget.ID != "tab-2" || summary.CurrentTarget.URL != "https://example.com/2" {
		t.Fatalf("expected current target identity in handoff, got %#v", summary.CurrentTarget)
	}
	if summary.TargetCount != 2 || summary.RouteCount != 1 || summary.ActiveRunID != "run-1" || summary.ActiveRunCount != 1 {
		t.Fatalf("expected compact route/run counters, got %#v", summary)
	}
	if summary.SelectedProfile != "workbench" || summary.SelectedProfileStatus != "running" || summary.BrowserApp != "Chromium" {
		t.Fatalf("expected selected profile identity in handoff, got %#v", summary)
	}
}

func TestBuildSharedSessionBrowserSessionHandoffSummaryReviewBeatsHealth(t *testing.T) {
	summary := BuildSharedSessionBrowserSessionHandoffSummary(SharedSessionBrowserSessionHandoffRequest{
		Routes: []SharedSessionBrowserRouteSnapshot{{
			Backend:                  "proxy",
			Profile:                  "workbench",
			RuntimeTarget:            "node",
			CurrentTargetID:          "tab-1",
			PendingTargetReviewCount: 2,
			PendingTargetReview: &BrowserSessionTargetReview{
				ID:       "tab-2",
				TabIndex: 2,
				URL:      "https://review.example",
				Title:    "Review target",
				Decision: "review_required",
			},
		}},
		Health: &SharedSessionBrowserHealthSummary{
			State:         "profile_reconnecting",
			NextStepAlias: "wait_for_reconnect",
		},
	})

	if summary == nil {
		t.Fatalf("expected handoff summary")
	}
	if summary.State != SharedSessionBrowserSessionHandoffStateTargetReviewRequired || summary.NextStepAlias != "review_target" {
		t.Fatalf("expected target review to be the handoff blocker, got %#v", summary)
	}
	if summary.PendingTargetReviewCount != 2 || summary.PendingTargetReview == nil || summary.PendingTargetReview.ID != "tab-2" {
		t.Fatalf("expected pending review summary, got %#v", summary)
	}
}

func TestBuildSharedSessionBrowserSessionHandoffSummaryHealthAttention(t *testing.T) {
	summary := BuildSharedSessionBrowserSessionHandoffSummary(SharedSessionBrowserSessionHandoffRequest{
		Routes: []SharedSessionBrowserRouteSnapshot{{
			Backend:         "proxy",
			Profile:         "workbench",
			RuntimeTarget:   "node",
			CurrentTargetID: "tab-1",
			Targets:         []SharedSessionBrowserRouteTarget{{ID: "tab-1", Current: true}},
		}},
		Health: &SharedSessionBrowserHealthSummary{
			State:         "profile_reconnecting",
			Reason:        "cdp reconnect in progress",
			NextStepAlias: "wait_for_reconnect",
		},
	})

	if summary == nil {
		t.Fatalf("expected handoff summary")
	}
	if summary.State != SharedSessionBrowserSessionHandoffStateHealthAttention || summary.NextStepAlias != "wait_for_reconnect" {
		t.Fatalf("expected health attention handoff, got %#v", summary)
	}
	if summary.CurrentTarget == nil || summary.CurrentTarget.ID != "tab-1" {
		t.Fatalf("expected handoff to keep current target while health needs attention, got %#v", summary.CurrentTarget)
	}
}

func TestBuildSharedSessionBrowserSessionHandoffSummaryEmpty(t *testing.T) {
	if summary := BuildSharedSessionBrowserSessionHandoffSummary(SharedSessionBrowserSessionHandoffRequest{}); summary != nil {
		t.Fatalf("expected empty handoff input to stay nil, got %#v", summary)
	}
}
