package tools

import (
	"context"
	"strings"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

func browserRuntimeApplyLifecycleResultProjection(
	payload *browserRuntimePayload,
	outcome agentxbrowserruntime.SharedSessionBrowserLifecycleActionOutcome,
) {
	if payload == nil {
		return
	}
	payload.PreparedProfile = strings.TrimSpace(outcome.PreparedProfile)
	payload.PrepareDecision = strings.TrimSpace(outcome.PrepareDecision)
	payload.PrepareReady = outcome.PrepareReady
	payload.RestartDecision = strings.TrimSpace(outcome.RestartDecision)
	payload.RestartReady = outcome.RestartReady
	payload.StopDecision = strings.TrimSpace(outcome.StopDecision)
	payload.StopReady = outcome.StopReady
	payload.CreateDecision = strings.TrimSpace(outcome.CreateDecision)
	payload.CreateReady = outcome.CreateReady
	payload.DeleteDecision = strings.TrimSpace(outcome.DeleteDecision)
	payload.DeleteReady = outcome.DeleteReady
}

func browserRuntimeApplyLifecycleActionOutcome(
	callCtx context.Context,
	repairScript string,
	payload *browserRuntimePayload,
	selectedInfo BrowserRuntimeInfo,
	outcome agentxbrowserruntime.SharedSessionBrowserLifecycleActionOutcome,
) {
	if payload == nil {
		return
	}
	browserRuntimeApplyLifecycleResultProjection(payload, outcome)
	browserRuntimeApplyPrepareResultSurface(
		callCtx,
		payload,
		selectedInfo,
		outcome.ExecutionProjection,
	)
	browserRuntimeMaybeApplyLifecycleLaunchDiagnostics(repairScript, payload, outcome.Action, outcome.Result, selectedInfo)
	if outcome.Err != nil {
		browserRuntimeApplyActionTerminalStatus(payload, browserRuntimeActionTerminalStatus{
			Status: firstNonEmpty(strings.TrimSpace(outcome.Status), "error"),
			Note:   firstNonEmpty(strings.TrimSpace(outcome.Note), outcome.Err.Error()),
		})
		if outcome.ApplyCoordinationDecisionOnError {
			payload.CoordinationDecision = strings.TrimSpace(outcome.Result.Decision)
		}
		return
	}
	if outcome.RememberOutcome != nil {
		browserRuntimeApplySessionActionOutcome(payload, *outcome.RememberOutcome)
	}
}
