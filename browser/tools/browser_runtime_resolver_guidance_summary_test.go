package tools

import (
	"testing"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

func TestBrowserRuntimeBuildResolverExplanationSummary(t *testing.T) {
	summary := browserRuntimeBuildResolverExplanationSummary(browserRuntimePayload{
		ResolverBlockedBy:         "multiple_candidates_filtered",
		ResolverAmbiguityClass:    "filtered_residual",
		ResolverCandidateKind:     "label",
		ResolverRetryDisposition:  "manual_only",
		ResolverManualRetryHint:   "add_ordinal",
		ResolverNextStepAlias:     "snapshot",
		ResolverSpecificityFields: []string{"tag", "type"},
	})
	if summary == nil {
		t.Fatalf("expected resolver explanation summary to be built")
	}
	if summary.State != "manual_resolution_required" ||
		summary.SummaryCode != "label_filtered_residual" ||
		summary.NextStepAlias != "snapshot" ||
		summary.ManualRetryHint != "add_ordinal" {
		t.Fatalf("unexpected resolver explanation summary: %#v", summary)
	}
}

func TestBrowserRuntimeBuildDiagnosticsExplanationSummary(t *testing.T) {
	summary := browserRuntimeBuildDiagnosticsExplanationSummary(browserRuntimePayload{
		ResolverBlockedBy:         "multiple_candidates_filtered",
		ResolverAmbiguityClass:    "filtered_residual",
		ResolverCandidateKind:     "label",
		ResolverRetryDisposition:  "manual_only",
		ResolverManualRetryHint:   "add_ordinal",
		ResolverNextStepAlias:     "snapshot",
		ResolverSpecificityFields: []string{"tag", "type"},
	})
	if summary == nil {
		t.Fatalf("expected diagnostics explanation summary to be built")
	}
	if summary.Category != "resolver" ||
		summary.State != "manual_resolution_required" ||
		summary.SummaryCode != "label_filtered_residual" ||
		summary.NextStepAlias != "snapshot" ||
		summary.ManualRetryHint != "add_ordinal" {
		t.Fatalf("unexpected diagnostics explanation summary: %#v", summary)
	}
}

func TestBrowserRuntimeBuildResolverExplanationSummaryEmpty(t *testing.T) {
	if summary := browserRuntimeBuildResolverExplanationSummary(browserRuntimePayload{}); summary != nil {
		t.Fatalf("expected empty payload not to build resolver explanation summary, got %#v", summary)
	}
}

func TestBrowserRuntimeBuildDiagnosticsExplanationSummaryEmpty(t *testing.T) {
	if summary := browserRuntimeBuildDiagnosticsExplanationSummary(browserRuntimePayload{}); summary != nil {
		t.Fatalf("expected empty payload not to build diagnostics explanation summary, got %#v", summary)
	}
}

func TestBrowserRuntimeClearResolverGuidanceSummaryClearsExplanation(t *testing.T) {
	payload := browserRuntimePayload{
		ResolverBlockedBy:         "multiple_candidates_filtered",
		ResolverAmbiguityClass:    "filtered_residual",
		ResolverCandidateKind:     "role_label",
		ResolverCandidateStrength: "strong",
		ResolverRetryDisposition:  "manual_only",
		ResolverManualRetryHint:   "add_ordinal",
		ResolverNextStepAlias:     "snapshot",
		ResolverSpecificityFields: []string{"href"},
		ResolverExplanation: &browserRuntimeResolverExplanationSummary{
			State:           "manual_resolution_required",
			SummaryCode:     "role_label_filtered_residual",
			NextStepAlias:   "snapshot",
			ManualRetryHint: "add_ordinal",
		},
		DiagnosticsExplanation: &browserRuntimeDiagnosticsExplanationSummary{
			Category:        "resolver",
			State:           "manual_resolution_required",
			SummaryCode:     "role_label_filtered_residual",
			NextStepAlias:   "snapshot",
			ManualRetryHint: "add_ordinal",
		},
	}

	browserRuntimeClearResolverGuidanceSummary(&payload)

	if payload.ResolverBlockedBy != "" ||
		payload.ResolverAmbiguityClass != "" ||
		payload.ResolverCandidateKind != "" ||
		payload.ResolverCandidateStrength != "" ||
		payload.ResolverRetryDisposition != "" ||
		payload.ResolverManualRetryHint != "" ||
		payload.ResolverNextStepAlias != "" ||
		len(payload.ResolverSpecificityFields) != 0 ||
		payload.ResolverExplanation != nil ||
		payload.DiagnosticsExplanation != nil {
		t.Fatalf("expected clear helper to remove resolver guidance and diagnostics explanation, got %#v", payload)
	}
}

func TestBrowserRuntimeBuildTopLevelSummaryUsesSharedProjectionHelpers(t *testing.T) {
	workbench := browserRuntimeBuildWorkbenchSummary(browserRuntimePayload{
		WorkbenchDiagnostics: &browserRuntimeWorkbenchDiagnosticsSummary{
			Category:             "coordination",
			State:                "action_plan_available",
			SummaryCode:          "workbench_action_plan",
			RepairCommand:        "browser_runtime repair",
			NextStepAlias:        "repair",
			ManualRetryHint:      "rerun",
			ResolvedViaFallback:  true,
			PrimaryBrowserAction: "browser.open",
			PrimaryNodeAction:    "browser_runtime start",
			NextStep:             "repair runtime",
		},
	})
	if workbench == nil {
		t.Fatalf("expected workbench summary to be built")
	}
	if workbench.Category != "coordination" ||
		workbench.SummaryCode != "workbench_action_plan" ||
		workbench.RepairCommand != "browser_runtime repair" ||
		workbench.PrimaryBrowserAction != "browser.open" ||
		workbench.PrimaryNodeAction != "browser_runtime start" ||
		workbench.NextStep != "repair runtime" ||
		!workbench.ResolvedViaFallback {
		t.Fatalf("unexpected workbench summary projection: %#v", workbench)
	}

	diagnostics := browserRuntimeBuildDiagnosticsAlias(browserRuntimePayload{
		DiagnosticsExplanation: &browserRuntimeDiagnosticsExplanationSummary{
			Category:        "resolver",
			State:           "resolved_via_fallback",
			SummaryCode:     "resolver_fallback",
			NextStepAlias:   "snapshot",
			ManualRetryHint: "retry-with-ordinal",
		},
	})
	if diagnostics == nil {
		t.Fatalf("expected diagnostics summary to be built")
	}
	if diagnostics.Category != "resolver" ||
		diagnostics.State != "resolved_via_fallback" ||
		diagnostics.SummaryCode != "resolver_fallback" ||
		diagnostics.NextStepAlias != "snapshot" ||
		diagnostics.ManualRetryHint != "retry-with-ordinal" ||
		!diagnostics.ResolvedViaFallback {
		t.Fatalf("unexpected diagnostics summary projection: %#v", diagnostics)
	}
}

func TestBrowserRuntimeBuildExplanationAliasFallsBackToResolverProjection(t *testing.T) {
	explanation := browserRuntimeBuildExplanationAlias(browserRuntimePayload{
		ResolverExplanation: &browserRuntimeResolverExplanationSummary{
			State:           "manual_resolution_required",
			SummaryCode:     "resolver_manual",
			NextStepAlias:   "snapshot",
			ManualRetryHint: "add-ordinal",
		},
	})
	if explanation == nil {
		t.Fatalf("expected explanation summary to be built from resolver explanation")
	}
	if explanation.Category != "resolver" ||
		explanation.State != "manual_resolution_required" ||
		explanation.SummaryCode != "resolver_manual" ||
		explanation.NextStepAlias != "snapshot" ||
		explanation.ManualRetryHint != "add-ordinal" ||
		explanation.ResolvedViaFallback {
		t.Fatalf("unexpected resolver explanation projection: %#v", explanation)
	}

	fallback := browserRuntimeBuildTopLevelSummary(browserRuntimePayload{
		ResolverExplanation: &browserRuntimeResolverExplanationSummary{
			State:           "resolved_via_fallback",
			SummaryCode:     "resolver_fallback",
			NextStepAlias:   "snapshot",
			ManualRetryHint: "add-ordinal",
		},
	})
	if fallback == nil {
		t.Fatalf("expected top-level summary to be built from resolver fallback explanation")
	}
	if fallback.Category != "resolver_fallback" || !fallback.ResolvedViaFallback {
		t.Fatalf("unexpected resolver fallback projection: %#v", fallback)
	}
}

func TestBrowserRuntimeProjectGuidanceShellBuildsSharedShells(t *testing.T) {
	projection := browserRuntimeProjectGuidanceShell(browserRuntimePayload{
		Action:                        "workbench",
		Status:                        "ok",
		ResolverBlockedBy:             "multiple_candidates_filtered",
		ResolverAmbiguityClass:        "filtered_residual",
		ResolverCandidateKind:         "label",
		ResolverRetryDisposition:      "manual_only",
		ResolverManualRetryHint:       "add_ordinal",
		ResolverNextStepAlias:         "snapshot",
		WorkbenchReady:                true,
		WorkbenchSections:             []string{"route"},
		WorkbenchPrimaryBrowserAction: "browser action=refresh",
		WorkbenchNextStep:             "browser action=refresh",
	}, true)

	if projection.ResolverExplanation == nil ||
		projection.ResolverExplanation.SummaryCode != "label_filtered_residual" {
		t.Fatalf("expected resolver explanation projection, got %#v", projection.ResolverExplanation)
	}
	if projection.DiagnosticsExplanation == nil ||
		projection.DiagnosticsExplanation.Category != "resolver" {
		t.Fatalf("expected diagnostics explanation projection, got %#v", projection.DiagnosticsExplanation)
	}
	if projection.WorkbenchDiagnostics == nil ||
		projection.WorkbenchDiagnostics.PrimaryBrowserAction != "browser action=refresh" ||
		projection.WorkbenchDiagnostics.NextStep != "browser action=refresh" {
		t.Fatalf("expected workbench diagnostics projection, got %#v", projection.WorkbenchDiagnostics)
	}
	if projection.WorkbenchDisplay == nil ||
		!projection.WorkbenchDisplay.Ready ||
		projection.WorkbenchDisplay.Category != "resolver" {
		t.Fatalf("expected workbench display projection, got %#v", projection.WorkbenchDisplay)
	}
	if projection.Explanation == nil ||
		projection.Explanation.Category != "resolver" {
		t.Fatalf("expected top-level explanation projection, got %#v", projection.Explanation)
	}
	if projection.Display == nil ||
		!projection.Display.Ready ||
		projection.Display.Category != "resolver" {
		t.Fatalf("expected top-level display projection, got %#v", projection.Display)
	}
}

func TestBrowserRuntimeProjectGuidanceShellBuildsActionabilityFailureShells(t *testing.T) {
	projection := browserRuntimeProjectGuidanceShell(browserRuntimePayload{
		Action:            "click",
		Status:            "failed",
		WorkbenchReady:    true,
		WorkbenchSections: []string{"diagnostics"},
		Actionability: &agentxbrowserruntime.BrowserActionabilityReport{
			Action:          "click",
			Status:          agentxbrowserruntime.BrowserActionabilityStatusFailed,
			FailedCheck:     "receives_events",
			FailureReason:   "actionability_receives_events_failed",
			ManualRetryHint: "choose_uncovered_target",
			RecoveryAction:  "browser action=snapshot",
		},
		FailureEvidence: &agentxbrowserruntime.BrowserActionFailureEvidence{
			Action:         "click",
			Status:         "failed",
			ReasonCode:     "actionability_receives_events_failed",
			RecoveryAction: "browser action=snapshot",
		},
	}, true)

	if projection.DiagnosticsExplanation == nil ||
		projection.DiagnosticsExplanation.Category != "actionability" ||
		projection.DiagnosticsExplanation.State != "actionability_failed" ||
		projection.DiagnosticsExplanation.SummaryCode != "actionability_receives_events_failed" ||
		projection.DiagnosticsExplanation.NextStepAlias != "snapshot" ||
		projection.DiagnosticsExplanation.ManualRetryHint != "choose_uncovered_target" {
		t.Fatalf("expected actionability diagnostics projection, got %#v", projection.DiagnosticsExplanation)
	}
	if projection.WorkbenchDiagnostics == nil ||
		projection.WorkbenchDiagnostics.Category != "actionability" ||
		projection.WorkbenchDiagnostics.PrimaryBrowserAction != "browser action=snapshot" ||
		projection.WorkbenchDiagnostics.NextStep != "browser action=snapshot" {
		t.Fatalf("expected workbench diagnostics to carry actionability next step, got %#v", projection.WorkbenchDiagnostics)
	}
	if projection.Display == nil ||
		!projection.Display.Ready ||
		projection.Display.Category != "actionability" ||
		projection.Display.PrimaryBrowserAction != "browser action=snapshot" {
		t.Fatalf("expected display to carry actionability next step, got %#v", projection.Display)
	}
}

func TestBrowserRuntimeApplyGuidanceShellProjectionWritesPayloadShells(t *testing.T) {
	payload := browserRuntimePayload{}
	projection := browserRuntimeGuidanceShellProjection{
		ResolverExplanation:    &browserRuntimeResolverExplanationSummary{State: "manual_resolution_required"},
		DiagnosticsExplanation: &browserRuntimeDiagnosticsExplanationSummary{Category: "resolver"},
		WorkbenchExplanation:   &browserRuntimeDiagnosticsExplanationSummary{Category: "resolver"},
		WorkbenchDiagnostics:   &browserRuntimeWorkbenchDiagnosticsSummary{Category: "coordination"},
		WorkbenchSummary:       &browserTopLevelSummary{Category: "coordination"},
		WorkbenchDisplay:       &browserRuntimeWorkbenchDisplaySummary{Category: "coordination"},
		Explanation:            &browserTopLevelSummary{Category: "resolver"},
		Diagnostics:            &browserTopLevelSummary{Category: "resolver"},
		Summary:                &browserTopLevelSummary{Category: "resolver"},
		Display:                &browserTopLevelDisplaySummary{Category: "resolver"},
	}

	browserRuntimeApplyGuidanceShellProjection(&payload, projection)

	if payload.ResolverExplanation == nil || payload.ResolverExplanation.State != "manual_resolution_required" {
		t.Fatalf("expected resolver explanation shell to be written, got %#v", payload.ResolverExplanation)
	}
	if payload.WorkbenchDiagnostics == nil || payload.WorkbenchDiagnostics.Category != "coordination" {
		t.Fatalf("expected workbench diagnostics shell to be written, got %#v", payload.WorkbenchDiagnostics)
	}
	if payload.WorkbenchDisplay == nil || payload.WorkbenchDisplay.Category != "coordination" {
		t.Fatalf("expected workbench display shell to be written, got %#v", payload.WorkbenchDisplay)
	}
	if payload.Display == nil || payload.Display.Category != "resolver" {
		t.Fatalf("expected top-level display shell to be written, got %#v", payload.Display)
	}
}

func TestBrowserRuntimeProjectWorkbenchShellBuildsWorkbenchSummaries(t *testing.T) {
	projection := browserRuntimeProjectWorkbenchShell(browserRuntimePayload{
		Action:                        "workbench",
		Status:                        "ok",
		ResolverBlockedBy:             "multiple_candidates_filtered",
		ResolverAmbiguityClass:        "filtered_residual",
		ResolverCandidateKind:         "label",
		ResolverRetryDisposition:      "manual_only",
		ResolverManualRetryHint:       "add_ordinal",
		ResolverNextStepAlias:         "snapshot",
		WorkbenchReady:                true,
		WorkbenchSections:             []string{"route"},
		WorkbenchPrimaryBrowserAction: "browser action=refresh",
		WorkbenchPrimaryNodeAction:    "nodes action=run_status",
		WorkbenchNextStep:             "browser action=refresh",
		BrowserSurface:                "explicit_managed_opt_in",
		BrowserOptInTargets:           []string{"node"},
	})

	if projection.WorkbenchExplanation == nil ||
		projection.WorkbenchExplanation.Category != "resolver" ||
		projection.WorkbenchExplanation.SummaryCode != "label_filtered_residual" {
		t.Fatalf("expected workbench explanation projection, got %#v", projection.WorkbenchExplanation)
	}
	if projection.WorkbenchDiagnostics == nil ||
		projection.WorkbenchDiagnostics.Category != "resolver" ||
		projection.WorkbenchDiagnostics.PrimaryBrowserAction != "browser action=refresh" ||
		projection.WorkbenchDiagnostics.PrimaryNodeAction != "nodes action=run_status" ||
		projection.WorkbenchDiagnostics.NextStep != "browser action=refresh" {
		t.Fatalf("expected workbench diagnostics projection, got %#v", projection.WorkbenchDiagnostics)
	}
	if projection.WorkbenchSummary == nil ||
		projection.WorkbenchSummary.Category != "resolver" ||
		projection.WorkbenchSummary.SummaryCode != "label_filtered_residual" {
		t.Fatalf("expected workbench summary projection, got %#v", projection.WorkbenchSummary)
	}
	if projection.WorkbenchDisplay == nil ||
		!projection.WorkbenchDisplay.Ready ||
		projection.WorkbenchDisplay.Category != "resolver" ||
		projection.WorkbenchDisplay.SummaryCode != "label_filtered_residual" ||
		projection.WorkbenchDisplay.PrimaryBrowserAction != "browser action=refresh" ||
		projection.WorkbenchDisplay.PrimaryNodeAction != "nodes action=run_status" ||
		projection.WorkbenchDisplay.NextStep != "browser action=refresh" {
		t.Fatalf("expected workbench display projection, got %#v", projection.WorkbenchDisplay)
	}
	if projection.WorkbenchSurface == nil ||
		!projection.WorkbenchSurface.Ready ||
		projection.WorkbenchSurface.BrowserSurface != "explicit_managed_opt_in" ||
		len(projection.WorkbenchSurface.BrowserOptInTargets) != 1 ||
		projection.WorkbenchSurface.BrowserOptInTargets[0] != "node" ||
		projection.WorkbenchSurface.Summary == nil ||
		projection.WorkbenchSurface.Summary.Category != "resolver" {
		t.Fatalf("expected workbench surface projection, got %#v", projection.WorkbenchSurface)
	}
}

func TestBrowserRuntimeApplyWorkbenchShellProjectionWritesPayloadShells(t *testing.T) {
	payload := browserRuntimePayload{}
	projection := browserRuntimeWorkbenchShellProjection{
		WorkbenchExplanation: &browserRuntimeDiagnosticsExplanationSummary{Category: "resolver"},
		WorkbenchDiagnostics: &browserRuntimeWorkbenchDiagnosticsSummary{Category: "coordination"},
		WorkbenchSummary:     &browserTopLevelSummary{Category: "coordination"},
		WorkbenchDisplay:     &browserRuntimeWorkbenchDisplaySummary{Category: "coordination", Ready: true},
		WorkbenchSurface:     &browserRuntimeWorkbenchSurfaceSummary{Ready: true, BrowserSurface: "explicit_managed_opt_in"},
	}

	browserRuntimeApplyWorkbenchShellProjection(&payload, projection)

	if payload.WorkbenchExplanation == nil || payload.WorkbenchExplanation.Category != "resolver" {
		t.Fatalf("expected workbench explanation shell to be written, got %#v", payload.WorkbenchExplanation)
	}
	if payload.WorkbenchDiagnostics == nil || payload.WorkbenchDiagnostics.Category != "coordination" {
		t.Fatalf("expected workbench diagnostics shell to be written, got %#v", payload.WorkbenchDiagnostics)
	}
	if payload.WorkbenchSummary == nil || payload.WorkbenchSummary.Category != "coordination" {
		t.Fatalf("expected workbench summary shell to be written, got %#v", payload.WorkbenchSummary)
	}
	if payload.WorkbenchDisplay == nil || !payload.WorkbenchDisplay.Ready || payload.WorkbenchDisplay.Category != "coordination" {
		t.Fatalf("expected workbench display shell to be written, got %#v", payload.WorkbenchDisplay)
	}
	if payload.Workbench == nil || !payload.Workbench.Ready || payload.Workbench.BrowserSurface != "explicit_managed_opt_in" {
		t.Fatalf("expected workbench surface shell to be written, got %#v", payload.Workbench)
	}
}
