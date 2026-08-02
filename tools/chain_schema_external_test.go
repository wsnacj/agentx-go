package tools_test

import (
	"context"
	"errors"
	"testing"

	toolcontract "github.com/wsnacj/agentx-go/components/tool"
	"github.com/wsnacj/agentx-go/tools"
)

func TestChainExecutorFallbackAndStableDefinitions(t *testing.T) {
	first := tools.NewRegistry()
	second := tools.NewRegistry()
	second.Register(definition("echo"), func(_ context.Context, call toolcontract.Call) (toolcontract.Result, error) {
		return call.Arguments, nil
	})

	executor := tools.ChainExecutor{Executors: []toolcontract.Executor{first, second}}
	output, err := executor.Execute(context.Background(), toolcontract.Call{Name: "echo", Arguments: "hello"})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if output != "hello" {
		t.Fatalf("output = %q, want hello", output)
	}
	definitions := executor.Definitions()
	if len(definitions) != 1 || definitions[0].Function.Name != "echo" {
		t.Fatalf("definitions = %#v", definitions)
	}
}

func TestChainExecutorStopsOnKnownToolError(t *testing.T) {
	first := tools.NewRegistry()
	first.Register(definition("echo"), func(context.Context, toolcontract.Call) (toolcontract.Result, error) {
		return "", errors.New("boom")
	})
	second := tools.NewRegistry()
	second.Register(definition("echo"), func(context.Context, toolcontract.Call) (toolcontract.Result, error) {
		return "unexpected", nil
	})

	_, err := (tools.ChainExecutor{Executors: []toolcontract.Executor{first, second}}).Execute(
		context.Background(), toolcontract.Call{Name: "echo"},
	)
	if err == nil || err.Error() != "boom" {
		t.Fatalf("error = %v, want boom", err)
	}
}

func TestSanitizeToolDefinitionDoesNotMutateSource(t *testing.T) {
	properties := map[string]any{
		"query": map[string]any{"type": []any{"string", "null"}},
	}
	definition := toolcontract.Definition{Function: toolcontract.Function{
		Name: "lookup",
		Parameters: map[string]any{
			"properties": properties,
			"required":   []any{"query", "missing", "query"},
		},
	}}

	sanitized := tools.SanitizeToolDefinitionForBackendCompatibility(definition)
	if sanitized.Type != "function" || sanitized.Function.Parameters["type"] != "object" {
		t.Fatalf("sanitized = %#v", sanitized)
	}
	sanitizedProperties := sanitized.Function.Parameters["properties"].(map[string]any)
	if got := sanitizedProperties["query"].(map[string]any)["type"]; got != "string" {
		t.Fatalf("query type = %#v", got)
	}
	required := sanitized.Function.Parameters["required"].([]any)
	if len(required) != 1 || required[0] != "query" {
		t.Fatalf("required = %#v", required)
	}
	if _, ok := properties["query"].(map[string]any)["type"].([]any); !ok {
		t.Fatalf("source schema mutated: %#v", properties)
	}
}

func TestNormalizeToolNameKeepsLegacyAliases(t *testing.T) {
	if got := tools.NormalizeToolName(" Bash "); got != "exec" {
		t.Fatalf("NormalizeToolName(Bash) = %q", got)
	}
	if got := tools.NormalizeToolName("apply-patch"); got != "apply_patch" {
		t.Fatalf("NormalizeToolName(apply-patch) = %q", got)
	}
}

func definition(name string) toolcontract.Definition {
	return toolcontract.Definition{Type: "function", Function: toolcontract.Function{
		Name: name,
		Parameters: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}}
}
