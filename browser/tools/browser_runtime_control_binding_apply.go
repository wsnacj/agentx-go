package tools

import (
	"context"
	"strings"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

func browserRuntimeSharedBindingRouteProjectionForPayload(
	payload browserRuntimePayload,
) agentxbrowserruntime.SharedSessionBrowserBindingRouteProjection {
	return agentxbrowserruntime.ProjectSharedSessionBrowserBindingRouteProjection(
		browserRuntimeSharedBindingEvaluationForPayload(payload, payload.SessionRoutes),
		browserRuntimeSharedSelectionProjectionFromPayload(payload),
		payload.DefaultProfile,
	)
}

func browserRuntimeHydrateSelectionProjectionFromBindingPayload(payload *browserRuntimePayload) {
	if payload == nil {
		return
	}
	profileSelection, targetSelection := browserRuntimeProjectedSelectionPtrsFromBindingPayload(*payload)
	if payload.SessionProfileSelection != nil && profileSelection != nil {
		payload.SessionProfileSelection = profileSelection
	}
	if payload.SessionTargetSelection != nil && targetSelection != nil {
		payload.SessionTargetSelection = targetSelection
	}
}

func browserRuntimeProjectedSelectionPtrsFromBindingPayload(
	payload browserRuntimePayload,
) (*browserRuntimeSessionProfileSelection, *browserRuntimeSessionTargetSelection) {
	projection := browserRuntimeProjectedSelectionProjectionFromBindingPayload(payload)
	if projection == nil {
		return nil, nil
	}
	return browserRuntimeSelectionPtrValue(projection.ProfileSelection),
		browserRuntimeSessionTargetSelectionPtrFromShared(projection.TargetSelection)
}

func browserRuntimeSelectedRouteInfoFromBindingPayload(
	payload browserRuntimePayload,
) BrowserRuntimeInfo {
	return normalizeBrowserRuntimeInfo(
		browserRuntimeSharedBindingRouteProjectionForPayload(payload).SelectedRouteInfo,
	)
}

func browserRuntimeBackfillSelectedRouteFromBindingPayload(payload *browserRuntimePayload) {
	if payload == nil || payload.SelectedRoute != nil {
		return
	}
	if payload.SessionBinding == nil && payload.SessionProfileSelection == nil && payload.SessionTargetSelection == nil {
		return
	}
	info := browserRuntimeSelectedRouteInfoFromBindingPayload(*payload)
	switch BrowserSubstratePosture(info.Backend, info.Target) {
	case BrowserSubstrateNodeRuntime, BrowserSubstrateSandboxRuntime:
	default:
		return
	}
	payload.SelectedRoute = browserRuntimeRouteDescriptorPtr(info)
}

func browserRuntimeBackfillSelectionsFromBindingPayload(payload *browserRuntimePayload) {
	if payload == nil || payload.SessionBinding == nil {
		return
	}
	if payload.SelectedRoute == nil {
		info := browserRuntimeSelectedRouteInfoFromBindingPayload(*payload)
		switch BrowserSubstratePosture(info.Backend, info.Target) {
		case BrowserSubstrateNodeRuntime, BrowserSubstrateSandboxRuntime:
		default:
			return
		}
	}
	profileSelection, targetSelection := browserRuntimeProjectedSelectionPtrsFromBindingPayload(*payload)
	if payload.SessionProfileSelection == nil {
		payload.SessionProfileSelection = profileSelection
	}
	if payload.SessionTargetSelection == nil {
		payload.SessionTargetSelection = targetSelection
	}
	browserRuntimeHydrateSelectionProjectionFromBindingPayload(payload)
}

func browserRuntimeBackfillBindingIdentityFromPayload(payload *browserRuntimePayload) {
	if payload == nil || payload.SessionBinding == nil {
		return
	}
	normalized := browserRuntimeBindingWithPayloadSelections(*payload.SessionBinding, *payload)
	payload.SessionBinding = &normalized
}

func browserRuntimeConfiguredInfoForBindingPayload(payload browserRuntimePayload) BrowserRuntimeInfo {
	if info := browserRuntimeInfoFromRouteDescriptor(payload.SelectedRoute); info != (BrowserRuntimeInfo{}) {
		return info
	}
	return normalizeBrowserRuntimeInfo(
		browserRuntimeSharedBindingRouteProjectionForPayload(payload).ConfiguredInfo,
	)
}

func browserRuntimeDefaultProfileForBindingPayload(payload browserRuntimePayload) string {
	return strings.TrimSpace(firstNonEmpty(
		payload.DefaultProfile,
		func() string {
			if payload.SelectedRoute != nil {
				return payload.SelectedRoute.Profile
			}
			return ""
		}(),
		browserRuntimeSharedBindingRouteProjectionForPayload(payload).DefaultProfile,
	))
}

func browserRuntimeRefreshDefaultProfileFromBindingPayload(payload *browserRuntimePayload) {
	if payload == nil || strings.TrimSpace(payload.DefaultProfile) != "" {
		return
	}
	if profile := browserRuntimeDefaultProfileForBindingPayload(*payload); profile != "" {
		payload.DefaultProfile = profile
	}
}

func browserRuntimeTopLevelProfileInventoryProjectionFromBindingPayload(
	payload browserRuntimePayload,
) *browserRuntimeTopLevelProfileInventoryProjection {
	if payload.SessionBinding == nil {
		return nil
	}
	configuredInfo := browserRuntimeConfiguredInfoForBindingPayload(payload)
	inventoryPayload := payload
	inventoryPayload.ProfileStatus = nil
	sharedEvaluation := browserRuntimeSharedBindingEvaluationForPayload(inventoryPayload, inventoryPayload.SessionRoutes)
	sharedEvaluation = agentxbrowserruntime.ProjectSharedSessionBrowserBindingEvaluationWithSelectionProjection(
		sharedEvaluation,
		browserRuntimeProjectedSelectionProjectionFromBindingPayload(payload),
	)
	shared := agentxbrowserruntime.ProjectSharedSessionBrowserTopLevelProfileInventoryFromBindingEvaluation(
		agentxbrowserruntime.SharedSessionBrowserBindingProfileInventoryProjectionRequest{
			Evaluation:              sharedEvaluation,
			SelectedInfo:            configuredInfo,
			RequestedDefaultProfile: payload.DefaultProfile,
			NeedProfileStatus:       payload.ProfileStatus == nil,
			NeedProfileInventory:    len(payload.Profiles) == 0,
		},
	)
	return browserRuntimeTopLevelProfileInventoryProjectionFromShared(configuredInfo, shared)
}

func browserRuntimeRefreshProfileInventoryFromBindingPayload(payload *browserRuntimePayload) {
	if payload == nil || payload.SessionBinding == nil {
		return
	}
	if projection := browserRuntimeTopLevelProfileInventoryProjectionFromBindingPayload(*payload); projection != nil {
		browserRuntimeApplyTopLevelProfileInventory(payload, *projection)
	}
}

func browserRuntimeConfiguredProfilesProjectionForBindingPayload(
	payload browserRuntimePayload,
	selectedInfo BrowserRuntimeInfo,
) browserRuntimeConfiguredProfilesProjection {
	return browserRuntimeConfiguredProfilesProjection{
		Profiles: agentxbrowserruntime.ProjectSharedSessionBrowserConfiguredProfiles(
			browserRuntimeSharedConfiguredProfilesProjectionRequest(payload, selectedInfo, true),
		),
	}
}

func browserRuntimeRefreshConfiguredProfilesFromBindingPayload(payload *browserRuntimePayload) {
	if payload == nil {
		return
	}
	if configuredInfo := browserRuntimeConfiguredInfoForBindingPayload(*payload); configuredInfo != (BrowserRuntimeInfo{}) {
		browserRuntimeApplyConfiguredProfilesProjection(
			payload,
			browserRuntimeConfiguredProfilesProjectionForBindingPayload(*payload, configuredInfo),
		)
	}
}

func browserRuntimeApplyTopLevelBindingProjection(
	callCtx context.Context,
	payload *browserRuntimePayload,
	projection agentxbrowserruntime.SharedSessionBrowserTopLevelBindingProjection,
) {
	if payload == nil {
		return
	}
	selectionProjection := agentxbrowserruntime.ProjectSharedSessionBrowserSelectionProjectionFromBindingEvaluation(
		projection.Evaluation,
		&agentxbrowserruntime.SharedSessionBrowserSelectionProjection{
			ProfileSelection: projection.ProfileSelection,
			TargetSelection:  projection.TargetSelection,
		},
	)
	if selectionProjection != nil {
		payload.SessionProfileSelection = browserRuntimeSelectionPtrValue(selectionProjection.ProfileSelection)
		payload.SessionTargetSelection = browserRuntimeSessionTargetSelectionPtrFromShared(selectionProjection.TargetSelection)
	} else {
		payload.SessionProfileSelection = nil
		payload.SessionTargetSelection = nil
	}
	payload.SessionBinding = browserRuntimeBuildSessionBinding(
		callCtx,
		nil,
		nil,
		nil,
		payload.SelectedRoute,
		nil,
		&projection.Evaluation,
	)
	if payload.SessionBinding != nil {
		payload.SessionHandoff = agentxbrowserruntime.CloneSharedSessionBrowserSessionHandoffSummary(payload.SessionBinding.SessionHandoff)
	}
	browserRuntimeHydrateSelectionProjectionFromBindingPayload(payload)
	browserRuntimeBackfillBindingIdentityFromPayload(payload)
	browserRuntimeBackfillSelectedRouteFromBindingPayload(payload)
	browserRuntimeRefreshProfileInventoryFromBindingPayload(payload)
	browserRuntimeRefreshDefaultProfileFromBindingPayload(payload)
	browserRuntimeMarkCurrentProfileSelected(payload)
	browserRuntimeRefreshConfiguredProfilesFromBindingPayload(payload)
	browserRuntimeSyncResolverGuidanceSummary(payload)
}
