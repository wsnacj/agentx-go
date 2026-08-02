package browserruntime

import (
	"context"
	"strings"
)

// SharedSessionBrowserLifecycleActionDispatchRequest carries the runtime-owned
// execution inputs needed to dispatch managed lifecycle actions through one
// shared browserruntime helper.
type SharedSessionBrowserLifecycleActionDispatchRequest struct {
	Action               string
	ExecutionRequest     SharedSessionBrowserExecutionRequest
	Control              BrowserRuntimeControlBackend
	Manager              BrowserRuntimeProfileManagementBackend
	ProfileCreateRequest SharedSessionBrowserProfileCreateRequest
}

// SharedSessionBrowserLifecycleActionDispatchResult captures the dispatched
// lifecycle execution result together with the default remember-profile posture
// for the selected action.
type SharedSessionBrowserLifecycleActionDispatchResult struct {
	Handled         bool
	RememberProfile bool
	Result          SharedSessionBrowserExecutionResult
	Err             error
}

// DispatchSharedSessionBrowserLifecycleAction routes prepare/start/restart/
// refresh/stop/create/delete/teardown execution requests through the shared
// browserruntime owner so tools callers do not need to keep their own execute
// selection switch.
func DispatchSharedSessionBrowserLifecycleAction(
	ctx context.Context,
	req SharedSessionBrowserLifecycleActionDispatchRequest,
) SharedSessionBrowserLifecycleActionDispatchResult {
	switch strings.ToLower(strings.TrimSpace(req.Action)) {
	case "prepare":
		result, err := ExecuteSharedSessionBrowserEnsurePrepared(ctx, req.Control, req.ExecutionRequest)
		return SharedSessionBrowserLifecycleActionDispatchResult{
			Handled:         true,
			RememberProfile: true,
			Result:          result,
			Err:             err,
		}
	case "start":
		result, err := ExecuteSharedSessionBrowserStart(ctx, req.Control, req.ExecutionRequest)
		return SharedSessionBrowserLifecycleActionDispatchResult{
			Handled:         true,
			RememberProfile: true,
			Result:          result,
			Err:             err,
		}
	case "restart", "refresh":
		result, err := ExecuteSharedSessionBrowserRestart(ctx, req.Control, req.ExecutionRequest)
		return SharedSessionBrowserLifecycleActionDispatchResult{
			Handled:         true,
			RememberProfile: true,
			Result:          result,
			Err:             err,
		}
	case "stop":
		result, err := ExecuteSharedSessionBrowserStop(ctx, req.Control, req.ExecutionRequest)
		return SharedSessionBrowserLifecycleActionDispatchResult{
			Handled: true,
			Result:  result,
			Err:     err,
		}
	case "teardown":
		result, err := ExecuteSharedSessionBrowserTeardown(ctx, req.Control, req.ExecutionRequest)
		return SharedSessionBrowserLifecycleActionDispatchResult{
			Handled: true,
			Result:  result,
			Err:     err,
		}
	case "create_profile":
		result, err := ExecuteSharedSessionBrowserCreateProfile(ctx, req.Manager, req.ProfileCreateRequest)
		return SharedSessionBrowserLifecycleActionDispatchResult{
			Handled:         true,
			RememberProfile: true,
			Result:          result,
			Err:             err,
		}
	case "delete_profile":
		result, err := ExecuteSharedSessionBrowserDeleteProfile(ctx, req.Manager, req.ExecutionRequest)
		return SharedSessionBrowserLifecycleActionDispatchResult{
			Handled: true,
			Result:  result,
			Err:     err,
		}
	default:
		return SharedSessionBrowserLifecycleActionDispatchResult{}
	}
}
