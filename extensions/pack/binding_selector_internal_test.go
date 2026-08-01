package pack

import (
	"testing"

	agentxexecution "github.com/wsnacj/agentx-go/runtime/executionpolicy"
	agentxworkflow "github.com/wsnacj/agentx-go/runtime/workflow"
)

func TestResolveBindingRejectsWhitespacePaddedSelectors(t *testing.T) {
	def := Definition{
		Manifest: Manifest{
			ID:                 "browserops",
			Version:            "1.0.0",
			Domain:             "browser",
			SupportedCaseTypes: []string{"browser.form_submission"},
			DefaultWorkflow:    "browser_form_submit_v1",
			PolicyProfiles:     []string{"browser_fill_review"},
		},
		Workflows: []agentxworkflow.Spec{
			{
				ID:              "browser_form_submit_v1",
				Pack:            "browserops",
				EntryNode:       "open_target",
				DefaultContract: "browser_fill_review",
				CaseTypes:       []string{"browser.form_submission"},
				Nodes: []agentxworkflow.NodeSpec{
					{ID: "open_target", Kind: agentxworkflow.NodeTool, Config: map[string]any{"tool": "browser_act"}},
				},
			},
		},
		PolicyProfiles: []PolicyProfile{
			{Name: "browser_fill_review", Contract: agentxexecution.Contract{ID: "browser_fill_review_contract"}},
		},
		MemorySchemas: []MemorySchema{
			{Name: "browser_run_memory"},
		},
	}

	if _, err := testCoordinator(t).resolveBindingFromDefinition(def, " browser.form_submission", ""); err == nil {
		t.Fatalf("expected whitespace-padded case type selector to fail")
	}
	if _, err := testCoordinator(t).resolveBindingFromDefinition(def, "browser.form_submission", " browser_form_submit_v1"); err == nil {
		t.Fatalf("expected whitespace-padded workflow selector to fail")
	}
}

func TestBindingSelectorHelpersRejectWhitespacePaddedNames(t *testing.T) {
	binding := Binding{
		PackID: "browserops",
		Definition: Definition{
			PolicyProfiles: []PolicyProfile{
				{Name: "browser_fill_review", Contract: agentxexecution.Contract{ID: "browser_fill_review_contract"}},
			},
			MemorySchemas: []MemorySchema{
				{Name: "browser_run_memory"},
			},
		},
	}

	if _, ok := binding.ContractByRef(" browser_fill_review"); ok {
		t.Fatalf("expected whitespace-padded contract ref lookup to fail")
	}
	if _, ok := binding.ContractByRef("browser_fill_review"); !ok {
		t.Fatalf("expected canonical contract ref lookup to succeed")
	}

	if _, ok, err := binding.ResolveMemorySchema(" browser_run_memory"); err == nil || ok {
		t.Fatalf("expected whitespace-padded memory schema lookup to fail, ok=%v err=%v", ok, err)
	}
	if _, ok, err := binding.ResolveMemorySchema("browser_run_memory"); err != nil || !ok {
		t.Fatalf("expected canonical memory schema lookup to succeed, ok=%v err=%v", ok, err)
	}
}
