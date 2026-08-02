package tools

import (
	"strings"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

func browserRuntimeSyncResolverGuidanceSummary(payload *browserRuntimePayload) {
	if payload == nil || payload.SessionBinding == nil {
		browserRuntimeClearResolverGuidanceSummary(payload)
		return
	}
	binding := payload.SessionBinding
	payload.ResolverBlockedBy = strings.TrimSpace(binding.SessionHealthResolverBlockedBy)
	payload.ResolverAmbiguityClass = strings.TrimSpace(binding.SessionHealthResolverAmbiguityClass)
	payload.ResolverCandidateKind = strings.TrimSpace(binding.SessionHealthResolverCandidateKind)
	payload.ResolverCandidateStrength = strings.TrimSpace(binding.SessionHealthResolverStrength)
	payload.ResolverRetryDisposition = strings.TrimSpace(binding.SessionHealthResolverRetryDisposition)
	payload.ResolverManualRetryHint = strings.TrimSpace(binding.SessionHealthResolverManualRetryHint)
	payload.ResolverNextStepAlias = strings.TrimSpace(binding.SessionHealthResolverNextStepAlias)
	if len(binding.SessionHealthResolverSpecificityFields) == 0 {
		payload.ResolverSpecificityFields = nil
	} else {
		payload.ResolverSpecificityFields = append([]string(nil), binding.SessionHealthResolverSpecificityFields...)
	}
	browserRuntimeSyncSharedGuidanceProjection(payload, false)
	browserRuntimeSyncTopLevelSurfaceSummary(payload)
}

func browserRuntimeClearResolverGuidanceSummary(payload *browserRuntimePayload) {
	if payload == nil {
		return
	}
	payload.ResolverBlockedBy = ""
	payload.ResolverAmbiguityClass = ""
	payload.ResolverCandidateKind = ""
	payload.ResolverCandidateStrength = ""
	payload.ResolverRetryDisposition = ""
	payload.ResolverManualRetryHint = ""
	payload.ResolverNextStepAlias = ""
	payload.ResolverSpecificityFields = nil
	payload.ResolverExplanation = nil
	payload.DiagnosticsExplanation = nil
	payload.WorkbenchExplanation = nil
	payload.WorkbenchDiagnostics = nil
	payload.WorkbenchSummary = nil
	payload.Workbench = nil
	payload.WorkbenchDisplay = nil
	payload.Review = nil
	payload.Explanation = nil
	payload.Diagnostics = nil
	payload.Summary = nil
	payload.Display = nil
	payload.Surface = nil
	payload.View = nil
}

func browserRuntimeSyncResolverExplanationSummary(payload *browserRuntimePayload) {
	if payload == nil {
		return
	}
	explanation := browserRuntimeBuildResolverExplanationSummary(*payload)
	if explanation == nil {
		payload.ResolverExplanation = nil
		return
	}
	payload.ResolverExplanation = explanation
}

func browserRuntimeSyncDiagnosticsExplanationSummary(payload *browserRuntimePayload) {
	if payload == nil {
		return
	}
	explanation := browserRuntimeBuildDiagnosticsExplanationSummary(*payload)
	if explanation == nil {
		payload.DiagnosticsExplanation = nil
		return
	}
	payload.DiagnosticsExplanation = explanation
}

func browserRuntimeSyncWorkbenchExplanationSummary(payload *browserRuntimePayload) {
	if payload == nil {
		return
	}
	if !browserRuntimeHasWorkbenchSurface(*payload) {
		payload.WorkbenchExplanation = nil
		return
	}
	explanation := payload.DiagnosticsExplanation
	if explanation == nil {
		explanation = browserRuntimeBuildDiagnosticsExplanationSummary(*payload)
	}
	if explanation == nil {
		payload.WorkbenchExplanation = nil
		return
	}
	copied := *explanation
	payload.WorkbenchExplanation = &copied
}

func browserRuntimeSyncWorkbenchDiagnosticsSummary(payload *browserRuntimePayload) {
	if payload == nil {
		return
	}
	if !browserRuntimeHasWorkbenchSurface(*payload) {
		payload.WorkbenchDiagnostics = nil
		return
	}
	explanation := payload.WorkbenchExplanation
	if explanation == nil {
		explanation = payload.DiagnosticsExplanation
	}
	primaryBrowserAction := strings.TrimSpace(payload.WorkbenchPrimaryBrowserAction)
	primaryNodeAction := strings.TrimSpace(payload.WorkbenchPrimaryNodeAction)
	nextStep := strings.TrimSpace(payload.WorkbenchNextStep)
	if explanation == nil && primaryBrowserAction == "" && primaryNodeAction == "" && nextStep == "" {
		payload.WorkbenchDiagnostics = nil
		return
	}
	summary := &browserRuntimeWorkbenchDiagnosticsSummary{
		PrimaryBrowserAction: primaryBrowserAction,
		PrimaryNodeAction:    primaryNodeAction,
		NextStep:             nextStep,
	}
	if explanation != nil {
		summary.Category = strings.TrimSpace(explanation.Category)
		summary.State = strings.TrimSpace(explanation.State)
		summary.SummaryCode = strings.TrimSpace(explanation.SummaryCode)
		summary.NextStepAlias = strings.TrimSpace(explanation.NextStepAlias)
		summary.ManualRetryHint = strings.TrimSpace(explanation.ManualRetryHint)
		if strings.EqualFold(summary.State, "resolved_via_fallback") {
			summary.ResolvedViaFallback = true
		}
	} else {
		summary.Category = "coordination"
		summary.State = "action_plan_available"
		summary.SummaryCode = "workbench_action_plan"
	}
	payload.WorkbenchDiagnostics = summary
	browserRuntimeApplyRepairCommandToPayloadShells(payload)
}

func browserRuntimeSyncWorkbenchSummary(payload *browserRuntimePayload) {
	if payload == nil {
		return
	}
	summary := browserRuntimeBuildWorkbenchSummary(*payload)
	if summary == nil {
		payload.WorkbenchSummary = nil
		return
	}
	payload.WorkbenchSummary = summary
	browserRuntimeApplyDefaultCandidateRouteToPayloadShells(payload)
	browserRuntimeApplyRepairCommandToPayloadShells(payload)
}

func browserRuntimeSyncWorkbenchSurfaceSummary(payload *browserRuntimePayload) {
	if payload == nil {
		return
	}
	payload.Workbench = browserRuntimeWorkbenchSurfaceSummaryFromShared(
		agentxbrowserruntime.BuildSharedSessionBrowserWorkbenchSurface(
			browserRuntimeSharedWorkbenchSurfaceRequest(payload),
		),
	)
	browserRuntimeApplyDefaultCandidateRouteToPayloadShells(payload)
	browserRuntimeApplyRepairCommandToPayloadShells(payload)
}

func browserRuntimeSyncReviewSurfaceSummary(payload *browserRuntimePayload) {
	if payload == nil {
		return
	}
	browserRuntimeSyncTopLevelSurfaceSummary(payload)
}

func browserRuntimeSyncTopLevelSurfaceSummary(payload *browserRuntimePayload) {
	if payload == nil {
		return
	}
	browserRuntimeApplyTopLevelSurfaceProjection(
		payload,
		browserRuntimeProjectTopLevelSurface(*payload),
	)
	browserRuntimeApplyDefaultCandidateRouteToPayloadShells(payload)
	browserRuntimeApplyRepairCommandToPayloadShells(payload)
}

func browserRuntimeSyncExplanationAlias(payload *browserRuntimePayload) {
	if payload == nil {
		return
	}
	explanation := browserRuntimeBuildExplanationAlias(*payload)
	if explanation == nil {
		payload.Explanation = nil
		return
	}
	payload.Explanation = explanation
}

func browserRuntimeSyncDiagnosticsAlias(payload *browserRuntimePayload) {
	if payload == nil {
		return
	}
	diagnostics := browserRuntimeBuildDiagnosticsAlias(*payload)
	if diagnostics == nil {
		payload.Diagnostics = nil
		return
	}
	payload.Diagnostics = diagnostics
}

func browserRuntimeSyncTopLevelSummary(payload *browserRuntimePayload) {
	if payload == nil {
		return
	}
	summary := browserRuntimeBuildTopLevelSummary(*payload)
	if summary == nil {
		payload.Summary = nil
		return
	}
	payload.Summary = summary
	browserRuntimeApplyDefaultCandidateRouteToPayloadShells(payload)
	browserRuntimeApplyRepairCommandToPayloadShells(payload)
}

func browserRuntimeSyncDisplayAlias(payload *browserRuntimePayload) {
	if payload == nil {
		return
	}
	display := browserRuntimeBuildDisplayAlias(*payload)
	if display == nil {
		payload.Display = nil
		return
	}
	payload.Display = display
	browserRuntimeApplyRepairCommandToPayloadShells(payload)
}

func browserRuntimeBuildWorkbenchSummary(payload browserRuntimePayload) *browserTopLevelSummary {
	if payload.WorkbenchSummary != nil {
		copied := *payload.WorkbenchSummary
		if browserUnifiedSummaryEmpty(copied) {
			return nil
		}
		return &copied
	}
	if summary := browserRuntimeTopLevelSummaryFromWorkbenchDiagnostics(payload.WorkbenchDiagnostics); summary != nil {
		return summary
	}
	if summary := browserRuntimeTopLevelSummaryFromDiagnosticsExplanation(payload.WorkbenchExplanation); summary != nil {
		return summary
	}
	return nil
}

func browserRuntimeBuildDisplayAlias(payload browserRuntimePayload) *browserTopLevelDisplaySummary {
	if payload.WorkbenchDisplay != nil {
		return browserTopLevelDisplayFromWorkbenchDisplay(payload.WorkbenchDisplay)
	}
	if payload.Summary != nil {
		return browserTopLevelDisplayFromSummary(payload.Summary)
	}
	if payload.Diagnostics != nil {
		return browserTopLevelDisplayFromSummary(payload.Diagnostics)
	}
	if payload.Explanation != nil {
		return browserTopLevelDisplayFromSummary(payload.Explanation)
	}
	return nil
}

func browserRuntimeBuildExplanationAlias(payload browserRuntimePayload) *browserTopLevelSummary {
	if summary := browserRuntimeTopLevelSummaryFromDiagnosticsExplanation(payload.WorkbenchExplanation); summary != nil {
		return summary
	}
	if summary := browserRuntimeTopLevelSummaryFromDiagnosticsExplanation(payload.DiagnosticsExplanation); summary != nil {
		return summary
	}
	if summary := browserRuntimeTopLevelSummaryFromResolverExplanation(payload.ResolverExplanation); summary != nil {
		return summary
	}
	return nil
}

func browserRuntimeBuildDiagnosticsAlias(payload browserRuntimePayload) *browserTopLevelSummary {
	if summary := browserRuntimeBuildWorkbenchSummary(payload); summary != nil {
		return summary
	}
	if summary := browserRuntimeTopLevelSummaryFromDiagnosticsExplanation(payload.DiagnosticsExplanation); summary != nil {
		return summary
	}
	return nil
}

func browserRuntimeBuildTopLevelSummary(payload browserRuntimePayload) *browserTopLevelSummary {
	if summary := browserRuntimeBuildWorkbenchSummary(payload); summary != nil {
		return summary
	}
	if summary := browserRuntimeTopLevelSummaryFromDiagnosticsExplanation(payload.DiagnosticsExplanation); summary != nil {
		return summary
	}
	if summary := browserRuntimeTopLevelSummaryFromResolverExplanation(payload.ResolverExplanation); summary != nil {
		return summary
	}
	return nil
}

func browserRuntimeTopLevelSummaryFromWorkbenchDiagnostics(
	summary *browserRuntimeWorkbenchDiagnosticsSummary,
) *browserTopLevelSummary {
	if summary == nil {
		return nil
	}
	out := &browserTopLevelSummary{
		Category:             strings.TrimSpace(summary.Category),
		State:                strings.TrimSpace(summary.State),
		SummaryCode:          strings.TrimSpace(summary.SummaryCode),
		RepairCommand:        strings.TrimSpace(summary.RepairCommand),
		NextStepAlias:        strings.TrimSpace(summary.NextStepAlias),
		ManualRetryHint:      strings.TrimSpace(summary.ManualRetryHint),
		ResolvedViaFallback:  summary.ResolvedViaFallback,
		PrimaryBrowserAction: strings.TrimSpace(summary.PrimaryBrowserAction),
		PrimaryNodeAction:    strings.TrimSpace(summary.PrimaryNodeAction),
		NextStep:             strings.TrimSpace(summary.NextStep),
	}
	if browserUnifiedSummaryEmpty(*out) {
		return nil
	}
	return out
}

func browserRuntimeTopLevelSummaryFromDiagnosticsExplanation(
	explanation *browserRuntimeDiagnosticsExplanationSummary,
) *browserTopLevelSummary {
	if explanation == nil {
		return nil
	}
	out := &browserTopLevelSummary{
		Category:        strings.TrimSpace(explanation.Category),
		State:           strings.TrimSpace(explanation.State),
		SummaryCode:     strings.TrimSpace(explanation.SummaryCode),
		NextStepAlias:   strings.TrimSpace(explanation.NextStepAlias),
		ManualRetryHint: strings.TrimSpace(explanation.ManualRetryHint),
	}
	if strings.EqualFold(out.State, "resolved_via_fallback") {
		out.ResolvedViaFallback = true
	}
	if browserUnifiedSummaryEmpty(*out) {
		return nil
	}
	return out
}

func browserRuntimeTopLevelSummaryFromResolverExplanation(
	explanation *browserRuntimeResolverExplanationSummary,
) *browserTopLevelSummary {
	if explanation == nil {
		return nil
	}
	out := &browserTopLevelSummary{
		State:           strings.TrimSpace(explanation.State),
		SummaryCode:     strings.TrimSpace(explanation.SummaryCode),
		NextStepAlias:   strings.TrimSpace(explanation.NextStepAlias),
		ManualRetryHint: strings.TrimSpace(explanation.ManualRetryHint),
	}
	if strings.EqualFold(out.State, "resolved_via_fallback") {
		out.Category = "resolver_fallback"
		out.ResolvedViaFallback = true
	} else if out.State != "" || out.SummaryCode != "" || out.NextStepAlias != "" || out.ManualRetryHint != "" {
		out.Category = "resolver"
	}
	if browserUnifiedSummaryEmpty(*out) {
		return nil
	}
	return out
}

func browserRuntimeBuildResolverExplanationSummary(payload browserRuntimePayload) *browserRuntimeResolverExplanationSummary {
	projection := agentxbrowserruntime.BuildSharedSessionBrowserGuidanceProjection(
		browserRuntimeSharedGuidanceProjectionRequest(&payload, false),
	)
	return browserRuntimeResolverExplanationSummaryFromShared(projection.ResolverExplanation)
}

func browserRuntimeBuildDiagnosticsExplanationSummary(payload browserRuntimePayload) *browserRuntimeDiagnosticsExplanationSummary {
	projection := agentxbrowserruntime.BuildSharedSessionBrowserGuidanceProjection(
		browserRuntimeSharedGuidanceProjectionRequest(&payload, false),
	)
	return browserRuntimeDiagnosticsExplanationSummaryFromShared(projection.DiagnosticsExplanation)
}

func browserRuntimeHasWorkbenchSurface(payload browserRuntimePayload) bool {
	if strings.EqualFold(strings.TrimSpace(payload.Action), "workbench") {
		return true
	}
	return payload.WorkbenchReady || len(payload.WorkbenchSections) > 0
}
