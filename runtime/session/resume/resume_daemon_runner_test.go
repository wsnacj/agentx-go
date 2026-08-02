package resume

import (
	"context"
	"strings"
	"testing"

	scheduler "github.com/wsnacj/agentx-go/runtime/scheduler"
	session "github.com/wsnacj/agentx-go/runtime/session"
)

func TestObjectiveRuntimeSchedulerResumeDaemonRunnerDrainsBoundedTicks(t *testing.T) {
	ctx := context.Background()
	queue := scheduler.NewMemoryQueue(scheduler.QueueConfig{})
	daemon := objectiveRuntimeSchedulerResumeRunnerReadyDaemon(queue)
	for _, suffix := range []string{"hostruntime_scheduler_daemon_runner_one", "hostruntime_scheduler_daemon_runner_two"} {
		enqueue := daemon.EnqueueSchedulerTick(ctx, ObjectiveRuntimeSchedulerResumeTickEnqueueInput{
			Enabled:       true,
			Payload:       daemonReadySchedulerResumePayload(suffix),
			TrustedCaller: true,
		})
		if !enqueue.TickEnqueued {
			t.Fatalf("enqueue %s: %#v", suffix, enqueue)
		}
	}

	report := daemon.RunSchedulerResumeDaemon(ctx, ObjectiveRuntimeSchedulerResumeDaemonRunInput{
		Enabled:  true,
		MaxTicks: 2,
	})
	if report.Status != "objective_runtime_scheduler_resume_daemon_runner_completed" ||
		!report.Enabled ||
		!report.Available ||
		!report.QueueAvailable ||
		!report.KindAwareQueue ||
		!report.WorkerAvailable ||
		report.MaxTicks != 2 ||
		report.TicksAttempted != 2 ||
		report.TicksLeased != 2 ||
		report.TicksAcked != 2 ||
		report.TicksFailed != 0 ||
		report.TicksIdle != 0 ||
		report.TicksBlocked != 0 ||
		report.WorkerCalls != 2 ||
		!report.QueueMutationByHost ||
		report.QueueMutationByCore ||
		!report.WorkerMutationByHost ||
		!report.HostRuntimeDispatchByHost ||
		report.LLMWakeDispatched ||
		report.RunnerDispatched ||
		report.RuntimeAdapterExecuted ||
		report.ToolExecuted ||
		report.WorkflowDispatched ||
		report.SchedulerApplied ||
		report.InstallerExecuted ||
		report.StoreMutationByCore ||
		report.LastObjectiveRunRef != "objective_run:hostruntime_scheduler_daemon_runner_two" ||
		report.LastGraphSnapshotRef != "objective_graph_snapshot:hostruntime_scheduler_daemon_runner_two" ||
		report.LastGraphRef != "objective_graph:hostruntime_scheduler_daemon_runner_two" ||
		report.LastGraphReadbackRef != "objective_graph_readback:hostruntime_scheduler_daemon_runner_two" ||
		report.LastGraphState != "running" ||
		report.LastGraphRevision != 5 ||
		len(report.LastReadyNodeRefs) != 1 ||
		report.LastReadyNodeRefs[0] != "objective_node:hostruntime_scheduler_daemon_runner_two_ready" ||
		len(report.ProcessReports) != 2 {
		t.Fatalf("unexpected runner report: %#v", report)
	}
	for _, want := range []string{
		"host_owned_objective_runtime_scheduler_resume_daemon_runner",
		"bounded_scheduler_resume_worker_loop",
		"scheduler_tick_job_ack_recorded",
	} {
		if !stringSliceContains(report.Boundaries, want) {
			t.Fatalf("runner report missing boundary %q: %#v", want, report.Boundaries)
		}
	}
}

func TestObjectiveRuntimeSchedulerResumeDaemonRunnerStopsOnIdle(t *testing.T) {
	ctx := context.Background()
	queue := scheduler.NewMemoryQueue(scheduler.QueueConfig{})
	daemon := objectiveRuntimeSchedulerResumeRunnerReadyDaemon(queue)

	report := daemon.RunSchedulerResumeDaemon(ctx, ObjectiveRuntimeSchedulerResumeDaemonRunInput{
		Enabled:  true,
		MaxTicks: 5,
	})
	if report.Status != "objective_runtime_scheduler_resume_daemon_runner_idle" ||
		report.TicksAttempted != 1 ||
		report.TicksIdle != 1 ||
		report.TicksAcked != 0 ||
		report.QueueMutationByHost ||
		len(report.ProcessReports) != 1 ||
		report.NextHostAction != "wait_for_scheduler_tick" {
		t.Fatalf("unexpected idle runner report: %#v", report)
	}
}

func TestObjectiveRuntimeSchedulerResumeDaemonRunnerDisabledDoesNotProcess(t *testing.T) {
	ctx := context.Background()
	queue := scheduler.NewMemoryQueue(scheduler.QueueConfig{})
	daemon := objectiveRuntimeSchedulerResumeRunnerReadyDaemon(queue)
	enqueue := daemon.EnqueueSchedulerTick(ctx, ObjectiveRuntimeSchedulerResumeTickEnqueueInput{
		Enabled:       true,
		Payload:       daemonReadySchedulerResumePayload("hostruntime_scheduler_daemon_runner_disabled"),
		TrustedCaller: true,
	})
	if !enqueue.TickEnqueued {
		t.Fatalf("enqueue disabled runner job: %#v", enqueue)
	}

	report := daemon.RunSchedulerResumeDaemon(ctx, ObjectiveRuntimeSchedulerResumeDaemonRunInput{})
	if report.Status != "blocked" ||
		report.Enabled ||
		report.TicksAttempted != 0 ||
		len(report.ProcessReports) != 0 ||
		!stringSliceContains(report.MissingInputs, "host:objective_runtime_scheduler_resume_daemon_runner_enabled") ||
		report.NextHostAction != "enable_objective_runtime_scheduler_resume_daemon_runner" {
		t.Fatalf("unexpected disabled runner report: %#v", report)
	}
	pending, err := queue.Pending(ctx, enqueue.JobID)
	if err != nil {
		t.Fatalf("pending disabled runner job: %v", err)
	}
	if !pending {
		t.Fatalf("disabled runner must leave queued tick pending")
	}
}

func TestObjectiveRuntimeSchedulerResumeDaemonRunnerFailsClosedWithoutKindAwareQueue(t *testing.T) {
	ctx := context.Background()
	queue := &schedulerResumeBasicQueueStub{
		dequeueJob: scheduler.Job{
			ID:      "scheduler_tick:daemon_runner_basic_queue",
			Lane:    scheduler.LaneBackground,
			JobKind: ObjectiveRuntimeSchedulerResumeTickJobKind,
			Payload: "{}",
		},
	}
	daemon := objectiveRuntimeSchedulerResumeRunnerReadyDaemon(queue)

	report := daemon.RunSchedulerResumeDaemon(ctx, ObjectiveRuntimeSchedulerResumeDaemonRunInput{
		Enabled: true,
	})
	if report.Status != "objective_runtime_scheduler_resume_daemon_runner_blocked" ||
		report.TicksAttempted != 1 ||
		report.TicksBlocked != 1 ||
		report.TicksLeased != 0 ||
		queue.dequeued ||
		report.KindAwareQueue ||
		!stringSliceContains(report.MissingInputs, "host:scheduler_kind_aware_queue") ||
		!stringSliceContains(report.BlockedReasons, "scheduler_kind_aware_queue_missing") ||
		report.QueueMutationByHost ||
		report.QueueMutationByCore {
		t.Fatalf("expected kind-aware fail-closed runner report, got report=%#v queue=%#v", report, queue)
	}
}

func objectiveRuntimeSchedulerResumeRunnerReadyDaemon(queue scheduler.Queue) ObjectiveRuntimeSchedulerResumeDaemon {
	return ObjectiveRuntimeSchedulerResumeDaemon{
		Queue: queue,
		Worker: ObjectiveRuntimeSchedulerResumeWorker{
			ContinuationReadback: func(ctx context.Context, input ObjectiveRuntimeSchedulerResumeContinuationReadbackInput) (ObjectiveRuntimeSchedulerResumeContinuationReadbackResult, error) {
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
					ObjectiveGraphValidationRef:    session.DisplaySafeRef("objective_graph_validation:" + input.ExpectedObjectiveRunRef),
					ObjectiveGraphReadbackRef:      session.DisplaySafeRef("objective_graph_readback:" + strings.TrimPrefix(string(input.ExpectedObjectiveRunRef), "objective_run:")),
					ObjectiveGraphState:            "running",
					ObjectiveGraphRevision:         5,
					ReadyNodeRefs: []session.DisplaySafeRef{
						session.DisplaySafeRef("objective_node:" + strings.TrimPrefix(string(input.ExpectedObjectiveRunRef), "objective_run:") + "_ready"),
					},
					TaskLedgerRef:                session.DisplaySafeRef("task_ledger:" + input.ExpectedObjectiveRunRef),
					HostRuntimeQueueRef:          session.DisplaySafeRef("host_runtime_queue:" + input.ExpectedRuntimeQueueRef),
					SchedulerRuntimeQueueRef:     input.ExpectedRuntimeQueueRef,
					SchedulerWakeContinuationRef: input.ExpectedContinuationRef,
					WakeDecision:                 "wake_llm",
					WakeContinuationEvidenceRefs: []session.DisplaySafeRef{session.DisplaySafeRef("evidence:" + input.ExpectedObjectiveRunRef)},
					WakeCursorVisible:            true,
					ManifestVisible:              true,
					DurableReadback:              true,
					CrossInstanceReadback:        true,
					Boundaries:                   []session.Boundary{"test_daemon_runner_readback"},
				}, nil
			},
			WakeDispatch: func(ctx context.Context, input ObjectiveRuntimeSchedulerResumeWakeDispatchInput) (ObjectiveRuntimeSchedulerResumeWakeDispatchResult, error) {
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
					EvidenceRefs:                []session.DisplaySafeRef{"evidence:daemon_runner_dispatch"},
					Boundaries:                  []string{"test_daemon_runner_dispatch"},
					NextHostAction:              "host_dispatch_runtime_wake_runner",
				}, nil
			},
		},
		Config: ObjectiveRuntimeSchedulerResumeDaemonConfig{
			Lane:        scheduler.LaneBackground,
			WorkerRef:   session.DisplaySafeRef("worker:daemon_runner"),
			ProducerRef: session.DisplaySafeRef("producer:daemon_runner"),
		},
	}
}
