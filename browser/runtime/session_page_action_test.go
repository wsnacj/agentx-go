package browserruntime

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestApplySharedSessionBrowserPageActionResultEventSeedsSiblingProviderFromPrimaryCachedEventCycle(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	runRegistryA := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-apply-page-action-result-event-sibling-provider": {{RunID: "run-a", Status: "running"}},
		}},
	}
	runRegistryB := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-apply-page-action-result-event-sibling-provider": {{RunID: "run-b", Status: "running"}},
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
	sessionID := "sess-apply-page-action-result-event-sibling-provider"
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

	result := ApplySharedSessionBrowserPageActionResultEvent(
		sessionRegistry,
		runRegistryA,
		nil,
		SharedSessionBrowserPageActionResultEventRequest{
			SessionID: sessionID,
			Route:     route,
			TabIndex:  2,
			URL:       "https://example.com/extracted",
			Title:     "Extracted",
			Source:    "browser_extract",
			Actor:     "browser_extract",
		},
		time.Minute,
	)
	if result.TargetID == "" || result.ReviewDecision != "" || result.ReviewReady {
		t.Fatalf("expected page action event to track target without review posture, got %#v", result)
	}

	seededB := boundB.ObserveWatchLoop(context.Background(), req)
	if len(seededB.View.Session.Routes) != 1 || seededB.View.Session.Routes[0].CurrentTargetID != result.TargetID {
		t.Fatalf("expected sibling watch loop to reuse page action result source, got %#v", seededB.View.Session.Routes)
	}
	if len(backend.statusReqs) != 2 || len(backend.profilesReqs) != 2 {
		t.Fatalf("expected sibling watch loop to reuse page action result source without extra polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}

func TestApplySharedSessionBrowserPageActionResultWithContextPreservesReviewPostureAndSeedsSiblingProvider(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	runRegistryA := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-apply-page-action-result-helper-review": {{RunID: "run-a", Status: "running"}},
		}},
	}
	runRegistryB := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-apply-page-action-result-helper-review": {{RunID: "run-b", Status: "running"}},
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
	sessionID := "sess-apply-page-action-result-helper-review"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}
	popup := sessionRegistry.TrackTab(sessionID, BrowserSessionTarget{
		ID:         "target-popup",
		TabIndex:   2,
		URL:        "https://popup.example/offer",
		Title:      "Offer",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
	}, true)
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

	result := ApplySharedSessionBrowserPageActionResultWithContext(
		mutationCtx,
		SharedSessionBrowserPageActionResultEventRequest{
			SessionID:         sessionID,
			Route:             route,
			PreferredTargetID: popup.ID,
			TabIndex:          popup.TabIndex,
			URL:               popup.URL,
			Title:             popup.Title,
			Source:            "browser_act_extract",
			Actor:             "browser_act extract",
			Force:             true,
			Review: SharedSessionBrowserPendingTargetReviewState{
				Review: &BrowserSessionTargetReview{
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
				},
				Count:       1,
				PolicyState: "popup_review_required",
			},
		},
	)
	if result.TargetID != popup.ID || result.ReviewDecision != "session_target_popup_review_confirmed" || !result.ReviewReady {
		t.Fatalf("expected page action helper to preserve review target and confirm review posture, got %#v", result)
	}
	if !strings.Contains(result.Note, "force=true") {
		t.Fatalf("expected popup review note in page action helper result, got %#v", result)
	}

	seededB := boundB.ObserveWatchLoop(context.Background(), req)
	if len(seededB.View.Session.Routes) != 1 || seededB.View.Session.Routes[0].CurrentTargetID != popup.ID {
		t.Fatalf("expected sibling watch loop to reuse reviewed page action result source, got %#v", seededB.View.Session.Routes)
	}
	if len(backend.statusReqs) != 2 || len(backend.profilesReqs) != 2 {
		t.Fatalf("expected reviewed helper-seeded sibling watch loop to avoid extra polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}

func TestSharedSessionBrowserWatchManagerObserveWatchLoopUsesBackendRawRouteMutationSourceAfterRawPageActionTargetConsumed(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	runRegistry := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-watch-loop-backend-raw-route-mutation-page-action-target": {{RunID: "run-1", Status: "running"}},
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
				Kind:             "page_action_result",
				TargetID:         "tab-popup",
				TabIndex:         2,
				SetCurrent:       true,
				URL:              "https://example.com/popup",
				Title:            "Popup",
				Source:           "runtime_click_source",
				Actor:            "browser_click",
				Force:            true,
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
	sessionID := "sess-watch-loop-backend-raw-route-mutation-page-action-target"
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
		t.Fatalf("expected raw page-action target source to restore current target, got %#v", first.View.Session.Routes)
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
		t.Fatalf("expected backend raw route-mutation source to restore current target after raw page-action target drain, got %#v", second.View.Session.Routes)
	}
	if backend.rawTargetCalls != 2 || backend.rawRouteMutationCalls != 1 {
		t.Fatalf("expected second observe to fall back to backend raw route-mutation after raw page-action target drain, got rawTarget=%d rawRouteMutation=%d", backend.rawTargetCalls, backend.rawRouteMutationCalls)
	}
	if len(backend.statusReqs) != 0 || len(backend.profilesReqs) != 0 {
		t.Fatalf("expected raw target/raw route-mutation sources to avoid RuntimeStatus/RuntimeProfiles polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}
