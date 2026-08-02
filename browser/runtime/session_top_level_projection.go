package browserruntime

import "strings"

// SharedSessionBrowserTopLevelSessionProjection captures the shared
// route/run/profile projection that runtime sessions/workbench style payloads
// surface at the top level.
type SharedSessionBrowserTopLevelSessionProjection struct {
	Routes      []SharedSessionBrowserRouteSnapshot
	TargetCount int
	Runs        []SharedSessionRunInfo
	Profiles    []SharedSessionBrowserProjectedProfileState
	Handoff     *SharedSessionBrowserSessionHandoffSummary
}

// ProjectSharedSessionBrowserTopLevelSessionFromBindingEvaluation lowers a
// lifecycle-owned binding evaluation onto the shared top-level session
// projection contract. This keeps finalized session/workbench surfaces on the
// same routes/runs/profile snapshot even when a selected route is no longer
// available at the call site.
func ProjectSharedSessionBrowserTopLevelSessionFromBindingEvaluation(
	evaluation SharedSessionBrowserBindingEvaluation,
) SharedSessionBrowserTopLevelSessionProjection {
	projection := SharedSessionBrowserTopLevelSessionProjection{
		TargetCount: evaluation.Snapshot.Summary.RouteTargetCount,
	}
	if len(evaluation.Routes) > 0 {
		projection.Routes = append([]SharedSessionBrowserRouteSnapshot(nil), evaluation.Routes...)
		if projection.TargetCount <= 0 {
			projection.TargetCount = SharedSessionBrowserRouteTargetCount(evaluation.Routes)
		}
	} else if route := sharedSessionBrowserMinimalRouteSnapshotFromBindingSnapshot(evaluation.Snapshot); route != nil {
		projection.Routes = []SharedSessionBrowserRouteSnapshot{*route}
		if projection.TargetCount <= 0 {
			projection.TargetCount = 1
		}
	}
	if len(evaluation.Snapshot.Runs) > 0 {
		projection.Runs = append([]SharedSessionRunInfo(nil), evaluation.Snapshot.Runs...)
	}
	projected := ProjectSharedSessionBrowserProfileSnapshot(
		evaluation.Snapshot.Profiles,
		evaluation.Snapshot.SelectedProfileSelection,
	)
	if len(projected) == 0 {
		projected = sharedSessionBrowserFallbackProjectedProfilesFromBindingEvaluation(
			evaluation,
			projection.Routes,
		)
	}
	if len(projected) > 0 {
		projection.Profiles = projected
	}
	projection.Handoff = sharedSessionBrowserBindingEvaluationHandoff(
		evaluation,
		projection.Routes,
		projection.Runs,
		projection.Profiles,
		projection.TargetCount,
	)
	return projection
}

func sharedSessionBrowserFallbackProjectedProfilesFromBindingEvaluation(
	evaluation SharedSessionBrowserBindingEvaluation,
	routes []SharedSessionBrowserRouteSnapshot,
) []SharedSessionBrowserProjectedProfileState {
	if len(routes) == 0 {
		return nil
	}
	var selection *SharedSessionBrowserProfileSelection
	if evaluation.Snapshot.SelectedProfileSelection != nil {
		cloned := *evaluation.Snapshot.SelectedProfileSelection
		selection = &cloned
	}
	route := BrowserRuntimeInfo{}
	if targetSelection := evaluation.Snapshot.SelectedTargetSelection; targetSelection != nil {
		route = BrowserRuntimeInfo{
			Backend: strings.TrimSpace(targetSelection.Backend),
			Profile: strings.TrimSpace(targetSelection.Profile),
			Target:  strings.TrimSpace(targetSelection.RuntimeTarget),
		}
	}
	if route == (BrowserRuntimeInfo{}) && selection != nil {
		route = BrowserRuntimeInfo{
			Backend: strings.TrimSpace(selection.Backend),
			Profile: strings.TrimSpace(selection.Profile),
			Target:  strings.TrimSpace(selection.RuntimeTarget),
		}
	}
	return ProjectSharedSessionBrowserFallbackProfilesFromRouteSnapshots(
		route,
		strings.TrimSpace(route.Profile),
		routes,
		selection,
	)
}

func sharedSessionBrowserMinimalRouteSnapshotFromBindingSnapshot(
	snapshot SharedSessionBrowserBindingSnapshot,
) *SharedSessionBrowserRouteSnapshot {
	var (
		targetSelection  *BrowserSessionTargetSelection
		profileSelection *SharedSessionBrowserProfileSelection
	)
	if snapshot.SelectedTargetSelection != nil {
		cloned := *snapshot.SelectedTargetSelection
		targetSelection = &cloned
	}
	if snapshot.SelectedProfileSelection != nil {
		cloned := *snapshot.SelectedProfileSelection
		profileSelection = &cloned
	}
	backend := strings.TrimSpace(firstNonEmptyString(
		func() string {
			if targetSelection != nil {
				return targetSelection.Backend
			}
			return ""
		}(),
		func() string {
			if profileSelection != nil {
				return profileSelection.Backend
			}
			return ""
		}(),
	))
	profile := strings.TrimSpace(firstNonEmptyString(
		func() string {
			if targetSelection != nil {
				return targetSelection.Profile
			}
			return ""
		}(),
		func() string {
			if profileSelection != nil {
				return profileSelection.Profile
			}
			return ""
		}(),
	))
	runtimeTarget := strings.TrimSpace(firstNonEmptyString(
		func() string {
			if targetSelection != nil {
				return targetSelection.RuntimeTarget
			}
			return ""
		}(),
		func() string {
			if profileSelection != nil {
				return profileSelection.RuntimeTarget
			}
			return ""
		}(),
	))
	browserApp := strings.TrimSpace(firstNonEmptyString(
		func() string {
			if targetSelection != nil {
				return targetSelection.BrowserApp
			}
			return ""
		}(),
		func() string {
			if profileSelection != nil {
				return profileSelection.BrowserApp
			}
			return ""
		}(),
	))
	currentTargetID := strings.TrimSpace(firstNonEmptyString(
		snapshot.CurrentTargetID,
		func() string {
			if targetSelection != nil {
				return targetSelection.ID
			}
			return ""
		}(),
	))
	currentTargetSource := ""
	if targetSelection != nil {
		currentTargetSource = strings.TrimSpace(targetSelection.Source)
	}
	if backend == "" && profile == "" && runtimeTarget == "" && currentTargetID == "" {
		return nil
	}
	return &SharedSessionBrowserRouteSnapshot{
		Backend:             backend,
		Profile:             profile,
		RuntimeTarget:       runtimeTarget,
		BrowserApp:          browserApp,
		CurrentTargetID:     currentTargetID,
		CurrentTargetSource: currentTargetSource,
	}
}

// ProjectSharedSessionBrowserTopLevelSessionView clones a session view snapshot
// into the shared top-level projection contract and optionally falls back to a
// supplied projected profile set when the raw session view omits profiles.
func ProjectSharedSessionBrowserTopLevelSessionView(
	view SharedSessionBrowserSessionViewSnapshot,
	fallbackProfiles []SharedSessionBrowserProjectedProfileState,
) SharedSessionBrowserTopLevelSessionProjection {
	projection := SharedSessionBrowserTopLevelSessionProjection{
		TargetCount: view.TargetCount,
	}
	if len(view.Routes) > 0 {
		projection.Routes = append([]SharedSessionBrowserRouteSnapshot(nil), view.Routes...)
		if projection.TargetCount <= 0 {
			projection.TargetCount = SharedSessionBrowserRouteTargetCount(view.Routes)
		}
	}
	if len(view.Runs) > 0 {
		projection.Runs = append([]SharedSessionRunInfo(nil), view.Runs...)
	}
	profiles := view.Profiles
	if len(profiles) == 0 && len(fallbackProfiles) > 0 {
		profiles = fallbackProfiles
	}
	if len(profiles) > 0 {
		projection.Profiles = append([]SharedSessionBrowserProjectedProfileState(nil), profiles...)
	}
	if view.Handoff != nil && len(fallbackProfiles) == 0 {
		projection.Handoff = CloneSharedSessionBrowserSessionHandoffSummary(view.Handoff)
	} else {
		projection.Handoff = BuildSharedSessionBrowserSessionHandoffSummary(SharedSessionBrowserSessionHandoffRequest{
			Routes:      projection.Routes,
			Runs:        projection.Runs,
			Profiles:    projection.Profiles,
			TargetCount: projection.TargetCount,
		})
	}
	return projection
}

// ProjectSharedSessionBrowserFallbackProfilesFromRouteSnapshots synthesizes a
// projected managed-profile snapshot from route-scoped session targets when no
// state-registry backed session profiles are available.
func ProjectSharedSessionBrowserFallbackProfilesFromRouteSnapshots(
	route BrowserRuntimeInfo,
	requestedProfile string,
	routes []SharedSessionBrowserRouteSnapshot,
	selection *SharedSessionBrowserProfileSelection,
) []SharedSessionBrowserProjectedProfileState {
	snapshot := sharedSessionBrowserProfileSnapshotFromRouteSnapshots(route, requestedProfile, routes)
	if len(snapshot) == 0 {
		return nil
	}
	sharedSelection := selection
	if sharedSelection == nil {
		profile := strings.TrimSpace(firstNonEmptyString(requestedProfile, route.Profile))
		if profile != "" {
			sharedSelection = &SharedSessionBrowserProfileSelection{
				Backend:       strings.TrimSpace(route.Backend),
				Profile:       profile,
				RuntimeTarget: strings.TrimSpace(route.Target),
				Source:        "requested_profile",
			}
		}
	}
	return ProjectSharedSessionBrowserProfileSnapshot(snapshot, sharedSelection)
}

func sharedSessionBrowserProfileSnapshotFromRouteSnapshots(
	route BrowserRuntimeInfo,
	requestedProfile string,
	routes []SharedSessionBrowserRouteSnapshot,
) []SharedSessionBrowserProfileState {
	if len(routes) == 0 {
		return nil
	}
	route.Backend = strings.TrimSpace(route.Backend)
	route.Profile = strings.TrimSpace(route.Profile)
	route.Target = strings.TrimSpace(route.Target)
	requestedProfile = strings.TrimSpace(requestedProfile)

	seen := map[string]bool{}
	out := make([]SharedSessionBrowserProfileState, 0, len(routes))
	for _, item := range routes {
		backend := strings.TrimSpace(item.Backend)
		profile := strings.TrimSpace(item.Profile)
		runtimeTarget := strings.TrimSpace(item.RuntimeTarget)
		if route.Backend != "" && browserSessionCanonicalBackend(backend) != browserSessionCanonicalBackend(route.Backend) {
			continue
		}
		if route.Target != "" && !strings.EqualFold(runtimeTarget, route.Target) {
			continue
		}
		if requestedProfile != "" && !strings.EqualFold(profile, requestedProfile) {
			continue
		}
		if profile == "" {
			continue
		}
		key := strings.Join([]string{
			browserSessionCanonicalBackend(backend),
			strings.ToLower(runtimeTarget),
			strings.ToLower(profile),
		}, "|")
		if seen[key] {
			continue
		}
		seen[key] = true
		browserApp := strings.TrimSpace(item.BrowserApp)
		if browserApp == "" {
			for _, target := range item.Targets {
				browserApp = strings.TrimSpace(target.BrowserApp)
				if browserApp != "" {
					break
				}
			}
		}
		out = append(out, SharedSessionBrowserProfileState{
			Backend:       backend,
			Profile:       profile,
			RuntimeTarget: runtimeTarget,
			BrowserApp:    browserApp,
			Note:          "cached route-scoped session snapshot",
		})
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
