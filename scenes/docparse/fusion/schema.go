package fusion

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/wsnacj/agentx-go/scenes/docparse/adapters"
)

func normalizeField(field map[string]any, output adapters.Output) map[string]any {
	out := cloneMap(field)
	if key := fieldKey(out); key != "" {
		out["key"] = key
	}
	if confidence, ok := normalizedConfidence(out["confidence"]); ok {
		out["confidence"] = confidence
	} else {
		out["confidence"] = float64(0)
	}
	if source := strings.TrimSpace(fmt.Sprint(out["source"])); source == "" || source == "<nil>" {
		out["source"] = fallbackSource(output)
	}
	if adapterID := strings.TrimSpace(output.AdapterID); adapterID != "" {
		out["adapter_id"] = adapterID
	}
	if routeKind := strings.TrimSpace(output.RouteKind); routeKind != "" {
		out["route_kind"] = routeKind
	}
	if pageRefs := normalizePageRefs(out["page_refs"], out["page_ref"]); len(pageRefs) > 0 {
		out["page_refs"] = pageRefs
	}
	if boxes := normalizeEvidenceItems(out["bounding_boxes"], out["bbox"]); len(boxes) > 0 {
		out["bounding_boxes"] = boxes
	}
	if cells := normalizeEvidenceItems(out["table_cells"], out["cell_refs"]); len(cells) > 0 {
		out["table_cells"] = cells
	}
	evidenceKinds := fieldEvidenceKinds(out)
	evidenceComplete := fieldHasValue(out) && len(evidenceKinds) > 0
	out["evidence_complete"] = evidenceComplete
	out["evidence_kinds"] = evidenceKinds
	warnings := readStringSlice(out["warnings"])
	if !fieldHasValue(out) {
		warnings = append(warnings, "field_value_missing")
	}
	if len(normalizePageRefs(out["page_refs"])) == 0 {
		warnings = append(warnings, "page_refs_not_available")
	}
	if len(normalizeEvidenceItems(out["bounding_boxes"])) == 0 {
		warnings = append(warnings, "bbox_not_available")
	}
	if len(evidenceKinds) == 0 {
		warnings = append(warnings, "evidence_missing")
	}
	if fieldNeedsReview(out) || !evidenceComplete {
		out["review_required"] = true
	}
	if len(warnings) > 0 {
		out["warnings"] = uniqueStrings(warnings)
	}
	return out
}

func normalizeTable(table map[string]any, output adapters.Output) map[string]any {
	out := cloneMap(table)
	if key := tableKey(out); key != "" {
		out["key"] = key
	}
	if confidence, ok := normalizedConfidence(out["confidence"]); ok {
		out["confidence"] = confidence
	}
	if source := strings.TrimSpace(fmt.Sprint(out["source"])); source == "" || source == "<nil>" {
		out["source"] = fallbackSource(output)
	}
	if adapterID := strings.TrimSpace(output.AdapterID); adapterID != "" {
		out["adapter_id"] = adapterID
	}
	if routeKind := strings.TrimSpace(output.RouteKind); routeKind != "" {
		out["route_kind"] = routeKind
	}
	if pageRefs := normalizePageRefs(out["page_refs"], out["page_ref"]); len(pageRefs) > 0 {
		out["page_refs"] = pageRefs
	}
	if cellRefs := normalizeEvidenceItems(out["cell_refs"], out["table_cells"]); len(cellRefs) > 0 {
		out["cell_refs"] = cellRefs
	}
	evidenceKinds := tableEvidenceKinds(out)
	evidenceComplete := len(evidenceKinds) > 0
	out["evidence_complete"] = evidenceComplete
	out["evidence_kinds"] = evidenceKinds
	warnings := readStringSlice(out["warnings"])
	if len(normalizePageRefs(out["page_refs"])) == 0 {
		warnings = append(warnings, "page_refs_not_available")
	}
	if len(normalizeEvidenceItems(out["cell_refs"])) == 0 {
		warnings = append(warnings, "cell_refs_not_available")
	}
	if !evidenceComplete {
		warnings = append(warnings, "evidence_missing")
		out["review_required"] = true
	}
	if len(warnings) > 0 {
		out["warnings"] = uniqueStrings(warnings)
	}
	return out
}

func fieldKey(field map[string]any) string {
	for _, key := range []string{"key", "field", "name"} {
		value := strings.TrimSpace(fmt.Sprint(field[key]))
		if value != "" && value != "<nil>" {
			return value
		}
	}
	return ""
}

func tableKey(table map[string]any) string {
	for _, key := range []string{"key", "table", "name"} {
		value := strings.TrimSpace(fmt.Sprint(table[key]))
		if value != "" && value != "<nil>" {
			return value
		}
	}
	return ""
}

func fieldConfidence(field map[string]any) float64 {
	confidence, _ := normalizedConfidence(field["confidence"])
	return confidence
}

func normalizedConfidence(value any) (float64, bool) {
	switch typed := value.(type) {
	case float64:
		return clampConfidence(typed), true
	case float32:
		return clampConfidence(float64(typed)), true
	case int:
		return clampConfidence(float64(typed)), true
	case int64:
		return clampConfidence(float64(typed)), true
	case string:
		parsed, err := strconv.ParseFloat(strings.TrimSpace(typed), 64)
		if err == nil {
			return clampConfidence(parsed), true
		}
	case jsonNumber:
		parsed, err := typed.Float64()
		if err == nil {
			return clampConfidence(parsed), true
		}
	}
	return 0, false
}

func clampConfidence(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

func fallbackSource(output adapters.Output) string {
	if adapterID := strings.TrimSpace(output.AdapterID); adapterID != "" {
		return adapterID
	}
	if routeKind := strings.TrimSpace(output.RouteKind); routeKind != "" {
		return routeKind
	}
	return "unknown_adapter"
}

func fieldEvidenceKinds(field map[string]any) []string {
	kinds := []string{}
	if hasTextEvidence(field["evidence"]) {
		kinds = append(kinds, "text_snippet")
	}
	if len(normalizePageRefs(field["page_refs"])) > 0 {
		kinds = append(kinds, "page_refs")
	}
	if len(normalizeEvidenceItems(field["bounding_boxes"])) > 0 {
		kinds = append(kinds, "bounding_boxes")
	}
	if len(normalizeEvidenceItems(field["table_cells"])) > 0 {
		kinds = append(kinds, "table_cells")
	}
	return uniqueStrings(kinds)
}

func tableEvidenceKinds(table map[string]any) []string {
	kinds := []string{}
	if hasTextEvidence(table["evidence"]) {
		kinds = append(kinds, "text_snippet")
	}
	if len(normalizePageRefs(table["page_refs"])) > 0 {
		kinds = append(kinds, "page_refs")
	}
	if len(normalizeEvidenceItems(table["cell_refs"])) > 0 {
		kinds = append(kinds, "cell_refs")
	}
	return uniqueStrings(kinds)
}

func hasTextEvidence(value any) bool {
	text := strings.TrimSpace(fmt.Sprint(value))
	return text != "" && text != "<nil>"
}

func normalizePageRefs(values ...any) []int {
	seen := map[int]bool{}
	out := []int{}
	var add func(any)
	add = func(value any) {
		switch typed := value.(type) {
		case nil:
			return
		case int:
			if typed > 0 && !seen[typed] {
				seen[typed] = true
				out = append(out, typed)
			}
		case int64:
			add(int(typed))
		case float64:
			add(int(typed))
		case jsonNumber:
			if parsed, err := typed.Float64(); err == nil {
				add(int(parsed))
			}
		case string:
			for _, token := range strings.FieldsFunc(typed, func(r rune) bool {
				return r == ',' || r == ';' || r == ' ' || r == '\n' || r == '\t'
			}) {
				parsed, err := strconv.Atoi(strings.TrimSpace(token))
				if err == nil {
					add(parsed)
				}
			}
		case []int:
			for _, item := range typed {
				add(item)
			}
		case []int64:
			for _, item := range typed {
				add(item)
			}
		case []float64:
			for _, item := range typed {
				add(item)
			}
		case []string:
			for _, item := range typed {
				add(item)
			}
		case []any:
			for _, item := range typed {
				add(item)
			}
		default:
			rv := reflect.ValueOf(value)
			if rv.IsValid() && (rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array) {
				for i := 0; i < rv.Len(); i++ {
					add(rv.Index(i).Interface())
				}
			}
		}
	}
	for _, value := range values {
		add(value)
	}
	sort.Ints(out)
	return out
}

func normalizeEvidenceItems(values ...any) []any {
	out := []any{}
	for _, value := range values {
		out = append(out, evidenceItems(value)...)
	}
	return out
}

func evidenceItems(value any) []any {
	switch typed := value.(type) {
	case nil:
		return nil
	case []any:
		out := make([]any, 0, len(typed))
		for _, item := range typed {
			if isUsableEvidenceItem(item) {
				out = append(out, item)
			}
		}
		return out
	case []map[string]any:
		out := make([]any, 0, len(typed))
		for _, item := range typed {
			if len(item) > 0 {
				out = append(out, cloneMap(item))
			}
		}
		return out
	case map[string]any:
		if len(typed) == 0 {
			return nil
		}
		return []any{cloneMap(typed)}
	default:
		rv := reflect.ValueOf(value)
		if rv.IsValid() && (rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array) {
			out := make([]any, 0, rv.Len())
			for i := 0; i < rv.Len(); i++ {
				item := rv.Index(i).Interface()
				if isUsableEvidenceItem(item) {
					out = append(out, item)
				}
			}
			return out
		}
		if isUsableEvidenceItem(value) {
			return []any{value}
		}
		return nil
	}
}

func isUsableEvidenceItem(value any) bool {
	if value == nil {
		return false
	}
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed) != ""
	case map[string]any:
		return len(typed) > 0
	default:
		return true
	}
}

type jsonNumber interface {
	Float64() (float64, error)
}
