package hostkit

import (
	"context"
	"testing"

	agentxcontrolplane "github.com/wsnacj/agentx-go/runtime/controlcontract"
)

func TestObjectiveRuntimeClosureProfileRunsBackendAndBuildsRuntimeHandoff(t *testing.T) {
	runtime := &fakeWorkerRuntime{
		result: WorkerResult{
			Completed:            true,
			ObservedWorkerRunRef: "worker_run:delegation_backend",
			WorkerResultRef:      "worker_result:delegation_backend",
			WorkerReadbackRef:    "worker_readback:delegation_backend",
			ObservationRef:       "observation:delegation_backend",
			EvidenceRefs: []agentxcontrolplane.EvidenceRef{{
				Ref:      "evidence:delegation_backend_worker_result",
				Kind:     "delegation_worker_result",
				Strength: agentxcontrolplane.EvidenceStrong,
				Source:   "worker:delegation_backend",
			}},
			Boundaries: []agentxcontrolplane.Boundary{"fake_worker_runtime_completed"},
		},
		readback: WorkerReadback{
			Ready:                true,
			ResultVisible:        true,
			ObservedWorkerRunRef: "worker_run:delegation_backend",
			WorkerResultRef:      "worker_result:delegation_backend",
			WorkerReadbackRef:    "worker_readback:delegation_backend",
			ObservationRef:       "observation:delegation_backend",
			EvidenceRefs: []agentxcontrolplane.EvidenceRef{{
				Ref:      "evidence:delegation_backend_readback",
				Kind:     "delegation_worker_readback",
				Strength: agentxcontrolplane.EvidenceAdequate,
				Source:   "readback:delegation_backend",
			}},
		},
	}
	profile := ObjectiveRuntimeClosureProfile{
		Enabled: true,
		Backend: Backend{
			Runtime:    runtime,
			Store:      NewInMemoryStateStore(),
			BackendRef: "backend:delegation_worker_runtime_test",
			Durable:    true,
		},
		BackendInput: readyInput(false),
		ClosureInput: ObjectiveRuntimeClosureInput{
			Run:                      objectiveRuntimeClosureProfileRun(),
			WorkerAttemptRef:         "attempt:delegation_backend_worker_1",
			MergeRef:                 "merge:delegation_backend",
			HandoffRef:               "handoff:delegation_backend",
			WorkerObservations:       objectiveRuntimeClosureProfileObservations(),
			ExpectedObservationKinds: []string{"delegation_worker_result"},
			Boundaries:               []agentxcontrolplane.Boundary{"test_delegation_worker_objective_closure"},
		},
	}

	report, err := RunObjectiveRuntimeClosureProfile(context.Background(), profile)
	if err != nil {
		t.Fatalf("RunObjectiveRuntimeClosureProfile error = %v", err)
	}
	if !report.Available ||
		!report.Ready ||
		report.Status != "delegation_worker_objective_runtime_closure_ready_for_host_persist" ||
		!report.BackendReady ||
		!report.HostWorkerRuntimeExecuted ||
		!report.WorkerResultRecorded ||
		!report.WorkerResultReadbackReady ||
		!report.ReadyForRuntimeLoopInput ||
		!report.RuntimeLoopReady ||
		!report.RuntimeLoopHostPersistReady ||
		report.WorkerOutputAcceptedAsFact ||
		!report.WorkerResultRequiresVerification {
		t.Fatalf("unexpected profile report: %#v", report)
	}
	if report.WorkerDispatchedByCore ||
		report.RunnerDispatchedByCore ||
		report.RuntimeAdapterExecutedByCore ||
		report.StoreMutationExecutedByCore ||
		report.CoreExecutionExecuted {
		t.Fatalf("profile must not report core side effects: %#v", report)
	}
	if runtime.invokeCalls != 1 || runtime.readbackCalls != 1 {
		t.Fatalf("unexpected worker calls: invoke=%d readback=%d", runtime.invokeCalls, runtime.readbackCalls)
	}
	if report.Closure.RuntimeLoopStatus != "ready_for_host_persist" ||
		report.Closure.RuntimeLoopAction != string(agentxcontrolplane.ObjectiveActionReturnSatisfied) ||
		report.Closure.RuntimeLoopStep.Run.Ledger.Attempts[0].Ref != "attempt:delegation_backend_worker_1" {
		t.Fatalf("objective runtime closure not ready: %#v", report.Closure)
	}
	if report.AsyncCompletion.Status != agentxcontrolplane.VerificationSatisfied ||
		!report.AsyncCompletion.ReadyForResume ||
		!report.AsyncCompletion.Durable ||
		report.AsyncCompletion.ProcessLocal ||
		len(report.AsyncCompletion.CompletedChildRefs) != 1 ||
		report.AsyncCompletion.CompletedChildRefs[0] != "subgoal:delegation_backend_research" ||
		len(report.AsyncCompletion.CompletionEnvelopes) != 1 ||
		report.AsyncCompletion.NextHostAction != "resume_parent_objective_for_delegation_merge" {
		t.Fatalf("async completion projection not ready: %#v", report.AsyncCompletion)
	}
	for _, boundary := range []string{
		"delegation_worker_objective_runtime_closure_profile",
		"profile_delegation_worker_runtime_backend",
		"profile_delegation_worker_objective_runtime_handoff",
		"profile_delegation_worker_async_completion",
		"delegation_worker_objective_runtime_loop_ready_for_host_persist",
		"no_delegation_worker_runtime_by_core",
	} {
		if !stringContains(report.Boundaries, boundary) {
			t.Fatalf("expected boundary %q in %#v", boundary, report.Boundaries)
		}
	}
}

func TestObjectiveRuntimeClosureProfileDefaultOffDoesNotInvokeWorker(t *testing.T) {
	runtime := &fakeWorkerRuntime{}
	report, err := RunObjectiveRuntimeClosureProfile(context.Background(), ObjectiveRuntimeClosureProfile{
		Enabled: false,
		Backend: Backend{
			Runtime: runtime,
			Store:   NewInMemoryStateStore(),
		},
		BackendInput: readyInput(true),
	})
	if err != nil {
		t.Fatalf("RunObjectiveRuntimeClosureProfile error = %v", err)
	}
	if report.Ready ||
		report.Status != "disabled" ||
		runtime.invokeCalls != 0 ||
		!stringContains(report.MissingInputs, "host:delegation_worker_objective_runtime_closure_profile_enabled") ||
		!stringContains(report.Boundaries, "delegation_worker_objective_runtime_closure_profile_default_off") {
		t.Fatalf("expected default-off profile block, got %#v", report)
	}
}

func TestObjectiveRuntimeClosureProfileStopsBeforeClosureWhenBackendBlocked(t *testing.T) {
	runtime := &fakeWorkerRuntime{
		result: WorkerResult{
			Completed:            true,
			ObservedWorkerRunRef: "worker_run:delegation_backend",
			WorkerResultRef:      "worker_result:delegation_backend",
			WorkerReadbackRef:    "worker_readback:delegation_backend",
			ObservationRef:       "observation:delegation_backend",
		},
	}
	report, err := RunObjectiveRuntimeClosureProfile(context.Background(), ObjectiveRuntimeClosureProfile{
		Enabled: true,
		Backend: Backend{
			Runtime: runtime,
			Store:   NewInMemoryStateStore(),
		},
		BackendInput: readyInput(true),
		ClosureInput: ObjectiveRuntimeClosureInput{
			Run:                      objectiveRuntimeClosureProfileRun(),
			WorkerAttemptRef:         "attempt:delegation_backend_worker_1",
			MergeRef:                 "merge:delegation_backend",
			HandoffRef:               "handoff:delegation_backend",
			WorkerObservations:       objectiveRuntimeClosureProfileObservations(),
			ExpectedObservationKinds: []string{"delegation_worker_result"},
		},
	})
	if err != nil {
		t.Fatalf("RunObjectiveRuntimeClosureProfile error = %v", err)
	}
	if report.Ready ||
		report.Status != "delegation_worker_objective_runtime_closure_backend_blocked" ||
		report.WorkerResultReadbackReady ||
		report.ReadyForRuntimeLoopInput ||
		report.Closure.Available ||
		!stringContains(report.BlockedReasons, "delegation_worker_result_evidence_refs_missing") ||
		!stringContains(report.BlockedReasons, "delegation_worker_runtime_backend_not_ready_for_objective_closure") {
		t.Fatalf("expected backend block before closure, got %#v", report)
	}
}

func objectiveRuntimeClosureProfileRun() agentxcontrolplane.ObjectiveRun {
	evidenceRefs := []agentxcontrolplane.EvidenceRef{{
		Ref:      "evidence:delegation_backend_worker_result",
		Kind:     "delegation_worker_result",
		Strength: agentxcontrolplane.EvidenceAdequate,
		Source:   "worker:delegation_backend",
	}}
	return agentxcontrolplane.BuildObjectiveRun(agentxcontrolplane.ObjectiveRunInput{
		Activation: agentxcontrolplane.ActivationManaged,
		Frame: agentxcontrolplane.ObjectiveFrame{
			ID:               "objective:delegation_backend",
			UserGoalDigest:   "delegation backend objective",
			ControlMode:      agentxcontrolplane.ControlModeObjective,
			Intensity:        agentxcontrolplane.IntensityL3ManagedObjective,
			SuccessCriteria:  []string{"delegation worker result satisfies parent evidence"},
			RequiredEvidence: evidenceRefs,
		},
		Ledger: agentxcontrolplane.AttemptLedgerPatch{
			ObjectiveID: "objective:delegation_backend",
			LedgerRef:   "ledger:delegation_backend",
		},
		Budget: agentxcontrolplane.ObjectiveBudgetSnapshot{
			BudgetRef: "budget:delegation_backend",
			Limit:     3,
			Remaining: 3,
		},
		Approval: agentxcontrolplane.ObjectiveApprovalState{Required: false},
		Strategies: []agentxcontrolplane.StrategyCandidate{{
			ID:               "worker:delegation_backend",
			Kind:             "delegation_worker_strategy",
			ControlMode:      agentxcontrolplane.ControlModeDelegated,
			MinIntensity:     agentxcontrolplane.IntensityL4DurableLongRun,
			MaxIntensity:     agentxcontrolplane.IntensityL4DurableLongRun,
			Owner:            "host",
			ExpectedEvidence: evidenceRefs,
		}},
		CurrentStrategyRef: "worker:delegation_backend",
	})
}

func objectiveRuntimeClosureProfileObservations() []agentxcontrolplane.Observation {
	evidenceRefs := []agentxcontrolplane.EvidenceRef{{
		Ref:      "evidence:delegation_backend_worker_result",
		Kind:     "delegation_worker_result",
		Strength: agentxcontrolplane.EvidenceStrong,
		Source:   "worker:delegation_backend",
	}}
	return []agentxcontrolplane.Observation{{
		Kind:            "delegation_worker_result",
		Source:          "worker:delegation_backend",
		Subject:         "objective:delegation_backend",
		Name:            "worker_result_verified",
		Value:           "true",
		Strength:        agentxcontrolplane.EvidenceStrong,
		EvidenceRefs:    evidenceRefs,
		DisplaySafeRefs: []agentxcontrolplane.DisplaySafeRef{"evidence:delegation_backend_worker_result"},
	}}
}
