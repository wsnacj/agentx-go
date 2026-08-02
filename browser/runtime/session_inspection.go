package browserruntime

import (
	"context"
	"time"
)

// SharedSessionBrowserInspectionObservation is kept as a compatibility alias
// while runtime inspection/watch entrypoints converge on the shared watch
// owner.
type SharedSessionBrowserInspectionObservation = SharedSessionBrowserWatchObservation

// ObserveSharedSessionBrowserInspectionForScope runs a shared scoped watch
// cycle and projects the binding/session view plus the scoped profile payload
// metadata consumed by runtime inspection surfaces.
func ObserveSharedSessionBrowserInspectionForScope(
	ctx context.Context,
	control BrowserRuntimeControlBackend,
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	bindingRoute BrowserSessionRoute,
	requestedProfile string,
	includeStatus bool,
	includeProfiles bool,
	includeSessionView bool,
	sessionViewInfo BrowserRuntimeInfo,
	sessionViewRouteFilter BrowserSessionRoute,
	sessionViewRequestedProfile string,
	registry *BrowserSessionRegistry,
	runRegistry SharedSessionRunRegistry,
	stateRegistry SharedSessionBrowserStateRegistry,
	reconnectWindow time.Duration,
) SharedSessionBrowserInspectionObservation {
	return sharedSessionBrowserObserverManager(
		registry,
		runRegistry,
		stateRegistry,
		reconnectWindow,
	).ObserveInspection(ctx, control, SharedSessionBrowserObserverRequest{
		SessionID:                   sessionID,
		SelectedInfo:                selectedInfo,
		BindingRoute:                bindingRoute,
		RequestedProfile:            requestedProfile,
		IncludeStatus:               includeStatus,
		IncludeProfiles:             includeProfiles,
		IncludeSessionView:          includeSessionView,
		SessionViewInfo:             sessionViewInfo,
		SessionViewRouteFilter:      sessionViewRouteFilter,
		SessionViewRequestedProfile: sessionViewRequestedProfile,
	})
}
