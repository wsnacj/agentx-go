package browserruntime

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestApplySharedSessionBrowserNavigationResultEventRecordsRedirectReviewAndRestoresSiblingProviderFromPrimaryCachedEventCycle(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	runRegistryA := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-apply-navigation-result-event-review-sibling-provider": {{RunID: "run-a", Status: "running"}},
		}},
	}
	runRegistryB := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-apply-navigation-result-event-review-sibling-provider": {{RunID: "run-b", Status: "running"}},
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
	sessionID := "sess-apply-navigation-result-event-review-sibling-provider"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}
	home := sessionRegistry.TrackTab(sessionID, BrowserSessionTarget{
		ID:         "tab-home",
		TabIndex:   1,
		URL:        "https://example.com/home",
		Title:      "Home",
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

	firstA := boundA.ObserveWatchLoop(context.Background(), req)
	firstB := boundB.ObserveWatchLoop(context.Background(), req)
	if len(firstA.View.Session.Routes) != 1 || firstA.View.Session.Routes[0].CurrentTargetID != home.ID {
		t.Fatalf("expected first provider to expose home target before navigation event, got %#v", firstA.View.Session.Routes)
	}
	if len(firstB.View.Session.Routes) != 1 || firstB.View.Session.Routes[0].CurrentTargetID != home.ID {
		t.Fatalf("expected second provider to expose home target before navigation event, got %#v", firstB.View.Session.Routes)
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

	result := ApplySharedSessionBrowserNavigationResultEvent(
		sessionRegistry,
		runRegistryA,
		nil,
		SharedSessionBrowserNavigationResultEventRequest{
			SessionID:      sessionID,
			Route:          route,
			RequestedURL:   "https://example.com/start",
			FinalURL:       "https://example.org/landing",
			Title:          "Landing",
			Source:         "browser_navigate",
			PriorSelection: SnapshotSharedSessionBrowserCurrentTargetSelection(sessionRegistry, sessionID, route),
		},
		time.Minute,
	)
	if result.TargetID == "" || result.Review == nil || !result.ReviewRequired || result.ReviewDecision != "navigate_redirect_review_required" || result.ReviewReady {
		t.Fatalf("expected navigation result event to require redirect review, got %#v", result)
	}
	if !strings.Contains(result.Note, "redirected across origin") {
		t.Fatalf("expected redirect review note, got %#v", result)
	}

	seededB := boundB.ObserveWatchLoop(context.Background(), req)
	if len(seededB.View.Session.Routes) != 1 {
		t.Fatalf("expected sibling watch loop to preserve refreshed route snapshot, got %#v", seededB.View.Session.Routes)
	}
	if seededB.View.Session.Routes[0].CurrentTargetID != home.ID {
		t.Fatalf("expected sibling watch loop to reuse restored prior selection, got %#v", seededB.View.Session.Routes)
	}
	if seededB.View.Session.Routes[0].PendingTargetReview == nil || seededB.View.Session.Routes[0].PendingTargetReview.ID != result.TargetID {
		t.Fatalf("expected sibling watch loop to expose pending redirect review, got %#v", seededB.View.Session.Routes)
	}
	if len(backend.statusReqs) != 2 || len(backend.profilesReqs) != 2 {
		t.Fatalf("expected sibling watch loop to reuse navigation result source without extra polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}

func TestApplySharedSessionBrowserNavigationResultWithContextSeedsSiblingProviderFromPrimaryCachedEventCycle(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	runRegistryA := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-apply-navigation-result-helper-sibling-provider": {{RunID: "run-a", Status: "running"}},
		}},
	}
	runRegistryB := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-apply-navigation-result-helper-sibling-provider": {{RunID: "run-b", Status: "running"}},
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
	sessionID := "sess-apply-navigation-result-helper-sibling-provider"
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

	result := ApplySharedSessionBrowserNavigationResultWithContext(
		mutationCtx,
		SharedSessionBrowserNavigationResultEventRequest{
			SessionID:    sessionID,
			Route:        route,
			TabIndex:     2,
			RequestedURL: "https://example.com/landing",
			FinalURL:     "https://example.com/landing",
			Title:        "Landing",
			Source:       "browser_act_navigate",
		},
	)
	if result.TargetID == "" || result.ReviewRequired || result.ReviewDecision != "" {
		t.Fatalf("expected navigation helper to track target without redirect review, got %#v", result)
	}

	seededB := boundB.ObserveWatchLoop(context.Background(), req)
	if len(seededB.View.Session.Routes) != 1 || seededB.View.Session.Routes[0].CurrentTargetID != result.TargetID {
		t.Fatalf("expected sibling watch loop to reuse navigation helper source, got %#v", seededB.View.Session.Routes)
	}
	if len(seededB.View.Session.Routes[0].Targets) != 1 || seededB.View.Session.Routes[0].Targets[0].TabIndex != 2 {
		t.Fatalf("expected sibling watch loop to expose navigation helper tracked tab, got %#v", seededB.View.Session.Routes)
	}
	if len(backend.statusReqs) != 2 || len(backend.profilesReqs) != 2 {
		t.Fatalf("expected helper-seeded sibling watch loop to avoid extra polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}

func TestSharedSessionBrowserWatchManagerObserveWatchLoopUsesRawNavigationSourceWhenCachesDrained(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	runRegistry := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-watch-loop-raw-navigation-source": {{RunID: "run-1", Status: "running"}},
		}},
	}
	manager := NewSharedSessionBrowserObserverManager(sessionRegistry, runRegistry, stateRegistry, time.Minute)
	observedAt := time.Now().Add(-2 * time.Second)
	backend := &statusProfilesObservationTestBackend{
		rawNavigation: func(_ context.Context, requestedProfile string) SharedSessionBrowserRawNavigationObservation {
			if requestedProfile != "isolated" {
				t.Fatalf("expected trimmed requested profile, got %q", requestedProfile)
			}
			return SharedSessionBrowserRawNavigationObservation{
				RequestedProfile: "isolated",
				RequestedURL:     "https://example.com/start",
				FinalURL:         "https://example.org/landing",
				Title:            "Landing",
				TabIndex:         2,
				ObservedAt:       observedAt,
			}
		},
	}
	bound := manager.Bind(backend)
	sessionID := "sess-watch-loop-raw-navigation-source"
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
	home := sessionRegistry.TrackTab(sessionID, BrowserSessionTarget{
		ID:         "tab-home",
		TabIndex:   1,
		URL:        "https://example.com/home",
		Title:      "Home",
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

	loop := bound.ObserveWatchLoop(context.Background(), req)
	if len(loop.View.Session.Routes) != 1 {
		t.Fatalf("expected watch loop to expose one route, got %#v", loop.View.Session.Routes)
	}
	seeded := loop.View.Session.Routes[0]
	if seeded.CurrentTargetID != home.ID {
		t.Fatalf("expected raw navigation source to restore prior current target, got %#v", seeded)
	}
	if seeded.PendingTargetReview == nil || seeded.PendingTargetReview.ID == "" || seeded.PendingTargetReview.URL != "https://example.org/landing" {
		t.Fatalf("expected raw navigation source to record pending redirect review, got %#v", seeded)
	}
	if len(seeded.Targets) != 2 || seeded.Targets[1].TabIndex != 2 || seeded.Targets[1].URL != "https://example.org/landing" {
		t.Fatalf("expected raw navigation source to track landing tab without polling, got %#v", seeded.Targets)
	}
	if len(backend.statusReqs) != 0 || len(backend.profilesReqs) != 0 {
		t.Fatalf("expected raw navigation source to avoid RuntimeStatus/RuntimeProfiles polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
	if backend.rawNavigationCalls != 1 {
		t.Fatalf("expected raw navigation source to be consumed once, got %d", backend.rawNavigationCalls)
	}
}

func TestSharedSessionBrowserWatchManagerObserveWatchLoopUsesForcedRawNavigationSourceWhenCachesDrained(t *testing.T) {
	observedAt := time.Now().Add(-2 * time.Second)
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	runRegistry := &testSharedSessionRunRegistry{
		items: map[string][]SharedSessionRunInfo{
			"sess-watch-loop-raw-navigation-force-source": {{RunID: "run-force", Status: "running"}},
		},
	}
	manager := SharedSessionBrowserObserverManagerFor(sessionRegistry, runRegistry, stateRegistry, time.Minute)
	backend := &statusProfilesObservationTestBackend{
		rawNavigation: func(_ context.Context, requestedProfile string) SharedSessionBrowserRawNavigationObservation {
			if requestedProfile != "isolated" {
				t.Fatalf("expected trimmed requested profile, got %q", requestedProfile)
			}
			return SharedSessionBrowserRawNavigationObservation{
				RequestedProfile: "isolated",
				RequestedURL:     "https://example.com/start",
				FinalURL:         "https://example.org/landing",
				Title:            "Landing",
				TabIndex:         2,
				Force:            true,
				PriorSelection:   &BrowserSessionTargetSelection{ID: "tab-home", Source: "tracked_current_target"},
				Note:             "redirect forced",
				ObservedAt:       observedAt,
			}
		},
	}
	bound := manager.Bind(backend)
	sessionID := "sess-watch-loop-raw-navigation-force-source"
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
	home := sessionRegistry.TrackTab(sessionID, BrowserSessionTarget{
		ID:         "tab-home",
		TabIndex:   1,
		URL:        "https://example.com/home",
		Title:      "Home",
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

	loop := bound.ObserveWatchLoop(context.Background(), req)
	if len(loop.View.Session.Routes) != 1 {
		t.Fatalf("expected watch loop to expose one route, got %#v", loop.View.Session.Routes)
	}
	seeded := loop.View.Session.Routes[0]
	if seeded.PendingTargetReview != nil {
		t.Fatalf("expected forced raw navigation source to skip pending redirect review, got %#v", seeded)
	}
	if seeded.CurrentTargetID == "" || seeded.CurrentTargetID == home.ID {
		t.Fatalf("expected forced raw navigation source to keep redirected target current, got %#v", seeded)
	}
	if len(seeded.Targets) != 2 || seeded.Targets[1].TabIndex != 2 || seeded.Targets[1].URL != "https://example.org/landing" {
		t.Fatalf("expected forced raw navigation source to track landing tab without polling, got %#v", seeded.Targets)
	}
	if len(backend.statusReqs) != 0 || len(backend.profilesReqs) != 0 {
		t.Fatalf("expected forced raw navigation source to avoid RuntimeStatus/RuntimeProfiles polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
	if backend.rawNavigationCalls != 1 {
		t.Fatalf("expected raw navigation source to be consumed once, got %d", backend.rawNavigationCalls)
	}
}

func TestSharedSessionBrowserWatchManagerObserveWatchLoopUsesBackendRawRouteMutationSourceAfterForcedRawNavigationConsumed(t *testing.T) {
	observedAt := time.Now().Add(-2 * time.Second)
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	runRegistry := &testSharedSessionRunRegistry{
		items: map[string][]SharedSessionRunInfo{
			"sess-watch-loop-backend-raw-route-mutation-navigation-force": {{RunID: "run-force", Status: "running"}},
		},
	}
	manager := SharedSessionBrowserObserverManagerFor(sessionRegistry, runRegistry, stateRegistry, time.Minute)
	rawNavigationDelivered := false
	rawRouteMutationDelivered := false
	backend := &statusProfilesObservationTestBackend{
		rawNavigation: func(_ context.Context, requestedProfile string) SharedSessionBrowserRawNavigationObservation {
			if requestedProfile != "isolated" {
				t.Fatalf("expected trimmed requested profile, got %q", requestedProfile)
			}
			if rawNavigationDelivered {
				return SharedSessionBrowserRawNavigationObservation{}
			}
			rawNavigationDelivered = true
			return SharedSessionBrowserRawNavigationObservation{
				RequestedProfile: "isolated",
				RequestedURL:     "https://example.com/start",
				FinalURL:         "https://example.org/landing",
				Title:            "Landing",
				TabIndex:         2,
				Force:            true,
				ExplicitTargetID: "tab-start",
				PriorSelection:   &BrowserSessionTargetSelection{ID: "tab-home", Source: "tracked_current_target"},
				Note:             "redirect forced",
				ObservedAt:       observedAt,
			}
		},
		rawRouteMutation: func(_ context.Context, requestedProfile string) SharedSessionBrowserRawRouteMutationObservation {
			if requestedProfile != "isolated" {
				t.Fatalf("expected trimmed requested profile, got %q", requestedProfile)
			}
			if rawRouteMutationDelivered {
				return SharedSessionBrowserRawRouteMutationObservation{}
			}
			rawRouteMutationDelivered = true
			return SharedSessionBrowserRawRouteMutationObservation{
				RequestedProfile: requestedProfile,
				Route:            BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"},
				Kind:             "navigation_result",
				RequestedURL:     "https://example.com/start",
				TargetID:         "tab-start",
				TabIndex:         2,
				URL:              "https://example.org/landing",
				FinalURL:         "https://example.org/landing",
				Title:            "Landing",
				Force:            true,
				PriorSelection:   &BrowserSessionTargetSelection{ID: "tab-home", Source: "tracked_current_target"},
				Note:             "redirect forced",
				Source:           "runtime_navigation_source",
				ObservedAt:       observedAt,
			}
		},
	}
	bound := manager.Bind(backend)
	sessionID := "sess-watch-loop-backend-raw-route-mutation-navigation-force"
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
	home := sessionRegistry.TrackTab(sessionID, BrowserSessionTarget{
		ID:         "tab-home",
		TabIndex:   1,
		URL:        "https://example.com/home",
		Title:      "Home",
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
	if len(first.View.Session.Routes) != 1 {
		t.Fatalf("expected first watch loop to expose one route, got %#v", first.View.Session.Routes)
	}
	if first.View.Session.Routes[0].PendingTargetReview != nil {
		t.Fatalf("expected forced raw navigation source to skip pending review, got %#v", first.View.Session.Routes)
	}
	if first.View.Session.Routes[0].CurrentTargetID == "" || first.View.Session.Routes[0].CurrentTargetID == home.ID {
		t.Fatalf("expected forced raw navigation source to keep redirected target current, got %#v", first.View.Session.Routes)
	}
	if backend.rawNavigationCalls != 1 || backend.rawRouteMutationCalls != 0 {
		t.Fatalf("expected first watch loop to consume only raw navigation source, got rawNavigation=%d rawRouteMutation=%d", backend.rawNavigationCalls, backend.rawRouteMutationCalls)
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
	if second.View.Session.Routes[0].PendingTargetReview != nil {
		t.Fatalf("expected backend raw route-mutation source to keep forced redirect review cleared, got %#v", second.View.Session.Routes)
	}
	if second.View.Session.Routes[0].CurrentTargetID == "" || second.View.Session.Routes[0].CurrentTargetID == home.ID {
		t.Fatalf("expected backend raw route-mutation source to restore redirected current target, got %#v", second.View.Session.Routes)
	}
	if backend.rawNavigationCalls != 2 || backend.rawRouteMutationCalls != 1 {
		t.Fatalf("expected second watch loop to fall back to backend raw route-mutation after raw navigation drain, got rawNavigation=%d rawRouteMutation=%d", backend.rawNavigationCalls, backend.rawRouteMutationCalls)
	}
	if len(backend.statusReqs) != 0 || len(backend.profilesReqs) != 0 {
		t.Fatalf("expected raw navigation/raw route-mutation sources to avoid RuntimeStatus/RuntimeProfiles polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}

func TestSharedSessionBrowserWatchManagerObserveWatchLoopUsesBackendRawRouteMutationSourceAfterRawNavigationRedirectConsumed(t *testing.T) {
	observedAt := time.Now().Add(-2 * time.Second)
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	runRegistry := &testSharedSessionRunRegistry{
		items: map[string][]SharedSessionRunInfo{
			"sess-watch-loop-backend-raw-route-mutation-navigation-review": {{RunID: "run-review", Status: "running"}},
		},
	}
	manager := SharedSessionBrowserObserverManagerFor(sessionRegistry, runRegistry, stateRegistry, time.Minute)
	rawNavigationDelivered := false
	rawRouteMutationDelivered := false
	backend := &statusProfilesObservationTestBackend{
		rawNavigation: func(_ context.Context, requestedProfile string) SharedSessionBrowserRawNavigationObservation {
			if requestedProfile != "isolated" {
				t.Fatalf("expected trimmed requested profile, got %q", requestedProfile)
			}
			if rawNavigationDelivered {
				return SharedSessionBrowserRawNavigationObservation{}
			}
			rawNavigationDelivered = true
			return SharedSessionBrowserRawNavigationObservation{
				RequestedProfile: "isolated",
				RequestedURL:     "https://example.com/start",
				FinalURL:         "https://example.org/landing",
				Title:            "Landing",
				TabIndex:         2,
				PriorSelection:   &BrowserSessionTargetSelection{ID: "tab-home", Source: "tracked_current_target"},
				ObservedAt:       observedAt,
			}
		},
		rawRouteMutation: func(_ context.Context, requestedProfile string) SharedSessionBrowserRawRouteMutationObservation {
			if requestedProfile != "isolated" {
				t.Fatalf("expected trimmed requested profile, got %q", requestedProfile)
			}
			if rawRouteMutationDelivered {
				return SharedSessionBrowserRawRouteMutationObservation{}
			}
			rawRouteMutationDelivered = true
			return SharedSessionBrowserRawRouteMutationObservation{
				RequestedProfile: requestedProfile,
				Route:            BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"},
				Kind:             "navigation_result",
				RequestedURL:     "https://example.com/start",
				TabIndex:         2,
				URL:              "https://example.org/landing",
				FinalURL:         "https://example.org/landing",
				Title:            "Landing",
				PriorSelection:   &BrowserSessionTargetSelection{ID: "tab-home", Source: "tracked_current_target"},
				Source:           "runtime_navigation_source",
				ObservedAt:       observedAt,
			}
		},
	}
	bound := manager.Bind(backend)
	sessionID := "sess-watch-loop-backend-raw-route-mutation-navigation-review"
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
	sessionRegistry.TrackTab(sessionID, BrowserSessionTarget{
		ID:         "tab-home",
		TabIndex:   1,
		URL:        "https://example.com/home",
		Title:      "Home",
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
	if len(first.View.Session.Routes) != 1 {
		t.Fatalf("expected first watch loop to expose one route, got %#v", first.View.Session.Routes)
	}
	if first.View.Session.Routes[0].PendingTargetReview == nil || first.View.Session.Routes[0].PendingTargetReview.URL != "https://example.org/landing" {
		t.Fatalf("expected raw navigation source to expose pending redirect review, got %#v", first.View.Session.Routes)
	}
	if backend.rawNavigationCalls != 1 || backend.rawRouteMutationCalls != 0 {
		t.Fatalf("expected first watch loop to consume only raw navigation source, got rawNavigation=%d rawRouteMutation=%d", backend.rawNavigationCalls, backend.rawRouteMutationCalls)
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
	if second.View.Session.Routes[0].PendingTargetReview == nil || second.View.Session.Routes[0].PendingTargetReview.URL != "https://example.org/landing" {
		t.Fatalf("expected backend raw route-mutation source to restore pending redirect review, got %#v", second.View.Session.Routes)
	}
	if backend.rawNavigationCalls != 2 || backend.rawRouteMutationCalls != 1 {
		t.Fatalf("expected second watch loop to fall back to backend raw route-mutation after raw navigation drain, got rawNavigation=%d rawRouteMutation=%d", backend.rawNavigationCalls, backend.rawRouteMutationCalls)
	}
	if len(backend.statusReqs) != 0 || len(backend.profilesReqs) != 0 {
		t.Fatalf("expected raw navigation/raw route-mutation sources to avoid RuntimeStatus/RuntimeProfiles polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}
