package retrieval

import "testing"

func TestSearchProviderCapabilities(t *testing.T) {
	capabilities := SearchProviderCapabilities([]string{"", "brave", "perplexity", "openrouter", "doubao_custom", "brave"})
	if len(capabilities) != 4 {
		t.Fatalf("expected deduped capabilities, got %#v", capabilities)
	}
	byProvider := map[string]ProviderCapability{}
	for _, item := range capabilities {
		byProvider[item.Provider] = item
	}
	if !byProvider["brave"].DateFilters || !byProvider["brave"].Locale || !byProvider["brave"].StructuredSnippets {
		t.Fatalf("unexpected brave capability: %#v", byProvider["brave"])
	}
	if !byProvider["perplexity"].DomainFilters || !byProvider["perplexity"].SynthesizedCitations {
		t.Fatalf("unexpected perplexity capability: %#v", byProvider["perplexity"])
	}
	if byProvider["openrouter"].DomainFilters {
		t.Fatalf("openrouter compatibility path should not claim native domain filtering: %#v", byProvider["openrouter"])
	}
	custom := byProvider[SearchProviderDoubaoCustom]
	if !custom.DateFilters || !custom.DomainFilters || !custom.AuthorityFilter || !custom.QueryRewrite || custom.Cache || custom.MaxResults != 50 {
		t.Fatalf("unexpected doubao custom capability: %#v", custom)
	}
}

func TestBuildProviderDiagnosticsNormalizesHealthAndFallback(t *testing.T) {
	diagnostics := BuildProviderDiagnostics(ProviderDiagnosticsInput{
		Tool:              "search",
		RequestedProvider: "Perplexity",
		EffectiveProvider: "Brave",
		ProviderNote:      "provider note",
		Capabilities:      SearchProviderCapabilities([]string{"brave"}),
		Health: []ProviderHealth{
			{Provider: "brave", Available: true},
			{Provider: "perplexity", Status: "missing_credentials", RequiresEnv: []string{"PERPLEXITY_API_KEY"}},
		},
		FallbackFrom:   "perplexity",
		FallbackTo:     "brave",
		FallbackReason: "missing_credentials",
		FallbackHint:   "retry with brave",
	})
	if diagnostics.RequestedProvider != "perplexity" || diagnostics.EffectiveProvider != "brave" {
		t.Fatalf("unexpected providers: %#v", diagnostics)
	}
	if diagnostics.Fallback == nil || diagnostics.Fallback.Reason != ProviderHealthCredentialMissing {
		t.Fatalf("unexpected fallback: %#v", diagnostics.Fallback)
	}
	if len(diagnostics.Health) != 2 || diagnostics.Health[1].Status != ProviderHealthCredentialMissing {
		t.Fatalf("unexpected health diagnostics: %#v", diagnostics.Health)
	}
}

func TestFetchProviderCapabilitiesAreToolSpecific(t *testing.T) {
	webFetchCaps := FetchProviderCapabilities("web_fetch", []string{ProviderDirectHTTP, ProviderFirecrawl, ProviderDirectHTTP})
	if len(webFetchCaps) != 2 {
		t.Fatalf("expected deduped fetch capabilities, got %#v", webFetchCaps)
	}
	byProvider := map[string]ProviderCapability{}
	for _, item := range webFetchCaps {
		byProvider[item.Provider] = item
	}
	if !byProvider[ProviderDirectHTTP].ContentExtraction ||
		!byProvider[ProviderDirectHTTP].PDFBinaryHandoff ||
		!byProvider[ProviderDirectHTTP].Cache {
		t.Fatalf("unexpected web_fetch direct_http capability: %#v", byProvider[ProviderDirectHTTP])
	}
	if !byProvider[ProviderFirecrawl].ContentExtraction || !byProvider[ProviderFirecrawl].Cache {
		t.Fatalf("unexpected firecrawl capability: %#v", byProvider[ProviderFirecrawl])
	}

	httpCaps := FetchProviderCapabilities("http_request", []string{ProviderDirectHTTP})
	if len(httpCaps) != 1 {
		t.Fatalf("expected one http_request capability, got %#v", httpCaps)
	}
	if httpCaps[0].ContentExtraction || httpCaps[0].PDFBinaryHandoff || httpCaps[0].Cache {
		t.Fatalf("http_request direct_http should not claim extraction/cache capability: %#v", httpCaps[0])
	}
}

func TestBuildFetchProviderDiagnosticsUsesSharedSchema(t *testing.T) {
	diagnostics := BuildFetchProviderDiagnostics(FetchProviderDiagnosticsInput{
		Tool:              "web_fetch",
		EffectiveProvider: ProviderFirecrawl,
		IncludeFirecrawl:  true,
		FallbackFrom:      ProviderDirectHTTP,
		FallbackTo:        ProviderFirecrawl,
		FallbackReason:    ProviderHealthRequestFailed,
		FallbackHint:      "external extractor recovered content",
	})
	if diagnostics == nil ||
		diagnostics.Tool != "web_fetch" ||
		diagnostics.EffectiveProvider != ProviderFirecrawl ||
		len(diagnostics.Capabilities) != 2 ||
		len(diagnostics.Health) != 2 {
		t.Fatalf("unexpected fetch provider diagnostics: %#v", diagnostics)
	}
	if diagnostics.Fallback == nil ||
		diagnostics.Fallback.From != ProviderDirectHTTP ||
		diagnostics.Fallback.To != ProviderFirecrawl ||
		diagnostics.Fallback.Reason != ProviderHealthRequestFailed {
		t.Fatalf("unexpected fallback diagnostics: %#v", diagnostics.Fallback)
	}
}
