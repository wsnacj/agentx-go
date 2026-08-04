package main

import (
	"context"
	"errors"
	"fmt"

	scheduler "github.com/wsnacj/agentx-go/runtime/scheduler"
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
	resumeOutput, err := runResume(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Println(resumeOutput)
}

func runResume(ctx context.Context) (string, error) {
	queue := scheduler.NewMemoryQueue(scheduler.QueueConfig{})
	enqueueRuntime, err := sessionhostkit.NewResumeRuntime(sessionhostkit.ResumeConfig{
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
	enqueue, err := enqueueRuntime.Enqueue(ctx, request)
	if err != nil || !enqueue.TickEnqueued {
		return "", fmt.Errorf("resume enqueue blocked: report=%#v err=%v", enqueue, err)
	}
	if err := enqueueRuntime.Shutdown(context.Background()); err != nil {
		return "", err
	}

	// A fresh construction consumes the tick from the Host-owned queue. This is
	// the minimum cross-process-shaped resume proof; production Hosts replace
	// MemoryQueue with a durable backend while keeping the same seam.
	workerRuntime, err := sessionhostkit.NewResumeRuntime(sessionhostkit.ResumeConfig{
		Queue:  queue,
		Worker: readyResumeWorker(),
		Lane:   scheduler.LaneBackground,
	})
	if err != nil {
		return "", err
	}
	report, err := workerRuntime.Run(ctx, sessionhostkit.ResumeRunRequest{
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
	if err := workerRuntime.Shutdown(context.Background()); err != nil {
		return "", err
	}
	if err := workerRuntime.Shutdown(context.Background()); err != nil {
		return "", err
	}
	if _, err := workerRuntime.Enqueue(context.Background(), request); !errors.Is(err, sessionhostkit.ErrResumeRuntimeClosed) {
		return "", fmt.Errorf("resume closed call error = %v", err)
	}
	return fmt.Sprintf("agentx-resume-hostkit-ok:%s:%d:%t:cross-construction", report.Status, report.TicksAcked, report.HostRuntimeDispatchByHost), nil
}
