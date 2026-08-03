package videoframes

func stringEnumSchema(description string, values ...string) map[string]any {
	return map[string]any{"type": "string", "description": description, "enum": append([]string(nil), values...)}
}

func stringArraySchema(description string) map[string]any {
	return map[string]any{"type": "array", "description": description, "items": map[string]any{"type": "string"}}
}

func looseObjectSchema(description string) map[string]any {
	return map[string]any{"type": "object", "description": description, "additionalProperties": true}
}

func objectArraySchema(description string, properties map[string]any) map[string]any {
	return map[string]any{"type": "array", "description": description, "items": map[string]any{"type": "object", "additionalProperties": false, "properties": properties}}
}

func closedOutputSchema(properties map[string]any, required []string) map[string]any {
	out := map[string]any{"type": "object", "additionalProperties": false, "properties": properties}
	if len(required) > 0 {
		out["required"] = append([]string(nil), required...)
	}
	return out
}
