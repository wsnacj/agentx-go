package tools

import (
	"fmt"
	"strings"
)

func defaultBrowserRuntimeInfo() BrowserRuntimeInfo {
	return BrowserRuntimeInfo{
		Backend: "system",
		Profile: "default",
		Target:  "host",
	}
}

func DefaultBrowserRuntimeInfo() BrowserRuntimeInfo {
	return defaultBrowserRuntimeInfo()
}

func defaultBrowserNodeRuntimeInfo() BrowserRuntimeInfo {
	return BrowserRuntimeInfo{
		Backend: "node",
		Profile: "isolated",
		Target:  "node",
	}
}

func DefaultBrowserNodeRuntimeInfo() BrowserRuntimeInfo {
	return defaultBrowserNodeRuntimeInfo()
}

func defaultBrowserSandboxRuntimeInfo() BrowserRuntimeInfo {
	return BrowserRuntimeInfo{
		Backend: "sandbox",
		Profile: "default",
		Target:  "sandbox",
	}
}

func DefaultBrowserSandboxRuntimeInfo() BrowserRuntimeInfo {
	return defaultBrowserSandboxRuntimeInfo()
}

func defaultBrowserProxyNodeRuntimeInfo() BrowserRuntimeInfo {
	info := defaultBrowserNodeRuntimeInfo()
	info.Backend = "proxy"
	return info
}

func DefaultBrowserProxyNodeRuntimeInfo() BrowserRuntimeInfo {
	return defaultBrowserProxyNodeRuntimeInfo()
}

func BrowserDefaultRuntimePromotionReady(capabilities BrowserCapabilities) bool {
	return capabilities.Open &&
		capabilities.Navigate &&
		capabilities.Tabs &&
		capabilities.Extract &&
		capabilities.Snapshot &&
		capabilities.Screenshot &&
		capabilities.Click &&
		capabilities.TypeText &&
		capabilities.Evaluate
}

func browserRuntimeOnlySurfaceEnabled(opts BrowserToolOptions) bool {
	enabled := buildEnabledToolSet(opts.EnabledTools)
	if len(enabled) == 0 {
		return false
	}
	hasBrowserTool := false
	for name := range enabled {
		if !isBrowserToolName(name) {
			continue
		}
		hasBrowserTool = true
		if name != "browser_runtime" {
			return false
		}
	}
	return hasBrowserTool
}

func browserSurfaceDefaultRoutePromotionReady(opts BrowserToolOptions, capabilities BrowserCapabilities) bool {
	if BrowserDefaultRuntimePromotionReady(capabilities) {
		return true
	}
	return browserRuntimeOnlySurfaceEnabled(opts) && capabilities.RuntimeStatus
}

func browserSurfacePromoteDefaultSubstrateAssessment(assessment browserDefaultSubstrateAssessment, route browserResolvedExecutionRoute) browserDefaultSubstrateAssessment {
	route.RuntimeInfo = normalizeBrowserRuntimeInfo(route.RuntimeInfo)
	if route.RuntimeInfo == (BrowserRuntimeInfo{}) {
		return assessment
	}
	assessment.DefaultRuntime = route.RuntimeInfo
	assessment.DefaultConcreteRoute = browserConcreteRouteAssessment{
		Configured:     true,
		RouteAvailable: true,
		Route:          route,
	}
	switch route.RuntimeInfo.Target {
	case "node":
		assessment.NodeRoute = browserDefaultPromotionRouteAssessment{
			Configured:     true,
			RouteAvailable: true,
			Ready:          true,
			Route:          route,
		}
	case "sandbox":
		assessment.SandboxConcreteRoute = browserConcreteRouteAssessment{
			Configured:     true,
			RouteAvailable: true,
			Route:          route,
		}
		assessment.SandboxRoute = browserDefaultPromotionRouteAssessment{
			Configured:     true,
			RouteAvailable: true,
			Ready:          true,
			Route:          route,
		}
	}
	return assessment
}

func browserSurfacePromotedDefaultSubstrateAssessmentForBackend(
	opts BrowserToolOptions,
	backend BrowserBackend,
	assessment browserDefaultSubstrateAssessment,
) (browserDefaultSubstrateAssessment, bool) {
	if !browserRuntimeUsesImplicitLegacyHostDefaultFallback(browserDefaultRouteRuntimeInfoForAssessment(assessment), assessment) {
		return assessment, false
	}
	resolver, ok := backend.(interface {
		ResolveBrowserExecutionRoute(BrowserRuntimeInfo) (browserResolvedExecutionRoute, error)
	})
	if !ok {
		return assessment, false
	}
	route, err := resolver.ResolveBrowserExecutionRoute(BrowserRuntimeInfo{})
	if err != nil || !browserSurfaceDefaultRoutePromotionReady(opts, route.Capabilities) {
		return assessment, false
	}
	return browserSurfacePromoteDefaultSubstrateAssessment(assessment, route), true
}

func BrowserPromoteDefaultRuntimeInfo(hostInfo BrowserRuntimeInfo, nodeInfo BrowserRuntimeInfo, nodePromotable bool, sandboxInfo BrowserRuntimeInfo, sandboxPromotable bool) BrowserRuntimeInfo {
	hostInfo = normalizeBrowserRuntimeInfo(hostInfo)
	nodeInfo = normalizeBrowserRuntimeInfo(nodeInfo)
	_ = sandboxInfo
	_ = sandboxPromotable
	if BrowserSubstratePosture(hostInfo.Backend, hostInfo.Target) != BrowserSubstrateLegacySystemHost {
		return hostInfo
	}
	if nodePromotable {
		return nodeInfo
	}
	return hostInfo
}

func BrowserDefaultRuntimeInfoForOptions(opts BrowserToolOptions) BrowserRuntimeInfo {
	return browserDefaultRuntimePreviewForToolOptions(opts).LogicalDefaultRoute
}

func browserConfiguredHostBackendForOptions(opts BrowserToolOptions) BrowserBackend {
	if opts.Backend != nil {
		return opts.Backend
	}
	return opts.ImplicitHostBackend
}

func browserHostBackendForOptions(opts BrowserToolOptions, policy outboundNetworkPolicy, timeoutMs int) BrowserBackend {
	_ = policy
	_ = timeoutMs
	return browserConfiguredHostBackendForOptions(opts)
}

type browserDefaultRouteOwner struct {
	backend             BrowserBackend
	substrateAssessment browserDefaultSubstrateAssessment
	substrateSummary    BrowserWorkbenchSubstrateSummary
	defaultRoute        BrowserRuntimeInfo
}

func browserDefaultRouteOwnerForOptions(opts BrowserToolOptions, policy outboundNetworkPolicy, timeoutMs int) browserDefaultRouteOwner {
	hostBackend := browserHostBackendForOptions(opts, policy, timeoutMs)
	substrateAssessment := browserDefaultSubstrateAssessmentForHostBackend(opts, hostBackend)
	backend := newBrowserRuntimeRouterBackendWithDispatchOptions(opts, hostBackend, policy, timeoutMs, substrateAssessment)
	substrateAssessment = browserSubstrateAssessmentForBackend(backend, substrateAssessment)
	return browserDefaultRouteOwner{
		backend:             backend,
		substrateAssessment: substrateAssessment,
		substrateSummary:    browserWorkbenchSubstrateSummaryForBackend(opts, backend, substrateAssessment),
		defaultRoute:        browserDefaultRouteRuntimeInfoForAssessment(substrateAssessment),
	}
}

func browserDefaultRouteOwnerDispatchOptionsForToolOptions(opts BrowserToolOptions) (outboundNetworkPolicy, int) {
	timeoutMs := opts.TimeoutMs
	if timeoutMs <= 0 {
		timeoutMs = 20_000
	}
	policy, err := newOutboundNetworkPolicy(outboundNetworkOptions{
		AllowPrivateHosts: opts.AllowPrivateHosts,
		AllowCIDRs:        append([]string(nil), opts.AllowCIDRs...),
		DenyCIDRs:         append([]string(nil), opts.DenyCIDRs...),
		AllowPorts:        append([]int(nil), opts.AllowPorts...),
		DenyPorts:         append([]int(nil), opts.DenyPorts...),
		DefaultDenyCIDRs:  defaultOutboundDeniedCIDRs,
		DefaultDenyPorts:  defaultOutboundDeniedPorts,
	})
	if err != nil {
		return outboundNetworkPolicy{}, timeoutMs
	}
	return policy, timeoutMs
}

func browserDefaultRouteOwnerForToolOptions(opts BrowserToolOptions) browserDefaultRouteOwner {
	policy, timeoutMs := browserDefaultRouteOwnerDispatchOptionsForToolOptions(opts)
	return browserDefaultRouteOwnerForOptions(opts, policy, timeoutMs)
}

func browserSurfaceDefaultRouteOwnerForToolOptions(opts BrowserToolOptions) browserDefaultRouteOwner {
	policy, timeoutMs := browserDefaultRouteOwnerDispatchOptionsForToolOptions(opts)
	return browserSurfaceDefaultRouteOwnerForOptions(opts, policy, timeoutMs)
}

func browserSurfaceDefaultRouteOwnerForOptions(opts BrowserToolOptions, policy outboundNetworkPolicy, timeoutMs int) browserDefaultRouteOwner {
	owner := browserDefaultRouteOwnerForOptions(opts, policy, timeoutMs)
	if !browserRuntimeUsesImplicitLegacyHostDefaultFallback(owner.defaultRoute, owner.substrateAssessment) {
		return owner
	}
	assessment, promoted := browserSurfacePromotedDefaultSubstrateAssessmentForBackend(opts, owner.backend, owner.substrateAssessment)
	if !promoted {
		return owner
	}
	owner.substrateAssessment = assessment
	owner.substrateSummary = browserWorkbenchSubstrateSummaryForAssessmentWithBackend(opts, owner.backend, owner.substrateAssessment)
	owner.defaultRoute = browserDefaultRouteRuntimeInfoForAssessment(owner.substrateAssessment)
	return owner
}

func browserVisibleDefaultRouteRuntimeInfo(defaultRoute BrowserRuntimeInfo, substrate browserDefaultSubstrateAssessment) BrowserRuntimeInfo {
	defaultRoute = normalizeBrowserRuntimeInfo(defaultRoute)
	if browserRuntimeUsesImplicitLegacyHostDefaultFallback(defaultRoute, substrate) {
		return BrowserRuntimeInfo{}
	}
	return defaultRoute
}

func browserVisibleDefaultRouteRuntimeInfoForAssessment(substrate browserDefaultSubstrateAssessment) BrowserRuntimeInfo {
	return browserVisibleDefaultRouteRuntimeInfo(browserDefaultRouteRuntimeInfoForAssessment(substrate), substrate)
}

func browserRuntimeUsesImplicitLegacyHostDefaultFallback(defaultRoute BrowserRuntimeInfo, substrate browserDefaultSubstrateAssessment) bool {
	defaultRoute = normalizeBrowserRuntimeInfo(defaultRoute)
	if BrowserSubstratePosture(defaultRoute.Backend, defaultRoute.Target) != BrowserSubstrateLegacySystemHost {
		return false
	}
	return !browserConcreteRouteAssessmentHasResult(substrate.DefaultConcreteRoute)
}

func browserDefaultPromotionRouteForConcreteBackend(backend BrowserBackend, fallback BrowserRuntimeInfo) (browserResolvedExecutionRoute, bool) {
	assessment := browserDefaultPromotionRouteAssessmentForConcreteBackend(backend, fallback, normalizeBrowserRuntimeInfo(fallback).Target)
	return assessment.Route, assessment.Ready
}

type browserDefaultPromotionRouteAssessment struct {
	Configured     bool
	RouteAvailable bool
	Ready          bool
	Route          browserResolvedExecutionRoute
	FailureReason  string
	FailureNote    string
}

type browserConcreteRouteAssessment struct {
	Configured     bool
	RouteAvailable bool
	Route          browserResolvedExecutionRoute
	FailureReason  string
	FailureNote    string
}

type browserDefaultSubstrateAssessment struct {
	HostRoute            browserConcreteRouteAssessment
	HostRuntime          BrowserRuntimeInfo
	NodeRoute            browserDefaultPromotionRouteAssessment
	SandboxRoute         browserDefaultPromotionRouteAssessment
	SandboxConcreteRoute browserConcreteRouteAssessment
	DefaultRuntime       BrowserRuntimeInfo
	DefaultConcreteRoute browserConcreteRouteAssessment
}

func browserConcreteRouteAssessmentForConcreteBackend(backend BrowserBackend, fallback BrowserRuntimeInfo, routeLabel string) browserConcreteRouteAssessment {
	routeLabel = strings.ToLower(strings.TrimSpace(routeLabel))
	if backend == nil {
		return browserConcreteRouteAssessment{}
	}
	assessment := browserConcreteRouteAssessment{Configured: true}
	route, err := resolveConcreteBrowserExecutionRoute(backend, fallback, BrowserRuntimeInfo{})
	if err != nil {
		assessment.FailureReason = fmt.Sprintf("%s runtime route is configured but could not be resolved: %v", routeLabel, err)
		assessment.FailureNote = err.Error()
		return assessment
	}
	assessment.RouteAvailable = true
	assessment.Route = route
	return assessment
}

func browserDefaultPromotionRouteAssessmentForConcreteBackend(backend BrowserBackend, fallback BrowserRuntimeInfo, routeLabel string) browserDefaultPromotionRouteAssessment {
	concrete := browserConcreteRouteAssessmentForConcreteBackend(backend, fallback, routeLabel)
	return browserDefaultPromotionRouteAssessmentForConcreteRouteAssessment(concrete, routeLabel)
}

func browserDefaultPromotionRouteAssessmentForConcreteRouteAssessment(concrete browserConcreteRouteAssessment, routeLabel string) browserDefaultPromotionRouteAssessment {
	if !concrete.Configured {
		return browserDefaultPromotionRouteAssessment{}
	}
	assessment := browserDefaultPromotionRouteAssessment{
		Configured:     concrete.Configured,
		RouteAvailable: concrete.RouteAvailable,
		Route:          concrete.Route,
		FailureNote:    concrete.FailureNote,
	}
	if !concrete.RouteAvailable {
		assessment.FailureReason = strings.Replace(concrete.FailureReason, "is configured but could not be resolved", "is configured but not the default because its concrete default route could not be resolved", 1)
		return assessment
	}
	if !BrowserDefaultRuntimePromotionReady(concrete.Route.Capabilities) {
		assessment.FailureReason = fmt.Sprintf("%s runtime route is configured but not the default because it does not yet advertise the required default browser capabilities; it remains available via `runtime_target=%s`", routeLabel, routeLabel)
		return assessment
	}
	assessment.Ready = true
	return assessment
}

func browserPromotedDefaultRuntimeInfoForAssessments(hostInfo BrowserRuntimeInfo, nodeAssessment browserDefaultPromotionRouteAssessment, sandboxAssessment browserDefaultPromotionRouteAssessment) BrowserRuntimeInfo {
	nodeInfo := defaultBrowserNodeRuntimeInfo()
	if nodeAssessment.Ready {
		nodeInfo = nodeAssessment.Route.RuntimeInfo
	}
	sandboxInfo := defaultBrowserSandboxRuntimeInfo()
	if sandboxAssessment.Ready {
		sandboxInfo = sandboxAssessment.Route.RuntimeInfo
	}
	return BrowserPromoteDefaultRuntimeInfo(hostInfo, nodeInfo, nodeAssessment.Ready, sandboxInfo, sandboxAssessment.Ready)
}

func browserDefaultSubstrateAssessmentForOptions(opts BrowserToolOptions) browserDefaultSubstrateAssessment {
	return browserDefaultRuntimePreviewForToolOptions(opts).SubstrateAssessment
}

func browserDefaultSubstrateAssessmentForHostBackend(opts BrowserToolOptions, hostBackend BrowserBackend) browserDefaultSubstrateAssessment {
	hostInfo, hostAssessment := browserConcreteHostRuntimeInfoForHostBackend(opts, hostBackend)
	nodeAssessment := browserDefaultPromotionRouteAssessmentForConcreteBackend(opts.NodeBackend, defaultBrowserNodeRuntimeInfo(), "node")
	sandboxConcreteAssessment := browserConcreteRouteAssessmentForConcreteBackend(opts.SandboxBackend, defaultBrowserSandboxRuntimeInfo(), "sandbox")
	sandboxAssessment := browserDefaultPromotionRouteAssessmentForConcreteRouteAssessment(sandboxConcreteAssessment, "sandbox")
	defaultRuntime := browserPromotedDefaultRuntimeInfoForAssessments(hostInfo, nodeAssessment, sandboxAssessment)
	assessment := browserDefaultSubstrateAssessment{
		HostRoute:            hostAssessment,
		HostRuntime:          hostInfo,
		NodeRoute:            nodeAssessment,
		SandboxRoute:         sandboxAssessment,
		SandboxConcreteRoute: sandboxConcreteAssessment,
		DefaultRuntime:       defaultRuntime,
	}
	assessment.DefaultConcreteRoute = browserSharedDefaultConcreteRouteAssessment(opts, defaultRuntime, assessment)
	return assessment
}

func browserSharedDefaultConcreteRouteAssessment(opts BrowserToolOptions, defaultRoute BrowserRuntimeInfo, substrate browserDefaultSubstrateAssessment) browserConcreteRouteAssessment {
	defaultRoute = normalizeBrowserRuntimeInfo(defaultRoute)
	if browserImplicitLegacyHostRouteAssessmentEnabled(opts, substrate.HostRuntime) && defaultRoute.Target == "host" {
		return browserConcreteRouteAssessment{}
	}
	return browserDefaultConcreteRouteAssessment(defaultRoute, substrate)
}

func browserConcreteHostRuntimeInfoForOptions(opts BrowserToolOptions) (BrowserRuntimeInfo, browserConcreteRouteAssessment) {
	return browserConcreteHostRuntimeInfoForHostBackend(opts, browserRegistrationHostBackend(opts))
}

func browserConcreteHostRuntimeInfoForHostBackend(opts BrowserToolOptions, hostBackend BrowserBackend) (BrowserRuntimeInfo, browserConcreteRouteAssessment) {
	hostInfo := browserRuntimeHostRuntimeInfoForOptions(opts)
	if browserImplicitLegacyHostRouteAssessmentEnabled(opts, hostInfo) {
		return hostInfo, browserImplicitLegacyHostRouteAssessment(hostBackend, hostInfo)
	}
	assessment := browserConcreteRouteAssessmentForConcreteBackend(hostBackend, hostInfo, "host")
	if assessment.RouteAvailable {
		hostInfo = assessment.Route.RuntimeInfo
	}
	return hostInfo, assessment
}

func browserImplicitLegacyHostRouteAssessmentEnabled(opts BrowserToolOptions, hostInfo BrowserRuntimeInfo) bool {
	return opts.Backend == nil &&
		BrowserSubstratePosture(hostInfo.Backend, hostInfo.Target) == BrowserSubstrateLegacySystemHost
}

func browserImplicitLegacyHostRouteAssessment(hostBackend BrowserBackend, hostInfo BrowserRuntimeInfo) browserConcreteRouteAssessment {
	capabilities := browserCapabilitiesForConcreteBackend(hostBackend)
	if capabilities == (BrowserCapabilities{}) {
		capabilities = defaultBrowserCapabilities()
	}
	return browserConcreteRouteAssessment{
		Configured:     true,
		RouteAvailable: true,
		Route: browserResolvedExecutionRoute{
			Backend:      hostBackend,
			RuntimeInfo:  hostInfo,
			Capabilities: capabilities,
		},
	}
}

func browserDefaultConcreteRouteAssessment(defaultRoute BrowserRuntimeInfo, substrate browserDefaultSubstrateAssessment) browserConcreteRouteAssessment {
	switch normalizeBrowserRuntimeInfo(defaultRoute).Target {
	case "node":
		return browserConcreteRouteAssessmentForDefaultPromotion(substrate.NodeRoute)
	case "sandbox":
		return substrate.SandboxConcreteRoute
	default:
		return substrate.HostRoute
	}
}

func browserDefaultRouteRuntimeInfoForAssessment(assessment browserDefaultSubstrateAssessment) BrowserRuntimeInfo {
	if assessment.DefaultConcreteRoute.RouteAvailable {
		return normalizeBrowserRuntimeInfo(assessment.DefaultConcreteRoute.Route.RuntimeInfo)
	}
	return normalizeBrowserRuntimeInfo(assessment.DefaultRuntime)
}

func browserRuntimeHostRuntimeInfoForOptions(opts BrowserToolOptions) BrowserRuntimeInfo {
	info := browserRuntimeInfoForBackend(opts, browserConfiguredHostBackendForOptions(opts))
	if opts.Backend != nil && strings.TrimSpace(info.Backend) == "" {
		info.Backend = "custom"
	}
	hostInfo := mergeBrowserRuntimeInfo(defaultBrowserRuntimeInfo(), info)
	hostInfo.Target = "host"
	return hostInfo
}

func browserRuntimeInfoForBackend(opts BrowserToolOptions, backend BrowserBackend) BrowserRuntimeInfo {
	if provider, ok := backend.(BrowserRuntimeInfoProvider); ok {
		return normalizeBrowserRuntimeInfo(provider.BrowserRuntimeInfo())
	}
	if opts.Backend != nil {
		return normalizeBrowserRuntimeInfo(BrowserRuntimeInfo{
			Backend: "custom",
		})
	}
	return defaultBrowserRuntimeInfo()
}

func browserRuntimeInfoForConcreteBackend(backend BrowserBackend, fallback BrowserRuntimeInfo) BrowserRuntimeInfo {
	if provider, ok := backend.(BrowserRuntimeInfoProvider); ok {
		info := normalizeBrowserRuntimeInfo(provider.BrowserRuntimeInfo())
		if info.Backend == "" {
			info.Backend = fallback.Backend
		}
		if info.Profile == "" {
			info.Profile = fallback.Profile
		}
		if info.Target == "" {
			info.Target = fallback.Target
		}
		return info
	}
	return normalizeBrowserRuntimeInfo(fallback)
}

func normalizeBrowserRuntimeInfo(info BrowserRuntimeInfo) BrowserRuntimeInfo {
	return BrowserRuntimeInfo{
		Backend: strings.ToLower(strings.TrimSpace(info.Backend)),
		Profile: strings.ToLower(strings.TrimSpace(info.Profile)),
		Target:  strings.ToLower(strings.TrimSpace(info.Target)),
	}
}
