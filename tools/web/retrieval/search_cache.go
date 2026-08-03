package retrieval

import (
	"strconv"
	"strings"
	"sync"
	"time"
)

type SearchCacheEntry struct {
	ExpiresAt time.Time
	Payload   SearchPayload
}

var (
	searchCacheMu sync.Mutex
	searchCache   = map[string]SearchCacheEntry{}
)

func SearchCacheKey(provider string, query string, count int, country string, searchLang string, uiLang string, freshness string, dateAfter string, dateBefore string, domainFilter []string, model string, trustedEnvProxy bool) string {
	return strings.ToLower(strings.TrimSpace(provider)) +
		"\n" + strings.ToLower(strings.TrimSpace(query)) +
		"\n" + strconv.Itoa(count) +
		"\n" + strings.ToLower(strings.TrimSpace(country)) +
		"\n" + strings.ToLower(strings.TrimSpace(searchLang)) +
		"\n" + strings.ToLower(strings.TrimSpace(uiLang)) +
		"\n" + strings.ToLower(strings.TrimSpace(freshness)) +
		"\n" + strings.ToLower(strings.TrimSpace(dateAfter)) +
		"\n" + strings.ToLower(strings.TrimSpace(dateBefore)) +
		"\n" + strings.ToLower(strings.Join(domainFilter, ",")) +
		"\n" + strconv.FormatBool(trustedEnvProxy) +
		"\n" + strings.ToLower(strings.TrimSpace(model))
}

func ReadSearchCache(key string) (SearchPayload, bool) {
	searchCacheMu.Lock()
	defer searchCacheMu.Unlock()
	entry, ok := searchCache[key]
	if !ok {
		return SearchPayload{}, false
	}
	if time.Now().After(entry.ExpiresAt) {
		delete(searchCache, key)
		return SearchPayload{}, false
	}
	return cloneSearchPayload(entry.Payload), true
}

func WriteSearchCache(key string, payload SearchPayload, ttl time.Duration) {
	if ttl <= 0 {
		return
	}
	searchCacheMu.Lock()
	defer searchCacheMu.Unlock()
	if len(searchCache) > 512 {
		now := time.Now()
		for cacheKey, item := range searchCache {
			if now.After(item.ExpiresAt) {
				delete(searchCache, cacheKey)
			}
		}
	}
	searchCache[key] = SearchCacheEntry{
		ExpiresAt: time.Now().Add(ttl),
		Payload:   cloneSearchPayload(payload),
	}
}

func ResetSearchCache() {
	searchCacheMu.Lock()
	defer searchCacheMu.Unlock()
	searchCache = map[string]SearchCacheEntry{}
}

func cloneSearchPayload(payload SearchPayload) SearchPayload {
	out := payload
	if len(payload.Results) > 0 {
		out.Results = append([]SearchResult(nil), payload.Results...)
	}
	if payload.ProviderDiagnostics != nil {
		cloned := CloneProviderDiagnostics(*payload.ProviderDiagnostics)
		out.ProviderDiagnostics = &cloned
	}
	return out
}
