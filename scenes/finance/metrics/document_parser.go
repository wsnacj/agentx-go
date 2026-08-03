package metrics

import "context"

// ReportDocumentParseInput is the pack-level contract for adapters that parse
// downloaded financial-report artifacts. The pack owns this neutral shape; the
// concrete parser implementation stays in project/plugin adapters.
type ReportDocumentParseInput struct {
	ReportPath string
	SourceURL  string
	SpecPath   string
	ModelName  string
}

// ReportDocumentMetricEvidence is the normalized evidence produced by a
// document parser adapter before it is evaluated by the pack guard.
type ReportDocumentMetricEvidence struct {
	CompanyName       string
	StockCode         string
	SelectionReason   string
	OfficialSource    string
	ReportPeriod      string
	Revenue           string
	RevenueGrowth     string
	NetProfit         string
	NetProfitGrowth   string
	OperatingCashFlow string
	ParserID          string
	SpecID            string
	ArtifactPath      string
	FieldEvidence     map[string]ReportDocumentMetricFieldEvidence `json:"field_evidence,omitempty"`
}

// ReportDocumentMetricFieldEvidence is source-neutral field-level provenance
// projected by document parser adapters. It intentionally avoids depending on
// docparse-specific types so OCR/table/vendor adapters can emit the same shape.
type ReportDocumentMetricFieldEvidence struct {
	Field           string                               `json:"field,omitempty"`
	Value           string                               `json:"value,omitempty"`
	Source          string                               `json:"source,omitempty"`
	Chapter         string                               `json:"chapter,omitempty"`
	Evidence        string                               `json:"evidence,omitempty"`
	Unit            string                               `json:"unit,omitempty"`
	Currency        string                               `json:"currency,omitempty"`
	Period          string                               `json:"period,omitempty"`
	PageRefs        []int                                `json:"page_refs,omitempty"`
	Confidence      float64                              `json:"confidence,omitempty"`
	ReviewRequired  bool                                 `json:"review_required,omitempty"`
	Warnings        []string                             `json:"warnings,omitempty"`
	SelectionReason string                               `json:"selection_reason,omitempty"`
	Candidates      []ReportDocumentMetricFieldCandidate `json:"candidates,omitempty"`
}

// ReportDocumentMetricFieldCandidate is a compact candidate projection for the
// field ranker/auditor. Adapters may omit it when the parser cannot expose
// losing candidates.
type ReportDocumentMetricFieldCandidate struct {
	Value           string  `json:"value,omitempty"`
	Source          string  `json:"source,omitempty"`
	Extractor       string  `json:"extractor,omitempty"`
	Evidence        string  `json:"evidence,omitempty"`
	Unit            string  `json:"unit,omitempty"`
	Currency        string  `json:"currency,omitempty"`
	Period          string  `json:"period,omitempty"`
	RowLabel        string  `json:"row_label,omitempty"`
	ColumnLabel     string  `json:"column_label,omitempty"`
	UnitSource      string  `json:"unit_source,omitempty"`
	PageRefs        []int   `json:"page_refs,omitempty"`
	LineNumber      int     `json:"line_number,omitempty"`
	Confidence      float64 `json:"confidence,omitempty"`
	Score           float64 `json:"score,omitempty"`
	Selected        bool    `json:"selected,omitempty"`
	SelectionReason string  `json:"selection_reason,omitempty"`
}

// ReportDocumentParser is the narrow adapter seam for turning local annual
// report artifacts into pack evidence. Implementations may use docparse, OCR,
// table extractors, or vendor APIs, but should not encode runtime routing.
type ReportDocumentParser interface {
	ParseReportDocument(context.Context, ReportDocumentParseInput) (ReportDocumentMetricEvidence, error)
}
