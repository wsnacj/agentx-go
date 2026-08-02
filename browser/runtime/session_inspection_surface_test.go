package browserruntime

import (
	"errors"
	"testing"
)

func TestBuildSharedSessionBrowserInspectionSurfaceProjectionBuildsWorkbenchSurface(t *testing.T) {
	surface := BuildSharedSessionBrowserInspectionSurfaceProjection(
		"workbench",
		BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
		SharedSessionBrowserInspectionProjection{
			Note:             "inspection note",
			HasProfileStatus: true,
			ProfileStatus: SharedSessionBrowserProfileState{
				Backend:       "proxy",
				Profile:       "workbench",
				RuntimeTarget: "node",
				Status:        "running",
			},
			HasSessionView: true,
			SessionProjection: SharedSessionBrowserTopLevelSessionProjection{
				Routes: []SharedSessionBrowserRouteSnapshot{{
					Backend:       "proxy",
					Profile:       "workbench",
					RuntimeTarget: "node",
				}},
				TargetCount: 1,
			},
		},
	)

	if !surface.ApplyWorkbenchView || surface.Note != "inspection note" {
		t.Fatalf("expected inspection surface to preserve workbench note/flag, got %#v", surface)
	}
	if surface.ProfileInventory == nil || !surface.ProfileInventory.HasProfileStatus || surface.ProfileInventory.ProfileStatus.Profile != "workbench" {
		t.Fatalf("expected inspection surface to project shared profile inventory, got %#v", surface.ProfileInventory)
	}
	if surface.WorkbenchSurface == nil || surface.WorkbenchSurface.ProfileInventory == nil || surface.WorkbenchSurface.SessionProjection == nil {
		t.Fatalf("expected inspection surface to carry shared workbench surface, got %#v", surface.WorkbenchSurface)
	}
	if surface.SessionProjection == nil ||
		surface.SessionProjection.Projection == nil ||
		len(surface.SessionProjection.Projection.Routes) != 1 ||
		surface.SessionProjection.Projection.TargetCount != 1 {
		t.Fatalf("expected inspection surface to preserve shared session projection, got %#v", surface.SessionProjection)
	}
	if surface.SessionProjection.ConfiguredInfo.Profile != "workbench" ||
		!surface.SessionProjection.ApplyConfiguredProfiles ||
		surface.SessionProjection.MissingSessionIDNote != "browser_runtime: no tool session context is available" {
		t.Fatalf("expected inspection surface to carry shared session apply contract, got %#v", surface.SessionProjection)
	}
}

func TestBuildSharedSessionBrowserInspectionSurfaceProjectionProfilesErrorBuildsTerminalStatus(t *testing.T) {
	surface := BuildSharedSessionBrowserInspectionSurfaceProjection(
		"profiles",
		BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
		SharedSessionBrowserInspectionProjection{
			ProfilesErr: errors.New("profiles unavailable"),
			Note:        "profiles unavailable",
		},
	)

	if surface.Status != "error" || surface.Note != "profiles unavailable" {
		t.Fatalf("expected inspection surface to own profiles terminal status/note, got %#v", surface)
	}
	if surface.ProfileInventory != nil || surface.SessionProjection != nil || surface.WorkbenchSurface != nil {
		t.Fatalf("expected profiles error surface not to project inventory/session shells, got %#v", surface)
	}
}
