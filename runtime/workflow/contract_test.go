package workflow

import (
	"encoding/json"
	"reflect"
	"testing"
)

type fieldContract struct {
	name string
	typ  reflect.Type
	tag  string
}

func TestExportedConstantsContract(t *testing.T) {
	got := []string{
		string(PlanningOpen),
		string(PlanningBounded),
		string(PlanningPlanless),
		string(NodeTool),
		string(NodeLLM),
		string(NodeAgent),
		string(NodeParallel),
		string(NodeCollect),
		string(NodeWait),
		string(NodeEvaluate),
		string(NodeApprove),
		string(NodeSubflow),
		string(NodeHumanInput),
		string(ExecInline),
		string(ExecTask),
		string(ExecRemote),
	}
	want := []string{
		"open", "bounded", "planless",
		"tool", "llm", "agent", "parallel", "collect", "wait",
		"evaluate", "approve", "subflow", "human_input",
		"inline", "task", "remote",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("constants = %#v, want %#v", got, want)
	}
}

func TestExportedStructContracts(t *testing.T) {
	assertStructContract(t, reflect.TypeFor[Spec](), []fieldContract{
		{name: "ID", typ: reflect.TypeFor[string](), tag: "id,omitempty"},
		{name: "Title", typ: reflect.TypeFor[string](), tag: "title,omitempty"},
		{name: "Description", typ: reflect.TypeFor[string](), tag: "description,omitempty"},
		{name: "Version", typ: reflect.TypeFor[string](), tag: "version,omitempty"},
		{name: "Pack", typ: reflect.TypeFor[string](), tag: "pack,omitempty"},
		{name: "CaseTypes", typ: reflect.TypeFor[[]string](), tag: "case_types,omitempty"},
		{name: "RouteHints", typ: reflect.TypeFor[[]string](), tag: "route_hints,omitempty"},
		{name: "PlanningMode", typ: reflect.TypeFor[PlanningMode](), tag: "planning_mode,omitempty"},
		{name: "EntryNode", typ: reflect.TypeFor[string](), tag: "entry_node,omitempty"},
		{name: "Nodes", typ: reflect.TypeFor[[]NodeSpec](), tag: "nodes,omitempty"},
		{name: "Edges", typ: reflect.TypeFor[[]EdgeSpec](), tag: "edges,omitempty"},
		{name: "StateSchema", typ: reflect.TypeFor[[]StateSlotSpec](), tag: "state_schema,omitempty"},
		{name: "ArtifactSchema", typ: reflect.TypeFor[[]ArtifactTypeRef](), tag: "artifact_schema,omitempty"},
		{name: "EvaluatorSchema", typ: reflect.TypeFor[[]EvaluatorRef](), tag: "evaluator_schema,omitempty"},
		{name: "DefaultContract", typ: reflect.TypeFor[string](), tag: "default_contract,omitempty"},
	})
	assertStructContract(t, reflect.TypeFor[NodeSpec](), []fieldContract{
		{name: "ID", typ: reflect.TypeFor[string](), tag: "id,omitempty"},
		{name: "Kind", typ: reflect.TypeFor[NodeKind](), tag: "kind,omitempty"},
		{name: "Title", typ: reflect.TypeFor[string](), tag: "title,omitempty"},
		{name: "Description", typ: reflect.TypeFor[string](), tag: "description,omitempty"},
		{name: "ContractRef", typ: reflect.TypeFor[string](), tag: "contract_ref,omitempty"},
		{name: "Inputs", typ: reflect.TypeFor[[]BindingSpec](), tag: "inputs,omitempty"},
		{name: "Outputs", typ: reflect.TypeFor[[]BindingSpec](), tag: "outputs,omitempty"},
		{name: "Retry", typ: reflect.TypeFor[RetryPolicy](), tag: "retry,omitempty"},
		{name: "TimeoutMs", typ: reflect.TypeFor[int64](), tag: "timeout_ms,omitempty"},
		{name: "ExecutionMode", typ: reflect.TypeFor[ExecutionMode](), tag: "execution_mode,omitempty"},
		{name: "Config", typ: reflect.TypeFor[map[string]any](), tag: "config,omitempty"},
	})
	assertStructContract(t, reflect.TypeFor[EdgeSpec](), []fieldContract{
		{name: "From", typ: reflect.TypeFor[string](), tag: "from,omitempty"},
		{name: "To", typ: reflect.TypeFor[string](), tag: "to,omitempty"},
		{name: "On", typ: reflect.TypeFor[string](), tag: "on,omitempty"},
		{name: "Condition", typ: reflect.TypeFor[string](), tag: "condition,omitempty"},
	})
	assertStructContract(t, reflect.TypeFor[StateSlotSpec](), []fieldContract{
		{name: "Name", typ: reflect.TypeFor[string](), tag: "name,omitempty"},
		{name: "Type", typ: reflect.TypeFor[string](), tag: "type,omitempty"},
		{name: "Required", typ: reflect.TypeFor[bool](), tag: "required,omitempty"},
	})
	assertStructContract(t, reflect.TypeFor[BindingSpec](), []fieldContract{
		{name: "From", typ: reflect.TypeFor[string](), tag: "from,omitempty"},
		{name: "To", typ: reflect.TypeFor[string](), tag: "to,omitempty"},
		{name: "Optional", typ: reflect.TypeFor[bool](), tag: "optional,omitempty"},
	})
	assertStructContract(t, reflect.TypeFor[RetryPolicy](), []fieldContract{
		{name: "MaxAttempts", typ: reflect.TypeFor[int](), tag: "max_attempts,omitempty"},
		{name: "BackoffMs", typ: reflect.TypeFor[[]int](), tag: "backoff_ms,omitempty"},
	})
	assertStructContract(t, reflect.TypeFor[ArtifactTypeRef](), []fieldContract{
		{name: "Type", typ: reflect.TypeFor[string](), tag: "type,omitempty"},
	})
	assertStructContract(t, reflect.TypeFor[EvaluatorRef](), []fieldContract{
		{name: "Name", typ: reflect.TypeFor[string](), tag: "name,omitempty"},
	})
}

func TestSpecJSONRoundTripAndZeroValue(t *testing.T) {
	zero, err := json.Marshal(Spec{})
	if err != nil {
		t.Fatalf("Marshal(zero): %v", err)
	}
	if got, want := string(zero), `{}`; got != want {
		t.Fatalf("zero JSON = %s, want %s", got, want)
	}

	in := Spec{
		ID:           "workflow-1",
		PlanningMode: PlanningBounded,
		EntryNode:    "collect",
		Nodes: []NodeSpec{{
			ID:            "collect",
			Kind:          NodeTool,
			ExecutionMode: ExecInline,
			Config:        map[string]any{"tool": "search"},
		}},
		Edges:           []EdgeSpec{{From: "collect", To: "finish", On: "success"}},
		StateSchema:     []StateSlotSpec{{Name: "query", Type: "string", Required: true}},
		ArtifactSchema:  []ArtifactTypeRef{{Type: "report"}},
		EvaluatorSchema: []EvaluatorRef{{Name: "quality"}},
	}
	payload, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("Marshal(): %v", err)
	}
	var out Spec
	if err := json.Unmarshal(payload, &out); err != nil {
		t.Fatalf("Unmarshal(): %v", err)
	}
	if !reflect.DeepEqual(out, in) {
		t.Fatalf("round trip\nout: %#v\nin: %#v", out, in)
	}
}

func assertStructContract(t *testing.T, typ reflect.Type, want []fieldContract) {
	t.Helper()
	if typ.NumField() != len(want) {
		t.Fatalf("%s field count = %d, want %d", typ, typ.NumField(), len(want))
	}
	for i, contract := range want {
		field := typ.Field(i)
		if field.Name != contract.name ||
			field.Type != contract.typ ||
			field.Tag.Get("json") != contract.tag {
			t.Fatalf(
				"%s field[%d] = %s %s json:%q, want %s %s json:%q",
				typ,
				i,
				field.Name,
				field.Type,
				field.Tag.Get("json"),
				contract.name,
				contract.typ,
				contract.tag,
			)
		}
	}
}
