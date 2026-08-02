package tools

import "strings"

func browserRuntimeUsesDoctorRouteInspectionSummary(action string) bool {
	switch browserRuntimeCanonicalAction(action) {
	case "doctor", "status", "profiles", "sessions", "workbench", "prepare", "coordinate", "start":
		return true
	default:
		return false
	}
}

func browserRuntimeDoctorRouteInspectionMetadata(
	ctx browserRegistrationContext,
	preview browserRuntimeDiagnosticsPreview,
	info BrowserRuntimeInfo,
) browserRuntimeCapabilityMetadata {
	metadata := browserRuntimeDiagnosticsSurfaceMetadataForRouteWithPreview(ctx, preview, info)
	if strings.TrimSpace(metadata.BrowserSurface) != "" || len(metadata.BrowserOptInTargets) > 0 {
		return metadata
	}
	backend := browserRuntimePreviewManagedTargetBackend(preview, info.Target)
	if backend == nil {
		return metadata
	}
	capabilities := browserCapabilitiesForConcreteBackend(backend)
	if !capabilities.SupportsAnyActKind() {
		return metadata
	}
	optInSurface := browserRuntimeManagedOptInSurfaceForResolvedRouteWithPreview(ctx, preview, info, capabilities)
	if browserRuntimeManagedOptInSurfaceLabel(optInSurface) != "" {
		return browserRuntimeCapabilityMetadataForDiagnosticsSurface(
			ctx,
			browserRuntimeDiagnosticsCapabilities(ctx),
			optInSurface,
		)
	}
	if info.Target != "node" && info.Target != "sandbox" {
		return metadata
	}
	optInSurface = browserRuntimeManagedOptInDiagnosticsSurface{
		Targets:      []string{info.Target},
		Capabilities: capabilities,
	}
	return browserRuntimeCapabilityMetadataForDiagnosticsSurface(
		ctx,
		browserRuntimeDiagnosticsCapabilities(ctx),
		optInSurface,
	)
}

func browserRuntimeHasDoctorRouteInspectionSummary(summary *browserTopLevelSummary) bool {
	if summary == nil {
		return false
	}
	switch strings.TrimSpace(summary.SummaryCode) {
	case "managed_route_not_default", "managed_route_hidden_by_legacy_host_default":
		return strings.TrimSpace(summary.Category) == "coordination" &&
			strings.TrimSpace(summary.State) == "managed_route_pending_default"
	default:
		return false
	}
}

func browserRuntimeHasDoctorRouteInspectionDisplay(display *browserTopLevelDisplaySummary) bool {
	if display == nil {
		return false
	}
	return browserRuntimeHasDoctorRouteInspectionSummary(browserTopLevelSummaryFromDisplay(display))
}

func browserRuntimeHasDoctorRouteInspectionWorkbenchSummary(summary *browserRuntimeWorkbenchDiagnosticsSummary) bool {
	if summary == nil {
		return false
	}
	return strings.TrimSpace(summary.Category) == "coordination" &&
		strings.TrimSpace(summary.State) == "managed_route_pending_default" &&
		(strings.TrimSpace(summary.SummaryCode) == "managed_route_not_default" ||
			strings.TrimSpace(summary.SummaryCode) == "managed_route_hidden_by_legacy_host_default")
}

func browserRuntimeHasDoctorRouteInspectionWorkbenchDisplay(display *browserRuntimeWorkbenchDisplaySummary) bool {
	if display == nil {
		return false
	}
	return browserRuntimeHasDoctorRouteInspectionSummary(
		browserTopLevelSummaryFromDisplay(browserTopLevelDisplayFromWorkbenchDisplay(display)),
	)
}

func browserRuntimeHasConcreteLaunchFailureNote(note string) bool {
	lower := strings.ToLower(strings.TrimSpace(note))
	switch {
	case lower == "":
		return false
	case strings.Contains(lower, "managed browserd boot failed"):
		return true
	case strings.Contains(lower, "bundled browserd bootstrap"):
		return true
	default:
		return false
	}
}

func browserRuntimeLaunchDiagnosticsHasConcreteFailure(summary *browserRuntimeLaunchDiagnosticsSummary) bool {
	if summary == nil {
		return false
	}
	return browserRuntimeHasConcreteLaunchFailureNote(summary.Note) ||
		strings.TrimSpace(summary.RepairCommand) != "" ||
		strings.TrimSpace(summary.BootstrapState) != "" ||
		strings.TrimSpace(summary.BootstrapErrorCode) != ""
}

func browserDoctorLaunchSummaryHasConcreteFailure(summary *BrowserDoctorLaunchSummary) bool {
	if summary == nil {
		return false
	}
	return browserRuntimeHasConcreteLaunchFailureNote(summary.Note) ||
		strings.TrimSpace(summary.BootstrapState) != "" ||
		strings.TrimSpace(summary.BootstrapErrorCode) != ""
}

func browserRuntimePayloadHasConcreteManagedLaunchFailureEvidence(payload *browserRuntimePayload) bool {
	if payload == nil {
		return false
	}
	if browserRuntimeHasConcreteLaunchFailureNote(payload.Note) ||
		browserRuntimeLaunchDiagnosticsHasConcreteFailure(payload.LaunchDiagnostics) ||
		browserRuntimeLaunchDiagnosticsHasConcreteFailure(payload.WorkbenchLaunchDiagnostics) {
		return true
	}
	if payload.Doctor != nil {
		if browserDoctorLaunchSummaryHasConcreteFailure(payload.Doctor.Launch) {
			return true
		}
	}
	return false
}

func browserRuntimeDoctorRouteInspectionSummaryBase(route *BrowserDoctorRouteSummary, nextStepAlias string) (browserTopLevelSummary, bool) {
	if route == nil {
		return browserTopLevelSummary{}, false
	}
	switch strings.TrimSpace(route.Code) {
	case "managed_route_not_default", "managed_route_hidden_by_legacy_host_default":
		return browserTopLevelSummary{
			Category:            "coordination",
			State:               "managed_route_pending_default",
			SummaryCode:         strings.TrimSpace(route.Code),
			NextStepAlias:       firstNonEmpty(strings.TrimSpace(nextStepAlias), "ready"),
			ManualRetryHint:     "promote_managed_default",
			ResolvedViaFallback: true,
		}, true
	default:
		return browserTopLevelSummary{}, false
	}
}

func browserRuntimeDoctorRouteInspectionCanonicalSummary(route *BrowserDoctorRouteSummary) string {
	if route == nil {
		return ""
	}
	switch strings.TrimSpace(route.Code) {
	case "managed_route_not_default":
		return "Managed browser route is configured, but browser workbench still resolves onto the legacy host path."
	case "managed_route_hidden_by_legacy_host_default":
		return "Managed browser route is configured, but the default route is still hidden behind the implicit legacy host fallback."
	default:
		return strings.TrimSpace(route.Summary)
	}
}

func browserRuntimeDoctorRouteInspectionNote(route *BrowserDoctorRouteSummary, current string) string {
	routeNote := strings.TrimSpace("")
	if route != nil {
		routeNote = strings.TrimSpace(route.Summary)
	}
	current = strings.TrimSpace(current)
	switch {
	case routeNote == "":
		return current
	case current == "":
		return routeNote
	case strings.Contains(current, routeNote):
		return current
	case strings.Contains(routeNote, current):
		return routeNote
	default:
		return routeNote + " Details: " + current
	}
}

func browserRuntimeDoctorRouteInspectionExplanationSummary(route *BrowserDoctorRouteSummary) *browserRuntimeDiagnosticsExplanationSummary {
	base, ok := browserRuntimeDoctorRouteInspectionSummaryBase(route, "ready")
	if !ok {
		return nil
	}
	return &browserRuntimeDiagnosticsExplanationSummary{
		Category:        strings.TrimSpace(base.Category),
		State:           strings.TrimSpace(base.State),
		SummaryCode:     strings.TrimSpace(base.SummaryCode),
		NextStepAlias:   strings.TrimSpace(base.NextStepAlias),
		ManualRetryHint: strings.TrimSpace(base.ManualRetryHint),
	}
}

func browserRuntimeDoctorRouteInspectionTopLevelSummary(
	route *BrowserDoctorRouteSummary,
	nextStepAlias string,
	primaryBrowserAction string,
	nextStep string,
) *browserTopLevelSummary {
	base, ok := browserRuntimeDoctorRouteInspectionSummaryBase(route, nextStepAlias)
	if !ok {
		return nil
	}
	base.PrimaryBrowserAction = strings.TrimSpace(primaryBrowserAction)
	base.NextStep = strings.TrimSpace(nextStep)
	if browserUnifiedSummaryEmpty(base) {
		return nil
	}
	return &base
}

func browserRuntimeDoctorRouteInspectionDefaultRouteDescriptor(route *BrowserDoctorRouteSummary) browserRuntimeRouteDescriptor {
	if route == nil {
		return browserRuntimeRouteDescriptor{}
	}
	return browserRuntimeRouteDescriptorFromInfoWithProvenance(
		BrowserRuntimeInfo{
			Backend: strings.TrimSpace(route.Backend),
			Profile: strings.TrimSpace(route.Profile),
			Target:  strings.TrimSpace(route.RuntimeTarget),
		},
		route.Source,
		route.Endpoint,
	)
}

func browserRuntimeDoctorRouteInspectionSelectionStrategy(
	preview browserRuntimeDiagnosticsPreview,
	managedDefault BrowserRuntimeInfo,
) string {
	managedDefault = normalizeBrowserRuntimeInfo(managedDefault)
	if managedDefault == (BrowserRuntimeInfo{}) {
		return ""
	}
	hostInfo := normalizeBrowserRuntimeInfo(preview.Registration.SubstrateSummary.HostRoute)
	if hostInfo == (BrowserRuntimeInfo{}) {
		hostInfo = normalizeBrowserRuntimeInfo(preview.Registration.SubstrateAssessment.HostRuntime)
	}
	return BrowserSubstrateSelectionStrategy(managedDefault, hostInfo)
}

func browserRuntimeApplyDoctorRouteInspectionSubstrateMatrixSummary(
	matrix []browserRuntimeSubstrateStatus,
	metadata browserRuntimeCapabilityMetadata,
	managedDefault BrowserRuntimeInfo,
	note string,
) []browserRuntimeSubstrateStatus {
	if len(matrix) == 0 {
		return matrix
	}
	managedDefault = normalizeBrowserRuntimeInfo(managedDefault)
	if managedDefault == (BrowserRuntimeInfo{}) {
		return matrix
	}
	rows := append([]browserRuntimeSubstrateStatus(nil), matrix...)
	defaultIdx := -1
	for idx, row := range rows {
		if row.Role == "default" {
			defaultIdx = idx
			break
		}
	}
	if defaultIdx < 0 {
		return rows
	}
	defaultRow := rows[defaultIdx]
	defaultRow.Role = "default"
	defaultRow.SelectionState = "default"
	defaultRow.SelectionReason = firstNonEmpty(strings.TrimSpace(note), strings.TrimSpace(defaultRow.SelectionReason))
	defaultRow.Profile = managedDefault.Profile
	defaultRow.RuntimeTarget = managedDefault.Target
	defaultRow.Backend = managedDefault.Backend
	defaultRow.Status = "unsupported"
	defaultRow.SubstratePosture = BrowserSubstratePosture(managedDefault.Backend, managedDefault.Target)
	defaultRow.SubstrateStatus = BrowserSubstrateStatus(managedDefault.Backend, managedDefault.Target)
	defaultRow.SubstrateReason = BrowserSubstrateReason(managedDefault.Backend, managedDefault.Target)
	defaultRow.Note = firstNonEmpty(strings.TrimSpace(note), strings.TrimSpace(defaultRow.Note))
	browserRuntimeApplyCapabilityMetadataToSubstrateStatus(&defaultRow, metadata)
	rows[defaultIdx] = defaultRow
	return rows
}

func browserRuntimeDoctorRouteWorkbenchDiagnosticsSummary(
	route *BrowserDoctorRouteSummary,
	nextStepAlias string,
	primaryBrowserAction string,
	nextStep string,
) *browserRuntimeWorkbenchDiagnosticsSummary {
	base, ok := browserRuntimeDoctorRouteInspectionSummaryBase(route, nextStepAlias)
	if !ok {
		return nil
	}
	return &browserRuntimeWorkbenchDiagnosticsSummary{
		Category:             strings.TrimSpace(base.Category),
		State:                strings.TrimSpace(base.State),
		SummaryCode:          strings.TrimSpace(base.SummaryCode),
		NextStepAlias:        strings.TrimSpace(base.NextStepAlias),
		ManualRetryHint:      strings.TrimSpace(base.ManualRetryHint),
		ResolvedViaFallback:  base.ResolvedViaFallback,
		PrimaryBrowserAction: strings.TrimSpace(primaryBrowserAction),
		NextStep:             strings.TrimSpace(nextStep),
	}
}

func browserRuntimeDoctorRouteWorkbenchDisplaySummary(
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

func browserRuntimeShouldApplyDoctorRouteInspectionSummary(
	payload *browserRuntimePayload,
	action string,
	requestedTarget string,
) (*BrowserDoctorRouteSummary, bool) {
	if payload == nil ||
		payload.Doctor == nil ||
		payload.Doctor.Route == nil ||
		strings.TrimSpace(requestedTarget) != "" ||
		strings.TrimSpace(payload.RequestedProfile) != "" ||
		!browserRuntimeUsesDoctorRouteInspectionSummary(action) {
		return nil, false
	}
	if browserRuntimeHasDoctorRouteInspectionSummary(payload.Diagnostics) ||
		browserRuntimeHasDoctorRouteInspectionSummary(payload.Summary) ||
		browserRuntimeHasDoctorRouteInspectionDisplay(payload.Display) {
		return nil, false
	}
	switch browserRuntimeCanonicalAction(action) {
	case "workbench":
		if browserRuntimeHasDoctorRouteInspectionWorkbenchSummary(payload.WorkbenchDiagnostics) ||
			browserRuntimeHasDoctorRouteInspectionSummary(payload.WorkbenchSummary) ||
			browserRuntimeHasDoctorRouteInspectionWorkbenchDisplay(payload.WorkbenchDisplay) {
			return nil, false
		}
	}
	route := payload.Doctor.Route
	_, ok := browserRuntimeDoctorRouteInspectionSummaryBase(route, "ready")
	if !ok {
		return nil, false
	}
	return route, true
}

func browserRuntimeShouldPromoteDoctorRouteInspectionSummary(
	ctx browserRegistrationContext,
	payload *browserRuntimePayload,
	action string,
	preview browserRuntimeDiagnosticsPreview,
	route *BrowserDoctorRouteSummary,
) bool {
	if payload == nil || route == nil {
		return false
	}
	action = browserRuntimeCanonicalAction(action)
	managedInfo := normalizeBrowserRuntimeInfo(BrowserRuntimeInfo{
		Backend: route.Backend,
		Profile: route.Profile,
		Target:  route.RuntimeTarget,
	})
	if strings.TrimSpace(route.Code) != "managed_route_hidden_by_legacy_host_default" {
		return false
	}
	if managedInfo.Target != "node" {
		return false
	}
	status := strings.TrimSpace(payload.Status)
	hasConcreteLaunchFailure := browserRuntimePayloadHasConcreteManagedLaunchFailureEvidence(payload)
	switch action {
	case "doctor", "status", "workbench":
		if status == "ok" {
			break
		}
		if status != "unsupported" || hasConcreteLaunchFailure {
			return false
		}
	case "prepare":
		if status == "unsupported" && hasConcreteLaunchFailure {
			return false
		}
		if status != "ok" && status != "unsupported" {
			return false
		}
	case "coordinate", "start":
		if status != "unsupported" || hasConcreteLaunchFailure {
			return false
		}
	case "profiles":
		if status == "unsupported" && hasConcreteLaunchFailure {
			return false
		}
		if status != "ok" && status != "unsupported" {
			return false
		}
	case "sessions":
		if status == "ok" {
			break
		}
		if status != "unsupported" || hasConcreteLaunchFailure {
			return false
		}
	default:
		return false
	}
	if action == "profiles" || action == "sessions" {
		if strings.TrimSpace(route.Code) != "managed_route_hidden_by_legacy_host_default" {
			return false
		}
		if browserCompatRegisteredOrEnabledToolForActKind(ctx, "open") == "" &&
			!browserRuntimeRegisteredOrEnabledTool(ctx, "browser") {
			return false
		}
		if strings.TrimSpace(browserRuntimeDoctorRouteInspectionMetadata(ctx, preview, managedInfo).BrowserSurface) == "" {
			return false
		}
	}
	return true
}

func browserRuntimeDoctorRouteInspectionPayloadStatus(action string, current string) string {
	switch browserRuntimeCanonicalAction(action) {
	case "doctor", "status", "workbench", "sessions":
		return "ok"
	case "prepare", "coordinate", "start", "profiles":
		return "unsupported"
	default:
		return strings.TrimSpace(current)
	}
}

func browserRuntimeMaybeApplyDoctorRouteInspectionSummary(
	ctx browserRegistrationContext,
	payload *browserRuntimePayload,
	action string,
	requestedTarget string,
	preview browserRuntimeDiagnosticsPreview,
) {
	route, ok := browserRuntimeShouldApplyDoctorRouteInspectionSummary(payload, action, requestedTarget)
	if !ok {
		return
	}
	if !browserRuntimeShouldPromoteDoctorRouteInspectionSummary(ctx, payload, action, preview, route) {
		return
	}
	browserRuntimeApplyDoctorRouteInspectionSummaryPayload(
		ctx,
		payload,
		action,
		preview,
		route,
		browserRuntimeActionHintsForRegistration(ctx),
		browserRuntimeCanonicalAction(action) == "workbench",
	)
}

func browserRuntimeApplyDoctorRouteInspectionSummaryPayload(
	ctx browserRegistrationContext,
	payload *browserRuntimePayload,
	action string,
	preview browserRuntimeDiagnosticsPreview,
	route *BrowserDoctorRouteSummary,
	hints browserRuntimeActionHints,
	includeWorkbench bool,
) {
	if payload == nil || route == nil {
		return
	}
	payload.Status = browserRuntimeDoctorRouteInspectionPayloadStatus(action, payload.Status)
	managedInfo := normalizeBrowserRuntimeInfo(BrowserRuntimeInfo{
		Backend: route.Backend,
		Profile: route.Profile,
		Target:  route.RuntimeTarget,
	})
	if managedInfo != (BrowserRuntimeInfo{}) {
		browserRuntimeApplyCapabilityMetadataToPayload(
			payload,
			browserRuntimeDoctorRouteInspectionMetadata(ctx, preview, managedInfo),
		)
	}
	payload.Note = browserRuntimeDoctorRouteInspectionNote(route, payload.Note)
	payload.SubstrateSelectionReason = firstNonEmpty(
		strings.TrimSpace(route.Summary),
		strings.TrimSpace(payload.SubstrateSelectionReason),
	)
	explanation := browserRuntimeDoctorRouteInspectionExplanationSummary(route)
	if explanation != nil {
		explanation.NextStepAlias = firstNonEmpty(strings.TrimSpace(hints.ReadyAlias), explanation.NextStepAlias)
	}
	if payload.DefaultRoute == (browserRuntimeRouteDescriptor{}) {
		payload.DefaultRoute = browserRuntimeDoctorRouteInspectionDefaultRouteDescriptor(route)
	}
	if payload.DefaultCandidateRoute == (browserRuntimeRouteDescriptor{}) {
		payload.DefaultCandidateRoute = browserRuntimeDoctorRouteInspectionDefaultRouteDescriptor(route)
	}
	managedInfo = browserRuntimeInfoFromRouteDescriptor(&payload.DefaultRoute)
	if managedInfo == (BrowserRuntimeInfo{}) {
		managedInfo = normalizeBrowserRuntimeInfo(BrowserRuntimeInfo{
			Backend: route.Backend,
			Profile: route.Profile,
			Target:  route.RuntimeTarget,
		})
	}
	payload.SubstrateSelectionStrategy = firstNonEmpty(
		strings.TrimSpace(browserRuntimeDoctorRouteInspectionSelectionStrategy(preview, managedInfo)),
		strings.TrimSpace(payload.SubstrateSelectionStrategy),
	)
	payload.SubstrateMatrix = browserRuntimeApplyDoctorRouteInspectionSubstrateMatrixSummary(
		payload.SubstrateMatrix,
		browserRuntimeDoctorRouteInspectionMetadata(ctx, preview, managedInfo),
		managedInfo,
		strings.TrimSpace(route.Summary),
	)
	payload.DiagnosticsExplanation = explanation
	payload.Explanation = browserTopLevelSummaryFromRuntimeDiagnosticsExplanation(explanation)
	primaryBrowserAction := strings.TrimSpace(firstNonEmpty(hints.ReadyCommand, "browser action=ready"))
	topLevelSummary := browserRuntimeDoctorRouteInspectionTopLevelSummary(
		route,
		firstNonEmpty(strings.TrimSpace(hints.ReadyAlias), "ready"),
		primaryBrowserAction,
		primaryBrowserAction,
	)
	payload.Diagnostics = browserCloneTopLevelSummary(topLevelSummary)
	payload.Summary = browserCloneTopLevelSummary(topLevelSummary)

	if includeWorkbench {
		payload.WorkbenchSections = mergeToolMetadataStrings(payload.WorkbenchSections, []string{"route"})
		payload.WorkbenchReady = len(payload.WorkbenchSections) > 0
		payload.WorkbenchPrimaryBrowserAction = primaryBrowserAction
		payload.WorkbenchNextStep = primaryBrowserAction
		payload.WorkbenchRecommendedBrowserActions = mergeToolMetadataStrings(
			payload.WorkbenchRecommendedBrowserActions,
			[]string{primaryBrowserAction},
		)
		payload.WorkbenchExplanation = explanation
		payload.WorkbenchDiagnostics = browserRuntimeDoctorRouteWorkbenchDiagnosticsSummary(
			route,
			firstNonEmpty(strings.TrimSpace(hints.ReadyAlias), "ready"),
			primaryBrowserAction,
			primaryBrowserAction,
		)
		payload.WorkbenchSummary = browserCloneTopLevelSummary(topLevelSummary)
		payload.WorkbenchDisplay = browserRuntimeDoctorRouteWorkbenchDisplaySummary(*payload, topLevelSummary)
		payload.Display = browserTopLevelDisplayFromWorkbenchDisplay(payload.WorkbenchDisplay)
		browserRuntimeSyncWorkbenchSurfaceSummary(payload)
	} else {
		payload.Display = browserTopLevelDisplayFromSummary(topLevelSummary)
	}

	browserRuntimeApplyDefaultCandidateRouteToPayloadShells(payload)
	browserRuntimeSyncTopLevelSurfaceSummary(payload)
}
