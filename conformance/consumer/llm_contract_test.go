package main

import (
	"encoding/json"
	"reflect"
	"testing"

	llm "github.com/wsnacj/agentx-go/components/llm"
)

func TestExternalStyleLLMComponentConsumer(t *testing.T) {
	t.Run("defensive clone", func(t *testing.T) {
		maxTokens := 128
		input := llm.ChatInput{
			Messages: llm.Conversation{{Role: "user", Content: "hello"}},
			Request: llm.RequestOptions{
				MaxTokens: &maxTokens,
				Headers:   map[string]string{"X-Request-ID": "request-1"},
				Metadata:  map[string]any{"tenant": "example"},
			},
			ToolChoice: &llm.ToolChoice{
				Type:     "function",
				Function: &llm.ToolChoiceFunction{Name: "lookup"},
			},
		}
		cloned := input.Clone()
		cloned.Messages[0].Content = "changed"
		cloned.Request.Headers["X-Request-ID"] = "request-2"
		cloned.Request.Metadata["tenant"] = "changed"
		cloned.ToolChoice.Function.Name = "changed"
		*cloned.Request.MaxTokens = 256

		if input.Messages[0].Content != "hello" ||
			input.Request.Headers["X-Request-ID"] != "request-1" ||
			input.Request.Metadata["tenant"] != "example" ||
			input.ToolChoice.Function.Name != "lookup" ||
			*input.Request.MaxTokens != 128 {
			t.Fatalf("Clone() retained mutable top-level state: %#v", input)
		}
	})

	t.Run("tool schema sanitize", func(t *testing.T) {
		original := map[string]any{
			"type": "object",
			"properties": map[string]any{
				"query": map[string]any{
					"type":     []any{"string", "null"},
					"nullable": true,
				},
			},
			"required": []any{"query", "missing"},
		}
		sanitized := llm.SanitizeFunctionParametersSchema(original)
		properties := sanitized["properties"].(map[string]any)
		query := properties["query"].(map[string]any)
		if query["type"] != "string" {
			t.Fatalf("sanitized query type = %#v", query["type"])
		}
		if _, ok := query["nullable"]; ok {
			t.Fatalf("sanitized schema retained nullable: %#v", query)
		}
		if got := sanitized["required"]; !reflect.DeepEqual(got, []string{"query"}) {
			t.Fatalf("sanitized required = %#v", got)
		}
		if original["properties"].(map[string]any)["query"].(map[string]any)["nullable"] != true {
			t.Fatalf("sanitize mutated caller input: %#v", original)
		}
	})

	t.Run("stream snapshot and stop reason", func(t *testing.T) {
		snapshot := llm.BuildStreamMessageSnapshot(
			map[int]string{2: "second", 0: "first"},
			map[int]string{1: "thinking"},
			map[int]llm.FunctionCall{3: {Name: "lookup", Arguments: `{"q":"x"}`}},
		)
		if !reflect.DeepEqual(snapshot.Text, []string{"first", "", "second", ""}) ||
			!reflect.DeepEqual(snapshot.Thinking, []string{"", "thinking", "", ""}) ||
			len(snapshot.ToolCalls) != 4 ||
			snapshot.ToolCalls[3].Name != "lookup" {
			t.Fatalf("unexpected snapshot: %#v", snapshot)
		}
		if got := llm.NormalizeStreamStopReason("MAX_TOKENS"); got != llm.StreamStopReasonLength {
			t.Fatalf("NormalizeStreamStopReason() = %q", got)
		}
	})

	t.Run("json contract round trip", func(t *testing.T) {
		type wire struct {
			Message llm.Message `json:"message"`
			Tool    llm.Tool    `json:"tool"`
			Usage   llm.Usage   `json:"usage"`
		}
		input := wire{
			Message: llm.Message{
				Role:    "assistant",
				Content: "done",
				ToolCalls: []llm.FunctionCall{{
					ID:        "call-1",
					Type:      "function",
					Name:      "lookup",
					Arguments: `{"q":"x"}`,
				}},
			},
			Tool: llm.Tool{
				Type: "function",
				Function: llm.Function{
					Name:       "lookup",
					Parameters: map[string]any{"type": "object"},
				},
			},
			Usage: llm.Usage{PromptTokens: 3, CompletionTokens: 2, TotalTokens: 5},
		}
		encoded, err := json.Marshal(input)
		if err != nil {
			t.Fatalf("Marshal() error = %v", err)
		}
		var decoded wire
		if err := json.Unmarshal(encoded, &decoded); err != nil {
			t.Fatalf("Unmarshal() error = %v", err)
		}
		if !reflect.DeepEqual(decoded, input) {
			t.Fatalf("round trip mismatch\ngot:  %#v\nwant: %#v", decoded, input)
		}
	})
}
