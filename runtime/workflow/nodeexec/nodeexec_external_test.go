package nodeexec_test

import (
	"context"
	"testing"

	nodeexec "github.com/wsnacj/agentx-go/runtime/workflow/nodeexec"
)

func TestExternalPackageConsumer(t *testing.T) {
	executor := externalExecutor{}
	coordinator := nodeexec.New(nodeexec.Dependencies{
		Basic:   executor,
		Node:    executor,
		Outcome: executor,
	})
	outcome, err := coordinator.Execute(context.Background(), nodeexec.Request{
		NodeExecutionID: "nodeexec-1",
		NodeID:          "collect",
		Call: nodeexec.Call{
			Name:      "public_source",
			Arguments: `{"query":"risk"}`,
		},
	})
	if err != nil {
		t.Fatalf("Execute(): %v", err)
	}
	if outcome.Output != "ready" || outcome.FinalStatus != "completed" {
		t.Fatalf("outcome = %#v", outcome)
	}
}

type externalExecutor struct{}

func (externalExecutor) Execute(context.Context, nodeexec.Call) (string, error) {
	return "basic", nil
}

func (externalExecutor) ExecuteNode(context.Context, nodeexec.Request) (string, error) {
	return "node", nil
}

func (externalExecutor) ExecuteNodeWithOutcome(context.Context, nodeexec.Request) (nodeexec.Outcome, error) {
	return nodeexec.Outcome{Output: "ready", FinalStatus: "completed"}, nil
}
