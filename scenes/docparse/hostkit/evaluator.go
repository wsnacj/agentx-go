package hostkit

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	docparsetypes "github.com/wsnacj/agentx-go/document/pipeline/types"
)

const (
	failureClassParseFailed          = "parse_failed"
	failureClassFieldMissing         = "field_missing"
	failureClassWrongValue           = "wrong_value"
	failureClassWrongPeriod          = "wrong_period"
	failureClassWrongUnit            = "wrong_unit"
	failureClassDocumentTypeMismatch = "document_type_mismatch"
	failureClassEvidenceMissing      = "evidence_missing"
	failureClassBBoxMissing          = "bbox_missing"
	failureClassTableCellMissing     = "table_cell_missing"
	failureClassTableRowMissing      = "table_row_missing"
	failureClassReviewRequired       = "review_required"
	failureCodeResultMissing         = "docparse_result_missing"
	failureCodeResultReadFailed      = "docparse_result_read_failed"
	failureCodeResultDecodeFailed    = "docparse_result_decode_failed"
)

type evidencePayload struct {
	Tool                 string          `json:"tool"`
	TaskKind             string          `json:"task_kind,omitempty"`
	Source               string          `json:"source,omitempty"`
	Status               string          `json:"status"`
	AdapterStatus        string          `json:"adapter_status,omitempty"`
	ResultPath           string          `json:"result_path,omitempty"`
	DocumentPath         string          `json:"document_path,omitempty"`
	DocumentType         string          `json:"document_type,omitempty"`
	ExpectedDocumentType string          `json:"expected_document_type,omitempty"`
	ProfileID            string          `json:"profile_id,omitempty"`
	PageCount            int             `json:"page_count,omitempty"`
	FieldCount           int             `json:"field_count"`
	TableCount           int             `json:"table_count"`
	EvidenceComplete     bool            `json:"evidence_complete"`
	ReviewRequired       bool            `json:"review_required"`
	Passed               bool            `json:"passed"`
	FailureCode          string          `json:"failure_code,omitempty"`
	FailureClass         string          `json:"failure_class,omitempty"`
	Summary              string          `json:"summary,omitempty"`
	Fields               []fieldEvidence `json:"fields,omitempty"`
	Tables               []tableEvidence `json:"tables,omitempty"`
	ProfileProposal      map[string]any  `json:"profile_proposal,omitempty"`
	Diagnostics          map[string]any  `json:"diagnostics,omitempty"`
}

type fieldEvidence struct {
	Key             string                       `json:"key"`
	Chapter         string                       `json:"chapter,omitempty"`
	Value           any                          `json:"value,omitempty"`
	RawValue        any                          `json:"raw_value,omitempty"`
	NormalizedValue any                          `json:"normalized_value,omitempty"`
	Source          string                       `json:"source,omitempty"`
	Confidence      float64                      `json:"confidence,omitempty"`
	Evidence        string                       `json:"evidence,omitempty"`
	Unit            string                       `json:"unit,omitempty"`
	Currency        string                       `json:"currency,omitempty"`
	Period          string                       `json:"period,omitempty"`
	PageRefs        []int                        `json:"page_refs,omitempty"`
	BoundingBoxes   []docparsetypes.BoundingBox  `json:"bounding_boxes,omitempty"`
	TableCells      []docparsetypes.TableCellRef `json:"table_cells,omitempty"`
	CandidateCount  int                          `json:"candidate_count,omitempty"`
	ReviewRequired  bool                         `json:"review_required,omitempty"`
	Warnings        []string                     `json:"warnings,omitempty"`
}

type tableEvidence struct {
	Key      string                       `json:"key,omitempty"`
	PageRefs []int                        `json:"page_refs,omitempty"`
	Columns  []string                     `json:"columns,omitempty"`
	RowCount int                          `json:"row_count,omitempty"`
	Rows     []map[string]any             `json:"rows,omitempty"`
	CellRefs []docparsetypes.TableCellRef `json:"cell_refs,omitempty"`
	Warnings []string                     `json:"warnings,omitempty"`
}

type validationPayload struct {
	Tool                 string            `json:"tool"`
	Source               string            `json:"source,omitempty"`
	Status               string            `json:"status"`
	AdapterStatus        string            `json:"adapter_status,omitempty"`
	ResultPath           string            `json:"result_path,omitempty"`
	DocumentType         string            `json:"document_type,omitempty"`
	ExpectedDocumentType string            `json:"expected_document_type,omitempty"`
	ProfileID            string            `json:"profile_id,omitempty"`
	Passed               bool              `json:"passed"`
	EvidenceComplete     bool              `json:"evidence_complete"`
	ReviewRequired       bool              `json:"review_required"`
	FieldCount           int               `json:"field_count"`
	TableCount           int               `json:"table_count"`
	MissingFields        []string          `json:"missing_fields,omitempty"`
	Issues               []validationIssue `json:"issues,omitempty"`
	FailureCode          string            `json:"failure_code,omitempty"`
	FailureClass         string            `json:"failure_class,omitempty"`
	Summary              string            `json:"summary,omitempty"`
	Fields               []fieldEvidence   `json:"fields,omitempty"`
	Tables               []tableEvidence   `json:"tables,omitempty"`
	ProfileProposal      map[string]any    `json:"profile_proposal,omitempty"`
	Diagnostics          map[string]any    `json:"diagnostics,omitempty"`
}

type validationIssue struct {
	Field        string `json:"field,omitempty"`
	Code         string `json:"code"`
	FailureClass string `json:"failure_class"`
	Message      string `json:"message,omitempty"`
	Expected     any    `json:"expected,omitempty"`
	Actual       any    `json:"actual,omitempty"`
}

type evidenceRequirements struct {
	PageRefs          bool
	BoundingBoxes     bool
	TableCells        bool
	CompleteTableRows bool
}

func (k *Kit) buildEvidencePayload(tool string, params map[string]any) evidencePayload {
	raw, resultPath, loadErr := loadRawParseResult(params, k.cfg.ResultLoader)
	base := evidencePayload{
		Tool:                 tool,
		Source:               k.cfg.Source,
		AdapterStatus:        "ok",
		ResultPath:           resultPath,
		DocumentType:         firstStringArg(params, "document_type", "actual_document_type", "detected_document_type"),
		ExpectedDocumentType: firstStringArg(params, "expected_document_type"),
		ProfileID:            firstStringArg(params, "profile_id", "parser_profile_id"),
	}
	if documentPath := firstStringArg(params, "document_path"); documentPath != "" {
		base.DocumentPath = documentPath
	}
	if loadErr != nil {
		base.Status = "failed"
		base.AdapterStatus = "failed"
		base.Passed = false
		base.ReviewRequired = true
		base.FailureCode = loadErr.code
		base.FailureClass = loadErr.failureClass
		base.Summary = loadErr.message
		return base
	}
	base.ProfileProposal = profileProposalFromRaw(raw)
	if status := strings.TrimSpace(readString(raw, "status")); strings.EqualFold(status, "failed") {
		base.Status = "failed"
		base.AdapterStatus = "failed"
		base.Passed = false
		base.ReviewRequired = true
		base.FailureCode = firstNonEmptyString(raw["failure_code"], "docparse_parse_failed")
		base.FailureClass = firstNonEmptyString(raw["failure_class"], firstNonEmptyString(raw["error_class"], failureClassParseFailed))
		base.Summary = firstNonEmptyString(raw["error"], "docparse parse result reports failure")
		return base
	}
	fields := compactFieldsFromRaw(raw)
	if len(fields) == 0 {
		fields = documentResultFields(raw)
	}
	requirements := readEvidenceRequirements(params)
	tables := compactTablesFromRaw(raw, requirements.CompleteTableRows)
	if len(tables) == 0 {
		tables = tablesFromFieldCells(fields)
	}
	base.Status = "success"
	base.Fields = fields
	base.Tables = tables
	base.FieldCount = len(fields)
	base.TableCount = len(tables)
	diagnostics := readMap(raw, "diagnostics")
	if len(diagnostics) == 0 {
		if nested := readMap(raw, "result"); len(nested) > 0 {
			diagnostics = readMap(nested, "diagnostics")
		}
	}
	base.DocumentType = firstNonEmptyString(base.DocumentType, documentTypeFromRaw(raw, diagnostics))
	base.ProfileID = firstNonEmptyString(base.ProfileID, profileIDFromRaw(raw, diagnostics))
	base.PageCount = firstNonZeroInt(readInt(raw, "page_count"), readInt(diagnostics, "page_count"))
	base.Diagnostics = diagnostics
	base.ReviewRequired = readBool(raw, "review_required") || fieldsReviewRequired(fields)
	base.EvidenceComplete = evidenceComplete(fields, tables) && (!requirements.CompleteTableRows || completeTableRows(tables))
	base.Passed = base.EvidenceComplete && !base.ReviewRequired
	if applyDocumentTypeMismatchToEvidence(&base) {
		return base
	}
	if base.Passed {
		base.Summary = fmt.Sprintf("docparse evidence ready: fields=%d tables=%d", base.FieldCount, base.TableCount)
	} else if base.ReviewRequired {
		base.FailureClass = failureClassReviewRequired
		base.FailureCode = "docparse_review_required"
		base.Summary = "docparse evidence requires review"
	} else {
		if requirements.CompleteTableRows && !completeTableRows(tables) {
			base.FailureClass = failureClassTableRowMissing
			base.FailureCode = "docparse_table_rows_incomplete"
			base.Summary = "docparse table rows are incomplete"
		} else {
			base.FailureClass = failureClassEvidenceMissing
			base.FailureCode = "docparse_evidence_incomplete"
			base.Summary = "docparse evidence is incomplete"
		}
	}
	return base
}

func validateEvidencePayload(tool string, source string, params map[string]any, evidence evidencePayload) validationPayload {
	out := validationPayload{
		Tool:                 tool,
		Source:               source,
		AdapterStatus:        evidence.AdapterStatus,
		ResultPath:           evidence.ResultPath,
		DocumentType:         evidence.DocumentType,
		ExpectedDocumentType: evidence.ExpectedDocumentType,
		ProfileID:            evidence.ProfileID,
		FieldCount:           evidence.FieldCount,
		TableCount:           evidence.TableCount,
		EvidenceComplete:     evidence.EvidenceComplete,
		ReviewRequired:       evidence.ReviewRequired,
		Fields:               evidence.Fields,
		Tables:               evidence.Tables,
		ProfileProposal:      evidence.ProfileProposal,
		Diagnostics:          evidence.Diagnostics,
	}
	if evidence.Status != "success" {
		out.Status = "failed"
		out.Passed = false
		out.ReviewRequired = true
		out.FailureCode = evidence.FailureCode
		out.FailureClass = firstNonEmptyString(evidence.FailureClass, failureClassParseFailed)
		out.Summary = evidence.Summary
		return out
	}
	fieldByKey := map[string]fieldEvidence{}
	for _, field := range evidence.Fields {
		if field.Key == "" {
			continue
		}
		fieldByKey[field.Key] = field
	}
	requirements := readEvidenceRequirements(params)
	issues := []validationIssue{}
	missing := []string{}
	expectedDocumentType := firstNonEmptyString(evidence.ExpectedDocumentType, firstStringArg(params, "expected_document_type"))
	actualDocumentType := firstNonEmptyString(evidence.DocumentType, firstStringArg(params, "document_type", "actual_document_type", "detected_document_type"))
	if documentTypeMismatch(expectedDocumentType, actualDocumentType) {
		issues = append(issues, validationIssue{Code: "document_type_mismatch", FailureClass: failureClassDocumentTypeMismatch, Message: "document type mismatch", Expected: expectedDocumentType, Actual: actualDocumentType})
		out.ExpectedDocumentType = expectedDocumentType
		out.DocumentType = actualDocumentType
	}
	for _, key := range requiredFieldKeys(params) {
		field, ok := fieldByKey[key]
		if !ok {
			missing = append(missing, key)
			issues = append(issues, validationIssue{Field: key, Code: "field_missing", FailureClass: failureClassFieldMissing, Message: "required field missing"})
			continue
		}
		issues = append(issues, validateFieldEvidence(field, requirements)...)
	}
	for _, expected := range expectedFieldAssertions(params) {
		key := fieldAssertionKey(expected)
		if key == "" {
			continue
		}
		field, ok := fieldByKey[key]
		if !ok {
			missing = append(missing, key)
			issues = append(issues, validationIssue{Field: key, Code: "field_missing", FailureClass: failureClassFieldMissing, Message: "expected field missing"})
			continue
		}
		issues = append(issues, validateExpectedField(field, expected)...)
		issues = append(issues, validateFieldEvidence(field, mergeEvidenceRequirements(requirements, readEvidenceRequirements(expected)))...)
	}
	if requirements.CompleteTableRows {
		issues = append(issues, validateCompleteTableRows(evidence.Tables)...)
	}
	if evidence.ReviewRequired && !readBool(params, "allow_review_required") {
		issues = append(issues, validationIssue{Code: "review_required", FailureClass: failureClassReviewRequired, Message: "one or more fields require review"})
	}
	requireEvidenceComplete := readBool(params, "require_evidence_complete")
	if tool == ToolDocparseGuard {
		requireEvidenceComplete = true
	}
	if requireEvidenceComplete && !evidence.EvidenceComplete {
		issues = append(issues, validationIssue{Code: "evidence_incomplete", FailureClass: failureClassEvidenceMissing, Message: "evidence bundle is incomplete"})
	}
	out.MissingFields = uniqueSortedStrings(missing)
	out.Issues = dedupeIssues(issues)
	out.Passed = len(out.Issues) == 0
	if out.Passed {
		out.Status = "passed"
		out.Summary = fmt.Sprintf("docparse validation passed: fields=%d tables=%d", out.FieldCount, out.TableCount)
		return out
	}
	out.Status = "failed"
	if first := firstIssue(out.Issues); first != nil {
		out.FailureClass = first.FailureClass
		out.FailureCode = "docparse_" + first.Code
	}
	out.ReviewRequired = out.ReviewRequired || out.FailureClass == failureClassReviewRequired
	out.Summary = fmt.Sprintf("docparse validation failed: issues=%d", len(out.Issues))
	return out
}

func validateFieldEvidence(field fieldEvidence, req evidenceRequirements) []validationIssue {
	issues := []validationIssue{}
	if fieldValueMissing(field) || fieldHasWarning(field, "field_not_found") {
		issues = append(issues, validationIssue{Field: field.Key, Code: "field_missing", FailureClass: failureClassFieldMissing, Message: "field value missing"})
	}
	if req.PageRefs && len(field.PageRefs) == 0 {
		issues = append(issues, validationIssue{Field: field.Key, Code: "evidence_missing", FailureClass: failureClassEvidenceMissing, Message: "page_refs required"})
	}
	if req.BoundingBoxes && len(field.BoundingBoxes) == 0 {
		issues = append(issues, validationIssue{Field: field.Key, Code: "bbox_missing", FailureClass: failureClassBBoxMissing, Message: "bounding_boxes required"})
	}
	if req.TableCells && len(field.TableCells) == 0 {
		issues = append(issues, validationIssue{Field: field.Key, Code: "table_cell_missing", FailureClass: failureClassTableCellMissing, Message: "table_cells required"})
	}
	return issues
}

func validateExpectedField(field fieldEvidence, expected map[string]any) []validationIssue {
	issues := []validationIssue{}
	if expectedValue, ok := firstPresent(expected, "value", "expected_value", "normalized_value"); ok && !scalarEqual(fieldValue(field), expectedValue) {
		issues = append(issues, validationIssue{Field: field.Key, Code: "wrong_value", FailureClass: failureClassWrongValue, Message: "field value mismatch", Expected: expectedValue, Actual: fieldValue(field)})
	}
	if expectedPeriod := firstStringArg(expected, "period", "expected_period"); expectedPeriod != "" && strings.TrimSpace(field.Period) != expectedPeriod {
		issues = append(issues, validationIssue{Field: field.Key, Code: "wrong_period", FailureClass: failureClassWrongPeriod, Message: "field period mismatch", Expected: expectedPeriod, Actual: field.Period})
	}
	if expectedUnit := firstStringArg(expected, "unit", "expected_unit"); expectedUnit != "" && strings.TrimSpace(field.Unit) != expectedUnit {
		issues = append(issues, validationIssue{Field: field.Key, Code: "wrong_unit", FailureClass: failureClassWrongUnit, Message: "field unit mismatch", Expected: expectedUnit, Actual: field.Unit})
	}
	if expectedCurrency := firstStringArg(expected, "currency", "expected_currency"); expectedCurrency != "" && strings.TrimSpace(field.Currency) != expectedCurrency {
		issues = append(issues, validationIssue{Field: field.Key, Code: "wrong_unit", FailureClass: failureClassWrongUnit, Message: "field currency mismatch", Expected: expectedCurrency, Actual: field.Currency})
	}
	return issues
}

type resultLoadError struct {
	code         string
	failureClass string
	message      string
}

func loadRawParseResult(params map[string]any, loader ResultLoader) (map[string]any, string, *resultLoadError) {
	if raw, ok := params["parse_result"]; ok {
		out, err := anyToMap(raw)
		if err != nil {
			return nil, "", &resultLoadError{code: failureCodeResultDecodeFailed, failureClass: failureClassParseFailed, message: err.Error()}
		}
		return out, "", nil
	}
	resultPath := firstStringArg(params, "result_path")
	if resultPath == "" {
		return nil, "", &resultLoadError{code: failureCodeResultMissing, failureClass: failureClassParseFailed, message: "parse_result or result_path is required"}
	}
	if looksLikeRemoteRef(resultPath) {
		return nil, resultPath, &resultLoadError{code: failureCodeResultReadFailed, failureClass: failureClassParseFailed, message: "remote result_path is not supported"}
	}
	if loader == nil {
		return nil, resultPath, &resultLoadError{code: failureCodeResultReadFailed, failureClass: failureClassParseFailed, message: "result loader is not configured"}
	}
	blob, err := loader(resultPath)
	if err != nil {
		return nil, resultPath, &resultLoadError{code: failureCodeResultReadFailed, failureClass: failureClassParseFailed, message: err.Error()}
	}
	out := map[string]any{}
	if err := json.Unmarshal(blob, &out); err != nil {
		return nil, resultPath, &resultLoadError{code: failureCodeResultDecodeFailed, failureClass: failureClassParseFailed, message: err.Error()}
	}
	return out, resultPath, nil
}

func compactFieldsFromRaw(raw map[string]any) []fieldEvidence {
	fields := []fieldEvidence{}
	for _, item := range readObjectArray(raw["fields"]) {
		field := fieldEvidence{}
		_ = decodeAny(item, &field)
		field.Key = firstNonEmptyString(item["key"], firstNonEmptyString(item["field"], firstNonEmptyString(item["name"], field.Key)))
		if field.CandidateCount == 0 {
			field.CandidateCount = len(readObjectArray(item["candidates"]))
		}
		if field.Key != "" {
			fields = append(fields, field)
		}
	}
	sort.Slice(fields, func(i, j int) bool { return fields[i].Key < fields[j].Key })
	return fields
}

func compactTablesFromRaw(raw map[string]any, includeRows bool) []tableEvidence {
	tables := []tableEvidence{}
	for _, item := range readObjectArray(raw["tables"]) {
		table := tableEvidence{}
		_ = decodeAny(item, &table)
		table.Key = firstNonEmptyString(item["key"], firstNonEmptyString(item["name"], table.Key))
		if !includeRows {
			table.Rows = nil
		}
		tables = append(tables, table)
	}
	sort.Slice(tables, func(i, j int) bool { return tables[i].Key < tables[j].Key })
	return tables
}

func documentResultFields(raw map[string]any) []fieldEvidence {
	resultRaw := raw
	if nested := readMap(raw, "result"); len(nested) > 0 {
		resultRaw = nested
	}
	var result docparsetypes.DocumentResult
	if err := decodeAny(resultRaw, &result); err != nil {
		return nil
	}
	order := append([]string(nil), result.ChapterOrder...)
	if len(order) == 0 {
		for key := range result.Chapters {
			order = append(order, key)
		}
		sort.Strings(order)
	}
	out := []fieldEvidence{}
	for _, chapterKey := range order {
		chapter := result.Chapters[chapterKey]
		if chapter == nil {
			continue
		}
		fieldKeys := make([]string, 0, len(chapter.Fields))
		for key := range chapter.Fields {
			fieldKeys = append(fieldKeys, key)
		}
		sort.Strings(fieldKeys)
		for _, key := range fieldKeys {
			field := chapter.Fields[key]
			out = append(out, fieldEvidence{
				Key:             firstNonEmptyString(field.Key, key),
				Chapter:         firstNonEmptyString(field.Chapter, chapter.Key),
				Value:           field.Value,
				RawValue:        field.RawValue,
				NormalizedValue: field.NormalizedValue,
				Source:          field.Source,
				Confidence:      field.Confidence,
				Evidence:        field.Evidence,
				Unit:            field.Unit,
				Currency:        field.Currency,
				Period:          field.Period,
				PageRefs:        append([]int(nil), field.PageRefs...),
				BoundingBoxes:   append([]docparsetypes.BoundingBox(nil), field.BoundingBoxes...),
				TableCells:      append([]docparsetypes.TableCellRef(nil), field.TableCells...),
				CandidateCount:  len(field.Candidates),
				ReviewRequired:  field.ReviewRequired,
				Warnings:        append([]string(nil), field.Warnings...),
			})
		}
	}
	return out
}

func tablesFromFieldCells(fields []fieldEvidence) []tableEvidence {
	grouped := map[string]*tableEvidence{}
	for _, field := range fields {
		for _, cell := range field.TableCells {
			key := fmt.Sprintf("table_%d_page_%d", cell.TableIndex, cell.Page)
			table := grouped[key]
			if table == nil {
				table = &tableEvidence{Key: key, PageRefs: []int{cell.Page}}
				grouped[key] = table
			}
			table.CellRefs = append(table.CellRefs, cell)
			if cell.EndRow+1 > table.RowCount {
				table.RowCount = cell.EndRow + 1
			}
		}
	}
	keys := make([]string, 0, len(grouped))
	for key := range grouped {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	out := make([]tableEvidence, 0, len(keys))
	for _, key := range keys {
		out = append(out, *grouped[key])
	}
	return out
}

func evidenceComplete(fields []fieldEvidence, tables []tableEvidence) bool {
	if len(fields) == 0 && len(tables) == 0 {
		return false
	}
	for _, field := range fields {
		if fieldValueMissing(field) || fieldHasWarning(field, "field_not_found") {
			return false
		}
		if !fieldHasAnyEvidence(field) {
			return false
		}
	}
	for _, table := range tables {
		if len(table.PageRefs) == 0 && len(table.CellRefs) == 0 {
			return false
		}
	}
	return true
}

func completeTableRows(tables []tableEvidence) bool {
	if len(tables) == 0 {
		return false
	}
	for _, table := range tables {
		if table.RowCount <= 0 || len(table.Rows) != table.RowCount || len(table.Columns) == 0 {
			return false
		}
		for _, row := range table.Rows {
			for _, column := range table.Columns {
				if _, ok := row[column]; !ok {
					return false
				}
			}
		}
	}
	return true
}

func validateCompleteTableRows(tables []tableEvidence) []validationIssue {
	issues := []validationIssue{}
	if len(tables) == 0 {
		return append(issues, validationIssue{Code: "table_rows_missing", FailureClass: failureClassTableRowMissing, Message: "complete table rows required"})
	}
	for _, table := range tables {
		if table.RowCount <= 0 {
			issues = append(issues, validationIssue{Field: table.Key, Code: "table_row_count_missing", FailureClass: failureClassTableRowMissing, Message: "positive table row_count required"})
			continue
		}
		if len(table.Rows) != table.RowCount {
			issues = append(issues, validationIssue{Field: table.Key, Code: "table_rows_incomplete", FailureClass: failureClassTableRowMissing, Message: "table rows must match row_count", Expected: table.RowCount, Actual: len(table.Rows)})
			continue
		}
		for rowIndex, row := range table.Rows {
			for _, column := range table.Columns {
				if _, ok := row[column]; !ok {
					issues = append(issues, validationIssue{Field: table.Key, Code: "table_row_column_missing", FailureClass: failureClassTableRowMissing, Message: fmt.Sprintf("row %d missing column %s", rowIndex, column), Expected: column})
				}
			}
		}
	}
	return issues
}

func fieldHasAnyEvidence(field fieldEvidence) bool {
	return strings.TrimSpace(field.Evidence) != "" || len(field.PageRefs) > 0 || len(field.BoundingBoxes) > 0 || len(field.TableCells) > 0
}

func fieldsReviewRequired(fields []fieldEvidence) bool {
	for _, field := range fields {
		if field.ReviewRequired || fieldHasWarning(field, "field_not_found") {
			return true
		}
	}
	return false
}

func applyDocumentTypeMismatchToEvidence(payload *evidencePayload) bool {
	if payload == nil || !documentTypeMismatch(payload.ExpectedDocumentType, payload.DocumentType) {
		return false
	}
	payload.Passed = false
	payload.ReviewRequired = true
	payload.FailureClass = failureClassDocumentTypeMismatch
	payload.FailureCode = "docparse_document_type_mismatch"
	payload.Summary = fmt.Sprintf("docparse document type mismatch: expected=%s actual=%s", payload.ExpectedDocumentType, payload.DocumentType)
	return true
}

func documentTypeFromRaw(raw map[string]any, diagnostics map[string]any) string {
	for _, values := range []map[string]any{
		raw,
		readMap(raw, "document_ref"),
		readMap(raw, "document"),
		readMap(raw, "metadata"),
		diagnostics,
	} {
		if value := firstStringArg(values, "document_type", "doc_type", "actual_document_type", "detected_document_type", "profile_document_type"); value != "" {
			return value
		}
	}
	if nested := readMap(raw, "result"); len(nested) > 0 {
		return documentTypeFromRaw(nested, readMap(nested, "diagnostics"))
	}
	return ""
}

func profileIDFromRaw(raw map[string]any, diagnostics map[string]any) string {
	for _, values := range []map[string]any{
		raw,
		readMap(raw, "document_ref"),
		readMap(raw, "metadata"),
		diagnostics,
	} {
		if value := firstStringArg(values, "profile_id", "parser_profile_id", "spec_id"); value != "" {
			return value
		}
	}
	if nested := readMap(raw, "result"); len(nested) > 0 {
		return profileIDFromRaw(nested, readMap(nested, "diagnostics"))
	}
	return ""
}

func profileProposalFromRaw(raw map[string]any) map[string]any {
	if proposal := readMap(raw, "profile_proposal"); len(proposal) > 0 {
		return proposal
	}
	if diagnostics := readMap(raw, "diagnostics"); len(diagnostics) > 0 {
		if proposal := readMap(diagnostics, "profile_proposal"); len(proposal) > 0 {
			return proposal
		}
	}
	if nested := readMap(raw, "result"); len(nested) > 0 {
		return profileProposalFromRaw(nested)
	}
	return nil
}

func documentTypeMismatch(expected string, actual string) bool {
	expected = normalizeDocumentType(expected)
	actual = normalizeDocumentType(actual)
	return expected != "" && actual != "" && expected != actual
}

func normalizeDocumentType(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return ""
	}
	var b strings.Builder
	underscore := false
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
			underscore = false
		case r >= '0' && r <= '9':
			b.WriteRune(r)
			underscore = false
		default:
			if !underscore && b.Len() > 0 {
				b.WriteByte('_')
				underscore = true
			}
		}
	}
	return strings.Trim(b.String(), "_")
}

func fieldValueMissing(field fieldEvidence) bool {
	value := fieldValue(field)
	if value == nil {
		return true
	}
	if text, ok := value.(string); ok && strings.TrimSpace(text) == "" {
		return true
	}
	return false
}

func fieldHasWarning(field fieldEvidence, warning string) bool {
	warning = strings.TrimSpace(warning)
	if warning == "" {
		return false
	}
	for _, item := range field.Warnings {
		if strings.EqualFold(strings.TrimSpace(item), warning) {
			return true
		}
	}
	return false
}

func requiredFieldKeys(params map[string]any) []string {
	out := readStringArray(params["required_fields"])
	sort.Strings(out)
	return uniqueSortedStrings(out)
}

func expectedFieldAssertions(params map[string]any) []map[string]any {
	return readObjectArray(params["expected_fields"])
}

func fieldAssertionKey(item map[string]any) string {
	return firstStringArg(item, "key", "field", "name")
}

func readEvidenceRequirements(params map[string]any) evidenceRequirements {
	req := evidenceRequirements{
		PageRefs:          readBool(params, "require_page_refs"),
		BoundingBoxes:     readBool(params, "require_bounding_boxes", "require_bbox"),
		TableCells:        readBool(params, "require_table_cells"),
		CompleteTableRows: readBool(params, "require_complete_table_rows"),
	}
	nested := readMap(params, "required_evidence")
	if len(nested) > 0 {
		req = mergeEvidenceRequirements(req, evidenceRequirements{
			PageRefs:          readBool(nested, "page_refs", "require_page_refs"),
			BoundingBoxes:     readBool(nested, "bounding_boxes", "bbox", "require_bounding_boxes", "require_bbox"),
			TableCells:        readBool(nested, "table_cells", "require_table_cells"),
			CompleteTableRows: readBool(nested, "complete_table_rows", "require_complete_table_rows"),
		})
	}
	return req
}

func mergeEvidenceRequirements(base evidenceRequirements, overlay evidenceRequirements) evidenceRequirements {
	return evidenceRequirements{
		PageRefs:          base.PageRefs || overlay.PageRefs,
		BoundingBoxes:     base.BoundingBoxes || overlay.BoundingBoxes,
		TableCells:        base.TableCells || overlay.TableCells,
		CompleteTableRows: base.CompleteTableRows || overlay.CompleteTableRows,
	}
}

func fieldValue(field fieldEvidence) any {
	if field.NormalizedValue != nil {
		return field.NormalizedValue
	}
	if field.Value != nil {
		return field.Value
	}
	return field.RawValue
}

func scalarEqual(actual any, expected any) bool {
	if actual == nil || expected == nil {
		return actual == expected
	}
	if actualNumber, ok := numericValue(actual); ok {
		if expectedNumber, ok := numericValue(expected); ok {
			return actualNumber == expectedNumber
		}
	}
	return strings.TrimSpace(fmt.Sprint(actual)) == strings.TrimSpace(fmt.Sprint(expected))
}

func numericValue(value any) (float64, bool) {
	switch typed := value.(type) {
	case int:
		return float64(typed), true
	case int64:
		return float64(typed), true
	case float64:
		return typed, true
	case float32:
		return float64(typed), true
	case json.Number:
		number, err := typed.Float64()
		return number, err == nil
	default:
		return 0, false
	}
}

func anyToMap(value any) (map[string]any, error) {
	if typed, ok := value.(map[string]any); ok {
		return typed, nil
	}
	if text, ok := value.(string); ok {
		out := map[string]any{}
		if err := json.Unmarshal([]byte(text), &out); err != nil {
			return nil, err
		}
		return out, nil
	}
	out := map[string]any{}
	if err := decodeAny(value, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func readObjectArray(raw any) []map[string]any {
	switch typed := raw.(type) {
	case []map[string]any:
		return typed
	case []any:
		out := make([]map[string]any, 0, len(typed))
		for _, item := range typed {
			if m, err := anyToMap(item); err == nil {
				out = append(out, m)
			}
		}
		return out
	default:
		return nil
	}
}

func readStringArray(raw any) []string {
	switch typed := raw.(type) {
	case []string:
		return append([]string(nil), typed...)
	case []any:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			if text := strings.TrimSpace(fmt.Sprint(item)); text != "" {
				out = append(out, text)
			}
		}
		return out
	default:
		return nil
	}
}

func readMap(values map[string]any, key string) map[string]any {
	if values == nil {
		return nil
	}
	out, _ := anyToMap(values[key])
	return out
}

func readString(values map[string]any, key string) string {
	if values == nil {
		return ""
	}
	return firstNonEmptyString(values[key], "")
}

func readInt(values map[string]any, keys ...string) int {
	for _, key := range keys {
		switch typed := values[key].(type) {
		case int:
			return typed
		case int64:
			return int(typed)
		case float64:
			return int(typed)
		case json.Number:
			number, err := typed.Int64()
			if err == nil {
				return int(number)
			}
		}
	}
	return 0
}

func firstNonZeroInt(values ...int) int {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}
	return 0
}

func readBool(values map[string]any, keys ...string) bool {
	for _, key := range keys {
		switch typed := values[key].(type) {
		case bool:
			return typed
		case string:
			return strings.EqualFold(strings.TrimSpace(typed), "true")
		}
	}
	return false
}

func firstPresent(values map[string]any, keys ...string) (any, bool) {
	for _, key := range keys {
		if value, ok := values[key]; ok {
			return value, true
		}
	}
	return nil, false
}

func decodeAny(value any, out any) error {
	blob, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return json.Unmarshal(blob, out)
}

func uniqueSortedStrings(values []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func dedupeIssues(values []validationIssue) []validationIssue {
	seen := map[string]bool{}
	out := make([]validationIssue, 0, len(values))
	for _, issue := range values {
		key := issue.Field + "|" + issue.Code + "|" + issue.FailureClass
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, issue)
	}
	return out
}

func firstIssue(values []validationIssue) *validationIssue {
	if len(values) == 0 {
		return nil
	}
	return &values[0]
}

func looksLikeRemoteRef(value string) bool {
	trimmed := strings.ToLower(strings.TrimSpace(value))
	return strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://")
}
