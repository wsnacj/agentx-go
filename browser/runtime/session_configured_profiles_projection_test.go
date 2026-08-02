package browserruntime

import "testing"

func TestProjectSharedSessionBrowserConfiguredProfilesDefaultsLegacyHostToSingleFallback(t *testing.T) {
	got := ProjectSharedSessionBrowserConfiguredProfiles(
		SharedSessionBrowserConfiguredProfilesProjectionRequest{
			LegacySystemHost:      true,
			LegacyFallbackProfile: "default",
		},
	)
	if len(got) != 1 || got[0] != "default" {
		t.Fatalf("expected legacy host configured profiles to collapse to default only, got %#v", got)
	}
}

func TestProjectSharedSessionBrowserConfiguredProfilesKeepsManagedFallbackOrder(t *testing.T) {
	got := ProjectSharedSessionBrowserConfiguredProfiles(
		SharedSessionBrowserConfiguredProfilesProjectionRequest{
			SelectedProfile:         "workbench",
			AppendManagedGenericSet: true,
		},
	)
	if len(got) != 4 || got[0] != "workbench" || got[1] != "default" || got[2] != "isolated" || got[3] != "relay" {
		t.Fatalf("expected managed configured profiles to keep selected profile ahead of generic fallback, got %#v", got)
	}
}

func TestProjectSharedSessionBrowserConfiguredProfilesIncludesBindingTargetProfile(t *testing.T) {
	got := ProjectSharedSessionBrowserConfiguredProfiles(
		SharedSessionBrowserConfiguredProfilesProjectionRequest{
			SessionTargetProfile:    "isolated",
			SelectedProfile:         "workbench",
			AppendManagedGenericSet: true,
		},
	)
	if len(got) != 4 || got[0] != "isolated" || got[1] != "workbench" || got[2] != "default" || got[3] != "relay" {
		t.Fatalf("expected binding target profile to lead configured profiles before selected profile, got %#v", got)
	}
}
