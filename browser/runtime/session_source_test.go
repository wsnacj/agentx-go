package browserruntime

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"
)

func TestObserveSharedSessionBrowserRawStatusTrimsProfileAndPreservesError(t *testing.T) {
	backend := &statusProfilesObservationTestBackend{statusErr: errors.New("status failed")}
	observation := ObserveSharedSessionBrowserRawStatus(context.Background(), backend, " work ")
	if observation.RequestedProfile != "work" {
		t.Fatalf("expected trimmed requested profile, got %#v", observation)
	}
	if observation.ObservedAt.IsZero() {
		t.Fatalf("expected raw status observation timestamp, got %#v", observation)
	}
	if observation.Err == nil || observation.Err.Error() != "status failed" {
		t.Fatalf("expected raw status error, got %#v", observation.Err)
	}
	if len(backend.statusReqs) != 1 || backend.statusReqs[0].Profile != "work" {
		t.Fatalf("expected raw status request for trimmed profile, got %#v", backend.statusReqs)
	}
}

func TestSharedSessionBrowserRawObservationExpired(t *testing.T) {
	now := time.Date(2026, time.May, 3, 9, 0, 0, 0, time.UTC)
	if !SharedSessionBrowserRawObservationExpired(time.Time{}, now, 2*time.Second) {
		t.Fatal("expected zero observation time to be expired")
	}
	if SharedSessionBrowserRawObservationExpired(now.Add(-2*time.Second), now, 2*time.Second) {
		t.Fatal("expected observation exactly at ttl boundary to remain fresh")
	}
	if !SharedSessionBrowserRawObservationExpired(now.Add(-2*time.Second-time.Nanosecond), now, 2*time.Second) {
		t.Fatal("expected observation older than ttl to expire")
	}
	if SharedSessionBrowserRawObservationExpired(now.Add(time.Second), now, 2*time.Second) {
		t.Fatal("expected future observation to remain fresh")
	}
}

func TestSharedSessionBrowserRawObservationReplayTTL(t *testing.T) {
	if SharedSessionBrowserRawObservationReplayTTL != 2*time.Second {
		t.Fatalf("unexpected raw observation replay ttl: %s", SharedSessionBrowserRawObservationReplayTTL)
	}
}

func TestFreshSharedSessionBrowserRawObservationHelpersRequireFreshSourceTime(t *testing.T) {
	now := time.Date(2026, time.May, 3, 9, 0, 0, 0, time.UTC)
	ttl := 2 * time.Second
	freshAt := now.Add(-time.Second)
	staleAt := now.Add(-ttl - time.Nanosecond)

	tests := []struct {
		name     string
		fresh    func() bool
		stale    func() bool
		zeroTime func() bool
	}{
		{
			name: "status",
			fresh: func() bool {
				got, ok := FreshSharedSessionBrowserRawStatusObservation(SharedSessionBrowserRawStatusObservation{
					RequestedProfile: " work ",
					Status:           &BrowserProfileStatusResult{Profile: "work"},
					ObservedAt:       freshAt,
				}, now, ttl)
				return ok && got.RequestedProfile == "work" && got.Status != nil
			},
			stale: func() bool {
				_, ok := FreshSharedSessionBrowserRawStatusObservation(SharedSessionBrowserRawStatusObservation{Status: &BrowserProfileStatusResult{}, ObservedAt: staleAt}, now, ttl)
				return ok
			},
			zeroTime: func() bool {
				_, ok := FreshSharedSessionBrowserRawStatusObservation(SharedSessionBrowserRawStatusObservation{Status: &BrowserProfileStatusResult{}}, now, ttl)
				return ok
			},
		},
		{
			name: "profiles",
			fresh: func() bool {
				got, ok := FreshSharedSessionBrowserRawProfilesObservation(SharedSessionBrowserRawProfilesObservation{
					RequestedProfile: " work ",
					Profiles:         &BrowserProfilesResult{DefaultProfile: "work"},
					ObservedAt:       freshAt,
				}, now, ttl)
				return ok && got.RequestedProfile == "work" && got.Profiles != nil
			},
			stale: func() bool {
				_, ok := FreshSharedSessionBrowserRawProfilesObservation(SharedSessionBrowserRawProfilesObservation{Profiles: &BrowserProfilesResult{}, ObservedAt: staleAt}, now, ttl)
				return ok
			},
			zeroTime: func() bool {
				_, ok := FreshSharedSessionBrowserRawProfilesObservation(SharedSessionBrowserRawProfilesObservation{Profiles: &BrowserProfilesResult{}}, now, ttl)
				return ok
			},
		},
		{
			name: "lifecycle",
			fresh: func() bool {
				got, ok := FreshSharedSessionBrowserRawLifecycleObservation(SharedSessionBrowserRawLifecycleObservation{
					Profile:    " work ",
					Status:     &BrowserProfileStatusResult{Profile: "work"},
					ObservedAt: freshAt,
				}, now, ttl)
				return ok && got.Profile == "work" && got.Status != nil
			},
			stale: func() bool {
				_, ok := FreshSharedSessionBrowserRawLifecycleObservation(SharedSessionBrowserRawLifecycleObservation{Status: &BrowserProfileStatusResult{}, ObservedAt: staleAt}, now, ttl)
				return ok
			},
			zeroTime: func() bool {
				_, ok := FreshSharedSessionBrowserRawLifecycleObservation(SharedSessionBrowserRawLifecycleObservation{Status: &BrowserProfileStatusResult{}}, now, ttl)
				return ok
			},
		},
		{
			name: "tabs",
			fresh: func() bool {
				got, ok := FreshSharedSessionBrowserRawTabsObservation(SharedSessionBrowserRawTabsObservation{
					RequestedProfile: " work ",
					Action:           " select ",
					ObservedAt:       freshAt,
				}, now, ttl)
				return ok && got.RequestedProfile == "work" && got.Action == "select"
			},
			stale: func() bool {
				_, ok := FreshSharedSessionBrowserRawTabsObservation(SharedSessionBrowserRawTabsObservation{Action: "select", ObservedAt: staleAt}, now, ttl)
				return ok
			},
			zeroTime: func() bool {
				_, ok := FreshSharedSessionBrowserRawTabsObservation(SharedSessionBrowserRawTabsObservation{Action: "select"}, now, ttl)
				return ok
			},
		},
		{
			name: "navigation",
			fresh: func() bool {
				got, ok := FreshSharedSessionBrowserRawNavigationObservation(SharedSessionBrowserRawNavigationObservation{
					RequestedProfile: " work ",
					RequestedURL:     " https://example.com ",
					ObservedAt:       freshAt,
				}, now, ttl)
				return ok && got.RequestedProfile == "work" && got.RequestedURL == "https://example.com"
			},
			stale: func() bool {
				_, ok := FreshSharedSessionBrowserRawNavigationObservation(SharedSessionBrowserRawNavigationObservation{RequestedURL: "https://example.com", ObservedAt: staleAt}, now, ttl)
				return ok
			},
			zeroTime: func() bool {
				_, ok := FreshSharedSessionBrowserRawNavigationObservation(SharedSessionBrowserRawNavigationObservation{RequestedURL: "https://example.com"}, now, ttl)
				return ok
			},
		},
		{
			name: "open",
			fresh: func() bool {
				got, ok := FreshSharedSessionBrowserRawOpenObservation(SharedSessionBrowserRawOpenObservation{
					RequestedProfile: " work ",
					URL:              " https://example.com ",
					ObservedAt:       freshAt,
				}, now, ttl)
				return ok && got.RequestedProfile == "work" && got.URL == "https://example.com"
			},
			stale: func() bool {
				_, ok := FreshSharedSessionBrowserRawOpenObservation(SharedSessionBrowserRawOpenObservation{URL: "https://example.com", ObservedAt: staleAt}, now, ttl)
				return ok
			},
			zeroTime: func() bool {
				_, ok := FreshSharedSessionBrowserRawOpenObservation(SharedSessionBrowserRawOpenObservation{URL: "https://example.com"}, now, ttl)
				return ok
			},
		},
		{
			name: "target",
			fresh: func() bool {
				got, ok := FreshSharedSessionBrowserRawTargetObservation(SharedSessionBrowserRawTargetObservation{
					RequestedProfile: " work ",
					Source:           " console ",
					ObservedAt:       freshAt,
				}, now, ttl)
				return ok && got.RequestedProfile == "work" && got.Source == "console"
			},
			stale: func() bool {
				_, ok := FreshSharedSessionBrowserRawTargetObservation(SharedSessionBrowserRawTargetObservation{Source: "console", ObservedAt: staleAt}, now, ttl)
				return ok
			},
			zeroTime: func() bool {
				_, ok := FreshSharedSessionBrowserRawTargetObservation(SharedSessionBrowserRawTargetObservation{Source: "console"}, now, ttl)
				return ok
			},
		},
		{
			name: "route mutation",
			fresh: func() bool {
				got, ok := FreshSharedSessionBrowserRawRouteMutationObservation(SharedSessionBrowserRawRouteMutationObservation{
					RequestedProfile: " work ",
					Kind:             " current_target ",
					ObservedAt:       freshAt,
				}, now, ttl)
				return ok && got.RequestedProfile == "work" && got.Kind == "current_target"
			},
			stale: func() bool {
				_, ok := FreshSharedSessionBrowserRawRouteMutationObservation(SharedSessionBrowserRawRouteMutationObservation{Kind: "current_target", ObservedAt: staleAt}, now, ttl)
				return ok
			},
			zeroTime: func() bool {
				_, ok := FreshSharedSessionBrowserRawRouteMutationObservation(SharedSessionBrowserRawRouteMutationObservation{Kind: "current_target"}, now, ttl)
				return ok
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.fresh() {
				t.Fatal("expected fresh source-time observation to replay")
			}
			if tt.stale() {
				t.Fatal("expected stale source-time observation to be rejected")
			}
			if tt.zeroTime() {
				t.Fatal("expected zero source-time observation to be rejected even with payload")
			}
		})
	}
}

func TestSharedSessionBrowserRawObservationKeys(t *testing.T) {
	tests := []struct {
		name            string
		requested       string
		resolved        string
		expectedProfile []string
	}{
		{name: "empty", expectedProfile: []string{""}},
		{name: "requested only", requested: " work ", expectedProfile: []string{"work", ""}},
		{name: "resolved only", resolved: " work ", expectedProfile: []string{"", "work"}},
		{name: "same requested and resolved", requested: " work ", resolved: "work", expectedProfile: []string{"work"}},
		{name: "different requested and resolved", requested: " relay ", resolved: " work ", expectedProfile: []string{"relay", "work"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SharedSessionBrowserRawObservationKeys(tt.requested, tt.resolved)
			if len(got) != len(tt.expectedProfile) {
				t.Fatalf("expected keys %#v, got %#v", tt.expectedProfile, got)
			}
			for i := range got {
				if got[i] != tt.expectedProfile[i] {
					t.Fatalf("expected keys %#v, got %#v", tt.expectedProfile, got)
				}
			}
		})
	}
}

func TestSharedSessionBrowserRawStatusObservationLookupKeysForProfiles(t *testing.T) {
	got := SharedSessionBrowserRawStatusObservationLookupKeysForProfiles(
		" relay ",
		&BrowserProfilesResult{DefaultProfile: " work "},
	)
	expected := []string{"relay", "work"}
	if len(got) != len(expected) {
		t.Fatalf("expected status lookup keys %#v, got %#v", expected, got)
	}
	for i := range got {
		if got[i] != expected[i] {
			t.Fatalf("expected status lookup keys %#v, got %#v", expected, got)
		}
	}

	got = SharedSessionBrowserRawStatusObservationLookupKeysForProfiles(" relay ", nil)
	expected = []string{"relay", ""}
	if len(got) != len(expected) {
		t.Fatalf("expected status fallback keys %#v, got %#v", expected, got)
	}
	for i := range got {
		if got[i] != expected[i] {
			t.Fatalf("expected status fallback keys %#v, got %#v", expected, got)
		}
	}
}

func TestSharedSessionBrowserRawProfilesObservationLookupKeysForStatus(t *testing.T) {
	got := SharedSessionBrowserRawProfilesObservationLookupKeysForStatus(
		" relay ",
		&BrowserProfileStatusResult{Profile: " work "},
	)
	expected := []string{"relay", "work"}
	if len(got) != len(expected) {
		t.Fatalf("expected profiles lookup keys %#v, got %#v", expected, got)
	}
	for i := range got {
		if got[i] != expected[i] {
			t.Fatalf("expected profiles lookup keys %#v, got %#v", expected, got)
		}
	}

	got = SharedSessionBrowserRawProfilesObservationLookupKeysForStatus("", &BrowserProfileStatusResult{Profile: " work "})
	expected = []string{"", "work"}
	if len(got) != len(expected) {
		t.Fatalf("expected profiles empty-request keys %#v, got %#v", expected, got)
	}
	for i := range got {
		if got[i] != expected[i] {
			t.Fatalf("expected profiles empty-request keys %#v, got %#v", expected, got)
		}
	}
}

func TestSharedSessionBrowserRawObservationProfileFilter(t *testing.T) {
	filter := BuildSharedSessionBrowserRawObservationProfileFilter(" work ", "", "relay", "work")
	if filter.Empty() {
		t.Fatal("expected profile filter to keep non-empty profiles")
	}
	if !filter.Matches(" work ") {
		t.Fatal("expected filter to match trimmed work profile")
	}
	if !filter.Matches("missing", " relay ") {
		t.Fatal("expected filter to match any candidate profile")
	}
	if filter.Matches("", "missing") {
		t.Fatal("expected filter to ignore empty and missing candidates")
	}

	empty := BuildSharedSessionBrowserRawObservationProfileFilter("", "   ")
	if !empty.Empty() {
		t.Fatalf("expected empty profile filter, got %#v", empty)
	}
	if empty.Matches("work") {
		t.Fatal("expected empty profile filter to match nothing")
	}
}

func TestSharedSessionBrowserRawObservationProfileCandidates(t *testing.T) {
	tests := []struct {
		name     string
		got      []string
		expected []string
	}{
		{
			name:     "status cache key and resolved profile",
			got:      SharedSessionBrowserRawStatusObservationProfileCandidates(" relay ", SharedSessionBrowserRawStatusObservation{Status: &BrowserProfileStatusResult{Profile: " work "}}),
			expected: []string{"relay", "work"},
		},
		{
			name:     "tabs deduplicates requested profile",
			got:      SharedSessionBrowserRawTabsObservationProfileCandidates(" work ", SharedSessionBrowserRawTabsObservation{RequestedProfile: "work"}),
			expected: []string{"work"},
		},
		{
			name: "route mutation keeps route profile",
			got: SharedSessionBrowserRawRouteMutationObservationProfileCandidates(" relay ", SharedSessionBrowserRawRouteMutationObservation{
				RequestedProfile: " requested ",
				Route:            BrowserSessionRoute{Profile: " route "},
			}),
			expected: []string{"relay", "requested", "route"},
		},
		{
			name:     "navigation uses requested profile",
			got:      SharedSessionBrowserRawNavigationObservationProfileCandidates(" relay ", SharedSessionBrowserRawNavigationObservation{RequestedProfile: " requested "}),
			expected: []string{"relay", "requested"},
		},
		{
			name:     "open uses requested profile",
			got:      SharedSessionBrowserRawOpenObservationProfileCandidates(" relay ", SharedSessionBrowserRawOpenObservation{RequestedProfile: " requested "}),
			expected: []string{"relay", "requested"},
		},
		{
			name:     "target uses requested profile",
			got:      SharedSessionBrowserRawTargetObservationProfileCandidates(" relay ", SharedSessionBrowserRawTargetObservation{RequestedProfile: " requested "}),
			expected: []string{"relay", "requested"},
		},
		{
			name:     "lifecycle cache key and resolved profile",
			got:      SharedSessionBrowserRawLifecycleObservationProfileCandidates(" requested ", SharedSessionBrowserRawLifecycleObservation{Profile: " resolved "}),
			expected: []string{"requested", "resolved"},
		},
		{
			name: "combined status and default profile",
			got: SharedSessionBrowserRawStatusAndProfilesObservationProfileCandidates(" relay ", SharedSessionBrowserRawStatusAndProfilesObservation{
				Status:   &BrowserProfileStatusResult{Profile: " work "},
				Profiles: &BrowserProfilesResult{DefaultProfile: " default "},
			}),
			expected: []string{"relay", "work", "default"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !slices.Equal(tt.got, tt.expected) {
				t.Fatalf("expected profile candidates %#v, got %#v", tt.expected, tt.got)
			}
		})
	}
}

func TestBuildSharedSessionBrowserRawStatusObservation(t *testing.T) {
	observedAt := time.Date(2026, time.May, 3, 9, 0, 0, 0, time.UTC)
	status := &BrowserProfileStatusResult{Profile: " work ", Status: "running"}
	got, ok := BuildSharedSessionBrowserRawStatusObservation(" work ", status, nil, observedAt)
	if !ok {
		t.Fatal("expected raw status observation")
	}
	status.Profile = "mutated"
	if got.RequestedProfile != "work" ||
		got.Status == nil ||
		got.Status.Profile != " work " ||
		got.Err != nil ||
		!got.ObservedAt.Equal(observedAt) {
		t.Fatalf("unexpected raw status observation: %#v", got)
	}

	if got, ok := BuildSharedSessionBrowserRawStatusObservation("work", nil, nil, time.Time{}); ok {
		t.Fatalf("expected empty status observation to be rejected, got %#v", got)
	}
}

func TestBuildSharedSessionBrowserRawProfilesObservation(t *testing.T) {
	observedAt := time.Date(2026, time.May, 3, 9, 0, 0, 0, time.UTC)
	profiles := &BrowserProfilesResult{DefaultProfile: " work "}
	got, ok := BuildSharedSessionBrowserRawProfilesObservation(" work ", profiles, nil, observedAt)
	if !ok {
		t.Fatal("expected raw profiles observation")
	}
	profiles.DefaultProfile = "mutated"
	if got.RequestedProfile != "work" ||
		got.Profiles == nil ||
		got.Profiles.DefaultProfile != " work " ||
		got.Err != nil ||
		!got.ObservedAt.Equal(observedAt) {
		t.Fatalf("unexpected raw profiles observation: %#v", got)
	}

	err := errors.New("profiles failed")
	got, ok = BuildSharedSessionBrowserRawProfilesObservation(" work ", nil, err, observedAt)
	if !ok || got.Err != err || !got.ObservedAt.Equal(observedAt) {
		t.Fatalf("expected raw profiles error observation, got ok=%v observation=%#v", ok, got)
	}
}

func TestBuildSharedSessionBrowserRawLifecycleObservation(t *testing.T) {
	observedAt := time.Date(2026, time.May, 3, 9, 0, 0, 0, time.UTC)
	status := &BrowserProfileStatusResult{Profile: " resolved ", Status: "running"}
	got, ok := BuildSharedSessionBrowserRawLifecycleObservation(" requested ", " fallback ", status, nil, observedAt)
	if !ok {
		t.Fatal("expected raw lifecycle observation")
	}
	status.Profile = "mutated"
	if got.Profile != "resolved" ||
		got.Status == nil ||
		got.Status.Profile != " resolved " ||
		got.Err != nil ||
		!got.ObservedAt.Equal(observedAt) {
		t.Fatalf("unexpected raw lifecycle observation: %#v", got)
	}

	got, ok = BuildSharedSessionBrowserRawLifecycleObservation("", " fallback ", nil, errors.New("start failed"), observedAt)
	if !ok || got.Profile != "fallback" || got.Err == nil || !got.ObservedAt.Equal(observedAt) {
		t.Fatalf("expected raw lifecycle error observation, got ok=%v observation=%#v", ok, got)
	}
}

func TestSharedSessionBrowserRawStatusAndProfilesObservationProvided(t *testing.T) {
	if SharedSessionBrowserRawStatusAndProfilesObservationProvided(SharedSessionBrowserRawStatusAndProfilesObservation{}) {
		t.Fatal("expected empty combined status/profiles observation to be absent")
	}
	if !SharedSessionBrowserRawStatusAndProfilesObservationProvided(SharedSessionBrowserRawStatusAndProfilesObservation{
		StatusObservedAt: time.Date(2026, time.May, 3, 9, 0, 0, 0, time.UTC),
	}) {
		t.Fatal("expected status timestamp to provide combined status/profiles observation")
	}
	if !SharedSessionBrowserRawStatusAndProfilesObservationProvided(SharedSessionBrowserRawStatusAndProfilesObservation{
		ProfilesErr: errors.New("profiles failed"),
	}) {
		t.Fatal("expected profiles error to provide combined status/profiles observation")
	}
}

func TestBuildSharedSessionBrowserRawStatusAndProfilesObservation(t *testing.T) {
	statusObservedAt := time.Date(2026, time.May, 3, 9, 0, 0, 0, time.UTC)
	profilesObservedAt := statusObservedAt.Add(time.Second)
	statusObservation := SharedSessionBrowserRawStatusObservation{
		RequestedProfile: " work ",
		Status:           &BrowserProfileStatusResult{Profile: "work", Status: "running"},
		ObservedAt:       statusObservedAt,
	}
	profilesObservation := SharedSessionBrowserRawProfilesObservation{
		RequestedProfile: " fallback ",
		Profiles:         &BrowserProfilesResult{DefaultProfile: "work"},
		ObservedAt:       profilesObservedAt,
	}
	got, ok := BuildSharedSessionBrowserRawStatusAndProfilesObservation(statusObservation, profilesObservation)
	if !ok {
		t.Fatal("expected combined raw status/profiles observation")
	}
	if got.RequestedProfile != "work" ||
		got.Status != statusObservation.Status ||
		got.StatusErr != nil ||
		!got.StatusObservedAt.Equal(statusObservedAt) ||
		got.Profiles != profilesObservation.Profiles ||
		got.ProfilesErr != nil ||
		!got.ProfilesObservedAt.Equal(profilesObservedAt) {
		t.Fatalf("unexpected combined raw status/profiles observation: %#v", got)
	}

	got, ok = BuildSharedSessionBrowserRawStatusAndProfilesObservation(
		SharedSessionBrowserRawStatusObservation{RequestedProfile: "work", Err: errors.New("status failed"), ObservedAt: statusObservedAt},
		SharedSessionBrowserRawProfilesObservation{},
	)
	if !ok || got.StatusErr == nil || !got.StatusObservedAt.Equal(statusObservedAt) {
		t.Fatalf("expected combined raw status error observation, got ok=%v observation=%#v", ok, got)
	}
}

func TestBuildFreshSharedSessionBrowserRawStatusAndProfilesObservation(t *testing.T) {
	now := time.Date(2026, time.May, 3, 9, 0, 0, 0, time.UTC)
	statusObservation := SharedSessionBrowserRawStatusObservation{
		RequestedProfile: "work",
		Status:           &BrowserProfileStatusResult{Profile: "work"},
		ObservedAt:       now.Add(-time.Second),
	}
	profilesObservation := SharedSessionBrowserRawProfilesObservation{
		RequestedProfile: "work",
		Profiles:         &BrowserProfilesResult{DefaultProfile: "work"},
		ObservedAt:       now.Add(-time.Second),
	}

	got, ok := BuildFreshSharedSessionBrowserRawStatusAndProfilesObservation(statusObservation, profilesObservation, now, 2*time.Second)
	if !ok || got.Status != statusObservation.Status || got.Profiles != profilesObservation.Profiles {
		t.Fatalf("expected fresh combined raw observation, got ok=%v observation=%#v", ok, got)
	}

	staleStatus := statusObservation
	staleStatus.ObservedAt = now.Add(-3 * time.Second)
	if got, ok := BuildFreshSharedSessionBrowserRawStatusAndProfilesObservation(staleStatus, profilesObservation, now, 2*time.Second); ok {
		t.Fatalf("expected stale status to reject combined observation, got %#v", got)
	}

	staleProfiles := profilesObservation
	staleProfiles.ObservedAt = now.Add(-3 * time.Second)
	if got, ok := BuildFreshSharedSessionBrowserRawStatusAndProfilesObservation(statusObservation, staleProfiles, now, 2*time.Second); ok {
		t.Fatalf("expected stale profiles to reject combined observation, got %#v", got)
	}
}

func TestBuildSharedSessionBrowserRawStatusAndProfilesObservationForRequest(t *testing.T) {
	statusObservedAt := time.Date(2026, time.May, 3, 9, 0, 0, 0, time.UTC)
	profilesObservedAt := statusObservedAt.Add(time.Second)
	statusObservation := SharedSessionBrowserRawStatusObservation{
		RequestedProfile: " status-profile ",
		Status:           &BrowserProfileStatusResult{Profile: "status-profile"},
		ObservedAt:       statusObservedAt,
	}
	profilesObservation := SharedSessionBrowserRawProfilesObservation{
		RequestedProfile: " profiles-profile ",
		Profiles:         &BrowserProfilesResult{DefaultProfile: "profiles-profile"},
		ObservedAt:       profilesObservedAt,
	}

	got := BuildSharedSessionBrowserRawStatusAndProfilesObservationForRequest(
		" requested ",
		true,
		statusObservation,
		false,
		profilesObservation,
	)
	if got.RequestedProfile != "requested" ||
		got.Status != statusObservation.Status ||
		!got.StatusObservedAt.Equal(statusObservedAt) ||
		got.Profiles != nil ||
		!got.ProfilesObservedAt.IsZero() {
		t.Fatalf("unexpected status-only combined raw observation: %#v", got)
	}

	got = BuildSharedSessionBrowserRawStatusAndProfilesObservationForRequest(
		"",
		false,
		statusObservation,
		true,
		profilesObservation,
	)
	if got.RequestedProfile != "profiles-profile" ||
		got.Status != nil ||
		!got.StatusObservedAt.IsZero() ||
		got.Profiles != profilesObservation.Profiles ||
		!got.ProfilesObservedAt.Equal(profilesObservedAt) {
		t.Fatalf("unexpected profiles-only combined raw observation: %#v", got)
	}
}

func TestPruneExpiredSharedSessionBrowserRawStatusAndProfilesObservation(t *testing.T) {
	now := time.Date(2026, time.May, 3, 9, 0, 0, 0, time.UTC)
	observation := SharedSessionBrowserRawStatusAndProfilesObservation{
		RequestedProfile:   "work",
		Status:             &BrowserProfileStatusResult{Profile: "work"},
		StatusErr:          errors.New("status stale"),
		StatusObservedAt:   now.Add(-3 * time.Second),
		Profiles:           &BrowserProfilesResult{DefaultProfile: "work"},
		ProfilesErr:        errors.New("profiles fresh"),
		ProfilesObservedAt: now.Add(-time.Second),
	}

	got := PruneExpiredSharedSessionBrowserRawStatusAndProfilesObservation(observation, now, 2*time.Second)
	if got.Status != nil || got.StatusErr != nil || !got.StatusObservedAt.IsZero() {
		t.Fatalf("expected stale status half to be pruned, got %#v", got)
	}
	if got.Profiles == nil || got.ProfilesErr == nil || !got.ProfilesObservedAt.Equal(now.Add(-time.Second)) {
		t.Fatalf("expected fresh profiles half to remain, got %#v", got)
	}

	got = PruneExpiredSharedSessionBrowserRawStatusAndProfilesObservation(observation, now, 500*time.Millisecond)
	if SharedSessionBrowserRawStatusAndProfilesObservationProvided(got) {
		t.Fatalf("expected both halves to be pruned, got %#v", got)
	}
}

func TestFreshSharedSessionBrowserRawStatusAndProfilesObservationPrunesAndReportsReplayable(t *testing.T) {
	now := time.Date(2026, time.May, 3, 9, 0, 0, 0, time.UTC)
	observation := SharedSessionBrowserRawStatusAndProfilesObservation{
		RequestedProfile:   " work ",
		Status:             &BrowserProfileStatusResult{Profile: "work"},
		StatusErr:          errors.New("status stale"),
		StatusObservedAt:   now.Add(-3 * time.Second),
		Profiles:           &BrowserProfilesResult{DefaultProfile: "work"},
		ProfilesErr:        errors.New("profiles fresh"),
		ProfilesObservedAt: now.Add(-time.Second),
	}

	got, ok := FreshSharedSessionBrowserRawStatusAndProfilesObservation(observation, now, 2*time.Second)
	if !ok {
		t.Fatal("expected profiles half to remain replayable")
	}
	if got.RequestedProfile != "work" {
		t.Fatalf("expected requested profile to be normalized, got %#v", got.RequestedProfile)
	}
	if got.Status != nil || got.StatusErr != nil || !got.StatusObservedAt.IsZero() {
		t.Fatalf("expected stale status half to be pruned, got %#v", got)
	}
	if got.Profiles == nil || got.ProfilesErr == nil || !got.ProfilesObservedAt.Equal(now.Add(-time.Second)) {
		t.Fatalf("expected fresh profiles half to remain, got %#v", got)
	}

	if got, ok := FreshSharedSessionBrowserRawStatusAndProfilesObservation(observation, now, 500*time.Millisecond); ok {
		t.Fatalf("expected fully stale combined observation to be rejected, got %#v", got)
	}
}

func TestSharedSessionBrowserRawStatusAndProfilesObservationKeys(t *testing.T) {
	got := SharedSessionBrowserRawStatusAndProfilesObservationKeys(
		SharedSessionBrowserRawStatusObservation{
			RequestedProfile: " relay ",
			Status:           &BrowserProfileStatusResult{Profile: " work "},
		},
		SharedSessionBrowserRawProfilesObservation{
			RequestedProfile: " relay ",
			Profiles:         &BrowserProfilesResult{DefaultProfile: " default "},
		},
	)
	expected := []string{"relay", "work", "default"}
	if len(got) != len(expected) {
		t.Fatalf("expected combined keys %#v, got %#v", expected, got)
	}
	for i := range got {
		if got[i] != expected[i] {
			t.Fatalf("expected combined keys %#v, got %#v", expected, got)
		}
	}

	got = SharedSessionBrowserRawStatusAndProfilesObservationKeys(
		SharedSessionBrowserRawStatusObservation{},
		SharedSessionBrowserRawProfilesObservation{},
	)
	if len(got) != 1 || got[0] != "" {
		t.Fatalf("expected empty combined key, got %#v", got)
	}
}

func TestSharedSessionBrowserRawRouteMutationObservationProvided(t *testing.T) {
	observedAt := time.Date(2026, time.May, 3, 9, 0, 0, 0, time.UTC)
	tests := []struct {
		name        string
		observation SharedSessionBrowserRawRouteMutationObservation
		want        bool
	}{
		{
			name: "empty",
		},
		{
			name: "requested profile and route only",
			observation: SharedSessionBrowserRawRouteMutationObservation{
				RequestedProfile: "work",
				Route: BrowserSessionRoute{
					Backend: "proxy",
					Profile: "work",
					Target:  "target-1",
				},
			},
		},
		{
			name: "whitespace only kind",
			observation: SharedSessionBrowserRawRouteMutationObservation{
				Kind: "   ",
			},
		},
		{
			name: "kind",
			observation: SharedSessionBrowserRawRouteMutationObservation{
				Kind: " track_current ",
			},
			want: true,
		},
		{
			name: "review policy state",
			observation: SharedSessionBrowserRawRouteMutationObservation{
				Review: SharedSessionBrowserPendingTargetReviewState{PolicyState: " pending "},
			},
			want: true,
		},
		{
			name: "prior selection",
			observation: SharedSessionBrowserRawRouteMutationObservation{
				PriorSelection: &BrowserSessionTargetSelection{ID: "target-1"},
			},
			want: true,
		},
		{
			name: "source timestamp",
			observation: SharedSessionBrowserRawRouteMutationObservation{
				ObservedAt: observedAt,
			},
			want: true,
		},
		{
			name: "tabs",
			observation: SharedSessionBrowserRawRouteMutationObservation{
				Tabs: []BrowserTab{{Index: 1, URL: "https://example.com"}},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SharedSessionBrowserRawRouteMutationObservationProvided(tt.observation); got != tt.want {
				t.Fatalf("expected provided=%v, got %v for %#v", tt.want, got, tt.observation)
			}
		})
	}
}

func TestNormalizeSharedSessionBrowserRawRouteMutationObservation(t *testing.T) {
	observedAt := time.Date(2026, time.May, 3, 9, 0, 0, 0, time.UTC)
	prior := &BrowserSessionTargetSelection{ID: " prior-target "}
	review := &BrowserSessionTargetReview{
		ID:       " review-1 ",
		URL:      " https://example.com/review ",
		Title:    " Review ",
		Decision: " pending ",
		Reason:   " needs review ",
	}
	observation := SharedSessionBrowserRawRouteMutationObservation{
		RequestedProfile: " provided ",
		Route: BrowserSessionRoute{
			Backend:    " Proxy ",
			Profile:    " Work ",
			Target:     " Node ",
			BrowserApp: " Chromium ",
		},
		Kind:                   " tabs_result ",
		Action:                 " focus ",
		Review:                 SharedSessionBrowserPendingTargetReviewState{Review: review, PolicyState: " review_required ", PolicyReason: " required "},
		Actor:                  " browser_tabs ",
		ExplicitTargetID:       " target-1 ",
		CandidateTargetID:      " target-2 ",
		PriorActiveTargetID:    " prior-active ",
		PriorRequestedTargetID: " prior-requested ",
		TargetID:               " target-3 ",
		FinalURL:               " https://example.com/final ",
		Decision:               " approved ",
		Reason:                 " ok ",
		PriorSelection:         prior,
		PendingTargetID:        " pending-target ",
		URL:                    " https://example.com/current ",
		Title:                  " Current ",
		Note:                   " note ",
		Source:                 " runtime_route_mutation_source ",
		Tabs:                   []BrowserTab{{Index: 1, TargetID: "target-1"}},
		ObservedAt:             observedAt,
	}

	got := NormalizeSharedSessionBrowserRawRouteMutationObservation(observation, " fallback ")
	observation.Tabs[0].TargetID = "mutated"
	prior.ID = "mutated"
	review.ID = "mutated"

	if got.RequestedProfile != "provided" ||
		got.Route.Backend != "proxy" ||
		got.Route.Profile != "work" ||
		got.Route.Target != "node" ||
		got.Route.BrowserApp != "chromium" ||
		got.Kind != "tabs_result" ||
		got.Action != "focus" ||
		got.Review.PolicyState != "review_required" ||
		got.Review.PolicyReason != "required" ||
		got.Review.Review == nil ||
		got.Review.Review.ID != "review-1" ||
		got.Review.Review.URL != "https://example.com/review" ||
		got.Review.Review.Title != "Review" ||
		got.Review.Review.Decision != "pending" ||
		got.Review.Review.Reason != "needs review" ||
		got.Actor != "browser_tabs" ||
		got.ExplicitTargetID != "target-1" ||
		got.CandidateTargetID != "target-2" ||
		got.PriorActiveTargetID != "prior-active" ||
		got.PriorRequestedTargetID != "prior-requested" ||
		got.TargetID != "target-3" ||
		got.FinalURL != "https://example.com/final" ||
		got.Decision != "approved" ||
		got.Reason != "ok" ||
		got.PriorSelection == nil ||
		got.PriorSelection.ID != " prior-target " ||
		got.PendingTargetID != "pending-target" ||
		got.URL != "https://example.com/current" ||
		got.Title != "Current" ||
		got.Note != "note" ||
		got.Source != "runtime_route_mutation_source" ||
		len(got.Tabs) != 1 ||
		got.Tabs[0].TargetID != "target-1" ||
		!got.ObservedAt.Equal(observedAt) {
		t.Fatalf("unexpected normalized raw route mutation: %#v", got)
	}
}

func TestBuildSharedSessionBrowserRawTabsObservation(t *testing.T) {
	observedAt := time.Date(2026, time.May, 3, 9, 0, 0, 0, time.UTC)

	t.Run("rejects nil result", func(t *testing.T) {
		got, ok := BuildSharedSessionBrowserRawTabsObservation("work", nil, nil, observedAt)
		if ok {
			t.Fatalf("expected no raw tabs observation without result, got %#v", got)
		}
	})

	t.Run("projects tabs result and request posture", func(t *testing.T) {
		result := &BrowserTabsResult{
			Action:      " focus ",
			ActiveIndex: 2,
			Tabs: []BrowserTab{
				{Index: 1, TargetID: "target-1"},
				{Index: 2, TargetID: "target-2", URL: "https://example.com/two", Title: "Two"},
			},
			Note: " done ",
		}
		req := &BrowserTabsRequest{
			Action:                 "list",
			TabIndex:               2,
			Force:                  true,
			RememberTarget:         true,
			Review:                 SharedSessionBrowserPendingTargetReviewState{Review: &BrowserSessionTargetReview{ID: " review-1 "}, PolicyState: " pending "},
			Actor:                  " browser_tabs ",
			ExplicitTargetID:       " target-2 ",
			PriorSelection:         &BrowserSessionTargetSelection{ID: "prior-target"},
			PriorActiveTargetID:    " prior-active ",
			PriorRequestedTargetID: " prior-requested ",
		}

		got, ok := BuildSharedSessionBrowserRawTabsObservation(" work ", req, result, observedAt)
		if !ok {
			t.Fatal("expected raw tabs observation")
		}
		result.Tabs[1].TargetID = "mutated"
		req.PriorSelection.ID = "mutated"
		req.Review.Review.ID = "mutated"

		if got.RequestedProfile != "work" ||
			got.Action != "focus" ||
			got.RequestedTabIndex != 2 ||
			!got.Force ||
			!got.RememberTarget ||
			got.Review.PolicyState != "pending" ||
			got.Review.Review == nil ||
			got.Review.Review.ID != "review-1" ||
			got.Actor != "browser_tabs" ||
			got.ExplicitTargetID != "target-2" ||
			got.PriorSelection == nil ||
			got.PriorSelection.ID != "prior-target" ||
			got.PriorActiveTargetID != "prior-active" ||
			got.PriorRequestedTargetID != "prior-requested" ||
			got.Note != "done" ||
			got.ActiveIndex != 2 ||
			len(got.Tabs) != 2 ||
			got.Tabs[1].TargetID != "target-2" ||
			!got.ObservedAt.Equal(observedAt) {
			t.Fatalf("unexpected raw tabs observation: %#v", got)
		}
	})
}

func TestBuildSharedSessionBrowserRawTabsRouteMutationObservation(t *testing.T) {
	observedAt := time.Date(2026, time.May, 3, 9, 0, 0, 0, time.UTC)
	route := BrowserSessionRoute{Backend: " proxy ", Profile: " work ", Target: " target-1 "}

	t.Run("rejects nil result", func(t *testing.T) {
		got, ok := BuildSharedSessionBrowserRawTabsRouteMutationObservation("work", route, nil, nil, observedAt)
		if ok {
			t.Fatalf("expected no tabs route mutation without result, got %#v", got)
		}
	})

	t.Run("projects list active candidate", func(t *testing.T) {
		result := &BrowserTabsResult{
			Action:      " list ",
			BrowserApp:  " Chrome ",
			ActiveIndex: 2,
			Tabs: []BrowserTab{
				{Index: 1, TargetID: "target-1"},
				{Index: 2, TargetID: "target-2", URL: "https://example.com/two", Title: "Two"},
			},
			Note: " done ",
		}
		got, ok := BuildSharedSessionBrowserRawTabsRouteMutationObservation(
			" work ",
			route,
			&BrowserTabsRequest{
				RememberTarget:      true,
				Review:              SharedSessionBrowserPendingTargetReviewState{PolicyState: " pending "},
				Actor:               " browser_tabs ",
				PriorActiveTargetID: " prior-active ",
				PriorSelection:      &BrowserSessionTargetSelection{ID: "prior-target"},
			},
			result,
			observedAt,
		)
		if !ok {
			t.Fatal("expected tabs route mutation")
		}
		result.Tabs[1].TargetID = "changed"
		if got.Kind != "tabs_result" ||
			got.Action != "list" ||
			got.RequestedProfile != "work" ||
			got.Route.Backend != "proxy" ||
			got.Route.Profile != "work" ||
			got.Route.Target != "target-1" ||
			got.Route.BrowserApp != "chrome" ||
			got.CandidateTargetID != "target-2" ||
			!got.RememberTarget ||
			got.Review.PolicyState != "pending" ||
			got.Actor != "browser_tabs" ||
			got.PriorActiveTargetID != "prior-active" ||
			got.PriorRequestedTargetID != "" ||
			got.PriorSelection == nil ||
			got.PriorSelection.ID != "prior-target" ||
			got.ActiveIndex != 2 ||
			len(got.Tabs) != 2 ||
			got.Tabs[1].TargetID != "target-2" ||
			got.Note != "done" ||
			!got.ObservedAt.Equal(observedAt) {
			t.Fatalf("unexpected list tabs route mutation: %#v", got)
		}
	})

	t.Run("projects focus requested-tab candidate", func(t *testing.T) {
		got, ok := BuildSharedSessionBrowserRawTabsRouteMutationObservation(
			"work",
			route,
			&BrowserTabsRequest{Action: " focus ", TabIndex: 3},
			&BrowserTabsResult{
				ActiveIndex: 2,
				Tabs: []BrowserTab{
					{Index: 2, TargetID: "active-target"},
					{Index: 3, TargetID: "requested-target"},
				},
			},
			observedAt,
		)
		if !ok {
			t.Fatal("expected focus tabs route mutation")
		}
		if got.Action != "focus" ||
			got.RequestedTabIndex != 3 ||
			got.CandidateTargetID != "requested-target" {
			t.Fatalf("unexpected focus tabs route mutation: %#v", got)
		}
	})

	t.Run("projects close prior candidate", func(t *testing.T) {
		got, ok := BuildSharedSessionBrowserRawTabsRouteMutationObservation(
			"work",
			route,
			&BrowserTabsRequest{Action: "close", PriorRequestedTargetID: "prior-target"},
			&BrowserTabsResult{Action: "close"},
			observedAt,
		)
		if !ok {
			t.Fatal("expected close tabs route mutation")
		}
		if got.CandidateTargetID != "prior-target" {
			t.Fatalf("expected close candidate from prior requested target, got %#v", got)
		}
	})
}

func TestBuildSharedSessionBrowserRawNavigationRouteMutationObservation(t *testing.T) {
	observedAt := time.Date(2026, time.May, 3, 9, 0, 0, 0, time.UTC)
	route := BrowserSessionRoute{Backend: " proxy ", Profile: " work ", Target: " target-1 "}

	t.Run("rejects observation without URL or title", func(t *testing.T) {
		got, ok := BuildSharedSessionBrowserRawNavigationRouteMutationObservation(
			SharedSessionBrowserRawNavigationObservation{
				RequestedProfile: "work",
				TabIndex:         2,
				Force:            true,
				ObservedAt:       observedAt,
			},
			route,
		)
		if ok {
			t.Fatalf("expected no route mutation without navigation URL/title, got %#v", got)
		}
	})

	t.Run("projects requested URL as final fallback", func(t *testing.T) {
		got, ok := BuildSharedSessionBrowserRawNavigationRouteMutationObservation(
			SharedSessionBrowserRawNavigationObservation{
				RequestedProfile: " work ",
				RequestedURL:     " https://example.com/requested ",
				Title:            " Requested ",
				TabIndex:         2,
				Force:            true,
				ExplicitTargetID: " target-2 ",
				PriorSelection:   &BrowserSessionTargetSelection{ID: "prior-target"},
				Note:             " redirected ",
				ObservedAt:       observedAt,
			},
			route,
		)
		if !ok {
			t.Fatal("expected navigation route mutation")
		}
		if got.Kind != "navigation_result" ||
			got.RequestedProfile != "work" ||
			got.Route.Backend != "proxy" ||
			got.Route.Profile != "work" ||
			got.Route.Target != "target-1" ||
			got.RequestedURL != "https://example.com/requested" ||
			got.FinalURL != "https://example.com/requested" ||
			got.URL != "https://example.com/requested" ||
			got.Title != "Requested" ||
			got.TargetID != "target-2" ||
			got.TabIndex != 2 ||
			!got.Force ||
			got.PriorSelection == nil ||
			got.PriorSelection.ID != "prior-target" ||
			got.Note != "redirected" ||
			got.Source != "runtime_navigation_source" ||
			!got.ObservedAt.Equal(observedAt) {
			t.Fatalf("unexpected navigation route mutation: %#v", got)
		}
	})

	t.Run("projects final URL over requested URL", func(t *testing.T) {
		got, ok := BuildSharedSessionBrowserRawNavigationRouteMutationObservation(
			SharedSessionBrowserRawNavigationObservation{
				RequestedProfile: "work",
				RequestedURL:     "https://example.com/requested",
				FinalURL:         "https://example.com/final",
				ObservedAt:       observedAt,
			},
			route,
		)
		if !ok {
			t.Fatal("expected navigation route mutation")
		}
		if got.URL != "https://example.com/final" || got.FinalURL != "https://example.com/final" {
			t.Fatalf("expected final URL projection, got %#v", got)
		}
	})
}

func TestBuildSharedSessionBrowserRawNavigationObservation(t *testing.T) {
	observedAt := time.Date(2026, time.May, 3, 9, 0, 0, 0, time.UTC)

	t.Run("rejects nil result", func(t *testing.T) {
		got, ok := BuildSharedSessionBrowserRawNavigationObservation("work", nil, nil, observedAt)
		if ok {
			t.Fatalf("expected no raw navigation observation without result, got %#v", got)
		}
	})

	t.Run("projects request and result", func(t *testing.T) {
		req := &BrowserNavigateRequest{
			URL:              " https://example.com/requested ",
			TabIndex:         2,
			Force:            true,
			ExplicitTargetID: " target-2 ",
			PriorSelection:   &BrowserSessionTargetSelection{ID: "prior-target"},
		}
		got, ok := BuildSharedSessionBrowserRawNavigationObservation(
			" work ",
			req,
			&BrowserNavigateResult{
				FinalURL: " https://example.com/final ",
				Title:    " Final ",
				Note:     " redirected ",
			},
			observedAt,
		)
		if !ok {
			t.Fatal("expected raw navigation observation")
		}
		req.PriorSelection.ID = "mutated"
		if got.RequestedProfile != "work" ||
			got.RequestedURL != "https://example.com/requested" ||
			got.FinalURL != "https://example.com/final" ||
			got.Title != "Final" ||
			got.TabIndex != 2 ||
			!got.Force ||
			got.ExplicitTargetID != "target-2" ||
			got.PriorSelection == nil ||
			got.PriorSelection.ID != "prior-target" ||
			got.Note != "redirected" ||
			!got.ObservedAt.Equal(observedAt) {
			t.Fatalf("unexpected raw navigation observation: %#v", got)
		}
	})

	t.Run("uses requested URL as final fallback", func(t *testing.T) {
		got, ok := BuildSharedSessionBrowserRawNavigationObservation(
			"work",
			&BrowserNavigateRequest{URL: "https://example.com/requested"},
			&BrowserNavigateResult{},
			observedAt,
		)
		if !ok {
			t.Fatal("expected raw navigation observation")
		}
		if got.FinalURL != "https://example.com/requested" {
			t.Fatalf("expected final URL fallback, got %#v", got)
		}
	})
}

func TestBuildSharedSessionBrowserRawOpenRouteMutationObservation(t *testing.T) {
	observedAt := time.Date(2026, time.May, 3, 9, 0, 0, 0, time.UTC)
	route := BrowserSessionRoute{Backend: "proxy", Profile: "work", Target: "target-1"}

	t.Run("rejects title without URL", func(t *testing.T) {
		got, ok := BuildSharedSessionBrowserRawOpenRouteMutationObservation(
			SharedSessionBrowserRawOpenObservation{
				RequestedProfile: "work",
				Title:            "Untargeted",
				ObservedAt:       observedAt,
			},
			route,
		)
		if ok {
			t.Fatalf("expected no open route mutation without URL, got %#v", got)
		}
	})

	t.Run("projects URL", func(t *testing.T) {
		got, ok := BuildSharedSessionBrowserRawOpenRouteMutationObservation(
			SharedSessionBrowserRawOpenObservation{
				RequestedProfile: " work ",
				URL:              " https://example.com/open ",
				Title:            " Open ",
				ObservedAt:       observedAt,
			},
			route,
		)
		if !ok {
			t.Fatal("expected open route mutation")
		}
		if got.Kind != "open_result" ||
			got.RequestedProfile != "work" ||
			got.URL != "https://example.com/open" ||
			got.Title != "Open" ||
			got.Source != "runtime_open_source" ||
			!got.ObservedAt.Equal(observedAt) {
			t.Fatalf("unexpected open route mutation: %#v", got)
		}
	})
}

func TestBuildSharedSessionBrowserRawOpenObservation(t *testing.T) {
	observedAt := time.Date(2026, time.May, 3, 9, 0, 0, 0, time.UTC)

	t.Run("rejects nil result", func(t *testing.T) {
		got, ok := BuildSharedSessionBrowserRawOpenObservation("work", "https://example.com", nil, observedAt)
		if ok {
			t.Fatalf("expected no raw open observation without result, got %#v", got)
		}
	})

	t.Run("projects requested URL", func(t *testing.T) {
		got, ok := BuildSharedSessionBrowserRawOpenObservation(
			" work ",
			" https://example.com/open ",
			&BrowserOpenResult{},
			observedAt,
		)
		if !ok {
			t.Fatal("expected raw open observation")
		}
		if got.RequestedProfile != "work" ||
			got.URL != "https://example.com/open" ||
			!got.ObservedAt.Equal(observedAt) {
			t.Fatalf("unexpected raw open observation: %#v", got)
		}
	})
}

func TestNormalizeSharedSessionBrowserRawTargetObservation(t *testing.T) {
	observedAt := time.Date(2026, time.May, 3, 9, 0, 0, 0, time.UTC)
	review := &BrowserSessionTargetReview{
		ID:       " review-1 ",
		URL:      " https://example.com/review ",
		Title:    " Review ",
		Decision: " pending ",
		Reason:   " needs review ",
	}
	observation := SharedSessionBrowserRawTargetObservation{
		RequestedProfile:  " work ",
		TabIndex:          2,
		TrackCurrent:      true,
		SetCurrent:        true,
		URL:               " https://example.com/current ",
		Title:             " Current ",
		Source:            " runtime_click_source ",
		PreferredTargetID: " target-1 ",
		Actor:             " browser_click ",
		Force:             true,
		Review:            SharedSessionBrowserPendingTargetReviewState{Review: review, PolicyState: " review_required ", PolicyReason: " required "},
		ReviewDecision:    " approved ",
		ReviewReady:       true,
		Note:              " done ",
		ObservedAt:        observedAt,
	}

	got := NormalizeSharedSessionBrowserRawTargetObservation(observation, "fallback")
	review.ID = "mutated"

	if got.RequestedProfile != "work" ||
		got.TabIndex != 2 ||
		!got.TrackCurrent ||
		!got.SetCurrent ||
		got.URL != "https://example.com/current" ||
		got.Title != "Current" ||
		got.Source != "runtime_click_source" ||
		got.PreferredTargetID != "target-1" ||
		got.Actor != "browser_click" ||
		!got.Force ||
		got.Review.PolicyState != "review_required" ||
		got.Review.PolicyReason != "required" ||
		got.Review.Review == nil ||
		got.Review.Review.ID != " review-1 " ||
		got.Review.Review.URL != " https://example.com/review " ||
		got.Review.Review.Title != " Review " ||
		got.Review.Review.Decision != " pending " ||
		got.Review.Review.Reason != " needs review " ||
		got.ReviewDecision != "approved" ||
		!got.ReviewReady ||
		got.Note != "done" ||
		!got.ObservedAt.Equal(observedAt) {
		t.Fatalf("unexpected normalized raw target observation: %#v", got)
	}
	if !SharedSessionBrowserRawTargetObservationProvided(got) {
		t.Fatalf("expected normalized raw target observation to be provided: %#v", got)
	}
}

func TestBuildSharedSessionBrowserRawTargetObservation(t *testing.T) {
	observedAt := time.Date(2026, time.May, 3, 9, 0, 0, 0, time.UTC)
	review := SharedSessionBrowserPendingTargetReviewState{
		Review:       &BrowserSessionTargetReview{ID: " review-1 "},
		PolicyState:  " review_required ",
		PolicyReason: " required ",
	}

	t.Run("rejects target observation without tracking intent", func(t *testing.T) {
		got, ok := BuildSharedSessionBrowserRawTargetObservation(
			"work",
			0,
			false,
			false,
			"https://example.com/current",
			"Current",
			"runtime_source",
			"",
			"",
			false,
			SharedSessionBrowserPendingTargetReviewState{},
			"",
			false,
			"",
			observedAt,
		)
		if ok {
			t.Fatalf("expected no raw target observation without tracking intent, got %#v", got)
		}
	})

	t.Run("projects and normalizes target posture", func(t *testing.T) {
		got, ok := BuildSharedSessionBrowserRawTargetObservation(
			" work ",
			2,
			true,
			true,
			" https://example.com/current ",
			" Current ",
			" runtime_click_source ",
			" target-1 ",
			" browser_click ",
			true,
			review,
			" approved ",
			true,
			" done ",
			observedAt,
		)
		if !ok {
			t.Fatal("expected raw target observation")
		}
		review.Review.ID = "mutated"
		if got.RequestedProfile != "work" ||
			got.TabIndex != 2 ||
			!got.TrackCurrent ||
			!got.SetCurrent ||
			got.URL != "https://example.com/current" ||
			got.Title != "Current" ||
			got.Source != "runtime_click_source" ||
			got.PreferredTargetID != "target-1" ||
			got.Actor != "browser_click" ||
			!got.Force ||
			got.Review.PolicyState != "review_required" ||
			got.Review.PolicyReason != "required" ||
			got.Review.Review == nil ||
			got.Review.Review.ID != " review-1 " ||
			got.ReviewDecision != "approved" ||
			!got.ReviewReady ||
			got.Note != "done" ||
			!got.ObservedAt.Equal(observedAt) {
			t.Fatalf("unexpected raw target observation: %#v", got)
		}
	})
}

func TestBuildSharedSessionBrowserRawTargetRouteMutationObservation(t *testing.T) {
	observedAt := time.Date(2026, time.May, 3, 9, 0, 0, 0, time.UTC)
	route := BrowserSessionRoute{
		Backend: " proxy ",
		Profile: " work ",
		Target:  " target-1 ",
	}

	t.Run("rejects non actionable target observation", func(t *testing.T) {
		if got, ok := BuildSharedSessionBrowserRawTargetRouteMutationObservation(
			SharedSessionBrowserRawTargetObservation{URL: "https://example.com", ObservedAt: observedAt},
			route,
		); ok {
			t.Fatalf("expected no route mutation for non-actionable target observation, got %#v", got)
		}
	})

	t.Run("projects page action posture", func(t *testing.T) {
		got, ok := BuildSharedSessionBrowserRawTargetRouteMutationObservation(
			SharedSessionBrowserRawTargetObservation{
				RequestedProfile:  " work ",
				TabIndex:          2,
				SetCurrent:        true,
				URL:               " https://example.com/page ",
				Title:             " Page ",
				Source:            " runtime_target_source ",
				PreferredTargetID: " target-2 ",
				Actor:             " browser_act ",
				Force:             true,
				Review: SharedSessionBrowserPendingTargetReviewState{
					Review: &BrowserSessionTargetReview{ID: " review-1 "},
					Count:  1,
				},
				ObservedAt: observedAt,
			},
			route,
		)
		if !ok {
			t.Fatal("expected page-action route mutation")
		}
		if got.Kind != "page_action_result" ||
			got.RequestedProfile != "work" ||
			got.Route.Backend != "proxy" ||
			got.Route.Profile != "work" ||
			got.Route.Target != "target-1" ||
			got.TargetID != "target-2" ||
			got.TabIndex != 2 ||
			!got.SetCurrent ||
			got.URL != "https://example.com/page" ||
			got.Title != "Page" ||
			got.Source != "runtime_target_source" ||
			got.Actor != "browser_act" ||
			!got.Force ||
			got.Review.Review == nil ||
			got.Review.Review.ID != "review-1" ||
			!got.ObservedAt.Equal(observedAt) {
			t.Fatalf("unexpected page-action route mutation: %#v", got)
		}
	})

	t.Run("projects generic action posture", func(t *testing.T) {
		got, ok := BuildSharedSessionBrowserRawTargetRouteMutationObservation(
			SharedSessionBrowserRawTargetObservation{
				RequestedProfile:  "work",
				TrackCurrent:      true,
				PreferredTargetID: "target-3",
				ReviewDecision:    "approved",
				ReviewReady:       true,
				Note:              "done",
				ObservedAt:        observedAt,
			},
			route,
		)
		if !ok {
			t.Fatal("expected generic action route mutation")
		}
		if got.Kind != "action_result" ||
			got.TargetID != "target-3" ||
			got.Decision != "approved" ||
			!got.Ready ||
			got.Note != "done" {
			t.Fatalf("unexpected generic action route mutation: %#v", got)
		}
	})

	t.Run("projects tab tracking", func(t *testing.T) {
		got, ok := BuildSharedSessionBrowserRawTargetRouteMutationObservation(
			SharedSessionBrowserRawTargetObservation{
				RequestedProfile: "work",
				TabIndex:         3,
				SetCurrent:       true,
				URL:              "https://example.com/tab",
				Title:            "Tab",
				ObservedAt:       observedAt,
			},
			route,
		)
		if !ok {
			t.Fatal("expected tab tracking route mutation")
		}
		if got.Kind != "track_tab" ||
			!got.SetCurrent ||
			got.Tab.Index != 3 ||
			got.Tab.URL != "https://example.com/tab" ||
			got.Tab.Title != "Tab" {
			t.Fatalf("unexpected tab tracking route mutation: %#v", got)
		}
	})

	t.Run("projects current target tracking", func(t *testing.T) {
		got, ok := BuildSharedSessionBrowserRawTargetRouteMutationObservation(
			SharedSessionBrowserRawTargetObservation{
				RequestedProfile: "work",
				TrackCurrent:     true,
				URL:              "https://example.com/current",
				Title:            "Current",
				ObservedAt:       observedAt,
			},
			route,
		)
		if !ok {
			t.Fatal("expected current target route mutation")
		}
		if got.Kind != "track_current" ||
			got.URL != "https://example.com/current" ||
			got.Title != "Current" {
			t.Fatalf("unexpected current target route mutation: %#v", got)
		}
	})
}

func TestObserveSharedSessionBrowserRawStatusReusesSharedZeroConfigWatchManager(t *testing.T) {
	backend := &statusProfilesObservationTestBackend{
		statusResp: BrowserProfileStatusResult{
			Backend: "proxy",
			Profile: "work",
			Status:  "running",
		},
	}

	first := ObserveSharedSessionBrowserRawStatus(context.Background(), backend, "work")
	second := ObserveSharedSessionBrowserRawStatus(context.Background(), backend, "work")

	if first.Status == nil || second.Status == nil {
		t.Fatalf("expected successful raw status observations, got first=%#v second=%#v", first, second)
	}
	if len(backend.statusReqs) != 1 {
		t.Fatalf("expected shared zero-config watch manager to reuse cached raw status, got %#v", backend.statusReqs)
	}
}

func TestObserveSharedSessionBrowserRawProfilesTrimsProfileAndPreservesError(t *testing.T) {
	backend := &statusProfilesObservationTestBackend{profilesErr: errors.New("profiles failed")}
	observation := ObserveSharedSessionBrowserRawProfiles(context.Background(), backend, " work ")
	if observation.RequestedProfile != "work" {
		t.Fatalf("expected trimmed requested profile, got %#v", observation)
	}
	if observation.ObservedAt.IsZero() {
		t.Fatalf("expected raw profiles observation timestamp, got %#v", observation)
	}
	if observation.Err == nil || observation.Err.Error() != "profiles failed" {
		t.Fatalf("expected raw profiles error, got %#v", observation.Err)
	}
	if len(backend.profilesReqs) != 1 || backend.profilesReqs[0].Profile != "work" {
		t.Fatalf("expected raw profiles request for trimmed profile, got %#v", backend.profilesReqs)
	}
}

func TestObserveSharedSessionBrowserRawStatusUsesRawObservationSourceBeforePolling(t *testing.T) {
	observedAt := time.Now().Add(-2 * time.Second)
	backend := &statusProfilesObservationTestBackend{
		rawStatus: func(_ context.Context, requestedProfile string) SharedSessionBrowserRawStatusObservation {
			if requestedProfile != "work" {
				t.Fatalf("expected trimmed requested profile, got %q", requestedProfile)
			}
			return SharedSessionBrowserRawStatusObservation{
				Status: &BrowserProfileStatusResult{
					Backend: "proxy",
					Profile: "work",
					Status:  "running",
				},
				ObservedAt: observedAt,
			}
		},
	}

	observation := ObserveSharedSessionBrowserRawStatus(context.Background(), backend, " work ")
	if observation.RequestedProfile != "work" || observation.Status == nil || observation.Status.Profile != "work" {
		t.Fatalf("expected raw status source observation, got %#v", observation)
	}
	if !observation.ObservedAt.Equal(observedAt) {
		t.Fatalf("expected raw status source timestamp %v, got %v", observedAt, observation.ObservedAt)
	}
	if len(backend.statusReqs) != 0 {
		t.Fatalf("expected raw status source to bypass RuntimeStatus polling, got %#v", backend.statusReqs)
	}
	if backend.rawStatusCalls != 1 {
		t.Fatalf("expected raw status source to be used once, got %d", backend.rawStatusCalls)
	}
}

func TestObserveSharedSessionBrowserRawStatusFallsBackToCombinedSourceAndSeedsProfilesCache(t *testing.T) {
	statusObservedAt := time.Now().Add(-4 * time.Second)
	profilesObservedAt := time.Now().Add(-2 * time.Second)
	backend := &statusProfilesObservationTestBackend{
		rawStatusAndProfiles: func(_ context.Context, requestedProfile string, includeStatus bool, includeProfiles bool) SharedSessionBrowserRawStatusAndProfilesObservation {
			if requestedProfile != "work" || !includeStatus || !includeProfiles {
				t.Fatalf("unexpected combined raw observation request: profile=%q includeStatus=%v includeProfiles=%v", requestedProfile, includeStatus, includeProfiles)
			}
			return SharedSessionBrowserRawStatusAndProfilesObservation{
				Status: &BrowserProfileStatusResult{
					Backend: "proxy",
					Profile: "work",
					Status:  "running",
				},
				StatusObservedAt: statusObservedAt,
				Profiles: &BrowserProfilesResult{
					Backend:        "proxy",
					DefaultProfile: "work",
					Profiles: []BrowserProfileInfo{{
						Profile: "work",
						Status:  "running",
					}},
				},
				ProfilesObservedAt: profilesObservedAt,
			}
		},
	}

	status := ObserveSharedSessionBrowserRawStatus(context.Background(), backend, " work ")
	if status.RequestedProfile != "work" || status.Status == nil || status.Status.Profile != "work" {
		t.Fatalf("expected combined raw source to satisfy raw status observation, got %#v", status)
	}
	if !status.ObservedAt.Equal(statusObservedAt) {
		t.Fatalf("expected combined raw status timestamp %v, got %v", statusObservedAt, status.ObservedAt)
	}
	profiles := ObserveSharedSessionBrowserRawProfiles(context.Background(), backend, "work")
	if profiles.Profiles == nil || profiles.Profiles.DefaultProfile != "work" {
		t.Fatalf("expected combined raw source to seed raw profiles cache, got %#v", profiles)
	}
	if !profiles.ObservedAt.Equal(profilesObservedAt) {
		t.Fatalf("expected combined raw profiles timestamp %v, got %v", profilesObservedAt, profiles.ObservedAt)
	}
	if backend.rawStatusAndProfilesCalls != 1 {
		t.Fatalf("expected combined raw source to be used once, got %d", backend.rawStatusAndProfilesCalls)
	}
	if len(backend.statusReqs) != 0 || len(backend.profilesReqs) != 0 {
		t.Fatalf("expected combined raw source to bypass RuntimeStatus/RuntimeProfiles polling, got status=%#v profiles=%#v", backend.statusReqs, backend.profilesReqs)
	}
}

func TestObserveSharedSessionBrowserRawProfilesFallsBackToCombinedSourceAndSeedsStatusCache(t *testing.T) {
	statusObservedAt := time.Now().Add(-4 * time.Second)
	profilesObservedAt := time.Now().Add(-2 * time.Second)
	backend := &statusProfilesObservationTestBackend{
		rawStatusAndProfiles: func(_ context.Context, requestedProfile string, includeStatus bool, includeProfiles bool) SharedSessionBrowserRawStatusAndProfilesObservation {
			if requestedProfile != "work" || !includeStatus || !includeProfiles {
				t.Fatalf("unexpected combined raw observation request: profile=%q includeStatus=%v includeProfiles=%v", requestedProfile, includeStatus, includeProfiles)
			}
			return SharedSessionBrowserRawStatusAndProfilesObservation{
				Status: &BrowserProfileStatusResult{
					Backend: "proxy",
					Profile: "work",
					Status:  "running",
				},
				StatusObservedAt: statusObservedAt,
				Profiles: &BrowserProfilesResult{
					Backend:        "proxy",
					DefaultProfile: "work",
					Profiles: []BrowserProfileInfo{{
						Profile: "work",
						Status:  "running",
					}},
				},
				ProfilesObservedAt: profilesObservedAt,
			}
		},
	}

	profiles := ObserveSharedSessionBrowserRawProfiles(context.Background(), backend, " work ")
	if profiles.RequestedProfile != "work" || profiles.Profiles == nil || profiles.Profiles.DefaultProfile != "work" {
		t.Fatalf("expected combined raw source to satisfy raw profiles observation, got %#v", profiles)
	}
	if !profiles.ObservedAt.Equal(profilesObservedAt) {
		t.Fatalf("expected combined raw profiles timestamp %v, got %v", profilesObservedAt, profiles.ObservedAt)
	}
	status := ObserveSharedSessionBrowserRawStatus(context.Background(), backend, "work")
	if status.Status == nil || status.Status.Profile != "work" || status.Status.Status != "running" {
		t.Fatalf("expected combined raw source to seed raw status cache, got %#v", status)
	}
	if !status.ObservedAt.Equal(statusObservedAt) {
		t.Fatalf("expected combined raw status timestamp %v, got %v", statusObservedAt, status.ObservedAt)
	}
	if backend.rawStatusAndProfilesCalls != 1 {
		t.Fatalf("expected combined raw source to be used once, got %d", backend.rawStatusAndProfilesCalls)
	}
	if len(backend.statusReqs) != 0 || len(backend.profilesReqs) != 0 {
		t.Fatalf("expected combined raw source to bypass RuntimeStatus/RuntimeProfiles polling, got status=%#v profiles=%#v", backend.statusReqs, backend.profilesReqs)
	}
}

func TestObserveSharedSessionBrowserRawTargetTrimsPageActionPosture(t *testing.T) {
	observedAt := time.Now().Add(-2 * time.Second)
	backend := &statusProfilesObservationTestBackend{
		rawTarget: func(_ context.Context, requestedProfile string) SharedSessionBrowserRawTargetObservation {
			if requestedProfile != "work" {
				t.Fatalf("expected trimmed requested profile, got %q", requestedProfile)
			}
			return SharedSessionBrowserRawTargetObservation{
				RequestedProfile:  "work",
				TabIndex:          2,
				SetCurrent:        true,
				URL:               "https://example.com/popup",
				Title:             "Popup",
				Source:            "runtime_click_source",
				PreferredTargetID: " tab-popup ",
				Actor:             " browser_click ",
				Force:             true,
				Review: SharedSessionBrowserPendingTargetReviewState{
					Review: &BrowserSessionTargetReview{
						ID:       "tab-popup",
						Decision: "session_target_popup_review_required",
					},
					Count:        1,
					PolicyState:  " review_required ",
					PolicyReason: " popup review required ",
				},
				ObservedAt: observedAt,
			}
		},
	}

	observation := ObserveSharedSessionBrowserRawTarget(context.Background(), backend, " work ")
	if observation.RequestedProfile != "work" ||
		observation.PreferredTargetID != "tab-popup" ||
		observation.Actor != "browser_click" ||
		!observation.Force ||
		observation.Review.Review == nil ||
		observation.Review.Review.ID != "tab-popup" ||
		observation.Review.PolicyState != "review_required" ||
		observation.Review.PolicyReason != "popup review required" {
		t.Fatalf("expected raw target source to preserve trimmed page-action posture, got %#v", observation)
	}
	if !observation.ObservedAt.Equal(observedAt) {
		t.Fatalf("expected raw target source timestamp %v, got %v", observedAt, observation.ObservedAt)
	}
	if backend.rawTargetCalls != 1 {
		t.Fatalf("expected raw target source to be consumed once, got %d", backend.rawTargetCalls)
	}
}

func TestObserveSharedSessionBrowserRawTargetTrimsGenericActionPosture(t *testing.T) {
	observedAt := time.Now().Add(-2 * time.Second)
	backend := &statusProfilesObservationTestBackend{
		rawTarget: func(_ context.Context, requestedProfile string) SharedSessionBrowserRawTargetObservation {
			if requestedProfile != "work" {
				t.Fatalf("expected trimmed requested profile, got %q", requestedProfile)
			}
			return SharedSessionBrowserRawTargetObservation{
				RequestedProfile:  "work",
				SetCurrent:        true,
				URL:               "https://example.com/archive.zip",
				Title:             "archive.zip",
				Source:            "runtime_wait_download_source",
				PreferredTargetID: " download-target ",
				ReviewDecision:    " session_target_download_review_confirmed ",
				ReviewReady:       true,
				Note:              " waited for download ",
				ObservedAt:        observedAt,
			}
		},
	}

	observation := ObserveSharedSessionBrowserRawTarget(context.Background(), backend, " work ")
	if observation.RequestedProfile != "work" ||
		observation.PreferredTargetID != "download-target" ||
		observation.ReviewDecision != "session_target_download_review_confirmed" ||
		!observation.ReviewReady ||
		observation.Note != "waited for download" {
		t.Fatalf("expected raw target source to preserve trimmed generic action posture, got %#v", observation)
	}
	if !observation.ObservedAt.Equal(observedAt) {
		t.Fatalf("expected raw target source timestamp %v, got %v", observedAt, observation.ObservedAt)
	}
	if backend.rawTargetCalls != 1 {
		t.Fatalf("expected raw target source to be consumed once, got %d", backend.rawTargetCalls)
	}
}

func TestObserveSharedSessionBrowserRawRouteMutationTrimsAttachState(t *testing.T) {
	observedAt := time.Now().Add(-2 * time.Second)
	backend := &statusProfilesObservationTestBackend{
		rawRouteMutation: func(_ context.Context, requestedProfile string) SharedSessionBrowserRawRouteMutationObservation {
			if requestedProfile != "work" {
				t.Fatalf("expected trimmed requested profile, got %q", requestedProfile)
			}
			return SharedSessionBrowserRawRouteMutationObservation{
				RequestedProfile: "work",
				Route: BrowserSessionRoute{
					Backend:    " proxy ",
					Profile:    " work ",
					Target:     " node ",
					BrowserApp: " Chromium ",
				},
				Kind:        " sync_tabs ",
				ActiveIndex: 2,
				Tabs: []BrowserTab{
					{Index: 1, URL: "https://example.com/home", Title: "Home"},
					{Index: 2, URL: "https://example.com/popup", Title: "Popup"},
				},
				ObservedAt: observedAt,
			}
		},
	}

	observation := ObserveSharedSessionBrowserRawRouteMutation(context.Background(), backend, " work ")
	if observation.RequestedProfile != "work" ||
		observation.Route.Backend != "proxy" ||
		observation.Route.Profile != "work" ||
		observation.Route.Target != "node" ||
		observation.Route.BrowserApp != "chromium" ||
		observation.Kind != "sync_tabs" ||
		observation.ActiveIndex != 2 ||
		len(observation.Tabs) != 2 ||
		observation.Tabs[1].Title != "Popup" {
		t.Fatalf("expected trimmed raw route-mutation source, got %#v", observation)
	}
	if !observation.ObservedAt.Equal(observedAt) {
		t.Fatalf("expected raw route-mutation timestamp %v, got %v", observedAt, observation.ObservedAt)
	}
	if backend.rawRouteMutationCalls != 1 {
		t.Fatalf("expected raw route-mutation source to be consumed once, got %d", backend.rawRouteMutationCalls)
	}
}

func TestObserveSharedSessionBrowserRawStartFallsBackToRequestedProfile(t *testing.T) {
	backend := &executionTestBackend{
		startResp: BrowserProfileStatusResult{
			Backend: "proxy",
			Status:  "started",
		},
	}
	observation := ObserveSharedSessionBrowserRawStart(context.Background(), backend, "isolated")
	if observation.Err != nil {
		t.Fatalf("expected successful raw start observation, got %v", observation.Err)
	}
	if observation.ObservedAt.IsZero() {
		t.Fatalf("expected raw start observation timestamp, got %#v", observation)
	}
	if len(backend.startReqs) != 1 || backend.startReqs[0].Profile != "isolated" {
		t.Fatalf("expected raw start request for isolated profile, got %#v", backend.startReqs)
	}
	if observation.Profile != "isolated" || observation.Status == nil || observation.Status.Status != "started" {
		t.Fatalf("expected raw start observation to preserve requested profile fallback, got %#v", observation)
	}
}

func TestObserveSharedSessionBrowserRawStartUsesRawObservationSourceBeforePolling(t *testing.T) {
	observedAt := time.Now().Add(-3 * time.Second)
	backend := &executionTestBackend{
		rawStart: func(_ context.Context, profile string) SharedSessionBrowserRawLifecycleObservation {
			if profile != "isolated" {
				t.Fatalf("expected requested profile to reach raw start source, got %q", profile)
			}
			return SharedSessionBrowserRawLifecycleObservation{
				Status: &BrowserProfileStatusResult{
					Backend: "proxy",
					Status:  "starting",
				},
				ObservedAt: observedAt,
			}
		},
	}

	observation := ObserveSharedSessionBrowserRawStart(context.Background(), backend, "isolated")
	if observation.Profile != "isolated" || observation.Status == nil || observation.Status.Status != "starting" {
		t.Fatalf("expected raw start source observation, got %#v", observation)
	}
	if !observation.ObservedAt.Equal(observedAt) {
		t.Fatalf("expected raw start source timestamp %v, got %v", observedAt, observation.ObservedAt)
	}
	if len(backend.startReqs) != 0 {
		t.Fatalf("expected raw start source to bypass RuntimeStart polling, got %#v", backend.startReqs)
	}
}

func TestObserveSharedSessionBrowserRawStopFallsBackToRequestedProfile(t *testing.T) {
	backend := &executionTestBackend{
		stopResp: BrowserProfileStatusResult{
			Backend: "proxy",
			Status:  "stopped",
		},
	}
	observation := ObserveSharedSessionBrowserRawStop(context.Background(), backend, "isolated")
	if observation.Err != nil {
		t.Fatalf("expected successful raw stop observation, got %v", observation.Err)
	}
	if observation.ObservedAt.IsZero() {
		t.Fatalf("expected raw stop observation timestamp, got %#v", observation)
	}
	if len(backend.stopReqs) != 1 || backend.stopReqs[0].Profile != "isolated" {
		t.Fatalf("expected raw stop request for isolated profile, got %#v", backend.stopReqs)
	}
	if observation.Profile != "isolated" || observation.Status == nil || observation.Status.Status != "stopped" {
		t.Fatalf("expected raw stop observation to preserve requested profile fallback, got %#v", observation)
	}
}

func TestObserveSharedSessionBrowserRawStopUsesRawObservationSourceBeforePolling(t *testing.T) {
	observedAt := time.Now().Add(-3 * time.Second)
	backend := &executionTestBackend{
		rawStop: func(_ context.Context, profile string) SharedSessionBrowserRawLifecycleObservation {
			if profile != "isolated" {
				t.Fatalf("expected requested profile to reach raw stop source, got %q", profile)
			}
			return SharedSessionBrowserRawLifecycleObservation{
				Status: &BrowserProfileStatusResult{
					Backend: "proxy",
					Status:  "stopped",
				},
				ObservedAt: observedAt,
			}
		},
	}

	observation := ObserveSharedSessionBrowserRawStop(context.Background(), backend, "isolated")
	if observation.Profile != "isolated" || observation.Status == nil || observation.Status.Status != "stopped" {
		t.Fatalf("expected raw stop source observation, got %#v", observation)
	}
	if !observation.ObservedAt.Equal(observedAt) {
		t.Fatalf("expected raw stop source timestamp %v, got %v", observedAt, observation.ObservedAt)
	}
	if len(backend.stopReqs) != 0 {
		t.Fatalf("expected raw stop source to bypass RuntimeStop polling, got %#v", backend.stopReqs)
	}
}

func TestObserveSharedSessionBrowserRawStatusAndProfilesPreservesIndependentErrors(t *testing.T) {
	backend := &statusProfilesObservationTestBackend{
		statusErr: errors.New("status failed"),
		profilesResp: BrowserProfilesResult{
			Backend: "proxy",
			Profiles: []BrowserProfileInfo{{
				Profile: "work",
				Status:  "running",
			}},
		},
	}
	observation := ObserveSharedSessionBrowserRawStatusAndProfiles(context.Background(), backend, " work ", true, true)
	if observation.RequestedProfile != "work" {
		t.Fatalf("expected trimmed requested profile, got %#v", observation)
	}
	if observation.StatusErr == nil || observation.StatusErr.Error() != "status failed" {
		t.Fatalf("expected independent raw status error, got %#v", observation.StatusErr)
	}
	if observation.Status != nil {
		t.Fatalf("expected no raw status payload on status error, got %#v", observation.Status)
	}
	if observation.StatusObservedAt.IsZero() || observation.ProfilesObservedAt.IsZero() {
		t.Fatalf("expected per-source timestamps, got %#v", observation)
	}
	if observation.Profiles == nil || len(observation.Profiles.Profiles) != 1 || observation.Profiles.Profiles[0].Profile != "work" {
		t.Fatalf("expected raw profiles payload to survive independent status error, got %#v", observation.Profiles)
	}
}

func TestObserveSharedSessionBrowserRawStatusAndProfilesFillsMissingProfilesAfterPartialCombinedSource(t *testing.T) {
	observedAt := time.Now().Add(-2 * time.Second)
	backend := &statusProfilesObservationTestBackend{
		rawStatusAndProfiles: func(_ context.Context, requestedProfile string, includeStatus bool, includeProfiles bool) SharedSessionBrowserRawStatusAndProfilesObservation {
			if requestedProfile != "work" || !includeStatus || !includeProfiles {
				t.Fatalf("unexpected combined raw observation request: profile=%q includeStatus=%v includeProfiles=%v", requestedProfile, includeStatus, includeProfiles)
			}
			return SharedSessionBrowserRawStatusAndProfilesObservation{
				Status: &BrowserProfileStatusResult{
					Backend: "proxy",
					Profile: "work",
					Status:  "running",
				},
				StatusObservedAt: observedAt,
			}
		},
		profilesResp: BrowserProfilesResult{
			Backend:        "proxy",
			DefaultProfile: "work",
			Profiles: []BrowserProfileInfo{{
				Profile: "work",
				Status:  "running",
			}},
		},
	}

	observation := ObserveSharedSessionBrowserRawStatusAndProfiles(context.Background(), backend, " work ", true, true)
	if observation.RequestedProfile != "work" {
		t.Fatalf("expected trimmed requested profile, got %#v", observation)
	}
	if observation.Status == nil || observation.Status.Profile != "work" || observation.Status.Status != "running" {
		t.Fatalf("expected combined raw status observation, got %#v", observation.Status)
	}
	if !observation.StatusObservedAt.Equal(observedAt) {
		t.Fatalf("expected combined raw status timestamp %v, got %v", observedAt, observation.StatusObservedAt)
	}
	if observation.Profiles == nil || observation.Profiles.DefaultProfile != "work" {
		t.Fatalf("expected missing profiles to fall back to RuntimeProfiles polling, got %#v", observation.Profiles)
	}
	if observation.ProfilesObservedAt.IsZero() {
		t.Fatalf("expected fallback profiles observation timestamp, got %#v", observation)
	}
	if backend.rawStatusAndProfilesCalls != 1 {
		t.Fatalf("expected one combined raw observation call, got %d", backend.rawStatusAndProfilesCalls)
	}
	if len(backend.statusReqs) != 0 {
		t.Fatalf("expected combined raw status observation to bypass RuntimeStatus polling, got %#v", backend.statusReqs)
	}
	if len(backend.profilesReqs) != 1 || backend.profilesReqs[0].Profile != "work" {
		t.Fatalf("expected RuntimeProfiles polling to fill missing combined observation side, got %#v", backend.profilesReqs)
	}
}

func TestObserveSharedSessionBrowserRawStatusAndProfilesUsesCombinedRawObservationSourceAndSeedsCaches(t *testing.T) {
	statusObservedAt := time.Now().Add(-4 * time.Second)
	profilesObservedAt := time.Now().Add(-2 * time.Second)
	backend := &statusProfilesObservationTestBackend{
		rawStatusAndProfiles: func(_ context.Context, requestedProfile string, includeStatus bool, includeProfiles bool) SharedSessionBrowserRawStatusAndProfilesObservation {
			if requestedProfile != "work" || !includeStatus || !includeProfiles {
				t.Fatalf("unexpected combined raw observation request: profile=%q includeStatus=%v includeProfiles=%v", requestedProfile, includeStatus, includeProfiles)
			}
			return SharedSessionBrowserRawStatusAndProfilesObservation{
				Status: &BrowserProfileStatusResult{
					Backend: "proxy",
					Profile: "work",
					Status:  "running",
				},
				StatusObservedAt: statusObservedAt,
				Profiles: &BrowserProfilesResult{
					Backend:        "proxy",
					DefaultProfile: "work",
					Profiles: []BrowserProfileInfo{{
						Profile: "work",
						Status:  "running",
					}},
				},
				ProfilesObservedAt: profilesObservedAt,
			}
		},
	}

	combined := ObserveSharedSessionBrowserRawStatusAndProfiles(context.Background(), backend, " work ", true, true)
	if combined.RequestedProfile != "work" || combined.Status == nil || combined.Profiles == nil {
		t.Fatalf("expected combined raw source observation, got %#v", combined)
	}
	if !combined.StatusObservedAt.Equal(statusObservedAt) || !combined.ProfilesObservedAt.Equal(profilesObservedAt) {
		t.Fatalf("expected combined raw source timestamps to be preserved, got %#v", combined)
	}
	status := ObserveSharedSessionBrowserRawStatus(context.Background(), backend, "work")
	profiles := ObserveSharedSessionBrowserRawProfiles(context.Background(), backend, "work")
	if status.Status == nil || profiles.Profiles == nil {
		t.Fatalf("expected per-source caches to be seeded from combined raw observation, got status=%#v profiles=%#v", status, profiles)
	}
	if len(backend.statusReqs) != 0 || len(backend.profilesReqs) != 0 {
		t.Fatalf("expected combined raw source to avoid RuntimeStatus/RuntimeProfiles polling, got status=%#v profiles=%#v", backend.statusReqs, backend.profilesReqs)
	}
	if backend.rawStatusAndProfilesCalls != 1 {
		t.Fatalf("expected combined raw source to be used once, got %d", backend.rawStatusAndProfilesCalls)
	}
}

func TestObserveSharedSessionBrowserRawTabsUsesRawObservationSource(t *testing.T) {
	observedAt := time.Now().Add(-1500 * time.Millisecond)
	backend := &statusProfilesObservationTestBackend{
		rawTabs: func(_ context.Context, requestedProfile string) SharedSessionBrowserRawTabsObservation {
			if requestedProfile != "workbench" {
				t.Fatalf("expected trimmed requested profile, got %q", requestedProfile)
			}
			return SharedSessionBrowserRawTabsObservation{
				RequestedProfile:  "workbench",
				Action:            "focus",
				RequestedTabIndex: 2,
				ActiveIndex:       2,
				Tabs: []BrowserTab{
					{Index: 1, URL: "https://example.com/one", Title: "One", Active: false},
					{Index: 2, URL: "https://example.com/two", Title: "Two", Active: true},
				},
				ObservedAt: observedAt,
			}
		},
	}

	observation := ObserveSharedSessionBrowserRawTabs(context.Background(), backend, " workbench ")
	if observation.RequestedProfile != "workbench" {
		t.Fatalf("expected trimmed requested profile, got %#v", observation)
	}
	if observation.Action != "focus" || observation.RequestedTabIndex != 2 || observation.ActiveIndex != 2 || len(observation.Tabs) != 2 || observation.Tabs[1].Title != "Two" {
		t.Fatalf("expected raw tabs source payload, got %#v", observation)
	}
	if !observation.ObservedAt.Equal(observedAt) {
		t.Fatalf("expected raw tabs source timestamp %v, got %v", observedAt, observation.ObservedAt)
	}
	if backend.rawTabsCalls != 1 {
		t.Fatalf("expected raw tabs source to be used once, got %d", backend.rawTabsCalls)
	}
}

func TestObserveSharedSessionBrowserRawNavigationUsesRawObservationSource(t *testing.T) {
	observedAt := time.Now().Add(-1200 * time.Millisecond)
	backend := &statusProfilesObservationTestBackend{
		rawNavigation: func(_ context.Context, requestedProfile string) SharedSessionBrowserRawNavigationObservation {
			if requestedProfile != "workbench" {
				t.Fatalf("expected trimmed requested profile, got %q", requestedProfile)
			}
			return SharedSessionBrowserRawNavigationObservation{
				RequestedProfile: "workbench",
				RequestedURL:     "https://example.com/start",
				FinalURL:         "https://example.org/landing",
				Title:            "Landing",
				TabIndex:         2,
				Force:            true,
				ExplicitTargetID: "tab-start",
				PriorSelection:   &BrowserSessionTargetSelection{ID: "tab-home", Source: "tracked_current_target"},
				Note:             "redirect ok",
				ObservedAt:       observedAt,
			}
		},
	}

	observation := ObserveSharedSessionBrowserRawNavigation(context.Background(), backend, " workbench ")
	if observation.RequestedProfile != "workbench" {
		t.Fatalf("expected trimmed requested profile, got %#v", observation)
	}
	if observation.RequestedURL != "https://example.com/start" ||
		observation.FinalURL != "https://example.org/landing" ||
		observation.Title != "Landing" ||
		observation.TabIndex != 2 ||
		!observation.Force ||
		observation.ExplicitTargetID != "tab-start" ||
		observation.PriorSelection == nil ||
		observation.PriorSelection.ID != "tab-home" ||
		observation.Note != "redirect ok" {
		t.Fatalf("expected raw navigation source payload, got %#v", observation)
	}
	if !observation.ObservedAt.Equal(observedAt) {
		t.Fatalf("expected raw navigation source timestamp %v, got %v", observedAt, observation.ObservedAt)
	}
	if backend.rawNavigationCalls != 1 {
		t.Fatalf("expected raw navigation source to be used once, got %d", backend.rawNavigationCalls)
	}
}

func TestObserveSharedSessionBrowserRawOpenUsesRawObservationSource(t *testing.T) {
	observedAt := time.Now().Add(-900 * time.Millisecond)
	backend := &statusProfilesObservationTestBackend{
		rawOpen: func(_ context.Context, requestedProfile string) SharedSessionBrowserRawOpenObservation {
			if requestedProfile != "workbench" {
				t.Fatalf("expected trimmed requested profile, got %q", requestedProfile)
			}
			return SharedSessionBrowserRawOpenObservation{
				RequestedProfile: "workbench",
				URL:              "https://example.com/opened",
				Title:            "Opened",
				ObservedAt:       observedAt,
			}
		},
	}

	observation := ObserveSharedSessionBrowserRawOpen(context.Background(), backend, " workbench ")
	if observation.RequestedProfile != "workbench" {
		t.Fatalf("expected trimmed requested profile, got %#v", observation)
	}
	if observation.URL != "https://example.com/opened" || observation.Title != "Opened" {
		t.Fatalf("expected raw open source payload, got %#v", observation)
	}
	if !observation.ObservedAt.Equal(observedAt) {
		t.Fatalf("expected raw open source timestamp %v, got %v", observedAt, observation.ObservedAt)
	}
	if backend.rawOpenCalls != 1 {
		t.Fatalf("expected raw open source to be used once, got %d", backend.rawOpenCalls)
	}
}

func TestObserveSharedSessionBrowserRawTargetUsesRawObservationSource(t *testing.T) {
	observedAt := time.Now().Add(-time.Second)
	backend := &statusProfilesObservationTestBackend{
		rawTarget: func(_ context.Context, requestedProfile string) SharedSessionBrowserRawTargetObservation {
			if requestedProfile != "workbench" {
				t.Fatalf("expected trimmed requested profile, got %q", requestedProfile)
			}
			return SharedSessionBrowserRawTargetObservation{
				RequestedProfile: "workbench",
				TabIndex:         2,
				SetCurrent:       true,
				URL:              "https://example.com/console",
				Title:            "Console",
				Source:           "runtime_console_source",
				ObservedAt:       observedAt,
			}
		},
	}

	observation := ObserveSharedSessionBrowserRawTarget(context.Background(), backend, "  workbench  ")
	if observation.RequestedProfile != "workbench" {
		t.Fatalf("expected normalized requested profile, got %#v", observation)
	}
	if observation.TabIndex != 2 || !observation.SetCurrent || observation.URL != "https://example.com/console" || observation.Title != "Console" || observation.Source != "runtime_console_source" {
		t.Fatalf("expected raw target source payload, got %#v", observation)
	}
	if !observation.ObservedAt.Equal(observedAt) {
		t.Fatalf("expected raw target source timestamp %v, got %v", observedAt, observation.ObservedAt)
	}
	if backend.rawTargetCalls != 1 {
		t.Fatalf("expected raw target source to be used once, got %d", backend.rawTargetCalls)
	}
}
