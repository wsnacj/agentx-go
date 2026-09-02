package tools_test

import (
	"errors"
	"testing"

	toolcontract "github.com/wsnacj/agentx-go/components/tool"
	tools "github.com/wsnacj/agentx-go/tools"
	"github.com/wsnacj/agentx-go/tools/weather"
)

func TestValidateCallArgumentsRequiresTrustedWeatherLocation(t *testing.T) {
	definition := weather.Definition()
	tests := []struct {
		name       string
		raw        string
		bindings   tools.BindingContext
		ready      bool
		wantSource tools.BindingSource
	}{
		{name: "user input", raw: `{"location":"北京"}`, bindings: tools.BindingContext{UserInput: "请告诉我北京今天天气如何"}, ready: true, wantSource: tools.BindingSourceUserInput},
		{name: "trusted host", raw: `{"location":"Beijing"}`, bindings: tools.BindingContext{TrustedHost: map[string]string{"$.location": "Beijing"}}, ready: true, wantSource: tools.BindingSourceTrustedHost},
		{name: "invented", raw: `{"location":"北京"}`, bindings: tools.BindingContext{UserInput: "请告诉我今天天气如何"}},
		{name: "missing", raw: `{}`, bindings: tools.BindingContext{UserInput: "请告诉我今天天气如何"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := tools.ValidateCallArguments(definition, test.raw, test.bindings)
			if err != nil {
				t.Fatalf("ValidateCallArguments() error = %v", err)
			}
			if result.Ready() != test.ready {
				t.Fatalf("Ready() = %v, result = %#v", result.Ready(), result)
			}
			if test.ready && (len(result.Evidence) != 1 || result.Evidence[0].Path != "$.location" || result.Evidence[0].Source != test.wantSource) {
				t.Fatalf("evidence = %#v", result.Evidence)
			}
			if !test.ready && (len(result.NeedsClarification) != 1 || result.NeedsClarification[0] != "$.location") {
				t.Fatalf("clarification = %#v", result.NeedsClarification)
			}
		})
	}
}

func TestValidateCallArgumentsRejectsMalformedWeatherArguments(t *testing.T) {
	definition := weather.Definition()
	for name, raw := range map[string]string{
		"unknown field": `{"location":"北京","locaton":"北京"}`,
		"wrong type":    `{"location":1}`,
		"empty":         `{"location":""}`,
		"code fence":    "```json\n{\"location\":\"北京\"}\n```",
		"array wrapper": `[{"location":"北京"}]`,
		"trailing json": `{"location":"北京"} {}`,
	} {
		t.Run(name, func(t *testing.T) {
			_, err := tools.ValidateCallArguments(definition, raw, tools.BindingContext{UserInput: "北京天气"})
			if !errors.Is(err, tools.ErrInvalidToolArguments) {
				t.Fatalf("error = %v", err)
			}
		})
	}
}

func TestValidateCallArgumentsRejectsUnsupportedDefinition(t *testing.T) {
	definition := toolcontract.Definition{Type: "function", Function: toolcontract.Function{
		Name: "fixture", Parameters: map[string]any{"type": "object", "unevaluatedProperties": false},
	}}
	_, err := tools.ValidateCallArguments(definition, `{}`, tools.BindingContext{})
	if !errors.Is(err, tools.ErrInvalidToolDefinition) {
		t.Fatalf("error = %v", err)
	}
}
