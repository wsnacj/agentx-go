package codex

import (
	"crypto/sha1" // #nosec G505 -- deterministic public call identity, not cryptography.
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	llm "github.com/wsnacj/agentx-go/components/llm"
	"github.com/wsnacj/agentx-go/providers/transport"
)

func (c *Client) buildChatPayload(cfg ModelConfig, req llm.ChatRequest) (map[string]any, transport.Settings) {
	payload := map[string]any{
		"model": req.Model, "instructions": req.System, "input": buildInput(req.Messages),
		"store": false, "parallel_tool_calls": true,
	}
	if tools := buildTools(req.Tools); len(tools) > 0 {
		payload["tools"] = tools
		payload["tool_choice"] = buildToolChoice(req.ToolChoice)
	} else if req.ToolChoice != nil {
		payload["tool_choice"] = buildToolChoice(req.ToolChoice)
	}
	mergeOptions(payload, cfg, req)
	return payload, transport.Resolve(c.transport, llm.RequestOptionsFromMap(req.Options))
}

func buildInput(messages llm.Conversation) []map[string]any {
	out := make([]map[string]any, 0, len(messages))
	for _, message := range messages {
		switch strings.ToLower(strings.TrimSpace(message.Role)) {
		case "system":
			continue
		case "user":
			out = append(out, map[string]any{"role": "user", "content": message.Content})
		case "tool":
			callID := responseCallID(message.ToolCallID)
			if callID != "" {
				out = append(out, map[string]any{"type": "function_call_output", "call_id": callID, "output": message.Content})
			}
		default:
			if content := strings.TrimSpace(message.Content); content != "" {
				out = append(out, map[string]any{"role": "assistant", "content": content})
			}
			for index, call := range message.ToolCalls {
				if item := buildFunctionCallInput(call, index); item != nil {
					out = append(out, item)
				}
			}
		}
	}
	return out
}

func buildFunctionCallInput(call llm.FunctionCall, index int) map[string]any {
	name := strings.TrimSpace(call.Name)
	if name == "" {
		return nil
	}
	arguments := strings.TrimSpace(call.Arguments)
	if arguments == "" {
		arguments = "{}"
	}
	callID := responseCallID(call.ID)
	if callID == "" {
		callID = deterministicCallID(name, arguments, index)
	}
	return map[string]any{"type": "function_call", "call_id": callID, "name": name, "arguments": arguments}
}

func buildTools(tools []llm.Tool) []map[string]any {
	if len(tools) == 0 {
		return nil
	}
	tools = llm.SanitizeToolSchemas(tools)
	out := make([]map[string]any, 0, len(tools))
	for _, tool := range tools {
		name := strings.TrimSpace(tool.Function.Name)
		if name == "" {
			continue
		}
		parameters := tool.Function.Parameters
		if parameters == nil {
			parameters = map[string]any{"type": "object", "properties": map[string]any{}}
		}
		strict := false
		if tool.Function.Strict != nil {
			strict = *tool.Function.Strict
		}
		out = append(out, map[string]any{
			"type": "function", "name": name, "description": tool.Function.Description,
			"parameters": parameters, "strict": strict,
		})
	}
	return out
}

func buildToolChoice(choice *llm.ToolChoice) any {
	if choice == nil {
		return "auto"
	}
	choiceType := strings.ToLower(strings.TrimSpace(choice.Type))
	if choiceType == "" && choice.Function != nil {
		choiceType = "function"
	}
	switch choiceType {
	case "", "auto", "none", "required":
		return firstNonEmpty(choiceType, "auto")
	case "function":
		if choice.Function != nil && strings.TrimSpace(choice.Function.Name) != "" {
			return map[string]any{"type": "function", "name": strings.TrimSpace(choice.Function.Name)}
		}
		return map[string]any{"type": "function"}
	default:
		return choiceType
	}
}

func mergeOptions(payload map[string]any, cfg ModelConfig, req llm.ChatRequest) {
	options := llm.SanitizeProviderOptionMap(req.Options)
	if sessionID, _ := options["session_id"].(string); strings.TrimSpace(sessionID) != "" {
		payload["prompt_cache_key"] = strings.TrimSpace(sessionID)
	}
	if cacheKey, _ := options["prompt_cache_key"].(string); strings.TrimSpace(cacheKey) != "" {
		payload["prompt_cache_key"] = strings.TrimSpace(cacheKey)
	}
	if serviceTier, _ := options["service_tier"].(string); strings.TrimSpace(serviceTier) != "" {
		payload["service_tier"] = strings.TrimSpace(serviceTier)
	}
	if include, ok := options["include"].([]any); ok {
		payload["include"] = include
	} else if include, ok := options["include"].([]string); ok {
		values := make([]any, 0, len(include))
		for _, value := range include {
			if strings.TrimSpace(value) != "" {
				values = append(values, strings.TrimSpace(value))
			}
		}
		if len(values) > 0 {
			payload["include"] = values
		}
	}
	if parallel, ok := options["parallel_tool_calls"].(bool); ok {
		payload["parallel_tool_calls"] = parallel
	}
	if toolChoice, ok := options["tool_choice"]; ok && payload["tool_choice"] == nil {
		payload["tool_choice"] = toolChoice
	}
	if reasoning, ok := options["reasoning"].(map[string]any); ok {
		payload["reasoning"] = reasoning
	} else if effort := resolveReasoningEffort(cfg, options); effort != "" {
		payload["reasoning"] = map[string]any{"effort": effort, "summary": "auto"}
		payload["include"] = []any{"reasoning.encrypted_content"}
	}
}

func resolveReasoningEffort(cfg ModelConfig, options map[string]any) string {
	if !cfg.Thinking && cfg.ReasoningEffort != nil && !*cfg.ReasoningEffort {
		return ""
	}
	if effort, _ := options["reasoning_effort"].(string); strings.TrimSpace(effort) != "" {
		return normalizeReasoningEffort(effort)
	}
	return normalizeReasoningEffort(cfg.ReasoningDefault)
}

func normalizeReasoningEffort(effort string) string {
	switch strings.ToLower(strings.TrimSpace(effort)) {
	case "minimal":
		return "low"
	case "low", "medium", "high", "xhigh":
		return strings.ToLower(strings.TrimSpace(effort))
	default:
		return ""
	}
}

func responseCallID(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ""
	}
	if strings.Contains(value, "|") {
		value = strings.TrimSpace(strings.SplitN(value, "|", 2)[0])
	}
	if strings.HasPrefix(value, "fc_") {
		return "call_" + strings.TrimPrefix(value, "fc_")
	}
	return value
}

func deterministicCallID(name, arguments string, index int) string {
	sum := sha1.Sum([]byte(fmt.Sprintf("%s:%s:%d", name, arguments, index)))
	return "call_" + hex.EncodeToString(sum[:])[:16]
}

func stringFromAny(value any) string {
	switch value := value.(type) {
	case string:
		return value
	case nil:
		return ""
	default:
		data, err := json.Marshal(value)
		if err == nil {
			return string(data)
		}
		return fmt.Sprint(value)
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
