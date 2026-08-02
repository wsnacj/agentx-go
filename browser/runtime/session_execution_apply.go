package browserruntime

import "time"

// ApplySharedSessionBrowserExecutionResultWithMutationContext applies a
// lifecycle execution result through the shared mutation seam while letting the
// browserruntime owner decide whether to use the event-backed provider path or
// the raw registry fallback.
func ApplySharedSessionBrowserExecutionResultWithMutationContext(
	ctx SharedSessionBrowserMutationContext,
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	requestedProfile string,
	result SharedSessionBrowserExecutionResult,
) SharedSessionBrowserExecutionApplication {
	return sharedSessionBrowserObserverManager(
		ctx.Registry,
		ctx.RunRegistry,
		ctx.StateRegistry,
		ctx.ReconnectWindow,
	).ApplyExecutionResult(
		sessionID,
		selectedInfo,
		requestedProfile,
		result,
	)
}

// ApplySharedSessionBrowserExecutionResultWithManager applies a lifecycle
// execution result through the provided observer manager, defaulting missing
// session/state dependencies to the supplied registries so tools callers do
// not need to reassemble browserruntime-owned provider wiring.
func ApplySharedSessionBrowserExecutionResultWithManager(
	manager SharedSessionBrowserObserverManager,
	sessionRegistry *BrowserSessionRegistry,
	stateRegistry SharedSessionBrowserStateRegistry,
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	requestedProfile string,
	result SharedSessionBrowserExecutionResult,
	reconnectWindow time.Duration,
) SharedSessionBrowserExecutionApplication {
	return ApplySharedSessionBrowserExecutionResultWithMutationContext(
		SharedSessionBrowserMutationContextFor(
			manager,
			sessionRegistry,
			stateRegistry,
			reconnectWindow,
		),
		sessionID,
		selectedInfo,
		requestedProfile,
		result,
	)
}
