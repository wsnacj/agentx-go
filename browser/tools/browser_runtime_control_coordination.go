package tools

import (
	"strings"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

type browserRuntimeCoordinationSummaryProjection struct {
	Clear   bool
	Summary *agentxbrowserruntime.SharedSessionBrowserCoordinationSummary
}

func browserRuntimeClearWorkbenchCoordinationSummary(payload *browserRuntimePayload) {
	browserRuntimeApplyCoordinationSummaryProjection(
		payload,
		browserRuntimeCoordinationSummaryProjection{Clear: true},
	)
}

func browserRuntimeApplyCoordinationSummaryProjection(
	payload *browserRuntimePayload,
	projection browserRuntimeCoordinationSummaryProjection,
) {
	if payload == nil {
		return
	}
	if projection.Clear {
		payload.CoordinationState = ""
		payload.CoordinationDecision = ""
		payload.CoordinationReady = false
		return
	}
	if projection.Summary == nil {
		return
	}
	payload.CoordinationState = projection.Summary.State
	payload.CoordinationDecision = projection.Summary.Decision
	payload.CoordinationReady = projection.Summary.Ready
}

func browserRuntimeBuildCoordinationSummaryProjection(
	payload browserRuntimePayload,
	action string,
	coordinationGoal string,
) browserRuntimeCoordinationSummaryProjection {
	summary, ok := agentxbrowserruntime.BuildSharedSessionBrowserCoordinationSummary(
		agentxbrowserruntime.SharedSessionBrowserCoordinationSummaryInput{
			Action:            action,
			CoordinationGoal:  coordinationGoal,
			Evaluation:        browserRuntimeSharedBindingEvaluationPtrFromPayload(payload),
			CoordinationState: browserRuntimeCoordinationStateFromPayload(payload),
			HealthState:       browserRuntimeCoordinationHealthStateFromPayload(payload),
			ProfileStatus:     browserRuntimeSharedProfileStatusResultPtr(payload.ProfileStatus),
			PrepareDecision:   payload.PrepareDecision,
			RestartDecision:   payload.RestartDecision,
			SyncSessionReady:  payload.SyncSessionReady,
		},
	)
	if !ok {
		return browserRuntimeCoordinationSummaryProjection{}
	}
	return browserRuntimeCoordinationSummaryProjection{Summary: &summary}
}

func browserRuntimeRefreshCoordinationSummary(payload *browserRuntimePayload, action string, coordinationGoal string) {
	if payload == nil {
		return
	}
	browserRuntimeApplyCoordinationSummaryProjection(
		payload,
		browserRuntimeBuildCoordinationSummaryProjection(*payload, action, coordinationGoal),
	)
}

func browserRuntimeCoordinationStateFromPayload(payload browserRuntimePayload) string {
	if payload.SessionBinding != nil && payload.SessionBinding.Coordination != nil {
		return strings.TrimSpace(payload.SessionBinding.Coordination.State)
	}
	return strings.TrimSpace(payload.CoordinationState)
}

func browserRuntimeCoordinationHealthStateFromPayload(payload browserRuntimePayload) string {
	if payload.SessionBinding == nil {
		return ""
	}
	return strings.TrimSpace(payload.SessionBinding.SessionHealthState)
}
