package browserruntime

import "testing"

func TestBuildSharedSessionBrowserLifecycleActionOutcomeProjectsExecutionSurface(t *testing.T) {
	stateRegistry := NewBrowserSessionStateRegistry()
	outcome := BuildSharedSessionBrowserLifecycleActionOutcome(
		SharedSessionBrowserLifecycleActionOutcomeRequest{
			Action:       "restart",
			SessionID:    "sess-lifecycle-outcome",
			SelectedInfo: BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
			MutationContext: SharedSessionBrowserMutationContext{
				StateRegistry: stateRegistry,
			},
			DispatchResult: SharedSessionBrowserLifecycleActionDispatchResult{
				Result: SharedSessionBrowserExecutionResult{
					Profile: "workbench",
					Profiles: &BrowserProfilesResult{
						DefaultProfile: "workbench",
						Profiles: []BrowserProfileInfo{
							{Profile: "workbench", BrowserApp: "Chromium", Status: "running", Running: true},
						},
						Note: "profiles synced",
					},
					Decision: "restart_started",
					Ready:    false,
					ProfileStatus: BrowserProfileStatusResult{
						Backend:    "proxy",
						Profile:    "workbench",
						BrowserApp: "Chromium",
						Status:     "starting",
						Running:    true,
					},
				},
			},
			RememberProfile: true,
		},
	)

	if outcome.Action != "restart" || outcome.Result.Decision != "restart_started" || outcome.Result.Profile != "workbench" {
		t.Fatalf("expected lifecycle outcome to preserve result semantics, got %#v", outcome)
	}
	if outcome.RememberOutcome == nil || outcome.RememberOutcome.Decision != "session_profile_remembered" || !outcome.RememberOutcome.Ready {
		t.Fatalf("expected lifecycle outcome to dispatch remember-profile contract, got %#v", outcome.RememberOutcome)
	}
	if outcome.RememberOutcome.SelectionProjection == nil || outcome.RememberOutcome.SelectionProjection.ProfileSelection == nil {
		t.Fatalf("expected lifecycle outcome remember result to include selection projection, got %#v", outcome.RememberOutcome)
	}
	if !outcome.ExecutionProjection.Surface.HasProfileState ||
		outcome.ExecutionProjection.Surface.ProfileState.Profile != "workbench" ||
		outcome.ExecutionProjection.Surface.ProfileState.Status != "reconnecting" {
		t.Fatalf("expected lifecycle outcome to project applied execution surface, got %#v", outcome.ExecutionProjection.Surface)
	}
	if outcome.ExecutionProjection.Surface.Note != "profiles synced" ||
		!outcome.ExecutionProjection.Surface.ApplyProfileInventory ||
		outcome.ExecutionProjection.Surface.DefaultProfile != "workbench" {
		t.Fatalf("expected lifecycle outcome to preserve projected execution surface, got %#v", outcome.ExecutionProjection.Surface)
	}
	if outcome.ExecutionProjection.InventoryProjection == nil ||
		!outcome.ExecutionProjection.InventoryProjection.HasProfileState ||
		outcome.ExecutionProjection.InventoryProjection.ProfileState.Profile != "workbench" {
		t.Fatalf("expected lifecycle outcome to carry execution inventory projection, got %#v", outcome.ExecutionProjection.InventoryProjection)
	}
}

func TestBuildSharedSessionBrowserLifecycleActionOutcomePreservesCoordinationErrorBehavior(t *testing.T) {
	outcome := BuildSharedSessionBrowserLifecycleActionOutcome(
		SharedSessionBrowserLifecycleActionOutcomeRequest{
			Action:           "coordinate",
			CoordinationGoal: "restart",
			SelectedInfo:     BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
			DispatchResult: SharedSessionBrowserLifecycleActionDispatchResult{
				Result: SharedSessionBrowserExecutionResult{
					Profile:  "workbench",
					Decision: "restart_started",
				},
				Err: errLifecycleOutcomeTest("restart failed"),
			},
			ApplyCoordinationDecisionOnError: true,
		},
	)

	if outcome.Action != "coordinate" || outcome.CoordinationGoal != "restart" || outcome.Err == nil {
		t.Fatalf("expected lifecycle outcome to preserve action/goal/error, got %#v", outcome)
	}
	if outcome.Status != "error" || outcome.Note != "restart failed" {
		t.Fatalf("expected lifecycle outcome to own error terminal status/note, got %#v", outcome)
	}
	if !outcome.ApplyCoordinationDecisionOnError {
		t.Fatalf("expected lifecycle outcome to preserve coordination error policy, got %#v", outcome)
	}
	if outcome.RememberOutcome != nil {
		t.Fatalf("expected lifecycle outcome not to dispatch remember-profile follow-up on error, got %#v", outcome.RememberOutcome)
	}
}

func TestBuildSharedSessionBrowserLifecycleActionOutcomeProjectsDecisionFields(t *testing.T) {
	t.Run("coordinate_restart_uses_restart_fields", func(t *testing.T) {
		outcome := BuildSharedSessionBrowserLifecycleActionOutcome(
			SharedSessionBrowserLifecycleActionOutcomeRequest{
				Action:           "coordinate",
				CoordinationGoal: "restart",
				SelectedInfo:     BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
				DispatchResult: SharedSessionBrowserLifecycleActionDispatchResult{
					Result: SharedSessionBrowserExecutionResult{
						Profile:  "workbench",
						Decision: "session_restarted",
						Ready:    true,
					},
				},
			},
		)

		if outcome.PreparedProfile != "workbench" {
			t.Fatalf("expected lifecycle outcome to keep prepared profile, got %#v", outcome)
		}
		if outcome.RestartDecision != "session_restarted" || !outcome.RestartReady {
			t.Fatalf("expected lifecycle outcome to project restart fields, got %#v", outcome)
		}
		if outcome.PrepareDecision != "" || outcome.PrepareReady {
			t.Fatalf("expected lifecycle outcome not to project prepare fields for coordinate restart, got %#v", outcome)
		}
	})

	t.Run("coordinate_teardown_uses_prepare_fields", func(t *testing.T) {
		outcome := BuildSharedSessionBrowserLifecycleActionOutcome(
			SharedSessionBrowserLifecycleActionOutcomeRequest{
				Action:           "coordinate",
				CoordinationGoal: "teardown",
				SelectedInfo:     BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
				DispatchResult: SharedSessionBrowserLifecycleActionDispatchResult{
					Result: SharedSessionBrowserExecutionResult{
						Profile:  "workbench",
						Decision: "session_stopped",
						Ready:    true,
					},
				},
			},
		)

		if outcome.PreparedProfile != "workbench" {
			t.Fatalf("expected lifecycle outcome to keep prepared profile, got %#v", outcome)
		}
		if outcome.PrepareDecision != "session_stopped" || !outcome.PrepareReady {
			t.Fatalf("expected lifecycle outcome to project prepare fields, got %#v", outcome)
		}
		if outcome.RestartDecision != "" || outcome.RestartReady {
			t.Fatalf("expected lifecycle outcome not to project restart fields for coordinate teardown, got %#v", outcome)
		}
	})
}

type errLifecycleOutcomeTest string

func (e errLifecycleOutcomeTest) Error() string { return string(e) }
