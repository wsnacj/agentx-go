package publicnews

import "testing"

func TestBuildLatestNewsBriefCaseInput(t *testing.T) {
	input, ok := BuildLatestNewsBriefCaseInput("帮我找下伊朗战争的最新新闻，说明来源和风险")
	if !ok {
		t.Fatal("expected latest news brief case input to be built")
	}
	if got := input["user_message"]; got != "帮我找下伊朗战争的最新新闻，说明来源和风险" {
		t.Fatalf("unexpected user_message: %#v", got)
	}
	topic, _ := input["topic"].(map[string]any)
	if topic["name"] != "伊朗战争" {
		t.Fatalf("expected topic name 伊朗战争, got %#v", topic["name"])
	}
	fields, _ := input["requested_fields"].([]any)
	if !containsCaseField(fields, "published_at") || !containsCaseField(fields, "source_site") || !containsCaseField(fields, "implications") {
		t.Fatalf("expected requested fields to include published_at, source_site and implications, got %#v", fields)
	}
	if got := input["freshness"]; got != "live" {
		t.Fatalf("expected live freshness, got %#v", got)
	}
	if got := input["stop_condition"]; got != "guard_passed" {
		t.Fatalf("expected guard_passed stop condition, got %#v", got)
	}
	if got := input["cross_check_policy"]; got != "at_least_two_independent_source_sites_for_key_facts" {
		t.Fatalf("expected independent source-site cross-check policy, got %#v", got)
	}
}

func TestBuildLatestNewsBriefCaseInputRejectsNonFreshQuery(t *testing.T) {
	if input, ok := BuildLatestNewsBriefCaseInput("介绍一下伊朗历史背景"); ok || input != nil {
		t.Fatalf("expected non-latest query to be rejected, got ok=%v input=%#v", ok, input)
	}
}

func containsCaseField(fields []any, target string) bool {
	for _, field := range fields {
		if field == target {
			return true
		}
	}
	return false
}
