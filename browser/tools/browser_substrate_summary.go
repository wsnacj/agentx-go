package tools

import "strings"

// BrowserWorkbenchSubstrateSummary captures the shared substrate/default-route
// observation used by workbench-style browser diagnostics.
type BrowserWorkbenchSubstrateSummary struct {
	DefaultRoute                 BrowserRuntimeInfo
	SubstrateSource              string
	SubstrateEndpoint            string
	DefaultCandidateRoute        BrowserRuntimeInfo
	DefaultCandidateSource       string
	DefaultCandidateEndpoint     string
	HostRoute                    BrowserRuntimeInfo
	HostRouteAvailable           bool
	HostFailureCause             string
	SubstratePosture             string
	SubstrateStatus              string
	SubstrateReason              string
	RepairCommand                string
	ConfiguredTargets            []string
	SelectionStrategy            string
	SelectionReason              string
	NodeConfigured               bool
	NodeRouteAvailable           bool
	NodePromotionReady           bool
	NodePromotionFailureCause    string
	SandboxConfigured            bool
	SandboxRouteAvailable        bool
	SandboxPromotionReady        bool
	SandboxPromotionFailureCause string
	SandboxFailureCause          string
}

func BrowserWorkbenchSubstrateSummaryForOptions(opts BrowserToolOptions) BrowserWorkbenchSubstrateSummary {
	return browserDefaultRuntimePreviewForToolOptions(opts).SubstrateSummary
}

func browserWorkbenchSubstrateSummaryForAssessment(opts BrowserToolOptions, assessment browserDefaultSubstrateAssessment) BrowserWorkbenchSubstrateSummary {
	return browserWorkbenchSubstrateSummaryForAssessmentWithBackend(opts, nil, assessment)
}

func browserWorkbenchSubstrateSummaryForAssessmentWithBackend(opts BrowserToolOptions, backend BrowserBackend, assessment browserDefaultSubstrateAssessment) BrowserWorkbenchSubstrateSummary {
	return browserWorkbenchSubstrateSummaryForBackend(opts, backend, assessment)
}

func browserWorkbenchSummaryDefaultRouteForAssessment(assessment browserDefaultSubstrateAssessment) BrowserRuntimeInfo {
	return browserVisibleDefaultRouteRuntimeInfoForAssessment(assessment)
}

func browserWorkbenchSubstrateSummaryForBackend(opts BrowserToolOptions, backend BrowserBackend, assessment browserDefaultSubstrateAssessment) BrowserWorkbenchSubstrateSummary {
	assessment = browserSubstrateAssessmentForBackend(backend, assessment)
	assessment = browserSubstrateAssessmentForConfiguredBackends(opts, assessment)
	if promotedAssessment, promoted := browserSurfacePromotedDefaultSubstrateAssessmentForBackend(opts, backend, assessment); promoted {
		assessment = promotedAssessment
	}
	logicalDefaultRoute := browserDefaultRouteRuntimeInfoForAssessment(assessment)
	defaultRoute := browserWorkbenchSummaryDefaultRouteForAssessment(assessment)
	selectionReason := browserWorkbenchSubstrateSelectionReason(logicalDefaultRoute, assessment.HostRuntime, assessment.HostRoute, assessment.NodeRoute, assessment.SandboxRoute)
	summary := BrowserWorkbenchSubstrateSummary{
		DefaultRoute:                 defaultRoute,
		DefaultCandidateRoute:        defaultRoute,
		HostRoute:                    assessment.HostRuntime,
		HostRouteAvailable:           assessment.HostRoute.RouteAvailable,
		HostFailureCause:             strings.TrimSpace(assessment.HostRoute.FailureReason),
		RepairCommand:                browserWorkbenchSubstrateRepairCommand(opts.RepairScript, selectionReason, assessment),
		ConfiguredTargets:            browserWorkbenchConfiguredTargetsForOptions(opts),
		SelectionStrategy:            BrowserSubstrateSelectionStrategy(logicalDefaultRoute, assessment.HostRuntime),
		SelectionReason:              selectionReason,
		NodeConfigured:               assessment.NodeRoute.Configured,
		NodeRouteAvailable:           assessment.NodeRoute.RouteAvailable,
		NodePromotionReady:           assessment.NodeRoute.Ready,
		NodePromotionFailureCause:    strings.TrimSpace(assessment.NodeRoute.FailureReason),
		SandboxConfigured:            assessment.SandboxRoute.Configured,
		SandboxRouteAvailable:        assessment.SandboxConcreteRoute.RouteAvailable,
		SandboxPromotionReady:        assessment.SandboxRoute.Ready,
		SandboxPromotionFailureCause: strings.TrimSpace(assessment.SandboxRoute.FailureReason),
		SandboxFailureCause:          strings.TrimSpace(assessment.SandboxConcreteRoute.FailureReason),
	}
	return browserCanonicalizeHiddenManagedRuntimePreview(browserDefaultRuntimePreview{
		EffectiveBackend:              backend,
		SubstrateAssessment:           assessment,
		SubstrateSummary:              summary,
		LogicalDefaultRoute:           logicalDefaultRoute,
		VisibleDefaultRoute:           defaultRoute,
		HiddenImplicitHostDefaultBase: normalizeBrowserRuntimeInfo(defaultRoute) == (BrowserRuntimeInfo{}),
	}).SubstrateSummary
}

func browserWorkbenchSubstrateDescriptor(summary BrowserWorkbenchSubstrateSummary) BrowserRuntimeInfo {
	if info := normalizeBrowserRuntimeInfo(summary.DefaultRoute); info != (BrowserRuntimeInfo{}) {
		return info
	}
	if strings.TrimSpace(summary.SelectionStrategy) == BrowserSubstrateSelectionLegacyHostDefault {
		return firstBrowserRuntimeInfo(summary.HostRoute, defaultBrowserRuntimeInfo())
	}
	return BrowserRuntimeInfo{}
}

func browserWorkbenchPopulateRouteProvenance(preview browserDefaultRuntimePreview) browserDefaultRuntimePreview {
	defaultRoute := normalizeBrowserRuntimeInfo(preview.LogicalDefaultRoute)
	diagnosticsPreview := browserRuntimeDiagnosticsPreview{
		Registration:           preview,
		DefaultRoute:           defaultRoute,
		DefaultRouteDescriptor: browserRuntimeVisibleDefaultRouteDescriptorForPreview(preview, defaultRoute),
		ConfiguredTargets:      append([]string(nil), preview.SubstrateSummary.ConfiguredTargets...),
	}
	if info := normalizeBrowserRuntimeInfo(browserWorkbenchSubstrateDescriptor(preview.SubstrateSummary)); info != (BrowserRuntimeInfo{}) {
		projection := browserRuntimeDoctorVisibleDefaultRouteProjection(diagnosticsPreview, info)
		preview.SubstrateSummary.SubstrateSource = strings.TrimSpace(projection.Metadata.Source)
		preview.SubstrateSummary.SubstrateEndpoint = strings.TrimSpace(projection.Metadata.Endpoint)
	}
	if candidate := normalizeBrowserRuntimeInfo(preview.SubstrateSummary.DefaultCandidateRoute); candidate != (BrowserRuntimeInfo{}) {
		if projection, ok := browserRuntimeDoctorManagedCandidateRouteProjection(diagnosticsPreview); ok {
			projected := normalizeBrowserRuntimeInfo(projection.Info)
			if projected == candidate ||
				(strings.EqualFold(strings.TrimSpace(projected.Backend), strings.TrimSpace(candidate.Backend)) &&
					strings.EqualFold(strings.TrimSpace(projected.Target), strings.TrimSpace(candidate.Target))) {
				preview.SubstrateSummary.DefaultCandidateSource = strings.TrimSpace(projection.Metadata.Source)
				preview.SubstrateSummary.DefaultCandidateEndpoint = strings.TrimSpace(projection.Metadata.Endpoint)
			}
		}
	}
	return preview
}

func browserWorkbenchSubstrateRepairCommand(
	repairScript string,
	selectionReason string,
	assessment browserDefaultSubstrateAssessment,
) string {
	bootstrapCode := browserRuntimeBootstrapErrorCodeFromFailureText(
		selectionReason,
		assessment.NodeRoute.FailureReason,
		assessment.NodeRoute.FailureNote,
		assessment.SandboxRoute.FailureReason,
		assessment.SandboxRoute.FailureNote,
		assessment.SandboxConcreteRoute.FailureReason,
		assessment.SandboxConcreteRoute.FailureNote,
	)
	return browserRuntimeBootstrapRepairCommand(repairScript, bootstrapCode)
}

func browserSubstrateAssessmentForConfiguredBackends(opts BrowserToolOptions, assessment browserDefaultSubstrateAssessment) browserDefaultSubstrateAssessment {
	storedDefaultRuntime := normalizeBrowserRuntimeInfo(assessment.DefaultRuntime)
	if opts.NodeBackend == nil {
		assessment.NodeRoute = browserDefaultPromotionRouteAssessment{}
		if normalizeBrowserRuntimeInfo(assessment.DefaultConcreteRoute.Route.RuntimeInfo).Target == "node" {
			assessment.DefaultConcreteRoute = browserConcreteRouteAssessment{}
		}
	}
	if opts.SandboxBackend == nil {
		assessment.SandboxRoute = browserDefaultPromotionRouteAssessment{}
		assessment.SandboxConcreteRoute = browserConcreteRouteAssessment{}
		if normalizeBrowserRuntimeInfo(assessment.DefaultConcreteRoute.Route.RuntimeInfo).Target == "sandbox" {
			assessment.DefaultConcreteRoute = browserConcreteRouteAssessment{}
		}
	}
	assessment.DefaultRuntime = browserPromotedDefaultRuntimeInfoForAssessments(
		assessment.HostRuntime,
		assessment.NodeRoute,
		assessment.SandboxRoute,
	)
	if normalizeBrowserRuntimeInfo(assessment.DefaultRuntime) == (BrowserRuntimeInfo{}) {
		assessment.DefaultRuntime = storedDefaultRuntime
	}
	assessment.DefaultConcreteRoute = browserSharedDefaultConcreteRouteAssessment(opts, assessment.DefaultRuntime, assessment)
	return assessment
}

func browserSubstrateAssessmentForBackend(backend BrowserBackend, assessment browserDefaultSubstrateAssessment) browserDefaultSubstrateAssessment {
	storedDefaultRuntime := normalizeBrowserRuntimeInfo(assessment.DefaultRuntime)
	if preferred, ok := browserRuntimeRouterPreferredDefaultConcreteRouteAssessment(backend); ok {
		assessment.DefaultConcreteRoute = preferred
		switch normalizeBrowserRuntimeInfo(preferred.Route.RuntimeInfo).Target {
		case "node":
			assessment.NodeRoute = browserDefaultPromotionRouteAssessmentForConcreteRouteAssessment(preferred, "node")
		case "sandbox":
			assessment.SandboxConcreteRoute = preferred
			assessment.SandboxRoute = browserDefaultPromotionRouteAssessmentForConcreteRouteAssessment(preferred, "sandbox")
		}
	} else if cached, ok := browserRuntimeRouterCachedDefaultConcreteRouteAssessment(backend); ok {
		assessment.DefaultConcreteRoute = cached
	}
	if cached, ok := browserRuntimeRouterCachedRouteAssessment(backend, BrowserRuntimeInfo{Target: "host"}); ok {
		assessment.HostRoute = cached
		if cached.RouteAvailable {
			assessment.HostRuntime = cached.Route.RuntimeInfo
		}
	}
	if cached, ok := browserRuntimeRouterCachedRouteAssessment(backend, BrowserRuntimeInfo{Target: "node"}); ok {
		assessment.NodeRoute = browserDefaultPromotionRouteAssessmentForConcreteRouteAssessment(cached, "node")
	}
	if cached, ok := browserRuntimeRouterCachedRouteAssessment(backend, BrowserRuntimeInfo{Target: "sandbox"}); ok {
		assessment.SandboxConcreteRoute = cached
		assessment.SandboxRoute = browserDefaultPromotionRouteAssessmentForConcreteRouteAssessment(cached, "sandbox")
	}
	assessment.DefaultRuntime = browserPromotedDefaultRuntimeInfoForAssessments(
		assessment.HostRuntime,
		assessment.NodeRoute,
		assessment.SandboxRoute,
	)
	if normalizeBrowserRuntimeInfo(assessment.DefaultRuntime) == (BrowserRuntimeInfo{}) {
		assessment.DefaultRuntime = storedDefaultRuntime
	}
	return assessment
}

func browserWorkbenchConfiguredTargetsForOptions(opts BrowserToolOptions) []string {
	targets := []string{"host"}
	if opts.SandboxBackend != nil {
		targets = append(targets, "sandbox")
	}
	if opts.NodeBackend != nil {
		targets = append(targets, "node")
	}
	return targets
}

func browserWorkbenchSubstrateSelectionReason(defaultInfo BrowserRuntimeInfo, hostInfo BrowserRuntimeInfo, hostAssessment browserConcreteRouteAssessment, nodeAssessment browserDefaultPromotionRouteAssessment, sandboxAssessment browserDefaultPromotionRouteAssessment) string {
	defaultInfo = normalizeBrowserRuntimeInfo(defaultInfo)
	if defaultInfo.Target == "host" && !hostAssessment.RouteAvailable && strings.TrimSpace(hostAssessment.FailureReason) != "" {
		return browserWorkbenchSpecificSelectionReason(strings.TrimSpace(hostAssessment.FailureReason))
	}
	if BrowserSubstrateSelectionStrategy(defaultInfo, hostInfo) == BrowserSubstrateSelectionLegacyHostDefault &&
		nodeAssessment.Configured &&
		!nodeAssessment.Ready &&
		strings.TrimSpace(nodeAssessment.FailureReason) != "" {
		return browserWorkbenchSpecificSelectionReason(strings.TrimSpace(nodeAssessment.FailureReason))
	}
	if BrowserSubstrateSelectionStrategy(defaultInfo, hostInfo) == BrowserSubstrateSelectionLegacyHostDefault &&
		sandboxAssessment.Configured &&
		!sandboxAssessment.Ready &&
		strings.TrimSpace(sandboxAssessment.FailureReason) != "" {
		return browserWorkbenchSpecificSelectionReason(strings.TrimSpace(sandboxAssessment.FailureReason))
	}
	return BrowserSubstrateSelectionReasonWithPromotionState(defaultInfo, hostInfo, nodeAssessment.Configured, nodeAssessment.Ready, sandboxAssessment.Configured, sandboxAssessment.Ready)
}

func browserWorkbenchSpecificSelectionReason(reason string) string {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return ""
	}
	if note := browserRuntimeBootstrapBlockedSurfaceNoteForFailureText(reason); note != "" {
		return note
	}
	return reason
}
