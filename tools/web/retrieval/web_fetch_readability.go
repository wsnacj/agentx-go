package retrieval

import (
	"bytes"
	"net/url"
	"strings"

	readability "codeberg.org/readeck/go-readability/v2"
	"github.com/PuerkitoBio/goquery"
	xhtml "golang.org/x/net/html"
)

const (
	webFetchReadabilityMaxHTMLChars             = 1_000_000
	webFetchReadabilityMaxEstimatedNestingDepth = 3_000
)

func ExtractReadableHTMLContent(trimmed string, mode string) WebFetchExtractResult {
	return extractReadableHTMLContentWithURL(trimmed, mode, "")
}

func ExtractReadableHTMLContentWithURL(trimmed string, mode string, pageURL string) WebFetchExtractResult {
	return extractReadableHTMLContentWithURL(trimmed, mode, pageURL)
}

func extractReadableHTMLContentWithURL(trimmed string, mode string, pageURL string) WebFetchExtractResult {
	title, fallbackText := extractHTMLText(trimmed)
	if result, ok := extractHTMLContentWithReadability(trimmed, mode, pageURL); ok {
		result.Title = firstNonBlank(result.Title, title)
		return fallbackHTMLExtractResult(title, result, fallbackText, trimmed)
	}
	return extractReadableHTMLContentHeuristic(trimmed, mode, title, fallbackText)
}

func extractHTMLContentWithReadability(trimmed string, mode string, pageURL string) (WebFetchExtractResult, bool) {
	if strings.TrimSpace(trimmed) == "" {
		return WebFetchExtractResult{}, false
	}
	cleanHTML := sanitizeHTMLForReadability(trimmed)
	if len(cleanHTML) > webFetchReadabilityMaxHTMLChars || exceedsEstimatedHTMLNestingDepth(cleanHTML, webFetchReadabilityMaxEstimatedNestingDepth) {
		return WebFetchExtractResult{}, false
	}

	var pageRef *url.URL
	if parsed, err := url.Parse(strings.TrimSpace(pageURL)); err == nil && parsed != nil && strings.TrimSpace(parsed.Scheme) != "" && strings.TrimSpace(parsed.Host) != "" {
		pageRef = parsed
	}

	parser := readability.NewParser()
	parser.CharThresholds = 0
	article, err := parser.Parse(strings.NewReader(cleanHTML), pageRef)
	if err != nil {
		return WebFetchExtractResult{}, false
	}

	root, text, excerpt, ok := parseReadableHTMLFragment(renderReadabilityArticleHTML(article))
	if !ok {
		text = compactWhitespace(renderReadabilityArticleText(article))
		excerpt = truncateToolText(firstNonBlank(readabilityArticleExcerpt(article), text), 220)
	}
	text = compactWhitespace(firstNonBlank(text, renderReadabilityArticleText(article)))
	if text == "" {
		return WebFetchExtractResult{}, false
	}

	content := text
	extractMode := "html_text"
	if normalizeWebFetchExtractMode(mode) == "markdown" {
		if root == nil || root.Length() == 0 {
			return WebFetchExtractResult{}, false
		}
		content = strings.TrimSpace(renderSelectionMarkdown(root))
		if content == "" {
			return WebFetchExtractResult{}, false
		}
		extractMode = "html_markdown"
	}

	return WebFetchExtractResult{
		Content:         content,
		Title:           strings.TrimSpace(article.Title()),
		SiteName:        strings.TrimSpace(article.SiteName()),
		Byline:          strings.TrimSpace(article.Byline()),
		Excerpt:         truncateToolText(firstNonBlank(readabilityArticleExcerpt(article), excerpt, text), 220),
		Mode:            extractMode,
		WordCount:       len(strings.Fields(text)),
		ReadabilityUsed: true,
	}, true
}

func renderReadabilityArticleHTML(article readability.Article) string {
	var buf bytes.Buffer
	if err := article.RenderHTML(&buf); err != nil {
		return ""
	}
	return buf.String()
}

func renderReadabilityArticleText(article readability.Article) string {
	var buf bytes.Buffer
	if err := article.RenderText(&buf); err != nil {
		return ""
	}
	return buf.String()
}

func readabilityArticleExcerpt(article readability.Article) string {
	if article.Node == nil {
		return ""
	}
	return strings.TrimSpace(article.Excerpt())
}

func sanitizeHTMLForReadability(trimmed string) string {
	sanitized := commentTagRe.ReplaceAllString(trimmed, " ")
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(sanitized))
	if err != nil {
		return sanitized
	}
	body := doc.Find("body").First()
	if body == nil || body.Length() == 0 {
		body = doc.Selection
	}
	sanitizeReadableSelection(body)
	rendered := renderGoqueryDocument(doc)
	if strings.TrimSpace(rendered) == "" {
		return sanitized
	}
	return rendered
}

func renderGoqueryDocument(doc *goquery.Document) string {
	if doc == nil || len(doc.Nodes) == 0 {
		return ""
	}
	var buf bytes.Buffer
	for _, node := range doc.Nodes {
		_ = xhtml.Render(&buf, node)
	}
	return buf.String()
}

func parseReadableHTMLFragment(fragment string) (*goquery.Selection, string, string, bool) {
	trimmed := strings.TrimSpace(fragment)
	if trimmed == "" {
		return nil, "", "", false
	}
	wrapped := "<html><body><article>" + trimmed + "</article></body></html>"
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(wrapped))
	if err != nil {
		return nil, "", "", false
	}
	root := doc.Find("article").First()
	if root == nil || root.Length() == 0 {
		return nil, "", "", false
	}
	sanitizeReadableSelection(root)
	text := compactWhitespace(root.Text())
	if text == "" {
		return nil, "", "", false
	}
	excerpt := truncateToolText(firstNonBlank(compactWhitespace(root.Find("p").First().Text()), text), 220)
	return root, text, excerpt, true
}

func extractReadableHTMLContentHeuristic(trimmed string, mode string, title string, fallbackText string) WebFetchExtractResult {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(trimmed))
	if err != nil {
		return fallbackHTMLExtractResult(title, WebFetchExtractResult{
			Title: title,
		}, fallbackText, trimmed)
	}
	title = firstNonBlank(
		metaContent(doc, "property", "og:title"),
		metaContent(doc, "name", "twitter:title"),
		metaContent(doc, "name", "title"),
		title,
	)
	body := doc.Find("body").First()
	if body == nil || body.Length() == 0 {
		body = doc.Selection
	}
	sanitizeReadableSelection(body)
	fallbackText = compactWhitespace(body.Text())
	root := pickReadableRoot(doc)
	sanitizeReadableSelection(root)
	siteName := firstNonBlank(
		metaContent(doc, "property", "og:site_name"),
		metaContent(doc, "name", "application-name"),
	)
	byline := firstNonBlank(
		metaContent(doc, "name", "author"),
		metaContent(doc, "property", "article:author"),
		compactWhitespace(root.Find(`[rel="author"], [itemprop="author"], .author, .byline`).First().Text()),
	)
	text := compactWhitespace(root.Text())
	content := text
	extractMode := "html_text"
	if normalizeWebFetchExtractMode(mode) == "markdown" {
		content = renderSelectionMarkdown(root)
		extractMode = "html_markdown"
	}
	if strings.TrimSpace(content) == "" {
		content = text
	}
	result := WebFetchExtractResult{
		Content:         content,
		Title:           title,
		SiteName:        siteName,
		Byline:          byline,
		Excerpt:         truncateToolText(firstNonBlank(compactWhitespace(root.Find("p").First().Text()), text), 220),
		Mode:            extractMode,
		WordCount:       len(strings.Fields(text)),
		ReadabilityUsed: true,
	}
	return fallbackHTMLExtractResult(title, result, fallbackText, trimmed)
}

func exceedsEstimatedHTMLNestingDepth(html string, maxDepth int) bool {
	if maxDepth <= 0 || html == "" {
		return false
	}
	voidTags := map[string]bool{
		"area": true, "base": true, "br": true, "col": true, "embed": true,
		"hr": true, "img": true, "input": true, "link": true, "meta": true,
		"param": true, "source": true, "track": true, "wbr": true,
	}
	depth := 0
	length := len(html)
	for i := 0; i < length; i++ {
		if html[i] != '<' {
			continue
		}
		if i+1 >= length {
			continue
		}
		next := html[i+1]
		if next == '!' || next == '?' {
			continue
		}
		j := i + 1
		closing := false
		if html[j] == '/' {
			closing = true
			j++
		}
		for j < length && html[j] <= ' ' {
			j++
		}
		start := j
		for j < length {
			c := html[j]
			if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == ':' || c == '-' {
				j++
				continue
			}
			break
		}
		tagName := strings.ToLower(html[start:j])
		if tagName == "" {
			continue
		}
		if closing {
			if depth > 0 {
				depth--
			}
			continue
		}
		if voidTags[tagName] {
			continue
		}
		selfClosing := false
		for k := j; k < length && k < j+200; k++ {
			if html[k] == '>' {
				if k > 0 && html[k-1] == '/' {
					selfClosing = true
				}
				break
			}
		}
		if selfClosing {
			continue
		}
		depth++
		if depth > maxDepth {
			return true
		}
	}
	return false
}
