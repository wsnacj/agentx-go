package fusion

import (
	"testing"

	"github.com/wsnacj/agentx-go/scenes/docparse/adapters"
)

func TestMergePrefersHigherConfidenceField(t *testing.T) {
	got := Merge([]adapters.Output{
		{RouteKind: "a", Status: "success", Fields: []map[string]any{{"key": "amount", "value": "10", "confidence": 0.2, "page_refs": []int{1}}}},
		{RouteKind: "b", Status: "success", Fields: []map[string]any{{"key": "amount", "value": "12", "confidence": 0.9, "page_refs": []int{2}}}},
	})
	if got.Status != StatusAnswerReady || got.FieldCount != 1 {
		t.Fatalf("unexpected merge result: %#v", got)
	}
	if got.Fields[0]["value"] != "12" {
		t.Fatalf("expected higher-confidence field, got %#v", got.Fields[0])
	}
}

func TestMergeRequiresReviewForMissingEvidence(t *testing.T) {
	got := Merge([]adapters.Output{
		{RouteKind: "a", Status: "success", Fields: []map[string]any{{"key": "amount", "value": "10"}}},
	})
	if got.Status != StatusReviewRequired || !got.ReviewRequired {
		t.Fatalf("expected review-required result: %#v", got)
	}
}

func TestMergePropagatesFieldReviewRequired(t *testing.T) {
	got := Merge([]adapters.Output{
		{RouteKind: "a", Status: "success", Fields: []map[string]any{{"key": "amount", "value": "10", "page_refs": []int{1}, "review_required": true}}},
	})
	if got.Status != StatusReviewRequired || !got.ReviewRequired {
		t.Fatalf("expected field review-required result: %#v", got)
	}
}

func TestMergeNormalizesFieldEvidenceSchema(t *testing.T) {
	got := Merge([]adapters.Output{
		{
			AdapterID: "ocrx_html_llm",
			RouteKind: "ocrx_html_llm",
			Status:    "success",
			Fields: []map[string]any{{
				"name":       "invoice_id",
				"value":      "INV-001",
				"confidence": "1.2",
				"page_ref":   "2",
				"bbox":       map[string]any{"page": 2, "x0": 0.1, "y0": 0.2, "x1": 0.5, "y1": 0.3},
			}},
		},
	})
	if got.SchemaVersion != SchemaVersion || got.Diagnostics["schema_version"] != SchemaVersion {
		t.Fatalf("expected schema version in result: %#v", got)
	}
	if got.Status != StatusAnswerReady || got.FieldCount != 1 {
		t.Fatalf("unexpected normalized merge result: %#v", got)
	}
	field := got.Fields[0]
	if field["key"] != "invoice_id" ||
		field["adapter_id"] != "ocrx_html_llm" ||
		field["route_kind"] != "ocrx_html_llm" ||
		field["source"] != "ocrx_html_llm" ||
		field["confidence"] != float64(1) ||
		field["evidence_complete"] != true {
		t.Fatalf("unexpected normalized field metadata: %#v", field)
	}
	if refs, ok := field["page_refs"].([]int); !ok || len(refs) != 1 || refs[0] != 2 {
		t.Fatalf("expected normalized page refs, got %#v", field["page_refs"])
	}
	if boxes, ok := field["bounding_boxes"].([]any); !ok || len(boxes) != 1 {
		t.Fatalf("expected normalized bounding boxes, got %#v", field["bounding_boxes"])
	}
	if !containsString(readStringSlice(field["evidence_kinds"]), "page_refs") ||
		!containsString(readStringSlice(field["evidence_kinds"]), "bounding_boxes") {
		t.Fatalf("expected evidence kinds, got %#v", field["evidence_kinds"])
	}
}

func TestMergeMarksMissingBBoxWithoutFailingPageGroundedField(t *testing.T) {
	got := Merge([]adapters.Output{
		{
			AdapterID: "host",
			RouteKind: "generic_text_projection",
			Status:    "success",
			Fields: []map[string]any{{
				"key":       "amount",
				"value":     "10",
				"page_refs": []int{1},
			}},
		},
	})
	if got.Status != StatusAnswerReady || got.ReviewRequired {
		t.Fatalf("expected page-grounded field to remain answer-ready: %#v", got)
	}
	field := got.Fields[0]
	if field["evidence_complete"] != true {
		t.Fatalf("expected evidence_complete=true from page_refs: %#v", field)
	}
	if !containsString(readStringSlice(field["warnings"]), "bbox_not_available") {
		t.Fatalf("expected bbox_not_available warning, got %#v", field["warnings"])
	}
	if !containsString(got.Warnings, "bbox_not_available") {
		t.Fatalf("expected result warnings to include field warning, got %#v", got.Warnings)
	}
}

func TestMergeNormalizesTableEvidenceSchema(t *testing.T) {
	got := Merge([]adapters.Output{
		{
			AdapterID: "table_adapter",
			RouteKind: "table_statement",
			Status:    "success",
			Tables: []map[string]any{{
				"name":        "transactions",
				"rows":        [][]string{{"date", "amount"}, {"2026-05-01", "10"}},
				"page_ref":    3,
				"table_cells": []map[string]any{{"row": 1, "col": 2, "page": 3}},
			}},
		},
	})
	if got.Status != StatusAnswerReady || got.TableCount != 1 {
		t.Fatalf("unexpected table merge result: %#v", got)
	}
	table := got.Tables[0]
	if table["key"] != "transactions" ||
		table["adapter_id"] != "table_adapter" ||
		table["route_kind"] != "table_statement" ||
		table["source"] != "table_adapter" ||
		table["evidence_complete"] != true {
		t.Fatalf("unexpected normalized table metadata: %#v", table)
	}
	if refs, ok := table["page_refs"].([]int); !ok || len(refs) != 1 || refs[0] != 3 {
		t.Fatalf("expected normalized table page refs, got %#v", table["page_refs"])
	}
	if cells, ok := table["cell_refs"].([]any); !ok || len(cells) != 1 {
		t.Fatalf("expected normalized cell refs, got %#v", table["cell_refs"])
	}
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
