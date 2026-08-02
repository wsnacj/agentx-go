package tools

import "context"

func (b browserRuntimeRouterBackend) Open(ctx context.Context, req BrowserOpenRequest) (BrowserOpenResult, error) {
	return invokeDirectURLAction(ctx, b, "browser backend open", req.BrowserApp, 0, req.URL, func(backend BrowserBackend, browserApp string) (BrowserOpenResult, error) {
		req.BrowserApp = browserApp
		return backend.Open(ctx, req)
	})
}

func (b browserRuntimeRouterBackend) Navigate(ctx context.Context, req BrowserNavigateRequest) (BrowserNavigateResult, error) {
	return invokeDirectURLAction(ctx, b, "browser backend navigate", req.BrowserApp, req.TabIndex, req.URL, func(backend BrowserBackend, browserApp string) (BrowserNavigateResult, error) {
		req.BrowserApp = browserApp
		return backend.Navigate(ctx, req)
	})
}

func (b browserRuntimeRouterBackend) Tabs(ctx context.Context, req BrowserTabsRequest) (BrowserTabsResult, error) {
	return invokeDirectTabsAction(ctx, b, "browser backend tabs", req, func(backend BrowserBackend, req BrowserTabsRequest) (BrowserTabsResult, error) {
		return backend.Tabs(ctx, req)
	})
}

func (b browserRuntimeRouterBackend) Extract(ctx context.Context, req BrowserExtractRequest) (BrowserExtractResult, error) {
	return invokeDirectPageAction(ctx, b, "browser backend extract", req.BrowserApp, req.TabIndex, req.URL, func(backend BrowserBackend, browserApp string) (BrowserExtractResult, error) {
		req.BrowserApp = browserApp
		return backend.Extract(ctx, req)
	})
}

func (b browserRuntimeRouterBackend) Snapshot(ctx context.Context, req BrowserSnapshotRequest) (BrowserSnapshotResult, error) {
	return invokeDirectPageAction(ctx, b, "browser backend snapshot", req.BrowserApp, req.TabIndex, req.URL, func(backend BrowserBackend, browserApp string) (BrowserSnapshotResult, error) {
		req.BrowserApp = browserApp
		return backend.Snapshot(ctx, req)
	})
}

func (b browserRuntimeRouterBackend) Screenshot(ctx context.Context, req BrowserScreenshotRequest) (BrowserScreenshotResult, error) {
	return invokeDirectPageAction(ctx, b, "browser backend screenshot", req.BrowserApp, req.TabIndex, req.URL, func(backend BrowserBackend, browserApp string) (BrowserScreenshotResult, error) {
		req.BrowserApp = browserApp
		return backend.Screenshot(ctx, req)
	})
}

func (b browserRuntimeRouterBackend) Click(ctx context.Context, req BrowserClickRequest) (BrowserClickResult, error) {
	return invokeDirectPageAction(ctx, b, "browser backend click", req.BrowserApp, req.TabIndex, req.URL, func(backend BrowserBackend, browserApp string) (BrowserClickResult, error) {
		req.BrowserApp = browserApp
		return backend.Click(ctx, req)
	})
}

func (b browserRuntimeRouterBackend) Type(ctx context.Context, req BrowserTypeRequest) (BrowserTypeResult, error) {
	return invokeDirectPageAction(ctx, b, "browser backend type", req.BrowserApp, req.TabIndex, req.URL, func(backend BrowserBackend, browserApp string) (BrowserTypeResult, error) {
		req.BrowserApp = browserApp
		return backend.Type(ctx, req)
	})
}

func (b browserRuntimeRouterBackend) Eval(ctx context.Context, req BrowserEvalRequest) (BrowserEvalResult, error) {
	return invokeDirectPageAction(ctx, b, "browser backend eval", req.BrowserApp, req.TabIndex, req.URL, func(backend BrowserBackend, browserApp string) (BrowserEvalResult, error) {
		req.BrowserApp = browserApp
		return backend.Eval(ctx, req)
	})
}
