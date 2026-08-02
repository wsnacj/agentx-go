package tools

import (
	"strings"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

func browserRuntimeApplySessionActionOutcome(
	payload *browserRuntimePayload,
	outcome agentxbrowserruntime.SharedSessionBrowserActionOutcome,
) {
	if payload == nil {
		return
	}
	if decision, ok := agentxbrowserruntime.ProjectSharedSessionBrowserActionDecision(outcome); ok {
		switch decision.Action {
		case "select_profile":
			payload.SelectDecision = decision.Decision
			payload.SelectReady = decision.Ready
		case "sync_session":
			payload.SyncSessionDecision = decision.Decision
			payload.SyncSessionReady = decision.Ready
		case "select_target":
			payload.SelectTargetDecision = decision.Decision
			payload.SelectTargetReady = decision.Ready
		case "remember_profile":
			payload.RememberDecision = decision.Decision
			payload.RememberReady = decision.Ready
		case "clear_profile", "clear_session", "clear_target":
			if decision.ClearProfileStatus {
				payload.ProfileStatus = nil
			}
			switch decision.Action {
			case "clear_profile":
				payload.ClearDecision = decision.Decision
				payload.ClearReady = decision.Ready
			case "clear_session":
				payload.ClearSessionDecision = decision.Decision
				payload.ClearSessionReady = decision.Ready
				payload.ClearedSessionProfiles = decision.ClearedSessionProfiles
				payload.ClearedSessionTargets = decision.ClearedSessionTargets
			case "clear_target":
				payload.ClearTargetDecision = decision.Decision
				payload.ClearTargetReady = decision.Ready
			}
		}
	}
	if outcome.Err != nil {
		browserRuntimeApplyActionTerminalStatus(payload, browserRuntimeActionTerminalStatus{
			Status: firstNonEmpty(strings.TrimSpace(outcome.Status), "error"),
			Note:   firstNonEmpty(strings.TrimSpace(outcome.Note), outcome.Err.Error()),
		})
		if outcome.ApplyCoordinationDecisionOnError {
			payload.CoordinationDecision = strings.TrimSpace(outcome.Decision)
		}
		return
	}
	browserRuntimeApplyActionTerminalStatus(payload, browserRuntimeActionTerminalStatus{
		Status:               strings.TrimSpace(outcome.Status),
		Note:                 strings.TrimSpace(outcome.Note),
		PreserveExistingNote: true,
	})
	if outcome.SelectionProjection != nil {
		browserRuntimeApplySessionSelectionProjection(payload, browserRuntimeSessionSelectionProjection{
			ProfileSelection:   browserRuntimeSelectionPtrValue(outcome.SelectionProjection.ProfileSelection),
			TargetSelection:    browserRuntimeSessionTargetSelectionPtrFromShared(outcome.SelectionProjection.TargetSelection),
			ApplyTargetToRoute: outcome.SelectionProjection.ApplyTargetToRoute,
		})
	}
	if outcome.ProfileInventoryProjection != nil {
		browserRuntimeApplyTopLevelProfileInventory(
			payload,
			*browserRuntimeTopLevelProfileInventoryProjectionFromShared(BrowserRuntimeInfo{}, outcome.ProfileInventoryProjection),
		)
	}
}

func browserRuntimeApplyMissingSessionActionInput(
	payload *browserRuntimePayload,
	action string,
	decision string,
) {
	browserRuntimeApplySessionActionOutcome(
		payload,
		agentxbrowserruntime.BuildSharedSessionBrowserMissingInputActionOutcome(action, decision),
	)
}

func browserRuntimeApplyMissingSelectTargetSelection(
	payload *browserRuntimePayload,
	decision string,
	note string,
) {
	browserRuntimeApplySessionActionOutcome(
		payload,
		agentxbrowserruntime.BuildSharedSessionBrowserSelectionActionOutcome(
			agentxbrowserruntime.SharedSessionBrowserSelectionActionOutcomeRequest{
				Action: "select_target",
				DispatchResult: agentxbrowserruntime.SharedSessionBrowserSelectionActionDispatchResult{
					Decision: strings.TrimSpace(decision),
					Note:     strings.TrimSpace(note),
				},
				MissingNote: strings.TrimSpace(note),
			},
		),
	)
}
