package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	agentxmedia "github.com/wsnacj/agentx-go/runtime/mediaartifact"
	agentxprotocol "github.com/wsnacj/agentx-go/runtime/protocol"
	agentxtelemetry "github.com/wsnacj/agentx-go/runtime/telemetry"
	agentxsafeerror "github.com/wsnacj/agentx-go/runtime/telemetry/safeerror"
	agentxtoolerrors "github.com/wsnacj/agentx-go/runtime/toolerrors"
	agentxworkflow "github.com/wsnacj/agentx-go/runtime/workflow"
	agentxbindingstate "github.com/wsnacj/agentx-go/runtime/workflow/bindingstate"
	agentxjournal "github.com/wsnacj/agentx-go/runtime/workflow/journal"
	agentxnodeexec "github.com/wsnacj/agentx-go/runtime/workflow/nodeexec"
	agentxtransition "github.com/wsnacj/agentx-go/runtime/workflow/transition"
)

func canonicalEventJSON() ([]byte, error) {
	event := agentxprotocol.NormalizeRunEvent(agentxprotocol.RunEvent{
		Envelope: agentxprotocol.Envelope{
			Kind:  " tool.call.completed ",
			RunID: " run-consumer-1 ",
		},
		Status:   " Completed ",
		ToolName: " Browser_Open ",
	})
	if err := agentxprotocol.ValidateRunEvent(event); err != nil {
		return nil, err
	}
	return json.Marshal(event)
}

func canonicalSafeErrorJSON() ([]byte, error) {
	cause := errors.New("private consumer sentinel")
	wrapped := agentxsafeerror.WrapWithIdentity(
		cause,
		"operation failed",
		"consumer-error-1",
	)
	if !errors.Is(wrapped, cause) {
		return nil, fmt.Errorf("safeerror wrapper lost cause")
	}
	projection := agentxsafeerror.Project(
		wrapped,
		" Runtime Error ",
		" UPSTREAM/FAILED ",
	)
	return json.Marshal(projection)
}

func canonicalMediaArtifactJSON() ([]byte, error) {
	hasAudio := false
	return json.Marshal(agentxmedia.Descriptor{
		Source:     "nodes",
		Kind:       "video",
		Path:       ".agentx/nodes/capture.mp4",
		MIMEType:   "video/mp4",
		Format:     "mp4",
		Bytes:      4096,
		DurationMs: 2500,
		FPS:        30,
		HasAudio:   &hasAudio,
	})
}

func canonicalToolArgumentError() (*agentxtoolerrors.ToolArgumentError, error) {
	cause := errors.New("decode: top-level JSON object is required")
	err := agentxtoolerrors.NewInvalidJSONToolArgumentError(" browser ", cause)
	wrapped := fmt.Errorf("consumer wrapper: %w", err)
	if !errors.Is(wrapped, cause) {
		return nil, fmt.Errorf("tool argument error lost cause")
	}
	typed, ok := agentxtoolerrors.AsToolArgumentError(wrapped)
	if !ok {
		return nil, fmt.Errorf("tool argument error lost typed identity")
	}
	return typed, nil
}

func canonicalTelemetryJSON() ([]byte, error) {
	events := agentxtelemetry.ProjectToolEvents(agentxtelemetry.Event{
		Component: "tool",
		Name:      "tool.finish",
		Tool:      " Browser_Open ",
		Status:    "ok",
		Attrs: map[string]any{
			"duration_ms": 42,
		},
	})
	if len(events) != 1 {
		return nil, fmt.Errorf("telemetry projection count = %d, want 1", len(events))
	}
	return json.Marshal(events[0])
}

func canonicalWorkflowJSON() ([]byte, error) {
	return json.Marshal(agentxworkflow.Spec{
		ID:           "consumer-workflow",
		Version:      "1",
		PlanningMode: agentxworkflow.PlanningBounded,
		EntryNode:    "collect",
		Nodes: []agentxworkflow.NodeSpec{{
			ID:            "collect",
			Kind:          agentxworkflow.NodeCollect,
			ExecutionMode: agentxworkflow.ExecInline,
			Inputs: []agentxworkflow.BindingSpec{{
				From: "request.query",
				To:   "query",
			}},
			Outputs: []agentxworkflow.BindingSpec{{
				From: "result",
				To:   "state.report",
			}},
			Retry: agentxworkflow.RetryPolicy{
				MaxAttempts: 2,
				BackoffMs:   []int{100},
			},
			Config: map[string]any{"format": "markdown"},
		}},
		Edges: []agentxworkflow.EdgeSpec{{
			From: "collect",
			To:   "collect",
			On:   "retry",
		}},
		StateSchema: []agentxworkflow.StateSlotSpec{{
			Name:     "report",
			Type:     "string",
			Required: true,
		}},
		ArtifactSchema: []agentxworkflow.ArtifactTypeRef{{
			Type: "report",
		}},
		EvaluatorSchema: []agentxworkflow.EvaluatorRef{{
			Name: "quality",
		}},
	})
}

func canonicalWorkflowBindingState() (map[string]any, error) {
	runtime := agentxbindingstate.New(agentxbindingstate.Inputs{
		SessionInput: map[string]any{"query": "risk"},
	})
	args, err := runtime.MaterializeArguments(
		"collect",
		`{}`,
		[]agentxworkflow.BindingSpec{{
			From: "session.input.query",
			To:   "args.query",
		}},
	)
	if err != nil {
		return nil, err
	}
	if args != `{"query":"risk"}` {
		return nil, fmt.Errorf("workflow binding arguments = %s", args)
	}
	if err := runtime.ApplyNodeOutputs(
		"collect",
		[]agentxworkflow.BindingSpec{{
			From: "result.report",
			To:   "state.report",
		}},
		agentxbindingstate.NewNodeResult("completed", `{"report":"ready"}`, ""),
	); err != nil {
		return nil, err
	}
	if err := runtime.ValidateRequiredSlots([]agentxworkflow.StateSlotSpec{{
		Name:     "report",
		Required: true,
	}}); err != nil {
		return nil, err
	}
	return runtime.State(), nil
}

func canonicalWorkflowTransition() ([]string, error) {
	machine := agentxtransition.New(agentxtransition.Plan{
		EntryNode: "collect",
		NodeIDs:   []string{"collect", "report"},
		Edges: []agentxworkflow.EdgeSpec{{
			From: "collect",
			To:   "report",
		}},
	})
	var visited []string
	for {
		nodeID, err := machine.Enter()
		if err != nil {
			return nil, err
		}
		if nodeID == "" {
			return visited, nil
		}
		visited = append(visited, nodeID)
		next, err := machine.Advance(agentxtransition.TriggerSuccess)
		if err != nil {
			return nil, err
		}
		if next == "" {
			return visited, nil
		}
	}
}

func canonicalWorkflowJournal() ([]string, error) {
	port := &consumerJournalPort{runs: map[string]agentxjournal.Run{}}
	ids := []string{
		"run-consumer-1",
		"workflow-start",
		"node-input",
		"node-start",
		"node-output",
		"node-finish",
		"workflow-finish",
	}
	nextID := func() string {
		value := ids[0]
		ids = ids[1:]
		return value
	}
	journal := agentxjournal.New(agentxjournal.Dependencies{
		Port:         port,
		NewRunID:     nextID,
		NewEventID:   nextID,
		NowUnixMilli: func() int64 { return 100 },
	})
	runID, err := journal.EnsureRun(context.Background(), agentxjournal.EnsureRunRequest{
		WorkflowID:      "consumer-workflow",
		WorkflowVersion: "1",
	})
	if err != nil {
		return nil, err
	}
	if err := journal.AppendRunStart(context.Background(), agentxjournal.StartRunEventRequest{
		RunID:      runID,
		WorkflowID: "consumer-workflow",
		EntryNode:  "collect",
	}); err != nil {
		return nil, err
	}
	node := agentxjournal.NodeExecution{
		NodeExecutionID: "node-consumer-1",
		RunID:           runID,
		NodeID:          "collect",
		Status:          "running",
		Attempt:         1,
		StartedAt:       101,
	}
	inputRef, err := journal.StartNode(context.Background(), agentxjournal.StartNodeRequest{
		Node:  node,
		State: map[string]any{"query": "risk"},
	})
	if err != nil {
		return nil, err
	}
	node.InputStateRef = inputRef
	node.Status = "completed"
	node.FinishedAt = 102
	if _, err := journal.FinishNode(context.Background(), agentxjournal.FinishNodeRequest{
		Node:  node,
		State: map[string]any{"report": "ready"},
	}); err != nil {
		return nil, err
	}
	if err := journal.FinishRun(context.Background(), agentxjournal.FinishRunRequest{
		RunID:           runID,
		WorkflowID:      "consumer-workflow",
		WorkflowVersion: "1",
		Status:          "completed",
		FinishedAt:      102,
	}); err != nil {
		return nil, err
	}
	return port.operations, nil
}

func canonicalWorkflowNodeExecution() (agentxnodeexec.Outcome, []int, error) {
	executor := &consumerNodeExecutor{}
	coordinator := agentxnodeexec.New(agentxnodeexec.Dependencies{
		Basic:   executor,
		Node:    executor,
		Outcome: executor,
	})
	outcome, err := coordinator.Execute(context.Background(), agentxnodeexec.Request{
		NodeExecutionID: "nodeexec-consumer-1",
		NodeID:          "collect",
		Kind:            agentxworkflow.NodeTool,
		Call: agentxnodeexec.Call{
			Name:      "public_source",
			Arguments: `{"query":"risk"}`,
		},
	})
	return outcome, []int{
		executor.basicCalls,
		executor.nodeCalls,
		executor.outcomeCalls,
	}, err
}

type consumerNodeExecutor struct {
	basicCalls   int
	nodeCalls    int
	outcomeCalls int
}

func (e *consumerNodeExecutor) Execute(context.Context, agentxnodeexec.Call) (string, error) {
	e.basicCalls++
	return "basic", nil
}

func (e *consumerNodeExecutor) ExecuteNode(context.Context, agentxnodeexec.Request) (string, error) {
	e.nodeCalls++
	return "node", nil
}

func (e *consumerNodeExecutor) ExecuteNodeWithOutcome(context.Context, agentxnodeexec.Request) (agentxnodeexec.Outcome, error) {
	e.outcomeCalls++
	return agentxnodeexec.Outcome{
		Output:      "ready",
		FinalStatus: "completed",
		ChildNodeExecutions: []agentxnodeexec.ChildNodeExecutionProjection{{
			NodeExecutionID: "child-1",
			Status:          "completed",
		}},
	}, nil
}

type consumerJournalPort struct {
	runs       map[string]agentxjournal.Run
	operations []string
}

func (p *consumerJournalPort) LoadRun(_ context.Context, runID string) (agentxjournal.Run, bool, error) {
	p.operations = append(p.operations, "load:"+runID)
	run, found := p.runs[runID]
	return run, found, nil
}

func (p *consumerJournalPort) CreateRun(_ context.Context, run agentxjournal.Run) error {
	p.operations = append(p.operations, "create:"+run.RunID)
	p.runs[run.RunID] = run
	return nil
}

func (p *consumerJournalPort) UpdateRun(_ context.Context, run agentxjournal.Run) error {
	p.operations = append(p.operations, "update:"+run.RunID)
	p.runs[run.RunID] = run
	return nil
}

func (p *consumerJournalPort) UpsertNodeExecution(_ context.Context, node agentxjournal.NodeExecution) error {
	p.operations = append(p.operations, "node:"+node.Status)
	return nil
}

func (p *consumerJournalPort) AppendEvent(_ context.Context, event agentxjournal.Event) error {
	p.operations = append(p.operations, "event:"+event.Name)
	return nil
}

func main() {
	payload, err := canonicalWorkflowJSON()
	if err != nil {
		panic(err)
	}
	fmt.Println(string(payload))
}
