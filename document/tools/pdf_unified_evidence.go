package tools

import (
	"strings"
	"unicode"
)

func groundPDFUnifiedAnswerEvidence(prompt string, answer string, documents []pdfUnifiedDocumentArtifacts, payload pdfUnifiedPayload) pdfUnifiedPayload {
	answer = strings.TrimSpace(answer)
	if answer == "" || len(documents) == 0 || len(payload.Documents) == 0 {
		return payload
	}
	for idx := range payload.Documents {
		if idx >= len(documents) || len(payload.Documents[idx].EvidencePages) == 0 {
			continue
		}
		textByPage := make(map[int]string, len(documents[idx].TextResult.Pages))
		for _, page := range documents[idx].TextResult.Pages {
			textByPage[page.Page] = page.Text
		}
		for evidenceIdx := range payload.Documents[idx].EvidencePages {
			evidence := &payload.Documents[idx].EvidencePages[evidenceIdx]
			excerpt := buildPDFUnifiedAnswerGroundedEvidenceExcerpt(prompt, answer, textByPage[evidence.Page], defaultPDFUnifiedEvidenceChars)
			if strings.TrimSpace(excerpt) != "" {
				evidence.Excerpt = excerpt
			}
		}
	}
	return payload
}

type pdfUnifiedEvidenceLine struct {
	Text     string
	Indented bool
}

func buildPDFUnifiedAnswerGroundedEvidenceExcerpt(prompt string, answer string, text string, maxChars int) string {
	groundingQuery := strings.TrimSpace(prompt) + "\n" + answer
	queryExcerpt := buildPDFUnifiedQueryEvidenceExcerpt(groundingQuery, text, maxChars)
	anchors := pdfUnifiedEvidenceNumberAnchors(answer)
	if len(anchors) == 0 || strings.TrimSpace(text) == "" {
		return queryExcerpt
	}

	rawLines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	lines := make([]pdfUnifiedEvidenceLine, 0, len(rawLines))
	for _, rawLine := range rawLines {
		trimmed := strings.TrimSpace(rawLine)
		if trimmed == "" {
			continue
		}
		lines = append(lines, pdfUnifiedEvidenceLine{
			Text:     trimmed,
			Indented: len(rawLine) > len(strings.TrimLeftFunc(rawLine, unicode.IsSpace)),
		})
	}
	if len(lines) == 0 {
		return queryExcerpt
	}

	lineAnchors := make([]map[string]struct{}, len(lines))
	for idx, line := range lines {
		values := pdfUnifiedEvidenceNumberAnchors(line.Text)
		lineAnchors[idx] = make(map[string]struct{}, len(values))
		for _, value := range values {
			lineAnchors[idx][value] = struct{}{}
		}
	}

	selected := make([]int, 0, len(anchors))
	selectedSet := make(map[int]struct{}, len(anchors))
	alternates := make([]int, 0, len(anchors))
	const maxGroundedEvidenceGroups = 6
	for _, anchor := range anchors {
		candidates := make([]PDFPageText, 0, 2)
		candidateIndexes := make([]int, 0, 2)
		for lineIdx := range lines {
			if _, ok := lineAnchors[lineIdx][anchor]; !ok {
				continue
			}
			candidateIndexes = append(candidateIndexes, lineIdx)
			candidates = append(candidates, PDFPageText{
				Page: lineIdx + 1,
				Text: pdfUnifiedGroundedEvidenceLineGroup(lines, lineIdx),
			})
		}
		if len(candidates) == 0 {
			continue
		}
		orderedIndexes := append([]int(nil), candidateIndexes...)
		if ranked, matched := rankPDFUnifiedPages(groundingQuery, candidates); matched {
			orderedIndexes = orderedIndexes[:0]
			for _, item := range ranked {
				orderedIndexes = append(orderedIndexes, item.Page-1)
			}
		}
		covered := false
		for _, lineIdx := range orderedIndexes {
			if _, ok := selectedSet[lineIdx]; ok {
				covered = true
				break
			}
		}
		if covered || len(selected) >= maxGroundedEvidenceGroups {
			continue
		}
		selectedSet[orderedIndexes[0]] = struct{}{}
		selected = append(selected, orderedIndexes[0])
		alternates = append(alternates, orderedIndexes[1:]...)
	}
	for _, lineIdx := range alternates {
		if len(selected) >= maxGroundedEvidenceGroups {
			break
		}
		if _, ok := selectedSet[lineIdx]; ok {
			continue
		}
		selectedSet[lineIdx] = struct{}{}
		selected = append(selected, lineIdx)
	}
	if len(selected) == 0 {
		return queryExcerpt
	}

	groups := make([]string, 0, len(selected))
	for _, lineIdx := range selected {
		groups = append(groups, pdfUnifiedGroundedEvidenceLineGroup(lines, lineIdx))
	}
	grounded := compactPDFUnifiedEvidenceGroups(groups, maxChars)
	if strings.TrimSpace(queryExcerpt) == "" || runeLen(grounded) >= maxChars {
		return grounded
	}
	separator := "\n...\n"
	remaining := maxChars - runeLen(grounded) - runeLen(separator)
	if remaining < 48 {
		return grounded
	}
	return grounded + separator + truncateToolText(queryExcerpt, remaining)
}

func pdfUnifiedGroundedEvidenceLineGroup(lines []pdfUnifiedEvidenceLine, lineIdx int) string {
	if lineIdx < 0 || lineIdx >= len(lines) {
		return ""
	}
	parts := make([]string, 0, 2)
	if lineIdx > 0 && lines[lineIdx].Indented && len(pdfUnifiedEvidenceNumberAnchors(lines[lineIdx-1].Text)) == 0 {
		parts = append(parts, lines[lineIdx-1].Text)
	}
	parts = append(parts, lines[lineIdx].Text)
	return strings.Join(parts, "\n")
}

func pdfUnifiedEvidenceNumberAnchors(value string) []string {
	runes := []rune(value)
	seen := make(map[string]struct{})
	out := make([]string, 0, 8)
	for start := 0; start < len(runes); {
		if !unicode.IsDigit(runes[start]) {
			start++
			continue
		}
		end := start
		hasSeparator := false
		digits := make([]rune, 0, 8)
		for end < len(runes) {
			r := runes[end]
			switch {
			case unicode.IsDigit(r):
				digits = append(digits, r)
			case r == ',' || r == '.' || r == '\u066c' || r == '\uff0c':
				hasSeparator = true
			default:
				goto anchorComplete
			}
			end++
		}

	anchorComplete:
		if len(digits) >= 5 || (hasSeparator && len(digits) >= 4) {
			anchor := string(digits)
			if _, ok := seen[anchor]; !ok {
				seen[anchor] = struct{}{}
				out = append(out, anchor)
				if len(out) == 32 {
					return out
				}
			}
		}
		if end == start {
			end++
		}
		start = end
	}
	return out
}

func compactPDFUnifiedEvidenceGroups(groups []string, maxChars int) string {
	if len(groups) == 0 || maxChars <= 0 {
		return ""
	}
	separator := "\n...\n"
	joined := strings.Join(groups, separator)
	if runeLen(joined) <= maxChars {
		return joined
	}
	contentBudget := maxChars - runeLen(separator)*(len(groups)-1)
	if contentBudget <= 0 {
		return truncateToolText(groups[0], maxChars)
	}
	perGroup := contentBudget / len(groups)
	parts := make([]string, 0, len(groups))
	for _, group := range groups {
		parts = append(parts, truncatePDFUnifiedEvidenceMiddle(group, perGroup))
	}
	return strings.Join(parts, separator)
}

func truncatePDFUnifiedEvidenceMiddle(value string, maxChars int) string {
	runes := []rune(strings.TrimSpace(value))
	if maxChars <= 0 {
		return ""
	}
	if len(runes) <= maxChars {
		return string(runes)
	}
	marker := []rune(" ... ")
	if maxChars <= len(marker)+2 {
		return string(runes[:maxChars])
	}
	remaining := maxChars - len(marker)
	head := remaining * 2 / 3
	tail := remaining - head
	return string(runes[:head]) + string(marker) + string(runes[len(runes)-tail:])
}

type pdfUnifiedEvidenceWindow struct {
	Start int
	End   int
	Text  string
}

func buildPDFUnifiedQueryEvidenceExcerpt(prompt string, text string, maxChars int) string {
	text = strings.TrimSpace(text)
	if text == "" || maxChars <= 0 {
		return ""
	}
	rawLines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	lines := make([]string, 0, len(rawLines))
	for _, line := range rawLines {
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}
	if len(lines) == 0 {
		return ""
	}

	windows := make([]pdfUnifiedEvidenceWindow, 0, len(lines))
	candidates := make([]PDFPageText, 0, len(lines))
	for idx := range lines {
		start := max(0, idx-1)
		end := minInt(len(lines), idx+3)
		windowText := strings.Join(lines[start:end], "\n")
		windows = append(windows, pdfUnifiedEvidenceWindow{Start: start, End: end, Text: windowText})
		candidates = append(candidates, PDFPageText{Page: idx + 1, Text: windowText})
	}
	ranked, matched := rankPDFUnifiedPages(prompt, candidates)
	if !matched {
		return truncateToolText(text, maxChars)
	}

	maxWindows := 3
	fieldCompare := classifyPDFUnifiedQuery(prompt) == pdfUnifiedQueryClassFieldCompare
	if fieldCompare {
		maxWindows = 2
	}
	selected := make([]pdfUnifiedEvidenceWindow, 0, maxWindows)
	addWindow := func(candidate pdfUnifiedEvidenceWindow) bool {
		for _, existing := range selected {
			if candidate.Start < existing.End && existing.Start < candidate.End {
				return false
			}
		}
		selected = append(selected, candidate)
		return true
	}
	if fieldCompare {
		bestIndex := -1
		bestRows := 0
		for index, candidate := range windows {
			rows := pdfUnifiedAlignedValueRowCount(candidate.Text)
			if rows > bestRows {
				bestIndex = index
				bestRows = rows
			}
		}
		if bestIndex >= 0 {
			addWindow(windows[bestIndex])
		}
	}
	for _, item := range ranked {
		if item.Matches <= 0 || item.Page < 1 || item.Page > len(windows) {
			continue
		}
		candidate := windows[item.Page-1]
		addWindow(candidate)
		if len(selected) == maxWindows {
			break
		}
	}
	if len(selected) == 0 {
		return truncateToolText(text, maxChars)
	}

	separator := "\n...\n"
	contentBudget := maxChars - runeLen(separator)*(len(selected)-1)
	if contentBudget <= 0 {
		return truncateToolText(selected[0].Text, maxChars)
	}
	perWindow := contentBudget / len(selected)
	parts := make([]string, 0, len(selected))
	for _, window := range selected {
		parts = append(parts, truncateToolText(window.Text, perWindow))
	}
	return truncateToolText(strings.Join(parts, separator), maxChars)
}

func filterPDFUnifiedFieldCompareEvidencePages(pages []int, focus pdfUnifiedDocumentFocus) []int {
	if len(pages) == 0 {
		return nil
	}
	blocked := make(map[int]struct{}, len(pages))
	for _, segment := range focus.Segments {
		switch segment.Kind {
		case pdfUnifiedSegmentLogisticsDoc, pdfUnifiedSegmentCoverNotice:
			for _, page := range segment.Pages {
				blocked[page] = struct{}{}
			}
		}
	}
	out := make([]int, 0, len(pages))
	for _, page := range pages {
		if _, blockedPage := blocked[page]; blockedPage {
			continue
		}
		out = append(out, page)
	}
	if len(out) == 0 {
		return pages
	}
	return out
}

func filterPDFUnifiedPayloadEvidencePages(document pdfUnifiedDocumentInfo) []pdfAnalyzePageItem {
	if len(document.EvidencePages) == 0 || len(document.Segments) == 0 {
		return document.EvidencePages
	}
	blocked := make(map[int]struct{}, len(document.EvidencePages))
	for _, segment := range document.Segments {
		switch segment.Kind {
		case pdfUnifiedSegmentLogisticsDoc, pdfUnifiedSegmentCoverNotice:
			for _, page := range segment.Pages {
				blocked[page] = struct{}{}
			}
		}
	}
	filtered := make([]pdfAnalyzePageItem, 0, len(document.EvidencePages))
	for _, item := range document.EvidencePages {
		if _, ok := blocked[item.Page]; ok {
			continue
		}
		filtered = append(filtered, item)
	}
	if len(filtered) == 0 {
		return document.EvidencePages
	}
	return filtered
}

func pdfAnalyzePageItemsForPages(pageMap []pdfAnalyzePageItem, pages []int, limit int) []pdfAnalyzePageItem {
	if len(pageMap) == 0 || len(pages) == 0 || limit <= 0 {
		return nil
	}
	pageSet := make(map[int]struct{}, len(pages))
	for _, page := range pages {
		if page <= 0 {
			continue
		}
		pageSet[page] = struct{}{}
	}
	out := make([]pdfAnalyzePageItem, 0, minInt(limit, len(pageSet)))
	for _, item := range pageMap {
		if _, ok := pageSet[item.Page]; !ok {
			continue
		}
		if strings.TrimSpace(item.Excerpt) == "" {
			continue
		}
		out = append(out, item)
		if len(out) >= limit {
			break
		}
	}
	return out
}
