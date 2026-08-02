package tools

import (
	"context"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

func browserRuntimeApplyFinalizedSessionActionSurface(
	callCtx context.Context,
	payload *browserRuntimePayload,
	surface agentxbrowserruntime.SharedSessionBrowserFinalizedActionSurface,
) {
	if payload == nil {
		return
	}
	if surface.UseWorkbenchSurface {
		if surface.WorkbenchSurface != nil {
			browserRuntimeApplySharedWorkbenchSessionSurface(
				callCtx,
				payload,
				*surface.WorkbenchSurface,
				nil,
				surface.SyncCoordinationSurface,
			)
		}
	} else if surface.BindingProjection != nil {
		browserRuntimeApplyTopLevelBindingProjection(callCtx, payload, *surface.BindingProjection)
	}
	browserRuntimeApplyCoordinationSummaryProjection(
		payload,
		browserRuntimeCoordinationSummaryProjection{Summary: surface.CoordinationSummary},
	)
}

func browserRuntimeTopLevelSessionProjectionFromBindingPayload(
	payload browserRuntimePayload,
) *browserRuntimeTopLevelSessionProjection {
	binding := payload.SessionBinding
	if binding == nil {
		return nil
	}
	sharedEvaluation := browserRuntimeSharedBindingEvaluationForPayload(payload, payload.SessionRoutes)
	sharedEvaluation = agentxbrowserruntime.ProjectSharedSessionBrowserBindingEvaluationWithSelectionProjection(
		sharedEvaluation,
		browserRuntimeProjectedSelectionProjectionFromBindingPayload(payload),
	)
	shared := agentxbrowserruntime.ProjectSharedSessionBrowserTopLevelSessionFromBindingEvaluation(
		sharedEvaluation,
	)
	projection := browserRuntimeTopLevelSessionProjectionFromShared(&shared)
	if projection == nil || (len(projection.Routes) == 0 && projection.TargetCount == 0 && len(projection.Runs) == 0 && len(projection.Profiles) == 0 && projection.Handoff == nil) {
		return nil
	}
	return projection
}

func browserRuntimeSharedSelectionProjectionFromPayload(
	payload browserRuntimePayload,
) *agentxbrowserruntime.SharedSessionBrowserSelectionProjection {
	profileSelection := browserRuntimeSharedProfileSelection(payload.SessionProfileSelection)
	targetSelection := browserRuntimeSharedTargetSelection(payload.SessionTargetSelection)
	if profileSelection == nil && targetSelection == nil {
		return nil
	}
	return &agentxbrowserruntime.SharedSessionBrowserSelectionProjection{
		ProfileSelection: profileSelection,
		TargetSelection:  targetSelection,
	}
}

func browserRuntimeSharedBindingEvaluationForPayload(
	payload browserRuntimePayload,
	routes []browserRuntimeSessionRoute,
) agentxbrowserruntime.SharedSessionBrowserBindingEvaluation {
	var evaluation agentxbrowserruntime.SharedSessionBrowserBindingEvaluation
	if payload.SessionBinding != nil {
		evaluation = browserRuntimeSharedBindingEvaluation(*payload.SessionBinding, routes)
	} else {
		evaluation.Routes = browserRuntimeSharedSessionRouteSnapshots(routes)
	}
	extraProfiles := make([]agentxbrowserruntime.SharedSessionBrowserProfileState, 0, len(payload.Profiles)+1)
	if payload.ProfileStatus != nil {
		extraProfiles = append(extraProfiles, browserRuntimeSharedSessionProfileState(*payload.ProfileStatus))
	}
	extraProfiles = append(extraProfiles, browserRuntimeSharedSessionProfileStates(payload.Profiles)...)
	if len(extraProfiles) > 0 {
		evaluation.Snapshot.Profiles = append(extraProfiles, evaluation.Snapshot.Profiles...)
	}
	return evaluation
}

func browserRuntimeProjectedSelectionProjectionFromBindingPayload(
	payload browserRuntimePayload,
) *agentxbrowserruntime.SharedSessionBrowserSelectionProjection {
	projection := browserRuntimeSharedSelectionProjectionFromPayload(payload)
	if payload.SessionBinding == nil && payload.ProfileStatus == nil && len(payload.Profiles) == 0 {
		return projection
	}
	return agentxbrowserruntime.ProjectSharedSessionBrowserSelectionProjectionFromBindingEvaluation(
		browserRuntimeSharedBindingEvaluationForPayload(payload, payload.SessionRoutes),
		projection,
	)
}

func browserRuntimeBindingWithPayloadSelections(
	binding browserRuntimeSessionBinding,
	payload browserRuntimePayload,
) browserRuntimeSessionBinding {
	projection := browserRuntimeProjectedSelectionProjectionFromBindingPayload(payload)
	if projection == nil {
		return binding
	}
	browserRuntimeApplySharedBindingEvaluation(
		&binding,
		agentxbrowserruntime.ProjectSharedSessionBrowserBindingEvaluationWithSelectionProjection(
			browserRuntimeSharedBindingEvaluation(binding, nil),
			projection,
		),
	)
	return binding
}

func browserRuntimeTopLevelSessionProjectionFromShared(
	projection *agentxbrowserruntime.SharedSessionBrowserTopLevelSessionProjection,
) *browserRuntimeTopLevelSessionProjection {
	if projection == nil {
		return nil
	}
	return &browserRuntimeTopLevelSessionProjection{
		Routes:      browserRuntimeSessionRoutesFromShared(projection.Routes),
		TargetCount: projection.TargetCount,
		Runs:        browserRuntimeSessionRunsFromShared(projection.Runs),
		Profiles:    browserRuntimeProfileStatesFromProjected(projection.Profiles),
		Handoff:     agentxbrowserruntime.CloneSharedSessionBrowserSessionHandoffSummary(projection.Handoff),
	}
}
