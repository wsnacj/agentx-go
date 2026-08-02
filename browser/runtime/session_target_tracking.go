package browserruntime

import (
	"strings"
	"time"
)

// SharedSessionBrowserMutationContext carries the optional shared-manager
// dependencies that let browserruntime helpers route their writes through the
// top-level mutation/event seam instead of falling back to raw registry
// mutations.
type SharedSessionBrowserMutationContext struct {
	Registry        *BrowserSessionRegistry
	RunRegistry     SharedSessionRunRegistry
	StateRegistry   SharedSessionBrowserStateRegistry
	ReconnectWindow time.Duration
}

// SharedSessionBrowserMutationContextFor merges a partially populated observer
// manager with fallback registry dependencies so tools callers can hand
// browserruntime one shared mutation context without re-implementing fallback
// wiring.
func SharedSessionBrowserMutationContextFor(
	manager SharedSessionBrowserObserverManager,
	sessionRegistry *BrowserSessionRegistry,
	stateRegistry SharedSessionBrowserStateRegistry,
	reconnectWindow time.Duration,
) SharedSessionBrowserMutationContext {
	ctx := SharedSessionBrowserMutationContext{
		Registry:        manager.SessionRegistry,
		RunRegistry:     manager.RunRegistry,
		StateRegistry:   manager.StateRegistry,
		ReconnectWindow: manager.ReconnectWindow,
	}
	if ctx.Registry == nil {
		ctx.Registry = sessionRegistry
	}
	if ctx.StateRegistry == nil {
		ctx.StateRegistry = stateRegistry
	}
	if ctx.ReconnectWindow == 0 {
		ctx.ReconnectWindow = reconnectWindow
	}
	return ctx
}

func (ctx SharedSessionBrowserMutationContext) usesWatchManagerEventSeam() bool {
	return ctx.Registry != nil &&
		ctx.ReconnectWindow > 0 &&
		(ctx.RunRegistry != nil || ctx.StateRegistry != nil)
}

// ResolveSharedSessionBrowserTabTargetID resolves a tracked target ID for a
// route-scoped tab index after pruning stale route state.
func ResolveSharedSessionBrowserTabTargetID(registry *BrowserSessionRegistry, sessionID string, route BrowserSessionRoute, tabIndex int) string {
	sessionID = strings.TrimSpace(sessionID)
	route = normalizeBrowserSessionRoute(route)
	if registry == nil || sessionID == "" || tabIndex <= 0 {
		return ""
	}
	registry.PruneStaleRouteState(sessionID, route)
	target, ok := registry.ResolveTabForRoute(sessionID, route, tabIndex)
	if !ok {
		return ""
	}
	return strings.TrimSpace(target.ID)
}

// SnapshotSharedSessionBrowserCurrentTargetSelection snapshots the currently
// selected target for a route-scoped session after pruning stale route state.
func SnapshotSharedSessionBrowserCurrentTargetSelection(registry *BrowserSessionRegistry, sessionID string, route BrowserSessionRoute) *BrowserSessionTargetSelection {
	return CurrentSharedSessionBrowserTargetSelection(registry, strings.TrimSpace(sessionID), normalizeBrowserSessionRoute(route))
}

// RestoreSharedSessionBrowserCurrentTargetSelectionWithContext restores a
// previously selected target for a route-scoped session and routes the write
// through the shared mutation seam when manager dependencies are available.
func RestoreSharedSessionBrowserCurrentTargetSelectionWithContext(
	ctx SharedSessionBrowserMutationContext,
	sessionID string,
	route BrowserSessionRoute,
	snapshot *BrowserSessionTargetSelection,
	source string,
) *BrowserSessionTargetSelection {
	sessionID = strings.TrimSpace(sessionID)
	route = normalizeBrowserSessionRoute(route)
	if ctx.Registry == nil || sessionID == "" {
		return nil
	}
	if ctx.usesWatchManagerEventSeam() {
		return RestoreSharedSessionBrowserCurrentTargetSelectionEvent(
			ctx.Registry,
			ctx.RunRegistry,
			ctx.StateRegistry,
			sessionID,
			route,
			snapshot,
			source,
			ctx.ReconnectWindow,
		)
	}
	if snapshot == nil || strings.TrimSpace(snapshot.ID) == "" {
		ctx.Registry.ClearCurrentTargetForRoute(sessionID, route)
		return nil
	}
	source = firstNonEmptyString(strings.TrimSpace(snapshot.Source), strings.TrimSpace(source), "popup_review_restore")
	if selected, ok := ctx.Registry.SelectTargetForRoute(sessionID, route, strings.TrimSpace(snapshot.ID), source); ok {
		return sharedSessionBrowserTargetSelectionFromTracked(selected, source)
	}
	if selected, ok := ctx.Registry.SelectTargetForRoute(sessionID, BrowserSessionRoute{}, strings.TrimSpace(snapshot.ID), source); ok {
		return sharedSessionBrowserTargetSelectionFromTracked(selected, source)
	}
	ctx.Registry.ClearCurrentTargetForRoute(sessionID, route)
	return nil
}

// RestoreSharedSessionBrowserCurrentTargetSelection restores a previously
// selected target for a route-scoped session, falling back to the default route
// when the target no longer resolves on the original route.
func RestoreSharedSessionBrowserCurrentTargetSelection(registry *BrowserSessionRegistry, sessionID string, route BrowserSessionRoute, snapshot *BrowserSessionTargetSelection, source string) *BrowserSessionTargetSelection {
	return RestoreSharedSessionBrowserCurrentTargetSelectionWithContext(
		SharedSessionBrowserMutationContext{Registry: registry},
		sessionID,
		route,
		snapshot,
		source,
	)
}

// RestoreSharedSessionBrowserCurrentTargetSelectionForPendingReviewWithContext
// restores a prior current-target selection unless that selection already
// matches the pending-review target that should remain detached from auto-
// follow, and routes the write through the shared mutation seam when possible.
func RestoreSharedSessionBrowserCurrentTargetSelectionForPendingReviewWithContext(
	ctx SharedSessionBrowserMutationContext,
	sessionID string,
	route BrowserSessionRoute,
	snapshot *BrowserSessionTargetSelection,
	pendingTargetID string,
	source string,
) *BrowserSessionTargetSelection {
	if snapshot != nil &&
		strings.TrimSpace(snapshot.ID) != "" &&
		strings.EqualFold(strings.TrimSpace(snapshot.ID), strings.TrimSpace(pendingTargetID)) {
		return snapshot
	}
	if ctx.usesWatchManagerEventSeam() {
		return RestoreSharedSessionBrowserCurrentTargetSelectionForPendingReviewEvent(
			ctx.Registry,
			ctx.RunRegistry,
			ctx.StateRegistry,
			sessionID,
			route,
			snapshot,
			pendingTargetID,
			source,
			ctx.ReconnectWindow,
		)
	}
	return RestoreSharedSessionBrowserCurrentTargetSelectionWithContext(ctx, sessionID, route, snapshot, source)
}

// RestoreSharedSessionBrowserCurrentTargetSelectionForPendingReview restores a
// prior current-target selection unless that selection already matches the
// pending-review target that should remain detached from auto-follow.
func RestoreSharedSessionBrowserCurrentTargetSelectionForPendingReview(registry *BrowserSessionRegistry, sessionID string, route BrowserSessionRoute, snapshot *BrowserSessionTargetSelection, pendingTargetID string, source string) *BrowserSessionTargetSelection {
	return RestoreSharedSessionBrowserCurrentTargetSelectionForPendingReviewWithContext(
		SharedSessionBrowserMutationContext{Registry: registry},
		sessionID,
		route,
		snapshot,
		pendingTargetID,
		source,
	)
}

// SharedSessionBrowserProfileStateNeedsTargetInvalidation reports whether a
// lifecycle state should detach the current target selection for a managed
// route because the underlying browser profile is no longer safe to reuse.
func SharedSessionBrowserProfileStateNeedsTargetInvalidation(state SharedSessionBrowserProfileState) bool {
	state = normalizeBrowserSessionProfileState(state)
	status := strings.ToLower(strings.TrimSpace(state.Status))
	switch status {
	case "reconnecting", "disconnected", "crashed", "stopped", "deleted", "delete_requested", "teardown_stopped", "teardown_already_stopped":
		return true
	}
	return state.Running && !state.Connected
}

// InvalidateSharedSessionBrowserCurrentTargetForProfileState clears the current
// selected target for a managed route when a lifecycle update indicates the
// profile has disconnected, stopped, or otherwise become unsafe to reuse.
func invalidateSharedSessionBrowserCurrentTargetForProfileState(registry *BrowserSessionRegistry, sessionID string, state SharedSessionBrowserProfileState, invalidate bool) bool {
	sessionID = strings.TrimSpace(sessionID)
	state = normalizeBrowserSessionProfileState(state)
	if registry == nil || sessionID == "" || !SharedSessionBrowserProfileStateNeedsTargetInvalidation(state) {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(state.RuntimeTarget)) {
	case "node", "sandbox":
	default:
		return false
	}
	route := BrowserSessionRoute{
		Backend:    strings.TrimSpace(state.Backend),
		Profile:    strings.TrimSpace(state.Profile),
		Target:     strings.TrimSpace(state.RuntimeTarget),
		BrowserApp: strings.TrimSpace(state.BrowserApp),
	}
	registry.PruneStaleRouteState(sessionID, route)
	if registry.clearCurrentTargetForRoute(sessionID, route, invalidate) {
		return true
	}
	if strings.TrimSpace(route.BrowserApp) == "" {
		return false
	}
	route.BrowserApp = ""
	registry.PruneStaleRouteState(sessionID, route)
	return registry.clearCurrentTargetForRoute(sessionID, route, invalidate)
}

func InvalidateSharedSessionBrowserCurrentTargetForProfileState(registry *BrowserSessionRegistry, sessionID string, state SharedSessionBrowserProfileState) bool {
	return invalidateSharedSessionBrowserCurrentTargetForProfileState(registry, sessionID, state, true)
}

// TrackSharedSessionBrowserTab tracks a route-scoped tab and optionally makes
// it the current target for that route.
func TrackSharedSessionBrowserTab(registry *BrowserSessionRegistry, sessionID string, route BrowserSessionRoute, tab BrowserTab, setCurrent bool) BrowserSessionTarget {
	return TrackSharedSessionBrowserTabWithContext(
		SharedSessionBrowserMutationContext{Registry: registry},
		sessionID,
		route,
		tab,
		setCurrent,
	)
}

// TrackSharedSessionBrowserTabWithContext tracks a route-scoped tab and
// optionally makes it the current target for that route, routing the write
// through the shared mutation seam when manager dependencies are available.
func TrackSharedSessionBrowserTabWithContext(ctx SharedSessionBrowserMutationContext, sessionID string, route BrowserSessionRoute, tab BrowserTab, setCurrent bool) BrowserSessionTarget {
	sessionID = strings.TrimSpace(sessionID)
	route = normalizeBrowserSessionRoute(route)
	tab.TargetID = strings.TrimSpace(tab.TargetID)
	if ctx.Registry == nil || sessionID == "" || tab.Index <= 0 {
		return BrowserSessionTarget{}
	}
	if ctx.usesWatchManagerEventSeam() {
		return TrackSharedSessionBrowserTabEvent(
			ctx.Registry,
			ctx.RunRegistry,
			ctx.StateRegistry,
			sessionID,
			route,
			tab,
			setCurrent,
			ctx.ReconnectWindow,
		)
	}
	target := ctx.Registry.TrackTab(sessionID, BrowserSessionTarget{
		TabIndex:   tab.Index,
		URL:        strings.TrimSpace(tab.URL),
		Title:      strings.TrimSpace(tab.Title),
		BrowserApp: strings.TrimSpace(route.BrowserApp),
		Backend:    strings.TrimSpace(route.Backend),
		Profile:    strings.TrimSpace(route.Profile),
		Target:     strings.TrimSpace(route.Target),
	}, setCurrent)
	return target
}

// TrackSharedSessionBrowserCurrentTarget tracks a route-scoped current target
// without requiring a tab index.
func TrackSharedSessionBrowserCurrentTarget(registry *BrowserSessionRegistry, sessionID string, route BrowserSessionRoute, url string, title string, source string) BrowserSessionTarget {
	return TrackSharedSessionBrowserCurrentTargetWithContext(
		SharedSessionBrowserMutationContext{Registry: registry},
		sessionID,
		route,
		url,
		title,
		source,
	)
}

// TrackSharedSessionBrowserCurrentTargetWithContext tracks a route-scoped
// current target without requiring a tab index and routes the write through the
// shared mutation seam when manager dependencies are available.
func TrackSharedSessionBrowserCurrentTargetWithContext(ctx SharedSessionBrowserMutationContext, sessionID string, route BrowserSessionRoute, url string, title string, source string) BrowserSessionTarget {
	sessionID = strings.TrimSpace(sessionID)
	route = normalizeBrowserSessionRoute(route)
	if ctx.Registry == nil || sessionID == "" {
		return BrowserSessionTarget{}
	}
	if ctx.usesWatchManagerEventSeam() {
		return TrackSharedSessionBrowserCurrentTargetEvent(
			ctx.Registry,
			ctx.RunRegistry,
			ctx.StateRegistry,
			sessionID,
			route,
			url,
			title,
			source,
			ctx.ReconnectWindow,
		)
	}
	target := ctx.Registry.TrackCurrentTarget(sessionID, BrowserSessionTarget{
		URL:        strings.TrimSpace(url),
		Title:      strings.TrimSpace(title),
		BrowserApp: strings.TrimSpace(route.BrowserApp),
		Backend:    strings.TrimSpace(route.Backend),
		Profile:    strings.TrimSpace(route.Profile),
		Target:     strings.TrimSpace(route.Target),
	}, strings.TrimSpace(source))
	return target
}

// TrackSharedSessionBrowserResolvedTarget tracks a resolved target using a tab
// index when one exists, otherwise it falls back to current-target tracking.
func TrackSharedSessionBrowserResolvedTarget(registry *BrowserSessionRegistry, sessionID string, route BrowserSessionRoute, tab BrowserTab, source string) BrowserSessionTarget {
	return TrackSharedSessionBrowserResolvedTargetWithContext(
		SharedSessionBrowserMutationContext{Registry: registry},
		sessionID,
		route,
		tab,
		source,
	)
}

// TrackSharedSessionBrowserResolvedTargetWithContext tracks a resolved target
// using a tab index when one exists, otherwise it falls back to current-target
// tracking, and routes the write through the shared mutation seam when manager
// dependencies are available.
func TrackSharedSessionBrowserResolvedTargetWithContext(ctx SharedSessionBrowserMutationContext, sessionID string, route BrowserSessionRoute, tab BrowserTab, source string) BrowserSessionTarget {
	if tab.Index > 0 {
		return TrackSharedSessionBrowserTabWithContext(ctx, sessionID, route, tab, true)
	}
	return TrackSharedSessionBrowserCurrentTargetWithContext(ctx, sessionID, route, tab.URL, tab.Title, source)
}

// ApplySharedSessionBrowserTargetWithContext applies a generic target-tracking
// runtime result and routes the write through the top-level mutation seam when
// manager dependencies are available.
func ApplySharedSessionBrowserTargetWithContext(
	ctx SharedSessionBrowserMutationContext,
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
	if ctx.Registry == nil || req.SessionID == "" {
		return result
	}
	if ctx.usesWatchManagerEventSeam() {
		return ApplySharedSessionBrowserTargetEvent(
			ctx.Registry,
			ctx.RunRegistry,
			ctx.StateRegistry,
			req,
			ctx.ReconnectWindow,
		)
	}

	if req.TabIndex > 0 {
		result.Target = TrackSharedSessionBrowserTabWithContext(
			ctx,
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
		result.Target = TrackSharedSessionBrowserCurrentTargetWithContext(
			ctx,
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

// ApplySharedSessionBrowserTarget applies a generic target-tracking runtime
// result through the shared target-tracking contract.
func ApplySharedSessionBrowserTarget(
	registry *BrowserSessionRegistry,
	req SharedSessionBrowserTargetEventRequest,
) SharedSessionBrowserTargetEventResult {
	return ApplySharedSessionBrowserTargetWithContext(
		SharedSessionBrowserMutationContext{Registry: registry},
		req,
	)
}

// ApplySharedSessionBrowserResolvedTargetWithContext applies a resolved-target
// runtime result and routes the write through the top-level mutation seam when
// manager dependencies are available.
func ApplySharedSessionBrowserResolvedTargetWithContext(
	ctx SharedSessionBrowserMutationContext,
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
	if ctx.Registry == nil || req.SessionID == "" {
		return result
	}
	if ctx.usesWatchManagerEventSeam() {
		return ApplySharedSessionBrowserResolvedTargetEvent(
			ctx.Registry,
			ctx.RunRegistry,
			ctx.StateRegistry,
			req,
			ctx.ReconnectWindow,
		)
	}

	tracked := TrackSharedSessionBrowserResolvedTargetWithContext(
		ctx,
		req.SessionID,
		req.Route,
		BrowserTab{
			Index: req.TabIndex,
			URL:   req.URL,
			Title: req.Title,
		},
		firstNonEmptyString(req.Source, "runtime_result"),
	)
	result.TargetID = firstNonEmptyString(result.TargetID, strings.TrimSpace(tracked.ID))

	if req.PendingReview {
		result.Review = RecordSharedSessionBrowserPendingTargetReviewWithContext(
			ctx,
			req.SessionID,
			req.Route,
			result.TargetID,
			req.TabIndex,
			req.URL,
			req.Title,
			req.PendingReviewDecision,
			req.PendingReviewReason,
		)
		result.RestoredSelection = RestoreSharedSessionBrowserCurrentTargetSelectionForPendingReviewWithContext(
			ctx,
			req.SessionID,
			req.Route,
			req.PriorSelection,
			result.TargetID,
			"popup_review_restore",
		)
	}

	return result
}

// ApplySharedSessionBrowserResolvedTarget applies a resolved-target runtime
// result through the shared target-tracking contract.
func ApplySharedSessionBrowserResolvedTarget(
	registry *BrowserSessionRegistry,
	req SharedSessionBrowserResolvedTargetEventRequest,
) SharedSessionBrowserResolvedTargetEventResult {
	return ApplySharedSessionBrowserResolvedTargetWithContext(
		SharedSessionBrowserMutationContext{Registry: registry},
		req,
	)
}

// SyncSharedSessionBrowserTabsForRoute tracks a full route-scoped tab list and
// projects the resulting target IDs back onto the browser tabs payload.
func SyncSharedSessionBrowserTabsForRoute(registry *BrowserSessionRegistry, sessionID string, route BrowserSessionRoute, activeIndex int, tabs []BrowserTab) []BrowserTab {
	return SyncSharedSessionBrowserTabsForRouteWithContext(
		SharedSessionBrowserMutationContext{Registry: registry},
		sessionID,
		route,
		activeIndex,
		tabs,
	)
}

// SyncSharedSessionBrowserTabsForRouteWithContext tracks a full route-scoped
// tab list and projects the resulting target IDs back onto the browser tabs
// payload, routing the write through the shared mutation seam when manager
// dependencies are available.
func SyncSharedSessionBrowserTabsForRouteWithContext(ctx SharedSessionBrowserMutationContext, sessionID string, route BrowserSessionRoute, activeIndex int, tabs []BrowserTab) []BrowserTab {
	out := append([]BrowserTab(nil), tabs...)
	sessionID = strings.TrimSpace(sessionID)
	route = normalizeBrowserSessionRoute(route)
	if ctx.Registry == nil || sessionID == "" {
		return out
	}
	if ctx.usesWatchManagerEventSeam() {
		return SyncSharedSessionBrowserTabsForRouteEvent(
			ctx.Registry,
			ctx.RunRegistry,
			ctx.StateRegistry,
			sessionID,
			route,
			activeIndex,
			tabs,
			ctx.ReconnectWindow,
		)
	}
	inputs := make([]BrowserSessionTarget, 0, len(tabs))
	for _, tab := range tabs {
		inputs = append(inputs, BrowserSessionTarget{
			TabIndex:   tab.Index,
			URL:        strings.TrimSpace(tab.URL),
			Title:      strings.TrimSpace(tab.Title),
			BrowserApp: strings.TrimSpace(route.BrowserApp),
			Backend:    strings.TrimSpace(route.Backend),
			Profile:    strings.TrimSpace(route.Profile),
			Target:     strings.TrimSpace(route.Target),
		})
	}
	tracked := ctx.Registry.SyncTabsForRoute(sessionID, route, inputs, activeIndex)
	for i := range out {
		if i < len(tracked) {
			out[i].TargetID = strings.TrimSpace(tracked[i].ID)
		}
	}
	return out
}

// ForgetSharedSessionBrowserTabForRoute forgets a tracked tab for the scoped
// session route.
func ForgetSharedSessionBrowserTabForRoute(registry *BrowserSessionRegistry, sessionID string, route BrowserSessionRoute, tabIndex int) {
	ForgetSharedSessionBrowserTabForRouteWithContext(
		SharedSessionBrowserMutationContext{Registry: registry},
		sessionID,
		route,
		tabIndex,
	)
}

// ForgetSharedSessionBrowserTabForRouteWithContext forgets a tracked tab for
// the scoped session route and routes the write through the shared mutation
// seam when manager dependencies are available.
func ForgetSharedSessionBrowserTabForRouteWithContext(ctx SharedSessionBrowserMutationContext, sessionID string, route BrowserSessionRoute, tabIndex int) {
	sessionID = strings.TrimSpace(sessionID)
	route = normalizeBrowserSessionRoute(route)
	if ctx.Registry == nil || sessionID == "" || tabIndex <= 0 {
		return
	}
	if ctx.usesWatchManagerEventSeam() {
		ForgetSharedSessionBrowserTabForRouteEvent(
			ctx.Registry,
			ctx.RunRegistry,
			ctx.StateRegistry,
			sessionID,
			route,
			tabIndex,
			ctx.ReconnectWindow,
		)
		return
	}
	ctx.Registry.ForgetTabForRoute(sessionID, route, tabIndex)
}
