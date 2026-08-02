package tools

import (
	"fmt"
	"sort"
	"strings"

	agentxmedia "github.com/wsnacj/agentx-go/runtime/mediaartifact"
)

type pdfAnalyzeDocument struct {
	Path            string               `json:"path"`
	Backend         string               `json:"backend"`
	BackendStatus   pdfBackendStatus     `json:"backend_status,omitempty"`
	PageCount       int                  `json:"page_count"`
	Excerpt         string               `json:"excerpt,omitempty"`
	TotalTextChars  int                  `json:"total_text_chars"`
	PagesWithText   int                  `json:"pages_with_text"`
	KeyPages        []pdfAnalyzePageItem `json:"key_pages,omitempty"`
	Hits            []pdfAnalyzeHit      `json:"hits,omitempty"`
	PageMap         []pdfAnalyzePageItem `json:"page_map,omitempty"`
	DocumentProfile pdfDocumentProfile   `json:"document_profile,omitempty"`
	MediaProfile    pdfMediaProfile      `json:"media_profile,omitempty"`
	AnalysisPlan    pdfAnalysisPlan      `json:"analysis_plan,omitempty"`
	VisualAnalysis  *pdfVisualAnalysis   `json:"visual_analysis,omitempty"`
	Segments        []pdfUnifiedSegment  `json:"segments,omitempty"`
	PrimarySegment  *pdfUnifiedSegment   `json:"primary_segment,omitempty"`
	Supporting      []pdfUnifiedSegment  `json:"supporting_segments,omitempty"`
	HasOutline      bool                 `json:"has_outline"`
	Outline         []pdfOutlineNode     `json:"outline,omitempty"`
	MatchCount      int                  `json:"match_count,omitempty"`
}

type pdfAnalyzeTopDocument struct {
	Path           string `json:"path"`
	PageCount      int    `json:"page_count"`
	TotalTextChars int    `json:"total_text_chars"`
	PagesWithText  int    `json:"pages_with_text"`
	Excerpt        string `json:"excerpt,omitempty"`
	MatchCount     int    `json:"match_count,omitempty"`
}

type pdfAnalyzeMultiPayload struct {
	Paths              []string                `json:"paths"`
	FilesTouched       []string                `json:"files_touched,omitempty"`
	DocumentCount      int                     `json:"document_count"`
	TotalPageCount     int                     `json:"total_page_count"`
	TotalTextChars     int                     `json:"total_text_chars"`
	TotalPagesWithText int                     `json:"total_pages_with_text"`
	Mode               string                  `json:"mode"`
	Query              string                  `json:"query,omitempty"`
	TotalMatches       int                     `json:"total_matches,omitempty"`
	AnalysisPlan       *pdfAnalysisPlan        `json:"analysis_plan,omitempty"`
	VisualAnalysis     *pdfVisualAnalysis      `json:"visual_analysis,omitempty"`
	FocusEnabled       bool                    `json:"focus_enabled,omitempty"`
	FocusQueryClass    string                  `json:"focus_query_class,omitempty"`
	FocusReasonCodes   []string                `json:"focus_reason_codes,omitempty"`
	FocusConfidence    string                  `json:"focus_confidence,omitempty"`
	Documents          []pdfAnalyzeDocument    `json:"documents"`
	TopDocuments       []pdfAnalyzeTopDocument `json:"top_documents,omitempty"`
}

type pdfAnalyzeSinglePayload struct {
	Path             string               `json:"path"`
	FilesTouched     []string             `json:"files_touched,omitempty"`
	Backend          string               `json:"backend"`
	BackendStatus    pdfBackendStatus     `json:"backend_status,omitempty"`
	PageCount        int                  `json:"page_count"`
	Mode             string               `json:"mode"`
	Query            string               `json:"query,omitempty"`
	Excerpt          string               `json:"excerpt,omitempty"`
	TotalTextChars   int                  `json:"total_text_chars"`
	PagesWithText    int                  `json:"pages_with_text"`
	KeyPages         []pdfAnalyzePageItem `json:"key_pages,omitempty"`
	Hits             []pdfAnalyzeHit      `json:"hits,omitempty"`
	PageMap          []pdfAnalyzePageItem `json:"page_map,omitempty"`
	DocumentProfile  pdfDocumentProfile   `json:"document_profile,omitempty"`
	MediaProfile     pdfMediaProfile      `json:"media_profile,omitempty"`
	AnalysisPlan     pdfAnalysisPlan      `json:"analysis_plan,omitempty"`
	VisualAnalysis   *pdfVisualAnalysis   `json:"visual_analysis,omitempty"`
	Segments         []pdfUnifiedSegment  `json:"segments,omitempty"`
	PrimarySegment   *pdfUnifiedSegment   `json:"primary_segment,omitempty"`
	Supporting       []pdfUnifiedSegment  `json:"supporting_segments,omitempty"`
	FocusEnabled     bool                 `json:"focus_enabled,omitempty"`
	FocusQueryClass  string               `json:"focus_query_class,omitempty"`
	FocusReasonCodes []string             `json:"focus_reason_codes,omitempty"`
	FocusConfidence  string               `json:"focus_confidence,omitempty"`
	HasOutline       bool                 `json:"has_outline"`
	Outline          []pdfOutlineNode     `json:"outline,omitempty"`
}

func buildPDFAnalyzeMultiPayload(
	artifactsList []pdfAnalysisArtifacts,
	query string,
	maxHits int,
	maxExcerptChars int,
	includePageMap bool,
	includeOutline bool,
	analysisPlan *pdfAnalysisPlan,
	visualAnalysis *pdfVisualAnalysis,
) pdfAnalyzeMultiPayload {
	payload := pdfAnalyzeMultiPayload{
		Mode:           "overview",
		Query:          strings.TrimSpace(query),
		AnalysisPlan:   clonePDFAnalysisPlanPtr(analysisPlan),
		VisualAnalysis: clonePDFVisualAnalysisPtr(visualAnalysis),
		FilesTouched:   pdfTouchedPathsFromVisualAnalyses(visualAnalysis),
		Documents:      make([]pdfAnalyzeDocument, 0, len(artifactsList)),
	}
	if payload.Query != "" {
		payload.Mode = "search"
	}
	for _, artifacts := range artifactsList {
		item := buildPDFAnalyzeDocument(artifacts, payload.Query, maxHits, maxExcerptChars, includePageMap, includeOutline)
		payload.Paths = append(payload.Paths, item.Path)
		payload.DocumentCount++
		payload.TotalPageCount += item.PageCount
		payload.TotalTextChars += item.TotalTextChars
		payload.TotalPagesWithText += item.PagesWithText
		payload.TotalMatches += item.MatchCount
		payload.Documents = append(payload.Documents, item)
		payload.FilesTouched = appendStringSlicesDedup(payload.FilesTouched, pdfVisualAnalysisTouchedPaths(item.VisualAnalysis))
	}
	payload.TopDocuments = buildPDFAnalyzeTopDocuments(payload.Documents, payload.Mode)
	return applyPDFUnifiedFocusToAnalyzeMultiPayload(payload, buildPDFUnifiedFocusSummaryFromAnalysisArtifacts(query, artifactsList))
}

func applyPDFUnifiedFocusToAnalyzeDocument(item pdfAnalyzeDocument, focus pdfUnifiedDocumentFocus, focusEnabled bool) pdfAnalyzeDocument {
	if !focusEnabled {
		return item
	}
	item.Segments = append([]pdfUnifiedSegment(nil), focus.Segments...)
	if focus.Primary != nil {
		primary := *focus.Primary
		item.PrimarySegment = &primary
		if item.MatchCount == 0 && strings.TrimSpace(primary.Excerpt) != "" {
			item.Excerpt = primary.Excerpt
		}
		focusedPages := filterPDFAnalyzePageMapByPages(item.PageMap, primary.Pages, 3)
		if len(focusedPages) > 0 {
			item.KeyPages = focusedPages
		}
	}
	item.Supporting = append([]pdfUnifiedSegment(nil), focus.Supporting...)
	return item
}

func applyPDFUnifiedFocusToAnalyzeSinglePayload(payload pdfAnalyzeSinglePayload, summary pdfUnifiedFocusSummary) pdfAnalyzeSinglePayload {
	if !summary.Enabled || len(summary.Documents) == 0 {
		return payload
	}
	payload.FocusEnabled = true
	payload.FocusQueryClass = summary.QueryClass
	payload.FocusReasonCodes = append([]string(nil), summary.ReasonCodes...)
	payload.FocusConfidence = summary.Confidence
	doc := pdfAnalyzeDocument{
		Excerpt:    payload.Excerpt,
		KeyPages:   append([]pdfAnalyzePageItem(nil), payload.KeyPages...),
		PageMap:    append([]pdfAnalyzePageItem(nil), payload.PageMap...),
		Segments:   append([]pdfUnifiedSegment(nil), payload.Segments...),
		Supporting: append([]pdfUnifiedSegment(nil), payload.Supporting...),
	}
	doc = applyPDFUnifiedFocusToAnalyzeDocument(doc, summary.Documents[0], true)
	payload.Excerpt = doc.Excerpt
	payload.KeyPages = doc.KeyPages
	payload.Segments = doc.Segments
	payload.PrimarySegment = doc.PrimarySegment
	payload.Supporting = doc.Supporting
	return payload
}

func applyPDFUnifiedFocusToAnalyzeMultiPayload(payload pdfAnalyzeMultiPayload, summary pdfUnifiedFocusSummary) pdfAnalyzeMultiPayload {
	if !summary.Enabled {
		return payload
	}
	payload.FocusEnabled = true
	payload.FocusQueryClass = summary.QueryClass
	payload.FocusReasonCodes = append([]string(nil), summary.ReasonCodes...)
	payload.FocusConfidence = summary.Confidence
	for idx := range payload.Documents {
		if idx >= len(summary.Documents) {
			break
		}
		payload.Documents[idx] = applyPDFUnifiedFocusToAnalyzeDocument(payload.Documents[idx], summary.Documents[idx], true)
	}
	payload.TopDocuments = buildPDFAnalyzeTopDocuments(payload.Documents, payload.Mode)
	return payload
}

func appendStringSlicesDedup(base []string, values []string) []string {
	if len(values) == 0 {
		return base
	}
	seen := make(map[string]bool, len(base)+len(values))
	out := make([]string, 0, len(base)+len(values))
	for _, value := range base {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func buildPDFDocumentSetAnalysisPlan(artifactsList []pdfAnalysisArtifacts) pdfAnalysisPlan {
	if len(artifactsList) == 0 {
		return pdfAnalysisPlan{}
	}
	docCount := len(artifactsList)
	needsVision := false
	needsOCR := false
	graphicDocs := 0
	scannedDocs := 0
	preferredBackend := ""
	for _, artifacts := range artifactsList {
		if preferredBackend == "" {
			preferredBackend = strings.TrimSpace(artifacts.BackendStatus.ExtractBackend)
		}
		if artifacts.AnalysisPlan.NeedsVision {
			needsVision = true
		}
		if artifacts.AnalysisPlan.NeedsOCR {
			needsOCR = true
			scannedDocs++
		}
		if artifacts.MediaProfile.LikelyGraphicDoc || artifacts.MediaProfile.LikelySlideDeck {
			graphicDocs++
		}
	}
	plan := pdfAnalysisPlan{
		PreferredBackend: preferredBackend,
	}
	switch {
	case needsOCR:
		plan.Mode = "vision_ocr"
		plan.NeedsVision = true
		plan.NeedsOCR = true
		plan.Reason = buildPDFDocumentSetReason(docCount, scannedDocs, graphicDocs, true)
		plan.SuggestedNextSteps = []string{
			"Use the document-set native PDF path to compare OCR-visible text and layout across the PDFs.",
			"Drill into documents[].page_map and documents[].hits for per-document evidence after the cross-document summary.",
		}
	case needsVision:
		plan.Mode = "hybrid_vision_text"
		plan.NeedsVision = true
		plan.Reason = buildPDFDocumentSetReason(docCount, scannedDocs, graphicDocs, false)
		plan.SuggestedNextSteps = []string{
			"Use the document-set native PDF path to compare charts, layouts, and visual differences across the PDFs.",
			"Use top_documents and document-level excerpts to decide which PDFs need deeper follow-up extraction.",
		}
	default:
		plan.Mode = "text_first"
		plan.Reason = fmt.Sprintf("document set has %d text-friendly PDFs; plain text extraction is sufficient for a first-pass comparison", docCount)
		plan.SuggestedNextSteps = []string{
			"Use documents[].excerpt and hits to compare the PDFs before escalating to visual analysis.",
		}
	}
	if plan.NeedsVision {
		plan.PreferredClients = preferredPDFVisionClients(plan.Mode)
		var candidates []pdfVisionModelCandidate
		for _, artifacts := range artifactsList {
			if len(artifacts.AnalysisPlan.CandidateModels) > 0 {
				candidates = artifacts.AnalysisPlan.CandidateModels
				break
			}
		}
		plan.CandidateModels = rankPDFVisionCandidates(plan.Mode, candidates, pdfModelResolverConfig{})
		plan.ProviderRouting = buildPDFProviderRouting(plan.Mode, plan.PreferredClients, plan.CandidateModels)
		plan.NativeProviderRouting = buildPDFNativeProviderRouting(plan.Mode, plan.CandidateModels)
		if len(plan.CandidateModels) == 0 {
			plan.Warning = "no vision-capable llmx submodel is currently configured; fall back to per-document text analysis or add a native/pdf-capable vision model"
		} else if strings.TrimSpace(plan.PreferredBackend) == "" {
			plan.PreferredBackend = plan.CandidateModels[0].Name
		}
	}
	return plan
}

func buildPDFDocumentSetReason(docCount int, scannedDocs int, graphicDocs int, needsOCR bool) string {
	switch {
	case needsOCR:
		return fmt.Sprintf("document set contains %d/%d scanned or text-sparse PDFs; OCR-aware visual analysis is recommended", scannedDocs, docCount)
	case graphicDocs > 0:
		return fmt.Sprintf("document set contains %d/%d graphics-heavy PDFs; cross-document visual analysis is recommended", graphicDocs, docCount)
	default:
		return fmt.Sprintf("document set contains %d PDFs with mixed text and layout signals; use visual analysis only if you need cross-document comparison", docCount)
	}
}

func buildPDFAggregateMediaProfile(artifactsList []pdfAnalysisArtifacts) pdfMediaProfile {
	if len(artifactsList) == 0 {
		return pdfMediaProfile{}
	}
	profile := pdfMediaProfile{}
	for _, artifacts := range artifactsList {
		profile.TotalRects += artifacts.MediaProfile.TotalRects
		profile.PagesWithGraphics += artifacts.MediaProfile.PagesWithGraphics
		if artifacts.MediaProfile.MaxRectsPerPage > profile.MaxRectsPerPage {
			profile.MaxRectsPerPage = artifacts.MediaProfile.MaxRectsPerPage
		}
		if artifacts.MediaProfile.LikelyGraphicDoc {
			profile.LikelyGraphicDoc = true
		}
		if artifacts.MediaProfile.LikelySlideDeck {
			profile.LikelySlideDeck = true
		}
	}
	return profile
}

func clonePDFAnalysisPlanPtr(plan *pdfAnalysisPlan) *pdfAnalysisPlan {
	if plan == nil {
		return nil
	}
	cloned := *plan
	cloned.PreferredClients = append([]string(nil), plan.PreferredClients...)
	cloned.CandidateModels = append([]pdfVisionModelCandidate(nil), plan.CandidateModels...)
	cloned.SuggestedNextSteps = append([]string(nil), plan.SuggestedNextSteps...)
	return &cloned
}

func clonePDFVisualAnalysisPtr(analysis *pdfVisualAnalysis) *pdfVisualAnalysis {
	if analysis == nil {
		return nil
	}
	cloned := *analysis
	cloned.Pages = append([]int(nil), analysis.Pages...)
	cloned.AttemptedModels = append([]string(nil), analysis.AttemptedModels...)
	cloned.RenderedPages = append([]agentxmedia.Descriptor(nil), analysis.RenderedPages...)
	cloned.PageTargets = append([]pdfVisualPageTarget(nil), analysis.PageTargets...)
	cloned.ExtractionBatches = append([]pdfVisualExtractionBatch(nil), analysis.ExtractionBatches...)
	cloned.SignalProfile.SummaryOutline = append([]string(nil), analysis.SignalProfile.SummaryOutline...)
	cloned.SignalProfile.ExtractionTargets = append([]string(nil), analysis.SignalProfile.ExtractionTargets...)
	cloned.SignalProfile.ExtractionSchema = append([]pdfVisualExtractionField(nil), analysis.SignalProfile.ExtractionSchema...)
	cloned.SignalProfile.ConfidenceNotes = append([]string(nil), analysis.SignalProfile.ConfidenceNotes...)
	cloned.SignalProfile.FocusAreas = append([]string(nil), analysis.SignalProfile.FocusAreas...)
	cloned.SignalProfile.SuggestedFollowUps = append([]string(nil), analysis.SignalProfile.SuggestedFollowUps...)
	return &cloned
}

func buildPDFAnalyzeDocument(
	artifacts pdfAnalysisArtifacts,
	query string,
	maxHits int,
	maxExcerptChars int,
	includePageMap bool,
	includeOutline bool,
) pdfAnalyzeDocument {
	item := pdfAnalyzeDocument{
		Path:            artifacts.DisplayPath,
		Backend:         artifacts.BackendStatus.ExtractBackend,
		BackendStatus:   artifacts.BackendStatus,
		PageCount:       artifacts.Metadata.PageCount,
		Excerpt:         truncateToolText(joinPDFPageTexts(artifacts.TextResult.Pages), maxExcerptChars),
		TotalTextChars:  pdfTotalTextChars(artifacts.TextResult.Pages),
		PagesWithText:   pdfPagesWithText(artifacts.TextResult.Pages),
		KeyPages:        topPDFAnalyzePages(artifacts.PageMap, 3),
		DocumentProfile: artifacts.DocumentProfile,
		MediaProfile:    artifacts.MediaProfile,
		AnalysisPlan:    artifacts.AnalysisPlan,
		VisualAnalysis:  artifacts.VisualAnalysis,
		HasOutline:      artifacts.Metadata.Outline != nil,
	}
	if includePageMap {
		item.PageMap = artifacts.PageMap
	}
	if includeOutline {
		item.Outline = convertPDFOutline(artifacts.Metadata.Outline)
	}
	if strings.TrimSpace(query) != "" {
		item.Hits = searchPDFAnalyzeHits(artifacts.TextResult.Pages, query, maxHits, maxExcerptChars)
		item.MatchCount = pdfAnalyzeHitCount(item.Hits)
		if len(item.Hits) > 0 {
			item.Excerpt = strings.TrimSpace(item.Hits[0].Excerpt)
			item.KeyPages = topPDFAnalyzePagesFromHits(item.Hits, artifacts.PageMap, 3)
		}
	}
	return item
}

func buildPDFAnalyzeTopDocuments(documents []pdfAnalyzeDocument, mode string) []pdfAnalyzeTopDocument {
	if len(documents) == 0 {
		return nil
	}
	items := make([]pdfAnalyzeTopDocument, 0, len(documents))
	for _, doc := range documents {
		items = append(items, pdfAnalyzeTopDocument{
			Path:           doc.Path,
			PageCount:      doc.PageCount,
			TotalTextChars: doc.TotalTextChars,
			PagesWithText:  doc.PagesWithText,
			Excerpt:        doc.Excerpt,
			MatchCount:     doc.MatchCount,
		})
	}
	sort.SliceStable(items, func(i, j int) bool {
		switch strings.ToLower(strings.TrimSpace(mode)) {
		case "search":
			if items[i].MatchCount == items[j].MatchCount {
				if items[i].TotalTextChars == items[j].TotalTextChars {
					return items[i].Path < items[j].Path
				}
				return items[i].TotalTextChars > items[j].TotalTextChars
			}
			return items[i].MatchCount > items[j].MatchCount
		default:
			if items[i].TotalTextChars == items[j].TotalTextChars {
				return items[i].Path < items[j].Path
			}
			return items[i].TotalTextChars > items[j].TotalTextChars
		}
	})
	return items
}

func pdfAnalyzeHitCount(hits []pdfAnalyzeHit) int {
	total := 0
	for _, hit := range hits {
		total += hit.Matches
	}
	return total
}

func filterPDFAnalyzePageMapByPages(pageMap []pdfAnalyzePageItem, pages []int, limit int) []pdfAnalyzePageItem {
	if len(pageMap) == 0 || len(pages) == 0 || limit <= 0 {
		return nil
	}
	allowed := make(map[int]struct{}, len(pages))
	for _, page := range pages {
		allowed[page] = struct{}{}
	}
	out := make([]pdfAnalyzePageItem, 0, minInt(len(pages), limit))
	for _, item := range pageMap {
		if _, ok := allowed[item.Page]; !ok {
			continue
		}
		out = append(out, item)
		if len(out) >= limit {
			break
		}
	}
	return out
}
