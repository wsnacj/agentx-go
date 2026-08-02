package browserruntime

import (
	"context"
	"testing"
	"time"
)

func TestSyncSharedSessionBrowserTabsForRouteProjectsTrackedTargetIDs(t *testing.T) {
	registry := NewBrowserSessionRegistry()
	route := BrowserSessionRoute{Backend: "proxy", Profile: "work", Target: "node", BrowserApp: "Chromium"}

	tabs := SyncSharedSessionBrowserTabsForRoute(registry, "sess-1", route, 2, []BrowserTab{
		{Index: 1, URL: "https://example.com/1", Title: "One"},
		{Index: 2, URL: "https://example.com/2", Title: "Two"},
	})

	if len(tabs) != 2 || tabs[0].TargetID == "" || tabs[1].TargetID == "" {
		t.Fatalf("expected tracked tabs with target IDs, got %#v", tabs)
	}
	current := CurrentSharedSessionBrowserTargetSelection(registry, "sess-1", route)
	if current == nil || current.ID != tabs[1].TargetID || current.Source != "tracked_active_tab" {
		t.Fatalf("expected active tab to become current target selection, got %#v", current)
	}
}

func TestTrackSharedSessionBrowserResolvedTargetFallsBackToCurrentTarget(t *testing.T) {
	registry := NewBrowserSessionRegistry()
	route := BrowserSessionRoute{Backend: "proxy", Profile: "work", Target: "node", BrowserApp: "Chromium"}

	tracked := TrackSharedSessionBrowserResolvedTarget(registry, "sess-1", route, BrowserTab{
		URL:   "https://example.com/current",
		Title: "Current",
	}, "browser_navigate")

	if tracked.ID == "" || tracked.TabIndex != 0 {
		t.Fatalf("expected current-target tracking fallback, got %#v", tracked)
	}
	current := CurrentSharedSessionBrowserTargetSelection(registry, "sess-1", route)
	if current == nil || current.ID != tracked.ID || current.Source != "browser_navigate" {
		t.Fatalf("expected tracked current target to become current selection, got %#v", current)
	}
}

func TestApplySharedSessionBrowserResolvedTargetWithContextSeedsSiblingProviderFromPrimaryCachedEventCycle(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	runRegistryA := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-resolved-target-helper-sibling-provider": {{RunID: "run-a", Status: "running"}},
		}},
	}
	runRegistryB := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-resolved-target-helper-sibling-provider": {{RunID: "run-b", Status: "running"}},
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
	sessionID := "sess-resolved-target-helper-sibling-provider"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}
	ctx := SharedSessionBrowserMutationContext{
		Registry:        sessionRegistry,
		RunRegistry:     runRegistryA,
		ReconnectWindow: time.Minute,
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

	_ = boundA.ObserveWatchLoop(context.Background(), req)
	_ = boundB.ObserveWatchLoop(context.Background(), req)
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

	result := ApplySharedSessionBrowserResolvedTargetWithContext(
		ctx,
		SharedSessionBrowserResolvedTargetEventRequest{
			SessionID: sessionID,
			Route:     route,
			URL:       "https://example.com/opened",
			Source:    "browser_open",
		},
	)
	if result.TargetID == "" {
		t.Fatalf("expected direct resolved-target helper to track current target, got %#v", result)
	}

	seededB := boundB.ObserveWatchLoop(context.Background(), req)
	if len(seededB.View.Session.Routes) != 1 || seededB.View.Session.Routes[0].CurrentTargetID != result.TargetID {
		t.Fatalf("expected sibling watch loop to reuse direct helper source, got %#v", seededB.View.Session.Routes)
	}
	if len(backend.statusReqs) != 2 || len(backend.profilesReqs) != 2 {
		t.Fatalf("expected direct helper sibling watch loop to avoid extra polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}

func TestApplySharedSessionBrowserResolvedTargetWithContextRestoresSiblingProviderDuringPendingReview(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	runRegistryA := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-resolved-target-helper-review-sibling-provider": {{RunID: "run-a", Status: "running"}},
		}},
	}
	runRegistryB := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-resolved-target-helper-review-sibling-provider": {{RunID: "run-b", Status: "running"}},
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
	sessionID := "sess-resolved-target-helper-review-sibling-provider"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}
	ctx := SharedSessionBrowserMutationContext{
		Registry:        sessionRegistry,
		RunRegistry:     runRegistryA,
		ReconnectWindow: time.Minute,
	}

	home := TrackSharedSessionBrowserTabWithContext(
		ctx,
		sessionID,
		route,
		BrowserTab{Index: 1, URL: "https://example.com/home", Title: "Home"},
		true,
	)
	priorSelection := SnapshotSharedSessionBrowserCurrentTargetSelection(sessionRegistry, sessionID, route)
	if home.ID == "" || priorSelection == nil {
		t.Fatalf("expected initial home target and prior selection, got home=%#v prior=%#v", home, priorSelection)
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

	_ = boundA.ObserveWatchLoop(context.Background(), req)
	_ = boundB.ObserveWatchLoop(context.Background(), req)
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

	result := ApplySharedSessionBrowserResolvedTargetWithContext(
		ctx,
		SharedSessionBrowserResolvedTargetEventRequest{
			SessionID:             sessionID,
			Route:                 route,
			URL:                   "https://example.com/redirected",
			Title:                 "Redirected",
			Source:                "browser_navigate",
			PendingReview:         true,
			PendingReviewDecision: "session_target_redirect_review_required",
			PendingReviewReason:   "redirect review",
			PriorSelection:        priorSelection,
		},
	)
	if result.TargetID == "" || result.Review == nil {
		t.Fatalf("expected direct resolved-target helper to track target and record review, got %#v", result)
	}

	seededB := boundB.ObserveWatchLoop(context.Background(), req)
	if len(seededB.View.Session.Routes) != 1 {
		t.Fatalf("expected sibling watch loop to preserve refreshed route snapshot, got %#v", seededB.View.Session.Routes)
	}
	if seededB.View.Session.Routes[0].CurrentTargetID != home.ID {
		t.Fatalf("expected direct helper to restore prior current target, got %#v", seededB.View.Session.Routes)
	}
	if seededB.View.Session.Routes[0].PendingTargetReview == nil || seededB.View.Session.Routes[0].PendingTargetReview.ID != result.TargetID {
		t.Fatalf("expected direct helper to record pending review on sibling provider, got %#v", seededB.View.Session.Routes)
	}
	if len(backend.statusReqs) != 2 || len(backend.profilesReqs) != 2 {
		t.Fatalf("expected direct helper sibling watch loop to avoid extra polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}

func TestTrackSharedSessionBrowserCurrentTargetWithContextSeedsSiblingProviderFromPrimaryCachedEventCycle(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	runRegistryA := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-track-current-target-helper-sibling-provider": {{RunID: "run-a", Status: "running"}},
		}},
	}
	runRegistryB := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-track-current-target-helper-sibling-provider": {{RunID: "run-b", Status: "running"}},
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
	sessionID := "sess-track-current-target-helper-sibling-provider"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}
	mutationCtx := SharedSessionBrowserMutationContext{
		Registry:        sessionRegistry,
		RunRegistry:     runRegistryA,
		ReconnectWindow: time.Minute,
	}

	initial := TrackSharedSessionBrowserCurrentTargetWithContext(
		mutationCtx,
		sessionID,
		route,
		"https://example.com/one",
		"One",
		"browser_navigate",
	)
	if initial.ID == "" {
		t.Fatalf("expected owner-aware current-target helper to create tracked target, got %#v", initial)
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
	if len(initialA.View.Session.Routes) != 1 || initialA.View.Session.Routes[0].CurrentTargetID != initial.ID {
		t.Fatalf("expected first provider to expose helper-tracked current target, got %#v", initialA.View.Session.Routes)
	}
	if len(initialB.View.Session.Routes) != 1 || initialB.View.Session.Routes[0].CurrentTargetID != initial.ID {
		t.Fatalf("expected second provider to expose helper-tracked current target, got %#v", initialB.View.Session.Routes)
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

	updated := TrackSharedSessionBrowserCurrentTargetWithContext(
		mutationCtx,
		sessionID,
		route,
		"https://example.com/two",
		"Two",
		"browser_navigate",
	)
	if updated.ID == "" {
		t.Fatalf("expected owner-aware current-target helper to update tracked target, got %#v", updated)
	}

	seededB := boundB.ObserveWatchLoop(context.Background(), req)
	if len(seededB.View.Session.Routes) != 1 || seededB.View.Session.Routes[0].CurrentTargetID != updated.ID {
		t.Fatalf("expected sibling watch loop to reuse helper-seeded current target, got %#v", seededB.View.Session.Routes)
	}
	if len(seededB.View.Session.Routes[0].Targets) != 1 || seededB.View.Session.Routes[0].Targets[0].Title != "Two" {
		t.Fatalf("expected sibling watch loop to expose updated helper-tracked target, got %#v", seededB.View.Session.Routes)
	}
	if len(backend.statusReqs) != 2 || len(backend.profilesReqs) != 2 {
		t.Fatalf("expected helper-seeded sibling watch loop to avoid extra polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}

func TestApplySharedSessionBrowserTargetWithContextSeedsSiblingProviderFromPrimaryCachedEventCycle(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	runRegistryA := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-apply-target-helper-sibling-provider": {{RunID: "run-a", Status: "running"}},
		}},
	}
	runRegistryB := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-apply-target-helper-sibling-provider": {{RunID: "run-b", Status: "running"}},
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
	sessionID := "sess-apply-target-helper-sibling-provider"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}
	mutationCtx := SharedSessionBrowserMutationContext{
		Registry:        sessionRegistry,
		RunRegistry:     runRegistryA,
		ReconnectWindow: time.Minute,
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

	_ = boundA.ObserveWatchLoop(context.Background(), req)
	_ = boundB.ObserveWatchLoop(context.Background(), req)
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

	result := ApplySharedSessionBrowserTargetWithContext(
		mutationCtx,
		SharedSessionBrowserTargetEventRequest{
			SessionID:  sessionID,
			Route:      route,
			TabIndex:   3,
			URL:        "https://example.com/third",
			Title:      "Third",
			Source:     "browser_act_response_body",
			SetCurrent: true,
		},
	)
	if result.TargetID == "" || result.Target.TabIndex != 3 {
		t.Fatalf("expected generic target helper to track tab handle, got %#v", result)
	}

	seededB := boundB.ObserveWatchLoop(context.Background(), req)
	if len(seededB.View.Session.Routes) != 1 || seededB.View.Session.Routes[0].CurrentTargetID != result.TargetID {
		t.Fatalf("expected sibling watch loop to reuse generic target helper source, got %#v", seededB.View.Session.Routes)
	}
	if len(seededB.View.Session.Routes[0].Targets) != 1 || seededB.View.Session.Routes[0].Targets[0].TabIndex != 3 {
		t.Fatalf("expected sibling watch loop to expose helper-tracked tab, got %#v", seededB.View.Session.Routes)
	}
	if len(backend.statusReqs) != 2 || len(backend.profilesReqs) != 2 {
		t.Fatalf("expected helper-seeded sibling watch loop to avoid extra polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}

func TestRestoreSharedSessionBrowserCurrentTargetSelectionRestoresPriorTarget(t *testing.T) {
	registry := NewBrowserSessionRegistry()
	sessionID := "sess-1"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "work", Target: "node", BrowserApp: "Chromium"}

	tracked := SyncSharedSessionBrowserTabsForRoute(registry, sessionID, route, 1, []BrowserTab{
		{Index: 1, URL: "https://example.com/1", Title: "One"},
		{Index: 2, URL: "https://example.com/2", Title: "Two"},
	})
	second, ok := registry.SelectTargetForRoute(sessionID, route, tracked[1].TargetID, "select_target")
	if !ok {
		t.Fatalf("expected second target to be selectable")
	}
	snapshot := SnapshotSharedSessionBrowserCurrentTargetSelection(registry, sessionID, route)
	if snapshot == nil || snapshot.ID != second.ID {
		t.Fatalf("expected snapshot for second target, got %#v", snapshot)
	}
	if _, ok := registry.SelectTargetForRoute(sessionID, route, tracked[0].TargetID, "select_target"); !ok {
		t.Fatalf("expected first target to be selectable")
	}

	restored := RestoreSharedSessionBrowserCurrentTargetSelection(registry, sessionID, route, snapshot, "popup_review_restore")
	if restored == nil || restored.ID != second.ID || restored.Source != "select_target" {
		t.Fatalf("expected prior current target to be restored, got %#v", restored)
	}
}

func TestRestoreSharedSessionBrowserCurrentTargetSelectionClearsMissingSnapshot(t *testing.T) {
	registry := NewBrowserSessionRegistry()
	sessionID := "sess-1"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "work", Target: "node", BrowserApp: "Chromium"}

	TrackSharedSessionBrowserResolvedTarget(registry, sessionID, route, BrowserTab{
		URL:   "https://example.com/current",
		Title: "Current",
	}, "browser_navigate")

	if cleared := RestoreSharedSessionBrowserCurrentTargetSelection(registry, sessionID, route, nil, "popup_review_restore"); cleared != nil {
		t.Fatalf("expected nil restore result when snapshot is missing, got %#v", cleared)
	}
	if current := CurrentSharedSessionBrowserTargetSelection(registry, sessionID, route); current != nil {
		t.Fatalf("expected current target selection to be cleared, got %#v", current)
	}
}

func TestInvalidateSharedSessionBrowserCurrentTargetForProfileStateClearsManagedSelection(t *testing.T) {
	registry := NewBrowserSessionRegistry()
	sessionID := "sess-1"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "work", Target: "node", BrowserApp: "Chromium"}

	tracked := SyncSharedSessionBrowserTabsForRoute(registry, sessionID, route, 1, []BrowserTab{
		{Index: 1, URL: "https://example.com/1", Title: "One"},
	})
	if len(tracked) != 1 || tracked[0].TargetID == "" {
		t.Fatalf("expected tracked tab before invalidation, got %#v", tracked)
	}

	cleared := InvalidateSharedSessionBrowserCurrentTargetForProfileState(registry, sessionID, SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "work",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "disconnected",
		Running:       true,
		Connected:     false,
	})
	if !cleared {
		t.Fatalf("expected managed current target selection to be cleared")
	}
	if current := CurrentSharedSessionBrowserTargetSelection(registry, sessionID, route); current != nil {
		t.Fatalf("expected current target selection to be removed, got %#v", current)
	}
	if _, ok := registry.ResolveTabForRoute(sessionID, route, 1); !ok {
		t.Fatalf("expected tracked tab to remain after target invalidation")
	}
}

func TestInvalidateSharedSessionBrowserCurrentTargetForProfileStateFallsBackWithoutBrowserApp(t *testing.T) {
	registry := NewBrowserSessionRegistry()
	sessionID := "sess-1"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "work", Target: "node"}

	tracked := SyncSharedSessionBrowserTabsForRoute(registry, sessionID, route, 1, []BrowserTab{
		{Index: 1, URL: "https://example.com/1", Title: "One"},
	})
	if len(tracked) != 1 || tracked[0].TargetID == "" {
		t.Fatalf("expected tracked tab before invalidation, got %#v", tracked)
	}

	cleared := InvalidateSharedSessionBrowserCurrentTargetForProfileState(registry, sessionID, SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "work",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "disconnected",
		Running:       true,
		Connected:     false,
	})
	if !cleared {
		t.Fatalf("expected managed current target selection to be cleared via browser-app fallback")
	}
	if current := CurrentSharedSessionBrowserTargetSelection(registry, sessionID, route); current != nil {
		t.Fatalf("expected current target selection to be removed, got %#v", current)
	}
	if _, ok := registry.ResolveTabForRoute(sessionID, route, 1); !ok {
		t.Fatalf("expected tracked tab to remain after target invalidation")
	}
}

func TestInvalidateSharedSessionBrowserCurrentTargetForProfileStatePreservesHostSelection(t *testing.T) {
	registry := NewBrowserSessionRegistry()
	sessionID := "sess-1"
	route := BrowserSessionRoute{Backend: "system", Profile: "default", Target: "host", BrowserApp: "Safari"}

	TrackSharedSessionBrowserResolvedTarget(registry, sessionID, route, BrowserTab{
		Index: 1,
		URL:   "https://example.com/host",
		Title: "Host",
	}, "browser_navigate")

	cleared := InvalidateSharedSessionBrowserCurrentTargetForProfileState(registry, sessionID, SharedSessionBrowserProfileState{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Status:        "disconnected",
		Running:       true,
		Connected:     false,
	})
	if cleared {
		t.Fatalf("expected host current target selection to be preserved")
	}
	if current := CurrentSharedSessionBrowserTargetSelection(registry, sessionID, route); current == nil {
		t.Fatalf("expected host current target selection to remain")
	}
}

func TestResolveSharedSessionBrowserCurrentTargetFallsBackToDefaultRoute(t *testing.T) {
	registry := NewBrowserSessionRegistry()
	sessionID := "sess-1"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "work", Target: "node", BrowserApp: "Chromium"}

	defaultTracked := TrackSharedSessionBrowserResolvedTarget(registry, sessionID, BrowserSessionRoute{}, BrowserTab{
		Index: 3,
		URL:   "https://example.com/default",
		Title: "Default",
	}, "browser_open")

	resolved, ok := ResolveSharedSessionBrowserCurrentTarget(registry, sessionID, route, true)
	if !ok || resolved.ID != defaultTracked.ID {
		t.Fatalf("expected default-route current target fallback, got %#v ok=%v", resolved, ok)
	}
}

func TestResolveSharedSessionBrowserCurrentTargetDoesNotFallbackWhenDisabled(t *testing.T) {
	registry := NewBrowserSessionRegistry()
	sessionID := "sess-1"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "work", Target: "node", BrowserApp: "Chromium"}

	TrackSharedSessionBrowserResolvedTarget(registry, sessionID, BrowserSessionRoute{}, BrowserTab{
		Index: 3,
		URL:   "https://example.com/default",
		Title: "Default",
	}, "browser_open")

	if resolved, ok := ResolveSharedSessionBrowserCurrentTarget(registry, sessionID, route, false); ok || resolved.ID != "" {
		t.Fatalf("expected scoped current-target lookup without default fallback, got %#v ok=%v", resolved, ok)
	}
}

func TestResolveSharedSessionBrowserTargetSupportsHandleTabAndCurrentFallback(t *testing.T) {
	registry := NewBrowserSessionRegistry()
	sessionID := "sess-1"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "work", Target: "node", BrowserApp: "Chromium"}

	tracked := SyncSharedSessionBrowserTabsForRoute(registry, sessionID, route, 2, []BrowserTab{
		{Index: 1, URL: "https://example.com/one", Title: "One"},
		{Index: 2, URL: "https://example.com/two", Title: "Two"},
	})
	if len(tracked) != 2 {
		t.Fatalf("expected tracked tabs, got %#v", tracked)
	}

	byHandle, ok := ResolveSharedSessionBrowserTarget(registry, sessionID, route, tracked[0].TargetID, 0, false)
	if !ok || byHandle.ID != tracked[0].TargetID {
		t.Fatalf("expected handle resolution for first tab, got %#v ok=%v", byHandle, ok)
	}

	byTab, ok := ResolveSharedSessionBrowserTarget(registry, sessionID, route, "", 1, false)
	if !ok || byTab.ID != tracked[0].TargetID {
		t.Fatalf("expected tab resolution for first tab, got %#v ok=%v", byTab, ok)
	}

	byCurrent, ok := ResolveSharedSessionBrowserTarget(registry, sessionID, route, "", 0, false)
	if !ok || byCurrent.ID != tracked[1].TargetID {
		t.Fatalf("expected current-target resolution for active tab, got %#v ok=%v", byCurrent, ok)
	}
}

func TestSyncSharedSessionBrowserTabsForRouteInvalidatesSharedWatchManagerCaches(t *testing.T) {
	registry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
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
	manager := SharedSessionBrowserObserverManagerFor(registry, nil, stateRegistry, time.Minute)
	sessionID := "sess-track-tabs-invalidates-watch"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}
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

	SyncSharedSessionBrowserTabsForRoute(registry, sessionID, route, 1, []BrowserTab{
		{Index: 1, URL: "https://example.com/one", Title: "One"},
	})
	first := manager.ObserveWatchLoop(context.Background(), backend, req)
	if len(first.View.Session.Routes) != 1 || len(first.View.Session.Routes[0].Targets) != 1 || first.View.Session.Routes[0].Targets[0].Title != "One" {
		t.Fatalf("expected initial watch loop to expose first tracked tab, got %#v", first.View.Session.Routes)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected initial watch loop to poll backend once, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}

	SyncSharedSessionBrowserTabsForRoute(registry, sessionID, route, 1, []BrowserTab{
		{Index: 1, URL: "https://example.com/one", Title: "Two"},
	})
	second := manager.ObserveWatchLoop(context.Background(), backend, req)
	if len(second.View.Session.Routes) != 1 || len(second.View.Session.Routes[0].Targets) != 1 || second.View.Session.Routes[0].Targets[0].Title != "Two" {
		t.Fatalf("expected tabs sync to invalidate cached watch loop, got %#v", second.View.Session.Routes)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected tabs sync invalidation to reuse cached raw status/profiles, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}

func TestForgetSharedSessionBrowserTabForRouteInvalidatesSharedWatchManagerCaches(t *testing.T) {
	registry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
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
	manager := SharedSessionBrowserObserverManagerFor(registry, nil, stateRegistry, time.Minute)
	sessionID := "sess-forget-tab-invalidates-watch"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}
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

	SyncSharedSessionBrowserTabsForRoute(registry, sessionID, route, 2, []BrowserTab{
		{Index: 1, URL: "https://example.com/one", Title: "One"},
		{Index: 2, URL: "https://example.com/two", Title: "Two"},
	})
	first := manager.ObserveWatchLoop(context.Background(), backend, req)
	if len(first.View.Session.Routes) != 1 || len(first.View.Session.Routes[0].Targets) != 2 {
		t.Fatalf("expected initial watch loop to expose two tracked tabs, got %#v", first.View.Session.Routes)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected initial watch loop to poll backend once, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}

	ForgetSharedSessionBrowserTabForRoute(registry, sessionID, route, 1)
	second := manager.ObserveWatchLoop(context.Background(), backend, req)
	if len(second.View.Session.Routes) != 1 || len(second.View.Session.Routes[0].Targets) != 1 || second.View.Session.Routes[0].Targets[0].TabIndex != 2 {
		t.Fatalf("expected forgetting a tracked tab to invalidate cached watch loop, got %#v", second.View.Session.Routes)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected forget-tab invalidation to reuse cached raw status/profiles, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}
