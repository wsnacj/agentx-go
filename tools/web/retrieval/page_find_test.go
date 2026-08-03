package retrieval

import "testing"

func TestFindInPageUsesMarkdownHeadings(t *testing.T) {
	page := Page{
		PageID:   "page_abc",
		Title:    "Annual Report",
		FinalURL: "https://example.com/report",
		Markdown: "# Overview\n\nRevenue increased to 31.51 billion yuan.\n\n## Profitability\n\nNet profit reached 2.99 billion yuan.",
	}
	result := FindInPage(page, "net profit", 5, 40)
	if result.PageID != "page_abc" || result.Query != "net profit" {
		t.Fatalf("unexpected metadata: %#v", result)
	}
	if result.MatchCount != 1 || len(result.Matches) != 1 {
		t.Fatalf("expected one match, got %#v", result)
	}
	if result.Matches[0].Heading != "Profitability" {
		t.Fatalf("expected profitability heading, got %#v", result.Matches[0])
	}
	if result.Matches[0].Snippet == "" {
		t.Fatalf("expected snippet, got %#v", result.Matches[0])
	}
	if result.Diagnostics == nil ||
		result.Diagnostics.Scope != "page_cache" ||
		!result.Diagnostics.PageCacheHit ||
		result.Diagnostics.MatchStatus != "matched" {
		t.Fatalf("expected matched diagnostics, got %#v", result.Diagnostics)
	}
}

func TestFindInPageClampsMatchesAndReportsHasMore(t *testing.T) {
	page := Page{
		PageID: "page_many",
		Text:   "profit one\nprofit two\nprofit three",
	}
	result := FindInPage(page, "profit", 2, 20)
	if result.MatchCount != 3 {
		t.Fatalf("expected 3 total matches, got %#v", result)
	}
	if len(result.Matches) != 2 || !result.HasMore {
		t.Fatalf("expected truncated visible matches, got %#v", result)
	}
}

func TestFindInPageCacheMissPayload(t *testing.T) {
	result := FindInPageCacheMiss("page_missing", "profit")
	if result.Status != "page_id_not_found" ||
		result.Error != "page_id_not_found" ||
		result.ErrorClass != "page_cache_miss" {
		t.Fatalf("unexpected cache miss result: %#v", result)
	}
	if result.Diagnostics == nil ||
		result.Diagnostics.Scope != "page_cache" ||
		result.Diagnostics.PageCacheHit ||
		result.Diagnostics.SuggestedNextTool != "open_page" ||
		!result.Diagnostics.Retryable {
		t.Fatalf("unexpected cache miss diagnostics: %#v", result.Diagnostics)
	}
}

func TestFindInPageNoSearchableTextReportsPageQualityDiagnostics(t *testing.T) {
	page := Page{
		PageID:      "page_empty",
		Title:       "Empty App",
		ContentType: "text/html; charset=utf-8",
		Text:        "<html><body></body></html>",
	}
	result := FindInPage(page, "profit", 5, 40)
	if result.MatchCount != 0 || len(result.Matches) != 0 {
		t.Fatalf("expected no matches for empty page, got %#v", result)
	}
	if result.Diagnostics == nil ||
		result.Diagnostics.Scope != "page_cache" ||
		!result.Diagnostics.PageCacheHit ||
		result.Diagnostics.PageTextOK ||
		result.Diagnostics.Issue != "page_no_text" ||
		result.Diagnostics.MatchStatus != "page_no_text" ||
		result.Diagnostics.SuggestedNextTool != "browser" ||
		result.Diagnostics.Retryable {
		t.Fatalf("expected page_no_text diagnostics with browser handoff, got %#v", result.Diagnostics)
	}
}

func TestFindInPageNoMatchOnJSShellSuggestsBrowser(t *testing.T) {
	page := Page{
		PageID:      "page_shell",
		Title:       "Loading App",
		ContentType: "text/html; charset=utf-8",
		Text:        "Loading... please enable JavaScript",
		WordCount:   4,
	}
	result := FindInPage(page, "profit", 5, 40)
	if result.MatchCount != 0 || len(result.Matches) != 0 {
		t.Fatalf("expected no matches for js shell, got %#v", result)
	}
	if result.Diagnostics == nil ||
		result.Diagnostics.PageTextOK ||
		result.Diagnostics.Issue != "js_shell" ||
		result.Diagnostics.MatchStatus != "no_match" ||
		result.Diagnostics.SuggestedNextTool != "browser" ||
		result.Diagnostics.Retryable {
		t.Fatalf("expected js_shell diagnostics with browser handoff, got %#v", result.Diagnostics)
	}
}

func TestFindInPageBinaryPageSuggestsPDFAndWarnings(t *testing.T) {
	page := Page{
		PageID:      "page_pdf",
		Title:       "Annual Report PDF",
		ContentType: "application/pdf",
		FinalURL:    "https://example.com/report.pdf",
	}
	result := FindInPage(page, "profit", 5, 40)
	if result.MatchCount != 0 || len(result.Matches) != 0 {
		t.Fatalf("expected no matches for binary page, got %#v", result)
	}
	if result.Diagnostics == nil ||
		result.Diagnostics.PageTextOK ||
		result.Diagnostics.Issue != "binary_or_pdf" ||
		result.Diagnostics.MatchStatus != "page_no_text" ||
		result.Diagnostics.SuggestedNextTool != "pdf" ||
		result.Diagnostics.Retryable {
		t.Fatalf("expected binary_or_pdf diagnostics with pdf handoff, got %#v", result.Diagnostics)
	}
	if len(result.Diagnostics.Warnings) == 0 || !containsString(result.Diagnostics.Warnings, "binary_or_pdf") {
		t.Fatalf("expected binary_or_pdf warning, got %#v", result.Diagnostics.Warnings)
	}
}

func TestFindInPageFallsBackToPlainTextAndDefaultLimit(t *testing.T) {
	page := Page{
		PageID: "page_plain",
		Title:  "Operations Notes",
		Text:   "profit one\nprofit two\nprofit three\nprofit four\nprofit five\nprofit six",
	}
	result := FindInPage(page, "profit", 0, 0)
	if result.MatchCount != 6 {
		t.Fatalf("expected 6 total matches, got %#v", result)
	}
	if len(result.Matches) != 5 || !result.HasMore {
		t.Fatalf("expected default max_matches clamp to 5 with has_more, got %#v", result)
	}
	if result.Matches[0].Heading != "Operations Notes" {
		t.Fatalf("expected plain-text fallback to inherit title as heading, got %#v", result.Matches[0])
	}
}

func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
