package tools

import (
	"context"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

type browserRuntimeClearDispatchOptions struct {
	Action               string
	Capabilities         BrowserCapabilities
	SessionRegistry      *BrowserSessionRegistry
	WatchManagerProvider agentxbrowserruntime.SharedSessionBrowserObserverManager
	StateRegistry        agentxbrowserruntime.SharedSessionBrowserStateRegistry
	SelectedInfo         BrowserRuntimeInfo
	SelectedRoute        *browserRuntimeRouteDescriptor
}

func browserRuntimeDispatchClearSessionAction(
	callCtx context.Context,
	payload *browserRuntimePayload,
	options browserRuntimeClearDispatchOptions,
) bool {
	action := browserRuntimeCanonicalAction(options.Action)
	switch action {
	case "clear_profile", "clear_session", "clear_target":
	default:
		return false
	}
	if !browserRuntimeActionSupported(options.Capabilities, action) {
		browserRuntimeApplyUnsupportedActionOutcome(payload, action)
		return true
	}
	request := browserRuntimeClearRequestFromPayload(
		callCtx,
		options.SessionRegistry,
		options.StateRegistry,
		options.SelectedInfo,
		options.SelectedRoute,
		*payload,
		payload.Force,
	)
	result := agentxbrowserruntime.ExecuteSharedSessionBrowserClearActionWithContext(
		action,
		browserRuntimeMutationContext(
			options.WatchManagerProvider,
			options.SessionRegistry,
			options.StateRegistry,
		),
		request,
	)
	browserRuntimeApplySessionActionOutcome(
		payload,
		agentxbrowserruntime.BuildSharedSessionBrowserClearActionOutcome(action, options.SelectedInfo, result),
	)
	return true
}
