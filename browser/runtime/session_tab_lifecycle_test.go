package browserruntime

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestApplySharedSessionBrowserTabRememberReviewUsesSharedProviderProjectionRefreshWhenRunRegistryProvided(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	runRegistry := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-remember-review-runtime-helper-shared-provider": {{RunID: "run-1", Status: "running"}},
		}},
	}
	manager := SharedSessionBrowserObserverManagerFor(sessionRegistry, runRegistry, nil, time.Minute)
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
	bound := manager.Bind(backend)
	sessionID := "sess-remember-review-runtime-helper-shared-provider"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}

	firstTabs := SyncSharedSessionBrowserTabsForRouteEvent(
		sessionRegistry,
		runRegistry,
		nil,
		sessionID,
		route,
		1,
		[]BrowserTab{{Index: 1, URL: "https://example.com/home", Title: "Home"}},
		time.Minute,
	)
	priorSelection := SnapshotSharedSessionBrowserCurrentTargetSelection(sessionRegistry, sessionID, route)
	trackedTabs := SyncSharedSessionBrowserTabsForRouteEvent(
		sessionRegistry,
		runRegistry,
		nil,
		sessionID,
		route,
		2,
		[]BrowserTab{
			{Index: 1, URL: "https://example.com/home", Title: "Home"},
			{Index: 2, URL: "https://popup.example/offer", Title: "Offer"},
		},
		time.Minute,
	)
	if len(firstTabs) != 1 || len(trackedTabs) != 2 || priorSelection == nil || strings.TrimSpace(trackedTabs[1].TargetID) == "" {
		t.Fatalf("expected tab sync to seed tracked tabs and prior selection, got first=%#v tracked=%#v prior=%#v", firstTabs, trackedTabs, priorSelection)
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

	first := bound.ObserveWatchLoop(context.Background(), req)
	if len(first.View.Session.Routes) != 1 || first.View.Session.Routes[0].CurrentTargetID != strings.TrimSpace(trackedTabs[1].TargetID) {
		t.Fatalf("expected initial watch loop to expose popup tab as current selection, got %#v", first.View.Session.Routes)
	}
	if runRegistry.callCount() != 1 {
		t.Fatalf("expected initial watch loop to snapshot runs once, got %d", runRegistry.callCount())
	}

	bound.state.mu.Lock()
	clear(bound.state.rawStatus)
	clear(bound.state.rawProfiles)
	eventCycleCount := len(bound.state.eventCycles)
	watchLoopCount := len(bound.state.watchLoops)
	bound.state.mu.Unlock()
	if eventCycleCount == 0 || watchLoopCount == 0 {
		t.Fatalf("expected cached event-cycle/watch-loop source before draining raw caches, got eventCycles=%d watchLoops=%d", eventCycleCount, watchLoopCount)
	}

	rememberReview := ApplySharedSessionBrowserTabRememberReview(SharedSessionBrowserTabRememberReviewRequest{
		Registry:            sessionRegistry,
		RunRegistry:         runRegistry,
		ReconnectWindow:     time.Minute,
		SessionID:           sessionID,
		Route:               route,
		Action:              "list",
		Force:               false,
		RememberTarget:      true,
		CandidateTargetID:   strings.TrimSpace(trackedTabs[1].TargetID),
		ActiveIndex:         2,
		PriorSelection:      priorSelection,
		Tabs:                trackedTabs,
		PriorActiveTargetID: "",
	})
	if rememberReview.Decision != "session_target_popup_review_required" || rememberReview.Ready {
		t.Fatalf("expected remember review to require popup confirmation, got %#v", rememberReview)
	}
	if runRegistry.callCount() != 2 {
		t.Fatalf("expected runtime helper remember review to refresh shared-provider projection once, got %d", runRegistry.callCount())
	}

	second := bound.ObserveWatchLoop(context.Background(), req)
	if len(second.View.Session.Routes) != 1 {
		t.Fatalf("expected refreshed watch loop to expose one route, got %#v", second.View.Session.Routes)
	}
	if second.View.Session.Routes[0].CurrentTargetID != strings.TrimSpace(priorSelection.ID) {
		t.Fatalf("expected runtime helper remember review to restore prior current target, got %#v", second.View.Session.Routes)
	}
	if second.View.Session.Routes[0].PendingTargetReview == nil || second.View.Session.Routes[0].PendingTargetReview.ID != strings.TrimSpace(trackedTabs[1].TargetID) {
		t.Fatalf("expected runtime helper remember review to record pending popup review, got %#v", second.View.Session.Routes)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected runtime helper remember review to reuse cached projection without extra polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
	if runRegistry.callCount() != 2 {
		t.Fatalf("expected next ObserveWatchLoop to reuse refreshed shared-provider projection cache, got %d", runRegistry.callCount())
	}
}

func TestApplySharedSessionBrowserTabsResultWithContextSeedsSiblingProviderFromPrimaryCachedEventCycle(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	runRegistryA := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-tabs-result-helper-close-sibling-provider": {{RunID: "run-a", Status: "running"}},
		}},
	}
	runRegistryB := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-tabs-result-helper-close-sibling-provider": {{RunID: "run-b", Status: "running"}},
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
	sessionID := "sess-tabs-result-helper-close-sibling-provider"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}
	ctx := SharedSessionBrowserMutationContext{
		Registry:        sessionRegistry,
		RunRegistry:     runRegistryA,
		ReconnectWindow: time.Minute,
	}

	initial := ApplySharedSessionBrowserTabsResultWithContext(
		ctx,
		SharedSessionBrowserTabsResultEventRequest{
			SessionID:   sessionID,
			Route:       route,
			Action:      "list",
			ActiveIndex: 2,
			Tabs: []BrowserTab{
				{Index: 1, URL: "https://example.com/home", Title: "Home"},
				{Index: 2, URL: "https://example.com/popup", Title: "Popup"},
			},
		},
	)
	homeID := strings.TrimSpace(initial.Tabs[0].TargetID)
	popupID := strings.TrimSpace(initial.Tabs[1].TargetID)
	if homeID == "" || popupID == "" {
		t.Fatalf("expected tabs-result helper to seed initial target ids, got %#v", initial)
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

	closed := ApplySharedSessionBrowserTabsResultWithContext(
		ctx,
		SharedSessionBrowserTabsResultEventRequest{
			SessionID:         sessionID,
			Route:             route,
			Action:            "close",
			RequestedTabIndex: 2,
			ActiveIndex:       1,
			Tabs: []BrowserTab{
				{Index: 1, URL: "https://example.com/home", Title: "Home"},
			},
		},
	)
	if closed.TargetID != popupID {
		t.Fatalf("expected tabs-result helper to preserve closed tab target id %q, got %#v", popupID, closed)
	}

	seededB := boundB.ObserveWatchLoop(context.Background(), req)
	if len(seededB.View.Session.Routes) != 1 || seededB.View.Session.Routes[0].CurrentTargetID != homeID {
		t.Fatalf("expected sibling watch loop to reuse helper-seeded close result, got %#v", seededB.View.Session.Routes)
	}
	if len(backend.statusReqs) != 2 || len(backend.profilesReqs) != 2 {
		t.Fatalf("expected helper-seeded sibling watch loop to avoid extra polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}

func TestApplySharedSessionBrowserTabsResultWithContextRestoresSiblingProviderDuringRememberReview(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	runRegistryA := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-tabs-result-helper-review-sibling-provider": {{RunID: "run-a", Status: "running"}},
		}},
	}
	runRegistryB := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-tabs-result-helper-review-sibling-provider": {{RunID: "run-b", Status: "running"}},
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
	sessionID := "sess-tabs-result-helper-review-sibling-provider"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}
	ctx := SharedSessionBrowserMutationContext{
		Registry:        sessionRegistry,
		RunRegistry:     runRegistryA,
		ReconnectWindow: time.Minute,
	}

	initial := ApplySharedSessionBrowserTabsResultWithContext(
		ctx,
		SharedSessionBrowserTabsResultEventRequest{
			SessionID:   sessionID,
			Route:       route,
			Action:      "list",
			ActiveIndex: 1,
			Tabs: []BrowserTab{
				{Index: 1, URL: "https://example.com/home", Title: "Home"},
			},
		},
	)
	homeID := strings.TrimSpace(initial.Tabs[0].TargetID)
	if homeID == "" {
		t.Fatalf("expected tabs-result helper to seed home target id, got %#v", initial)
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

	reviewed := ApplySharedSessionBrowserTabsResultWithContext(
		ctx,
		SharedSessionBrowserTabsResultEventRequest{
			SessionID:   sessionID,
			Route:       route,
			Action:      "list",
			ActiveIndex: 2,
			Tabs: []BrowserTab{
				{Index: 1, URL: "https://example.com/home", Title: "Home"},
				{Index: 2, URL: "https://popup.example/offer", Title: "Offer"},
			},
			RememberTarget: true,
		},
	)
	if reviewed.RememberReview.Decision != "session_target_popup_review_required" || reviewed.RememberReview.Ready {
		t.Fatalf("expected tabs-result helper to require popup remember review, got %#v", reviewed)
	}
	if strings.TrimSpace(reviewed.TargetID) == "" {
		t.Fatalf("expected tabs-result helper to resolve popup target id, got %#v", reviewed)
	}

	seededB := boundB.ObserveWatchLoop(context.Background(), req)
	if len(seededB.View.Session.Routes) != 1 {
		t.Fatalf("expected sibling watch loop to preserve refreshed route snapshot, got %#v", seededB.View.Session.Routes)
	}
	if seededB.View.Session.Routes[0].CurrentTargetID != homeID {
		t.Fatalf("expected tabs-result helper to restore prior current target, got %#v", seededB.View.Session.Routes)
	}
	if seededB.View.Session.Routes[0].PendingTargetReview == nil || seededB.View.Session.Routes[0].PendingTargetReview.ID != strings.TrimSpace(reviewed.TargetID) {
		t.Fatalf("expected tabs-result helper to record popup review on sibling provider, got %#v", seededB.View.Session.Routes)
	}
	if len(backend.statusReqs) != 2 || len(backend.profilesReqs) != 2 {
		t.Fatalf("expected helper-seeded sibling watch loop to avoid extra polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}
