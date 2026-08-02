package tools

import (
	"context"
	"fmt"
	"sort"
	"strings"

	types "github.com/wsnacj/agentx-go/components/llm"
)

func buildPDFUnifiedPromptVisuals(
	ctx context.Context,
	documents []pdfUnifiedDocumentArtifacts,
	focuses []pdfUnifiedDocumentFocus,
	queryClass string,
	maxVisualPages int,
) ([]types.VisualContent, []pdfUnifiedDocumentArtifacts, []string, func(), error) {
	if len(documents) == 0 {
		return nil, nil, nil, nil, nil
	}
	out := append([]pdfUnifiedDocumentArtifacts(nil), documents...)
	visuals := make([]types.VisualContent, 0, len(documents)*4)
	warnings := make([]string, 0, len(documents))
	cleanups := make([]func() error, 0, len(documents))
	cleanup := func() {
		for _, fn := range cleanups {
			if fn != nil {
				_ = fn()
			}
		}
	}
	for idx := range out {
		var focus pdfUnifiedDocumentFocus
		if idx < len(focuses) {
			focus = focuses[idx]
		}
		pages := selectPDFUnifiedVisualPages(out[idx], focus, queryClass, maxVisualPages)
		if len(pages) == 0 {
			continue
		}
		if len(pages) > maxVisualPages {
			pages = append([]int(nil), pages[:maxVisualPages]...)
		}
		rendered, fn, err := pdfRenderPDFPages(ctx, out[idx].Path, pages, defaultPDFVisualDPI)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("render %s visuals failed: %v", out[idx].DisplayPath, err))
			continue
		}
		if fn != nil {
			cleanups = append(cleanups, fn)
		}
		visuals = append(visuals, types.NewTextBlock(fmt.Sprintf("Rendered pages for %s", out[idx].DisplayPath)))
		renderedVisuals, err := buildPDFVisualContents(rendered, out[idx].PageMap)
		if err != nil {
			cleanup()
			return nil, nil, nil, nil, err
		}
		visuals = append(visuals, renderedVisuals...)
		out[idx].VisualPages = append([]int(nil), pages...)
	}
	return visuals, out, warnings, cleanup, nil
}

func selectPDFUnifiedVisualPages(document pdfUnifiedDocumentArtifacts, focus pdfUnifiedDocumentFocus, queryClass string, maxVisualPages int) []int {
	limit := clampToolLimit(maxVisualPages, defaultPDFVisualPages, hardPDFVisualPages)
	if len(document.SelectedPages) > 0 {
		pages := append([]int(nil), document.SelectedPages...)
		if len(pages) > limit {
			pages = pages[:limit]
		}
		return pages
	}
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
	switch strings.TrimSpace(queryClass) {
	case pdfUnifiedQueryClassChartSummary:
		if focus.Primary != nil {
			for _, page := range focus.Primary.Pages {
				add(page)
			}
		}
		for _, segment := range focus.Supporting {
			for _, page := range segment.Pages {
				add(page)
			}
		}
	case pdfUnifiedQueryClassFieldCompare:
		if focus.Primary != nil {
			for _, page := range focus.Primary.Pages {
				add(page)
			}
		}
		for _, segment := range selectPDFUnifiedSupportingSegmentsForPrompt(focus, queryClass) {
			for _, page := range segment.Pages {
				add(page)
			}
		}
	}
	if len(selected) < limit {
		for _, page := range selectPDFVisualPages(document.PageMap, document.MediaProfile, maxVisualPages) {
			add(page)
		}
	}
	sort.Ints(selected)
	return selected
}
