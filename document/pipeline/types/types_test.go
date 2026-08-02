package types

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestFieldCandidateLayoutReferencesAreOmittedWhenEmpty(t *testing.T) {
	payload, err := json.Marshal(FieldCandidate{
		Value:      float64(100),
		Source:     "table",
		Confidence: 0.9,
	})
	if err != nil {
		t.Fatalf("marshal field candidate: %v", err)
	}
	text := string(payload)
	if strings.Contains(text, "bounding_boxes") || strings.Contains(text, "table_cells") {
		t.Fatalf("expected empty layout references to be omitted, got %s", text)
	}
}

func TestFieldResultLayoutReferencesRoundTrip(t *testing.T) {
	input := FieldResult{
		Key:   "Revenue",
		Value: float64(100),
		BoundingBoxes: []BoundingBox{{
			Page:             1,
			X0:               0,
			Y0:               12,
			X1:               80,
			Y1:               24,
			Unit:             "pixel",
			CoordinateSystem: "page_top_left",
			Source:           "pdfparser",
		}},
		TableCells: []TableCellRef{{
			Page:       1,
			TableIndex: 1,
			StartRow:   2,
			StartCol:   1,
			EndRow:     2,
			EndCol:     1,
			Source:     "pdfparser",
		}},
	}

	payload, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("marshal field result: %v", err)
	}
	var got FieldResult
	if err := json.Unmarshal(payload, &got); err != nil {
		t.Fatalf("unmarshal field result: %v", err)
	}
	if len(got.BoundingBoxes) != 1 || got.BoundingBoxes[0].X0 != 0 || got.BoundingBoxes[0].X1 != 80 {
		t.Fatalf("unexpected bounding box round trip: %#v", got.BoundingBoxes)
	}
	if len(got.TableCells) != 1 || got.TableCells[0].TableIndex != 1 {
		t.Fatalf("unexpected table cell round trip: %#v", got.TableCells)
	}
}

func TestFieldResultUnmarshalsLegacyPayloadWithoutLayoutReferences(t *testing.T) {
	var got FieldResult
	if err := json.Unmarshal([]byte(`{"key":"Revenue","value":100,"source":"table","page_refs":[1]}`), &got); err != nil {
		t.Fatalf("unmarshal legacy field result: %v", err)
	}
	if got.Key != "Revenue" || len(got.BoundingBoxes) != 0 || len(got.TableCells) != 0 {
		t.Fatalf("unexpected legacy payload result: %#v", got)
	}
}
