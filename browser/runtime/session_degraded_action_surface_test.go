package browserruntime

import "testing"

func TestBuildSharedSessionBrowserDegradedActionSurfaceProjectionBuildsSessionsSurface(t *testing.T) {
	surface := BuildSharedSessionBrowserDegradedActionSurfaceProjection(
		"sessions",
		SharedSessionBrowserDegradedActionSurfaceRequest{
			SelectedInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
			SessionProjection: &SharedSessionBrowserTopLevelSessionProjection{
				Routes: []SharedSessionBrowserRouteSnapshot{{
					Backend:       "proxy",
					Profile:       "workbench",
					RuntimeTarget: "node",
				}},
				TargetCount: 2,
				Runs: []SharedSessionRunInfo{{
					RunID:  "run-1",
					Status: "running",
				}},
				Profiles: []SharedSessionBrowserProjectedProfileState{{
					State: SharedSessionBrowserProfileState{
						Backend:       "proxy",
						Profile:       "workbench",
						RuntimeTarget: "node",
						Status:        "running",
					},
				}},
			},
		},
	)

	if surface.ConfiguredInfo != (BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}) {
		t.Fatalf("expected degraded action surface to keep configured info, got %#v", surface.ConfiguredInfo)
	}
	if surface.SessionProjection == nil || surface.SessionProjection.Projection == nil || len(surface.SessionProjection.Projection.Routes) != 1 {
		t.Fatalf("expected degraded action surface to keep shared session projection, got %#v", surface.SessionProjection)
	}
	if !surface.SessionProjection.ApplyConfiguredProfiles || surface.SessionProjection.MissingSessionIDNote != "" {
		t.Fatalf("expected degraded action surface to keep session apply contract, got %#v", surface.SessionProjection)
	}
}

func TestBuildSharedSessionBrowserDegradedActionSurfaceProjectionBuildsStatusInventory(t *testing.T) {
	surface := BuildSharedSessionBrowserDegradedActionSurfaceProjection(
		"status",
		SharedSessionBrowserDegradedActionSurfaceRequest{
			SelectedInfo:            BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
			RequestedDefaultProfile: "workbench",
			ProfileStatus: &SharedSessionBrowserProfileState{
				Backend:       "proxy",
				Profile:       "workbench",
				RuntimeTarget: "node",
				Status:        "running",
				Running:       true,
				Connected:     true,
			},
		},
	)

	if surface.ProfileInventory == nil ||
		!surface.ProfileInventory.HasProfileStatus ||
		surface.ProfileInventory.ProfileStatus.Profile != "workbench" ||
		!surface.ProfileInventory.ApplyProfileInventory ||
		surface.ProfileInventory.DefaultProfile != "workbench" {
		t.Fatalf("expected degraded status surface to keep shared profile inventory, got %#v", surface.ProfileInventory)
	}
}
