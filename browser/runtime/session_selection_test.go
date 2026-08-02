package browserruntime

import (
	"context"
	"testing"
	"time"
)

func TestSelectSharedSessionBrowserProfileValidatesProfilesAndPersistsSelection(t *testing.T) {
	registry := NewBrowserSessionStateRegistry()
	backend := &executionTestBackend{
		profilesResp: BrowserProfilesResult{
			Backend: "proxy",
			Profiles: []BrowserProfileInfo{
				{Profile: "isolated", BrowserApp: "Chromium", Status: "stopped"},
				{Profile: "workbench", BrowserApp: "Chromium", Status: "running", Running: true, Connected: true},
			},
		},
	}
	selection, decision, ok, err := SelectSharedSessionBrowserProfile(
		context.Background(),
		registry,
		"s1",
		BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
		"",
		backend,
		true,
		"select_profile",
	)
	if err != nil {
		t.Fatalf("SelectSharedSessionBrowserProfile returned error: %v", err)
	}
	if !ok || decision != "session_profile_selected" {
		t.Fatalf("expected selection to succeed, got ok=%v decision=%q", ok, decision)
	}
	if len(backend.profilesReqs) != 1 {
		t.Fatalf("expected profile validation to hit backend once, got %#v", backend.profilesReqs)
	}
	if selection.BrowserApp != "Chromium" || selection.Profile != "workbench" || selection.RuntimeTarget != "node" || selection.Source != "select_profile" {
		t.Fatalf("unexpected selection: %#v", selection)
	}
	stored, storedOK := registry.SelectedBrowserProfile("s1", "node")
	if !storedOK || stored.Profile != "workbench" || stored.BrowserApp != "Chromium" {
		t.Fatalf("expected selection to persist in registry, got %#v ok=%v", stored, storedOK)
	}
}

func TestSelectSharedSessionBrowserProfileReusesExistingSelectionBrowserApp(t *testing.T) {
	registry := NewBrowserSessionStateRegistry()
	registry.SelectBrowserProfile("s1", SharedSessionBrowserProfileSelection{
		Backend:       "proxy",
		Profile:       "workbench",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Source:        "sync_session",
	})
	selection, decision, ok, err := SelectSharedSessionBrowserProfile(
		context.Background(),
		registry,
		"s1",
		BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
		"",
		nil,
		false,
		"select_profile",
	)
	if err != nil {
		t.Fatalf("SelectSharedSessionBrowserProfile returned error: %v", err)
	}
	if !ok || decision != "session_profile_already_selected" {
		t.Fatalf("expected existing selection reuse, got ok=%v decision=%q", ok, decision)
	}
	if selection.BrowserApp != "Chromium" {
		t.Fatalf("expected existing browser app to be preserved, got %#v", selection)
	}
}

func TestSelectSharedSessionBrowserProfileReportsMissingValidatedProfile(t *testing.T) {
	registry := NewBrowserSessionStateRegistry()
	backend := &executionTestBackend{
		profilesResp: BrowserProfilesResult{
			Backend: "proxy",
			Profiles: []BrowserProfileInfo{
				{Profile: "isolated", BrowserApp: "Chromium", Status: "stopped"},
			},
		},
	}
	_, decision, ok, err := SelectSharedSessionBrowserProfile(
		context.Background(),
		registry,
		"s1",
		BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
		"",
		backend,
		true,
		"select_profile",
	)
	if err == nil || ok {
		t.Fatalf("expected missing validated profile to fail, got ok=%v err=%v", ok, err)
	}
	if decision != "session_profile_missing" {
		t.Fatalf("unexpected decision: %q", decision)
	}
}

func TestPromoteSharedSessionBrowserProfileFromTargetSelectionPersistsSelection(t *testing.T) {
	registry := NewBrowserSessionStateRegistry()
	selection, ok := PromoteSharedSessionBrowserProfileFromTargetSelection(
		registry,
		"s1",
		&BrowserSessionTargetSelection{
			Backend:       "proxy-extract",
			Profile:       "workbench",
			RuntimeTarget: "node",
			BrowserApp:    "Chromium",
			Source:        "select_target",
		},
	)
	if !ok {
		t.Fatalf("expected target-derived profile selection to be promoted")
	}
	if selection.Backend != "proxy" || selection.Profile != "workbench" || selection.RuntimeTarget != "node" || selection.BrowserApp != "Chromium" || selection.Source != "select_target" {
		t.Fatalf("unexpected promoted selection: %#v", selection)
	}
	stored, storedOK := registry.SelectedBrowserProfile("s1", "node")
	if !storedOK || stored.Profile != "workbench" || stored.Backend != "proxy" {
		t.Fatalf("expected promoted selection to persist, got %#v ok=%v", stored, storedOK)
	}
}

func TestPromoteSharedSessionBrowserProfileFromTargetSelectionReusesExistingSelection(t *testing.T) {
	registry := NewBrowserSessionStateRegistry()
	registry.SelectBrowserProfile("s1", SharedSessionBrowserProfileSelection{
		Backend:       "proxy",
		Profile:       "workbench",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Source:        "remember_target",
	})
	selection, ok := PromoteSharedSessionBrowserProfileFromTargetSelection(
		registry,
		"s1",
		&BrowserSessionTargetSelection{
			Backend:       "proxy",
			Profile:       "workbench",
			RuntimeTarget: "node",
			Source:        "select_target",
		},
	)
	if !ok {
		t.Fatalf("expected target-derived profile selection to reuse existing selection")
	}
	if selection.BrowserApp != "Chromium" || selection.Source != "remember_target" {
		t.Fatalf("expected existing richer selection to be preserved, got %#v", selection)
	}
}

func TestRememberSharedSessionBrowserProfilePersistsSelection(t *testing.T) {
	registry := NewBrowserSessionStateRegistry()
	selection, decision, ok := RememberSharedSessionBrowserProfile(
		registry,
		"s1",
		BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
		&BrowserProfileStatusResult{
			Backend:    "proxy",
			BrowserApp: "Chromium",
			Profile:    "workbench",
			Status:     "running",
			Running:    true,
			Connected:  true,
		},
		"",
		"",
		"",
	)
	if !ok || decision != "session_profile_remembered" {
		t.Fatalf("expected remember_profile selection, got ok=%v decision=%q", ok, decision)
	}
	if selection.Profile != "workbench" || selection.RuntimeTarget != "node" || selection.BrowserApp != "Chromium" || selection.Source != "remember_profile" {
		t.Fatalf("unexpected remembered selection: %#v", selection)
	}
	stored, storedOK := registry.SelectedBrowserProfile("s1", "node")
	if !storedOK || stored.Profile != "workbench" || stored.Source != "remember_profile" {
		t.Fatalf("expected remembered selection to persist, got %#v ok=%v", stored, storedOK)
	}
}

func TestRememberSharedSessionBrowserProfileReusesExistingSelection(t *testing.T) {
	registry := NewBrowserSessionStateRegistry()
	registry.SelectBrowserProfile("s1", SharedSessionBrowserProfileSelection{
		Backend:       "proxy",
		Profile:       "workbench",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Source:        "remember_profile",
	})
	selection, decision, ok := RememberSharedSessionBrowserProfile(
		registry,
		"s1",
		BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
		&BrowserProfileStatusResult{
			Backend: "proxy",
			Profile: "workbench",
			Status:  "running",
		},
		"",
		"",
		"",
	)
	if !ok || decision != "session_profile_already_remembered" {
		t.Fatalf("expected remembered selection reuse, got ok=%v decision=%q", ok, decision)
	}
	if selection.BrowserApp != "Chromium" || selection.Source != "remember_profile" {
		t.Fatalf("expected existing remembered selection to be preserved, got %#v", selection)
	}
}

func TestSyncSharedSessionBrowserRouteSelectionPersistsProfileAndTarget(t *testing.T) {
	stateRegistry := NewBrowserSessionStateRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	sessionID := "browser-runtime-sync-session-selection"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "workbench", Target: "node"}
	tracked := sessionRegistry.TrackTabs(sessionID, []BrowserSessionTarget{{
		TabIndex:   1,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}}, 1)

	result, err := SyncSharedSessionBrowserRouteSelection(
		context.Background(),
		stateRegistry,
		sessionRegistry,
		sessionID,
		BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
		route,
		"",
		nil,
		false,
		"sync_session",
	)
	if err != nil {
		t.Fatalf("SyncSharedSessionBrowserRouteSelection returned error: %v", err)
	}
	if result.Decision != "session_route_synced" || !result.Ready {
		t.Fatalf("unexpected sync result: %#v", result)
	}
	if result.ProfileSelection == nil || result.ProfileSelection.Profile != "workbench" || result.ProfileSelection.Source != "sync_session" {
		t.Fatalf("unexpected synced profile selection: %#v", result.ProfileSelection)
	}
	if result.TargetSelection == nil || result.TargetSelection.ID != tracked[0].ID || result.TargetSelection.Source != "sync_session" {
		t.Fatalf("unexpected synced target selection: %#v", result.TargetSelection)
	}
	if stored, ok := stateRegistry.SelectedBrowserProfile(sessionID, "node"); !ok || stored.Profile != "workbench" || stored.Source != "sync_session" {
		t.Fatalf("expected synced profile selection to persist, got %#v ok=%v", stored, ok)
	}
}

func TestSyncSharedSessionBrowserRouteSelectionClearsMismatchedTarget(t *testing.T) {
	stateRegistry := NewBrowserSessionStateRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	sessionID := "browser-runtime-sync-session-clears-mismatch"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "alternate", Target: "node"}
	initialRoute := BrowserSessionRoute{Backend: "proxy", Profile: "workbench", Target: "node"}
	tracked := sessionRegistry.TrackTabs(sessionID, []BrowserSessionTarget{{
		TabIndex:   1,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}}, 1)
	if _, ok := sessionRegistry.SelectTargetForRoute(sessionID, initialRoute, tracked[0].ID, "select_target"); !ok {
		t.Fatalf("expected initial target selection to succeed")
	}

	result, err := SyncSharedSessionBrowserRouteSelection(
		context.Background(),
		stateRegistry,
		sessionRegistry,
		sessionID,
		BrowserRuntimeInfo{Backend: "proxy", Profile: "alternate", Target: "node"},
		route,
		"",
		nil,
		false,
		"sync_session",
	)
	if err != nil {
		t.Fatalf("SyncSharedSessionBrowserRouteSelection returned error: %v", err)
	}
	if result.Decision != "session_profile_synced" || !result.Ready {
		t.Fatalf("unexpected mismatched-target sync result: %#v", result)
	}
	if result.ProfileSelection == nil || result.ProfileSelection.Profile != "alternate" {
		t.Fatalf("expected alternate profile selection, got %#v", result.ProfileSelection)
	}
	if result.TargetSelection != nil {
		t.Fatalf("expected mismatched target to clear instead of sync, got %#v", result.TargetSelection)
	}
	if current := CurrentSharedSessionBrowserTargetSelection(sessionRegistry, sessionID, route); current != nil {
		t.Fatalf("expected mismatched current target to clear, got %#v", current)
	}
}

func TestRememberSharedSessionBrowserProfileForRouteSyncsCurrentTarget(t *testing.T) {
	stateRegistry := NewBrowserSessionStateRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	sessionID := "browser-runtime-remember-profile-syncs-target"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "workbench", Target: "node"}
	tracked := sessionRegistry.TrackTabs(sessionID, []BrowserSessionTarget{{
		TabIndex:   1,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}}, 1)

	result := RememberSharedSessionBrowserProfileForRoute(
		stateRegistry,
		sessionRegistry,
		sessionID,
		BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
		route,
		&BrowserProfileStatusResult{
			Backend:    "proxy",
			BrowserApp: "Chromium",
			Profile:    "workbench",
			Status:     "running",
			Running:    true,
			Connected:  true,
		},
		"",
		"",
		"",
	)
	if result.Decision != "session_profile_remembered" || !result.Ready {
		t.Fatalf("unexpected remember-profile result: %#v", result)
	}
	if result.ProfileSelection == nil || result.ProfileSelection.Profile != "workbench" || result.ProfileSelection.Source != "remember_profile" {
		t.Fatalf("unexpected remembered selection: %#v", result.ProfileSelection)
	}
	if result.TargetSelection == nil || result.TargetSelection.ID != tracked[0].ID || result.TargetSelection.Source != "remember_profile" {
		t.Fatalf("expected remembered profile to sync current target, got %#v", result.TargetSelection)
	}
	if result.SelectionProjection == nil || result.SelectionProjection.ProfileSelection == nil || result.SelectionProjection.ProfileSelection.Profile != "workbench" {
		t.Fatalf("expected remember-profile result to carry shared selection projection, got %#v", result.SelectionProjection)
	}
	if result.SelectionProjection.TargetSelection == nil || result.SelectionProjection.TargetSelection.ID != tracked[0].ID || !result.SelectionProjection.ApplyTargetToRoute {
		t.Fatalf("expected remember-profile result to carry target projection, got %#v", result.SelectionProjection)
	}
}

func TestRememberSharedSessionBrowserProfileForRouteClearsMismatchedTarget(t *testing.T) {
	stateRegistry := NewBrowserSessionStateRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	sessionID := "browser-runtime-remember-profile-clears-mismatch"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "alternate", Target: "node"}
	initialRoute := BrowserSessionRoute{Backend: "proxy", Profile: "workbench", Target: "node"}
	tracked := sessionRegistry.TrackTabs(sessionID, []BrowserSessionTarget{{
		TabIndex:   1,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}}, 1)
	if _, ok := sessionRegistry.SelectTargetForRoute(sessionID, initialRoute, tracked[0].ID, "select_target"); !ok {
		t.Fatalf("expected initial target selection to succeed")
	}

	result := RememberSharedSessionBrowserProfileForRoute(
		stateRegistry,
		sessionRegistry,
		sessionID,
		BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
		route,
		&BrowserProfileStatusResult{
			Backend:    "proxy",
			BrowserApp: "Chromium",
			Profile:    "alternate",
			Status:     "running",
			Running:    true,
			Connected:  true,
		},
		"",
		"",
		"",
	)
	if result.Decision != "session_profile_remembered" || !result.Ready {
		t.Fatalf("unexpected remember-profile mismatch result: %#v", result)
	}
	if result.ProfileSelection == nil || result.ProfileSelection.Profile != "alternate" {
		t.Fatalf("expected alternate remembered selection, got %#v", result.ProfileSelection)
	}
	if result.TargetSelection != nil {
		t.Fatalf("expected mismatched remembered target to clear, got %#v", result.TargetSelection)
	}
	if result.SelectionProjection == nil || result.SelectionProjection.ProfileSelection == nil || result.SelectionProjection.ProfileSelection.Profile != "alternate" || !result.SelectionProjection.ApplyTargetToRoute {
		t.Fatalf("expected remember-profile mismatch result to keep shared selection projection, got %#v", result.SelectionProjection)
	}
	if current := CurrentSharedSessionBrowserTargetSelection(sessionRegistry, sessionID, route); current != nil {
		t.Fatalf("expected mismatched remembered target to clear, got %#v", current)
	}
}

func TestDispatchSharedSessionBrowserRememberProfileUsesRememberProfileContract(t *testing.T) {
	stateRegistry := NewBrowserSessionStateRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	sessionID := "browser-runtime-remember-profile-dispatch"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "workbench", Target: "node"}
	tracked := sessionRegistry.TrackTabs(sessionID, []BrowserSessionTarget{{
		TabIndex:   1,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}}, 1)

	result := DispatchSharedSessionBrowserRememberProfile(
		SharedSessionBrowserRememberProfileDispatchRequest{
			MutationContext: SharedSessionBrowserMutationContext{
				Registry:      sessionRegistry,
				StateRegistry: stateRegistry,
			},
			SessionID:    sessionID,
			SelectedInfo: BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
			Route:        route,
			ProfileStatus: &BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "workbench",
				Status:     "running",
				Running:    true,
				Connected:  true,
			},
		},
	)
	if result.Decision != "session_profile_remembered" || !result.Ready {
		t.Fatalf("unexpected remember-profile dispatch result: %#v", result)
	}
	if result.ProfileSelection == nil || result.ProfileSelection.Profile != "workbench" || result.ProfileSelection.Source != "remember_profile" {
		t.Fatalf("expected remember-profile dispatch helper to persist selection, got %#v", result.ProfileSelection)
	}
	if result.TargetSelection == nil || result.TargetSelection.ID != tracked[0].ID || result.TargetSelection.Source != "remember_profile" {
		t.Fatalf("expected remember-profile dispatch helper to sync current target, got %#v", result.TargetSelection)
	}
	if result.SelectionProjection == nil || result.SelectionProjection.ProfileSelection == nil || result.SelectionProjection.TargetSelection == nil {
		t.Fatalf("expected remember-profile dispatch helper to preserve shared selection projection, got %#v", result.SelectionProjection)
	}
}

func TestCoordinateSharedSessionBrowserRouteSyncPrefersCurrentTargetSelection(t *testing.T) {
	stateRegistry := NewBrowserSessionStateRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	sessionID := "browser-runtime-coordinate-sync-target-first"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "workbench", Target: "node"}
	tracked := sessionRegistry.TrackTabs(sessionID, []BrowserSessionTarget{{
		TabIndex:   1,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}}, 1)

	result, err := CoordinateSharedSessionBrowserRouteSync(
		context.Background(),
		stateRegistry,
		sessionRegistry,
		sessionID,
		BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
		route,
		"",
		nil,
		false,
	)
	if err != nil {
		t.Fatalf("CoordinateSharedSessionBrowserRouteSync returned error: %v", err)
	}
	if result.Decision != "session_target_synced" || !result.Ready {
		t.Fatalf("unexpected coordination sync result: %#v", result)
	}
	if result.ProfileSelection != nil {
		t.Fatalf("expected target-first sync to avoid profile selection churn, got %#v", result.ProfileSelection)
	}
	if result.TargetSelection == nil || result.TargetSelection.ID != tracked[0].ID || result.TargetSelection.Source != "sync_session" {
		t.Fatalf("unexpected target-first sync selection: %#v", result.TargetSelection)
	}
}

func TestDispatchSharedSessionBrowserSyncActionUsesSyncSessionRouteSelection(t *testing.T) {
	stateRegistry := NewBrowserSessionStateRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	sessionID := "browser-runtime-sync-dispatch-route-selection"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "workbench", Target: "node"}
	tracked := sessionRegistry.TrackTabs(sessionID, []BrowserSessionTarget{{
		TabIndex:   1,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}}, 1)

	dispatched := DispatchSharedSessionBrowserSyncAction(
		context.Background(),
		SharedSessionBrowserSyncActionDispatchRequest{
			Action: "sync_session",
			MutationContext: SharedSessionBrowserMutationContext{
				Registry:      sessionRegistry,
				StateRegistry: stateRegistry,
			},
			SessionID:    sessionID,
			SelectedInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
			Route:        route,
		},
	)
	if !dispatched.Handled || dispatched.Err != nil {
		t.Fatalf("expected sync dispatch helper to handle sync_session, got %#v", dispatched)
	}
	if dispatched.Result.Decision != "session_route_synced" || !dispatched.Result.Ready {
		t.Fatalf("unexpected sync dispatch result: %#v", dispatched.Result)
	}
	if dispatched.Result.ProfileSelection == nil || dispatched.Result.ProfileSelection.Profile != "workbench" {
		t.Fatalf("expected sync dispatch helper to persist profile selection, got %#v", dispatched.Result.ProfileSelection)
	}
	if dispatched.Result.TargetSelection == nil || dispatched.Result.TargetSelection.ID != tracked[0].ID || dispatched.Result.TargetSelection.Source != "sync_session" {
		t.Fatalf("expected sync dispatch helper to persist target selection, got %#v", dispatched.Result.TargetSelection)
	}
}

func TestDispatchSharedSessionBrowserSyncActionUsesCoordinateTargetFirstContract(t *testing.T) {
	stateRegistry := NewBrowserSessionStateRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	sessionID := "browser-runtime-sync-dispatch-coordinate-sync"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "workbench", Target: "node"}
	tracked := sessionRegistry.TrackTabs(sessionID, []BrowserSessionTarget{{
		TabIndex:   1,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}}, 1)

	dispatched := DispatchSharedSessionBrowserSyncAction(
		context.Background(),
		SharedSessionBrowserSyncActionDispatchRequest{
			Action:           "coordinate",
			CoordinationGoal: "sync",
			MutationContext: SharedSessionBrowserMutationContext{
				Registry:      sessionRegistry,
				StateRegistry: stateRegistry,
			},
			SessionID:    sessionID,
			SelectedInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
			Route:        route,
		},
	)
	if !dispatched.Handled || dispatched.Err != nil {
		t.Fatalf("expected sync dispatch helper to handle coordinate(sync), got %#v", dispatched)
	}
	if dispatched.Result.Decision != "session_target_synced" || !dispatched.Result.Ready {
		t.Fatalf("unexpected coordinate(sync) dispatch result: %#v", dispatched.Result)
	}
	if dispatched.Result.ProfileSelection != nil {
		t.Fatalf("expected coordinate(sync) dispatch to preserve target-first contract, got %#v", dispatched.Result.ProfileSelection)
	}
	if dispatched.Result.TargetSelection == nil || dispatched.Result.TargetSelection.ID != tracked[0].ID || dispatched.Result.TargetSelection.Source != "sync_session" {
		t.Fatalf("expected coordinate(sync) dispatch to sync current target, got %#v", dispatched.Result.TargetSelection)
	}
}

func TestDispatchSharedSessionBrowserSyncActionReturnsUnhandledForOtherActions(t *testing.T) {
	dispatched := DispatchSharedSessionBrowserSyncAction(
		context.Background(),
		SharedSessionBrowserSyncActionDispatchRequest{Action: "select_profile"},
	)
	if dispatched.Handled {
		t.Fatalf("expected non-sync action to remain unhandled, got %#v", dispatched)
	}
}

func TestDispatchSharedSessionBrowserSelectionActionUsesSelectProfileContract(t *testing.T) {
	stateRegistry := NewBrowserSessionStateRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	sessionID := "browser-runtime-selection-dispatch-profile"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "workbench", Target: "node"}
	tracked := sessionRegistry.TrackTabs(sessionID, []BrowserSessionTarget{{
		TabIndex:   1,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}}, 1)

	dispatched := DispatchSharedSessionBrowserSelectionAction(
		context.Background(),
		SharedSessionBrowserSelectionActionDispatchRequest{
			Action: "select_profile",
			MutationContext: SharedSessionBrowserMutationContext{
				Registry:      sessionRegistry,
				StateRegistry: stateRegistry,
			},
			SessionID:    sessionID,
			SelectedInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
			Route:        route,
			Source:       "select_profile",
		},
	)
	if !dispatched.Handled || dispatched.Err != nil {
		t.Fatalf("expected selection dispatch helper to handle select_profile, got %#v", dispatched)
	}
	if dispatched.Decision != "session_profile_selected" || !dispatched.Ready {
		t.Fatalf("unexpected select_profile dispatch result: %#v", dispatched)
	}
	if dispatched.ProfileSelection == nil || dispatched.ProfileSelection.Profile != "workbench" || dispatched.ProfileSelection.Source != "select_profile" {
		t.Fatalf("expected selection dispatch helper to persist profile selection, got %#v", dispatched.ProfileSelection)
	}
	if dispatched.TargetSelection == nil || dispatched.TargetSelection.ID != tracked[0].ID || dispatched.TargetSelection.Source != "select_profile" {
		t.Fatalf("expected selection dispatch helper to sync current target for profile selection, got %#v", dispatched.TargetSelection)
	}
}

func TestDispatchSharedSessionBrowserSelectionActionUsesSelectTargetPromotion(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	sessionID := "browser-runtime-selection-dispatch-target"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "workbench", Target: "node"}
	tracked := sessionRegistry.TrackTabs(sessionID, []BrowserSessionTarget{{
		TabIndex:   2,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}}, 2)

	dispatched := DispatchSharedSessionBrowserSelectionAction(
		context.Background(),
		SharedSessionBrowserSelectionActionDispatchRequest{
			Action: "select_target",
			MutationContext: SharedSessionBrowserMutationContext{
				Registry:      sessionRegistry,
				StateRegistry: stateRegistry,
			},
			SessionID:    sessionID,
			SelectedInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
			Route:        route,
			Source:       "select_target",
			TargetRequest: &SharedSessionBrowserSelectTargetRequest{
				SessionID: sessionID,
				Route:     route,
				TabIndex:  2,
				Source:    "select_target",
				Actor:     "browser_runtime target selection",
			},
		},
	)
	if !dispatched.Handled || dispatched.Err != nil {
		t.Fatalf("expected selection dispatch helper to handle select_target, got %#v", dispatched)
	}
	if dispatched.Decision != "session_target_already_selected" || !dispatched.Ready {
		t.Fatalf("unexpected select_target dispatch result: %#v", dispatched)
	}
	if dispatched.TargetSelection == nil || dispatched.TargetSelection.ID != tracked[0].ID || dispatched.TargetSelection.Source != "select_target" {
		t.Fatalf("expected selection dispatch helper to project target selection, got %#v", dispatched.TargetSelection)
	}
	if dispatched.ProfileSelection == nil || dispatched.ProfileSelection.Profile != "workbench" || dispatched.ProfileSelection.Source != "select_target" {
		t.Fatalf("expected selection dispatch helper to promote profile selection, got %#v", dispatched.ProfileSelection)
	}
}

func TestDispatchSharedSessionBrowserSelectionActionReturnsUnhandledForOtherActions(t *testing.T) {
	dispatched := DispatchSharedSessionBrowserSelectionAction(
		context.Background(),
		SharedSessionBrowserSelectionActionDispatchRequest{Action: "sync_session"},
	)
	if dispatched.Handled {
		t.Fatalf("expected non-selection action to remain unhandled, got %#v", dispatched)
	}
}

func TestSelectSharedSessionBrowserProfileWithContextSeedsSiblingProviderBindingFromPrimaryCachedEventCycle(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	runRegistryA := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-select-profile-helper-sibling-provider": {{RunID: "run-a", Status: "running"}},
		}},
	}
	runRegistryB := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-select-profile-helper-sibling-provider": {{RunID: "run-b", Status: "running"}},
		}},
	}
	managerA := SharedSessionBrowserObserverManagerFor(sessionRegistry, runRegistryA, stateRegistry, time.Minute)
	managerB := SharedSessionBrowserObserverManagerFor(sessionRegistry, runRegistryB, stateRegistry, time.Minute)
	backend := &statusProfilesObservationTestBackend{
		statusResp: BrowserProfileStatusResult{
			Backend:    "proxy",
			Profile:    "workbench",
			BrowserApp: "Chromium",
			Status:     "running",
			Running:    true,
			Connected:  true,
		},
		profilesResp: BrowserProfilesResult{
			Backend:        "proxy",
			DefaultProfile: "workbench",
			Profiles: []BrowserProfileInfo{
				{Profile: "workbench", BrowserApp: "Chromium", Status: "connected", Running: true, Connected: true},
			},
		},
	}
	boundA := managerA.Bind(backend)
	boundB := managerB.Bind(backend)
	sessionID := "sess-select-profile-helper-sibling-provider"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "workbench", Target: "node", BrowserApp: "Chromium"}
	mutationCtx := SharedSessionBrowserMutationContext{
		Registry:        sessionRegistry,
		RunRegistry:     runRegistryA,
		StateRegistry:   stateRegistry,
		ReconnectWindow: time.Minute,
	}

	req := SharedSessionBrowserObserverRequest{
		SessionID:        sessionID,
		SelectedInfo:     BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
		BindingRoute:     route,
		RequestedProfile: "workbench",
		IncludeStatus:    true,
		IncludeProfiles:  true,
	}

	initialA := boundA.ObserveBinding(context.Background(), req)
	initialB := boundB.ObserveBinding(context.Background(), req)
	if initialA.Evaluation.Snapshot.SelectedProfileSelection != nil || initialB.Evaluation.Snapshot.SelectedProfileSelection != nil {
		t.Fatalf("expected initial binding snapshots to have no selected profile, got A=%#v B=%#v", initialA.Evaluation.Snapshot.SelectedProfileSelection, initialB.Evaluation.Snapshot.SelectedProfileSelection)
	}
	if len(backend.statusReqs) != 2 || len(backend.profilesReqs) != 2 {
		t.Fatalf("expected initial sibling bindings to poll backend once each, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
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

	selection, decision, ok, err := SelectSharedSessionBrowserProfileWithContext(
		mutationCtx,
		context.Background(),
		sessionID,
		BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
		"",
		nil,
		false,
		"select_profile",
	)
	if err != nil {
		t.Fatalf("SelectSharedSessionBrowserProfileWithContext returned error: %v", err)
	}
	if !ok || decision != "session_profile_selected" {
		t.Fatalf("expected owner-aware select-profile helper to succeed, got ok=%v decision=%q", ok, decision)
	}
	if selection.Profile != "workbench" || selection.Source != "select_profile" {
		t.Fatalf("unexpected owner-aware selected profile: %#v", selection)
	}

	seededB := boundB.ObserveBinding(context.Background(), req)
	if seededB.Evaluation.Snapshot.SelectedProfileSelection == nil ||
		seededB.Evaluation.Snapshot.SelectedProfileSelection.Profile != "workbench" ||
		seededB.Evaluation.Snapshot.SelectedProfileSelection.Source != "select_profile" {
		t.Fatalf("expected sibling binding to reuse helper-seeded selected profile, got %#v", seededB.Evaluation.Snapshot.SelectedProfileSelection)
	}
	if len(backend.statusReqs) != 2 || len(backend.profilesReqs) != 2 {
		t.Fatalf("expected helper-seeded sibling binding to avoid extra polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}

func TestSyncSharedSessionBrowserRouteSelectionWithContextSeedsSiblingProviderFromPrimaryCachedEventCycle(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	runRegistryA := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-sync-route-selection-helper-sibling-provider": {{RunID: "run-a", Status: "running"}},
		}},
	}
	runRegistryB := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-sync-route-selection-helper-sibling-provider": {{RunID: "run-b", Status: "running"}},
		}},
	}
	managerA := SharedSessionBrowserObserverManagerFor(sessionRegistry, runRegistryA, stateRegistry, time.Minute)
	managerB := SharedSessionBrowserObserverManagerFor(sessionRegistry, runRegistryB, stateRegistry, time.Minute)
	backend := &statusProfilesObservationTestBackend{
		statusResp: BrowserProfileStatusResult{
			Backend:    "proxy",
			Profile:    "workbench",
			BrowserApp: "Chromium",
			Status:     "running",
			Running:    true,
			Connected:  true,
		},
		profilesResp: BrowserProfilesResult{
			Backend:        "proxy",
			DefaultProfile: "workbench",
			Profiles: []BrowserProfileInfo{
				{Profile: "workbench", BrowserApp: "Chromium", Status: "connected", Running: true, Connected: true},
			},
		},
	}
	boundA := managerA.Bind(backend)
	boundB := managerB.Bind(backend)
	sessionID := "sess-sync-route-selection-helper-sibling-provider"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "workbench", Target: "node", BrowserApp: "Chromium"}
	mutationCtx := SharedSessionBrowserMutationContext{
		Registry:        sessionRegistry,
		RunRegistry:     runRegistryA,
		StateRegistry:   stateRegistry,
		ReconnectWindow: time.Minute,
	}
	tracked := SyncSharedSessionBrowserTabsForRoute(sessionRegistry, sessionID, route, 1, []BrowserTab{
		{Index: 1, URL: "https://example.com/one", Title: "One"},
	})

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

	initialA := boundA.ObserveWatchLoop(context.Background(), req)
	initialB := boundB.ObserveWatchLoop(context.Background(), req)
	if len(initialA.View.Session.Routes) != 1 || initialA.View.Session.Routes[0].CurrentTargetSource != "tracked_active_tab" {
		t.Fatalf("expected first provider to expose tracked_active_tab before sync, got %#v", initialA.View.Session.Routes)
	}
	if len(initialB.View.Session.Routes) != 1 || initialB.View.Session.Routes[0].CurrentTargetSource != "tracked_active_tab" {
		t.Fatalf("expected second provider to expose tracked_active_tab before sync, got %#v", initialB.View.Session.Routes)
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

	result, err := SyncSharedSessionBrowserRouteSelectionWithContext(
		mutationCtx,
		context.Background(),
		sessionID,
		BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
		route,
		"",
		nil,
		false,
		"sync_session",
	)
	if err != nil {
		t.Fatalf("SyncSharedSessionBrowserRouteSelectionWithContext returned error: %v", err)
	}
	if !result.Ready || result.Decision != "session_route_synced" {
		t.Fatalf("unexpected owner-aware route sync result: %#v", result)
	}
	if result.ProfileSelection == nil || result.TargetSelection == nil || result.TargetSelection.ID != tracked[0].TargetID {
		t.Fatalf("expected owner-aware route sync to persist profile and target selections, got %#v", result)
	}

	seededB := boundB.ObserveWatchLoop(context.Background(), req)
	if len(seededB.View.Session.Routes) != 1 ||
		seededB.View.Session.Routes[0].CurrentTargetID != tracked[0].TargetID ||
		seededB.View.Session.Routes[0].CurrentTargetSource != "sync_session" {
		t.Fatalf("expected sibling watch loop to reuse helper-seeded sync_session target, got %#v", seededB.View.Session.Routes)
	}
	if stored, ok := stateRegistry.SelectedBrowserProfile(sessionID, "node"); !ok || stored.Profile != "workbench" || stored.Source != "sync_session" {
		t.Fatalf("expected owner-aware route sync to persist selected profile, got %#v ok=%v", stored, ok)
	}
	if len(backend.statusReqs) != 2 || len(backend.profilesReqs) != 2 {
		t.Fatalf("expected helper-seeded sibling watch loop to avoid extra polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}

func TestRememberSharedSessionBrowserProfileForRouteWithContextSeedsSiblingProviderFromPrimaryCachedEventCycle(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	runRegistryA := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-remember-profile-route-helper-sibling-provider": {{RunID: "run-a", Status: "running"}},
		}},
	}
	runRegistryB := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-remember-profile-route-helper-sibling-provider": {{RunID: "run-b", Status: "running"}},
		}},
	}
	managerA := SharedSessionBrowserObserverManagerFor(sessionRegistry, runRegistryA, stateRegistry, time.Minute)
	managerB := SharedSessionBrowserObserverManagerFor(sessionRegistry, runRegistryB, stateRegistry, time.Minute)
	backend := &statusProfilesObservationTestBackend{
		statusResp: BrowserProfileStatusResult{
			Backend:    "proxy",
			Profile:    "workbench",
			BrowserApp: "Chromium",
			Status:     "running",
			Running:    true,
			Connected:  true,
		},
		profilesResp: BrowserProfilesResult{
			Backend:        "proxy",
			DefaultProfile: "workbench",
			Profiles: []BrowserProfileInfo{
				{Profile: "workbench", BrowserApp: "Chromium", Status: "connected", Running: true, Connected: true},
			},
		},
	}
	boundA := managerA.Bind(backend)
	boundB := managerB.Bind(backend)
	sessionID := "sess-remember-profile-route-helper-sibling-provider"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "workbench", Target: "node", BrowserApp: "Chromium"}
	mutationCtx := SharedSessionBrowserMutationContext{
		Registry:        sessionRegistry,
		RunRegistry:     runRegistryA,
		StateRegistry:   stateRegistry,
		ReconnectWindow: time.Minute,
	}
	tracked := SyncSharedSessionBrowserTabsForRoute(sessionRegistry, sessionID, route, 1, []BrowserTab{
		{Index: 1, URL: "https://example.com/one", Title: "One"},
	})

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

	initialA := boundA.ObserveWatchLoop(context.Background(), req)
	initialB := boundB.ObserveWatchLoop(context.Background(), req)
	if len(initialA.View.Session.Routes) != 1 || initialA.View.Session.Routes[0].CurrentTargetSource != "tracked_active_tab" {
		t.Fatalf("expected first provider to expose tracked_active_tab before remember_profile, got %#v", initialA.View.Session.Routes)
	}
	if len(initialB.View.Session.Routes) != 1 || initialB.View.Session.Routes[0].CurrentTargetSource != "tracked_active_tab" {
		t.Fatalf("expected second provider to expose tracked_active_tab before remember_profile, got %#v", initialB.View.Session.Routes)
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

	result := RememberSharedSessionBrowserProfileForRouteWithContext(
		mutationCtx,
		sessionID,
		BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
		route,
		&BrowserProfileStatusResult{
			Backend:    "proxy",
			BrowserApp: "Chromium",
			Profile:    "workbench",
			Status:     "running",
			Running:    true,
			Connected:  true,
		},
		"",
		"",
		"",
	)
	if !result.Ready || result.Decision != "session_profile_remembered" {
		t.Fatalf("unexpected owner-aware remember-profile result: %#v", result)
	}
	if result.ProfileSelection == nil || result.TargetSelection == nil || result.TargetSelection.ID != tracked[0].TargetID {
		t.Fatalf("expected owner-aware remember-profile helper to persist profile and target selections, got %#v", result)
	}

	seededB := boundB.ObserveWatchLoop(context.Background(), req)
	if len(seededB.View.Session.Routes) != 1 ||
		seededB.View.Session.Routes[0].CurrentTargetID != tracked[0].TargetID ||
		seededB.View.Session.Routes[0].CurrentTargetSource != "remember_profile" {
		t.Fatalf("expected sibling watch loop to reuse helper-seeded remember_profile target, got %#v", seededB.View.Session.Routes)
	}
	if stored, ok := stateRegistry.SelectedBrowserProfile(sessionID, "node"); !ok || stored.Profile != "workbench" || stored.Source != "remember_profile" {
		t.Fatalf("expected owner-aware remember-profile helper to persist selected profile, got %#v ok=%v", stored, ok)
	}
	if len(backend.statusReqs) != 2 || len(backend.profilesReqs) != 2 {
		t.Fatalf("expected helper-seeded sibling watch loop to avoid extra polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}
