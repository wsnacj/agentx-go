package tools

import (
	"strings"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

func browserRuntimeSharedWorkbenchProjectionRequest(
	payload *browserRuntimePayload,
	options browserRuntimeWorkbenchProjectionSync,
) agentxbrowserruntime.SharedSessionBrowserWorkbenchProjectionRequest {
	if payload == nil {
		return agentxbrowserruntime.SharedSessionBrowserWorkbenchProjectionRequest{
			ExtraSections:           append([]string(nil), options.ExtraSections...),
			SyncCoordinationSurface: options.SyncCoordinationSurface,
			ClearActionPlan:         options.ClearActionPlan,
		}
	}
	return agentxbrowserruntime.SharedSessionBrowserWorkbenchProjectionRequest{
		HasProfileStatus:        payload.ProfileStatus != nil,
		HasProfiles:             len(payload.Profiles) > 0,
		HasSessionProjection:    browserRuntimeWorkbenchHasSessionProjection(*payload),
		ExtraSections:           append([]string(nil), options.ExtraSections...),
		SyncCoordinationSurface: options.SyncCoordinationSurface,
		ClearActionPlan:         options.ClearActionPlan,
		Evaluation:              browserRuntimeSharedBindingEvaluationPtrFromPayload(*payload),
		CoordinationPlan:        browserRuntimeSharedCoordinationPlanPtrFromPayload(*payload),
	}
}

func browserRuntimeSharedCoordinationPlanPtrFromPayload(
	payload browserRuntimePayload,
) *agentxbrowserruntime.SharedSessionBrowserCoordinationPlan {
	if payload.SessionBinding != nil && payload.SessionBinding.Coordination != nil {
		if plan := normalizeBrowserRuntimeSharedCoordinationPlanPtr(
			ptrBrowserRuntimeSharedCoordinationPlan(
				browserRuntimeSharedCoordinationPlan(*payload.SessionBinding.Coordination),
			),
		); plan != nil {
			return plan
		}
	}
	if evaluation := browserRuntimeSharedBindingEvaluationPtrFromPayload(payload); evaluation != nil {
		return normalizeBrowserRuntimeSharedCoordinationPlanPtr(&evaluation.Coordination.Plan)
	}
	return nil
}

func browserRuntimeSharedCoordinationPlan(
	coordination browserRuntimeCoordination,
) agentxbrowserruntime.SharedSessionBrowserCoordinationPlan {
	return agentxbrowserruntime.SharedSessionBrowserCoordinationPlan{
		State:                     strings.TrimSpace(coordination.State),
		BrowserOnNode:             coordination.BrowserOnNode,
		HasActiveNodeRun:          coordination.HasActiveNodeRun,
		HasRunningBrowserProfile:  coordination.HasRunningBrowserProfile,
		SyncAction:                strings.TrimSpace(coordination.SyncBrowserAction),
		PrepareAction:             strings.TrimSpace(coordination.PrepareBrowserAction),
		RestartAction:             strings.TrimSpace(coordination.RestartBrowserAction),
		TeardownAction:            strings.TrimSpace(coordination.TeardownBrowserAction),
		PrimaryBrowserAction:      strings.TrimSpace(coordination.PrimaryBrowserAction),
		PrimaryNodeAction:         strings.TrimSpace(coordination.PrimaryNodeAction),
		NextStep:                  strings.TrimSpace(coordination.NextStep),
		RecommendedBrowserActions: append([]string(nil), coordination.RecommendedBrowserActions...),
		RecommendedNodeActions:    append([]string(nil), coordination.RecommendedNodeActions...),
	}
}

func ptrBrowserRuntimeSharedCoordinationPlan(
	plan agentxbrowserruntime.SharedSessionBrowserCoordinationPlan,
) *agentxbrowserruntime.SharedSessionBrowserCoordinationPlan {
	return &plan
}

func normalizeBrowserRuntimeSharedCoordinationPlanPtr(
	plan *agentxbrowserruntime.SharedSessionBrowserCoordinationPlan,
) *agentxbrowserruntime.SharedSessionBrowserCoordinationPlan {
	if plan == nil {
		return nil
	}
	normalized := agentxbrowserruntime.SharedSessionBrowserCoordinationPlan{
		State:                     strings.TrimSpace(plan.State),
		BrowserOnNode:             plan.BrowserOnNode,
		HasActiveNodeRun:          plan.HasActiveNodeRun,
		HasRunningBrowserProfile:  plan.HasRunningBrowserProfile,
		NeedsSessionSync:          plan.NeedsSessionSync,
		SyncAction:                strings.TrimSpace(plan.SyncAction),
		PrepareAction:             strings.TrimSpace(plan.PrepareAction),
		RestartAction:             strings.TrimSpace(plan.RestartAction),
		TeardownAction:            strings.TrimSpace(plan.TeardownAction),
		PrimaryBrowserAction:      strings.TrimSpace(plan.PrimaryBrowserAction),
		PrimaryNodeAction:         strings.TrimSpace(plan.PrimaryNodeAction),
		NextStep:                  strings.TrimSpace(plan.NextStep),
		RecommendedBrowserActions: append([]string(nil), plan.RecommendedBrowserActions...),
		RecommendedNodeActions:    append([]string(nil), plan.RecommendedNodeActions...),
	}
	if normalized.State == "" &&
		normalized.SyncAction == "" &&
		normalized.PrepareAction == "" &&
		normalized.RestartAction == "" &&
		normalized.TeardownAction == "" &&
		normalized.PrimaryBrowserAction == "" &&
		normalized.PrimaryNodeAction == "" &&
		normalized.NextStep == "" &&
		len(normalized.RecommendedBrowserActions) == 0 &&
		len(normalized.RecommendedNodeActions) == 0 &&
		!normalized.BrowserOnNode &&
		!normalized.HasActiveNodeRun &&
		!normalized.HasRunningBrowserProfile &&
		!normalized.NeedsSessionSync {
		return nil
	}
	return &normalized
}

func browserRuntimeApplySharedWorkbenchProjection(
	payload *browserRuntimePayload,
	projection agentxbrowserruntime.SharedSessionBrowserWorkbenchProjection,
) {
	if payload == nil {
		return
	}
	payload.WorkbenchSections = append([]string(nil), projection.Sections...)
	payload.WorkbenchReady = projection.Ready
	if projection.ClearActionPlan {
		browserRuntimeClearWorkbenchActionPlan(payload)
	}
	if projection.ActionPlan != nil {
		browserRuntimeApplySharedWorkbenchActionPlan(payload, *projection.ActionPlan)
	}
	browserRuntimeApplyCoordinationSummaryProjection(
		payload,
		browserRuntimeCoordinationSummaryProjection{
			Clear:   projection.ClearCoordinationSummary,
			Summary: projection.CoordinationSummary,
		},
	)
}

func browserRuntimeApplySharedWorkbenchActionPlan(
	payload *browserRuntimePayload,
	plan agentxbrowserruntime.SharedSessionBrowserWorkbenchActionPlan,
) {
	if payload == nil {
		return
	}
	payload.WorkbenchPrimaryBrowserAction = strings.TrimSpace(plan.PrimaryBrowserAction)
	payload.WorkbenchPrimaryNodeAction = strings.TrimSpace(plan.PrimaryNodeAction)
	payload.WorkbenchNextStep = strings.TrimSpace(plan.NextStep)
	payload.WorkbenchRecommendedBrowserActions = mergeToolMetadataStrings(nil, plan.RecommendedBrowserActions)
	payload.WorkbenchRecommendedNodeActions = mergeToolMetadataStrings(nil, plan.RecommendedNodeActions)
}
