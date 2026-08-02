package browserruntime

import "testing"

func TestProjectSharedSessionBrowserBindingRouteProjectionKeepsRouteAndConfiguredPriority(t *testing.T) {
	projection := ProjectSharedSessionBrowserBindingRouteProjection(
		SharedSessionBrowserBindingEvaluation{
			Snapshot: SharedSessionBrowserBindingSnapshot{
				Profiles: []SharedSessionBrowserProfileState{
					{
						Backend:       "proxy",
						Profile:       "workbench",
						RuntimeTarget: "node",
						BrowserApp:    "Chromium",
						Status:        "running",
						Running:       true,
						Connected:     true,
					},
					{
						Backend:       "sandbox",
						Profile:       "isolated",
						RuntimeTarget: "sandbox",
						BrowserApp:    "Firefox",
						Status:        "running",
						Running:       true,
						Connected:     true,
					},
				},
			},
		},
		&SharedSessionBrowserSelectionProjection{
			ProfileSelection: &SharedSessionBrowserProfileSelection{
				Backend:       "proxy",
				Profile:       "workbench",
				RuntimeTarget: "node",
				Source:        "select_profile",
			},
			TargetSelection: &BrowserSessionTargetSelection{
				ID:            "tab-2",
				Backend:       "sandbox",
				Profile:       "isolated",
				RuntimeTarget: "sandbox",
				Source:        "tracked_active_tab",
			},
		},
		"",
	)

	if projection.SelectedRouteInfo != (BrowserRuntimeInfo{Backend: "sandbox", Profile: "isolated", Target: "sandbox"}) {
		t.Fatalf("expected route projection to prefer target selection identity, got %#v", projection.SelectedRouteInfo)
	}
	if projection.ConfiguredInfo != (BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}) {
		t.Fatalf("expected configured projection to prefer profile selection identity, got %#v", projection.ConfiguredInfo)
	}
	if projection.DefaultProfile != "workbench" {
		t.Fatalf("expected default profile to prefer configured profile selection, got %#v", projection)
	}
}

func TestProjectSharedSessionBrowserBindingRouteProjectionFallsBackToProfileSnapshot(t *testing.T) {
	projection := ProjectSharedSessionBrowserBindingRouteProjection(
		SharedSessionBrowserBindingEvaluation{
			Snapshot: SharedSessionBrowserBindingSnapshot{
				Profiles: []SharedSessionBrowserProfileState{{
					Backend:       "proxy",
					Profile:       "workbench",
					RuntimeTarget: "node",
					BrowserApp:    "Chromium",
					Status:        "running",
					Running:       true,
					Connected:     true,
				}},
			},
		},
		nil,
		"",
	)

	if projection.SelectedRouteInfo != (BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}) {
		t.Fatalf("expected route projection to fall back to profile snapshot, got %#v", projection.SelectedRouteInfo)
	}
	if projection.ConfiguredInfo != (BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}) {
		t.Fatalf("expected configured projection to fall back to profile snapshot, got %#v", projection.ConfiguredInfo)
	}
	if projection.DefaultProfile != "workbench" {
		t.Fatalf("expected default profile to fall back to projected profile inventory, got %#v", projection)
	}
}
