package browserruntime

import (
	"context"
	"time"
)

// SharedSessionBrowserWatchLoopObservation captures the explicit source-time
// watch loop owner that stitches together the shared event cycle, observer
// projection, and watch payload view from one scoped pass.
type SharedSessionBrowserWatchLoopObservation struct {
	Cycle         SharedSessionBrowserEventCycleObservation
	Observer      SharedSessionBrowserObserverObservation
	Watch         SharedSessionBrowserWatchObservation
	View          SharedSessionBrowserViewObservation
	ReferenceTime time.Time
}

// ObserveSharedSessionBrowserWatchLoopForScope runs the explicit shared watch
// loop owner for a scoped route/profile selection. The loop reuses the same
// source-time event cycle across observer, binding, session-view, and watch
// payload projection so consumers do not need to layer wrappers on top of one
// another.
func ObserveSharedSessionBrowserWatchLoopForScope(
	ctx context.Context,
	control BrowserRuntimeControlBackend,
	req SharedSessionBrowserObserverRequest,
	registry *BrowserSessionRegistry,
	runRegistry SharedSessionRunRegistry,
	stateRegistry SharedSessionBrowserStateRegistry,
	reconnectWindow time.Duration,
) SharedSessionBrowserWatchLoopObservation {
	return sharedSessionBrowserObserverManager(
		registry,
		runRegistry,
		stateRegistry,
		reconnectWindow,
	).ObserveWatchLoop(ctx, control, req)
}
