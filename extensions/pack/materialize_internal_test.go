package pack

import (
	"errors"
	"strings"
	"testing"

	agentxworkflow "github.com/wsnacj/agentx-go/runtime/workflow"
)

type packWorkflowValidatorStub struct {
	calls int
	err   error
}

func (s *packWorkflowValidatorStub) ValidateSpec(agentxworkflow.Spec) error {
	s.calls++
	return s.err
}

func TestValidateDefinitionUsesInjectedWorkflowValidator(t *testing.T) {
	sentinel := errors.New("injected definition validator")
	validator := &packWorkflowValidatorStub{err: sentinel}

	err := validateDefinitionWithWorkflowValidator(caseLibraryTestDefinition(), validator)

	if !errors.Is(err, sentinel) {
		t.Fatalf("validation error = %v, want injected sentinel", err)
	}
	if validator.calls != 1 {
		t.Fatalf("validator calls = %d, want 1", validator.calls)
	}
}

func TestMaterializeWorkflowSpecUsesInjectedWorkflowValidator(t *testing.T) {
	sentinel := errors.New("injected materialization validator")
	validator := &packWorkflowValidatorStub{err: sentinel}

	_, err := materializeWorkflowSpecWithWorkflowValidator(agentxworkflow.Spec{
		ID:        "materialize-injected-validator",
		EntryNode: "collect",
		Nodes: []agentxworkflow.NodeSpec{{
			ID:     "collect",
			Kind:   agentxworkflow.NodeTool,
			Config: map[string]any{"tool": "collect"},
		}},
	}, nil, validator, testToolArgumentLowerer())

	if !errors.Is(err, sentinel) {
		t.Fatalf("materialization error = %v, want injected sentinel", err)
	}
	if validator.calls != 1 {
		t.Fatalf("validator calls = %d, want 1", validator.calls)
	}
}

func TestMemoryRegistryGetRejectsWhitespacePaddedPackID(t *testing.T) {
	reg, err := NewMemoryRegistry(testCoordinator(t))
	if err != nil {
		t.Fatalf("new registry: %v", err)
	}
	def := Definition{
		Manifest: Manifest{
			ID:                 "browserops",
			Version:            "1.0.0",
			Domain:             "browser",
			SupportedCaseTypes: []string{"browser.form_submission"},
			DefaultWorkflow:    "browser_form_submit_v1",
		},
		Workflows: []agentxworkflow.Spec{
			{
				ID:        "browser_form_submit_v1",
				Pack:      "browserops",
				EntryNode: "open_target",
				CaseTypes: []string{"browser.form_submission"},
				Nodes: []agentxworkflow.NodeSpec{
					{ID: "open_target", Kind: agentxworkflow.NodeTool, Config: map[string]any{"tool": "browser_act"}},
				},
			},
		},
	}
	if err := reg.Register(def); err != nil {
		t.Fatalf("register definition: %v", err)
	}
	if _, ok := reg.Get("browserops"); !ok {
		t.Fatalf("expected canonical pack id lookup to succeed")
	}
	if _, ok := reg.Get(" browserops"); ok {
		t.Fatalf("expected whitespace-padded pack id lookup to fail")
	}
}

func TestMaterializeWorkflowRejectsWhitespacePaddedWorkflowID(t *testing.T) {
	def := Definition{
		Manifest: Manifest{
			ID:                 "browserops",
			Version:            "1.0.0",
			Domain:             "browser",
			SupportedCaseTypes: []string{"browser.form_submission"},
			DefaultWorkflow:    "browser_form_submit_v1",
		},
		Workflows: []agentxworkflow.Spec{
			{
				ID:        "browser_form_submit_v1",
				Pack:      "browserops",
				EntryNode: "open_target",
				CaseTypes: []string{"browser.form_submission"},
				Nodes: []agentxworkflow.NodeSpec{
					{ID: "open_target", Kind: agentxworkflow.NodeTool, Config: map[string]any{"tool": "browser_act"}},
				},
			},
		},
	}
	coordinator := testCoordinator(t)
	if _, err := coordinator.MaterializeWorkflow(def, " browser_form_submit_v1"); err == nil {
		t.Fatalf("expected whitespace-padded workflow id to fail materialization")
	}
	if _, err := coordinator.MaterializeWorkflow(def, "browser_form_submit_v1"); err != nil {
		t.Fatalf("expected canonical workflow id to materialize, got %v", err)
	}
}

func TestSemanticToolIndexStopsTrimmingStoredToolNames(t *testing.T) {
	def := Definition{
		Tools: []SemanticTool{
			{Name: " browser_capture_submission_evidence", RuntimeTool: "browser_act"},
		},
	}
	index := def.semanticToolIndex()
	if _, ok := index["browser_capture_submission_evidence"]; ok {
		t.Fatalf("expected trimmed semantic tool name not to match malformed stored key")
	}
	if _, ok := index[" browser_capture_submission_evidence"]; !ok {
		t.Fatalf("expected raw semantic tool name to be preserved in index")
	}
}

func TestMaterializeToolConfigPreservesRawRuntimeTool(t *testing.T) {
	config, err := materializeToolConfig(agentxworkflow.NodeSpec{
		ID:   "open_target",
		Kind: agentxworkflow.NodeTool,
		Config: map[string]any{
			"tool": "browser_open_target",
		},
	}, SemanticTool{
		Name:        "browser_open_target",
		RuntimeTool: " browser_act ",
	}, testToolArgumentLowerer())
	if err != nil {
		t.Fatalf("materialize tool config: %v", err)
	}
	if got := config["tool"]; got != " browser_act " {
		t.Fatalf("expected raw runtime tool to be preserved, got %#v", got)
	}
}

func TestCloneEvalSuitesPreservesRawMode(t *testing.T) {
	cloned := cloneEvalSuites([]EvalSuite{
		{
			Name: "browser_submit_success_suite",
			Mode: " shadow ",
		},
	})
	if len(cloned) != 1 || cloned[0].Mode != " shadow " {
		t.Fatalf("expected raw eval suite mode to be preserved, got %#v", cloned)
	}
}

func TestMaterializeWorkflowSpecPreservesRawNodeIDInSemanticToolError(t *testing.T) {
	_, err := materializeWorkflowSpecWithWorkflowValidator(agentxworkflow.Spec{
		ID:        "browser_form_submit_v1",
		EntryNode: " open_target ",
		Nodes: []agentxworkflow.NodeSpec{
			{
				ID:   " open_target ",
				Kind: agentxworkflow.NodeTool,
				Config: map[string]any{
					"tool":           "browser_open_target",
					"arguments_json": true,
				},
			},
		},
	}, map[string]SemanticTool{
		"browser_open_target": {
			Name:        "browser_open_target",
			RuntimeTool: "browser_act",
			RuntimeArgs: map[string]any{"selector": "#submit"},
		},
	}, validatorFunc(func(agentxworkflow.Spec) error { return nil }), testToolArgumentLowerer())
	if err == nil {
		t.Fatalf("expected semantic tool materialization error")
	}
	if !strings.Contains(err.Error(), `node " open_target "`) {
		t.Fatalf("expected raw node id in materialization error, got %v", err)
	}
}
