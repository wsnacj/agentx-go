package tools

func browserRuntimeRouterHostBackend(host BrowserBackend, fallbackPolicy outboundNetworkPolicy, fallbackTimeoutMs int) (BrowserBackend, outboundNetworkPolicy, int) {
	return host, fallbackPolicy, fallbackTimeoutMs
}

func browserRuntimeRouterHostDeferredAssessment(assessment browserConcreteRouteAssessment, hostBackend BrowserBackend) browserConcreteRouteAssessment {
	if !assessment.RouteAvailable {
		return assessment
	}
	route := assessment.Route
	route.RuntimeInfo = normalizeBrowserRuntimeInfo(route.RuntimeInfo)
	if route.Backend == nil && BrowserSubstratePosture(route.RuntimeInfo.Backend, route.RuntimeInfo.Target) == BrowserSubstrateLegacySystemHost {
		route.Backend = hostBackend
		if route.Capabilities == (BrowserCapabilities{}) {
			route.Capabilities = browserCapabilitiesForConcreteBackend(route.Backend)
		}
		assessment.Route = route
	}
	return assessment
}

func (b browserRuntimeRouterBackend) BrowserCapabilities() BrowserCapabilities {
	capabilities := BrowserCapabilities{}
	if assessment, ok := b.cachedDefaultConcreteRouteAssessment(); ok && assessment.RouteAvailable {
		capabilities = mergeBrowserCapabilities(capabilities, b.resolvedRouteCapabilities(assessment.Route))
	}
	if assessment, ok := b.cachedImplicitManagedDefaultRouteProbeAssessment(BrowserRuntimeInfo{Target: "node"}); ok && assessment.RouteAvailable {
		capabilities = mergeBrowserCapabilities(capabilities, b.resolvedRouteCapabilities(assessment.Route))
	}
	if b.hostRoute.RouteAvailable {
		capabilities = mergeBrowserCapabilities(capabilities, b.resolvedRouteCapabilities(b.hostRoute.Route))
	}
	if b.sandboxRoute.RouteAvailable {
		capabilities = mergeBrowserCapabilities(capabilities, b.resolvedRouteCapabilities(b.sandboxRoute.Route))
	}
	if b.nodeRoute.RouteAvailable {
		capabilities = mergeBrowserCapabilities(capabilities, b.resolvedRouteCapabilities(b.nodeRoute.Route))
	}
	return capabilities
}

func (b browserRuntimeRouterBackend) BrowserRuntimeInfo() BrowserRuntimeInfo {
	return b.executionPreview().DefaultRoute
}
