package controlcontract

import "testing"

func TestBuildObjectiveSafeDefaultProposalAppliesHostOwnedDefaults(t *testing.T) {
	proposal := BuildObjectiveSafeDefaultProposal(ObjectiveSafeDefaultProposalInput{
		MissingInfoPolicy: ObjectiveSpecMissingInfoDefaultSafe,
		MissingInputs:     []MissingInput{"slot:travel_date"},
		Candidates: []ObjectiveSafeDefaultCandidate{{
			MissingInput: "slot:travel_date",
			DefaultRef:   "default:host_policy_today",
			PolicyRef:    "policy:default_travel_date_today",
			EvidenceRef: EvidenceRef{
				Ref:      "evidence:host_default_policy",
				Kind:     "host_policy",
				Strength: EvidenceAdequate,
				Source:   "host:policy",
			},
		}},
		PolicyRefs: []DisplaySafeRef{"policy:default_travel_date_today"},
	})

	if proposal.Status != VerificationSatisfied ||
		!proposal.ReadyForHostApply ||
		proposal.NextHostAction != "host_may_apply_safe_defaults" ||
		len(proposal.Defaults) != 1 ||
		proposal.Defaults[0].DefaultRef != "default:host_policy_today" ||
		len(proposal.MissingInputs) != 0 ||
		!objectiveReplanPolicyBoundaryContains(proposal.Boundaries, "host_policy_owned_defaults") ||
		!objectiveReplanPolicyBoundaryContains(proposal.Boundaries, "host_must_materialize_default_values") ||
		!objectiveReplanPolicyBoundaryContains(proposal.Boundaries, "no_prompt_heuristic_fallback") {
		t.Fatalf("unexpected safe default proposal = %#v", proposal)
	}
}

func TestBuildObjectiveSafeDefaultProposalAsksUserWhenPolicyRequiresClarification(t *testing.T) {
	proposal := BuildObjectiveSafeDefaultProposal(ObjectiveSafeDefaultProposalInput{
		MissingInfoPolicy: ObjectiveSpecMissingInfoAskUser,
		MissingInputs:     []MissingInput{"slot:target"},
	})

	if proposal.Status != VerificationBlocked ||
		proposal.ReadyForHostApply ||
		proposal.NextHostAction != "ask_user_clarification" ||
		proposal.FailureClass != FailureInsufficientInformation ||
		!objectiveReplanPolicyBoundaryContains(proposal.Boundaries, "missing_info_policy_ask_user") {
		t.Fatalf("unexpected ask-user safe default proposal = %#v", proposal)
	}
}

func TestBuildObjectiveSideEffectSplitProposalKeepsReadOnlyAndBlocksEffects(t *testing.T) {
	graph := ObjectiveGraph{
		GraphRef: "graph:side_effect_split",
		Nodes: []ObjectiveNode{
			objectiveReplanPolicyNode("node:query", ObjectiveCapabilitySideEffectReadOnly),
			objectiveReplanPolicyNode("node:purchase", ObjectiveCapabilitySideEffectPayment),
		},
	}
	spec := ObjectiveSpec{
		SpecRef:          "spec:side_effect_split",
		SideEffectPolicy: ObjectiveSpecSideEffectReadOnly,
	}

	proposal := BuildObjectiveSideEffectSplitProposal(ObjectiveSideEffectSplitProposalInput{
		Spec:  spec,
		Graph: graph,
	})

	if proposal.Status != VerificationPartial ||
		!proposal.ReadyForReadOnlyContinuation ||
		!proposal.BlockedSideEffects ||
		proposal.NextHostAction != "run_read_only_nodes_then_request_approval" ||
		proposal.FailureClass != FailurePolicyBlocked ||
		!objectiveReplanPolicyDisplayRefContains(proposal.ReadOnlyNodeRefs, "node:query") ||
		!objectiveReplanPolicyDisplayRefContains(proposal.BlockedNodeRefs, "node:purchase") ||
		!objectiveReplanPolicyStringContains(proposal.BlockedSideEffectClasses, "payment") ||
		!objectiveReplanPolicyMissingInputContains(proposal.MissingInputs, "host:side_effect_approval_or_read_only_split") ||
		!objectiveReplanPolicyBoundaryContains(proposal.Boundaries, "read_only_prefix_allowed") ||
		!objectiveReplanPolicyBoundaryContains(proposal.Boundaries, "host_must_not_execute_blocked_side_effect_nodes") {
		t.Fatalf("unexpected side-effect split proposal = %#v", proposal)
	}
}

func TestBuildObjectiveNoProgressSwitchGateRequiresRepeatedSameSignature(t *testing.T) {
	gate := BuildObjectiveNoProgressSwitchGate(ObjectiveNoProgressSwitchGateInput{
		CurrentStrategyRef: "strategy:primary",
		Verification: VerificationResult{
			Status:        VerificationFailed,
			FailureClass:  FailureVerificationFailed,
			MissingInputs: []MissingInput{"evidence:target"},
		},
		Candidates: []StrategyCandidate{
			objectiveReplanPolicyStrategy("strategy:primary"),
			objectiveReplanPolicyStrategy("strategy:fallback"),
		},
		Attempts: []AttemptSummary{
			{
				Ref:            "attempt:one",
				StrategyID:     "strategy:primary",
				Status:         VerificationFailed,
				FailureClass:   FailureVerificationFailed,
				MissingInputs:  []MissingInput{"evidence:target"},
				NextHostAction: "",
			},
			{
				Ref:            "attempt:other_signature",
				StrategyID:     "strategy:primary",
				Status:         VerificationFailed,
				FailureClass:   FailureVerificationFailed,
				MissingInputs:  []MissingInput{"evidence:different"},
				NextHostAction: "",
			},
		},
	})
	if gate.RepeatedNoProgress ||
		gate.ReadyForSwitch ||
		len(gate.SupportAttemptRefs) != 1 ||
		gate.NextHostAction != "continue_objective_runtime_loop" ||
		!objectiveReplanPolicyBoundaryContains(gate.Boundaries, "no_progress_gate_not_triggered") {
		t.Fatalf("unexpected untriggered no-progress gate = %#v", gate)
	}

	triggered := BuildObjectiveNoProgressSwitchGate(ObjectiveNoProgressSwitchGateInput{
		CurrentStrategyRef: "strategy:primary",
		Verification: VerificationResult{
			Status:        VerificationFailed,
			FailureClass:  FailureVerificationFailed,
			MissingInputs: []MissingInput{"evidence:target"},
		},
		Candidates: []StrategyCandidate{
			objectiveReplanPolicyStrategy("strategy:primary"),
			objectiveReplanPolicyStrategy("strategy:fallback"),
		},
		Attempts: []AttemptSummary{
			{
				Ref:           "attempt:one",
				StrategyID:    "strategy:primary",
				Status:        VerificationFailed,
				FailureClass:  FailureVerificationFailed,
				MissingInputs: []MissingInput{"evidence:target"},
			},
			{
				Ref:           "attempt:two",
				StrategyID:    "strategy:primary",
				Status:        VerificationFailed,
				FailureClass:  FailureVerificationFailed,
				MissingInputs: []MissingInput{"evidence:target"},
			},
		},
	})
	if !triggered.RepeatedNoProgress ||
		!triggered.ReadyForSwitch ||
		triggered.ReadyForCloseout ||
		triggered.FailureClass != FailureRepeatedNoProgress ||
		triggered.NextHostAction != "host_may_switch_strategy" ||
		len(triggered.SupportAttemptRefs) != 2 ||
		!objectiveReplanPolicyDisplayRefContains(triggered.SwitchCandidateRefs, "strategy:fallback") ||
		!objectiveReplanPolicyBoundaryContains(triggered.Boundaries, "objective_replanner_no_progress_switch_strategy") {
		t.Fatalf("unexpected triggered no-progress gate = %#v", triggered)
	}
}

func objectiveReplanPolicyNode(ref DisplaySafeRef, sideEffect ObjectiveCapabilitySideEffectClass) ObjectiveNode {
	return ObjectiveNode{
		NodeRef:             ref,
		Kind:                "test_node",
		CapabilityRef:       DisplaySafeRef("capability:" + string(ref)),
		InputSchemaRef:      DisplaySafeRef("schema:" + string(ref) + "_input"),
		OutputSchemaRef:     DisplaySafeRef("schema:" + string(ref) + "_output"),
		EvidenceContractRef: DisplaySafeRef("contract:" + string(ref) + "_evidence"),
		RequiredEvidence: []EvidenceRef{{
			Ref:      DisplaySafeRef("evidence:" + string(ref)),
			Kind:     "test",
			Strength: EvidenceAdequate,
			Source:   "adapter:test",
		}},
		SideEffectClass: sideEffect,
	}
}

func objectiveReplanPolicyStrategy(ref string) StrategyCandidate {
	return StrategyCandidate{
		ID:              ref,
		CapabilityRefs:  []DisplaySafeRef{DisplaySafeRef("capability:" + ref)},
		SideEffectClass: "read_only",
	}
}

func objectiveReplanPolicyBoundaryContains(values []Boundary, want Boundary) bool {
	for _, value := range normalizeBoundaries(values) {
		if value == want {
			return true
		}
	}
	return false
}

func objectiveReplanPolicyMissingInputContains(values []MissingInput, want MissingInput) bool {
	for _, value := range normalizeMissingInputs(values) {
		if value == want {
			return true
		}
	}
	return false
}

func objectiveReplanPolicyDisplayRefContains(values []DisplaySafeRef, want DisplaySafeRef) bool {
	for _, value := range normalizeDisplaySafeRefs(values) {
		if value == want {
			return true
		}
	}
	return false
}

func objectiveReplanPolicyStringContains(values []string, want string) bool {
	for _, value := range normalizeControlTokenList(values) {
		if value == want {
			return true
		}
	}
	return false
}
