package controlcontract

import "testing"

func TestObjectiveRuntimeProductizationReady(t *testing.T) {
	step := BuildObjectiveRuntimeLoopStep(ObjectiveRuntimeLoopInput{
		Activation: ActivationManaged,
		Frame: ObjectiveFrame{
			ID:              "objective:productization",
			UserGoalDigest:  "goal:productization",
			ControlMode:     ControlModeObjective,
			Intensity:       IntensityL3ManagedObjective,
			SuccessCriteria: []string{"runtime productization is visible to host"},
		},
		Ledger: AttemptLedgerPatch{
			ObjectiveID: "objective:productization",
			LedgerRef:   "ledger:productization",
		},
		LedgerPatch: AttemptLedgerPatch{
			ObjectiveID: "objective:productization",
			LedgerRef:   "ledger:productization",
			Attempts: []AttemptSummary{{
				Ref:              "attempt:productization_1",
				ObjectiveID:      "objective:productization",
				StrategyID:       "strategy:productization",
				ControlMode:      ControlModeObjective,
				Intensity:        IntensityL3ManagedObjective,
				Status:           VerificationPartial,
				EvidenceRefs:     []EvidenceRef{{Ref: "evidence:productization_attempt", Kind: "runtime", Strength: EvidenceStrong}},
				ObservationCount: 1,
			}},
			EvidenceRefs: []EvidenceRef{{Ref: "evidence:productization_ledger", Kind: "runtime", Strength: EvidenceStrong}},
		},
		Budget: ObjectiveBudgetSnapshot{BudgetRef: "budget:productization", Limit: 3, Remaining: 2},
		Approval: ObjectiveApprovalState{
			Approved:     true,
			ApprovalRefs: []DisplaySafeRef{"approval:productization"},
		},
		Strategies: []StrategyCandidate{{
			ID:           "strategy:productization",
			Kind:         "host_runtime",
			ControlMode:  ControlModeObjective,
			MinIntensity: IntensityL3ManagedObjective,
			MaxIntensity: IntensityL3ManagedObjective,
			Owner:        "host",
		}},
		CurrentStrategyRef: "strategy:productization",
		Verification: ObjectiveVerificationGateResult{
			Status:         VerificationPartial,
			FailureClass:   FailureEvidenceMissing,
			EvidenceRefs:   []EvidenceRef{{Ref: "evidence:productization_verification", Kind: "runtime", Strength: EvidenceAdequate}},
			NextHostAction: "request_replan_or_return_partial",
		},
		EvidenceRefs: []EvidenceRef{{Ref: "evidence:productization_input", Kind: "runtime", Strength: EvidenceAdequate}},
	})
	report := BuildObjectiveRuntimeProductization(ObjectiveRuntimeProductizationInput{
		Activation:          ActivationManaged,
		RuntimeLoop:         step,
		ObjectiveRunRef:     "objective_run:productization",
		TaskLedgerRef:       "task_ledger:productization",
		TrajectoryRef:       "trajectory:productization",
		WatchdogRef:         "watchdog:productization",
		HostRuntimeQueueRef: "runtime_queue:productization",
	})
	if report.Status != "ready_for_host_runtime_productization" ||
		!report.ReadyForHostProductization ||
		!report.ReadyForRuntimeContinuation ||
		!report.RuntimeLoopReadyForPersist ||
		!report.RuntimeLoopReadyForContinue ||
		!report.TaskLedgerReady ||
		!report.TrajectoryReady ||
		!report.WatchdogReady ||
		report.NoProgressDetected ||
		report.AttemptCount != 1 ||
		report.CurrentStrategyRef != "strategy:productization" ||
		report.NextHostAction != "continue_objective_runtime_loop" ||
		report.RunnerEffect != "none" ||
		report.PromptEffect != "none" ||
		report.RuntimeEffect != "none" {
		t.Fatalf("unexpected runtime productization report: %#v", report)
	}
	for _, want := range []Boundary{
		"objective_runtime_productization_task_ledger_ready",
		"objective_runtime_productization_trajectory_ready",
		"objective_runtime_productization_watchdog_ready",
		"no_runner_dispatch",
		"no_runtime_adapter_execution",
		"no_store_mutation_by_core",
	} {
		if !objectiveRuntimeProductizationBoundaryContains(report.Boundaries, want) {
			t.Fatalf("runtime productization missing boundary %q: %#v", want, report.Boundaries)
		}
	}
}

func TestObjectiveRuntimeProductizationWatchdogStopsRepeatedNoProgress(t *testing.T) {
	step := objectiveRuntimeProductizationNoProgressStep()
	report := BuildObjectiveRuntimeProductization(ObjectiveRuntimeProductizationInput{
		Activation:          ActivationManaged,
		RuntimeLoop:         step,
		ObjectiveRunRef:     "objective_run:productization_no_progress",
		TaskLedgerRef:       "task_ledger:productization_no_progress",
		TrajectoryRef:       "trajectory:productization_no_progress",
		WatchdogRef:         "watchdog:productization_no_progress",
		HostRuntimeQueueRef: "runtime_queue:productization_no_progress",
	})
	if report.Status != "watchdog_blocked_repeated_no_progress" ||
		!report.NoProgressDetected ||
		!report.ReadyForWatchdogStop ||
		report.ReadyForHostProductization ||
		report.FailureClass != FailureRepeatedNoProgress ||
		report.NextHostAction != "return_blocked" ||
		report.NoProgressAttemptCount < 2 {
		t.Fatalf("unexpected no-progress watchdog report: %#v", report)
	}
	if !objectiveRuntimeProductizationBoundaryContains(report.Boundaries, "objective_runtime_productization_no_progress_watchdog_stop") {
		t.Fatalf("expected watchdog stop boundary, got %#v", report.Boundaries)
	}
}

func TestObjectiveRuntimeProductizationBlocksInactiveAndRawRefs(t *testing.T) {
	step := objectiveRuntimeProductizationNoProgressStep()
	inactive := BuildObjectiveRuntimeProductization(ObjectiveRuntimeProductizationInput{
		Activation:      ActivationAdvisory,
		RuntimeLoop:     step,
		ObjectiveRunRef: "objective_run:inactive",
		TaskLedgerRef:   "task_ledger:inactive",
		TrajectoryRef:   "trajectory:inactive",
		WatchdogRef:     "watchdog:inactive",
	})
	if inactive.Status != "inactive" ||
		inactive.Available ||
		inactive.FailureClass != FailurePolicyBlocked ||
		!objectiveRuntimeProductizationMissingContains(inactive.MissingInputs, "host:managed_control_plane_activation") {
		t.Fatalf("unexpected inactive productization report: %#v", inactive)
	}

	raw := BuildObjectiveRuntimeProductization(ObjectiveRuntimeProductizationInput{
		Activation:      ActivationManaged,
		RuntimeLoop:     step,
		ObjectiveRunRef: "secret://example.invalid/run",
		TaskLedgerRef:   "task_ledger:raw",
		TrajectoryRef:   "trajectory:raw",
		WatchdogRef:     "watchdog:raw",
	})
	if raw.Status != "review_required" ||
		raw.FailureClass != FailureEvidenceWeak ||
		!objectiveRuntimeProductizationMissingContains(raw.MissingInputs, "host:display_safe_refs") ||
		!objectiveRuntimeProductizationBoundaryContains(raw.Boundaries, "raw_output_not_allowed") {
		t.Fatalf("unexpected raw productization report: %#v", raw)
	}
}

func objectiveRuntimeProductizationNoProgressStep() ObjectiveRuntimeLoopStep {
	return BuildObjectiveRuntimeLoopStep(ObjectiveRuntimeLoopInput{
		Activation: ActivationManaged,
		Frame: ObjectiveFrame{
			ID:             "objective:productization_no_progress",
			UserGoalDigest: "goal:productization_no_progress",
			ControlMode:    ControlModeObjective,
			Intensity:      IntensityL3ManagedObjective,
		},
		Ledger: AttemptLedgerPatch{
			ObjectiveID: "objective:productization_no_progress",
			LedgerRef:   "ledger:productization_no_progress",
			Attempts: []AttemptSummary{
				{
					Ref:          "attempt:no_progress_1",
					ObjectiveID:  "objective:productization_no_progress",
					StrategyID:   "strategy:no_progress",
					ControlMode:  ControlModeObjective,
					Intensity:    IntensityL3ManagedObjective,
					Status:       VerificationFailed,
					FailureClass: FailureVerificationFailed,
				},
				{
					Ref:          "attempt:no_progress_2",
					ObjectiveID:  "objective:productization_no_progress",
					StrategyID:   "strategy:no_progress",
					ControlMode:  ControlModeObjective,
					Intensity:    IntensityL3ManagedObjective,
					Status:       VerificationFailed,
					FailureClass: FailureVerificationFailed,
				},
			},
			EvidenceRefs: []EvidenceRef{{Ref: "evidence:no_progress_ledger", Kind: "runtime", Strength: EvidenceAdequate}},
		},
		LedgerPatch: AttemptLedgerPatch{
			ObjectiveID: "objective:productization_no_progress",
			LedgerRef:   "ledger:productization_no_progress",
			Attempts: []AttemptSummary{{
				Ref:          "attempt:no_progress_3",
				ObjectiveID:  "objective:productization_no_progress",
				StrategyID:   "strategy:no_progress",
				ControlMode:  ControlModeObjective,
				Intensity:    IntensityL3ManagedObjective,
				Status:       VerificationFailed,
				FailureClass: FailureVerificationFailed,
			}},
		},
		Budget:             ObjectiveBudgetSnapshot{BudgetRef: "budget:no_progress", Limit: 3, Remaining: 1},
		Approval:           ObjectiveApprovalState{Approved: true},
		CurrentStrategyRef: "strategy:no_progress",
		Verification: ObjectiveVerificationGateResult{
			Status:         VerificationPartial,
			FailureClass:   FailureEvidenceMissing,
			NextHostAction: "request_replan_or_return_partial",
		},
		EvidenceRefs: []EvidenceRef{{Ref: "evidence:no_progress_runtime", Kind: "runtime", Strength: EvidenceAdequate}},
	})
}

func objectiveRuntimeProductizationBoundaryContains(items []Boundary, target Boundary) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func objectiveRuntimeProductizationMissingContains(items []MissingInput, target MissingInput) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
