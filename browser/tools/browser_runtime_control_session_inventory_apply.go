package tools

import (
	"strings"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

type browserRuntimeTopLevelSessionProjection struct {
	Routes                  []browserRuntimeSessionRoute
	TargetCount             int
	Runs                    []browserRuntimeSessionRun
	Profiles                []browserRuntimeProfileState
	Handoff                 *browserRuntimeSessionHandoffSummary
	ConfiguredInfo          BrowserRuntimeInfo
	ApplyConfiguredProfiles bool
	MissingSessionIDNote    string
}

func browserRuntimeApplyTopLevelSessionProjection(
	payload *browserRuntimePayload,
	projection browserRuntimeTopLevelSessionProjection,
) {
	if payload == nil {
		return
	}
	payload.SessionRoutes = projection.Routes
	payload.SessionTargetCount = projection.TargetCount
	payload.SessionRuns = projection.Runs
	payload.SessionProfiles = projection.Profiles
	payload.SessionHandoff = agentxbrowserruntime.CloneSharedSessionBrowserSessionHandoffSummary(projection.Handoff)
	if projection.ApplyConfiguredProfiles {
		browserRuntimeApplyConfiguredProfilesProjection(
			payload,
			browserRuntimeConfiguredProfilesProjectionForPayload(*payload, projection.ConfiguredInfo),
		)
	}
	if note := strings.TrimSpace(projection.MissingSessionIDNote); note != "" && payload.SessionID == "" {
		browserRuntimeApplyActionTerminalStatus(payload, browserRuntimeActionTerminalStatus{
			Note:                 note,
			PreserveExistingNote: true,
		})
	}
}
