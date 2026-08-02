package tools

import (
	"strings"
	"sync"

	llmxtools "github.com/wsnacj/agentx-go/tools"
)

type browserRuntimeDiagnosticsSurfaceMetadataCacheKey struct {
	BaseCapabilities  BrowserCapabilities
	OptInCapabilities BrowserCapabilities
	Targets           string
}

type browserRuntimeManagedOptInResolvedSurfaceCacheKey struct {
	Target       string
	Capabilities BrowserCapabilities
}

type browserRuntimeResolvedRouteSurfaceMetadataCacheKey struct {
	Target       string
	Capabilities BrowserCapabilities
}

type browserRuntimeAssessmentSurfaceProjectionCacheKey struct {
	Role                              string
	Info                              BrowserRuntimeInfo
	RouteAvailable                    bool
	RouteInfo                         BrowserRuntimeInfo
	Capabilities                      BrowserCapabilities
	PreserveDefaultMetadataWhenHidden bool
}

type browserRuntimeRouteFallbackInfoCacheKey struct {
	Profile string
	Target  string
}

type browserRegistrationDerivedCache struct {
	registeredBrowserToolsOnce sync.Once
	registeredBrowserTools     map[string]bool

	registeredSurfaceMu sync.Mutex
	registeredSurface   map[BrowserCapabilities]browserRuntimeRegisteredSurface

	capabilityMetadataMu sync.Mutex
	capabilityMetadata   map[BrowserCapabilities]browserRuntimeCapabilityMetadata

	managedOptInSurfaceOnce sync.Once
	managedOptInSurface     browserRuntimeManagedOptInDiagnosticsSurface

	diagnosticsMetadataOnce sync.Once
	diagnosticsMetadata     browserRuntimeCapabilityMetadata

	managedOptInRouteSurfaceMu sync.Mutex
	managedOptInRouteSurface   map[string]browserRuntimeManagedOptInDiagnosticsSurface

	diagnosticsSurfaceMetadataMu sync.Mutex
	diagnosticsSurfaceMetadata   map[browserRuntimeDiagnosticsSurfaceMetadataCacheKey]browserRuntimeCapabilityMetadata

	managedOptInResolvedSurfaceMu sync.Mutex
	managedOptInResolvedSurface   map[browserRuntimeManagedOptInResolvedSurfaceCacheKey]browserRuntimeManagedOptInDiagnosticsSurface

	resolvedRouteSurfaceMetadataMu    sync.Mutex
	resolvedRouteSurfaceMetadataCache map[browserRuntimeResolvedRouteSurfaceMetadataCacheKey]browserRuntimeCapabilityMetadata

	assessmentSurfaceProjectionMu    sync.Mutex
	assessmentSurfaceProjectionCache map[browserRuntimeAssessmentSurfaceProjectionCacheKey]browserRuntimeAssessmentSurfaceProjection

	routeFallbackInfoMu    sync.Mutex
	routeFallbackInfoCache map[browserRuntimeRouteFallbackInfoCacheKey]BrowserRuntimeInfo
}

func newBrowserRegistrationDerivedCache() *browserRegistrationDerivedCache {
	return &browserRegistrationDerivedCache{}
}

func (c *browserRegistrationDerivedCache) registeredToolSet(reg *llmxtools.Registry) map[string]bool {
	if c == nil {
		return browserRuntimeRegisteredToolSetFromRegistry(reg)
	}
	c.registeredBrowserToolsOnce.Do(func() {
		c.registeredBrowserTools = browserRuntimeRegisteredToolSetFromRegistry(reg)
	})
	return c.registeredBrowserTools
}

func (c *browserRegistrationDerivedCache) managedOptInDiagnosticsSurface(ctx browserRegistrationContext) browserRuntimeManagedOptInDiagnosticsSurface {
	if c == nil {
		return browserRuntimeManagedOptInCapabilitiesUncached(ctx)
	}
	c.managedOptInSurfaceOnce.Do(func() {
		c.managedOptInSurface = browserRuntimeManagedOptInCapabilitiesUncached(ctx)
	})
	return cloneBrowserRuntimeManagedOptInDiagnosticsSurface(c.managedOptInSurface)
}

func (c *browserRegistrationDerivedCache) capabilityMetadataForCapabilities(
	ctx browserRegistrationContext,
	capabilities BrowserCapabilities,
) browserRuntimeCapabilityMetadata {
	if c == nil {
		return browserRuntimeCapabilityMetadataForCapabilitiesUncached(ctx, capabilities)
	}
	c.capabilityMetadataMu.Lock()
	defer c.capabilityMetadataMu.Unlock()
	if c.capabilityMetadata == nil {
		c.capabilityMetadata = map[BrowserCapabilities]browserRuntimeCapabilityMetadata{}
	}
	if metadata, ok := c.capabilityMetadata[capabilities]; ok {
		return cloneBrowserRuntimeCapabilityMetadata(metadata)
	}
	metadata := browserRuntimeCapabilityMetadataForCapabilitiesUncached(ctx, capabilities)
	c.capabilityMetadata[capabilities] = cloneBrowserRuntimeCapabilityMetadata(metadata)
	return cloneBrowserRuntimeCapabilityMetadata(metadata)
}

func (c *browserRegistrationDerivedCache) diagnosticsCapabilityMetadata(ctx browserRegistrationContext) browserRuntimeCapabilityMetadata {
	if c == nil {
		return browserRuntimeDiagnosticsMetadataUncached(ctx)
	}
	c.diagnosticsMetadataOnce.Do(func() {
		c.diagnosticsMetadata = browserRuntimeDiagnosticsMetadataUncached(ctx)
	})
	return cloneBrowserRuntimeCapabilityMetadata(c.diagnosticsMetadata)
}

func (c *browserRegistrationDerivedCache) registeredSurfaceForCapabilities(
	ctx browserRegistrationContext,
	capabilities BrowserCapabilities,
) browserRuntimeRegisteredSurface {
	if c == nil {
		return browserRuntimeRegisteredSurfaceForCapabilitiesUncached(ctx, capabilities)
	}
	c.registeredSurfaceMu.Lock()
	defer c.registeredSurfaceMu.Unlock()
	if c.registeredSurface == nil {
		c.registeredSurface = map[BrowserCapabilities]browserRuntimeRegisteredSurface{}
	}
	if surface, ok := c.registeredSurface[capabilities]; ok {
		return cloneBrowserRuntimeRegisteredSurface(surface)
	}
	surface := browserRuntimeRegisteredSurfaceForCapabilitiesUncached(ctx, capabilities)
	c.registeredSurface[capabilities] = cloneBrowserRuntimeRegisteredSurface(surface)
	return cloneBrowserRuntimeRegisteredSurface(surface)
}

func (c *browserRegistrationDerivedCache) managedOptInSurfaceForRoute(
	ctx browserRegistrationContext,
	runtimeInfo BrowserRuntimeInfo,
) browserRuntimeManagedOptInDiagnosticsSurface {
	if c == nil {
		return browserRuntimeManagedOptInSurfaceForRouteUncached(ctx, runtimeInfo)
	}
	key, ok := browserRuntimeManagedOptInSurfaceRouteCacheKey(runtimeInfo)
	if !ok {
		return browserRuntimeManagedOptInSurfaceForRouteUncached(ctx, runtimeInfo)
	}
	c.managedOptInRouteSurfaceMu.Lock()
	defer c.managedOptInRouteSurfaceMu.Unlock()
	if c.managedOptInRouteSurface == nil {
		c.managedOptInRouteSurface = map[string]browserRuntimeManagedOptInDiagnosticsSurface{}
	}
	if surface, ok := c.managedOptInRouteSurface[key]; ok {
		return cloneBrowserRuntimeManagedOptInDiagnosticsSurface(surface)
	}
	surface := browserRuntimeManagedOptInSurfaceForRouteUncached(ctx, runtimeInfo)
	c.managedOptInRouteSurface[key] = cloneBrowserRuntimeManagedOptInDiagnosticsSurface(surface)
	return cloneBrowserRuntimeManagedOptInDiagnosticsSurface(surface)
}

func (c *browserRegistrationDerivedCache) diagnosticsSurfaceCapabilityMetadata(
	ctx browserRegistrationContext,
	baseCapabilities BrowserCapabilities,
	optInSurface browserRuntimeManagedOptInDiagnosticsSurface,
) browserRuntimeCapabilityMetadata {
	if c == nil {
		return browserRuntimeCapabilityMetadataForDiagnosticsSurfaceUncached(ctx, baseCapabilities, optInSurface)
	}
	key := browserRuntimeDiagnosticsSurfaceMetadataCacheKey{
		BaseCapabilities:  baseCapabilities,
		OptInCapabilities: optInSurface.Capabilities,
		Targets:           strings.Join(mergeToolMetadataStrings(nil, optInSurface.Targets), "\n"),
	}
	c.diagnosticsSurfaceMetadataMu.Lock()
	defer c.diagnosticsSurfaceMetadataMu.Unlock()
	if c.diagnosticsSurfaceMetadata == nil {
		c.diagnosticsSurfaceMetadata = map[browserRuntimeDiagnosticsSurfaceMetadataCacheKey]browserRuntimeCapabilityMetadata{}
	}
	if metadata, ok := c.diagnosticsSurfaceMetadata[key]; ok {
		return cloneBrowserRuntimeCapabilityMetadata(metadata)
	}
	metadata := browserRuntimeCapabilityMetadataForDiagnosticsSurfaceUncached(ctx, baseCapabilities, optInSurface)
	c.diagnosticsSurfaceMetadata[key] = cloneBrowserRuntimeCapabilityMetadata(metadata)
	return cloneBrowserRuntimeCapabilityMetadata(metadata)
}

func (c *browserRegistrationDerivedCache) managedOptInSurfaceForResolvedRoute(
	ctx browserRegistrationContext,
	runtimeInfo BrowserRuntimeInfo,
	capabilities BrowserCapabilities,
) browserRuntimeManagedOptInDiagnosticsSurface {
	if c == nil {
		return browserRuntimeManagedOptInSurfaceForResolvedRouteUncached(ctx, runtimeInfo, capabilities)
	}
	key, ok := browserRuntimeManagedOptInResolvedSurfaceCacheKeyForRoute(runtimeInfo, capabilities)
	if !ok {
		return browserRuntimeManagedOptInSurfaceForResolvedRouteUncached(ctx, runtimeInfo, capabilities)
	}
	c.managedOptInResolvedSurfaceMu.Lock()
	defer c.managedOptInResolvedSurfaceMu.Unlock()
	if c.managedOptInResolvedSurface == nil {
		c.managedOptInResolvedSurface = map[browserRuntimeManagedOptInResolvedSurfaceCacheKey]browserRuntimeManagedOptInDiagnosticsSurface{}
	}
	if surface, ok := c.managedOptInResolvedSurface[key]; ok {
		return cloneBrowserRuntimeManagedOptInDiagnosticsSurface(surface)
	}
	surface := browserRuntimeManagedOptInSurfaceForResolvedRouteUncached(ctx, runtimeInfo, capabilities)
	c.managedOptInResolvedSurface[key] = cloneBrowserRuntimeManagedOptInDiagnosticsSurface(surface)
	return cloneBrowserRuntimeManagedOptInDiagnosticsSurface(surface)
}

func (c *browserRegistrationDerivedCache) resolvedRouteSurfaceMetadata(
	ctx browserRegistrationContext,
	route browserResolvedExecutionRoute,
) (browserRuntimeCapabilityMetadata, BrowserCapabilities) {
	capabilities := browserCapabilitiesForRuntimeInspection(ctx, route)
	if c == nil {
		return browserRuntimeResolvedRouteSurfaceMetadataUncached(ctx, route, capabilities), capabilities
	}
	key, ok := browserRuntimeResolvedRouteSurfaceMetadataCacheKeyForRoute(route.RuntimeInfo, capabilities)
	if !ok {
		return browserRuntimeResolvedRouteSurfaceMetadataUncached(ctx, route, capabilities), capabilities
	}
	c.resolvedRouteSurfaceMetadataMu.Lock()
	defer c.resolvedRouteSurfaceMetadataMu.Unlock()
	if c.resolvedRouteSurfaceMetadataCache == nil {
		c.resolvedRouteSurfaceMetadataCache = map[browserRuntimeResolvedRouteSurfaceMetadataCacheKey]browserRuntimeCapabilityMetadata{}
	}
	if metadata, ok := c.resolvedRouteSurfaceMetadataCache[key]; ok {
		return cloneBrowserRuntimeCapabilityMetadata(metadata), capabilities
	}
	metadata := browserRuntimeResolvedRouteSurfaceMetadataUncached(ctx, route, capabilities)
	c.resolvedRouteSurfaceMetadataCache[key] = cloneBrowserRuntimeCapabilityMetadata(metadata)
	return cloneBrowserRuntimeCapabilityMetadata(metadata), capabilities
}

func (c *browserRegistrationDerivedCache) assessmentSurfaceProjection(
	ctx browserRegistrationContext,
	role string,
	info BrowserRuntimeInfo,
	assessment browserConcreteRouteAssessment,
	preserveDefaultMetadataWhenHidden bool,
) browserRuntimeAssessmentSurfaceProjection {
	info = normalizeBrowserRuntimeInfo(info)
	capabilities := BrowserCapabilities{}
	if assessment.RouteAvailable {
		capabilities = browserCapabilitiesForRuntimeInspection(ctx, assessment.Route)
	}
	if c == nil {
		return browserRuntimeAssessmentSurfaceProjectionUncached(ctx, role, info, assessment, capabilities, preserveDefaultMetadataWhenHidden)
	}
	key := browserRuntimeAssessmentSurfaceProjectionCacheKey{
		Role:                              strings.ToLower(strings.TrimSpace(role)),
		Info:                              info,
		RouteAvailable:                    assessment.RouteAvailable,
		Capabilities:                      capabilities,
		PreserveDefaultMetadataWhenHidden: preserveDefaultMetadataWhenHidden,
	}
	if assessment.RouteAvailable {
		key.RouteInfo = normalizeBrowserRuntimeInfo(assessment.Route.RuntimeInfo)
	}
	c.assessmentSurfaceProjectionMu.Lock()
	defer c.assessmentSurfaceProjectionMu.Unlock()
	if c.assessmentSurfaceProjectionCache == nil {
		c.assessmentSurfaceProjectionCache = map[browserRuntimeAssessmentSurfaceProjectionCacheKey]browserRuntimeAssessmentSurfaceProjection{}
	}
	if projection, ok := c.assessmentSurfaceProjectionCache[key]; ok {
		return cloneBrowserRuntimeAssessmentSurfaceProjection(projection)
	}
	projection := browserRuntimeAssessmentSurfaceProjectionUncached(ctx, role, info, assessment, capabilities, preserveDefaultMetadataWhenHidden)
	c.assessmentSurfaceProjectionCache[key] = cloneBrowserRuntimeAssessmentSurfaceProjection(projection)
	return cloneBrowserRuntimeAssessmentSurfaceProjection(projection)
}

func (c *browserRegistrationDerivedCache) routeFallbackInfo(
	ctx browserRegistrationContext,
	profile string,
	target string,
) BrowserRuntimeInfo {
	if c == nil {
		return browserRuntimeRouteFallbackInfoForPreviewTarget(
			ctx,
			browserRuntimeDiagnosticsPreviewBaseForRegistration(ctx),
			profile,
			target,
		)
	}
	key := browserRuntimeRouteFallbackInfoCacheKey{
		Profile: strings.ToLower(strings.TrimSpace(profile)),
		Target:  strings.ToLower(strings.TrimSpace(target)),
	}
	c.routeFallbackInfoMu.Lock()
	defer c.routeFallbackInfoMu.Unlock()
	if c.routeFallbackInfoCache == nil {
		c.routeFallbackInfoCache = map[browserRuntimeRouteFallbackInfoCacheKey]BrowserRuntimeInfo{}
	}
	if info, ok := c.routeFallbackInfoCache[key]; ok {
		return info
	}
	info := browserRuntimeRouteFallbackInfoForPreviewTarget(
		ctx,
		browserRuntimeDiagnosticsPreviewBaseForRegistration(ctx),
		profile,
		target,
	)
	c.routeFallbackInfoCache[key] = info
	return info
}

func browserRuntimeManagedOptInSurfaceRouteCacheKey(runtimeInfo BrowserRuntimeInfo) (string, bool) {
	runtimeInfo = normalizeBrowserRuntimeInfo(runtimeInfo)
	switch runtimeInfo.Target {
	case "node", "sandbox":
	default:
		return "", false
	}
	return runtimeInfo.Target + "\n" + runtimeInfo.Profile, true
}

func browserRuntimeManagedOptInResolvedSurfaceCacheKeyForRoute(
	runtimeInfo BrowserRuntimeInfo,
	capabilities BrowserCapabilities,
) (browserRuntimeManagedOptInResolvedSurfaceCacheKey, bool) {
	runtimeInfo = normalizeBrowserRuntimeInfo(runtimeInfo)
	switch runtimeInfo.Target {
	case "node", "sandbox":
		return browserRuntimeManagedOptInResolvedSurfaceCacheKey{
			Target:       runtimeInfo.Target,
			Capabilities: capabilities,
		}, true
	default:
		return browserRuntimeManagedOptInResolvedSurfaceCacheKey{}, false
	}
}

func browserRuntimeResolvedRouteSurfaceMetadataCacheKeyForRoute(
	runtimeInfo BrowserRuntimeInfo,
	capabilities BrowserCapabilities,
) (browserRuntimeResolvedRouteSurfaceMetadataCacheKey, bool) {
	runtimeInfo = normalizeBrowserRuntimeInfo(runtimeInfo)
	return browserRuntimeResolvedRouteSurfaceMetadataCacheKey{
		Target:       runtimeInfo.Target,
		Capabilities: capabilities,
	}, true
}

func cloneBrowserRuntimeManagedOptInDiagnosticsSurface(surface browserRuntimeManagedOptInDiagnosticsSurface) browserRuntimeManagedOptInDiagnosticsSurface {
	surface.ToolNames = append([]string(nil), surface.ToolNames...)
	surface.ActKinds = append([]string(nil), surface.ActKinds...)
	surface.Targets = append([]string(nil), surface.Targets...)
	return surface
}

func cloneBrowserRuntimeRegisteredSurface(surface browserRuntimeRegisteredSurface) browserRuntimeRegisteredSurface {
	surface.BrowserTools = append([]string(nil), surface.BrowserTools...)
	surface.ArtifactTools = append([]string(nil), surface.ArtifactTools...)
	surface.BrowserActKinds = append([]string(nil), surface.BrowserActKinds...)
	return surface
}

func cloneBrowserRuntimeCapabilityMetadata(metadata browserRuntimeCapabilityMetadata) browserRuntimeCapabilityMetadata {
	metadata.RuntimeActions = append([]string(nil), metadata.RuntimeActions...)
	metadata.BrowserTools = append([]string(nil), metadata.BrowserTools...)
	metadata.ArtifactTools = append([]string(nil), metadata.ArtifactTools...)
	metadata.ArtifactKinds = append([]string(nil), metadata.ArtifactKinds...)
	metadata.BrowserActKinds = append([]string(nil), metadata.BrowserActKinds...)
	metadata.BrowserOptInTargets = append([]string(nil), metadata.BrowserOptInTargets...)
	if metadata.Capabilities != nil {
		cloned := make(map[string]bool, len(metadata.Capabilities))
		for key, value := range metadata.Capabilities {
			cloned[key] = value
		}
		metadata.Capabilities = cloned
	}
	return metadata
}

func cloneBrowserRuntimeAssessmentSurfaceProjection(projection browserRuntimeAssessmentSurfaceProjection) browserRuntimeAssessmentSurfaceProjection {
	projection.Metadata = cloneBrowserRuntimeCapabilityMetadata(projection.Metadata)
	return projection
}
