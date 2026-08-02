package browserruntime

import (
	"testing"
	"time"
)

func TestProjectSharedSessionBrowserTopLevelBindingBuildsSelectionsAndMergesCurrentProfileState(t *testing.T) {
	sessionID := "sess-binding-projection"
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	tracked := sessionRegistry.TrackCurrentTarget(sessionID, BrowserSessionTarget{
		ID:         "tab-1",
		TabIndex:   1,
		URL:        "https://example.com",
		Title:      "Example",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "tracked_active_tab")
	stateRegistry.SelectBrowserProfile(sessionID, SharedSessionBrowserProfileSelection{
		Backend:       "proxy",
		Profile:       "workbench",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Source:        "select_profile",
	})

	projection := ProjectSharedSessionBrowserTopLevelBinding(
		sessionID,
		BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
		BrowserSessionRoute{Backend: "proxy", Profile: "workbench", Target: "node"},
		nil,
		sessionRegistry,
		nil,
		stateRegistry,
		nil,
		&SharedSessionBrowserProfileState{
			Backend:       "proxy",
			Profile:       "workbench",
			RuntimeTarget: "node",
			BrowserApp:    "Chromium",
			Status:        "running",
			Running:       true,
			Connected:     true,
			Note:          "fresh current status",
		},
		time.Minute,
	)

	if projection.ProfileSelection == nil || projection.ProfileSelection.Profile != "workbench" || projection.ProfileSelection.Source != "select_profile" {
		t.Fatalf("expected top-level binding projection to expose selected profile, got %#v", projection.ProfileSelection)
	}
	if projection.TargetSelection == nil || projection.TargetSelection.ID != tracked.ID || projection.TargetSelection.Source != "tracked_active_tab" {
		t.Fatalf("expected top-level binding projection to expose selected target, got %#v", projection.TargetSelection)
	}
	if len(projection.Evaluation.Snapshot.Profiles) != 1 || projection.Evaluation.Snapshot.Profiles[0].Profile != "workbench" || projection.Evaluation.Snapshot.Profiles[0].Status != "running" || projection.Evaluation.Snapshot.Profiles[0].Note != "fresh current status" {
		t.Fatalf("expected top-level binding projection to merge current profile state, got %#v", projection.Evaluation.Snapshot.Profiles)
	}
	if projection.Evaluation.Snapshot.CurrentTargetID != tracked.ID || projection.Evaluation.Snapshot.Summary.CurrentTargetID != tracked.ID {
		t.Fatalf("expected top-level binding projection to preserve current target id, got %#v", projection.Evaluation.Snapshot)
	}
}

func TestProjectSharedSessionBrowserTopLevelBindingReusesProvidedEvaluation(t *testing.T) {
	projection := ProjectSharedSessionBrowserTopLevelBinding(
		"",
		BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
		BrowserSessionRoute{Backend: "proxy", Profile: "workbench", Target: "node"},
		nil,
		nil,
		nil,
		nil,
		&SharedSessionBrowserBindingEvaluation{
			Snapshot: SharedSessionBrowserBindingSnapshot{
				CurrentTargetID: "tab-2",
				SelectedProfileSelection: &SharedSessionBrowserProfileSelection{
					Backend:       "proxy",
					Profile:       "workbench",
					RuntimeTarget: "node",
					BrowserApp:    "Chromium",
					Source:        "remember_profile",
				},
				SelectedTargetSelection: &BrowserSessionTargetSelection{
					ID:            "tab-2",
					TabIndex:      2,
					Backend:       "proxy",
					Profile:       "workbench",
					RuntimeTarget: "node",
					BrowserApp:    "Chromium",
					Source:        "tracked_active_tab",
				},
			},
		},
		nil,
		time.Minute,
	)

	if projection.ProfileSelection == nil || projection.ProfileSelection.Profile != "workbench" || projection.ProfileSelection.Source != "remember_profile" {
		t.Fatalf("expected top-level binding projection to reuse provided profile selection, got %#v", projection.ProfileSelection)
	}
	if projection.TargetSelection == nil || projection.TargetSelection.ID != "tab-2" || projection.TargetSelection.Source != "tracked_active_tab" {
		t.Fatalf("expected top-level binding projection to reuse provided target selection, got %#v", projection.TargetSelection)
	}
	if projection.Evaluation.Snapshot.CurrentTargetID != "tab-2" {
		t.Fatalf("expected top-level binding projection to preserve provided evaluation snapshot, got %#v", projection.Evaluation.Snapshot)
	}
}
