package tools

import (
	stdruntime "runtime"
)

func fullBrowserCapabilities() BrowserCapabilities {
	return BrowserCapabilities{
		Open:       true,
		Navigate:   true,
		Tabs:       true,
		Extract:    true,
		Snapshot:   true,
		Screenshot: true,
		Click:      true,
		TypeText:   true,
		Evaluate:   true,
		Wait:       true,
	}
}

func defaultBrowserCapabilities() BrowserCapabilities {
	if stdruntime.GOOS == "darwin" {
		return fullBrowserCapabilities()
	}
	return BrowserCapabilities{
		Open:     true,
		Extract:  true,
		Snapshot: true,
		Wait:     true,
	}
}

func browserCapabilitiesForRegistration(opts BrowserToolOptions) BrowserCapabilities {
	preview := browserDefaultRuntimePreviewForToolOptions(opts)
	return preview.RegistrationCapabilities
}

func browserCapabilitiesForRegistrationWithAssessment(substrate browserDefaultSubstrateAssessment) BrowserCapabilities {
	return browserCapabilitiesForRegistrationWithBackend(nil, substrate)
}

func browserCapabilitiesForRegistrationWithBackend(backend BrowserBackend, substrate browserDefaultSubstrateAssessment) BrowserCapabilities {
	hostAssessment := browserCapabilitiesRegistrationRouteAssessment(
		backend,
		BrowserRuntimeInfo{Target: "host"},
		substrate.HostRoute,
	)
	nodeAssessment := browserCapabilitiesRegistrationRouteAssessment(
		backend,
		BrowserRuntimeInfo{Target: "node"},
		browserConcreteRouteAssessmentForDefaultPromotion(substrate.NodeRoute),
	)
	sandboxAssessment := browserCapabilitiesRegistrationRouteAssessment(
		backend,
		BrowserRuntimeInfo{Target: "sandbox"},
		substrate.SandboxConcreteRoute,
	)
	if browserRegistrationSurfaceBlockedByHostFailureWithAssessments(substrate.HostRuntime, hostAssessment, nodeAssessment, sandboxAssessment) {
		return BrowserCapabilities{}
	}
	capabilities := BrowserCapabilities{}
	if hostAssessment.RouteAvailable {
		capabilities = hostAssessment.Route.Capabilities
	}
	if sandboxAssessment.RouteAvailable {
		capabilities = mergeBrowserCapabilities(capabilities, sandboxAssessment.Route.Capabilities)
	}
	if nodeAssessment.RouteAvailable {
		capabilities = mergeBrowserCapabilities(capabilities, nodeAssessment.Route.Capabilities)
	}
	return capabilities
}

func browserRegistrationSurfaceBlockedByHostFailure(substrate browserDefaultSubstrateAssessment) bool {
	return browserRegistrationSurfaceBlockedByHostFailureWithAssessments(
		substrate.HostRuntime,
		substrate.HostRoute,
		browserConcreteRouteAssessmentForDefaultPromotion(substrate.NodeRoute),
		substrate.SandboxConcreteRoute,
	)
}

func browserRegistrationSurfaceBlockedByHostFailureWithAssessments(hostRuntime BrowserRuntimeInfo, hostAssessment browserConcreteRouteAssessment, nodeAssessment browserConcreteRouteAssessment, sandboxAssessment browserConcreteRouteAssessment) bool {
	if !hostAssessment.Configured || hostAssessment.RouteAvailable {
		return false
	}
	if BrowserSubstratePosture(hostRuntime.Backend, hostRuntime.Target) == BrowserSubstrateLegacySystemHost {
		return false
	}
	// Keep the specialist surface visible when an explicit managed lane is still
	// resolvable, even if the custom host default route is broken.
	return !nodeAssessment.RouteAvailable && !sandboxAssessment.RouteAvailable
}

func browserCapabilitiesRegistrationRouteAssessment(backend BrowserBackend, requested BrowserRuntimeInfo, fallback browserConcreteRouteAssessment) browserConcreteRouteAssessment {
	if assessment, ok := browserRuntimeRouterCachedRouteAssessment(backend, requested); ok {
		return assessment
	}
	return fallback
}

func browserRegistrationHostBackend(opts BrowserToolOptions) BrowserBackend {
	return browserConfiguredHostBackendForOptions(opts)
}

func browserCapabilitiesForConcreteRoutePreview(backend BrowserBackend, fallback BrowserRuntimeInfo) (BrowserCapabilities, bool) {
	assessment := browserConcreteRouteAssessmentForConcreteBackend(backend, fallback, normalizeBrowserRuntimeInfo(fallback).Target)
	if !assessment.RouteAvailable {
		return BrowserCapabilities{}, false
	}
	return assessment.Route.Capabilities, true
}
