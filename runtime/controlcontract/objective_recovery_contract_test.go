package controlcontract

import "testing"

func TestBuildObjectiveRecoveryContractFromJSONBuildsReplanProposal(t *testing.T) {
	report := BuildObjectiveRecoveryContractFromJSON(ObjectiveRecoveryContractJSONDecodeInput{
		RawJSON: []byte(`{
			"answer_contract": {
				"final_answer_recommended": false,
				"recovery_recommended": "True",
				"recovery_reason": "recoverable_missing_evidence",
				"suggested_recovery_tools": ["capability:lookup_a"],
				"recovery_targets": [{
					"target_ref": "target:item_a",
					"missing_dimension": "dimension_a",
					"failure_class": "evidence_missing",
					"suggested_tools": ["capability:lookup_a"]
				}, {
					"target_ref": "target:item_b",
					"missing_input": "evidence:dimension_b",
					"failure_class": "evidence_weak",
					"suggested_tools": ["capability:lookup_b"]
				}]
			}
		}`),
		ContractRef:        "contract:recovery_case",
		SourceRef:          "source:primary_attempt",
		Producer:           "adapter:primary",
		ObjectiveID:        "objective:recovery",
		CurrentStrategyRef: "strategy:primary",
	})

	if !report.Decoded ||
		report.Status != VerificationPartial ||
		report.FailureClass != FailureEvidenceMissing ||
		report.NextHostAction != "host_may_add_evidence_node" {
		t.Fatalf("unexpected recovery decode report: %#v", report)
	}
	contract := report.Contract
	if !contract.Recommended ||
		contract.FinalAnswerRecommended ||
		contract.TargetCount != 2 ||
		contract.ReplannerSource.SourceKind != ReplannerSourceRecovery ||
		contract.ReplannerSource.Candidate.ID != "capability:lookup_a" ||
		contract.ReplannerSource.Proposal.Kind != "objective_recovery_proposal" {
		t.Fatalf("unexpected recovery contract/source: %#v", contract)
	}
	if contract.ReplanProposal.Action != ObjectiveReplanProposalActionAddEvidenceNode ||
		contract.ReplanProposal.NextHostAction != "host_may_add_evidence_node" ||
		len(contract.ReplanProposal.Steps) != 2 {
		t.Fatalf("unexpected recovery replan proposal: %#v", contract.ReplanProposal)
	}
	first := contract.ReplanProposal.Steps[0]
	if first.Action != ObjectiveReplanProposalActionAddEvidenceNode ||
		first.NextStrategy != "capability:lookup_a" ||
		!objectiveReplanProposalMissingInputContains(first.MissingInputs, "evidence:dimension_a") ||
		!objectiveRecoveryEvidenceContains(first.RequiredEvidence, "evidence:dimension_a") {
		t.Fatalf("unexpected first recovery step: %#v", first)
	}
	second := contract.ReplanProposal.Steps[1]
	if second.NextStrategy != "capability:lookup_a" ||
		!objectiveReplanProposalDisplayRefContains(second.CapabilityRefs, "capability:lookup_b") ||
		!objectiveReplanProposalMissingInputContains(second.MissingInputs, "evidence:dimension_b") {
		t.Fatalf("unexpected second recovery step: %#v", second)
	}
	if !objectiveReplanProposalBoundaryContains(contract.Boundaries, "display_safe_refs_only") ||
		!objectiveReplanProposalBoundaryContains(contract.ReplanProposal.Boundaries, "objective_recovery_replan_proposal") {
		t.Fatalf("expected display-safe recovery boundaries, contract=%#v proposal=%#v", contract.Boundaries, contract.ReplanProposal.Boundaries)
	}
}

func TestBuildObjectiveRecoveryContractDoesNotReplanWhenFinalIsAllowed(t *testing.T) {
	contract := BuildObjectiveRecoveryContract(ObjectiveRecoveryContractInput{
		ContractRef:            "contract:recovery_not_needed",
		SourceRef:              "source:ready_attempt",
		RecoveryRecommended:    true,
		FinalAnswerRecommended: true,
		Targets: []ObjectiveRecoveryTarget{{
			MissingInput:      "evidence:dimension_ready",
			SuggestedToolRefs: []DisplaySafeRef{"capability:lookup_ready"},
		}},
	})
	if contract.Status != VerificationNotApplicable ||
		contract.FailureClass != FailureNone ||
		contract.NextHostAction != "return_current_result" ||
		contract.ReplanProposal.Action != ObjectiveReplanProposalActionNone {
		t.Fatalf("final-ready recovery contract should not request replan: %#v", contract)
	}
}

func TestBuildObjectiveRecoveryContractDowngradesUnsafeRefs(t *testing.T) {
	contract := BuildObjectiveRecoveryContract(ObjectiveRecoveryContractInput{
		ContractRef:         "contract:unsafe_recovery",
		SourceRef:           "/tmp/raw-source.json",
		RecoveryRecommended: true,
		Targets: []ObjectiveRecoveryTarget{{
			MissingInput:      "evidence:dimension_a",
			SuggestedToolRefs: []DisplaySafeRef{"capability:lookup_a"},
		}},
	})
	if contract.Status != VerificationReviewRequired ||
		contract.FailureClass != FailureEvidenceWeak ||
		contract.NextHostAction != "provide_display_safe_refs" ||
		!objectiveReplanProposalMissingInputContains(contract.MissingInputs, "host:display_safe_refs") ||
		!objectiveReplanProposalBoundaryContains(contract.Boundaries, "raw_output_not_allowed") {
		t.Fatalf("unsafe recovery contract should require display-safe review: %#v", contract)
	}
}

func objectiveRecoveryEvidenceContains(values []EvidenceRef, want DisplaySafeRef) bool {
	for _, value := range normalizeEvidenceRefs(values) {
		if value.Ref == want {
			return true
		}
	}
	return false
}
