package browserruntime

import (
	"context"
	"testing"
	"time"
)

func TestApplySharedSessionBrowserOpenResultEventSeedsSiblingProviderFromPrimaryCachedEventCycle(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	runRegistryA := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-apply-open-result-event-sibling-provider": {{RunID: "run-a", Status: "running"}},
		}},
	}
	runRegistryB := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-apply-open-result-event-sibling-provider": {{RunID: "run-b", Status: "running"}},
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
	sessionID := "sess-apply-open-result-event-sibling-provider"
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

	result := ApplySharedSessionBrowserOpenResultEvent(
		sessionRegistry,
		runRegistryA,
		nil,
		SharedSessionBrowserOpenResultEventRequest{
			SessionID: sessionID,
			Route:     route,
			URL:       "https://example.com/opened",
			Title:     "Opened",
			Source:    "browser_open",
		},
		time.Minute,
	)
	if result.TargetID == "" {
		t.Fatalf("expected open result event to track current target, got %#v", result)
	}

	seededB := boundB.ObserveWatchLoop(context.Background(), req)
	if len(seededB.View.Session.Routes) != 1 || seededB.View.Session.Routes[0].CurrentTargetID != result.TargetID {
		t.Fatalf("expected sibling watch loop to reuse open result current target, got %#v", seededB.View.Session.Routes)
	}
	if len(backend.statusReqs) != 2 || len(backend.profilesReqs) != 2 {
		t.Fatalf("expected sibling watch loop to reuse open result source without extra polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}

func TestApplySharedSessionBrowserOpenResultWithContextSeedsSiblingProviderFromPrimaryCachedEventCycle(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	runRegistryA := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-apply-open-result-helper-sibling-provider": {{RunID: "run-a", Status: "running"}},
		}},
	}
	runRegistryB := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-apply-open-result-helper-sibling-provider": {{RunID: "run-b", Status: "running"}},
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
	sessionID := "sess-apply-open-result-helper-sibling-provider"
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

	result := ApplySharedSessionBrowserOpenResultWithContext(
		mutationCtx,
		SharedSessionBrowserOpenResultEventRequest{
			SessionID: sessionID,
			Route:     route,
			URL:       "https://example.com/opened",
			Title:     "Opened",
			Source:    "browser_act_open",
		},
	)
	if result.TargetID == "" {
		t.Fatalf("expected open result helper to track current target, got %#v", result)
	}

	seededB := boundB.ObserveWatchLoop(context.Background(), req)
	if len(seededB.View.Session.Routes) != 1 || seededB.View.Session.Routes[0].CurrentTargetID != result.TargetID {
		t.Fatalf("expected sibling watch loop to reuse open helper source, got %#v", seededB.View.Session.Routes)
	}
	if len(backend.statusReqs) != 2 || len(backend.profilesReqs) != 2 {
		t.Fatalf("expected helper-seeded sibling watch loop to avoid extra polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}

func TestSharedSessionBrowserWatchManagerObserveWatchLoopUsesRawOpenSourceWhenCachesDrained(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	runRegistry := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-watch-loop-raw-open-source": {{RunID: "run-1", Status: "running"}},
		}},
	}
	manager := NewSharedSessionBrowserObserverManager(sessionRegistry, runRegistry, stateRegistry, time.Minute)
	observedAt := time.Now().Add(-1500 * time.Millisecond)
	backend := &statusProfilesObservationTestBackend{
		rawOpen: func(_ context.Context, requestedProfile string) SharedSessionBrowserRawOpenObservation {
			if requestedProfile != "isolated" {
				t.Fatalf("expected trimmed requested profile, got %q", requestedProfile)
			}
			return SharedSessionBrowserRawOpenObservation{
				RequestedProfile: "isolated",
				URL:              "https://example.com/opened",
				Title:            "Opened",
				ObservedAt:       observedAt,
			}
		},
	}
	bound := manager.Bind(backend)
	sessionID := "sess-watch-loop-raw-open-source"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}

	stateRegistry.RecordBrowserProfileState(sessionID, SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "running",
		Running:       true,
		Connected:     true,
		ObservedAt:    observedAt,
		StatusSince:   observedAt,
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

	loop := bound.ObserveWatchLoop(context.Background(), req)
	if len(loop.View.Session.Routes) != 1 {
		t.Fatalf("expected watch loop to expose one route, got %#v", loop.View.Session.Routes)
	}
	seeded := loop.View.Session.Routes[0]
	if seeded.CurrentTargetID == "" || len(seeded.Targets) != 1 {
		t.Fatalf("expected raw open source to track current target without polling, got %#v", seeded)
	}
	if seeded.Targets[0].URL != "https://example.com/opened" || seeded.Targets[0].Title != "Opened" {
		t.Fatalf("expected raw open source to seed opened target, got %#v", seeded.Targets)
	}
	if len(backend.statusReqs) != 0 || len(backend.profilesReqs) != 0 {
		t.Fatalf("expected raw open source to avoid RuntimeStatus/RuntimeProfiles polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
	if backend.rawOpenCalls != 1 {
		t.Fatalf("expected raw open source to be consumed once, got %d", backend.rawOpenCalls)
	}
}
