package browserruntime

import (
	"context"
	"time"
)

// SharedSessionBrowserWatchObservation captures the unified source-time watch
// cycle consumed by runtime status/workbench/profiles/sessions surfaces.
type SharedSessionBrowserWatchObservation struct {
	View               SharedSessionBrowserViewObservation
	Profiles           []SharedSessionBrowserProjectedProfileState
	DiscoveredProfiles []string
	DefaultProfile     string
	Note               string
	ReferenceTime      time.Time
}

// ObserveSharedSessionBrowserWatchForScope runs the shared scoped watch cycle
// and projects the binding/session view together with scoped profile payload
// metadata from the same source-time observation.
func ObserveSharedSessionBrowserWatchForScope(
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
) SharedSessionBrowserWatchObservation {
	return sharedSessionBrowserObserverManager(
		registry,
		runRegistry,
		stateRegistry,
		reconnectWindow,
	).ObserveWatch(ctx, control, SharedSessionBrowserObserverRequest{
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
