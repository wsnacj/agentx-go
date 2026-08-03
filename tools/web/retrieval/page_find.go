package retrieval

import (
	"regexp"
	"strings"
)

type PageMatch struct {
	Start   int    `json:"start"`
	End     int    `json:"end"`
	Snippet string `json:"snippet"`
	Heading string `json:"heading,omitempty"`
}

type FindInPageResult struct {
	PageID      string                 `json:"page_id"`
	Query       string                 `json:"query"`
	Status      string                 `json:"status,omitempty"`
	Error       string                 `json:"error,omitempty"`
	ErrorClass  string                 `json:"error_class,omitempty"`
	Title       string                 `json:"title,omitempty"`
	FinalURL    string                 `json:"final_url,omitempty"`
	MatchCount  int                    `json:"match_count"`
	HasMore     bool                   `json:"has_more,omitempty"`
	Matches     []PageMatch            `json:"matches,omitempty"`
	Diagnostics *FindInPageDiagnostics `json:"diagnostics,omitempty"`
}

type FindInPageDiagnostics struct {
	Scope             string   `json:"scope,omitempty"`
	PageCacheHit      bool     `json:"page_cache_hit"`
	PageTextOK        bool     `json:"page_text_ok,omitempty"`
	Issue             string   `json:"issue,omitempty"`
	MatchStatus       string   `json:"match_status,omitempty"`
	SuggestedNextTool string   `json:"suggested_next_tool,omitempty"`
	Retryable         bool     `json:"retryable,omitempty"`
	Warnings          []string `json:"warnings,omitempty"`
}

type pageSearchSegment struct {
	Text    string
	Heading string
	Offset  int
}

var (
	pageFindSpaceRe       = regexp.MustCompile(`\s+`)
	pageFindLinkRe        = regexp.MustCompile(`\[(.*?)\]\([^)]+\)`)
	pageFindImageRe       = regexp.MustCompile(`!\[[^\]]*]\([^)]+\)`)
	pageFindInlineCodeRe  = regexp.MustCompile("`([^`]+)`")
	pageFindBulletRe      = regexp.MustCompile(`^\s*[-*+]\s+`)
	pageFindNumberLineRe  = regexp.MustCompile(`^\s*\d+\.\s+`)
	pageFindHeadingLineRe = regexp.MustCompile(`^\s*#{1,6}\s+`)
)

func FindInPage(page Page, query string, maxMatches int, contextChars int) FindInPageResult {
	query = strings.TrimSpace(query)
	result := FindInPageResult{
		PageID:   strings.TrimSpace(page.PageID),
		Query:    query,
		Status:   "ok",
		Title:    strings.TrimSpace(page.Title),
		FinalURL: strings.TrimSpace(page.FinalURL),
	}
	if query == "" {
		result.Diagnostics = findInPageDiagnostics(page, "empty_query")
		return result
	}
	if maxMatches <= 0 {
		maxMatches = 5
	}
	if maxMatches > 20 {
		maxMatches = 20
	}
	if contextChars <= 0 {
		contextChars = 140
	}
	if contextChars > 400 {
		contextChars = 400
	}
	segments := buildPageSearchSegments(page)
	if len(segments) == 0 {
		result.Diagnostics = findInPageDiagnostics(page, "page_no_text")
		return result
	}
	lowerQuery := strings.ToLower(query)
	for _, segment := range segments {
		if strings.TrimSpace(segment.Text) == "" {
			continue
		}
		matches := findAllSegmentMatchIndexes(segment.Text, lowerQuery)
		if len(matches) == 0 {
			continue
		}
		result.MatchCount += len(matches)
		for _, byteIndex := range matches {
			if len(result.Matches) >= maxMatches {
				result.HasMore = true
				continue
			}
			start := runeIndexForByteIndex(segment.Text, byteIndex)
			end := start + len([]rune(query))
			result.Matches = append(result.Matches, PageMatch{
				Start:   segment.Offset + start,
				End:     segment.Offset + end,
				Snippet: snippetAroundRunes(segment.Text, start, end, contextChars),
				Heading: strings.TrimSpace(segment.Heading),
			})
		}
	}
	if result.MatchCount > len(result.Matches) {
		result.HasMore = true
	}
	if result.MatchCount > 0 {
		result.Diagnostics = findInPageDiagnostics(page, "matched")
	} else {
		result.Diagnostics = findInPageDiagnostics(page, "no_match")
	}
	return result
}

func FindInPageCacheMiss(pageID string, query string) FindInPageResult {
	return FindInPageResult{
		PageID:     strings.TrimSpace(pageID),
		Query:      strings.TrimSpace(query),
		Status:     "page_id_not_found",
		Error:      "page_id_not_found",
		ErrorClass: "page_cache_miss",
		Diagnostics: &FindInPageDiagnostics{
			Scope:             "page_cache",
			PageCacheHit:      false,
			MatchStatus:       "page_not_found",
			SuggestedNextTool: "open_page",
			Retryable:         true,
		},
	}
}

func findInPageDiagnostics(page Page, matchStatus string) *FindInPageDiagnostics {
	pageTextOK := openPageTextOK(page)
	issue := strings.TrimSpace(InferOpenPageWarning(page))
	contentType := normalizeWebFetchContentType(page.ContentType)
	if issue == "" && strings.Contains(contentType, "application/pdf") && !pageTextOK {
		issue = "binary_or_pdf"
	}
	if issue == "" && !pageTextOK && strings.TrimSpace(matchStatus) == "page_no_text" {
		issue = "page_no_text"
	}
	suggestedNextTool := ""
	retryable := false
	switch strings.TrimSpace(matchStatus) {
	case "matched":
		suggestedNextTool = ""
	case "no_match":
		if pageTextOK {
			suggestedNextTool = "find_in_page"
			retryable = true
		} else {
			suggestedNextTool = findInPagePageQualityNextTool(page, pageTextOK, issue)
		}
	case "empty_query":
		suggestedNextTool = "find_in_page"
		retryable = true
	case "page_no_text":
		suggestedNextTool = findInPagePageQualityNextTool(page, pageTextOK, issue)
	default:
		suggestedNextTool = "find_in_page"
	}
	diagnostics := &FindInPageDiagnostics{
		Scope:             "page_cache",
		PageCacheHit:      true,
		PageTextOK:        pageTextOK,
		Issue:             issue,
		MatchStatus:       strings.TrimSpace(matchStatus),
		SuggestedNextTool: suggestedNextTool,
		Retryable:         retryable,
	}
	diagnostics.Warnings = openPageDiagnosticWarnings(page, &PageDiagnostics{
		PageTextOK: pageTextOK,
		Issue:      issue,
	})
	return diagnostics
}

func findInPagePageQualityNextTool(page Page, pageTextOK bool, issue string) string {
	suggested := strings.TrimSpace(openPageSuggestedNextTool(page, &PageDiagnostics{
		PageTextOK: pageTextOK,
		Issue:      strings.TrimSpace(issue),
	}))
	if suggested == "" || suggested == "find_in_page" {
		return "open_page"
	}
	return suggested
}

func buildPageSearchSegments(page Page) []pageSearchSegment {
	if segments := buildMarkdownPageSearchSegments(page.Markdown, page.Title); len(segments) > 0 {
		return segments
	}
	return buildPlainTextPageSearchSegments(openPageReadableText(page), page.Title)
}

func buildMarkdownPageSearchSegments(markdown string, title string) []pageSearchSegment {
	lines := strings.Split(strings.TrimSpace(markdown), "\n")
	if len(lines) == 0 {
		return nil
	}
	currentHeading := strings.TrimSpace(title)
	currentLines := make([]string, 0, 4)
	segments := make([]pageSearchSegment, 0, 8)
	offset := 0
	flush := func() {
		if len(currentLines) == 0 {
			return
		}
		text := pageFindCompact(strings.Join(currentLines, " "))
		currentLines = currentLines[:0]
		if text == "" {
			return
		}
		segments = append(segments, pageSearchSegment{
			Text:    text,
			Heading: currentHeading,
			Offset:  offset,
		})
		offset += len([]rune(text)) + 1
	}
	for _, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		if line == "" {
			flush()
			continue
		}
		if heading, ok := markdownHeadingText(line); ok {
			flush()
			currentHeading = heading
			continue
		}
		plain := normalizeMarkdownSearchLine(line)
		if plain == "" {
			continue
		}
		currentLines = append(currentLines, plain)
	}
	flush()
	return segments
}

func buildPlainTextPageSearchSegments(text string, title string) []pageSearchSegment {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return nil
	}
	parts := strings.Split(trimmed, "\n")
	segments := make([]pageSearchSegment, 0, len(parts))
	offset := 0
	currentHeading := strings.TrimSpace(title)
	for _, raw := range parts {
		line := pageFindCompact(raw)
		if line == "" {
			continue
		}
		segments = append(segments, pageSearchSegment{
			Text:    line,
			Heading: currentHeading,
			Offset:  offset,
		})
		offset += len([]rune(line)) + 1
	}
	if len(segments) == 0 {
		segments = append(segments, pageSearchSegment{
			Text:    pageFindCompact(trimmed),
			Heading: currentHeading,
			Offset:  0,
		})
	}
	return segments
}

func markdownHeadingText(line string) (string, bool) {
	if !pageFindHeadingLineRe.MatchString(line) {
		return "", false
	}
	trimmed := pageFindHeadingLineRe.ReplaceAllString(line, "")
	trimmed = normalizeMarkdownSearchLine(trimmed)
	if trimmed == "" {
		return "", false
	}
	return trimmed, true
}

func normalizeMarkdownSearchLine(line string) string {
	line = strings.TrimSpace(line)
	if line == "" {
		return ""
	}
	line = strings.TrimPrefix(line, ">")
	line = strings.TrimSpace(line)
	line = pageFindImageRe.ReplaceAllString(line, "")
	line = pageFindLinkRe.ReplaceAllString(line, "$1")
	line = pageFindInlineCodeRe.ReplaceAllString(line, "$1")
	line = pageFindBulletRe.ReplaceAllString(line, "")
	line = pageFindNumberLineRe.ReplaceAllString(line, "")
	line = strings.NewReplacer("*", "", "_", "", "~", "", "#", "").Replace(line)
	return pageFindCompact(line)
}

func pageFindCompact(raw string) string {
	return strings.TrimSpace(pageFindSpaceRe.ReplaceAllString(strings.TrimSpace(raw), " "))
}

func findAllSegmentMatchIndexes(text string, lowerQuery string) []int {
	if strings.TrimSpace(text) == "" || lowerQuery == "" {
		return nil
	}
	lowerText := strings.ToLower(text)
	out := make([]int, 0, 4)
	offset := 0
	for offset < len(lowerText) {
		idx := strings.Index(lowerText[offset:], lowerQuery)
		if idx < 0 {
			break
		}
		absolute := offset + idx
		out = append(out, absolute)
		offset = absolute + len(lowerQuery)
	}
	return out
}

func runeIndexForByteIndex(text string, byteIndex int) int {
	if byteIndex <= 0 {
		return 0
	}
	if byteIndex >= len(text) {
		return len([]rune(text))
	}
	return len([]rune(text[:byteIndex]))
}

func snippetAroundRunes(text string, start int, end int, contextChars int) string {
	runes := []rune(text)
	if start < 0 {
		start = 0
	}
	if end < start {
		end = start
	}
	if end > len(runes) {
		end = len(runes)
	}
	snippetStart := start - contextChars
	if snippetStart < 0 {
		snippetStart = 0
	}
	snippetEnd := end + contextChars
	if snippetEnd > len(runes) {
		snippetEnd = len(runes)
	}
	snippet := string(runes[snippetStart:snippetEnd])
	snippet = pageFindCompact(snippet)
	if snippetStart > 0 {
		snippet = "... " + snippet
	}
	if snippetEnd < len(runes) {
		snippet += " ..."
	}
	return snippet
}
