package codex

import (
	"encoding/json"
	"fmt"
	"strings"

	llm "github.com/wsnacj/agentx-go/components/llm"
)

type responseEnvelope struct {
	Output     []responseOutputItem `json:"output"`
	OutputText string               `json:"output_text"`
	Usage      responseUsage        `json:"usage"`
	Status     string               `json:"status"`
	Error      any                  `json:"error"`
}

type responseOutputItem struct {
	Type      string                `json:"type"`
	Role      string                `json:"role"`
	Status    string                `json:"status"`
	ID        string                `json:"id"`
	CallID    string                `json:"call_id"`
	Name      string                `json:"name"`
	Arguments any                   `json:"arguments"`
	Input     any                   `json:"input"`
	Content   []responseContentPart `json:"content"`
}

type responseContentPart struct {
	Type string `json:"type"`
	Text string `json:"text"`
}
type responseUsage struct {
	InputTokens      int `json:"input_tokens"`
	OutputTokens     int `json:"output_tokens"`
	TotalTokens      int `json:"total_tokens"`
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
}

func parseResponse(body []byte) (*llm.ChatResponse, error) {
	var envelope responseEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, fmt.Errorf("decode codex response: %w", err)
	}
	status := strings.ToLower(strings.TrimSpace(envelope.Status))
	if status == "failed" || status == "cancelled" {
		return nil, fmt.Errorf("codex response status %s: %s", status, stringFromAny(envelope.Error))
	}
	var textParts []string
	var calls []llm.FunctionCall
	for index, item := range envelope.Output {
		switch strings.ToLower(strings.TrimSpace(item.Type)) {
		case "message":
			if text := extractMessageText(item); text != "" {
				textParts = append(textParts, text)
			}
		case "function_call", "custom_tool_call":
			if call := convertFunctionCall(item, index); call != nil {
				calls = append(calls, *call)
			}
		}
	}
	content := strings.TrimSpace(strings.Join(textParts, "\n"))
	if content == "" {
		content = strings.TrimSpace(envelope.OutputText)
	}
	if len(envelope.Output) == 0 && content == "" {
		return nil, fmt.Errorf("codex response has no output")
	}
	return &llm.ChatResponse{Content: content, Raw: body, Calls: calls}, nil
}

func extractUsage(body []byte) *llm.Usage {
	var envelope responseEnvelope
	if json.Unmarshal(body, &envelope) != nil {
		return nil
	}
	usage := envelope.Usage
	prompt := usage.InputTokens
	if prompt == 0 {
		prompt = usage.PromptTokens
	}
	completion := usage.OutputTokens
	if completion == 0 {
		completion = usage.CompletionTokens
	}
	total := usage.TotalTokens
	if total == 0 && (prompt > 0 || completion > 0) {
		total = prompt + completion
	}
	if prompt == 0 && completion == 0 && total == 0 {
		return nil
	}
	return &llm.Usage{PromptTokens: prompt, CompletionTokens: completion, TotalTokens: total}
}

func extractMessageText(item responseOutputItem) string {
	var parts []string
	for _, part := range item.Content {
		typeName := strings.ToLower(strings.TrimSpace(part.Type))
		if (typeName == "output_text" || typeName == "text") && part.Text != "" {
			parts = append(parts, part.Text)
		}
	}
	return strings.TrimSpace(strings.Join(parts, ""))
}

func convertFunctionCall(item responseOutputItem, index int) *llm.FunctionCall {
	status := strings.ToLower(strings.TrimSpace(item.Status))
	if status == "queued" || status == "in_progress" || status == "incomplete" {
		return nil
	}
	name := strings.TrimSpace(item.Name)
	if name == "" {
		return nil
	}
	arguments := stringFromAny(item.Arguments)
	if strings.EqualFold(item.Type, "custom_tool_call") {
		arguments = stringFromAny(item.Input)
	}
	arguments = strings.TrimSpace(arguments)
	if arguments == "" {
		arguments = "{}"
	}
	callID := responseCallID(item.CallID)
	if callID == "" {
		callID = responseCallID(item.ID)
	}
	if callID == "" {
		callID = deterministicCallID(name, arguments, index)
	}
	return &llm.FunctionCall{ID: callID, Type: "function", Name: name, Arguments: arguments}
}
