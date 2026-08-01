package runstore

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func TestRecordJSONContract(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want string
	}{
		{name: "run", in: Run{RunID: "run-1"}, want: `{"run_id":"run-1"}`},
		{name: "node", in: NodeExecution{NodeExecID: "node-1", RunID: "run-1"}, want: `{"node_exec_id":"node-1","run_id":"run-1"}`},
		{name: "event", in: Event{EventID: "event-1", RunID: "run-1"}, want: `{"event_id":"event-1","run_id":"run-1"}`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			payload, err := json.Marshal(tc.in)
			if err != nil {
				t.Fatalf("Marshal(): %v", err)
			}
			if string(payload) != tc.want {
				t.Fatalf("JSON = %s, want %s", payload, tc.want)
			}
		})
	}
}

func TestMemoryStoreNilReceiverContract(t *testing.T) {
	var store *MemoryStore
	ctx := context.Background()
	if err := store.CreateRun(ctx, Run{}); err != nil {
		t.Fatalf("nil CreateRun() error = %v", err)
	}
	if err := store.UpdateRun(ctx, Run{}); err != nil {
		t.Fatalf("nil UpdateRun() error = %v", err)
	}
	if _, err := store.GetRun(ctx, "run-1"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("nil GetRun() error = %v", err)
	}
	if err := store.AppendEvent(ctx, Event{}); err != nil {
		t.Fatalf("nil AppendEvent() error = %v", err)
	}
	if events, err := store.ListEvents(ctx, "run-1", 0); err != nil || events != nil {
		t.Fatalf("nil ListEvents() = %#v, %v", events, err)
	}
	if err := store.UpsertNodeExecution(ctx, NodeExecution{}); err != nil {
		t.Fatalf("nil UpsertNodeExecution() error = %v", err)
	}
	if nodes, err := store.ListNodeExecutions(ctx, "run-1"); err != nil || nodes != nil {
		t.Fatalf("nil ListNodeExecutions() = %#v, %v", nodes, err)
	}
}

func TestMemoryStoreNormalizationAndValidationContract(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()
	if err := store.CreateRun(ctx, Run{RunID: " run-1 ", Attempt: -1}); err != nil {
		t.Fatalf("CreateRun(): %v", err)
	}
	run, err := store.GetRun(ctx, "run-1")
	if err != nil || run.Attempt != 0 {
		t.Fatalf("GetRun() = %#v, %v", run, err)
	}
	if err := store.UpsertNodeExecution(ctx, NodeExecution{NodeExecID: " node-1 ", RunID: " run-1 ", Attempt: -2}); err != nil {
		t.Fatalf("UpsertNodeExecution(): %v", err)
	}
	nodes, err := store.ListNodeExecutions(ctx, "run-1")
	if err != nil || len(nodes) != 1 || nodes[0].Attempt != 0 || nodes[0].NodeExecID != "node-1" {
		t.Fatalf("ListNodeExecutions() = %#v, %v", nodes, err)
	}
	if err := store.CreateRun(ctx, Run{}); err == nil || err.Error() != "agentx/runstore: run id is required" {
		t.Fatalf("empty CreateRun() error = %v", err)
	}
	if err := store.AppendEvent(ctx, Event{EventID: "event-1"}); err == nil || err.Error() != "agentx/runstore: event id and run id are required" {
		t.Fatalf("empty AppendEvent() error = %v", err)
	}
	if err := store.UpsertNodeExecution(ctx, NodeExecution{RunID: "run-1"}); err == nil || err.Error() != "agentx/runstore: node exec id and run id are required" {
		t.Fatalf("empty UpsertNodeExecution() error = %v", err)
	}
}

func TestMemoryStoreEventIDsAreGloballyUnique(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()
	for _, runID := range []string{"run-1", "run-2"} {
		if err := store.CreateRun(ctx, Run{RunID: runID}); err != nil {
			t.Fatalf("CreateRun(%s): %v", runID, err)
		}
	}
	if err := store.AppendEvent(ctx, Event{EventID: "event-1", RunID: "run-1"}); err != nil {
		t.Fatalf("AppendEvent(): %v", err)
	}
	if err := store.AppendEvent(ctx, Event{EventID: "event-1", RunID: "run-2"}); !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("cross-run duplicate error = %v", err)
	}
}
