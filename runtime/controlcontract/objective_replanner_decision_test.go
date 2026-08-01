package controlcontract

import "testing"

func TestObjectiveReplannerDecisionSwitchesStrategyOnlyForManagedL3(t *testing.T) {
	controller := objectiveReplannerTestController(VerificationFailed, FailureVerificationFailed, "strategy:metrics", objectiveReplannerTestStrategies())

	decision := BuildObjectiveReplannerDecision(ObjectiveReplannerDecisionInput{
		Activation:   ActivationManaged,
		Controller:   controller,
		StrategyPlan: StrategyPlannerResult{MaxAllowedIntensity: IntensityL3ManagedObjective},
	})

	if decision.Status != VerificationPartial ||
		decision.Action != ObjectiveReplannerActionSwitchStrategy ||
		decision.CurrentStrategyRef != "strategy:metrics" ||
		decision.NextStrategyRef != "strategy:container_inventory" ||
		decision.NextHostAction != "host_may_switch_strategy" ||
		decision.RunnerEffect != "none" ||
		decision.PromptEffect != "none" {
		t.Fatalf("switch decision = %#v", decision)
	}
	for _, want := range []Boundary{"objective_replanner_decision", "decision_only", "no_strategy_dispatch", "host_must_apply_replanner_decision", "ready_for_host_strategy_switch"} {
		if !objectiveReplannerBoundaryContains(decision.Boundaries, want) {
			t.Fatalf("switch decision missing boundary %q: %#v", want, decision.Boundaries)
		}
	}

	l2SameStrategy := BuildObjectiveReplannerDecision(ObjectiveReplannerDecisionInput{
		Activation:   ActivationManaged,
		Controller:   controller,
		StrategyPlan: StrategyPlannerResult{MaxAllowedIntensity: IntensityL2BoundedToolLoop},
	})
	if l2SameStrategy.Action != ObjectiveReplannerActionRetrySameStrategy ||
		l2SameStrategy.NextStrategyRef != "strategy:metrics" ||
		l2SameStrategy.NextHostAction != "host_may_retry_same_strategy" {
		t.Fatalf("L2 same-strategy decision = %#v", l2SameStrategy)
	}

	l2CrossStrategyOnly := objectiveReplannerTestController(VerificationFailed, FailureVerificationFailed, "strategy:metrics", []StrategyCandidate{objectiveReplannerContainerInventoryStrategy()})
	blocked := BuildObjectiveReplannerDecision(ObjectiveReplannerDecisionInput{
		Activation:   ActivationManaged,
		Controller:   l2CrossStrategyOnly,
		StrategyPlan: StrategyPlannerResult{MaxAllowedIntensity: IntensityL2BoundedToolLoop},
	})
	if blocked.Action != ObjectiveReplannerActionReturnPartial ||
		blocked.FailureClass != FailurePolicyBlocked ||
		blocked.NextHostAction != "request_intensity_upgrade_confirmation" ||
		!objectiveReplannerBoundaryContains(blocked.Boundaries, "l2_cross_strategy_blocked") {
		t.Fatalf("L2 cross-strategy decision = %#v", blocked)
	}
}

func TestObjectiveReplannerDecisionRetriesSameStrategyForPartialEvidence(t *testing.T) {
	controller := objectiveReplannerTestController(VerificationPartial, FailureEvidenceMissing, "strategy:metrics", objectiveReplannerTestStrategies())

	decision := BuildObjectiveReplannerDecision(ObjectiveReplannerDecisionInput{
		Activation:   ActivationManaged,
		Controller:   controller,
		StrategyPlan: StrategyPlannerResult{MaxAllowedIntensity: IntensityL3ManagedObjective},
	})

	if decision.Action != ObjectiveReplannerActionRetrySameStrategy ||
		decision.NextStrategyRef != "strategy:metrics" ||
		decision.NextHostAction != "host_may_retry_same_strategy" ||
		!objectiveReplannerBoundaryContains(decision.Boundaries, "ready_for_host_same_strategy_retry") {
		t.Fatalf("retry decision = %#v", decision)
	}
}

func TestObjectiveReplannerDecisionRoutesCapabilityAndApprovalWithoutApplying(t *testing.T) {
	capability := BuildObjectiveReplannerDecision(ObjectiveReplannerDecisionInput{
		Activation: ActivationManaged,
		Controller: objectiveReplannerTestController(VerificationBlocked, FailureCapabilityMissing, "strategy:metrics", objectiveReplannerTestStrategies()),
	})
	if capability.Action != ObjectiveReplannerActionEnterCapabilityResolution ||
		capability.FailureClass != FailureCapabilityMissing ||
		capability.NextHostAction != "enter_capability_resolution" ||
		!objectiveReplannerBoundaryContains(capability.Boundaries, "capability_gap_proposal_only") ||
		!objectiveReplannerBoundaryContains(capability.Boundaries, "no_install_or_schedule_apply") {
		t.Fatalf("capability decision = %#v", capability)
	}

	approval := BuildObjectiveReplannerDecision(ObjectiveReplannerDecisionInput{
		Activation: ActivationManaged,
		Controller: objectiveReplannerTestController(VerificationReviewRequired, FailureApprovalRequired, "strategy:metrics", objectiveReplannerTestStrategies()),
		Approval:   ObjectiveApprovalState{Required: true},
	})
	if approval.Action != ObjectiveReplannerActionRequestApproval ||
		approval.FailureClass != FailureApprovalRequired ||
		approval.NextHostAction != "request_host_approval" ||
		!objectiveReplannerMissingInputContains(approval.MissingInputs, "host:objective_approval") {
		t.Fatalf("approval decision = %#v", approval)
	}
}

func TestObjectiveReplannerDecisionRoutesCredentialPolicyAndSourceFailures(t *testing.T) {
	credentialController := objectiveReplannerTestController(VerificationBlocked, FailureCredentialMissing, "strategy:metrics", objectiveReplannerTestStrategies())
	credential := BuildObjectiveReplannerDecision(ObjectiveReplannerDecisionInput{
		Activation: ActivationManaged,
		Controller: credentialController,
	})
	if credential.Action != ObjectiveReplannerActionRequestApproval ||
		credential.FailureClass != FailureCredentialMissing ||
		credential.NextHostAction != "request_host_credential" ||
		!objectiveReplannerMissingInputContains(credential.MissingInputs, "host:credential") ||
		!objectiveReplannerBoundaryContains(credential.Boundaries, "objective_replanner_requires_credential") {
		t.Fatalf("credential decision = %#v", credential)
	}

	policyController := objectiveReplannerTestController(VerificationBlocked, FailurePolicyBlocked, "strategy:metrics", objectiveReplannerTestStrategies())
	policyController.EvidenceRefs = nil
	policyController.Verification.EvidenceRefs = nil
	policy := BuildObjectiveReplannerDecision(ObjectiveReplannerDecisionInput{
		Activation: ActivationManaged,
		Controller: policyController,
	})
	if policy.Action != ObjectiveReplannerActionReturnBlocked ||
		policy.NextStrategyRef != "" ||
		policy.FailureClass != FailurePolicyBlocked ||
		policy.NextHostAction != "return_blocked" ||
		!objectiveReplannerBoundaryContains(policy.Boundaries, "objective_replanner_policy_blocked") {
		t.Fatalf("policy decision = %#v", policy)
	}

	sourceUnavailable := BuildObjectiveReplannerDecision(ObjectiveReplannerDecisionInput{
		Activation: ActivationManaged,
		Controller: objectiveReplannerTestController(VerificationBlocked, FailureTargetUnavailable, "strategy:metrics", objectiveReplannerTestStrategies()),
		StrategyPlan: StrategyPlannerResult{
			Activation:          ActivationManaged,
			MaxAllowedIntensity: IntensityL3ManagedObjective,
		},
	})
	if sourceUnavailable.Action != ObjectiveReplannerActionSwitchStrategy ||
		sourceUnavailable.NextStrategyRef != "strategy:container_inventory" ||
		sourceUnavailable.NextHostAction != "host_may_switch_strategy" ||
		!objectiveReplannerBoundaryContains(sourceUnavailable.Boundaries, "objective_replanner_source_unavailable_switch_strategy") {
		t.Fatalf("source-unavailable decision = %#v", sourceUnavailable)
	}
}

func TestObjectiveReplannerDecisionHandlesBudgetAndNoProgress(t *testing.T) {
	controller := objectiveReplannerTestController(VerificationPartial, FailureEvidenceMissing, "strategy:metrics", objectiveReplannerTestStrategies())
	budget := BuildObjectiveReplannerDecision(ObjectiveReplannerDecisionInput{
		Activation: ActivationManaged,
		Controller: controller,
		Budget:     ObjectiveBudgetSnapshot{BudgetRef: "budget:objective", Limit: 2, Used: 2},
		EvidenceRefs: []EvidenceRef{{
			Ref:      "evidence:partial",
			Kind:     "metric",
			Strength: EvidenceAdequate,
			Source:   "adapter:metrics",
		}},
	})
	if budget.Action != ObjectiveReplannerActionReturnPartial ||
		budget.FailureClass != FailureBudgetExhausted ||
		budget.NextHostAction != "return_partial_or_request_budget" ||
		!objectiveReplannerBoundaryContains(budget.Boundaries, "objective_replanner_budget_exhausted") {
		t.Fatalf("budget decision = %#v", budget)
	}

	noProgressController := objectiveReplannerTestController(VerificationFailed, FailureVerificationFailed, "strategy:metrics", []StrategyCandidate{objectiveReplannerMetricStrategy()})
	noProgressController.Verification.EvidenceRefs = nil
	noProgressController.EvidenceRefs = nil
	noProgress := BuildObjectiveReplannerDecision(ObjectiveReplannerDecisionInput{
		Activation: ActivationManaged,
		Controller: noProgressController,
		Attempts: []AttemptSummary{
			{Ref: "attempt:one", StrategyID: "strategy:metrics", Status: VerificationFailed, FailureClass: FailureVerificationFailed},
			{Ref: "attempt:two", StrategyID: "strategy:metrics", Status: VerificationFailed, FailureClass: FailureVerificationFailed},
		},
	})
	if noProgress.Action != ObjectiveReplannerActionReturnBlocked ||
		noProgress.FailureClass != FailureRepeatedNoProgress ||
		noProgress.NextHostAction != "return_blocked" ||
		!objectiveReplannerBoundaryContains(noProgress.Boundaries, "objective_replanner_repeated_no_progress") {
		t.Fatalf("no-progress decision = %#v", noProgress)
	}

	noProgressSwitchController := objectiveReplannerTestController(VerificationFailed, FailureVerificationFailed, "strategy:metrics", objectiveReplannerTestStrategies())
	noProgressSwitchController.Verification.EvidenceRefs = nil
	noProgressSwitchController.EvidenceRefs = nil
	noProgressSwitch := BuildObjectiveReplannerDecision(ObjectiveReplannerDecisionInput{
		Activation: ActivationManaged,
		Controller: noProgressSwitchController,
		StrategyPlan: StrategyPlannerResult{
			Activation:          ActivationManaged,
			MaxAllowedIntensity: IntensityL3ManagedObjective,
		},
		Attempts: []AttemptSummary{
			{Ref: "attempt:one", StrategyID: "strategy:metrics", Status: VerificationFailed, FailureClass: FailureVerificationFailed},
			{Ref: "attempt:two", StrategyID: "strategy:metrics", Status: VerificationFailed, FailureClass: FailureVerificationFailed},
		},
	})
	if noProgressSwitch.Action != ObjectiveReplannerActionSwitchStrategy ||
		noProgressSwitch.FailureClass != FailureRepeatedNoProgress ||
		noProgressSwitch.NextStrategyRef != "strategy:container_inventory" ||
		noProgressSwitch.NextHostAction != "host_may_switch_strategy" ||
		!objectiveReplannerBoundaryContains(noProgressSwitch.Boundaries, "objective_replanner_no_progress_switch_strategy") {
		t.Fatalf("no-progress switch decision = %#v", noProgressSwitch)
	}
}

func TestObjectiveReplannerDecisionStopsL2BudgetExhaustedWithoutEscalation(t *testing.T) {
	controller := objectiveReplannerTestController(VerificationPartial, FailureEvidenceMissing, "strategy:metrics", objectiveReplannerTestStrategies())
	decision := BuildObjectiveReplannerDecision(ObjectiveReplannerDecisionInput{
		Activation: ActivationManaged,
		Controller: controller,
		StrategyPlan: StrategyPlannerResult{
			Activation:          ActivationManaged,
			MaxAllowedIntensity: IntensityL2BoundedToolLoop,
			CurrentStrategyRef:  "strategy:metrics",
			RankedCandidates: []StrategyPlanCandidate{{
				Candidate: objectiveReplannerMetricStrategy(),
				Status:    VerificationSatisfied,
			}},
		},
		Budget: ObjectiveBudgetSnapshot{BudgetRef: "budget:l2", Limit: 1, Used: 1},
		EvidenceRefs: []EvidenceRef{{
			Ref:      "evidence:l2_partial",
			Kind:     "metric",
			Strength: EvidenceAdequate,
			Source:   "adapter:metrics",
		}},
	})
	if decision.Action != ObjectiveReplannerActionReturnPartial ||
		decision.Status != VerificationPartial ||
		decision.FailureClass != FailureBudgetExhausted ||
		decision.NextHostAction != "return_partial_or_request_budget" ||
		decision.RunnerEffect != "none" ||
		decision.PromptEffect != "none" {
		t.Fatalf("L2 budget exhausted decision = %#v", decision)
	}
	for _, want := range []Boundary{
		"objective_replanner_budget_exhausted",
		"no_strategy_dispatch",
		"no_runtime_adapter_execution",
		"no_install_or_schedule_apply",
	} {
		if !objectiveReplannerBoundaryContains(decision.Boundaries, want) {
			t.Fatalf("L2 budget exhausted decision missing boundary %q: %#v", want, decision.Boundaries)
		}
	}
	if decision.NextStrategyRef != "" {
		t.Fatalf("L2 budget exhausted must not select another strategy: %#v", decision)
	}
}

func TestObjectiveReplannerDecisionBlocksOutsideManagedAndRawOutput(t *testing.T) {
	inactive := BuildObjectiveReplannerDecision(ObjectiveReplannerDecisionInput{
		Activation: ActivationAdvisory,
		Controller: objectiveReplannerTestController(VerificationFailed, FailureVerificationFailed, "strategy:metrics", objectiveReplannerTestStrategies()),
	})
	if inactive.Action != ObjectiveReplannerActionNone ||
		inactive.FailureClass != FailurePolicyBlocked ||
		inactive.NextHostAction != "enable_managed_objective" ||
		!objectiveReplannerBoundaryContains(inactive.Boundaries, "objective_replanner_requires_managed_activation") {
		t.Fatalf("inactive decision = %#v", inactive)
	}

	raw := BuildObjectiveReplannerDecision(ObjectiveReplannerDecisionInput{
		Activation:         ActivationManaged,
		CurrentStrategyRef: "secret://example.invalid/token",
		Controller:         objectiveReplannerTestController(VerificationFailed, FailureVerificationFailed, "strategy:metrics", objectiveReplannerTestStrategies()),
	})
	if raw.Action != ObjectiveReplannerActionReviewDisplaySafeRefs ||
		raw.Status != VerificationReviewRequired ||
		raw.FailureClass != FailureEvidenceWeak ||
		raw.NextHostAction != "provide_display_safe_refs" ||
		!objectiveReplannerBoundaryContains(raw.Boundaries, "raw_output_not_allowed") {
		t.Fatalf("raw decision = %#v", raw)
	}

	nestedRawController := objectiveReplannerTestController(VerificationFailed, FailureVerificationFailed, "strategy:metrics", objectiveReplannerTestStrategies())
	nestedRawController.CurrentStrategyRef = "secret://example.invalid/token"
	nestedRaw := BuildObjectiveReplannerDecision(ObjectiveReplannerDecisionInput{
		Activation: ActivationManaged,
		Controller: nestedRawController,
	})
	if nestedRaw.Action != ObjectiveReplannerActionReviewDisplaySafeRefs ||
		nestedRaw.Status != VerificationReviewRequired ||
		nestedRaw.FailureClass != FailureEvidenceWeak ||
		!objectiveReplannerMissingInputContains(nestedRaw.MissingInputs, "host:display_safe_refs") {
		t.Fatalf("nested raw decision = %#v", nestedRaw)
	}
}

func TestObjectiveReplannerDecisionReturnsSatisfiedWithoutReplan(t *testing.T) {
	decision := BuildObjectiveReplannerDecision(ObjectiveReplannerDecisionInput{
		Activation: ActivationManaged,
		Controller: objectiveReplannerTestController(VerificationSatisfied, FailureNone, "strategy:metrics", objectiveReplannerTestStrategies()),
	})
	if decision.Action != ObjectiveReplannerActionReturnSatisfied ||
		decision.Status != VerificationSatisfied ||
		decision.FailureClass != FailureNone ||
		decision.NextHostAction != "return_satisfied" ||
		!objectiveReplannerBoundaryContains(decision.Boundaries, "objective_already_satisfied") {
		t.Fatalf("satisfied decision = %#v", decision)
	}
}

func objectiveReplannerTestController(status VerificationStatus, failure FailureClass, current DisplaySafeRef, strategies []StrategyCandidate) ObjectiveControllerDecision {
	return BuildObjectiveControllerDecision(ObjectiveControllerInput{
		Activation: ActivationManaged,
		Frame: ObjectiveFrame{
			ID:              "objective:replanner",
			UserGoalDigest:  "goal:digest",
			ControlMode:     ControlModeObjective,
			Intensity:       IntensityL3ManagedObjective,
			SuccessCriteria: []string{"collect evidence"},
		},
		Ledger:             AttemptLedgerPatch{ObjectiveID: "objective:replanner", LedgerRef: "ledger:replanner"},
		Budget:             ObjectiveBudgetSnapshot{BudgetRef: "budget:objective", Limit: 3},
		CurrentStrategyRef: current,
		Strategies:         strategies,
		Verification: VerificationResult{
			Status:       status,
			FailureClass: failure,
			EvidenceRefs: []EvidenceRef{{
				Ref:      "evidence:replanner",
				Kind:     "metric",
				Strength: EvidenceAdequate,
				Source:   "adapter:metrics",
			}},
			MissingInputs: objectiveReplannerTestMissingInputs(failure),
		},
	})
}

func objectiveReplannerTestMissingInputs(failure FailureClass) []MissingInput {
	switch failure {
	case FailureCapabilityMissing:
		return []MissingInput{"capability:metric_reader"}
	case FailureCredentialMissing:
		return []MissingInput{"host:credential"}
	case FailureApprovalRequired:
		return []MissingInput{"host:objective_approval"}
	case FailurePolicyBlocked:
		return []MissingInput{"host:policy_review"}
	default:
		return nil
	}
}

func objectiveReplannerTestStrategies() []StrategyCandidate {
	return []StrategyCandidate{
		objectiveReplannerMetricStrategy(),
		objectiveReplannerContainerInventoryStrategy(),
	}
}

func objectiveReplannerContainerInventoryStrategy() StrategyCandidate {
	return StrategyCandidate{
		ID:              "strategy:container_inventory",
		Kind:            "host_adapter",
		ControlMode:     ControlModeObjective,
		MinIntensity:    IntensityL3ManagedObjective,
		MaxIntensity:    IntensityL3ManagedObjective,
		Owner:           "host",
		SideEffectClass: "read_only",
		ExpectedEvidence: []EvidenceRef{{
			Ref:      "evidence:container_inventory",
			Kind:     "container_inventory",
			Strength: EvidenceAdequate,
			Source:   "adapter:container",
		}},
	}
}

func objectiveReplannerMetricStrategy() StrategyCandidate {
	return StrategyCandidate{
		ID:              "strategy:metrics",
		Kind:            "host_adapter",
		ControlMode:     ControlModeOperations,
		MinIntensity:    IntensityL3ManagedObjective,
		MaxIntensity:    IntensityL3ManagedObjective,
		Owner:           "host",
		SideEffectClass: "read_only",
		ExpectedEvidence: []EvidenceRef{{
			Ref:      "evidence:metric",
			Kind:     "metric",
			Strength: EvidenceAdequate,
			Source:   "adapter:metrics",
		}},
	}
}

func objectiveReplannerBoundaryContains(values []Boundary, want Boundary) bool {
	for _, value := range normalizeBoundaries(values) {
		if value == want {
			return true
		}
	}
	return false
}

func objectiveReplannerMissingInputContains(values []MissingInput, want MissingInput) bool {
	for _, value := range normalizeMissingInputs(values) {
		if value == want {
			return true
		}
	}
	return false
}
