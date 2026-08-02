package tools

import (
	"strings"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

type browserRuntimeGuidanceShellProjection struct {
	ResolverExplanation    *browserRuntimeResolverExplanationSummary
	DiagnosticsExplanation *browserRuntimeDiagnosticsExplanationSummary
	WorkbenchExplanation   *browserRuntimeDiagnosticsExplanationSummary
	WorkbenchDiagnostics   *browserRuntimeWorkbenchDiagnosticsSummary
	WorkbenchSummary       *browserTopLevelSummary
	WorkbenchDisplay       *browserRuntimeWorkbenchDisplaySummary
	Explanation            *browserTopLevelSummary
	Diagnostics            *browserTopLevelSummary
	Summary                *browserTopLevelSummary
	Display                *browserTopLevelDisplaySummary
}

func browserRuntimeSharedGuidanceProjectionRequest(
	payload *browserRuntimePayload,
	includeWorkbenchSurface bool,
) agentxbrowserruntime.SharedSessionBrowserGuidanceProjectionRequest {
	if payload == nil {
		return agentxbrowserruntime.SharedSessionBrowserGuidanceProjectionRequest{
			IncludeWorkbenchSurface: includeWorkbenchSurface,
		}
	}
	req := agentxbrowserruntime.SharedSessionBrowserGuidanceProjectionRequest{
		IncludeWorkbenchSurface:       includeWorkbenchSurface,
		ActionKind:                    payload.Action,
		ActionStatus:                  payload.Status,
		ResolverBlockedBy:             payload.ResolverBlockedBy,
		ResolverAmbiguityClass:        payload.ResolverAmbiguityClass,
		ResolverCandidateKind:         payload.ResolverCandidateKind,
		ResolverRetryDisposition:      payload.ResolverRetryDisposition,
		ResolverManualRetryHint:       payload.ResolverManualRetryHint,
		ResolverNextStepAlias:         payload.ResolverNextStepAlias,
		Routes:                        browserRuntimeRouteCoordinationInputs(payload.SessionRoutes),
		WorkbenchReady:                payload.WorkbenchReady,
		WorkbenchSections:             append([]string(nil), payload.WorkbenchSections...),
		WorkbenchPrimaryBrowserAction: payload.WorkbenchPrimaryBrowserAction,
		WorkbenchPrimaryNodeAction:    payload.WorkbenchPrimaryNodeAction,
		WorkbenchNextStep:             payload.WorkbenchNextStep,
	}
	browserPopulateActionabilityGuidanceProjectionRequest(&req, payload.Actionability, payload.FailureEvidence)
	return req
}

func browserRuntimeSyncSharedGuidanceProjection(payload *browserRuntimePayload, includeWorkbenchSurface bool) {
	if payload == nil {
		return
	}
	browserRuntimeApplyGuidanceShellProjection(
		payload,
		browserRuntimeProjectGuidanceShell(*payload, includeWorkbenchSurface),
	)
	browserRuntimeApplyDefaultCandidateRouteToPayloadShells(payload)
	browserRuntimeApplyRepairCommandToPayloadShells(payload)
}

func browserRuntimeMaybeSyncSharedGuidanceProjection(payload *browserRuntimePayload, includeWorkbenchSurface bool) bool {
	if payload == nil {
		return false
	}
	projection := browserRuntimeProjectGuidanceShell(*payload, includeWorkbenchSurface)
	if browserRuntimeGuidanceShellProjectionEmpty(projection) {
		return false
	}
	browserRuntimeApplyGuidanceShellProjection(payload, projection)
	browserRuntimeApplyDefaultCandidateRouteToPayloadShells(payload)
	browserRuntimeApplyRepairCommandToPayloadShells(payload)
	return true
}

func browserRuntimeProjectGuidanceShell(
	payload browserRuntimePayload,
	includeWorkbenchSurface bool,
) browserRuntimeGuidanceShellProjection {
	return browserRuntimeGuidanceShellProjectionFromShared(
		agentxbrowserruntime.BuildSharedSessionBrowserGuidanceProjection(
			browserRuntimeSharedGuidanceProjectionRequest(&payload, includeWorkbenchSurface),
		),
	)
}

func browserRuntimeGuidanceShellProjectionFromShared(
	projection agentxbrowserruntime.SharedSessionBrowserGuidanceProjection,
) browserRuntimeGuidanceShellProjection {
	return browserRuntimeGuidanceShellProjection{
		ResolverExplanation:    browserRuntimeResolverExplanationSummaryFromShared(projection.ResolverExplanation),
		DiagnosticsExplanation: browserRuntimeDiagnosticsExplanationSummaryFromShared(projection.DiagnosticsExplanation),
		WorkbenchExplanation:   browserRuntimeDiagnosticsExplanationSummaryFromShared(projection.WorkbenchExplanation),
		WorkbenchDiagnostics:   browserRuntimeWorkbenchDiagnosticsSummaryFromShared(projection.WorkbenchDiagnostics),
		WorkbenchSummary:       browserRuntimeTopLevelSummaryFromShared(projection.WorkbenchSummary),
		WorkbenchDisplay:       browserRuntimeWorkbenchDisplaySummaryFromShared(projection.WorkbenchDisplay),
		Explanation:            browserRuntimeTopLevelSummaryFromShared(projection.Explanation),
		Diagnostics:            browserRuntimeTopLevelSummaryFromShared(projection.Diagnostics),
		Summary:                browserRuntimeTopLevelSummaryFromShared(projection.Summary),
		Display:                browserRuntimeTopLevelDisplaySummaryFromShared(projection.Display),
	}
}

func browserRuntimeGuidanceShellProjectionEmpty(projection browserRuntimeGuidanceShellProjection) bool {
	return projection.ResolverExplanation == nil &&
		projection.DiagnosticsExplanation == nil &&
		projection.WorkbenchExplanation == nil &&
		projection.WorkbenchDiagnostics == nil &&
		projection.WorkbenchSummary == nil &&
		projection.WorkbenchDisplay == nil &&
		projection.Explanation == nil &&
		projection.Diagnostics == nil &&
		projection.Summary == nil &&
		projection.Display == nil
}

func browserRuntimeApplyGuidanceShellProjection(
	payload *browserRuntimePayload,
	projection browserRuntimeGuidanceShellProjection,
) {
	if payload == nil {
		return
	}
	payload.ResolverExplanation = projection.ResolverExplanation
	payload.DiagnosticsExplanation = projection.DiagnosticsExplanation
	payload.WorkbenchExplanation = projection.WorkbenchExplanation
	payload.WorkbenchDiagnostics = projection.WorkbenchDiagnostics
	payload.WorkbenchSummary = projection.WorkbenchSummary
	payload.WorkbenchDisplay = projection.WorkbenchDisplay
	payload.Explanation = projection.Explanation
	payload.Diagnostics = projection.Diagnostics
	payload.Summary = projection.Summary
	payload.Display = projection.Display
}

func browserRuntimeResolverExplanationSummaryFromShared(
	summary *agentxbrowserruntime.SharedSessionBrowserSummary,
) *browserRuntimeResolverExplanationSummary {
	if summary == nil {
		return nil
	}
	out := &browserRuntimeResolverExplanationSummary{
		State:           strings.TrimSpace(summary.State),
		SummaryCode:     strings.TrimSpace(summary.SummaryCode),
		NextStepAlias:   strings.TrimSpace(summary.NextStepAlias),
		ManualRetryHint: strings.TrimSpace(summary.ManualRetryHint),
	}
	if out.State == "" && out.SummaryCode == "" && out.NextStepAlias == "" && out.ManualRetryHint == "" {
		return nil
	}
	return out
}

func browserRuntimeDiagnosticsExplanationSummaryFromShared(
	summary *agentxbrowserruntime.SharedSessionBrowserSummary,
) *browserRuntimeDiagnosticsExplanationSummary {
	if summary == nil {
		return nil
	}
	out := &browserRuntimeDiagnosticsExplanationSummary{
		Category:        strings.TrimSpace(summary.Category),
		State:           strings.TrimSpace(summary.State),
		SummaryCode:     strings.TrimSpace(summary.SummaryCode),
		NextStepAlias:   strings.TrimSpace(summary.NextStepAlias),
		ManualRetryHint: strings.TrimSpace(summary.ManualRetryHint),
	}
	if out.Category == "" && out.State == "" && out.SummaryCode == "" && out.NextStepAlias == "" && out.ManualRetryHint == "" {
		return nil
	}
	return out
}

func browserDiagnosticsExplanationSummaryFromShared(
	summary *agentxbrowserruntime.SharedSessionBrowserSummary,
) *browserDiagnosticsExplanationSummary {
	if summary == nil {
		return nil
	}
	out := &browserDiagnosticsExplanationSummary{
		Category:        strings.TrimSpace(summary.Category),
		State:           strings.TrimSpace(summary.State),
		SummaryCode:     strings.TrimSpace(summary.SummaryCode),
		NextStepAlias:   strings.TrimSpace(summary.NextStepAlias),
		ManualRetryHint: strings.TrimSpace(summary.ManualRetryHint),
	}
	if out.Category == "" && out.State == "" && out.SummaryCode == "" && out.NextStepAlias == "" && out.ManualRetryHint == "" {
		return nil
	}
	return out
}

func browserRuntimeWorkbenchDiagnosticsSummaryFromShared(
	summary *agentxbrowserruntime.SharedSessionBrowserSummary,
) *browserRuntimeWorkbenchDiagnosticsSummary {
	if summary == nil {
		return nil
	}
	out := &browserRuntimeWorkbenchDiagnosticsSummary{
		Category:             strings.TrimSpace(summary.Category),
		State:                strings.TrimSpace(summary.State),
		SummaryCode:          strings.TrimSpace(summary.SummaryCode),
		NextStepAlias:        strings.TrimSpace(summary.NextStepAlias),
		ManualRetryHint:      strings.TrimSpace(summary.ManualRetryHint),
		ResolvedViaFallback:  summary.ResolvedViaFallback,
		PrimaryBrowserAction: strings.TrimSpace(summary.PrimaryBrowserAction),
		PrimaryNodeAction:    strings.TrimSpace(summary.PrimaryNodeAction),
		NextStep:             strings.TrimSpace(summary.NextStep),
	}
	if out.Category == "" &&
		out.State == "" &&
		out.SummaryCode == "" &&
		out.RepairCommand == "" &&
		out.NextStepAlias == "" &&
		out.ManualRetryHint == "" &&
		!out.ResolvedViaFallback &&
		out.PrimaryBrowserAction == "" &&
		out.PrimaryNodeAction == "" &&
		out.NextStep == "" {
		return nil
	}
	return out
}

func browserRuntimeTopLevelSummaryFromShared(
	summary *agentxbrowserruntime.SharedSessionBrowserSummary,
) *browserTopLevelSummary {
	if summary == nil {
		return nil
	}
	out := &browserTopLevelSummary{
		Category:             strings.TrimSpace(summary.Category),
		State:                strings.TrimSpace(summary.State),
		SummaryCode:          strings.TrimSpace(summary.SummaryCode),
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

func browserRuntimeWorkbenchDisplaySummaryFromShared(
	display *agentxbrowserruntime.SharedSessionBrowserDisplay,
) *browserRuntimeWorkbenchDisplaySummary {
	if display == nil {
		return nil
	}
	out := &browserRuntimeWorkbenchDisplaySummary{
		Ready:                display.Ready,
		Sections:             append([]string(nil), display.Sections...),
		Category:             strings.TrimSpace(display.Category),
		State:                strings.TrimSpace(display.State),
		SummaryCode:          strings.TrimSpace(display.SummaryCode),
		NextStepAlias:        strings.TrimSpace(display.NextStepAlias),
		ManualRetryHint:      strings.TrimSpace(display.ManualRetryHint),
		ResolvedViaFallback:  display.ResolvedViaFallback,
		PrimaryBrowserAction: strings.TrimSpace(display.PrimaryBrowserAction),
		PrimaryNodeAction:    strings.TrimSpace(display.PrimaryNodeAction),
		NextStep:             strings.TrimSpace(display.NextStep),
	}
	if browserUnifiedWorkbenchDisplayEmpty(*out) {
		return nil
	}
	return out
}

func browserRuntimeTopLevelDisplaySummaryFromShared(
	display *agentxbrowserruntime.SharedSessionBrowserDisplay,
) *browserTopLevelDisplaySummary {
	if display == nil {
		return nil
	}
	out := &browserTopLevelDisplaySummary{
		Ready:                display.Ready,
		Sections:             append([]string(nil), display.Sections...),
		Category:             strings.TrimSpace(display.Category),
		State:                strings.TrimSpace(display.State),
		SummaryCode:          strings.TrimSpace(display.SummaryCode),
		NextStepAlias:        strings.TrimSpace(display.NextStepAlias),
		ManualRetryHint:      strings.TrimSpace(display.ManualRetryHint),
		ResolvedViaFallback:  display.ResolvedViaFallback,
		PrimaryBrowserAction: strings.TrimSpace(display.PrimaryBrowserAction),
		PrimaryNodeAction:    strings.TrimSpace(display.PrimaryNodeAction),
		NextStep:             strings.TrimSpace(display.NextStep),
	}
	if browserTopLevelDisplayEmpty(*out) {
		return nil
	}
	return out
}
