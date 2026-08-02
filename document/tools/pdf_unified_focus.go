package tools

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
)

const (
	pdfUnifiedQueryClassGeneric      = "generic"
	pdfUnifiedQueryClassFieldCompare = "field_compare"
	pdfUnifiedQueryClassChartSummary = "chart_summary"

	pdfUnifiedContentLayerBody      = "body"
	pdfUnifiedContentLayerFurniture = "furniture"

	pdfUnifiedStructureBlockText      = "text_block"
	pdfUnifiedStructureBlockChart     = "chart_table_region"
	pdfUnifiedStructureBlockKeyValue  = "key_value_region"
	pdfUnifiedStructureBlockSignature = "signature_region"

	pdfUnifiedSegmentBusinessDoc    = "business_doc"
	pdfUnifiedSegmentLogisticsDoc   = "logistics_doc"
	pdfUnifiedSegmentSupportingDoc  = "supporting_doc"
	pdfUnifiedSegmentCoverNotice    = "cover_or_notice"
	pdfUnifiedSegmentSignatureStamp = "signature_or_stamp"
	pdfUnifiedSegmentChartLayout    = "chart_table_layout"
)

func buildPDFUnifiedDocumentFocuses(prompt string, documents []pdfUnifiedDocumentArtifacts) (string, []string, string, []pdfUnifiedDocumentFocus) {
	queryClass := classifyPDFUnifiedQuery(prompt)
	focuses := make([]pdfUnifiedDocumentFocus, 0, len(documents))
	hasMixed := false
	hasOCRSensitive := false
	hasVisualLayout := false
	for _, document := range documents {
		focus := buildPDFUnifiedDocumentFocus(document, queryClass)
		focuses = append(focuses, focus)
		if focus.Mixed {
			hasMixed = true
		}
		if document.AnalysisPlan.NeedsOCR || document.DocumentProfile.LikelyScanned {
			hasOCRSensitive = true
		}
		if document.AnalysisPlan.NeedsVision || document.MediaProfile.LikelyGraphicDoc || document.MediaProfile.LikelySlideDeck {
			hasVisualLayout = true
		}
	}

	reasons := make([]string, 0, 4)
	if queryClass != pdfUnifiedQueryClassGeneric {
		reasons = append(reasons, "query_class_"+queryClass)
	}
	if len(documents) > 1 {
		reasons = append(reasons, "multi_document")
	}
	if hasMixed {
		reasons = append(reasons, "mixed_subdocuments")
	}
	if queryClass == pdfUnifiedQueryClassFieldCompare && hasOCRSensitive {
		reasons = append(reasons, "ocr_sensitive")
	}
	if queryClass == pdfUnifiedQueryClassChartSummary && hasVisualLayout {
		reasons = append(reasons, "visual_layout")
	}
	if len(reasons) == 0 {
		return "", nil, "", focuses
	}
	return queryClass, reasons, computePDFUnifiedFocusConfidence(prompt, queryClass, reasons, documents, focuses), focuses
}

func buildPDFUnifiedFocusSummary(prompt string, documents []pdfUnifiedDocumentArtifacts) pdfUnifiedFocusSummary {
	queryClass, reasonCodes, confidence, focuses := buildPDFUnifiedDocumentFocuses(prompt, documents)
	return pdfUnifiedFocusSummary{
		Enabled:     len(reasonCodes) > 0,
		QueryClass:  queryClass,
		ReasonCodes: append([]string(nil), reasonCodes...),
		Confidence:  confidence,
		Documents:   append([]pdfUnifiedDocumentFocus(nil), focuses...),
	}
}

func buildPDFUnifiedFocusSummaryFromAnalysisArtifacts(prompt string, artifactsList []pdfAnalysisArtifacts) pdfUnifiedFocusSummary {
	if len(artifactsList) == 0 {
		return pdfUnifiedFocusSummary{}
	}
	documents := make([]pdfUnifiedDocumentArtifacts, 0, len(artifactsList))
	for _, artifacts := range artifactsList {
		documents = append(documents, pdfUnifiedDocumentArtifacts{
			Path:            artifacts.Path,
			DisplayPath:     artifacts.DisplayPath,
			Metadata:        artifacts.Metadata,
			TextResult:      artifacts.TextResult,
			BackendStatus:   artifacts.BackendStatus,
			PageMap:         artifacts.PageMap,
			StructureItems:  append([]pdfUnifiedStructureItem(nil), artifacts.StructureItems...),
			DocumentProfile: artifacts.DocumentProfile,
			MediaProfile:    artifacts.MediaProfile,
			AnalysisPlan:    artifacts.AnalysisPlan,
		})
	}
	return buildPDFUnifiedFocusSummary(prompt, documents)
}

func classifyPDFUnifiedQuery(prompt string) string {
	normalized := strings.ToLower(strings.TrimSpace(prompt))
	if normalized == "" {
		return pdfUnifiedQueryClassGeneric
	}
	strongChartHints := []string{
		"图表", "图示", "图像", "图片", "流程图", "柱状图", "折线图", "饼图",
		"chart", "diagram", "graph", "trend", "visualization", "可视化",
	}
	for _, hint := range strongChartHints {
		if strings.Contains(normalized, hint) {
			return pdfUnifiedQueryClassChartSummary
		}
	}
	fieldHints := []string{
		"比较", "差异", "对比", "主体", "金额", "日期", "甲方", "乙方",
		"签署", "合同", "期限", "违约", "提取", "抽取", "字段", "确认", "核对", "验证", "明确",
		"party", "amount", "date", "compare", "difference", "extract", "confirm", "verify",
	}
	for _, hint := range fieldHints {
		if strings.Contains(normalized, hint) {
			return pdfUnifiedQueryClassFieldCompare
		}
	}
	weakChartHints := []string{
		"图", "表格", "布局", "table",
	}
	for _, hint := range weakChartHints {
		if strings.Contains(normalized, hint) {
			return pdfUnifiedQueryClassChartSummary
		}
	}
	return pdfUnifiedQueryClassGeneric
}

func computePDFUnifiedFocusConfidence(prompt string, queryClass string, reasonCodes []string, documents []pdfUnifiedDocumentArtifacts, focuses []pdfUnifiedDocumentFocus) string {
	if len(reasonCodes) == 0 {
		return ""
	}
	if queryClass == pdfUnifiedQueryClassFieldCompare && len(documents) > 0 {
		strongEvidence := 0
		for _, document := range documents {
			if hasStrongPDFUnifiedQueryEvidence(prompt, document) {
				strongEvidence++
			}
		}
		if strongEvidence == len(documents) {
			return "high"
		}
		if strongEvidence > 0 {
			return "medium"
		}
	}
	for _, focus := range focuses {
		if focus.Primary == nil || strings.TrimSpace(focus.Primary.Confidence) == "" || focus.Primary.Confidence == "low" {
			return "low"
		}
	}
	for _, code := range reasonCodes {
		if code == "mixed_subdocuments" || code == "multi_document" {
			return "high"
		}
	}
	return "medium"
}

func hasStrongPDFUnifiedQueryEvidence(prompt string, document pdfUnifiedDocumentArtifacts) bool {
	if !document.BackendStatus.LayoutPreserved || len(document.TextResult.Pages) == 0 {
		return false
	}
	switch document.SelectionStrategy {
	case pdfUnifiedSelectionAll, pdfUnifiedSelectionExplicit, pdfUnifiedSelectionQuery:
	default:
		return false
	}
	ranked, matched := rankPDFUnifiedPages(prompt, document.TextResult.Pages)
	if !matched {
		return false
	}
	for _, item := range ranked {
		if item.Score <= 0 {
			break
		}
		if item.Matches > 1 {
			return true
		}
	}
	return false
}

func buildPDFUnifiedDocumentFocus(document pdfUnifiedDocumentArtifacts, queryClass string) pdfUnifiedDocumentFocus {
	segments := buildPDFUnifiedSegments(document)
	if len(segments) == 0 {
		return pdfUnifiedDocumentFocus{}
	}
	ordered := append([]pdfUnifiedSegment(nil), segments...)
	sort.SliceStable(ordered, func(i, j int) bool {
		left := pdfUnifiedSegmentPriority(queryClass, ordered[i].Kind)
		right := pdfUnifiedSegmentPriority(queryClass, ordered[j].Kind)
		if left != right {
			return left > right
		}
		leftConfidence := pdfUnifiedConfidenceWeight(ordered[i].Confidence)
		rightConfidence := pdfUnifiedConfidenceWeight(ordered[j].Confidence)
		if leftConfidence != rightConfidence {
			return leftConfidence > rightConfidence
		}
		if len(ordered[i].text) != len(ordered[j].text) {
			return len(ordered[i].text) > len(ordered[j].text)
		}
		return ordered[i].PageStart < ordered[j].PageStart
	})

	primary := ordered[0]
	supporting := make([]pdfUnifiedSegment, 0, 2)
	mixed := pdfUnifiedHasMixedSegments(segments)
	basePriority := pdfUnifiedSegmentPriority(queryClass, primary.Kind)
	for _, candidate := range ordered[1:] {
		priority := pdfUnifiedSegmentPriority(queryClass, candidate.Kind)
		if len(supporting) >= 2 {
			break
		}
		if priority >= basePriority-30 ||
			(primary.Kind == pdfUnifiedSegmentBusinessDoc && (candidate.Kind == pdfUnifiedSegmentSupportingDoc || candidate.Kind == pdfUnifiedSegmentSignatureStamp)) ||
			(mixed && candidate.Kind != primary.Kind) {
			supporting = append(supporting, candidate)
		}
	}
	primaryCopy := primary
	return pdfUnifiedDocumentFocus{
		Segments:   segments,
		Primary:    &primaryCopy,
		Supporting: supporting,
		Mixed:      mixed,
	}
}

func buildPDFUnifiedSegments(document pdfUnifiedDocumentArtifacts) []pdfUnifiedSegment {
	structureItems := document.StructureItems
	if len(structureItems) == 0 {
		structureItems = buildPDFUnifiedStructureItems(document.TextResult.Pages, document.PageMap, document.Metadata.PageCount, document.MediaProfile)
	}
	if len(structureItems) > 0 {
		return buildPDFUnifiedSegmentsFromStructure(structureItems)
	}
	if len(document.PageMap) == 0 {
		return nil
	}
	pageTexts := make(map[int]string, len(document.TextResult.Pages))
	for _, page := range document.TextResult.Pages {
		pageTexts[page.Page] = strings.TrimSpace(page.Text)
	}
	graphicHeavy := make(map[int]struct{}, len(document.MediaProfile.GraphicHeavyPages))
	for _, page := range document.MediaProfile.GraphicHeavyPages {
		graphicHeavy[page] = struct{}{}
	}
	segments := make([]pdfUnifiedSegment, 0, len(document.PageMap))
	for _, page := range document.PageMap {
		text := pageTexts[page.Page]
		kind, confidence, signalCodes, anchors := classifyPDFUnifiedPageSegment(page, text, document.Metadata.PageCount, document.MediaProfile, graphicHeavy)
		excerpt := strings.TrimSpace(page.Excerpt)
		if excerpt == "" {
			excerpt = truncateToolText(text, 240)
		}
		segment := pdfUnifiedSegment{
			Kind:        kind,
			Pages:       []int{page.Page},
			PageStart:   page.Page,
			PageEnd:     page.Page,
			Confidence:  confidence,
			Anchors:     anchors,
			Excerpt:     excerpt,
			SignalCodes: signalCodes,
			text:        text,
		}
		if len(segments) > 0 {
			last := &segments[len(segments)-1]
			if last.Kind == segment.Kind && last.PageEnd+1 == segment.PageStart {
				last.Pages = append(last.Pages, segment.Pages...)
				last.PageEnd = segment.PageEnd
				last.Confidence = strongerPDFUnifiedConfidence(last.Confidence, segment.Confidence)
				last.Anchors = appendPDFUnifiedStringSet(last.Anchors, segment.Anchors...)
				last.SignalCodes = appendPDFUnifiedStringSet(last.SignalCodes, segment.SignalCodes...)
				last.Excerpt = joinPDFUnifiedSegmentExcerpt(last.Excerpt, segment.Excerpt, 240)
				last.text = joinPDFUnifiedSegmentText(last.text, segment.text)
				continue
			}
		}
		segments = append(segments, segment)
	}
	for idx := range segments {
		segments[idx].ID = fmt.Sprintf("seg-%d", idx+1)
	}
	return segments
}

func buildPDFUnifiedSegmentsFromStructure(items []pdfUnifiedStructureItem) []pdfUnifiedSegment {
	if len(items) == 0 {
		return nil
	}
	candidates := make([]pdfUnifiedStructureItem, 0, len(items))
	for _, item := range items {
		if item.ContentLayer != pdfUnifiedContentLayerBody {
			continue
		}
		if strings.TrimSpace(item.Role) == "" || strings.TrimSpace(item.text) == "" {
			continue
		}
		candidates = append(candidates, item)
	}
	if len(candidates) == 0 {
		for _, item := range items {
			if strings.TrimSpace(item.Role) == "" || strings.TrimSpace(item.text) == "" {
				continue
			}
			candidates = append(candidates, item)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	segments := make([]pdfUnifiedSegment, 0, len(candidates))
	for _, item := range candidates {
		segment := pdfUnifiedSegment{
			Kind:        item.Role,
			Pages:       []int{item.Page},
			PageStart:   item.Page,
			PageEnd:     item.Page,
			Confidence:  item.Confidence,
			Anchors:     append([]string(nil), item.Anchors...),
			Excerpt:     item.Excerpt,
			SignalCodes: append([]string(nil), item.SignalCodes...),
			text:        item.text,
		}
		if len(segments) > 0 {
			last := &segments[len(segments)-1]
			if last.Kind == segment.Kind && last.PageEnd+1 == segment.PageStart {
				last.Pages = append(last.Pages, segment.Pages...)
				last.PageEnd = segment.PageEnd
				last.Confidence = strongerPDFUnifiedConfidence(last.Confidence, segment.Confidence)
				last.Anchors = appendPDFUnifiedStringSet(last.Anchors, segment.Anchors...)
				last.SignalCodes = appendPDFUnifiedStringSet(last.SignalCodes, segment.SignalCodes...)
				last.Excerpt = joinPDFUnifiedSegmentExcerpt(last.Excerpt, segment.Excerpt, 240)
				last.text = joinPDFUnifiedSegmentText(last.text, segment.text)
				continue
			}
		}
		segments = append(segments, segment)
	}
	for idx := range segments {
		segments[idx].ID = fmt.Sprintf("seg-%d", idx+1)
	}
	return segments
}

func buildPDFUnifiedStructureItems(pageTexts []PDFPageText, pageMap []pdfAnalyzePageItem, pageCount int, media pdfMediaProfile) []pdfUnifiedStructureItem {
	if len(pageMap) == 0 {
		return nil
	}
	pageTextByPage := make(map[int]string, len(pageTexts))
	for _, page := range pageTexts {
		pageTextByPage[page.Page] = strings.TrimSpace(page.Text)
	}
	repeatedHeaders := buildPDFUnifiedRepeatedEdgeLineSet(pageTexts, true)
	repeatedFooters := buildPDFUnifiedRepeatedEdgeLineSet(pageTexts, false)
	graphicHeavy := make(map[int]struct{}, len(media.GraphicHeavyPages))
	for _, page := range media.GraphicHeavyPages {
		graphicHeavy[page] = struct{}{}
	}
	items := make([]pdfUnifiedStructureItem, 0, len(pageMap)*2)
	for _, page := range pageMap {
		text := pageTextByPage[page.Page]
		lines := splitPDFUnifiedPageLines(text)
		headerText, bodyText, footerText := splitPDFUnifiedPageStructure(lines, repeatedHeaders, repeatedFooters)
		if headerText != "" {
			items = append(items, buildPDFUnifiedFurnitureItem(page.Page, headerText, "repeated_header"))
		}
		bodyText = strings.TrimSpace(bodyText)
		if bodyText == "" {
			bodyText = strings.TrimSpace(text)
		}
		if bodyText != "" {
			blockKind := classifyPDFUnifiedStructureBlockKind(page, bodyText, pageCount, media, graphicHeavy)
			kind, confidence, signalCodes, anchors := classifyPDFUnifiedPageSegment(page, bodyText, pageCount, media, graphicHeavy)
			signalCodes = appendPDFUnifiedStringSet(signalCodes, "block_kind_"+blockKind, "content_layer_"+pdfUnifiedContentLayerBody)
			items = append(items, pdfUnifiedStructureItem{
				Page:         page.Page,
				ContentLayer: pdfUnifiedContentLayerBody,
				BlockKind:    blockKind,
				Role:         kind,
				Confidence:   confidence,
				Anchors:      anchors,
				Excerpt:      truncateToolText(bodyText, 240),
				SignalCodes:  signalCodes,
				text:         bodyText,
			})
		}
		if footerText != "" {
			items = append(items, buildPDFUnifiedFurnitureItem(page.Page, footerText, "repeated_footer"))
		}
	}
	for idx := range items {
		items[idx].ID = fmt.Sprintf("item-%d", idx+1)
	}
	return items
}

func buildPDFUnifiedFurnitureItem(page int, text string, reasonCode string) pdfUnifiedStructureItem {
	text = strings.TrimSpace(text)
	return pdfUnifiedStructureItem{
		Page:         page,
		ContentLayer: pdfUnifiedContentLayerFurniture,
		BlockKind:    pdfUnifiedStructureBlockText,
		Role:         pdfUnifiedSegmentCoverNotice,
		Confidence:   "medium",
		Anchors:      prependPDFUnifiedAnchor(nil, firstPDFUnifiedAnchorLine(text)),
		Excerpt:      truncateToolText(text, 160),
		SignalCodes:  []string{"content_layer_" + pdfUnifiedContentLayerFurniture, reasonCode},
		text:         text,
	}
}

func buildPDFUnifiedRepeatedEdgeLineSet(pageTexts []PDFPageText, isHeader bool) map[string]struct{} {
	counts := map[string]int{}
	for _, page := range pageTexts {
		lines := splitPDFUnifiedPageLines(page.Text)
		if len(lines) == 0 {
			continue
		}
		line := lines[0]
		if !isHeader {
			line = lines[len(lines)-1]
		}
		normalized := normalizePDFUnifiedEdgeLine(line)
		if normalized == "" {
			continue
		}
		counts[normalized]++
	}
	out := map[string]struct{}{}
	for line, count := range counts {
		if count >= 2 {
			out[line] = struct{}{}
		}
	}
	return out
}

func splitPDFUnifiedPageLines(text string) []string {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		out = append(out, trimmed)
	}
	return out
}

func splitPDFUnifiedPageStructure(lines []string, repeatedHeaders map[string]struct{}, repeatedFooters map[string]struct{}) (string, string, string) {
	if len(lines) == 0 {
		return "", "", ""
	}
	start := 0
	end := len(lines)
	header := ""
	footer := ""
	if _, ok := repeatedHeaders[normalizePDFUnifiedEdgeLine(lines[0])]; ok {
		header = lines[0]
		start = 1
	}
	if end > start {
		lastIdx := end - 1
		if _, ok := repeatedFooters[normalizePDFUnifiedEdgeLine(lines[lastIdx])]; ok {
			footer = lines[lastIdx]
			end = lastIdx
		}
	}
	bodyLines := lines[start:end]
	if len(bodyLines) == 0 && (header != "" || footer != "") {
		return "", strings.Join(lines, "\n"), ""
	}
	return header, strings.Join(bodyLines, "\n"), footer
}

func normalizePDFUnifiedEdgeLine(line string) string {
	line = strings.TrimSpace(strings.ToLower(line))
	if line == "" {
		return ""
	}
	line = strings.Join(strings.Fields(line), " ")
	if len(line) < 6 || len(line) > 120 {
		return ""
	}
	alphaNum := 0
	letters := 0
	for _, r := range line {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			alphaNum++
		}
		if unicode.IsLetter(r) {
			letters++
		}
	}
	if alphaNum == 0 || letters == 0 {
		return ""
	}
	return line
}

func classifyPDFUnifiedStructureBlockKind(page pdfAnalyzePageItem, text string, pageCount int, media pdfMediaProfile, graphicHeavy map[int]struct{}) string {
	normalized := strings.ToLower(strings.TrimSpace(text))
	if len(pdfUnifiedKeywordHits(normalized, []string{
		"签章", "盖章", "签字", "签名", "授权代表", "法定代表人", "stamp", "seal",
	})) > 0 || (page.Page == pageCount && page.Chars <= 240) {
		return pdfUnifiedStructureBlockSignature
	}
	keyValueSignals := 0
	for _, line := range splitPDFUnifiedPageLines(text) {
		if strings.Contains(line, ":") || strings.Contains(line, "：") {
			keyValueSignals++
		}
	}
	keyValueSignals += len(pdfUnifiedKeywordHits(normalized, []string{
		"寄件人", "收件人", "单号", "运单", "到付", "金额", "日期", "甲方", "乙方", "公司",
	}))
	if keyValueSignals >= 2 {
		return pdfUnifiedStructureBlockKeyValue
	}
	if len(pdfUnifiedKeywordHits(normalized, []string{
		"图表", "chart", "diagram", "table", "趋势", "销量", "保有量", "同比",
	})) > 0 {
		return pdfUnifiedStructureBlockChart
	}
	if _, ok := graphicHeavy[page.Page]; ok {
		return pdfUnifiedStructureBlockChart
	}
	if media.LikelyGraphicDoc && page.Empty {
		return pdfUnifiedStructureBlockChart
	}
	return pdfUnifiedStructureBlockText
}

func classifyPDFUnifiedPageSegment(page pdfAnalyzePageItem, text string, pageCount int, media pdfMediaProfile, graphicHeavy map[int]struct{}) (string, string, []string, []string) {
	normalized := strings.ToLower(strings.TrimSpace(text + "\n" + page.Excerpt))
	scores := map[string]int{}
	codesByKind := map[string][]string{}
	anchors := make([]string, 0, 3)
	addScore := func(kind string, score int, code string) {
		if score <= 0 || strings.TrimSpace(kind) == "" {
			return
		}
		scores[kind] += score
		if strings.TrimSpace(code) != "" {
			codesByKind[kind] = appendPDFUnifiedStringSet(codesByKind[kind], code)
		}
	}
	addMatches := func(kind string, keywords []string, code string, base int) {
		hits := pdfUnifiedKeywordHits(normalized, keywords)
		if len(hits) == 0 {
			return
		}
		addScore(kind, base+len(hits), code)
		anchors = appendPDFUnifiedStringSet(anchors, hits...)
	}

	addMatches(pdfUnifiedSegmentBusinessDoc, []string{
		"合同", "协议", "询证函", "对账", "往来", "甲方", "乙方", "付款", "签署",
		"公司", "采购", "主协议", "master agreement", "isda",
	}, "keyword_business_doc", 3)
	addMatches(pdfUnifiedSegmentLogisticsDoc, []string{
		"运单", "快件", "物流", "顺丰", "收件人", "寄件人", "单号", "到付", "签收", "运费",
	}, "keyword_logistics_doc", 4)
	addMatches(pdfUnifiedSegmentSupportingDoc, []string{
		"附件", "附录", "附表", "appendix", "annex", "schedule", "confirmation", "备注", "说明",
	}, "keyword_supporting_doc", 2)
	addMatches(pdfUnifiedSegmentCoverNotice, []string{
		"目录", "contents", "封面", "首页", "概览", "notice", "摘要",
	}, "keyword_cover_notice", 2)
	addMatches(pdfUnifiedSegmentSignatureStamp, []string{
		"签章", "盖章", "签字", "签名", "签署日期", "授权代表", "法定代表人", "stamp", "seal",
	}, "keyword_signature_stamp", 3)
	addMatches(pdfUnifiedSegmentChartLayout, []string{
		"图表", "chart", "diagram", "同比", "趋势", "销量", "保有量", "table",
	}, "keyword_chart_layout", 2)

	if _, ok := graphicHeavy[page.Page]; ok {
		addScore(pdfUnifiedSegmentChartLayout, 5, "graphic_heavy_page")
	}
	if media.LikelyGraphicDoc && page.Chars <= 220 {
		addScore(pdfUnifiedSegmentChartLayout, 2, "graphic_doc_layout")
	}
	if page.Page == 1 && page.Chars <= 180 {
		addScore(pdfUnifiedSegmentCoverNotice, 2, "leading_short_page")
	}
	if page.Page == pageCount && page.Chars <= 240 {
		addScore(pdfUnifiedSegmentSignatureStamp, 2, "trailing_signature_page")
	}
	if page.Empty && media.LikelyGraphicDoc {
		addScore(pdfUnifiedSegmentChartLayout, 2, "empty_graphic_page")
	}
	if len(scores) == 0 {
		if page.Empty {
			if _, ok := graphicHeavy[page.Page]; ok || media.LikelyGraphicDoc {
				scores[pdfUnifiedSegmentChartLayout] = 1
			} else {
				scores[pdfUnifiedSegmentSupportingDoc] = 1
			}
		} else {
			scores[pdfUnifiedSegmentBusinessDoc] = 1
			codesByKind[pdfUnifiedSegmentBusinessDoc] = []string{"default_body_text"}
		}
	}

	kind, topScore, secondScore := selectPDFUnifiedSegmentKind(scores)
	confidence := "low"
	switch {
	case topScore >= 8 && topScore-secondScore >= 3:
		confidence = "high"
	case topScore >= 4:
		confidence = "medium"
	}
	anchors = prependPDFUnifiedAnchor(anchors, firstPDFUnifiedAnchorLine(text))
	return kind, confidence, codesByKind[kind], anchors
}

func selectPDFUnifiedSegmentKind(scores map[string]int) (string, int, int) {
	order := []string{
		pdfUnifiedSegmentBusinessDoc,
		pdfUnifiedSegmentLogisticsDoc,
		pdfUnifiedSegmentChartLayout,
		pdfUnifiedSegmentSupportingDoc,
		pdfUnifiedSegmentSignatureStamp,
		pdfUnifiedSegmentCoverNotice,
	}
	bestKind := pdfUnifiedSegmentBusinessDoc
	bestScore := -1
	secondScore := -1
	for _, kind := range order {
		score := scores[kind]
		if score > bestScore {
			secondScore = bestScore
			bestScore = score
			bestKind = kind
			continue
		}
		if score > secondScore {
			secondScore = score
		}
	}
	if bestScore < 0 {
		return pdfUnifiedSegmentBusinessDoc, 0, 0
	}
	if secondScore < 0 {
		secondScore = 0
	}
	return bestKind, bestScore, secondScore
}

func pdfUnifiedSegmentPriority(queryClass string, kind string) int {
	switch queryClass {
	case pdfUnifiedQueryClassFieldCompare:
		switch kind {
		case pdfUnifiedSegmentBusinessDoc:
			return 100
		case pdfUnifiedSegmentSupportingDoc:
			return 70
		case pdfUnifiedSegmentSignatureStamp:
			return 60
		case pdfUnifiedSegmentLogisticsDoc:
			return 30
		case pdfUnifiedSegmentChartLayout:
			return 25
		case pdfUnifiedSegmentCoverNotice:
			return 20
		}
	case pdfUnifiedQueryClassChartSummary:
		switch kind {
		case pdfUnifiedSegmentChartLayout:
			return 100
		case pdfUnifiedSegmentBusinessDoc:
			return 70
		case pdfUnifiedSegmentSupportingDoc:
			return 55
		case pdfUnifiedSegmentSignatureStamp:
			return 20
		case pdfUnifiedSegmentCoverNotice:
			return 20
		case pdfUnifiedSegmentLogisticsDoc:
			return 10
		}
	default:
		switch kind {
		case pdfUnifiedSegmentBusinessDoc:
			return 80
		case pdfUnifiedSegmentChartLayout:
			return 60
		case pdfUnifiedSegmentSupportingDoc:
			return 45
		case pdfUnifiedSegmentSignatureStamp:
			return 35
		case pdfUnifiedSegmentLogisticsDoc:
			return 30
		case pdfUnifiedSegmentCoverNotice:
			return 20
		}
	}
	return 0
}

func pdfUnifiedConfidenceWeight(confidence string) int {
	switch strings.TrimSpace(confidence) {
	case "high":
		return 20
	case "medium":
		return 10
	default:
		return 0
	}
}

func pdfUnifiedHasMixedSegments(segments []pdfUnifiedSegment) bool {
	distinct := make(map[string]struct{}, len(segments))
	for _, segment := range segments {
		switch segment.Kind {
		case pdfUnifiedSegmentBusinessDoc, pdfUnifiedSegmentLogisticsDoc, pdfUnifiedSegmentSupportingDoc, pdfUnifiedSegmentChartLayout:
			distinct[segment.Kind] = struct{}{}
		}
	}
	return len(distinct) > 1
}

func pdfUnifiedKeywordHits(text string, keywords []string) []string {
	hits := make([]string, 0, len(keywords))
	for _, keyword := range keywords {
		trimmed := strings.TrimSpace(strings.ToLower(keyword))
		if trimmed == "" {
			continue
		}
		if strings.Contains(text, trimmed) {
			hits = append(hits, trimmed)
		}
	}
	return hits
}

func firstPDFUnifiedAnchorLine(text string) string {
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		return truncateToolText(trimmed, 80)
	}
	return ""
}

func prependPDFUnifiedAnchor(existing []string, anchor string) []string {
	anchor = strings.TrimSpace(anchor)
	if anchor == "" {
		return existing
	}
	out := []string{anchor}
	out = appendPDFUnifiedStringSet(out, existing...)
	if len(out) > 4 {
		out = out[:4]
	}
	return out
}

func appendPDFUnifiedStringSet(existing []string, values ...string) []string {
	seen := make(map[string]struct{}, len(existing))
	out := make([]string, 0, len(existing)+len(values))
	for _, item := range existing {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	for _, item := range values {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	return out
}

func strongerPDFUnifiedConfidence(left string, right string) string {
	if pdfUnifiedConfidenceWeight(right) > pdfUnifiedConfidenceWeight(left) {
		return right
	}
	return left
}

func joinPDFUnifiedSegmentExcerpt(left string, right string, maxChars int) string {
	left = strings.TrimSpace(left)
	right = strings.TrimSpace(right)
	switch {
	case left == "":
		return truncateToolText(right, maxChars)
	case right == "":
		return truncateToolText(left, maxChars)
	case strings.Contains(left, right):
		return truncateToolText(left, maxChars)
	default:
		return truncateToolText(left+"\n"+right, maxChars)
	}
}

func joinPDFUnifiedSegmentText(left string, right string) string {
	left = strings.TrimSpace(left)
	right = strings.TrimSpace(right)
	switch {
	case left == "":
		return right
	case right == "":
		return left
	default:
		return left + "\n\n" + right
	}
}
