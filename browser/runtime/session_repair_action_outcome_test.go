package browserruntime

import "testing"

func TestBuildSharedSessionBrowserRepairActionOutcomeUsesSuccessGuidanceProjection(t *testing.T) {
	outcome := BuildSharedSessionBrowserRepairActionOutcome(
		SharedSessionBrowserRepairActionOutcomeRequest{
			Status:         "ok",
			Note:           "Bundled browserd bootstrap repair completed.",
			RepairDecision: "repaired",
			RepairReady:    true,
			RepairOutput:   "repair-completed",
		},
	)

	if outcome.Status != "ok" || outcome.RepairDecision != "repaired" || !outcome.RepairReady {
		t.Fatalf("expected repaired outcome fields to be preserved, got %#v", outcome)
	}
	if outcome.GuidanceProjection.Summary == nil ||
		outcome.GuidanceProjection.Summary.Category != "coordination" ||
		outcome.GuidanceProjection.Summary.State != "completed" ||
		outcome.GuidanceProjection.Summary.SummaryCode != "repair_completed" {
		t.Fatalf("expected repaired outcome to project shared success summary, got %#v", outcome.GuidanceProjection.Summary)
	}
	if outcome.GuidanceProjection.Display == nil ||
		outcome.GuidanceProjection.Display.Category != "coordination" ||
		outcome.GuidanceProjection.Display.State != "completed" ||
		outcome.GuidanceProjection.Display.SummaryCode != "repair_completed" {
		t.Fatalf("expected repaired outcome to project shared success display, got %#v", outcome.GuidanceProjection.Display)
	}
}

func TestBuildSharedSessionBrowserRepairActionOutcomeUsesFailureTerminalProjection(t *testing.T) {
	outcome := BuildSharedSessionBrowserRepairActionOutcome(
		SharedSessionBrowserRepairActionOutcomeRequest{
			Status:         "error",
			Note:           "repair-failed",
			RepairDecision: "repair_failed",
			RepairOutput:   "repair-failed",
			RepairAction:   "browser_runtime action=repair",
		},
	)

	if outcome.GuidanceProjection.DiagnosticsExplanation == nil ||
		outcome.GuidanceProjection.DiagnosticsExplanation.State != "failed" ||
		outcome.GuidanceProjection.DiagnosticsExplanation.SummaryCode != "repair_failed" ||
		outcome.GuidanceProjection.DiagnosticsExplanation.NextStepAlias != "repair" ||
		outcome.GuidanceProjection.DiagnosticsExplanation.ManualRetryHint != "repair_bootstrap" {
		t.Fatalf("expected repair failure diagnostics explanation, got %#v", outcome.GuidanceProjection.DiagnosticsExplanation)
	}
	if outcome.GuidanceProjection.Summary == nil ||
		outcome.GuidanceProjection.Summary.PrimaryBrowserAction != "browser_runtime action=repair" ||
		outcome.GuidanceProjection.Summary.NextStep != "browser_runtime action=repair" {
		t.Fatalf("expected repair failure summary to retain repair action command, got %#v", outcome.GuidanceProjection.Summary)
	}
}

func TestBuildSharedSessionBrowserRepairActionOutcomeUsesUnavailableTerminalProjection(t *testing.T) {
	outcome := BuildSharedSessionBrowserRepairActionOutcome(
		SharedSessionBrowserRepairActionOutcomeRequest{
			Status:         "unsupported",
			Note:           "repair unavailable",
			RepairDecision: "repair_command_unavailable",
			DoctorAction:   "browser_runtime action=doctor",
		},
	)

	if outcome.GuidanceProjection.DiagnosticsExplanation == nil ||
		outcome.GuidanceProjection.DiagnosticsExplanation.State != "unsupported" ||
		outcome.GuidanceProjection.DiagnosticsExplanation.SummaryCode != "repair_command_unavailable" {
		t.Fatalf("expected repair unavailable diagnostics explanation, got %#v", outcome.GuidanceProjection.DiagnosticsExplanation)
	}
	if outcome.GuidanceProjection.Summary == nil ||
		outcome.GuidanceProjection.Summary.PrimaryBrowserAction != "browser_runtime action=doctor" ||
		outcome.GuidanceProjection.Summary.NextStep != "browser_runtime action=doctor" {
		t.Fatalf("expected repair unavailable summary to retain doctor action command, got %#v", outcome.GuidanceProjection.Summary)
	}
}
