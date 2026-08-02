package tools

import (
	"context"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

type browserRuntimeActionSessionResultPostProcess struct {
	Action            string
	CoordinationGoal  string
	BindingEvaluation *agentxbrowserruntime.SharedSessionBrowserBindingEvaluation
}

func browserRuntimeFinalizeActionSessionPayload(
	ctx browserRegistrationContext,
	callCtx context.Context,
	payload *browserRuntimePayload,
	options browserRuntimeActionSessionResultPostProcess,
) {
	if payload == nil {
		return
	}
	profileStatus := browserRuntimeSharedProfileStatusResultPtr(payload.ProfileStatus)
	surface := agentxbrowserruntime.BuildSharedSessionBrowserFinalizedActionSurface(
		agentxbrowserruntime.SharedSessionBrowserFinalizedActionSurfaceRequest{
			Action:            options.Action,
			CoordinationGoal:  options.CoordinationGoal,
			SessionID:         ToolSessionIDFromContext(callCtx),
			Route:             browserRuntimeSessionRouteFilter(payload.SelectedRoute),
			Routes:            browserRuntimeSharedSessionRouteSnapshots(payload.SessionRoutes),
			Registry:          ctx.sessionRegistry,
			RunRegistry:       ctx.sessionRunRegistry,
			StateRegistry:     ctx.sessionStateRegistry,
			BindingEvaluation: options.BindingEvaluation,
			ProfileStatus:     profileStatus,
			HealthSummary:     payload.finalizedSessionHealthSummary,
			ReconnectWindow:   browserRuntimeReconnectWatchdogWindow,
			CoordinationSummary: agentxbrowserruntime.SharedSessionBrowserCoordinationSummaryInput{
				PrepareDecision:  payload.PrepareDecision,
				RestartDecision:  payload.RestartDecision,
				SyncSessionReady: payload.SyncSessionReady,
			},
		},
	)
	payload.finalizedSessionHealthSummary = nil
	browserRuntimeApplyFinalizedSessionActionSurface(
		callCtx,
		payload,
		surface,
	)
	if !surface.UseWorkbenchSurface {
		browserRuntimeSyncSharedGuidanceProjection(payload, false)
		browserRuntimeSyncTopLevelSurfaceSummary(payload)
	}
}

func browserRuntimeSharedProfileStatusResultPtr(
	profile *browserRuntimeProfileState,
) *agentxbrowserruntime.BrowserProfileStatusResult {
	if profile == nil {
		return nil
	}
	status := agentxbrowserruntime.SharedSessionBrowserProfileStatusResultFromState(
		browserRuntimeSharedSessionProfileState(*profile),
		BrowserRuntimeInfo{
			Backend: profile.Backend,
			Profile: profile.Profile,
			Target:  profile.RuntimeTarget,
		},
		profile.Profile,
	)
	return &status
}
