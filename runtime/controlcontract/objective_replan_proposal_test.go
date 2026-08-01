package controlcontract

import "testing"

func TestBuildObjectiveReplanProposalAddsEvidenceNodeBeforeRetry(t *testing.T) {
	decision := BuildObjectiveReplannerDecision(ObjectiveReplannerDecisionInput{
		Activation:   ActivationManaged,
		Controller:   objectiveReplannerTestController(VerificationPartial, FailureEvidenceMissing, "strategy:metrics", objectiveReplannerTestStrategies()),
		StrategyPlan: StrategyPlannerResult{MaxAllowedIntensity: IntensityL3ManagedObjective},
	})
	verification := ObjectiveVerificationGateResult{
		Status:       VerificationPartial,
		FailureClass: FailureEvidenceMissing,
		Frame: ObjectiveFrame{
			ID: "objective:replanner",
			RequiredEvidence: []EvidenceRef{{
				Ref:      "evidence:missing_metric",
				Kind:     "metric",
				Strength: EvidenceAdequate,
				Source:   "adapter:metrics",
			}},
		},
		MissingInputs:  []MissingInput{"evidence:missing_metric"},
		NextHostAction: "request_replan_or_return_partial",
	}

	proposal := BuildObjectiveReplanProposal(ObjectiveReplanProposalInput{
		Decision:         decision,
		Verification:     verification,
		AttemptLedgerRef: "ledger:replanner",
	})

	if proposal.Action != ObjectiveReplanProposalActionAddEvidenceNode ||
		proposal.Status != VerificationPartial ||
		proposal.NextOwner != "host" ||
		proposal.NextHostAction != "host_may_add_evidence_node" ||
		len(proposal.Steps) != 2 ||
		proposal.Steps[0].Action != ObjectiveReplanProposalActionAddEvidenceNode ||
		proposal.Steps[0].NextHostAction != "host_may_add_evidence_node" ||
		proposal.Steps[1].Action != ObjectiveReplanProposalActionRetrySameStrategy ||
		!objectiveReplanProposalMissingInputContains(proposal.MissingInputs, "evidence:missing_metric") ||
		!objectiveReplanProposalBoundaryContains(proposal.Boundaries, "no_strategy_dispatch") ||
		!objectiveReplanProposalBoundaryContains(proposal.Boundaries, "no_runtime_adapter_execution") {
		t.Fatalf("unexpected add-evidence proposal = %#v", proposal)
	}
}

func TestBuildObjectiveReplanProposalSwitchesCapabilityWithoutDispatch(t *testing.T) {
	decision := BuildObjectiveReplannerDecision(ObjectiveReplannerDecisionInput{
		Activation:   ActivationManaged,
		Controller:   objectiveReplannerTestController(VerificationFailed, FailureVerificationFailed, "strategy:metrics", objectiveReplannerTestStrategies()),
		StrategyPlan: StrategyPlannerResult{MaxAllowedIntensity: IntensityL3ManagedObjective},
	})

	proposal := BuildObjectiveReplanProposal(ObjectiveReplanProposalInput{Decision: decision})

	if proposal.Action != ObjectiveReplanProposalActionSwitchCapability ||
		proposal.NextStrategyRef != "strategy:container_inventory" ||
		proposal.NextHostAction != "host_may_switch_strategy" ||
		len(proposal.Steps) != 1 ||
		proposal.Steps[0].Action != ObjectiveReplanProposalActionSwitchCapability ||
		proposal.Steps[0].NextStrategy != "strategy:container_inventory" ||
		!objectiveReplanProposalBoundaryContains(proposal.Boundaries, "host_must_apply_replan_proposal") ||
		!objectiveReplanProposalBoundaryContains(proposal.Boundaries, "no_strategy_dispatch") {
		t.Fatalf("unexpected switch proposal = %#v", proposal)
	}
}

func TestBuildObjectiveReplanProposalRoutesCapabilityGapProposalOnly(t *testing.T) {
	decision := BuildObjectiveReplannerDecision(ObjectiveReplannerDecisionInput{
		Activation: ActivationManaged,
		Controller: objectiveReplannerTestController(VerificationBlocked, FailureCapabilityMissing, "strategy:metrics", objectiveReplannerTestStrategies()),
	})

	proposal := BuildObjectiveReplanProposal(ObjectiveReplanProposalInput{Decision: decision})

	if proposal.Action != ObjectiveReplanProposalActionCapabilityGap ||
		proposal.NextHostAction != "enter_capability_resolution" ||
		len(proposal.Steps) != 1 ||
		proposal.Steps[0].Action != ObjectiveReplanProposalActionCapabilityGap ||
		!objectiveReplanProposalDisplayRefContains(proposal.CapabilityGapRefs, "capability:metric_reader") ||
		!objectiveReplanProposalBoundaryContains(proposal.Boundaries, "capability_gap_proposal_only") ||
		!objectiveReplanProposalBoundaryContains(proposal.Boundaries, "no_install_or_schedule_apply") {
		t.Fatalf("unexpected capability gap proposal = %#v", proposal)
	}
}

func TestBuildObjectiveReplanProposalAsksUserBeforeBlockedCloseout(t *testing.T) {
	decision := ObjectiveReplannerDecision{
		ContractVersion: ContractVersion,
		Projected:       true,
		Status:          VerificationBlocked,
		Action:          ObjectiveReplannerActionReturnBlocked,
		ObjectiveID:     "objective:clarify",
		FailureClass:    FailureInsufficientInformation,
		MissingInputs:   []MissingInput{"user:clarify_target"},
		NextHostAction:  "return_blocked",
		RunnerEffect:    "none",
		PromptEffect:    "none",
	}.Normalize()

	proposal := BuildObjectiveReplanProposal(ObjectiveReplanProposalInput{Decision: decision})

	if proposal.Action != ObjectiveReplanProposalActionAskUser ||
		proposal.NextOwner != "user" ||
		proposal.NextHostAction != "ask_user_clarification" ||
		len(proposal.Steps) != 2 ||
		proposal.Steps[0].Action != ObjectiveReplanProposalActionAskUser ||
		proposal.Steps[1].Action != ObjectiveReplanProposalActionBlockedCloseout ||
		!objectiveReplanProposalMissingInputContains(proposal.MissingInputs, "user:clarify_target") {
		t.Fatalf("unexpected ask-user proposal = %#v", proposal)
	}
}

func TestBuildObjectiveReplanProposalReturnsPartialAndBlockedCloseouts(t *testing.T) {
	partialDecision := BuildObjectiveReplannerDecision(ObjectiveReplannerDecisionInput{
		Activation: ActivationManaged,
		Controller: objectiveReplannerTestController(VerificationPartial, FailureEvidenceMissing, "strategy:metrics", objectiveReplannerTestStrategies()),
		Budget:     ObjectiveBudgetSnapshot{BudgetRef: "budget:objective", Limit: 1, Used: 1},
		EvidenceRefs: []EvidenceRef{{
			Ref:      "evidence:partial",
			Kind:     "metric",
			Strength: EvidenceAdequate,
			Source:   "adapter:metrics",
		}},
	})
	partial := BuildObjectiveReplanProposal(ObjectiveReplanProposalInput{Decision: partialDecision})
	if partial.Action != ObjectiveReplanProposalActionPartialCloseout ||
		partial.NextHostAction != "return_partial_or_request_budget" ||
		len(partial.Steps) != 1 ||
		partial.Steps[0].Action != ObjectiveReplanProposalActionPartialCloseout {
		t.Fatalf("unexpected partial closeout proposal = %#v", partial)
	}

	blockedController := objectiveReplannerTestController(VerificationBlocked, FailurePolicyBlocked, "strategy:metrics", objectiveReplannerTestStrategies())
	blockedController.EvidenceRefs = nil
	blockedController.Verification.EvidenceRefs = nil
	blockedDecision := BuildObjectiveReplannerDecision(ObjectiveReplannerDecisionInput{
		Activation: ActivationManaged,
		Controller: blockedController,
	})
	blocked := BuildObjectiveReplanProposal(ObjectiveReplanProposalInput{Decision: blockedDecision})
	if blocked.Action != ObjectiveReplanProposalActionBlockedCloseout ||
		blocked.NextHostAction != "return_blocked" ||
		len(blocked.Steps) != 1 ||
		blocked.Steps[0].Action != ObjectiveReplanProposalActionBlockedCloseout {
		t.Fatalf("unexpected blocked closeout proposal = %#v", blocked)
	}
}

func TestBuildObjectiveReplanProposalDowngradesUnsafeRefs(t *testing.T) {
	proposal := BuildObjectiveReplanProposal(ObjectiveReplanProposalInput{
		ProposalRef: "secret://example.invalid/token",
		Decision:    BuildObjectiveReplannerDecision(ObjectiveReplannerDecisionInput{Activation: ActivationManaged}),
	})
	if proposal.Action != ObjectiveReplanProposalActionReviewRefs ||
		proposal.Status != VerificationReviewRequired ||
		proposal.NextHostAction != "provide_display_safe_refs" ||
		!proposal.RawOutputLoaded ||
		!objectiveReplanProposalMissingInputContains(proposal.MissingInputs, "host:display_safe_refs") ||
		!objectiveReplanProposalBoundaryContains(proposal.Boundaries, "raw_output_not_allowed") {
		t.Fatalf("unexpected unsafe proposal = %#v", proposal)
	}
}

func TestObjectiveReplanProposalActionMetadataRequiresDispatchOnlyForAttemptActions(t *testing.T) {
	tests := []struct {
		action       ObjectiveReplanProposalAction
		wantOwner    string
		wantDispatch bool
		wantUser     bool
		wantApproval bool
		wantTerminal bool
	}{
		{ObjectiveReplanProposalActionRetrySameStrategy, "host", true, false, false, false},
		{ObjectiveReplanProposalActionSwitchCapability, "host", true, false, false, false},
		{ObjectiveReplanProposalActionAddEvidenceNode, "host", true, false, false, false},
		{ObjectiveReplanProposalActionAskUser, "user", false, true, false, false},
		{ObjectiveReplanProposalActionRequestApproval, "host", false, false, true, false},
		{ObjectiveReplanProposalActionCapabilityGap, "host", false, false, false, false},
		{ObjectiveReplanProposalActionPartialCloseout, "host", false, false, false, true},
		{ObjectiveReplanProposalActionBlockedCloseout, "host", false, false, false, true},
		{ObjectiveReplanProposalActionSatisfiedCloseout, "host", false, false, false, true},
		{ObjectiveReplanProposalActionReviewRefs, "host", false, false, false, false},
		{ObjectiveReplanProposalActionNone, "host", false, false, false, false},
	}

	for _, tc := range tests {
		meta := ObjectiveReplanProposalActionMetadataFor(tc.action)
		if meta.Owner != tc.wantOwner ||
			meta.RequiresHostDispatchBinding != tc.wantDispatch ||
			meta.RequiresUserInput != tc.wantUser ||
			meta.RequiresApproval != tc.wantApproval ||
			meta.Terminal != tc.wantTerminal {
			t.Fatalf("metadata mismatch for %s: %#v", tc.action, meta)
		}
	}
}

func TestObjectiveReplanProposalRequiresHostDispatchBinding(t *testing.T) {
	proposal := BuildObjectiveReplanProposal(ObjectiveReplanProposalInput{
		Decision: BuildObjectiveReplannerDecision(ObjectiveReplannerDecisionInput{
			Activation:   ActivationManaged,
			Controller:   objectiveReplannerTestController(VerificationPartial, FailureEvidenceMissing, "strategy:metrics", objectiveReplannerTestStrategies()),
			StrategyPlan: StrategyPlannerResult{MaxAllowedIntensity: IntensityL3ManagedObjective},
		}),
		Verification: ObjectiveVerificationGateResult{
			Status:        VerificationPartial,
			FailureClass:  FailureEvidenceMissing,
			MissingInputs: []MissingInput{"evidence:missing_metric"},
		},
	})
	if !ObjectiveReplanProposalRequiresHostDispatchBinding(&proposal) {
		t.Fatalf("expected add-evidence proposal to require host dispatch binding: %#v", proposal)
	}

	closeout := BuildObjectiveReplanProposal(ObjectiveReplanProposalInput{
		Decision: BuildObjectiveReplannerDecision(ObjectiveReplannerDecisionInput{
			Activation: ActivationManaged,
			Controller: objectiveReplannerTestController(VerificationBlocked, FailurePolicyBlocked, "strategy:metrics", objectiveReplannerTestStrategies()),
		}),
	})
	if ObjectiveReplanProposalRequiresHostDispatchBinding(&closeout) {
		t.Fatalf("blocked closeout should not require host dispatch binding: %#v", closeout)
	}
}

func objectiveReplanProposalBoundaryContains(values []Boundary, want Boundary) bool {
	for _, value := range normalizeBoundaries(values) {
		if value == want {
			return true
		}
	}
	return false
}

func objectiveReplanProposalMissingInputContains(values []MissingInput, want MissingInput) bool {
	for _, value := range normalizeMissingInputs(values) {
		if value == want {
			return true
		}
	}
	return false
}

func objectiveReplanProposalDisplayRefContains(values []DisplaySafeRef, want DisplaySafeRef) bool {
	for _, value := range normalizeDisplaySafeRefs(values) {
		if value == want {
			return true
		}
	}
	return false
}
