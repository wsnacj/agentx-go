package tools

import (
	"testing"

	types "github.com/wsnacj/agentx-go/components/llm"
)

func TestBuiltinBrowserActionSchemaFieldContractsCoverRegisteredSchemas(t *testing.T) {
	contracts := BuiltinBrowserActionSchemaFieldContracts()
	if len(contracts) == 0 {
		t.Fatal("expected builtin browser action schema field contracts")
	}
	if findings := AuditBrowserActionSchemaFieldContracts(contracts); len(findings) > 0 {
		t.Fatalf("expected builtin browser action schema field contracts to pass audit, got %+v", findings)
	}

	for _, def := range []types.Tool{
		browserDefinition([]string{"status", "prepare", "coordinate", "profiles", "sessions"}, fullBrowserActKinds),
		browserActDefinition(fullBrowserActKinds),
		browserOpenDefinition(),
		browserNavigateDefinition(),
		browserTabsDefinition(),
		browserExtractDefinition(),
		browserScreenshotDefinition(),
		browserClickDefinition(),
		browserTypeDefinition(),
		browserEvalDefinition(),
	} {
		if findings := AuditBrowserActionToolSchema(def, contracts); len(findings) > 0 {
			t.Fatalf("%s schema should be fully covered by backend-neutral contracts, got %+v", def.Function.Name, findings)
		}
	}
}

func TestAuditBrowserActionToolSchemaRejectsUnregisteredField(t *testing.T) {
	def := types.Tool{
		Type: "function",
		Function: types.Function{
			Name: "browser_act",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"kind":                 map[string]any{"type": "string"},
					"aria_snapshot_filter": map[string]any{"type": "string"},
				},
			},
		},
	}
	findings := AuditBrowserActionToolSchema(def, BuiltinBrowserActionSchemaFieldContracts())
	if !browserActionSchemaFieldFindingExists(findings, "browser_act", "aria_snapshot_filter", "browser_action_schema_field_contract_required") {
		t.Fatalf("expected unregistered field contract rejection, got %+v", findings)
	}
}

func TestAuditBrowserActionToolSchemaRejectsPlaywrightLocatorDSLFields(t *testing.T) {
	def := types.Tool{
		Type: "function",
		Function: types.Function{
			Name: "browser",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"action":        map[string]any{"type": "string"},
					"get_by_role":   map[string]any{"type": "string"},
					"locator_chain": map[string]any{"type": "array"},
					"target": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"has_text": map[string]any{"type": "string"},
							"nth":      map[string]any{"type": "integer"},
						},
					},
				},
			},
		},
	}
	findings := AuditBrowserActionToolSchema(def, BuiltinBrowserActionSchemaFieldContracts())
	for _, field := range []string{"get_by_role", "locator_chain", "has_text", "nth"} {
		if !browserActionSchemaFieldFindingExists(findings, "browser", field, "browser_action_schema_playwright_locator_field_forbidden") {
			t.Fatalf("expected Playwright locator DSL rejection for %s, got %+v", field, findings)
		}
	}
}

func TestAuditBrowserActionSchemaFieldContractsRejectsPlaywrightOnlyContract(t *testing.T) {
	findings := AuditBrowserActionSchemaFieldContracts([]BrowserActionSchemaFieldContract{
		{
			Field:                "get_by_role",
			Owner:                "tools/browser_action_schema",
			Kind:                 BrowserActionSchemaFieldKindTarget,
			BackendNeutral:       true,
			LLMActionSchema:      true,
			PlaywrightLocatorDSL: true,
		},
	})
	if !browserActionSchemaFieldFindingExists(findings, "", "get_by_role", "browser_action_schema_playwright_locator_field_forbidden") {
		t.Fatalf("expected Playwright-only field contract rejection, got %+v", findings)
	}
}

func TestAuditBrowserActionSchemaFieldContractsRequiresBackendNeutralLLMSchema(t *testing.T) {
	findings := AuditBrowserActionSchemaFieldContracts([]BrowserActionSchemaFieldContract{
		{
			Field:           "selector_engine",
			Owner:           "tools/browser_action_schema",
			Kind:            BrowserActionSchemaFieldKindTarget,
			LLMActionSchema: true,
			BackendNeutral:  false,
		},
	})
	if !browserActionSchemaFieldFindingExists(findings, "", "selector_engine", "browser_action_schema_field_backend_neutral_required") {
		t.Fatalf("expected backend-neutral schema rejection, got %+v", findings)
	}
}

func TestAuditBrowserActionSchemaFieldContractsKeepSelectorAsNeutralDOMHint(t *testing.T) {
	var selector BrowserActionSchemaFieldContract
	for _, contract := range BuiltinBrowserActionSchemaFieldContracts() {
		if contract.Field == "selector" {
			selector = contract
			break
		}
	}
	if selector.Field == "" {
		t.Fatal("expected selector field contract")
	}
	if !selector.BackendNeutral || selector.PlaywrightLocatorDSL || selector.Kind != BrowserActionSchemaFieldKindTarget {
		t.Fatalf("selector should remain a backend-neutral DOM target hint, got %+v", selector)
	}
}

func browserActionSchemaFieldFindingExists(findings []BrowserActionSchemaFieldFinding, surface string, field string, rule string) bool {
	for _, finding := range findings {
		if finding.Surface == surface && finding.Field == field && finding.Rule == rule {
			return true
		}
	}
	return false
}
