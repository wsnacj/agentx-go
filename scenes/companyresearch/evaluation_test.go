package companyresearch

import (
	"strings"
	"testing"
)

func TestEvaluateCompanyResearchEvidencePassesCompleteSourceBackedEvidence(t *testing.T) {
	eval := EvaluateCompanyResearchEvidence(CompanyResearchEvaluationInput{
		SubjectCount:         1,
		ExpectedSubjectCount: 1,
		SubjectResolved:      true,
		AnswerReady:          true,
		GuardStatus:          "passed",
		RequestedDimensions:  []string{"financials", "market_data", "news"},
		ReadyDimensions:      []string{"financials", "market_data", "news"},
		SourceBacked:         true,
		FreshnessConfirmed:   true,
	})
	if !eval.Passed {
		t.Fatalf("expected evaluation to pass, got %#v", eval)
	}
}

func TestEvaluateCompanyResearchEvidenceAllowsBoundedPartialButDoesNotPassGate(t *testing.T) {
	eval := EvaluateCompanyResearchEvidence(CompanyResearchEvaluationInput{
		SubjectCount:                3,
		ExpectedSubjectCount:        3,
		SubjectResolved:             true,
		AnswerReady:                 false,
		GuardStatus:                 "needs_review",
		RequestedDimensions:         []string{"financials", "market_data", "news"},
		ReadyDimensions:             []string{"financials", "market_data"},
		MissingDimensions:           []string{"news"},
		SourceBacked:                true,
		FreshnessConfirmed:          true,
		AnswerContractRecommended:   true,
		FinalAnswerBoundaryObserved: true,
	})
	if eval.Passed {
		t.Fatalf("bounded partial evidence should not pass complete gate: %#v", eval)
	}
	if !eval.BoundaryOK || !strings.Contains(eval.FailureReason, "evidence_incomplete") {
		t.Fatalf("expected bounded partial to preserve boundary but fail completeness, got %#v", eval)
	}
}

func TestEvaluateCompanyResearchEvidenceRejectsOverClaim(t *testing.T) {
	eval := EvaluateCompanyResearchEvidence(CompanyResearchEvaluationInput{
		SubjectResolved:             true,
		AnswerReady:                 true,
		GuardStatus:                 "passed",
		RequestedDimensions:         []string{"financials"},
		ReadyDimensions:             []string{"financials"},
		SourceBacked:                true,
		FreshnessConfirmed:          true,
		OverClaimDetected:           true,
		AnswerContractRecommended:   false,
		FinalAnswerBoundaryObserved: false,
	})
	if eval.Passed || eval.BoundaryOK || !strings.Contains(eval.FailureReason, "over_claim_detected") {
		t.Fatalf("expected over-claim to fail boundary and gate, got %#v", eval)
	}
}

func TestEvaluateCompanyResearchEvidenceRejectsTaskConflictAndSubjectDrift(t *testing.T) {
	eval := EvaluateCompanyResearchEvidence(CompanyResearchEvaluationInput{
		SubjectCount:           1,
		ExpectedSubjectCount:   1,
		SubjectResolved:        true,
		AnswerReady:            true,
		GuardStatus:            "passed",
		RequestedDimensions:    []string{"financials", "market_data"},
		ReadyDimensions:        []string{"financials", "market_data"},
		SourceBacked:           true,
		FreshnessConfirmed:     true,
		TaskConflictCount:      1,
		SubjectResolutionDrift: true,
	})
	if eval.Passed || eval.TaskConflictFree || !eval.SubjectResolutionDrift {
		t.Fatalf("expected task conflict and subject drift to fail gate, got %#v", eval)
	}
	if !strings.Contains(eval.FailureReason, "task_conflict_detected") ||
		!strings.Contains(eval.FailureReason, "subject_resolution_drift") {
		t.Fatalf("expected task conflict failure reasons, got %#v", eval)
	}
}
