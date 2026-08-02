package browserruntime

import (
	"context"
	"strings"
	"testing"
	"time"
)

type runtimeInfoStatusProfilesObservationTestBackend struct {
	*statusProfilesObservationTestBackend
	runtimeInfo BrowserRuntimeInfo
}

func (b *runtimeInfoStatusProfilesObservationTestBackend) BrowserRuntimeInfo() BrowserRuntimeInfo {
	return b.runtimeInfo
}

func TestResolveSharedSessionBrowserProfileStatusEventFallsBackWithoutRegistry(t *testing.T) {
	resolved, synced, ok := ResolveSharedSessionBrowserProfileStatusEvent(
		nil,
		"",
		BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		BrowserProfileStatusResult{
			Backend:    "proxy",
			Profile:    "isolated",
			BrowserApp: "Chromium",
			Status:     "running",
			Running:    true,
			Connected:  true,
		},
		time.Date(2026, time.March, 29, 9, 0, 0, 0, time.UTC),
		54*time.Second,
	)
	if ok {
		t.Fatalf("expected fallback resolution without synced registry state")
	}
	if resolved.Profile != "isolated" || resolved.Status != "running" {
		t.Fatalf("expected raw status fallback, got %#v", resolved)
	}
	if synced.Profile != "isolated" || synced.RuntimeTarget != "node" || synced.Status != "running" || !synced.Connected {
		t.Fatalf("expected fallback shared state from raw status, got %#v", synced)
	}
}

func TestResolveSharedSessionBrowserProfilesEventFallsBackToObservedSnapshotWithoutRegistry(t *testing.T) {
	snapshot := ResolveSharedSessionBrowserProfilesEvent(
		nil,
		"",
		BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
		"",
		BrowserProfilesResult{
			Backend: "proxy",
			Profiles: []BrowserProfileInfo{
				{Profile: "isolated", BrowserApp: "Chromium", Status: "running", Running: true, Connected: true},
			},
		},
		time.Date(2026, time.March, 29, 9, 5, 0, 0, time.UTC),
		54*time.Second,
	)
	if len(snapshot) != 1 {
		t.Fatalf("expected raw observed snapshot fallback, got %#v", snapshot)
	}
	if snapshot[0].Backend != "proxy" || snapshot[0].Profile != "isolated" || snapshot[0].RuntimeTarget != "node" || snapshot[0].BrowserApp != "Chromium" {
		t.Fatalf("unexpected fallback snapshot: %#v", snapshot[0])
	}
}

func TestResolveSharedSessionBrowserStatusAndProfilesEventUsesRegistryResolution(t *testing.T) {
	registry := NewBrowserSessionStateRegistry()
	sessionID := "sess-status-profiles-resolution"

	registry.RecordBrowserProfileState(sessionID, SharedSessionBrowserProfileState{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Status:        "running",
		Running:       true,
		Connected:     true,
	})

	resolved, synced, ok, snapshot := ResolveSharedSessionBrowserStatusAndProfilesEvent(
		registry,
		sessionID,
		BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
		"",
		&BrowserProfileStatusResult{
			Backend:    "proxy",
			Profile:    "isolated",
			BrowserApp: "Chromium",
			Status:     "running",
			Running:    true,
			Connected:  true,
		},
		time.Date(2026, time.March, 29, 9, 10, 0, 0, time.UTC),
		&BrowserProfilesResult{
			Backend: "proxy",
			Profiles: []BrowserProfileInfo{
				{Profile: "isolated", BrowserApp: "Chromium", Status: "running", Running: true, Connected: true},
			},
		},
		time.Date(2026, time.March, 29, 9, 10, 0, 0, time.UTC),
		54*time.Second,
	)
	if !ok {
		t.Fatalf("expected synced scoped state")
	}
	if resolved.Profile != "isolated" || resolved.Status != "running" {
		t.Fatalf("expected lifecycle-owned status, got %#v", resolved)
	}
	if synced.Profile != "isolated" || synced.RuntimeTarget != "node" || synced.BrowserApp != "Chromium" {
		t.Fatalf("unexpected synced state: %#v", synced)
	}
	if len(snapshot) != 1 || snapshot[0].Backend != "proxy" || snapshot[0].Profile != "isolated" || snapshot[0].RuntimeTarget != "node" {
		t.Fatalf("expected scoped synced snapshot, got %#v", snapshot)
	}
	fullSnapshot := registry.SnapshotSessionBrowserProfiles(sessionID)
	if len(fullSnapshot) != 2 {
		t.Fatalf("expected registry full snapshot to preserve other routes, got %#v", fullSnapshot)
	}
	if fullSnapshot[1].Backend != "system" || fullSnapshot[1].Profile != "default" || fullSnapshot[1].RuntimeTarget != "host" {
		t.Fatalf("expected host route state to remain in registry snapshot, got %#v", fullSnapshot)
	}
}

func TestResolveSharedSessionBrowserExecutionEventUsesRegistryResolution(t *testing.T) {
	registry := NewBrowserSessionStateRegistry()
	sessionID := "sess-execution-resolution"

	resolution := ResolveSharedSessionBrowserExecutionEvent(
		registry,
		sessionID,
		BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
		"",
		SharedSessionBrowserExecutionResult{
			Profile: "isolated",
			Profiles: &BrowserProfilesResult{
				Backend:        "proxy",
				DefaultProfile: "isolated",
				Profiles: []BrowserProfileInfo{
					{Profile: "isolated", BrowserApp: "Chromium", Status: "running", Running: true, Connected: false},
					{Profile: "relay", BrowserApp: "Chromium", Status: "stopped", Running: false, Connected: false},
				},
			},
			Decision: "started",
			ProfileStatus: BrowserProfileStatusResult{
				Backend:   "proxy",
				Profile:   "isolated",
				Status:    "started",
				Running:   true,
				Connected: false,
			},
		},
		time.Minute,
	)
	if !resolution.HasSyncedState {
		t.Fatalf("expected registry-backed execution resolution to return synced state")
	}
	if resolution.ResolvedStatus.Profile != "isolated" || resolution.ResolvedStatus.Status != "starting" || !resolution.ResolvedStatus.Running || resolution.ResolvedStatus.Connected {
		t.Fatalf("expected resolved status to use lifecycle-owned starting state, got %#v", resolution.ResolvedStatus)
	}
	if len(resolution.Snapshot) != 2 || resolution.Snapshot[0].Profile != "isolated" || resolution.Snapshot[0].Status != "starting" {
		t.Fatalf("expected registry-backed execution resolution to return scoped snapshot, got %#v", resolution.Snapshot)
	}
}

func TestResolveSharedSessionBrowserExecutionEventFallsBackWithoutRegistry(t *testing.T) {
	resolution := ResolveSharedSessionBrowserExecutionEvent(
		nil,
		"",
		BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		"",
		SharedSessionBrowserExecutionResult{
			Profile: "isolated",
			Profiles: &BrowserProfilesResult{
				Backend:        "proxy",
				DefaultProfile: "isolated",
				Profiles: []BrowserProfileInfo{
					{Profile: "isolated", BrowserApp: "Chromium", Status: "running", Running: true, Connected: false},
				},
			},
			Decision: "started",
			ProfileStatus: BrowserProfileStatusResult{
				Backend:   "proxy",
				Profile:   "isolated",
				Status:    "started",
				Running:   true,
				Connected: false,
			},
		},
		time.Minute,
	)
	if !resolution.HasSyncedState {
		t.Fatalf("expected lifecycle fallback resolution to produce synced state")
	}
	if resolution.ResolvedStatus.Profile != "isolated" || resolution.ResolvedStatus.Status != "starting" || !resolution.ResolvedStatus.Running || resolution.ResolvedStatus.Connected {
		t.Fatalf("expected fallback resolution to use lifecycle-owned starting state, got %#v", resolution.ResolvedStatus)
	}
	if len(resolution.Snapshot) != 1 || resolution.Snapshot[0].Profile != "isolated" || resolution.Snapshot[0].Status != "running" {
		t.Fatalf("expected fallback resolution to preserve raw discovered snapshot, got %#v", resolution.Snapshot)
	}
}

func TestSyncSharedSessionBrowserProfilesEventSyncsFullRouteScopeWhenProfileFilterBlank(t *testing.T) {
	registry := NewBrowserSessionStateRegistry()
	sessionID := "sess-profiles-sync"

	registry.RecordBrowserProfileState(sessionID, SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "relay",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "disconnected",
		Running:       true,
		Connected:     false,
		Note:          "stale relay state",
	})
	registry.RecordBrowserProfileState(sessionID, SharedSessionBrowserProfileState{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Status:        "running",
		Running:       true,
		Connected:     true,
	})

	scopedSnapshot := SyncSharedSessionBrowserProfilesEvent(
		registry,
		sessionID,
		BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
		"",
		BrowserProfilesResult{
			Backend: "proxy",
			Profiles: []BrowserProfileInfo{
				{Profile: "isolated", BrowserApp: "Chromium", Status: "running", Running: true, Connected: true},
			},
		},
		time.Time{},
		54*time.Second,
	)
	if len(scopedSnapshot) != 1 || scopedSnapshot[0].Profile != "isolated" || scopedSnapshot[0].RuntimeTarget != "node" {
		t.Fatalf("expected scoped synced snapshot, got %#v", scopedSnapshot)
	}

	snapshot := registry.SnapshotSessionBrowserProfiles(sessionID)
	if len(snapshot) != 2 {
		t.Fatalf("expected full-scope sync to keep host state and replace node route state, got %#v", snapshot)
	}
	if snapshot[0].Backend != "proxy" || snapshot[0].Profile != "isolated" || snapshot[1].Backend != "system" || snapshot[1].Profile != "default" {
		t.Fatalf("unexpected full-scope synced snapshot: %#v", snapshot)
	}
}

func TestSyncSharedSessionBrowserProfileStatusEventInvalidatesManagedCurrentTargetSelection(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	sessionID := "sess-status-sync"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node"}

	sessionRegistry.TrackTab(sessionID, BrowserSessionTarget{
		TabIndex:   1,
		URL:        "https://node.example/home",
		Title:      "Home",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
	}, true)
	if _, ok := sessionRegistry.CurrentTargetForRoute(sessionID, route); !ok {
		t.Fatalf("expected initial managed current target selection")
	}

	state, ok := SyncSharedSessionBrowserProfileStatusEvent(
		stateRegistry,
		sessionRegistry,
		sessionID,
		BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		BrowserProfileStatusResult{
			Backend:    "proxy",
			Profile:    "isolated",
			BrowserApp: "Chromium",
			Status:     "disconnected",
			Running:    true,
			Connected:  false,
		},
		time.Time{},
		54*time.Second,
	)
	if !ok || state.Status != "disconnected" {
		t.Fatalf("expected synced disconnected state, got %#v ok=%v", state, ok)
	}
	if _, ok := sessionRegistry.CurrentTargetForRoute(sessionID, route); ok {
		t.Fatalf("expected managed current target selection to be invalidated")
	}
	if _, ok := sessionRegistry.ResolveTabForRoute(sessionID, route, 1); !ok {
		t.Fatalf("expected tracked tab to remain after selection invalidation")
	}
}

func TestSyncSharedSessionBrowserProfileLifecycleEventSyncsRouteProfileScope(t *testing.T) {
	registry := NewBrowserSessionStateRegistry()
	sessionID := "sess-lifecycle-sync"
	base := time.Now().Add(-12 * time.Second)

	registry.RecordBrowserProfileState(sessionID, SharedSessionBrowserProfileState{
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
	registry.RecordBrowserProfileState(sessionID, SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chrome",
		Status:        "stopped",
	})

	state, ok := SyncSharedSessionBrowserProfileLifecycleEvent(
		registry,
		nil,
		nil,
		sessionID,
		BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
		"isolated",
		BrowserProfileStatusResult{
			Backend:   "proxy",
			Profile:   "isolated",
			Status:    "starting",
			Running:   true,
			Connected: false,
		},
		"restart_started",
		time.Time{},
		54*time.Second,
	)
	if !ok || state.BrowserApp != "Chromium" || state.Status != "reconnecting" || state.Note != "restart requested" {
		t.Fatalf("expected lifecycle sync to preserve managed reconnecting state, got %#v ok=%v", state, ok)
	}

	snapshot := registry.SnapshotSessionBrowserProfiles(sessionID)
	if len(snapshot) != 1 {
		t.Fatalf("expected lifecycle sync to collapse stale duplicate profile states, got %#v", snapshot)
	}
	if snapshot[0].BrowserApp != "Chromium" || snapshot[0].Status != "reconnecting" || snapshot[0].Note != "restart requested" {
		t.Fatalf("unexpected lifecycle-synced profile state: %#v", snapshot[0])
	}
	if !snapshot[0].StatusSince.Equal(base) {
		t.Fatalf("expected lifecycle sync to preserve status_since for unchanged reconnecting health, got %#v", snapshot[0])
	}
}

func TestSyncSharedSessionBrowserProfileStatusEventSeedsSharedWatchManagerRawStatusCache(t *testing.T) {
	backend := &statusProfilesObservationTestBackend{
		statusResp: BrowserProfileStatusResult{
			Backend:   "proxy",
			Profile:   "isolated",
			Status:    "stopped",
			Running:   false,
			Connected: false,
		},
	}
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	manager := SharedSessionBrowserObserverManagerFor(sessionRegistry, nil, stateRegistry, time.Minute)
	bound := manager.Bind(backend)
	sessionID := "sess-status-seed"
	observedAt := time.Date(2026, time.March, 30, 7, 0, 0, 0, time.UTC)

	initial := bound.ObserveRawStatus(context.Background(), "isolated")
	if initial.Status == nil || initial.Status.Status != "stopped" {
		t.Fatalf("expected initial raw status poll to reflect backend result, got %#v", initial)
	}
	if len(backend.statusReqs) != 1 {
		t.Fatalf("expected initial raw status observation to poll backend once, got %#v", backend.statusReqs)
	}

	state, ok := SyncSharedSessionBrowserProfileStatusEvent(
		stateRegistry,
		sessionRegistry,
		sessionID,
		BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		BrowserProfileStatusResult{
			Backend:    "proxy",
			Profile:    "isolated",
			BrowserApp: "Chromium",
			Status:     "running",
			Running:    true,
			Connected:  true,
		},
		observedAt,
		time.Minute,
	)
	if !ok || state.Status != "running" {
		t.Fatalf("expected synced running state, got %#v ok=%v", state, ok)
	}

	seeded := bound.ObserveRawStatus(context.Background(), "isolated")
	if seeded.Status == nil || seeded.Status.Status != "running" || !seeded.Status.Connected {
		t.Fatalf("expected seeded raw status observation, got %#v", seeded)
	}
	if !seeded.ObservedAt.Equal(observedAt) {
		t.Fatalf("expected seeded raw status timestamp %v, got %v", observedAt, seeded.ObservedAt)
	}
	if len(backend.statusReqs) != 1 {
		t.Fatalf("expected event-synced raw status cache to avoid a second poll, got %#v", backend.statusReqs)
	}
}

func TestSharedSessionBrowserObserverManagerSyncProfileStatusEventSeedsWatchLoopFromCachedEventCycleWhenProfilesDrained(t *testing.T) {
	backend := &statusProfilesObservationTestBackend{
		statusResp: BrowserProfileStatusResult{
			Backend:   "proxy",
			Profile:   "isolated",
			Status:    "stopped",
			Running:   false,
			Connected: false,
		},
		profilesResp: BrowserProfilesResult{
			Backend:        "proxy",
			DefaultProfile: "isolated",
			Profiles: []BrowserProfileInfo{{
				Profile:    "isolated",
				BrowserApp: "Chromium",
				Status:     "stopped",
				Running:    false,
				Connected:  false,
			}},
		},
	}
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	sessionID := "sess-direct-status-watchloop-refresh"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}
	sessionRegistry.TrackTab(sessionID, BrowserSessionTarget{
		TabIndex:   1,
		URL:        "https://example.com/one",
		Title:      "One",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
	}, true)
	stateRegistry.SelectBrowserProfile(sessionID, SharedSessionBrowserProfileSelection{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Source:        "remember_profile",
	})
	runRegistry := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			sessionID: {{RunID: "run-1", Status: "running"}},
		}},
	}
	manager := SharedSessionBrowserObserverManagerFor(sessionRegistry, runRegistry, stateRegistry, time.Minute)
	bound := manager.Bind(backend)
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

	initial := bound.ObserveWatchLoop(context.Background(), req)
	if len(initial.View.Session.Routes) != 1 || len(initial.View.Session.Routes[0].Targets) != 1 || initial.View.Session.Routes[0].Targets[0].Title != "One" {
		t.Fatalf("expected initial watch loop to expose tracked route snapshot, got %#v", initial.View.Session.Routes)
	}
	if len(initial.Watch.Profiles) != 1 || initial.Watch.Profiles[0].State.Status != "stopped" {
		t.Fatalf("expected initial watch loop to reflect stopped profile, got %#v", initial.Watch.Profiles)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected initial watch loop to poll backend once, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
	if runRegistry.callCount() != 1 {
		t.Fatalf("expected initial watch loop to snapshot runs once, got %d", runRegistry.callCount())
	}
	initialProfilesObservedAt := initial.Cycle.Observation.ProfilesObservedAt

	bound.state.mu.Lock()
	clear(bound.state.rawStatus)
	clear(bound.state.rawProfiles)
	eventCycleCount := len(bound.state.eventCycles)
	watchLoopCount := len(bound.state.watchLoops)
	bound.state.mu.Unlock()
	if eventCycleCount == 0 || watchLoopCount == 0 {
		t.Fatalf("expected cached event-cycle/watch-loop source before draining raw caches, got eventCycles=%d watchLoops=%d", eventCycleCount, watchLoopCount)
	}

	observedAt := time.Date(2026, time.March, 30, 7, 2, 0, 0, time.UTC)
	state, ok := manager.SyncProfileStatusEvent(
		sessionID,
		BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		BrowserProfileStatusResult{
			Backend:    "proxy",
			Profile:    "isolated",
			BrowserApp: "Chromium",
			Status:     "reconnecting",
			Running:    true,
			Connected:  false,
		},
		observedAt,
	)
	if !ok || state.Status != "reconnecting" {
		t.Fatalf("expected reconnecting synced state, got %#v ok=%v", state, ok)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected status-only sync to reuse cached event-cycle source, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
	if runRegistry.callCount() != 2 {
		t.Fatalf("expected status-only sync to refresh watch-loop projection once, got %d", runRegistry.callCount())
	}

	seeded := bound.ObserveWatchLoop(context.Background(), req)
	if len(seeded.View.Session.Routes) != 1 || seeded.View.Session.Routes[0].CurrentTargetID != "" {
		t.Fatalf("expected status-only sync to invalidate current target in refreshed watch loop, got %#v", seeded.View.Session.Routes)
	}
	if len(seeded.Watch.Profiles) != 1 || seeded.Watch.Profiles[0].State.Status != "reconnecting" {
		t.Fatalf("expected refreshed watch loop to reuse cached profiles and synced reconnecting status, got %#v", seeded.Watch.Profiles)
	}
	if !seeded.Cycle.Observation.StatusObservedAt.Equal(observedAt) {
		t.Fatalf("expected refreshed watch loop to use status timestamp %v, got %v", observedAt, seeded.Cycle.Observation.StatusObservedAt)
	}
	if !seeded.Cycle.Observation.ProfilesObservedAt.Equal(initialProfilesObservedAt) {
		t.Fatalf("expected refreshed watch loop to reuse cached profiles timestamp %v, got %v", initialProfilesObservedAt, seeded.Cycle.Observation.ProfilesObservedAt)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected next ObserveWatchLoop to reuse refreshed projection cache, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
	if runRegistry.callCount() != 2 {
		t.Fatalf("expected next ObserveWatchLoop to avoid second projection rebuild, got %d", runRegistry.callCount())
	}
}

func TestSyncSharedSessionBrowserProfilesEventSeedsSharedWatchManagerRawProfilesCache(t *testing.T) {
	backend := &statusProfilesObservationTestBackend{
		profilesResp: BrowserProfilesResult{
			Backend:        "proxy",
			DefaultProfile: "isolated",
			Profiles: []BrowserProfileInfo{{
				Profile:   "isolated",
				Status:    "stopped",
				Running:   false,
				Connected: false,
			}},
		},
	}
	stateRegistry := NewBrowserSessionStateRegistry()
	manager := SharedSessionBrowserObserverManagerFor(nil, nil, stateRegistry, time.Minute)
	bound := manager.Bind(backend)
	sessionID := "sess-profiles-seed"
	observedAt := time.Date(2026, time.March, 30, 7, 5, 0, 0, time.UTC)

	initial := bound.ObserveRawProfiles(context.Background(), "isolated")
	if initial.Profiles == nil || initial.Profiles.DefaultProfile != "isolated" {
		t.Fatalf("expected initial raw profiles poll to reflect backend result, got %#v", initial)
	}
	if len(backend.profilesReqs) != 1 {
		t.Fatalf("expected initial raw profiles observation to poll backend once, got %#v", backend.profilesReqs)
	}

	snapshot := SyncSharedSessionBrowserProfilesEvent(
		stateRegistry,
		sessionID,
		BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		"isolated",
		BrowserProfilesResult{
			Backend:        "proxy",
			DefaultProfile: "isolated",
			Profiles: []BrowserProfileInfo{{
				Profile:    "isolated",
				BrowserApp: "Chromium",
				Status:     "running",
				Running:    true,
				Connected:  true,
			}},
		},
		observedAt,
		time.Minute,
	)
	if len(snapshot) != 1 || snapshot[0].Status != "running" {
		t.Fatalf("expected synced profiles snapshot, got %#v", snapshot)
	}

	seeded := bound.ObserveRawProfiles(context.Background(), "isolated")
	if seeded.Profiles == nil || len(seeded.Profiles.Profiles) != 1 || seeded.Profiles.Profiles[0].Status != "running" {
		t.Fatalf("expected seeded raw profiles observation, got %#v", seeded)
	}
	if !seeded.ObservedAt.Equal(observedAt) {
		t.Fatalf("expected seeded raw profiles timestamp %v, got %v", observedAt, seeded.ObservedAt)
	}
	if len(backend.profilesReqs) != 1 {
		t.Fatalf("expected event-synced raw profiles cache to avoid a second poll, got %#v", backend.profilesReqs)
	}
}

func TestSyncSharedSessionBrowserProfileLifecycleEventSeedsSharedWatchManagerRawStatusCache(t *testing.T) {
	backend := &statusProfilesObservationTestBackend{
		statusResp: BrowserProfileStatusResult{
			Backend:   "proxy",
			Profile:   "isolated",
			Status:    "stopped",
			Running:   false,
			Connected: false,
		},
	}
	stateRegistry := NewBrowserSessionStateRegistry()
	manager := SharedSessionBrowserObserverManagerFor(nil, nil, stateRegistry, time.Minute)
	bound := manager.Bind(backend)
	sessionID := "sess-lifecycle-seed"
	observedAt := time.Date(2026, time.March, 30, 7, 10, 0, 0, time.UTC)

	initial := bound.ObserveRawStatus(context.Background(), "isolated")
	if initial.Status == nil || initial.Status.Status != "stopped" {
		t.Fatalf("expected initial raw status poll to reflect backend result, got %#v", initial)
	}
	if len(backend.statusReqs) != 1 {
		t.Fatalf("expected initial raw status observation to poll backend once, got %#v", backend.statusReqs)
	}

	state, ok := SyncSharedSessionBrowserProfileLifecycleEvent(
		stateRegistry,
		nil,
		nil,
		sessionID,
		BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		"isolated",
		BrowserProfileStatusResult{
			Backend:   "proxy",
			Profile:   "isolated",
			Status:    "starting",
			Running:   true,
			Connected: false,
		},
		"restart_started",
		observedAt,
		time.Minute,
	)
	if !ok || state.Status != "starting" && state.Status != "reconnecting" {
		t.Fatalf("expected lifecycle-synced state, got %#v ok=%v", state, ok)
	}

	seeded := bound.ObserveRawStatus(context.Background(), "isolated")
	if seeded.Status == nil || seeded.Status.Status != "starting" {
		t.Fatalf("expected lifecycle event to seed raw status observation, got %#v", seeded)
	}
	if !seeded.ObservedAt.Equal(observedAt) {
		t.Fatalf("expected seeded lifecycle raw status timestamp %v, got %v", observedAt, seeded.ObservedAt)
	}
	if len(backend.statusReqs) != 1 {
		t.Fatalf("expected lifecycle event to avoid a second raw status poll, got %#v", backend.statusReqs)
	}
}

func TestSyncSharedSessionBrowserProfileLifecycleEventSeedsWatchLoopFromCachedEventCycleWhenProfilesDrained(t *testing.T) {
	backend := &statusProfilesObservationTestBackend{
		statusResp: BrowserProfileStatusResult{
			Backend:   "proxy",
			Profile:   "isolated",
			Status:    "stopped",
			Running:   false,
			Connected: false,
		},
		profilesResp: BrowserProfilesResult{
			Backend:        "proxy",
			DefaultProfile: "isolated",
			Profiles: []BrowserProfileInfo{{
				Profile:    "isolated",
				BrowserApp: "Chromium",
				Status:     "stopped",
				Running:    false,
				Connected:  false,
			}},
		},
	}
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	sessionID := "sess-lifecycle-watchloop-refresh"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}
	sessionRegistry.TrackTab(sessionID, BrowserSessionTarget{
		TabIndex:   1,
		URL:        "https://example.com/one",
		Title:      "One",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
	}, true)
	stateRegistry.SelectBrowserProfile(sessionID, SharedSessionBrowserProfileSelection{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Source:        "remember_profile",
	})
	runRegistry := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			sessionID: {{RunID: "run-1", Status: "running"}},
		}},
	}
	manager := SharedSessionBrowserObserverManagerFor(sessionRegistry, runRegistry, stateRegistry, time.Minute)
	bound := manager.Bind(backend)
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

	initial := bound.ObserveWatchLoop(context.Background(), req)
	if len(initial.View.Session.Routes) != 1 || len(initial.View.Session.Routes[0].Targets) != 1 || initial.View.Session.Routes[0].Targets[0].Title != "One" {
		t.Fatalf("expected initial watch loop to expose tracked route snapshot, got %#v", initial.View.Session.Routes)
	}
	if len(initial.Watch.Profiles) != 1 || initial.Watch.Profiles[0].State.Status != "stopped" {
		t.Fatalf("expected initial watch loop to reflect stopped profile, got %#v", initial.Watch.Profiles)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected initial watch loop to poll backend once, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
	if runRegistry.callCount() != 1 {
		t.Fatalf("expected initial watch loop to snapshot runs once, got %d", runRegistry.callCount())
	}
	initialProfilesObservedAt := initial.Cycle.Observation.ProfilesObservedAt

	bound.state.mu.Lock()
	clear(bound.state.rawStatus)
	clear(bound.state.rawProfiles)
	eventCycleCount := len(bound.state.eventCycles)
	watchLoopCount := len(bound.state.watchLoops)
	bound.state.mu.Unlock()
	if eventCycleCount == 0 || watchLoopCount == 0 {
		t.Fatalf("expected cached event-cycle/watch-loop source before draining raw caches, got eventCycles=%d watchLoops=%d", eventCycleCount, watchLoopCount)
	}

	observedAt := time.Date(2026, time.March, 30, 7, 12, 0, 0, time.UTC)
	state, ok := SyncSharedSessionBrowserProfileLifecycleEvent(
		stateRegistry,
		sessionRegistry,
		runRegistry,
		sessionID,
		BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		"isolated",
		BrowserProfileStatusResult{
			Backend:    "proxy",
			Profile:    "isolated",
			BrowserApp: "Chromium",
			Status:     "starting",
			Running:    true,
			Connected:  false,
		},
		"restart_started",
		observedAt,
		time.Minute,
	)
	if !ok || (state.Status != "starting" && state.Status != "reconnecting") {
		t.Fatalf("expected lifecycle-synced starting/reconnecting state, got %#v ok=%v", state, ok)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected lifecycle-only sync to reuse cached event-cycle source, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
	if runRegistry.callCount() != 2 {
		t.Fatalf("expected lifecycle-only sync to refresh watch-loop projection once, got %d", runRegistry.callCount())
	}

	seeded := bound.ObserveWatchLoop(context.Background(), req)
	if len(seeded.View.Session.Routes) != 1 || seeded.View.Session.Routes[0].CurrentTargetID != "" {
		t.Fatalf("expected lifecycle sync to invalidate current target in refreshed watch loop, got %#v", seeded.View.Session.Routes)
	}
	if len(seeded.Watch.Profiles) != 1 || seeded.Watch.Profiles[0].State.Status != state.Status {
		t.Fatalf("expected refreshed watch loop to reuse cached profiles and lifecycle-updated status %q, got %#v", state.Status, seeded.Watch.Profiles)
	}
	if !seeded.Cycle.Observation.StatusObservedAt.Equal(observedAt) {
		t.Fatalf("expected refreshed watch loop to use lifecycle status timestamp %v, got %v", observedAt, seeded.Cycle.Observation.StatusObservedAt)
	}
	if !seeded.Cycle.Observation.ProfilesObservedAt.Equal(initialProfilesObservedAt) {
		t.Fatalf("expected refreshed watch loop to reuse cached profiles timestamp %v, got %v", initialProfilesObservedAt, seeded.Cycle.Observation.ProfilesObservedAt)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected next ObserveWatchLoop to reuse refreshed lifecycle projection cache, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
	if runRegistry.callCount() != 2 {
		t.Fatalf("expected next ObserveWatchLoop to avoid second lifecycle projection rebuild, got %d", runRegistry.callCount())
	}
}

func TestSyncSharedSessionBrowserStatusAndProfilesEventSeedsSharedWatchManagerEventCycle(t *testing.T) {
	backend := &statusProfilesObservationTestBackend{
		statusResp: BrowserProfileStatusResult{
			Backend:   "proxy",
			Profile:   "isolated",
			Status:    "stopped",
			Running:   false,
			Connected: false,
		},
		profilesResp: BrowserProfilesResult{
			Backend:        "proxy",
			DefaultProfile: "isolated",
			Profiles: []BrowserProfileInfo{{
				Profile:   "isolated",
				Status:    "stopped",
				Running:   false,
				Connected: false,
			}},
		},
	}
	stateRegistry := NewBrowserSessionStateRegistry()
	manager := SharedSessionBrowserObserverManagerFor(nil, nil, stateRegistry, time.Minute)
	bound := manager.Bind(backend)
	sessionID := "sess-status-profiles-seed"
	req := SharedSessionBrowserObserverRequest{
		SessionID:        sessionID,
		SelectedInfo:     BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		BindingRoute:     BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node"},
		RequestedProfile: "isolated",
		IncludeStatus:    true,
		IncludeProfiles:  true,
	}
	initial := bound.ObserveEventCycle(context.Background(), req)
	if initial.Observation.Status == nil || initial.Observation.Profiles == nil {
		t.Fatalf("expected initial event cycle to poll backend, got %#v", initial)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected initial event cycle to poll once, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}

	statusObservedAt := time.Date(2026, time.March, 30, 7, 20, 0, 0, time.UTC)
	profilesObservedAt := time.Date(2026, time.March, 30, 7, 20, 1, 0, time.UTC)
	resolved, synced, ok, snapshot := SyncSharedSessionBrowserStatusAndProfilesEvent(
		stateRegistry,
		nil,
		sessionID,
		BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		"isolated",
		&BrowserProfileStatusResult{
			Backend:    "proxy",
			Profile:    "isolated",
			BrowserApp: "Chromium",
			Status:     "running",
			Running:    true,
			Connected:  true,
		},
		statusObservedAt,
		&BrowserProfilesResult{
			Backend:        "proxy",
			DefaultProfile: "isolated",
			Profiles: []BrowserProfileInfo{{
				Profile:    "isolated",
				BrowserApp: "Chromium",
				Status:     "running",
				Running:    true,
				Connected:  true,
			}},
		},
		profilesObservedAt,
		time.Minute,
	)
	if !ok || resolved.Status != "running" || synced.Status != "running" || len(snapshot) != 1 || snapshot[0].Status != "running" {
		t.Fatalf("expected combined sync event to resolve running state, got resolved=%#v synced=%#v ok=%v snapshot=%#v", resolved, synced, ok, snapshot)
	}

	seeded := bound.ObserveEventCycle(context.Background(), req)
	if seeded.Observation.Status == nil || seeded.Observation.Status.Status != "running" || seeded.Observation.Profiles == nil || len(seeded.Observation.Snapshot) != 1 || seeded.Observation.Snapshot[0].Status != "running" {
		t.Fatalf("expected combined sync event to seed next event cycle, got %#v", seeded)
	}
	if !seeded.Observation.StatusObservedAt.Equal(statusObservedAt) || !seeded.Observation.ProfilesObservedAt.Equal(profilesObservedAt) {
		t.Fatalf("expected seeded event cycle timestamps status=%v profiles=%v, got status=%v profiles=%v", statusObservedAt, profilesObservedAt, seeded.Observation.StatusObservedAt, seeded.Observation.ProfilesObservedAt)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected combined sync event to avoid a second backend poll, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}

func TestSyncSharedSessionBrowserStatusAndProfilesEventSeedsEventCycleCacheWithoutSecondResolution(t *testing.T) {
	backend := &statusProfilesObservationTestBackend{
		statusResp: BrowserProfileStatusResult{
			Backend:   "proxy",
			Profile:   "isolated",
			Status:    "stopped",
			Running:   false,
			Connected: false,
		},
		profilesResp: BrowserProfilesResult{
			Backend:        "proxy",
			DefaultProfile: "isolated",
			Profiles: []BrowserProfileInfo{{
				Profile:   "isolated",
				Status:    "stopped",
				Running:   false,
				Connected: false,
			}},
		},
	}
	stateRegistry := &countingBrowserSessionStateRegistry{BrowserSessionStateRegistry: NewBrowserSessionStateRegistry()}
	manager := SharedSessionBrowserObserverManagerFor(nil, nil, stateRegistry, time.Minute)
	bound := manager.Bind(backend)
	sessionID := "sess-status-profiles-cycle-cache"
	req := SharedSessionBrowserObserverRequest{
		SessionID:        sessionID,
		SelectedInfo:     BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		BindingRoute:     BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node"},
		RequestedProfile: "isolated",
		IncludeStatus:    true,
		IncludeProfiles:  true,
	}

	initial := bound.ObserveEventCycle(context.Background(), req)
	if initial.Observation.Status == nil || initial.Observation.Profiles == nil {
		t.Fatalf("expected initial event cycle to include status and profiles, got %#v", initial)
	}
	if stateRegistry.resolutionCallCount() != 1 {
		t.Fatalf("expected initial event cycle to resolve state once, got %d", stateRegistry.resolutionCallCount())
	}

	statusObservedAt := time.Date(2026, time.March, 30, 7, 30, 0, 0, time.UTC)
	profilesObservedAt := time.Date(2026, time.March, 30, 7, 30, 1, 0, time.UTC)
	_, _, ok, snapshot := SyncSharedSessionBrowserStatusAndProfilesEvent(
		stateRegistry,
		nil,
		sessionID,
		BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		"isolated",
		&BrowserProfileStatusResult{
			Backend:    "proxy",
			Profile:    "isolated",
			BrowserApp: "Chromium",
			Status:     "running",
			Running:    true,
			Connected:  true,
		},
		statusObservedAt,
		&BrowserProfilesResult{
			Backend:        "proxy",
			DefaultProfile: "isolated",
			Profiles: []BrowserProfileInfo{{
				Profile:    "isolated",
				BrowserApp: "Chromium",
				Status:     "running",
				Running:    true,
				Connected:  true,
			}},
		},
		profilesObservedAt,
		time.Minute,
	)
	if !ok || len(snapshot) != 1 || snapshot[0].Status != "running" {
		t.Fatalf("expected combined sync event to resolve running snapshot, got ok=%v snapshot=%#v", ok, snapshot)
	}
	if stateRegistry.resolutionCallCount() != 2 {
		t.Fatalf("expected sync event to perform one additional resolution, got %d", stateRegistry.resolutionCallCount())
	}

	seeded := bound.ObserveEventCycle(context.Background(), req)
	if seeded.Observation.Status == nil || seeded.Observation.Status.Status != "running" || seeded.Observation.Profiles == nil || len(seeded.Observation.Snapshot) != 1 || seeded.Observation.Snapshot[0].Status != "running" {
		t.Fatalf("expected event cycle cache to return seeded running observation, got %#v", seeded)
	}
	if !seeded.Observation.StatusObservedAt.Equal(statusObservedAt) || !seeded.Observation.ProfilesObservedAt.Equal(profilesObservedAt) {
		t.Fatalf("expected seeded event cycle timestamps status=%v profiles=%v, got status=%v profiles=%v", statusObservedAt, profilesObservedAt, seeded.Observation.StatusObservedAt, seeded.Observation.ProfilesObservedAt)
	}
	if stateRegistry.resolutionCallCount() != 2 {
		t.Fatalf("expected seeded event cycle cache to avoid a second resolution, got %d", stateRegistry.resolutionCallCount())
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected seeded event cycle cache to avoid extra backend polls, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}

func TestSharedSessionBrowserObserverManagerSyncStatusAndProfilesEventSeedsWatchLoopCache(t *testing.T) {
	backend := &statusProfilesObservationTestBackend{
		statusResp: BrowserProfileStatusResult{
			Backend:   "proxy",
			Profile:   "isolated",
			Status:    "stopped",
			Running:   false,
			Connected: false,
		},
		profilesResp: BrowserProfilesResult{
			Backend:        "proxy",
			DefaultProfile: "isolated",
			Profiles: []BrowserProfileInfo{{
				Profile:   "isolated",
				Status:    "stopped",
				Running:   false,
				Connected: false,
			}},
		},
	}
	sessionRegistry := NewBrowserSessionRegistry()
	sessionID := "sess-status-profiles-watchloop-cache"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}
	sessionRegistry.TrackTab(sessionID, BrowserSessionTarget{
		TabIndex:   1,
		URL:        "https://example.com/one",
		Title:      "One",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
	}, true)

	stateRegistry := NewBrowserSessionStateRegistry()
	stateRegistry.SelectBrowserProfile(sessionID, SharedSessionBrowserProfileSelection{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Source:        "remember_profile",
	})
	runRegistry := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			sessionID: {{RunID: "run-1", Status: "running"}},
		}},
	}
	manager := SharedSessionBrowserObserverManagerFor(sessionRegistry, runRegistry, stateRegistry, time.Minute)
	bound := manager.Bind(backend)
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

	initial := bound.ObserveWatchLoop(context.Background(), req)
	if len(initial.View.Session.Routes) != 1 || len(initial.View.Session.Routes[0].Targets) != 1 || initial.View.Session.Routes[0].Targets[0].Title != "One" {
		t.Fatalf("expected initial watch loop to expose tracked tab, got %#v", initial.View.Session.Routes)
	}
	if len(initial.Watch.Profiles) != 1 || initial.Watch.Profiles[0].State.Status != "stopped" {
		t.Fatalf("expected initial watch loop to reflect stopped profile, got %#v", initial.Watch.Profiles)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected initial watch loop to poll backend once, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
	if runRegistry.callCount() != 1 {
		t.Fatalf("expected initial watch loop to snapshot runs once, got %d", runRegistry.callCount())
	}

	statusObservedAt := time.Date(2026, time.March, 30, 7, 40, 0, 0, time.UTC)
	profilesObservedAt := time.Date(2026, time.March, 30, 7, 40, 1, 0, time.UTC)
	resolved, synced, ok, snapshot := manager.SyncStatusAndProfilesEvent(
		sessionID,
		BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		"isolated",
		&BrowserProfileStatusResult{
			Backend:    "proxy",
			Profile:    "isolated",
			BrowserApp: "Chromium",
			Status:     "running",
			Running:    true,
			Connected:  true,
		},
		statusObservedAt,
		&BrowserProfilesResult{
			Backend:        "proxy",
			DefaultProfile: "isolated",
			Profiles: []BrowserProfileInfo{{
				Profile:    "isolated",
				BrowserApp: "Chromium",
				Status:     "running",
				Running:    true,
				Connected:  true,
			}},
		},
		profilesObservedAt,
	)
	if !ok || resolved.Status != "running" || synced.Status != "running" || len(snapshot) != 1 || snapshot[0].Status != "running" {
		t.Fatalf("expected sync event to resolve running state, got resolved=%#v synced=%#v ok=%v snapshot=%#v", resolved, synced, ok, snapshot)
	}
	if runRegistry.callCount() != 2 {
		t.Fatalf("expected sync event to refresh watch-loop projection once, got %d", runRegistry.callCount())
	}

	seeded := bound.ObserveWatchLoop(context.Background(), req)
	if len(seeded.View.Session.Routes) != 1 || len(seeded.View.Session.Routes[0].Targets) != 1 || seeded.View.Session.Routes[0].Targets[0].Title != "One" {
		t.Fatalf("expected seeded watch loop to preserve session view, got %#v", seeded.View.Session.Routes)
	}
	if len(seeded.Watch.Profiles) != 1 || seeded.Watch.Profiles[0].State.Status != "running" {
		t.Fatalf("expected seeded watch loop to reflect running profile, got %#v", seeded.Watch.Profiles)
	}
	if !seeded.Cycle.Observation.StatusObservedAt.Equal(statusObservedAt) || !seeded.Cycle.Observation.ProfilesObservedAt.Equal(profilesObservedAt) {
		t.Fatalf("expected seeded watch loop timestamps status=%v profiles=%v, got status=%v profiles=%v", statusObservedAt, profilesObservedAt, seeded.Cycle.Observation.StatusObservedAt, seeded.Cycle.Observation.ProfilesObservedAt)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected seeded watch loop cache to avoid extra backend polls, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
	if runRegistry.callCount() != 2 {
		t.Fatalf("expected seeded watch loop cache to avoid second projection rebuild, got %d", runRegistry.callCount())
	}
}

func TestSharedSessionBrowserObserverManagerSyncStatusAndProfilesEventSeedsSiblingProviderWatchLoopCaches(t *testing.T) {
	backend := &statusProfilesObservationTestBackend{
		statusResp: BrowserProfileStatusResult{
			Backend:   "proxy",
			Profile:   "isolated",
			Status:    "stopped",
			Running:   false,
			Connected: false,
		},
		profilesResp: BrowserProfilesResult{
			Backend:        "proxy",
			DefaultProfile: "isolated",
			Profiles: []BrowserProfileInfo{{
				Profile:   "isolated",
				Status:    "stopped",
				Running:   false,
				Connected: false,
			}},
		},
	}
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	sessionID := "sess-direct-status-profiles-sibling-provider-watchloop-cache"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}
	sessionRegistry.TrackTab(sessionID, BrowserSessionTarget{
		TabIndex:   1,
		URL:        "https://example.com/one",
		Title:      "One",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
	}, true)
	stateRegistry.SelectBrowserProfile(sessionID, SharedSessionBrowserProfileSelection{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Source:        "remember_profile",
	})

	runRegistryA := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			sessionID: {{RunID: "run-a", Status: "running"}},
		}},
	}
	runRegistryB := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			sessionID: {{RunID: "run-b", Status: "running"}},
		}},
	}
	managerA := SharedSessionBrowserObserverManagerFor(sessionRegistry, runRegistryA, stateRegistry, time.Minute)
	managerB := SharedSessionBrowserObserverManagerFor(sessionRegistry, runRegistryB, stateRegistry, time.Minute)
	boundA := managerA.Bind(backend)
	boundB := managerB.Bind(backend)
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
	if len(initialA.Watch.Profiles) != 1 || initialA.Watch.Profiles[0].State.Status != "stopped" {
		t.Fatalf("expected first sibling watch loop to reflect stopped profile, got %#v", initialA.Watch.Profiles)
	}
	if len(initialB.Watch.Profiles) != 1 || initialB.Watch.Profiles[0].State.Status != "stopped" {
		t.Fatalf("expected second sibling watch loop to reflect stopped profile, got %#v", initialB.Watch.Profiles)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected first sibling watch loop to seed stopped lifecycle source for the second provider, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
	if runRegistryA.callCount() != 1 || runRegistryB.callCount() != 1 {
		t.Fatalf("expected initial sibling watch loops to snapshot runs once each, got runA=%d runB=%d", runRegistryA.callCount(), runRegistryB.callCount())
	}

	statusObservedAt := time.Date(2026, time.March, 30, 7, 45, 0, 0, time.UTC)
	profilesObservedAt := time.Date(2026, time.March, 30, 7, 45, 1, 0, time.UTC)
	resolved, synced, ok, snapshot := managerA.SyncStatusAndProfilesEvent(
		sessionID,
		BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		"isolated",
		&BrowserProfileStatusResult{
			Backend:    "proxy",
			Profile:    "isolated",
			BrowserApp: "Chromium",
			Status:     "running",
			Running:    true,
			Connected:  true,
		},
		statusObservedAt,
		&BrowserProfilesResult{
			Backend:        "proxy",
			DefaultProfile: "isolated",
			Profiles: []BrowserProfileInfo{{
				Profile:    "isolated",
				BrowserApp: "Chromium",
				Status:     "running",
				Running:    true,
				Connected:  true,
			}},
		},
		profilesObservedAt,
	)
	if !ok || resolved.Status != "running" || synced.Status != "running" || len(snapshot) != 1 || snapshot[0].Status != "running" {
		t.Fatalf("expected direct manager sync to resolve running state, got resolved=%#v synced=%#v ok=%v snapshot=%#v", resolved, synced, ok, snapshot)
	}
	if runRegistryA.callCount() != 2 || runRegistryB.callCount() != 2 {
		t.Fatalf("expected direct manager sync to refresh both sibling watch-loop projections once, got runA=%d runB=%d", runRegistryA.callCount(), runRegistryB.callCount())
	}

	seededA := boundA.ObserveWatchLoop(context.Background(), req)
	seededB := boundB.ObserveWatchLoop(context.Background(), req)
	if len(seededA.Watch.Profiles) != 1 || seededA.Watch.Profiles[0].State.Status != "running" {
		t.Fatalf("expected first sibling watch loop to reuse seeded running profile, got %#v", seededA.Watch.Profiles)
	}
	if len(seededB.Watch.Profiles) != 1 || seededB.Watch.Profiles[0].State.Status != "running" {
		t.Fatalf("expected second sibling watch loop to reuse seeded running profile, got %#v", seededB.Watch.Profiles)
	}
	if !seededA.Cycle.Observation.StatusObservedAt.Equal(statusObservedAt) || !seededA.Cycle.Observation.ProfilesObservedAt.Equal(profilesObservedAt) {
		t.Fatalf("expected first sibling watch loop timestamps status=%v profiles=%v, got status=%v profiles=%v", statusObservedAt, profilesObservedAt, seededA.Cycle.Observation.StatusObservedAt, seededA.Cycle.Observation.ProfilesObservedAt)
	}
	if !seededB.Cycle.Observation.StatusObservedAt.Equal(statusObservedAt) || !seededB.Cycle.Observation.ProfilesObservedAt.Equal(profilesObservedAt) {
		t.Fatalf("expected second sibling watch loop timestamps status=%v profiles=%v, got status=%v profiles=%v", statusObservedAt, profilesObservedAt, seededB.Cycle.Observation.StatusObservedAt, seededB.Cycle.Observation.ProfilesObservedAt)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected sibling watch loops to reuse event-seeded caches without extra polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
	if runRegistryA.callCount() != 2 || runRegistryB.callCount() != 2 {
		t.Fatalf("expected sibling watch loops to reuse seeded projections without extra rebuilds, got runA=%d runB=%d", runRegistryA.callCount(), runRegistryB.callCount())
	}
}

func TestSyncSharedSessionBrowserStatusAndProfilesEventSeedsSiblingProviderWatchLoopCaches(t *testing.T) {
	backend := &statusProfilesObservationTestBackend{
		statusResp: BrowserProfileStatusResult{
			Backend:   "proxy",
			Profile:   "isolated",
			Status:    "stopped",
			Running:   false,
			Connected: false,
		},
		profilesResp: BrowserProfilesResult{
			Backend:        "proxy",
			DefaultProfile: "isolated",
			Profiles: []BrowserProfileInfo{{
				Profile:   "isolated",
				Status:    "stopped",
				Running:   false,
				Connected: false,
			}},
		},
	}
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	sessionID := "sess-status-profiles-sibling-provider-watchloop-cache"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}
	sessionRegistry.TrackTab(sessionID, BrowserSessionTarget{
		TabIndex:   1,
		URL:        "https://example.com/one",
		Title:      "One",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
	}, true)
	stateRegistry.SelectBrowserProfile(sessionID, SharedSessionBrowserProfileSelection{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Source:        "remember_profile",
	})

	runRegistryA := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			sessionID: {{RunID: "run-a", Status: "running"}},
		}},
	}
	runRegistryB := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			sessionID: {{RunID: "run-b", Status: "running"}},
		}},
	}
	managerA := SharedSessionBrowserObserverManagerFor(sessionRegistry, runRegistryA, stateRegistry, time.Minute)
	managerB := SharedSessionBrowserObserverManagerFor(sessionRegistry, runRegistryB, stateRegistry, time.Minute)
	boundA := managerA.Bind(backend)
	boundB := managerB.Bind(backend)
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
	if len(initialA.Watch.Profiles) != 1 || initialA.Watch.Profiles[0].State.Status != "stopped" {
		t.Fatalf("expected first sibling watch loop to reflect stopped profile, got %#v", initialA.Watch.Profiles)
	}
	if len(initialB.Watch.Profiles) != 1 || initialB.Watch.Profiles[0].State.Status != "stopped" {
		t.Fatalf("expected second sibling watch loop to reflect stopped profile, got %#v", initialB.Watch.Profiles)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected first sibling watch loop to seed stopped lifecycle source for the second provider, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
	if runRegistryA.callCount() != 1 || runRegistryB.callCount() != 1 {
		t.Fatalf("expected initial sibling watch loops to snapshot runs once each, got runA=%d runB=%d", runRegistryA.callCount(), runRegistryB.callCount())
	}

	statusObservedAt := time.Date(2026, time.March, 30, 7, 50, 0, 0, time.UTC)
	profilesObservedAt := time.Date(2026, time.March, 30, 7, 50, 1, 0, time.UTC)
	resolved, synced, ok, snapshot := SyncSharedSessionBrowserStatusAndProfilesEvent(
		stateRegistry,
		sessionRegistry,
		sessionID,
		BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		"isolated",
		&BrowserProfileStatusResult{
			Backend:    "proxy",
			Profile:    "isolated",
			BrowserApp: "Chromium",
			Status:     "running",
			Running:    true,
			Connected:  true,
		},
		statusObservedAt,
		&BrowserProfilesResult{
			Backend:        "proxy",
			DefaultProfile: "isolated",
			Profiles: []BrowserProfileInfo{{
				Profile:    "isolated",
				BrowserApp: "Chromium",
				Status:     "running",
				Running:    true,
				Connected:  true,
			}},
		},
		profilesObservedAt,
		time.Minute,
	)
	if !ok || resolved.Status != "running" || synced.Status != "running" || len(snapshot) != 1 || snapshot[0].Status != "running" {
		t.Fatalf("expected top-level sync to resolve running state, got resolved=%#v synced=%#v ok=%v snapshot=%#v", resolved, synced, ok, snapshot)
	}
	if runRegistryA.callCount() != 2 || runRegistryB.callCount() != 2 {
		t.Fatalf("expected top-level sync to refresh both sibling watch-loop projections once, got runA=%d runB=%d", runRegistryA.callCount(), runRegistryB.callCount())
	}

	seededA := boundA.ObserveWatchLoop(context.Background(), req)
	seededB := boundB.ObserveWatchLoop(context.Background(), req)
	if len(seededA.Watch.Profiles) != 1 || seededA.Watch.Profiles[0].State.Status != "running" {
		t.Fatalf("expected first sibling watch loop to reuse seeded running profile, got %#v", seededA.Watch.Profiles)
	}
	if len(seededB.Watch.Profiles) != 1 || seededB.Watch.Profiles[0].State.Status != "running" {
		t.Fatalf("expected second sibling watch loop to reuse seeded running profile, got %#v", seededB.Watch.Profiles)
	}
	if !seededA.Cycle.Observation.StatusObservedAt.Equal(statusObservedAt) || !seededA.Cycle.Observation.ProfilesObservedAt.Equal(profilesObservedAt) {
		t.Fatalf("expected first sibling watch loop timestamps status=%v profiles=%v, got status=%v profiles=%v", statusObservedAt, profilesObservedAt, seededA.Cycle.Observation.StatusObservedAt, seededA.Cycle.Observation.ProfilesObservedAt)
	}
	if !seededB.Cycle.Observation.StatusObservedAt.Equal(statusObservedAt) || !seededB.Cycle.Observation.ProfilesObservedAt.Equal(profilesObservedAt) {
		t.Fatalf("expected second sibling watch loop timestamps status=%v profiles=%v, got status=%v profiles=%v", statusObservedAt, profilesObservedAt, seededB.Cycle.Observation.StatusObservedAt, seededB.Cycle.Observation.ProfilesObservedAt)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected sibling watch loops to reuse event-seeded caches without extra polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
	if runRegistryA.callCount() != 2 || runRegistryB.callCount() != 2 {
		t.Fatalf("expected sibling watch loops to reuse seeded projections without extra rebuilds, got runA=%d runB=%d", runRegistryA.callCount(), runRegistryB.callCount())
	}
}

func TestSyncSharedSessionBrowserTabsForRouteEventRefreshesSharedProviderProjectionFromCachedEventCycleWhenRawDrained(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	runRegistry := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-sync-tabs-event-shared-provider": {{RunID: "run-1", Status: "running"}},
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
	sessionID := "sess-sync-tabs-event-shared-provider"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}

	seeded := SyncSharedSessionBrowserTabsForRouteEvent(
		sessionRegistry,
		runRegistry,
		nil,
		sessionID,
		route,
		1,
		[]BrowserTab{{Index: 1, URL: "https://example.com/one", Title: "One"}},
		time.Minute,
	)
	if len(seeded) != 1 || seeded[0].Title != "One" {
		t.Fatalf("expected top-level tabs sync to seed initial tracked tab, got %#v", seeded)
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

	tracked := SyncSharedSessionBrowserTabsForRouteEvent(
		sessionRegistry,
		runRegistry,
		nil,
		sessionID,
		route,
		1,
		[]BrowserTab{{Index: 1, URL: "https://example.com/one", Title: "Two"}},
		time.Minute,
	)
	if len(tracked) != 1 || tracked[0].Title != "Two" {
		t.Fatalf("expected top-level tabs sync event to update tracked tab payload, got %#v", tracked)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected top-level tabs sync refresh to reuse cached event-cycle source, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
	if runRegistry.callCount() != 2 {
		t.Fatalf("expected top-level tabs sync to refresh projection once, got %d", runRegistry.callCount())
	}

	second := bound.ObserveWatchLoop(context.Background(), req)
	if len(second.View.Session.Routes) != 1 || len(second.View.Session.Routes[0].Targets) != 1 || second.View.Session.Routes[0].Targets[0].Title != "Two" {
		t.Fatalf("expected top-level tabs sync to refresh shared-provider watch-loop cache, got %#v", second.View.Session.Routes)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected next ObserveWatchLoop to reuse refreshed projection cache, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
	if runRegistry.callCount() != 2 {
		t.Fatalf("expected next ObserveWatchLoop to avoid second projection rebuild, got %d", runRegistry.callCount())
	}
}

func TestTrackSharedSessionBrowserCurrentTargetEventRefreshesSharedProviderProjectionFromCachedEventCycleWhenRawDrained(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	runRegistry := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-track-current-target-event-shared-provider": {{RunID: "run-1", Status: "running"}},
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
	sessionID := "sess-track-current-target-event-shared-provider"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}

	initial := TrackSharedSessionBrowserCurrentTargetEvent(
		sessionRegistry,
		runRegistry,
		nil,
		sessionID,
		route,
		"https://example.com/one",
		"One",
		"browser_navigate",
		time.Minute,
	)
	if initial.ID == "" {
		t.Fatalf("expected top-level current-target event to create a tracked target, got %#v", initial)
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

	updated := TrackSharedSessionBrowserCurrentTargetEvent(
		sessionRegistry,
		runRegistry,
		nil,
		sessionID,
		route,
		"https://example.com/two",
		"Two",
		"browser_navigate",
		time.Minute,
	)
	if updated.ID == "" {
		t.Fatalf("expected top-level current-target event to return tracked target, got %#v", updated)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected top-level current-target event refresh to reuse cached event-cycle source, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
	if runRegistry.callCount() != 2 {
		t.Fatalf("expected top-level current-target event to refresh projection once, got %d", runRegistry.callCount())
	}

	second := bound.ObserveWatchLoop(context.Background(), req)
	if len(second.View.Session.Routes) != 1 || second.View.Session.Routes[0].CurrentTargetID != updated.ID {
		t.Fatalf("expected top-level current-target event to refresh shared-provider watch-loop cache, got %#v", second.View.Session.Routes)
	}
	if len(second.View.Session.Routes[0].Targets) != 1 || second.View.Session.Routes[0].Targets[0].Title != "Two" || second.View.Session.Routes[0].Targets[0].URL != "https://example.com/two" {
		t.Fatalf("expected refreshed watch loop to expose updated current target, got %#v", second.View.Session.Routes)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected next ObserveWatchLoop to reuse refreshed projection cache, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
	if runRegistry.callCount() != 2 {
		t.Fatalf("expected next ObserveWatchLoop to avoid second projection rebuild, got %d", runRegistry.callCount())
	}
}

func TestSyncSharedSessionBrowserTabsForRouteEventSeedsSiblingProviderFromPrimaryCachedEventCycle(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	runRegistryA := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-sync-tabs-event-sibling-provider": {{RunID: "run-a", Status: "running"}},
		}},
	}
	runRegistryB := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-sync-tabs-event-sibling-provider": {{RunID: "run-b", Status: "running"}},
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
	sessionID := "sess-sync-tabs-event-sibling-provider"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}

	seeded := SyncSharedSessionBrowserTabsForRouteEvent(
		sessionRegistry,
		runRegistryA,
		nil,
		sessionID,
		route,
		1,
		[]BrowserTab{{Index: 1, URL: "https://example.com/one", Title: "One"}},
		time.Minute,
	)
	if len(seeded) != 1 || seeded[0].Title != "One" {
		t.Fatalf("expected initial top-level tabs sync to seed tracked tab, got %#v", seeded)
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
	if len(initialA.View.Session.Routes) != 1 || len(initialA.View.Session.Routes[0].Targets) != 1 || initialA.View.Session.Routes[0].Targets[0].Title != "One" {
		t.Fatalf("expected first sibling watch loop to expose tracked tab, got %#v", initialA.View.Session.Routes)
	}
	if len(initialB.View.Session.Routes) != 1 || len(initialB.View.Session.Routes[0].Targets) != 1 || initialB.View.Session.Routes[0].Targets[0].Title != "One" {
		t.Fatalf("expected second sibling watch loop to expose tracked tab, got %#v", initialB.View.Session.Routes)
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

	tracked := SyncSharedSessionBrowserTabsForRouteEvent(
		sessionRegistry,
		runRegistryA,
		nil,
		sessionID,
		route,
		1,
		[]BrowserTab{{Index: 1, URL: "https://example.com/one", Title: "Two"}},
		time.Minute,
	)
	if len(tracked) != 1 || tracked[0].Title != "Two" {
		t.Fatalf("expected tabs sync event to update tracked tab payload, got %#v", tracked)
	}

	seededB := boundB.ObserveWatchLoop(context.Background(), req)
	if len(seededB.View.Session.Routes) != 1 || len(seededB.View.Session.Routes[0].Targets) != 1 || seededB.View.Session.Routes[0].Targets[0].Title != "Two" {
		t.Fatalf("expected sibling watch loop to reuse primary provider source for updated tab payload, got %#v", seededB.View.Session.Routes)
	}
	if len(backend.statusReqs) != 2 || len(backend.profilesReqs) != 2 {
		t.Fatalf("expected sibling watch loop to reuse seeded source without extra polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}

func TestSyncSharedSessionBrowserTabsForRouteEventSeedsPrimaryAndSiblingProvidersFromStateSnapshotWithoutPriorPolling(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	sessionID := "sess-sync-tabs-event-state-snapshot-source"
	observedAt := time.Date(2026, time.March, 30, 8, 15, 0, 0, time.UTC)
	stateRegistry.RecordBrowserProfileState(sessionID, SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "connected",
		Running:       true,
		Connected:     true,
		ObservedAt:    observedAt,
	})

	runRegistryA := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			sessionID: {{RunID: "run-a", Status: "running"}},
		}},
	}
	runRegistryB := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			sessionID: {{RunID: "run-b", Status: "running"}},
		}},
	}
	managerA := SharedSessionBrowserObserverManagerFor(sessionRegistry, runRegistryA, stateRegistry, time.Minute)
	managerB := SharedSessionBrowserObserverManagerFor(sessionRegistry, runRegistryB, stateRegistry, time.Minute)
	backend := &statusProfilesObservationTestBackend{
		statusResp: BrowserProfileStatusResult{
			Profile:   "isolated",
			Status:    "connected",
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
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}

	tracked := SyncSharedSessionBrowserTabsForRouteEvent(
		sessionRegistry,
		runRegistryA,
		stateRegistry,
		sessionID,
		route,
		1,
		[]BrowserTab{{Index: 1, URL: "https://example.com/one", Title: "One"}},
		time.Minute,
	)
	if len(tracked) != 1 || tracked[0].Title != "One" {
		t.Fatalf("expected top-level tabs sync to update tracked tab payload, got %#v", tracked)
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

	seededA := boundA.ObserveWatchLoop(context.Background(), req)
	seededB := boundB.ObserveWatchLoop(context.Background(), req)
	if len(seededA.View.Session.Routes) != 1 || len(seededA.View.Session.Routes[0].Targets) != 1 || seededA.View.Session.Routes[0].Targets[0].Title != "One" {
		t.Fatalf("expected primary provider watch loop to reuse state-snapshot source for tracked tab, got %#v", seededA.View.Session.Routes)
	}
	if len(seededB.View.Session.Routes) != 1 || len(seededB.View.Session.Routes[0].Targets) != 1 || seededB.View.Session.Routes[0].Targets[0].Title != "One" {
		t.Fatalf("expected sibling provider watch loop to reuse state-snapshot source for tracked tab, got %#v", seededB.View.Session.Routes)
	}
	if backend.statusRequestCount() != 0 || backend.profilesRequestCount() != 0 {
		t.Fatalf("expected state-snapshot source to avoid backend polling entirely, got status=%d profiles=%d", backend.statusRequestCount(), backend.profilesRequestCount())
	}
}

func TestSyncSharedSessionBrowserTabsForRouteEventSeedsSiblingRawRouteMutationSourceWhenRouteMutationCycleDrained(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	sessionID := "sess-sync-tabs-event-raw-route-mutation-source"
	observedAt := time.Date(2026, time.March, 30, 8, 16, 0, 0, time.UTC)
	stateRegistry.RecordBrowserProfileState(sessionID, SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "connected",
		Running:       true,
		Connected:     true,
		ObservedAt:    observedAt,
	})

	runRegistryA := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			sessionID: {{RunID: "run-a", Status: "running"}},
		}},
	}
	runRegistryB := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			sessionID: {{RunID: "run-b", Status: "running"}},
		}},
	}
	managerA := SharedSessionBrowserObserverManagerFor(sessionRegistry, runRegistryA, stateRegistry, time.Minute)
	managerB := SharedSessionBrowserObserverManagerFor(sessionRegistry, runRegistryB, stateRegistry, time.Minute)
	backend := &statusProfilesObservationTestBackend{}
	boundA := managerA.Bind(backend)
	boundB := managerB.Bind(backend)
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}

	tracked := SyncSharedSessionBrowserTabsForRouteEvent(
		sessionRegistry,
		runRegistryA,
		stateRegistry,
		sessionID,
		route,
		1,
		[]BrowserTab{{Index: 1, URL: "https://example.com/one", Title: "One"}},
		time.Minute,
	)
	if len(tracked) != 1 || tracked[0].Title != "One" {
		t.Fatalf("expected top-level tabs sync to update tracked tab payload, got %#v", tracked)
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

	boundB.state.mu.Lock()
	clear(boundB.state.rawStatus)
	clear(boundB.state.rawProfiles)
	clear(boundB.state.routeMutations)
	clear(boundB.state.eventCycles)
	clear(boundB.state.bindings)
	clear(boundB.state.views)
	clear(boundB.state.watchLoops)
	clear(boundB.state.eventCyclesInFlight)
	clear(boundB.state.bindingsInFlight)
	clear(boundB.state.viewsInFlight)
	clear(boundB.state.watchLoopsInFlight)
	rawRouteMutationCount := len(boundB.state.rawRouteMutations)
	boundB.state.mu.Unlock()
	if rawRouteMutationCount == 0 {
		t.Fatalf("expected sibling provider raw route-mutation source before draining route-mutation cycle cache")
	}

	seededB := boundB.ObserveWatchLoop(context.Background(), req)
	if len(seededB.View.Session.Routes) != 1 || len(seededB.View.Session.Routes[0].Targets) != 1 || seededB.View.Session.Routes[0].Targets[0].Title != "One" {
		t.Fatalf("expected sibling watch loop to reuse raw route-mutation source for tracked tab, got %#v", seededB.View.Session.Routes)
	}
	if backend.statusRequestCount() != 0 || backend.profilesRequestCount() != 0 {
		t.Fatalf("expected sibling raw route-mutation source to avoid backend polling, got status=%d profiles=%d", backend.statusRequestCount(), backend.profilesRequestCount())
	}
}

func TestSyncSharedSessionBrowserTabsForRouteEventSeedsSiblingProviderFromPrimaryRouteMutationSourceWhenPrimaryCachesDrained(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	runRegistryA := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-sync-tabs-event-primary-route-mutation-source": {{RunID: "run-a", Status: "running"}},
		}},
	}
	runRegistryB := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-sync-tabs-event-primary-route-mutation-source": {{RunID: "run-b", Status: "running"}},
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
	sessionID := "sess-sync-tabs-event-primary-route-mutation-source"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}

	seeded := SyncSharedSessionBrowserTabsForRouteEvent(
		sessionRegistry,
		runRegistryA,
		nil,
		sessionID,
		route,
		1,
		[]BrowserTab{{Index: 1, URL: "https://example.com/one", Title: "One"}},
		time.Minute,
	)
	if len(seeded) != 1 || seeded[0].Title != "One" {
		t.Fatalf("expected initial top-level tabs sync to seed tracked tab, got %#v", seeded)
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
	if len(initialA.View.Session.Routes) != 1 || len(initialA.View.Session.Routes[0].Targets) != 1 || initialA.View.Session.Routes[0].Targets[0].Title != "One" {
		t.Fatalf("expected first sibling watch loop to expose tracked tab, got %#v", initialA.View.Session.Routes)
	}
	if len(initialB.View.Session.Routes) != 1 || len(initialB.View.Session.Routes[0].Targets) != 1 || initialB.View.Session.Routes[0].Targets[0].Title != "One" {
		t.Fatalf("expected second sibling watch loop to expose tracked tab, got %#v", initialB.View.Session.Routes)
	}
	if len(backend.statusReqs) != 2 || len(backend.profilesReqs) != 2 {
		t.Fatalf("expected initial sibling watch loops to poll backend once each, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}

	updated := SyncSharedSessionBrowserTabsForRouteEvent(
		sessionRegistry,
		runRegistryA,
		nil,
		sessionID,
		route,
		1,
		[]BrowserTab{{Index: 1, URL: "https://example.com/two", Title: "Two"}},
		time.Minute,
	)
	if len(updated) != 1 || updated[0].Title != "Two" {
		t.Fatalf("expected second tabs sync to update tracked tab payload, got %#v", updated)
	}

	boundA.state.mu.Lock()
	clear(boundA.state.rawStatus)
	clear(boundA.state.rawProfiles)
	clear(boundA.state.eventCycles)
	clear(boundA.state.bindings)
	clear(boundA.state.views)
	clear(boundA.state.watchLoops)
	clear(boundA.state.eventCyclesInFlight)
	clear(boundA.state.bindingsInFlight)
	clear(boundA.state.viewsInFlight)
	clear(boundA.state.watchLoopsInFlight)
	routeMutationCountA := len(boundA.state.routeMutations)
	boundA.state.mu.Unlock()
	if routeMutationCountA == 0 {
		t.Fatalf("expected primary provider to retain route-mutation source before draining event/raw caches")
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

	third := SyncSharedSessionBrowserTabsForRouteEvent(
		sessionRegistry,
		runRegistryA,
		nil,
		sessionID,
		route,
		1,
		[]BrowserTab{{Index: 1, URL: "https://example.com/three", Title: "Three"}},
		time.Minute,
	)
	if len(third) != 1 || third[0].Title != "Three" {
		t.Fatalf("expected third tabs sync to update tracked tab payload, got %#v", third)
	}

	seededB := boundB.ObserveWatchLoop(context.Background(), req)
	if len(seededB.View.Session.Routes) != 1 || len(seededB.View.Session.Routes[0].Targets) != 1 || seededB.View.Session.Routes[0].Targets[0].Title != "Three" {
		t.Fatalf("expected sibling watch loop to reuse primary route-mutation source for updated tab payload, got %#v", seededB.View.Session.Routes)
	}
	if len(backend.statusReqs) != 2 || len(backend.profilesReqs) != 2 {
		t.Fatalf("expected sibling watch loop to reuse primary route-mutation source without extra polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}

func TestTrackSharedSessionBrowserCurrentTargetEventSeedsSiblingProviderFromPrimaryCachedEventCycle(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	runRegistryA := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-track-current-target-event-sibling-provider": {{RunID: "run-a", Status: "running"}},
		}},
	}
	runRegistryB := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-track-current-target-event-sibling-provider": {{RunID: "run-b", Status: "running"}},
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
	sessionID := "sess-track-current-target-event-sibling-provider"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}

	initial := TrackSharedSessionBrowserCurrentTargetEvent(
		sessionRegistry,
		runRegistryA,
		nil,
		sessionID,
		route,
		"https://example.com/one",
		"One",
		"browser_navigate",
		time.Minute,
	)
	if initial.ID == "" {
		t.Fatalf("expected initial top-level current-target event to create tracked target, got %#v", initial)
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
		t.Fatalf("expected first sibling watch loop to expose tracked current target, got %#v", initialA.View.Session.Routes)
	}
	if len(initialB.View.Session.Routes) != 1 || initialB.View.Session.Routes[0].CurrentTargetID != initial.ID {
		t.Fatalf("expected second sibling watch loop to expose tracked current target, got %#v", initialB.View.Session.Routes)
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

	updated := TrackSharedSessionBrowserCurrentTargetEvent(
		sessionRegistry,
		runRegistryA,
		nil,
		sessionID,
		route,
		"https://example.com/two",
		"Two",
		"browser_navigate",
		time.Minute,
	)
	if updated.ID == "" {
		t.Fatalf("expected current-target event to return tracked target, got %#v", updated)
	}

	seededB := boundB.ObserveWatchLoop(context.Background(), req)
	if len(seededB.View.Session.Routes) != 1 || seededB.View.Session.Routes[0].CurrentTargetID != updated.ID {
		t.Fatalf("expected sibling watch loop to reuse primary provider source for updated current target, got %#v", seededB.View.Session.Routes)
	}
	if len(seededB.View.Session.Routes[0].Targets) != 1 || seededB.View.Session.Routes[0].Targets[0].Title != "Two" || seededB.View.Session.Routes[0].Targets[0].URL != "https://example.com/two" {
		t.Fatalf("expected sibling watch loop to expose updated current target payload, got %#v", seededB.View.Session.Routes)
	}
	if len(backend.statusReqs) != 2 || len(backend.profilesReqs) != 2 {
		t.Fatalf("expected sibling watch loop to reuse seeded source without extra polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}

func TestTrackSharedSessionBrowserCurrentTargetEventSeedsPrimaryAndSiblingProvidersFromStateSnapshotWithoutPriorPolling(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	sessionID := "sess-track-current-target-event-state-snapshot-source"
	observedAt := time.Date(2026, time.March, 30, 8, 25, 0, 0, time.UTC)
	stateRegistry.RecordBrowserProfileState(sessionID, SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "connected",
		Running:       true,
		Connected:     true,
		ObservedAt:    observedAt,
	})

	runRegistryA := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			sessionID: {{RunID: "run-a", Status: "running"}},
		}},
	}
	runRegistryB := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			sessionID: {{RunID: "run-b", Status: "running"}},
		}},
	}
	managerA := SharedSessionBrowserObserverManagerFor(sessionRegistry, runRegistryA, stateRegistry, time.Minute)
	managerB := SharedSessionBrowserObserverManagerFor(sessionRegistry, runRegistryB, stateRegistry, time.Minute)
	backend := &statusProfilesObservationTestBackend{
		statusResp: BrowserProfileStatusResult{
			Profile:   "isolated",
			Status:    "connected",
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
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}

	target := TrackSharedSessionBrowserCurrentTargetEvent(
		sessionRegistry,
		runRegistryA,
		stateRegistry,
		sessionID,
		route,
		"https://example.com/one",
		"One",
		"browser_navigate",
		time.Minute,
	)
	if target.ID == "" {
		t.Fatalf("expected top-level current-target event to create tracked target, got %#v", target)
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

	seededA := boundA.ObserveWatchLoop(context.Background(), req)
	seededB := boundB.ObserveWatchLoop(context.Background(), req)
	if len(seededA.View.Session.Routes) != 1 || seededA.View.Session.Routes[0].CurrentTargetID != target.ID {
		t.Fatalf("expected primary provider watch loop to reuse state-snapshot source for current target, got %#v", seededA.View.Session.Routes)
	}
	if len(seededB.View.Session.Routes) != 1 || seededB.View.Session.Routes[0].CurrentTargetID != target.ID {
		t.Fatalf("expected sibling provider watch loop to reuse state-snapshot source for current target, got %#v", seededB.View.Session.Routes)
	}
	if backend.statusRequestCount() != 0 || backend.profilesRequestCount() != 0 {
		t.Fatalf("expected state-snapshot source to avoid backend polling entirely, got status=%d profiles=%d", backend.statusRequestCount(), backend.profilesRequestCount())
	}
}

func TestTrackSharedSessionBrowserCurrentTargetEventSeedsSiblingRawRouteMutationSourceWhenRouteMutationCycleDrained(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	sessionID := "sess-track-current-target-event-raw-route-mutation-source"
	observedAt := time.Date(2026, time.March, 30, 8, 26, 0, 0, time.UTC)
	stateRegistry.RecordBrowserProfileState(sessionID, SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "connected",
		Running:       true,
		Connected:     true,
		ObservedAt:    observedAt,
	})

	runRegistryA := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			sessionID: {{RunID: "run-a", Status: "running"}},
		}},
	}
	runRegistryB := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			sessionID: {{RunID: "run-b", Status: "running"}},
		}},
	}
	managerA := SharedSessionBrowserObserverManagerFor(sessionRegistry, runRegistryA, stateRegistry, time.Minute)
	managerB := SharedSessionBrowserObserverManagerFor(sessionRegistry, runRegistryB, stateRegistry, time.Minute)
	backend := &statusProfilesObservationTestBackend{}
	boundA := managerA.Bind(backend)
	boundB := managerB.Bind(backend)
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}

	target := TrackSharedSessionBrowserCurrentTargetEvent(
		sessionRegistry,
		runRegistryA,
		stateRegistry,
		sessionID,
		route,
		"https://example.com/one",
		"One",
		"browser_navigate",
		time.Minute,
	)
	if target.ID == "" {
		t.Fatalf("expected top-level current-target event to create tracked target, got %#v", target)
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

	boundB.state.mu.Lock()
	clear(boundB.state.rawStatus)
	clear(boundB.state.rawProfiles)
	clear(boundB.state.routeMutations)
	clear(boundB.state.eventCycles)
	clear(boundB.state.bindings)
	clear(boundB.state.views)
	clear(boundB.state.watchLoops)
	clear(boundB.state.eventCyclesInFlight)
	clear(boundB.state.bindingsInFlight)
	clear(boundB.state.viewsInFlight)
	clear(boundB.state.watchLoopsInFlight)
	rawRouteMutationCount := len(boundB.state.rawRouteMutations)
	boundB.state.mu.Unlock()
	if rawRouteMutationCount == 0 {
		t.Fatalf("expected sibling provider raw route-mutation source before draining route-mutation cycle cache")
	}

	seededB := boundB.ObserveWatchLoop(context.Background(), req)
	if len(seededB.View.Session.Routes) != 1 || seededB.View.Session.Routes[0].CurrentTargetID != target.ID {
		t.Fatalf("expected sibling watch loop to reuse raw route-mutation source for current target, got %#v", seededB.View.Session.Routes)
	}
	if backend.statusRequestCount() != 0 || backend.profilesRequestCount() != 0 {
		t.Fatalf("expected sibling raw route-mutation source to avoid backend polling, got status=%d profiles=%d", backend.statusRequestCount(), backend.profilesRequestCount())
	}
}

func TestTrackSharedSessionBrowserCurrentTargetEventSeedsSiblingProviderFromPrimaryRouteMutationSourceWhenPrimaryCachesDrained(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	runRegistryA := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-track-current-target-event-primary-route-mutation-source": {{RunID: "run-a", Status: "running"}},
		}},
	}
	runRegistryB := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-track-current-target-event-primary-route-mutation-source": {{RunID: "run-b", Status: "running"}},
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
	sessionID := "sess-track-current-target-event-primary-route-mutation-source"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}

	initial := TrackSharedSessionBrowserCurrentTargetEvent(
		sessionRegistry,
		runRegistryA,
		nil,
		sessionID,
		route,
		"https://example.com/one",
		"One",
		"browser_navigate",
		time.Minute,
	)
	if initial.ID == "" {
		t.Fatalf("expected initial top-level current-target event to create tracked target, got %#v", initial)
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
		t.Fatalf("expected first sibling watch loop to expose tracked current target, got %#v", initialA.View.Session.Routes)
	}
	if len(initialB.View.Session.Routes) != 1 || initialB.View.Session.Routes[0].CurrentTargetID != initial.ID {
		t.Fatalf("expected second sibling watch loop to expose tracked current target, got %#v", initialB.View.Session.Routes)
	}
	if len(backend.statusReqs) != 2 || len(backend.profilesReqs) != 2 {
		t.Fatalf("expected initial sibling watch loops to poll backend once each, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}

	updated := TrackSharedSessionBrowserCurrentTargetEvent(
		sessionRegistry,
		runRegistryA,
		nil,
		sessionID,
		route,
		"https://example.com/two",
		"Two",
		"browser_navigate",
		time.Minute,
	)
	if updated.ID == "" {
		t.Fatalf("expected second current-target event to return tracked target, got %#v", updated)
	}

	boundA.state.mu.Lock()
	clear(boundA.state.rawStatus)
	clear(boundA.state.rawProfiles)
	clear(boundA.state.eventCycles)
	clear(boundA.state.bindings)
	clear(boundA.state.views)
	clear(boundA.state.watchLoops)
	clear(boundA.state.eventCyclesInFlight)
	clear(boundA.state.bindingsInFlight)
	clear(boundA.state.viewsInFlight)
	clear(boundA.state.watchLoopsInFlight)
	routeMutationCountA := len(boundA.state.routeMutations)
	boundA.state.mu.Unlock()
	if routeMutationCountA == 0 {
		t.Fatalf("expected primary provider to retain route-mutation source before draining event/raw caches")
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

	third := TrackSharedSessionBrowserCurrentTargetEvent(
		sessionRegistry,
		runRegistryA,
		nil,
		sessionID,
		route,
		"https://example.com/three",
		"Three",
		"browser_navigate",
		time.Minute,
	)
	if third.ID == "" {
		t.Fatalf("expected third current-target event to return tracked target, got %#v", third)
	}

	seededB := boundB.ObserveWatchLoop(context.Background(), req)
	if len(seededB.View.Session.Routes) != 1 || seededB.View.Session.Routes[0].CurrentTargetID != third.ID {
		t.Fatalf("expected sibling watch loop to reuse primary route-mutation source for updated current target, got %#v", seededB.View.Session.Routes)
	}
	if len(seededB.View.Session.Routes[0].Targets) != 1 || seededB.View.Session.Routes[0].Targets[0].Title != "Three" || seededB.View.Session.Routes[0].Targets[0].URL != "https://example.com/three" {
		t.Fatalf("expected sibling watch loop to expose updated current target payload, got %#v", seededB.View.Session.Routes)
	}
	if len(backend.statusReqs) != 2 || len(backend.profilesReqs) != 2 {
		t.Fatalf("expected sibling watch loop to reuse primary route-mutation source without extra polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}

func TestRecordSharedSessionBrowserPendingTargetPopupReviewEventSeedsSiblingRawRouteMutationSourceWhenRouteMutationCycleDrained(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	sessionID := "sess-pending-popup-review-event-raw-route-mutation-source"
	observedAt := time.Date(2026, time.March, 30, 8, 27, 0, 0, time.UTC)
	stateRegistry.RecordBrowserProfileState(sessionID, SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "connected",
		Running:       true,
		Connected:     true,
		ObservedAt:    observedAt,
	})

	runRegistryA := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			sessionID: {{RunID: "run-a", Status: "running"}},
		}},
	}
	runRegistryB := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			sessionID: {{RunID: "run-b", Status: "running"}},
		}},
	}
	managerA := SharedSessionBrowserObserverManagerFor(sessionRegistry, runRegistryA, stateRegistry, time.Minute)
	managerB := SharedSessionBrowserObserverManagerFor(sessionRegistry, runRegistryB, stateRegistry, time.Minute)
	backend := &statusProfilesObservationTestBackend{}
	boundA := managerA.Bind(backend)
	boundB := managerB.Bind(backend)
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}

	popup := TrackSharedSessionBrowserTabEvent(
		sessionRegistry,
		runRegistryA,
		stateRegistry,
		sessionID,
		route,
		BrowserTab{Index: 2, URL: "https://popup.example/offer", Title: "Offer"},
		false,
		time.Minute,
	)
	if popup.ID == "" {
		t.Fatalf("expected top-level tab event to create popup target, got %#v", popup)
	}
	review := RecordSharedSessionBrowserPendingTargetPopupReviewEvent(
		sessionRegistry,
		runRegistryA,
		stateRegistry,
		sessionID,
		route,
		BrowserTab{Index: popup.TabIndex, URL: popup.URL, Title: popup.Title, TargetID: popup.ID},
		"session_target_popup_review_required",
		"pending popup review",
		time.Minute,
	)
	if review == nil || review.ID != popup.ID {
		t.Fatalf("expected top-level popup review event to resolve tracked popup target, got %#v", review)
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

	boundB.state.mu.Lock()
	clear(boundB.state.rawStatus)
	clear(boundB.state.rawProfiles)
	clear(boundB.state.routeMutations)
	clear(boundB.state.eventCycles)
	clear(boundB.state.bindings)
	clear(boundB.state.views)
	clear(boundB.state.watchLoops)
	clear(boundB.state.eventCyclesInFlight)
	clear(boundB.state.bindingsInFlight)
	clear(boundB.state.viewsInFlight)
	clear(boundB.state.watchLoopsInFlight)
	rawRouteMutationCount := len(boundB.state.rawRouteMutations)
	boundB.state.mu.Unlock()
	if rawRouteMutationCount == 0 {
		t.Fatalf("expected sibling provider raw route-mutation source before draining route-mutation cycle cache")
	}

	seededB := boundB.ObserveWatchLoop(context.Background(), req)
	if len(seededB.View.Session.Routes) != 1 || seededB.View.Session.Routes[0].PendingTargetReview == nil || seededB.View.Session.Routes[0].PendingTargetReview.ID != popup.ID {
		t.Fatalf("expected sibling watch loop to reuse raw route-mutation source for pending popup review, got %#v", seededB.View.Session.Routes)
	}
	if backend.statusRequestCount() != 0 || backend.profilesRequestCount() != 0 {
		t.Fatalf("expected sibling raw route-mutation source to avoid backend polling, got status=%d profiles=%d", backend.statusRequestCount(), backend.profilesRequestCount())
	}
}

func TestRestoreSharedSessionBrowserCurrentTargetSelectionForPendingReviewEventSeedsSiblingRawRouteMutationSourceWhenRouteMutationCycleDrained(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	sessionID := "sess-restore-pending-review-event-raw-route-mutation-source"
	observedAt := time.Date(2026, time.March, 30, 8, 28, 0, 0, time.UTC)
	stateRegistry.RecordBrowserProfileState(sessionID, SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "connected",
		Running:       true,
		Connected:     true,
		ObservedAt:    observedAt,
	})

	runRegistryA := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			sessionID: {{RunID: "run-a", Status: "running"}},
		}},
	}
	runRegistryB := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			sessionID: {{RunID: "run-b", Status: "running"}},
		}},
	}
	managerA := SharedSessionBrowserObserverManagerFor(sessionRegistry, runRegistryA, stateRegistry, time.Minute)
	managerB := SharedSessionBrowserObserverManagerFor(sessionRegistry, runRegistryB, stateRegistry, time.Minute)
	backend := &statusProfilesObservationTestBackend{}
	boundA := managerA.Bind(backend)
	boundB := managerB.Bind(backend)
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}

	prior := TrackSharedSessionBrowserTabEvent(
		sessionRegistry,
		runRegistryA,
		stateRegistry,
		sessionID,
		route,
		BrowserTab{Index: 1, URL: "https://example.com/one", Title: "One"},
		true,
		time.Minute,
	)
	if prior.ID == "" {
		t.Fatalf("expected top-level tab event to create prior target, got %#v", prior)
	}
	priorSelection := SnapshotSharedSessionBrowserCurrentTargetSelection(sessionRegistry, sessionID, route)
	if priorSelection == nil || priorSelection.ID != prior.ID {
		t.Fatalf("expected prior current-target snapshot to match first target, got %#v", priorSelection)
	}
	pending := TrackSharedSessionBrowserTabEvent(
		sessionRegistry,
		runRegistryA,
		stateRegistry,
		sessionID,
		route,
		BrowserTab{Index: 2, URL: "https://popup.example/offer", Title: "Offer"},
		true,
		time.Minute,
	)
	if pending.ID == "" {
		t.Fatalf("expected second top-level tab event to create pending target, got %#v", pending)
	}
	review := RecordSharedSessionBrowserPendingTargetPopupReviewEvent(
		sessionRegistry,
		runRegistryA,
		stateRegistry,
		sessionID,
		route,
		BrowserTab{Index: pending.TabIndex, URL: pending.URL, Title: pending.Title, TargetID: pending.ID},
		"session_target_popup_review_required",
		"pending popup review",
		time.Minute,
	)
	if review == nil || review.ID != pending.ID {
		t.Fatalf("expected popup review event before restore, got %#v", review)
	}
	restored := RestoreSharedSessionBrowserCurrentTargetSelectionForPendingReviewEvent(
		sessionRegistry,
		runRegistryA,
		stateRegistry,
		sessionID,
		route,
		priorSelection,
		pending.ID,
		"popup_review_restore",
		time.Minute,
	)
	if restored == nil || restored.ID != prior.ID {
		t.Fatalf("expected top-level restore event to switch current selection back to prior target, got %#v", restored)
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

	boundB.state.mu.Lock()
	clear(boundB.state.rawStatus)
	clear(boundB.state.rawProfiles)
	clear(boundB.state.routeMutations)
	clear(boundB.state.eventCycles)
	clear(boundB.state.bindings)
	clear(boundB.state.views)
	clear(boundB.state.watchLoops)
	clear(boundB.state.eventCyclesInFlight)
	clear(boundB.state.bindingsInFlight)
	clear(boundB.state.viewsInFlight)
	clear(boundB.state.watchLoopsInFlight)
	rawRouteMutationCount := len(boundB.state.rawRouteMutations)
	boundB.state.mu.Unlock()
	if rawRouteMutationCount == 0 {
		t.Fatalf("expected sibling provider raw route-mutation source before draining route-mutation cycle cache")
	}

	seededB := boundB.ObserveWatchLoop(context.Background(), req)
	if len(seededB.View.Session.Routes) != 1 || seededB.View.Session.Routes[0].CurrentTargetID != prior.ID {
		t.Fatalf("expected sibling watch loop to reuse raw route-mutation source for prior current target restore, got %#v", seededB.View.Session.Routes)
	}
	if seededB.View.Session.Routes[0].PendingTargetReview == nil || seededB.View.Session.Routes[0].PendingTargetReview.ID != pending.ID {
		t.Fatalf("expected sibling watch loop to preserve pending popup review after restore, got %#v", seededB.View.Session.Routes)
	}
	if backend.statusRequestCount() != 0 || backend.profilesRequestCount() != 0 {
		t.Fatalf("expected sibling raw route-mutation source to avoid backend polling, got status=%d profiles=%d", backend.statusRequestCount(), backend.profilesRequestCount())
	}
}

func TestApplySharedSessionBrowserExecutionResultEventSeedsSiblingProviderFromPrimaryCachedEventCycle(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	runRegistryA := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-apply-execution-event-sibling-provider": {{RunID: "run-a", Status: "running"}},
		}},
	}
	runRegistryB := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-apply-execution-event-sibling-provider": {{RunID: "run-b", Status: "running"}},
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
	sessionID := "sess-apply-execution-event-sibling-provider"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node"}

	tracked := sessionRegistry.TrackTab(sessionID, BrowserSessionTarget{
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

	initialA := boundA.ObserveWatchLoop(context.Background(), req)
	initialB := boundB.ObserveWatchLoop(context.Background(), req)
	if len(initialA.View.Session.Routes) != 1 || initialA.View.Session.Routes[0].CurrentTargetID != tracked.ID {
		t.Fatalf("expected first provider to expose tracked current target, got %#v", initialA.View.Session.Routes)
	}
	if len(initialB.View.Session.Routes) != 1 || initialB.View.Session.Routes[0].CurrentTargetID != tracked.ID {
		t.Fatalf("expected second provider to expose tracked current target, got %#v", initialB.View.Session.Routes)
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

	application := ApplySharedSessionBrowserExecutionResultEvent(
		sessionRegistry,
		runRegistryA,
		nil,
		sessionID,
		BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
		"",
		SharedSessionBrowserExecutionResult{
			Profile: "isolated",
			ProfileStatus: BrowserProfileStatusResult{
				Backend:   "proxy",
				Profile:   "isolated",
				Status:    "stopped",
				Running:   false,
				Connected: false,
			},
			Decision:                 "stopped",
			InvalidateSessionTargets: true,
		},
		time.Minute,
	)
	if application.Cleanup.ClearedSessionTargets != 1 {
		t.Fatalf("expected top-level execution event cleanup to clear tracked route target, got %#v", application.Cleanup)
	}

	seededB := boundB.ObserveWatchLoop(context.Background(), req)
	if len(seededB.View.Session.Routes) > 1 {
		t.Fatalf("expected sibling watch loop to preserve a single refreshed route snapshot, got %#v", seededB.View.Session.Routes)
	}
	if len(seededB.View.Session.Routes) == 1 && seededB.View.Session.Routes[0].CurrentTargetID != "" {
		t.Fatalf("expected sibling watch loop to reuse execution-event-seeded cleared target, got %#v", seededB.View.Session.Routes)
	}
	if len(backend.statusReqs) != 2 || len(backend.profilesReqs) != 2 {
		t.Fatalf("expected sibling watch loop to reuse top-level execution event source without extra polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}

func TestSyncSharedSessionBrowserRouteSelectionEventSeedsSiblingProviderFromPrimaryCachedEventCycle(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	runRegistryA := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-sync-route-selection-event-sibling-provider": {{RunID: "run-a", Status: "running"}},
		}},
	}
	runRegistryB := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-sync-route-selection-event-sibling-provider": {{RunID: "run-b", Status: "running"}},
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
	sessionID := "sess-sync-route-selection-event-sibling-provider"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "workbench", Target: "node", BrowserApp: "Chromium"}
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

	result, err := SyncSharedSessionBrowserRouteSelectionEvent(
		context.Background(),
		sessionRegistry,
		runRegistryA,
		stateRegistry,
		sessionID,
		BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
		route,
		"",
		nil,
		false,
		"sync_session",
		time.Minute,
	)
	if err != nil {
		t.Fatalf("SyncSharedSessionBrowserRouteSelectionEvent returned error: %v", err)
	}
	if !result.Ready || result.Decision != "session_route_synced" {
		t.Fatalf("unexpected top-level route sync result: %#v", result)
	}
	if result.ProfileSelection == nil || result.TargetSelection == nil || result.TargetSelection.ID != tracked[0].TargetID {
		t.Fatalf("expected top-level route sync to persist profile and target selections, got %#v", result)
	}

	seededB := boundB.ObserveWatchLoop(context.Background(), req)
	if len(seededB.View.Session.Routes) != 1 ||
		seededB.View.Session.Routes[0].CurrentTargetID != tracked[0].TargetID ||
		seededB.View.Session.Routes[0].CurrentTargetSource != "sync_session" {
		t.Fatalf("expected sibling watch loop to reuse top-level sync_session target, got %#v", seededB.View.Session.Routes)
	}
	if stored, ok := stateRegistry.SelectedBrowserProfile(sessionID, "node"); !ok || stored.Profile != "workbench" || stored.Source != "sync_session" {
		t.Fatalf("expected top-level route sync to persist selected profile, got %#v ok=%v", stored, ok)
	}
	if len(backend.statusReqs) != 2 || len(backend.profilesReqs) != 2 {
		t.Fatalf("expected top-level route sync to avoid extra polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}

func TestCoordinateSharedSessionBrowserRouteSyncEventSeedsSiblingProviderFromPrimaryCachedEventCycle(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	runRegistryA := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-coordinate-route-sync-event-sibling-provider": {{RunID: "run-a", Status: "running"}},
		}},
	}
	runRegistryB := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-coordinate-route-sync-event-sibling-provider": {{RunID: "run-b", Status: "running"}},
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
	sessionID := "sess-coordinate-route-sync-event-sibling-provider"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "workbench", Target: "node", BrowserApp: "Chromium"}
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
		t.Fatalf("expected first provider to expose tracked_active_tab before coordination sync, got %#v", initialA.View.Session.Routes)
	}
	if len(initialB.View.Session.Routes) != 1 || initialB.View.Session.Routes[0].CurrentTargetSource != "tracked_active_tab" {
		t.Fatalf("expected second provider to expose tracked_active_tab before coordination sync, got %#v", initialB.View.Session.Routes)
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

	result, err := CoordinateSharedSessionBrowserRouteSyncEvent(
		context.Background(),
		sessionRegistry,
		runRegistryA,
		stateRegistry,
		sessionID,
		BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
		route,
		"",
		nil,
		false,
		time.Minute,
	)
	if err != nil {
		t.Fatalf("CoordinateSharedSessionBrowserRouteSyncEvent returned error: %v", err)
	}
	if !result.Ready || result.Decision != "session_target_synced" {
		t.Fatalf("unexpected top-level coordination sync result: %#v", result)
	}
	if result.ProfileSelection != nil {
		t.Fatalf("expected top-level coordination sync to avoid profile selection churn, got %#v", result.ProfileSelection)
	}
	if result.TargetSelection == nil || result.TargetSelection.ID != tracked[0].TargetID || result.TargetSelection.Source != "sync_session" {
		t.Fatalf("unexpected top-level coordination target selection: %#v", result.TargetSelection)
	}

	seededB := boundB.ObserveWatchLoop(context.Background(), req)
	if len(seededB.View.Session.Routes) != 1 ||
		seededB.View.Session.Routes[0].CurrentTargetID != tracked[0].TargetID ||
		seededB.View.Session.Routes[0].CurrentTargetSource != "sync_session" {
		t.Fatalf("expected sibling watch loop to reuse top-level coordination sync target, got %#v", seededB.View.Session.Routes)
	}
	if len(backend.statusReqs) != 2 || len(backend.profilesReqs) != 2 {
		t.Fatalf("expected top-level coordination sync to avoid extra polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}

func TestApplySharedSessionBrowserResolvedTargetEventSeedsSiblingProviderFromPrimaryCachedEventCycle(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	runRegistryA := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-apply-resolved-target-event-sibling-provider": {{RunID: "run-a", Status: "running"}},
		}},
	}
	runRegistryB := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-apply-resolved-target-event-sibling-provider": {{RunID: "run-b", Status: "running"}},
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
	sessionID := "sess-apply-resolved-target-event-sibling-provider"
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

	result := ApplySharedSessionBrowserResolvedTargetEvent(
		sessionRegistry,
		runRegistryA,
		nil,
		SharedSessionBrowserResolvedTargetEventRequest{
			SessionID: sessionID,
			Route:     route,
			URL:       "https://example.com/opened",
			Source:    "browser_open",
		},
		time.Minute,
	)
	if strings.TrimSpace(result.TargetID) == "" {
		t.Fatalf("expected resolved-target event to track current target, got %#v", result)
	}

	seededB := boundB.ObserveWatchLoop(context.Background(), req)
	if len(seededB.View.Session.Routes) != 1 || seededB.View.Session.Routes[0].CurrentTargetID != result.TargetID {
		t.Fatalf("expected sibling watch loop to reuse resolved-target event current target, got %#v", seededB.View.Session.Routes)
	}
	if len(backend.statusReqs) != 2 || len(backend.profilesReqs) != 2 {
		t.Fatalf("expected sibling watch loop to reuse resolved-target source without extra polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}

func TestApplySharedSessionBrowserTargetEventSeedsSiblingProviderFromPrimaryCachedEventCycle(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	runRegistryA := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-apply-target-event-sibling-provider": {{RunID: "run-a", Status: "running"}},
		}},
	}
	runRegistryB := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-apply-target-event-sibling-provider": {{RunID: "run-b", Status: "running"}},
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
	sessionID := "sess-apply-target-event-sibling-provider"
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

	result := ApplySharedSessionBrowserTargetEvent(
		sessionRegistry,
		runRegistryA,
		nil,
		SharedSessionBrowserTargetEventRequest{
			SessionID:  sessionID,
			Route:      route,
			TabIndex:   2,
			URL:        "https://example.com/second",
			Title:      "Second",
			Source:     "browser_act_download",
			SetCurrent: true,
		},
		time.Minute,
	)
	if strings.TrimSpace(result.TargetID) == "" || result.Target.TabIndex != 2 {
		t.Fatalf("expected generic target event to track tab handle, got %#v", result)
	}

	seededB := boundB.ObserveWatchLoop(context.Background(), req)
	if len(seededB.View.Session.Routes) != 1 || seededB.View.Session.Routes[0].CurrentTargetID != result.TargetID {
		t.Fatalf("expected sibling watch loop to reuse generic target event current target, got %#v", seededB.View.Session.Routes)
	}
	if len(seededB.View.Session.Routes[0].Targets) != 1 || seededB.View.Session.Routes[0].Targets[0].TabIndex != 2 {
		t.Fatalf("expected sibling watch loop to expose generic target event tracked tab, got %#v", seededB.View.Session.Routes)
	}
	if len(backend.statusReqs) != 2 || len(backend.profilesReqs) != 2 {
		t.Fatalf("expected sibling watch loop to reuse generic target event source without extra polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}

func TestApplySharedSessionBrowserResolvedTargetEventRecordsPendingReviewAndRestoresSiblingProviderFromPrimaryCachedEventCycle(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	runRegistryA := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-apply-resolved-target-review-event-sibling-provider": {{RunID: "run-a", Status: "running"}},
		}},
	}
	runRegistryB := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-apply-resolved-target-review-event-sibling-provider": {{RunID: "run-b", Status: "running"}},
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
	sessionID := "sess-apply-resolved-target-review-event-sibling-provider"
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
		t.Fatalf("expected first provider to expose home target before review event, got %#v", firstA.View.Session.Routes)
	}
	if len(firstB.View.Session.Routes) != 1 || firstB.View.Session.Routes[0].CurrentTargetID != home.ID {
		t.Fatalf("expected second provider to expose home target before review event, got %#v", firstB.View.Session.Routes)
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

	result := ApplySharedSessionBrowserResolvedTargetEvent(
		sessionRegistry,
		runRegistryA,
		nil,
		SharedSessionBrowserResolvedTargetEventRequest{
			SessionID:             sessionID,
			Route:                 route,
			URL:                   "https://example.com/redirected",
			Title:                 "Redirected",
			Source:                "browser_navigate",
			PendingReview:         true,
			PendingReviewDecision: "session_target_redirect_review_required",
			PendingReviewReason:   "redirect review",
			PriorSelection:        SnapshotSharedSessionBrowserCurrentTargetSelection(sessionRegistry, sessionID, route),
		},
		time.Minute,
	)
	if strings.TrimSpace(result.TargetID) == "" || result.Review == nil {
		t.Fatalf("expected resolved-target review event to track target and record review, got %#v", result)
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
		t.Fatalf("expected sibling watch loop to reuse resolved-target review source without extra polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}

func TestApplySharedSessionBrowserTabsResultEventSeedsSiblingProviderAndPreservesClosedTargetID(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	runRegistryA := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-apply-tabs-result-event-close-sibling-provider": {{RunID: "run-a", Status: "running"}},
		}},
	}
	runRegistryB := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-apply-tabs-result-event-close-sibling-provider": {{RunID: "run-b", Status: "running"}},
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
	sessionID := "sess-apply-tabs-result-event-close-sibling-provider"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}
	trackedTabs := SyncSharedSessionBrowserTabsForRouteEvent(
		sessionRegistry,
		runRegistryA,
		nil,
		sessionID,
		route,
		2,
		[]BrowserTab{
			{Index: 1, URL: "https://example.com/home", Title: "Home"},
			{Index: 2, URL: "https://example.com/popup", Title: "Popup"},
		},
		time.Minute,
	)
	homeID := strings.TrimSpace(trackedTabs[0].TargetID)
	popupID := strings.TrimSpace(trackedTabs[1].TargetID)
	if homeID == "" || popupID == "" {
		t.Fatalf("expected initial tab sync to seed target ids, got %#v", trackedTabs)
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

	result := ApplySharedSessionBrowserTabsResultEvent(
		sessionRegistry,
		runRegistryA,
		nil,
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
		time.Minute,
	)
	if result.TargetID != popupID {
		t.Fatalf("expected top-level tabs result to preserve closed tab target id %q, got %#v", popupID, result)
	}
	if len(result.Tabs) != 1 || result.Tabs[0].Index != 1 {
		t.Fatalf("expected top-level tabs result to keep synced tab snapshot, got %#v", result.Tabs)
	}

	seededB := boundB.ObserveWatchLoop(context.Background(), req)
	if len(seededB.View.Session.Routes) != 1 || seededB.View.Session.Routes[0].CurrentTargetID != homeID {
		t.Fatalf("expected sibling watch loop to reuse close event source and keep home current, got %#v", seededB.View.Session.Routes)
	}
	if len(backend.statusReqs) != 2 || len(backend.profilesReqs) != 2 {
		t.Fatalf("expected sibling watch loop to reuse tabs-result source without extra polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}

func TestApplySharedSessionBrowserTabsResultEventRestoresSiblingProviderDuringRememberReview(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	runRegistryA := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-apply-tabs-result-event-review-sibling-provider": {{RunID: "run-a", Status: "running"}},
		}},
	}
	runRegistryB := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-apply-tabs-result-event-review-sibling-provider": {{RunID: "run-b", Status: "running"}},
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
	sessionID := "sess-apply-tabs-result-event-review-sibling-provider"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}
	trackedTabs := SyncSharedSessionBrowserTabsForRouteEvent(
		sessionRegistry,
		runRegistryA,
		nil,
		sessionID,
		route,
		1,
		[]BrowserTab{
			{Index: 1, URL: "https://example.com/home", Title: "Home"},
		},
		time.Minute,
	)
	homeID := strings.TrimSpace(trackedTabs[0].TargetID)
	if homeID == "" {
		t.Fatalf("expected initial tab sync to seed home target id, got %#v", trackedTabs)
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

	result := ApplySharedSessionBrowserTabsResultEvent(
		sessionRegistry,
		runRegistryA,
		nil,
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
		time.Minute,
	)
	if result.RememberReview.Decision != "session_target_popup_review_required" || result.RememberReview.Ready {
		t.Fatalf("expected top-level tabs result to require popup remember review, got %#v", result)
	}
	if strings.TrimSpace(result.TargetID) == "" {
		t.Fatalf("expected top-level tabs result to resolve active popup target id, got %#v", result)
	}

	seededB := boundB.ObserveWatchLoop(context.Background(), req)
	if len(seededB.View.Session.Routes) != 1 {
		t.Fatalf("expected sibling watch loop to preserve refreshed route snapshot, got %#v", seededB.View.Session.Routes)
	}
	if seededB.View.Session.Routes[0].CurrentTargetID != homeID {
		t.Fatalf("expected top-level tabs result to restore prior current target, got %#v", seededB.View.Session.Routes)
	}
	if seededB.View.Session.Routes[0].PendingTargetReview == nil || seededB.View.Session.Routes[0].PendingTargetReview.ID != strings.TrimSpace(result.TargetID) {
		t.Fatalf("expected top-level tabs result to record popup review on sibling provider, got %#v", seededB.View.Session.Routes)
	}
	if len(backend.statusReqs) != 2 || len(backend.profilesReqs) != 2 {
		t.Fatalf("expected sibling watch loop to reuse tabs-result review source without extra polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}

func TestApplySharedSessionBrowserTabsResultEventCarriesFocusReviewPostureFromPendingReview(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	runRegistryA := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-apply-tabs-result-event-focus-review-sibling-provider": {{RunID: "run-a", Status: "running"}},
		}},
	}
	runRegistryB := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-apply-tabs-result-event-focus-review-sibling-provider": {{RunID: "run-b", Status: "running"}},
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
	sessionID := "sess-apply-tabs-result-event-focus-review-sibling-provider"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}
	trackedTabs := SyncSharedSessionBrowserTabsForRouteEvent(
		sessionRegistry,
		runRegistryA,
		nil,
		sessionID,
		route,
		1,
		[]BrowserTab{
			{Index: 1, URL: "https://example.com/home", Title: "Home"},
			{Index: 2, URL: "https://popup.example/offer", Title: "Offer"},
		},
		time.Minute,
	)
	popupID := strings.TrimSpace(trackedTabs[1].TargetID)
	if popupID == "" {
		t.Fatalf("expected initial tab sync to seed popup target id, got %#v", trackedTabs)
	}
	RecordSharedSessionBrowserPendingTargetReview(
		sessionRegistry,
		sessionID,
		route,
		popupID,
		2,
		"https://popup.example/offer",
		"Offer",
		"session_target_popup_review_required",
		"pending popup review",
	)
	focusReview := SharedSessionBrowserPendingTargetReviewStateForTarget(sessionRegistry, sessionID, route, "", 2)
	if focusReview.Review == nil {
		t.Fatalf("expected pending review for focused tab")
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

	result := ApplySharedSessionBrowserTabsResultEvent(
		sessionRegistry,
		runRegistryA,
		nil,
		SharedSessionBrowserTabsResultEventRequest{
			SessionID:         sessionID,
			Route:             route,
			Action:            "focus",
			RequestedTabIndex: 2,
			ActiveIndex:       2,
			Tabs: []BrowserTab{
				{Index: 1, URL: "https://example.com/home", Title: "Home"},
				{Index: 2, URL: "https://popup.example/offer", Title: "Offer"},
			},
			Force:  true,
			Review: focusReview,
			Actor:  "browser_act focus_tab",
		},
		time.Minute,
	)
	if result.TargetID != popupID || result.ReviewDecision != "session_target_popup_review_confirmed" || !result.ReviewReady {
		t.Fatalf("expected top-level tabs result to preserve focus review posture, got %#v", result)
	}
	if !strings.Contains(result.Note, "force=true") {
		t.Fatalf("expected forced focus review note, got %#v", result)
	}

	seededB := boundB.ObserveWatchLoop(context.Background(), req)
	if len(seededB.View.Session.Routes) != 1 || seededB.View.Session.Routes[0].CurrentTargetID != popupID {
		t.Fatalf("expected sibling watch loop to reuse focused popup target, got %#v", seededB.View.Session.Routes)
	}
	if len(backend.statusReqs) != 2 || len(backend.profilesReqs) != 2 {
		t.Fatalf("expected sibling watch loop to reuse focus review source without extra polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}

func TestApplySharedSessionBrowserTabRememberReviewEventRefreshesSharedProviderProjectionOnce(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	runRegistry := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-apply-tab-remember-review-event-shared-provider": {{RunID: "run-1", Status: "running"}},
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
	sessionID := "sess-apply-tab-remember-review-event-shared-provider"
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
		t.Fatalf("expected top-level tab sync to seed tracked tabs and prior selection, got first=%#v tracked=%#v prior=%#v", firstTabs, trackedTabs, priorSelection)
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

	rememberReview := ApplySharedSessionBrowserTabRememberReviewEvent(
		sessionRegistry,
		runRegistry,
		nil,
		SharedSessionBrowserTabRememberReviewRequest{
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
		},
		time.Minute,
	)
	if rememberReview.Decision != "session_target_popup_review_required" || rememberReview.Ready {
		t.Fatalf("expected top-level remember review to require popup confirmation, got %#v", rememberReview)
	}
	if runRegistry.callCount() != 2 {
		t.Fatalf("expected top-level remember review to refresh shared-provider projection once, got %d", runRegistry.callCount())
	}

	second := bound.ObserveWatchLoop(context.Background(), req)
	if len(second.View.Session.Routes) != 1 {
		t.Fatalf("expected refreshed watch loop to expose one route, got %#v", second.View.Session.Routes)
	}
	if second.View.Session.Routes[0].CurrentTargetID != strings.TrimSpace(priorSelection.ID) {
		t.Fatalf("expected top-level remember review to restore prior current target, got %#v", second.View.Session.Routes)
	}
	if second.View.Session.Routes[0].PendingTargetReview == nil || second.View.Session.Routes[0].PendingTargetReview.ID != strings.TrimSpace(trackedTabs[1].TargetID) {
		t.Fatalf("expected top-level remember review to record pending popup review, got %#v", second.View.Session.Routes)
	}
	if runRegistry.callCount() != 2 {
		t.Fatalf("expected next ObserveWatchLoop to reuse refreshed shared-provider projection cache, got %d", runRegistry.callCount())
	}
}

func TestApplySharedSessionBrowserTabRememberReviewEventSeedsSiblingRawRouteMutationSourceWhenRouteMutationCycleDrained(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	sessionID := "sess-apply-tab-remember-review-event-raw-route-mutation-source"
	observedAt := time.Date(2026, time.March, 30, 8, 29, 0, 0, time.UTC)
	stateRegistry.RecordBrowserProfileState(sessionID, SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "connected",
		Running:       true,
		Connected:     true,
		ObservedAt:    observedAt,
	})

	runRegistryA := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			sessionID: {{RunID: "run-a", Status: "running"}},
		}},
	}
	runRegistryB := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			sessionID: {{RunID: "run-b", Status: "running"}},
		}},
	}
	managerA := SharedSessionBrowserObserverManagerFor(sessionRegistry, runRegistryA, stateRegistry, time.Minute)
	managerB := SharedSessionBrowserObserverManagerFor(sessionRegistry, runRegistryB, stateRegistry, time.Minute)
	backend := &statusProfilesObservationTestBackend{}
	boundA := managerA.Bind(backend)
	boundB := managerB.Bind(backend)
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}

	firstTabs := SyncSharedSessionBrowserTabsForRouteEvent(
		sessionRegistry,
		runRegistryA,
		stateRegistry,
		sessionID,
		route,
		1,
		[]BrowserTab{{Index: 1, URL: "https://example.com/home", Title: "Home"}},
		time.Minute,
	)
	priorSelection := SnapshotSharedSessionBrowserCurrentTargetSelection(sessionRegistry, sessionID, route)
	trackedTabs := SyncSharedSessionBrowserTabsForRouteEvent(
		sessionRegistry,
		runRegistryA,
		stateRegistry,
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
		t.Fatalf("expected top-level tab sync to seed tracked tabs and prior selection, got first=%#v tracked=%#v prior=%#v", firstTabs, trackedTabs, priorSelection)
	}

	rememberReview := ApplySharedSessionBrowserTabRememberReviewEvent(
		sessionRegistry,
		runRegistryA,
		stateRegistry,
		SharedSessionBrowserTabRememberReviewRequest{
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
		},
		time.Minute,
	)
	if rememberReview.Decision != "session_target_popup_review_required" || rememberReview.Ready {
		t.Fatalf("expected top-level remember review to require popup confirmation, got %#v", rememberReview)
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

	boundB.state.mu.Lock()
	clear(boundB.state.rawStatus)
	clear(boundB.state.rawProfiles)
	clear(boundB.state.routeMutations)
	clear(boundB.state.eventCycles)
	clear(boundB.state.bindings)
	clear(boundB.state.views)
	clear(boundB.state.watchLoops)
	clear(boundB.state.eventCyclesInFlight)
	clear(boundB.state.bindingsInFlight)
	clear(boundB.state.viewsInFlight)
	clear(boundB.state.watchLoopsInFlight)
	rawRouteMutationCount := len(boundB.state.rawRouteMutations)
	boundB.state.mu.Unlock()
	if rawRouteMutationCount == 0 {
		t.Fatalf("expected sibling provider raw route-mutation source before draining route-mutation cycle cache")
	}

	seededB := boundB.ObserveWatchLoop(context.Background(), req)
	if len(seededB.View.Session.Routes) != 1 {
		t.Fatalf("expected sibling watch loop to expose one route, got %#v", seededB.View.Session.Routes)
	}
	if seededB.View.Session.Routes[0].CurrentTargetID != strings.TrimSpace(priorSelection.ID) {
		t.Fatalf("expected sibling watch loop to reuse raw route-mutation source for prior current target restore, got %#v", seededB.View.Session.Routes)
	}
	if seededB.View.Session.Routes[0].PendingTargetReview == nil || seededB.View.Session.Routes[0].PendingTargetReview.ID != strings.TrimSpace(trackedTabs[1].TargetID) {
		t.Fatalf("expected sibling watch loop to reuse raw route-mutation source for pending popup review, got %#v", seededB.View.Session.Routes)
	}
	if backend.statusRequestCount() != 0 || backend.profilesRequestCount() != 0 {
		t.Fatalf("expected sibling raw route-mutation source to avoid backend polling, got status=%d profiles=%d", backend.statusRequestCount(), backend.profilesRequestCount())
	}
}

func TestSelectSharedSessionBrowserTargetEventSeedsSiblingProviderFromPrimaryCachedEventCycle(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	runRegistryA := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-select-target-event-sibling-provider": {{RunID: "run-a", Status: "running"}},
		}},
	}
	runRegistryB := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-select-target-event-sibling-provider": {{RunID: "run-b", Status: "running"}},
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
	sessionID := "sess-select-target-event-sibling-provider"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}

	firstTarget := sessionRegistry.TrackTab(sessionID, BrowserSessionTarget{
		ID:         "tab-1",
		TabIndex:   1,
		URL:        "https://example.com/one",
		Title:      "One",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
	}, true)
	secondTarget := sessionRegistry.TrackTab(sessionID, BrowserSessionTarget{
		ID:         "tab-2",
		TabIndex:   2,
		URL:        "https://example.com/two",
		Title:      "Two",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
	}, false)
	if firstTarget.ID == "" || secondTarget.ID == "" {
		t.Fatalf("expected tracked targets to expose ids, got first=%#v second=%#v", firstTarget, secondTarget)
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
	if len(initialA.View.Session.Routes) != 1 || initialA.View.Session.Routes[0].CurrentTargetID != firstTarget.ID {
		t.Fatalf("expected first sibling watch loop to expose first current target, got %#v", initialA.View.Session.Routes)
	}
	if len(initialB.View.Session.Routes) != 1 || initialB.View.Session.Routes[0].CurrentTargetID != firstTarget.ID {
		t.Fatalf("expected second sibling watch loop to expose first current target, got %#v", initialB.View.Session.Routes)
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

	result, err := SelectSharedSessionBrowserTargetEvent(
		sessionRegistry,
		runRegistryA,
		nil,
		SharedSessionBrowserSelectTargetRequest{
			SessionID: sessionID,
			Route:     route,
			TargetID:  secondTarget.ID,
			Source:    "select_target",
			Actor:     "top-level select target",
		},
		time.Minute,
	)
	if err != nil {
		t.Fatalf("expected top-level target selection to succeed, got %v", err)
	}
	if !result.Ready || result.Selection == nil || result.Selection.ID != secondTarget.ID {
		t.Fatalf("expected top-level target selection to pick second target, got %#v", result)
	}

	seededB := boundB.ObserveWatchLoop(context.Background(), req)
	if len(seededB.View.Session.Routes) != 1 || seededB.View.Session.Routes[0].CurrentTargetID != secondTarget.ID {
		t.Fatalf("expected sibling watch loop to reuse primary provider source for selected target, got %#v", seededB.View.Session.Routes)
	}
	if len(seededB.View.Session.Routes[0].Targets) != 2 {
		t.Fatalf("expected sibling watch loop to preserve tracked targets, got %#v", seededB.View.Session.Routes)
	}
	if len(backend.statusReqs) != 2 || len(backend.profilesReqs) != 2 {
		t.Fatalf("expected sibling watch loop to reuse seeded source without extra polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}

func TestExecuteSharedSessionBrowserClearTargetEventSeedsSiblingProviderFromPrimaryCachedEventCycle(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	runRegistryA := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-clear-target-event-sibling-provider": {{RunID: "run-a", Status: "running"}},
		}},
	}
	runRegistryB := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-clear-target-event-sibling-provider": {{RunID: "run-b", Status: "running"}},
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
	sessionID := "sess-clear-target-event-sibling-provider"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}
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
	if tracked.ID == "" {
		t.Fatalf("expected tracked current target id, got %#v", tracked)
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
	if len(initialA.View.Session.Routes) != 1 || initialA.View.Session.Routes[0].CurrentTargetID != tracked.ID {
		t.Fatalf("expected first sibling watch loop to expose tracked target, got %#v", initialA.View.Session.Routes)
	}
	if len(initialB.View.Session.Routes) != 1 || initialB.View.Session.Routes[0].CurrentTargetID != tracked.ID {
		t.Fatalf("expected second sibling watch loop to expose tracked target, got %#v", initialB.View.Session.Routes)
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

	result := ExecuteSharedSessionBrowserClearTargetEvent(
		sessionRegistry,
		runRegistryA,
		nil,
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
		time.Minute,
	)
	if !result.Ready || !result.ClearedTargetSelection {
		t.Fatalf("expected top-level clear target event to clear current target, got %#v", result)
	}

	seededB := boundB.ObserveWatchLoop(context.Background(), req)
	if len(seededB.View.Session.Routes) != 1 || seededB.View.Session.Routes[0].CurrentTargetID != "" {
		t.Fatalf("expected sibling watch loop to reuse primary provider source for cleared target, got %#v", seededB.View.Session.Routes)
	}
	if len(backend.statusReqs) != 2 || len(backend.profilesReqs) != 2 {
		t.Fatalf("expected sibling watch loop to reuse seeded source without extra polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}

func TestExecuteSharedSessionBrowserClearTargetEventRefreshesPrimaryProviderFromStateSnapshot(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	runRegistry := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-clear-target-event-primary-provider": {{RunID: "run-a", Status: "running"}},
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
	sessionID := "sess-clear-target-event-primary-provider"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}

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
	if tracked.ID == "" {
		t.Fatalf("expected tracked current target id, got %#v", tracked)
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

	initial := bound.ObserveWatchLoop(context.Background(), req)
	if len(initial.View.Session.Routes) != 1 || initial.View.Session.Routes[0].CurrentTargetID != tracked.ID {
		t.Fatalf("expected initial watch loop to expose tracked target, got %#v", initial.View.Session.Routes)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected initial watch loop to poll backend once, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}

	result := ExecuteSharedSessionBrowserClearTargetEvent(
		sessionRegistry,
		runRegistry,
		nil,
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
		time.Minute,
	)
	if !result.Ready || !result.ClearedTargetSelection {
		t.Fatalf("expected top-level clear target event to clear current target, got %#v", result)
	}

	updated := bound.ObserveWatchLoop(context.Background(), req)
	if len(updated.View.Session.Routes) != 1 || updated.View.Session.Routes[0].CurrentTargetID != "" {
		t.Fatalf("expected primary watch loop to refresh from mutation source after clear_target, got %#v", updated.View.Session.Routes)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected primary watch loop to reuse mutation source without extra polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}

func TestExecuteSharedSessionBrowserClearTargetEventRefreshesPrimaryStatusWatchFromStateSnapshot(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	runRegistry := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-clear-target-event-primary-status-watch": {{RunID: "run-a", Status: "running"}},
		}},
	}
	manager := SharedSessionBrowserObserverManagerFor(sessionRegistry, runRegistry, nil, time.Minute)
	backend := &runtimeInfoStatusProfilesObservationTestBackend{
		statusProfilesObservationTestBackend: &statusProfilesObservationTestBackend{
			statusResp: BrowserProfileStatusResult{
				Profile:   "workbench",
				Status:    "running",
				Running:   true,
				Connected: true,
			},
			profilesResp: BrowserProfilesResult{
				Backend:        "proxy",
				DefaultProfile: "isolated",
				Profiles: []BrowserProfileInfo{
					{Profile: "workbench", BrowserApp: "Chromium", Status: "connected", Running: true, Connected: true},
				},
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
	}
	bound := manager.Bind(backend)
	sessionID := "sess-clear-target-event-primary-status-watch"

	tracked := sessionRegistry.TrackTab(sessionID, BrowserSessionTarget{
		ID:         "tab-1",
		TabIndex:   1,
		URL:        "https://example.com/one",
		Title:      "One",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, true)
	if _, ok := sessionRegistry.SelectTargetForRoute(sessionID, BrowserSessionRoute{Backend: "proxy", Profile: "workbench", Target: "node"}, tracked.ID, "select_target"); !ok {
		t.Fatalf("expected explicit target selection, got %#v", tracked)
	}

	req := SharedSessionBrowserObserverRequest{
		SessionID:        sessionID,
		SelectedInfo:     BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
		BindingRoute:     BrowserSessionRoute{Backend: "proxy", Profile: "workbench", Target: "node"},
		RequestedProfile: "workbench",
		IncludeStatus:    true,
		IncludeProfiles:  true,
	}

	initial := bound.ObserveWatch(context.Background(), req)
	if initial.View.Binding.Snapshot.SelectedTargetSelection == nil || initial.View.Binding.Snapshot.SelectedTargetSelection.ID != tracked.ID {
		t.Fatalf("expected initial status watch to expose selected target, got %#v", initial.View.Binding.Snapshot.SelectedTargetSelection)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected initial status watch to poll backend once, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}

	result := ExecuteSharedSessionBrowserClearTargetEvent(
		sessionRegistry,
		runRegistry,
		nil,
		BuildSharedSessionBrowserClearRequest(
			sessionRegistry,
			nil,
			sessionID,
			BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
			BrowserSessionRoute{Backend: "proxy", Profile: "workbench", Target: "node"},
			false,
			"",
			SharedSessionBrowserHealthInput{},
			time.Minute,
		),
		time.Minute,
	)
	if !result.Ready || !result.ClearedTargetSelection {
		t.Fatalf("expected top-level clear target event to clear current target, got %#v", result)
	}

	updated := bound.ObserveWatch(context.Background(), req)
	if updated.View.Binding.Snapshot.SelectedTargetSelection != nil {
		t.Fatalf("expected status watch binding to clear selected target after clear_target, got %#v", updated.View.Binding.Snapshot.SelectedTargetSelection)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected status watch to reuse mutation source without extra polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}
