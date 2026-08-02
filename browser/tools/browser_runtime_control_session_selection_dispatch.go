package tools

import (
	"context"
	"fmt"
	"strings"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

type browserRuntimeSessionSelectionDispatchOptions struct {
	Action               string
	Capabilities         BrowserCapabilities
	SelectedBackend      BrowserBackend
	WatchManagerProvider agentxbrowserruntime.SharedSessionBrowserObserverManager
	StateRegistry        agentxbrowserruntime.SharedSessionBrowserStateRegistry
	SelectedInfo         BrowserRuntimeInfo
	SelectedRoute        *browserRuntimeRouteDescriptor
	RequestedBrowserApp  string
	Params               map[string]any
	Force                bool
}

func browserRuntimeDispatchSessionSelectionAction(
	callCtx context.Context,
	payload *browserRuntimePayload,
	options browserRuntimeSessionSelectionDispatchOptions,
) bool {
	action := browserRuntimeCanonicalAction(options.Action)
	switch action {
	case "select_profile":
		if !browserRuntimeActionSupported(options.Capabilities, action) {
			browserRuntimeApplyUnsupportedActionOutcome(payload, action)
			return true
		}
		control, _ := options.SelectedBackend.(BrowserRuntimeControlBackend)
		dispatched := agentxbrowserruntime.DispatchSharedSessionBrowserSelectionAction(
			callCtx,
			agentxbrowserruntime.SharedSessionBrowserSelectionActionDispatchRequest{
				Action: action,
				MutationContext: browserRuntimeMutationContext(
					options.WatchManagerProvider,
					nil,
					options.StateRegistry,
				),
				SessionID:            ToolSessionIDFromContext(callCtx),
				SelectedInfo:         options.SelectedInfo,
				Route:                browserRuntimeSessionRouteFilter(options.SelectedRoute),
				BrowserApp:           options.RequestedBrowserApp,
				Control:              control,
				ValidateWithProfiles: control != nil && browserRuntimeActionSupported(options.Capabilities, "profiles"),
				Source:               "select_profile",
			},
		)
		browserRuntimeApplySessionActionOutcome(
			payload,
			agentxbrowserruntime.BuildSharedSessionBrowserSelectionActionOutcome(
				agentxbrowserruntime.SharedSessionBrowserSelectionActionOutcomeRequest{
					Action:         action,
					DispatchResult: dispatched,
				},
			),
		)
		return true
	case "select_target":
		if !browserRuntimeActionSupported(options.Capabilities, action) {
			browserRuntimeApplyUnsupportedActionOutcome(payload, action)
			return true
		}
		request, err := browserRuntimeSelectTargetRequest(callCtx, options.SelectedRoute, options.Params, options.Force)
		if err == nil {
			dispatched := agentxbrowserruntime.DispatchSharedSessionBrowserSelectionAction(
				callCtx,
				agentxbrowserruntime.SharedSessionBrowserSelectionActionDispatchRequest{
					Action: "select_target",
					MutationContext: browserRuntimeMutationContext(
						options.WatchManagerProvider,
						nil,
						options.StateRegistry,
					),
					SessionID:     ToolSessionIDFromContext(callCtx),
					SelectedInfo:  options.SelectedInfo,
					Route:         browserRuntimeSessionRouteFilter(options.SelectedRoute),
					Source:        "select_target",
					TargetRequest: request,
				},
			)
			browserRuntimeApplySessionActionOutcome(
				payload,
				agentxbrowserruntime.BuildSharedSessionBrowserSelectionActionOutcome(
					agentxbrowserruntime.SharedSessionBrowserSelectionActionOutcomeRequest{
						Action:             action,
						DispatchResult:     dispatched,
						ApplyTargetToRoute: true,
					},
				),
			)
			return true
		}
		browserRuntimeApplySessionActionOutcome(
			payload,
			agentxbrowserruntime.BuildSharedSessionBrowserInvalidSelectTargetActionOutcome(err),
		)
		return true
	default:
		return false
	}
}

func browserRuntimeSelectTargetRequest(
	callCtx context.Context,
	selectedRoute *browserRuntimeRouteDescriptor,
	params map[string]any,
	force bool,
) (*agentxbrowserruntime.SharedSessionBrowserSelectTargetRequest, error) {
	if selectedRoute == nil {
		return nil, nil
	}
	sessionID := ToolSessionIDFromContext(callCtx)
	if sessionID == "" {
		return nil, nil
	}
	rawTarget := strings.TrimSpace(firstString(params, "target"))
	request := &agentxbrowserruntime.SharedSessionBrowserSelectTargetRequest{
		SessionID: sessionID,
		Route:     browserRuntimeSessionRouteFilter(selectedRoute),
		Force:     force,
		Source:    "select_target",
		Actor:     "browser_runtime target selection",
	}
	switch {
	case rawTarget != "":
		target, err := parseBrowserToolTarget(rawTarget)
		if err != nil {
			return nil, fmt.Errorf("browser_runtime: %w", err)
		}
		request.Current = target.Value == "current"
		request.TargetID = strings.TrimSpace(target.TargetID)
		request.TabIndex = target.TabIndex
	case firstInt(params, "tab_index", "index") > 0:
		request.TabIndex = firstInt(params, "tab_index", "index")
	default:
		return nil, nil
	}
	return request, nil
}
