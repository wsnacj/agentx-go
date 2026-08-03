package retrieval

import (
	"strings"
	"testing"
	"time"
)

func TestPrepareSearchRequestProviderOverrideAccepted(t *testing.T) {
	t.Parallel()

	plan, validation := PrepareSearchRequest(SearchPrepareOptions{
		ConfiguredProvider: "baidu",
		RequestedProvider:  "brave",
		Query:              "agentx",
		Count:              2,
		CanUseProvider: func(provider string) bool {
			return provider == "brave"
		},
	})
	if validation != nil {
		t.Fatalf("unexpected validation: %#v", validation)
	}
	if plan.EffectiveProvider != "brave" || plan.RequestedProvider != "brave" {
		t.Fatalf("unexpected plan: %#v", plan)
	}
	if !strings.Contains(strings.ToLower(plan.ProviderNote), "accepted") {
		t.Fatalf("expected accepted provider note, got %#v", plan)
	}
}

func TestPrepareSearchRequestProviderOverrideIgnored(t *testing.T) {
	t.Parallel()

	plan, validation := PrepareSearchRequest(SearchPrepareOptions{
		ConfiguredProvider: "baidu",
		RequestedProvider:  "brave",
		Query:              "agentx",
		Count:              1,
	})
	if validation != nil {
		t.Fatalf("unexpected validation: %#v", validation)
	}
	if plan.EffectiveProvider != "baidu" || plan.RequestedProvider != "brave" {
		t.Fatalf("unexpected plan: %#v", plan)
	}
	if !strings.Contains(strings.ToLower(plan.ProviderNote), "ignored") {
		t.Fatalf("expected ignored provider note, got %#v", plan)
	}
}

func TestPrepareSearchRequestDateFilterSwitchesToPerplexity(t *testing.T) {
	t.Parallel()

	plan, validation := PrepareSearchRequest(SearchPrepareOptions{
		ConfiguredProvider:                    "baidu",
		RequestedProvider:                     "baidu",
		Query:                                 "agentx latest",
		Count:                                 1,
		DateAfter:                             "2026-03-09",
		CanUsePerplexityStructuredDateFilters: true,
		CanUseProvider: func(provider string) bool {
			return provider == "baidu"
		},
	})
	if validation != nil {
		t.Fatalf("unexpected validation: %#v", validation)
	}
	if plan.EffectiveProvider != "perplexity" {
		t.Fatalf("expected perplexity fallback, got %#v", plan)
	}
	if plan.PerplexityDateAfter != "3/9/2026" {
		t.Fatalf("expected formatted perplexity date, got %#v", plan)
	}
	if !strings.Contains(strings.ToLower(plan.ProviderNote), "switched to perplexity") {
		t.Fatalf("expected switch note, got %#v", plan)
	}
}

func TestPrepareSearchRequestDateFilterIgnoredWithoutFallback(t *testing.T) {
	t.Parallel()

	plan, validation := PrepareSearchRequest(SearchPrepareOptions{
		ConfiguredProvider: "baidu",
		RequestedProvider:  "baidu",
		Query:              "agentx latest",
		Count:              1,
		DateAfter:          "2026-03-09",
		CanUseProvider: func(provider string) bool {
			return provider == "baidu"
		},
	})
	if validation != nil {
		t.Fatalf("unexpected validation: %#v", validation)
	}
	if plan.EffectiveProvider != "baidu" || plan.DateAfter != "" || plan.DateBefore != "" {
		t.Fatalf("expected ignored date filter, got %#v", plan)
	}
	if !strings.Contains(strings.ToLower(plan.ProviderNote), "ignored") {
		t.Fatalf("expected ignored note, got %#v", plan)
	}
}

func TestPrepareSearchRequestBuildsBraveDateRangeFreshness(t *testing.T) {
	t.Parallel()

	plan, validation := PrepareSearchRequest(SearchPrepareOptions{
		ConfiguredProvider: "brave",
		Query:              "agentx",
		Count:              1,
		DateAfter:          "2026-02-01",
		Now: func() time.Time {
			return time.Date(2026, 2, 7, 12, 0, 0, 0, time.UTC)
		},
	})
	if validation != nil {
		t.Fatalf("unexpected validation: %#v", validation)
	}
	if plan.Freshness != "2026-02-01to2026-02-07" {
		t.Fatalf("unexpected freshness: %#v", plan)
	}
}

func TestPrepareSearchRequestRejectsUnsupportedOpenRouterDateFilter(t *testing.T) {
	t.Parallel()

	_, validation := PrepareSearchRequest(SearchPrepareOptions{
		ConfiguredProvider: "openrouter",
		Query:              "agentx",
		Count:              1,
		DateAfter:          "2026-03-09",
	})
	if validation == nil || validation.Code != "unsupported_date_filter" {
		t.Fatalf("expected openrouter date filter validation, got %#v", validation)
	}
}

func TestPrepareSearchRequestRejectsUnsupportedDomainFilter(t *testing.T) {
	t.Parallel()

	_, validation := PrepareSearchRequest(SearchPrepareOptions{
		ConfiguredProvider: "brave",
		Query:              "agentx",
		Count:              1,
		DomainFilter:       []string{"nature.com"},
	})
	if validation == nil || validation.Code != "unsupported_domain_filter" {
		t.Fatalf("expected unsupported domain filter validation, got %#v", validation)
	}
}
