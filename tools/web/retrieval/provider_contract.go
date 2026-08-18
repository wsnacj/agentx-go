package retrieval

import (
	"sort"
	"strings"
)

const (
	ProviderHealthAvailable          = "available"
	ProviderHealthCredentialMissing  = "credential_missing"
	ProviderHealthCredentialInvalid  = "credential_invalid"
	ProviderHealthQuotaLimited       = "quota_limited"
	ProviderHealthRateLimited        = "rate_limited"
	ProviderHealthNetworkBlocked     = "network_blocked"
	ProviderHealthUnsupportedFeature = "unsupported_feature"
	ProviderHealthStatusError        = "status_error"
	ProviderHealthDecodeError        = "decode_error"
	ProviderHealthRequestFailed      = "request_failed"
	ProviderHealthTimeout            = "timeout"
	ProviderHealthUnknown            = "unknown"
)

const (
	ProviderDirectHTTP = "direct_http"
	ProviderFirecrawl  = "firecrawl"
)

type ProviderCapability struct {
	Provider             string `json:"provider"`
	Kind                 string `json:"kind,omitempty"`
	DateFilters          bool   `json:"date_filters"`
	DomainFilters        bool   `json:"domain_filters"`
	AuthorityFilter      bool   `json:"authority_filter"`
	QueryRewrite         bool   `json:"query_rewrite"`
	Locale               bool   `json:"locale"`
	StructuredSnippets   bool   `json:"structured_snippets"`
	SynthesizedCitations bool   `json:"synthesized_citations"`
	ContentExtraction    bool   `json:"content_extraction"`
	PDFBinaryHandoff     bool   `json:"pdf_binary_handoff"`
	Cache                bool   `json:"cache"`
	MaxResults           int    `json:"max_results,omitempty"`
}

type ProviderHealth struct {
	Provider    string   `json:"provider"`
	Available   bool     `json:"available"`
	Configured  bool     `json:"configured"`
	Status      string   `json:"status"`
	Reason      string   `json:"reason,omitempty"`
	RequiresEnv []string `json:"requires_env,omitempty"`
	Retryable   bool     `json:"retryable,omitempty"`
}

type ProviderFallback struct {
	From   string `json:"from,omitempty"`
	To     string `json:"to,omitempty"`
	Reason string `json:"reason,omitempty"`
	Hint   string `json:"hint,omitempty"`
}

type ProviderDiagnostics struct {
	Tool              string               `json:"tool,omitempty"`
	RequestedProvider string               `json:"requested_provider,omitempty"`
	EffectiveProvider string               `json:"effective_provider,omitempty"`
	ProviderNote      string               `json:"provider_note,omitempty"`
	Capabilities      []ProviderCapability `json:"capabilities,omitempty"`
	Health            []ProviderHealth     `json:"health,omitempty"`
	Fallback          *ProviderFallback    `json:"fallback,omitempty"`
}

type ProviderDiagnosticsInput struct {
	Tool              string
	RequestedProvider string
	EffectiveProvider string
	ProviderNote      string
	Capabilities      []ProviderCapability
	Health            []ProviderHealth
	FallbackFrom      string
	FallbackTo        string
	FallbackReason    string
	FallbackHint      string
}

type FetchProviderDiagnosticsInput struct {
	Tool              string
	RequestedProvider string
	EffectiveProvider string
	ProviderNote      string
	IncludeFirecrawl  bool
	FallbackFrom      string
	FallbackTo        string
	FallbackReason    string
	FallbackHint      string
}

func BuildProviderDiagnostics(input ProviderDiagnosticsInput) ProviderDiagnostics {
	out := ProviderDiagnostics{
		Tool:              strings.TrimSpace(input.Tool),
		RequestedProvider: normalizeProviderOptional(input.RequestedProvider),
		EffectiveProvider: normalizeProviderOptional(input.EffectiveProvider),
		ProviderNote:      strings.TrimSpace(input.ProviderNote),
		Capabilities:      NormalizeProviderCapabilities(input.Capabilities),
		Health:            NormalizeProviderHealth(input.Health),
	}
	fallback := ProviderFallback{
		From:   normalizeProviderOptional(input.FallbackFrom),
		To:     normalizeProviderOptional(input.FallbackTo),
		Reason: NormalizeProviderHealthStatus(input.FallbackReason),
		Hint:   strings.TrimSpace(input.FallbackHint),
	}
	if fallback.From != "" || fallback.To != "" || fallback.Reason != "" || fallback.Hint != "" {
		out.Fallback = &fallback
	}
	return out
}

func BuildFetchProviderDiagnostics(input FetchProviderDiagnosticsInput) *ProviderDiagnostics {
	tool := strings.TrimSpace(input.Tool)
	effectiveProvider := normalizeProviderOptional(input.EffectiveProvider)
	if effectiveProvider == "" {
		effectiveProvider = ProviderDirectHTTP
	}
	providers := []string{ProviderDirectHTTP, effectiveProvider}
	if input.IncludeFirecrawl {
		providers = append(providers, ProviderFirecrawl)
	}
	health := []ProviderHealth{availableProviderHealth(ProviderDirectHTTP)}
	if effectiveProvider == ProviderFirecrawl || input.IncludeFirecrawl {
		health = append(health, availableProviderHealth(ProviderFirecrawl))
	}
	diagnostics := BuildProviderDiagnostics(ProviderDiagnosticsInput{
		Tool:              tool,
		RequestedProvider: input.RequestedProvider,
		EffectiveProvider: effectiveProvider,
		ProviderNote:      input.ProviderNote,
		Capabilities:      FetchProviderCapabilities(tool, providers),
		Health:            health,
		FallbackFrom:      input.FallbackFrom,
		FallbackTo:        input.FallbackTo,
		FallbackReason:    input.FallbackReason,
		FallbackHint:      input.FallbackHint,
	})
	return &diagnostics
}

func CloneProviderDiagnostics(input ProviderDiagnostics) ProviderDiagnostics {
	out := input
	if len(input.Capabilities) > 0 {
		out.Capabilities = append([]ProviderCapability(nil), input.Capabilities...)
	}
	if len(input.Health) > 0 {
		out.Health = append([]ProviderHealth(nil), input.Health...)
		for i := range out.Health {
			out.Health[i].RequiresEnv = append([]string(nil), out.Health[i].RequiresEnv...)
		}
	}
	if input.Fallback != nil {
		fallback := *input.Fallback
		out.Fallback = &fallback
	}
	return out
}

func FetchProviderCapability(tool string, provider string) ProviderCapability {
	tool = strings.TrimSpace(tool)
	provider = normalizeProviderOptional(provider)
	capability := ProviderCapability{Provider: provider}
	switch provider {
	case ProviderDirectHTTP:
		capability.Kind = "direct_http"
		switch tool {
		case "open_page", "web_fetch":
			capability.ContentExtraction = true
			capability.PDFBinaryHandoff = true
			capability.Cache = true
		case "http_request":
			capability.Cache = false
		default:
			capability.ContentExtraction = true
			capability.PDFBinaryHandoff = true
		}
	case ProviderFirecrawl:
		capability.Kind = "external_extractor"
		capability.ContentExtraction = true
		capability.Cache = true
	default:
		capability.Kind = ""
	}
	return capability
}

func FetchProviderCapabilities(tool string, providers []string) []ProviderCapability {
	if len(providers) == 0 {
		return nil
	}
	out := make([]ProviderCapability, 0, len(providers))
	seen := map[string]bool{}
	for _, raw := range providers {
		provider := normalizeProviderOptional(raw)
		if provider == "" || seen[provider] {
			continue
		}
		seen[provider] = true
		out = append(out, FetchProviderCapability(tool, provider))
	}
	return NormalizeProviderCapabilities(out)
}

func FetchProviderForExtractor(extractor string) string {
	switch strings.ToLower(strings.TrimSpace(extractor)) {
	case ProviderFirecrawl, "firecrawl_markdown", "firecrawl_text":
		return ProviderFirecrawl
	default:
		return ProviderDirectHTTP
	}
}

func availableProviderHealth(provider string) ProviderHealth {
	return ProviderHealth{
		Provider:   normalizeProviderOptional(provider),
		Available:  true,
		Configured: true,
		Status:     ProviderHealthAvailable,
	}
}

func NormalizeProviderHealthStatus(raw string) string {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	switch normalized {
	case ProviderHealthAvailable,
		ProviderHealthCredentialMissing,
		ProviderHealthCredentialInvalid,
		ProviderHealthQuotaLimited,
		ProviderHealthRateLimited,
		ProviderHealthNetworkBlocked,
		ProviderHealthUnsupportedFeature,
		ProviderHealthStatusError,
		ProviderHealthDecodeError,
		ProviderHealthRequestFailed,
		ProviderHealthTimeout:
		return normalized
	case "missing_credentials":
		return ProviderHealthCredentialMissing
	case "invalid_credentials", "invalid_credential", "invalid_api_key", "invalid_token", "subscription_token_invalid", "auth_invalid", "authentication_failed", "unauthorized", "forbidden":
		return ProviderHealthCredentialInvalid
	default:
		if normalized == "" {
			return ""
		}
		return ProviderHealthUnknown
	}
}

func normalizeProviderOptional(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	return NormalizeSearchProvider(trimmed)
}

func SearchProviderCapability(provider string) ProviderCapability {
	provider = normalizeProviderOptional(provider)
	capability := ProviderCapability{
		Provider:   provider,
		Kind:       searchProviderKindString(ResolveSearchProviderKind(provider)),
		Cache:      true,
		MaxResults: 10,
	}
	switch provider {
	case "brave":
		capability.DateFilters = true
		capability.Locale = true
		capability.StructuredSnippets = true
	case SearchProviderDoubaoCustom:
		capability.Cache = false
		capability.DateFilters = true
		capability.DomainFilters = true
		capability.AuthorityFilter = true
		capability.QueryRewrite = true
		capability.StructuredSnippets = true
		capability.MaxResults = 50
	case "baidu":
		capability.DateFilters = true
		capability.StructuredSnippets = true
	case "perplexity":
		capability.DateFilters = true
		capability.DomainFilters = true
		capability.StructuredSnippets = true
		capability.SynthesizedCitations = true
	case "openrouter":
		capability.DateFilters = true
		capability.SynthesizedCitations = true
	default:
		capability.Provider = provider
		capability.Kind = ""
		capability.Cache = false
		capability.MaxResults = 0
	}
	return capability
}

func SearchProviderCapabilities(providers []string) []ProviderCapability {
	if len(providers) == 0 {
		return nil
	}
	out := make([]ProviderCapability, 0, len(providers))
	seen := map[string]bool{}
	for _, raw := range providers {
		provider := normalizeProviderOptional(raw)
		if provider == "" || seen[provider] {
			continue
		}
		seen[provider] = true
		out = append(out, SearchProviderCapability(provider))
	}
	return NormalizeProviderCapabilities(out)
}

func NormalizeProviderCapabilities(items []ProviderCapability) []ProviderCapability {
	if len(items) == 0 {
		return nil
	}
	seen := map[string]bool{}
	out := make([]ProviderCapability, 0, len(items))
	for _, item := range items {
		item.Provider = normalizeProviderOptional(item.Provider)
		if item.Provider == "" || seen[item.Provider] {
			continue
		}
		seen[item.Provider] = true
		item.Kind = strings.TrimSpace(item.Kind)
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Provider < out[j].Provider
	})
	if len(out) == 0 {
		return nil
	}
	return out
}

func NormalizeProviderHealth(items []ProviderHealth) []ProviderHealth {
	if len(items) == 0 {
		return nil
	}
	seen := map[string]bool{}
	out := make([]ProviderHealth, 0, len(items))
	for _, item := range items {
		item.Provider = normalizeProviderOptional(item.Provider)
		if item.Provider == "" || seen[item.Provider] {
			continue
		}
		seen[item.Provider] = true
		item.Status = NormalizeProviderHealthStatus(item.Status)
		if item.Status == "" {
			if item.Available {
				item.Status = ProviderHealthAvailable
			} else {
				item.Status = ProviderHealthUnknown
			}
		}
		item.Reason = strings.TrimSpace(item.Reason)
		item.RequiresEnv = normalizeProviderStringList(item.RequiresEnv)
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Provider < out[j].Provider
	})
	if len(out) == 0 {
		return nil
	}
	return out
}

func normalizeProviderStringList(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		normalized := strings.TrimSpace(value)
		if normalized == "" || seen[normalized] {
			continue
		}
		seen[normalized] = true
		out = append(out, normalized)
	}
	sort.Strings(out)
	if len(out) == 0 {
		return nil
	}
	return out
}
