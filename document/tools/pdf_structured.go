package tools

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

type pdfAnalysisArtifacts struct {
	Root            string
	Path            string
	DisplayPath     string
	TextResult      PDFTextResult
	Metadata        PDFMetadataResult
	BackendStatus   pdfBackendStatus
	PageMap         []pdfAnalyzePageItem
	StructureItems  []pdfUnifiedStructureItem
	DocumentProfile pdfDocumentProfile
	MediaProfile    pdfMediaProfile
	AnalysisPlan    pdfAnalysisPlan
	VisualAnalysis  *pdfVisualAnalysis
}

type pdfStructuredPayload struct {
	Path              string                         `json:"path"`
	FilesTouched      []string                       `json:"files_touched,omitempty"`
	Backend           string                         `json:"backend"`
	BackendStatus     pdfBackendStatus               `json:"backend_status,omitempty"`
	PageCount         int                            `json:"page_count"`
	Query             string                         `json:"query,omitempty"`
	Mode              string                         `json:"mode,omitempty"`
	Status            string                         `json:"status,omitempty"`
	ReviewRequired    bool                           `json:"review_required,omitempty"`
	ReviewSummary     *pdfStructuredReviewSummary    `json:"review_summary,omitempty"`
	Warning           string                         `json:"warning,omitempty"`
	PageMap           []pdfAnalyzePageItem           `json:"page_map,omitempty"`
	DocumentProfile   pdfDocumentProfile             `json:"document_profile,omitempty"`
	MediaProfile      pdfMediaProfile                `json:"media_profile,omitempty"`
	AnalysisPlan      pdfAnalysisPlan                `json:"analysis_plan,omitempty"`
	VisualAnalysis    *pdfVisualAnalysis             `json:"visual_analysis,omitempty"`
	Segments          []pdfUnifiedSegment            `json:"segments,omitempty"`
	PrimarySegment    *pdfUnifiedSegment             `json:"primary_segment,omitempty"`
	Supporting        []pdfUnifiedSegment            `json:"supporting_segments,omitempty"`
	FocusEnabled      bool                           `json:"focus_enabled,omitempty"`
	FocusQueryClass   string                         `json:"focus_query_class,omitempty"`
	FocusReasonCodes  []string                       `json:"focus_reason_codes,omitempty"`
	FocusConfidence   string                         `json:"focus_confidence,omitempty"`
	StructuredBatches []pdfStructuredBatchResult     `json:"structured_batches,omitempty"`
	ResultEvaluation  *pdfStructuredResultEvaluation `json:"result_evaluation,omitempty"`
}

type pdfStructuredMultiPayload struct {
	Paths            []string                       `json:"paths"`
	FilesTouched     []string                       `json:"files_touched,omitempty"`
	DocumentCount    int                            `json:"document_count"`
	Query            string                         `json:"query,omitempty"`
	Mode             string                         `json:"mode,omitempty"`
	Status           string                         `json:"status,omitempty"`
	ReviewRequired   bool                           `json:"review_required,omitempty"`
	ReviewSummary    *pdfStructuredReviewSummary    `json:"review_summary,omitempty"`
	AnalysisPlan     *pdfAnalysisPlan               `json:"analysis_plan,omitempty"`
	VisualAnalysis   *pdfVisualAnalysis             `json:"visual_analysis,omitempty"`
	FocusEnabled     bool                           `json:"focus_enabled,omitempty"`
	FocusQueryClass  string                         `json:"focus_query_class,omitempty"`
	FocusReasonCodes []string                       `json:"focus_reason_codes,omitempty"`
	FocusConfidence  string                         `json:"focus_confidence,omitempty"`
	Warning          string                         `json:"warning,omitempty"`
	Documents        []pdfStructuredPayload         `json:"documents"`
	TopDocuments     []pdfStructuredTopDocument     `json:"top_documents,omitempty"`
	ResultEvaluation *pdfStructuredResultEvaluation `json:"result_evaluation,omitempty"`
}

type pdfStructuredTopDocument struct {
	Path                   string   `json:"path"`
	Status                 string   `json:"status,omitempty"`
	ReviewRequired         bool     `json:"review_required,omitempty"`
	BatchCount             int      `json:"batch_count,omitempty"`
	AverageCompletionRatio float64  `json:"average_completion_ratio,omitempty"`
	BatchesRequiringReview int      `json:"batches_requiring_review,omitempty"`
	LowConfidenceFields    int      `json:"low_confidence_fields,omitempty"`
	SelectionReasonCodes   []string `json:"selection_reason_codes,omitempty"`
	SelectionReasons       []string `json:"selection_reasons,omitempty"`
	TopNotes               []string `json:"top_notes,omitempty"`
}

type pdfStructuredReviewSummary struct {
	BatchesRequiringReview int      `json:"batches_requiring_review,omitempty"`
	LowConfidenceFields    int      `json:"low_confidence_fields,omitempty"`
	BatchTargets           []string `json:"batch_targets,omitempty"`
	FocusTarget            string   `json:"focus_target,omitempty"`
	ReviewReasonCodes      []string `json:"review_reason_codes,omitempty"`
	ReviewDrivers          []string `json:"review_drivers,omitempty"`
	TopNotes               []string `json:"top_notes,omitempty"`
}

type pdfStructuredBatchResult struct {
	Target                string                       `json:"target,omitempty"`
	Pages                 []int                        `json:"pages,omitempty"`
	Priority              string                       `json:"priority,omitempty"`
	FocusAligned          bool                         `json:"focus_aligned,omitempty"`
	FocusTarget           string                       `json:"focus_target,omitempty"`
	Status                string                       `json:"status,omitempty"`
	ReviewRequired        bool                         `json:"review_required,omitempty"`
	ReviewNotes           []string                     `json:"review_notes,omitempty"`
	ReviewReasonCodes     []string                     `json:"review_reason_codes,omitempty"`
	ReviewDrivers         []string                     `json:"review_drivers,omitempty"`
	LowConfidenceFields   []string                     `json:"low_confidence_fields,omitempty"`
	Reasons               []string                     `json:"reasons,omitempty"`
	SummaryMode           string                       `json:"summary_mode,omitempty"`
	CompletionRatio       float64                      `json:"completion_ratio,omitempty"`
	RenderedOutput        []string                     `json:"rendered_output,omitempty"`
	Result                map[string]any               `json:"result,omitempty"`
	MissingRequiredFields []string                     `json:"missing_required_fields,omitempty"`
	ValidationChecks      []string                     `json:"validation_checks,omitempty"`
	NormalizationRules    []string                     `json:"normalization_rules,omitempty"`
	AggregationStrategy   string                       `json:"aggregation_strategy,omitempty"`
	AggregationRules      []string                     `json:"aggregation_rules,omitempty"`
	Sections              []pdfStructuredSectionResult `json:"sections,omitempty"`
}

type pdfStructuredSectionResult struct {
	Name               string                     `json:"name,omitempty"`
	Purpose            string                     `json:"purpose,omitempty"`
	Status             string                     `json:"status,omitempty"`
	NeedsReview        bool                       `json:"needs_review,omitempty"`
	ReviewNotes        []string                   `json:"review_notes,omitempty"`
	RenderTemplate     string                     `json:"render_template,omitempty"`
	Rendered           string                     `json:"rendered,omitempty"`
	MissingFields      []string                   `json:"missing_fields,omitempty"`
	CompletionCriteria []string                   `json:"completion_criteria,omitempty"`
	QualityChecks      []string                   `json:"quality_checks,omitempty"`
	Fields             []pdfStructuredFieldResult `json:"fields,omitempty"`
}

type pdfStructuredFieldResult struct {
	Name                  string   `json:"name,omitempty"`
	Kind                  string   `json:"kind,omitempty"`
	Required              bool     `json:"required,omitempty"`
	Filled                bool     `json:"filled,omitempty"`
	NeedsReview           bool     `json:"needs_review,omitempty"`
	CrossDocumentPriority bool     `json:"cross_document_priority,omitempty"`
	Value                 any      `json:"value,omitempty"`
	Source                string   `json:"source,omitempty"`
	Confidence            string   `json:"confidence,omitempty"`
	Evidence              []string `json:"evidence,omitempty"`
	IssueFlags            []string `json:"issue_flags,omitempty"`
	Notes                 []string `json:"notes,omitempty"`
}

type pdfStructuredFieldCandidate struct {
	Value      any
	Source     string
	Confidence string
	Evidence   []string
	Notes      []string
}

type pdfStructuredFieldInferenceContext struct {
	FocusAligned  bool
	FocusTarget   string
	PriorityField bool
}

func buildPDFAnalysisArtifacts(
	ctx context.Context,
	runtime pdfBackendRuntime,
	resolved string,
	display string,
	query string,
	maxExcerptChars int,
	ocrxConfigPath string,
	includeVisualAnalysis bool,
	visionModel string,
	maxVisualPages int,
) (pdfAnalysisArtifacts, error) {
	textResult, backendStatus, err := runtime.runText(ctx, func(backend PDFBackend) (PDFTextResult, error) {
		return backend.ExtractAllText(ctx, resolved)
	})
	if err != nil {
		return pdfAnalysisArtifacts{}, err
	}
	metadata, metaStatus, err := runtime.runMetadata(ctx, false, func(backend PDFBackend, includeFonts bool) (PDFMetadataResult, error) {
		return backend.ReadMetadata(ctx, resolved, includeFonts)
	})
	if err != nil {
		return pdfAnalysisArtifacts{}, err
	}
	backendStatus = mergePDFBackendStatus(backendStatus, metaStatus)
	pageMap := buildPDFAnalyzePageMap(textResult.Pages, maxExcerptChars)
	documentProfile := buildPDFDocumentProfile(metadata.PageCount, metadata.Outline != nil, pageMap)
	mediaProfile := buildPDFMediaProfile(metadata, documentProfile, pageMap)
	analysisPlan := buildPDFAnalysisPlan(documentProfile, mediaProfile, backendStatus)
	ocrUsed := false
	textResult, backendStatus, ocrUsed = maybeSupplementPDFUnifiedTextWithOCR(ctx, resolved, nil, textResult, backendStatus, documentProfile, analysisPlan, ocrxConfigPath)
	if ocrUsed {
		pageMap = buildPDFAnalyzePageMap(textResult.Pages, maxExcerptChars)
		documentProfile = buildPDFDocumentProfile(metadata.PageCount, metadata.Outline != nil, pageMap)
		mediaProfile = buildPDFMediaProfile(metadata, documentProfile, pageMap)
		analysisPlan = buildPDFAnalysisPlan(documentProfile, mediaProfile, backendStatus)
	}
	structureItems := buildPDFUnifiedStructureItems(textResult.Pages, pageMap, metadata.PageCount, mediaProfile)
	var visualAnalysis *pdfVisualAnalysis
	if includeVisualAnalysis {
		result := runPDFVisualAnalysis(ctx, runtime.root, resolved, query, pageMap, mediaProfile, analysisPlan, visionModel, maxVisualPages)
		visualAnalysis = &result
	}
	return pdfAnalysisArtifacts{
		Root:            runtime.root,
		Path:            resolved,
		DisplayPath:     display,
		TextResult:      textResult,
		Metadata:        metadata,
		BackendStatus:   backendStatus,
		PageMap:         pageMap,
		StructureItems:  structureItems,
		DocumentProfile: documentProfile,
		MediaProfile:    mediaProfile,
		AnalysisPlan:    analysisPlan,
		VisualAnalysis:  visualAnalysis,
	}, nil
}

func enrichPDFAnalysisArtifactsWithVisualAnalysis(
	ctx context.Context,
	artifacts pdfAnalysisArtifacts,
	query string,
	visionModel string,
	maxVisualPages int,
) pdfAnalysisArtifacts {
	result := runPDFVisualAnalysis(ctx, artifacts.Root, artifacts.Path, query, artifacts.PageMap, artifacts.MediaProfile, artifacts.AnalysisPlan, visionModel, maxVisualPages)
	artifacts.VisualAnalysis = &result
	return artifacts
}

func buildPDFStructuredPayload(
	artifacts pdfAnalysisArtifacts,
	query string,
	includePageMap bool,
	documentSetPlan *pdfAnalysisPlan,
	documentSetVisual *pdfVisualAnalysis,
) pdfStructuredPayload {
	focusSummary := buildPDFUnifiedFocusSummaryFromAnalysisArtifacts(query, []pdfAnalysisArtifacts{artifacts})
	var documentFocus pdfUnifiedDocumentFocus
	if len(focusSummary.Documents) > 0 {
		documentFocus = focusSummary.Documents[0]
	}
	visual := artifacts.VisualAnalysis
	signalProfile := pdfVisualSignalProfile{}
	if visual != nil {
		signalProfile = visual.SignalProfile
	}
	documentSetFocus := firstNonEmpty(
		pdfStructuredDocumentSetFocusTarget(documentSetPlan, documentSetVisual),
		pdfStructuredInferFocusTargetFromQuery(query),
	)
	var batches []pdfVisualExtractionBatch
	if visual != nil && len(visual.ExtractionBatches) > 0 {
		batches = visual.ExtractionBatches
	} else {
		if pdfVisualSignalProfileIsEmpty(signalProfile) {
			signalProfile = buildPDFVisualSignalProfile("", artifacts.MediaProfile, artifacts.AnalysisPlan)
		}
		pageTargets := buildPDFVisualPageTargets(artifacts.PageMap, allPDFPagesFromMap(artifacts.PageMap), artifacts.MediaProfile, signalProfile)
		batches = buildPDFVisualExtractionBatches(pageTargets, signalProfile)
	}
	batches = applyPDFUnifiedFocusToStructuredBatches(batches, documentFocus, focusSummary.QueryClass, focusSummary.Enabled)
	batches = pdfStructuredEnsureDocumentSetFocusBatch(batches, signalProfile, documentSetFocus)
	structured := make([]pdfStructuredBatchResult, 0, len(batches))
	status := "empty"
	warnings := make([]string, 0, 2)
	reviewRequired := false
	reviewBatchTargets := make([]string, 0, len(batches))
	reviewNotes := make([]string, 0, len(batches))
	lowConfidenceFieldCount := 0
	for _, batch := range batches {
		item := buildPDFStructuredBatchResult(batch, artifacts.PageMap, visual, documentSetFocus)
		if item.Status == "complete" {
			status = "complete"
		} else if item.Status == "partial" && status == "empty" {
			status = "partial"
		}
		if item.Status == "partial" && status == "complete" {
			status = "partial"
		}
		if len(item.MissingRequiredFields) > 0 {
			warnings = append(warnings, fmt.Sprintf("%s batch missing required fields: %s", batch.Target, strings.Join(item.MissingRequiredFields, ", ")))
		}
		if item.ReviewRequired {
			reviewRequired = true
			reviewBatchTargets = append(reviewBatchTargets, item.Target)
			reviewNotes = append(reviewNotes, item.ReviewNotes...)
			lowConfidenceFieldCount += len(item.LowConfidenceFields)
		}
		structured = append(structured, item)
	}
	if status == "empty" && len(structured) > 0 {
		status = "partial"
	}
	payload := pdfStructuredPayload{
		Path:              artifacts.DisplayPath,
		FilesTouched:      pdfVisualAnalysisTouchedPaths(visual),
		Backend:           artifacts.BackendStatus.ExtractBackend,
		BackendStatus:     artifacts.BackendStatus,
		PageCount:         artifacts.Metadata.PageCount,
		Query:             strings.TrimSpace(query),
		Mode:              "structured_extract",
		Status:            status,
		ReviewRequired:    reviewRequired,
		DocumentProfile:   artifacts.DocumentProfile,
		MediaProfile:      artifacts.MediaProfile,
		AnalysisPlan:      artifacts.AnalysisPlan,
		VisualAnalysis:    visual,
		StructuredBatches: structured,
	}
	if reviewRequired {
		payload.ReviewSummary = &pdfStructuredReviewSummary{
			BatchesRequiringReview: len(dedupePDFVisualStrings(reviewBatchTargets)),
			LowConfidenceFields:    lowConfidenceFieldCount,
			BatchTargets:           dedupePDFVisualStrings(reviewBatchTargets),
			FocusTarget:            documentSetFocus,
			ReviewReasonCodes:      pdfStructuredPrependFocusReasonCode(pdfStructuredAggregateBatchReviewReasonCodes(structured), documentSetFocus),
			ReviewDrivers:          pdfStructuredPrependFocusDriver(pdfStructuredAggregateBatchReviewDrivers(structured), documentSetFocus),
			TopNotes:               truncatePDFStructuredNotes(reviewNotes, 6),
		}
	}
	if includePageMap {
		payload.PageMap = artifacts.PageMap
	}
	if len(warnings) > 0 {
		payload.Warning = strings.Join(dedupePDFVisualStrings(warnings), "; ")
	}
	payload.ResultEvaluation = buildPDFStructuredResultEvaluation(batches, payload.StructuredBatches, payload.VisualAnalysis, documentSetFocus)
	return applyPDFUnifiedFocusToStructuredPayload(payload, focusSummary, 0)
}

func applyPDFUnifiedFocusToStructuredPayload(payload pdfStructuredPayload, summary pdfUnifiedFocusSummary, index int) pdfStructuredPayload {
	if !summary.Enabled || index < 0 || index >= len(summary.Documents) {
		return payload
	}
	focus := summary.Documents[index]
	payload.FocusEnabled = true
	payload.FocusQueryClass = summary.QueryClass
	payload.FocusReasonCodes = append([]string(nil), summary.ReasonCodes...)
	payload.FocusConfidence = summary.Confidence
	payload.Segments = append([]pdfUnifiedSegment(nil), focus.Segments...)
	if focus.Primary != nil {
		primary := *focus.Primary
		payload.PrimarySegment = &primary
	}
	payload.Supporting = append([]pdfUnifiedSegment(nil), focus.Supporting...)
	if payload.ReviewSummary != nil {
		payload.ReviewSummary.TopNotes = prependPDFStructuredStrings(
			payload.ReviewSummary.TopNotes,
			pdfUnifiedStructuredFocusNote(focus),
		)
		if focus.Primary != nil {
			payload.ReviewSummary.ReviewReasonCodes = truncatePDFStructuredCodes(
				append([]string{"subdocument_focus_" + strings.TrimSpace(focus.Primary.Kind)}, payload.ReviewSummary.ReviewReasonCodes...),
				8,
			)
		}
	}
	return payload
}

func applyPDFUnifiedFocusToStructuredBatches(
	batches []pdfVisualExtractionBatch,
	focus pdfUnifiedDocumentFocus,
	queryClass string,
	focusEnabled bool,
) []pdfVisualExtractionBatch {
	if !focusEnabled || len(batches) == 0 || focus.Primary == nil {
		return batches
	}
	allowedPages := pdfUnifiedStructuredAllowedPages(focus, queryClass)
	if len(allowedPages) == 0 {
		return batches
	}
	out := make([]pdfVisualExtractionBatch, 0, len(batches))
	for _, batch := range batches {
		pruned := intersectPDFPagesPreserveOrder(batch.Pages, allowedPages)
		if len(pruned) == 0 {
			continue
		}
		if len(pruned) < len(batch.Pages) {
			batch.Reasons = prependPDFStructuredStrings(
				batch.Reasons,
				fmt.Sprintf("Subdocument focus limited this batch to pages %s.", formatPDFPageSelection(pruned)),
			)
		}
		batch.Pages = pruned
		out = append(out, batch)
	}
	if len(out) == 0 {
		return batches
	}
	return out
}

func pdfUnifiedStructuredAllowedPages(focus pdfUnifiedDocumentFocus, queryClass string) []int {
	pages := make([]int, 0, 6)
	if focus.Primary != nil {
		pages = append(pages, focus.Primary.Pages...)
	}
	supporting := selectPDFUnifiedSupportingSegmentsForPrompt(focus, queryClass)
	if queryClass == pdfUnifiedQueryClassChartSummary && len(supporting) == 0 {
		supporting = append([]pdfUnifiedSegment(nil), focus.Supporting...)
	}
	for _, segment := range supporting {
		pages = append(pages, segment.Pages...)
	}
	return dedupeSortedPDFPages(pages)
}

func intersectPDFPagesPreserveOrder(pages []int, allowed []int) []int {
	if len(pages) == 0 || len(allowed) == 0 {
		return nil
	}
	allowedSet := make(map[int]struct{}, len(allowed))
	for _, page := range allowed {
		allowedSet[page] = struct{}{}
	}
	out := make([]int, 0, len(pages))
	for _, page := range pages {
		if _, ok := allowedSet[page]; ok {
			out = append(out, page)
		}
	}
	return dedupeSortedPDFPages(out)
}

func dedupeSortedPDFPages(pages []int) []int {
	if len(pages) == 0 {
		return nil
	}
	out := make([]int, 0, len(pages))
	seen := map[int]struct{}{}
	for _, page := range pages {
		if page <= 0 {
			continue
		}
		if _, ok := seen[page]; ok {
			continue
		}
		seen[page] = struct{}{}
		out = append(out, page)
	}
	sort.Ints(out)
	return out
}

func applyPDFUnifiedFocusToStructuredMultiPayload(payload pdfStructuredMultiPayload, summary pdfUnifiedFocusSummary) pdfStructuredMultiPayload {
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
		payload.Documents[idx] = applyPDFUnifiedFocusToStructuredPayload(payload.Documents[idx], summary, idx)
	}
	if payload.ReviewSummary != nil {
		payload.ReviewSummary.TopNotes = prependPDFStructuredStrings(
			payload.ReviewSummary.TopNotes,
			pdfUnifiedStructuredSummaryNote(summary),
		)
	}
	return payload
}

func pdfUnifiedStructuredFocusNote(focus pdfUnifiedDocumentFocus) string {
	if focus.Primary == nil {
		return ""
	}
	return fmt.Sprintf(
		"Primary subdocument focus: %s on pages %s; treat other segments as supporting evidence unless they change the conclusion.",
		strings.TrimSpace(focus.Primary.Kind),
		formatPDFPageSelection(focus.Primary.Pages),
	)
}

func pdfUnifiedStructuredSummaryNote(summary pdfUnifiedFocusSummary) string {
	if !summary.Enabled {
		return ""
	}
	return fmt.Sprintf(
		"Subdocument focus active for %s queries with %s confidence.",
		strings.TrimSpace(summary.QueryClass),
		strings.TrimSpace(summary.Confidence),
	)
}

func buildPDFStructuredMultiPayload(
	documents []pdfStructuredPayload,
	query string,
	analysisPlan *pdfAnalysisPlan,
	visualAnalysis *pdfVisualAnalysis,
) pdfStructuredMultiPayload {
	contextualDocuments := make([]pdfStructuredPayload, 0, len(documents))
	for _, doc := range documents {
		contextualDocuments = append(contextualDocuments, applyPDFStructuredDocumentSetContext(doc, analysisPlan, visualAnalysis))
	}
	payload := pdfStructuredMultiPayload{
		Query:          strings.TrimSpace(query),
		Mode:           "structured_extract",
		AnalysisPlan:   clonePDFAnalysisPlanPtr(analysisPlan),
		VisualAnalysis: clonePDFVisualAnalysisPtr(visualAnalysis),
		FilesTouched:   pdfTouchedPathsFromVisualAnalyses(visualAnalysis),
		Documents:      contextualDocuments,
	}
	if len(contextualDocuments) == 0 {
		return payload
	}
	reviewRequired := false
	allComplete := true
	hasProgress := false
	warnings := make([]string, 0, len(contextualDocuments))
	reviewTargets := make([]string, 0, len(contextualDocuments))
	reviewNotes := make([]string, 0, len(contextualDocuments))
	reviewDrivers := make([]string, 0, len(contextualDocuments)*2)
	lowConfidenceFields := 0
	focusTarget := firstNonEmpty(
		pdfStructuredDocumentSetFocusTarget(payload.AnalysisPlan, payload.VisualAnalysis),
		pdfStructuredPayloadsFocusTarget(contextualDocuments),
		pdfStructuredInferFocusTargetFromQuery(query),
	)
	for _, doc := range contextualDocuments {
		payload.Paths = append(payload.Paths, doc.Path)
		payload.FilesTouched = appendStringSlicesDedup(payload.FilesTouched, doc.FilesTouched)
		payload.DocumentCount++
		switch strings.ToLower(strings.TrimSpace(doc.Status)) {
		case "complete":
			hasProgress = true
		case "partial":
			allComplete = false
			hasProgress = true
		default:
			allComplete = false
		}
		if strings.TrimSpace(doc.Warning) != "" {
			warnings = append(warnings, fmt.Sprintf("%s: %s", doc.Path, strings.TrimSpace(doc.Warning)))
		}
		if !doc.ReviewRequired {
			continue
		}
		reviewRequired = true
		if doc.ReviewSummary != nil {
			lowConfidenceFields += doc.ReviewSummary.LowConfidenceFields
			for _, target := range doc.ReviewSummary.BatchTargets {
				target = strings.TrimSpace(target)
				if target == "" {
					continue
				}
				reviewTargets = append(reviewTargets, fmt.Sprintf("%s:%s", doc.Path, target))
			}
			for _, note := range doc.ReviewSummary.TopNotes {
				note = strings.TrimSpace(note)
				if note == "" {
					continue
				}
				reviewNotes = append(reviewNotes, fmt.Sprintf("%s: %s", doc.Path, note))
			}
			for _, driver := range doc.ReviewSummary.ReviewDrivers {
				driver = strings.TrimSpace(driver)
				if driver == "" {
					continue
				}
				reviewDrivers = append(reviewDrivers, fmt.Sprintf("%s: %s", doc.Path, driver))
			}
		}
	}
	switch {
	case allComplete:
		payload.Status = "complete"
	case hasProgress:
		payload.Status = "partial"
	default:
		payload.Status = "empty"
	}
	payload.ReviewRequired = reviewRequired
	payload.TopDocuments = buildPDFStructuredTopDocuments(contextualDocuments, payload.AnalysisPlan, payload.VisualAnalysis)
	payload.ResultEvaluation = buildPDFStructuredMultiResultEvaluation(payload.Documents, payload.AnalysisPlan, payload.VisualAnalysis, payload.TopDocuments)
	if len(warnings) > 0 {
		payload.Warning = strings.Join(dedupePDFVisualStrings(warnings), "; ")
	}
	if payload.VisualAnalysis != nil && strings.TrimSpace(payload.VisualAnalysis.Warning) != "" {
		if strings.TrimSpace(payload.Warning) == "" {
			payload.Warning = strings.TrimSpace(payload.VisualAnalysis.Warning)
		} else {
			payload.Warning = payload.Warning + "; " + strings.TrimSpace(payload.VisualAnalysis.Warning)
		}
	}
	if reviewRequired {
		topNotes := append(
			pdfStructuredDocumentSetReviewNotes(payload.AnalysisPlan, payload.VisualAnalysis),
			reviewNotes...,
		)
		payload.ReviewSummary = &pdfStructuredReviewSummary{
			BatchesRequiringReview: len(dedupePDFVisualStrings(reviewTargets)),
			LowConfidenceFields:    lowConfidenceFields,
			BatchTargets:           dedupePDFVisualStrings(reviewTargets),
			FocusTarget:            focusTarget,
			ReviewReasonCodes:      pdfStructuredPrependFocusReasonCode(pdfStructuredAggregateDocumentReviewReasonCodes(contextualDocuments), focusTarget),
			ReviewDrivers:          pdfStructuredPrependFocusDriver(truncatePDFStructuredNotes(reviewDrivers, 8), focusTarget),
			TopNotes:               truncatePDFStructuredNotes(topNotes, 8),
		}
	}
	return payload
}

func applyPDFStructuredDocumentSetContext(
	doc pdfStructuredPayload,
	analysisPlan *pdfAnalysisPlan,
	visualAnalysis *pdfVisualAnalysis,
) pdfStructuredPayload {
	focus := firstNonEmpty(pdfStructuredDocumentSetFocusTarget(analysisPlan, visualAnalysis), pdfStructuredPayloadFocusTarget(doc))
	if focus == "" || len(doc.StructuredBatches) == 0 {
		return doc
	}
	hasFocusBatch := false
	for _, batch := range doc.StructuredBatches {
		if pdfStructuredNormalizeDocumentTarget(batch.Target) == focus {
			hasFocusBatch = true
			break
		}
	}
	if !hasFocusBatch {
		return doc
	}

	adjusted := make([]pdfStructuredBatchResult, 0, len(doc.StructuredBatches))
	for _, batch := range doc.StructuredBatches {
		target := pdfStructuredNormalizeDocumentTarget(batch.Target)
		if pdfStructuredShouldPruneBatchForDocumentSetFocus(target, focus) {
			continue
		}
		if target == focus {
			batch.Priority = "high"
			batch.FocusAligned = true
			batch.FocusTarget = focus
			batch.Reasons = append([]string{
				fmt.Sprintf("aligned with document-set %s focus", pdfStructuredDocumentSetFocusLabel(focus)),
			}, batch.Reasons...)
			batch.Reasons = dedupePDFVisualStrings(batch.Reasons)
			batch = applyPDFStructuredFieldLevelContext(batch, focus)
		} else if target != "" && strings.TrimSpace(batch.Priority) == "" {
			batch.Priority = "low"
		}
		adjusted = append(adjusted, batch)
	}
	if len(adjusted) == 0 {
		return doc
	}
	sort.SliceStable(adjusted, func(i, j int) bool {
		left := pdfStructuredNormalizeDocumentTarget(adjusted[i].Target)
		right := pdfStructuredNormalizeDocumentTarget(adjusted[j].Target)
		if left == focus && right != focus {
			return true
		}
		if left != focus && right == focus {
			return false
		}
		if pdfStructuredBatchPriorityRank(adjusted[i].Priority) == pdfStructuredBatchPriorityRank(adjusted[j].Priority) {
			if adjusted[i].CompletionRatio == adjusted[j].CompletionRatio {
				return adjusted[i].Target < adjusted[j].Target
			}
			return adjusted[i].CompletionRatio > adjusted[j].CompletionRatio
		}
		return pdfStructuredBatchPriorityRank(adjusted[i].Priority) > pdfStructuredBatchPriorityRank(adjusted[j].Priority)
	})
	doc.StructuredBatches = adjusted
	return rebuildPDFStructuredPayloadDerivedFields(doc, analysisPlan, visualAnalysis)
}

func pdfStructuredShouldPruneBatchForDocumentSetFocus(target string, focus string) bool {
	if focus == "" || target == "" {
		return false
	}
	if focus == "layout" {
		return false
	}
	if target == "layout" {
		return true
	}
	switch focus {
	case "chart", "table", "diagram":
		return target == "ocr_text"
	default:
		return false
	}
}

func pdfStructuredBatchPriorityRank(priority string) int {
	switch strings.ToLower(strings.TrimSpace(priority)) {
	case "high":
		return 3
	case "medium":
		return 2
	case "low":
		return 1
	default:
		return 0
	}
}

func rebuildPDFStructuredPayloadDerivedFields(
	doc pdfStructuredPayload,
	analysisPlan *pdfAnalysisPlan,
	visualAnalysis *pdfVisualAnalysis,
) pdfStructuredPayload {
	status := "empty"
	warnings := make([]string, 0, len(doc.StructuredBatches))
	reviewRequired := false
	reviewBatchTargets := make([]string, 0, len(doc.StructuredBatches))
	reviewNotes := make([]string, 0, len(doc.StructuredBatches))
	lowConfidenceFieldCount := 0
	for i := range doc.StructuredBatches {
		doc.StructuredBatches[i] = pdfStructuredFinalizeBatchExplainability(doc.StructuredBatches[i])
		batch := doc.StructuredBatches[i]
		if batch.Status == "complete" {
			status = "complete"
		} else if batch.Status == "partial" && status == "empty" {
			status = "partial"
		}
		if batch.Status == "partial" && status == "complete" {
			status = "partial"
		}
		if len(batch.MissingRequiredFields) > 0 {
			warnings = append(warnings, fmt.Sprintf("%s batch missing required fields: %s", batch.Target, strings.Join(batch.MissingRequiredFields, ", ")))
		}
		if batch.ReviewRequired {
			reviewRequired = true
			reviewBatchTargets = append(reviewBatchTargets, batch.Target)
			reviewNotes = append(reviewNotes, batch.ReviewNotes...)
			lowConfidenceFieldCount += len(batch.LowConfidenceFields)
		}
	}
	if status == "empty" && len(doc.StructuredBatches) > 0 {
		status = "partial"
	}
	doc.Status = status
	doc.ReviewRequired = reviewRequired
	if reviewRequired {
		topNotes := append([]string(nil), reviewNotes...)
		if note := pdfStructuredDocumentSetDocumentNote(doc, analysisPlan, visualAnalysis); note != "" {
			topNotes = append([]string{note}, topNotes...)
		}
		doc.ReviewSummary = &pdfStructuredReviewSummary{
			BatchesRequiringReview: len(dedupePDFVisualStrings(reviewBatchTargets)),
			LowConfidenceFields:    lowConfidenceFieldCount,
			BatchTargets:           dedupePDFVisualStrings(reviewBatchTargets),
			FocusTarget:            firstNonEmpty(pdfStructuredDocumentSetFocusTarget(analysisPlan, visualAnalysis), pdfStructuredPayloadFocusTarget(doc)),
			ReviewReasonCodes: pdfStructuredPrependFocusReasonCode(
				pdfStructuredAggregateBatchReviewReasonCodes(doc.StructuredBatches),
				firstNonEmpty(pdfStructuredDocumentSetFocusTarget(analysisPlan, visualAnalysis), pdfStructuredPayloadFocusTarget(doc)),
			),
			ReviewDrivers: pdfStructuredPrependFocusDriver(
				pdfStructuredAggregateBatchReviewDrivers(doc.StructuredBatches),
				firstNonEmpty(pdfStructuredDocumentSetFocusTarget(analysisPlan, visualAnalysis), pdfStructuredPayloadFocusTarget(doc)),
			),
			TopNotes: truncatePDFStructuredNotes(topNotes, 6),
		}
	} else {
		doc.ReviewSummary = nil
	}
	if len(warnings) > 0 {
		doc.Warning = strings.Join(dedupePDFVisualStrings(warnings), "; ")
	} else {
		doc.Warning = ""
	}
	return refreshPDFStructuredResultEvaluation(doc, analysisPlan, visualAnalysis)
}

func applyPDFStructuredFieldLevelContext(batch pdfStructuredBatchResult, focus string) pdfStructuredBatchResult {
	priorityFields := pdfStructuredDocumentSetPriorityFieldSet(focus)
	if len(priorityFields) == 0 {
		return pdfStructuredFinalizeBatchExplainability(batch)
	}
	batch.ValidationChecks = prependPDFStructuredStrings(
		batch.ValidationChecks,
		fmt.Sprintf("document_set_focus_%s: verify primary %s fields before secondary details", focus, pdfStructuredDocumentSetFocusLabel(focus)),
	)
	batch.AggregationRules = prependPDFStructuredStrings(
		batch.AggregationRules,
		fmt.Sprintf("document_set_focus_%s: prefer %s fields for cross-document comparison", focus, strings.Join(pdfStructuredDocumentSetPriorityFieldNames(focus), ", ")),
	)
	reviewNotes := append([]string(nil), batch.ReviewNotes...)
	lowConfidenceFields := append([]string(nil), batch.LowConfidenceFields...)
	for si := range batch.Sections {
		section := &batch.Sections[si]
		sectionHasPriority := false
		sectionNeedsFocusReview := false
		for fi := range section.Fields {
			field := &section.Fields[fi]
			if !priorityFields[field.Name] {
				continue
			}
			sectionHasPriority = true
			field.CrossDocumentPriority = true
			field.Notes = prependPDFStructuredStrings(
				field.Notes,
				fmt.Sprintf("Cross-document priority field for %s comparison.", pdfStructuredDocumentSetFocusLabel(focus)),
			)
			if !field.Filled {
				field.IssueFlags = prependPDFStructuredStrings(field.IssueFlags, "cross_document_priority_missing")
				field.IssueFlags = dedupePDFVisualStrings(field.IssueFlags)
				field.NeedsReview = true
				sectionNeedsFocusReview = true
				lowConfidenceFields = append(lowConfidenceFields, field.Name)
				reviewNotes = append(reviewNotes, fmt.Sprintf("%s requires review: cross_document_priority_missing", field.Name))
			}
		}
		if !sectionHasPriority {
			continue
		}
		section.QualityChecks = prependPDFStructuredStrings(
			section.QualityChecks,
			fmt.Sprintf("Cross-document priority: validate the primary %s fields before optional detail.", pdfStructuredDocumentSetFocusLabel(focus)),
		)
		if sectionNeedsFocusReview {
			section.ReviewNotes = prependPDFStructuredStrings(
				section.ReviewNotes,
				fmt.Sprintf("Cross-document focus: review primary %s fields first.", pdfStructuredDocumentSetFocusLabel(focus)),
			)
			section.NeedsReview = true
		}
	}
	batch.ReviewNotes = dedupePDFVisualStrings(reviewNotes)
	batch.LowConfidenceFields = dedupePDFVisualStrings(lowConfidenceFields)
	if len(batch.ReviewNotes) > 0 {
		batch.ReviewRequired = true
	}
	return pdfStructuredFinalizeBatchExplainability(batch)
}

func pdfStructuredDocumentSetPriorityFieldSet(focus string) map[string]bool {
	names := pdfStructuredDocumentSetPriorityFieldNames(focus)
	if len(names) == 0 {
		return nil
	}
	out := make(map[string]bool, len(names))
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		out[name] = true
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func pdfStructuredDocumentSetPriorityFieldNames(focus string) []string {
	switch focus {
	case "chart":
		return []string{"slide_title", "chart_type", "main_trend", "axes_or_categories", "series_labels"}
	case "table":
		return []string{"table_subject", "key_headers", "key_rows", "totals_or_comparisons"}
	case "diagram":
		return []string{"diagram_subject", "main_components", "relationships", "legend_or_labels"}
	case "ocr_text":
		return []string{"primary_heading", "section_headers", "captions_or_callouts", "scan_quality_notes"}
	case "layout":
		return []string{"layout_type", "dominant_regions", "key_takeaway"}
	default:
		return nil
	}
}

func pdfStructuredEnsureDocumentSetFocusBatch(
	batches []pdfVisualExtractionBatch,
	profile pdfVisualSignalProfile,
	focus string,
) []pdfVisualExtractionBatch {
	focus = pdfStructuredNormalizeDocumentTarget(focus)
	if focus == "" || focus == "ocr_text" || len(batches) == 0 {
		return batches
	}
	for _, batch := range batches {
		if pdfStructuredNormalizeDocumentTarget(batch.Target) == focus {
			return batches
		}
	}
	pages := pdfStructuredDocumentSetFallbackPages(batches)
	if len(pages) == 0 {
		return batches
	}
	derived := derivePDFVisualSignalProfileForTarget(profile, focus)
	focusBatch := pdfVisualExtractionBatch{
		Target:                 focus,
		Pages:                  pages,
		Priority:               "high",
		Reasons:                []string{fmt.Sprintf("document-set %s focus added a specialized extraction batch from the available pages", pdfStructuredDocumentSetFocusLabel(focus))},
		SummaryMode:            derived.SummaryMode,
		SummaryOutline:         append([]string(nil), derived.SummaryOutline...),
		ExtractionTargets:      append([]string(nil), derived.ExtractionTargets...),
		ExtractionSchema:       append([]pdfVisualExtractionField(nil), derived.ExtractionSchema...),
		ResultSections:         buildPDFVisualResultSections(derived),
		ResultTemplate:         buildPDFVisualResultTemplate(derived.ExtractionSchema),
		ValidationChecks:       buildPDFVisualValidationChecks(derived),
		NormalizationRules:     buildPDFVisualNormalizationRules(derived),
		ExtractionInstructions: buildPDFVisualExtractionInstructions(derived),
		AggregationStrategy:    buildPDFVisualAggregationStrategy(derived),
		AggregationRules:       buildPDFVisualAggregationRules(derived),
	}
	return append([]pdfVisualExtractionBatch{focusBatch}, batches...)
}

func pdfStructuredDocumentSetFallbackPages(batches []pdfVisualExtractionBatch) []int {
	pages := make([]int, 0, len(batches)*2)
	for _, batch := range batches {
		pages = append(pages, batch.Pages...)
	}
	return dedupeSortedPDFPages(pages)
}

func prependPDFStructuredStrings(values []string, leading ...string) []string {
	out := make([]string, 0, len(leading)+len(values))
	for _, item := range leading {
		item = strings.TrimSpace(item)
		if item != "" {
			out = append(out, item)
		}
	}
	for _, item := range values {
		item = strings.TrimSpace(item)
		if item != "" {
			out = append(out, item)
		}
	}
	return dedupePDFVisualStrings(out)
}

func buildPDFStructuredTopDocuments(
	documents []pdfStructuredPayload,
	analysisPlan *pdfAnalysisPlan,
	visualAnalysis *pdfVisualAnalysis,
) []pdfStructuredTopDocument {
	if len(documents) == 0 {
		return nil
	}
	type rankedTopDocument struct {
		item          pdfStructuredTopDocument
		contextScore  int
		completionKey float64
	}
	out := make([]rankedTopDocument, 0, len(documents))
	for _, doc := range documents {
		avgCompletion := averagePDFStructuredCompletionRatio(doc.StructuredBatches)
		out = append(out, rankedTopDocument{
			item: pdfStructuredTopDocument{
				Path:                   doc.Path,
				Status:                 doc.Status,
				ReviewRequired:         doc.ReviewRequired,
				BatchCount:             len(doc.StructuredBatches),
				AverageCompletionRatio: avgCompletion,
				BatchesRequiringReview: pdfStructuredReviewBatchCount(doc),
				LowConfidenceFields:    pdfStructuredReviewLowConfidenceFields(doc),
				SelectionReasonCodes:   pdfStructuredSelectionReasonCodes(doc, analysisPlan, visualAnalysis),
				SelectionReasons:       pdfStructuredSelectionReasons(doc, analysisPlan, visualAnalysis),
				TopNotes:               pdfStructuredTopNotes(doc, analysisPlan, visualAnalysis),
			},
			contextScore:  pdfStructuredDocumentSetAlignmentScore(doc, analysisPlan, visualAnalysis),
			completionKey: avgCompletion,
		})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if pdfStructuredStatusRank(out[i].item.Status) == pdfStructuredStatusRank(out[j].item.Status) {
			if out[i].contextScore == out[j].contextScore {
				if out[i].completionKey == out[j].completionKey {
					if out[i].item.ReviewRequired == out[j].item.ReviewRequired {
						return out[i].item.Path < out[j].item.Path
					}
					return !out[i].item.ReviewRequired && out[j].item.ReviewRequired
				}
				return out[i].completionKey > out[j].completionKey
			}
			return out[i].contextScore > out[j].contextScore
		}
		return pdfStructuredStatusRank(out[i].item.Status) > pdfStructuredStatusRank(out[j].item.Status)
	})
	items := make([]pdfStructuredTopDocument, 0, len(out))
	for _, item := range out {
		items = append(items, item.item)
	}
	return items
}

func averagePDFStructuredCompletionRatio(batches []pdfStructuredBatchResult) float64 {
	if len(batches) == 0 {
		return 0
	}
	total := 0.0
	for _, batch := range batches {
		total += batch.CompletionRatio
	}
	return total / float64(len(batches))
}

func pdfStructuredStatusRank(status string) int {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "complete":
		return 2
	case "partial":
		return 1
	default:
		return 0
	}
}

func pdfStructuredReviewBatchCount(doc pdfStructuredPayload) int {
	if doc.ReviewSummary == nil {
		return 0
	}
	return doc.ReviewSummary.BatchesRequiringReview
}

func pdfStructuredReviewLowConfidenceFields(doc pdfStructuredPayload) int {
	if doc.ReviewSummary == nil {
		return 0
	}
	return doc.ReviewSummary.LowConfidenceFields
}

func pdfStructuredTopNotes(
	doc pdfStructuredPayload,
	analysisPlan *pdfAnalysisPlan,
	visualAnalysis *pdfVisualAnalysis,
) []string {
	notes := make([]string, 0, 4)
	if note := pdfStructuredDocumentSetDocumentNote(doc, analysisPlan, visualAnalysis); note != "" {
		notes = append(notes, note)
	}
	if doc.ReviewSummary != nil {
		notes = append(notes, doc.ReviewSummary.TopNotes...)
	}
	return truncatePDFStructuredNotes(notes, 3)
}

func pdfStructuredSelectionReasons(
	doc pdfStructuredPayload,
	analysisPlan *pdfAnalysisPlan,
	visualAnalysis *pdfVisualAnalysis,
) []string {
	reasons := make([]string, 0, 4)
	if note := pdfStructuredDocumentSetDocumentNote(doc, analysisPlan, visualAnalysis); note != "" {
		reasons = append(reasons, note)
	}
	if status := strings.TrimSpace(doc.Status); status != "" {
		reasons = append(reasons, fmt.Sprintf("Structured extraction status: %s with average completion %.2f.", status, averagePDFStructuredCompletionRatio(doc.StructuredBatches)))
	}
	if doc.ReviewSummary != nil {
		for _, driver := range doc.ReviewSummary.ReviewDrivers {
			driver = strings.TrimSpace(driver)
			if driver != "" {
				reasons = append(reasons, driver)
			}
		}
	}
	return truncatePDFStructuredNotes(reasons, 3)
}

func pdfStructuredSelectionReasonCodes(
	doc pdfStructuredPayload,
	analysisPlan *pdfAnalysisPlan,
	visualAnalysis *pdfVisualAnalysis,
) []string {
	codes := make([]string, 0, 6)
	focus := firstNonEmpty(pdfStructuredDocumentSetFocusTarget(analysisPlan, visualAnalysis), pdfStructuredPayloadFocusTarget(doc))
	if focus != "" && pdfStructuredDocumentSetAlignmentScore(doc, analysisPlan, visualAnalysis) > 0 {
		codes = append(codes, "focus_aligned", "focus_target_"+focus)
	}
	if status := strings.ToLower(strings.TrimSpace(doc.Status)); status != "" {
		codes = append(codes, "status_"+status)
	}
	if doc.ReviewRequired {
		codes = append(codes, "review_required")
	}
	if doc.ReviewSummary != nil && len(doc.ReviewSummary.ReviewReasonCodes) > 0 {
		codes = append(codes, doc.ReviewSummary.ReviewReasonCodes...)
	}
	return truncatePDFStructuredCodes(codes, 6)
}

func pdfStructuredDocumentSetReviewNotes(analysisPlan *pdfAnalysisPlan, visualAnalysis *pdfVisualAnalysis) []string {
	notes := make([]string, 0, 2)
	if note := pdfStructuredDocumentSetFocusNote(analysisPlan, visualAnalysis); note != "" {
		notes = append(notes, note)
	}
	if visualAnalysis != nil {
		if summary := strings.TrimSpace(visualAnalysis.Summary); summary != "" {
			notes = append(notes, "Cross-document visual summary: "+truncateToolText(summary, 220))
		}
	}
	return dedupePDFVisualStrings(notes)
}

func pdfStructuredDocumentSetDocumentNote(
	doc pdfStructuredPayload,
	analysisPlan *pdfAnalysisPlan,
	visualAnalysis *pdfVisualAnalysis,
) string {
	focus := firstNonEmpty(pdfStructuredDocumentSetFocusTarget(analysisPlan, visualAnalysis), pdfStructuredPayloadFocusTarget(doc))
	if focus == "" {
		return ""
	}
	if pdfStructuredDocumentSetAlignmentScore(doc, analysisPlan, visualAnalysis) <= 0 {
		return ""
	}
	return fmt.Sprintf("Cross-document focus: this PDF aligns with the document-set %s priority.", pdfStructuredDocumentSetFocusLabel(focus))
}

func pdfStructuredDocumentSetFocusNote(analysisPlan *pdfAnalysisPlan, visualAnalysis *pdfVisualAnalysis) string {
	focus := pdfStructuredDocumentSetFocusTarget(analysisPlan, visualAnalysis)
	if focus == "" {
		return ""
	}
	return fmt.Sprintf("Cross-document focus: prioritize %s signals across the PDF set.", pdfStructuredDocumentSetFocusDescription(focus))
}

func pdfStructuredDocumentSetAlignmentScore(
	doc pdfStructuredPayload,
	analysisPlan *pdfAnalysisPlan,
	visualAnalysis *pdfVisualAnalysis,
) int {
	focus := firstNonEmpty(pdfStructuredDocumentSetFocusTarget(analysisPlan, visualAnalysis), pdfStructuredPayloadFocusTarget(doc))
	if focus == "" {
		return 0
	}
	score := 0
	for _, batch := range doc.StructuredBatches {
		if pdfStructuredNormalizeDocumentTarget(batch.Target) == focus {
			score += 40
			break
		}
	}
	if doc.VisualAnalysis != nil && pdfStructuredNormalizeDocumentTarget(doc.VisualAnalysis.SignalProfile.PrimaryVisualTarget) == focus {
		score += 20
	}
	switch focus {
	case "chart":
		if doc.MediaProfile.LikelyGraphicDoc || doc.MediaProfile.LikelySlideDeck {
			score += 10
		}
	case "table", "diagram":
		if doc.MediaProfile.LikelyGraphicDoc {
			score += 10
		}
	case "ocr_text":
		if doc.DocumentProfile.LikelyScanned {
			score += 10
		}
	case "layout":
		if !doc.MediaProfile.LikelyGraphicDoc && !doc.DocumentProfile.LikelyScanned {
			score += 5
		}
	}
	return score
}

func pdfStructuredDocumentSetFocusTarget(analysisPlan *pdfAnalysisPlan, visualAnalysis *pdfVisualAnalysis) string {
	if visualAnalysis != nil {
		if target := pdfStructuredNormalizeDocumentTarget(visualAnalysis.SignalProfile.PrimaryVisualTarget); target != "" {
			return target
		}
		switch strings.TrimSpace(visualAnalysis.SignalProfile.SummaryMode) {
		case "slide_chart_summary", "chart_summary":
			return "chart"
		case "table_summary":
			return "table"
		case "diagram_summary":
			return "diagram"
		case "ocr_layout_summary":
			return "ocr_text"
		case "slide_visual_summary", "layout_summary":
			return "layout"
		}
	}
	if analysisPlan == nil {
		return ""
	}
	switch strings.TrimSpace(analysisPlan.Mode) {
	case "vision_ocr":
		return "ocr_text"
	case "text_first":
		return "layout"
	default:
		return ""
	}
}

func pdfStructuredInferFocusTargetFromQuery(query string) string {
	normalized := strings.ToLower(strings.TrimSpace(query))
	if normalized == "" {
		return ""
	}
	switch {
	case strings.Contains(normalized, "table"), strings.Contains(normalized, "表格"):
		return "table"
	case strings.Contains(normalized, "diagram"), strings.Contains(normalized, "图示"):
		return "diagram"
	case strings.Contains(normalized, "ocr"), strings.Contains(normalized, "scan"), strings.Contains(normalized, "扫描"):
		return "ocr_text"
	case strings.Contains(normalized, "layout"), strings.Contains(normalized, "布局"):
		return "layout"
	case strings.Contains(normalized, "chart"), strings.Contains(normalized, "trend"), strings.Contains(normalized, "图表"), strings.Contains(normalized, "可视化"), strings.Contains(normalized, "图"):
		return "chart"
	default:
		return ""
	}
}

func pdfStructuredPayloadFocusTarget(doc pdfStructuredPayload) string {
	if doc.ReviewSummary != nil {
		if focus := strings.TrimSpace(doc.ReviewSummary.FocusTarget); focus != "" {
			return focus
		}
	}
	if doc.ResultEvaluation != nil && doc.ResultEvaluation.Planner != nil {
		if focus := strings.TrimSpace(doc.ResultEvaluation.Planner.FocusTarget); focus != "" {
			return focus
		}
	}
	for _, batch := range doc.StructuredBatches {
		if focus := strings.TrimSpace(batch.FocusTarget); focus != "" {
			return focus
		}
	}
	return ""
}

func pdfStructuredPayloadsFocusTarget(documents []pdfStructuredPayload) string {
	for _, doc := range documents {
		if focus := pdfStructuredPayloadFocusTarget(doc); focus != "" {
			return focus
		}
	}
	return ""
}

func pdfStructuredNormalizeDocumentTarget(target string) string {
	switch strings.TrimSpace(target) {
	case "chart":
		return "chart"
	case "table":
		return "table"
	case "diagram":
		return "diagram"
	case "ocr_text":
		return "ocr_text"
	case "slide_visual", "layout":
		return "layout"
	default:
		return ""
	}
}

func pdfStructuredDocumentSetFocusLabel(target string) string {
	switch target {
	case "chart":
		return "chart"
	case "table":
		return "table"
	case "diagram":
		return "diagram"
	case "ocr_text":
		return "OCR"
	default:
		return "layout"
	}
}

func pdfStructuredDocumentSetFocusDescription(target string) string {
	switch target {
	case "chart":
		return "chart trends, series labels, and numeric comparisons"
	case "table":
		return "table structure, key rows, and totals"
	case "diagram":
		return "diagram components, flows, and labeled relationships"
	case "ocr_text":
		return "OCR-visible headings, captions, and scan-quality cues"
	default:
		return "overall layout and dominant visual regions"
	}
}

func pdfStructuredPrependFocusDriver(drivers []string, focusTarget string) []string {
	if strings.TrimSpace(focusTarget) == "" {
		return drivers
	}
	leading := fmt.Sprintf(
		"Cross-document focus currently prioritizes %s signals.",
		pdfStructuredDocumentSetFocusDescription(focusTarget),
	)
	return truncatePDFStructuredNotes(prependPDFStructuredStrings(drivers, leading), 6)
}

func pdfStructuredPrependFocusReasonCode(codes []string, focusTarget string) []string {
	if strings.TrimSpace(focusTarget) == "" {
		return truncatePDFStructuredCodes(codes, 8)
	}
	return truncatePDFStructuredCodes(append([]string{"focus_target_" + focusTarget}, codes...), 8)
}

func truncatePDFStructuredCodes(codes []string, limit int) []string {
	if len(codes) == 0 {
		return nil
	}
	normalized := make([]string, 0, len(codes))
	for _, code := range codes {
		code = strings.ToLower(strings.TrimSpace(code))
		if code != "" {
			normalized = append(normalized, code)
		}
	}
	normalized = dedupePDFVisualStrings(normalized)
	if limit > 0 && len(normalized) > limit {
		return append([]string(nil), normalized[:limit]...)
	}
	return normalized
}

func pdfStructuredAggregateDocumentReviewReasonCodes(documents []pdfStructuredPayload) []string {
	codes := make([]string, 0, len(documents)*3)
	for _, doc := range documents {
		if doc.ReviewSummary == nil {
			continue
		}
		codes = append(codes, doc.ReviewSummary.ReviewReasonCodes...)
	}
	return truncatePDFStructuredCodes(codes, 8)
}

func pdfStructuredAggregateBatchReviewReasonCodes(batches []pdfStructuredBatchResult) []string {
	codes := make([]string, 0, len(batches)*3)
	for _, batch := range batches {
		if !batch.ReviewRequired {
			continue
		}
		codes = append(codes, pdfStructuredFinalizeBatchExplainability(batch).ReviewReasonCodes...)
	}
	return truncatePDFStructuredCodes(codes, 8)
}

func pdfStructuredAggregateBatchReviewDrivers(batches []pdfStructuredBatchResult) []string {
	drivers := make([]string, 0, len(batches)*2)
	for _, batch := range batches {
		if !batch.ReviewRequired {
			continue
		}
		for _, driver := range pdfStructuredFinalizeBatchExplainability(batch).ReviewDrivers {
			driver = strings.TrimSpace(driver)
			if driver != "" {
				drivers = append(drivers, driver)
			}
		}
	}
	return truncatePDFStructuredNotes(dedupePDFVisualStrings(drivers), 6)
}

func pdfStructuredFinalizeBatchExplainability(batch pdfStructuredBatchResult) pdfStructuredBatchResult {
	batch.ReviewReasonCodes = pdfStructuredReviewReasonCodesForBatch(batch)
	batch.ReviewDrivers = pdfStructuredReviewDriversForBatch(batch)
	return batch
}

func pdfStructuredReviewReasonCodesForBatch(batch pdfStructuredBatchResult) []string {
	codes := make([]string, 0, 8)
	focusTarget := firstNonEmpty(strings.TrimSpace(batch.FocusTarget), pdfStructuredNormalizeDocumentTarget(batch.Target))
	if batch.FocusAligned {
		codes = append(codes, "focus_aligned")
		if focusTarget != "" {
			codes = append(codes, "focus_target_"+focusTarget)
		}
	}
	for _, section := range batch.Sections {
		for _, field := range section.Fields {
			if !field.NeedsReview {
				continue
			}
			codes = append(codes, field.IssueFlags...)
		}
	}
	if batch.ReviewRequired {
		codes = append(codes, "review_required")
	}
	if len(codes) == 0 && batch.ReviewRequired {
		codes = append(codes, "manual_review")
	}
	return truncatePDFStructuredCodes(codes, 8)
}

func pdfStructuredReviewDriversForBatch(batch pdfStructuredBatchResult) []string {
	drivers := make([]string, 0, 6)
	focusTarget := firstNonEmpty(strings.TrimSpace(batch.FocusTarget), pdfStructuredNormalizeDocumentTarget(batch.Target))
	if batch.FocusAligned && focusTarget != "" {
		drivers = append(drivers, fmt.Sprintf("Prioritized because this batch aligns with the cross-document %s focus.", pdfStructuredDocumentSetFocusLabel(focusTarget)))
	}

	priorityMissing := make([]string, 0, 4)
	requiredMissing := make([]string, 0, 4)
	missingEvidence := make([]string, 0, 4)
	summaryInference := make([]string, 0, 4)
	profileInference := make([]string, 0, 4)
	for _, section := range batch.Sections {
		for _, field := range section.Fields {
			if !field.NeedsReview {
				continue
			}
			for _, flag := range field.IssueFlags {
				switch strings.TrimSpace(flag) {
				case "cross_document_priority_missing":
					priorityMissing = append(priorityMissing, field.Name)
				case "required_missing":
					requiredMissing = append(requiredMissing, field.Name)
				case "missing_evidence":
					missingEvidence = append(missingEvidence, field.Name)
				case "summary_inference":
					summaryInference = append(summaryInference, field.Name)
				case "profile_inference", "derived_signal":
					profileInference = append(profileInference, field.Name)
				}
			}
		}
	}
	if len(priorityMissing) > 0 {
		drivers = append(drivers, fmt.Sprintf("Primary %s fields still need review: %s.", pdfStructuredDocumentSetFocusLabel(focusTarget), strings.Join(dedupePDFVisualStrings(priorityMissing), ", ")))
	}
	if len(requiredMissing) > 0 {
		drivers = append(drivers, fmt.Sprintf("Required fields are missing: %s.", strings.Join(dedupePDFVisualStrings(requiredMissing), ", ")))
	}
	if len(missingEvidence) > 0 {
		drivers = append(drivers, fmt.Sprintf("Evidence is still thin for: %s.", strings.Join(dedupePDFVisualStrings(missingEvidence), ", ")))
	}
	if len(summaryInference) > 0 {
		drivers = append(drivers, fmt.Sprintf("Some fields still rely on summary inference: %s.", strings.Join(dedupePDFVisualStrings(summaryInference), ", ")))
	}
	if len(profileInference) > 0 {
		drivers = append(drivers, fmt.Sprintf("Some fields still rely on profile-derived signals: %s.", strings.Join(dedupePDFVisualStrings(profileInference), ", ")))
	}
	if len(drivers) == 0 && batch.ReviewRequired {
		drivers = append(drivers, "This batch still requires manual review.")
	}
	return truncatePDFStructuredNotes(dedupePDFVisualStrings(drivers), 4)
}

func buildPDFStructuredBatchResult(
	batch pdfVisualExtractionBatch,
	pageMap []pdfAnalyzePageItem,
	visual *pdfVisualAnalysis,
	documentSetFocus string,
) pdfStructuredBatchResult {
	resultValues := clonePDFStructuredTemplate(batch.ResultTemplate)
	sections := make([]pdfStructuredSectionResult, 0, len(batch.ResultSections))
	missingRequired := make([]string, 0, 4)
	rendered := make([]string, 0, len(batch.ResultSections))
	filledRequired := 0
	totalRequired := 0
	anyFilledOverall := false
	reviewRequired := false
	reviewNotes := make([]string, 0, len(batch.ResultSections))
	lowConfidenceFields := make([]string, 0, 4)
	excerpts := pdfExcerptsForPages(pageMap, batch.Pages)
	excerptEvidence := pdfEvidenceFromPageMap(pageMap, batch.Pages, 2)
	visualSummary := ""
	signalProfile := pdfVisualSignalProfile{}
	if visual != nil {
		visualSummary = strings.TrimSpace(visual.Summary)
		signalProfile = visual.SignalProfile
	}
	focusAligned := pdfStructuredNormalizeDocumentTarget(batch.Target) == documentSetFocus && strings.TrimSpace(documentSetFocus) != ""
	batchFocusTarget := ""
	if focusAligned {
		batchFocusTarget = documentSetFocus
	}
	priorityFields := pdfStructuredDocumentSetPriorityFieldSet(documentSetFocus)
	for _, section := range batch.ResultSections {
		sectionResult := pdfStructuredSectionResult{
			Name:               section.Name,
			Purpose:            section.Purpose,
			RenderTemplate:     section.RenderTemplate,
			CompletionCriteria: append([]string(nil), section.CompletionCriteria...),
			QualityChecks:      append([]string(nil), section.QualityChecks...),
		}
		fieldResults := make([]pdfStructuredFieldResult, 0, len(section.FieldNames))
		sectionMissing := make([]string, 0, len(section.FieldNames))
		requiredFieldNames := make([]string, 0, len(section.FieldNames))
		requiredFilled := 0
		for _, fieldName := range section.FieldNames {
			schemaField, ok := pdfStructuredSchemaField(batch.ExtractionSchema, fieldName)
			if !ok {
				continue
			}
			candidate := inferPDFStructuredFieldCandidate(
				schemaField,
				batch,
				signalProfile,
				visualSummary,
				excerpts,
				excerptEvidence,
				pdfStructuredFieldInferenceContext{
					FocusAligned:  focusAligned,
					FocusTarget:   documentSetFocus,
					PriorityField: priorityFields[schemaField.Name],
				},
			)
			filled := pdfStructuredHasValue(candidate.Value)
			if filled {
				resultValues[schemaField.Name] = candidate.Value
			}
			fieldResult := pdfStructuredFieldResult{
				Name:                  schemaField.Name,
				Kind:                  schemaField.Kind,
				Required:              schemaField.Required,
				Filled:                filled,
				CrossDocumentPriority: candidatePriorityFlag(candidate, focusAligned, documentSetFocus, schemaField.Name, priorityFields),
				Value:                 resultValues[schemaField.Name],
				Source:                candidate.Source,
				Confidence:            firstNonEmpty(strings.TrimSpace(candidate.Confidence), strings.TrimSpace(section.FieldConfidencePolicy[schemaField.Name])),
				Evidence:              append([]string(nil), candidate.Evidence...),
				Notes:                 append([]string(nil), candidate.Notes...),
			}
			fieldResult.IssueFlags = pdfStructuredFieldIssueFlags(fieldResult)
			fieldResult.NeedsReview = len(fieldResult.IssueFlags) > 0
			fieldResults = append(fieldResults, fieldResult)
			if filled {
				anyFilledOverall = true
			}
			if schemaField.Required {
				requiredFieldNames = append(requiredFieldNames, schemaField.Name)
				if filled {
					requiredFilled++
				}
			}
			if schemaField.Required && !filled {
				sectionMissing = append(sectionMissing, schemaField.Name)
			}
		}
		if len(requiredFieldNames) > 0 {
			switch {
			case section.RequiredAny:
				totalRequired++
				if requiredFilled > 0 {
					filledRequired++
					sectionMissing = nil
					for i := range fieldResults {
						if !fieldResults[i].Required || fieldResults[i].Filled {
							continue
						}
						fieldResults[i].IssueFlags = filterPDFStructuredIssueFlags(fieldResults[i].IssueFlags, "required_missing", "missing_evidence")
						fieldResults[i].NeedsReview = len(fieldResults[i].IssueFlags) > 0
					}
				} else {
					missingRequired = append(missingRequired, requiredFieldNames...)
				}
			default:
				totalRequired += len(requiredFieldNames)
				filledRequired += requiredFilled
				if requiredFilled < len(requiredFieldNames) {
					missingRequired = append(missingRequired, sectionMissing...)
				}
			}
		}
		for _, field := range fieldResults {
			if !field.NeedsReview {
				continue
			}
			reviewRequired = true
			lowConfidenceFields = append(lowConfidenceFields, field.Name)
			reason := strings.Join(field.IssueFlags, ", ")
			if reason == "" {
				reason = "manual_review"
			}
			reviewNotes = append(reviewNotes, fmt.Sprintf("%s requires review: %s", field.Name, reason))
		}
		sectionResult.Fields = fieldResults
		sectionResult.MissingFields = dedupePDFVisualStrings(sectionMissing)
		sectionResult.Rendered = renderPDFStructuredSection(section.RenderTemplate, fieldResults)
		sectionResult.ReviewNotes = pdfStructuredSectionReviewNotes(fieldResults)
		sectionResult.NeedsReview = len(sectionResult.ReviewNotes) > 0
		requiredSatisfied := len(requiredFieldNames) == 0 || (!section.RequiredAny && requiredFilled == len(requiredFieldNames)) || (section.RequiredAny && requiredFilled > 0)
		switch {
		case requiredSatisfied && sectionResult.Rendered != "":
			sectionResult.Status = "complete"
		case len(fieldResults) > 0 && pdfStructuredAnyFilled(fieldResults):
			sectionResult.Status = "partial"
		default:
			sectionResult.Status = "empty"
		}
		if sectionResult.Rendered != "" {
			rendered = append(rendered, sectionResult.Rendered)
		}
		sections = append(sections, sectionResult)
	}
	completionRatio := 1.0
	if totalRequired > 0 {
		completionRatio = float64(filledRequired) / float64(totalRequired)
	}
	status := "empty"
	switch {
	case totalRequired > 0 && completionRatio >= 1 && len(sections) > 0:
		status = "complete"
	case totalRequired > 0 && completionRatio > 0:
		status = "partial"
	case anyFilledOverall || len(rendered) > 0:
		status = "partial"
	}
	return pdfStructuredFinalizeBatchExplainability(pdfStructuredBatchResult{
		Target:                batch.Target,
		Pages:                 append([]int(nil), batch.Pages...),
		Priority:              batch.Priority,
		FocusAligned:          focusAligned,
		FocusTarget:           batchFocusTarget,
		Status:                status,
		Reasons:               append([]string(nil), batch.Reasons...),
		SummaryMode:           batch.SummaryMode,
		CompletionRatio:       completionRatio,
		RenderedOutput:        rendered,
		Result:                resultValues,
		ReviewRequired:        reviewRequired,
		ReviewNotes:           dedupePDFVisualStrings(reviewNotes),
		LowConfidenceFields:   dedupePDFVisualStrings(lowConfidenceFields),
		MissingRequiredFields: dedupePDFVisualStrings(missingRequired),
		ValidationChecks:      append([]string(nil), batch.ValidationChecks...),
		NormalizationRules:    append([]string(nil), batch.NormalizationRules...),
		AggregationStrategy:   batch.AggregationStrategy,
		AggregationRules:      append([]string(nil), batch.AggregationRules...),
		Sections:              sections,
	})
}

func candidatePriorityFlag(candidate pdfStructuredFieldCandidate, focusAligned bool, focusTarget string, fieldName string, priorityFields map[string]bool) bool {
	if !focusAligned {
		return false
	}
	if strings.TrimSpace(focusTarget) == "" {
		return false
	}
	return priorityFields[fieldName]
}

func inferPDFStructuredFieldCandidate(
	field pdfVisualExtractionField,
	batch pdfVisualExtractionBatch,
	profile pdfVisualSignalProfile,
	visualSummary string,
	excerpts []string,
	excerptEvidence []string,
	ctx pdfStructuredFieldInferenceContext,
) pdfStructuredFieldCandidate {
	text := strings.TrimSpace(strings.Join(excerpts, "\n"))
	summaryEvidence := pdfEvidenceForSummaryOnPages(visualSummary, batch.Pages)
	switch field.Name {
	case "slide_title", "table_subject", "diagram_subject", "primary_heading":
		value := pdfFirstHeading(excerpts)
		candidate := pdfStructuredFieldCandidate{
			Value:      value,
			Source:     "page_excerpt",
			Confidence: "excerpt_heading_first",
			Evidence:   pdfLimitEvidence(excerptEvidence, 1),
		}
		if routed, ok := pdfStructuredApplyFocusEvidenceRoute(field.Name, candidate, summaryEvidence, excerptEvidence, ctx, "excerpt_first"); ok {
			return routed
		}
		return candidate
	case "chart_type":
		candidate := pdfStructuredFieldCandidate{
			Value:      inferPDFChartType(visualSummary, text),
			Source:     "visual_summary",
			Confidence: "summary_chart_keyword",
			Evidence:   summaryEvidence,
		}
		if routed, ok := pdfStructuredApplyFocusEvidenceRoute(field.Name, candidate, summaryEvidence, excerptEvidence, ctx, "summary_first"); ok {
			return routed
		}
		return candidate
	case "main_trend", "key_takeaway":
		source := "visual_summary"
		confidence := "summary_primary_sentence"
		evidence := summaryEvidence
		if strings.TrimSpace(visualSummary) == "" {
			source = "page_excerpt"
			confidence = "excerpt_heading_first"
			evidence = pdfLimitEvidence(excerptEvidence, 1)
		}
		candidate := pdfStructuredFieldCandidate{
			Value:      pdfPrimaryTakeaway(visualSummary, text),
			Source:     source,
			Confidence: confidence,
			Evidence:   evidence,
		}
		if routed, ok := pdfStructuredApplyFocusEvidenceRoute(field.Name, candidate, summaryEvidence, excerptEvidence, ctx, "summary_first"); ok {
			return routed
		}
		return candidate
	case "supporting_annotation":
		return pdfStructuredFieldCandidate{
			Value:      pdfSupportingSentence(visualSummary),
			Source:     "visual_summary",
			Confidence: "summary_secondary_sentence",
			Evidence:   summaryEvidence,
		}
	case "key_headers", "axes_or_categories":
		candidate := pdfStructuredFieldCandidate{
			Value:      pdfHeaderCandidates(excerpts),
			Source:     "page_excerpt",
			Confidence: "excerpt_header_candidates",
			Evidence:   pdfLimitEvidence(excerptEvidence, 2),
		}
		if routed, ok := pdfStructuredApplyFocusEvidenceRoute(field.Name, candidate, summaryEvidence, excerptEvidence, ctx, pdfStructuredFocusRouteForField(ctx, field.Name, "excerpt_first")); ok {
			return routed
		}
		return candidate
	case "key_rows":
		candidate := pdfStructuredFieldCandidate{
			Value:      pdfKeyRowCandidates(excerpts),
			Source:     "page_excerpt",
			Confidence: "excerpt_row_candidates",
			Evidence:   pdfLimitEvidence(excerptEvidence, 2),
		}
		if routed, ok := pdfStructuredApplyFocusEvidenceRoute(field.Name, candidate, summaryEvidence, excerptEvidence, ctx, "excerpt_first"); ok {
			return routed
		}
		return candidate
	case "totals_or_comparisons":
		candidate := pdfStructuredFieldCandidate{
			Value:      pdfNumericHighlights(excerpts, visualSummary),
			Source:     "page_excerpt",
			Confidence: "numeric_highlights",
			Evidence:   pdfLimitEvidence(excerptEvidence, 2),
		}
		if routed, ok := pdfStructuredApplyFocusEvidenceRoute(field.Name, candidate, summaryEvidence, excerptEvidence, ctx, pdfStructuredFocusRouteForField(ctx, field.Name, "summary_first")); ok {
			return routed
		}
		return candidate
	case "main_components":
		candidate := pdfStructuredFieldCandidate{
			Value:      pdfKeywordList(visualSummary, text, 4),
			Source:     "visual_summary",
			Confidence: "summary_keyword_components",
			Evidence:   summaryEvidence,
		}
		if routed, ok := pdfStructuredApplyFocusEvidenceRoute(field.Name, candidate, summaryEvidence, excerptEvidence, ctx, "summary_first"); ok {
			return routed
		}
		return candidate
	case "relationships":
		candidate := pdfStructuredFieldCandidate{
			Value:      pdfRelationshipCandidates(visualSummary),
			Source:     "visual_summary",
			Confidence: "summary_relationships",
			Evidence:   summaryEvidence,
		}
		if routed, ok := pdfStructuredApplyFocusEvidenceRoute(field.Name, candidate, summaryEvidence, excerptEvidence, ctx, "summary_first"); ok {
			return routed
		}
		return candidate
	case "legend_or_labels", "series_labels":
		candidate := pdfStructuredFieldCandidate{
			Value:      pdfLabelCandidates(excerpts, visualSummary),
			Source:     "page_excerpt",
			Confidence: "visible_labels",
			Evidence:   pdfLimitEvidence(excerptEvidence, 2),
		}
		if routed, ok := pdfStructuredApplyFocusEvidenceRoute(field.Name, candidate, summaryEvidence, excerptEvidence, ctx, pdfStructuredFocusRouteForField(ctx, field.Name, "summary_first")); ok {
			return routed
		}
		return candidate
	case "section_headers":
		candidate := pdfStructuredFieldCandidate{
			Value:      pdfSectionHeaders(excerpts),
			Source:     "page_excerpt",
			Confidence: "excerpt_sections",
			Evidence:   pdfLimitEvidence(excerptEvidence, 2),
		}
		if routed, ok := pdfStructuredApplyFocusEvidenceRoute(field.Name, candidate, summaryEvidence, excerptEvidence, ctx, "excerpt_first"); ok {
			return routed
		}
		return candidate
	case "captions_or_callouts":
		candidate := pdfStructuredFieldCandidate{
			Value:      pdfCaptionsOrCallouts(excerpts, visualSummary),
			Source:     "page_excerpt",
			Confidence: "visible_callouts",
			Evidence:   pdfLimitEvidence(excerptEvidence, 2),
		}
		if routed, ok := pdfStructuredApplyFocusEvidenceRoute(field.Name, candidate, summaryEvidence, excerptEvidence, ctx, pdfStructuredFocusRouteForField(ctx, field.Name, "excerpt_first")); ok {
			return routed
		}
		return candidate
	case "scan_quality_notes":
		candidate := pdfStructuredFieldCandidate{
			Value:      append([]string(nil), profile.ConfidenceNotes...),
			Source:     "signal_profile",
			Confidence: "signal_profile_confidence_notes",
			Evidence:   append([]string(nil), profile.ConfidenceNotes...),
		}
		if routed, ok := pdfStructuredApplyFocusEvidenceRoute(field.Name, candidate, summaryEvidence, excerptEvidence, ctx, "excerpt_first"); ok {
			return routed
		}
		return candidate
	case "layout_type":
		source := "signal_profile"
		confidence := "signal_profile_layout"
		evidence := []string{fmt.Sprintf("Pages %s signal profile: %s", formatPDFPageSelection(batch.Pages), firstNonEmpty(profile.LayoutType, batch.Target))}
		if strings.TrimSpace(visualSummary) == "" {
			source = "page_excerpt"
			confidence = "excerpt_layout_heuristic"
			evidence = pdfLimitEvidence(excerptEvidence, 1)
		}
		candidate := pdfStructuredFieldCandidate{
			Value:      firstNonEmpty(profile.LayoutType, batch.Target),
			Source:     source,
			Confidence: confidence,
			Evidence:   evidence,
		}
		if routed, ok := pdfStructuredApplyFocusEvidenceRoute(field.Name, candidate, summaryEvidence, excerptEvidence, ctx, "summary_first"); ok {
			return routed
		}
		return candidate
	case "dominant_regions":
		value := append([]string(nil), profile.FocusAreas...)
		source := "signal_profile"
		confidence := "signal_profile_focus_areas"
		evidence := []string{fmt.Sprintf("Pages %s signal profile: %s", formatPDFPageSelection(batch.Pages), strings.Join(profile.FocusAreas, ", "))}
		if strings.TrimSpace(visualSummary) == "" {
			value = pdfDominantRegionCandidates(excerpts)
			if len(value) == 0 {
				value = append([]string(nil), profile.FocusAreas...)
			}
			source = "page_excerpt"
			confidence = "excerpt_region_candidates"
			evidence = pdfLimitEvidence(excerptEvidence, 2)
		}
		candidate := pdfStructuredFieldCandidate{
			Value:      value,
			Source:     source,
			Confidence: confidence,
			Evidence:   evidence,
		}
		if routed, ok := pdfStructuredApplyFocusEvidenceRoute(field.Name, candidate, summaryEvidence, excerptEvidence, ctx, "summary_first"); ok {
			return routed
		}
		return candidate
	default:
		return pdfStructuredFieldCandidate{
			Value: defaultPDFVisualTemplateValue(field.Kind),
		}
	}
}

func pdfStructuredApplyFocusEvidenceRoute(
	fieldName string,
	candidate pdfStructuredFieldCandidate,
	summaryEvidence []string,
	excerptEvidence []string,
	ctx pdfStructuredFieldInferenceContext,
	route string,
) (pdfStructuredFieldCandidate, bool) {
	if !ctx.FocusAligned || !ctx.PriorityField {
		return candidate, false
	}
	route = strings.ToLower(strings.TrimSpace(route))
	switch route {
	case "summary_first":
		if len(summaryEvidence) > 0 {
			candidate.Source = "cross_document_visual_summary"
			candidate.Confidence = "cross_document_priority_summary_first"
			candidate.Evidence = pdfStructuredMergeEvidence(summaryEvidence, excerptEvidence)
			candidate.Notes = prependPDFStructuredStrings(candidate.Notes, fmt.Sprintf("Cross-document focus routed %s through summary-first evidence.", fieldName))
			return candidate, true
		}
		if len(excerptEvidence) > 0 {
			candidate.Source = "cross_document_page_excerpt"
			candidate.Confidence = "cross_document_priority_excerpt_fallback"
			candidate.Evidence = pdfStructuredMergeEvidence(excerptEvidence, summaryEvidence)
			candidate.Notes = prependPDFStructuredStrings(candidate.Notes, fmt.Sprintf("Cross-document focus fell back to excerpt evidence for %s.", fieldName))
			return candidate, true
		}
	case "excerpt_first":
		if len(excerptEvidence) > 0 {
			candidate.Source = "cross_document_page_excerpt"
			candidate.Confidence = "cross_document_priority_excerpt_first"
			candidate.Evidence = pdfStructuredMergeEvidence(excerptEvidence, summaryEvidence)
			candidate.Notes = prependPDFStructuredStrings(candidate.Notes, fmt.Sprintf("Cross-document focus routed %s through excerpt-first evidence.", fieldName))
			return candidate, true
		}
		if len(summaryEvidence) > 0 {
			candidate.Source = "cross_document_visual_summary"
			candidate.Confidence = "cross_document_priority_summary_fallback"
			candidate.Evidence = pdfStructuredMergeEvidence(summaryEvidence, excerptEvidence)
			candidate.Notes = prependPDFStructuredStrings(candidate.Notes, fmt.Sprintf("Cross-document focus fell back to summary evidence for %s.", fieldName))
			return candidate, true
		}
	}
	candidate.Notes = prependPDFStructuredStrings(candidate.Notes, fmt.Sprintf("Cross-document focus kept the default evidence path for %s.", fieldName))
	return candidate, true
}

func pdfStructuredFocusRouteForField(ctx pdfStructuredFieldInferenceContext, fieldName string, defaultRoute string) string {
	if !ctx.FocusAligned || !ctx.PriorityField {
		return defaultRoute
	}
	switch strings.TrimSpace(ctx.FocusTarget) {
	case "chart", "diagram", "layout":
		return "summary_first"
	case "table", "ocr_text":
		return "excerpt_first"
	default:
		return defaultRoute
	}
}

func pdfStructuredMergeEvidence(primary []string, secondary []string) []string {
	merged := make([]string, 0, len(primary)+len(secondary))
	for _, item := range primary {
		item = strings.TrimSpace(item)
		if item != "" {
			merged = append(merged, item)
		}
	}
	for _, item := range secondary {
		item = strings.TrimSpace(item)
		if item != "" {
			merged = append(merged, item)
		}
	}
	return dedupePDFVisualStrings(merged)
}

func pdfStructuredSchemaField(schema []pdfVisualExtractionField, name string) (pdfVisualExtractionField, bool) {
	for _, item := range schema {
		if item.Name == name {
			return item, true
		}
	}
	return pdfVisualExtractionField{}, false
}

func clonePDFStructuredTemplate(template map[string]any) map[string]any {
	if len(template) == 0 {
		return map[string]any{}
	}
	out := make(map[string]any, len(template))
	for k, v := range template {
		switch typed := v.(type) {
		case []string:
			out[k] = append([]string(nil), typed...)
		default:
			out[k] = typed
		}
	}
	return out
}

func pdfVisualSignalProfileIsEmpty(profile pdfVisualSignalProfile) bool {
	return strings.TrimSpace(profile.LayoutType) == "" &&
		strings.TrimSpace(profile.SummaryMode) == "" &&
		strings.TrimSpace(profile.PrimaryVisualTarget) == "" &&
		!profile.ChartLike &&
		!profile.TableLike &&
		!profile.DiagramLike &&
		!profile.ImageDocument &&
		!profile.TextSparse &&
		len(profile.SummaryOutline) == 0 &&
		len(profile.ExtractionTargets) == 0 &&
		len(profile.ExtractionSchema) == 0 &&
		len(profile.ConfidenceNotes) == 0 &&
		len(profile.FocusAreas) == 0 &&
		len(profile.SuggestedFollowUps) == 0
}

func renderPDFStructuredSection(template string, fields []pdfStructuredFieldResult) string {
	rendered := strings.TrimSpace(template)
	if rendered == "" {
		return ""
	}
	for _, field := range fields {
		token := "{{" + field.Name + "}}"
		rendered = strings.ReplaceAll(rendered, token, pdfStructuredRenderValue(field.Value))
	}
	rendered = strings.TrimSpace(strings.ReplaceAll(rendered, "  ", " "))
	rendered = strings.TrimSpace(strings.Trim(rendered, "|"))
	if !strings.ContainsAny(rendered, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789") {
		return ""
	}
	return rendered
}

func pdfStructuredRenderValue(value any) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case []string:
		filtered := make([]string, 0, len(typed))
		for _, item := range typed {
			item = strings.TrimSpace(item)
			if item != "" {
				filtered = append(filtered, item)
			}
		}
		return strings.Join(filtered, ", ")
	default:
		return ""
	}
}

func pdfStructuredAnyFilled(fields []pdfStructuredFieldResult) bool {
	for _, field := range fields {
		if field.Filled {
			return true
		}
	}
	return false
}

func pdfStructuredFieldIssueFlags(field pdfStructuredFieldResult) []string {
	flags := make([]string, 0, 4)
	switch {
	case field.Required && !field.Filled:
		flags = append(flags, "required_missing")
	case !field.Filled:
		return nil
	}
	if len(field.Evidence) == 0 {
		flags = append(flags, "missing_evidence")
	}
	confidence := strings.TrimSpace(strings.ToLower(field.Confidence))
	source := strings.TrimSpace(strings.ToLower(field.Source))
	switch {
	case strings.HasPrefix(confidence, "signal_profile"):
		flags = append(flags, "profile_inference")
	case strings.HasPrefix(confidence, "summary_"):
		flags = append(flags, "summary_inference")
	}
	if source == "signal_profile" {
		flags = append(flags, "derived_signal")
	}
	return dedupePDFVisualStrings(flags)
}

func pdfStructuredSectionReviewNotes(fields []pdfStructuredFieldResult) []string {
	out := make([]string, 0, len(fields))
	for _, field := range fields {
		if !field.NeedsReview {
			continue
		}
		reason := strings.Join(field.IssueFlags, ", ")
		if reason == "" {
			reason = "manual_review"
		}
		out = append(out, fmt.Sprintf("%s: %s", field.Name, reason))
	}
	return dedupePDFVisualStrings(out)
}

func filterPDFStructuredIssueFlags(flags []string, drop ...string) []string {
	if len(flags) == 0 || len(drop) == 0 {
		return flags
	}
	blocked := make(map[string]struct{}, len(drop))
	for _, item := range drop {
		item = strings.TrimSpace(item)
		if item != "" {
			blocked[item] = struct{}{}
		}
	}
	out := make([]string, 0, len(flags))
	for _, flag := range flags {
		if _, ok := blocked[flag]; ok {
			continue
		}
		out = append(out, flag)
	}
	return dedupePDFVisualStrings(out)
}

func truncatePDFStructuredNotes(notes []string, limit int) []string {
	deduped := dedupePDFVisualStrings(notes)
	if limit <= 0 || len(deduped) <= limit {
		return deduped
	}
	return deduped[:limit]
}

func pdfStructuredHasValue(value any) bool {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed) != ""
	case []string:
		for _, item := range typed {
			if strings.TrimSpace(item) != "" {
				return true
			}
		}
		return false
	default:
		return false
	}
}

func pdfExcerptsForPages(pageMap []pdfAnalyzePageItem, pages []int) []string {
	if len(pageMap) == 0 || len(pages) == 0 {
		return nil
	}
	pageSet := make(map[int]struct{}, len(pages))
	for _, page := range pages {
		pageSet[page] = struct{}{}
	}
	out := make([]string, 0, len(pages))
	for _, item := range pageMap {
		if _, ok := pageSet[item.Page]; !ok {
			continue
		}
		excerpt := strings.TrimSpace(item.Excerpt)
		if excerpt != "" {
			out = append(out, excerpt)
		}
	}
	return out
}

func allPDFPagesFromMap(pageMap []pdfAnalyzePageItem) []int {
	if len(pageMap) == 0 {
		return nil
	}
	out := make([]int, 0, len(pageMap))
	for _, item := range pageMap {
		out = append(out, item.Page)
	}
	sort.Ints(out)
	return out
}

func pdfFirstHeading(excerpts []string) string {
	for _, excerpt := range excerpts {
		for _, line := range strings.Split(excerpt, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			return truncateToolText(line, 120)
		}
	}
	return ""
}

func inferPDFChartType(summary string, text string) string {
	lower := strings.ToLower(summary + "\n" + text)
	switch {
	case strings.Contains(lower, "bar"):
		return "bar"
	case strings.Contains(lower, "line"):
		return "line"
	case strings.Contains(lower, "pie"):
		return "pie"
	case strings.Contains(lower, "scatter"):
		return "scatter"
	case strings.Contains(lower, "chart"), strings.Contains(lower, "graph"), strings.Contains(lower, "plot"):
		return "chart"
	default:
		return ""
	}
}

func pdfPrimaryTakeaway(summary string, text string) string {
	for _, sentence := range pdfSentenceCandidates(summary) {
		if strings.TrimSpace(sentence) != "" {
			return truncateToolText(sentence, 180)
		}
	}
	return truncateToolText(pdfFirstHeading([]string{text}), 180)
}

func pdfSupportingSentence(summary string) string {
	candidates := pdfSentenceCandidates(summary)
	if len(candidates) > 1 {
		return truncateToolText(candidates[1], 160)
	}
	return ""
}

func pdfHeaderCandidates(excerpts []string) []string {
	out := make([]string, 0, 4)
	for _, excerpt := range excerpts {
		line := pdfFirstHeading([]string{excerpt})
		if line == "" {
			continue
		}
		for _, part := range strings.FieldsFunc(line, func(r rune) bool {
			return r == '|' || r == ',' || r == ';'
		}) {
			part = strings.TrimSpace(part)
			if part != "" && len(part) <= 40 {
				out = append(out, part)
			}
		}
	}
	return dedupePDFVisualStrings(out)
}

func pdfDominantRegionCandidates(excerpts []string) []string {
	out := make([]string, 0, 4)
	if len(excerpts) == 0 {
		return nil
	}
	firstHeading := pdfFirstHeading(excerpts)
	if firstHeading != "" {
		out = append(out, "heading_region")
	}
	for _, excerpt := range excerpts {
		lines := strings.Split(excerpt, "\n")
		if len(lines) > 1 {
			out = append(out, "body_text_region")
			break
		}
	}
	if len(out) == 0 {
		out = append(out, "single_text_region")
	}
	return dedupePDFVisualStrings(out)
}

func pdfKeyRowCandidates(excerpts []string) []string {
	out := make([]string, 0, 4)
	for _, excerpt := range excerpts {
		lines := strings.Split(excerpt, "\n")
		for _, line := range lines[1:] {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			if strings.ContainsAny(line, "0123456789") || strings.Contains(line, "|") {
				out = append(out, truncateToolText(line, 80))
			}
		}
	}
	return dedupePDFVisualStrings(out)
}

func pdfNumericHighlights(excerpts []string, summary string) []string {
	out := make([]string, 0, 4)
	for _, sentence := range pdfSentenceCandidates(summary) {
		if strings.ContainsAny(sentence, "0123456789%") {
			out = append(out, truncateToolText(sentence, 120))
		}
	}
	for _, excerpt := range excerpts {
		for _, line := range strings.Split(excerpt, "\n") {
			line = strings.TrimSpace(line)
			if strings.ContainsAny(line, "0123456789%") {
				out = append(out, truncateToolText(line, 120))
			}
		}
	}
	return dedupePDFVisualStrings(out)
}

func pdfKeywordList(summary string, text string, limit int) []string {
	source := strings.TrimSpace(summary)
	if source == "" {
		source = strings.TrimSpace(text)
	}
	if source == "" {
		return nil
	}
	tokens := strings.FieldsFunc(source, func(r rune) bool {
		return r == ',' || r == ';' || r == ':' || r == '\n'
	})
	out := make([]string, 0, limit)
	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if token == "" || len(token) > 60 {
			continue
		}
		out = append(out, token)
		if len(out) >= limit {
			break
		}
	}
	return dedupePDFVisualStrings(out)
}

func pdfRelationshipCandidates(summary string) []string {
	out := make([]string, 0, 3)
	for _, sentence := range pdfSentenceCandidates(summary) {
		lower := strings.ToLower(sentence)
		if strings.Contains(lower, "->") || strings.Contains(lower, "flow") || strings.Contains(lower, "relationship") || strings.Contains(lower, "connect") {
			out = append(out, truncateToolText(sentence, 140))
		}
	}
	return dedupePDFVisualStrings(out)
}

func pdfLabelCandidates(excerpts []string, summary string) []string {
	out := pdfHeaderCandidates(excerpts)
	if len(out) == 0 {
		out = pdfKeywordList(summary, "", 4)
	}
	return out
}

func pdfSectionHeaders(excerpts []string) []string {
	out := make([]string, 0, 5)
	for _, excerpt := range excerpts {
		for _, line := range strings.Split(excerpt, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			if len(line) <= 60 {
				out = append(out, truncateToolText(line, 60))
			}
			if len(out) >= 5 {
				return dedupePDFVisualStrings(out)
			}
		}
	}
	return dedupePDFVisualStrings(out)
}

func pdfCaptionsOrCallouts(excerpts []string, summary string) []string {
	out := make([]string, 0, 3)
	for _, excerpt := range excerpts {
		for _, line := range strings.Split(excerpt, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			if strings.Contains(line, ":") || strings.Contains(strings.ToLower(line), "note") {
				out = append(out, truncateToolText(line, 100))
			}
		}
	}
	if len(out) == 0 {
		if supporting := pdfSupportingSentence(summary); supporting != "" {
			out = append(out, supporting)
		}
	}
	return dedupePDFVisualStrings(out)
}

func pdfEvidenceFromExcerpts(excerpts []string, limit int) []string {
	if limit <= 0 {
		limit = 1
	}
	out := make([]string, 0, limit)
	for _, excerpt := range excerpts {
		excerpt = strings.TrimSpace(excerpt)
		if excerpt == "" {
			continue
		}
		out = append(out, truncateToolText(excerpt, 160))
		if len(out) >= limit {
			break
		}
	}
	return out
}

func pdfEvidenceForSummary(summary string) []string {
	summary = strings.TrimSpace(summary)
	if summary == "" {
		return nil
	}
	return []string{truncateToolText(summary, 180)}
}

func pdfEvidenceForSummaryOnPages(summary string, pages []int) []string {
	summary = strings.TrimSpace(summary)
	if summary == "" {
		return nil
	}
	label := "Page"
	pageText := formatPDFPageSelection(pages)
	if strings.Contains(pageText, ",") {
		label = "Pages"
	}
	if pageText == "" {
		return []string{truncateToolText(summary, 180)}
	}
	return []string{fmt.Sprintf("%s %s visual summary: %s", label, pageText, truncateToolText(summary, 180))}
}

func pdfEvidenceFromPageMap(pageMap []pdfAnalyzePageItem, pages []int, limit int) []string {
	items := pdfAnalyzePageItemsForPages(pageMap, pages, limit)
	if len(items) == 0 {
		return nil
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		excerpt := strings.TrimSpace(item.Excerpt)
		if excerpt == "" {
			continue
		}
		out = append(out, fmt.Sprintf("Page %d: %s", item.Page, truncateToolText(excerpt, 160)))
	}
	return out
}

func pdfLimitEvidence(evidence []string, limit int) []string {
	if limit <= 0 {
		limit = 1
	}
	out := make([]string, 0, minInt(limit, len(evidence)))
	for _, item := range evidence {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		out = append(out, item)
		if len(out) >= limit {
			break
		}
	}
	return out
}

func pdfSentenceCandidates(text string) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	replacer := strings.NewReplacer("\n", ". ", "!", ".", "?", ".")
	text = replacer.Replace(text)
	parts := strings.Split(text, ".")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
