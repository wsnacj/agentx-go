package resume

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	scheduler "github.com/wsnacj/agentx-go/runtime/scheduler"
	session "github.com/wsnacj/agentx-go/runtime/session"
)

func TestObjectiveRuntimeSchedulerResumeWorkerDispatchesReadyTick(t *testing.T) {
	suffix := "hostruntime_scheduler_resume"
	payload := BuildObjectiveRuntimeSchedulerResumeTickPayload(ObjectiveRuntimeSchedulerResumeTickPayload{
		TickRef:                      session.DisplaySafeRef("scheduler_tick:" + suffix),
		SchedulerJobRef:              session.DisplaySafeRef("scheduler_job:" + suffix),
		SchedulerRuntimeQueueRef:     session.DisplaySafeRef("runtime_queue:" + suffix),
		SchedulerWakeContinuationRef: session.DisplaySafeRef("wake_continuation:" + suffix),
		ContinuationStoreRef:         session.DisplaySafeRef("objective_runtime_wake_continuation_store:" + suffix),
		ContinuationApplyRef:         session.DisplaySafeRef("objective_runtime_wake_continuation_apply:" + suffix),
		ContinuationReadbackRef:      session.DisplaySafeRef("objective_runtime_wake_continuation_readback:" + suffix),
		WakeCursorRef:                session.DisplaySafeRef("wake_cursor:" + suffix),
		ObjectiveRunRef:              session.DisplaySafeRef("objective_run:" + suffix),
		ObjectiveGraphSnapshotRef:    session.DisplaySafeRef("objective_graph_snapshot:" + suffix),
		ObjectiveGraphRef:            session.DisplaySafeRef("objective_graph:" + suffix),
		ObjectiveGraphReadbackRef:    session.DisplaySafeRef("objective_graph_readback:" + suffix),
		DispatchRef:                  session.DisplaySafeRef("objective_runtime_wake_dispatch:" + suffix),
		RuntimeDispatchRef:           session.DisplaySafeRef("runtime_dispatch:" + suffix),
		HostRunnerRef:                session.DisplaySafeRef("runner:" + suffix),
		HostRunnerVersionRef:         session.DisplaySafeRef("runner_version:" + suffix),
		OperatorApprovalRef:          session.DisplaySafeRef("approval:" + suffix),
		Boundaries:                   []session.Boundary{"test_scheduler_resume_tick"},
	})
	payloadJSON, err := payload.JSON()
	if err != nil {
		t.Fatalf("marshal scheduler resume payload: %v", err)
	}
	readbackCalled := false
	dispatchCalled := false
	report, err := (ObjectiveRuntimeSchedulerResumeWorker{
		ContinuationReadback: func(ctx context.Context, input ObjectiveRuntimeSchedulerResumeContinuationReadbackInput) (ObjectiveRuntimeSchedulerResumeContinuationReadbackResult, error) {
			readbackCalled = true
			if input.ExpectedWakeCursorRef != "wake_cursor:"+session.DisplaySafeRef(suffix) ||
				input.ExpectedGraphSnapshotRef != "objective_graph_snapshot:"+session.DisplaySafeRef(suffix) ||
				input.ExpectedObjectiveGraphRef != "objective_graph:"+session.DisplaySafeRef(suffix) ||
				input.ExpectedRuntimeQueueRef != "runtime_queue:"+session.DisplaySafeRef(suffix) {
				t.Fatalf("unexpected readback input: %#v", input)
			}
			return ObjectiveRuntimeSchedulerResumeContinuationReadbackResult{
				ReadyForWakeContinuationResume: true,
				Status:                         "wake_continuation_readback_recorded",
				BackendKind:                    "fixture_durable_continuation_store",
				ContinuationStoreConfigured:    true,
				ContinuationStoreRef:           input.ContinuationStoreRef,
				ContinuationApplyRef:           input.ContinuationApplyRef,
				ContinuationReadbackRef:        input.ContinuationReadbackRef,
				WakeCursorRef:                  input.ExpectedWakeCursorRef,
				ObjectiveRunRef:                input.ExpectedObjectiveRunRef,
				ObjectiveGraphSnapshotRef:      input.ExpectedGraphSnapshotRef,
				ObjectiveGraphRef:              input.ExpectedObjectiveGraphRef,
				ObjectiveGraphValidationRef:    session.DisplaySafeRef("objective_graph_validation:" + suffix),
				ObjectiveGraphReadbackRef:      session.DisplaySafeRef("objective_graph_readback:" + suffix),
				ObjectiveGraphState:            "running",
				ObjectiveGraphRevision:         4,
				ReadyNodeRefs: []session.DisplaySafeRef{
					session.DisplaySafeRef("objective_node:" + suffix + "_fetch"),
					session.DisplaySafeRef("objective_node:" + suffix + "_verify"),
				},
				TaskLedgerRef:                session.DisplaySafeRef("task_ledger:" + suffix),
				HostRuntimeQueueRef:          session.DisplaySafeRef("host_runtime_queue:" + suffix),
				SchedulerRuntimeQueueRef:     input.ExpectedRuntimeQueueRef,
				SchedulerWakeContinuationRef: input.ExpectedContinuationRef,
				WakeDecision:                 "wake_llm",
				WakeContinuationEvidenceRefs: []session.DisplaySafeRef{"evidence:" + session.DisplaySafeRef(suffix)},
				WakeCursorVisible:            true,
				ManifestVisible:              true,
				DurableReadback:              true,
				CrossInstanceReadback:        true,
				Boundaries:                   []session.Boundary{"test_readback"},
			}, nil
		},
		WakeDispatch: func(ctx context.Context, input ObjectiveRuntimeSchedulerResumeWakeDispatchInput) (ObjectiveRuntimeSchedulerResumeWakeDispatchResult, error) {
			dispatchCalled = true
			if input.Continuation.WakeDecision != "wake_llm" ||
				input.DispatchRef != "objective_runtime_wake_dispatch:"+session.DisplaySafeRef(suffix) {
				t.Fatalf("unexpected dispatch input: %#v", input)
			}
			return ObjectiveRuntimeSchedulerResumeWakeDispatchResult{
				Status:                      "objective_runtime_wake_dispatch_request_ready",
				ReadyForRuntimeWakeDispatch: true,
				DispatchRequestRecorded:     true,
				HostRuntimeDispatchByHost:   true,
				DispatchRef:                 input.DispatchRef,
				RuntimeDispatchRef:          input.RuntimeDispatchRef,
				HostRunnerRef:               input.HostRunnerRef,
				HostRunnerVersionRef:        input.HostRunnerVersionRef,
				OperatorApprovalRef:         input.OperatorApprovalRef,
				ObjectiveGraphSnapshotRef:   input.Continuation.ObjectiveGraphSnapshotRef,
				ObjectiveGraphRef:           input.Continuation.ObjectiveGraphRef,
				ObjectiveGraphReadbackRef:   input.Continuation.ObjectiveGraphReadbackRef,
				ObjectiveGraphState:         input.Continuation.ObjectiveGraphState,
				ObjectiveGraphRevision:      input.Continuation.ObjectiveGraphRevision,
				ReadyNodeRefs:               input.Continuation.ReadyNodeRefs,
				EvidenceRefs:                []session.DisplaySafeRef{"evidence:" + session.DisplaySafeRef(suffix) + "_dispatch"},
				Boundaries:                  []string{"test_dispatch"},
				NextHostAction:              "host_dispatch_runtime_wake_runner",
			}, nil
		},
	}).HandleSchedulerTick(context.Background(), scheduler.Job{
		ID:      "scheduler_resume_tick_job",
		Lane:    scheduler.LaneBackground,
		JobKind: "objective_runtime_scheduler_resume_tick",
		Payload: payloadJSON,
	})
	if err != nil {
		t.Fatalf("handle scheduler resume tick: %v", err)
	}
	if !readbackCalled || !dispatchCalled {
		t.Fatalf("expected readback and dispatch callbacks, readback=%t dispatch=%t", readbackCalled, dispatchCalled)
	}
	if report.Status != "objective_runtime_scheduler_resume_dispatch_ready" ||
		report.WorkerKind != ObjectiveRuntimeSchedulerResumeWorkerKind ||
		!report.SchedulerTickObserved ||
		!report.PayloadAccepted ||
		!report.DisplaySafePayload ||
		!report.WakeContinuationReadbackReady ||
		!report.ReadyForObjectiveRunResume ||
		!report.ReadyForRuntimeWakeDispatch ||
		!report.DispatchRequestRecorded ||
		!report.HostRuntimeDispatchByHost ||
		!report.WorkerMutationByHost ||
		report.LLMWakeDispatched ||
		report.RunnerDispatched ||
		report.ToolExecuted ||
		report.WorkflowDispatched ||
		report.SchedulerApplied ||
		report.InstallerExecuted ||
		report.StoreMutationByCore ||
		report.ObjectiveGraphSnapshotRef != "objective_graph_snapshot:"+suffix ||
		report.ObjectiveGraphRef != "objective_graph:"+suffix ||
		report.ObjectiveGraphReadbackRef != "objective_graph_readback:"+suffix ||
		report.ObjectiveGraphState != "running" ||
		report.ObjectiveGraphRevision != 4 ||
		len(report.ReadyNodeRefs) != 2 ||
		report.NextHostAction != "host_dispatch_runtime_wake_runner" {
		t.Fatalf("unexpected scheduler resume worker report: %#v", report)
	}
	for _, want := range []string{
		"host_owned_objective_runtime_scheduler_resume_worker",
		"production_scheduler_tick_worker",
		"durable_objective_run_resume_worker",
		"host_owned_scheduler_tick_to_wake_dispatch",
		"scheduler_tick_resumed_objective_run",
		"scheduler_tick_bound_to_runtime_wake_dispatch",
		"no_runner_dispatch_by_core",
		"no_store_mutation_by_core",
	} {
		if !stringSliceContains(report.Boundaries, want) {
			t.Fatalf("scheduler resume worker missing boundary %q: %#v", want, report.Boundaries)
		}
	}
}

func TestObjectiveRuntimeSchedulerResumeWorkerRejectsUnsafePayload(t *testing.T) {
	rawRef := "/tmp/raw-scheduler-resume-secret"
	payload := ObjectiveRuntimeSchedulerResumeTickPayload{
		TickRef:                      "scheduler_tick:unsafe_resume",
		SchedulerRuntimeQueueRef:     "runtime_queue:unsafe_resume",
		SchedulerWakeContinuationRef: "wake_continuation:unsafe_resume",
		ContinuationStoreRef:         "objective_runtime_wake_continuation_store:unsafe_resume",
		ContinuationApplyRef:         session.DisplaySafeRef(rawRef),
		ContinuationReadbackRef:      "objective_runtime_wake_continuation_readback:unsafe_resume",
		WakeCursorRef:                "wake_cursor:unsafe_resume",
		ObjectiveRunRef:              "objective_run:unsafe_resume",
		DispatchRef:                  "objective_runtime_wake_dispatch:unsafe_resume",
		RuntimeDispatchRef:           "runtime_dispatch:unsafe_resume",
		HostRunnerRef:                "runner:unsafe_resume",
		HostRunnerVersionRef:         "runner_version:unsafe_resume",
		OperatorApprovalRef:          "approval:unsafe_resume",
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal unsafe payload: %v", err)
	}
	report, err := (ObjectiveRuntimeSchedulerResumeWorker{
		ContinuationReadback: func(context.Context, ObjectiveRuntimeSchedulerResumeContinuationReadbackInput) (ObjectiveRuntimeSchedulerResumeContinuationReadbackResult, error) {
			t.Fatal("unsafe payload must not reach readback")
			return ObjectiveRuntimeSchedulerResumeContinuationReadbackResult{}, nil
		},
		WakeDispatch: func(context.Context, ObjectiveRuntimeSchedulerResumeWakeDispatchInput) (ObjectiveRuntimeSchedulerResumeWakeDispatchResult, error) {
			t.Fatal("unsafe payload must not reach dispatch")
			return ObjectiveRuntimeSchedulerResumeWakeDispatchResult{}, nil
		},
	}).HandleSchedulerTick(context.Background(), scheduler.Job{
		ID:      "scheduler_resume_unsafe",
		JobKind: "objective_runtime_scheduler_resume_tick",
		Payload: string(payloadJSON),
	})
	if err != nil {
		t.Fatalf("handle unsafe scheduler resume tick: %v", err)
	}
	if report.Status != "blocked" ||
		report.ReadyForRuntimeWakeDispatch ||
		report.DispatchRequestRecorded ||
		!stringSliceContains(report.MissingInputs, "host:display_safe_scheduler_tick_payload") ||
		!stringSliceContains(report.BlockedReasons, "scheduler_tick_payload_unsafe") {
		t.Fatalf("expected unsafe scheduler resume payload block, got %#v", report)
	}
	blob, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("marshal unsafe scheduler resume report: %v", err)
	}
	if strings.Contains(string(blob), rawRef) {
		t.Fatalf("scheduler resume worker leaked unsafe payload %q in %s", rawRef, blob)
	}
}

func stringSliceContains(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
