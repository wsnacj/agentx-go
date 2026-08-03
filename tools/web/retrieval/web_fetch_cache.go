package retrieval

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type webFetchCacheEntry struct {
	ExpiresAt time.Time
	Payload   []byte
}

var (
	webFetchCacheMu sync.Mutex
	webFetchCache   = map[string]webFetchCacheEntry{}
)

func WebFetchCacheKey(urlValue string, maxChars int, maxResponseBytes int, extractMode string, requestSignature string) string {
	mode := strings.TrimSpace(extractMode)
	material := strings.TrimSpace(urlValue) + "\n" + fmt.Sprintf("%d", maxChars) + "\n" + fmt.Sprintf("%d", maxResponseBytes) + "\n" + mode + "\n" + strings.TrimSpace(requestSignature)
	return "web_fetch_" + requestIdentitySHA256(material)
}

func ReadWebFetchCache(key string) ([]byte, bool) {
	webFetchCacheMu.Lock()
	defer webFetchCacheMu.Unlock()
	entry, ok := webFetchCache[key]
	if !ok {
		return nil, false
	}
	if time.Now().After(entry.ExpiresAt) {
		delete(webFetchCache, key)
		return nil, false
	}
	return append([]byte(nil), entry.Payload...), true
}

func WriteWebFetchCache(key string, payload []byte, ttl time.Duration) {
	if ttl <= 0 {
		return
	}
	now := time.Now()
	webFetchCacheMu.Lock()
	if len(webFetchCache) > 512 {
		for name, entry := range webFetchCache {
			if now.After(entry.ExpiresAt) {
				delete(webFetchCache, name)
			}
		}
	}
	webFetchCache[key] = webFetchCacheEntry{
		ExpiresAt: now.Add(ttl),
		Payload:   append([]byte(nil), payload...),
	}
	webFetchCacheMu.Unlock()
}

func ResetWebFetchCache() {
	webFetchCacheMu.Lock()
	webFetchCache = map[string]webFetchCacheEntry{}
	webFetchCacheMu.Unlock()
}
