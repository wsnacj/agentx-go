package browserruntime

import "testing"

func TestSnapshotSharedSessionBrowserRoutesProjectsPendingReviewPolicy(t *testing.T) {
	registry := NewBrowserSessionRegistry()
	sessionID := "s1"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "workbench", Target: "node", BrowserApp: "Chromium"}
	home := registry.TrackTab(sessionID, BrowserSessionTarget{
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
		TabIndex:   1,
		URL:        "https://example.com",
		Title:      "Home",
		BrowserApp: "Chromium",
	}, true)
	popup := registry.TrackTab(sessionID, BrowserSessionTarget{
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
		TabIndex:   2,
		URL:        "https://example.com/popup",
		Title:      "Popup",
		BrowserApp: "Chromium",
	}, false)
	registry.RecordPendingTargetReviewForRoute(sessionID, route, BrowserSessionTargetReview{
		ID:         popup.ID,
		TabIndex:   popup.TabIndex,
		URL:        popup.URL,
		Title:      popup.Title,
		BrowserApp: popup.BrowserApp,
		Backend:    popup.Backend,
		Profile:    popup.Profile,
		Target:     popup.Target,
		Decision:   "session_target_popup_review_required",
		Reason:     "newly opened active tab",
	})
	routes := SnapshotSharedSessionBrowserRoutes(registry, sessionID, BrowserSessionRoute{Backend: "proxy", Target: "node"})
	if len(routes) != 1 {
		t.Fatalf("expected single projected route, got %#v", routes)
	}
	if routes[0].CurrentTargetID != home.ID || routes[0].PendingTargetReview == nil || routes[0].PendingTargetReview.ID != popup.ID {
		t.Fatalf("unexpected route snapshot: %#v", routes[0])
	}
	if routes[0].FollowPolicyState != "popup_review_required" || routes[0].PopupPolicyState != "popup_review_required" {
		t.Fatalf("expected popup review posture, got %#v", routes[0])
	}
	if len(routes[0].Targets) != 2 || !routes[0].Targets[0].Current {
		t.Fatalf("expected projected targets with current marker, got %#v", routes[0].Targets)
	}
}

func TestSnapshotSharedSessionBrowserRoutesUsesRedirectFollowPolicyWithoutPopupPosture(t *testing.T) {
	registry := NewBrowserSessionRegistry()
	sessionID := "s1"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "workbench", Target: "node", BrowserApp: "Chromium"}
	redirected := registry.TrackTab(sessionID, BrowserSessionTarget{
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
		TabIndex:   3,
		URL:        "https://example.com/redirect",
		Title:      "Redirected",
		BrowserApp: "Chromium",
	}, false)
	registry.RecordPendingTargetReviewForRoute(sessionID, route, BrowserSessionTargetReview{
		ID:         redirected.ID,
		TabIndex:   redirected.TabIndex,
		URL:        redirected.URL,
		Title:      redirected.Title,
		BrowserApp: redirected.BrowserApp,
		Backend:    redirected.Backend,
		Profile:    redirected.Profile,
		Target:     redirected.Target,
		Decision:   "session_target_redirect_review_required",
		Reason:     "cross-origin redirect",
	})
	routes := SnapshotSharedSessionBrowserRoutes(registry, sessionID, route)
	if len(routes) != 1 {
		t.Fatalf("expected single projected route, got %#v", routes)
	}
	if routes[0].FollowPolicyState != "redirect_review_required" || routes[0].PopupPolicyState != "" {
		t.Fatalf("expected redirect follow policy without popup posture, got %#v", routes[0])
	}
}
