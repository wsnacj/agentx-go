package browserruntime

import (
	"context"
	"strings"
	"time"
)

func (m SharedSessionBrowserWatchManager) cachedEventCycle(req SharedSessionBrowserObserverRequest) (SharedSessionBrowserEventCycleObservation, bool) {
	m.syncProjectionGeneration()
	if m.state == nil {
		return SharedSessionBrowserEventCycleObservation{}, false
	}
	now := time.Now()
	m.state.mu.RLock()
	cached, ok := m.state.eventCycles[req]
	if ok && now.Sub(cached.cachedAt) <= sharedSessionBrowserWatchManagerEventCycleCacheTTL {
		m.state.mu.RUnlock()
		return cached.observation, true
	}
	for cachedReq, cached := range m.state.eventCycles {
		if !sharedSessionBrowserEventCycleRequestCanSatisfy(cachedReq, req) {
			continue
		}
		if now.Sub(cached.cachedAt) > sharedSessionBrowserWatchManagerEventCycleCacheTTL {
			continue
		}
		observation := projectSharedSessionBrowserEventCycleObservationForRequest(cached.observation, req)
		m.state.mu.RUnlock()
		return observation, true
	}
	m.state.mu.RUnlock()
	return SharedSessionBrowserEventCycleObservation{}, false
}

func (m SharedSessionBrowserWatchManager) cachedWatchLoop(req SharedSessionBrowserObserverRequest) (SharedSessionBrowserWatchLoopObservation, bool) {
	m.syncProjectionGeneration()
	if m.state == nil {
		return SharedSessionBrowserWatchLoopObservation{}, false
	}
	now := time.Now()
	m.state.mu.RLock()
	cached, ok := m.state.watchLoops[req]
	if ok && now.Sub(cached.cachedAt) <= sharedSessionBrowserWatchManagerEventCycleCacheTTL {
		m.state.mu.RUnlock()
		return cached.observation, true
	}
	for cachedReq, cached := range m.state.watchLoops {
		if !sharedSessionBrowserWatchLoopRequestCanSatisfy(cachedReq, req) {
			continue
		}
		if now.Sub(cached.cachedAt) > sharedSessionBrowserWatchManagerEventCycleCacheTTL {
			continue
		}
		observation := projectSharedSessionBrowserWatchLoopObservationForRequest(cached.observation, req)
		m.state.mu.RUnlock()
		return observation, true
	}
	m.state.mu.RUnlock()
	return SharedSessionBrowserWatchLoopObservation{}, false
}

func (m SharedSessionBrowserWatchManager) cachedBinding(req SharedSessionBrowserObserverRequest) (SharedSessionBrowserBindingObservation, bool) {
	m.syncProjectionGeneration()
	if m.state == nil {
		return SharedSessionBrowserBindingObservation{}, false
	}
	now := time.Now()
	m.state.mu.RLock()
	cached, ok := m.state.bindings[req]
	m.state.mu.RUnlock()
	if !ok || now.Sub(cached.cachedAt) > sharedSessionBrowserWatchManagerEventCycleCacheTTL {
		return SharedSessionBrowserBindingObservation{}, false
	}
	return cached.observation, true
}

func (m SharedSessionBrowserWatchManager) cachedView(req SharedSessionBrowserObserverRequest) (SharedSessionBrowserViewObservation, bool) {
	m.syncProjectionGeneration()
	if m.state == nil {
		return SharedSessionBrowserViewObservation{}, false
	}
	now := time.Now()
	m.state.mu.RLock()
	cached, ok := m.state.views[req]
	m.state.mu.RUnlock()
	if !ok || now.Sub(cached.cachedAt) > sharedSessionBrowserWatchManagerEventCycleCacheTTL {
		return SharedSessionBrowserViewObservation{}, false
	}
	return cached.observation, true
}

func (m SharedSessionBrowserWatchManager) cachedRawStatus(requestedProfile string) (SharedSessionBrowserRawStatusObservation, bool) {
	m.syncRawGeneration()
	if m.state == nil {
		return SharedSessionBrowserRawStatusObservation{}, false
	}
	m.state.mu.RLock()
	cached, ok := m.state.rawStatus[requestedProfile]
	m.state.mu.RUnlock()
	if !ok || time.Since(cached.cachedAt) > sharedSessionBrowserWatchManagerEventCycleCacheTTL {
		return SharedSessionBrowserRawStatusObservation{}, false
	}
	return cached.observation, true
}

func (m SharedSessionBrowserWatchManager) cachedRawProfiles(requestedProfile string) (SharedSessionBrowserRawProfilesObservation, bool) {
	m.syncRawGeneration()
	if m.state == nil {
		return SharedSessionBrowserRawProfilesObservation{}, false
	}
	m.state.mu.RLock()
	cached, ok := m.state.rawProfiles[requestedProfile]
	m.state.mu.RUnlock()
	if !ok || time.Since(cached.cachedAt) > sharedSessionBrowserWatchManagerEventCycleCacheTTL {
		return SharedSessionBrowserRawProfilesObservation{}, false
	}
	return cached.observation, true
}

func normalizeSharedSessionBrowserRouteMutationSourceKey(
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	requestedProfile string,
) sharedSessionBrowserRouteMutationSourceKey {
	return sharedSessionBrowserRouteMutationSourceKey{
		sessionID: strings.TrimSpace(sessionID),
		selectedInfo: BrowserRuntimeInfo{
			Backend: strings.TrimSpace(selectedInfo.Backend),
			Profile: strings.TrimSpace(selectedInfo.Profile),
			Target:  strings.TrimSpace(selectedInfo.Target),
		},
		requestedProfile: strings.TrimSpace(requestedProfile),
	}
}

func sharedSessionBrowserRouteMutationSourceMatches(
	candidate sharedSessionBrowserRouteMutationSourceKey,
	target sharedSessionBrowserRouteMutationSourceKey,
) bool {
	if candidate.sessionID != target.sessionID || candidate.selectedInfo != target.selectedInfo {
		return false
	}
	return candidate.requestedProfile == target.requestedProfile ||
		candidate.requestedProfile == "" ||
		target.requestedProfile == ""
}

func (m SharedSessionBrowserWatchManager) cachedRouteMutationSource(
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	requestedProfile string,
) (SharedSessionBrowserEventCycleObservation, bool) {
	m.syncRawGeneration()
	if m.state == nil {
		return SharedSessionBrowserEventCycleObservation{}, false
	}
	now := time.Now()
	key := normalizeSharedSessionBrowserRouteMutationSourceKey(sessionID, selectedInfo, requestedProfile)
	m.state.mu.RLock()
	cached, ok := m.state.routeMutations[key]
	if ok && now.Sub(cached.cachedAt) <= sharedSessionBrowserWatchManagerEventCycleCacheTTL {
		m.state.mu.RUnlock()
		return cached.observation, true
	}
	for cachedKey, candidate := range m.state.routeMutations {
		if !sharedSessionBrowserRouteMutationSourceMatches(cachedKey, key) {
			continue
		}
		if now.Sub(candidate.cachedAt) > sharedSessionBrowserWatchManagerEventCycleCacheTTL {
			continue
		}
		m.state.mu.RUnlock()
		return candidate.observation, true
	}
	m.state.mu.RUnlock()
	return SharedSessionBrowserEventCycleObservation{}, false
}

func (m SharedSessionBrowserWatchManager) cachedRawRouteMutationSource(
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	requestedProfile string,
) (SharedSessionBrowserRawRouteMutationObservation, bool) {
	m.syncRawGeneration()
	if m.state == nil {
		return SharedSessionBrowserRawRouteMutationObservation{}, false
	}
	now := time.Now()
	key := normalizeSharedSessionBrowserRouteMutationSourceKey(sessionID, selectedInfo, requestedProfile)
	m.state.mu.RLock()
	cached, ok := m.state.rawRouteMutations[key]
	if ok && now.Sub(cached.cachedAt) <= sharedSessionBrowserWatchManagerEventCycleCacheTTL {
		m.state.mu.RUnlock()
		return cached.observation, true
	}
	for cachedKey, candidate := range m.state.rawRouteMutations {
		if !sharedSessionBrowserRouteMutationSourceMatches(cachedKey, key) {
			continue
		}
		if now.Sub(candidate.cachedAt) > sharedSessionBrowserWatchManagerEventCycleCacheTTL {
			continue
		}
		m.state.mu.RUnlock()
		return candidate.observation, true
	}
	m.state.mu.RUnlock()
	return SharedSessionBrowserRawRouteMutationObservation{}, false
}

func (m SharedSessionBrowserWatchManager) storeRouteMutationSource(
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	requestedProfile string,
	observation SharedSessionBrowserEventCycleObservation,
) {
	if m.state == nil {
		return
	}
	key := normalizeSharedSessionBrowserRouteMutationSourceKey(sessionID, selectedInfo, requestedProfile)
	m.state.mu.Lock()
	m.state.routeMutations[key] = sharedSessionBrowserCachedEventCycleObservation{
		cachedAt:    time.Now(),
		observation: observation,
	}
	m.state.mu.Unlock()
}

func (m SharedSessionBrowserWatchManager) storeRawRouteMutationSource(
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	requestedProfile string,
	observation SharedSessionBrowserRawRouteMutationObservation,
) {
	if m.state == nil {
		return
	}
	key := normalizeSharedSessionBrowserRouteMutationSourceKey(sessionID, selectedInfo, requestedProfile)
	m.state.mu.Lock()
	m.state.rawRouteMutations[key] = sharedSessionBrowserCachedRawRouteMutationObservation{
		cachedAt:    time.Now(),
		observation: observation,
	}
	m.state.mu.Unlock()
}

func (m SharedSessionBrowserWatchManager) storeRawStatus(requestedProfile string, observation SharedSessionBrowserRawStatusObservation) {
	if m.state == nil {
		return
	}
	m.state.mu.Lock()
	m.state.rawStatus[requestedProfile] = sharedSessionBrowserCachedRawStatusObservation{
		cachedAt:    time.Now(),
		observation: observation,
	}
	m.state.mu.Unlock()
}

func (m SharedSessionBrowserWatchManager) storeRawProfiles(requestedProfile string, observation SharedSessionBrowserRawProfilesObservation) {
	if m.state == nil {
		return
	}
	m.state.mu.Lock()
	m.state.rawProfiles[requestedProfile] = sharedSessionBrowserCachedRawProfilesObservation{
		cachedAt:    time.Now(),
		observation: observation,
	}
	m.state.mu.Unlock()
}

func (m SharedSessionBrowserWatchManager) seedRawStatusFromExecutionLifecycle(profile string, observation SharedSessionBrowserExecutionLifecycleObservation) {
	if m.state == nil || sharedSessionBrowserProfileStatusResultEmpty(observation.Status) {
		return
	}
	observedAt := observation.ObservedAt
	if observedAt.IsZero() {
		observedAt = time.Now()
	}
	keys := sharedSessionBrowserRawObservationCacheKeys(
		"",
		profile,
		observation.Profile,
		observation.Status.Profile,
	)
	for _, requestedProfile := range keys {
		status := observation.Status
		m.storeRawStatus(
			requestedProfile,
			normalizeSharedSessionBrowserRawStatusObservation(
				SharedSessionBrowserRawStatusObservation{
					RequestedProfile: requestedProfile,
					Status:           &status,
					ObservedAt:       observedAt,
				},
				requestedProfile,
			),
		)
	}
}

func (m SharedSessionBrowserWatchManager) snapshotRawProfilesForExecutionLifecycle(values ...string) map[string]SharedSessionBrowserRawProfilesObservation {
	keys := sharedSessionBrowserRawObservationCacheKeys(values...)
	if len(keys) == 0 {
		return nil
	}
	snapshots := make(map[string]SharedSessionBrowserRawProfilesObservation, len(keys))
	for _, requestedProfile := range keys {
		cached, ok := m.cachedRawProfiles(requestedProfile)
		if !ok || cached.Profiles == nil {
			continue
		}
		snapshots[requestedProfile] = cached
	}
	if len(snapshots) == 0 {
		return nil
	}
	return snapshots
}

func sharedSessionBrowserLifecycleSeededProfilesResult(
	current BrowserProfilesResult,
	observation SharedSessionBrowserExecutionLifecycleObservation,
) (BrowserProfilesResult, bool) {
	profile := firstNonEmptyString(
		strings.TrimSpace(observation.Profile),
		strings.TrimSpace(observation.Status.Profile),
	)
	if profile == "" {
		return BrowserProfilesResult{}, false
	}
	if len(current.Profiles) == 0 {
		return BrowserProfilesResult{}, false
	}
	updated := current
	updated.Backend = firstNonEmptyString(strings.TrimSpace(updated.Backend), strings.TrimSpace(observation.Status.Backend))
	updated.Profiles = append([]BrowserProfileInfo(nil), current.Profiles...)
	patched := false
	for i, item := range updated.Profiles {
		if !strings.EqualFold(strings.TrimSpace(item.Profile), profile) {
			continue
		}
		updated.Profiles[i] = BrowserProfileInfo{
			Profile:    firstNonEmptyString(strings.TrimSpace(item.Profile), profile),
			BrowserApp: firstNonEmptyString(strings.TrimSpace(observation.Status.BrowserApp), strings.TrimSpace(item.BrowserApp)),
			Status:     strings.TrimSpace(observation.Status.Status),
			Running:    observation.Status.Running,
			Connected:  observation.Status.Connected,
			Note:       strings.TrimSpace(observation.Status.Note),
		}
		patched = true
		break
	}
	if !patched {
		updated.Profiles = append(updated.Profiles, BrowserProfileInfo{
			Profile:    profile,
			BrowserApp: strings.TrimSpace(observation.Status.BrowserApp),
			Status:     strings.TrimSpace(observation.Status.Status),
			Running:    observation.Status.Running,
			Connected:  observation.Status.Connected,
			Note:       strings.TrimSpace(observation.Status.Note),
		})
	}
	if strings.TrimSpace(updated.DefaultProfile) == "" {
		updated.DefaultProfile = profile
	}
	return updated, true
}

func (m SharedSessionBrowserWatchManager) seedRawProfilesFromExecutionLifecycle(
	profile string,
	observation SharedSessionBrowserExecutionLifecycleObservation,
	prior map[string]SharedSessionBrowserRawProfilesObservation,
) {
	if m.state == nil || sharedSessionBrowserProfileStatusResultEmpty(observation.Status) {
		return
	}
	observedAt := observation.ObservedAt
	if observedAt.IsZero() {
		observedAt = time.Now()
	}
	keys := sharedSessionBrowserRawObservationCacheKeys(
		"",
		profile,
		observation.Profile,
		observation.Status.Profile,
	)
	for _, requestedProfile := range keys {
		cached, ok := prior[requestedProfile]
		if !ok {
			var liveOK bool
			cached, liveOK = m.cachedRawProfiles(requestedProfile)
			ok = liveOK
		}
		if !ok || cached.Profiles == nil {
			continue
		}
		seeded, updated := sharedSessionBrowserLifecycleSeededProfilesResult(*cached.Profiles, observation)
		if !updated {
			continue
		}
		m.storeRawProfiles(
			requestedProfile,
			normalizeSharedSessionBrowserRawProfilesObservation(
				SharedSessionBrowserRawProfilesObservation{
					RequestedProfile: requestedProfile,
					Profiles:         &seeded,
					ObservedAt:       observedAt,
				},
				requestedProfile,
			),
		)
	}
}

func (m SharedSessionBrowserWatchManager) storeEventCycle(req SharedSessionBrowserObserverRequest, observation SharedSessionBrowserEventCycleObservation) {
	if m.state == nil {
		return
	}
	m.state.mu.Lock()
	m.state.eventCycles[req] = sharedSessionBrowserCachedEventCycleObservation{
		cachedAt:    time.Now(),
		observation: observation,
	}
	m.state.mu.Unlock()
}

func (m SharedSessionBrowserWatchManager) storeBinding(req SharedSessionBrowserObserverRequest, observation SharedSessionBrowserBindingObservation) {
	if m.state == nil {
		return
	}
	m.state.mu.Lock()
	m.state.bindings[req] = sharedSessionBrowserCachedBindingObservation{
		cachedAt:    time.Now(),
		observation: observation,
	}
	m.state.mu.Unlock()
}

func (m SharedSessionBrowserWatchManager) storeView(req SharedSessionBrowserObserverRequest, observation SharedSessionBrowserViewObservation) {
	if m.state == nil {
		return
	}
	m.state.mu.Lock()
	m.state.views[req] = sharedSessionBrowserCachedViewObservation{
		cachedAt:    time.Now(),
		observation: observation,
	}
	m.state.mu.Unlock()
}

func (m SharedSessionBrowserWatchManager) storeWatchLoop(req SharedSessionBrowserObserverRequest, observation SharedSessionBrowserWatchLoopObservation) {
	if m.state == nil {
		return
	}
	m.state.mu.Lock()
	m.state.watchLoops[req] = sharedSessionBrowserCachedWatchLoopObservation{
		cachedAt:    time.Now(),
		observation: observation,
	}
	m.state.mu.Unlock()
}

func (m SharedSessionBrowserWatchManager) beginRawStatusInFlight(requestedProfile string) (*sharedSessionBrowserInFlightRawStatusObservation, bool) {
	m.syncRawGeneration()
	if m.state == nil {
		return nil, true
	}
	m.state.mu.Lock()
	defer m.state.mu.Unlock()
	if inFlight, ok := m.state.rawStatusInFlight[requestedProfile]; ok {
		return inFlight, false
	}
	inFlight := &sharedSessionBrowserInFlightRawStatusObservation{ready: make(chan struct{})}
	m.state.rawStatusInFlight[requestedProfile] = inFlight
	return inFlight, true
}

func (m SharedSessionBrowserWatchManager) awaitRawStatusInFlight(
	ctx context.Context,
	requestedProfile string,
	inFlight *sharedSessionBrowserInFlightRawStatusObservation,
) SharedSessionBrowserRawStatusObservation {
	if inFlight == nil {
		return SharedSessionBrowserRawStatusObservation{RequestedProfile: requestedProfile}
	}
	select {
	case <-inFlight.ready:
		return inFlight.observation
	case <-ctx.Done():
		return SharedSessionBrowserRawStatusObservation{
			RequestedProfile: requestedProfile,
			Err:              ctx.Err(),
		}
	}
}

func (m SharedSessionBrowserWatchManager) finishRawStatusInFlight(
	requestedProfile string,
	inFlight *sharedSessionBrowserInFlightRawStatusObservation,
	observation SharedSessionBrowserRawStatusObservation,
) {
	if inFlight == nil {
		return
	}
	if m.state != nil {
		m.state.mu.Lock()
		if current, ok := m.state.rawStatusInFlight[requestedProfile]; ok && current == inFlight {
			delete(m.state.rawStatusInFlight, requestedProfile)
		}
		inFlight.observation = observation
		close(inFlight.ready)
		m.state.mu.Unlock()
		return
	}
	inFlight.observation = observation
	close(inFlight.ready)
}

func (m SharedSessionBrowserWatchManager) beginRawProfilesInFlight(requestedProfile string) (*sharedSessionBrowserInFlightRawProfilesObservation, bool) {
	m.syncRawGeneration()
	if m.state == nil {
		return nil, true
	}
	m.state.mu.Lock()
	defer m.state.mu.Unlock()
	if inFlight, ok := m.state.rawProfilesInFlight[requestedProfile]; ok {
		return inFlight, false
	}
	inFlight := &sharedSessionBrowserInFlightRawProfilesObservation{ready: make(chan struct{})}
	m.state.rawProfilesInFlight[requestedProfile] = inFlight
	return inFlight, true
}

func (m SharedSessionBrowserWatchManager) beginRawStartInFlight(profile string) (*sharedSessionBrowserInFlightRawLifecycleObservation, bool) {
	m.syncRawGeneration()
	if m.state == nil {
		return nil, true
	}
	profile = strings.TrimSpace(profile)
	m.state.mu.Lock()
	defer m.state.mu.Unlock()
	if inFlight, ok := m.state.rawStartsInFlight[profile]; ok {
		return inFlight, false
	}
	inFlight := &sharedSessionBrowserInFlightRawLifecycleObservation{ready: make(chan struct{})}
	m.state.rawStartsInFlight[profile] = inFlight
	return inFlight, true
}

func (m SharedSessionBrowserWatchManager) beginRawStopInFlight(profile string) (*sharedSessionBrowserInFlightRawLifecycleObservation, bool) {
	m.syncRawGeneration()
	if m.state == nil {
		return nil, true
	}
	profile = strings.TrimSpace(profile)
	m.state.mu.Lock()
	defer m.state.mu.Unlock()
	if inFlight, ok := m.state.rawStopsInFlight[profile]; ok {
		return inFlight, false
	}
	inFlight := &sharedSessionBrowserInFlightRawLifecycleObservation{ready: make(chan struct{})}
	m.state.rawStopsInFlight[profile] = inFlight
	return inFlight, true
}

func (m SharedSessionBrowserWatchManager) awaitRawProfilesInFlight(
	ctx context.Context,
	requestedProfile string,
	inFlight *sharedSessionBrowserInFlightRawProfilesObservation,
) SharedSessionBrowserRawProfilesObservation {
	if inFlight == nil {
		return SharedSessionBrowserRawProfilesObservation{RequestedProfile: requestedProfile}
	}
	select {
	case <-inFlight.ready:
		return inFlight.observation
	case <-ctx.Done():
		return SharedSessionBrowserRawProfilesObservation{
			RequestedProfile: requestedProfile,
			Err:              ctx.Err(),
		}
	}
}

func (m SharedSessionBrowserWatchManager) awaitRawLifecycleInFlight(
	ctx context.Context,
	profile string,
	inFlight *sharedSessionBrowserInFlightRawLifecycleObservation,
) SharedSessionBrowserRawLifecycleObservation {
	if inFlight == nil {
		return SharedSessionBrowserRawLifecycleObservation{Profile: strings.TrimSpace(profile)}
	}
	select {
	case <-inFlight.ready:
		return inFlight.observation
	case <-ctx.Done():
		return SharedSessionBrowserRawLifecycleObservation{
			Profile: strings.TrimSpace(profile),
			Err:     ctx.Err(),
		}
	}
}

func (m SharedSessionBrowserWatchManager) finishRawProfilesInFlight(
	requestedProfile string,
	inFlight *sharedSessionBrowserInFlightRawProfilesObservation,
	observation SharedSessionBrowserRawProfilesObservation,
) {
	if inFlight == nil {
		return
	}
	if m.state != nil {
		m.state.mu.Lock()
		if current, ok := m.state.rawProfilesInFlight[requestedProfile]; ok && current == inFlight {
			delete(m.state.rawProfilesInFlight, requestedProfile)
		}
		inFlight.observation = observation
		close(inFlight.ready)
		m.state.mu.Unlock()
		return
	}
	inFlight.observation = observation
	close(inFlight.ready)
}

func (m SharedSessionBrowserWatchManager) finishRawStartInFlight(
	profile string,
	inFlight *sharedSessionBrowserInFlightRawLifecycleObservation,
	observation SharedSessionBrowserRawLifecycleObservation,
) {
	if inFlight == nil {
		return
	}
	profile = strings.TrimSpace(profile)
	if m.state != nil {
		m.state.mu.Lock()
		if current, ok := m.state.rawStartsInFlight[profile]; ok && current == inFlight {
			delete(m.state.rawStartsInFlight, profile)
		}
		inFlight.observation = observation
		close(inFlight.ready)
		m.state.mu.Unlock()
		return
	}
	inFlight.observation = observation
	close(inFlight.ready)
}

func (m SharedSessionBrowserWatchManager) finishRawStopInFlight(
	profile string,
	inFlight *sharedSessionBrowserInFlightRawLifecycleObservation,
	observation SharedSessionBrowserRawLifecycleObservation,
) {
	if inFlight == nil {
		return
	}
	profile = strings.TrimSpace(profile)
	if m.state != nil {
		m.state.mu.Lock()
		if current, ok := m.state.rawStopsInFlight[profile]; ok && current == inFlight {
			delete(m.state.rawStopsInFlight, profile)
		}
		inFlight.observation = observation
		close(inFlight.ready)
		m.state.mu.Unlock()
		return
	}
	inFlight.observation = observation
	close(inFlight.ready)
}

func (m SharedSessionBrowserWatchManager) beginEventCycleInFlight(req SharedSessionBrowserObserverRequest) (*sharedSessionBrowserInFlightEventCycleObservation, bool) {
	m.syncProjectionGeneration()
	if m.state == nil {
		return nil, true
	}
	m.state.mu.Lock()
	defer m.state.mu.Unlock()
	if inFlight, ok := m.state.eventCyclesInFlight[req]; ok {
		return inFlight, false
	}
	for inFlightReq, inFlight := range m.state.eventCyclesInFlight {
		if sharedSessionBrowserEventCycleRequestCanSatisfy(inFlightReq, req) {
			return inFlight, false
		}
	}
	inFlight := &sharedSessionBrowserInFlightEventCycleObservation{ready: make(chan struct{})}
	m.state.eventCyclesInFlight[req] = inFlight
	return inFlight, true
}

func (m SharedSessionBrowserWatchManager) awaitEventCycleInFlight(
	ctx context.Context,
	req SharedSessionBrowserObserverRequest,
	inFlight *sharedSessionBrowserInFlightEventCycleObservation,
) SharedSessionBrowserEventCycleObservation {
	if inFlight == nil {
		return SharedSessionBrowserEventCycleObservation{}
	}
	select {
	case <-inFlight.ready:
		return projectSharedSessionBrowserEventCycleObservationForRequest(inFlight.observation, req)
	case <-ctx.Done():
		return SharedSessionBrowserEventCycleObservation{}
	}
}

func (m SharedSessionBrowserWatchManager) finishEventCycleInFlight(
	req SharedSessionBrowserObserverRequest,
	inFlight *sharedSessionBrowserInFlightEventCycleObservation,
	observation SharedSessionBrowserEventCycleObservation,
) {
	if inFlight == nil {
		return
	}
	if m.state != nil {
		m.state.mu.Lock()
		if current, ok := m.state.eventCyclesInFlight[req]; ok && current == inFlight {
			delete(m.state.eventCyclesInFlight, req)
		}
		inFlight.observation = observation
		close(inFlight.ready)
		m.state.mu.Unlock()
		return
	}
	inFlight.observation = observation
	close(inFlight.ready)
}

func (m SharedSessionBrowserWatchManager) beginWatchLoopInFlight(req SharedSessionBrowserObserverRequest) (*sharedSessionBrowserInFlightWatchLoopObservation, bool) {
	m.syncProjectionGeneration()
	if m.state == nil {
		return nil, true
	}
	m.state.mu.Lock()
	defer m.state.mu.Unlock()
	if inFlight, ok := m.state.watchLoopsInFlight[req]; ok {
		return inFlight, false
	}
	for inFlightReq, inFlight := range m.state.watchLoopsInFlight {
		if sharedSessionBrowserWatchLoopRequestCanSatisfy(inFlightReq, req) {
			return inFlight, false
		}
	}
	inFlight := &sharedSessionBrowserInFlightWatchLoopObservation{ready: make(chan struct{})}
	m.state.watchLoopsInFlight[req] = inFlight
	return inFlight, true
}

func (m SharedSessionBrowserWatchManager) beginBindingInFlight(req SharedSessionBrowserObserverRequest) (*sharedSessionBrowserInFlightBindingObservation, bool) {
	m.syncProjectionGeneration()
	if m.state == nil {
		return nil, true
	}
	m.state.mu.Lock()
	defer m.state.mu.Unlock()
	if inFlight, ok := m.state.bindingsInFlight[req]; ok {
		return inFlight, false
	}
	inFlight := &sharedSessionBrowserInFlightBindingObservation{ready: make(chan struct{})}
	m.state.bindingsInFlight[req] = inFlight
	return inFlight, true
}

func (m SharedSessionBrowserWatchManager) beginViewInFlight(req SharedSessionBrowserObserverRequest) (*sharedSessionBrowserInFlightViewObservation, bool) {
	m.syncProjectionGeneration()
	if m.state == nil {
		return nil, true
	}
	m.state.mu.Lock()
	defer m.state.mu.Unlock()
	if inFlight, ok := m.state.viewsInFlight[req]; ok {
		return inFlight, false
	}
	inFlight := &sharedSessionBrowserInFlightViewObservation{ready: make(chan struct{})}
	m.state.viewsInFlight[req] = inFlight
	return inFlight, true
}

func (m SharedSessionBrowserWatchManager) awaitBindingInFlight(
	ctx context.Context,
	inFlight *sharedSessionBrowserInFlightBindingObservation,
) SharedSessionBrowserBindingObservation {
	if inFlight == nil {
		return SharedSessionBrowserBindingObservation{}
	}
	select {
	case <-inFlight.ready:
		return inFlight.observation
	case <-ctx.Done():
		return SharedSessionBrowserBindingObservation{}
	}
}

func (m SharedSessionBrowserWatchManager) awaitViewInFlight(
	ctx context.Context,
	inFlight *sharedSessionBrowserInFlightViewObservation,
) SharedSessionBrowserViewObservation {
	if inFlight == nil {
		return SharedSessionBrowserViewObservation{}
	}
	select {
	case <-inFlight.ready:
		return inFlight.observation
	case <-ctx.Done():
		return SharedSessionBrowserViewObservation{}
	}
}

func (m SharedSessionBrowserWatchManager) finishBindingInFlight(
	req SharedSessionBrowserObserverRequest,
	inFlight *sharedSessionBrowserInFlightBindingObservation,
	observation SharedSessionBrowserBindingObservation,
) {
	if inFlight == nil {
		return
	}
	if m.state != nil {
		m.state.mu.Lock()
		if current, ok := m.state.bindingsInFlight[req]; ok && current == inFlight {
			delete(m.state.bindingsInFlight, req)
		}
		inFlight.observation = observation
		close(inFlight.ready)
		m.state.mu.Unlock()
		return
	}
	inFlight.observation = observation
	close(inFlight.ready)
}

func (m SharedSessionBrowserWatchManager) finishViewInFlight(
	req SharedSessionBrowserObserverRequest,
	inFlight *sharedSessionBrowserInFlightViewObservation,
	observation SharedSessionBrowserViewObservation,
) {
	if inFlight == nil {
		return
	}
	if m.state != nil {
		m.state.mu.Lock()
		if current, ok := m.state.viewsInFlight[req]; ok && current == inFlight {
			delete(m.state.viewsInFlight, req)
		}
		inFlight.observation = observation
		close(inFlight.ready)
		m.state.mu.Unlock()
		return
	}
	inFlight.observation = observation
	close(inFlight.ready)
}

func (m SharedSessionBrowserWatchManager) awaitWatchLoopInFlight(
	ctx context.Context,
	req SharedSessionBrowserObserverRequest,
	inFlight *sharedSessionBrowserInFlightWatchLoopObservation,
) SharedSessionBrowserWatchLoopObservation {
	if inFlight == nil {
		return SharedSessionBrowserWatchLoopObservation{}
	}
	select {
	case <-inFlight.ready:
		return projectSharedSessionBrowserWatchLoopObservationForRequest(inFlight.observation, req)
	case <-ctx.Done():
		return SharedSessionBrowserWatchLoopObservation{}
	}
}

func (m SharedSessionBrowserWatchManager) finishWatchLoopInFlight(
	req SharedSessionBrowserObserverRequest,
	inFlight *sharedSessionBrowserInFlightWatchLoopObservation,
	observation SharedSessionBrowserWatchLoopObservation,
) {
	if inFlight == nil {
		return
	}
	if m.state != nil {
		m.state.mu.Lock()
		if current, ok := m.state.watchLoopsInFlight[req]; ok && current == inFlight {
			delete(m.state.watchLoopsInFlight, req)
		}
		inFlight.observation = observation
		close(inFlight.ready)
		m.state.mu.Unlock()
		return
	}
	inFlight.observation = observation
	close(inFlight.ready)
}

func sharedSessionBrowserEventCycleRequestCanSatisfy(candidate SharedSessionBrowserObserverRequest, requested SharedSessionBrowserObserverRequest) bool {
	if candidate.SessionID != requested.SessionID {
		return false
	}
	if candidate.SelectedInfo != requested.SelectedInfo {
		return false
	}
	if candidate.BindingRoute != requested.BindingRoute {
		return false
	}
	if candidate.RequestedProfile != requested.RequestedProfile {
		return false
	}
	if requested.IncludeStatus && !candidate.IncludeStatus {
		return false
	}
	if requested.IncludeProfiles && !candidate.IncludeProfiles {
		return false
	}
	return true
}

func sharedSessionBrowserWatchLoopRequestCanSatisfy(candidate SharedSessionBrowserObserverRequest, requested SharedSessionBrowserObserverRequest) bool {
	if candidate.SessionID != requested.SessionID {
		return false
	}
	if candidate.SelectedInfo != requested.SelectedInfo {
		return false
	}
	if candidate.BindingRoute != requested.BindingRoute {
		return false
	}
	if candidate.RequestedProfile != requested.RequestedProfile {
		return false
	}
	if candidate.IncludeStatus != requested.IncludeStatus {
		return false
	}
	if candidate.IncludeProfiles != requested.IncludeProfiles {
		return false
	}
	if requested.IncludeSessionView {
		return candidate.IncludeSessionView &&
			candidate.SessionViewInfo == requested.SessionViewInfo &&
			candidate.SessionViewRouteFilter == requested.SessionViewRouteFilter &&
			candidate.SessionViewRequestedProfile == requested.SessionViewRequestedProfile
	}
	return true
}

func projectSharedSessionBrowserEventCycleObservationForRequest(
	observation SharedSessionBrowserEventCycleObservation,
	req SharedSessionBrowserObserverRequest,
) SharedSessionBrowserEventCycleObservation {
	projected := observation
	if !req.IncludeStatus {
		projected.Observation.Status = nil
		projected.Observation.StatusErr = nil
		projected.Observation.StatusObservedAt = time.Time{}
		projected.Observation.ResolvedStatus = BrowserProfileStatusResult{}
		projected.Observation.SyncedState = SharedSessionBrowserProfileState{}
		projected.Observation.HasSyncedState = false
	}
	if !req.IncludeProfiles {
		projected.Observation.Profiles = nil
		projected.Observation.ProfilesErr = nil
		projected.Observation.ProfilesObservedAt = time.Time{}
		projected.Observation.Snapshot = nil
	}
	projected.ReferenceTime = sharedSessionBrowserLatestEventCycleObservedAt(
		projected.Observation.StatusObservedAt,
		projected.Observation.ProfilesObservedAt,
	)
	return projected
}

func projectSharedSessionBrowserWatchLoopObservationForRequest(
	observation SharedSessionBrowserWatchLoopObservation,
	req SharedSessionBrowserObserverRequest,
) SharedSessionBrowserWatchLoopObservation {
	if req.IncludeSessionView {
		return observation
	}
	projected := observation
	projected.Observer.Session = SharedSessionBrowserSessionViewSnapshot{}
	projected.View.Session = SharedSessionBrowserSessionViewSnapshot{}
	projected.Watch.View.Session = SharedSessionBrowserSessionViewSnapshot{}
	return projected
}

func (m SharedSessionBrowserWatchManager) invalidateEventCycleCache() {
	if m.state == nil {
		return
	}
	m.state.mu.Lock()
	clear(m.state.rawStatus)
	clear(m.state.rawProfiles)
	clear(m.state.rawStatusInFlight)
	clear(m.state.rawProfilesInFlight)
	clear(m.state.rawStartsInFlight)
	clear(m.state.rawStopsInFlight)
	clear(m.state.eventCycles)
	clear(m.state.bindings)
	clear(m.state.views)
	clear(m.state.watchLoops)
	clear(m.state.eventCyclesInFlight)
	clear(m.state.bindingsInFlight)
	clear(m.state.viewsInFlight)
	clear(m.state.watchLoopsInFlight)
	m.state.generation = m.Observer.currentGeneration()
	m.state.rawGeneration = m.Observer.currentRawGeneration()
	m.state.mu.Unlock()
}

func (m SharedSessionBrowserWatchManager) touch() {
	m.Observer.touchProvider()
	if m.state == nil {
		return
	}
	m.state.mu.Lock()
	m.state.lastActiveAt = time.Now()
	m.state.mu.Unlock()
}

func (m SharedSessionBrowserWatchManager) syncRawGeneration() {
	if m.state == nil {
		return
	}
	current := m.Observer.currentRawGeneration()
	m.state.mu.RLock()
	if m.state.rawGeneration == current {
		m.state.mu.RUnlock()
		return
	}
	m.state.mu.RUnlock()
	m.state.mu.Lock()
	if m.state.rawGeneration != current {
		clear(m.state.rawStatus)
		clear(m.state.rawProfiles)
		clear(m.state.rawRouteMutations)
		clear(m.state.routeMutations)
		clear(m.state.rawStatusInFlight)
		clear(m.state.rawProfilesInFlight)
		clear(m.state.rawStartsInFlight)
		clear(m.state.rawStopsInFlight)
		m.state.rawGeneration = current
	}
	m.state.mu.Unlock()
}

func (m SharedSessionBrowserWatchManager) syncProjectionGeneration() {
	if m.state == nil {
		return
	}
	current := m.Observer.currentGeneration()
	m.state.mu.RLock()
	if m.state.generation == current {
		m.state.mu.RUnlock()
		return
	}
	m.state.mu.RUnlock()
	m.state.mu.Lock()
	if m.state.generation != current {
		clear(m.state.eventCycles)
		clear(m.state.bindings)
		clear(m.state.views)
		clear(m.state.watchLoops)
		clear(m.state.eventCyclesInFlight)
		clear(m.state.bindingsInFlight)
		clear(m.state.viewsInFlight)
		clear(m.state.watchLoopsInFlight)
		m.state.generation = current
	}
	m.state.mu.Unlock()
}
