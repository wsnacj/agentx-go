package pipeline

import (
	"fmt"
	"strings"
)

// PDFParseMode controls the host document loader's PDF extraction strategy.
type PDFParseMode int

const (
	PDFParseSimple PDFParseMode = iota
	PDFParseNormal
	PDFParseForceOCR
)

// DocumentExtractionMode controls the preferred source order used by the host
// document loader.
type DocumentExtractionMode string

const (
	DocumentExtractionModeDefault        DocumentExtractionMode = ""
	DocumentExtractionModeLegacy         DocumentExtractionMode = "legacy"
	DocumentExtractionModeTableFirst     DocumentExtractionMode = "table_first"
	DocumentExtractionModeTextLayerFirst DocumentExtractionMode = "text_layer_first"
	DocumentExtractionModeOCRFirst       DocumentExtractionMode = "ocr_first"
	DocumentExtractionModeAuto           DocumentExtractionMode = "auto"
)

// NormalizeDocumentExtractionMode normalizes documented aliases without
// selecting a concrete extractor.
func NormalizeDocumentExtractionMode(raw string) (DocumentExtractionMode, error) {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	normalized = strings.ReplaceAll(normalized, "-", "_")
	normalized = strings.ReplaceAll(normalized, " ", "_")
	switch normalized {
	case "", "default", "legacy":
		return DocumentExtractionModeLegacy, nil
	case "table", "tablefirst", "table_first":
		return DocumentExtractionModeTableFirst, nil
	case "text", "pdf", "pdf_text", "pdf_text_layer", "textlayerfirst", "text_layer_first":
		return DocumentExtractionModeTextLayerFirst, nil
	case "ocr", "ocrfirst", "ocr_first":
		return DocumentExtractionModeOCRFirst, nil
	case "auto":
		return DocumentExtractionModeAuto, nil
	default:
		return "", fmt.Errorf("unknown document extraction mode %q", raw)
	}
}
