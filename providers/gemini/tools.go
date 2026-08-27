package gemini

import (
	"encoding/json"
	"fmt"
	"strings"

	llm "github.com/wsnacj/agentx-go/components/llm"
)

func buildGeminiTools(tools []llm.Tool) ([]map[string]any, error) {
	if len(tools) == 0 {
		return nil, nil
	}
	tools = llm.SanitizeToolSchemas(tools)
	declarations := make([]map[string]any, 0, len(tools))
	for _, tool := range tools {
		kind := strings.ToLower(strings.TrimSpace(tool.Type))
		if kind != "" && kind != "function" {
			return nil, fmt.Errorf("gemini: unsupported tool type %q", kind)
		}
		name := strings.TrimSpace(tool.Function.Name)
		if name == "" {
			return nil, fmt.Errorf("gemini: function name is required")
		}
		declaration := map[string]any{"name": name}
		if description := strings.TrimSpace(tool.Function.Description); description != "" {
			declaration["description"] = description
		}
		if len(tool.Function.Parameters) > 0 {
			// Canonical Tool schemas are standard JSON Schema. Gemini's legacy
			// `parameters` field uses its protobuf Schema enum, so use the
			// dedicated JSON Schema field to preserve lowercase JSON types.
			declaration["parametersJsonSchema"] = tool.Function.Parameters
		}
		if len(tool.Function.OutputSchema) > 0 {
			declaration["responseJsonSchema"] = tool.Function.OutputSchema
		}
		declarations = append(declarations, declaration)
	}
	return []map[string]any{{"functionDeclarations": declarations}}, nil
}

func buildGeminiToolConfig(choice *llm.ToolChoice, tools []llm.Tool) (map[string]any, error) {
	if choice == nil {
		return nil, nil
	}
	kind := strings.ToLower(strings.TrimSpace(choice.Type))
	if kind == "" && choice.Function != nil {
		kind = "function"
	}
	config := map[string]any{}
	switch kind {
	case "", "auto":
		config["mode"] = "AUTO"
	case "none":
		config["mode"] = "NONE"
	case "required":
		if len(tools) == 0 {
			return nil, fmt.Errorf("gemini: required tool choice has no tools")
		}
		config["mode"] = "ANY"
	case "function", "tool":
		if len(tools) == 0 || choice.Function == nil || strings.TrimSpace(choice.Function.Name) == "" {
			return nil, fmt.Errorf("gemini: named tool choice is invalid")
		}
		selected := strings.TrimSpace(choice.Function.Name)
		found := false
		for _, tool := range tools {
			if strings.TrimSpace(tool.Function.Name) == selected {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("gemini: named tool choice is not declared")
		}
		config["mode"] = "ANY"
		config["allowedFunctionNames"] = []string{selected}
	default:
		return nil, fmt.Errorf("gemini: unsupported tool choice %q", kind)
	}
	return map[string]any{"functionCallingConfig": config}, nil
}

func buildMessageContents(messages llm.Conversation) ([]map[string]any, error) {
	contents := make([]map[string]any, 0, len(messages))
	for _, message := range messages {
		role := strings.ToLower(strings.TrimSpace(message.Role))
		parts := make([]map[string]any, 0, 1+len(message.ToolCalls))
		switch role {
		case "assistant", "model":
			if message.Content != "" {
				parts = append(parts, map[string]any{"text": message.Content})
			}
			for _, call := range message.ToolCalls {
				name := strings.TrimSpace(call.Name)
				if name == "" {
					return nil, fmt.Errorf("gemini: assistant function name is required")
				}
				args := map[string]any{}
				if raw := strings.TrimSpace(call.Arguments); raw != "" {
					if err := json.Unmarshal([]byte(raw), &args); err != nil {
						return nil, fmt.Errorf("gemini: decode assistant function arguments: %w", err)
					}
				}
				parts = append(parts, map[string]any{"functionCall": map[string]any{"name": name, "args": args}})
			}
			if len(parts) > 0 {
				contents = append(contents, map[string]any{"role": "model", "parts": parts})
			}
		case "tool":
			name := strings.TrimSpace(message.ToolName)
			if name == "" {
				name = strings.TrimSpace(message.ToolCallID)
			}
			if name == "" {
				return nil, fmt.Errorf("gemini: tool response name is required")
			}
			parts = append(parts, map[string]any{"functionResponse": map[string]any{
				"name": name, "response": geminiFunctionResponse(message.Content),
			}})
			contents = append(contents, map[string]any{"role": "user", "parts": parts})
		default:
			if message.Content != "" {
				contents = append(contents, map[string]any{"role": "user", "parts": []map[string]any{{"text": message.Content}}})
			}
		}
	}
	return contents, nil
}

func geminiFunctionResponse(output string) map[string]any {
	trimmed := strings.TrimSpace(output)
	if trimmed == "" {
		return map[string]any{"output": ""}
	}
	var decoded any
	if err := json.Unmarshal([]byte(trimmed), &decoded); err == nil {
		if object, ok := decoded.(map[string]any); ok {
			return object
		}
		return map[string]any{"output": decoded}
	}
	return map[string]any{"output": output}
}

func extractFunctionCalls(resp *GenerateContentResponse) []llm.FunctionCall {
	if resp == nil || len(resp.Candidates) == 0 {
		return nil
	}
	calls := make([]llm.FunctionCall, 0)
	for _, part := range resp.Candidates[0].Content.Parts {
		if part.FunctionCall == nil || strings.TrimSpace(part.FunctionCall.Name) == "" {
			continue
		}
		argsMap := part.FunctionCall.Args
		if argsMap == nil {
			argsMap = map[string]any{}
		}
		args, err := json.Marshal(argsMap)
		if err != nil {
			continue
		}
		calls = append(calls, llm.FunctionCall{Type: "function", Name: part.FunctionCall.Name, Arguments: string(args)})
	}
	return calls
}
