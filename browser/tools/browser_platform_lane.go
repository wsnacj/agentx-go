package tools

import stdruntime "runtime"

type BrowserPlatformLane interface {
	Name() string
	DefaultRuntime() BrowserRuntimeInfo
	DefaultCapabilities() BrowserCapabilities
	DefaultBackend(opts BrowserToolOptions, policy outboundNetworkPolicy, timeoutMs int) BrowserBackend
}

func currentBrowserPlatformLane(opts BrowserToolOptions) BrowserPlatformLane {
	switch stdruntime.GOOS {
	case "darwin":
		return browserDarwinPlatformLane{}
	case "linux":
		return browserLinuxPlatformLane{}
	default:
		return browserDefaultPlatformLane{}
	}
}
