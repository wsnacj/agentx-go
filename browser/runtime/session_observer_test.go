package browserruntime

import (
	"context"
	"testing"
	"time"
)

func TestObserveSharedSessionBrowserObserverForScopeProjectsBindingViewAndProfiles(t *testing.T) {
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
			Note:           "profiles ok",
			Profiles: []BrowserProfileInfo{
				{Profile: "isolated", BrowserApp: "Chromium", Status: "connected", Running: true, Connected: true},
			},
		},
	}
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	sessionID := "sess-observer"
	runRegistry := testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
		sessionID: {{RunID: "run-1", Status: "running"}},
	}}

	sessionRegistry.TrackTab(sessionID, BrowserSessionTarget{
		TabIndex:   1,
		URL:        "https://example.com",
		Title:      "Example",
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
		BrowserApp: "Chromium",
	}, true)
	stateRegistry.SelectBrowserProfile(sessionID, SharedSessionBrowserProfileSelection{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Source:        "remember_profile",
	})

	observation := ObserveSharedSessionBrowserObserverForScope(
		context.Background(),
		backend,
		SharedSessionBrowserObserverRequest{
			SessionID:                   sessionID,
			SelectedInfo:                BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
			BindingRoute:                BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node"},
			RequestedProfile:            "isolated",
			IncludeStatus:               true,
			IncludeProfiles:             true,
			IncludeSessionView:          true,
			SessionViewInfo:             BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
			SessionViewRouteFilter:      BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node"},
			SessionViewRequestedProfile: "isolated",
		},
		sessionRegistry,
		runRegistry,
		stateRegistry,
		time.Minute,
	)

	if observation.Observation.Status == nil || observation.Observation.Profiles == nil {
		t.Fatalf("expected observer to include status and profiles observation, got %#v", observation.Observation)
	}
	if len(observation.Profiles) != 1 || observation.Profiles[0].State.Profile != "isolated" || !observation.Profiles[0].Selected {
		t.Fatalf("expected projected scoped profiles, got %#v", observation.Profiles)
	}
	if len(observation.Session.Routes) != 1 || len(observation.Session.Runs) != 1 || len(observation.Session.Profiles) != 1 {
		t.Fatalf("expected observer to include session view snapshot, got %#v", observation.Session)
	}
	if observation.Binding.Health.Summary == nil || observation.Binding.Health.Summary.State != "healthy" {
		t.Fatalf("expected healthy binding evaluation, got %#v", observation.Binding)
	}
	if observation.ReferenceTime.IsZero() || !observation.ReferenceTime.Equal(observation.Binding.ReferenceTime) {
		t.Fatalf("expected observer reference time to follow binding evaluation, got %v want %v", observation.ReferenceTime, observation.Binding.ReferenceTime)
	}
}

func TestObserveSharedSessionBrowserObserverForScopeFallsBackWithoutRegistry(t *testing.T) {
	backend := &statusProfilesObservationTestBackend{
		profilesResp: BrowserProfilesResult{
			Backend:        "proxy",
			DefaultProfile: "isolated",
			Profiles: []BrowserProfileInfo{
				{Profile: "isolated", BrowserApp: "Chromium", Status: "running", Running: true},
			},
		},
	}

	observation := ObserveSharedSessionBrowserObserverForScope(
		context.Background(),
		backend,
		SharedSessionBrowserObserverRequest{
			SelectedInfo:     BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
			BindingRoute:     BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node"},
			RequestedProfile: "isolated",
			IncludeProfiles:  true,
		},
		nil,
		nil,
		nil,
		time.Minute,
	)

	if observation.Observation.Profiles == nil {
		t.Fatalf("expected profiles observation")
	}
	if len(observation.Profiles) != 1 || observation.Profiles[0].Selected {
		t.Fatalf("expected observed profiles fallback without registry selection, got %#v", observation.Profiles)
	}
	if observation.DefaultProfile != "isolated" {
		t.Fatalf("expected default profile fallback, got %#v", observation)
	}
}

func TestObserveSharedSessionBrowserObserverForScopeUsesEventCycleReferenceTimeWithoutProfiles(t *testing.T) {
	backend := &statusProfilesObservationTestBackend{
		statusResp: BrowserProfileStatusResult{
			Profile:   "isolated",
			Status:    "running",
			Running:   true,
			Connected: true,
		},
	}

	observation := ObserveSharedSessionBrowserObserverForScope(
		context.Background(),
		backend,
		SharedSessionBrowserObserverRequest{
			SelectedInfo:     BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
			BindingRoute:     BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node"},
			RequestedProfile: "isolated",
			IncludeStatus:    true,
		},
		nil,
		nil,
		nil,
		time.Minute,
	)

	if observation.Observation.Status == nil {
		t.Fatalf("expected status observation")
	}
	if observation.ReferenceTime.IsZero() || !observation.ReferenceTime.Equal(observation.Observation.StatusObservedAt) {
		t.Fatalf("expected observer to fall back to status observed_at as reference time, got %v want %v", observation.ReferenceTime, observation.Observation.StatusObservedAt)
	}
}

func TestSharedSessionBrowserWatchManagerObserveWatchLoopSeedsSiblingProviderAfterLifecycleInvalidation(t *testing.T) {
	backend := &statusProfilesObservationTestBackend{
		statusResp: BrowserProfileStatusResult{
			Backend:    "proxy",
			Profile:    "isolated",
			BrowserApp: "Chromium",
			Status:     "running",
			Running:    true,
			Connected:  true,
		},
	}
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	sessionID := "sess-observer-sibling-provider-watchloop-cache"
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
	watchReq := SharedSessionBrowserObserverRequest{
		SessionID:                   sessionID,
		SelectedInfo:                BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		BindingRoute:                route,
		RequestedProfile:            "isolated",
		IncludeStatus:               true,
		IncludeProfiles:             false,
		IncludeSessionView:          true,
		SessionViewInfo:             BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		SessionViewRouteFilter:      route,
		SessionViewRequestedProfile: "isolated",
	}

	initialA := boundA.ObserveWatchLoop(context.Background(), watchReq)
	initialB := boundB.ObserveWatchLoop(context.Background(), watchReq)
	if len(initialA.View.Session.Routes) != 1 || initialA.View.Session.Routes[0].CurrentTargetID == "" {
		t.Fatalf("expected first provider watch loop to retain current target, got %#v", initialA.View.Session.Routes)
	}
	if len(initialB.View.Session.Routes) != 1 || initialB.View.Session.Routes[0].CurrentTargetID == "" {
		t.Fatalf("expected second provider watch loop to retain current target, got %#v", initialB.View.Session.Routes)
	}
	if len(backend.statusReqs) != 2 || len(backend.profilesReqs) != 0 {
		t.Fatalf("expected sibling watch loops to poll RuntimeStatus once each, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
	if runRegistryA.callCount() != 1 || runRegistryB.callCount() != 1 {
		t.Fatalf("expected initial sibling watch loops to snapshot runs once each, got runA=%d runB=%d", runRegistryA.callCount(), runRegistryB.callCount())
	}

	backend.statusResp = BrowserProfileStatusResult{
		Backend:    "proxy",
		Profile:    "isolated",
		BrowserApp: "Chromium",
		Status:     "disconnected",
		Running:    true,
		Connected:  false,
	}
	managerA.Invalidate()
	seededA := boundA.ObserveWatchLoop(context.Background(), watchReq)
	if len(seededA.View.Session.Routes) != 1 || seededA.View.Session.Routes[0].CurrentTargetID != "" {
		t.Fatalf("expected first watch loop to clear current target after lifecycle invalidation, got %#v", seededA.View.Session.Routes)
	}
	if !seededA.Observer.Observation.HasSyncedState || seededA.Observer.Observation.SyncedState.Status != "disconnected" {
		t.Fatalf("expected first watch loop to sync disconnected lifecycle state, got %#v", seededA.Observer.Observation)
	}
	if runRegistryA.callCount() != 2 || runRegistryB.callCount() != 2 {
		t.Fatalf("expected watch-loop invalidation to refresh both sibling provider projections once, got runA=%d runB=%d", runRegistryA.callCount(), runRegistryB.callCount())
	}

	seededB := boundB.ObserveWatchLoop(context.Background(), watchReq)
	if len(seededB.View.Session.Routes) != 1 || seededB.View.Session.Routes[0].CurrentTargetID != "" {
		t.Fatalf("expected sibling watch loop to reuse invalidated projection cache, got %#v", seededB.View.Session.Routes)
	}
	if len(backend.statusReqs) != 3 || len(backend.profilesReqs) != 0 {
		t.Fatalf("expected sibling watch loop to reuse watch-loop-seeded source without extra polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
	if runRegistryA.callCount() != 2 || runRegistryB.callCount() != 2 {
		t.Fatalf("expected sibling watch loop to reuse watch-loop-seeded projection cache without extra rebuilds, got runA=%d runB=%d", runRegistryA.callCount(), runRegistryB.callCount())
	}
}
