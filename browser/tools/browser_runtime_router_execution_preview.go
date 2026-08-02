package tools

import "strings"

type browserRuntimeRouterExecutionPreview struct {
	DefaultRoute                  BrowserRuntimeInfo
	DefaultCandidateRoute         BrowserRuntimeInfo
	DefaultCandidateDescriptor    browserRuntimeRouteDescriptor
	DispatchBase                  BrowserRuntimeInfo
	HiddenImplicitHostDefaultBase bool
	DefaultTarget                 string
}

func (b browserRuntimeRouterBackend) executionPreview() browserRuntimeRouterExecutionPreview {
	defaultRoute := normalizeBrowserRuntimeInfo(b.baseDefaultRuntimeInfo())
	defaultConcreteRouteAvailable := false
	if assessment, ok := b.preferredDefaultConcreteRouteAssessment(); ok && assessment.RouteAvailable {
		defaultRoute = normalizeBrowserRuntimeInfo(assessment.Route.RuntimeInfo)
		defaultConcreteRouteAvailable = true
	}
	dispatchBase := defaultRoute
	hiddenImplicitHostDefaultBase := false
	if BrowserSubstratePosture(normalizeBrowserRuntimeInfo(b.baseDefaultRuntimeInfo()).Backend, normalizeBrowserRuntimeInfo(b.baseDefaultRuntimeInfo()).Target) == BrowserSubstrateLegacySystemHost &&
		BrowserSubstratePosture(defaultRoute.Backend, defaultRoute.Target) == BrowserSubstrateLegacySystemHost &&
		!defaultConcreteRouteAvailable {
		dispatchBase = BrowserRuntimeInfo{}
		hiddenImplicitHostDefaultBase = true
	}
	defaultCandidateRoute := b.executionPreviewDefaultCandidateRoute(defaultRoute, dispatchBase, hiddenImplicitHostDefaultBase)
	return browserRuntimeRouterExecutionPreview{
		DefaultRoute:                  defaultRoute,
		DefaultCandidateRoute:         defaultCandidateRoute,
		DefaultCandidateDescriptor:    b.executionPreviewDefaultCandidateDescriptor(defaultCandidateRoute, hiddenImplicitHostDefaultBase),
		DispatchBase:                  normalizeBrowserRuntimeInfo(dispatchBase),
		HiddenImplicitHostDefaultBase: hiddenImplicitHostDefaultBase,
		DefaultTarget:                 strings.TrimSpace(defaultRoute.Target),
	}
}

func (b browserRuntimeRouterBackend) executionPreviewDefaultCandidateRoute(
	defaultRoute BrowserRuntimeInfo,
	dispatchBase BrowserRuntimeInfo,
	hiddenImplicitHostDefaultBase bool,
) BrowserRuntimeInfo {
	dispatchBase = normalizeBrowserRuntimeInfo(dispatchBase)
	if dispatchBase != (BrowserRuntimeInfo{}) {
		return dispatchBase
	}
	defaultRoute = normalizeBrowserRuntimeInfo(defaultRoute)
	if !hiddenImplicitHostDefaultBase {
		return defaultRoute
	}
	if assessment, ok := b.cachedImplicitManagedDefaultConcreteRouteAssessment(); ok {
		if info := normalizeBrowserRuntimeInfo(assessment.Route.RuntimeInfo); info != (BrowserRuntimeInfo{}) {
			return info
		}
	}
	for _, target := range []string{"node", "sandbox"} {
		if !browserImplicitManagedDefaultBackendConfigured(target, b) {
			continue
		}
		fallback := defaultBrowserNodeRuntimeInfo()
		if target == "sandbox" {
			fallback = defaultBrowserSandboxRuntimeInfo()
		}
		if info := normalizeBrowserRuntimeInfo(browserRuntimeInfoForConcreteBackend(
			browserRuntimeRouterManagedTargetBackend(b, target),
			fallback,
		)); info != (BrowserRuntimeInfo{}) {
			return info
		}
	}
	return BrowserRuntimeInfo{}
}

func browserRuntimeRouteDescriptorValueFromResolvedRoute(route browserResolvedExecutionRoute) browserRuntimeRouteDescriptor {
	if descriptor := browserRuntimeRouteDescriptorPtrFromResolvedRoute(route); descriptor != nil {
		return *descriptor
	}
	return browserRuntimeRouteDescriptor{}
}

func browserRuntimeRouteDescriptorForAssessmentInfo(
	info BrowserRuntimeInfo,
	assessment browserConcreteRouteAssessment,
) browserRuntimeRouteDescriptor {
	info = normalizeBrowserRuntimeInfo(info)
	if info == (BrowserRuntimeInfo{}) || !assessment.RouteAvailable {
		return browserRuntimeRouteDescriptor{}
	}
	if normalizeBrowserRuntimeInfo(assessment.Route.RuntimeInfo) != info {
		return browserRuntimeRouteDescriptor{}
	}
	return browserRuntimeRouteDescriptorValueFromResolvedRoute(assessment.Route)
}

func (b browserRuntimeRouterBackend) executionPreviewDefaultCandidateBackend(info BrowserRuntimeInfo) BrowserBackend {
	switch strings.ToLower(strings.TrimSpace(info.Target)) {
	case "host":
		return b.hostBackend
	case "node":
		return b.nodeBackend
	case "sandbox":
		return b.sandboxBackend
	default:
		return nil
	}
}

func (b browserRuntimeRouterBackend) executionPreviewDefaultCandidateDescriptor(
	defaultCandidateRoute BrowserRuntimeInfo,
	hiddenImplicitHostDefaultBase bool,
) browserRuntimeRouteDescriptor {
	defaultCandidateRoute = normalizeBrowserRuntimeInfo(defaultCandidateRoute)
	if defaultCandidateRoute == (BrowserRuntimeInfo{}) {
		return browserRuntimeRouteDescriptor{}
	}
	for _, assessment := range []browserConcreteRouteAssessment{
		b.defaultRoute,
		b.hostRoute,
		browserConcreteRouteAssessmentForDefaultPromotion(b.nodeRoute),
		b.sandboxRoute,
	} {
		if descriptor := browserRuntimeRouteDescriptorForAssessmentInfo(defaultCandidateRoute, assessment); descriptor != (browserRuntimeRouteDescriptor{}) {
			return descriptor
		}
	}
	if hiddenImplicitHostDefaultBase {
		if assessment, ok := b.cachedImplicitManagedDefaultConcreteRouteAssessment(); ok {
			if descriptor := browserRuntimeRouteDescriptorForAssessmentInfo(defaultCandidateRoute, assessment); descriptor != (browserRuntimeRouteDescriptor{}) {
				return descriptor
			}
		}
	}
	metadata := browserDoctorRouteMetadataForBackend(b.executionPreviewDefaultCandidateBackend(defaultCandidateRoute))
	return browserRuntimeRouteDescriptorFromInfoWithProvenance(
		defaultCandidateRoute,
		metadata.Source,
		metadata.Endpoint,
	)
}

func (b browserRuntimeRouterBackend) ResolveBrowserDefaultRequestBase(params map[string]any, base BrowserRuntimeInfo) BrowserRuntimeInfo {
	executionPreview := b.executionPreview()
	base = normalizeBrowserRuntimeInfo(base)
	if !browserRuntimeRouterCanReuseExecutionPreviewForDefaultRequest(params, base) {
		return base
	}
	if base == (BrowserRuntimeInfo{}) {
		return executionPreview.DispatchBase
	}
	if base == executionPreview.DispatchBase {
		return base
	}
	if base == executionPreview.DefaultRoute {
		return executionPreview.DispatchBase
	}
	if BrowserSubstratePosture(base.Backend, base.Target) == BrowserSubstrateLegacySystemHost {
		return executionPreview.DispatchBase
	}
	return base
}

func (b browserRuntimeRouterBackend) ResolveBrowserDefaultRequestPreview(params map[string]any, base BrowserRuntimeInfo, hiddenImplicitHostDefaultBase bool) (BrowserRuntimeInfo, bool) {
	base = normalizeBrowserRuntimeInfo(base)
	if !browserRuntimeRouterCanReuseExecutionPreviewForDefaultRequest(params, base) {
		return base, hiddenImplicitHostDefaultBase
	}
	if hiddenImplicitHostDefaultBase {
		sessionExecutionPreview := b.sessionExecutionPreview(base, hiddenImplicitHostDefaultBase)
		return sessionExecutionPreview.Base, sessionExecutionPreview.HiddenImplicitHostDefaultBase
	}
	return b.ResolveBrowserDefaultRequestBase(params, base), hiddenImplicitHostDefaultBase
}

func browserRuntimeRouterCanReuseExecutionPreviewForDefaultRequest(params map[string]any, base BrowserRuntimeInfo) bool {
	if strings.TrimSpace(firstString(params, "target", "url")) != "" {
		return false
	}
	if firstInt(params, "tab_index", "index") > 0 {
		return false
	}
	if browserHasKey(params, "tab_index") || browserHasKey(params, "index") || browserHasKey(params, "browser_app") || browserHasKey(params, "browser") || browserHasKey(params, "app") {
		return base == (BrowserRuntimeInfo{})
	}
	return true
}

func browserHasKey(params map[string]any, keys ...string) bool {
	for _, key := range keys {
		if _, ok := params[key]; ok {
			return true
		}
	}
	return false
}
