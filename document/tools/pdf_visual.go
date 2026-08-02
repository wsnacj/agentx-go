package tools

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"sort"
	"strings"

	types "github.com/wsnacj/agentx-go/components/llm"
	agentxmedia "github.com/wsnacj/agentx-go/runtime/mediaartifact"
)

const (
	defaultPDFVisualPages = 2
	hardPDFVisualPages    = 4
	defaultPDFVisualDPI   = 180
)

type pdfVisualAnalysis struct {
	Status            string                     `json:"status,omitempty"`
	Mode              string                     `json:"mode,omitempty"`
	Model             string                     `json:"model,omitempty"`
	Client            string                     `json:"client,omitempty"`
	NativePDF         bool                       `json:"native_pdf,omitempty"`
	DocumentCount     int                        `json:"document_count,omitempty"`
	PageCount         int                        `json:"page_count,omitempty"`
	Pages             []int                      `json:"pages,omitempty"`
	AttemptedModels   []string                   `json:"attempted_models,omitempty"`
	FallbackUsed      bool                       `json:"fallback_used,omitempty"`
	Summary           string                     `json:"summary,omitempty"`
	SignalProfile     pdfVisualSignalProfile     `json:"signal_profile,omitempty"`
	PageTargets       []pdfVisualPageTarget      `json:"page_targets,omitempty"`
	ExtractionBatches []pdfVisualExtractionBatch `json:"extraction_batches,omitempty"`
	Warning           string                     `json:"warning,omitempty"`
	RenderedImages    int                        `json:"rendered_images,omitempty"`
	RenderedPages     []agentxmedia.Descriptor   `json:"rendered_pages,omitempty"`
}

type pdfVisualPageTarget struct {
	Page     int    `json:"page"`
	Target   string `json:"target,omitempty"`
	Priority string `json:"priority,omitempty"`
	Reason   string `json:"reason,omitempty"`
}

type pdfVisualExtractionField struct {
	Name        string `json:"name,omitempty"`
	Kind        string `json:"kind,omitempty"`
	Required    bool   `json:"required,omitempty"`
	Description string `json:"description,omitempty"`
}

type pdfVisualResultSection struct {
	Name                  string                        `json:"name,omitempty"`
	Purpose               string                        `json:"purpose,omitempty"`
	FieldNames            []string                      `json:"field_names,omitempty"`
	RequiredAny           bool                          `json:"required_any,omitempty"`
	FieldFillRules        []string                      `json:"field_fill_rules,omitempty"`
	CompletionCriteria    []string                      `json:"completion_criteria,omitempty"`
	MissingFieldPolicy    string                        `json:"missing_field_policy,omitempty"`
	FieldEvidenceRules    map[string][]string           `json:"field_evidence_rules,omitempty"`
	FieldConfidencePolicy map[string]string             `json:"field_confidence_policy,omitempty"`
	FieldConflictPolicy   map[string]string             `json:"field_conflict_policy,omitempty"`
	FieldPriorityRules    map[string][]string           `json:"field_priority_rules,omitempty"`
	FieldSourceWeights    map[string]map[string]float64 `json:"field_source_weights,omitempty"`
	FieldResolutionOrder  map[string][]string           `json:"field_resolution_order,omitempty"`
	SummaryStyle          string                        `json:"summary_style,omitempty"`
	PostprocessRules      []string                      `json:"postprocess_rules,omitempty"`
	ExampleOutputs        []string                      `json:"example_outputs,omitempty"`
	QualityChecks         []string                      `json:"quality_checks,omitempty"`
	RenderTemplate        string                        `json:"render_template,omitempty"`
	RenderPolicy          string                        `json:"render_policy,omitempty"`
	RenderConditions      []string                      `json:"render_conditions,omitempty"`
	OmitWhenFieldsEmpty   bool                          `json:"omit_when_fields_empty,omitempty"`
}

type pdfVisualExtractionBatch struct {
	Target                 string                     `json:"target,omitempty"`
	Pages                  []int                      `json:"pages,omitempty"`
	Priority               string                     `json:"priority,omitempty"`
	Reasons                []string                   `json:"reasons,omitempty"`
	SummaryMode            string                     `json:"summary_mode,omitempty"`
	SummaryOutline         []string                   `json:"summary_outline,omitempty"`
	ExtractionTargets      []string                   `json:"extraction_targets,omitempty"`
	ExtractionSchema       []pdfVisualExtractionField `json:"extraction_schema,omitempty"`
	ResultSections         []pdfVisualResultSection   `json:"result_sections,omitempty"`
	ResultTemplate         map[string]any             `json:"result_template,omitempty"`
	ValidationChecks       []string                   `json:"validation_checks,omitempty"`
	NormalizationRules     []string                   `json:"normalization_rules,omitempty"`
	ExtractionInstructions []string                   `json:"extraction_instructions,omitempty"`
	AggregationStrategy    string                     `json:"aggregation_strategy,omitempty"`
	AggregationRules       []string                   `json:"aggregation_rules,omitempty"`
}

type pdfVisualSignalProfile struct {
	LayoutType          string                     `json:"layout_type,omitempty"`
	SummaryMode         string                     `json:"summary_mode,omitempty"`
	PrimaryVisualTarget string                     `json:"primary_visual_target,omitempty"`
	ChartLike           bool                       `json:"chart_like,omitempty"`
	TableLike           bool                       `json:"table_like,omitempty"`
	DiagramLike         bool                       `json:"diagram_like,omitempty"`
	ImageDocument       bool                       `json:"image_document,omitempty"`
	TextSparse          bool                       `json:"text_sparse,omitempty"`
	SummaryOutline      []string                   `json:"summary_outline,omitempty"`
	ExtractionTargets   []string                   `json:"extraction_targets,omitempty"`
	ExtractionSchema    []pdfVisualExtractionField `json:"extraction_schema,omitempty"`
	ConfidenceNotes     []string                   `json:"confidence_notes,omitempty"`
	FocusAreas          []string                   `json:"focus_areas,omitempty"`
	SuggestedFollowUps  []string                   `json:"suggested_follow_ups,omitempty"`
}

var pdfVisualSystemPrompt = "You analyze rendered PDF pages. Focus on layout, charts, tables, figures, and text that may not be captured by raw PDF extraction. Return concise, high-signal findings."

func pdfVisionAnalyzeWithInput(ctx context.Context, input types.VisionInput) (*types.VisualResponse, error) {
	host := pdfHostFromContext(ctx)
	if host.Vision == nil {
		return nil, fmt.Errorf("pdf vision analyzer is not configured")
	}
	return host.Vision(ctx, input)
}

func pdfNativeAnalyze(ctx context.Context, candidate pdfVisionModelCandidate, prompt string, paths []string) (string, error) {
	return runPDFNativeAnalysis(ctx, candidate, prompt, paths)
}

func pdfRenderPDFPages(ctx context.Context, path string, pages []int, dpi int) ([]pdfRenderedPage, func() error, error) {
	host := pdfHostFromContext(ctx)
	if host.Render == nil {
		return nil, nil, fmt.Errorf("pdf page renderer is not configured")
	}
	return host.Render(ctx, path, append([]int(nil), pages...), dpi)
}

func runPDFVisualAnalysis(ctx context.Context, root string, path string, query string, pageMap []pdfAnalyzePageItem, mediaProfile pdfMediaProfile, plan pdfAnalysisPlan, modelOverride string, maxPages int) pdfVisualAnalysis {
	analysis := pdfVisualAnalysis{
		Status: "skipped",
		Mode:   plan.Mode,
	}
	if !plan.NeedsVision {
		analysis.Warning = "visual analysis not required for this document profile"
		return analysis
	}

	candidates := append([]pdfVisionModelCandidate(nil), plan.CandidateModels...)
	if override := strings.TrimSpace(modelOverride); override != "" {
		overrideCandidate := pdfVisionModelCandidate{Name: override}
		for _, candidate := range candidates {
			if strings.EqualFold(strings.TrimSpace(candidate.Name), override) {
				overrideCandidate = candidate
				break
			}
		}
		candidates = []pdfVisionModelCandidate{overrideCandidate}
	}
	if len(candidates) == 0 {
		analysis.Status = "unavailable"
		analysis.Warning = strings.TrimSpace(plan.Warning)
		if analysis.Warning == "" {
			analysis.Warning = "no vision-capable model is configured"
		}
		return analysis
	}

	pages := selectPDFVisualPages(pageMap, mediaProfile, maxPages)
	var rendered []pdfRenderedPage
	var visuals []types.VisualContent
	var cleanup func() error
	var preparedRender bool
	var renderedPageWarning string
	prepareRenderedVisuals := func() error {
		if preparedRender {
			return nil
		}
		preparedRender = true
		if len(pages) == 0 {
			return fmt.Errorf("no suitable visual pages selected")
		}
		var err error
		rendered, cleanup, err = pdfRenderPDFPages(ctx, path, pages, defaultPDFVisualDPI)
		if err != nil {
			return fmt.Errorf("render visual pages: %w", err)
		}
		visuals, err = buildPDFVisualContents(rendered, pageMap)
		if err != nil {
			return fmt.Errorf("prepare visual payload: %w", err)
		}
		renderedPages, artifactErr := persistPDFRenderedPageArtifacts(ctx, root, path, rendered)
		if artifactErr != nil {
			renderedPageWarning = fmt.Sprintf("rendered page artifacts unavailable: %v", artifactErr)
		} else {
			analysis.RenderedPages = renderedPages
		}
		return nil
	}
	defer func() {
		if cleanup != nil {
			_ = cleanup()
		}
	}()
	var firstErr error
	for idx, candidate := range candidates {
		model := strings.TrimSpace(candidate.Name)
		if model == "" {
			continue
		}
		analysis.AttemptedModels = append(analysis.AttemptedModels, model)
		configName := pdfVisionCandidateConfigName(candidate)
		var summary string
		var err error
		nativeUsed := false
		if candidate.NativePDF {
			summary, err = pdfNativeAnalyze(ctx, candidate, buildPDFNativePrompt(query, mediaProfile, plan), []string{path})
			nativeUsed = err == nil
		} else {
			err = prepareRenderedVisuals()
			if err == nil {
				var resp *types.VisualResponse
				resp, err = pdfVisionAnalyzeWithInput(ctx, types.VisionInput{
					ConfigName:   configName,
					SystemPrompt: pdfVisualSystemPrompt,
					Messages: toolUserConversation(
						buildPDFVisualPrompt(query, pageMap, rendered, mediaProfile, plan),
					),
					Visuals: visuals,
				})
				if err == nil && resp != nil {
					summary = strings.TrimSpace(resp.Content)
				}
			}
		}
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		analysis.Status = "success"
		analysis.Model = model
		analysis.Client = strings.TrimSpace(candidate.Client)
		analysis.NativePDF = nativeUsed
		analysis.Pages = append([]int(nil), pages...)
		analysis.PageCount = len(pages)
		analysis.RenderedImages = len(rendered)
		analysis.FallbackUsed = idx > 0
		analysis.Summary = summary
		analysis.SignalProfile = buildPDFVisualSignalProfile(analysis.Summary, mediaProfile, plan)
		analysis.PageTargets = buildPDFVisualPageTargets(pageMap, pages, mediaProfile, analysis.SignalProfile)
		analysis.ExtractionBatches = buildPDFVisualExtractionBatches(analysis.PageTargets, analysis.SignalProfile)
		if analysis.Summary == "" {
			analysis.Warning = "vision backend returned empty summary"
		} else if idx > 0 && firstErr != nil {
			analysis.Warning = fmt.Sprintf("primary vision model failed; used fallback %s", model)
		}
		if strings.TrimSpace(renderedPageWarning) != "" {
			analysis.Warning = mergePDFToolWarnings(analysis.Warning, renderedPageWarning)
		}
		return analysis
	}
	analysis.Status = "failed"
	if len(analysis.AttemptedModels) > 0 {
		analysis.Model = analysis.AttemptedModels[0]
	}
	analysis.Pages = append([]int(nil), pages...)
	analysis.PageCount = len(pages)
	analysis.RenderedImages = len(rendered)
	if firstErr != nil {
		analysis.Warning = fmt.Sprintf("vision analysis failed across %d candidate models: %v", len(analysis.AttemptedModels), firstErr)
	}
	if strings.TrimSpace(renderedPageWarning) != "" {
		analysis.Warning = mergePDFToolWarnings(analysis.Warning, renderedPageWarning)
	}
	return analysis
}

func runPDFDocumentSetVisualAnalysis(
	ctx context.Context,
	paths []string,
	query string,
	artifactsList []pdfAnalysisArtifacts,
	plan pdfAnalysisPlan,
	modelOverride string,
) *pdfVisualAnalysis {
	analysis := &pdfVisualAnalysis{
		Status:        "skipped",
		Mode:          plan.Mode,
		DocumentCount: len(paths),
	}
	if !plan.NeedsVision {
		analysis.Warning = "document-set visual analysis not required for this pdf set"
		return analysis
	}

	candidates := append([]pdfVisionModelCandidate(nil), plan.CandidateModels...)
	override := strings.TrimSpace(modelOverride)
	if override != "" {
		overrideCandidate := pdfVisionModelCandidate{}
		for _, candidate := range candidates {
			if strings.EqualFold(strings.TrimSpace(candidate.Name), override) || strings.EqualFold(strings.TrimSpace(candidate.ConfigKey), override) {
				overrideCandidate = candidate
				break
			}
		}
		if strings.TrimSpace(overrideCandidate.Name) == "" {
			overrideCandidate = pdfVisionModelCandidate{Name: override, ConfigKey: override}
		}
		candidates = []pdfVisionModelCandidate{overrideCandidate}
	}
	candidates = filterPDFNativeCandidates(candidates)
	if len(candidates) == 0 {
		analysis.Status = "unavailable"
		analysis.Warning = strings.TrimSpace(plan.Warning)
		if override != "" {
			analysis.Warning = fmt.Sprintf("vision model override %s does not support native pdf document-set analysis", override)
		} else if strings.TrimSpace(plan.NativeProviderRouting) == "" {
			analysis.Warning = "no native pdf-capable llmx submodel is currently configured for document-set analysis"
		} else if analysis.Warning == "" {
			analysis.Warning = "native pdf document-set analysis is unavailable"
		}
		return analysis
	}

	aggregateMedia := buildPDFAggregateMediaProfile(artifactsList)
	totalPages := 0
	for _, artifacts := range artifactsList {
		totalPages += artifacts.Metadata.PageCount
	}
	prompt := buildPDFDocumentSetVisualPrompt(query, artifactsList, plan)
	var firstErr error
	for idx, candidate := range candidates {
		model := strings.TrimSpace(candidate.Name)
		if model == "" {
			continue
		}
		analysis.AttemptedModels = append(analysis.AttemptedModels, model)
		summary, err := pdfNativeAnalyze(ctx, candidate, prompt, paths)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		analysis.Status = "success"
		analysis.Model = model
		analysis.Client = strings.TrimSpace(candidate.Client)
		analysis.NativePDF = true
		analysis.PageCount = totalPages
		analysis.FallbackUsed = idx > 0
		analysis.Summary = strings.TrimSpace(summary)
		analysis.SignalProfile = buildPDFVisualSignalProfile(analysis.Summary, aggregateMedia, plan)
		if analysis.Summary == "" {
			analysis.Warning = "native pdf provider returned empty summary"
		} else if idx > 0 && firstErr != nil {
			analysis.Warning = fmt.Sprintf("primary native pdf model failed; used fallback %s", model)
		}
		return analysis
	}
	analysis.Status = "failed"
	if len(analysis.AttemptedModels) > 0 {
		analysis.Model = analysis.AttemptedModels[0]
	}
	analysis.PageCount = totalPages
	if firstErr != nil {
		analysis.Warning = fmt.Sprintf("native pdf document-set analysis failed across %d candidate models: %v", len(analysis.AttemptedModels), firstErr)
	}
	return analysis
}

func buildPDFDocumentSetVisualPrompt(query string, artifactsList []pdfAnalysisArtifacts, plan pdfAnalysisPlan) string {
	lines := []string{
		"Analyze the provided PDF document set directly and compare the documents at a high level.",
		fmt.Sprintf("Planned mode: %s.", plan.Mode),
		fmt.Sprintf("Document count: %d.", len(artifactsList)),
		"Focus on cross-document similarities, differences, standout charts/tables/diagrams, and any important OCR-visible text or layout cues.",
		"Return a concise summary grounded in the visible document content.",
	}
	if trimmed := strings.TrimSpace(query); trimmed != "" {
		lines = append(lines, fmt.Sprintf("Focus query: %s.", trimmed))
	}
	lines = append(lines, "Documents:")
	for _, artifacts := range artifactsList {
		lines = append(lines, fmt.Sprintf("- %s (%d pages): %s", artifacts.DisplayPath, artifacts.Metadata.PageCount, truncateToolText(joinPDFPageTexts(artifacts.TextResult.Pages), 180)))
	}
	lines = append(lines,
		"Highlight the most relevant document or documents for follow-up analysis.",
		"If the documents conflict, call that out explicitly and mention the visible evidence driving the difference.",
	)
	return strings.Join(lines, "\n")
}

func selectPDFVisualPages(pageMap []pdfAnalyzePageItem, mediaProfile pdfMediaProfile, maxPages int) []int {
	limit := clampToolLimit(maxPages, defaultPDFVisualPages, hardPDFVisualPages)
	selected := make([]int, 0, limit)
	seen := make(map[int]struct{}, limit)
	add := func(page int) {
		if page <= 0 || len(selected) >= limit {
			return
		}
		if _, ok := seen[page]; ok {
			return
		}
		seen[page] = struct{}{}
		selected = append(selected, page)
	}
	for _, page := range mediaProfile.GraphicHeavyPages {
		add(page)
	}
	for _, item := range pageMap {
		if item.Page > 0 && item.Chars > 0 {
			add(item.Page)
		}
		if len(selected) >= limit {
			break
		}
	}
	sort.Ints(selected)
	return selected
}

func buildPDFVisualContents(rendered []pdfRenderedPage, pageMap []pdfAnalyzePageItem) ([]types.VisualContent, error) {
	excerpts := make(map[int]string, len(pageMap))
	for _, item := range pageMap {
		if strings.TrimSpace(item.Excerpt) == "" {
			continue
		}
		excerpts[item.Page] = item.Excerpt
	}
	visuals := make([]types.VisualContent, 0, len(rendered)*2)
	for _, page := range rendered {
		if excerpt := strings.TrimSpace(excerpts[page.Page]); excerpt != "" {
			visuals = append(visuals, types.NewTextBlock(fmt.Sprintf("Page %d extracted text excerpt: %s", page.Page, excerpt)))
		}
		dataURI, err := encodePDFVisualDataURI(page)
		if err != nil {
			return nil, err
		}
		visuals = append(visuals, types.VisualContent{
			Type:    "image_url",
			DataURI: dataURI,
			Detail:  types.DetailHigh,
			Labels:  []string{fmt.Sprintf("page-%d", page.Page)},
		})
	}
	return visuals, nil
}

func encodePDFVisualDataURI(page pdfRenderedPage) (string, error) {
	data := page.Data
	if len(data) == 0 {
		return "", fmt.Errorf("rendered page %d has no image data", page.Page)
	}
	mimeType := strings.TrimSpace(page.MIMEType)
	if mimeType == "" {
		mimeType = http.DetectContentType(data)
	}
	format := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(mimeType)), "image/")
	if format == "" || format == mimeType {
		return "", fmt.Errorf("unsupported visual image format %q", mimeType)
	}
	return fmt.Sprintf("data:image/%s;base64,%s", format, base64.StdEncoding.EncodeToString(data)), nil
}

func mergePDFToolWarnings(current string, extra string) string {
	current = strings.TrimSpace(current)
	extra = strings.TrimSpace(extra)
	switch {
	case current == "":
		return extra
	case extra == "":
		return current
	case strings.Contains(current, extra):
		return current
	default:
		return current + "; " + extra
	}
}

func buildPDFVisualPrompt(query string, pageMap []pdfAnalyzePageItem, rendered []pdfRenderedPage, mediaProfile pdfMediaProfile, plan pdfAnalysisPlan) string {
	promptMode := selectPDFVisualPromptMode(mediaProfile, plan)
	lines := []string{
		"Analyze the rendered PDF pages and summarize visual/layout insights that are not obvious from plain text extraction.",
		fmt.Sprintf("Planned mode: %s.", plan.Mode),
		fmt.Sprintf("Specialized visual prompt mode: %s.", promptMode),
	}
	if strings.TrimSpace(query) != "" {
		lines = append(lines, fmt.Sprintf("Focus query: %s.", strings.TrimSpace(query)))
	}
	excerpts := make(map[int]string, len(pageMap))
	for _, item := range pageMap {
		if strings.TrimSpace(item.Excerpt) == "" {
			continue
		}
		excerpts[item.Page] = item.Excerpt
	}
	for _, page := range rendered {
		if excerpt := strings.TrimSpace(excerpts[page.Page]); excerpt != "" {
			lines = append(lines, fmt.Sprintf("Page %d text excerpt: %s", page.Page, excerpt))
		} else {
			lines = append(lines, fmt.Sprintf("Page %d has little or no extracted text; rely more on visual/OCR cues.", page.Page))
		}
	}
	lines = append(lines,
		"Mention charts, tables, slide-like layouts, diagrams, and OCR text when relevant.",
		"Prefer a concise 3-5 bullet style summary covering layout type, dominant visual elements, OCR/text cues, and the most useful follow-up extraction target.",
	)
	lines = append(lines, pdfVisualPromptInstructions(promptMode)...)
	return strings.Join(lines, "\n")
}

func buildPDFVisualSignalProfile(summary string, mediaProfile pdfMediaProfile, plan pdfAnalysisPlan) pdfVisualSignalProfile {
	lower := strings.ToLower(strings.TrimSpace(summary))
	profile := pdfVisualSignalProfile{
		TextSparse:    plan.NeedsOCR,
		ImageDocument: plan.Mode == "vision_ocr",
	}
	switch {
	case mediaProfile.LikelySlideDeck || strings.Contains(lower, "slide"):
		profile.LayoutType = "slide_deck"
	case plan.Mode == "vision_ocr" || strings.Contains(lower, "scanned"):
		profile.LayoutType = "scanned_document"
	case containsAny(lower, "table", "tabular", "spreadsheet", "rows", "columns"):
		profile.LayoutType = "tabular_report"
	case containsAny(lower, "chart", "graph", "trend", "bar", "line", "axis", "plot"):
		profile.LayoutType = "chart_report"
	case mediaProfile.LikelyGraphicDoc:
		profile.LayoutType = "graphic_report"
	default:
		profile.LayoutType = "mixed_document"
	}
	profile.ChartLike = containsAny(lower, "chart", "graph", "trend", "bar", "line", "pie", "plot", "axis")
	profile.TableLike = containsAny(lower, "table", "tabular", "spreadsheet", "grid", "rows", "columns")
	profile.DiagramLike = containsAny(lower, "diagram", "flowchart", "workflow", "architecture", "schematic", "org chart")
	if !profile.ImageDocument {
		profile.ImageDocument = containsAny(lower, "scanned", "ocr", "screenshot", "photo", "raster")
	}
	profile.SummaryMode, profile.PrimaryVisualTarget = classifyPDFVisualSummaryMode(profile, plan)
	profile.SummaryOutline = buildPDFVisualSummaryOutline(profile)
	profile.ExtractionTargets = buildPDFVisualExtractionTargets(profile)
	profile.ExtractionSchema = buildPDFVisualExtractionSchema(profile)
	profile.ConfidenceNotes = buildPDFVisualConfidenceNotes(profile, plan)
	profile.FocusAreas = buildPDFVisualFocusAreas(profile)
	profile.SuggestedFollowUps = buildPDFVisualFollowUps(profile)
	return profile
}

func buildPDFVisualSummaryOutline(profile pdfVisualSignalProfile) []string {
	outline := []string{"layout_type", "dominant_visual", "key_takeaway"}
	switch profile.SummaryMode {
	case "slide_chart_summary":
		outline = append(outline, "slide_headline", "chart_trend", "supporting_annotation")
	case "table_summary":
		outline = append(outline, "table_subject", "header_structure", "important_rows_or_totals")
	case "diagram_summary":
		outline = append(outline, "diagram_subject", "main_components", "relationship_flow")
	case "ocr_layout_summary":
		outline = append(outline, "ocr_headings", "captions_or_callouts", "scan_quality_notes")
	case "chart_summary":
		outline = append(outline, "chart_type", "series_and_axes", "main_numeric_change")
	default:
		outline = append(outline, "section_structure")
	}
	return dedupePDFVisualStrings(outline)
}

func buildPDFVisualExtractionTargets(profile pdfVisualSignalProfile) []string {
	targets := make([]string, 0, 6)
	switch profile.PrimaryVisualTarget {
	case "chart":
		targets = append(targets, "chart_type", "chart_axes", "chart_series", "chart_annotations")
	case "table":
		targets = append(targets, "table_headers", "table_key_rows", "table_totals", "table_footnotes")
	case "diagram":
		targets = append(targets, "diagram_nodes", "diagram_edges", "diagram_labels", "diagram_flow")
	case "ocr_text":
		targets = append(targets, "ocr_title_blocks", "ocr_section_headers", "ocr_captions", "ocr_annotations")
	default:
		targets = append(targets, "page_layout", "visual_priority_regions")
	}
	if profile.LayoutType == "slide_deck" {
		targets = append(targets, "slide_title", "slide_supporting_visual")
	}
	return dedupePDFVisualStrings(targets)
}

func buildPDFVisualExtractionSchema(profile pdfVisualSignalProfile) []pdfVisualExtractionField {
	switch profile.SummaryMode {
	case "slide_chart_summary":
		return []pdfVisualExtractionField{
			{Name: "slide_title", Kind: "string", Required: true, Description: "Primary headline or title shown on the slide."},
			{Name: "chart_type", Kind: "string", Required: true, Description: "Dominant chart form, such as bar, line, pie, or combo."},
			{Name: "main_trend", Kind: "string", Required: true, Description: "Most important trend or comparison visible in the chart."},
			{Name: "supporting_annotation", Kind: "string", Required: false, Description: "Caption, callout, or side note reinforcing the chart takeaway."},
		}
	case "table_summary":
		return []pdfVisualExtractionField{
			{Name: "table_subject", Kind: "string", Required: true, Description: "What the table is about."},
			{Name: "key_headers", Kind: "string_list", Required: true, Description: "Most important row or column headers."},
			{Name: "key_rows", Kind: "string_list", Required: false, Description: "Rows or row groups that matter most."},
			{Name: "totals_or_comparisons", Kind: "string_list", Required: false, Description: "Important totals, deltas, or comparison values."},
		}
	case "diagram_summary":
		return []pdfVisualExtractionField{
			{Name: "diagram_subject", Kind: "string", Required: true, Description: "What the diagram or workflow depicts."},
			{Name: "main_components", Kind: "string_list", Required: true, Description: "Important nodes, boxes, or system components."},
			{Name: "relationships", Kind: "string_list", Required: true, Description: "Key relationships, connectors, or directional flows."},
			{Name: "legend_or_labels", Kind: "string_list", Required: false, Description: "Labels, legends, or annotations explaining the diagram."},
		}
	case "ocr_layout_summary":
		return []pdfVisualExtractionField{
			{Name: "primary_heading", Kind: "string", Required: true, Description: "Top-level heading or title recovered by OCR."},
			{Name: "section_headers", Kind: "string_list", Required: false, Description: "Important section titles or structural markers."},
			{Name: "captions_or_callouts", Kind: "string_list", Required: false, Description: "Captions, notes, or marginal annotations."},
			{Name: "scan_quality_notes", Kind: "string_list", Required: false, Description: "Anything that lowers OCR confidence, such as blur or raster artifacts."},
		}
	case "chart_summary":
		return []pdfVisualExtractionField{
			{Name: "chart_type", Kind: "string", Required: true, Description: "Dominant chart form."},
			{Name: "axes_or_categories", Kind: "string_list", Required: false, Description: "Visible axes, categories, or grouping dimensions."},
			{Name: "series_labels", Kind: "string_list", Required: false, Description: "Series, legends, or data group labels."},
			{Name: "main_trend", Kind: "string", Required: true, Description: "Most important numeric trend or comparison."},
		}
	default:
		return []pdfVisualExtractionField{
			{Name: "layout_type", Kind: "string", Required: true, Description: "Overall page or document layout category."},
			{Name: "dominant_regions", Kind: "string_list", Required: true, Description: "Most salient visual or textual regions to inspect."},
			{Name: "key_takeaway", Kind: "string", Required: true, Description: "Highest-value conclusion from the page or document layout."},
		}
	}
}

func buildPDFVisualConfidenceNotes(profile pdfVisualSignalProfile, plan pdfAnalysisPlan) []string {
	notes := make([]string, 0, 4)
	if profile.ImageDocument || profile.TextSparse || plan.Mode == "vision_ocr" || plan.NeedsOCR {
		notes = append(notes, "ocr_required_for_full_fidelity")
	}
	if profile.ImageDocument {
		notes = append(notes, "scan_or_raster_artifacts_may_reduce_accuracy")
	}
	if profile.ChartLike || profile.TableLike || profile.DiagramLike {
		notes = append(notes, "visual_interpretation_should_be_cross_checked_with_embedded_text_when_available")
	}
	if profile.LayoutType == "slide_deck" {
		notes = append(notes, "slide_takeaways_may_be_abbreviated_and_context_dependent")
	}
	return dedupePDFVisualStrings(notes)
}

func buildPDFVisualPageTargets(pageMap []pdfAnalyzePageItem, pages []int, mediaProfile pdfMediaProfile, profile pdfVisualSignalProfile) []pdfVisualPageTarget {
	if len(pages) == 0 {
		return nil
	}
	charsByPage := make(map[int]int, len(pageMap))
	for _, item := range pageMap {
		charsByPage[item.Page] = item.Chars
	}
	graphicHeavy := make(map[int]struct{}, len(mediaProfile.GraphicHeavyPages))
	for _, page := range mediaProfile.GraphicHeavyPages {
		graphicHeavy[page] = struct{}{}
	}
	targets := make([]pdfVisualPageTarget, 0, len(pages))
	for idx, page := range pages {
		target := pdfVisualPageTarget{Page: page}
		chars := charsByPage[page]
		_, isGraphicHeavy := graphicHeavy[page]
		switch {
		case profile.ImageDocument || profile.TextSparse || chars <= 40:
			target.Target = "ocr_text"
			target.Reason = "text is sparse or OCR-dependent on this page"
		case profile.PrimaryVisualTarget == "chart" && isGraphicHeavy:
			target.Target = "chart"
			target.Reason = "graphics-heavy page likely contains a chart or chart-like visual"
		case profile.PrimaryVisualTarget == "table" && (isGraphicHeavy || chars >= 180):
			target.Target = "table"
			target.Reason = "page likely contains tabular structure worth targeted extraction"
		case profile.PrimaryVisualTarget == "diagram" && isGraphicHeavy:
			target.Target = "diagram"
			target.Reason = "graphics-heavy page likely contains diagram structure and connectors"
		case profile.LayoutType == "slide_deck":
			target.Target = "slide_visual"
			target.Reason = "slide-like page should be summarized as a headline plus supporting visual"
		default:
			target.Target = "layout"
			target.Reason = "page should be interpreted primarily by layout and dominant visual regions"
		}
		switch {
		case idx == 0 || isGraphicHeavy:
			target.Priority = "high"
		case chars == 0 || chars <= 80:
			target.Priority = "medium"
		default:
			target.Priority = "low"
		}
		targets = append(targets, target)
	}
	return targets
}

func buildPDFVisualExtractionBatches(pageTargets []pdfVisualPageTarget, profile pdfVisualSignalProfile) []pdfVisualExtractionBatch {
	if len(pageTargets) == 0 {
		return nil
	}
	type batchState struct {
		pages    []int
		priority string
		reasons  []string
	}
	order := make([]string, 0, len(pageTargets))
	stateByTarget := make(map[string]*batchState, len(pageTargets))
	for _, item := range pageTargets {
		target := strings.TrimSpace(item.Target)
		if target == "" {
			target = "layout"
		}
		state := stateByTarget[target]
		if state == nil {
			state = &batchState{}
			stateByTarget[target] = state
			order = append(order, target)
		}
		state.pages = append(state.pages, item.Page)
		state.priority = strongerPDFVisualPriority(state.priority, item.Priority)
		if reason := strings.TrimSpace(item.Reason); reason != "" {
			state.reasons = append(state.reasons, reason)
		}
	}
	out := make([]pdfVisualExtractionBatch, 0, len(order))
	for _, target := range order {
		state := stateByTarget[target]
		derived := derivePDFVisualSignalProfileForTarget(profile, target)
		sort.Ints(state.pages)
		out = append(out, pdfVisualExtractionBatch{
			Target:                 target,
			Pages:                  append([]int(nil), state.pages...),
			Priority:               state.priority,
			Reasons:                dedupePDFVisualStrings(state.reasons),
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
		})
	}
	return out
}

func derivePDFVisualSignalProfileForTarget(profile pdfVisualSignalProfile, target string) pdfVisualSignalProfile {
	derived := profile
	switch strings.TrimSpace(target) {
	case "chart":
		derived.SummaryMode = "chart_summary"
		if derived.LayoutType == "slide_deck" {
			derived.SummaryMode = "slide_chart_summary"
		}
		derived.PrimaryVisualTarget = "chart"
		derived.ChartLike = true
	case "table":
		derived.SummaryMode = "table_summary"
		derived.PrimaryVisualTarget = "table"
		derived.TableLike = true
	case "diagram":
		derived.SummaryMode = "diagram_summary"
		derived.PrimaryVisualTarget = "diagram"
		derived.DiagramLike = true
	case "ocr_text":
		derived.SummaryMode = "ocr_layout_summary"
		derived.PrimaryVisualTarget = "ocr_text"
		derived.TextSparse = true
		derived.ImageDocument = true
	case "slide_visual":
		derived.SummaryMode = "slide_visual_summary"
		derived.PrimaryVisualTarget = "layout"
	default:
		derived.SummaryMode = "layout_summary"
		derived.PrimaryVisualTarget = "layout"
	}
	derived.SummaryOutline = buildPDFVisualSummaryOutline(derived)
	derived.ExtractionTargets = buildPDFVisualExtractionTargets(derived)
	derived.ExtractionSchema = buildPDFVisualExtractionSchema(derived)
	return derived
}

func strongerPDFVisualPriority(current string, next string) string {
	rank := map[string]int{
		"high":   3,
		"medium": 2,
		"low":    1,
		"":       0,
	}
	if rank[next] > rank[current] {
		return next
	}
	return current
}

func buildPDFVisualResultTemplate(schema []pdfVisualExtractionField) map[string]any {
	if len(schema) == 0 {
		return nil
	}
	out := make(map[string]any, len(schema))
	for _, field := range schema {
		name := strings.TrimSpace(field.Name)
		if name == "" {
			continue
		}
		out[name] = defaultPDFVisualTemplateValue(field.Kind)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func buildPDFVisualResultSections(profile pdfVisualSignalProfile) []pdfVisualResultSection {
	switch profile.SummaryMode {
	case "slide_chart_summary":
		return finalizePDFVisualResultSections([]pdfVisualResultSection{
			{
				Name:        "headline",
				Purpose:     "Capture the slide-level framing before chart details.",
				FieldNames:  []string{"slide_title"},
				RequiredAny: true,
				FieldFillRules: []string{
					"fill slide_title from the clearest visible title or headline on the earliest high-priority page",
				},
				SummaryStyle:        "single_line_title",
				PostprocessRules:    []string{"trim filler punctuation from title edges", "preserve original title casing when it appears intentional"},
				ExampleOutputs:      []string{"Slide title: Q4 revenue review", "Slide title: 2026 roadmap overview"},
				QualityChecks:       []string{"title fits on one line", "title contains no page numbering noise"},
				RenderTemplate:      "Slide title: {{slide_title}}",
				RenderPolicy:        "always_when_present",
				RenderConditions:    []string{"render when slide_title is non-empty"},
				OmitWhenFieldsEmpty: true,
			},
			{
				Name:        "chart_takeaway",
				Purpose:     "Summarize the main chart form and dominant trend.",
				FieldNames:  []string{"chart_type", "main_trend"},
				RequiredAny: true,
				FieldFillRules: []string{
					"fill chart_type before main_trend using only visible visual form and labels",
					"fill main_trend with one dominant trend or comparison grounded in the chart",
				},
				SummaryStyle:        "single_sentence_takeaway",
				PostprocessRules:    []string{"normalize chart_type to a common chart family label", "keep main_trend concise and grounded in visible evidence"},
				ExampleOutputs:      []string{"Chart (bar): revenue increases quarter over quarter", "Chart (line): signups decline after May"},
				QualityChecks:       []string{"contains exactly one dominant trend", "does not speculate beyond visible chart evidence"},
				RenderTemplate:      "Chart ({{chart_type}}): {{main_trend}}",
				RenderPolicy:        "always_when_present",
				RenderConditions:    []string{"render when chart_type or main_trend is non-empty"},
				OmitWhenFieldsEmpty: true,
			},
			{
				Name:        "supporting_notes",
				Purpose:     "Keep visible callouts or captions that materially support the chart takeaway.",
				FieldNames:  []string{"supporting_annotation"},
				RequiredAny: false,
				FieldFillRules: []string{
					"fill supporting_annotation only when captions or callouts are visibly present and relevant",
				},
				SummaryStyle:        "supporting_note",
				PostprocessRules:    []string{"drop decorative copy with no analytical value", "deduplicate repeated callout phrasing"},
				ExampleOutputs:      []string{"Supporting note: target excludes one-time promotions"},
				QualityChecks:       []string{"keeps only materially relevant annotations"},
				RenderTemplate:      "Supporting note: {{supporting_annotation}}",
				RenderPolicy:        "optional",
				RenderConditions:    []string{"render only when supporting_annotation is non-empty and not redundant with chart_takeaway"},
				OmitWhenFieldsEmpty: true,
			},
		})
	case "table_summary":
		return finalizePDFVisualResultSections([]pdfVisualResultSection{
			{
				Name:        "table_context",
				Purpose:     "State what the table is about before extracting details.",
				FieldNames:  []string{"table_subject"},
				RequiredAny: true,
				FieldFillRules: []string{
					"fill table_subject from the visible title, heading, or nearby label that best describes the table",
				},
				SummaryStyle:        "single_line_heading",
				PostprocessRules:    []string{"prefer exact visible subject phrasing", "remove duplicated punctuation and excess spacing"},
				ExampleOutputs:      []string{"Table subject: regional revenue by quarter"},
				QualityChecks:       []string{"subject is specific and source-aligned"},
				RenderTemplate:      "Table subject: {{table_subject}}",
				RenderPolicy:        "always_when_present",
				RenderConditions:    []string{"render when table_subject is non-empty"},
				OmitWhenFieldsEmpty: true,
			},
			{
				Name:        "table_structure",
				Purpose:     "Capture the visible header structure and key rows.",
				FieldNames:  []string{"key_headers", "key_rows"},
				RequiredAny: true,
				FieldFillRules: []string{
					"fill key_headers before key_rows using only visible row or column headers",
					"fill key_rows with the most important rows or row groups, not every row in the table",
				},
				SummaryStyle:        "structured_table_overview",
				PostprocessRules:    []string{"preserve visible header order", "keep row labels short and source-aligned"},
				ExampleOutputs:      []string{"Headers: region, q1, q2 | Key rows: China, Europe, US"},
				QualityChecks:       []string{"header order matches the page", "row labels are not expanded with invented detail"},
				RenderTemplate:      "Headers: {{key_headers}} | Key rows: {{key_rows}}",
				RenderPolicy:        "always_when_present",
				RenderConditions:    []string{"render when key_headers or key_rows has at least one visible entry"},
				OmitWhenFieldsEmpty: true,
			},
			{
				Name:        "table_highlights",
				Purpose:     "Preserve totals, comparisons, or emphasized values.",
				FieldNames:  []string{"totals_or_comparisons"},
				RequiredAny: false,
				FieldFillRules: []string{
					"fill totals_or_comparisons only with visually emphasized totals, deltas, or comparisons",
				},
				SummaryStyle:        "numeric_highlights",
				PostprocessRules:    []string{"normalize number formatting without changing units", "drop empty highlight placeholders"},
				ExampleOutputs:      []string{"Highlights: total revenue 12.4M; YoY +18%"},
				QualityChecks:       []string{"highlights keep units intact", "only emphasized values are included"},
				RenderTemplate:      "Highlights: {{totals_or_comparisons}}",
				RenderPolicy:        "optional",
				RenderConditions:    []string{"render only when totals_or_comparisons has at least one non-empty value"},
				OmitWhenFieldsEmpty: true,
			},
		})
	case "diagram_summary":
		return finalizePDFVisualResultSections([]pdfVisualResultSection{
			{
				Name:        "diagram_context",
				Purpose:     "Identify the diagram subject and scope.",
				FieldNames:  []string{"diagram_subject"},
				RequiredAny: true,
				FieldFillRules: []string{
					"fill diagram_subject with the clearest visible title or subject phrase describing the diagram",
				},
				SummaryStyle:        "single_line_heading",
				PostprocessRules:    []string{"prefer visible subject labels", "strip duplicated punctuation at edges"},
				ExampleOutputs:      []string{"Diagram subject: order fulfillment flow"},
				QualityChecks:       []string{"subject names the workflow or system clearly"},
				RenderTemplate:      "Diagram subject: {{diagram_subject}}",
				RenderPolicy:        "always_when_present",
				RenderConditions:    []string{"render when diagram_subject is non-empty"},
				OmitWhenFieldsEmpty: true,
			},
			{
				Name:        "diagram_components",
				Purpose:     "List the main boxes, nodes, or components.",
				FieldNames:  []string{"main_components"},
				RequiredAny: true,
				FieldFillRules: []string{
					"fill main_components with visible node or component labels after deduping synonymous repeats",
				},
				SummaryStyle:        "entity_list",
				PostprocessRules:    []string{"deduplicate repeated component names", "preserve component ordering when visually obvious"},
				ExampleOutputs:      []string{"Components: customer, payment service, warehouse"},
				QualityChecks:       []string{"component names are deduplicated", "ordering follows the visible flow when possible"},
				RenderTemplate:      "Components: {{main_components}}",
				RenderPolicy:        "always_when_present",
				RenderConditions:    []string{"render when main_components has at least one visible component"},
				OmitWhenFieldsEmpty: true,
			},
			{
				Name:        "diagram_relationships",
				Purpose:     "Describe directional links, flows, or dependencies.",
				FieldNames:  []string{"relationships", "legend_or_labels"},
				RequiredAny: true,
				FieldFillRules: []string{
					"fill relationships with directional flows or dependencies only when connectors are visible",
					"fill legend_or_labels with visible legends or explanatory labels when present",
				},
				SummaryStyle:        "relationship_summary",
				PostprocessRules:    []string{"prefer directional verbs when arrows are visible", "avoid inventing unlabeled relationships"},
				ExampleOutputs:      []string{"Relationships: customer submits order -> payment approves -> warehouse ships | Labels: SLA, fraud check"},
				QualityChecks:       []string{"relationships stay directional when arrows are visible", "labels remain separate from the relationship list"},
				RenderTemplate:      "Relationships: {{relationships}} | Labels: {{legend_or_labels}}",
				RenderPolicy:        "always_when_present",
				RenderConditions:    []string{"render when relationships or legend_or_labels contains visible content"},
				OmitWhenFieldsEmpty: true,
			},
		})
	case "ocr_layout_summary":
		return finalizePDFVisualResultSections([]pdfVisualResultSection{
			{
				Name:        "ocr_headings",
				Purpose:     "Capture the primary heading and major section markers.",
				FieldNames:  []string{"primary_heading", "section_headers"},
				RequiredAny: true,
				FieldFillRules: []string{
					"fill primary_heading from the most prominent OCR-recoverable heading",
					"fill section_headers with visible structural markers in page order",
				},
				SummaryStyle:        "ocr_structure_summary",
				PostprocessRules:    []string{"preserve heading order from the scanned page", "avoid over-correcting uncertain OCR spellings"},
				ExampleOutputs:      []string{"Heading: Insurance claim form | Sections: claimant details, incident summary, signatures"},
				QualityChecks:       []string{"section order follows the scanned page", "uncertain OCR text is not over-normalized"},
				RenderTemplate:      "Heading: {{primary_heading}} | Sections: {{section_headers}}",
				RenderPolicy:        "always_when_present",
				RenderConditions:    []string{"render when primary_heading or section_headers contains OCR-recovered structure"},
				OmitWhenFieldsEmpty: true,
			},
			{
				Name:        "ocr_notes",
				Purpose:     "Preserve callouts and scan quality issues that affect confidence.",
				FieldNames:  []string{"captions_or_callouts", "scan_quality_notes"},
				RequiredAny: false,
				FieldFillRules: []string{
					"fill captions_or_callouts only when they are visually distinct from body text",
					"fill scan_quality_notes whenever OCR confidence is lowered by blur, skew, or raster artifacts",
				},
				SummaryStyle:        "quality_notes",
				PostprocessRules:    []string{"keep notes short and uncertainty-focused", "drop generic warnings with no effect on extraction confidence"},
				ExampleOutputs:      []string{"Notes: handwritten margin note on page bottom | OCR quality: skewed scan, low contrast"},
				QualityChecks:       []string{"notes only mention confidence-affecting issues"},
				RenderTemplate:      "Notes: {{captions_or_callouts}} | OCR quality: {{scan_quality_notes}}",
				RenderPolicy:        "optional",
				RenderConditions:    []string{"render only when captions_or_callouts or scan_quality_notes contains non-empty content"},
				OmitWhenFieldsEmpty: true,
			},
		})
	case "chart_summary":
		return finalizePDFVisualResultSections([]pdfVisualResultSection{
			{
				Name:        "chart_identity",
				Purpose:     "Identify the chart form and visible labels.",
				FieldNames:  []string{"chart_type", "axes_or_categories", "series_labels"},
				RequiredAny: true,
				FieldFillRules: []string{
					"fill chart_type first, then axes_or_categories and series_labels using only visible labels",
				},
				SummaryStyle:        "chart_identity_summary",
				PostprocessRules:    []string{"normalize chart labels into a compact identity line", "preserve visible category order where possible"},
				ExampleOutputs:      []string{"Chart: line | Axes/categories: month, revenue | Series: product A, product B"},
				QualityChecks:       []string{"identity line includes chart form and visible labels", "series ordering follows the legend when visible"},
				RenderTemplate:      "Chart: {{chart_type}} | Axes/categories: {{axes_or_categories}} | Series: {{series_labels}}",
				RenderPolicy:        "always_when_present",
				RenderConditions:    []string{"render when chart_type or visible labels are present"},
				OmitWhenFieldsEmpty: true,
			},
			{
				Name:        "chart_takeaway",
				Purpose:     "Summarize the single most important trend or comparison.",
				FieldNames:  []string{"main_trend"},
				RequiredAny: true,
				FieldFillRules: []string{
					"fill main_trend with one dominant trend or comparison rather than multiple competing statements",
				},
				SummaryStyle:        "single_sentence_takeaway",
				PostprocessRules:    []string{"keep the takeaway to one clear statement", "avoid mixing multiple unrelated trends"},
				ExampleOutputs:      []string{"Main trend: revenue rises steadily from January to June"},
				QualityChecks:       []string{"takeaway is a single statement", "takeaway references only visible chart behavior"},
				RenderTemplate:      "Main trend: {{main_trend}}",
				RenderPolicy:        "always_when_present",
				RenderConditions:    []string{"render when main_trend is non-empty"},
				OmitWhenFieldsEmpty: true,
			},
		})
	default:
		return finalizePDFVisualResultSections([]pdfVisualResultSection{
			{
				Name:        "layout_overview",
				Purpose:     "Describe the overall layout and dominant regions.",
				FieldNames:  []string{"layout_type", "dominant_regions"},
				RequiredAny: true,
				FieldFillRules: []string{
					"fill layout_type first, then dominant_regions with the most salient visible regions in reading order",
				},
				SummaryStyle:        "layout_region_summary",
				PostprocessRules:    []string{"preserve top-to-bottom region ordering", "drop decorative regions with no extraction value"},
				ExampleOutputs:      []string{"Layout: report_page | Regions: title band, two-column body, footer notes"},
				QualityChecks:       []string{"regions are ordered top-to-bottom", "decorative regions are omitted unless analytically relevant"},
				RenderTemplate:      "Layout: {{layout_type}} | Regions: {{dominant_regions}}",
				RenderPolicy:        "always_when_present",
				RenderConditions:    []string{"render when layout_type or dominant_regions contains visible structure"},
				OmitWhenFieldsEmpty: true,
			},
			{
				Name:        "layout_takeaway",
				Purpose:     "Capture the highest-value visible conclusion.",
				FieldNames:  []string{"key_takeaway"},
				RequiredAny: true,
				FieldFillRules: []string{
					"fill key_takeaway with the highest-value visible conclusion without inferring hidden context",
				},
				SummaryStyle:        "single_sentence_takeaway",
				PostprocessRules:    []string{"prefer one page-level insight", "avoid speculative language not grounded in visible content"},
				ExampleOutputs:      []string{"Takeaway: the page is a comparison layout with the main result highlighted in the right column"},
				QualityChecks:       []string{"takeaway remains page-level", "takeaway does not infer hidden narrative"},
				RenderTemplate:      "Takeaway: {{key_takeaway}}",
				RenderPolicy:        "always_when_present",
				RenderConditions:    []string{"render when key_takeaway is non-empty"},
				OmitWhenFieldsEmpty: true,
			},
		})
	}
}

func finalizePDFVisualResultSections(sections []pdfVisualResultSection) []pdfVisualResultSection {
	for i := range sections {
		if len(sections[i].CompletionCriteria) == 0 {
			sections[i].CompletionCriteria = buildPDFVisualCompletionCriteria(sections[i])
		}
		if strings.TrimSpace(sections[i].MissingFieldPolicy) == "" {
			sections[i].MissingFieldPolicy = buildPDFVisualMissingFieldPolicy(sections[i])
		}
		if len(sections[i].FieldEvidenceRules) == 0 {
			sections[i].FieldEvidenceRules = buildPDFVisualFieldEvidenceRules(sections[i])
		}
		if len(sections[i].FieldConfidencePolicy) == 0 {
			sections[i].FieldConfidencePolicy = buildPDFVisualFieldConfidencePolicy(sections[i])
		}
		if len(sections[i].FieldConflictPolicy) == 0 {
			sections[i].FieldConflictPolicy = buildPDFVisualFieldConflictPolicy(sections[i])
		}
		if len(sections[i].FieldPriorityRules) == 0 {
			sections[i].FieldPriorityRules = buildPDFVisualFieldPriorityRules(sections[i])
		}
		if len(sections[i].FieldSourceWeights) == 0 {
			sections[i].FieldSourceWeights = buildPDFVisualFieldSourceWeights(sections[i])
		}
		if len(sections[i].FieldResolutionOrder) == 0 {
			sections[i].FieldResolutionOrder = buildPDFVisualFieldResolutionOrder(sections[i])
		}
	}
	return sections
}

func buildPDFVisualCompletionCriteria(section pdfVisualResultSection) []string {
	criteria := make([]string, 0, 4)
	fieldCount := len(section.FieldNames)
	switch {
	case section.RequiredAny && fieldCount == 1:
		criteria = append(criteria, "the primary field is populated with source-grounded content")
	case section.RequiredAny && fieldCount > 1:
		criteria = append(criteria, "at least one primary field is populated with source-grounded content")
	case !section.RequiredAny && fieldCount > 0:
		criteria = append(criteria, "when rendered, at least one optional field contains materially useful content")
	}
	if len(section.RenderConditions) > 0 {
		criteria = append(criteria, "render conditions are satisfied before the section is emitted")
	}
	switch strings.TrimSpace(section.SummaryStyle) {
	case "single_line_title", "single_line_heading":
		criteria = append(criteria, "rendered output stays on a single line")
	case "single_sentence_takeaway":
		criteria = append(criteria, "rendered takeaway remains a single concise statement")
	}
	if len(section.QualityChecks) > 0 {
		criteria = append(criteria, "quality checks pass without placeholder or speculative text")
	}
	return dedupePDFVisualStrings(criteria)
}

func buildPDFVisualMissingFieldPolicy(section pdfVisualResultSection) string {
	switch {
	case !section.RequiredAny:
		return "omit_optional_section"
	case section.OmitWhenFieldsEmpty:
		return "downgrade_to_partial_when_possible_else_omit"
	default:
		return "downgrade_to_partial_output"
	}
}

func buildPDFVisualFieldEvidenceRules(section pdfVisualResultSection) map[string][]string {
	if len(section.FieldNames) == 0 {
		return nil
	}
	out := make(map[string][]string, len(section.FieldNames))
	for _, field := range section.FieldNames {
		name := strings.TrimSpace(field)
		if name == "" {
			continue
		}
		rules := []string{"must_be_grounded_in_visible_page_evidence"}
		switch {
		case strings.Contains(name, "title"), strings.Contains(name, "heading"), strings.Contains(name, "subject"):
			rules = append(rules, "prefer explicit visible headings or labels over inferred paraphrases")
		case strings.Contains(name, "trend"), strings.Contains(name, "takeaway"), strings.Contains(name, "comparison"):
			rules = append(rules, "only derive from visible chart, table, or layout signals")
		case strings.Contains(name, "header"), strings.Contains(name, "row"), strings.Contains(name, "component"), strings.Contains(name, "relationship"), strings.Contains(name, "label"):
			rules = append(rules, "must map to visible structural labels or connectors")
		case strings.Contains(name, "annotation"), strings.Contains(name, "caption"), strings.Contains(name, "note"):
			rules = append(rules, "include only if visibly distinct and materially relevant")
		}
		out[name] = dedupePDFVisualStrings(rules)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func buildPDFVisualFieldConfidencePolicy(section pdfVisualResultSection) map[string]string {
	if len(section.FieldNames) == 0 {
		return nil
	}
	out := make(map[string]string, len(section.FieldNames))
	for _, field := range section.FieldNames {
		name := strings.TrimSpace(field)
		if name == "" {
			continue
		}
		policy := "allow_when_visually_supported"
		switch {
		case strings.Contains(name, "title"), strings.Contains(name, "heading"), strings.Contains(name, "subject"):
			policy = "require_explicit_label_or_strong_visual_heading"
		case strings.Contains(name, "trend"), strings.Contains(name, "takeaway"), strings.Contains(name, "comparison"):
			policy = "require_clear_primary_signal_else_omit"
		case strings.Contains(name, "annotation"), strings.Contains(name, "caption"), strings.Contains(name, "note"):
			policy = "emit_only_when_confidence_is_moderate_or_higher"
		case strings.Contains(name, "label"), strings.Contains(name, "row"), strings.Contains(name, "header"), strings.Contains(name, "component"), strings.Contains(name, "relationship"):
			policy = "prefer_exact_visible_tokens_else_keep_partial"
		}
		out[name] = policy
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func buildPDFVisualFieldConflictPolicy(section pdfVisualResultSection) map[string]string {
	if len(section.FieldNames) == 0 {
		return nil
	}
	out := make(map[string]string, len(section.FieldNames))
	for _, field := range section.FieldNames {
		name := strings.TrimSpace(field)
		if name == "" {
			continue
		}
		policy := "prefer_direct_page_evidence_then_omit"
		switch {
		case strings.Contains(name, "title"), strings.Contains(name, "heading"), strings.Contains(name, "subject"):
			policy = "prefer_visible_heading_then_caption_then_omit"
		case strings.Contains(name, "trend"), strings.Contains(name, "takeaway"), strings.Contains(name, "comparison"):
			policy = "prefer_dominant_primary_signal_then_supporting_callout"
		case strings.Contains(name, "header"), strings.Contains(name, "row"), strings.Contains(name, "component"), strings.Contains(name, "relationship"), strings.Contains(name, "label"):
			policy = "prefer_exact_structural_label_then_partial_visible_token"
		case strings.Contains(name, "annotation"), strings.Contains(name, "caption"), strings.Contains(name, "note"):
			policy = "prefer_explicit_callout_then_caption_then_omit"
		}
		out[name] = policy
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func buildPDFVisualFieldPriorityRules(section pdfVisualResultSection) map[string][]string {
	if len(section.FieldNames) == 0 {
		return nil
	}
	out := make(map[string][]string, len(section.FieldNames))
	for _, field := range section.FieldNames {
		name := strings.TrimSpace(field)
		if name == "" {
			continue
		}
		rules := []string{"prefer_high_priority_pages_over_medium_or_low_priority_pages"}
		switch {
		case strings.Contains(name, "title"), strings.Contains(name, "heading"), strings.Contains(name, "subject"):
			rules = append(rules, "prefer_top_of_page_or_visually_prominent_labels")
		case strings.Contains(name, "trend"), strings.Contains(name, "takeaway"), strings.Contains(name, "comparison"):
			rules = append(rules, "prefer the dominant chart_or_table signal before secondary notes")
		case strings.Contains(name, "header"), strings.Contains(name, "row"), strings.Contains(name, "component"), strings.Contains(name, "relationship"), strings.Contains(name, "label"):
			rules = append(rules, "prefer repeated or central structural labels over edge annotations")
		case strings.Contains(name, "annotation"), strings.Contains(name, "caption"), strings.Contains(name, "note"):
			rules = append(rules, "prefer concise supporting notes after primary content fields are filled")
		}
		out[name] = dedupePDFVisualStrings(rules)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func buildPDFVisualFieldSourceWeights(section pdfVisualResultSection) map[string]map[string]float64 {
	if len(section.FieldNames) == 0 {
		return nil
	}
	out := make(map[string]map[string]float64, len(section.FieldNames))
	for _, field := range section.FieldNames {
		name := strings.TrimSpace(field)
		if name == "" {
			continue
		}
		weights := map[string]float64{
			"high_priority_page":   1.0,
			"medium_priority_page": 0.75,
			"low_priority_page":    0.5,
		}
		switch {
		case strings.Contains(name, "title"), strings.Contains(name, "heading"), strings.Contains(name, "subject"):
			weights["visible_heading"] = 1.0
			weights["caption_or_subtitle"] = 0.7
			weights["supporting_note"] = 0.4
		case strings.Contains(name, "trend"), strings.Contains(name, "takeaway"), strings.Contains(name, "comparison"):
			weights["dominant_visual_signal"] = 1.0
			weights["supporting_callout"] = 0.65
			weights["ambient_context"] = 0.3
		case strings.Contains(name, "header"), strings.Contains(name, "row"), strings.Contains(name, "component"), strings.Contains(name, "relationship"), strings.Contains(name, "label"):
			weights["explicit_structural_label"] = 1.0
			weights["partial_visible_token"] = 0.6
			weights["inferred_structure"] = 0.2
		case strings.Contains(name, "annotation"), strings.Contains(name, "caption"), strings.Contains(name, "note"):
			weights["explicit_callout"] = 1.0
			weights["caption"] = 0.8
			weights["ambient_text"] = 0.25
		default:
			weights["direct_page_evidence"] = 1.0
			weights["secondary_context"] = 0.5
		}
		out[name] = weights
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func buildPDFVisualFieldResolutionOrder(section pdfVisualResultSection) map[string][]string {
	if len(section.FieldNames) == 0 {
		return nil
	}
	out := make(map[string][]string, len(section.FieldNames))
	for _, field := range section.FieldNames {
		name := strings.TrimSpace(field)
		if name == "" {
			continue
		}
		order := []string{"high_priority_page", "medium_priority_page", "low_priority_page"}
		switch {
		case strings.Contains(name, "title"), strings.Contains(name, "heading"), strings.Contains(name, "subject"):
			order = append([]string{"visible_heading", "caption_or_subtitle", "supporting_note"}, order...)
		case strings.Contains(name, "trend"), strings.Contains(name, "takeaway"), strings.Contains(name, "comparison"):
			order = append([]string{"dominant_visual_signal", "supporting_callout", "ambient_context"}, order...)
		case strings.Contains(name, "header"), strings.Contains(name, "row"), strings.Contains(name, "component"), strings.Contains(name, "relationship"), strings.Contains(name, "label"):
			order = append([]string{"explicit_structural_label", "partial_visible_token", "inferred_structure"}, order...)
		case strings.Contains(name, "annotation"), strings.Contains(name, "caption"), strings.Contains(name, "note"):
			order = append([]string{"explicit_callout", "caption", "ambient_text"}, order...)
		default:
			order = append([]string{"direct_page_evidence", "secondary_context"}, order...)
		}
		out[name] = dedupePDFVisualStrings(order)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func defaultPDFVisualTemplateValue(kind string) any {
	switch strings.TrimSpace(kind) {
	case "string_list":
		return []string{}
	case "number":
		return 0
	case "number_list":
		return []int{}
	case "bool":
		return false
	default:
		return ""
	}
}

func buildPDFVisualValidationChecks(profile pdfVisualSignalProfile) []string {
	checks := []string{"required_fields_present", "no_placeholder_values_left"}
	switch profile.SummaryMode {
	case "slide_chart_summary":
		checks = append(checks,
			"slide_title_not_empty",
			"main_trend_describes_direction_or_comparison",
			"chart_type_is_named_when_chart_present",
		)
	case "table_summary":
		checks = append(checks,
			"table_subject_not_empty",
			"key_headers_has_at_least_one_entry",
			"totals_or_comparisons_only_include_visible_values",
		)
	case "diagram_summary":
		checks = append(checks,
			"diagram_subject_not_empty",
			"main_components_has_at_least_one_entry",
			"relationships_describe_direction_or_linkage",
		)
	case "ocr_layout_summary":
		checks = append(checks,
			"primary_heading_not_empty_when_detected",
			"scan_quality_notes_capture_uncertainty",
			"ocr_fields_match_visible_regions_only",
		)
	case "chart_summary":
		checks = append(checks,
			"chart_type_not_empty",
			"main_trend_not_empty",
			"series_labels_only_include_visible_labels",
		)
	default:
		checks = append(checks,
			"layout_type_not_empty",
			"key_takeaway_grounded_in_visible_content",
		)
	}
	return dedupePDFVisualStrings(checks)
}

func buildPDFVisualNormalizationRules(profile pdfVisualSignalProfile) []string {
	rules := []string{
		"trim_whitespace_on_all_string_fields",
		"dedupe_string_list_values_case_insensitively",
		"preserve_visible_order_for_list_fields_when_possible",
	}
	switch profile.SummaryMode {
	case "slide_chart_summary":
		rules = append(rules,
			"normalize_chart_type_to_lower_snake_case_when_possible",
			"keep_slide_title_in_original_case",
		)
	case "table_summary":
		rules = append(rules,
			"normalize_headers_using_visible_text_not_inferred_labels",
			"keep_totals_and_comparisons_as_displayed_strings",
		)
	case "diagram_summary":
		rules = append(rules,
			"normalize_component_names_using_visible_labels",
			"use_simple_subject_verb_object_phrasing_for_relationships",
		)
	case "ocr_layout_summary":
		rules = append(rules,
			"preserve_ocr_uncertainty_in_scan_quality_notes",
			"avoid_normalizing_unclear_characters_into_confident_values",
		)
	case "chart_summary":
		rules = append(rules,
			"normalize_axes_or_categories_as_visible_labels",
			"prefer_compact_trend_statements_over_long prose",
		)
	default:
		rules = append(rules,
			"prefer_compact_layout_labels",
			"avoid_inventing_regions_not_visibly_present",
		)
	}
	return dedupePDFVisualStrings(rules)
}

func buildPDFVisualExtractionInstructions(profile pdfVisualSignalProfile) []string {
	instructions := []string{
		"Use only visually present evidence from the rendered pages and extracted text excerpts.",
		"Fill the result template field by field and leave optional fields empty when evidence is absent.",
	}
	switch profile.SummaryMode {
	case "slide_chart_summary":
		instructions = append(instructions,
			"Capture the visible slide headline first, then identify the dominant chart type and the most important trend or comparison.",
			"Use supporting_annotation only for visible captions, callouts, or side notes; do not infer presenter intent.",
		)
	case "table_summary":
		instructions = append(instructions,
			"Name the table subject from visible title or nearby heading text.",
			"Populate key_headers and key_rows using visible labels only, and keep totals_or_comparisons as displayed strings.",
		)
	case "diagram_summary":
		instructions = append(instructions,
			"Identify the main diagram subject, then list visible components and the directional or logical relationships between them.",
			"Describe relationships from arrows, connectors, or proximity; avoid inventing hidden flow semantics.",
		)
	case "ocr_layout_summary":
		instructions = append(instructions,
			"Prefer OCR-recoverable headings, section titles, captions, and callouts that are actually legible on the page.",
			"Record uncertainty and scan artifacts in scan_quality_notes instead of forcing a confident extraction.",
		)
	case "chart_summary":
		instructions = append(instructions,
			"Name the chart type, then summarize the main numeric trend and any visible axes, categories, or series labels.",
			"Keep trend descriptions compact and do not invent numeric values that are not clearly visible.",
		)
	default:
		instructions = append(instructions,
			"Describe the dominant layout regions and provide a single grounded key_takeaway from visible content.",
		)
	}
	return dedupePDFVisualStrings(instructions)
}

func buildPDFVisualAggregationStrategy(profile pdfVisualSignalProfile) string {
	switch profile.SummaryMode {
	case "slide_chart_summary":
		return "slide_first_then_chart_support"
	case "table_summary":
		return "header_once_merge_rows"
	case "diagram_summary":
		return "merge_components_then_relationships"
	case "ocr_layout_summary":
		return "page_order_text_merge"
	case "chart_summary":
		return "chart_first_then_labels"
	default:
		return "page_order_layout_merge"
	}
}

func buildPDFVisualAggregationRules(profile pdfVisualSignalProfile) []string {
	rules := []string{
		"process_pages_in_page_order",
		"prefer_consistent_field_values_over_page_local_variants",
		"dedupe_repeated_observations_across_pages",
	}
	switch profile.SummaryMode {
	case "slide_chart_summary":
		rules = append(rules,
			"keep_the_first_clear_slide_title_as_primary",
			"merge_chart_observations_into_one_main_trend_and_one_supporting_annotation",
		)
	case "table_summary":
		rules = append(rules,
			"capture_headers_once_then_append_unique_rows_or_totals",
			"prefer_visually emphasized totals when duplicate values appear",
		)
	case "diagram_summary":
		rules = append(rules,
			"merge_duplicate_nodes_by_visible_label",
			"prefer_directional_relationships_when_arrows_or_connectors_are_clear",
		)
	case "ocr_layout_summary":
		rules = append(rules,
			"preserve_page_order_for_headings_and_sections",
			"keep_uncertain_or_low_confidence_text_in_scan_quality_notes",
		)
	case "chart_summary":
		rules = append(rules,
			"merge_axes_categories_and_series_labels_without inventing missing labels",
			"prefer a single dominant trend statement over multiple overlapping trend claims",
		)
	default:
		rules = append(rules,
			"merge_layout_regions_from highest priority pages first",
			"prefer document-level takeaway over page-local repetition",
		)
	}
	return dedupePDFVisualStrings(rules)
}

func buildPDFVisualFocusAreas(profile pdfVisualSignalProfile) []string {
	areas := make([]string, 0, 5)
	areas = append(areas, "page_layout_and_section_boundaries")
	switch profile.PrimaryVisualTarget {
	case "chart":
		areas = append(areas, "chart_title_axes_series_and_outliers")
	case "table":
		areas = append(areas, "table_structure_headers_cells_and_footnotes")
	case "diagram":
		areas = append(areas, "diagram_nodes_edges_and_directionality")
	case "ocr_text":
		areas = append(areas, "ocr_text_quality_titles_and_annotations")
	}
	if profile.ChartLike {
		areas = append(areas, "chart_values_legends_and_trends")
	}
	if profile.TableLike {
		areas = append(areas, "table_headers_key_rows_and_columns")
	}
	if profile.DiagramLike {
		areas = append(areas, "diagram_structure_connectors_and_labels")
	}
	if profile.TextSparse || profile.ImageDocument {
		areas = append(areas, "ocr_text_regions_titles_and_captions")
	}
	if profile.LayoutType == "slide_deck" {
		areas = append(areas, "titles_bullets_and_slide_takeaways")
	}
	return dedupePDFVisualStrings(areas)
}

func buildPDFVisualFollowUps(profile pdfVisualSignalProfile) []string {
	steps := make([]string, 0, 4)
	switch profile.SummaryMode {
	case "slide_chart_summary":
		steps = append(steps, "Summarize the slide headline, the main chart takeaway, and supporting annotation text.")
	case "table_summary":
		steps = append(steps, "Extract the most important rows, columns, totals, and comparison points from the table.")
	case "diagram_summary":
		steps = append(steps, "Describe the major components, directional flow, and labeled relationships in the diagram.")
	case "ocr_layout_summary":
		steps = append(steps, "Run OCR-oriented summarization and extract the dominant headings, captions, and callouts.")
	case "chart_summary":
		steps = append(steps, "Extract chart type, x/y axes, series labels, and the most important numeric trend.")
	}
	if profile.ChartLike {
		steps = append(steps, "Extract chart trends, series labels, and notable numeric comparisons.")
	}
	if profile.TableLike {
		steps = append(steps, "Follow up with row/column level extraction for key tables.")
	}
	if profile.DiagramLike {
		steps = append(steps, "Summarize nodes, flows, and labeled relationships in the diagram.")
	}
	if profile.TextSparse || profile.ImageDocument {
		steps = append(steps, "Use OCR-aware summarization for rendered pages with limited embedded text.")
	}
	if profile.LayoutType == "slide_deck" {
		steps = append(steps, "Summarize each slide's headline and the main supporting visual.")
	}
	return dedupePDFVisualStrings(steps)
}

func dedupePDFVisualStrings(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	out := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, trimmed)
	}
	return out
}

func selectPDFVisualPromptMode(mediaProfile pdfMediaProfile, plan pdfAnalysisPlan) string {
	switch {
	case mediaProfile.LikelySlideDeck:
		return "slide_visual_summary"
	case plan.Mode == "vision_ocr" || plan.NeedsOCR:
		return "ocr_layout_summary"
	case mediaProfile.LikelyGraphicDoc:
		return "graphic_layout_summary"
	default:
		return "layout_summary"
	}
}

func pdfVisualPromptInstructions(mode string) []string {
	switch mode {
	case "slide_visual_summary":
		return []string{
			"Prioritize the slide title, the dominant visual, supporting bullets, and the main takeaway a presenter would want to emphasize.",
			"If charts are present, name the chart type, trend direction, and the most important comparison.",
		}
	case "ocr_layout_summary":
		return []string{
			"Prioritize readable titles, section headers, captions, and any diagram labels that likely require OCR.",
			"Call out where image quality or scan artifacts may reduce confidence.",
		}
	case "graphic_layout_summary":
		return []string{
			"Prioritize charts, tables, and diagrams over prose; identify the dominant visual artifact on each page.",
			"Call out the best next extraction target: chart values, table rows, or diagram structure.",
		}
	default:
		return []string{
			"Balance layout overview with the most information-dense visual elements.",
		}
	}
}

func classifyPDFVisualSummaryMode(profile pdfVisualSignalProfile, plan pdfAnalysisPlan) (string, string) {
	switch {
	case profile.LayoutType == "slide_deck" && profile.ChartLike:
		return "slide_chart_summary", "chart"
	case profile.TableLike:
		return "table_summary", "table"
	case profile.DiagramLike:
		return "diagram_summary", "diagram"
	case profile.LayoutType == "slide_deck":
		return "slide_visual_summary", "layout"
	case profile.ChartLike:
		return "chart_summary", "chart"
	case profile.ImageDocument || profile.TextSparse || plan.Mode == "vision_ocr":
		return "ocr_layout_summary", "ocr_text"
	default:
		return "layout_summary", "layout"
	}
}

func containsAny(s string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(s, needle) {
			return true
		}
	}
	return false
}
