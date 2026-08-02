package browserruntime

import "testing"

func TestProjectSharedSessionBrowserTopLevelProfileInventoryProjectionPrefersExplicitProjection(t *testing.T) {
	projection := ProjectSharedSessionBrowserTopLevelProfileInventoryProjection(
		SharedSessionBrowserTopLevelProfileInventoryProjectionRequest{
			SelectedInfo:          BrowserRuntimeInfo{Backend: "proxy", Profile: "requested", Target: "node"},
			NeedProfileStatus:     true,
			NeedProfileInventory:  true,
			HasProfileStatus:      true,
			ProfileStatus:         SharedSessionBrowserProfileState{Backend: "proxy", Profile: "workbench", RuntimeTarget: "node", Status: "running", Running: true, Connected: true},
			Profiles:              []SharedSessionBrowserProjectedProfileState{{State: SharedSessionBrowserProfileState{Backend: "proxy", Profile: "workbench", RuntimeTarget: "node", Status: "running"}, Selected: true}},
			DiscoveredProfiles:    []string{"workbench"},
			DefaultProfile:        " workbench ",
			ApplyProfileInventory: true,
			SessionProjection: &SharedSessionBrowserTopLevelSessionProjection{
				Profiles: []SharedSessionBrowserProjectedProfileState{{State: SharedSessionBrowserProfileState{Backend: "proxy", Profile: "fallback", RuntimeTarget: "node", Status: "stopped"}}},
			},
		},
	)

	if projection == nil {
		t.Fatalf("expected explicit profile inventory projection")
	}
	if !projection.HasProfileStatus || projection.ProfileStatus.Profile != "workbench" || projection.ProfileStatus.Status != "running" {
		t.Fatalf("expected explicit profile status to win, got %#v", projection)
	}
	if !projection.ApplyProfileInventory || len(projection.Profiles) != 1 || projection.Profiles[0].State.Profile != "workbench" {
		t.Fatalf("expected explicit profile inventory to win, got %#v", projection)
	}
	if projection.DefaultProfile != "workbench" || len(projection.DiscoveredProfiles) != 1 || projection.DiscoveredProfiles[0] != "workbench" {
		t.Fatalf("expected explicit inventory metadata to be preserved, got %#v", projection)
	}
}

func TestProjectSharedSessionBrowserTopLevelProfileInventoryProjectionFallsBackToSessionProjection(t *testing.T) {
	projection := ProjectSharedSessionBrowserTopLevelProfileInventoryProjection(
		SharedSessionBrowserTopLevelProfileInventoryProjectionRequest{
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
							BrowserApp:    "Chromium",
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
		t.Fatalf("expected fallback profile inventory projection")
	}
	if !projection.HasProfileStatus || projection.ProfileStatus.Profile != "workbench" || projection.ProfileStatus.RuntimeTarget != "node" {
		t.Fatalf("expected fallback status to prefer requested/default profile, got %#v", projection)
	}
	if !projection.ApplyProfileInventory || len(projection.Profiles) != 2 || projection.Profiles[0].State.Profile != "workbench" {
		t.Fatalf("expected fallback inventory to retain session projection profiles, got %#v", projection)
	}
	if projection.DefaultProfile != "workbench" || len(projection.DiscoveredProfiles) != 2 {
		t.Fatalf("expected fallback inventory metadata to be preserved, got %#v", projection)
	}
}

func TestProjectSharedSessionBrowserTopLevelProfileInventoryProjectionReturnsNilWhenEmpty(t *testing.T) {
	if projection := ProjectSharedSessionBrowserTopLevelProfileInventoryProjection(
		SharedSessionBrowserTopLevelProfileInventoryProjectionRequest{},
	); projection != nil {
		t.Fatalf("expected empty profile inventory request to return nil, got %#v", projection)
	}
}
