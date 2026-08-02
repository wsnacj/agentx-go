package browserruntime

import (
	"context"
	"strings"
	"time"
)

// ResolveSharedSessionBrowserProfileStatusEvent resolves a raw RuntimeStatus
// event through the shared session-state contract and returns the lifecycle-
// owned effective status together with any synced scoped state.
func ResolveSharedSessionBrowserProfileStatusEvent(
	registry SharedSessionBrowserStateRegistry,
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	result BrowserProfileStatusResult,
	observedAt time.Time,
	reconnectWindow time.Duration,
) (BrowserProfileStatusResult, SharedSessionBrowserProfileState, bool) {
	return sharedSessionBrowserObserverManager(
		nil,
		nil,
		registry,
		reconnectWindow,
	).ResolveProfileStatusEvent(sessionID, selectedInfo, result, observedAt)
}

// ResolveSharedSessionBrowserProfilesEvent resolves a raw RuntimeProfiles event
// through the shared session-state contract and falls back to the raw scoped
// snapshot when no registry-backed synced snapshot is available.
func ResolveSharedSessionBrowserProfilesEvent(
	registry SharedSessionBrowserStateRegistry,
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	requestedProfile string,
	result BrowserProfilesResult,
	observedAt time.Time,
	reconnectWindow time.Duration,
) []SharedSessionBrowserProfileState {
	return sharedSessionBrowserObserverManager(
		nil,
		nil,
		registry,
		reconnectWindow,
	).ResolveProfilesEvent(sessionID, selectedInfo, requestedProfile, result, observedAt)
}

// ResolveSharedSessionBrowserStatusAndProfilesEvent resolves a raw combined
// status/profiles watch cycle through the shared session-state contract and
// returns the lifecycle-owned effective status together with the scoped
// synced snapshot.
func ResolveSharedSessionBrowserStatusAndProfilesEvent(
	registry SharedSessionBrowserStateRegistry,
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	requestedProfile string,
	status *BrowserProfileStatusResult,
	statusObservedAt time.Time,
	profiles *BrowserProfilesResult,
	profilesObservedAt time.Time,
	reconnectWindow time.Duration,
) (BrowserProfileStatusResult, SharedSessionBrowserProfileState, bool, []SharedSessionBrowserProfileState) {
	return sharedSessionBrowserObserverManager(
		nil,
		nil,
		registry,
		reconnectWindow,
	).ResolveStatusAndProfilesEvent(
		sessionID,
		selectedInfo,
		requestedProfile,
		status,
		statusObservedAt,
		profiles,
		profilesObservedAt,
	)
}

// ResolveSharedSessionBrowserExecutionEvent resolves a lifecycle-owned
// execution result through the shared session-state contract and falls back to
// the raw execution observation when no registry-backed synced state is
// available.
func ResolveSharedSessionBrowserExecutionEvent(
	registry SharedSessionBrowserStateRegistry,
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	requestedProfile string,
	result SharedSessionBrowserExecutionResult,
	reconnectWindow time.Duration,
) SharedSessionBrowserExecutionResolution {
	return sharedSessionBrowserObserverManager(
		nil,
		nil,
		registry,
		reconnectWindow,
	).ResolveExecutionEvent(sessionID, selectedInfo, requestedProfile, result)
}

// ApplySharedSessionBrowserExecutionResultEvent applies a lifecycle-owned
// execution result through the shared observer-manager seam so primary and
// sibling providers can refresh from the same source-time writeback event.
func ApplySharedSessionBrowserExecutionResultEvent(
	sessionRegistry *BrowserSessionRegistry,
	runRegistry SharedSessionRunRegistry,
	stateRegistry SharedSessionBrowserStateRegistry,
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	requestedProfile string,
	result SharedSessionBrowserExecutionResult,
	reconnectWindow time.Duration,
) SharedSessionBrowserExecutionApplication {
	return sharedSessionBrowserObserverManager(
		sessionRegistry,
		runRegistry,
		stateRegistry,
		reconnectWindow,
	).ApplyExecutionResult(
		sessionID,
		selectedInfo,
		requestedProfile,
		result,
	)
}

// SharedSessionBrowserResolvedTargetEventRequest carries the route-scoped
// runtime result needed to track a resolved target and optionally gate it
// behind pending popup/redirect review.
type SharedSessionBrowserResolvedTargetEventRequest struct {
	SessionID             string
	Route                 BrowserSessionRoute
	ExplicitTargetID      string
	TabIndex              int
	URL                   string
	Title                 string
	Source                string
	PendingReview         bool
	PendingReviewDecision string
	PendingReviewReason   string
	PriorSelection        *BrowserSessionTargetSelection
}

// SharedSessionBrowserResolvedTargetEventResult captures the tracked target id
// together with any pending review and restored prior selection produced by a
// resolved-target event.
type SharedSessionBrowserResolvedTargetEventResult struct {
	TargetID          string
	Review            *BrowserSessionTargetReview
	RestoredSelection *BrowserSessionTargetSelection
}

// SharedSessionBrowserTargetEventRequest carries the generic route-scoped
// runtime result needed to track a tab/current target without additional
// popup-review or tab-sync semantics.
type SharedSessionBrowserTargetEventRequest struct {
	SessionID        string
	Route            BrowserSessionRoute
	ExplicitTargetID string
	TabIndex         int
	URL              string
	Title            string
	Source           string
	SetCurrent       bool
}

// SharedSessionBrowserTargetEventResult captures the tracked target handle
// produced by a generic route-scoped runtime result.
type SharedSessionBrowserTargetEventResult struct {
	Target   BrowserSessionTarget
	TargetID string
}

// SharedSessionBrowserTabsResultEventRequest carries the route-scoped runtime
// result from list/focus/close style tab actions so tab sync, target
// resolution, and remember-target popup review can share one top-level event
// contract.
type SharedSessionBrowserTabsResultEventRequest struct {
	SessionID              string
	Route                  BrowserSessionRoute
	Action                 string
	RequestedTabIndex      int
	ActiveIndex            int
	Tabs                   []BrowserTab
	ExplicitTargetID       string
	PriorSelection         *BrowserSessionTargetSelection
	PriorActiveTargetID    string
	PriorRequestedTargetID string
	Force                  bool
	RememberTarget         bool
	Review                 SharedSessionBrowserPendingTargetReviewState
	Actor                  string
	Note                   string
}

// SharedSessionBrowserTabsResultEventResult captures the tracked tab snapshot,
// resolved target handle, and remember-review posture produced by a tab action
// runtime result.
type SharedSessionBrowserTabsResultEventResult struct {
	Tabs           []BrowserTab
	TargetID       string
	RememberReview SharedSessionBrowserTabRememberReviewResult
	ReviewDecision string
	ReviewReady    bool
	Note           string
}

// ApplySharedSessionBrowserResolvedTargetEvent applies a runtime result through
// the shared route-scoped mutation seam: it tracks the resolved target and, if
// requested, records a pending popup/redirect review and restores the prior
// current-target selection through the same source-time contract.
func (m SharedSessionBrowserObserverManager) ApplyResolvedTargetEvent(
	req SharedSessionBrowserResolvedTargetEventRequest,
) SharedSessionBrowserResolvedTargetEventResult {
	req.SessionID = strings.TrimSpace(req.SessionID)
	req.Route = normalizeBrowserSessionRoute(req.Route)
	req.ExplicitTargetID = strings.TrimSpace(req.ExplicitTargetID)
	req.URL = strings.TrimSpace(req.URL)
	req.Title = strings.TrimSpace(req.Title)
	req.Source = strings.TrimSpace(req.Source)
	req.PendingReviewDecision = strings.TrimSpace(req.PendingReviewDecision)
	req.PendingReviewReason = strings.TrimSpace(req.PendingReviewReason)

	result := SharedSessionBrowserResolvedTargetEventResult{
		TargetID: req.ExplicitTargetID,
	}
	if m.SessionRegistry == nil || req.SessionID == "" {
		return result
	}

	if req.TabIndex > 0 {
		tracked := m.TrackTabEvent(
			req.SessionID,
			req.Route,
			BrowserTab{
				Index: req.TabIndex,
				URL:   req.URL,
				Title: req.Title,
			},
			true,
		)
		result.TargetID = firstNonEmptyString(result.TargetID, strings.TrimSpace(tracked.ID))
	} else {
		tracked := m.TrackCurrentTargetEvent(
			req.SessionID,
			req.Route,
			req.URL,
			req.Title,
			firstNonEmptyString(req.Source, "runtime_result"),
		)
		result.TargetID = firstNonEmptyString(result.TargetID, strings.TrimSpace(tracked.ID))
	}

	if req.PendingReview {
		result.Review = m.RecordPendingTargetReviewEvent(
			req.SessionID,
			req.Route,
			result.TargetID,
			req.TabIndex,
			req.URL,
			req.Title,
			req.PendingReviewDecision,
			req.PendingReviewReason,
		)
		result.RestoredSelection = m.RestoreCurrentTargetSelectionForPendingReviewEvent(
			req.SessionID,
			req.Route,
			req.PriorSelection,
			result.TargetID,
			"popup_review_restore",
		)
	}

	return result
}

// ApplySharedSessionBrowserResolvedTargetEvent applies a runtime result through
// the shared route-scoped mutation seam: it tracks the resolved target and, if
// requested, records a pending popup/redirect review and restores the prior
// current-target selection through the same source-time contract.
func ApplySharedSessionBrowserResolvedTargetEvent(
	sessionRegistry *BrowserSessionRegistry,
	runRegistry SharedSessionRunRegistry,
	stateRegistry SharedSessionBrowserStateRegistry,
	req SharedSessionBrowserResolvedTargetEventRequest,
	reconnectWindow time.Duration,
) SharedSessionBrowserResolvedTargetEventResult {
	return sharedSessionBrowserObserverManager(
		sessionRegistry,
		runRegistry,
		stateRegistry,
		reconnectWindow,
	).ApplyResolvedTargetEvent(req)
}

// ApplySharedSessionBrowserTargetEvent applies a generic route-scoped runtime
// result through the shared mutation seam so target tracking can share the same
// source-time contract as the richer resolved-target and tabs-result helpers.
func (m SharedSessionBrowserObserverManager) ApplyTargetEvent(
	req SharedSessionBrowserTargetEventRequest,
) SharedSessionBrowserTargetEventResult {
	req.SessionID = strings.TrimSpace(req.SessionID)
	req.Route = normalizeBrowserSessionRoute(req.Route)
	req.ExplicitTargetID = strings.TrimSpace(req.ExplicitTargetID)
	req.URL = strings.TrimSpace(req.URL)
	req.Title = strings.TrimSpace(req.Title)
	req.Source = strings.TrimSpace(req.Source)

	result := SharedSessionBrowserTargetEventResult{
		TargetID: req.ExplicitTargetID,
	}
	if m.SessionRegistry == nil || req.SessionID == "" {
		return result
	}
	if req.TabIndex > 0 {
		result.Target = m.TrackTabEvent(
			req.SessionID,
			req.Route,
			BrowserTab{
				Index: req.TabIndex,
				URL:   req.URL,
				Title: req.Title,
			},
			req.SetCurrent,
		)
	} else {
		result.Target = m.TrackCurrentTargetEvent(
			req.SessionID,
			req.Route,
			req.URL,
			req.Title,
			firstNonEmptyString(req.Source, "runtime_result"),
		)
	}
	result.TargetID = firstNonEmptyString(result.TargetID, strings.TrimSpace(result.Target.ID))
	return result
}

// ApplySharedSessionBrowserTargetEvent applies a generic route-scoped runtime
// result through the shared mutation seam so target tracking can share the same
// source-time contract as the richer resolved-target and tabs-result helpers.
func ApplySharedSessionBrowserTargetEvent(
	sessionRegistry *BrowserSessionRegistry,
	runRegistry SharedSessionRunRegistry,
	stateRegistry SharedSessionBrowserStateRegistry,
	req SharedSessionBrowserTargetEventRequest,
	reconnectWindow time.Duration,
) SharedSessionBrowserTargetEventResult {
	return sharedSessionBrowserObserverManager(
		sessionRegistry,
		runRegistry,
		stateRegistry,
		reconnectWindow,
	).ApplyTargetEvent(req)
}

// ApplySharedSessionBrowserTabsResultEvent applies a route-scoped tab action
// runtime result through the shared mutation seam: it syncs tracked tabs,
// preserves any pre-sync target handle needed by close/focus actions, resolves
// the effective target after sync, and then applies remember-target popup
// review through the same owner.
func ApplySharedSessionBrowserTabsResultEvent(
	sessionRegistry *BrowserSessionRegistry,
	runRegistry SharedSessionRunRegistry,
	stateRegistry SharedSessionBrowserStateRegistry,
	req SharedSessionBrowserTabsResultEventRequest,
	reconnectWindow time.Duration,
) SharedSessionBrowserTabsResultEventResult {
	req.SessionID = strings.TrimSpace(req.SessionID)
	req.Route = normalizeBrowserSessionRoute(req.Route)
	req.Action = strings.TrimSpace(req.Action)
	req.ExplicitTargetID = strings.TrimSpace(req.ExplicitTargetID)
	req.PriorActiveTargetID = strings.TrimSpace(req.PriorActiveTargetID)
	req.PriorRequestedTargetID = strings.TrimSpace(req.PriorRequestedTargetID)
	req.Actor = strings.TrimSpace(req.Actor)
	req.Note = strings.TrimSpace(req.Note)
	if req.PriorSelection != nil {
		priorSelectionCopy := *req.PriorSelection
		req.PriorSelection = &priorSelectionCopy
	}

	result := SharedSessionBrowserTabsResultEventResult{
		TargetID: req.ExplicitTargetID,
		Tabs:     append([]BrowserTab(nil), req.Tabs...),
	}
	if sessionRegistry == nil || req.SessionID == "" {
		return result
	}

	priorSelection := req.PriorSelection
	if priorSelection == nil {
		priorSelection = SnapshotSharedSessionBrowserCurrentTargetSelection(sessionRegistry, req.SessionID, req.Route)
	}
	priorActiveTargetID := req.PriorActiveTargetID
	if priorActiveTargetID == "" && req.ActiveIndex > 0 {
		priorActiveTargetID = ResolveSharedSessionBrowserTabTargetID(sessionRegistry, req.SessionID, req.Route, req.ActiveIndex)
	}
	priorRequestedTargetID := req.PriorRequestedTargetID
	if priorRequestedTargetID == "" && req.RequestedTabIndex > 0 {
		priorRequestedTargetID = ResolveSharedSessionBrowserTabTargetID(sessionRegistry, req.SessionID, req.Route, req.RequestedTabIndex)
	}

	result.Tabs = SyncSharedSessionBrowserTabsForRouteEvent(
		sessionRegistry,
		runRegistry,
		stateRegistry,
		req.SessionID,
		req.Route,
		req.ActiveIndex,
		req.Tabs,
		reconnectWindow,
	)

	result.TargetID = firstNonEmptyString(result.TargetID, priorRequestedTargetID)
	switch req.Action {
	case "list":
		if result.TargetID == "" && req.ActiveIndex > 0 {
			result.TargetID = ResolveSharedSessionBrowserTabTargetID(sessionRegistry, req.SessionID, req.Route, req.ActiveIndex)
		}
	case "focus":
		if result.TargetID == "" && req.RequestedTabIndex > 0 {
			result.TargetID = ResolveSharedSessionBrowserTabTargetID(sessionRegistry, req.SessionID, req.Route, req.RequestedTabIndex)
		}
		if result.TargetID == "" && req.ActiveIndex > 0 {
			result.TargetID = ResolveSharedSessionBrowserTabTargetID(sessionRegistry, req.SessionID, req.Route, req.ActiveIndex)
		}
	case "close":
		if result.TargetID == "" {
			result.TargetID = priorRequestedTargetID
		}
	}

	result.RememberReview = ApplySharedSessionBrowserTabRememberReviewEvent(
		sessionRegistry,
		runRegistry,
		stateRegistry,
		SharedSessionBrowserTabRememberReviewRequest{
			Registry:            sessionRegistry,
			RunRegistry:         runRegistry,
			StateRegistry:       stateRegistry,
			ReconnectWindow:     reconnectWindow,
			SessionID:           req.SessionID,
			Route:               req.Route,
			Action:              req.Action,
			Force:               req.Force,
			RememberTarget:      req.RememberTarget,
			CandidateTargetID:   result.TargetID,
			RequestedTabIndex:   req.RequestedTabIndex,
			ActiveIndex:         req.ActiveIndex,
			PriorActiveTargetID: priorActiveTargetID,
			PriorSelection:      priorSelection,
			Tabs:                result.Tabs,
		},
		reconnectWindow,
	)

	if req.Review.Review != nil {
		result.ReviewDecision = SharedSessionBrowserPendingTargetReviewDecision(req.Review, req.Force)
		result.ReviewReady = req.Force
		result.Note = firstNonEmptyString(
			req.Note,
			SharedSessionBrowserPendingTargetReviewReason(
				firstNonEmptyString(req.Actor, "browser tabs action"),
				req.Review,
				req.Force,
			),
		)
		return result
	}
	if strings.TrimSpace(result.RememberReview.Decision) != "" {
		result.Note = firstNonEmptyString(req.Note, strings.TrimSpace(result.RememberReview.Note))
		return result
	}
	result.Note = req.Note

	return result
}

// SyncSharedSessionBrowserProfileStatusEvent applies a single route-scoped
// RuntimeStatus observation to the shared session state contract and performs
// managed current-target invalidation when the resulting lifecycle state is no
// longer safe to reuse.
func SyncSharedSessionBrowserProfileStatusEvent(
	registry SharedSessionBrowserStateRegistry,
	sessionRegistry *BrowserSessionRegistry,
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	result BrowserProfileStatusResult,
	observedAt time.Time,
	reconnectWindow time.Duration,
) (SharedSessionBrowserProfileState, bool) {
	manager := sharedSessionBrowserObserverManager(
		sessionRegistry,
		nil,
		registry,
		reconnectWindow,
	)
	return manager.SyncProfileStatusEvent(sessionID, selectedInfo, result, observedAt)
}

// SyncSharedSessionBrowserProfilesEvent applies a route-scoped RuntimeProfiles
// observation to the shared session state contract and returns the final scoped
// lifecycle snapshot.
func SyncSharedSessionBrowserProfilesEvent(
	registry SharedSessionBrowserStateRegistry,
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	requestedProfile string,
	result BrowserProfilesResult,
	observedAt time.Time,
	reconnectWindow time.Duration,
) []SharedSessionBrowserProfileState {
	manager := sharedSessionBrowserObserverManager(
		nil,
		nil,
		registry,
		reconnectWindow,
	)
	return manager.SyncProfilesEvent(sessionID, selectedInfo, requestedProfile, result, observedAt)
}

// SyncSharedSessionBrowserStatusAndProfilesEvent applies a combined
// route-scoped RuntimeStatus and RuntimeProfiles observation to the shared
// session state contract, performs managed current-target invalidation, and
// seeds the shared watch-manager raw caches so the next event cycle can reuse
// the source-time event directly.
func SyncSharedSessionBrowserStatusAndProfilesEvent(
	registry SharedSessionBrowserStateRegistry,
	sessionRegistry *BrowserSessionRegistry,
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	requestedProfile string,
	status *BrowserProfileStatusResult,
	statusObservedAt time.Time,
	profiles *BrowserProfilesResult,
	profilesObservedAt time.Time,
	reconnectWindow time.Duration,
) (BrowserProfileStatusResult, SharedSessionBrowserProfileState, bool, []SharedSessionBrowserProfileState) {
	manager := sharedSessionBrowserObserverManager(
		sessionRegistry,
		nil,
		registry,
		reconnectWindow,
	)
	return manager.SyncStatusAndProfilesEvent(
		sessionID,
		selectedInfo,
		requestedProfile,
		status,
		statusObservedAt,
		profiles,
		profilesObservedAt,
	)
}

// SyncSharedSessionBrowserProfileLifecycleEvent applies a lifecycle-owned
// status/decision event to the shared session state contract and performs
// managed current-target invalidation when the resulting lifecycle state is no
// longer safe to reuse.
func SyncSharedSessionBrowserProfileLifecycleEvent(
	registry SharedSessionBrowserStateRegistry,
	sessionRegistry *BrowserSessionRegistry,
	runRegistry SharedSessionRunRegistry,
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	profile string,
	result BrowserProfileStatusResult,
	decision string,
	observedAt time.Time,
	reconnectWindow time.Duration,
) (SharedSessionBrowserProfileState, bool) {
	manager := sharedSessionBrowserObserverManager(
		sessionRegistry,
		runRegistry,
		registry,
		reconnectWindow,
	)
	return manager.SyncProfileLifecycleEvent(sessionID, selectedInfo, profile, result, decision, observedAt)
}

// InvalidateSharedSessionBrowserCurrentTargetForProfileStateEvent applies a
// lifecycle-owned managed current-target invalidation event through the shared
// observer-manager seam so provider-owned watch caches can refresh from fresh
// source-time observations instead of waiting for a passive rebuild.
func InvalidateSharedSessionBrowserCurrentTargetForProfileStateEvent(
	sessionRegistry *BrowserSessionRegistry,
	runRegistry SharedSessionRunRegistry,
	stateRegistry SharedSessionBrowserStateRegistry,
	sessionID string,
	state SharedSessionBrowserProfileState,
	reconnectWindow time.Duration,
) bool {
	return sharedSessionBrowserObserverManager(
		sessionRegistry,
		runRegistry,
		stateRegistry,
		reconnectWindow,
	).InvalidateCurrentTargetForProfileStateEvent(sessionID, state)
}

// SyncSharedSessionBrowserTabsForRouteEvent applies a route-scoped tab sync
// event through the shared observer-manager seam so provider-owned watch caches
// can refresh from cached source-time projections instead of waiting for the
// next passive rebuild.
func SyncSharedSessionBrowserTabsForRouteEvent(
	sessionRegistry *BrowserSessionRegistry,
	runRegistry SharedSessionRunRegistry,
	stateRegistry SharedSessionBrowserStateRegistry,
	sessionID string,
	route BrowserSessionRoute,
	activeIndex int,
	tabs []BrowserTab,
	reconnectWindow time.Duration,
) []BrowserTab {
	return sharedSessionBrowserObserverManager(
		sessionRegistry,
		runRegistry,
		stateRegistry,
		reconnectWindow,
	).SyncTabsForRouteEvent(sessionID, route, activeIndex, tabs)
}

// TrackSharedSessionBrowserTabEvent applies a single route-scoped tab
// attach/update event through the shared observer-manager seam.
func TrackSharedSessionBrowserTabEvent(
	sessionRegistry *BrowserSessionRegistry,
	runRegistry SharedSessionRunRegistry,
	stateRegistry SharedSessionBrowserStateRegistry,
	sessionID string,
	route BrowserSessionRoute,
	tab BrowserTab,
	setCurrent bool,
	reconnectWindow time.Duration,
) BrowserSessionTarget {
	return sharedSessionBrowserObserverManager(
		sessionRegistry,
		runRegistry,
		stateRegistry,
		reconnectWindow,
	).TrackTabEvent(sessionID, route, tab, setCurrent)
}

// TrackSharedSessionBrowserCurrentTargetEvent applies a route-scoped
// current-target sync event through the shared observer-manager seam.
func TrackSharedSessionBrowserCurrentTargetEvent(
	sessionRegistry *BrowserSessionRegistry,
	runRegistry SharedSessionRunRegistry,
	stateRegistry SharedSessionBrowserStateRegistry,
	sessionID string,
	route BrowserSessionRoute,
	url string,
	title string,
	source string,
	reconnectWindow time.Duration,
) BrowserSessionTarget {
	return sharedSessionBrowserObserverManager(
		sessionRegistry,
		runRegistry,
		stateRegistry,
		reconnectWindow,
	).TrackCurrentTargetEvent(sessionID, route, url, title, source)
}

// ForgetSharedSessionBrowserTabForRouteEvent applies a route-scoped tab
// detach/close event through the shared observer-manager seam.
func ForgetSharedSessionBrowserTabForRouteEvent(
	sessionRegistry *BrowserSessionRegistry,
	runRegistry SharedSessionRunRegistry,
	stateRegistry SharedSessionBrowserStateRegistry,
	sessionID string,
	route BrowserSessionRoute,
	tabIndex int,
	reconnectWindow time.Duration,
) {
	sharedSessionBrowserObserverManager(
		sessionRegistry,
		runRegistry,
		stateRegistry,
		reconnectWindow,
	).ForgetTabForRouteEvent(sessionID, route, tabIndex)
}

// RecordSharedSessionBrowserPendingTargetReviewEvent applies a route-scoped
// pending target review event through the shared observer-manager seam.
func RecordSharedSessionBrowserPendingTargetReviewEvent(
	sessionRegistry *BrowserSessionRegistry,
	runRegistry SharedSessionRunRegistry,
	stateRegistry SharedSessionBrowserStateRegistry,
	sessionID string,
	route BrowserSessionRoute,
	targetID string,
	tabIndex int,
	finalURL string,
	title string,
	decision string,
	reason string,
	reconnectWindow time.Duration,
) *BrowserSessionTargetReview {
	return sharedSessionBrowserObserverManager(
		sessionRegistry,
		runRegistry,
		stateRegistry,
		reconnectWindow,
	).RecordPendingTargetReviewEvent(
		sessionID,
		route,
		targetID,
		tabIndex,
		finalURL,
		title,
		decision,
		reason,
	)
}

// RecordSharedSessionBrowserPendingTargetPopupReviewEvent applies a
// route-scoped popup review event through the shared observer-manager seam.
func RecordSharedSessionBrowserPendingTargetPopupReviewEvent(
	sessionRegistry *BrowserSessionRegistry,
	runRegistry SharedSessionRunRegistry,
	stateRegistry SharedSessionBrowserStateRegistry,
	sessionID string,
	route BrowserSessionRoute,
	activeTab BrowserTab,
	decision string,
	reason string,
	reconnectWindow time.Duration,
) *BrowserSessionTargetReview {
	return sharedSessionBrowserObserverManager(
		sessionRegistry,
		runRegistry,
		stateRegistry,
		reconnectWindow,
	).RecordPendingTargetPopupReviewEvent(sessionID, route, activeTab, decision, reason)
}

// RestoreSharedSessionBrowserCurrentTargetSelectionEvent applies a route-scoped
// current-target restore event through the shared observer-manager seam.
func RestoreSharedSessionBrowserCurrentTargetSelectionEvent(
	sessionRegistry *BrowserSessionRegistry,
	runRegistry SharedSessionRunRegistry,
	stateRegistry SharedSessionBrowserStateRegistry,
	sessionID string,
	route BrowserSessionRoute,
	snapshot *BrowserSessionTargetSelection,
	source string,
	reconnectWindow time.Duration,
) *BrowserSessionTargetSelection {
	return sharedSessionBrowserObserverManager(
		sessionRegistry,
		runRegistry,
		stateRegistry,
		reconnectWindow,
	).RestoreCurrentTargetSelectionEvent(sessionID, route, snapshot, source)
}

// RestoreSharedSessionBrowserCurrentTargetSelectionForPendingReviewEvent
// applies a route-scoped current-target restore event for popup/redirect review
// posture through the shared observer-manager seam.
func RestoreSharedSessionBrowserCurrentTargetSelectionForPendingReviewEvent(
	sessionRegistry *BrowserSessionRegistry,
	runRegistry SharedSessionRunRegistry,
	stateRegistry SharedSessionBrowserStateRegistry,
	sessionID string,
	route BrowserSessionRoute,
	snapshot *BrowserSessionTargetSelection,
	pendingTargetID string,
	source string,
	reconnectWindow time.Duration,
) *BrowserSessionTargetSelection {
	return sharedSessionBrowserObserverManager(
		sessionRegistry,
		runRegistry,
		stateRegistry,
		reconnectWindow,
	).RestoreCurrentTargetSelectionForPendingReviewEvent(
		sessionID,
		route,
		snapshot,
		pendingTargetID,
		source,
	)
}

// ApplySharedSessionBrowserTabRememberReviewEvent applies the shared
// popup-review guardrail for tab lifecycle actions through the shared
// observer-manager seam.
func ApplySharedSessionBrowserTabRememberReviewEvent(
	sessionRegistry *BrowserSessionRegistry,
	runRegistry SharedSessionRunRegistry,
	stateRegistry SharedSessionBrowserStateRegistry,
	req SharedSessionBrowserTabRememberReviewRequest,
	reconnectWindow time.Duration,
) SharedSessionBrowserTabRememberReviewResult {
	return sharedSessionBrowserObserverManager(
		sessionRegistry,
		runRegistry,
		stateRegistry,
		reconnectWindow,
	).ApplyTabRememberReviewEvent(req)
}

// SelectSharedSessionBrowserProfileEvent applies the shared profile-selection
// contract through the shared observer-manager seam.
func SelectSharedSessionBrowserProfileEvent(
	ctx context.Context,
	sessionRegistry *BrowserSessionRegistry,
	runRegistry SharedSessionRunRegistry,
	stateRegistry SharedSessionBrowserStateRegistry,
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	browserApp string,
	control BrowserRuntimeControlBackend,
	validateWithProfiles bool,
	source string,
	reconnectWindow time.Duration,
) (SharedSessionBrowserProfileSelection, string, bool, error) {
	return sharedSessionBrowserObserverManager(
		sessionRegistry,
		runRegistry,
		stateRegistry,
		reconnectWindow,
	).SelectProfile(
		ctx,
		sessionID,
		selectedInfo,
		browserApp,
		control,
		validateWithProfiles,
		source,
	)
}

// SyncOrClearSharedSessionBrowserCurrentTargetForProfileSelectionEvent applies
// the shared current-target sync/clear contract through the shared
// observer-manager seam.
func SyncOrClearSharedSessionBrowserCurrentTargetForProfileSelectionEvent(
	sessionRegistry *BrowserSessionRegistry,
	runRegistry SharedSessionRunRegistry,
	stateRegistry SharedSessionBrowserStateRegistry,
	sessionID string,
	route BrowserSessionRoute,
	profileSelection *SharedSessionBrowserProfileSelection,
	source string,
	reconnectWindow time.Duration,
) (*BrowserSessionTargetSelection, string, error) {
	return sharedSessionBrowserObserverManager(
		sessionRegistry,
		runRegistry,
		stateRegistry,
		reconnectWindow,
	).SyncOrClearCurrentTargetForProfileSelection(
		sessionID,
		route,
		profileSelection,
		source,
	)
}

// SyncSharedSessionBrowserRouteSelectionEvent applies the shared route-scoped
// profile/target sync contract through the shared observer-manager seam.
func SyncSharedSessionBrowserRouteSelectionEvent(
	ctx context.Context,
	sessionRegistry *BrowserSessionRegistry,
	runRegistry SharedSessionRunRegistry,
	stateRegistry SharedSessionBrowserStateRegistry,
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	route BrowserSessionRoute,
	browserApp string,
	control BrowserRuntimeControlBackend,
	validateWithProfiles bool,
	source string,
	reconnectWindow time.Duration,
) (SharedSessionBrowserSyncSelectionResult, error) {
	return sharedSessionBrowserObserverManager(
		sessionRegistry,
		runRegistry,
		stateRegistry,
		reconnectWindow,
	).SyncRouteSelection(
		ctx,
		sessionID,
		selectedInfo,
		route,
		browserApp,
		control,
		validateWithProfiles,
		source,
	)
}

// CoordinateSharedSessionBrowserRouteSyncEvent applies the target-first
// route-sync contract through the shared observer-manager seam.
func CoordinateSharedSessionBrowserRouteSyncEvent(
	ctx context.Context,
	sessionRegistry *BrowserSessionRegistry,
	runRegistry SharedSessionRunRegistry,
	stateRegistry SharedSessionBrowserStateRegistry,
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	route BrowserSessionRoute,
	browserApp string,
	control BrowserRuntimeControlBackend,
	validateWithProfiles bool,
	reconnectWindow time.Duration,
) (SharedSessionBrowserSyncSelectionResult, error) {
	return sharedSessionBrowserObserverManager(
		sessionRegistry,
		runRegistry,
		stateRegistry,
		reconnectWindow,
	).CoordinateRouteSync(
		ctx,
		sessionID,
		selectedInfo,
		route,
		browserApp,
		control,
		validateWithProfiles,
	)
}

// SelectSharedSessionBrowserTargetEvent applies the shared route-scoped target
// selection contract through the shared observer-manager seam.
func SelectSharedSessionBrowserTargetEvent(
	sessionRegistry *BrowserSessionRegistry,
	runRegistry SharedSessionRunRegistry,
	stateRegistry SharedSessionBrowserStateRegistry,
	req SharedSessionBrowserSelectTargetRequest,
	reconnectWindow time.Duration,
) (SharedSessionBrowserSelectTargetResult, error) {
	return sharedSessionBrowserObserverManager(
		sessionRegistry,
		runRegistry,
		stateRegistry,
		reconnectWindow,
	).SelectTarget(req)
}

// RememberSharedSessionBrowserTargetEvent applies the shared remember-target
// contract through the shared observer-manager seam.
func RememberSharedSessionBrowserTargetEvent(
	sessionRegistry *BrowserSessionRegistry,
	runRegistry SharedSessionRunRegistry,
	stateRegistry SharedSessionBrowserStateRegistry,
	req SharedSessionBrowserRememberTargetRequest,
	reconnectWindow time.Duration,
) SharedSessionBrowserRememberTargetResult {
	return sharedSessionBrowserObserverManager(
		sessionRegistry,
		runRegistry,
		stateRegistry,
		reconnectWindow,
	).RememberTarget(req)
}

// RememberSharedSessionBrowserProfileForRouteEvent applies the shared
// remember-profile contract through the shared observer-manager seam.
func RememberSharedSessionBrowserProfileForRouteEvent(
	sessionRegistry *BrowserSessionRegistry,
	runRegistry SharedSessionRunRegistry,
	stateRegistry SharedSessionBrowserStateRegistry,
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	route BrowserSessionRoute,
	profileStatus *BrowserProfileStatusResult,
	preparedProfile string,
	requestedProfile string,
	requestedBrowserApp string,
	reconnectWindow time.Duration,
) SharedSessionBrowserRememberProfileResult {
	return sharedSessionBrowserObserverManager(
		sessionRegistry,
		runRegistry,
		stateRegistry,
		reconnectWindow,
	).RememberProfileForRoute(
		sessionID,
		selectedInfo,
		route,
		profileStatus,
		preparedProfile,
		requestedProfile,
		requestedBrowserApp,
	)
}

// ExecuteSharedSessionBrowserClearProfileEvent applies the shared clear-profile
// contract through the shared observer-manager seam.
func ExecuteSharedSessionBrowserClearProfileEvent(
	sessionRegistry *BrowserSessionRegistry,
	runRegistry SharedSessionRunRegistry,
	stateRegistry SharedSessionBrowserStateRegistry,
	req SharedSessionBrowserClearRequest,
	reconnectWindow time.Duration,
) SharedSessionBrowserClearResult {
	return sharedSessionBrowserObserverManager(
		sessionRegistry,
		runRegistry,
		stateRegistry,
		reconnectWindow,
	).ExecuteClearProfile(req)
}

// ExecuteSharedSessionBrowserClearTargetEvent applies the shared clear-target
// contract through the shared observer-manager seam.
func ExecuteSharedSessionBrowserClearTargetEvent(
	sessionRegistry *BrowserSessionRegistry,
	runRegistry SharedSessionRunRegistry,
	stateRegistry SharedSessionBrowserStateRegistry,
	req SharedSessionBrowserClearRequest,
	reconnectWindow time.Duration,
) SharedSessionBrowserClearResult {
	return sharedSessionBrowserObserverManager(
		sessionRegistry,
		runRegistry,
		stateRegistry,
		reconnectWindow,
	).ExecuteClearTarget(req)
}

// ExecuteSharedSessionBrowserClearSessionEvent applies the shared clear-session
// contract through the shared observer-manager seam.
func ExecuteSharedSessionBrowserClearSessionEvent(
	sessionRegistry *BrowserSessionRegistry,
	runRegistry SharedSessionRunRegistry,
	stateRegistry SharedSessionBrowserStateRegistry,
	req SharedSessionBrowserClearRequest,
	reconnectWindow time.Duration,
) SharedSessionBrowserClearResult {
	return sharedSessionBrowserObserverManager(
		sessionRegistry,
		runRegistry,
		stateRegistry,
		reconnectWindow,
	).ExecuteClearSession(req)
}
