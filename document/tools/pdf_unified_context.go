package tools

import (
	"fmt"
	"strings"
)

func buildPDFUnifiedPromptChunks(prompt string, documents []pdfUnifiedDocumentArtifacts, focuses []pdfUnifiedDocumentFocus, evidencePages [][]pdfAnalyzePageItem, focusEnabled bool, focusQueryClass string, focusConfidence string, maxContextChars int) []string {
	share := maxContextChars
	if len(documents) > 0 {
		share = maxInt(2_000, maxContextChars/len(documents))
	}
	chunks := []string{
		fmt.Sprintf("User request:\n%s", prompt),
		"Use only the provided PDF evidence. If a claim is not supported by the PDF content, say so.",
		"Match every requested field's label, reporting period, entity, and scope exactly. Do not replace it with a related total, subtotal, adjusted value, or narrower/broader metric; mark the field missing or uncertain when exact evidence is absent. If a nearby row is a plausible alternative, name it briefly and explain why it is not the requested field.",
		"For each resolved field, reproduce the exact source row or label verbatim so downstream consumers can distinguish duplicate values attached to different scopes.",
		"When the same value appears on multiple pages or labels, cite the page with the exact requested label, not a later roll-forward, equity, summary, or component row. An explicit total row matching the requested entity and scope remains authoritative even when adjacent rows decompose it into narrower components.",
		"Copy numeric literals exactly from the source. Unless the request asks for a conversion, do not recompute, round, or restate a resolved value with altered digits in supplemental prose.",
		"For tables, preserve the association between row labels, column headers, and values. If extraction separates value columns from their labels, align them by the table headers and row order; do not relabel a prior-period value as the requested period.",
		"For every material claim, cite the supporting page inline as [pN] or [pp.N-M]. When comparing multiple PDFs, include the PDF label together with the page citation.",
	}
	useFocusedContext := shouldUsePDFUnifiedFocusedContext(documents, focuses, focusEnabled, focusQueryClass, focusConfidence)
	if useFocusedContext {
		chunks = append(chunks, fmt.Sprintf("Focus mode is active for %s queries. Prioritize each PDF's primary segment before supporting segments.", focusQueryClass))
		if focusQueryClass == pdfUnifiedQueryClassFieldCompare {
			chunks = append(chunks, "For 主体/金额/日期/结论 style comparisons, use each PDF's primary business segment as the default source of truth. Treat logistics/waybill, cover/notice, and other auxiliary segments as supplemental evidence only, and only use them when the primary segment lacks the field.")
		}
	}
	for idx, document := range documents {
		text := strings.TrimSpace(buildPDFUnifiedPageMarkedQueryText(prompt, document.TextResult.Pages))
		if useFocusedContext && idx < len(focuses) {
			text = buildPDFUnifiedFocusedDocumentText(prompt, document, focuses[idx], focusQueryClass, share)
		}
		if trimmed, changed := trimToMaxChars(text, share); changed {
			text = trimmed
		}
		header := fmt.Sprintf("[PDF %d] %s", idx+1, document.DisplayPath)
		if len(document.SelectedPages) > 0 {
			header = fmt.Sprintf("%s\nSelected pages: %s", header, formatPDFPageSelection(document.SelectedPages))
		}
		if idx < len(evidencePages) && len(evidencePages[idx]) > 0 {
			lines := make([]string, 0, len(evidencePages[idx]))
			excerptChars := 180
			if strings.TrimSpace(focusQueryClass) == pdfUnifiedQueryClassFieldCompare {
				excerptChars = 420
			}
			for _, item := range evidencePages[idx] {
				excerpt := strings.TrimSpace(item.Excerpt)
				if excerpt == "" {
					continue
				}
				lines = append(lines, fmt.Sprintf("- [p%d] %s", item.Page, truncateToolText(excerpt, excerptChars)))
			}
			if len(lines) > 0 {
				header += "\nCitable page evidence:\n" + strings.Join(lines, "\n")
			}
		}
		if text == "" {
			chunks = append(chunks, header+"\nNo reliable embedded text was extracted from the selected pages.")
			continue
		}
		chunks = append(chunks, header+"\nExtracted text:\n"+text)
	}
	return chunks
}

func buildPDFUnifiedPageMarkedQueryText(prompt string, pages []PDFPageText) string {
	if len(pages) == 0 {
		return ""
	}
	ordered := orderPDFUnifiedPageTextsByQuery(prompt, pages)
	parts := make([]string, 0, len(ordered))
	for _, page := range ordered {
		text := strings.TrimSpace(page.Text)
		if text == "" {
			continue
		}
		parts = append(parts, fmt.Sprintf("[p%d]\n%s", page.Page, text))
	}
	return strings.Join(parts, "\n\n")
}

func orderPDFUnifiedPageTextsByQuery(prompt string, pages []PDFPageText) []PDFPageText {
	if len(pages) < 2 {
		return append([]PDFPageText(nil), pages...)
	}
	ranked, matched := rankPDFUnifiedPages(prompt, pages)
	if !matched {
		return append([]PDFPageText(nil), pages...)
	}
	byPage := make(map[int]PDFPageText, len(pages))
	for _, page := range pages {
		byPage[page.Page] = page
	}
	ordered := make([]PDFPageText, 0, len(pages))
	seen := make(map[int]struct{}, len(pages))
	add := func(pageNum int) {
		if _, ok := seen[pageNum]; ok {
			return
		}
		page, ok := byPage[pageNum]
		if !ok {
			return
		}
		seen[pageNum] = struct{}{}
		ordered = append(ordered, page)
	}
	if classifyPDFUnifiedQuery(prompt) == pdfUnifiedQueryClassFieldCompare {
		aligned := rankPDFUnifiedAlignedTablePages(prompt, pages)
		aligned = aligned[:minInt(fieldPDFUnifiedAlignedAnchors, len(aligned))]
		for _, item := range aligned {
			add(item.Page)
			add(item.Page + 1)
		}
	}
	anchors := make([]pdfUnifiedPageRelevance, 0, len(ranked))
	for _, item := range ranked {
		if item.Score <= 0 {
			break
		}
		anchors = append(anchors, item)
	}
	anchorBatch := pdfUnifiedEvidenceAnchorBatch(classifyPDFUnifiedQuery(prompt), len(anchors))
	for _, item := range anchors[:anchorBatch] {
		add(item.Page)
	}
	for _, item := range anchors[:anchorBatch] {
		add(item.Page + 1)
	}
	for _, item := range anchors[:anchorBatch] {
		add(item.Page - 1)
	}
	for _, item := range anchors[anchorBatch:] {
		add(item.Page)
	}
	for _, item := range ranked {
		add(item.Page)
	}
	return ordered
}

func pdfUnifiedEvidenceAnchorBatch(queryClass string, rankedCount int) int {
	limit := 3
	if strings.TrimSpace(queryClass) == pdfUnifiedQueryClassFieldCompare {
		limit = 6
	}
	return minInt(limit, rankedCount)
}

func shouldUsePDFUnifiedFocusedContext(documents []pdfUnifiedDocumentArtifacts, focuses []pdfUnifiedDocumentFocus, focusEnabled bool, focusQueryClass string, focusConfidence string) bool {
	if !focusEnabled || strings.TrimSpace(focusConfidence) == "low" {
		return false
	}
	if len(documents) > 1 {
		return true
	}
	for _, focus := range focuses {
		if focus.Mixed {
			return true
		}
	}
	return focusQueryClass == pdfUnifiedQueryClassChartSummary
}

func buildPDFUnifiedFocusedDocumentText(prompt string, document pdfUnifiedDocumentArtifacts, focus pdfUnifiedDocumentFocus, queryClass string, maxChars int) string {
	if focus.Primary == nil {
		return strings.TrimSpace(buildPDFUnifiedPageMarkedQueryText(prompt, document.TextResult.Pages))
	}
	parts := make([]string, 0, 3)
	contentBudget := maxChars
	if strings.TrimSpace(queryClass) == pdfUnifiedQueryClassFieldCompare {
		queryEvidence := strings.TrimSpace(buildPDFUnifiedPageMarkedQueryText(prompt, document.TextResult.Pages))
		queryEvidence = truncateToolText(queryEvidence, maxInt(800, int(float64(maxChars)*0.85)))
		if queryEvidence != "" {
			parts = append(parts, "Query-relevant page evidence:\n"+queryEvidence)
			contentBudget = maxInt(400, maxChars-runeLen(strings.Join(parts, "\n\n")))
		}
	}
	primary := strings.TrimSpace(buildPDFUnifiedPageMarkedQueryText(prompt, pdfUnifiedPageTextsForPages(document.TextResult.Pages, focus.Primary.Pages)))
	if primary == "" {
		primary = strings.TrimSpace(focus.Primary.text)
	}
	if primary == "" {
		primary = strings.TrimSpace(buildPDFUnifiedPageMarkedQueryText(prompt, document.TextResult.Pages))
	}
	if primary == "" {
		return ""
	}
	primaryBudget := contentBudget
	if len(focus.Supporting) > 0 {
		primaryBudget = int(float64(contentBudget) * 0.65)
	}
	primary = truncateToolText(primary, primaryBudget)
	parts = append(parts, fmt.Sprintf(
		"Primary segment [%s] pages %s:\n%s",
		focus.Primary.Kind,
		formatPDFPageSelection(focus.Primary.Pages),
		primary,
	))
	supporting := selectPDFUnifiedSupportingSegmentsForPrompt(focus, queryClass)
	if len(supporting) == 0 {
		return strings.Join(parts, "\n\n")
	}
	remaining := maxChars - runeLen(strings.Join(parts, "\n\n"))
	if remaining <= 0 {
		return strings.Join(parts, "\n\n")
	}
	perSupporting := maxInt(240, remaining/len(supporting))
	for _, segment := range supporting {
		text := strings.TrimSpace(buildPDFUnifiedPageMarkedQueryText(prompt, pdfUnifiedPageTextsForPages(document.TextResult.Pages, segment.Pages)))
		if text == "" {
			text = strings.TrimSpace(segment.text)
		}
		text = truncateToolText(text, perSupporting)
		if text == "" {
			continue
		}
		parts = append(parts, fmt.Sprintf("Supporting segment [%s] pages %s:\n%s", segment.Kind, formatPDFPageSelection(segment.Pages), text))
	}
	return strings.Join(parts, "\n\n")
}

func pdfUnifiedPageTextsForPages(pageTexts []PDFPageText, pages []int) []PDFPageText {
	if len(pageTexts) == 0 || len(pages) == 0 {
		return nil
	}
	selected := make(map[int]struct{}, len(pages))
	for _, page := range pages {
		selected[page] = struct{}{}
	}
	out := make([]PDFPageText, 0, len(selected))
	for _, page := range pageTexts {
		if _, ok := selected[page.Page]; ok {
			out = append(out, page)
		}
	}
	return out
}

func selectPDFUnifiedSupportingSegmentsForPrompt(focus pdfUnifiedDocumentFocus, queryClass string) []pdfUnifiedSegment {
	if len(focus.Supporting) == 0 {
		return nil
	}
	if strings.TrimSpace(queryClass) != pdfUnifiedQueryClassFieldCompare {
		return append([]pdfUnifiedSegment(nil), focus.Supporting...)
	}
	if focus.Primary == nil {
		return append([]pdfUnifiedSegment(nil), focus.Supporting...)
	}
	out := make([]pdfUnifiedSegment, 0, len(focus.Supporting))
	for _, segment := range focus.Supporting {
		switch focus.Primary.Kind {
		case pdfUnifiedSegmentBusinessDoc:
			switch segment.Kind {
			case pdfUnifiedSegmentSupportingDoc, pdfUnifiedSegmentSignatureStamp:
				out = append(out, segment)
			}
		case pdfUnifiedSegmentSupportingDoc:
			switch segment.Kind {
			case pdfUnifiedSegmentBusinessDoc, pdfUnifiedSegmentSignatureStamp:
				out = append(out, segment)
			}
		default:
			switch segment.Kind {
			case pdfUnifiedSegmentCoverNotice:
				continue
			default:
				out = append(out, segment)
			}
		}
	}
	return out
}

func buildPDFUnifiedDocumentEvidencePages(prompt string, documents []pdfUnifiedDocumentArtifacts, focuses []pdfUnifiedDocumentFocus, focusEnabled bool, queryClass string, limit int) [][]pdfAnalyzePageItem {
	if len(documents) == 0 || limit <= 0 {
		return nil
	}
	out := make([][]pdfAnalyzePageItem, 0, len(documents))
	for idx, document := range documents {
		var focus pdfUnifiedDocumentFocus
		if idx < len(focuses) {
			focus = focuses[idx]
		}
		out = append(out, buildPDFUnifiedEvidencePagesForDocument(prompt, document, focus, focusEnabled, queryClass, limit))
	}
	return out
}

func buildPDFUnifiedEvidencePagesForDocument(prompt string, document pdfUnifiedDocumentArtifacts, focus pdfUnifiedDocumentFocus, focusEnabled bool, queryClass string, limit int) []pdfAnalyzePageItem {
	if len(document.PageMap) == 0 || limit <= 0 {
		return nil
	}
	pages := make([]int, 0, limit)
	seen := make(map[int]struct{}, limit)
	add := func(page int) {
		if page <= 0 || len(pages) >= limit {
			return
		}
		if _, ok := seen[page]; ok {
			return
		}
		seen[page] = struct{}{}
		pages = append(pages, page)
	}
	if strings.TrimSpace(queryClass) == pdfUnifiedQueryClassFieldCompare {
		aligned := rankPDFUnifiedAlignedTablePages(prompt, document.TextResult.Pages)
		aligned = aligned[:minInt(fieldPDFUnifiedAlignedAnchors, len(aligned))]
		for _, item := range aligned {
			add(item.Page)
			add(item.Page + 1)
		}
	}
	if ranked, matched := rankPDFUnifiedPages(prompt, document.TextResult.Pages); matched {
		anchors := make([]pdfUnifiedPageRelevance, 0, len(ranked))
		for _, item := range ranked {
			if item.Score <= 0 {
				break
			}
			anchors = append(anchors, item)
		}
		anchorBatch := pdfUnifiedEvidenceAnchorBatch(queryClass, len(anchors))
		for _, item := range anchors[:anchorBatch] {
			add(item.Page)
		}
		for _, item := range anchors[:anchorBatch] {
			add(item.Page + 1)
		}
		for _, item := range anchors[:anchorBatch] {
			add(item.Page - 1)
		}
		for _, item := range anchors[anchorBatch:] {
			add(item.Page)
		}
	}
	if len(document.SelectedPages) > 0 {
		for _, page := range document.SelectedPages {
			add(page)
		}
	}
	if len(pages) < limit && focusEnabled && focus.Primary != nil {
		for _, page := range focus.Primary.Pages {
			add(page)
		}
		supporting := selectPDFUnifiedSupportingSegmentsForPrompt(focus, queryClass)
		if strings.TrimSpace(queryClass) == pdfUnifiedQueryClassChartSummary && len(supporting) == 0 {
			supporting = append([]pdfUnifiedSegment(nil), focus.Supporting...)
		}
		for _, segment := range supporting {
			for _, page := range segment.Pages {
				add(page)
			}
		}
		if strings.TrimSpace(queryClass) == pdfUnifiedQueryClassFieldCompare && len(pages) > 0 {
			pages = filterPDFUnifiedFieldCompareEvidencePages(pages, focus)
			items := pdfAnalyzePageItemsForPages(document.PageMap, pages, limit)
			return enrichPDFUnifiedEvidenceExcerpts(prompt, document.TextResult.Pages, items)
		}
	}
	if len(pages) < limit {
		candidates := append([]int(nil), document.VisualPages...)
		if len(candidates) == 0 {
			for _, item := range topPDFAnalyzePages(document.PageMap, limit) {
				candidates = append(candidates, item.Page)
			}
		}
		for _, page := range candidates {
			add(page)
		}
	}
	if len(pages) == 0 {
		return nil
	}
	items := pdfAnalyzePageItemsForPages(document.PageMap, pages, limit)
	return enrichPDFUnifiedEvidenceExcerpts(prompt, document.TextResult.Pages, items)
}

func enrichPDFUnifiedEvidenceExcerpts(prompt string, pageTexts []PDFPageText, items []pdfAnalyzePageItem) []pdfAnalyzePageItem {
	if len(pageTexts) == 0 || len(items) == 0 {
		return items
	}
	textByPage := make(map[int]string, len(pageTexts))
	for _, page := range pageTexts {
		textByPage[page.Page] = page.Text
	}
	out := append([]pdfAnalyzePageItem(nil), items...)
	for idx := range out {
		excerpt := buildPDFUnifiedQueryEvidenceExcerpt(prompt, textByPage[out[idx].Page], defaultPDFUnifiedEvidenceChars)
		if strings.TrimSpace(excerpt) != "" {
			out[idx].Excerpt = excerpt
		}
	}
	return out
}
