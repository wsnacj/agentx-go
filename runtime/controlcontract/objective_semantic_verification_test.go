package controlcontract

import (
	"context"
	"errors"
	"testing"
)

func TestBuildObjectiveSemanticVerificationFromJSONAdvisesOnly(t *testing.T) {
	raw := []byte(`{
		"advice_ref":"advice:semantic:test",
		"suggested_status":"satisfied",
		"coverage":[{
			"criterion_ref":"criterion:summary",
			"status":"covered",
			"covered_evidence_refs":[{"ref":"evidence:page_summary","kind":"summary","strength":"strong"}]
		}],
		"evidence_refs":[{"ref":"evidence:page_summary","kind":"summary","strength":"strong"}],
		"findings":["evidence covers the requested summary"]
	}`)

	got := BuildObjectiveSemanticVerificationFromJSON(ObjectiveSemanticVerificationJSONDecodeInput{
		RawJSON:   raw,
		AdviceRef: "advice:semantic:decode",
		SourceRef: "llm:semantic_verifier",
	})

	if !got.Decoded {
		t.Fatalf("expected decoded")
	}
	if got.Status != VerificationReviewRequired {
		t.Fatalf("semantic advice must stay review-required, got %s", got.Status)
	}
	if got.FailureClass != FailureNone {
		t.Fatalf("expected no failure class, got %s", got.FailureClass)
	}
	if got.NextHostAction != "run_deterministic_verification_gate" {
		t.Fatalf("expected deterministic gate next action, got %s", got.NextHostAction)
	}
	if !boundariesContain(got.Boundaries, "does_not_mark_objective_satisfied") {
		t.Fatalf("missing advice-only boundary: %#v", got.Boundaries)
	}
}

func TestBuildObjectiveSemanticVerificationFromJSONStrictDecode(t *testing.T) {
	got := BuildObjectiveSemanticVerificationFromJSON(ObjectiveSemanticVerificationJSONDecodeInput{
		RawJSON: []byte(`{"advice_ref":"advice:bad","unknown":true}`),
	})

	if got.Decoded {
		t.Fatalf("unexpected decoded report")
	}
	if got.Status != VerificationBlocked {
		t.Fatalf("expected blocked, got %s", got.Status)
	}
	if got.FailureClass != FailureInvalidInput {
		t.Fatalf("expected invalid input, got %s", got.FailureClass)
	}
	if !boundariesContain(got.Boundaries, "objective_semantic_verification_json_invalid") {
		t.Fatalf("missing invalid-json boundary: %#v", got.Boundaries)
	}
}

func TestBuildObjectiveSemanticVerificationReportsMissingConstraint(t *testing.T) {
	verifier := ObjectiveSemanticVerifierFunc(func(context.Context, ObjectiveSemanticVerifierRequest) (ObjectiveSemanticVerifierResponse, error) {
		return ObjectiveSemanticVerifierResponse{
			ResponseRef: "llm:semantic:response",
			Advice: ObjectiveSemanticVerificationAdvice{
				SuggestedStatus: VerificationPartial,
				Coverage: []ObjectiveSemanticCoverageAssessment{{
					CriterionRef:    "criterion:seat_inventory",
					Status:          ObjectiveSemanticCoverageMissing,
					MissingEvidence: []EvidenceRef{{Ref: "evidence:second_class_inventory", Kind: "inventory", Strength: EvidenceStrong}},
					MissingInputs:   []MissingInput{"host:second_class_inventory_evidence"},
					Findings:        []string{"available evidence does not cover second-class inventory"},
				}},
			},
		}, nil
	})

	got := BuildObjectiveSemanticVerification(context.Background(), ObjectiveSemanticVerificationInput{
		Enabled:  true,
		Verifier: verifier,
		Request: ObjectiveSemanticVerifierRequest{
			RequestRef: "request:semantic:test",
			Spec: ObjectiveSpec{
				SpecRef: "spec:ticket",
				SuccessCriteria: []ObjectiveSuccessCriterion{{
					CriteriaRef:      "criterion:seat_inventory",
					RequiredEvidence: []EvidenceRef{{Ref: "evidence:second_class_inventory", Kind: "inventory", Strength: EvidenceStrong}},
				}},
				RequiredEvidence: []EvidenceRef{{Ref: "evidence:second_class_inventory", Kind: "inventory", Strength: EvidenceStrong}},
			},
		},
	})

	if !got.VerifierCalled {
		t.Fatalf("expected verifier call")
	}
	if got.Status != VerificationPartial {
		t.Fatalf("expected partial, got %s", got.Status)
	}
	if got.FailureClass != FailureEvidenceMissing {
		t.Fatalf("expected evidence missing, got %s", got.FailureClass)
	}
	if got.NextHostAction != "add_evidence_node" {
		t.Fatalf("expected add evidence node, got %s", got.NextHostAction)
	}
	if !missingInputsContain(got.MissingInputs, "host:second_class_inventory_evidence") {
		t.Fatalf("missing evidence input not propagated: %#v", got.MissingInputs)
	}
}

func TestBuildObjectiveSemanticVerificationDisabledAndFailureAreSafe(t *testing.T) {
	disabled := BuildObjectiveSemanticVerification(context.Background(), ObjectiveSemanticVerificationInput{
		Enabled: false,
	})
	if disabled.Status != VerificationNotApplicable {
		t.Fatalf("expected not applicable when disabled, got %s", disabled.Status)
	}
	if disabled.VerifierCalled {
		t.Fatalf("disabled verifier should not be called")
	}
	if disabled.NextHostAction != "continue_objective_runtime_loop" {
		t.Fatalf("unexpected disabled next action: %s", disabled.NextHostAction)
	}

	failing := BuildObjectiveSemanticVerification(context.Background(), ObjectiveSemanticVerificationInput{
		Enabled: true,
		Verifier: ObjectiveSemanticVerifierFunc(func(context.Context, ObjectiveSemanticVerifierRequest) (ObjectiveSemanticVerifierResponse, error) {
			return ObjectiveSemanticVerifierResponse{}, errors.New("llm unavailable")
		}),
	})
	if failing.Status != VerificationBlocked {
		t.Fatalf("expected blocked on verifier error, got %s", failing.Status)
	}
	if failing.FailureClass != FailureExternalDependencyUnavailable {
		t.Fatalf("expected external dependency unavailable, got %s", failing.FailureClass)
	}
	if !boundariesContain(failing.Boundaries, "semantic_verification_deterministic_blocked_fallback") {
		t.Fatalf("missing deterministic fallback boundary: %#v", failing.Boundaries)
	}
}

func boundariesContain(values []Boundary, want Boundary) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func missingInputsContain(values []MissingInput, want MissingInput) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
