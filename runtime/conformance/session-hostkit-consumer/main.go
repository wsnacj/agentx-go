package main

import (
	"context"
	"errors"
	"fmt"

	scheduler "github.com/wsnacj/agentx-go/runtime/scheduler"
	session "github.com/wsnacj/agentx-go/runtime/session"
	sessionhostkit "github.com/wsnacj/agentx-go/runtime/session/hostkit"
	resume "github.com/wsnacj/agentx-go/runtime/session/resume"
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
	resumeOutput, err := runResume(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Println(resumeOutput)
}

func runResume(ctx context.Context) (string, error) {
	queue := scheduler.NewMemoryQueue(scheduler.QueueConfig{})
	runtime, err := sessionhostkit.NewResumeRuntime(sessionhostkit.ResumeConfig{
		Queue:  queue,
		Worker: readyResumeWorker(),
		Lane:   scheduler.LaneBackground,
	})
	if err != nil {
		return "", err
	}
	request := sessionhostkit.ResumeEnqueueRequest{
		Enabled:        true,
		Payload:        readyResumePayload(),
		TrustedCaller:  true,
		IdempotencyKey: "resume-conformance",
	}
	enqueue, err := runtime.Enqueue(ctx, request)
	if err != nil || !enqueue.TickEnqueued {
		return "", fmt.Errorf("resume enqueue blocked: report=%#v err=%v", enqueue, err)
	}
	report, err := runtime.Run(ctx, sessionhostkit.ResumeRunRequest{
		Enabled:          true,
		MaxCycles:        1,
		MaxTicksPerCycle: 1,
	})
	if err != nil {
		return "", err
	}
	if report.TicksAcked != 1 || !report.HostRuntimeDispatchByHost {
		return "", fmt.Errorf("resume run blocked: status=%s acked=%d dispatch=%t", report.Status, report.TicksAcked, report.HostRuntimeDispatchByHost)
	}
	if err := runtime.Shutdown(context.Background()); err != nil {
		return "", err
	}
	if err := runtime.Shutdown(context.Background()); err != nil {
		return "", err
	}
	if _, err := runtime.Enqueue(context.Background(), request); !errors.Is(err, sessionhostkit.ErrResumeRuntimeClosed) {
		return "", fmt.Errorf("resume closed call error = %v", err)
	}
	return fmt.Sprintf("agentx-resume-hostkit-ok:%s:%d:%t", report.Status, report.TicksAcked, report.HostRuntimeDispatchByHost), nil
}

func readyResumeWorker() resume.Worker {
	return resume.Worker{
		ContinuationReadback: func(_ context.Context, input resume.ObjectiveRuntimeSchedulerResumeContinuationReadbackInput) (resume.ObjectiveRuntimeSchedulerResumeContinuationReadbackResult, error) {
			return resume.ObjectiveRuntimeSchedulerResumeContinuationReadbackResult{
				ReadyForWakeContinuationResume: true,
				Status:                         "wake_continuation_readback_recorded",
				BackendKind:                    "conformance_durable_store",
				ContinuationStoreConfigured:    true,
				ContinuationStoreRef:           input.ContinuationStoreRef,
				ContinuationApplyRef:           input.ContinuationApplyRef,
				ContinuationReadbackRef:        input.ContinuationReadbackRef,
				WakeCursorRef:                  input.ExpectedWakeCursorRef,
				ObjectiveRunRef:                input.ExpectedObjectiveRunRef,
				ObjectiveGraphSnapshotRef:      input.ExpectedGraphSnapshotRef,
				ObjectiveGraphRef:              input.ExpectedObjectiveGraphRef,
				ObjectiveGraphReadbackRef:      "graph_readback:conformance_resume",
				ObjectiveGraphState:            "running",
				ObjectiveGraphRevision:         1,
				ReadyNodeRefs:                  []session.DisplaySafeRef{"node:conformance_resume"},
				SchedulerRuntimeQueueRef:       input.ExpectedRuntimeQueueRef,
				SchedulerWakeContinuationRef:   input.ExpectedContinuationRef,
				WakeDecision:                   "wake_llm",
				WakeContinuationEvidenceRefs:   []session.DisplaySafeRef{"evidence:conformance_resume"},
				WakeCursorVisible:              true,
				ManifestVisible:                true,
				DurableReadback:                true,
				CrossInstanceReadback:          true,
			}, nil
		},
		WakeDispatch: func(_ context.Context, input resume.ObjectiveRuntimeSchedulerResumeWakeDispatchInput) (resume.ObjectiveRuntimeSchedulerResumeWakeDispatchResult, error) {
			return resume.ObjectiveRuntimeSchedulerResumeWakeDispatchResult{
				Status:                      "objective_runtime_wake_dispatch_request_ready",
				ReadyForRuntimeWakeDispatch: true,
				DispatchRequestRecorded:     true,
				HostRuntimeDispatchByHost:   true,
				DispatchRef:                 input.DispatchRef,
				RuntimeDispatchRef:          input.RuntimeDispatchRef,
				HostRunnerRef:               input.HostRunnerRef,
				HostRunnerVersionRef:        input.HostRunnerVersionRef,
				OperatorApprovalRef:         input.OperatorApprovalRef,
				NextHostAction:              "host_dispatch_runtime_wake_runner",
			}, nil
		},
	}
}

func readyResumePayload() resume.ObjectiveRuntimeSchedulerResumeTickPayload {
	return resume.ObjectiveRuntimeSchedulerResumeTickPayload{
		TickRef:                      "tick:conformance_resume",
		SchedulerJobRef:              "job:conformance_resume",
		SchedulerRuntimeQueueRef:     "queue:conformance_resume",
		SchedulerWakeContinuationRef: "continuation:conformance_resume",
		ContinuationStoreRef:         "store:conformance_resume",
		ContinuationApplyRef:         "apply:conformance_resume",
		ContinuationReadbackRef:      "readback:conformance_resume",
		WakeCursorRef:                "cursor:conformance_resume",
		ObjectiveRunRef:              "objective_run:conformance_resume",
		ObjectiveGraphSnapshotRef:    "snapshot:conformance_resume",
		ObjectiveGraphRef:            "graph:conformance_resume",
		ObjectiveGraphReadbackRef:    "graph_readback:conformance_resume",
		DispatchRef:                  "dispatch:conformance_resume",
		RuntimeDispatchRef:           "runtime_dispatch:conformance_resume",
		HostRunnerRef:                "runner:conformance_resume",
		HostRunnerVersionRef:         "runner_version:conformance_resume",
		OperatorApprovalRef:          "approval:conformance_resume",
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
