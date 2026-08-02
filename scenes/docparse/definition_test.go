package docparse

import (
	"testing"

	agentxworkflow "github.com/wsnacj/agentx-go/runtime/workflow"
	docparsehostkit "github.com/wsnacj/agentx-go/scenes/docparse/hostkit"
)

func TestDefinitionValidates(t *testing.T) {
	def := Definition()
	if def.Manifest.ID != PackID {
		t.Fatalf("pack id = %q, want %q", def.Manifest.ID, PackID)
	}
	if len(def.Manifest.SupportedCaseTypes) != len(supportedCaseTypes()) {
		t.Fatalf("supported case type count = %d", len(def.Manifest.SupportedCaseTypes))
	}
	if _, ok := def.SemanticToolByName(docparsehostkit.ToolDocparseExtractFields); !ok {
		t.Fatalf("definition missing semantic tool %q", docparsehostkit.ToolDocparseExtractFields)
	}
	if _, ok := def.SemanticToolByName(docparsehostkit.ToolDocparseProfileProbe); !ok {
		t.Fatalf("definition missing semantic tool %q", docparsehostkit.ToolDocparseProfileProbe)
	}
}

func TestDefinitionContainsAllStartCriteriaWorkflows(t *testing.T) {
	def := Definition()
	for _, item := range []struct {
		caseType string
		workflow string
	}{
		{CaseTypeProfileProbe, ProfileProbeWorkflow},
		{CaseTypeExtractFields, ExtractFieldsWorkflow},
		{CaseTypeExtractTable, ExtractTableWorkflow},
		{CaseTypeVerifyFields, VerifyFieldsWorkflow},
		{CaseTypeEvidenceTrace, EvidenceTraceWorkflow},
		{CaseTypeGuard, GuardWorkflow},
	} {
		if !def.Manifest.SupportsCaseType(item.caseType) {
			t.Fatalf("manifest does not support case type %q", item.caseType)
		}
		workflow, err := def.ResolveWorkflowForCaseType(item.caseType, "")
		if err != nil {
			t.Fatalf("ResolveWorkflowForCaseType(%q) returned error: %v", item.caseType, err)
		}
		if workflow.ID != item.workflow {
			t.Fatalf("case type %q workflow = %q, want %q", item.caseType, workflow.ID, item.workflow)
		}
	}
}

func TestDefinitionPassesEvaluatorInputsToWorkflow(t *testing.T) {
	def := Definition()
	workflow, err := def.ResolveWorkflowForCaseType(CaseTypeVerifyFields, "")
	if err != nil {
		t.Fatalf("ResolveWorkflowForCaseType returned error: %v", err)
	}
	if len(workflow.Nodes) != 1 {
		t.Fatalf("workflow node count = %d, want 1", len(workflow.Nodes))
	}
	for _, item := range []struct {
		from string
		to   string
	}{
		{from: "case.input.parse_result", to: "args.parse_result"},
		{from: "case.input.page_range", to: "args.page_range"},
		{from: "case.input.required_evidence", to: "args.required_evidence"},
		{from: "case.input.require_page_refs", to: "args.require_page_refs"},
		{from: "case.input.require_bounding_boxes", to: "args.require_bounding_boxes"},
		{from: "case.input.require_table_cells", to: "args.require_table_cells"},
		{from: "case.input.require_complete_table_rows", to: "args.require_complete_table_rows"},
		{from: "case.input.allow_review_required", to: "args.allow_review_required"},
	} {
		if !bindingExists(workflow.Nodes[0].Inputs, item.from, item.to) {
			t.Fatalf("workflow missing binding %s -> %s: %#v", item.from, item.to, workflow.Nodes[0].Inputs)
		}
	}
	for _, binding := range workflow.Nodes[0].Inputs {
		if binding.From == "case.input.user_message" {
			continue
		}
		if !binding.Optional {
			t.Fatalf("expected non-required docparse binding to be optional: %#v", binding)
		}
	}
	for _, binding := range workflow.Nodes[0].Outputs {
		switch binding.From {
		case "result.status", "result.evidence_complete", "result.passed", "result.summary":
			continue
		default:
			if !binding.Optional {
				t.Fatalf("expected diagnostic docparse output binding to be optional: %#v", binding)
			}
		}
	}
}

func TestDefinitionUsesStateBasedEvidencePayloads(t *testing.T) {
	def := Definition()
	if len(def.Manifest.ArtifactTypes) != 0 {
		t.Fatalf("docparse semantic tools return evidence payload/state, not registry artifacts: %#v", def.Manifest.ArtifactTypes)
	}
	for _, workflow := range def.Workflows {
		if len(workflow.ArtifactSchema) != 0 {
			t.Fatalf("workflow %q should not require registry artifacts for evidence payload output: %#v", workflow.ID, workflow.ArtifactSchema)
		}
	}
	for _, tool := range def.Tools {
		if len(tool.ArtifactTypes) != 0 {
			t.Fatalf("semantic tool %q should not declare registry artifact types for evidence payload output: %#v", tool.Name, tool.ArtifactTypes)
		}
	}
	for _, suite := range def.EvalSuites {
		if len(suite.RequiredArtifacts) != 0 {
			t.Fatalf("eval suite %q should gate docparse payloads through state, not registry artifacts: %#v", suite.Name, suite.RequiredArtifacts)
		}
		if !stringSliceContains(suite.RequiredState, "docparse.evidence_complete") {
			t.Fatalf("eval suite %q should require docparse.evidence_complete state: %#v", suite.Name, suite.RequiredState)
		}
	}
}

func bindingExists(bindings []agentxworkflow.BindingSpec, from string, to string) bool {
	for _, binding := range bindings {
		if binding.From == from && binding.To == to {
			return true
		}
	}
	return false
}

func stringSliceContains(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}
