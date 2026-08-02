package tools

import (
	"fmt"
	"strings"
)

func (b browserRuntimeRouterBackend) resolveBrowserRoute(requested BrowserRuntimeInfo) (browserResolvedExecutionRoute, error) {
	requested = normalizeBrowserRuntimeInfo(requested)
	requestedNoBackend := requested
	requestedNoBackend.Backend = ""
	if requested.Target == "" {
		if route, ok := b.resolveImplicitLegacyHostManagedDefaultRoute(requestedNoBackend); ok {
			return route, nil
		}
		if err := b.implicitManagedDefaultRouteFailure(requestedNoBackend); err != nil {
			return browserResolvedExecutionRoute{}, err
		}
		executionPreview := b.executionPreview()
		if err := browserImplicitLegacyHostDefaultRequestError(
			executionPreview.HiddenImplicitHostDefaultBase,
			requestedNoBackend,
			firstNonEmpty(b.hostRuntime.Profile, defaultBrowserRuntimeInfo().Profile),
		); err != nil {
			return browserResolvedExecutionRoute{}, err
		}
		if assessment, ok := b.cachedDefaultConcreteRouteAssessmentForRequestedProfile(requestedNoBackend); ok {
			return b.resolvedExecutionRouteFromAssessment(assessment, "default browser route is unavailable")
		}
		defaultRequested := requestedNoBackend
		defaultRequested.Target = executionPreview.DefaultTarget
		return b.resolveTargetBrowserRoute(defaultRequested, "default browser route is unavailable")
	}
	return b.resolveTargetBrowserRoute(requestedNoBackend, requested.Target+" browser route is unavailable")
}

func (b browserRuntimeRouterBackend) resolveImplicitLegacyHostManagedDefaultRoute(requested BrowserRuntimeInfo) (browserResolvedExecutionRoute, bool) {
	if !b.usesImplicitLegacyHostDefaultFallback() {
		return browserResolvedExecutionRoute{}, false
	}
	requested = normalizeBrowserRuntimeInfo(requested)
	if strings.TrimSpace(requested.Target) != "" {
		return browserResolvedExecutionRoute{}, false
	}
	defaultProfile := strings.TrimSpace(firstNonEmpty(b.hostRuntime.Profile, defaultBrowserRuntimeInfo().Profile))
	requestedProfile := strings.TrimSpace(requested.Profile)
	defaultProfileRequest := requestedProfile == "" || strings.EqualFold(requestedProfile, defaultProfile)
	if defaultProfileRequest {
		requested.Profile = ""
	}
	for _, target := range []string{"node"} {
		route, err := b.resolveImplicitLegacyHostManagedDefaultTargetRoute(target, requested, defaultProfileRequest)
		if err == nil {
			return route, true
		}
	}
	return browserResolvedExecutionRoute{}, false
}

func (b browserRuntimeRouterBackend) implicitManagedDefaultRouteFailure(requested BrowserRuntimeInfo) error {
	if !b.usesImplicitLegacyHostDefaultFallback() {
		return nil
	}
	requested = normalizeBrowserRuntimeInfo(requested)
	if strings.TrimSpace(requested.Target) != "" {
		return nil
	}
	defaultProfile := strings.TrimSpace(firstNonEmpty(b.hostRuntime.Profile, defaultBrowserRuntimeInfo().Profile))
	requestedProfile := strings.TrimSpace(requested.Profile)
	defaultProfileRequest := requestedProfile == "" || strings.EqualFold(requestedProfile, defaultProfile)
	if !defaultProfileRequest {
		return nil
	}
	requested.Profile = ""
	for _, target := range []string{"node"} {
		candidate := requested
		candidate.Target = target
		assessment, ok := b.cachedImplicitManagedDefaultRouteAssessment(candidate)
		if !ok || !assessment.Configured || assessment.RouteAvailable {
			continue
		}
		if note := strings.TrimSpace(assessment.FailureNote); note != "" {
			return fmt.Errorf("%s", note)
		}
		if reason := strings.TrimSpace(assessment.FailureReason); reason != "" {
			return fmt.Errorf("%s", reason)
		}
	}
	return nil
}

func (b browserRuntimeRouterBackend) resolveImplicitLegacyHostManagedDefaultTargetRoute(target string, requested BrowserRuntimeInfo, defaultProfileRequest bool) (browserResolvedExecutionRoute, error) {
	candidate := requested
	candidate.Target = target
	if !defaultProfileRequest {
		return b.resolveTargetBrowserRoute(candidate, target+" browser route is unavailable")
	}
	if assessment, ok := b.cachedImplicitManagedDefaultRouteAssessment(candidate); ok {
		return b.resolvedExecutionRouteFromAssessment(assessment, target+" browser route is unavailable")
	}
	route, err := b.resolveImplicitManagedDefaultConcreteRoute(candidate)
	if err != nil {
		b.storeImplicitManagedDefaultRouteAssessment(candidate, browserConcreteRouteAssessment{
			Configured:    browserImplicitManagedDefaultBackendConfigured(target, b),
			FailureReason: err.Error(),
			FailureNote:   err.Error(),
		})
		return browserResolvedExecutionRoute{}, err
	}
	b.storeImplicitManagedDefaultRouteAssessment(candidate, browserConcreteRouteAssessment{
		Configured:     true,
		RouteAvailable: true,
		Route:          route,
	})
	return route, nil
}

func (b browserRuntimeRouterBackend) resolveImplicitManagedDefaultConcreteRoute(requested BrowserRuntimeInfo) (browserResolvedExecutionRoute, error) {
	requested = normalizeBrowserRuntimeInfo(requested)
	switch requested.Target {
	case "node":
		if b.nodeBackend == nil {
			return browserResolvedExecutionRoute{}, fmt.Errorf("runtime_target %q is unsupported for backend %q (no node backend configured)", requested.Target, browserRuntimeBackendName(b.hostRuntime))
		}
		return b.resolveConcreteBrowserRoute(b.nodeBackend, BrowserRuntimeInfo{Backend: "node", Target: "node"}, requested)
	case "sandbox":
		if b.sandboxBackend == nil {
			return browserResolvedExecutionRoute{}, fmt.Errorf("runtime_target %q is unsupported for backend %q (no sandbox backend configured)", requested.Target, browserRuntimeBackendName(b.hostRuntime))
		}
		return b.resolveConcreteBrowserRoute(b.sandboxBackend, BrowserRuntimeInfo{Backend: "sandbox", Target: "sandbox"}, requested)
	default:
		return browserResolvedExecutionRoute{}, fmt.Errorf("runtime_target %q is unsupported", requested.Target)
	}
}

func (b browserRuntimeRouterBackend) cachedImplicitManagedDefaultRouteAssessment(requested BrowserRuntimeInfo) (browserConcreteRouteAssessment, bool) {
	info, ok := browserRuntimeImplicitManagedDefaultCacheInfo(requested)
	if !ok {
		return browserConcreteRouteAssessment{}, false
	}
	return b.implicitManagedDefaultRoute.Load(info.Profile, info.Target)
}

func (b browserRuntimeRouterBackend) storeImplicitManagedDefaultRouteAssessment(requested BrowserRuntimeInfo, assessment browserConcreteRouteAssessment) {
	info, ok := browserRuntimeImplicitManagedDefaultCacheInfo(requested)
	if !ok {
		return
	}
	b.implicitManagedDefaultRoute.Store(info.Profile, info.Target, assessment)
}

func browserRuntimeImplicitManagedDefaultCacheInfo(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, bool) {
	requested = normalizeBrowserRuntimeInfo(requested)
	target := strings.TrimSpace(requested.Target)
	if target == "" {
		return BrowserRuntimeInfo{}, false
	}
	profile := strings.TrimSpace(requested.Profile)
	if profile == "" {
		profile = "__default__"
	}
	return BrowserRuntimeInfo{Profile: profile, Target: target}, true
}

func browserImplicitManagedDefaultBackendConfigured(target string, router browserRuntimeRouterBackend) bool {
	switch strings.TrimSpace(target) {
	case "node":
		return router.nodeBackend != nil
	case "sandbox":
		return router.sandboxBackend != nil
	default:
		return false
	}
}

func (b browserRuntimeRouterBackend) resolveTargetBrowserRoute(requested BrowserRuntimeInfo, unavailableMessage string) (browserResolvedExecutionRoute, error) {
	requested = normalizeBrowserRuntimeInfo(requested)
	if assessment, ok := b.cachedTargetRouteAssessmentForRouteResolve(requested); ok {
		return b.resolvedExecutionRouteFromAssessment(assessment, unavailableMessage)
	}
	switch requested.Target {
	case "host":
		return b.resolveConcreteBrowserRouteWithDynamicCache(b.hostBackend, b.hostRuntime, requested)
	case "sandbox":
		if b.sandboxBackend == nil {
			return browserResolvedExecutionRoute{}, fmt.Errorf("runtime_target %q is unsupported for backend %q (no sandbox backend configured)", requested.Target, browserRuntimeBackendName(b.hostRuntime))
		}
		return b.resolveConcreteBrowserRouteWithDynamicCache(b.sandboxBackend, BrowserRuntimeInfo{
			Backend: "sandbox",
			Profile: requested.Profile,
			Target:  "sandbox",
		}, requested)
	case "node":
		if b.nodeBackend == nil {
			return browserResolvedExecutionRoute{}, fmt.Errorf("runtime_target %q is unsupported for backend %q (no node backend configured)", requested.Target, browserRuntimeBackendName(b.hostRuntime))
		}
		return b.resolveConcreteBrowserRouteWithDynamicCache(b.nodeBackend, BrowserRuntimeInfo{
			Backend: "node",
			Profile: requested.Profile,
			Target:  "node",
		}, requested)
	default:
		return browserResolvedExecutionRoute{}, fmt.Errorf("runtime_target %q is unsupported", requested.Target)
	}
}

func (b browserRuntimeRouterBackend) cachedTargetRouteAssessmentForRouteResolve(requested BrowserRuntimeInfo) (browserConcreteRouteAssessment, bool) {
	requested = normalizeBrowserRuntimeInfo(requested)
	profile := strings.ToLower(strings.TrimSpace(requested.Profile))
	switch requested.Target {
	case "host":
		cachedProfile := browserCachedRouteAssessmentProfile(b.hostRoute, b.hostRuntime.Profile)
		return browserRouterCachedRouteAssessmentForRequestedProfile(profile, cachedProfile, b.hostRoute)
	case "node":
		nodeAssessment := browserConcreteRouteAssessmentForDefaultPromotion(b.nodeRoute)
		cachedProfile := browserCachedRouteAssessmentProfile(nodeAssessment, defaultBrowserNodeRuntimeInfo().Profile)
		return browserRouterCachedRouteAssessmentForRequestedProfile(profile, cachedProfile, nodeAssessment)
	case "sandbox":
		cachedProfile := browserCachedRouteAssessmentProfile(b.sandboxRoute, defaultBrowserSandboxRuntimeInfo().Profile)
		return browserRouterCachedRouteAssessmentForRequestedProfile(profile, cachedProfile, b.sandboxRoute)
	default:
		return browserConcreteRouteAssessment{}, false
	}
}

func (b browserRuntimeRouterBackend) cachedDefaultConcreteRouteAssessmentForRequestedProfile(requested BrowserRuntimeInfo) (browserConcreteRouteAssessment, bool) {
	requested = normalizeBrowserRuntimeInfo(requested)
	profile := strings.ToLower(strings.TrimSpace(requested.Profile))
	assessment, ok := b.cachedDefaultConcreteRouteAssessment()
	if !ok {
		return browserConcreteRouteAssessment{}, false
	}
	cachedProfile := browserCachedRouteAssessmentProfile(assessment, b.defaultRuntimeInfo().Profile)
	if assessment, ok := browserRouterCachedRouteAssessmentForRequestedProfile(profile, cachedProfile, assessment); ok {
		return assessment, true
	}
	return browserCachedRouteAssessmentForDiagnosticsProfile(profile, cachedProfile, assessment)
}

func (b browserRuntimeRouterBackend) usesImplicitLegacyHostDefaultFallback() bool {
	defaultRuntime := normalizeBrowserRuntimeInfo(b.baseDefaultRuntimeInfo())
	if BrowserSubstratePosture(defaultRuntime.Backend, defaultRuntime.Target) != BrowserSubstrateLegacySystemHost {
		return false
	}
	return !b.defaultRoute.RouteAvailable
}

func (b browserRuntimeRouterBackend) cachedTargetRouteAssessment(requested BrowserRuntimeInfo) (browserConcreteRouteAssessment, bool) {
	requested = normalizeBrowserRuntimeInfo(requested)
	profile := strings.ToLower(strings.TrimSpace(requested.Profile))
	switch requested.Target {
	case "host":
		cachedProfile := browserCachedRouteAssessmentProfile(b.hostRoute, b.hostRuntime.Profile)
		if assessment, ok := browserRouterCachedRouteAssessmentForRequestedProfile(profile, cachedProfile, b.hostRoute); ok {
			return assessment, true
		}
		return browserCachedRouteAssessmentForDiagnosticsProfile(profile, cachedProfile, b.hostRoute)
	case "node":
		if profile == "" {
			if assessment, ok := b.cachedImplicitManagedDefaultRouteProbeAssessment(BrowserRuntimeInfo{Target: "node"}); ok && assessment.RouteAvailable {
				return assessment, true
			}
		}
		nodeAssessment := browserConcreteRouteAssessmentForDefaultPromotion(b.nodeRoute)
		cachedProfile := browserCachedRouteAssessmentProfile(nodeAssessment, defaultBrowserNodeRuntimeInfo().Profile)
		if assessment, ok := browserRouterCachedRouteAssessmentForRequestedProfile(profile, cachedProfile, nodeAssessment); ok {
			return assessment, true
		}
		return browserCachedRouteAssessmentForDiagnosticsProfile(profile, cachedProfile, nodeAssessment)
	case "sandbox":
		cachedProfile := browserCachedRouteAssessmentProfile(b.sandboxRoute, defaultBrowserSandboxRuntimeInfo().Profile)
		if assessment, ok := browserRouterCachedRouteAssessmentForRequestedProfile(profile, cachedProfile, b.sandboxRoute); ok {
			return assessment, true
		}
		return browserCachedRouteAssessmentForDiagnosticsProfile(profile, cachedProfile, b.sandboxRoute)
	default:
		return browserConcreteRouteAssessment{}, false
	}
}

func (b browserRuntimeRouterBackend) dynamicRouteCacheInfo(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, bool) {
	requested = normalizeBrowserRuntimeInfo(requested)
	if strings.TrimSpace(requested.Profile) == "" {
		return BrowserRuntimeInfo{}, false
	}
	if strings.TrimSpace(requested.Target) == "" {
		requested.Target = b.defaultRuntimeInfo().Target
	}
	if strings.TrimSpace(requested.Target) == "" {
		return BrowserRuntimeInfo{}, false
	}
	requested.Backend = ""
	return requested, true
}

func (b browserRuntimeRouterBackend) cachedDynamicRouteAssessment(requested BrowserRuntimeInfo) (browserConcreteRouteAssessment, bool) {
	info, ok := b.dynamicRouteCacheInfo(requested)
	if !ok {
		return browserConcreteRouteAssessment{}, false
	}
	return b.dynamicRoutes.Load(info.Profile, info.Target)
}

func (b browserRuntimeRouterBackend) cachedRouteAssessment(requested BrowserRuntimeInfo) (browserConcreteRouteAssessment, bool) {
	requested = normalizeBrowserRuntimeInfo(requested)
	requested.Backend = ""
	if requested.Target == "" {
		if assessment, ok := b.cachedDefaultConcreteRouteAssessmentForRequestedProfile(requested); ok {
			return assessment, true
		}
	} else {
		if assessment, ok := b.cachedDefaultConcreteRouteAssessmentForRequestedTarget(requested); ok {
			return assessment, true
		}
		if assessment, ok := b.cachedTargetRouteAssessment(requested); ok {
			return assessment, true
		}
	}
	return b.cachedDynamicRouteAssessment(requested)
}

func (b browserRuntimeRouterBackend) cachedDefaultConcreteRouteAssessmentForRequestedTarget(requested BrowserRuntimeInfo) (browserConcreteRouteAssessment, bool) {
	requested = normalizeBrowserRuntimeInfo(requested)
	if strings.TrimSpace(requested.Target) == "" || strings.TrimSpace(requested.Profile) != "" {
		return browserConcreteRouteAssessment{}, false
	}
	assessment, ok := b.cachedDefaultConcreteRouteAssessment()
	if !ok || !assessment.RouteAvailable {
		return browserConcreteRouteAssessment{}, false
	}
	if normalizeBrowserRuntimeInfo(assessment.Route.RuntimeInfo).Target != requested.Target {
		return browserConcreteRouteAssessment{}, false
	}
	return assessment, true
}

func (b browserRuntimeRouterBackend) defaultRuntimeInfo() BrowserRuntimeInfo {
	return b.executionPreview().DefaultRoute
}

func (b browserRuntimeRouterBackend) baseDefaultRuntimeInfo() BrowserRuntimeInfo {
	sandboxPromotion := browserDefaultPromotionRouteAssessmentForConcreteRouteAssessment(b.sandboxRoute, "sandbox")
	return browserDefaultRouteRuntimeInfoForAssessment(browserDefaultSubstrateAssessment{
		HostRuntime:          b.hostRuntime,
		NodeRoute:            b.nodeRoute,
		SandboxRoute:         sandboxPromotion,
		SandboxConcreteRoute: b.sandboxRoute,
		DefaultRuntime: browserPromotedDefaultRuntimeInfoForAssessments(
			b.hostRuntime,
			b.nodeRoute,
			sandboxPromotion,
		),
		DefaultConcreteRoute: b.defaultRoute,
	})
}

func (b browserRuntimeRouterBackend) resolvedRouteCapabilities(route browserResolvedExecutionRoute) BrowserCapabilities {
	if route.Capabilities != (BrowserCapabilities{}) {
		return route.Capabilities
	}
	if route.Backend != nil {
		return browserCapabilitiesForConcreteBackend(route.Backend)
	}
	return BrowserCapabilities{}
}

func (b browserRuntimeRouterBackend) resolvedExecutionRouteFromAssessment(assessment browserConcreteRouteAssessment, unavailableMessage string) (browserResolvedExecutionRoute, error) {
	if !assessment.RouteAvailable {
		return browserResolvedExecutionRouteFromAssessment(assessment, unavailableMessage)
	}
	route := assessment.Route
	if route.Backend != nil || route.Capabilities != (BrowserCapabilities{}) {
		return route, nil
	}
	return browserResolvedExecutionRouteFromAssessment(assessment, unavailableMessage)
}

func (b browserRuntimeRouterBackend) storeDynamicRouteAssessment(requested BrowserRuntimeInfo, assessment browserConcreteRouteAssessment) {
	info, ok := b.dynamicRouteCacheInfo(requested)
	if !ok {
		return
	}
	b.dynamicRoutes.Store(info.Profile, info.Target, assessment)
}

func browserRuntimeRouterCachedRouteAssessment(backend BrowserBackend, requested BrowserRuntimeInfo) (browserConcreteRouteAssessment, bool) {
	switch routed := backend.(type) {
	case browserRuntimeRouterBackend:
		return routed.cachedRouteAssessment(requested)
	case *browserRuntimeRouterBackend:
		return routed.cachedRouteAssessment(requested)
	default:
		return browserConcreteRouteAssessment{}, false
	}
}

func browserRuntimeRouterCachedDefaultConcreteRouteAssessment(backend BrowserBackend) (browserConcreteRouteAssessment, bool) {
	switch routed := backend.(type) {
	case browserRuntimeRouterBackend:
		return routed.cachedDefaultConcreteRouteAssessment()
	case *browserRuntimeRouterBackend:
		return routed.cachedDefaultConcreteRouteAssessment()
	default:
		return browserConcreteRouteAssessment{}, false
	}
}

func browserRuntimeRouterPreferredDefaultConcreteRouteAssessment(backend BrowserBackend) (browserConcreteRouteAssessment, bool) {
	switch routed := backend.(type) {
	case browserRuntimeRouterBackend:
		return routed.preferredDefaultConcreteRouteAssessment()
	case *browserRuntimeRouterBackend:
		return routed.preferredDefaultConcreteRouteAssessment()
	default:
		return browserConcreteRouteAssessment{}, false
	}
}

func (b browserRuntimeRouterBackend) cachedDefaultConcreteRouteAssessment() (browserConcreteRouteAssessment, bool) {
	if assessment, ok := b.cachedImplicitManagedDefaultConcreteRouteAssessment(); ok {
		return assessment, true
	}
	if b.defaultRoute.Configured ||
		b.defaultRoute.RouteAvailable ||
		strings.TrimSpace(b.defaultRoute.FailureReason) != "" ||
		strings.TrimSpace(b.defaultRoute.FailureNote) != "" {
		return b.defaultRoute, true
	}
	return browserConcreteRouteAssessment{}, false
}

func (b browserRuntimeRouterBackend) preferredDefaultConcreteRouteAssessment() (browserConcreteRouteAssessment, bool) {
	if assessment, ok := b.cachedDefaultConcreteRouteAssessment(); ok && assessment.RouteAvailable {
		return assessment, true
	}
	return b.warmImplicitManagedDefaultConcreteRouteAssessment()
}

func (b browserRuntimeRouterBackend) warmImplicitManagedDefaultConcreteRouteAssessment() (browserConcreteRouteAssessment, bool) {
	if !b.usesImplicitLegacyHostDefaultFallback() {
		return browserConcreteRouteAssessment{}, false
	}
	for _, target := range []string{"node"} {
		candidate := BrowserRuntimeInfo{Target: target}
		if assessment, ok := b.cachedImplicitManagedDefaultRouteProbeAssessment(candidate); ok {
			if assessment.RouteAvailable && BrowserDefaultRuntimePromotionReady(assessment.Route.Capabilities) {
				return assessment, true
			}
			continue
		}
		route, err := b.resolveImplicitManagedDefaultConcreteRoute(candidate)
		if err != nil {
			b.storeImplicitManagedDefaultRouteAssessment(candidate, browserConcreteRouteAssessment{
				Configured:    browserImplicitManagedDefaultBackendConfigured(target, b),
				FailureReason: err.Error(),
				FailureNote:   err.Error(),
			})
			continue
		}
		assessment := browserConcreteRouteAssessment{
			Configured:     true,
			RouteAvailable: true,
			Route:          route,
		}
		b.storeImplicitManagedDefaultRouteAssessment(candidate, assessment)
		if BrowserDefaultRuntimePromotionReady(route.Capabilities) {
			return assessment, true
		}
	}
	return browserConcreteRouteAssessment{}, false
}

func (b browserRuntimeRouterBackend) cachedImplicitManagedDefaultRouteProbeAssessment(requested BrowserRuntimeInfo) (browserConcreteRouteAssessment, bool) {
	info, ok := browserRuntimeImplicitManagedDefaultCacheInfo(requested)
	if !ok {
		return browserConcreteRouteAssessment{}, false
	}
	return b.implicitManagedDefaultRoute.Load(info.Profile, info.Target)
}

func (b browserRuntimeRouterBackend) cachedImplicitManagedDefaultConcreteRouteAssessment() (browserConcreteRouteAssessment, bool) {
	if !b.usesImplicitLegacyHostDefaultFallback() {
		return browserConcreteRouteAssessment{}, false
	}
	for _, target := range []string{"node"} {
		if assessment, ok := b.implicitManagedDefaultRoute.Load("__default__", target); ok &&
			assessment.RouteAvailable &&
			BrowserDefaultRuntimePromotionReady(assessment.Route.Capabilities) {
			return assessment, true
		}
	}
	return browserConcreteRouteAssessment{}, false
}

func browserRouterCachedRouteAssessmentForRequestedProfile(profile string, cachedProfile string, assessment browserConcreteRouteAssessment) (browserConcreteRouteAssessment, bool) {
	if profile == "" {
		if assessment.RouteAvailable || strings.TrimSpace(assessment.FailureReason) != "" || strings.TrimSpace(assessment.FailureNote) != "" {
			return assessment, true
		}
		return browserConcreteRouteAssessment{}, false
	}
	return browserCachedRouteAssessmentForProfile(profile, cachedProfile, assessment)
}

func (b browserRuntimeRouterBackend) resolveConcreteBrowserRoute(backend BrowserBackend, fallback BrowserRuntimeInfo, requested BrowserRuntimeInfo) (browserResolvedExecutionRoute, error) {
	return resolveConcreteBrowserExecutionRoute(backend, fallback, requested)
}

func (b browserRuntimeRouterBackend) resolveConcreteBrowserRouteWithDynamicCache(backend BrowserBackend, fallback BrowserRuntimeInfo, requested BrowserRuntimeInfo) (browserResolvedExecutionRoute, error) {
	if assessment, ok := b.cachedDynamicRouteAssessment(requested); ok {
		return b.resolvedExecutionRouteFromAssessment(assessment, "browser route is unavailable")
	}
	route, err := b.resolveConcreteBrowserRoute(backend, fallback, requested)
	if err != nil {
		b.storeDynamicRouteAssessment(requested, browserConcreteRouteAssessment{
			Configured:    backend != nil,
			FailureReason: err.Error(),
			FailureNote:   err.Error(),
		})
		return browserResolvedExecutionRoute{}, err
	}
	b.storeDynamicRouteAssessment(requested, browserConcreteRouteAssessment{
		Configured:     backend != nil,
		RouteAvailable: true,
		Route:          route,
	})
	return route, nil
}

func resolveConcreteBrowserExecutionRoute(backend BrowserBackend, fallback BrowserRuntimeInfo, requested BrowserRuntimeInfo) (browserResolvedExecutionRoute, error) {
	laneTarget := normalizeBrowserRuntimeInfo(fallback).Target
	info := browserRuntimeInfoForConcreteBackend(backend, fallback)
	if info.Target == "" {
		info.Target = laneTarget
	}
	requestedInfo := mergeBrowserRuntimeInfo(info, requested)
	if laneTarget != "" {
		requestedInfo.Target = laneTarget
	}
	if resolver, ok := backend.(browserExecutionRouteResolver); ok {
		route, err := resolver.ResolveBrowserExecutionRoute(requestedInfo)
		if err != nil {
			return browserResolvedExecutionRoute{}, err
		}
		route = browserNormalizeResolvedExecutionRoute(route, requestedInfo, requestedInfo.Profile != "", requestedInfo.Target != "")
		if laneTarget != "" {
			route.RuntimeInfo.Target = laneTarget
		}
		return route, nil
	}
	if resolver, ok := backend.(BrowserRuntimeRouteResolver); ok {
		resolved, err := resolver.ResolveBrowserRuntimeRoute(requestedInfo)
		if err != nil {
			return browserResolvedExecutionRoute{}, err
		}
		info = mergeBrowserRuntimeInfo(info, resolved)
	} else {
		info = requestedInfo
	}
	if laneTarget != "" {
		info.Target = laneTarget
	}
	return browserResolvedExecutionRoute{
		Backend:      backend,
		RuntimeInfo:  info,
		Capabilities: browserCapabilitiesForConcreteBackend(backend),
	}, nil
}

func browserResolvedExecutionRouteFromAssessment(assessment browserConcreteRouteAssessment, unavailableMessage string) (browserResolvedExecutionRoute, error) {
	if assessment.RouteAvailable {
		return assessment.Route, nil
	}
	if note := strings.TrimSpace(assessment.FailureNote); note != "" {
		return browserResolvedExecutionRoute{}, fmt.Errorf("%s", note)
	}
	if reason := strings.TrimSpace(assessment.FailureReason); reason != "" {
		return browserResolvedExecutionRoute{}, fmt.Errorf("%s", reason)
	}
	return browserResolvedExecutionRoute{}, fmt.Errorf("%s", unavailableMessage)
}

func mergeBrowserRuntimeInfo(base BrowserRuntimeInfo, overlay BrowserRuntimeInfo) BrowserRuntimeInfo {
	base = normalizeBrowserRuntimeInfo(base)
	overlay = normalizeBrowserRuntimeInfo(overlay)
	if overlay.Backend != "" {
		base.Backend = overlay.Backend
	}
	if overlay.Profile != "" {
		base.Profile = overlay.Profile
	}
	if overlay.Target != "" {
		base.Target = overlay.Target
	}
	return base
}

func browserCapabilitiesForConcreteBackend(backend BrowserBackend) BrowserCapabilities {
	if provider, ok := backend.(BrowserCapabilityProvider); ok {
		return provider.BrowserCapabilities()
	}
	if backend != nil {
		return fullBrowserCapabilities()
	}
	return BrowserCapabilities{}
}
