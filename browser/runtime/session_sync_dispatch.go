package browserruntime

import (
	"context"
	"strings"
)

// SharedSessionBrowserSyncActionDispatchRequest carries the runtime-owned
// inputs needed to dispatch sync_session and coordinate(goal=sync) through one
// shared browserruntime helper.
type SharedSessionBrowserSyncActionDispatchRequest struct {
	Action               string
	CoordinationGoal     string
	MutationContext      SharedSessionBrowserMutationContext
	SessionID            string
	SelectedInfo         BrowserRuntimeInfo
	Route                BrowserSessionRoute
	BrowserApp           string
	Control              BrowserRuntimeControlBackend
	ValidateWithProfiles bool
}

// SharedSessionBrowserSyncActionDispatchResult captures the shared sync action
// outcome together with whether the helper claimed the requested action.
type SharedSessionBrowserSyncActionDispatchResult struct {
	Handled bool
	Result  SharedSessionBrowserSyncSelectionResult
	Err     error
}

// DispatchSharedSessionBrowserSyncAction routes sync_session and
// coordinate(goal=sync) through the shared browserruntime owner so tools
// callers no longer keep their own route-sync execute switch.
func DispatchSharedSessionBrowserSyncAction(
	ctx context.Context,
	req SharedSessionBrowserSyncActionDispatchRequest,
) SharedSessionBrowserSyncActionDispatchResult {
	action := strings.ToLower(strings.TrimSpace(req.Action))
	switch {
	case action == "sync_session":
		result, err := SyncSharedSessionBrowserRouteSelectionWithContext(
			req.MutationContext,
			ctx,
			req.SessionID,
			req.SelectedInfo,
			req.Route,
			req.BrowserApp,
			req.Control,
			req.ValidateWithProfiles,
			"sync_session",
		)
		return SharedSessionBrowserSyncActionDispatchResult{
			Handled: true,
			Result:  result,
			Err:     err,
		}
	case action == "coordinate" && strings.EqualFold(strings.TrimSpace(req.CoordinationGoal), "sync"):
		result, err := CoordinateSharedSessionBrowserRouteSyncWithContext(
			req.MutationContext,
			ctx,
			req.SessionID,
			req.SelectedInfo,
			req.Route,
			req.BrowserApp,
			req.Control,
			req.ValidateWithProfiles,
		)
		return SharedSessionBrowserSyncActionDispatchResult{
			Handled: true,
			Result:  result,
			Err:     err,
		}
	default:
		return SharedSessionBrowserSyncActionDispatchResult{}
	}
}
