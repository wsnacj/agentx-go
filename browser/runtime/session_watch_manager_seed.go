package browserruntime

import (
	"strings"
	"time"
)

type sharedSessionBrowserProjectionSeedPlan struct {
	eventCycles map[SharedSessionBrowserObserverRequest]struct{}
	bindings    map[SharedSessionBrowserObserverRequest]struct{}
	views       map[SharedSessionBrowserObserverRequest]struct{}
	watchLoops  map[SharedSessionBrowserObserverRequest]struct{}
}

func (m SharedSessionBrowserObserverManager) seedBoundManagersForRawStatus(
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	requestedProfiles []string,
	observation SharedSessionBrowserRawStatusObservation,
) {
	if !sharedSessionBrowserRawStatusObservationProvided(observation) {
		return
	}
	m.seedBoundManagersRawObservations(sessionID, selectedInfo, requestedProfiles, &observation, nil)
}

func (m SharedSessionBrowserObserverManager) seedBoundManagersForRawProfiles(
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	requestedProfiles []string,
	observation SharedSessionBrowserRawProfilesObservation,
) {
	if !sharedSessionBrowserRawProfilesObservationProvided(observation) {
		return
	}
	m.seedBoundManagersRawObservations(sessionID, selectedInfo, requestedProfiles, nil, &observation)
}

func (m SharedSessionBrowserObserverManager) seedBoundManagersRouteMutationSource(
	sessionID string,
	route BrowserSessionRoute,
) {
	if m.cache == nil {
		return
	}
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return
	}
	route = normalizeBrowserSessionRoute(route)
	selectedInfo := BrowserRuntimeInfo{
		Backend: strings.TrimSpace(route.Backend),
		Profile: strings.TrimSpace(route.Profile),
		Target:  strings.TrimSpace(route.Target),
	}
	now := time.Now()
	m.pruneIdleBoundManagers(now)

	m.cache.mu.RLock()
	managers := make([]SharedSessionBrowserWatchManager, 0, len(m.cache.managers))
	for _, manager := range m.cache.managers {
		managers = append(managers, manager)
	}
	m.cache.mu.RUnlock()

	for _, manager := range managers {
		if !sharedSessionBrowserWatchManagerMatchesRuntimeInfo(manager, selectedInfo) {
			continue
		}
		observation, ok := m.syntheticRouteMutationSourceObservation(
			sessionID,
			selectedInfo,
			route,
			route.Profile,
		)
		if !ok {
			observation, ok = sharedSessionBrowserCachedRouteMutationSourceObservationFromManager(
				manager,
				sessionID,
				selectedInfo,
				route.Profile,
			)
		}
		if !ok {
			continue
		}
		manager.storeRouteMutationSource(sessionID, selectedInfo, route.Profile, observation)
	}
}

func (m SharedSessionBrowserObserverManager) seedBoundManagersRawRouteMutationSource(
	sessionID string,
	route BrowserSessionRoute,
	observation SharedSessionBrowserRawRouteMutationObservation,
) {
	if m.cache == nil || !sharedSessionBrowserRawRouteMutationObservationProvided(observation) {
		return
	}
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return
	}
	route = normalizeBrowserSessionRoute(route)
	selectedInfo := BrowserRuntimeInfo{
		Backend: strings.TrimSpace(route.Backend),
		Profile: strings.TrimSpace(route.Profile),
		Target:  strings.TrimSpace(route.Target),
	}
	observation = normalizeSharedSessionBrowserRawRouteMutationObservation(observation, route.Profile)
	now := time.Now()
	m.pruneIdleBoundManagers(now)

	m.cache.mu.RLock()
	managers := make([]SharedSessionBrowserWatchManager, 0, len(m.cache.managers))
	for _, manager := range m.cache.managers {
		managers = append(managers, manager)
	}
	currentRawGeneration := m.cache.rawGen.Load()
	m.cache.mu.RUnlock()

	key := normalizeSharedSessionBrowserRouteMutationSourceKey(sessionID, selectedInfo, route.Profile)
	for _, manager := range managers {
		if !sharedSessionBrowserWatchManagerMatchesRuntimeInfo(manager, selectedInfo) || manager.state == nil {
			continue
		}
		manager.state.mu.Lock()
		if manager.state.rawGeneration != currentRawGeneration {
			clear(manager.state.rawStatus)
			clear(manager.state.rawProfiles)
			clear(manager.state.rawRouteMutations)
			clear(manager.state.routeMutations)
			clear(manager.state.rawStatusInFlight)
			clear(manager.state.rawProfilesInFlight)
			clear(manager.state.rawStartsInFlight)
			clear(manager.state.rawStopsInFlight)
			manager.state.rawGeneration = currentRawGeneration
		}
		manager.state.rawRouteMutations[key] = sharedSessionBrowserCachedRawRouteMutationObservation{
			cachedAt:    now,
			observation: observation,
		}
		manager.state.mu.Unlock()
	}
}

func (m SharedSessionBrowserObserverManager) syntheticRouteMutationSourceObservation(
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	route BrowserSessionRoute,
	requestedProfile string,
) (SharedSessionBrowserEventCycleObservation, bool) {
	status, profiles := sharedSessionBrowserSyntheticRawObservationsForRouteMutation(
		m.StateRegistry,
		sessionID,
		selectedInfo,
		requestedProfile,
	)
	if status == nil && profiles == nil {
		return SharedSessionBrowserEventCycleObservation{}, false
	}
	req := SharedSessionBrowserObserverRequest{
		SessionID: strings.TrimSpace(sessionID),
		SelectedInfo: BrowserRuntimeInfo{
			Backend: strings.TrimSpace(selectedInfo.Backend),
			Profile: strings.TrimSpace(selectedInfo.Profile),
			Target:  strings.TrimSpace(selectedInfo.Target),
		},
		BindingRoute:     normalizeBrowserSessionRoute(route),
		RequestedProfile: strings.TrimSpace(requestedProfile),
		IncludeStatus:    status != nil,
		IncludeProfiles:  profiles != nil,
	}
	observation := m.seededEventCycleObservationForRequest(req.SessionID, req, status, profiles)
	if !sharedSessionBrowserSeededEventCycleObservationProvided(observation, req) {
		return SharedSessionBrowserEventCycleObservation{}, false
	}
	return observation, true
}

func (m SharedSessionBrowserObserverManager) seedSiblingProvidersForRouteMutation(
	sessionID string,
	route BrowserSessionRoute,
) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" || m.SessionRegistry == nil || m.cache == nil {
		return
	}
	route = normalizeBrowserSessionRoute(route)
	selectedInfo := BrowserRuntimeInfo{
		Backend: route.Backend,
		Profile: route.Profile,
		Target:  route.Target,
	}
	selectedInfo.Backend = strings.TrimSpace(selectedInfo.Backend)
	selectedInfo.Profile = strings.TrimSpace(selectedInfo.Profile)
	selectedInfo.Target = strings.TrimSpace(selectedInfo.Target)
	hasBoundManagers := m.hasBoundManagers()
	m.seedBoundManagersRouteMutationSource(sessionID, route)
	primaryHasRouteMutationSource := m.boundManagersHaveRouteMutationSource(sessionID, selectedInfo, route.Profile)
	status, profiles := m.cachedRawObservationsForRouteMutation(sessionID, selectedInfo, route.Profile)
	if status == nil && profiles == nil {
		if !hasBoundManagers {
			m.invalidateBoundProjectionManagers()
		}
		return
	}
	requestedProfiles := sharedSessionBrowserRawObservationCacheKeys(route.Profile)
	if primaryHasRouteMutationSource {
		m.refreshBoundManagersProjectionCaches()
	} else {
		m.seedBoundManagersRawObservations(
			sessionID,
			selectedInfo,
			requestedProfiles,
			status,
			profiles,
		)
	}
	if !hasBoundManagers {
		m.invalidateBoundProjectionManagers()
	}
	seedRelatedSharedSessionBrowserObserverManagersRawObservations(
		m,
		m.SessionRegistry,
		m.StateRegistry,
		sessionID,
		selectedInfo,
		requestedProfiles,
		status,
		profiles,
	)
}

func (m SharedSessionBrowserObserverManager) seedBoundManagersRawObservations(
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	requestedProfiles []string,
	status *SharedSessionBrowserRawStatusObservation,
	profiles *SharedSessionBrowserRawProfilesObservation,
) {
	if m.cache == nil {
		return
	}
	now := time.Now()
	m.pruneIdleBoundManagers(now)

	m.cache.mu.RLock()
	managers := make([]SharedSessionBrowserWatchManager, 0, len(m.cache.managers))
	for _, manager := range m.cache.managers {
		managers = append(managers, manager)
	}
	currentGeneration := m.cache.generation.Load()
	currentRawGeneration := m.cache.rawGen.Load()
	m.cache.mu.RUnlock()

	keys := sharedSessionBrowserRawObservationCacheKeys(requestedProfiles...)
	if len(keys) == 0 {
		keys = []string{""}
	}
	keySet := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		keySet[key] = struct{}{}
	}

	for _, manager := range managers {
		if !sharedSessionBrowserWatchManagerMatchesRuntimeInfo(manager, selectedInfo) || manager.state == nil {
			continue
		}
		manager.state.mu.Lock()
		plan := sharedSessionBrowserProjectionSeedPlanFromStateLocked(manager.state)
		cachedEventCycles := make(map[SharedSessionBrowserObserverRequest]SharedSessionBrowserEventCycleObservation, len(manager.state.eventCycles)+len(manager.state.watchLoops))
		for key, cached := range manager.state.eventCycles {
			normalized := normalizeSharedSessionBrowserEventCycleRequest(key)
			cachedEventCycles[normalized] = cached.observation
		}
		for key, cached := range manager.state.watchLoops {
			normalized := normalizeSharedSessionBrowserEventCycleRequest(key)
			if _, ok := cachedEventCycles[normalized]; ok {
				continue
			}
			cachedEventCycles[normalized] = cached.observation.Cycle
		}
		eventCycleRequests := sharedSessionBrowserProjectionSeedRequestsLocked(
			manager.state,
			sessionID,
			selectedInfo,
			keySet,
			status != nil,
			profiles != nil,
		)
		resetSharedSessionBrowserWatchManagerObservationCachesLocked(manager.state, currentGeneration, currentRawGeneration)
		seededRawStatus := make(map[string]SharedSessionBrowserRawStatusObservation, len(keys))
		seededRawProfiles := make(map[string]SharedSessionBrowserRawProfilesObservation, len(keys))
		for _, key := range keys {
			if status != nil {
				observation := sharedSessionBrowserRawStatusObservationForRequestedProfile(*status, key)
				manager.state.rawStatus[key] = sharedSessionBrowserCachedRawStatusObservation{
					cachedAt:    now,
					observation: observation,
				}
				seededRawStatus[key] = observation
			}
			if profiles != nil {
				observation := sharedSessionBrowserRawProfilesObservationForRequestedProfile(*profiles, key)
				manager.state.rawProfiles[key] = sharedSessionBrowserCachedRawProfilesObservation{
					cachedAt:    now,
					observation: observation,
				}
				seededRawProfiles[key] = observation
			}
		}
		for _, req := range eventCycleRequests {
			observation, ok := m.seededProjectionEventCycleForRequest(
				req,
				seededRawStatus,
				seededRawProfiles,
				cachedEventCycles,
				nil,
			)
			if !ok {
				continue
			}
			manager.state.eventCycles[req] = sharedSessionBrowserCachedEventCycleObservation{
				cachedAt:    now,
				observation: observation,
			}
		}
		for req := range plan.bindings {
			normalizedCycleReq := normalizeSharedSessionBrowserEventCycleRequest(req)
			cycle, ok := manager.state.eventCycles[normalizedCycleReq]
			if !ok {
				continue
			}
			observation := observeSharedSessionBrowserBindingForScopeFromCycle(
				req,
				cycle.observation,
				m.SessionRegistry,
				m.RunRegistry,
				m.StateRegistry,
				m.ReconnectWindow,
			)
			manager.state.bindings[req] = sharedSessionBrowserCachedBindingObservation{
				cachedAt:    now,
				observation: observation,
			}
		}
		for req := range plan.views {
			bindingReq := normalizeSharedSessionBrowserBindingRequest(req)
			binding, ok := manager.state.bindings[bindingReq]
			if !ok {
				continue
			}
			observation := observeSharedSessionBrowserViewForScopeFromBinding(
				req,
				binding.observation,
				m.SessionRegistry,
				m.RunRegistry,
				m.StateRegistry,
			)
			manager.state.views[req] = sharedSessionBrowserCachedViewObservation{
				cachedAt:    now,
				observation: observation,
			}
		}
		for req := range plan.watchLoops {
			normalizedCycleReq := normalizeSharedSessionBrowserEventCycleRequest(req)
			cycle, ok := manager.state.eventCycles[normalizedCycleReq]
			if !ok {
				continue
			}
			observer := observeSharedSessionBrowserObserverForScopeFromCycle(
				req,
				cycle.observation,
				m.SessionRegistry,
				m.RunRegistry,
				m.StateRegistry,
				m.ReconnectWindow,
			)
			view := SharedSessionBrowserViewObservation{
				Observation: observer.Observation,
				Binding:     observer.Binding,
				Session:     observer.Session,
			}
			watch := SharedSessionBrowserWatchObservation{
				View:               view,
				Profiles:           observer.Profiles,
				DiscoveredProfiles: observer.DiscoveredProfiles,
				DefaultProfile:     observer.DefaultProfile,
				Note:               observer.Note,
				ReferenceTime:      observer.ReferenceTime,
			}
			manager.state.watchLoops[req] = sharedSessionBrowserCachedWatchLoopObservation{
				cachedAt: now,
				observation: SharedSessionBrowserWatchLoopObservation{
					Cycle:         cycle.observation,
					Observer:      observer,
					Watch:         watch,
					View:          view,
					ReferenceTime: observer.ReferenceTime,
				},
			}
		}
		manager.state.lastActiveAt = now
		manager.state.mu.Unlock()
	}
}

func (m SharedSessionBrowserObserverManager) boundManagersHaveRouteMutationSource(
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	requestedProfile string,
) bool {
	if m.cache == nil {
		return false
	}
	now := time.Now()
	m.pruneIdleBoundManagers(now)

	m.cache.mu.RLock()
	managers := make([]SharedSessionBrowserWatchManager, 0, len(m.cache.managers))
	for _, manager := range m.cache.managers {
		managers = append(managers, manager)
	}
	m.cache.mu.RUnlock()

	for _, manager := range managers {
		if !sharedSessionBrowserWatchManagerMatchesRuntimeInfo(manager, selectedInfo) {
			continue
		}
		if _, ok := manager.cachedRouteMutationSource(sessionID, selectedInfo, requestedProfile); ok {
			return true
		}
	}
	return false
}

func (m SharedSessionBrowserObserverManager) hasBoundManagers() bool {
	if m.cache == nil {
		return false
	}
	now := time.Now()
	m.pruneIdleBoundManagers(now)
	m.cache.mu.RLock()
	defer m.cache.mu.RUnlock()
	return len(m.cache.managers) > 0
}

func (m SharedSessionBrowserObserverManager) cachedRawObservationsForRouteMutation(
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	requestedProfile string,
) (*SharedSessionBrowserRawStatusObservation, *SharedSessionBrowserRawProfilesObservation) {
	if m.cache == nil {
		return nil, nil
	}
	now := time.Now()
	m.pruneIdleBoundManagers(now)

	m.cache.mu.RLock()
	managers := make([]SharedSessionBrowserWatchManager, 0, len(m.cache.managers))
	for _, manager := range m.cache.managers {
		managers = append(managers, manager)
	}
	m.cache.mu.RUnlock()

	requestedProfiles := sharedSessionBrowserRawObservationCacheKeys(requestedProfile)
	if len(requestedProfiles) == 0 {
		requestedProfiles = []string{""}
	}
	for _, manager := range managers {
		if !sharedSessionBrowserWatchManagerMatchesRuntimeInfo(manager, selectedInfo) || manager.state == nil {
			continue
		}
		status, profiles := sharedSessionBrowserCachedRouteMutationRawObservationsFromManager(
			manager,
			sessionID,
			selectedInfo,
			requestedProfile,
			requestedProfiles,
		)
		if status != nil || profiles != nil {
			return status, profiles
		}
	}
	return sharedSessionBrowserSyntheticRawObservationsForRouteMutation(
		m.StateRegistry,
		sessionID,
		selectedInfo,
		requestedProfile,
	)
}

func sharedSessionBrowserSyntheticRawObservationsForRouteMutation(
	stateRegistry SharedSessionBrowserStateRegistry,
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	requestedProfile string,
) (*SharedSessionBrowserRawStatusObservation, *SharedSessionBrowserRawProfilesObservation) {
	sessionID = strings.TrimSpace(sessionID)
	requestedProfile = strings.TrimSpace(requestedProfile)
	if stateRegistry == nil || sessionID == "" {
		return nil, nil
	}
	snapshot := stateRegistry.SnapshotSessionBrowserProfilesForScope(sessionID, selectedInfo, requestedProfile)
	if len(snapshot) == 0 {
		return nil, nil
	}

	observedAt := time.Time{}
	profiles := make([]BrowserProfileInfo, 0, len(snapshot))
	defaultProfile := firstNonEmptyString(requestedProfile, strings.TrimSpace(selectedInfo.Profile))
	for _, state := range snapshot {
		if state.ObservedAt.After(observedAt) {
			observedAt = state.ObservedAt
		}
		profiles = append(profiles, BrowserProfileInfo{
			Profile:    strings.TrimSpace(state.Profile),
			BrowserApp: strings.TrimSpace(state.BrowserApp),
			Status:     strings.TrimSpace(state.Status),
			Running:    state.Running,
			Connected:  state.Connected,
			Note:       strings.TrimSpace(state.Note),
		})
		if defaultProfile == "" && strings.TrimSpace(state.Profile) != "" {
			defaultProfile = strings.TrimSpace(state.Profile)
		}
	}
	if observedAt.IsZero() {
		observedAt = time.Now()
	}

	profilesResult := BrowserProfilesResult{
		Backend:        firstNonEmptyString(strings.TrimSpace(snapshot[0].Backend), strings.TrimSpace(selectedInfo.Backend)),
		DefaultProfile: defaultProfile,
		Profiles:       profiles,
	}
	profilesObservation := &SharedSessionBrowserRawProfilesObservation{
		RequestedProfile: requestedProfile,
		Profiles:         &profilesResult,
		ObservedAt:       observedAt,
	}

	var statusObservation *SharedSessionBrowserRawStatusObservation
	statusProfile := firstNonEmptyString(requestedProfile, strings.TrimSpace(selectedInfo.Profile), defaultProfile)
	if statusProfile != "" {
		fallbackStatus := BrowserProfileStatusResult{}
		for _, state := range snapshot {
			if !strings.EqualFold(strings.TrimSpace(state.Profile), statusProfile) {
				continue
			}
			fallbackStatus = SharedSessionBrowserProfileStatusResultFromState(state, selectedInfo, statusProfile)
			if !state.ObservedAt.IsZero() {
				observedAt = state.ObservedAt
			}
			break
		}
		if resolvedStatus, ok := stateRegistry.ResolveSessionBrowserProfileStatus(sessionID, selectedInfo, statusProfile, fallbackStatus); ok {
			status := resolvedStatus
			statusObservation = &SharedSessionBrowserRawStatusObservation{
				RequestedProfile: requestedProfile,
				Status:           &status,
				ObservedAt:       observedAt,
			}
		}
	}

	return statusObservation, profilesObservation
}

func sharedSessionBrowserCachedRouteMutationSourceObservationFromManager(
	manager SharedSessionBrowserWatchManager,
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	requestedProfile string,
) (SharedSessionBrowserEventCycleObservation, bool) {
	if manager.state == nil {
		return SharedSessionBrowserEventCycleObservation{}, false
	}
	if observation, ok := manager.cachedRouteMutationSource(sessionID, selectedInfo, requestedProfile); ok {
		return observation, true
	}

	now := time.Now()
	manager.state.mu.RLock()
	cachedEventCycles := make(map[SharedSessionBrowserObserverRequest]SharedSessionBrowserEventCycleObservation, len(manager.state.eventCycles)+len(manager.state.watchLoops))
	for key, cached := range manager.state.eventCycles {
		if now.Sub(cached.cachedAt) > sharedSessionBrowserWatchManagerEventCycleCacheTTL {
			continue
		}
		normalized := normalizeSharedSessionBrowserEventCycleRequest(key)
		cachedEventCycles[normalized] = cached.observation
	}
	for key, cached := range manager.state.watchLoops {
		if now.Sub(cached.cachedAt) > sharedSessionBrowserWatchManagerEventCycleCacheTTL {
			continue
		}
		normalized := normalizeSharedSessionBrowserEventCycleRequest(key)
		if _, ok := cachedEventCycles[normalized]; ok {
			continue
		}
		cachedEventCycles[normalized] = cached.observation.Cycle
	}
	manager.state.mu.RUnlock()

	return sharedSessionBrowserCachedRouteMutationEventCycle(sessionID, selectedInfo, requestedProfile, cachedEventCycles)
}

func sharedSessionBrowserCachedRouteMutationRawObservationsFromManager(
	manager SharedSessionBrowserWatchManager,
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	requestedProfile string,
	requestedProfiles []string,
) (*SharedSessionBrowserRawStatusObservation, *SharedSessionBrowserRawProfilesObservation) {
	if manager.state == nil {
		return nil, nil
	}
	routeMutationSource, routeMutationSourceOK := sharedSessionBrowserCachedRouteMutationSourceObservationFromManager(
		manager,
		sessionID,
		selectedInfo,
		requestedProfile,
	)

	var status *SharedSessionBrowserRawStatusObservation
	if routeMutationSourceOK {
		observation := sharedSessionBrowserRawStatusObservationForRequestedProfile(
			sharedSessionBrowserRawStatusObservationFromEventCycle(routeMutationSource, requestedProfile),
			requestedProfile,
		)
		if sharedSessionBrowserRawStatusObservationProvided(observation) {
			status = &observation
		}
	}

	var profiles *SharedSessionBrowserRawProfilesObservation
	if routeMutationSourceOK {
		observation := sharedSessionBrowserRawProfilesObservationForRequestedProfile(
			sharedSessionBrowserRawProfilesObservationFromEventCycle(routeMutationSource, requestedProfile),
			requestedProfile,
		)
		if sharedSessionBrowserRawProfilesObservationProvided(observation) {
			profiles = &observation
		}
	}

	manager.state.mu.RLock()
	defer manager.state.mu.RUnlock()

	if status == nil {
		for _, key := range requestedProfiles {
			cached, ok := manager.state.rawStatus[strings.TrimSpace(key)]
			if !ok || !sharedSessionBrowserRawStatusObservationProvided(cached.observation) {
				continue
			}
			observation := sharedSessionBrowserRawStatusObservationForRequestedProfile(cached.observation, requestedProfile)
			if !sharedSessionBrowserRawStatusObservationProvided(observation) {
				continue
			}
			status = &observation
			break
		}
	}

	if profiles == nil {
		for _, key := range requestedProfiles {
			cached, ok := manager.state.rawProfiles[strings.TrimSpace(key)]
			if !ok || !sharedSessionBrowserRawProfilesObservationProvided(cached.observation) {
				continue
			}
			observation := sharedSessionBrowserRawProfilesObservationForRequestedProfile(cached.observation, requestedProfile)
			if !sharedSessionBrowserRawProfilesObservationProvided(observation) {
				continue
			}
			profiles = &observation
			break
		}
	}

	if status != nil && profiles != nil {
		return status, profiles
	}

	cachedEventCycles := make(map[SharedSessionBrowserObserverRequest]SharedSessionBrowserEventCycleObservation, len(manager.state.eventCycles)+len(manager.state.watchLoops))
	for key, cached := range manager.state.eventCycles {
		normalized := normalizeSharedSessionBrowserEventCycleRequest(key)
		cachedEventCycles[normalized] = cached.observation
	}
	for key, cached := range manager.state.watchLoops {
		normalized := normalizeSharedSessionBrowserEventCycleRequest(key)
		if _, ok := cachedEventCycles[normalized]; ok {
			continue
		}
		cachedEventCycles[normalized] = cached.observation.Cycle
	}
	fallbackCycle, ok := sharedSessionBrowserCachedRouteMutationEventCycle(sessionID, selectedInfo, requestedProfile, cachedEventCycles)
	if !ok {
		return status, profiles
	}
	if status == nil {
		observation := sharedSessionBrowserRawStatusObservationForRequestedProfile(
			sharedSessionBrowserRawStatusObservationFromEventCycle(fallbackCycle, requestedProfile),
			requestedProfile,
		)
		if sharedSessionBrowserRawStatusObservationProvided(observation) {
			status = &observation
		}
	}
	if profiles == nil {
		observation := sharedSessionBrowserRawProfilesObservationForRequestedProfile(
			sharedSessionBrowserRawProfilesObservationFromEventCycle(fallbackCycle, requestedProfile),
			requestedProfile,
		)
		if sharedSessionBrowserRawProfilesObservationProvided(observation) {
			profiles = &observation
		}
	}
	return status, profiles
}

func (m SharedSessionBrowserObserverManager) refreshBoundManagersProjectionCaches() {
	m.touchProvider()
	if m.cache == nil {
		return
	}
	nextGeneration := m.cache.generation.Add(1)
	now := time.Now()
	m.pruneIdleBoundManagers(now)

	m.cache.mu.RLock()
	managers := make([]SharedSessionBrowserWatchManager, 0, len(m.cache.managers))
	for _, manager := range m.cache.managers {
		managers = append(managers, manager)
	}
	m.cache.mu.RUnlock()

	for _, manager := range managers {
		m.refreshBoundManagerProjectionCaches(now, nextGeneration, manager)
	}
}

func (m SharedSessionBrowserObserverManager) refreshBoundManagerProjectionCaches(
	now time.Time,
	nextGeneration uint64,
	manager SharedSessionBrowserWatchManager,
) {
	if manager.state == nil {
		return
	}

	manager.state.mu.Lock()
	rawStatus := make(map[string]SharedSessionBrowserRawStatusObservation, len(manager.state.rawStatus))
	for key, cached := range manager.state.rawStatus {
		rawStatus[key] = cached.observation
	}
	rawProfiles := make(map[string]SharedSessionBrowserRawProfilesObservation, len(manager.state.rawProfiles))
	for key, cached := range manager.state.rawProfiles {
		rawProfiles[key] = cached.observation
	}
	cachedEventCycles := make(map[SharedSessionBrowserObserverRequest]SharedSessionBrowserEventCycleObservation, len(manager.state.eventCycles)+len(manager.state.watchLoops))
	routeMutationCycles := make(map[sharedSessionBrowserRouteMutationSourceKey]SharedSessionBrowserEventCycleObservation, len(manager.state.routeMutations))
	for key, cached := range manager.state.eventCycles {
		normalized := normalizeSharedSessionBrowserEventCycleRequest(key)
		cachedEventCycles[normalized] = cached.observation
	}
	for key, cached := range manager.state.watchLoops {
		normalized := normalizeSharedSessionBrowserEventCycleRequest(key)
		if _, ok := cachedEventCycles[normalized]; ok {
			continue
		}
		cachedEventCycles[normalized] = cached.observation.Cycle
	}
	for key, cached := range manager.state.routeMutations {
		if now.Sub(cached.cachedAt) > sharedSessionBrowserWatchManagerEventCycleCacheTTL {
			continue
		}
		routeMutationCycles[key] = cached.observation
	}
	plan := sharedSessionBrowserProjectionSeedPlanFromStateLocked(manager.state)
	clear(manager.state.eventCycles)
	clear(manager.state.bindings)
	clear(manager.state.views)
	clear(manager.state.watchLoops)
	clear(manager.state.eventCyclesInFlight)
	clear(manager.state.bindingsInFlight)
	clear(manager.state.viewsInFlight)
	clear(manager.state.watchLoopsInFlight)
	manager.state.generation = nextGeneration
	manager.state.lastActiveAt = now
	manager.state.mu.Unlock()

	eventCycles := make(map[SharedSessionBrowserObserverRequest]SharedSessionBrowserEventCycleObservation, len(plan.eventCycles))
	for req := range plan.eventCycles {
		observation, ok := m.seededProjectionEventCycleForRequest(req, rawStatus, rawProfiles, cachedEventCycles, routeMutationCycles)
		if !ok {
			continue
		}
		eventCycles[req] = observation
	}

	manager.state.mu.Lock()
	defer manager.state.mu.Unlock()
	for req, observation := range eventCycles {
		manager.state.eventCycles[req] = sharedSessionBrowserCachedEventCycleObservation{
			cachedAt:    now,
			observation: observation,
		}
	}
	for req := range plan.bindings {
		cycle, ok := eventCycles[normalizeSharedSessionBrowserEventCycleRequest(req)]
		if !ok {
			continue
		}
		observation := observeSharedSessionBrowserBindingForScopeFromCycle(
			req,
			cycle,
			m.SessionRegistry,
			m.RunRegistry,
			m.StateRegistry,
			m.ReconnectWindow,
		)
		manager.state.bindings[req] = sharedSessionBrowserCachedBindingObservation{
			cachedAt:    now,
			observation: observation,
		}
	}
	for req := range plan.views {
		binding, ok := manager.state.bindings[normalizeSharedSessionBrowserBindingRequest(req)]
		if !ok {
			continue
		}
		observation := observeSharedSessionBrowserViewForScopeFromBinding(
			req,
			binding.observation,
			m.SessionRegistry,
			m.RunRegistry,
			m.StateRegistry,
		)
		manager.state.views[req] = sharedSessionBrowserCachedViewObservation{
			cachedAt:    now,
			observation: observation,
		}
	}
	for req := range plan.watchLoops {
		cycle, ok := eventCycles[normalizeSharedSessionBrowserEventCycleRequest(req)]
		if !ok {
			continue
		}
		observer := observeSharedSessionBrowserObserverForScopeFromCycle(
			req,
			cycle,
			m.SessionRegistry,
			m.RunRegistry,
			m.StateRegistry,
			m.ReconnectWindow,
		)
		view := SharedSessionBrowserViewObservation{
			Observation: observer.Observation,
			Binding:     observer.Binding,
			Session:     observer.Session,
		}
		watch := SharedSessionBrowserWatchObservation{
			View:               view,
			Profiles:           observer.Profiles,
			DiscoveredProfiles: observer.DiscoveredProfiles,
			DefaultProfile:     observer.DefaultProfile,
			Note:               observer.Note,
			ReferenceTime:      observer.ReferenceTime,
		}
		manager.state.watchLoops[req] = sharedSessionBrowserCachedWatchLoopObservation{
			cachedAt: now,
			observation: SharedSessionBrowserWatchLoopObservation{
				Cycle:         cycle,
				Observer:      observer,
				Watch:         watch,
				View:          view,
				ReferenceTime: observer.ReferenceTime,
			},
		}
	}
}

func (m SharedSessionBrowserObserverManager) seededEventCycleObservationForRequest(
	sessionID string,
	req SharedSessionBrowserObserverRequest,
	status *SharedSessionBrowserRawStatusObservation,
	profiles *SharedSessionBrowserRawProfilesObservation,
) SharedSessionBrowserEventCycleObservation {
	req = normalizeSharedSessionBrowserEventCycleRequest(req)
	observation := SharedSessionBrowserEventCycleObservation{}
	requestedProfile := firstNonEmptyBindingString(
		strings.TrimSpace(req.RequestedProfile),
		strings.TrimSpace(req.SelectedInfo.Profile),
		strings.TrimSpace(req.BindingRoute.Profile),
	)
	if req.IncludeStatus && status != nil {
		statusObservation := sharedSessionBrowserRawStatusObservationForRequestedProfile(*status, req.RequestedProfile)
		observation.Observation.Status = statusObservation.Status
		observation.Observation.StatusErr = statusObservation.Err
		observation.Observation.StatusObservedAt = statusObservation.ObservedAt
		if statusObservation.Status != nil {
			observation.Observation.ResolvedStatus = *statusObservation.Status
			if m.StateRegistry != nil && strings.TrimSpace(sessionID) != "" {
				if resolved, ok := m.StateRegistry.ResolveSessionBrowserProfileStatus(sessionID, req.SelectedInfo, requestedProfile, *statusObservation.Status); ok {
					observation.Observation.ResolvedStatus = resolved
				}
				desired := SharedSessionBrowserProfileStateFromStatus(req.SelectedInfo, *statusObservation.Status)
				desired.Profile = firstNonEmptyString(strings.TrimSpace(desired.Profile), requestedProfile)
				if synced, ok := m.StateRegistry.ResolveSessionBrowserProfileState(sessionID, desired); ok {
					observation.Observation.SyncedState = synced
					observation.Observation.HasSyncedState = true
				}
			}
		}
	}
	if req.IncludeProfiles && profiles != nil {
		profilesObservation := sharedSessionBrowserRawProfilesObservationForRequestedProfile(*profiles, req.RequestedProfile)
		observation.Observation.Profiles = profilesObservation.Profiles
		observation.Observation.ProfilesErr = profilesObservation.Err
		observation.Observation.ProfilesObservedAt = profilesObservation.ObservedAt
		if profilesObservation.Profiles != nil {
			observation.Observation.Snapshot = SharedSessionBrowserProfileStatesFromObservedProfiles(
				req.SelectedInfo,
				*profilesObservation.Profiles,
				profilesObservation.ObservedAt,
			)
		}
		if m.StateRegistry != nil && strings.TrimSpace(sessionID) != "" {
			if snapshot := m.StateRegistry.SnapshotSessionBrowserProfilesForScope(sessionID, req.SelectedInfo, requestedProfile); len(snapshot) > 0 {
				observation.Observation.Snapshot = snapshot
			}
		}
	}
	observation.ReferenceTime = sharedSessionBrowserLatestEventCycleObservedAt(
		observation.Observation.StatusObservedAt,
		observation.Observation.ProfilesObservedAt,
	)
	return observation
}

func (m SharedSessionBrowserObserverManager) seededProjectionEventCycleForRequest(
	req SharedSessionBrowserObserverRequest,
	rawStatus map[string]SharedSessionBrowserRawStatusObservation,
	rawProfiles map[string]SharedSessionBrowserRawProfilesObservation,
	cachedEventCycles map[SharedSessionBrowserObserverRequest]SharedSessionBrowserEventCycleObservation,
	routeMutationCycles map[sharedSessionBrowserRouteMutationSourceKey]SharedSessionBrowserEventCycleObservation,
) (SharedSessionBrowserEventCycleObservation, bool) {
	req = normalizeSharedSessionBrowserEventCycleRequest(req)
	fallbackCycle, haveFallbackCycle := sharedSessionBrowserProjectionSeedFallbackEventCycle(req, cachedEventCycles, routeMutationCycles)
	var status *SharedSessionBrowserRawStatusObservation
	if req.IncludeStatus {
		observation, ok := rawStatus[strings.TrimSpace(req.RequestedProfile)]
		if ok && sharedSessionBrowserRawStatusObservationProvided(observation) {
			normalized := sharedSessionBrowserRawStatusObservationForRequestedProfile(observation, req.RequestedProfile)
			status = &normalized
		} else {
			if !haveFallbackCycle {
				return SharedSessionBrowserEventCycleObservation{}, false
			}
			normalized := sharedSessionBrowserRawStatusObservationForRequestedProfile(
				sharedSessionBrowserRawStatusObservationFromEventCycle(fallbackCycle, req.RequestedProfile),
				req.RequestedProfile,
			)
			if !sharedSessionBrowserRawStatusObservationProvided(normalized) {
				return SharedSessionBrowserEventCycleObservation{}, false
			}
			status = &normalized
		}
	}
	var profiles *SharedSessionBrowserRawProfilesObservation
	if req.IncludeProfiles {
		observation, ok := rawProfiles[strings.TrimSpace(req.RequestedProfile)]
		if ok && sharedSessionBrowserRawProfilesObservationProvided(observation) {
			normalized := sharedSessionBrowserRawProfilesObservationForRequestedProfile(observation, req.RequestedProfile)
			profiles = &normalized
		} else {
			if !haveFallbackCycle {
				return SharedSessionBrowserEventCycleObservation{}, false
			}
			normalized := sharedSessionBrowserRawProfilesObservationForRequestedProfile(
				sharedSessionBrowserRawProfilesObservationFromEventCycle(fallbackCycle, req.RequestedProfile),
				req.RequestedProfile,
			)
			if !sharedSessionBrowserRawProfilesObservationProvided(normalized) {
				return SharedSessionBrowserEventCycleObservation{}, false
			}
			profiles = &normalized
		}
	}
	observation := m.seededEventCycleObservationForRequest(req.SessionID, req, status, profiles)
	if !sharedSessionBrowserSeededEventCycleObservationProvided(observation, req) {
		return SharedSessionBrowserEventCycleObservation{}, false
	}
	return observation, true
}

func sharedSessionBrowserProjectionSeedFallbackEventCycle(
	req SharedSessionBrowserObserverRequest,
	cachedEventCycles map[SharedSessionBrowserObserverRequest]SharedSessionBrowserEventCycleObservation,
	routeMutationCycles map[sharedSessionBrowserRouteMutationSourceKey]SharedSessionBrowserEventCycleObservation,
) (SharedSessionBrowserEventCycleObservation, bool) {
	if len(routeMutationCycles) != 0 {
		if observation, ok := sharedSessionBrowserCachedRouteMutationEventCycleFromRouteMutationSources(
			req.SessionID,
			req.SelectedInfo,
			req.RequestedProfile,
			routeMutationCycles,
		); ok {
			return observation, true
		}
	}
	if len(cachedEventCycles) != 0 {
		observation, ok := cachedEventCycles[normalizeSharedSessionBrowserEventCycleRequest(req)]
		if ok && sharedSessionBrowserSeededEventCycleObservationProvided(observation, req) {
			return observation, true
		}
	}
	return SharedSessionBrowserEventCycleObservation{}, false
}

func sharedSessionBrowserCachedRouteMutationEventCycleFromRouteMutationSources(
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	requestedProfile string,
	routeMutationCycles map[sharedSessionBrowserRouteMutationSourceKey]SharedSessionBrowserEventCycleObservation,
) (SharedSessionBrowserEventCycleObservation, bool) {
	if len(routeMutationCycles) == 0 {
		return SharedSessionBrowserEventCycleObservation{}, false
	}
	target := normalizeSharedSessionBrowserRouteMutationSourceKey(sessionID, selectedInfo, requestedProfile)
	if observation, ok := routeMutationCycles[target]; ok &&
		(observation.Observation.Status != nil || observation.Observation.Profiles != nil) {
		return observation, true
	}
	for key, observation := range routeMutationCycles {
		if !sharedSessionBrowserRouteMutationSourceMatches(key, target) {
			continue
		}
		if observation.Observation.Status != nil || observation.Observation.Profiles != nil {
			return observation, true
		}
	}
	return SharedSessionBrowserEventCycleObservation{}, false
}

func sharedSessionBrowserCachedRouteMutationEventCycle(
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	requestedProfile string,
	cachedEventCycles map[SharedSessionBrowserObserverRequest]SharedSessionBrowserEventCycleObservation,
) (SharedSessionBrowserEventCycleObservation, bool) {
	if len(cachedEventCycles) == 0 {
		return SharedSessionBrowserEventCycleObservation{}, false
	}
	sessionID = strings.TrimSpace(sessionID)
	requestedProfile = strings.TrimSpace(requestedProfile)
	exactReq := normalizeSharedSessionBrowserEventCycleRequest(SharedSessionBrowserObserverRequest{
		SessionID:        sessionID,
		SelectedInfo:     selectedInfo,
		BindingRoute:     normalizeBrowserSessionRoute(BrowserSessionRoute{Backend: selectedInfo.Backend, Profile: selectedInfo.Profile, Target: selectedInfo.Target}),
		RequestedProfile: requestedProfile,
		IncludeStatus:    true,
		IncludeProfiles:  true,
	})
	if observation, ok := cachedEventCycles[exactReq]; ok &&
		(observation.Observation.Status != nil || observation.Observation.Profiles != nil) {
		return observation, true
	}
	for req, observation := range cachedEventCycles {
		if req.SessionID != sessionID || req.SelectedInfo != selectedInfo {
			continue
		}
		if strings.TrimSpace(req.RequestedProfile) != requestedProfile {
			continue
		}
		if observation.Observation.Status != nil || observation.Observation.Profiles != nil {
			return observation, true
		}
	}
	for req, observation := range cachedEventCycles {
		if req.SessionID != sessionID || req.SelectedInfo != selectedInfo {
			continue
		}
		if observation.Observation.Status != nil || observation.Observation.Profiles != nil {
			return observation, true
		}
	}
	return SharedSessionBrowserEventCycleObservation{}, false
}

func sharedSessionBrowserRawStatusObservationFromEventCycle(
	observation SharedSessionBrowserEventCycleObservation,
	requestedProfile string,
) SharedSessionBrowserRawStatusObservation {
	return SharedSessionBrowserRawStatusObservation{
		RequestedProfile: strings.TrimSpace(requestedProfile),
		Status:           observation.Observation.Status,
		Err:              observation.Observation.StatusErr,
		ObservedAt:       observation.Observation.StatusObservedAt,
	}
}

func sharedSessionBrowserRawProfilesObservationFromEventCycle(
	observation SharedSessionBrowserEventCycleObservation,
	requestedProfile string,
) SharedSessionBrowserRawProfilesObservation {
	return SharedSessionBrowserRawProfilesObservation{
		RequestedProfile: strings.TrimSpace(requestedProfile),
		Profiles:         observation.Observation.Profiles,
		Err:              observation.Observation.ProfilesErr,
		ObservedAt:       observation.Observation.ProfilesObservedAt,
	}
}

func resetSharedSessionBrowserWatchManagerObservationCachesLocked(state *sharedSessionBrowserWatchManagerState, generation uint64, rawGeneration uint64) {
	if state == nil {
		return
	}
	clear(state.rawStatus)
	clear(state.rawProfiles)
	clear(state.rawRouteMutations)
	clear(state.rawStatusInFlight)
	clear(state.rawProfilesInFlight)
	clear(state.rawStartsInFlight)
	clear(state.rawStopsInFlight)
	clear(state.eventCycles)
	clear(state.bindings)
	clear(state.views)
	clear(state.watchLoops)
	clear(state.eventCyclesInFlight)
	clear(state.bindingsInFlight)
	clear(state.viewsInFlight)
	clear(state.watchLoopsInFlight)
	state.generation = generation
	state.rawGeneration = rawGeneration
}

func sharedSessionBrowserWatchManagerMatchesRuntimeInfo(manager SharedSessionBrowserWatchManager, selectedInfo BrowserRuntimeInfo) bool {
	control := manager.Control
	if control == nil {
		return false
	}
	provider, ok := control.(BrowserRuntimeInfoProvider)
	if !ok {
		return true
	}
	info := provider.BrowserRuntimeInfo()
	if info.Backend != "" && selectedInfo.Backend != "" && !strings.EqualFold(strings.TrimSpace(info.Backend), strings.TrimSpace(selectedInfo.Backend)) {
		return false
	}
	if info.Target != "" && selectedInfo.Target != "" && !strings.EqualFold(strings.TrimSpace(info.Target), strings.TrimSpace(selectedInfo.Target)) {
		return false
	}
	return true
}

func sharedSessionBrowserRawObservationCacheKeys(values ...string) []string {
	seen := make(map[string]struct{}, len(values))
	keys := make([]string, 0, len(values))
	for _, value := range values {
		key := strings.TrimSpace(value)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		keys = append(keys, key)
	}
	return keys
}

func sharedSessionBrowserProjectionSeedRequestsLocked(
	state *sharedSessionBrowserWatchManagerState,
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	profileKeys map[string]struct{},
	haveStatus bool,
	haveProfiles bool,
) []SharedSessionBrowserObserverRequest {
	if state == nil {
		return nil
	}
	sessionID = strings.TrimSpace(sessionID)
	seen := make(map[SharedSessionBrowserObserverRequest]struct{})
	out := make([]SharedSessionBrowserObserverRequest, 0)
	maybeAdd := func(req SharedSessionBrowserObserverRequest) {
		req = normalizeSharedSessionBrowserEventCycleRequest(req)
		if req.SessionID != sessionID || req.SelectedInfo != selectedInfo {
			return
		}
		effectiveProfile := firstNonEmptyBindingString(
			strings.TrimSpace(req.RequestedProfile),
			strings.TrimSpace(req.SelectedInfo.Profile),
			strings.TrimSpace(req.BindingRoute.Profile),
		)
		if len(profileKeys) > 0 {
			if _, ok := profileKeys[effectiveProfile]; !ok {
				return
			}
		}
		if _, ok := seen[req]; ok {
			return
		}
		if !haveStatus && !haveProfiles {
			return
		}
		seen[req] = struct{}{}
		out = append(out, req)
	}
	for req := range state.eventCycles {
		maybeAdd(req)
	}
	for req := range state.bindings {
		maybeAdd(req)
	}
	for req := range state.views {
		maybeAdd(req)
	}
	for req := range state.watchLoops {
		maybeAdd(req)
	}
	return out
}

func sharedSessionBrowserSeededEventCycleObservationProvided(
	observation SharedSessionBrowserEventCycleObservation,
	req SharedSessionBrowserObserverRequest,
) bool {
	if req.IncludeStatus && sharedSessionBrowserRawStatusObservationProvided(SharedSessionBrowserRawStatusObservation{
		Status:     observation.Observation.Status,
		Err:        observation.Observation.StatusErr,
		ObservedAt: observation.Observation.StatusObservedAt,
	}) {
		return true
	}
	if req.IncludeProfiles && sharedSessionBrowserRawProfilesObservationProvided(SharedSessionBrowserRawProfilesObservation{
		Profiles:   observation.Observation.Profiles,
		Err:        observation.Observation.ProfilesErr,
		ObservedAt: observation.Observation.ProfilesObservedAt,
	}) {
		return true
	}
	return false
}

func sharedSessionBrowserProjectionSeedPlanFromStateLocked(state *sharedSessionBrowserWatchManagerState) sharedSessionBrowserProjectionSeedPlan {
	plan := sharedSessionBrowserProjectionSeedPlan{
		eventCycles: map[SharedSessionBrowserObserverRequest]struct{}{},
		bindings:    map[SharedSessionBrowserObserverRequest]struct{}{},
		views:       map[SharedSessionBrowserObserverRequest]struct{}{},
		watchLoops:  map[SharedSessionBrowserObserverRequest]struct{}{},
	}
	if state == nil {
		return plan
	}
	for req := range state.eventCycles {
		plan.eventCycles[normalizeSharedSessionBrowserEventCycleRequest(req)] = struct{}{}
	}
	for req := range state.bindings {
		normalized := normalizeSharedSessionBrowserBindingRequest(req)
		plan.bindings[normalized] = struct{}{}
		plan.eventCycles[normalizeSharedSessionBrowserEventCycleRequest(normalized)] = struct{}{}
	}
	for req := range state.views {
		normalized := normalizeSharedSessionBrowserViewRequest(req)
		plan.views[normalized] = struct{}{}
		bindingReq := normalizeSharedSessionBrowserBindingRequest(normalized)
		plan.bindings[bindingReq] = struct{}{}
		plan.eventCycles[normalizeSharedSessionBrowserEventCycleRequest(bindingReq)] = struct{}{}
	}
	for req := range state.watchLoops {
		normalized := normalizeSharedSessionBrowserWatchLoopRequest(req)
		plan.watchLoops[normalized] = struct{}{}
		plan.eventCycles[normalizeSharedSessionBrowserEventCycleRequest(normalized)] = struct{}{}
	}
	return plan
}

func sharedSessionBrowserRawStatusObservationForRequestedProfile(
	observation SharedSessionBrowserRawStatusObservation,
	requestedProfile string,
) SharedSessionBrowserRawStatusObservation {
	observation.RequestedProfile = strings.TrimSpace(requestedProfile)
	return normalizeSharedSessionBrowserRawStatusObservation(observation, observation.RequestedProfile)
}

func sharedSessionBrowserRawProfilesObservationForRequestedProfile(
	observation SharedSessionBrowserRawProfilesObservation,
	requestedProfile string,
) SharedSessionBrowserRawProfilesObservation {
	observation.RequestedProfile = strings.TrimSpace(requestedProfile)
	return normalizeSharedSessionBrowserRawProfilesObservation(observation, observation.RequestedProfile)
}
