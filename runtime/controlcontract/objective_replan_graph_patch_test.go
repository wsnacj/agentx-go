package controlcontract

import "testing"

func TestBuildObjectiveReplanGraphPatchBuildsHostReviewNodesForRecoveryProposal(t *testing.T) {
	report := BuildObjectiveRecoveryContractFromJSON(ObjectiveRecoveryContractJSONDecodeInput{
		RawJSON: []byte(`{
			"answer_contract": {
				"final_answer_recommended": false,
				"recovery_recommended": "True",
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
		ContractRef:        "contract:recovery_patch_case",
		SourceRef:          "source:primary_attempt",
		Producer:           "adapter:primary",
		ObjectiveID:        "objective:recovery_patch",
		CurrentStrategyRef: "strategy:primary",
	})
	if report.Contract.ReplanProposal.Action != ObjectiveReplanProposalActionAddEvidenceNode {
		t.Fatalf("expected recovery proposal, got %#v", report.Contract.ReplanProposal)
	}

	patch := BuildObjectiveReplanGraphPatch(ObjectiveReplanGraphPatchInput{
		PatchRef:       "patch:recovery_case",
		SourceGraphRef: "graph:objective",
		SourceNodeRef:  "node:primary_attempt",
		Proposal:       report.Contract.ReplanProposal,
	})

	if !patch.ReadyForHostReview ||
		patch.ReadyForGraphApply ||
		patch.Status != VerificationPartial ||
		patch.Action != ObjectiveReplanProposalActionAddEvidenceNode ||
		patch.NextHostAction != "review_objective_replan_graph_patch" ||
		len(patch.PatchNodes) != 2 ||
		!objectiveReplanGraphPatchBoundaryContains(patch.Boundaries, "objective_replan_graph_patch_proposal_only") ||
		!objectiveReplanGraphPatchBoundaryContains(patch.Boundaries, "no_graph_mutation_by_core") ||
		!objectiveReplanGraphPatchMissingInputContains(patch.MissingInputs, "host:objective_replan_graph_patch_review") {
		t.Fatalf("unexpected graph patch: %#v", patch)
	}

	first := patch.PatchNodes[0]
	if first.State != ObjectiveNodeStatePending ||
		first.Kind != "objective_replan_recovery_node" ||
		first.CapabilityRef != "capability:lookup_a" ||
		first.StrategyRef != "capability:lookup_a" ||
		first.AttemptPolicy.MaxAttempts != 1 ||
		!first.AttemptPolicy.NoProgressGate ||
		first.SideEffectClass != ObjectiveCapabilitySideEffectUnspecified ||
		!objectiveReplanGraphPatchEvidenceContains(first.RequiredEvidence, "evidence:dimension_a") ||
		!objectiveReplanGraphPatchBoundaryContains(first.Boundaries, "host_must_bind_recovery_node_before_runtime") ||
		!objectiveReplanGraphPatchMissingInputContains(first.MissingInputs, "host:capability_descriptor_binding") {
		t.Fatalf("unexpected first recovery node: %#v", first)
	}
}

func TestBuildObjectiveReplanGraphPatchSkipsTerminalCloseout(t *testing.T) {
	proposal := ObjectiveReplanProposal{
		ContractVersion: ContractVersion,
		Projected:       true,
		ProposalRef:     "proposal:blocked_closeout",
		Status:          VerificationBlocked,
		Action:          ObjectiveReplanProposalActionBlockedCloseout,
		Steps: []ObjectiveReplanProposalStep{{
			StepRef: "replan_step:blocked_closeout",
			Action:  ObjectiveReplanProposalActionBlockedCloseout,
			Owner:   "host",
		}},
		NextHostAction: "return_blocked",
		RunnerEffect:   "none",
		PromptEffect:   "none",
	}.Normalize()

	patch := BuildObjectiveReplanGraphPatch(ObjectiveReplanGraphPatchInput{Proposal: proposal})
	if patch.ReadyForHostReview ||
		patch.Status != VerificationNotApplicable ||
		len(patch.PatchNodes) != 0 ||
		patch.NextHostAction != "return_blocked" ||
		!objectiveReplanGraphPatchBoundaryContains(patch.Boundaries, "objective_replan_graph_patch_not_required") {
		t.Fatalf("terminal closeout should not produce a graph patch: %#v", patch)
	}
}

func TestBuildObjectiveReplanGraphPatchDowngradesUnsafeRefs(t *testing.T) {
	proposal := ObjectiveReplanProposal{
		ContractVersion: ContractVersion,
		Projected:       true,
		ProposalRef:     "secret://example.invalid/token",
		Status:          VerificationPartial,
		Action:          ObjectiveReplanProposalActionAddEvidenceNode,
		Steps: []ObjectiveReplanProposalStep{{
			StepRef: "replan_step:add_evidence",
			Action:  ObjectiveReplanProposalActionAddEvidenceNode,
		}},
	}

	patch := BuildObjectiveReplanGraphPatch(ObjectiveReplanGraphPatchInput{Proposal: proposal})
	if patch.ReadyForHostReview ||
		patch.Status != VerificationReviewRequired ||
		patch.Action != ObjectiveReplanProposalActionReviewRefs ||
		patch.NextHostAction != "provide_display_safe_refs" ||
		!patch.RawOutputLoaded ||
		!objectiveReplanGraphPatchMissingInputContains(patch.MissingInputs, "host:display_safe_refs") ||
		!objectiveReplanGraphPatchBoundaryContains(patch.Boundaries, "raw_output_not_allowed") {
		t.Fatalf("unsafe graph patch input should require display-safe review: %#v", patch)
	}
}

func objectiveReplanGraphPatchBoundaryContains(values []Boundary, want Boundary) bool {
	for _, value := range normalizeBoundaries(values) {
		if value == want {
			return true
		}
	}
	return false
}

func objectiveReplanGraphPatchMissingInputContains(values []MissingInput, want MissingInput) bool {
	for _, value := range normalizeMissingInputs(values) {
		if value == want {
			return true
		}
	}
	return false
}

func objectiveReplanGraphPatchEvidenceContains(values []EvidenceRef, want DisplaySafeRef) bool {
	for _, value := range normalizeEvidenceRefs(values) {
		if value.Ref == want {
			return true
		}
	}
	return false
}
