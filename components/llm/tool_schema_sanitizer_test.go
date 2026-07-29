package llm

import "testing"

func TestSanitizeToolSchemasNormalizesFunctionParameters(t *testing.T) {
	tools := []Tool{{
		Type: "function",
		Function: Function{
			Name: "lookup",
			Parameters: map[string]any{
				"type":     []any{"object", "null"},
				"nullable": true,
				"properties": map[string]any{
					"query": map[string]any{
						"type":     []any{"string", "null"},
						"nullable": true,
					},
					"type": map[string]any{
						"type": []string{"string", "null"},
					},
				},
				"required": []any{"query", "missing", "query", 7},
			},
			OutputSchema: map[string]any{"type": "object"},
		},
	}}

	sanitized := SanitizeToolSchemas(tools)
	params := sanitized[0].Function.Parameters
	if got := params["type"]; got != "object" {
		t.Fatalf("expected top-level object schema, got %#v", params)
	}
	if _, ok := params["nullable"]; ok {
		t.Fatalf("expected nullable to be removed from top-level schema, got %#v", params)
	}
	props, ok := params["properties"].(map[string]any)
	if !ok {
		t.Fatalf("expected properties object, got %#v", params["properties"])
	}
	query, ok := props["query"].(map[string]any)
	if !ok || query["type"] != "string" {
		t.Fatalf("expected nullable query union to collapse to string, got %#v", props["query"])
	}
	if _, ok := query["nullable"]; ok {
		t.Fatalf("expected nested nullable to be removed, got %#v", query)
	}
	typeProp, ok := props["type"].(map[string]any)
	if !ok || typeProp["type"] != "string" {
		t.Fatalf("expected property named type to remain a property schema, got %#v", props["type"])
	}
	required, ok := params["required"].([]string)
	if !ok || len(required) != 1 || required[0] != "query" {
		t.Fatalf("expected required fields filtered and deduped, got %#v", params["required"])
	}
	if sanitized[0].Function.OutputSchema["type"] != "object" {
		t.Fatalf("expected output schema to remain untouched, got %#v", sanitized[0].Function.OutputSchema)
	}
	if _, ok := tools[0].Function.Parameters["nullable"]; !ok {
		t.Fatalf("expected original tool schema to remain unchanged")
	}
}

func TestSanitizeToolSchemasDefaultsMissingParametersToObject(t *testing.T) {
	sanitized := SanitizeToolSchemas([]Tool{{
		Type:     "function",
		Function: Function{Name: "empty"},
	}})

	params := sanitized[0].Function.Parameters
	if params["type"] != "object" {
		t.Fatalf("expected default object schema, got %#v", params)
	}
	props, ok := params["properties"].(map[string]any)
	if !ok || len(props) != 0 {
		t.Fatalf("expected empty properties object, got %#v", params["properties"])
	}
}
