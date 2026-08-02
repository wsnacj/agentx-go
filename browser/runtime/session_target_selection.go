package browserruntime

import (
	"fmt"
	"strings"
)

// SharedSessionBrowserSelectTargetRequest carries the route-scoped target
// selection inputs shared by browser runtime control actions.
type SharedSessionBrowserSelectTargetRequest struct {
	SessionID string
	Route     BrowserSessionRoute
	TargetID  string
	TabIndex  int
	Current   bool
	Force     bool
	Source    string
	Actor     string
}

// SharedSessionBrowserSelectTargetResult captures the lifecycle-owned outcome
// of a route-scoped target selection.
type SharedSessionBrowserSelectTargetResult struct {
	Selection *BrowserSessionTargetSelection
	Decision  string
	Ready     bool
	Note      string
}

// SharedSessionBrowserRememberTargetRequest carries the route-scoped target
// remember inputs shared by browser action/runtime tools.
type SharedSessionBrowserRememberTargetRequest struct {
	SessionID string
	Route     BrowserSessionRoute
	TargetID  string
	TabIndex  int
	Source    string
}

// SharedSessionBrowserRememberTargetResult captures the lifecycle-owned outcome
// of remembering a route-scoped target selection.
type SharedSessionBrowserRememberTargetResult struct {
	Selection *BrowserSessionTargetSelection
	Decision  string
	Ready     bool
}

// CurrentSharedSessionBrowserTargetSelection returns the current selected target
// for a scoped session route as a shared target selection payload.
func CurrentSharedSessionBrowserTargetSelection(registry *BrowserSessionRegistry, sessionID string, route BrowserSessionRoute) *BrowserSessionTargetSelection {
	sessionID = strings.TrimSpace(sessionID)
	if registry == nil || sessionID == "" {
		return nil
	}
	registry.PruneStaleRouteState(sessionID, route)
	target, source, ok := registry.CurrentTargetSelectionForRoute(sessionID, route)
	if !ok {
		return nil
	}
	return sharedSessionBrowserTargetSelectionFromTracked(target, source)
}

// ClearSharedSessionBrowserTargetSelection clears the current selected target
// for a scoped session route.
func ClearSharedSessionBrowserTargetSelectionWithContext(
	ctx SharedSessionBrowserMutationContext,
	sessionID string,
	route BrowserSessionRoute,
) bool {
	sessionID = strings.TrimSpace(sessionID)
	if ctx.Registry == nil || sessionID == "" {
		return false
	}
	if ctx.usesWatchManagerEventSeam() {
		req := BuildSharedSessionBrowserClearRequest(
			ctx.Registry,
			ctx.StateRegistry,
			sessionID,
			BrowserRuntimeInfo{
				Backend: strings.TrimSpace(route.Backend),
				Profile: strings.TrimSpace(route.Profile),
				Target:  strings.TrimSpace(route.Target),
			},
			route,
			false,
			"",
			SharedSessionBrowserHealthInput{},
			ctx.ReconnectWindow,
		)
		result := ExecuteSharedSessionBrowserClearTargetEvent(
			ctx.Registry,
			ctx.RunRegistry,
			ctx.StateRegistry,
			req,
			ctx.ReconnectWindow,
		)
		return result.ClearedTargetSelection
	}
	return ctx.Registry.ClearCurrentTargetForRoute(sessionID, route)
}

func ClearSharedSessionBrowserTargetSelection(registry *BrowserSessionRegistry, sessionID string, route BrowserSessionRoute) bool {
	return ClearSharedSessionBrowserTargetSelectionWithContext(
		SharedSessionBrowserMutationContext{Registry: registry},
		sessionID,
		route,
	)
}

// SyncSharedSessionBrowserCurrentTarget selects the current tracked target for
// the scoped session route, preserving the route-local source semantics.
func SyncSharedSessionBrowserCurrentTarget(registry *BrowserSessionRegistry, sessionID string, route BrowserSessionRoute, source string) (*BrowserSessionTargetSelection, string, error) {
	sessionID = strings.TrimSpace(sessionID)
	if registry == nil || sessionID == "" {
		return nil, "", nil
	}
	registry.PruneStaleRouteState(sessionID, route)
	current, currentSource, ok := registry.CurrentTargetSelectionForRoute(sessionID, route)
	if !ok || strings.TrimSpace(current.ID) == "" {
		return nil, "", nil
	}
	source = firstNonEmptyString(strings.TrimSpace(source), "sync_session")
	if strings.EqualFold(strings.TrimSpace(currentSource), source) {
		return sharedSessionBrowserTargetSelectionFromTracked(current, source), "session_target_already_selected", nil
	}
	selected, ok := registry.SelectTargetForRoute(sessionID, route, current.ID, source)
	if !ok {
		return nil, "", fmt.Errorf("browser_runtime: current target is not available on the selected route")
	}
	return sharedSessionBrowserTargetSelectionFromTracked(selected, source), "session_target_selected", nil
}

// SelectSharedSessionBrowserTarget applies the shared route-scoped target
// selection contract, including pending-review confirmation posture and
// route-aware target fallback.
func SelectSharedSessionBrowserTargetWithContext(
	ctx SharedSessionBrowserMutationContext,
	req SharedSessionBrowserSelectTargetRequest,
) (SharedSessionBrowserSelectTargetResult, error) {
	req.SessionID = strings.TrimSpace(req.SessionID)
	req.Route = normalizeBrowserSessionRoute(req.Route)
	req.TargetID = strings.TrimSpace(req.TargetID)
	req.Source = firstNonEmptyString(strings.TrimSpace(req.Source), "select_target")
	req.Actor = firstNonEmptyString(strings.TrimSpace(req.Actor), "browser_runtime target selection")
	if ctx.usesWatchManagerEventSeam() {
		return SelectSharedSessionBrowserTargetEvent(
			ctx.Registry,
			ctx.RunRegistry,
			ctx.StateRegistry,
			req,
			ctx.ReconnectWindow,
		)
	}
	registry := ctx.Registry
	if registry == nil || req.SessionID == "" {
		return SharedSessionBrowserSelectTargetResult{}, nil
	}
	registry.PruneStaleRouteState(req.SessionID, req.Route)
	switch {
	case req.Current:
		current, ok := registry.CurrentTargetForRoute(req.SessionID, req.Route)
		if !ok || strings.TrimSpace(current.ID) == "" {
			return SharedSessionBrowserSelectTargetResult{Decision: "session_target_missing"}, fmt.Errorf("browser_runtime: current target is not available on the selected route")
		}
		if selected, ok := registry.SelectTargetForRoute(req.SessionID, req.Route, current.ID, req.Source); ok {
			current = selected
		}
		return SharedSessionBrowserSelectTargetResult{
			Selection: sharedSessionBrowserTargetSelectionFromTracked(current, req.Source),
			Decision:  "session_target_already_selected",
			Ready:     true,
		}, nil
	case req.TargetID != "":
		pendingReview := SharedSessionBrowserPendingTargetReviewStateForTarget(registry, req.SessionID, req.Route, req.TargetID, 0)
		if pendingReview.Review != nil && !req.Force {
			return SharedSessionBrowserSelectTargetResult{
				Decision: SharedSessionBrowserPendingTargetReviewDecision(pendingReview, req.Force),
				Note:     SharedSessionBrowserPendingTargetReviewReason(req.Actor, pendingReview, req.Force),
			}, nil
		}
		if current, ok := registry.CurrentTargetForRoute(req.SessionID, req.Route); ok && strings.EqualFold(strings.TrimSpace(current.ID), req.TargetID) {
			if selected, ok := registry.SelectTargetForRoute(req.SessionID, req.Route, current.ID, req.Source); ok {
				current = selected
			}
			return SharedSessionBrowserSelectTargetResult{
				Selection: sharedSessionBrowserTargetSelectionFromTracked(current, req.Source),
				Decision:  "session_target_already_selected",
				Ready:     true,
			}, nil
		}
		selected, ok := registry.SelectTargetForRoute(req.SessionID, req.Route, req.TargetID, req.Source)
		if !ok {
			resolved, resolvedOK := registry.ResolveTarget(req.SessionID, req.TargetID)
			if !resolvedOK || !sharedSessionBrowserTrackedTargetMatchesRoute(resolved, req.Route) {
				return SharedSessionBrowserSelectTargetResult{Decision: "session_target_missing"}, fmt.Errorf("browser_runtime: target %q is not available on the selected route", req.TargetID)
			}
			selected, ok = registry.SelectTargetForRoute(req.SessionID, browserSessionRouteFromTarget(resolved), req.TargetID, req.Source)
			if !ok {
				return SharedSessionBrowserSelectTargetResult{Decision: "session_target_missing"}, fmt.Errorf("browser_runtime: target %q is not available on the selected route", req.TargetID)
			}
		}
		return SharedSessionBrowserSelectTargetResult{
			Selection: sharedSessionBrowserTargetSelectionFromTracked(selected, req.Source),
			Decision:  SharedSessionBrowserSelectedTargetDecision(pendingReview, req.Force),
			Ready:     true,
			Note:      SharedSessionBrowserPendingTargetReviewReason(req.Actor, pendingReview, req.Force),
		}, nil
	case req.TabIndex > 0:
		pendingReview := SharedSessionBrowserPendingTargetReviewStateForTarget(registry, req.SessionID, req.Route, "", req.TabIndex)
		if pendingReview.Review != nil && !req.Force {
			return SharedSessionBrowserSelectTargetResult{
				Decision: SharedSessionBrowserPendingTargetReviewDecision(pendingReview, req.Force),
				Note:     SharedSessionBrowserPendingTargetReviewReason(req.Actor, pendingReview, req.Force),
			}, nil
		}
		if current, ok := registry.CurrentTargetForRoute(req.SessionID, req.Route); ok && current.TabIndex == req.TabIndex {
			if selected, ok := registry.SelectTabForRoute(req.SessionID, req.Route, current.TabIndex, req.Source); ok {
				current = selected
			}
			return SharedSessionBrowserSelectTargetResult{
				Selection: sharedSessionBrowserTargetSelectionFromTracked(current, req.Source),
				Decision:  "session_target_already_selected",
				Ready:     true,
			}, nil
		}
		selected, ok := registry.SelectTabForRoute(req.SessionID, req.Route, req.TabIndex, req.Source)
		if !ok {
			return SharedSessionBrowserSelectTargetResult{Decision: "session_target_missing"}, fmt.Errorf("browser_runtime: tab %d is not available on the selected route", req.TabIndex)
		}
		return SharedSessionBrowserSelectTargetResult{
			Selection: sharedSessionBrowserTargetSelectionFromTracked(selected, req.Source),
			Decision:  SharedSessionBrowserSelectedTargetDecision(pendingReview, req.Force),
			Ready:     true,
			Note:      SharedSessionBrowserPendingTargetReviewReason(req.Actor, pendingReview, req.Force),
		}, nil
	default:
		return SharedSessionBrowserSelectTargetResult{}, nil
	}
}

func SelectSharedSessionBrowserTarget(registry *BrowserSessionRegistry, req SharedSessionBrowserSelectTargetRequest) (SharedSessionBrowserSelectTargetResult, error) {
	return SelectSharedSessionBrowserTargetWithContext(
		SharedSessionBrowserMutationContext{Registry: registry},
		req,
	)
}

// RememberSharedSessionBrowserTarget persists a route-scoped target selection
// while preserving the current-target fast path and default-route fallback used
// by browser action/runtime remember_target flows.
func RememberSharedSessionBrowserTargetWithContext(
	ctx SharedSessionBrowserMutationContext,
	req SharedSessionBrowserRememberTargetRequest,
) SharedSessionBrowserRememberTargetResult {
	req.SessionID = strings.TrimSpace(req.SessionID)
	req.Route = normalizeBrowserSessionRoute(req.Route)
	req.TargetID = strings.TrimSpace(req.TargetID)
	req.Source = firstNonEmptyString(strings.TrimSpace(req.Source), "remember_target")
	if ctx.usesWatchManagerEventSeam() {
		return RememberSharedSessionBrowserTargetEvent(
			ctx.Registry,
			ctx.RunRegistry,
			ctx.StateRegistry,
			req,
			ctx.ReconnectWindow,
		)
	}
	registry := ctx.Registry
	if registry == nil || req.SessionID == "" {
		return SharedSessionBrowserRememberTargetResult{
			Decision: "session_target_not_remembered",
		}
	}

	registry.PruneStaleRouteState(req.SessionID, req.Route)
	registry.PruneStaleRouteState(req.SessionID, BrowserSessionRoute{})

	if req.TargetID != "" {
		if current, _, ok := registry.CurrentTargetSelectionForRoute(req.SessionID, req.Route); ok &&
			strings.EqualFold(strings.TrimSpace(current.ID), req.TargetID) {
			if selected, ok := registry.SelectTargetForRoute(req.SessionID, req.Route, current.ID, req.Source); ok {
				return SharedSessionBrowserRememberTargetResult{
					Selection: sharedSessionBrowserTargetSelectionFromTracked(selected, req.Source),
					Decision:  "session_target_already_remembered",
					Ready:     true,
				}
			}
			if selected, ok := registry.SelectTargetForRoute(req.SessionID, BrowserSessionRoute{}, current.ID, req.Source); ok {
				return SharedSessionBrowserRememberTargetResult{
					Selection: sharedSessionBrowserTargetSelectionFromTracked(selected, req.Source),
					Decision:  "session_target_already_remembered",
					Ready:     true,
				}
			}
		}
		if selected, ok := registry.SelectTargetForRoute(req.SessionID, req.Route, req.TargetID, req.Source); ok {
			return SharedSessionBrowserRememberTargetResult{
				Selection: sharedSessionBrowserTargetSelectionFromTracked(selected, req.Source),
				Decision:  "session_target_remembered",
				Ready:     true,
			}
		}
		if selected, ok := registry.SelectTargetForRoute(req.SessionID, BrowserSessionRoute{}, req.TargetID, req.Source); ok {
			return SharedSessionBrowserRememberTargetResult{
				Selection: sharedSessionBrowserTargetSelectionFromTracked(selected, req.Source),
				Decision:  "session_target_remembered",
				Ready:     true,
			}
		}
		return SharedSessionBrowserRememberTargetResult{
			Decision: "session_target_not_remembered",
		}
	}

	if req.TabIndex <= 0 {
		return SharedSessionBrowserRememberTargetResult{
			Decision: "session_target_not_remembered",
		}
	}

	if current, _, ok := registry.CurrentTargetSelectionForRoute(req.SessionID, req.Route); ok && current.TabIndex == req.TabIndex {
		if selected, ok := registry.SelectTabForRoute(req.SessionID, req.Route, current.TabIndex, req.Source); ok {
			return SharedSessionBrowserRememberTargetResult{
				Selection: sharedSessionBrowserTargetSelectionFromTracked(selected, req.Source),
				Decision:  "session_target_already_remembered",
				Ready:     true,
			}
		}
		if selected, ok := registry.SelectTabForRoute(req.SessionID, BrowserSessionRoute{}, current.TabIndex, req.Source); ok {
			return SharedSessionBrowserRememberTargetResult{
				Selection: sharedSessionBrowserTargetSelectionFromTracked(selected, req.Source),
				Decision:  "session_target_already_remembered",
				Ready:     true,
			}
		}
	}
	if selected, ok := registry.SelectTabForRoute(req.SessionID, req.Route, req.TabIndex, req.Source); ok {
		return SharedSessionBrowserRememberTargetResult{
			Selection: sharedSessionBrowserTargetSelectionFromTracked(selected, req.Source),
			Decision:  "session_target_remembered",
			Ready:     true,
		}
	}
	if selected, ok := registry.SelectTabForRoute(req.SessionID, BrowserSessionRoute{}, req.TabIndex, req.Source); ok {
		return SharedSessionBrowserRememberTargetResult{
			Selection: sharedSessionBrowserTargetSelectionFromTracked(selected, req.Source),
			Decision:  "session_target_remembered",
			Ready:     true,
		}
	}
	return SharedSessionBrowserRememberTargetResult{
		Decision: "session_target_not_remembered",
	}
}

func RememberSharedSessionBrowserTarget(registry *BrowserSessionRegistry, req SharedSessionBrowserRememberTargetRequest) SharedSessionBrowserRememberTargetResult {
	return RememberSharedSessionBrowserTargetWithContext(
		SharedSessionBrowserMutationContext{Registry: registry},
		req,
	)
}

// SyncOrClearSharedSessionBrowserCurrentTargetForProfileSelection synchronizes
// the current target for a route-scoped profile selection or clears a stale
// mismatched target selection when the remembered target belongs to a different
// managed profile.
func SyncOrClearSharedSessionBrowserCurrentTargetForProfileSelectionWithContext(
	ctx SharedSessionBrowserMutationContext,
	sessionID string,
	route BrowserSessionRoute,
	profileSelection *SharedSessionBrowserProfileSelection,
	source string,
) (*BrowserSessionTargetSelection, string, error) {
	sessionID = strings.TrimSpace(sessionID)
	route = normalizeBrowserSessionRoute(route)
	if ctx.usesWatchManagerEventSeam() {
		return SyncOrClearSharedSessionBrowserCurrentTargetForProfileSelectionEvent(
			ctx.Registry,
			ctx.RunRegistry,
			ctx.StateRegistry,
			sessionID,
			route,
			profileSelection,
			source,
			ctx.ReconnectWindow,
		)
	}
	return SyncOrClearSharedSessionBrowserCurrentTargetForProfileSelection(
		ctx.Registry,
		sessionID,
		route,
		profileSelection,
		source,
	)
}

// SyncOrClearSharedSessionBrowserCurrentTargetForProfileSelection synchronizes
// the current target for a route-scoped profile selection or clears a stale
// mismatched target selection when the remembered target belongs to a different
// managed profile.
func SyncOrClearSharedSessionBrowserCurrentTargetForProfileSelection(registry *BrowserSessionRegistry, sessionID string, route BrowserSessionRoute, profileSelection *SharedSessionBrowserProfileSelection, source string) (*BrowserSessionTargetSelection, string, error) {
	targetSelection, targetDecision, err := SyncSharedSessionBrowserCurrentTarget(registry, sessionID, route, source)
	if err != nil || targetSelection != nil {
		return targetSelection, targetDecision, err
	}
	if ShouldClearSharedSessionBrowserTargetOnProfileSelect(registry, sessionID, route, profileSelection) {
		if ClearSharedSessionBrowserTargetSelection(registry, sessionID, route) {
			return nil, "session_target_cleared_mismatched_profile", nil
		}
	}
	return nil, targetDecision, nil
}

// ShouldClearSharedSessionBrowserTargetOnProfileClear reports whether clearing
// the current session profile should also clear the route-scoped target
// selection.
func ShouldClearSharedSessionBrowserTargetOnProfileClear(registry *BrowserSessionRegistry, sessionID string, route BrowserSessionRoute) bool {
	selection := CurrentSharedSessionBrowserTargetSelection(registry, sessionID, route)
	if selection == nil || !sharedSessionBrowserTargetSelectionMatchesRoute(selection, route) {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(selection.Source)) {
	case "select_profile", "sync_session", "remember_profile":
		return true
	default:
		return false
	}
}

// ShouldClearSharedSessionBrowserTargetOnProfileSelect reports whether a stale
// current target should be cleared when a new session profile selection no
// longer matches the target's remembered profile.
func ShouldClearSharedSessionBrowserTargetOnProfileSelect(registry *BrowserSessionRegistry, sessionID string, route BrowserSessionRoute, profileSelection *SharedSessionBrowserProfileSelection) bool {
	selection := CurrentSharedSessionBrowserTargetSelection(registry, sessionID, route)
	if selection == nil || !sharedSessionBrowserTargetSelectionMatchesRoute(selection, route) || profileSelection == nil {
		return false
	}
	if strings.TrimSpace(selection.Profile) == "" || strings.TrimSpace(profileSelection.Profile) == "" {
		return false
	}
	return !strings.EqualFold(strings.TrimSpace(selection.Profile), strings.TrimSpace(profileSelection.Profile))
}

// ShouldClearSharedSessionBrowserProfileOnTargetClear reports whether clearing
// the current route-scoped target selection should also clear the remembered
// managed profile selection for the same route.
func ShouldClearSharedSessionBrowserProfileOnTargetClear(registry *BrowserSessionRegistry, stateRegistry SharedSessionBrowserStateRegistry, sessionID string, route BrowserSessionRoute) bool {
	targetSelection := CurrentSharedSessionBrowserTargetSelection(registry, sessionID, route)
	if targetSelection == nil || !sharedSessionBrowserTargetSelectionMatchesRoute(targetSelection, route) {
		return false
	}
	targetSource := strings.TrimSpace(targetSelection.Source)
	switch {
	case strings.EqualFold(targetSource, "select_target"):
	case strings.EqualFold(targetSource, "remember_target"):
	case strings.EqualFold(targetSource, "remember_profile"):
	default:
		return false
	}
	if stateRegistry == nil || sessionID == "" {
		return false
	}
	profileSelection, ok := stateRegistry.SelectedBrowserProfile(strings.TrimSpace(sessionID), strings.TrimSpace(route.Target))
	if !ok || !sharedSessionBrowserProfileSelectionMatchesTarget(profileSelection, targetSelection.Backend, targetSelection.RuntimeTarget, targetSelection.Profile) {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(profileSelection.Source), targetSource)
}

func sharedSessionBrowserTargetSelectionFromTracked(target BrowserSessionTarget, source string) *BrowserSessionTargetSelection {
	target.ID = strings.TrimSpace(target.ID)
	target.URL = strings.TrimSpace(target.URL)
	target.Title = strings.TrimSpace(target.Title)
	target.BrowserApp = strings.TrimSpace(target.BrowserApp)
	target.Backend = strings.TrimSpace(target.Backend)
	target.Profile = strings.TrimSpace(target.Profile)
	target.Target = strings.TrimSpace(target.Target)
	if target.ID == "" && target.TabIndex <= 0 && target.URL == "" && target.Title == "" {
		return nil
	}
	return &BrowserSessionTargetSelection{
		ID:            strings.TrimSpace(target.ID),
		TabIndex:      target.TabIndex,
		URL:           strings.TrimSpace(target.URL),
		Title:         strings.TrimSpace(target.Title),
		Backend:       strings.TrimSpace(target.Backend),
		Profile:       strings.TrimSpace(target.Profile),
		RuntimeTarget: strings.TrimSpace(target.Target),
		BrowserApp:    strings.TrimSpace(target.BrowserApp),
		Source:        strings.TrimSpace(source),
	}
}

func sharedSessionBrowserTargetSelectionMatchesRoute(selection *BrowserSessionTargetSelection, route BrowserSessionRoute) bool {
	if selection == nil {
		return false
	}
	if selected := strings.TrimSpace(selection.Backend); selected != "" && strings.TrimSpace(route.Backend) != "" && !browserSessionSameLogicalRoute(BrowserSessionRoute{Backend: selected}, BrowserSessionRoute{Backend: strings.TrimSpace(route.Backend)}) {
		return false
	}
	if selected := strings.TrimSpace(selection.RuntimeTarget); selected != "" && strings.TrimSpace(route.Target) != "" && !strings.EqualFold(selected, strings.TrimSpace(route.Target)) {
		return false
	}
	if selected := strings.TrimSpace(selection.Profile); selected != "" && strings.TrimSpace(route.Profile) != "" && !strings.EqualFold(selected, strings.TrimSpace(route.Profile)) {
		return false
	}
	return true
}

func sharedSessionBrowserProfileSelectionMatchesTarget(selection SharedSessionBrowserProfileSelection, backend string, runtimeTarget string, profile string) bool {
	return (strings.TrimSpace(selection.Backend) == "" || strings.TrimSpace(backend) == "" ||
		browserSessionSameLogicalRoute(BrowserSessionRoute{Backend: strings.TrimSpace(selection.Backend)}, BrowserSessionRoute{Backend: strings.TrimSpace(backend)})) &&
		(strings.TrimSpace(selection.RuntimeTarget) == "" || strings.TrimSpace(runtimeTarget) == "" ||
			strings.EqualFold(strings.TrimSpace(selection.RuntimeTarget), strings.TrimSpace(runtimeTarget))) &&
		(strings.TrimSpace(selection.Profile) == "" || strings.TrimSpace(profile) == "" ||
			strings.EqualFold(strings.TrimSpace(selection.Profile), strings.TrimSpace(profile)))
}

func sharedSessionBrowserTrackedTargetMatchesRoute(target BrowserSessionTarget, route BrowserSessionRoute) bool {
	if selected := strings.TrimSpace(route.Backend); selected != "" && !browserSessionSameLogicalRoute(BrowserSessionRoute{Backend: selected}, BrowserSessionRoute{Backend: strings.TrimSpace(target.Backend)}) {
		return false
	}
	if selected := strings.TrimSpace(route.Target); selected != "" && !strings.EqualFold(selected, strings.TrimSpace(target.Target)) {
		return false
	}
	if selected := strings.TrimSpace(route.BrowserApp); selected != "" && !strings.EqualFold(selected, strings.TrimSpace(target.BrowserApp)) {
		return false
	}
	if selected := strings.TrimSpace(route.Profile); selected != "" && !strings.EqualFold(selected, strings.TrimSpace(target.Profile)) {
		return false
	}
	return true
}
