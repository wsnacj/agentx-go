package hostkit

import (
	"context"
	"errors"
	"strings"
	"testing"

	workflow "github.com/wsnacj/agentx-go/runtime/workflow"
)

type acceptingValidator struct {
	err error
}

func (v acceptingValidator) ValidateSpec(workflow.Spec) error     { return v.err }
func (v acceptingValidator) ValidateNode(workflow.NodeSpec) error { return v.err }

type fixedMapper struct {
	err error
}

func (m fixedMapper) MapNode(workflow.NodeSpec, workflow.ExecutionMode) (MappedCall, error) {
	return MappedCall{Name: "echo", Arguments: map[string]any{"value": "ok"}}, m.err
}

type fixedBasicExecutor struct {
	output     string
	err        error
	contextErr error
	calls      int
}

func (e *fixedBasicExecutor) Execute(ctx context.Context, call Call) (string, error) {
	e.calls++
	e.contextErr = ctx.Err()
	if call.Name != "echo" || call.Arguments != `{"value":"ok"}` {
		return "", errors.New("unexpected call")
	}
	return e.output, e.err
}

func validConfig() Config {
	return Config{
		Validator:          acceptingValidator{},
		Mapper:             fixedMapper{},
		BasicExecutor:      &fixedBasicExecutor{output: "ok"},
		NewRunID:           func() string { return "run-1" },
		NewNodeExecutionID: func() string { return "nodeexec-1" },
		NowUnixMilli:       func() int64 { return 1 },
	}
}

func validSpec() workflow.Spec {
	return workflow.Spec{
		ID:        "workflow-1",
		EntryNode: "step",
		Nodes: []workflow.NodeSpec{{
			ID:   "step",
			Kind: workflow.NodeTool,
		}},
	}
}

func TestNewFailsClosedInStableOrder(t *testing.T) {
	tests := []struct {
		name  string
		clear func(*Config)
		want  string
	}{
		{name: "validator", clear: func(c *Config) { c.Validator = nil }, want: "validator is required"},
		{name: "mapper", clear: func(c *Config) { c.Mapper = nil }, want: "mapper is required"},
		{name: "executor", clear: func(c *Config) { c.BasicExecutor = nil }, want: "executor is required"},
		{name: "run id", clear: func(c *Config) { c.NewRunID = nil }, want: "run id generator is required"},
		{
			name: "event id",
			clear: func(c *Config) {
				c.JournalPort = inertJournalPort{}
				c.NewEventID = nil
			},
			want: "event id generator is required when journal port is configured",
		},
		{name: "node execution id", clear: func(c *Config) { c.NewNodeExecutionID = nil }, want: "node execution id generator is required"},
		{name: "clock", clear: func(c *Config) { c.NowUnixMilli = nil }, want: "clock is required"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := validConfig()
			test.clear(&config)
			runtime, err := New(config)
			if runtime != nil || err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("New() = %#v, %v; want %q", runtime, err, test.want)
			}
		})
	}
}

func TestRunUsesCanonicalWorkflowAndPreservesErrorIdentity(t *testing.T) {
	executor := &fixedBasicExecutor{output: "portable-ok"}
	config := validConfig()
	config.BasicExecutor = executor
	runtime, err := New(config)
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	result, err := runtime.Run(context.Background(), validSpec(), Inputs{})
	if err != nil {
		t.Fatalf("Run(): %v", err)
	}
	if result.LoweringPlan.SpecID != "workflow-1" ||
		result.Execution.RunID != "run-1" ||
		result.Execution.FinalStatus != "completed" ||
		result.Execution.NodeOutput["step"] != "portable-ok" ||
		executor.calls != 1 {
		t.Fatalf("result=%#v calls=%d", result, executor.calls)
	}

	sentinel := errors.New("raw validation error")
	config = validConfig()
	config.Validator = acceptingValidator{err: sentinel}
	runtime, err = New(config)
	if err != nil {
		t.Fatalf("New(validation): %v", err)
	}
	result, err = runtime.Run(context.Background(), validSpec(), Inputs{})
	if !errors.Is(err, sentinel) || result.Execution.RunID != "" {
		t.Fatalf("Run(validation) = %#v, %v", result, err)
	}
}

func TestRunPassesCancellationAndPartialExecutionThrough(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	executor := &fixedBasicExecutor{output: "dependency-decides", err: context.Canceled}
	config := validConfig()
	config.BasicExecutor = executor
	runtime, err := New(config)
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	result, err := runtime.Run(ctx, validSpec(), Inputs{})
	if !errors.Is(err, context.Canceled) ||
		!errors.Is(executor.contextErr, context.Canceled) ||
		result.LoweringPlan.SpecID != "workflow-1" ||
		result.Execution.RunID != "run-1" ||
		len(result.Execution.NodeResults) != 1 {
		t.Fatalf("Run(cancelled) = %#v, %v; context=%v", result, err, executor.contextErr)
	}
}

func TestNilRuntimeFailsClosed(t *testing.T) {
	var runtime *Runtime
	if _, err := runtime.Run(context.Background(), validSpec(), Inputs{}); err == nil ||
		!strings.Contains(err.Error(), "runtime is required") {
		t.Fatalf("Run(nil) error = %v", err)
	}
}

type inertJournalPort struct{}

func (inertJournalPort) LoadRun(context.Context, string) (JournalRun, bool, error) {
	return JournalRun{}, false, nil
}
func (inertJournalPort) CreateRun(context.Context, JournalRun) error { return nil }
func (inertJournalPort) UpdateRun(context.Context, JournalRun) error { return nil }
func (inertJournalPort) UpsertNodeExecution(context.Context, JournalNodeExecution) error {
	return nil
}
func (inertJournalPort) AppendEvent(context.Context, JournalEvent) error { return nil }
