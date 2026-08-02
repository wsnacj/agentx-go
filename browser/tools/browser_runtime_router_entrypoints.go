package tools

import (
	"context"
)

func (b browserRuntimeRouterBackend) ResolveBrowserExecutionRoute(requested BrowserRuntimeInfo) (browserResolvedExecutionRoute, error) {
	return b.resolveBrowserRoute(requested)
}

func (b browserRuntimeRouterBackend) ResolveBrowserExecutionRouteForSession(ctx context.Context, params map[string]any, base BrowserRuntimeInfo, hiddenImplicitHostDefaultBase bool) (browserResolvedExecutionRoute, error) {
	executionPreview := b.sessionExecutionPreview(base, hiddenImplicitHostDefaultBase)
	request, err := b.buildSessionPreflightRequest(
		ctx,
		params,
		executionPreview.Base,
		executionPreview.DefaultCandidateRoute,
		executionPreview.DefaultCandidateDescriptor,
		executionPreview.HiddenImplicitHostDefaultBase,
	)
	if err != nil {
		return browserResolvedExecutionRoute{}, err
	}
	preflight, err := b.resolveSessionPreflight(request)
	if err != nil {
		return browserResolvedExecutionRoute{}, err
	}
	return preflight.Route, nil
}

func (b browserRuntimeRouterBackend) ResolveBrowserRuntimeRouteForSession(ctx context.Context, params map[string]any, base BrowserRuntimeInfo, hiddenImplicitHostDefaultBase bool) (BrowserRuntimeInfo, error) {
	executionPreview := b.sessionExecutionPreview(base, hiddenImplicitHostDefaultBase)
	request, err := b.buildSessionPreflightRequest(
		ctx,
		params,
		executionPreview.Base,
		executionPreview.DefaultCandidateRoute,
		executionPreview.DefaultCandidateDescriptor,
		executionPreview.HiddenImplicitHostDefaultBase,
	)
	if err != nil {
		return BrowserRuntimeInfo{}, err
	}
	preflight, err := b.resolveSessionPreflight(request)
	if err != nil {
		return BrowserRuntimeInfo{}, err
	}
	return preflight.Route.RuntimeInfo, nil
}

func (b browserRuntimeRouterBackend) ResolveBrowserRuntimeRoute(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
	route, err := b.resolveBrowserRoute(requested)
	return route.RuntimeInfo, err
}

func (b browserRuntimeRouterBackend) ResolveBrowserBackend(requested BrowserRuntimeInfo) (BrowserBackend, BrowserRuntimeInfo, error) {
	route, err := b.resolveBrowserRoute(requested)
	return route.Backend, route.RuntimeInfo, err
}
