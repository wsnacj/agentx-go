package journal_test

import (
	"context"
	"testing"

	journal "github.com/wsnacj/agentx-go/runtime/workflow/journal"
)

func TestExternalPackageConsumer(t *testing.T) {
	port := &memoryPort{runs: map[string]journal.Run{}}
	runtime := journal.New(journal.Dependencies{
		Port:         port,
		NewRunID:     func() string { return "run-1" },
		NewEventID:   sequence("event-start", "input-ref", "event-node-start", "output-ref", "event-node-finish", "event-finish"),
		NowUnixMilli: sequenceInt64(100, 101),
	})
	runID, err := runtime.EnsureRun(context.Background(), journal.EnsureRunRequest{
		WorkflowID:      "consumer-workflow",
		WorkflowVersion: "1",
	})
	if err != nil {
		t.Fatalf("EnsureRun(): %v", err)
	}
	if err := runtime.AppendRunStart(context.Background(), journal.StartRunEventRequest{
		RunID:      runID,
		WorkflowID: "consumer-workflow",
		EntryNode:  "collect",
	}); err != nil {
		t.Fatalf("AppendRunStart(): %v", err)
	}
	inputRef, err := runtime.StartNode(context.Background(), journal.StartNodeRequest{
		Node: journal.NodeExecution{
			NodeExecutionID: "node-1",
			RunID:           runID,
			NodeID:          "collect",
			Status:          "running",
			Attempt:         1,
			StartedAt:       102,
		},
		State: map[string]any{"query": "risk"},
	})
	if err != nil {
		t.Fatalf("StartNode(): %v", err)
	}
	if _, err := runtime.FinishNode(context.Background(), journal.FinishNodeRequest{
		Node: journal.NodeExecution{
			NodeExecutionID: "node-1",
			RunID:           runID,
			NodeID:          "collect",
			Status:          "completed",
			Attempt:         1,
			InputStateRef:   inputRef,
			StartedAt:       102,
			FinishedAt:      103,
		},
		State: map[string]any{"report": "ready"},
	}); err != nil {
		t.Fatalf("FinishNode(): %v", err)
	}
	if err := runtime.FinishRun(context.Background(), journal.FinishRunRequest{
		RunID:           runID,
		WorkflowID:      "consumer-workflow",
		WorkflowVersion: "1",
		Status:          "completed",
		FinishedAt:      103,
	}); err != nil {
		t.Fatalf("FinishRun(): %v", err)
	}
	if port.runs[runID].Status != "completed" ||
		len(port.nodes) != 2 ||
		len(port.events) != 6 {
		t.Fatalf("durable projection runs=%#v nodes=%#v events=%#v", port.runs, port.nodes, port.events)
	}
}

type memoryPort struct {
	runs   map[string]journal.Run
	nodes  []journal.NodeExecution
	events []journal.Event
}

func (p *memoryPort) LoadRun(_ context.Context, runID string) (journal.Run, bool, error) {
	run, ok := p.runs[runID]
	return run, ok, nil
}

func (p *memoryPort) CreateRun(_ context.Context, run journal.Run) error {
	p.runs[run.RunID] = run
	return nil
}

func (p *memoryPort) UpdateRun(_ context.Context, run journal.Run) error {
	p.runs[run.RunID] = run
	return nil
}

func (p *memoryPort) UpsertNodeExecution(_ context.Context, node journal.NodeExecution) error {
	p.nodes = append(p.nodes, node)
	return nil
}

func (p *memoryPort) AppendEvent(_ context.Context, event journal.Event) error {
	p.events = append(p.events, event)
	return nil
}

func sequence(values ...string) func() string {
	index := 0
	return func() string {
		value := values[index]
		index++
		return value
	}
}

func sequenceInt64(values ...int64) func() int64 {
	index := 0
	return func() int64 {
		value := values[index]
		index++
		return value
	}
}
