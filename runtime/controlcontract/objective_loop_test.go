package controlcontract

import "testing"

func TestObjectiveRunDoesNotCreateFullRunOutsideManagedActivation(t *testing.T) {
	for _, activation := range []Activation{ActivationOff, ActivationObserveOnly, ActivationAdvisory} {
		run := BuildObjectiveRun(ObjectiveRunInput{
			Activation: activation,
			Frame: ObjectiveFrame{
				ID:              "objective:inactive",
				UserGoalDigest:  "goal:digest",
				SuccessCriteria: []string{"collect evidence"},
			},
			Ledger: AttemptLedgerPatch{LedgerRef: "ledger:inactive"},
			Budget: ObjectiveBudgetSnapshot{BudgetRef: "budget:inactive", Limit: 3},
			Strategies: []StrategyCandidate{{
				ID:           "strategy:metrics",
				ControlMode:  ControlModeOperations,
				MinIntensity: IntensityL3ManagedObjective,
			}},
		})
		if run.FullRun ||
			run.State != ObjectiveControllerInactive ||
			run.FailureClass != FailurePolicyBlocked ||
			run.RunnerEffect != "none" ||
			run.PromptEffect != "none" ||
			run.NextHostAction != "enable_managed_objective" {
			t.Fatalf("activation %q created unexpected run: %#v", activation, run)
		}
		for _, want := range []Boundary{
			"objective_run_projection_only",
			"no_runner_dispatch",
			"no_runtime_adapter_execution",
			"no_install_or_schedule_apply",
			"objective_controller_activation_required",
		} {
			if !objectiveLoopBoundaryContains(run.Boundaries, want) {
				t.Fatalf("activation %q missing boundary %q: %#v", activation, want, run.Boundaries)
			}
		}

		decision := BuildObjectiveControllerDecision(ObjectiveControllerInput{Run: run})
		if decision.State != ObjectiveControllerInactive ||
			decision.Action != ObjectiveActionNone ||
			decision.RunnerEffect != "none" ||
			decision.PromptEffect != "none" {
			t.Fatalf("activation %q decision = %#v", activation, decision)
		}
	}
}

func TestObjectiveControllerRequiresManagedContractInputs(t *testing.T) {
	decision := BuildObjectiveControllerDecision(ObjectiveControllerInput{
		Activation: ActivationManaged,
		Frame: ObjectiveFrame{
			ID:             "objective:missing",
			UserGoalDigest: "goal:digest",
		},
		Ledger: AttemptLedgerPatch{LedgerRef: "ledger:missing"},
	})
	if decision.State != ObjectiveControllerNeedsContract ||
		decision.Action != ObjectiveActionProvideBudgetPolicy ||
		decision.FailureClass != FailureConfigMissing ||
		decision.NextHostAction != "provide_objective_contract" {
		t.Fatalf("missing contract decision = %#v", decision)
	}
	for _, want := range []MissingInput{"host:success_criteria", "contract:budget"} {
		if !objectiveLoopMissingInputContains(decision.MissingInputs, want) {
			t.Fatalf("expected missing input %q, got %#v", want, decision.MissingInputs)
		}
	}
	if !objectiveLoopBoundaryContains(decision.Boundaries, "host_must_apply_controller_decision") ||
		!objectiveLoopBoundaryContains(decision.Boundaries, "no_strategy_dispatch") {
		t.Fatalf("missing decision boundaries = %#v", decision.Boundaries)
	}
}

func TestObjectiveControllerRequiresApprovalBeforePlanning(t *testing.T) {
	input := objectiveLoopReadyInput()
	input.Approval = ObjectiveApprovalState{Required: true}
	decision := BuildObjectiveControllerDecision(input)
	if decision.State != ObjectiveControllerNeedsAction ||
		decision.Action != ObjectiveActionRequestHostApproval ||
		decision.FailureClass != FailureApprovalRequired ||
		decision.NextHostAction != "request_host_approval" {
		t.Fatalf("approval decision = %#v", decision)
	}
	if !objectiveLoopMissingInputContains(decision.MissingInputs, "host:objective_approval") ||
		!objectiveLoopBoundaryContains(decision.Boundaries, "objective_requires_host_approval") {
		t.Fatalf("approval missing/boundaries = %#v / %#v", decision.MissingInputs, decision.Boundaries)
	}

	input.Approval.Approved = true
	withoutRef := BuildObjectiveControllerDecision(input)
	if withoutRef.State != ObjectiveControllerReviewRequired ||
		withoutRef.Action != ObjectiveActionProvideApprovalRef ||
		withoutRef.FailureClass != FailureEvidenceMissing {
		t.Fatalf("approval without ref decision = %#v", withoutRef)
	}

	input.Approval.ApprovalRefs = []DisplaySafeRef{"approval:objective"}
	ready := BuildObjectiveControllerDecision(input)
	if ready.State != ObjectiveControllerRunning ||
		ready.Action != ObjectiveActionPlanStrategy ||
		ready.NextHostAction != "host_may_select_strategy" ||
		ready.FailureClass != FailureNone {
		t.Fatalf("ready decision = %#v", ready)
	}
}

func TestObjectiveControllerMapsVerificationToNextAction(t *testing.T) {
	partialInput := objectiveLoopReadyInput()
	partialInput.Verification = VerificationResult{
		Status:       VerificationPartial,
		FailureClass: FailureEvidenceMissing,
		EvidenceRefs: []EvidenceRef{{Ref: "evidence:partial", Kind: "metric", Strength: EvidenceWeak}},
	}
	partial := BuildObjectiveControllerDecision(partialInput)
	if partial.State != ObjectiveControllerPartial ||
		partial.Action != ObjectiveActionRequestReplanDecision ||
		partial.FailureClass != FailureEvidenceMissing ||
		partial.NextHostAction != "request_host_replanner_decision" ||
		len(partial.AllowedStrategies) != 1 {
		t.Fatalf("partial decision = %#v", partial)
	}
	if !objectiveLoopDisplaySafeRefContains(partial.DecisionBasis, "verification_status:partial") {
		t.Fatalf("partial decision basis = %#v", partial.DecisionBasis)
	}

	satisfiedInput := objectiveLoopReadyInput()
	satisfiedInput.Verification = VerificationResult{
		Status:       VerificationSatisfied,
		EvidenceRefs: []EvidenceRef{{Ref: "evidence:complete", Kind: "metric", Strength: EvidenceStrong}},
	}
	satisfied := BuildObjectiveControllerDecision(satisfiedInput)
	if satisfied.State != ObjectiveControllerSatisfied ||
		satisfied.Action != ObjectiveActionReturnSatisfied ||
		satisfied.FailureClass != FailureNone ||
		satisfied.NextHostAction != "return_satisfied" ||
		satisfied.Verification.Status != VerificationSatisfied ||
		!satisfied.Verification.Satisfied {
		t.Fatalf("satisfied decision = %#v", satisfied)
	}

	blockedInput := objectiveLoopReadyInput()
	blockedInput.Budget = ObjectiveBudgetSnapshot{BudgetRef: "budget:objective", Limit: 2, Used: 2}
	blocked := BuildObjectiveControllerDecision(blockedInput)
	if blocked.State != ObjectiveControllerBlocked ||
		blocked.Action != ObjectiveActionReturnBlocked ||
		blocked.FailureClass != FailureBudgetExhausted ||
		blocked.NextHostAction != "return_partial_or_request_budget" {
		t.Fatalf("blocked decision = %#v", blocked)
	}
}

func TestObjectiveControllerDowngradesRawOutput(t *testing.T) {
	input := objectiveLoopReadyInput()
	input.RawOutputLoaded = true
	decision := BuildObjectiveControllerDecision(input)
	if decision.State != ObjectiveControllerReviewRequired ||
		decision.Action != ObjectiveActionReviewDisplaySafeRefs ||
		decision.FailureClass != FailureEvidenceWeak ||
		decision.NextHostAction != "provide_display_safe_refs" ||
		!objectiveLoopMissingInputContains(decision.MissingInputs, "host:display_safe_refs") ||
		!objectiveLoopBoundaryContains(decision.Boundaries, "raw_output_not_allowed") {
		t.Fatalf("raw-output decision = %#v", decision)
	}
}

func objectiveLoopReadyInput() ObjectiveControllerInput {
	return ObjectiveControllerInput{
		Activation: ActivationManaged,
		Frame: ObjectiveFrame{
			ID:              "objective:ready",
			UserGoalDigest:  "goal:digest",
			ControlMode:     ControlModeObjective,
			Intensity:       IntensityL3ManagedObjective,
			SuccessCriteria: []string{"collect evidence"},
		},
		Ledger: AttemptLedgerPatch{
			ObjectiveID: "objective:ready",
			LedgerRef:   "ledger:ready",
		},
		Budget: ObjectiveBudgetSnapshot{
			BudgetRef: "budget:objective",
			Limit:     3,
		},
		Approval: ObjectiveApprovalState{
			Required: false,
		},
		Strategies: []StrategyCandidate{{
			ID:           "strategy:metrics",
			Kind:         "host_adapter",
			ControlMode:  ControlModeOperations,
			MinIntensity: IntensityL3ManagedObjective,
			MaxIntensity: IntensityL3ManagedObjective,
			Owner:        "host",
			ExpectedEvidence: []EvidenceRef{{
				Ref:      "evidence:metric_contract",
				Kind:     "metric",
				Strength: EvidenceAdequate,
			}},
		}},
	}
}

func objectiveLoopBoundaryContains(values []Boundary, want Boundary) bool {
	for _, value := range normalizeBoundaries(values) {
		if value == want {
			return true
		}
	}
	return false
}

func objectiveLoopMissingInputContains(values []MissingInput, want MissingInput) bool {
	for _, value := range normalizeMissingInputs(values) {
		if value == want {
			return true
		}
	}
	return false
}

func objectiveLoopDisplaySafeRefContains(values []DisplaySafeRef, want DisplaySafeRef) bool {
	for _, value := range normalizeDisplaySafeRefs(values) {
		if value == want {
			return true
		}
	}
	return false
}
