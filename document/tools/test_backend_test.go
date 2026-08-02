package tools

import "context"

type stubPDFBackend struct {
	allPages  []PDFPageText
	pageCount int
	available bool
}

func (s stubPDFBackend) Name() string { return "stub" }

func (s stubPDFBackend) Availability(context.Context) PDFBackendAvailability {
	return PDFBackendAvailability{Name: s.Name(), Available: s.available}
}

func (s stubPDFBackend) ExtractAllText(context.Context, string) (PDFTextResult, error) {
	return PDFTextResult{Pages: append([]PDFPageText(nil), s.allPages...)}, nil
}

func (s stubPDFBackend) ExtractPageText(_ context.Context, _ string, pages []int) (PDFTextResult, error) {
	if len(pages) == 0 {
		return s.ExtractAllText(context.Background(), "")
	}
	byPage := make(map[int]string, len(s.allPages))
	for _, page := range s.allPages {
		byPage[page.Page] = page.Text
	}
	out := make([]PDFPageText, 0, len(pages))
	for _, page := range pages {
		out = append(out, PDFPageText{Page: page, Text: byPage[page]})
	}
	return PDFTextResult{Pages: out}, nil
}

func (s stubPDFBackend) ReadMetadata(context.Context, string, bool) (PDFMetadataResult, error) {
	return PDFMetadataResult{PageCount: s.pageCount}, nil
}
