package browserruntime

import (
	"strings"
	"time"
)

// SharedSessionBrowserTopLevelBindingProjection captures the lifecycle-owned
// selection and binding evaluation that runtime top-level payloads surface
// alongside route/session projections.
type SharedSessionBrowserTopLevelBindingProjection struct {
	ProfileSelection *SharedSessionBrowserProfileSelection
	TargetSelection  *BrowserSessionTargetSelection
	Evaluation       SharedSessionBrowserBindingEvaluation
}

// ProjectSharedSessionBrowserTopLevelBindingFromEvaluation refreshes the
// shared selection/binding projection from an existing lifecycle-owned binding
// evaluation even when the caller no longer has a scoped selected route.
func ProjectSharedSessionBrowserTopLevelBindingFromEvaluation(
	sessionID string,
	routes []SharedSessionBrowserRouteSnapshot,
	registry *BrowserSessionRegistry,
	runRegistry SharedSessionRunRegistry,
	stateRegistry SharedSessionBrowserStateRegistry,
	evaluation SharedSessionBrowserBindingEvaluation,
	current *SharedSessionBrowserProfileState,
	reconnectWindow time.Duration,
) SharedSessionBrowserTopLevelBindingProjection {
	selectedInfo := sharedSessionBrowserTopLevelBindingSelectedInfo(evaluation, current)
	return ProjectSharedSessionBrowserTopLevelBinding(
		sessionID,
		selectedInfo,
		BrowserSessionRoute{
			Backend: selectedInfo.Backend,
			Profile: selectedInfo.Profile,
			Target:  selectedInfo.Target,
		},
		routes,
		registry,
		runRegistry,
		stateRegistry,
		&evaluation,
		current,
		reconnectWindow,
	)
}

// ProjectSharedSessionBrowserTopLevelBinding refreshes the shared
// selection/binding projection for a scoped route and optionally merges a
// current profile observation into that lifecycle-owned snapshot.
func ProjectSharedSessionBrowserTopLevelBinding(
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	route BrowserSessionRoute,
	routes []SharedSessionBrowserRouteSnapshot,
	registry *BrowserSessionRegistry,
	runRegistry SharedSessionRunRegistry,
	stateRegistry SharedSessionBrowserStateRegistry,
	evaluation *SharedSessionBrowserBindingEvaluation,
	current *SharedSessionBrowserProfileState,
	reconnectWindow time.Duration,
) SharedSessionBrowserTopLevelBindingProjection {
	sessionID = strings.TrimSpace(sessionID)
	selectedInfo.Backend = strings.TrimSpace(selectedInfo.Backend)
	selectedInfo.Profile = strings.TrimSpace(selectedInfo.Profile)
	selectedInfo.Target = strings.TrimSpace(selectedInfo.Target)
	route = normalizeBrowserSessionRoute(route)

	projection := SharedSessionBrowserTopLevelBindingProjection{}
	if evaluation != nil {
		projection.Evaluation = *evaluation
		if len(routes) > 0 && len(projection.Evaluation.Routes) == 0 {
			projection.Evaluation.Routes = append([]SharedSessionBrowserRouteSnapshot(nil), routes...)
		}
	} else {
		projection.Evaluation = EvaluateSharedSessionBrowserBindingForScope(
			sessionID,
			selectedInfo,
			route,
			routes,
			registry,
			runRegistry,
			stateRegistry,
			reconnectWindow,
		)
	}

	if current != nil {
		merged := *current
		merged.Backend = strings.TrimSpace(merged.Backend)
		merged.Profile = strings.TrimSpace(merged.Profile)
		merged.RuntimeTarget = strings.TrimSpace(merged.RuntimeTarget)
		merged.BrowserApp = strings.TrimSpace(merged.BrowserApp)
		merged.Status = strings.TrimSpace(merged.Status)
		merged.Note = strings.TrimSpace(merged.Note)
		projection.Evaluation = MergeSharedSessionBrowserBindingEvaluationProfileState(
			stateRegistry,
			sessionID,
			selectedInfo,
			firstNonEmptyBindingString(strings.TrimSpace(selectedInfo.Profile), strings.TrimSpace(route.Profile)),
			projection.Evaluation,
			merged,
			reconnectWindow,
		)
	}

	if selection := projection.Evaluation.Snapshot.SelectedProfileSelection; selection != nil {
		cloned := *selection
		projection.ProfileSelection = &cloned
	}
	if selection := projection.Evaluation.Snapshot.SelectedTargetSelection; selection != nil {
		cloned := *selection
		projection.TargetSelection = &cloned
	}
	return projection
}

func sharedSessionBrowserTopLevelBindingSelectedInfo(
	evaluation SharedSessionBrowserBindingEvaluation,
	current *SharedSessionBrowserProfileState,
) BrowserRuntimeInfo {
	if selection := evaluation.Snapshot.SelectedProfileSelection; selection != nil {
		info := BrowserRuntimeInfo{
			Backend: strings.TrimSpace(selection.Backend),
			Profile: strings.TrimSpace(selection.Profile),
			Target:  strings.TrimSpace(selection.RuntimeTarget),
		}
		if info != (BrowserRuntimeInfo{}) {
			return info
		}
	}
	if selection := evaluation.Snapshot.SelectedTargetSelection; selection != nil {
		info := BrowserRuntimeInfo{
			Backend: strings.TrimSpace(selection.Backend),
			Profile: strings.TrimSpace(selection.Profile),
			Target:  strings.TrimSpace(selection.RuntimeTarget),
		}
		if info != (BrowserRuntimeInfo{}) {
			return info
		}
	}
	if current != nil {
		info := BrowserRuntimeInfo{
			Backend: strings.TrimSpace(current.Backend),
			Profile: strings.TrimSpace(current.Profile),
			Target:  strings.TrimSpace(current.RuntimeTarget),
		}
		if info != (BrowserRuntimeInfo{}) {
			return info
		}
	}
	for _, state := range evaluation.Snapshot.Profiles {
		info := BrowserRuntimeInfo{
			Backend: strings.TrimSpace(state.Backend),
			Profile: strings.TrimSpace(state.Profile),
			Target:  strings.TrimSpace(state.RuntimeTarget),
		}
		if info != (BrowserRuntimeInfo{}) {
			return info
		}
	}
	return BrowserRuntimeInfo{}
}
