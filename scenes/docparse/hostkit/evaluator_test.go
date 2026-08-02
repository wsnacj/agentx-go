package hostkit

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func newTestKit() *Kit {
	return New(Config{ResultLoader: os.ReadFile})
}

func TestTraceEvidenceFromSyntheticResultPath(t *testing.T) {
	kit := newTestKit()

	evidence := kit.buildEvidencePayload(ToolDocparseTraceEvidence, map[string]any{
		"result_path": syntheticResultPath(t),
	})

	if evidence.Status != "success" || !evidence.Passed || !evidence.EvidenceComplete {
		t.Fatalf("unexpected evidence status: %#v", evidence)
	}
	if evidence.FieldCount != 2 || evidence.TableCount != 1 || evidence.PageCount != 1 {
		t.Fatalf("unexpected evidence counts: fields=%d tables=%d pages=%d", evidence.FieldCount, evidence.TableCount, evidence.PageCount)
	}
	invoiceID, ok := fieldByKey(evidence.Fields, "invoice_id")
	if !ok || len(invoiceID.PageRefs) != 1 || len(invoiceID.BoundingBoxes) != 1 {
		t.Fatalf("invoice_id evidence not projected: %#v", invoiceID)
	}
	totalAmount, ok := fieldByKey(evidence.Fields, "total_amount")
	if !ok || len(totalAmount.TableCells) != 1 {
		t.Fatalf("total_amount table evidence not projected: %#v", totalAmount)
	}
}

func TestExtractFieldsProjectsRequestedFieldsFromSyntheticResult(t *testing.T) {
	kit := newTestKit()

	out, err := kit.ExtractFields(context.Background(), map[string]any{
		"result_path":      syntheticResultPath(t),
		"requested_fields": []string{"invoice_id"},
	})
	if err != nil {
		t.Fatalf("ExtractFields returned error: %v", err)
	}
	payload, ok := out.(evidencePayload)
	if !ok {
		t.Fatalf("ExtractFields returned %T, want evidencePayload", out)
	}
	if payload.Tool != ToolDocparseExtractFields || payload.TaskKind != "document.extract_fields" {
		t.Fatalf("unexpected extraction identity: %#v", payload)
	}
	if payload.Status != "success" || !payload.Passed || payload.FieldCount != 1 || payload.TableCount != 0 {
		t.Fatalf("unexpected field extraction payload: %#v", payload)
	}
	field, ok := fieldByKey(payload.Fields, "invoice_id")
	if !ok || len(field.BoundingBoxes) != 1 {
		t.Fatalf("invoice_id field evidence not projected: %#v", payload.Fields)
	}
}

func TestExtractFieldsMarksMissingRequestedFieldsForReview(t *testing.T) {
	kit := newTestKit()

	out, err := kit.ExtractFields(context.Background(), map[string]any{
		"result_path":      syntheticResultPath(t),
		"requested_fields": []string{"invoice_id", "missing_tax_number"},
	})
	if err != nil {
		t.Fatalf("ExtractFields returned error: %v", err)
	}
	payload, ok := out.(evidencePayload)
	if !ok {
		t.Fatalf("ExtractFields returned %T, want evidencePayload", out)
	}
	if payload.Status != "success" || payload.Passed || !payload.ReviewRequired || payload.FailureClass != failureClassReviewRequired {
		t.Fatalf("expected missing requested field to require review, got %#v", payload)
	}
	missing, ok := fieldByKey(payload.Fields, "missing_tax_number")
	if !ok || !missing.ReviewRequired || len(missing.Warnings) == 0 || missing.Warnings[0] != "field_not_found" {
		t.Fatalf("missing requested field was not surfaced: %#v", payload.Fields)
	}
}

func TestExtractFieldsPreservesTopLevelReviewRequired(t *testing.T) {
	kit := newTestKit()

	out, err := kit.ExtractFields(context.Background(), map[string]any{
		"parse_result": map[string]any{
			"review_required": true,
			"fields": []map[string]any{
				{
					"key":       "amount",
					"value":     "10",
					"evidence":  "Amount 10",
					"page_refs": []int{1},
				},
			},
		},
		"requested_fields": []string{"amount"},
	})
	if err != nil {
		t.Fatalf("ExtractFields returned error: %v", err)
	}
	payload, ok := out.(evidencePayload)
	if !ok {
		t.Fatalf("ExtractFields returned %T, want evidencePayload", out)
	}
	if !payload.ReviewRequired || payload.Passed || payload.FailureClass != failureClassReviewRequired {
		t.Fatalf("expected top-level review_required to be preserved, got %#v", payload)
	}
}

func TestExtractFieldsPreservesProfileProposal(t *testing.T) {
	kit := newTestKit()

	out, err := kit.ExtractFields(context.Background(), map[string]any{
		"parse_result": map[string]any{
			"review_required": true,
			"profile_proposal": map[string]any{
				"candidate_document_types": []map[string]any{{
					"type": "alpha_import_declaration",
				}},
				"discovery_results": []map[string]any{{
					"title": "Official Alpha Import Declaration Guide",
				}},
				"review_required": true,
			},
		},
	})
	if err != nil {
		t.Fatalf("ExtractFields returned error: %v", err)
	}
	payload, ok := out.(evidencePayload)
	if !ok {
		t.Fatalf("ExtractFields returned %T, want evidencePayload", out)
	}
	if !payload.ReviewRequired || payload.Passed {
		t.Fatalf("expected proposal-only extraction to require review, got %#v", payload)
	}
	if len(payload.ProfileProposal) == 0 {
		t.Fatalf("profile proposal was not preserved: %#v", payload)
	}
	if got := payload.ProfileProposal["review_required"]; got != true {
		t.Fatalf("profile proposal review_required = %#v, want true", got)
	}
}

func TestProfileProbeReturnsReviewRequiredProposalOnly(t *testing.T) {
	kit := newTestKit()

	out, err := kit.ProfileProbe(context.Background(), map[string]any{
		"parse_result": map[string]any{
			"review_required": true,
			"profile_proposal": map[string]any{
				"candidate_document_types": []map[string]any{{"type": "alpha_import_declaration"}},
				"suggested_fields":         []string{"declaration_number", "issuer"},
				"review_required":          true,
			},
			"fields": []map[string]any{{"key": "declaration_number", "value": "A-001", "page_refs": []int{1}}},
		},
	})
	if err != nil {
		t.Fatalf("ProfileProbe returned error: %v", err)
	}
	payload, ok := out.(evidencePayload)
	if !ok {
		t.Fatalf("ProfileProbe returned %T, want evidencePayload", out)
	}
	if payload.Tool != ToolDocparseProfileProbe || payload.TaskKind != "document.profile_probe" {
		t.Fatalf("unexpected profile probe identity: %#v", payload)
	}
	if !payload.ReviewRequired || payload.Passed || payload.EvidenceComplete {
		t.Fatalf("profile probe should remain review-required and not passed: %#v", payload)
	}
	if len(payload.Fields) != 0 || len(payload.Tables) != 0 || payload.FieldCount != 0 || payload.TableCount != 0 {
		t.Fatalf("profile probe should not expose final fields/tables: %#v", payload)
	}
	if len(payload.ProfileProposal) == 0 {
		t.Fatalf("profile proposal was not preserved: %#v", payload)
	}
}

func TestExtractFieldsTreatsFieldNotFoundWarningAsIncomplete(t *testing.T) {
	kit := newTestKit()

	out, err := kit.ExtractFields(context.Background(), map[string]any{
		"parse_result": map[string]any{
			"fields": []map[string]any{
				{
					"key":       "invoice_date",
					"page_refs": []int{1},
					"warnings":  []string{"field_not_found", "bbox_not_available"},
				},
			},
		},
		"requested_fields":  []string{"invoice_date"},
		"require_page_refs": true,
	})
	if err != nil {
		t.Fatalf("ExtractFields returned error: %v", err)
	}
	payload, ok := out.(evidencePayload)
	if !ok {
		t.Fatalf("ExtractFields returned %T, want evidencePayload", out)
	}
	if payload.Status != "success" || payload.Passed || payload.EvidenceComplete || !payload.ReviewRequired {
		t.Fatalf("expected field_not_found to fail readiness, got %#v", payload)
	}
	if payload.FailureClass != failureClassReviewRequired {
		t.Fatalf("failure class = %q, want %q", payload.FailureClass, failureClassReviewRequired)
	}
	validation := validateEvidencePayload(ToolDocparseValidate, DefaultSource, map[string]any{
		"parse_result":      map[string]any{},
		"required_fields":   []string{"invoice_date"},
		"require_page_refs": true,
	}, payload)
	if validation.Passed || validation.FailureClass != failureClassFieldMissing {
		t.Fatalf("expected validation to report field_missing, got %#v", validation)
	}
}

func TestExtractFieldsMarksDocumentTypeMismatch(t *testing.T) {
	kit := newTestKit()

	out, err := kit.ExtractFields(context.Background(), map[string]any{
		"parse_result": map[string]any{
			"document_type": "air_waybill",
			"fields": []map[string]any{
				{
					"key":       "awb_number",
					"value":     "123-45678901",
					"evidence":  "AWB 123-45678901",
					"page_refs": []int{1},
				},
			},
		},
		"expected_document_type": "invoice",
		"requested_fields":       []string{"awb_number"},
	})
	if err != nil {
		t.Fatalf("ExtractFields returned error: %v", err)
	}
	payload, ok := out.(evidencePayload)
	if !ok {
		t.Fatalf("ExtractFields returned %T, want evidencePayload", out)
	}
	if payload.Status != "success" || payload.Passed || !payload.ReviewRequired {
		t.Fatalf("expected document type mismatch to block readiness, got %#v", payload)
	}
	if payload.FailureClass != failureClassDocumentTypeMismatch || payload.FailureCode != "docparse_document_type_mismatch" {
		t.Fatalf("unexpected mismatch failure: class=%q code=%q", payload.FailureClass, payload.FailureCode)
	}
	if payload.DocumentType != "air_waybill" || payload.ExpectedDocumentType != "invoice" {
		t.Fatalf("document type metadata not preserved: %#v", payload)
	}
}

func TestExtractTableProjectsTablesFromSyntheticResult(t *testing.T) {
	kit := newTestKit()

	out, err := kit.ExtractTable(context.Background(), map[string]any{
		"result_path": syntheticResultPath(t),
	})
	if err != nil {
		t.Fatalf("ExtractTable returned error: %v", err)
	}
	payload, ok := out.(evidencePayload)
	if !ok {
		t.Fatalf("ExtractTable returned %T, want evidencePayload", out)
	}
	if payload.Tool != ToolDocparseExtractTable || payload.TaskKind != "document.extract_table" {
		t.Fatalf("unexpected extraction identity: %#v", payload)
	}
	if payload.Status != "success" || !payload.Passed || payload.FieldCount != 0 || payload.TableCount != 1 {
		t.Fatalf("unexpected table extraction payload: %#v", payload)
	}
	if len(payload.Fields) != 0 || len(payload.Tables) != 1 || len(payload.Tables[0].CellRefs) != 1 {
		t.Fatalf("table evidence not projected cleanly: %#v", payload)
	}
}

func TestExtractTableProjectsCompleteRowsOnlyWhenRequested(t *testing.T) {
	kit := newTestKit()
	parseResult := map[string]any{
		"tables": []map[string]any{{
			"key":       "transactions",
			"page_refs": []int{1},
			"columns":   []string{"date", "amount"},
			"row_count": 2,
			"rows": []map[string]any{
				{"date": "2026-05-01", "amount": "10.00"},
				{"date": "2026-05-02", "amount": "20.00"},
			},
		}},
	}

	compactOut, err := kit.ExtractTable(context.Background(), map[string]any{"parse_result": parseResult})
	if err != nil {
		t.Fatalf("ExtractTable compact returned error: %v", err)
	}
	compact := compactOut.(evidencePayload)
	if len(compact.Tables) != 1 || len(compact.Tables[0].Rows) != 0 {
		t.Fatalf("default projection must remain compact: %#v", compact.Tables)
	}

	completeOut, err := kit.ExtractTable(context.Background(), map[string]any{
		"parse_result":                parseResult,
		"require_complete_table_rows": true,
	})
	if err != nil {
		t.Fatalf("ExtractTable complete returned error: %v", err)
	}
	complete := completeOut.(evidencePayload)
	if !complete.Passed || !complete.EvidenceComplete || len(complete.Tables) != 1 || len(complete.Tables[0].Rows) != 2 {
		t.Fatalf("complete rows were not projected: %#v", complete)
	}
}

func TestExtractTableRejectsIncompleteRequestedRows(t *testing.T) {
	kit := newTestKit()
	out, err := kit.ExtractTable(context.Background(), map[string]any{
		"require_complete_table_rows": true,
		"parse_result": map[string]any{
			"tables": []map[string]any{{
				"key":       "transactions",
				"page_refs": []int{1},
				"columns":   []string{"date", "amount"},
				"row_count": 2,
				"rows":      []map[string]any{{"date": "2026-05-01", "amount": "10.00"}},
			}},
		},
	})
	if err != nil {
		t.Fatalf("ExtractTable returned error: %v", err)
	}
	payload := out.(evidencePayload)
	if payload.Passed || payload.EvidenceComplete || payload.FailureClass != failureClassTableRowMissing || payload.FailureCode != "docparse_table_rows_incomplete" {
		t.Fatalf("incomplete rows must fail closed: %#v", payload)
	}
}

func TestValidateReportsDocumentTypeMismatch(t *testing.T) {
	kit := newTestKit()
	params := map[string]any{
		"parse_result": map[string]any{
			"document_ref": map[string]any{"document_type": "bank_statement"},
			"fields": []map[string]any{
				{
					"key":       "transaction_date",
					"value":     "2026-05-25",
					"evidence":  "2026-05-25 payroll",
					"page_refs": []int{1},
				},
			},
		},
		"expected_document_type": "invoice",
		"expected_fields": []map[string]any{
			{"key": "transaction_date", "value": "2026-05-25"},
		},
	}
	evidence := kit.buildEvidencePayload(ToolDocparseTraceEvidence, params)

	validation := validateEvidencePayload(ToolDocparseValidate, DefaultSource, params, evidence)

	if validation.Status != "failed" || validation.Passed {
		t.Fatalf("expected validation to fail, got %#v", validation)
	}
	if validation.FailureClass != failureClassDocumentTypeMismatch {
		t.Fatalf("failure class = %q, want %q", validation.FailureClass, failureClassDocumentTypeMismatch)
	}
	if !issueExists(validation.Issues, "", "document_type_mismatch") {
		t.Fatalf("expected document_type_mismatch issue, got %#v", validation.Issues)
	}
}

func TestValidateExpectedFieldsPasses(t *testing.T) {
	kit := newTestKit()
	params := map[string]any{
		"result_path":     syntheticResultPath(t),
		"required_fields": []string{"invoice_id", "total_amount"},
		"expected_fields": []map[string]any{
			{
				"key":   "invoice_id",
				"value": "INV-001",
				"required_evidence": map[string]any{
					"page_refs":      true,
					"bounding_boxes": true,
				},
			},
			{
				"key":    "total_amount",
				"value":  "123.45",
				"period": "2026-05",
				"unit":   "USD",
				"required_evidence": map[string]any{
					"page_refs":   true,
					"table_cells": true,
				},
			},
		},
	}
	evidence := kit.buildEvidencePayload(ToolDocparseTraceEvidence, params)

	validation := validateEvidencePayload(ToolDocparseValidate, DefaultSource, params, evidence)

	if validation.Status != "passed" || !validation.Passed {
		t.Fatalf("expected validation to pass, got %#v", validation)
	}
	if len(validation.Issues) != 0 || len(validation.MissingFields) != 0 {
		t.Fatalf("unexpected validation issues: %#v missing=%#v", validation.Issues, validation.MissingFields)
	}
}

func TestValidateReportsWrongUnitAndMissingField(t *testing.T) {
	kit := newTestKit()
	params := map[string]any{
		"result_path": syntheticResultPath(t),
		"expected_fields": []map[string]any{
			{"key": "total_amount", "unit": "EUR"},
			{"key": "customer_id", "value": "C-001"},
		},
	}
	evidence := kit.buildEvidencePayload(ToolDocparseTraceEvidence, params)

	validation := validateEvidencePayload(ToolDocparseValidate, DefaultSource, params, evidence)

	if validation.Status != "failed" || validation.Passed {
		t.Fatalf("expected validation to fail, got %#v", validation)
	}
	if validation.FailureClass != failureClassWrongUnit {
		t.Fatalf("failure class = %q, want %q", validation.FailureClass, failureClassWrongUnit)
	}
	if !issueExists(validation.Issues, "total_amount", "wrong_unit") {
		t.Fatalf("expected wrong_unit issue, got %#v", validation.Issues)
	}
	if !issueExists(validation.Issues, "customer_id", "field_missing") {
		t.Fatalf("expected field_missing issue, got %#v", validation.Issues)
	}
}

func TestGuardReportsReviewRequired(t *testing.T) {
	kit := newTestKit()

	out, err := kit.Guard(context.Background(), map[string]any{
		"parse_result": map[string]any{
			"fields": []map[string]any{
				{
					"key":             "invoice_id",
					"value":           "INV-001",
					"page_refs":       []int{1},
					"review_required": true,
				},
			},
		},
		"required_fields": []string{"invoice_id"},
	})
	if err != nil {
		t.Fatalf("Guard returned error: %v", err)
	}
	guard, ok := out.(validationPayload)
	if !ok {
		t.Fatalf("Guard returned %T, want validationPayload", out)
	}
	if guard.Status != "review_required" || !guard.ReviewRequired || guard.FailureClass != failureClassReviewRequired {
		t.Fatalf("unexpected guard payload: %#v", guard)
	}
}

func TestGuardRejectsEvidenceIncomplete(t *testing.T) {
	kit := newTestKit()

	out, err := kit.Guard(context.Background(), map[string]any{
		"parse_result": map[string]any{
			"fields": []map[string]any{
				{"key": "invoice_id", "value": "INV-001"},
			},
		},
		"expected_fields": []map[string]any{
			{"key": "invoice_id", "value": "INV-001"},
		},
	})
	if err != nil {
		t.Fatalf("Guard returned error: %v", err)
	}
	guard, ok := out.(validationPayload)
	if !ok {
		t.Fatalf("Guard returned %T, want validationPayload", out)
	}
	if guard.Status != "failed" || guard.Passed || guard.FailureClass != failureClassEvidenceMissing {
		t.Fatalf("expected incomplete evidence to fail guard, got %#v", guard)
	}
	if guard.FailureCode != "docparse_evidence_incomplete" {
		t.Fatalf("failure code = %q, want docparse_evidence_incomplete", guard.FailureCode)
	}
}

func TestValidateAllowsIncompleteOptionalEvidenceUnlessRequired(t *testing.T) {
	kit := newTestKit()

	out, err := kit.Validate(context.Background(), map[string]any{
		"parse_result": map[string]any{
			"review_required": true,
			"fields": []map[string]any{
				{
					"key":      "optional_tax_id",
					"warnings": []string{"field_not_found", "bbox_not_available"},
				},
			},
		},
		"allow_review_required": true,
	})
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
	payload, ok := out.(validationPayload)
	if !ok {
		t.Fatalf("Validate returned %T, want validationPayload", out)
	}
	if !payload.Passed || payload.Status != "passed" || !payload.ReviewRequired || payload.EvidenceComplete {
		t.Fatalf("expected advisory validation to pass with review_required optional missing field, got %#v", payload)
	}

	out, err = kit.Validate(context.Background(), map[string]any{
		"parse_result": map[string]any{
			"fields": []map[string]any{
				{"key": "invoice_id", "value": "INV-001"},
			},
		},
		"expected_fields": []map[string]any{
			{"key": "invoice_id", "value": "INV-001"},
		},
		"require_evidence_complete": true,
	})
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
	payload, ok = out.(validationPayload)
	if !ok {
		t.Fatalf("Validate returned %T, want validationPayload", out)
	}
	if payload.Passed || payload.FailureClass != failureClassEvidenceMissing {
		t.Fatalf("expected required evidence completeness to fail, got %#v", payload)
	}
}

func syntheticResultPath(t *testing.T) string {
	t.Helper()
	return filepath.Join("testdata", "synthetic_result.json")
}

func fieldByKey(fields []fieldEvidence, key string) (fieldEvidence, bool) {
	for _, field := range fields {
		if field.Key == key {
			return field, true
		}
	}
	return fieldEvidence{}, false
}

func issueExists(issues []validationIssue, field string, code string) bool {
	for _, issue := range issues {
		if issue.Field == field && issue.Code == code {
			return true
		}
	}
	return false
}
