package runstore

import (
	"context"
	"errors"
)

var (
	ErrNotFound      = errors.New("agentx/runstore: not found")
	ErrAlreadyExists = errors.New("agentx/runstore: already exists")
)

type Run struct {
	RunID       string `json:"run_id"`
	CaseID      string `json:"case_id,omitempty"`
	WorkflowID  string `json:"workflow_id,omitempty"`
	WorkflowVer string `json:"workflow_ver,omitempty"`
	Status      string `json:"status,omitempty"`
	Attempt     int    `json:"attempt,omitempty"`
	ParentRunID string `json:"parent_run_id,omitempty"`
	RootRunID   string `json:"root_run_id,omitempty"`
	ContractID  string `json:"contract_id,omitempty"`
	StartedAt   int64  `json:"started_at,omitempty"`
	FinishedAt  int64  `json:"finished_at,omitempty"`
	Summary     string `json:"summary,omitempty"`
}

type NodeExecution struct {
	NodeExecID                string `json:"node_exec_id"`
	RunID                     string `json:"run_id"`
	BranchID                  string `json:"branch_id,omitempty"`
	ParentNodeExecID          string `json:"parent_node_exec_id,omitempty"`
	NodeID                    string `json:"node_id,omitempty"`
	Kind                      string `json:"kind,omitempty"`
	Status                    string `json:"status,omitempty"`
	Attempt                   int    `json:"attempt,omitempty"`
	InputStateRef             string `json:"input_state_ref,omitempty"`
	OutputStateRef            string `json:"output_state_ref,omitempty"`
	ExecutionContractID       string `json:"execution_contract_id,omitempty"`
	ExecutionContractDiffJSON string `json:"execution_contract_diff_json,omitempty"`
	TerminationJSON           string `json:"termination_json,omitempty"`
	DelegatedExecutionJSON    string `json:"delegated_execution_json,omitempty"`
	StartedAt                 int64  `json:"started_at,omitempty"`
	FinishedAt                int64  `json:"finished_at,omitempty"`
}

type Event struct {
	EventID     string `json:"event_id"`
	RunID       string `json:"run_id"`
	BranchID    string `json:"branch_id,omitempty"`
	NodeExecID  string `json:"node_exec_id,omitempty"`
	Name        string `json:"name,omitempty"`
	PayloadJSON string `json:"payload_json,omitempty"`
	CreatedAt   int64  `json:"created_at,omitempty"`
}

type Store interface {
	CreateRun(ctx context.Context, run Run) error
	UpdateRun(ctx context.Context, run Run) error
	GetRun(ctx context.Context, runID string) (Run, error)

	AppendEvent(ctx context.Context, event Event) error
	ListEvents(ctx context.Context, runID string, limit int) ([]Event, error)

	UpsertNodeExecution(ctx context.Context, node NodeExecution) error
	ListNodeExecutions(ctx context.Context, runID string) ([]NodeExecution, error)
}
