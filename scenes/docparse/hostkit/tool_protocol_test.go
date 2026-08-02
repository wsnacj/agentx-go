package hostkit

import (
	"strings"
	"testing"

	types "github.com/wsnacj/agentx-go/components/llm"
)

func TestToolNamesExposeStableDocparseSurface(t *testing.T) {
	names := ToolNames()
	for _, want := range []string{
		ToolDocparseSpecSelect,
		ToolDocparseProfileProbe,
		ToolDocparseExtractFields,
		ToolDocparseExtractTable,
		ToolDocparseTraceEvidence,
		ToolDocparseValidate,
		ToolDocparseGuard,
	} {
		if !stringSliceContains(names, want) {
			t.Fatalf("ToolNames() = %#v, want %q", names, want)
		}
	}
}

func TestToolDefinitionsUseFunctionTools(t *testing.T) {
	for _, tool := range []struct {
		name string
		def  func() string
	}{
		{name: ToolDocparseSpecSelect, def: func() string { return DocparseSpecSelectTool().Function.Name }},
		{name: ToolDocparseProfileProbe, def: func() string { return DocparseProfileProbeTool().Function.Name }},
		{name: ToolDocparseExtractFields, def: func() string { return DocparseExtractFieldsTool().Function.Name }},
		{name: ToolDocparseExtractTable, def: func() string { return DocparseExtractTableTool().Function.Name }},
	} {
		if got := tool.def(); got != tool.name {
			t.Fatalf("unexpected tool definition for %q", tool.name)
		}
	}
}

func TestToolDefinitionsDescribeParameters(t *testing.T) {
	for _, tool := range []types.Tool{
		DocparseSpecSelectTool(),
		DocparseProfileProbeTool(),
		DocparseExtractFieldsTool(),
		DocparseExtractTableTool(),
		DocparseTraceEvidenceTool(),
		DocparseValidateTool(),
		DocparseGuardTool(),
	} {
		if strings.TrimSpace(tool.Function.Description) == "" {
			t.Fatalf("%s missing description", tool.Function.Name)
		}
		assertToolSchemaDescriptions(t, tool.Function.Name, tool.Function.Parameters, "")
	}
}

func TestExtractionToolDefinitionsAcceptExistingParseResults(t *testing.T) {
	for _, tool := range []types.Tool{
		DocparseExtractFieldsTool(),
		DocparseExtractTableTool(),
	} {
		params := tool.Function.Parameters
		props, _ := params["properties"].(map[string]any)
		for _, want := range []string{"document_path", "result_path", "parse_result"} {
			if _, ok := props[want]; !ok {
				t.Fatalf("%s missing property %q", tool.Function.Name, want)
			}
		}
		for _, required := range readStringArrayFromSchema(params["required"]) {
			if required == "document_path" {
				t.Fatalf("%s should not force document_path when result_path/parse_result is supported", tool.Function.Name)
			}
		}
	}
}

func assertToolSchemaDescriptions(t *testing.T, owner string, schema map[string]any, prefix string) {
	t.Helper()
	props, _ := schema["properties"].(map[string]any)
	for name, raw := range props {
		path := name
		if prefix != "" {
			path = prefix + "." + name
		}
		prop, ok := raw.(map[string]any)
		if !ok {
			t.Fatalf("%s property %s should be an object schema", owner, path)
		}
		if strings.TrimSpace(readToolSchemaString(prop["description"])) == "" {
			t.Fatalf("%s property %s missing description", owner, path)
		}
		if nested, _ := prop["properties"].(map[string]any); len(nested) != 0 {
			assertToolSchemaDescriptions(t, owner, prop, path)
		}
		if items, _ := prop["items"].(map[string]any); items != nil {
			if nested, _ := items["properties"].(map[string]any); len(nested) != 0 {
				assertToolSchemaDescriptions(t, owner, items, path+"[]")
			}
		}
	}
}

func readToolSchemaString(value any) string {
	out, _ := value.(string)
	return out
}

func readStringArrayFromSchema(value any) []string {
	values, _ := value.([]string)
	if values != nil {
		return values
	}
	raw, _ := value.([]any)
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		if text, ok := item.(string); ok {
			out = append(out, text)
		}
	}
	return out
}

func stringSliceContains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
