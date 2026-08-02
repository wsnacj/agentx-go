package browserruntime

import (
	"context"
	"strings"
	"time"
)

// SharedSessionBrowserObserverRequest captures the scoped watch-cycle inputs
// needed to observe raw source state, sync lifecycle events, and project the
// derived binding/session/profile views from the same source-time cycle.
type SharedSessionBrowserObserverRequest struct {
	SessionID                   string
	SelectedInfo                BrowserRuntimeInfo
	BindingRoute                BrowserSessionRoute
	RequestedProfile            string
	IncludeStatus               bool
	IncludeProfiles             bool
	IncludeSessionView          bool
	SessionViewInfo             BrowserRuntimeInfo
	SessionViewRouteFilter      BrowserSessionRoute
	SessionViewRequestedProfile string
}

// SharedSessionBrowserObserverObservation captures the unified source-time
// observer cycle consumed by binding/view/watch surfaces.
type SharedSessionBrowserObserverObservation struct {
	Observation        SharedSessionBrowserStatusAndProfilesObservation
	Binding            SharedSessionBrowserBindingEvaluation
	Session            SharedSessionBrowserSessionViewSnapshot
	Profiles           []SharedSessionBrowserProjectedProfileState
	DiscoveredProfiles []string
	DefaultProfile     string
	Note               string
	ReferenceTime      time.Time
}

// ObserveSharedSessionBrowserObserverForScope runs the shared source-time
// observer cycle for a scoped route/profile selection and projects binding,
// session-view, and profile payloads from the same observed source of truth.
func ObserveSharedSessionBrowserObserverForScope(
	ctx context.Context,
	control BrowserRuntimeControlBackend,
	req SharedSessionBrowserObserverRequest,
	registry *BrowserSessionRegistry,
	runRegistry SharedSessionRunRegistry,
	stateRegistry SharedSessionBrowserStateRegistry,
	reconnectWindow time.Duration,
) SharedSessionBrowserObserverObservation {
	return sharedSessionBrowserObserverManager(
		registry,
		runRegistry,
		stateRegistry,
		reconnectWindow,
	).ObserveObserver(ctx, control, req)
}

func observeSharedSessionBrowserObserverForScopeFromCycle(
	req SharedSessionBrowserObserverRequest,
	cycle SharedSessionBrowserEventCycleObservation,
	registry *BrowserSessionRegistry,
	runRegistry SharedSessionRunRegistry,
	stateRegistry SharedSessionBrowserStateRegistry,
	reconnectWindow time.Duration,
) SharedSessionBrowserObserverObservation {
	observation, _ := observeSharedSessionBrowserObserverForScopeFromCycleWithInvalidation(
		req,
		cycle,
		registry,
		runRegistry,
		stateRegistry,
		reconnectWindow,
	)
	return observation
}

func observeSharedSessionBrowserObserverForScopeFromCycleWithInvalidation(
	req SharedSessionBrowserObserverRequest,
	cycle SharedSessionBrowserEventCycleObservation,
	registry *BrowserSessionRegistry,
	runRegistry SharedSessionRunRegistry,
	stateRegistry SharedSessionBrowserStateRegistry,
	reconnectWindow time.Duration,
) (SharedSessionBrowserObserverObservation, bool) {
	req.SessionID = strings.TrimSpace(req.SessionID)
	req.SelectedInfo.Backend = strings.TrimSpace(req.SelectedInfo.Backend)
	req.SelectedInfo.Profile = strings.TrimSpace(req.SelectedInfo.Profile)
	req.SelectedInfo.Target = strings.TrimSpace(req.SelectedInfo.Target)
	req.BindingRoute = normalizeBrowserSessionRoute(req.BindingRoute)
	req.RequestedProfile = firstNonEmptyBindingString(
		strings.TrimSpace(req.RequestedProfile),
		strings.TrimSpace(req.SelectedInfo.Profile),
		strings.TrimSpace(req.BindingRoute.Profile),
	)
	req.SessionViewRequestedProfile = strings.TrimSpace(req.SessionViewRequestedProfile)
	req.SessionViewRouteFilter = normalizeBrowserSessionRoute(req.SessionViewRouteFilter)
	req.SessionViewInfo.Backend = strings.TrimSpace(req.SessionViewInfo.Backend)
	req.SessionViewInfo.Profile = strings.TrimSpace(req.SessionViewInfo.Profile)
	req.SessionViewInfo.Target = strings.TrimSpace(req.SessionViewInfo.Target)

	bindingObservation, invalidated := observeSharedSessionBrowserBindingForScopeFromCycleWithInvalidation(
		req,
		cycle,
		registry,
		runRegistry,
		stateRegistry,
		reconnectWindow,
	)
	observation := bindingObservation.Observation
	binding := bindingObservation.Evaluation

	watch := SharedSessionBrowserObserverObservation{
		Observation: observation,
		Binding:     binding,
	}
	watch.ReferenceTime = binding.ReferenceTime
	if watch.ReferenceTime.IsZero() {
		watch.ReferenceTime = cycle.ReferenceTime
	}
	if req.IncludeSessionView && req.SessionID != "" {
		if sharedSessionBrowserSessionViewMatchesBindingScope(req) {
			watch.Session = sharedSessionBrowserSessionViewSnapshotFromBinding(binding)
		} else {
			watch.Session = SnapshotSharedSessionBrowserSessionView(
				req.SessionID,
				req.SessionViewInfo,
				req.SessionViewRequestedProfile,
				req.SessionViewRouteFilter,
				registry,
				runRegistry,
				stateRegistry,
			)
		}
	}
	if observation.Profiles == nil {
		return watch, invalidated
	}

	watch.DiscoveredProfiles = SharedSessionBrowserDiscoveredProfiles(observation.Profiles.Profiles)
	watch.DefaultProfile = strings.TrimSpace(observation.Profiles.DefaultProfile)
	watch.Note = strings.TrimSpace(observation.Profiles.Note)
	selection := sharedSessionBrowserSelectedProfileForTarget(stateRegistry, req.SessionID, req.SelectedInfo.Target)
	if len(observation.Snapshot) > 0 {
		watch.Profiles = ProjectSharedSessionBrowserProfileSnapshot(observation.Snapshot, selection)
		return watch, invalidated
	}
	watch.Profiles = ProjectSharedSessionBrowserObservedProfiles(
		observation.Profiles.Backend,
		req.SelectedInfo.Target,
		observation.Profiles.Profiles,
		selection,
	)
	return watch, invalidated
}

func sharedSessionBrowserSessionViewMatchesBindingScope(req SharedSessionBrowserObserverRequest) bool {
	bindingRoute := normalizeBrowserSessionRoute(req.BindingRoute)
	sessionViewRoute := normalizeBrowserSessionRoute(req.SessionViewRouteFilter)
	if bindingRoute != sessionViewRoute {
		return false
	}
	selectedInfo := req.SelectedInfo
	selectedInfo.Backend = strings.TrimSpace(selectedInfo.Backend)
	selectedInfo.Profile = strings.TrimSpace(selectedInfo.Profile)
	selectedInfo.Target = strings.TrimSpace(selectedInfo.Target)
	sessionViewInfo := req.SessionViewInfo
	sessionViewInfo.Backend = strings.TrimSpace(sessionViewInfo.Backend)
	sessionViewInfo.Profile = strings.TrimSpace(sessionViewInfo.Profile)
	sessionViewInfo.Target = strings.TrimSpace(sessionViewInfo.Target)
	if selectedInfo != sessionViewInfo {
		return false
	}
	requestedProfile := firstNonEmptyBindingString(
		strings.TrimSpace(req.RequestedProfile),
		strings.TrimSpace(selectedInfo.Profile),
		strings.TrimSpace(bindingRoute.Profile),
	)
	sessionViewRequestedProfile := firstNonEmptyBindingString(
		strings.TrimSpace(req.SessionViewRequestedProfile),
		strings.TrimSpace(sessionViewInfo.Profile),
		strings.TrimSpace(sessionViewRoute.Profile),
	)
	return requestedProfile == sessionViewRequestedProfile
}

func sharedSessionBrowserSessionViewSnapshotFromBinding(binding SharedSessionBrowserBindingEvaluation) SharedSessionBrowserSessionViewSnapshot {
	session := SharedSessionBrowserSessionViewSnapshot{
		TargetCount: SharedSessionBrowserRouteTargetCount(binding.Routes),
	}
	if len(binding.Routes) > 0 {
		session.Routes = append([]SharedSessionBrowserRouteSnapshot(nil), binding.Routes...)
	} else if route := sharedSessionBrowserMinimalRouteSnapshotFromBindingSnapshot(binding.Snapshot); route != nil {
		session.Routes = []SharedSessionBrowserRouteSnapshot{*route}
		if session.TargetCount <= 0 {
			session.TargetCount = 1
		}
	}
	if len(binding.Snapshot.Runs) > 0 {
		session.Runs = append([]SharedSessionRunInfo(nil), binding.Snapshot.Runs...)
	}
	session.Profiles = ProjectSharedSessionBrowserProfileSnapshot(
		binding.Snapshot.Profiles,
		binding.Snapshot.SelectedProfileSelection,
	)
	if len(session.Profiles) == 0 {
		session.Profiles = sharedSessionBrowserFallbackProjectedProfilesFromBindingEvaluation(
			binding,
			session.Routes,
		)
	}
	session.Handoff = sharedSessionBrowserBindingEvaluationHandoff(
		binding,
		session.Routes,
		session.Runs,
		session.Profiles,
		session.TargetCount,
	)
	return session
}
