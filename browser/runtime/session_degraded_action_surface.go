package browserruntime

import "strings"

// SharedSessionBrowserDegradedActionSurfaceRequest carries the shared degraded
// route/session/profile inputs needed to project the stable action-surface
// contract without rebuilding tool-local mirrors.
type SharedSessionBrowserDegradedActionSurfaceRequest struct {
	SelectedInfo            BrowserRuntimeInfo
	RequestedDefaultProfile string
	ProfileStatus           *SharedSessionBrowserProfileState
	Profiles                []SharedSessionBrowserProjectedProfileState
	SessionProjection       *SharedSessionBrowserTopLevelSessionProjection
	WorkbenchSurface        *SharedSessionBrowserWorkbenchSessionSurfaceProjection
}

// BuildSharedSessionBrowserDegradedActionSurfaceProjection lowers degraded
// route/session/profile snapshots into the shared action-surface contract so
// tools callers only bridge payload-local types into shared projections.
func BuildSharedSessionBrowserDegradedActionSurfaceProjection(
	action string,
	req SharedSessionBrowserDegradedActionSurfaceRequest,
) SharedSessionBrowserActionSurfaceProjection {
	action = strings.ToLower(strings.TrimSpace(action))
	selectedInfo := sharedSessionBrowserNormalizedRuntimeInfo(req.SelectedInfo)
	surface := SharedSessionBrowserActionSurfaceProjection{
		ConfiguredInfo: selectedInfo,
	}

	switch action {
	case "status":
		surface.ProfileInventory = ProjectSharedSessionBrowserTopLevelProfileInventoryProjection(
			SharedSessionBrowserTopLevelProfileInventoryProjectionRequest{
				SelectedInfo:            selectedInfo,
				RequestedDefaultProfile: strings.TrimSpace(req.RequestedDefaultProfile),
				NeedProfileStatus:       true,
				NeedProfileInventory:    true,
				HasProfileStatus:        req.ProfileStatus != nil,
				DefaultProfile:          strings.TrimSpace(req.RequestedDefaultProfile),
				ApplyProfileInventory:   true,
			},
		)
		if req.ProfileStatus != nil {
			surface.ProfileInventory = ProjectSharedSessionBrowserTopLevelProfileInventoryProjection(
				SharedSessionBrowserTopLevelProfileInventoryProjectionRequest{
					SelectedInfo:            selectedInfo,
					RequestedDefaultProfile: strings.TrimSpace(req.RequestedDefaultProfile),
					NeedProfileStatus:       true,
					NeedProfileInventory:    true,
					ProfileStatus:           *req.ProfileStatus,
					HasProfileStatus:        true,
					DefaultProfile:          strings.TrimSpace(req.RequestedDefaultProfile),
					ApplyProfileInventory:   true,
				},
			)
		}
	case "profiles":
		surface.ProfileInventory = ProjectSharedSessionBrowserTopLevelProfileInventoryProjection(
			SharedSessionBrowserTopLevelProfileInventoryProjectionRequest{
				SelectedInfo:            selectedInfo,
				RequestedDefaultProfile: strings.TrimSpace(req.RequestedDefaultProfile),
				NeedProfileInventory:    true,
				Profiles:                cloneSharedSessionBrowserProjectedProfiles(req.Profiles),
				DefaultProfile:          strings.TrimSpace(req.RequestedDefaultProfile),
				ApplyProfileInventory:   true,
			},
		)
	case "sessions":
		surface.SessionProjection = sharedSessionBrowserWorkbenchSessionProjectionApplication(
			req.SessionProjection,
			selectedInfo,
			true,
			false,
		)
	case "workbench":
		surface.ApplyWorkbenchView = true
		surface.WorkbenchSurface = req.WorkbenchSurface
	}

	return surface
}
