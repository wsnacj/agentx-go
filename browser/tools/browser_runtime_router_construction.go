package tools

func newBrowserRuntimeRouterBackend(opts BrowserToolOptions, host BrowserBackend) BrowserBackend {
	substrate := browserDefaultSubstrateAssessmentForHostBackend(opts, host)
	return newBrowserRuntimeRouterBackendWithAssessment(opts, host, substrate)
}

func newBrowserRuntimeRouterBackendWithAssessment(opts BrowserToolOptions, host BrowserBackend, substrate browserDefaultSubstrateAssessment) BrowserBackend {
	policy, timeoutMs := browserDefaultRouteOwnerDispatchOptionsForToolOptions(opts)
	return newBrowserRuntimeRouterBackendWithDispatchOptions(opts, host, policy, timeoutMs, substrate)
}

func newBrowserRuntimeRouterBackendWithDispatchOptions(opts BrowserToolOptions, host BrowserBackend, hostPolicy outboundNetworkPolicy, hostTimeoutMs int, substrate browserDefaultSubstrateAssessment) BrowserBackend {
	hostBackend, hostPolicy, hostTimeoutMs := browserRuntimeRouterHostBackend(host, hostPolicy, hostTimeoutMs)
	router := browserRuntimeRouterBackend{
		hostBackend:                 hostBackend,
		hostRuntime:                 substrate.HostRuntime,
		hostPolicy:                  hostPolicy,
		hostTimeoutMs:               hostTimeoutMs,
		sessionRegistry:             opts.SessionRegistry,
		stateRegistry:               opts.SessionStateRegistry,
		defaultRoute:                browserRuntimeRouterHostDeferredAssessment(substrate.DefaultConcreteRoute, hostBackend),
		hostRoute:                   browserRuntimeRouterHostDeferredAssessment(substrate.HostRoute, hostBackend),
		sandboxBackend:              opts.SandboxBackend,
		sandboxRoute:                substrate.SandboxConcreteRoute,
		nodeBackend:                 opts.NodeBackend,
		nodeRoute:                   substrate.NodeRoute,
		dynamicRoutes:               newBrowserRuntimeRouteAssessmentCache(),
		implicitManagedDefaultRoute: newBrowserRuntimeRouteAssessmentCache(),
	}
	return router
}
