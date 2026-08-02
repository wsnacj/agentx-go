package browserruntime

import (
	"context"
	"fmt"
	"strings"
)

// SharedSessionBrowserSyncSelectionResult captures the shared route-scoped
// profile/target selection outcome used by sync_session style actions.
type SharedSessionBrowserSyncSelectionResult struct {
	ProfileSelection *SharedSessionBrowserProfileSelection
	TargetSelection  *BrowserSessionTargetSelection
	Decision         string
	Ready            bool
}

// SharedSessionBrowserRememberProfileResult captures the remembered profile
// selection and any route-scoped target sync/clear that follows it.
type SharedSessionBrowserRememberProfileResult struct {
	ProfileSelection    *SharedSessionBrowserProfileSelection
	TargetSelection     *BrowserSessionTargetSelection
	SelectionProjection *SharedSessionBrowserSelectionProjection
	Decision            string
	Ready               bool
}

// SelectSharedSessionBrowserProfile validates and records a session-scoped
// preferred managed browser profile selection through the shared registry
// contract.
func SelectSharedSessionBrowserProfileWithContext(
	mutationCtx SharedSessionBrowserMutationContext,
	ctx context.Context,
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	browserApp string,
	control BrowserRuntimeControlBackend,
	validateWithProfiles bool,
	source string,
) (SharedSessionBrowserProfileSelection, string, bool, error) {
	if mutationCtx.usesWatchManagerEventSeam() {
		return SelectSharedSessionBrowserProfileEvent(
			ctx,
			mutationCtx.Registry,
			mutationCtx.RunRegistry,
			mutationCtx.StateRegistry,
			sessionID,
			selectedInfo,
			browserApp,
			control,
			validateWithProfiles,
			source,
			mutationCtx.ReconnectWindow,
		)
	}
	registry := mutationCtx.StateRegistry
	if registry == nil {
		return SharedSessionBrowserProfileSelection{}, "", false, nil
	}
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return SharedSessionBrowserProfileSelection{}, "", false, nil
	}
	profile := strings.TrimSpace(selectedInfo.Profile)
	if profile == "" {
		return SharedSessionBrowserProfileSelection{}, "", false, nil
	}
	resolvedBrowserApp := strings.TrimSpace(browserApp)
	if control != nil && validateWithProfiles {
		observation := sharedSessionBrowserObserverManager(nil, nil, nil, 0).
			ObserveProfiles(ctx, control, "", selectedInfo, "")
		if observation.ProfilesErr != nil {
			return SharedSessionBrowserProfileSelection{}, "session_profile_validation_failed", false, fmt.Errorf("browser_runtime: failed to load profiles for select_profile: %w", observation.ProfilesErr)
		}
		if observation.Profiles != nil {
			validated := false
			found := false
			resolvedBrowserApp, validated, found = ValidateSharedSessionBrowserSelectedProfile(profile, resolvedBrowserApp, *observation.Profiles)
			if validated && !found {
				return SharedSessionBrowserProfileSelection{}, "session_profile_missing", false, fmt.Errorf("browser_runtime: profile %q is not available on the selected route", profile)
			}
		}
	}
	return persistSharedSessionBrowserProfileSelection(
		registry,
		sessionID,
		SharedSessionBrowserProfileSelection{
			Backend:       strings.TrimSpace(selectedInfo.Backend),
			Profile:       profile,
			RuntimeTarget: strings.TrimSpace(selectedInfo.Target),
			BrowserApp:    resolvedBrowserApp,
			Source:        firstNonEmptyString(strings.TrimSpace(source), "session_profile_selected"),
		},
	)
}

func SelectSharedSessionBrowserProfile(
	ctx context.Context,
	registry SharedSessionBrowserStateRegistry,
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	browserApp string,
	control BrowserRuntimeControlBackend,
	validateWithProfiles bool,
	source string,
) (SharedSessionBrowserProfileSelection, string, bool, error) {
	return SelectSharedSessionBrowserProfileWithContext(
		SharedSessionBrowserMutationContext{StateRegistry: registry},
		ctx,
		sessionID,
		selectedInfo,
		browserApp,
		control,
		validateWithProfiles,
		source,
	)
}

// SyncSharedSessionBrowserRouteSelection applies the shared profile-selection
// validation/writeback and current-target sync/clear contract used by
// browser_runtime sync_session style actions.
func SyncSharedSessionBrowserRouteSelectionWithContext(
	mutationCtx SharedSessionBrowserMutationContext,
	ctx context.Context,
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	route BrowserSessionRoute,
	browserApp string,
	control BrowserRuntimeControlBackend,
	validateWithProfiles bool,
	source string,
) (SharedSessionBrowserSyncSelectionResult, error) {
	if mutationCtx.usesWatchManagerEventSeam() {
		return SyncSharedSessionBrowserRouteSelectionEvent(
			ctx,
			mutationCtx.Registry,
			mutationCtx.RunRegistry,
			mutationCtx.StateRegistry,
			sessionID,
			selectedInfo,
			route,
			browserApp,
			control,
			validateWithProfiles,
			source,
			mutationCtx.ReconnectWindow,
		)
	}
	sessionID = strings.TrimSpace(sessionID)
	route = normalizeBrowserSessionRoute(route)
	if sessionID == "" || sharedSessionBrowserRouteEmpty(route) {
		return SharedSessionBrowserSyncSelectionResult{}, nil
	}
	source = firstNonEmptyString(strings.TrimSpace(source), "sync_session")
	profileSelection, profileDecision, ok, err := SelectSharedSessionBrowserProfileWithContext(
		mutationCtx,
		ctx,
		sessionID,
		selectedInfo,
		browserApp,
		control,
		validateWithProfiles,
		source,
	)
	if err != nil {
		return SharedSessionBrowserSyncSelectionResult{Decision: "session_profile_validation_failed"}, err
	}
	result := SharedSessionBrowserSyncSelectionResult{}
	var profileSelectionPtr *SharedSessionBrowserProfileSelection
	if ok {
		result.ProfileSelection = &profileSelection
		profileSelectionPtr = &profileSelection
	}
	targetSelection, targetDecision, err := SyncOrClearSharedSessionBrowserCurrentTargetForProfileSelection(
		mutationCtx.Registry,
		sessionID,
		route,
		profileSelectionPtr,
		source,
	)
	if err != nil {
		result.Decision = "session_target_sync_failed"
		return result, err
	}
	result.TargetSelection = targetSelection
	switch {
	case result.ProfileSelection != nil && result.TargetSelection != nil:
		if profileDecision == "session_profile_already_selected" && targetDecision == "session_target_already_selected" {
			result.Decision = "session_route_already_synced"
		} else {
			result.Decision = "session_route_synced"
		}
		result.Ready = true
	case result.ProfileSelection != nil:
		if profileDecision == "session_profile_already_selected" {
			result.Decision = "session_profile_already_synced"
		} else {
			result.Decision = "session_profile_synced"
		}
		result.Ready = true
	case result.TargetSelection != nil:
		if targetDecision == "session_target_already_selected" {
			result.Decision = "session_target_already_synced"
		} else {
			result.Decision = "session_target_synced"
		}
		result.Ready = true
	default:
		result.Decision = "session_route_sync_unavailable"
	}
	return result, nil
}

func SyncSharedSessionBrowserRouteSelection(
	ctx context.Context,
	stateRegistry SharedSessionBrowserStateRegistry,
	sessionRegistry *BrowserSessionRegistry,
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	route BrowserSessionRoute,
	browserApp string,
	control BrowserRuntimeControlBackend,
	validateWithProfiles bool,
	source string,
) (SharedSessionBrowserSyncSelectionResult, error) {
	return SyncSharedSessionBrowserRouteSelectionWithContext(
		SharedSessionBrowserMutationContext{
			Registry:      sessionRegistry,
			StateRegistry: stateRegistry,
		},
		ctx,
		sessionID,
		selectedInfo,
		route,
		browserApp,
		control,
		validateWithProfiles,
		source,
	)
}

// CoordinateSharedSessionBrowserRouteSync applies the target-first sync
// semantics used by browser_runtime coordination goal=sync: if a current target
// can be refreshed in-place for the route, it preserves the target-only sync
// contract; otherwise it falls back to the full route selection sync path.
func CoordinateSharedSessionBrowserRouteSyncWithContext(
	mutationCtx SharedSessionBrowserMutationContext,
	ctx context.Context,
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	route BrowserSessionRoute,
	browserApp string,
	control BrowserRuntimeControlBackend,
	validateWithProfiles bool,
) (SharedSessionBrowserSyncSelectionResult, error) {
	if mutationCtx.usesWatchManagerEventSeam() {
		return CoordinateSharedSessionBrowserRouteSyncEvent(
			ctx,
			mutationCtx.Registry,
			mutationCtx.RunRegistry,
			mutationCtx.StateRegistry,
			sessionID,
			selectedInfo,
			route,
			browserApp,
			control,
			validateWithProfiles,
			mutationCtx.ReconnectWindow,
		)
	}
	sessionID = strings.TrimSpace(sessionID)
	route = normalizeBrowserSessionRoute(route)
	if sessionID == "" || sharedSessionBrowserRouteEmpty(route) {
		return SharedSessionBrowserSyncSelectionResult{}, nil
	}
	targetSelection, targetDecision, err := SyncSharedSessionBrowserCurrentTarget(mutationCtx.Registry, sessionID, route, "sync_session")
	if err != nil {
		return SharedSessionBrowserSyncSelectionResult{Decision: "session_target_sync_failed"}, err
	}
	if targetSelection != nil {
		result := SharedSessionBrowserSyncSelectionResult{
			TargetSelection: targetSelection,
			Ready:           true,
		}
		switch targetDecision {
		case "session_target_selected":
			result.Decision = "session_target_synced"
		case "session_target_already_selected":
			result.Decision = "session_target_already_synced"
		default:
			result.Decision = firstNonEmptyString(strings.TrimSpace(targetDecision), "session_target_synced")
		}
		return result, nil
	}
	return SyncSharedSessionBrowserRouteSelectionWithContext(
		mutationCtx,
		ctx,
		sessionID,
		selectedInfo,
		route,
		browserApp,
		control,
		validateWithProfiles,
		"sync_session",
	)
}

func CoordinateSharedSessionBrowserRouteSync(
	ctx context.Context,
	stateRegistry SharedSessionBrowserStateRegistry,
	sessionRegistry *BrowserSessionRegistry,
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	route BrowserSessionRoute,
	browserApp string,
	control BrowserRuntimeControlBackend,
	validateWithProfiles bool,
) (SharedSessionBrowserSyncSelectionResult, error) {
	return CoordinateSharedSessionBrowserRouteSyncWithContext(
		SharedSessionBrowserMutationContext{
			Registry:      sessionRegistry,
			StateRegistry: stateRegistry,
		},
		ctx,
		sessionID,
		selectedInfo,
		route,
		browserApp,
		control,
		validateWithProfiles,
	)
}

// RememberSharedSessionBrowserProfile records a remember-profile preference
// from the current execution result without requiring a separate control-plane
// candidate-building pass.
func RememberSharedSessionBrowserProfile(
	registry SharedSessionBrowserStateRegistry,
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	profileStatus *BrowserProfileStatusResult,
	preparedProfile string,
	requestedProfile string,
	requestedBrowserApp string,
) (SharedSessionBrowserProfileSelection, string, bool) {
	if registry == nil {
		return SharedSessionBrowserProfileSelection{}, "session_profile_not_remembered", false
	}
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return SharedSessionBrowserProfileSelection{}, "session_profile_not_remembered", false
	}
	selection := SharedSessionBrowserProfileSelection{
		Backend: strings.TrimSpace(firstNonEmptyString(
			firstStringValue(profileStatus, func(v *BrowserProfileStatusResult) string { return v.Backend }),
			selectedInfo.Backend,
		)),
		Profile: firstNonEmptyString(
			firstStringValue(profileStatus, func(v *BrowserProfileStatusResult) string { return v.Profile }),
			strings.TrimSpace(preparedProfile),
			strings.TrimSpace(requestedProfile),
			selectedInfo.Profile,
		),
		RuntimeTarget: strings.TrimSpace(selectedInfo.Target),
		BrowserApp: firstNonEmptyString(
			firstStringValue(profileStatus, func(v *BrowserProfileStatusResult) string { return v.BrowserApp }),
			strings.TrimSpace(requestedBrowserApp),
		),
		Source: "remember_profile",
	}
	if strings.TrimSpace(selection.Profile) == "" {
		return SharedSessionBrowserProfileSelection{}, "session_profile_not_remembered", false
	}
	selection, decision, ok, _ := persistSharedSessionBrowserProfileSelection(registry, sessionID, selection)
	if !ok {
		return SharedSessionBrowserProfileSelection{}, "session_profile_not_remembered", false
	}
	switch decision {
	case "session_profile_already_selected":
		return selection, "session_profile_already_remembered", true
	default:
		return selection, "session_profile_remembered", true
	}
}

// RememberSharedSessionBrowserProfileForRoute records a remember-profile
// selection and applies the shared current-target sync/clear semantics for the
// scoped route.
func RememberSharedSessionBrowserProfileForRouteWithContext(
	mutationCtx SharedSessionBrowserMutationContext,
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	route BrowserSessionRoute,
	profileStatus *BrowserProfileStatusResult,
	preparedProfile string,
	requestedProfile string,
	requestedBrowserApp string,
) SharedSessionBrowserRememberProfileResult {
	if mutationCtx.usesWatchManagerEventSeam() {
		return RememberSharedSessionBrowserProfileForRouteEvent(
			mutationCtx.Registry,
			mutationCtx.RunRegistry,
			mutationCtx.StateRegistry,
			sessionID,
			selectedInfo,
			route,
			profileStatus,
			preparedProfile,
			requestedProfile,
			requestedBrowserApp,
			mutationCtx.ReconnectWindow,
		)
	}
	sessionID = strings.TrimSpace(sessionID)
	selection, decision, ok := RememberSharedSessionBrowserProfile(
		mutationCtx.StateRegistry,
		sessionID,
		selectedInfo,
		profileStatus,
		preparedProfile,
		requestedProfile,
		requestedBrowserApp,
	)
	result := SharedSessionBrowserRememberProfileResult{
		Decision: decision,
	}
	if !ok {
		return result
	}
	result.ProfileSelection = &selection
	result.Ready = true
	route = normalizeBrowserSessionRoute(route)
	if mutationCtx.Registry != nil && sessionID != "" && !sharedSessionBrowserRouteEmpty(route) {
		if targetSelection, _, err := SyncOrClearSharedSessionBrowserCurrentTargetForProfileSelection(
			mutationCtx.Registry,
			sessionID,
			route,
			&selection,
			"remember_profile",
		); err == nil && targetSelection != nil {
			result.TargetSelection = targetSelection
		}
	}
	result.SelectionProjection = &SharedSessionBrowserSelectionProjection{
		ProfileSelection:   result.ProfileSelection,
		TargetSelection:    result.TargetSelection,
		ApplyTargetToRoute: true,
	}
	return result
}

func RememberSharedSessionBrowserProfileForRoute(
	stateRegistry SharedSessionBrowserStateRegistry,
	sessionRegistry *BrowserSessionRegistry,
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	route BrowserSessionRoute,
	profileStatus *BrowserProfileStatusResult,
	preparedProfile string,
	requestedProfile string,
	requestedBrowserApp string,
) SharedSessionBrowserRememberProfileResult {
	return RememberSharedSessionBrowserProfileForRouteWithContext(
		SharedSessionBrowserMutationContext{
			Registry:      sessionRegistry,
			StateRegistry: stateRegistry,
		},
		sessionID,
		selectedInfo,
		route,
		profileStatus,
		preparedProfile,
		requestedProfile,
		requestedBrowserApp,
	)
}

// PromoteSharedSessionBrowserProfileFromTargetSelection records a remembered
// profile selection derived from a current target selection when the target has
// enough route/profile identity to safely promote.
func PromoteSharedSessionBrowserProfileFromTargetSelection(
	registry SharedSessionBrowserStateRegistry,
	sessionID string,
	selection *BrowserSessionTargetSelection,
) (SharedSessionBrowserProfileSelection, bool) {
	sessionID = strings.TrimSpace(sessionID)
	if selection == nil || registry == nil || sessionID == "" {
		return SharedSessionBrowserProfileSelection{}, false
	}
	profile := strings.TrimSpace(selection.Profile)
	runtimeTarget := strings.TrimSpace(selection.RuntimeTarget)
	if profile == "" || runtimeTarget == "" {
		return SharedSessionBrowserProfileSelection{}, false
	}
	next := SharedSessionBrowserProfileSelection{
		Backend:       browserSessionCanonicalBackend(strings.TrimSpace(selection.Backend)),
		Profile:       profile,
		RuntimeTarget: runtimeTarget,
		BrowserApp:    strings.TrimSpace(selection.BrowserApp),
		Source:        firstNonEmptyString(strings.TrimSpace(selection.Source), "select_target"),
	}
	if current, ok := registry.SelectedBrowserProfile(sessionID, runtimeTarget); ok {
		if sameSharedSessionBrowserProfileSelection(current, next) {
			return current, true
		}
	}
	registry.SelectBrowserProfile(sessionID, next)
	return next, true
}

func sameSharedSessionBrowserProfileSelection(left SharedSessionBrowserProfileSelection, right SharedSessionBrowserProfileSelection) bool {
	return strings.EqualFold(strings.TrimSpace(left.Backend), strings.TrimSpace(right.Backend)) &&
		strings.EqualFold(strings.TrimSpace(left.Profile), strings.TrimSpace(right.Profile)) &&
		strings.EqualFold(strings.TrimSpace(left.RuntimeTarget), strings.TrimSpace(right.RuntimeTarget)) &&
		(strings.TrimSpace(left.BrowserApp) == "" || strings.TrimSpace(right.BrowserApp) == "" ||
			strings.EqualFold(strings.TrimSpace(left.BrowserApp), strings.TrimSpace(right.BrowserApp)))
}

func persistSharedSessionBrowserProfileSelection(registry SharedSessionBrowserStateRegistry, sessionID string, selection SharedSessionBrowserProfileSelection) (SharedSessionBrowserProfileSelection, string, bool, error) {
	if registry == nil {
		return SharedSessionBrowserProfileSelection{}, "", false, nil
	}
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return SharedSessionBrowserProfileSelection{}, "", false, nil
	}
	if strings.TrimSpace(selection.Profile) == "" {
		return SharedSessionBrowserProfileSelection{}, "", false, nil
	}
	if existing, ok := registry.SelectedBrowserProfile(sessionID, selection.RuntimeTarget); ok {
		if sameSharedSessionBrowserProfileSelection(existing, selection) {
			if strings.TrimSpace(selection.BrowserApp) == "" {
				selection.BrowserApp = strings.TrimSpace(existing.BrowserApp)
			}
			if strings.TrimSpace(selection.Source) == "" {
				selection.Source = strings.TrimSpace(existing.Source)
			}
			registry.SelectBrowserProfile(sessionID, selection)
			return selection, "session_profile_already_selected", true, nil
		}
	}
	registry.SelectBrowserProfile(sessionID, selection)
	return selection, "session_profile_selected", true, nil
}

func firstStringValue[T any](item *T, pick func(*T) string) string {
	if item == nil {
		return ""
	}
	return strings.TrimSpace(pick(item))
}

func sharedSessionBrowserRouteEmpty(route BrowserSessionRoute) bool {
	route = normalizeBrowserSessionRoute(route)
	return route == (BrowserSessionRoute{})
}
