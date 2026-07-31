// Package composition 提供 Workflow lowering 与 run orchestration 的
// substrate-neutral construction 和组合机制。
//
// 当前 package 处于 Experimental/private validation。具体 validation/mapping
// policy、executor、RunStore、identity、clock 与 error display policy 必须由
// host 显式注入。
package composition

import (
	"context"
	"fmt"

	workflow "github.com/wsnacj/agentx-go/runtime/workflow"
	workflowlowering "github.com/wsnacj/agentx-go/runtime/workflow/lowering"
	workfloworchestration "github.com/wsnacj/agentx-go/runtime/workflow/orchestration"
)

// Dependencies contains the canonical lowering and orchestration dependencies
// validated and retained by Runtime.
type Dependencies struct {
	Lowering      workflowlowering.Dependencies
	Orchestration workfloworchestration.Dependencies
}

// Inputs contains the optional runtime Workflow identity and portable run
// inputs.
type Inputs struct {
	WorkflowID string
	Run        workfloworchestration.Inputs
}

// Result preserves both the lowering plan and partial or complete execution
// result.
type Result struct {
	LoweringPlan workflowlowering.Plan
	Execution    workfloworchestration.Result
}

// Runtime is an immutable composition of canonical Workflow dependencies.
type Runtime struct {
	dependencies Dependencies
}

// New validates required dependencies and constructs a Workflow Runtime
// composition.
func New(dependencies Dependencies) (*Runtime, error) {
	switch {
	case dependencies.Lowering.Validator == nil:
		return nil, fmt.Errorf("workflow composition: lowering validator is required")
	case dependencies.Lowering.Mapper == nil:
		return nil, fmt.Errorf("workflow composition: lowering mapper is required")
	case dependencies.Orchestration.Journal == nil:
		return nil, fmt.Errorf("workflow composition: journal is required")
	case dependencies.Orchestration.NodeExecution == nil:
		return nil, fmt.Errorf("workflow composition: node execution is required")
	case dependencies.Orchestration.NewNodeExecutionID == nil:
		return nil, fmt.Errorf("workflow composition: node execution id generator is required")
	case dependencies.Orchestration.NowUnixMilli == nil:
		return nil, fmt.Errorf("workflow composition: clock is required")
	default:
		return &Runtime{dependencies: dependencies}, nil
	}
}

// Run lowers the Spec, resolves the runtime Workflow identity, and executes the
// canonical orchestration loop.
func (r *Runtime) Run(ctx context.Context, spec workflow.Spec, inputs Inputs) (Result, error) {
	if r == nil {
		return Result{}, fmt.Errorf("workflow composition: runtime is required")
	}
	plan, err := workflowlowering.LowerSpec(spec, r.dependencies.Lowering)
	if err != nil {
		return Result{}, err
	}
	workflowID := inputs.WorkflowID
	if workflowID == "" {
		workflowID = plan.SpecID
	}
	execution, err := workfloworchestration.Run(
		ctx,
		plan.OrchestrationPlan(workflowID),
		inputs.Run,
		r.dependencies.Orchestration,
	)
	return Result{
		LoweringPlan: plan,
		Execution:    execution,
	}, err
}
