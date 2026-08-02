package tools

import "strings"

type pdfStructuredResultEvaluation struct {
	Planner  *pdfStructuredPlannerEvaluation  `json:"planner,omitempty"`
	Executor *pdfStructuredExecutorEvaluation `json:"executor,omitempty"`
}

type pdfStructuredPlannerEvaluation struct {
	FocusTarget          string   `json:"focus_target,omitempty"`
	VisualStatus         string   `json:"visual_status,omitempty"`
	NativeVisual         bool     `json:"native_visual,omitempty"`
	PlannedBatchCount    int      `json:"planned_batch_count,omitempty"`
	PlannedBatchTargets  []string `json:"planned_batch_targets,omitempty"`
	RetainedBatchCount   int      `json:"retained_batch_count,omitempty"`
	RetainedBatchTargets []string `json:"retained_batch_targets,omitempty"`
	PrunedBatchTargets   []string `json:"pruned_batch_targets,omitempty"`
}

type pdfStructuredExecutorEvaluation struct {
	Status                    string   `json:"status,omitempty"`
	DocumentCount             int      `json:"document_count,omitempty"`
	CompleteDocumentCount     int      `json:"complete_document_count,omitempty"`
	PartialDocumentCount      int      `json:"partial_document_count,omitempty"`
	EmptyDocumentCount        int      `json:"empty_document_count,omitempty"`
	ReviewDocumentCount       int      `json:"review_document_count,omitempty"`
	BatchCount                int      `json:"batch_count,omitempty"`
	CompleteBatchCount        int      `json:"complete_batch_count,omitempty"`
	PartialBatchCount         int      `json:"partial_batch_count,omitempty"`
	EmptyBatchCount           int      `json:"empty_batch_count,omitempty"`
	ReviewBatchCount          int      `json:"review_batch_count,omitempty"`
	FocusAlignedBatchCount    int      `json:"focus_aligned_batch_count,omitempty"`
	MissingRequiredFieldCount int      `json:"missing_required_field_count,omitempty"`
	LowConfidenceFieldCount   int      `json:"low_confidence_field_count,omitempty"`
	AverageCompletionRatio    float64  `json:"average_completion_ratio,omitempty"`
	TopDocumentPaths          []string `json:"top_document_paths,omitempty"`
	TopFollowUpDocumentPaths  []string `json:"top_follow_up_document_paths,omitempty"`
	TopFollowUpReasonCodes    []string `json:"top_follow_up_reason_codes,omitempty"`
	TopFollowUpNotes          []string `json:"top_follow_up_notes,omitempty"`
}

func buildPDFStructuredResultEvaluation(
	planned []pdfVisualExtractionBatch,
	retained []pdfStructuredBatchResult,
	visual *pdfVisualAnalysis,
	focusTarget string,
) *pdfStructuredResultEvaluation {
	return buildPDFStructuredResultEvaluationFromPlannerSnapshot(
		len(planned),
		pdfStructuredPlannedBatchTargets(planned),
		retained,
		visual,
		focusTarget,
	)
}

func buildPDFStructuredResultEvaluationFromPlannerSnapshot(
	plannedBatchCount int,
	plannedBatchTargets []string,
	retained []pdfStructuredBatchResult,
	visual *pdfVisualAnalysis,
	focusTarget string,
) *pdfStructuredResultEvaluation {
	planner := buildPDFStructuredPlannerEvaluation(plannedBatchCount, plannedBatchTargets, retained, visual, focusTarget)
	executor := buildPDFStructuredExecutorEvaluation(retained)
	if planner == nil && executor == nil {
		return nil
	}
	return &pdfStructuredResultEvaluation{
		Planner:  planner,
		Executor: executor,
	}
}

func refreshPDFStructuredResultEvaluation(
	doc pdfStructuredPayload,
	analysisPlan *pdfAnalysisPlan,
	visualAnalysis *pdfVisualAnalysis,
) pdfStructuredPayload {
	focusTarget := firstNonEmpty(
		pdfStructuredDocumentSetFocusTarget(analysisPlan, visualAnalysis),
		func() string {
			if doc.ResultEvaluation != nil && doc.ResultEvaluation.Planner != nil {
				return strings.TrimSpace(doc.ResultEvaluation.Planner.FocusTarget)
			}
			return ""
		}(),
		pdfStructuredPayloadFocusTarget(doc),
	)
	plannedBatchCount := len(doc.StructuredBatches)
	var plannedBatchTargets []string
	if doc.ResultEvaluation != nil && doc.ResultEvaluation.Planner != nil {
		plannedBatchCount = doc.ResultEvaluation.Planner.PlannedBatchCount
		plannedBatchTargets = append([]string(nil), doc.ResultEvaluation.Planner.PlannedBatchTargets...)
	}
	if len(plannedBatchTargets) == 0 {
		plannedBatchTargets = pdfStructuredResultBatchTargets(doc.StructuredBatches)
	}
	doc.ResultEvaluation = buildPDFStructuredResultEvaluationFromPlannerSnapshot(
		plannedBatchCount,
		plannedBatchTargets,
		doc.StructuredBatches,
		visualAnalysis,
		focusTarget,
	)
	return doc
}

func buildPDFStructuredMultiResultEvaluation(
	documents []pdfStructuredPayload,
	analysisPlan *pdfAnalysisPlan,
	visualAnalysis *pdfVisualAnalysis,
	topDocuments []pdfStructuredTopDocument,
) *pdfStructuredResultEvaluation {
	if len(documents) == 0 {
		return nil
	}
	focusTarget := firstNonEmpty(
		pdfStructuredDocumentSetFocusTarget(analysisPlan, visualAnalysis),
		pdfStructuredPayloadsFocusTarget(documents),
	)
	plannedBatchCount := 0
	plannedBatchTargets := make([]string, 0, len(documents))
	retained := make([]pdfStructuredBatchResult, 0)
	completeDocuments := 0
	partialDocuments := 0
	emptyDocuments := 0
	reviewDocuments := 0
	for _, doc := range documents {
		retained = append(retained, doc.StructuredBatches...)
		if doc.ReviewRequired {
			reviewDocuments++
		}
		switch strings.ToLower(strings.TrimSpace(doc.Status)) {
		case "complete":
			completeDocuments++
		case "partial":
			partialDocuments++
		default:
			emptyDocuments++
		}
		if doc.ResultEvaluation != nil && doc.ResultEvaluation.Planner != nil {
			plannedBatchCount += doc.ResultEvaluation.Planner.PlannedBatchCount
			plannedBatchTargets = append(plannedBatchTargets, doc.ResultEvaluation.Planner.PlannedBatchTargets...)
			continue
		}
		plannedBatchCount += len(doc.StructuredBatches)
		plannedBatchTargets = append(plannedBatchTargets, pdfStructuredResultBatchTargets(doc.StructuredBatches)...)
	}
	planner := buildPDFStructuredPlannerEvaluation(plannedBatchCount, plannedBatchTargets, retained, visualAnalysis, focusTarget)
	executor := buildPDFStructuredExecutorEvaluation(retained)
	if executor != nil {
		executor.DocumentCount = len(documents)
		executor.CompleteDocumentCount = completeDocuments
		executor.PartialDocumentCount = partialDocuments
		executor.EmptyDocumentCount = emptyDocuments
		executor.ReviewDocumentCount = reviewDocuments
		executor.TopDocumentPaths = pdfStructuredTopDocumentPaths(topDocuments, 3)
		executor.TopFollowUpDocumentPaths = pdfStructuredTopFollowUpDocumentPaths(documents, topDocuments, 3)
		executor.TopFollowUpReasonCodes = pdfStructuredTopFollowUpReasonCodes(documents, topDocuments, executor.TopFollowUpDocumentPaths, 8)
		executor.TopFollowUpNotes = pdfStructuredTopFollowUpNotes(documents, topDocuments, executor.TopFollowUpDocumentPaths, 6)
		executor.Status = pdfStructuredMultiExecutorStatus(completeDocuments, partialDocuments, emptyDocuments, reviewDocuments, executor)
	}
	if planner == nil && executor == nil {
		return nil
	}
	return &pdfStructuredResultEvaluation{
		Planner:  planner,
		Executor: executor,
	}
}

func buildPDFStructuredPlannerEvaluation(
	plannedBatchCount int,
	plannedBatchTargets []string,
	retained []pdfStructuredBatchResult,
	visual *pdfVisualAnalysis,
	focusTarget string,
) *pdfStructuredPlannerEvaluation {
	retainedTargets := pdfStructuredResultBatchTargets(retained)
	prunedTargets := pdfStructuredSubtractTargets(plannedBatchTargets, retainedTargets)
	visualStatus := ""
	nativeVisual := false
	if visual != nil {
		visualStatus = strings.TrimSpace(visual.Status)
		nativeVisual = visual.NativePDF
	}
	if plannedBatchCount == 0 && len(plannedBatchTargets) == 0 && len(retained) == 0 && visualStatus == "" && strings.TrimSpace(focusTarget) == "" {
		return nil
	}
	return &pdfStructuredPlannerEvaluation{
		FocusTarget:          strings.TrimSpace(focusTarget),
		VisualStatus:         visualStatus,
		NativeVisual:         nativeVisual,
		PlannedBatchCount:    plannedBatchCount,
		PlannedBatchTargets:  dedupePDFVisualStrings(plannedBatchTargets),
		RetainedBatchCount:   len(retained),
		RetainedBatchTargets: retainedTargets,
		PrunedBatchTargets:   prunedTargets,
	}
}

func buildPDFStructuredExecutorEvaluation(batches []pdfStructuredBatchResult) *pdfStructuredExecutorEvaluation {
	if len(batches) == 0 {
		return nil
	}
	eval := &pdfStructuredExecutorEvaluation{
		BatchCount:             len(batches),
		AverageCompletionRatio: averagePDFStructuredCompletionRatio(batches),
	}
	for _, batch := range batches {
		switch strings.ToLower(strings.TrimSpace(batch.Status)) {
		case "complete":
			eval.CompleteBatchCount++
		case "partial":
			eval.PartialBatchCount++
		default:
			eval.EmptyBatchCount++
		}
		if batch.ReviewRequired {
			eval.ReviewBatchCount++
		}
		if batch.FocusAligned {
			eval.FocusAlignedBatchCount++
		}
		eval.MissingRequiredFieldCount += len(batch.MissingRequiredFields)
		eval.LowConfidenceFieldCount += len(batch.LowConfidenceFields)
	}
	eval.Status = pdfStructuredExecutorStatus(eval.CompleteBatchCount, eval.PartialBatchCount, eval.EmptyBatchCount, eval.ReviewBatchCount)
	return eval
}

func pdfStructuredExecutorStatus(completeCount int, partialCount int, emptyCount int, reviewCount int) string {
	switch {
	case reviewCount > 0:
		return "review_required"
	case partialCount > 0:
		return "partial"
	case completeCount > 0 && emptyCount == 0:
		return "complete"
	case completeCount > 0:
		return "partial"
	case emptyCount > 0:
		return "empty"
	default:
		return ""
	}
}

func pdfStructuredMultiExecutorStatus(
	completeDocuments int,
	partialDocuments int,
	emptyDocuments int,
	reviewDocuments int,
	executor *pdfStructuredExecutorEvaluation,
) string {
	switch {
	case reviewDocuments > 0:
		return "review_required"
	case partialDocuments > 0:
		return "partial"
	case completeDocuments > 0 && emptyDocuments == 0:
		return "complete"
	case executor != nil:
		return pdfStructuredExecutorStatus(executor.CompleteBatchCount, executor.PartialBatchCount, executor.EmptyBatchCount, executor.ReviewBatchCount)
	default:
		return ""
	}
}

func pdfStructuredPlannedBatchTargets(planned []pdfVisualExtractionBatch) []string {
	targets := make([]string, 0, len(planned))
	for _, batch := range planned {
		target := strings.TrimSpace(batch.Target)
		if target == "" {
			continue
		}
		targets = append(targets, target)
	}
	return dedupePDFVisualStrings(targets)
}

func pdfStructuredResultBatchTargets(batches []pdfStructuredBatchResult) []string {
	targets := make([]string, 0, len(batches))
	for _, batch := range batches {
		target := strings.TrimSpace(batch.Target)
		if target == "" {
			continue
		}
		targets = append(targets, target)
	}
	return dedupePDFVisualStrings(targets)
}

func pdfStructuredSubtractTargets(left []string, right []string) []string {
	if len(left) == 0 {
		return nil
	}
	blocked := make(map[string]struct{}, len(right))
	for _, item := range right {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		blocked[item] = struct{}{}
	}
	out := make([]string, 0, len(left))
	for _, item := range left {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := blocked[item]; ok {
			continue
		}
		out = append(out, item)
	}
	return dedupePDFVisualStrings(out)
}

func pdfStructuredTopDocumentPaths(topDocuments []pdfStructuredTopDocument, limit int) []string {
	if len(topDocuments) == 0 {
		return nil
	}
	if limit <= 0 || limit > len(topDocuments) {
		limit = len(topDocuments)
	}
	out := make([]string, 0, limit)
	for _, item := range topDocuments[:limit] {
		path := strings.TrimSpace(item.Path)
		if path == "" {
			continue
		}
		out = append(out, path)
	}
	return dedupePDFVisualStrings(out)
}

func pdfStructuredTopFollowUpDocumentPaths(
	documents []pdfStructuredPayload,
	topDocuments []pdfStructuredTopDocument,
	limit int,
) []string {
	if len(documents) == 0 || len(topDocuments) == 0 {
		return nil
	}
	if limit <= 0 || limit > len(topDocuments) {
		limit = len(topDocuments)
	}
	docsByPath := make(map[string]pdfStructuredPayload, len(documents))
	for _, doc := range documents {
		path := strings.TrimSpace(doc.Path)
		if path == "" {
			continue
		}
		docsByPath[path] = doc
	}
	out := make([]string, 0, limit)
	appendPath := func(path string) {
		path = strings.TrimSpace(path)
		if path == "" || len(out) >= limit {
			return
		}
		out = append(out, path)
		out = dedupePDFVisualStrings(out)
	}
	for _, item := range topDocuments {
		path := strings.TrimSpace(item.Path)
		doc, ok := docsByPath[path]
		if !ok || !pdfStructuredDocumentNeedsFollowUp(doc) {
			continue
		}
		appendPath(path)
		if len(out) >= limit {
			return out
		}
	}
	if len(out) > 0 || len(topDocuments) <= 1 {
		return out
	}
	for _, item := range topDocuments[1:] {
		appendPath(item.Path)
		if len(out) >= limit {
			break
		}
	}
	return out
}

func pdfStructuredDocumentNeedsFollowUp(doc pdfStructuredPayload) bool {
	if doc.ReviewRequired {
		return true
	}
	switch strings.ToLower(strings.TrimSpace(doc.Status)) {
	case "partial", "empty":
		return true
	}
	if doc.ReviewSummary != nil {
		if doc.ReviewSummary.BatchesRequiringReview > 0 || doc.ReviewSummary.LowConfidenceFields > 0 {
			return true
		}
	}
	for _, batch := range doc.StructuredBatches {
		if batch.ReviewRequired || len(batch.MissingRequiredFields) > 0 || len(batch.LowConfidenceFields) > 0 {
			return true
		}
	}
	return false
}

func pdfStructuredTopFollowUpReasonCodes(
	documents []pdfStructuredPayload,
	topDocuments []pdfStructuredTopDocument,
	paths []string,
	limit int,
) []string {
	if len(paths) == 0 {
		return nil
	}
	docsByPath := pdfStructuredPayloadsByPath(documents)
	topDocsByPath := pdfStructuredTopDocumentsByPath(topDocuments)
	codes := make([]string, 0, len(paths)*4)
	for _, path := range paths {
		if topDoc, ok := topDocsByPath[path]; ok {
			codes = append(codes, topDoc.SelectionReasonCodes...)
		}
		if doc, ok := docsByPath[path]; ok && doc.ReviewSummary != nil {
			codes = append(codes, doc.ReviewSummary.ReviewReasonCodes...)
		}
	}
	return truncatePDFStructuredCodes(codes, limit)
}

func pdfStructuredTopFollowUpNotes(
	documents []pdfStructuredPayload,
	topDocuments []pdfStructuredTopDocument,
	paths []string,
	limit int,
) []string {
	if len(paths) == 0 {
		return nil
	}
	docsByPath := pdfStructuredPayloadsByPath(documents)
	topDocsByPath := pdfStructuredTopDocumentsByPath(topDocuments)
	notes := make([]string, 0, len(paths)*4)
	for _, path := range paths {
		if topDoc, ok := topDocsByPath[path]; ok {
			notes = append(notes, topDoc.SelectionReasons...)
			notes = append(notes, topDoc.TopNotes...)
		}
		if doc, ok := docsByPath[path]; ok && doc.ReviewSummary != nil {
			notes = append(notes, doc.ReviewSummary.ReviewDrivers...)
		}
	}
	return truncatePDFStructuredNotes(notes, limit)
}

func pdfStructuredPayloadsByPath(documents []pdfStructuredPayload) map[string]pdfStructuredPayload {
	if len(documents) == 0 {
		return nil
	}
	out := make(map[string]pdfStructuredPayload, len(documents))
	for _, doc := range documents {
		path := strings.TrimSpace(doc.Path)
		if path == "" {
			continue
		}
		out[path] = doc
	}
	return out
}

func pdfStructuredTopDocumentsByPath(topDocuments []pdfStructuredTopDocument) map[string]pdfStructuredTopDocument {
	if len(topDocuments) == 0 {
		return nil
	}
	out := make(map[string]pdfStructuredTopDocument, len(topDocuments))
	for _, doc := range topDocuments {
		path := strings.TrimSpace(doc.Path)
		if path == "" {
			continue
		}
		out[path] = doc
	}
	return out
}
