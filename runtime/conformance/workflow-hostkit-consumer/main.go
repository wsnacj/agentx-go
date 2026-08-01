package main

import (
	"context"
	"fmt"

	workflow "github.com/wsnacj/agentx-go/runtime/workflow"
	workflowhostkit "github.com/wsnacj/agentx-go/runtime/workflow/hostkit"
)

type validator struct{}

func (validator) ValidateSpec(workflow.Spec) error     { return nil }
func (validator) ValidateNode(workflow.NodeSpec) error { return nil }

type mapper struct{}

func (mapper) MapNode(workflow.NodeSpec, workflow.ExecutionMode) (workflowhostkit.MappedCall, error) {
	return workflowhostkit.MappedCall{
		Name:      "echo",
		Arguments: map[string]any{"source": "fixed-version-consumer"},
	}, nil
}

type executor struct{}

func (executor) Execute(context.Context, workflowhostkit.Call) (string, error) {
	return "workflow-hostkit-conformance", nil
}

func run(ctx context.Context) (string, error) {
	runtime, err := workflowhostkit.New(workflowhostkit.Config{
		Validator:          validator{},
		Mapper:             mapper{},
		BasicExecutor:      executor{},
		NewRunID:           func() string { return "run-workflow-conformance" },
		NewNodeExecutionID: func() string { return "nodeexec-workflow-conformance" },
		NowUnixMilli:       func() int64 { return 1 },
	})
	if err != nil {
		return "", err
	}
	result, err := runtime.Run(ctx, workflow.Spec{
		ID:        "workflow-conformance",
		Version:   "1",
		EntryNode: "echo",
		Nodes: []workflow.NodeSpec{{
			ID:   "echo",
			Kind: workflow.NodeTool,
		}},
	}, workflowhostkit.Inputs{})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(
		"agentx-workflow-hostkit-ok:%s:%s:%s",
		result.Execution.RunID,
		result.Execution.FinalStatus,
		result.Execution.NodeOutput["echo"],
	), nil
}

func main() {
	output, err := run(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Println(output)
}
