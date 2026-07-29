package llm

import "strings"

// SanitizeToolSchemas returns a copied tool list with backend-compatible
// function parameter schemas. It does not alter OutputSchema, which is an
// internal AgentX contract and is not sent to OpenAI-compatible tool payloads.
func SanitizeToolSchemas(tools []Tool) []Tool {
	if len(tools) == 0 {
		return tools
	}
	out := make([]Tool, len(tools))
	for i, tool := range tools {
		out[i] = tool
		out[i].Function.Parameters = SanitizeFunctionParametersSchema(tool.Function.Parameters)
	}
	return out
}

func SanitizeFunctionParametersSchema(schema map[string]any) map[string]any {
	out := sanitizeJSONSchemaMap(schema)
	out["type"] = "object"
	properties, ok := schemaMap(out["properties"])
	if !ok {
		properties = map[string]any{}
	}
	out["properties"] = properties
	if required := sanitizeRequiredFields(out["required"], properties); len(required) > 0 {
		out["required"] = required
	} else {
		delete(out, "required")
	}
	return out
}

func sanitizeJSONSchemaMap(schema map[string]any) map[string]any {
	out := map[string]any{}
	for key, value := range schema {
		if strings.TrimSpace(key) == "" {
			continue
		}
		switch key {
		case "properties", "required", "enum", "const", "type", "nullable":
			out[key] = cloneJSONSchemaValue(value)
		default:
			out[key] = sanitizeJSONSchemaValue(value)
		}
	}
	if schemaType, ok := sanitizeJSONSchemaType(out["type"]); ok {
		out["type"] = schemaType
	} else {
		delete(out, "type")
	}
	delete(out, "nullable")
	if props, ok := schemaMap(out["properties"]); ok {
		cleaned := make(map[string]any, len(props))
		for name, value := range props {
			if strings.TrimSpace(name) == "" {
				continue
			}
			if nested, ok := schemaMap(value); ok {
				cleaned[name] = sanitizeJSONSchemaMap(nested)
			} else {
				cleaned[name] = value
			}
		}
		out["properties"] = cleaned
	}
	if props, ok := schemaMap(out["properties"]); ok {
		if required := sanitizeRequiredFields(out["required"], props); len(required) > 0 {
			out["required"] = required
		} else {
			delete(out, "required")
		}
	}
	return out
}

func sanitizeJSONSchemaValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return sanitizeJSONSchemaMap(typed)
	case []any:
		out := make([]any, len(typed))
		for i, item := range typed {
			out[i] = sanitizeJSONSchemaValue(item)
		}
		return out
	default:
		return typed
	}
}

func cloneJSONSchemaValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(typed))
		for key, item := range typed {
			out[key] = cloneJSONSchemaValue(item)
		}
		return out
	case []any:
		out := make([]any, len(typed))
		for i, item := range typed {
			out[i] = cloneJSONSchemaValue(item)
		}
		return out
	case []string:
		out := make([]string, len(typed))
		copy(out, typed)
		return out
	default:
		return typed
	}
}

func sanitizeJSONSchemaType(value any) (any, bool) {
	switch typed := value.(type) {
	case string:
		item := strings.ToLower(strings.TrimSpace(typed))
		if item == "" || item == "null" {
			return nil, false
		}
		return item, true
	case []string:
		return sanitizeJSONSchemaTypeStrings(typed)
	case []any:
		items := make([]string, 0, len(typed))
		for _, item := range typed {
			if text, ok := item.(string); ok {
				items = append(items, text)
			}
		}
		return sanitizeJSONSchemaTypeStrings(items)
	default:
		return nil, false
	}
}

func sanitizeJSONSchemaTypeStrings(items []string) (any, bool) {
	for _, item := range items {
		normalized := strings.ToLower(strings.TrimSpace(item))
		if normalized == "" || normalized == "null" {
			continue
		}
		return normalized, true
	}
	return nil, false
}

func sanitizeRequiredFields(value any, properties map[string]any) []string {
	if len(properties) == 0 {
		return nil
	}
	seen := map[string]bool{}
	out := make([]string, 0)
	appendName := func(name string) {
		name = strings.TrimSpace(name)
		if name == "" || seen[name] {
			return
		}
		if _, ok := properties[name]; !ok {
			return
		}
		seen[name] = true
		out = append(out, name)
	}
	switch typed := value.(type) {
	case []string:
		for _, item := range typed {
			appendName(item)
		}
	case []any:
		for _, item := range typed {
			if text, ok := item.(string); ok {
				appendName(text)
			}
		}
	}
	return out
}

func schemaMap(value any) (map[string]any, bool) {
	switch typed := value.(type) {
	case map[string]any:
		return typed, true
	default:
		return nil, false
	}
}
