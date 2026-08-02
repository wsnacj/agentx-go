package openaicompat

import (
	"testing"

	llm "github.com/wsnacj/agentx-go/components/llm"
)

func TestPayloadCompatibilityRules(t *testing.T) {
	client, err := New(Config{BaseURL: "https://example.invalid/v1"})
	if err != nil {
		t.Fatal(err)
	}
	parallel := true
	payload, _ := client.buildChatPayload(ModelConfig{Model: "fixture"}, llm.ChatRequest{
		Model: "fixture", MaxTokens: 33,
		Messages: llm.Conversation{{Role: "tool", Content: `{"ok":true}`}},
		Tools: []llm.Tool{{Type: "function", Function: llm.Function{
			Name:         "lookup",
			Parameters:   map[string]any{"type": []any{"object", "null"}, "properties": map[string]any{"query": map[string]any{"type": []any{"string", "null"}}}, "required": []any{"query", "missing"}},
			OutputSchema: map[string]any{"type": "object"},
		}}},
		Options: llm.RequestOptions{ParallelToolCalls: &parallel}.ToMap(),
	}, true)
	if payload["max_completion_tokens"] != 33 || payload["parallel_tool_calls"] != true {
		t.Fatalf("payload = %#v", payload)
	}
	messages := payload["messages"].([]map[string]any)
	if messages[0]["role"] != "assistant" {
		t.Fatalf("messages = %#v", messages)
	}
	tools := payload["tools"].([]map[string]any)
	function := tools[0]["function"].(map[string]any)
	if function["strict"] != false {
		t.Fatalf("function = %#v", function)
	}
	if _, ok := function["output_schema"]; ok {
		t.Fatalf("function = %#v", function)
	}
	params := function["parameters"].(map[string]any)
	if params["type"] != "object" {
		t.Fatalf("params = %#v", params)
	}
	streamOptions := payload["stream_options"].(map[string]any)
	if streamOptions["include_usage"] != true {
		t.Fatalf("stream_options = %#v", streamOptions)
	}
}

func TestProfileSpecificMaxTokensAndReasoning(t *testing.T) {
	azure, _ := New(Config{BaseURL: "https://example.invalid", CompatProfile: "azure-openai"})
	payload, _ := azure.buildChatPayload(ModelConfig{}, llm.ChatRequest{Options: map[string]any{"max_completion_tokens": 41}}, false)
	if payload["max_tokens"] != 41 {
		t.Fatalf("azure payload = %#v", payload)
	}
	openRouter, _ := New(Config{BaseURL: "https://example.invalid", CompatProfile: "openrouter"})
	payload, _ = openRouter.buildChatPayload(ModelConfig{}, llm.ChatRequest{Options: map[string]any{"reasoning_effort": "high"}}, false)
	reasoning := payload["reasoning"].(map[string]any)
	if reasoning["effort"] != "high" {
		t.Fatalf("openrouter payload = %#v", payload)
	}
}
