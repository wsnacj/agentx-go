package executionpolicy_test

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	executionpolicy "github.com/wsnacj/agentx-go/runtime/executionpolicy"
)

func TestContractJSONShape(t *testing.T) {
	contract := executionpolicy.Contract{
		ID:      "readonly",
		Version: 1,
		Strict:  true,
		Identity: executionpolicy.Identity{
			Pack:       "research",
			WorkflowID: "collect-v1",
		},
		Visibility: executionpolicy.VisibilityPolicy{
			AllowTools:      []string{"search"},
			RequireDeclared: true,
		},
		Budget: executionpolicy.BudgetPolicy{MaxToolCalls: 3},
	}
	want := `{"id":"readonly","version":1,"strict":true,"identity":{"pack":"research","workflow_id":"collect-v1"},"visibility":{"allow_tools":["search"],"require_declared":true},"budget":{"max_tool_calls":3},"loop":{},"approval":{},"replay":{},"runtime_controls":{"control_plane":{}},"side_effects":{},"sandbox":{},"evidence":{},"inherit":{},"audit":{}}`
	encoded, err := json.Marshal(contract)
	if err != nil {
		t.Fatalf("Marshal(): %v", err)
	}
	if string(encoded) != want {
		t.Fatalf("JSON = %s, want %s", encoded, want)
	}
	var decoded executionpolicy.Contract
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Unmarshal(): %v", err)
	}
	if !reflect.DeepEqual(decoded, contract) {
		t.Fatalf("round trip = %#v, want %#v", decoded, contract)
	}
}

func TestCompilerPortPreservesContextAndErrorIdentity(t *testing.T) {
	type contextKey string
	ctx := context.WithValue(context.Background(), contextKey("identity"), "kept")
	sentinel := errors.New("host policy failure")
	compiler := compilerFunc(func(got context.Context, input executionpolicy.CompileInput) (executionpolicy.Snapshot, error) {
		if got != ctx || got.Value(contextKey("identity")) != "kept" {
			t.Fatal("context identity changed")
		}
		if input.Identity.RunID != "run-1" {
			t.Fatalf("input = %#v", input)
		}
		return executionpolicy.Snapshot{}, sentinel
	})
	_, err := compiler.Compile(ctx, executionpolicy.CompileInput{Identity: executionpolicy.Identity{RunID: "run-1"}})
	if !errors.Is(err, sentinel) {
		t.Fatalf("Compile() error = %v", err)
	}
}

type compilerFunc func(context.Context, executionpolicy.CompileInput) (executionpolicy.Snapshot, error)

func (f compilerFunc) Compile(ctx context.Context, input executionpolicy.CompileInput) (executionpolicy.Snapshot, error) {
	return f(ctx, input)
}
