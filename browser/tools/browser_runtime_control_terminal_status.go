package tools

import (
	"strings"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

type browserRuntimeActionTerminalStatus struct {
	Status               string
	Note                 string
	PreserveExistingNote bool
}

func browserRuntimeApplyActionTerminalStatus(
	payload *browserRuntimePayload,
	terminal browserRuntimeActionTerminalStatus,
) {
	if payload == nil {
		return
	}
	if status := strings.TrimSpace(terminal.Status); status != "" {
		payload.Status = status
	}
	if note := strings.TrimSpace(terminal.Note); note != "" {
		if !terminal.PreserveExistingNote || strings.TrimSpace(payload.Note) == "" {
			payload.Note = note
		}
	}
}

func browserRuntimeApplyUnsupportedActionOutcome(payload *browserRuntimePayload, action string) {
	outcome := agentxbrowserruntime.BuildSharedSessionBrowserUnsupportedActionOutcome(action)
	browserRuntimeApplyActionTerminalStatus(payload, browserRuntimeActionTerminalStatus{
		Status: outcome.Status,
		Note:   outcome.Note,
	})
}

func browserRuntimeApplyUnsupportedRouteActionOutcome(
	payload *browserRuntimePayload,
	action string,
	routeErr error,
) {
	outcome := agentxbrowserruntime.BuildSharedSessionBrowserUnsupportedRouteActionOutcome(action, routeErr)
	browserRuntimeApplyActionTerminalStatus(payload, browserRuntimeActionTerminalStatus{
		Status: outcome.Status,
		Note:   outcome.Note,
	})
}
