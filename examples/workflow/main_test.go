package main

import (
	"context"
	"testing"
)

func TestWorkflowExample(t *testing.T) {
	result, err := run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if result.Execution.RunID != "run-workflow-example" || result.Execution.FinalStatus != "completed" || result.Execution.NodeOutput["echo"] != "executed:echo" {
		t.Fatalf("execution=%#v", result.Execution)
	}
}
