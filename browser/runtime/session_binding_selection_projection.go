package browserruntime

import "strings"

// ProjectSharedSessionBrowserProfileSelectionFromBindingSnapshot hydrates a
// binding-shell profile selection from the shared profile snapshot so tools
// callers do not need to re-derive backend/profile/runtime-target/browser-app
// fallback locally.
func ProjectSharedSessionBrowserProfileSelectionFromBindingSnapshot(
	selection *SharedSessionBrowserProfileSelection,
	fallbackTargetSelection *BrowserSessionTargetSelection,
	profiles []SharedSessionBrowserProfileState,
) *SharedSessionBrowserProfileSelection {
	if selection == nil {
		return nil
	}
	projected := *selection
	state := sharedSessionBrowserBindingProfileStateForProjection(
		profiles,
		firstNonEmptyBindingString(
			strings.TrimSpace(projected.Backend),
			firstStringValue(fallbackTargetSelection, func(v *BrowserSessionTargetSelection) string { return v.Backend }),
		),
		firstNonEmptyBindingString(
			strings.TrimSpace(projected.Profile),
			firstStringValue(fallbackTargetSelection, func(v *BrowserSessionTargetSelection) string { return v.Profile }),
		),
		firstNonEmptyBindingString(
			strings.TrimSpace(projected.RuntimeTarget),
			firstStringValue(fallbackTargetSelection, func(v *BrowserSessionTargetSelection) string { return v.RuntimeTarget }),
		),
	)
	if state != nil {
		if strings.TrimSpace(projected.Backend) == "" {
			projected.Backend = strings.TrimSpace(state.Backend)
		}
		if strings.TrimSpace(projected.Profile) == "" {
			projected.Profile = strings.TrimSpace(state.Profile)
		}
		if strings.TrimSpace(projected.RuntimeTarget) == "" {
			projected.RuntimeTarget = strings.TrimSpace(state.RuntimeTarget)
		}
		if strings.TrimSpace(projected.BrowserApp) == "" {
			projected.BrowserApp = strings.TrimSpace(state.BrowserApp)
		}
	}
	if sharedSessionBrowserProfileSelectionEmpty(projected) {
		return nil
	}
	return &projected
}

// ProjectSharedSessionBrowserTargetSelectionFromBindingSnapshot hydrates a
// binding-shell target selection from the shared profile snapshot so tools
// callers do not need to re-derive route/profile/browser-app fallback locally.
func ProjectSharedSessionBrowserTargetSelectionFromBindingSnapshot(
	selection *BrowserSessionTargetSelection,
	fallbackProfileSelection *SharedSessionBrowserProfileSelection,
	profiles []SharedSessionBrowserProfileState,
) *BrowserSessionTargetSelection {
	if selection == nil {
		return nil
	}
	projected := *selection
	state := sharedSessionBrowserBindingProfileStateForProjection(
		profiles,
		firstNonEmptyBindingString(
			strings.TrimSpace(projected.Backend),
			firstStringValue(fallbackProfileSelection, func(v *SharedSessionBrowserProfileSelection) string { return v.Backend }),
		),
		firstNonEmptyBindingString(
			strings.TrimSpace(projected.Profile),
			firstStringValue(fallbackProfileSelection, func(v *SharedSessionBrowserProfileSelection) string { return v.Profile }),
		),
		firstNonEmptyBindingString(
			strings.TrimSpace(projected.RuntimeTarget),
			firstStringValue(fallbackProfileSelection, func(v *SharedSessionBrowserProfileSelection) string { return v.RuntimeTarget }),
		),
	)
	if state != nil {
		if strings.TrimSpace(projected.Backend) == "" {
			projected.Backend = strings.TrimSpace(state.Backend)
		}
		if strings.TrimSpace(projected.Profile) == "" {
			projected.Profile = strings.TrimSpace(state.Profile)
		}
		if strings.TrimSpace(projected.RuntimeTarget) == "" {
			projected.RuntimeTarget = strings.TrimSpace(state.RuntimeTarget)
		}
		if strings.TrimSpace(projected.BrowserApp) == "" {
			projected.BrowserApp = strings.TrimSpace(state.BrowserApp)
		}
	}
	if sharedSessionBrowserTargetSelectionEmpty(projected) {
		return nil
	}
	return &projected
}

// ProjectSharedSessionBrowserSelectionProjectionFromBindingEvaluation applies a
// tool-facing selection overlay to a shared binding evaluation and then
// reprojects profile/target selections through the shared profile snapshot so
// callers do not need to locally rehydrate browser-app or route identity
// fields from payload/profile status fallbacks.
func ProjectSharedSessionBrowserSelectionProjectionFromBindingEvaluation(
	evaluation SharedSessionBrowserBindingEvaluation,
	projection *SharedSessionBrowserSelectionProjection,
) *SharedSessionBrowserSelectionProjection {
	projectedEvaluation := ProjectSharedSessionBrowserBindingEvaluationWithSelectionProjection(
		evaluation,
		projection,
	)
	profileSelection := ProjectSharedSessionBrowserProfileSelectionFromBindingSnapshot(
		projectedEvaluation.Snapshot.SelectedProfileSelection,
		projectedEvaluation.Snapshot.SelectedTargetSelection,
		projectedEvaluation.Snapshot.Profiles,
	)
	targetSelection := ProjectSharedSessionBrowserTargetSelectionFromBindingSnapshot(
		projectedEvaluation.Snapshot.SelectedTargetSelection,
		profileSelection,
		projectedEvaluation.Snapshot.Profiles,
	)
	applyTargetToRoute := false
	if projection != nil {
		applyTargetToRoute = projection.ApplyTargetToRoute
	}
	if profileSelection == nil && targetSelection == nil && !applyTargetToRoute {
		return nil
	}
	return &SharedSessionBrowserSelectionProjection{
		ProfileSelection:   profileSelection,
		TargetSelection:    targetSelection,
		ApplyTargetToRoute: applyTargetToRoute,
	}
}

func sharedSessionBrowserBindingProfileStateForProjection(
	profiles []SharedSessionBrowserProfileState,
	backend string,
	profile string,
	runtimeTarget string,
) *SharedSessionBrowserProfileState {
	if len(profiles) == 0 {
		return nil
	}
	backend = strings.TrimSpace(backend)
	profile = strings.TrimSpace(profile)
	runtimeTarget = strings.TrimSpace(runtimeTarget)
	match := func(state SharedSessionBrowserProfileState, allowPartial bool) bool {
		stateBackend := strings.TrimSpace(state.Backend)
		stateProfile := strings.TrimSpace(state.Profile)
		stateTarget := strings.TrimSpace(state.RuntimeTarget)
		if backend != "" && browserSessionCanonicalBackend(stateBackend) != browserSessionCanonicalBackend(backend) {
			return false
		}
		if profile != "" && !strings.EqualFold(stateProfile, profile) {
			return false
		}
		if runtimeTarget != "" && !strings.EqualFold(stateTarget, runtimeTarget) {
			return false
		}
		if !allowPartial && (profile == "" || runtimeTarget == "") {
			return false
		}
		return true
	}
	for i := range profiles {
		if match(profiles[i], false) {
			state := profiles[i]
			return &state
		}
	}
	for i := range profiles {
		if match(profiles[i], true) {
			state := profiles[i]
			return &state
		}
	}
	if len(profiles) == 1 {
		state := profiles[0]
		return &state
	}
	return nil
}

func sharedSessionBrowserProfileSelectionEmpty(selection SharedSessionBrowserProfileSelection) bool {
	return strings.TrimSpace(selection.Backend) == "" &&
		strings.TrimSpace(selection.Profile) == "" &&
		strings.TrimSpace(selection.RuntimeTarget) == "" &&
		strings.TrimSpace(selection.BrowserApp) == "" &&
		strings.TrimSpace(selection.Source) == ""
}

func sharedSessionBrowserTargetSelectionEmpty(selection BrowserSessionTargetSelection) bool {
	return strings.TrimSpace(selection.ID) == "" &&
		selection.TabIndex <= 0 &&
		strings.TrimSpace(selection.URL) == "" &&
		strings.TrimSpace(selection.Title) == "" &&
		strings.TrimSpace(selection.Backend) == "" &&
		strings.TrimSpace(selection.Profile) == "" &&
		strings.TrimSpace(selection.RuntimeTarget) == "" &&
		strings.TrimSpace(selection.BrowserApp) == "" &&
		strings.TrimSpace(selection.Source) == ""
}
