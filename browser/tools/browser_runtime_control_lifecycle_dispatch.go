package tools

import (
	"context"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

type browserRuntimeLifecycleDispatchOptions struct {
	RepairScript         string
	Action               string
	Capabilities         BrowserCapabilities
	SelectedBackend      BrowserBackend
	EffectiveProfile     string
	SelectedInfo         BrowserRuntimeInfo
	SelectedRoute        *browserRuntimeRouteDescriptor
	WatchManagerProvider agentxbrowserruntime.SharedSessionBrowserObserverManager
	SessionRegistry      *BrowserSessionRegistry
	StateRegistry        agentxbrowserruntime.SharedSessionBrowserStateRegistry
}

func browserRuntimeDispatchLifecycleAction(
	callCtx context.Context,
	payload *browserRuntimePayload,
	options browserRuntimeLifecycleDispatchOptions,
) bool {
	action := browserRuntimeCanonicalAction(options.Action)
	switch action {
	case "prepare", "start", "restart", "refresh", "stop":
		control, ok := options.SelectedBackend.(BrowserRuntimeControlBackend)
		if !ok || !browserRuntimeActionSupported(options.Capabilities, action) {
			browserRuntimeApplyUnsupportedActionOutcome(payload, action)
			return true
		}
		dispatched := agentxbrowserruntime.DispatchSharedSessionBrowserLifecycleAction(
			callCtx,
			agentxbrowserruntime.SharedSessionBrowserLifecycleActionDispatchRequest{
				Action:  action,
				Control: control,
				ExecutionRequest: browserRuntimeExecutionRequestFromPayload(
					callCtx,
					options.StateRegistry,
					options.EffectiveProfile,
					options.SelectedInfo,
					*payload,
					payload.Force,
				),
			},
		)
		outcome := agentxbrowserruntime.BuildSharedSessionBrowserLifecycleActionOutcome(
			agentxbrowserruntime.SharedSessionBrowserLifecycleActionOutcomeRequest{
				Action:              action,
				MutationContext:     browserRuntimeMutationContext(options.WatchManagerProvider, options.SessionRegistry, options.StateRegistry),
				SessionID:           ToolSessionIDFromContext(callCtx),
				SelectedInfo:        options.SelectedInfo,
				Route:               browserRuntimeSessionRouteFilter(options.SelectedRoute),
				RequestedProfile:    payload.RequestedProfile,
				RequestedBrowserApp: payload.RequestedBrowserApp,
				DispatchResult:      dispatched,
				RememberProfile:     dispatched.RememberProfile,
			},
		)
		browserRuntimeApplyLifecycleActionOutcome(
			callCtx,
			options.RepairScript,
			payload,
			options.SelectedInfo,
			outcome,
		)
		browserRuntimeRefreshLifecycleBindingEvaluation(callCtx, payload, options.StateRegistry, options.SelectedInfo, outcome.Result.ProfileStatus.SessionHealth)
		return true
	case "create_profile", "delete_profile":
		manager, ok := options.SelectedBackend.(BrowserRuntimeProfileManagementBackend)
		if !ok || !browserRuntimeActionSupported(options.Capabilities, action) {
			browserRuntimeApplyUnsupportedActionOutcome(payload, action)
			return true
		}
		dispatched := agentxbrowserruntime.DispatchSharedSessionBrowserLifecycleAction(
			callCtx,
			agentxbrowserruntime.SharedSessionBrowserLifecycleActionDispatchRequest{
				Action:  action,
				Manager: manager,
				ExecutionRequest: browserRuntimeExecutionRequestFromPayload(
					callCtx,
					options.StateRegistry,
					payload.RequestedProfile,
					options.SelectedInfo,
					*payload,
					payload.Force,
				),
				ProfileCreateRequest: agentxbrowserruntime.SharedSessionBrowserProfileCreateRequest{
					RequestedProfile: payload.RequestedProfile,
					SelectedInfo:     options.SelectedInfo,
					BrowserApp:       payload.RequestedBrowserApp,
					Color:            payload.RequestedColor,
					CopyFrom:         payload.RequestedCopyFrom,
				},
			},
		)
		outcome := agentxbrowserruntime.BuildSharedSessionBrowserLifecycleActionOutcome(
			agentxbrowserruntime.SharedSessionBrowserLifecycleActionOutcomeRequest{
				Action:              action,
				MutationContext:     browserRuntimeMutationContext(options.WatchManagerProvider, options.SessionRegistry, options.StateRegistry),
				SessionID:           ToolSessionIDFromContext(callCtx),
				SelectedInfo:        options.SelectedInfo,
				Route:               browserRuntimeSessionRouteFilter(options.SelectedRoute),
				RequestedProfile:    payload.RequestedProfile,
				RequestedBrowserApp: payload.RequestedBrowserApp,
				DispatchResult:      dispatched,
				RememberProfile:     dispatched.RememberProfile,
			},
		)
		browserRuntimeApplyLifecycleActionOutcome(
			callCtx,
			options.RepairScript,
			payload,
			options.SelectedInfo,
			outcome,
		)
		browserRuntimeRefreshLifecycleBindingEvaluation(callCtx, payload, options.StateRegistry, options.SelectedInfo, outcome.Result.ProfileStatus.SessionHealth)
		return true
	default:
		return false
	}
}
