package main

import "testing"

func TestFixedVersionProductShellConsumer(t *testing.T) {
	got, err := run()
	if err != nil {
		t.Fatalf("run(): %v", err)
	}
	const want = "agentx-productshell-ok:portable-research:research.lookup:collect-v1:portable-review:case-001:AgentX"
	if got != want {
		t.Fatalf("run() = %q, want %q", got, want)
	}
}

func TestFixedVersionTemporaryWorkflowPlanningConsumer(t *testing.T) {
	got, err := runTemporaryWorkflowPlanning()
	if err != nil {
		t.Fatalf("runTemporaryWorkflowPlanning(): %v", err)
	}
	const want = "agentx-productshell-planning-ok:temp_workflow_external_consumer:1:lookup:true"
	if got != want {
		t.Fatalf("runTemporaryWorkflowPlanning() = %q, want %q", got, want)
	}
}
