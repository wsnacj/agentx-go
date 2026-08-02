package browserruntime

import "strings"

// SharedSessionBrowserExecutionSurface captures the lifecycle-owned profile
// inventory and cleanup surface derived from an applied execution result.
type SharedSessionBrowserExecutionSurface struct {
	Note                  string
	ProfileState          SharedSessionBrowserProfileState
	HasProfileState       bool
	ProfileStatus         BrowserProfileStatusResult
	HasProfileStatus      bool
	Profiles              []SharedSessionBrowserProjectedProfileState
	DiscoveredProfiles    []string
	DefaultProfile        string
	ApplyProfileInventory bool
	ClearedSessionTargets int
}

// SharedSessionBrowserExecutionInventoryProjection captures the lifecycle-owned
// profile status/inventory slice that tools can bridge into payload-specific
// inventory fields without re-deriving fallback rules locally.
type SharedSessionBrowserExecutionInventoryProjection struct {
	ProfileState          SharedSessionBrowserProfileState
	HasProfileState       bool
	ProfileStatus         BrowserProfileStatusResult
	HasProfileStatus      bool
	Profiles              []SharedSessionBrowserProjectedProfileState
	DiscoveredProfiles    []string
	DefaultProfile        string
	ApplyProfileInventory bool
}

// SharedSessionBrowserExecutionSurfaceProjection captures the lifecycle-owned
// execution surface plus the narrowed inventory slice that tools bridge into
// payload-specific action/profile-inventory fields.
type SharedSessionBrowserExecutionSurfaceProjection struct {
	Surface             SharedSessionBrowserExecutionSurface
	InventoryProjection *SharedSessionBrowserExecutionInventoryProjection
}

// ProjectSharedSessionBrowserExecutionSurface projects a shared lifecycle
// execution application into the stable tool-facing inventory/cleanup surface.
func ProjectSharedSessionBrowserExecutionSurface(
	selectedInfo BrowserRuntimeInfo,
	result SharedSessionBrowserExecutionResult,
	application SharedSessionBrowserExecutionApplication,
) SharedSessionBrowserExecutionSurface {
	surface := SharedSessionBrowserExecutionSurface{
		Note:                  strings.TrimSpace(sharedSessionBrowserExecutionProfilesNote(result)),
		ClearedSessionTargets: application.Cleanup.ClearedSessionTargets,
	}
	if application.Resolution.HasSyncedState {
		surface.ProfileState = application.Resolution.SyncedState
		surface.HasProfileState = true
		surface.ProfileStatus = SharedSessionBrowserProfileStatusResultFromState(
			application.Resolution.SyncedState,
			selectedInfo,
			result.Profile,
		)
		surface.HasProfileStatus = true
	} else if !sharedSessionBrowserProfileStatusResultEmpty(application.Resolution.ResolvedStatus) {
		surface.ProfileStatus = application.Resolution.ResolvedStatus
		surface.HasProfileStatus = true
	}
	if result.Profiles != nil {
		surface.Profiles = cloneSharedSessionBrowserProjectedProfiles(application.ProjectedProfiles)
		surface.DiscoveredProfiles = append([]string(nil), SharedSessionBrowserDiscoveredProfiles(result.Profiles.Profiles)...)
		surface.DefaultProfile = strings.TrimSpace(result.Profiles.DefaultProfile)
		surface.ApplyProfileInventory = true
	}
	return surface
}

// ProjectSharedSessionBrowserExecutionInventoryProjection narrows the shared
// execution surface onto the profile-status/profile-inventory subset that
// lifecycle tools bridge into top-level payload inventory fields.
func ProjectSharedSessionBrowserExecutionInventoryProjection(
	surface SharedSessionBrowserExecutionSurface,
) *SharedSessionBrowserExecutionInventoryProjection {
	projection := SharedSessionBrowserExecutionInventoryProjection{}
	hasProjection := false
	if surface.HasProfileState {
		projection.ProfileState = surface.ProfileState
		projection.HasProfileState = true
		projection.ProfileStatus = surface.ProfileStatus
		projection.HasProfileStatus = true
		hasProjection = true
	} else if surface.HasProfileStatus {
		projection.ProfileStatus = surface.ProfileStatus
		projection.HasProfileStatus = true
		hasProjection = true
	}
	if surface.ApplyProfileInventory {
		projection.Profiles = cloneSharedSessionBrowserProjectedProfiles(surface.Profiles)
		projection.DiscoveredProfiles = append([]string(nil), surface.DiscoveredProfiles...)
		projection.DefaultProfile = strings.TrimSpace(surface.DefaultProfile)
		projection.ApplyProfileInventory = true
		hasProjection = true
	}
	if !hasProjection {
		return nil
	}
	return &projection
}

// ProjectSharedSessionBrowserTopLevelProfileInventoryFromExecutionInventory
// lowers the execution-inventory subset onto the shared top-level inventory
// contract so tools callers can reuse the same action-surface bridge as other
// lifecycle-owned surfaces.
func ProjectSharedSessionBrowserTopLevelProfileInventoryFromExecutionInventory(
	selectedInfo BrowserRuntimeInfo,
	projection *SharedSessionBrowserExecutionInventoryProjection,
) *SharedSessionBrowserTopLevelProfileInventoryProjection {
	if projection == nil {
		return nil
	}
	req := SharedSessionBrowserTopLevelProfileInventoryProjectionRequest{
		SelectedInfo:          sharedSessionBrowserNormalizedRuntimeInfo(selectedInfo),
		NeedProfileStatus:     projection.HasProfileState || projection.HasProfileStatus,
		NeedProfileInventory:  projection.ApplyProfileInventory,
		Profiles:              cloneSharedSessionBrowserProjectedProfiles(projection.Profiles),
		DiscoveredProfiles:    append([]string(nil), projection.DiscoveredProfiles...),
		DefaultProfile:        strings.TrimSpace(projection.DefaultProfile),
		ApplyProfileInventory: projection.ApplyProfileInventory,
	}
	if projection.HasProfileState {
		req.ProfileStatus = projection.ProfileState
		req.HasProfileStatus = true
	} else if projection.HasProfileStatus {
		req.ProfileStatus = SharedSessionBrowserProfileStateFromStatus(selectedInfo, projection.ProfileStatus)
		req.HasProfileStatus = true
	}
	return ProjectSharedSessionBrowserTopLevelProfileInventoryProjection(req)
}

// BuildSharedSessionBrowserExecutionSurfaceProjection lowers a shared
// lifecycle execution application into the stable execution-surface contract
// plus the inventory subset that tools bridge into payload fields.
func BuildSharedSessionBrowserExecutionSurfaceProjection(
	selectedInfo BrowserRuntimeInfo,
	result SharedSessionBrowserExecutionResult,
	application SharedSessionBrowserExecutionApplication,
) SharedSessionBrowserExecutionSurfaceProjection {
	return BuildSharedSessionBrowserExecutionSurfaceProjectionFromSurface(
		ProjectSharedSessionBrowserExecutionSurface(selectedInfo, result, application),
	)
}

// BuildSharedSessionBrowserExecutionSurfaceProjectionFromSurface narrows the
// shared execution surface into the reusable projection contract consumed by
// lifecycle and prepare-result tool bridges.
func BuildSharedSessionBrowserExecutionSurfaceProjectionFromSurface(
	surface SharedSessionBrowserExecutionSurface,
) SharedSessionBrowserExecutionSurfaceProjection {
	return SharedSessionBrowserExecutionSurfaceProjection{
		Surface:             surface,
		InventoryProjection: ProjectSharedSessionBrowserExecutionInventoryProjection(surface),
	}
}

func cloneSharedSessionBrowserProjectedProfiles(items []SharedSessionBrowserProjectedProfileState) []SharedSessionBrowserProjectedProfileState {
	if len(items) == 0 {
		return nil
	}
	out := make([]SharedSessionBrowserProjectedProfileState, 0, len(items))
	for _, item := range items {
		state := item.State
		state.Backend = strings.TrimSpace(state.Backend)
		state.Profile = strings.TrimSpace(state.Profile)
		state.RuntimeTarget = strings.TrimSpace(state.RuntimeTarget)
		state.BrowserApp = strings.TrimSpace(state.BrowserApp)
		state.Status = strings.TrimSpace(state.Status)
		state.Note = strings.TrimSpace(state.Note)
		out = append(out, SharedSessionBrowserProjectedProfileState{
			State:    state,
			Selected: item.Selected,
		})
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func sharedSessionBrowserExecutionProfilesNote(result SharedSessionBrowserExecutionResult) string {
	if result.Profiles == nil {
		return ""
	}
	return result.Profiles.Note
}
