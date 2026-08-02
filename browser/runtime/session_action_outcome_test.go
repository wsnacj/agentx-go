package browserruntime

import "testing"

func TestBuildSharedSessionBrowserSelectionActionOutcomeSelectProfileProjectsSelection(t *testing.T) {
	outcome := BuildSharedSessionBrowserSelectionActionOutcome(
		SharedSessionBrowserSelectionActionOutcomeRequest{
			Action: "select_profile",
			DispatchResult: SharedSessionBrowserSelectionActionDispatchResult{
				Decision: "session_profile_selected",
				Ready:    true,
				ProfileSelection: &SharedSessionBrowserProfileSelection{
					Backend:       "proxy",
					Profile:       "workbench",
					RuntimeTarget: "node",
					Source:        "select_profile",
				},
				TargetSelection: &BrowserSessionTargetSelection{
					ID:            "target-1",
					Backend:       "proxy",
					Profile:       "workbench",
					RuntimeTarget: "node",
					Source:        "select_profile",
				},
			},
		},
	)

	if !outcome.ApplyDecision || outcome.Action != "select_profile" || outcome.Decision != "session_profile_selected" || !outcome.Ready {
		t.Fatalf("expected lowered select_profile outcome to preserve decision/ready, got %#v", outcome)
	}
	if outcome.SelectDecision != "session_profile_selected" || !outcome.SelectReady {
		t.Fatalf("expected lowered select_profile outcome to project action-specific decision fields, got %#v", outcome)
	}
	if outcome.SelectionProjection == nil || outcome.SelectionProjection.ProfileSelection == nil || outcome.SelectionProjection.ProfileSelection.Profile != "workbench" {
		t.Fatalf("expected lowered select_profile outcome to carry profile projection, got %#v", outcome.SelectionProjection)
	}
	if outcome.SelectionProjection.TargetSelection == nil || outcome.SelectionProjection.TargetSelection.ID != "target-1" {
		t.Fatalf("expected lowered select_profile outcome to preserve target projection, got %#v", outcome.SelectionProjection)
	}
	if outcome.Status != "" || outcome.Note != "" {
		t.Fatalf("expected lowered select_profile outcome not to set terminal status/note on success, got %#v", outcome)
	}
}

func TestBuildSharedSessionBrowserSelectionActionOutcomeRequiresReviewForPendingTarget(t *testing.T) {
	outcome := BuildSharedSessionBrowserSelectionActionOutcome(
		SharedSessionBrowserSelectionActionOutcomeRequest{
			Action: "select_target",
			DispatchResult: SharedSessionBrowserSelectionActionDispatchResult{
				Decision: "session_target_popup_review_required",
				Note:     "pending popup target",
			},
			ApplyTargetToRoute: true,
		},
	)

	if outcome.Decision != "session_target_popup_review_required" || outcome.Ready {
		t.Fatalf("expected lowered select_target review outcome to preserve decision/ready=false, got %#v", outcome)
	}
	if outcome.SelectTargetDecision != "session_target_popup_review_required" || outcome.SelectTargetReady {
		t.Fatalf("expected lowered select_target review outcome to project action-specific decision fields, got %#v", outcome)
	}
	if outcome.Status != "review_required" || outcome.Note != "pending popup target" {
		t.Fatalf("expected lowered select_target review outcome to keep review-required status/note, got %#v", outcome)
	}
	if outcome.SelectionProjection != nil {
		t.Fatalf("expected review-required outcome without target selection to skip projection, got %#v", outcome.SelectionProjection)
	}
}

func TestBuildSharedSessionBrowserSelectionActionOutcomeUsesMissingInputNote(t *testing.T) {
	outcome := BuildSharedSessionBrowserSelectionActionOutcome(
		SharedSessionBrowserSelectionActionOutcomeRequest{
			Action: "select_target",
		},
	)

	if outcome.Status != "error" || outcome.Note != "browser_runtime: target or tab_index is required for action select_target" {
		t.Fatalf("expected lowered select_target outcome to surface missing-input error, got %#v", outcome)
	}
}

func TestBuildSharedSessionBrowserMissingInputActionOutcome(t *testing.T) {
	outcome := BuildSharedSessionBrowserMissingInputActionOutcome(" select_profile ", "session_profile_required")

	if !outcome.ApplyDecision || outcome.Action != "select_profile" || outcome.Decision != "session_profile_required" {
		t.Fatalf("expected missing-input action outcome to preserve action semantics, got %#v", outcome)
	}
	if outcome.SelectDecision != "session_profile_required" || outcome.SelectReady {
		t.Fatalf("expected missing-input action outcome to project select_profile decision fields, got %#v", outcome)
	}
	if outcome.Status != "error" || outcome.Note != "browser_runtime: profile is required for action select_profile" {
		t.Fatalf("expected missing-input action outcome to own canonical status/note, got %#v", outcome)
	}
}

func TestBuildSharedSessionBrowserSyncActionOutcomeCarriesRouteMutationProjection(t *testing.T) {
	outcome := BuildSharedSessionBrowserSyncActionOutcome(
		"coordinate",
		"sync",
		SharedSessionBrowserSyncActionDispatchResult{
			Result: SharedSessionBrowserSyncSelectionResult{
				Decision: "session_route_synced",
				Ready:    true,
				ProfileSelection: &SharedSessionBrowserProfileSelection{
					Profile:       "workbench",
					RuntimeTarget: "node",
					Source:        "sync_session",
				},
				TargetSelection: &BrowserSessionTargetSelection{
					ID:            "target-2",
					Profile:       "workbench",
					RuntimeTarget: "node",
					Source:        "sync_session",
				},
			},
		},
		true,
		true,
	)

	if outcome.Action != "coordinate" || outcome.CoordinationGoal != "sync" || outcome.Decision != "session_route_synced" || !outcome.Ready {
		t.Fatalf("expected lowered sync outcome to preserve action semantics, got %#v", outcome)
	}
	if outcome.SyncSessionDecision != "session_route_synced" || !outcome.SyncSessionReady {
		t.Fatalf("expected lowered sync outcome to project sync_session decision fields, got %#v", outcome)
	}
	if !outcome.ApplyCoordinationDecisionOnError {
		t.Fatalf("expected lowered sync outcome to preserve coordination error behavior, got %#v", outcome)
	}
	if outcome.SelectionProjection == nil || !outcome.SelectionProjection.ApplyTargetToRoute || outcome.SelectionProjection.TargetSelection == nil || outcome.SelectionProjection.TargetSelection.ID != "target-2" {
		t.Fatalf("expected lowered sync outcome to carry route-mutation projection, got %#v", outcome.SelectionProjection)
	}
	projected, ok := ProjectSharedSessionBrowserActionDecision(outcome)
	if !ok || projected.Action != "sync_session" || projected.Decision != "session_route_synced" || !projected.Ready {
		t.Fatalf("expected sync outcome projection to target sync_session decision fields, got %#v ok=%v", projected, ok)
	}
}

func TestBuildSharedSessionBrowserClearActionOutcomeCarriesClearResult(t *testing.T) {
	outcome := BuildSharedSessionBrowserClearActionOutcome(
		"clear_session",
		BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
		SharedSessionBrowserClearResult{
			Decision:               "session_route_cleared",
			Ready:                  true,
			ClearedSessionProfiles: 2,
			ClearedSessionTargets:  1,
			ProfileState: SharedSessionBrowserProfileState{
				Backend:       "proxy",
				Profile:       "workbench",
				RuntimeTarget: "node",
				Status:        "stopped",
			},
			HasProfileState: true,
		},
	)

	if !outcome.ApplyDecision || outcome.Action != "clear_session" {
		t.Fatalf("expected lowered clear outcome to preserve action semantics, got %#v", outcome)
	}
	if outcome.Decision != "session_route_cleared" || !outcome.Ready || !outcome.ClearProfileStatus || outcome.ClearedSessionProfiles != 2 || outcome.ClearedSessionTargets != 1 {
		t.Fatalf("expected lowered clear outcome to keep clear-result projection details, got %#v", outcome)
	}
	if outcome.ClearSessionDecision != "session_route_cleared" || !outcome.ClearSessionReady {
		t.Fatalf("expected lowered clear outcome to project clear_session decision fields, got %#v", outcome)
	}
	if outcome.ProfileInventoryProjection == nil || !outcome.ProfileInventoryProjection.HasProfileStatus || outcome.ProfileInventoryProjection.ProfileStatus.Profile != "workbench" || outcome.ProfileInventoryProjection.ProfileStatus.RuntimeTarget != "node" {
		t.Fatalf("expected lowered clear outcome to carry shared profile inventory projection, got %#v", outcome.ProfileInventoryProjection)
	}
	projected, ok := ProjectSharedSessionBrowserActionDecision(outcome)
	if !ok || projected.Action != "clear_session" || projected.Decision != "session_route_cleared" || !projected.Ready || !projected.ClearProfileStatus || projected.ClearedSessionProfiles != 2 || projected.ClearedSessionTargets != 1 {
		t.Fatalf("expected clear outcome projection to carry clear counters, got %#v ok=%v", projected, ok)
	}
}

func TestBuildSharedSessionBrowserRememberActionOutcomeCarriesSelectionProjection(t *testing.T) {
	outcome := BuildSharedSessionBrowserRememberActionOutcome(
		SharedSessionBrowserRememberProfileResult{
			Decision: "session_profile_remembered",
			Ready:    true,
			SelectionProjection: &SharedSessionBrowserSelectionProjection{
				ProfileSelection: &SharedSessionBrowserProfileSelection{
					Backend:       "proxy",
					Profile:       "workbench",
					RuntimeTarget: "node",
					Source:        "remember_profile",
				},
				TargetSelection: &BrowserSessionTargetSelection{
					ID:            "target-1",
					Backend:       "proxy",
					Profile:       "workbench",
					RuntimeTarget: "node",
					Source:        "remember_profile",
				},
				ApplyTargetToRoute: true,
			},
		},
	)

	if !outcome.ApplyDecision || outcome.Action != "remember_profile" || outcome.Decision != "session_profile_remembered" || !outcome.Ready {
		t.Fatalf("expected lowered remember outcome to preserve decision semantics, got %#v", outcome)
	}
	if outcome.RememberDecision != "session_profile_remembered" || !outcome.RememberReady {
		t.Fatalf("expected lowered remember outcome to project action-specific decision fields, got %#v", outcome)
	}
	if outcome.SelectionProjection == nil || outcome.SelectionProjection.ProfileSelection == nil || outcome.SelectionProjection.TargetSelection == nil || !outcome.SelectionProjection.ApplyTargetToRoute {
		t.Fatalf("expected lowered remember outcome to keep selection projection, got %#v", outcome.SelectionProjection)
	}
}

func TestBuildSharedSessionBrowserBasicActionOutcomeProjectsDecisionFields(t *testing.T) {
	outcome := BuildSharedSessionBrowserBasicActionOutcome(
		SharedSessionBrowserBasicActionOutcomeRequest{
			ApplyDecision: true,
			Action:        "select_profile",
			Decision:      "session_profile_required",
			Status:        "error",
			Note:          "browser_runtime: profile is required for action select_profile",
		},
	)

	if !outcome.ApplyDecision || outcome.Action != "select_profile" || outcome.Decision != "session_profile_required" {
		t.Fatalf("expected basic action outcome to preserve action semantics, got %#v", outcome)
	}
	if outcome.SelectDecision != "session_profile_required" || outcome.SelectReady {
		t.Fatalf("expected basic action outcome to project select_profile decision fields, got %#v", outcome)
	}
	if outcome.Status != "error" || outcome.Note != "browser_runtime: profile is required for action select_profile" {
		t.Fatalf("expected basic action outcome to preserve status/note, got %#v", outcome)
	}
}

func TestBuildSharedSessionBrowserBasicActionOutcomePreservesError(t *testing.T) {
	outcome := BuildSharedSessionBrowserBasicActionOutcome(
		SharedSessionBrowserBasicActionOutcomeRequest{
			ApplyDecision: true,
			Action:        "select_target",
			Decision:      "session_target_invalid",
			Err:           errLifecycleOutcomeTest("browser_runtime: invalid target"),
		},
	)

	if outcome.Decision != "session_target_invalid" || outcome.Err == nil {
		t.Fatalf("expected basic action outcome to preserve error-bearing decision, got %#v", outcome)
	}
	if outcome.SelectTargetDecision != "session_target_invalid" || outcome.SelectTargetReady {
		t.Fatalf("expected basic action outcome to project select_target decision fields, got %#v", outcome)
	}
	if outcome.Status != "error" || outcome.Note != "browser_runtime: invalid target" {
		t.Fatalf("expected basic action outcome to own error terminal status/note, got %#v", outcome)
	}
}

func TestBuildSharedSessionBrowserInvalidSelectTargetActionOutcome(t *testing.T) {
	outcome := BuildSharedSessionBrowserInvalidSelectTargetActionOutcome(
		errLifecycleOutcomeTest("browser_runtime: invalid target"),
	)

	if outcome.Decision != "session_target_invalid" || outcome.Err == nil {
		t.Fatalf("expected invalid select_target outcome to preserve canonical decision and error, got %#v", outcome)
	}
	if outcome.SelectTargetDecision != "session_target_invalid" || outcome.SelectTargetReady {
		t.Fatalf("expected invalid select_target outcome to project action-specific decision fields, got %#v", outcome)
	}
	if outcome.Status != "error" || outcome.Note != "browser_runtime: invalid target" {
		t.Fatalf("expected invalid select_target outcome to own error terminal status/note, got %#v", outcome)
	}
}

func TestBuildSharedSessionBrowserUnsupportedActionOutcome(t *testing.T) {
	outcome := BuildSharedSessionBrowserUnsupportedActionOutcome(" Refresh ")

	if outcome.Action != "refresh" {
		t.Fatalf("expected unsupported action outcome to normalize action, got %#v", outcome)
	}
	if outcome.Status != "unsupported" || outcome.Note != "browser_runtime: selected route does not support action refresh" {
		t.Fatalf("expected unsupported action outcome to own unsupported status/note, got %#v", outcome)
	}
	if outcome.ApplyDecision || outcome.Decision != "" || outcome.Err != nil {
		t.Fatalf("expected unsupported action outcome to stay terminal-only, got %#v", outcome)
	}
}

func TestBuildSharedSessionBrowserUnsupportedRouteActionOutcomePreservesRouteErrorNote(t *testing.T) {
	outcome := BuildSharedSessionBrowserUnsupportedRouteActionOutcome(
		" refresh ",
		errLifecycleOutcomeTest("browser_runtime: managed browserd boot failed"),
	)

	if outcome.Action != "refresh" {
		t.Fatalf("expected unsupported route outcome to normalize action, got %#v", outcome)
	}
	if outcome.Status != "unsupported" || outcome.Note != "browser_runtime: managed browserd boot failed" {
		t.Fatalf("expected unsupported route outcome to preserve route-error note, got %#v", outcome)
	}
	if outcome.ApplyDecision || outcome.Decision != "" || outcome.Err != nil {
		t.Fatalf("expected unsupported route outcome to stay terminal-only, got %#v", outcome)
	}
}
