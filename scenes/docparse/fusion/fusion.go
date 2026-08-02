// Package fusion merges outputs from document parsing routes into a single
// evidence-oriented result.
package fusion

import (
	"fmt"
	"sort"
	"strings"

	"github.com/wsnacj/agentx-go/scenes/docparse/adapters"
)

const (
	StatusAnswerReady    = "answer_ready"
	StatusReviewRequired = "review_required"
	StatusFailed         = "failed"
	SchemaVersion        = "agentx_docparse_fusion_v1"
)

// Result is the fused document understanding output.
type Result struct {
	SchemaVersion    string           `json:"schema_version,omitempty"`
	Status           string           `json:"status"`
	Fields           []map[string]any `json:"fields,omitempty"`
	Tables           []map[string]any `json:"tables,omitempty"`
	FieldCount       int              `json:"field_count"`
	TableCount       int              `json:"table_count"`
	EvidenceComplete bool             `json:"evidence_complete"`
	ReviewRequired   bool             `json:"review_required"`
	Warnings         []string         `json:"warnings,omitempty"`
	Diagnostics      map[string]any   `json:"diagnostics,omitempty"`
}

// Merge combines adapter outputs without inventing missing fields.
func Merge(outputs []adapters.Output) Result {
	result := Result{
		SchemaVersion: SchemaVersion,
		Diagnostics: map[string]any{
			"adapter_count":  len(outputs),
			"schema_version": SchemaVersion,
		},
	}
	if len(outputs) == 0 {
		result.Status = StatusFailed
		result.EvidenceComplete = false
		result.ReviewRequired = true
		result.Warnings = []string{"no_adapter_outputs"}
		return result
	}
	fieldsByKey := map[string]map[string]any{}
	tables := []map[string]any{}
	warnings := []string{}
	routeKinds := []string{}
	for _, output := range outputs {
		if output.ReviewRequired {
			result.ReviewRequired = true
		}
		if route := strings.TrimSpace(output.RouteKind); route != "" {
			routeKinds = append(routeKinds, route)
		}
		warnings = append(warnings, output.Warnings...)
		for _, field := range output.Fields {
			field = normalizeField(field, output)
			key := fieldKey(field)
			if key == "" {
				continue
			}
			warnings = append(warnings, readStringSlice(field["warnings"])...)
			if existing, ok := fieldsByKey[key]; ok {
				fieldsByKey[key] = betterField(existing, field)
				continue
			}
			fieldsByKey[key] = cloneMap(field)
		}
		for _, table := range output.Tables {
			table = normalizeTable(table, output)
			warnings = append(warnings, readStringSlice(table["warnings"])...)
			if value, ok := table["review_required"].(bool); ok && value {
				result.ReviewRequired = true
			}
			tables = append(tables, table)
		}
	}
	keys := make([]string, 0, len(fieldsByKey))
	for key := range fieldsByKey {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		field := fieldsByKey[key]
		if fieldNeedsReview(field) {
			result.ReviewRequired = true
		}
		result.Fields = append(result.Fields, field)
	}
	result.Tables = tables
	result.FieldCount = len(result.Fields)
	result.TableCount = len(result.Tables)
	result.EvidenceComplete = evidenceComplete(result.Fields, result.Tables)
	result.Warnings = uniqueStrings(warnings)
	result.Diagnostics["route_kinds"] = uniqueStrings(routeKinds)
	switch {
	case result.ReviewRequired || !result.EvidenceComplete:
		result.Status = StatusReviewRequired
		result.ReviewRequired = true
		if !result.EvidenceComplete {
			result.Warnings = uniqueStrings(append(result.Warnings, "evidence_incomplete"))
		}
	case result.EvidenceComplete:
		result.Status = StatusAnswerReady
	default:
		result.Status = StatusFailed
		result.ReviewRequired = true
		result.Warnings = uniqueStrings(append(result.Warnings, "evidence_incomplete"))
	}
	return result
}

func betterField(existing map[string]any, candidate map[string]any) map[string]any {
	if fieldHasValue(candidate) && !fieldHasValue(existing) {
		return cloneMap(candidate)
	}
	if fieldConfidence(candidate) > fieldConfidence(existing) {
		return cloneMap(candidate)
	}
	return existing
}

func fieldNeedsReview(field map[string]any) bool {
	if value, ok := field["review_required"].(bool); ok && value {
		return true
	}
	for _, warning := range readStringSlice(field["warnings"]) {
		switch strings.TrimSpace(warning) {
		case "field_not_found", "review_required", "bbox_not_available_review_required":
			return true
		}
	}
	return false
}

func fieldHasValue(field map[string]any) bool {
	for _, key := range []string{"normalized_value", "value", "raw_value"} {
		value, ok := field[key]
		if !ok || value == nil {
			continue
		}
		if text, ok := value.(string); ok {
			if strings.TrimSpace(text) != "" {
				return true
			}
			continue
		}
		return true
	}
	return false
}

func evidenceComplete(fields []map[string]any, tables []map[string]any) bool {
	if len(fields) == 0 && len(tables) == 0 {
		return false
	}
	for _, field := range fields {
		if !fieldHasValue(field) || !fieldHasEvidence(field) {
			return false
		}
	}
	for _, table := range tables {
		if !tableHasEvidence(table) {
			return false
		}
	}
	return true
}

func fieldHasEvidence(field map[string]any) bool {
	if hasTextEvidence(field["evidence"]) {
		return true
	}
	return len(readAnySlice(field["page_refs"])) > 0 ||
		len(readAnySlice(field["bounding_boxes"])) > 0 ||
		len(readAnySlice(field["table_cells"])) > 0
}

func tableHasEvidence(table map[string]any) bool {
	if hasTextEvidence(table["evidence"]) {
		return true
	}
	return len(readAnySlice(table["page_refs"])) > 0 || len(readAnySlice(table["cell_refs"])) > 0
}

func readStringSlice(value any) []string {
	items := readAnySlice(value)
	out := make([]string, 0, len(items))
	for _, item := range items {
		text := strings.TrimSpace(fmt.Sprint(item))
		if text != "" && text != "<nil>" {
			out = append(out, text)
		}
	}
	return out
}

func readAnySlice(value any) []any {
	switch typed := value.(type) {
	case []any:
		return typed
	case []string:
		out := make([]any, 0, len(typed))
		for _, item := range typed {
			out = append(out, item)
		}
		return out
	case []int:
		out := make([]any, 0, len(typed))
		for _, item := range typed {
			out = append(out, item)
		}
		return out
	default:
		return nil
	}
}

func cloneMap(input map[string]any) map[string]any {
	out := make(map[string]any, len(input))
	for key, value := range input {
		out[key] = value
	}
	return out
}

func uniqueStrings(values []string) []string {
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
