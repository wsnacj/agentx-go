package tools

import "strings"

type browserDefaultRuntimePreview struct {
	EffectiveBackend              BrowserBackend
	SubstrateAssessment           browserDefaultSubstrateAssessment
	SubstrateSummary              BrowserWorkbenchSubstrateSummary
	LogicalDefaultRoute           BrowserRuntimeInfo
	VisibleDefaultRoute           BrowserRuntimeInfo
	RegistrationCapabilities      BrowserCapabilities
	HiddenImplicitHostDefaultBase bool
}

func browserVisibleDefaultRuntimeInfoForPreview(preview browserDefaultRuntimePreview) BrowserRuntimeInfo {
	return normalizeBrowserRuntimeInfo(preview.VisibleDefaultRoute)
}

func browserDefaultRuntimePreviewLogicalDefaultRoute(summaryDefaultRoute BrowserRuntimeInfo, assessment browserDefaultSubstrateAssessment, backend BrowserBackend, fallback BrowserRuntimeInfo) BrowserRuntimeInfo {
	if info := normalizeBrowserRuntimeInfo(summaryDefaultRoute); info != (BrowserRuntimeInfo{}) {
		return info
	}
	if browserRegistrationHasStoredSubstrateAssessment(assessment) {
		return browserDefaultRouteRuntimeInfoForAssessment(assessment)
	}
	fallback = normalizeBrowserRuntimeInfo(fallback)
	if fallback == (BrowserRuntimeInfo{}) {
		fallback = defaultBrowserRuntimeInfo()
	}
	return browserRuntimeInfoForConcreteBackend(backend, fallback)
}

func browserDefaultRuntimePreviewForOwner(opts BrowserToolOptions, owner browserDefaultRouteOwner) browserDefaultRuntimePreview {
	effectiveBackend := owner.backend
	assessment := owner.substrateAssessment
	summary := owner.substrateSummary
	logicalDefaultRoute := normalizeBrowserRuntimeInfo(owner.defaultRoute)
	if logicalDefaultRoute == (BrowserRuntimeInfo{}) {
		logicalDefaultRoute = browserDefaultRuntimePreviewLogicalDefaultRoute(summary.DefaultRoute, assessment, effectiveBackend, owner.defaultRoute)
	}
	visibleDefaultRoute := normalizeBrowserRuntimeInfo(summary.DefaultRoute)
	if visibleDefaultRoute == (BrowserRuntimeInfo{}) {
		visibleDefaultRoute = browserVisibleDefaultRouteRuntimeInfo(logicalDefaultRoute, assessment)
	}
	return browserCanonicalizeHiddenManagedRuntimePreview(browserDefaultRuntimePreview{
		EffectiveBackend:              effectiveBackend,
		SubstrateAssessment:           assessment,
		SubstrateSummary:              summary,
		LogicalDefaultRoute:           logicalDefaultRoute,
		VisibleDefaultRoute:           visibleDefaultRoute,
		RegistrationCapabilities:      browserCapabilitiesForRegistrationWithBackend(effectiveBackend, assessment),
		HiddenImplicitHostDefaultBase: normalizeBrowserRuntimeInfo(visibleDefaultRoute) == (BrowserRuntimeInfo{}),
	})
}

func browserDefaultRuntimePreviewForToolOptions(opts BrowserToolOptions) browserDefaultRuntimePreview {
	return browserDefaultRuntimePreviewForOwner(opts, browserSurfaceDefaultRouteOwnerForToolOptions(opts))
}

func browserDefaultRuntimePreviewForDispatchOptions(opts BrowserToolOptions, policy outboundNetworkPolicy, timeoutMs int) browserDefaultRuntimePreview {
	return browserDefaultRuntimePreviewForOwner(opts, browserSurfaceDefaultRouteOwnerForOptions(opts, policy, timeoutMs))
}

func browserRegistrationFallbackPreviewTimeoutMs(ctx browserRegistrationContext) int {
	timeoutMs := ctx.timeoutMs
	if timeoutMs <= 0 {
		timeoutMs = ctx.opts.TimeoutMs
	}
	if timeoutMs <= 0 {
		timeoutMs = 20_000
	}
	return timeoutMs
}

func browserRegistrationFallbackRuntimePreview(ctx browserRegistrationContext) browserDefaultRuntimePreview {
	return browserDefaultRuntimePreviewForDispatchOptions(ctx.opts, ctx.policy, browserRegistrationFallbackPreviewTimeoutMs(ctx))
}

func browserRegistrationStoredSummaryRequiresFallbackPreview(ctx browserRegistrationContext, fallback browserDefaultRuntimePreview) bool {
	storedDefaultRoute := normalizeBrowserRuntimeInfo(ctx.substrateSummary.DefaultRoute)
	fallbackVisibleDefaultRoute := normalizeBrowserRuntimeInfo(fallback.VisibleDefaultRoute)
	if fallback.HiddenImplicitHostDefaultBase && storedDefaultRoute != (BrowserRuntimeInfo{}) {
		return true
	}
	if fallbackVisibleDefaultRoute != (BrowserRuntimeInfo{}) && storedDefaultRoute != fallbackVisibleDefaultRoute {
		return true
	}
	if browserRegistrationStoredSummaryMissesFallbackRouteEvidence(ctx.substrateSummary, fallback.SubstrateSummary) {
		return true
	}
	if storedDefaultRoute == (BrowserRuntimeInfo{}) || ctx.backend == nil || fallback.EffectiveBackend == nil {
		return false
	}
	provider, ok := ctx.backend.(BrowserRuntimeInfoProvider)
	if !ok {
		return false
	}
	return normalizeBrowserRuntimeInfo(provider.BrowserRuntimeInfo()) != storedDefaultRoute
}

func browserRegistrationStoredSummaryMissesFallbackRouteEvidence(
	stored BrowserWorkbenchSubstrateSummary,
	fallback BrowserWorkbenchSubstrateSummary,
) bool {
	stored = browserCanonicalizeHiddenManagedSubstrateSummaryForComparison(stored)
	fallback = browserCanonicalizeHiddenManagedSubstrateSummaryForComparison(fallback)
	fallbackHasSelectionEvidence := browserRegistrationFallbackSummaryHasSelectionEvidence(fallback)
	fallbackDefaultRoute := normalizeBrowserRuntimeInfo(fallback.DefaultRoute)
	if fallbackDefaultRoute != (BrowserRuntimeInfo{}) &&
		normalizeBrowserRuntimeInfo(stored.DefaultRoute) != fallbackDefaultRoute {
		return true
	}
	if normalizeBrowserRuntimeInfo(stored.DefaultCandidateRoute) != normalizeBrowserRuntimeInfo(fallback.DefaultCandidateRoute) {
		return true
	}
	if browserRegistrationStoredSummaryMissingTargets(stored.ConfiguredTargets, fallback.ConfiguredTargets, fallbackHasSelectionEvidence) {
		return true
	}
	if strings.TrimSpace(stored.SelectionStrategy) != strings.TrimSpace(fallback.SelectionStrategy) {
		return true
	}
	if strings.TrimSpace(stored.SelectionReason) != strings.TrimSpace(fallback.SelectionReason) {
		return true
	}
	if strings.TrimSpace(stored.SubstratePosture) != strings.TrimSpace(fallback.SubstratePosture) {
		return true
	}
	if strings.TrimSpace(stored.SubstrateStatus) != strings.TrimSpace(fallback.SubstrateStatus) {
		return true
	}
	if strings.TrimSpace(stored.SubstrateReason) != strings.TrimSpace(fallback.SubstrateReason) {
		return true
	}
	if strings.TrimSpace(stored.RepairCommand) != strings.TrimSpace(fallback.RepairCommand) {
		return true
	}
	if browserRegistrationStoredSummaryMissesFallbackHostLaneEvidence(stored, fallback) {
		return true
	}
	if browserRegistrationStoredSummaryMissesFallbackManagedLaneEvidence(
		stored.NodeConfigured,
		stored.NodeRouteAvailable,
		stored.NodePromotionReady,
		stored.NodePromotionFailureCause,
		"",
		fallback.NodeConfigured,
		fallback.NodeRouteAvailable,
		fallback.NodePromotionReady,
		fallback.NodePromotionFailureCause,
		"",
	) {
		return true
	}
	return browserRegistrationStoredSummaryMissesFallbackManagedLaneEvidence(
		stored.SandboxConfigured,
		stored.SandboxRouteAvailable,
		stored.SandboxPromotionReady,
		stored.SandboxPromotionFailureCause,
		stored.SandboxFailureCause,
		fallback.SandboxConfigured,
		fallback.SandboxRouteAvailable,
		fallback.SandboxPromotionReady,
		fallback.SandboxPromotionFailureCause,
		fallback.SandboxFailureCause,
	)
}

func browserRegistrationFallbackSummaryHasSelectionEvidence(summary BrowserWorkbenchSubstrateSummary) bool {
	if normalizeBrowserRuntimeInfo(summary.DefaultRoute) != (BrowserRuntimeInfo{}) {
		return true
	}
	if normalizeBrowserRuntimeInfo(summary.DefaultCandidateRoute) != (BrowserRuntimeInfo{}) {
		return true
	}
	return len(mergeToolMetadataStrings(nil, summary.ConfiguredTargets)) > 1 ||
		summary.NodeConfigured ||
		summary.NodeRouteAvailable ||
		summary.NodePromotionReady ||
		summary.SandboxConfigured ||
		summary.SandboxRouteAvailable ||
		summary.SandboxPromotionReady
}

func browserRegistrationSummaryHasHostLaneEvidence(summary BrowserWorkbenchSubstrateSummary) bool {
	hostRoute := normalizeBrowserRuntimeInfo(summary.HostRoute)
	return strings.TrimSpace(summary.HostFailureCause) != "" ||
		!summary.HostRouteAvailable ||
		(hostRoute != (BrowserRuntimeInfo{}) &&
			BrowserSubstratePosture(hostRoute.Backend, hostRoute.Target) != BrowserSubstrateLegacySystemHost)
}

func browserRegistrationStoredSummaryMissesFallbackHostLaneEvidence(
	stored BrowserWorkbenchSubstrateSummary,
	fallback BrowserWorkbenchSubstrateSummary,
) bool {
	if !browserRegistrationSummaryHasHostLaneEvidence(stored) &&
		!browserRegistrationSummaryHasHostLaneEvidence(fallback) {
		return false
	}
	return normalizeBrowserRuntimeInfo(stored.HostRoute) != normalizeBrowserRuntimeInfo(fallback.HostRoute) ||
		stored.HostRouteAvailable != fallback.HostRouteAvailable ||
		strings.TrimSpace(stored.HostFailureCause) != strings.TrimSpace(fallback.HostFailureCause)
}

func browserRegistrationStoredSummaryMissesFallbackManagedLaneEvidence(
	storedConfigured bool,
	storedRouteAvailable bool,
	storedPromotionReady bool,
	storedPromotionFailure string,
	storedRouteFailure string,
	fallbackConfigured bool,
	fallbackRouteAvailable bool,
	fallbackPromotionReady bool,
	fallbackPromotionFailure string,
	fallbackRouteFailure string,
) bool {
	storedHasEvidence := storedConfigured ||
		storedRouteAvailable ||
		storedPromotionReady ||
		strings.TrimSpace(storedPromotionFailure) != "" ||
		strings.TrimSpace(storedRouteFailure) != ""
	fallbackHasEvidence := fallbackConfigured ||
		fallbackRouteAvailable ||
		fallbackPromotionReady ||
		strings.TrimSpace(fallbackPromotionFailure) != "" ||
		strings.TrimSpace(fallbackRouteFailure) != ""
	if !storedHasEvidence && !fallbackHasEvidence {
		return false
	}
	return storedConfigured != fallbackConfigured ||
		storedRouteAvailable != fallbackRouteAvailable ||
		storedPromotionReady != fallbackPromotionReady ||
		strings.TrimSpace(storedPromotionFailure) != strings.TrimSpace(fallbackPromotionFailure) ||
		strings.TrimSpace(storedRouteFailure) != strings.TrimSpace(fallbackRouteFailure)
}

func browserRegistrationStoredSummaryMissingTargets(storedTargets []string, fallbackTargets []string, fallbackHasSelectionEvidence bool) bool {
	_ = fallbackHasSelectionEvidence
	storedTargets = mergeToolMetadataStrings(nil, storedTargets)
	fallbackTargets = mergeToolMetadataStrings(nil, fallbackTargets)
	if len(storedTargets) == 0 && len(fallbackTargets) == 0 {
		return false
	}
	if len(storedTargets) != len(fallbackTargets) {
		return true
	}
	storedSet := map[string]bool{}
	for _, target := range storedTargets {
		storedSet[target] = true
	}
	for _, target := range fallbackTargets {
		if !storedSet[target] {
			return true
		}
	}
	fallbackSet := map[string]bool{}
	for _, target := range fallbackTargets {
		fallbackSet[target] = true
	}
	for _, target := range storedTargets {
		if !fallbackSet[target] {
			return true
		}
	}
	return false
}

func browserRegistrationShouldUseFallbackOwnerBackendForStoredAssessment(
	ctx browserRegistrationContext,
	fallback browserDefaultRuntimePreview,
) bool {
	if fallback.EffectiveBackend == nil {
		return false
	}
	if browserRegistrationUsesLegacyHostFallbackBackend(ctx.backend) {
		return true
	}
	storedDefaultRoute := browserDefaultRouteRuntimeInfoForAssessment(ctx.substrateAssessment)
	if browserRegistrationStoredAssessmentMissesFallbackHostLaneEvidence(
		ctx.substrateAssessment,
		fallback.SubstrateAssessment,
	) && (ctx.backend == nil || browserRegistrationUsesLegacyHostFallbackBackend(ctx.backend)) {
		return true
	}
	fallbackDefaultRoute := normalizeBrowserRuntimeInfo(fallback.LogicalDefaultRoute)
	if browserRegistrationUsesLegacyHostFallbackBackend(ctx.backend) &&
		fallbackDefaultRoute != (BrowserRuntimeInfo{}) &&
		!browserRuntimeUsesImplicitLegacyHostDefaultFallback(
			fallbackDefaultRoute,
			fallback.SubstrateAssessment,
		) {
		return true
	}
	if (ctx.backend == nil || browserRegistrationUsesLegacyHostFallbackBackend(ctx.backend)) &&
		(storedDefaultRoute == (BrowserRuntimeInfo{}) ||
			browserRuntimeUsesImplicitLegacyHostDefaultFallback(storedDefaultRoute, ctx.substrateAssessment)) &&
		(browserRegistrationStoredAssessmentMissesFallbackManagedLaneEvidence(
			ctx.substrateAssessment.NodeRoute,
			fallback.SubstrateAssessment.NodeRoute,
		) ||
			browserRegistrationStoredAssessmentMissesFallbackManagedLaneEvidence(
				ctx.substrateAssessment.SandboxRoute,
				fallback.SubstrateAssessment.SandboxRoute,
			)) {
		return true
	}
	if ctx.backend != nil {
		return false
	}
	if storedDefaultRoute != (BrowserRuntimeInfo{}) &&
		!browserRuntimeUsesImplicitLegacyHostDefaultFallback(storedDefaultRoute, ctx.substrateAssessment) {
		return false
	}
	if fallbackDefaultRoute == (BrowserRuntimeInfo{}) {
		return false
	}
	return !browserRuntimeUsesImplicitLegacyHostDefaultFallback(
		fallbackDefaultRoute,
		fallback.SubstrateAssessment,
	)
}

func browserRegistrationUsesLegacyHostFallbackBackend(backend BrowserBackend) bool {
	return false
}

func browserRegistrationAssessmentHasHostLaneEvidence(assessment browserDefaultSubstrateAssessment) bool {
	hostRoute := normalizeBrowserRuntimeInfo(assessment.HostRuntime)
	resolvedHostRoute := normalizeBrowserRuntimeInfo(assessment.HostRoute.Route.RuntimeInfo)
	return strings.TrimSpace(assessment.HostRoute.FailureReason) != "" ||
		(hostRoute != (BrowserRuntimeInfo{}) &&
			BrowserSubstratePosture(hostRoute.Backend, hostRoute.Target) != BrowserSubstrateLegacySystemHost) ||
		(resolvedHostRoute != (BrowserRuntimeInfo{}) &&
			BrowserSubstratePosture(resolvedHostRoute.Backend, resolvedHostRoute.Target) != BrowserSubstrateLegacySystemHost)
}

func browserRegistrationStoredAssessmentMissesFallbackHostLaneEvidence(
	stored browserDefaultSubstrateAssessment,
	fallback browserDefaultSubstrateAssessment,
) bool {
	if !browserRegistrationAssessmentHasHostLaneEvidence(stored) &&
		!browserRegistrationAssessmentHasHostLaneEvidence(fallback) {
		return false
	}
	return normalizeBrowserRuntimeInfo(stored.HostRuntime) != normalizeBrowserRuntimeInfo(fallback.HostRuntime) ||
		normalizeBrowserRuntimeInfo(stored.HostRoute.Route.RuntimeInfo) != normalizeBrowserRuntimeInfo(fallback.HostRoute.Route.RuntimeInfo) ||
		stored.HostRoute.RouteAvailable != fallback.HostRoute.RouteAvailable ||
		strings.TrimSpace(stored.HostRoute.FailureReason) != strings.TrimSpace(fallback.HostRoute.FailureReason)
}

func browserRegistrationStoredAssessmentMissesFallbackManagedLaneEvidence(
	stored browserDefaultPromotionRouteAssessment,
	fallback browserDefaultPromotionRouteAssessment,
) bool {
	if !fallback.Configured {
		return false
	}
	return stored.Configured != fallback.Configured ||
		stored.RouteAvailable != fallback.RouteAvailable ||
		stored.Ready != fallback.Ready ||
		normalizeBrowserRuntimeInfo(stored.Route.RuntimeInfo) != normalizeBrowserRuntimeInfo(fallback.Route.RuntimeInfo) ||
		strings.TrimSpace(stored.FailureReason) != strings.TrimSpace(fallback.FailureReason) ||
		strings.TrimSpace(stored.FailureNote) != strings.TrimSpace(fallback.FailureNote)
}

func browserRegistrationPreviewEffectiveBackend(
	ctx browserRegistrationContext,
	fallback browserDefaultRuntimePreview,
	hasStoredAssessment bool,
) BrowserBackend {
	if hasStoredAssessment {
		if browserRegistrationShouldUseFallbackOwnerBackendForStoredAssessment(ctx, fallback) {
			return fallback.EffectiveBackend
		}
		return ctx.backend
	}
	if fallback.EffectiveBackend != nil {
		return fallback.EffectiveBackend
	}
	if ctx.backend != nil {
		return ctx.backend
	}
	return fallback.EffectiveBackend
}

func browserRegistrationDefaultRuntimePreview(ctx browserRegistrationContext) browserDefaultRuntimePreview {
	hasStoredAssessment := browserRegistrationHasStoredSubstrateAssessment(ctx.substrateAssessment)
	hasStoredSummary := browserRegistrationHasStoredSubstrateSummary(ctx.substrateSummary)
	fallback := browserDefaultRuntimePreview{}
	if !hasStoredAssessment || ctx.backend == nil || browserRegistrationUsesLegacyHostFallbackBackend(ctx.backend) {
		fallback = browserRegistrationFallbackRuntimePreview(ctx)
	}
	if !hasStoredAssessment {
		if !hasStoredSummary || browserRegistrationStoredSummaryRequiresFallbackPreview(ctx, fallback) {
			return fallback
		}
	}
	assessment := ctx.substrateAssessment
	if !hasStoredAssessment {
		assessment = fallback.SubstrateAssessment
	}
	effectiveBackend := browserRegistrationPreviewEffectiveBackend(ctx, fallback, hasStoredAssessment)
	assessment = browserSubstrateAssessmentForBackend(effectiveBackend, assessment)
	assessment = browserSubstrateAssessmentForConfiguredBackends(ctx.opts, assessment)
	runtimeOnlyManagedDefaultPromoted := false
	if promotedAssessment, promoted := browserSurfacePromotedDefaultSubstrateAssessmentForBackend(ctx.opts, effectiveBackend, assessment); promoted {
		assessment = promotedAssessment
		runtimeOnlyManagedDefaultPromoted = true
	}
	summary := ctx.substrateSummary
	switch {
	case runtimeOnlyManagedDefaultPromoted:
		summary = browserWorkbenchSubstrateSummaryForAssessmentWithBackend(ctx.opts, effectiveBackend, assessment)
	case hasStoredAssessment:
		summary = browserWorkbenchSubstrateSummaryForBackend(ctx.opts, effectiveBackend, assessment)
	case hasStoredSummary:
		summary = ctx.substrateSummary
	default:
		summary = fallback.SubstrateSummary
	}
	logicalDefaultFallback := BrowserRuntimeInfo{}
	if !hasStoredAssessment {
		logicalDefaultFallback = fallback.LogicalDefaultRoute
	}
	logicalDefaultRoute := browserDefaultRuntimePreviewLogicalDefaultRoute(summary.DefaultRoute, assessment, effectiveBackend, logicalDefaultFallback)
	visibleDefaultRoute := browserVisibleDefaultRouteRuntimeInfo(logicalDefaultRoute, assessment)
	return browserCanonicalizeHiddenManagedRuntimePreview(browserDefaultRuntimePreview{
		EffectiveBackend:              effectiveBackend,
		SubstrateAssessment:           assessment,
		SubstrateSummary:              summary,
		LogicalDefaultRoute:           logicalDefaultRoute,
		VisibleDefaultRoute:           visibleDefaultRoute,
		HiddenImplicitHostDefaultBase: normalizeBrowserRuntimeInfo(visibleDefaultRoute) == (BrowserRuntimeInfo{}),
	})
}
