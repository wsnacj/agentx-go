package tools

import (
	"context"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

type browserRuntimeSessionSelectionExecutionPreview struct {
	DefaultCandidateRoute         BrowserRuntimeInfo
	DefaultCandidateDescriptor    browserRuntimeRouteDescriptor
	Base                          BrowserRuntimeInfo
	HiddenImplicitHostDefaultBase bool
	SessionSelectionPreview       browserRuntimeSessionSelectionPreview
}

func browserRuntimeSessionExecutionPreviewForBackend(
	ctx context.Context,
	stateRegistry agentxbrowserruntime.SharedSessionBrowserStateRegistry,
	sessionRegistry *BrowserSessionRegistry,
	params map[string]any,
	base BrowserRuntimeInfo,
	hiddenImplicitHostDefaultBase bool,
	backend BrowserBackend,
) browserRuntimeSessionSelectionExecutionPreview {
	switch routed := backend.(type) {
	case browserRuntimeRouterBackend:
		sessionExecutionPreview := routed.sessionExecutionPreview(base, hiddenImplicitHostDefaultBase)
		return browserRuntimeSessionSelectionExecutionPreview{
			DefaultCandidateRoute:         normalizeBrowserRuntimeInfo(sessionExecutionPreview.DefaultCandidateRoute),
			DefaultCandidateDescriptor:    sessionExecutionPreview.DefaultCandidateDescriptor,
			Base:                          sessionExecutionPreview.Base,
			HiddenImplicitHostDefaultBase: sessionExecutionPreview.HiddenImplicitHostDefaultBase,
			SessionSelectionPreview: browserRuntimePreviewSessionSelections(
				ctx,
				stateRegistry,
				sessionRegistry,
				params,
				sessionExecutionPreview.Base,
				sessionExecutionPreview.HiddenImplicitHostDefaultBase,
			),
		}
	case *browserRuntimeRouterBackend:
		if routed != nil {
			sessionExecutionPreview := routed.sessionExecutionPreview(base, hiddenImplicitHostDefaultBase)
			return browserRuntimeSessionSelectionExecutionPreview{
				DefaultCandidateRoute:         normalizeBrowserRuntimeInfo(sessionExecutionPreview.DefaultCandidateRoute),
				DefaultCandidateDescriptor:    sessionExecutionPreview.DefaultCandidateDescriptor,
				Base:                          sessionExecutionPreview.Base,
				HiddenImplicitHostDefaultBase: sessionExecutionPreview.HiddenImplicitHostDefaultBase,
				SessionSelectionPreview: browserRuntimePreviewSessionSelections(
					ctx,
					stateRegistry,
					sessionRegistry,
					params,
					sessionExecutionPreview.Base,
					sessionExecutionPreview.HiddenImplicitHostDefaultBase,
				),
			}
		}
	}
	base, hiddenImplicitHostDefaultBase = browserResolveDefaultRequestPreview(
		params,
		base,
		hiddenImplicitHostDefaultBase,
		backend,
	)
	return browserRuntimeSessionSelectionExecutionPreview{
		Base:                          base,
		HiddenImplicitHostDefaultBase: hiddenImplicitHostDefaultBase,
		SessionSelectionPreview: browserRuntimePreviewSessionSelections(
			ctx,
			stateRegistry,
			sessionRegistry,
			params,
			base,
			hiddenImplicitHostDefaultBase,
		),
	}
}

func browserRuntimePreviewSessionSelectionsForExecution(
	ctx context.Context,
	stateRegistry agentxbrowserruntime.SharedSessionBrowserStateRegistry,
	sessionRegistry *BrowserSessionRegistry,
	params map[string]any,
	base BrowserRuntimeInfo,
	hiddenImplicitHostDefaultBase bool,
	backend BrowserBackend,
) browserRuntimeSessionSelectionExecutionPreview {
	return browserRuntimeSessionExecutionPreviewForBackend(
		ctx,
		stateRegistry,
		sessionRegistry,
		params,
		base,
		hiddenImplicitHostDefaultBase,
		backend,
	)
}
