package tools

import "strings"

type browserRuntimeTopLevelRouteProjection struct {
	DefaultRoute          browserRuntimeRouteDescriptor
	DefaultCandidateRoute browserRuntimeRouteDescriptor
	SelectedRoute         *browserRuntimeRouteDescriptor
	ConfiguredTargets     []string
	Metadata              browserRuntimeCapabilityMetadata
}

type browserRuntimeAssessmentSurfaceProjection struct {
	Info        BrowserRuntimeInfo
	Metadata    browserRuntimeCapabilityMetadata
	HasMetadata bool
}

func browserRuntimeDiagnosticsSurfaceMetadata(ctx browserRegistrationContext) browserRuntimeCapabilityMetadata {
	return browserRuntimeDiagnosticsMetadata(ctx)
}

func browserRuntimeDiagnosticsSurfaceMetadataForRoute(
	ctx browserRegistrationContext,
	route BrowserRuntimeInfo,
) browserRuntimeCapabilityMetadata {
	optInSurface := browserRuntimeManagedOptInDiagnosticsSurfaceForRoute(ctx, route)
	if browserRuntimeManagedOptInSurfaceLabel(optInSurface) != "" {
		return browserRuntimeCapabilityMetadataForDiagnosticsSurface(
			ctx,
			browserRuntimeDiagnosticsCapabilities(ctx),
			optInSurface,
		)
	}
	if metadata, ok := browserRuntimeDegradedRouteSurfaceMetadata(ctx, route); ok {
		return metadata
	}
	return browserRuntimeDiagnosticsSurfaceMetadata(ctx)
}

func browserRuntimeManagedLaunchFailureConcreteCapabilities(
	ctx browserRegistrationContext,
	runtimeInfo BrowserRuntimeInfo,
) (BrowserCapabilities, bool) {
	runtimeInfo = normalizeBrowserRuntimeInfo(runtimeInfo)
	var backend BrowserBackend
	switch runtimeInfo.Target {
	case "node":
		backend = ctx.opts.NodeBackend
	case "sandbox":
		backend = ctx.opts.SandboxBackend
	default:
		return BrowserCapabilities{}, false
	}
	if backend == nil {
		return BrowserCapabilities{}, false
	}
	capabilities := browserCapabilitiesForConcreteBackend(backend)
	if capabilities == (BrowserCapabilities{}) {
		return BrowserCapabilities{}, false
	}
	return capabilities, true
}

func browserRuntimeManagedLaunchFailureSurfaceMetadata(
	ctx browserRegistrationContext,
	route BrowserRuntimeInfo,
) (browserRuntimeCapabilityMetadata, bool) {
	capabilities, ok := browserRuntimeManagedLaunchFailureConcreteCapabilities(ctx, route)
	if !ok {
		return browserRuntimeCapabilityMetadata{}, false
	}
	optInSurface := browserRuntimeManagedOptInSurfaceForResolvedRoute(ctx, route, capabilities)
	return browserRuntimeCapabilityMetadataForDiagnosticsSurface(
		ctx,
		browserRuntimeManagedLaunchFailureActionCapabilities(ctx),
		optInSurface,
	), true
}

func browserRuntimeDiagnosticsSurfaceMetadataWithPreview(
	ctx browserRegistrationContext,
	preview browserRuntimeDiagnosticsPreview,
) browserRuntimeCapabilityMetadata {
	return browserRuntimeDiagnosticsSurfaceMetadataForRouteWithPreview(ctx, preview, preview.DefaultRoute)
}

func browserRuntimeDiagnosticsSurfaceMetadataForRouteWithPreview(
	ctx browserRegistrationContext,
	preview browserRuntimeDiagnosticsPreview,
	route BrowserRuntimeInfo,
) browserRuntimeCapabilityMetadata {
	route = browserRuntimeDiagnosticsMetadataRouteForPreview(preview, route)
	optInSurface := browserRuntimeManagedOptInDiagnosticsSurfaceForRouteWithPreview(ctx, preview, route)
	if browserRuntimeManagedOptInSurfaceLabel(optInSurface) != "" {
		return browserRuntimeCapabilityMetadataForDiagnosticsSurface(
			ctx,
			browserRuntimeDiagnosticsCapabilities(ctx),
			optInSurface,
		)
	}
	if metadata, ok := browserRuntimeDegradedRouteSurfaceMetadataWithPreview(ctx, preview, route); ok {
		return metadata
	}
	return browserRuntimeCapabilityMetadataForDiagnosticsSurface(
		ctx,
		browserRuntimeDiagnosticsCapabilities(ctx),
		browserRuntimeManagedOptInDiagnosticsSurface{},
	)
}

func browserRuntimeDiagnosticsPreferredDefaultMetadataRouteForPreview(
	preview browserRuntimeDiagnosticsPreview,
) BrowserRuntimeInfo {
	if info := browserVisibleDefaultRuntimeInfoForPreview(preview.Registration); info != (BrowserRuntimeInfo{}) {
		return info
	}
	if info := normalizeBrowserRuntimeInfo(preview.Registration.SubstrateSummary.DefaultCandidateRoute); info != (BrowserRuntimeInfo{}) {
		return info
	}
	if projection, ok := browserRuntimeDoctorManagedCandidateRouteProjection(preview); ok {
		return projection.Info
	}
	return BrowserRuntimeInfo{}
}

func browserRuntimeDiagnosticsMetadataRouteForPreview(
	preview browserRuntimeDiagnosticsPreview,
	route BrowserRuntimeInfo,
) BrowserRuntimeInfo {
	route = normalizeBrowserRuntimeInfo(route)
	defaultRoute := normalizeBrowserRuntimeInfo(preview.DefaultRoute)
	preferred := browserRuntimeDiagnosticsPreferredDefaultMetadataRouteForPreview(preview)
	if preferred != (BrowserRuntimeInfo{}) &&
		(route == (BrowserRuntimeInfo{}) ||
			((preview.Registration.HiddenImplicitHostDefaultBase ||
				browserRuntimeUsesImplicitLegacyHostDefaultFallback(defaultRoute, preview.Registration.SubstrateAssessment)) &&
				route == defaultRoute)) {
		return preferred
	}
	if route != (BrowserRuntimeInfo{}) {
		return route
	}
	if preferred != (BrowserRuntimeInfo{}) {
		return preferred
	}
	return defaultRoute
}

func browserRuntimeManagedLaunchFailureSurfaceMetadataWithPreview(
	ctx browserRegistrationContext,
	preview browserRuntimeDiagnosticsPreview,
	route BrowserRuntimeInfo,
) (browserRuntimeCapabilityMetadata, bool) {
	capabilities, ok := browserRuntimeManagedLaunchFailureConcreteCapabilities(ctx, route)
	if !ok {
		return browserRuntimeCapabilityMetadata{}, false
	}
	optInSurface := browserRuntimeManagedOptInSurfaceForResolvedRouteWithPreview(ctx, preview, route, capabilities)
	return browserRuntimeCapabilityMetadataForDiagnosticsSurface(
		ctx,
		browserRuntimeManagedLaunchFailureActionCapabilities(ctx),
		optInSurface,
	), true
}

func browserRuntimeManagedLaunchFailureProjectedSurfaceMetadata(
	ctx browserRegistrationContext,
	route BrowserRuntimeInfo,
) (browserRuntimeCapabilityMetadata, bool) {
	if metadata, ok := browserRuntimeManagedLaunchFailureSurfaceMetadata(ctx, route); ok {
		return metadata, true
	}
	optInSurface := browserRuntimeManagedOptInDiagnosticsSurfaceForRoute(ctx, route)
	if browserRuntimeManagedOptInSurfaceLabel(optInSurface) == "" {
		return browserRuntimeCapabilityMetadata{}, false
	}
	return browserRuntimeCapabilityMetadataForDiagnosticsSurface(
		ctx,
		browserRuntimeManagedLaunchFailureActionCapabilities(ctx),
		optInSurface,
	), true
}

func browserRuntimeManagedLaunchFailureProjectedSurfaceMetadataWithPreview(
	ctx browserRegistrationContext,
	preview browserRuntimeDiagnosticsPreview,
	route BrowserRuntimeInfo,
) (browserRuntimeCapabilityMetadata, bool) {
	if metadata, ok := browserRuntimeManagedLaunchFailureSurfaceMetadataWithPreview(ctx, preview, route); ok {
		return metadata, true
	}
	optInSurface := browserRuntimeManagedOptInDiagnosticsSurfaceForRouteWithPreview(ctx, preview, route)
	if browserRuntimeManagedOptInSurfaceLabel(optInSurface) == "" {
		return browserRuntimeCapabilityMetadata{}, false
	}
	return browserRuntimeCapabilityMetadataForDiagnosticsSurface(
		ctx,
		browserRuntimeManagedLaunchFailureActionCapabilities(ctx),
		optInSurface,
	), true
}

func browserRuntimeResolvedRouteSurfaceMetadata(
	ctx browserRegistrationContext,
	route browserResolvedExecutionRoute,
) (browserRuntimeCapabilityMetadata, BrowserCapabilities) {
	capabilities := browserCapabilitiesForRuntimeInspection(ctx, route)
	if ctx.derivedCache != nil {
		return ctx.derivedCache.resolvedRouteSurfaceMetadata(ctx, route)
	}
	return browserRuntimeResolvedRouteSurfaceMetadataUncached(ctx, route, capabilities), capabilities
}

func browserRuntimeResolvedRouteSurfaceMetadataWithPreview(
	ctx browserRegistrationContext,
	preview browserRuntimeDiagnosticsPreview,
	route browserResolvedExecutionRoute,
) (browserRuntimeCapabilityMetadata, BrowserCapabilities) {
	capabilities := browserCapabilitiesForRuntimeInspection(ctx, route)
	metadata := browserRuntimeCapabilityMetadataForCapabilities(ctx, capabilities)
	optInSurface := browserRuntimeManagedOptInSurfaceForResolvedRouteWithPreview(ctx, preview, route.RuntimeInfo, capabilities)
	metadata.BrowserSurface = browserRuntimeManagedOptInSurfaceLabel(optInSurface)
	metadata.BrowserOptInTargets = append([]string(nil), optInSurface.Targets...)
	return metadata, capabilities
}

func browserRuntimeResolvedRouteSurfaceMetadataUncached(
	ctx browserRegistrationContext,
	route browserResolvedExecutionRoute,
	capabilities BrowserCapabilities,
) browserRuntimeCapabilityMetadata {
	metadata := browserRuntimeCapabilityMetadataForCapabilities(ctx, capabilities)
	optInSurface := browserRuntimeManagedOptInSurfaceForResolvedRoute(ctx, route.RuntimeInfo, capabilities)
	metadata.BrowserSurface = browserRuntimeManagedOptInSurfaceLabel(optInSurface)
	metadata.BrowserOptInTargets = append([]string(nil), optInSurface.Targets...)
	return metadata
}

func browserRuntimeDiagnosticsRouteProjection(
	ctx browserRegistrationContext,
	preview browserRuntimeDiagnosticsPreview,
) browserRuntimeTopLevelRouteProjection {
	return browserRuntimeTopLevelRouteProjection{
		DefaultRoute:          preview.DefaultRouteDescriptor,
		DefaultCandidateRoute: browserRuntimeDefaultCandidateRouteDescriptor(preview),
		ConfiguredTargets:     append([]string(nil), preview.ConfiguredTargets...),
		Metadata:              browserRuntimeDiagnosticsSurfaceMetadataForRouteWithPreview(ctx, preview, preview.DefaultRoute),
	}
}

func browserRuntimeDegradedRouteSurfaceMetadata(
	ctx browserRegistrationContext,
	route BrowserRuntimeInfo,
) (browserRuntimeCapabilityMetadata, bool) {
	route = normalizeBrowserRuntimeInfo(route)
	if route == (BrowserRuntimeInfo{}) {
		return browserRuntimeCapabilityMetadata{}, false
	}
	optInSurface := browserRuntimeManagedOptInDiagnosticsSurfaceForRoute(ctx, route)
	if browserRuntimeManagedOptInSurfaceLabel(optInSurface) == "" {
		return browserRuntimeCapabilityMetadata{}, false
	}
	return browserRuntimeCapabilityMetadataForDiagnosticsSurface(
		ctx,
		browserRuntimeDiagnosticsCapabilities(ctx),
		optInSurface,
	), true
}

func browserRuntimeDegradedRouteSurfaceMetadataWithPreview(
	ctx browserRegistrationContext,
	preview browserRuntimeDiagnosticsPreview,
	route BrowserRuntimeInfo,
) (browserRuntimeCapabilityMetadata, bool) {
	route = normalizeBrowserRuntimeInfo(route)
	if route == (BrowserRuntimeInfo{}) {
		return browserRuntimeCapabilityMetadata{}, false
	}
	optInSurface := browserRuntimeManagedOptInDiagnosticsSurfaceForRouteWithPreview(ctx, preview, route)
	if browserRuntimeManagedOptInSurfaceLabel(optInSurface) == "" {
		return browserRuntimeCapabilityMetadata{}, false
	}
	return browserRuntimeCapabilityMetadataForDiagnosticsSurface(
		ctx,
		browserRuntimeDiagnosticsCapabilities(ctx),
		optInSurface,
	), true
}

func browserRuntimeSelectedRouteProjection(
	ctx browserRegistrationContext,
	preview browserRuntimeDiagnosticsPreview,
	route browserResolvedExecutionRoute,
) (browserRuntimeTopLevelRouteProjection, BrowserCapabilities) {
	metadata, capabilities := browserRuntimeResolvedRouteSurfaceMetadataWithPreview(ctx, preview, route)
	return browserRuntimeTopLevelRouteProjection{
		DefaultRoute:          preview.DefaultRouteDescriptor,
		DefaultCandidateRoute: browserRuntimeDefaultCandidateRouteDescriptor(preview),
		SelectedRoute:         browserRuntimeRouteDescriptorPtrFromResolvedRoute(route),
		ConfiguredTargets:     append([]string(nil), preview.ConfiguredTargets...),
		Metadata:              metadata,
	}, capabilities
}

func browserRuntimeApplyTopLevelRouteProjection(
	payload *browserRuntimePayload,
	projection browserRuntimeTopLevelRouteProjection,
) {
	if payload == nil {
		return
	}
	payload.DefaultRoute = projection.DefaultRoute
	payload.DefaultCandidateRoute = projection.DefaultCandidateRoute
	if projection.SelectedRoute != nil {
		selected := *projection.SelectedRoute
		payload.SelectedRoute = &selected
	} else {
		payload.SelectedRoute = nil
	}
	payload.ConfiguredTargets = append([]string(nil), projection.ConfiguredTargets...)
	browserRuntimeApplyCapabilityMetadataToPayload(payload, projection.Metadata)
	browserRuntimeApplyDefaultCandidateRouteToPayloadShells(payload)
}

func browserRuntimeApplyCapabilityMetadataToRouteStatus(routeStatus *browserRuntimeRouteStatus, metadata browserRuntimeCapabilityMetadata) {
	if routeStatus == nil {
		return
	}
	routeStatus.RuntimeActions = metadata.RuntimeActions
	routeStatus.BrowserTools = metadata.BrowserTools
	routeStatus.ArtifactTools = metadata.ArtifactTools
	routeStatus.ArtifactKinds = metadata.ArtifactKinds
	routeStatus.ArtifactContract = metadata.ArtifactContract
	routeStatus.BrowserActKinds = metadata.BrowserActKinds
	routeStatus.BrowserSurface = strings.TrimSpace(metadata.BrowserSurface)
	routeStatus.BrowserOptInTargets = append([]string(nil), metadata.BrowserOptInTargets...)
	routeStatus.Capabilities = metadata.Capabilities
}

func browserRuntimeApplyCapabilityMetadataToSubstrateStatus(substrateStatus *browserRuntimeSubstrateStatus, metadata browserRuntimeCapabilityMetadata) {
	if substrateStatus == nil {
		return
	}
	substrateStatus.RuntimeActions = metadata.RuntimeActions
	substrateStatus.BrowserTools = metadata.BrowserTools
	substrateStatus.ArtifactTools = metadata.ArtifactTools
	substrateStatus.ArtifactKinds = metadata.ArtifactKinds
	substrateStatus.ArtifactContract = metadata.ArtifactContract
	substrateStatus.BrowserActKinds = metadata.BrowserActKinds
	substrateStatus.BrowserSurface = strings.TrimSpace(metadata.BrowserSurface)
	substrateStatus.BrowserOptInTargets = append([]string(nil), metadata.BrowserOptInTargets...)
	substrateStatus.Capabilities = metadata.Capabilities
}

func browserRuntimeStatusRouteMetadataForAssessment(
	ctx browserRegistrationContext,
	info BrowserRuntimeInfo,
	assessment browserConcreteRouteAssessment,
) browserDoctorRouteMetadata {
	return browserRuntimeStatusRouteMetadataForAssessmentWithPreview(
		ctx,
		browserRuntimeDiagnosticsPreviewBaseForRegistration(ctx),
		info,
		assessment,
	)
}

func browserRuntimeStatusRouteMetadataForAssessmentWithPreview(
	ctx browserRegistrationContext,
	preview browserRuntimeDiagnosticsPreview,
	info BrowserRuntimeInfo,
	assessment browserConcreteRouteAssessment,
) browserDoctorRouteMetadata {
	if assessment.RouteAvailable {
		if metadata := browserRuntimeStatusRouteMetadataForResolvedRoute(assessment.Route); metadata != (browserDoctorRouteMetadata{}) {
			return metadata
		}
	}
	info = normalizeBrowserRuntimeInfo(info)
	if info == (BrowserRuntimeInfo{}) {
		return browserDoctorRouteMetadata{}
	}
	if metadata := browserRuntimeDoctorRouteMetadataForAssessment(
		assessment,
		browserRuntimeStatusRouteMetadataFallbackBackend(ctx, preview, info.Target),
	); metadata != (browserDoctorRouteMetadata{}) {
		return metadata
	}
	return browserDoctorRouteMetadata{}
}

func browserRuntimeStatusRouteMetadataForResolvedRoute(route browserResolvedExecutionRoute) browserDoctorRouteMetadata {
	if source := strings.TrimSpace(route.Source); source != "" || strings.TrimSpace(route.Endpoint) != "" {
		return browserDoctorRouteMetadata{
			Source:   source,
			Endpoint: strings.TrimSpace(route.Endpoint),
		}
	}
	return browserDoctorRouteMetadataForBackend(route.Backend)
}

func browserRuntimeStatusRouteMetadataFallbackBackend(
	ctx browserRegistrationContext,
	preview browserRuntimeDiagnosticsPreview,
	target string,
) BrowserBackend {
	switch strings.ToLower(strings.TrimSpace(target)) {
	case "node":
		if backend := browserRuntimePreviewManagedTargetBackend(preview, target); backend != nil {
			return backend
		}
		return ctx.opts.NodeBackend
	case "sandbox":
		if backend := browserRuntimePreviewManagedTargetBackend(preview, target); backend != nil {
			return backend
		}
		return ctx.opts.SandboxBackend
	case "host":
		return browserConfiguredHostBackendForOptions(ctx.opts)
	default:
		return nil
	}
}

func browserRuntimeRouteAssessmentSurfaceProjection(
	ctx browserRegistrationContext,
	info BrowserRuntimeInfo,
	assessment browserConcreteRouteAssessment,
) browserRuntimeAssessmentSurfaceProjection {
	return browserRuntimeAssessmentSurfaceProjectionForRole(ctx, "default", info, assessment, false)
}

func browserRuntimeRouteAssessmentSurfaceProjectionForPreview(
	ctx browserRegistrationContext,
	preview browserRuntimeDiagnosticsPreview,
	info BrowserRuntimeInfo,
	assessment browserConcreteRouteAssessment,
) browserRuntimeAssessmentSurfaceProjection {
	info = normalizeBrowserRuntimeInfo(info)
	if !assessment.RouteAvailable {
		defaultRoute := normalizeBrowserRuntimeInfo(preview.DefaultRoute)
		preserveDefaultMetadataWhenHidden := info == defaultRoute
		if browserRuntimeShouldHideImplicitLegacyHostDefaultInfoForDefaultRoute(
			"default",
			info,
			assessment,
			defaultRoute,
			preview.Registration.SubstrateAssessment,
		) {
			info = BrowserRuntimeInfo{}
		}
		projection := browserRuntimeAssessmentSurfaceProjection{Info: info}
		if browserRuntimeShouldKeepManagedLaunchFailureActionSurface(assessment) {
			if metadata, ok := browserRuntimeManagedLaunchFailureProjectedSurfaceMetadataWithPreview(ctx, preview, info); ok {
				projection.Metadata = metadata
				projection.HasMetadata = true
			}
		} else if preserveDefaultMetadataWhenHidden || info == defaultRoute {
			projection.Metadata = browserRuntimeDiagnosticsSurfaceMetadataForRouteWithPreview(ctx, preview, info)
			projection.HasMetadata = true
		}
		return projection
	}
	metadata, _ := browserRuntimeResolvedRouteSurfaceMetadataWithPreview(ctx, preview, assessment.Route)
	return browserRuntimeAssessmentSurfaceProjection{
		Info:        normalizeBrowserRuntimeInfo(assessment.Route.RuntimeInfo),
		Metadata:    metadata,
		HasMetadata: true,
	}
}

func browserRuntimeSubstrateAssessmentSurfaceProjection(
	ctx browserRegistrationContext,
	role string,
	info BrowserRuntimeInfo,
	assessment browserConcreteRouteAssessment,
) browserRuntimeAssessmentSurfaceProjection {
	return browserRuntimeAssessmentSurfaceProjectionForRole(ctx, role, info, assessment, true)
}

func browserRuntimeAssessmentSurfaceProjectionForRole(
	ctx browserRegistrationContext,
	role string,
	info BrowserRuntimeInfo,
	assessment browserConcreteRouteAssessment,
	preserveDefaultMetadataWhenHidden bool,
) browserRuntimeAssessmentSurfaceProjection {
	if ctx.derivedCache != nil {
		return ctx.derivedCache.assessmentSurfaceProjection(ctx, role, info, assessment, preserveDefaultMetadataWhenHidden)
	}
	capabilities := BrowserCapabilities{}
	if assessment.RouteAvailable {
		capabilities = browserCapabilitiesForRuntimeInspection(ctx, assessment.Route)
	}
	return browserRuntimeAssessmentSurfaceProjectionUncached(ctx, role, info, assessment, capabilities, preserveDefaultMetadataWhenHidden)
}

func browserRuntimeAssessmentSurfaceProjectionUncached(
	ctx browserRegistrationContext,
	role string,
	info BrowserRuntimeInfo,
	assessment browserConcreteRouteAssessment,
	capabilities BrowserCapabilities,
	preserveDefaultMetadataWhenHidden bool,
) browserRuntimeAssessmentSurfaceProjection {
	info = normalizeBrowserRuntimeInfo(info)
	if !assessment.RouteAvailable {
		if browserRuntimeShouldHideImplicitLegacyHostDefaultInfo(ctx, role, info, assessment) {
			info = BrowserRuntimeInfo{}
		}
		projection := browserRuntimeAssessmentSurfaceProjection{Info: info}
		if browserRuntimeShouldKeepManagedLaunchFailureActionSurface(assessment) {
			if metadata, ok := browserRuntimeManagedLaunchFailureProjectedSurfaceMetadata(ctx, info); ok {
				projection.Metadata = metadata
				projection.HasMetadata = true
			}
		} else if strings.EqualFold(strings.TrimSpace(role), "default") &&
			(preserveDefaultMetadataWhenHidden || info == normalizeBrowserRuntimeInfo(browserRegistrationDefaultRuntimeInfo(ctx))) {
			projection.Metadata = browserRuntimeDiagnosticsSurfaceMetadataForRoute(ctx, info)
			projection.HasMetadata = true
		}
		return projection
	}
	metadata := browserRuntimeResolvedRouteSurfaceMetadataUncached(ctx, assessment.Route, capabilities)
	return browserRuntimeAssessmentSurfaceProjection{
		Info:        normalizeBrowserRuntimeInfo(assessment.Route.RuntimeInfo),
		Metadata:    metadata,
		HasMetadata: true,
	}
}

func browserRuntimeSubstrateAssessmentSurfaceProjectionForPreview(
	ctx browserRegistrationContext,
	role string,
	preview browserRuntimeDiagnosticsPreview,
	info BrowserRuntimeInfo,
	assessment browserConcreteRouteAssessment,
) browserRuntimeAssessmentSurfaceProjection {
	info = normalizeBrowserRuntimeInfo(info)
	if !assessment.RouteAvailable {
		if browserRuntimeShouldHideImplicitLegacyHostDefaultInfoForDefaultRoute(
			role,
			info,
			assessment,
			preview.DefaultRoute,
			preview.Registration.SubstrateAssessment,
		) {
			info = BrowserRuntimeInfo{}
		}
		projection := browserRuntimeAssessmentSurfaceProjection{Info: info}
		if browserRuntimeShouldKeepManagedLaunchFailureActionSurface(assessment) {
			if metadata, ok := browserRuntimeManagedLaunchFailureProjectedSurfaceMetadataWithPreview(ctx, preview, info); ok {
				projection.Metadata = metadata
				projection.HasMetadata = true
			}
		} else if strings.EqualFold(strings.TrimSpace(role), "default") {
			projection.Metadata = browserRuntimeDiagnosticsSurfaceMetadataForRouteWithPreview(ctx, preview, info)
			projection.HasMetadata = true
		}
		return projection
	}
	metadata, _ := browserRuntimeResolvedRouteSurfaceMetadataWithPreview(ctx, preview, assessment.Route)
	return browserRuntimeAssessmentSurfaceProjection{
		Info:        normalizeBrowserRuntimeInfo(assessment.Route.RuntimeInfo),
		Metadata:    metadata,
		HasMetadata: true,
	}
}
