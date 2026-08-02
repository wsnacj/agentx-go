package main

import (
	"context"
	"testing"
)

func TestExternalStyleStandardConstructionPaths(t *testing.T) {
	modelTool, err := runModelToolProbe(context.Background())
	if err != nil {
		t.Fatalf("runModelToolProbe() error = %v", err)
	}
	if want := "run-model-tool:completed:model-tool-ok"; modelTool != want {
		t.Fatalf("runModelToolProbe() = %q, want %q", modelTool, want)
	}

	workflowResult, err := runWorkflowProbe(context.Background())
	if err != nil {
		t.Fatalf("runWorkflowProbe() error = %v", err)
	}
	if want := "run-workflow:completed:workflow-ok"; workflowResult != want {
		t.Fatalf("runWorkflowProbe() = %q, want %q", workflowResult, want)
	}
}

func TestExternalStyleUnifiedConsumer(t *testing.T) {
	output, err := runAllProbes(context.Background())
	if err != nil {
		t.Fatalf("runAllProbes() error = %v", err)
	}
	const want = "agentx-core-developer-preview-ok:model-tool=run-model-tool:completed:model-tool-ok:workflow=run-workflow:completed:workflow-ok"
	if output != want {
		t.Fatalf("runAllProbes() = %q, want %q", output, want)
	}
}
