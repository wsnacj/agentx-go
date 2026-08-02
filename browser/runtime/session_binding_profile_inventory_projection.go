package browserruntime

import "strings"

// SharedSessionBrowserBindingProfileInventoryProjectionRequest carries the
// shared inputs needed to derive top-level profile inventory from a lifecycle-
// owned binding evaluation.
type SharedSessionBrowserBindingProfileInventoryProjectionRequest struct {
	Evaluation              SharedSessionBrowserBindingEvaluation
	SelectedInfo            BrowserRuntimeInfo
	RequestedDefaultProfile string
	NeedProfileStatus       bool
	NeedProfileInventory    bool
}

// ProjectSharedSessionBrowserTopLevelProfileInventoryFromBindingEvaluation
// projects a lifecycle-owned binding evaluation onto the shared top-level
// profile inventory contract so tools callers do not have to rebuild
// profile-status/profile-list/default-profile fallback from binding shells.
func ProjectSharedSessionBrowserTopLevelProfileInventoryFromBindingEvaluation(
	req SharedSessionBrowserBindingProfileInventoryProjectionRequest,
) *SharedSessionBrowserTopLevelProfileInventoryProjection {
	selectedInfo := normalizeSharedSessionBrowserBindingProfileInventorySelectedInfo(
		req.SelectedInfo,
		req.Evaluation,
	)
	sessionProjection := ProjectSharedSessionBrowserTopLevelSessionFromBindingEvaluation(req.Evaluation)
	if len(sessionProjection.Profiles) == 0 {
		return nil
	}
	return ProjectSharedSessionBrowserTopLevelProfileInventoryProjection(
		SharedSessionBrowserTopLevelProfileInventoryProjectionRequest{
			SelectedInfo:            selectedInfo,
			RequestedDefaultProfile: strings.TrimSpace(req.RequestedDefaultProfile),
			NeedProfileStatus:       req.NeedProfileStatus,
			NeedProfileInventory:    req.NeedProfileInventory,
			SessionProjection:       &sessionProjection,
		},
	)
}

func normalizeSharedSessionBrowserBindingProfileInventorySelectedInfo(
	selectedInfo BrowserRuntimeInfo,
	evaluation SharedSessionBrowserBindingEvaluation,
) BrowserRuntimeInfo {
	selectedInfo.Backend = strings.TrimSpace(selectedInfo.Backend)
	selectedInfo.Profile = strings.TrimSpace(selectedInfo.Profile)
	selectedInfo.Target = strings.TrimSpace(selectedInfo.Target)
	fallback := sharedSessionBrowserTopLevelBindingSelectedInfo(evaluation, nil)
	if selectedInfo.Backend == "" {
		selectedInfo.Backend = fallback.Backend
	}
	if selectedInfo.Profile == "" {
		selectedInfo.Profile = fallback.Profile
	}
	if selectedInfo.Target == "" {
		selectedInfo.Target = fallback.Target
	}
	return selectedInfo
}
