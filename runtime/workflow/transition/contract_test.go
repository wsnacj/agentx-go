package transition

import (
	"strings"
	"testing"

	workflow "github.com/wsnacj/agentx-go/runtime/workflow"
)

func TestMachineTraversesExactIdentifiersAndDefaultSuccess(t *testing.T) {
	machine := New(Plan{
		EntryNode: " prepare ",
		NodeIDs:   []string{" prepare ", "finish"},
		Edges: []workflow.EdgeSpec{
			{From: "prepare", To: "wrong", On: "success"},
			{From: " prepare ", To: "finish"},
		},
	})
	if current, err := machine.Enter(); err != nil || current != " prepare " {
		t.Fatalf("Enter() = %q, %v", current, err)
	}
	if next, err := machine.Advance(TriggerSuccess); err != nil || next != "finish" {
		t.Fatalf("Advance(success) = %q, %v", next, err)
	}
	if current, err := machine.Enter(); err != nil || current != "finish" {
		t.Fatalf("second Enter() = %q, %v", current, err)
	}
	if next, err := machine.Advance(TriggerSuccess); err != nil || next != "" {
		t.Fatalf("terminal Advance(success) = %q, %v", next, err)
	}
	if current, err := machine.Enter(); err != nil || current != "" {
		t.Fatalf("terminal Enter() = %q, %v", current, err)
	}
}

func TestMachineRoutesFailureAndAlwaysEdges(t *testing.T) {
	tests := []struct {
		name    string
		trigger Trigger
		edgeOn  string
	}{
		{name: "failure", trigger: TriggerFailure, edgeOn: "failure"},
		{name: "always after success", trigger: TriggerSuccess, edgeOn: "always"},
		{name: "always after failure", trigger: TriggerFailure, edgeOn: "always"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			machine := New(Plan{
				EntryNode: "start",
				NodeIDs:   []string{"start", "finish"},
				Edges:     []workflow.EdgeSpec{{From: "start", To: "finish", On: tt.edgeOn}},
			})
			if _, err := machine.Enter(); err != nil {
				t.Fatalf("Enter(): %v", err)
			}
			if next, err := machine.Advance(tt.trigger); err != nil || next != "finish" {
				t.Fatalf("Advance(%s) = %q, %v", tt.trigger, next, err)
			}
		})
	}
}

func TestMachineRejectsMultipleMatchingEdgesWithoutAdvancing(t *testing.T) {
	machine := New(Plan{
		EntryNode: "start",
		NodeIDs:   []string{"start", "one", "two"},
		Edges: []workflow.EdgeSpec{
			{From: "start", To: "one", On: "success"},
			{From: "start", To: "two", On: "always"},
		},
	})
	if _, err := machine.Enter(); err != nil {
		t.Fatalf("Enter(): %v", err)
	}
	_, err := machine.Advance(TriggerSuccess)
	if err == nil {
		t.Fatal("expected multiple-edge error")
	}
	want := `workflow: node "start" has multiple outgoing success edges; inline executor only supports a single path`
	if err.Error() != want {
		t.Fatalf("error = %q, want %q", err, want)
	}
}

func TestMachineRejectsMissingNodeAndCycleWithExactErrors(t *testing.T) {
	missing := New(Plan{EntryNode: "missing"})
	if _, err := missing.Enter(); err == nil || err.Error() != `workflow: lowered node "missing" missing` {
		t.Fatalf("missing error = %v", err)
	}

	cycle := New(Plan{
		EntryNode: "loop",
		NodeIDs:   []string{"loop"},
		Edges:     []workflow.EdgeSpec{{From: "loop", To: "loop"}},
	})
	if _, err := cycle.Enter(); err != nil {
		t.Fatalf("first Enter(): %v", err)
	}
	if _, err := cycle.Advance(TriggerSuccess); err != nil {
		t.Fatalf("Advance(): %v", err)
	}
	if _, err := cycle.Enter(); err == nil || !strings.Contains(err.Error(), `detected cycle at node "loop"`) {
		t.Fatalf("cycle error = %v", err)
	}
}

func TestNewCopiesPlanAndNormalizeFinalStatusPreservesRules(t *testing.T) {
	nodeIDs := []string{"start", "finish"}
	edges := []workflow.EdgeSpec{{From: "start", To: "finish"}}
	machine := New(Plan{EntryNode: "start", NodeIDs: nodeIDs, Edges: edges})
	nodeIDs[1] = "mutated"
	edges[0].To = "mutated"

	if _, err := machine.Enter(); err != nil {
		t.Fatalf("Enter(): %v", err)
	}
	if next, err := machine.Advance(TriggerSuccess); err != nil || next != "finish" {
		t.Fatalf("Advance(success) = %q, %v", next, err)
	}

	tests := []struct {
		status string
		failed bool
		want   string
	}{
		{status: "completed", want: "completed"},
		{status: "failed", want: "failed"},
		{status: "incomplete", want: "incomplete"},
		{status: " completed ", want: "completed"},
		{status: "unknown", want: "completed"},
		{status: "incomplete", failed: true, want: "failed"},
		{status: " failed ", failed: true, want: "failed"},
	}
	for _, tt := range tests {
		if got := NormalizeFinalStatus(tt.status, tt.failed); got != tt.want {
			t.Errorf("NormalizeFinalStatus(%q, %v) = %q, want %q", tt.status, tt.failed, got, tt.want)
		}
	}
}
