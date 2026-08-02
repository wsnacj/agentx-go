package tools

type BrowserExecutionLane struct {
	Runtime      BrowserRuntimeInfo
	Backend      BrowserBackend
	Capabilities BrowserCapabilities
	Platform     string
	Substrate    string
}

func browserExecutionLaneFromRoute(platform BrowserPlatformLane, route browserResolvedExecutionRoute) BrowserExecutionLane {
	if route.Capabilities == (BrowserCapabilities{}) {
		route.Capabilities = browserCapabilitiesForConcreteBackend(route.Backend)
	}
	info := normalizeBrowserRuntimeInfo(route.RuntimeInfo)
	return BrowserExecutionLane{
		Runtime:      info,
		Backend:      route.Backend,
		Capabilities: route.Capabilities,
		Platform:     platform.Name(),
		Substrate:    BrowserSubstratePosture(info.Backend, info.Target),
	}
}

func browserExecutionLaneFromRuntime(platform BrowserPlatformLane, backend BrowserBackend, runtimeInfo BrowserRuntimeInfo, capabilities BrowserCapabilities) BrowserExecutionLane {
	info := normalizeBrowserRuntimeInfo(runtimeInfo)
	return BrowserExecutionLane{
		Runtime:      info,
		Backend:      backend,
		Capabilities: capabilities,
		Platform:     platform.Name(),
		Substrate:    BrowserSubstratePosture(info.Backend, info.Target),
	}
}
