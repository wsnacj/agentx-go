package gemini

import (
	"context"
	"encoding/json"
	"testing"

	llm "github.com/wsnacj/agentx-go/components/llm"
)

func TestBuildGeneratePayloadProjectsCanonicalToolContract(t *testing.T) {
	payload, err := buildGeneratePayload(context.Background(), nil, ModelConfig{Capability: Capability{ToolCalling: true}}, llm.ChatRequest{
		Messages: llm.Conversation{
			{Role: "user", Content: "lookup fixture"},
			{Role: "assistant", ToolCalls: []llm.FunctionCall{{Name: "lookup", Arguments: `{"id":"fixture"}`}}},
			{Role: "tool", ToolName: "lookup", Content: `{"value":"ready"}`},
		},
		Tools: []llm.Tool{{Type: "function", Function: llm.Function{
			Name: "lookup", Description: "lookup fixture", Parameters: map[string]any{
				"type": "object", "properties": map[string]any{"id": map[string]any{"type": "string"}},
			},
		}}},
		ToolChoice: &llm.ToolChoice{Type: "function", Function: &llm.ToolChoiceFunction{Name: "lookup"}},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	tools, ok := payload["tools"].([]map[string]any)
	if !ok || len(tools) != 1 {
		t.Fatalf("tools = %#v", payload["tools"])
	}
	declarations, ok := tools[0]["functionDeclarations"].([]map[string]any)
	if !ok || len(declarations) != 1 || declarations[0]["name"] != "lookup" {
		t.Fatalf("functionDeclarations = %#v", tools[0]["functionDeclarations"])
	}
	config := payload["toolConfig"].(map[string]any)["functionCallingConfig"].(map[string]any)
	if config["mode"] != "ANY" {
		t.Fatalf("toolConfig = %#v", config)
	}
	allowed := config["allowedFunctionNames"].([]string)
	if len(allowed) != 1 || allowed[0] != "lookup" {
		t.Fatalf("allowedFunctionNames = %#v", allowed)
	}
	contents := payload["contents"].([]map[string]any)
	if len(contents) != 3 || contents[1]["role"] != "model" || contents[2]["role"] != "user" {
		t.Fatalf("contents = %#v", contents)
	}
	assistantParts := contents[1]["parts"].([]map[string]any)
	if _, ok := assistantParts[0]["functionCall"]; !ok {
		t.Fatalf("assistant parts = %#v", assistantParts)
	}
	toolParts := contents[2]["parts"].([]map[string]any)
	response := toolParts[0]["functionResponse"].(map[string]any)["response"].(map[string]any)
	if response["value"] != "ready" {
		t.Fatalf("functionResponse = %#v", response)
	}
}

func TestBuildGeneratePayloadToolContractFailsClosed(t *testing.T) {
	request := llm.ChatRequest{
		Tools:      []llm.Tool{{Type: "function", Function: llm.Function{Name: "lookup"}}},
		ToolChoice: &llm.ToolChoice{Type: "function", Function: &llm.ToolChoiceFunction{Name: "missing"}},
	}
	if _, err := buildGeneratePayload(context.Background(), nil, ModelConfig{}, request, nil); err == nil {
		t.Fatal("expected disabled capability error")
	}
	if _, err := buildGeneratePayload(context.Background(), nil, ModelConfig{Capability: Capability{ToolCalling: true}}, request, nil); err == nil {
		t.Fatal("expected undeclared function error")
	}
}

func TestExtractFunctionCallsAndReasoningUsage(t *testing.T) {
	response := &GenerateContentResponse{
		Candidates: []Candidate{{Content: Content{Parts: []Part{{FunctionCall: &FunctionCall{
			Name: "lookup", Args: map[string]any{"id": "fixture"},
		}}}}}},
		UsageMetadata: &UsageMetadata{PromptTokenCount: 2, CandidatesTokenCount: 3, TotalTokenCount: 8, ThoughtsTokenCount: 3},
	}
	calls := extractFunctionCalls(response)
	if len(calls) != 1 || calls[0].Name != "lookup" || calls[0].Arguments != `{"id":"fixture"}` {
		t.Fatalf("calls = %#v", calls)
	}
	usage := extractUsage(response)
	if usage == nil || usage.TotalTokens != 8 || usage.ReasoningTokens != 3 {
		t.Fatalf("usage = %#v", usage)
	}
}

func TestGeminiStreamParserEmitsNormalizedToolEvents(t *testing.T) {
	parser := &geminiStreamEventParser{}
	events := parser.ParseResponse(&GenerateContentResponse{
		Candidates: []Candidate{{
			Content: Content{Parts: []Part{{
				FunctionCall: &FunctionCall{Name: "lookup", Args: map[string]any{"id": "fixture"}},
			}}},
			FinishReason: "STOP",
		}},
	}, []byte("fixture"))
	if len(events) != 2 || events[0].Type != llm.StreamEventToolCallStart || events[1].Type != llm.StreamEventToolCallDelta {
		t.Fatalf("events = %#v", events)
	}
	if events[1].ToolCall == nil || events[1].ToolCall.Name != "lookup" || !json.Valid([]byte(events[1].ToolCall.Arguments)) {
		t.Fatalf("tool delta = %#v", events[1].ToolCall)
	}
	done := parser.DoneEvents(nil)
	if len(done) != 2 || done[0].Type != llm.StreamEventToolCallEnd || done[1].StopReason != llm.StreamStopReasonToolUse {
		t.Fatalf("done = %#v", done)
	}
}
