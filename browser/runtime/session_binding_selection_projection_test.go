package browserruntime

import "testing"

func TestProjectSharedSessionBrowserProfileSelectionFromBindingSnapshotHydratesFromProfileSnapshot(t *testing.T) {
	projection := ProjectSharedSessionBrowserProfileSelectionFromBindingSnapshot(
		&SharedSessionBrowserProfileSelection{
			Profile: "workbench",
			Source:  "select_profile",
		},
		&BrowserSessionTargetSelection{
			RuntimeTarget: "node",
		},
		[]SharedSessionBrowserProfileState{{
			Backend:       "proxy",
			Profile:       "workbench",
			RuntimeTarget: "node",
			BrowserApp:    "Chromium",
			Status:        "running",
			Running:       true,
			Connected:     true,
		}},
	)

	if projection == nil ||
		projection.Backend != "proxy" ||
		projection.Profile != "workbench" ||
		projection.RuntimeTarget != "node" ||
		projection.BrowserApp != "Chromium" ||
		projection.Source != "select_profile" {
		t.Fatalf("expected hydrated profile selection projection, got %#v", projection)
	}
}

func TestProjectSharedSessionBrowserTargetSelectionFromBindingSnapshotKeepsStoredIdentity(t *testing.T) {
	projection := ProjectSharedSessionBrowserTargetSelectionFromBindingSnapshot(
		&BrowserSessionTargetSelection{
			ID:            "tab-2",
			TabIndex:      2,
			Backend:       "proxy",
			Profile:       "workbench",
			RuntimeTarget: "node",
			BrowserApp:    "Chromium",
			Source:        "tracked_active_tab",
		},
		&SharedSessionBrowserProfileSelection{
			Backend:       "proxy",
			Profile:       "workbench",
			RuntimeTarget: "node",
		},
		[]SharedSessionBrowserProfileState{{
			Backend:       "sandbox",
			Profile:       "workbench",
			RuntimeTarget: "node",
			BrowserApp:    "Firefox",
			Status:        "running",
			Running:       true,
			Connected:     true,
		}},
	)

	if projection == nil ||
		projection.ID != "tab-2" ||
		projection.TabIndex != 2 ||
		projection.Backend != "proxy" ||
		projection.Profile != "workbench" ||
		projection.RuntimeTarget != "node" ||
		projection.BrowserApp != "Chromium" ||
		projection.Source != "tracked_active_tab" {
		t.Fatalf("expected stored target selection identity to win over mismatched snapshot, got %#v", projection)
	}
}

func TestProjectSharedSessionBrowserSelectionProjectionFromBindingEvaluationHydratesBrowserApp(t *testing.T) {
	projection := ProjectSharedSessionBrowserSelectionProjectionFromBindingEvaluation(
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
		&SharedSessionBrowserSelectionProjection{
			ProfileSelection: &SharedSessionBrowserProfileSelection{
				Backend:       "proxy",
				Profile:       "workbench",
				RuntimeTarget: "node",
				Source:        "select_profile",
			},
			TargetSelection: &BrowserSessionTargetSelection{
				ID:            "tab-2",
				Backend:       "proxy",
				Profile:       "workbench",
				RuntimeTarget: "node",
				Source:        "tracked_active_tab",
			},
			ApplyTargetToRoute: true,
		},
	)

	if projection == nil ||
		projection.ProfileSelection == nil ||
		projection.ProfileSelection.BrowserApp != "Chromium" ||
		projection.TargetSelection == nil ||
		projection.TargetSelection.BrowserApp != "Chromium" ||
		!projection.ApplyTargetToRoute {
		t.Fatalf("expected hydrated shared selection projection from binding evaluation, got %#v", projection)
	}
}

func TestProjectSharedSessionBrowserSelectionProjectionFromBindingEvaluationPreservesStoredIdentity(t *testing.T) {
	projection := ProjectSharedSessionBrowserSelectionProjectionFromBindingEvaluation(
		SharedSessionBrowserBindingEvaluation{
			Snapshot: SharedSessionBrowserBindingSnapshot{
				SelectedProfileSelection: &SharedSessionBrowserProfileSelection{
					Backend:       "proxy",
					Profile:       "workbench",
					RuntimeTarget: "node",
					BrowserApp:    "Chromium",
					Source:        "remember_profile",
				},
				SelectedTargetSelection: &BrowserSessionTargetSelection{
					ID:            "tab-1",
					TabIndex:      1,
					Backend:       "proxy",
					Profile:       "workbench",
					RuntimeTarget: "node",
					BrowserApp:    "Chromium",
					Source:        "tracked_active_tab",
				},
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
		&SharedSessionBrowserSelectionProjection{
			ProfileSelection: &SharedSessionBrowserProfileSelection{
				BrowserApp: "Firefox",
				Source:     "other",
			},
			TargetSelection: &BrowserSessionTargetSelection{
				ID:         "tab-2",
				BrowserApp: "Firefox",
				Source:     "other",
			},
		},
	)

	if projection == nil ||
		projection.ProfileSelection == nil ||
		projection.ProfileSelection.BrowserApp != "Chromium" ||
		projection.ProfileSelection.Source != "remember_profile" ||
		projection.TargetSelection == nil ||
		projection.TargetSelection.ID != "tab-1" ||
		projection.TargetSelection.BrowserApp != "Chromium" ||
		projection.TargetSelection.Source != "tracked_active_tab" {
		t.Fatalf("expected stored selection identity to win in shared projection, got %#v", projection)
	}
}
