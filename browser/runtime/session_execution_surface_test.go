package browserruntime

import "testing"

func TestProjectSharedSessionBrowserExecutionSurfacePrefersSyncedState(t *testing.T) {
	surface := ProjectSharedSessionBrowserExecutionSurface(
		BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
		SharedSessionBrowserExecutionResult{
			Profile: "workbench",
			Profiles: &BrowserProfilesResult{
				DefaultProfile: "workbench",
				Profiles: []BrowserProfileInfo{
					{Profile: "workbench", Status: "running"},
					{Profile: "relay", Status: "stopped"},
				},
				Note: "profiles synced",
			},
		},
		SharedSessionBrowserExecutionApplication{
			Resolution: SharedSessionBrowserExecutionResolution{
				SyncedState: SharedSessionBrowserProfileState{
					Backend:       "proxy",
					Profile:       "workbench",
					RuntimeTarget: "node",
					Status:        "running",
					Running:       true,
					Connected:     true,
					Note:          "synced lifecycle",
				},
				HasSyncedState: true,
			},
			Cleanup: SharedSessionBrowserExecutionCleanup{
				ClearedSessionTargets: 2,
			},
			ProjectedProfiles: []SharedSessionBrowserProjectedProfileState{
				{State: SharedSessionBrowserProfileState{Backend: "proxy", Profile: "workbench", RuntimeTarget: "node", Status: "running"}, Selected: true},
				{State: SharedSessionBrowserProfileState{Backend: "proxy", Profile: "relay", RuntimeTarget: "node", Status: "stopped"}},
			},
		},
	)

	if surface.Note != "profiles synced" || surface.ClearedSessionTargets != 2 {
		t.Fatalf("expected execution surface to preserve note/cleanup, got %#v", surface)
	}
	if !surface.HasProfileState || surface.ProfileState.Profile != "workbench" || surface.ProfileState.RuntimeTarget != "node" || surface.ProfileState.Status != "running" {
		t.Fatalf("expected execution surface to preserve synced profile state, got %#v", surface.ProfileState)
	}
	if !surface.HasProfileStatus || surface.ProfileStatus.Profile != "workbench" || surface.ProfileStatus.Status != "running" || !surface.ProfileStatus.Connected {
		t.Fatalf("expected execution surface to derive status from synced state, got %#v", surface.ProfileStatus)
	}
	if !surface.ApplyProfileInventory || len(surface.Profiles) != 2 || len(surface.DiscoveredProfiles) != 2 || surface.DefaultProfile != "workbench" {
		t.Fatalf("expected execution surface to preserve profile inventory, got %#v", surface)
	}
}

func TestProjectSharedSessionBrowserExecutionSurfaceFallsBackToResolvedStatus(t *testing.T) {
	surface := ProjectSharedSessionBrowserExecutionSurface(
		BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
		SharedSessionBrowserExecutionResult{},
		SharedSessionBrowserExecutionApplication{
			Resolution: SharedSessionBrowserExecutionResolution{
				ResolvedStatus: BrowserProfileStatusResult{
					Backend:    "proxy",
					Profile:    "isolated",
					BrowserApp: "Chromium",
					Status:     "connected",
					Running:    true,
					Connected:  true,
				},
			},
		},
	)

	if surface.HasProfileState {
		t.Fatalf("expected execution surface without synced state not to expose profile state, got %#v", surface.ProfileState)
	}
	if !surface.HasProfileStatus || surface.ProfileStatus.Profile != "isolated" || surface.ProfileStatus.Status != "connected" || !surface.ProfileStatus.Running {
		t.Fatalf("expected execution surface to preserve resolved status fallback, got %#v", surface.ProfileStatus)
	}
	if surface.ApplyProfileInventory || len(surface.Profiles) != 0 {
		t.Fatalf("expected execution surface without profiles inventory to keep inventory disabled, got %#v", surface)
	}
}

func TestProjectSharedSessionBrowserExecutionSurfaceClonesProjectedProfiles(t *testing.T) {
	projected := []SharedSessionBrowserProjectedProfileState{
		{State: SharedSessionBrowserProfileState{Backend: " proxy ", Profile: " workbench ", RuntimeTarget: " node ", Status: " running ", Note: " note "}, Selected: true},
	}
	surface := ProjectSharedSessionBrowserExecutionSurface(
		BrowserRuntimeInfo{},
		SharedSessionBrowserExecutionResult{
			Profiles: &BrowserProfilesResult{Profiles: []BrowserProfileInfo{{Profile: "workbench"}}},
		},
		SharedSessionBrowserExecutionApplication{
			ProjectedProfiles: projected,
		},
	)
	projected[0].State.Profile = "mutated"

	if len(surface.Profiles) != 1 || surface.Profiles[0].State.Profile != "workbench" || surface.Profiles[0].State.Backend != "proxy" || surface.Profiles[0].State.Note != "note" {
		t.Fatalf("expected execution surface to clone and trim projected profiles, got %#v", surface.Profiles)
	}
}

func TestProjectSharedSessionBrowserExecutionInventoryProjectionPrefersSyncedStateAndInventory(t *testing.T) {
	projection := ProjectSharedSessionBrowserExecutionInventoryProjection(
		SharedSessionBrowserExecutionSurface{
			ProfileState: SharedSessionBrowserProfileState{
				Backend:       "proxy",
				Profile:       "workbench",
				RuntimeTarget: "node",
				Status:        "running",
				Running:       true,
				Connected:     true,
			},
			HasProfileState: true,
			ProfileStatus: BrowserProfileStatusResult{
				Backend:   "proxy",
				Profile:   "workbench",
				Status:    "running",
				Running:   true,
				Connected: true,
			},
			HasProfileStatus: true,
			Profiles: []SharedSessionBrowserProjectedProfileState{
				{State: SharedSessionBrowserProfileState{Backend: " proxy ", Profile: " workbench ", RuntimeTarget: " node ", Status: " running "}, Selected: true},
			},
			DiscoveredProfiles:    []string{"workbench"},
			DefaultProfile:        " workbench ",
			ApplyProfileInventory: true,
		},
	)

	if projection == nil {
		t.Fatalf("expected execution inventory projection to be built")
	}
	if !projection.HasProfileState || projection.ProfileState.Profile != "workbench" || !projection.HasProfileStatus || projection.ProfileStatus.Profile != "workbench" {
		t.Fatalf("expected execution inventory projection to preserve lifecycle profile status/state, got %#v", projection)
	}
	if !projection.ApplyProfileInventory || len(projection.Profiles) != 1 || projection.Profiles[0].State.Profile != "workbench" || projection.DefaultProfile != "workbench" {
		t.Fatalf("expected execution inventory projection to clone profile inventory, got %#v", projection)
	}
}

func TestProjectSharedSessionBrowserExecutionInventoryProjectionReturnsNilWhenEmpty(t *testing.T) {
	if projection := ProjectSharedSessionBrowserExecutionInventoryProjection(SharedSessionBrowserExecutionSurface{}); projection != nil {
		t.Fatalf("expected empty execution surface not to synthesize inventory projection, got %#v", projection)
	}
}

func TestBuildSharedSessionBrowserExecutionSurfaceProjectionCarriesSurfaceAndInventory(t *testing.T) {
	projection := BuildSharedSessionBrowserExecutionSurfaceProjection(
		BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
		SharedSessionBrowserExecutionResult{
			Profile: "workbench",
			Profiles: &BrowserProfilesResult{
				DefaultProfile: "workbench",
				Profiles: []BrowserProfileInfo{
					{Profile: "workbench", Status: "running"},
				},
				Note: "profiles synced",
			},
		},
		SharedSessionBrowserExecutionApplication{
			Resolution: SharedSessionBrowserExecutionResolution{
				SyncedState: SharedSessionBrowserProfileState{
					Backend:       "proxy",
					Profile:       "workbench",
					RuntimeTarget: "node",
					Status:        "running",
					Running:       true,
					Connected:     true,
				},
				HasSyncedState: true,
			},
			Cleanup: SharedSessionBrowserExecutionCleanup{
				ClearedSessionTargets: 2,
			},
			ProjectedProfiles: []SharedSessionBrowserProjectedProfileState{
				{State: SharedSessionBrowserProfileState{Backend: "proxy", Profile: "workbench", RuntimeTarget: "node", Status: "running"}, Selected: true},
			},
		},
	)

	if projection.Surface.Note != "profiles synced" || projection.Surface.ClearedSessionTargets != 2 {
		t.Fatalf("expected execution projection to preserve surface note/cleanup, got %#v", projection)
	}
	if projection.InventoryProjection == nil ||
		!projection.InventoryProjection.HasProfileState ||
		projection.InventoryProjection.ProfileState.Profile != "workbench" ||
		!projection.InventoryProjection.ApplyProfileInventory ||
		projection.InventoryProjection.DefaultProfile != "workbench" {
		t.Fatalf("expected execution projection to carry inventory subset, got %#v", projection.InventoryProjection)
	}
}

func TestProjectSharedSessionBrowserTopLevelProfileInventoryFromExecutionInventory(t *testing.T) {
	projection := ProjectSharedSessionBrowserTopLevelProfileInventoryFromExecutionInventory(
		BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
		&SharedSessionBrowserExecutionInventoryProjection{
			ProfileState: SharedSessionBrowserProfileState{
				Backend:       "proxy",
				Profile:       "workbench",
				RuntimeTarget: "node",
				Status:        "running",
				Running:       true,
				Connected:     true,
			},
			HasProfileState: true,
			Profiles: []SharedSessionBrowserProjectedProfileState{
				{State: SharedSessionBrowserProfileState{Backend: "proxy", Profile: "workbench", RuntimeTarget: "node", Status: "running"}, Selected: true},
				{State: SharedSessionBrowserProfileState{Backend: "proxy", Profile: "relay", RuntimeTarget: "node", Status: "stopped"}},
			},
			DiscoveredProfiles:    []string{"workbench", "relay"},
			DefaultProfile:        "workbench",
			ApplyProfileInventory: true,
		},
	)

	if projection == nil {
		t.Fatalf("expected top-level profile inventory projection from execution inventory")
	}
	if !projection.HasProfileStatus || projection.ProfileStatus.Profile != "workbench" || projection.ProfileStatus.Status != "running" {
		t.Fatalf("expected execution inventory helper to preserve profile status, got %#v", projection)
	}
	if !projection.ApplyProfileInventory || len(projection.Profiles) != 2 || projection.DefaultProfile != "workbench" {
		t.Fatalf("expected execution inventory helper to preserve profile inventory, got %#v", projection)
	}
}
