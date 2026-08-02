package browserruntime

import (
	"context"
	"strings"
	"time"
)

// ObserveRawStatus loads the raw RuntimeStatus source of truth used by
// higher-level observation and execution helpers.
func (m SharedSessionBrowserObserverManager) ObserveRawStatus(
	ctx context.Context,
	control BrowserRuntimeControlBackend,
	requestedProfile string,
) SharedSessionBrowserRawStatusObservation {
	return m.Bind(control).ObserveRawStatus(ctx, requestedProfile)
}

func (m SharedSessionBrowserObserverManager) observeRawStatusDirect(
	ctx context.Context,
	control BrowserRuntimeControlBackend,
	requestedProfile string,
) SharedSessionBrowserRawStatusObservation {
	observation := m.observeRawStatusSourceDirect(ctx, control, requestedProfile)
	if sharedSessionBrowserRawStatusObservationProvided(observation) {
		return observation
	}
	return m.observeRawStatusPollingDirect(ctx, control, requestedProfile)
}

func (m SharedSessionBrowserObserverManager) observeRawStatusSourceDirect(
	ctx context.Context,
	control BrowserRuntimeControlBackend,
	requestedProfile string,
) SharedSessionBrowserRawStatusObservation {
	observation := SharedSessionBrowserRawStatusObservation{
		RequestedProfile: strings.TrimSpace(requestedProfile),
	}
	if control == nil {
		return observation
	}
	if source, ok := control.(BrowserRuntimeRawStatusObservationBackend); ok {
		observation := normalizeSharedSessionBrowserRawStatusObservation(
			source.ObserveRawBrowserRuntimeStatus(ctx, observation.RequestedProfile),
			observation.RequestedProfile,
		)
		if sharedSessionBrowserRawStatusObservationProvided(observation) {
			return observation
		}
	}
	return observation
}

func (m SharedSessionBrowserObserverManager) observeRawStatusPollingDirect(
	ctx context.Context,
	control BrowserRuntimeControlBackend,
	requestedProfile string,
) SharedSessionBrowserRawStatusObservation {
	observation := SharedSessionBrowserRawStatusObservation{
		RequestedProfile: strings.TrimSpace(requestedProfile),
	}
	if control == nil {
		return observation
	}
	observation.ObservedAt = time.Now()
	result, err := control.RuntimeStatus(ctx, BrowserProfileStatusRequest{Profile: observation.RequestedProfile})
	if err != nil {
		observation.Err = err
		return observation
	}
	observation.Status = &result
	return observation
}

// ObserveRawProfiles loads the raw RuntimeProfiles source of truth used by
// higher-level observation and selection helpers.
func (m SharedSessionBrowserObserverManager) ObserveRawProfiles(
	ctx context.Context,
	control BrowserRuntimeControlBackend,
	requestedProfile string,
) SharedSessionBrowserRawProfilesObservation {
	return m.Bind(control).ObserveRawProfiles(ctx, requestedProfile)
}

func (m SharedSessionBrowserObserverManager) observeRawProfilesDirect(
	ctx context.Context,
	control BrowserRuntimeControlBackend,
	requestedProfile string,
) SharedSessionBrowserRawProfilesObservation {
	observation := m.observeRawProfilesSourceDirect(ctx, control, requestedProfile)
	if sharedSessionBrowserRawProfilesObservationProvided(observation) {
		return observation
	}
	return m.observeRawProfilesPollingDirect(ctx, control, requestedProfile)
}

func (m SharedSessionBrowserObserverManager) observeRawProfilesSourceDirect(
	ctx context.Context,
	control BrowserRuntimeControlBackend,
	requestedProfile string,
) SharedSessionBrowserRawProfilesObservation {
	observation := SharedSessionBrowserRawProfilesObservation{
		RequestedProfile: strings.TrimSpace(requestedProfile),
	}
	if control == nil {
		return observation
	}
	if source, ok := control.(BrowserRuntimeRawProfilesObservationBackend); ok {
		observation := normalizeSharedSessionBrowserRawProfilesObservation(
			source.ObserveRawBrowserRuntimeProfiles(ctx, observation.RequestedProfile),
			observation.RequestedProfile,
		)
		if sharedSessionBrowserRawProfilesObservationProvided(observation) {
			return observation
		}
	}
	return observation
}

func (m SharedSessionBrowserObserverManager) observeRawProfilesPollingDirect(
	ctx context.Context,
	control BrowserRuntimeControlBackend,
	requestedProfile string,
) SharedSessionBrowserRawProfilesObservation {
	observation := SharedSessionBrowserRawProfilesObservation{
		RequestedProfile: strings.TrimSpace(requestedProfile),
	}
	if control == nil {
		return observation
	}
	observation.ObservedAt = time.Now()
	result, err := control.RuntimeProfiles(ctx, BrowserProfilesRequest{Profile: observation.RequestedProfile})
	if err != nil {
		observation.Err = err
		return observation
	}
	observation.Profiles = &result
	return observation
}

// ObserveRawStart loads the raw RuntimeStart source of truth used by
// lifecycle execution helpers.
func (m SharedSessionBrowserObserverManager) ObserveRawStart(
	ctx context.Context,
	control BrowserRuntimeControlBackend,
	profile string,
) SharedSessionBrowserRawLifecycleObservation {
	return m.Bind(control).ObserveRawStart(ctx, profile)
}

func (m SharedSessionBrowserObserverManager) observeRawStartDirect(
	ctx context.Context,
	control BrowserRuntimeControlBackend,
	profile string,
) SharedSessionBrowserRawLifecycleObservation {
	observation := SharedSessionBrowserRawLifecycleObservation{
		Profile: strings.TrimSpace(profile),
	}
	if control == nil {
		return observation
	}
	if source, ok := control.(BrowserRuntimeRawLifecycleObservationBackend); ok {
		observation := normalizeSharedSessionBrowserRawLifecycleObservation(
			source.ObserveRawBrowserRuntimeStart(ctx, observation.Profile),
			observation.Profile,
		)
		if sharedSessionBrowserRawLifecycleObservationProvided(observation) {
			return observation
		}
	}
	observation.ObservedAt = time.Now()
	result, err := control.RuntimeStart(ctx, BrowserProfileLifecycleRequest{Profile: observation.Profile})
	if err != nil {
		observation.Err = err
		return observation
	}
	observation.Profile = firstNonEmptyString(strings.TrimSpace(result.Profile), observation.Profile)
	observation.Status = &result
	return observation
}

// ObserveRawStop loads the raw RuntimeStop source of truth used by lifecycle
// execution helpers.
func (m SharedSessionBrowserObserverManager) ObserveRawStop(
	ctx context.Context,
	control BrowserRuntimeControlBackend,
	profile string,
) SharedSessionBrowserRawLifecycleObservation {
	return m.Bind(control).ObserveRawStop(ctx, profile)
}

func (m SharedSessionBrowserObserverManager) observeRawStopDirect(
	ctx context.Context,
	control BrowserRuntimeControlBackend,
	profile string,
) SharedSessionBrowserRawLifecycleObservation {
	observation := SharedSessionBrowserRawLifecycleObservation{
		Profile: strings.TrimSpace(profile),
	}
	if control == nil {
		return observation
	}
	if source, ok := control.(BrowserRuntimeRawLifecycleObservationBackend); ok {
		observation := normalizeSharedSessionBrowserRawLifecycleObservation(
			source.ObserveRawBrowserRuntimeStop(ctx, observation.Profile),
			observation.Profile,
		)
		if sharedSessionBrowserRawLifecycleObservationProvided(observation) {
			return observation
		}
	}
	observation.ObservedAt = time.Now()
	result, err := control.RuntimeStop(ctx, BrowserProfileLifecycleRequest{Profile: observation.Profile})
	if err != nil {
		observation.Err = err
		return observation
	}
	observation.Profile = firstNonEmptyString(strings.TrimSpace(result.Profile), observation.Profile)
	observation.Status = &result
	return observation
}

// ObserveRawStatusAndProfiles loads a single raw polling cycle for optional
// RuntimeStatus and RuntimeProfiles sources.
func (m SharedSessionBrowserObserverManager) ObserveRawStatusAndProfiles(
	ctx context.Context,
	control BrowserRuntimeControlBackend,
	requestedProfile string,
	includeStatus bool,
	includeProfiles bool,
) SharedSessionBrowserRawStatusAndProfilesObservation {
	return m.Bind(control).ObserveRawStatusAndProfiles(ctx, requestedProfile, includeStatus, includeProfiles)
}

// ResolveProfileStatusEvent resolves a raw RuntimeStatus event through the
// shared session-state contract and returns the lifecycle-owned effective
// status together with any synced scoped state.
func (m SharedSessionBrowserObserverManager) ResolveProfileStatusEvent(
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	result BrowserProfileStatusResult,
	observedAt time.Time,
) (BrowserProfileStatusResult, SharedSessionBrowserProfileState, bool) {
	m.touchProvider()
	resolved := result
	synced := SharedSessionBrowserProfileStateFromStatus(selectedInfo, result)
	sessionID = strings.TrimSpace(sessionID)
	if m.StateRegistry == nil || sessionID == "" {
		return resolved, synced, false
	}
	if next, state, ok := m.StateRegistry.SyncSessionBrowserProfileStatusResolution(sessionID, selectedInfo, result, observedAt, m.ReconnectWindow); ok {
		return next, state, true
	}
	return resolved, synced, false
}

// ResolveProfilesEvent resolves a raw RuntimeProfiles event through the shared
// session-state contract and falls back to the raw scoped snapshot when no
// registry-backed synced snapshot is available.
func (m SharedSessionBrowserObserverManager) ResolveProfilesEvent(
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	requestedProfile string,
	result BrowserProfilesResult,
	observedAt time.Time,
) []SharedSessionBrowserProfileState {
	m.touchProvider()
	snapshot := SharedSessionBrowserProfileStatesFromObservedProfiles(selectedInfo, result, observedAt)
	sessionID = strings.TrimSpace(sessionID)
	if m.StateRegistry == nil || sessionID == "" {
		return snapshot
	}
	if synced := m.StateRegistry.SyncSessionBrowserProfilesResolution(sessionID, selectedInfo, requestedProfile, result, observedAt, m.ReconnectWindow); len(synced) > 0 {
		return synced
	}
	return snapshot
}

// ResolveStatusAndProfilesEvent resolves a raw combined status/profiles watch
// cycle through the shared session-state contract and returns the
// lifecycle-owned effective status together with the scoped synced snapshot.
func (m SharedSessionBrowserObserverManager) ResolveStatusAndProfilesEvent(
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	requestedProfile string,
	status *BrowserProfileStatusResult,
	statusObservedAt time.Time,
	profiles *BrowserProfilesResult,
	profilesObservedAt time.Time,
) (BrowserProfileStatusResult, SharedSessionBrowserProfileState, bool, []SharedSessionBrowserProfileState) {
	m.touchProvider()
	var resolved BrowserProfileStatusResult
	if status != nil {
		resolved = *status
	}
	var snapshot []SharedSessionBrowserProfileState
	if profiles != nil {
		snapshot = SharedSessionBrowserProfileStatesFromObservedProfiles(selectedInfo, *profiles, profilesObservedAt)
	}
	sessionID = strings.TrimSpace(sessionID)
	if m.StateRegistry == nil || sessionID == "" || (status == nil && profiles == nil) {
		return resolved, SharedSessionBrowserProfileState{}, false, snapshot
	}
	nextResolved, synced, ok, syncedSnapshot := m.StateRegistry.SyncSessionBrowserStatusAndProfilesResolution(
		sessionID,
		selectedInfo,
		requestedProfile,
		status,
		statusObservedAt,
		profiles,
		profilesObservedAt,
		m.ReconnectWindow,
	)
	if status != nil {
		resolved = nextResolved
	}
	return resolved, synced, ok, syncedSnapshot
}

// ResolveExecutionEvent resolves a lifecycle-owned execution result through
// the shared session-state contract and falls back to the raw execution
// observation when no registry-backed synced state is available.
func (m SharedSessionBrowserObserverManager) ResolveExecutionEvent(
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	requestedProfile string,
	result SharedSessionBrowserExecutionResult,
) SharedSessionBrowserExecutionResolution {
	m.touchProvider()
	resolution := SharedSessionBrowserExecutionResolution{}
	sessionID = strings.TrimSpace(sessionID)
	if m.StateRegistry != nil && sessionID != "" {
		resolution.ResolvedStatus, resolution.SyncedState, resolution.HasSyncedState, resolution.Snapshot = m.StateRegistry.SyncSessionBrowserExecutionResolution(
			sessionID,
			selectedInfo,
			requestedProfile,
			result.Profile,
			result.Profiles,
			result.ProfilesObservedAt,
			result.ProfileStatus,
			result.ProfileStatusObservedAt,
			result.Decision,
			m.ReconnectWindow,
		)
		return resolution
	}

	if result.Profiles != nil {
		resolution.Snapshot = SharedSessionBrowserProfileStatesFromObservedProfiles(selectedInfo, *result.Profiles, result.ProfilesObservedAt)
	}
	profile := firstNonEmptyString(
		strings.TrimSpace(result.Profile),
		strings.TrimSpace(result.ProfileStatus.Profile),
		strings.TrimSpace(requestedProfile),
		strings.TrimSpace(selectedInfo.Profile),
	)
	if strings.TrimSpace(result.Decision) != "" {
		if state, ok := SharedSessionBrowserProfileStateFromObservedLifecycle(selectedInfo, profile, result.ProfileStatus, result.Decision, result.ProfileStatusObservedAt); ok {
			resolution.SyncedState = state
			resolution.HasSyncedState = true
			resolution.ResolvedStatus = SharedSessionBrowserProfileStatusResultFromState(state, selectedInfo, profile)
			return resolution
		}
	}
	if !sharedSessionBrowserProfileStatusResultEmpty(result.ProfileStatus) {
		state := SharedSessionBrowserProfileStateFromObservedStatus(selectedInfo, result.ProfileStatus, result.ProfileStatusObservedAt)
		if strings.TrimSpace(state.Profile) != "" || strings.TrimSpace(state.Backend) != "" {
			resolution.SyncedState = state
			resolution.HasSyncedState = true
			resolution.ResolvedStatus = SharedSessionBrowserProfileStatusResultFromState(state, selectedInfo, profile)
			return resolution
		}
		resolution.ResolvedStatus = result.ProfileStatus
	}
	return resolution
}

// SyncProfileStatusEvent applies a single route-scoped RuntimeStatus
// observation to the shared session state contract and performs managed
// current-target invalidation when the resulting lifecycle state is no longer
// safe to reuse.
func (m SharedSessionBrowserObserverManager) SyncProfileStatusEvent(
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	result BrowserProfileStatusResult,
	observedAt time.Time,
) (SharedSessionBrowserProfileState, bool) {
	m.touchProvider()
	_, state, _ := m.ResolveProfileStatusEvent(sessionID, selectedInfo, result, observedAt)
	sessionID = strings.TrimSpace(sessionID)
	if sessionID != "" {
		requestedProfiles := sharedSessionBrowserRawObservationCacheKeys("", selectedInfo.Profile, result.Profile)
		observation := normalizeSharedSessionBrowserRawStatusObservation(SharedSessionBrowserRawStatusObservation{
			RequestedProfile: strings.TrimSpace(result.Profile),
			Status:           &result,
			ObservedAt:       observedAt,
		}, strings.TrimSpace(result.Profile))
		m.seedBoundManagersForRawStatus(
			sessionID,
			selectedInfo,
			requestedProfiles,
			observation,
		)
		seedRelatedSharedSessionBrowserObserverManagersRawObservations(
			m,
			m.SessionRegistry,
			m.StateRegistry,
			sessionID,
			selectedInfo,
			requestedProfiles,
			&observation,
			nil,
		)
	}
	m.InvalidateCurrentTargetForProfileStateEvent(sessionID, state)
	return state, true
}

// SyncProfilesEvent applies a route-scoped RuntimeProfiles observation to the
// shared session state contract and returns the final scoped lifecycle
// snapshot.
func (m SharedSessionBrowserObserverManager) SyncProfilesEvent(
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	requestedProfile string,
	result BrowserProfilesResult,
	observedAt time.Time,
) []SharedSessionBrowserProfileState {
	m.touchProvider()
	sessionID = strings.TrimSpace(sessionID)
	if m.StateRegistry == nil || sessionID == "" {
		return nil
	}
	snapshot := m.StateRegistry.SyncSessionBrowserProfilesResolution(sessionID, selectedInfo, requestedProfile, result, observedAt, m.ReconnectWindow)
	requestedProfiles := sharedSessionBrowserRawObservationCacheKeys(requestedProfile, selectedInfo.Profile, result.DefaultProfile)
	observation := normalizeSharedSessionBrowserRawProfilesObservation(SharedSessionBrowserRawProfilesObservation{
		RequestedProfile: strings.TrimSpace(requestedProfile),
		Profiles:         &result,
		ObservedAt:       observedAt,
	}, strings.TrimSpace(requestedProfile))
	m.seedBoundManagersForRawProfiles(
		sessionID,
		selectedInfo,
		requestedProfiles,
		observation,
	)
	seedRelatedSharedSessionBrowserObserverManagersRawObservations(
		m,
		m.SessionRegistry,
		m.StateRegistry,
		sessionID,
		selectedInfo,
		requestedProfiles,
		nil,
		&observation,
	)
	return snapshot
}

// SyncStatusAndProfilesEvent applies a combined route-scoped RuntimeStatus and
// RuntimeProfiles observation to the shared session state contract, performs
// managed current-target invalidation when the lifecycle state is no longer
// safe to reuse, and seeds the bound watch-manager raw caches so the next
// event cycle can reuse the source-time event directly.
func (m SharedSessionBrowserObserverManager) SyncStatusAndProfilesEvent(
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	requestedProfile string,
	status *BrowserProfileStatusResult,
	statusObservedAt time.Time,
	profiles *BrowserProfilesResult,
	profilesObservedAt time.Time,
) (BrowserProfileStatusResult, SharedSessionBrowserProfileState, bool, []SharedSessionBrowserProfileState) {
	m.touchProvider()
	resolved, synced, ok, snapshot := m.ResolveStatusAndProfilesEvent(
		sessionID,
		selectedInfo,
		requestedProfile,
		status,
		statusObservedAt,
		profiles,
		profilesObservedAt,
	)
	sessionID = strings.TrimSpace(sessionID)
	if sessionID != "" {
		var rawStatus *SharedSessionBrowserRawStatusObservation
		if status != nil {
			normalized := normalizeSharedSessionBrowserRawStatusObservation(SharedSessionBrowserRawStatusObservation{
				RequestedProfile: strings.TrimSpace(requestedProfile),
				Status:           status,
				ObservedAt:       statusObservedAt,
			}, strings.TrimSpace(requestedProfile))
			rawStatus = &normalized
		}
		var rawProfiles *SharedSessionBrowserRawProfilesObservation
		if profiles != nil {
			normalized := normalizeSharedSessionBrowserRawProfilesObservation(SharedSessionBrowserRawProfilesObservation{
				RequestedProfile: strings.TrimSpace(requestedProfile),
				Profiles:         profiles,
				ObservedAt:       profilesObservedAt,
			}, strings.TrimSpace(requestedProfile))
			rawProfiles = &normalized
		}
		requestedProfiles := sharedSessionBrowserRawObservationCacheKeys(
			requestedProfile,
			selectedInfo.Profile,
			firstNonEmptyString(
				func() string {
					if status == nil {
						return ""
					}
					return status.Profile
				}(),
				func() string {
					if profiles == nil {
						return ""
					}
					return profiles.DefaultProfile
				}(),
			),
		)
		m.seedBoundManagersRawObservations(
			sessionID,
			selectedInfo,
			requestedProfiles,
			rawStatus,
			rawProfiles,
		)
		seedRelatedSharedSessionBrowserObserverManagersRawObservations(
			m,
			m.SessionRegistry,
			m.StateRegistry,
			sessionID,
			selectedInfo,
			requestedProfiles,
			rawStatus,
			rawProfiles,
		)
	}
	m.InvalidateCurrentTargetForProfileStateEvent(sessionID, synced)
	return resolved, synced, ok, snapshot
}

// SyncProfileLifecycleEvent applies a lifecycle-owned status/decision event to
// the shared session state contract and performs managed current-target
// invalidation when the resulting lifecycle state is no longer safe to reuse.
func (m SharedSessionBrowserObserverManager) SyncProfileLifecycleEvent(
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	profile string,
	result BrowserProfileStatusResult,
	decision string,
	observedAt time.Time,
) (SharedSessionBrowserProfileState, bool) {
	m.touchProvider()
	profile = firstNonEmptyString(
		strings.TrimSpace(profile),
		strings.TrimSpace(result.Profile),
		strings.TrimSpace(selectedInfo.Profile),
	)
	state, ok := SharedSessionBrowserProfileStateFromObservedLifecycle(selectedInfo, profile, result, decision, observedAt)
	sessionID = strings.TrimSpace(sessionID)
	if m.StateRegistry != nil && sessionID != "" {
		if _, synced, syncedOK := m.StateRegistry.SyncSessionBrowserProfileLifecycleResolution(sessionID, selectedInfo, profile, result, decision, observedAt, m.ReconnectWindow); syncedOK {
			state = synced
			ok = true
		}
	}
	if !ok {
		return SharedSessionBrowserProfileState{}, false
	}
	if sessionID != "" {
		requestedProfiles := sharedSessionBrowserRawObservationCacheKeys("", selectedInfo.Profile, profile, result.Profile)
		observation := normalizeSharedSessionBrowserRawStatusObservation(SharedSessionBrowserRawStatusObservation{
			RequestedProfile: strings.TrimSpace(profile),
			Status:           &result,
			ObservedAt:       observedAt,
		}, strings.TrimSpace(profile))
		m.seedBoundManagersForRawStatus(
			sessionID,
			selectedInfo,
			requestedProfiles,
			observation,
		)
		seedRelatedSharedSessionBrowserObserverManagersRawObservations(
			m,
			m.SessionRegistry,
			m.StateRegistry,
			sessionID,
			selectedInfo,
			requestedProfiles,
			&observation,
			nil,
		)
	}
	m.InvalidateCurrentTargetForProfileStateEvent(sessionID, state)
	return state, true
}

// ObserveEventCycle runs the shared source-time status/profiles cycle used by
// watch, inspection, and execution observers before any higher-level
// projection is applied.
func (m SharedSessionBrowserObserverManager) ObserveEventCycle(
	ctx context.Context,
	control BrowserRuntimeControlBackend,
	req SharedSessionBrowserObserverRequest,
) SharedSessionBrowserEventCycleObservation {
	return m.Bind(control).ObserveEventCycle(ctx, req)
}

// ObserveWatchLoop runs the explicit shared source-time watch loop for a
// scoped route/profile selection and projects event-cycle, observer, watch,
// and session-view state from the same manager-owned dependencies.
func (m SharedSessionBrowserObserverManager) ObserveWatchLoop(
	ctx context.Context,
	control BrowserRuntimeControlBackend,
	req SharedSessionBrowserObserverRequest,
) SharedSessionBrowserWatchLoopObservation {
	return m.Bind(control).ObserveWatchLoop(ctx, req)
}

// ObserveStatus runs the shared source-time status-only cycle for a scoped
// route/profile selection.
func (m SharedSessionBrowserObserverManager) ObserveStatus(
	ctx context.Context,
	control BrowserRuntimeControlBackend,
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	requestedProfile string,
) SharedSessionBrowserStatusObservation {
	return m.Bind(control).ObserveStatus(ctx, sessionID, selectedInfo, requestedProfile)
}

// ObserveStatusAndProfiles runs the shared source-time status/profiles cycle
// for a scoped route/profile selection.
func (m SharedSessionBrowserObserverManager) ObserveStatusAndProfiles(
	ctx context.Context,
	control BrowserRuntimeControlBackend,
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	requestedProfile string,
	includeStatus bool,
	includeProfiles bool,
) SharedSessionBrowserStatusAndProfilesObservation {
	return m.Bind(control).ObserveStatusAndProfiles(ctx, sessionID, selectedInfo, requestedProfile, includeStatus, includeProfiles)
}

// ObserveProfiles runs the shared source-time profiles watch cycle for a
// scoped route/profile selection.
func (m SharedSessionBrowserObserverManager) ObserveProfiles(
	ctx context.Context,
	control BrowserRuntimeControlBackend,
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	requestedProfile string,
) SharedSessionBrowserProfilesObservation {
	return m.Bind(control).ObserveProfiles(ctx, sessionID, selectedInfo, requestedProfile)
}

// ObserveExecutionStatus runs the shared source-time execution status cycle
// used by execution paths together with the lifecycle-owned resolved status
// they should consume on both success and fallback paths.
func (m SharedSessionBrowserObserverManager) ObserveExecutionStatus(
	ctx context.Context,
	control BrowserRuntimeControlBackend,
	req SharedSessionBrowserExecutionRequest,
	profile string,
	fallback BrowserProfileStatusResult,
) SharedSessionBrowserExecutionStatusObservation {
	return m.Bind(control).ObserveExecutionStatus(ctx, req, profile, fallback)
}

// ObserveExecutionStart loads a RuntimeStart observation for a managed profile
// and applies the shared lifecycle decision mapping expected by execution
// paths.
func (m SharedSessionBrowserObserverManager) ObserveExecutionStart(
	ctx context.Context,
	control BrowserRuntimeControlBackend,
	req SharedSessionBrowserExecutionRequest,
	profile string,
	decision string,
	fallback BrowserProfileStatusResult,
) SharedSessionBrowserExecutionLifecycleObservation {
	return m.Bind(control).ObserveExecutionStart(ctx, req, profile, decision, fallback)
}

// ObserveExecutionStop loads a RuntimeStop observation for a managed profile
// and applies the shared lifecycle decision mapping expected by execution
// paths.
func (m SharedSessionBrowserObserverManager) ObserveExecutionStop(
	ctx context.Context,
	control BrowserRuntimeControlBackend,
	req SharedSessionBrowserExecutionRequest,
	profile string,
	decision string,
	fallback BrowserProfileStatusResult,
) SharedSessionBrowserExecutionLifecycleObservation {
	return m.Bind(control).ObserveExecutionStop(ctx, req, profile, decision, fallback)
}

// ObserveObserver runs the shared source-time observer cycle for a scoped
// route/profile selection.
func (m SharedSessionBrowserObserverManager) ObserveObserver(
	ctx context.Context,
	control BrowserRuntimeControlBackend,
	req SharedSessionBrowserObserverRequest,
) SharedSessionBrowserObserverObservation {
	return m.Bind(control).ObserveObserver(ctx, req)
}

// ObserveWatch runs the shared source-time watch cycle for a scoped
// route/profile selection.
func (m SharedSessionBrowserObserverManager) ObserveWatch(
	ctx context.Context,
	control BrowserRuntimeControlBackend,
	req SharedSessionBrowserObserverRequest,
) SharedSessionBrowserWatchObservation {
	return m.Bind(control).ObserveWatch(ctx, req)
}

// ObserveView runs the shared source-time session-view cycle for a scoped
// route/profile selection.
func (m SharedSessionBrowserObserverManager) ObserveView(
	ctx context.Context,
	control BrowserRuntimeControlBackend,
	req SharedSessionBrowserObserverRequest,
) SharedSessionBrowserViewObservation {
	return m.Bind(control).ObserveView(ctx, req)
}

// ObserveBinding runs the shared source-time binding cycle for a scoped
// route/profile selection.
func (m SharedSessionBrowserObserverManager) ObserveBinding(
	ctx context.Context,
	control BrowserRuntimeControlBackend,
	req SharedSessionBrowserObserverRequest,
) SharedSessionBrowserBindingObservation {
	return m.Bind(control).ObserveBinding(ctx, req)
}

// ObserveInspection runs the shared source-time inspection cycle for a scoped
// route/profile selection.
func (m SharedSessionBrowserObserverManager) ObserveInspection(
	ctx context.Context,
	control BrowserRuntimeControlBackend,
	req SharedSessionBrowserObserverRequest,
) SharedSessionBrowserInspectionObservation {
	return m.Bind(control).ObserveInspection(ctx, req)
}

// ObserveInspectionAction lowers an action-scoped inspection request onto the
// shared observer/watch contract before binding it to the shared lifecycle
// manager.
func (m SharedSessionBrowserObserverManager) ObserveInspectionAction(
	ctx context.Context,
	control BrowserRuntimeControlBackend,
	req SharedSessionBrowserInspectionActionRequest,
) SharedSessionBrowserInspectionObservation {
	return m.Bind(control).ObserveInspectionAction(ctx, req)
}
