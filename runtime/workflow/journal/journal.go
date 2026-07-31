// Package journal 提供与存储 backend 无关的 Workflow durable write ordering。
//
// 当前 package 处于 Experimental/private validation。它不拥有具体 RunStore、
// lowering、node executor、retry、resume、provider 或 Scene policy。
package journal

import (
	"context"
	"encoding/json"
	"fmt"
)

// Run is the portable workflow run record used by the journal port.
type Run struct {
	RunID           string
	CaseID          string
	WorkflowID      string
	WorkflowVersion string
	Status          string
	Attempt         int
	ParentRunID     string
	RootRunID       string
	ContractID      string
	StartedAt       int64
	FinishedAt      int64
	Summary         string
}

// NodeExecution is the portable node execution record used by the journal port.
type NodeExecution struct {
	NodeExecutionID           string
	RunID                     string
	BranchID                  string
	ParentNodeExecutionID     string
	NodeID                    string
	Kind                      string
	Status                    string
	Attempt                   int
	InputStateRef             string
	OutputStateRef            string
	ExecutionContractID       string
	ExecutionContractDiffJSON string
	TerminationJSON           string
	DelegatedExecutionJSON    string
	StartedAt                 int64
	FinishedAt                int64
}

// Event is the portable durable event record used by the journal port.
type Event struct {
	EventID         string
	RunID           string
	BranchID        string
	NodeExecutionID string
	Name            string
	PayloadJSON     string
	CreatedAt       int64
}

// Port is the minimal storage boundary required by Journal.
type Port interface {
	LoadRun(context.Context, string) (Run, bool, error)
	CreateRun(context.Context, Run) error
	UpdateRun(context.Context, Run) error
	UpsertNodeExecution(context.Context, NodeExecution) error
	AppendEvent(context.Context, Event) error
}

// Dependencies contains host-owned identity, clock, and storage adapters.
type Dependencies struct {
	Port         Port
	NewRunID     func() string
	NewEventID   func() string
	NowUnixMilli func() int64
}

// Journal owns fail-fast durable write ordering for one workflow host.
type Journal struct {
	port         Port
	newRunID     func() string
	newEventID   func() string
	nowUnixMilli func() int64
}

// EnsureRunRequest contains the host projection required to create or resume a
// run.
type EnsureRunRequest struct {
	RunID           string
	CaseID          string
	WorkflowID      string
	WorkflowVersion string
}

// StartRunEventRequest contains the host projection for workflow.start.
type StartRunEventRequest struct {
	RunID      string
	BranchID   string
	WorkflowID string
	EntryNode  string
}

// FinishRunRequest contains the host projection required to finish a run.
type FinishRunRequest struct {
	RunID           string
	WorkflowID      string
	WorkflowVersion string
	Status          string
	FinishedAt      int64
	ErrorText       string
}

// StartNodeRequest contains the record, state, and event payload for node start.
type StartNodeRequest struct {
	Node         NodeExecution
	State        map[string]any
	EventPayload map[string]any
}

// FinishNodeRequest contains the record, state, and event payload for node
// finish.
type FinishNodeRequest struct {
	Node         NodeExecution
	State        map[string]any
	EventPayload map[string]any
}

// New constructs a durable journal from host-owned dependencies.
func New(dependencies Dependencies) *Journal {
	return &Journal{
		port:         dependencies.Port,
		newRunID:     dependencies.NewRunID,
		newEventID:   dependencies.NewEventID,
		nowUnixMilli: dependencies.NowUnixMilli,
	}
}

// EnsureRun creates or resumes a run.
func (j *Journal) EnsureRun(ctx context.Context, request EnsureRunRequest) (string, error) {
	runID := request.RunID
	if j == nil {
		if runID == "" {
			return "", fmt.Errorf("workflow journal: journal is required to generate run id")
		}
		return runID, nil
	}
	if runID == "" {
		generated, err := generatedValue(j.newRunID, "run id")
		if err != nil {
			return "", err
		}
		runID = generated
	}
	if j.port == nil {
		return runID, nil
	}

	current, found, err := j.port.LoadRun(ctx, runID)
	if err != nil {
		return "", err
	}
	if found {
		current.CaseID = firstNonEmpty(request.CaseID, current.CaseID)
		current.WorkflowID = firstNonEmpty(request.WorkflowID, current.WorkflowID)
		current.WorkflowVersion = firstNonEmpty(request.WorkflowVersion, current.WorkflowVersion)
		current.Status = "running"
		if current.Attempt <= 0 {
			current.Attempt = 1
		}
		if current.RootRunID == "" {
			current.RootRunID = runID
		}
		if current.StartedAt <= 0 {
			startedAt, err := currentTime(j.nowUnixMilli)
			if err != nil {
				return "", err
			}
			current.StartedAt = startedAt
		}
		if err := j.port.UpdateRun(ctx, current); err != nil {
			return "", err
		}
	} else {
		startedAt, err := currentTime(j.nowUnixMilli)
		if err != nil {
			return "", err
		}
		if err := j.port.CreateRun(ctx, Run{
			RunID:           runID,
			CaseID:          request.CaseID,
			WorkflowID:      request.WorkflowID,
			WorkflowVersion: request.WorkflowVersion,
			Status:          "running",
			Attempt:         1,
			RootRunID:       runID,
			StartedAt:       startedAt,
		}); err != nil {
			return "", err
		}
	}
	return runID, nil
}

// AppendRunStart appends workflow.start after the host has initialized its
// observable execution result.
func (j *Journal) AppendRunStart(ctx context.Context, request StartRunEventRequest) error {
	if j == nil || j.port == nil {
		return nil
	}
	createdAt, err := currentTime(j.nowUnixMilli)
	if err != nil {
		return err
	}
	eventID, err := generatedValue(j.newEventID, "event id")
	if err != nil {
		return err
	}
	return j.appendEvent(ctx, Event{
		EventID:   eventID,
		RunID:     request.RunID,
		BranchID:  request.BranchID,
		Name:      "workflow.start",
		CreatedAt: createdAt,
	}, map[string]any{
		"workflow_id": request.WorkflowID,
		"entry_node":  request.EntryNode,
	})
}

// StartNode appends the input snapshot, upserts the running node, then appends
// workflow.node.start.
func (j *Journal) StartNode(ctx context.Context, request StartNodeRequest) (string, error) {
	if j == nil || j.port == nil {
		return "", nil
	}
	inputStateRef, err := j.appendStateSnapshot(
		ctx,
		request.Node,
		"workflow.node.state.input",
		request.State,
		request.Node.StartedAt,
	)
	if err != nil {
		return "", err
	}
	node := request.Node
	node.InputStateRef = inputStateRef
	if err := j.port.UpsertNodeExecution(ctx, node); err != nil {
		return "", err
	}
	eventID, err := generatedValue(j.newEventID, "event id")
	if err != nil {
		return "", err
	}
	if err := j.appendEvent(ctx, Event{
		EventID:         eventID,
		RunID:           node.RunID,
		BranchID:        node.BranchID,
		NodeExecutionID: node.NodeExecutionID,
		Name:            "workflow.node.start",
		CreatedAt:       node.StartedAt,
	}, request.EventPayload); err != nil {
		return "", err
	}
	return inputStateRef, nil
}

// FinishNode appends the output snapshot, upserts the final node, then appends
// workflow.node.finish.
func (j *Journal) FinishNode(ctx context.Context, request FinishNodeRequest) (string, error) {
	if j == nil || j.port == nil {
		return "", nil
	}
	outputStateRef, err := j.appendStateSnapshot(
		ctx,
		request.Node,
		"workflow.node.state.output",
		request.State,
		request.Node.FinishedAt,
	)
	if err != nil {
		return "", err
	}
	node := request.Node
	node.OutputStateRef = outputStateRef
	if err := j.port.UpsertNodeExecution(ctx, node); err != nil {
		return "", err
	}
	eventID, err := generatedValue(j.newEventID, "event id")
	if err != nil {
		return "", err
	}
	if err := j.appendEvent(ctx, Event{
		EventID:         eventID,
		RunID:           node.RunID,
		BranchID:        node.BranchID,
		NodeExecutionID: node.NodeExecutionID,
		Name:            "workflow.node.finish",
		CreatedAt:       node.FinishedAt,
	}, request.EventPayload); err != nil {
		return "", err
	}
	return outputStateRef, nil
}

// FinishRun updates the run, then appends workflow.finish.
func (j *Journal) FinishRun(ctx context.Context, request FinishRunRequest) error {
	if j == nil || j.port == nil {
		return nil
	}
	current, _, err := j.port.LoadRun(ctx, request.RunID)
	if err != nil {
		return err
	}
	current.RunID = request.RunID
	current.WorkflowID = firstNonEmpty(request.WorkflowID, current.WorkflowID)
	current.WorkflowVersion = firstNonEmpty(request.WorkflowVersion, current.WorkflowVersion)
	current.Status = request.Status
	if current.Attempt <= 0 {
		current.Attempt = 1
	}
	if current.RootRunID == "" {
		current.RootRunID = request.RunID
	}
	current.FinishedAt = request.FinishedAt
	current.Summary = request.ErrorText
	if err := j.port.UpdateRun(ctx, current); err != nil {
		return err
	}
	eventID, err := generatedValue(j.newEventID, "event id")
	if err != nil {
		return err
	}
	return j.appendEvent(ctx, Event{
		EventID:   eventID,
		RunID:     request.RunID,
		Name:      "workflow.finish",
		CreatedAt: request.FinishedAt,
	}, map[string]any{
		"workflow_id": request.WorkflowID,
		"status":      request.Status,
		"error":       request.ErrorText,
	})
}

func (j *Journal) appendStateSnapshot(ctx context.Context, node NodeExecution, name string, state map[string]any, createdAt int64) (string, error) {
	eventID, err := generatedValue(j.newEventID, "event id")
	if err != nil {
		return "", err
	}
	err = j.appendEvent(ctx, Event{
		EventID:         eventID,
		RunID:           node.RunID,
		BranchID:        node.BranchID,
		NodeExecutionID: node.NodeExecutionID,
		Name:            name,
		CreatedAt:       createdAt,
	}, map[string]any{
		"node_id": node.NodeID,
		"state":   state,
	})
	if err != nil {
		return "", err
	}
	return eventID, nil
}

func (j *Journal) appendEvent(ctx context.Context, event Event, payload map[string]any) error {
	if len(payload) > 0 {
		raw, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		event.PayloadJSON = string(raw)
	}
	return j.port.AppendEvent(ctx, event)
}

func generatedValue(generator func() string, name string) (string, error) {
	if generator == nil {
		return "", fmt.Errorf("workflow journal: %s source is required", name)
	}
	value := generator()
	if value == "" {
		return "", fmt.Errorf("workflow journal: %s source returned empty value", name)
	}
	return value, nil
}

func currentTime(now func() int64) (int64, error) {
	if now == nil {
		return 0, fmt.Errorf("workflow journal: clock is required")
	}
	return now(), nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
