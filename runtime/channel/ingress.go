package channel

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	agentxsafeerror "github.com/wsnacj/agentx-go/runtime/telemetry/safeerror"
)

// ErrIngressRuntimeClosed cancels active channel work during shutdown.
var ErrIngressRuntimeClosed = errors.New("channel ingress runtime is closing or closed")

// IngressOverloadPolicy controls submission behavior when the ingress queue is full.
type IngressOverloadPolicy string

const (
	IngressOverloadReject IngressOverloadPolicy = "reject"
	IngressOverloadWait   IngressOverloadPolicy = "wait"
)

// IngressSubmitReason is the stable rejection or acceptance classification for a submission.
type IngressSubmitReason string

const (
	IngressSubmitAccepted           IngressSubmitReason = "accepted"
	IngressSubmitDuplicate          IngressSubmitReason = "duplicate"
	IngressSubmitOverloaded         IngressSubmitReason = "overloaded"
	IngressSubmitClosed             IngressSubmitReason = "closed"
	IngressSubmitRuntimeUnavailable IngressSubmitReason = "runtime_unavailable"
	IngressSubmitContextCanceled    IngressSubmitReason = "context_canceled"
)

// IngressSubmitResult reports whether ownership of a message moved to the runtime.
type IngressSubmitResult struct {
	Accepted bool                `json:"accepted"`
	Reason   IngressSubmitReason `json:"reason"`
}

// IngressRuntimeOptions configures worker, queue, overload, and shutdown bounds.
type IngressRuntimeOptions struct {
	MaxConcurrency int
	QueueCapacity  int
	OverloadPolicy IngressOverloadPolicy
	CloseTimeout   time.Duration
}

const (
	defaultIngressMaxConcurrency = 4
	defaultIngressQueueCapacity  = 64
	defaultIngressCloseTimeout   = 10 * time.Second
)

type ingressWork struct {
	ctx          context.Context
	processor    InboundProcessor
	message      Message
	reservations []DedupeReservation
}

// IngressRuntime owns a bounded shared worker pool for asynchronous channel messages.
type IngressRuntime struct {
	mu        sync.Mutex
	logMu     sync.Mutex
	options   IngressRuntimeOptions
	queue     chan ingressWork
	ctx       context.Context
	cancel    context.CancelCauseFunc
	accepting bool
	workers   sync.WaitGroup
	closeOnce sync.Once
	closeDone chan struct{}
}

// NewIngressRuntime starts a bounded ingress runtime. Call Close when its host stops.
func NewIngressRuntime(opts IngressRuntimeOptions) *IngressRuntime {
	opts = normalizeIngressRuntimeOptions(opts)
	ctx, cancel := context.WithCancelCause(context.Background())
	runtime := &IngressRuntime{
		options:   opts,
		queue:     make(chan ingressWork, opts.QueueCapacity),
		ctx:       ctx,
		cancel:    cancel,
		accepting: true,
		closeDone: make(chan struct{}),
	}
	for i := 0; i < opts.MaxConcurrency; i++ {
		runtime.workers.Add(1)
		go runtime.worker()
	}
	return runtime
}

func normalizeIngressRuntimeOptions(opts IngressRuntimeOptions) IngressRuntimeOptions {
	if opts.MaxConcurrency <= 0 {
		opts.MaxConcurrency = defaultIngressMaxConcurrency
	}
	if opts.QueueCapacity <= 0 {
		opts.QueueCapacity = defaultIngressQueueCapacity
	}
	switch IngressOverloadPolicy(strings.ToLower(strings.TrimSpace(string(opts.OverloadPolicy)))) {
	case IngressOverloadWait:
		opts.OverloadPolicy = IngressOverloadWait
	default:
		opts.OverloadPolicy = IngressOverloadReject
	}
	if opts.CloseTimeout <= 0 {
		opts.CloseTimeout = defaultIngressCloseTimeout
	}
	return opts
}

// Submit reserves dedupe keys and transfers accepted work to the shared runtime.
func (r *IngressRuntime) Submit(ctx context.Context, processor InboundProcessor, message Message) IngressSubmitResult {
	if r == nil {
		return IngressSubmitResult{Reason: IngressSubmitRuntimeUnavailable}
	}
	processor.Logger = r.synchronizedLogger(processor.Logger)
	reservations, ok := processor.beginDedupe(message)
	if !ok {
		return IngressSubmitResult{Reason: IngressSubmitDuplicate}
	}
	baseCtx := context.Background()
	if ctx != nil {
		baseCtx = context.WithoutCancel(ctx)
	}
	work := ingressWork{
		ctx:          baseCtx,
		processor:    processor,
		message:      message,
		reservations: reservations,
	}
	result := r.enqueue(ctx, work)
	if !result.Accepted {
		processor.releaseDedupe(reservations)
	}
	return result
}

func (r *IngressRuntime) synchronizedLogger(logger EventLogger) EventLogger {
	if r == nil || logger == nil {
		return logger
	}
	return func(format string, args ...any) {
		r.logMu.Lock()
		defer r.logMu.Unlock()
		logger(format, args...)
	}
}

func (r *IngressRuntime) log(logger EventLogger, format string, args ...any) {
	if logger := r.synchronizedLogger(logger); logger != nil {
		logger(format, args...)
	}
}

func (r *IngressRuntime) enqueue(ctx context.Context, work ingressWork) IngressSubmitResult {
	if r.options.OverloadPolicy == IngressOverloadWait {
		r.mu.Lock()
		accepting := r.accepting
		r.mu.Unlock()
		if !accepting {
			return IngressSubmitResult{Reason: IngressSubmitClosed}
		}
		submitCtx := context.Background()
		if ctx != nil {
			submitCtx = ctx
		}
		select {
		case r.queue <- work:
			return IngressSubmitResult{Accepted: true, Reason: IngressSubmitAccepted}
		case <-r.ctx.Done():
			return IngressSubmitResult{Reason: IngressSubmitClosed}
		case <-submitCtx.Done():
			return IngressSubmitResult{Reason: IngressSubmitContextCanceled}
		}
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.accepting {
		return IngressSubmitResult{Reason: IngressSubmitClosed}
	}
	select {
	case r.queue <- work:
		return IngressSubmitResult{Accepted: true, Reason: IngressSubmitAccepted}
	default:
		return IngressSubmitResult{Reason: IngressSubmitOverloaded}
	}
}

func (r *IngressRuntime) worker() {
	defer r.workers.Done()
	for {
		select {
		case <-r.ctx.Done():
			r.releaseQueued()
			return
		case work := <-r.queue:
			if r.ctx.Err() != nil {
				work.processor.releaseDedupe(work.reservations)
				continue
			}
			r.process(work)
		}
	}
}

func (r *IngressRuntime) process(work ingressWork) {
	runCtx, cancel := context.WithCancelCause(work.ctx)
	stopRuntimeCancel := context.AfterFunc(r.ctx, func() {
		cancel(ErrIngressRuntimeClosed)
	})
	err := work.processor.processReserved(runCtx, work.message, work.reservations)
	stopRuntimeCancel()
	cancel(context.Canceled)
	if err != nil && work.processor.Logger != nil {
		projection := agentxsafeerror.Project(err, "channel_ingress", "processing_failed")
		work.processor.Logger(
			"channel event failed platform=%s account=%s session=%s chat=%s msg=%s error_class=%s error_code=%s error_identity=%s\n",
			work.message.Platform,
			work.message.AccountID,
			work.message.SessionID,
			work.message.ChatID,
			work.message.MessageID,
			projection.Class,
			projection.Code,
			projection.Identity,
		)
	}
}

func (r *IngressRuntime) releaseQueued() {
	for {
		select {
		case work := <-r.queue:
			work.processor.releaseDedupe(work.reservations)
		default:
			return
		}
	}
}

// Close stops acceptance, cancels active work, releases queued work, and waits up to CloseTimeout.
func (r *IngressRuntime) Close() error {
	if r == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), r.options.CloseTimeout)
	defer cancel()
	return r.Shutdown(ctx)
}

// Shutdown performs Close semantics with a caller-provided wait bound.
func (r *IngressRuntime) Shutdown(ctx context.Context) error {
	if r == nil {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	r.closeOnce.Do(func() {
		r.mu.Lock()
		r.accepting = false
		r.mu.Unlock()
		r.cancel(ErrIngressRuntimeClosed)
		r.releaseQueued()
		go func() {
			r.workers.Wait()
			close(r.closeDone)
		}()
	})
	select {
	case <-r.closeDone:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("channel ingress shutdown: %w", ctx.Err())
	}
}
