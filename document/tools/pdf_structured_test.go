package tools

import (
	"context"
	"strings"
	"testing"
)

func TestBuildPDFStructuredBatchResult_RequiredAnyCountsAsSingleRequirement(t *testing.T) {
	batch := pdfVisualExtractionBatch{
		Target: "chart",
		Pages:  []int{1},
		ExtractionSchema: []pdfVisualExtractionField{
			{Name: "slide_title", Kind: "string", Required: true},
			{Name: "secondary_claim", Kind: "string", Required: true},
		},
		ResultTemplate: map[string]any{
			"slide_title":     "",
			"secondary_claim": "",
		},
		ResultSections: []pdfVisualResultSection{
			{
				Name:           "headline",
				FieldNames:     []string{"slide_title", "secondary_claim"},
				RequiredAny:    true,
				RenderTemplate: "Slide title: {{slide_title}}",
			},
		},
	}
	pageMap := []pdfAnalyzePageItem{
		{Page: 1, Excerpt: "Quarterly review\nRevenue is up quarter over quarter."},
	}

	result := buildPDFStructuredBatchResult(batch, pageMap, nil, "")
	if result.Status != "complete" {
		t.Fatalf("expected complete status, got %#v", result.Status)
	}
	if result.CompletionRatio != 1 {
		t.Fatalf("expected completion_ratio=1, got %#v", result.CompletionRatio)
	}
	if len(result.MissingRequiredFields) != 0 {
		t.Fatalf("expected no missing required fields, got %#v", result.MissingRequiredFields)
	}
	if len(result.Sections) != 1 {
		t.Fatalf("expected one section, got %#v", result.Sections)
	}
	if result.Sections[0].Status != "complete" {
		t.Fatalf("expected complete section, got %#v", result.Sections[0].Status)
	}
	if len(result.Sections[0].MissingFields) != 0 {
		t.Fatalf("expected no missing fields in required_any section, got %#v", result.Sections[0].MissingFields)
	}
	if got := result.Result["slide_title"]; got != "Quarterly review" {
		t.Fatalf("expected slide_title to be filled, got %#v", got)
	}
	if result.ReviewRequired {
		t.Fatalf("expected no batch review required, got %#v", result.ReviewNotes)
	}
}

func TestBuildPDFStructuredBatchResult_RequiredFieldsRemainPartialWhenMissing(t *testing.T) {
	batch := pdfVisualExtractionBatch{
		Target: "chart",
		Pages:  []int{1},
		ExtractionSchema: []pdfVisualExtractionField{
			{Name: "slide_title", Kind: "string", Required: true},
			{Name: "secondary_claim", Kind: "string", Required: true},
		},
		ResultTemplate: map[string]any{
			"slide_title":     "",
			"secondary_claim": "",
		},
		ResultSections: []pdfVisualResultSection{
			{
				Name:           "headline",
				FieldNames:     []string{"slide_title", "secondary_claim"},
				RequiredAny:    false,
				RenderTemplate: "Slide title: {{slide_title}}",
			},
		},
	}
	pageMap := []pdfAnalyzePageItem{
		{Page: 1, Excerpt: "Quarterly review\nRevenue is up quarter over quarter."},
	}

	result := buildPDFStructuredBatchResult(batch, pageMap, nil, "")
	if result.Status != "partial" {
		t.Fatalf("expected partial status, got %#v", result.Status)
	}
	if result.CompletionRatio != 0.5 {
		t.Fatalf("expected completion_ratio=0.5, got %#v", result.CompletionRatio)
	}
	if len(result.MissingRequiredFields) != 1 || result.MissingRequiredFields[0] != "secondary_claim" {
		t.Fatalf("expected missing secondary_claim, got %#v", result.MissingRequiredFields)
	}
	if len(result.Sections) != 1 {
		t.Fatalf("expected one section, got %#v", result.Sections)
	}
	if result.Sections[0].Status != "partial" {
		t.Fatalf("expected partial section, got %#v", result.Sections[0].Status)
	}
	if len(result.Sections[0].MissingFields) != 1 || result.Sections[0].MissingFields[0] != "secondary_claim" {
		t.Fatalf("expected section missing secondary_claim, got %#v", result.Sections[0].MissingFields)
	}
	if result.Sections[0].Rendered == "" {
		t.Fatalf("expected rendered section output, got %#v", result.Sections[0].Rendered)
	}
	if !result.ReviewRequired {
		t.Fatalf("expected batch review required when required field is missing")
	}
	if len(result.LowConfidenceFields) == 0 || result.LowConfidenceFields[0] != "secondary_claim" {
		t.Fatalf("expected secondary_claim in low confidence fields, got %#v", result.LowConfidenceFields)
	}
	if !result.Sections[0].NeedsReview {
		t.Fatalf("expected section review required")
	}
	if len(result.Sections[0].ReviewNotes) == 0 {
		t.Fatalf("expected section review notes, got %#v", result.Sections[0].ReviewNotes)
	}
}

func TestBuildPDFStructuredBatchResult_SummaryOnlyFieldsMarkedForReview(t *testing.T) {
	batch := pdfVisualExtractionBatch{
		Target: "layout",
		Pages:  []int{1},
		ExtractionSchema: []pdfVisualExtractionField{
			{Name: "layout_type", Kind: "string", Required: true},
			{Name: "dominant_regions", Kind: "string_list", Required: true},
		},
		ResultTemplate: map[string]any{
			"layout_type":      "",
			"dominant_regions": []string{},
		},
		ResultSections: []pdfVisualResultSection{
			{
				Name:           "layout",
				FieldNames:     []string{"layout_type", "dominant_regions"},
				RequiredAny:    false,
				RenderTemplate: "Layout: {{layout_type}} | Regions: {{dominant_regions}}",
			},
		},
	}

	result := buildPDFStructuredBatchResult(batch, nil, &pdfVisualAnalysis{
		Summary: "Graphic report with chart area and notes sidebar.",
		SignalProfile: pdfVisualSignalProfile{
			LayoutType: "graphic_report",
			FocusAreas: []string{"chart area", "notes sidebar"},
		},
	}, "")

	if result.Status != "complete" {
		t.Fatalf("expected complete status from signal_profile-backed fields, got %#v", result.Status)
	}
	if !result.ReviewRequired {
		t.Fatalf("expected review required for signal-profile-derived fields")
	}
	if len(result.LowConfidenceFields) != 2 {
		t.Fatalf("expected low confidence fields for both derived fields, got %#v", result.LowConfidenceFields)
	}
	fields := result.Sections[0].Fields
	if len(fields) != 2 {
		t.Fatalf("expected two fields, got %#v", fields)
	}
	if !fields[0].NeedsReview || !fields[1].NeedsReview {
		t.Fatalf("expected both fields to require review, got %#v", fields)
	}
	if len(fields[0].IssueFlags) == 0 || len(fields[1].IssueFlags) == 0 {
		t.Fatalf("expected issue flags on derived fields, got %#v", fields)
	}
}

func TestBuildPDFStructuredPayload_SurfacesReviewSummary(t *testing.T) {
	artifacts := pdfAnalysisArtifacts{
		DisplayPath: "sample.pdf",
		Metadata:    PDFMetadataResult{PageCount: 1},
		BackendStatus: pdfBackendStatus{
			ExtractBackend: "stub",
		},
		PageMap: []pdfAnalyzePageItem{
			{Page: 1, Excerpt: "Graphic report\nRevenue trend and notes sidebar.", Chars: 64},
		},
		DocumentProfile: pdfDocumentProfile{HasOutline: false},
		MediaProfile:    pdfMediaProfile{LikelyGraphicDoc: true},
		AnalysisPlan:    pdfAnalysisPlan{Mode: "hybrid_vision_text"},
		VisualAnalysis: &pdfVisualAnalysis{
			Summary: "Graphic report with chart area and notes sidebar.",
			SignalProfile: pdfVisualSignalProfile{
				LayoutType: "graphic_report",
				FocusAreas: []string{"chart area", "notes sidebar"},
			},
		},
	}

	payload := buildPDFStructuredPayload(artifacts, "revenue", false, nil, nil)
	if !payload.ReviewRequired {
		t.Fatalf("expected payload review_required")
	}
	if payload.ReviewSummary == nil {
		t.Fatalf("expected review_summary")
	}
	if payload.ReviewSummary.BatchesRequiringReview == 0 {
		t.Fatalf("expected batches_requiring_review > 0, got %#v", payload.ReviewSummary)
	}
	if payload.ReviewSummary.LowConfidenceFields == 0 {
		t.Fatalf("expected low_confidence_fields > 0, got %#v", payload.ReviewSummary)
	}
	if len(payload.ReviewSummary.BatchTargets) == 0 {
		t.Fatalf("expected batch targets in review summary, got %#v", payload.ReviewSummary)
	}
	if len(payload.ReviewSummary.TopNotes) == 0 {
		t.Fatalf("expected top notes in review summary, got %#v", payload.ReviewSummary)
	}
	if len(payload.ReviewSummary.ReviewDrivers) == 0 {
		t.Fatalf("expected review drivers in review summary, got %#v", payload.ReviewSummary)
	}
	if len(payload.ReviewSummary.ReviewReasonCodes) == 0 {
		t.Fatalf("expected review reason codes in review summary, got %#v", payload.ReviewSummary)
	}
}

func TestApplyPDFUnifiedFocusToStructuredBatches_PrunesLogisticsPagesForFieldCompare(t *testing.T) {
	batches := []pdfVisualExtractionBatch{
		{
			Target:  "ocr_text",
			Pages:   []int{1, 2, 3},
			Reasons: []string{"default batch pages"},
		},
	}
	focus := pdfUnifiedDocumentFocus{
		Primary: &pdfUnifiedSegment{
			Kind:  pdfUnifiedSegmentBusinessDoc,
			Pages: []int{2},
		},
		Supporting: []pdfUnifiedSegment{
			{Kind: pdfUnifiedSegmentLogisticsDoc, Pages: []int{1}},
			{Kind: pdfUnifiedSegmentSignatureStamp, Pages: []int{3}},
		},
	}

	got := applyPDFUnifiedFocusToStructuredBatches(batches, focus, pdfUnifiedQueryClassFieldCompare, true)
	if len(got) != 1 {
		t.Fatalf("expected one batch after focus prune, got %#v", got)
	}
	if len(got[0].Pages) != 2 || got[0].Pages[0] != 2 || got[0].Pages[1] != 3 {
		t.Fatalf("expected focused pages [2 3], got %#v", got[0].Pages)
	}
	if len(got[0].Reasons) == 0 || !strings.Contains(strings.ToLower(got[0].Reasons[0]), "subdocument focus limited") {
		t.Fatalf("expected focus prune reason, got %#v", got[0].Reasons)
	}
}

func TestBuildPDFStructuredPayload_BuildsResultEvaluation(t *testing.T) {
	artifacts := pdfAnalysisArtifacts{
		DisplayPath: "sample.pdf",
		Metadata:    PDFMetadataResult{PageCount: 1},
		BackendStatus: pdfBackendStatus{
			ExtractBackend: "stub",
		},
		PageMap: []pdfAnalyzePageItem{
			{Page: 1, Excerpt: "Quarterly review\nRevenue is up quarter over quarter.", Chars: 64},
		},
		AnalysisPlan: pdfAnalysisPlan{Mode: "hybrid_vision_text"},
		VisualAnalysis: &pdfVisualAnalysis{
			Status: "success",
			SignalProfile: pdfVisualSignalProfile{
				PrimaryVisualTarget: "chart",
			},
			ExtractionBatches: []pdfVisualExtractionBatch{
				{
					Target: "chart",
					Pages:  []int{1},
					ExtractionSchema: []pdfVisualExtractionField{
						{Name: "slide_title", Kind: "string", Required: true},
						{Name: "secondary_claim", Kind: "string", Required: true},
					},
					ResultTemplate: map[string]any{
						"slide_title":     "",
						"secondary_claim": "",
					},
					ResultSections: []pdfVisualResultSection{
						{
							Name:           "headline",
							FieldNames:     []string{"slide_title", "secondary_claim"},
							RenderTemplate: "Slide title: {{slide_title}}",
						},
					},
				},
			},
		},
	}

	payload := buildPDFStructuredPayload(artifacts, "revenue", false, nil, nil)
	if payload.ResultEvaluation == nil {
		t.Fatalf("expected result_evaluation")
	}
	if payload.ResultEvaluation.Planner == nil {
		t.Fatalf("expected planner evaluation, got %#v", payload.ResultEvaluation)
	}
	if payload.ResultEvaluation.Planner.VisualStatus != "success" {
		t.Fatalf("expected planner visual_status=success, got %#v", payload.ResultEvaluation.Planner)
	}
	if payload.ResultEvaluation.Planner.PlannedBatchCount != 1 || len(payload.ResultEvaluation.Planner.PlannedBatchTargets) != 1 || payload.ResultEvaluation.Planner.PlannedBatchTargets[0] != "chart" {
		t.Fatalf("expected chart planner snapshot, got %#v", payload.ResultEvaluation.Planner)
	}
	if payload.ResultEvaluation.Executor == nil {
		t.Fatalf("expected executor evaluation, got %#v", payload.ResultEvaluation)
	}
	if payload.ResultEvaluation.Executor.Status != "review_required" {
		t.Fatalf("expected executor status=review_required, got %#v", payload.ResultEvaluation.Executor)
	}
	if payload.ResultEvaluation.Executor.BatchCount != 1 || payload.ResultEvaluation.Executor.ReviewBatchCount != 1 {
		t.Fatalf("expected one review batch, got %#v", payload.ResultEvaluation.Executor)
	}
	if payload.ResultEvaluation.Executor.MissingRequiredFieldCount != 1 || payload.ResultEvaluation.Executor.LowConfidenceFieldCount != 1 {
		t.Fatalf("expected missing/low-confidence counts to surface, got %#v", payload.ResultEvaluation.Executor)
	}
}

func TestBuildPDFStructuredPayload_SynthesizesDocumentSetFocusBatchFromOCRFallback(t *testing.T) {
	profile := pdfVisualSignalProfile{
		LayoutType:          "slide_deck",
		SummaryMode:         "ocr_layout_summary",
		PrimaryVisualTarget: "ocr_text",
		TextSparse:          true,
		ImageDocument:       true,
	}
	artifacts := pdfAnalysisArtifacts{
		DisplayPath: "alpha.pdf",
		Metadata:    PDFMetadataResult{PageCount: 2},
		BackendStatus: pdfBackendStatus{
			ExtractBackend: "stub",
		},
		PageMap: []pdfAnalyzePageItem{
			{Page: 1, Excerpt: "Executive slide\nRevenue chart climbs sharply after launch. North, South, West.", Chars: 22},
			{Page: 2, Excerpt: "Chart notes\nQ2 to Q4 trend improves and margin expands.", Chars: 26},
		},
		MediaProfile: pdfMediaProfile{
			LikelyGraphicDoc: true,
			LikelySlideDeck:  true,
		},
		AnalysisPlan: pdfAnalysisPlan{Mode: "hybrid_vision_text"},
		VisualAnalysis: &pdfVisualAnalysis{
			Status:        "success",
			SignalProfile: profile,
			ExtractionBatches: buildPDFVisualExtractionBatches([]pdfVisualPageTarget{
				{Page: 1, Target: "ocr_text", Priority: "high", Reason: "text is sparse or OCR-dependent on this page"},
				{Page: 2, Target: "ocr_text", Priority: "high", Reason: "text is sparse or OCR-dependent on this page"},
			}, profile),
		},
	}

	payload := buildPDFStructuredPayload(artifacts, "identify the primary chart-focused evidence and what still needs review", false, nil, nil)
	if len(payload.StructuredBatches) < 2 {
		t.Fatalf("expected synthesized chart batch alongside OCR fallback, got %#v", payload.StructuredBatches)
	}
	if got := pdfStructuredResultBatchTargets(payload.StructuredBatches); len(got) < 2 || got[0] != "chart" || got[1] != "ocr_text" {
		t.Fatalf("expected chart+ocr_text planned targets, got %#v", got)
	}
	if !payload.StructuredBatches[0].FocusAligned || payload.StructuredBatches[0].FocusTarget != "chart" {
		t.Fatalf("expected synthesized chart batch to carry chart focus, got %#v", payload.StructuredBatches[0])
	}
	if payload.StructuredBatches[1].FocusTarget != "" {
		t.Fatalf("expected non-focus OCR batch to omit batch-level focus_target, got %#v", payload.StructuredBatches[1])
	}

	pruned := applyPDFStructuredDocumentSetContext(payload, &pdfAnalysisPlan{Mode: "hybrid_vision_text"}, nil)
	if len(pruned.StructuredBatches) != 1 || pruned.StructuredBatches[0].Target != "chart" {
		t.Fatalf("expected document-set context to retain only chart batch, got %#v", pruned.StructuredBatches)
	}
	if pruned.ResultEvaluation == nil || pruned.ResultEvaluation.Planner == nil {
		t.Fatalf("expected refreshed planner evaluation, got %#v", pruned.ResultEvaluation)
	}
	if len(pruned.ResultEvaluation.Planner.PrunedBatchTargets) == 0 || pruned.ResultEvaluation.Planner.PrunedBatchTargets[0] != "ocr_text" {
		t.Fatalf("expected OCR fallback to appear in pruned_batch_targets, got %#v", pruned.ResultEvaluation.Planner)
	}
}

func TestApplyPDFStructuredDocumentSetContext_RefreshesResultEvaluationAfterPrune(t *testing.T) {
	doc := pdfStructuredPayload{
		Path: "alpha.pdf",
		StructuredBatches: []pdfStructuredBatchResult{
			{Target: "chart", Status: "complete", CompletionRatio: 1, FocusAligned: true},
			{Target: "layout", Status: "partial", CompletionRatio: 0.5, ReviewRequired: true, MissingRequiredFields: []string{"layout_type"}, LowConfidenceFields: []string{"layout_type"}},
		},
		ResultEvaluation: buildPDFStructuredResultEvaluation(
			[]pdfVisualExtractionBatch{
				{Target: "chart"},
				{Target: "layout"},
			},
			[]pdfStructuredBatchResult{
				{Target: "chart", Status: "complete", CompletionRatio: 1, FocusAligned: true},
				{Target: "layout", Status: "partial", CompletionRatio: 0.5, ReviewRequired: true, MissingRequiredFields: []string{"layout_type"}, LowConfidenceFields: []string{"layout_type"}},
			},
			&pdfVisualAnalysis{Status: "success", SignalProfile: pdfVisualSignalProfile{PrimaryVisualTarget: "chart"}},
			"chart",
		),
	}

	got := applyPDFStructuredDocumentSetContext(
		doc,
		&pdfAnalysisPlan{Mode: "hybrid_vision_text"},
		&pdfVisualAnalysis{Status: "success", SignalProfile: pdfVisualSignalProfile{PrimaryVisualTarget: "chart"}},
	)
	if got.ResultEvaluation == nil || got.ResultEvaluation.Planner == nil {
		t.Fatalf("expected refreshed planner evaluation, got %#v", got.ResultEvaluation)
	}
	if got.ResultEvaluation.Planner.PlannedBatchCount != 2 {
		t.Fatalf("expected original planned_batch_count retained, got %#v", got.ResultEvaluation.Planner)
	}
	if len(got.ResultEvaluation.Planner.RetainedBatchTargets) != 1 || got.ResultEvaluation.Planner.RetainedBatchTargets[0] != "chart" {
		t.Fatalf("expected only chart retained after focus prune, got %#v", got.ResultEvaluation.Planner)
	}
	if len(got.ResultEvaluation.Planner.PrunedBatchTargets) != 1 || got.ResultEvaluation.Planner.PrunedBatchTargets[0] != "layout" {
		t.Fatalf("expected layout to appear as pruned planner target, got %#v", got.ResultEvaluation.Planner)
	}
}

func TestBuildPDFStructuredMultiPayload_AggregatesResultEvaluation(t *testing.T) {
	documents := []pdfStructuredPayload{
		{
			Path:           "alpha.pdf",
			Status:         "complete",
			ReviewRequired: false,
			StructuredBatches: []pdfStructuredBatchResult{
				{Target: "chart", Status: "complete", CompletionRatio: 1, FocusAligned: true},
			},
		},
		{
			Path:           "beta.pdf",
			Status:         "partial",
			ReviewRequired: true,
			ReviewSummary:  &pdfStructuredReviewSummary{BatchesRequiringReview: 1, LowConfidenceFields: 1, BatchTargets: []string{"beta.pdf:layout"}},
			StructuredBatches: []pdfStructuredBatchResult{
				{Target: "layout", Status: "partial", CompletionRatio: 0.5, ReviewRequired: true, MissingRequiredFields: []string{"layout_type"}, LowConfidenceFields: []string{"layout_type"}},
			},
		},
	}

	payload := buildPDFStructuredMultiPayload(
		documents,
		"revenue trend",
		&pdfAnalysisPlan{Mode: "hybrid_vision_text"},
		&pdfVisualAnalysis{Status: "success", SignalProfile: pdfVisualSignalProfile{PrimaryVisualTarget: "chart"}},
	)
	if payload.ResultEvaluation == nil || payload.ResultEvaluation.Planner == nil || payload.ResultEvaluation.Executor == nil {
		t.Fatalf("expected aggregated result evaluation, got %#v", payload.ResultEvaluation)
	}
	if payload.ResultEvaluation.Planner.FocusTarget != "chart" {
		t.Fatalf("expected planner focus_target=chart, got %#v", payload.ResultEvaluation.Planner)
	}
	if payload.ResultEvaluation.Executor.DocumentCount != 2 || payload.ResultEvaluation.Executor.ReviewDocumentCount != 1 {
		t.Fatalf("expected aggregated document counts, got %#v", payload.ResultEvaluation.Executor)
	}
	if payload.ResultEvaluation.Executor.Status != "review_required" {
		t.Fatalf("expected aggregated executor status=review_required, got %#v", payload.ResultEvaluation.Executor)
	}
	if len(payload.ResultEvaluation.Executor.TopDocumentPaths) == 0 || payload.ResultEvaluation.Executor.TopDocumentPaths[0] != "alpha.pdf" {
		t.Fatalf("expected top document path ranking to surface in evaluation, got %#v", payload.ResultEvaluation.Executor)
	}
	if len(payload.ResultEvaluation.Executor.TopFollowUpDocumentPaths) == 0 || payload.ResultEvaluation.Executor.TopFollowUpDocumentPaths[0] != "beta.pdf" {
		t.Fatalf("expected follow-up document path ranking to surface review-target doc, got %#v", payload.ResultEvaluation.Executor)
	}
	foundReviewRequired := false
	for _, code := range payload.ResultEvaluation.Executor.TopFollowUpReasonCodes {
		if code == "review_required" {
			foundReviewRequired = true
			break
		}
	}
	if !foundReviewRequired {
		t.Fatalf("expected follow-up reason codes to surface review_required, got %#v", payload.ResultEvaluation.Executor.TopFollowUpReasonCodes)
	}
	if len(payload.ResultEvaluation.Executor.TopFollowUpNotes) == 0 {
		t.Fatalf("expected follow-up notes to surface top-level rationale, got %#v", payload.ResultEvaluation.Executor)
	}
}

func TestBuildPDFStructuredMultiPayload_FallsBackToSecondaryDocumentForFollowUp(t *testing.T) {
	documents := []pdfStructuredPayload{
		{
			Path:           "alpha.pdf",
			Status:         "complete",
			ReviewRequired: false,
			MediaProfile:   pdfMediaProfile{LikelyGraphicDoc: true},
			VisualAnalysis: &pdfVisualAnalysis{
				SignalProfile: pdfVisualSignalProfile{PrimaryVisualTarget: "chart"},
			},
			StructuredBatches: []pdfStructuredBatchResult{
				{Target: "chart", Status: "complete", CompletionRatio: 0.9, FocusAligned: true},
			},
		},
		{
			Path:           "beta.pdf",
			Status:         "complete",
			ReviewRequired: false,
			MediaProfile:   pdfMediaProfile{LikelyGraphicDoc: true},
			VisualAnalysis: &pdfVisualAnalysis{
				SignalProfile: pdfVisualSignalProfile{PrimaryVisualTarget: "diagram"},
			},
			StructuredBatches: []pdfStructuredBatchResult{
				{Target: "diagram", Status: "complete", CompletionRatio: 0.7},
			},
		},
	}

	payload := buildPDFStructuredMultiPayload(
		documents,
		"revenue trend",
		&pdfAnalysisPlan{Mode: "hybrid_vision_text"},
		&pdfVisualAnalysis{Status: "success", SignalProfile: pdfVisualSignalProfile{PrimaryVisualTarget: "chart"}},
	)
	if payload.ResultEvaluation == nil || payload.ResultEvaluation.Executor == nil {
		t.Fatalf("expected aggregated executor evaluation, got %#v", payload.ResultEvaluation)
	}
	if len(payload.ResultEvaluation.Executor.TopDocumentPaths) == 0 || payload.ResultEvaluation.Executor.TopDocumentPaths[0] != "alpha.pdf" {
		t.Fatalf("expected alpha.pdf to remain top document, got %#v", payload.ResultEvaluation.Executor)
	}
	if len(payload.ResultEvaluation.Executor.TopFollowUpDocumentPaths) == 0 || payload.ResultEvaluation.Executor.TopFollowUpDocumentPaths[0] != "beta.pdf" {
		t.Fatalf("expected secondary ranked doc to become follow-up path when no review docs exist, got %#v", payload.ResultEvaluation.Executor)
	}
	foundCompleteStatus := false
	for _, code := range payload.ResultEvaluation.Executor.TopFollowUpReasonCodes {
		if code == "status_complete" {
			foundCompleteStatus = true
			break
		}
	}
	if !foundCompleteStatus {
		t.Fatalf("expected fallback follow-up rationale to surface secondary doc status, got %#v", payload.ResultEvaluation.Executor.TopFollowUpReasonCodes)
	}
	if len(payload.ResultEvaluation.Executor.TopFollowUpNotes) == 0 {
		t.Fatalf("expected fallback follow-up notes to surface secondary doc rationale, got %#v", payload.ResultEvaluation.Executor)
	}
}

func TestBuildPDFStructuredPayload_NoReviewSummaryWhenNotNeeded(t *testing.T) {
	query := "quarterly review notes"
	runtime := newPDFBackendRuntime(PDFToolOptions{
		Backend: stubPDFBackend{
			allPages: []PDFPageText{
				{Page: 1, Text: "Quarterly review\nRevenue is up quarter over quarter."},
			},
			pageCount: 1,
			available: true,
		},
	})
	artifacts, err := buildPDFAnalysisArtifacts(context.Background(), runtime, "sample.pdf", "sample.pdf", query, 120, "", false, "", 0)
	if err != nil {
		t.Fatalf("build artifacts: %v", err)
	}

	payload := buildPDFStructuredPayload(artifacts, query, false, nil, nil)
	if payload.ReviewRequired {
		t.Fatalf("expected no payload review_required, got %#v", payload.ReviewSummary)
	}
	if payload.ReviewSummary != nil {
		t.Fatalf("expected nil review_summary, got %#v", payload.ReviewSummary)
	}
}

func TestBuildPDFStructuredTopDocuments_PrefersDocumentSetAlignedDoc(t *testing.T) {
	documents := []pdfStructuredPayload{
		{
			Path:           "alpha.pdf",
			Status:         "complete",
			ReviewRequired: false,
			MediaProfile:   pdfMediaProfile{LikelyGraphicDoc: true},
			VisualAnalysis: &pdfVisualAnalysis{
				SignalProfile: pdfVisualSignalProfile{PrimaryVisualTarget: "chart"},
			},
			StructuredBatches: []pdfStructuredBatchResult{
				{Target: "chart", CompletionRatio: 0.55},
			},
			ReviewSummary: &pdfStructuredReviewSummary{
				TopNotes: []string{"Alpha note"},
			},
		},
		{
			Path:           "beta.pdf",
			Status:         "complete",
			ReviewRequired: false,
			MediaProfile:   pdfMediaProfile{LikelyGraphicDoc: true},
			VisualAnalysis: &pdfVisualAnalysis{
				SignalProfile: pdfVisualSignalProfile{PrimaryVisualTarget: "diagram"},
			},
			StructuredBatches: []pdfStructuredBatchResult{
				{Target: "diagram", CompletionRatio: 0.95},
			},
			ReviewSummary: &pdfStructuredReviewSummary{
				TopNotes: []string{"Beta note"},
			},
		},
	}
	analysisPlan := &pdfAnalysisPlan{Mode: "hybrid_vision_text"}
	visualAnalysis := &pdfVisualAnalysis{
		Summary: "Cross-document chart comparison shows alpha contains the primary revenue chart.",
		SignalProfile: pdfVisualSignalProfile{
			PrimaryVisualTarget: "chart",
			SummaryMode:         "chart_summary",
		},
	}

	topDocs := buildPDFStructuredTopDocuments(documents, analysisPlan, visualAnalysis)
	if len(topDocs) != 2 {
		t.Fatalf("expected 2 top documents, got %#v", topDocs)
	}
	if topDocs[0].Path != "alpha.pdf" {
		t.Fatalf("expected alpha.pdf to rank first from document-set chart focus, got %#v", topDocs)
	}
	if len(topDocs[0].TopNotes) == 0 || !strings.Contains(strings.ToLower(topDocs[0].TopNotes[0]), "document-set chart priority") {
		t.Fatalf("expected context-aware top note on alpha, got %#v", topDocs[0].TopNotes)
	}
	if len(topDocs[0].SelectionReasons) == 0 || !strings.Contains(strings.ToLower(topDocs[0].SelectionReasons[0]), "document-set chart priority") {
		t.Fatalf("expected context-aware selection reason on alpha, got %#v", topDocs[0].SelectionReasons)
	}
	if len(topDocs[0].SelectionReasonCodes) == 0 || topDocs[0].SelectionReasonCodes[0] != "focus_aligned" {
		t.Fatalf("expected focus-aligned selection reason code on alpha, got %#v", topDocs[0].SelectionReasonCodes)
	}
}

func TestBuildPDFStructuredMultiPayload_IncludesDocumentSetReviewContext(t *testing.T) {
	documents := []pdfStructuredPayload{
		{
			Path:           "alpha.pdf",
			Status:         "partial",
			ReviewRequired: true,
			StructuredBatches: []pdfStructuredBatchResult{
				{
					Target:              "chart",
					CompletionRatio:     0.5,
					ReviewRequired:      true,
					ReviewNotes:         []string{"alpha needs review"},
					LowConfidenceFields: []string{"main_trend", "chart_type"},
				},
			},
			ReviewSummary: &pdfStructuredReviewSummary{
				BatchesRequiringReview: 1,
				LowConfidenceFields:    2,
				BatchTargets:           []string{"chart"},
				TopNotes:               []string{"alpha needs review"},
			},
		},
	}
	analysisPlan := &pdfAnalysisPlan{Mode: "hybrid_vision_text"}
	visualAnalysis := &pdfVisualAnalysis{
		Summary: "Cross-document chart comparison highlights alpha as the key chart-heavy PDF.",
		SignalProfile: pdfVisualSignalProfile{
			PrimaryVisualTarget: "chart",
			SummaryMode:         "chart_summary",
		},
	}

	payload := buildPDFStructuredMultiPayload(documents, "revenue trend", analysisPlan, visualAnalysis)
	if payload.ReviewSummary == nil {
		t.Fatalf("expected review summary, got %#v", payload)
	}
	if len(payload.ReviewSummary.TopNotes) < 2 {
		t.Fatalf("expected context and document review notes, got %#v", payload.ReviewSummary.TopNotes)
	}
	if payload.ReviewSummary.FocusTarget != "chart" {
		t.Fatalf("expected chart focus target, got %#v", payload.ReviewSummary.FocusTarget)
	}
	if len(payload.ReviewSummary.ReviewReasonCodes) == 0 || payload.ReviewSummary.ReviewReasonCodes[0] != "focus_target_chart" {
		t.Fatalf("expected chart focus review code first, got %#v", payload.ReviewSummary.ReviewReasonCodes)
	}
	if !strings.Contains(strings.ToLower(payload.ReviewSummary.TopNotes[0]), "cross-document focus") {
		t.Fatalf("expected cross-document focus note first, got %#v", payload.ReviewSummary.TopNotes)
	}
	foundSummary := false
	for _, note := range payload.ReviewSummary.TopNotes {
		if strings.Contains(strings.ToLower(note), "cross-document visual summary") {
			foundSummary = true
			break
		}
	}
	if !foundSummary {
		t.Fatalf("expected cross-document visual summary note, got %#v", payload.ReviewSummary.TopNotes)
	}
	if len(payload.ReviewSummary.ReviewDrivers) == 0 || !strings.Contains(strings.ToLower(payload.ReviewSummary.ReviewDrivers[0]), "cross-document focus") {
		t.Fatalf("expected focus driver first, got %#v", payload.ReviewSummary.ReviewDrivers)
	}
	foundDriverPath := false
	for _, driver := range payload.ReviewSummary.ReviewDrivers {
		if strings.Contains(strings.ToLower(driver), "alpha.pdf") {
			foundDriverPath = true
			break
		}
	}
	if !foundDriverPath {
		t.Fatalf("expected review drivers annotated with document path, got %#v", payload.ReviewSummary.ReviewDrivers)
	}
}

func TestBuildPDFStructuredMultiPayload_FallsBackToQueryFocusTarget(t *testing.T) {
	documents := []pdfStructuredPayload{
		{
			Path:           "alpha.pdf",
			Status:         "partial",
			ReviewRequired: true,
			StructuredBatches: []pdfStructuredBatchResult{
				{
					Target:          "chart",
					FocusAligned:    true,
					FocusTarget:     "chart",
					Status:          "partial",
					CompletionRatio: 0.5,
					ReviewRequired:  true,
					ReviewNotes:     []string{"main trend needs visual confirmation"},
					LowConfidenceFields: []string{
						"main_trend",
					},
					Sections: []pdfStructuredSectionResult{
						{
							Name: "summary",
							Fields: []pdfStructuredFieldResult{
								{
									Name:                  "main_trend",
									Required:              true,
									Filled:                false,
									NeedsReview:           true,
									CrossDocumentPriority: true,
									Evidence:              []string{"page 1"},
									IssueFlags:            []string{"cross_document_priority_missing"},
								},
							},
						},
					},
				},
				{
					Target:          "layout",
					Status:          "partial",
					CompletionRatio: 0.5,
				},
			},
			ReviewSummary: &pdfStructuredReviewSummary{
				BatchesRequiringReview: 1,
				LowConfidenceFields:    1,
				BatchTargets:           []string{"chart"},
				FocusTarget:            "chart",
				ReviewReasonCodes:      []string{"focus_target_chart"},
				TopNotes:               []string{"alpha chart needs review"},
			},
			ResultEvaluation: &pdfStructuredResultEvaluation{
				Planner: &pdfStructuredPlannerEvaluation{
					FocusTarget:          "chart",
					PlannedBatchCount:    2,
					PlannedBatchTargets:  []string{"chart", "layout"},
					RetainedBatchCount:   1,
					RetainedBatchTargets: []string{"chart"},
					PrunedBatchTargets:   []string{"layout"},
				},
			},
		},
	}

	payload := buildPDFStructuredMultiPayload(documents, "identify the primary chart-focused evidence and follow-up review", nil, nil)
	if payload.ReviewSummary == nil || payload.ReviewSummary.FocusTarget != "chart" {
		t.Fatalf("expected query fallback focus_target=chart in review summary, got %#v", payload.ReviewSummary)
	}
	if payload.ResultEvaluation == nil || payload.ResultEvaluation.Planner == nil || payload.ResultEvaluation.Planner.FocusTarget != "chart" {
		t.Fatalf("expected planner focus_target=chart, got %#v", payload.ResultEvaluation)
	}
	if len(payload.ResultEvaluation.Planner.PrunedBatchTargets) == 0 || payload.ResultEvaluation.Planner.PrunedBatchTargets[0] != "layout" {
		t.Fatalf("expected pruned layout batch to survive evaluation refresh, got %#v", payload.ResultEvaluation.Planner)
	}
	if payload.ResultEvaluation.Executor == nil || payload.ResultEvaluation.Executor.TopDocumentPaths[0] != "alpha.pdf" {
		t.Fatalf("expected executor top document to remain alpha.pdf, got %#v", payload.ResultEvaluation)
	}
}

func TestApplyPDFStructuredDocumentSetContext_PrioritizesFocusAndPrunesLayout(t *testing.T) {
	doc := pdfStructuredPayload{
		Path: "alpha.pdf",
		StructuredBatches: []pdfStructuredBatchResult{
			{
				Target:          "layout",
				Priority:        "high",
				Status:          "partial",
				CompletionRatio: 0.30,
			},
			{
				Target:          "chart",
				Priority:        "low",
				Status:          "complete",
				CompletionRatio: 0.90,
			},
		},
	}
	analysisPlan := &pdfAnalysisPlan{Mode: "hybrid_vision_text"}
	visualAnalysis := &pdfVisualAnalysis{
		SignalProfile: pdfVisualSignalProfile{
			PrimaryVisualTarget: "chart",
			SummaryMode:         "chart_summary",
		},
	}

	got := applyPDFStructuredDocumentSetContext(doc, analysisPlan, visualAnalysis)
	if len(got.StructuredBatches) != 1 {
		t.Fatalf("expected layout batch pruned, got %#v", got.StructuredBatches)
	}
	if got.StructuredBatches[0].Target != "chart" {
		t.Fatalf("expected chart batch to remain first, got %#v", got.StructuredBatches)
	}
	if got.StructuredBatches[0].Priority != "high" {
		t.Fatalf("expected focus batch promoted to high priority, got %#v", got.StructuredBatches[0].Priority)
	}
	if len(got.StructuredBatches[0].Reasons) == 0 || !strings.Contains(strings.ToLower(got.StructuredBatches[0].Reasons[0]), "document-set chart focus") {
		t.Fatalf("expected document-set focus reason, got %#v", got.StructuredBatches[0].Reasons)
	}
	if got.Status != "complete" {
		t.Fatalf("expected status rebuilt from remaining batch, got %#v", got.Status)
	}
}

func TestApplyPDFStructuredDocumentSetContext_PrunesOCRForChartFocus(t *testing.T) {
	doc := pdfStructuredPayload{
		Path: "alpha.pdf",
		StructuredBatches: []pdfStructuredBatchResult{
			{
				Target:          "chart",
				Priority:        "high",
				FocusAligned:    true,
				FocusTarget:     "chart",
				Status:          "partial",
				CompletionRatio: 0.7,
			},
			{
				Target:          "ocr_text",
				Priority:        "high",
				Status:          "complete",
				CompletionRatio: 1.0,
			},
		},
		ResultEvaluation: buildPDFStructuredResultEvaluation(
			[]pdfVisualExtractionBatch{
				{Target: "chart"},
				{Target: "ocr_text"},
			},
			[]pdfStructuredBatchResult{
				{Target: "chart", FocusAligned: true, FocusTarget: "chart", Status: "partial", CompletionRatio: 0.7},
				{Target: "ocr_text", Status: "complete", CompletionRatio: 1.0},
			},
			&pdfVisualAnalysis{Status: "success", SignalProfile: pdfVisualSignalProfile{PrimaryVisualTarget: "chart"}},
			"chart",
		),
	}

	got := applyPDFStructuredDocumentSetContext(doc, &pdfAnalysisPlan{Mode: "hybrid_vision_text"}, nil)
	if len(got.StructuredBatches) != 1 || got.StructuredBatches[0].Target != "chart" {
		t.Fatalf("expected OCR fallback pruned for chart focus, got %#v", got.StructuredBatches)
	}
	if got.ResultEvaluation == nil || got.ResultEvaluation.Planner == nil || len(got.ResultEvaluation.Planner.PrunedBatchTargets) == 0 || got.ResultEvaluation.Planner.PrunedBatchTargets[0] != "ocr_text" {
		t.Fatalf("expected pruned planner targets to include ocr_text, got %#v", got.ResultEvaluation)
	}
}

func TestApplyPDFStructuredDocumentSetContext_RebuildsReviewSummaryAndWarnings(t *testing.T) {
	doc := pdfStructuredPayload{
		Path: "alpha.pdf",
		StructuredBatches: []pdfStructuredBatchResult{
			{
				Target:                "layout",
				Priority:              "medium",
				Status:                "partial",
				CompletionRatio:       0.25,
				ReviewRequired:        true,
				ReviewNotes:           []string{"layout requires review"},
				LowConfidenceFields:   []string{"layout_type"},
				MissingRequiredFields: []string{"layout_type"},
			},
			{
				Target:                "chart",
				Priority:              "medium",
				Status:                "partial",
				CompletionRatio:       0.70,
				ReviewRequired:        true,
				ReviewNotes:           []string{"chart requires review"},
				LowConfidenceFields:   []string{"main_trend"},
				MissingRequiredFields: []string{"main_trend"},
			},
		},
	}
	analysisPlan := &pdfAnalysisPlan{Mode: "hybrid_vision_text"}
	visualAnalysis := &pdfVisualAnalysis{
		Summary: "Cross-document chart comparison highlights alpha.",
		SignalProfile: pdfVisualSignalProfile{
			PrimaryVisualTarget: "chart",
			SummaryMode:         "chart_summary",
		},
	}

	got := applyPDFStructuredDocumentSetContext(doc, analysisPlan, visualAnalysis)
	if got.ReviewSummary == nil {
		t.Fatalf("expected review summary, got %#v", got)
	}
	if got.ReviewSummary.BatchesRequiringReview != 1 {
		t.Fatalf("expected only chart batch to remain in review summary, got %#v", got.ReviewSummary)
	}
	if len(got.ReviewSummary.BatchTargets) != 1 || got.ReviewSummary.BatchTargets[0] != "chart" {
		t.Fatalf("expected chart-only batch target, got %#v", got.ReviewSummary.BatchTargets)
	}
	if len(got.ReviewSummary.TopNotes) == 0 || !strings.Contains(strings.ToLower(got.ReviewSummary.TopNotes[0]), "document-set chart priority") {
		t.Fatalf("expected document-set context note first, got %#v", got.ReviewSummary.TopNotes)
	}
	if got.ReviewSummary.FocusTarget != "chart" {
		t.Fatalf("expected chart focus target, got %#v", got.ReviewSummary.FocusTarget)
	}
	if len(got.ReviewSummary.ReviewReasonCodes) == 0 || got.ReviewSummary.ReviewReasonCodes[0] != "focus_target_chart" {
		t.Fatalf("expected explainable review reason codes, got %#v", got.ReviewSummary.ReviewReasonCodes)
	}
	if len(got.ReviewSummary.ReviewDrivers) == 0 || !strings.Contains(strings.ToLower(got.ReviewSummary.ReviewDrivers[0]), "cross-document focus") {
		t.Fatalf("expected explainable review drivers, got %#v", got.ReviewSummary.ReviewDrivers)
	}
	if !strings.Contains(strings.ToLower(got.Warning), "chart batch missing required fields") {
		t.Fatalf("expected warning rebuilt from remaining chart batch, got %#v", got.Warning)
	}
	if strings.Contains(strings.ToLower(got.Warning), "layout batch") {
		t.Fatalf("did not expect pruned layout warning to remain, got %#v", got.Warning)
	}
}

func TestApplyPDFStructuredFieldLevelContext_MarksPriorityFields(t *testing.T) {
	batch := pdfStructuredBatchResult{
		Target: "chart",
		Sections: []pdfStructuredSectionResult{
			{
				Name:          "chart_takeaway",
				QualityChecks: []string{"base quality check"},
				Fields: []pdfStructuredFieldResult{
					{Name: "chart_type", Filled: true},
					{Name: "main_trend", Filled: true},
					{Name: "supporting_annotation", Filled: true},
				},
			},
		},
		ValidationChecks: []string{"base validation"},
		AggregationRules: []string{"base aggregation"},
	}

	got := applyPDFStructuredFieldLevelContext(batch, "chart")
	if !got.FocusAligned && got.FocusTarget != "" {
		t.Fatalf("unexpected focus state before outer context, got %#v", got)
	}
	if len(got.ValidationChecks) == 0 || !strings.Contains(strings.ToLower(got.ValidationChecks[0]), "document_set_focus_chart") {
		t.Fatalf("expected chart focus validation check, got %#v", got.ValidationChecks)
	}
	if len(got.AggregationRules) == 0 || !strings.Contains(strings.ToLower(got.AggregationRules[0]), "cross-document comparison") {
		t.Fatalf("expected chart focus aggregation rule, got %#v", got.AggregationRules)
	}
	fields := got.Sections[0].Fields
	if !fields[0].CrossDocumentPriority || !fields[1].CrossDocumentPriority {
		t.Fatalf("expected chart_type/main_trend marked as cross-document priority, got %#v", fields)
	}
	if fields[2].CrossDocumentPriority {
		t.Fatalf("did not expect supporting_annotation to be priority, got %#v", fields[2])
	}
	if len(fields[0].Notes) == 0 || !strings.Contains(strings.ToLower(fields[0].Notes[0]), "cross-document priority field") {
		t.Fatalf("expected priority note on chart_type, got %#v", fields[0].Notes)
	}
	if len(got.Sections[0].QualityChecks) == 0 || !strings.Contains(strings.ToLower(got.Sections[0].QualityChecks[0]), "cross-document priority") {
		t.Fatalf("expected section quality check to mention cross-document priority, got %#v", got.Sections[0].QualityChecks)
	}
}

func TestApplyPDFStructuredFieldLevelContext_AddsPriorityMissingIssueFlag(t *testing.T) {
	batch := pdfStructuredBatchResult{
		Target: "chart",
		Sections: []pdfStructuredSectionResult{
			{
				Name: "chart_takeaway",
				Fields: []pdfStructuredFieldResult{
					{Name: "main_trend", Filled: false, NeedsReview: false, IssueFlags: nil},
				},
			},
		},
	}

	got := applyPDFStructuredFieldLevelContext(batch, "chart")
	field := got.Sections[0].Fields[0]
	if !field.CrossDocumentPriority {
		t.Fatalf("expected main_trend to be cross-document priority, got %#v", field)
	}
	if !field.NeedsReview {
		t.Fatalf("expected missing priority field to require review, got %#v", field)
	}
	found := false
	for _, flag := range field.IssueFlags {
		if flag == "cross_document_priority_missing" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected cross_document_priority_missing in issue flags, got %#v", field.IssueFlags)
	}
	if len(got.ReviewNotes) == 0 || !strings.Contains(strings.ToLower(got.ReviewNotes[0]), "cross_document_priority_missing") {
		t.Fatalf("expected batch review note for missing priority field, got %#v", got.ReviewNotes)
	}
	if len(got.ReviewDrivers) == 0 || !strings.Contains(strings.ToLower(got.ReviewDrivers[0]), "primary chart fields") {
		t.Fatalf("expected batch review drivers, got %#v", got.ReviewDrivers)
	}
	if len(got.ReviewReasonCodes) == 0 || got.ReviewReasonCodes[0] != "cross_document_priority_missing" {
		t.Fatalf("expected batch review reason codes, got %#v", got.ReviewReasonCodes)
	}
	if !got.ReviewRequired {
		t.Fatalf("expected batch review required, got %#v", got)
	}
	if len(got.LowConfidenceFields) == 0 || got.LowConfidenceFields[0] != "main_trend" {
		t.Fatalf("expected main_trend in low confidence fields, got %#v", got.LowConfidenceFields)
	}
}

func TestInferPDFStructuredFieldCandidate_UsesSummaryFirstForChartFocus(t *testing.T) {
	candidate := inferPDFStructuredFieldCandidate(
		pdfVisualExtractionField{Name: "main_trend", Kind: "string", Required: true},
		pdfVisualExtractionBatch{Target: "chart", Pages: []int{1}},
		pdfVisualSignalProfile{},
		"Revenue rises sharply after Q2 and remains above the prior baseline.",
		[]string{"Revenue overview\nQ2 revenue 120, Q3 revenue 180"},
		[]string{"Page 1: Revenue overview\nQ2 revenue 120, Q3 revenue 180"},
		pdfStructuredFieldInferenceContext{
			FocusAligned:  true,
			FocusTarget:   "chart",
			PriorityField: true,
		},
	)
	if candidate.Source != "cross_document_visual_summary" {
		t.Fatalf("expected summary-first source, got %#v", candidate.Source)
	}
	if candidate.Confidence != "cross_document_priority_summary_first" {
		t.Fatalf("expected summary-first confidence, got %#v", candidate.Confidence)
	}
	if len(candidate.Evidence) == 0 || !strings.Contains(strings.ToLower(candidate.Evidence[0]), "page 1") || !strings.Contains(strings.ToLower(candidate.Evidence[0]), "revenue rises sharply") {
		t.Fatalf("expected summary evidence first, got %#v", candidate.Evidence)
	}
	if len(candidate.Notes) == 0 || !strings.Contains(strings.ToLower(candidate.Notes[0]), "summary-first evidence") {
		t.Fatalf("expected summary-first routing note, got %#v", candidate.Notes)
	}
}

func TestInferPDFStructuredFieldCandidate_UsesExcerptFirstForTableFocus(t *testing.T) {
	candidate := inferPDFStructuredFieldCandidate(
		pdfVisualExtractionField{Name: "key_headers", Kind: "string_list", Required: true},
		pdfVisualExtractionBatch{Target: "table", Pages: []int{2}},
		pdfVisualSignalProfile{},
		"Summary mentions totals and a comparison table.",
		[]string{"Region | Revenue | Margin\nUS | 100 | 25%"},
		[]string{"Page 2: Region | Revenue | Margin\nUS | 100 | 25%"},
		pdfStructuredFieldInferenceContext{
			FocusAligned:  true,
			FocusTarget:   "table",
			PriorityField: true,
		},
	)
	if candidate.Source != "cross_document_page_excerpt" {
		t.Fatalf("expected excerpt-first source, got %#v", candidate.Source)
	}
	if candidate.Confidence != "cross_document_priority_excerpt_first" {
		t.Fatalf("expected excerpt-first confidence, got %#v", candidate.Confidence)
	}
	if len(candidate.Evidence) == 0 || !strings.Contains(strings.ToLower(candidate.Evidence[0]), "page 2") || !strings.Contains(strings.ToLower(candidate.Evidence[0]), "region | revenue | margin") {
		t.Fatalf("expected excerpt evidence first, got %#v", candidate.Evidence)
	}
	if len(candidate.Notes) == 0 || !strings.Contains(strings.ToLower(candidate.Notes[0]), "excerpt-first evidence") {
		t.Fatalf("expected excerpt-first routing note, got %#v", candidate.Notes)
	}
}
