package runstore

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func TestMemoryStoreRunLifecycleAndEvents(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	if err := store.CreateRun(ctx, Run{
		RunID:      " run-1 ",
		WorkflowID: " browser-ops ",
		Status:     "running",
		Attempt:    1,
	}); err != nil {
		t.Fatalf("create run: %v", err)
	}
	run, err := store.GetRun(ctx, "run-1")
	if err != nil {
		t.Fatalf("get run: %v", err)
	}
	if run.RunID != "run-1" || run.WorkflowID != " browser-ops " {
		t.Fatalf("unexpected run: %#v", run)
	}
	if err := store.AppendEvent(ctx, Event{
		EventID:   "event-2",
		RunID:     "run-1",
		Name:      "tool.finish",
		CreatedAt: 20,
	}); err != nil {
		t.Fatalf("append second event: %v", err)
	}
	if err := store.AppendEvent(ctx, Event{
		EventID:   "event-1",
		RunID:     "run-1",
		Name:      "run.start",
		CreatedAt: 10,
	}); err != nil {
		t.Fatalf("append first event: %v", err)
	}
	if err := store.AppendEvent(ctx, Event{
		EventID:   "event-1",
		RunID:     "run-1",
		Name:      "run.start.duplicate",
		CreatedAt: 30,
	}); !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("expected duplicate event id to return ErrAlreadyExists, got %v", err)
	}
	events, err := store.ListEvents(ctx, "run-1", 0)
	if err != nil {
		t.Fatalf("list events: %v", err)
	}
	if len(events) != 2 || events[0].EventID != "event-1" || events[1].EventID != "event-2" {
		t.Fatalf("unexpected events ordering: %#v", events)
	}
	if err := store.UpdateRun(ctx, Run{
		RunID:      "run-1",
		WorkflowID: "browser-ops",
		Status:     "completed",
		Attempt:    1,
		FinishedAt: 30,
	}); err != nil {
		t.Fatalf("update run: %v", err)
	}
	run, err = store.GetRun(ctx, "run-1")
	if err != nil {
		t.Fatalf("get run after update: %v", err)
	}
	if run.Status != "completed" || run.FinishedAt != 30 {
		t.Fatalf("unexpected updated run: %#v", run)
	}
}

func TestMemoryStoreRunLifecyclePreservesRawOptionalFields(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	if err := store.CreateRun(ctx, Run{
		RunID:       " run-1 ",
		CaseID:      " case-1 ",
		WorkflowID:  " browser-ops ",
		WorkflowVer: " v1 ",
		Status:      " running ",
		ParentRunID: " parent-run ",
		RootRunID:   " root-run ",
		ContractID:  " contract-1 ",
		Summary:     " raw summary ",
	}); err != nil {
		t.Fatalf("create run: %v", err)
	}

	run, err := store.GetRun(ctx, "run-1")
	if err != nil {
		t.Fatalf("get run: %v", err)
	}
	if run.RunID != "run-1" ||
		run.CaseID != " case-1 " ||
		run.WorkflowID != " browser-ops " ||
		run.WorkflowVer != " v1 " ||
		run.Status != " running " ||
		run.ParentRunID != " parent-run " ||
		run.RootRunID != " root-run " ||
		run.ContractID != " contract-1 " ||
		run.Summary != " raw summary " {
		t.Fatalf("expected raw optional run fields, got %#v", run)
	}
}

func TestMemoryStoreNodeExecutionLifecycle(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()
	if err := store.CreateRun(ctx, Run{RunID: "run-1"}); err != nil {
		t.Fatalf("create run: %v", err)
	}
	if err := store.UpsertNodeExecution(ctx, NodeExecution{
		NodeExecID: "node-exec-2",
		RunID:      "run-1",
		NodeID:     "evaluate",
		StartedAt:  20,
	}); err != nil {
		t.Fatalf("upsert second node exec: %v", err)
	}
	if err := store.UpsertNodeExecution(ctx, NodeExecution{
		NodeExecID:                "node-exec-1",
		RunID:                     "run-1",
		ParentNodeExecID:          "node-root",
		NodeID:                    "plan",
		ExecutionContractID:       "contract-1",
		ExecutionContractDiffJSON: ` {"changed_fields":["visibility"]} `,
		TerminationJSON:           ` {"kind":"max_rounds","checkpoint_stage":"max_rounds_break"} `,
		DelegatedExecutionJSON:    ` {"driver":"open_tool_loop","rounds":[{"round":1,"outcome_kind":"completed"}]} `,
		StartedAt:                 10,
	}); err != nil {
		t.Fatalf("upsert first node exec: %v", err)
	}
	items, err := store.ListNodeExecutions(ctx, "run-1")
	if err != nil {
		t.Fatalf("list node executions: %v", err)
	}
	if len(items) != 2 || items[0].NodeExecID != "node-exec-1" || items[1].NodeExecID != "node-exec-2" {
		t.Fatalf("unexpected node execution ordering: %#v", items)
	}
	var delegated map[string]any
	if err := json.Unmarshal([]byte(items[0].DelegatedExecutionJSON), &delegated); err != nil {
		t.Fatalf("decode delegated execution json: %v", err)
	}
	if delegated["driver"] != "open_tool_loop" {
		t.Fatalf("unexpected delegated execution json: %#v", delegated)
	}
	if items[0].ExecutionContractID != "contract-1" {
		t.Fatalf("expected execution contract id, got %#v", items[0])
	}
	if items[0].ParentNodeExecID != "node-root" {
		t.Fatalf("expected parent node exec id, got %#v", items[0])
	}
	if diff := items[0].ExecutionContractDiff(); len(diff) != 1 || diff[0] != "visibility" {
		t.Fatalf("expected execution contract diff, got %#v", diff)
	}
	if termination := items[0].TerminationProjection(); termination == nil || termination.Kind != "max_rounds" {
		t.Fatalf("expected termination projection, got %#v", termination)
	}
}

func TestMemoryStoreEventAndNodeExecutionPreserveRawOptionalFields(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()
	if err := store.CreateRun(ctx, Run{RunID: "run-1"}); err != nil {
		t.Fatalf("create run: %v", err)
	}
	if err := store.AppendEvent(ctx, Event{
		EventID:     " event-1 ",
		RunID:       " run-1 ",
		BranchID:    " branch-1 ",
		NodeExecID:  " node-exec-1 ",
		Name:        " run.start ",
		PayloadJSON: " {\"ok\":true} ",
		CreatedAt:   10,
	}); err != nil {
		t.Fatalf("append event: %v", err)
	}
	events, err := store.ListEvents(ctx, "run-1", 0)
	if err != nil {
		t.Fatalf("list events: %v", err)
	}
	if len(events) != 1 ||
		events[0].EventID != "event-1" ||
		events[0].RunID != "run-1" ||
		events[0].BranchID != " branch-1 " ||
		events[0].NodeExecID != " node-exec-1 " ||
		events[0].Name != " run.start " ||
		events[0].PayloadJSON != " {\"ok\":true} " {
		t.Fatalf("expected raw optional event fields, got %#v", events)
	}

	if err := store.UpsertNodeExecution(ctx, NodeExecution{
		NodeExecID:                " node-exec-1 ",
		RunID:                     " run-1 ",
		BranchID:                  " branch-1 ",
		ParentNodeExecID:          " node-root ",
		NodeID:                    " main/round:1 ",
		Kind:                      " tool ",
		Status:                    " completed ",
		InputStateRef:             " input-ref ",
		OutputStateRef:            " output-ref ",
		ExecutionContractID:       " contract-1 ",
		ExecutionContractDiffJSON: ` {"changed_fields":[" visibility "]} `,
		TerminationJSON:           ` {"kind":" max_rounds "} `,
		DelegatedExecutionJSON:    ` {"driver":" open_tool_loop ","rounds":[{"node_exec_id":" round-1 ","round":1,"stop_reason":" max_rounds "}]} `,
		StartedAt:                 10,
	}); err != nil {
		t.Fatalf("upsert node execution: %v", err)
	}
	nodes, err := store.ListNodeExecutions(ctx, "run-1")
	if err != nil {
		t.Fatalf("list node executions: %v", err)
	}
	if len(nodes) != 1 ||
		nodes[0].NodeExecID != "node-exec-1" ||
		nodes[0].RunID != "run-1" ||
		nodes[0].BranchID != " branch-1 " ||
		nodes[0].ParentNodeExecID != " node-root " ||
		nodes[0].NodeID != " main/round:1 " ||
		nodes[0].Kind != " tool " ||
		nodes[0].Status != " completed " ||
		nodes[0].InputStateRef != " input-ref " ||
		nodes[0].OutputStateRef != " output-ref " ||
		nodes[0].ExecutionContractID != " contract-1 " ||
		nodes[0].ExecutionContractDiffJSON != ` {"changed_fields":[" visibility "]} ` ||
		nodes[0].TerminationJSON != ` {"kind":" max_rounds "} ` ||
		nodes[0].DelegatedExecutionJSON != ` {"driver":" open_tool_loop ","rounds":[{"node_exec_id":" round-1 ","round":1,"stop_reason":" max_rounds "}]} ` {
		t.Fatalf("expected raw optional node fields, got %#v", nodes)
	}
	if diff := nodes[0].ExecutionContractDiff(); len(diff) != 1 || diff[0] != " visibility " {
		t.Fatalf("expected raw diff field, got %#v", diff)
	}
	if termination := nodes[0].TerminationProjection(); termination == nil || termination.Kind != " max_rounds " {
		t.Fatalf("expected raw termination projection, got %#v", termination)
	}
	if delegated := nodes[0].DelegatedExecutionProjection(); delegated == nil || delegated.Driver != " open_tool_loop " {
		t.Fatalf("expected raw delegated projection, got %#v", delegated)
	}
}

func TestMemoryStoreRejectsMissingRun(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()
	if err := store.AppendEvent(ctx, Event{
		EventID: "event-1",
		RunID:   "missing-run",
	}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected append event to fail with ErrNotFound, got %v", err)
	}
	if err := store.UpsertNodeExecution(ctx, NodeExecution{
		NodeExecID: "node-exec-1",
		RunID:      "missing-run",
	}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected upsert node execution to fail with ErrNotFound, got %v", err)
	}
}
