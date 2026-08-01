package channel

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestIngressRuntimeBoundsConcurrencyAndRejectsOverload(t *testing.T) {
	runner := &ingressBlockingRunner{
		started: make(chan Message, 8),
		release: make(chan struct{}),
	}
	runtime := NewIngressRuntime(IngressRuntimeOptions{
		MaxConcurrency: 2,
		QueueCapacity:  2,
		OverloadPolicy: IngressOverloadReject,
		CloseTimeout:   2 * time.Second,
	})
	deduper := NewDeduper(time.Minute)
	var logs strings.Builder
	processor := InboundProcessor{
		Runner:  runner,
		Deduper: deduper,
		Ingress: runtime,
		Logger:  NewWriterLogger(&logs),
		BuildReservations: func(message Message) []DedupeReservation {
			return []DedupeReservation{{Key: message.MessageID}}
		},
	}

	for _, id := range []string{"m1", "m2"} {
		if result := processor.SubmitAsync(context.Background(), Message{MessageID: id}); !result.Accepted {
			t.Fatalf("submit %s = %#v, want accepted", id, result)
		}
	}
	waitIngressStarts(t, runner.started, 2)
	for _, id := range []string{"m3", "m4"} {
		if result := processor.SubmitAsync(context.Background(), Message{MessageID: id}); !result.Accepted {
			t.Fatalf("queue %s = %#v, want accepted", id, result)
		}
	}
	result := processor.SubmitAsync(context.Background(), Message{MessageID: "m5"})
	if result.Accepted || result.Reason != IngressSubmitOverloaded {
		t.Fatalf("overload result = %#v, want overloaded rejection", result)
	}
	if !deduper.Begin("m5") {
		t.Fatal("overload rejection did not release dedupe reservation")
	}
	if !strings.Contains(logs.String(), "reason=overloaded") {
		t.Fatalf("missing structured overload reason in log: %q", logs.String())
	}

	close(runner.release)
	if err := runtime.Close(); err != nil {
		t.Fatalf("close ingress runtime: %v", err)
	}
	if got := runner.max.Load(); got != 2 {
		t.Fatalf("peak concurrency = %d, want 2", got)
	}
}

func TestIngressRuntimeLoggerProjectsProcessingError(t *testing.T) {
	secret := "channel-ingress-secret-sentinel"
	logs := make(chan string, 1)
	runtime := NewIngressRuntime(IngressRuntimeOptions{MaxConcurrency: 1, QueueCapacity: 1})
	processor := InboundProcessor{
		Runner:  ingressErrorRunner{err: errors.New(secret)},
		Ingress: runtime,
		Logger: func(format string, args ...any) {
			logs <- fmt.Sprintf(format, args...)
		},
	}
	if result := processor.SubmitAsync(context.Background(), Message{MessageID: "safe-error"}); !result.Accepted {
		t.Fatalf("submit: %#v", result)
	}
	var observed string
	select {
	case observed = <-logs:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for ingress log")
	}
	if err := runtime.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	if strings.Contains(observed, secret) || !strings.Contains(observed, "error_identity=") {
		t.Fatalf("unsafe ingress log: %q", observed)
	}
}

func TestIngressRuntimePreservesFIFOQueueOrder(t *testing.T) {
	runner := &ingressOrderedRunner{
		firstStarted: make(chan struct{}),
		releaseFirst: make(chan struct{}),
		done:         make(chan struct{}, 3),
	}
	runtime := NewIngressRuntime(IngressRuntimeOptions{MaxConcurrency: 1, QueueCapacity: 3})
	processor := InboundProcessor{Runner: runner, Ingress: runtime}

	if result := processor.SubmitAsync(context.Background(), Message{MessageID: "m1"}); !result.Accepted {
		t.Fatalf("submit m1: %#v", result)
	}
	select {
	case <-runner.firstStarted:
	case <-time.After(time.Second):
		t.Fatal("first message did not start")
	}
	for _, id := range []string{"m2", "m3"} {
		if result := processor.SubmitAsync(context.Background(), Message{MessageID: id}); !result.Accepted {
			t.Fatalf("submit %s: %#v", id, result)
		}
	}
	close(runner.releaseFirst)
	for i := 0; i < 3; i++ {
		select {
		case <-runner.done:
		case <-time.After(time.Second):
			t.Fatal("queued message did not complete")
		}
	}
	if err := runtime.Close(); err != nil {
		t.Fatalf("close ingress runtime: %v", err)
	}
	if got := runner.snapshot(); len(got) != 3 || got[0] != "m1" || got[1] != "m2" || got[2] != "m3" {
		t.Fatalf("execution order = %#v, want FIFO", got)
	}
}

func TestIngressRuntimeShutdownCancelsActiveAndReleasesQueuedReservations(t *testing.T) {
	runner := &ingressBlockingRunner{
		started: make(chan Message, 2),
		release: make(chan struct{}),
	}
	runtime := NewIngressRuntime(IngressRuntimeOptions{MaxConcurrency: 1, QueueCapacity: 2})
	deduper := NewDeduper(time.Minute)
	processor := InboundProcessor{
		Runner:  runner,
		Deduper: deduper,
		Ingress: runtime,
		BuildReservations: func(message Message) []DedupeReservation {
			return []DedupeReservation{{Key: message.MessageID}}
		},
	}
	if result := processor.SubmitAsync(context.Background(), Message{MessageID: "active"}); !result.Accepted {
		t.Fatalf("submit active: %#v", result)
	}
	waitIngressStarts(t, runner.started, 1)
	if result := processor.SubmitAsync(context.Background(), Message{MessageID: "queued"}); !result.Accepted {
		t.Fatalf("submit queued: %#v", result)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := runtime.Shutdown(ctx); err != nil {
		t.Fatalf("shutdown ingress runtime: %v", err)
	}
	if !deduper.Begin("active") || !deduper.Begin("queued") {
		t.Fatal("shutdown did not release active and queued dedupe reservations")
	}
	result := processor.SubmitAsync(context.Background(), Message{MessageID: "after-close"})
	if result.Accepted || result.Reason != IngressSubmitClosed {
		t.Fatalf("submit after shutdown = %#v, want closed", result)
	}
	if cause := runner.lastCause(); !errors.Is(cause, ErrIngressRuntimeClosed) {
		t.Fatalf("active runner cancellation cause = %v, want ErrIngressRuntimeClosed", cause)
	}
}

func TestIngressRuntimeWaitPolicyHonorsSubmitContextAndReleasesDedupe(t *testing.T) {
	runner := &ingressBlockingRunner{
		started: make(chan Message, 2),
		release: make(chan struct{}),
	}
	runtime := NewIngressRuntime(IngressRuntimeOptions{
		MaxConcurrency: 1,
		QueueCapacity:  1,
		OverloadPolicy: IngressOverloadWait,
	})
	defer runtime.Close()
	deduper := NewDeduper(time.Minute)
	processor := InboundProcessor{
		Runner:  runner,
		Deduper: deduper,
		Ingress: runtime,
		BuildReservations: func(message Message) []DedupeReservation {
			return []DedupeReservation{{Key: message.MessageID}}
		},
	}
	if result := processor.SubmitAsync(context.Background(), Message{MessageID: "active"}); !result.Accepted {
		t.Fatalf("submit active: %#v", result)
	}
	waitIngressStarts(t, runner.started, 1)
	if result := processor.SubmitAsync(context.Background(), Message{MessageID: "queued"}); !result.Accepted {
		t.Fatalf("submit queued: %#v", result)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	result := processor.SubmitAsync(ctx, Message{MessageID: "waiting"})
	if result.Accepted || result.Reason != IngressSubmitContextCanceled {
		t.Fatalf("wait-policy result = %#v, want context cancellation", result)
	}
	if !deduper.Begin("waiting") {
		t.Fatal("canceled wait did not release dedupe reservation")
	}
	close(runner.release)
}

func TestInboundProcessorAsyncRequiresOwnedRuntime(t *testing.T) {
	var logs strings.Builder
	processor := InboundProcessor{
		Runner: workerReplyRunner{reply: "ignored"},
		Logger: NewWriterLogger(&logs),
	}
	result := processor.SubmitAsync(context.Background(), Message{MessageID: "m1"})
	if result.Accepted || result.Reason != IngressSubmitRuntimeUnavailable {
		t.Fatalf("submit without runtime = %#v, want fail closed", result)
	}
	if !strings.Contains(logs.String(), "reason=runtime_unavailable") {
		t.Fatalf("missing runtime-unavailable log: %q", logs.String())
	}
}

func TestIngressRuntimeRetainsPerMessageTimeout(t *testing.T) {
	runner := &ingressCauseRunner{done: make(chan struct{})}
	runtime := NewIngressRuntime(IngressRuntimeOptions{MaxConcurrency: 1, QueueCapacity: 1})
	processor := InboundProcessor{Runner: runner, Ingress: runtime, Timeout: 30 * time.Millisecond}
	if result := processor.SubmitAsync(context.Background(), Message{MessageID: "timeout"}); !result.Accepted {
		t.Fatalf("submit timeout message: %#v", result)
	}
	select {
	case <-runner.done:
	case <-time.After(time.Second):
		t.Fatal("per-message timeout did not cancel runner")
	}
	if err := runtime.Close(); err != nil {
		t.Fatalf("close ingress runtime: %v", err)
	}
	if cause := runner.cause(); !errors.Is(cause, context.DeadlineExceeded) {
		t.Fatalf("runner cause = %v, want deadline exceeded", cause)
	}
}

func TestIngressRuntimeShutdownIsBoundedForNonCooperativeRunner(t *testing.T) {
	runner := &ingressNonCooperativeRunner{
		started: make(chan struct{}),
		release: make(chan struct{}),
	}
	runtime := NewIngressRuntime(IngressRuntimeOptions{
		MaxConcurrency: 1,
		QueueCapacity:  2,
		CloseTimeout:   30 * time.Millisecond,
	})
	deduper := NewDeduper(time.Minute)
	processor := InboundProcessor{
		Runner:  runner,
		Deduper: deduper,
		Ingress: runtime,
		BuildReservations: func(message Message) []DedupeReservation {
			return []DedupeReservation{{Key: message.MessageID}}
		},
	}
	if result := processor.SubmitAsync(context.Background(), Message{MessageID: "active"}); !result.Accepted {
		t.Fatalf("submit active: %#v", result)
	}
	select {
	case <-runner.started:
	case <-time.After(time.Second):
		t.Fatal("non-cooperative runner did not start")
	}
	if result := processor.SubmitAsync(context.Background(), Message{MessageID: "queued"}); !result.Accepted {
		t.Fatalf("submit queued: %#v", result)
	}
	if err := runtime.Close(); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("bounded close error = %v, want deadline exceeded", err)
	}
	if !deduper.Begin("queued") {
		t.Fatal("bounded close did not immediately release queued reservation")
	}
	close(runner.release)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := runtime.Shutdown(ctx); err != nil {
		t.Fatalf("finish shutdown after runner release: %v", err)
	}
	if !deduper.Begin("active") {
		t.Fatal("stale non-cooperative completion should release active reservation")
	}
}

type ingressBlockingRunner struct {
	started chan Message
	release chan struct{}
	current atomic.Int64
	max     atomic.Int64
	mu      sync.Mutex
	cause   error
}

type ingressErrorRunner struct{ err error }

func (r ingressErrorRunner) RunTurn(context.Context, Message) (string, error) {
	return "", r.err
}

func (ingressErrorRunner) WorkspaceDir() string { return "." }
func (ingressErrorRunner) Profile() string      { return "safe" }

type ingressNonCooperativeRunner struct {
	started chan struct{}
	release chan struct{}
	once    sync.Once
}

func (r *ingressNonCooperativeRunner) RunTurn(context.Context, Message) (string, error) {
	r.once.Do(func() { close(r.started) })
	<-r.release
	return "", nil
}

func (r *ingressNonCooperativeRunner) WorkspaceDir() string { return "." }
func (r *ingressNonCooperativeRunner) Profile() string      { return "safe" }

func (r *ingressBlockingRunner) RunTurn(ctx context.Context, message Message) (string, error) {
	current := r.current.Add(1)
	defer r.current.Add(-1)
	for {
		max := r.max.Load()
		if current <= max || r.max.CompareAndSwap(max, current) {
			break
		}
	}
	r.started <- message
	select {
	case <-r.release:
		return "", nil
	case <-ctx.Done():
		cause := context.Cause(ctx)
		r.mu.Lock()
		r.cause = cause
		r.mu.Unlock()
		return "", cause
	}
}

func (r *ingressBlockingRunner) WorkspaceDir() string { return "." }
func (r *ingressBlockingRunner) Profile() string      { return "safe" }
func (r *ingressBlockingRunner) lastCause() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.cause
}

type ingressOrderedRunner struct {
	firstStarted chan struct{}
	releaseFirst chan struct{}
	done         chan struct{}
	firstOnce    sync.Once
	mu           sync.Mutex
	order        []string
}

func (r *ingressOrderedRunner) RunTurn(_ context.Context, message Message) (string, error) {
	r.mu.Lock()
	r.order = append(r.order, message.MessageID)
	r.mu.Unlock()
	if message.MessageID == "m1" {
		r.firstOnce.Do(func() { close(r.firstStarted) })
		<-r.releaseFirst
	}
	r.done <- struct{}{}
	return "", nil
}

func (r *ingressOrderedRunner) WorkspaceDir() string { return "." }
func (r *ingressOrderedRunner) Profile() string      { return "safe" }
func (r *ingressOrderedRunner) snapshot() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.order...)
}

type ingressCauseRunner struct {
	done      chan struct{}
	doneOnce  sync.Once
	mu        sync.Mutex
	lastCause error
}

func (r *ingressCauseRunner) RunTurn(ctx context.Context, _ Message) (string, error) {
	<-ctx.Done()
	cause := context.Cause(ctx)
	r.mu.Lock()
	r.lastCause = cause
	r.mu.Unlock()
	r.doneOnce.Do(func() { close(r.done) })
	return "", cause
}

func (r *ingressCauseRunner) WorkspaceDir() string { return "." }
func (r *ingressCauseRunner) Profile() string      { return "safe" }
func (r *ingressCauseRunner) cause() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.lastCause
}

func waitIngressStarts(t *testing.T, started <-chan Message, count int) {
	t.Helper()
	for i := 0; i < count; i++ {
		select {
		case <-started:
		case <-time.After(time.Second):
			t.Fatalf("timed out waiting for ingress start %d/%d", i+1, count)
		}
	}
}
