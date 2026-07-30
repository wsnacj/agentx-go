package transition_test

import (
	"testing"

	workflow "github.com/wsnacj/agentx-go/runtime/workflow"
	transition "github.com/wsnacj/agentx-go/runtime/workflow/transition"
)

func TestExternalPackageConsumer(t *testing.T) {
	machine := transition.New(transition.Plan{
		EntryNode: "collect",
		NodeIDs:   []string{"collect", "report"},
		Edges: []workflow.EdgeSpec{{
			From: "collect",
			To:   "report",
			On:   "success",
		}},
	})
	current, err := machine.Enter()
	if err != nil || current != "collect" {
		t.Fatalf("Enter() = %q, %v", current, err)
	}
	next, err := machine.Advance(transition.TriggerSuccess)
	if err != nil || next != "report" {
		t.Fatalf("Advance(success) = %q, %v", next, err)
	}
	if got := transition.NormalizeFinalStatus("", false); got != "completed" {
		t.Fatalf("NormalizeFinalStatus() = %q, want completed", got)
	}
}
