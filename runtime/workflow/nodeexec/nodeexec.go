// Package nodeexec 提供与执行 substrate 无关的 Workflow node execution
// coordination。
//
// 当前 package 处于 Experimental/private validation。它不拥有 lowering、
// concrete executor、RunStore、retry、provider、credential 或 Scene policy。
package nodeexec

import (
	"context"
	"fmt"

	workflow "github.com/wsnacj/agentx-go/runtime/workflow"
)

// Call is the portable tool-call projection passed to a basic executor.
type Call struct {
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"`
}

// Request contains the portable node execution input.
type Request struct {
	NodeExecutionID string                 `json:"node_execution_id,omitempty"`
	NodeID          string                 `json:"node_id,omitempty"`
	Kind            workflow.NodeKind      `json:"kind,omitempty"`
	ExecutionMode   workflow.ExecutionMode `json:"execution_mode,omitempty"`
	OriginalConfig  map[string]any         `json:"original_config,omitempty"`
	Spec            workflow.NodeSpec      `json:"spec,omitempty"`
	Call            Call                   `json:"call"`
}

// Termination contains portable node termination evidence.
type Termination struct {
	Kind            string `json:"kind,omitempty"`
	CheckpointStage string `json:"checkpoint_stage,omitempty"`
	CheckpointError string `json:"checkpoint_error,omitempty"`
	EventName       string `json:"event_name,omitempty"`
	EventStatus     string `json:"event_status,omitempty"`
	ReplyPersisted  bool   `json:"reply_persisted,omitempty"`
}

// DelegatedRound contains one portable delegated execution round.
type DelegatedRound struct {
	NodeExecID  string `json:"node_exec_id,omitempty"`
	Round       int    `json:"round,omitempty"`
	OutcomeKind string `json:"outcome_kind,omitempty"`
	StopReason  string `json:"stop_reason,omitempty"`
	ToolCalls   int    `json:"tool_calls,omitempty"`
	ToolRuns    int    `json:"tool_runs,omitempty"`
}

// DelegatedExecution contains portable delegated execution evidence.
type DelegatedExecution struct {
	Driver      string           `json:"driver,omitempty"`
	OutcomeKind string           `json:"outcome_kind,omitempty"`
	RoundCount  int              `json:"round_count,omitempty"`
	ToolCalls   int              `json:"tool_calls,omitempty"`
	Rounds      []DelegatedRound `json:"rounds,omitempty"`
}

// NodeExecutionProjection is the recursive portable projection for child
// executions returned by an outcome-aware executor.
type NodeExecutionProjection struct {
	NodeExecID            string                    `json:"node_exec_id,omitempty"`
	RunID                 string                    `json:"run_id,omitempty"`
	BranchID              string                    `json:"branch_id,omitempty"`
	ParentNodeExecID      string                    `json:"parent_node_exec_id,omitempty"`
	NodeID                string                    `json:"node_id,omitempty"`
	Kind                  string                    `json:"kind,omitempty"`
	Status                string                    `json:"status,omitempty"`
	Attempt               int                       `json:"attempt,omitempty"`
	InputStateRef         string                    `json:"input_state_ref,omitempty"`
	OutputStateRef        string                    `json:"output_state_ref,omitempty"`
	ExecutionContractID   string                    `json:"execution_contract_id,omitempty"`
	ExecutionContractDiff []string                  `json:"execution_contract_diff,omitempty"`
	Termination           *Termination              `json:"termination,omitempty"`
	DelegatedExecution    *DelegatedExecution       `json:"delegated_execution,omitempty"`
	ChildNodeExecutions   []NodeExecutionProjection `json:"child_node_executions,omitempty"`
	StartedAt             int64                     `json:"started_at,omitempty"`
	FinishedAt            int64                     `json:"finished_at,omitempty"`
}

// ChildNodeExecutionProjection is kept as a source-compatible name for the
// recursive portable projection.
//
// Deprecated: use NodeExecutionProjection.
type ChildNodeExecutionProjection = NodeExecutionProjection

// Outcome is the portable result of one node executor invocation.
type Outcome struct {
	Output                string                    `json:"output,omitempty"`
	FinalStatus           string                    `json:"final_status,omitempty"`
	StopReason            string                    `json:"stop_reason,omitempty"`
	ExecutionContractID   string                    `json:"execution_contract_id,omitempty"`
	ExecutionContractDiff []string                  `json:"execution_contract_diff,omitempty"`
	Termination           *Termination              `json:"termination,omitempty"`
	DelegatedExecution    *DelegatedExecution       `json:"delegated_execution,omitempty"`
	ChildNodeExecutions   []NodeExecutionProjection `json:"child_node_executions,omitempty"`
}

// BasicExecutor executes a portable call without node metadata.
type BasicExecutor interface {
	Execute(context.Context, Call) (string, error)
}

// NodeExecutor executes a node-aware portable request.
type NodeExecutor interface {
	ExecuteNode(context.Context, Request) (string, error)
}

// OutcomeExecutor executes a request and returns a rich portable outcome.
type OutcomeExecutor interface {
	ExecuteNodeWithOutcome(context.Context, Request) (Outcome, error)
}

// Dependencies contains host-owned execution capabilities.
type Dependencies struct {
	Basic       BasicExecutor
	Node        NodeExecutor
	Outcome     OutcomeExecutor
	BindContext func(context.Context, string, string) context.Context
}

// Coordinator selects exactly one executor capability for each request.
type Coordinator struct {
	basic       BasicExecutor
	node        NodeExecutor
	outcome     OutcomeExecutor
	bindContext func(context.Context, string, string) context.Context
}

// New constructs an immutable node execution coordinator.
func New(dependencies Dependencies) *Coordinator {
	return &Coordinator{
		basic:       dependencies.Basic,
		node:        dependencies.Node,
		outcome:     dependencies.Outcome,
		bindContext: dependencies.BindContext,
	}
}

// Execute binds the node context and invokes exactly one executor. Rich outcome
// capability takes precedence over node-aware and basic capabilities.
func (c *Coordinator) Execute(ctx context.Context, request Request) (Outcome, error) {
	if c == nil {
		return Outcome{}, fmt.Errorf("workflow nodeexec: coordinator is required")
	}
	if c.bindContext != nil {
		if bound := c.bindContext(ctx, request.NodeExecutionID, request.NodeID); bound != nil {
			ctx = bound
		}
	}
	if c.outcome != nil {
		return c.outcome.ExecuteNodeWithOutcome(ctx, request)
	}
	if c.node != nil {
		output, err := c.node.ExecuteNode(ctx, request)
		return Outcome{Output: output}, err
	}
	if c.basic == nil {
		return Outcome{}, fmt.Errorf("workflow nodeexec: basic executor is required")
	}
	output, err := c.basic.Execute(ctx, request.Call)
	return Outcome{Output: output}, err
}
