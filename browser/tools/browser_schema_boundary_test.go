package tools

import (
	"strings"
	"testing"

	types "github.com/wsnacj/agentx-go/components/llm"
)

func TestBrowserToolSchemasDoNotExposePlaywrightOnlyLocatorDSL(t *testing.T) {
	for _, def := range []types.Tool{
		browserActDefinition([]string{"snapshot", "click", "type", "screenshot"}),
		browserDefinition([]string{"status", "prepare", "coordinate"}, []string{"snapshot", "click", "type", "screenshot"}),
	} {
		keys := browserSchemaPropertyKeys(def.Function.Parameters)
		for _, forbidden := range []string{
			"get_by_role",
			"get_by_text",
			"get_by_label",
			"get_by_placeholder",
			"get_by_alt_text",
			"get_by_title",
			"get_by_test_id",
			"test_id",
			"data_test_id",
			"locator",
			"locator_chain",
			"has",
			"has_not",
			"has_text",
			"has_not_text",
			"nth",
			"exact",
		} {
			if keys[forbidden] {
				t.Fatalf("%s schema exposed Playwright-only locator field %q", def.Function.Name, forbidden)
			}
		}
	}
}

func TestBrowserToolSchemasExposeBackendNeutralSemanticTargetHints(t *testing.T) {
	for _, def := range []types.Tool{
		browserActDefinition([]string{"snapshot", "click", "type", "screenshot"}),
		browserDefinition([]string{"status", "prepare", "coordinate"}, []string{"snapshot", "click", "type", "screenshot"}),
		browserClickDefinition(),
	} {
		props := browserSchemaProperties(def.Function.Parameters)
		for _, field := range []string{"element", "label", "text"} {
			prop, ok := props[field]
			if !ok {
				t.Fatalf("%s schema missing backend-neutral semantic target hint %q", def.Function.Name, field)
			}
			if got := prop["type"]; got != "string" {
				t.Fatalf("%s schema semantic hint %q should be a string, got %#v", def.Function.Name, field, prop)
			}
		}
	}
}

func browserSchemaProperties(value any) map[string]map[string]any {
	params, ok := value.(map[string]any)
	if !ok {
		return nil
	}
	raw, ok := params["properties"].(map[string]any)
	if !ok {
		return nil
	}
	out := make(map[string]map[string]any, len(raw))
	for key, value := range raw {
		prop, ok := value.(map[string]any)
		if !ok {
			continue
		}
		out[strings.ToLower(strings.TrimSpace(key))] = prop
	}
	return out
}

func browserSchemaPropertyKeys(value any) map[string]bool {
	out := map[string]bool{}
	var walk func(any)
	walk = func(current any) {
		switch typed := current.(type) {
		case map[string]any:
			if props, ok := typed["properties"].(map[string]any); ok {
				for key, nested := range props {
					out[strings.ToLower(strings.TrimSpace(key))] = true
					walk(nested)
				}
			}
			for _, nested := range typed {
				walk(nested)
			}
		case []any:
			for _, nested := range typed {
				walk(nested)
			}
		case []map[string]any:
			for _, nested := range typed {
				walk(nested)
			}
		}
	}
	walk(value)
	return out
}
