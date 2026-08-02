package openaicompat

import (
	"strings"

	llm "github.com/wsnacj/agentx-go/components/llm"
)

type profile struct {
	maxTokensField                   string
	supportsUsageInStreaming         bool
	supportsReasoningEffort          bool
	supportsParallelToolCalls        bool
	reasoningFormat                  string
	supportsStrictMode               bool
	requiresToolResultName           bool
	requiresAssistantAfterToolResult bool
	requiresToolCallIDForToolRole    bool
}

func newProfile(name string) profile {
	name = strings.ToLower(strings.TrimSpace(name))
	p := profile{
		maxTokensField: "max_completion_tokens", supportsUsageInStreaming: true,
		supportsReasoningEffort: true, supportsParallelToolCalls: true,
		reasoningFormat: "openai", supportsStrictMode: true,
		requiresToolCallIDForToolRole: true,
	}
	if name == "azure-openai" {
		p.maxTokensField = "max_tokens"
	} else if name == "openrouter" {
		p.reasoningFormat = "openrouter"
	}
	return p
}

func (p profile) applyTooling(payload map[string]any, tools []llm.Tool, choice *llm.ToolChoice) {
	if serialized := p.serializeTools(tools); len(serialized) > 0 {
		payload["tools"] = serialized
	}
	if serialized := serializeToolChoice(choice); serialized != nil {
		payload["tool_choice"] = serialized
	}
}

func serializeToolChoice(choice *llm.ToolChoice) any {
	if choice == nil {
		return nil
	}
	kind := choice.Type
	if kind == "" && choice.Function != nil {
		kind = "function"
	}
	switch kind {
	case "", "auto", "none", "required":
		return kind
	case "function":
		if choice.Function != nil && choice.Function.Name != "" {
			return map[string]any{"type": "function", "function": map[string]any{"name": choice.Function.Name}}
		}
		return map[string]any{"type": "function"}
	default:
		return kind
	}
}

func (p profile) mergeOptions(payload map[string]any, capability Capability, options map[string]any) {
	if options == nil {
		options = map[string]any{}
	}
	if capability.Thinking {
		if enabled, ok := options["thinking"].(bool); ok {
			mode := "disabled"
			if enabled {
				mode = "enabled"
			}
			payload["thinking"] = map[string]string{"type": mode}
			payload["enable_thinking"] = enabled
			if _, exists := payload["stream"]; !exists {
				payload["stream"] = enabled
			}
		}
	}
	if effort, ok := stringOption(options, "reasoning_effort"); ok && !optionDisabled(capability.ReasoningEffort) && p.supportsReasoningEffort {
		if p.reasoningFormat == "openrouter" {
			if _, exists := options["reasoning"]; !exists {
				payload["reasoning"] = map[string]any{"effort": effort}
			}
		} else {
			payload["reasoning_effort"] = effort
		}
	}
	if parallel, ok := options["parallel_tool_calls"].(bool); ok && !optionDisabled(capability.ParallelToolCalls) && p.supportsParallelToolCalls {
		payload["parallel_tool_calls"] = parallel
	}
	for key, value := range options {
		switch key {
		case "thinking", "max_tokens", "max_completion_tokens", "reasoning_effort", "parallel_tool_calls":
		default:
			payload[key] = value
		}
	}
}

func (p profile) applyMaxTokens(payload map[string]any, fallback int, options map[string]any) {
	field := p.maxTokensField
	if field == "" {
		field = "max_completion_tokens"
	}
	if value, ok := intOption(options, field); ok {
		payload[field] = value
		return
	}
	alternate := "max_tokens"
	if field == "max_tokens" {
		alternate = "max_completion_tokens"
	}
	if value, ok := intOption(options, alternate); ok {
		payload[field] = value
		return
	}
	if fallback > 0 {
		payload[field] = fallback
	}
}

func (p profile) applyStreamingDefaults(payload map[string]any) {
	stream, _ := payload["stream"].(bool)
	if !p.supportsUsageInStreaming || !stream {
		return
	}
	if _, exists := payload["stream_options"]; !exists {
		payload["stream_options"] = map[string]any{"include_usage": true}
	}
}

func (p profile) mergeEmbeddingOptions(payload map[string]any, cfg EmbeddingConfig, options map[string]any) {
	if value, ok := options["dimensions"]; ok && !optionDisabled(cfg.Capability.EmbeddingDimensions) {
		payload["dimensions"] = value
	}
	if value, ok := options["encoding_format"]; ok {
		if !(strings.EqualFold(asString(value), "base64") && optionDisabled(cfg.Capability.EmbeddingEncodingBase64)) {
			payload["encoding_format"] = value
		}
	} else if value, ok := options["encoding"]; ok {
		if _, exists := payload["encoding_format"]; !exists && !(strings.EqualFold(asString(value), "base64") && optionDisabled(cfg.Capability.EmbeddingEncodingBase64)) {
			payload["encoding_format"] = value
		}
	}
	if value, ok := options["instructions"]; ok && !optionDisabled(cfg.Capability.EmbeddingInstructions) {
		payload["instructions"] = value
	}
	if value, ok := options["sparse_embedding"]; ok && !optionDisabled(cfg.Capability.SparseEmbedding) {
		payload["sparse_embedding"] = normalizeSparse(value)
	}
}

func (p profile) serializeTools(tools []llm.Tool) []map[string]any {
	if len(tools) == 0 {
		return nil
	}
	tools = llm.SanitizeToolSchemas(tools)
	out := make([]map[string]any, 0, len(tools))
	for _, tool := range tools {
		kind := strings.TrimSpace(tool.Type)
		if kind == "" {
			kind = "function"
		}
		function := map[string]any{"name": tool.Function.Name}
		if value := strings.TrimSpace(tool.Function.Description); value != "" {
			function["description"] = value
		}
		if len(tool.Function.Parameters) > 0 {
			function["parameters"] = tool.Function.Parameters
		}
		if p.supportsStrictMode {
			if tool.Function.Strict != nil {
				function["strict"] = *tool.Function.Strict
			} else {
				function["strict"] = false
			}
		}
		out = append(out, map[string]any{"type": kind, "function": function})
	}
	return out
}

func optionDisabled(value *bool) bool { return value != nil && !*value }
func stringOption(values map[string]any, key string) (string, bool) {
	value, ok := values[key].(string)
	value = strings.TrimSpace(value)
	return value, ok && value != ""
}
func asString(value any) string { text, _ := value.(string); return text }
func normalizeSparse(value any) any {
	if enabled, ok := value.(bool); ok {
		mode := "disabled"
		if enabled {
			mode = "enabled"
		}
		return map[string]string{"type": mode}
	}
	if text, ok := value.(string); ok && (text == "enabled" || text == "disabled") {
		return map[string]string{"type": text}
	}
	return value
}
func intOption(values map[string]any, key string) (int, bool) {
	value, exists := values[key]
	if !exists {
		return 0, false
	}
	switch value := value.(type) {
	case int:
		return value, true
	case int8:
		return int(value), true
	case int16:
		return int(value), true
	case int32:
		return int(value), true
	case int64:
		return int(value), true
	case uint:
		return int(value), true
	case uint8:
		return int(value), true
	case uint16:
		return int(value), true
	case uint32:
		return int(value), true
	case uint64:
		return int(value), true
	case float32:
		return int(value), true
	case float64:
		return int(value), true
	default:
		return 0, false
	}
}
