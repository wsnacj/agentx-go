package tools

type browserDarwinPlatformLane struct{}

func (browserDarwinPlatformLane) Name() string {
	return "darwin"
}

func (browserDarwinPlatformLane) DefaultRuntime() BrowserRuntimeInfo {
	return defaultBrowserRuntimeInfo()
}

func (browserDarwinPlatformLane) DefaultCapabilities() BrowserCapabilities {
	return BrowserCapabilities{}
}

func (browserDarwinPlatformLane) DefaultBackend(opts BrowserToolOptions, policy outboundNetworkPolicy, timeoutMs int) BrowserBackend {
	return browserHostBackendForOptions(opts, policy, timeoutMs)
}

type browserLinuxPlatformLane struct{}

func (browserLinuxPlatformLane) Name() string {
	return "linux"
}

func (browserLinuxPlatformLane) DefaultRuntime() BrowserRuntimeInfo {
	return defaultBrowserRuntimeInfo()
}

func (browserLinuxPlatformLane) DefaultCapabilities() BrowserCapabilities {
	return BrowserCapabilities{}
}

func (browserLinuxPlatformLane) DefaultBackend(opts BrowserToolOptions, policy outboundNetworkPolicy, timeoutMs int) BrowserBackend {
	return browserHostBackendForOptions(opts, policy, timeoutMs)
}

type browserDefaultPlatformLane struct{}

func (browserDefaultPlatformLane) Name() string {
	return "default"
}

func (browserDefaultPlatformLane) DefaultRuntime() BrowserRuntimeInfo {
	return defaultBrowserRuntimeInfo()
}

func (browserDefaultPlatformLane) DefaultCapabilities() BrowserCapabilities {
	return BrowserCapabilities{}
}

func (browserDefaultPlatformLane) DefaultBackend(opts BrowserToolOptions, policy outboundNetworkPolicy, timeoutMs int) BrowserBackend {
	return browserHostBackendForOptions(opts, policy, timeoutMs)
}
