package retrieval

import (
	neturl "net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var markdownOpenPageLinkRe = regexp.MustCompile(`\[(.*?)\]\(([^)]+)\)`)

func ExtractPageLinks(raw string, contentType string, pageURL string, limit int) []PageLink {
	return extractPageLinks(raw, contentType, pageURL, limit)
}

func extractPageLinks(raw string, contentType string, pageURL string, limit int) []PageLink {
	if limit <= 0 {
		limit = 24
	}
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	if strings.Contains(strings.ToLower(strings.TrimSpace(contentType)), "html") || looksLikeHTML(raw) {
		return extractHTMLPageLinks(raw, pageURL, limit)
	}
	if looksLikeMarkdownContentType(contentType) {
		return extractMarkdownPageLinks(raw, pageURL, limit)
	}
	return nil
}

func extractHTMLPageLinks(raw string, pageURL string, limit int) []PageLink {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(raw))
	if err != nil {
		return nil
	}
	base, _ := neturl.Parse(strings.TrimSpace(pageURL))
	links := make([]PageLink, 0, limit)
	seen := map[string]bool{}
	doc.Find("a[href]").EachWithBreak(func(_ int, sel *goquery.Selection) bool {
		href := strings.TrimSpace(sel.AttrOr("href", ""))
		if href == "" {
			return true
		}
		resolved := resolvePageLinkURL(base, href)
		if strings.TrimSpace(resolved) == "" || seen[resolved] {
			return true
		}
		seen[resolved] = true
		text := compactWhitespace(firstNonBlank(sel.Text(), sel.AttrOr("title", ""), resolved))
		links = append(links, PageLink{
			Text: text,
			URL:  resolved,
		})
		return len(links) < limit
	})
	return links
}

func extractMarkdownPageLinks(raw string, pageURL string, limit int) []PageLink {
	base, _ := neturl.Parse(strings.TrimSpace(pageURL))
	links := make([]PageLink, 0, limit)
	seen := map[string]bool{}
	matches := markdownOpenPageLinkRe.FindAllStringSubmatch(raw, limit)
	for _, match := range matches {
		if len(match) < 3 {
			continue
		}
		resolved := resolvePageLinkURL(base, match[2])
		if strings.TrimSpace(resolved) == "" || seen[resolved] {
			continue
		}
		seen[resolved] = true
		links = append(links, PageLink{
			Text: compactWhitespace(match[1]),
			URL:  resolved,
		})
	}
	return links
}

func resolvePageLinkURL(base *neturl.URL, href string) string {
	trimmed := strings.TrimSpace(href)
	if trimmed == "" {
		return ""
	}
	parsed, err := neturl.Parse(trimmed)
	if err != nil {
		return ""
	}
	if base != nil {
		parsed = base.ResolveReference(parsed)
	}
	if strings.TrimSpace(parsed.Scheme) == "" || strings.TrimSpace(parsed.Host) == "" {
		return ""
	}
	return evidenceURL(parsed.String())
}

func DedupePageLinks(items []PageLink) []PageLink {
	if len(items) == 0 {
		return nil
	}
	out := make([]PageLink, 0, len(items))
	seen := map[string]bool{}
	for _, item := range items {
		urlValue := evidenceURL(item.URL)
		if urlValue == "" || seen[urlValue] {
			continue
		}
		seen[urlValue] = true
		text := strings.TrimSpace(item.Text)
		if text == strings.TrimSpace(item.URL) {
			text = urlValue
		}
		out = append(out, PageLink{
			Text: text,
			URL:  urlValue,
		})
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
