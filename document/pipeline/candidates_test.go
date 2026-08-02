package pipeline

import (
	"github.com/wsnacj/agentx-go/document/pipeline/configs"
	"github.com/wsnacj/agentx-go/document/pipeline/types"
	"strings"
	"testing"
)

func TestSelectFieldCandidateAppliesFieldPolicy(t *testing.T) {
	field := configs.FieldSpec{
		Key:               "Profit",
		Aliases:           []string{"Profit attributable to owners"},
		PreferredChapters: []string{"income_statement"},
		PeriodPolicy:      "current",
		UnitPolicy:        "prefer",
		PreferPatterns:    []string{`(?i)attributable to owners`},
		DisallowPatterns:  []string{`(?i)per share|eps`},
	}
	candidates := []types.FieldCandidate{
		{
			Value:      float64(10),
			Source:     "table",
			Confidence: 0.95,
			RowLabel:   "Basic EPS",
			Evidence:   "Basic EPS per share 10",
		},
		{
			Value:       float64(194073),
			Source:      "table",
			Confidence:  0.9,
			RowLabel:    "Profit attributable to owners",
			ColumnLabel: "2024",
			Period:      "2024",
			Evidence:    "Profit attributable to owners 194,073",
		},
		{
			Value:       float64(224842),
			Source:      "table",
			Confidence:  0.7,
			RowLabel:    "Profit attributable to owners",
			ColumnLabel: "2025",
			Period:      "2025",
			Unit:        "RMB million",
			Evidence:    "Profit attributable to owners 224,842",
		},
	}

	got, ok := selectFieldCandidate(field, "income_statement", candidates)
	if !ok {
		t.Fatal("expected selected field")
	}
	if got.Value != float64(224842) {
		t.Fatalf("expected current-period policy candidate, got %#v", got)
	}
	if len(got.Candidates) != 2 {
		t.Fatalf("expected disallowed EPS candidate to be filtered, got %#v", got.Candidates)
	}
	if got.Candidates[1].Score <= got.Candidates[0].Score {
		t.Fatalf("expected policy score to outrank higher base confidence: %#v", got.Candidates)
	}
}

func TestSelectFieldCandidateHonorsRequiredUnitPolicy(t *testing.T) {
	field := configs.FieldSpec{
		Key:        "Revenue",
		UnitPolicy: "required",
	}
	candidates := []types.FieldCandidate{{
		Value:      float64(100),
		Source:     "table",
		Confidence: 0.9,
		RowLabel:   "Revenue",
	}}

	if got, ok := selectFieldCandidate(field, "summary", candidates); ok {
		t.Fatalf("expected no selected field without source unit, got %#v", got)
	}
}

func TestSelectFieldCandidatePropagatesLayoutReferences(t *testing.T) {
	field := configs.FieldSpec{Key: "Revenue"}
	bbox := types.BoundingBox{
		Page:             3,
		X0:               10,
		Y0:               20,
		X1:               120,
		Y1:               42,
		Unit:             "pixel",
		CoordinateSystem: "page_top_left",
		Source:           "pdfparser",
	}
	cell := types.TableCellRef{
		Page:          3,
		TableIndex:    2,
		StartRow:      4,
		StartCol:      1,
		EndRow:        4,
		EndCol:        1,
		BoundingBoxes: []types.BoundingBox{bbox},
		Source:        "pdfparser",
	}

	got, ok := selectFieldCandidate(field, "summary", []types.FieldCandidate{{
		Value:         float64(100),
		Source:        "table",
		Confidence:    0.9,
		BoundingBoxes: []types.BoundingBox{bbox},
		TableCells:    []types.TableCellRef{cell},
	}})
	if !ok {
		t.Fatal("expected selected field")
	}
	if len(got.BoundingBoxes) != 1 || got.BoundingBoxes[0].X1 != 120 {
		t.Fatalf("expected selected candidate bbox to propagate to field, got %#v", got.BoundingBoxes)
	}
	if len(got.TableCells) != 1 || got.TableCells[0].TableIndex != 2 || len(got.Candidates[0].TableCells) != 1 {
		t.Fatalf("expected selected candidate table cell ref to propagate, got %#v", got)
	}
}

func TestSelectFieldCandidateBoostsIndependentExtractorAgreement(t *testing.T) {
	field := configs.FieldSpec{Key: "Revenue"}
	candidates := []types.FieldCandidate{
		{
			Value:      float64(100),
			Source:     "regex",
			Extractor:  "regex",
			Confidence: 0.70,
		},
		{
			Value:      float64(100),
			Source:     "table",
			Extractor:  "table",
			Confidence: 0.69,
		},
		{
			Value:      float64(200),
			Source:     "llm",
			Extractor:  "llm",
			Confidence: 0.78,
		},
	}

	got, ok := selectFieldCandidate(field, "summary", candidates)
	if !ok {
		t.Fatal("expected selected field")
	}
	if got.Value != float64(100) {
		t.Fatalf("expected independently agreed value to win, got %#v", got)
	}
	if got.Candidates[0].Score <= 0.70 || got.Candidates[1].Score <= 0.69 {
		t.Fatalf("expected agreed candidates to receive score boost, got %#v", got.Candidates)
	}
	if !strings.Contains(got.SelectionReason, "ensemble_agreement") || !got.ReviewRequired || !hasString(got.Warnings, fieldWarningCandidateConflict) {
		t.Fatalf("expected ensemble agreement selection with conflict review, got %#v", got)
	}
}

func TestSelectFieldCandidateDoesNotBoostDuplicateSameSource(t *testing.T) {
	field := configs.FieldSpec{Key: "Revenue"}
	candidates := []types.FieldCandidate{
		{
			Value:      float64(100),
			Source:     "table",
			Extractor:  "table",
			Confidence: 0.70,
		},
		{
			Value:      float64(100),
			Source:     "table",
			Extractor:  "table",
			Confidence: 0.69,
		},
		{
			Value:      float64(200),
			Source:     "llm",
			Extractor:  "llm",
			Confidence: 0.75,
		},
	}

	got, ok := selectFieldCandidate(field, "summary", candidates)
	if !ok {
		t.Fatal("expected selected field")
	}
	if got.Value != float64(200) {
		t.Fatalf("expected higher independent candidate to win without duplicate-source boost, got %#v", got)
	}
	if strings.Contains(got.SelectionReason, "ensemble_agreement") {
		t.Fatalf("did not expect duplicate same-source agreement, got %#v", got)
	}
}
