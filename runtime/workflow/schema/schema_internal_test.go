package schema

import "testing"

func TestWorkflowSchemaValueMatchesTypeRejectsNonCanonicalTypeLabels(t *testing.T) {
	cases := []string{" string ", "STRING"}
	for _, schemaType := range cases {
		if workflowSchemaValueMatchesType("ok", schemaType) {
			t.Fatalf("expected workflow schema helper to reject non-canonical type %q", schemaType)
		}
		if isSupportedWorkflowSchemaType(schemaType) {
			t.Fatalf("expected workflow schema type whitelist to reject non-canonical type %q", schemaType)
		}
	}
}
