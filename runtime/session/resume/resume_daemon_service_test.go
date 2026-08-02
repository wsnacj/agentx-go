package resume

import (
	"context"
	"errors"
	"testing"
	"time"

	scheduler "github.com/wsnacj/agentx-go/runtime/scheduler"
	session "github.com/wsnacj/agentx-go/runtime/session"
)

func TestObjectiveRuntimeSchedulerResumeDaemonServiceRunsBoundedCycles(t *testing.T) {
	ctx := context.Background()
	queue := scheduler.NewMemoryQueue(scheduler.QueueConfig{})
	daemon := objectiveRuntimeSchedulerResumeRunnerReadyDaemon(queue)
	for _, suffix := range []string{"hostruntime_scheduler_daemon_service_one", "hostruntime_scheduler_daemon_service_two"} {
		enqueue := daemon.EnqueueSchedulerTick(ctx, ObjectiveRuntimeSchedulerResumeTickEnqueueInput{
			Enabled:       true,
			Payload:       daemonReadySchedulerResumePayload(suffix),
			TrustedCaller: true,
		})
		if !enqueue.TickEnqueued {
			t.Fatalf("enqueue %s: %#v", suffix, enqueue)
		}
	}
	waits := 0
	report := (ObjectiveRuntimeSchedulerResumeDaemonService{
		Daemon: daemon,
		Wait: func(ctx context.Context, input ObjectiveRuntimeSchedulerResumeDaemonServiceWaitInput) error {
			waits++
			if input.CycleIndex != 1 ||
				input.CycleInterval != 5*time.Millisecond ||
				input.ServiceRef != "service:daemon_service" ||
				input.ConfigRef != "config:daemon_service" ||
				input.DeploymentRef != "deployment:daemon_service" {
				t.Fatalf("unexpected wait input: %#v", input)
			}
			return nil
		},
	}).Run(ctx, ObjectiveRuntimeSchedulerResumeDaemonServiceInput{
		Enabled:             true,
		MaxCycles:           2,
		MaxTicksPerCycle:    1,
		CycleInterval:       5 * time.Millisecond,
		ServiceRef:          session.DisplaySafeRef("service:daemon_service"),
		ConfigRef:           session.DisplaySafeRef("config:daemon_service"),
		DeploymentRef:       session.DisplaySafeRef("deployment:daemon_service"),
		OperatorApprovalRef: session.DisplaySafeRef("approval:daemon_service"),
	})
	if report.Status != "objective_runtime_scheduler_resume_daemon_service_completed" ||
		!report.Enabled ||
		!report.Available ||
		!report.ServiceConfigured ||
		!report.ServiceStartRequested ||
		!report.ServiceStartedByHost ||
		!report.ServiceStopRequested ||
		!report.ServiceStoppedByHost ||
		!report.ServiceLoopCompleted ||
		report.ServiceMutationByCore ||
		!report.QueueAvailable ||
		!report.KindAwareQueue ||
		!report.WorkerAvailable ||
		report.MaxCycles != 2 ||
		report.MaxTicksPerCycle != 1 ||
		report.CyclesStarted != 2 ||
		report.CyclesCompleted != 2 ||
		report.CyclesIdle != 0 ||
		report.CyclesFailed != 0 ||
		report.CyclesBlocked != 0 ||
		report.WaitsRequested != 1 ||
		report.WaitsCompleted != 1 ||
		waits != 1 ||
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
		report.ServiceRef != "service:daemon_service" ||
		report.ConfigRef != "config:daemon_service" ||
		report.DeploymentRef != "deployment:daemon_service" ||
		report.OperatorApprovalRef != "approval:daemon_service" ||
		report.LastObjectiveRunRef != "objective_run:hostruntime_scheduler_daemon_service_two" ||
		report.LastGraphSnapshotRef != "objective_graph_snapshot:hostruntime_scheduler_daemon_service_two" ||
		report.LastGraphRef != "objective_graph:hostruntime_scheduler_daemon_service_two" ||
		report.LastGraphReadbackRef != "objective_graph_readback:hostruntime_scheduler_daemon_service_two" ||
		report.LastGraphState != "running" ||
		report.LastGraphRevision != 5 ||
		len(report.LastReadyNodeRefs) != 1 ||
		report.LastReadyNodeRefs[0] != "objective_node:hostruntime_scheduler_daemon_service_two_ready" ||
		len(report.RunReports) != 2 {
		t.Fatalf("unexpected service report: %#v", report)
	}
	for _, want := range []string{
		"host_owned_objective_runtime_scheduler_resume_daemon_service",
		"bounded_service_lifecycle_by_host",
		"service_start_stop_by_host",
		"objective_runtime_scheduler_resume_daemon_service_cycle",
	} {
		if !stringSliceContains(report.Boundaries, want) {
			t.Fatalf("service report missing boundary %q: %#v", want, report.Boundaries)
		}
	}
}

func TestObjectiveRuntimeSchedulerResumeDaemonServiceDisabledDoesNotProcess(t *testing.T) {
	ctx := context.Background()
	queue := scheduler.NewMemoryQueue(scheduler.QueueConfig{})
	daemon := objectiveRuntimeSchedulerResumeRunnerReadyDaemon(queue)
	enqueue := daemon.EnqueueSchedulerTick(ctx, ObjectiveRuntimeSchedulerResumeTickEnqueueInput{
		Enabled:       true,
		Payload:       daemonReadySchedulerResumePayload("hostruntime_scheduler_daemon_service_disabled"),
		TrustedCaller: true,
	})
	if !enqueue.TickEnqueued {
		t.Fatalf("enqueue disabled service job: %#v", enqueue)
	}

	report := (ObjectiveRuntimeSchedulerResumeDaemonService{Daemon: daemon}).Run(ctx, ObjectiveRuntimeSchedulerResumeDaemonServiceInput{})
	if report.Status != "blocked" ||
		report.Enabled ||
		report.ServiceStartRequested ||
		report.ServiceStartedByHost ||
		report.CyclesStarted != 0 ||
		len(report.RunReports) != 0 ||
		!stringSliceContains(report.MissingInputs, "host:objective_runtime_scheduler_resume_daemon_service_enabled") ||
		report.NextHostAction != "enable_objective_runtime_scheduler_resume_daemon_service" {
		t.Fatalf("unexpected disabled service report: %#v", report)
	}
	pending, err := queue.Pending(ctx, enqueue.JobID)
	if err != nil {
		t.Fatalf("pending disabled service job: %v", err)
	}
	if !pending {
		t.Fatalf("disabled service must leave queued tick pending")
	}
}

func TestObjectiveRuntimeSchedulerResumeDaemonServiceStopsOnIdleByDefault(t *testing.T) {
	ctx := context.Background()
	queue := scheduler.NewMemoryQueue(scheduler.QueueConfig{})
	daemon := objectiveRuntimeSchedulerResumeRunnerReadyDaemon(queue)

	report := (ObjectiveRuntimeSchedulerResumeDaemonService{Daemon: daemon}).Run(ctx, ObjectiveRuntimeSchedulerResumeDaemonServiceInput{
		Enabled:          true,
		MaxCycles:        5,
		MaxTicksPerCycle: 1,
	})
	if report.Status != "objective_runtime_scheduler_resume_daemon_service_idle" ||
		report.CyclesStarted != 1 ||
		report.CyclesIdle != 1 ||
		report.CyclesCompleted != 0 ||
		report.WaitsRequested != 0 ||
		report.TicksIdle != 1 ||
		report.QueueMutationByHost ||
		len(report.RunReports) != 1 ||
		report.NextHostAction != "wait_for_scheduler_tick" {
		t.Fatalf("unexpected idle service report: %#v", report)
	}
}

func TestObjectiveRuntimeSchedulerResumeDaemonServiceInactivityWatchdogStopsOnRepeatedIdle(t *testing.T) {
	ctx := context.Background()
	queue := scheduler.NewMemoryQueue(scheduler.QueueConfig{})
	daemon := objectiveRuntimeSchedulerResumeRunnerReadyDaemon(queue)

	report := (ObjectiveRuntimeSchedulerResumeDaemonService{Daemon: daemon}).Run(ctx, ObjectiveRuntimeSchedulerResumeDaemonServiceInput{
		Enabled:                                true,
		MaxCycles:                              5,
		MaxTicksPerCycle:                       1,
		ContinueOnIdle:                         true,
		InactivityWatchdogEnabled:              true,
		InactivityWatchdogRef:                  "watchdog:daemon_service_idle",
		InactivityWatchdogReviewRef:            "review:daemon_service_idle",
		InactivityWatchdogHumanInterventionRef: "human_intervention:daemon_service_idle",
		MaxConsecutiveIdleCycles:               2,
	})
	if report.Status != "objective_runtime_scheduler_resume_daemon_service_watchdog_triggered" ||
		!report.InactivityWatchdogEnabled ||
		!report.InactivityWatchdogReady ||
		!report.InactivityWatchdogTriggered ||
		!report.ReadyForWatchdogStop ||
		!report.ReadyForHumanIntervention ||
		report.InactivityWatchdogReason != "consecutive_idle_cycles_exceeded" ||
		report.CyclesStarted != 2 ||
		report.CyclesIdle != 2 ||
		report.ConsecutiveIdleCycles != 2 ||
		report.ConsecutiveNoProgressCycles != 2 ||
		report.WaitsRequested != 1 ||
		report.TicksIdle != 2 ||
		report.QueueMutationByHost ||
		report.ServiceMutationByCore ||
		report.StoreMutationByCore ||
		report.NextHostAction != "review_objective_runtime_inactivity_watchdog" {
		t.Fatalf("unexpected watchdog idle service report: %#v", report)
	}
	for _, want := range []string{
		"objective_runtime_inactivity_watchdog_by_host",
		"objective_runtime_inactivity_watchdog_triggered",
		"watchdog_stop_by_host",
		"human_intervention_requested_by_host",
	} {
		if !stringSliceContains(report.Boundaries, want) {
			t.Fatalf("watchdog service report missing boundary %q: %#v", want, report.Boundaries)
		}
	}
}

func TestObjectiveRuntimeSchedulerResumeDaemonServiceWatchdogConfigBlocksBeforeProcessing(t *testing.T) {
	ctx := context.Background()
	queue := scheduler.NewMemoryQueue(scheduler.QueueConfig{})
	daemon := objectiveRuntimeSchedulerResumeRunnerReadyDaemon(queue)
	enqueue := daemon.EnqueueSchedulerTick(ctx, ObjectiveRuntimeSchedulerResumeTickEnqueueInput{
		Enabled:       true,
		Payload:       daemonReadySchedulerResumePayload("hostruntime_scheduler_daemon_service_watchdog_blocked"),
		TrustedCaller: true,
	})
	if !enqueue.TickEnqueued {
		t.Fatalf("enqueue watchdog blocked service job: %#v", enqueue)
	}

	report := (ObjectiveRuntimeSchedulerResumeDaemonService{Daemon: daemon}).Run(ctx, ObjectiveRuntimeSchedulerResumeDaemonServiceInput{
		Enabled:                   true,
		InactivityWatchdogEnabled: true,
	})
	if report.Status != "blocked" ||
		!report.InactivityWatchdogEnabled ||
		report.InactivityWatchdogReady ||
		report.ServiceConfigured ||
		report.ServiceStartRequested ||
		report.CyclesStarted != 0 ||
		!stringSliceContains(report.MissingInputs, "host:objective_runtime_inactivity_watchdog_ref") ||
		!stringSliceContains(report.MissingInputs, "host:objective_runtime_inactivity_watchdog_review_ref") ||
		!stringSliceContains(report.MissingInputs, "host:objective_runtime_inactivity_watchdog_human_intervention_ref") ||
		!stringSliceContains(report.MissingInputs, "host:objective_runtime_inactivity_watchdog_threshold") ||
		report.NextHostAction != "provide_objective_runtime_inactivity_watchdog_config" {
		t.Fatalf("unexpected watchdog blocked service report: %#v", report)
	}
	pending, err := queue.Pending(ctx, enqueue.JobID)
	if err != nil {
		t.Fatalf("pending watchdog blocked service job: %v", err)
	}
	if !pending {
		t.Fatalf("watchdog config block must leave queued tick pending")
	}
}

func TestObjectiveRuntimeSchedulerResumeDaemonServicePropagatesKindAwareBlock(t *testing.T) {
	ctx := context.Background()
	queue := &schedulerResumeBasicQueueStub{
		dequeueJob: scheduler.Job{
			ID:      "scheduler_tick:daemon_service_basic_queue",
			Lane:    scheduler.LaneBackground,
			JobKind: ObjectiveRuntimeSchedulerResumeTickJobKind,
			Payload: "{}",
		},
	}
	daemon := objectiveRuntimeSchedulerResumeRunnerReadyDaemon(queue)

	report := (ObjectiveRuntimeSchedulerResumeDaemonService{Daemon: daemon}).Run(ctx, ObjectiveRuntimeSchedulerResumeDaemonServiceInput{
		Enabled: true,
	})
	if report.Status != "objective_runtime_scheduler_resume_daemon_service_blocked" ||
		report.CyclesStarted != 1 ||
		report.CyclesBlocked != 1 ||
		report.TicksBlocked != 1 ||
		queue.dequeued ||
		report.KindAwareQueue ||
		!stringSliceContains(report.MissingInputs, "host:scheduler_kind_aware_queue") ||
		!stringSliceContains(report.BlockedReasons, "scheduler_kind_aware_queue_missing") ||
		report.QueueMutationByHost ||
		report.ServiceMutationByCore {
		t.Fatalf("expected service kind-aware fail-closed, got report=%#v queue=%#v", report, queue)
	}
}

func TestObjectiveRuntimeSchedulerResumeDaemonServiceStopsOnWaitFailure(t *testing.T) {
	ctx := context.Background()
	queue := scheduler.NewMemoryQueue(scheduler.QueueConfig{})
	daemon := objectiveRuntimeSchedulerResumeRunnerReadyDaemon(queue)
	for _, suffix := range []string{"hostruntime_scheduler_daemon_service_wait_one", "hostruntime_scheduler_daemon_service_wait_two"} {
		enqueue := daemon.EnqueueSchedulerTick(ctx, ObjectiveRuntimeSchedulerResumeTickEnqueueInput{
			Enabled:       true,
			Payload:       daemonReadySchedulerResumePayload(suffix),
			TrustedCaller: true,
		})
		if !enqueue.TickEnqueued {
			t.Fatalf("enqueue wait failure job: %#v", enqueue)
		}
	}
	report := (ObjectiveRuntimeSchedulerResumeDaemonService{
		Daemon: daemon,
		Wait: func(context.Context, ObjectiveRuntimeSchedulerResumeDaemonServiceWaitInput) error {
			return errors.New("test wait failure")
		},
	}).Run(ctx, ObjectiveRuntimeSchedulerResumeDaemonServiceInput{
		Enabled:          true,
		MaxCycles:        2,
		MaxTicksPerCycle: 1,
		ContinueOnIdle:   true,
	})
	if report.Status != "objective_runtime_scheduler_resume_daemon_service_wait_failed" ||
		report.CyclesStarted != 1 ||
		report.CyclesCompleted != 1 ||
		report.WaitsRequested != 1 ||
		report.WaitsCompleted != 0 ||
		report.TicksAcked != 1 ||
		!stringSliceContains(report.BlockedReasons, "objective_runtime_scheduler_resume_daemon_service_wait_failed") {
		t.Fatalf("unexpected wait failure service report: %#v", report)
	}
}
