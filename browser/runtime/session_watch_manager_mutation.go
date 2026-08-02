package browserruntime

import (
	"context"
	"strings"
)

func (m SharedSessionBrowserObserverManager) withSuppressedRegistryWatchInvalidation(fn func()) {
	if fn == nil {
		return
	}
	if m.SessionRegistry == nil {
		fn()
		return
	}
	m.SessionRegistry.withSuppressedWatchManagerInvalidation(fn)
}

func sharedSessionBrowserRouteFromProfileSelection(
	selection SharedSessionBrowserProfileSelection,
	fallback BrowserSessionRoute,
	fallbackBrowserApp string,
) BrowserSessionRoute {
	route := normalizeBrowserSessionRoute(fallback)
	if backend := strings.TrimSpace(selection.Backend); backend != "" {
		route.Backend = backend
	}
	if profile := strings.TrimSpace(selection.Profile); profile != "" {
		route.Profile = profile
	}
	if runtimeTarget := strings.TrimSpace(selection.RuntimeTarget); runtimeTarget != "" {
		route.Target = runtimeTarget
	}
	if browserApp := firstNonEmptyString(strings.TrimSpace(selection.BrowserApp), strings.TrimSpace(fallbackBrowserApp)); browserApp != "" {
		route.BrowserApp = browserApp
	}
	return normalizeBrowserSessionRoute(route)
}

func sharedSessionBrowserRouteFromTargetSelection(
	selection *BrowserSessionTargetSelection,
	fallback BrowserSessionRoute,
) BrowserSessionRoute {
	route := normalizeBrowserSessionRoute(fallback)
	if selection == nil {
		return route
	}
	if backend := strings.TrimSpace(selection.Backend); backend != "" {
		route.Backend = backend
	}
	if profile := strings.TrimSpace(selection.Profile); profile != "" {
		route.Profile = profile
	}
	if runtimeTarget := strings.TrimSpace(selection.RuntimeTarget); runtimeTarget != "" {
		route.Target = runtimeTarget
	}
	if browserApp := strings.TrimSpace(selection.BrowserApp); browserApp != "" {
		route.BrowserApp = browserApp
	}
	return normalizeBrowserSessionRoute(route)
}

func sharedSessionBrowserRouteFromSelections(
	fallback BrowserSessionRoute,
	profileSelection *SharedSessionBrowserProfileSelection,
	targetSelection *BrowserSessionTargetSelection,
) BrowserSessionRoute {
	route := normalizeBrowserSessionRoute(fallback)
	if profileSelection != nil {
		route = sharedSessionBrowserRouteFromProfileSelection(*profileSelection, route, "")
	}
	if targetSelection != nil {
		route = sharedSessionBrowserRouteFromTargetSelection(targetSelection, route)
	}
	return normalizeBrowserSessionRoute(route)
}

// ApplyTabRememberReviewEvent applies the shared popup-review guardrail for tab
// lifecycle actions through the observer manager so route-scoped current-target
// restore and pending popup-review writes reuse the same manager-owned
// mutation seam.
func (m SharedSessionBrowserObserverManager) ApplyTabRememberReviewEvent(
	req SharedSessionBrowserTabRememberReviewRequest,
) SharedSessionBrowserTabRememberReviewResult {
	m.touchProvider()
	if req.Registry == nil {
		req.Registry = m.SessionRegistry
	}
	req.SessionID = strings.TrimSpace(req.SessionID)
	req.Route = normalizeBrowserSessionRoute(req.Route)
	req.Action = strings.TrimSpace(req.Action)
	req.CandidateTargetID = strings.TrimSpace(req.CandidateTargetID)
	req.PriorActiveTargetID = strings.TrimSpace(req.PriorActiveTargetID)

	result := SharedSessionBrowserTabRememberReviewResult{}
	if req.Registry == nil || req.SessionID == "" {
		return result
	}
	var observation *SharedSessionBrowserRawRouteMutationObservation
	req.Registry.withSuppressedWatchManagerInvalidation(func() {
		refreshed := false

		if (req.Action == "list" || req.Action == "close") && req.ActiveIndex > 0 {
			activeReview := SharedSessionBrowserPendingTargetReviewStateForTarget(
				req.Registry,
				req.SessionID,
				req.Route,
				"",
				req.ActiveIndex,
			)
			if activeReview.Review != nil && !(req.RememberTarget && req.Force) {
				RestoreSharedSessionBrowserCurrentTargetSelection(
					req.Registry,
					req.SessionID,
					req.Route,
					req.PriorSelection,
					"popup_review_restore",
				)
				refreshed = true
			}
		}

		switch req.Action {
		case "list", "close":
			result.RememberTabIndex = req.ActiveIndex
		case "focus":
			if req.ActiveIndex > 0 {
				result.RememberTabIndex = req.ActiveIndex
			} else {
				result.RememberTabIndex = req.RequestedTabIndex
			}
		}
		if result.RememberTabIndex <= 0 {
			result.RememberTargetID = req.CandidateTargetID
		}

		popupReview := req.RememberTarget &&
			sharedSessionBrowserRememberTargetNeedsPopupReview(req.Action, result.RememberTabIndex, req.PriorActiveTargetID)
		if popupReview {
			activeTab := sharedSessionBrowserTabByIndex(req.Tabs, result.RememberTabIndex)
			result.Decision = SharedSessionBrowserRememberTargetPopupReviewDecision(req.Force)
			result.Ready = req.Force
			result.Note = SharedSessionBrowserRememberTargetPopupReviewReason(activeTab, req.Force)
			if !req.Force && req.Registry != nil {
				_, review := req.Registry.restoreCurrentTargetSelectionAndRecordPendingTargetPopupReview(
					req.SessionID,
					req.Route,
					req.PriorSelection,
					activeTab,
					result.Decision,
					result.Note,
					"popup_review_restore",
				)
				refreshed = review != nil
			}
		}
		if refreshed && m.SessionRegistry != nil && strings.TrimSpace(req.SessionID) != "" {
			rawObservation := SharedSessionBrowserRawRouteMutationObservation{
				RequestedProfile:    req.Route.Profile,
				Route:               req.Route,
				Kind:                "remember_review",
				Action:              req.Action,
				Force:               req.Force,
				RememberTarget:      req.RememberTarget,
				CandidateTargetID:   req.CandidateTargetID,
				RequestedTabIndex:   req.RequestedTabIndex,
				PriorActiveTargetID: req.PriorActiveTargetID,
				ActiveIndex:         req.ActiveIndex,
				PriorSelection:      req.PriorSelection,
				Tabs:                req.Tabs,
			}
			observation = &rawObservation
		}
	})
	if observation != nil && m.SessionRegistry != nil && strings.TrimSpace(req.SessionID) != "" {
		m.seedSiblingProvidersForRouteMutation(req.SessionID, req.Route)
		m.seedBoundManagersRawRouteMutationSource(req.SessionID, req.Route, *observation)
		seedRelatedSharedSessionBrowserObserverManagersRawRouteMutation(
			m,
			m.SessionRegistry,
			m.StateRegistry,
			req.SessionID,
			req.Route,
			*observation,
		)
	}
	return result
}

// SyncTabsForRouteEvent applies a route-scoped tab attach/detach event through
// the shared tracking contract and refreshes standalone manager projections so
// the next watch loop can reuse cached source-time observations.
func (m SharedSessionBrowserObserverManager) SyncTabsForRouteEvent(
	sessionID string,
	route BrowserSessionRoute,
	activeIndex int,
	tabs []BrowserTab,
) []BrowserTab {
	m.touchProvider()
	var tracked []BrowserTab
	m.withSuppressedRegistryWatchInvalidation(func() {
		tracked = SyncSharedSessionBrowserTabsForRoute(m.SessionRegistry, sessionID, route, activeIndex, tabs)
	})
	if m.SessionRegistry != nil && strings.TrimSpace(sessionID) != "" {
		observation := SharedSessionBrowserRawRouteMutationObservation{
			RequestedProfile: route.Profile,
			Route:            route,
			Kind:             "sync_tabs",
			ActiveIndex:      activeIndex,
			Tabs:             tabs,
		}
		m.seedSiblingProvidersForRouteMutation(sessionID, route)
		m.seedBoundManagersRawRouteMutationSource(sessionID, route, observation)
		seedRelatedSharedSessionBrowserObserverManagersRawRouteMutation(
			m,
			m.SessionRegistry,
			m.StateRegistry,
			sessionID,
			route,
			observation,
		)
	}
	return tracked
}

// TrackTabEvent applies a single route-scoped tab attach/update event through
// the shared tracking contract and refreshes standalone manager projections so
// the next watch loop can reuse cached source-time observations.
func (m SharedSessionBrowserObserverManager) TrackTabEvent(
	sessionID string,
	route BrowserSessionRoute,
	tab BrowserTab,
	setCurrent bool,
) BrowserSessionTarget {
	m.touchProvider()
	var tracked BrowserSessionTarget
	m.withSuppressedRegistryWatchInvalidation(func() {
		tracked = TrackSharedSessionBrowserTab(m.SessionRegistry, sessionID, route, tab, setCurrent)
	})
	if m.SessionRegistry != nil && strings.TrimSpace(sessionID) != "" && tab.Index > 0 {
		observation := SharedSessionBrowserRawRouteMutationObservation{
			RequestedProfile: route.Profile,
			Route:            route,
			Kind:             "track_tab",
			Tab:              tab,
			SetCurrent:       setCurrent,
		}
		m.seedSiblingProvidersForRouteMutation(sessionID, route)
		m.seedBoundManagersRawRouteMutationSource(sessionID, route, observation)
		seedRelatedSharedSessionBrowserObserverManagersRawRouteMutation(
			m,
			m.SessionRegistry,
			m.StateRegistry,
			sessionID,
			route,
			observation,
		)
	}
	return tracked
}

// TrackCurrentTargetEvent applies a route-scoped current-target attach/update
// event through the shared tracking contract and refreshes standalone manager
// projections so the next watch loop can reuse cached source-time observations.
func (m SharedSessionBrowserObserverManager) TrackCurrentTargetEvent(
	sessionID string,
	route BrowserSessionRoute,
	url string,
	title string,
	source string,
) BrowserSessionTarget {
	m.touchProvider()
	var tracked BrowserSessionTarget
	m.withSuppressedRegistryWatchInvalidation(func() {
		tracked = TrackSharedSessionBrowserCurrentTarget(m.SessionRegistry, sessionID, route, url, title, source)
	})
	if m.SessionRegistry != nil && strings.TrimSpace(sessionID) != "" {
		observation := SharedSessionBrowserRawRouteMutationObservation{
			RequestedProfile: route.Profile,
			Route:            route,
			Kind:             "track_current",
			URL:              url,
			Title:            title,
			Source:           source,
		}
		m.seedSiblingProvidersForRouteMutation(sessionID, route)
		m.seedBoundManagersRawRouteMutationSource(sessionID, route, observation)
		seedRelatedSharedSessionBrowserObserverManagersRawRouteMutation(
			m,
			m.SessionRegistry,
			m.StateRegistry,
			sessionID,
			route,
			observation,
		)
	}
	return tracked
}

// RestoreCurrentTargetSelectionEvent applies a route-scoped current-target
// restore event through the shared tracking contract and refreshes standalone
// manager projections so the next watch loop can reuse cached source-time
// observations.
func (m SharedSessionBrowserObserverManager) RestoreCurrentTargetSelectionEvent(
	sessionID string,
	route BrowserSessionRoute,
	snapshot *BrowserSessionTargetSelection,
	source string,
) *BrowserSessionTargetSelection {
	m.touchProvider()
	var selection *BrowserSessionTargetSelection
	m.withSuppressedRegistryWatchInvalidation(func() {
		selection = RestoreSharedSessionBrowserCurrentTargetSelection(
			m.SessionRegistry,
			sessionID,
			route,
			snapshot,
			source,
		)
	})
	if m.SessionRegistry != nil && strings.TrimSpace(sessionID) != "" {
		if snapshot != nil {
			observation := SharedSessionBrowserRawRouteMutationObservation{
				RequestedProfile: route.Profile,
				Route:            route,
				Kind:             "restore_current",
				PriorSelection:   snapshot,
				Source:           source,
			}
			m.seedSiblingProvidersForRouteMutation(sessionID, route)
			m.seedBoundManagersRawRouteMutationSource(sessionID, route, observation)
			seedRelatedSharedSessionBrowserObserverManagersRawRouteMutation(
				m,
				m.SessionRegistry,
				m.StateRegistry,
				sessionID,
				route,
				observation,
			)
			return selection
		}
		m.seedSiblingProvidersForRouteMutation(sessionID, route)
	}
	return selection
}

// RestoreCurrentTargetSelectionForPendingReviewEvent applies a route-scoped
// current-target restore event for popup/redirect review posture and refreshes
// standalone manager projections so the next watch loop can reuse cached
// source-time observations.
func (m SharedSessionBrowserObserverManager) RestoreCurrentTargetSelectionForPendingReviewEvent(
	sessionID string,
	route BrowserSessionRoute,
	snapshot *BrowserSessionTargetSelection,
	pendingTargetID string,
	source string,
) *BrowserSessionTargetSelection {
	m.touchProvider()
	var selection *BrowserSessionTargetSelection
	m.withSuppressedRegistryWatchInvalidation(func() {
		selection = RestoreSharedSessionBrowserCurrentTargetSelectionForPendingReview(
			m.SessionRegistry,
			sessionID,
			route,
			snapshot,
			pendingTargetID,
			source,
		)
	})
	if m.SessionRegistry != nil &&
		strings.TrimSpace(sessionID) != "" &&
		(snapshot == nil ||
			strings.TrimSpace(snapshot.ID) == "" ||
			!strings.EqualFold(strings.TrimSpace(snapshot.ID), strings.TrimSpace(pendingTargetID))) {
		if snapshot != nil || strings.TrimSpace(pendingTargetID) != "" {
			observation := SharedSessionBrowserRawRouteMutationObservation{
				RequestedProfile: route.Profile,
				Route:            route,
				Kind:             "restore_pending_review",
				PriorSelection:   snapshot,
				PendingTargetID:  pendingTargetID,
				Source:           source,
			}
			m.seedSiblingProvidersForRouteMutation(sessionID, route)
			m.seedBoundManagersRawRouteMutationSource(sessionID, route, observation)
			seedRelatedSharedSessionBrowserObserverManagersRawRouteMutation(
				m,
				m.SessionRegistry,
				m.StateRegistry,
				sessionID,
				route,
				observation,
			)
			return selection
		}
		m.seedSiblingProvidersForRouteMutation(sessionID, route)
	}
	return selection
}

// ForgetTabForRouteEvent applies a route-scoped tab detach/close event through
// the shared tracking contract and refreshes standalone manager projections so
// the next watch loop can reuse cached source-time observations.
func (m SharedSessionBrowserObserverManager) ForgetTabForRouteEvent(
	sessionID string,
	route BrowserSessionRoute,
	tabIndex int,
) {
	m.touchProvider()
	m.withSuppressedRegistryWatchInvalidation(func() {
		ForgetSharedSessionBrowserTabForRoute(m.SessionRegistry, sessionID, route, tabIndex)
	})
	if m.SessionRegistry != nil && strings.TrimSpace(sessionID) != "" && tabIndex > 0 {
		observation := SharedSessionBrowserRawRouteMutationObservation{
			RequestedProfile: route.Profile,
			Route:            route,
			Kind:             "forget_tab",
			TabIndex:         tabIndex,
		}
		m.seedSiblingProvidersForRouteMutation(sessionID, route)
		m.seedBoundManagersRawRouteMutationSource(sessionID, route, observation)
		seedRelatedSharedSessionBrowserObserverManagersRawRouteMutation(
			m,
			m.SessionRegistry,
			m.StateRegistry,
			sessionID,
			route,
			observation,
		)
	}
}

// RecordPendingTargetReviewEvent applies a route-scoped pending target review
// event through the shared review contract and refreshes standalone manager
// projections so the next watch loop can reuse cached source-time observations.
func (m SharedSessionBrowserObserverManager) RecordPendingTargetReviewEvent(
	sessionID string,
	route BrowserSessionRoute,
	targetID string,
	tabIndex int,
	finalURL string,
	title string,
	decision string,
	reason string,
) *BrowserSessionTargetReview {
	m.touchProvider()
	var review *BrowserSessionTargetReview
	m.withSuppressedRegistryWatchInvalidation(func() {
		review = RecordSharedSessionBrowserPendingTargetReview(
			m.SessionRegistry,
			sessionID,
			route,
			targetID,
			tabIndex,
			finalURL,
			title,
			decision,
			reason,
		)
	})
	if review != nil && m.SessionRegistry != nil && strings.TrimSpace(sessionID) != "" {
		observation := SharedSessionBrowserRawRouteMutationObservation{
			RequestedProfile: route.Profile,
			Route:            route,
			Kind:             "pending_review",
			TabIndex:         tabIndex,
			TargetID:         targetID,
			FinalURL:         finalURL,
			Title:            title,
			Decision:         decision,
			Reason:           reason,
		}
		m.seedSiblingProvidersForRouteMutation(sessionID, route)
		m.seedBoundManagersRawRouteMutationSource(sessionID, route, observation)
		seedRelatedSharedSessionBrowserObserverManagersRawRouteMutation(
			m,
			m.SessionRegistry,
			m.StateRegistry,
			sessionID,
			route,
			observation,
		)
	}
	return review
}

// RecordPendingTargetPopupReviewEvent applies a route-scoped popup review event
// through the shared review contract and refreshes standalone manager
// projections so the next watch loop can reuse cached source-time observations.
func (m SharedSessionBrowserObserverManager) RecordPendingTargetPopupReviewEvent(
	sessionID string,
	route BrowserSessionRoute,
	activeTab BrowserTab,
	decision string,
	reason string,
) *BrowserSessionTargetReview {
	m.touchProvider()
	var review *BrowserSessionTargetReview
	m.withSuppressedRegistryWatchInvalidation(func() {
		review = RecordSharedSessionBrowserPendingTargetPopupReview(
			m.SessionRegistry,
			sessionID,
			route,
			activeTab,
			decision,
			reason,
		)
	})
	if review != nil && m.SessionRegistry != nil && strings.TrimSpace(sessionID) != "" {
		observation := SharedSessionBrowserRawRouteMutationObservation{
			RequestedProfile: route.Profile,
			Route:            route,
			Kind:             "pending_popup_review",
			Tab:              activeTab,
			Decision:         decision,
			Reason:           reason,
		}
		m.seedSiblingProvidersForRouteMutation(sessionID, route)
		m.seedBoundManagersRawRouteMutationSource(sessionID, route, observation)
		seedRelatedSharedSessionBrowserObserverManagersRawRouteMutation(
			m,
			m.SessionRegistry,
			m.StateRegistry,
			sessionID,
			route,
			observation,
		)
	}
	return review
}

// InvalidateCurrentTargetForProfileStateEvent applies a lifecycle-owned
// current-target invalidation event through the shared tracking contract and
// refreshes standalone manager projections so the next watch loop can reuse
// freshly seeded source-time observations.
func (m SharedSessionBrowserObserverManager) InvalidateCurrentTargetForProfileStateEvent(
	sessionID string,
	state SharedSessionBrowserProfileState,
) bool {
	m.touchProvider()
	cleared := InvalidateSharedSessionBrowserCurrentTargetForProfileState(
		m.SessionRegistry,
		sessionID,
		state,
	)
	if cleared && m.SessionRegistry != nil && strings.TrimSpace(sessionID) != "" {
		m.invalidateStandaloneBoundManagers()
	}
	return cleared
}

// SelectProfile applies the shared session profile-selection contract through
// the bound observer manager. Shared providers rely on state-registry owned
// cache coherency; standalone managers still refresh local bound caches after a
// successful writeback.
func (m SharedSessionBrowserObserverManager) SelectProfile(
	ctx context.Context,
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	browserApp string,
	control BrowserRuntimeControlBackend,
	validateWithProfiles bool,
	source string,
) (SharedSessionBrowserProfileSelection, string, bool, error) {
	m.touchProvider()
	var (
		selection SharedSessionBrowserProfileSelection
		decision  string
		ok        bool
		err       error
	)
	m.withSuppressedRegistryWatchInvalidation(func() {
		selection, decision, ok, err = SelectSharedSessionBrowserProfile(
			ctx,
			m.StateRegistry,
			sessionID,
			selectedInfo,
			browserApp,
			control,
			validateWithProfiles,
			source,
		)
	})
	if err == nil && ok {
		if m.SessionRegistry != nil && strings.TrimSpace(sessionID) != "" {
			m.seedSiblingProvidersForRouteMutation(
				sessionID,
				sharedSessionBrowserRouteFromProfileSelection(
					selection,
					BrowserSessionRoute{
						Backend:    strings.TrimSpace(selectedInfo.Backend),
						Profile:    strings.TrimSpace(selectedInfo.Profile),
						Target:     strings.TrimSpace(selectedInfo.Target),
						BrowserApp: strings.TrimSpace(browserApp),
					},
					browserApp,
				),
			)
		}
	}
	return selection, decision, ok, err
}

// SyncOrClearCurrentTargetForProfileSelection applies the shared current-target
// sync/clear contract. Shared providers rely on session-registry owned cache
// coherency; standalone managers still refresh local bound caches when the
// route-scoped target selection changes.
func (m SharedSessionBrowserObserverManager) SyncOrClearCurrentTargetForProfileSelection(
	sessionID string,
	route BrowserSessionRoute,
	profileSelection *SharedSessionBrowserProfileSelection,
	source string,
) (*BrowserSessionTargetSelection, string, error) {
	m.touchProvider()
	var (
		targetSelection *BrowserSessionTargetSelection
		decision        string
		err             error
	)
	m.withSuppressedRegistryWatchInvalidation(func() {
		targetSelection, decision, err = SyncOrClearSharedSessionBrowserCurrentTargetForProfileSelection(
			m.SessionRegistry,
			sessionID,
			route,
			profileSelection,
			source,
		)
	})
	if err == nil && (targetSelection != nil || strings.TrimSpace(decision) != "") {
		if m.SessionRegistry != nil && strings.TrimSpace(sessionID) != "" {
			m.seedSiblingProvidersForRouteMutation(
				sessionID,
				sharedSessionBrowserRouteFromTargetSelection(targetSelection, route),
			)
		}
	}
	return targetSelection, decision, err
}

// SyncRouteSelection applies the shared profile/target route sync contract and
// refreshes standalone manager caches after a ready writeback. Shared
// providers rely on session/state registry owned coherency.
func (m SharedSessionBrowserObserverManager) SyncRouteSelection(
	ctx context.Context,
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	route BrowserSessionRoute,
	browserApp string,
	control BrowserRuntimeControlBackend,
	validateWithProfiles bool,
	source string,
) (SharedSessionBrowserSyncSelectionResult, error) {
	m.touchProvider()
	var (
		result SharedSessionBrowserSyncSelectionResult
		err    error
	)
	m.withSuppressedRegistryWatchInvalidation(func() {
		result, err = SyncSharedSessionBrowserRouteSelection(
			ctx,
			m.StateRegistry,
			m.SessionRegistry,
			sessionID,
			selectedInfo,
			route,
			browserApp,
			control,
			validateWithProfiles,
			source,
		)
	})
	if err == nil && result.Ready {
		if m.SessionRegistry != nil && strings.TrimSpace(sessionID) != "" {
			m.seedSiblingProvidersForRouteMutation(
				sessionID,
				sharedSessionBrowserRouteFromSelections(route, result.ProfileSelection, result.TargetSelection),
			)
		}
	}
	return result, err
}

// CoordinateRouteSync applies the target-first route sync contract through the
// observer manager. Shared providers rely on registry-owned coherency;
// standalone managers still refresh local bound caches after a ready writeback.
func (m SharedSessionBrowserObserverManager) CoordinateRouteSync(
	ctx context.Context,
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	route BrowserSessionRoute,
	browserApp string,
	control BrowserRuntimeControlBackend,
	validateWithProfiles bool,
) (SharedSessionBrowserSyncSelectionResult, error) {
	m.touchProvider()
	var (
		result SharedSessionBrowserSyncSelectionResult
		err    error
	)
	m.withSuppressedRegistryWatchInvalidation(func() {
		result, err = CoordinateSharedSessionBrowserRouteSync(
			ctx,
			m.StateRegistry,
			m.SessionRegistry,
			sessionID,
			selectedInfo,
			route,
			browserApp,
			control,
			validateWithProfiles,
		)
	})
	if err == nil && result.Ready {
		if m.SessionRegistry != nil && strings.TrimSpace(sessionID) != "" {
			m.seedSiblingProvidersForRouteMutation(
				sessionID,
				sharedSessionBrowserRouteFromSelections(route, result.ProfileSelection, result.TargetSelection),
			)
		}
	}
	return result, err
}

// SelectTarget applies the shared route-scoped target-selection contract and
// refreshes standalone manager caches after a ready writeback. Shared
// providers rely on session-registry owned coherency.
func (m SharedSessionBrowserObserverManager) SelectTarget(
	req SharedSessionBrowserSelectTargetRequest,
) (SharedSessionBrowserSelectTargetResult, error) {
	m.touchProvider()
	var (
		result SharedSessionBrowserSelectTargetResult
		err    error
	)
	m.withSuppressedRegistryWatchInvalidation(func() {
		result, err = SelectSharedSessionBrowserTarget(m.SessionRegistry, req)
	})
	if err == nil && result.Ready {
		if m.SessionRegistry != nil && strings.TrimSpace(req.SessionID) != "" {
			m.seedSiblingProvidersForRouteMutation(
				req.SessionID,
				sharedSessionBrowserRouteFromTargetSelection(result.Selection, req.Route),
			)
		}
	}
	return result, err
}

// RememberTarget applies the shared remember-target contract and invalidates
// standalone manager caches after a ready writeback. Shared providers rely on
// session-registry owned coherency.
func (m SharedSessionBrowserObserverManager) RememberTarget(
	req SharedSessionBrowserRememberTargetRequest,
) SharedSessionBrowserRememberTargetResult {
	m.touchProvider()
	var result SharedSessionBrowserRememberTargetResult
	m.withSuppressedRegistryWatchInvalidation(func() {
		result = RememberSharedSessionBrowserTarget(m.SessionRegistry, req)
	})
	if result.Ready {
		if m.SessionRegistry != nil && strings.TrimSpace(req.SessionID) != "" {
			m.seedSiblingProvidersForRouteMutation(
				req.SessionID,
				sharedSessionBrowserRouteFromTargetSelection(result.Selection, req.Route),
			)
		}
	}
	return result
}

// RememberProfileForRoute applies the shared remember-profile contract and
// refreshes standalone manager caches after a ready writeback. Shared
// providers rely on session/state registry owned coherency.
func (m SharedSessionBrowserObserverManager) RememberProfileForRoute(
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	route BrowserSessionRoute,
	profileStatus *BrowserProfileStatusResult,
	preparedProfile string,
	requestedProfile string,
	requestedBrowserApp string,
) SharedSessionBrowserRememberProfileResult {
	m.touchProvider()
	var result SharedSessionBrowserRememberProfileResult
	m.withSuppressedRegistryWatchInvalidation(func() {
		result = RememberSharedSessionBrowserProfileForRoute(
			m.StateRegistry,
			m.SessionRegistry,
			sessionID,
			selectedInfo,
			route,
			profileStatus,
			preparedProfile,
			requestedProfile,
			requestedBrowserApp,
		)
	})
	if result.Ready {
		if m.SessionRegistry != nil && strings.TrimSpace(sessionID) != "" {
			m.seedSiblingProvidersForRouteMutation(
				sessionID,
				sharedSessionBrowserRouteFromSelections(route, result.ProfileSelection, result.TargetSelection),
			)
		}
	}
	return result
}

// ExecuteClearProfile applies the shared clear-profile contract and invalidates
// standalone manager caches when route-scoped profile/target selection
// changes. Shared providers rely on session/state registry owned coherency.
func (m SharedSessionBrowserObserverManager) ExecuteClearProfile(req SharedSessionBrowserClearRequest) SharedSessionBrowserClearResult {
	m.touchProvider()
	var result SharedSessionBrowserClearResult
	m.withSuppressedRegistryWatchInvalidation(func() {
		result = ExecuteSharedSessionBrowserClearProfile(req)
	})
	if result.ClearedProfileSelection || result.ClearedTargetSelection {
		if m.SessionRegistry != nil && strings.TrimSpace(req.SessionID) != "" {
			m.seedSiblingProvidersForRouteMutation(req.SessionID, req.Route)
		}
	}
	return result
}

// ExecuteClearTarget applies the shared clear-target contract and invalidates
// standalone manager caches when route-scoped profile/target selection
// changes. Shared providers rely on session/state registry owned coherency.
func (m SharedSessionBrowserObserverManager) ExecuteClearTarget(req SharedSessionBrowserClearRequest) SharedSessionBrowserClearResult {
	m.touchProvider()
	var result SharedSessionBrowserClearResult
	m.withSuppressedRegistryWatchInvalidation(func() {
		result = ExecuteSharedSessionBrowserClearTarget(req)
	})
	if result.ClearedProfileSelection || result.ClearedTargetSelection {
		if m.SessionRegistry != nil && strings.TrimSpace(req.SessionID) != "" {
			m.seedSiblingProvidersForRouteMutation(req.SessionID, req.Route)
		}
	}
	return result
}

// ExecuteClearSession applies the shared clear-session contract and invalidates
// standalone manager caches when route-scoped session state changes. Shared
// providers rely on session/state registry owned coherency.
func (m SharedSessionBrowserObserverManager) ExecuteClearSession(req SharedSessionBrowserClearRequest) SharedSessionBrowserClearResult {
	m.touchProvider()
	var result SharedSessionBrowserClearResult
	m.withSuppressedRegistryWatchInvalidation(func() {
		result = ExecuteSharedSessionBrowserClearSession(req)
	})
	if result.ClearedProfileSelection || result.ClearedTargetSelection || result.ClearedSessionProfiles > 0 || result.ClearedSessionTargets > 0 {
		if m.SessionRegistry != nil && strings.TrimSpace(req.SessionID) != "" {
			m.seedSiblingProvidersForRouteMutation(req.SessionID, req.Route)
		}
	}
	return result
}
