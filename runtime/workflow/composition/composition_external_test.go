package composition_test

import (
	"context"
	"testing"

	workflow "github.com/wsnacj/agentx-go/runtime/workflow"
	composition "github.com/wsnacj/agentx-go/runtime/workflow/composition"
	workflowjournal "github.com/wsnacj/agentx-go/runtime/workflow/journal"
	workflowlowering "github.com/wsnacj/agentx-go/runtime/workflow/lowering"
	workflownodeexec "github.com/wsnacj/agentx-go/runtime/workflow/nodeexec"
	workfloworchestration "github.com/wsnacj/agentx-go/runtime/workflow/orchestration"
)

func TestExternalConsumerCanConstructAndRun(t *testing.T) {
	runtime, err := composition.New(composition.Dependencies{
		Lowering: workflowlowering.Dependencies{
			Validator: acceptingValidator{},
			Mapper:    fixedMapper{},
		},
		Orchestration: workfloworchestration.Dependencies{
			Journal: workflowjournal.New(workflowjournal.Dependencies{
				NewRunID:     func() string { return "run-external" },
				NewEventID:   func() string { return "event-external" },
				NowUnixMilli: func() int64 { return 1 },
			}),
			NodeExecution:      fixedExecutor{},
			NewNodeExecutionID: func() string { return "nodeexec-external" },
			NowUnixMilli:       func() int64 { return 1 },
		},
	})
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	result, err := runtime.Run(context.Background(), workflow.Spec{
		ID:        "external-workflow",
		EntryNode: "step",
		Nodes: []workflow.NodeSpec{{
			ID:   "step",
			Kind: workflow.NodeTool,
		}},
	}, composition.Inputs{})
	if err != nil {
		t.Fatalf("Run(): %v", err)
	}
	if result.LoweringPlan.SpecID != "external-workflow" ||
		result.Execution.RunID != "run-external" ||
		result.Execution.FinalStatus != "completed" ||
		result.Execution.NodeOutput["step"] != "external-ok" {
		t.Fatalf("result = %#v", result)
	}
}

type acceptingValidator struct{}

func (acceptingValidator) ValidateSpec(workflow.Spec) error     { return nil }
func (acceptingValidator) ValidateNode(workflow.NodeSpec) error { return nil }

type fixedMapper struct{}

func (fixedMapper) MapNode(workflow.NodeSpec, workflow.ExecutionMode) (workflowlowering.MappedCall, error) {
	return workflowlowering.MappedCall{Name: "external_tool"}, nil
}

type fixedExecutor struct{}

func (fixedExecutor) Execute(context.Context, workflownodeexec.Request) (workflownodeexec.Outcome, error) {
	return workflownodeexec.Outcome{Output: "external-ok"}, nil
}
