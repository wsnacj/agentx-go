package hostkit_test

import (
	"context"
	"errors"
	"testing"

	controlcontract "github.com/wsnacj/agentx-go/runtime/controlcontract"
	sessionhostkit "github.com/wsnacj/agentx-go/runtime/session/hostkit"
)

func TestRuntimeSuccessShutdownAndClosedContract(t *testing.T) {
	worker := &stubWorker{}
	runtime, err := sessionhostkit.New(sessionhostkit.Config{
		Worker:     worker,
		Store:      sessionhostkit.NewInMemoryStateStore(),
		BackendRef: "backend:external_session_hostkit",
		Durable:    true,
	})
	if err != nil {
		t.Fatal(err)
	}

	result, err := runtime.Run(context.Background(), sessionhostkit.RunRequest{BackendInput: readyInput()})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Completed || !result.Backend.WorkerResultReadbackReady {
		t.Fatalf("unexpected result: %#v", result)
	}
	if worker.invokeCalls != 1 || worker.readbackCalls != 1 {
		t.Fatalf("worker calls = invoke:%d readback:%d, want 1/1", worker.invokeCalls, worker.readbackCalls)
	}
	if result.Backend.WorkerOutputAcceptedAsFact || !result.Backend.WorkerResultRequiresVerification {
		t.Fatalf("child output bypassed parent verification: %#v", result.Backend)
	}

	if err := runtime.Shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Shutdown(context.Background()); err != nil {
		t.Fatalf("second shutdown: %v", err)
	}
	if _, err := runtime.Run(context.Background(), sessionhostkit.RunRequest{BackendInput: readyInput()}); !errors.Is(err, sessionhostkit.ErrClosed) {
		t.Fatalf("Run after Shutdown error = %v, want ErrClosed", err)
	}
}

func TestRuntimeCancellationAndReadbackMismatchFailClosed(t *testing.T) {
	worker := &stubWorker{}
	runtime, err := sessionhostkit.New(sessionhostkit.Config{
		Worker: worker,
		Store:  sessionhostkit.NewInMemoryStateStore(),
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result, err := runtime.Run(ctx, sessionhostkit.RunRequest{BackendInput: readyInput()})
	if err != nil {
		t.Fatal(err)
	}
	if result.Completed || !result.Backend.ReadyForFailureReview {
		t.Fatalf("cancelled worker must fail closed: %#v", result)
	}

	mismatch := &stubWorker{readbackMismatch: true}
	runtime, err = sessionhostkit.New(sessionhostkit.Config{
		Worker: mismatch,
		Store:  sessionhostkit.NewInMemoryStateStore(),
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err = runtime.Run(context.Background(), sessionhostkit.RunRequest{BackendInput: readyInput()})
	if err != nil {
		t.Fatal(err)
	}
	if result.Completed || result.Backend.WorkerResultReadbackReady || result.Backend.FailureClass != controlcontract.FailureVerificationFailed {
		t.Fatalf("mismatched readback must fail closed: %#v", result)
	}
}

func TestNewRejectsMissingAndTypedNilPorts(t *testing.T) {
	if _, err := sessionhostkit.New(sessionhostkit.Config{}); !errors.Is(err, sessionhostkit.ErrInvalidConfig) {
		t.Fatalf("New error = %v, want ErrInvalidConfig", err)
	}
	var worker *stubWorker
	if _, err := sessionhostkit.New(sessionhostkit.Config{Worker: worker, Store: sessionhostkit.NewInMemoryStateStore()}); !errors.Is(err, sessionhostkit.ErrInvalidConfig) {
		t.Fatalf("typed-nil worker error = %v, want ErrInvalidConfig", err)
	}
}

type stubWorker struct {
	readbackMismatch bool
	invokeCalls      int
	readbackCalls    int
}

func (w *stubWorker) InvokeDelegationWorker(ctx context.Context, request sessionhostkit.WorkerRequest) (sessionhostkit.WorkerResult, error) {
	w.invokeCalls++
	if err := ctx.Err(); err != nil {
		return sessionhostkit.WorkerResult{}, err
	}
	return sessionhostkit.WorkerResult{
		Completed:            true,
		ObservedWorkerRunRef: request.WorkerRunRef,
		WorkerResultRef:      "worker_result:external_session_hostkit",
		WorkerReadbackRef:    "worker_readback:external_session_hostkit",
		ObservationRef:       "observation:external_session_hostkit",
		EvidenceRefs: []controlcontract.EvidenceRef{{
			Ref:      "evidence:external_session_hostkit",
			Kind:     "delegation_worker_result",
			Strength: controlcontract.EvidenceAdequate,
			Source:   "worker:external_session_hostkit",
		}},
	}, nil
}

func (w *stubWorker) ReadDelegationWorkerResult(_ context.Context, request sessionhostkit.WorkerReadbackRequest) (sessionhostkit.WorkerReadback, error) {
	w.readbackCalls++
	workerRunRef := request.WorkerRunRef
	if w.readbackMismatch {
		workerRunRef = "worker_run:mismatch"
	}
	return sessionhostkit.WorkerReadback{
		Ready:                true,
		ResultVisible:        true,
		ObservedWorkerRunRef: workerRunRef,
		WorkerResultRef:      request.WorkerResultRef,
		WorkerReadbackRef:    request.WorkerReadbackRef,
		ObservationRef:       "observation:external_session_hostkit",
		EvidenceRefs: []controlcontract.EvidenceRef{{
			Ref:      "evidence:external_session_hostkit_readback",
			Kind:     "delegation_worker_readback",
			Strength: controlcontract.EvidenceAdequate,
			Source:   "readback:external_session_hostkit",
		}},
	}, nil
}

func readyInput() sessionhostkit.BackendInput {
	readiness := controlcontract.BuildHostOwnedDelegationWorkerRuntimeReadiness(controlcontract.HostOwnedDelegationWorkerRuntimeReadinessInput{
		Request: controlcontract.BuildDelegationRequestProjection(controlcontract.DelegationRequestInput{
			Activation:         controlcontract.ActivationManaged,
			RequestedIntensity: controlcontract.IntensityL4DurableLongRun,
			Frame: controlcontract.ObjectiveFrame{
				ID:              "objective:external_session_hostkit",
				UserGoalDigest:  "sha256:external_session_hostkit",
				ControlMode:     controlcontract.ControlModeDelegated,
				Intensity:       controlcontract.IntensityL4DurableLongRun,
				SuccessCriteria: []string{"child result is verified before parent merge"},
				RequiredEvidence: []controlcontract.EvidenceRef{{
					Ref:      "evidence:external_session_hostkit",
					Kind:     "delegation_worker_result",
					Strength: controlcontract.EvidenceAdequate,
					Source:   "worker:external_session_hostkit",
				}},
			},
			SubgoalRef:                        "subgoal:external_session_hostkit",
			WorkerRef:                         "worker:external_session_hostkit",
			AllowedToolRefs:                   []controlcontract.DisplaySafeRef{"tool:read"},
			Budget:                            controlcontract.ObjectiveBudgetSnapshot{BudgetRef: "budget:external_session_hostkit", Limit: 1, Remaining: 1},
			EvidenceRequirements:              []controlcontract.EvidenceRef{{Ref: "evidence:external_session_hostkit", Kind: "delegation_worker_result", Strength: controlcontract.EvidenceAdequate, Source: "worker:external_session_hostkit"}},
			StopConditionRefs:                 []controlcontract.DisplaySafeRef{"stop:external_session_hostkit_verified"},
			RedactionPolicyRef:                "redaction:external_session_hostkit",
			MergePolicyRef:                    "merge:external_session_hostkit",
			ExecutionContractAllowsDelegation: true,
			HostAllowsL4Delegation:            true,
			UserConfirmed:                     true,
			HostApproved:                      true,
			ApprovalRefs:                      []controlcontract.DisplaySafeRef{"approval:external_session_hostkit"},
			PolicyRefs:                        []controlcontract.DisplaySafeRef{"policy:external_session_hostkit"},
		}),
		WorkerRuntimeGate: controlcontract.BuildProductionAdapterIndependentEffectGate(controlcontract.ProductionAdapterIndependentEffectGateSpec{
			Kind:                  controlcontract.ProductionAdapterEffectGateDelegationWorker,
			GateRef:               "gate:external_session_hostkit",
			AdapterRef:            "adapter:external_session_hostkit",
			ContractRef:           "contract:external_session_hostkit",
			PolicyRef:             "policy:external_session_hostkit",
			ApprovalRef:           "approval:external_session_hostkit",
			BudgetRef:             "budget:external_session_hostkit",
			IdempotencyRef:        "idempotency:external_session_hostkit",
			ReadbackRef:           "readback:external_session_hostkit",
			EvalRef:               "eval:external_session_hostkit",
			FailureReviewRef:      "review:external_session_hostkit",
			CompensationReviewRef: "review:external_session_hostkit_compensation",
			EvidenceRefs:          []controlcontract.DisplaySafeRef{"evidence:external_session_hostkit"},
		}),
		AdapterRef:           "adapter:external_session_hostkit",
		AdapterVersionRef:    "adapter_version:external_session_hostkit_v1",
		AdapterCapabilityRef: "capability:external_session_hostkit",
		AdapterContractRef:   "contract:external_session_hostkit",
		HostConfirmationRef:  "confirmation:external_session_hostkit",
		WorkerRunRef:         "worker_run:external_session_hostkit",
		WorkerRequestRef:     "worker_request:external_session_hostkit",
		InvocationRef:        "invocation:external_session_hostkit",
		ResultBindingRef:     "worker_result:external_session_hostkit",
		ReadbackBindingRef:   "worker_readback:external_session_hostkit",
		IdempotencyRef:       "idempotency:external_session_hostkit",
		BudgetRef:            "budget:external_session_hostkit",
		VerificationRef:      "verification:external_session_hostkit",
		FailureBindingRef:    "failure:external_session_hostkit",
		CompensationRef:      "compensation:external_session_hostkit",
	})
	return sessionhostkit.BackendInput{
		Enabled:                 true,
		Readiness:               readiness,
		InvocationReportRef:     "invocation_report:external_session_hostkit",
		HostWorkerRuntimeRunRef: "worker_runtime_run:external_session_hostkit",
		VisibleToolRefs:         []controlcontract.DisplaySafeRef{"tool:read"},
		ContextRefs:             []controlcontract.DisplaySafeRef{"context:external_session_hostkit"},
		TimeoutRef:              "timeout:external_session_hostkit",
		ParallelismRef:          "parallelism:external_session_hostkit",
		FailureRef:              "failure:external_session_hostkit",
		CompensationRef:         "compensation:external_session_hostkit",
	}
}
