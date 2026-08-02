package tools

import "strings"

func browserCanonicalizeHiddenManagedRuntimePreview(preview browserDefaultRuntimePreview) browserDefaultRuntimePreview {
	if candidate, ok := browserCanonicalizeHiddenManagedDefaultCandidateRoute(preview); ok {
		preview.SubstrateSummary.DefaultCandidateRoute = candidate
	}
	selectionStrategy, selectionReason, ok := browserCanonicalizeHiddenManagedSelectionSummary(
		preview,
		strings.TrimSpace(preview.SubstrateSummary.SelectionStrategy),
		strings.TrimSpace(preview.SubstrateSummary.SelectionReason),
	)
	if ok {
		preview.SubstrateSummary.SelectionStrategy = selectionStrategy
		preview.SubstrateSummary.SelectionReason = selectionReason
	}
	preview.SubstrateSummary = browserWorkbenchApplyTopLevelSubstrateSummary(preview.SubstrateSummary)
	return browserWorkbenchPopulateRouteProvenance(preview)
}

func browserCanonicalizeHiddenManagedDefaultCandidateRoute(preview browserDefaultRuntimePreview) (BrowserRuntimeInfo, bool) {
	if info := normalizeBrowserRuntimeInfo(preview.SubstrateSummary.DefaultCandidateRoute); info != (BrowserRuntimeInfo{}) {
		return info, true
	}
	if info := normalizeBrowserRuntimeInfo(preview.VisibleDefaultRoute); info != (BrowserRuntimeInfo{}) {
		return info, true
	}
	logicalDefaultRoute := normalizeBrowserRuntimeInfo(preview.LogicalDefaultRoute)
	if !preview.HiddenImplicitHostDefaultBase ||
		!browserRuntimeUsesImplicitLegacyHostDefaultFallback(logicalDefaultRoute, preview.SubstrateAssessment) {
		return BrowserRuntimeInfo{}, false
	}
	diagnosticsPreview := browserRuntimeDiagnosticsPreview{
		Registration:           preview,
		DefaultRoute:           logicalDefaultRoute,
		DefaultRouteDescriptor: browserRuntimeRouteDescriptor{},
		ConfiguredTargets:      append([]string(nil), preview.SubstrateSummary.ConfiguredTargets...),
	}
	route := browserRuntimeDoctorRouteSummary(&browserRuntimePayload{
		ConfiguredTargets:          append([]string(nil), diagnosticsPreview.ConfiguredTargets...),
		SubstrateSelectionStrategy: strings.TrimSpace(preview.SubstrateSummary.SelectionStrategy),
		SubstrateSelectionReason:   strings.TrimSpace(preview.SubstrateSummary.SelectionReason),
	}, diagnosticsPreview)
	if route == nil || strings.TrimSpace(route.Code) != "managed_route_hidden_by_legacy_host_default" {
		return BrowserRuntimeInfo{}, false
	}
	info := normalizeBrowserRuntimeInfo(BrowserRuntimeInfo{
		Backend: strings.TrimSpace(route.Backend),
		Profile: strings.TrimSpace(route.Profile),
		Target:  strings.TrimSpace(route.RuntimeTarget),
	})
	if info == (BrowserRuntimeInfo{}) {
		return BrowserRuntimeInfo{}, false
	}
	return info, true
}

func browserCanonicalizeHiddenManagedSubstrateSummaryForComparison(summary BrowserWorkbenchSubstrateSummary) BrowserWorkbenchSubstrateSummary {
	logicalDefaultRoute := normalizeBrowserRuntimeInfo(summary.DefaultRoute)
	if logicalDefaultRoute == (BrowserRuntimeInfo{}) {
		logicalDefaultRoute = firstBrowserRuntimeInfo(summary.HostRoute, defaultBrowserRuntimeInfo())
	}
	visibleDefaultRoute := normalizeBrowserRuntimeInfo(summary.DefaultRoute)
	assessment := browserDefaultSubstrateAssessment{
		HostRuntime: firstBrowserRuntimeInfo(summary.HostRoute, defaultBrowserRuntimeInfo()),
		HostRoute: browserConcreteRouteAssessment{
			Configured:     summary.HostRouteAvailable || normalizeBrowserRuntimeInfo(summary.HostRoute) != (BrowserRuntimeInfo{}) || strings.TrimSpace(summary.HostFailureCause) != "",
			RouteAvailable: summary.HostRouteAvailable,
			Route: browserResolvedExecutionRoute{
				RuntimeInfo: normalizeBrowserRuntimeInfo(summary.HostRoute),
			},
			FailureReason: strings.TrimSpace(summary.HostFailureCause),
			FailureNote:   strings.TrimSpace(summary.HostFailureCause),
		},
		DefaultRuntime: logicalDefaultRoute,
		NodeRoute: browserDefaultPromotionRouteAssessment{
			Configured:     summary.NodeConfigured,
			RouteAvailable: summary.NodeRouteAvailable,
			Ready:          summary.NodePromotionReady,
			Route: browserResolvedExecutionRoute{
				RuntimeInfo: defaultBrowserNodeRuntimeInfo(),
			},
			FailureReason: strings.TrimSpace(summary.NodePromotionFailureCause),
			FailureNote:   strings.TrimSpace(summary.NodePromotionFailureCause),
		},
		SandboxConcreteRoute: browserConcreteRouteAssessment{
			Configured:     summary.SandboxConfigured || summary.SandboxRouteAvailable || strings.TrimSpace(summary.SandboxFailureCause) != "",
			RouteAvailable: summary.SandboxRouteAvailable,
			Route: browserResolvedExecutionRoute{
				RuntimeInfo: defaultBrowserSandboxRuntimeInfo(),
			},
			FailureReason: strings.TrimSpace(summary.SandboxFailureCause),
			FailureNote:   strings.TrimSpace(summary.SandboxFailureCause),
		},
		SandboxRoute: browserDefaultPromotionRouteAssessment{
			Configured:     summary.SandboxConfigured,
			RouteAvailable: summary.SandboxRouteAvailable,
			Ready:          summary.SandboxPromotionReady,
			Route: browserResolvedExecutionRoute{
				RuntimeInfo: defaultBrowserSandboxRuntimeInfo(),
			},
			FailureReason: strings.TrimSpace(summary.SandboxPromotionFailureCause),
			FailureNote:   strings.TrimSpace(summary.SandboxPromotionFailureCause),
		},
	}
	if visibleDefaultRoute != (BrowserRuntimeInfo{}) {
		assessment.DefaultConcreteRoute = browserConcreteRouteAssessment{
			Configured:     true,
			RouteAvailable: true,
			Route: browserResolvedExecutionRoute{
				RuntimeInfo: visibleDefaultRoute,
			},
		}
	}
	return browserCanonicalizeHiddenManagedRuntimePreview(browserDefaultRuntimePreview{
		SubstrateAssessment:           assessment,
		SubstrateSummary:              summary,
		LogicalDefaultRoute:           logicalDefaultRoute,
		VisibleDefaultRoute:           visibleDefaultRoute,
		HiddenImplicitHostDefaultBase: visibleDefaultRoute == (BrowserRuntimeInfo{}),
	}).SubstrateSummary
}

func browserCanonicalizeHiddenManagedSelectionSummaryForDiagnosticsPreview(
	preview browserRuntimeDiagnosticsPreview,
	selectionStrategy string,
	selectionReason string,
) (string, string, bool) {
	registration := preview.Registration
	if normalizeBrowserRuntimeInfo(registration.LogicalDefaultRoute) == (BrowserRuntimeInfo{}) {
		registration.LogicalDefaultRoute = normalizeBrowserRuntimeInfo(preview.DefaultRoute)
	}
	if len(registration.SubstrateSummary.ConfiguredTargets) == 0 && len(preview.ConfiguredTargets) != 0 {
		registration.SubstrateSummary.ConfiguredTargets = append([]string(nil), preview.ConfiguredTargets...)
	}
	return browserCanonicalizeHiddenManagedSelectionSummary(
		registration,
		selectionStrategy,
		selectionReason,
	)
}

func browserCanonicalizeHiddenManagedPayloadSelectionSummary(
	payload *browserRuntimePayload,
	preview browserRuntimeDiagnosticsPreview,
) {
	if payload == nil {
		return
	}
	selectionStrategy, selectionReason, ok := browserCanonicalizeHiddenManagedSelectionSummaryForDiagnosticsPreview(
		preview,
		strings.TrimSpace(payload.SubstrateSelectionStrategy),
		strings.TrimSpace(payload.SubstrateSelectionReason),
	)
	if !ok {
		return
	}
	payload.SubstrateSelectionStrategy = selectionStrategy
	payload.SubstrateSelectionReason = selectionReason
}

func browserCanonicalizeHiddenManagedSelectionSummary(
	registration browserDefaultRuntimePreview,
	selectionStrategy string,
	selectionReason string,
) (string, string, bool) {
	logicalDefaultRoute := normalizeBrowserRuntimeInfo(registration.LogicalDefaultRoute)
	if !registration.HiddenImplicitHostDefaultBase ||
		!browserRuntimeUsesImplicitLegacyHostDefaultFallback(logicalDefaultRoute, registration.SubstrateAssessment) {
		return "", "", false
	}
	preview := browserRuntimeDiagnosticsPreview{
		Registration:           registration,
		DefaultRoute:           logicalDefaultRoute,
		DefaultRouteDescriptor: browserRuntimeRouteDescriptor{},
		ConfiguredTargets:      append([]string(nil), registration.SubstrateSummary.ConfiguredTargets...),
	}
	route := browserRuntimeDoctorRouteSummary(&browserRuntimePayload{
		ConfiguredTargets:          append([]string(nil), preview.ConfiguredTargets...),
		SubstrateSelectionStrategy: strings.TrimSpace(selectionStrategy),
		SubstrateSelectionReason:   strings.TrimSpace(selectionReason),
	}, preview)
	if route == nil ||
		strings.TrimSpace(route.Code) != "managed_route_hidden_by_legacy_host_default" ||
		strings.TrimSpace(route.RuntimeTarget) != "node" {
		return "", "", false
	}
	canonicalStrategy := strings.TrimSpace(selectionStrategy)
	if browserShouldCanonicalizeHiddenManagedSelectionStrategy(selectionReason) {
		canonicalStrategy = firstNonEmpty(
			strings.TrimSpace(route.SelectionStrategy),
			canonicalStrategy,
		)
	}
	canonicalReason := strings.TrimSpace(selectionReason)
	if browserShouldCanonicalizeHiddenManagedSelectionReason(selectionReason) {
		canonicalReason = firstNonEmpty(
			strings.TrimSpace(browserRuntimeDoctorRouteInspectionCanonicalSummary(route)),
			canonicalReason,
		)
	}
	if canonicalStrategy == strings.TrimSpace(selectionStrategy) &&
		canonicalReason == strings.TrimSpace(selectionReason) {
		return "", "", false
	}
	return canonicalStrategy, canonicalReason, true
}

func browserShouldCanonicalizeHiddenManagedSelectionStrategy(reason string) bool {
	reason = strings.ToLower(strings.TrimSpace(reason))
	switch {
	case reason == "":
		return true
	case strings.Contains(reason, "does not yet advertise the required default browser capabilities"):
		return false
	case strings.Contains(reason, "remains available via `runtime_target=node`"):
		return false
	default:
		return true
	}
}

func browserShouldCanonicalizeHiddenManagedSelectionReason(reason string) bool {
	reason = strings.ToLower(strings.TrimSpace(reason))
	switch {
	case reason == "":
		return true
	case strings.HasPrefix(reason, "bundled browserd bootstrap is blocked because"):
		return false
	case strings.Contains(reason, "could not be resolved"):
		return false
	case strings.Contains(reason, "managed browserd boot failed"):
		return false
	case strings.Contains(reason, "managed_route_unavailable"):
		return false
	case strings.Contains(reason, "does not yet advertise the required default browser capabilities"):
		return false
	case strings.Contains(reason, "not the default because"):
		return false
	case strings.Contains(reason, "runtime_target=node"):
		return false
	case strings.Contains(reason, "runtime_target=sandbox"):
		return false
	default:
		return true
	}
}
