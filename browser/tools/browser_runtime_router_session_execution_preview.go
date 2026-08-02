package tools

type browserRuntimeRouterSessionExecutionPreview struct {
	ExecutionPreview              browserRuntimeRouterExecutionPreview
	DefaultCandidateRoute         BrowserRuntimeInfo
	DefaultCandidateDescriptor    browserRuntimeRouteDescriptor
	Base                          BrowserRuntimeInfo
	HiddenImplicitHostDefaultBase bool
}

func (b browserRuntimeRouterBackend) sessionExecutionPreview(base BrowserRuntimeInfo, hiddenImplicitHostDefaultBase bool) browserRuntimeRouterSessionExecutionPreview {
	executionPreview := b.executionPreview()
	base = normalizeBrowserRuntimeInfo(base)
	if hiddenImplicitHostDefaultBase && b.sessionExecutionPreviewUsesDefaultBase(base, executionPreview) {
		base = executionPreview.DispatchBase
		hiddenImplicitHostDefaultBase = executionPreview.HiddenImplicitHostDefaultBase
	}
	return browserRuntimeRouterSessionExecutionPreview{
		ExecutionPreview:              executionPreview,
		DefaultCandidateRoute:         normalizeBrowserRuntimeInfo(executionPreview.DefaultCandidateRoute),
		DefaultCandidateDescriptor:    executionPreview.DefaultCandidateDescriptor,
		Base:                          base,
		HiddenImplicitHostDefaultBase: hiddenImplicitHostDefaultBase,
	}
}

func (b browserRuntimeRouterBackend) sessionExecutionPreviewUsesDefaultBase(base BrowserRuntimeInfo, executionPreview browserRuntimeRouterExecutionPreview) bool {
	base = normalizeBrowserRuntimeInfo(base)
	if base == (BrowserRuntimeInfo{}) {
		return false
	}
	if BrowserSubstratePosture(base.Backend, base.Target) == BrowserSubstrateLegacySystemHost {
		return true
	}
	if base == normalizeBrowserRuntimeInfo(b.baseDefaultRuntimeInfo()) {
		return true
	}
	if base == executionPreview.DefaultRoute {
		return true
	}
	return base == executionPreview.DispatchBase
}
