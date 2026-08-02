package browserruntime

import "testing"

func TestProjectSharedSessionBrowserTopLevelProfileInventoryFromSessionProjectionFallsBackToSessionView(t *testing.T) {
	projection := ProjectSharedSessionBrowserTopLevelProfileInventoryFromSessionProjection(
		SharedSessionBrowserSessionProjectionProfileInventoryRequest{
			SelectedInfo:         BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
			NeedProfileStatus:    true,
			NeedProfileInventory: true,
			DiscoveredProfiles:   []string{"workbench", "relay"},
			SessionProjection: &SharedSessionBrowserTopLevelSessionProjection{
				Profiles: []SharedSessionBrowserProjectedProfileState{
					{
						State: SharedSessionBrowserProfileState{
							Backend:       "proxy",
							Profile:       "workbench",
							RuntimeTarget: "node",
							Status:        "running",
							Running:       true,
							Connected:     true,
						},
						Selected: true,
					},
					{
						State: SharedSessionBrowserProfileState{
							Backend:       "proxy",
							Profile:       "relay",
							RuntimeTarget: "node",
							Status:        "stopped",
						},
					},
				},
			},
		},
	)

	if projection == nil {
		t.Fatalf("expected session projection fallback")
	}
	if !projection.HasProfileStatus || projection.ProfileStatus.Profile != "workbench" {
		t.Fatalf("expected session projection fallback status, got %#v", projection)
	}
	if !projection.ApplyProfileInventory || len(projection.Profiles) != 2 {
		t.Fatalf("expected session projection fallback inventory, got %#v", projection)
	}
	if projection.DefaultProfile != "workbench" || len(projection.DiscoveredProfiles) != 2 {
		t.Fatalf("expected session projection fallback metadata, got %#v", projection)
	}
}

func TestProjectSharedSessionBrowserTopLevelProfileInventoryFromInspectionProjectionPrefersExplicitStatusAndProfiles(t *testing.T) {
	projection := ProjectSharedSessionBrowserTopLevelProfileInventoryFromInspectionProjection(
		BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
		SharedSessionBrowserInspectionProjection{
			ProfileStatus: SharedSessionBrowserProfileState{
				Backend:       "proxy",
				Profile:       "workbench",
				RuntimeTarget: "node",
				Status:        "running",
				Running:       true,
				Connected:     true,
			},
			HasProfileStatus: true,
			Profiles: []SharedSessionBrowserProjectedProfileState{
				{
					State: SharedSessionBrowserProfileState{
						Backend:       "proxy",
						Profile:       "workbench",
						RuntimeTarget: "node",
						Status:        "running",
					},
					Selected: true,
				},
			},
			HasProfiles:        true,
			DiscoveredProfiles: []string{"workbench"},
			DefaultProfile:     " workbench ",
			HasSessionView:     true,
			SessionProjection: SharedSessionBrowserTopLevelSessionProjection{
				Profiles: []SharedSessionBrowserProjectedProfileState{
					{State: SharedSessionBrowserProfileState{Backend: "proxy", Profile: "relay", RuntimeTarget: "node", Status: "stopped"}},
				},
			},
		},
	)

	if projection == nil {
		t.Fatalf("expected inspection projection")
	}
	if !projection.HasProfileStatus || projection.ProfileStatus.Profile != "workbench" || projection.ProfileStatus.Status != "running" {
		t.Fatalf("expected explicit inspection status to win, got %#v", projection)
	}
	if !projection.ApplyProfileInventory || len(projection.Profiles) != 1 || projection.Profiles[0].State.Profile != "workbench" {
		t.Fatalf("expected explicit inspection profiles to win, got %#v", projection)
	}
	if projection.DefaultProfile != "workbench" || len(projection.DiscoveredProfiles) != 1 {
		t.Fatalf("expected explicit inspection metadata to be preserved, got %#v", projection)
	}
}
