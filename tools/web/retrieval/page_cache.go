package retrieval

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"
)

type PageLink struct {
	Text string `json:"text,omitempty"`
	URL  string `json:"url"`
}

type Page struct {
	PageID              string               `json:"page_id"`
	RequestURL          string               `json:"request_url"`
	FinalURL            string               `json:"final_url"`
	Status              int                  `json:"status"`
	ContentType         string               `json:"content_type,omitempty"`
	SiteName            string               `json:"site_name,omitempty"`
	Title               string               `json:"title,omitempty"`
	Byline              string               `json:"byline,omitempty"`
	Excerpt             string               `json:"excerpt,omitempty"`
	Text                string               `json:"text"`
	Markdown            string               `json:"markdown,omitempty"`
	Links               []PageLink           `json:"links,omitempty"`
	Extractor           string               `json:"extractor,omitempty"`
	FallbackUsed        bool                 `json:"fallback_used,omitempty"`
	ReadabilityUsed     bool                 `json:"readability_used,omitempty"`
	Truncated           bool                 `json:"truncated,omitempty"`
	Warning             string               `json:"warning,omitempty"`
	WordCount           int                  `json:"word_count,omitempty"`
	FetchedAt           int64                `json:"fetched_at"`
	TookMs              int64                `json:"took_ms,omitempty"`
	Diagnostics         *PageDiagnostics     `json:"diagnostics,omitempty"`
	ProviderDiagnostics *ProviderDiagnostics `json:"provider_diagnostics,omitempty"`
}

type PageDiagnostics struct {
	PageTextOK        bool     `json:"page_text_ok"`
	Issue             string   `json:"issue,omitempty"`
	ExtractionMethod  string   `json:"extraction_method,omitempty"`
	ExtractMode       string   `json:"extract_mode,omitempty"`
	Redirected        bool     `json:"redirected,omitempty"`
	RedirectCount     int      `json:"redirect_count,omitempty"`
	CacheHit          bool     `json:"cache_hit,omitempty"`
	BodyTruncated     bool     `json:"body_truncated,omitempty"`
	ContentTruncated  bool     `json:"content_truncated,omitempty"`
	ResponseBytes     int      `json:"response_bytes,omitempty"`
	SuggestedNextTool string   `json:"suggested_next_tool,omitempty"`
	Warnings          []string `json:"warnings,omitempty"`
}

type pageCacheEntry struct {
	ExpiresAt time.Time
	Page      Page
}

var (
	pageCacheMu    sync.Mutex
	pageCacheByKey = map[string]pageCacheEntry{}
	pageCacheByID  = map[string]pageCacheEntry{}
)

func OpenPageCacheKey(urlValue string, maxChars int, maxResponseBytes int, requestSignature string) string {
	material := strings.TrimSpace(urlValue) + "\n" +
		fmt.Sprintf("%d", maxChars) + "\n" +
		fmt.Sprintf("%d", maxResponseBytes) + "\n" +
		strings.TrimSpace(requestSignature)
	return "open_page_" + requestIdentitySHA256(material)
}

func OpenPageIDForCacheKey(key string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(key)))
	return "page_" + hex.EncodeToString(sum[:8])
}

func ReadOpenPageCache(key string) (Page, bool) {
	pageCacheMu.Lock()
	defer pageCacheMu.Unlock()
	entry, ok := pageCacheByKey[key]
	if !ok {
		return Page{}, false
	}
	if time.Now().After(entry.ExpiresAt) {
		delete(pageCacheByKey, key)
		if strings.TrimSpace(entry.Page.PageID) != "" {
			delete(pageCacheByID, entry.Page.PageID)
		}
		return Page{}, false
	}
	return cloneCachedPage(entry.Page), true
}

func ReadOpenPageCacheByID(pageID string) (Page, bool) {
	pageCacheMu.Lock()
	defer pageCacheMu.Unlock()
	entry, ok := pageCacheByID[strings.TrimSpace(pageID)]
	if !ok {
		return Page{}, false
	}
	if time.Now().After(entry.ExpiresAt) {
		delete(pageCacheByID, strings.TrimSpace(pageID))
		return Page{}, false
	}
	return cloneCachedPage(entry.Page), true
}

func ReadOpenPageCacheByURL(urlValue string) (Page, bool) {
	urlValue = evidenceURL(urlValue)
	if urlValue == "" {
		return Page{}, false
	}
	now := time.Now()
	pageCacheMu.Lock()
	defer pageCacheMu.Unlock()
	for key, entry := range pageCacheByKey {
		if now.After(entry.ExpiresAt) {
			delete(pageCacheByKey, key)
			if strings.TrimSpace(entry.Page.PageID) != "" {
				delete(pageCacheByID, entry.Page.PageID)
			}
			continue
		}
		if strings.TrimSpace(entry.Page.RequestURL) == urlValue ||
			strings.TrimSpace(entry.Page.FinalURL) == urlValue {
			return cloneCachedPage(entry.Page), true
		}
	}
	return Page{}, false
}

func WriteOpenPageCache(key string, page Page, ttl time.Duration) {
	if ttl <= 0 {
		return
	}
	now := time.Now()
	pageCacheMu.Lock()
	defer pageCacheMu.Unlock()
	if len(pageCacheByKey) > 512 {
		for name, entry := range pageCacheByKey {
			if now.After(entry.ExpiresAt) {
				delete(pageCacheByKey, name)
				if strings.TrimSpace(entry.Page.PageID) != "" {
					delete(pageCacheByID, entry.Page.PageID)
				}
			}
		}
	}
	entry := pageCacheEntry{
		ExpiresAt: now.Add(ttl),
		Page:      cloneCachedPage(page),
	}
	pageCacheByKey[key] = entry
	if strings.TrimSpace(page.PageID) != "" {
		pageCacheByID[strings.TrimSpace(page.PageID)] = entry
	}
}

func ResetOpenPageCache() {
	pageCacheMu.Lock()
	defer pageCacheMu.Unlock()
	pageCacheByKey = map[string]pageCacheEntry{}
	pageCacheByID = map[string]pageCacheEntry{}
}

func cloneCachedPage(page Page) Page {
	cloned := page
	if len(page.Links) > 0 {
		cloned.Links = append([]PageLink(nil), page.Links...)
	}
	if page.Diagnostics != nil {
		diagnostics := *page.Diagnostics
		if len(page.Diagnostics.Warnings) > 0 {
			diagnostics.Warnings = append([]string(nil), page.Diagnostics.Warnings...)
		}
		cloned.Diagnostics = &diagnostics
	}
	if page.ProviderDiagnostics != nil {
		providerDiagnostics := CloneProviderDiagnostics(*page.ProviderDiagnostics)
		cloned.ProviderDiagnostics = &providerDiagnostics
	}
	return cloned
}
