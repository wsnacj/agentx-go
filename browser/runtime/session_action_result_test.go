package browserruntime

import (
	"context"
	"testing"
	"time"
)

func TestApplySharedSessionBrowserActionResultEventSeedsSiblingProviderAndCarriesReviewPosture(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	runRegistryA := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-apply-action-result-event-sibling-provider": {{RunID: "run-a", Status: "running"}},
		}},
	}
	runRegistryB := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-apply-action-result-event-sibling-provider": {{RunID: "run-b", Status: "running"}},
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
	sessionID := "sess-apply-action-result-event-sibling-provider"
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

	result := ApplySharedSessionBrowserActionResultEvent(
		sessionRegistry,
		runRegistryA,
		nil,
		SharedSessionBrowserActionResultEventRequest{
			SessionID:      sessionID,
			Route:          route,
			TabIndex:       2,
			URL:            "https://example.com/archive.zip",
			Title:          "Archive",
			Source:         "browser_act_download",
			SetCurrent:     true,
			ReviewDecision: "download_review_confirmed",
			ReviewReady:    true,
			Note:           "browser_act file download review acknowledged via force=true",
		},
		time.Minute,
	)
	if result.TargetID == "" || result.ReviewDecision != "download_review_confirmed" || !result.ReviewReady {
		t.Fatalf("expected action result event to track target with review posture, got %#v", result)
	}

	seededB := boundB.ObserveWatchLoop(context.Background(), req)
	if len(seededB.View.Session.Routes) != 1 || seededB.View.Session.Routes[0].CurrentTargetID != result.TargetID {
		t.Fatalf("expected sibling watch loop to reuse action result source, got %#v", seededB.View.Session.Routes)
	}
	if len(backend.statusReqs) != 2 || len(backend.profilesReqs) != 2 {
		t.Fatalf("expected sibling watch loop to reuse action result source without extra polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}

func TestApplySharedSessionBrowserActionResultWithContextPreservesReviewPostureAndSeedsSiblingProvider(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	runRegistryA := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-apply-action-result-helper-sibling-provider": {{RunID: "run-a", Status: "running"}},
		}},
	}
	runRegistryB := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-apply-action-result-helper-sibling-provider": {{RunID: "run-b", Status: "running"}},
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
	sessionID := "sess-apply-action-result-helper-sibling-provider"
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

	result := ApplySharedSessionBrowserActionResultWithContext(
		mutationCtx,
		SharedSessionBrowserActionResultEventRequest{
			SessionID:      sessionID,
			Route:          route,
			TabIndex:       2,
			URL:            "https://example.com/report.pdf",
			Title:          "Report",
			Source:         "browser_act_save_pdf",
			SetCurrent:     true,
			ReviewDecision: "save_pdf_review_confirmed",
			ReviewReady:    true,
			Note:           "browser_act page export review acknowledged via force=true",
		},
	)
	if result.TargetID == "" || result.ReviewDecision != "save_pdf_review_confirmed" || !result.ReviewReady {
		t.Fatalf("expected helper result to preserve review posture, got %#v", result)
	}

	seededB := boundB.ObserveWatchLoop(context.Background(), req)
	if len(seededB.View.Session.Routes) != 1 || seededB.View.Session.Routes[0].CurrentTargetID != result.TargetID {
		t.Fatalf("expected sibling watch loop to reuse helper action result source, got %#v", seededB.View.Session.Routes)
	}
	if len(backend.statusReqs) != 2 || len(backend.profilesReqs) != 2 {
		t.Fatalf("expected helper action result to avoid extra polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}

func TestSharedSessionBrowserWatchManagerObserveWatchLoopUsesRawTargetSourceWhenCachesDrained(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	runRegistry := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-watch-loop-raw-target-source": {{RunID: "run-1", Status: "running"}},
		}},
	}
	manager := NewSharedSessionBrowserObserverManager(sessionRegistry, runRegistry, stateRegistry, time.Minute)
	observedAt := time.Now().Add(-1500 * time.Millisecond)
	backend := &statusProfilesObservationTestBackend{
		rawTarget: func(_ context.Context, requestedProfile string) SharedSessionBrowserRawTargetObservation {
			if requestedProfile != "isolated" {
				t.Fatalf("expected trimmed requested profile, got %q", requestedProfile)
			}
			return SharedSessionBrowserRawTargetObservation{
				RequestedProfile: "isolated",
				TabIndex:         2,
				SetCurrent:       true,
				URL:              "https://example.com/console",
				Title:            "Console",
				Source:           "runtime_console_source",
				ObservedAt:       observedAt,
			}
		},
	}
	bound := manager.Bind(backend)
	sessionID := "sess-watch-loop-raw-target-source"
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
		t.Fatalf("expected raw target source to track current target without polling, got %#v", seeded)
	}
	if seeded.Targets[0].TabIndex != 2 || seeded.Targets[0].URL != "https://example.com/console" || seeded.Targets[0].Title != "Console" {
		t.Fatalf("expected raw target source to seed tracked target, got %#v", seeded.Targets)
	}
	if len(backend.statusReqs) != 0 || len(backend.profilesReqs) != 0 {
		t.Fatalf("expected raw target source to avoid RuntimeStatus/RuntimeProfiles polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
	if backend.rawTargetCalls != 1 {
		t.Fatalf("expected raw target source to be consumed once, got %d", backend.rawTargetCalls)
	}
}

func TestSharedSessionBrowserWatchManagerObserveWatchLoopUsesSetCurrentOnlyRawTargetSourceWhenCachesDrained(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	runRegistry := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-watch-loop-raw-target-set-current-source": {{RunID: "run-1", Status: "running"}},
		}},
	}
	manager := NewSharedSessionBrowserObserverManager(sessionRegistry, runRegistry, stateRegistry, time.Minute)
	observedAt := time.Now().Add(-1500 * time.Millisecond)
	backend := &statusProfilesObservationTestBackend{
		rawTarget: func(_ context.Context, requestedProfile string) SharedSessionBrowserRawTargetObservation {
			if requestedProfile != "isolated" {
				t.Fatalf("expected trimmed requested profile, got %q", requestedProfile)
			}
			return SharedSessionBrowserRawTargetObservation{
				RequestedProfile: "isolated",
				SetCurrent:       true,
				URL:              "https://example.com/archive.zip",
				Title:            "archive.zip",
				Source:           "runtime_wait_download_source",
				ObservedAt:       observedAt,
			}
		},
	}
	bound := manager.Bind(backend)
	sessionID := "sess-watch-loop-raw-target-set-current-source"
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
		t.Fatalf("expected setCurrent-only raw target source to track current target without polling, got %#v", seeded)
	}
	if seeded.Targets[0].TabIndex != 0 || seeded.Targets[0].URL != "https://example.com/archive.zip" || seeded.Targets[0].Title != "archive.zip" {
		t.Fatalf("expected setCurrent-only raw target source to seed tracked target, got %#v", seeded.Targets)
	}
	if len(backend.statusReqs) != 0 || len(backend.profilesReqs) != 0 {
		t.Fatalf("expected setCurrent-only raw target source to avoid RuntimeStatus/RuntimeProfiles polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
	if backend.rawTargetCalls != 1 {
		t.Fatalf("expected setCurrent-only raw target source to be consumed once, got %d", backend.rawTargetCalls)
	}
}

func TestSharedSessionBrowserWatchManagerObserveWatchLoopUsesRawPageActionTargetSourceWhenCachesDrained(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	runRegistry := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-watch-loop-raw-page-action-target-source": {{RunID: "run-1", Status: "running"}},
		}},
	}
	manager := NewSharedSessionBrowserObserverManager(sessionRegistry, runRegistry, stateRegistry, time.Minute)
	observedAt := time.Now().Add(-2 * time.Second)
	backend := &statusProfilesObservationTestBackend{
		rawTarget: func(_ context.Context, requestedProfile string) SharedSessionBrowserRawTargetObservation {
			if requestedProfile != "isolated" {
				t.Fatalf("expected trimmed requested profile, got %q", requestedProfile)
			}
			return SharedSessionBrowserRawTargetObservation{
				RequestedProfile:  "isolated",
				TabIndex:          2,
				SetCurrent:        true,
				URL:               "https://example.com/popup",
				Title:             "Popup",
				Source:            "runtime_click_source",
				PreferredTargetID: "tab-popup",
				Actor:             "browser_click",
				Force:             true,
				Review: SharedSessionBrowserPendingTargetReviewState{
					Review: &BrowserSessionTargetReview{
						ID:       "tab-popup",
						TabIndex: 2,
						URL:      "https://example.com/popup",
						Title:    "Popup",
						Decision: "session_target_popup_review_required",
						Reason:   "popup review required",
					},
					Count:        1,
					PolicyState:  "review_required",
					PolicyReason: "popup review required",
				},
				ObservedAt: observedAt,
			}
		},
	}
	bound := manager.Bind(backend)
	sessionID := "sess-watch-loop-raw-page-action-target-source"
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
		t.Fatalf("expected raw page-action target source to track current target without polling, got %#v", seeded)
	}
	if seeded.Targets[0].TabIndex != 2 || seeded.Targets[0].URL != "https://example.com/popup" || seeded.Targets[0].Title != "Popup" {
		t.Fatalf("expected raw page-action target source to seed tracked target, got %#v", seeded.Targets)
	}
	if len(backend.statusReqs) != 0 || len(backend.profilesReqs) != 0 {
		t.Fatalf("expected raw page-action target source to avoid RuntimeStatus/RuntimeProfiles polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
	if backend.rawTargetCalls != 1 {
		t.Fatalf("expected raw page-action target source to be consumed once, got %d", backend.rawTargetCalls)
	}
}

func TestSharedSessionBrowserWatchManagerObserveWatchLoopUsesRawGenericActionTargetSourceWhenCachesDrained(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	runRegistry := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-watch-loop-raw-generic-action-target-source": {{RunID: "run-1", Status: "running"}},
		}},
	}
	manager := NewSharedSessionBrowserObserverManager(sessionRegistry, runRegistry, stateRegistry, time.Minute)
	observedAt := time.Now().Add(-2 * time.Second)
	backend := &statusProfilesObservationTestBackend{
		rawTarget: func(_ context.Context, requestedProfile string) SharedSessionBrowserRawTargetObservation {
			if requestedProfile != "isolated" {
				t.Fatalf("expected trimmed requested profile, got %q", requestedProfile)
			}
			return SharedSessionBrowserRawTargetObservation{
				RequestedProfile:  "isolated",
				SetCurrent:        true,
				URL:               "https://example.com/archive.zip",
				Title:             "archive.zip",
				Source:            "runtime_wait_download_source",
				PreferredTargetID: "download-target",
				ReviewDecision:    "session_target_download_review_confirmed",
				ReviewReady:       true,
				Note:              "waited for download",
				ObservedAt:        observedAt,
			}
		},
	}
	bound := manager.Bind(backend)
	sessionID := "sess-watch-loop-raw-generic-action-target-source"
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
		t.Fatalf("expected raw generic-action target source to track current target without polling, got %#v", seeded)
	}
	if seeded.Targets[0].TabIndex != 0 || seeded.Targets[0].URL != "https://example.com/archive.zip" || seeded.Targets[0].Title != "archive.zip" {
		t.Fatalf("expected raw generic-action target source to seed tracked target, got %#v", seeded.Targets)
	}
	if len(backend.statusReqs) != 0 || len(backend.profilesReqs) != 0 {
		t.Fatalf("expected raw generic-action target source to avoid RuntimeStatus/RuntimeProfiles polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
	if backend.rawTargetCalls != 1 {
		t.Fatalf("expected raw generic-action target source to be consumed once, got %d", backend.rawTargetCalls)
	}
}

func TestSharedSessionBrowserWatchManagerObserveWatchLoopUsesBackendRawRouteMutationSourceAfterRawGenericActionTargetConsumed(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	runRegistry := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-watch-loop-backend-raw-route-mutation-generic-action-target": {{RunID: "run-1", Status: "running"}},
		}},
	}
	manager := NewSharedSessionBrowserObserverManager(sessionRegistry, runRegistry, stateRegistry, time.Minute)
	observedAt := time.Now().Add(-2 * time.Second)
	rawTargetDelivered := false
	rawRouteMutationDelivered := false
	backend := &statusProfilesObservationTestBackend{
		rawTarget: func(_ context.Context, requestedProfile string) SharedSessionBrowserRawTargetObservation {
			if requestedProfile != "isolated" {
				t.Fatalf("expected trimmed requested profile, got %q", requestedProfile)
			}
			if rawTargetDelivered {
				return SharedSessionBrowserRawTargetObservation{}
			}
			rawTargetDelivered = true
			return SharedSessionBrowserRawTargetObservation{
				RequestedProfile:  "isolated",
				SetCurrent:        true,
				URL:               "https://example.com/archive.zip",
				Title:             "archive.zip",
				Source:            "runtime_wait_download_source",
				PreferredTargetID: "download-target",
				ReviewDecision:    "session_target_download_review_confirmed",
				ReviewReady:       true,
				Note:              "waited for download",
				ObservedAt:        observedAt,
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
				Kind:             "action_result",
				TargetID:         "download-target",
				URL:              "https://example.com/archive.zip",
				Title:            "archive.zip",
				SetCurrent:       true,
				Decision:         "session_target_download_review_confirmed",
				Ready:            true,
				Note:             "waited for download",
				Source:           "runtime_wait_download_source",
				ObservedAt:       observedAt,
			}
		},
	}
	bound := manager.Bind(backend)
	sessionID := "sess-watch-loop-backend-raw-route-mutation-generic-action-target"
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

	first := bound.ObserveWatchLoop(context.Background(), req)
	if len(first.View.Session.Routes) != 1 || first.View.Session.Routes[0].CurrentTargetID == "" {
		t.Fatalf("expected raw generic-action target source to restore current target, got %#v", first.View.Session.Routes)
	}
	if backend.rawTargetCalls != 1 || backend.rawRouteMutationCalls != 0 {
		t.Fatalf("expected first observe to consume only raw target source, got rawTarget=%d rawRouteMutation=%d", backend.rawTargetCalls, backend.rawRouteMutationCalls)
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
		t.Fatalf("expected backend raw route-mutation source to restore current target after raw generic-action target drain, got %#v", second.View.Session.Routes)
	}
	if backend.rawTargetCalls != 2 || backend.rawRouteMutationCalls != 1 {
		t.Fatalf("expected second observe to fall back to backend raw route-mutation after raw generic-action target drain, got rawTarget=%d rawRouteMutation=%d", backend.rawTargetCalls, backend.rawRouteMutationCalls)
	}
	if len(backend.statusReqs) != 0 || len(backend.profilesReqs) != 0 {
		t.Fatalf("expected raw target/raw route-mutation sources to avoid RuntimeStatus/RuntimeProfiles polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}
