package resume

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	scheduler "github.com/wsnacj/agentx-go/runtime/scheduler"
	session "github.com/wsnacj/agentx-go/runtime/session"
)

func TestObjectiveRuntimeSchedulerResumeDaemonEnqueuesAndProcessesTick(t *testing.T) {
	ctx := context.Background()
	suffix := "hostruntime_scheduler_daemon"
	queue := scheduler.NewMemoryQueue(scheduler.QueueConfig{})
	daemon := ObjectiveRuntimeSchedulerResumeDaemon{
		Queue: queue,
		Worker: ObjectiveRuntimeSchedulerResumeWorker{
			ContinuationReadback: func(ctx context.Context, input ObjectiveRuntimeSchedulerResumeContinuationReadbackInput) (ObjectiveRuntimeSchedulerResumeContinuationReadbackResult, error) {
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
					ObjectiveGraphRevision:         3,
					ReadyNodeRefs: []session.DisplaySafeRef{
						session.DisplaySafeRef("objective_node:" + suffix + "_collect"),
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
					Boundaries:                   []session.Boundary{"test_daemon_readback"},
				}, nil
			},
			WakeDispatch: func(ctx context.Context, input ObjectiveRuntimeSchedulerResumeWakeDispatchInput) (ObjectiveRuntimeSchedulerResumeWakeDispatchResult, error) {
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
					EvidenceRefs:                []session.DisplaySafeRef{"evidence:" + session.DisplaySafeRef(suffix) + "_dispatch"},
					Boundaries:                  []string{"test_daemon_dispatch"},
					NextHostAction:              "host_dispatch_runtime_wake_runner",
				}, nil
			},
		},
		Config: ObjectiveRuntimeSchedulerResumeDaemonConfig{
			Lane:        scheduler.LaneBackground,
			WorkerRef:   session.DisplaySafeRef("worker:" + suffix),
			ProducerRef: session.DisplaySafeRef("producer:" + suffix),
		},
	}
	payload := daemonReadySchedulerResumePayload(suffix)
	enqueue := daemon.EnqueueSchedulerTick(ctx, ObjectiveRuntimeSchedulerResumeTickEnqueueInput{
		Enabled:       true,
		Payload:       payload,
		TrustedCaller: true,
	})
	if enqueue.Status != "objective_runtime_scheduler_tick_enqueued" ||
		!enqueue.TickEnqueued ||
		!enqueue.QueueMutationByHost ||
		enqueue.QueueMutationByCore ||
		!enqueue.QueuePendingReadback ||
		!enqueue.QueuePending ||
		enqueue.JobID != "scheduler_tick:"+suffix ||
		enqueue.ObjectiveRunRef != "objective_run:"+suffix ||
		enqueue.ObjectiveGraphSnapshotRef != "objective_graph_snapshot:"+suffix ||
		enqueue.ObjectiveGraphRef != "objective_graph:"+suffix ||
		enqueue.ObjectiveGraphReadbackRef != "objective_graph_readback:"+suffix ||
		enqueue.NextHostAction != "run_objective_runtime_scheduler_resume_worker" {
		t.Fatalf("unexpected enqueue report: %#v", enqueue)
	}
	process := daemon.ProcessNextSchedulerTick(ctx, ObjectiveRuntimeSchedulerResumeProcessInput{
		Enabled: true,
	})
	if process.Status != "objective_runtime_scheduler_resume_worker_ack_recorded" ||
		!process.LeaseRequested ||
		!process.LeaseAcquired ||
		!process.WorkerCalled ||
		!process.WorkerMutationByHost ||
		!process.QueueAcked ||
		process.QueueFailed ||
		!process.QueueResultReadbackReady ||
		process.QueueResultStatus != "completed" ||
		!process.QueueResultSucceeded ||
		process.ObjectiveRunRef != "objective_run:"+suffix ||
		process.ObjectiveGraphSnapshotRef != "objective_graph_snapshot:"+suffix ||
		process.ObjectiveGraphRef != "objective_graph:"+suffix ||
		process.ObjectiveGraphReadbackRef != "objective_graph_readback:"+suffix ||
		process.ObjectiveGraphState != "running" ||
		process.ObjectiveGraphRevision != 3 ||
		len(process.ReadyNodeRefs) != 2 ||
		process.ReadyNodeRefs[0] != "objective_node:"+suffix+"_collect" ||
		process.ReadyNodeRefs[1] != "objective_node:"+suffix+"_verify" ||
		!process.ReadyForRuntimeWakeDispatch ||
		!process.DispatchRequestRecorded ||
		!process.HostRuntimeDispatchByHost ||
		process.LLMWakeDispatched ||
		process.RunnerDispatched ||
		process.RuntimeAdapterExecuted ||
		process.ToolExecuted ||
		process.WorkflowDispatched ||
		process.SchedulerApplied ||
		process.InstallerExecuted ||
		process.StoreMutationByCore ||
		process.NextHostAction != "host_dispatch_runtime_wake_runner" {
		t.Fatalf("unexpected process report: %#v", process)
	}
	for _, want := range []string{
		"production_scheduler_tick_enqueue",
		"scheduler_tick_queued_for_resume_worker",
	} {
		if !stringSliceContains(enqueue.Boundaries, want) {
			t.Fatalf("enqueue report missing boundary %q: %#v", want, enqueue.Boundaries)
		}
	}
	for _, want := range []string{
		"cross_process_scheduler_resume_worker",
		"scheduler_queue_lease_ack_readback",
		"scheduler_tick_job_leased_by_host",
		"scheduler_tick_job_ack_recorded",
		"scheduler_tick_bound_to_runtime_wake_dispatch",
	} {
		if !stringSliceContains(process.Boundaries, want) {
			t.Fatalf("process report missing boundary %q: %#v", want, process.Boundaries)
		}
	}
}

func TestObjectiveRuntimeSchedulerResumeDaemonRejectsUnsafePayloadWithoutQueueMutation(t *testing.T) {
	ctx := context.Background()
	queue := scheduler.NewMemoryQueue(scheduler.QueueConfig{})
	rawRef := "/tmp/raw-scheduler-daemon-secret"
	daemon := ObjectiveRuntimeSchedulerResumeDaemon{Queue: queue}
	report := daemon.EnqueueSchedulerTick(ctx, ObjectiveRuntimeSchedulerResumeTickEnqueueInput{
		Enabled: true,
		Payload: ObjectiveRuntimeSchedulerResumeTickPayload{
			TickRef:                      "scheduler_tick:unsafe_daemon",
			SchedulerRuntimeQueueRef:     "runtime_queue:unsafe_daemon",
			SchedulerWakeContinuationRef: "wake_continuation:unsafe_daemon",
			ContinuationStoreRef:         "objective_runtime_wake_continuation_store:unsafe_daemon",
			ContinuationApplyRef:         session.DisplaySafeRef(rawRef),
		},
	})
	if report.Status != "blocked" ||
		report.TickEnqueued ||
		report.QueueMutationByHost ||
		!stringSliceContains(report.MissingInputs, "host:display_safe_scheduler_tick_payload") ||
		!stringSliceContains(report.BlockedReasons, "scheduler_tick_payload_unsafe") {
		t.Fatalf("expected unsafe enqueue block, got %#v", report)
	}
	if _, err := queue.Dequeue(ctx, scheduler.LaneBackground); err != scheduler.ErrQueueEmpty {
		t.Fatalf("expected no queued unsafe tick, got %v", err)
	}
	blob, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("marshal unsafe enqueue report: %v", err)
	}
	if strings.Contains(string(blob), rawRef) {
		t.Fatalf("unsafe enqueue report leaked raw ref %q in %s", rawRef, blob)
	}
}

func TestObjectiveRuntimeSchedulerResumeDaemonFailsBlockedWorkerWithReadback(t *testing.T) {
	ctx := context.Background()
	suffix := "hostruntime_scheduler_daemon_blocked"
	queue := scheduler.NewMemoryQueue(scheduler.QueueConfig{})
	daemon := ObjectiveRuntimeSchedulerResumeDaemon{
		Queue: queue,
		Worker: ObjectiveRuntimeSchedulerResumeWorker{
			ContinuationReadback: func(ctx context.Context, input ObjectiveRuntimeSchedulerResumeContinuationReadbackInput) (ObjectiveRuntimeSchedulerResumeContinuationReadbackResult, error) {
				return ObjectiveRuntimeSchedulerResumeContinuationReadbackResult{
					Status:                  "blocked",
					ContinuationStoreRef:    input.ContinuationStoreRef,
					ContinuationApplyRef:    input.ContinuationApplyRef,
					ContinuationReadbackRef: input.ContinuationReadbackRef,
					WakeCursorRef:           input.ExpectedWakeCursorRef,
					ObjectiveRunRef:         input.ExpectedObjectiveRunRef,
					MissingInputs:           []string{"host:objective_runtime_wake_continuation_readback"},
					BlockedReasons:          []string{"objective_runtime_wake_continuation_readback_not_ready"},
				}, nil
			},
			WakeDispatch: func(context.Context, ObjectiveRuntimeSchedulerResumeWakeDispatchInput) (ObjectiveRuntimeSchedulerResumeWakeDispatchResult, error) {
				t.Fatal("wake dispatch must not run when readback is blocked")
				return ObjectiveRuntimeSchedulerResumeWakeDispatchResult{}, nil
			},
		},
		Config: ObjectiveRuntimeSchedulerResumeDaemonConfig{Lane: scheduler.LaneBackground},
	}
	enqueue := daemon.EnqueueSchedulerTick(ctx, ObjectiveRuntimeSchedulerResumeTickEnqueueInput{
		Enabled: true,
		Payload: daemonReadySchedulerResumePayload(suffix),
	})
	if !enqueue.TickEnqueued {
		t.Fatalf("expected enqueue before blocked processing, got %#v", enqueue)
	}
	process := daemon.ProcessNextSchedulerTick(ctx, ObjectiveRuntimeSchedulerResumeProcessInput{Enabled: true})
	if process.Status != "objective_runtime_scheduler_resume_worker_failure_recorded" ||
		!process.LeaseAcquired ||
		!process.WorkerCalled ||
		process.QueueAcked ||
		!process.QueueFailed ||
		!process.QueueResultReadbackReady ||
		process.QueueResultStatus != "failed" ||
		process.QueueResultSucceeded ||
		process.ReadyForRuntimeWakeDispatch ||
		!stringSliceContains(process.BlockedReasons, "objective_runtime_scheduler_resume_worker_not_ready") {
		t.Fatalf("expected failed blocked worker readback, got %#v", process)
	}
}

func TestObjectiveRuntimeSchedulerResumeDaemonSkipsUnexpectedJobKind(t *testing.T) {
	ctx := context.Background()
	queue := scheduler.NewMemoryQueue(scheduler.QueueConfig{})
	if err := queue.Enqueue(ctx, scheduler.Job{
		ID:      "scheduler_tick:wrong_kind",
		Lane:    scheduler.LaneBackground,
		JobKind: "unrelated_job",
		Payload: "{}",
	}); err != nil {
		t.Fatalf("enqueue wrong-kind job: %v", err)
	}
	daemon := ObjectiveRuntimeSchedulerResumeDaemon{
		Queue: queue,
		Worker: ObjectiveRuntimeSchedulerResumeWorker{
			ContinuationReadback: func(context.Context, ObjectiveRuntimeSchedulerResumeContinuationReadbackInput) (ObjectiveRuntimeSchedulerResumeContinuationReadbackResult, error) {
				t.Fatal("readback must not run for unexpected job kind")
				return ObjectiveRuntimeSchedulerResumeContinuationReadbackResult{}, nil
			},
			WakeDispatch: func(context.Context, ObjectiveRuntimeSchedulerResumeWakeDispatchInput) (ObjectiveRuntimeSchedulerResumeWakeDispatchResult, error) {
				t.Fatal("dispatch must not run for unexpected job kind")
				return ObjectiveRuntimeSchedulerResumeWakeDispatchResult{}, nil
			},
		},
		Config: ObjectiveRuntimeSchedulerResumeDaemonConfig{Lane: scheduler.LaneBackground},
	}
	report := daemon.ProcessNextSchedulerTick(ctx, ObjectiveRuntimeSchedulerResumeProcessInput{Enabled: true})
	if report.Status != "objective_runtime_scheduler_resume_worker_idle" ||
		report.LeaseAcquired ||
		report.WorkerCalled ||
		report.QueueAcked ||
		report.QueueFailed {
		t.Fatalf("expected unexpected job kind to remain unclaimed, got %#v", report)
	}
	pending, err := queue.Pending(ctx, "scheduler_tick:wrong_kind")
	if err != nil {
		t.Fatalf("pending wrong-kind job: %v", err)
	}
	if !pending {
		t.Fatalf("expected wrong-kind job to remain pending")
	}
}

func TestObjectiveRuntimeSchedulerResumeDaemonDoesNotLeaseWithoutWorker(t *testing.T) {
	ctx := context.Background()
	queue := scheduler.NewMemoryQueue(scheduler.QueueConfig{})
	if err := queue.Enqueue(ctx, scheduler.Job{
		ID:      "scheduler_tick:no_worker",
		Lane:    scheduler.LaneBackground,
		JobKind: ObjectiveRuntimeSchedulerResumeTickJobKind,
		Payload: "{}",
	}); err != nil {
		t.Fatalf("enqueue no-worker job: %v", err)
	}
	daemon := ObjectiveRuntimeSchedulerResumeDaemon{Queue: queue}
	report := daemon.ProcessNextSchedulerTick(ctx, ObjectiveRuntimeSchedulerResumeProcessInput{Enabled: true})
	if report.LeaseAcquired ||
		report.WorkerCalled ||
		!stringSliceContains(report.MissingInputs, "host:objective_runtime_scheduler_resume_worker") ||
		!stringSliceContains(report.BlockedReasons, "objective_runtime_scheduler_resume_worker_missing") {
		t.Fatalf("expected no lease without worker, got %#v", report)
	}
	pending, err := queue.Pending(ctx, "scheduler_tick:no_worker")
	if err != nil {
		t.Fatalf("pending no-worker job: %v", err)
	}
	if !pending {
		t.Fatalf("expected no-worker job to remain pending")
	}
}

func TestObjectiveRuntimeSchedulerResumeDaemonDoesNotLeaseWithoutKindAwareQueue(t *testing.T) {
	ctx := context.Background()
	queue := &schedulerResumeBasicQueueStub{
		dequeueJob: scheduler.Job{
			ID:      "scheduler_tick:basic_queue",
			Lane:    scheduler.LaneBackground,
			JobKind: ObjectiveRuntimeSchedulerResumeTickJobKind,
			Payload: "{}",
		},
	}
	daemon := ObjectiveRuntimeSchedulerResumeDaemon{
		Queue: queue,
		Worker: ObjectiveRuntimeSchedulerResumeWorker{
			ContinuationReadback: func(context.Context, ObjectiveRuntimeSchedulerResumeContinuationReadbackInput) (ObjectiveRuntimeSchedulerResumeContinuationReadbackResult, error) {
				t.Fatal("readback must not run without kind-aware queue")
				return ObjectiveRuntimeSchedulerResumeContinuationReadbackResult{}, nil
			},
			WakeDispatch: func(context.Context, ObjectiveRuntimeSchedulerResumeWakeDispatchInput) (ObjectiveRuntimeSchedulerResumeWakeDispatchResult, error) {
				t.Fatal("dispatch must not run without kind-aware queue")
				return ObjectiveRuntimeSchedulerResumeWakeDispatchResult{}, nil
			},
		},
	}
	report := daemon.ProcessNextSchedulerTick(ctx, ObjectiveRuntimeSchedulerResumeProcessInput{Enabled: true})
	if report.LeaseRequested ||
		report.LeaseAcquired ||
		queue.dequeued ||
		!stringSliceContains(report.MissingInputs, "host:scheduler_kind_aware_queue") ||
		!stringSliceContains(report.BlockedReasons, "scheduler_kind_aware_queue_missing") {
		t.Fatalf("expected fail-closed without kind-aware queue, got report=%#v queue=%#v", report, queue)
	}
}

func daemonReadySchedulerResumePayload(suffix string) ObjectiveRuntimeSchedulerResumeTickPayload {
	return ObjectiveRuntimeSchedulerResumeTickPayload{
		TickRef:                      session.DisplaySafeRef("scheduler_tick:" + suffix),
		SchedulerJobRef:              session.DisplaySafeRef("scheduler_tick:" + suffix),
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
		Boundaries:                   []session.Boundary{"test_daemon_tick"},
	}
}

type schedulerResumeBasicQueueStub struct {
	dequeueJob scheduler.Job
	dequeued   bool
}

func (q *schedulerResumeBasicQueueStub) Enqueue(context.Context, scheduler.Job) error {
	return nil
}

func (q *schedulerResumeBasicQueueStub) Dequeue(context.Context, scheduler.Lane) (scheduler.Job, error) {
	q.dequeued = true
	return q.dequeueJob, nil
}

func (q *schedulerResumeBasicQueueStub) Ack(context.Context, scheduler.Result) error {
	return nil
}

func (q *schedulerResumeBasicQueueStub) Fail(context.Context, scheduler.Result) error {
	return nil
}

func (q *schedulerResumeBasicQueueStub) Result(context.Context, string) (scheduler.Result, bool, error) {
	return scheduler.Result{}, false, nil
}

func (q *schedulerResumeBasicQueueStub) Pending(context.Context, string) (bool, error) {
	return true, nil
}
