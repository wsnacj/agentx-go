package tools

import (
	"context"
	"strings"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

type browserRuntimeSessionSyncDispatchOptions struct {
	Action               string
	CoordinationGoal     string
	Capabilities         BrowserCapabilities
	SelectedBackend      BrowserBackend
	WatchManagerProvider agentxbrowserruntime.SharedSessionBrowserObserverManager
	SelectedInfo         BrowserRuntimeInfo
	SelectedRoute        *browserRuntimeRouteDescriptor
	RequestedBrowserApp  string
}

func browserRuntimeDispatchSessionSyncAction(
	callCtx context.Context,
	payload *browserRuntimePayload,
	options browserRuntimeSessionSyncDispatchOptions,
) bool {
	action := browserRuntimeCanonicalAction(options.Action)
	switch {
	case action == "sync_session":
		if !browserRuntimeActionSupported(options.Capabilities, "sync_session") {
			browserRuntimeApplyUnsupportedActionOutcome(payload, action)
			return true
		}
		control, _ := options.SelectedBackend.(BrowserRuntimeControlBackend)
		dispatched := agentxbrowserruntime.DispatchSharedSessionBrowserSyncAction(
			callCtx,
			agentxbrowserruntime.SharedSessionBrowserSyncActionDispatchRequest{
				Action:               action,
				MutationContext:      browserRuntimeMutationContext(options.WatchManagerProvider, nil, nil),
				SessionID:            ToolSessionIDFromContext(callCtx),
				SelectedInfo:         options.SelectedInfo,
				Route:                browserRuntimeSessionRouteFilter(options.SelectedRoute),
				BrowserApp:           options.RequestedBrowserApp,
				Control:              control,
				ValidateWithProfiles: control != nil && browserRuntimeActionSupported(options.Capabilities, "profiles"),
			},
		)
		browserRuntimeApplySessionActionOutcome(
			payload,
			agentxbrowserruntime.BuildSharedSessionBrowserSyncActionOutcome(
				action,
				"",
				dispatched,
				false,
				false,
			),
		)
		return true
	case action == "coordinate" && strings.EqualFold(strings.TrimSpace(options.CoordinationGoal), "sync"):
		control, ok := options.SelectedBackend.(BrowserRuntimeControlBackend)
		if !ok || !browserRuntimeActionSupported(options.Capabilities, "coordinate") {
			browserRuntimeApplyUnsupportedActionOutcome(payload, action)
			return true
		}
		dispatched := agentxbrowserruntime.DispatchSharedSessionBrowserSyncAction(
			callCtx,
			agentxbrowserruntime.SharedSessionBrowserSyncActionDispatchRequest{
				Action:               action,
				CoordinationGoal:     options.CoordinationGoal,
				MutationContext:      browserRuntimeMutationContext(options.WatchManagerProvider, nil, nil),
				SessionID:            ToolSessionIDFromContext(callCtx),
				SelectedInfo:         options.SelectedInfo,
				Route:                browserRuntimeSessionRouteFilter(options.SelectedRoute),
				BrowserApp:           options.RequestedBrowserApp,
				Control:              control,
				ValidateWithProfiles: browserRuntimeActionSupported(options.Capabilities, "profiles"),
			},
		)
		browserRuntimeApplySessionActionOutcome(
			payload,
			agentxbrowserruntime.BuildSharedSessionBrowserSyncActionOutcome(
				action,
				options.CoordinationGoal,
				dispatched,
				true,
				true,
			),
		)
		return true
	default:
		return false
	}
}
