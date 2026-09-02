package tools_test

import (
	"context"
	"errors"
	"testing"

	toolcontract "github.com/wsnacj/agentx-go/components/tool"
	"github.com/wsnacj/agentx-go/tools"
)

func TestRegistryOrderRepairAndTypedError(t *testing.T) {
	registry := tools.NewRegistry()
	registry.Register(toolcontract.Definition{Type: "function", Function: toolcontract.Function{Name: "zeta_tool"}}, func(_ context.Context, call toolcontract.Call) (toolcontract.Result, error) { return call.Name, nil })
	registry.Register(toolcontract.Definition{Type: "function", Function: toolcontract.Function{Name: "alpha"}}, func(_ context.Context, call toolcontract.Call) (toolcontract.Result, error) { return call.Name, nil })
	definitions := registry.Definitions()
	if len(definitions) != 2 || definitions[0].Function.Name != "alpha" || definitions[1].Function.Name != "zeta_tool" {
		t.Fatalf("definitions=%#v", definitions)
	}
	result, err := registry.Execute(context.Background(), toolcontract.Call{Name: "zeta-tool"})
	if err != nil || result != "zeta_tool" {
		t.Fatalf("result=%q err=%v", result, err)
	}
	_, err = registry.Execute(context.Background(), toolcontract.Call{Name: "missing"})
	var nameErr *tools.ToolNameError
	if !errors.As(err, &nameErr) || nameErr.Code() != "invalid_tool_name" {
		t.Fatalf("error=%#v", err)
	}
}

func TestRegistryConcurrentUse(t *testing.T) {
	registry := tools.NewRegistry()
	registry.Register(toolcontract.Definition{Type: "function", Function: toolcontract.Function{Name: "echo"}}, func(_ context.Context, call toolcontract.Call) (toolcontract.Result, error) {
		return call.Arguments, nil
	})
	done := make(chan struct{}, 16)
	for index := 0; index < cap(done); index++ {
		go func() {
			defer func() { done <- struct{}{} }()
			_, _ = registry.Execute(context.Background(), toolcontract.Call{Name: "echo", Arguments: "ready"})
			_ = registry.Definitions()
			_ = registry.Version()
		}()
	}
	for index := 0; index < cap(done); index++ {
		<-done
	}
}

func TestRegistryValidateCallUsesRegisteredDefinitionAndNameRepair(t *testing.T) {
	registry := tools.NewRegistry()
	registry.Register(toolcontract.Definition{Type: "function", Function: toolcontract.Function{
		Name: "weather_lookup", Parameters: map[string]any{
			"type": "object", "additionalProperties": false,
			"properties": map[string]any{"location": map[string]any{
				"type": "string", "minLength": 1, "x-agentx-binding-sources": []any{"user_input"},
			}},
			"required": []string{"location"},
		},
	}}, func(context.Context, toolcontract.Call) (toolcontract.Result, error) { return "unused", nil })
	result, err := registry.ValidateCall(toolcontract.Call{Name: "weather-lookup", Arguments: `{"location":"北京"}`}, tools.BindingContext{UserInput: "北京天气"})
	if err != nil || !result.Ready() || len(result.Evidence) != 1 || result.Evidence[0].Source != tools.BindingSourceUserInput {
		t.Fatalf("result=%#v err=%v", result, err)
	}
}
