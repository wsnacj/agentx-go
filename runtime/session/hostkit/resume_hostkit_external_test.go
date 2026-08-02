package hostkit_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	scheduler "github.com/wsnacj/agentx-go/runtime/scheduler"
	session "github.com/wsnacj/agentx-go/runtime/session"
	hostkit "github.com/wsnacj/agentx-go/runtime/session/hostkit"
	resume "github.com/wsnacj/agentx-go/runtime/session/resume"
)

func TestResumeRuntimeExternalLifecycle(t *testing.T) {
	queue := scheduler.NewMemoryQueue(scheduler.QueueConfig{})
	runtime, err := hostkit.NewResumeRuntime(hostkit.ResumeConfig{
		Queue:  queue,
		Worker: readyResumeWorker(),
		Lane:   scheduler.LaneBackground,
	})
	if err != nil {
		t.Fatal(err)
	}

	enqueue, err := runtime.Enqueue(context.Background(), hostkit.ResumeEnqueueRequest{
		Enabled:       true,
		Payload:       readyResumePayload("external"),
		TrustedCaller: true,
	})
	if err != nil || !enqueue.TickEnqueued {
		t.Fatalf("enqueue = %#v, %v", enqueue, err)
	}
	result, err := runtime.Run(context.Background(), hostkit.ResumeRunRequest{
		Enabled:             true,
		MaxCycles:           1,
		MaxTicksPerCycle:    1,
		ServiceRef:          "service:external",
		ConfigRef:           "config:external",
		DeploymentRef:       "deployment:external",
		OperatorApprovalRef: "approval:external",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.TicksAcked != 1 || result.WorkerCalls != 1 || !result.HostRuntimeDispatchByHost || result.LLMWakeDispatched {
		t.Fatalf("unexpected resume result: %#v", result)
	}
	if err := runtime.Shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := runtime.Enqueue(context.Background(), hostkit.ResumeEnqueueRequest{}); !errors.Is(err, hostkit.ErrResumeRuntimeClosed) {
		t.Fatalf("enqueue after shutdown = %v", err)
	}
}

func TestResumeRuntimeExternalBusyAndBoundedShutdown(t *testing.T) {
	waitStarted := make(chan struct{})
	runtime, err := hostkit.NewResumeRuntime(hostkit.ResumeConfig{
		Queue:  scheduler.NewMemoryQueue(scheduler.QueueConfig{}),
		Worker: readyResumeWorker(),
		Wait: func(ctx context.Context, _ resume.ObjectiveRuntimeSchedulerResumeDaemonServiceWaitInput) error {
			select {
			case <-waitStarted:
			default:
				close(waitStarted)
			}
			<-ctx.Done()
			return ctx.Err()
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	runDone := make(chan error, 1)
	go func() {
		_, runErr := runtime.Run(context.Background(), hostkit.ResumeRunRequest{
			Enabled:          true,
			MaxCycles:        2,
			MaxTicksPerCycle: 1,
			ContinueOnIdle:   true,
			CycleInterval:    time.Millisecond,
			ServiceRef:       "service:shutdown",
			ConfigRef:        "config:shutdown",
			DeploymentRef:    "deployment:shutdown",
		})
		runDone <- runErr
	}()
	select {
	case <-waitStarted:
	case <-time.After(time.Second):
		t.Fatal("resume runtime did not enter bounded wait")
	}
	if _, err := runtime.Run(context.Background(), hostkit.ResumeRunRequest{}); !errors.Is(err, hostkit.ErrResumeRuntimeBusy) {
		t.Fatalf("concurrent Run error = %v", err)
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := runtime.Shutdown(shutdownCtx); err != nil {
		t.Fatal(err)
	}
	if err := <-runDone; err != nil {
		t.Fatal(err)
	}
	if _, err := runtime.Run(context.Background(), hostkit.ResumeRunRequest{}); !errors.Is(err, hostkit.ErrResumeRuntimeClosed) {
		t.Fatalf("Run after shutdown error = %v", err)
	}
}

func readyResumeWorker() resume.ObjectiveRuntimeSchedulerResumeWorker {
	return resume.ObjectiveRuntimeSchedulerResumeWorker{
		ContinuationReadback: func(_ context.Context, input resume.ObjectiveRuntimeSchedulerResumeContinuationReadbackInput) (resume.ObjectiveRuntimeSchedulerResumeContinuationReadbackResult, error) {
			return resume.ObjectiveRuntimeSchedulerResumeContinuationReadbackResult{
				ReadyForWakeContinuationResume: true,
				Status:                         "wake_continuation_readback_recorded",
				BackendKind:                    "fixture_durable_store",
				ContinuationStoreConfigured:    true,
				ContinuationStoreRef:           input.ContinuationStoreRef,
				ContinuationApplyRef:           input.ContinuationApplyRef,
				ContinuationReadbackRef:        input.ContinuationReadbackRef,
				WakeCursorRef:                  input.ExpectedWakeCursorRef,
				ObjectiveRunRef:                input.ExpectedObjectiveRunRef,
				ObjectiveGraphSnapshotRef:      input.ExpectedGraphSnapshotRef,
				ObjectiveGraphRef:              input.ExpectedObjectiveGraphRef,
				ObjectiveGraphReadbackRef:      "graph_readback:external",
				ObjectiveGraphState:            "running",
				ObjectiveGraphRevision:         1,
				ReadyNodeRefs:                  []session.DisplaySafeRef{"node:external"},
				SchedulerRuntimeQueueRef:       input.ExpectedRuntimeQueueRef,
				SchedulerWakeContinuationRef:   input.ExpectedContinuationRef,
				WakeDecision:                   "wake_llm",
				WakeContinuationEvidenceRefs:   []session.DisplaySafeRef{"evidence:external"},
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

func readyResumePayload(suffix string) resume.ObjectiveRuntimeSchedulerResumeTickPayload {
	suffix = strings.TrimSpace(suffix)
	return resume.ObjectiveRuntimeSchedulerResumeTickPayload{
		TickRef:                      session.DisplaySafeRef("tick:" + suffix),
		SchedulerJobRef:              session.DisplaySafeRef("job:" + suffix),
		SchedulerRuntimeQueueRef:     session.DisplaySafeRef("queue:" + suffix),
		SchedulerWakeContinuationRef: session.DisplaySafeRef("continuation:" + suffix),
		ContinuationStoreRef:         session.DisplaySafeRef("store:" + suffix),
		ContinuationApplyRef:         session.DisplaySafeRef("apply:" + suffix),
		ContinuationReadbackRef:      session.DisplaySafeRef("readback:" + suffix),
		WakeCursorRef:                session.DisplaySafeRef("cursor:" + suffix),
		ObjectiveRunRef:              session.DisplaySafeRef("objective_run:" + suffix),
		ObjectiveGraphSnapshotRef:    session.DisplaySafeRef("snapshot:" + suffix),
		ObjectiveGraphRef:            session.DisplaySafeRef("graph:" + suffix),
		ObjectiveGraphReadbackRef:    session.DisplaySafeRef("graph_readback:" + suffix),
		DispatchRef:                  session.DisplaySafeRef("dispatch:" + suffix),
		RuntimeDispatchRef:           session.DisplaySafeRef("runtime_dispatch:" + suffix),
		HostRunnerRef:                session.DisplaySafeRef("runner:" + suffix),
		HostRunnerVersionRef:         session.DisplaySafeRef("runner_version:" + suffix),
		OperatorApprovalRef:          session.DisplaySafeRef("approval:" + suffix),
	}
}
