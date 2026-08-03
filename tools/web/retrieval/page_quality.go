package retrieval

import "strings"

var openPageJSShellMarkers = []string{
	"enable javascript",
	"javascript required",
	"javascript is required",
	"please turn on javascript",
	"please enable javascript",
	"run this app",
	"loading...",
	"checking your browser",
}

var openPageConsentWallMarkers = []string{
	"we use cookies",
	"use analytics cookies",
	"tracking technologies",
	"cookie settings",
	"accept all cookies",
	"reject all cookies",
	"manage cookies",
	"cookie preferences",
	"privacy preferences",
}

// InferOpenPageWarning derives a generic page-quality warning for page follow-up flows.
// It intentionally stays domain-neutral and only classifies generic shell/no-text/consent-wall cases.
func InferOpenPageWarning(page Page) string {
	if strings.TrimSpace(page.Warning) != "" {
		return strings.TrimSpace(page.Warning)
	}
	if !strings.Contains(normalizeWebFetchContentType(page.ContentType), "text/html") {
		return ""
	}
	text := openPageReadableText(page)
	wordCount := page.WordCount
	if wordCount <= 0 {
		wordCount = len(strings.Fields(text))
	}
	if strings.TrimSpace(text) == "" || wordCount == 0 {
		return "page_no_text"
	}
	lower := strings.ToLower(strings.Join([]string{
		strings.TrimSpace(page.Title),
		strings.TrimSpace(page.Excerpt),
		text,
	}, "\n"))
	if wordCount <= 32 {
		for _, marker := range openPageJSShellMarkers {
			if strings.Contains(lower, marker) {
				return "js_shell"
			}
		}
	}
	if openPageLooksLikeConsentWall(lower, wordCount) {
		return "consent_wall"
	}
	if wordCount <= 16 && len(page.Links) >= 8 {
		return "page_no_text"
	}
	return ""
}

func openPageLooksLikeConsentWall(lower string, wordCount int) bool {
	lower = strings.TrimSpace(lower)
	if lower == "" || wordCount <= 0 || wordCount > 140 {
		return false
	}
	hasCookieContext := strings.Contains(lower, "cookie") || strings.Contains(lower, "cookies")
	if !hasCookieContext && !strings.Contains(lower, "consent") {
		return false
	}
	hits := 0
	for _, marker := range openPageConsentWallMarkers {
		if strings.Contains(lower, marker) {
			hits++
		}
	}
	if hits >= 2 {
		return true
	}
	return hasCookieContext &&
		strings.Contains(lower, "tracking") &&
		(strings.Contains(lower, "analytics") || strings.Contains(lower, "preferences"))
}

type PageDiagnosticsInput struct {
	ExtractMode      string
	Redirected       bool
	RedirectCount    int
	CacheHit         bool
	BodyTruncated    bool
	ContentTruncated bool
	ResponseBytes    int
}

func BuildOpenPageDiagnostics(page Page, input PageDiagnosticsInput) *PageDiagnostics {
	diagnostics := &PageDiagnostics{
		PageTextOK:       openPageTextOK(page),
		Issue:            strings.TrimSpace(InferOpenPageWarning(page)),
		ExtractionMethod: strings.TrimSpace(page.Extractor),
		ExtractMode:      strings.TrimSpace(input.ExtractMode),
		Redirected:       input.Redirected,
		RedirectCount:    input.RedirectCount,
		CacheHit:         input.CacheHit,
		BodyTruncated:    input.BodyTruncated,
		ContentTruncated: page.Truncated || input.ContentTruncated,
		ResponseBytes:    input.ResponseBytes,
	}
	diagnostics.Warnings = openPageDiagnosticWarnings(page, diagnostics)
	diagnostics.SuggestedNextTool = openPageSuggestedNextTool(page, diagnostics)
	return diagnostics
}

func openPageTextOK(page Page) bool {
	text := openPageReadableText(page)
	if text == "" {
		return false
	}
	if strings.TrimSpace(InferOpenPageWarning(page)) != "" {
		return false
	}
	wordCount := page.WordCount
	if wordCount <= 0 {
		wordCount = len(strings.Fields(text))
	}
	return wordCount > 0
}

func openPageReadableText(page Page) string {
	text := strings.TrimSpace(page.Text)
	if strings.Contains(normalizeWebFetchContentType(page.ContentType), "text/html") && looksLikeHTML(text) {
		_, htmlText := extractHTMLText(text)
		text = htmlText
	}
	return compactWhitespace(firstNonBlank(text, markdownToPlainText(page.Markdown), page.Excerpt))
}

func openPageDiagnosticWarnings(page Page, diagnostics *PageDiagnostics) []string {
	if diagnostics == nil {
		return nil
	}
	seen := map[string]bool{}
	out := make([]string, 0, 4)
	add := func(value string) {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			return
		}
		seen[value] = true
		out = append(out, value)
	}
	add(page.Warning)
	add(diagnostics.Issue)
	if diagnostics.Redirected {
		add("redirected")
	}
	if diagnostics.BodyTruncated {
		add("body_truncated")
	}
	if diagnostics.ContentTruncated {
		add("content_truncated")
	}
	if strings.Contains(normalizeWebFetchContentType(page.ContentType), "application/pdf") {
		add("binary_or_pdf")
	}
	return out
}

func openPageSuggestedNextTool(page Page, diagnostics *PageDiagnostics) string {
	if diagnostics == nil {
		return ""
	}
	issue := strings.TrimSpace(diagnostics.Issue)
	contentType := normalizeWebFetchContentType(page.ContentType)
	switch {
	case strings.Contains(contentType, "application/pdf"):
		return "pdf"
	case issue == "js_shell":
		return "browser"
	case issue == "page_no_text":
		return "browser"
	case issue == "consent_wall":
		return "browser"
	case diagnostics.PageTextOK:
		return "find_in_page"
	default:
		return "open_page"
	}
}
