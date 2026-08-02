package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	types "github.com/wsnacj/agentx-go/components/llm"
	toolcontract "github.com/wsnacj/agentx-go/components/tool"
)

const (
	defaultPDFExtractMaxChars = 120_000
	hardPDFExtractMaxChars    = 500_000
	defaultPDFPageMaxChars    = 24_000
	hardPDFPageMaxChars       = 120_000
	defaultPDFMaxPages        = 80
	hardPDFMaxPages           = 500
)

type PDFToolOptions struct {
	Root                      string
	EnabledTools              []string
	MaxExtractChars           int
	MaxPageChars              int
	MaxPages                  int
	TimeoutMs                 int
	MaxResponseBytes          int
	Backend                   PDFBackend
	FallbackBackend           PDFBackend
	OCRProfile                string
	PreferredModel            string
	FallbackModels            []string
	PreferNative              *bool
	NativePageSelectionPolicy string
	ArtifactCacheSize         int
	Models                    []PDFModelCandidate
	Host                      PDFHost
}

type pdfPageItem struct {
	Page      int    `json:"page"`
	Text      string `json:"text"`
	Chars     int    `json:"chars"`
	Truncated bool   `json:"truncated"`
}

type pdfOutlineNode struct {
	Title    string           `json:"title"`
	Children []pdfOutlineNode `json:"children,omitempty"`
}

type pdfAnalyzePageItem struct {
	Page    int    `json:"page"`
	Chars   int    `json:"chars"`
	Empty   bool   `json:"empty,omitempty"`
	Excerpt string `json:"excerpt,omitempty"`
}

type pdfAnalyzeHit struct {
	Page    int    `json:"page"`
	Matches int    `json:"matches"`
	Score   int    `json:"score"`
	Excerpt string `json:"excerpt,omitempty"`
}

type pdfBackendStatus struct {
	PrimaryBackend    string                   `json:"primary_backend,omitempty"`
	ExtractBackend    string                   `json:"extract_backend,omitempty"`
	LayoutBackend     string                   `json:"layout_backend,omitempty"`
	LayoutPreserved   bool                     `json:"layout_preserved,omitempty"`
	MetadataBackend   string                   `json:"metadata_backend,omitempty"`
	FallbackBackend   string                   `json:"fallback_backend,omitempty"`
	FallbackUsed      bool                     `json:"fallback_used,omitempty"`
	Degraded          bool                     `json:"degraded,omitempty"`
	Warning           string                   `json:"warning,omitempty"`
	AvailableBackends []PDFBackendAvailability `json:"available_backends,omitempty"`
}

type pdfDocumentProfile struct {
	PagesWithoutText    int     `json:"pages_without_text,omitempty"`
	EmptyPageRatio      float64 `json:"empty_page_ratio,omitempty"`
	AvgCharsPerTextPage int     `json:"avg_chars_per_text_page,omitempty"`
	MaxPageChars        int     `json:"max_page_chars,omitempty"`
	LikelyScanned       bool    `json:"likely_scanned,omitempty"`
	HasOutline          bool    `json:"has_outline,omitempty"`
}

type pdfMediaProfile struct {
	TotalRects        int   `json:"total_rects,omitempty"`
	PagesWithGraphics int   `json:"pages_with_graphics,omitempty"`
	MaxRectsPerPage   int   `json:"max_rects_per_page,omitempty"`
	GraphicHeavyPages []int `json:"graphic_heavy_pages,omitempty"`
	LikelyGraphicDoc  bool  `json:"likely_graphic_doc,omitempty"`
	LikelySlideDeck   bool  `json:"likely_slide_deck,omitempty"`
}

type pdfAnalysisPlan struct {
	Mode                  string                    `json:"mode,omitempty"`
	NeedsVision           bool                      `json:"needs_vision,omitempty"`
	NeedsOCR              bool                      `json:"needs_ocr,omitempty"`
	PreferredBackend      string                    `json:"preferred_backend,omitempty"`
	PreferredClients      []string                  `json:"preferred_clients,omitempty"`
	ProviderRouting       string                    `json:"provider_routing,omitempty"`
	NativeProviderRouting string                    `json:"native_provider_routing,omitempty"`
	Reason                string                    `json:"reason,omitempty"`
	Warning               string                    `json:"warning,omitempty"`
	CandidateModels       []pdfVisionModelCandidate `json:"candidate_models,omitempty"`
	SuggestedNextSteps    []string                  `json:"suggested_next_steps,omitempty"`
}

type pdfToolSurface struct {
	Unified           bool
	Extract           bool
	ReadPages         bool
	Outline           bool
	Analyze           bool
	ExtractStructured bool
}

type pdfBackendRuntime struct {
	root          string
	primary       PDFBackend
	fallback      PDFBackend
	primaryAvail  PDFBackendAvailability
	fallbackAvail PDFBackendAvailability
	hasFallback   bool
}

func newPDFBackendRuntime(opts PDFToolOptions) pdfBackendRuntime {
	primary := opts.Backend
	fallback := opts.FallbackBackend
	rt := pdfBackendRuntime{
		root:    strings.TrimSpace(opts.Root),
		primary: primary,
	}
	if primary != nil {
		rt.primaryAvail = primary.Availability(context.Background())
	}
	if fallback != nil && (primary == nil || strings.TrimSpace(fallback.Name()) != strings.TrimSpace(primary.Name())) {
		rt.fallback = fallback
		rt.fallbackAvail = fallback.Availability(context.Background())
		rt.hasFallback = true
	}
	return rt
}

func (rt pdfBackendRuntime) registrationAvailable() bool {
	if rt.primaryAvail.Available {
		return true
	}
	return rt.hasFallback && rt.fallbackAvail.Available
}

func (rt pdfBackendRuntime) availableBackends() []PDFBackendAvailability {
	out := make([]PDFBackendAvailability, 0, 2)
	if strings.TrimSpace(rt.primaryAvail.Name) != "" {
		out = append(out, rt.primaryAvail)
	}
	if rt.hasFallback && strings.TrimSpace(rt.fallbackAvail.Name) != "" {
		out = append(out, rt.fallbackAvail)
	}
	return out
}

func (rt pdfBackendRuntime) runText(ctx context.Context, op func(PDFBackend) (PDFTextResult, error)) (PDFTextResult, pdfBackendStatus, error) {
	status := pdfBackendStatus{
		PrimaryBackend:    backendName(rt.primary),
		FallbackBackend:   backendName(rt.fallback),
		AvailableBackends: rt.availableBackends(),
	}
	if rt.primaryAvail.Available {
		result, err := op(rt.primary)
		if err == nil {
			status.ExtractBackend = backendName(rt.primary)
			return result, status, nil
		}
		if !rt.hasFallback || !rt.fallbackAvail.Available {
			return PDFTextResult{}, status, err
		}
		status.Degraded = true
		status.Warning = fmt.Sprintf("primary backend %s failed; used fallback %s", backendName(rt.primary), backendName(rt.fallback))
		result, fallbackErr := op(rt.fallback)
		if fallbackErr != nil {
			return PDFTextResult{}, status, fmt.Errorf("%w; fallback %s failed: %v", err, backendName(rt.fallback), fallbackErr)
		}
		status.ExtractBackend = backendName(rt.fallback)
		status.FallbackUsed = true
		return result, status, nil
	}
	if rt.hasFallback && rt.fallbackAvail.Available {
		status.Degraded = true
		if reason := strings.TrimSpace(rt.primaryAvail.Reason); reason != "" {
			status.Warning = fmt.Sprintf("primary backend %s unavailable (%s); used fallback %s", backendName(rt.primary), reason, backendName(rt.fallback))
		} else {
			status.Warning = fmt.Sprintf("primary backend %s unavailable; used fallback %s", backendName(rt.primary), backendName(rt.fallback))
		}
		result, err := op(rt.fallback)
		if err != nil {
			return PDFTextResult{}, status, err
		}
		status.ExtractBackend = backendName(rt.fallback)
		status.FallbackUsed = true
		return result, status, nil
	}
	if reason := strings.TrimSpace(rt.primaryAvail.Reason); reason != "" {
		return PDFTextResult{}, status, fmt.Errorf("%s unavailable: %s", backendName(rt.primary), reason)
	}
	return PDFTextResult{}, status, fmt.Errorf("%s unavailable", backendName(rt.primary))
}

func (rt pdfBackendRuntime) runLayoutText(ctx context.Context, path string, pages []int) (PDFTextResult, string, bool, error) {
	type candidate struct {
		backend      PDFBackend
		availability PDFBackendAvailability
	}
	candidates := []candidate{{backend: rt.primary, availability: rt.primaryAvail}}
	if rt.hasFallback {
		candidates = append(candidates, candidate{backend: rt.fallback, availability: rt.fallbackAvail})
	}
	var firstErr error
	capabilityAvailable := false
	for _, item := range candidates {
		if item.backend == nil || !item.availability.Available {
			continue
		}
		layoutBackend := asPDFLayoutBackend(item.backend)
		if layoutBackend == nil {
			continue
		}
		capabilityAvailable = true
		result, err := layoutBackend.extractLayoutText(ctx, path, pages)
		if err == nil {
			return result, backendName(item.backend), true, nil
		}
		if firstErr == nil {
			firstErr = fmt.Errorf("layout backend %s failed: %w", backendName(item.backend), err)
		}
	}
	return PDFTextResult{}, "", capabilityAvailable, firstErr
}

func (rt pdfBackendRuntime) runMetadata(ctx context.Context, includeFonts bool, op func(PDFBackend, bool) (PDFMetadataResult, error)) (PDFMetadataResult, pdfBackendStatus, error) {
	status := pdfBackendStatus{
		PrimaryBackend:    backendName(rt.primary),
		FallbackBackend:   backendName(rt.fallback),
		AvailableBackends: rt.availableBackends(),
	}
	if rt.primaryAvail.Available {
		result, err := op(rt.primary, includeFonts)
		if err == nil {
			status.MetadataBackend = backendName(rt.primary)
			return result, status, nil
		}
		if !rt.hasFallback || !rt.fallbackAvail.Available {
			return PDFMetadataResult{}, status, err
		}
		status.Degraded = true
		status.Warning = fmt.Sprintf("primary backend %s metadata failed; used fallback %s", backendName(rt.primary), backendName(rt.fallback))
		result, fallbackErr := op(rt.fallback, includeFonts)
		if fallbackErr != nil {
			return PDFMetadataResult{}, status, fmt.Errorf("%w; fallback %s metadata failed: %v", err, backendName(rt.fallback), fallbackErr)
		}
		status.MetadataBackend = backendName(rt.fallback)
		status.FallbackUsed = true
		return result, status, nil
	}
	if rt.hasFallback && rt.fallbackAvail.Available {
		status.Degraded = true
		if reason := strings.TrimSpace(rt.primaryAvail.Reason); reason != "" {
			status.Warning = fmt.Sprintf("primary backend %s unavailable (%s); used fallback %s metadata", backendName(rt.primary), reason, backendName(rt.fallback))
		} else {
			status.Warning = fmt.Sprintf("primary backend %s unavailable; used fallback %s metadata", backendName(rt.primary), backendName(rt.fallback))
		}
		result, err := op(rt.fallback, includeFonts)
		if err != nil {
			return PDFMetadataResult{}, status, err
		}
		status.MetadataBackend = backendName(rt.fallback)
		status.FallbackUsed = true
		return result, status, nil
	}
	if reason := strings.TrimSpace(rt.primaryAvail.Reason); reason != "" {
		return PDFMetadataResult{}, status, fmt.Errorf("%s unavailable: %s", backendName(rt.primary), reason)
	}
	return PDFMetadataResult{}, status, fmt.Errorf("%s unavailable", backendName(rt.primary))
}

func mergePDFBackendStatus(base pdfBackendStatus, extra pdfBackendStatus) pdfBackendStatus {
	if strings.TrimSpace(base.PrimaryBackend) == "" {
		base.PrimaryBackend = extra.PrimaryBackend
	}
	if strings.TrimSpace(base.ExtractBackend) == "" {
		base.ExtractBackend = extra.ExtractBackend
	}
	if strings.TrimSpace(base.LayoutBackend) == "" {
		base.LayoutBackend = extra.LayoutBackend
	}
	if extra.LayoutPreserved {
		base.LayoutPreserved = true
	}
	if strings.TrimSpace(extra.MetadataBackend) != "" {
		base.MetadataBackend = extra.MetadataBackend
	}
	if strings.TrimSpace(base.FallbackBackend) == "" {
		base.FallbackBackend = extra.FallbackBackend
	}
	if extra.FallbackUsed {
		base.FallbackUsed = true
	}
	if extra.Degraded {
		base.Degraded = true
	}
	if strings.TrimSpace(base.Warning) == "" {
		base.Warning = extra.Warning
	} else if strings.TrimSpace(extra.Warning) != "" && !strings.Contains(base.Warning, extra.Warning) {
		base.Warning += "; " + extra.Warning
	}
	if len(base.AvailableBackends) == 0 {
		base.AvailableBackends = extra.AvailableBackends
	}
	return base
}

func appendPDFBackendWarning(status pdfBackendStatus, warning string) pdfBackendStatus {
	warning = strings.TrimSpace(warning)
	if warning == "" {
		return status
	}
	if strings.TrimSpace(status.Warning) == "" {
		status.Warning = warning
		return status
	}
	if !strings.Contains(status.Warning, warning) {
		status.Warning += "; " + warning
	}
	return status
}

func backendName(backend PDFBackend) string {
	if backend == nil {
		return ""
	}
	return strings.TrimSpace(backend.Name())
}

func buildPDFDocumentProfile(pageCount int, hasOutline bool, pageMap []pdfAnalyzePageItem) pdfDocumentProfile {
	if pageCount <= 0 {
		return pdfDocumentProfile{HasOutline: hasOutline}
	}
	withoutText := 0
	totalChars := 0
	maxChars := 0
	textPages := 0
	for _, page := range pageMap {
		if page.Empty || page.Chars == 0 {
			withoutText++
			continue
		}
		textPages++
		totalChars += page.Chars
		if page.Chars > maxChars {
			maxChars = page.Chars
		}
	}
	avgChars := 0
	if textPages > 0 {
		avgChars = totalChars / textPages
	}
	emptyRatio := float64(withoutText) / float64(pageCount)
	return pdfDocumentProfile{
		PagesWithoutText:    withoutText,
		EmptyPageRatio:      emptyRatio,
		AvgCharsPerTextPage: avgChars,
		MaxPageChars:        maxChars,
		LikelyScanned:       textPages == 0 || (pageCount >= 3 && emptyRatio >= 0.8),
		HasOutline:          hasOutline,
	}
}

func buildPDFMediaProfile(metadata PDFMetadataResult, document pdfDocumentProfile, pageMap []pdfAnalyzePageItem) pdfMediaProfile {
	profile := pdfMediaProfile{
		TotalRects:        metadata.TotalRects,
		PagesWithGraphics: metadata.PagesWithGraphics,
		MaxRectsPerPage:   metadata.MaxRectsPerPage,
	}
	if metadata.PageCount <= 0 {
		return profile
	}
	if metadata.PagesWithGraphics == 0 && metadata.TotalRects == 0 {
		return profile
	}
	heavy := make([]int, 0)
	pageChars := make(map[int]int, len(pageMap))
	for _, item := range pageMap {
		pageChars[item.Page] = item.Chars
	}
	for _, page := range metadata.GraphicPages {
		rects := metadata.RectsByPage[page]
		if rects >= 12 || (rects >= 4 && pageChars[page] <= 120) {
			heavy = append(heavy, page)
		}
	}
	profile.GraphicHeavyPages = heavy
	graphicRatio := float64(metadata.PagesWithGraphics) / float64(metadata.PageCount)
	profile.LikelyGraphicDoc = metadata.TotalRects >= metadata.PageCount*4 || graphicRatio >= 0.6
	profile.LikelySlideDeck = metadata.PageCount >= 5 && graphicRatio >= 0.5 && document.AvgCharsPerTextPage > 0 && document.AvgCharsPerTextPage <= 220
	return profile
}

func maybeSupplementPDFToolTextWithOCR(
	ctx context.Context,
	path string,
	selectedPages []int,
	textResult PDFTextResult,
	metadata PDFMetadataResult,
	backendStatus pdfBackendStatus,
	maxPageChars int,
	ocrxConfigPath string,
) (PDFTextResult, pdfBackendStatus, bool) {
	pageMap := buildPDFAnalyzePageMap(textResult.Pages, maxPageChars)
	pageCountForProfile := metadata.PageCount
	if pageCountForProfile <= 0 {
		if len(selectedPages) > 0 {
			pageCountForProfile = len(selectedPages)
		} else {
			pageCountForProfile = len(textResult.Pages)
		}
	}
	if len(selectedPages) > 0 && len(selectedPages) < pageCountForProfile {
		pageCountForProfile = len(selectedPages)
	}
	documentProfile := buildPDFDocumentProfile(pageCountForProfile, metadata.Outline != nil, pageMap)
	mediaProfile := buildPDFMediaProfile(metadata, documentProfile, pageMap)
	analysisPlan := buildPDFAnalysisPlan(documentProfile, mediaProfile, backendStatus)
	return maybeSupplementPDFUnifiedTextWithOCR(ctx, path, selectedPages, textResult, backendStatus, documentProfile, analysisPlan, ocrxConfigPath)
}

func preferredPDFVisionClients(mode string) []string {
	switch strings.TrimSpace(mode) {
	case "vision_ocr":
		return []string{"gemini", "anthropic", "openai", "ark", "openrouter"}
	case "hybrid_vision_text":
		return []string{"anthropic", "gemini", "openai", "ark", "openrouter"}
	default:
		return []string{"anthropic", "gemini", "openai", "ark", "openrouter"}
	}
}

func buildPDFProviderRouting(mode string, preferred []string, candidates []pdfVisionModelCandidate) string {
	if len(candidates) == 0 {
		return ""
	}
	parts := make([]string, 0, len(candidates))
	seen := make(map[string]struct{}, len(candidates))
	for _, candidate := range candidates {
		client := pdfVisionCandidateRouteLabel(candidate)
		if client == "" {
			continue
		}
		key := strings.ToLower(client)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		parts = append(parts, client)
	}
	if len(parts) == 0 {
		return ""
	}
	if len(preferred) == 0 {
		return fmt.Sprintf("vision mode %s will try providers in available order: %s", mode, strings.Join(parts, " -> "))
	}
	return fmt.Sprintf("vision mode %s prefers providers %s; available order is %s", mode, strings.Join(preferred, " -> "), strings.Join(parts, " -> "))
}

func pdfVisionCandidateRoutingKey(candidate pdfVisionModelCandidate) string {
	if trimmed := strings.ToLower(strings.TrimSpace(candidate.Vendor)); trimmed != "" {
		return trimmed
	}
	return strings.ToLower(strings.TrimSpace(candidate.Client))
}

func pdfVisionCandidateRouteLabel(candidate pdfVisionModelCandidate) string {
	if trimmed := strings.TrimSpace(candidate.Vendor); trimmed != "" {
		return trimmed
	}
	return strings.TrimSpace(candidate.Client)
}

func buildPDFNativeProviderRouting(mode string, candidates []pdfVisionModelCandidate) string {
	native := filterPDFNativeCandidates(candidates)
	if len(native) == 0 {
		return ""
	}
	parts := make([]string, 0, len(native))
	seen := make(map[string]struct{}, len(native))
	for _, candidate := range native {
		vendor := strings.TrimSpace(candidate.Vendor)
		if vendor == "" {
			continue
		}
		key := strings.ToLower(vendor)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		parts = append(parts, vendor)
	}
	if len(parts) == 0 {
		return ""
	}
	return fmt.Sprintf("vision mode %s has native PDF-capable providers available: %s", mode, strings.Join(parts, " -> "))
}

func filterPDFNativeCandidates(candidates []pdfVisionModelCandidate) []pdfVisionModelCandidate {
	out := make([]pdfVisionModelCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate.NativePDF {
			out = append(out, candidate)
		}
	}
	return out
}

func buildPDFAnalysisPlan(document pdfDocumentProfile, media pdfMediaProfile, backendStatus pdfBackendStatus) pdfAnalysisPlan {
	plan := pdfAnalysisPlan{
		PreferredBackend: strings.TrimSpace(backendStatus.ExtractBackend),
	}
	surface := defaultPDFToolSurface()
	switch {
	case document.LikelyScanned:
		plan.Mode = "vision_ocr"
		plan.NeedsVision = true
		plan.NeedsOCR = true
		plan.Reason = "document appears scanned or text-sparse; OCR/vision analysis is recommended"
	case media.LikelySlideDeck || media.LikelyGraphicDoc:
		plan.Mode = "hybrid_vision_text"
		plan.NeedsVision = true
		plan.Reason = "document looks graphics-heavy; combine text extraction with visual page analysis"
	default:
		plan.Mode = "text_first"
		plan.Reason = "document is text-friendly; text extraction should be sufficient for first-pass analysis"
	}
	plan.SuggestedNextSteps = pdfAnalysisPlanSuggestedNextSteps(plan.Mode, surface)
	return plan
}

func defaultPDFToolSurface() pdfToolSurface {
	return pdfToolSurface{
		Unified:           true,
		Extract:           true,
		ReadPages:         true,
		Outline:           true,
		Analyze:           true,
		ExtractStructured: true,
	}
}

func buildPDFToolSurface(enabled map[string]bool, unifiedAvailable bool) pdfToolSurface {
	return pdfToolSurface{
		Unified:           unifiedAvailable && toolEnabled(enabled, "pdf"),
		Extract:           toolEnabled(enabled, "pdf_extract"),
		ReadPages:         toolEnabled(enabled, "pdf_read_pages"),
		Outline:           toolEnabled(enabled, "pdf_outline"),
		Analyze:           toolEnabled(enabled, "pdf_analyze"),
		ExtractStructured: toolEnabled(enabled, "pdf_extract_structured"),
	}
}

func pdfAnalysisPlanSuggestedNextSteps(mode string, surface pdfToolSurface) []string {
	readPagesGuidance := "Use pdf_read_pages for exact page-level excerpts that already parse cleanly."
	structuredGuidance := "Escalate to pdf_extract_structured when layout-heavy fields, charts, tables, or OCR-visible structure need structured extraction."
	switch strings.TrimSpace(mode) {
	case "vision_ocr":
		switch {
		case surface.Unified:
			return []string{
				"Use unified pdf (`pdf`) for OCR-aware question answering or summarization over the relevant pages.",
				"Use pdf_read_pages only for exact page-level snippets that already parse cleanly, and escalate to pdf_extract_structured when layout-heavy fields need structured extraction.",
			}
		case surface.ExtractStructured && surface.ReadPages:
			return []string{
				"Use pdf_extract_structured for OCR-aware extraction and layout-heavy field recovery across the relevant pages.",
				readPagesGuidance,
			}
		case surface.ExtractStructured:
			return []string{
				"Use pdf_extract_structured for OCR-aware extraction and layout-heavy field recovery across the relevant pages.",
			}
		case surface.ReadPages:
			return []string{readPagesGuidance}
		case surface.Extract:
			return []string{"Use pdf_extract for full-text OCR output when specialist layout tools are unavailable."}
		default:
			return nil
		}
	case "hybrid_vision_text":
		switch {
		case surface.Unified:
			return []string{
				"Use unified pdf (`pdf`) for first-pass chart, layout, or slide interpretation across text and rendered pages.",
				"Escalate to pdf_extract_structured for chart/table/layout-heavy structured output, and keep pdf_read_pages for exact supporting excerpts.",
			}
		case surface.ExtractStructured && surface.ReadPages:
			return []string{
				"Use pdf_extract_structured for first-pass chart, layout, or slide interpretation across extracted text and rendered pages.",
				readPagesGuidance,
			}
		case surface.ExtractStructured:
			return []string{
				"Use pdf_extract_structured for first-pass chart, layout, or slide interpretation across extracted text and rendered pages.",
			}
		case surface.ReadPages:
			return []string{
				"Use pdf_read_pages for exact page-level excerpts from the likely chart or layout-heavy pages.",
			}
		case surface.Extract:
			return []string{
				"Use pdf_extract for plain-text fallback, but expect chart and layout fidelity to be limited.",
			}
		default:
			return nil
		}
	default:
		switch {
		case surface.Unified:
			return []string{
				"Use unified pdf (`pdf`) for first-pass summarization or Q&A, and use pdf_read_pages only when you need exact page-level excerpts.",
			}
		case surface.ReadPages && surface.ExtractStructured:
			return []string{
				readPagesGuidance + " " + structuredGuidance,
			}
		case surface.ReadPages:
			return []string{readPagesGuidance}
		case surface.ExtractStructured:
			return []string{structuredGuidance}
		case surface.Extract:
			return []string{"Use pdf_extract for full-text extraction when the unified `pdf` entrypoint is unavailable."}
		default:
			return nil
		}
	}
}

func forcePDFVisualAnalysisPlan(plan pdfAnalysisPlan) pdfAnalysisPlan {
	if plan.NeedsVision {
		return plan
	}
	reason := strings.TrimSpace(plan.Reason)
	plan.Mode = "hybrid_vision_text"
	plan.NeedsVision = true
	switch {
	case reason == "":
		plan.Reason = "visual analysis explicitly requested by the tool caller"
	case strings.Contains(strings.ToLower(reason), "visual analysis explicitly requested"):
		plan.Reason = reason
	default:
		plan.Reason = reason + "; visual analysis explicitly requested by the tool caller"
	}
	return plan
}

func adaptPDFAnalysisPlanForToolSurface(plan pdfAnalysisPlan, resolver pdfModelResolverConfig, surface pdfToolSurface, forceVisualAnalysis bool) pdfAnalysisPlan {
	if forceVisualAnalysis {
		plan = forcePDFVisualAnalysisPlan(plan)
	}
	plan = applyPDFModelResolverToPlan(plan, resolver)
	plan.SuggestedNextSteps = pdfAnalysisPlanSuggestedNextSteps(plan.Mode, surface)
	return plan
}

func pdfSpecialistToolDescription(task string, unifiedAvailable bool) string {
	trimmedTask := strings.TrimSpace(task)
	citationNote := " Preserve page-grounded evidence or inline citations from this tool in downstream answers when available."
	if !unifiedAvailable {
		return fmt.Sprintf("Specialist PDF tool for %s over a workspace PDF file, file:// URL, or http(s) PDF URL. Use this surface directly when the unified `pdf` entrypoint is unavailable.%s", trimmedTask, citationNote)
	}
	return fmt.Sprintf("Specialist companion to the unified `pdf` entrypoint for %s over a workspace PDF file, file:// URL, or http(s) PDF URL. Prefer `pdf` for default question-answering, summarization, and comparison work.%s", trimmedTask, citationNote)
}

func RegisterPDFTools(reg toolcontract.Registrar, opts PDFToolOptions) error {
	if reg == nil {
		return fmt.Errorf("pdf tool registrar is required")
	}
	if err := opts.Host.validate(); err != nil {
		return err
	}
	runtime := newPDFBackendRuntime(opts)
	if !runtime.registrationAvailable() {
		return nil
	}
	maxExtractChars := clampToolLimit(opts.MaxExtractChars, defaultPDFExtractMaxChars, hardPDFExtractMaxChars)
	maxPageChars := clampToolLimit(opts.MaxPageChars, defaultPDFPageMaxChars, hardPDFPageMaxChars)
	maxPages := clampToolLimit(opts.MaxPages, defaultPDFMaxPages, hardPDFMaxPages)
	defaultRemoteMaxBytes := normalizePDFToolMaxResponseBytes(opts.MaxResponseBytes)
	defaultRemoteTimeoutMs := normalizePDFToolTimeoutMs(opts.TimeoutMs)
	resolverConfig := newPDFModelResolverConfig(opts)
	enabled := buildEnabledToolSet(opts.EnabledTools)
	unifiedAvailable := pdfUnifiedModelAvailable(resolverConfig)
	unifiedRegistered := (unifiedAvailable || pdfConfiguredModelRoutePresent(resolverConfig)) && toolEnabled(enabled, "pdf")
	artifactCache := newPDFArtifactCache(opts.ArtifactCacheSize)
	unifiedCache := newPDFUnifiedArtifactCache(opts.ArtifactCacheSize)
	unifiedVisualCache := newPDFUnifiedVisualCache(opts.ArtifactCacheSize)
	surface := buildPDFToolSurface(enabled, unifiedRegistered)

	if unifiedRegistered {
		reg.Register(pdfUnifiedDefinition(), func(ctx context.Context, call types.FunctionCall) (string, error) {
			ctx = contextWithPDFHost(ctx, opts.Host)
			params, err := decodeArgs(call.Arguments)
			if err != nil {
				return "", err
			}
			return runPDFUnifiedTool(ctx, params, pdfUnifiedToolOptions{
				Root:                      opts.Root,
				Runtime:                   runtime,
				DefaultTimeoutMs:          defaultRemoteTimeoutMs,
				DefaultMaxBytes:           defaultRemoteMaxBytes,
				MaxPages:                  maxPages,
				MaxPageChars:              maxPageChars,
				OCRXConfigPath:            opts.OCRProfile,
				Resolver:                  resolverConfig,
				NativePageSelectionPolicy: opts.NativePageSelectionPolicy,
				Cache:                     unifiedCache,
				VisualCache:               unifiedVisualCache,
			})
		})
	}

	if toolEnabled(enabled, "pdf_extract") {
		reg.Register(pdfExtractDefinition(surface.Unified), func(ctx context.Context, call types.FunctionCall) (string, error) {
			ctx = contextWithPDFHost(ctx, opts.Host)
			params, err := decodeArgs(call.Arguments)
			if err != nil {
				return "", err
			}
			input, err := resolvePDFToolInput(ctx, opts.Root, params, defaultRemoteTimeoutMs, defaultRemoteMaxBytes)
			if err != nil {
				return "", fmt.Errorf("pdf_extract: %w", err)
			}
			defer input.Cleanup()
			resolved, display := input.Path, input.Display
			metadata, metaStatus, err := runtime.runMetadata(ctx, false, func(backend PDFBackend, includeFonts bool) (PDFMetadataResult, error) {
				return backend.ReadMetadata(ctx, resolved, includeFonts)
			})
			if err != nil {
				return "", fmt.Errorf("pdf_extract: %w", err)
			}
			textResult, backendStatus, err := runtime.runText(ctx, func(backend PDFBackend) (PDFTextResult, error) {
				return backend.ExtractAllText(ctx, resolved)
			})
			if err != nil {
				return "", fmt.Errorf("pdf_extract: %w", err)
			}
			backendStatus = mergePDFBackendStatus(metaStatus, backendStatus)
			textResult, backendStatus, _ = maybeSupplementPDFToolTextWithOCR(ctx, resolved, nil, textResult, metadata, backendStatus, maxPageChars, opts.OCRProfile)
			pageCount := metadata.PageCount
			if pageCount <= 0 {
				pageCount = len(textResult.Pages)
			}
			limit := firstInt(params, "max_chars")
			if limit <= 0 || limit > maxExtractChars {
				limit = maxExtractChars
			}

			text := joinPDFPageTexts(textResult.Pages)
			truncated := false
			if trimmed, changed := trimToMaxChars(text, limit); changed {
				text = trimmed
				truncated = true
			}

			payload := struct {
				Path             string           `json:"path"`
				Backend          string           `json:"backend"`
				BackendStatus    pdfBackendStatus `json:"backend_status,omitempty"`
				PageCount        int              `json:"page_count"`
				Text             string           `json:"text"`
				TextChars        int              `json:"text_chars"`
				Truncated        bool             `json:"truncated"`
				Pages            []pdfPageItem    `json:"pages,omitempty"`
				PageLimitApplied bool             `json:"page_limit_applied,omitempty"`
				Outline          []pdfOutlineNode `json:"outline,omitempty"`
			}{
				Path:          display,
				Backend:       backendStatus.ExtractBackend,
				BackendStatus: backendStatus,
				PageCount:     pageCount,
				Text:          text,
				TextChars:     runeLen(text),
				Truncated:     truncated,
			}
			if readBool(params, "include_pages") {
				pageItems, limited := buildPDFPageItems(textResult.Pages, maxPageChars, maxPages, false)
				payload.Pages = pageItems
				payload.PageLimitApplied = limited
			}
			if readBool(params, "include_outline") {
				payload.Outline = convertPDFOutline(metadata.Outline)
			}

			blob, err := json.Marshal(payload)
			if err != nil {
				return "", err
			}
			return string(blob), nil
		})
	}

	if toolEnabled(enabled, "pdf_read_pages") {
		reg.Register(pdfReadPagesDefinition(surface.Unified), func(ctx context.Context, call types.FunctionCall) (string, error) {
			ctx = contextWithPDFHost(ctx, opts.Host)
			params, err := decodeArgs(call.Arguments)
			if err != nil {
				return "", err
			}
			input, err := resolvePDFToolInput(ctx, opts.Root, params, defaultRemoteTimeoutMs, defaultRemoteMaxBytes)
			if err != nil {
				return "", fmt.Errorf("pdf_read_pages: %w", err)
			}
			defer input.Cleanup()
			resolved, display := input.Path, input.Display
			metadata, backendStatus, err := runtime.runMetadata(ctx, false, func(backend PDFBackend, includeFonts bool) (PDFMetadataResult, error) {
				return backend.ReadMetadata(ctx, resolved, includeFonts)
			})
			if err != nil {
				return "", fmt.Errorf("pdf_read_pages: %w", err)
			}
			pageSelection, pageLimitApplied, err := resolvePDFPageSelection(params, metadata.PageCount, maxPages)
			if err != nil {
				return "", fmt.Errorf("pdf_read_pages: %w", err)
			}
			textResult, textStatus, err := runtime.runText(ctx, func(backend PDFBackend) (PDFTextResult, error) {
				return backend.ExtractPageText(ctx, resolved, pageSelection)
			})
			if err != nil {
				return "", fmt.Errorf("pdf_read_pages: %w", err)
			}
			backendStatus = mergePDFBackendStatus(backendStatus, textStatus)
			textResult, backendStatus, _ = maybeSupplementPDFToolTextWithOCR(ctx, resolved, pageSelection, textResult, metadata, backendStatus, maxPageChars, opts.OCRProfile)
			pageCharsLimit := firstInt(params, "max_chars_per_page", "max_page_chars")
			if pageCharsLimit <= 0 || pageCharsLimit > maxPageChars {
				pageCharsLimit = maxPageChars
			}
			includeEmpty := readBool(params, "include_empty")
			textByPage := make(map[int]string, len(textResult.Pages))
			for _, page := range textResult.Pages {
				textByPage[page.Page] = page.Text
			}
			pages := make([]pdfPageItem, 0, len(pageSelection))
			for _, pageNum := range pageSelection {
				raw := textByPage[pageNum]
				if !includeEmpty && strings.TrimSpace(raw) == "" {
					continue
				}
				pageText := raw
				pageTruncated := false
				if trimmed, changed := trimToMaxChars(pageText, pageCharsLimit); changed {
					pageText = trimmed
					pageTruncated = true
				}
				pages = append(pages, pdfPageItem{
					Page:      pageNum,
					Text:      pageText,
					Chars:     runeLen(pageText),
					Truncated: pageTruncated,
				})
			}

			payload := struct {
				Path             string           `json:"path"`
				Backend          string           `json:"backend"`
				BackendStatus    pdfBackendStatus `json:"backend_status,omitempty"`
				PageCount        int              `json:"page_count"`
				SelectedPages    []int            `json:"selected_pages"`
				PageLimitApplied bool             `json:"page_limit_applied,omitempty"`
				Pages            []pdfPageItem    `json:"pages"`
			}{
				Path:             display,
				Backend:          backendStatus.ExtractBackend,
				BackendStatus:    backendStatus,
				PageCount:        metadata.PageCount,
				SelectedPages:    pageSelection,
				PageLimitApplied: pageLimitApplied,
				Pages:            pages,
			}
			blob, err := json.Marshal(payload)
			if err != nil {
				return "", err
			}
			return string(blob), nil
		})
	}

	if toolEnabled(enabled, "pdf_outline") {
		reg.Register(pdfOutlineDefinition(surface.Unified), func(ctx context.Context, call types.FunctionCall) (string, error) {
			ctx = contextWithPDFHost(ctx, opts.Host)
			params, err := decodeArgs(call.Arguments)
			if err != nil {
				return "", err
			}
			input, err := resolvePDFToolInput(ctx, opts.Root, params, defaultRemoteTimeoutMs, defaultRemoteMaxBytes)
			if err != nil {
				return "", fmt.Errorf("pdf_outline: %w", err)
			}
			defer input.Cleanup()
			resolved, display := input.Path, input.Display
			includeFonts := readBool(params, "include_fonts")
			metadata, backendStatus, err := runtime.runMetadata(ctx, includeFonts, func(backend PDFBackend, includeFonts bool) (PDFMetadataResult, error) {
				return backend.ReadMetadata(ctx, resolved, includeFonts)
			})
			if err != nil {
				return "", fmt.Errorf("pdf_outline: %w", err)
			}
			outline := convertPDFOutline(metadata.Outline)
			payload := struct {
				Path          string           `json:"path"`
				Backend       string           `json:"backend"`
				BackendStatus pdfBackendStatus `json:"backend_status,omitempty"`
				PageCount     int              `json:"page_count"`
				HasOutline    bool             `json:"has_outline"`
				Outline       []pdfOutlineNode `json:"outline,omitempty"`
				Fonts         []PDFFont        `json:"fonts,omitempty"`
			}{
				Path:          display,
				Backend:       backendStatus.MetadataBackend,
				BackendStatus: backendStatus,
				PageCount:     metadata.PageCount,
				HasOutline:    len(outline) > 0,
				Outline:       outline,
			}
			if includeFonts {
				payload.Fonts = append([]PDFFont(nil), metadata.Fonts...)
			}
			blob, err := json.Marshal(payload)
			if err != nil {
				return "", err
			}
			return string(blob), nil
		})
	}

	if toolEnabled(enabled, "pdf_analyze") {
		reg.Register(pdfAnalyzeDefinition(surface.Unified), func(ctx context.Context, call types.FunctionCall) (string, error) {
			ctx = contextWithPDFHost(ctx, opts.Host)
			params, err := decodeArgs(call.Arguments)
			if err != nil {
				return "", err
			}
			inputs, err := resolvePDFToolInputs(ctx, opts.Root, params, defaultRemoteTimeoutMs, defaultRemoteMaxBytes, hardPDFToolInputCount)
			if err != nil {
				return "", fmt.Errorf("pdf_analyze: %w", err)
			}
			defer cleanupPDFToolInputs(inputs)
			query := strings.TrimSpace(firstString(params, "query"))
			maxHits := clampToolLimit(firstInt(params, "max_hits"), 5, 20)
			maxExcerptChars := clampToolLimit(firstInt(params, "max_excerpt_chars"), 280, 1_200)
			includeVisualAnalysis := readBool(params, "include_visual_analysis")
			visionModel := strings.TrimSpace(firstString(params, "vision_model"))
			maxVisualPages := clampToolLimit(firstInt(params, "max_visual_pages"), defaultPDFVisualPages, hardPDFVisualPages)
			includeOutline := readBool(params, "include_outline")
			includePageMap := readBool(params, "include_page_map")
			if len(inputs) > 1 {
				artifactsList := make([]pdfAnalysisArtifacts, 0, len(inputs))
				for _, input := range inputs {
					artifacts, err := buildCachedPDFAnalysisArtifacts(ctx, artifactCache, runtime, resolverConfig, surface, input, query, maxExcerptChars, opts.OCRProfile, false, false, visionModel, maxVisualPages)
					if err != nil {
						return "", fmt.Errorf("pdf_analyze: %w", err)
					}
					artifactsList = append(artifactsList, artifacts)
				}
				var documentSetPlan *pdfAnalysisPlan
				var documentSetVisual *pdfVisualAnalysis
				if includeVisualAnalysis {
					plan := buildPDFDocumentSetAnalysisPlan(artifactsList)
					documentSetPlan = &plan
					documentSetVisual = runPDFDocumentSetVisualAnalysis(ctx, resolvedPDFToolInputPaths(inputs), query, artifactsList, plan, visionModel)
					if documentSetVisual == nil || strings.ToLower(strings.TrimSpace(documentSetVisual.Status)) != "success" {
						for i := range artifactsList {
							artifactsList[i] = enrichPDFAnalysisArtifactsWithVisualAnalysis(ctx, artifactsList[i], query, visionModel, maxVisualPages)
						}
					}
				}
				payload := buildPDFAnalyzeMultiPayload(artifactsList, query, maxHits, maxExcerptChars, includePageMap, includeOutline, documentSetPlan, documentSetVisual)
				blob, err := json.Marshal(payload)
				if err != nil {
					return "", err
				}
				return string(blob), nil
			}
			input := inputs[0]
			artifacts, err := buildCachedPDFAnalysisArtifacts(ctx, artifactCache, runtime, resolverConfig, surface, input, query, maxExcerptChars, opts.OCRProfile, includeVisualAnalysis, false, visionModel, maxVisualPages)
			if err != nil {
				return "", fmt.Errorf("pdf_analyze: %w", err)
			}
			payload := pdfAnalyzeSinglePayload{
				Path:            artifacts.DisplayPath,
				FilesTouched:    pdfVisualAnalysisTouchedPaths(artifacts.VisualAnalysis),
				Backend:         artifacts.BackendStatus.ExtractBackend,
				BackendStatus:   artifacts.BackendStatus,
				PageCount:       artifacts.Metadata.PageCount,
				Mode:            "overview",
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
				payload.PageMap = artifacts.PageMap
			}
			if includeOutline {
				payload.Outline = convertPDFOutline(artifacts.Metadata.Outline)
			}
			if query != "" {
				payload.Mode = "search"
				payload.Query = query
				payload.Hits = searchPDFAnalyzeHits(artifacts.TextResult.Pages, query, maxHits, maxExcerptChars)
				if len(payload.Hits) > 0 {
					payload.KeyPages = topPDFAnalyzePagesFromHits(payload.Hits, artifacts.PageMap, 3)
				}
			}
			payload = applyPDFUnifiedFocusToAnalyzeSinglePayload(payload, buildPDFUnifiedFocusSummaryFromAnalysisArtifacts(query, []pdfAnalysisArtifacts{artifacts}))
			blob, err := json.Marshal(payload)
			if err != nil {
				return "", err
			}
			return string(blob), nil
		})
	}

	if toolEnabled(enabled, "pdf_extract_structured") {
		reg.Register(pdfExtractStructuredDefinition(surface.Unified), func(ctx context.Context, call types.FunctionCall) (string, error) {
			ctx = contextWithPDFHost(ctx, opts.Host)
			params, err := decodeArgs(call.Arguments)
			if err != nil {
				return "", err
			}
			inputs, err := resolvePDFToolInputs(ctx, opts.Root, params, defaultRemoteTimeoutMs, defaultRemoteMaxBytes, hardPDFToolInputCount)
			if err != nil {
				return "", fmt.Errorf("pdf_extract_structured: %w", err)
			}
			defer cleanupPDFToolInputs(inputs)
			query := strings.TrimSpace(firstString(params, "query"))
			maxExcerptChars := clampToolLimit(firstInt(params, "max_excerpt_chars"), 280, 1_200)
			includePageMap := readBool(params, "include_page_map")
			includeVisualAnalysis := true
			if _, ok := params["include_visual_analysis"]; ok {
				includeVisualAnalysis = readBool(params, "include_visual_analysis")
			}
			visionModel := strings.TrimSpace(firstString(params, "vision_model"))
			maxVisualPages := clampToolLimit(firstInt(params, "max_visual_pages"), defaultPDFVisualPages, hardPDFVisualPages)
			if len(inputs) > 1 {
				var documentSetPlan *pdfAnalysisPlan
				var documentSetVisual *pdfVisualAnalysis
				artifactsList := make([]pdfAnalysisArtifacts, 0, len(inputs))
				for _, input := range inputs {
					artifacts, err := buildCachedPDFAnalysisArtifacts(ctx, artifactCache, runtime, resolverConfig, surface, input, query, maxExcerptChars, opts.OCRProfile, includeVisualAnalysis, includeVisualAnalysis, visionModel, maxVisualPages)
					if err != nil {
						return "", fmt.Errorf("pdf_extract_structured: %w", err)
					}
					artifactsList = append(artifactsList, artifacts)
				}
				if includeVisualAnalysis {
					plan := buildPDFDocumentSetAnalysisPlan(artifactsList)
					documentSetPlan = &plan
					documentSetVisual = runPDFDocumentSetVisualAnalysis(ctx, resolvedPDFToolInputPaths(inputs), query, artifactsList, plan, visionModel)
				}
				focusSummary := buildPDFUnifiedFocusSummaryFromAnalysisArtifacts(query, artifactsList)
				documents := make([]pdfStructuredPayload, 0, len(artifactsList))
				for idx, artifacts := range artifactsList {
					documents = append(documents, applyPDFUnifiedFocusToStructuredPayload(buildPDFStructuredPayload(artifacts, query, includePageMap, documentSetPlan, documentSetVisual), focusSummary, idx))
				}
				payload := applyPDFUnifiedFocusToStructuredMultiPayload(buildPDFStructuredMultiPayload(documents, query, documentSetPlan, documentSetVisual), focusSummary)
				blob, err := json.Marshal(payload)
				if err != nil {
					return "", err
				}
				return string(blob), nil
			}
			input := inputs[0]
			artifacts, err := buildCachedPDFAnalysisArtifacts(ctx, artifactCache, runtime, resolverConfig, surface, input, query, maxExcerptChars, opts.OCRProfile, includeVisualAnalysis, includeVisualAnalysis, visionModel, maxVisualPages)
			if err != nil {
				return "", fmt.Errorf("pdf_extract_structured: %w", err)
			}
			payload := buildPDFStructuredPayload(artifacts, query, includePageMap, nil, nil)
			blob, err := json.Marshal(payload)
			if err != nil {
				return "", err
			}
			return string(blob), nil
		})
	}
	return nil
}

func pdfExtractDefinition(unifiedAvailable bool) types.Tool {
	return types.Tool{
		Type: "function",
		Function: types.Function{
			Name:        "pdf_extract",
			Description: pdfSpecialistToolDescription("precise full-text extraction", unifiedAvailable),
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"pdf":                map[string]any{"type": "string"},
					"path":               map[string]any{"type": "string"},
					"file_path":          map[string]any{"type": "string"},
					"max_chars":          map[string]any{"type": "integer", "minimum": 1},
					"max_response_bytes": map[string]any{"type": "integer", "minimum": 1},
					"max_bytes_mb":       map[string]any{"type": "number", "minimum": 0},
					"maxBytesMb":         map[string]any{"type": "number", "minimum": 0},
					"timeout_ms":         map[string]any{"type": "integer", "minimum": 1},
					"include_pages":      map[string]any{"type": "boolean"},
					"include_outline":    map[string]any{"type": "boolean"},
				},
			},
		},
	}
}

// PDFDefinition returns one model-facing PDF tool declaration. The boolean
// controls only specialist guidance text and never enables execution.
func PDFDefinition(name string, unifiedAvailable bool) (types.Tool, bool) {
	switch strings.TrimSpace(name) {
	case "pdf":
		return pdfUnifiedDefinition(), true
	case "pdf_extract":
		return pdfExtractDefinition(unifiedAvailable), true
	case "pdf_read_pages":
		return pdfReadPagesDefinition(unifiedAvailable), true
	case "pdf_outline":
		return pdfOutlineDefinition(unifiedAvailable), true
	case "pdf_analyze":
		return pdfAnalyzeDefinition(unifiedAvailable), true
	case "pdf_extract_structured":
		return pdfExtractStructuredDefinition(unifiedAvailable), true
	default:
		return types.Tool{}, false
	}
}

func pdfReadPagesDefinition(unifiedAvailable bool) types.Tool {
	return types.Tool{
		Type: "function",
		Function: types.Function{
			Name:        "pdf_read_pages",
			Description: pdfSpecialistToolDescription("exact selected-page text reads", unifiedAvailable),
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"pdf":       map[string]any{"type": "string"},
					"path":      map[string]any{"type": "string"},
					"file_path": map[string]any{"type": "string"},
					"pages": map[string]any{
						"anyOf": []any{
							map[string]any{"type": "array", "items": map[string]any{"type": "integer", "minimum": 1}},
							map[string]any{"type": "string", "description": `Page range like "1-3,5".`},
						},
					},
					"page_range":         map[string]any{"type": "string", "description": `Page range like "1-3,5".`},
					"pageRange":          map[string]any{"type": "string", "description": `Page range like "1-3,5".`},
					"start_page":         map[string]any{"type": "integer", "minimum": 1},
					"end_page":           map[string]any{"type": "integer", "minimum": 1},
					"max_response_bytes": map[string]any{"type": "integer", "minimum": 1},
					"max_bytes_mb":       map[string]any{"type": "number", "minimum": 0},
					"maxBytesMb":         map[string]any{"type": "number", "minimum": 0},
					"timeout_ms":         map[string]any{"type": "integer", "minimum": 1},
					"max_chars_per_page": map[string]any{"type": "integer", "minimum": 1},
					"include_empty":      map[string]any{"type": "boolean"},
				},
			},
		},
	}
}

func pdfOutlineDefinition(unifiedAvailable bool) types.Tool {
	return types.Tool{
		Type: "function",
		Function: types.Function{
			Name:        "pdf_outline",
			Description: pdfSpecialistToolDescription("outline and TOC inspection", unifiedAvailable),
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"pdf":                map[string]any{"type": "string"},
					"path":               map[string]any{"type": "string"},
					"file_path":          map[string]any{"type": "string"},
					"max_response_bytes": map[string]any{"type": "integer", "minimum": 1},
					"max_bytes_mb":       map[string]any{"type": "number", "minimum": 0},
					"maxBytesMb":         map[string]any{"type": "number", "minimum": 0},
					"timeout_ms":         map[string]any{"type": "integer", "minimum": 1},
					"include_fonts":      map[string]any{"type": "boolean"},
				},
			},
		},
	}
}

func pdfAnalyzeDefinition(unifiedAvailable bool) types.Tool {
	return types.Tool{
		Type: "function",
		Function: types.Function{
			Name:        "pdf_analyze",
			Description: pdfSpecialistToolDescription("semantic overview, search signals, and key-page inspection across one or more PDFs", unifiedAvailable),
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"pdf":                     map[string]any{"type": "string"},
					"pdfs":                    map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
					"path":                    map[string]any{"type": "string"},
					"file_path":               map[string]any{"type": "string"},
					"max_response_bytes":      map[string]any{"type": "integer", "minimum": 1},
					"max_bytes_mb":            map[string]any{"type": "number", "minimum": 0},
					"maxBytesMb":              map[string]any{"type": "number", "minimum": 0},
					"timeout_ms":              map[string]any{"type": "integer", "minimum": 1},
					"query":                   map[string]any{"type": "string"},
					"max_hits":                map[string]any{"type": "integer", "minimum": 1},
					"max_excerpt_chars":       map[string]any{"type": "integer", "minimum": 40},
					"include_visual_analysis": map[string]any{"type": "boolean"},
					"vision_model":            map[string]any{"type": "string"},
					"max_visual_pages":        map[string]any{"type": "integer", "minimum": 1},
					"include_outline":         map[string]any{"type": "boolean"},
					"include_page_map":        map[string]any{"type": "boolean"},
				},
			},
		},
	}
}

func pdfExtractStructuredDefinition(unifiedAvailable bool) types.Tool {
	return types.Tool{
		Type: "function",
		Function: types.Function{
			Name:        "pdf_extract_structured",
			Description: pdfSpecialistToolDescription("chart, table, diagram, OCR, and layout-heavy structured extraction across one or more PDFs", unifiedAvailable),
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"pdf":                     map[string]any{"type": "string"},
					"pdfs":                    map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
					"path":                    map[string]any{"type": "string"},
					"file_path":               map[string]any{"type": "string"},
					"max_response_bytes":      map[string]any{"type": "integer", "minimum": 1},
					"max_bytes_mb":            map[string]any{"type": "number", "minimum": 0},
					"maxBytesMb":              map[string]any{"type": "number", "minimum": 0},
					"timeout_ms":              map[string]any{"type": "integer", "minimum": 1},
					"query":                   map[string]any{"type": "string"},
					"max_excerpt_chars":       map[string]any{"type": "integer", "minimum": 40},
					"include_visual_analysis": map[string]any{"type": "boolean"},
					"vision_model":            map[string]any{"type": "string"},
					"max_visual_pages":        map[string]any{"type": "integer", "minimum": 1},
					"include_page_map":        map[string]any{"type": "boolean"},
				},
			},
		},
	}
}

func clampToolLimit(value int, defaultValue int, hardMax int) int {
	if value <= 0 {
		value = defaultValue
	}
	if value > hardMax {
		value = hardMax
	}
	return value
}

func buildPDFPageItems(pageTexts []PDFPageText, maxPageChars int, maxPages int, includeEmpty bool) ([]pdfPageItem, bool) {
	if len(pageTexts) == 0 {
		return nil, false
	}
	limit := len(pageTexts)
	limited := false
	if limit > maxPages {
		limit = maxPages
		limited = true
	}
	items := make([]pdfPageItem, 0, limit)
	for i := 0; i < limit; i++ {
		text := pageTexts[i].Text
		if !includeEmpty && strings.TrimSpace(text) == "" {
			continue
		}
		truncated := false
		if trimmed, changed := trimToMaxChars(text, maxPageChars); changed {
			text = trimmed
			truncated = true
		}
		items = append(items, pdfPageItem{
			Page:      pageTexts[i].Page,
			Text:      text,
			Chars:     runeLen(text),
			Truncated: truncated,
		})
	}
	return items, limited
}

func unsupportedPDFToolReference(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	lower := strings.ToLower(trimmed)
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") || strings.HasPrefix(lower, "file://") {
		return ""
	}
	if len(trimmed) >= 3 && ((trimmed[1] == ':' && (trimmed[2] == '\\' || trimmed[2] == '/')) || (trimmed[0] == '\\' && trimmed[1] == '\\')) {
		return ""
	}
	if idx := strings.Index(trimmed, ":"); idx > 0 {
		scheme := trimmed[:idx]
		if isValidImageToolSchemeToken(scheme) {
			return trimmed
		}
	}
	return ""
}

func joinPDFPageTexts(pageTexts []PDFPageText) string {
	if len(pageTexts) == 0 {
		return ""
	}
	parts := make([]string, 0, len(pageTexts))
	for _, page := range pageTexts {
		parts = append(parts, page.Text)
	}
	return strings.Join(parts, "\n\n")
}

func buildPDFAnalyzePageMap(pageTexts []PDFPageText, maxExcerptChars int) []pdfAnalyzePageItem {
	if len(pageTexts) == 0 {
		return nil
	}
	items := make([]pdfAnalyzePageItem, 0, len(pageTexts))
	for _, page := range pageTexts {
		text := strings.TrimSpace(page.Text)
		items = append(items, pdfAnalyzePageItem{
			Page:    page.Page,
			Chars:   runeLen(text),
			Empty:   text == "",
			Excerpt: truncateToolText(text, maxExcerptChars),
		})
	}
	return items
}

func topPDFAnalyzePages(pageMap []pdfAnalyzePageItem, limit int) []pdfAnalyzePageItem {
	if len(pageMap) == 0 || limit <= 0 {
		return nil
	}
	cloned := append([]pdfAnalyzePageItem(nil), pageMap...)
	sort.SliceStable(cloned, func(i, j int) bool {
		if cloned[i].Chars == cloned[j].Chars {
			return cloned[i].Page < cloned[j].Page
		}
		return cloned[i].Chars > cloned[j].Chars
	})
	if len(cloned) > limit {
		cloned = cloned[:limit]
	}
	return cloned
}

func topPDFAnalyzePagesFromHits(hits []pdfAnalyzeHit, pageMap []pdfAnalyzePageItem, limit int) []pdfAnalyzePageItem {
	if len(hits) == 0 || limit <= 0 {
		return nil
	}
	byPage := make(map[int]pdfAnalyzePageItem, len(pageMap))
	for _, item := range pageMap {
		byPage[item.Page] = item
	}
	out := make([]pdfAnalyzePageItem, 0, minInt(limit, len(hits)))
	for _, hit := range hits {
		if item, ok := byPage[hit.Page]; ok {
			out = append(out, item)
		}
		if len(out) >= limit {
			break
		}
	}
	return out
}

func pdfTotalTextChars(pageTexts []PDFPageText) int {
	total := 0
	for _, page := range pageTexts {
		total += runeLen(strings.TrimSpace(page.Text))
	}
	return total
}

func pdfPagesWithText(pageTexts []PDFPageText) int {
	count := 0
	for _, page := range pageTexts {
		if strings.TrimSpace(page.Text) != "" {
			count++
		}
	}
	return count
}

func searchPDFAnalyzeHits(pageTexts []PDFPageText, query string, limit int, maxExcerptChars int) []pdfAnalyzeHit {
	query = strings.TrimSpace(query)
	if query == "" || limit <= 0 {
		return nil
	}
	tokens := pdfAnalyzeQueryTokens(query)
	if len(tokens) == 0 {
		return nil
	}
	hits := make([]pdfAnalyzeHit, 0)
	for _, page := range pageTexts {
		raw := strings.TrimSpace(page.Text)
		if raw == "" {
			continue
		}
		lower := strings.ToLower(raw)
		matches := 0
		for _, token := range tokens {
			matches += strings.Count(lower, token)
		}
		if matches == 0 {
			continue
		}
		hits = append(hits, pdfAnalyzeHit{
			Page:    page.Page,
			Matches: matches,
			Score:   matches*100 + runeLen(raw),
			Excerpt: buildPDFAnalyzeExcerpt(raw, tokens, maxExcerptChars),
		})
	}
	sort.SliceStable(hits, func(i, j int) bool {
		if hits[i].Score == hits[j].Score {
			return hits[i].Page < hits[j].Page
		}
		return hits[i].Score > hits[j].Score
	})
	if len(hits) > limit {
		hits = hits[:limit]
	}
	return hits
}

func pdfAnalyzeQueryTokens(query string) []string {
	parts := strings.Fields(strings.ToLower(strings.TrimSpace(query)))
	if len(parts) == 0 {
		return nil
	}
	seen := make(map[string]bool, len(parts))
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if len(part) < 2 || seen[part] {
			continue
		}
		seen[part] = true
		out = append(out, part)
	}
	return out
}

func buildPDFAnalyzeExcerpt(raw string, tokens []string, maxExcerptChars int) string {
	text := strings.TrimSpace(raw)
	if text == "" {
		return ""
	}
	lower := strings.ToLower(text)
	best := -1
	for _, token := range tokens {
		if idx := strings.Index(lower, token); idx >= 0 && (best == -1 || idx < best) {
			best = idx
		}
	}
	if best < 0 {
		return truncateToolText(text, maxExcerptChars)
	}
	runes := []rune(text)
	start := max(0, best-maxExcerptChars/4)
	end := minInt(len(runes), start+maxExcerptChars)
	return strings.TrimSpace(string(runes[start:end]))
}

func minInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func resolvePDFPageSelection(params map[string]any, totalPages int, maxPages int) ([]int, bool, error) {
	if totalPages <= 0 {
		return nil, false, nil
	}
	explicitPages := false
	pageList, usedPageRange, err := readPDFPageRange(params)
	if err != nil {
		return nil, false, err
	}
	if usedPageRange {
		explicitPages = true
	}
	if len(pageList) == 0 {
		pageList = readIntSlice(params["pages"])
		if _, ok := params["pages"]; ok {
			explicitPages = true
		}
	}
	if len(pageList) == 0 {
		start := firstInt(params, "start_page", "from_page", "from")
		end := firstInt(params, "end_page", "to_page", "to")
		if start > 0 || end > 0 {
			explicitPages = true
			if start <= 0 {
				start = 1
			}
			if end <= 0 {
				end = totalPages
			}
			if start > end {
				return nil, false, fmt.Errorf("invalid page range: start_page (%d) > end_page (%d)", start, end)
			}
			pageList = make([]int, 0, end-start+1)
			for i := start; i <= end; i++ {
				pageList = append(pageList, i)
			}
		}
	}
	if len(pageList) == 0 {
		limit := totalPages
		limited := false
		if limit > maxPages {
			limit = maxPages
			limited = true
		}
		pages := make([]int, 0, limit)
		for i := 1; i <= limit; i++ {
			pages = append(pages, i)
		}
		return pages, limited, nil
	}

	uniq := map[int]bool{}
	for _, page := range pageList {
		if page < 1 || page > totalPages {
			return nil, false, fmt.Errorf("page out of range: %d (total=%d)", page, totalPages)
		}
		uniq[page] = true
	}
	out := make([]int, 0, len(uniq))
	for page := range uniq {
		out = append(out, page)
	}
	sort.Ints(out)
	if explicitPages && len(out) > maxPages {
		return nil, false, fmt.Errorf("too many pages requested: %d > %d", len(out), maxPages)
	}
	return out, false, nil
}

func readPDFPageRange(params map[string]any) ([]int, bool, error) {
	raw := strings.TrimSpace(firstString(params, "page_range", "pageRange"))
	if raw == "" {
		if text, ok := params["pages"].(string); ok {
			raw = strings.TrimSpace(text)
		}
	}
	if raw == "" {
		return nil, false, nil
	}
	pages, err := parsePDFPageRange(raw)
	if err != nil {
		return nil, true, err
	}
	return pages, true, nil
}

func parsePDFPageRange(raw string) ([]int, error) {
	trimmed := strings.TrimSpace(raw)
	if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
		var jsonPages []int
		if err := json.Unmarshal([]byte(trimmed), &jsonPages); err == nil {
			if len(jsonPages) == 0 {
				return nil, fmt.Errorf("page range is required")
			}
			for _, page := range jsonPages {
				if page <= 0 {
					return nil, fmt.Errorf("invalid page range token: %s", trimmed)
				}
			}
			return jsonPages, nil
		}
		trimmed = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(trimmed, "["), "]"))
	}
	parts := strings.Split(trimmed, ",")
	pages := make([]int, 0, len(parts))
	for _, part := range parts {
		token := strings.TrimSpace(part)
		if token == "" {
			continue
		}
		if strings.Contains(token, "-") {
			rangeParts := strings.SplitN(token, "-", 2)
			if len(rangeParts) != 2 {
				return nil, fmt.Errorf("invalid page range token: %s", token)
			}
			start, err := strconv.Atoi(strings.TrimSpace(rangeParts[0]))
			if err != nil || start <= 0 {
				return nil, fmt.Errorf("invalid page range token: %s", token)
			}
			end, err := strconv.Atoi(strings.TrimSpace(rangeParts[1]))
			if err != nil || end <= 0 {
				return nil, fmt.Errorf("invalid page range token: %s", token)
			}
			if start > end {
				return nil, fmt.Errorf("invalid page range token: %s", token)
			}
			for page := start; page <= end; page++ {
				pages = append(pages, page)
			}
			continue
		}
		page, err := strconv.Atoi(token)
		if err != nil || page <= 0 {
			return nil, fmt.Errorf("invalid page range token: %s", token)
		}
		pages = append(pages, page)
	}
	if len(pages) == 0 {
		return nil, fmt.Errorf("page range is required")
	}
	return pages, nil
}

func readIntSlice(raw any) []int {
	if raw == nil {
		return nil
	}
	out := make([]int, 0)
	appendValue := func(value int) {
		if value > 0 {
			out = append(out, value)
		}
	}
	switch typed := raw.(type) {
	case []any:
		for _, item := range typed {
			appendValue(normalizeInt(item))
		}
	case []int:
		for _, item := range typed {
			appendValue(item)
		}
	case []int64:
		for _, item := range typed {
			appendValue(int(item))
		}
	case []float64:
		for _, item := range typed {
			appendValue(int(item))
		}
	case string:
		parts := strings.FieldsFunc(typed, func(r rune) bool {
			switch r {
			case ',', ' ', '\t', '\n', '\r':
				return true
			default:
				return false
			}
		})
		for _, part := range parts {
			v, err := strconv.Atoi(strings.TrimSpace(part))
			if err == nil {
				appendValue(v)
			}
		}
	default:
		appendValue(normalizeInt(typed))
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func normalizeInt(raw any) int {
	switch value := raw.(type) {
	case int:
		return value
	case int8:
		return int(value)
	case int16:
		return int(value)
	case int32:
		return int(value)
	case int64:
		return int(value)
	case uint:
		return int(value)
	case uint8:
		return int(value)
	case uint16:
		return int(value)
	case uint32:
		return int(value)
	case uint64:
		return int(value)
	case float32:
		return int(value)
	case float64:
		return int(value)
	case json.Number:
		if i, err := value.Int64(); err == nil {
			return int(i)
		}
		if f, err := value.Float64(); err == nil {
			return int(f)
		}
	case string:
		if v, err := strconv.Atoi(strings.TrimSpace(value)); err == nil {
			return v
		}
	}
	return 0
}

func convertPDFOutline(root *PDFOutline) []pdfOutlineNode {
	if root == nil {
		return nil
	}
	node := convertPDFOutlineNode(*root)
	if strings.TrimSpace(node.Title) == "" {
		if len(node.Children) == 0 {
			return nil
		}
		return node.Children
	}
	return []pdfOutlineNode{node}
}

func convertPDFOutlineNode(node PDFOutline) pdfOutlineNode {
	converted := pdfOutlineNode{
		Title: strings.TrimSpace(node.Title),
	}
	if len(node.Children) == 0 {
		return converted
	}
	children := make([]pdfOutlineNode, 0, len(node.Children))
	for _, child := range node.Children {
		convertedChild := convertPDFOutlineNode(child)
		if convertedChild.Title == "" && len(convertedChild.Children) == 0 {
			continue
		}
		children = append(children, convertedChild)
	}
	converted.Children = children
	return converted
}
