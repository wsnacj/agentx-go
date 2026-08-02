package browserruntime

import "strings"

// SharedSessionBrowserRepairActionOutcome captures the shared tool-facing
// repair result contract so tools callers do not keep local repair-specific
// guidance/terminal summary builders.
type SharedSessionBrowserRepairActionOutcome struct {
	Status             string
	Note               string
	RepairDecision     string
	RepairReady        bool
	RepairOutput       string
	GuidanceProjection SharedSessionBrowserGuidanceProjection
}

// SharedSessionBrowserRepairActionOutcomeRequest carries the minimal repair
// execution result plus operator-facing action commands needed to lower repair
// outcomes into shared guidance shells.
type SharedSessionBrowserRepairActionOutcomeRequest struct {
	Status         string
	Note           string
	RepairDecision string
	RepairReady    bool
	RepairOutput   string
	RepairAction   string
	DoctorAction   string
}

func buildSharedSessionBrowserRepairTerminalProjection(
	state string,
	summaryCode string,
	nextStepAlias string,
	manualRetryHint string,
	primaryBrowserAction string,
) SharedSessionBrowserGuidanceProjection {
	diagnosticsExplanation := &SharedSessionBrowserSummary{
		Category:        "coordination",
		State:           strings.TrimSpace(state),
		SummaryCode:     strings.TrimSpace(summaryCode),
		NextStepAlias:   strings.TrimSpace(nextStepAlias),
		ManualRetryHint: strings.TrimSpace(manualRetryHint),
	}
	summary := cloneSharedSessionBrowserSummary(diagnosticsExplanation)
	if summary != nil {
		summary.PrimaryBrowserAction = strings.TrimSpace(primaryBrowserAction)
		summary.NextStep = strings.TrimSpace(primaryBrowserAction)
	}
	var display *SharedSessionBrowserDisplay
	if summary != nil {
		display = &SharedSessionBrowserDisplay{
			SharedSessionBrowserSummary: *cloneSharedSessionBrowserSummary(summary),
		}
	}
	return SharedSessionBrowserGuidanceProjection{
		DiagnosticsExplanation: cloneSharedSessionBrowserSummary(diagnosticsExplanation),
		Explanation:            cloneSharedSessionBrowserSummary(summary),
		Diagnostics:            cloneSharedSessionBrowserSummary(summary),
		Summary:                cloneSharedSessionBrowserSummary(summary),
		Display:                display,
	}
}

// BuildSharedSessionBrowserRepairActionOutcome lowers repair execution results
// into the shared guidance/result contract consumed by tools payload bridges.
func BuildSharedSessionBrowserRepairActionOutcome(
	req SharedSessionBrowserRepairActionOutcomeRequest,
) SharedSessionBrowserRepairActionOutcome {
	outcome := SharedSessionBrowserRepairActionOutcome{
		Status:         strings.TrimSpace(req.Status),
		Note:           strings.TrimSpace(req.Note),
		RepairDecision: strings.TrimSpace(req.RepairDecision),
		RepairReady:    req.RepairReady,
		RepairOutput:   strings.TrimSpace(req.RepairOutput),
	}
	switch outcome.RepairDecision {
	case "repaired":
		outcome.GuidanceProjection = BuildSharedSessionBrowserGuidanceProjection(
			SharedSessionBrowserGuidanceProjectionRequest{
				ActionKind:   "repair",
				ActionStatus: "ok",
			},
		)
	case "repair_failed":
		outcome.GuidanceProjection = buildSharedSessionBrowserRepairTerminalProjection(
			"failed",
			"repair_failed",
			"repair",
			"repair_bootstrap",
			req.RepairAction,
		)
	case "repair_command_unavailable":
		outcome.GuidanceProjection = buildSharedSessionBrowserRepairTerminalProjection(
			"unsupported",
			"repair_command_unavailable",
			"",
			"",
			req.DoctorAction,
		)
	}
	return outcome
}
