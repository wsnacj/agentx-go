// Package hostkit provides the portable coordination layer for one
// host-owned child worker lifecycle.
//
// It invokes a caller-supplied worker exactly once, records and reads back the
// result through a caller-supplied store, and can project the verified child
// result into the parent Objective runtime. It does not own a Runner, process,
// queue, scheduler, credential, authorization policy, or concrete backend.
package hostkit

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"time"

	controlcontract "github.com/wsnacj/agentx-go/runtime/session"
)

var (
	// ErrInvalidConfig identifies a missing or typed-nil Host Kit dependency.
	ErrInvalidConfig = errors.New("session hostkit: invalid config")
	// ErrClosed identifies Run calls made after Shutdown.
	ErrClosed = errors.New("session hostkit: closed")
)

// Config supplies all side-effecting child worker capabilities.
type Config struct {
	Worker      WorkerRuntime
	Store       StateStore
	BackendRef  controlcontract.DisplaySafeRef
	BackendKind string
	Durable     bool
	Now         func() time.Time
}

// ConfigError reports which explicit Host dependency is missing.
type ConfigError struct {
	Field string
}

func (e *ConfigError) Error() string {
	if e == nil || e.Field == "" {
		return ErrInvalidConfig.Error()
	}
	return ErrInvalidConfig.Error() + ": " + e.Field + " is required"
}

// Unwrap makes ConfigError compatible with errors.Is(err, ErrInvalidConfig).
func (*ConfigError) Unwrap() error { return ErrInvalidConfig }

// Runtime is safe for concurrent Run calls when the supplied WorkerRuntime and
// StateStore implementations are safe for the same concurrency.
type Runtime struct {
	backend Backend

	mu     sync.RWMutex
	closed bool
}

// RunRequest describes one child worker lifecycle. Closure is optional; when
// present the verified readback is also projected into the parent Objective.
type RunRequest struct {
	BackendInput BackendInput
	Closure      *ObjectiveRuntimeClosureInput
	Boundaries   []controlcontract.Boundary
}

// RunResult is the normalized Host Kit result. Backend is always populated;
// Closure is populated only when RunRequest.Closure is present.
type RunResult struct {
	Completed bool                                 `json:"completed"`
	Status    string                               `json:"status,omitempty"`
	Backend   BackendReport                        `json:"backend"`
	Closure   ObjectiveRuntimeClosureProfileReport `json:"closure,omitempty"`
}

// New validates the required Host ports and constructs an inert Runtime. It
// does not start a process, worker, queue, or goroutine.
func New(config Config) (*Runtime, error) {
	if interfaceNil(config.Worker) {
		return nil, &ConfigError{Field: "Worker"}
	}
	if interfaceNil(config.Store) {
		return nil, &ConfigError{Field: "Store"}
	}
	return &Runtime{backend: Backend{
		Runtime:     config.Worker,
		Store:       config.Store,
		BackendRef:  config.BackendRef,
		BackendKind: config.BackendKind,
		Durable:     config.Durable,
		Now:         config.Now,
	}}, nil
}

// Run invokes one explicitly enabled child worker lifecycle. It never accepts
// the child output as parent fact without the canonical verification path.
func (r *Runtime) Run(ctx context.Context, request RunRequest) (RunResult, error) {
	if r == nil {
		return RunResult{}, ErrClosed
	}
	r.mu.RLock()
	closed := r.closed
	r.mu.RUnlock()
	if closed {
		return RunResult{}, ErrClosed
	}
	if ctx == nil {
		ctx = context.Background()
	}

	if request.Closure != nil {
		profile, err := RunObjectiveRuntimeClosureProfile(ctx, ObjectiveRuntimeClosureProfile{
			Enabled:      request.BackendInput.Enabled,
			Backend:      r.backend,
			BackendInput: request.BackendInput,
			ClosureInput: *request.Closure,
			Boundaries:   append([]controlcontract.Boundary(nil), request.Boundaries...),
		})
		return RunResult{
			Completed: profile.Ready && profile.RuntimeLoopHostPersistReady,
			Status:    profile.Status,
			Backend:   profile.Backend,
			Closure:   profile,
		}, err
	}

	backend, err := r.backend.RunDelegationWorkerRuntime(ctx, request.BackendInput)
	return RunResult{
		Completed: backend.WorkerResultReadbackReady,
		Status:    string(backend.Status),
		Backend:   backend,
	}, err
}

// Shutdown prevents future Run calls. It is bounded and idempotent because the
// Host Kit owns no background worker; the Host remains responsible for its
// injected worker and store lifecycles.
func (r *Runtime) Shutdown(ctx context.Context) error {
	if r == nil {
		return nil
	}
	if ctx != nil {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}
	r.mu.Lock()
	r.closed = true
	r.mu.Unlock()
	return nil
}

func interfaceNil(value any) bool {
	if value == nil {
		return true
	}
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return rv.IsNil()
	default:
		return false
	}
}
