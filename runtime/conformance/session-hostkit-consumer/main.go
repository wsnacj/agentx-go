package main

import (
	"context"
	"errors"
	"fmt"

	session "github.com/wsnacj/agentx-go/runtime/session"
	sessionhostkit "github.com/wsnacj/agentx-go/runtime/session/hostkit"
)

func run(ctx context.Context) (string, error) {
	worker := &stubWorker{}
	kit, err := sessionhostkit.New(sessionhostkit.Config{
		Worker:     worker,
		Store:      sessionhostkit.NewInMemoryStateStore(),
		BackendRef: "backend:conformance_session_hostkit",
		Durable:    true,
	})
	if err != nil {
		return "", err
	}
	result, err := kit.Run(ctx, sessionhostkit.RunRequest{BackendInput: readyInput()})
	if err != nil {
		return "", err
	}
	if !result.Completed || !result.Backend.WorkerResultReadbackReady {
		return "", fmt.Errorf("child lifecycle blocked: status=%s failure=%s next=%s", result.Status, result.Backend.FailureClass, result.Backend.NextHostAction)
	}
	if err := kit.Shutdown(context.Background()); err != nil {
		return "", err
	}
	if err := kit.Shutdown(context.Background()); err != nil {
		return "", err
	}
	if _, err := kit.Run(context.Background(), sessionhostkit.RunRequest{BackendInput: readyInput()}); !errors.Is(err, sessionhostkit.ErrClosed) {
		return "", fmt.Errorf("closed call error = %v", err)
	}
	return fmt.Sprintf("agentx-session-hostkit-ok:%s:%t:%d:%d", result.Status, result.Backend.WorkerResultRequiresVerification, worker.invokeCalls, worker.readbackCalls), nil
}

func main() {
	output, err := run(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Println(output)
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
		EvidenceRefs: []session.EvidenceRef{{
			Ref:      "evidence:external_session_hostkit",
			Kind:     "delegation_worker_result",
			Strength: session.EvidenceAdequate,
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
		EvidenceRefs: []session.EvidenceRef{{
			Ref:      "evidence:external_session_hostkit_readback",
			Kind:     "delegation_worker_readback",
			Strength: session.EvidenceAdequate,
			Source:   "readback:external_session_hostkit",
		}},
	}, nil
}

func readyInput() sessionhostkit.BackendInput {
	readiness := session.BuildHostOwnedDelegationWorkerRuntimeReadiness(session.HostOwnedDelegationWorkerRuntimeReadinessInput{
		Request: session.BuildDelegationRequestProjection(session.DelegationRequestInput{
			Activation:         session.ActivationManaged,
			RequestedIntensity: session.IntensityL4DurableLongRun,
			Frame: session.ObjectiveFrame{
				ID:              "objective:external_session_hostkit",
				UserGoalDigest:  "sha256:external_session_hostkit",
				ControlMode:     session.ControlModeDelegated,
				Intensity:       session.IntensityL4DurableLongRun,
				SuccessCriteria: []string{"child result is verified before parent merge"},
				RequiredEvidence: []session.EvidenceRef{{
					Ref:      "evidence:external_session_hostkit",
					Kind:     "delegation_worker_result",
					Strength: session.EvidenceAdequate,
					Source:   "worker:external_session_hostkit",
				}},
			},
			SubgoalRef:                        "subgoal:external_session_hostkit",
			WorkerRef:                         "worker:external_session_hostkit",
			AllowedToolRefs:                   []session.DisplaySafeRef{"tool:read"},
			Budget:                            session.ObjectiveBudgetSnapshot{BudgetRef: "budget:external_session_hostkit", Limit: 1, Remaining: 1},
			EvidenceRequirements:              []session.EvidenceRef{{Ref: "evidence:external_session_hostkit", Kind: "delegation_worker_result", Strength: session.EvidenceAdequate, Source: "worker:external_session_hostkit"}},
			StopConditionRefs:                 []session.DisplaySafeRef{"stop:external_session_hostkit_verified"},
			RedactionPolicyRef:                "redaction:external_session_hostkit",
			MergePolicyRef:                    "merge:external_session_hostkit",
			ExecutionContractAllowsDelegation: true,
			HostAllowsL4Delegation:            true,
			UserConfirmed:                     true,
			HostApproved:                      true,
			ApprovalRefs:                      []session.DisplaySafeRef{"approval:external_session_hostkit"},
			PolicyRefs:                        []session.DisplaySafeRef{"policy:external_session_hostkit"},
		}),
		WorkerRuntimeGate: session.BuildProductionAdapterIndependentEffectGate(session.ProductionAdapterIndependentEffectGateSpec{
			Kind:                  session.ProductionAdapterEffectGateDelegationWorker,
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
			EvidenceRefs:          []session.DisplaySafeRef{"evidence:external_session_hostkit"},
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
		VisibleToolRefs:         []session.DisplaySafeRef{"tool:read"},
		ContextRefs:             []session.DisplaySafeRef{"context:external_session_hostkit"},
		TimeoutRef:              "timeout:external_session_hostkit",
		ParallelismRef:          "parallelism:external_session_hostkit",
		FailureRef:              "failure:external_session_hostkit",
		CompensationRef:         "compensation:external_session_hostkit",
	}
}
