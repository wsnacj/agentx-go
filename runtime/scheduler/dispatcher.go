package scheduler

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	agentxsafeerror "github.com/wsnacj/agentx-go/runtime/telemetry/safeerror"
)

type DispatcherConfig struct {
	LaneConcurrency map[Lane]int
	PollInterval    time.Duration
	// TerminalTimeout bounds durable Ack/Fail work after the run context ends.
	TerminalTimeout time.Duration
	// HeartbeatFailureObserver lets a Host preserve telemetry without making
	// the portable scheduler depend on a concrete logger.
	HeartbeatFailureObserver func(context.Context, HeartbeatFailure)
}

// HeartbeatFailure is the display-safe scheduler lease failure projection
// delivered to an optional Host observer.
type HeartbeatFailure struct {
	JobID             string
	Lane              Lane
	Attempt           int
	HeartbeatInterval time.Duration
	Err               error
}

type Dispatcher struct {
	queue                   Queue
	handlers                *handlerRegistry
	metrics                 *Metrics
	poll                    time.Duration
	terminal                time.Duration
	workers                 map[Lane]int
	wg                      sync.WaitGroup
	runOnce                 sync.Once
	started                 atomic.Bool
	done                    chan struct{}
	errMu                   sync.Mutex
	waitErr                 error
	errCount                uint64
	observeHeartbeatFailure func(context.Context, HeartbeatFailure)
}

func NewDispatcher(queue Queue, cfg DispatcherConfig) *Dispatcher {
	workers := map[Lane]int{
		LaneMain:       1,
		LaneSubtask:    3,
		LaneBackground: 2,
	}
	for lane, value := range cfg.LaneConcurrency {
		if value <= 0 {
			continue
		}
		workers[lane] = value
	}
	poll := cfg.PollInterval
	if poll <= 0 {
		poll = 20 * time.Millisecond
	}
	terminal := cfg.TerminalTimeout
	if terminal <= 0 {
		terminal = defaultTerminalPersistenceTimeout
	}
	return &Dispatcher{
		queue:                   queue,
		handlers:                newHandlerRegistry(),
		metrics:                 &Metrics{},
		poll:                    poll,
		terminal:                terminal,
		workers:                 workers,
		done:                    make(chan struct{}),
		observeHeartbeatFailure: cfg.HeartbeatFailureObserver,
	}
}

func (d *Dispatcher) RegisterHandler(lane Lane, handler Handler) error {
	if !isKnownLane(lane) {
		return ErrInvalidLane
	}
	if handler == nil {
		return ErrInvalidHandler
	}
	d.handlers.Set(lane, handler)
	return nil
}

func (d *Dispatcher) Enqueue(ctx context.Context, job Job) error {
	if strings.TrimSpace(job.ID) == "" {
		return ErrUnknownJob
	}
	if err := d.queue.Enqueue(ctx, job); err != nil {
		return err
	}
	d.metrics.RecordEnqueue(job.Lane)
	return nil
}

func (d *Dispatcher) Run(ctx context.Context) {
	if d == nil {
		return
	}
	d.runOnce.Do(func() {
		d.started.Store(true)
		for _, lane := range []Lane{LaneMain, LaneSubtask, LaneBackground} {
			workerCount := d.workers[lane]
			if workerCount <= 0 {
				continue
			}
			for i := 0; i < workerCount; i++ {
				d.wg.Add(1)
				go d.worker(ctx, lane)
			}
		}
		go func() {
			d.wg.Wait()
			close(d.done)
		}()
	})
}

// Wait blocks until every worker and its bounded terminal persistence work exits.
func (d *Dispatcher) Wait() error {
	if d == nil {
		return nil
	}
	if !d.started.Load() {
		return nil
	}
	<-d.done
	return d.waitError()
}

func (d *Dispatcher) WaitContext(ctx context.Context) error {
	if d == nil {
		return nil
	}
	if !d.started.Load() {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	select {
	case <-d.done:
		return d.waitError()
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (d *Dispatcher) Metrics() map[Lane]LaneMetrics {
	return d.metrics.Snapshot()
}

func (d *Dispatcher) worker(ctx context.Context, lane Lane) {
	defer d.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		job, err := d.queue.Dequeue(ctx, lane)
		if err != nil {
			if err == ErrQueueEmpty {
				select {
				case <-ctx.Done():
					return
				case <-time.After(d.poll):
				}
				continue
			}
			select {
			case <-ctx.Done():
				return
			case <-time.After(d.poll):
			}
			continue
		}
		d.metrics.RecordDequeue(lane)
		handler, ok := d.handlers.Get(lane)
		if !ok || handler == nil {
			outcome := ResultOutcomeFailed
			failureErr := ErrInvalidHandler
			if ctx.Err() != nil {
				outcome = ResultOutcomeCanceled
				failureErr = context.Cause(ctx)
				if failureErr == nil {
					failureErr = ctx.Err()
				}
			}
			d.failJobTerminal(ctx, lane, "missing_handler", Result{
				JobID:   job.ID,
				Lane:    lane,
				Attempt: job.Attempt,
				Outcome: outcome,
				Error:   schedulerFailureSummary(failureErr, schedulerFailureCode(failureErr)),
			})
			d.metrics.RecordFail(lane)
			continue
		}
		attemptBaseCtx, cancelAttempt := context.WithCancelCause(ctx)
		heartbeat := d.newLeaseHeartbeat(job, lane, cancelAttempt)
		attemptCtx := withExecutionAttempt(attemptBaseCtx, d.executionAttempt(job), heartbeat.Validate)
		heartbeat.Start(attemptCtx)
		handlerErr := invokeSchedulerHandler(attemptCtx, handler, job)
		heartbeat.Stop()
		leaseErr := heartbeat.Err()
		if leaseErr != nil {
			if handlerErr == nil {
				d.metrics.RecordStaleHandlerCompletion(lane)
			} else if errors.Is(handlerErr, context.Canceled) || errors.Is(handlerErr, ErrLeaseLost) {
				d.metrics.RecordLeaseCancellationObserved(lane)
			}
			d.failJobTerminal(ctx, lane, "lease_lost", Result{
				JobID:   job.ID,
				Lane:    lane,
				Attempt: job.Attempt,
				Outcome: ResultOutcomeFailed,
				Error:   schedulerFailureSummary(leaseErr, "lease_lost"),
			})
			d.metrics.RecordFail(lane)
			cancelAttempt(nil)
			continue
		}
		if ctx.Err() != nil {
			rootErr := context.Cause(ctx)
			if rootErr == nil {
				rootErr = ctx.Err()
			}
			cancelAttempt(nil)
			d.failJobTerminal(ctx, lane, "root_canceled", Result{
				JobID:   job.ID,
				Lane:    lane,
				Attempt: job.Attempt,
				Outcome: ResultOutcomeCanceled,
				Error:   schedulerFailureSummary(rootErr, schedulerFailureCode(rootErr)),
			})
			d.metrics.RecordFail(lane)
			continue
		}
		cancelAttempt(nil)
		if handlerErr != nil {
			outcome := ResultOutcomeFailed
			if errors.Is(handlerErr, context.Canceled) {
				outcome = ResultOutcomeCanceled
			}
			d.failJobTerminal(ctx, lane, "handler_result", Result{
				JobID:   job.ID,
				Lane:    lane,
				Attempt: job.Attempt,
				Outcome: outcome,
				Error:   schedulerFailureSummary(handlerErr, schedulerFailureCode(handlerErr)),
			})
			d.metrics.RecordFail(lane)
			continue
		}
		ackResult := Result{
			JobID:   job.ID,
			Lane:    lane,
			Attempt: job.Attempt,
			Outcome: ResultOutcomeCompleted,
		}
		if err := d.ackJobTerminal(ctx, lane, ackResult); err != nil {
			d.metrics.RecordAckFailed(lane)
			d.failJobTerminal(ctx, lane, "ack_compensation", Result{
				JobID:   job.ID,
				Lane:    lane,
				Attempt: job.Attempt,
				Outcome: ResultOutcomeFailed,
				Error:   schedulerFailureSummary(err, "ack_failed"),
			})
			d.metrics.RecordFail(lane)
			continue
		}
		d.metrics.RecordAck(lane)
	}
}

var errSchedulerHandlerPanic = errors.New("scheduler handler panic")

func invokeSchedulerHandler(ctx context.Context, handler Handler, job Job) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = agentxsafeerror.WrapWithIdentity(
				errSchedulerHandlerPanic,
				"scheduler handler panic",
				agentxsafeerror.Identity(fmt.Sprint(recovered)),
			)
		}
	}()
	return handler(ctx, job)
}

func schedulerFailureCode(err error) string {
	switch {
	case errors.Is(err, errSchedulerHandlerPanic):
		return "handler_panic"
	case errors.Is(err, context.Canceled):
		return "canceled"
	case errors.Is(err, context.DeadlineExceeded):
		return "deadline_exceeded"
	case errors.Is(err, ErrLeaseLost):
		return "lease_lost"
	default:
		return "handler_failed"
	}
}

func schedulerFailureSummary(err error, code string) string {
	return agentxsafeerror.Summary(agentxsafeerror.Project(err, "scheduler", code))
}

func (d *Dispatcher) recordPostFailOutcome(ctx context.Context, lane Lane, jobID string) {
	if d == nil || d.queue == nil {
		return
	}
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return
	}
	if result, ok, err := d.queue.Result(ctx, jobID); err == nil && ok {
		if strings.EqualFold(strings.TrimSpace(result.Status), "dead_letter") {
			d.metrics.RecordDeadLetter(lane)
			return
		}
	}
	if pending, err := d.queue.Pending(ctx, jobID); err == nil && pending {
		d.metrics.RecordRequeue(lane)
	}
}

func (d *Dispatcher) executionAttempt(job Job) ExecutionAttempt {
	attempt := ExecutionAttempt{
		JobID:          strings.TrimSpace(job.ID),
		Lane:           job.Lane,
		Attempt:        job.Attempt,
		IdempotencyKey: strings.TrimSpace(job.IdempotencyKey),
	}
	if provider, ok := d.queue.(LeaseIdentityProvider); ok && provider != nil {
		attempt.LeaseOwner = strings.TrimSpace(provider.LeaseIdentity(job))
	}
	return attempt
}

type leaseHeartbeat struct {
	dispatcher    *Dispatcher
	queue         HeartbeatCapableQueue
	job           Job
	lane          Lane
	interval      time.Duration
	cancelAttempt context.CancelCauseFunc

	startOnce sync.Once
	stopOnce  sync.Once
	failOnce  sync.Once
	cancel    context.CancelFunc
	done      chan struct{}

	errMu sync.RWMutex
	err   error
}

func (d *Dispatcher) newLeaseHeartbeat(job Job, lane Lane, cancelAttempt context.CancelCauseFunc) *leaseHeartbeat {
	heartbeat := &leaseHeartbeat{
		dispatcher:    d,
		job:           job,
		lane:          lane,
		cancelAttempt: cancelAttempt,
		done:          make(chan struct{}),
	}
	if queue, ok := d.queue.(HeartbeatCapableQueue); ok && queue != nil {
		heartbeat.queue = queue
		heartbeat.interval = queue.HeartbeatInterval()
	}
	return heartbeat
}

func (h *leaseHeartbeat) Start(ctx context.Context) {
	if h == nil {
		return
	}
	h.startOnce.Do(func() {
		if h.queue == nil || h.interval <= 0 {
			close(h.done)
			return
		}
		heartbeatCtx, cancel := context.WithCancel(ctx)
		h.cancel = cancel
		go func() {
			defer close(h.done)
			ticker := time.NewTicker(h.interval)
			defer ticker.Stop()
			for {
				select {
				case <-heartbeatCtx.Done():
					return
				case <-ticker.C:
					if err := h.Validate(heartbeatCtx); err != nil {
						return
					}
				}
			}
		}()
	})
}

func (h *leaseHeartbeat) Validate(ctx context.Context) error {
	if h == nil || h.queue == nil || h.interval <= 0 {
		return nil
	}
	if err := h.Err(); err != nil {
		return err
	}
	if err := h.queue.Heartbeat(ctx, h.job); err != nil {
		leaseErr := executionLeaseLostError(err)
		h.failOnce.Do(func() {
			if h.dispatcher != nil && h.dispatcher.observeHeartbeatFailure != nil {
				h.dispatcher.observeHeartbeatFailure(ctx, HeartbeatFailure{
					JobID:             strings.TrimSpace(h.job.ID),
					Lane:              h.lane,
					Attempt:           h.job.Attempt,
					HeartbeatInterval: h.interval,
					Err:               err,
				})
			}
			h.errMu.Lock()
			h.err = leaseErr
			h.errMu.Unlock()
			if h.dispatcher != nil && h.dispatcher.metrics != nil {
				h.dispatcher.metrics.RecordHeartbeatFailed(h.lane)
			}
			if h.cancelAttempt != nil {
				h.cancelAttempt(leaseErr)
			}
		})
		return leaseErr
	}
	return nil
}

func (h *leaseHeartbeat) Stop() {
	if h == nil {
		return
	}
	h.StopNoWait()
	<-h.done
}

func (h *leaseHeartbeat) StopNoWait() {
	if h == nil {
		return
	}
	h.stopOnce.Do(func() {
		if h.cancel != nil {
			h.cancel()
		}
	})
}

func (h *leaseHeartbeat) Err() error {
	if h == nil {
		return nil
	}
	h.errMu.RLock()
	defer h.errMu.RUnlock()
	return h.err
}
