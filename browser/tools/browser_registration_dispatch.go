package tools

import (
	"context"
	"strings"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

type browserRegistrationPageActionDispatchOptions struct {
	UseManagedRoute               bool
	UseManagedDefaultDispatchBase bool
}

type browserRegistrationPageActionDispatch struct {
	Route                         browserResolvedExecutionRoute
	Backend                       BrowserBackend
	RuntimeInfo                   BrowserRuntimeInfo
	HiddenImplicitHostDefaultBase bool
	RequestedBrowserApp           string
	ExplicitRuntimeTarget         bool
	Target                        browserToolTarget
	BrowserApp                    string
	RouteFallback                 *browserRuntimeRouterURLFallback
}

type browserRememberTargetApplyOptions struct {
	Route                    BrowserSessionRoute
	TargetID                 string
	TabIndex                 int
	ExistingDecision         string
	ExistingReady            bool
	PreserveExistingOnSelect bool
}

type browserRememberTargetApplyResult struct {
	Selection        *agentxbrowserruntime.BrowserSessionTargetSelection
	ProfileSelection *browserRuntimeSessionProfileSelection
	Decision         string
	Ready            bool
}

func browserRegistrationDefaultRuntimeInfo(ctx browserRegistrationContext) BrowserRuntimeInfo {
	return browserRegistrationExecutionPreviewForContext(ctx).DefaultRoute
}

func browserRegistrationDispatchRuntime(ctx browserRegistrationContext) (BrowserRuntimeInfo, bool) {
	preview := browserRegistrationExecutionPreviewForContext(ctx)
	return preview.DispatchBase, preview.HiddenImplicitHostDefaultBase
}

func resolveBrowserRegistrationPageActionDispatch(ctx browserRegistrationContext, callCtx context.Context, params map[string]any, options browserRegistrationPageActionDispatchOptions) (browserRegistrationPageActionDispatch, error) {
	return resolveBrowserRegistrationPageActionDispatchWithPreview(
		ctx,
		browserRegistrationExecutionPreviewForContext(ctx),
		callCtx,
		params,
		options,
	)
}

func resolveBrowserRegistrationPageActionDispatchWithPreview(ctx browserRegistrationContext, executionPreview browserRegistrationExecutionPreview, callCtx context.Context, params map[string]any, options browserRegistrationPageActionDispatchOptions) (browserRegistrationPageActionDispatch, error) {
	dispatchBase := executionPreview.DispatchBase
	hiddenImplicitHostDefaultBase := executionPreview.HiddenImplicitHostDefaultBase
	if options.UseManagedDefaultDispatchBase {
		dispatchBase = browserPageActionDispatchBaseForManagedDefaultParams(dispatchBase, hiddenImplicitHostDefaultBase, params)
	}
	backend := executionPreview.EffectiveBackend
	var (
		routedBackend BrowserBackend
		runtimeInfo   BrowserRuntimeInfo
		route         browserResolvedExecutionRoute
		err           error
	)
	if options.UseManagedRoute {
		route, err = resolveBrowserManagedRouteForSession(
			callCtx,
			ctx.sessionStateRegistry,
			ctx.sessionRegistry,
			ctx.watchManagerProvider,
			params,
			dispatchBase,
			hiddenImplicitHostDefaultBase,
			backend,
			ctx.maxChars,
		)
		if err != nil {
			return browserRegistrationPageActionDispatch{}, err
		}
		routedBackend = route.Backend
		runtimeInfo = route.RuntimeInfo
		hiddenImplicitHostDefaultBase = route.hiddenImplicitHostDefaultBase
	} else {
		routedBackend, runtimeInfo, err = resolveBrowserExecutionBackendForSession(
			callCtx,
			ctx.sessionStateRegistry,
			ctx.sessionRegistry,
			params,
			dispatchBase,
			hiddenImplicitHostDefaultBase,
			backend,
		)
		if err != nil {
			return browserRegistrationPageActionDispatch{}, err
		}
	}
	requestedBrowserApp := firstNonEmpty(strings.TrimSpace(firstString(params, "browser", "browser_app", "app")), strings.TrimSpace(ctx.opts.DefaultBrowserApp))
	explicitRuntimeTarget := browserHasExplicitRuntimeTarget(params)
	target, err := resolveBrowserToolTarget(callCtx, ctx.sessionRegistry, runtimeInfo, hiddenImplicitHostDefaultBase, requestedBrowserApp, params)
	if err != nil {
		return browserRegistrationPageActionDispatch{}, err
	}
	browserApp := browserEffectiveBrowserApp(callCtx, ctx.sessionRegistry, runtimeInfo, hiddenImplicitHostDefaultBase, requestedBrowserApp, target)
	return browserRegistrationPageActionDispatch{
		Route:                         route,
		Backend:                       routedBackend,
		RuntimeInfo:                   runtimeInfo,
		HiddenImplicitHostDefaultBase: hiddenImplicitHostDefaultBase,
		RequestedBrowserApp:           requestedBrowserApp,
		ExplicitRuntimeTarget:         explicitRuntimeTarget,
		Target:                        target,
		BrowserApp:                    browserApp,
	}, nil
}

func applyBrowserRememberTargetSelection(callCtx context.Context, ctx browserRegistrationContext, options browserRememberTargetApplyOptions) browserRememberTargetApplyResult {
	rememberResult := agentxbrowserruntime.DispatchSharedSessionBrowserRememberTarget(
		agentxbrowserruntime.SharedSessionBrowserRememberTargetDispatchRequest{
			MutationContext: browserSharedMutationContext(ctx.watchManagerProvider, ctx.sessionRegistry),
			SessionID:       ToolSessionIDFromContext(callCtx),
			Route:           options.Route,
			TargetID:        strings.TrimSpace(options.TargetID),
			TabIndex:        options.TabIndex,
			Source:          "remember_target",
		},
	)
	result := browserRememberTargetApplyResult{
		Selection: rememberResult.Selection,
		Decision:  rememberResult.Decision,
		Ready:     rememberResult.Ready,
	}
	if options.PreserveExistingOnSelect &&
		strings.TrimSpace(options.ExistingDecision) != "" &&
		rememberResult.Selection != nil {
		result.Decision = strings.TrimSpace(options.ExistingDecision)
		result.Ready = options.ExistingReady
	}
	result.ProfileSelection = browserRuntimeSelectionPtrValue(rememberResult.ProfileSelection)
	return result
}
