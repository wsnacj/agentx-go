package model

import (
	"time"
)

// OperationKind indicates which OCR pipeline to invoke.
type OperationKind string

const (
	OperationKindAny   OperationKind = "any"
	OperationKindOCR   OperationKind = "ocr"
	OperationKindTable OperationKind = "table"
	OperationKindStamp OperationKind = "stamp"
)

// Request defines the standard input for a recognition workflow.
type Request struct {
	Paths     []string
	MaxPages  int
	Options   map[string]any
	Meta      map[string]string
	CreatedAt time.Time
}

// Response is the generic output returned by a pipeline.
type Response struct {
	Meta    Meta
	Payload any
	Diff    *DiffSummary
}

// Meta captures execution metadata.
type Meta struct {
	ProcessedPaths []string
	Duration       time.Duration
	Warnings       []string
	Diagnostics    map[string]any
}

// TableResponse is the specialized response for table OCR requests.
type TableResponse struct {
	Meta    Meta
	Payload TablePayload
	Diff    *DiffSummary
}

// StampResponse is the specialized response for stamp detection.
type StampResponse struct {
	Meta    Meta
	Payload StampPayload
}

// TablePayload represents table structured recognition result.
type TablePayload struct {
	Pages                  []TablePage
	Raw                    [][]byte
	Files                  []string
	HTMLPages              []string
	CombinedHTML           string
	Text                   []string
	NormalizedTexts        []string
	NormalizedCombinedText string
}

// TablePage captures the parsed payload for a single table page.
type TablePage struct {
	Index       int
	Width       int
	Height      int
	Recognized  string
	Coordinates []Coordinate
}

// StampPayload collects stamp detection details.
type StampPayload struct {
	Pages []StampPage
	Raw   [][]byte
	Files []string
}

// DiffSummary 汇总 diff/fuzzy 相关信息。
type DiffSummary struct {
	OCRDiff   *DiffResultSummary `json:"ocr_diff,omitempty"`
	TableDiff *DiffResultSummary `json:"table_diff,omitempty"`
	Notes     []string           `json:"notes,omitempty"`
}

type DiffResultSummary struct {
	HasDiff       bool     `json:"has_diff"`
	ExtraPages    int      `json:"extra_pages"`
	DiffPageCount int      `json:"diff_page_count"`
	MappingScheme string   `json:"mapping_scheme"`
	Preview       []string `json:"preview,omitempty"`
}

// OCRPayload carries raw provider responses for later parsing or merging.
type OCRPayload struct {
	Files               []string
	RawResponses        [][]byte
	HTMLPages           []string
	CombinedHTML        string
	RecognizedText      string
	NormalizedText      string
	Coordinates         []Coordinate
	TextBoxes           []TextBox
	PageTexts           []string
	NormalizedPageTexts []string
}

// StampPage holds detection details per page.
type StampPage struct {
	Index int
	Stamp []StampDetail
}

// StampDetail holds metadata for a single detected stamp.
type StampDetail struct {
	Text     string
	Type     string
	Shape    string
	Color    string
	Position []int
}

// Coordinate captures bounding box information.
type Coordinate struct {
	Left   int
	Top    int
	Right  int
	Bottom int
}

// TextBox associates recognized text with a bounding box on a page.
type TextBox struct {
	Page       int
	Text       string
	Coordinate Coordinate
	PageWidth  int
	PageHeight int
	Confidence float64
	Level      string
}
