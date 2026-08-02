package browserruntime

import "strings"

// SharedSessionBrowserTopLevelProfileInventoryProjection captures the shared
// profile-status/profile-inventory shell that tools bridge into payload-level
// profile inventory fields.
type SharedSessionBrowserTopLevelProfileInventoryProjection struct {
	ProfileStatus         SharedSessionBrowserProfileState
	HasProfileStatus      bool
	Profiles              []SharedSessionBrowserProjectedProfileState
	DiscoveredProfiles    []string
	DefaultProfile        string
	ApplyProfileInventory bool
}

// SharedSessionBrowserTopLevelProfileInventoryProjectionRequest carries the
// shared inputs needed to build the stable top-level profile inventory shell,
// including session-projection fallback when no explicit profile inventory is
// present at the call site.
type SharedSessionBrowserTopLevelProfileInventoryProjectionRequest struct {
	SelectedInfo            BrowserRuntimeInfo
	RequestedDefaultProfile string
	NeedProfileStatus       bool
	NeedProfileInventory    bool
	ProfileStatus           SharedSessionBrowserProfileState
	HasProfileStatus        bool
	Profiles                []SharedSessionBrowserProjectedProfileState
	DiscoveredProfiles      []string
	DefaultProfile          string
	ApplyProfileInventory   bool
	SessionProjection       *SharedSessionBrowserTopLevelSessionProjection
}

// ProjectSharedSessionBrowserTopLevelProfileInventoryProjection lowers shared
// status/profile/session inputs into a single lifecycle-owned profile
// inventory projection so tools callers do not have to re-derive fallback
// status or profile lists from session projection locally.
func ProjectSharedSessionBrowserTopLevelProfileInventoryProjection(
	req SharedSessionBrowserTopLevelProfileInventoryProjectionRequest,
) *SharedSessionBrowserTopLevelProfileInventoryProjection {
	projection := SharedSessionBrowserTopLevelProfileInventoryProjection{}
	hasProjection := false

	explicitDefault := strings.TrimSpace(req.DefaultProfile)
	if req.NeedProfileStatus && req.HasProfileStatus {
		projection.ProfileStatus = req.ProfileStatus
		projection.HasProfileStatus = true
		hasProjection = true
	}
	if req.NeedProfileInventory && req.ApplyProfileInventory {
		projection.Profiles = cloneSharedSessionBrowserProjectedProfiles(req.Profiles)
		projection.DiscoveredProfiles = append([]string(nil), req.DiscoveredProfiles...)
		projection.DefaultProfile = explicitDefault
		projection.ApplyProfileInventory = true
		hasProjection = true
	}

	needFallbackStatus := req.NeedProfileStatus && !projection.HasProfileStatus
	needFallbackInventory := req.NeedProfileInventory && !projection.ApplyProfileInventory
	if (needFallbackStatus || needFallbackInventory) && req.SessionProjection != nil && len(req.SessionProjection.Profiles) > 0 {
		fallbackProfiles := cloneSharedSessionBrowserProjectedProfiles(req.SessionProjection.Profiles)
		fallbackDefault := sharedSessionBrowserProfileInventoryDefaultProfile(
			fallbackProfiles,
			explicitDefault,
			req.RequestedDefaultProfile,
			req.SelectedInfo.Profile,
		)
		if needFallbackInventory {
			projection.Profiles = fallbackProfiles
			projection.DiscoveredProfiles = append([]string(nil), req.DiscoveredProfiles...)
			projection.DefaultProfile = fallbackDefault
			projection.ApplyProfileInventory = true
			hasProjection = true
		}
		if needFallbackStatus {
			if state, ok := sharedSessionBrowserPreferredProjectedProfileState(fallbackProfiles, fallbackDefault); ok {
				projection.ProfileStatus = state
				projection.HasProfileStatus = true
				hasProjection = true
			}
		}
	}

	if !hasProjection {
		return nil
	}
	return &projection
}

func sharedSessionBrowserProfileInventoryDefaultProfile(
	profiles []SharedSessionBrowserProjectedProfileState,
	values ...string,
) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	if len(profiles) == 0 {
		return ""
	}
	states := make([]SharedSessionBrowserProfileState, 0, len(profiles))
	for _, item := range profiles {
		states = append(states, item.State)
	}
	activeProfile, _ := SummarizeSharedSessionBrowserProfiles(states)
	return strings.TrimSpace(activeProfile)
}

func sharedSessionBrowserPreferredProjectedProfileState(
	profiles []SharedSessionBrowserProjectedProfileState,
	preferredProfile string,
) (SharedSessionBrowserProfileState, bool) {
	preferredProfile = strings.TrimSpace(preferredProfile)
	if preferredProfile != "" {
		for _, item := range profiles {
			if strings.EqualFold(strings.TrimSpace(item.State.Profile), preferredProfile) {
				return item.State, true
			}
		}
	}
	if len(profiles) == 1 {
		return profiles[0].State, true
	}
	if len(profiles) == 0 {
		return SharedSessionBrowserProfileState{}, false
	}
	return SharedSessionBrowserProfileState{}, false
}
