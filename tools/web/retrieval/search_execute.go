package retrieval

import (
	"context"
	"time"
)

type SearchExecuteOptions struct {
	Prepared             PreparedSearchRequest
	Endpoint             string
	ProviderAPIKey       string
	PerplexityAPIKey     string
	PerplexityBaseURL    string
	PerplexityModel      string
	TimeoutMs            int
	CacheTTL             time.Duration
	TrustedEnvProxy      bool
	Prepare              Preparer
	ClassifyNetworkError NetworkErrorClassifier
	Audit                func(SearchAuditEvent)
}

// SearchAuditEvent is the policy-neutral observation emitted by search.
// Persistence, redaction and enablement remain Host responsibilities.
type SearchAuditEvent struct {
	Prepared    PreparedSearchRequest
	Provider    string
	Event       string
	CacheHit    bool
	ResultCount int
	StartedAt   time.Time
	Err         error
}

func ExecutePreparedSearch(ctx context.Context, opts SearchExecuteOptions) (SearchPayload, error) {
	started := time.Now()
	provider := NormalizeSearchProvider(opts.Prepared.EffectiveProvider)
	if provider == "" {
		provider = DefaultSearchProvider
	}
	cacheKey := SearchCacheKey(
		provider,
		opts.Prepared.Query,
		opts.Prepared.Count,
		opts.Prepared.Country,
		opts.Prepared.SearchLang,
		opts.Prepared.UILang,
		opts.Prepared.Freshness,
		opts.Prepared.DateAfter,
		opts.Prepared.DateBefore,
		opts.Prepared.DomainFilter,
		opts.PerplexityModel,
		opts.TrustedEnvProxy,
	)
	if opts.CacheTTL > 0 {
		if cached, ok := ReadSearchCache(cacheKey); ok {
			payload := applyPreparedSearchMetadata(cached, opts.Prepared, true)
			emitSearchAudit(opts, provider, "search_provider_cache_hit", true, len(payload.Results), started, nil)
			return payload, nil
		}
	}
	payload, err := RunSearch(ctx, SearchRunOptions{
		Provider:               provider,
		Query:                  opts.Prepared.Query,
		Count:                  opts.Prepared.Count,
		Country:                opts.Prepared.Country,
		SearchLang:             opts.Prepared.SearchLang,
		UILang:                 opts.Prepared.UILang,
		Freshness:              opts.Prepared.Freshness,
		TimeoutMs:              opts.TimeoutMs,
		BraveEndpoint:          opts.Endpoint,
		BraveAPIKey:            opts.ProviderAPIKey,
		ArkEndpoint:            opts.Endpoint,
		ArkAPIKey:              opts.ProviderAPIKey,
		ArkTimeRange:           opts.Prepared.ArkTimeRange,
		BaiduEndpoint:          opts.Endpoint,
		BaiduAPIKey:            opts.ProviderAPIKey,
		BaiduRecency:           opts.Prepared.BaiduRecency,
		PerplexityAPIKey:       opts.PerplexityAPIKey,
		PerplexityBaseURL:      opts.PerplexityBaseURL,
		PerplexityModel:        opts.PerplexityModel,
		PerplexityRecency:      opts.Prepared.PerplexityRecency,
		PerplexityDateAfter:    opts.Prepared.PerplexityDateAfter,
		PerplexityDateBefore:   opts.Prepared.PerplexityDateBefore,
		PerplexityDomainFilter: append([]string(nil), opts.Prepared.DomainFilter...),
		Prepare:                opts.Prepare,
		ClassifyNetworkError:   opts.ClassifyNetworkError,
	})
	if err != nil {
		emitSearchAudit(opts, provider, "search_provider_request", false, 0, started, err)
		return SearchPayload{}, err
	}
	payload = applyPreparedSearchMetadata(payload, opts.Prepared, false)
	emitSearchAudit(opts, provider, "search_provider_request", false, len(payload.Results), started, nil)
	if opts.CacheTTL > 0 {
		WriteSearchCache(cacheKey, payload, opts.CacheTTL)
	}
	return payload, nil
}

func emitSearchAudit(opts SearchExecuteOptions, provider, event string, cacheHit bool, resultCount int, started time.Time, err error) {
	if opts.Audit == nil {
		return
	}
	opts.Audit(SearchAuditEvent{Prepared: opts.Prepared, Provider: provider, Event: event, CacheHit: cacheHit, ResultCount: resultCount, StartedAt: started, Err: err})
}

func applyPreparedSearchMetadata(payload SearchPayload, prepared PreparedSearchRequest, cached bool) SearchPayload {
	out := cloneSearchPayload(payload)
	if stringsTrimmed := prepared.RequestedProvider; stringsTrimmed != "" {
		out.RequestedProvider = stringsTrimmed
	}
	if stringsTrimmed := prepared.ProviderNote; stringsTrimmed != "" {
		out.ProviderNote = stringsTrimmed
	}
	out.Cached = cached
	return out
}
