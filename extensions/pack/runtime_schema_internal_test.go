package pack

import (
	"strings"
	"testing"

	agentxworkflow "github.com/wsnacj/agentx-go/runtime/workflow"
)

func TestValidateSemanticToolInputSchemaRejectsMalformedSchemaDefinitionAtRuntime(t *testing.T) {
	err := validateSemanticToolInputSchema(agentxworkflow.NodeSpec{
		ID:   "tool_node",
		Kind: agentxworkflow.NodeTool,
		Config: map[string]any{
			"tool": "browser_act",
		},
	}, map[string]any{
		"tool": "browser_act",
	}, map[string]any{
		"type": " object ",
	}, testToolArgumentLowerer())
	if err == nil {
		t.Fatalf("expected malformed semantic tool schema to fail")
	}
	if !strings.Contains(err.Error(), "semantic_tool_input.type") || !strings.Contains(err.Error(), "surrounding whitespace") {
		t.Fatalf("unexpected runtime semantic tool schema error: %v", err)
	}
}

func TestValidateSemanticToolInputSchemaRejectsTypeSliceMismatchAtRuntime(t *testing.T) {
	err := validateSemanticToolInputSchema(agentxworkflow.NodeSpec{
		ID:   "tool_node",
		Kind: agentxworkflow.NodeTool,
	}, map[string]any{
		"tool": "browser_act",
		"args": map[string]any{
			"score": "bad",
		},
	}, map[string]any{
		"type": []string{"object"},
		"properties": map[string]any{
			"score": map[string]any{
				"type": []string{"number"},
			},
		},
		"required": []string{"score"},
	}, testToolArgumentLowerer())
	if err == nil {
		t.Fatalf("expected semantic tool input type-slice mismatch to fail")
	}
	if !strings.Contains(err.Error(), "semantic_tool_input.score") || !strings.Contains(err.Error(), "declared type") {
		t.Fatalf("unexpected runtime semantic tool schema error: %v", err)
	}
}

func TestSemanticToolSchemaPlaceholderSupportsTypeSlice(t *testing.T) {
	placeholder := semanticToolSchemaPlaceholder(map[string]any{
		"type": []string{"integer"},
	})
	value, ok := placeholder.(int)
	if !ok || value != 0 {
		t.Fatalf("expected integer placeholder for type-slice schema, got %#v", placeholder)
	}
}

func TestPackSchemaTypeHelpersRejectNonCanonicalTypeLabels(t *testing.T) {
	cases := []string{" string ", "STRING"}
	for _, schemaType := range cases {
		if packSchemaValueMatchesType("ok", schemaType) {
			t.Fatalf("expected pack schema helper to reject non-canonical type %q", schemaType)
		}
		if isSupportedPackSchemaType(schemaType) {
			t.Fatalf("expected pack schema whitelist to reject non-canonical type %q", schemaType)
		}
	}
}

func TestRuntimePackSchemaTypesIgnoresMalformedRequiredShape(t *testing.T) {
	types := runtimePackSchemaTypes(map[string]any{
		"required": []string{" value "},
	})
	if len(types) != 0 {
		t.Fatalf("expected malformed required shape not to infer object type, got %#v", types)
	}
}

func TestPackMaterializedPayloadPathHelpersRejectMalformedPath(t *testing.T) {
	if got := normalizeSemanticToolInputBindingTarget(" args.value"); got != " args.value" {
		t.Fatalf("expected target helper to stop trimming malformed path, got %q", got)
	}
	if got := normalizeSemanticToolInputBindingTarget("args.value"); got != "value" {
		t.Fatalf("expected canonical args target to strip args prefix, got %q", got)
	}
	for _, path := range []string{" args.value", "args. value", "args..value"} {
		if parts := splitMaterializedPayloadPath(path); parts != nil {
			t.Fatalf("expected malformed path %q to be rejected, got %#v", path, parts)
		}
	}
	if parts := splitMaterializedPayloadPath("args.value"); len(parts) != 2 || parts[0] != "args" || parts[1] != "value" {
		t.Fatalf("expected canonical path to split cleanly, got %#v", parts)
	}
}

func TestExtractArgumentContainerRejectsWhitespaceOnlyAndTrimDependentJSON(t *testing.T) {
	if _, _, _, err := extractArgumentContainer(map[string]any{"args": "   "}); err == nil || !strings.Contains(err.Error(), "whitespace-only") {
		t.Fatalf("expected whitespace-only args to fail, got %v", err)
	}
	if _, _, _, err := extractArgumentContainer(map[string]any{"arguments_json": " {\"ok\":true} "}); err == nil || !strings.Contains(err.Error(), "surrounding whitespace") {
		t.Fatalf("expected trim-dependent arguments_json to fail, got %v", err)
	}
}

func TestWorkflowToolNameStopsTrimmingLookupValue(t *testing.T) {
	if got := workflowToolName(map[string]any{"tool": " browser_act "}); got != " browser_act " {
		t.Fatalf("expected workflow tool name helper to preserve raw value, got %q", got)
	}
}
