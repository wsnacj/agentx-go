package tools

import (
	"context"
	"strings"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

type browserRuntimeWorkbenchSessionSurface struct {
	ProfileInventory        *browserRuntimeTopLevelProfileInventoryProjection
	SessionProjection       *browserRuntimeTopLevelSessionProjection
	BindingProjection       *agentxbrowserruntime.SharedSessionBrowserTopLevelBindingProjection
	ApplyDirectBinding      bool
	SessionProfileSelection *browserRuntimeSessionProfileSelection
	SessionTargetSelection  *browserRuntimeSessionTargetSelection
	SessionBinding          *browserRuntimeSessionBinding
	ExtraSections           []string
	SyncCoordinationSurface bool
}

func browserRuntimeSharedWorkbenchSessionSurfaceRequest(
	payload browserRuntimePayload,
	surface browserRuntimeWorkbenchSessionSurface,
) agentxbrowserruntime.SharedSessionBrowserWorkbenchSessionSurfaceRequest {
	selectedInfo := BrowserRuntimeInfo{}
	if surface.SessionProjection != nil {
		selectedInfo = normalizeBrowserRuntimeInfo(surface.SessionProjection.ConfiguredInfo)
	}
	if selectedInfo == (BrowserRuntimeInfo{}) {
		selectedInfo = normalizeBrowserRuntimeInfo(browserRuntimeConfiguredInfoForBindingPayload(payload))
	}
	configuredInfo := selectedInfo
	applyConfiguredProfiles := true
	missingSessionIDNoteEnabled := true
	if surface.SessionProjection != nil {
		if explicitInfo := normalizeBrowserRuntimeInfo(surface.SessionProjection.ConfiguredInfo); explicitInfo != (BrowserRuntimeInfo{}) {
			configuredInfo = explicitInfo
		}
		applyConfiguredProfiles = surface.SessionProjection.ApplyConfiguredProfiles
		missingSessionIDNoteEnabled = strings.TrimSpace(surface.SessionProjection.MissingSessionIDNote) != ""
	}
	req := agentxbrowserruntime.SharedSessionBrowserWorkbenchSessionSurfaceRequest{
		SelectedInfo:              selectedInfo,
		RequestedDefaultProfile:   payload.DefaultProfile,
		NeedProfileStatus:         payload.ProfileStatus == nil,
		NeedProfileInventory:      len(payload.Profiles) == 0,
		ConfiguredInfo:            configuredInfo,
		ApplyConfiguredProfiles:   applyConfiguredProfiles,
		ApplyMissingSessionIDNote: missingSessionIDNoteEnabled,
		BindingProjection:         surface.BindingProjection,
		SessionProjection:         browserRuntimeSharedTopLevelSessionProjectionPtr(surface.SessionProjection),
	}
	if !surface.ApplyDirectBinding || (surface.SessionBinding == nil && surface.SessionProfileSelection == nil && surface.SessionTargetSelection == nil) {
		if surface.SessionProjection == nil && payload.SessionBinding != nil {
			evaluation := browserRuntimeSharedBindingEvaluationForPayload(payload, payload.SessionRoutes)
			if projection := browserRuntimeProjectedSelectionProjectionFromBindingPayload(payload); projection != nil {
				evaluation = agentxbrowserruntime.ProjectSharedSessionBrowserBindingEvaluationWithSelectionProjection(
					evaluation,
					projection,
				)
			}
			req.BindingEvaluation = &evaluation
		}
		return req
	}
	directPayload := payload
	if surface.SessionProfileSelection != nil {
		directPayload.SessionProfileSelection = surface.SessionProfileSelection
	}
	if surface.SessionTargetSelection != nil {
		directPayload.SessionTargetSelection = surface.SessionTargetSelection
	}
	if surface.SessionBinding != nil {
		directPayload.SessionBinding = surface.SessionBinding
	}
	evaluation := browserRuntimeSharedBindingEvaluationForPayload(directPayload, directPayload.SessionRoutes)
	if projection := browserRuntimeProjectedSelectionProjectionFromBindingPayload(directPayload); projection != nil {
		evaluation = agentxbrowserruntime.ProjectSharedSessionBrowserBindingEvaluationWithSelectionProjection(
			evaluation,
			projection,
		)
	}
	req.BindingEvaluation = &evaluation
	req.ApplyBindingEvaluation = true
	return req
}

func browserRuntimeApplyWorkbenchSessionSurface(
	callCtx context.Context,
	payload *browserRuntimePayload,
	surface browserRuntimeWorkbenchSessionSurface,
) {
	if payload == nil {
		return
	}
	if surface.ProfileInventory != nil {
		browserRuntimeApplyTopLevelProfileInventory(payload, *surface.ProfileInventory)
	}
	shared := agentxbrowserruntime.BuildSharedSessionBrowserWorkbenchSessionSurfaceProjection(
		browserRuntimeSharedWorkbenchSessionSurfaceRequest(*payload, surface),
	)
	browserRuntimeApplySharedWorkbenchSessionSurface(
		callCtx,
		payload,
		shared,
		surface.ExtraSections,
		surface.SyncCoordinationSurface,
	)
	if shared.BindingProjection != nil && payload.SessionBinding == nil && surface.SessionBinding != nil {
		normalized := browserRuntimeBindingWithPayloadSelections(*surface.SessionBinding, *payload)
		payload.SessionBinding = &normalized
	}
}

func browserRuntimeApplySharedWorkbenchSessionSurface(
	callCtx context.Context,
	payload *browserRuntimePayload,
	shared agentxbrowserruntime.SharedSessionBrowserWorkbenchSessionSurfaceProjection,
	extraSections []string,
	syncCoordinationSurface bool,
) {
	if payload == nil {
		return
	}
	if shared.BindingProjection != nil {
		projection := shared.BindingProjection
		browserRuntimeApplyTopLevelBindingProjection(callCtx, payload, *projection)
	}
	if projection := browserRuntimeTopLevelSessionProjectionApplicationFromShared(shared.SessionProjection); projection != nil {
		browserRuntimeApplyTopLevelSessionProjection(payload, *projection)
	}
	if projection := browserRuntimeTopLevelSessionProjectionApplicationFromShared(shared.FallbackSessionProjection); projection != nil {
		browserRuntimeApplyTopLevelSessionProjection(payload, *projection)
	}
	configuredInfo := normalizeBrowserRuntimeInfo(browserRuntimeConfiguredInfoForBindingPayload(*payload))
	switch {
	case shared.SessionProjection != nil && normalizeBrowserRuntimeInfo(shared.SessionProjection.ConfiguredInfo) != (BrowserRuntimeInfo{}):
		configuredInfo = normalizeBrowserRuntimeInfo(shared.SessionProjection.ConfiguredInfo)
	case shared.FallbackSessionProjection != nil && normalizeBrowserRuntimeInfo(shared.FallbackSessionProjection.ConfiguredInfo) != (BrowserRuntimeInfo{}):
		configuredInfo = normalizeBrowserRuntimeInfo(shared.FallbackSessionProjection.ConfiguredInfo)
	}
	if shared.ProfileInventory != nil {
		browserRuntimeApplyTopLevelProfileInventory(
			payload,
			*browserRuntimeTopLevelProfileInventoryProjectionFromShared(
				configuredInfo,
				shared.ProfileInventory,
			),
		)
	}
	browserRuntimeBackfillSelectedRouteFromBindingPayload(payload)
	browserRuntimeBackfillSelectionsFromBindingPayload(payload)
	browserRuntimeBackfillBindingIdentityFromPayload(payload)
	browserRuntimeRefreshProfileInventoryFromBindingPayload(payload)
	browserRuntimeRefreshDefaultProfileFromBindingPayload(payload)
	browserRuntimeRefreshConfiguredProfilesFromBindingPayload(payload)
	browserRuntimeSyncWorkbenchProjection(payload, browserRuntimeWorkbenchProjectionSync{
		ExtraSections:           extraSections,
		ClearActionPlan:         syncCoordinationSurface,
		SyncCoordinationSurface: syncCoordinationSurface,
	})
}

func browserRuntimeTopLevelSessionProjectionApplicationFromShared(
	application *agentxbrowserruntime.SharedSessionBrowserWorkbenchSessionProjectionApplication,
) *browserRuntimeTopLevelSessionProjection {
	if application == nil || application.Projection == nil {
		return nil
	}
	projection := browserRuntimeTopLevelSessionProjectionFromShared(application.Projection)
	if projection == nil {
		return nil
	}
	projection.ConfiguredInfo = normalizeBrowserRuntimeInfo(application.ConfiguredInfo)
	projection.ApplyConfiguredProfiles = application.ApplyConfiguredProfiles
	projection.MissingSessionIDNote = strings.TrimSpace(application.MissingSessionIDNote)
	return projection
}
