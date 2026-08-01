package runstore

import (
	"os"
	"strings"
	"testing"
)

func TestNodeExecutionProjectionCanonicalOwner(t *testing.T) {
	source, err := os.ReadFile("node_execution_projection.go")
	if err != nil {
		t.Fatalf("read projection source: %v", err)
	}
	text := string(source)
	if !strings.Contains(text, `"github.com/wsnacj/agentx-go/runtime/workflow/nodeexec"`) {
		t.Fatal("runstore projection compatibility names must consume canonical nodeexec")
	}
	for _, duplicate := range []string{
		"type NodeExecutionProjection struct",
		"type NodeTerminationProjection struct",
		"type NodeDelegatedExecutionProjection struct",
		"type NodeDelegatedRoundProjection struct",
	} {
		if strings.Contains(text, duplicate) {
			t.Errorf("runstore restores duplicate portable contract %q", duplicate)
		}
	}
}

func TestNodeDelegatedExecutionProjectionFromJSON(t *testing.T) {
	projection := NodeDelegatedExecutionProjectionFromJSON(` {
		"driver":" open_tool_loop ",
		"outcome_kind":" terminated ",
		"round_count":2,
		"tool_calls":1,
		"rounds":[
			{"node_exec_id":" nodeexec-round-1 ","round":1,"outcome_kind":"continue","tool_calls":1,"tool_runs":1},
			{"node_exec_id":"nodeexec-round-2","round":2,"outcome_kind":"terminated","stop_reason":" max_rounds ","tool_calls":0,"tool_runs":0}
		]
	}`)
	if projection == nil {
		t.Fatalf("expected projection")
	}
	if projection.Driver != " open_tool_loop " || projection.OutcomeKind != " terminated " || projection.RoundCount != 2 {
		t.Fatalf("unexpected projection: %#v", projection)
	}
	if len(projection.Rounds) != 2 || projection.Rounds[1].StopReason != " max_rounds " || projection.Rounds[0].NodeExecID != " nodeexec-round-1 " {
		t.Fatalf("unexpected projection rounds: %#v", projection)
	}
}

func TestCloneNodeDelegatedExecutionProjection(t *testing.T) {
	in := &NodeDelegatedExecutionProjection{
		Driver:      " open_tool_loop ",
		OutcomeKind: " completed ",
		RoundCount:  1,
		Rounds: []NodeDelegatedRoundProjection{
			{NodeExecID: " nodeexec-round-1 ", Round: 1, OutcomeKind: " completed ", StopReason: " "},
		},
	}
	out := CloneNodeDelegatedExecutionProjection(in)
	if out == nil {
		t.Fatalf("expected clone")
	}
	if out == in || &out.Rounds[0] == &in.Rounds[0] {
		t.Fatalf("expected deep clone")
	}
	if out.Driver != " open_tool_loop " || out.OutcomeKind != " completed " || out.Rounds[0].OutcomeKind != " completed " || out.Rounds[0].NodeExecID != " nodeexec-round-1 " {
		t.Fatalf("expected raw-preserve clone, got %#v", out)
	}
}

func TestNodeDelegatedExecutionProjectionFromJSON_EmptyProjectionReturnsNil(t *testing.T) {
	projection := NodeDelegatedExecutionProjectionFromJSON(` {
		"driver":" ",
		"outcome_kind":" ",
		"round_count":0,
		"tool_calls":0,
		"rounds":[{}]
	}`)
	if projection != nil {
		t.Fatalf("expected empty delegated execution projection to normalize to nil, got %#v", projection)
	}
}

func TestNodeTerminationProjectionFromJSONPreservesRawFields(t *testing.T) {
	projection := NodeTerminationProjectionFromJSON(`{
		"kind":" max_rounds ",
		"checkpoint_stage":" checkpoint_stage ",
		"checkpoint_error":" checkpoint_error ",
		"event_name":" tool.max_rounds ",
		"event_status":" error "
	}`)
	if projection == nil {
		t.Fatalf("expected termination projection")
	}
	if projection.Kind != " max_rounds " ||
		projection.CheckpointStage != " checkpoint_stage " ||
		projection.CheckpointError != " checkpoint_error " ||
		projection.EventName != " tool.max_rounds " ||
		projection.EventStatus != " error " {
		t.Fatalf("expected raw termination fields, got %#v", projection)
	}
}

func TestCloneNodeTerminationProjectionPreservesRawFields(t *testing.T) {
	out := CloneNodeTerminationProjection(&NodeTerminationProjection{
		Kind:            " max_rounds ",
		CheckpointStage: " checkpoint_stage ",
		CheckpointError: " checkpoint_error ",
		EventName:       " tool.max_rounds ",
		EventStatus:     " error ",
	})
	if out == nil {
		t.Fatalf("expected clone")
	}
	if out.Kind != " max_rounds " ||
		out.CheckpointStage != " checkpoint_stage " ||
		out.CheckpointError != " checkpoint_error " ||
		out.EventName != " tool.max_rounds " ||
		out.EventStatus != " error " {
		t.Fatalf("expected raw-preserve termination clone, got %#v", out)
	}
}

func TestCloneNodeDelegatedExecutionProjection_EmptyProjectionReturnsNil(t *testing.T) {
	out := CloneNodeDelegatedExecutionProjection(&NodeDelegatedExecutionProjection{
		Driver:      " ",
		OutcomeKind: " ",
		Rounds:      []NodeDelegatedRoundProjection{{}},
	})
	if out != nil {
		t.Fatalf("expected empty delegated execution projection clone to normalize to nil, got %#v", out)
	}
}

func TestNodeDelegatedExecutionLastStopReason(t *testing.T) {
	projection := &NodeDelegatedExecutionProjection{
		Rounds: []NodeDelegatedRoundProjection{
			{Round: 1, OutcomeKind: "continue"},
			{Round: 2, OutcomeKind: "terminated", StopReason: "max_rounds"},
		},
	}
	if got := NodeDelegatedExecutionLastStopReason(projection); got != "max_rounds" {
		t.Fatalf("expected last stop reason max_rounds, got %q", got)
	}
}

func TestNodeDelegatedExecutionLastStopReasonPreservesRawStopReason(t *testing.T) {
	projection := &NodeDelegatedExecutionProjection{
		Rounds: []NodeDelegatedRoundProjection{
			{Round: 1, OutcomeKind: "continue"},
			{Round: 2, OutcomeKind: "terminated", StopReason: " max_rounds "},
		},
	}
	if got := NodeDelegatedExecutionLastStopReason(projection); got != " max_rounds " {
		t.Fatalf("expected raw stop reason, got %q", got)
	}
}

func TestDelegatedOutcomeKindFromRoundsPreservesRawMatchedOutcome(t *testing.T) {
	rounds := []NodeDelegatedRoundProjection{
		{Round: 1, OutcomeKind: " continue "},
		{Round: 2, OutcomeKind: " terminated "},
	}
	if got := delegatedOutcomeKindFromRounds(rounds); got != " terminated " {
		t.Fatalf("expected raw matched outcome kind, got %q", got)
	}
}

func TestNodeDelegatedExecutionProjectionFromChildNodeExecutions(t *testing.T) {
	children := []NodeExecutionProjection{
		{
			NodeExecID: "node-round-1",
			NodeID:     "main/round:1",
			Status:     "completed",
			StartedAt:  100,
			DelegatedExecution: &NodeDelegatedExecutionProjection{
				Driver:      "open_tool_loop",
				OutcomeKind: "completed",
				ToolCalls:   1,
				Rounds: []NodeDelegatedRoundProjection{{
					NodeExecID:  "node-round-1",
					Round:       1,
					OutcomeKind: "completed",
					ToolCalls:   1,
				}},
			},
		},
		{
			NodeExecID: "node-round-3",
			NodeID:     "main/round:3",
			Status:     "incomplete",
			StartedAt:  300,
			Termination: &NodeTerminationProjection{
				Kind: "max_rounds",
			},
			DelegatedExecution: &NodeDelegatedExecutionProjection{
				Driver:      "open_tool_loop",
				OutcomeKind: "terminated",
				ToolCalls:   1,
				Rounds: []NodeDelegatedRoundProjection{{
					NodeExecID:  "node-round-3",
					Round:       3,
					OutcomeKind: "terminated",
					StopReason:  "max_rounds",
					ToolCalls:   1,
				}},
			},
		},
	}

	projection := NodeDelegatedExecutionProjectionFromChildNodeExecutions(children)
	if projection == nil {
		t.Fatalf("expected delegated execution projection from child nodes")
	}
	if projection.Driver != "open_tool_loop" ||
		projection.OutcomeKind != "terminated" ||
		projection.RoundCount != 3 ||
		projection.ToolCalls != 2 {
		t.Fatalf("unexpected delegated execution projection from child nodes: %#v", projection)
	}
	if len(projection.Rounds) != 2 ||
		projection.Rounds[0].NodeExecID != "node-round-1" ||
		projection.Rounds[1].NodeExecID != "node-round-3" ||
		projection.Rounds[1].StopReason != "max_rounds" {
		t.Fatalf("unexpected delegated execution child rounds: %#v", projection.Rounds)
	}
}

func TestNodeDelegatedExecutionProjectionFromChildNodeExecutionsPreservesRawFields(t *testing.T) {
	children := []NodeExecutionProjection{
		{
			NodeExecID: " node-round-1 ",
			NodeID:     "main/round:1",
			Status:     "completed",
			StartedAt:  100,
			DelegatedExecution: &NodeDelegatedExecutionProjection{
				Driver:      " open_tool_loop ",
				OutcomeKind: " completed ",
				ToolCalls:   1,
				Rounds: []NodeDelegatedRoundProjection{{
					NodeExecID:  " delegated-round-1 ",
					Round:       1,
					OutcomeKind: " completed ",
					ToolCalls:   1,
				}},
			},
		},
		{
			NodeExecID: " node-round-2 ",
			NodeID:     "main/round:2",
			Status:     "incomplete",
			StartedAt:  200,
			DelegatedExecution: &NodeDelegatedExecutionProjection{
				Driver:      " open_tool_loop ",
				OutcomeKind: " terminated ",
				ToolCalls:   1,
				Rounds: []NodeDelegatedRoundProjection{{
					NodeExecID:  " delegated-round-2 ",
					Round:       2,
					OutcomeKind: " terminated ",
					StopReason:  " max_rounds ",
					ToolCalls:   1,
				}},
			},
		},
	}

	projection := NodeDelegatedExecutionProjectionFromChildNodeExecutions(children)
	if projection == nil {
		t.Fatalf("expected delegated execution projection from child nodes")
	}
	if projection.Driver != " open_tool_loop " {
		t.Fatalf("expected raw driver, got %#v", projection)
	}
	if projection.OutcomeKind != " terminated " {
		t.Fatalf("expected raw delegated outcome kind, got %#v", projection)
	}
	if len(projection.Rounds) != 2 ||
		projection.Rounds[0].NodeExecID != " delegated-round-1 " ||
		projection.Rounds[0].OutcomeKind != " completed " ||
		projection.Rounds[1].NodeExecID != " delegated-round-2 " ||
		projection.Rounds[1].OutcomeKind != " terminated " ||
		projection.Rounds[1].StopReason != " max_rounds " {
		t.Fatalf("expected raw delegated rounds, got %#v", projection.Rounds)
	}
}

func TestNodeExecutionProjectionFromStoredFields(t *testing.T) {
	node := NodeExecution{
		NodeExecID:                "node-1",
		RunID:                     "run-1",
		ParentNodeExecID:          "node-root",
		NodeID:                    "capture_evidence",
		Kind:                      "tool",
		Status:                    "incomplete",
		ExecutionContractID:       "contract-node",
		ExecutionContractDiffJSON: `{"changed_fields":["visibility","evidence"]}`,
		TerminationJSON:           `{"kind":"max_rounds","checkpoint_stage":"max_rounds_break","event_name":"tool.max_rounds"}`,
		DelegatedExecutionJSON:    `{"driver":"open_tool_loop","round_count":2,"rounds":[{"node_exec_id":"round-node-2","round":2,"stop_reason":"max_rounds"}]}`,
	}
	projection := node.Projection()
	if projection == nil {
		t.Fatalf("expected node execution projection")
	}
	if projection.ExecutionContractID != "contract-node" {
		t.Fatalf("unexpected execution contract id: %#v", projection)
	}
	if projection.ParentNodeExecID != "node-root" {
		t.Fatalf("unexpected parent node exec id: %#v", projection)
	}
	if len(projection.ExecutionContractDiff) != 2 || projection.ExecutionContractDiff[0] != "visibility" {
		t.Fatalf("unexpected execution contract diff: %#v", projection)
	}
	if projection.Termination == nil || projection.Termination.Kind != "max_rounds" || projection.Termination.EventName != "tool.max_rounds" {
		t.Fatalf("unexpected termination projection: %#v", projection)
	}
	if projection.DelegatedExecution == nil || projection.DelegatedExecution.RoundCount != 2 {
		t.Fatalf("unexpected delegated execution projection: %#v", projection)
	}
	if len(projection.DelegatedExecution.Rounds) != 1 || projection.DelegatedExecution.Rounds[0].NodeExecID != "round-node-2" {
		t.Fatalf("unexpected delegated execution round projection: %#v", projection.DelegatedExecution)
	}
}

func TestNodeExecutionProjectionFromStoredFieldsPreservesRawCopiedFields(t *testing.T) {
	node := NodeExecution{
		NodeExecID:                " node-1 ",
		RunID:                     " run-1 ",
		BranchID:                  " branch-1 ",
		ParentNodeExecID:          " node-root ",
		NodeID:                    " capture_evidence ",
		Kind:                      " tool ",
		Status:                    " incomplete ",
		InputStateRef:             " input-ref ",
		OutputStateRef:            " output-ref ",
		ExecutionContractID:       " contract-node ",
		ExecutionContractDiffJSON: `{"changed_fields":[" visibility "," evidence "]}`,
		TerminationJSON:           `{"kind":" max_rounds "}`,
		DelegatedExecutionJSON:    `{"driver":" open_tool_loop ","rounds":[{"node_exec_id":" round-node-1 ","round":1,"outcome_kind":" terminated ","stop_reason":" max_rounds "}]} `,
	}
	projection := node.Projection()
	if projection == nil {
		t.Fatalf("expected node execution projection")
	}
	if projection.NodeExecID != " node-1 " ||
		projection.RunID != " run-1 " ||
		projection.BranchID != " branch-1 " ||
		projection.ParentNodeExecID != " node-root " ||
		projection.NodeID != " capture_evidence " ||
		projection.Kind != " tool " ||
		projection.Status != " incomplete " ||
		projection.InputStateRef != " input-ref " ||
		projection.OutputStateRef != " output-ref " ||
		projection.ExecutionContractID != " contract-node " {
		t.Fatalf("expected raw copied fields, got %#v", projection)
	}
	if len(projection.ExecutionContractDiff) != 2 ||
		projection.ExecutionContractDiff[0] != " visibility " ||
		projection.ExecutionContractDiff[1] != " evidence " {
		t.Fatalf("expected raw contract diff fields, got %#v", projection.ExecutionContractDiff)
	}
	if projection.Termination == nil || projection.Termination.Kind != " max_rounds " {
		t.Fatalf("expected raw termination kind, got %#v", projection.Termination)
	}
	if projection.DelegatedExecution == nil ||
		projection.DelegatedExecution.Driver != " open_tool_loop " ||
		len(projection.DelegatedExecution.Rounds) != 1 ||
		projection.DelegatedExecution.Rounds[0].NodeExecID != " round-node-1 " ||
		projection.DelegatedExecution.Rounds[0].OutcomeKind != " terminated " ||
		projection.DelegatedExecution.Rounds[0].StopReason != " max_rounds " {
		t.Fatalf("expected raw delegated execution fields, got %#v", projection.DelegatedExecution)
	}
}

func TestNodeExecutionProjectionOutcomeKindPreservesRawDelegatedOutcome(t *testing.T) {
	item := NodeExecutionProjection{
		Status: "incomplete",
		DelegatedExecution: &NodeDelegatedExecutionProjection{
			OutcomeKind: " terminated ",
			Rounds: []NodeDelegatedRoundProjection{
				{Round: 1, OutcomeKind: " continue "},
				{Round: 2, OutcomeKind: " terminated "},
			},
		},
	}
	if got := nodeExecutionProjectionOutcomeKind(item); got != " terminated " {
		t.Fatalf("expected raw delegated outcome kind, got %q", got)
	}
}

func TestNodeExecutionProjectionStopReasonPreservesRawTerminationKind(t *testing.T) {
	item := NodeExecutionProjection{
		Termination: &NodeTerminationProjection{Kind: " budget_exhausted "},
	}
	if got := nodeExecutionProjectionStopReason(item); got != " budget_exhausted " {
		t.Fatalf("expected raw termination kind, got %q", got)
	}
}

func TestNodeExecutionContractDiffFromJSONPreservesRawFields(t *testing.T) {
	diff := NodeExecutionContractDiffFromJSON(`{"changed_fields":[" visibility "," evidence "]}`)
	if len(diff) != 2 || diff[0] != " visibility " || diff[1] != " evidence " {
		t.Fatalf("expected raw contract diff fields, got %#v", diff)
	}
}

func TestSelectTerminalNodeExecutionProjection(t *testing.T) {
	items := []NodeExecutionProjection{
		{
			NodeExecID: "node-1",
			NodeID:     "plan",
			Status:     "completed",
			StartedAt:  100,
			FinishedAt: 200,
		},
		{
			NodeExecID: "node-2",
			NodeID:     "main",
			Status:     "incomplete",
			StartedAt:  250,
			FinishedAt: 300,
			DelegatedExecution: &NodeDelegatedExecutionProjection{
				Driver:      "open_tool_loop",
				OutcomeKind: "terminated",
				Rounds: []NodeDelegatedRoundProjection{
					{Round: 1, StopReason: "max_rounds"},
				},
			},
		},
		{
			NodeExecID: "node-3",
			NodeID:     "cleanup",
			Status:     "completed",
			StartedAt:  275,
			FinishedAt: 290,
		},
		{
			NodeExecID:       "node-4",
			ParentNodeExecID: "node-2",
			NodeID:           "main/round:2",
			Status:           "completed",
			StartedAt:        400,
			FinishedAt:       450,
		},
	}

	selected := SelectTerminalNodeExecutionProjection(items)
	if selected == nil {
		t.Fatalf("expected terminal node execution projection")
	}
	if selected.NodeExecID != "node-2" || selected.NodeID != "main" {
		t.Fatalf("unexpected terminal node execution projection: %#v", selected)
	}
	if selected.DelegatedExecution == nil || selected.DelegatedExecution.Driver != "open_tool_loop" {
		t.Fatalf("expected delegated execution to be preserved, got %#v", selected)
	}
	selected.NodeID = "mutated"
	if items[1].NodeID != "main" {
		t.Fatalf("expected cloned terminal projection, got %#v", items[1])
	}
}

func TestSelectTerminalNodeExecutionProjectionTieBreakUsesRawNodeExecID(t *testing.T) {
	items := []NodeExecutionProjection{
		{
			NodeExecID: " z-node ",
			NodeID:     "main",
			Status:     "completed",
			StartedAt:  100,
			FinishedAt: 200,
		},
		{
			NodeExecID: "a-node",
			NodeID:     "main",
			Status:     "completed",
			StartedAt:  100,
			FinishedAt: 200,
		},
	}

	selected := SelectTerminalNodeExecutionProjection(items)
	if selected == nil || selected.NodeExecID != "a-node" {
		t.Fatalf("expected raw node exec id tie-break, got %#v", selected)
	}
}

func TestSelectChildNodeExecutionProjections(t *testing.T) {
	items := []NodeExecutionProjection{
		{
			NodeExecID:       "round-2",
			ParentNodeExecID: "node-2",
			NodeID:           "main/round:2",
			StartedAt:        200,
			DelegatedExecution: &NodeDelegatedExecutionProjection{
				Rounds: []NodeDelegatedRoundProjection{{Round: 2}},
			},
		},
		{
			NodeExecID:       "round-1",
			ParentNodeExecID: "node-2",
			NodeID:           "main/round:1",
			StartedAt:        100,
			DelegatedExecution: &NodeDelegatedExecutionProjection{
				Rounds: []NodeDelegatedRoundProjection{{Round: 1}},
			},
		},
		{
			NodeExecID:       "round-other",
			ParentNodeExecID: "node-3",
			NodeID:           "other/round:1",
			StartedAt:        50,
		},
	}

	selected := SelectChildNodeExecutionProjections(items, "node-2")
	if len(selected) != 2 {
		t.Fatalf("expected 2 child projections, got %#v", selected)
	}
	if selected[0].NodeExecID != "round-1" || selected[1].NodeExecID != "round-2" {
		t.Fatalf("expected child projections ordered by round, got %#v", selected)
	}
	selected[0].NodeID = "mutated"
	if items[1].NodeID != "main/round:1" {
		t.Fatalf("expected cloned child projection, got %#v", items[1])
	}
}

func TestSelectChildNodeExecutionProjectionsTieBreakUsesRawNodeExecID(t *testing.T) {
	items := []NodeExecutionProjection{
		{
			NodeExecID:       " z-round ",
			ParentNodeExecID: "node-2",
			NodeID:           "main/round:1",
			StartedAt:        100,
			DelegatedExecution: &NodeDelegatedExecutionProjection{
				Rounds: []NodeDelegatedRoundProjection{{Round: 1}},
			},
		},
		{
			NodeExecID:       "a-round",
			ParentNodeExecID: "node-2",
			NodeID:           "main/round:1",
			StartedAt:        100,
			DelegatedExecution: &NodeDelegatedExecutionProjection{
				Rounds: []NodeDelegatedRoundProjection{{Round: 1}},
			},
		},
	}

	selected := SelectChildNodeExecutionProjections(items, "node-2")
	if len(selected) != 2 || selected[0].NodeExecID != " z-round " || selected[1].NodeExecID != "a-round" {
		t.Fatalf("expected raw node exec id ordering, got %#v", selected)
	}
}

func TestSelectChildNodeExecutionProjectionsRequiresExactParentNodeExecID(t *testing.T) {
	items := []NodeExecutionProjection{{
		NodeExecID:       "round-1",
		ParentNodeExecID: " node-parent ",
	}}

	if selected := SelectChildNodeExecutionProjections(items, "node-parent"); len(selected) != 0 {
		t.Fatalf("expected exact parent node exec id match, got %#v", selected)
	}
	if selected := SelectChildNodeExecutionProjections(items, " node-parent "); len(selected) != 1 {
		t.Fatalf("expected raw parent node exec id match, got %#v", selected)
	}
}

func TestAttachChildNodeExecutionProjections(t *testing.T) {
	items := []NodeExecutionProjection{
		{
			NodeExecID: "node-root",
			NodeID:     "main",
			Status:     "incomplete",
			DelegatedExecution: &NodeDelegatedExecutionProjection{
				Driver:      "open_tool_loop",
				OutcomeKind: "terminated",
				RoundCount:  2,
			},
		},
		{
			NodeExecID:       "node-round-2",
			ParentNodeExecID: "node-root",
			NodeID:           "main/round:2",
			Status:           "completed",
			StartedAt:        200,
			DelegatedExecution: &NodeDelegatedExecutionProjection{
				Rounds: []NodeDelegatedRoundProjection{{NodeExecID: "node-round-2", Round: 2}},
			},
		},
		{
			NodeExecID:       "node-round-1",
			ParentNodeExecID: "node-root",
			NodeID:           "main/round:1",
			Status:           "completed",
			StartedAt:        100,
			DelegatedExecution: &NodeDelegatedExecutionProjection{
				Rounds: []NodeDelegatedRoundProjection{{NodeExecID: "node-round-1", Round: 1}},
			},
		},
	}

	attached := AttachChildNodeExecutionProjections(items)
	if len(attached) != 1 {
		t.Fatalf("expected one top-level node with attached children, got %#v", attached)
	}
	if attached[0].NodeExecID != "node-root" {
		t.Fatalf("unexpected top-level node: %#v", attached[0])
	}
	if len(attached[0].ChildNodeExecutions) != 2 {
		t.Fatalf("expected attached child round nodes, got %#v", attached[0])
	}
	if attached[0].ChildNodeExecutions[0].NodeExecID != "node-round-1" || attached[0].ChildNodeExecutions[1].NodeExecID != "node-round-2" {
		t.Fatalf("expected child nodes ordered by round, got %#v", attached[0].ChildNodeExecutions)
	}
	attached[0].ChildNodeExecutions[0].NodeID = "mutated"
	if items[2].NodeID != "main/round:1" {
		t.Fatalf("expected attached child nodes to be cloned, got %#v", items[2])
	}
}

func TestFindNodeExecutionProjection(t *testing.T) {
	items := AttachChildNodeExecutionProjections([]NodeExecutionProjection{
		{
			NodeExecID: "node-root",
			NodeID:     "main",
		},
		{
			NodeExecID:       "node-round-1",
			ParentNodeExecID: "node-root",
			NodeID:           "main/round:1",
			DelegatedExecution: &NodeDelegatedExecutionProjection{
				Rounds: []NodeDelegatedRoundProjection{{NodeExecID: "node-round-1", Round: 1}},
			},
		},
	})

	found := FindNodeExecutionProjection(items, "node-round-1")
	if found == nil || found.NodeExecID != "node-round-1" {
		t.Fatalf("expected to find child node execution projection, got %#v", found)
	}
	found.NodeID = "mutated"
	if items[0].ChildNodeExecutions[0].NodeID != "main/round:1" {
		t.Fatalf("expected found projection to be cloned, got %#v", items[0].ChildNodeExecutions[0])
	}
}

func TestFindNodeExecutionProjectionRequiresExactNodeExecID(t *testing.T) {
	items := AttachChildNodeExecutionProjections([]NodeExecutionProjection{{
		NodeExecID: " node-root ",
		NodeID:     "main",
	}})

	if found := FindNodeExecutionProjection(items, "node-root"); found != nil {
		t.Fatalf("expected exact node exec id lookup, got %#v", found)
	}
	if found := FindNodeExecutionProjection(items, " node-root "); found == nil || found.NodeExecID != " node-root " {
		t.Fatalf("expected raw node exec id lookup, got %#v", found)
	}
}

func TestAttachChildNodeExecutionProjectionsRequiresExactParentNodeExecID(t *testing.T) {
	attached := AttachChildNodeExecutionProjections([]NodeExecutionProjection{
		{NodeExecID: " node-root ", NodeID: "main"},
		{NodeExecID: "child-1", ParentNodeExecID: " node-root ", NodeID: "main/round:1"},
		{NodeExecID: "child-2", ParentNodeExecID: "node-root", NodeID: "main/round:2"},
	})

	if len(attached) != 1 {
		t.Fatalf("expected one top-level node, got %#v", attached)
	}
	if len(attached[0].ChildNodeExecutions) != 1 || attached[0].ChildNodeExecutions[0].NodeExecID != "child-1" {
		t.Fatalf("expected exact raw parent attachment only, got %#v", attached[0].ChildNodeExecutions)
	}
}

func TestNodeExecutionProjectionRoundNumberRequiresCanonicalRoundSuffix(t *testing.T) {
	if got := nodeExecutionProjectionRoundNumber(NodeExecutionProjection{NodeID: "main/round: 2"}); got != 0 {
		t.Fatalf("expected malformed round suffix to stop parsing, got %d", got)
	}
	if got := nodeExecutionProjectionRoundNumber(NodeExecutionProjection{NodeID: "main/round:2"}); got != 2 {
		t.Fatalf("expected canonical round suffix, got %d", got)
	}
}

func TestNodeExecutionProjectionRoundOrderRequiresCanonicalRoundSuffix(t *testing.T) {
	if got := nodeExecutionProjectionRoundOrder(NodeExecutionProjection{NodeID: "main/round: 3"}); got != 0 {
		t.Fatalf("expected malformed round order suffix to stop parsing, got %d", got)
	}
	if got := nodeExecutionProjectionRoundOrder(NodeExecutionProjection{NodeID: "main/round:3"}); got != 3 {
		t.Fatalf("expected canonical round order suffix, got %d", got)
	}
}
