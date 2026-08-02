package tools

import (
	"context"
	"strings"
)

func browserExecutionLaneForRegistrationDispatch(ctx browserRegistrationContext, dispatch browserRegistrationPageActionDispatch) BrowserExecutionLane {
	platform := currentBrowserPlatformLane(ctx.opts)
	if dispatch.Route.Backend != nil {
		return browserExecutionLaneFromRoute(platform, dispatch.Route)
	}
	return browserExecutionLaneFromRuntime(
		platform,
		dispatch.Backend,
		dispatch.RuntimeInfo,
		browserCapabilitiesForConcreteBackend(dispatch.Backend),
	)
}

func browserDefaultExecutionLaneForOptions(opts BrowserToolOptions) BrowserExecutionLane {
	platform := currentBrowserPlatformLane(opts)
	preview := browserDefaultRuntimePreviewForToolOptions(opts)
	return browserExecutionLaneFromRuntime(
		platform,
		preview.EffectiveBackend,
		preview.LogicalDefaultRoute,
		preview.RegistrationCapabilities,
	)
}

func resolveBrowserExecutionLaneForRegistrationDispatch(ctx browserRegistrationContext, callCtx context.Context, params map[string]any, options browserRegistrationPageActionDispatchOptions) (BrowserExecutionLane, browserRegistrationPageActionDispatch, error) {
	dispatch, err := resolveBrowserRegistrationPageActionDispatch(ctx, callCtx, params, options)
	if err != nil {
		return BrowserExecutionLane{}, browserRegistrationPageActionDispatch{}, err
	}
	return browserExecutionLaneForRegistrationDispatch(ctx, dispatch), dispatch, nil
}

func resolveBrowserExecutionLaneForRegistrationURLAction(ctx browserRegistrationContext, callCtx context.Context, params map[string]any, options browserRegistrationPageActionDispatchOptions) (BrowserExecutionLane, browserRegistrationPageActionDispatch, error) {
	preview := browserRegistrationExecutionPreviewForContext(ctx)
	preview.DispatchBase = browserURLActionDispatchBaseForManagedDefaultParams(
		preview.DispatchBase,
		preview.HiddenImplicitHostDefaultBase,
		params,
	)
	dispatch, err := resolveBrowserRegistrationPageActionDispatchWithPreview(ctx, preview, callCtx, params, options)
	if err != nil {
		return BrowserExecutionLane{}, browserRegistrationPageActionDispatch{}, err
	}
	return browserExecutionLaneForRegistrationDispatch(ctx, dispatch), dispatch, nil
}

func resolveBrowserExecutionLaneForRegistrationDirectURLAction(ctx browserRegistrationContext, callCtx context.Context, params map[string]any, actor string, requestURL string) (BrowserExecutionLane, browserRegistrationPageActionDispatch, error) {
	executionPreview := browserRegistrationExecutionPreviewForContext(ctx)
	router, ok := executionPreview.EffectiveBackend.(browserRuntimeRouterBackend)
	if !ok {
		return resolveBrowserExecutionLaneForRegistrationURLAction(ctx, callCtx, params, browserRegistrationPageActionDispatchOptions{})
	}
	requestedBrowserApp := firstNonEmpty(
		strings.TrimSpace(firstString(params, "browser", "browser_app", "app")),
		strings.TrimSpace(ctx.opts.DefaultBrowserApp),
	)
	preflight, err := router.resolveDirectURLActionPreflight(
		callCtx,
		strings.TrimSpace(actor),
		requestedBrowserApp,
		firstInt(params, "tab_index"),
		strings.TrimSpace(requestURL),
	)
	if err != nil {
		return BrowserExecutionLane{}, browserRegistrationPageActionDispatch{}, err
	}
	dispatch := browserRegistrationPageActionDispatch{
		Backend:                       preflight.Lane.Backend,
		RuntimeInfo:                   preflight.Lane.Runtime,
		HiddenImplicitHostDefaultBase: preflight.HiddenImplicitHostDefaultBase,
		RequestedBrowserApp:           requestedBrowserApp,
		ExplicitRuntimeTarget:         preflight.ExplicitRuntimeTarget,
		Target:                        preflight.Target,
		RouteFallback:                 preflight.Fallback,
	}
	dispatch.BrowserApp = browserEffectiveBrowserApp(
		callCtx,
		ctx.sessionRegistry,
		dispatch.RuntimeInfo,
		dispatch.HiddenImplicitHostDefaultBase,
		dispatch.RequestedBrowserApp,
		dispatch.Target,
	)
	return preflight.Lane, dispatch, nil
}

func resolveBrowserExecutionLaneForRegistrationPageAction(ctx browserRegistrationContext, callCtx context.Context, params map[string]any, options browserRegistrationPageActionDispatchOptions) (BrowserExecutionLane, browserRegistrationPageActionDispatch, error) {
	return resolveBrowserExecutionLaneForRegistrationDispatch(ctx, callCtx, params, options)
}
