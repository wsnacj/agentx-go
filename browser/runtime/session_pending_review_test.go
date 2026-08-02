package browserruntime

import (
	"context"
	"testing"
	"time"
)

func TestSharedSessionBrowserPendingTargetReviewStateForTargetPrefersRouteMatch(t *testing.T) {
	registry := NewBrowserSessionRegistry()
	sessionID := "browser-runtime-pending-review-target-route-match"
	matchedRoute := BrowserSessionRoute{Backend: "proxy", Profile: "workbench", Target: "node-a", BrowserApp: "Chromium"}
	otherRoute := BrowserSessionRoute{Backend: "proxy", Profile: "workbench", Target: "node-b", BrowserApp: "Chromium"}

	otherPopup := registry.TrackTab(sessionID, BrowserSessionTarget{
		TabIndex:   2,
		URL:        "https://popup.example/other",
		Title:      "Other",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node-b",
	}, false)
	matchedPopup := registry.TrackTab(sessionID, BrowserSessionTarget{
		TabIndex:   2,
		URL:        "https://popup.example/matched",
		Title:      "Matched",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node-a",
	}, false)
	registry.RecordPendingTargetReviewForRoute(sessionID, otherRoute, BrowserSessionTargetReview{
		ID:         otherPopup.ID,
		TabIndex:   otherPopup.TabIndex,
		URL:        otherPopup.URL,
		Title:      otherPopup.Title,
		BrowserApp: otherPopup.BrowserApp,
		Backend:    otherPopup.Backend,
		Profile:    otherPopup.Profile,
		Target:     otherPopup.Target,
		Decision:   "session_target_popup_review_required",
		Reason:     "other pending popup",
	})
	registry.RecordPendingTargetReviewForRoute(sessionID, matchedRoute, BrowserSessionTargetReview{
		ID:         matchedPopup.ID,
		TabIndex:   matchedPopup.TabIndex,
		URL:        matchedPopup.URL,
		Title:      matchedPopup.Title,
		BrowserApp: matchedPopup.BrowserApp,
		Backend:    matchedPopup.Backend,
		Profile:    matchedPopup.Profile,
		Target:     matchedPopup.Target,
		Decision:   "session_target_popup_review_required",
		Reason:     "matched pending popup",
	})

	state := SharedSessionBrowserPendingTargetReviewStateForTarget(registry, sessionID, matchedRoute, matchedPopup.ID, 0)
	if state.Review == nil || state.Review.Target != "node-a" || state.PolicyState != "popup_review_required" {
		t.Fatalf("expected matched route pending review, got %#v", state)
	}
}

func TestSharedSessionBrowserPendingTargetReviewStateForRouteClearsAmbiguousFallback(t *testing.T) {
	registry := NewBrowserSessionRegistry()
	sessionID := "browser-runtime-pending-review-route-ambiguous"
	first := registry.TrackTab(sessionID, BrowserSessionTarget{
		TabIndex:   2,
		URL:        "https://popup.example/a",
		Title:      "Popup A",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node-a",
	}, false)
	second := registry.TrackTab(sessionID, BrowserSessionTarget{
		TabIndex:   3,
		URL:        "https://popup.example/b",
		Title:      "Popup B",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node-b",
	}, false)

	registry.RecordPendingTargetReviewForRoute(sessionID, BrowserSessionRoute{Backend: "proxy", Profile: "workbench", Target: "node-a", BrowserApp: "Chromium"}, BrowserSessionTargetReview{
		ID:         first.ID,
		TabIndex:   first.TabIndex,
		URL:        first.URL,
		Title:      first.Title,
		BrowserApp: first.BrowserApp,
		Backend:    first.Backend,
		Profile:    first.Profile,
		Target:     first.Target,
		Decision:   "session_target_popup_review_required",
		Reason:     "popup a",
	})
	registry.RecordPendingTargetReviewForRoute(sessionID, BrowserSessionRoute{Backend: "proxy", Profile: "workbench", Target: "node-b", BrowserApp: "Chromium"}, BrowserSessionTargetReview{
		ID:         second.ID,
		TabIndex:   second.TabIndex,
		URL:        second.URL,
		Title:      second.Title,
		BrowserApp: second.BrowserApp,
		Backend:    second.Backend,
		Profile:    second.Profile,
		Target:     second.Target,
		Decision:   "session_target_popup_review_required",
		Reason:     "popup b",
	})

	state := SharedSessionBrowserPendingTargetReviewStateForRoute(registry, sessionID, BrowserSessionRoute{Backend: "proxy", Profile: "workbench"})
	if state.Review != nil {
		t.Fatalf("expected ambiguous route review to clear fallback, got %#v", state)
	}
}

func TestSharedSessionBrowserAutoFollowPendingTargetReviewStateFallsBackToRoute(t *testing.T) {
	registry := NewBrowserSessionRegistry()
	sessionID := "browser-runtime-pending-review-auto-follow"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "workbench", Target: "node", BrowserApp: "Chromium"}
	popup := registry.TrackTab(sessionID, BrowserSessionTarget{
		TabIndex:   4,
		URL:        "https://popup.example/route",
		Title:      "Route Popup",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
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
		Decision:   "session_target_redirect_review_required",
		Reason:     "redirect review",
	})

	state := SharedSessionBrowserAutoFollowPendingTargetReviewState(registry, sessionID, route, "", 0)
	if state.Review == nil || state.Review.ID != popup.ID {
		t.Fatalf("expected auto-follow fallback to route review, got %#v", state)
	}
	if decision := SharedSessionBrowserPendingTargetReviewDecision(state, false); decision != "session_target_redirect_review_required" {
		t.Fatalf("unexpected pending review decision %q", decision)
	}
}

func TestRecordSharedSessionBrowserPendingTargetReviewResolvesTrackedTab(t *testing.T) {
	registry := NewBrowserSessionRegistry()
	sessionID := "browser-runtime-record-pending-review-tab"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "workbench", Target: "node", BrowserApp: "Chromium"}
	popup := registry.TrackTab(sessionID, BrowserSessionTarget{
		TabIndex:   5,
		URL:        "https://popup.example/review",
		Title:      "Popup Review",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, false)

	review := RecordSharedSessionBrowserPendingTargetReview(
		registry,
		sessionID,
		route,
		"",
		5,
		popup.URL,
		popup.Title,
		"session_target_popup_review_required",
		"pending popup review",
	)
	if review == nil || review.ID != popup.ID {
		t.Fatalf("expected pending review to resolve tracked tab target, got %#v", review)
	}

	state := SharedSessionBrowserPendingTargetReviewStateForTarget(registry, sessionID, route, popup.ID, 0)
	if state.Review == nil || state.Review.ID != popup.ID {
		t.Fatalf("expected recorded pending review to be visible in route state, got %#v", state)
	}
}

func TestRecordSharedSessionBrowserPendingTargetPopupReviewUsesActiveTabSelection(t *testing.T) {
	registry := NewBrowserSessionRegistry()
	sessionID := "browser-runtime-record-pending-popup-review"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "workbench", Target: "node", BrowserApp: "Chromium"}
	popup := registry.TrackTab(sessionID, BrowserSessionTarget{
		TabIndex:   7,
		URL:        "https://popup.example/offer",
		Title:      "Offer",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, false)

	review := RecordSharedSessionBrowserPendingTargetPopupReview(
		registry,
		sessionID,
		route,
		BrowserTab{
			Index:    popup.TabIndex,
			URL:      popup.URL,
			Title:    popup.Title,
			TargetID: popup.ID,
		},
		"session_target_popup_review_required",
		"pending popup review",
	)
	if review == nil || review.ID != popup.ID || review.TabIndex != popup.TabIndex {
		t.Fatalf("unexpected popup review record: %#v", review)
	}
}

func TestRecordSharedSessionBrowserPendingTargetPopupReviewWithContextSeedsSiblingProviderFromPrimaryCachedEventCycle(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	runRegistryA := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-record-pending-popup-review-helper-sibling-provider": {{RunID: "run-a", Status: "running"}},
		}},
	}
	runRegistryB := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-record-pending-popup-review-helper-sibling-provider": {{RunID: "run-b", Status: "running"}},
		}},
	}
	managerA := SharedSessionBrowserObserverManagerFor(sessionRegistry, runRegistryA, nil, time.Minute)
	managerB := SharedSessionBrowserObserverManagerFor(sessionRegistry, runRegistryB, nil, time.Minute)
	backend := &statusProfilesObservationTestBackend{
		statusResp: BrowserProfileStatusResult{
			Profile:   "isolated",
			Status:    "running",
			Running:   true,
			Connected: true,
		},
		profilesResp: BrowserProfilesResult{
			Backend:        "proxy",
			DefaultProfile: "isolated",
			Profiles: []BrowserProfileInfo{
				{Profile: "isolated", BrowserApp: "Chromium", Status: "connected", Running: true, Connected: true},
			},
		},
	}
	boundA := managerA.Bind(backend)
	boundB := managerB.Bind(backend)
	sessionID := "sess-record-pending-popup-review-helper-sibling-provider"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}
	mutationCtx := SharedSessionBrowserMutationContext{
		Registry:        sessionRegistry,
		RunRegistry:     runRegistryA,
		ReconnectWindow: time.Minute,
	}

	home := TrackSharedSessionBrowserTabWithContext(
		mutationCtx,
		sessionID,
		route,
		BrowserTab{Index: 1, URL: "https://example.com/home", Title: "Home"},
		true,
	)
	popup := TrackSharedSessionBrowserTabWithContext(
		mutationCtx,
		sessionID,
		route,
		BrowserTab{Index: 2, URL: "https://popup.example/offer", Title: "Offer"},
		false,
	)
	if home.ID == "" || popup.ID == "" {
		t.Fatalf("expected owner-aware helper setup to create tracked targets, got home=%#v popup=%#v", home, popup)
	}

	req := SharedSessionBrowserObserverRequest{
		SessionID:                   sessionID,
		SelectedInfo:                BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		BindingRoute:                route,
		RequestedProfile:            "isolated",
		IncludeStatus:               true,
		IncludeProfiles:             true,
		IncludeSessionView:          true,
		SessionViewInfo:             BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		SessionViewRouteFilter:      route,
		SessionViewRequestedProfile: "isolated",
	}

	initialA := boundA.ObserveWatchLoop(context.Background(), req)
	initialB := boundB.ObserveWatchLoop(context.Background(), req)
	if len(initialA.View.Session.Routes) != 1 || initialA.View.Session.Routes[0].CurrentTargetID != home.ID {
		t.Fatalf("expected first provider to expose tracked home target, got %#v", initialA.View.Session.Routes)
	}
	if len(initialB.View.Session.Routes) != 1 || initialB.View.Session.Routes[0].CurrentTargetID != home.ID {
		t.Fatalf("expected second provider to expose tracked home target, got %#v", initialB.View.Session.Routes)
	}
	if len(backend.statusReqs) != 2 || len(backend.profilesReqs) != 2 {
		t.Fatalf("expected initial sibling watch loops to poll backend once each, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}

	boundB.state.mu.Lock()
	clear(boundB.state.rawStatus)
	clear(boundB.state.rawProfiles)
	clear(boundB.state.eventCycles)
	clear(boundB.state.bindings)
	clear(boundB.state.views)
	clear(boundB.state.watchLoops)
	clear(boundB.state.eventCyclesInFlight)
	clear(boundB.state.bindingsInFlight)
	clear(boundB.state.viewsInFlight)
	clear(boundB.state.watchLoopsInFlight)
	boundB.state.mu.Unlock()

	review := RecordSharedSessionBrowserPendingTargetPopupReviewWithContext(
		mutationCtx,
		sessionID,
		route,
		BrowserTab{
			Index:    popup.TabIndex,
			URL:      popup.URL,
			Title:    popup.Title,
			TargetID: popup.ID,
		},
		"session_target_popup_review_required",
		"pending popup review",
	)
	if review == nil || review.ID != popup.ID || review.TabIndex != popup.TabIndex {
		t.Fatalf("expected owner-aware popup-review helper to record popup target, got %#v", review)
	}

	seededB := boundB.ObserveWatchLoop(context.Background(), req)
	if len(seededB.View.Session.Routes) != 1 || seededB.View.Session.Routes[0].CurrentTargetID != home.ID {
		t.Fatalf("expected helper-seeded popup review to preserve current target on sibling provider, got %#v", seededB.View.Session.Routes)
	}
	if seededB.View.Session.Routes[0].PendingTargetReview == nil || seededB.View.Session.Routes[0].PendingTargetReview.ID != popup.ID {
		t.Fatalf("expected sibling watch loop to expose helper-seeded pending popup review, got %#v", seededB.View.Session.Routes)
	}
	if len(backend.statusReqs) != 2 || len(backend.profilesReqs) != 2 {
		t.Fatalf("expected helper-seeded popup review to avoid extra polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}

func TestRecordSharedSessionBrowserPendingTargetReviewInvalidatesSharedWatchManagerCaches(t *testing.T) {
	registry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	backend := &statusProfilesObservationTestBackend{
		statusResp: BrowserProfileStatusResult{
			Profile:   "workbench",
			Status:    "running",
			Running:   true,
			Connected: true,
		},
		profilesResp: BrowserProfilesResult{
			Backend:        "proxy",
			DefaultProfile: "workbench",
			Profiles: []BrowserProfileInfo{
				{Profile: "workbench", BrowserApp: "Chromium", Status: "connected", Running: true, Connected: true},
			},
		},
	}
	manager := SharedSessionBrowserObserverManagerFor(registry, nil, stateRegistry, time.Minute)
	sessionID := "browser-runtime-pending-review-invalidates-watch"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "workbench", Target: "node", BrowserApp: "Chromium"}
	req := SharedSessionBrowserObserverRequest{
		SessionID:                   sessionID,
		SelectedInfo:                BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
		BindingRoute:                route,
		RequestedProfile:            "workbench",
		IncludeStatus:               true,
		IncludeProfiles:             true,
		IncludeSessionView:          true,
		SessionViewInfo:             BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
		SessionViewRouteFilter:      route,
		SessionViewRequestedProfile: "workbench",
	}

	popup := TrackSharedSessionBrowserTab(registry, sessionID, route, BrowserTab{
		Index: 3,
		URL:   "https://popup.example/review",
		Title: "Popup Review",
	}, false)
	first := manager.ObserveWatchLoop(context.Background(), backend, req)
	if len(first.View.Session.Routes) != 1 || first.View.Session.Routes[0].PendingTargetReview != nil || first.View.Session.Routes[0].PendingTargetReviewCount != 0 {
		t.Fatalf("expected initial watch loop without pending review, got %#v", first.View.Session.Routes)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected initial watch loop to poll backend once, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}

	review := RecordSharedSessionBrowserPendingTargetReview(
		registry,
		sessionID,
		route,
		popup.ID,
		popup.TabIndex,
		popup.URL,
		popup.Title,
		"session_target_popup_review_required",
		"pending popup review",
	)
	if review == nil || review.ID != popup.ID {
		t.Fatalf("expected pending review to be recorded, got %#v", review)
	}

	second := manager.ObserveWatchLoop(context.Background(), backend, req)
	if len(second.View.Session.Routes) != 1 || second.View.Session.Routes[0].PendingTargetReview == nil {
		t.Fatalf("expected pending review to invalidate cached watch loop, got %#v", second.View.Session.Routes)
	}
	if second.View.Session.Routes[0].PendingTargetReview.ID != popup.ID || second.View.Session.Routes[0].PendingTargetReviewCount != 1 {
		t.Fatalf("expected refreshed watch loop to expose pending review, got %#v", second.View.Session.Routes[0])
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected pending review invalidation to reuse cached raw status/profiles, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}
