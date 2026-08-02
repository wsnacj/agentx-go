package browserruntime

import (
	"context"
	"testing"
	"time"
)

func TestExecuteSharedSessionBrowserClearProfileClearsMatchingTargetSelection(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	sessionID := "browser-runtime-clear-profile"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "workbench", Target: "node"}

	sessionRegistry.TrackCurrentTarget(sessionID, BrowserSessionTarget{
		ID:         "target-2",
		TabIndex:   2,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "select_profile")
	stateRegistry.SelectBrowserProfile(sessionID, SharedSessionBrowserProfileSelection{
		Backend:       "proxy",
		Profile:       "workbench",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Source:        "select_profile",
	})

	result := ExecuteSharedSessionBrowserClearProfile(BuildSharedSessionBrowserClearRequest(
		sessionRegistry,
		stateRegistry,
		sessionID,
		BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
		route,
		false,
		"",
		SharedSessionBrowserHealthInput{},
		time.Minute,
	))
	if result.Decision != "session_profile_cleared" || !result.Ready || !result.ClearedProfileSelection || !result.ClearedTargetSelection {
		t.Fatalf("expected clear_profile to clear profile+target selection, got %#v", result)
	}
	if selection := CurrentSharedSessionBrowserTargetSelection(sessionRegistry, sessionID, route); selection != nil {
		t.Fatalf("expected target selection to be cleared, got %#v", selection)
	}
	if _, ok := stateRegistry.SelectedBrowserProfile(sessionID, "node"); ok {
		t.Fatalf("expected profile selection to be cleared")
	}
}

func TestExecuteSharedSessionBrowserClearTargetClearsRememberedProfileSelection(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	sessionID := "browser-runtime-clear-target"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "workbench", Target: "node"}

	sessionRegistry.TrackCurrentTarget(sessionID, BrowserSessionTarget{
		ID:         "target-9",
		TabIndex:   9,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "remember_profile")
	stateRegistry.SelectBrowserProfile(sessionID, SharedSessionBrowserProfileSelection{
		Backend:       "proxy",
		Profile:       "workbench",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Source:        "remember_profile",
	})

	result := ExecuteSharedSessionBrowserClearTarget(BuildSharedSessionBrowserClearRequest(
		sessionRegistry,
		stateRegistry,
		sessionID,
		BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
		route,
		false,
		"",
		SharedSessionBrowserHealthInput{},
		time.Minute,
	))
	if result.Decision != "session_target_cleared" || !result.Ready || !result.ClearedTargetSelection || !result.ClearedProfileSelection {
		t.Fatalf("expected clear_target to clear target+remembered profile selection, got %#v", result)
	}
	if selection := CurrentSharedSessionBrowserTargetSelection(sessionRegistry, sessionID, route); selection != nil {
		t.Fatalf("expected target selection to be cleared, got %#v", selection)
	}
	if _, ok := stateRegistry.SelectedBrowserProfile(sessionID, "node"); ok {
		t.Fatalf("expected remembered profile selection to be cleared")
	}
}

func TestExecuteSharedSessionBrowserClearSessionClearsRouteState(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	sessionID := "browser-runtime-clear-session"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "workbench", Target: "node"}

	sessionRegistry.TrackTab(sessionID, BrowserSessionTarget{
		TabIndex:   3,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, true)
	stateRegistry.SelectBrowserProfile(sessionID, SharedSessionBrowserProfileSelection{
		Backend:       "proxy",
		Profile:       "workbench",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Source:        "select_profile",
	})
	if _, ok := stateRegistry.SyncSessionBrowserProfileObservation(sessionID, SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "workbench",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "running",
		Running:       true,
		Connected:     true,
	}, time.Minute); !ok {
		t.Fatalf("expected shared profile observation sync to succeed")
	}

	result := ExecuteSharedSessionBrowserClearSession(BuildSharedSessionBrowserClearRequest(
		sessionRegistry,
		stateRegistry,
		sessionID,
		BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
		route,
		false,
		"",
		SharedSessionBrowserHealthInput{},
		time.Minute,
	))
	if result.Decision != "session_route_cleared" || !result.Ready {
		t.Fatalf("expected clear_session to report route cleared, got %#v", result)
	}
	if !result.ClearedProfileSelection || !result.ClearedTargetSelection || result.ClearedSessionProfiles != 1 || result.ClearedSessionTargets != 1 {
		t.Fatalf("expected clear_session to clear route selections and state, got %#v", result)
	}
}

func TestExecuteSharedSessionBrowserClearSessionBlockedActiveNodeRunKeepsEffectiveLifecycleState(t *testing.T) {
	result := ExecuteSharedSessionBrowserClearSession(BuildSharedSessionBrowserClearRequest(
		nil,
		nil,
		"browser-runtime-clear-session-blocked",
		BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
		BrowserSessionRoute{Backend: "proxy", Profile: "workbench", Target: "node"},
		false,
		"run-77",
		SharedSessionBrowserHealthInput{
			Profiles: []SharedSessionBrowserProfileState{{
				Backend:       "proxy",
				Profile:       "workbench",
				RuntimeTarget: "node",
				BrowserApp:    "Chromium",
				Status:        "running",
				Running:       true,
				Connected:     true,
			}},
		},
		time.Minute,
	))
	if result.Decision != "clear_session_blocked_active_node_run" || result.Ready || result.ProfileStatus.Profile != "workbench" || result.ProfileStatus.Status != "running" || !result.ProfileStatus.Running || !result.ProfileStatus.Connected {
		t.Fatalf("expected blocked clear_session to preserve effective lifecycle state, got %#v", result)
	}
}

func TestExecuteSharedSessionBrowserClearTargetWithContextSeedsSiblingProviderFromPrimaryCachedEventCycle(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	runRegistryA := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-clear-target-helper-sibling-provider": {{RunID: "run-a", Status: "running"}},
		}},
	}
	runRegistryB := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-clear-target-helper-sibling-provider": {{RunID: "run-b", Status: "running"}},
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
	sessionID := "sess-clear-target-helper-sibling-provider"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}
	mutationCtx := SharedSessionBrowserMutationContext{
		Registry:        sessionRegistry,
		RunRegistry:     runRegistryA,
		ReconnectWindow: time.Minute,
	}

	tracked := sessionRegistry.TrackTab(sessionID, BrowserSessionTarget{
		ID:         "tab-1",
		TabIndex:   1,
		URL:        "https://example.com/one",
		Title:      "One",
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

	initialA := boundA.ObserveWatchLoop(context.Background(), req)
	initialB := boundB.ObserveWatchLoop(context.Background(), req)
	if len(initialA.View.Session.Routes) != 1 || initialA.View.Session.Routes[0].CurrentTargetID != tracked.ID {
		t.Fatalf("expected first provider to expose tracked target, got %#v", initialA.View.Session.Routes)
	}
	if len(initialB.View.Session.Routes) != 1 || initialB.View.Session.Routes[0].CurrentTargetID != tracked.ID {
		t.Fatalf("expected second provider to expose tracked target, got %#v", initialB.View.Session.Routes)
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

	result := ExecuteSharedSessionBrowserClearTargetWithContext(
		mutationCtx,
		BuildSharedSessionBrowserClearRequest(
			sessionRegistry,
			nil,
			sessionID,
			BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
			route,
			false,
			"",
			SharedSessionBrowserHealthInput{},
			time.Minute,
		),
	)
	if !result.Ready || !result.ClearedTargetSelection {
		t.Fatalf("expected owner-aware clear-target helper to clear current target, got %#v", result)
	}

	seededB := boundB.ObserveWatchLoop(context.Background(), req)
	if len(seededB.View.Session.Routes) != 1 || seededB.View.Session.Routes[0].CurrentTargetID != "" {
		t.Fatalf("expected sibling watch loop to reuse helper-seeded cleared target, got %#v", seededB.View.Session.Routes)
	}
	if len(backend.statusReqs) != 2 || len(backend.profilesReqs) != 2 {
		t.Fatalf("expected helper-seeded sibling watch loop to avoid extra polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}

func TestExecuteSharedSessionBrowserClearActionWithContextDispatchesClearSession(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	sessionID := "browser-runtime-clear-action-helper"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "workbench", Target: "node"}

	sessionRegistry.TrackCurrentTarget(sessionID, BrowserSessionTarget{
		ID:         "tab-1",
		TabIndex:   1,
		URL:        "https://example.com/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "tracked_active_tab")
	stateRegistry.SelectBrowserProfile(sessionID, SharedSessionBrowserProfileSelection{
		Backend:       "proxy",
		Profile:       "workbench",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Source:        "select_profile",
	})
	if _, ok := stateRegistry.SyncSessionBrowserProfileObservation(sessionID, SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "workbench",
		RuntimeTarget: "node",
		Status:        "running",
		Running:       true,
		Connected:     true,
	}, time.Minute); !ok {
		t.Fatalf("expected shared profile observation sync to succeed")
	}

	result := ExecuteSharedSessionBrowserClearActionWithContext(
		"clear_session",
		SharedSessionBrowserMutationContextFor(
			SharedSessionBrowserObserverManager{},
			sessionRegistry,
			stateRegistry,
			time.Minute,
		),
		BuildSharedSessionBrowserClearRequest(
			sessionRegistry,
			stateRegistry,
			sessionID,
			BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
			route,
			false,
			"",
			SharedSessionBrowserHealthInput{},
			time.Minute,
		),
	)
	if result.Decision != "session_route_cleared" || !result.Ready {
		t.Fatalf("expected clear action helper to dispatch clear_session, got %#v", result)
	}
	if !result.ClearedProfileSelection || !result.ClearedTargetSelection || result.ClearedSessionProfiles != 1 || result.ClearedSessionTargets != 1 {
		t.Fatalf("expected clear action helper to preserve route cleanup counts, got %#v", result)
	}
}
