package tools

import (
	"context"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

func browserExecutionLaneForActDispatch(opts BrowserToolOptions, dispatch browserActDispatch) BrowserExecutionLane {
	return browserExecutionLaneFromRoute(currentBrowserPlatformLane(opts), dispatch.Route)
}

func resolveBrowserExecutionLaneForActDispatch(
	ctx context.Context,
	backend BrowserBackend,
	sessionRegistry *BrowserSessionRegistry,
	sessionStateRegistry agentxbrowserruntime.SharedSessionBrowserStateRegistry,
	watchManagerProvider agentxbrowserruntime.SharedSessionBrowserObserverManager,
	baseRuntimeInfo BrowserRuntimeInfo,
	hiddenImplicitHostDefaultBase bool,
	opts BrowserToolOptions,
	maxChars int,
	params map[string]any,
	kind string,
) (BrowserExecutionLane, browserActDispatch, error) {
	dispatch, err := resolveBrowserActDispatch(
		ctx,
		backend,
		sessionRegistry,
		sessionStateRegistry,
		watchManagerProvider,
		baseRuntimeInfo,
		hiddenImplicitHostDefaultBase,
		opts,
		maxChars,
		params,
		kind,
	)
	if err != nil {
		return BrowserExecutionLane{}, browserActDispatch{}, err
	}
	return browserExecutionLaneForActDispatch(opts, dispatch), dispatch, nil
}
