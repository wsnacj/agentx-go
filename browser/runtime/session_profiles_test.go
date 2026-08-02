package browserruntime

import (
	"context"
	"testing"
	"time"
)

func TestSnapshotSharedSessionBrowserProjectedProfilesForScopeMarksSelectionAndFillsBrowserApp(t *testing.T) {
	registry := NewBrowserSessionStateRegistry()
	registry.SyncSessionBrowserProfiles("s1", SharedSessionBrowserProfileState{}, []SharedSessionBrowserProfileState{{
		Backend:       "proxy-tabs",
		Profile:       "work",
		RuntimeTarget: "node",
		Status:        "running",
		Running:       true,
		Connected:     true,
	}})
	registry.SelectBrowserProfile("s1", SharedSessionBrowserProfileSelection{
		Backend:       "proxy",
		Profile:       "work",
		RuntimeTarget: "node",
		BrowserApp:    "Arc",
		Source:        "remember_profile",
	})

	projected := SnapshotSharedSessionBrowserProjectedProfilesForScope(registry, "s1", BrowserRuntimeInfo{
		Backend: "proxy-tabs",
		Target:  "node",
	}, "work")
	if len(projected) != 1 {
		t.Fatalf("expected 1 projected profile, got %d", len(projected))
	}
	if !projected[0].Selected {
		t.Fatalf("expected projected profile to be selected")
	}
	if projected[0].State.BrowserApp != "Arc" {
		t.Fatalf("expected missing browser app to be backfilled from selection, got %q", projected[0].State.BrowserApp)
	}
}

func TestProjectSharedSessionBrowserObservedProfilesMarksSelectedProfile(t *testing.T) {
	projected := ProjectSharedSessionBrowserObservedProfiles("proxy-cdp", "node", []BrowserProfileInfo{{
		Profile: "work",
		Status:  "running",
		Running: true,
	}}, &SharedSessionBrowserProfileSelection{
		Backend:       "proxy",
		Profile:       "work",
		RuntimeTarget: "node",
	})
	if len(projected) != 1 {
		t.Fatalf("expected 1 projected profile, got %d", len(projected))
	}
	if !projected[0].Selected {
		t.Fatalf("expected observed profile to be marked selected")
	}
	if projected[0].State.Backend != "proxy-cdp" || projected[0].State.RuntimeTarget != "node" {
		t.Fatalf("expected projected observation to preserve backend/target identity, got %+v", projected[0].State)
	}
}

func TestSharedSessionBrowserDiscoveredProfilesDedupesTrimmedNames(t *testing.T) {
	profiles := SharedSessionBrowserDiscoveredProfiles([]BrowserProfileInfo{
		{Profile: " work "},
		{Profile: "relay"},
		{Profile: "work"},
		{Profile: "  "},
	})
	if len(profiles) != 2 || profiles[0] != "work" || profiles[1] != "relay" {
		t.Fatalf("unexpected discovered profile list: %#v", profiles)
	}
}

func TestObserveSharedSessionBrowserProfilesUsesRegistryScopedSnapshotAndSelection(t *testing.T) {
	registry := NewBrowserSessionStateRegistry()
	registry.SelectBrowserProfile("s1", SharedSessionBrowserProfileSelection{
		Backend:       "proxy",
		Profile:       "work",
		RuntimeTarget: "node",
		BrowserApp:    "Arc",
		Source:        "remember_profile",
	})
	backend := &executionTestBackend{
		profilesResp: BrowserProfilesResult{
			Backend:        "proxy",
			DefaultProfile: "work",
			Profiles: []BrowserProfileInfo{
				{Profile: "work", Status: "running", Running: true, Connected: true},
			},
		},
	}

	observation := ObserveSharedSessionBrowserProfiles(
		context.Background(),
		backend,
		registry,
		"s1",
		BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
		"work",
		time.Minute,
	)

	if len(backend.profilesReqs) != 1 || backend.profilesReqs[0].Profile != "work" {
		t.Fatalf("expected one RuntimeProfiles call for work, got %#v", backend.profilesReqs)
	}
	if observation.Profiles == nil || observation.Profiles.Backend != "proxy" {
		t.Fatalf("expected raw profiles observation, got %#v", observation.Profiles)
	}
	if len(observation.Snapshot) != 1 || observation.Snapshot[0].Profile != "work" || observation.Snapshot[0].RuntimeTarget != "node" {
		t.Fatalf("expected scoped synced snapshot, got %#v", observation.Snapshot)
	}
	if len(observation.Projected) != 1 || !observation.Projected[0].Selected || observation.Projected[0].State.BrowserApp != "Arc" {
		t.Fatalf("expected projected selection/browser app from registry, got %#v", observation.Projected)
	}
}

func TestObserveSharedSessionBrowserProfilesFallsBackToRawProjectionWithoutRegistry(t *testing.T) {
	backend := &executionTestBackend{
		profilesResp: BrowserProfilesResult{
			Backend: "proxy",
			Profiles: []BrowserProfileInfo{
				{Profile: "work", BrowserApp: "Chromium", Status: "running", Running: true},
			},
		},
	}

	observation := ObserveSharedSessionBrowserProfiles(
		context.Background(),
		backend,
		nil,
		"",
		BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
		"work",
		time.Minute,
	)

	if observation.ProfilesErr != nil {
		t.Fatalf("expected successful raw observation, got %v", observation.ProfilesErr)
	}
	if len(observation.Snapshot) != 1 || observation.Snapshot[0].Profile != "work" {
		t.Fatalf("expected raw mapped snapshot, got %#v", observation.Snapshot)
	}
	if len(observation.Projected) != 1 || observation.Projected[0].Selected || observation.Projected[0].State.BrowserApp != "Chromium" {
		t.Fatalf("expected raw projected profiles without selection, got %#v", observation.Projected)
	}
}

func TestValidateSharedSessionBrowserSelectedProfileResolvesBrowserAppAndMissing(t *testing.T) {
	browserApp, validated, found := ValidateSharedSessionBrowserSelectedProfile(
		"work",
		"",
		BrowserProfilesResult{
			DefaultProfile: "work",
			Profiles: []BrowserProfileInfo{
				{Profile: "work", BrowserApp: "Arc"},
			},
		},
	)
	if !validated || !found || browserApp != "Arc" {
		t.Fatalf("expected validated profile with resolved browser app, got browserApp=%q validated=%v found=%v", browserApp, validated, found)
	}

	browserApp, validated, found = ValidateSharedSessionBrowserSelectedProfile(
		"missing",
		"",
		BrowserProfilesResult{
			DefaultProfile: "work",
			Note:           "profiles available",
		},
	)
	if !validated || found || browserApp != "" {
		t.Fatalf("expected validated missing profile, got browserApp=%q validated=%v found=%v", browserApp, validated, found)
	}
}
