package tools

import (
	"strings"

	llmxtools "github.com/wsnacj/agentx-go/tools"
)

type browserRuntimeManagedOptInDiagnosticsSurface struct {
	ToolNames    []string
	ActKinds     []string
	Targets      []string
	Capabilities BrowserCapabilities
}

func browserRuntimeManagedOptInSurfaceLabel(surface browserRuntimeManagedOptInDiagnosticsSurface) string {
	if len(surface.Targets) == 0 {
		return ""
	}
	return "explicit_managed_opt_in"
}

func browserRuntimeRegisteredToolSet(ctx browserRegistrationContext) map[string]bool {
	if ctx.derivedCache != nil {
		return ctx.derivedCache.registeredToolSet(ctx.reg)
	}
	return browserRuntimeRegisteredToolSetFromRegistry(ctx.reg)
}

func browserRuntimeRegisteredToolSetFromRegistry(reg *llmxtools.Registry) map[string]bool {
	if reg == nil {
		return nil
	}
	defs := reg.Definitions()
	if len(defs) == 0 {
		return nil
	}
	out := map[string]bool{}
	for _, def := range defs {
		name := NormalizeToolName(def.Function.Name)
		if !isBrowserToolName(name) {
			continue
		}
		out[name] = true
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func browserRuntimeManagedOptInCapabilities(ctx browserRegistrationContext) browserRuntimeManagedOptInDiagnosticsSurface {
	if ctx.derivedCache != nil {
		return ctx.derivedCache.managedOptInDiagnosticsSurface(ctx)
	}
	return browserRuntimeManagedOptInCapabilitiesUncached(ctx)
}

func browserRuntimeManagedOptInCapabilitiesUncached(ctx browserRegistrationContext) browserRuntimeManagedOptInDiagnosticsSurface {
	registered := browserRuntimeRegisteredToolSet(ctx)
	if len(registered) == 0 {
		return browserRuntimeManagedOptInDiagnosticsSurface{}
	}
	return browserManagedOptInProjectionForCapabilities(
		registered,
		ctx.capabilities,
		ctx.opts.NodeBackend,
		ctx.opts.SandboxBackend,
	).DiagnosticsSurface(registered)
}

func browserRuntimeManagedOptInSurfaceForTarget(
	ctx browserRegistrationContext,
	runtimeTarget string,
) browserRuntimeManagedOptInDiagnosticsSurface {
	return browserRuntimeManagedOptInSurfaceForRoute(
		ctx,
		BrowserRuntimeInfo{Target: runtimeTarget},
	)
}

func browserRuntimeManagedOptInSurfaceForRoute(
	ctx browserRegistrationContext,
	runtimeInfo BrowserRuntimeInfo,
) browserRuntimeManagedOptInDiagnosticsSurface {
	runtimeInfo = normalizeBrowserRuntimeInfo(runtimeInfo)
	if ctx.derivedCache != nil {
		return ctx.derivedCache.managedOptInSurfaceForRoute(ctx, runtimeInfo)
	}
	return browserRuntimeManagedOptInSurfaceForRouteUncached(ctx, runtimeInfo)
}

func browserRuntimeManagedOptInSurfaceForRouteUncached(
	ctx browserRegistrationContext,
	runtimeInfo BrowserRuntimeInfo,
) browserRuntimeManagedOptInDiagnosticsSurface {
	runtimeInfo = normalizeBrowserRuntimeInfo(runtimeInfo)
	if runtimeInfo.Target == "" {
		return browserRuntimeManagedOptInDiagnosticsSurface{}
	}
	capabilities, ok := browserRuntimeManagedOptInCapabilitiesForRoute(ctx, runtimeInfo)
	if !ok {
		return browserRuntimeManagedOptInDiagnosticsSurface{}
	}
	return browserRuntimeManagedOptInSurfaceForResolvedRoute(ctx, runtimeInfo, capabilities)
}

func browserRuntimeManagedOptInSurfaceForRouteWithPreview(
	ctx browserRegistrationContext,
	preview browserRuntimeDiagnosticsPreview,
	runtimeInfo BrowserRuntimeInfo,
) browserRuntimeManagedOptInDiagnosticsSurface {
	runtimeInfo = normalizeBrowserRuntimeInfo(runtimeInfo)
	if runtimeInfo.Target == "" {
		return browserRuntimeManagedOptInDiagnosticsSurface{}
	}
	if !browserRuntimePreviewConfiguresManagedTarget(preview, runtimeInfo.Target) {
		return browserRuntimeManagedOptInDiagnosticsSurface{}
	}
	capabilities, ok := browserRuntimeManagedOptInCapabilitiesForRouteWithPreview(ctx, preview, runtimeInfo)
	if !ok {
		return browserRuntimeManagedOptInDiagnosticsSurface{}
	}
	return browserRuntimeManagedOptInSurfaceForResolvedRouteWithPreview(
		ctx,
		preview,
		runtimeInfo,
		capabilities,
	)
}

func browserRuntimeManagedOptInCapabilitiesForRoute(
	ctx browserRegistrationContext,
	runtimeInfo BrowserRuntimeInfo,
) (BrowserCapabilities, bool) {
	runtimeInfo = normalizeBrowserRuntimeInfo(runtimeInfo)
	requested := runtimeInfo
	var (
		backend  BrowserBackend
		fallback BrowserRuntimeInfo
	)
	switch runtimeInfo.Target {
	case "node":
		backend = ctx.opts.NodeBackend
		fallback = defaultBrowserNodeRuntimeInfo()
	case "sandbox":
		backend = ctx.opts.SandboxBackend
		fallback = defaultBrowserSandboxRuntimeInfo()
	default:
		return BrowserCapabilities{}, false
	}
	if backend == nil {
		return BrowserCapabilities{}, false
	}
	// Managed opt-in diagnostics should resolve against the destination lane's
	// concrete backend identity. Carrying the source host backend through a
	// host->node/sandbox probe can bypass execution-route narrowing entirely.
	requested.Backend = ""
	route, err := resolveConcreteBrowserExecutionRoute(backend, fallback, requested)
	if err != nil {
		return BrowserCapabilities{}, false
	}
	capabilities := route.Capabilities
	if capabilities == (BrowserCapabilities{}) {
		capabilities = browserCapabilitiesForConcreteBackend(route.Backend)
	}
	if !capabilities.SupportsAnyActKind() {
		return BrowserCapabilities{}, false
	}
	return capabilities, true
}

func browserRuntimeManagedOptInCapabilitiesForRouteWithPreview(
	ctx browserRegistrationContext,
	preview browserRuntimeDiagnosticsPreview,
	runtimeInfo BrowserRuntimeInfo,
) (BrowserCapabilities, bool) {
	runtimeInfo = normalizeBrowserRuntimeInfo(runtimeInfo)
	if !browserRuntimePreviewConfiguresManagedTarget(preview, runtimeInfo.Target) {
		return BrowserCapabilities{}, false
	}
	if assessment, ok := browserRuntimePreviewManagedRouteAssessment(preview, runtimeInfo.Target); ok {
		if !assessment.RouteAvailable {
			return browserRuntimeManagedOptInCapabilitiesForRoute(ctx, runtimeInfo)
		}
		capabilities := assessment.Route.Capabilities
		if capabilities == (BrowserCapabilities{}) {
			capabilities = browserCapabilitiesForConcreteBackend(assessment.Route.Backend)
		}
		if !capabilities.SupportsAnyActKind() {
			return BrowserCapabilities{}, false
		}
		return capabilities, true
	}
	return browserRuntimeManagedOptInCapabilitiesForRoute(ctx, runtimeInfo)
}

func browserRuntimePreviewManagedRouteAssessment(preview browserRuntimeDiagnosticsPreview, runtimeTarget string) (browserConcreteRouteAssessment, bool) {
	switch strings.ToLower(strings.TrimSpace(runtimeTarget)) {
	case "node":
		assessment := browserConcreteRouteAssessmentForDefaultPromotion(preview.Registration.SubstrateAssessment.NodeRoute)
		return assessment, browserConcreteRouteAssessmentHasResult(assessment)
	case "sandbox":
		assessment := preview.Registration.SubstrateAssessment.SandboxConcreteRoute
		return assessment, browserConcreteRouteAssessmentHasResult(assessment)
	default:
		return browserConcreteRouteAssessment{}, false
	}
}

func browserRuntimeManagedOptInSurfaceForResolvedRoute(
	ctx browserRegistrationContext,
	runtimeInfo BrowserRuntimeInfo,
	capabilities BrowserCapabilities,
) browserRuntimeManagedOptInDiagnosticsSurface {
	if ctx.derivedCache != nil {
		return ctx.derivedCache.managedOptInSurfaceForResolvedRoute(ctx, runtimeInfo, capabilities)
	}
	return browserRuntimeManagedOptInSurfaceForResolvedRouteUncached(ctx, runtimeInfo, capabilities)
}

func browserRuntimeManagedOptInSurfaceForResolvedRouteUncached(
	ctx browserRegistrationContext,
	runtimeInfo BrowserRuntimeInfo,
	capabilities BrowserCapabilities,
) browserRuntimeManagedOptInDiagnosticsSurface {
	return browserRuntimeManagedOptInSurfaceForResolvedRouteWithPreview(
		ctx,
		browserRuntimeDiagnosticsPreviewBaseForRegistration(ctx),
		runtimeInfo,
		capabilities,
	)
}

func browserRuntimeManagedOptInSurfaceForResolvedRouteWithPreview(
	ctx browserRegistrationContext,
	preview browserRuntimeDiagnosticsPreview,
	runtimeInfo BrowserRuntimeInfo,
	capabilities BrowserCapabilities,
) browserRuntimeManagedOptInDiagnosticsSurface {
	runtimeInfo = normalizeBrowserRuntimeInfo(runtimeInfo)
	if runtimeInfo.Target != "node" && runtimeInfo.Target != "sandbox" {
		return browserRuntimeManagedOptInDiagnosticsSurface{}
	}
	visibleDefaultRoute := browserVisibleDefaultRuntimeInfoForPreview(preview.Registration)
	if visibleDefaultRoute != (BrowserRuntimeInfo{}) && visibleDefaultRoute.Target == runtimeInfo.Target {
		return browserRuntimeManagedOptInDiagnosticsSurface{}
	}
	if visibleDefaultRoute == (BrowserRuntimeInfo{}) {
		defaultCandidateRoute := browserRuntimeDiagnosticsPreferredDefaultMetadataRouteForPreview(preview)
		if defaultCandidateRoute == (BrowserRuntimeInfo{}) || defaultCandidateRoute.Target != runtimeInfo.Target {
			return browserRuntimeManagedOptInDiagnosticsSurface{}
		}
	}
	if visibleDefaultRoute == (BrowserRuntimeInfo{}) && !browserRuntimePreviewConfiguresManagedTarget(preview, runtimeInfo.Target) {
		return browserRuntimeManagedOptInDiagnosticsSurface{}
	}
	selectedSurface := browserRuntimeRegisteredSurfaceForCapabilities(ctx, capabilities)
	defaultCapabilities := BrowserCapabilities{}
	defaultAssessment := browserRuntimeDefaultSubstrateRouteAssessment(
		preview.DefaultRoute,
		preview.Registration.SubstrateAssessment,
	)
	if defaultAssessment.RouteAvailable {
		defaultCapabilities = browserCapabilitiesForRuntimeInspection(ctx, defaultAssessment.Route)
	}
	defaultSurface := browserRuntimeRegisteredSurfaceForCapabilities(ctx, defaultCapabilities)
	extraCompatTools := browserRuntimeSelectedOnlyCompatTools(selectedSurface.BrowserTools, defaultSurface.BrowserTools)
	extraActKinds := browserRuntimeSelectedOnlyItems(selectedSurface.BrowserActKinds, defaultSurface.BrowserActKinds)
	if len(extraCompatTools) == 0 && len(extraActKinds) == 0 {
		return browserRuntimeManagedOptInDiagnosticsSurface{}
	}
	surface := browserRuntimeManagedOptInDiagnosticsSurface{}
	surface.ToolNames = extraCompatTools
	surface.ActKinds = extraActKinds
	surface.Targets = []string{runtimeInfo.Target}
	surface.Capabilities = BrowserCapabilitiesForActKinds(extraActKinds)
	for _, name := range extraCompatTools {
		if kind := browserCompatManagedOptInActKind(name); kind != "" {
			surface.Capabilities = mergeBrowserCapabilities(
				surface.Capabilities,
				BrowserCapabilitiesForActKinds([]string{kind}),
			)
		}
	}
	return surface
}

func browserRuntimeManagedOptInDiagnosticsSurfaceForRoute(
	ctx browserRegistrationContext,
	runtimeInfo BrowserRuntimeInfo,
) browserRuntimeManagedOptInDiagnosticsSurface {
	runtimeInfo = normalizeBrowserRuntimeInfo(runtimeInfo)
	switch runtimeInfo.Target {
	case "node", "sandbox":
		return browserRuntimeManagedOptInSurfaceForRoute(ctx, runtimeInfo)
	case "", "host":
		merged := browserRuntimeManagedOptInDiagnosticsSurface{}
		for _, target := range []string{"node", "sandbox"} {
			candidate := runtimeInfo
			candidate.Target = target
			surface := browserRuntimeManagedOptInSurfaceForRoute(ctx, candidate)
			merged = browserRuntimeMergeManagedOptInDiagnosticsSurfaces(merged, surface)
		}
		return merged
	default:
		return browserRuntimeManagedOptInDiagnosticsSurface{}
	}
}

func browserRuntimeManagedOptInDiagnosticsSurfaceForRouteWithPreview(
	ctx browserRegistrationContext,
	preview browserRuntimeDiagnosticsPreview,
	runtimeInfo BrowserRuntimeInfo,
) browserRuntimeManagedOptInDiagnosticsSurface {
	runtimeInfo = normalizeBrowserRuntimeInfo(runtimeInfo)
	switch runtimeInfo.Target {
	case "node", "sandbox":
		return browserRuntimeManagedOptInSurfaceForRouteWithPreview(ctx, preview, runtimeInfo)
	case "", "host":
		merged := browserRuntimeManagedOptInDiagnosticsSurface{}
		for _, target := range []string{"node", "sandbox"} {
			if !browserRuntimePreviewConfiguresManagedTarget(preview, target) {
				continue
			}
			candidate := runtimeInfo
			candidate.Target = target
			surface := browserRuntimeManagedOptInSurfaceForRouteWithPreview(ctx, preview, candidate)
			merged = browserRuntimeMergeManagedOptInDiagnosticsSurfaces(merged, surface)
		}
		return merged
	default:
		return browserRuntimeManagedOptInDiagnosticsSurface{}
	}
}

func browserRuntimePreviewConfiguresManagedTarget(preview browserRuntimeDiagnosticsPreview, target string) bool {
	target = strings.ToLower(strings.TrimSpace(target))
	for _, configured := range mergeToolMetadataStrings(nil, preview.ConfiguredTargets) {
		if strings.EqualFold(strings.TrimSpace(configured), target) {
			return true
		}
	}
	switch target {
	case "node":
		return preview.Registration.SubstrateSummary.NodeConfigured || preview.Registration.SubstrateAssessment.NodeRoute.Configured
	case "sandbox":
		return preview.Registration.SubstrateSummary.SandboxConfigured || preview.Registration.SubstrateAssessment.SandboxRoute.Configured
	default:
		return false
	}
}

func browserRuntimeMergeManagedOptInDiagnosticsSurfaces(
	base browserRuntimeManagedOptInDiagnosticsSurface,
	overlay browserRuntimeManagedOptInDiagnosticsSurface,
) browserRuntimeManagedOptInDiagnosticsSurface {
	return browserRuntimeManagedOptInDiagnosticsSurface{
		ToolNames:    mergeToolMetadataStrings(base.ToolNames, overlay.ToolNames),
		ActKinds:     mergeToolMetadataStrings(base.ActKinds, overlay.ActKinds),
		Targets:      mergeToolMetadataStrings(base.Targets, overlay.Targets),
		Capabilities: mergeBrowserCapabilities(base.Capabilities, overlay.Capabilities),
	}
}

func browserRuntimeSelectedOnlyCompatTools(selected []string, defaults []string) []string {
	out := make([]string, 0, len(selected))
	defaultSet := map[string]bool{}
	for _, name := range mergeToolMetadataStrings(nil, defaults) {
		defaultSet[name] = true
	}
	for _, rawName := range mergeToolMetadataStrings(nil, selected) {
		name := NormalizeToolName(rawName)
		if name == "" || !IsBrowserCompatToolName(name) || defaultSet[name] {
			continue
		}
		out = append(out, name)
	}
	return mergeToolMetadataStrings(nil, out)
}

func browserRuntimeSelectedOnlyItems(selected []string, defaults []string) []string {
	out := make([]string, 0, len(selected))
	defaultSet := map[string]bool{}
	for _, item := range mergeToolMetadataStrings(nil, defaults) {
		defaultSet[item] = true
	}
	for _, item := range mergeToolMetadataStrings(nil, selected) {
		if defaultSet[item] {
			continue
		}
		out = append(out, item)
	}
	return mergeToolMetadataStrings(nil, out)
}
