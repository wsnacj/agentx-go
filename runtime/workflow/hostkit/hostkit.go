// Package hostkit 组合 canonical Workflow lowering、journal、node execution
// coordination 和 orchestration，为不依赖特定 Runner 的宿主提供最小构造入口。
//
// 本package进入v0.2.0 Developer Preview核心兼容候选面。它不提供
// validation/mapping policy、executor、RunStore backend、provider、credential
// 或 Scene。
package hostkit

import (
	"context"
	"fmt"

	workflow "github.com/wsnacj/agentx-go/runtime/workflow"
	workflowcomposition "github.com/wsnacj/agentx-go/runtime/workflow/composition"
	workflowjournal "github.com/wsnacj/agentx-go/runtime/workflow/journal"
	workflowlowering "github.com/wsnacj/agentx-go/runtime/workflow/lowering"
	workflownodeexec "github.com/wsnacj/agentx-go/runtime/workflow/nodeexec"
	workfloworchestration "github.com/wsnacj/agentx-go/runtime/workflow/orchestration"
)

// Validator validates complete Specs and individual Nodes before mapping.
type Validator interface {
	ValidateSpec(workflow.Spec) error
	ValidateNode(workflow.NodeSpec) error
}

// Mapper maps one validated node to a portable call.
type Mapper interface {
	MapNode(workflow.NodeSpec, workflow.ExecutionMode) (MappedCall, error)
}

// MappedCall is the portable pre-JSON mapping result.
type MappedCall = workflowlowering.MappedCall

// Call is the portable basic executor input.
type Call = workflownodeexec.Call

// NodeRequest is the portable node-aware executor input.
type NodeRequest = workflownodeexec.Request

// NodeOutcome is the portable rich node result.
type NodeOutcome = workflownodeexec.Outcome

// BasicExecutor executes a mapped call without node metadata.
type BasicExecutor = workflownodeexec.BasicExecutor

// NodeExecutor executes one node-aware request.
type NodeExecutor = workflownodeexec.NodeExecutor

// OutcomeExecutor executes one request and returns a rich portable outcome.
type OutcomeExecutor = workflownodeexec.OutcomeExecutor

// JournalRun is the portable durable Run record.
type JournalRun = workflowjournal.Run

// JournalNodeExecution is the portable durable node record.
type JournalNodeExecution = workflowjournal.NodeExecution

// JournalEvent is the portable durable event record.
type JournalEvent = workflowjournal.Event

// JournalPort is the optional durable storage boundary.
type JournalPort = workflowjournal.Port

// Inputs contains the optional Workflow identity and portable run inputs.
type Inputs = workflowcomposition.Inputs

// RunInputs contains portable run identity and binding roots.
type RunInputs = workfloworchestration.Inputs

// Result preserves both the lowering plan and partial or complete execution.
type Result = workflowcomposition.Result

// Config contains only host-selected capabilities. It never installs a
// provider, backend, identity policy, clock, or product default.
type Config struct {
	Validator Validator
	Mapper    Mapper

	BasicExecutor   BasicExecutor
	NodeExecutor    NodeExecutor
	OutcomeExecutor OutcomeExecutor

	JournalPort JournalPort

	NewRunID           func() string
	NewEventID         func() string
	NewNodeExecutionID func() string
	NowUnixMilli       func() int64

	BindNodeExecutionContext func(context.Context, string, string) context.Context
	ProjectError             func(error) string
}

// Runtime is an immutable host composition of the canonical Workflow owners.
type Runtime struct {
	composition *workflowcomposition.Runtime
}

// New validates host capabilities and composes the canonical Workflow Runtime.
func New(config Config) (*Runtime, error) {
	switch {
	case config.Validator == nil:
		return nil, fmt.Errorf("agentx workflow host kit: validator is required")
	case config.Mapper == nil:
		return nil, fmt.Errorf("agentx workflow host kit: mapper is required")
	case config.BasicExecutor == nil && config.NodeExecutor == nil && config.OutcomeExecutor == nil:
		return nil, fmt.Errorf("agentx workflow host kit: executor is required")
	case config.NewRunID == nil:
		return nil, fmt.Errorf("agentx workflow host kit: run id generator is required")
	case config.JournalPort != nil && config.NewEventID == nil:
		return nil, fmt.Errorf("agentx workflow host kit: event id generator is required when journal port is configured")
	case config.NewNodeExecutionID == nil:
		return nil, fmt.Errorf("agentx workflow host kit: node execution id generator is required")
	case config.NowUnixMilli == nil:
		return nil, fmt.Errorf("agentx workflow host kit: clock is required")
	}

	journal := workflowjournal.New(workflowjournal.Dependencies{
		Port:         config.JournalPort,
		NewRunID:     config.NewRunID,
		NewEventID:   config.NewEventID,
		NowUnixMilli: config.NowUnixMilli,
	})
	nodeExecution := workflownodeexec.New(workflownodeexec.Dependencies{
		Basic:       config.BasicExecutor,
		Node:        config.NodeExecutor,
		Outcome:     config.OutcomeExecutor,
		BindContext: config.BindNodeExecutionContext,
	})
	composition, err := workflowcomposition.New(workflowcomposition.Dependencies{
		Lowering: workflowlowering.Dependencies{
			Validator: config.Validator,
			Mapper:    config.Mapper,
		},
		Orchestration: workfloworchestration.Dependencies{
			Journal:            journal,
			NodeExecution:      nodeExecution,
			NewNodeExecutionID: config.NewNodeExecutionID,
			NowUnixMilli:       config.NowUnixMilli,
			ProjectError:       config.ProjectError,
		},
	})
	if err != nil {
		return nil, err
	}
	return &Runtime{composition: composition}, nil
}

// Run delegates validation, lowering and execution to the canonical
// composition and preserves partial results and error identity.
func (r *Runtime) Run(ctx context.Context, spec workflow.Spec, inputs Inputs) (Result, error) {
	if r == nil || r.composition == nil {
		return Result{}, fmt.Errorf("agentx workflow host kit: runtime is required")
	}
	return r.composition.Run(ctx, spec, inputs)
}
