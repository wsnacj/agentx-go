package types

type FieldResult struct {
	Key             string           `json:"key"`
	Chapter         string           `json:"chapter,omitempty"`
	Value           interface{}      `json:"value,omitempty"`
	RawValue        interface{}      `json:"raw_value,omitempty"`
	NormalizedValue interface{}      `json:"normalized_value,omitempty"`
	Source          string           `json:"source,omitempty"` // regex|table|llm|script|derived
	Confidence      float64          `json:"confidence,omitempty"`
	Evidence        string           `json:"evidence,omitempty"`
	Unit            string           `json:"unit,omitempty"`
	Currency        string           `json:"currency,omitempty"`
	Period          string           `json:"period,omitempty"`
	PageRefs        []int            `json:"page_refs,omitempty"`
	BoundingBoxes   []BoundingBox    `json:"bounding_boxes,omitempty"`
	TableCells      []TableCellRef   `json:"table_cells,omitempty"`
	Warnings        []string         `json:"warnings,omitempty"`
	ReviewRequired  bool             `json:"review_required,omitempty"`
	SelectionReason string           `json:"selection_reason,omitempty"`
	Candidates      []FieldCandidate `json:"candidates,omitempty"`
}

type FieldCandidate struct {
	Chapter         string         `json:"chapter,omitempty"`
	Value           interface{}    `json:"value,omitempty"`
	RawValue        interface{}    `json:"raw_value,omitempty"`
	NormalizedValue interface{}    `json:"normalized_value,omitempty"`
	Source          string         `json:"source,omitempty"`
	Extractor       string         `json:"extractor,omitempty"`
	Confidence      float64        `json:"confidence,omitempty"`
	Score           float64        `json:"score,omitempty"`
	Evidence        string         `json:"evidence,omitempty"`
	Unit            string         `json:"unit,omitempty"`
	Currency        string         `json:"currency,omitempty"`
	Period          string         `json:"period,omitempty"`
	PageRefs        []int          `json:"page_refs,omitempty"`
	BoundingBoxes   []BoundingBox  `json:"bounding_boxes,omitempty"`
	TableCells      []TableCellRef `json:"table_cells,omitempty"`
	RowLabel        string         `json:"row_label,omitempty"`
	ColumnLabel     string         `json:"column_label,omitempty"`
	UnitSource      string         `json:"unit_source,omitempty"`
	LineNumber      int            `json:"line_number,omitempty"`
	Warnings        []string       `json:"warnings,omitempty"`
	Selected        bool           `json:"selected,omitempty"`
	SelectionReason string         `json:"selection_reason,omitempty"`
}

// BoundingBox identifies a source-provided rectangular region in a document page.
type BoundingBox struct {
	Page             int     `json:"page,omitempty"`
	X0               float64 `json:"x0"`
	Y0               float64 `json:"y0"`
	X1               float64 `json:"x1"`
	Y1               float64 `json:"y1"`
	Unit             string  `json:"unit,omitempty"`
	CoordinateSystem string  `json:"coordinate_system,omitempty"`
	Source           string  `json:"source,omitempty"`
}

// TableCellRef identifies a source-provided table cell or merged-cell range.
type TableCellRef struct {
	Page          int           `json:"page,omitempty"`
	TableIndex    int           `json:"table_index,omitempty"`
	StartRow      int           `json:"start_row,omitempty"`
	StartCol      int           `json:"start_col,omitempty"`
	EndRow        int           `json:"end_row,omitempty"`
	EndCol        int           `json:"end_col,omitempty"`
	BoundingBoxes []BoundingBox `json:"bounding_boxes,omitempty"`
	Source        string        `json:"source,omitempty"`
}

type ChapterResult struct {
	Key             string                 `json:"key"`
	TextSize        int                    `json:"text_size"`
	Fields          map[string]FieldResult `json:"fields"`
	RawLLM          string                 `json:"raw_llm,omitempty"`
	Prompt          string                 `json:"prompt,omitempty"`
	LLMRepairPrompt string                 `json:"llm_repair_prompt,omitempty"`
	LLMRepairRaw    string                 `json:"llm_repair_raw,omitempty"`
	LLMRepairReason string                 `json:"llm_repair_reason,omitempty"`
	LLMRepairFields []string               `json:"llm_repair_fields,omitempty"`
	Fallback        bool                   `json:"llm_fallback,omitempty"`
}

type DocumentResult struct {
	Chapters     map[string]*ChapterResult `json:"chapters"`
	ChapterOrder []string                  `json:"chapter_order"`
	OutputDir    string                    `json:"output_dir"`
	Fingerprint  *ParseFingerprint         `json:"fingerprint,omitempty"`
	Cache        *ParseCacheInfo           `json:"cache,omitempty"`
	Validations  []ValidationResult        `json:"validations,omitempty"`
	Diagnostics  *DocumentDiagnostics      `json:"diagnostics,omitempty"`
	// 派生字段诊断：未就绪/循环依赖等
	DerivedDiagnostics []DerivedDiagnostic `json:"derived_diagnostics,omitempty"`
}

type ParseFingerprint struct {
	Algorithm      string `json:"algorithm"`
	CacheKey       string `json:"cache_key"`
	DocumentSHA256 string `json:"document_sha256,omitempty"`
	SpecSHA256     string `json:"spec_sha256,omitempty"`
	SpecFileCount  int    `json:"spec_file_count,omitempty"`
	ModelName      string `json:"model_name,omitempty"`
	PDFParseMode   string `json:"pdf_parse_mode,omitempty"`
	ExtractionMode string `json:"extraction_mode,omitempty"`
	MaxChunkChars  int    `json:"max_chunk_chars,omitempty"`
	PageLimit      int    `json:"page_limit,omitempty"`
}

type ParseCacheInfo struct {
	Policy   string `json:"policy,omitempty"`
	Hit      bool   `json:"hit"`
	Written  bool   `json:"written,omitempty"`
	EntryDir string `json:"entry_dir,omitempty"`
}

type DocumentDiagnostics struct {
	Status      string                       `json:"status,omitempty"`
	PageCount   int                          `json:"page_count,omitempty"`
	TextQuality string                       `json:"text_quality,omitempty"`
	TextSource  string                       `json:"text_source,omitempty"`
	Stages      []StageDiagnostic            `json:"stages,omitempty"`
	Sections    map[string]SectionDiagnostic `json:"sections,omitempty"`
	Fields      map[string]FieldDiagnostic   `json:"fields,omitempty"`
	Warnings    []string                     `json:"warnings,omitempty"`
	Fallbacks   []string                     `json:"fallbacks,omitempty"`
}

type StageDiagnostic struct {
	Name       string   `json:"name"`
	Status     string   `json:"status"`
	DurationMS int64    `json:"duration_ms,omitempty"`
	Error      string   `json:"error,omitempty"`
	Warnings   []string `json:"warnings,omitempty"`
}

type SectionDiagnostic struct {
	Key            string  `json:"key"`
	Status         string  `json:"status"`
	MatchedBy      string  `json:"matched_by,omitempty"`
	PageCount      int     `json:"page_count,omitempty"`
	CandidatePages []int   `json:"candidate_pages,omitempty"`
	Confidence     float64 `json:"confidence,omitempty"`
	MissingReason  string  `json:"missing_reason,omitempty"`
}

type FieldDiagnostic struct {
	Chapter              string   `json:"chapter"`
	Field                string   `json:"field"`
	Status               string   `json:"status"`
	Source               string   `json:"source,omitempty"`
	MissingReason        string   `json:"missing_reason,omitempty"`
	NormalizationWarning string   `json:"normalization_warning,omitempty"`
	CandidateCount       int      `json:"candidate_count,omitempty"`
	CandidateValueCount  int      `json:"candidate_value_count,omitempty"`
	CandidateSources     []string `json:"candidate_sources,omitempty"`
	AgreementSourceCount int      `json:"agreement_source_count,omitempty"`
	SelectionReason      string   `json:"selection_reason,omitempty"`
	ReviewRequired       bool     `json:"review_required,omitempty"`
	Warnings             []string `json:"warnings,omitempty"`
}

type ValidationResult struct {
	Name     string `json:"name"`
	Passed   bool   `json:"passed"`
	Severity string `json:"severity"`
	Message  string `json:"message,omitempty"`
}

// DerivedDiagnostic 用于报告派生字段在分阶段求值后仍未就绪的原因
type DerivedDiagnostic struct {
	ID          string   `json:"id"` // chapter.field
	Chapter     string   `json:"chapter"`
	Field       string   `json:"field"`
	Formula     string   `json:"formula"`
	MissingDeps []string `json:"missing_deps,omitempty"` // 在原始抽取结果中未找到的基础依赖
	BlockedBy   []string `json:"blocked_by,omitempty"`   // 被哪些未计算的派生项阻塞
	Cycle       bool     `json:"cycle,omitempty"`        // 是否处于循环依赖环中
}
