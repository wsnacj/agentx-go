package tools

import "testing"

func TestPDFSurfaceClassification(t *testing.T) {
	for _, name := range []string{"pdf", " PDF "} {
		if got := PDFSurfaceForToolName(name); got != PDFSurfaceUnified {
			t.Fatalf("PDFSurfaceForToolName(%q) = %q, want %q", name, got, PDFSurfaceUnified)
		}
	}
	for _, name := range []string{"pdf_extract", "pdf_extract_structured", "pdf_analyze", "pdf_read_pages", "pdf_outline"} {
		if got := PDFSurfaceForToolName(name); got != PDFSurfaceSpecialist {
			t.Fatalf("PDFSurfaceForToolName(%q) = %q, want %q", name, got, PDFSurfaceSpecialist)
		}
	}
	if got := PDFSurfaceForToolName("exec"); got != "" {
		t.Fatalf("PDFSurfaceForToolName(non-pdf) = %q, want empty", got)
	}
}

func TestNormalizeAndResolvePDFSurface(t *testing.T) {
	if got := NormalizePDFSurface(" pdf_specialist "); got != PDFSurfaceSpecialist {
		t.Fatalf("NormalizePDFSurface() = %q, want %q", got, PDFSurfaceSpecialist)
	}
	if got := PDFSurfaceFallbackEntrypoint(PDFSurfaceUnified); got != "pdf" {
		t.Fatalf("PDFSurfaceFallbackEntrypoint(unified) = %q, want pdf", got)
	}
	if got := PDFSurfaceFallbackEntrypoint(PDFSurfaceSpecialist); got != "pdf_extract" {
		t.Fatalf("PDFSurfaceFallbackEntrypoint(specialist) = %q, want pdf_extract", got)
	}
	if surface, entrypoint := ResolvePDFSurface("pdf_specialist", ""); surface != PDFSurfaceSpecialist || entrypoint != "pdf_extract" {
		t.Fatalf("ResolvePDFSurface(surface only) = (%q, %q), want (%q, %q)", surface, entrypoint, PDFSurfaceSpecialist, "pdf_extract")
	}
	if surface, entrypoint := ResolvePDFSurface("", "pdf"); surface != PDFSurfaceUnified || entrypoint != "pdf" {
		t.Fatalf("ResolvePDFSurface(entrypoint unified) = (%q, %q), want (%q, %q)", surface, entrypoint, PDFSurfaceUnified, "pdf")
	}
	if surface, entrypoint := ResolvePDFSurface("", "pdf_read_pages"); surface != PDFSurfaceSpecialist || entrypoint != "pdf_read_pages" {
		t.Fatalf("ResolvePDFSurface(entrypoint specialist) = (%q, %q), want (%q, %q)", surface, entrypoint, PDFSurfaceSpecialist, "pdf_read_pages")
	}
}
