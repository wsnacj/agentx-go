package metrics

import (
	"encoding/json"
	"strings"
	"testing"

	agentxpack "github.com/wsnacj/agentx-go/extensions/pack"
	agentxworkflow "github.com/wsnacj/agentx-go/runtime/workflow"
)

type testValidator struct{}

func (testValidator) ValidateSpec(agentxworkflow.Spec) error { return nil }

type testLowerer struct{}

func (testLowerer) LowerToolArguments(node agentxworkflow.NodeSpec) (string, error) {
	arguments, _ := node.Config["args"].(map[string]any)
	encoded, err := json.Marshal(arguments)
	return string(encoded), err
}

func testCoordinator(t *testing.T) *agentxpack.Coordinator {
	t.Helper()
	coordinator, err := agentxpack.NewCoordinator(testValidator{}, testLowerer{})
	if err != nil {
		t.Fatalf("new pack coordinator: %v", err)
	}
	return coordinator
}

func testRegistry(t *testing.T) *agentxpack.MemoryRegistry {
	t.Helper()
	registry, err := agentxpack.NewMemoryRegistry(testCoordinator(t))
	if err != nil {
		t.Fatalf("new pack registry: %v", err)
	}
	return registry
}

func TestDefinitionValidatesAndMaterializes(t *testing.T) {
	def := Definition()
	if err := testCoordinator(t).ValidateDefinition(def); err != nil {
		t.Fatalf("validate financial report metrics pack: %v", err)
	}

	spec, err := MaterializedDefaultWorkflow(testCoordinator(t))
	if err != nil {
		t.Fatalf("materialize financial report metrics workflow: %v", err)
	}
	if spec.ID != DefaultWorkflow || spec.Pack != PackID {
		t.Fatalf("unexpected materialized workflow identity: %#v", spec)
	}
	if !containsString(spec.CaseTypes, CaseTypeLatest) || !containsString(spec.CaseTypes, CaseTypeTrend) {
		t.Fatalf("expected workflow to support latest and trend case types, got %#v", spec.CaseTypes)
	}

	if tool := nodeTool(spec, "generate_candidates"); tool != "report_metrics_candidates" {
		t.Fatalf("expected generate_candidates to materialize to report_metrics_candidates, got %q", tool)
	}
	if tool := nodeTool(spec, "open_candidate"); tool != "open_page" {
		t.Fatalf("expected open_candidate to materialize to open_page, got %q", tool)
	}
	if hasNodeOutputTarget(spec, "open_candidate", "state.page.page_id") {
		t.Fatalf("open_candidate should not require page_id; adapters must be able to continue from final_url when open_page cannot cache text")
	}
	if tool := nodeTool(spec, "extract_metrics"); tool != "report_metrics_extract" {
		t.Fatalf("expected extract_metrics to materialize to report_metrics_extract, got %q", tool)
	}
	if tool := nodeTool(spec, "guard_metrics"); tool != "report_metrics_guard" {
		t.Fatalf("expected guard_metrics to materialize to report_metrics_guard, got %q", tool)
	}
	if !hasStateSlot(spec, "metrics.field_sources_accepted") {
		t.Fatalf("expected workflow state to expose metrics.field_sources_accepted for field provenance gates")
	}
	if !hasNodeOutputTarget(spec, "guard_metrics", "state.metrics.field_sources_accepted") {
		t.Fatalf("expected guard_metrics to project evaluation.field_sources_accepted into workflow state")
	}
	if target := nodeInputTarget(spec, "guard_metrics", 0); target != "args.user_message" {
		t.Fatalf("expected guard_metrics first input target args.user_message, got %q", target)
	}
	for _, nodeID := range []string{"generate_candidates", "extract_metrics", "guard_metrics"} {
		if !hasNodeInputBinding(spec, nodeID, "case.input.requested_metrics", "args.requested_metrics") {
			t.Fatalf("%s should receive requested_metrics from the pack task frame", nodeID)
		}
		if !hasNodeInputBinding(spec, nodeID, "case.input.requested_outputs", "args.requested_outputs") {
			t.Fatalf("%s should receive requested_outputs from the pack task frame", nodeID)
		}
		if !hasNodeInputBinding(spec, nodeID, "case.input.assessment.kind", "args.assessment_kind") {
			t.Fatalf("%s should receive assessment.kind from the pack task frame", nodeID)
		}
		if !hasNodeInputBinding(spec, nodeID, "case.input.period_policy", "args.period_scope") {
			t.Fatalf("%s should receive period_policy as period_scope from the pack task frame", nodeID)
		}
	}
	for _, nodeID := range []string{"extract_metrics", "guard_metrics"} {
		if !hasNodeInputBinding(spec, nodeID, "case.input.entity.name", "args.entity_name") {
			t.Fatalf("%s should receive entity.name from the pack task frame", nodeID)
		}
	}
	if suite, ok := def.EvalSuiteForWorkflow(DefaultWorkflow); !ok || suite.PassPath != "metrics.passed" {
		t.Fatalf("expected default eval suite to gate on metrics.passed, got ok=%v suite=%#v", ok, suite)
	}
}

func TestLatestMetricsCaseSchemaRequiresStructuredTaskFrame(t *testing.T) {
	reg := testRegistry(t)
	if err := RegisterInto(reg); err != nil {
		t.Fatalf("register financial report metrics pack: %v", err)
	}
	binding, ok, err := testCoordinator(t).ResolveBinding(reg, PackID, CaseTypeLatest, "")
	if err != nil || !ok {
		t.Fatalf("resolve latest metrics binding: ok=%v err=%v", ok, err)
	}
	err = binding.ValidateCaseInput(map[string]any{
		"user_message":      "帮我查贵州茅台最新财报里的营收和净利润",
		"requested_metrics": []any{"revenue", "net_profit"},
	})
	if err == nil || !strings.Contains(err.Error(), "case_input.entity is required") {
		t.Fatalf("expected latest metrics case input to require entity frame, got %v", err)
	}
	if err := binding.ValidateCaseInput(map[string]any{
		"user_message": "帮我查贵州茅台最新财报里的营收和净利润",
		"entity": map[string]any{
			"name": "贵州茅台",
			"identifiers": map[string]any{
				"stock_code": "600519",
				"exchange":   "SH",
			},
		},
		"requested_metrics": []any{"revenue", "net_profit", "revenue_growth", "net_profit_growth"},
		"requested_outputs": []any{"metrics"},
		"assessment": map[string]any{
			"kind":               AssessmentKindNone,
			"scope":              AssessmentScopeMetricsOnly,
			"requires_valuation": false,
		},
		"period_policy":  "latest_disclosed_annual",
		"source_policy":  "public_web_prefer_official_or_accepted_financial_data_source",
		"freshness":      "live",
		"stop_condition": "guard_passed",
	}); err != nil {
		t.Fatalf("expected structured latest metrics case input to validate: %v", err)
	}
}

func TestTrendMetricsCaseSchemaRequiresStructuredTaskFrame(t *testing.T) {
	reg := testRegistry(t)
	if err := RegisterInto(reg); err != nil {
		t.Fatalf("register financial report metrics pack: %v", err)
	}
	binding, ok, err := testCoordinator(t).ResolveBinding(reg, PackID, CaseTypeTrend, "")
	if err != nil || !ok {
		t.Fatalf("resolve trend metrics binding: ok=%v err=%v", ok, err)
	}
	if err := binding.ValidateCaseInput(map[string]any{
		"user_message": "帮我查腾讯音乐近几年收入和利润趋势",
		"entity": map[string]any{
			"name": "腾讯音乐",
		},
		"requested_metrics": []any{"revenue", "net_profit"},
		"requested_outputs": []any{"metrics", "performance_assessment"},
		"assessment": map[string]any{
			"kind":               AssessmentKindBusinessPerformance,
			"scope":              AssessmentScopeMetricsOnly,
			"requires_valuation": false,
		},
		"period_policy":  "recent_years",
		"source_policy":  "public_web_prefer_official_or_accepted_financial_data_source",
		"freshness":      "live",
		"stop_condition": "guard_passed_or_review_required",
	}); err != nil {
		t.Fatalf("expected structured trend metrics case input to validate: %v", err)
	}
}

func nodeTool(spec agentxworkflow.Spec, nodeID string) string {
	raw, _ := nodeConfigAny(spec, nodeID, "tool").(string)
	return raw
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
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

func hasNodeOutputTarget(spec agentxworkflow.Spec, nodeID string, target string) bool {
	for _, node := range spec.Nodes {
		if node.ID != nodeID {
			continue
		}
		for _, output := range node.Outputs {
			if output.To == target {
				return true
			}
		}
	}
	return false
}

func hasStateSlot(spec agentxworkflow.Spec, name string) bool {
	for _, slot := range spec.StateSchema {
		if slot.Name == name {
			return true
		}
	}
	return false
}

func hasNodeInputBinding(spec agentxworkflow.Spec, nodeID string, from string, to string) bool {
	for _, node := range spec.Nodes {
		if node.ID != nodeID {
			continue
		}
		for _, input := range node.Inputs {
			if input.From == from && input.To == to {
				return true
			}
		}
	}
	return false
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
