package composition

import (
	"context"
	"errors"
	"strings"
	"testing"

	workflow "github.com/wsnacj/agentx-go/runtime/workflow"
	workflowjournal "github.com/wsnacj/agentx-go/runtime/workflow/journal"
	workflowlowering "github.com/wsnacj/agentx-go/runtime/workflow/lowering"
	workflownodeexec "github.com/wsnacj/agentx-go/runtime/workflow/nodeexec"
	workfloworchestration "github.com/wsnacj/agentx-go/runtime/workflow/orchestration"
)

type validator struct {
	err error
}

func (v validator) ValidateSpec(workflow.Spec) error     { return v.err }
func (v validator) ValidateNode(workflow.NodeSpec) error { return v.err }

type mapper struct{}

func (mapper) MapNode(workflow.NodeSpec, workflow.ExecutionMode) (workflowlowering.MappedCall, error) {
	return workflowlowering.MappedCall{Name: "echo"}, nil
}

type nodeExecution struct{}

func (nodeExecution) Execute(context.Context, workflownodeexec.Request) (workflownodeexec.Outcome, error) {
	return workflownodeexec.Outcome{Output: "ok"}, nil
}

func validDependencies() Dependencies {
	var next int
	return Dependencies{
		Lowering: workflowlowering.Dependencies{
			Validator: validator{},
			Mapper:    mapper{},
		},
		Orchestration: workfloworchestration.Dependencies{
			Journal: workflowjournal.New(workflowjournal.Dependencies{
				NewRunID:     func() string { return "run" },
				NewEventID:   func() string { return "event" },
				NowUnixMilli: func() int64 { return int64(next) },
			}),
			NodeExecution: nodeExecution{},
			NewNodeExecutionID: func() string {
				next++
				return "nodeexec"
			},
			NowUnixMilli: func() int64 { return int64(next) },
		},
	}
}

func validSpec() workflow.Spec {
	return workflow.Spec{
		ID:        "spec-workflow",
		EntryNode: "step",
		Nodes: []workflow.NodeSpec{{
			ID:   "step",
			Kind: workflow.NodeTool,
		}},
	}
}

func TestNewFailsClosedForRequiredDependencies(t *testing.T) {
	tests := map[string]struct {
		clear func(*Dependencies)
		want  string
	}{
		"validator": {
			clear: func(deps *Dependencies) { deps.Lowering.Validator = nil },
			want:  "lowering validator is required",
		},
		"mapper": {
			clear: func(deps *Dependencies) { deps.Lowering.Mapper = nil },
			want:  "lowering mapper is required",
		},
		"journal": {
			clear: func(deps *Dependencies) { deps.Orchestration.Journal = nil },
			want:  "journal is required",
		},
		"node execution": {
			clear: func(deps *Dependencies) { deps.Orchestration.NodeExecution = nil },
			want:  "node execution is required",
		},
		"node execution id": {
			clear: func(deps *Dependencies) { deps.Orchestration.NewNodeExecutionID = nil },
			want:  "node execution id generator is required",
		},
		"clock": {
			clear: func(deps *Dependencies) { deps.Orchestration.NowUnixMilli = nil },
			want:  "clock is required",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			deps := validDependencies()
			test.clear(&deps)
			runtime, err := New(deps)
			if runtime != nil || err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("New() = %#v, %v, want %q", runtime, err, test.want)
			}
		})
	}
}

func TestRunPreservesRawWorkflowIDOverrideAndFallback(t *testing.T) {
	for name, inputs := range map[string]Inputs{
		"fallback": {},
		"override": {WorkflowID: " raw-workflow "},
	} {
		t.Run(name, func(t *testing.T) {
			deps := validDependencies()
			var observed string
			deps.Orchestration.Journal = workflowjournal.New(workflowjournal.Dependencies{
				NewRunID:   func() string { return "run" },
				NewEventID: func() string { return "event" },
				NowUnixMilli: func() int64 {
					return 1
				},
				Port: journalPortFunc(func(run workflowjournal.Run) {
					observed = run.WorkflowID
				}),
			})
			runtime, err := New(deps)
			if err != nil {
				t.Fatalf("New(): %v", err)
			}
			result, err := runtime.Run(context.Background(), validSpec(), inputs)
			if err != nil {
				t.Fatalf("Run(): %v", err)
			}
			want := "spec-workflow"
			if inputs.WorkflowID != "" {
				want = inputs.WorkflowID
			}
			if observed != want || result.LoweringPlan.SpecID != "spec-workflow" ||
				result.Execution.FinalStatus != "completed" {
				t.Fatalf("observed=%q result=%#v, want workflow=%q", observed, result, want)
			}
		})
	}
}

func TestRunStopsAtLoweringAndPreservesCause(t *testing.T) {
	sentinel := errors.New("validation sentinel")
	deps := validDependencies()
	deps.Lowering.Validator = validator{err: sentinel}
	runtime, err := New(deps)
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	result, err := runtime.Run(context.Background(), validSpec(), Inputs{})
	if !errors.Is(err, sentinel) ||
		result.LoweringPlan.SpecID != "" ||
		result.Execution.RunID != "" {
		t.Fatalf("Run() = %#v, %v", result, err)
	}
}

func TestRunPreservesPlanAndPartialExecutionOnOrchestrationFailure(t *testing.T) {
	sentinel := errors.New("start event sentinel")
	deps := validDependencies()
	deps.Orchestration.Journal = workflowjournal.New(workflowjournal.Dependencies{
		NewRunID:     func() string { return "run-partial" },
		NewEventID:   func() string { return "event-partial" },
		NowUnixMilli: func() int64 { return 1 },
		Port:         failingJournalPort{err: sentinel},
	})
	runtime, err := New(deps)
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	result, err := runtime.Run(context.Background(), validSpec(), Inputs{})
	if !errors.Is(err, sentinel) {
		t.Fatalf("Run() error = %v, want sentinel", err)
	}
	if result.LoweringPlan.SpecID != "spec-workflow" ||
		result.Execution.RunID != "run-partial" ||
		result.Execution.NodeOutput == nil ||
		result.Execution.State == nil {
		t.Fatalf("partial result = %#v", result)
	}
}

type journalPortFunc func(workflowjournal.Run)

func (fn journalPortFunc) LoadRun(context.Context, string) (workflowjournal.Run, bool, error) {
	return workflowjournal.Run{}, false, nil
}
func (fn journalPortFunc) CreateRun(_ context.Context, run workflowjournal.Run) error {
	fn(run)
	return nil
}
func (journalPortFunc) UpdateRun(context.Context, workflowjournal.Run) error {
	return nil
}
func (journalPortFunc) UpsertNodeExecution(context.Context, workflowjournal.NodeExecution) error {
	return nil
}
func (journalPortFunc) AppendEvent(context.Context, workflowjournal.Event) error {
	return nil
}

type failingJournalPort struct {
	err error
}

func (failingJournalPort) LoadRun(context.Context, string) (workflowjournal.Run, bool, error) {
	return workflowjournal.Run{}, false, nil
}
func (failingJournalPort) CreateRun(context.Context, workflowjournal.Run) error {
	return nil
}
func (failingJournalPort) UpdateRun(context.Context, workflowjournal.Run) error {
	return nil
}
func (failingJournalPort) UpsertNodeExecution(context.Context, workflowjournal.NodeExecution) error {
	return nil
}
func (p failingJournalPort) AppendEvent(context.Context, workflowjournal.Event) error {
	return p.err
}
