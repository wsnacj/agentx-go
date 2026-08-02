package browserruntime

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestSharedSessionBrowserObserverManagerSelectTargetInvalidatesBoundWatchLoopCache(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	manager := NewSharedSessionBrowserObserverManager(sessionRegistry, nil, nil, time.Minute)
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
	sessionID := "sess-manager-select-target-cache"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node"}

	sessionRegistry.TrackTab(sessionID, BrowserSessionTarget{
		ID:         "tab-1",
		TabIndex:   1,
		URL:        "https://example.com/one",
		Title:      "One",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
	}, true)
	sessionRegistry.TrackTab(sessionID, BrowserSessionTarget{
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

	first := bound.ObserveWatchLoop(context.Background(), req)
	if len(first.View.Session.Routes) != 1 || first.View.Session.Routes[0].CurrentTargetID == "" {
		t.Fatalf("expected initial watch loop to expose tab-1 as current target, got %#v", first.View.Session.Routes)
	}
	if len(first.View.Session.Routes[0].Targets) != 2 {
		t.Fatalf("expected initial watch loop to expose both tracked targets, got %#v", first.View.Session.Routes)
	}
	nextTargetID := first.View.Session.Routes[0].Targets[1].ID
	if nextTargetID == "" {
		t.Fatalf("expected second tracked target to expose an ID, got %#v", first.View.Session.Routes[0].Targets)
	}

	result, err := manager.SelectTarget(SharedSessionBrowserSelectTargetRequest{
		SessionID: sessionID,
		Route:     route,
		TargetID:  nextTargetID,
		Source:    "select_target",
		Actor:     "test select target",
	})
	if err != nil {
		t.Fatalf("expected target selection to succeed, got %v", err)
	}
	if !result.Ready || result.Selection == nil || result.Selection.ID != nextTargetID {
		t.Fatalf("expected manager target selection to select tab-2, got %#v", result)
	}

	second := bound.ObserveWatchLoop(context.Background(), req)
	if len(second.View.Session.Routes) != 1 || second.View.Session.Routes[0].CurrentTargetID != nextTargetID {
		t.Fatalf("expected target selection to invalidate bound watch-loop cache, got %#v", second.View.Session.Routes)
	}
}

func TestSharedSessionBrowserObserverManagerSelectTargetRefreshesStandaloneProjectionFromCachedEventCycleWhenRawDrained(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	runRegistry := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-manager-select-target-refreshes-from-cached-cycle": {{RunID: "run-1", Status: "running"}},
		}},
	}
	manager := NewSharedSessionBrowserObserverManager(sessionRegistry, runRegistry, nil, time.Minute)
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
	sessionID := "sess-manager-select-target-refreshes-from-cached-cycle"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node"}

	sessionRegistry.TrackTab(sessionID, BrowserSessionTarget{
		ID:         "tab-1",
		TabIndex:   1,
		URL:        "https://example.com/one",
		Title:      "One",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
	}, true)
	sessionRegistry.TrackTab(sessionID, BrowserSessionTarget{
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

	first := bound.ObserveWatchLoop(context.Background(), req)
	if len(first.View.Session.Routes) != 1 || first.View.Session.Routes[0].CurrentTargetID == "" {
		t.Fatalf("expected initial watch loop to expose current target, got %#v", first.View.Session.Routes)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected initial watch loop to poll backend once, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
	if runRegistry.callCount() != 1 {
		t.Fatalf("expected initial watch loop to snapshot runs once, got %d", runRegistry.callCount())
	}
	nextTargetID := first.View.Session.Routes[0].Targets[1].ID
	if nextTargetID == "" {
		t.Fatalf("expected second tracked target to expose an ID, got %#v", first.View.Session.Routes[0].Targets)
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

	result, err := manager.SelectTarget(SharedSessionBrowserSelectTargetRequest{
		SessionID: sessionID,
		Route:     route,
		TargetID:  nextTargetID,
		Source:    "select_target",
		Actor:     "test select target",
	})
	if err != nil {
		t.Fatalf("expected target selection to succeed, got %v", err)
	}
	if !result.Ready || result.Selection == nil || result.Selection.ID != nextTargetID {
		t.Fatalf("expected target selection to select tab-2, got %#v", result)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected standalone projection refresh to reuse cached event-cycle source, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
	if runRegistry.callCount() != 2 {
		t.Fatalf("expected standalone target selection to refresh watch-loop projection once, got %d", runRegistry.callCount())
	}

	second := bound.ObserveWatchLoop(context.Background(), req)
	if len(second.View.Session.Routes) != 1 || second.View.Session.Routes[0].CurrentTargetID != nextTargetID {
		t.Fatalf("expected target selection to refresh standalone watch-loop cache, got %#v", second.View.Session.Routes)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected next ObserveWatchLoop to reuse refreshed standalone projection cache, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
	if runRegistry.callCount() != 2 {
		t.Fatalf("expected next ObserveWatchLoop to avoid second standalone projection rebuild, got %d", runRegistry.callCount())
	}
}

func TestSharedSessionBrowserObserverManagerSyncTabsForRouteEventRefreshesStandaloneProjectionFromCachedEventCycleWhenRawDrained(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	runRegistry := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-manager-sync-tabs-refreshes-from-cached-cycle": {{RunID: "run-1", Status: "running"}},
		}},
	}
	manager := NewSharedSessionBrowserObserverManager(sessionRegistry, runRegistry, nil, time.Minute)
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
	sessionID := "sess-manager-sync-tabs-refreshes-from-cached-cycle"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}

	manager.SyncTabsForRouteEvent(sessionID, route, 1, []BrowserTab{
		{Index: 1, URL: "https://example.com/one", Title: "One"},
	})

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
	if len(first.View.Session.Routes) != 1 || len(first.View.Session.Routes[0].Targets) != 1 || first.View.Session.Routes[0].Targets[0].Title != "One" {
		t.Fatalf("expected initial watch loop to expose first tracked tab, got %#v", first.View.Session.Routes)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected initial watch loop to poll backend once, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
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

	tracked := manager.SyncTabsForRouteEvent(sessionID, route, 1, []BrowserTab{
		{Index: 1, URL: "https://example.com/one", Title: "Two"},
	})
	if len(tracked) != 1 || tracked[0].Title != "Two" {
		t.Fatalf("expected tabs event to update tracked tab payload, got %#v", tracked)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected tabs event refresh to reuse cached event-cycle source, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
	if runRegistry.callCount() != 2 {
		t.Fatalf("expected tabs event to refresh projection once, got %d", runRegistry.callCount())
	}

	second := bound.ObserveWatchLoop(context.Background(), req)
	if len(second.View.Session.Routes) != 1 || len(second.View.Session.Routes[0].Targets) != 1 || second.View.Session.Routes[0].Targets[0].Title != "Two" {
		t.Fatalf("expected tabs event to refresh standalone watch-loop cache, got %#v", second.View.Session.Routes)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected next ObserveWatchLoop to reuse refreshed standalone projection cache, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
	if runRegistry.callCount() != 2 {
		t.Fatalf("expected next ObserveWatchLoop to avoid second standalone projection rebuild, got %d", runRegistry.callCount())
	}
}

func TestSharedSessionBrowserObserverManagerSyncTabsForRouteEventObserveWatchLoopReusesRouteMutationSourceWhenCachesDrained(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	runRegistry := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-manager-sync-tabs-refreshes-from-route-mutation-source": {{RunID: "run-1", Status: "running"}},
		}},
	}
	manager := NewSharedSessionBrowserObserverManager(sessionRegistry, runRegistry, nil, time.Minute)
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
	sessionID := "sess-manager-sync-tabs-refreshes-from-route-mutation-source"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}

	manager.SyncTabsForRouteEvent(sessionID, route, 1, []BrowserTab{
		{Index: 1, URL: "https://example.com/one", Title: "One"},
	})

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
	if len(first.View.Session.Routes) != 1 || len(first.View.Session.Routes[0].Targets) != 1 || first.View.Session.Routes[0].Targets[0].Title != "One" {
		t.Fatalf("expected initial watch loop to expose first tracked tab, got %#v", first.View.Session.Routes)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected initial watch loop to poll backend once, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
	if runRegistry.callCount() != 1 {
		t.Fatalf("expected initial watch loop to snapshot runs once, got %d", runRegistry.callCount())
	}

	secondTracked := manager.SyncTabsForRouteEvent(sessionID, route, 1, []BrowserTab{
		{Index: 1, URL: "https://example.com/two", Title: "Two"},
	})
	if len(secondTracked) != 1 || secondTracked[0].Title != "Two" {
		t.Fatalf("expected second tabs event to update tracked tab payload, got %#v", secondTracked)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected second tabs event refresh to reuse cached event-cycle source, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
	if runRegistry.callCount() != 2 {
		t.Fatalf("expected second tabs event to refresh projection once, got %d", runRegistry.callCount())
	}

	bound.state.mu.Lock()
	clear(bound.state.rawStatus)
	clear(bound.state.rawProfiles)
	clear(bound.state.eventCycles)
	clear(bound.state.bindings)
	clear(bound.state.views)
	clear(bound.state.watchLoops)
	clear(bound.state.eventCyclesInFlight)
	clear(bound.state.bindingsInFlight)
	clear(bound.state.viewsInFlight)
	clear(bound.state.watchLoopsInFlight)
	routeMutationCount := len(bound.state.routeMutations)
	bound.state.mu.Unlock()
	if routeMutationCount == 0 {
		t.Fatalf("expected route-mutation source before draining raw/projection caches")
	}

	thirdTracked := manager.SyncTabsForRouteEvent(sessionID, route, 1, []BrowserTab{
		{Index: 1, URL: "https://example.com/three", Title: "Three"},
	})
	if len(thirdTracked) != 1 || thirdTracked[0].Title != "Three" {
		t.Fatalf("expected third tabs event to update tracked tab payload, got %#v", thirdTracked)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected third tabs event to avoid eager polling while retaining route-mutation source, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
	if runRegistry.callCount() != 2 {
		t.Fatalf("expected third tabs event to defer rebuild until observe reuses route-mutation source, got %d", runRegistry.callCount())
	}

	second := bound.ObserveWatchLoop(context.Background(), req)
	if len(second.View.Session.Routes) != 1 || len(second.View.Session.Routes[0].Targets) != 1 || second.View.Session.Routes[0].Targets[0].Title != "Three" {
		t.Fatalf("expected tabs event to refresh standalone watch-loop cache from route-mutation source, got %#v", second.View.Session.Routes)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected next ObserveWatchLoop to reuse route-mutation refreshed projection cache, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
	if runRegistry.callCount() != 3 {
		t.Fatalf("expected next ObserveWatchLoop to rebuild once from route-mutation source, got %d", runRegistry.callCount())
	}
}

func TestSharedSessionBrowserObserverManagerForgetTabForRouteEventRefreshesStandaloneProjectionFromCachedEventCycleWhenRawDrained(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	runRegistry := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-manager-forget-tab-refreshes-from-cached-cycle": {{RunID: "run-1", Status: "running"}},
		}},
	}
	manager := NewSharedSessionBrowserObserverManager(sessionRegistry, runRegistry, nil, time.Minute)
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
	sessionID := "sess-manager-forget-tab-refreshes-from-cached-cycle"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}

	manager.SyncTabsForRouteEvent(sessionID, route, 2, []BrowserTab{
		{Index: 1, URL: "https://example.com/one", Title: "One"},
		{Index: 2, URL: "https://example.com/two", Title: "Two"},
	})

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
	if len(first.View.Session.Routes) != 1 || len(first.View.Session.Routes[0].Targets) != 2 {
		t.Fatalf("expected initial watch loop to expose tracked tabs, got %#v", first.View.Session.Routes)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected initial watch loop to poll backend once, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
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

	manager.ForgetTabForRouteEvent(sessionID, route, 1)
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected forget-tab event refresh to reuse cached event-cycle source, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
	if runRegistry.callCount() != 2 {
		t.Fatalf("expected forget-tab event to refresh projection once, got %d", runRegistry.callCount())
	}

	second := bound.ObserveWatchLoop(context.Background(), req)
	if len(second.View.Session.Routes) != 1 || len(second.View.Session.Routes[0].Targets) != 1 || second.View.Session.Routes[0].Targets[0].TabIndex != 2 {
		t.Fatalf("expected forget-tab event to refresh standalone watch-loop cache, got %#v", second.View.Session.Routes)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected next ObserveWatchLoop to reuse refreshed standalone projection cache, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
	if runRegistry.callCount() != 2 {
		t.Fatalf("expected next ObserveWatchLoop to avoid second standalone projection rebuild, got %d", runRegistry.callCount())
	}
}

func TestSharedSessionBrowserObserverManagerTrackCurrentTargetEventRefreshesStandaloneProjectionFromCachedEventCycleWhenRawDrained(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	runRegistry := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-manager-track-current-target-refreshes-from-cached-cycle": {{RunID: "run-1", Status: "running"}},
		}},
	}
	manager := NewSharedSessionBrowserObserverManager(sessionRegistry, runRegistry, nil, time.Minute)
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
	sessionID := "sess-manager-track-current-target-refreshes-from-cached-cycle"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}

	initial := manager.TrackCurrentTargetEvent(
		sessionID,
		route,
		"https://example.com/one",
		"One",
		"browser_navigate",
	)
	if initial.ID == "" {
		t.Fatalf("expected current-target event to create a tracked target, got %#v", initial)
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
	if len(first.View.Session.Routes) != 1 || first.View.Session.Routes[0].CurrentTargetID != initial.ID {
		t.Fatalf("expected initial watch loop to expose current target, got %#v", first.View.Session.Routes)
	}
	if len(first.View.Session.Routes[0].Targets) != 1 || first.View.Session.Routes[0].Targets[0].Title != "One" {
		t.Fatalf("expected initial watch loop to expose tracked current target, got %#v", first.View.Session.Routes)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected initial watch loop to poll backend once, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
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

	updated := manager.TrackCurrentTargetEvent(
		sessionID,
		route,
		"https://example.com/two",
		"Two",
		"browser_navigate",
	)
	if updated.ID == "" {
		t.Fatalf("expected current-target event to return tracked target, got %#v", updated)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected current-target event refresh to reuse cached event-cycle source, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
	if runRegistry.callCount() != 2 {
		t.Fatalf("expected current-target event to refresh projection once, got %d", runRegistry.callCount())
	}

	second := bound.ObserveWatchLoop(context.Background(), req)
	if len(second.View.Session.Routes) != 1 || second.View.Session.Routes[0].CurrentTargetID != updated.ID {
		t.Fatalf("expected current-target event to refresh standalone watch-loop cache, got %#v", second.View.Session.Routes)
	}
	if len(second.View.Session.Routes[0].Targets) != 1 || second.View.Session.Routes[0].Targets[0].Title != "Two" || second.View.Session.Routes[0].Targets[0].URL != "https://example.com/two" {
		t.Fatalf("expected refreshed watch loop to expose updated current target, got %#v", second.View.Session.Routes)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected next ObserveWatchLoop to reuse refreshed standalone projection cache, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
	if runRegistry.callCount() != 2 {
		t.Fatalf("expected next ObserveWatchLoop to avoid second standalone projection rebuild, got %d", runRegistry.callCount())
	}
}

func TestSharedSessionBrowserObserverManagerTrackCurrentTargetEventObserveWatchLoopReusesRouteMutationSourceWhenCachesDrained(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	runRegistry := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-manager-track-current-target-refreshes-from-route-mutation-source": {{RunID: "run-1", Status: "running"}},
		}},
	}
	manager := NewSharedSessionBrowserObserverManager(sessionRegistry, runRegistry, nil, time.Minute)
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
	sessionID := "sess-manager-track-current-target-refreshes-from-route-mutation-source"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}

	initial := manager.TrackCurrentTargetEvent(
		sessionID,
		route,
		"https://example.com/one",
		"One",
		"browser_navigate",
	)
	if initial.ID == "" {
		t.Fatalf("expected current-target event to create a tracked target, got %#v", initial)
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
	if len(first.View.Session.Routes) != 1 || first.View.Session.Routes[0].CurrentTargetID != initial.ID {
		t.Fatalf("expected initial watch loop to expose current target, got %#v", first.View.Session.Routes)
	}
	if len(first.View.Session.Routes[0].Targets) != 1 || first.View.Session.Routes[0].Targets[0].Title != "One" {
		t.Fatalf("expected initial watch loop to expose tracked current target, got %#v", first.View.Session.Routes)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected initial watch loop to poll backend once, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
	if runRegistry.callCount() != 1 {
		t.Fatalf("expected initial watch loop to snapshot runs once, got %d", runRegistry.callCount())
	}

	secondTracked := manager.TrackCurrentTargetEvent(
		sessionID,
		route,
		"https://example.com/two",
		"Two",
		"browser_navigate",
	)
	if secondTracked.ID == "" {
		t.Fatalf("expected second current-target event to return tracked target, got %#v", secondTracked)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected second current-target event refresh to reuse cached event-cycle source, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
	if runRegistry.callCount() != 2 {
		t.Fatalf("expected second current-target event to refresh projection once, got %d", runRegistry.callCount())
	}

	bound.state.mu.Lock()
	clear(bound.state.rawStatus)
	clear(bound.state.rawProfiles)
	clear(bound.state.eventCycles)
	clear(bound.state.bindings)
	clear(bound.state.views)
	clear(bound.state.watchLoops)
	clear(bound.state.eventCyclesInFlight)
	clear(bound.state.bindingsInFlight)
	clear(bound.state.viewsInFlight)
	clear(bound.state.watchLoopsInFlight)
	routeMutationCount := len(bound.state.routeMutations)
	bound.state.mu.Unlock()
	if routeMutationCount == 0 {
		t.Fatalf("expected route-mutation source before draining raw/projection caches")
	}

	thirdTracked := manager.TrackCurrentTargetEvent(
		sessionID,
		route,
		"https://example.com/three",
		"Three",
		"browser_navigate",
	)
	if thirdTracked.ID == "" {
		t.Fatalf("expected third current-target event to return tracked target, got %#v", thirdTracked)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected third current-target event to avoid eager polling while retaining route-mutation source, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
	if runRegistry.callCount() != 2 {
		t.Fatalf("expected third current-target event to defer rebuild until observe reuses route-mutation source, got %d", runRegistry.callCount())
	}

	second := bound.ObserveWatchLoop(context.Background(), req)
	if len(second.View.Session.Routes) != 1 || second.View.Session.Routes[0].CurrentTargetID != thirdTracked.ID {
		t.Fatalf("expected current-target event to refresh standalone watch-loop cache from route-mutation source, got %#v", second.View.Session.Routes)
	}
	if len(second.View.Session.Routes[0].Targets) != 1 || second.View.Session.Routes[0].Targets[0].Title != "Three" || second.View.Session.Routes[0].Targets[0].URL != "https://example.com/three" {
		t.Fatalf("expected refreshed watch loop to expose updated current target from route-mutation source, got %#v", second.View.Session.Routes)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected next ObserveWatchLoop to reuse route-mutation refreshed projection cache, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
	if runRegistry.callCount() != 3 {
		t.Fatalf("expected next ObserveWatchLoop to rebuild once from route-mutation source, got %d", runRegistry.callCount())
	}
}

func TestSharedSessionBrowserObserverManagerRecordPendingTargetPopupReviewEventRefreshesStandaloneProjectionFromCachedEventCycleWhenRawDrained(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	runRegistry := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-manager-pending-review-refreshes-from-cached-cycle": {{RunID: "run-1", Status: "running"}},
		}},
	}
	manager := NewSharedSessionBrowserObserverManager(sessionRegistry, runRegistry, nil, time.Minute)
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
	sessionID := "sess-manager-pending-review-refreshes-from-cached-cycle"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}

	popup := manager.TrackTabEvent(sessionID, route, BrowserTab{
		Index: 2,
		URL:   "https://popup.example/offer",
		Title: "Offer",
	}, false)
	if popup.ID == "" {
		t.Fatalf("expected track-tab event to create popup target, got %#v", popup)
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
	if len(first.View.Session.Routes) != 1 || first.View.Session.Routes[0].PendingTargetReview != nil {
		t.Fatalf("expected initial watch loop without pending review, got %#v", first.View.Session.Routes)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected initial watch loop to poll backend once, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
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

	review := manager.RecordPendingTargetPopupReviewEvent(
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
	if review == nil || review.ID != popup.ID {
		t.Fatalf("expected popup review event to resolve tracked popup target, got %#v", review)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected popup review event refresh to reuse cached event-cycle source, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
	if runRegistry.callCount() != 2 {
		t.Fatalf("expected popup review event to refresh projection once, got %d", runRegistry.callCount())
	}

	second := bound.ObserveWatchLoop(context.Background(), req)
	if len(second.View.Session.Routes) != 1 || second.View.Session.Routes[0].PendingTargetReview == nil || second.View.Session.Routes[0].PendingTargetReview.ID != popup.ID {
		t.Fatalf("expected popup review event to refresh standalone watch-loop cache, got %#v", second.View.Session.Routes)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected next ObserveWatchLoop to reuse refreshed standalone projection cache, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
	if runRegistry.callCount() != 2 {
		t.Fatalf("expected next ObserveWatchLoop to avoid second standalone projection rebuild, got %d", runRegistry.callCount())
	}
}

func TestSharedSessionBrowserObserverManagerRestoreCurrentTargetSelectionForPendingReviewEventRefreshesStandaloneProjectionFromCachedEventCycleWhenRawDrained(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	runRegistry := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-manager-restore-current-target-refreshes-from-cached-cycle": {{RunID: "run-1", Status: "running"}},
		}},
	}
	manager := NewSharedSessionBrowserObserverManager(sessionRegistry, runRegistry, nil, time.Minute)
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
	sessionID := "sess-manager-restore-current-target-refreshes-from-cached-cycle"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}

	prior := manager.TrackTabEvent(sessionID, route, BrowserTab{
		Index: 1,
		URL:   "https://example.com/one",
		Title: "One",
	}, true)
	if prior.ID == "" {
		t.Fatalf("expected first track-tab event to create prior target, got %#v", prior)
	}
	priorSelection := SnapshotSharedSessionBrowserCurrentTargetSelection(sessionRegistry, sessionID, route)
	if priorSelection == nil || priorSelection.ID != prior.ID {
		t.Fatalf("expected prior current-target snapshot to match first target, got %#v", priorSelection)
	}
	pending := manager.TrackTabEvent(sessionID, route, BrowserTab{
		Index: 2,
		URL:   "https://popup.example/offer",
		Title: "Offer",
	}, true)
	if pending.ID == "" {
		t.Fatalf("expected second track-tab event to create pending target, got %#v", pending)
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
	if len(first.View.Session.Routes) != 1 || first.View.Session.Routes[0].CurrentTargetID != pending.ID {
		t.Fatalf("expected initial watch loop to expose pending target as current selection, got %#v", first.View.Session.Routes)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected initial watch loop to poll backend once, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
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

	restored := manager.RestoreCurrentTargetSelectionForPendingReviewEvent(
		sessionID,
		route,
		priorSelection,
		pending.ID,
		"popup_review_restore",
	)
	if restored == nil || restored.ID != prior.ID {
		t.Fatalf("expected restore event to switch current selection back to prior target, got %#v", restored)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected restore event refresh to reuse cached event-cycle source, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
	if runRegistry.callCount() != 2 {
		t.Fatalf("expected restore event to refresh projection once, got %d", runRegistry.callCount())
	}

	second := bound.ObserveWatchLoop(context.Background(), req)
	if len(second.View.Session.Routes) != 1 || second.View.Session.Routes[0].CurrentTargetID != prior.ID {
		t.Fatalf("expected restore event to refresh standalone watch-loop cache, got %#v", second.View.Session.Routes)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected next ObserveWatchLoop to reuse refreshed standalone projection cache, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
	if runRegistry.callCount() != 2 {
		t.Fatalf("expected next ObserveWatchLoop to avoid second standalone projection rebuild, got %d", runRegistry.callCount())
	}
}

func TestSharedSessionBrowserObserverManagerApplyTabRememberReviewEventRefreshesStandaloneProjectionFromCachedEventCycleWhenRawDrained(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	runRegistry := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-manager-apply-tab-remember-review-refreshes-from-cached-cycle": {{RunID: "run-1", Status: "running"}},
		}},
	}
	manager := NewSharedSessionBrowserObserverManager(sessionRegistry, runRegistry, nil, time.Minute)
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
	sessionID := "sess-manager-apply-tab-remember-review-refreshes-from-cached-cycle"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}

	firstTabs := manager.SyncTabsForRouteEvent(sessionID, route, 1, []BrowserTab{
		{Index: 1, URL: "https://example.com/home", Title: "Home"},
	})
	if len(firstTabs) != 1 || strings.TrimSpace(firstTabs[0].TargetID) == "" {
		t.Fatalf("expected first tab sync to produce tracked target, got %#v", firstTabs)
	}
	priorSelection := SnapshotSharedSessionBrowserCurrentTargetSelection(sessionRegistry, sessionID, route)
	if priorSelection == nil || strings.TrimSpace(priorSelection.ID) == "" {
		t.Fatalf("expected prior selection snapshot after first tab sync, got %#v", priorSelection)
	}
	trackedTabs := manager.SyncTabsForRouteEvent(sessionID, route, 2, []BrowserTab{
		{Index: 1, URL: "https://example.com/home", Title: "Home"},
		{Index: 2, URL: "https://popup.example/offer", Title: "Offer"},
	})
	if len(trackedTabs) != 2 || strings.TrimSpace(trackedTabs[1].TargetID) == "" {
		t.Fatalf("expected second tab sync to produce tracked popup target, got %#v", trackedTabs)
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
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected initial watch loop to poll backend once, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
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

	rememberReview := manager.ApplyTabRememberReviewEvent(SharedSessionBrowserTabRememberReviewRequest{
		SessionID:           sessionID,
		Route:               route,
		Action:              "list",
		Force:               false,
		RememberTarget:      true,
		CandidateTargetID:   strings.TrimSpace(trackedTabs[1].TargetID),
		RequestedTabIndex:   0,
		ActiveIndex:         2,
		PriorActiveTargetID: "",
		PriorSelection:      priorSelection,
		Tabs:                trackedTabs,
	})
	if rememberReview.Decision != "session_target_popup_review_required" || rememberReview.Ready {
		t.Fatalf("expected remember review to require popup confirmation, got %#v", rememberReview)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected tab-remember review to reuse cached event-cycle source, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
	if runRegistry.callCount() != 2 {
		t.Fatalf("expected tab-remember review to refresh projection once, got %d", runRegistry.callCount())
	}

	second := bound.ObserveWatchLoop(context.Background(), req)
	if len(second.View.Session.Routes) != 1 {
		t.Fatalf("expected refreshed watch loop to expose one route, got %#v", second.View.Session.Routes)
	}
	if second.View.Session.Routes[0].CurrentTargetID != strings.TrimSpace(priorSelection.ID) {
		t.Fatalf("expected tab-remember review to restore prior current target, got %#v", second.View.Session.Routes)
	}
	if second.View.Session.Routes[0].PendingTargetReview == nil || second.View.Session.Routes[0].PendingTargetReview.ID != strings.TrimSpace(trackedTabs[1].TargetID) {
		t.Fatalf("expected tab-remember review to record pending popup review, got %#v", second.View.Session.Routes)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected next ObserveWatchLoop to reuse refreshed standalone projection cache, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
	if runRegistry.callCount() != 2 {
		t.Fatalf("expected next ObserveWatchLoop to avoid second standalone projection rebuild, got %d", runRegistry.callCount())
	}
}

func TestSharedSessionBrowserObserverManagerApplyTabRememberReviewEventRefreshesSharedProviderProjectionOnce(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	runRegistry := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-manager-apply-tab-remember-review-shared-provider": {{RunID: "run-1", Status: "running"}},
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
	sessionID := "sess-manager-apply-tab-remember-review-shared-provider"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}

	firstTabs := manager.SyncTabsForRouteEvent(sessionID, route, 1, []BrowserTab{
		{Index: 1, URL: "https://example.com/home", Title: "Home"},
	})
	priorSelection := SnapshotSharedSessionBrowserCurrentTargetSelection(sessionRegistry, sessionID, route)
	trackedTabs := manager.SyncTabsForRouteEvent(sessionID, route, 2, []BrowserTab{
		{Index: 1, URL: "https://example.com/home", Title: "Home"},
		{Index: 2, URL: "https://popup.example/offer", Title: "Offer"},
	})
	if len(firstTabs) != 1 || len(trackedTabs) != 2 || priorSelection == nil || strings.TrimSpace(trackedTabs[1].TargetID) == "" {
		t.Fatalf("expected initial tab sync to seed tracked tabs and prior selection, got first=%#v tracked=%#v prior=%#v", firstTabs, trackedTabs, priorSelection)
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

	rememberReview := manager.ApplyTabRememberReviewEvent(SharedSessionBrowserTabRememberReviewRequest{
		SessionID:           sessionID,
		Route:               route,
		Action:              "list",
		Force:               false,
		RememberTarget:      true,
		CandidateTargetID:   strings.TrimSpace(trackedTabs[1].TargetID),
		ActiveIndex:         2,
		PriorActiveTargetID: "",
		PriorSelection:      priorSelection,
		Tabs:                trackedTabs,
	})
	if rememberReview.Decision != "session_target_popup_review_required" || rememberReview.Ready {
		t.Fatalf("expected remember review to require popup confirmation, got %#v", rememberReview)
	}
	if runRegistry.callCount() != 2 {
		t.Fatalf("expected shared-provider remember review to refresh watch-loop projection once, got %d", runRegistry.callCount())
	}

	second := bound.ObserveWatchLoop(context.Background(), req)
	if len(second.View.Session.Routes) != 1 {
		t.Fatalf("expected refreshed watch loop to expose one route, got %#v", second.View.Session.Routes)
	}
	if second.View.Session.Routes[0].CurrentTargetID != strings.TrimSpace(priorSelection.ID) {
		t.Fatalf("expected remember review to restore prior current target, got %#v", second.View.Session.Routes)
	}
	if second.View.Session.Routes[0].PendingTargetReview == nil || second.View.Session.Routes[0].PendingTargetReview.ID != strings.TrimSpace(trackedTabs[1].TargetID) {
		t.Fatalf("expected remember review to record pending popup review, got %#v", second.View.Session.Routes)
	}
	if runRegistry.callCount() != 2 {
		t.Fatalf("expected next ObserveWatchLoop to reuse refreshed shared-provider projection cache, got %d", runRegistry.callCount())
	}
}

func TestSharedSessionBrowserObserverManagerApplyTabRememberReviewEventObserveWatchLoopReusesRawRouteMutationSourceWhenRouteMutationCycleDrained(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	runRegistry := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-manager-apply-tab-remember-review-raw-route-mutation": {{RunID: "run-1", Status: "running"}},
		}},
	}
	manager := NewSharedSessionBrowserObserverManager(sessionRegistry, runRegistry, stateRegistry, time.Minute)
	backend := &statusProfilesObservationTestBackend{}
	bound := manager.Bind(backend)
	sessionID := "sess-manager-apply-tab-remember-review-raw-route-mutation"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}
	stateRegistry.SyncSessionBrowserProfileObservation(sessionID, SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "running",
		Running:       true,
		Connected:     true,
	}, time.Minute)

	firstTabs := manager.SyncTabsForRouteEvent(sessionID, route, 1, []BrowserTab{
		{Index: 1, URL: "https://example.com/home", Title: "Home"},
	})
	if len(firstTabs) != 1 || strings.TrimSpace(firstTabs[0].TargetID) == "" {
		t.Fatalf("expected first tab sync to produce tracked target, got %#v", firstTabs)
	}
	priorSelection := SnapshotSharedSessionBrowserCurrentTargetSelection(sessionRegistry, sessionID, route)
	if priorSelection == nil || strings.TrimSpace(priorSelection.ID) == "" {
		t.Fatalf("expected prior selection snapshot after first tab sync, got %#v", priorSelection)
	}
	trackedTabs := manager.SyncTabsForRouteEvent(sessionID, route, 2, []BrowserTab{
		{Index: 1, URL: "https://example.com/home", Title: "Home"},
		{Index: 2, URL: "https://popup.example/offer", Title: "Offer"},
	})
	if len(trackedTabs) != 2 || strings.TrimSpace(trackedTabs[1].TargetID) == "" {
		t.Fatalf("expected second tab sync to produce tracked popup target, got %#v", trackedTabs)
	}

	rememberReview := manager.ApplyTabRememberReviewEvent(SharedSessionBrowserTabRememberReviewRequest{
		SessionID:           sessionID,
		Route:               route,
		Action:              "list",
		Force:               false,
		RememberTarget:      true,
		CandidateTargetID:   strings.TrimSpace(trackedTabs[1].TargetID),
		RequestedTabIndex:   0,
		ActiveIndex:         2,
		PriorActiveTargetID: "",
		PriorSelection:      priorSelection,
		Tabs:                trackedTabs,
	})
	if rememberReview.Decision != "session_target_popup_review_required" || rememberReview.Ready {
		t.Fatalf("expected remember review to require popup confirmation, got %#v", rememberReview)
	}

	bound.state.mu.Lock()
	clear(bound.state.rawStatus)
	clear(bound.state.rawProfiles)
	clear(bound.state.routeMutations)
	clear(bound.state.eventCycles)
	clear(bound.state.bindings)
	clear(bound.state.views)
	clear(bound.state.watchLoops)
	clear(bound.state.eventCyclesInFlight)
	clear(bound.state.bindingsInFlight)
	clear(bound.state.viewsInFlight)
	clear(bound.state.watchLoopsInFlight)
	rawRouteMutationCount := len(bound.state.rawRouteMutations)
	bound.state.mu.Unlock()
	if rawRouteMutationCount == 0 {
		t.Fatalf("expected raw route-mutation source before draining route-mutation cycle cache")
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

	watch := bound.ObserveWatchLoop(context.Background(), req)
	if len(watch.View.Session.Routes) != 1 {
		t.Fatalf("expected watch loop to expose one route, got %#v", watch.View.Session.Routes)
	}
	if watch.View.Session.Routes[0].CurrentTargetID != strings.TrimSpace(priorSelection.ID) {
		t.Fatalf("expected raw route-mutation source to restore prior current target, got %#v", watch.View.Session.Routes)
	}
	if watch.View.Session.Routes[0].PendingTargetReview == nil || watch.View.Session.Routes[0].PendingTargetReview.ID != strings.TrimSpace(trackedTabs[1].TargetID) {
		t.Fatalf("expected raw route-mutation source to restore pending popup review, got %#v", watch.View.Session.Routes)
	}
	if len(backend.statusReqs) != 0 || len(backend.profilesReqs) != 0 {
		t.Fatalf("expected raw route-mutation source to avoid RuntimeStatus/RuntimeProfiles polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}

func TestSharedSessionBrowserObserverManagerExecuteClearTargetInvalidatesBoundWatchLoopCache(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	manager := NewSharedSessionBrowserObserverManager(sessionRegistry, nil, nil, time.Minute)
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
	sessionID := "sess-manager-clear-target-cache"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node"}

	sessionRegistry.TrackTab(sessionID, BrowserSessionTarget{
		ID:         "tab-1",
		TabIndex:   1,
		URL:        "https://example.com",
		Title:      "Example",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
	}, true)

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
	if len(first.View.Session.Routes) != 1 || first.View.Session.Routes[0].CurrentTargetID == "" {
		t.Fatalf("expected initial watch loop to expose current target, got %#v", first.View.Session.Routes)
	}

	result := manager.ExecuteClearTarget(BuildSharedSessionBrowserClearRequest(
		sessionRegistry,
		nil,
		sessionID,
		BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		route,
		false,
		"",
		SharedSessionBrowserHealthInput{},
		time.Minute,
	))
	if !result.Ready || !result.ClearedTargetSelection {
		t.Fatalf("expected clear target to clear tracked current target, got %#v", result)
	}

	second := bound.ObserveWatchLoop(context.Background(), req)
	if len(second.View.Session.Routes) != 1 {
		t.Fatalf("expected invalidated watch loop to preserve route snapshot, got %#v", second.View.Session.Routes)
	}
	if second.View.Session.Routes[0].CurrentTargetID != "" {
		t.Fatalf("expected clear target to invalidate bound watch-loop cache, got %#v", second.View.Session.Routes)
	}
}

func TestSharedSessionBrowserWatchManagerObserveWatchLoopUsesRawTabsSourceWhenCachesDrained(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	runRegistry := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-manager-raw-tabs-source": {{RunID: "run-1", Status: "running"}},
		}},
	}
	manager := NewSharedSessionBrowserObserverManager(sessionRegistry, runRegistry, stateRegistry, time.Minute)
	backend := &statusProfilesObservationTestBackend{
		rawTabs: func(_ context.Context, requestedProfile string) SharedSessionBrowserRawTabsObservation {
			if requestedProfile != "isolated" {
				t.Fatalf("expected requested profile to reach raw tabs source, got %q", requestedProfile)
			}
			return SharedSessionBrowserRawTabsObservation{
				RequestedProfile: "isolated",
				Action:           "list",
				ActiveIndex:      2,
				Tabs: []BrowserTab{
					{Index: 1, URL: "https://example.com/home", Title: "Home", Active: false},
					{Index: 2, URL: "https://example.com/popup", Title: "Popup", Active: true},
				},
				ObservedAt: time.Now().Add(-500 * time.Millisecond),
			}
		},
	}
	bound := manager.Bind(backend)
	sessionID := "sess-manager-raw-tabs-source"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}
	stateRegistry.SyncSessionBrowserProfileObservation(sessionID, SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "running",
		Running:       true,
		Connected:     true,
	}, time.Minute)

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

	watch := bound.ObserveWatchLoop(context.Background(), req)
	if len(watch.View.Session.Routes) != 1 {
		t.Fatalf("expected raw tabs source to seed one session route, got %#v", watch.View.Session.Routes)
	}
	if len(watch.View.Session.Routes[0].Targets) != 2 || watch.View.Session.Routes[0].CurrentTargetID == "" {
		t.Fatalf("expected raw tabs source to seed current target and tracked tabs, got %#v", watch.View.Session.Routes)
	}
	if watch.View.Session.Routes[0].Targets[1].Title != "Popup" {
		t.Fatalf("expected raw tabs source to preserve tab metadata, got %#v", watch.View.Session.Routes[0].Targets)
	}
	if len(backend.statusReqs) != 0 || len(backend.profilesReqs) != 0 {
		t.Fatalf("expected raw tabs source to avoid RuntimeStatus/RuntimeProfiles polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
	if backend.rawTabsCalls != 1 {
		t.Fatalf("expected raw tabs source to be consumed once, got %d", backend.rawTabsCalls)
	}
}

func TestSharedSessionBrowserWatchManagerObserveWatchLoopUsesRawTabsResultSourceForRememberReviewWhenCachesDrained(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	runRegistry := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-manager-raw-tabs-remember-review-source": {{RunID: "run-1", Status: "running"}},
		}},
	}
	manager := NewSharedSessionBrowserObserverManager(sessionRegistry, runRegistry, stateRegistry, time.Minute)
	var priorSelection *BrowserSessionTargetSelection
	backend := &statusProfilesObservationTestBackend{
		rawTabs: func(_ context.Context, requestedProfile string) SharedSessionBrowserRawTabsObservation {
			if requestedProfile != "isolated" {
				t.Fatalf("expected requested profile to reach raw tabs source, got %q", requestedProfile)
			}
			return SharedSessionBrowserRawTabsObservation{
				RequestedProfile: "isolated",
				Action:           "list",
				RememberTarget:   true,
				Actor:            "browser_tabs list",
				PriorSelection:   priorSelection,
				Note:             "tabs ok",
				ActiveIndex:      2,
				Tabs: []BrowserTab{
					{Index: 1, URL: "https://example.com/home", Title: "Home", Active: false},
					{Index: 2, URL: "https://popup.example/offer", Title: "Offer", Active: true},
				},
				ObservedAt: time.Now().Add(-500 * time.Millisecond),
			}
		},
	}
	bound := manager.Bind(backend)
	sessionID := "sess-manager-raw-tabs-remember-review-source"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}
	stateRegistry.SyncSessionBrowserProfileObservation(sessionID, SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "running",
		Running:       true,
		Connected:     true,
	}, time.Minute)

	firstTabs := manager.SyncTabsForRouteEvent(sessionID, route, 1, []BrowserTab{
		{Index: 1, URL: "https://example.com/home", Title: "Home"},
	})
	if len(firstTabs) != 1 || strings.TrimSpace(firstTabs[0].TargetID) == "" {
		t.Fatalf("expected initial tab sync to create tracked target, got %#v", firstTabs)
	}
	priorSelection = SnapshotSharedSessionBrowserCurrentTargetSelection(sessionRegistry, sessionID, route)
	if priorSelection == nil || strings.TrimSpace(priorSelection.ID) == "" {
		t.Fatalf("expected prior selection snapshot after initial sync, got %#v", priorSelection)
	}
	bound.state.mu.Lock()
	clear(bound.state.rawStatus)
	clear(bound.state.rawProfiles)
	clear(bound.state.routeMutations)
	clear(bound.state.eventCycles)
	clear(bound.state.watchLoops)
	bound.state.mu.Unlock()

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

	watch := bound.ObserveWatchLoop(context.Background(), req)
	if len(watch.View.Session.Routes) != 1 {
		t.Fatalf("expected one session route from raw tabs-result source, got %#v", watch.View.Session.Routes)
	}
	seeded := watch.View.Session.Routes[0]
	if seeded.PendingTargetReview == nil {
		t.Fatalf("expected raw tabs-result source to record pending popup review, got %#v", seeded)
	}
	if seeded.PendingTargetReview.Decision != "session_target_popup_review_required" {
		t.Fatalf("expected raw tabs-result source to require popup review, got %#v", seeded.PendingTargetReview)
	}
	if seeded.CurrentTargetID != priorSelection.ID {
		t.Fatalf("expected raw tabs-result source to restore prior current target, got %#v", seeded)
	}
	if !strings.Contains(seeded.PendingTargetReview.Reason, "rerun with force=true") {
		t.Fatalf("expected remember-review reason from raw tabs-result source, got %#v", seeded.PendingTargetReview)
	}
	if len(seeded.Targets) != 2 {
		t.Fatalf("expected raw tabs-result source to keep tracked tabs, got %#v", seeded.Targets)
	}
	if len(backend.statusReqs) != 0 || len(backend.profilesReqs) != 0 {
		t.Fatalf("expected raw tabs-result source to avoid RuntimeStatus/RuntimeProfiles polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
	if backend.rawTabsCalls != 1 {
		t.Fatalf("expected raw tabs-result source to be consumed once, got %d", backend.rawTabsCalls)
	}
}

func TestSharedSessionBrowserWatchManagerObserveWatchLoopUsesBackendRawRouteMutationTabsResultAfterRawTabsConsumed(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	runRegistry := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-manager-backend-raw-route-mutation-tabs-remember-review": {{RunID: "run-1", Status: "running"}},
		}},
	}
	manager := NewSharedSessionBrowserObserverManager(sessionRegistry, runRegistry, stateRegistry, time.Minute)
	rawTabsDelivered := false
	rawRouteMutationDelivered := false
	var priorSelection *BrowserSessionTargetSelection
	backend := &statusProfilesObservationTestBackend{
		rawTabs: func(_ context.Context, requestedProfile string) SharedSessionBrowserRawTabsObservation {
			if requestedProfile != "isolated" {
				t.Fatalf("expected requested profile isolated, got %q", requestedProfile)
			}
			if rawTabsDelivered {
				return SharedSessionBrowserRawTabsObservation{}
			}
			rawTabsDelivered = true
			return SharedSessionBrowserRawTabsObservation{
				RequestedProfile: requestedProfile,
				Action:           "list",
				RememberTarget:   true,
				PriorSelection:   priorSelection,
				Tabs: []BrowserTab{
					{Index: 1, URL: "https://example.com/home", Title: "Home", TargetID: "tab-home"},
					{Index: 2, URL: "https://popup.example/offer", Title: "Offer", TargetID: "tab-popup", Active: true},
				},
				ActiveIndex: 2,
				ObservedAt:  time.Now().Add(-time.Second),
			}
		},
		rawRouteMutation: func(_ context.Context, requestedProfile string) SharedSessionBrowserRawRouteMutationObservation {
			if requestedProfile != "isolated" {
				t.Fatalf("expected requested profile isolated, got %q", requestedProfile)
			}
			if rawRouteMutationDelivered {
				return SharedSessionBrowserRawRouteMutationObservation{}
			}
			rawRouteMutationDelivered = true
			return SharedSessionBrowserRawRouteMutationObservation{
				RequestedProfile:  requestedProfile,
				Route:             BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"},
				Kind:              "tabs_result",
				Action:            "list",
				RememberTarget:    true,
				CandidateTargetID: "tab-popup",
				ActiveIndex:       2,
				PriorSelection:    priorSelection,
				Tabs: []BrowserTab{
					{Index: 1, URL: "https://example.com/home", Title: "Home", TargetID: "tab-home"},
					{Index: 2, URL: "https://popup.example/offer", Title: "Offer", TargetID: "tab-popup", Active: true},
				},
				ObservedAt: time.Now().Add(-time.Second),
			}
		},
	}
	bound := manager.Bind(backend)
	sessionID := "sess-manager-backend-raw-route-mutation-tabs-remember-review"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}
	stateRegistry.SyncSessionBrowserProfileObservation(sessionID, SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "running",
		Running:       true,
		Connected:     true,
	}, time.Minute)
	firstTabs := manager.SyncTabsForRouteEvent(sessionID, route, 1, []BrowserTab{
		{Index: 1, URL: "https://example.com/home", Title: "Home"},
	})
	if len(firstTabs) != 1 || strings.TrimSpace(firstTabs[0].TargetID) == "" {
		t.Fatalf("expected initial tab sync to create tracked target, got %#v", firstTabs)
	}
	priorSelection = SnapshotSharedSessionBrowserCurrentTargetSelection(sessionRegistry, sessionID, route)
	if priorSelection == nil || strings.TrimSpace(priorSelection.ID) == "" {
		t.Fatalf("expected prior selection snapshot after initial sync, got %#v", priorSelection)
	}
	bound.state.mu.Lock()
	clear(bound.state.rawStatus)
	clear(bound.state.rawProfiles)
	clear(bound.state.routeMutations)
	clear(bound.state.eventCycles)
	clear(bound.state.watchLoops)
	bound.state.mu.Unlock()

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
	if len(first.View.Session.Routes) != 1 {
		t.Fatalf("expected first watch loop to expose one route, got %#v", first.View.Session.Routes)
	}
	if first.View.Session.Routes[0].PendingTargetReview == nil || first.View.Session.Routes[0].PendingTargetReview.URL != "https://popup.example/offer" {
		t.Fatalf("expected raw tabs source to expose pending popup review, got %#v", first.View.Session.Routes)
	}
	if first.View.Session.Routes[0].CurrentTargetID != priorSelection.ID {
		t.Fatalf("expected raw tabs source to restore prior current target, got %#v", first.View.Session.Routes)
	}
	if backend.rawTabsCalls != 1 || backend.rawRouteMutationCalls != 0 {
		t.Fatalf("expected first observe to consume only raw tabs source, got rawTabs=%d rawRouteMutation=%d", backend.rawTabsCalls, backend.rawRouteMutationCalls)
	}

	bound.state.mu.Lock()
	clear(bound.state.rawStatus)
	clear(bound.state.rawProfiles)
	clear(bound.state.rawRouteMutations)
	clear(bound.state.routeMutations)
	clear(bound.state.eventCycles)
	clear(bound.state.bindings)
	clear(bound.state.views)
	clear(bound.state.watchLoops)
	clear(bound.state.eventCyclesInFlight)
	clear(bound.state.bindingsInFlight)
	clear(bound.state.viewsInFlight)
	clear(bound.state.watchLoopsInFlight)
	bound.state.mu.Unlock()

	second := bound.ObserveWatchLoop(context.Background(), req)
	if len(second.View.Session.Routes) != 1 {
		t.Fatalf("expected second watch loop to expose one route, got %#v", second.View.Session.Routes)
	}
	if second.View.Session.Routes[0].PendingTargetReview == nil || second.View.Session.Routes[0].PendingTargetReview.URL != "https://popup.example/offer" {
		t.Fatalf("expected backend raw tabs-result route-mutation source to restore pending popup review, got %#v", second.View.Session.Routes)
	}
	if second.View.Session.Routes[0].CurrentTargetID != priorSelection.ID {
		t.Fatalf("expected backend raw tabs-result route-mutation source to restore prior current target, got %#v", second.View.Session.Routes)
	}
	if backend.rawTabsCalls != 2 || backend.rawRouteMutationCalls != 1 {
		t.Fatalf("expected second observe to fall back to backend raw route-mutation after raw tabs drain, got rawTabs=%d rawRouteMutation=%d", backend.rawTabsCalls, backend.rawRouteMutationCalls)
	}
	if len(backend.statusReqs) != 0 || len(backend.profilesReqs) != 0 {
		t.Fatalf("expected raw tabs/raw route-mutation sources to avoid RuntimeStatus/RuntimeProfiles polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}

func TestSharedSessionBrowserObserverManagerSyncTabsForRouteEventObserveWatchLoopReusesRawRouteMutationSourceWhenRouteMutationCycleDrained(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	runRegistry := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-manager-raw-route-mutation-sync-tabs": {{RunID: "run-1", Status: "running"}},
		}},
	}
	manager := NewSharedSessionBrowserObserverManager(sessionRegistry, runRegistry, stateRegistry, time.Minute)
	backend := &statusProfilesObservationTestBackend{}
	bound := manager.Bind(backend)
	sessionID := "sess-manager-raw-route-mutation-sync-tabs"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}
	stateRegistry.SyncSessionBrowserProfileObservation(sessionID, SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "running",
		Running:       true,
		Connected:     true,
	}, time.Minute)

	manager.SyncTabsForRouteEvent(sessionID, route, 2, []BrowserTab{
		{Index: 1, URL: "https://example.com/one", Title: "One"},
		{Index: 2, URL: "https://example.com/two", Title: "Two"},
	})

	bound.state.mu.Lock()
	clear(bound.state.rawStatus)
	clear(bound.state.rawProfiles)
	clear(bound.state.routeMutations)
	clear(bound.state.eventCycles)
	clear(bound.state.bindings)
	clear(bound.state.views)
	clear(bound.state.watchLoops)
	clear(bound.state.eventCyclesInFlight)
	clear(bound.state.bindingsInFlight)
	clear(bound.state.viewsInFlight)
	clear(bound.state.watchLoopsInFlight)
	rawRouteMutationCount := len(bound.state.rawRouteMutations)
	bound.state.mu.Unlock()
	if rawRouteMutationCount == 0 {
		t.Fatalf("expected raw route-mutation source before draining route-mutation cycle cache")
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

	watch := bound.ObserveWatchLoop(context.Background(), req)
	if len(watch.View.Session.Routes) != 1 || len(watch.View.Session.Routes[0].Targets) != 2 {
		t.Fatalf("expected raw route-mutation source to restore tracked tabs, got %#v", watch.View.Session.Routes)
	}
	if watch.View.Session.Routes[0].CurrentTargetID == "" || watch.View.Session.Routes[0].Targets[1].Title != "Two" {
		t.Fatalf("expected raw route-mutation source to restore active tab projection, got %#v", watch.View.Session.Routes)
	}
	if len(backend.statusReqs) != 0 || len(backend.profilesReqs) != 0 {
		t.Fatalf("expected raw route-mutation source to avoid RuntimeStatus/RuntimeProfiles polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}

func TestSharedSessionBrowserWatchManagerObserveWatchLoopUsesBackendRawRouteMutationSourceWhenSyntheticCachesMissing(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	runRegistry := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-manager-backend-raw-route-mutation-sync-tabs": {{RunID: "run-1", Status: "running"}},
		}},
	}
	manager := NewSharedSessionBrowserObserverManager(sessionRegistry, runRegistry, stateRegistry, time.Minute)
	backend := &statusProfilesObservationTestBackend{
		rawRouteMutation: func(_ context.Context, requestedProfile string) SharedSessionBrowserRawRouteMutationObservation {
			if requestedProfile != "isolated" {
				t.Fatalf("expected requested profile isolated, got %q", requestedProfile)
			}
			return SharedSessionBrowserRawRouteMutationObservation{
				RequestedProfile: requestedProfile,
				Route:            BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"},
				Kind:             "sync_tabs",
				ActiveIndex:      2,
				Tabs: []BrowserTab{
					{Index: 1, URL: "https://example.com/home", Title: "Home"},
					{Index: 2, URL: "https://example.com/popup", Title: "Popup"},
				},
				ObservedAt: time.Now().Add(-time.Second),
			}
		},
	}
	bound := manager.Bind(backend)
	sessionID := "sess-manager-backend-raw-route-mutation-sync-tabs"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}
	stateRegistry.SyncSessionBrowserProfileObservation(sessionID, SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "running",
		Running:       true,
		Connected:     true,
	}, time.Minute)

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

	watch := bound.ObserveWatchLoop(context.Background(), req)
	if len(watch.View.Session.Routes) != 1 || len(watch.View.Session.Routes[0].Targets) != 2 {
		t.Fatalf("expected backend raw route-mutation source to restore tracked tabs, got %#v", watch.View.Session.Routes)
	}
	if watch.View.Session.Routes[0].CurrentTargetID == "" || watch.View.Session.Routes[0].Targets[1].Title != "Popup" {
		t.Fatalf("expected backend raw route-mutation source to restore active tab projection, got %#v", watch.View.Session.Routes)
	}
	if backend.rawRouteMutationCalls != 1 {
		t.Fatalf("expected backend raw route-mutation source to be consumed once, got %d", backend.rawRouteMutationCalls)
	}
	if len(backend.statusReqs) != 0 || len(backend.profilesReqs) != 0 {
		t.Fatalf("expected backend raw route-mutation source to avoid RuntimeStatus/RuntimeProfiles polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}

func TestSharedSessionBrowserWatchManagerObserveWatchLoopUsesBackendRawRouteMutationSourceAfterRawOpenConsumed(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	runRegistry := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-manager-backend-raw-route-mutation-open": {{RunID: "run-1", Status: "running"}},
		}},
	}
	manager := NewSharedSessionBrowserObserverManager(sessionRegistry, runRegistry, stateRegistry, time.Minute)
	rawOpenDelivered := false
	rawRouteMutationDelivered := false
	backend := &statusProfilesObservationTestBackend{
		rawOpen: func(_ context.Context, requestedProfile string) SharedSessionBrowserRawOpenObservation {
			if requestedProfile != "isolated" {
				t.Fatalf("expected requested profile isolated, got %q", requestedProfile)
			}
			if rawOpenDelivered {
				return SharedSessionBrowserRawOpenObservation{}
			}
			rawOpenDelivered = true
			return SharedSessionBrowserRawOpenObservation{
				RequestedProfile: requestedProfile,
				URL:              "https://example.com/opened",
				ObservedAt:       time.Now().Add(-time.Second),
			}
		},
		rawRouteMutation: func(_ context.Context, requestedProfile string) SharedSessionBrowserRawRouteMutationObservation {
			if requestedProfile != "isolated" {
				t.Fatalf("expected requested profile isolated, got %q", requestedProfile)
			}
			if rawRouteMutationDelivered {
				return SharedSessionBrowserRawRouteMutationObservation{}
			}
			rawRouteMutationDelivered = true
			return SharedSessionBrowserRawRouteMutationObservation{
				RequestedProfile: requestedProfile,
				Route:            BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node"},
				Kind:             "open_result",
				URL:              "https://example.com/opened",
				Source:           "runtime_open_source",
				ObservedAt:       time.Now().Add(-time.Second),
			}
		},
	}
	bound := manager.Bind(backend)
	sessionID := "sess-manager-backend-raw-route-mutation-open"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node"}
	stateRegistry.SyncSessionBrowserProfileObservation(sessionID, SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		Status:        "running",
		Running:       true,
		Connected:     true,
	}, time.Minute)

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
	if len(first.View.Session.Routes) != 1 || first.View.Session.Routes[0].CurrentTargetID == "" {
		t.Fatalf("expected raw open source to restore current target, got %#v", first.View.Session.Routes)
	}
	if backend.rawOpenCalls != 1 || backend.rawRouteMutationCalls != 0 {
		t.Fatalf("expected first observe to consume only raw open source, got rawOpen=%d rawRouteMutation=%d", backend.rawOpenCalls, backend.rawRouteMutationCalls)
	}

	bound.state.mu.Lock()
	clear(bound.state.rawStatus)
	clear(bound.state.rawProfiles)
	clear(bound.state.rawRouteMutations)
	clear(bound.state.routeMutations)
	clear(bound.state.eventCycles)
	clear(bound.state.bindings)
	clear(bound.state.views)
	clear(bound.state.watchLoops)
	clear(bound.state.eventCyclesInFlight)
	clear(bound.state.bindingsInFlight)
	clear(bound.state.viewsInFlight)
	clear(bound.state.watchLoopsInFlight)
	bound.state.mu.Unlock()

	second := bound.ObserveWatchLoop(context.Background(), req)
	if len(second.View.Session.Routes) != 1 || second.View.Session.Routes[0].CurrentTargetID == "" {
		t.Fatalf("expected backend raw route-mutation source to restore current target after raw open drain, got %#v", second.View.Session.Routes)
	}
	if backend.rawOpenCalls != 2 || backend.rawRouteMutationCalls != 1 {
		t.Fatalf("expected second observe to fall back to backend raw route-mutation after raw open drain, got rawOpen=%d rawRouteMutation=%d", backend.rawOpenCalls, backend.rawRouteMutationCalls)
	}
	if len(backend.statusReqs) != 0 || len(backend.profilesReqs) != 0 {
		t.Fatalf("expected raw open/raw route-mutation sources to avoid RuntimeStatus/RuntimeProfiles polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}

func TestSharedSessionBrowserObserverManagerForgetTabForRouteEventObserveWatchLoopReusesRawRouteMutationSourceWhenRouteMutationCycleDrained(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	runRegistry := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-manager-raw-route-mutation-forget-tab": {{RunID: "run-1", Status: "running"}},
		}},
	}
	manager := NewSharedSessionBrowserObserverManager(sessionRegistry, runRegistry, stateRegistry, time.Minute)
	backend := &statusProfilesObservationTestBackend{}
	bound := manager.Bind(backend)
	sessionID := "sess-manager-raw-route-mutation-forget-tab"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}
	stateRegistry.SyncSessionBrowserProfileObservation(sessionID, SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "running",
		Running:       true,
		Connected:     true,
	}, time.Minute)

	manager.SyncTabsForRouteEvent(sessionID, route, 2, []BrowserTab{
		{Index: 1, URL: "https://example.com/one", Title: "One"},
		{Index: 2, URL: "https://example.com/two", Title: "Two"},
	})
	manager.ForgetTabForRouteEvent(sessionID, route, 1)

	bound.state.mu.Lock()
	clear(bound.state.rawStatus)
	clear(bound.state.rawProfiles)
	clear(bound.state.routeMutations)
	clear(bound.state.eventCycles)
	clear(bound.state.bindings)
	clear(bound.state.views)
	clear(bound.state.watchLoops)
	clear(bound.state.eventCyclesInFlight)
	clear(bound.state.bindingsInFlight)
	clear(bound.state.viewsInFlight)
	clear(bound.state.watchLoopsInFlight)
	rawRouteMutationCount := len(bound.state.rawRouteMutations)
	bound.state.mu.Unlock()
	if rawRouteMutationCount == 0 {
		t.Fatalf("expected raw route-mutation source before draining route-mutation cycle cache")
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

	watch := bound.ObserveWatchLoop(context.Background(), req)
	if len(watch.View.Session.Routes) != 1 || len(watch.View.Session.Routes[0].Targets) != 1 {
		t.Fatalf("expected raw route-mutation source to restore detached tab projection, got %#v", watch.View.Session.Routes)
	}
	if watch.View.Session.Routes[0].Targets[0].TabIndex != 2 {
		t.Fatalf("expected raw route-mutation source to preserve surviving tab, got %#v", watch.View.Session.Routes)
	}
	if len(backend.statusReqs) != 0 || len(backend.profilesReqs) != 0 {
		t.Fatalf("expected raw route-mutation source to avoid RuntimeStatus/RuntimeProfiles polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}

func TestSharedSessionBrowserObserverManagerTrackCurrentTargetEventObserveWatchLoopReusesRawRouteMutationSourceWhenRouteMutationCycleDrained(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	runRegistry := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-manager-raw-route-mutation-track-current": {{RunID: "run-1", Status: "running"}},
		}},
	}
	manager := NewSharedSessionBrowserObserverManager(sessionRegistry, runRegistry, stateRegistry, time.Minute)
	backend := &statusProfilesObservationTestBackend{}
	bound := manager.Bind(backend)
	sessionID := "sess-manager-raw-route-mutation-track-current"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}
	stateRegistry.SyncSessionBrowserProfileObservation(sessionID, SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "running",
		Running:       true,
		Connected:     true,
	}, time.Minute)

	manager.TrackCurrentTargetEvent(sessionID, route, "https://example.com/current", "Current", "browser_navigate")

	bound.state.mu.Lock()
	clear(bound.state.rawStatus)
	clear(bound.state.rawProfiles)
	clear(bound.state.routeMutations)
	clear(bound.state.eventCycles)
	clear(bound.state.bindings)
	clear(bound.state.views)
	clear(bound.state.watchLoops)
	clear(bound.state.eventCyclesInFlight)
	clear(bound.state.bindingsInFlight)
	clear(bound.state.viewsInFlight)
	clear(bound.state.watchLoopsInFlight)
	rawRouteMutationCount := len(bound.state.rawRouteMutations)
	bound.state.mu.Unlock()
	if rawRouteMutationCount == 0 {
		t.Fatalf("expected raw route-mutation source before draining route-mutation cycle cache")
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

	watch := bound.ObserveWatchLoop(context.Background(), req)
	if len(watch.View.Session.Routes) != 1 || watch.View.Session.Routes[0].CurrentTargetID == "" {
		t.Fatalf("expected raw route-mutation source to restore current target projection, got %#v", watch.View.Session.Routes)
	}
	if len(watch.View.Session.Routes[0].Targets) != 1 || watch.View.Session.Routes[0].Targets[0].URL != "https://example.com/current" {
		t.Fatalf("expected raw route-mutation source to restore tracked current target, got %#v", watch.View.Session.Routes)
	}
	if len(backend.statusReqs) != 0 || len(backend.profilesReqs) != 0 {
		t.Fatalf("expected raw route-mutation source to avoid RuntimeStatus/RuntimeProfiles polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}

func TestSharedSessionBrowserObserverManagerRecordPendingTargetPopupReviewEventObserveWatchLoopReusesRawRouteMutationSourceWhenRouteMutationCycleDrained(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	runRegistry := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-manager-raw-route-mutation-pending-popup-review": {{RunID: "run-1", Status: "running"}},
		}},
	}
	manager := NewSharedSessionBrowserObserverManager(sessionRegistry, runRegistry, stateRegistry, time.Minute)
	backend := &statusProfilesObservationTestBackend{}
	bound := manager.Bind(backend)
	sessionID := "sess-manager-raw-route-mutation-pending-popup-review"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}
	stateRegistry.SyncSessionBrowserProfileObservation(sessionID, SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "running",
		Running:       true,
		Connected:     true,
	}, time.Minute)

	popup := manager.TrackTabEvent(sessionID, route, BrowserTab{
		Index: 2,
		URL:   "https://popup.example/offer",
		Title: "Offer",
	}, false)
	if popup.ID == "" {
		t.Fatalf("expected track-tab event to create popup target, got %#v", popup)
	}

	review := manager.RecordPendingTargetPopupReviewEvent(
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
	if review == nil || review.ID != popup.ID {
		t.Fatalf("expected popup review event to resolve tracked popup target, got %#v", review)
	}

	bound.state.mu.Lock()
	clear(bound.state.rawStatus)
	clear(bound.state.rawProfiles)
	clear(bound.state.routeMutations)
	clear(bound.state.eventCycles)
	clear(bound.state.bindings)
	clear(bound.state.views)
	clear(bound.state.watchLoops)
	clear(bound.state.eventCyclesInFlight)
	clear(bound.state.bindingsInFlight)
	clear(bound.state.viewsInFlight)
	clear(bound.state.watchLoopsInFlight)
	rawRouteMutationCount := len(bound.state.rawRouteMutations)
	bound.state.mu.Unlock()
	if rawRouteMutationCount == 0 {
		t.Fatalf("expected raw route-mutation source before draining route-mutation cycle cache")
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

	watch := bound.ObserveWatchLoop(context.Background(), req)
	if len(watch.View.Session.Routes) != 1 || watch.View.Session.Routes[0].PendingTargetReview == nil || watch.View.Session.Routes[0].PendingTargetReview.ID != popup.ID {
		t.Fatalf("expected raw route-mutation source to restore pending popup review, got %#v", watch.View.Session.Routes)
	}
	if len(backend.statusReqs) != 0 || len(backend.profilesReqs) != 0 {
		t.Fatalf("expected raw route-mutation source to avoid RuntimeStatus/RuntimeProfiles polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}

func TestSharedSessionBrowserObserverManagerRestoreCurrentTargetSelectionForPendingReviewEventObserveWatchLoopReusesRawRouteMutationSourceWhenRouteMutationCycleDrained(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	runRegistry := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-manager-raw-route-mutation-restore-pending-review": {{RunID: "run-1", Status: "running"}},
		}},
	}
	manager := NewSharedSessionBrowserObserverManager(sessionRegistry, runRegistry, stateRegistry, time.Minute)
	backend := &statusProfilesObservationTestBackend{}
	bound := manager.Bind(backend)
	sessionID := "sess-manager-raw-route-mutation-restore-pending-review"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}
	stateRegistry.SyncSessionBrowserProfileObservation(sessionID, SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "running",
		Running:       true,
		Connected:     true,
	}, time.Minute)

	prior := manager.TrackTabEvent(sessionID, route, BrowserTab{
		Index: 1,
		URL:   "https://example.com/one",
		Title: "One",
	}, true)
	if prior.ID == "" {
		t.Fatalf("expected first track-tab event to create prior target, got %#v", prior)
	}
	priorSelection := SnapshotSharedSessionBrowserCurrentTargetSelection(sessionRegistry, sessionID, route)
	if priorSelection == nil || priorSelection.ID != prior.ID {
		t.Fatalf("expected prior current-target snapshot to match first target, got %#v", priorSelection)
	}
	pending := manager.TrackTabEvent(sessionID, route, BrowserTab{
		Index: 2,
		URL:   "https://popup.example/offer",
		Title: "Offer",
	}, true)
	if pending.ID == "" {
		t.Fatalf("expected second track-tab event to create pending target, got %#v", pending)
	}
	review := manager.RecordPendingTargetPopupReviewEvent(
		sessionID,
		route,
		BrowserTab{
			Index:    pending.TabIndex,
			URL:      pending.URL,
			Title:    pending.Title,
			TargetID: pending.ID,
		},
		"session_target_popup_review_required",
		"pending popup review",
	)
	if review == nil || review.ID != pending.ID {
		t.Fatalf("expected popup review state before restore, got %#v", review)
	}

	restored := manager.RestoreCurrentTargetSelectionForPendingReviewEvent(
		sessionID,
		route,
		priorSelection,
		pending.ID,
		"popup_review_restore",
	)
	if restored == nil || restored.ID != prior.ID {
		t.Fatalf("expected restore event to switch current selection back to prior target, got %#v", restored)
	}

	bound.state.mu.Lock()
	clear(bound.state.rawStatus)
	clear(bound.state.rawProfiles)
	clear(bound.state.routeMutations)
	clear(bound.state.eventCycles)
	clear(bound.state.bindings)
	clear(bound.state.views)
	clear(bound.state.watchLoops)
	clear(bound.state.eventCyclesInFlight)
	clear(bound.state.bindingsInFlight)
	clear(bound.state.viewsInFlight)
	clear(bound.state.watchLoopsInFlight)
	rawRouteMutationCount := len(bound.state.rawRouteMutations)
	bound.state.mu.Unlock()
	if rawRouteMutationCount == 0 {
		t.Fatalf("expected raw route-mutation source before draining route-mutation cycle cache")
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

	watch := bound.ObserveWatchLoop(context.Background(), req)
	if len(watch.View.Session.Routes) != 1 || watch.View.Session.Routes[0].CurrentTargetID != prior.ID {
		t.Fatalf("expected raw route-mutation source to restore prior current target, got %#v", watch.View.Session.Routes)
	}
	if watch.View.Session.Routes[0].PendingTargetReview == nil || watch.View.Session.Routes[0].PendingTargetReview.ID != pending.ID {
		t.Fatalf("expected raw route-mutation source to preserve pending popup review, got %#v", watch.View.Session.Routes)
	}
	if len(backend.statusReqs) != 0 || len(backend.profilesReqs) != 0 {
		t.Fatalf("expected raw route-mutation source to avoid RuntimeStatus/RuntimeProfiles polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}

func TestSharedSessionBrowserObserverManagerSelectTargetBumpsGenerationOnceWhenProviderBacked(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	sessionID := "sess-manager-select-target-generation"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node"}

	first := sessionRegistry.TrackTab(sessionID, BrowserSessionTarget{
		ID:         "tab-1",
		TabIndex:   1,
		URL:        "https://example.com/one",
		Title:      "One",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
	}, true)
	second := sessionRegistry.TrackTab(sessionID, BrowserSessionTarget{
		ID:         "tab-2",
		TabIndex:   2,
		URL:        "https://example.com/two",
		Title:      "Two",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
	}, false)
	if first.ID == "" || second.ID == "" {
		t.Fatalf("expected tracked targets to expose ids, got first=%#v second=%#v", first, second)
	}

	manager := SharedSessionBrowserObserverManagerFor(sessionRegistry, nil, nil, time.Minute)
	initialGeneration := manager.currentGeneration()

	result, err := manager.SelectTarget(SharedSessionBrowserSelectTargetRequest{
		SessionID: sessionID,
		Route:     route,
		TargetID:  second.ID,
		Source:    "select_target",
		Actor:     "test select target",
	})
	if err != nil {
		t.Fatalf("expected target selection to succeed, got %v", err)
	}
	if !result.Ready || result.Selection == nil || result.Selection.ID != second.ID {
		t.Fatalf("expected target selection to remember tab-2, got %#v", result)
	}
	if got := manager.currentGeneration(); got != initialGeneration+1 {
		t.Fatalf("expected provider-backed target selection to bump generation once from %d to %d, got %d", initialGeneration, initialGeneration+1, got)
	}
}

func TestSharedSessionBrowserObserverManagerSelectProfileBumpsGenerationOnceWhenProviderBacked(t *testing.T) {
	stateRegistry := NewBrowserSessionStateRegistry()
	manager := SharedSessionBrowserObserverManagerFor(nil, nil, stateRegistry, time.Minute)
	initialGeneration := manager.currentGeneration()

	selection, decision, ok, err := manager.SelectProfile(
		context.Background(),
		"sess-manager-select-profile-generation",
		BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		"Chromium",
		nil,
		false,
		"select_profile",
	)
	if err != nil {
		t.Fatalf("expected profile selection to succeed, got %v", err)
	}
	if !ok {
		t.Fatalf("expected profile selection to persist a remembered selection, got ok=false decision=%q selection=%#v", decision, selection)
	}
	if got := manager.currentGeneration(); got != initialGeneration+1 {
		t.Fatalf("expected provider-backed profile selection to bump generation once from %d to %d, got %d", initialGeneration, initialGeneration+1, got)
	}
}
