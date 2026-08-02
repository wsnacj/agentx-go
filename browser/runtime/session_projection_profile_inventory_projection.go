package browserruntime

// SharedSessionBrowserSessionProjectionProfileInventoryRequest carries the
// shared inputs needed to derive top-level profile inventory from a lifecycle-
// owned session projection without reconstructing request semantics in tools.
type SharedSessionBrowserSessionProjectionProfileInventoryRequest struct {
	SelectedInfo            BrowserRuntimeInfo
	RequestedDefaultProfile string
	NeedProfileStatus       bool
	NeedProfileInventory    bool
	DiscoveredProfiles      []string
	SessionProjection       *SharedSessionBrowserTopLevelSessionProjection
}

// ProjectSharedSessionBrowserTopLevelProfileInventoryFromSessionProjection
// projects a lifecycle-owned session projection onto the shared top-level
// profile inventory contract so tools callers only bridge the result.
func ProjectSharedSessionBrowserTopLevelProfileInventoryFromSessionProjection(
	req SharedSessionBrowserSessionProjectionProfileInventoryRequest,
) *SharedSessionBrowserTopLevelProfileInventoryProjection {
	return ProjectSharedSessionBrowserTopLevelProfileInventoryProjection(
		SharedSessionBrowserTopLevelProfileInventoryProjectionRequest{
			SelectedInfo:            sharedSessionBrowserNormalizedRuntimeInfo(req.SelectedInfo),
			RequestedDefaultProfile: req.RequestedDefaultProfile,
			NeedProfileStatus:       req.NeedProfileStatus,
			NeedProfileInventory:    req.NeedProfileInventory,
			DiscoveredProfiles:      append([]string(nil), req.DiscoveredProfiles...),
			SessionProjection:       req.SessionProjection,
		},
	)
}

// ProjectSharedSessionBrowserTopLevelProfileInventoryFromInspectionProjection
// lowers a lifecycle-owned inspection projection onto the shared top-level
// profile inventory contract so tools callers do not need to restitch
// inspection status/profile/session fallback locally.
func ProjectSharedSessionBrowserTopLevelProfileInventoryFromInspectionProjection(
	selectedInfo BrowserRuntimeInfo,
	projection SharedSessionBrowserInspectionProjection,
) *SharedSessionBrowserTopLevelProfileInventoryProjection {
	var sessionProjection *SharedSessionBrowserTopLevelSessionProjection
	if projection.HasSessionView {
		cloned := projection.SessionProjection
		sessionProjection = &cloned
	}
	return ProjectSharedSessionBrowserTopLevelProfileInventoryProjection(
		SharedSessionBrowserTopLevelProfileInventoryProjectionRequest{
			SelectedInfo:          sharedSessionBrowserNormalizedRuntimeInfo(selectedInfo),
			NeedProfileStatus:     true,
			NeedProfileInventory:  true,
			ProfileStatus:         projection.ProfileStatus,
			HasProfileStatus:      projection.HasProfileStatus,
			Profiles:              cloneSharedSessionBrowserProjectedProfiles(projection.Profiles),
			DiscoveredProfiles:    append([]string(nil), projection.DiscoveredProfiles...),
			DefaultProfile:        projection.DefaultProfile,
			ApplyProfileInventory: projection.HasProfiles,
			SessionProjection:     sessionProjection,
		},
	)
}
