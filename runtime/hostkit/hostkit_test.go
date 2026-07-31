package hostkit

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	agentx "github.com/wsnacj/agentx-go"
	"github.com/wsnacj/agentx-go/runtime/execution"
	"github.com/wsnacj/agentx-go/runtime/toolloop"
)

var errHostKitRound = errors.New("host kit round failed")

type factoryFunc struct {
	build    func(context.Context, execution.Request) (RunConfig, error)
	shutdown func(context.Context) error
	classify func(error) agentx.ErrorCode
}

func (factory *factoryFunc) BuildRun(ctx context.Context, request execution.Request) (RunConfig, error) {
	return factory.build(ctx, request)
}

func (factory *factoryFunc) Shutdown(ctx context.Context) error {
	if factory.shutdown == nil {
		return nil
	}
	return factory.shutdown(ctx)
}

func (factory *factoryFunc) ClassifyError(err error) agentx.ErrorCode {
	if factory.classify == nil {
		return agentx.CodeExecutionFailed
	}
	return factory.classify(err)
}

type stepExecutor struct {
	mu      sync.Mutex
	steps   []toolloop.RoundExecutionResult
	err     error
	inputs  []toolloop.RoundExecutionInput
	context context.Context
}

func (executor *stepExecutor) ExecuteRound(ctx context.Context, input toolloop.RoundExecutionInput) (toolloop.RoundExecutionResult, error) {
	executor.mu.Lock()
	defer executor.mu.Unlock()
	executor.context = ctx
	executor.inputs = append(executor.inputs, input)
	if executor.err != nil {
		return toolloop.RoundExecutionResult{}, executor.err
	}
	index := len(executor.inputs) - 1
	if index >= len(executor.steps) {
		return toolloop.RoundExecutionResult{Kind: toolloop.OutcomeCompleted}, nil
	}
	return executor.steps[index], nil
}

func TestNewRunsPortableAssemblyThroughRootClient(t *testing.T) {
	ctxKey := struct{ name string }{"host-kit-context"}
	ctx := context.WithValue(context.Background(), ctxKey, "same")
	executor := &stepExecutor{steps: []toolloop.RoundExecutionResult{
		{
			Kind:  toolloop.OutcomeContinue,
			Reply: "working",
			Continuation: &toolloop.RoundContinuation{
				NextChunks: []string{"tool result"},
			},
		},
		{Kind: toolloop.OutcomeCompleted, Reply: "done"},
	}}
	factory := &factoryFunc{build: func(buildCtx context.Context, request execution.Request) (RunConfig, error) {
		if buildCtx != ctx {
			t.Fatalf("BuildRun context identity changed")
		}
		if request.Input != "inspect" || request.SessionID != "session-1" {
			t.Fatalf("request = %#v", request)
		}
		return RunConfig{
			RunID:     "run-1",
			SessionID: request.SessionID,
			Assembly: toolloop.AssemblyConfig{
				MaxRounds: 3,
				Coordinator: toolloop.CoordinatorConfig{
					Executor: executor,
				},
				Initial: toolloop.RoundState{Chunks: []string{request.Input}},
			},
		}, nil
	}}
	client, err := New(Config{Factory: factory})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	result, err := client.Run(ctx, agentx.RunRequest{Input: "inspect", SessionID: " session-1 "})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.RunID != "run-1" || result.SessionID != "session-1" || result.Status != "completed" || result.Reply != "done" {
		t.Fatalf("result = %#v", result)
	}
	if !reflect.DeepEqual(result.Evidence, []string{"run:run-1", "session:session-1"}) {
		t.Fatalf("evidence = %#v", result.Evidence)
	}
	if len(executor.inputs) != 2 || executor.inputs[0].Round != 1 || executor.inputs[1].Round != 2 {
		t.Fatalf("inputs = %#v", executor.inputs)
	}
	if executor.context.Value(ctxKey) != "same" {
		t.Fatalf("executor context identity changed")
	}
}

func TestExecuteMapsTerminationAndMaxRounds(t *testing.T) {
	t.Run("termination", func(t *testing.T) {
		executor := &stepExecutor{steps: []toolloop.RoundExecutionResult{{
			Kind:  toolloop.OutcomeContinue,
			Reply: "partial",
			Continuation: &toolloop.RoundContinuation{
				Failures: []toolloop.FailureObservation{{Tool: "lookup", Failed: true, ErrorClass: "other"}},
			},
		}}}
		result, err := Execute(context.Background(), RunConfig{
			RunID: "run-terminated",
			Assembly: toolloop.AssemblyConfig{
				MaxRounds: 2,
				Coordinator: toolloop.CoordinatorConfig{
					Executor:    executor,
					FailureFuse: toolloop.NewFailureFuse(toolloop.FailureFuseConfig{Enabled: true, Threshold: 1}),
				},
			},
		})
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if result.Status != "incomplete" || result.Driver.Kind != toolloop.OutcomeTerminated || result.Reply != "partial" {
			t.Fatalf("result = %#v", result)
		}
		if result.Termination == nil || result.Termination.Kind != toolloop.TerminationFailureFuse {
			t.Fatalf("termination = %#v", result.Termination)
		}
	})

	t.Run("max rounds", func(t *testing.T) {
		executor := &stepExecutor{steps: []toolloop.RoundExecutionResult{{
			Kind:         toolloop.OutcomeContinue,
			Reply:        "last reply",
			Continuation: &toolloop.RoundContinuation{},
		}}}
		result, err := Execute(context.Background(), RunConfig{
			Assembly: toolloop.AssemblyConfig{
				MaxRounds:   1,
				Coordinator: toolloop.CoordinatorConfig{Executor: executor},
			},
		})
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if result.Status != "incomplete" || result.Driver.Kind != toolloop.OutcomeMaxRounds || result.Reply != "last reply" {
			t.Fatalf("result = %#v", result)
		}
	})
}

func TestRunPreservesPartialResultAndErrorIdentity(t *testing.T) {
	executor := &stepExecutor{err: errHostKitRound}
	factory := &factoryFunc{
		build: func(context.Context, execution.Request) (RunConfig, error) {
			return RunConfig{
				RunID:     "run-error",
				SessionID: "session-error",
				Assembly: toolloop.AssemblyConfig{
					MaxRounds:   2,
					Coordinator: toolloop.CoordinatorConfig{Executor: executor},
					Initial:     toolloop.RoundState{FinalReply: "partial reply"},
				},
			}, nil
		},
		classify: func(err error) agentx.ErrorCode {
			if errors.Is(err, errHostKitRound) {
				return agentx.CodeExecutionFailed
			}
			return agentx.CodeExecutionFailed
		},
	}
	client, err := New(Config{Factory: factory})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	result, err := client.Run(context.Background(), agentx.RunRequest{Input: "fail"})
	if !errors.Is(err, errHostKitRound) {
		t.Fatalf("Run() error = %v", err)
	}
	var typed *agentx.Error
	if !errors.As(err, &typed) || typed.Code != agentx.CodeExecutionFailed {
		t.Fatalf("typed error = %#v", typed)
	}
	if result.RunID != "run-error" || result.SessionID != "session-error" || result.Reply != "partial reply" || result.Status != "failed" {
		t.Fatalf("result = %#v", result)
	}
}

func TestRunCancellationAndDeadlineRemainTyped(t *testing.T) {
	for _, test := range []struct {
		name string
		ctx  func() (context.Context, context.CancelFunc)
		code agentx.ErrorCode
	}{
		{
			name: "canceled",
			ctx: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx, func() {}
			},
			code: agentx.CodeCanceled,
		},
		{
			name: "deadline",
			ctx: func() (context.Context, context.CancelFunc) {
				return context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
			},
			code: agentx.CodeDeadlineExceeded,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			factory := &factoryFunc{build: func(ctx context.Context, _ execution.Request) (RunConfig, error) {
				return RunConfig{}, ctx.Err()
			}}
			client, err := New(Config{Factory: factory})
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}
			ctx, cancel := test.ctx()
			defer cancel()
			_, err = client.Run(ctx, agentx.RunRequest{Input: "cancel"})
			var typed *agentx.Error
			if !errors.As(err, &typed) || typed.Code != test.code {
				t.Fatalf("error = %#v", err)
			}
		})
	}
}

func TestShutdownIsDelegatedAndClosedClientRejectsRun(t *testing.T) {
	var shutdownCalls int
	factory := &factoryFunc{
		build: func(context.Context, execution.Request) (RunConfig, error) {
			return RunConfig{}, errors.New("unexpected run")
		},
		shutdown: func(ctx context.Context) error {
			if ctx == nil || ctx.Err() != nil {
				return errors.New("invalid shutdown context")
			}
			shutdownCalls++
			return nil
		},
	}
	client, err := New(Config{Factory: factory})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if err := client.Shutdown(context.Background()); err != nil {
		t.Fatalf("first Shutdown() error = %v", err)
	}
	if err := client.Shutdown(context.Background()); err != nil {
		t.Fatalf("second Shutdown() error = %v", err)
	}
	if shutdownCalls != 2 {
		t.Fatalf("shutdown calls = %d", shutdownCalls)
	}
	_, err = client.Run(context.Background(), agentx.RunRequest{Input: "closed"})
	var typed *agentx.Error
	if !errors.As(err, &typed) || typed.Code != agentx.CodeClientClosed {
		t.Fatalf("closed Run() error = %#v", err)
	}
}

func TestValidationFailsBeforeOwnershipTransfer(t *testing.T) {
	if _, err := New(Config{}); err == nil || err.Error() != "agentx host kit: factory is required" {
		t.Fatalf("New() error = %v", err)
	}
	factory := &factoryFunc{build: func(context.Context, execution.Request) (RunConfig, error) {
		return RunConfig{}, nil
	}}
	if _, err := New(Config{Factory: factory, Profile: agentx.ExecutionProfile{Driver: "workflow"}}); err == nil {
		t.Fatal("New() accepted unsupported profile")
	}
	if _, err := Execute(nil, RunConfig{}); err == nil || err.Error() != "agentx host kit: nil run context" {
		t.Fatalf("Execute(nil) error = %v", err)
	}
}
