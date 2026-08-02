package browserruntime

import (
	"context"
	"strings"
)

func (m SharedSessionBrowserWatchManager) ObserveRawStatus(ctx context.Context, requestedProfile string) SharedSessionBrowserRawStatusObservation {
	m.touch()
	requestedProfile = strings.TrimSpace(requestedProfile)
	if cached, ok := m.cachedRawStatus(requestedProfile); ok {
		return cached
	}
	inFlight, leader := m.beginRawStatusInFlight(requestedProfile)
	if !leader {
		return m.awaitRawStatusInFlight(ctx, requestedProfile, inFlight)
	}
	observation := m.Observer.observeRawStatusSourceDirect(ctx, m.Control, requestedProfile)
	if !sharedSessionBrowserRawStatusObservationProvided(observation) {
		if combined, ok := m.observeRawStatusFromCombinedSource(ctx, requestedProfile); ok {
			observation = combined
		} else {
			observation = m.Observer.observeRawStatusPollingDirect(ctx, m.Control, requestedProfile)
		}
	}
	m.storeRawStatus(requestedProfile, observation)
	m.finishRawStatusInFlight(requestedProfile, inFlight, observation)
	return observation
}

func (m SharedSessionBrowserWatchManager) ObserveRawProfiles(ctx context.Context, requestedProfile string) SharedSessionBrowserRawProfilesObservation {
	m.touch()
	requestedProfile = strings.TrimSpace(requestedProfile)
	if cached, ok := m.cachedRawProfiles(requestedProfile); ok {
		return cached
	}
	inFlight, leader := m.beginRawProfilesInFlight(requestedProfile)
	if !leader {
		return m.awaitRawProfilesInFlight(ctx, requestedProfile, inFlight)
	}
	observation := m.Observer.observeRawProfilesSourceDirect(ctx, m.Control, requestedProfile)
	if !sharedSessionBrowserRawProfilesObservationProvided(observation) {
		if combined, ok := m.observeRawProfilesFromCombinedSource(ctx, requestedProfile); ok {
			observation = combined
		} else {
			observation = m.Observer.observeRawProfilesPollingDirect(ctx, m.Control, requestedProfile)
		}
	}
	m.storeRawProfiles(requestedProfile, observation)
	m.finishRawProfilesInFlight(requestedProfile, inFlight, observation)
	return observation
}

func (m SharedSessionBrowserWatchManager) observeRawStatusFromCombinedSource(ctx context.Context, requestedProfile string) (SharedSessionBrowserRawStatusObservation, bool) {
	observation := SharedSessionBrowserRawStatusObservation{
		RequestedProfile: strings.TrimSpace(requestedProfile),
	}
	if m.Control == nil {
		return observation, false
	}
	source, ok := m.Control.(BrowserRuntimeRawStatusAndProfilesObservationBackend)
	if !ok {
		return observation, false
	}
	combined := normalizeSharedSessionBrowserRawStatusAndProfilesObservation(
		source.ObserveRawBrowserRuntimeStatusAndProfiles(ctx, observation.RequestedProfile, true, true),
		observation.RequestedProfile,
		true,
		true,
	)
	if !sharedSessionBrowserRawStatusAndProfilesObservationProvided(combined) {
		return observation, false
	}
	if sharedSessionBrowserRawProfilesObservationProvided(SharedSessionBrowserRawProfilesObservation{
		RequestedProfile: combined.RequestedProfile,
		Profiles:         combined.Profiles,
		Err:              combined.ProfilesErr,
		ObservedAt:       combined.ProfilesObservedAt,
	}) {
		m.storeRawProfiles(combined.RequestedProfile, SharedSessionBrowserRawProfilesObservation{
			RequestedProfile: combined.RequestedProfile,
			Profiles:         combined.Profiles,
			Err:              combined.ProfilesErr,
			ObservedAt:       combined.ProfilesObservedAt,
		})
	}
	observation = normalizeSharedSessionBrowserRawStatusObservation(SharedSessionBrowserRawStatusObservation{
		RequestedProfile: combined.RequestedProfile,
		Status:           combined.Status,
		Err:              combined.StatusErr,
		ObservedAt:       combined.StatusObservedAt,
	}, observation.RequestedProfile)
	return observation, sharedSessionBrowserRawStatusObservationProvided(observation)
}

func (m SharedSessionBrowserWatchManager) observeRawProfilesFromCombinedSource(ctx context.Context, requestedProfile string) (SharedSessionBrowserRawProfilesObservation, bool) {
	observation := SharedSessionBrowserRawProfilesObservation{
		RequestedProfile: strings.TrimSpace(requestedProfile),
	}
	if m.Control == nil {
		return observation, false
	}
	source, ok := m.Control.(BrowserRuntimeRawStatusAndProfilesObservationBackend)
	if !ok {
		return observation, false
	}
	combined := normalizeSharedSessionBrowserRawStatusAndProfilesObservation(
		source.ObserveRawBrowserRuntimeStatusAndProfiles(ctx, observation.RequestedProfile, true, true),
		observation.RequestedProfile,
		true,
		true,
	)
	if !sharedSessionBrowserRawStatusAndProfilesObservationProvided(combined) {
		return observation, false
	}
	if sharedSessionBrowserRawStatusObservationProvided(SharedSessionBrowserRawStatusObservation{
		RequestedProfile: combined.RequestedProfile,
		Status:           combined.Status,
		Err:              combined.StatusErr,
		ObservedAt:       combined.StatusObservedAt,
	}) {
		m.storeRawStatus(combined.RequestedProfile, SharedSessionBrowserRawStatusObservation{
			RequestedProfile: combined.RequestedProfile,
			Status:           combined.Status,
			Err:              combined.StatusErr,
			ObservedAt:       combined.StatusObservedAt,
		})
	}
	observation = normalizeSharedSessionBrowserRawProfilesObservation(SharedSessionBrowserRawProfilesObservation{
		RequestedProfile: combined.RequestedProfile,
		Profiles:         combined.Profiles,
		Err:              combined.ProfilesErr,
		ObservedAt:       combined.ProfilesObservedAt,
	}, observation.RequestedProfile)
	return observation, sharedSessionBrowserRawProfilesObservationProvided(observation)
}

func (m SharedSessionBrowserWatchManager) ObserveRawStart(ctx context.Context, profile string) SharedSessionBrowserRawLifecycleObservation {
	m.touch()
	inFlight, leader := m.beginRawStartInFlight(profile)
	if !leader {
		return m.awaitRawLifecycleInFlight(ctx, profile, inFlight)
	}
	observation := m.Observer.observeRawStartDirect(ctx, m.Control, profile)
	m.finishRawStartInFlight(profile, inFlight, observation)
	m.invalidateEventCycleCache()
	return observation
}

func (m SharedSessionBrowserWatchManager) ObserveRawStop(ctx context.Context, profile string) SharedSessionBrowserRawLifecycleObservation {
	m.touch()
	inFlight, leader := m.beginRawStopInFlight(profile)
	if !leader {
		return m.awaitRawLifecycleInFlight(ctx, profile, inFlight)
	}
	observation := m.Observer.observeRawStopDirect(ctx, m.Control, profile)
	m.finishRawStopInFlight(profile, inFlight, observation)
	m.invalidateEventCycleCache()
	return observation
}

func (m SharedSessionBrowserWatchManager) ObserveRawStatusAndProfiles(ctx context.Context, requestedProfile string, includeStatus bool, includeProfiles bool) SharedSessionBrowserRawStatusAndProfilesObservation {
	m.touch()
	observation := SharedSessionBrowserRawStatusAndProfilesObservation{
		RequestedProfile: strings.TrimSpace(requestedProfile),
	}
	if m.Control == nil {
		return observation
	}
	if source, ok := m.Control.(BrowserRuntimeRawStatusAndProfilesObservationBackend); ok {
		observation = normalizeSharedSessionBrowserRawStatusAndProfilesObservation(
			source.ObserveRawBrowserRuntimeStatusAndProfiles(ctx, observation.RequestedProfile, includeStatus, includeProfiles),
			observation.RequestedProfile,
			includeStatus,
			includeProfiles,
		)
		if sharedSessionBrowserRawStatusAndProfilesObservationProvided(observation) {
			if includeStatus && (observation.Status != nil || observation.StatusErr != nil) {
				m.storeRawStatus(observation.RequestedProfile, SharedSessionBrowserRawStatusObservation{
					RequestedProfile: observation.RequestedProfile,
					Status:           observation.Status,
					Err:              observation.StatusErr,
					ObservedAt:       observation.StatusObservedAt,
				})
			}
			if includeProfiles && (observation.Profiles != nil || observation.ProfilesErr != nil) {
				m.storeRawProfiles(observation.RequestedProfile, SharedSessionBrowserRawProfilesObservation{
					RequestedProfile: observation.RequestedProfile,
					Profiles:         observation.Profiles,
					Err:              observation.ProfilesErr,
					ObservedAt:       observation.ProfilesObservedAt,
				})
			}
		}
	}
	if includeStatus && !sharedSessionBrowserRawStatusObservationProvided(SharedSessionBrowserRawStatusObservation{
		RequestedProfile: observation.RequestedProfile,
		Status:           observation.Status,
		Err:              observation.StatusErr,
		ObservedAt:       observation.StatusObservedAt,
	}) {
		statusObservation := m.observeRawStatusWithoutCombinedSource(ctx, observation.RequestedProfile)
		observation.Status = statusObservation.Status
		observation.StatusErr = statusObservation.Err
		observation.StatusObservedAt = statusObservation.ObservedAt
	}
	if includeProfiles && !sharedSessionBrowserRawProfilesObservationProvided(SharedSessionBrowserRawProfilesObservation{
		RequestedProfile: observation.RequestedProfile,
		Profiles:         observation.Profiles,
		Err:              observation.ProfilesErr,
		ObservedAt:       observation.ProfilesObservedAt,
	}) {
		profilesObservation := m.observeRawProfilesWithoutCombinedSource(ctx, observation.RequestedProfile)
		observation.Profiles = profilesObservation.Profiles
		observation.ProfilesErr = profilesObservation.Err
		observation.ProfilesObservedAt = profilesObservation.ObservedAt
	}
	return observation
}

func (m SharedSessionBrowserWatchManager) observeRawStatusWithoutCombinedSource(ctx context.Context, requestedProfile string) SharedSessionBrowserRawStatusObservation {
	if cached, ok := m.cachedRawStatus(requestedProfile); ok {
		return cached
	}
	observation := m.Observer.observeRawStatusSourceDirect(ctx, m.Control, requestedProfile)
	if !sharedSessionBrowserRawStatusObservationProvided(observation) {
		observation = m.Observer.observeRawStatusPollingDirect(ctx, m.Control, requestedProfile)
	}
	m.storeRawStatus(strings.TrimSpace(requestedProfile), observation)
	return observation
}

func (m SharedSessionBrowserWatchManager) observeRawProfilesWithoutCombinedSource(ctx context.Context, requestedProfile string) SharedSessionBrowserRawProfilesObservation {
	if cached, ok := m.cachedRawProfiles(requestedProfile); ok {
		return cached
	}
	observation := m.Observer.observeRawProfilesSourceDirect(ctx, m.Control, requestedProfile)
	if !sharedSessionBrowserRawProfilesObservationProvided(observation) {
		observation = m.Observer.observeRawProfilesPollingDirect(ctx, m.Control, requestedProfile)
	}
	m.storeRawProfiles(strings.TrimSpace(requestedProfile), observation)
	return observation
}

func (m SharedSessionBrowserWatchManager) ObserveEventCycle(ctx context.Context, req SharedSessionBrowserObserverRequest) SharedSessionBrowserEventCycleObservation {
	m.touch()
	req = normalizeSharedSessionBrowserEventCycleRequest(req)
	if cached, ok := m.cachedEventCycle(req); ok {
		return cached
	}
	inFlight, leader := m.beginEventCycleInFlight(req)
	if !leader {
		return m.awaitEventCycleInFlight(ctx, req, inFlight)
	}
	observation := SharedSessionBrowserEventCycleObservation{}
	if m.Control == nil {
		m.finishEventCycleInFlight(req, inFlight, observation)
		return observation
	}
	if routeMutationSource, ok := m.cachedRouteMutationSource(req.SessionID, req.SelectedInfo, req.RequestedProfile); ok {
		projected := projectSharedSessionBrowserEventCycleObservationForRequest(routeMutationSource, req)
		if sharedSessionBrowserSeededEventCycleObservationProvided(projected, req) {
			m.storeEventCycle(req, projected)
			m.finishEventCycleInFlight(req, inFlight, projected)
			return projected
		}
	}
	if projected, ok := m.observeEventCycleFromRawOpenSource(ctx, req); ok {
		m.storeEventCycle(req, projected)
		m.finishEventCycleInFlight(req, inFlight, projected)
		return projected
	}
	if projected, ok := m.observeEventCycleFromRawNavigationSource(ctx, req); ok {
		m.storeEventCycle(req, projected)
		m.finishEventCycleInFlight(req, inFlight, projected)
		return projected
	}
	if projected, ok := m.observeEventCycleFromRawTabsSource(ctx, req); ok {
		m.storeEventCycle(req, projected)
		m.finishEventCycleInFlight(req, inFlight, projected)
		return projected
	}
	if projected, ok := m.observeEventCycleFromRawTargetSource(ctx, req); ok {
		m.storeEventCycle(req, projected)
		m.finishEventCycleInFlight(req, inFlight, projected)
		return projected
	}
	if projected, ok := m.observeEventCycleFromBackendRawRouteMutationSource(ctx, req); ok {
		m.storeEventCycle(req, projected)
		m.finishEventCycleInFlight(req, inFlight, projected)
		return projected
	}
	if projected, ok := m.observeEventCycleFromRawRouteMutationSource(req); ok {
		m.storeEventCycle(req, projected)
		m.finishEventCycleInFlight(req, inFlight, projected)
		return projected
	}
	rawObservation := m.ObserveRawStatusAndProfiles(ctx, req.RequestedProfile, req.IncludeStatus, req.IncludeProfiles)
	req.RequestedProfile = rawObservation.RequestedProfile

	if rawObservation.Status != nil {
		observation.Observation.Status = rawObservation.Status
		observation.Observation.StatusObservedAt = rawObservation.StatusObservedAt
		observation.Observation.ResolvedStatus = *rawObservation.Status
	}
	observation.Observation.StatusErr = rawObservation.StatusErr

	if rawObservation.Profiles != nil {
		observation.Observation.Profiles = rawObservation.Profiles
		observation.Observation.ProfilesObservedAt = rawObservation.ProfilesObservedAt
		observation.Observation.Snapshot = SharedSessionBrowserProfileStatesFromObservedProfiles(
			req.SelectedInfo,
			*rawObservation.Profiles,
			rawObservation.ProfilesObservedAt,
		)
	}
	observation.Observation.ProfilesErr = rawObservation.ProfilesErr

	resolvedStatus, syncedState, syncedOK, snapshot := m.Observer.ResolveStatusAndProfilesEvent(
		req.SessionID,
		req.SelectedInfo,
		req.RequestedProfile,
		observation.Observation.Status,
		observation.Observation.StatusObservedAt,
		observation.Observation.Profiles,
		observation.Observation.ProfilesObservedAt,
	)
	if observation.Observation.Status != nil {
		observation.Observation.ResolvedStatus = resolvedStatus
	}
	observation.Observation.SyncedState = syncedState
	observation.Observation.HasSyncedState = syncedOK
	observation.Observation.Snapshot = snapshot
	observation.ReferenceTime = sharedSessionBrowserLatestEventCycleObservedAt(
		observation.Observation.StatusObservedAt,
		observation.Observation.ProfilesObservedAt,
	)
	m.storeEventCycle(req, observation)
	m.finishEventCycleInFlight(req, inFlight, observation)
	return observation
}

func (m SharedSessionBrowserWatchManager) observeEventCycleFromRawOpenSource(
	ctx context.Context,
	req SharedSessionBrowserObserverRequest,
) (SharedSessionBrowserEventCycleObservation, bool) {
	if m.Control == nil || m.Observer.SessionRegistry == nil || strings.TrimSpace(req.SessionID) == "" {
		return SharedSessionBrowserEventCycleObservation{}, false
	}
	source, ok := m.Control.(BrowserRuntimeRawOpenObservationBackend)
	if !ok {
		return SharedSessionBrowserEventCycleObservation{}, false
	}
	observation := normalizeSharedSessionBrowserRawOpenObservation(
		source.ObserveRawBrowserOpen(ctx, req.RequestedProfile),
		req.RequestedProfile,
	)
	if !sharedSessionBrowserRawOpenObservationProvided(observation) {
		return SharedSessionBrowserEventCycleObservation{}, false
	}
	route := normalizeBrowserSessionRoute(req.BindingRoute)
	if route == (BrowserSessionRoute{}) {
		route = normalizeBrowserSessionRoute(BrowserSessionRoute{
			Backend: req.SelectedInfo.Backend,
			Profile: req.SelectedInfo.Profile,
			Target:  req.SelectedInfo.Target,
		})
	}
	m.Observer.ApplyOpenResultEvent(
		SharedSessionBrowserOpenResultEventRequest{
			SessionID: req.SessionID,
			Route:     route,
			URL:       observation.URL,
			Title:     observation.Title,
			Source:    "runtime_open_source",
		},
	)
	routeMutationSource, ok := m.cachedRouteMutationSource(req.SessionID, req.SelectedInfo, req.RequestedProfile)
	if !ok {
		return SharedSessionBrowserEventCycleObservation{}, false
	}
	projected := projectSharedSessionBrowserEventCycleObservationForRequest(routeMutationSource, req)
	if !sharedSessionBrowserSeededEventCycleObservationProvided(projected, req) {
		return SharedSessionBrowserEventCycleObservation{}, false
	}
	return projected, true
}

func (m SharedSessionBrowserWatchManager) observeEventCycleFromRawNavigationSource(
	ctx context.Context,
	req SharedSessionBrowserObserverRequest,
) (SharedSessionBrowserEventCycleObservation, bool) {
	if m.Control == nil || m.Observer.SessionRegistry == nil || strings.TrimSpace(req.SessionID) == "" {
		return SharedSessionBrowserEventCycleObservation{}, false
	}
	source, ok := m.Control.(BrowserRuntimeRawNavigationObservationBackend)
	if !ok {
		return SharedSessionBrowserEventCycleObservation{}, false
	}
	observation := normalizeSharedSessionBrowserRawNavigationObservation(
		source.ObserveRawBrowserNavigation(ctx, req.RequestedProfile),
		req.RequestedProfile,
	)
	if !sharedSessionBrowserRawNavigationObservationProvided(observation) {
		return SharedSessionBrowserEventCycleObservation{}, false
	}
	route := normalizeBrowserSessionRoute(req.BindingRoute)
	if route == (BrowserSessionRoute{}) {
		route = normalizeBrowserSessionRoute(BrowserSessionRoute{
			Backend: req.SelectedInfo.Backend,
			Profile: req.SelectedInfo.Profile,
			Target:  req.SelectedInfo.Target,
		})
	}
	priorSelection := observation.PriorSelection
	if priorSelection == nil {
		priorSelection = SnapshotSharedSessionBrowserCurrentTargetSelection(m.Observer.SessionRegistry, req.SessionID, route)
	}
	m.Observer.ApplyNavigationResultEvent(
		SharedSessionBrowserNavigationResultEventRequest{
			SessionID:        req.SessionID,
			Route:            route,
			ExplicitTargetID: observation.ExplicitTargetID,
			TabIndex:         observation.TabIndex,
			RequestedURL:     observation.RequestedURL,
			FinalURL:         firstNonEmptyString(observation.FinalURL, observation.RequestedURL),
			Title:            observation.Title,
			Source:           "runtime_navigation_source",
			Force:            observation.Force,
			PriorSelection:   priorSelection,
			Note:             observation.Note,
		},
	)
	m.Observer.seedBoundManagersRouteMutationSource(req.SessionID, route)
	routeMutationSource, ok := m.cachedRouteMutationSource(req.SessionID, req.SelectedInfo, req.RequestedProfile)
	if !ok {
		return SharedSessionBrowserEventCycleObservation{}, false
	}
	projected := projectSharedSessionBrowserEventCycleObservationForRequest(routeMutationSource, req)
	if !sharedSessionBrowserSeededEventCycleObservationProvided(projected, req) {
		return SharedSessionBrowserEventCycleObservation{}, false
	}
	return projected, true
}

func (m SharedSessionBrowserWatchManager) observeEventCycleFromRawRouteMutationSource(
	req SharedSessionBrowserObserverRequest,
) (SharedSessionBrowserEventCycleObservation, bool) {
	observation, ok := m.cachedRawRouteMutationSource(req.SessionID, req.SelectedInfo, req.RequestedProfile)
	if !ok || !sharedSessionBrowserRawRouteMutationObservationProvided(observation) {
		return SharedSessionBrowserEventCycleObservation{}, false
	}
	return m.observeEventCycleFromRawRouteMutationObservation(req, observation)
}

func (m SharedSessionBrowserWatchManager) observeEventCycleFromBackendRawRouteMutationSource(
	ctx context.Context,
	req SharedSessionBrowserObserverRequest,
) (SharedSessionBrowserEventCycleObservation, bool) {
	if m.Control == nil {
		return SharedSessionBrowserEventCycleObservation{}, false
	}
	source, ok := m.Control.(BrowserRuntimeRawRouteMutationObservationBackend)
	if !ok {
		return SharedSessionBrowserEventCycleObservation{}, false
	}
	observation := normalizeSharedSessionBrowserRawRouteMutationObservation(
		source.ObserveRawBrowserRouteMutation(ctx, req.RequestedProfile),
		req.RequestedProfile,
	)
	if !sharedSessionBrowserRawRouteMutationObservationProvided(observation) {
		return SharedSessionBrowserEventCycleObservation{}, false
	}
	return m.observeEventCycleFromRawRouteMutationObservation(req, observation)
}

func (m SharedSessionBrowserWatchManager) observeEventCycleFromRawRouteMutationObservation(
	req SharedSessionBrowserObserverRequest,
	observation SharedSessionBrowserRawRouteMutationObservation,
) (SharedSessionBrowserEventCycleObservation, bool) {
	route := normalizeBrowserSessionRoute(observation.Route)
	if route == (BrowserSessionRoute{}) {
		route = normalizeBrowserSessionRoute(req.BindingRoute)
	}
	if route == (BrowserSessionRoute{}) {
		route = normalizeBrowserSessionRoute(BrowserSessionRoute{
			Backend: req.SelectedInfo.Backend,
			Profile: req.SelectedInfo.Profile,
			Target:  req.SelectedInfo.Target,
		})
	}
	switch observation.Kind {
	case "open_result":
		m.Observer.ApplyOpenResultEvent(
			SharedSessionBrowserOpenResultEventRequest{
				SessionID: req.SessionID,
				Route:     route,
				URL:       observation.URL,
				Title:     observation.Title,
				Source:    firstNonEmptyString(observation.Source, "runtime_route_mutation_source"),
			},
		)
		m.Observer.seedBoundManagersRouteMutationSource(req.SessionID, route)
	case "navigation_result":
		ApplySharedSessionBrowserNavigationResultEvent(
			m.Observer.SessionRegistry,
			m.Observer.RunRegistry,
			m.Observer.StateRegistry,
			SharedSessionBrowserNavigationResultEventRequest{
				SessionID:        req.SessionID,
				Route:            route,
				ExplicitTargetID: firstNonEmptyString(observation.TargetID, observation.ExplicitTargetID),
				TabIndex:         observation.TabIndex,
				RequestedURL:     observation.RequestedURL,
				FinalURL:         firstNonEmptyString(observation.FinalURL, observation.URL),
				Title:            observation.Title,
				Source:           firstNonEmptyString(observation.Source, "runtime_route_mutation_source"),
				Force:            observation.Force,
				PriorSelection:   observation.PriorSelection,
				Note:             observation.Note,
			},
			m.Observer.ReconnectWindow,
		)
		m.Observer.seedBoundManagersRouteMutationSource(req.SessionID, route)
	case "action_result":
		ApplySharedSessionBrowserActionResultEvent(
			m.Observer.SessionRegistry,
			m.Observer.RunRegistry,
			m.Observer.StateRegistry,
			SharedSessionBrowserActionResultEventRequest{
				SessionID:         req.SessionID,
				Route:             route,
				PreferredTargetID: observation.TargetID,
				TabIndex:          observation.TabIndex,
				TrackCurrent:      observation.SetCurrent,
				URL:               firstNonEmptyString(observation.FinalURL, observation.URL),
				Title:             observation.Title,
				Source:            firstNonEmptyString(observation.Source, "runtime_route_mutation_source"),
				SetCurrent:        observation.SetCurrent,
				ReviewDecision:    observation.Decision,
				ReviewReady:       observation.Ready,
				Note:              observation.Note,
			},
			m.Observer.ReconnectWindow,
		)
		m.Observer.seedBoundManagersRouteMutationSource(req.SessionID, route)
	case "page_action_result":
		ApplySharedSessionBrowserPageActionResultEvent(
			m.Observer.SessionRegistry,
			m.Observer.RunRegistry,
			m.Observer.StateRegistry,
			SharedSessionBrowserPageActionResultEventRequest{
				SessionID:         req.SessionID,
				Route:             route,
				PreferredTargetID: observation.TargetID,
				TabIndex:          observation.TabIndex,
				URL:               firstNonEmptyString(observation.FinalURL, observation.URL),
				Title:             observation.Title,
				Source:            firstNonEmptyString(observation.Source, "runtime_route_mutation_source"),
				Actor:             observation.Actor,
				Force:             observation.Force,
				Review:            observation.Review,
			},
			m.Observer.ReconnectWindow,
		)
		m.Observer.seedBoundManagersRouteMutationSource(req.SessionID, route)
	case "tabs_result":
		ApplySharedSessionBrowserTabsResultEvent(
			m.Observer.SessionRegistry,
			m.Observer.RunRegistry,
			m.Observer.StateRegistry,
			SharedSessionBrowserTabsResultEventRequest{
				SessionID:              req.SessionID,
				Route:                  route,
				Action:                 firstNonEmptyString(observation.Action, "list"),
				RequestedTabIndex:      observation.RequestedTabIndex,
				ActiveIndex:            observation.ActiveIndex,
				Tabs:                   observation.Tabs,
				ExplicitTargetID:       observation.ExplicitTargetID,
				PriorSelection:         observation.PriorSelection,
				PriorActiveTargetID:    observation.PriorActiveTargetID,
				PriorRequestedTargetID: observation.PriorRequestedTargetID,
				Force:                  observation.Force,
				RememberTarget:         observation.RememberTarget,
				Review:                 observation.Review,
				Actor:                  observation.Actor,
				Note:                   observation.Note,
			},
			m.Observer.ReconnectWindow,
		)
		m.Observer.seedBoundManagersRouteMutationSource(req.SessionID, route)
	case "sync_tabs":
		m.Observer.SyncTabsForRouteEvent(req.SessionID, route, observation.ActiveIndex, observation.Tabs)
	case "track_tab":
		m.Observer.TrackTabEvent(req.SessionID, route, observation.Tab, observation.SetCurrent)
	case "track_current":
		m.Observer.TrackCurrentTargetEvent(
			req.SessionID,
			route,
			observation.URL,
			observation.Title,
			firstNonEmptyString(observation.Source, "runtime_route_mutation_source"),
		)
	case "resolved_target":
		m.Observer.ApplyResolvedTargetEvent(
			SharedSessionBrowserResolvedTargetEventRequest{
				SessionID:        req.SessionID,
				Route:            route,
				ExplicitTargetID: firstNonEmptyString(observation.TargetID, observation.ExplicitTargetID),
				TabIndex:         observation.TabIndex,
				URL:              firstNonEmptyString(observation.FinalURL, observation.URL),
				Title:            observation.Title,
				Source:           firstNonEmptyString(observation.Source, "runtime_route_mutation_source"),
				PriorSelection:   observation.PriorSelection,
			},
		)
	case "forget_tab":
		m.Observer.ForgetTabForRouteEvent(req.SessionID, route, observation.TabIndex)
	case "pending_review":
		m.Observer.RecordPendingTargetReviewEvent(
			req.SessionID,
			route,
			observation.TargetID,
			observation.TabIndex,
			observation.FinalURL,
			observation.Title,
			observation.Decision,
			observation.Reason,
		)
	case "pending_popup_review":
		m.Observer.RecordPendingTargetPopupReviewEvent(
			req.SessionID,
			route,
			observation.Tab,
			observation.Decision,
			observation.Reason,
		)
	case "restore_current":
		m.Observer.RestoreCurrentTargetSelectionEvent(
			req.SessionID,
			route,
			observation.PriorSelection,
			firstNonEmptyString(observation.Source, "runtime_route_mutation_source"),
		)
	case "restore_pending_review":
		m.Observer.RestoreCurrentTargetSelectionForPendingReviewEvent(
			req.SessionID,
			route,
			observation.PriorSelection,
			observation.PendingTargetID,
			firstNonEmptyString(observation.Source, "runtime_route_mutation_source"),
		)
	case "resolved_pending_review":
		m.Observer.ApplyResolvedTargetEvent(
			SharedSessionBrowserResolvedTargetEventRequest{
				SessionID:             req.SessionID,
				Route:                 route,
				ExplicitTargetID:      observation.TargetID,
				TabIndex:              observation.TabIndex,
				URL:                   firstNonEmptyString(observation.FinalURL, observation.URL),
				Title:                 observation.Title,
				Source:                firstNonEmptyString(observation.Source, "runtime_route_mutation_source"),
				PendingReview:         true,
				PendingReviewDecision: observation.Decision,
				PendingReviewReason:   observation.Reason,
				PriorSelection:        observation.PriorSelection,
			},
		)
	case "remember_review":
		m.Observer.ApplyTabRememberReviewEvent(SharedSessionBrowserTabRememberReviewRequest{
			SessionID:           req.SessionID,
			Route:               route,
			Action:              observation.Action,
			Force:               observation.Force,
			RememberTarget:      observation.RememberTarget,
			CandidateTargetID:   observation.CandidateTargetID,
			RequestedTabIndex:   observation.RequestedTabIndex,
			ActiveIndex:         observation.ActiveIndex,
			PriorActiveTargetID: observation.PriorActiveTargetID,
			PriorSelection:      observation.PriorSelection,
			Tabs:                observation.Tabs,
		})
	default:
		return SharedSessionBrowserEventCycleObservation{}, false
	}
	routeMutationSource, ok := m.cachedRouteMutationSource(req.SessionID, req.SelectedInfo, req.RequestedProfile)
	if !ok {
		return SharedSessionBrowserEventCycleObservation{}, false
	}
	projected := projectSharedSessionBrowserEventCycleObservationForRequest(routeMutationSource, req)
	if !sharedSessionBrowserSeededEventCycleObservationProvided(projected, req) {
		return SharedSessionBrowserEventCycleObservation{}, false
	}
	return projected, true
}

func (m SharedSessionBrowserWatchManager) observeEventCycleFromRawTabsSource(
	ctx context.Context,
	req SharedSessionBrowserObserverRequest,
) (SharedSessionBrowserEventCycleObservation, bool) {
	if m.Control == nil || m.Observer.SessionRegistry == nil || strings.TrimSpace(req.SessionID) == "" {
		return SharedSessionBrowserEventCycleObservation{}, false
	}
	source, ok := m.Control.(BrowserRuntimeRawTabsObservationBackend)
	if !ok {
		return SharedSessionBrowserEventCycleObservation{}, false
	}
	observation := normalizeSharedSessionBrowserRawTabsObservation(
		source.ObserveRawBrowserTabs(ctx, req.RequestedProfile),
		req.RequestedProfile,
	)
	if !sharedSessionBrowserRawTabsObservationProvided(observation) {
		return SharedSessionBrowserEventCycleObservation{}, false
	}
	route := normalizeBrowserSessionRoute(req.BindingRoute)
	if route == (BrowserSessionRoute{}) {
		route = normalizeBrowserSessionRoute(BrowserSessionRoute{
			Backend: req.SelectedInfo.Backend,
			Profile: req.SelectedInfo.Profile,
			Target:  req.SelectedInfo.Target,
		})
	}
	applyTabsResult := func() (SharedSessionBrowserEventCycleObservation, bool) {
		ApplySharedSessionBrowserTabsResultEvent(
			m.Observer.SessionRegistry,
			m.Observer.RunRegistry,
			m.Observer.StateRegistry,
			SharedSessionBrowserTabsResultEventRequest{
				SessionID:              req.SessionID,
				Route:                  route,
				Action:                 firstNonEmptyString(observation.Action, "list"),
				RequestedTabIndex:      observation.RequestedTabIndex,
				Force:                  observation.Force,
				RememberTarget:         observation.RememberTarget,
				Actor:                  observation.Actor,
				ExplicitTargetID:       observation.ExplicitTargetID,
				PriorSelection:         observation.PriorSelection,
				PriorActiveTargetID:    observation.PriorActiveTargetID,
				PriorRequestedTargetID: observation.PriorRequestedTargetID,
				Review:                 observation.Review,
				Note:                   observation.Note,
				ActiveIndex:            observation.ActiveIndex,
				Tabs:                   observation.Tabs,
			},
			m.Observer.ReconnectWindow,
		)
		m.Observer.seedBoundManagersRouteMutationSource(req.SessionID, route)
		routeMutationSource, ok := m.cachedRouteMutationSource(req.SessionID, req.SelectedInfo, req.RequestedProfile)
		if !ok {
			return SharedSessionBrowserEventCycleObservation{}, false
		}
		projected := projectSharedSessionBrowserEventCycleObservationForRequest(routeMutationSource, req)
		if !sharedSessionBrowserSeededEventCycleObservationProvided(projected, req) {
			return SharedSessionBrowserEventCycleObservation{}, false
		}
		return projected, true
	}
	if projected, ok := applyTabsResult(); ok {
		return projected, true
	}
	m.Observer.SyncTabsForRouteEvent(req.SessionID, route, observation.ActiveIndex, observation.Tabs)
	routeMutationSource, ok := m.cachedRouteMutationSource(req.SessionID, req.SelectedInfo, req.RequestedProfile)
	if !ok {
		return SharedSessionBrowserEventCycleObservation{}, false
	}
	projected := projectSharedSessionBrowserEventCycleObservationForRequest(routeMutationSource, req)
	if !sharedSessionBrowserSeededEventCycleObservationProvided(projected, req) {
		return SharedSessionBrowserEventCycleObservation{}, false
	}
	return projected, true
}

func (m SharedSessionBrowserWatchManager) observeEventCycleFromRawTargetSource(
	ctx context.Context,
	req SharedSessionBrowserObserverRequest,
) (SharedSessionBrowserEventCycleObservation, bool) {
	if m.Control == nil || m.Observer.SessionRegistry == nil || strings.TrimSpace(req.SessionID) == "" {
		return SharedSessionBrowserEventCycleObservation{}, false
	}
	source, ok := m.Control.(BrowserRuntimeRawTargetObservationBackend)
	if !ok {
		return SharedSessionBrowserEventCycleObservation{}, false
	}
	observation := normalizeSharedSessionBrowserRawTargetObservation(
		source.ObserveRawBrowserTarget(ctx, req.RequestedProfile),
		req.RequestedProfile,
	)
	if !sharedSessionBrowserRawTargetObservationProvided(observation) {
		return SharedSessionBrowserEventCycleObservation{}, false
	}
	if observation.TabIndex <= 0 && !observation.TrackCurrent && !observation.SetCurrent {
		return SharedSessionBrowserEventCycleObservation{}, false
	}
	route := normalizeBrowserSessionRoute(req.BindingRoute)
	if route == (BrowserSessionRoute{}) {
		route = normalizeBrowserSessionRoute(BrowserSessionRoute{
			Backend: req.SelectedInfo.Backend,
			Profile: req.SelectedInfo.Profile,
			Target:  req.SelectedInfo.Target,
		})
	}
	if observation.Actor != "" ||
		observation.Force ||
		observation.Review.Review != nil ||
		observation.Review.Count > 0 ||
		observation.Review.PolicyState != "" ||
		observation.Review.PolicyReason != "" {
		ApplySharedSessionBrowserPageActionResultEvent(
			m.Observer.SessionRegistry,
			m.Observer.RunRegistry,
			m.Observer.StateRegistry,
			SharedSessionBrowserPageActionResultEventRequest{
				SessionID:         req.SessionID,
				Route:             route,
				PreferredTargetID: observation.PreferredTargetID,
				TabIndex:          observation.TabIndex,
				URL:               observation.URL,
				Title:             observation.Title,
				Source:            firstNonEmptyString(observation.Source, "runtime_target_source"),
				Actor:             observation.Actor,
				Force:             observation.Force,
				Review:            observation.Review,
			},
			m.Observer.ReconnectWindow,
		)
		m.Observer.seedBoundManagersRouteMutationSource(req.SessionID, route)
	} else if observation.PreferredTargetID != "" ||
		observation.ReviewDecision != "" ||
		observation.ReviewReady ||
		observation.Note != "" {
		ApplySharedSessionBrowserActionResultEvent(
			m.Observer.SessionRegistry,
			m.Observer.RunRegistry,
			m.Observer.StateRegistry,
			SharedSessionBrowserActionResultEventRequest{
				SessionID:         req.SessionID,
				Route:             route,
				PreferredTargetID: observation.PreferredTargetID,
				TabIndex:          observation.TabIndex,
				TrackCurrent:      observation.TrackCurrent,
				URL:               observation.URL,
				Title:             observation.Title,
				Source:            firstNonEmptyString(observation.Source, "runtime_target_source"),
				SetCurrent:        observation.SetCurrent,
				ReviewDecision:    observation.ReviewDecision,
				ReviewReady:       observation.ReviewReady,
				Note:              observation.Note,
			},
			m.Observer.ReconnectWindow,
		)
		m.Observer.seedBoundManagersRouteMutationSource(req.SessionID, route)
	} else {
		m.Observer.ApplyTargetEvent(
			SharedSessionBrowserTargetEventRequest{
				SessionID:  req.SessionID,
				Route:      route,
				TabIndex:   observation.TabIndex,
				URL:        observation.URL,
				Title:      observation.Title,
				Source:     firstNonEmptyString(observation.Source, "runtime_target_source"),
				SetCurrent: observation.SetCurrent,
			},
		)
	}
	routeMutationSource, ok := m.cachedRouteMutationSource(req.SessionID, req.SelectedInfo, req.RequestedProfile)
	if !ok {
		return SharedSessionBrowserEventCycleObservation{}, false
	}
	projected := projectSharedSessionBrowserEventCycleObservationForRequest(routeMutationSource, req)
	if !sharedSessionBrowserSeededEventCycleObservationProvided(projected, req) {
		return SharedSessionBrowserEventCycleObservation{}, false
	}
	return projected, true
}

func (m SharedSessionBrowserWatchManager) ObserveWatchLoop(ctx context.Context, req SharedSessionBrowserObserverRequest) SharedSessionBrowserWatchLoopObservation {
	m.touch()
	req = normalizeSharedSessionBrowserWatchLoopRequest(req)
	if cached, ok := m.cachedWatchLoop(req); ok {
		return cached
	}
	inFlight, leader := m.beginWatchLoopInFlight(req)
	if !leader {
		return m.awaitWatchLoopInFlight(ctx, req, inFlight)
	}
	cycle := m.ObserveEventCycle(ctx, req)
	observer, invalidated := observeSharedSessionBrowserObserverForScopeFromCycleWithInvalidation(
		req,
		cycle,
		m.Observer.SessionRegistry,
		m.Observer.RunRegistry,
		m.Observer.StateRegistry,
		m.Observer.ReconnectWindow,
	)
	if invalidated {
		m.seedSiblingProvidersForBindingInvalidation(req, observer.Observation)
	}
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
	loop := SharedSessionBrowserWatchLoopObservation{
		Cycle:         cycle,
		Observer:      observer,
		Watch:         watch,
		View:          view,
		ReferenceTime: observer.ReferenceTime,
	}
	m.storeWatchLoop(req, loop)
	m.finishWatchLoopInFlight(req, inFlight, loop)
	return loop
}

func (m SharedSessionBrowserWatchManager) ObserveStatus(ctx context.Context, sessionID string, selectedInfo BrowserRuntimeInfo, requestedProfile string) SharedSessionBrowserStatusObservation {
	m.touch()
	cycle := m.ObserveEventCycle(ctx, SharedSessionBrowserObserverRequest{
		SessionID:        sessionID,
		SelectedInfo:     selectedInfo,
		BindingRoute:     BrowserSessionRoute{Backend: selectedInfo.Backend, Profile: selectedInfo.Profile, Target: selectedInfo.Target},
		RequestedProfile: requestedProfile,
		IncludeStatus:    true,
	})
	return SharedSessionBrowserStatusObservation{
		Status:         cycle.Observation.Status,
		StatusErr:      cycle.Observation.StatusErr,
		ObservedAt:     cycle.Observation.StatusObservedAt,
		ResolvedStatus: cycle.Observation.ResolvedStatus,
		SyncedState:    cycle.Observation.SyncedState,
		HasSyncedState: cycle.Observation.HasSyncedState,
	}
}

func (m SharedSessionBrowserWatchManager) ObserveStatusAndProfiles(ctx context.Context, sessionID string, selectedInfo BrowserRuntimeInfo, requestedProfile string, includeStatus bool, includeProfiles bool) SharedSessionBrowserStatusAndProfilesObservation {
	m.touch()
	return m.ObserveEventCycle(ctx, SharedSessionBrowserObserverRequest{
		SessionID:        sessionID,
		SelectedInfo:     selectedInfo,
		BindingRoute:     BrowserSessionRoute{Backend: selectedInfo.Backend, Profile: selectedInfo.Profile, Target: selectedInfo.Target},
		RequestedProfile: requestedProfile,
		IncludeStatus:    includeStatus,
		IncludeProfiles:  includeProfiles,
	}).Observation
}

func (m SharedSessionBrowserWatchManager) ObserveProfiles(ctx context.Context, sessionID string, selectedInfo BrowserRuntimeInfo, requestedProfile string) SharedSessionBrowserProfilesObservation {
	m.touch()
	cycle := m.ObserveEventCycle(ctx, SharedSessionBrowserObserverRequest{
		SessionID:        sessionID,
		SelectedInfo:     selectedInfo,
		BindingRoute:     BrowserSessionRoute{Backend: selectedInfo.Backend, Profile: selectedInfo.Profile, Target: selectedInfo.Target},
		RequestedProfile: requestedProfile,
		IncludeProfiles:  true,
	})
	return sharedSessionBrowserProfilesObservationFromCycle(
		m.Observer.StateRegistry,
		sessionID,
		selectedInfo,
		cycle.Observation,
	)
}

func (m SharedSessionBrowserWatchManager) ObserveExecutionStatus(ctx context.Context, req SharedSessionBrowserExecutionRequest, profile string, fallback BrowserProfileStatusResult) SharedSessionBrowserExecutionStatusObservation {
	m.touch()
	observation := SharedSessionBrowserExecutionStatusObservation{
		ResolvedStatus: resolveSharedSessionBrowserExecutionStatus(req, profile, fallback),
	}
	cycle := m.ObserveEventCycle(ctx, SharedSessionBrowserObserverRequest{
		SelectedInfo:     req.SelectedInfo,
		BindingRoute:     BrowserSessionRoute{Backend: req.SelectedInfo.Backend, Profile: req.SelectedInfo.Profile, Target: req.SelectedInfo.Target},
		RequestedProfile: profile,
		IncludeStatus:    true,
	})
	if cycle.Observation.StatusErr != nil {
		observation.StatusErr = cycle.Observation.StatusErr
		return observation
	}
	observation.ObservedAt = cycle.ReferenceTime
	if cycle.Observation.Status != nil {
		observation.Status = *cycle.Observation.Status
		observation.HasStatus = true
		observation.ResolvedStatus = resolveSharedSessionBrowserExecutionStatus(req, profile, cycle.Observation.ResolvedStatus)
	}
	return observation
}

func (m SharedSessionBrowserWatchManager) ObserveExecutionStart(ctx context.Context, req SharedSessionBrowserExecutionRequest, profile string, decision string, fallback BrowserProfileStatusResult) SharedSessionBrowserExecutionLifecycleObservation {
	m.touch()
	observation := SharedSessionBrowserExecutionLifecycleObservation{
		Profile: strings.TrimSpace(profile),
		Status:  fallback,
	}
	priorProfiles := m.snapshotRawProfilesForExecutionLifecycle("", profile, fallback.Profile, req.SelectedInfo.Profile)
	rawObservation := m.ObserveRawStart(ctx, profile)
	if rawObservation.Err != nil {
		observation.Err = rawObservation.Err
		return observation
	}
	observation.ObservedAt = rawObservation.ObservedAt
	observation.Profile = firstNonEmptyString(strings.TrimSpace(rawObservation.Profile), observation.Profile)
	status := resolveSharedSessionBrowserExecutionStatus(req, observation.Profile, fallback)
	if rawObservation.Status != nil {
		status = *rawObservation.Status
	}
	observation.Status = SharedSessionBrowserLifecycleDecisionStatus(req.SelectedInfo, observation.Profile, status, decision)
	observation.Ready = SharedSessionBrowserLifecycleDecisionReady(req.SelectedInfo, observation.Profile, observation.Status, decision)
	m.seedRawStatusFromExecutionLifecycle(profile, observation)
	m.seedRawProfilesFromExecutionLifecycle(profile, observation, priorProfiles)
	return observation
}

func (m SharedSessionBrowserWatchManager) ObserveExecutionStop(ctx context.Context, req SharedSessionBrowserExecutionRequest, profile string, decision string, fallback BrowserProfileStatusResult) SharedSessionBrowserExecutionLifecycleObservation {
	m.touch()
	observation := SharedSessionBrowserExecutionLifecycleObservation{
		Profile: strings.TrimSpace(profile),
		Status:  fallback,
	}
	priorProfiles := m.snapshotRawProfilesForExecutionLifecycle("", profile, fallback.Profile, req.SelectedInfo.Profile)
	rawObservation := m.ObserveRawStop(ctx, profile)
	if rawObservation.Err != nil {
		observation.Err = rawObservation.Err
		return observation
	}
	observation.ObservedAt = rawObservation.ObservedAt
	observation.Profile = firstNonEmptyString(strings.TrimSpace(rawObservation.Profile), observation.Profile)
	status := resolveSharedSessionBrowserExecutionStatus(req, observation.Profile, fallback)
	if rawObservation.Status != nil {
		status = *rawObservation.Status
	}
	observation.Status = SharedSessionBrowserLifecycleDecisionStatus(req.SelectedInfo, observation.Profile, status, decision)
	observation.Ready = SharedSessionBrowserLifecycleDecisionReady(req.SelectedInfo, observation.Profile, observation.Status, decision)
	m.seedRawStatusFromExecutionLifecycle(profile, observation)
	m.seedRawProfilesFromExecutionLifecycle(profile, observation, priorProfiles)
	return observation
}

func (m SharedSessionBrowserWatchManager) ObserveObserver(ctx context.Context, req SharedSessionBrowserObserverRequest) SharedSessionBrowserObserverObservation {
	m.touch()
	return m.ObserveWatchLoop(ctx, req).Observer
}

func (m SharedSessionBrowserWatchManager) ObserveWatch(ctx context.Context, req SharedSessionBrowserObserverRequest) SharedSessionBrowserWatchObservation {
	m.touch()
	return m.ObserveWatchLoop(ctx, req).Watch
}

func (m SharedSessionBrowserWatchManager) ObserveView(ctx context.Context, req SharedSessionBrowserObserverRequest) SharedSessionBrowserViewObservation {
	m.touch()
	req = normalizeSharedSessionBrowserViewRequest(req)
	if cached, ok := m.cachedView(req); ok {
		return cached
	}
	inFlight, leader := m.beginViewInFlight(req)
	if !leader {
		return m.awaitViewInFlight(ctx, inFlight)
	}
	observation := observeSharedSessionBrowserViewForScopeFromBinding(
		req,
		m.ObserveBinding(ctx, req),
		m.Observer.SessionRegistry,
		m.Observer.RunRegistry,
		m.Observer.StateRegistry,
	)
	m.storeView(req, observation)
	m.finishViewInFlight(req, inFlight, observation)
	return observation
}

func (m SharedSessionBrowserWatchManager) ObserveBinding(ctx context.Context, req SharedSessionBrowserObserverRequest) SharedSessionBrowserBindingObservation {
	m.touch()
	req = normalizeSharedSessionBrowserBindingRequest(req)
	if cached, ok := m.cachedBinding(req); ok {
		return cached
	}
	inFlight, leader := m.beginBindingInFlight(req)
	if !leader {
		return m.awaitBindingInFlight(ctx, inFlight)
	}
	observation, invalidated := observeSharedSessionBrowserBindingForScopeFromCycleWithInvalidation(
		req,
		m.ObserveEventCycle(ctx, req),
		m.Observer.SessionRegistry,
		m.Observer.RunRegistry,
		m.Observer.StateRegistry,
		m.Observer.ReconnectWindow,
	)
	if invalidated {
		m.seedSiblingProvidersForBindingInvalidation(req, observation.Observation)
	}
	m.storeBinding(req, observation)
	m.finishBindingInFlight(req, inFlight, observation)
	return observation
}

func (m SharedSessionBrowserWatchManager) seedSiblingProvidersForBindingInvalidation(
	req SharedSessionBrowserObserverRequest,
	observation SharedSessionBrowserStatusAndProfilesObservation,
) {
	m.seedSiblingProvidersForObservedStatusAndProfiles(req, observation)
}

func (m SharedSessionBrowserWatchManager) seedSiblingProvidersForObservedStatusAndProfiles(
	req SharedSessionBrowserObserverRequest,
	observation SharedSessionBrowserStatusAndProfilesObservation,
) {
	sessionID := strings.TrimSpace(req.SessionID)
	if sessionID == "" {
		return
	}
	eventCycle := SharedSessionBrowserEventCycleObservation{Observation: observation}
	statusObservation := sharedSessionBrowserRawStatusObservationFromEventCycle(eventCycle, req.RequestedProfile)
	profilesObservation := sharedSessionBrowserRawProfilesObservationFromEventCycle(eventCycle, req.RequestedProfile)
	var rawStatus *SharedSessionBrowserRawStatusObservation
	if sharedSessionBrowserRawStatusObservationProvided(statusObservation) {
		rawStatus = &statusObservation
	}
	var rawProfiles *SharedSessionBrowserRawProfilesObservation
	if sharedSessionBrowserRawProfilesObservationProvided(profilesObservation) {
		rawProfiles = &profilesObservation
	}
	if rawStatus == nil && rawProfiles == nil {
		return
	}
	requestedProfiles := sharedSessionBrowserRawObservationCacheKeys(
		req.RequestedProfile,
		req.SelectedInfo.Profile,
		req.BindingRoute.Profile,
		statusObservation.RequestedProfile,
		profilesObservation.RequestedProfile,
	)
	if rawStatus != nil && rawStatus.Status != nil {
		requestedProfiles = sharedSessionBrowserRawObservationCacheKeys(
			append(requestedProfiles, rawStatus.Status.Profile)...,
		)
	}
	if rawProfiles != nil && rawProfiles.Profiles != nil {
		requestedProfiles = sharedSessionBrowserRawObservationCacheKeys(
			append(requestedProfiles, rawProfiles.Profiles.DefaultProfile)...,
		)
	}
	seedRelatedSharedSessionBrowserObserverManagersRawObservations(
		m.Observer,
		m.Observer.SessionRegistry,
		m.Observer.StateRegistry,
		sessionID,
		req.SelectedInfo,
		requestedProfiles,
		rawStatus,
		rawProfiles,
	)
}

func (m SharedSessionBrowserWatchManager) ObserveInspection(ctx context.Context, req SharedSessionBrowserObserverRequest) SharedSessionBrowserInspectionObservation {
	m.touch()
	return m.ObserveWatch(ctx, req)
}

// ObserveInspectionAction lowers an action-scoped inspection request onto the
// shared observer/watch contract before reusing the lifecycle-owned watch
// manager.
func (m SharedSessionBrowserWatchManager) ObserveInspectionAction(ctx context.Context, req SharedSessionBrowserInspectionActionRequest) SharedSessionBrowserInspectionObservation {
	m.touch()
	return m.ObserveInspection(ctx, BuildSharedSessionBrowserInspectionObserverRequest(req))
}

func normalizeSharedSessionBrowserEventCycleRequest(req SharedSessionBrowserObserverRequest) SharedSessionBrowserObserverRequest {
	req.SessionID = strings.TrimSpace(req.SessionID)
	req.SelectedInfo.Backend = strings.TrimSpace(req.SelectedInfo.Backend)
	req.SelectedInfo.Profile = strings.TrimSpace(req.SelectedInfo.Profile)
	req.SelectedInfo.Target = strings.TrimSpace(req.SelectedInfo.Target)
	req.BindingRoute = normalizeBrowserSessionRoute(req.BindingRoute)
	req.RequestedProfile = strings.TrimSpace(req.RequestedProfile)
	req.IncludeSessionView = false
	req.SessionViewInfo = BrowserRuntimeInfo{}
	req.SessionViewRouteFilter = BrowserSessionRoute{}
	req.SessionViewRequestedProfile = ""
	return req
}

func normalizeSharedSessionBrowserWatchLoopRequest(req SharedSessionBrowserObserverRequest) SharedSessionBrowserObserverRequest {
	req.SessionID = strings.TrimSpace(req.SessionID)
	req.SelectedInfo.Backend = strings.TrimSpace(req.SelectedInfo.Backend)
	req.SelectedInfo.Profile = strings.TrimSpace(req.SelectedInfo.Profile)
	req.SelectedInfo.Target = strings.TrimSpace(req.SelectedInfo.Target)
	req.BindingRoute = normalizeBrowserSessionRoute(req.BindingRoute)
	req.RequestedProfile = strings.TrimSpace(req.RequestedProfile)
	if !req.IncludeSessionView {
		req.SessionViewInfo = BrowserRuntimeInfo{}
		req.SessionViewRouteFilter = BrowserSessionRoute{}
		req.SessionViewRequestedProfile = ""
		return req
	}
	req.SessionViewInfo.Backend = strings.TrimSpace(req.SessionViewInfo.Backend)
	req.SessionViewInfo.Profile = strings.TrimSpace(req.SessionViewInfo.Profile)
	req.SessionViewInfo.Target = strings.TrimSpace(req.SessionViewInfo.Target)
	req.SessionViewRouteFilter = normalizeBrowserSessionRoute(req.SessionViewRouteFilter)
	req.SessionViewRequestedProfile = strings.TrimSpace(req.SessionViewRequestedProfile)
	return req
}

func normalizeSharedSessionBrowserBindingRequest(req SharedSessionBrowserObserverRequest) SharedSessionBrowserObserverRequest {
	return normalizeSharedSessionBrowserEventCycleRequest(req)
}

func normalizeSharedSessionBrowserViewRequest(req SharedSessionBrowserObserverRequest) SharedSessionBrowserObserverRequest {
	return normalizeSharedSessionBrowserWatchLoopRequest(req)
}
