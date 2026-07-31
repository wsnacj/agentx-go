package orchestration_test

import (
	"context"
	"testing"

	workflow "github.com/wsnacj/agentx-go/runtime/workflow"
	workflowjournal "github.com/wsnacj/agentx-go/runtime/workflow/journal"
	workflownodeexec "github.com/wsnacj/agentx-go/runtime/workflow/nodeexec"
	orchestration "github.com/wsnacj/agentx-go/runtime/workflow/orchestration"
)

func TestExternalPackageConsumer(t *testing.T) {
	result, err := orchestration.Run(context.Background(), orchestration.Plan{
		WorkflowID: "external-workflow",
		EntryNode:  "step",
		NodeIDs:    []string{"step"},
		Nodes: map[string]orchestration.PlannedNode{
			"step": {
				Spec: workflow.NodeSpec{ID: "step", Kind: workflow.NodeTool},
				Kind: workflow.NodeTool,
				Call: workflownodeexec.Call{Name: "echo", Arguments: `{}`},
			},
		},
	}, orchestration.Inputs{RunID: "external-run"}, orchestration.Dependencies{
		Journal:            workflowjournal.New(workflowjournal.Dependencies{}),
		NodeExecution:      externalExecution{},
		NewNodeExecutionID: func() string { return "nodeexec-1" },
		NowUnixMilli:       func() int64 { return 1 },
	})
	if err != nil {
		t.Fatalf("Run(): %v", err)
	}
	if result.FinalStatus != "completed" || result.NodeOutput["step"] != "ready" {
		t.Fatalf("result = %#v", result)
	}
}

type externalExecution struct{}

func (externalExecution) Execute(context.Context, workflownodeexec.Request) (workflownodeexec.Outcome, error) {
	return workflownodeexec.Outcome{Output: "ready"}, nil
}
