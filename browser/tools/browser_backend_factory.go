package tools

func newBrowserBackend(opts BrowserToolOptions, policy outboundNetworkPolicy, timeoutMs int) BrowserBackend {
	return browserDefaultRuntimePreviewForDispatchOptions(opts, policy, timeoutMs).EffectiveBackend
}
