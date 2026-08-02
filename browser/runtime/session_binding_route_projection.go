package browserruntime

import "strings"

// SharedSessionBrowserBindingRouteProjection captures the shared route/configured
// runtime identity that tools payloads can bridge from a lifecycle-owned binding
// evaluation without rebuilding selection/default-profile precedence locally.
type SharedSessionBrowserBindingRouteProjection struct {
	SelectedRouteInfo BrowserRuntimeInfo
	ConfiguredInfo    BrowserRuntimeInfo
	DefaultProfile    string
}

// ProjectSharedSessionBrowserBindingRouteProjection lowers a shared binding
// evaluation plus any tool-facing selection overlay onto the stable
// selected-route/configured/default-profile contract used by tools payloads.
func ProjectSharedSessionBrowserBindingRouteProjection(
	evaluation SharedSessionBrowserBindingEvaluation,
	projection *SharedSessionBrowserSelectionProjection,
	explicitDefaultProfile string,
) SharedSessionBrowserBindingRouteProjection {
	hydratedProjection := sharedSessionBrowserRequestedSelectionProjection(
		evaluation,
		projection,
	)
	storedProjection := ProjectSharedSessionBrowserSelectionProjectionFromBindingEvaluation(
		evaluation,
		nil,
	)
	projected := SharedSessionBrowserBindingRouteProjection{
		SelectedRouteInfo: sharedSessionBrowserNormalizedRuntimeInfo(
			sharedSessionBrowserBindingRouteInfo(
				hydratedProjection,
				storedProjection,
				evaluation,
			),
		),
		ConfiguredInfo: sharedSessionBrowserNormalizedRuntimeInfo(
			sharedSessionBrowserBindingConfiguredInfo(
				hydratedProjection,
				storedProjection,
				evaluation,
			),
		),
		DefaultProfile: strings.TrimSpace(explicitDefaultProfile),
	}
	if projected.DefaultProfile == "" {
		projected.DefaultProfile = strings.TrimSpace(sharedSessionBrowserBindingDefaultProfile(
			hydratedProjection,
			storedProjection,
		))
	}
	if projected.DefaultProfile == "" {
		sessionProjection := ProjectSharedSessionBrowserTopLevelSessionFromBindingEvaluation(evaluation)
		projected.DefaultProfile = sharedSessionBrowserProfileInventoryDefaultProfile(
			sessionProjection.Profiles,
			projected.ConfiguredInfo.Profile,
		)
	}
	return projected
}

func sharedSessionBrowserBindingRouteInfo(
	requested *SharedSessionBrowserSelectionProjection,
	stored *SharedSessionBrowserSelectionProjection,
	evaluation SharedSessionBrowserBindingEvaluation,
) BrowserRuntimeInfo {
	if info := sharedSessionBrowserBindingInfoFromTargetSelection(firstTargetSelectionValue(requested)); info != (BrowserRuntimeInfo{}) {
		return info
	}
	if info := sharedSessionBrowserBindingInfoFromProfileSelection(firstProfileSelectionValue(requested)); info != (BrowserRuntimeInfo{}) {
		return info
	}
	if info := sharedSessionBrowserBindingInfoFromTargetSelection(firstTargetSelectionValue(stored)); info != (BrowserRuntimeInfo{}) {
		return info
	}
	if info := sharedSessionBrowserBindingInfoFromProfileSelection(firstProfileSelectionValue(stored)); info != (BrowserRuntimeInfo{}) {
		return info
	}
	fallback := evaluation
	fallback.Snapshot.SelectedProfileSelection = nil
	fallback.Snapshot.SelectedTargetSelection = nil
	return sharedSessionBrowserNormalizedRuntimeInfo(
		sharedSessionBrowserTopLevelBindingSelectedInfo(fallback, nil),
	)
}

func sharedSessionBrowserBindingConfiguredInfo(
	requested *SharedSessionBrowserSelectionProjection,
	stored *SharedSessionBrowserSelectionProjection,
	evaluation SharedSessionBrowserBindingEvaluation,
) BrowserRuntimeInfo {
	if info := sharedSessionBrowserBindingInfoFromProfileSelection(firstProfileSelectionValue(requested)); info != (BrowserRuntimeInfo{}) {
		return info
	}
	if info := sharedSessionBrowserBindingInfoFromTargetSelection(firstTargetSelectionValue(requested)); info != (BrowserRuntimeInfo{}) {
		return info
	}
	if info := sharedSessionBrowserBindingInfoFromProfileSelection(firstProfileSelectionValue(stored)); info != (BrowserRuntimeInfo{}) {
		return info
	}
	if info := sharedSessionBrowserBindingInfoFromTargetSelection(firstTargetSelectionValue(stored)); info != (BrowserRuntimeInfo{}) {
		return info
	}
	return sharedSessionBrowserNormalizedRuntimeInfo(
		sharedSessionBrowserTopLevelBindingSelectedInfo(evaluation, nil),
	)
}

func sharedSessionBrowserBindingDefaultProfile(
	requested *SharedSessionBrowserSelectionProjection,
	stored *SharedSessionBrowserSelectionProjection,
) string {
	return strings.TrimSpace(firstNonEmptyBindingString(
		firstStringValue(firstProfileSelectionValue(requested), func(v *SharedSessionBrowserProfileSelection) string { return v.Profile }),
		firstStringValue(firstTargetSelectionValue(requested), func(v *BrowserSessionTargetSelection) string { return v.Profile }),
		firstStringValue(firstProfileSelectionValue(stored), func(v *SharedSessionBrowserProfileSelection) string { return v.Profile }),
		firstStringValue(firstTargetSelectionValue(stored), func(v *BrowserSessionTargetSelection) string { return v.Profile }),
	))
}

func sharedSessionBrowserBindingInfoFromProfileSelection(
	selection *SharedSessionBrowserProfileSelection,
) BrowserRuntimeInfo {
	if selection == nil {
		return BrowserRuntimeInfo{}
	}
	return sharedSessionBrowserNormalizedRuntimeInfo(BrowserRuntimeInfo{
		Backend: selection.Backend,
		Profile: selection.Profile,
		Target:  selection.RuntimeTarget,
	})
}

func sharedSessionBrowserBindingInfoFromTargetSelection(
	selection *BrowserSessionTargetSelection,
) BrowserRuntimeInfo {
	if selection == nil {
		return BrowserRuntimeInfo{}
	}
	return sharedSessionBrowserNormalizedRuntimeInfo(BrowserRuntimeInfo{
		Backend: selection.Backend,
		Profile: selection.Profile,
		Target:  selection.RuntimeTarget,
	})
}

func sharedSessionBrowserNormalizedRuntimeInfo(info BrowserRuntimeInfo) BrowserRuntimeInfo {
	info.Backend = strings.TrimSpace(info.Backend)
	info.Profile = strings.TrimSpace(info.Profile)
	info.Target = strings.TrimSpace(info.Target)
	return info
}

func sharedSessionBrowserRequestedSelectionProjection(
	evaluation SharedSessionBrowserBindingEvaluation,
	projection *SharedSessionBrowserSelectionProjection,
) *SharedSessionBrowserSelectionProjection {
	if projection == nil {
		return nil
	}
	profileFallback := firstTargetSelectionValue(projection)
	if profileFallback == nil && evaluation.Snapshot.SelectedTargetSelection != nil {
		profileFallback = evaluation.Snapshot.SelectedTargetSelection
	}
	profileSelection := ProjectSharedSessionBrowserProfileSelectionFromBindingSnapshot(
		firstProfileSelectionValue(projection),
		profileFallback,
		evaluation.Snapshot.Profiles,
	)
	targetFallback := profileSelection
	if targetFallback == nil && evaluation.Snapshot.SelectedProfileSelection != nil {
		targetFallback = evaluation.Snapshot.SelectedProfileSelection
	}
	targetSelection := ProjectSharedSessionBrowserTargetSelectionFromBindingSnapshot(
		firstTargetSelectionValue(projection),
		targetFallback,
		evaluation.Snapshot.Profiles,
	)
	if profileSelection == nil && targetSelection == nil && !projection.ApplyTargetToRoute {
		return nil
	}
	return &SharedSessionBrowserSelectionProjection{
		ProfileSelection:   profileSelection,
		TargetSelection:    targetSelection,
		ApplyTargetToRoute: projection.ApplyTargetToRoute,
	}
}

func firstProfileSelectionValue(
	projection *SharedSessionBrowserSelectionProjection,
) *SharedSessionBrowserProfileSelection {
	if projection == nil {
		return nil
	}
	return projection.ProfileSelection
}

func firstTargetSelectionValue(
	projection *SharedSessionBrowserSelectionProjection,
) *BrowserSessionTargetSelection {
	if projection == nil {
		return nil
	}
	return projection.TargetSelection
}
