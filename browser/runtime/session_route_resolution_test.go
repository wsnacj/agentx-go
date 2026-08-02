package browserruntime

import "testing"

func TestResolveSharedSessionBrowserRouteResolutionPrefersExplicitAndTrackedSources(t *testing.T) {
	resolution, ok := ResolveSharedSessionBrowserRouteResolution(
		"workbench",
		"node",
		BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
		BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
		&SharedSessionBrowserProfileSelection{
			Profile:       "workbench",
			RuntimeTarget: "node",
			Source:        "select_profile",
		},
		&BrowserSessionTargetSelection{Source: "tracked_active_tab"},
	)
	if !ok {
		t.Fatalf("expected route resolution to be present")
	}
	if resolution.ProfileSource != "explicit_request" || resolution.RuntimeTargetSource != "explicit_request" || resolution.TargetSource != "tracked_active_tab" {
		t.Fatalf("unexpected route resolution: %#v", resolution)
	}
}

func TestResolveSharedSessionBrowserRouteResolutionUsesSelectionAndDefaultFallbacks(t *testing.T) {
	resolution, ok := ResolveSharedSessionBrowserRouteResolution(
		"",
		"",
		BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
		BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
		&SharedSessionBrowserProfileSelection{
			Profile:       "workbench",
			RuntimeTarget: "node",
			Source:        "select_profile",
		},
		nil,
	)
	if !ok {
		t.Fatalf("expected route resolution to be present")
	}
	if resolution.ProfileSource != "select_profile" || resolution.RuntimeTargetSource != "select_profile" || resolution.TargetSource != "" {
		t.Fatalf("unexpected selection-based route resolution: %#v", resolution)
	}

	resolution, ok = ResolveSharedSessionBrowserRouteResolution(
		"",
		"",
		BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
		BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
		nil,
		nil,
	)
	if !ok {
		t.Fatalf("expected default route resolution to be present")
	}
	if resolution.ProfileSource != "default_route" || resolution.RuntimeTargetSource != "default_route" || resolution.TargetSource != "" {
		t.Fatalf("unexpected default route resolution: %#v", resolution)
	}
}

func TestSharedSessionBrowserProfileStateSelectedUsesFallbackRuntimeTarget(t *testing.T) {
	selected := SharedSessionBrowserProfileStateSelected(
		&SharedSessionBrowserProfileSelection{
			Backend:       "proxy",
			Profile:       "workbench",
			RuntimeTarget: "node",
		},
		"node",
		SharedSessionBrowserProfileState{
			Backend: "proxy-cdp",
			Profile: "workbench",
		},
	)
	if !selected {
		t.Fatalf("expected fallback runtime target to preserve selected-state match")
	}

	selected = SharedSessionBrowserProfileStateSelected(
		&SharedSessionBrowserProfileSelection{
			Backend:       "proxy",
			Profile:       "workbench",
			RuntimeTarget: "node",
		},
		"node",
		SharedSessionBrowserProfileState{
			Backend: "proxy-cdp",
			Profile: "isolated",
		},
	)
	if selected {
		t.Fatalf("expected mismatched profile not to be marked selected")
	}
}
