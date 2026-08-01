package pack

import (
	"strings"
	"testing"

	agentxworkflow "github.com/wsnacj/agentx-go/runtime/workflow"
)

func TestValidateDefinitionAcceptsPackOwnedCaseLibrary(t *testing.T) {
	def := caseLibraryTestDefinition()
	if err := testCoordinator(t).ValidateDefinition(def); err != nil {
		t.Fatalf("validate definition with case library: %v", err)
	}

	item, ok := def.CaseLibraryCaseByID("browser_action_failure_payload.visible")
	if !ok {
		t.Fatalf("expected case library item lookup to succeed")
	}
	if item.CaseType != "browser.action_failure_payload" || item.Locale != "en-US" || item.ReviewStatus != CaseReviewStatusApproved {
		t.Fatalf("unexpected case library item metadata: %#v", item)
	}
	items := def.CaseLibraryCasesForType("browser.action_failure_payload")
	if len(items) != 1 || items[0].ID != item.ID {
		t.Fatalf("expected one case library item for type, got %#v", items)
	}
}

func TestValidateDefinitionRejectsMalformedCaseLibrary(t *testing.T) {
	tests := []struct {
		name string
		edit func(*Definition)
		want string
	}{
		{
			name: "unsupported_case_type",
			edit: func(def *Definition) {
				def.CaseLibrary[0].CaseType = "browser.unknown"
			},
			want: "case_type \"browser.unknown\" is not declared",
		},
		{
			name: "missing_input",
			edit: func(def *Definition) {
				def.CaseLibrary[0].Input = nil
			},
			want: "input is required",
		},
		{
			name: "missing_expected_output",
			edit: func(def *Definition) {
				def.CaseLibrary[0].ExpectedOutput = nil
			},
			want: "expected_output is required",
		},
		{
			name: "malformed_expected_output_path",
			edit: func(def *Definition) {
				def.CaseLibrary[0].ExpectedOutput = map[string]any{" review.passed": true}
			},
			want: "must not include surrounding whitespace",
		},
		{
			name: "unsupported_review_status",
			edit: func(def *Definition) {
				def.CaseLibrary[0].ReviewStatus = "approved "
			},
			want: "must not include surrounding whitespace",
		},
		{
			name: "duplicate_placeholder",
			edit: func(def *Definition) {
				def.CaseLibrary[0].InputPlaceholders = append(def.CaseLibrary[0].InputPlaceholders, CaseInputPlaceholder{Name: "payloads"})
			},
			want: "duplicate case library",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			def := caseLibraryTestDefinition()
			tt.edit(&def)
			err := testCoordinator(t).ValidateDefinition(def)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("expected %q validation error, got %v", tt.want, err)
			}
		})
	}
}

func TestCloneDefinitionDeepCopiesCaseLibrary(t *testing.T) {
	def := caseLibraryTestDefinition()
	lookup, ok := def.CaseLibraryCaseByID("browser_action_failure_payload.visible")
	if !ok {
		t.Fatalf("expected case library lookup to succeed")
	}
	lookup.Input["payloads"] = []any{"mutated"}
	if got := def.CaseLibrary[0].Input["payloads"].([]any); len(got) != 1 || got[0] != "{{payloads}}" {
		t.Fatalf("expected original input to remain unchanged after lookup mutation, got %#v", got)
	}

	cloned := cloneDefinition(def)

	cloned.CaseLibrary[0].Input["payloads"] = []any{"mutated"}
	cloned.CaseLibrary[0].InputPlaceholders[0].Example = []any{"mutated"}
	cloned.CaseLibrary[0].ExpectedOutput["passed"] = false
	cloned.CaseLibrary[0].Tags[0] = "mutated"

	if got := def.CaseLibrary[0].Input["payloads"].([]any); len(got) != 1 || got[0] != "{{payloads}}" {
		t.Fatalf("expected original input to remain unchanged, got %#v", got)
	}
	if got := def.CaseLibrary[0].InputPlaceholders[0].Example.([]any); len(got) != 0 {
		t.Fatalf("expected original placeholder example to remain unchanged, got %#v", got)
	}
	if got := def.CaseLibrary[0].ExpectedOutput["passed"]; got != true {
		t.Fatalf("expected original expected output to remain unchanged, got %#v", got)
	}
	if got := def.CaseLibrary[0].Tags[0]; got != "regression" {
		t.Fatalf("expected original tags to remain unchanged, got %q", got)
	}
}

func caseLibraryTestDefinition() Definition {
	return Definition{
		Manifest: Manifest{
			ID:                 "case-library-test-pack",
			Version:            "0.1.0",
			Domain:             "browser_operations",
			SupportedCaseTypes: []string{"browser.action_failure_payload"},
			DefaultWorkflow:    "browser_action_failure_payload_v1",
		},
		CaseSchemas: []CaseSchema{
			{CaseType: "browser.action_failure_payload"},
		},
		CaseLibrary: []CaseLibraryCase{
			{
				ID:          "browser_action_failure_payload.visible",
				CaseType:    "browser.action_failure_payload",
				Locale:      "en-US",
				Description: "Visible actionability failure payload regression.",
				Input: map[string]any{
					"payloads": []any{"{{payloads}}"},
				},
				InputPlaceholders: []CaseInputPlaceholder{
					{
						Name:        "payloads",
						Path:        "payloads",
						Description: "Captured browser_act action_failed payloads.",
						Required:    true,
						Example:     []any{},
					},
				},
				ExpectedOutput: map[string]any{
					"passed":                 true,
					"required_failed_checks": []any{"visible"},
				},
				ReviewStatus: CaseReviewStatusApproved,
				Tags:         []string{"regression", "browser"},
			},
		},
		Workflows: []agentxworkflow.Spec{
			{
				ID:        "browser_action_failure_payload_v1",
				Pack:      "case-library-test-pack",
				CaseTypes: []string{"browser.action_failure_payload"},
				EntryNode: "evaluate_payload",
				Nodes: []agentxworkflow.NodeSpec{
					{
						ID:     "evaluate_payload",
						Kind:   agentxworkflow.NodeTool,
						Config: map[string]any{"tool": "evaluate_payload"},
					},
				},
			},
		},
	}
}
