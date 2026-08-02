package browserruntime

// ProjectSharedSessionBrowserTopLevelProfileInventoryFromClearResult projects a
// lifecycle-owned clear result onto the shared top-level profile inventory
// contract so tools callers do not have to rebuild clear-result
// profile-state/profile-status fallback locally.
func ProjectSharedSessionBrowserTopLevelProfileInventoryFromClearResult(
	selectedInfo BrowserRuntimeInfo,
	result SharedSessionBrowserClearResult,
) *SharedSessionBrowserTopLevelProfileInventoryProjection {
	if result.HasProfileState {
		return ProjectSharedSessionBrowserTopLevelProfileInventoryProjection(
			SharedSessionBrowserTopLevelProfileInventoryProjectionRequest{
				SelectedInfo:      selectedInfo,
				NeedProfileStatus: true,
				ProfileStatus:     result.ProfileState,
				HasProfileStatus:  true,
			},
		)
	}
	if sharedSessionBrowserProfileStatusResultEmpty(result.ProfileStatus) {
		return nil
	}
	return ProjectSharedSessionBrowserTopLevelProfileInventoryProjection(
		SharedSessionBrowserTopLevelProfileInventoryProjectionRequest{
			SelectedInfo:      selectedInfo,
			NeedProfileStatus: true,
			ProfileStatus:     SharedSessionBrowserProfileStateFromStatus(selectedInfo, result.ProfileStatus),
			HasProfileStatus:  true,
		},
	)
}
