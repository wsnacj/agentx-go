package tools

import (
	"context"
	"strings"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

func browserRuntimeTopLevelBindingProjection(
	ctx context.Context,
	sessionRegistry *BrowserSessionRegistry,
	sessionRunRegistry BrowserSessionRunRegistry,
	sessionStateRegistry agentxbrowserruntime.SharedSessionBrowserStateRegistry,
	selectedRoute *browserRuntimeRouteDescriptor,
	routes []browserRuntimeSessionRoute,
	evaluation *agentxbrowserruntime.SharedSessionBrowserBindingEvaluation,
	current *browserRuntimeProfileState,
) agentxbrowserruntime.SharedSessionBrowserTopLevelBindingProjection {
	if selectedRoute == nil {
		if evaluation != nil {
			var sharedCurrent *agentxbrowserruntime.SharedSessionBrowserProfileState
			if current != nil {
				state := browserRuntimeSharedSessionProfileState(*current)
				sharedCurrent = &state
			}
			return agentxbrowserruntime.ProjectSharedSessionBrowserTopLevelBindingFromEvaluation(
				ToolSessionIDFromContext(ctx),
				browserRuntimeSharedSessionRouteSnapshots(routes),
				sessionRegistry,
				sessionRunRegistry,
				sessionStateRegistry,
				*evaluation,
				sharedCurrent,
				browserRuntimeReconnectWatchdogWindow,
			)
		}
		return agentxbrowserruntime.SharedSessionBrowserTopLevelBindingProjection{}
	}
	var sharedCurrent *agentxbrowserruntime.SharedSessionBrowserProfileState
	if current != nil {
		state := browserRuntimeSharedSessionProfileState(*current)
		sharedCurrent = &state
	}
	return agentxbrowserruntime.ProjectSharedSessionBrowserTopLevelBinding(
		ToolSessionIDFromContext(ctx),
		agentxbrowserruntime.BrowserRuntimeInfo{
			Backend: strings.TrimSpace(selectedRoute.Backend),
			Profile: strings.TrimSpace(selectedRoute.Profile),
			Target:  strings.TrimSpace(selectedRoute.RuntimeTarget),
		},
		browserRuntimeSessionRouteFilter(selectedRoute),
		browserRuntimeSharedSessionRouteSnapshots(routes),
		sessionRegistry,
		sessionRunRegistry,
		sessionStateRegistry,
		evaluation,
		sharedCurrent,
		browserRuntimeReconnectWatchdogWindow,
	)
}
