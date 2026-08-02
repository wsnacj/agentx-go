package hostkit

import (
	"testing"

	agentxcontrolplane "github.com/wsnacj/agentx-go/runtime/controlcontract"
)

func TestDelegationWorkerObjectiveRuntimeHandoffFeedsSatisfiedLoop(t *testing.T) {
	report := BuildDelegationWorkerObjectiveRuntimeHandoff(DelegationWorkerObjectiveRuntimeHandoffInput{
		Run:              delegationWorkerObjectiveTestRun(),
		Invocation:       delegationWorkerObjectiveTestInvocation(),
		WorkerAttemptRef: "attempt:delegation_worker_closure_1",
		MergeRef:         "merge:delegation_worker_closure",
		HandoffRef:       "handoff:delegation_worker_closure",
		WorkerObservations: []agentxcontrolplane.Observation{{
			Kind:     "delegation_worker_result",
			Source:   "worker:delegation_worker_closure",
			Subject:  "objective:delegation_worker_closure",
			Name:     "worker_result_verified",
			Value:    "true",
			Strength: agentxcontrolplane.EvidenceStrong,
			EvidenceRefs: []agentxcontrolplane.EvidenceRef{{
				Ref:      "evidence:delegation_worker_closure",
				Kind:     "delegation_worker_result",
				Strength: agentxcontrolplane.EvidenceStrong,
				Source:   "worker:delegation_worker_closure",
			}},
			DisplaySafeRefs: []agentxcontrolplane.DisplaySafeRef{"evidence:delegation_worker_closure"},
		}},
		ExpectedObservationKinds: []string{"delegation_worker_result"},
		Boundaries:               []agentxcontrolplane.Boundary{"hostruntime_delegation_worker_closure_test"},
	})

	if report.Status != "ready_for_host_persist" ||
		report.ParentMergeStatus != string(agentxcontrolplane.VerificationSatisfied) ||
		report.RuntimeHandoffStatus != string(agentxcontrolplane.HostActionRecorded) ||
		report.RuntimeLoopStatus != "ready_for_host_persist" ||
		report.RuntimeLoopState != string(agentxcontrolplane.ObjectiveControllerSatisfied) ||
		report.RuntimeLoopAction != string(agentxcontrolplane.ObjectiveActionReturnSatisfied) ||
		!report.ReadyForParentMerge ||
		!report.ReadyForRuntimeLoopInput ||
		!report.RuntimeLoopReady ||
		!report.RuntimeLoopHostPersistReady ||
		report.ReadyForNextRuntimeAction ||
		report.WorkerOutputAcceptedAsFact ||
		!report.WorkerResultRequiresVerification {
		t.Fatalf("unexpected delegation objective handoff report: %#v", report)
	}
	if report.WorkerDispatchedByCore ||
		report.RunnerDispatchedByCore ||
		report.RuntimeAdapterExecutedByCore ||
		report.StoreMutationExecutedByCore ||
		report.CoreExecutionExecuted {
		t.Fatalf("hostruntime handoff must remain side-effect free: %#v", report)
	}
	if len(report.RuntimeLoopStep.Run.Ledger.Attempts) != 1 ||
		report.RuntimeLoopStep.Run.Ledger.Attempts[0].Ref != "attempt:delegation_worker_closure_1" {
		t.Fatalf("delegation worker attempt was not merged into parent ledger: %#v", report.RuntimeLoopStep.Run.Ledger.Attempts)
	}
	for _, boundary := range []string{
		"hostruntime_delegation_worker_objective_handoff",
		"delegation_parent_merge_ready_for_runtime_loop",
		"ready_for_objective_runtime_loop",
		"objective_runtime_loop_step",
		"delegation_worker_runtime_loop_ready_for_host_persist",
		"no_worker_dispatch",
		"no_store_mutation_by_core",
	} {
		if !hostruntimeStringContains(report.Boundaries, boundary) {
			t.Fatalf("expected boundary %q in %#v", boundary, report.Boundaries)
		}
	}
}

func TestDelegationWorkerObjectiveRuntimeHandoffBlocksMissingObservation(t *testing.T) {
	report := BuildDelegationWorkerObjectiveRuntimeHandoff(DelegationWorkerObjectiveRuntimeHandoffInput{
		Run:                      delegationWorkerObjectiveTestRun(),
		Invocation:               delegationWorkerObjectiveTestInvocation(),
		WorkerAttemptRef:         "attempt:delegation_worker_closure_1",
		MergeRef:                 "merge:delegation_worker_closure",
		HandoffRef:               "handoff:delegation_worker_closure",
		ExpectedObservationKinds: []string{"delegation_worker_result"},
	})

	if report.Status == "ready_for_host_persist" ||
		report.ReadyForParentMerge ||
		report.ReadyForRuntimeLoopInput ||
		report.RuntimeLoopHostPersistReady ||
		report.WorkerOutputAcceptedAsFact ||
		!hostruntimeStringContains(report.MissingInputs, "host:delegation_worker_observations") ||
		!hostruntimeStringContains(report.BlockedReasons, "delegation_worker_observations_missing") {
		t.Fatalf("expected missing observation block, got %#v", report)
	}
}

func TestDelegationWorkerObjectiveRuntimeHandoffRawOutputForcesReview(t *testing.T) {
	report := BuildDelegationWorkerObjectiveRuntimeHandoff(DelegationWorkerObjectiveRuntimeHandoffInput{
		Run:              delegationWorkerObjectiveTestRun(),
		Invocation:       delegationWorkerObjectiveTestInvocation(),
		WorkerAttemptRef: "attempt:delegation_worker_closure_1",
		MergeRef:         "/Users/example/raw-worker-result",
		HandoffRef:       "handoff:delegation_worker_closure",
		WorkerObservations: []agentxcontrolplane.Observation{{
			Kind:            "delegation_worker_result",
			Source:          "worker:delegation_worker_closure",
			Strength:        agentxcontrolplane.EvidenceStrong,
			RawOutputLoaded: true,
			DisplaySafeRefs: []agentxcontrolplane.DisplaySafeRef{"/Users/example/raw-worker-result"},
		}},
		RawOutputLoaded: true,
	})

	if report.Status != "review_required" ||
		report.ReadyForParentMerge ||
		report.ReadyForRuntimeLoopInput ||
		report.RuntimeLoopHostPersistReady ||
		!report.RawOutputLoaded ||
		!hostruntimeStringContains(report.MissingInputs, "host:display_safe_refs") ||
		!hostruntimeStringContains(report.Boundaries, "raw_output_not_allowed") {
		t.Fatalf("expected raw-output review, got %#v", report)
	}
}

func delegationWorkerObjectiveTestRun() agentxcontrolplane.ObjectiveRun {
	return agentxcontrolplane.BuildObjectiveRun(agentxcontrolplane.ObjectiveRunInput{
		Activation: agentxcontrolplane.ActivationManaged,
		Frame: agentxcontrolplane.ObjectiveFrame{
			ID:              "objective:delegation_worker_closure",
			UserGoalDigest:  "delegation worker closure objective",
			ControlMode:     agentxcontrolplane.ControlModeObjective,
			Intensity:       agentxcontrolplane.IntensityL3ManagedObjective,
			SuccessCriteria: []string{"delegation worker result satisfies parent evidence"},
			RequiredEvidence: []agentxcontrolplane.EvidenceRef{{
				Ref:      "evidence:delegation_worker_closure",
				Kind:     "delegation_worker_result",
				Strength: agentxcontrolplane.EvidenceAdequate,
				Source:   "worker:delegation_worker_closure",
			}},
		},
		Ledger: agentxcontrolplane.AttemptLedgerPatch{
			ObjectiveID: "objective:delegation_worker_closure",
			LedgerRef:   "ledger:delegation_worker_closure",
		},
		Budget: agentxcontrolplane.ObjectiveBudgetSnapshot{
			BudgetRef: "budget:delegation_worker_closure",
			Limit:     3,
			Remaining: 3,
		},
		Approval: agentxcontrolplane.ObjectiveApprovalState{Required: false},
		Strategies: []agentxcontrolplane.StrategyCandidate{{
			ID:           "worker:delegation_worker_closure",
			Kind:         "delegation_worker_strategy",
			ControlMode:  agentxcontrolplane.ControlModeDelegated,
			MinIntensity: agentxcontrolplane.IntensityL4DurableLongRun,
			MaxIntensity: agentxcontrolplane.IntensityL4DurableLongRun,
			Owner:        "host",
			ExpectedEvidence: []agentxcontrolplane.EvidenceRef{{
				Ref:      "evidence:delegation_worker_closure",
				Kind:     "delegation_worker_result",
				Strength: agentxcontrolplane.EvidenceAdequate,
				Source:   "worker:delegation_worker_closure",
			}},
		}},
		CurrentStrategyRef: "worker:delegation_worker_closure",
	})
}

func delegationWorkerObjectiveTestInvocation() agentxcontrolplane.HostOwnedDelegationWorkerRuntimeInvocation {
	return agentxcontrolplane.BuildHostOwnedDelegationWorkerRuntimeInvocation(agentxcontrolplane.HostOwnedDelegationWorkerRuntimeInvocationInput{
		Readiness:               delegationWorkerObjectiveTestReadiness(),
		InvocationReportRef:     "invocation_report:delegation_worker_closure",
		ObservedInvocationRef:   "invocation:delegation_worker_closure",
		HostWorkerRuntimeRunRef: "worker_runtime_run:delegation_worker_closure",
		ObservedWorkerRunRef:    "worker_run:delegation_worker_closure",
		WorkerResultRef:         "worker_result:delegation_worker_closure",
		WorkerReadbackRef:       "worker_readback:delegation_worker_closure",
		ObservationRef:          "observation:delegation_worker_closure",
		HostInvocationReported:  true,
		HostInvocationCompleted: true,
		EvidenceRefs: []agentxcontrolplane.EvidenceRef{{
			Ref:      "evidence:delegation_worker_closure",
			Kind:     "delegation_worker_result",
			Strength: agentxcontrolplane.EvidenceStrong,
			Source:   "worker:delegation_worker_closure",
		}},
		Boundaries: []agentxcontrolplane.Boundary{"host_owned_delegation_worker_runtime_backend"},
	})
}

func delegationWorkerObjectiveTestReadiness() agentxcontrolplane.HostOwnedDelegationWorkerRuntimeReadiness {
	return agentxcontrolplane.BuildHostOwnedDelegationWorkerRuntimeReadiness(agentxcontrolplane.HostOwnedDelegationWorkerRuntimeReadinessInput{
		Request: agentxcontrolplane.BuildDelegationRequestProjection(agentxcontrolplane.DelegationRequestInput{
			Activation:         agentxcontrolplane.ActivationManaged,
			RequestedIntensity: agentxcontrolplane.IntensityL4DurableLongRun,
			Frame: agentxcontrolplane.ObjectiveFrame{
				ID:              "objective:delegation_worker_closure",
				UserGoalDigest:  "delegation worker closure objective",
				ControlMode:     agentxcontrolplane.ControlModeDelegated,
				Intensity:       agentxcontrolplane.IntensityL4DurableLongRun,
				SuccessCriteria: []string{"delegation worker result satisfies parent evidence"},
				RequiredEvidence: []agentxcontrolplane.EvidenceRef{{
					Ref:      "evidence:delegation_worker_closure",
					Kind:     "delegation_worker_result",
					Strength: agentxcontrolplane.EvidenceAdequate,
					Source:   "worker:delegation_worker_closure",
				}},
			},
			SubgoalRef:                        "subgoal:delegation_worker_closure",
			WorkerRef:                         "worker:delegation_worker_closure",
			AllowedToolRefs:                   []agentxcontrolplane.DisplaySafeRef{"tool:read"},
			DeniedToolRefs:                    []agentxcontrolplane.DisplaySafeRef{"tool:write"},
			Budget:                            agentxcontrolplane.ObjectiveBudgetSnapshot{BudgetRef: "budget:delegation_worker_closure", Limit: 1, Remaining: 1},
			EvidenceRequirements:              []agentxcontrolplane.EvidenceRef{{Ref: "evidence:delegation_worker_closure", Kind: "delegation_worker_result", Strength: agentxcontrolplane.EvidenceAdequate, Source: "worker:delegation_worker_closure"}},
			StopConditionRefs:                 []agentxcontrolplane.DisplaySafeRef{"stop:delegation_worker_closure_parent_verified"},
			RedactionPolicyRef:                "redaction:delegation_worker_closure",
			MergePolicyRef:                    "merge:delegation_worker_closure",
			ExecutionContractAllowsDelegation: true,
			HostAllowsL4Delegation:            true,
			UserConfirmed:                     true,
			HostApproved:                      true,
			ApprovalRefs:                      []agentxcontrolplane.DisplaySafeRef{"approval:delegation_worker_closure"},
			PolicyRefs:                        []agentxcontrolplane.DisplaySafeRef{"policy:delegation_worker_closure"},
		}),
		WorkerRuntimeGate: agentxcontrolplane.BuildProductionAdapterIndependentEffectGate(agentxcontrolplane.ProductionAdapterIndependentEffectGateSpec{
			Kind:                  agentxcontrolplane.ProductionAdapterEffectGateDelegationWorker,
			GateRef:               "gate:delegation_worker_closure",
			AdapterRef:            "adapter:delegation_worker_closure",
			ContractRef:           "contract:delegation_worker_closure",
			PolicyRef:             "policy:delegation_worker_closure",
			ApprovalRef:           "approval:delegation_worker_closure",
			BudgetRef:             "budget:delegation_worker_closure",
			IdempotencyRef:        "idempotency:delegation_worker_closure",
			ReadbackRef:           "readback:delegation_worker_closure",
			EvalRef:               "eval:delegation_worker_closure",
			FailureReviewRef:      "review:delegation_worker_closure_failure",
			CompensationReviewRef: "review:delegation_worker_closure_compensation",
			EvidenceRefs:          []agentxcontrolplane.DisplaySafeRef{"evidence:delegation_worker_closure_gate"},
		}),
		AdapterRef:           "adapter:delegation_worker_closure",
		AdapterVersionRef:    "adapter_version:delegation_worker_closure_v1",
		AdapterCapabilityRef: "capability:delegation_worker_closure",
		AdapterContractRef:   "contract:delegation_worker_closure",
		HostConfirmationRef:  "confirmation:delegation_worker_closure",
		WorkerRunRef:         "worker_run:delegation_worker_closure",
		WorkerRequestRef:     "worker_request:delegation_worker_closure",
		InvocationRef:        "invocation:delegation_worker_closure",
		ResultBindingRef:     "worker_result:delegation_worker_closure",
		ReadbackBindingRef:   "worker_readback:delegation_worker_closure",
		IdempotencyRef:       "idempotency:delegation_worker_closure",
		BudgetRef:            "budget:delegation_worker_closure",
		VerificationRef:      "verification:delegation_worker_closure_parent",
		FailureBindingRef:    "failure:delegation_worker_closure",
		CompensationRef:      "compensation:delegation_worker_closure",
	})
}

func hostruntimeStringContains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
