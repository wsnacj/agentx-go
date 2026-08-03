package retrieval

import "testing"

func TestInferOpenPageWarningConsentWall(t *testing.T) {
	page := Page{
		ContentType: "text/html; charset=utf-8",
		Title:       "Investors",
		Text:        "Welcome to example.com! We would like to use analytics cookies and other similar tracking technologies to help us improve our website. Manage cookies or accept all cookies.",
		WordCount:   24,
	}

	if got := InferOpenPageWarning(page); got != "consent_wall" {
		t.Fatalf("expected consent_wall, got %q", got)
	}
	diagnostics := BuildOpenPageDiagnostics(page, PageDiagnosticsInput{ExtractMode: "markdown"})
	if diagnostics == nil || diagnostics.PageTextOK || diagnostics.Issue != "consent_wall" || diagnostics.SuggestedNextTool != "browser" {
		t.Fatalf("expected consent wall diagnostics with browser next step, got %#v", diagnostics)
	}
}

func TestInferOpenPageWarningJSShell(t *testing.T) {
	page := Page{
		ContentType: "text/html; charset=utf-8",
		Title:       "App",
		Text:        "Loading... Please enable JavaScript to run this app.",
		WordCount:   8,
	}

	if got := InferOpenPageWarning(page); got != "js_shell" {
		t.Fatalf("expected js_shell, got %q", got)
	}
	diagnostics := BuildOpenPageDiagnostics(page, PageDiagnosticsInput{ExtractMode: "markdown"})
	if diagnostics == nil || diagnostics.PageTextOK || diagnostics.Issue != "js_shell" || diagnostics.SuggestedNextTool != "browser" {
		t.Fatalf("expected JS shell diagnostics with browser next step, got %#v", diagnostics)
	}
}

func TestInferOpenPageWarningPageNoText(t *testing.T) {
	page := Page{
		ContentType: "text/html; charset=utf-8",
		Text:        "",
		Markdown:    "",
		WordCount:   0,
	}

	if got := InferOpenPageWarning(page); got != "page_no_text" {
		t.Fatalf("expected page_no_text, got %q", got)
	}
	diagnostics := BuildOpenPageDiagnostics(page, PageDiagnosticsInput{ExtractMode: "markdown"})
	if diagnostics == nil || diagnostics.PageTextOK || diagnostics.Issue != "page_no_text" || diagnostics.SuggestedNextTool != "browser" {
		t.Fatalf("expected no-text diagnostics with browser next step, got %#v", diagnostics)
	}
}

func TestInferOpenPageWarningRawEmptyHTMLFallback(t *testing.T) {
	page := Page{
		ContentType: "text/html; charset=utf-8",
		Text:        "<html><body></body></html>",
		WordCount:   1,
		Extractor:   "html_fallback",
	}

	if got := InferOpenPageWarning(page); got != "page_no_text" {
		t.Fatalf("expected page_no_text for raw empty HTML fallback, got %q", got)
	}
	diagnostics := BuildOpenPageDiagnostics(page, PageDiagnosticsInput{ExtractMode: "html_raw_fallback"})
	if diagnostics == nil || diagnostics.PageTextOK || diagnostics.Issue != "page_no_text" || diagnostics.SuggestedNextTool != "browser" {
		t.Fatalf("expected raw empty HTML fallback to be marked not text-ready, got %#v", diagnostics)
	}
}

func TestInferOpenPageWarningDoesNotFlagNormalCookieArticle(t *testing.T) {
	page := Page{
		ContentType: "text/html; charset=utf-8",
		Title:       "Cookie usage guide",
		Text:        "This article explains how HTTP cookies work in browsers, including request headers, response headers, cache behavior, authentication sessions, privacy tradeoffs, and testing strategies for web applications.",
		WordCount:   25,
	}

	if got := InferOpenPageWarning(page); got != "" {
		t.Fatalf("expected no warning for normal cookie article, got %q", got)
	}
}
