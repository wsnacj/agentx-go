package hostkit

import (
	"context"
	"errors"
	"sync"

	scheduler "github.com/wsnacj/agentx-go/runtime/scheduler"
	resume "github.com/wsnacj/agentx-go/runtime/session/resume"
)

var (
	ErrInvalidResumeConfig = errors.New("agentx session hostkit: invalid resume config")
	ErrResumeRuntimeClosed = errors.New("agentx session hostkit: resume runtime closed")
	ErrResumeRuntimeBusy   = errors.New("agentx session hostkit: resume runtime already running")
)

// ResumeConfig wires the portable scheduler and wake-continuation mechanism.
// The Host still owns the concrete continuation store, runtime dispatch,
// authorization, credentials, process lifecycle, and durable queue backend.
type ResumeConfig struct {
	Queue  scheduler.Queue
	Worker resume.Worker
	Lane   scheduler.Lane
	Wait   resume.ServiceWaitFunc
}

type ResumeEnqueueRequest = resume.EnqueueRequest
type ResumeEnqueueResult = resume.EnqueueResult
type ResumeRunRequest = resume.ServiceInput
type ResumeRunResult = resume.ServiceReport

// ResumeRuntime is a bounded Host Kit for enqueueing wake-continuation ticks
// and processing them through a single service loop.
type ResumeRuntime struct {
	service resume.Service

	mu      sync.Mutex
	closed  bool
	running bool
	cancel  context.CancelFunc
	done    chan struct{}
}

func NewResumeRuntime(cfg ResumeConfig) (*ResumeRuntime, error) {
	if cfg.Queue == nil || cfg.Worker.ContinuationReadback == nil || cfg.Worker.WakeDispatch == nil {
		return nil, ErrInvalidResumeConfig
	}
	daemon := resume.Daemon{
		Queue:  cfg.Queue,
		Worker: cfg.Worker,
		Config: resume.DaemonConfig{Lane: cfg.Lane},
	}
	return &ResumeRuntime{
		service: resume.Service{Daemon: daemon, Wait: cfg.Wait},
	}, nil
}

// Enqueue records a display-safe resume tick in the Host-provided queue.
func (r *ResumeRuntime) Enqueue(ctx context.Context, request ResumeEnqueueRequest) (ResumeEnqueueResult, error) {
	if r == nil {
		return ResumeEnqueueResult{}, ErrInvalidResumeConfig
	}
	r.mu.Lock()
	closed := r.closed
	r.mu.Unlock()
	if closed {
		return ResumeEnqueueResult{}, ErrResumeRuntimeClosed
	}
	return r.service.Daemon.EnqueueSchedulerTick(ctx, request), nil
}

// Run processes a bounded resume service loop. Only one loop may execute at a
// time; callers can continue enqueueing through the concurrent-safe Queue.
func (r *ResumeRuntime) Run(ctx context.Context, request ResumeRunRequest) (ResumeRunResult, error) {
	if r == nil {
		return ResumeRunResult{}, ErrInvalidResumeConfig
	}
	if ctx == nil {
		ctx = context.Background()
	}
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return ResumeRunResult{}, ErrResumeRuntimeClosed
	}
	if r.running {
		r.mu.Unlock()
		return ResumeRunResult{}, ErrResumeRuntimeBusy
	}
	runCtx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	r.running = true
	r.cancel = cancel
	r.done = done
	r.mu.Unlock()

	result := r.service.Run(runCtx, request)
	cancel()
	close(done)
	r.mu.Lock()
	if r.done == done {
		r.running = false
		r.cancel = nil
		r.done = nil
	}
	r.mu.Unlock()
	return result, nil
}

// Shutdown is idempotent. It prevents future Enqueue/Run calls, cancels an
// active service loop, and waits until that loop exits or ctx expires.
func (r *ResumeRuntime) Shutdown(ctx context.Context) error {
	if r == nil {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	r.mu.Lock()
	r.closed = true
	cancel := r.cancel
	done := r.done
	r.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	if done == nil {
		return nil
	}
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
