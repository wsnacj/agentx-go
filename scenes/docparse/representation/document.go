// Package representation defines provider-neutral document views used by Docparse.
package representation

import (
	"fmt"
	"strings"
	"unicode"
)

const (
	DefaultPlainTextChunkSize = 800
	DefaultMaxPromptChars     = 24000
	minUsableTextRunes        = 16

	TextSourcePlainText    = "text"
	TextSourcePDFTextLayer = "pdf_text_layer"
	TextSourceOCRX         = "ocrx"
	TextSourceOCRXTable    = "ocrx_table"

	ExtractionModeTableFirst     ExtractionMode = "table_first"
	ExtractionModeTextLayerFirst ExtractionMode = "text_layer_first"
	ExtractionModeOCRFirst       ExtractionMode = "ocr_first"
	ExtractionModeAuto           ExtractionMode = "auto"

	DefaultExtractionMode = ExtractionModeTableFirst
)

// ExtractionMode controls the first-phase document representation strategy.
type ExtractionMode string

// Page is the normalized per-page text representation consumed by docparse
// adapters. HTML/table/stamp fields will be added by later route stages.
type Page struct {
	Number int    `json:"number"`
	Text   string `json:"text"`
	Source string `json:"source,omitempty"`
}

// Document is the shared representation passed through docparse planning and
// adapter execution.
type Document struct {
	DocumentPath string      `json:"document_path"`
	Category     string      `json:"category,omitempty"`
	TextSource   string      `json:"text_source,omitempty"`
	Pages        []Page      `json:"pages,omitempty"`
	FullText     string      `json:"full_text,omitempty"`
	Diagnostics  Diagnostics `json:"diagnostics,omitempty"`
}

// Clone returns a copy that can be safely held by caches or adapters.
func (d Document) Clone() Document {
	out := d
	out.Pages = append([]Page(nil), d.Pages...)
	out.Diagnostics.PDFTextLayerErrors = append([]string(nil), d.Diagnostics.PDFTextLayerErrors...)
	out.Diagnostics.Warnings = append([]string(nil), d.Diagnostics.Warnings...)
	return out
}

// Diagnostics records extraction choices without embedding business semantics.
type Diagnostics struct {
	ExtractionMode     string   `json:"extraction_mode,omitempty"`
	PDFTextLayerEngine string   `json:"pdf_text_layer_engine,omitempty"`
	PDFTextLayerErrors []string `json:"pdf_text_layer_errors,omitempty"`
	OCRXFallbackUsed   bool     `json:"ocrx_fallback_used,omitempty"`
	OCRXTableUsed      bool     `json:"ocrx_table_used,omitempty"`
	Warnings           []string `json:"warnings,omitempty"`
}

// PageRange identifies a contiguous 1-based page range.
type PageRange struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

// NormalizeExtractionMode normalizes host-facing extraction mode aliases.
func NormalizeExtractionMode(raw string) (ExtractionMode, error) {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	normalized = strings.ReplaceAll(normalized, "-", "_")
	normalized = strings.ReplaceAll(normalized, " ", "_")
	switch normalized {
	case "", "default", "table", "tablefirst", "table_first":
		return ExtractionModeTableFirst, nil
	case "text", "pdf", "pdf_text", "pdf_text_layer", "textlayerfirst", "text_layer_first":
		return ExtractionModeTextLayerFirst, nil
	case "ocr", "ocrfirst", "ocr_first":
		return ExtractionModeOCRFirst, nil
	case "auto":
		return ExtractionModeAuto, nil
	default:
		return "", fmt.Errorf("unknown docparse extraction mode %q", raw)
	}
}

// FromTextPages wraps host-provided pages as a representation.
func FromTextPages(path string, pages []string, source string) (Document, error) {
	pages = TrimPages(pages)
	if !PagesHaveUsableText(pages) {
		return Document{}, fmt.Errorf("document representation has no usable text")
	}
	source = strings.TrimSpace(source)
	if source == "" {
		source = TextSourcePlainText
	}
	return Document{
		DocumentPath: path,
		TextSource:   source,
		Pages:        makePages(pages, source),
		FullText:     strings.Join(pages, "\n\n"),
	}, nil
}

// TextPages returns page text in order.
func (d Document) TextPages() []string {
	out := make([]string, 0, len(d.Pages))
	for _, page := range d.Pages {
		text := strings.TrimSpace(page.Text)
		if text != "" {
			out = append(out, text)
		}
	}
	return out
}

// PromptText renders page text with [PAGE n] labels for LLM projection.
func (d Document) PromptText(maxChars int) string {
	return BuildPagePromptTextFromPages(d.Pages, maxChars)
}

// SelectPageRange returns a document view limited to a contiguous page range.
// Page labels keep their original document numbers.
func (d Document) SelectPageRange(pageRange PageRange) (Document, error) {
	if pageRange.Start <= 0 && pageRange.End <= 0 {
		return d, nil
	}
	if pageRange.Start <= 0 {
		pageRange.Start = 1
	}
	if pageRange.End <= 0 {
		pageRange.End = pageRange.Start
	}
	if pageRange.End < pageRange.Start {
		return Document{}, fmt.Errorf("invalid page range: %d-%d", pageRange.Start, pageRange.End)
	}
	selected := make([]Page, 0, len(d.Pages))
	texts := make([]string, 0, len(d.Pages))
	for _, page := range d.Pages {
		if page.Number < pageRange.Start || page.Number > pageRange.End {
			continue
		}
		if strings.TrimSpace(page.Text) == "" {
			continue
		}
		selected = append(selected, page)
		texts = append(texts, strings.TrimSpace(page.Text))
	}
	if len(selected) == 0 {
		return Document{}, fmt.Errorf("page range %d-%d selected no usable pages", pageRange.Start, pageRange.End)
	}
	out := d
	out.Pages = selected
	out.FullText = strings.Join(texts, "\n\n")
	return out, nil
}

// ParsePageRange parses a compact 1-based page range such as "2" or "2-4".
func ParsePageRange(value string) (PageRange, bool, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return PageRange{}, false, nil
	}
	parts := strings.Split(value, "-")
	if len(parts) > 2 {
		return PageRange{}, false, fmt.Errorf("invalid page range %q", value)
	}
	start, err := parsePositiveInt(parts[0])
	if err != nil {
		return PageRange{}, false, fmt.Errorf("invalid page range %q", value)
	}
	end := start
	if len(parts) == 2 {
		end, err = parsePositiveInt(parts[1])
		if err != nil {
			return PageRange{}, false, fmt.Errorf("invalid page range %q", value)
		}
	}
	if end < start {
		return PageRange{}, false, fmt.Errorf("invalid page range %q", value)
	}
	return PageRange{Start: start, End: end}, true, nil
}

// BuildPagePromptText renders page text with stable labels.
func BuildPagePromptText(pages []string, maxChars int) string {
	return BuildPagePromptTextFromPages(makePages(pages, TextSourcePlainText), maxChars)
}

// BuildPagePromptTextFromPages renders page text with stable page labels.
func BuildPagePromptTextFromPages(pages []Page, maxChars int) string {
	if maxChars <= 0 {
		maxChars = DefaultMaxPromptChars
	}
	var b strings.Builder
	used := 0
	for idx, page := range pages {
		text := strings.TrimSpace(page.Text)
		if text == "" {
			continue
		}
		pageNumber := page.Number
		if pageNumber <= 0 {
			pageNumber = idx + 1
		}
		next := fmt.Sprintf("[PAGE %d]\n%s\n\n", pageNumber, text)
		nextRunes := []rune(next)
		if used+len(nextRunes) > maxChars {
			remain := maxChars - used
			if remain > 0 {
				b.WriteString(string(nextRunes[:remain]))
			}
			break
		}
		b.WriteString(next)
		used += len(nextRunes)
	}
	return b.String()
}

func parsePositiveInt(value string) (int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, fmt.Errorf("empty page number")
	}
	out := 0
	for _, r := range value {
		if r < '0' || r > '9' {
			return 0, fmt.Errorf("invalid page number")
		}
		out = out*10 + int(r-'0')
	}
	if out <= 0 {
		return 0, fmt.Errorf("page number must be positive")
	}
	return out, nil
}

// SplitTextByChars splits plain text into stable pseudo-pages.
func SplitTextByChars(text string, chunkSize int) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	if chunkSize <= 0 {
		chunkSize = DefaultPlainTextChunkSize
	}
	runes := []rune(text)
	pages := make([]string, 0, (len(runes)+chunkSize-1)/chunkSize)
	for start := 0; start < len(runes); start += chunkSize {
		end := start + chunkSize
		if end > len(runes) {
			end = len(runes)
		}
		page := strings.TrimSpace(string(runes[start:end]))
		if page != "" {
			pages = append(pages, page)
		}
	}
	return pages
}

// TrimPages removes empty pages and trims whitespace.
func TrimPages(pages []string) []string {
	out := make([]string, 0, len(pages))
	for _, page := range pages {
		page = strings.TrimSpace(page)
		if page != "" {
			out = append(out, page)
		}
	}
	return out
}

// PagesHaveUsableText returns true when pages have enough non-space text to
// feed downstream parsing.
func PagesHaveUsableText(pages []string) bool {
	usableRunes := 0
	for _, page := range pages {
		for _, r := range strings.TrimSpace(page) {
			if !unicode.IsSpace(r) {
				usableRunes++
				if usableRunes >= minUsableTextRunes {
					return true
				}
			}
		}
	}
	return false
}

// PageTextScore scores usable text quantity for PDF engine selection.
func PageTextScore(pages []string) int {
	score := 0
	for _, page := range pages {
		for _, r := range page {
			if !unicode.IsSpace(r) {
				score++
			}
		}
	}
	return score
}

func makePages(pages []string, source string) []Page {
	out := make([]Page, 0, len(pages))
	for idx, text := range pages {
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}
		out = append(out, Page{
			Number: idx + 1,
			Text:   text,
			Source: source,
		})
	}
	return out
}
