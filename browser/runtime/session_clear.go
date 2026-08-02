package browserruntime

import (
	"strings"
	"time"
)

// SharedSessionBrowserClearRequest carries the scoped session state needed to
// execute clear-profile / clear-target / clear-session actions through the
// shared browserruntime contract.
type SharedSessionBrowserClearRequest struct {
	SessionRegistry *BrowserSessionRegistry
	StateRegistry   SharedSessionBrowserStateRegistry
	SessionID       string
	SelectedInfo    BrowserRuntimeInfo
	Route           BrowserSessionRoute
	Force           bool
	ActiveNodeRunID string
	HealthInput     SharedSessionBrowserHealthInput
	ReconnectWindow time.Duration
}

// SharedSessionBrowserClearResult captures the lifecycle-owned decision and
// cleanup outcomes for shared session state clear actions.
type SharedSessionBrowserClearResult struct {
	Decision                string
	Ready                   bool
	ProfileStatus           BrowserProfileStatusResult
	ProfileState            SharedSessionBrowserProfileState
	HasProfileState         bool
	ClearedProfileSelection bool
	ClearedTargetSelection  bool
	ClearedSessionProfiles  int
	ClearedSessionTargets   int
}

// BuildSharedSessionBrowserClearRequest assembles the shared clear-action
// request used by session-scoped selection/route cleanup actions.
func BuildSharedSessionBrowserClearRequest(
	sessionRegistry *BrowserSessionRegistry,
	stateRegistry SharedSessionBrowserStateRegistry,
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	route BrowserSessionRoute,
	force bool,
	activeNodeRunID string,
	healthInput SharedSessionBrowserHealthInput,
	reconnectWindow time.Duration,
) SharedSessionBrowserClearRequest {
	return SharedSessionBrowserClearRequest{
		SessionRegistry: sessionRegistry,
		StateRegistry:   stateRegistry,
		SessionID:       strings.TrimSpace(sessionID),
		SelectedInfo:    selectedInfo,
		Route:           normalizeBrowserSessionRoute(route),
		Force:           force,
		ActiveNodeRunID: strings.TrimSpace(activeNodeRunID),
		HealthInput:     healthInput,
		ReconnectWindow: reconnectWindow,
	}
}

// ExecuteSharedSessionBrowserClearProfile applies the shared clear-profile
// action contract for a scoped managed-browser route.
func ExecuteSharedSessionBrowserClearProfileWithContext(
	ctx SharedSessionBrowserMutationContext,
	req SharedSessionBrowserClearRequest,
) SharedSessionBrowserClearResult {
	if ctx.Registry != nil {
		req.SessionRegistry = ctx.Registry
	}
	if ctx.StateRegistry != nil {
		req.StateRegistry = ctx.StateRegistry
	}
	if ctx.usesWatchManagerEventSeam() {
		return ExecuteSharedSessionBrowserClearProfileEvent(
			ctx.Registry,
			ctx.RunRegistry,
			ctx.StateRegistry,
			req,
			ctx.ReconnectWindow,
		)
	}
	if blocked, ok := blockedSharedSessionBrowserClearResult(req, "clear_profile_blocked_active_node_run"); ok {
		return blocked
	}
	result := SharedSessionBrowserClearResult{
		Decision: "session_profile_already_clear",
		Ready:    true,
	}
	result.ClearedProfileSelection = clearSharedSessionBrowserProfileSelectionForRoute(req.StateRegistry, req.SessionID, req.Route)
	if result.ClearedProfileSelection && ShouldClearSharedSessionBrowserTargetOnProfileClear(req.SessionRegistry, req.SessionID, req.Route) {
		result.ClearedTargetSelection = ClearSharedSessionBrowserTargetSelection(req.SessionRegistry, req.SessionID, req.Route)
	}
	if result.ClearedProfileSelection {
		result.Decision = "session_profile_cleared"
	}
	return result
}

func ExecuteSharedSessionBrowserClearProfile(req SharedSessionBrowserClearRequest) SharedSessionBrowserClearResult {
	return ExecuteSharedSessionBrowserClearProfileWithContext(
		SharedSessionBrowserMutationContext{
			Registry:      req.SessionRegistry,
			StateRegistry: req.StateRegistry,
		},
		req,
	)
}

// ExecuteSharedSessionBrowserClearTarget applies the shared clear-target action
// contract for a scoped managed-browser route.
func ExecuteSharedSessionBrowserClearTargetWithContext(
	ctx SharedSessionBrowserMutationContext,
	req SharedSessionBrowserClearRequest,
) SharedSessionBrowserClearResult {
	if ctx.Registry != nil {
		req.SessionRegistry = ctx.Registry
	}
	if ctx.StateRegistry != nil {
		req.StateRegistry = ctx.StateRegistry
	}
	if ctx.usesWatchManagerEventSeam() {
		return ExecuteSharedSessionBrowserClearTargetEvent(
			ctx.Registry,
			ctx.RunRegistry,
			ctx.StateRegistry,
			req,
			ctx.ReconnectWindow,
		)
	}
	if blocked, ok := blockedSharedSessionBrowserClearResult(req, "clear_target_blocked_active_node_run"); ok {
		return blocked
	}
	result := SharedSessionBrowserClearResult{
		Decision: "session_target_already_clear",
		Ready:    true,
	}
	if ShouldClearSharedSessionBrowserProfileOnTargetClear(req.SessionRegistry, req.StateRegistry, req.SessionID, req.Route) {
		result.ClearedProfileSelection = clearSharedSessionBrowserProfileSelectionForRoute(req.StateRegistry, req.SessionID, req.Route)
	}
	result.ClearedTargetSelection = ClearSharedSessionBrowserTargetSelection(req.SessionRegistry, req.SessionID, req.Route)
	if result.ClearedTargetSelection {
		result.Decision = "session_target_cleared"
	}
	return result
}

func ExecuteSharedSessionBrowserClearTarget(req SharedSessionBrowserClearRequest) SharedSessionBrowserClearResult {
	return ExecuteSharedSessionBrowserClearTargetWithContext(
		SharedSessionBrowserMutationContext{
			Registry:      req.SessionRegistry,
			StateRegistry: req.StateRegistry,
		},
		req,
	)
}

// ExecuteSharedSessionBrowserClearSession applies the shared clear-session
// action contract for a scoped managed-browser route.
func ExecuteSharedSessionBrowserClearSessionWithContext(
	ctx SharedSessionBrowserMutationContext,
	req SharedSessionBrowserClearRequest,
) SharedSessionBrowserClearResult {
	if ctx.Registry != nil {
		req.SessionRegistry = ctx.Registry
	}
	if ctx.StateRegistry != nil {
		req.StateRegistry = ctx.StateRegistry
	}
	if ctx.usesWatchManagerEventSeam() {
		return ExecuteSharedSessionBrowserClearSessionEvent(
			ctx.Registry,
			ctx.RunRegistry,
			ctx.StateRegistry,
			req,
			ctx.ReconnectWindow,
		)
	}
	if blocked, ok := blockedSharedSessionBrowserClearResult(req, "clear_session_blocked_active_node_run"); ok {
		return blocked
	}
	result := SharedSessionBrowserClearResult{
		Decision: "session_route_already_clear",
		Ready:    true,
	}
	result.ClearedProfileSelection = clearSharedSessionBrowserProfileSelectionForRoute(req.StateRegistry, req.SessionID, req.Route)
	result.ClearedTargetSelection = ClearSharedSessionBrowserTargetSelection(req.SessionRegistry, req.SessionID, req.Route)
	result.ClearedSessionProfiles = clearSharedSessionBrowserProfileStateForRoute(req.StateRegistry, req.SessionID, req.Route)
	if req.SessionRegistry != nil && req.SessionID != "" {
		result.ClearedSessionTargets = req.SessionRegistry.ClearRoute(req.SessionID, req.Route)
	}
	if result.ClearedProfileSelection || result.ClearedTargetSelection || result.ClearedSessionProfiles > 0 || result.ClearedSessionTargets > 0 {
		result.Decision = "session_route_cleared"
	}
	return result
}

func ExecuteSharedSessionBrowserClearSession(req SharedSessionBrowserClearRequest) SharedSessionBrowserClearResult {
	return ExecuteSharedSessionBrowserClearSessionWithContext(
		SharedSessionBrowserMutationContext{
			Registry:      req.SessionRegistry,
			StateRegistry: req.StateRegistry,
		},
		req,
	)
}

// ExecuteSharedSessionBrowserClearActionWithContext dispatches a clear action
// through the shared mutation seam so tools callers do not need to choose
// between event-backed and raw-registry clear helpers themselves.
func ExecuteSharedSessionBrowserClearActionWithContext(
	action string,
	ctx SharedSessionBrowserMutationContext,
	req SharedSessionBrowserClearRequest,
) SharedSessionBrowserClearResult {
	switch strings.ToLower(strings.TrimSpace(action)) {
	case "clear_profile":
		return ExecuteSharedSessionBrowserClearProfileWithContext(ctx, req)
	case "clear_target":
		return ExecuteSharedSessionBrowserClearTargetWithContext(ctx, req)
	case "clear_session":
		return ExecuteSharedSessionBrowserClearSessionWithContext(ctx, req)
	default:
		return SharedSessionBrowserClearResult{}
	}
}

func blockedSharedSessionBrowserClearResult(req SharedSessionBrowserClearRequest, decision string) (SharedSessionBrowserClearResult, bool) {
	if req.Force || strings.TrimSpace(req.ActiveNodeRunID) == "" {
		return SharedSessionBrowserClearResult{}, false
	}
	profile := firstNonEmptyString(
		strings.TrimSpace(req.SelectedInfo.Profile),
		strings.TrimSpace(req.Route.Profile),
	)
	status := resolveSharedSessionBrowserExecutionStatus(
		SharedSessionBrowserExecutionRequest{
			Registry:         req.StateRegistry,
			SessionID:        req.SessionID,
			RequestedProfile: profile,
			SelectedInfo:     req.SelectedInfo,
			Force:            req.Force,
			ActiveNodeRunID:  req.ActiveNodeRunID,
			HealthInput:      req.HealthInput,
			ReconnectWindow:  req.ReconnectWindow,
		},
		profile,
		BrowserProfileStatusResult{
			Backend: strings.TrimSpace(firstNonEmptyString(
				strings.TrimSpace(req.SelectedInfo.Backend),
				strings.TrimSpace(req.Route.Backend),
			)),
			Profile: profile,
		},
	)
	return SharedSessionBrowserClearResult{
		Decision:        decision,
		Ready:           false,
		ProfileStatus:   status,
		ProfileState:    SharedSessionBrowserProfileStateFromStatus(req.SelectedInfo, status),
		HasProfileState: true,
	}, true
}

func clearSharedSessionBrowserProfileSelectionForRoute(registry SharedSessionBrowserStateRegistry, sessionID string, route BrowserSessionRoute) bool {
	if registry == nil {
		return false
	}
	sessionID = strings.TrimSpace(sessionID)
	runtimeTarget := strings.TrimSpace(route.Target)
	if sessionID == "" || runtimeTarget == "" {
		return false
	}
	if _, ok := registry.SelectedBrowserProfile(sessionID, runtimeTarget); !ok {
		return false
	}
	registry.ClearSelectedBrowserProfile(sessionID, runtimeTarget)
	return true
}

func clearSharedSessionBrowserProfileStateForRoute(registry SharedSessionBrowserStateRegistry, sessionID string, route BrowserSessionRoute) int {
	if registry == nil {
		return 0
	}
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return 0
	}
	return registry.ClearSessionBrowserProfiles(sessionID, SharedSessionBrowserProfileState{
		Backend:       strings.TrimSpace(route.Backend),
		Profile:       strings.TrimSpace(route.Profile),
		RuntimeTarget: strings.TrimSpace(route.Target),
		BrowserApp:    strings.TrimSpace(route.BrowserApp),
	})
}
