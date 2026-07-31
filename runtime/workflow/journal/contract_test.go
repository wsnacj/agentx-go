package journal

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

func TestRunLifecyclePreservesFieldsPayloadsAndWriteOrder(t *testing.T) {
	port := newCapturePort()
	ids := stringSequence("run-generated", "event-start", "event-finish")
	journal := New(Dependencies{
		Port:         port,
		NewRunID:     ids,
		NewEventID:   ids,
		NowUnixMilli: int64Sequence(101, 102),
	})
	runID, err := journal.EnsureRun(context.Background(), EnsureRunRequest{
		CaseID:          " case-1 ",
		WorkflowID:      " workflow-1 ",
		WorkflowVersion: " version-1 ",
	})
	if err != nil {
		t.Fatalf("EnsureRun(): %v", err)
	}
	if err := journal.AppendRunStart(context.Background(), StartRunEventRequest{
		RunID:      runID,
		BranchID:   " branch-1 ",
		WorkflowID: " workflow-1 ",
		EntryNode:  " entry-1 ",
	}); err != nil {
		t.Fatalf("AppendRunStart(): %v", err)
	}
	if got, want := port.operations, []string{
		"load:run-generated",
		"create:run-generated",
		"event:workflow.start",
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("start operations = %#v, want %#v", got, want)
	}
	created := port.runs[runID]
	if created.CaseID != " case-1 " ||
		created.WorkflowID != " workflow-1 " ||
		created.WorkflowVersion != " version-1 " ||
		created.Status != "running" ||
		created.Attempt != 1 ||
		created.RootRunID != runID ||
		created.StartedAt != 101 {
		t.Fatalf("created run = %#v", created)
	}
	start := port.events[0]
	if start.EventID != "event-start" ||
		start.BranchID != " branch-1 " ||
		start.CreatedAt != 102 {
		t.Fatalf("start event = %#v", start)
	}
	requireJSONMap(t, start.PayloadJSON, map[string]any{
		"workflow_id": " workflow-1 ",
		"entry_node":  " entry-1 ",
	})

	port.operations = nil
	if err := journal.FinishRun(context.Background(), FinishRunRequest{
		RunID:           runID,
		WorkflowID:      " runtime-workflow ",
		WorkflowVersion: " runtime-version ",
		Status:          " completed ",
		FinishedAt:      456,
		ErrorText:       " raw summary ",
	}); err != nil {
		t.Fatalf("FinishRun(): %v", err)
	}
	if got, want := port.operations, []string{
		"load:run-generated",
		"update:run-generated",
		"event:workflow.finish",
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("finish operations = %#v, want %#v", got, want)
	}
	finished := port.runs[runID]
	if finished.WorkflowID != " runtime-workflow " ||
		finished.WorkflowVersion != " runtime-version " ||
		finished.Status != " completed " ||
		finished.FinishedAt != 456 ||
		finished.Summary != " raw summary " {
		t.Fatalf("finished run = %#v", finished)
	}
	requireJSONMap(t, port.events[1].PayloadJSON, map[string]any{
		"workflow_id": " runtime-workflow ",
		"status":      " completed ",
		"error":       " raw summary ",
	})
}

func TestEnsureRunResumesWithoutOverwritingFallbackFields(t *testing.T) {
	port := newCapturePort()
	port.runs[" run-1 "] = Run{
		RunID:           " run-1 ",
		CaseID:          " original-case ",
		WorkflowID:      " original-workflow ",
		WorkflowVersion: " original-version ",
		Attempt:         3,
		RootRunID:       " original-root ",
		StartedAt:       88,
	}
	journal := New(Dependencies{Port: port})
	runID, err := journal.EnsureRun(context.Background(), EnsureRunRequest{
		RunID:      " run-1 ",
		WorkflowID: " override-workflow ",
	})
	if err != nil {
		t.Fatalf("EnsureRun(): %v", err)
	}
	got := port.runs[runID]
	if got.CaseID != " original-case " ||
		got.WorkflowID != " override-workflow " ||
		got.WorkflowVersion != " original-version " ||
		got.Status != "running" ||
		got.Attempt != 3 ||
		got.RootRunID != " original-root " ||
		got.StartedAt != 88 {
		t.Fatalf("updated run = %#v", got)
	}
}

func TestNodeLifecyclePreservesSnapshotReferencesAndWriteOrder(t *testing.T) {
	port := newCapturePort()
	journal := New(Dependencies{
		Port:       port,
		NewEventID: stringSequence("input-ref", "event-start", "output-ref", "event-finish"),
	})
	node := NodeExecution{
		NodeExecutionID: " nodeexec-1 ",
		RunID:           " run-1 ",
		BranchID:        " branch-1 ",
		NodeID:          " node-1 ",
		Kind:            " tool ",
		Status:          "running",
		Attempt:         1,
		StartedAt:       123,
	}
	inputRef, err := journal.StartNode(context.Background(), StartNodeRequest{
		Node:  node,
		State: map[string]any{"phase": "input"},
		EventPayload: map[string]any{
			"node_id":   " node-1 ",
			"tool_name": " echo ",
		},
	})
	if err != nil {
		t.Fatalf("StartNode(): %v", err)
	}
	node.Status = " completed "
	node.InputStateRef = inputRef
	node.ExecutionContractID = " contract-1 "
	node.FinishedAt = 456
	outputRef, err := journal.FinishNode(context.Background(), FinishNodeRequest{
		Node:         node,
		State:        map[string]any{"phase": "output"},
		EventPayload: map[string]any{"node_id": " node-1 ", "status": " completed "},
	})
	if err != nil {
		t.Fatalf("FinishNode(): %v", err)
	}
	if inputRef != "input-ref" || outputRef != "output-ref" {
		t.Fatalf("refs = %q/%q", inputRef, outputRef)
	}
	want := []string{
		"event:workflow.node.state.input",
		"node: nodeexec-1 ",
		"event:workflow.node.start",
		"event:workflow.node.state.output",
		"node: nodeexec-1 ",
		"event:workflow.node.finish",
	}
	if !reflect.DeepEqual(port.operations, want) {
		t.Fatalf("operations = %#v, want %#v", port.operations, want)
	}
	if len(port.nodes) != 2 ||
		port.nodes[0].InputStateRef != "input-ref" ||
		port.nodes[0].OutputStateRef != "" ||
		port.nodes[1].InputStateRef != "input-ref" ||
		port.nodes[1].OutputStateRef != "output-ref" {
		t.Fatalf("node records = %#v", port.nodes)
	}
	requireJSONMap(t, port.events[0].PayloadJSON, map[string]any{
		"node_id": " node-1 ",
		"state":   map[string]any{"phase": "input"},
	})
}

func TestJournalStopsAtFirstPortError(t *testing.T) {
	sentinel := errors.New("durable failure")
	tests := []struct {
		name   string
		failAt int
		call   func(*Journal) error
		want   []string
	}{
		{
			name:   "snapshot",
			failAt: 1,
			call: func(j *Journal) error {
				_, err := j.StartNode(context.Background(), StartNodeRequest{
					Node: NodeExecution{NodeExecutionID: "node-1", RunID: "run-1", NodeID: "step"},
				})
				return err
			},
			want: []string{"event:workflow.node.state.input"},
		},
		{
			name:   "node upsert",
			failAt: 2,
			call: func(j *Journal) error {
				_, err := j.StartNode(context.Background(), StartNodeRequest{
					Node: NodeExecution{NodeExecutionID: "node-1", RunID: "run-1", NodeID: "step"},
				})
				return err
			},
			want: []string{"event:workflow.node.state.input", "node:node-1"},
		},
		{
			name:   "finish update",
			failAt: 2,
			call: func(j *Journal) error {
				return j.FinishRun(context.Background(), FinishRunRequest{RunID: "run-1"})
			},
			want: []string{"load:run-1", "update:run-1"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			port := newCapturePort()
			port.failAt = tt.failAt
			port.failure = sentinel
			journal := New(Dependencies{
				Port:       port,
				NewEventID: stringSequence("event-1", "event-2"),
			})
			if err := tt.call(journal); !errors.Is(err, sentinel) {
				t.Fatalf("error = %v, want sentinel", err)
			}
			if !reflect.DeepEqual(port.operations, tt.want) {
				t.Fatalf("operations = %#v, want %#v", port.operations, tt.want)
			}
		})
	}
}

func TestNilPortIsBoundedNoop(t *testing.T) {
	journal := New(Dependencies{NewRunID: stringSequence("run-generated")})
	runID, err := journal.EnsureRun(context.Background(), EnsureRunRequest{})
	if err != nil || runID != "run-generated" {
		t.Fatalf("EnsureRun() = %q, %v", runID, err)
	}
	if err := journal.AppendRunStart(context.Background(), StartRunEventRequest{RunID: runID}); err != nil {
		t.Fatalf("AppendRunStart(): %v", err)
	}
	if ref, err := journal.StartNode(context.Background(), StartNodeRequest{}); err != nil || ref != "" {
		t.Fatalf("StartNode() = %q, %v", ref, err)
	}
	if ref, err := journal.FinishNode(context.Background(), FinishNodeRequest{}); err != nil || ref != "" {
		t.Fatalf("FinishNode() = %q, %v", ref, err)
	}
	if err := journal.FinishRun(context.Background(), FinishRunRequest{}); err != nil {
		t.Fatalf("FinishRun(): %v", err)
	}
}

type capturePort struct {
	runs       map[string]Run
	nodes      []NodeExecution
	events     []Event
	operations []string
	failAt     int
	failure    error
}

func newCapturePort() *capturePort {
	return &capturePort{runs: map[string]Run{}}
}

func (p *capturePort) LoadRun(_ context.Context, runID string) (Run, bool, error) {
	if err := p.record("load:" + runID); err != nil {
		return Run{}, false, err
	}
	run, ok := p.runs[runID]
	return run, ok, nil
}

func (p *capturePort) CreateRun(_ context.Context, run Run) error {
	if err := p.record("create:" + run.RunID); err != nil {
		return err
	}
	p.runs[run.RunID] = run
	return nil
}

func (p *capturePort) UpdateRun(_ context.Context, run Run) error {
	if err := p.record("update:" + run.RunID); err != nil {
		return err
	}
	p.runs[run.RunID] = run
	return nil
}

func (p *capturePort) UpsertNodeExecution(_ context.Context, node NodeExecution) error {
	if err := p.record("node:" + node.NodeExecutionID); err != nil {
		return err
	}
	p.nodes = append(p.nodes, node)
	return nil
}

func (p *capturePort) AppendEvent(_ context.Context, event Event) error {
	if err := p.record("event:" + event.Name); err != nil {
		return err
	}
	p.events = append(p.events, event)
	return nil
}

func (p *capturePort) record(operation string) error {
	p.operations = append(p.operations, operation)
	if p.failAt > 0 && len(p.operations) == p.failAt {
		return p.failure
	}
	return nil
}

func stringSequence(values ...string) func() string {
	index := 0
	return func() string {
		value := values[index]
		index++
		return value
	}
}

func int64Sequence(values ...int64) func() int64 {
	index := 0
	return func() int64 {
		value := values[index]
		index++
		return value
	}
}

func requireJSONMap(t *testing.T, raw string, want map[string]any) {
	t.Helper()
	var got map[string]any
	if err := json.Unmarshal([]byte(raw), &got); err != nil {
		t.Fatalf("decode JSON %q: %v", raw, err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("JSON = %#v, want %#v", got, want)
	}
}
