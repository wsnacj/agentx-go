package tools

import (
	"errors"
	"strings"
)

type browserRuntimeManagedLaunchFailureGuidance struct {
	NextStepAlias        string
	ManualRetryHint      string
	PrimaryBrowserAction string
	PrimaryNodeAction    string
	NextStep             string
}

func firstBrowserRuntimeRouteDescriptor(items ...browserRuntimeRouteDescriptor) browserRuntimeRouteDescriptor {
	for _, item := range items {
		if item != (browserRuntimeRouteDescriptor{}) {
			return item
		}
	}
	return browserRuntimeRouteDescriptor{}
}

func browserRuntimeRouteErrIsManagedLaunchFailure(err error) bool {
	if err == nil {
		return false
	}
	var managedErr *browserManagedRouteUnavailableError
	return errors.As(err, &managedErr)
}

func browserRuntimeUsesManagedLaunchFailureInspectionSummary(action string) bool {
	switch browserRuntimeCanonicalAction(action) {
	case "status", "profiles", "sessions", "workbench", "prepare", "coordinate", "start":
		return true
	default:
		return false
	}
}

func browserRuntimeManagedLaunchFailureSummaryBase(nextStepAlias string, manualRetryHint string) browserTopLevelSummary {
	return browserTopLevelSummary{
		Category:        "coordination",
		State:           "runtime_unavailable",
		SummaryCode:     "managed_default_launch_failed",
		NextStepAlias:   strings.TrimSpace(nextStepAlias),
		ManualRetryHint: firstNonEmpty(strings.TrimSpace(manualRetryHint), "retry_launch"),
	}
}

func browserRuntimeManagedLaunchFailureExplanationSummary(nextStepAlias string, manualRetryHint string) *browserRuntimeDiagnosticsExplanationSummary {
	base := browserRuntimeManagedLaunchFailureSummaryBase(nextStepAlias, manualRetryHint)
	return &browserRuntimeDiagnosticsExplanationSummary{
		Category:        base.Category,
		State:           base.State,
		SummaryCode:     base.SummaryCode,
		NextStepAlias:   base.NextStepAlias,
		ManualRetryHint: base.ManualRetryHint,
	}
}

func browserRuntimeManagedLaunchFailureTopLevelSummary(
	guidance browserRuntimeManagedLaunchFailureGuidance,
) *browserTopLevelSummary {
	base := browserRuntimeManagedLaunchFailureSummaryBase(guidance.NextStepAlias, guidance.ManualRetryHint)
	base.PrimaryBrowserAction = strings.TrimSpace(guidance.PrimaryBrowserAction)
	base.PrimaryNodeAction = strings.TrimSpace(guidance.PrimaryNodeAction)
	base.NextStep = strings.TrimSpace(guidance.NextStep)
	if browserUnifiedSummaryEmpty(base) {
		return nil
	}
	return &base
}

func browserRuntimeManagedLaunchFailurePrimaryTarget(preview browserRuntimeDiagnosticsPreview) string {
	for _, target := range []string{"node", "sandbox"} {
		if assessment, ok := browserRuntimePreviewManagedRouteAssessment(preview, target); ok &&
			browserRuntimeShouldKeepManagedLaunchFailureActionSurface(assessment) {
			return target
		}
	}
	if browserRuntimePreviewConfiguresManagedTarget(preview, "node") {
		return "node"
	}
	if browserRuntimePreviewConfiguresManagedTarget(preview, "sandbox") {
		return "sandbox"
	}
	return ""
}

func browserRuntimeManagedLaunchFailurePrimaryInfoForPreview(
	ctx browserRegistrationContext,
	preview browserRuntimeDiagnosticsPreview,
) BrowserRuntimeInfo {
	target := browserRuntimeManagedLaunchFailurePrimaryTarget(preview)
	if target == "" {
		return BrowserRuntimeInfo{}
	}
	fallback := defaultBrowserNodeRuntimeInfo()
	if target == "sandbox" {
		fallback = defaultBrowserSandboxRuntimeInfo()
	}
	return firstBrowserRuntimeInfo(
		browserRuntimePreviewFallbackInfoForManagedTarget(preview, target, fallback),
		browserRuntimeRouteFallbackInfoForPreviewTarget(ctx, preview, strings.TrimSpace(preview.DefaultRoute.Profile), target),
		BrowserRuntimeInfo{Target: target},
	)
}

func browserRuntimeManagedLaunchFailureWorkbenchDiagnosticsSummary(
	guidance browserRuntimeManagedLaunchFailureGuidance,
) *browserRuntimeWorkbenchDiagnosticsSummary {
	base := browserRuntimeManagedLaunchFailureSummaryBase(guidance.NextStepAlias, guidance.ManualRetryHint)
	return &browserRuntimeWorkbenchDiagnosticsSummary{
		Category:             base.Category,
		State:                base.State,
		SummaryCode:          base.SummaryCode,
		NextStepAlias:        base.NextStepAlias,
		ManualRetryHint:      base.ManualRetryHint,
		ResolvedViaFallback:  base.ResolvedViaFallback,
		PrimaryBrowserAction: strings.TrimSpace(guidance.PrimaryBrowserAction),
		PrimaryNodeAction:    strings.TrimSpace(guidance.PrimaryNodeAction),
		NextStep:             strings.TrimSpace(guidance.NextStep),
	}
}

func browserRuntimeManagedLaunchFailureGuidanceForHints(
	hints browserRuntimeActionHints,
	launchDiagnostics *browserRuntimeLaunchDiagnosticsSummary,
	fallbackPrimaryBrowserAction string,
	fallbackPrimaryNodeAction string,
	fallbackNextStep string,
) browserRuntimeManagedLaunchFailureGuidance {
	guidance := browserRuntimeManagedLaunchFailureGuidance{
		NextStepAlias:        strings.TrimSpace(hints.LaunchAlias),
		ManualRetryHint:      "retry_launch",
		PrimaryBrowserAction: strings.TrimSpace(firstNonEmpty(fallbackPrimaryBrowserAction, hints.LaunchCommand, "browser action=launch")),
		PrimaryNodeAction:    strings.TrimSpace(fallbackPrimaryNodeAction),
	}
	guidance.NextStep = strings.TrimSpace(firstNonEmpty(fallbackNextStep, guidance.PrimaryBrowserAction))
	if launchDiagnostics != nil &&
		browserRuntimeBootstrapErrorCodeSupportsRepair(launchDiagnostics.BootstrapErrorCode) &&
		strings.TrimSpace(hints.RepairCommand) != "" {
		guidance.NextStepAlias = firstNonEmpty(strings.TrimSpace(hints.RepairAlias), "repair")
		guidance.ManualRetryHint = "repair_bootstrap"
		guidance.PrimaryBrowserAction = strings.TrimSpace(hints.RepairCommand)
		guidance.NextStep = strings.TrimSpace(hints.RepairCommand)
	}
	return guidance
}

func browserRuntimeManagedLaunchFailureWorkbenchDisplaySummary(
	payload browserRuntimePayload,
	summary *browserTopLevelSummary,
) *browserRuntimeWorkbenchDisplaySummary {
	if summary == nil {
		return nil
	}
	display := &browserRuntimeWorkbenchDisplaySummary{
		Ready:                payload.WorkbenchReady,
		Sections:             append([]string(nil), payload.WorkbenchSections...),
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
	if browserUnifiedWorkbenchDisplayEmpty(*display) {
		return nil
	}
	return display
}

func browserRuntimeShouldApplyManagedLaunchFailureInspectionSummary(
	payload *browserRuntimePayload,
	action string,
	requestedTarget string,
	preview browserRuntimeDiagnosticsPreview,
) (string, bool) {
	hiddenImplicitHostDefaultPayload := payload != nil &&
		(payload.DefaultRoute == (browserRuntimeRouteDescriptor{}) ||
			strings.TrimSpace(payload.SubstrateSelectionStrategy) == BrowserSubstrateSelectionLegacyHostDefault)
	if payload == nil ||
		!hiddenImplicitHostDefaultPayload ||
		strings.TrimSpace(requestedTarget) != "" ||
		!browserRuntimeUsesManagedLaunchFailureInspectionSummary(action) {
		return "", false
	}
	note := strings.TrimSpace(browserRuntimeManagedLaunchFailureNoteForPreview(preview))
	if note == "" {
		return "", false
	}
	if payload.Status == "unsupported" {
		return note, true
	}
	if payload.Status != "ok" {
		return "", false
	}
	if payload.SelectedRoute != nil &&
		BrowserSubstratePosture(payload.SelectedRoute.Backend, payload.SelectedRoute.RuntimeTarget) != BrowserSubstrateLegacySystemHost {
		return "", false
	}
	return note, true
}

func browserRuntimeApplyManagedLaunchFailureRouteMatrixSummary(
	routes []browserRuntimeRouteStatus,
	metadata browserRuntimeCapabilityMetadata,
	defaultCandidateRoute browserRuntimeRouteDescriptor,
	note string,
) []browserRuntimeRouteStatus {
	if len(routes) == 0 {
		return routes
	}
	rows := append([]browserRuntimeRouteStatus(nil), routes...)
	defaultIdx := -1
	hostFallbackExists := false
	var hostDefaultRow browserRuntimeRouteStatus
	for idx, row := range rows {
		if BrowserSubstratePosture(row.Backend, row.RuntimeTarget) != BrowserSubstrateLegacySystemHost {
			continue
		}
		if row.Status == "default" && defaultIdx < 0 {
			defaultIdx = idx
			hostDefaultRow = row
			continue
		}
		hostFallbackExists = true
	}
	if defaultIdx < 0 {
		return rows
	}
	defaultRow := browserRuntimeRouteStatus{
		DefaultCandidateRoute: firstBrowserRuntimeRouteDescriptor(
			rows[defaultIdx].DefaultCandidateRoute,
			defaultCandidateRoute,
		),
		Status: "unsupported",
		Note:   strings.TrimSpace(note),
	}
	browserRuntimeApplyCapabilityMetadataToRouteStatus(&defaultRow, metadata)
	rows[defaultIdx] = defaultRow
	if !hostFallbackExists && strings.TrimSpace(hostDefaultRow.RuntimeTarget) == "host" {
		hostDefaultRow.Status = "available"
		rows = append(rows, hostDefaultRow)
	}
	return rows
}

func browserRuntimeApplyManagedLaunchFailureSubstrateMatrixSummary(
	matrix []browserRuntimeSubstrateStatus,
	metadata browserRuntimeCapabilityMetadata,
	managedDefault BrowserRuntimeInfo,
	defaultCandidateRoute browserRuntimeRouteDescriptor,
	note string,
) []browserRuntimeSubstrateStatus {
	if len(matrix) == 0 {
		return matrix
	}
	rows := append([]browserRuntimeSubstrateStatus(nil), matrix...)
	defaultIdx := -1
	hostFallbackExists := false
	var hostDefaultRow browserRuntimeSubstrateStatus
	for idx, row := range rows {
		if row.Role == "host" {
			hostFallbackExists = true
		}
		if row.Role == "default" && defaultIdx < 0 {
			defaultIdx = idx
			hostDefaultRow = row
		}
	}
	if defaultIdx < 0 {
		return rows
	}
	defaultRow := browserRuntimeSubstrateStatus{
		Role:           "default",
		SelectionState: "unsupported",
		SelectionReason: firstNonEmpty(
			strings.TrimSpace(hostDefaultRow.SelectionReason),
			strings.TrimSpace(note),
		),
		DefaultCandidateRoute: firstBrowserRuntimeRouteDescriptor(
			rows[defaultIdx].DefaultCandidateRoute,
			defaultCandidateRoute,
		),
		Status: "unsupported",
		Note:   strings.TrimSpace(note),
	}
	browserRuntimeApplyCapabilityMetadataToSubstrateStatus(&defaultRow, metadata)
	rows[defaultIdx] = defaultRow
	if !hostFallbackExists && strings.TrimSpace(hostDefaultRow.RuntimeTarget) == "host" {
		hostInfo := BrowserRuntimeInfo{
			Backend: hostDefaultRow.Backend,
			Profile: hostDefaultRow.Profile,
			Target:  hostDefaultRow.RuntimeTarget,
		}
		hostRow := hostDefaultRow
		hostRow.Role = "host"
		hostRow.SelectionState = "explicit_fallback"
		hostRow.SelectionReason = firstNonEmpty(
			browserRuntimeSubstrateSelectionReason("host", hostInfo, managedDefault),
			hostDefaultRow.SelectionReason,
		)
		if strings.TrimSpace(hostRow.Status) == "" || hostRow.Status == "default" {
			hostRow.Status = "available"
		}
		rows = append(rows, hostRow)
	}
	return rows
}

func browserRuntimeMaybeApplyManagedLaunchFailureMatrixSummary(
	ctx browserRegistrationContext,
	payload *browserRuntimePayload,
	preview browserRuntimeDiagnosticsPreview,
	note string,
) {
	if payload == nil || (len(payload.Routes) == 0 && len(payload.SubstrateMatrix) == 0) {
		return
	}
	managedDefault := browserRuntimeManagedLaunchFailurePrimaryInfoForPreview(ctx, preview)
	metadata, ok := browserRuntimeManagedLaunchFailureProjectedSurfaceMetadataWithPreview(ctx, preview, managedDefault)
	if !ok {
		return
	}
	managedCandidateRoute := browserRuntimeRouteDescriptor{}
	if projection, ok := browserRuntimeDoctorManagedCandidateRouteProjection(preview); ok {
		managedCandidateRoute = browserRuntimeRouteDescriptorFromInfoWithProvenance(
			projection.Info,
			projection.Metadata.Source,
			projection.Metadata.Endpoint,
		)
	}
	defaultCandidateRoute := firstBrowserRuntimeRouteDescriptor(
		browserRuntimeDefaultCandidateRouteDescriptor(preview),
		managedCandidateRoute,
		browserRuntimeRouteDescriptorFromInfo(managedDefault),
	)
	payload.Routes = browserRuntimeApplyManagedLaunchFailureRouteMatrixSummary(
		payload.Routes,
		metadata,
		defaultCandidateRoute,
		note,
	)
	payload.SubstrateMatrix = browserRuntimeApplyManagedLaunchFailureSubstrateMatrixSummary(
		payload.SubstrateMatrix,
		metadata,
		managedDefault,
		defaultCandidateRoute,
		note,
	)
}

func browserRuntimeMaybeApplyManagedLaunchFailureInspectionSummary(
	ctx browserRegistrationContext,
	payload *browserRuntimePayload,
	action string,
	requestedTarget string,
	preview browserRuntimeDiagnosticsPreview,
) {
	note, ok := browserRuntimeShouldApplyManagedLaunchFailureInspectionSummary(
		payload,
		action,
		requestedTarget,
		preview,
	)
	if !ok {
		return
	}
	rawNote := strings.TrimSpace(note)
	launchDiagnostics := browserRuntimeLaunchDiagnosticsSummaryFromManagedLaunchFailure(ctx, preview, rawNote)
	note = browserRuntimeManagedLaunchFailureSurfaceNoteWithCandidate(payload, launchDiagnostics, rawNote)
	hasSpecificSurfaceNote := strings.TrimSpace(note) != "" && strings.TrimSpace(note) != rawNote
	payload.Status = "unsupported"
	if payloadNote := strings.TrimSpace(payload.Note); payloadNote == "" ||
		hasSpecificSurfaceNote ||
		strings.HasPrefix(payloadNote, "browser_runtime: no tool session context is available") ||
		strings.Contains(payloadNote, "selected route does not support action") ||
		!strings.Contains(strings.ToLower(payloadNote), "managed_route_unavailable") {
		payload.Note = note
	}
	if metadata, ok := browserRuntimeManagedLaunchFailureSurfaceMetadataWithPreview(
		ctx,
		preview,
		BrowserRuntimeInfo{Target: "node"},
	); ok {
		browserRuntimeApplyCapabilityMetadataToPayload(payload, metadata)
	}

	hints := browserRuntimeActionHintsForRegistration(ctx)
	guidance := browserRuntimeManagedLaunchFailureGuidanceForHints(
		hints,
		launchDiagnostics,
		payload.WorkbenchPrimaryBrowserAction,
		payload.WorkbenchPrimaryNodeAction,
		payload.WorkbenchNextStep,
	)
	explanation := browserRuntimeManagedLaunchFailureExplanationSummary(guidance.NextStepAlias, guidance.ManualRetryHint)
	payload.DiagnosticsExplanation = explanation
	payload.Explanation = &browserTopLevelSummary{
		Category:        explanation.Category,
		State:           explanation.State,
		SummaryCode:     explanation.SummaryCode,
		NextStepAlias:   explanation.NextStepAlias,
		ManualRetryHint: explanation.ManualRetryHint,
	}

	action = browserRuntimeCanonicalAction(action)
	topLevelSummary := browserRuntimeManagedLaunchFailureTopLevelSummary(guidance)
	payload.Diagnostics = browserCloneTopLevelSummary(topLevelSummary)
	payload.Summary = browserCloneTopLevelSummary(topLevelSummary)
	browserRuntimeApplyLaunchDiagnostics(
		payload,
		action,
		launchDiagnostics,
	)

	if action == "workbench" {
		payload.WorkbenchSections = mergeToolMetadataStrings(payload.WorkbenchSections, []string{"coordination"})
		payload.WorkbenchReady = len(payload.WorkbenchSections) > 0
		payload.WorkbenchPrimaryBrowserAction = guidance.PrimaryBrowserAction
		payload.WorkbenchPrimaryNodeAction = guidance.PrimaryNodeAction
		payload.WorkbenchNextStep = guidance.NextStep
		payload.WorkbenchRecommendedBrowserActions = mergeToolMetadataStrings(
			payload.WorkbenchRecommendedBrowserActions,
			[]string{guidance.PrimaryBrowserAction},
		)
		payload.WorkbenchExplanation = explanation
		payload.WorkbenchDiagnostics = browserRuntimeManagedLaunchFailureWorkbenchDiagnosticsSummary(guidance)
		payload.WorkbenchSummary = browserCloneTopLevelSummary(topLevelSummary)
		payload.WorkbenchDisplay = browserRuntimeManagedLaunchFailureWorkbenchDisplaySummary(*payload, topLevelSummary)
		payload.Display = browserTopLevelDisplayFromWorkbenchDisplay(payload.WorkbenchDisplay)
		browserRuntimeSyncWorkbenchSurfaceSummary(payload)
	} else {
		payload.Display = browserTopLevelDisplayFromSummary(topLevelSummary)
	}

	browserRuntimeSyncTopLevelSurfaceSummary(payload)
	browserRuntimeMaybeApplyManagedLaunchFailureMatrixSummary(ctx, payload, preview, note)
}
