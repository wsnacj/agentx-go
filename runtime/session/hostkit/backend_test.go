package hostkit

import (
	"context"
	"testing"
	"time"

	agentxcontrolplane "github.com/wsnacj/agentx-go/runtime/controlcontract"
)

func TestBackendRunsWorkerStoresReadbackAndReturnsParentVerificationReview(t *testing.T) {
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
			Boundaries: []agentxcontrolplane.Boundary{"fake_worker_runtime_readback_ready"},
		},
	}
	backend := Backend{
		Runtime:    runtime,
		Store:      NewInMemoryStateStore(),
		BackendRef: "backend:delegation_worker_runtime_test",
		Now:        func() time.Time { return time.Unix(100, 0).UTC() },
	}

	report, err := backend.RunDelegationWorkerRuntime(context.Background(), readyInput(true))
	if err != nil {
		t.Fatalf("RunDelegationWorkerRuntime error = %v", err)
	}
	if report.Status != agentxcontrolplane.HostActionRecorded ||
		!report.ReadyForWorkerResultReview ||
		report.ReadyForFailureReview ||
		!report.WorkerRunAttempted ||
		!report.WorkerResultRecorded ||
		!report.WorkerReadbackAttempted ||
		!report.WorkerResultReadbackReady ||
		report.WorkerOutputAcceptedAsFact ||
		!report.WorkerResultRequiresVerification {
		t.Fatalf("unexpected report: %#v", report)
	}
	if report.Invocation.ResultReview.Status != agentxcontrolplane.VerificationBlocked ||
		report.Invocation.ResultReview.ReadyForParentMerge ||
		!missingContains(report.Invocation.ResultReview.MissingInputs, "host:parent_verification") {
		t.Fatalf("worker result must wait for parent verification: %#v", report.Invocation.ResultReview)
	}
	if report.Invocation.CoreWorkerRuntimeInvoked ||
		report.Invocation.WorkerDispatched ||
		report.Invocation.RunnerDispatched ||
		report.Invocation.RuntimeAdapterExecuted ||
		report.Invocation.WorkflowDispatched ||
		report.Invocation.SchedulerApplied ||
		report.Invocation.InstallerExecuted ||
		report.Invocation.StoreMutationExecuted ||
		report.Invocation.CompensationExecuted {
		t.Fatalf("control plane invocation must not report core side effects: %#v", report.Invocation)
	}
	if runtime.invokeCalls != 1 || runtime.readbackCalls != 1 {
		t.Fatalf("unexpected runtime calls: invoke=%d readback=%d", runtime.invokeCalls, runtime.readbackCalls)
	}
	if !displayRefContains(runtime.lastRequest.VisibleToolRefs, "tool:search") ||
		!displayRefContains(runtime.lastRequest.ContextRefs, "context:delegation_parent_frame") ||
		runtime.lastRequest.BudgetRef != "budget:delegation_worker_runtime" ||
		runtime.lastRequest.TimeoutRef != "timeout:delegation_worker_2m" ||
		runtime.lastRequest.ParallelismRef != "parallelism:delegation_worker_2" {
		t.Fatalf("worker scope not propagated: %#v", runtime.lastRequest)
	}
	if !boundaryContains(report.Boundaries, "host_owned_delegation_worker_runtime_backend") ||
		!boundaryContains(report.Boundaries, "worker_result_readback_verified_by_host_backend") ||
		!boundaryContains(report.Boundaries, "worker_result_requires_parent_verification") {
		t.Fatalf("missing backend boundaries: %#v", report.Boundaries)
	}
}

func TestBackendDefaultOffDoesNotInvokeWorker(t *testing.T) {
	runtime := &fakeWorkerRuntime{}
	backend := Backend{Runtime: runtime, Store: NewInMemoryStateStore()}

	report, err := backend.RunDelegationWorkerRuntime(context.Background(), readyInput(false))
	if err != nil {
		t.Fatalf("RunDelegationWorkerRuntime error = %v", err)
	}
	if report.Status == agentxcontrolplane.HostActionRecorded ||
		report.WorkerRunAttempted ||
		runtime.invokeCalls != 0 ||
		!missingContains(report.MissingInputs, "host:delegation_worker_runtime_enablement") ||
		!stringContains(report.BlockedReasons, "delegation_worker_runtime_backend_not_enabled") {
		t.Fatalf("expected default-off block without worker invocation, got %#v", report)
	}
}

func TestBackendBlocksCompletedWorkerResultWithoutEvidence(t *testing.T) {
	runtime := &fakeWorkerRuntime{
		result: WorkerResult{
			Completed:            true,
			ObservedWorkerRunRef: "worker_run:delegation_backend",
			WorkerResultRef:      "worker_result:delegation_backend",
			WorkerReadbackRef:    "worker_readback:delegation_backend",
			ObservationRef:       "observation:delegation_backend",
		},
	}
	backend := Backend{Runtime: runtime, Store: NewInMemoryStateStore()}

	report, err := backend.RunDelegationWorkerRuntime(context.Background(), readyInput(true))
	if err != nil {
		t.Fatalf("RunDelegationWorkerRuntime error = %v", err)
	}
	if report.Status == agentxcontrolplane.HostActionRecorded ||
		!report.WorkerRunAttempted ||
		report.WorkerResultRecorded ||
		report.WorkerReadbackAttempted ||
		!missingContains(report.MissingInputs, "host:delegation_worker_result_evidence_refs") ||
		!stringContains(report.BlockedReasons, "delegation_worker_result_evidence_refs_missing") {
		t.Fatalf("expected evidence block, got %#v", report)
	}
}

func TestBackendBlocksCompletedWorkerResultMissingRefsBeforeStoreReadback(t *testing.T) {
	runtime := &fakeWorkerRuntime{
		result: WorkerResult{
			Completed:            true,
			ObservedWorkerRunRef: "worker_run:delegation_backend",
			WorkerResultRef:      "worker_result:delegation_backend",
			EvidenceRefs: []agentxcontrolplane.EvidenceRef{{
				Ref:      "evidence:delegation_backend_worker_result",
				Kind:     "delegation_worker_result",
				Strength: agentxcontrolplane.EvidenceAdequate,
				Source:   "worker:delegation_backend",
			}},
		},
	}
	backend := Backend{Runtime: runtime, Store: NewInMemoryStateStore()}

	report, err := backend.RunDelegationWorkerRuntime(context.Background(), readyInput(true))
	if err != nil {
		t.Fatalf("RunDelegationWorkerRuntime error = %v", err)
	}
	if report.Status == agentxcontrolplane.HostActionRecorded ||
		!report.WorkerRunAttempted ||
		report.WorkerResultRecorded ||
		report.WorkerReadbackAttempted ||
		!missingContains(report.MissingInputs, "host:delegation_worker_readback_ref") ||
		!missingContains(report.MissingInputs, "host:delegation_worker_observation_ref") ||
		!stringContains(report.BlockedReasons, "delegation_worker_readback_ref_missing") {
		t.Fatalf("expected missing result refs block before readback, got %#v", report)
	}
}

func TestBackendRecordsWorkerFailureForFailureReview(t *testing.T) {
	runtime := &fakeWorkerRuntime{
		result: WorkerResult{
			Failed:               true,
			ObservedWorkerRunRef: "worker_run:delegation_backend",
			FailureRef:           "failure:delegation_worker_runtime_failed",
			CompensationRef:      "compensation:delegation_worker_runtime_failed",
			EvidenceRefs: []agentxcontrolplane.EvidenceRef{{
				Ref:      "evidence:delegation_worker_runtime_failure",
				Kind:     "delegation_worker_failure",
				Strength: agentxcontrolplane.EvidenceAdequate,
				Source:   "worker:delegation_backend",
			}},
		},
	}
	backend := Backend{Runtime: runtime, Store: NewInMemoryStateStore()}

	report, err := backend.RunDelegationWorkerRuntime(context.Background(), readyInput(true))
	if err != nil {
		t.Fatalf("RunDelegationWorkerRuntime error = %v", err)
	}
	if report.Status != agentxcontrolplane.HostActionRecorded ||
		report.ReadyForWorkerResultReview ||
		!report.ReadyForFailureReview ||
		report.WorkerResultRecorded ||
		report.WorkerReadbackAttempted ||
		report.Invocation.FailureRef != "failure:delegation_worker_runtime_failed" ||
		report.Invocation.CompensationRef != "compensation:delegation_worker_runtime_failed" {
		t.Fatalf("unexpected failure report: %#v", report)
	}
}

func TestBackendRejectsUnsafeRefsBeforeWorkerInvocation(t *testing.T) {
	runtime := &fakeWorkerRuntime{}
	backend := Backend{Runtime: runtime, Store: NewInMemoryStateStore()}
	input := readyInput(true)
	input.VisibleToolRefs = []agentxcontrolplane.DisplaySafeRef{"http://localhost/raw-tool"}

	report, err := backend.RunDelegationWorkerRuntime(context.Background(), input)
	if err != nil {
		t.Fatalf("RunDelegationWorkerRuntime error = %v", err)
	}
	if !report.RawOutputLoaded ||
		report.WorkerRunAttempted ||
		runtime.invokeCalls != 0 ||
		!missingContains(report.MissingInputs, "host:display_safe_refs") ||
		!boundaryContains(report.Boundaries, "raw_output_not_allowed") {
		t.Fatalf("expected unsafe ref block before invocation, got %#v", report)
	}
}

type fakeWorkerRuntime struct {
	result        WorkerResult
	readback      WorkerReadback
	invokeCalls   int
	readbackCalls int
	lastRequest   WorkerRequest
}

func (f *fakeWorkerRuntime) InvokeDelegationWorker(_ context.Context, request WorkerRequest) (WorkerResult, error) {
	f.invokeCalls++
	f.lastRequest = request
	return f.result, nil
}

func (f *fakeWorkerRuntime) ReadDelegationWorkerResult(_ context.Context, _ WorkerReadbackRequest) (WorkerReadback, error) {
	f.readbackCalls++
	return f.readback, nil
}

func readyInput(enabled bool) BackendInput {
	return BackendInput{
		Enabled:                 enabled,
		Readiness:               readyReadiness(),
		InvocationReportRef:     "invocation_report:delegation_worker_runtime_backend",
		HostWorkerRuntimeRunRef: "worker_runtime_run:delegation_backend",
		VisibleToolRefs:         []agentxcontrolplane.DisplaySafeRef{"tool:search", "tool:read"},
		ContextRefs:             []agentxcontrolplane.DisplaySafeRef{"context:delegation_parent_frame", "context:delegation_subgoal"},
		TimeoutRef:              "timeout:delegation_worker_2m",
		ParallelismRef:          "parallelism:delegation_worker_2",
		FailureRef:              "failure:delegation_worker_runtime",
		CompensationRef:         "compensation:delegation_worker_runtime",
	}
}

func readyReadiness() agentxcontrolplane.HostOwnedDelegationWorkerRuntimeReadiness {
	return agentxcontrolplane.BuildHostOwnedDelegationWorkerRuntimeReadiness(agentxcontrolplane.HostOwnedDelegationWorkerRuntimeReadinessInput{
		Request: agentxcontrolplane.BuildDelegationRequestProjection(agentxcontrolplane.DelegationRequestInput{
			Activation:         agentxcontrolplane.ActivationManaged,
			RequestedIntensity: agentxcontrolplane.IntensityL4DurableLongRun,
			Frame: agentxcontrolplane.ObjectiveFrame{
				ID:             "objective:delegation_backend",
				UserGoalDigest: "delegation backend objective",
				ControlMode:    agentxcontrolplane.ControlModeDelegated,
				Intensity:      agentxcontrolplane.IntensityL4DurableLongRun,
				SuccessCriteria: []string{
					"worker result is verified before parent merge",
				},
				RequiredEvidence: []agentxcontrolplane.EvidenceRef{{
					Ref:      "evidence:delegation_backend_worker_result",
					Kind:     "delegation_worker_result",
					Strength: agentxcontrolplane.EvidenceAdequate,
					Source:   "worker:delegation_backend",
				}},
			},
			SubgoalRef:                        "subgoal:delegation_backend_research",
			WorkerRef:                         "worker:delegation_backend_readonly",
			AllowedToolRefs:                   []agentxcontrolplane.DisplaySafeRef{"tool:read", "tool:search"},
			DeniedToolRefs:                    []agentxcontrolplane.DisplaySafeRef{"tool:write"},
			Budget:                            agentxcontrolplane.ObjectiveBudgetSnapshot{BudgetRef: "budget:delegation_backend", Limit: 2, Remaining: 2},
			EvidenceRequirements:              []agentxcontrolplane.EvidenceRef{{Ref: "evidence:delegation_backend_worker_result", Kind: "delegation_worker_result", Strength: agentxcontrolplane.EvidenceAdequate, Source: "worker:delegation_backend"}},
			StopConditionRefs:                 []agentxcontrolplane.DisplaySafeRef{"stop:delegation_backend_max_2_workers", "stop:delegation_backend_parent_verified"},
			RedactionPolicyRef:                "redaction:delegation_backend",
			MergePolicyRef:                    "merge:delegation_backend_verify_before_merge",
			ExecutionContractAllowsDelegation: true,
			HostAllowsL4Delegation:            true,
			UserConfirmed:                     true,
			HostApproved:                      true,
			ApprovalRefs:                      []agentxcontrolplane.DisplaySafeRef{"approval:delegation_backend"},
			PolicyRefs:                        []agentxcontrolplane.DisplaySafeRef{"policy:delegation_backend"},
		}),
		WorkerRuntimeGate: agentxcontrolplane.BuildProductionAdapterIndependentEffectGate(agentxcontrolplane.ProductionAdapterIndependentEffectGateSpec{
			Kind:                  agentxcontrolplane.ProductionAdapterEffectGateDelegationWorker,
			GateRef:               "gate:delegation_worker_runtime",
			AdapterRef:            "adapter:delegation_worker_runtime",
			ContractRef:           "contract:delegation_worker_runtime",
			PolicyRef:             "policy:delegation_worker_runtime",
			ApprovalRef:           "approval:delegation_worker_runtime",
			BudgetRef:             "budget:delegation_worker_runtime",
			IdempotencyRef:        "idempotency:delegation_worker_runtime",
			ReadbackRef:           "readback:delegation_worker_runtime",
			EvalRef:               "eval:delegation_worker_runtime",
			FailureReviewRef:      "review:delegation_worker_runtime_failure",
			CompensationReviewRef: "review:delegation_worker_runtime_compensation",
			EvidenceRefs:          []agentxcontrolplane.DisplaySafeRef{"evidence:delegation_worker_runtime_gate"},
		}),
		AdapterRef:           "adapter:delegation_worker_runtime",
		AdapterVersionRef:    "adapter_version:delegation_worker_runtime_v1",
		AdapterCapabilityRef: "capability:delegation_worker_runtime",
		AdapterContractRef:   "contract:delegation_worker_runtime",
		HostConfirmationRef:  "confirmation:delegation_worker_runtime",
		WorkerRunRef:         "worker_run:delegation_backend",
		WorkerRequestRef:     "worker_request:delegation_backend",
		InvocationRef:        "invocation:delegation_worker_runtime_backend",
		ResultBindingRef:     "worker_result:delegation_backend",
		ReadbackBindingRef:   "worker_readback:delegation_backend",
		IdempotencyRef:       "idempotency:delegation_worker_runtime",
		BudgetRef:            "budget:delegation_worker_runtime",
		VerificationRef:      "verification:delegation_worker_parent",
		FailureBindingRef:    "failure:delegation_worker_runtime",
		CompensationRef:      "compensation:delegation_worker_runtime",
	})
}

func missingContains(values []agentxcontrolplane.MissingInput, needle agentxcontrolplane.MissingInput) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

func stringContains(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

func boundaryContains(values []agentxcontrolplane.Boundary, needle agentxcontrolplane.Boundary) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

func displayRefContains(values []agentxcontrolplane.DisplaySafeRef, needle agentxcontrolplane.DisplaySafeRef) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}
