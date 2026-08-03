// Package web provides portable Web Search, Fetch, OpenPage and FindInPage tools.
//
// The package owns provider protocols, extraction, caching and model-facing
// coordination. A Host must inject a retrieval.Preparer and remains the sole
// owner of outbound policy, proxy, redirect validation, credentials and audit
// persistence.
package web

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	toolcontract "github.com/wsnacj/agentx-go/components/tool"
	agentxtoolerrors "github.com/wsnacj/agentx-go/runtime/toolerrors"
	"github.com/wsnacj/agentx-go/tools/web/retrieval"
)

const (
	SearchName     = "search"
	WebSearchName  = "web_search"
	WebFetchName   = "web_fetch"
	OpenPageName   = "open_page"
	FindInPageName = "find_in_page"
)

// ProviderConfig explicitly supplies one search provider's credentials and
// endpoint. This package never discovers credentials from the environment.
type ProviderConfig struct {
	Provider string
	Endpoint string
	APIKey   string
	Model    string
}

// SearchOptions configures the portable provider protocol implementation.
type SearchOptions struct {
	DefaultProvider      string
	Providers            map[string]ProviderConfig
	Prepare              retrieval.Preparer
	ClassifyNetworkError retrieval.NetworkErrorClassifier
	Audit                func(retrieval.SearchAuditEvent)
	TimeoutMs            int
	CacheTTL             time.Duration
	TrustedProxyVariant  bool
}

// FetchOptions configures direct page retrieval and optional Firecrawl fallback.
type FetchOptions struct {
	Prepare              retrieval.Preparer
	ClassifyNetworkError retrieval.NetworkErrorClassifier
	TimeoutMs            int
	MaxChars             int
	MaxResponseBytes     int
	MaxRedirects         int
	CacheTTL             time.Duration
	UserAgent            string
	RequestSignature     string
	Firecrawl            retrieval.WebFetchFirecrawlConfig
}

// Options configures the complete portable web tool cohort.
type Options struct {
	Search SearchOptions
	Fetch  FetchOptions
}

// SearchRequest is the model- and Go-facing portable search input.
type SearchRequest struct {
	Query        string   `json:"query"`
	Provider     string   `json:"provider,omitempty"`
	MaxResults   int      `json:"max_results,omitempty"`
	Count        int      `json:"count,omitempty"`
	Country      string   `json:"country,omitempty"`
	SearchLang   string   `json:"search_lang,omitempty"`
	UILang       string   `json:"ui_lang,omitempty"`
	Freshness    string   `json:"freshness,omitempty"`
	DateAfter    string   `json:"date_after,omitempty"`
	DateBefore   string   `json:"date_before,omitempty"`
	DomainFilter []string `json:"domain_filter,omitempty"`
	TimeoutMs    int      `json:"timeout_ms,omitempty"`
}

// FetchRequest is the Go-facing input shared by WebFetch and OpenPage.
type FetchRequest struct {
	URL              string `json:"url"`
	MaxChars         int    `json:"max_chars,omitempty"`
	MaxResponseBytes int    `json:"max_response_bytes,omitempty"`
	TimeoutMs        int    `json:"timeout_ms,omitempty"`
	ExtractMode      string `json:"extract_mode,omitempty"`
}

// FindRequest searches a page previously cached by OpenPage.
type FindRequest struct {
	PageID       string `json:"page_id"`
	Query        string `json:"query"`
	MaxMatches   int    `json:"max_matches,omitempty"`
	ContextChars int    `json:"context_chars,omitempty"`
}

// Register adds all web tools whose required Host ports are configured.
func Register(reg toolcontract.Registrar, opts Options) {
	RegisterSearch(reg, opts.Search)
	RegisterFetch(reg, opts.Fetch)
	RegisterOpenPage(reg, opts.Fetch)
	RegisterFindInPage(reg)
}

// RegisterSearch adds search and web_search aliases.
func RegisterSearch(reg toolcontract.Registrar, opts SearchOptions) {
	if reg == nil || opts.Prepare == nil {
		return
	}
	reg.Register(SearchDefinition(), NewSearchHandler(SearchName, opts))
	reg.Register(WebSearchDefinition(), NewSearchHandler(WebSearchName, opts))
}

// RegisterFetch adds the low-level web_fetch tool.
func RegisterFetch(reg toolcontract.Registrar, opts FetchOptions) {
	if reg == nil || opts.Prepare == nil {
		return
	}
	reg.Register(WebFetchDefinition(), NewFetchHandler(opts))
}

// RegisterOpenPage adds the readable-page tool.
func RegisterOpenPage(reg toolcontract.Registrar, opts FetchOptions) {
	if reg == nil || opts.Prepare == nil {
		return
	}
	reg.Register(OpenPageDefinition(), NewOpenPageHandler(opts))
}

// RegisterFindInPage adds the cache-only page search tool.
func RegisterFindInPage(reg toolcontract.Registrar) {
	if reg != nil {
		reg.Register(FindInPageDefinition(), NewFindInPageHandler())
	}
}

// RunSearch executes a configured provider without environment discovery or
// implicit product fallback policy.
func RunSearch(ctx context.Context, request SearchRequest, opts SearchOptions) (retrieval.SearchPayload, error) {
	query := strings.TrimSpace(request.Query)
	if query == "" {
		return retrieval.SearchPayload{}, agentxtoolerrors.NewMissingRequiredToolArgumentError(SearchName, []string{"query"}, "search: query is required")
	}
	provider := retrieval.NormalizeSearchProvider(request.Provider)
	if strings.TrimSpace(request.Provider) == "" {
		provider = retrieval.NormalizeSearchProvider(opts.DefaultProvider)
	}
	config, ok := lookupProvider(opts.Providers, provider)
	canUse := func(candidate string) bool {
		candidateConfig, exists := lookupProvider(opts.Providers, candidate)
		return exists && strings.TrimSpace(candidateConfig.APIKey) != ""
	}
	count := request.MaxResults
	if count <= 0 {
		count = request.Count
	}
	if count <= 0 {
		count = 5
	}
	if count > 10 {
		count = 10
	}
	prepared, validation := retrieval.PrepareSearchRequest(retrieval.SearchPrepareOptions{
		ConfiguredProvider: provider, RequestedProvider: request.Provider, Query: query, Count: count,
		Country: request.Country, SearchLang: request.SearchLang, UILang: request.UILang,
		Freshness: request.Freshness, DateAfter: request.DateAfter, DateBefore: request.DateBefore,
		DomainFilter: request.DomainFilter, CanUseProvider: canUse,
		CanUsePerplexityStructuredDateFilters: canUse("perplexity"),
	})
	if validation != nil {
		return retrieval.SearchPayload{}, fmt.Errorf("search: %s: %s", validation.Code, validation.Message)
	}
	if prepared.EffectiveProvider != provider {
		config, ok = lookupProvider(opts.Providers, prepared.EffectiveProvider)
	}
	if !ok || strings.TrimSpace(config.APIKey) == "" {
		return retrieval.SearchPayload{}, fmt.Errorf("search: provider %s credentials are unavailable", prepared.EffectiveProvider)
	}
	endpoint := strings.TrimSpace(config.Endpoint)
	if endpoint == "" {
		switch prepared.EffectiveProvider {
		case "perplexity":
			endpoint = retrieval.DefaultSearchPerplexityURL
		case "openrouter":
			endpoint = retrieval.DefaultSearchOpenRouterURL
		default:
			endpoint = retrieval.DefaultSearchEndpointForProvider(prepared.EffectiveProvider)
		}
	}
	model := strings.TrimSpace(config.Model)
	if model == "" && (prepared.EffectiveProvider == "perplexity" || prepared.EffectiveProvider == "openrouter") {
		model = "sonar"
	}
	timeoutMs := request.TimeoutMs
	if timeoutMs <= 0 {
		timeoutMs = opts.TimeoutMs
	}
	if timeoutMs <= 0 {
		timeoutMs = 15_000
	}
	return retrieval.ExecutePreparedSearch(ctx, retrieval.SearchExecuteOptions{
		Prepared: prepared, Endpoint: endpoint, ProviderAPIKey: config.APIKey,
		PerplexityAPIKey: config.APIKey, PerplexityBaseURL: endpoint, PerplexityModel: model,
		TimeoutMs: timeoutMs, CacheTTL: opts.CacheTTL, TrustedEnvProxy: opts.TrustedProxyVariant,
		Prepare: opts.Prepare, ClassifyNetworkError: opts.ClassifyNetworkError, Audit: opts.Audit,
	})
}

// RunFetch executes direct retrieval and extraction.
func RunFetch(ctx context.Context, request FetchRequest, opts FetchOptions) (retrieval.WebFetchExecutionResult, error) {
	return runFetch(ctx, WebFetchName, request, opts, true, "auto")
}

// RunOpenPage fetches a readable page and stores it for FindInPage.
func RunOpenPage(ctx context.Context, request FetchRequest, opts FetchOptions) (retrieval.Page, bool, error) {
	if strings.TrimSpace(request.URL) == "" {
		return retrieval.Page{}, false, agentxtoolerrors.NewMissingRequiredToolArgumentError(OpenPageName, []string{"url"}, "open_page: url is required")
	}
	maxChars, maxBytes := effectiveFetchLimits(request, opts)
	cacheKey := retrieval.OpenPageCacheKey(request.URL, maxChars, maxBytes, opts.RequestSignature)
	if opts.CacheTTL > 0 {
		if cached, ok := retrieval.ReadOpenPageCache(cacheKey); ok {
			return cached, true, nil
		}
	}
	result, err := runFetch(ctx, OpenPageName, request, opts, false, "markdown")
	if err != nil {
		return retrieval.Page{}, false, err
	}
	payload := result.Payload
	markdown := ""
	text := strings.TrimSpace(payload.Content)
	if strings.Contains(strings.TrimSpace(payload.ExtractMode), "markdown") {
		markdown = text
		text = strings.TrimSpace(retrieval.MarkdownToPlainText(markdown))
	}
	page := retrieval.Page{
		PageID: retrieval.OpenPageIDForCacheKey(cacheKey), RequestURL: payload.URL, FinalURL: payload.FinalURL,
		Status: payload.Status, ContentType: payload.ContentType, SiteName: payload.SiteName, Title: payload.Title,
		Byline: payload.Byline, Excerpt: payload.Excerpt, Text: text, Markdown: markdown,
		Links: retrieval.DedupePageLinks(result.Links), Extractor: payload.Extractor, FallbackUsed: payload.FallbackUsed,
		ReadabilityUsed: payload.ReadabilityUsed, Truncated: payload.Truncated, Warning: payload.Warning,
		WordCount: payload.WordCount, FetchedAt: payload.FetchedAt, TookMs: payload.TookMs,
	}
	if payload.ProviderDiagnostics != nil {
		cloned := retrieval.CloneProviderDiagnostics(*payload.ProviderDiagnostics)
		page.ProviderDiagnostics = &cloned
	}
	page.Warning = retrieval.InferOpenPageWarning(page)
	page.Diagnostics = retrieval.BuildOpenPageDiagnostics(page, retrieval.PageDiagnosticsInput{
		ExtractMode: payload.ExtractMode, Redirected: payload.Redirected, RedirectCount: payload.RedirectCount,
		BodyTruncated: payload.BodyTruncated, ContentTruncated: payload.Truncated, ResponseBytes: payload.ResponseBytes,
	})
	if opts.CacheTTL > 0 {
		retrieval.WriteOpenPageCache(cacheKey, page, opts.CacheTTL)
	}
	return page, false, nil
}

// RunFindInPage searches a cached page without network access.
func RunFindInPage(request FindRequest) retrieval.FindInPageResult {
	page, ok := retrieval.ReadOpenPageCacheByID(strings.TrimSpace(request.PageID))
	if !ok {
		return retrieval.FindInPageCacheMiss(request.PageID, request.Query)
	}
	return retrieval.FindInPage(page, request.Query, request.MaxMatches, request.ContextChars)
}

func runFetch(ctx context.Context, toolName string, request FetchRequest, opts FetchOptions, useCache bool, defaultMode string) (retrieval.WebFetchExecutionResult, error) {
	if strings.TrimSpace(request.URL) == "" {
		return retrieval.WebFetchExecutionResult{}, agentxtoolerrors.NewMissingRequiredToolArgumentError(toolName, []string{"url"}, toolName+": url is required")
	}
	maxChars, maxBytes := effectiveFetchLimits(request, opts)
	timeoutMs := request.TimeoutMs
	if timeoutMs <= 0 {
		timeoutMs = opts.TimeoutMs
	}
	if timeoutMs <= 0 {
		timeoutMs = 30_000
	}
	maxRedirects := opts.MaxRedirects
	if maxRedirects <= 0 {
		maxRedirects = 5
	}
	userAgent := strings.TrimSpace(opts.UserAgent)
	if userAgent == "" {
		userAgent = "agentx-web/1.0"
	}
	result, err := retrieval.ExecuteWebFetch(ctx, retrieval.ExecuteWebFetchOptions{
		ToolName: toolName, RawURL: request.URL, DefaultExtractMode: defaultMode,
		RequestedExtractMode: request.ExtractMode, UseCache: useCache,
		RequestMaxChars: maxChars, RequestMaxResponseBytes: maxBytes, RequestTimeoutMs: timeoutMs,
		MaxRedirects: maxRedirects, UserAgent: userAgent, RequestSignature: opts.RequestSignature,
		CacheTTL: opts.CacheTTL, Firecrawl: opts.Firecrawl, Prepare: opts.Prepare,
		ClassifyNetworkError: opts.ClassifyNetworkError,
	})
	if err == nil && strings.TrimSpace(result.Payload.Content) != "" {
		result.Payload.ExternalContent = &retrieval.ExternalContentMeta{Untrusted: true, Source: toolName, Fields: []string{"content"}}
	}
	return result, err
}

func effectiveFetchLimits(request FetchRequest, opts FetchOptions) (int, int) {
	maxChars := request.MaxChars
	if maxChars <= 0 || (opts.MaxChars > 0 && maxChars > opts.MaxChars) {
		maxChars = opts.MaxChars
	}
	if maxChars <= 0 {
		maxChars = 50_000
	}
	maxBytes := request.MaxResponseBytes
	if maxBytes <= 0 || (opts.MaxResponseBytes > 0 && maxBytes > opts.MaxResponseBytes) {
		maxBytes = opts.MaxResponseBytes
	}
	if maxBytes <= 0 {
		maxBytes = 2_000_000
	}
	return maxChars, maxBytes
}

func lookupProvider(configs map[string]ProviderConfig, provider string) (ProviderConfig, bool) {
	provider = retrieval.NormalizeSearchProvider(provider)
	for key, config := range configs {
		name := retrieval.NormalizeSearchProvider(config.Provider)
		if strings.TrimSpace(config.Provider) == "" {
			name = retrieval.NormalizeSearchProvider(key)
		}
		if name == provider {
			config.Provider = name
			return config, true
		}
	}
	return ProviderConfig{}, false
}

// NewSearchHandler returns a model-facing search handler.
func NewSearchHandler(name string, opts SearchOptions) toolcontract.Handler {
	return func(ctx context.Context, call toolcontract.Call) (toolcontract.Result, error) {
		var request SearchRequest
		if err := decode(call.Arguments, &request, name); err != nil {
			return "", err
		}
		payload, err := RunSearch(ctx, request, opts)
		return marshal(payload, err)
	}
}

// NewFetchHandler returns a model-facing web_fetch handler.
func NewFetchHandler(opts FetchOptions) toolcontract.Handler {
	return func(ctx context.Context, call toolcontract.Call) (toolcontract.Result, error) {
		var request FetchRequest
		if err := decode(call.Arguments, &request, WebFetchName); err != nil {
			return "", err
		}
		result, err := RunFetch(ctx, request, opts)
		return marshal(result.Payload, err)
	}
}

// NewOpenPageHandler returns a model-facing open_page handler.
func NewOpenPageHandler(opts FetchOptions) toolcontract.Handler {
	return func(ctx context.Context, call toolcontract.Call) (toolcontract.Result, error) {
		var request FetchRequest
		if err := decode(call.Arguments, &request, OpenPageName); err != nil {
			return "", err
		}
		page, cached, err := RunOpenPage(ctx, request, opts)
		return marshal(struct {
			retrieval.Page
			FromCache bool `json:"from_cache,omitempty"`
		}{Page: page, FromCache: cached}, err)
	}
}

// NewFindInPageHandler returns a cache-only find_in_page handler.
func NewFindInPageHandler() toolcontract.Handler {
	return func(_ context.Context, call toolcontract.Call) (toolcontract.Result, error) {
		var request FindRequest
		if err := decode(call.Arguments, &request, FindInPageName); err != nil {
			return "", err
		}
		if strings.TrimSpace(request.PageID) == "" {
			return "", agentxtoolerrors.NewMissingRequiredToolArgumentError(FindInPageName, []string{"page_id"}, "find_in_page: page_id is required")
		}
		if strings.TrimSpace(request.Query) == "" {
			return "", agentxtoolerrors.NewMissingRequiredToolArgumentError(FindInPageName, []string{"query"}, "find_in_page: query is required")
		}
		return marshal(RunFindInPage(request), nil)
	}
}

func decode(raw string, target any, toolName string) error {
	if strings.TrimSpace(raw) == "" {
		raw = "{}"
	}
	if err := json.Unmarshal([]byte(raw), target); err != nil {
		return agentxtoolerrors.NewInvalidJSONToolArgumentError(toolName, err)
	}
	return nil
}

func marshal(value any, err error) (toolcontract.Result, error) {
	if err != nil {
		return "", err
	}
	blob, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(blob), nil
}
