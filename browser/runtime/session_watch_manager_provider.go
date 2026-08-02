package browserruntime

import "time"

// Invalidate clears the short-lived source-time caches held by all bound watch
// managers that share this provider. Callers should invoke it after registry or
// route-selection writeback performed outside the manager-owned event/execution
// contracts.
func (m SharedSessionBrowserObserverManager) Invalidate() {
	m.touchProvider()
	m.invalidateBoundManagers()
}

func (m SharedSessionBrowserObserverManager) invalidateStandaloneBoundManagers() {
	if m.provider != nil {
		return
	}
	m.refreshBoundManagersProjectionCaches()
}

// Bind attaches a concrete backend control to the shared observer/watchdog
// manager so callers can reuse one provider across multiple source-time watch
// entrypoints without rebuilding the registry/run/reconnect dependencies.
func (m SharedSessionBrowserObserverManager) Bind(control BrowserRuntimeControlBackend) SharedSessionBrowserWatchManager {
	m.touchProvider()
	manager := SharedSessionBrowserWatchManager{
		Observer: m,
		Control:  control,
		state: &sharedSessionBrowserWatchManagerState{
			generation:          m.currentGeneration(),
			rawGeneration:       m.currentRawGeneration(),
			lastActiveAt:        time.Now(),
			rawStatus:           make(map[string]sharedSessionBrowserCachedRawStatusObservation),
			rawProfiles:         make(map[string]sharedSessionBrowserCachedRawProfilesObservation),
			rawRouteMutations:   make(map[sharedSessionBrowserRouteMutationSourceKey]sharedSessionBrowserCachedRawRouteMutationObservation),
			routeMutations:      make(map[sharedSessionBrowserRouteMutationSourceKey]sharedSessionBrowserCachedEventCycleObservation),
			rawStatusInFlight:   make(map[string]*sharedSessionBrowserInFlightRawStatusObservation),
			rawProfilesInFlight: make(map[string]*sharedSessionBrowserInFlightRawProfilesObservation),
			rawStartsInFlight:   make(map[string]*sharedSessionBrowserInFlightRawLifecycleObservation),
			rawStopsInFlight:    make(map[string]*sharedSessionBrowserInFlightRawLifecycleObservation),
			eventCycles:         make(map[SharedSessionBrowserObserverRequest]sharedSessionBrowserCachedEventCycleObservation),
			bindings:            make(map[SharedSessionBrowserObserverRequest]sharedSessionBrowserCachedBindingObservation),
			views:               make(map[SharedSessionBrowserObserverRequest]sharedSessionBrowserCachedViewObservation),
			watchLoops:          make(map[SharedSessionBrowserObserverRequest]sharedSessionBrowserCachedWatchLoopObservation),
			eventCyclesInFlight: make(map[SharedSessionBrowserObserverRequest]*sharedSessionBrowserInFlightEventCycleObservation),
			bindingsInFlight:    make(map[SharedSessionBrowserObserverRequest]*sharedSessionBrowserInFlightBindingObservation),
			viewsInFlight:       make(map[SharedSessionBrowserObserverRequest]*sharedSessionBrowserInFlightViewObservation),
			watchLoopsInFlight:  make(map[SharedSessionBrowserObserverRequest]*sharedSessionBrowserInFlightWatchLoopObservation),
		},
	}
	key, ok := sharedSessionBrowserWatchManagerCacheKeyForControl(control)
	if !ok || m.cache == nil {
		manager.touch()
		return manager
	}
	m.pruneIdleBoundManagers(time.Now())
	m.cache.mu.RLock()
	cached, found := m.cache.managers[key]
	m.cache.mu.RUnlock()
	if found {
		cached.touch()
		return cached
	}
	m.cache.mu.Lock()
	defer m.cache.mu.Unlock()
	if cached, found = m.cache.managers[key]; found {
		cached.touch()
		return cached
	}
	m.cache.managers[key] = manager
	manager.touch()
	return manager
}

func (m SharedSessionBrowserObserverManager) invalidateBoundManagers() {
	m.touchProvider()
	if m.cache == nil {
		return
	}
	m.cache.rawGen.Add(1)
	m.cache.generation.Add(1)
}

func (m SharedSessionBrowserObserverManager) invalidateBoundProjectionManagers() {
	m.touchProvider()
	if m.cache == nil {
		return
	}
	m.cache.generation.Add(1)
}

func (m SharedSessionBrowserObserverManager) touchProvider() {
	if m.provider == nil {
		return
	}
	m.provider.touch(time.Now())
}

func (m SharedSessionBrowserObserverManager) pruneIdleBoundManagers(now time.Time) {
	if m.cache == nil {
		return
	}
	m.cache.mu.Lock()
	defer m.cache.mu.Unlock()
	for key, manager := range m.cache.managers {
		if manager.state == nil {
			delete(m.cache.managers, key)
			continue
		}
		manager.state.mu.RLock()
		lastActiveAt := manager.state.lastActiveAt
		manager.state.mu.RUnlock()
		if lastActiveAt.IsZero() || now.Sub(lastActiveAt) > sharedSessionBrowserWatchManagerIdleTTL {
			delete(m.cache.managers, key)
		}
	}
}

func (m SharedSessionBrowserObserverManager) currentGeneration() uint64 {
	if m.cache == nil {
		return 0
	}
	return m.cache.generation.Load()
}

func (m SharedSessionBrowserObserverManager) currentRawGeneration() uint64 {
	if m.cache == nil {
		return 0
	}
	return m.cache.rawGen.Load()
}

func (e *sharedSessionBrowserObserverManagerProviderEntry) touch(now time.Time) {
	if e == nil {
		return
	}
	e.mu.Lock()
	e.lastActiveAt = now
	e.mu.Unlock()
}

func pruneIdleSharedSessionBrowserObserverManagerProvidersLocked(now time.Time) {
	for key, entry := range sharedSessionBrowserObserverManagerProviders.providers {
		if entry == nil {
			delete(sharedSessionBrowserObserverManagerProviders.providers, key)
			continue
		}
		entry.mu.RLock()
		lastActiveAt := entry.lastActiveAt
		entry.mu.RUnlock()
		if lastActiveAt.IsZero() || now.Sub(lastActiveAt) > sharedSessionBrowserObserverManagerIdleTTL {
			delete(sharedSessionBrowserObserverManagerProviders.providers, key)
		}
	}
}

func invalidateSharedSessionBrowserObserverManagersForSessionRegistry(registry *BrowserSessionRegistry) {
	managers := sharedSessionBrowserObserverManagersForSessionRegistry(registry)
	for _, manager := range managers {
		manager.refreshBoundManagersProjectionCaches()
	}
}

func invalidateSharedSessionBrowserObserverManagersForStateRegistry(registry *BrowserSessionStateRegistry) {
	managers := sharedSessionBrowserObserverManagersForStateRegistry(registry)
	for _, manager := range managers {
		manager.refreshBoundManagersProjectionCaches()
	}
}

func sharedSessionBrowserObserverManagersForSessionRegistry(registry *BrowserSessionRegistry) []SharedSessionBrowserObserverManager {
	if registry == nil {
		return nil
	}
	registryKey, ok := sharedSessionBrowserWatchManagerCacheKeyForValue(registry)
	if !ok {
		return nil
	}
	return sharedSessionBrowserObserverManagersMatching(func(
		key sharedSessionBrowserObserverManagerProviderCacheKey,
		entry *sharedSessionBrowserObserverManagerProviderEntry,
	) bool {
		return entry != nil && key.sessionRegistry == registryKey
	})
}

func sharedSessionBrowserObserverManagersForStateRegistry(registry SharedSessionBrowserStateRegistry) []SharedSessionBrowserObserverManager {
	if registry == nil {
		return nil
	}
	registryKey, ok := sharedSessionBrowserWatchManagerCacheKeyForValue(registry)
	if !ok {
		return nil
	}
	return sharedSessionBrowserObserverManagersMatching(func(
		key sharedSessionBrowserObserverManagerProviderCacheKey,
		entry *sharedSessionBrowserObserverManagerProviderEntry,
	) bool {
		return entry != nil && key.stateRegistry == registryKey
	})
}

func sharedSessionBrowserObserverManagersMatching(
	match func(sharedSessionBrowserObserverManagerProviderCacheKey, *sharedSessionBrowserObserverManagerProviderEntry) bool,
) []SharedSessionBrowserObserverManager {
	if match == nil {
		return nil
	}
	sharedSessionBrowserObserverManagerProviders.mu.RLock()
	defer sharedSessionBrowserObserverManagerProviders.mu.RUnlock()
	managers := make([]SharedSessionBrowserObserverManager, 0, len(sharedSessionBrowserObserverManagerProviders.providers))
	for key, entry := range sharedSessionBrowserObserverManagerProviders.providers {
		if !match(key, entry) {
			continue
		}
		managers = append(managers, entry.manager)
	}
	return managers
}

func sharedSessionBrowserObserverManagersShareProvider(
	left SharedSessionBrowserObserverManager,
	right SharedSessionBrowserObserverManager,
) bool {
	if left.provider != nil && right.provider != nil {
		return left.provider == right.provider
	}
	if left.cache != nil && right.cache != nil {
		return left.cache == right.cache
	}
	return false
}

func seedRelatedSharedSessionBrowserObserverManagersRawObservations(
	primary SharedSessionBrowserObserverManager,
	sessionRegistry *BrowserSessionRegistry,
	stateRegistry SharedSessionBrowserStateRegistry,
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	requestedProfiles []string,
	status *SharedSessionBrowserRawStatusObservation,
	profiles *SharedSessionBrowserRawProfilesObservation,
) {
	var managers []SharedSessionBrowserObserverManager
	if stateRegistry != nil {
		managers = sharedSessionBrowserObserverManagersForStateRegistry(stateRegistry)
	} else {
		managers = sharedSessionBrowserObserverManagersForSessionRegistry(sessionRegistry)
	}
	for _, manager := range managers {
		if sharedSessionBrowserObserverManagersShareProvider(manager, primary) {
			continue
		}
		manager.seedBoundManagersRawObservations(
			sessionID,
			selectedInfo,
			requestedProfiles,
			status,
			profiles,
		)
	}
}

func seedRelatedSharedSessionBrowserObserverManagersRawRouteMutation(
	primary SharedSessionBrowserObserverManager,
	sessionRegistry *BrowserSessionRegistry,
	stateRegistry SharedSessionBrowserStateRegistry,
	sessionID string,
	route BrowserSessionRoute,
	observation SharedSessionBrowserRawRouteMutationObservation,
) {
	var managers []SharedSessionBrowserObserverManager
	if stateRegistry != nil {
		managers = sharedSessionBrowserObserverManagersForStateRegistry(stateRegistry)
	} else {
		managers = sharedSessionBrowserObserverManagersForSessionRegistry(sessionRegistry)
	}
	for _, manager := range managers {
		if sharedSessionBrowserObserverManagersShareProvider(manager, primary) {
			continue
		}
		manager.seedBoundManagersRawRouteMutationSource(sessionID, route, observation)
	}
}
