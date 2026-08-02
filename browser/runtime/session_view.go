package browserruntime

import (
	"context"
	"strings"
	"time"
)

// SharedSessionBrowserSessionViewSnapshot is the shared session-scoped
// projection surfaced by runtime sessions/workbench views.
type SharedSessionBrowserSessionViewSnapshot struct {
	Routes      []SharedSessionBrowserRouteSnapshot
	Runs        []SharedSessionRunInfo
	Profiles    []SharedSessionBrowserProjectedProfileState
	TargetCount int
	Handoff     *SharedSessionBrowserSessionHandoffSummary
}

// SharedSessionBrowserViewObservation combines the scoped status/profiles watch
// cycle and binding evaluation with the broader session-scoped route/run/profile
// snapshot used by workbench and sessions payloads.
type SharedSessionBrowserViewObservation struct {
	Observation SharedSessionBrowserStatusAndProfilesObservation
	Binding     SharedSessionBrowserBindingEvaluation
	Session     SharedSessionBrowserSessionViewSnapshot
}

// ObserveSharedSessionBrowserViewForScope refreshes the scoped binding watch
// cycle and, when requested, snapshots the broader session-scoped route/run/
// profile view from the same shared owner.
func ObserveSharedSessionBrowserViewForScope(
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
) SharedSessionBrowserViewObservation {
	return sharedSessionBrowserObserverManager(
		registry,
		runRegistry,
		stateRegistry,
		reconnectWindow,
	).ObserveView(ctx, control, SharedSessionBrowserObserverRequest{
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

func observeSharedSessionBrowserViewForScopeFromBinding(
	req SharedSessionBrowserObserverRequest,
	binding SharedSessionBrowserBindingObservation,
	registry *BrowserSessionRegistry,
	runRegistry SharedSessionRunRegistry,
	stateRegistry SharedSessionBrowserStateRegistry,
) SharedSessionBrowserViewObservation {
	req.SessionID = strings.TrimSpace(req.SessionID)
	req.SessionViewRequestedProfile = strings.TrimSpace(req.SessionViewRequestedProfile)
	req.SessionViewRouteFilter = normalizeBrowserSessionRoute(req.SessionViewRouteFilter)
	req.SessionViewInfo.Backend = strings.TrimSpace(req.SessionViewInfo.Backend)
	req.SessionViewInfo.Profile = strings.TrimSpace(req.SessionViewInfo.Profile)
	req.SessionViewInfo.Target = strings.TrimSpace(req.SessionViewInfo.Target)

	view := SharedSessionBrowserViewObservation{
		Observation: binding.Observation,
		Binding:     binding.Evaluation,
	}
	if !req.IncludeSessionView || req.SessionID == "" {
		return view
	}
	if sharedSessionBrowserSessionViewMatchesBindingScope(req) {
		view.Session = sharedSessionBrowserSessionViewSnapshotFromBinding(binding.Evaluation)
		return view
	}
	view.Session = SnapshotSharedSessionBrowserSessionView(
		req.SessionID,
		req.SessionViewInfo,
		req.SessionViewRequestedProfile,
		req.SessionViewRouteFilter,
		registry,
		runRegistry,
		stateRegistry,
	)
	return view
}

// SnapshotSharedSessionBrowserSessionView returns the shared route/run/profile
// snapshot surfaced by runtime sessions/workbench payloads.
func SnapshotSharedSessionBrowserSessionView(
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	requestedProfile string,
	routeFilter BrowserSessionRoute,
	registry *BrowserSessionRegistry,
	runRegistry SharedSessionRunRegistry,
	stateRegistry SharedSessionBrowserStateRegistry,
) SharedSessionBrowserSessionViewSnapshot {
	sessionID = strings.TrimSpace(sessionID)
	selectedInfo.Backend = strings.TrimSpace(selectedInfo.Backend)
	selectedInfo.Profile = strings.TrimSpace(selectedInfo.Profile)
	selectedInfo.Target = strings.TrimSpace(selectedInfo.Target)
	requestedProfile = strings.TrimSpace(requestedProfile)
	routeFilter = normalizeBrowserSessionRoute(routeFilter)
	if sessionID == "" {
		return SharedSessionBrowserSessionViewSnapshot{}
	}

	snapshot := SharedSessionBrowserSessionViewSnapshot{}
	if registry != nil {
		snapshot.Routes = SnapshotSharedSessionBrowserRoutes(registry, sessionID, routeFilter)
	}
	if runRegistry != nil {
		snapshot.Runs = runRegistry.SnapshotSessionRuns(sessionID)
	}
	if stateRegistry != nil {
		snapshot.Profiles = SnapshotSharedSessionBrowserProjectedProfilesForScope(stateRegistry, sessionID, selectedInfo, requestedProfile)
	}
	snapshot.TargetCount = SharedSessionBrowserRouteTargetCount(snapshot.Routes)
	snapshot.Handoff = BuildSharedSessionBrowserSessionHandoffSummary(SharedSessionBrowserSessionHandoffRequest{
		Routes:      snapshot.Routes,
		Runs:        snapshot.Runs,
		Profiles:    snapshot.Profiles,
		TargetCount: snapshot.TargetCount,
	})
	return snapshot
}

// SharedSessionBrowserRouteTargetCount returns the total number of tracked
// targets across the supplied route snapshots.
func SharedSessionBrowserRouteTargetCount(routes []SharedSessionBrowserRouteSnapshot) int {
	total := 0
	for _, route := range routes {
		total += len(route.Targets)
	}
	return total
}
