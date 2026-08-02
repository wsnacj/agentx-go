package browserruntime

import (
	"testing"
	"time"
)

func TestBrowserSessionStateRegistry_SelectedBrowserProfile(t *testing.T) {
	registry := NewBrowserSessionStateRegistry()
	registry.SelectBrowserProfile("s1", SharedSessionBrowserProfileSelection{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Source:        "select_profile",
	})

	selected, ok := registry.SelectedBrowserProfile("s1", "node")
	if !ok {
		t.Fatalf("expected selected browser profile")
	}
	if selected.Profile != "isolated" || selected.RuntimeTarget != "node" || selected.Backend != "proxy" {
		t.Fatalf("unexpected selected profile: %#v", selected)
	}
	if selected.Source != "select_profile" {
		t.Fatalf("expected selected profile source to be preserved, got %#v", selected)
	}

	snapshot := registry.SnapshotSelectedBrowserProfiles("s1")
	if len(snapshot) != 1 || snapshot[0].Profile != "isolated" || snapshot[0].Source != "select_profile" {
		t.Fatalf("unexpected selected profile snapshot: %#v", snapshot)
	}
}

func TestBrowserSessionStateRegistry_SelectedBrowserProfileFallsBackWhenSingle(t *testing.T) {
	registry := NewBrowserSessionStateRegistry()
	registry.SelectBrowserProfile("s1", SharedSessionBrowserProfileSelection{
		Profile:       "isolated",
		RuntimeTarget: "node",
	})

	selected, ok := registry.SelectedBrowserProfile("s1", "")
	if !ok || selected.Profile != "isolated" || selected.RuntimeTarget != "node" {
		t.Fatalf("expected single selected profile fallback, got %#v ok=%v", selected, ok)
	}
}

func TestBrowserSessionStateRegistry_ClearSelectedBrowserProfile(t *testing.T) {
	registry := NewBrowserSessionStateRegistry()
	registry.SelectBrowserProfile("s1", SharedSessionBrowserProfileSelection{
		Profile:       "isolated",
		RuntimeTarget: "node",
	})
	registry.ClearSelectedBrowserProfile("s1", "node")

	if selected, ok := registry.SelectedBrowserProfile("s1", "node"); ok {
		t.Fatalf("expected selected profile to be cleared, got %#v", selected)
	}
	if snapshot := registry.SnapshotSelectedBrowserProfiles("s1"); len(snapshot) != 0 {
		t.Fatalf("expected empty selected profile snapshot, got %#v", snapshot)
	}
}

func TestBrowserSessionStateRegistryRecordBrowserProfileStateMergesMissingFields(t *testing.T) {
	registry := NewBrowserSessionStateRegistry()
	registry.RecordBrowserProfileState("s1", SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Running:       true,
		Connected:     true,
		Note:          "healthy",
	})
	registry.RecordBrowserProfileState("s1", SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
	})

	snapshot := registry.SnapshotSessionBrowserProfiles("s1")
	if len(snapshot) != 1 {
		t.Fatalf("expected one browser profile state, got %#v", snapshot)
	}
	got := snapshot[0]
	if got.RuntimeTarget != "node" || got.BrowserApp != "Chromium" || got.Status != "" || got.Note != "healthy" {
		t.Fatalf("expected missing fields to inherit from current state, got %#v", got)
	}
	if !got.Running || !got.Connected {
		t.Fatalf("expected running/connected flags to be preserved, got %#v", got)
	}
}

func TestBrowserSessionStateRegistrySnapshotSessionBrowserProfilesSortsAndFilters(t *testing.T) {
	registry := NewBrowserSessionStateRegistry()
	registry.RecordBrowserProfileState("s1", SharedSessionBrowserProfileState{
		Backend:       "system",
		RuntimeTarget: "host",
		Profile:       "default",
		BrowserApp:    "Safari",
	})
	registry.RecordBrowserProfileState("s1", SharedSessionBrowserProfileState{
		Backend:       "proxy",
		RuntimeTarget: "node-b",
		Profile:       "workbench",
		BrowserApp:    "Chromium",
	})
	registry.RecordBrowserProfileState("s1", SharedSessionBrowserProfileState{
		Backend:       "proxy",
		RuntimeTarget: "node-a",
		Profile:       "isolated",
		BrowserApp:    "Chromium",
	})

	snapshot := registry.SnapshotSessionBrowserProfiles("s1")
	if len(snapshot) != 3 {
		t.Fatalf("expected three browser profile states, got %#v", snapshot)
	}
	if snapshot[0].Backend != "proxy" || snapshot[0].RuntimeTarget != "node-a" {
		t.Fatalf("expected snapshot to sort by backend/runtime target/profile/browser app, got %#v", snapshot)
	}

	cleared := registry.ClearSessionBrowserProfiles("s1", SharedSessionBrowserProfileState{Backend: "proxy"})
	if cleared != 2 {
		t.Fatalf("expected two proxy states to be cleared, got %d", cleared)
	}
	remaining := registry.SnapshotSessionBrowserProfiles("s1")
	if len(remaining) != 1 || remaining[0].Backend != "system" {
		t.Fatalf("expected only system state to remain, got %#v", remaining)
	}
}

func TestBrowserSessionStateRegistrySelectBrowserProfileMergesAndUsesHostFallback(t *testing.T) {
	registry := NewBrowserSessionStateRegistry()
	registry.SelectBrowserProfile("s1", SharedSessionBrowserProfileSelection{
		Backend:    "system",
		Profile:    "default",
		BrowserApp: "Safari",
		Source:     "remember_profile",
	})
	registry.SelectBrowserProfile("s1", SharedSessionBrowserProfileSelection{
		Profile: "default",
	})

	selected, ok := registry.SelectedBrowserProfile("s1", "")
	if !ok {
		t.Fatalf("expected host-fallback selected profile")
	}
	if selected.Backend != "system" || selected.BrowserApp != "Safari" || selected.Source != "remember_profile" {
		t.Fatalf("expected missing fields to inherit on selected profile merge, got %#v", selected)
	}
	if browserSessionSelectionKey("") != "host" {
		t.Fatalf("expected blank runtime target to normalize to host")
	}
}

func TestBrowserSessionStateRegistryRecordBrowserProfileStatePreservesStatusSinceUntilHealthStateChanges(t *testing.T) {
	registry := NewBrowserSessionStateRegistry()
	base := time.Now().Add(-2 * time.Minute)
	registry.RecordBrowserProfileState("s1", SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		Status:        "reconnecting",
		Running:       true,
		Connected:     false,
		ObservedAt:    base,
		StatusSince:   base,
	})
	registry.RecordBrowserProfileState("s1", SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		Status:        "reconnecting",
		Running:       true,
		Connected:     false,
		ObservedAt:    time.Now(),
	})

	snapshot := registry.SnapshotSessionBrowserProfiles("s1")
	if len(snapshot) != 1 {
		t.Fatalf("expected one browser profile state, got %#v", snapshot)
	}
	if !snapshot[0].StatusSince.Equal(base) {
		t.Fatalf("expected reconnecting status_since to be preserved, got %#v", snapshot[0])
	}

	registry.RecordBrowserProfileState("s1", SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		Status:        "running",
		Running:       true,
		Connected:     true,
		ObservedAt:    time.Now(),
	})
	snapshot = registry.SnapshotSessionBrowserProfiles("s1")
	if len(snapshot) != 1 {
		t.Fatalf("expected one browser profile state after status change, got %#v", snapshot)
	}
	if snapshot[0].StatusSince.Equal(base) {
		t.Fatalf("expected status_since to reset after health state change, got %#v", snapshot[0])
	}
}

func TestBrowserSessionStateRegistrySyncSessionBrowserProfilesReplacesRouteScopedSnapshot(t *testing.T) {
	registry := NewBrowserSessionStateRegistry()
	base := time.Now().Add(-2 * time.Minute)
	registry.RecordBrowserProfileState("s1", SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "reconnecting",
		Running:       true,
		Connected:     false,
		ObservedAt:    base,
		StatusSince:   base,
	})
	registry.RecordBrowserProfileState("s1", SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "relay",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "stopped",
	})
	registry.RecordBrowserProfileState("s1", SharedSessionBrowserProfileState{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Status:        "running",
		Running:       true,
		Connected:     true,
	})

	registry.SyncSessionBrowserProfiles("s1", SharedSessionBrowserProfileState{
		Backend:       "proxy",
		RuntimeTarget: "node",
	}, []SharedSessionBrowserProfileState{{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "reconnecting",
		Running:       true,
		Connected:     false,
		ObservedAt:    time.Now(),
	}})

	snapshot := registry.SnapshotSessionBrowserProfiles("s1")
	if len(snapshot) != 2 {
		t.Fatalf("expected route-scoped sync to keep host state and replace node route states, got %#v", snapshot)
	}
	if snapshot[0].Backend != "proxy" || snapshot[0].Profile != "isolated" || snapshot[1].Backend != "system" || snapshot[1].Profile != "default" {
		t.Fatalf("unexpected sync snapshot: %#v", snapshot)
	}
	if !snapshot[0].StatusSince.Equal(base) {
		t.Fatalf("expected sync to preserve status_since for unchanged health state, got %#v", snapshot[0])
	}
}

func TestBrowserSessionStateRegistrySyncSessionBrowserProfilesRespectsProfileFilter(t *testing.T) {
	registry := NewBrowserSessionStateRegistry()
	registry.RecordBrowserProfileState("s1", SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "stopped",
	})
	registry.RecordBrowserProfileState("s1", SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "relay",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "running",
		Running:       true,
		Connected:     true,
	})

	registry.SyncSessionBrowserProfiles("s1", SharedSessionBrowserProfileState{
		Backend:       "proxy",
		RuntimeTarget: "node",
		Profile:       "isolated",
	}, []SharedSessionBrowserProfileState{{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "running",
		Running:       true,
		Connected:     true,
	}})

	snapshot := registry.SnapshotSessionBrowserProfiles("s1")
	if len(snapshot) != 2 {
		t.Fatalf("expected filtered sync to preserve other profiles on the same route, got %#v", snapshot)
	}
	if snapshot[0].Profile != "isolated" || snapshot[0].Status != "running" || snapshot[1].Profile != "relay" || snapshot[1].Status != "running" {
		t.Fatalf("unexpected filtered sync snapshot: %#v", snapshot)
	}
}

func TestBrowserSessionStateRegistryResolveSessionBrowserProfileStatePrefersHealthMatchThenAliasFallback(t *testing.T) {
	registry := NewBrowserSessionStateRegistry()
	registry.RecordBrowserProfileState("s1", SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "reconnecting",
		Running:       true,
		Connected:     false,
	})
	registry.RecordBrowserProfileState("s1", SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chrome",
		Status:        "stopped",
		Running:       false,
		Connected:     false,
	})

	resolved, ok := registry.ResolveSessionBrowserProfileState("s1", SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		Status:        "reconnecting",
		Running:       true,
		Connected:     false,
	})
	if !ok || resolved.BrowserApp != "Chromium" {
		t.Fatalf("expected resolve to prefer reconnecting health match, got %#v ok=%v", resolved, ok)
	}

	resolved, ok = registry.ResolveSessionBrowserProfileState("s1", SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
	})
	if !ok || resolved.BrowserApp != "Chrome" {
		t.Fatalf("expected resolve without alias to fall back to stable sorted concrete browser app, got %#v ok=%v", resolved, ok)
	}
}

func TestBrowserSessionStateRegistryResolveSessionBrowserProfileStatusUsesSelectedRouteFallbacks(t *testing.T) {
	registry := NewBrowserSessionStateRegistry()
	registry.RecordBrowserProfileState("s1", SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "reconnecting",
		Running:       true,
		Connected:     false,
		Note:          "cdp reconnect in progress",
	})

	resolved, ok := registry.ResolveSessionBrowserProfileStatus("s1", BrowserRuntimeInfo{
		Backend: "proxy",
		Profile: "isolated",
		Target:  "node",
	}, "", BrowserProfileStatusResult{
		BrowserApp: "Chromium",
	})
	if !ok {
		t.Fatalf("expected resolve status to find scoped profile state")
	}
	if resolved.Backend != "proxy" || resolved.Profile != "isolated" || resolved.Status != "reconnecting" || !resolved.Running || resolved.Connected || resolved.Note != "cdp reconnect in progress" {
		t.Fatalf("unexpected resolved profile status: %#v", resolved)
	}
}

func TestPrepareManagedSessionBrowserProfileTransitionPreservesReconnectAndResetsTimedOutWatchdog(t *testing.T) {
	base := time.Now().Add(-2 * time.Minute)
	current := SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "reconnecting",
		Running:       true,
		Connected:     false,
		ObservedAt:    base,
		StatusSince:   base,
		Note:          "cdp reconnect in progress",
	}

	preserved := PrepareManagedSessionBrowserProfileTransition(current, SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		Status:        "running",
		Running:       true,
		Connected:     false,
	}, 30*time.Second)
	if preserved.Status != "reconnecting" || preserved.BrowserApp != "Chromium" || preserved.Note != "cdp reconnect in progress" {
		t.Fatalf("expected managed transition to preserve reconnecting state, got %#v", preserved)
	}

	reset := PrepareManagedSessionBrowserProfileTransition(current, SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		Status:        "reconnecting",
		Running:       true,
		Connected:     false,
		ObservedAt:    base.Add(3 * time.Minute),
	}, 30*time.Second)
	if reset.Status != "reconnecting" {
		t.Fatalf("expected reconnecting state to survive timeout reset, got %#v", reset)
	}
	if !reset.StatusSince.Equal(base.Add(3*time.Minute)) || !reset.ObservedAt.Equal(base.Add(3*time.Minute)) {
		t.Fatalf("expected timed-out reconnecting state to refresh with source observation time, got %#v", reset)
	}
}

func TestSharedSessionBrowserProfileStateFromStatusUsesSelectedRouteFallbacks(t *testing.T) {
	state := SharedSessionBrowserProfileStateFromStatus(BrowserRuntimeInfo{
		Backend: "proxy",
		Profile: "isolated",
		Target:  "node",
	}, BrowserProfileStatusResult{
		BrowserApp: "Chromium",
		Status:     "running",
		Running:    true,
		Connected:  true,
	})

	if state.Backend != "proxy" || state.Profile != "isolated" || state.RuntimeTarget != "node" {
		t.Fatalf("expected status mapping to inherit selected route identity, got %#v", state)
	}
	if state.BrowserApp != "Chromium" || state.Status != "running" || !state.Running || !state.Connected {
		t.Fatalf("unexpected status-derived state: %#v", state)
	}
}

func TestBrowserSessionStateRegistrySyncSessionBrowserProfileObservationReturnsFinalTransitionState(t *testing.T) {
	registry := NewBrowserSessionStateRegistry()
	base := time.Now().Add(-2 * time.Minute)
	registry.RecordBrowserProfileState("s1", SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "reconnecting",
		Running:       true,
		Connected:     false,
		ObservedAt:    base,
		StatusSince:   base,
		Note:          "cdp reconnect in progress",
	})

	synced, ok := registry.SyncSessionBrowserProfileObservation("s1", SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "running",
		Running:       true,
		Connected:     false,
	}, 54*time.Second)
	if !ok {
		t.Fatalf("expected sync observation result")
	}
	if synced.Status != "reconnecting" || synced.Note != "cdp reconnect in progress" {
		t.Fatalf("expected returned state to preserve managed reconnect transition, got %#v", synced)
	}
	if !synced.StatusSince.Equal(base) {
		t.Fatalf("expected returned state to preserve reconnecting status_since, got %#v", synced)
	}
}

func TestBrowserSessionStateRegistrySyncSessionBrowserProfileStatusObservationUsesRouteFallbacksAndManagedTransition(t *testing.T) {
	registry := NewBrowserSessionStateRegistry()
	base := time.Now().Add(-2 * time.Minute)
	registry.RecordBrowserProfileState("s1", SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "reconnecting",
		Running:       true,
		Connected:     false,
		ObservedAt:    base,
		StatusSince:   base,
		Note:          "cdp reconnect in progress",
	})

	synced, ok := registry.SyncSessionBrowserProfileStatusObservation("s1", BrowserRuntimeInfo{
		Backend: "proxy",
		Profile: "isolated",
		Target:  "node",
	}, BrowserProfileStatusResult{
		BrowserApp: "Chromium",
		Status:     "running",
		Running:    true,
		Connected:  false,
	}, base.Add(30*time.Second), 54*time.Second)
	if !ok {
		t.Fatalf("expected synced status observation result")
	}
	if synced.Backend != "proxy" || synced.Profile != "isolated" || synced.RuntimeTarget != "node" {
		t.Fatalf("expected route fallback identity in synced state, got %#v", synced)
	}
	if synced.Status != "reconnecting" || synced.Note != "cdp reconnect in progress" {
		t.Fatalf("expected managed transition to be preserved, got %#v", synced)
	}
	if !synced.ObservedAt.Equal(base.Add(30 * time.Second)) {
		t.Fatalf("expected synced state to preserve source observation time, got %#v", synced)
	}
}

func TestBrowserSessionStateRegistrySyncSessionBrowserProfileStatusResolutionReturnsResolvedStatus(t *testing.T) {
	registry := NewBrowserSessionStateRegistry()
	base := time.Now().Add(-2 * time.Minute)
	registry.RecordBrowserProfileState("s1", SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "reconnecting",
		Running:       true,
		Connected:     false,
		ObservedAt:    base,
		StatusSince:   base,
		Note:          "cdp reconnect in progress",
	})

	status, synced, ok := registry.SyncSessionBrowserProfileStatusResolution("s1", BrowserRuntimeInfo{
		Backend: "proxy",
		Profile: "isolated",
		Target:  "node",
	}, BrowserProfileStatusResult{
		BrowserApp: "Chromium",
		Status:     "running",
		Running:    true,
		Connected:  false,
	}, base.Add(30*time.Second), 54*time.Second)
	if !ok {
		t.Fatalf("expected resolved status observation")
	}
	if status.Profile != "isolated" || status.Status != "reconnecting" || status.Note != "cdp reconnect in progress" {
		t.Fatalf("expected lifecycle-owned resolved status, got %#v", status)
	}
	if synced.Profile != "isolated" || synced.RuntimeTarget != "node" || synced.Status != "reconnecting" {
		t.Fatalf("unexpected synced state: %#v", synced)
	}
}

func TestSharedSessionBrowserProfileStatesFromProfilesUsesSelectedRouteFallbacks(t *testing.T) {
	states := SharedSessionBrowserProfileStatesFromProfiles(BrowserRuntimeInfo{
		Backend: "proxy",
		Target:  "node",
	}, BrowserProfilesResult{
		Profiles: []BrowserProfileInfo{
			{Profile: "isolated", BrowserApp: "Chromium", Status: "running", Running: true, Connected: true},
		},
	})
	if len(states) != 1 {
		t.Fatalf("expected one mapped profile state, got %#v", states)
	}
	if states[0].Backend != "proxy" || states[0].RuntimeTarget != "node" || states[0].Profile != "isolated" {
		t.Fatalf("expected profile mapping to inherit selected route, got %#v", states[0])
	}
}

func TestBrowserSessionStateRegistrySyncSessionBrowserProfilesObservationReturnsScopedSyncedStates(t *testing.T) {
	registry := NewBrowserSessionStateRegistry()
	base := time.Now().Add(-2 * time.Minute)
	registry.RecordBrowserProfileState("s1", SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "reconnecting",
		Running:       true,
		Connected:     false,
		ObservedAt:    base,
		StatusSince:   base,
		Note:          "cdp reconnect in progress",
	})
	registry.RecordBrowserProfileState("s1", SharedSessionBrowserProfileState{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Status:        "running",
		Running:       true,
		Connected:     true,
	})

	synced := registry.SyncSessionBrowserProfilesObservation("s1", BrowserRuntimeInfo{
		Backend: "proxy",
		Target:  "node",
	}, "", BrowserProfilesResult{
		Backend: "proxy",
		Profiles: []BrowserProfileInfo{
			{Profile: "isolated", BrowserApp: "Chromium", Status: "running", Running: true, Connected: false},
			{Profile: "relay", BrowserApp: "Chromium", Status: "stopped", Running: false, Connected: false},
		},
	}, base.Add(30*time.Second), 54*time.Second)

	if len(synced) != 2 {
		t.Fatalf("expected synced scoped route profiles, got %#v", synced)
	}
	if synced[0].Profile != "isolated" || synced[0].Status != "reconnecting" || synced[0].Note != "cdp reconnect in progress" {
		t.Fatalf("expected scoped sync to preserve managed transition, got %#v", synced)
	}
	if synced[1].Profile != "relay" || synced[1].Status != "stopped" {
		t.Fatalf("expected scoped sync to include observed sibling profile, got %#v", synced)
	}

	snapshot := registry.SnapshotSessionBrowserProfiles("s1")
	if len(snapshot) != 3 {
		t.Fatalf("expected full snapshot to retain host route alongside synced node scope, got %#v", snapshot)
	}
}

func TestBrowserSessionStateRegistrySyncSessionBrowserProfilesResolutionReturnsScopedSnapshot(t *testing.T) {
	registry := NewBrowserSessionStateRegistry()
	base := time.Now().Add(-2 * time.Minute)
	registry.RecordBrowserProfileState("s1", SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "reconnecting",
		Running:       true,
		Connected:     false,
		ObservedAt:    base,
		StatusSince:   base,
		Note:          "cdp reconnect in progress",
	})

	snapshot := registry.SyncSessionBrowserProfilesResolution("s1", BrowserRuntimeInfo{
		Backend: "proxy",
		Target:  "node",
	}, "isolated", BrowserProfilesResult{
		Backend: "proxy",
		Profiles: []BrowserProfileInfo{
			{Profile: "isolated", BrowserApp: "Chromium", Status: "running", Running: true, Connected: false},
		},
	}, base.Add(30*time.Second), 54*time.Second)

	if len(snapshot) != 1 {
		t.Fatalf("expected synced scoped snapshot, got %#v", snapshot)
	}
	if snapshot[0].Profile != "isolated" || snapshot[0].Status != "reconnecting" || snapshot[0].Note != "cdp reconnect in progress" {
		t.Fatalf("expected profiles resolution to preserve managed transition, got %#v", snapshot[0])
	}
}

func TestBrowserSessionStateRegistrySnapshotSessionBrowserProfilesForScopeFiltersRouteAndProfile(t *testing.T) {
	registry := NewBrowserSessionStateRegistry()
	registry.RecordBrowserProfileState("s1", SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "running",
		Running:       true,
		Connected:     true,
	})
	registry.RecordBrowserProfileState("s1", SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "relay",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "stopped",
	})
	registry.RecordBrowserProfileState("s1", SharedSessionBrowserProfileState{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Status:        "running",
		Running:       true,
		Connected:     true,
	})

	scoped := registry.SnapshotSessionBrowserProfilesForScope("s1", BrowserRuntimeInfo{
		Backend: "proxy",
		Target:  "node",
	}, "isolated")
	if len(scoped) != 1 || scoped[0].Profile != "isolated" || scoped[0].RuntimeTarget != "node" {
		t.Fatalf("expected scoped snapshot to filter route/profile, got %#v", scoped)
	}
}

func TestSharedSessionBrowserProfileStateFromLifecycleMapsRestartAndStopDecisions(t *testing.T) {
	restarting, ok := SharedSessionBrowserProfileStateFromLifecycle(BrowserRuntimeInfo{
		Backend: "proxy",
		Target:  "node",
	}, "isolated", BrowserProfileStatusResult{
		Backend:    "proxy",
		Profile:    "isolated",
		BrowserApp: "Chromium",
		Status:     "starting",
		Running:    true,
		Connected:  false,
	}, "restart_started")
	if !ok || restarting.Status != "reconnecting" || restarting.Note != "restart requested" {
		t.Fatalf("expected restart decision to map to reconnecting state, got %#v ok=%v", restarting, ok)
	}

	reconnectInProgress, ok := SharedSessionBrowserProfileStateFromLifecycle(BrowserRuntimeInfo{
		Backend: "proxy",
		Target:  "node",
	}, "isolated", BrowserProfileStatusResult{
		Backend:    "proxy",
		Profile:    "isolated",
		BrowserApp: "Chromium",
		Status:     "reconnecting",
		Running:    true,
		Connected:  false,
		Note:       "cdp reconnect in progress",
	}, "restart_reconnect_in_progress")
	if !ok || reconnectInProgress.Status != "reconnecting" || !reconnectInProgress.Running || reconnectInProgress.Connected {
		t.Fatalf("expected reconnect-in-progress decision to preserve lifecycle state, got %#v ok=%v", reconnectInProgress, ok)
	}

	alreadyReady, ok := SharedSessionBrowserProfileStateFromLifecycle(BrowserRuntimeInfo{
		Backend: "proxy",
		Target:  "node",
	}, "isolated", BrowserProfileStatusResult{
		Backend:    "proxy",
		Profile:    "isolated",
		BrowserApp: "Chromium",
		Status:     "connected",
		Running:    true,
		Connected:  true,
	}, "already_ready")
	if !ok || alreadyReady.Status != "connected" || !alreadyReady.Running || !alreadyReady.Connected {
		t.Fatalf("expected already_ready decision to preserve healthy lifecycle state, got %#v ok=%v", alreadyReady, ok)
	}

	stopped, ok := SharedSessionBrowserProfileStateFromLifecycle(BrowserRuntimeInfo{
		Backend: "proxy",
		Target:  "node",
	}, "isolated", BrowserProfileStatusResult{
		Backend:    "proxy",
		Profile:    "isolated",
		BrowserApp: "Chromium",
		Status:     "running",
		Running:    true,
		Connected:  true,
	}, "stopped")
	if !ok || stopped.Status != "stopped" || stopped.Running || stopped.Connected {
		t.Fatalf("expected stop decision to map to stopped state, got %#v ok=%v", stopped, ok)
	}
}

func TestBrowserSessionStateRegistrySyncSessionBrowserProfileLifecycleObservationReturnsSyncedState(t *testing.T) {
	registry := NewBrowserSessionStateRegistry()
	base := time.Now().Add(-2 * time.Minute)
	registry.RecordBrowserProfileState("s1", SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "reconnecting",
		Running:       true,
		Connected:     false,
		ObservedAt:    base,
		StatusSince:   base,
		Note:          "cdp reconnect in progress",
	})

	synced, ok := registry.SyncSessionBrowserProfileLifecycleObservation("s1", BrowserRuntimeInfo{
		Backend: "proxy",
		Target:  "node",
	}, "isolated", BrowserProfileStatusResult{
		Backend:    "proxy",
		Profile:    "isolated",
		BrowserApp: "Chromium",
		Status:     "starting",
		Running:    true,
		Connected:  false,
	}, "restart_started", base.Add(30*time.Second), 54*time.Second)
	if !ok {
		t.Fatalf("expected synced lifecycle observation result")
	}
	if synced.Backend != "proxy" || synced.Profile != "isolated" || synced.RuntimeTarget != "node" {
		t.Fatalf("expected route fallback identity in synced lifecycle state, got %#v", synced)
	}
	if synced.Status != "reconnecting" || !synced.Running || synced.Connected {
		t.Fatalf("expected lifecycle sync to return reconnecting synced state, got %#v", synced)
	}
}

func TestBrowserSessionStateRegistrySyncSessionBrowserProfileLifecycleResolutionReturnsResolvedStatus(t *testing.T) {
	registry := NewBrowserSessionStateRegistry()
	status, synced, ok := registry.SyncSessionBrowserProfileLifecycleResolution("s1", BrowserRuntimeInfo{
		Backend: "proxy",
		Target:  "node",
	}, "isolated", BrowserProfileStatusResult{
		Backend:    "proxy",
		Profile:    "isolated",
		BrowserApp: "Chromium",
		Status:     "starting",
		Running:    true,
		Connected:  false,
	}, "restart_started", time.Now(), 54*time.Second)
	if !ok {
		t.Fatalf("expected resolved lifecycle observation")
	}
	if status.Profile != "isolated" || status.Status != "reconnecting" || status.Note != "restart requested" {
		t.Fatalf("expected lifecycle-owned resolved status, got %#v", status)
	}
	if synced.Profile != "isolated" || synced.RuntimeTarget != "node" || synced.Status != "reconnecting" {
		t.Fatalf("unexpected synced lifecycle state: %#v", synced)
	}
}

func TestBrowserSessionStateRegistrySyncSessionBrowserExecutionObservationsReturnsSyncedScope(t *testing.T) {
	registry := NewBrowserSessionStateRegistry()
	synced, ok, snapshot := registry.SyncSessionBrowserExecutionObservations("s1", BrowserRuntimeInfo{
		Backend: "proxy",
		Target:  "node",
	}, "", "isolated", &BrowserProfilesResult{
		Backend: "proxy",
		Profiles: []BrowserProfileInfo{
			{Profile: "isolated", BrowserApp: "Chromium", Status: "running", Running: true},
			{Profile: "relay", BrowserApp: "Chromium", Status: "stopped"},
		},
	}, time.Now(), BrowserProfileStatusResult{
		Backend: "proxy",
		Profile: "isolated",
		Status:  "started",
		Running: true,
	}, time.Now(), "started", 54*time.Second)
	if !ok {
		t.Fatalf("expected synced execution observation state")
	}
	if synced.Profile != "isolated" || synced.RuntimeTarget != "node" || synced.Status != "starting" {
		t.Fatalf("unexpected synced execution state: %#v", synced)
	}
	if len(snapshot) != 2 {
		t.Fatalf("expected scoped synced snapshot, got %#v", snapshot)
	}
	if snapshot[0].Profile != "isolated" || snapshot[0].Status != "starting" {
		t.Fatalf("expected active profile to reflect lifecycle-owned state, got %#v", snapshot)
	}
	if snapshot[1].Profile != "relay" || snapshot[1].Status != "stopped" {
		t.Fatalf("expected companion profile to remain in scoped snapshot, got %#v", snapshot)
	}
}

func TestBrowserSessionStateRegistrySyncSessionBrowserStatusAndProfilesObservationsReturnsSyncedScope(t *testing.T) {
	registry := NewBrowserSessionStateRegistry()
	base := time.Now().Add(-2 * time.Minute)
	registry.RecordBrowserProfileState("s1", SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "reconnecting",
		Running:       true,
		Connected:     false,
		ObservedAt:    base,
		StatusSince:   base,
		Note:          "cdp reconnect in progress",
	})
	synced, ok, snapshot := registry.SyncSessionBrowserStatusAndProfilesObservations("s1", BrowserRuntimeInfo{
		Backend: "proxy",
		Target:  "node",
	}, "", &BrowserProfileStatusResult{
		Backend:    "proxy",
		Profile:    "isolated",
		BrowserApp: "Chromium",
		Status:     "running",
		Running:    true,
		Connected:  false,
		Note:       "cdp reconnect in progress",
	}, base.Add(30*time.Second), &BrowserProfilesResult{
		Backend: "proxy",
		Profiles: []BrowserProfileInfo{
			{Profile: "isolated", BrowserApp: "Chromium", Status: "running", Running: true, Connected: false, Note: "cdp reconnect in progress"},
			{Profile: "relay", BrowserApp: "Chromium", Status: "stopped"},
		},
	}, base.Add(30*time.Second), 54*time.Second)
	if !ok {
		t.Fatalf("expected synced status observation result")
	}
	if synced.Profile != "isolated" || synced.RuntimeTarget != "node" || synced.Status != "reconnecting" {
		t.Fatalf("unexpected synced status state: %#v", synced)
	}
	if len(snapshot) != 2 {
		t.Fatalf("expected scoped synced snapshot, got %#v", snapshot)
	}
	if snapshot[0].Profile != "isolated" || snapshot[0].Status != "reconnecting" {
		t.Fatalf("expected active profile to preserve managed transition, got %#v", snapshot)
	}
	if snapshot[1].Profile != "relay" || snapshot[1].Status != "stopped" {
		t.Fatalf("expected companion profile to remain available, got %#v", snapshot)
	}
}

func TestBrowserSessionStateRegistrySyncSessionBrowserStatusAndProfilesResolutionReturnsResolvedStatus(t *testing.T) {
	registry := NewBrowserSessionStateRegistry()
	base := time.Now().Add(-2 * time.Minute)
	registry.RecordBrowserProfileState("s1", SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "reconnecting",
		Running:       true,
		Connected:     false,
		ObservedAt:    base,
		StatusSince:   base,
		Note:          "cdp reconnect in progress",
	})
	status, synced, ok, snapshot := registry.SyncSessionBrowserStatusAndProfilesResolution("s1", BrowserRuntimeInfo{
		Backend: "proxy",
		Target:  "node",
	}, "", &BrowserProfileStatusResult{
		Backend:    "proxy",
		Profile:    "isolated",
		BrowserApp: "Chromium",
		Status:     "running",
		Running:    true,
		Connected:  false,
		Note:       "cdp reconnect in progress",
	}, base.Add(30*time.Second), &BrowserProfilesResult{
		Backend: "proxy",
		Profiles: []BrowserProfileInfo{
			{Profile: "isolated", BrowserApp: "Chromium", Status: "running", Running: true, Connected: false, Note: "cdp reconnect in progress"},
			{Profile: "relay", BrowserApp: "Chromium", Status: "stopped"},
		},
	}, base.Add(30*time.Second), 54*time.Second)
	if !ok {
		t.Fatalf("expected resolved status result from synced observation")
	}
	if status.Profile != "isolated" || status.Status != "reconnecting" {
		t.Fatalf("expected resolved lifecycle-owned status, got %#v", status)
	}
	if synced.Profile != "isolated" || synced.Status != "reconnecting" {
		t.Fatalf("unexpected synced state: %#v", synced)
	}
	if len(snapshot) != 2 {
		t.Fatalf("expected scoped synced snapshot, got %#v", snapshot)
	}
}

func TestBrowserSessionStateRegistrySyncSessionBrowserExecutionResolutionReturnsResolvedStatus(t *testing.T) {
	registry := NewBrowserSessionStateRegistry()
	status, synced, ok, snapshot := registry.SyncSessionBrowserExecutionResolution("s1", BrowserRuntimeInfo{
		Backend: "proxy",
		Target:  "node",
	}, "", "isolated", &BrowserProfilesResult{
		Backend: "proxy",
		Profiles: []BrowserProfileInfo{
			{Profile: "isolated", BrowserApp: "Chromium", Status: "running", Running: true},
			{Profile: "relay", BrowserApp: "Chromium", Status: "stopped"},
		},
	}, time.Now(), BrowserProfileStatusResult{
		Backend: "proxy",
		Profile: "isolated",
		Status:  "started",
		Running: true,
	}, time.Now(), "started", 54*time.Second)
	if !ok {
		t.Fatalf("expected resolved execution status from synced observation")
	}
	if status.Profile != "isolated" || status.Status != "starting" {
		t.Fatalf("expected lifecycle-owned execution status, got %#v", status)
	}
	if synced.Profile != "isolated" || synced.Status != "starting" {
		t.Fatalf("unexpected synced execution state: %#v", synced)
	}
	if len(snapshot) != 2 {
		t.Fatalf("expected scoped synced snapshot, got %#v", snapshot)
	}
}

func TestBrowserSessionStateRegistrySyncSessionBrowserProfileStateReplacesScopeAndPreservesManagedTransition(t *testing.T) {
	registry := NewBrowserSessionStateRegistry()
	base := time.Now().Add(-12 * time.Second)
	registry.RecordBrowserProfileState("s1", SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "reconnecting",
		Running:       true,
		Connected:     false,
		ObservedAt:    base,
		StatusSince:   base,
		Note:          "restart requested",
	})
	registry.RecordBrowserProfileState("s1", SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chrome",
		Status:        "stopped",
	})

	registry.SyncSessionBrowserProfileState("s1", SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "running",
		Running:       true,
		Connected:     false,
	}, 30*time.Second)

	snapshot := registry.SnapshotSessionBrowserProfiles("s1")
	if len(snapshot) != 1 {
		t.Fatalf("expected scoped sync to collapse duplicate aliases, got %#v", snapshot)
	}
	if snapshot[0].BrowserApp != "Chromium" || snapshot[0].Status != "reconnecting" || snapshot[0].Note != "restart requested" {
		t.Fatalf("expected managed transition to be preserved during scoped sync, got %#v", snapshot[0])
	}
	if !snapshot[0].StatusSince.Equal(base) {
		t.Fatalf("expected reconnecting status_since to be preserved during scoped sync, got %#v", snapshot[0])
	}
}
