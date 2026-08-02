package tools

import (
	"strings"

	agentxtools "github.com/wsnacj/agentx-go/tools"
)

const (
	PDFSurfaceUnified    = "pdf_unified"
	PDFSurfaceSpecialist = "pdf_specialist"
)

var pdfUnifiedToolNames = []string{"pdf"}

var pdfSpecialistToolNames = []string{
	"pdf_extract",
	"pdf_extract_structured",
	"pdf_analyze",
	"pdf_read_pages",
	"pdf_outline",
}

func PDFUnifiedToolNames() []string {
	out := make([]string, len(pdfUnifiedToolNames))
	copy(out, pdfUnifiedToolNames)
	return out
}

func PDFSpecialistToolNames() []string {
	out := make([]string, len(pdfSpecialistToolNames))
	copy(out, pdfSpecialistToolNames)
	return out
}

func PDFAllToolNames() []string {
	out := make([]string, 0, len(pdfUnifiedToolNames)+len(pdfSpecialistToolNames))
	out = append(out, pdfUnifiedToolNames...)
	out = append(out, pdfSpecialistToolNames...)
	return out
}

func IsPDFUnifiedToolName(name string) bool {
	return agentxtools.NormalizeToolName(name) == "pdf"
}

func IsPDFSpecialistToolName(name string) bool {
	switch agentxtools.NormalizeToolName(name) {
	case "pdf_extract", "pdf_extract_structured", "pdf_analyze", "pdf_read_pages", "pdf_outline":
		return true
	default:
		return false
	}
}

func PDFSurfaceForToolName(name string) string {
	switch {
	case IsPDFUnifiedToolName(name):
		return PDFSurfaceUnified
	case IsPDFSpecialistToolName(name):
		return PDFSurfaceSpecialist
	default:
		return ""
	}
}

func NormalizePDFSurface(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case PDFSurfaceUnified:
		return PDFSurfaceUnified
	case PDFSurfaceSpecialist:
		return PDFSurfaceSpecialist
	default:
		return ""
	}
}

func PDFSurfaceFallbackEntrypoint(surface string) string {
	switch NormalizePDFSurface(surface) {
	case PDFSurfaceUnified:
		return "pdf"
	case PDFSurfaceSpecialist:
		return "pdf_extract"
	default:
		return ""
	}
}

func ResolvePDFSurface(surface string, entrypoint string) (string, string) {
	normalizedSurface := NormalizePDFSurface(surface)
	normalizedEntrypoint := strings.TrimSpace(entrypoint)
	if normalizedSurface != "" {
		if normalizedEntrypoint == "" {
			normalizedEntrypoint = PDFSurfaceFallbackEntrypoint(normalizedSurface)
		}
		return normalizedSurface, normalizedEntrypoint
	}
	switch normalizedEntrypoint {
	case "pdf":
		return PDFSurfaceUnified, normalizedEntrypoint
	case "pdf_extract", "pdf_extract_structured", "pdf_analyze", "pdf_read_pages", "pdf_outline":
		return PDFSurfaceSpecialist, normalizedEntrypoint
	default:
		return "", normalizedEntrypoint
	}
}
