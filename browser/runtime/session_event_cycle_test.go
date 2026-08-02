package browserruntime

import (
	"context"
	"testing"
	"time"
)

func TestObserveSharedSessionBrowserEventCycleUsesRegistryResolutionAndReferenceTime(t *testing.T) {
	backend := &statusProfilesObservationTestBackend{
		statusResp: BrowserProfileStatusResult{
			Profile:   "isolated",
			Status:    "running",
			Running:   true,
			Connected: true,
		},
		profilesResp: BrowserProfilesResult{
			Backend:        "proxy",
			DefaultProfile: "isolated",
			Profiles: []BrowserProfileInfo{
				{Profile: "isolated", BrowserApp: "Chromium", Status: "connected", Running: true, Connected: true},
			},
		},
	}
	registry := NewBrowserSessionStateRegistry()

	cycle := ObserveSharedSessionBrowserEventCycle(
		context.Background(),
		backend,
		registry,
		"sess-cycle",
		BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		"isolated",
		true,
		true,
		time.Minute,
	)

	if cycle.Observation.Status == nil || cycle.Observation.Profiles == nil {
		t.Fatalf("expected raw status and profiles observation, got %#v", cycle.Observation)
	}
	if !cycle.Observation.HasSyncedState {
		t.Fatalf("expected synced state from shared event cycle, got %#v", cycle.Observation)
	}
	if len(cycle.Observation.Snapshot) != 1 {
		t.Fatalf("expected scoped synced snapshot, got %#v", cycle.Observation.Snapshot)
	}
	expectedReference := cycle.Observation.StatusObservedAt
	if cycle.Observation.ProfilesObservedAt.After(expectedReference) {
		expectedReference = cycle.Observation.ProfilesObservedAt
	}
	if !cycle.ReferenceTime.Equal(expectedReference) {
		t.Fatalf("expected event cycle reference time %v, got %v", expectedReference, cycle.ReferenceTime)
	}
}

func TestObserveSharedSessionBrowserEventCycleStatusOnlyUsesStatusObservedAt(t *testing.T) {
	backend := &statusProfilesObservationTestBackend{
		statusResp: BrowserProfileStatusResult{
			Profile: "isolated",
			Status:  "running",
		},
	}

	cycle := ObserveSharedSessionBrowserEventCycle(
		context.Background(),
		backend,
		nil,
		"",
		BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		"isolated",
		true,
		false,
		time.Minute,
	)

	if cycle.Observation.Status == nil {
		t.Fatalf("expected status observation")
	}
	if cycle.ReferenceTime.IsZero() || !cycle.ReferenceTime.Equal(cycle.Observation.StatusObservedAt) {
		t.Fatalf("expected status-only cycle to use status observed_at as reference time, got %v want %v", cycle.ReferenceTime, cycle.Observation.StatusObservedAt)
	}
}
