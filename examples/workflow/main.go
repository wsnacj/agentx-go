// workflow 展示显式图执行：Host注入validator、mapper、executor和identity。
package main

import (
	"context"
	"fmt"
	"os"

	workflow "github.com/wsnacj/agentx-go/runtime/workflow"
	workflowhostkit "github.com/wsnacj/agentx-go/runtime/workflow/hostkit"
)

type validator struct{}

func (validator) ValidateSpec(workflow.Spec) error     { return nil }
func (validator) ValidateNode(workflow.NodeSpec) error { return nil }

type mapper struct{}

func (mapper) MapNode(node workflow.NodeSpec, _ workflow.ExecutionMode) (workflowhostkit.MappedCall, error) {
	return workflowhostkit.MappedCall{Name: "echo", Arguments: map[string]any{"node": node.ID}}, nil
}

type executor struct{}

func (executor) Execute(_ context.Context, call workflowhostkit.Call) (string, error) {
	return fmt.Sprintf("executed:%s", call.Name), nil
}

func run(ctx context.Context) (workflowhostkit.Result, error) {
	runtime, err := workflowhostkit.New(workflowhostkit.Config{
		Validator: validator{}, Mapper: mapper{}, BasicExecutor: executor{},
		NewRunID:           func() string { return "run-workflow-example" },
		NewNodeExecutionID: func() string { return "nodeexec-workflow-example" },
		NowUnixMilli:       func() int64 { return 1 },
	})
	if err != nil {
		return workflowhostkit.Result{}, err
	}
	return runtime.Run(ctx, workflow.Spec{
		ID: "workflow-example", Version: "1", EntryNode: "echo",
		Nodes: []workflow.NodeSpec{{ID: "echo", Kind: workflow.NodeTool}},
	}, workflowhostkit.Inputs{})
}

func main() {
	result, err := run(context.Background())
	if err != nil || result.Execution.FinalStatus != "completed" {
		fmt.Fprintln(os.Stderr, "workflow example failed", result.Execution.FinalStatus, err)
		os.Exit(1)
	}
	fmt.Println(result.Execution.NodeOutput["echo"])
}
