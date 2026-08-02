package tools

import "context"

// PDFFont is a provider-neutral font projection used by pdf_outline.
type PDFFont struct {
	Name string  `json:"name"`
	Size float64 `json:"size"`
}

// PDFOutline is a recursive document outline item.
type PDFOutline struct {
	Title    string       `json:"title"`
	Children []PDFOutline `json:"children,omitempty"`
}

// PDFPageText is one physical page and its extracted text.
type PDFPageText struct {
	Page int    `json:"page"`
	Text string `json:"text"`
}

// PDFTextResult is the ordered text returned by a PDFBackend.
type PDFTextResult struct {
	Pages []PDFPageText `json:"pages"`
}

// PDFMetadataResult contains portable metadata required by the tool layer.
type PDFMetadataResult struct {
	PageCount         int
	Outline           *PDFOutline
	Fonts             []PDFFont
	TotalRects        int
	PagesWithGraphics int
	MaxRectsPerPage   int
	GraphicPages      []int
	RectsByPage       map[int]int
}

// PDFBackendAvailability is an explicit Host probe result. Canonical tools do
// not discover binaries, Python installations or commercial SDK licenses.
type PDFBackendAvailability struct {
	Name      string `json:"name"`
	Available bool   `json:"available"`
	Reason    string `json:"reason,omitempty"`
}

// PDFBackendAvailabilityDetails describes the configured primary/fallback pair.
type PDFBackendAvailabilityDetails struct {
	Primary  PDFBackendAvailability `json:"primary"`
	Fallback PDFBackendAvailability `json:"fallback"`
}

// PDFBackend owns concrete PDF extraction and metadata access.
type PDFBackend interface {
	Name() string
	Availability(context.Context) PDFBackendAvailability
	ExtractAllText(context.Context, string) (PDFTextResult, error)
	ExtractPageText(context.Context, string, []int) (PDFTextResult, error)
	ReadMetadata(context.Context, string, bool) (PDFMetadataResult, error)
}

// PDFLayoutBackend is the optional layout-preserving extension used for table
// and field-alignment evidence.
type PDFLayoutBackend interface {
	ExtractLayoutText(context.Context, string, []int) (PDFTextResult, error)
}

type pdfLayoutTextBackend interface {
	extractLayoutText(context.Context, string, []int) (PDFTextResult, error)
}

type exportedLayoutBackendAdapter struct{ PDFLayoutBackend }

func (a exportedLayoutBackendAdapter) extractLayoutText(ctx context.Context, path string, pages []int) (PDFTextResult, error) {
	return a.ExtractLayoutText(ctx, path, pages)
}

func asPDFLayoutBackend(backend PDFBackend) pdfLayoutTextBackend {
	if internal, ok := backend.(pdfLayoutTextBackend); ok {
		return internal
	}
	if exported, ok := backend.(PDFLayoutBackend); ok {
		return exportedLayoutBackendAdapter{PDFLayoutBackend: exported}
	}
	return nil
}
