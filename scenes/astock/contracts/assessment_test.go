package contracts

import "testing"

func TestBuildAssessmentBoundaryForInvestment(t *testing.T) {
	ready := BuildReadiness(AdapterStatusOK, FailureCodeNone, true, nil, nil)
	got := BuildAssessmentBoundary(AssessmentKindInvestment, "valuation snapshot", ready, []string{"PE is available"}, nil)
	if !got.NotInvestmentAdvice || got.AdviceBoundary == "" || len(got.VerifiedFacts) != 1 || !got.Readiness.AnswerReady {
		t.Fatalf("unexpected investment assessment boundary: %#v", got)
	}
}

func TestBuildAssessmentBoundaryForIncompleteEvidence(t *testing.T) {
	readiness := BuildReadiness(AdapterStatusEvidenceIncomplete, FailureCodeMissingFields, false, []string{"pe_ttm"}, nil)
	got := BuildAssessmentBoundary(AssessmentKindValuation, "valuation snapshot", readiness, []string{"price is available"}, []string{"pe_ttm"})
	if !got.NotInvestmentAdvice || got.AdviceBoundary == "" || got.Readiness.AnswerReady || len(got.MissingInputs) != 1 {
		t.Fatalf("unexpected incomplete assessment boundary: %#v", got)
	}
}
