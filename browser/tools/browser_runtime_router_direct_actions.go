package tools

import (
	"context"
	"strings"
)

func browserRuntimeRouterDirectParams(browserApp string, tabIndex int) map[string]any {
	return map[string]any{
		"browser_app": browserApp,
		"tab_index":   tabIndex,
	}
}

func (b browserRuntimeRouterBackend) directPageActionPreflightRequest(
	ctx context.Context,
	browserApp string,
	tabIndex int,
	requestURL string,
) (browserRuntimeRouterDirectPreflightRequest, error) {
	executionPreview := b.executionPreview()
	target := browserToolTarget{TabIndex: tabIndex}
	base := browserPageActionDispatchBaseForManagedDefaultRoute(
		executionPreview.DefaultRoute,
		executionPreview.HiddenImplicitHostDefaultBase,
		target,
		requestURL,
		"",
	)
	return b.buildDirectPreflightRequest(ctx, browserRuntimeRouterDirectPreflightArgs{
		Params:     browserRuntimeRouterDirectParams(browserApp, tabIndex),
		Base:       base,
		Target:     target,
		RequestURL: requestURL,
	})
}

func (b browserRuntimeRouterBackend) directURLActionPreflightRequest(
	ctx context.Context,
	browserApp string,
	tabIndex int,
	requestURL string,
) (browserRuntimeRouterDirectPreflightRequest, error) {
	target := browserToolTarget{TabIndex: tabIndex}
	base := b.directRequestBaseForDefaultManagedRoute(target, requestURL, "")
	return b.buildDirectPreflightRequest(ctx, browserRuntimeRouterDirectPreflightArgs{
		Params:     browserRuntimeRouterDirectParams(browserApp, tabIndex),
		Base:       base,
		Target:     target,
		RequestURL: requestURL,
	})
}

func (b browserRuntimeRouterBackend) directTabsActionPreflightRequest(
	ctx context.Context,
	req BrowserTabsRequest,
) (browserRuntimeRouterDirectPreflightRequest, error) {
	target := browserToolTarget{TabIndex: req.TabIndex}
	base := b.directTabsRequestBaseForManagedDefaultRoute(req)
	return b.buildDirectPreflightRequest(ctx, browserRuntimeRouterDirectPreflightArgs{
		Params: browserRuntimeRouterDirectParams(req.BrowserApp, req.TabIndex),
		Base:   base,
		Target: target,
	})
}

func (b browserRuntimeRouterBackend) directRequestRuntimeInfo(ctx context.Context, params map[string]any, base BrowserRuntimeInfo, hiddenImplicitHostDefaultBase bool) (BrowserRuntimeInfo, browserRuntimeSessionSelectionPreview, bool, error) {
	preview := browserRuntimePreviewSessionSelections(
		ctx,
		b.stateRegistry,
		b.sessionRegistry,
		params,
		base,
		hiddenImplicitHostDefaultBase,
	)
	requestPreview, err := browserResolveRuntimeRequestFromSessionPreview(
		params,
		preview,
		base,
		hiddenImplicitHostDefaultBase,
		b,
	)
	if err != nil {
		return BrowserRuntimeInfo{}, browserRuntimeSessionSelectionPreview{}, false, err
	}
	return requestPreview.Requested, preview, requestPreview.ExplicitRuntimeTarget, nil
}

func (b browserRuntimeRouterBackend) directRequestBrowserApp(current string, preview browserRuntimeSessionSelectionPreview) string {
	if strings.TrimSpace(current) != "" {
		return current
	}
	return preview.RequestedBrowserApp
}

func (b browserRuntimeRouterBackend) directRequestBaseForDefaultManagedRoute(target browserToolTarget, requestURL string, action string) BrowserRuntimeInfo {
	executionPreview := b.executionPreview()
	base := executionPreview.DefaultRoute
	if !executionPreview.HiddenImplicitHostDefaultBase {
		return base
	}
	if strings.TrimSpace(target.TargetID) != "" || target.TabIndex > 0 {
		return base
	}
	if strings.TrimSpace(requestURL) != "" {
		return BrowserRuntimeInfo{}
	}
	if strings.EqualFold(strings.TrimSpace(action), "list") {
		return BrowserRuntimeInfo{}
	}
	return base
}

func (b browserRuntimeRouterBackend) directTabsRequestBaseForManagedDefaultRoute(req BrowserTabsRequest) BrowserRuntimeInfo {
	target := browserToolTarget{TabIndex: req.TabIndex}
	base := b.directRequestBaseForDefaultManagedRoute(target, "", req.Action)
	executionPreview := b.executionPreview()
	if !executionPreview.HiddenImplicitHostDefaultBase {
		return base
	}
	if strings.EqualFold(strings.TrimSpace(req.Action), "list") || req.TabIndex <= 0 {
		return base
	}
	if route, ok := b.resolveImplicitLegacyHostManagedDefaultRoute(BrowserRuntimeInfo{}); ok {
		return route.RuntimeInfo
	}
	return base
}

func (b browserRuntimeRouterBackend) resolveDirectURLActionPreflight(ctx context.Context, actor string, browserApp string, tabIndex int, requestURL string) (browserRuntimeRouterDirectPreflight, error) {
	target := browserToolTarget{TabIndex: tabIndex}
	return b.resolveDirectPreflight(ctx, browserRuntimeRouterDirectPreflightArgs{
		Params:           browserRuntimeRouterDirectParams(browserApp, tabIndex),
		Base:             b.directRequestBaseForDefaultManagedRoute(target, requestURL, ""),
		Target:           target,
		RequestURL:       requestURL,
		LocalURLFallback: true,
		FallbackCheck: func(request browserRuntimeRouterDirectPreflightRequest) error {
			return b.checkDirectURLActionFallback(actor, request, requestURL)
		},
	})
}

func (b browserRuntimeRouterBackend) resolveDirectTabsActionPreflight(ctx context.Context, actor string, req BrowserTabsRequest) (browserRuntimeRouterDirectPreflight, error) {
	target := browserToolTarget{TabIndex: req.TabIndex}
	var fallbackCheck func(browserRuntimeRouterDirectPreflightRequest) error
	if req.Action == "list" || req.TabIndex > 0 {
		fallbackCheck = func(request browserRuntimeRouterDirectPreflightRequest) error {
			return b.checkDirectTabsActionFallback(actor, request, req.Action)
		}
	}
	return b.resolveDirectPreflight(ctx, browserRuntimeRouterDirectPreflightArgs{
		Params:        browserRuntimeRouterDirectParams(req.BrowserApp, req.TabIndex),
		Base:          b.directTabsRequestBaseForManagedDefaultRoute(req),
		Target:        target,
		FallbackCheck: fallbackCheck,
	})
}

func (b browserRuntimeRouterBackend) resolveDirectPageActionPreflight(ctx context.Context, actor string, browserApp string, tabIndex int, requestURL string) (browserRuntimeRouterDirectPreflight, error) {
	executionPreview := b.executionPreview()
	target := browserToolTarget{TabIndex: tabIndex}
	base := browserPageActionDispatchBaseForManagedDefaultRoute(
		executionPreview.DefaultRoute,
		executionPreview.HiddenImplicitHostDefaultBase,
		target,
		requestURL,
		"",
	)
	return b.resolveDirectPreflight(ctx, browserRuntimeRouterDirectPreflightArgs{
		Params:     browserRuntimeRouterDirectParams(browserApp, tabIndex),
		Base:       base,
		Target:     target,
		RequestURL: requestURL,
		FallbackCheck: func(request browserRuntimeRouterDirectPreflightRequest) error {
			return b.checkDirectPageActionFallback(actor, request, requestURL)
		},
	})
}

func (b browserRuntimeRouterBackend) checkDirectURLActionFallback(actor string, request browserRuntimeRouterDirectPreflightRequest, requestURL string) error {
	return browserImplicitLegacyHostURLNavigationFallbackError(
		actor,
		request.HiddenImplicitHostDefaultBase,
		request.ExplicitRuntimeTarget,
		request.Requested,
		request.Target,
		requestURL,
	)
}

func (b browserRuntimeRouterBackend) checkDirectTabsActionFallback(actor string, request browserRuntimeRouterDirectPreflightRequest, action string) error {
	return browserImplicitLegacyHostTabsActionFallbackError(
		actor,
		request.HiddenImplicitHostDefaultBase,
		request.Requested,
		action,
		request.Target,
	)
}

func (b browserRuntimeRouterBackend) checkDirectPageActionFallback(actor string, request browserRuntimeRouterDirectPreflightRequest, requestURL string) error {
	return browserImplicitLegacyHostDirectPageActionFallbackErrorForRuntime(
		actor,
		request.HiddenImplicitHostDefaultBase,
		request.ExplicitRuntimeTarget,
		request.Requested,
		request.Target,
		requestURL,
	)
}

func browserRuntimeRouterPlatformLane() BrowserPlatformLane {
	return currentBrowserPlatformLane(BrowserToolOptions{})
}

func browserRuntimeRouterExecutionLane(route browserResolvedExecutionRoute) BrowserExecutionLane {
	return browserExecutionLaneFromRoute(browserRuntimeRouterPlatformLane(), route)
}

func (b browserRuntimeRouterBackend) resolveBrowserExecutionLane(requested BrowserRuntimeInfo) (BrowserExecutionLane, error) {
	route, err := b.resolveBrowserRoute(requested)
	if err != nil {
		return BrowserExecutionLane{}, err
	}
	return browserRuntimeRouterExecutionLane(route), nil
}

func invokeDirectURLAction[T any](ctx context.Context, b browserRuntimeRouterBackend, actor string, currentBrowserApp string, tabIndex int, requestURL string, invoke func(BrowserBackend, string) (T, error)) (T, error) {
	preflight, err := b.resolveDirectURLActionPreflight(ctx, actor, currentBrowserApp, tabIndex, requestURL)
	if err != nil {
		var zero T
		return zero, err
	}
	result, err := invoke(preflight.Lane.Backend, b.directRequestBrowserApp(currentBrowserApp, preflight.Preview))
	if err != nil {
		var zero T
		return zero, err
	}
	return browserAnnotateURLFallbackResult(result, preflight.Fallback), nil
}

func invokeDirectTabsAction[T any](ctx context.Context, b browserRuntimeRouterBackend, actor string, req BrowserTabsRequest, invoke func(BrowserBackend, BrowserTabsRequest) (T, error)) (T, error) {
	preflight, err := b.resolveDirectTabsActionPreflight(ctx, actor, req)
	if err != nil {
		var zero T
		return zero, err
	}
	req.BrowserApp = b.directRequestBrowserApp(req.BrowserApp, preflight.Preview)
	return invoke(preflight.Lane.Backend, req)
}

func invokeDirectPageAction[T any](ctx context.Context, b browserRuntimeRouterBackend, actor string, currentBrowserApp string, tabIndex int, requestURL string, invoke func(BrowserBackend, string) (T, error)) (T, error) {
	preflight, err := b.resolveDirectPageActionPreflight(ctx, actor, currentBrowserApp, tabIndex, requestURL)
	if err != nil {
		var zero T
		return zero, err
	}
	return invoke(preflight.Lane.Backend, b.directRequestBrowserApp(currentBrowserApp, preflight.Preview))
}
