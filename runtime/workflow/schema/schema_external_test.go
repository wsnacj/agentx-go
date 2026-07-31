package schema_test

import (
	"reflect"
	"strings"
	"testing"

	schema "github.com/wsnacj/agentx-go/runtime/workflow/schema"
)

func TestNormalizeContract(t *testing.T) {
	object := map[string]any{"type": "object"}
	got, err := schema.Normalize(object, "config.schema")
	if err != nil {
		t.Fatalf("Normalize(map): %v", err)
	}
	if !reflect.DeepEqual(got, object) {
		t.Fatalf("Normalize(map) = %#v, want %#v", got, object)
	}

	got, err = schema.Normalize(`{"type":"string","minLength":1}`, "config.schema")
	if err != nil {
		t.Fatalf("Normalize(string): %v", err)
	}
	if got["type"] != "string" || got["minLength"] != float64(1) {
		t.Fatalf("Normalize(string) = %#v", got)
	}

	for _, raw := range []any{map[string]any{}, "", nil} {
		got, err = schema.Normalize(raw, "config.schema")
		if raw == nil {
			if err == nil || err.Error() != "workflow: config.schema must be a JSON object or JSON object string" {
				t.Fatalf("Normalize(nil) error = %v", err)
			}
			continue
		}
		if err != nil || got != nil {
			t.Fatalf("Normalize(%#v) = %#v, %v; want nil, nil", raw, got, err)
		}
	}
}

func TestNormalizePreservesWhitespaceAndDecodeErrors(t *testing.T) {
	tests := []struct {
		name string
		raw  any
		want string
	}{
		{
			name: "whitespace only",
			raw:  "  ",
			want: `workflow: config.schema "  " must not be whitespace-only`,
		},
		{
			name: "surrounding whitespace",
			raw:  ` {"type":"object"} `,
			want: `workflow: config.schema " {\"type\":\"object\"} " must not include surrounding whitespace`,
		},
		{
			name: "non object",
			raw:  []any{"object"},
			want: "workflow: config.schema must be a JSON object or JSON object string",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := schema.Normalize(tc.raw, "config.schema")
			if err == nil || err.Error() != tc.want {
				t.Fatalf("Normalize() error = %v, want %q", err, tc.want)
			}
		})
	}

	_, err := schema.Normalize(`{"type":`, "config.schema")
	if err == nil || !strings.HasPrefix(err.Error(), "workflow: config.schema must be a valid JSON object:") {
		t.Fatalf("Normalize(invalid JSON) error = %v", err)
	}
}

func TestValidateDefinitionAcceptsPortableSchema(t *testing.T) {
	definition := map[string]any{
		"type":          "object",
		"required":      []string{"answer"},
		"minProperties": 1,
		"properties": map[string]any{
			"answer": map[string]any{
				"type":      "string",
				"minLength": 1,
				"pattern":   `^[a-z]+$`,
				"enum":      []string{"pass", "fail"},
			},
			"score": map[string]any{
				"type":    []any{"number", "null"},
				"minimum": 0,
				"maximum": 1,
			},
		},
		"additionalProperties": false,
	}
	if err := schema.ValidateDefinition(definition, "config.schema"); err != nil {
		t.Fatalf("ValidateDefinition(): %v", err)
	}
}

func TestValidateDefinitionErrorContract(t *testing.T) {
	tests := []struct {
		name   string
		schema map[string]any
		want   string
	}{
		{
			name:   "canonical type",
			schema: map[string]any{"type": "STRING"},
			want:   "workflow: config.schema.type: must use canonical lowercase",
		},
		{
			name:   "unsupported type",
			schema: map[string]any{"type": "date"},
			want:   `workflow: config.schema.type: unsupported type "date"`,
		},
		{
			name:   "keyword applicability",
			schema: map[string]any{"type": "string", "required": []string{"answer"}},
			want:   "workflow: config.schema.required requires declared type to include object",
		},
		{
			name:   "duplicate required",
			schema: map[string]any{"type": "object", "required": []string{"answer", "answer"}},
			want:   "workflow: config.schema.required: must not contain duplicate entries",
		},
		{
			name:   "enum type",
			schema: map[string]any{"type": "integer", "enum": []any{1, "one"}},
			want:   "workflow: config.schema.enum: [1]: does not match declared type constraint",
		},
		{
			name:   "const constraint",
			schema: map[string]any{"type": "string", "minLength": 2, "const": "x"},
			want:   "workflow: config.schema.const: violates minLength",
		},
		{
			name:   "range",
			schema: map[string]any{"type": "number", "minimum": 2, "maximum": 1},
			want:   "workflow: config.schema.minimum: must be <= maximum",
		},
		{
			name:   "nested property",
			schema: map[string]any{"type": "object", "properties": map[string]any{"answer": "string"}},
			want:   "workflow: config.schema.properties.answer: must be an object",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := schema.ValidateDefinition(tc.schema, "config.schema")
			if err == nil || err.Error() != tc.want {
				t.Fatalf("ValidateDefinition() error = %v, want %q", err, tc.want)
			}
		})
	}
}
