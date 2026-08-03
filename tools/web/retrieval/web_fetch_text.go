package retrieval

import (
	"encoding/json"
	"fmt"
	neturl "net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var (
	htmlTagRe    = regexp.MustCompile(`(?s)<[^>]*>`)
	htmlSpaceRe  = regexp.MustCompile(`[\t\r\f\v]+`)
	htmlTitleRe  = regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)
	scriptTagRe  = regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`)
	styleTagRe   = regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`)
	commentTagRe = regexp.MustCompile(`(?s)<!--.*?-->`)
	mdFenceRe    = regexp.MustCompile("(?s)```[^\n]*\n?(.*?)```")
	mdInlineRe   = regexp.MustCompile("`([^`]+)`")
	mdImageRe    = regexp.MustCompile(`!\[[^\]]*]\([^)]+\)`)
	mdLinkRe     = regexp.MustCompile(`\[([^\]]+)\]\([^)]+\)`)
	mdHeadingRe  = regexp.MustCompile(`(?m)^#{1,6}\s+`)
	mdBulletRe   = regexp.MustCompile(`(?m)^\s*[-*+]\s+`)
	mdNumberRe   = regexp.MustCompile(`(?m)^\s*\d+\.\s+`)
)

func normalizeWebFetchContentType(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	if idx := strings.Index(trimmed, ";"); idx >= 0 {
		trimmed = trimmed[:idx]
	}
	return strings.ToLower(strings.TrimSpace(trimmed))
}

func normalizeWebFetchExtractMode(raw string) string {
	mode := strings.ToLower(strings.TrimSpace(raw))
	switch mode {
	case "", "auto":
		return "auto"
	case "text", "readable", "extract":
		return "text"
	case "markdown", "md":
		return "markdown"
	case "raw", "source", "raw_html":
		return "raw"
	default:
		return "auto"
	}
}

// NormalizeWebFetchContentType canonicalizes content-type strings for shared fetch/page flows.
func NormalizeWebFetchContentType(raw string) string {
	return normalizeWebFetchContentType(raw)
}

// NormalizeWebFetchExtractMode canonicalizes fetch extract modes for shared fetch/page flows.
func NormalizeWebFetchExtractMode(raw string) string {
	return normalizeWebFetchExtractMode(raw)
}

func NormalizeFetchContent(raw string, contentType string, requestedMode string) WebFetchExtractResult {
	return normalizeFetchContentWithURL(raw, contentType, requestedMode, "")
}

func NormalizeFetchContentWithURL(raw string, contentType string, requestedMode string, pageURL string) WebFetchExtractResult {
	return normalizeFetchContentWithURL(raw, contentType, requestedMode, pageURL)
}

func LooksLikeMarkdownContentType(contentType string) bool {
	return looksLikeMarkdownContentType(contentType)
}

func CompactWhitespace(value string) string {
	return compactWhitespace(value)
}

func FirstNonBlank(values ...string) string {
	return firstNonBlank(values...)
}

func MarkdownToPlainText(value string) string {
	return markdownToPlainText(value)
}

func LooksLikeHTML(value string) bool {
	return looksLikeHTML(value)
}

func TruncateToolText(value string, max int) string {
	return truncateToolText(value, max)
}

func normalizeFetchContentWithURL(raw string, contentType string, requestedMode string, pageURL string) WebFetchExtractResult {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return WebFetchExtractResult{Mode: "empty"}
	}
	mode := normalizeWebFetchExtractMode(requestedMode)
	if mode == "raw" {
		return WebFetchExtractResult{
			Content: compactWhitespace(trimmed),
			Mode:    "raw",
		}
	}
	lowerType := strings.ToLower(strings.TrimSpace(contentType))
	isHTML := strings.Contains(lowerType, "text/html") || looksLikeHTML(trimmed)
	if isHTML {
		return extractReadableHTMLContentWithURL(trimmed, mode, pageURL)
	}
	if looksLikeMarkdownContentType(lowerType) {
		return normalizeMarkdownFetchContent(trimmed, mode)
	}

	if mode != "raw" && looksLikeJSON(lowerType, trimmed) {
		if normalizedJSON, ok := normalizeJSONContent(trimmed); ok {
			return WebFetchExtractResult{
				Content: normalizedJSON,
				Mode:    "json",
			}
		}
	}
	return WebFetchExtractResult{
		Content: compactWhitespace(trimmed),
		Mode:    "text",
	}
}

func shouldTryWebFetchFirecrawlHTMLFallback(contentType string, extract WebFetchExtractResult) bool {
	if normalizeWebFetchExtractMode(extract.Mode) == "raw" {
		return false
	}
	lowerType := strings.ToLower(strings.TrimSpace(contentType))
	if !strings.Contains(lowerType, "text/html") {
		return false
	}
	if strings.TrimSpace(extract.Content) == "" {
		return true
	}
	switch strings.TrimSpace(extract.Mode) {
	case "html_raw_fallback", "raw_fallback", "empty":
		return true
	default:
		return false
	}
}

func resolveWebFetchExtractor(contentType string, mode string) string {
	normalizedMode := strings.TrimSpace(mode)
	switch {
	case strings.HasPrefix(normalizedMode, "firecrawl_"):
		return "firecrawl"
	case normalizedMode == "html_text" || normalizedMode == "html_markdown":
		return "readability"
	case normalizedMode == "html_raw_fallback":
		return "html_fallback"
	case normalizedMode == "raw_fallback":
		if strings.Contains(contentType, "text/html") {
			return "html_fallback"
		}
		return "raw_fallback"
	case normalizedMode == "markdown" || normalizedMode == "markdown_text":
		if strings.Contains(contentType, "text/markdown") || strings.Contains(contentType, "application/markdown") || strings.Contains(contentType, "text/x-markdown") || strings.Contains(contentType, "text/md") {
			return "cf_markdown"
		}
		return "markdown"
	case normalizedMode == "json":
		return "json"
	case normalizedMode == "raw":
		return "raw"
	case normalizedMode == "text":
		return "text"
	case normalizedMode == "binary_unsupported":
		return "binary_unsupported"
	case normalizedMode == "empty":
		return "empty"
	default:
		return normalizedMode
	}
}

func looksLikeMarkdownContentType(contentType string) bool {
	lower := strings.ToLower(strings.TrimSpace(contentType))
	return strings.Contains(lower, "text/markdown") ||
		strings.Contains(lower, "text/x-markdown") ||
		strings.Contains(lower, "application/markdown") ||
		strings.Contains(lower, "text/md")
}

func normalizeMarkdownFetchContent(trimmed string, mode string) WebFetchExtractResult {
	if mode == "text" {
		text := markdownToPlainText(trimmed)
		return WebFetchExtractResult{
			Content:   text,
			Mode:      "markdown_text",
			WordCount: len(strings.Fields(text)),
		}
	}
	normalized := normalizeMarkdownWhitespace(trimmed)
	return WebFetchExtractResult{
		Content:   normalized,
		Mode:      "markdown",
		WordCount: len(strings.Fields(markdownToPlainText(normalized))),
	}
}

func fallbackHTMLExtractResult(title string, result WebFetchExtractResult, fallbackText string, trimmed string) WebFetchExtractResult {
	if strings.TrimSpace(result.Content) != "" {
		if runeLen(result.Content) < 48 && len(trimmed) > 512 {
			fallback := compactWhitespace(firstNonBlank(fallbackText, trimmed))
			return WebFetchExtractResult{
				Content:         fallback,
				Title:           firstNonBlank(result.Title, title),
				SiteName:        result.SiteName,
				Byline:          result.Byline,
				Excerpt:         result.Excerpt,
				Mode:            "html_raw_fallback",
				WordCount:       len(strings.Fields(fallback)),
				FallbackUsed:    true,
				ReadabilityUsed: result.ReadabilityUsed,
			}
		}
		return result
	}
	fallback := compactWhitespace(firstNonBlank(fallbackText, trimmed))
	if fallback == "" {
		fallback = trimmed
	}
	return WebFetchExtractResult{
		Content:         fallback,
		Title:           firstNonBlank(result.Title, title),
		SiteName:        result.SiteName,
		Byline:          result.Byline,
		Excerpt:         result.Excerpt,
		Mode:            "html_raw_fallback",
		WordCount:       len(strings.Fields(fallback)),
		FallbackUsed:    true,
		ReadabilityUsed: result.ReadabilityUsed,
	}
}

func extractHTMLText(trimmed string) (title string, text string) {
	if match := htmlTitleRe.FindStringSubmatch(trimmed); len(match) > 1 {
		title = compactWhitespace(match[1])
	}
	sanitized := commentTagRe.ReplaceAllString(trimmed, " ")
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(sanitized))
	if err == nil {
		body := doc.Find("body").First()
		if body == nil || body.Length() == 0 {
			body = doc.Selection
		}
		sanitizeReadableSelection(body)
		text = compactWhitespace(body.Text())
		if text != "" {
			return title, text
		}
	}
	text = scriptTagRe.ReplaceAllString(sanitized, " ")
	text = styleTagRe.ReplaceAllString(text, " ")
	text = htmlTagRe.ReplaceAllString(text, " ")
	text = strings.ReplaceAll(text, "&nbsp;", " ")
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = compactWhitespace(text)
	return title, text
}

func pickReadableRoot(doc *goquery.Document) *goquery.Selection {
	if doc == nil {
		return nil
	}
	best := doc.Find("body").First()
	bestScore := readableSelectionScore(best)
	candidates := []string{
		"article",
		"main",
		`[role="main"]`,
		".article-content",
		".entry-content",
		".post-content",
		".content",
		"#content",
	}
	for _, selector := range candidates {
		doc.Find(selector).Each(func(_ int, sel *goquery.Selection) {
			score := readableSelectionScore(sel)
			if score > bestScore {
				best = sel
				bestScore = score
			}
		})
	}
	if best == nil || best.Length() == 0 {
		return doc.Selection
	}
	return best
}

func readableSelectionScore(sel *goquery.Selection) int {
	if sel == nil || sel.Length() == 0 {
		return 0
	}
	return runeLen(compactWhitespace(sel.Text()))
}

func renderSelectionMarkdown(sel *goquery.Selection) string {
	if sel == nil || sel.Length() == 0 {
		return ""
	}
	blocks := make([]string, 0, 16)
	sel.Find("h1, h2, h3, h4, h5, h6, p, li, blockquote, pre").Each(func(_ int, node *goquery.Selection) {
		text := strings.TrimSpace(node.Text())
		if text == "" {
			return
		}
		switch goquery.NodeName(node) {
		case "h1":
			blocks = append(blocks, "# "+compactWhitespace(text))
		case "h2":
			blocks = append(blocks, "## "+compactWhitespace(text))
		case "h3":
			blocks = append(blocks, "### "+compactWhitespace(text))
		case "h4":
			blocks = append(blocks, "#### "+compactWhitespace(text))
		case "h5":
			blocks = append(blocks, "##### "+compactWhitespace(text))
		case "h6":
			blocks = append(blocks, "###### "+compactWhitespace(text))
		case "li":
			blocks = append(blocks, "- "+compactWhitespace(text))
		case "blockquote":
			blocks = append(blocks, "> "+compactWhitespace(text))
		case "pre":
			blocks = append(blocks, "```\n"+strings.TrimSpace(text)+"\n```")
		default:
			blocks = append(blocks, compactWhitespace(text))
		}
	})
	if len(blocks) == 0 {
		return compactWhitespace(sel.Text())
	}
	return strings.TrimSpace(strings.Join(blocks, "\n\n"))
}

func metaContent(doc *goquery.Document, attr string, value string) string {
	if doc == nil {
		return ""
	}
	selector := fmt.Sprintf(`meta[%s="%s"]`, attr, value)
	return compactWhitespace(doc.Find(selector).First().AttrOr("content", ""))
}

func firstNonBlank(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func resolveWebFetchSiteName(urlValue string) string {
	parsed, err := neturl.Parse(strings.TrimSpace(urlValue))
	if err != nil {
		return ""
	}
	host := strings.TrimSpace(parsed.Hostname())
	host = strings.TrimPrefix(host, "www.")
	return host
}

func looksLikeJSON(contentType string, content string) bool {
	if strings.Contains(contentType, "application/json") || strings.Contains(contentType, "+json") {
		return true
	}
	trimmed := strings.TrimSpace(content)
	return strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[")
}

func normalizeJSONContent(raw string) (string, bool) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", false
	}
	var payload any
	if err := json.Unmarshal([]byte(trimmed), &payload); err != nil {
		return "", false
	}
	blob, err := json.Marshal(payload)
	if err != nil {
		return "", false
	}
	return string(blob), true
}

func compactWhitespace(value string) string {
	if value == "" {
		return ""
	}
	value = stripInvisibleUnicode(value)
	value = htmlSpaceRe.ReplaceAllString(value, " ")
	value = strings.ReplaceAll(value, "\n", " ")
	value = strings.ReplaceAll(value, "\r", " ")
	value = strings.ReplaceAll(value, "  ", " ")
	return strings.TrimSpace(value)
}

func normalizeMarkdownWhitespace(value string) string {
	if value == "" {
		return ""
	}
	value = stripInvisibleUnicode(value)
	value = strings.ReplaceAll(value, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")
	lines := strings.Split(value, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}
	out := strings.TrimSpace(strings.Join(lines, "\n"))
	for strings.Contains(out, "\n\n\n\n") {
		out = strings.ReplaceAll(out, "\n\n\n\n", "\n\n\n")
	}
	return out
}

func markdownToPlainText(value string) string {
	if value == "" {
		return ""
	}
	text := strings.ReplaceAll(value, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	text = mdImageRe.ReplaceAllString(text, "")
	text = mdLinkRe.ReplaceAllString(text, "$1")
	text = mdFenceRe.ReplaceAllString(text, "$1")
	text = mdInlineRe.ReplaceAllString(text, "$1")
	text = mdHeadingRe.ReplaceAllString(text, "")
	text = mdBulletRe.ReplaceAllString(text, "")
	text = mdNumberRe.ReplaceAllString(text, "")
	return normalizeMarkdownWhitespace(text)
}

func looksLikeHTML(value string) bool {
	lower := strings.ToLower(strings.TrimSpace(value))
	if lower == "" {
		return false
	}
	return strings.HasPrefix(lower, "<!doctype html") || strings.HasPrefix(lower, "<html")
}

func truncateToolText(value string, max int) string {
	return truncateWithEllipsis(value, max)
}

func joinWebFetchWarnings(parts ...string) string {
	out := make([]string, 0, len(parts))
	seen := map[string]bool{}
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" || seen[trimmed] {
			continue
		}
		seen[trimmed] = true
		out = append(out, trimmed)
	}
	return strings.Join(out, "; ")
}
