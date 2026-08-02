package browserruntime

import "testing"

func TestProjectSharedSessionBrowserBindingEvaluationWithSelectionProjectionOverlaysCurrentTarget(t *testing.T) {
	evaluation := SharedSessionBrowserBindingEvaluation{
		Snapshot: SharedSessionBrowserBindingSnapshot{
			SelectedProfileSelection: &SharedSessionBrowserProfileSelection{
				Backend: "proxy",
				Profile: "workbench",
				Source:  "remember_profile",
			},
			Summary: SharedSessionBrowserBindingSummary{},
		},
	}

	projected := ProjectSharedSessionBrowserBindingEvaluationWithSelectionProjection(
		evaluation,
		&SharedSessionBrowserSelectionProjection{
			ProfileSelection: &SharedSessionBrowserProfileSelection{
				RuntimeTarget: "node",
				BrowserApp:    "Chromium",
			},
			TargetSelection: &BrowserSessionTargetSelection{
				ID:            "tab-2",
				TabIndex:      2,
				Backend:       "proxy",
				Profile:       "workbench",
				RuntimeTarget: "node",
				BrowserApp:    "Chromium",
				Source:        "tracked_active_tab",
			},
		},
	)

	if projected.Snapshot.SelectedProfileSelection == nil ||
		projected.Snapshot.SelectedProfileSelection.Backend != "proxy" ||
		projected.Snapshot.SelectedProfileSelection.Profile != "workbench" ||
		projected.Snapshot.SelectedProfileSelection.RuntimeTarget != "node" ||
		projected.Snapshot.SelectedProfileSelection.BrowserApp != "Chromium" ||
		projected.Snapshot.SelectedProfileSelection.Source != "remember_profile" {
		t.Fatalf("expected merged profile selection projection, got %#v", projected.Snapshot.SelectedProfileSelection)
	}
	if projected.Snapshot.SelectedTargetSelection == nil ||
		projected.Snapshot.SelectedTargetSelection.ID != "tab-2" ||
		projected.Snapshot.SelectedTargetSelection.TabIndex != 2 ||
		projected.Snapshot.SelectedTargetSelection.BrowserApp != "Chromium" {
		t.Fatalf("expected merged target selection projection, got %#v", projected.Snapshot.SelectedTargetSelection)
	}
	if projected.Snapshot.CurrentTargetID != "tab-2" || projected.Snapshot.Summary.CurrentTargetID != "tab-2" {
		t.Fatalf("expected selection overlay to backfill current target identity, got snapshot=%#v summary=%#v", projected.Snapshot.CurrentTargetID, projected.Snapshot.Summary.CurrentTargetID)
	}
}

func TestProjectSharedSessionBrowserBindingEvaluationWithSelectionProjectionPreservesExistingSelectionIdentity(t *testing.T) {
	evaluation := SharedSessionBrowserBindingEvaluation{
		Snapshot: SharedSessionBrowserBindingSnapshot{
			CurrentTargetID: "tab-1",
			SelectedProfileSelection: &SharedSessionBrowserProfileSelection{
				Backend:       "proxy",
				Profile:       "workbench",
				RuntimeTarget: "node",
				BrowserApp:    "Chromium",
				Source:        "select_profile",
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
			Summary: SharedSessionBrowserBindingSummary{
				CurrentTargetID: "tab-1",
			},
		},
	}

	projected := ProjectSharedSessionBrowserBindingEvaluationWithSelectionProjection(
		evaluation,
		&SharedSessionBrowserSelectionProjection{
			ProfileSelection: &SharedSessionBrowserProfileSelection{
				Backend:       "sandbox",
				Profile:       "workbench",
				RuntimeTarget: "node",
				BrowserApp:    "Firefox",
				Source:        "other",
			},
			TargetSelection: &BrowserSessionTargetSelection{
				ID:            "tab-2",
				TabIndex:      2,
				Backend:       "sandbox",
				Profile:       "workbench",
				RuntimeTarget: "node",
				BrowserApp:    "Firefox",
				Source:        "other",
			},
		},
	)

	if projected.Snapshot.SelectedProfileSelection == nil ||
		projected.Snapshot.SelectedProfileSelection.Backend != "proxy" ||
		projected.Snapshot.SelectedProfileSelection.BrowserApp != "Chromium" ||
		projected.Snapshot.SelectedProfileSelection.Source != "select_profile" {
		t.Fatalf("expected existing profile selection identity to win, got %#v", projected.Snapshot.SelectedProfileSelection)
	}
	if projected.Snapshot.SelectedTargetSelection == nil ||
		projected.Snapshot.SelectedTargetSelection.ID != "tab-1" ||
		projected.Snapshot.SelectedTargetSelection.TabIndex != 1 ||
		projected.Snapshot.SelectedTargetSelection.BrowserApp != "Chromium" ||
		projected.Snapshot.SelectedTargetSelection.Source != "tracked_active_tab" {
		t.Fatalf("expected existing target selection identity to win, got %#v", projected.Snapshot.SelectedTargetSelection)
	}
	if projected.Snapshot.CurrentTargetID != "tab-1" || projected.Snapshot.Summary.CurrentTargetID != "tab-1" {
		t.Fatalf("expected existing current target identity to win, got snapshot=%#v summary=%#v", projected.Snapshot.CurrentTargetID, projected.Snapshot.Summary.CurrentTargetID)
	}
}
