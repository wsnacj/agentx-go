package publicnews

import (
	"strings"
	"testing"
)

func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func TestLatestNewsLookupAnswerContractPassedDraft(t *testing.T) {
	payload := LatestNewsLookupPayload{
		Tool:                 ToolLatestNewsLookup,
		GuardStatus:          "passed",
		NewsFieldsReady:      true,
		CrossCheckReady:      true,
		Passed:               true,
		FreshnessConfirmed:   true,
		SourceAccepted:       true,
		StopAfterGuardPassed: true,
		SourceURL:            "https://news.example.com/a",
		PublishedAt:          "2026-05-18T09:00:00",
		Summary:              "示例事件出现最新进展",
		Intent: LatestNewsLookupIntent{
			Topic:            "示例事件最新进展",
			EntityMentions:   []string{"示例事件"},
			RequestedOutputs: []string{"brief", "impact_assessment", "risk_summary"},
		},
		Guard: &Payload{
			SupportingSources: []Evidence{{
				SourceURL: "https://wire.example.net/b",
			}},
		},
	}
	contract := LatestNewsLookupAnswerContract(payload)
	if contract == nil || !contract.FinalAnswerRecommended || contract.Reason != "guard_passed" {
		t.Fatalf("expected passed answer contract, got %#v", contract)
	}
	if !strings.Contains(contract.PossibleImpact, "不足以独立量化") || !strings.Contains(contract.RiskBoundary, "已打开并通过 guard") {
		t.Fatalf("expected structured impact and risk boundary, got %#v", contract)
	}
	for _, expected := range []string{"查询目标：示例事件", "一句话摘要：示例事件出现最新进展。", "发布时间：2026-05-18T09:00:00", "来源：https://news.example.com/a", "交叉核对：https://wire.example.net/b", "可能影响：", "不足以独立量化", "风险边界："} {
		if !strings.Contains(contract.FinalAnswerDraft, expected) {
			t.Fatalf("expected draft to contain %q, got %q", expected, contract.FinalAnswerDraft)
		}
	}
	if !containsString(contract.DoNotRetryTools, ToolLatestNewsLookup) || !containsString(contract.DoNotRetryTools, "search") {
		t.Fatalf("expected do-not-retry tool list, got %#v", contract.DoNotRetryTools)
	}
}

func TestLatestNewsLookupAnswerContractFallsBackToTopicTarget(t *testing.T) {
	payload := LatestNewsLookupPayload{
		Tool:                 ToolLatestNewsLookup,
		GuardStatus:          "passed",
		NewsFieldsReady:      true,
		CrossCheckReady:      true,
		Passed:               true,
		FreshnessConfirmed:   true,
		SourceAccepted:       true,
		StopAfterGuardPassed: true,
		SourceURL:            "https://news.example.com/a",
		PublishedAt:          "2026-05-18T09:00:00",
		Summary:              "示例事件出现最新进展",
		Intent: LatestNewsLookupIntent{
			Topic: "腾讯云 AI Agent 产品",
		},
	}
	contract := LatestNewsLookupAnswerContract(payload)
	if contract == nil {
		t.Fatalf("expected answer contract")
	}
	if !strings.Contains(contract.FinalAnswerDraft, "查询目标：腾讯云 AI Agent 产品") {
		t.Fatalf("expected draft to include topic target, got %q", contract.FinalAnswerDraft)
	}
}

func TestLatestNewsLookupAnswerContractProviderUnavailableDraft(t *testing.T) {
	payload := LatestNewsLookupPayload{
		Tool:                  ToolLatestNewsLookup,
		AdapterStatus:         "provider_unavailable",
		FailureCode:           "search_provider_failure_missing",
		FailureClass:          "auth_missing",
		RetrySuppressedReason: "terminal_failure_class:auth_missing",
		Sources: &LatestNewsSourcesPayload{
			Provider:          "brave",
			EffectiveProvider: "brave",
			ProviderStatus:    "missing_credentials",
			FallbackHint:      "configure BRAVE_API_KEY",
		},
	}
	contract := LatestNewsLookupAnswerContract(payload)
	if contract == nil || !contract.FinalAnswerRecommended || contract.Reason != "search_provider_config_invalid" {
		t.Fatalf("expected provider-unavailable answer contract, got %#v", contract)
	}
	if !strings.Contains(contract.RiskBoundary, "不能把模型猜测") {
		t.Fatalf("expected structured provider risk boundary, got %#v", contract)
	}
	for _, expected := range []string{"当前不能完成这次最新新闻查询", "provider=brave", "missing_credentials", "auth_missing", "terminal_failure_class:auth_missing", "configure BRAVE_API_KEY"} {
		if !strings.Contains(contract.FinalAnswerDraft, expected) {
			t.Fatalf("expected provider draft to contain %q, got %q", expected, contract.FinalAnswerDraft)
		}
	}
	if !containsString(contract.DoNotRetryTools, ToolLatestNewsLookup) || !containsString(contract.DoNotRetryTools, "search") {
		t.Fatalf("config/auth provider failures should block repeated lookup/search, got %#v", contract.DoNotRetryTools)
	}
}

func TestLatestNewsLookupAnswerContractNeedsReviewDoesNotEmit(t *testing.T) {
	payload := LatestNewsLookupPayload{
		Tool:            ToolLatestNewsLookup,
		GuardStatus:     "needs_cross_check",
		NewsFieldsReady: true,
		CrossCheckReady: false,
		Passed:          false,
	}
	if contract := LatestNewsLookupAnswerContract(payload); contract != nil {
		t.Fatalf("did not expect answer contract for review payload, got %#v", contract)
	}
}

func TestLatestNewsLookupAnswerContractSourceQualityDiagnosticDraft(t *testing.T) {
	payload := LatestNewsLookupPayload{
		Tool:              ToolLatestNewsLookup,
		AdapterStatus:     "needs_review",
		FailureCode:       "latest_news_missing_fields",
		GuardStatus:       "missing_news_fields",
		NewsFieldsReady:   false,
		CrossCheckReady:   false,
		Passed:            false,
		SourceURL:         "https://news.example.com/list",
		PublishedAt:       "2026-05-23T07:19:15",
		Summary:           "候选来源提到示例事件，但正文未打开",
		MissingNewsFields: []string{"grounded_page_text"},
		ReviewReasons:     []string{"single_source_only"},
		Intent: LatestNewsLookupIntent{
			EntityMentions:   []string{"NVIDIA", "英伟达"},
			RequestedOutputs: []string{"brief", "impact_assessment", "risk_summary"},
		},
	}
	contract := LatestNewsLookupAnswerContract(payload)
	if contract == nil ||
		!contract.FinalAnswerRecommended ||
		contract.Reason != "source_quality_needs_review" ||
		contract.AllowedSummaryScope != LatestNewsAnswerScopeSourceDiagnostic {
		t.Fatalf("expected source-quality diagnostic answer contract, got %#v", contract)
	}
	for _, expected := range []string{"不能把这次查询包装成已完整核验", "查询目标：NVIDIA、英伟达", "https://news.example.com/list", "可能影响：", "不能可靠判断", "风险边界：", "grounded_page_text", "single_source_only"} {
		if !strings.Contains(contract.FinalAnswerDraft, expected) {
			t.Fatalf("expected source-quality draft to contain %q, got %q", expected, contract.FinalAnswerDraft)
		}
	}
	if contract.PossibleImpact == "" {
		t.Fatalf("expected bounded impact field, got %#v", contract)
	}
	if !containsString(contract.DoNotRetryTools, "open_page") || !containsString(contract.DoNotRetryTools, "web_fetch") || !containsString(contract.DoNotRetryTools, "search") {
		t.Fatalf("expected source-quality diagnostic to block repeated low-level recovery, got %#v", contract.DoNotRetryTools)
	}
}

func TestLatestNewsLookupAnswerContractNoSourceDiagnosticDraft(t *testing.T) {
	payload := LatestNewsLookupPayload{
		Tool:              ToolLatestNewsLookup,
		AdapterStatus:     "evidence_incomplete",
		FailureCode:       "latest_news_search_open_page_no_grounded_sources",
		FailureClass:      "evidence_missing",
		GuardStatus:       "missing_news_fields",
		MissingNewsFields: []string{"source_url", "grounded_page_text"},
		ReviewReasons:     []string{"ungrounded_news_fields", "no_usable_source"},
		SourceURL:         "unknown",
		Intent: LatestNewsLookupIntent{
			EntityMentions:   []string{"欧盟AI监管"},
			RequestedOutputs: []string{"brief", "impact_assessment", "risk_summary"},
		},
	}
	contract := LatestNewsLookupAnswerContract(payload)
	if contract == nil || contract.Reason != "source_quality_needs_review" {
		t.Fatalf("expected no-source diagnostic answer contract, got %#v", contract)
	}
	for _, expected := range []string{"未取得可打开且可 grounding 的候选来源", "未取得可核验的事件证据", "no_usable_source", "不能据此推导任何事件事实或影响"} {
		if !strings.Contains(contract.FinalAnswerDraft, expected) {
			t.Fatalf("expected no-source draft to contain %q, got %q", expected, contract.FinalAnswerDraft)
		}
	}
	for _, forbidden := range []string{"已发现候选来源", "候选事件的实际影响", "single_source_only"} {
		if strings.Contains(contract.FinalAnswerDraft, forbidden) {
			t.Fatalf("no-source draft must not claim %q, got %q", forbidden, contract.FinalAnswerDraft)
		}
	}
}

func TestLatestNewsLookupAnswerContractSearchSnippetPrimaryDoesNotEmit(t *testing.T) {
	payload := LatestNewsLookupPayload{
		Tool:            ToolLatestNewsLookup,
		GuardStatus:     "passed",
		NewsFieldsReady: true,
		CrossCheckReady: true,
		Passed:          true,
		Warnings:        []string{"latest_news_search_snippet_primary_used"},
	}
	if contract := LatestNewsLookupAnswerContract(payload); contract != nil {
		t.Fatalf("did not expect terminal answer contract for snippet-primary evidence, got %#v", contract)
	}
}

func TestLatestNewsLookupAnswerContractSearchSnippetSupportingDoesNotEmit(t *testing.T) {
	payload := LatestNewsLookupPayload{
		Tool:            ToolLatestNewsLookup,
		GuardStatus:     "passed",
		NewsFieldsReady: true,
		CrossCheckReady: true,
		Passed:          true,
		Sources: &LatestNewsSourcesPayload{
			Warnings: []string{"latest_news_search_snippet_supporting_source_used"},
		},
	}
	if contract := LatestNewsLookupAnswerContract(payload); contract != nil {
		t.Fatalf("did not expect terminal answer contract for snippet-supporting evidence, got %#v", contract)
	}
}

func TestLatestNewsLookupAnswerContractAllowsSnippetWarningAfterGuardGrounding(t *testing.T) {
	payload := LatestNewsLookupPayload{
		Tool:            ToolLatestNewsLookup,
		GuardStatus:     "passed",
		NewsFieldsReady: true,
		CrossCheckReady: true,
		Passed:          true,
		Warnings:        []string{"latest_news_search_snippet_supporting_source_used"},
		SourceURL:       "https://news.example.com/a",
		PublishedAt:     "2026-05-18T09:00:00",
		Summary:         "示例事件出现最新进展",
		Guard: &Payload{
			Evidence: Evidence{
				SourceURL:             "https://news.example.com/a",
				SourceSite:            "news.example.com",
				PublishedAt:           "2026-05-18T09:00:00",
				KeyUpdate:             "示例事件出现最新进展",
				GroundedTextAvailable: true,
			},
			SupportingSources: []Evidence{{
				SourceURL:             "https://wire.example.net/b",
				SourceSite:            "wire.example.net",
				PublishedAt:           "2026-05-18T09:05:00",
				KeyUpdate:             "第二来源确认示例事件出现最新进展",
				GroundedTextAvailable: true,
			}},
		},
	}
	contract := LatestNewsLookupAnswerContract(payload)
	if contract == nil || !contract.FinalAnswerRecommended || contract.Reason != "guard_passed" {
		t.Fatalf("expected grounded guard evidence to allow answer contract despite stale snippet warning, got %#v", contract)
	}
}
