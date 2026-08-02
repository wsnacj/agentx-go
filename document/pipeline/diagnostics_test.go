package pipeline

import (
	"github.com/wsnacj/agentx-go/document/pipeline/types"
	"reflect"
	"testing"
)

func TestMatchedFieldDiagnosticSummarizesCandidateCalibration(t *testing.T) {
	result := types.FieldResult{
		Key:             "Revenue",
		Source:          "table",
		SelectionReason: "selected_highest_score_with_ensemble_agreement_with_conflict",
		ReviewRequired:  true,
		Candidates: []types.FieldCandidate{
			{Value: float64(100), Source: "regex", Extractor: "regex", Selected: true},
			{Value: float64(100), Source: "table", Extractor: "table"},
			{Value: float64(200), Source: "llm", Extractor: "llm"},
		},
	}

	got := matchedFieldDiagnostic("summary", "Revenue", result, "", []string{"candidate_conflict"})
	if got.CandidateCount != 3 || got.CandidateValueCount != 2 || got.AgreementSourceCount != 2 {
		t.Fatalf("unexpected candidate summary: %#v", got)
	}
	if !reflect.DeepEqual(got.CandidateSources, []string{"llm/llm", "regex/regex", "table/table"}) {
		t.Fatalf("unexpected candidate sources: %#v", got.CandidateSources)
	}
	if got.SelectionReason != result.SelectionReason || !got.ReviewRequired || len(got.Warnings) != 1 {
		t.Fatalf("unexpected diagnostic selection fields: %#v", got)
	}
}
