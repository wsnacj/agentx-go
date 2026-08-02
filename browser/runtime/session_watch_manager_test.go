package browserruntime

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type countingSharedSessionRunRegistry struct {
	testSharedSessionRunRegistry
	mu      sync.Mutex
	calls   int
	started chan struct{}
	block   <-chan struct{}
}

func (r *countingSharedSessionRunRegistry) SnapshotSessionRuns(sessionID string) []SharedSessionRunInfo {
	r.mu.Lock()
	r.calls++
	r.mu.Unlock()
	if r.started != nil {
		select {
		case r.started <- struct{}{}:
		default:
		}
	}
	if r.block != nil {
		<-r.block
	}
	return r.testSharedSessionRunRegistry.SnapshotSessionRuns(sessionID)
}

func (r *countingSharedSessionRunRegistry) callCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.calls
}

type countingBrowserSessionStateRegistry struct {
	*BrowserSessionStateRegistry
	mu                               sync.Mutex
	statusAndProfilesResolutionCalls int
	started                          chan struct{}
	block                            <-chan struct{}
}

func (r *countingBrowserSessionStateRegistry) SyncSessionBrowserStatusAndProfilesResolution(
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	requestedProfile string,
	status *BrowserProfileStatusResult,
	statusObservedAt time.Time,
	profiles *BrowserProfilesResult,
	profilesObservedAt time.Time,
	reconnectWindow time.Duration,
) (BrowserProfileStatusResult, SharedSessionBrowserProfileState, bool, []SharedSessionBrowserProfileState) {
	r.mu.Lock()
	r.statusAndProfilesResolutionCalls++
	r.mu.Unlock()
	if r.started != nil {
		select {
		case r.started <- struct{}{}:
		default:
		}
	}
	if r.block != nil {
		<-r.block
	}
	return r.BrowserSessionStateRegistry.SyncSessionBrowserStatusAndProfilesResolution(
		sessionID,
		selectedInfo,
		requestedProfile,
		status,
		statusObservedAt,
		profiles,
		profilesObservedAt,
		reconnectWindow,
	)
}

func (r *countingBrowserSessionStateRegistry) resolutionCallCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.statusAndProfilesResolutionCalls
}

func TestSharedSessionBrowserObserverManagerObserveEventCycleUsesRegistryResolutionAndReferenceTime(t *testing.T) {
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
	manager := NewSharedSessionBrowserObserverManager(nil, nil, NewBrowserSessionStateRegistry(), time.Minute)
	cycle := manager.ObserveEventCycle(
		context.Background(),
		backend,
		SharedSessionBrowserObserverRequest{
			SessionID:        "sess-manager-cycle",
			SelectedInfo:     BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
			BindingRoute:     BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node"},
			RequestedProfile: "isolated",
			IncludeStatus:    true,
			IncludeProfiles:  true,
		},
	)

	if cycle.Observation.Status == nil || cycle.Observation.Profiles == nil {
		t.Fatalf("expected manager event cycle to include raw status and profiles, got %#v", cycle.Observation)
	}
	if !cycle.Observation.HasSyncedState || len(cycle.Observation.Snapshot) != 1 {
		t.Fatalf("expected manager event cycle to sync registry state, got %#v", cycle.Observation)
	}
	expectedReference := cycle.Observation.StatusObservedAt
	if cycle.Observation.ProfilesObservedAt.After(expectedReference) {
		expectedReference = cycle.Observation.ProfilesObservedAt
	}
	if !cycle.ReferenceTime.Equal(expectedReference) {
		t.Fatalf("expected manager event cycle reference time %v, got %v", expectedReference, cycle.ReferenceTime)
	}
}

func TestSharedSessionBrowserObserverManagerObserveWatchLoopProjectsCycleObserverAndWatch(t *testing.T) {
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
	sessionID := "sess-watch-manager"
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

	manager := NewSharedSessionBrowserObserverManager(sessionRegistry, runRegistry, stateRegistry, time.Minute)
	loop := manager.ObserveWatchLoop(
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
	)

	if loop.Cycle.Observation.Status == nil || loop.Cycle.Observation.Profiles == nil {
		t.Fatalf("expected manager watch loop to include cycle status/profiles, got %#v", loop.Cycle)
	}
	if loop.Observer.Binding.Health.Summary == nil || loop.Observer.Binding.Health.Summary.State != "healthy" {
		t.Fatalf("expected manager watch loop to project healthy binding, got %#v", loop.Observer.Binding)
	}
	if len(loop.Watch.Profiles) != 1 || !loop.Watch.Profiles[0].Selected {
		t.Fatalf("expected manager watch loop to project selected profile, got %#v", loop.Watch.Profiles)
	}
	if len(loop.View.Session.Routes) != 1 || len(loop.View.Session.Runs) != 1 || len(loop.View.Session.Profiles) != 1 {
		t.Fatalf("expected manager watch loop to project session view, got %#v", loop.View.Session)
	}
	if loop.ReferenceTime.IsZero() || !loop.ReferenceTime.Equal(loop.Observer.ReferenceTime) || !loop.ReferenceTime.Equal(loop.Watch.ReferenceTime) {
		t.Fatalf("expected manager watch loop reference time to align, got loop=%v observer=%v watch=%v", loop.ReferenceTime, loop.Observer.ReferenceTime, loop.Watch.ReferenceTime)
	}
}

func TestSharedSessionBrowserWatchManagerObserveWatchLoopProjectsCycleObserverAndWatch(t *testing.T) {
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
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	sessionID := "sess-watch-manager-bound"
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

	manager := NewSharedSessionBrowserWatchManager(backend, sessionRegistry, runRegistry, stateRegistry, time.Minute)
	loop := manager.ObserveWatchLoop(
		context.Background(),
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
	)

	if loop.Cycle.Observation.Status == nil || loop.Cycle.Observation.Profiles == nil {
		t.Fatalf("expected bound watch manager loop to include cycle status/profiles, got %#v", loop.Cycle)
	}
	if loop.Observer.Binding.Health.Summary == nil || loop.Observer.Binding.Health.Summary.State != "healthy" {
		t.Fatalf("expected bound watch manager to project healthy binding, got %#v", loop.Observer.Binding)
	}
	if len(loop.View.Session.Routes) != 1 || len(loop.Watch.Profiles) != 1 || !loop.Watch.Profiles[0].Selected {
		t.Fatalf("expected bound watch manager to project session view and selected profile, got view=%#v profiles=%#v", loop.View.Session, loop.Watch.Profiles)
	}
}

func TestSharedSessionBrowserObserverManagerBindCachesBoundManagersByControl(t *testing.T) {
	backendA := &statusProfilesObservationTestBackend{}
	backendB := &statusProfilesObservationTestBackend{}
	manager := NewSharedSessionBrowserObserverManager(nil, nil, nil, time.Minute)

	first := manager.Bind(backendA)
	second := manager.Bind(backendA)
	third := manager.Bind(backendB)

	if first.Control != second.Control {
		t.Fatalf("expected repeated bind for same control to preserve bound control, got first=%#v second=%#v", first.Control, second.Control)
	}
	if manager.cache == nil {
		t.Fatalf("expected observer manager to initialize bound-manager cache")
	}
	if got := len(manager.cache.managers); got != 2 {
		t.Fatalf("expected two cached bound managers, got %d", got)
	}
	keyA, ok := sharedSessionBrowserWatchManagerCacheKeyForControl(backendA)
	if !ok {
		t.Fatalf("expected cache key for backendA")
	}
	cachedA, ok := manager.cache.managers[keyA]
	if !ok || cachedA.Control != backendA {
		t.Fatalf("expected cached bound manager for backendA, got %#v", cachedA)
	}
	keyB, ok := sharedSessionBrowserWatchManagerCacheKeyForControl(backendB)
	if !ok {
		t.Fatalf("expected cache key for backendB")
	}
	cachedB, ok := manager.cache.managers[keyB]
	if !ok || cachedB.Control != backendB {
		t.Fatalf("expected cached bound manager for backendB, got %#v", cachedB)
	}
	_ = third
}

func TestSharedSessionBrowserObserverManagerBindPrunesIdleBoundManagers(t *testing.T) {
	backendA := &statusProfilesObservationTestBackend{}
	backendB := &statusProfilesObservationTestBackend{}
	backendC := &statusProfilesObservationTestBackend{}
	manager := NewSharedSessionBrowserObserverManager(nil, nil, nil, time.Minute)

	first := manager.Bind(backendA)
	second := manager.Bind(backendB)
	first.touch()
	second.touch()

	staleAt := time.Now().Add(-sharedSessionBrowserWatchManagerIdleTTL - time.Minute)
	first.state.mu.Lock()
	first.state.lastActiveAt = staleAt
	first.state.mu.Unlock()

	third := manager.Bind(backendC)

	if manager.cache == nil {
		t.Fatalf("expected observer manager to initialize bound-manager cache")
	}
	if got := len(manager.cache.managers); got != 2 {
		t.Fatalf("expected stale bound manager to be pruned before adding third manager, got %d", got)
	}
	keyA, ok := sharedSessionBrowserWatchManagerCacheKeyForControl(backendA)
	if !ok {
		t.Fatalf("expected cache key for backendA")
	}
	keyB, ok := sharedSessionBrowserWatchManagerCacheKeyForControl(backendB)
	if !ok {
		t.Fatalf("expected cache key for backendB")
	}
	keyC, ok := sharedSessionBrowserWatchManagerCacheKeyForControl(backendC)
	if !ok {
		t.Fatalf("expected cache key for backendC")
	}
	if _, found := manager.cache.managers[keyA]; found {
		t.Fatalf("expected stale backendA manager to be pruned")
	}
	if _, found := manager.cache.managers[keyB]; !found {
		t.Fatalf("expected active backendB manager to be retained")
	}
	if _, found := manager.cache.managers[keyC]; !found {
		t.Fatalf("expected new backendC manager to be cached")
	}
	if third.Control != backendC {
		t.Fatalf("expected bound backendC manager, got %#v", third.Control)
	}
}

func TestSharedSessionBrowserObserverManagerInvalidateLazilyRefreshesBoundWatchLoopByGeneration(t *testing.T) {
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
	sessionRegistry := NewBrowserSessionRegistry()
	manager := NewSharedSessionBrowserObserverManager(sessionRegistry, nil, nil, time.Minute)
	bound := manager.Bind(backend)
	sessionID := "sess-manager-invalidate-generation"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node"}

	sessionRegistry.TrackTab(sessionID, BrowserSessionTarget{
		TabIndex:   1,
		URL:        "https://example.com/first",
		Title:      "First",
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
		BrowserApp: "Chromium",
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
	if len(first.View.Session.Routes) != 1 || first.View.Session.Routes[0].Targets[0].Title != "First" {
		t.Fatalf("expected initial watch loop to cache first route title, got %#v", first.View.Session.Routes)
	}
	if bound.state == nil || len(bound.state.watchLoops) == 0 {
		t.Fatalf("expected initial watch loop to populate bound manager cache")
	}
	initialGeneration := bound.state.generation

	sessionRegistry.TrackTab(sessionID, BrowserSessionTarget{
		TabIndex:   1,
		URL:        "https://example.com/second",
		Title:      "Second",
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
		BrowserApp: "Chromium",
	}, true)
	manager.Invalidate()

	if manager.currentGeneration() != initialGeneration+1 {
		t.Fatalf("expected provider invalidation to bump generation from %d to %d, got %d", initialGeneration, initialGeneration+1, manager.currentGeneration())
	}
	if bound.state.generation != initialGeneration {
		t.Fatalf("expected bound manager generation to remain stale until next observation, got %d want %d", bound.state.generation, initialGeneration)
	}
	if len(bound.state.watchLoops) == 0 {
		t.Fatalf("expected lazy invalidation to preserve cached watch loops until next observation")
	}

	second := bound.ObserveWatchLoop(context.Background(), req)
	if bound.state.generation != manager.currentGeneration() {
		t.Fatalf("expected next observation to sync bound generation to provider generation, got bound=%d provider=%d", bound.state.generation, manager.currentGeneration())
	}
	if len(second.View.Session.Routes) != 1 || second.View.Session.Routes[0].Targets[0].Title != "Second" {
		t.Fatalf("expected lazy invalidation to refresh watch loop on next observation, got %#v", second.View.Session.Routes)
	}
}

func TestSharedSessionBrowserObserverManagerForCachesProvidersByDependencies(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()

	first := SharedSessionBrowserObserverManagerFor(sessionRegistry, nil, stateRegistry, time.Minute)
	second := SharedSessionBrowserObserverManagerFor(sessionRegistry, nil, stateRegistry, time.Minute)
	third := SharedSessionBrowserObserverManagerFor(nil, nil, stateRegistry, time.Minute)

	if first.cache == nil || second.cache == nil {
		t.Fatalf("expected shared provider helper to initialize provider cache")
	}
	if first.cache != second.cache {
		t.Fatalf("expected identical dependencies to reuse cached provider, got first=%p second=%p", first.cache, second.cache)
	}
	if third.cache == nil {
		t.Fatalf("expected alternate dependencies to initialize provider cache")
	}
	if third.cache == first.cache {
		t.Fatalf("expected different dependency set to produce distinct provider cache")
	}
}

func TestSharedSessionBrowserObserverManagerForPrunesIdleProviders(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()

	first := SharedSessionBrowserObserverManagerFor(sessionRegistry, nil, stateRegistry, time.Minute)
	key, ok := sharedSessionBrowserObserverManagerProviderCacheKeyForDependencies(sessionRegistry, nil, stateRegistry, time.Minute)
	if !ok {
		t.Fatalf("expected provider cache key")
	}
	sharedSessionBrowserObserverManagerProviders.mu.RLock()
	entry := sharedSessionBrowserObserverManagerProviders.providers[key]
	sharedSessionBrowserObserverManagerProviders.mu.RUnlock()
	if entry == nil {
		t.Fatalf("expected cached provider entry")
	}
	entry.mu.Lock()
	entry.lastActiveAt = time.Now().Add(-sharedSessionBrowserObserverManagerIdleTTL - time.Minute)
	entry.mu.Unlock()

	second := SharedSessionBrowserObserverManagerFor(sessionRegistry, nil, stateRegistry, time.Minute)
	sharedSessionBrowserObserverManagerProviders.mu.RLock()
	replacement := sharedSessionBrowserObserverManagerProviders.providers[key]
	sharedSessionBrowserObserverManagerProviders.mu.RUnlock()

	if replacement == nil {
		t.Fatalf("expected replacement provider entry after pruning")
	}
	if replacement == entry {
		t.Fatalf("expected stale provider entry to be pruned and replaced")
	}
	if first.cache == second.cache {
		t.Fatalf("expected stale provider cache to be replaced after pruning")
	}
}

func TestSharedSessionBrowserObserverManagerBindRefreshesProviderActivity(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()

	manager := SharedSessionBrowserObserverManagerFor(sessionRegistry, nil, stateRegistry, time.Minute)
	key, ok := sharedSessionBrowserObserverManagerProviderCacheKeyForDependencies(sessionRegistry, nil, stateRegistry, time.Minute)
	if !ok {
		t.Fatalf("expected provider cache key")
	}
	sharedSessionBrowserObserverManagerProviders.mu.RLock()
	entry := sharedSessionBrowserObserverManagerProviders.providers[key]
	sharedSessionBrowserObserverManagerProviders.mu.RUnlock()
	if entry == nil {
		t.Fatalf("expected cached provider entry")
	}

	staleAt := time.Now().Add(-sharedSessionBrowserObserverManagerIdleTTL - time.Minute)
	entry.mu.Lock()
	entry.lastActiveAt = staleAt
	entry.mu.Unlock()

	manager.Bind(&statusProfilesObservationTestBackend{})

	entry.mu.RLock()
	refreshedAt := entry.lastActiveAt
	entry.mu.RUnlock()
	if !refreshedAt.After(staleAt) {
		t.Fatalf("expected bind to refresh provider activity, got stale=%v refreshed=%v", staleAt, refreshedAt)
	}

	second := SharedSessionBrowserObserverManagerFor(sessionRegistry, nil, stateRegistry, time.Minute)
	if manager.cache != second.cache {
		t.Fatalf("expected refreshed provider activity to preserve cached provider")
	}
}

func TestSharedSessionBrowserObserverManagerObserveEventCycleRefreshesProviderActivity(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
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
	manager := SharedSessionBrowserObserverManagerFor(sessionRegistry, nil, stateRegistry, time.Minute)
	key, ok := sharedSessionBrowserObserverManagerProviderCacheKeyForDependencies(sessionRegistry, nil, stateRegistry, time.Minute)
	if !ok {
		t.Fatalf("expected provider cache key")
	}
	sharedSessionBrowserObserverManagerProviders.mu.RLock()
	entry := sharedSessionBrowserObserverManagerProviders.providers[key]
	sharedSessionBrowserObserverManagerProviders.mu.RUnlock()
	if entry == nil {
		t.Fatalf("expected cached provider entry")
	}

	staleAt := time.Now().Add(-sharedSessionBrowserObserverManagerIdleTTL - time.Minute)
	entry.mu.Lock()
	entry.lastActiveAt = staleAt
	entry.mu.Unlock()

	manager.ObserveEventCycle(context.Background(), backend, SharedSessionBrowserObserverRequest{
		SessionID:        "sess-provider-touch-cycle",
		SelectedInfo:     BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		BindingRoute:     BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node"},
		RequestedProfile: "isolated",
		IncludeStatus:    true,
		IncludeProfiles:  true,
	})

	entry.mu.RLock()
	refreshedAt := entry.lastActiveAt
	entry.mu.RUnlock()
	if !refreshedAt.After(staleAt) {
		t.Fatalf("expected direct provider observation to refresh provider activity, got stale=%v refreshed=%v", staleAt, refreshedAt)
	}

	second := SharedSessionBrowserObserverManagerFor(sessionRegistry, nil, stateRegistry, time.Minute)
	if manager.cache != second.cache {
		t.Fatalf("expected refreshed provider activity to preserve cached provider")
	}
}

func TestSharedSessionBrowserObserverManagerApplyExecutionResultRefreshesProviderActivity(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	manager := SharedSessionBrowserObserverManagerFor(sessionRegistry, nil, stateRegistry, time.Minute)
	key, ok := sharedSessionBrowserObserverManagerProviderCacheKeyForDependencies(sessionRegistry, nil, stateRegistry, time.Minute)
	if !ok {
		t.Fatalf("expected provider cache key")
	}
	sharedSessionBrowserObserverManagerProviders.mu.RLock()
	entry := sharedSessionBrowserObserverManagerProviders.providers[key]
	sharedSessionBrowserObserverManagerProviders.mu.RUnlock()
	if entry == nil {
		t.Fatalf("expected cached provider entry")
	}

	staleAt := time.Now().Add(-sharedSessionBrowserObserverManagerIdleTTL - time.Minute)
	entry.mu.Lock()
	entry.lastActiveAt = staleAt
	entry.mu.Unlock()

	manager.ApplyExecutionResult("", BrowserRuntimeInfo{}, "", SharedSessionBrowserExecutionResult{})

	entry.mu.RLock()
	refreshedAt := entry.lastActiveAt
	entry.mu.RUnlock()
	if !refreshedAt.After(staleAt) {
		t.Fatalf("expected execution apply to refresh provider activity, got stale=%v refreshed=%v", staleAt, refreshedAt)
	}

	second := SharedSessionBrowserObserverManagerFor(sessionRegistry, nil, stateRegistry, time.Minute)
	if manager.cache != second.cache {
		t.Fatalf("expected refreshed provider activity to preserve cached provider")
	}
}

func TestSharedSessionBrowserWatchManagerObserveEventCycleCachesShortLivedSourceTimeCycle(t *testing.T) {
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
	manager := NewSharedSessionBrowserWatchManager(backend, nil, nil, nil, time.Minute)
	req := SharedSessionBrowserObserverRequest{
		SessionID:        "sess-cache",
		SelectedInfo:     BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		BindingRoute:     BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node"},
		RequestedProfile: "isolated",
		IncludeStatus:    true,
		IncludeProfiles:  true,
	}

	first := manager.ObserveEventCycle(context.Background(), req)
	second := manager.ObserveEventCycle(context.Background(), req)

	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected cached event cycle to reuse raw status/profiles polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
	if !second.ReferenceTime.Equal(first.ReferenceTime) {
		t.Fatalf("expected cached event cycle to preserve reference time, got first=%v second=%v", first.ReferenceTime, second.ReferenceTime)
	}
}

func TestSharedSessionBrowserObserverManagerObserveRawStatusReusesBoundCacheByControl(t *testing.T) {
	backend := &statusProfilesObservationTestBackend{
		statusResp: BrowserProfileStatusResult{
			Backend: "proxy",
			Profile: "isolated",
			Status:  "running",
		},
	}
	manager := NewSharedSessionBrowserObserverManager(nil, nil, nil, time.Minute)

	first := manager.ObserveRawStatus(context.Background(), backend, "isolated")
	second := manager.ObserveRawStatus(context.Background(), backend, "isolated")

	if first.Status == nil || second.Status == nil {
		t.Fatalf("expected successful raw status observations, got first=%#v second=%#v", first, second)
	}
	if len(backend.statusReqs) != 1 {
		t.Fatalf("expected provider observe raw status to reuse bound cache, got %#v", backend.statusReqs)
	}
	if !second.ObservedAt.Equal(first.ObservedAt) {
		t.Fatalf("expected cached provider raw status to preserve observed_at, got first=%v second=%v", first.ObservedAt, second.ObservedAt)
	}
}

func TestSharedSessionBrowserObserverManagerObserveEventCycleReusesBoundCacheByControl(t *testing.T) {
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
	manager := NewSharedSessionBrowserObserverManager(nil, nil, nil, time.Minute)
	req := SharedSessionBrowserObserverRequest{
		SessionID:        "sess-provider-cycle-cache",
		SelectedInfo:     BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		BindingRoute:     BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node"},
		RequestedProfile: "isolated",
		IncludeStatus:    true,
		IncludeProfiles:  true,
	}

	first := manager.ObserveEventCycle(context.Background(), backend, req)
	second := manager.ObserveEventCycle(context.Background(), backend, req)

	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected provider observe event cycle to reuse bound cache, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
	if !second.ReferenceTime.Equal(first.ReferenceTime) {
		t.Fatalf("expected cached provider event cycle to preserve reference time, got first=%v second=%v", first.ReferenceTime, second.ReferenceTime)
	}
}

func TestSharedSessionBrowserObserverManagerMixedObservePathsReuseBoundCacheByControl(t *testing.T) {
	backend := &statusProfilesObservationTestBackend{
		statusResp: BrowserProfileStatusResult{
			Backend:   "proxy",
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
	manager := NewSharedSessionBrowserObserverManager(nil, nil, nil, time.Minute)
	selectedInfo := BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"}

	first := manager.ObserveStatusAndProfiles(
		context.Background(),
		backend,
		"sess-provider-mixed-cache",
		selectedInfo,
		"isolated",
		true,
		true,
	)
	second := manager.ObserveProfiles(
		context.Background(),
		backend,
		"sess-provider-mixed-cache",
		selectedInfo,
		"isolated",
	)

	if first.Profiles == nil || second.Profiles == nil {
		t.Fatalf("expected successful mixed observe payloads, got first=%#v second=%#v", first, second)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected mixed provider observe paths to reuse bound cache, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
	if !second.ObservedAt.Equal(first.ProfilesObservedAt) {
		t.Fatalf("expected mixed provider observe paths to preserve profiles observed_at, got first=%v second=%v", first.ProfilesObservedAt, second.ObservedAt)
	}
}

func TestSharedSessionBrowserWatchManagerObserveEventCycleProjectsStatusOnlyFromCachedSuperset(t *testing.T) {
	backend := &statusProfilesObservationTestBackend{
		statusResp: BrowserProfileStatusResult{
			Backend:   "proxy",
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
	stateRegistry := &countingBrowserSessionStateRegistry{
		BrowserSessionStateRegistry: NewBrowserSessionStateRegistry(),
	}
	manager := NewSharedSessionBrowserWatchManager(backend, nil, nil, stateRegistry, time.Minute)
	fullReq := SharedSessionBrowserObserverRequest{
		SessionID:        "sess-cycle-superset-cache",
		SelectedInfo:     BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		BindingRoute:     BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node"},
		RequestedProfile: "isolated",
		IncludeStatus:    true,
		IncludeProfiles:  true,
	}

	full := manager.ObserveEventCycle(context.Background(), fullReq)
	statusOnly := manager.ObserveEventCycle(context.Background(), SharedSessionBrowserObserverRequest{
		SessionID:        fullReq.SessionID,
		SelectedInfo:     fullReq.SelectedInfo,
		BindingRoute:     fullReq.BindingRoute,
		RequestedProfile: fullReq.RequestedProfile,
		IncludeStatus:    true,
	})

	if full.Observation.Status == nil || full.Observation.Profiles == nil {
		t.Fatalf("expected cached superset source cycle, got %#v", full)
	}
	if statusOnly.Observation.Status == nil || statusOnly.Observation.Profiles != nil {
		t.Fatalf("expected status-only projection from cached superset, got %#v", statusOnly.Observation)
	}
	if statusOnly.Observation.StatusErr != nil || statusOnly.Observation.ProfilesErr != nil {
		t.Fatalf("expected projected status-only cycle to keep successful fields only, got %#v", statusOnly.Observation)
	}
	if len(statusOnly.Observation.Snapshot) != 0 {
		t.Fatalf("expected status-only projection to clear profile snapshot, got %#v", statusOnly.Observation.Snapshot)
	}
	if !statusOnly.Observation.StatusObservedAt.Equal(full.Observation.StatusObservedAt) {
		t.Fatalf("expected status-only projection to preserve status observed_at, got full=%v projected=%v", full.Observation.StatusObservedAt, statusOnly.Observation.StatusObservedAt)
	}
	if !statusOnly.ReferenceTime.Equal(full.Observation.StatusObservedAt) {
		t.Fatalf("expected status-only projection to recompute reference time from status observed_at, got status=%v reference=%v", full.Observation.StatusObservedAt, statusOnly.ReferenceTime)
	}
	if stateRegistry.resolutionCallCount() != 1 {
		t.Fatalf("expected cached superset event cycle to avoid duplicate resolution, got %d calls", stateRegistry.resolutionCallCount())
	}
	if backend.statusRequestCount() != 1 || backend.profilesRequestCount() != 1 {
		t.Fatalf("expected cached superset event cycle to avoid duplicate backend polling, got status=%d profiles=%d", backend.statusRequestCount(), backend.profilesRequestCount())
	}
}

func TestSharedSessionBrowserWatchManagerObserveProfilesReusesCompatibleEventCycleInFlight(t *testing.T) {
	resolutionBlock := make(chan struct{})
	resolutionStarted := make(chan struct{}, 1)
	backend := &statusProfilesObservationTestBackend{
		statusResp: BrowserProfileStatusResult{
			Backend:   "proxy",
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
	stateRegistry := &countingBrowserSessionStateRegistry{
		BrowserSessionStateRegistry: NewBrowserSessionStateRegistry(),
		started:                     resolutionStarted,
		block:                       resolutionBlock,
	}
	manager := NewSharedSessionBrowserWatchManager(backend, nil, nil, stateRegistry, time.Minute)
	fullReq := SharedSessionBrowserObserverRequest{
		SessionID:        "sess-cycle-superset-in-flight",
		SelectedInfo:     BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		BindingRoute:     BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node"},
		RequestedProfile: "isolated",
		IncludeStatus:    true,
		IncludeProfiles:  true,
	}

	fullCh := make(chan SharedSessionBrowserEventCycleObservation, 1)
	profilesCh := make(chan SharedSessionBrowserProfilesObservation, 1)
	go func() {
		fullCh <- manager.ObserveEventCycle(context.Background(), fullReq)
	}()

	select {
	case <-resolutionStarted:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for superset event-cycle resolution to start")
	}

	go func() {
		profilesCh <- manager.ObserveProfiles(context.Background(), fullReq.SessionID, fullReq.SelectedInfo, fullReq.RequestedProfile)
	}()
	time.Sleep(20 * time.Millisecond)

	if stateRegistry.resolutionCallCount() != 1 {
		t.Fatalf("expected profiles-only observe to reuse broader in-flight resolution, got %d calls", stateRegistry.resolutionCallCount())
	}
	if backend.statusRequestCount() != 1 || backend.profilesRequestCount() != 1 {
		t.Fatalf("expected profiles-only observe to reuse broader in-flight raw polling, got status=%d profiles=%d", backend.statusRequestCount(), backend.profilesRequestCount())
	}

	close(resolutionBlock)
	full := <-fullCh
	profiles := <-profilesCh

	if full.Observation.Profiles == nil || profiles.Profiles == nil {
		t.Fatalf("expected successful broader and projected profiles observations, got full=%#v profiles=%#v", full, profiles)
	}
	if !profiles.ObservedAt.Equal(full.Observation.ProfilesObservedAt) {
		t.Fatalf("expected profiles-only projection to preserve broader profiles observed_at, got full=%v projected=%v", full.Observation.ProfilesObservedAt, profiles.ObservedAt)
	}
	if stateRegistry.resolutionCallCount() != 1 {
		t.Fatalf("expected profiles-only observe to finish without extra resolution, got %d calls", stateRegistry.resolutionCallCount())
	}
}

func TestSharedSessionBrowserWatchManagerObserveProfilesProjectsFromEventCycleWithoutRunSnapshot(t *testing.T) {
	backend := &statusProfilesObservationTestBackend{
		profilesResp: BrowserProfilesResult{
			Backend:        "proxy",
			DefaultProfile: "isolated",
			Profiles: []BrowserProfileInfo{
				{Profile: "isolated", BrowserApp: "Chromium", Status: "connected", Running: true, Connected: true},
			},
		},
	}
	stateRegistry := NewBrowserSessionStateRegistry()
	stateRegistry.SelectBrowserProfile("sess-profiles-event-cycle", SharedSessionBrowserProfileSelection{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Source:        "remember_profile",
	})
	runRegistry := &countingSharedSessionRunRegistry{}
	manager := NewSharedSessionBrowserWatchManager(backend, NewBrowserSessionRegistry(), runRegistry, stateRegistry, time.Minute)

	profiles := manager.ObserveProfiles(
		context.Background(),
		"sess-profiles-event-cycle",
		BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		"isolated",
	)

	if profiles.Profiles == nil || len(profiles.Projected) != 1 || !profiles.Projected[0].Selected {
		t.Fatalf("expected profiles-only observe to project selected profiles from event cycle, got %#v", profiles)
	}
	if backend.statusRequestCount() != 0 || backend.profilesRequestCount() != 1 {
		t.Fatalf("expected profiles-only observe to skip status polling and use one profiles poll, got status=%d profiles=%d", backend.statusRequestCount(), backend.profilesRequestCount())
	}
	if runRegistry.callCount() != 0 {
		t.Fatalf("expected profiles-only observe to avoid session-view/run snapshot rebuilding, got %d run snapshots", runRegistry.callCount())
	}
}

func TestSharedSessionBrowserWatchManagerObserveProfilesReusesCachedSupersetEventCycleWithoutRunSnapshot(t *testing.T) {
	backend := &statusProfilesObservationTestBackend{
		statusResp: BrowserProfileStatusResult{
			Backend:   "proxy",
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
	stateRegistry := &countingBrowserSessionStateRegistry{
		BrowserSessionStateRegistry: NewBrowserSessionStateRegistry(),
	}
	stateRegistry.SelectBrowserProfile("sess-profiles-superset-cache", SharedSessionBrowserProfileSelection{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Source:        "remember_profile",
	})
	runRegistry := &countingSharedSessionRunRegistry{}
	manager := NewSharedSessionBrowserWatchManager(backend, NewBrowserSessionRegistry(), runRegistry, stateRegistry, time.Minute)
	selectedInfo := BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"}

	full := manager.ObserveStatusAndProfiles(context.Background(), "sess-profiles-superset-cache", selectedInfo, "isolated", true, true)
	profiles := manager.ObserveProfiles(context.Background(), "sess-profiles-superset-cache", selectedInfo, "isolated")

	if full.Profiles == nil || profiles.Profiles == nil || len(profiles.Projected) != 1 || !profiles.Projected[0].Selected {
		t.Fatalf("expected cached superset event cycle to satisfy profiles-only observe, got full=%#v profiles=%#v", full, profiles)
	}
	if !profiles.ObservedAt.Equal(full.ProfilesObservedAt) {
		t.Fatalf("expected cached superset profiles observe to preserve observed_at, got full=%v projected=%v", full.ProfilesObservedAt, profiles.ObservedAt)
	}
	if stateRegistry.resolutionCallCount() != 1 {
		t.Fatalf("expected cached superset cycle to avoid extra resolution, got %d calls", stateRegistry.resolutionCallCount())
	}
	if backend.statusRequestCount() != 1 || backend.profilesRequestCount() != 1 {
		t.Fatalf("expected cached superset cycle to avoid extra backend polling, got status=%d profiles=%d", backend.statusRequestCount(), backend.profilesRequestCount())
	}
	if runRegistry.callCount() != 0 {
		t.Fatalf("expected profiles-only observe from cached superset cycle to avoid run snapshots, got %d", runRegistry.callCount())
	}
}

func TestSharedSessionBrowserWatchManagerObserveBindingCachesWithoutWatchLoopProjection(t *testing.T) {
	backend := &statusProfilesObservationTestBackend{
		statusResp: BrowserProfileStatusResult{
			Backend:   "proxy",
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
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	runRegistry := &countingSharedSessionRunRegistry{}
	manager := NewSharedSessionBrowserWatchManager(backend, sessionRegistry, runRegistry, stateRegistry, time.Minute)
	req := SharedSessionBrowserObserverRequest{
		SessionID:        "sess-binding-cache",
		SelectedInfo:     BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		BindingRoute:     BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node"},
		RequestedProfile: "isolated",
		IncludeStatus:    true,
		IncludeProfiles:  true,
	}

	first := manager.ObserveBinding(context.Background(), req)
	second := manager.ObserveBinding(context.Background(), req)

	if first.Observation.Status == nil || first.Observation.Profiles == nil || second.Observation.Status == nil || second.Observation.Profiles == nil {
		t.Fatalf("expected cached binding observe to preserve status/profiles, got first=%#v second=%#v", first, second)
	}
	if backend.statusRequestCount() != 1 || backend.profilesRequestCount() != 1 {
		t.Fatalf("expected binding cache to avoid extra backend polling, got status=%d profiles=%d", backend.statusRequestCount(), backend.profilesRequestCount())
	}
	if runRegistry.callCount() != 1 {
		t.Fatalf("expected binding cache to avoid duplicate binding evaluation, got %d run snapshots", runRegistry.callCount())
	}
	if manager.state == nil || len(manager.state.bindings) != 1 {
		t.Fatalf("expected direct binding observe to populate binding cache, got %#v", manager.state)
	}
	if len(manager.state.watchLoops) != 0 {
		t.Fatalf("expected direct binding observe to avoid watch-loop cache, got %d cached watch loops", len(manager.state.watchLoops))
	}
}

func TestSharedSessionBrowserWatchManagerObserveBindingCoalescesConcurrentProjection(t *testing.T) {
	runBlock := make(chan struct{})
	runStarted := make(chan struct{}, 1)
	backend := &statusProfilesObservationTestBackend{
		statusResp: BrowserProfileStatusResult{
			Backend:   "proxy",
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
	runRegistry := &countingSharedSessionRunRegistry{
		started: runStarted,
		block:   runBlock,
	}
	manager := NewSharedSessionBrowserWatchManager(backend, NewBrowserSessionRegistry(), runRegistry, NewBrowserSessionStateRegistry(), time.Minute)
	req := SharedSessionBrowserObserverRequest{
		SessionID:        "sess-binding-in-flight",
		SelectedInfo:     BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		BindingRoute:     BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node"},
		RequestedProfile: "isolated",
		IncludeStatus:    true,
		IncludeProfiles:  true,
	}

	firstCh := make(chan SharedSessionBrowserBindingObservation, 1)
	secondCh := make(chan SharedSessionBrowserBindingObservation, 1)
	go func() {
		firstCh <- manager.ObserveBinding(context.Background(), req)
	}()

	select {
	case <-runStarted:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for binding evaluation to start")
	}

	go func() {
		secondCh <- manager.ObserveBinding(context.Background(), req)
	}()
	time.Sleep(20 * time.Millisecond)

	if runRegistry.callCount() != 1 {
		t.Fatalf("expected concurrent binding observe to reuse one in-flight binding evaluation, got %d run snapshots", runRegistry.callCount())
	}
	if backend.statusRequestCount() != 1 || backend.profilesRequestCount() != 1 {
		t.Fatalf("expected concurrent binding observe to reuse one backend cycle, got status=%d profiles=%d", backend.statusRequestCount(), backend.profilesRequestCount())
	}

	close(runBlock)
	first := <-firstCh
	second := <-secondCh

	if first.Observation.Status == nil || first.Observation.Profiles == nil || second.Observation.Status == nil || second.Observation.Profiles == nil {
		t.Fatalf("expected in-flight binding observe to return projected status/profiles, got first=%#v second=%#v", first, second)
	}
	if runRegistry.callCount() != 1 {
		t.Fatalf("expected concurrent binding observe to finish without extra binding evaluation, got %d run snapshots", runRegistry.callCount())
	}
}

func TestSharedSessionBrowserWatchManagerObserveViewCachesWithoutWatchLoopProjection(t *testing.T) {
	backend := &statusProfilesObservationTestBackend{
		statusResp: BrowserProfileStatusResult{
			Backend:   "proxy",
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
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	runRegistry := &countingSharedSessionRunRegistry{}
	sessionID := "sess-view-cache"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node"}
	sessionRegistry.TrackTab(sessionID, BrowserSessionTarget{
		TabIndex:   1,
		URL:        "https://example.com",
		Title:      "Example",
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
		BrowserApp: "Chromium",
	}, true)
	manager := NewSharedSessionBrowserWatchManager(backend, sessionRegistry, runRegistry, stateRegistry, time.Minute)
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

	first := manager.ObserveView(context.Background(), req)
	second := manager.ObserveView(context.Background(), req)

	if first.Observation.Status == nil || first.Observation.Profiles == nil || len(first.Session.Routes) != 1 || len(second.Session.Routes) != 1 {
		t.Fatalf("expected cached view observe to preserve binding/session payloads, got first=%#v second=%#v", first, second)
	}
	if backend.statusRequestCount() != 1 || backend.profilesRequestCount() != 1 {
		t.Fatalf("expected view cache to avoid extra backend polling, got status=%d profiles=%d", backend.statusRequestCount(), backend.profilesRequestCount())
	}
	if runRegistry.callCount() != 1 {
		t.Fatalf("expected view cache to avoid duplicate session-view projection, got %d run snapshots", runRegistry.callCount())
	}
	if manager.state == nil || len(manager.state.views) != 1 {
		t.Fatalf("expected direct view observe to populate view cache, got %#v", manager.state)
	}
	if len(manager.state.watchLoops) != 0 {
		t.Fatalf("expected direct view observe to avoid watch-loop cache, got %d cached watch loops", len(manager.state.watchLoops))
	}
}

func TestSharedSessionBrowserWatchManagerObserveViewCoalescesConcurrentProjection(t *testing.T) {
	runBlock := make(chan struct{})
	runStarted := make(chan struct{}, 1)
	backend := &statusProfilesObservationTestBackend{
		statusResp: BrowserProfileStatusResult{
			Backend:   "proxy",
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
	sessionRegistry := NewBrowserSessionRegistry()
	runRegistry := &countingSharedSessionRunRegistry{
		started: runStarted,
		block:   runBlock,
	}
	sessionID := "sess-view-in-flight"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node"}
	sessionRegistry.TrackTab(sessionID, BrowserSessionTarget{
		TabIndex:   1,
		URL:        "https://example.com",
		Title:      "Example",
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
		BrowserApp: "Chromium",
	}, true)
	manager := NewSharedSessionBrowserWatchManager(backend, sessionRegistry, runRegistry, NewBrowserSessionStateRegistry(), time.Minute)
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

	firstCh := make(chan SharedSessionBrowserViewObservation, 1)
	secondCh := make(chan SharedSessionBrowserViewObservation, 1)
	go func() {
		firstCh <- manager.ObserveView(context.Background(), req)
	}()

	select {
	case <-runStarted:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for view projection to start")
	}

	go func() {
		secondCh <- manager.ObserveView(context.Background(), req)
	}()
	time.Sleep(20 * time.Millisecond)

	if runRegistry.callCount() != 1 {
		t.Fatalf("expected concurrent view observe to reuse one in-flight session-view projection, got %d run snapshots", runRegistry.callCount())
	}
	if backend.statusRequestCount() != 1 || backend.profilesRequestCount() != 1 {
		t.Fatalf("expected concurrent view observe to reuse one backend cycle, got status=%d profiles=%d", backend.statusRequestCount(), backend.profilesRequestCount())
	}

	close(runBlock)
	first := <-firstCh
	second := <-secondCh

	if len(first.Session.Routes) != 1 || len(second.Session.Routes) != 1 {
		t.Fatalf("expected in-flight view observe to return session snapshot, got first=%#v second=%#v", first.Session, second.Session)
	}
	if runRegistry.callCount() != 1 {
		t.Fatalf("expected concurrent view observe to finish without extra session-view projection, got %d run snapshots", runRegistry.callCount())
	}
}

func TestSharedSessionBrowserWatchManagerObserveWatchReusesCachedWatchLoopWithoutSessionView(t *testing.T) {
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
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	sessionID := "sess-watch-loop-superset-cache"
	runRegistry := &countingSharedSessionRunRegistry{testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
		sessionID: {{RunID: "run-1", Status: "running"}},
	}}}

	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node"}
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

	manager := NewSharedSessionBrowserWatchManager(backend, sessionRegistry, runRegistry, stateRegistry, time.Minute)
	full := manager.ObserveWatchLoop(context.Background(), SharedSessionBrowserObserverRequest{
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
	})
	watch := manager.ObserveWatch(context.Background(), SharedSessionBrowserObserverRequest{
		SessionID:        sessionID,
		SelectedInfo:     BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		BindingRoute:     route,
		RequestedProfile: "isolated",
		IncludeStatus:    true,
		IncludeProfiles:  true,
	})

	if len(full.View.Session.Routes) != 1 || len(watch.View.Session.Routes) != 0 {
		t.Fatalf("expected cached broader watch loop to satisfy narrower watch without leaking session view, got full=%#v watch=%#v", full.View.Session, watch.View.Session)
	}
	if watch.View.Binding.ReferenceTime.IsZero() || watch.ReferenceTime.IsZero() {
		t.Fatalf("expected narrower watch projection to preserve binding/reference time, got %#v", watch)
	}
	if backend.statusRequestCount() != 1 || backend.profilesRequestCount() != 1 {
		t.Fatalf("expected cached broader watch loop to avoid extra backend polling, got status=%d profiles=%d", backend.statusRequestCount(), backend.profilesRequestCount())
	}
	if runRegistry.callCount() != 1 {
		t.Fatalf("expected cached broader watch loop to avoid extra run snapshot, got %d", runRegistry.callCount())
	}
}

func TestSharedSessionBrowserWatchManagerObserveWatchReusesCompatibleWatchLoopInFlight(t *testing.T) {
	runBlock := make(chan struct{})
	runStarted := make(chan struct{}, 1)
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
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	sessionID := "sess-watch-loop-superset-in-flight"
	runRegistry := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			sessionID: {{RunID: "run-1", Status: "running"}},
		}},
		started: runStarted,
		block:   runBlock,
	}
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node"}

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

	manager := NewSharedSessionBrowserWatchManager(backend, sessionRegistry, runRegistry, stateRegistry, time.Minute)
	fullReq := SharedSessionBrowserObserverRequest{
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

	fullCh := make(chan SharedSessionBrowserWatchLoopObservation, 1)
	watchCh := make(chan SharedSessionBrowserWatchObservation, 1)
	go func() {
		fullCh <- manager.ObserveWatchLoop(context.Background(), fullReq)
	}()

	select {
	case <-runStarted:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for broader watch-loop projection to start")
	}

	go func() {
		watchCh <- manager.ObserveWatch(context.Background(), SharedSessionBrowserObserverRequest{
			SessionID:        sessionID,
			SelectedInfo:     fullReq.SelectedInfo,
			BindingRoute:     route,
			RequestedProfile: "isolated",
			IncludeStatus:    true,
			IncludeProfiles:  true,
		})
	}()
	time.Sleep(20 * time.Millisecond)

	if runRegistry.callCount() != 1 {
		t.Fatalf("expected narrower watch to reuse broader in-flight watch-loop projection, got %d run snapshots", runRegistry.callCount())
	}

	close(runBlock)
	full := <-fullCh
	watch := <-watchCh

	if len(full.View.Session.Routes) != 1 || len(watch.View.Session.Routes) != 0 {
		t.Fatalf("expected projected narrower watch to clear session view while reusing broader in-flight loop, got full=%#v watch=%#v", full.View.Session, watch.View.Session)
	}
	if backend.statusRequestCount() != 1 || backend.profilesRequestCount() != 1 {
		t.Fatalf("expected broader in-flight watch loop reuse to avoid extra backend polling, got status=%d profiles=%d", backend.statusRequestCount(), backend.profilesRequestCount())
	}
}

func TestSharedSessionBrowserWatchManagerObserveWatchLoopCachesShortLivedSessionView(t *testing.T) {
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
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	sessionID := "sess-watch-loop-cache"
	runRegistry := testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
		sessionID: {{RunID: "run-1", Status: "running"}},
	}}

	sessionRegistry.TrackTab(sessionID, BrowserSessionTarget{
		TabIndex:   1,
		URL:        "https://example.com",
		Title:      "Before Cache",
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

	manager := NewSharedSessionBrowserWatchManager(backend, sessionRegistry, runRegistry, stateRegistry, time.Minute)
	req := SharedSessionBrowserObserverRequest{
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
	}

	first := manager.ObserveWatchLoop(context.Background(), req)
	sessionRegistry.TrackTab(sessionID, BrowserSessionTarget{
		TabIndex:   1,
		URL:        "https://example.com",
		Title:      "After Mutation",
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
		BrowserApp: "Chromium",
	}, true)
	second := manager.ObserveWatchLoop(context.Background(), req)

	if len(first.View.Session.Routes) != 1 || len(second.View.Session.Routes) != 1 {
		t.Fatalf("expected cached watch loop to keep single route snapshot, got first=%#v second=%#v", first.View.Session, second.View.Session)
	}
	if len(first.View.Session.Routes[0].Targets) != 1 || len(second.View.Session.Routes[0].Targets) != 1 {
		t.Fatalf("expected cached watch loop to keep single route target snapshot, got first=%#v second=%#v", first.View.Session.Routes, second.View.Session.Routes)
	}
	if first.View.Session.Routes[0].Targets[0].Title != "Before Cache" {
		t.Fatalf("expected first watch loop snapshot to capture initial title, got %#v", first.View.Session.Routes)
	}
	if second.View.Session.Routes[0].Targets[0].Title != first.View.Session.Routes[0].Targets[0].Title {
		t.Fatalf("expected cached watch loop to preserve session view title, got first=%q second=%q", first.View.Session.Routes[0].Targets[0].Title, second.View.Session.Routes[0].Targets[0].Title)
	}
}

func TestSharedSessionBrowserWatchManagerObserveRawStatusAndProfilesCacheShortLivedSourceState(t *testing.T) {
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
	manager := NewSharedSessionBrowserWatchManager(backend, nil, nil, nil, time.Minute)

	firstStatus := manager.ObserveRawStatus(context.Background(), "isolated")
	secondStatus := manager.ObserveRawStatus(context.Background(), "isolated")
	firstProfiles := manager.ObserveRawProfiles(context.Background(), "isolated")
	secondProfiles := manager.ObserveRawProfiles(context.Background(), "isolated")

	if len(backend.statusReqs) != 1 {
		t.Fatalf("expected raw status cache to suppress duplicate polling, got %d calls", len(backend.statusReqs))
	}
	if len(backend.profilesReqs) != 1 {
		t.Fatalf("expected raw profiles cache to suppress duplicate polling, got %d calls", len(backend.profilesReqs))
	}
	if !secondStatus.ObservedAt.Equal(firstStatus.ObservedAt) {
		t.Fatalf("expected cached raw status observation to preserve observed_at, got first=%v second=%v", firstStatus.ObservedAt, secondStatus.ObservedAt)
	}
	if !secondProfiles.ObservedAt.Equal(firstProfiles.ObservedAt) {
		t.Fatalf("expected cached raw profiles observation to preserve observed_at, got first=%v second=%v", firstProfiles.ObservedAt, secondProfiles.ObservedAt)
	}
}

func TestSharedSessionBrowserWatchManagerObserveRawStatusCoalescesConcurrentPolling(t *testing.T) {
	statusBlock := make(chan struct{})
	statusStarted := make(chan struct{}, 1)
	backend := &statusProfilesObservationTestBackend{
		statusResp: BrowserProfileStatusResult{
			Backend:   "proxy",
			Profile:   "isolated",
			Status:    "running",
			Running:   true,
			Connected: true,
		},
		statusStarted: statusStarted,
		statusBlock:   statusBlock,
	}
	manager := NewSharedSessionBrowserWatchManager(backend, nil, nil, nil, time.Minute)

	firstCh := make(chan SharedSessionBrowserRawStatusObservation, 1)
	secondCh := make(chan SharedSessionBrowserRawStatusObservation, 1)
	go func() {
		firstCh <- manager.ObserveRawStatus(context.Background(), "isolated")
	}()

	select {
	case <-statusStarted:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for first raw status poll to start")
	}

	go func() {
		secondCh <- manager.ObserveRawStatus(context.Background(), "isolated")
	}()
	time.Sleep(20 * time.Millisecond)

	if got := backend.statusRequestCount(); got != 1 {
		t.Fatalf("expected concurrent raw status observe to coalesce into one backend poll, got %d", got)
	}

	close(statusBlock)
	first := <-firstCh
	second := <-secondCh

	if first.Status == nil || second.Status == nil {
		t.Fatalf("expected successful concurrent raw status observations, got first=%#v second=%#v", first, second)
	}
	if !second.ObservedAt.Equal(first.ObservedAt) {
		t.Fatalf("expected concurrent raw status observe to preserve observed_at, got first=%v second=%v", first.ObservedAt, second.ObservedAt)
	}
}

func TestSharedSessionBrowserWatchManagerObserveEventCycleCoalescesConcurrentRawPolling(t *testing.T) {
	statusBlock := make(chan struct{})
	profilesBlock := make(chan struct{})
	statusStarted := make(chan struct{}, 1)
	profilesStarted := make(chan struct{}, 1)
	backend := &statusProfilesObservationTestBackend{
		statusResp: BrowserProfileStatusResult{
			Backend:   "proxy",
			Profile:   "isolated",
			Status:    "running",
			Running:   true,
			Connected: true,
		},
		statusStarted: statusStarted,
		statusBlock:   statusBlock,
		profilesResp: BrowserProfilesResult{
			Backend:        "proxy",
			DefaultProfile: "isolated",
			Profiles: []BrowserProfileInfo{
				{Profile: "isolated", BrowserApp: "Chromium", Status: "connected", Running: true, Connected: true},
			},
		},
		profilesStarted: profilesStarted,
		profilesBlock:   profilesBlock,
	}
	manager := NewSharedSessionBrowserWatchManager(backend, nil, nil, nil, time.Minute)
	req := SharedSessionBrowserObserverRequest{
		SessionID:        "sess-concurrent-cycle",
		SelectedInfo:     BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		BindingRoute:     BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node"},
		RequestedProfile: "isolated",
		IncludeStatus:    true,
		IncludeProfiles:  true,
	}

	firstCh := make(chan SharedSessionBrowserEventCycleObservation, 1)
	secondCh := make(chan SharedSessionBrowserEventCycleObservation, 1)
	go func() {
		firstCh <- manager.ObserveEventCycle(context.Background(), req)
	}()

	select {
	case <-statusStarted:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for first event-cycle status poll to start")
	}

	go func() {
		secondCh <- manager.ObserveEventCycle(context.Background(), req)
	}()
	time.Sleep(20 * time.Millisecond)

	if got := backend.statusRequestCount(); got != 1 {
		t.Fatalf("expected concurrent event-cycle observe to coalesce status polling, got %d", got)
	}

	close(statusBlock)

	select {
	case <-profilesStarted:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for first event-cycle profiles poll to start")
	}
	time.Sleep(20 * time.Millisecond)

	if got := backend.profilesRequestCount(); got != 1 {
		t.Fatalf("expected concurrent event-cycle observe to coalesce profiles polling, got %d", got)
	}

	close(profilesBlock)
	first := <-firstCh
	second := <-secondCh

	if first.Observation.Status == nil || second.Observation.Status == nil || first.Observation.Profiles == nil || second.Observation.Profiles == nil {
		t.Fatalf("expected successful concurrent event-cycle observations, got first=%#v second=%#v", first, second)
	}
	if !second.ReferenceTime.Equal(first.ReferenceTime) {
		t.Fatalf("expected concurrent event-cycle observe to preserve reference time, got first=%v second=%v", first.ReferenceTime, second.ReferenceTime)
	}
}

func TestSharedSessionBrowserWatchManagerObserveEventCycleCoalescesConcurrentResolution(t *testing.T) {
	resolutionBlock := make(chan struct{})
	resolutionStarted := make(chan struct{}, 1)
	backend := &statusProfilesObservationTestBackend{
		statusResp: BrowserProfileStatusResult{
			Backend:   "proxy",
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
	stateRegistry := &countingBrowserSessionStateRegistry{
		BrowserSessionStateRegistry: NewBrowserSessionStateRegistry(),
		started:                     resolutionStarted,
		block:                       resolutionBlock,
	}
	manager := NewSharedSessionBrowserWatchManager(backend, nil, nil, stateRegistry, time.Minute)
	req := SharedSessionBrowserObserverRequest{
		SessionID:        "sess-concurrent-cycle-resolution",
		SelectedInfo:     BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		BindingRoute:     BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node"},
		RequestedProfile: "isolated",
		IncludeStatus:    true,
		IncludeProfiles:  true,
	}

	firstCh := make(chan SharedSessionBrowserEventCycleObservation, 1)
	secondCh := make(chan SharedSessionBrowserEventCycleObservation, 1)
	go func() {
		firstCh <- manager.ObserveEventCycle(context.Background(), req)
	}()

	select {
	case <-resolutionStarted:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for first event-cycle resolution to start")
	}

	go func() {
		secondCh <- manager.ObserveEventCycle(context.Background(), req)
	}()
	time.Sleep(20 * time.Millisecond)

	if got := stateRegistry.resolutionCallCount(); got != 1 {
		t.Fatalf("expected concurrent event-cycle observe to coalesce registry resolution, got %d", got)
	}

	close(resolutionBlock)
	first := <-firstCh
	second := <-secondCh

	if first.Observation.Status == nil || second.Observation.Status == nil || first.Observation.Profiles == nil || second.Observation.Profiles == nil {
		t.Fatalf("expected successful concurrent event-cycle observations, got first=%#v second=%#v", first, second)
	}
	if !second.ReferenceTime.Equal(first.ReferenceTime) {
		t.Fatalf("expected concurrent event-cycle resolution to preserve reference time, got first=%v second=%v", first.ReferenceTime, second.ReferenceTime)
	}
}

func TestSharedSessionBrowserWatchManagerObserveWatchLoopCoalescesConcurrentProjection(t *testing.T) {
	runBlock := make(chan struct{})
	runStarted := make(chan struct{}, 1)
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
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	sessionID := "sess-watch-loop-concurrent-projection"
	runRegistry := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			sessionID: {{RunID: "run-1", Status: "running"}},
		}},
		started: runStarted,
		block:   runBlock,
	}

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

	manager := NewSharedSessionBrowserWatchManager(backend, sessionRegistry, runRegistry, stateRegistry, time.Minute)
	req := SharedSessionBrowserObserverRequest{
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
	}

	firstCh := make(chan SharedSessionBrowserWatchLoopObservation, 1)
	secondCh := make(chan SharedSessionBrowserWatchLoopObservation, 1)
	go func() {
		firstCh <- manager.ObserveWatchLoop(context.Background(), req)
	}()

	select {
	case <-runStarted:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for first watch-loop projection to start")
	}

	go func() {
		secondCh <- manager.ObserveWatchLoop(context.Background(), req)
	}()
	time.Sleep(20 * time.Millisecond)

	if got := runRegistry.callCount(); got != 1 {
		t.Fatalf("expected concurrent watch-loop observe to coalesce projection work, got %d run snapshots", got)
	}

	close(runBlock)
	first := <-firstCh
	second := <-secondCh

	if len(first.View.Session.Routes) != 1 || len(second.View.Session.Routes) != 1 {
		t.Fatalf("expected successful concurrent watch-loop observations, got first=%#v second=%#v", first.View.Session, second.View.Session)
	}
	if got := runRegistry.callCount(); got != 1 {
		t.Fatalf("expected concurrent watch-loop observe to reuse single binding/session-view projection, got %d run snapshots", got)
	}
	if !second.ReferenceTime.Equal(first.ReferenceTime) {
		t.Fatalf("expected concurrent watch-loop projection to preserve reference time, got first=%v second=%v", first.ReferenceTime, second.ReferenceTime)
	}
}

func TestSharedSessionBrowserWatchManagerObserveWatchLoopReusesBindingForMatchingSessionView(t *testing.T) {
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
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	sessionID := "sess-watch-loop-reuse-binding"
	runRegistry := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			sessionID: {{RunID: "run-1", Status: "running"}},
		}},
	}

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

	manager := NewSharedSessionBrowserWatchManager(backend, sessionRegistry, runRegistry, stateRegistry, time.Minute)
	req := SharedSessionBrowserObserverRequest{
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
	}

	loop := manager.ObserveWatchLoop(context.Background(), req)

	if got := runRegistry.callCount(); got != 1 {
		t.Fatalf("expected matching session-view scope to reuse binding snapshot, got %d run snapshots", got)
	}
	if len(loop.View.Session.Routes) != 1 || len(loop.View.Session.Runs) != 1 || len(loop.View.Session.Profiles) != 1 {
		t.Fatalf("expected reused binding snapshot to populate session view, got %#v", loop.View.Session)
	}
	if !loop.View.Session.Profiles[0].Selected {
		t.Fatalf("expected reused binding snapshot to preserve selected profile metadata, got %#v", loop.View.Session.Profiles)
	}
}

func TestSharedSessionBrowserWatchManagerObserveRawStartInvalidatesEventCycleCache(t *testing.T) {
	backend := &executionTestBackend{
		statusResp: BrowserProfileStatusResult{
			Backend:   "proxy",
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
		startResp: BrowserProfileStatusResult{
			Backend:   "proxy",
			Profile:   "isolated",
			Status:    "started",
			Running:   true,
			Connected: false,
		},
	}
	manager := NewSharedSessionBrowserWatchManager(backend, nil, nil, nil, time.Minute)
	req := SharedSessionBrowserObserverRequest{
		SessionID:        "sess-cache-invalidate",
		SelectedInfo:     BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		BindingRoute:     BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node"},
		RequestedProfile: "isolated",
		IncludeStatus:    true,
		IncludeProfiles:  true,
	}

	manager.ObserveEventCycle(context.Background(), req)
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected initial event cycle to poll once, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}

	startObservation := manager.ObserveExecutionStart(
		context.Background(),
		SharedSessionBrowserExecutionRequest{
			SelectedInfo: BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
		},
		"isolated",
		"started",
		BrowserProfileStatusResult{},
	)
	if startObservation.Err != nil {
		t.Fatalf("expected successful start observation, got %v", startObservation.Err)
	}

	manager.ObserveEventCycle(context.Background(), req)
	if len(backend.startReqs) != 1 {
		t.Fatalf("expected start observation to reach backend once, got %#v", backend.startReqs)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected start observation to invalidate cached event cycle while reusing lifecycle-seeded status/profiles, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
	refreshed := manager.ObserveEventCycle(context.Background(), req)
	if refreshed.Observation.Profiles == nil || len(refreshed.Observation.Profiles.Profiles) != 1 || refreshed.Observation.Profiles.Profiles[0].Status != "starting" || !refreshed.Observation.Profiles.Profiles[0].Running || refreshed.Observation.Profiles.Profiles[0].Connected {
		t.Fatalf("expected refreshed event cycle to expose lifecycle-seeded starting profile state, got %#v", refreshed.Observation.Profiles)
	}
}

func TestSharedSessionBrowserWatchManagerObserveRawStartInvalidatesWatchLoopCache(t *testing.T) {
	backend := &executionTestBackend{
		statusResp: BrowserProfileStatusResult{
			Backend:   "proxy",
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
		startResp: BrowserProfileStatusResult{
			Backend:   "proxy",
			Profile:   "isolated",
			Status:    "started",
			Running:   true,
			Connected: false,
		},
	}
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	sessionID := "sess-watch-loop-cache-invalidate"
	runRegistry := testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
		sessionID: {{RunID: "run-1", Status: "running"}},
	}}

	sessionRegistry.TrackTab(sessionID, BrowserSessionTarget{
		TabIndex:   1,
		URL:        "https://example.com",
		Title:      "Before Start",
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

	manager := NewSharedSessionBrowserWatchManager(backend, sessionRegistry, runRegistry, stateRegistry, time.Minute)
	req := SharedSessionBrowserObserverRequest{
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
	}

	first := manager.ObserveWatchLoop(context.Background(), req)
	sessionRegistry.TrackTab(sessionID, BrowserSessionTarget{
		TabIndex:   1,
		URL:        "https://example.com",
		Title:      "After Start",
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
		BrowserApp: "Chromium",
	}, true)
	manager.ObserveExecutionStart(
		context.Background(),
		SharedSessionBrowserExecutionRequest{
			SelectedInfo: BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
		},
		"isolated",
		"started",
		BrowserProfileStatusResult{},
	)
	second := manager.ObserveWatchLoop(context.Background(), req)

	if len(first.View.Session.Routes) != 1 || len(second.View.Session.Routes) != 1 {
		t.Fatalf("expected watch-loop route snapshots before and after invalidation, got first=%#v second=%#v", first.View.Session, second.View.Session)
	}
	if len(first.View.Session.Routes[0].Targets) != 1 || len(second.View.Session.Routes[0].Targets) != 1 {
		t.Fatalf("expected watch-loop target snapshots before and after invalidation, got first=%#v second=%#v", first.View.Session.Routes, second.View.Session.Routes)
	}
	if first.View.Session.Routes[0].Targets[0].Title != "Before Start" {
		t.Fatalf("expected initial watch-loop snapshot to capture pre-start title, got %#v", first.View.Session.Routes)
	}
	if second.View.Session.Routes[0].Targets[0].Title != "After Start" {
		t.Fatalf("expected start observation to invalidate cached watch loop, got %#v", second.View.Session.Routes)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected refreshed watch loop to reuse lifecycle-seeded status/profiles without extra polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}

func TestSharedSessionBrowserObserverManagerObserveExecutionStatusUsesResolvedFallbackOnError(t *testing.T) {
	backend := &executionTestBackend{statusErr: errors.New("status unavailable")}
	manager := NewSharedSessionBrowserObserverManager(nil, nil, nil, time.Minute)
	observation := manager.ObserveExecutionStatus(context.Background(), backend, SharedSessionBrowserExecutionRequest{
		SelectedInfo: BrowserRuntimeInfo{
			Backend: "proxy",
			Profile: "isolated",
			Target:  "node",
		},
		HealthInput: SharedSessionBrowserHealthInput{
			Profiles: []SharedSessionBrowserProfileState{{
				Backend:       "proxy",
				Profile:       "isolated",
				RuntimeTarget: "node",
				Status:        "reconnecting",
				Running:       true,
				Connected:     false,
			}},
		},
		ReconnectWindow: time.Minute,
	}, "isolated", BrowserProfileStatusResult{})

	if observation.StatusErr == nil || observation.StatusErr.Error() != "status unavailable" {
		t.Fatalf("expected raw status error, got %#v", observation.StatusErr)
	}
	if observation.HasStatus {
		t.Fatalf("expected no raw status on error, got %#v", observation)
	}
	if observation.ResolvedStatus.Profile != "isolated" || observation.ResolvedStatus.Status != "reconnecting" || !observation.ResolvedStatus.Running || observation.ResolvedStatus.Connected {
		t.Fatalf("expected manager execution status to preserve resolved fallback, got %#v", observation.ResolvedStatus)
	}
}

func TestSharedSessionBrowserWatchManagerObserveExecutionStartUsesLifecycleDecisionStatus(t *testing.T) {
	backend := &executionTestBackend{
		startResp: BrowserProfileStatusResult{
			Backend:   "proxy",
			Profile:   "isolated",
			Status:    "started",
			Running:   true,
			Connected: false,
		},
	}
	manager := NewSharedSessionBrowserWatchManager(backend, nil, nil, nil, time.Minute)
	observation := manager.ObserveExecutionStart(context.Background(), SharedSessionBrowserExecutionRequest{
		SelectedInfo: BrowserRuntimeInfo{
			Backend: "proxy",
			Target:  "node",
		},
	}, "isolated", "started", BrowserProfileStatusResult{})

	if observation.Err != nil {
		t.Fatalf("expected successful bound watch-manager start observation, got %v", observation.Err)
	}
	if observation.Profile != "isolated" || observation.Status.Status != "starting" || !observation.Status.Running || observation.Status.Connected || observation.Status.Note != "start requested" || observation.Ready {
		t.Fatalf("expected lifecycle-owned starting observation from bound watch manager, got %#v", observation)
	}
}

func TestSharedSessionBrowserWatchManagerObserveExecutionStartUsesFallbackWhenRawLifecycleStatusMissing(t *testing.T) {
	backend := &executionTestBackend{
		rawStart: func(ctx context.Context, profile string) SharedSessionBrowserRawLifecycleObservation {
			return SharedSessionBrowserRawLifecycleObservation{
				Profile:    profile,
				ObservedAt: time.Date(2026, time.April, 10, 13, 2, 0, 0, time.UTC),
			}
		},
	}
	manager := NewSharedSessionBrowserWatchManager(backend, nil, nil, nil, time.Minute)
	observation := manager.ObserveExecutionStart(context.Background(), SharedSessionBrowserExecutionRequest{
		SelectedInfo: BrowserRuntimeInfo{
			Backend: "proxy",
			Target:  "node",
		},
	}, "isolated", "started", BrowserProfileStatusResult{})

	if observation.Err != nil {
		t.Fatalf("expected fallback start observation to stay successful, got %v", observation.Err)
	}
	if len(backend.startReqs) != 0 {
		t.Fatalf("expected provided raw lifecycle observation to avoid direct runtime start fallback, got %#v", backend.startReqs)
	}
	if observation.Profile != "isolated" || observation.Status.Status != "starting" || observation.Status.Connected || observation.Status.Note != "start requested" || observation.Ready {
		t.Fatalf("expected missing raw lifecycle status to collapse into lifecycle-owned starting fallback, got %#v", observation)
	}
}

func TestSharedSessionBrowserWatchManagerObserveExecutionStartSeedsLifecycleStatusForFollowUpExecutionStatus(t *testing.T) {
	backend := &executionTestBackend{
		startResp: BrowserProfileStatusResult{
			Backend:   "proxy",
			Profile:   "isolated",
			Status:    "started",
			Running:   true,
			Connected: false,
		},
	}
	manager := NewSharedSessionBrowserWatchManager(backend, nil, nil, nil, time.Minute)
	req := SharedSessionBrowserExecutionRequest{
		SelectedInfo: BrowserRuntimeInfo{
			Backend: "proxy",
			Profile: "isolated",
			Target:  "node",
		},
	}

	startObservation := manager.ObserveExecutionStart(context.Background(), req, "isolated", "started", BrowserProfileStatusResult{})
	if startObservation.Err != nil {
		t.Fatalf("expected successful start observation, got %v", startObservation.Err)
	}

	statusObservation := manager.ObserveExecutionStatus(context.Background(), req, "isolated", BrowserProfileStatusResult{})
	if statusObservation.StatusErr != nil {
		t.Fatalf("expected follow-up execution status to reuse seeded lifecycle status, got %v", statusObservation.StatusErr)
	}
	if len(backend.startReqs) != 1 || len(backend.statusReqs) != 0 {
		t.Fatalf("expected follow-up execution status to reuse lifecycle-seeded raw status without backend polling, got start=%d status=%d", len(backend.startReqs), len(backend.statusReqs))
	}
	if !statusObservation.HasStatus || statusObservation.Status.Profile != "isolated" || statusObservation.Status.Status != "starting" || !statusObservation.Status.Running || statusObservation.Status.Connected {
		t.Fatalf("expected follow-up execution status to expose lifecycle-owned starting state, got %#v", statusObservation)
	}
}

func TestSharedSessionBrowserWatchManagerObserveExecutionStartReusesRawStartInFlight(t *testing.T) {
	release := make(chan struct{})
	entered := make(chan struct{}, 2)
	var (
		mu    sync.Mutex
		calls int
	)
	backend := &executionTestBackend{
		rawStart: func(ctx context.Context, profile string) SharedSessionBrowserRawLifecycleObservation {
			mu.Lock()
			calls++
			mu.Unlock()
			select {
			case entered <- struct{}{}:
			default:
			}
			<-release
			return SharedSessionBrowserRawLifecycleObservation{
				Profile: profile,
				Status: &BrowserProfileStatusResult{
					Backend:   "proxy",
					Profile:   profile,
					Status:    "started",
					Running:   true,
					Connected: false,
				},
				ObservedAt: time.Date(2026, time.April, 10, 13, 0, 0, 0, time.UTC),
			}
		},
	}
	manager := NewSharedSessionBrowserWatchManager(backend, nil, nil, nil, time.Minute)
	req := SharedSessionBrowserExecutionRequest{
		SelectedInfo: BrowserRuntimeInfo{
			Backend: "proxy",
			Target:  "node",
		},
	}
	results := make(chan SharedSessionBrowserExecutionLifecycleObservation, 2)
	var wg sync.WaitGroup
	invoke := func() {
		defer wg.Done()
		results <- manager.ObserveExecutionStart(context.Background(), req, "isolated", "started", BrowserProfileStatusResult{})
	}

	wg.Add(1)
	go invoke()
	<-entered

	wg.Add(1)
	go invoke()
	select {
	case <-entered:
		close(release)
		t.Fatalf("expected concurrent execution start to reuse in-flight raw start observation")
	case <-time.After(50 * time.Millisecond):
	}

	close(release)
	wg.Wait()
	close(results)

	mu.Lock()
	defer mu.Unlock()
	if calls != 1 {
		t.Fatalf("expected single backend raw start call for concurrent execution start observations, got %d", calls)
	}
	for observation := range results {
		if observation.Err != nil {
			t.Fatalf("expected successful shared in-flight execution start observation, got %v", observation.Err)
		}
		if observation.Profile != "isolated" || observation.Status.Status != "starting" || !observation.Status.Running || observation.Status.Connected || observation.Ready {
			t.Fatalf("expected lifecycle-owned starting observation from shared in-flight raw start, got %#v", observation)
		}
	}
}

func TestSharedSessionBrowserWatchManagerObserveExecutionStopReusesRawStopInFlight(t *testing.T) {
	release := make(chan struct{})
	entered := make(chan struct{}, 2)
	var (
		mu    sync.Mutex
		calls int
	)
	backend := &executionTestBackend{
		rawStop: func(ctx context.Context, profile string) SharedSessionBrowserRawLifecycleObservation {
			mu.Lock()
			calls++
			mu.Unlock()
			select {
			case entered <- struct{}{}:
			default:
			}
			<-release
			return SharedSessionBrowserRawLifecycleObservation{
				Profile: profile,
				Status: &BrowserProfileStatusResult{
					Backend:   "proxy",
					Profile:   profile,
					Status:    "stopped",
					Running:   false,
					Connected: false,
				},
				ObservedAt: time.Date(2026, time.April, 10, 13, 1, 0, 0, time.UTC),
			}
		},
	}
	manager := NewSharedSessionBrowserWatchManager(backend, nil, nil, nil, time.Minute)
	req := SharedSessionBrowserExecutionRequest{
		SelectedInfo: BrowserRuntimeInfo{
			Backend: "proxy",
			Target:  "node",
		},
	}
	results := make(chan SharedSessionBrowserExecutionLifecycleObservation, 2)
	var wg sync.WaitGroup
	invoke := func() {
		defer wg.Done()
		results <- manager.ObserveExecutionStop(context.Background(), req, "isolated", "stopped", BrowserProfileStatusResult{})
	}

	wg.Add(1)
	go invoke()
	<-entered

	wg.Add(1)
	go invoke()
	select {
	case <-entered:
		close(release)
		t.Fatalf("expected concurrent execution stop to reuse in-flight raw stop observation")
	case <-time.After(50 * time.Millisecond):
	}

	close(release)
	wg.Wait()
	close(results)

	mu.Lock()
	defer mu.Unlock()
	if calls != 1 {
		t.Fatalf("expected single backend raw stop call for concurrent execution stop observations, got %d", calls)
	}
	for observation := range results {
		if observation.Err != nil {
			t.Fatalf("expected successful shared in-flight execution stop observation, got %v", observation.Err)
		}
		if observation.Profile != "isolated" || observation.Status.Status != "stopped" || observation.Status.Running || observation.Status.Connected || !observation.Ready {
			t.Fatalf("expected lifecycle-owned stopped observation from shared in-flight raw stop, got %#v", observation)
		}
	}
}

func TestSharedSessionBrowserWatchManagerObserveExecutionStopUsesFallbackWhenRawLifecycleStatusMissing(t *testing.T) {
	backend := &executionTestBackend{
		rawStop: func(ctx context.Context, profile string) SharedSessionBrowserRawLifecycleObservation {
			return SharedSessionBrowserRawLifecycleObservation{
				Profile:    profile,
				ObservedAt: time.Date(2026, time.April, 10, 13, 3, 0, 0, time.UTC),
			}
		},
	}
	manager := NewSharedSessionBrowserWatchManager(backend, nil, nil, nil, time.Minute)
	observation := manager.ObserveExecutionStop(context.Background(), SharedSessionBrowserExecutionRequest{
		SelectedInfo: BrowserRuntimeInfo{
			Backend: "proxy",
			Target:  "node",
		},
	}, "isolated", "stopped", BrowserProfileStatusResult{})

	if observation.Err != nil {
		t.Fatalf("expected fallback stop observation to stay successful, got %v", observation.Err)
	}
	if len(backend.stopReqs) != 0 {
		t.Fatalf("expected provided raw lifecycle observation to avoid direct runtime stop fallback, got %#v", backend.stopReqs)
	}
	if observation.Profile != "isolated" || observation.Status.Status != "stopped" || observation.Status.Running || observation.Status.Connected || !observation.Ready {
		t.Fatalf("expected missing raw lifecycle status to collapse into lifecycle-owned stopped fallback, got %#v", observation)
	}
}

func TestSharedSessionBrowserWatchManagerObserveExecutionStopSeedsLifecycleStatusForFollowUpExecutionStatus(t *testing.T) {
	backend := &executionTestBackend{
		stopResp: BrowserProfileStatusResult{
			Backend:   "proxy",
			Profile:   "isolated",
			Status:    "stopped",
			Running:   false,
			Connected: false,
		},
	}
	manager := NewSharedSessionBrowserWatchManager(backend, nil, nil, nil, time.Minute)
	req := SharedSessionBrowserExecutionRequest{
		SelectedInfo: BrowserRuntimeInfo{
			Backend: "proxy",
			Profile: "isolated",
			Target:  "node",
		},
	}

	stopObservation := manager.ObserveExecutionStop(context.Background(), req, "isolated", "stopped", BrowserProfileStatusResult{
		Backend:   "proxy",
		Profile:   "isolated",
		Status:    "running",
		Running:   true,
		Connected: true,
	})
	if stopObservation.Err != nil {
		t.Fatalf("expected successful stop observation, got %v", stopObservation.Err)
	}

	statusObservation := manager.ObserveExecutionStatus(context.Background(), req, "isolated", BrowserProfileStatusResult{})
	if statusObservation.StatusErr != nil {
		t.Fatalf("expected follow-up execution status to reuse seeded stopped lifecycle status, got %v", statusObservation.StatusErr)
	}
	if len(backend.stopReqs) != 1 || len(backend.statusReqs) != 0 {
		t.Fatalf("expected follow-up execution status to reuse lifecycle-seeded stopped status without backend polling, got stop=%d status=%d", len(backend.stopReqs), len(backend.statusReqs))
	}
	if !statusObservation.HasStatus || statusObservation.Status.Profile != "isolated" || statusObservation.Status.Status != "stopped" || statusObservation.Status.Running || statusObservation.Status.Connected {
		t.Fatalf("expected follow-up execution status to expose lifecycle-owned stopped state, got %#v", statusObservation)
	}
}

func TestSharedSessionBrowserWatchManagerObserveExecutionStopSeedsLifecycleProfilesForFollowUpEventCycle(t *testing.T) {
	backend := &executionTestBackend{
		statusResp: BrowserProfileStatusResult{
			Backend:   "proxy",
			Profile:   "isolated",
			Status:    "running",
			Running:   true,
			Connected: true,
		},
		profilesResp: BrowserProfilesResult{
			Backend:        "proxy",
			DefaultProfile: "isolated",
			Profiles: []BrowserProfileInfo{{
				Profile:   "isolated",
				Status:    "running",
				Running:   true,
				Connected: true,
			}},
		},
		stopResp: BrowserProfileStatusResult{
			Backend:   "proxy",
			Profile:   "isolated",
			Status:    "stopped",
			Running:   false,
			Connected: false,
		},
	}
	manager := NewSharedSessionBrowserWatchManager(backend, nil, nil, nil, time.Minute)
	req := SharedSessionBrowserObserverRequest{
		SessionID:        "sess-stop-seed-profiles",
		SelectedInfo:     BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		BindingRoute:     BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node"},
		RequestedProfile: "isolated",
		IncludeStatus:    true,
		IncludeProfiles:  true,
	}

	initial := manager.ObserveEventCycle(context.Background(), req)
	if initial.Observation.Profiles == nil || len(initial.Observation.Profiles.Profiles) != 1 {
		t.Fatalf("expected initial event cycle to expose one discovered profile, got %#v", initial.Observation.Profiles)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected initial event cycle to poll backend once, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}

	stopObservation := manager.ObserveExecutionStop(context.Background(), SharedSessionBrowserExecutionRequest{
		SelectedInfo: BrowserRuntimeInfo{
			Backend: "proxy",
			Profile: "isolated",
			Target:  "node",
		},
	}, "isolated", "stopped", BrowserProfileStatusResult{
		Backend:   "proxy",
		Profile:   "isolated",
		Status:    "running",
		Running:   true,
		Connected: true,
	})
	if stopObservation.Err != nil {
		t.Fatalf("expected successful stop observation, got %v", stopObservation.Err)
	}

	refreshed := manager.ObserveEventCycle(context.Background(), req)
	if len(backend.stopReqs) != 1 || len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected refreshed event cycle to reuse lifecycle-seeded stop status/profiles without extra polling, got stop=%d status=%d profiles=%d", len(backend.stopReqs), len(backend.statusReqs), len(backend.profilesReqs))
	}
	if refreshed.Observation.Profiles == nil || len(refreshed.Observation.Profiles.Profiles) != 1 {
		t.Fatalf("expected refreshed event cycle to retain one lifecycle-seeded profile, got %#v", refreshed.Observation.Profiles)
	}
	if refreshed.Observation.Profiles.Profiles[0].Profile != "isolated" || refreshed.Observation.Profiles.Profiles[0].Status != "stopped" || refreshed.Observation.Profiles.Profiles[0].Running || refreshed.Observation.Profiles.Profiles[0].Connected {
		t.Fatalf("expected refreshed event cycle to expose lifecycle-seeded stopped profile state, got %#v", refreshed.Observation.Profiles.Profiles[0])
	}
}

func TestSharedSessionBrowserObserverManagerResolveExecutionEventUsesRegistryResolution(t *testing.T) {
	manager := NewSharedSessionBrowserObserverManager(nil, nil, NewBrowserSessionStateRegistry(), time.Minute)
	resolution := manager.ResolveExecutionEvent(
		"sess-manager-execution",
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
	)

	if !resolution.HasSyncedState {
		t.Fatalf("expected manager execution resolution to return synced state")
	}
	if resolution.ResolvedStatus.Profile != "isolated" || resolution.ResolvedStatus.Status != "starting" || !resolution.ResolvedStatus.Running || resolution.ResolvedStatus.Connected {
		t.Fatalf("expected manager execution resolution to use lifecycle-owned starting state, got %#v", resolution.ResolvedStatus)
	}
	if len(resolution.Snapshot) != 2 || resolution.Snapshot[0].Profile != "isolated" || resolution.Snapshot[0].Status != "starting" {
		t.Fatalf("expected manager execution resolution to return scoped snapshot, got %#v", resolution.Snapshot)
	}
}

func TestSharedSessionBrowserObserverManagerSyncProfileStatusEventInvalidatesManagedCurrentTargetSelection(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	manager := NewSharedSessionBrowserObserverManager(sessionRegistry, nil, stateRegistry, 54*time.Second)
	sessionID := "sess-manager-status-sync"
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
		t.Fatalf("expected tracked current target before invalidation")
	}

	state, ok := manager.SyncProfileStatusEvent(
		sessionID,
		BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		BrowserProfileStatusResult{
			Backend:   "proxy",
			Profile:   "isolated",
			Status:    "reconnecting",
			Running:   true,
			Connected: false,
		},
		time.Date(2026, time.March, 29, 9, 20, 0, 0, time.UTC),
	)
	if !ok || state.Profile != "isolated" || state.Status != "reconnecting" {
		t.Fatalf("expected manager sync to return reconnecting state, got state=%#v ok=%v", state, ok)
	}
	if _, ok := sessionRegistry.CurrentTargetForRoute(sessionID, route); ok {
		t.Fatalf("expected manager sync to invalidate managed current target selection")
	}
}

func TestSharedSessionBrowserObserverManagerSyncProfileStatusEventInvalidatesBoundWatchLoopCache(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	manager := NewSharedSessionBrowserObserverManager(sessionRegistry, nil, stateRegistry, 54*time.Second)
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
	sessionID := "sess-manager-sync-cache"
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

	first := bound.ObserveWatchLoop(context.Background(), SharedSessionBrowserObserverRequest{
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
	})
	if len(first.View.Session.Routes) != 1 || first.View.Session.Routes[0].CurrentTargetID == "" {
		t.Fatalf("expected initial watch loop to expose tracked current target, got %#v", first.View.Session.Routes)
	}

	state, ok := manager.SyncProfileStatusEvent(
		sessionID,
		BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		BrowserProfileStatusResult{
			Backend:   "proxy",
			Profile:   "isolated",
			Status:    "reconnecting",
			Running:   true,
			Connected: false,
		},
		time.Date(2026, time.March, 29, 9, 25, 0, 0, time.UTC),
	)
	if !ok || state.Status != "reconnecting" {
		t.Fatalf("expected sync status event to return reconnecting state, got state=%#v ok=%v", state, ok)
	}

	second := bound.ObserveWatchLoop(context.Background(), SharedSessionBrowserObserverRequest{
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
	})
	if len(second.View.Session.Routes) != 1 {
		t.Fatalf("expected invalidated watch loop to preserve route snapshot, got %#v", second.View.Session.Routes)
	}
	if second.View.Session.Routes[0].CurrentTargetID != "" {
		t.Fatalf("expected sync status event to invalidate bound watch-loop cache, got %#v", second.View.Session.Routes)
	}
}

func TestSharedSessionBrowserObserverManagerApplyExecutionResultRefreshesBoundWatchLoopCacheFromExecutionSource(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	manager := NewSharedSessionBrowserObserverManager(sessionRegistry, nil, stateRegistry, time.Minute)
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
	sessionID := "sess-manager-apply-cache"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node"}

	sessionRegistry.TrackTab(sessionID, BrowserSessionTarget{
		TabIndex:   1,
		URL:        "https://example.com",
		Title:      "Example",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
	}, true)

	first := bound.ObserveWatchLoop(context.Background(), SharedSessionBrowserObserverRequest{
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
	})
	if len(first.View.Session.Routes) != 1 || first.View.Session.Routes[0].CurrentTargetID == "" {
		t.Fatalf("expected initial watch loop to expose tracked current target, got %#v", first.View.Session.Routes)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected initial watch loop to poll backend once, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}

	application := manager.ApplyExecutionResult(
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
	)
	if application.Cleanup.ClearedSessionTargets != 1 {
		t.Fatalf("expected execution application cleanup to clear tracked route target, got %#v", application.Cleanup)
	}

	second := bound.ObserveWatchLoop(context.Background(), SharedSessionBrowserObserverRequest{
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
	})
	if len(second.View.Session.Routes) > 1 {
		t.Fatalf("expected refreshed watch loop to preserve a single route snapshot, got %#v", second.View.Session.Routes)
	}
	if len(second.View.Session.Routes) == 1 && second.View.Session.Routes[0].CurrentTargetID != "" {
		t.Fatalf("expected execution application to refresh bound watch-loop cache with cleared target, got %#v", second.View.Session.Routes)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected execution application refresh to avoid extra polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}

func TestSharedSessionBrowserObserverManagerApplyExecutionResultSeedsSiblingProviderFromPrimaryCachedEventCycle(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	runRegistryA := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-manager-apply-sibling-provider": {{RunID: "run-a", Status: "running"}},
		}},
	}
	runRegistryB := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-manager-apply-sibling-provider": {{RunID: "run-b", Status: "running"}},
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
	sessionID := "sess-manager-apply-sibling-provider"
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

	application := managerA.ApplyExecutionResult(
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
	)
	if application.Cleanup.ClearedSessionTargets != 1 {
		t.Fatalf("expected execution application cleanup to clear tracked route target, got %#v", application.Cleanup)
	}

	seededB := boundB.ObserveWatchLoop(context.Background(), req)
	if len(seededB.View.Session.Routes) > 1 {
		t.Fatalf("expected sibling watch loop to preserve a single refreshed route snapshot, got %#v", seededB.View.Session.Routes)
	}
	if len(seededB.View.Session.Routes) == 1 && seededB.View.Session.Routes[0].CurrentTargetID != "" {
		t.Fatalf("expected sibling watch loop to reuse execution-seeded cleared target, got %#v", seededB.View.Session.Routes)
	}
	if len(backend.statusReqs) != 2 || len(backend.profilesReqs) != 2 {
		t.Fatalf("expected sibling watch loop to reuse execution-seeded source without extra polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}

func TestSharedSessionBrowserObserverManagerApplyExecutionResultInvalidatesCurrentTargetFromLifecycleStateWithoutExplicitCleanup(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	manager := NewSharedSessionBrowserObserverManager(sessionRegistry, nil, stateRegistry, time.Minute)
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
	sessionID := "sess-manager-apply-lifecycle-invalidation"
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

	first := bound.ObserveWatchLoop(context.Background(), req)
	if len(first.View.Session.Routes) != 1 || first.View.Session.Routes[0].CurrentTargetID != tracked.ID {
		t.Fatalf("expected initial watch loop to expose tracked current target, got %#v", first.View.Session.Routes)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected initial watch loop to poll backend once, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}

	application := manager.ApplyExecutionResult(
		sessionID,
		BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
		"",
		SharedSessionBrowserExecutionResult{
			ProfileStatus: BrowserProfileStatusResult{
				Backend:   "proxy",
				Profile:   "isolated",
				Status:    "disconnected",
				Running:   true,
				Connected: false,
			},
			Decision: "status_failed",
		},
	)
	if application.Cleanup.ClearedSessionTargets != 0 {
		t.Fatalf("expected lifecycle invalidation path to avoid explicit route cleanup, got %#v", application.Cleanup)
	}
	if selection := CurrentSharedSessionBrowserTargetSelection(sessionRegistry, sessionID, route); selection != nil {
		t.Fatalf("expected execution lifecycle invalidation to clear tracked current target, got %#v", selection)
	}

	second := bound.ObserveWatchLoop(context.Background(), req)
	if len(second.View.Session.Routes) > 1 {
		t.Fatalf("expected refreshed watch loop to preserve a single route snapshot, got %#v", second.View.Session.Routes)
	}
	if len(second.View.Session.Routes) == 1 && second.View.Session.Routes[0].CurrentTargetID != "" {
		t.Fatalf("expected lifecycle invalidation to clear current target without explicit cleanup, got %#v", second.View.Session.Routes)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected lifecycle invalidation refresh to avoid extra polling, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}
