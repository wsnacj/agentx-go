package browserruntime

import "strings"

// SharedSessionBrowserActionSurfaceProjection captures the shared
// action-aware top-level/workbench surface contract before tools bridge it
// into payload-specific fields.
type SharedSessionBrowserActionSurfaceProjection struct {
	Status             string
	Note               string
	ConfiguredInfo     BrowserRuntimeInfo
	ProfileInventory   *SharedSessionBrowserTopLevelProfileInventoryProjection
	WorkbenchSurface   *SharedSessionBrowserWorkbenchSessionSurfaceProjection
	SessionProjection  *SharedSessionBrowserWorkbenchSessionProjectionApplication
	ApplyWorkbenchView bool
}

// SharedSessionBrowserInspectionSurfaceProjection is kept as a compatibility
// alias while inspection callers migrate onto the generic action-surface
// contract.
type SharedSessionBrowserInspectionSurfaceProjection = SharedSessionBrowserActionSurfaceProjection

// BuildSharedSessionBrowserInspectionSurfaceProjection lowers an inspection
// projection into the shared action-aware surface contract so tools callers no
// longer need to restitch workbench/status/sessions surface composition
// locally.
func BuildSharedSessionBrowserInspectionSurfaceProjection(
	action string,
	selectedInfo BrowserRuntimeInfo,
	projection SharedSessionBrowserInspectionProjection,
) SharedSessionBrowserActionSurfaceProjection {
	selectedInfo = sharedSessionBrowserNormalizedRuntimeInfo(selectedInfo)
	surface := SharedSessionBrowserActionSurfaceProjection{
		Status:         "",
		Note:           strings.TrimSpace(projection.Note),
		ConfiguredInfo: selectedInfo,
		ProfileInventory: ProjectSharedSessionBrowserTopLevelProfileInventoryFromInspectionProjection(
			selectedInfo,
			projection,
		),
	}
	if strings.EqualFold(strings.TrimSpace(action), "profiles") && projection.ProfilesErr != nil {
		surface.Status = "error"
	}
	if projection.HasSessionView {
		sessionProjection := projection.SessionProjection
		surface.SessionProjection = sharedSessionBrowserWorkbenchSessionProjectionApplication(
			&sessionProjection,
			selectedInfo,
			true,
			true,
		)
	}
	if strings.EqualFold(strings.TrimSpace(action), "workbench") {
		surface.ApplyWorkbenchView = true
		if surface.SessionProjection != nil || surface.ProfileInventory != nil {
			surface.WorkbenchSurface = &SharedSessionBrowserWorkbenchSessionSurfaceProjection{
				SessionProjection: surface.SessionProjection,
				ProfileInventory:  surface.ProfileInventory,
			}
		}
	}
	return surface
}
