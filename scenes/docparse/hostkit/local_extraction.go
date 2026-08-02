package hostkit

import "fmt"

func hasLocalParseResult(params map[string]any) bool {
	if _, ok := params["parse_result"]; ok {
		return true
	}
	return firstStringArg(params, "result_path") != ""
}

func (k *Kit) buildLocalFieldExtractionPayload(params map[string]any) evidencePayload {
	out := k.buildEvidencePayload(ToolDocparseExtractFields, params)
	out.TaskKind = firstNonEmptyString(params["task_kind"], "document.extract_fields")
	if out.Status != "success" {
		return out
	}
	if requested := requestedFieldKeys(params); len(requested) > 0 {
		reviewRequired := out.ReviewRequired
		out.Fields = filterFieldsByKey(out.Fields, requested)
		out.Fields = appendMissingRequestedFields(out.Fields, requested)
		out.FieldCount = len(out.Fields)
		out.Tables = tablesFromFieldCells(out.Fields)
		out.TableCount = len(out.Tables)
		out.ReviewRequired = reviewRequired || fieldsReviewRequired(out.Fields)
		out.EvidenceComplete = evidenceComplete(out.Fields, out.Tables)
		out.Passed = out.EvidenceComplete && !out.ReviewRequired
		if out.ReviewRequired {
			out.FailureClass = failureClassReviewRequired
			out.FailureCode = "docparse_review_required"
		} else if !out.Passed {
			out.FailureClass = failureClassEvidenceMissing
			out.FailureCode = "docparse_evidence_incomplete"
		} else {
			out.FailureClass = ""
			out.FailureCode = ""
		}
	}
	out.Summary = fmt.Sprintf("docparse fields projected from parse result: fields=%d tables=%d", out.FieldCount, out.TableCount)
	applyDocumentTypeMismatchToEvidence(&out)
	return out
}

func (k *Kit) buildLocalTableExtractionPayload(params map[string]any) evidencePayload {
	out := k.buildEvidencePayload(ToolDocparseExtractTable, params)
	out.TaskKind = firstNonEmptyString(params["task_kind"], "document.extract_table")
	if out.Status != "success" {
		return out
	}
	out.Fields = nil
	out.FieldCount = 0
	out.ReviewRequired = false
	requireCompleteRows := readEvidenceRequirements(params).CompleteTableRows
	out.EvidenceComplete = evidenceComplete(nil, out.Tables) && (!requireCompleteRows || completeTableRows(out.Tables))
	out.Passed = out.EvidenceComplete && !out.ReviewRequired
	if out.Passed {
		out.FailureClass = ""
		out.FailureCode = ""
		out.Summary = fmt.Sprintf("docparse tables projected from parse result: tables=%d", out.TableCount)
		applyDocumentTypeMismatchToEvidence(&out)
		return out
	}
	if requireCompleteRows && !completeTableRows(out.Tables) {
		out.FailureClass = failureClassTableRowMissing
		out.FailureCode = "docparse_table_rows_incomplete"
		out.Summary = "docparse table rows are incomplete"
	} else {
		out.FailureClass = failureClassTableCellMissing
		out.FailureCode = "docparse_table_evidence_missing"
		out.Summary = "docparse table evidence is incomplete"
	}
	applyDocumentTypeMismatchToEvidence(&out)
	return out
}

func (k *Kit) buildProfileProbePayload(params map[string]any) evidencePayload {
	out := k.buildEvidencePayload(ToolDocparseProfileProbe, params)
	out.TaskKind = "document.profile_probe"
	out.Fields = nil
	out.Tables = nil
	out.FieldCount = 0
	out.TableCount = 0
	out.EvidenceComplete = false
	out.Passed = false
	out.ReviewRequired = true
	if out.Status == "success" {
		out.FailureClass = failureClassReviewRequired
		out.FailureCode = "docparse_profile_probe_requires_review"
		out.Summary = "docparse profile probe requires review"
	}
	return out
}

func requestedFieldKeys(params map[string]any) []string {
	return uniqueSortedStrings(readStringArray(params["requested_fields"]))
}

func filterFieldsByKey(fields []fieldEvidence, keys []string) []fieldEvidence {
	if len(keys) == 0 {
		return fields
	}
	wanted := map[string]bool{}
	for _, key := range keys {
		wanted[key] = true
	}
	out := make([]fieldEvidence, 0, len(fields))
	for _, field := range fields {
		if wanted[field.Key] {
			out = append(out, field)
		}
	}
	return out
}

func appendMissingRequestedFields(fields []fieldEvidence, keys []string) []fieldEvidence {
	if len(keys) == 0 {
		return fields
	}
	seen := map[string]bool{}
	for _, field := range fields {
		if field.Key != "" {
			seen[field.Key] = true
		}
	}
	out := append([]fieldEvidence(nil), fields...)
	for _, key := range keys {
		if seen[key] {
			continue
		}
		out = append(out, fieldEvidence{
			Key:            key,
			ReviewRequired: true,
			Warnings:       []string{"field_not_found"},
		})
	}
	return out
}
