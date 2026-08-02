package browserruntime

import "testing"

func TestProjectSharedSessionBrowserTopLevelProfileInventoryFromBindingEvaluationUsesBindingSnapshot(t *testing.T) {
	projection := ProjectSharedSessionBrowserTopLevelProfileInventoryFromBindingEvaluation(
		SharedSessionBrowserBindingProfileInventoryProjectionRequest{
			Evaluation: SharedSessionBrowserBindingEvaluation{
				Snapshot: SharedSessionBrowserBindingSnapshot{
					SelectedProfileSelection: &SharedSessionBrowserProfileSelection{
						Backend:       "proxy",
						Profile:       "workbench",
						RuntimeTarget: "node",
						BrowserApp:    "Chromium",
						Source:        "select_profile",
					},
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
					},
				},
			},
			SelectedInfo:            BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
			RequestedDefaultProfile: "",
			NeedProfileStatus:       true,
			NeedProfileInventory:    true,
		},
	)

	if projection == nil {
		t.Fatalf("expected binding profile inventory projection")
	}
	if !projection.HasProfileStatus || projection.ProfileStatus.Profile != "workbench" || projection.ProfileStatus.RuntimeTarget != "node" {
		t.Fatalf("expected binding projection to recover profile status, got %#v", projection)
	}
	if !projection.ApplyProfileInventory || len(projection.Profiles) != 1 || projection.Profiles[0].State.Profile != "workbench" {
		t.Fatalf("expected binding projection to recover profile inventory, got %#v", projection)
	}
	if projection.DefaultProfile != "workbench" {
		t.Fatalf("expected binding projection to recover default profile, got %#v", projection)
	}
}

func TestProjectSharedSessionBrowserTopLevelProfileInventoryFromBindingEvaluationFallsBackToBindingSelectedInfo(t *testing.T) {
	projection := ProjectSharedSessionBrowserTopLevelProfileInventoryFromBindingEvaluation(
		SharedSessionBrowserBindingProfileInventoryProjectionRequest{
			Evaluation: SharedSessionBrowserBindingEvaluation{
				Snapshot: SharedSessionBrowserBindingSnapshot{
					SelectedTargetSelection: &BrowserSessionTargetSelection{
						ID:            "tab-1",
						Backend:       "proxy",
						Profile:       "workbench",
						RuntimeTarget: "node",
						BrowserApp:    "Chromium",
						Source:        "tracked_active_tab",
					},
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
							Backend:       "proxy",
							Profile:       "relay",
							RuntimeTarget: "node",
							Status:        "stopped",
						},
					},
				},
			},
			NeedProfileStatus:    true,
			NeedProfileInventory: true,
		},
	)

	if projection == nil {
		t.Fatalf("expected binding fallback profile inventory projection")
	}
	if !projection.HasProfileStatus || projection.ProfileStatus.Profile != "workbench" {
		t.Fatalf("expected binding fallback to use selected target profile, got %#v", projection)
	}
	if !projection.ApplyProfileInventory || len(projection.Profiles) != 2 {
		t.Fatalf("expected binding fallback to keep all projected profiles, got %#v", projection)
	}
	if projection.DefaultProfile != "workbench" {
		t.Fatalf("expected binding fallback default profile to come from selection, got %#v", projection)
	}
}
