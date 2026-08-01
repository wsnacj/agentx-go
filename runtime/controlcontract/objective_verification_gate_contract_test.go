package controlcontract

import "testing"

func TestObjectiveVerificationGateUsesNormalizedEvidence(t *testing.T) {
	frame := ObjectiveFrame{
		ID:              "objective:metrics",
		ControlMode:     ControlModeObjective,
		Intensity:       IntensityL3ManagedObjective,
		SuccessCriteria: []string{"return a verified metric"},
		RequiredEvidence: []EvidenceRef{{
			Ref: "evidence:metric", Kind: "metric", Strength: EvidenceStrong,
			Source: "adapter:metrics_readonly",
		}},
	}.Normalize()
	normalized := ObservationNormalizationResult{
		Status: VerificationSatisfied,
		Frame:  frame,
		Observations: []Observation{{
			Kind: "metric", Source: "adapter:metrics_readonly", Subject: "objective:metrics",
			Name: "cpu_usage", Value: "10.5", Unit: "percent", Strength: EvidenceStrong,
			EvidenceRefs: []EvidenceRef{{
				Ref: "evidence:metric", Kind: "metric", Strength: EvidenceStrong,
				Source: "adapter:metrics_readonly",
			}},
		}},
	}.Normalize()

	gate := BuildObjectiveVerificationGate(ObjectiveVerificationGateInput{
		Frame: frame, Normalization: normalized,
	})
	if gate.Status != VerificationSatisfied || !gate.Satisfied || len(gate.Requirements) != 1 || !gate.Requirements[0].Satisfied {
		t.Fatalf("verification gate = %#v", gate)
	}
}

func TestObservationNormalizationResultRejectsRawPayload(t *testing.T) {
	result := (ObservationNormalizationResult{
		Status:          VerificationSatisfied,
		Observations:    []Observation{{Kind: "metric", Source: "adapter:metrics"}},
		RawOutputLoaded: true,
	}).Normalize()
	if result.ReadyForVerification || result.Status != VerificationReviewRequired || result.FailureClass != FailureEvidenceWeak {
		t.Fatalf("unsafe normalization result = %#v", result)
	}
}
