package tools

import "strings"

type browserRuntimeDiagnosticsPreview struct {
	Registration           browserDefaultRuntimePreview
	DefaultRoute           BrowserRuntimeInfo
	DefaultRouteDescriptor browserRuntimeRouteDescriptor
	ConfiguredTargets      []string
	AvailableActions       []string
}

func browserRuntimeDiagnosticsPreviewForRegistration(ctx browserRegistrationContext) browserRuntimeDiagnosticsPreview {
	return browserRuntimeDiagnosticsPreviewForRegistrationPreview(ctx, browserRegistrationDefaultRuntimePreview(ctx))
}

func browserRuntimeDiagnosticsPreviewBaseForRegistration(ctx browserRegistrationContext) browserRuntimeDiagnosticsPreview {
	return browserRuntimeDiagnosticsPreviewBaseForRegistrationPreview(browserRegistrationDefaultRuntimePreview(ctx))
}

func browserRuntimeDiagnosticsPreviewForExecutionPreview(_ browserRegistrationContext, executionPreview browserRegistrationExecutionPreview) browserRuntimeDiagnosticsPreview {
	return browserRuntimeDiagnosticsPreviewBaseForRegistrationPreview(executionPreview.Registration)
}

func browserRuntimeDiagnosticsPreviewBaseForRegistrationPreview(registration browserDefaultRuntimePreview) browserRuntimeDiagnosticsPreview {
	registration = browserRuntimeDiagnosticsPreviewCanonicalRegistration(registration)
	defaultRoute := normalizeBrowserRuntimeInfo(registration.LogicalDefaultRoute)
	return browserRuntimeDiagnosticsPreview{
		Registration:           registration,
		DefaultRoute:           defaultRoute,
		DefaultRouteDescriptor: browserRuntimeVisibleDefaultRouteDescriptorForPreview(registration, defaultRoute),
		ConfiguredTargets:      append([]string(nil), registration.SubstrateSummary.ConfiguredTargets...),
	}
}

func browserRuntimeVisibleDefaultRouteDescriptorForPreview(
	registration browserDefaultRuntimePreview,
	defaultRoute BrowserRuntimeInfo,
) browserRuntimeRouteDescriptor {
	visible := browserVisibleDefaultRuntimeInfoForPreview(registration)
	if visible == (BrowserRuntimeInfo{}) {
		return browserRuntimeRouteDescriptor{}
	}
	preview := browserRuntimeDiagnosticsPreview{
		Registration:      registration,
		DefaultRoute:      normalizeBrowserRuntimeInfo(defaultRoute),
		ConfiguredTargets: append([]string(nil), registration.SubstrateSummary.ConfiguredTargets...),
	}
	projection := browserRuntimeDoctorVisibleDefaultRouteProjection(preview, visible)
	source := firstNonEmpty(
		strings.TrimSpace(projection.Metadata.Source),
		strings.TrimSpace(registration.SubstrateSummary.SubstrateSource),
	)
	endpoint := firstNonEmpty(
		strings.TrimSpace(projection.Metadata.Endpoint),
		strings.TrimSpace(registration.SubstrateSummary.SubstrateEndpoint),
	)
	return browserRuntimeRouteDescriptorFromInfoWithProvenance(
		projection.Info,
		source,
		endpoint,
	)
}

func browserRuntimeDefaultCandidateRouteDescriptor(preview browserRuntimeDiagnosticsPreview) browserRuntimeRouteDescriptor {
	if preview.DefaultRouteDescriptor != (browserRuntimeRouteDescriptor{}) {
		return browserRuntimeRouteDescriptor{}
	}
	if route := browserRuntimeRouteDescriptorFromInfoWithProvenance(
		preview.Registration.SubstrateSummary.DefaultCandidateRoute,
		preview.Registration.SubstrateSummary.DefaultCandidateSource,
		preview.Registration.SubstrateSummary.DefaultCandidateEndpoint,
	); route != (browserRuntimeRouteDescriptor{}) {
		return route
	}
	if projection, ok := browserRuntimeDoctorManagedCandidateRouteProjection(preview); ok {
		return browserRuntimeRouteDescriptorFromInfoWithProvenance(
			projection.Info,
			projection.Metadata.Source,
			projection.Metadata.Endpoint,
		)
	}
	return browserRuntimeRouteDescriptor{}
}

func browserRuntimeDiagnosticsPreviewCanonicalRegistration(registration browserDefaultRuntimePreview) browserDefaultRuntimePreview {
	return browserCanonicalizeHiddenManagedRuntimePreview(registration)
}

func browserRuntimeDiagnosticsPreviewForRegistrationPreview(ctx browserRegistrationContext, registration browserDefaultRuntimePreview) browserRuntimeDiagnosticsPreview {
	preview := browserRuntimeDiagnosticsPreviewBaseForRegistrationPreview(registration)
	actions := browserRuntimeDiagnosticsCapabilities(ctx).SupportedRuntimeActions()
	for _, assessment := range browserRuntimeAvailableActionRouteAssessmentsForBackend(
		ctx,
		registration.EffectiveBackend,
		registration.SubstrateAssessment,
		registration.SubstrateSummary,
	) {
		capabilities, ok := browserRuntimeActionCapabilitiesForAssessment(ctx, assessment)
		if !ok {
			continue
		}
		actions = append(actions, capabilities.SupportedRuntimeActions()...)
	}
	preview.AvailableActions = mergeToolMetadataStrings(nil, actions)
	return preview
}

func browserRuntimePopulateSubstrateContextWithPreview(ctx browserRegistrationContext, payload *browserRuntimePayload, preview browserRuntimeDiagnosticsPreview) {
	summary := preview.Registration.SubstrateSummary
	payload.SubstrateSelectionStrategy = summary.SelectionStrategy
	payload.SubstrateSelectionReason = summary.SelectionReason
	payload.SubstrateMatrix = browserRuntimeSubstrateMatrixWithPreview(ctx, preview)
}

func browserRuntimeSubstrateMatrixWithPreview(ctx browserRegistrationContext, preview browserRuntimeDiagnosticsPreview) []browserRuntimeSubstrateStatus {
	defaultRoute := preview.DefaultRoute
	substrate := preview.Registration.SubstrateSummary
	substrateAssessment := preview.Registration.SubstrateAssessment
	backend := preview.Registration.EffectiveBackend
	defaultAssessment := browserRuntimeSubstrateRouteAssessmentForBackend(
		backend,
		BrowserRuntimeInfo{},
		browserRuntimeDefaultSubstrateRouteAssessment(defaultRoute, substrateAssessment),
	)
	matrix := []browserRuntimeSubstrateStatus{
		browserRuntimeSubstrateDefaultRouteStatusForPreview(
			ctx,
			preview,
			"default",
			substrate.SelectionReason,
			defaultRoute,
			defaultAssessment,
		),
	}
	if defaultRoute.Target != "host" || browserRuntimeUsesImplicitLegacyHostDefaultFallback(defaultRoute, substrateAssessment) {
		hostSelectionState := "explicit_fallback"
		hostReason := browserRuntimeSubstrateSelectionReason("host", substrate.HostRoute, defaultRoute)
		hostAssessment := browserRuntimeSubstrateRouteAssessmentForBackend(
			backend,
			BrowserRuntimeInfo{Target: "host"},
			substrateAssessment.HostRoute,
		)
		if !substrate.HostRouteAvailable {
			hostSelectionState = "unsupported"
			hostReason = firstNonEmpty(strings.TrimSpace(substrate.HostFailureCause), hostReason)
		}
		matrix = append(matrix, browserRuntimeSubstrateRouteStatusForAssessmentWithPreview(
			ctx,
			preview,
			"host",
			hostSelectionState,
			hostReason,
			substrate.HostRoute,
			hostAssessment,
		))
	}
	for _, managed := range browserRuntimeManagedRouteAssessmentsForPreview(ctx, preview) {
		if !managed.Configured || defaultRoute.Target == managed.Role {
			continue
		}
		managed.Assessment = browserRuntimeSubstrateRouteAssessmentForBackend(
			backend,
			BrowserRuntimeInfo{Target: managed.Role},
			managed.Assessment,
		)
		matrix = append(matrix, browserRuntimeSubstrateRouteStatusForAssessmentWithPreview(
			ctx,
			preview,
			managed.Role,
			managed.SelectionState(),
			managed.SelectionReason(defaultRoute),
			managed.RuntimeInfo,
			managed.Assessment,
		))
	}
	return matrix
}

func browserRuntimeSubstrateDefaultRouteStatusForPreview(
	ctx browserRegistrationContext,
	preview browserRuntimeDiagnosticsPreview,
	selectionState string,
	selectionReason string,
	info BrowserRuntimeInfo,
	assessment browserConcreteRouteAssessment,
) browserRuntimeSubstrateStatus {
	defaultCandidateRoute := browserRuntimeDefaultCandidateRouteDescriptor(preview)
	if assessment.RouteAvailable {
		status := browserRuntimeSubstrateRouteStatusForAssessmentWithPreview(
			ctx,
			preview,
			"default",
			selectionState,
			selectionReason,
			info,
			assessment,
		)
		status.DefaultCandidateRoute = defaultCandidateRoute
		return status
	}
	info = normalizeBrowserRuntimeInfo(info)
	if browserRuntimeShouldHideImplicitLegacyHostDefaultInfoForDefaultRoute(
		"default",
		info,
		assessment,
		preview.DefaultRoute,
		preview.Registration.SubstrateAssessment,
	) {
		info = BrowserRuntimeInfo{}
	}
	failureSurfaceNote := browserRuntimeAssessmentFailureSurfaceNote(assessment, selectionReason)
	status := browserRuntimeSubstrateStatus{
		Role:                  "default",
		SelectionState:        firstNonEmpty(strings.TrimSpace(selectionState), "unsupported"),
		SelectionReason:       firstNonEmpty(failureSurfaceNote, strings.TrimSpace(selectionReason), strings.TrimSpace(assessment.FailureReason)),
		Profile:               strings.TrimSpace(info.Profile),
		RuntimeTarget:         strings.TrimSpace(info.Target),
		Backend:               strings.TrimSpace(info.Backend),
		DefaultCandidateRoute: defaultCandidateRoute,
		Status:                "unsupported",
		Note:                  firstNonEmpty(failureSurfaceNote, strings.TrimSpace(assessment.FailureNote), strings.TrimSpace(assessment.FailureReason)),
	}
	metadata := browserRuntimeDiagnosticsSurfaceMetadataForRouteWithPreview(ctx, preview, info)
	if browserRuntimeShouldKeepManagedLaunchFailureActionSurface(assessment) {
		if managedMetadata, ok := browserRuntimeManagedLaunchFailureProjectedSurfaceMetadataWithPreview(ctx, preview, info); ok {
			metadata = managedMetadata
		}
	}
	browserRuntimeApplyCapabilityMetadataToSubstrateStatus(&status, metadata)
	return status
}

func browserRuntimeSubstrateDefaultRouteStatus(
	ctx browserRegistrationContext,
	selectionState string,
	selectionReason string,
	info BrowserRuntimeInfo,
	assessment browserConcreteRouteAssessment,
) browserRuntimeSubstrateStatus {
	defaultCandidateRoute := browserRuntimeDefaultCandidateRouteDescriptor(browserRuntimeDiagnosticsPreviewBaseForRegistration(ctx))
	if assessment.RouteAvailable {
		status := browserRuntimeSubstrateRouteStatusForAssessment(
			ctx,
			"default",
			selectionState,
			selectionReason,
			info,
			assessment,
		)
		status.DefaultCandidateRoute = defaultCandidateRoute
		return status
	}
	metadataInfo := normalizeBrowserRuntimeInfo(info)
	info = metadataInfo
	if browserRuntimeShouldHideImplicitLegacyHostDefaultInfo(ctx, "default", info, assessment) {
		info = BrowserRuntimeInfo{}
	}
	failureSurfaceNote := browserRuntimeAssessmentFailureSurfaceNote(assessment, selectionReason)
	status := browserRuntimeSubstrateStatus{
		Role:                  "default",
		SelectionState:        firstNonEmpty(strings.TrimSpace(selectionState), "unsupported"),
		SelectionReason:       firstNonEmpty(failureSurfaceNote, strings.TrimSpace(selectionReason), strings.TrimSpace(assessment.FailureReason)),
		Profile:               strings.TrimSpace(info.Profile),
		RuntimeTarget:         strings.TrimSpace(info.Target),
		Backend:               strings.TrimSpace(info.Backend),
		DefaultCandidateRoute: defaultCandidateRoute,
		Status:                "unsupported",
		Note:                  firstNonEmpty(failureSurfaceNote, strings.TrimSpace(assessment.FailureNote), strings.TrimSpace(assessment.FailureReason)),
	}
	metadata := browserRuntimeDiagnosticsSurfaceMetadata(ctx)
	if browserRuntimeShouldKeepManagedLaunchFailureActionSurface(assessment) {
		if managedMetadata, ok := browserRuntimeManagedLaunchFailureProjectedSurfaceMetadata(ctx, info); ok {
			metadata = managedMetadata
		}
	} else if metadataInfo != (BrowserRuntimeInfo{}) &&
		!browserRuntimeShouldHideImplicitLegacyHostDefaultInfo(ctx, "default", metadataInfo, assessment) {
		metadata = browserRuntimeDiagnosticsSurfaceMetadataForRoute(ctx, metadataInfo)
	}
	browserRuntimeApplyCapabilityMetadataToSubstrateStatus(&status, metadata)
	return status
}
