package scheduler

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

type Lane string

const (
	LaneMain       Lane = "main"
	LaneSubtask    Lane = "subtask"
	LaneBackground Lane = "background"
)

type Job struct {
	ID               string
	Lane             Lane
	SessionID        string
	Payload          string
	Attempt          int
	IdempotencyKey   string
	JobKind          string
	CoalescingKey    string
	ProtectedJobName string
	TrustedCaller    bool
	CreatedAt        time.Time
}

// ResultOutcome is the machine-readable terminal outcome used by scheduler
// control flow. Result.Error remains display-only diagnostic text.
type ResultOutcome string

const (
	// ResultOutcomeCompleted indicates successful terminal completion.
	ResultOutcomeCompleted ResultOutcome = "completed"
	// ResultOutcomeFailed indicates a retryable or terminal execution failure.
	ResultOutcomeFailed ResultOutcome = "failed"
	// ResultOutcomeCanceled indicates an explicit cancellation.
	ResultOutcomeCanceled ResultOutcome = "canceled"
	// ResultOutcomeDeadLetter indicates retry exhaustion.
	ResultOutcomeDeadLetter ResultOutcome = "dead_letter"
)

type Result struct {
	JobID      string
	Lane       Lane
	Outcome    ResultOutcome
	Status     string
	Succeeded  bool
	Attempt    int
	FinishedAt time.Time
	Error      string
}

func resultOutcome(result Result, fallback ResultOutcome) ResultOutcome {
	if outcome := normalizeResultOutcome(result.Outcome); outcome != "" {
		return outcome
	}
	if outcome := normalizeResultOutcome(ResultOutcome(result.Status)); outcome != "" {
		return outcome
	}
	return fallback
}

func normalizeResultOutcome(outcome ResultOutcome) ResultOutcome {
	switch ResultOutcome(strings.ToLower(strings.TrimSpace(string(outcome)))) {
	case ResultOutcomeCompleted:
		return ResultOutcomeCompleted
	case ResultOutcomeFailed:
		return ResultOutcomeFailed
	case ResultOutcomeCanceled, "cancelled":
		return ResultOutcomeCanceled
	case ResultOutcomeDeadLetter:
		return ResultOutcomeDeadLetter
	default:
		return ""
	}
}

type Handler func(ctx context.Context, job Job) error

// ExecutionAttempt identifies the scheduler lease attempt currently executing.
// Side-effect runtimes can read and validate it before starting new external work.
type ExecutionAttempt struct {
	JobID          string
	Lane           Lane
	Attempt        int
	IdempotencyKey string
	LeaseOwner     string
}

type executionAttemptContext struct {
	attempt  ExecutionAttempt
	validate func(context.Context) error
}

type executionAttemptContextKey struct{}

var (
	ErrQueueEmpty              = errors.New("scheduler: queue empty")
	ErrQueueLimit              = errors.New("scheduler: queue limit reached")
	ErrQueueUnavailable        = errors.New("scheduler: queue unavailable")
	ErrUnknownJob              = errors.New("scheduler: unknown job")
	ErrInvalidLane             = errors.New("scheduler: invalid lane")
	ErrInvalidHandler          = errors.New("scheduler: invalid handler")
	ErrLeaseLost               = errors.New("scheduler: execution lease lost")
	ErrCancellationUnsupported = errors.New("scheduler: queue cancellation unsupported")
)

// ExecutionAttemptFromContext returns the scheduler attempt attached by a
// Dispatcher. Contexts outside scheduler execution do not carry an attempt.
func ExecutionAttemptFromContext(ctx context.Context) (ExecutionAttempt, bool) {
	if ctx == nil {
		return ExecutionAttempt{}, false
	}
	state, ok := ctx.Value(executionAttemptContextKey{}).(*executionAttemptContext)
	if !ok || state == nil || strings.TrimSpace(state.attempt.JobID) == "" {
		return ExecutionAttempt{}, false
	}
	return state.attempt, true
}

// ValidateExecutionAttempt checks that a scheduler attempt still owns its
// lease. It is a no-op for non-scheduler contexts.
func ValidateExecutionAttempt(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	state, ok := ctx.Value(executionAttemptContextKey{}).(*executionAttemptContext)
	if !ok || state == nil {
		return nil
	}
	if cause := context.Cause(ctx); cause != nil {
		return cause
	}
	if state.validate == nil {
		return nil
	}
	return state.validate(ctx)
}

func withExecutionAttempt(ctx context.Context, attempt ExecutionAttempt, validate func(context.Context) error) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, executionAttemptContextKey{}, &executionAttemptContext{
		attempt:  attempt,
		validate: validate,
	})
}

func executionLeaseLostError(err error) error {
	if err == nil {
		return ErrLeaseLost
	}
	return fmt.Errorf("%w: %v", ErrLeaseLost, err)
}

type Queue interface {
	Enqueue(ctx context.Context, job Job) error
	Dequeue(ctx context.Context, lane Lane) (Job, error)
	Ack(ctx context.Context, result Result) error
	Fail(ctx context.Context, result Result) error
	Result(ctx context.Context, jobID string) (Result, bool, error)
	Pending(ctx context.Context, jobID string) (bool, error)
}

// CancelableQueue is the host-control extension for canceling queued or
// currently leased jobs. Worker execution failures continue to use Queue.Fail.
type CancelableQueue interface {
	Cancel(ctx context.Context, result Result) error
}

// CancelJob applies the typed host-cancellation contract without falling back
// to worker failure semantics.
func CancelJob(ctx context.Context, queue Queue, result Result) error {
	if queue == nil {
		return ErrQueueUnavailable
	}
	cancelable, ok := queue.(CancelableQueue)
	if !ok || cancelable == nil {
		return ErrCancellationUnsupported
	}
	result.Outcome = ResultOutcomeCanceled
	result.Status = string(ResultOutcomeCanceled)
	result.Succeeded = false
	return cancelable.Cancel(ctx, result)
}

// KindAwareQueue is an optional Queue extension for workers that must only
// lease jobs of a specific kind from a shared lane.
type KindAwareQueue interface {
	DequeueByKind(ctx context.Context, lane Lane, jobKind string) (Job, error)
}

// RuntimeVisibleQueue is an optional extension used by orchestration helpers
// that need to know whether a queue backend can expose fresh queued/running
// runtime state back to callers. Backends that return false should be treated
// conservatively: a queued/running task record may still represent a live job
// even when Pending/Result cannot currently prove it.
type RuntimeVisibleQueue interface {
	HasRuntimeVisibility() bool
}

// HeartbeatCapableQueue is an optional extension used by distributed backends.
// Dispatcher workers call Heartbeat periodically while a job handler is running.
type HeartbeatCapableQueue interface {
	Heartbeat(ctx context.Context, job Job) error
	HeartbeatInterval() time.Duration
}

// LeaseIdentityProvider is an optional queue extension used to expose a
// display-safe lease owner to the handler execution context.
type LeaseIdentityProvider interface {
	LeaseIdentity(job Job) string
}
