package publicnews

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

func TestDefinitionValidatesAndMaterializes(t *testing.T) {
	def := Definition()
	coordinator := testCoordinator(t)
	if err := coordinator.ValidateDefinition(def); err != nil {
		t.Fatalf("validate public news brief pack: %v", err)
	}

	spec, err := MaterializedDefaultWorkflow(coordinator)
	if err != nil {
		t.Fatalf("materialize public news brief workflow: %v", err)
	}
	if spec.ID != DefaultWorkflow || spec.Pack != PackID {
		t.Fatalf("unexpected materialized workflow identity: %#v", spec)
	}

	if tool := nodeTool(spec, "lookup_news"); tool != "latest_news_lookup" {
		t.Fatalf("expected lookup_news to materialize to latest_news_lookup, got %q", tool)
	}
	if kind := nodeKind(spec, "lookup_news"); kind != agentxworkflow.NodeTool {
		t.Fatalf("expected lookup_news to be a tool node, got %q", kind)
	}
	if target := nodeInputTarget(spec, "lookup_news", 0); target != "args.user_message" {
		t.Fatalf("expected lookup_news first input target args.user_message, got %q", target)
	}
	if !nodeHasInput(spec, "lookup_news", "case.input.topic.name", "args.topic") {
		t.Fatalf("expected lookup_news to bind structured topic name")
	}
	if !nodeHasOutput(spec, "lookup_news", "result.passed", "state.news.passed") {
		t.Fatalf("expected lookup_news to project top-level guard pass state")
	}
	if !nodeHasOutput(spec, "lookup_news", "result.source_url", "state.news.source_url") {
		t.Fatalf("expected lookup_news to project top-level source URL")
	}
	if nodeTool(spec, "select_candidate") != "" || nodeKind(spec, "select_candidate") != "" {
		t.Fatalf("did not expect default public-news workflow to retain llm_task candidate selection")
	}
	if suite, ok := def.EvalSuiteForWorkflow(DefaultWorkflow); !ok || suite.PassPath != "news.passed" {
		t.Fatalf("expected default eval suite to gate on news.passed, got ok=%v suite=%#v", ok, suite)
	}
}

func TestLatestNewsBriefCaseSchemaRequiresStructuredTaskFrame(t *testing.T) {
	coordinator := testCoordinator(t)
	reg, err := agentxpack.NewMemoryRegistry(coordinator)
	if err != nil {
		t.Fatalf("new pack registry: %v", err)
	}
	if err := RegisterInto(reg); err != nil {
		t.Fatalf("register public news brief pack: %v", err)
	}
	binding, ok, err := coordinator.ResolveBinding(reg, PackID, CaseTypeLatestBrief, "")
	if err != nil || !ok {
		t.Fatalf("resolve latest news brief binding: ok=%v err=%v", ok, err)
	}
	err = binding.ValidateCaseInput(map[string]any{
		"user_message":     "帮我找下伊朗战争的最新新闻",
		"requested_fields": []any{"headline", "published_at", "key_update", "source_url"},
	})
	if err == nil || !strings.Contains(err.Error(), "case_input.topic is required") {
		t.Fatalf("expected latest news brief case input to require topic frame, got %v", err)
	}
	if err := binding.ValidateCaseInput(map[string]any{
		"user_message": "帮我找下伊朗战争的最新新闻",
		"topic": map[string]any{
			"name":     "伊朗战争",
			"entities": []any{"伊朗"},
		},
		"requested_fields":   []any{"headline", "published_at", "key_update", "source_url"},
		"source_policy":      "public_web_prefer_official_or_authoritative_news_source",
		"freshness":          "live",
		"cross_check_policy": "at_least_two_independent_source_sites_for_key_facts",
		"stop_condition":     "guard_passed",
	}); err != nil {
		t.Fatalf("expected structured latest news brief case input to validate: %v", err)
	}
}

func nodeTool(spec agentxworkflow.Spec, nodeID string) string {
	raw, _ := nodeConfigAny(spec, nodeID, "tool").(string)
	return raw
}

func nodeKind(spec agentxworkflow.Spec, nodeID string) agentxworkflow.NodeKind {
	for _, node := range spec.Nodes {
		if node.ID == nodeID {
			return node.Kind
		}
	}
	return ""
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

func nodeHasOutput(spec agentxworkflow.Spec, nodeID string, from string, to string) bool {
	for _, node := range spec.Nodes {
		if node.ID != nodeID {
			continue
		}
		for _, output := range node.Outputs {
			if output.From == from && output.To == to {
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
