package browserruntime

import (
	"context"
	"testing"
	"time"
)

func TestSyncSharedSessionBrowserCurrentTargetTracksSource(t *testing.T) {
	registry := NewBrowserSessionRegistry()
	sessionID := "browser-runtime-sync-current-target-source"
	tracked := registry.TrackTabs(sessionID, []BrowserSessionTarget{{
		TabIndex:   1,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}}, 1)
	route := BrowserSessionRoute{Backend: "proxy", Target: "node", Profile: "workbench"}

	first, decision, err := SyncSharedSessionBrowserCurrentTarget(registry, sessionID, route, "sync_session")
	if err != nil {
		t.Fatalf("SyncSharedSessionBrowserCurrentTarget returned error: %v", err)
	}
	if decision != "session_target_selected" || first == nil || first.ID != tracked[0].ID || first.Source != "sync_session" {
		t.Fatalf("unexpected first sync result: selection=%#v decision=%q", first, decision)
	}

	second, decision, err := SyncSharedSessionBrowserCurrentTarget(registry, sessionID, route, "sync_session")
	if err != nil {
		t.Fatalf("SyncSharedSessionBrowserCurrentTarget second call returned error: %v", err)
	}
	if decision != "session_target_already_selected" || second == nil || second.ID != tracked[0].ID || second.Source != "sync_session" {
		t.Fatalf("unexpected second sync result: selection=%#v decision=%q", second, decision)
	}
}

func TestShouldClearSharedSessionBrowserTargetOnProfileSelectForMismatch(t *testing.T) {
	registry := NewBrowserSessionRegistry()
	sessionID := "browser-runtime-sync-clear-mismatch"
	registry.TrackTabs(sessionID, []BrowserSessionTarget{{
		TabIndex:   1,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}}, 1)
	route := BrowserSessionRoute{Backend: "proxy", Target: "node"}

	if !ShouldClearSharedSessionBrowserTargetOnProfileSelect(
		registry,
		sessionID,
		route,
		&SharedSessionBrowserProfileSelection{
			Backend:       "proxy",
			Profile:       "alternate",
			RuntimeTarget: "node",
			Source:        "sync_session",
		},
	) {
		t.Fatalf("expected mismatched profile selection to request target clear")
	}
}

func TestShouldClearSharedSessionBrowserProfileOnTargetClearForRememberedSelection(t *testing.T) {
	registry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	sessionID := "browser-runtime-clear-profile-on-target-clear"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "workbench", Target: "node"}

	tracked := registry.TrackTabs(sessionID, []BrowserSessionTarget{{
		TabIndex:   1,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}}, 1)
	if _, ok := registry.SelectTargetForRoute(sessionID, route, tracked[0].ID, "remember_profile"); !ok {
		t.Fatalf("expected target selection to succeed")
	}
	stateRegistry.SelectBrowserProfile(sessionID, SharedSessionBrowserProfileSelection{
		Backend:       "proxy",
		Profile:       "workbench",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Source:        "remember_profile",
	})

	if !ShouldClearSharedSessionBrowserProfileOnTargetClear(registry, stateRegistry, sessionID, route) {
		t.Fatalf("expected remembered profile selection to clear with remembered target")
	}
}

func TestSelectSharedSessionBrowserTargetRequiresPendingPopupReview(t *testing.T) {
	registry := NewBrowserSessionRegistry()
	sessionID := "browser-runtime-select-target-popup-review-shared"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "workbench", Target: "node", BrowserApp: "Chromium"}
	current := registry.TrackTab(sessionID, BrowserSessionTarget{
		TabIndex:   1,
		URL:        "https://node.example/current",
		Title:      "Current",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, true)
	popup := registry.TrackTab(sessionID, BrowserSessionTarget{
		TabIndex:   2,
		URL:        "https://popup.example/offer",
		Title:      "Offer",
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
		Decision:   "session_target_popup_review_required",
		Reason:     "pending popup review",
	})

	result, err := SelectSharedSessionBrowserTarget(registry, SharedSessionBrowserSelectTargetRequest{
		SessionID: sessionID,
		Route:     route,
		TargetID:  popup.ID,
		Source:    "select_target",
		Actor:     "browser_runtime target selection",
	})
	if err != nil {
		t.Fatalf("SelectSharedSessionBrowserTarget returned error: %v", err)
	}
	if result.Selection != nil || result.Decision != "session_target_popup_review_required" || result.Ready {
		t.Fatalf("unexpected blocked popup review result: %#v", result)
	}
	if got, ok := registry.CurrentTargetForRoute(sessionID, route); !ok || got.ID != current.ID {
		t.Fatalf("expected current route target to remain unchanged, got %#v ok=%v", got, ok)
	}
}

func TestSelectSharedSessionBrowserTargetConfirmsPopupReviewWithForce(t *testing.T) {
	registry := NewBrowserSessionRegistry()
	sessionID := "browser-runtime-select-target-popup-force-shared"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "workbench", Target: "node", BrowserApp: "Chromium"}
	registry.TrackTab(sessionID, BrowserSessionTarget{
		TabIndex:   1,
		URL:        "https://node.example/current",
		Title:      "Current",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, true)
	popup := registry.TrackTab(sessionID, BrowserSessionTarget{
		TabIndex:   2,
		URL:        "https://popup.example/offer",
		Title:      "Offer",
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
		Decision:   "session_target_popup_review_required",
		Reason:     "pending popup review",
	})

	result, err := SelectSharedSessionBrowserTarget(registry, SharedSessionBrowserSelectTargetRequest{
		SessionID: sessionID,
		Route:     route,
		TargetID:  popup.ID,
		Force:     true,
		Source:    "select_target",
		Actor:     "browser_runtime target selection",
	})
	if err != nil {
		t.Fatalf("SelectSharedSessionBrowserTarget returned error: %v", err)
	}
	if result.Selection == nil || result.Selection.ID != popup.ID || result.Decision != "session_target_popup_review_confirmed" || !result.Ready {
		t.Fatalf("unexpected forced popup selection result: %#v", result)
	}
	if got, ok := registry.CurrentTargetForRoute(sessionID, route); !ok || got.ID != popup.ID {
		t.Fatalf("expected forced popup review to promote target, got %#v ok=%v", got, ok)
	}
}

func TestSelectSharedSessionBrowserTargetFallsBackToResolvedRoute(t *testing.T) {
	registry := NewBrowserSessionRegistry()
	sessionID := "browser-runtime-select-target-fallback-route"
	tracked := registry.TrackTab(sessionID, BrowserSessionTarget{
		TabIndex:   4,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, false)

	result, err := SelectSharedSessionBrowserTarget(registry, SharedSessionBrowserSelectTargetRequest{
		SessionID: sessionID,
		Route:     BrowserSessionRoute{Target: "node"},
		TargetID:  tracked.ID,
		Source:    "select_target",
	})
	if err != nil {
		t.Fatalf("SelectSharedSessionBrowserTarget returned error: %v", err)
	}
	if result.Selection == nil || result.Selection.ID != tracked.ID || result.Decision != "session_target_selected" || !result.Ready {
		t.Fatalf("unexpected resolved-route fallback result: %#v", result)
	}
	if got, ok := registry.CurrentTargetForRoute(sessionID, BrowserSessionRoute{Backend: "proxy", Profile: "workbench", Target: "node", BrowserApp: "Chromium"}); !ok || got.ID != tracked.ID {
		t.Fatalf("expected fallback route selection to persist on resolved route, got %#v ok=%v", got, ok)
	}
}

func TestRememberSharedSessionBrowserTargetPrefersRouteAndFallsBackToDefault(t *testing.T) {
	registry := NewBrowserSessionRegistry()
	sessionID := "browser-runtime-remember-target-route-fallback"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "workbench", Target: "node", BrowserApp: "Chromium"}
	tracked := registry.TrackTab(sessionID, BrowserSessionTarget{
		TabIndex:   3,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, false)

	result := RememberSharedSessionBrowserTarget(registry, SharedSessionBrowserRememberTargetRequest{
		SessionID: sessionID,
		Route:     route,
		TabIndex:  tracked.TabIndex,
		Source:    "remember_target",
	})
	if result.Selection == nil || result.Selection.ID != tracked.ID || result.Decision != "session_target_remembered" || !result.Ready {
		t.Fatalf("unexpected remembered target result: %#v", result)
	}

	fallback := RememberSharedSessionBrowserTarget(registry, SharedSessionBrowserRememberTargetRequest{
		SessionID: sessionID,
		Route:     BrowserSessionRoute{Target: "node"},
		TargetID:  tracked.ID,
		Source:    "remember_target",
	})
	if fallback.Selection == nil || fallback.Selection.ID != tracked.ID || !fallback.Ready {
		t.Fatalf("unexpected fallback remembered target result: %#v", fallback)
	}
	if fallback.Decision != "session_target_remembered" && fallback.Decision != "session_target_already_remembered" {
		t.Fatalf("unexpected fallback remember decision: %#v", fallback)
	}
}

func TestRememberSharedSessionBrowserTargetReusesCurrentSelection(t *testing.T) {
	registry := NewBrowserSessionRegistry()
	sessionID := "browser-runtime-remember-target-current"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "workbench", Target: "node", BrowserApp: "Chromium"}
	tracked := registry.TrackTab(sessionID, BrowserSessionTarget{
		TabIndex:   2,
		URL:        "https://node.example/current",
		Title:      "Current",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, true)

	result := RememberSharedSessionBrowserTarget(registry, SharedSessionBrowserRememberTargetRequest{
		SessionID: sessionID,
		Route:     route,
		TargetID:  tracked.ID,
		Source:    "remember_target",
	})
	if result.Selection == nil || result.Selection.ID != tracked.ID || result.Decision != "session_target_already_remembered" || !result.Ready {
		t.Fatalf("unexpected current remembered target result: %#v", result)
	}
}

func TestDispatchSharedSessionBrowserRememberTargetPromotesProfileSelection(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	sessionID := "browser-runtime-remember-target-dispatch"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "workbench", Target: "node", BrowserApp: "Chromium"}
	tracked := sessionRegistry.TrackTab(sessionID, BrowserSessionTarget{
		TabIndex:   2,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, false)

	result := DispatchSharedSessionBrowserRememberTarget(
		SharedSessionBrowserRememberTargetDispatchRequest{
			MutationContext: SharedSessionBrowserMutationContext{
				Registry:      sessionRegistry,
				StateRegistry: stateRegistry,
			},
			SessionID: sessionID,
			Route:     route,
			TargetID:  tracked.ID,
			Source:    "remember_target",
		},
	)
	if result.Selection == nil || result.Selection.ID != tracked.ID || result.Decision != "session_target_remembered" || !result.Ready {
		t.Fatalf("unexpected remember-target dispatch result: %#v", result)
	}
	if result.ProfileSelection == nil || result.ProfileSelection.Profile != "workbench" || result.ProfileSelection.RuntimeTarget != "node" || result.ProfileSelection.Source != "remember_target" {
		t.Fatalf("expected remember-target dispatch helper to promote profile selection, got %#v", result.ProfileSelection)
	}
}

func TestSelectSharedSessionBrowserTargetWithContextSeedsSiblingProviderFromPrimaryCachedEventCycle(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	runRegistryA := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-select-target-helper-sibling-provider": {{RunID: "run-a", Status: "running"}},
		}},
	}
	runRegistryB := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-select-target-helper-sibling-provider": {{RunID: "run-b", Status: "running"}},
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
	sessionID := "sess-select-target-helper-sibling-provider"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}
	mutationCtx := SharedSessionBrowserMutationContext{
		Registry:        sessionRegistry,
		RunRegistry:     runRegistryA,
		ReconnectWindow: time.Minute,
	}

	firstTarget := sessionRegistry.TrackTab(sessionID, BrowserSessionTarget{
		ID:         "tab-1",
		TabIndex:   1,
		URL:        "https://example.com/one",
		Title:      "One",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
	}, true)
	secondTarget := sessionRegistry.TrackTab(sessionID, BrowserSessionTarget{
		ID:         "tab-2",
		TabIndex:   2,
		URL:        "https://example.com/two",
		Title:      "Two",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
	}, false)

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
	if len(initialA.View.Session.Routes) != 1 || initialA.View.Session.Routes[0].CurrentTargetID != firstTarget.ID {
		t.Fatalf("expected first provider to expose first current target, got %#v", initialA.View.Session.Routes)
	}
	if len(initialB.View.Session.Routes) != 1 || initialB.View.Session.Routes[0].CurrentTargetID != firstTarget.ID {
		t.Fatalf("expected second provider to expose first current target, got %#v", initialB.View.Session.Routes)
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

	result, err := SelectSharedSessionBrowserTargetWithContext(
		mutationCtx,
		SharedSessionBrowserSelectTargetRequest{
			SessionID: sessionID,
			Route:     route,
			TargetID:  secondTarget.ID,
			Source:    "select_target",
			Actor:     "helper select target",
		},
	)
	if err != nil {
		t.Fatalf("SelectSharedSessionBrowserTargetWithContext returned error: %v", err)
	}
	if !result.Ready || result.Selection == nil || result.Selection.ID != secondTarget.ID || result.Selection.Source != "select_target" {
		t.Fatalf("expected owner-aware select-target helper to pick second target, got %#v", result)
	}

	seededB := boundB.ObserveWatchLoop(context.Background(), req)
	if len(seededB.View.Session.Routes) != 1 || seededB.View.Session.Routes[0].CurrentTargetID != secondTarget.ID || seededB.View.Session.Routes[0].CurrentTargetSource != "select_target" {
		t.Fatalf("expected sibling watch loop to reuse helper-seeded selected target, got %#v", seededB.View.Session.Routes)
	}
	if len(backend.statusReqs) != 2 || len(backend.profilesReqs) != 2 {
		t.Fatalf("expected helper-seeded sibling watch loop to avoid extra polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}

func TestRememberSharedSessionBrowserTargetWithContextSeedsSiblingProviderFromPrimaryCachedEventCycle(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	runRegistryA := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-remember-target-helper-sibling-provider": {{RunID: "run-a", Status: "running"}},
		}},
	}
	runRegistryB := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-remember-target-helper-sibling-provider": {{RunID: "run-b", Status: "running"}},
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
	sessionID := "sess-remember-target-helper-sibling-provider"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}
	mutationCtx := SharedSessionBrowserMutationContext{
		Registry:        sessionRegistry,
		RunRegistry:     runRegistryA,
		ReconnectWindow: time.Minute,
	}

	firstTarget := sessionRegistry.TrackTab(sessionID, BrowserSessionTarget{
		ID:         "tab-1",
		TabIndex:   1,
		URL:        "https://example.com/one",
		Title:      "One",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
	}, true)
	secondTarget := sessionRegistry.TrackTab(sessionID, BrowserSessionTarget{
		ID:         "tab-2",
		TabIndex:   2,
		URL:        "https://example.com/two",
		Title:      "Two",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
	}, false)

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
	if len(initialA.View.Session.Routes) != 1 || initialA.View.Session.Routes[0].CurrentTargetID != firstTarget.ID {
		t.Fatalf("expected first provider to expose first current target, got %#v", initialA.View.Session.Routes)
	}
	if len(initialB.View.Session.Routes) != 1 || initialB.View.Session.Routes[0].CurrentTargetID != firstTarget.ID {
		t.Fatalf("expected second provider to expose first current target, got %#v", initialB.View.Session.Routes)
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

	result := RememberSharedSessionBrowserTargetWithContext(
		mutationCtx,
		SharedSessionBrowserRememberTargetRequest{
			SessionID: sessionID,
			Route:     route,
			TargetID:  secondTarget.ID,
			Source:    "remember_target",
		},
	)
	if !result.Ready || result.Selection == nil || result.Selection.ID != secondTarget.ID || result.Selection.Source != "remember_target" {
		t.Fatalf("expected owner-aware remember-target helper to pick second target, got %#v", result)
	}

	seededB := boundB.ObserveWatchLoop(context.Background(), req)
	if len(seededB.View.Session.Routes) != 1 || seededB.View.Session.Routes[0].CurrentTargetID != secondTarget.ID || seededB.View.Session.Routes[0].CurrentTargetSource != "remember_target" {
		t.Fatalf("expected sibling watch loop to reuse helper-seeded remembered target, got %#v", seededB.View.Session.Routes)
	}
	if len(backend.statusReqs) != 2 || len(backend.profilesReqs) != 2 {
		t.Fatalf("expected helper-seeded sibling watch loop to avoid extra polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}
