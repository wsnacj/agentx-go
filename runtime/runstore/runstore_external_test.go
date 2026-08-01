package runstore_test

import (
	"context"
	"errors"
	"testing"

	runstore "github.com/wsnacj/agentx-go/runtime/runstore"
)

var _ runstore.Store = (*runstore.MemoryStore)(nil)

func TestExternalPackageRunDataPlaneConsumer(t *testing.T) {
	ctx := context.Background()
	store := runstore.NewMemoryStore()
	if err := store.CreateRun(ctx, runstore.Run{
		RunID:      " run-1 ",
		WorkflowID: " workflow-1 ",
		Status:     "running",
	}); err != nil {
		t.Fatalf("CreateRun(): %v", err)
	}
	if err := store.CreateRun(ctx, runstore.Run{RunID: "run-1"}); !errors.Is(err, runstore.ErrAlreadyExists) {
		t.Fatalf("duplicate CreateRun() error = %v", err)
	}

	for _, event := range []runstore.Event{
		{EventID: "event-2", RunID: "run-1", Name: "node.finish", CreatedAt: 20},
		{EventID: "event-1", RunID: "run-1", Name: "run.start", CreatedAt: 10},
	} {
		if err := store.AppendEvent(ctx, event); err != nil {
			t.Fatalf("AppendEvent(%s): %v", event.EventID, err)
		}
	}
	events, err := store.ListEvents(ctx, "run-1", 1)
	if err != nil {
		t.Fatalf("ListEvents(): %v", err)
	}
	if len(events) != 1 || events[0].EventID != "event-1" {
		t.Fatalf("events = %#v", events)
	}

	for _, node := range []runstore.NodeExecution{
		{
			NodeExecID:       "child-1",
			RunID:            "run-1",
			ParentNodeExecID: "root-1",
			NodeID:           "main/round:1",
			Status:           "completed",
			StartedAt:        10,
			FinishedAt:       20,
		},
		{
			NodeExecID:      "root-1",
			RunID:           "run-1",
			NodeID:          "main",
			Status:          "incomplete",
			TerminationJSON: `{"kind":"max_rounds"}`,
			StartedAt:       1,
			FinishedAt:      30,
		},
	} {
		if err := store.UpsertNodeExecution(ctx, node); err != nil {
			t.Fatalf("UpsertNodeExecution(%s): %v", node.NodeExecID, err)
		}
	}
	nodes, err := store.ListNodeExecutions(ctx, "run-1")
	if err != nil {
		t.Fatalf("ListNodeExecutions(): %v", err)
	}
	projections := make([]runstore.NodeExecutionProjection, 0, len(nodes))
	for _, node := range nodes {
		projections = append(projections, *node.Projection())
	}
	tree := runstore.AttachChildNodeExecutionProjections(projections)
	if len(tree) != 1 || tree[0].NodeExecID != "root-1" || len(tree[0].ChildNodeExecutions) != 1 {
		t.Fatalf("projection tree = %#v", tree)
	}
	selected := runstore.SelectTerminalNodeExecutionProjection(tree)
	if selected == nil || selected.NodeExecID != "root-1" || selected.Termination == nil || selected.Termination.Kind != "max_rounds" {
		t.Fatalf("terminal projection = %#v", selected)
	}
	selected.ChildNodeExecutions[0].NodeID = "mutated"
	if tree[0].ChildNodeExecutions[0].NodeID != "main/round:1" {
		t.Fatalf("selection leaked mutable child projection: %#v", tree)
	}
}
