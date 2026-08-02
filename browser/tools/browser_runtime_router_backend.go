package tools

type browserRuntimeRouterBackend struct {
	hostBackend                 BrowserBackend
	hostRuntime                 BrowserRuntimeInfo
	hostPolicy                  outboundNetworkPolicy
	hostTimeoutMs               int
	sessionRegistry             *BrowserSessionRegistry
	stateRegistry               *BrowserSessionStateRegistry
	defaultRoute                browserConcreteRouteAssessment
	hostRoute                   browserConcreteRouteAssessment
	sandboxBackend              BrowserBackend
	sandboxRoute                browserConcreteRouteAssessment
	nodeBackend                 BrowserBackend
	nodeRoute                   browserDefaultPromotionRouteAssessment
	dynamicRoutes               *browserRuntimeRouteAssessmentCache
	implicitManagedDefaultRoute *browserRuntimeRouteAssessmentCache
}
