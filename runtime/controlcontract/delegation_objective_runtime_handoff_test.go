package controlcontract

import "testing"

func TestDelegationObjectiveRuntimeHandoffFeedsRuntimeLoop(t *testing.T) {
	merge := BuildDelegationWorkerParentMerge(DelegationWorkerParentMergeInput{
		Invocation:       delegationReadyWorkerInvocation(),
		ParentLedgerRef:  "ledger:delegation_parent",
		WorkerAttemptRef: "attempt:delegation_worker_1",
		MergeRef:         "merge:delegation_worker_1",
		WorkerObservations: []Observation{{
			Kind:     "delegation_worker_result",
			Source:   "worker:delegation_fixture",
			Subject:  "objective:delegation_fixture",
			Name:     "worker_result_verified",
			Value:    "true",
			Strength: EvidenceAdequate,
			EvidenceRefs: []EvidenceRef{{
				Ref:      "evidence:delegation_worker_verified",
				Kind:     "delegation_worker_result",
				Strength: EvidenceAdequate,
				Source:   "worker:delegation_fixture",
			}},
		}},
		ExpectedObservationKinds: []string{"delegation_worker_result"},
	})
	run := delegationObjectiveRuntimeTestRun()
	handoff := BuildDelegationObjectiveRuntimeHandoff(DelegationObjectiveRuntimeHandoffInput{
		HandoffRef:  "handoff:delegation_parent_merge",
		Run:         run,
		ParentMerge: merge,
	})
	if handoff.Status != HostActionRecorded ||
		!handoff.ReadyForRuntimeLoopInput ||
		handoff.ReadyForFailureReview ||
		handoff.NextHostAction != "run_objective_runtime_loop_step" ||
		handoff.LedgerPatch.LedgerRef != "ledger:delegation_parent" ||
		len(handoff.LedgerPatch.Attempts) != 1 ||
		handoff.LedgerPatch.Attempts[0].ControlMode != ControlModeDelegated ||
		handoff.WorkerDispatched ||
		handoff.RunnerDispatched ||
		handoff.StoreMutationExecuted {
		t.Fatalf("delegation runtime handoff = %#v", handoff)
	}
	step := BuildObjectiveRuntimeLoopStep(handoff.RuntimeLoopInput)
	if step.Status != "ready_for_host_persist" ||
		step.Run.State != ObjectiveControllerSatisfied ||
		step.ControllerDecision.Action != ObjectiveActionReturnSatisfied ||
		!step.ReadyForHostPersist ||
		len(step.Run.Ledger.Attempts) != 1 ||
		step.Run.Ledger.Attempts[0].Ref != "attempt:delegation_worker_1" {
		t.Fatalf("runtime loop did not consume delegation handoff: %#v", step)
	}
	for _, boundary := range []Boundary{
		"delegation_objective_runtime_handoff",
		"delegation_parent_merge_ready_for_runtime_loop",
		"worker_output_not_fact",
		"no_store_mutation_by_core",
	} {
		if !delegationBoundaryContains(handoff.Boundaries, boundary) {
			t.Fatalf("missing boundary %q in %#v", boundary, handoff.Boundaries)
		}
	}
}

func TestDelegationObjectiveRuntimeHandoffRoutesFailureReview(t *testing.T) {
	review := BuildDelegationWorkerFailureReview(DelegationWorkerFailureReviewInput{
		Request:          BuildDelegationRequestProjection(delegationReadyRequestInput(IntensityL4DurableLongRun)),
		FailureReviewRef: "failure_review:delegation_worker_no_progress",
		FailureRef:       "failure:delegation_worker_no_progress",
		CompensationRef:  "compensation:delegation_worker_no_progress",
		WorkerAttempts: []AttemptSummary{
			{
				Ref:          "attempt:delegation_worker_1",
				ObjectiveID:  "objective:delegation_fixture",
				StrategyID:   "worker:delegation_fixture_readonly",
				ControlMode:  ControlModeDelegated,
				Intensity:    IntensityL4DurableLongRun,
				Status:       VerificationPartial,
				FailureClass: FailureEvidenceMissing,
			},
			{
				Ref:          "attempt:delegation_worker_2",
				ObjectiveID:  "objective:delegation_fixture",
				StrategyID:   "worker:delegation_fixture_readonly",
				ControlMode:  ControlModeDelegated,
				Intensity:    IntensityL4DurableLongRun,
				Status:       VerificationBlocked,
				FailureClass: FailureEvidenceMissing,
			},
		},
	})
	handoff := BuildDelegationObjectiveRuntimeHandoff(DelegationObjectiveRuntimeHandoffInput{
		HandoffRef:    "handoff:delegation_failure_review",
		Run:           delegationObjectiveRuntimeTestRun(),
		FailureReview: review,
	})
	if handoff.Status != HostActionRecorded ||
		handoff.ReadyForRuntimeLoopInput ||
		!handoff.ReadyForFailureReview ||
		handoff.FailureClass != FailureRepeatedNoProgress ||
		handoff.NextHostAction != "review_delegation_worker_failure" ||
		!delegationBoundaryContains(handoff.Boundaries, "delegation_worker_failure_review_handoff_ready") {
		t.Fatalf("delegation failure review handoff = %#v", handoff)
	}
}

func TestDelegationObjectiveRuntimeHandoffBlocksMismatchedParentLedger(t *testing.T) {
	merge := BuildDelegationWorkerParentMerge(DelegationWorkerParentMergeInput{
		Invocation:       delegationReadyWorkerInvocation(),
		ParentLedgerRef:  "ledger:delegation_parent",
		WorkerAttemptRef: "attempt:delegation_worker_1",
		MergeRef:         "merge:delegation_worker_1",
		WorkerObservations: []Observation{{
			Kind:     "delegation_worker_result",
			Source:   "worker:delegation_fixture",
			Strength: EvidenceAdequate,
			EvidenceRefs: []EvidenceRef{{
				Ref:      "evidence:delegation_worker_verified",
				Kind:     "delegation_worker_result",
				Strength: EvidenceAdequate,
				Source:   "worker:delegation_fixture",
			}},
		}},
		ExpectedObservationKinds: []string{"delegation_worker_result"},
	})
	run := delegationObjectiveRuntimeTestRun()
	run.Ledger.LedgerRef = "ledger:other_parent"
	handoff := BuildDelegationObjectiveRuntimeHandoff(DelegationObjectiveRuntimeHandoffInput{
		HandoffRef:  "handoff:delegation_parent_merge",
		Run:         run,
		ParentMerge: merge,
	})
	if handoff.ReadyForRuntimeLoopInput ||
		handoff.Status == HostActionRecorded ||
		handoff.FailureClass != FailureVerificationFailed ||
		!delegationStringContains(handoff.BlockedReasons, "parent_ledger_ref_mismatch") ||
		!delegationBoundaryContains(handoff.Boundaries, "delegation_parent_ledger_ref_mismatch") {
		t.Fatalf("expected ledger mismatch block, got %#v", handoff)
	}
}

func delegationObjectiveRuntimeTestRun() ObjectiveRun {
	return BuildObjectiveRun(ObjectiveRunInput{
		Activation: ActivationManaged,
		Frame: ObjectiveFrame{
			ID:              "objective:delegation_fixture",
			UserGoalDigest:  "delegation fixture objective",
			ControlMode:     ControlModeDelegated,
			Intensity:       IntensityL4DurableLongRun,
			SuccessCriteria: []string{"worker result satisfies parent evidence"},
			RequiredEvidence: []EvidenceRef{{
				Ref:      "evidence:delegation_worker_verified",
				Kind:     "delegation_worker_result",
				Strength: EvidenceAdequate,
				Source:   "worker:delegation_fixture",
			}},
		},
		Ledger: AttemptLedgerPatch{
			ObjectiveID: "objective:delegation_fixture",
			LedgerRef:   "ledger:delegation_parent",
		},
		Budget: ObjectiveBudgetSnapshot{
			BudgetRef: "budget:delegation_parent",
			Limit:     3,
			Remaining: 3,
		},
		Approval: ObjectiveApprovalState{Required: false},
		Strategies: []StrategyCandidate{{
			ID:           "worker:delegation_fixture_readonly",
			Kind:         "delegation_worker_strategy",
			ControlMode:  ControlModeDelegated,
			MinIntensity: IntensityL4DurableLongRun,
			MaxIntensity: IntensityL4DurableLongRun,
			Owner:        "host",
			ExpectedEvidence: []EvidenceRef{{
				Ref:      "evidence:delegation_worker_verified",
				Kind:     "delegation_worker_result",
				Strength: EvidenceAdequate,
			}},
		}},
	})
}
