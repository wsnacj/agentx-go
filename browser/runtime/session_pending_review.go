package browserruntime

import (
	"fmt"
	"strings"
)

// SharedSessionBrowserPendingTargetReviewState captures the route-scoped
// pending-review posture shared across browser runtime and action tools.
type SharedSessionBrowserPendingTargetReviewState struct {
	Review       *BrowserSessionTargetReview
	Count        int
	PolicyState  string
	PolicyReason string
}

// RecordSharedSessionBrowserPendingTargetReview records a route-scoped pending
// review posture, resolving the target handle from the tracked tab when needed.
func RecordSharedSessionBrowserPendingTargetReview(
	registry *BrowserSessionRegistry,
	sessionID string,
	route BrowserSessionRoute,
	targetID string,
	tabIndex int,
	finalURL string,
	title string,
	decision string,
	reason string,
) *BrowserSessionTargetReview {
	return RecordSharedSessionBrowserPendingTargetReviewWithContext(
		SharedSessionBrowserMutationContext{Registry: registry},
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

// RecordSharedSessionBrowserPendingTargetReviewWithContext records a route-
// scoped pending review posture, resolving the target handle from the tracked
// tab when needed and routing the write through the shared mutation seam when
// manager dependencies are available.
func RecordSharedSessionBrowserPendingTargetReviewWithContext(
	ctx SharedSessionBrowserMutationContext,
	sessionID string,
	route BrowserSessionRoute,
	targetID string,
	tabIndex int,
	finalURL string,
	title string,
	decision string,
	reason string,
) *BrowserSessionTargetReview {
	sessionID = strings.TrimSpace(sessionID)
	route = normalizeBrowserSessionRoute(route)
	targetID = strings.TrimSpace(targetID)
	if ctx.Registry == nil || sessionID == "" {
		return nil
	}
	if ctx.usesWatchManagerEventSeam() {
		return RecordSharedSessionBrowserPendingTargetReviewEvent(
			ctx.Registry,
			ctx.RunRegistry,
			ctx.StateRegistry,
			sessionID,
			route,
			targetID,
			tabIndex,
			finalURL,
			title,
			decision,
			reason,
			ctx.ReconnectWindow,
		)
	}
	if targetID == "" {
		targetID = ResolveSharedSessionBrowserTabTargetID(ctx.Registry, sessionID, route, tabIndex)
	}
	if targetID == "" {
		return nil
	}
	review := BrowserSessionTargetReview{
		ID:         targetID,
		TabIndex:   tabIndex,
		URL:        strings.TrimSpace(finalURL),
		Title:      strings.TrimSpace(title),
		BrowserApp: strings.TrimSpace(route.BrowserApp),
		Backend:    strings.TrimSpace(route.Backend),
		Profile:    strings.TrimSpace(route.Profile),
		Target:     strings.TrimSpace(route.Target),
		Decision:   strings.TrimSpace(decision),
		Reason:     strings.TrimSpace(reason),
	}
	ctx.Registry.RecordPendingTargetReviewForRoute(sessionID, route, review)
	return &review
}

// RecordSharedSessionBrowserPendingTargetPopupReview records a pending popup
// review posture for a tracked active tab.
func RecordSharedSessionBrowserPendingTargetPopupReview(
	registry *BrowserSessionRegistry,
	sessionID string,
	route BrowserSessionRoute,
	activeTab BrowserTab,
	decision string,
	reason string,
) *BrowserSessionTargetReview {
	return RecordSharedSessionBrowserPendingTargetPopupReviewWithContext(
		SharedSessionBrowserMutationContext{Registry: registry},
		sessionID,
		route,
		activeTab,
		decision,
		reason,
	)
}

// RecordSharedSessionBrowserPendingTargetPopupReviewWithContext records a
// pending popup review posture for a tracked active tab and routes the write
// through the shared mutation seam when manager dependencies are available.
func RecordSharedSessionBrowserPendingTargetPopupReviewWithContext(
	ctx SharedSessionBrowserMutationContext,
	sessionID string,
	route BrowserSessionRoute,
	activeTab BrowserTab,
	decision string,
	reason string,
) *BrowserSessionTargetReview {
	if ctx.usesWatchManagerEventSeam() {
		return RecordSharedSessionBrowserPendingTargetPopupReviewEvent(
			ctx.Registry,
			ctx.RunRegistry,
			ctx.StateRegistry,
			sessionID,
			route,
			activeTab,
			decision,
			reason,
			ctx.ReconnectWindow,
		)
	}
	return RecordSharedSessionBrowserPendingTargetReviewWithContext(
		ctx,
		sessionID,
		route,
		strings.TrimSpace(activeTab.TargetID),
		activeTab.Index,
		strings.TrimSpace(activeTab.URL),
		strings.TrimSpace(activeTab.Title),
		decision,
		reason,
	)
}

// SharedSessionBrowserPendingTargetReviewStateForTarget returns the pending
// review posture for a concrete target scoped to the selected route.
func SharedSessionBrowserPendingTargetReviewStateForTarget(registry *BrowserSessionRegistry, sessionID string, route BrowserSessionRoute, targetID string, tabIndex int) SharedSessionBrowserPendingTargetReviewState {
	if registry == nil || strings.TrimSpace(sessionID) == "" {
		return SharedSessionBrowserPendingTargetReviewState{}
	}
	targetID = strings.TrimSpace(targetID)
	var fallback SharedSessionBrowserPendingTargetReviewState
	fallbackAmbiguous := false
	for _, routeState := range registry.Snapshot(strings.TrimSpace(sessionID)) {
		if routeState.PendingTargetReview == nil || !sharedSessionBrowserPendingTargetReviewMatchesTarget(*routeState.PendingTargetReview, targetID, tabIndex) {
			continue
		}
		state := SharedSessionBrowserPendingTargetReviewState{
			Review: routeState.PendingTargetReview,
			Count:  routeState.PendingTargetReviewCount,
		}
		state.PolicyState, state.PolicyReason = SharedSessionBrowserPendingTargetReviewPolicy(state.Review, state.Count)
		if browserSessionRouteMatchesFilter(routeState.Route, route) {
			return state
		}
		if fallback.Review == nil {
			fallback = state
			continue
		}
		if !strings.EqualFold(strings.TrimSpace(fallback.Review.ID), strings.TrimSpace(routeState.PendingTargetReview.ID)) {
			fallbackAmbiguous = true
		}
	}
	if fallbackAmbiguous {
		return SharedSessionBrowserPendingTargetReviewState{}
	}
	return fallback
}

// SharedSessionBrowserPendingTargetReviewStateForRoute returns the pending
// review posture for a scoped route regardless of a specific target request.
func SharedSessionBrowserPendingTargetReviewStateForRoute(registry *BrowserSessionRegistry, sessionID string, route BrowserSessionRoute) SharedSessionBrowserPendingTargetReviewState {
	if registry == nil || strings.TrimSpace(sessionID) == "" {
		return SharedSessionBrowserPendingTargetReviewState{}
	}
	var matched SharedSessionBrowserPendingTargetReviewState
	ambiguous := false
	for _, routeState := range registry.Snapshot(strings.TrimSpace(sessionID)) {
		if routeState.PendingTargetReview == nil || !browserSessionRouteMatchesFilter(routeState.Route, route) {
			continue
		}
		state := SharedSessionBrowserPendingTargetReviewState{
			Review: routeState.PendingTargetReview,
			Count:  routeState.PendingTargetReviewCount,
		}
		state.PolicyState, state.PolicyReason = SharedSessionBrowserPendingTargetReviewPolicy(state.Review, state.Count)
		if matched.Review == nil {
			matched = state
			continue
		}
		if !strings.EqualFold(strings.TrimSpace(matched.Review.ID), strings.TrimSpace(routeState.PendingTargetReview.ID)) {
			ambiguous = true
		}
	}
	if ambiguous {
		return SharedSessionBrowserPendingTargetReviewState{}
	}
	return matched
}

// SharedSessionBrowserAutoFollowPendingTargetReviewState returns the
// route-scoped pending-review posture used by auto-follow decisions.
func SharedSessionBrowserAutoFollowPendingTargetReviewState(registry *BrowserSessionRegistry, sessionID string, route BrowserSessionRoute, targetID string, tabIndex int) SharedSessionBrowserPendingTargetReviewState {
	state := SharedSessionBrowserPendingTargetReviewStateForTarget(registry, sessionID, route, targetID, tabIndex)
	if state.Review != nil {
		return state
	}
	if strings.TrimSpace(targetID) != "" || tabIndex > 0 {
		return state
	}
	return SharedSessionBrowserPendingTargetReviewStateForRoute(registry, sessionID, route)
}

// SharedSessionBrowserPendingTargetReviewDecision resolves the decision token
// for a pending target review posture.
func SharedSessionBrowserPendingTargetReviewDecision(state SharedSessionBrowserPendingTargetReviewState, force bool) string {
	if state.Review == nil {
		return ""
	}
	switch SharedSessionBrowserPendingTargetReviewCategory(state.Review) {
	case "redirect":
		if force {
			return "session_target_redirect_review_confirmed"
		}
		return "session_target_redirect_review_required"
	default:
		if force {
			return "session_target_popup_review_confirmed"
		}
		return "session_target_popup_review_required"
	}
}

// SharedSessionBrowserSelectedTargetDecision resolves the final decision token
// for a target selection after considering pending-review confirmation.
func SharedSessionBrowserSelectedTargetDecision(state SharedSessionBrowserPendingTargetReviewState, force bool) string {
	if state.Review != nil && force {
		return SharedSessionBrowserPendingTargetReviewDecision(state, force)
	}
	return "session_target_selected"
}

// SharedSessionBrowserPendingTargetReviewReason returns the user-facing reason
// string for a pending target review posture.
func SharedSessionBrowserPendingTargetReviewReason(actor string, state SharedSessionBrowserPendingTargetReviewState, force bool) string {
	if state.Review == nil {
		return ""
	}
	label := SharedSessionBrowserTargetReviewLabel(state.Review)
	actor = firstNonEmptyString(strings.TrimSpace(actor), "browser action")
	if SharedSessionBrowserPendingTargetReviewCategory(state.Review) == "redirect" {
		if force {
			return fmt.Sprintf("%s redirect review acknowledged via force=true; confirm the redirected target %q is intended before continuing", actor, label)
		}
		return fmt.Sprintf("%s hit redirected target %q after cross-origin navigation; rerun with force=true before adopting or following it", actor, label)
	}
	if force {
		return fmt.Sprintf("%s popup review acknowledged via force=true; confirm the pending popup target %q is intended before continuing", actor, label)
	}
	if strings.EqualFold(strings.TrimSpace(state.PolicyState), "popup_storm_review_required") && strings.TrimSpace(state.PolicyReason) != "" {
		return fmt.Sprintf("%s: %s", actor, strings.TrimSpace(state.PolicyReason))
	}
	return fmt.Sprintf("%s hit pending popup target %q; rerun with force=true before adopting or following it", actor, label)
}

func sharedSessionBrowserPendingTargetReviewMatchesTarget(review BrowserSessionTargetReview, targetID string, tabIndex int) bool {
	if targetID != "" && strings.EqualFold(strings.TrimSpace(review.ID), targetID) {
		return true
	}
	return tabIndex > 0 && review.TabIndex == tabIndex
}
