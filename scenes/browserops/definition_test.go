package browserops

import (
	"encoding/json"
	"strings"
	"testing"

	agentxpack "github.com/wsnacj/agentx-go/extensions/pack"
	agentxworkflow "github.com/wsnacj/agentx-go/runtime/workflow"
)

func TestDefinitionValidatesAndMaterializes(t *testing.T) {
	def := Definition()
	coordinator := newTestCoordinator(t)
	if err := coordinator.ValidateDefinition(def); err != nil {
		t.Fatalf("validate browserops pack: %v", err)
	}

	spec, err := MaterializedDefaultWorkflow(coordinator)
	if err != nil {
		t.Fatalf("materialize browserops workflow: %v", err)
	}
	if spec.ID != DefaultWorkflow || spec.Pack != PackID {
		t.Fatalf("unexpected materialized workflow identity: %#v", spec)
	}

	if tool := nodeTool(spec, "open_target"); tool != "browser_act" {
		t.Fatalf("expected open_target to materialize to browser_act, got %q", tool)
	}
	if tool := nodeTool(spec, "capture_evidence"); tool != "browser_screenshot" {
		t.Fatalf("expected capture_evidence to materialize to browser_screenshot, got %q", tool)
	}
	if target := nodeInputTarget(spec, "open_target", 0); target != "args.url" {
		t.Fatalf("expected open_target input target args.url, got %q", target)
	}
	if target := nodeInputTarget(spec, "final_gate", 0); target != "args.input" {
		t.Fatalf("expected final_gate primary input target args.input, got %q", target)
	}
	if !stringSliceContains(def.EvalSuites[0].RequiredState, "review.snapshot") ||
		!stringSliceContains(def.EvalSuites[0].RequiredState, "review.evidence_path") {
		t.Fatalf("expected eval suite to require snapshot and screenshot evidence state, got %#v", def.EvalSuites[0].RequiredState)
	}
	schema := def.Evaluators[0].OutputSchema
	props, _ := schema["properties"].(map[string]any)
	if _, ok := props["visual_evidence_ready"]; !ok {
		t.Fatalf("expected submit evaluator schema to expose visual_evidence_ready, got %#v", schema)
	}
	for _, key := range []string{"final_url_ready", "artifact_refs_ready", "evidence_bundle"} {
		if _, ok := props[key]; !ok {
			t.Fatalf("expected submit evaluator schema to expose %s, got %#v", key, schema)
		}
	}
	if kind := nodeConfigString(spec, "fill_fields", "kind"); kind != "fill" {
		t.Fatalf("expected fill_fields kind fill, got %q", kind)
	}
	if value := nodeConfigAny(spec, "capture_evidence", "full_page"); value != true {
		t.Fatalf("expected capture_evidence full_page=true, got %#v", value)
	}

	failureSpec, err := coordinator.MaterializeWorkflow(def, ActionFailurePayloadWorkflow)
	if err != nil {
		t.Fatalf("materialize browser action failure workflow: %v", err)
	}
	if failureSpec.ID != ActionFailurePayloadWorkflow || failureSpec.Pack != PackID {
		t.Fatalf("unexpected action failure workflow identity: %#v", failureSpec)
	}
	if evaluator := nodeConfigString(failureSpec, "failure_payload_gate", "evaluator"); evaluator != "browser_action_failure_payload_gate" {
		t.Fatalf("expected failure payload evaluator, got %q", evaluator)
	}
	if len(def.Manifest.Evaluators) != 6 || !stringSliceContains(def.Manifest.Evaluators, "browser_action_failure_payload_gate") {
		t.Fatalf("expected browser action failure evaluator in manifest, got %#v", def.Manifest.Evaluators)
	}
	if len(def.Manifest.EvalSuites) != 6 || !stringSliceContains(def.Manifest.EvalSuites, "browser_action_failure_payload_suite") {
		t.Fatalf("expected browser action failure eval suite in manifest, got %#v", def.Manifest.EvalSuites)
	}
	if !stringSliceContains(def.Manifest.ArtifactTypes, BrowserArtifactTypeActionTrace) {
		t.Fatalf("expected browser action trace artifact type in manifest, got %#v", def.Manifest.ArtifactTypes)
	}
	if item, ok := def.CaseLibraryCaseByID("browser_action_failure_payload.core_actionability_regression"); !ok {
		t.Fatalf("expected browser action failure case library item")
	} else if item.CaseType != "browser.action_failure_payload" || item.ReviewStatus != agentxpack.CaseReviewStatusApproved {
		t.Fatalf("unexpected browser action failure case library metadata: %#v", item)
	}
	if items := def.CaseLibraryCasesForType("browser.action_failure_payload"); len(items) != 1 {
		t.Fatalf("expected one browser action failure case library item, got %#v", items)
	}
	if tmpl, ok := def.PromptTemplateByName("browser_action_failure_payload_gate_instruction"); !ok {
		t.Fatalf("expected browser action failure prompt template")
	} else if len(tmpl.SourceAttributions) != 1 || tmpl.SourceAttributions[0].SourceType != agentxpack.SourceAttributionTypePack {
		t.Fatalf("expected pack source-attributed prompt template, got %#v", tmpl)
	}

	pageStateSpec, err := coordinator.MaterializeWorkflow(def, VerifyPageStateWorkflow)
	if err != nil {
		t.Fatalf("materialize browser page state workflow: %v", err)
	}
	if pageStateSpec.ID != VerifyPageStateWorkflow || pageStateSpec.DefaultContract != "browser_page_state_review" {
		t.Fatalf("unexpected page state workflow metadata: %#v", pageStateSpec)
	}
	if tool := nodeTool(pageStateSpec, "capture_evidence"); tool != "browser_screenshot" {
		t.Fatalf("expected page_state capture_evidence to materialize to browser_screenshot, got %q", tool)
	}
	if evaluator := nodeConfigString(pageStateSpec, "page_state_gate", "evaluator"); evaluator != "browser_page_state_gate" {
		t.Fatalf("expected page state evaluator, got %q", evaluator)
	}
	if !stringSliceContains(def.Manifest.SupportedCaseTypes, "browser.verify_page_state") ||
		!stringSliceContains(def.Manifest.Evaluators, "browser_page_state_gate") ||
		!stringSliceContains(def.Manifest.EvalSuites, "browser_page_state_success_suite") {
		t.Fatalf("expected page state case/evaluator/suite in manifest: %#v", def.Manifest)
	}
	if tmpl, ok := def.PromptTemplateByName("browser_page_state_gate_instruction"); !ok {
		t.Fatalf("expected browser page state prompt template")
	} else if len(tmpl.SourceAttributions) != 1 || tmpl.SourceAttributions[0].SourceType != agentxpack.SourceAttributionTypePack {
		t.Fatalf("expected pack source-attributed page state prompt template, got %#v", tmpl)
	}

	extractSpec, err := coordinator.MaterializeWorkflow(def, ExtractStructuredDataWorkflow)
	if err != nil {
		t.Fatalf("materialize browser structured data workflow: %v", err)
	}
	if extractSpec.ID != ExtractStructuredDataWorkflow || extractSpec.DefaultContract != "browser_extract_readonly_review" {
		t.Fatalf("unexpected structured data workflow metadata: %#v", extractSpec)
	}
	if tool := nodeTool(extractSpec, "capture_evidence"); tool != "browser_screenshot" {
		t.Fatalf("expected structured_data capture_evidence to materialize to browser_screenshot, got %q", tool)
	}
	if evaluator := nodeConfigString(extractSpec, "structured_data_gate", "evaluator"); evaluator != "browser_structured_data_gate" {
		t.Fatalf("expected structured data evaluator, got %q", evaluator)
	}
	if !stringSliceContains(def.Manifest.SupportedCaseTypes, "browser.extract_structured_data") ||
		!stringSliceContains(def.Manifest.Evaluators, "browser_structured_data_gate") ||
		!stringSliceContains(def.Manifest.EvalSuites, "browser_structured_data_success_suite") {
		t.Fatalf("expected structured data case/evaluator/suite in manifest: %#v", def.Manifest)
	}
	if tmpl, ok := def.PromptTemplateByName("browser_structured_data_gate_instruction"); !ok {
		t.Fatalf("expected browser structured data prompt template")
	} else if len(tmpl.SourceAttributions) != 1 || tmpl.SourceAttributions[0].SourceType != agentxpack.SourceAttributionTypePack {
		t.Fatalf("expected pack source-attributed structured data prompt template, got %#v", tmpl)
	}

	siteSearchSpec, err := coordinator.MaterializeWorkflow(def, SiteSearchWorkflow)
	if err != nil {
		t.Fatalf("materialize browser site search workflow: %v", err)
	}
	if siteSearchSpec.ID != SiteSearchWorkflow || siteSearchSpec.DefaultContract != "browser_site_search_review" {
		t.Fatalf("unexpected site search workflow metadata: %#v", siteSearchSpec)
	}
	if tool := nodeTool(siteSearchSpec, "submit_search"); tool != "browser_act" {
		t.Fatalf("expected site_search submit_search to materialize to browser_act, got %q", tool)
	}
	if kind := nodeConfigString(siteSearchSpec, "submit_search", "kind"); kind != "fill" {
		t.Fatalf("expected site_search submit_search kind fill, got %q", kind)
	}
	if evaluator := nodeConfigString(siteSearchSpec, "site_search_gate", "evaluator"); evaluator != "browser_site_search_gate" {
		t.Fatalf("expected site search evaluator, got %q", evaluator)
	}
	if !stringSliceContains(def.Manifest.SupportedCaseTypes, "browser.site_search") ||
		!stringSliceContains(def.Manifest.Evaluators, "browser_site_search_gate") ||
		!stringSliceContains(def.Manifest.EvalSuites, "browser_site_search_success_suite") {
		t.Fatalf("expected site search case/evaluator/suite in manifest: %#v", def.Manifest)
	}
	if tmpl, ok := def.PromptTemplateByName("browser_site_search_gate_instruction"); !ok {
		t.Fatalf("expected browser site search prompt template")
	} else if len(tmpl.SourceAttributions) != 1 || tmpl.SourceAttributions[0].SourceType != agentxpack.SourceAttributionTypePack {
		t.Fatalf("expected pack source-attributed site search prompt template, got %#v", tmpl)
	}

	downloadSpec, err := coordinator.MaterializeWorkflow(def, DownloadFileWorkflow)
	if err != nil {
		t.Fatalf("materialize browser download file workflow: %v", err)
	}
	if downloadSpec.ID != DownloadFileWorkflow || downloadSpec.DefaultContract != "browser_download_review" {
		t.Fatalf("unexpected download file workflow metadata: %#v", downloadSpec)
	}
	if tool := nodeTool(downloadSpec, "download_file"); tool != "browser_act" {
		t.Fatalf("expected download_file to materialize to browser_act, got %q", tool)
	}
	if kind := nodeConfigString(downloadSpec, "download_file", "kind"); kind != "download" {
		t.Fatalf("expected download_file kind download, got %q", kind)
	}
	if evaluator := nodeConfigString(downloadSpec, "download_gate", "evaluator"); evaluator != "browser_download_file_gate" {
		t.Fatalf("expected download file evaluator, got %q", evaluator)
	}
	if !stringSliceContains(def.Manifest.SupportedCaseTypes, "browser.download_file") ||
		!stringSliceContains(def.Manifest.Evaluators, "browser_download_file_gate") ||
		!stringSliceContains(def.Manifest.EvalSuites, "browser_download_file_success_suite") {
		t.Fatalf("expected download file case/evaluator/suite in manifest: %#v", def.Manifest)
	}
	if tmpl, ok := def.PromptTemplateByName("browser_download_file_gate_instruction"); !ok {
		t.Fatalf("expected browser download file prompt template")
	} else if len(tmpl.SourceAttributions) != 1 || tmpl.SourceAttributions[0].SourceType != agentxpack.SourceAttributionTypePack {
		t.Fatalf("expected pack source-attributed download file prompt template, got %#v", tmpl)
	}
}

func TestDefinitionRecordUpdateSchemaMatchesWorkflowInputs(t *testing.T) {
	def := Definition()
	if len(def.Manifest.OptionalSkills) != 0 {
		t.Fatalf("expected browserops canonical pack not to advertise optional skills without implementation, got %#v", def.Manifest.OptionalSkills)
	}
	coordinator := newTestCoordinator(t)
	reg, err := agentxpack.NewMemoryRegistry(coordinator)
	if err != nil {
		t.Fatalf("new browserops registry: %v", err)
	}
	if err := reg.Register(def); err != nil {
		t.Fatalf("register browserops pack: %v", err)
	}
	binding, ok, err := coordinator.ResolveBinding(reg, PackID, "browser.record_update", "")
	if err != nil || !ok {
		t.Fatalf("resolve record_update binding: ok=%v err=%v", ok, err)
	}
	err = binding.ValidateCaseInput(map[string]any{})
	if err == nil || !strings.Contains(err.Error(), "case_input.target_url is required") {
		t.Fatalf("expected record_update to require canonical workflow inputs, got %v", err)
	}
}

func TestDefinitionVerifyPageStateSchemaMatchesWorkflowInputs(t *testing.T) {
	def := Definition()
	coordinator := newTestCoordinator(t)
	reg, err := agentxpack.NewMemoryRegistry(coordinator)
	if err != nil {
		t.Fatalf("new browserops registry: %v", err)
	}
	if err := reg.Register(def); err != nil {
		t.Fatalf("register browserops pack: %v", err)
	}
	binding, ok, err := coordinator.ResolveBinding(reg, PackID, "browser.verify_page_state", "")
	if err != nil || !ok {
		t.Fatalf("resolve verify_page_state binding: ok=%v err=%v", ok, err)
	}
	if binding.WorkflowID != VerifyPageStateWorkflow {
		t.Fatalf("expected verify_page_state workflow %q, got %q", VerifyPageStateWorkflow, binding.WorkflowID)
	}
	err = binding.ValidateCaseInput(map[string]any{})
	if err == nil || !strings.Contains(err.Error(), "case_input.target_url is required") {
		t.Fatalf("expected verify_page_state to require target_url, got %v", err)
	}
	err = binding.ValidateCaseInput(map[string]any{
		"target_url": "https://example.com/status",
		"expectations": map[string]any{
			"required_text":      []any{"Healthy"},
			"forbidden_text":     []any{"Access denied"},
			"url_contains":       "status",
			"require_screenshot": true,
			"min_snapshot_chars": 20,
		},
	})
	if err != nil {
		t.Fatalf("expected verify_page_state case input to validate: %v", err)
	}
}

func TestDefinitionExtractStructuredDataSchemaMatchesWorkflowInputs(t *testing.T) {
	def := Definition()
	coordinator := newTestCoordinator(t)
	reg, err := agentxpack.NewMemoryRegistry(coordinator)
	if err != nil {
		t.Fatalf("new browserops registry: %v", err)
	}
	if err := reg.Register(def); err != nil {
		t.Fatalf("register browserops pack: %v", err)
	}
	binding, ok, err := coordinator.ResolveBinding(reg, PackID, "browser.extract_structured_data", "")
	if err != nil || !ok {
		t.Fatalf("resolve extract_structured_data binding: ok=%v err=%v", ok, err)
	}
	if binding.WorkflowID != ExtractStructuredDataWorkflow {
		t.Fatalf("expected extract_structured_data workflow %q, got %q", ExtractStructuredDataWorkflow, binding.WorkflowID)
	}
	err = binding.ValidateCaseInput(map[string]any{})
	if err == nil || !strings.Contains(err.Error(), "case_input.target_url is required") {
		t.Fatalf("expected extract_structured_data to require target_url, got %v", err)
	}
	err = binding.ValidateCaseInput(map[string]any{
		"target_url": "https://example.com/accounts/acme",
		"extraction": map[string]any{
			"fields": []any{
				map[string]any{"name": "company_name", "required": true, "expected_text": "Acme Robotics"},
				map[string]any{"name": "status", "required": true, "expected_text": "Active"},
			},
			"url_contains":       "accounts/acme",
			"require_screenshot": true,
			"min_snapshot_chars": 20,
		},
	})
	if err != nil {
		t.Fatalf("expected extract_structured_data case input to validate: %v", err)
	}
}

func TestDefinitionSiteSearchSchemaMatchesWorkflowInputs(t *testing.T) {
	def := Definition()
	coordinator := newTestCoordinator(t)
	reg, err := agentxpack.NewMemoryRegistry(coordinator)
	if err != nil {
		t.Fatalf("new browserops registry: %v", err)
	}
	if err := reg.Register(def); err != nil {
		t.Fatalf("register browserops pack: %v", err)
	}
	binding, ok, err := coordinator.ResolveBinding(reg, PackID, "browser.site_search", "")
	if err != nil || !ok {
		t.Fatalf("resolve site_search binding: ok=%v err=%v", ok, err)
	}
	if binding.WorkflowID != SiteSearchWorkflow {
		t.Fatalf("expected site_search workflow %q, got %q", SiteSearchWorkflow, binding.WorkflowID)
	}
	err = binding.ValidateCaseInput(map[string]any{})
	if err == nil || !strings.Contains(err.Error(), "case_input.target_url is required") {
		t.Fatalf("expected site_search to require target_url, got %v", err)
	}
	err = binding.ValidateCaseInput(map[string]any{
		"target_url": "https://example.com/search",
		"search": map[string]any{
			"query": "Acme Robotics",
			"fields": []any{
				map[string]any{"selector": "input[name=q]", "value": "Acme Robotics", "type": "text"},
			},
			"submit": true,
			"expected_results": []any{
				map[string]any{"title": "Acme Robotics Account", "url_contains": "/accounts/acme"},
			},
			"url_contains": "q=Acme",
		},
	})
	if err != nil {
		t.Fatalf("expected site_search case input to validate: %v", err)
	}
}

func TestDefinitionDownloadFileSchemaMatchesWorkflowInputs(t *testing.T) {
	def := Definition()
	coordinator := newTestCoordinator(t)
	reg, err := agentxpack.NewMemoryRegistry(coordinator)
	if err != nil {
		t.Fatalf("new browserops registry: %v", err)
	}
	if err := reg.Register(def); err != nil {
		t.Fatalf("register browserops pack: %v", err)
	}
	binding, ok, err := coordinator.ResolveBinding(reg, PackID, "browser.download_file", "")
	if err != nil || !ok {
		t.Fatalf("resolve download_file binding: ok=%v err=%v", ok, err)
	}
	if binding.WorkflowID != DownloadFileWorkflow {
		t.Fatalf("expected download_file workflow %q, got %q", DownloadFileWorkflow, binding.WorkflowID)
	}
	err = binding.ValidateCaseInput(map[string]any{})
	if err == nil || !strings.Contains(err.Error(), "case_input.target_url is required") {
		t.Fatalf("expected download_file to require target_url, got %v", err)
	}
	err = binding.ValidateCaseInput(map[string]any{
		"target_url": "https://example.com/downloads",
		"download": map[string]any{
			"url":                   "https://example.com/downloads/report.csv",
			"mode":                  "download",
			"expected_filename":     "report.csv",
			"expected_content_type": "text/csv",
			"min_bytes":             64,
			"require_screenshot":    true,
		},
	})
	if err != nil {
		t.Fatalf("expected download_file case input to validate: %v", err)
	}
}

type testValidator struct{}

func (testValidator) ValidateSpec(agentxworkflow.Spec) error { return nil }

type testLowerer struct{}

func (testLowerer) LowerToolArguments(node agentxworkflow.NodeSpec) (string, error) {
	arguments, _ := node.Config["args"].(map[string]any)
	encoded, err := json.Marshal(arguments)
	return string(encoded), err
}

func newTestCoordinator(t *testing.T) *agentxpack.Coordinator {
	t.Helper()
	coordinator, err := agentxpack.NewCoordinator(testValidator{}, testLowerer{})
	if err != nil {
		t.Fatalf("new browserops coordinator: %v", err)
	}
	return coordinator
}

func nodeTool(spec agentxworkflow.Spec, nodeID string) string {
	raw, _ := nodeConfigAny(spec, nodeID, "tool").(string)
	return raw
}

func nodeInputTarget(spec agentxworkflow.Spec, nodeID string, idx int) string {
	for _, node := range spec.Nodes {
		if node.ID != nodeID {
			continue
		}
		if idx < 0 || idx >= len(node.Inputs) {
			return ""
		}
		return node.Inputs[idx].To
	}
	return ""
}

func nodeConfigString(spec agentxworkflow.Spec, nodeID string, key string) string {
	raw, _ := nodeConfigAny(spec, nodeID, key).(string)
	return raw
}

func nodeConfigAny(spec agentxworkflow.Spec, nodeID string, key string) any {
	for _, node := range spec.Nodes {
		if node.ID != nodeID {
			continue
		}
		return node.Config[key]
	}
	return nil
}

func stringSliceContains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
