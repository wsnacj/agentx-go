package controlcontract

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestBuildObjectiveRequiredEvidenceContractCompilesExplicitBindings(t *testing.T) {
	frame := objectiveRequiredEvidenceTestFrame()
	contract := BuildObjectiveRequiredEvidenceContract(ObjectiveRequiredEvidenceContractInput{
		Frame:       frame,
		ContractRef: "contract:metrics_required_evidence",
		SourceRef:   "scene:agentx_operations",
		Bindings: []ObjectiveRequiredEvidenceBinding{
			{
				CriteriaIndex: 1,
				CriteriaRef:   "criteria:collect_metrics",
				Evidence: EvidenceRef{
					Ref:      "evidence:metrics_snapshot",
					Kind:     "metric",
					Source:   "scene:agentx_operations",
					Strength: EvidenceAdequate,
				},
			},
			{
				CriteriaIndex: 2,
				CriteriaRef:   "criteria:evaluate_thresholds",
				Evidence: EvidenceRef{
					Ref:      "evidence:threshold_judgement",
					Kind:     "target_status",
					Source:   "scene:agentx_operations",
					Strength: EvidenceAdequate,
				},
			},
		},
	})

	if contract.Status != VerificationSatisfied ||
		contract.NextHostAction != "run_objective_verification_gate" ||
		contract.RunnerEffect != "none" ||
		contract.PromptEffect != "none" ||
		len(contract.MissingInputs) != 0 ||
		len(contract.Frame.RequiredEvidence) != 2 {
		t.Fatalf("unexpected required evidence contract: %#v", contract)
	}
	if !objectiveRequiredEvidenceBoundaryContains(contract.Boundaries, "host_scene_explicit_evidence_mapping") ||
		!objectiveRequiredEvidenceBoundaryContains(contract.Boundaries, "no_success_criteria_text_inference") ||
		!objectiveRequiredEvidenceBoundaryContains(contract.Boundaries, "required_evidence_contract_ready") {
		t.Fatalf("unexpected boundaries: %#v", contract.Boundaries)
	}

	normalization := (ObservationNormalizationResult{
		Status:     VerificationSatisfied,
		Frame:      contract.Frame,
		SourceKind: "operations",
		SourceRef:  "scene:agentx_operations",
		Observations: []Observation{
			{
				Kind:     "metric",
				Source:   "scene:agentx_operations",
				Subject:  "objective:local_metrics",
				Name:     "cpu_percent",
				Value:    "12.5",
				Strength: EvidenceStrong,
				EvidenceRefs: []EvidenceRef{{
					Ref:      "evidence:metrics_snapshot",
					Kind:     "metric",
					Source:   "scene:agentx_operations",
					Strength: EvidenceStrong,
				}},
			},
			{
				Kind:     "target_status",
				Source:   "scene:agentx_operations",
				Subject:  "objective:local_metrics",
				Name:     "threshold_status",
				Value:    "normal",
				Strength: EvidenceStrong,
				EvidenceRefs: []EvidenceRef{{
					Ref:      "evidence:threshold_judgement",
					Kind:     "target_status",
					Source:   "scene:agentx_operations",
					Strength: EvidenceStrong,
				}},
			},
		},
		ObservationKinds: []string{"metric", "target_status"},
	}).Normalize()
	gate := BuildObjectiveVerificationGate(ObjectiveVerificationGateInput{
		Frame:         contract.Frame,
		Normalization: normalization,
	})
	if gate.Status != VerificationSatisfied || !gate.Satisfied || len(gate.Requirements) != 2 {
		t.Fatalf("expected compiled contract to satisfy verification gate, got %#v", gate)
	}
}

func TestBuildObjectiveRequiredEvidenceContractBlocksIncompleteMapping(t *testing.T) {
	contract := BuildObjectiveRequiredEvidenceContract(ObjectiveRequiredEvidenceContractInput{
		Frame: objectiveRequiredEvidenceTestFrame(),
		Bindings: []ObjectiveRequiredEvidenceBinding{{
			CriteriaIndex: 1,
			Evidence: EvidenceRef{
				Ref:    "evidence:metrics_snapshot",
				Kind:   "metric",
				Source: "scene:agentx_operations",
			},
		}},
	})

	if contract.Status != VerificationBlocked ||
		contract.FailureClass != FailureEvidenceMissing ||
		contract.NextHostAction != "provide_required_evidence_contract" ||
		!objectiveRequiredEvidenceMissingContains(contract.MissingInputs, "host:required_evidence_criteria_2") ||
		len(contract.Frame.RequiredEvidence) != 0 {
		t.Fatalf("expected incomplete mapping to block, got %#v", contract)
	}
}

func TestBuildObjectiveRequiredEvidenceContractBlocksMissingKind(t *testing.T) {
	contract := BuildObjectiveRequiredEvidenceContract(ObjectiveRequiredEvidenceContractInput{
		Frame: objectiveRequiredEvidenceTestFrame(),
		Bindings: []ObjectiveRequiredEvidenceBinding{
			{
				CriteriaIndex: 1,
				Evidence: EvidenceRef{
					Ref:    "evidence:metrics_snapshot",
					Source: "scene:agentx_operations",
				},
			},
			{
				CriteriaIndex: 2,
				Evidence: EvidenceRef{
					Ref:    "evidence:threshold_judgement",
					Kind:   "target_status",
					Source: "scene:agentx_operations",
				},
			},
		},
	})

	if contract.Status != VerificationBlocked ||
		!objectiveRequiredEvidenceMissingContains(contract.MissingInputs, "host:required_evidence_kind") {
		t.Fatalf("expected missing evidence kind to block, got %#v", contract)
	}
}

func TestBuildObjectiveRequiredEvidenceContractBlocksPrecompiledEvidenceWithoutKind(t *testing.T) {
	contract := BuildObjectiveRequiredEvidenceContract(ObjectiveRequiredEvidenceContractInput{
		Frame: objectiveRequiredEvidenceTestFrame(),
		RequiredEvidence: []EvidenceRef{{
			Ref:    "evidence:metrics_snapshot",
			Source: "scene:agentx_operations",
		}},
	})

	if contract.Status != VerificationBlocked ||
		!objectiveRequiredEvidenceMissingContains(contract.MissingInputs, "host:required_evidence_kind") {
		t.Fatalf("expected precompiled required evidence without kind to block, got %#v", contract)
	}
}

func TestBuildObjectiveRequiredEvidenceContractRejectsUnsafeRefsWithoutLeak(t *testing.T) {
	unsafeRef := "/" + "Users/example/" + "secret-token"
	contract := BuildObjectiveRequiredEvidenceContract(ObjectiveRequiredEvidenceContractInput{
		Frame:       objectiveRequiredEvidenceTestFrame(),
		ContractRef: "contract:metrics_required_evidence",
		Bindings: []ObjectiveRequiredEvidenceBinding{{
			CriteriaIndex: 1,
			Evidence: EvidenceRef{
				Ref:    DisplaySafeRef(unsafeRef),
				Kind:   "metric",
				Source: "scene:agentx_operations",
			},
		}},
	})

	if contract.Status != VerificationReviewRequired ||
		contract.FailureClass != FailureEvidenceWeak ||
		contract.NextHostAction != "provide_display_safe_refs" ||
		!objectiveRequiredEvidenceMissingContains(contract.MissingInputs, "host:display_safe_refs") {
		t.Fatalf("expected unsafe refs to require review, got %#v", contract)
	}
	payload, err := json.Marshal(contract)
	if err != nil {
		t.Fatalf("marshal contract: %v", err)
	}
	if strings.Contains(string(payload), unsafeRef) || strings.Contains(string(payload), "secret-token") {
		t.Fatalf("unsafe ref leaked in %s", payload)
	}
}

func objectiveRequiredEvidenceTestFrame() ObjectiveFrame {
	return ObjectiveFrame{
		ID:              "objective:local_metrics",
		ControlMode:     ControlModeOperations,
		Intensity:       IntensityL2BoundedToolLoop,
		SuccessCriteria: []string{"collect local metrics", "evaluate thresholds"},
	}
}

func objectiveRequiredEvidenceBoundaryContains(values []Boundary, want Boundary) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func objectiveRequiredEvidenceMissingContains(values []MissingInput, want MissingInput) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
