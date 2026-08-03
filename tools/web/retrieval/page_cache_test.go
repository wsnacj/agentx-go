package retrieval

import (
	"testing"
	"time"
)

func TestOpenPageCacheRoundTripAndLookupByID(t *testing.T) {
	ResetOpenPageCache()
	key := OpenPageCacheKey("https://example.com/report", 4000, 2000000, "sig=v1")
	pageID := OpenPageIDForCacheKey(key)
	page := Page{
		PageID:     pageID,
		RequestURL: "https://example.com/report",
		FinalURL:   "https://example.com/report",
		Title:      "Quarterly Report",
		Text:       "hello world",
		Links: []PageLink{
			{Text: "Investor Relations", URL: "https://example.com/ir"},
		},
	}
	WriteOpenPageCache(key, page, time.Minute)

	byKey, ok := ReadOpenPageCache(key)
	if !ok {
		t.Fatal("expected page cache hit by key")
	}
	if byKey.PageID != pageID || byKey.Title != "Quarterly Report" {
		t.Fatalf("unexpected page by key: %#v", byKey)
	}
	byID, ok := ReadOpenPageCacheByID(pageID)
	if !ok {
		t.Fatal("expected page cache hit by id")
	}
	if byID.FinalURL != "https://example.com/report" || len(byID.Links) != 1 {
		t.Fatalf("unexpected page by id: %#v", byID)
	}
	byURL, ok := ReadOpenPageCacheByURL("https://example.com/report")
	if !ok {
		t.Fatal("expected page cache hit by url")
	}
	if byURL.PageID != pageID || byURL.Title != "Quarterly Report" {
		t.Fatalf("unexpected page by url: %#v", byURL)
	}
}

func TestOpenPageCacheExpires(t *testing.T) {
	ResetOpenPageCache()
	key := OpenPageCacheKey("https://example.com/report", 4000, 2000000, "sig=v1")
	page := Page{
		PageID: OpenPageIDForCacheKey(key),
		Title:  "Expiring Report",
		Text:   "hello",
	}
	WriteOpenPageCache(key, page, 5*time.Millisecond)
	time.Sleep(25 * time.Millisecond)
	if _, ok := ReadOpenPageCache(key); ok {
		t.Fatal("expected page cache entry to expire")
	}
	if _, ok := ReadOpenPageCacheByID(page.PageID); ok {
		t.Fatal("expected page cache id entry to expire")
	}
	if _, ok := ReadOpenPageCacheByURL("https://example.com/report"); ok {
		t.Fatal("expected page cache url entry to expire")
	}
}

func TestOpenPageCacheReturnsClonedPages(t *testing.T) {
	ResetOpenPageCache()
	key := OpenPageCacheKey("https://example.com/report", 4000, 2000000, "sig=v1")
	page := Page{
		PageID: "page_clone",
		Title:  "Cloned Report",
		Links: []PageLink{
			{Text: "IR", URL: "https://example.com/ir"},
		},
		Diagnostics: &PageDiagnostics{
			PageTextOK:        true,
			SuggestedNextTool: "find_in_page",
			Warnings:          []string{"redirected"},
		},
	}
	WriteOpenPageCache(key, page, time.Minute)

	first, ok := ReadOpenPageCache(key)
	if !ok {
		t.Fatal("expected first page cache hit")
	}
	first.Title = "mutated"
	first.Links[0].Text = "mutated"
	first.Diagnostics.PageTextOK = false
	first.Diagnostics.Warnings[0] = "mutated"

	second, ok := ReadOpenPageCache(key)
	if !ok {
		t.Fatal("expected second page cache hit")
	}
	if second.Title != "Cloned Report" || second.Links[0].Text != "IR" {
		t.Fatalf("expected cached page to remain immutable across reads, got %#v", second)
	}
	if second.Diagnostics == nil || !second.Diagnostics.PageTextOK || second.Diagnostics.Warnings[0] != "redirected" {
		t.Fatalf("expected cached page diagnostics to remain immutable across reads, got %#v", second.Diagnostics)
	}
}
