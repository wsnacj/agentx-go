package pack

import (
	"testing"

	agentxworkflow "github.com/wsnacj/agentx-go/runtime/workflow"
)

func TestPackLookupHelpersRejectWhitespacePaddedInputs(t *testing.T) {
	def := Definition{
		Manifest: Manifest{
			SupportedCaseTypes: []string{"browser.form_submission"},
		},
		CaseSchemas: []CaseSchema{
			{CaseType: "browser.form_submission"},
		},
		Workflows: []agentxworkflow.Spec{
			{ID: "browser_form_submit_v1", CaseTypes: []string{"browser.form_submission"}},
			{ID: "browser_form_shadow_v1", CaseTypes: []string{"browser.shadow_submission"}},
		},
		Tools: []SemanticTool{
			{Name: "browser_capture_submission_evidence"},
		},
		Evaluators: []Evaluator{
			{Name: "browser_submit_evidence_gate"},
		},
		EvalSuites: []EvalSuite{
			{Name: "browser_submit_success_suite", WorkflowIDs: []string{"browser_form_submit_v1"}},
			{Name: "browser_shadow_suite", WorkflowIDs: []string{"browser_form_shadow_v1"}},
		},
		PolicyProfiles: []PolicyProfile{
			{Name: "browser_fill_review"},
		},
		MemorySchemas: []MemorySchema{
			{Name: "browser_run_memory"},
		},
	}

	if def.Manifest.SupportsCaseType(" browser.form_submission") {
		t.Fatalf("expected whitespace-padded case type lookup to fail")
	}
	if _, ok := def.WorkflowByID(" browser_form_submit_v1"); ok {
		t.Fatalf("expected whitespace-padded workflow lookup to fail")
	}
	if matches := def.WorkflowsForCaseType(" browser.form_submission"); len(matches) != 0 {
		t.Fatalf("expected whitespace-padded case type workflow lookup to fail, got %#v", matches)
	}
	if _, ok := def.SemanticToolByName(" browser_capture_submission_evidence"); ok {
		t.Fatalf("expected whitespace-padded semantic tool lookup to fail")
	}
	if _, ok := def.CaseSchemaByType(" browser.form_submission"); ok {
		t.Fatalf("expected whitespace-padded case schema lookup to fail")
	}
	if _, ok := def.PolicyProfileByName(" browser_fill_review"); ok {
		t.Fatalf("expected whitespace-padded policy profile lookup to fail")
	}
	if _, ok := def.EvaluatorByName(" browser_submit_evidence_gate"); ok {
		t.Fatalf("expected whitespace-padded evaluator lookup to fail")
	}
	if _, ok := def.EvalSuiteByName(" browser_submit_success_suite"); ok {
		t.Fatalf("expected whitespace-padded eval suite lookup to fail")
	}
	if _, ok := def.MemorySchemaByName(" browser_run_memory"); ok {
		t.Fatalf("expected whitespace-padded memory schema lookup to fail")
	}
	if suites := def.EvalSuitesForWorkflow(" browser_form_submit_v1"); len(suites) != 0 {
		t.Fatalf("expected whitespace-padded eval suite workflow lookup to fail, got %#v", suites)
	}
}

func TestPackLookupHelpersDoNotTrimStoredIdentifiers(t *testing.T) {
	def := Definition{
		Manifest: Manifest{
			SupportedCaseTypes: []string{" browser.form_submission"},
		},
		CaseSchemas: []CaseSchema{
			{CaseType: " browser.form_submission"},
		},
		Workflows: []agentxworkflow.Spec{
			{ID: " browser_form_submit_v1", CaseTypes: []string{" browser.form_submission"}},
		},
		Tools: []SemanticTool{
			{Name: " browser_capture_submission_evidence"},
		},
		Evaluators: []Evaluator{
			{Name: " browser_submit_evidence_gate"},
		},
		EvalSuites: []EvalSuite{
			{Name: " browser_submit_success_suite", WorkflowIDs: []string{" browser_form_submit_v1"}},
			{Name: " browser_shadow_suite", WorkflowIDs: []string{" browser_form_shadow_v1"}},
		},
		PolicyProfiles: []PolicyProfile{
			{Name: " browser_fill_review"},
		},
		MemorySchemas: []MemorySchema{
			{Name: " browser_run_memory"},
		},
	}

	if def.Manifest.SupportsCaseType("browser.form_submission") {
		t.Fatalf("expected trimmed case type not to match malformed stored value")
	}
	if _, ok := def.WorkflowByID("browser_form_submit_v1"); ok {
		t.Fatalf("expected trimmed workflow id not to match malformed stored value")
	}
	if _, ok := def.SemanticToolByName("browser_capture_submission_evidence"); ok {
		t.Fatalf("expected trimmed semantic tool name not to match malformed stored value")
	}
	if _, ok := def.CaseSchemaByType("browser.form_submission"); ok {
		t.Fatalf("expected trimmed case schema type not to match malformed stored value")
	}
	if _, ok := def.PolicyProfileByName("browser_fill_review"); ok {
		t.Fatalf("expected trimmed policy profile name not to match malformed stored value")
	}
	if _, ok := def.EvaluatorByName("browser_submit_evidence_gate"); ok {
		t.Fatalf("expected trimmed evaluator name not to match malformed stored value")
	}
	if _, ok := def.EvalSuiteByName("browser_submit_success_suite"); ok {
		t.Fatalf("expected trimmed eval suite name not to match malformed stored value")
	}
	if _, ok := def.MemorySchemaByName("browser_run_memory"); ok {
		t.Fatalf("expected trimmed memory schema name not to match malformed stored value")
	}
	if suites := def.EvalSuitesForWorkflow("browser_form_submit_v1"); len(suites) != 0 {
		t.Fatalf("expected trimmed workflow id not to match malformed eval suite binding, got %#v", suites)
	}
}

func TestEvalSuitesForWorkflowStopsNormalizingStoredMode(t *testing.T) {
	def := Definition{
		EvalSuites: []EvalSuite{
			{
				Name:        "a_suite",
				Mode:        " shadow ",
				WorkflowIDs: []string{"browser_form_submit_v1"},
			},
			{
				Name:        "b_suite",
				Mode:        EvalSuiteModeGate,
				WorkflowIDs: []string{"browser_form_submit_v1"},
			},
		},
	}

	suites := def.EvalSuitesForWorkflow("browser_form_submit_v1")
	if len(suites) != 2 {
		t.Fatalf("expected two eval suites, got %#v", suites)
	}
	if suites[0].Name != "a_suite" || suites[0].Mode != " shadow " {
		t.Fatalf("expected malformed raw mode to remain raw and affect exact ordering, got %#v", suites)
	}
}

func TestWorkflowSelectionErrorPreservesRawIdentifiers(t *testing.T) {
	err := workflowSelectionError(" browserops ", " browser.form_submission ", " browser_form_submit_v1 ", "workflow_not_found")
	if err == nil {
		t.Fatalf("expected workflow selection error")
	}
	text := err.Error()
	if text != `pack: definition " browserops " case type " browser.form_submission " workflow " browser_form_submit_v1 " workflow not found` {
		t.Fatalf("expected raw workflow selection error fields, got %q", text)
	}
}

func TestNormalizeEvalSuiteModeRequiresExactCanonicalValue(t *testing.T) {
	if got := NormalizeEvalSuiteMode(""); got != EvalSuiteModeGate {
		t.Fatalf("expected empty mode to keep default gate semantics, got %q", got)
	}
	if got := NormalizeEvalSuiteMode(EvalSuiteModeGate); got != EvalSuiteModeGate {
		t.Fatalf("expected canonical gate mode, got %q", got)
	}
	if got := NormalizeEvalSuiteMode(EvalSuiteModeShadow); got != EvalSuiteModeShadow {
		t.Fatalf("expected canonical shadow mode, got %q", got)
	}
	if got := NormalizeEvalSuiteMode(" shadow "); got != "" {
		t.Fatalf("expected whitespace-padded mode to stay invalid, got %q", got)
	}
	if got := NormalizeEvalSuiteMode("SHADOW"); got != "" {
		t.Fatalf("expected uppercase mode to stay invalid, got %q", got)
	}
}
