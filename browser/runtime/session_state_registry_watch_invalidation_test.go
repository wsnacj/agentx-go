package browserruntime

import (
	"context"
	"testing"
	"time"
)

func TestBrowserSessionStateRegistrySelectBrowserProfileInvalidatesSharedWatchManagerCaches(t *testing.T) {
	stateRegistry := NewBrowserSessionStateRegistry()
	backend := &statusProfilesObservationTestBackend{
		profilesResp: BrowserProfilesResult{
			Backend:        "proxy",
			DefaultProfile: "isolated",
			Profiles: []BrowserProfileInfo{
				{Profile: "isolated", BrowserApp: "Chromium", Status: "connected", Running: true, Connected: true},
				{Profile: "work", BrowserApp: "Chromium", Status: "connected", Running: true, Connected: true},
			},
		},
	}
	manager := SharedSessionBrowserObserverManagerFor(nil, nil, stateRegistry, time.Minute)
	sessionID := "sess-state-select-invalidates-watch"
	selectedInfo := BrowserRuntimeInfo{Backend: "proxy", Target: "node"}

	stateRegistry.SelectBrowserProfile(sessionID, SharedSessionBrowserProfileSelection{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Source:        "remember_profile",
	})
	first := manager.ObserveProfiles(context.Background(), backend, sessionID, selectedInfo, "")
	if !sharedSessionBrowserProjectedProfileSelected(first.Projected, "isolated") {
		t.Fatalf("expected initial projected profiles to select isolated, got %#v", first.Projected)
	}
	if len(backend.profilesReqs) != 1 {
		t.Fatalf("expected initial profile observation to poll backend once, got %d", len(backend.profilesReqs))
	}

	stateRegistry.SelectBrowserProfile(sessionID, SharedSessionBrowserProfileSelection{
		Backend:       "proxy",
		Profile:       "work",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Source:        "remember_profile",
	})
	second := manager.ObserveProfiles(context.Background(), backend, sessionID, selectedInfo, "")
	if !sharedSessionBrowserProjectedProfileSelected(second.Projected, "work") {
		t.Fatalf("expected selection change to invalidate cached profiles watch, got %#v", second.Projected)
	}
	if len(backend.profilesReqs) != 1 {
		t.Fatalf("expected selection invalidation to reuse cached raw profiles, got %d", len(backend.profilesReqs))
	}
}

func TestBrowserSessionStateRegistryClearSelectedBrowserProfileInvalidatesSharedWatchManagerCaches(t *testing.T) {
	stateRegistry := NewBrowserSessionStateRegistry()
	backend := &statusProfilesObservationTestBackend{
		profilesResp: BrowserProfilesResult{
			Backend:        "proxy",
			DefaultProfile: "isolated",
			Profiles: []BrowserProfileInfo{
				{Profile: "isolated", BrowserApp: "Chromium", Status: "connected", Running: true, Connected: true},
				{Profile: "work", BrowserApp: "Chromium", Status: "connected", Running: true, Connected: true},
			},
		},
	}
	manager := SharedSessionBrowserObserverManagerFor(nil, nil, stateRegistry, time.Minute)
	sessionID := "sess-state-clear-select-invalidates-watch"
	selectedInfo := BrowserRuntimeInfo{Backend: "proxy", Target: "node"}

	stateRegistry.SelectBrowserProfile(sessionID, SharedSessionBrowserProfileSelection{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Source:        "remember_profile",
	})
	first := manager.ObserveProfiles(context.Background(), backend, sessionID, selectedInfo, "")
	if !sharedSessionBrowserProjectedProfileSelected(first.Projected, "isolated") {
		t.Fatalf("expected initial projected profiles to select isolated, got %#v", first.Projected)
	}
	if len(backend.profilesReqs) != 1 {
		t.Fatalf("expected initial profile observation to poll backend once, got %d", len(backend.profilesReqs))
	}

	stateRegistry.ClearSelectedBrowserProfile(sessionID, "node")
	second := manager.ObserveProfiles(context.Background(), backend, sessionID, selectedInfo, "")
	if sharedSessionBrowserAnyProjectedProfileSelected(second.Projected) {
		t.Fatalf("expected clear-selected write to invalidate cached profiles watch, got %#v", second.Projected)
	}
	if len(backend.profilesReqs) != 1 {
		t.Fatalf("expected clear-selected invalidation to reuse cached raw profiles, got %d", len(backend.profilesReqs))
	}
}

func sharedSessionBrowserProjectedProfileSelected(items []SharedSessionBrowserProjectedProfileState, profile string) bool {
	for _, item := range items {
		if item.Selected && item.State.Profile == profile {
			return true
		}
	}
	return false
}

func sharedSessionBrowserAnyProjectedProfileSelected(items []SharedSessionBrowserProjectedProfileState) bool {
	for _, item := range items {
		if item.Selected {
			return true
		}
	}
	return false
}
