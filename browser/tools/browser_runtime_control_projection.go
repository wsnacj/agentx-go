package tools

import (
	"strings"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

type browserRuntimeWorkbenchProjectionSync struct {
	ExtraSections           []string
	ClearActionPlan         bool
	SyncCoordinationSurface bool
}

type browserRuntimeWorkbenchShellProjection struct {
	WorkbenchExplanation *browserRuntimeDiagnosticsExplanationSummary
	WorkbenchDiagnostics *browserRuntimeWorkbenchDiagnosticsSummary
	WorkbenchSummary     *browserTopLevelSummary
	WorkbenchDisplay     *browserRuntimeWorkbenchDisplaySummary
	WorkbenchSurface     *browserRuntimeWorkbenchSurfaceSummary
}

type browserRuntimeProfileInventoryClear struct {
	ClearStatus         bool
	ClearProfiles       bool
	ClearDefaultProfile bool
}

func browserRuntimeClearTopLevelProfileInventory(
	payload *browserRuntimePayload,
	options browserRuntimeProfileInventoryClear,
) {
	if payload == nil {
		return
	}
	if options.ClearStatus {
		payload.ProfileStatus = nil
	}
	if options.ClearProfiles {
		payload.Profiles = nil
		payload.discoveredProfiles = nil
	}
	if options.ClearDefaultProfile {
		payload.DefaultProfile = ""
	}
}

func browserRuntimeSyncWorkbenchProjection(
	payload *browserRuntimePayload,
	options browserRuntimeWorkbenchProjectionSync,
) {
	if payload == nil {
		return
	}
	browserRuntimeSyncResolverGuidanceSummary(payload)
	browserRuntimeApplySharedWorkbenchProjection(
		payload,
		agentxbrowserruntime.BuildSharedSessionBrowserWorkbenchProjection(
			browserRuntimeSharedWorkbenchProjectionRequest(payload, options),
		),
	)
	browserRuntimeSyncSharedGuidanceProjection(payload, true)
	browserRuntimeApplyWorkbenchShellProjection(
		payload,
		browserRuntimeProjectWorkbenchShell(*payload),
	)
	browserRuntimeSyncReviewSurfaceSummary(payload)
}

func browserRuntimeProjectWorkbenchShell(payload browserRuntimePayload) browserRuntimeWorkbenchShellProjection {
	if !browserRuntimeHasWorkbenchSurface(payload) {
		return browserRuntimeWorkbenchShellProjection{}
	}
	explanation := payload.DiagnosticsExplanation
	if explanation == nil {
		explanation = browserRuntimeBuildDiagnosticsExplanationSummary(payload)
	}
	var workbenchExplanation *browserRuntimeDiagnosticsExplanationSummary
	if explanation != nil {
		copied := *explanation
		workbenchExplanation = &copied
	}

	var workbenchDiagnostics *browserRuntimeWorkbenchDiagnosticsSummary
	primaryBrowserAction := strings.TrimSpace(payload.WorkbenchPrimaryBrowserAction)
	primaryNodeAction := strings.TrimSpace(payload.WorkbenchPrimaryNodeAction)
	nextStep := strings.TrimSpace(payload.WorkbenchNextStep)
	if workbenchExplanation != nil || primaryBrowserAction != "" || primaryNodeAction != "" || nextStep != "" {
		workbenchDiagnostics = &browserRuntimeWorkbenchDiagnosticsSummary{
			PrimaryBrowserAction: primaryBrowserAction,
			PrimaryNodeAction:    primaryNodeAction,
			NextStep:             nextStep,
		}
		if workbenchExplanation != nil {
			workbenchDiagnostics.Category = strings.TrimSpace(workbenchExplanation.Category)
			workbenchDiagnostics.State = strings.TrimSpace(workbenchExplanation.State)
			workbenchDiagnostics.SummaryCode = strings.TrimSpace(workbenchExplanation.SummaryCode)
			workbenchDiagnostics.NextStepAlias = strings.TrimSpace(workbenchExplanation.NextStepAlias)
			workbenchDiagnostics.ManualRetryHint = strings.TrimSpace(workbenchExplanation.ManualRetryHint)
			if strings.EqualFold(workbenchDiagnostics.State, "resolved_via_fallback") {
				workbenchDiagnostics.ResolvedViaFallback = true
			}
		} else {
			workbenchDiagnostics.Category = "coordination"
			workbenchDiagnostics.State = "action_plan_available"
			workbenchDiagnostics.SummaryCode = "workbench_action_plan"
		}
	}

	workbenchPayload := payload
	workbenchPayload.WorkbenchExplanation = workbenchExplanation
	workbenchPayload.WorkbenchDiagnostics = workbenchDiagnostics
	workbenchSummary := browserRuntimeBuildWorkbenchSummary(workbenchPayload)
	workbenchPayload.WorkbenchSummary = workbenchSummary
	workbenchSurface := browserRuntimeWorkbenchSurfaceSummaryFromShared(
		agentxbrowserruntime.BuildSharedSessionBrowserWorkbenchSurface(
			browserRuntimeSharedWorkbenchSurfaceRequest(&workbenchPayload),
		),
	)
	workbenchDisplay := browserRuntimeBuildWorkbenchDisplaySummary(
		workbenchSurface,
		workbenchSummary,
		workbenchDiagnostics,
		workbenchExplanation,
		payload.WorkbenchReady,
		payload.WorkbenchSections,
	)

	return browserRuntimeWorkbenchShellProjection{
		WorkbenchExplanation: workbenchExplanation,
		WorkbenchDiagnostics: workbenchDiagnostics,
		WorkbenchSummary:     workbenchSummary,
		WorkbenchDisplay:     workbenchDisplay,
		WorkbenchSurface:     workbenchSurface,
	}
}

func browserRuntimeBuildWorkbenchDisplaySummary(
	surface *browserRuntimeWorkbenchSurfaceSummary,
	summary *browserTopLevelSummary,
	diagnostics *browserRuntimeWorkbenchDiagnosticsSummary,
	explanation *browserRuntimeDiagnosticsExplanationSummary,
	ready bool,
	sections []string,
) *browserRuntimeWorkbenchDisplaySummary {
	display := &browserRuntimeWorkbenchDisplaySummary{
		Ready:    ready,
		Sections: append([]string(nil), sections...),
	}
	if surface != nil {
		display.Ready = display.Ready || surface.Ready
		if len(display.Sections) == 0 && len(surface.Sections) > 0 {
			display.Sections = append([]string(nil), surface.Sections...)
		}
		display.RepairCommand = strings.TrimSpace(surface.RepairCommand)
		display.DefaultCandidateRoute = surface.DefaultCandidateRoute
	}
	source := summary
	if source == nil {
		source = browserRuntimeTopLevelSummaryFromWorkbenchDiagnostics(diagnostics)
	}
	if source == nil {
		source = browserRuntimeTopLevelSummaryFromDiagnosticsExplanation(explanation)
	}
	if source != nil {
		display.Category = strings.TrimSpace(source.Category)
		display.State = strings.TrimSpace(source.State)
		display.SummaryCode = strings.TrimSpace(source.SummaryCode)
		if display.RepairCommand == "" {
			display.RepairCommand = strings.TrimSpace(source.RepairCommand)
		}
		if display.DefaultCandidateRoute == (browserRuntimeRouteDescriptor{}) {
			display.DefaultCandidateRoute = source.DefaultCandidateRoute
		}
		display.NextStepAlias = strings.TrimSpace(source.NextStepAlias)
		display.ManualRetryHint = strings.TrimSpace(source.ManualRetryHint)
		display.ResolvedViaFallback = source.ResolvedViaFallback
		display.PrimaryBrowserAction = strings.TrimSpace(source.PrimaryBrowserAction)
		display.PrimaryNodeAction = strings.TrimSpace(source.PrimaryNodeAction)
		display.NextStep = strings.TrimSpace(source.NextStep)
	}
	if browserUnifiedWorkbenchDisplayEmpty(*display) {
		return nil
	}
	return display
}

func browserRuntimeApplyWorkbenchShellProjection(
	payload *browserRuntimePayload,
	projection browserRuntimeWorkbenchShellProjection,
) {
	if payload == nil {
		return
	}
	payload.WorkbenchExplanation = projection.WorkbenchExplanation
	payload.WorkbenchDiagnostics = projection.WorkbenchDiagnostics
	payload.WorkbenchSummary = projection.WorkbenchSummary
	payload.WorkbenchDisplay = projection.WorkbenchDisplay
	payload.Workbench = projection.WorkbenchSurface
	browserRuntimeApplyDefaultCandidateRouteToPayloadShells(payload)
	browserRuntimeApplyRepairCommandToPayloadShells(payload)
}

func browserRuntimeRefreshWorkbenchProjection(payload *browserRuntimePayload, extraSections ...string) {
	browserRuntimeSyncWorkbenchProjection(payload, browserRuntimeWorkbenchProjectionSync{
		ExtraSections: extraSections,
	})
}

func browserRuntimeWorkbenchHasSessionProjection(payload browserRuntimePayload) bool {
	return payload.SessionTargetCount > 0 ||
		len(payload.SessionRoutes) > 0 ||
		len(payload.SessionRuns) > 0 ||
		len(payload.SessionProfiles) > 0 ||
		payload.SessionHandoff != nil ||
		payload.SessionProfileSelection != nil ||
		payload.SessionTargetSelection != nil ||
		browserRuntimeBindingHasSessionSurface(payload.SessionBinding)
}

func browserRuntimeBindingHasSessionSurface(binding *browserRuntimeSessionBinding) bool {
	if binding == nil {
		return false
	}
	evaluation := browserRuntimeSharedBindingEvaluation(*binding, nil)
	snapshot := evaluation.Snapshot
	if strings.TrimSpace(snapshot.CurrentTargetID) != "" ||
		snapshot.SelectedProfileSelection != nil ||
		snapshot.SelectedTargetSelection != nil ||
		len(snapshot.Runs) > 0 ||
		len(snapshot.Profiles) > 0 ||
		evaluation.Handoff != nil {
		return true
	}
	summary := snapshot.Summary
	return summary.RouteTargetCount > 0 ||
		summary.PendingTargetReviewCount > 0 ||
		summary.BlockedAutoFollowRouteCount > 0 ||
		summary.PopupStormRouteCount > 0 ||
		summary.NodeRunCount > 0 ||
		strings.TrimSpace(summary.ActiveNodeRunID) != "" ||
		len(summary.NodeRunStatusCounts) > 0 ||
		summary.BrowserProfileCount > 0 ||
		strings.TrimSpace(summary.ActiveBrowserProfile) != "" ||
		len(summary.BrowserProfileStatusCounts) > 0
}

func browserRuntimeClearWorkbenchActionPlan(payload *browserRuntimePayload) {
	if payload == nil {
		return
	}
	payload.WorkbenchPrimaryBrowserAction = ""
	payload.WorkbenchPrimaryNodeAction = ""
	payload.WorkbenchNextStep = ""
	payload.WorkbenchRecommendedBrowserActions = nil
	payload.WorkbenchRecommendedNodeActions = nil
}

func browserRuntimeHideImplicitLegacyHostSessionSurface(payload *browserRuntimePayload) {
	if payload == nil {
		return
	}
	if payload.SessionProfileSelection != nil {
		target := strings.ToLower(strings.TrimSpace(payload.SessionProfileSelection.RuntimeTarget))
		if target == "" || target == "host" {
			payload.SessionProfileSelection = nil
		}
	}
	if payload.SessionTargetSelection != nil {
		target := strings.ToLower(strings.TrimSpace(payload.SessionTargetSelection.RuntimeTarget))
		if target == "" || target == "host" {
			payload.SessionTargetSelection = nil
		}
	}
	if payload.SessionBinding != nil {
		payload.SessionBinding.SelectedBrowserBackend = ""
		payload.SessionBinding.SelectedBrowserApp = ""
		payload.SessionBinding.SelectedBrowserProfile = ""
		payload.SessionBinding.SelectedBrowserProfileSource = ""
		payload.SessionBinding.CurrentTargetID = ""
		payload.SessionBinding.SelectedBrowserTargetID = ""
		payload.SessionBinding.SelectedBrowserTabIndex = 0
		payload.SessionBinding.SelectedBrowserTarget = ""
		payload.SessionBinding.SelectedBrowserTargetSource = ""
		payload.SessionBinding.Coordination = nil
		payload.SessionBinding.BrowserProfileCount = 0
		payload.SessionBinding.ActiveBrowserProfile = ""
		payload.SessionBinding.BrowserProfileStatusCounts = nil
		payload.SessionBinding.BrowserProfiles = nil
		payload.SessionBinding.SessionHealthState = ""
		payload.SessionBinding.SessionHealthReason = ""
		payload.SessionBinding.SessionHealthRecoveryAction = ""
		payload.SessionBinding.SessionHealthReconnectHint = ""
		payload.SessionBinding.SessionHealthDisconnectCount = 0
		payload.SessionBinding.SessionHealthDisconnectBurstCount = 0
		payload.SessionBinding.SessionHealthDisconnectBurstWindowMs = 0
		payload.SessionBinding.SessionHealthCooldownRemainingMs = 0
		payload.SessionBinding.SessionHealthRetryBackoffRemainingMs = 0
		payload.SessionBinding.SessionHealthRestartAttemptCount = 0
		payload.SessionBinding.SessionHealthRestartFailureCount = 0
		payload.SessionBinding.SessionHealthLastDisconnectUnixMilli = 0
		payload.SessionBinding.SessionHealthLastReconnectUnixMilli = 0
		payload.SessionBinding.SessionHealthLastRestartAttemptUnixMilli = 0
		payload.SessionBinding.SessionHealthLastRestartResult = ""
		payload.SessionBinding.SessionHealthLastRestartError = ""
		payload.SessionBinding.SessionHealthRecommendedBackoffMs = 0
		payload.SessionBinding.SessionHealthResolverBlockedBy = ""
		payload.SessionBinding.SessionHealthResolverAmbiguityClass = ""
		payload.SessionBinding.SessionHealthResolverCandidateKind = ""
		payload.SessionBinding.SessionHealthResolverStrength = ""
		payload.SessionBinding.SessionHealthResolverRetryDisposition = ""
		payload.SessionBinding.SessionHealthResolverManualRetryHint = ""
		payload.SessionBinding.SessionHealthResolverNextStepAlias = ""
		payload.SessionBinding.SessionHealthResolverSpecificityFields = nil
		payload.SessionBinding.SessionHandoff = nil
	}
	payload.SessionHandoff = nil
	browserRuntimeClearResolverGuidanceSummary(payload)
	browserRuntimeClearWorkbenchCoordinationSummary(payload)
}

func browserRuntimeHideInspectionProjection(payload *browserRuntimePayload, action string) {
	if payload == nil {
		return
	}
	switch browserRuntimeCanonicalAction(action) {
	case "doctor":
		browserRuntimeClearTopLevelProfileInventory(payload, browserRuntimeProfileInventoryClear{
			ClearStatus:         true,
			ClearDefaultProfile: true,
		})
	case "prepare":
		browserRuntimeClearTopLevelProfileInventory(payload, browserRuntimeProfileInventoryClear{
			ClearStatus:         true,
			ClearDefaultProfile: true,
		})
	case "status":
		browserRuntimeClearTopLevelProfileInventory(payload, browserRuntimeProfileInventoryClear{
			ClearStatus:         true,
			ClearDefaultProfile: true,
		})
	case "profiles":
		browserRuntimeClearTopLevelProfileInventory(payload, browserRuntimeProfileInventoryClear{
			ClearProfiles:       true,
			ClearDefaultProfile: true,
		})
	case "workbench":
		browserRuntimeClearTopLevelProfileInventory(payload, browserRuntimeProfileInventoryClear{
			ClearStatus:         true,
			ClearProfiles:       true,
			ClearDefaultProfile: true,
		})
		browserRuntimeSyncWorkbenchProjection(payload, browserRuntimeWorkbenchProjectionSync{
			ClearActionPlan: true,
		})
	}
}
