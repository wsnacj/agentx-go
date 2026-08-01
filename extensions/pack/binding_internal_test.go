package pack

import (
	"strings"
	"testing"

	agentxworkflow "github.com/wsnacj/agentx-go/runtime/workflow"
)

func TestReadBindingConfigStringPreservesRawValue(t *testing.T) {
	got := readBindingConfigString(map[string]any{
		"evaluator": " browser_submit_evidence_gate ",
	}, "evaluator")
	if got != " browser_submit_evidence_gate " {
		t.Fatalf("expected raw binding config string, got %q", got)
	}
}

func TestBuildPackEvaluatorInstructionPreservesRawDescription(t *testing.T) {
	got := buildPackEvaluatorInstruction(Evaluator{
		Description: " keep evaluator prompt raw ",
	})
	if got != " keep evaluator prompt raw " {
		t.Fatalf("expected raw evaluator description, got %q", got)
	}
}

func TestValidateWorkflowEvalSuiteCoverageRejectsWhitespacePaddedRequiredStatePath(t *testing.T) {
	spec := agentxworkflow.Spec{
		ID: "browser_form_submit_v1",
		StateSchema: []agentxworkflow.StateSlotSpec{
			{Name: "review.summary"},
		},
	}
	err := validateWorkflowEvalSuiteCoverageWithProducedArtifacts(spec, []EvalSuite{
		{
			Name:          "browser_submit_success_suite",
			RequiredState: []string{" review.summary"},
		},
	}, nil)
	if err == nil {
		t.Fatalf("expected whitespace-padded required_state path to fail")
	}
	if !strings.Contains(err.Error(), `undeclared state path " review.summary"`) {
		t.Fatalf("unexpected eval suite coverage error: %v", err)
	}
}

func TestValidateWorkflowEvalSuiteCoverageRejectsWhitespacePaddedSummaryPath(t *testing.T) {
	spec := agentxworkflow.Spec{
		ID: "browser_form_submit_v1",
		StateSchema: []agentxworkflow.StateSlotSpec{
			{Name: "review.summary"},
		},
	}
	err := validateWorkflowEvalSuiteCoverageWithProducedArtifacts(spec, []EvalSuite{
		{
			Name:        "browser_submit_success_suite",
			SummaryPath: "review.summary ",
		},
	}, nil)
	if err == nil {
		t.Fatalf("expected whitespace-padded summary_path to fail")
	}
	if !strings.Contains(err.Error(), `undeclared state path "review.summary "`) {
		t.Fatalf("unexpected eval suite coverage error: %v", err)
	}
}

func TestValidateWorkflowEvalSuiteCoveragePreservesRawArtifactTypeAndSuiteName(t *testing.T) {
	spec := agentxworkflow.Spec{
		ID: " browser_form_submit_v1 ",
	}
	err := validateWorkflowEvalSuiteCoverageWithProducedArtifacts(spec, []EvalSuite{
		{
			Name:              " browser_submit_success_suite ",
			RequiredArtifacts: []string{" browser.page.screenshot "},
		},
	}, map[string]bool{
		"browser.page.screenshot": true,
	})
	if err == nil {
		t.Fatalf("expected whitespace-padded artifact type to remain unmatched")
	}
	if !strings.Contains(err.Error(), `workflow " browser_form_submit_v1 "`) ||
		!strings.Contains(err.Error(), `eval suite " browser_submit_success_suite "`) ||
		!strings.Contains(err.Error(), `unproduced artifact type " browser.page.screenshot "`) {
		t.Fatalf("expected raw workflow/suite/artifact fields in error, got %v", err)
	}
}

func TestValidateWorkflowArtifactCoveragePreservesRawArtifactTypeAndWorkflowID(t *testing.T) {
	spec := agentxworkflow.Spec{
		ID: " browser_form_submit_v1 ",
		ArtifactSchema: []agentxworkflow.ArtifactTypeRef{
			{Type: " browser.page.screenshot "},
		},
	}
	err := validateWorkflowArtifactCoverage(spec)
	if err == nil {
		t.Fatalf("expected whitespace-padded artifact type to remain missing")
	}
	if !strings.Contains(err.Error(), `workflow " browser_form_submit_v1 "`) ||
		!strings.Contains(err.Error(), `artifact types " browser.page.screenshot "`) {
		t.Fatalf("expected raw workflow/artifact fields in error, got %v", err)
	}
}
