package tools

import (
	"context"
	"strings"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

type browserRuntimeCoordinateDispatchOptions struct {
	RepairScript         string
	Action               string
	CoordinationGoal     string
	Capabilities         BrowserCapabilities
	SelectedBackend      BrowserBackend
	EffectiveProfile     string
	SelectedInfo         BrowserRuntimeInfo
	SelectedRoute        *browserRuntimeRouteDescriptor
	WatchManagerProvider agentxbrowserruntime.SharedSessionBrowserObserverManager
	SessionRegistry      *BrowserSessionRegistry
	StateRegistry        agentxbrowserruntime.SharedSessionBrowserStateRegistry
	RequestedBrowserApp  string
	Force                bool
}

func browserRuntimeDispatchCoordinateAction(
	callCtx context.Context,
	payload *browserRuntimePayload,
	options browserRuntimeCoordinateDispatchOptions,
) bool {
	action := browserRuntimeCanonicalAction(options.Action)
	if action != "coordinate" {
		return false
	}
	control, ok := options.SelectedBackend.(BrowserRuntimeControlBackend)
	if !ok || !browserRuntimeActionSupported(options.Capabilities, action) {
		browserRuntimeApplyUnsupportedActionOutcome(payload, action)
		return true
	}
	goal := strings.ToLower(strings.TrimSpace(options.CoordinationGoal))
	lifecycleAction := "prepare"
	switch goal {
	case "teardown":
		lifecycleAction = "teardown"
	case "restart":
		lifecycleAction = "restart"
	}
	dispatched := agentxbrowserruntime.DispatchSharedSessionBrowserLifecycleAction(
		callCtx,
		agentxbrowserruntime.SharedSessionBrowserLifecycleActionDispatchRequest{
			Action:  lifecycleAction,
			Control: control,
			ExecutionRequest: browserRuntimeExecutionRequestFromPayload(
				callCtx,
				options.StateRegistry,
				options.EffectiveProfile,
				options.SelectedInfo,
				*payload,
				options.Force,
			),
		},
	)
	remember := dispatched.RememberProfile
	if goal == "sync" {
		remember = false
	}
	browserRuntimeApplyLifecycleActionOutcome(
		callCtx,
		options.RepairScript,
		payload,
		options.SelectedInfo,
		agentxbrowserruntime.BuildSharedSessionBrowserLifecycleActionOutcome(
			agentxbrowserruntime.SharedSessionBrowserLifecycleActionOutcomeRequest{
				Action:                           action,
				CoordinationGoal:                 options.CoordinationGoal,
				MutationContext:                  browserRuntimeMutationContext(options.WatchManagerProvider, options.SessionRegistry, options.StateRegistry),
				SessionID:                        ToolSessionIDFromContext(callCtx),
				SelectedInfo:                     options.SelectedInfo,
				Route:                            browserRuntimeSessionRouteFilter(options.SelectedRoute),
				RequestedProfile:                 payload.RequestedProfile,
				RequestedBrowserApp:              payload.RequestedBrowserApp,
				DispatchResult:                   dispatched,
				RememberProfile:                  remember,
				ApplyCoordinationDecisionOnError: true,
			},
		),
	)
	if dispatched.Err != nil {
		return true
	}
	if payload.Status == "ok" {
		browserRuntimeDispatchSessionSyncAction(
			callCtx,
			payload,
			browserRuntimeSessionSyncDispatchOptions{
				Action:               action,
				CoordinationGoal:     options.CoordinationGoal,
				Capabilities:         options.Capabilities,
				SelectedBackend:      options.SelectedBackend,
				WatchManagerProvider: options.WatchManagerProvider,
				SelectedInfo:         options.SelectedInfo,
				SelectedRoute:        payload.SelectedRoute,
				RequestedBrowserApp:  options.RequestedBrowserApp,
			},
		)
	}
	return true
}
