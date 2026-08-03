package globalstock

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
		t.Fatalf("validate global-stock quote pack: %v", err)
	}
	if !containsString(def.Manifest.PolicyProfiles, "global_stock_comparison_readonly") {
		t.Fatalf("manifest must expose the comparison policy profile: %#v", def.Manifest.PolicyProfiles)
	}
	coordinator := testCoordinator(t)
	spec, err := MaterializedDefaultWorkflow(coordinator)
	if err != nil {
		t.Fatalf("materialize global-stock quote workflow: %v", err)
	}
	if spec.ID != DefaultWorkflow || spec.Pack != PackID {
		t.Fatalf("unexpected materialized workflow identity: %#v", spec)
	}
	if tool := nodeTool(spec, "lookup_quote"); tool != ToolGlobalStockQuoteLookup {
		t.Fatalf("expected lookup_quote to materialize to %q, got %q", ToolGlobalStockQuoteLookup, tool)
	}
	if !nodeHasInput(spec, "lookup_quote", "case.input.user_message", "args.user_message") ||
		!nodeHasInput(spec, "lookup_quote", "case.input.entity.name", "args.entity_name") ||
		!nodeHasInput(spec, "lookup_quote", "case.input.quote_fields", "args.quote_fields") {
		t.Fatalf("lookup_quote missing required task-frame inputs")
	}
	if !nodeHasInput(spec, "format_answer", "case.input.quote_fields", "args.requested_fields") {
		t.Fatalf("format_answer must receive the verified quote field scope")
	}
	if nodeHasInputPrefix(spec, "lookup_quote", "case.input.entity.identifiers.") {
		t.Fatalf("lookup_quote must not bind optional identifiers; adapter should verify subject and market from the required task frame")
	}
	for _, output := range []string{
		"result.subject.entity_name",
		"result.subject.stock_code",
		"result.subject.market",
		"result.subject.exchange",
		"result.subject.currency",
		"result.freshness.as_of",
		"result.evidence.source_url",
		"result.quote.price.value",
		"result.quote.change_percent.value",
		"result.quote.market_cap.value",
	} {
		if nodeHasOutputFrom(spec, "lookup_quote", output) {
			t.Fatalf("lookup_quote must not hard-bind optional degraded payload field %q", output)
		}
	}
	suite, ok := def.EvalSuiteForWorkflow(DefaultWorkflow)
	if !ok {
		t.Fatalf("expected default workflow eval suite")
	}
	if containsString(suite.RequiredState, "quote.subject_name") || containsString(suite.RequiredState, "quote.subject_stock_code") {
		t.Fatalf("eval suite must not require optional subject fields for degraded quote payloads: %#v", suite.RequiredState)
	}
	if suite.PassPath != "" {
		t.Fatalf("eval suite must keep degraded quote payloads as bounded output instead of a failed run, got pass path %q", suite.PassPath)
	}
	comparison, err := MaterializedComparisonWorkflow(coordinator)
	if err != nil {
		t.Fatalf("materialize global-stock comparison workflow: %v", err)
	}
	if comparison.ID != ComparisonWorkflow || comparison.Pack != PackID {
		t.Fatalf("unexpected materialized comparison workflow identity: %#v", comparison)
	}
	if tool := nodeTool(comparison, "lookup_comparison"); tool != ToolGlobalStockInvestigation {
		t.Fatalf("expected lookup_comparison to materialize to %q, got %q", ToolGlobalStockInvestigation, tool)
	}
	if !nodeHasInput(comparison, "lookup_comparison", "case.input.entities", "args.entity_mentions") ||
		!nodeHasInput(comparison, "lookup_comparison", "case.input.requested_fields", "args.requested_fields") ||
		!nodeHasInput(comparison, "lookup_comparison", "case.input.requested_outputs", "args.requested_outputs") {
		t.Fatalf("lookup_comparison missing required multi-subject task-frame inputs")
	}
	if !nodeInputOptional(comparison, "lookup_comparison", "case.input.requested_fields", "args.requested_fields") ||
		!nodeInputOptional(comparison, "lookup_comparison", "case.input.requested_outputs", "args.requested_outputs") {
		t.Fatalf("comparison field/output overrides must remain optional")
	}
	if got := nodeConfigString(comparison, "lookup_comparison", "task_kind"); got != "comparison" {
		t.Fatalf("expected comparison task kind, got %q", got)
	}
	if got := strings.Join(nodeConfigStrings(comparison, "lookup_comparison", "default_requested_fields"), ","); got != "price,pe_ttm,pb,market_cap" {
		t.Fatalf("expected default valuation fields, got %q", got)
	}
	if got := strings.Join(nodeConfigStrings(comparison, "lookup_comparison", "default_requested_outputs"), ","); got != "comparison,valuation_snapshot" {
		t.Fatalf("expected default comparison outputs, got %q", got)
	}
	if got := nodeConfigString(comparison, "format_answer", "answer_kind"); got != "comparison" {
		t.Fatalf("expected comparison answer kind, got %q", got)
	}
}

func TestQuoteCaseSchemaAllowsEntityWithoutIdentifiers(t *testing.T) {
	reg := testRegistry(t)
	if err := RegisterInto(reg); err != nil {
		t.Fatalf("register global-stock quote pack: %v", err)
	}
	binding, ok, err := testCoordinator(t).ResolveBinding(reg, PackID, CaseTypeQuote, "")
	if err != nil || !ok {
		t.Fatalf("resolve quote binding: ok=%v err=%v", ok, err)
	}
	if err := binding.ValidateCaseInput(map[string]any{
		"user_message": "联想集团最近港股行情有什么需要注意？",
		"entity":       map[string]any{"name": "联想集团"},
		"quote_fields": []any{"price", "pe_ttm", "pb", "market_cap"},
	}); err != nil {
		t.Fatalf("entity without identifiers and unused requested_outputs should validate: %v", err)
	}
}

func TestComparisonCaseSchemaAcceptsMultipleEntities(t *testing.T) {
	reg := testRegistry(t)
	if err := RegisterInto(reg); err != nil {
		t.Fatalf("register global-stock quote pack: %v", err)
	}
	binding, ok, err := testCoordinator(t).ResolveBinding(reg, PackID, CaseTypeComparison, ComparisonWorkflow)
	if err != nil || !ok {
		t.Fatalf("resolve comparison binding: ok=%v err=%v", ok, err)
	}
	if err := binding.ValidateCaseInput(map[string]any{
		"user_message": "Compare NVIDIA, AMD, and Intel current valuation snapshot.",
		"entities":     []any{"NVIDIA", "AMD", "Intel"},
	}); err != nil {
		t.Fatalf("comparison defaults should allow omitted inferred fields: %v", err)
	}
	if err := binding.ValidateCaseInput(map[string]any{
		"user_message":      "Compare NVIDIA, AMD, and Intel current valuation snapshot.",
		"entities":          []any{"NVIDIA", "AMD", "Intel"},
		"requested_fields":  []any{"price", "pe_ttm", "pb", "market_cap"},
		"requested_outputs": []any{"comparison", "valuation_snapshot"},
	}); err != nil {
		t.Fatalf("multi-subject comparison input should validate: %v", err)
	}
}

func TestDefinitionRoutesComparisonWithoutChangingSingleSubjectDefault(t *testing.T) {
	reg := testRegistry(t)
	if err := RegisterInto(reg); err != nil {
		t.Fatalf("register global-stock quote pack: %v", err)
	}
	selection, ok := agentxpack.SelectBinding(reg, "Compare NVIDIA, AMD, and Intel current valuation snapshot. If a field is unavailable, say which one.", agentxpack.SelectOptions{})
	if !ok {
		t.Fatalf("expected multi-subject comparison route, got %#v", selection)
	}
	if selection.Selected.CaseType != CaseTypeComparison || selection.Selected.WorkflowID != ComparisonWorkflow {
		t.Fatalf("expected comparison workflow, got %#v", selection.Selected)
	}

	selection, ok = agentxpack.SelectBinding(reg, "Compare Adobe (ADBE), Salesforce (CRM), and ServiceNow (NOW) current valuation snapshots. Include price, PE, PB, market cap, timestamp, and identify every unavailable field.", agentxpack.SelectOptions{})
	if !ok || selection.Selected.CaseType != CaseTypeComparison || selection.Selected.WorkflowID != ComparisonWorkflow {
		t.Fatalf("field names must not override an explicit multi-subject comparison route, got %#v", selection)
	}

	selection, ok = agentxpack.SelectBinding(reg, "NVIDIA current valuation snapshot.", agentxpack.SelectOptions{})
	if !ok {
		t.Fatalf("expected single-subject valuation route, got %#v", selection)
	}
	if selection.Selected.CaseType != CaseTypeValuation || selection.Selected.WorkflowID != DefaultWorkflow {
		t.Fatalf("expected default single-subject workflow, got %#v", selection.Selected)
	}

	selection, ok = agentxpack.SelectBinding(reg, "Adobe current valuation snapshot with PE, PB, and market cap.", agentxpack.SelectOptions{})
	if !ok || selection.Selected.CaseType != CaseTypeValuation || selection.Selected.WorkflowID != DefaultWorkflow {
		t.Fatalf("field names must preserve the single-subject valuation route, got %#v", selection)
	}

	selection, ok = agentxpack.SelectBinding(reg, "给我一个 Apple 当前估值快照，包含价格、PE、PB。", agentxpack.SelectOptions{})
	if !ok {
		t.Fatalf("expected Chinese single-subject valuation route, got %#v", selection)
	}
	if selection.Selected.CaseType != CaseTypeValuation || selection.Selected.WorkflowID != DefaultWorkflow {
		t.Fatalf("expected Chinese prompt to keep the default single-subject workflow, got %#v", selection.Selected)
	}
}

func nodeTool(spec agentxworkflow.Spec, nodeID string) string {
	for _, node := range spec.Nodes {
		if node.ID != nodeID {
			continue
		}
		raw, _ := node.Config["tool"].(string)
		return raw
	}
	return ""
}

func nodeConfigString(spec agentxworkflow.Spec, nodeID string, key string) string {
	for _, node := range spec.Nodes {
		if node.ID != nodeID {
			continue
		}
		value, _ := node.Config[key].(string)
		return value
	}
	return ""
}

func nodeConfigStrings(spec agentxworkflow.Spec, nodeID string, key string) []string {
	for _, node := range spec.Nodes {
		if node.ID != nodeID {
			continue
		}
		values, _ := node.Config[key].([]any)
		out := make([]string, 0, len(values))
		for _, value := range values {
			item, _ := value.(string)
			out = append(out, item)
		}
		return out
	}
	return nil
}

func nodeHasInput(spec agentxworkflow.Spec, nodeID string, from string, to string) bool {
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

func nodeInputOptional(spec agentxworkflow.Spec, nodeID string, from string, to string) bool {
	for _, node := range spec.Nodes {
		if node.ID != nodeID {
			continue
		}
		for _, input := range node.Inputs {
			if input.From == from && input.To == to {
				return input.Optional
			}
		}
	}
	return false
}

func nodeHasInputPrefix(spec agentxworkflow.Spec, nodeID string, fromPrefix string) bool {
	for _, node := range spec.Nodes {
		if node.ID != nodeID {
			continue
		}
		for _, input := range node.Inputs {
			if strings.HasPrefix(input.From, fromPrefix) {
				return true
			}
		}
	}
	return false
}

func nodeHasOutputFrom(spec agentxworkflow.Spec, nodeID string, from string) bool {
	for _, node := range spec.Nodes {
		if node.ID != nodeID {
			continue
		}
		for _, output := range node.Outputs {
			if output.From == from {
				return true
			}
		}
	}
	return false
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
