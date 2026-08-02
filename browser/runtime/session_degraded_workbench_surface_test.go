package browserruntime

import "testing"

func TestBuildSharedSessionBrowserDegradedWorkbenchSurfaceBuildsSharedSurface(t *testing.T) {
	surface := BuildSharedSessionBrowserDegradedWorkbenchSurface(
		SharedSessionBrowserDegradedWorkbenchSurfaceRequest{
			SelectedInfo:            BrowserRuntimeInfo{Backend: "custom-playwright", Profile: "workbench", Target: "host"},
			RequestedDefaultProfile: "workbench",
			ProfileStatus: &SharedSessionBrowserProfileState{
				Backend:       "custom-playwright",
				Profile:       "workbench",
				RuntimeTarget: "host",
				Status:        "running",
				Running:       true,
				Connected:     true,
				Note:          "cached route-scoped session snapshot",
			},
			Profiles: []SharedSessionBrowserProjectedProfileState{{
				State: SharedSessionBrowserProfileState{
					Backend:       "custom-playwright",
					Profile:       "workbench",
					RuntimeTarget: "host",
					BrowserApp:    "Chrome",
					Note:          "cached route-scoped session snapshot",
				},
				Selected: true,
			}},
			SessionProjection: &SharedSessionBrowserTopLevelSessionProjection{
				Routes: []SharedSessionBrowserRouteSnapshot{{
					Backend:             "custom-playwright",
					Profile:             "workbench",
					RuntimeTarget:       "host",
					BrowserApp:          "Chrome",
					CurrentTargetID:     "host-current",
					CurrentTargetSource: "tracked_active_tab",
				}},
				TargetCount: 1,
				Profiles: []SharedSessionBrowserProjectedProfileState{{
					State: SharedSessionBrowserProfileState{
						Backend:       "custom-playwright",
						Profile:       "workbench",
						RuntimeTarget: "host",
						BrowserApp:    "Chrome",
						Note:          "cached route-scoped session snapshot",
					},
					Selected: true,
				}},
			},
			BindingEvaluation: &SharedSessionBrowserBindingEvaluation{
				Routes: []SharedSessionBrowserRouteSnapshot{{
					Backend:             "custom-playwright",
					Profile:             "workbench",
					RuntimeTarget:       "host",
					BrowserApp:          "Chrome",
					CurrentTargetID:     "host-current",
					CurrentTargetSource: "tracked_active_tab",
				}},
			},
			TargetSelection: &BrowserSessionTargetSelection{
				ID:            "host-current",
				Backend:       "custom-playwright",
				Profile:       "workbench",
				RuntimeTarget: "host",
				BrowserApp:    "Chrome",
				Source:        "tracked_active_tab",
			},
		},
	)

	if surface == nil {
		t.Fatalf("expected degraded workbench surface")
	}
	if surface.SessionProjection == nil || surface.SessionProjection.Projection == nil || len(surface.SessionProjection.Projection.Routes) != 1 {
		t.Fatalf("expected degraded workbench surface to keep shared session projection, got %#v", surface.SessionProjection)
	}
	if surface.SessionProjection.ConfiguredInfo != (BrowserRuntimeInfo{Backend: "custom-playwright", Profile: "workbench", Target: "host"}) ||
		!surface.SessionProjection.ApplyConfiguredProfiles {
		t.Fatalf("expected degraded workbench surface to carry configured info/apply flags, got %#v", surface.SessionProjection)
	}
	if surface.ProfileInventory == nil ||
		!surface.ProfileInventory.HasProfileStatus ||
		surface.ProfileInventory.ProfileStatus.Profile != "workbench" ||
		!surface.ProfileInventory.ApplyProfileInventory ||
		surface.ProfileInventory.DefaultProfile != "workbench" ||
		len(surface.ProfileInventory.Profiles) != 1 {
		t.Fatalf("expected degraded workbench surface to keep shared profile inventory, got %#v", surface.ProfileInventory)
	}
	if surface.BindingProjection == nil || surface.BindingProjection.TargetSelection == nil || surface.BindingProjection.TargetSelection.ID != "host-current" {
		t.Fatalf("expected degraded workbench surface to keep shared binding projection, got %#v", surface.BindingProjection)
	}
}
