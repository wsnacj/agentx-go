package tools

import (
	"strings"

	types "github.com/wsnacj/agentx-go/components/llm"
)

// SanitizeToolDefinitionsForBackendCompatibility normalizes tool schemas before
// they reach model/provider request assembly. It keeps the canonical tool owner
// unchanged while making the final exposed schema shape backend-friendly.
func SanitizeToolDefinitionsForBackendCompatibility(defs []types.Tool) []types.Tool {
	if len(defs) == 0 {
		return nil
	}
	out := make([]types.Tool, 0, len(defs))
	for _, def := range defs {
		out = append(out, SanitizeToolDefinitionForBackendCompatibility(def))
	}
	return out
}

func SanitizeToolDefinitionForBackendCompatibility(def types.Tool) types.Tool {
	if strings.TrimSpace(def.Type) == "" {
		def.Type = "function"
	}
	def.Function.Parameters = sanitizeToolParameterSchema(def.Function.Parameters)
	if len(def.Function.OutputSchema) > 0 {
		def.Function.OutputSchema = sanitizeToolJSONSchema(def.Function.OutputSchema, false)
	}
	return def
}

func sanitizeToolParameterSchema(schema map[string]any) map[string]any {
	return sanitizeToolJSONSchema(schema, true)
}

func sanitizeToolJSONSchema(schema map[string]any, forceObjectRoot bool) map[string]any {
	out := sanitizeToolJSONSchemaMap(schema)
	if forceObjectRoot {
		out["type"] = "object"
		if _, ok := out["properties"].(map[string]any); !ok {
			out["properties"] = map[string]any{}
		}
	}
	if schemaTypeIsObject(out["type"]) {
		if _, ok := out["properties"].(map[string]any); !ok {
			out["properties"] = map[string]any{}
		}
	}
	sanitizeToolRequiredFields(out)
	return out
}

func sanitizeToolJSONSchemaMap(schema map[string]any) map[string]any {
	out := make(map[string]any, len(schema)+2)
	for key, value := range schema {
		if strings.TrimSpace(key) == "" {
			continue
		}
		if key == "type" {
			if sanitized, ok := sanitizeToolSchemaType(value); ok {
				out[key] = sanitized
			}
			continue
		}
		if key == "properties" {
			out[key] = sanitizeToolSchemaProperties(value)
			continue
		}
		out[key] = sanitizeToolJSONSchemaValue(value)
	}
	if props, ok := out["properties"].(map[string]any); ok {
		for name, prop := range props {
			props[name] = sanitizeToolJSONSchemaValue(prop)
		}
	}
	return out
}

func sanitizeToolSchemaProperties(value any) any {
	props, ok := value.(map[string]any)
	if !ok {
		return value
	}
	out := make(map[string]any, len(props))
	for name, prop := range props {
		if strings.TrimSpace(name) == "" {
			continue
		}
		out[name] = sanitizeToolJSONSchemaValue(prop)
	}
	return out
}

func sanitizeToolJSONSchemaValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return sanitizeToolJSONSchema(typed, false)
	case []any:
		out := make([]any, 0, len(typed))
		for _, item := range typed {
			out = append(out, sanitizeToolJSONSchemaValue(item))
		}
		return out
	case []map[string]any:
		out := make([]any, 0, len(typed))
		for _, item := range typed {
			out = append(out, sanitizeToolJSONSchema(item, false))
		}
		return out
	default:
		return value
	}
}

func sanitizeToolSchemaType(value any) (any, bool) {
	switch typed := value.(type) {
	case string:
		trimmed := strings.TrimSpace(typed)
		if trimmed == "" {
			return nil, false
		}
		return trimmed, true
	case []any:
		return sanitizeToolSchemaTypeArray(typed)
	case []string:
		items := make([]any, 0, len(typed))
		for _, item := range typed {
			items = append(items, item)
		}
		return sanitizeToolSchemaTypeArray(items)
	default:
		return value, true
	}
}

func sanitizeToolSchemaTypeArray(items []any) (any, bool) {
	seen := map[string]bool{}
	nonNull := make([]string, 0, len(items))
	for _, item := range items {
		text, ok := item.(string)
		if !ok {
			continue
		}
		normalized := strings.TrimSpace(text)
		if normalized == "" || strings.EqualFold(normalized, "null") {
			continue
		}
		if seen[normalized] {
			continue
		}
		seen[normalized] = true
		nonNull = append(nonNull, normalized)
	}
	switch len(nonNull) {
	case 0:
		return nil, false
	case 1:
		return nonNull[0], true
	default:
		out := make([]any, 0, len(nonNull))
		for _, item := range nonNull {
			out = append(out, item)
		}
		return out, true
	}
}

func schemaTypeIsObject(value any) bool {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed) == "object"
	case []any:
		for _, item := range typed {
			if text, ok := item.(string); ok && strings.TrimSpace(text) == "object" {
				return true
			}
		}
	case []string:
		for _, item := range typed {
			if strings.TrimSpace(item) == "object" {
				return true
			}
		}
	}
	return false
}

func sanitizeToolRequiredFields(schema map[string]any) {
	raw, ok := schema["required"]
	if !ok {
		return
	}
	properties, hasProperties := schema["properties"].(map[string]any)
	allowed := map[string]bool{}
	for name := range properties {
		allowed[name] = true
	}
	fields := stringListFromAny(raw)
	if len(fields) == 0 {
		delete(schema, "required")
		return
	}
	seen := map[string]bool{}
	out := make([]any, 0, len(fields))
	for _, field := range fields {
		trimmed := strings.TrimSpace(field)
		if trimmed == "" || seen[trimmed] {
			continue
		}
		if hasProperties && !allowed[trimmed] {
			continue
		}
		seen[trimmed] = true
		out = append(out, trimmed)
	}
	if len(out) == 0 {
		delete(schema, "required")
		return
	}
	schema["required"] = out
}

func stringListFromAny(value any) []string {
	switch typed := value.(type) {
	case []string:
		return append([]string(nil), typed...)
	case []any:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			if text, ok := item.(string); ok {
				out = append(out, text)
			}
		}
		return out
	default:
		return nil
	}
}
