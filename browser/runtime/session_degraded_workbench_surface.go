package browserruntime

import "strings"

// SharedSessionBrowserDegradedWorkbenchSurfaceRequest carries the shared
// degraded-route inputs needed to build the lifecycle-owned workbench surface
// contract without restitching binding/session/profile inventory in tools.
type SharedSessionBrowserDegradedWorkbenchSurfaceRequest struct {
	SelectedInfo            BrowserRuntimeInfo
	RequestedDefaultProfile string
	ProfileStatus           *SharedSessionBrowserProfileState
	Profiles                []SharedSessionBrowserProjectedProfileState
	SessionProjection       *SharedSessionBrowserTopLevelSessionProjection
	BindingEvaluation       *SharedSessionBrowserBindingEvaluation
	ProfileSelection        *SharedSessionBrowserProfileSelection
	TargetSelection         *BrowserSessionTargetSelection
}

// BuildSharedSessionBrowserDegradedWorkbenchSurface lowers degraded cached
// route/session/profile inputs into the shared workbench-surface contract so
// tools callers only bridge payload-local types into shared projections.
func BuildSharedSessionBrowserDegradedWorkbenchSurface(
	req SharedSessionBrowserDegradedWorkbenchSurfaceRequest,
) *SharedSessionBrowserWorkbenchSessionSurfaceProjection {
	selectedInfo := sharedSessionBrowserNormalizedRuntimeInfo(req.SelectedInfo)
	sessionProjection := sharedSessionBrowserWorkbenchSessionProjectionApplication(
		req.SessionProjection,
		selectedInfo,
		true,
		false,
	)

	profileInventory := ProjectSharedSessionBrowserTopLevelProfileInventoryProjection(
		SharedSessionBrowserTopLevelProfileInventoryProjectionRequest{
			SelectedInfo:            selectedInfo,
			RequestedDefaultProfile: strings.TrimSpace(req.RequestedDefaultProfile),
			NeedProfileStatus:       true,
			NeedProfileInventory:    true,
			Profiles:                cloneSharedSessionBrowserProjectedProfiles(req.Profiles),
			DefaultProfile:          strings.TrimSpace(req.RequestedDefaultProfile),
			ApplyProfileInventory:   true,
			SessionProjection: func() *SharedSessionBrowserTopLevelSessionProjection {
				if sessionProjection == nil {
					return nil
				}
				return sessionProjection.Projection
			}(),
		},
	)
	if req.ProfileStatus != nil {
		profileInventory = ProjectSharedSessionBrowserTopLevelProfileInventoryProjection(
			SharedSessionBrowserTopLevelProfileInventoryProjectionRequest{
				SelectedInfo:            selectedInfo,
				RequestedDefaultProfile: strings.TrimSpace(req.RequestedDefaultProfile),
				NeedProfileStatus:       true,
				NeedProfileInventory:    true,
				ProfileStatus:           *req.ProfileStatus,
				HasProfileStatus:        true,
				Profiles:                cloneSharedSessionBrowserProjectedProfiles(req.Profiles),
				DefaultProfile:          strings.TrimSpace(req.RequestedDefaultProfile),
				ApplyProfileInventory:   true,
				SessionProjection: func() *SharedSessionBrowserTopLevelSessionProjection {
					if sessionProjection == nil {
						return nil
					}
					return sessionProjection.Projection
				}(),
			},
		)
	}

	var bindingProjection *SharedSessionBrowserTopLevelBindingProjection
	if req.BindingEvaluation != nil {
		evaluation := *req.BindingEvaluation
		bindingProjection = &SharedSessionBrowserTopLevelBindingProjection{
			Evaluation: evaluation,
		}
		if req.ProfileSelection != nil {
			cloned := *req.ProfileSelection
			bindingProjection.ProfileSelection = &cloned
		}
		if req.TargetSelection != nil {
			cloned := *req.TargetSelection
			bindingProjection.TargetSelection = &cloned
		}
		if bindingProjection.ProfileSelection != nil || bindingProjection.TargetSelection != nil {
			bindingProjection.Evaluation = ProjectSharedSessionBrowserBindingEvaluationWithSelectionProjection(
				bindingProjection.Evaluation,
				&SharedSessionBrowserSelectionProjection{
					ProfileSelection: bindingProjection.ProfileSelection,
					TargetSelection:  bindingProjection.TargetSelection,
				},
			)
		}
	}

	if sessionProjection == nil && profileInventory == nil && bindingProjection == nil {
		return nil
	}
	return &SharedSessionBrowserWorkbenchSessionSurfaceProjection{
		BindingProjection: bindingProjection,
		SessionProjection: sessionProjection,
		ProfileInventory:  profileInventory,
	}
}
