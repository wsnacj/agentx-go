package orchestration

import (
	"context"
	"errors"
	"reflect"
	"testing"

	workflow "github.com/wsnacj/agentx-go/runtime/workflow"
	workflowjournal "github.com/wsnacj/agentx-go/runtime/workflow/journal"
	workflownodeexec "github.com/wsnacj/agentx-go/runtime/workflow/nodeexec"
)

func TestRunPreservesStateAndDurableOrder(t *testing.T) {
	port := newCapturePort()
	result, err := Run(context.Background(), Plan{
		WorkflowID: "workflow-1",
		Version:    "v1",
		EntryNode:  "collect",
		NodeIDs:    []string{"collect"},
		Nodes: map[string]PlannedNode{
			"collect": {
				Spec: workflow.NodeSpec{
					ID:      "collect",
					Kind:    workflow.NodeTool,
					Outputs: []workflow.BindingSpec{{From: "result.value", To: "state.value"}},
				},
				Kind: workflow.NodeTool,
				Call: workflownodeexec.Call{Name: "collect", Arguments: `{}`},
			},
		},
		StateSchema: []workflow.StateSlotSpec{{Name: "value", Required: true}},
	}, Inputs{}, testDependencies(port, staticNodeExecution{
		outcome: workflownodeexec.Outcome{Output: `{"value":"ready"}`},
	}))
	if err != nil {
		t.Fatalf("Run(): %v", err)
	}
	if result.FinalStatus != "completed" ||
		result.FinalNode != "collect" ||
		result.State["value"] != "ready" ||
		result.NodeOutput["collect"] != `{"value":"ready"}` {
		t.Fatalf("result = %#v", result)
	}
	want := []string{
		"load",
		"create",
		"event:workflow.start",
		"event:workflow.node.state.input",
		"node:running",
		"event:workflow.node.start",
		"event:workflow.node.state.output",
		"node:completed",
		"event:workflow.node.finish",
		"load",
		"update",
		"event:workflow.finish",
	}
	if !reflect.DeepEqual(port.operations, want) {
		t.Fatalf("operations = %#v, want %#v", port.operations, want)
	}
}

func TestRunProjectsDurableErrorAndPreservesCause(t *testing.T) {
	sentinel := errors.New("private dependency failure")
	port := newCapturePort()
	result, err := Run(context.Background(), Plan{
		WorkflowID: "workflow-1",
		EntryNode:  "step",
		NodeIDs:    []string{"step"},
		Nodes: map[string]PlannedNode{
			"step": {
				Spec: workflow.NodeSpec{ID: "step", Kind: workflow.NodeTool},
				Kind: workflow.NodeTool,
				Call: workflownodeexec.Call{Name: "step", Arguments: `{}`},
			},
		},
	}, Inputs{}, testDependencies(port, staticNodeExecution{err: sentinel}))
	if !errors.Is(err, sentinel) {
		t.Fatalf("error = %v, want sentinel identity", err)
	}
	if result.FinalStatus != "failed" ||
		result.StopReason != "class=tool code=execution_error" ||
		len(result.NodeResults) != 1 ||
		result.NodeResults[0].Error != "class=tool code=execution_error" {
		t.Fatalf("result = %#v", result)
	}
	run := port.runs[result.RunID]
	if run.Status != "failed" || run.Summary != "class=tool code=execution_error" {
		t.Fatalf("durable run = %#v", run)
	}
}

func TestRunPassesCancellationCauseThroughNodeExecution(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	sentinel := errors.New("cancelled by consumer")
	cancel(sentinel)
	result, err := Run(ctx, Plan{
		WorkflowID: "workflow-1",
		EntryNode:  "step",
		NodeIDs:    []string{"step"},
		Nodes: map[string]PlannedNode{
			"step": {
				Spec: workflow.NodeSpec{ID: "step", Kind: workflow.NodeTool},
				Kind: workflow.NodeTool,
				Call: workflownodeexec.Call{Name: "step", Arguments: `{}`},
			},
		},
	}, Inputs{}, testDependencies(newCapturePort(), contextNodeExecution{}))
	if !errors.Is(err, sentinel) {
		t.Fatalf("error = %v, want cancellation cause", err)
	}
	if result.FinalStatus != "failed" {
		t.Fatalf("result = %#v", result)
	}
}

func TestRunFailsClosedForMissingCapabilities(t *testing.T) {
	_, err := Run(context.Background(), Plan{}, Inputs{}, Dependencies{})
	if err == nil || err.Error() != "workflow orchestration: node execution is required" {
		t.Fatalf("error = %v", err)
	}
	_, err = Run(context.Background(), Plan{}, Inputs{}, Dependencies{
		NodeExecution: staticNodeExecution{},
	})
	if err == nil || err.Error() != "workflow orchestration: clock is required" {
		t.Fatalf("error = %v", err)
	}
}

type staticNodeExecution struct {
	outcome workflownodeexec.Outcome
	err     error
}

type contextNodeExecution struct{}

func (contextNodeExecution) Execute(ctx context.Context, _ workflownodeexec.Request) (workflownodeexec.Outcome, error) {
	return workflownodeexec.Outcome{}, context.Cause(ctx)
}

func (e staticNodeExecution) Execute(context.Context, workflownodeexec.Request) (workflownodeexec.Outcome, error) {
	return e.outcome, e.err
}

func testDependencies(port *capturePort, execution NodeExecution) Dependencies {
	eventID := 0
	now := int64(100)
	return Dependencies{
		Journal: workflowjournal.New(workflowjournal.Dependencies{
			Port:         port,
			NewRunID:     func() string { return "run-1" },
			NewEventID:   func() string { eventID++; return "event" },
			NowUnixMilli: func() int64 { now++; return now },
		}),
		NodeExecution:      execution,
		NewNodeExecutionID: func() string { return "nodeexec-1" },
		NowUnixMilli:       func() int64 { now++; return now },
		ProjectError:       func(error) string { return "class=tool code=execution_error" },
	}
}

type capturePort struct {
	runs       map[string]workflowjournal.Run
	operations []string
}

func newCapturePort() *capturePort {
	return &capturePort{runs: map[string]workflowjournal.Run{}}
}

func (p *capturePort) LoadRun(_ context.Context, runID string) (workflowjournal.Run, bool, error) {
	p.operations = append(p.operations, "load")
	run, ok := p.runs[runID]
	return run, ok, nil
}

func (p *capturePort) CreateRun(_ context.Context, run workflowjournal.Run) error {
	p.operations = append(p.operations, "create")
	p.runs[run.RunID] = run
	return nil
}

func (p *capturePort) UpdateRun(_ context.Context, run workflowjournal.Run) error {
	p.operations = append(p.operations, "update")
	p.runs[run.RunID] = run
	return nil
}

func (p *capturePort) UpsertNodeExecution(_ context.Context, node workflowjournal.NodeExecution) error {
	p.operations = append(p.operations, "node:"+node.Status)
	return nil
}

func (p *capturePort) AppendEvent(_ context.Context, event workflowjournal.Event) error {
	p.operations = append(p.operations, "event:"+event.Name)
	return nil
}
