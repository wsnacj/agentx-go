package llmtask

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	types "github.com/wsnacj/agentx-go/components/llm"
	"github.com/wsnacj/agentx-go/tools"
)

func containsToolUserChunk(input types.ChatInput, needle string) bool {
	for _, message := range input.Messages {
		if strings.Contains(message.Content, needle) {
			return true
		}
	}
	return false
}

func TestRegister_SuccessWithSchemaValidation(t *testing.T) {
	reg := tools.NewRegistry()
	var gotModel string
	var gotInput types.ChatInput
	Register(reg, Options{
		ModelConfig: "test-model",
		ChatWithInput: func(_ context.Context, input types.ChatInput) (*types.ChatResponse, error) {
			gotModel = input.ConfigName
			gotInput = input.Clone()
			return &types.ChatResponse{Content: `{"status":"ok","score":0.92}`}, nil
		},
	})
	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name: "llm_task",
		Arguments: `{
			"instruction":"summarize",
			"input":"demo text",
			"schema":{
				"type":"object",
				"required":["status"],
				"properties":{
					"status":{"type":"string"},
					"score":{"type":"number"}
				}
			}
		}`,
	})
	if err != nil {
		t.Fatalf("execute llm_task: %v", err)
	}
	if gotModel != "test-model" {
		t.Fatalf("unexpected model: %q", gotModel)
	}
	if !containsToolUserChunk(gotInput, `"required":["status"]`) || !containsToolUserChunk(gotInput, `"score":{"type":"number"}`) {
		t.Fatalf("expected schema included in llm_task prompt, got %#v", gotInput.Messages)
	}
	if gotInput.ToolChoice == nil || gotInput.ToolChoice.Type != "none" {
		t.Fatalf("expected tool_choice=none, got %#v", gotInput.ToolChoice)
	}
	responseFormat, ok := gotInput.Request.ResponseFormat.(map[string]any)
	if !ok || responseFormat["type"] != "json_schema" {
		t.Fatalf("expected json_schema response format, got %#v", gotInput.Request.ResponseFormat)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	result, ok := payload["result"].(map[string]any)
	if !ok {
		t.Fatalf("missing result object: %#v", payload)
	}
	if result["status"] != "ok" {
		t.Fatalf("unexpected result status: %#v", result)
	}
	if payload["schema_applied"] != true {
		t.Fatalf("expected schema_applied=true, got %#v", payload["schema_applied"])
	}
}

func TestRegister_PrefersTypedChatWithInput(t *testing.T) {
	reg := tools.NewRegistry()
	var gotInput types.ChatInput
	Register(reg, Options{
		ModelConfig: "test-model",
		ChatWithInput: func(_ context.Context, input types.ChatInput) (*types.ChatResponse, error) {
			gotInput = input.Clone()
			return &types.ChatResponse{Content: `{"status":"ok"}`}, nil
		},
	})
	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name: "llm_task",
		Arguments: `{
			"instruction":"summarize",
			"input":"demo text",
			"schema":{
				"type":"object",
				"required":["status"],
				"properties":{"status":{"type":"string"}}
			}
		}`,
	})
	if err != nil {
		t.Fatalf("execute llm_task: %v", err)
	}
	if gotInput.ConfigName != "test-model" || strings.TrimSpace(gotInput.SystemPrompt) == "" {
		t.Fatalf("unexpected typed llm_task input: %#v", gotInput)
	}
	if len(gotInput.Messages) != 1 || !strings.Contains(gotInput.Messages[0].Content, "Instruction:\nsummarize") {
		t.Fatalf("expected instruction to be encoded in typed llm_task input, got %#v", gotInput.Messages)
	}
	if gotInput.ToolChoice == nil || gotInput.ToolChoice.Type != "none" {
		t.Fatalf("expected tool_choice=none, got %#v", gotInput.ToolChoice)
	}
	responseFormat, ok := gotInput.Request.ResponseFormat.(map[string]any)
	if !ok || responseFormat["type"] != "json_schema" {
		t.Fatalf("expected typed json_schema response format, got %#v", gotInput.Request.ResponseFormat)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload["schema_applied"] != true {
		t.Fatalf("expected schema_applied=true, got %#v", payload["schema_applied"])
	}
}

func TestRegister_InvalidJSON(t *testing.T) {
	reg := tools.NewRegistry()
	Register(reg, Options{
		ModelConfig: "test-model",
		ChatWithInput: func(_ context.Context, _ types.ChatInput) (*types.ChatResponse, error) {
			return &types.ChatResponse{Content: "not-json"}, nil
		},
	})
	_, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "llm_task",
		Arguments: `{"instruction":"extract"}`,
	})
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "valid json") {
		t.Fatalf("expected invalid json error, got %v", err)
	}
}

func TestRegister_SchemaValidationFail(t *testing.T) {
	reg := tools.NewRegistry()
	Register(reg, Options{
		ModelConfig: "test-model",
		ChatWithInput: func(_ context.Context, _ types.ChatInput) (*types.ChatResponse, error) {
			return &types.ChatResponse{Content: `{"status":123}`}, nil
		},
	})
	_, err := reg.Execute(context.Background(), types.FunctionCall{
		Name: "llm_task",
		Arguments: `{
			"instruction":"extract",
			"schema":{
				"type":"object",
				"properties":{"status":{"type":"string"}},
				"required":["status"]
			}
		}`,
	})
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "schema validation failed") {
		t.Fatalf("expected schema validation error, got %v", err)
	}
}

func TestRegister_ModelOverridePolicy(t *testing.T) {
	reg := tools.NewRegistry()
	Register(reg, Options{
		ModelConfig: "default-model",
		ChatWithInput: func(_ context.Context, _ types.ChatInput) (*types.ChatResponse, error) {
			return &types.ChatResponse{Content: `{"ok":true}`}, nil
		},
	})
	_, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "llm_task",
		Arguments: `{"instruction":"extract","model":"other-model"}`,
	})
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "override is not allowed") {
		t.Fatalf("expected model override blocked error, got %v", err)
	}
}

func TestRegister_SameModelValueDoesNotCountAsOverride(t *testing.T) {
	reg := tools.NewRegistry()
	Register(reg, Options{
		ModelConfig: "default-model",
		ChatWithInput: func(_ context.Context, input types.ChatInput) (*types.ChatResponse, error) {
			if input.ConfigName != "default-model" {
				t.Fatalf("expected default model preserved, got %q", input.ConfigName)
			}
			return &types.ChatResponse{Content: `{"ok":true}`}, nil
		},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "llm_task",
		Arguments: `{"instruction":"extract","model":"default-model"}`,
	})
	if err != nil {
		t.Fatalf("expected same-model override value to be allowed, got %v", err)
	}
	if !strings.Contains(out, `"schema_applied":false`) {
		t.Fatalf("unexpected output payload: %s", out)
	}
}

func TestRegister_CodeFencePayload(t *testing.T) {
	reg := tools.NewRegistry()
	Register(reg, Options{
		ModelConfig: "test-model",
		ChatWithInput: func(_ context.Context, _ types.ChatInput) (*types.ChatResponse, error) {
			return &types.ChatResponse{Content: "```json\n{\"ok\":true}\n```"}, nil
		},
	})
	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "llm_task",
		Arguments: `{"instruction":"extract"}`,
	})
	if err != nil {
		t.Fatalf("execute llm_task: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	result, ok := payload["result"].(map[string]any)
	if !ok || result["ok"] != true {
		t.Fatalf("unexpected code fence parse result: %#v", payload)
	}
}

func TestRegister_GoalAliasAccepted(t *testing.T) {
	reg := tools.NewRegistry()
	Register(reg, Options{
		ModelConfig: "test-model",
		ChatWithInput: func(_ context.Context, input types.ChatInput) (*types.ChatResponse, error) {
			if !containsToolUserChunk(input, "extract key updates") {
				t.Fatalf("expected goal alias to populate instruction, chunks=%v", input.Messages)
			}
			return &types.ChatResponse{Content: `{"ok":true}`}, nil
		},
	})
	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "llm_task",
		Arguments: `{"goal":"extract key updates"}`,
	})
	if err != nil {
		t.Fatalf("execute llm_task with goal alias: %v", err)
	}
	if !strings.Contains(out, `"schema_applied":false`) {
		t.Fatalf("unexpected output payload: %s", out)
	}
}

func TestRegister_ToleratesTrailingArgsJunk(t *testing.T) {
	reg := tools.NewRegistry()
	Register(reg, Options{
		ModelConfig: "test-model",
		ChatWithInput: func(_ context.Context, _ types.ChatInput) (*types.ChatResponse, error) {
			return &types.ChatResponse{Content: `{"ok":true}`}, nil
		},
	})
	_, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "llm_task",
		Arguments: `{"instruction":"extract"}]`,
	})
	if err != nil {
		t.Fatalf("expected trailing-args junk fallback to succeed, got %v", err)
	}
}

func TestRegister_ParsesFencedJSONWithComments(t *testing.T) {
	reg := tools.NewRegistry()
	Register(reg, Options{
		ModelConfig: "test-model",
		ChatWithInput: func(_ context.Context, _ types.ChatInput) (*types.ChatResponse, error) {
			return &types.ChatResponse{Content: "```json\n{\n  // line comment should be stripped\n  \"ok\": true,\n  \"url\": \"https://example.com/a//b\"\n}\n```\nextra notes"}, nil
		},
	})
	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "llm_task",
		Arguments: `{"instruction":"extract"}`,
	})
	if err != nil {
		t.Fatalf("execute llm_task with commented fenced json: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	result, ok := payload["result"].(map[string]any)
	if !ok || result["ok"] != true {
		t.Fatalf("unexpected parse result: %#v", payload)
	}
}

func TestRegister_DefaultTimeoutApplied(t *testing.T) {
	reg := tools.NewRegistry()
	Register(reg, Options{
		ModelConfig:      "test-model",
		DefaultTimeoutMs: 15,
		ChatWithInput: func(ctx context.Context, _ types.ChatInput) (*types.ChatResponse, error) {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(200 * time.Millisecond):
				return &types.ChatResponse{Content: `{"ok":true}`}, nil
			}
		},
	})
	_, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "llm_task",
		Arguments: `{"instruction":"extract"}`,
	})
	if err == nil || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected default timeout to trigger deadline exceeded, got %v", err)
	}
}

func TestValidateSchemaDefinition_InvalidType(t *testing.T) {
	err := validateSchemaDefinition(map[string]any{
		"type": "objectx",
	}, "$")
	if err == nil || !strings.Contains(err.Error(), "unsupported type") {
		t.Fatalf("expected unsupported type error, got %v", err)
	}
}

func TestValidateSchemaDefinitionRejectsImpossibleEnumAndConst(t *testing.T) {
	tests := []struct {
		name   string
		schema map[string]any
		expect string
	}{
		{
			name: "enum item type mismatch",
			schema: map[string]any{
				"type": "string",
				"enum": []any{1},
			},
			expect: "type constraint",
		},
		{
			name: "enum item violates pattern",
			schema: map[string]any{
				"type":    "string",
				"pattern": "^pass$",
				"enum":    []any{"fail"},
			},
			expect: "pattern",
		},
		{
			name: "enum duplicate entry",
			schema: map[string]any{
				"type": "string",
				"enum": []string{"pass", "pass"},
			},
			expect: "duplicate entries",
		},
		{
			name: "enum numeric duplicate entry",
			schema: map[string]any{
				"type": "number",
				"enum": []any{1, 1.0},
			},
			expect: "duplicate entries",
		},
		{
			name: "const violates minimum",
			schema: map[string]any{
				"type":    "number",
				"minimum": 1,
				"const":   0,
			},
			expect: "minimum",
		},
		{
			name: "object keyword requires object type",
			schema: map[string]any{
				"type":     "string",
				"required": []string{"status"},
			},
			expect: "requires declared type to include object",
		},
		{
			name: "array keyword requires array type",
			schema: map[string]any{
				"type":  "string",
				"items": map[string]any{"type": "string"},
			},
			expect: "requires declared type to include array",
		},
		{
			name: "string keyword requires string type",
			schema: map[string]any{
				"type":    "number",
				"pattern": "^pass$",
			},
			expect: "requires declared type to include string",
		},
		{
			name: "numeric keyword requires number type",
			schema: map[string]any{
				"type":    "object",
				"minimum": 1,
			},
			expect: "requires declared type to include number or integer",
		},
		{
			name: "invalid range",
			schema: map[string]any{
				"type":     "array",
				"minItems": 2,
				"maxItems": 1,
				"items":    map[string]any{"type": "string"},
			},
			expect: "maxItems",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateSchemaDefinition(tc.schema, "$")
			if err == nil || !strings.Contains(err.Error(), tc.expect) {
				t.Fatalf("expected %q error, got %v", tc.expect, err)
			}
		})
	}
}

func TestValidateSchemaDefinitionAcceptsStringSlices(t *testing.T) {
	err := validateSchemaDefinition(map[string]any{
		"type":     []string{"object"},
		"required": []string{"status"},
		"properties": map[string]any{
			"status": map[string]any{"type": "string"},
		},
		"enum": []string{},
	}, "$")
	if err == nil || !strings.Contains(err.Error(), "non-empty array") {
		t.Fatalf("expected empty []string enum to fail shape check, got %v", err)
	}

	err = validateSchemaDefinition(map[string]any{
		"type": []string{"string"},
		"enum": []string{"pass", "fail"},
	}, "$")
	if err != nil {
		t.Fatalf("expected []string type/enum to pass, got %v", err)
	}

	err = validateSchemaDefinition(map[string]any{
		"type": "integer",
		"enum": []int{1, 2},
	}, "$")
	if err != nil {
		t.Fatalf("expected []int enum to pass, got %v", err)
	}
}

func TestValidateJSONSchemaSubset_AdditionalPropertiesBlocked(t *testing.T) {
	schema := map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"properties": map[string]any{
			"id": map[string]any{"type": "string"},
		},
	}
	err := validateJSONSchemaSubset(map[string]any{
		"id":    "A-1",
		"extra": "x",
	}, schema, "$")
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "additional property") {
		t.Fatalf("expected additional property error, got %v", err)
	}
}

func TestValidateJSONSchemaSubset_EnumAndConstUseNumericEquality(t *testing.T) {
	enumSchema := map[string]any{
		"type": "number",
		"enum": []any{1},
	}
	if err := validateJSONSchemaSubset(float64(1), enumSchema, "$"); err != nil {
		t.Fatalf("expected numeric enum equality to pass, got %v", err)
	}

	constSchema := map[string]any{
		"type":  "number",
		"const": 1,
	}
	if err := validateJSONSchemaSubset(float64(1), constSchema, "$"); err != nil {
		t.Fatalf("expected numeric const equality to pass, got %v", err)
	}
}
