package anthropic

import (
	"encoding/json"
	"fmt"
	"strings"

	llm "github.com/wsnacj/agentx-go/components/llm"
)

func buildMessagesPayload(req llm.ChatRequest) (map[string]any, error) {
	payload := map[string]any{
		"model":      req.Model,
		"messages":   buildMessages(req.Messages),
		"max_tokens": max(1, req.MaxTokens),
	}
	if strings.TrimSpace(req.System) != "" {
		payload["system"] = req.System
	}
	if req.Temperature != 0 {
		payload["temperature"] = req.Temperature
	}
	if tools := buildTools(req.Tools); len(tools) > 0 {
		payload["tools"] = tools
	}
	if choice := buildToolChoice(req.ToolChoice); choice != nil {
		payload["tool_choice"] = choice
	}
	return payload, nil
}

func buildMessages(messages llm.Conversation) []map[string]any {
	out := make([]map[string]any, 0, len(messages))
	for _, message := range messages {
		role := strings.ToLower(strings.TrimSpace(message.Role))
		switch role {
		case "system":
			if strings.TrimSpace(message.Content) != "" {
				out = append(out, map[string]any{"role": "user", "content": message.Content})
			}
		case "assistant":
			out = append(out, buildAssistantMessage(message))
		case "tool":
			out = append(out, map[string]any{
				"role": "user",
				"content": []map[string]any{{
					"type":        "tool_result",
					"tool_use_id": strings.TrimSpace(message.ToolCallID),
					"content":     message.Content,
				}},
			})
		default:
			out = append(out, map[string]any{"role": "user", "content": message.Content})
		}
	}
	if len(out) == 0 {
		return []map[string]any{{"role": "user", "content": ""}}
	}
	return out
}

func buildAssistantMessage(message llm.Message) map[string]any {
	if len(message.ToolCalls) == 0 {
		return map[string]any{"role": "assistant", "content": message.Content}
	}
	content := make([]map[string]any, 0, 1+len(message.ToolCalls))
	if strings.TrimSpace(message.Content) != "" {
		content = append(content, map[string]any{"type": "text", "text": message.Content})
	}
	for _, call := range message.ToolCalls {
		input := map[string]any{}
		if strings.TrimSpace(call.Arguments) != "" {
			var parsed any
			if err := json.Unmarshal([]byte(call.Arguments), &parsed); err == nil && parsed != nil {
				input = normalizeToolUseInput(parsed)
			}
		}
		content = append(content, map[string]any{
			"type": "tool_use", "id": firstNonEmpty(call.ID, call.Name),
			"name": call.Name, "input": input,
		})
	}
	return map[string]any{"role": "assistant", "content": content}
}

func normalizeToolUseInput(value any) map[string]any {
	if mapped, ok := value.(map[string]any); ok {
		return mapped
	}
	return map[string]any{"value": value}
}

func buildTools(tools []llm.Tool) []map[string]any {
	if len(tools) == 0 {
		return nil
	}
	out := make([]map[string]any, 0, len(tools))
	for _, tool := range tools {
		if strings.TrimSpace(tool.Function.Name) == "" {
			continue
		}
		schema := tool.Function.Parameters
		if len(schema) == 0 {
			schema = map[string]any{"type": "object", "properties": map[string]any{}}
		}
		out = append(out, map[string]any{
			"name": tool.Function.Name, "description": tool.Function.Description,
			"input_schema": schema,
		})
	}
	return out
}

func buildToolChoice(choice *llm.ToolChoice) map[string]any {
	if choice == nil {
		return nil
	}
	switch strings.ToLower(strings.TrimSpace(choice.Type)) {
	case "", "auto":
		return map[string]any{"type": "auto"}
	case "none":
		return nil
	case "required":
		return map[string]any{"type": "any"}
	case "function", "tool":
		if choice.Function != nil && strings.TrimSpace(choice.Function.Name) != "" {
			return map[string]any{"type": "tool", "name": strings.TrimSpace(choice.Function.Name)}
		}
		return map[string]any{"type": "any"}
	default:
		return nil
	}
}

type messagesResponse struct {
	Content []struct {
		Type  string          `json:"type"`
		Text  string          `json:"text"`
		ID    string          `json:"id"`
		Name  string          `json:"name"`
		Input json.RawMessage `json:"input"`
	} `json:"content"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

func parseMessagesResponse(raw []byte) (*llm.ChatResponse, *llm.Usage, error) {
	var parsed messagesResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, nil, fmt.Errorf("decode anthropic messages response: %w", err)
	}
	var text strings.Builder
	calls := make([]llm.FunctionCall, 0)
	for _, item := range parsed.Content {
		switch item.Type {
		case "text":
			text.WriteString(item.Text)
		case "tool_use":
			calls = append(calls, llm.FunctionCall{ID: item.ID, Type: "function", Name: item.Name, Arguments: string(item.Input)})
		}
	}
	usage := &llm.Usage{
		PromptTokens: parsed.Usage.InputTokens, CompletionTokens: parsed.Usage.OutputTokens,
		TotalTokens: parsed.Usage.InputTokens + parsed.Usage.OutputTokens,
	}
	return &llm.ChatResponse{Content: text.String(), Calls: calls, Raw: raw, Usage: usage}, usage, nil
}
