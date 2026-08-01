package hostkit_test

import (
	"context"
	"testing"

	workflow "github.com/wsnacj/agentx-go/runtime/workflow"
	workflowhostkit "github.com/wsnacj/agentx-go/runtime/workflow/hostkit"
)

type validator struct{}

func (validator) ValidateSpec(workflow.Spec) error     { return nil }
func (validator) ValidateNode(workflow.NodeSpec) error { return nil }

type mapper struct{}

func (mapper) MapNode(workflow.NodeSpec, workflow.ExecutionMode) (workflowhostkit.MappedCall, error) {
	return workflowhostkit.MappedCall{Name: "echo"}, nil
}

type executor struct{}

func (executor) Execute(context.Context, workflowhostkit.Call) (string, error) {
	return "external-ok", nil
}

func TestExternalConsumerNeedsOnlyWorkflowAndHostKit(t *testing.T) {
	runtime, err := workflowhostkit.New(workflowhostkit.Config{
		Validator:          validator{},
		Mapper:             mapper{},
		BasicExecutor:      executor{},
		NewRunID:           func() string { return "run-external" },
		NewNodeExecutionID: func() string { return "nodeexec-external" },
		NowUnixMilli:       func() int64 { return 1 },
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
	}, workflowhostkit.Inputs{})
	if err != nil {
		t.Fatalf("Run(): %v", err)
	}
	if result.Execution.FinalStatus != "completed" ||
		result.Execution.NodeOutput["step"] != "external-ok" {
		t.Fatalf("result = %#v", result)
	}
}
