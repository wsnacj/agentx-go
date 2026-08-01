package controlcontract

import "testing"

func TestExecutionIntensityPreGateHonorsExplicitActivationOff(t *testing.T) {
	gate := BuildExecutionIntensityPreGate(IntensityGateInput{
		Activation:           ActivationOff,
		Policy:               objectiveLoopIntensityPolicy(),
		RequestedControlMode: ControlModeObjective,
		RequestedIntensity:   IntensityL3ManagedObjective,
		UserConfirmed:        true,
		Budget:               ObjectiveBudgetSnapshot{BudgetRef: "budget:objective", Limit: 3},
	})
	if gate.Allowed ||
		gate.Activation != ActivationOff ||
		gate.FailureClass != FailurePolicyBlocked ||
		gate.NextHostAction != "enable_control_plane" ||
		!intensityGateBoundaryContains(gate.Boundaries, "intensity_gate_activation_off") {
		t.Fatalf("explicit off gate = %#v", gate)
	}
}

func TestExecutionIntensityPreGateBlocksModelSuggestedL3WithoutPolicyUpgrade(t *testing.T) {
	policy := objectiveLoopIntensityPolicy()
	policy.AllowModelSuggestedUpgrade = false

	gate := BuildExecutionIntensityPreGate(IntensityGateInput{
		Activation:           ActivationManaged,
		Policy:               policy,
		RequestedControlMode: ControlModeObjective,
		RequestedIntensity:   IntensityL3ManagedObjective,
		RouteSuggestion: IntensityRouteSuggestion{
			ControlMode:    ControlModeObjective,
			Intensity:      IntensityL3ManagedObjective,
			ModelSuggested: true,
			ReasonRef:      "route:objective_hint",
		},
		Budget: ObjectiveBudgetSnapshot{BudgetRef: "budget:objective", Limit: 3},
	})
	if gate.Allowed ||
		gate.Status != VerificationBlocked ||
		gate.FailureClass != FailureApprovalRequired ||
		gate.NextHostAction != "request_intensity_upgrade_confirmation" ||
		gate.Upgrade.RequestedBy != "model" {
		t.Fatalf("model suggested L3 gate = %#v", gate)
	}
	if !intensityGateBoundaryContains(gate.Boundaries, "model_route_is_not_authorization") ||
		!intensityGateMissingInputContains(gate.MissingInputs, "host:intensity_upgrade_confirmation") {
		t.Fatalf("model suggested boundaries/missing = %#v / %#v", gate.Boundaries, gate.MissingInputs)
	}
}

func TestExecutionIntensityPreGateRequiresUserConfirmationForL3(t *testing.T) {
	gate := BuildExecutionIntensityPreGate(IntensityGateInput{
		Activation:           ActivationManaged,
		Policy:               objectiveLoopIntensityPolicy(),
		RequestedControlMode: ControlModeObjective,
		RequestedIntensity:   IntensityL3ManagedObjective,
		RouteSuggestion: IntensityRouteSuggestion{
			ControlMode:  ControlModeObjective,
			Intensity:    IntensityL3ManagedObjective,
			UserExplicit: true,
		},
		Budget: ObjectiveBudgetSnapshot{BudgetRef: "budget:objective", Limit: 3},
	})
	if gate.Allowed ||
		!gate.RequiresUserConfirmation ||
		gate.FailureClass != FailureApprovalRequired ||
		gate.NextHostAction != "request_user_confirmation" {
		t.Fatalf("L3 without user confirmation = %#v", gate)
	}
	if !intensityGateMissingInputContains(gate.MissingInputs, "user:intensity_confirmation") {
		t.Fatalf("missing user confirmation input = %#v", gate.MissingInputs)
	}

	confirmed := BuildExecutionIntensityPreGate(IntensityGateInput{
		Activation:           ActivationManaged,
		Policy:               objectiveLoopIntensityPolicy(),
		RequestedControlMode: ControlModeObjective,
		RequestedIntensity:   IntensityL3ManagedObjective,
		UserConfirmed:        true,
		Budget:               ObjectiveBudgetSnapshot{BudgetRef: "budget:objective", Limit: 3},
	})
	if !confirmed.Allowed ||
		confirmed.Status != VerificationSatisfied ||
		confirmed.ApprovedIntensity != IntensityL3ManagedObjective ||
		confirmed.NextHostAction != "host_may_plan_strategy" {
		t.Fatalf("confirmed L3 gate = %#v", confirmed)
	}
}

func TestExecutionIntensityPreGateStopsWhenUserDeniesL3(t *testing.T) {
	gate := BuildExecutionIntensityPreGate(IntensityGateInput{
		Activation:           ActivationManaged,
		Policy:               objectiveLoopIntensityPolicy(),
		RequestedControlMode: ControlModeObjective,
		RequestedIntensity:   IntensityL3ManagedObjective,
		UserDenied:           true,
		Budget:               ObjectiveBudgetSnapshot{BudgetRef: "budget:objective", Limit: 3},
	})
	if gate.Allowed ||
		gate.FailureClass != FailurePermissionDenied ||
		gate.NextHostAction != "return_partial_or_stop" ||
		!gate.RequiresUserConfirmation ||
		!intensityGateBoundaryContains(gate.Boundaries, "user_denied_intensity_upgrade") {
		t.Fatalf("user denied L3 gate = %#v", gate)
	}
}

func TestExecutionIntensityPreGateRequiresHostApprovalForL4(t *testing.T) {
	input := IntensityGateInput{
		Activation:           ActivationManaged,
		Policy:               objectiveLoopIntensityPolicy(),
		RequestedControlMode: ControlModeOperations,
		RequestedIntensity:   IntensityL4DurableLongRun,
		UserConfirmed:        true,
		Budget:               ObjectiveBudgetSnapshot{BudgetRef: "budget:objective", Limit: 3},
	}
	gate := BuildExecutionIntensityPreGate(input)
	if gate.Allowed ||
		!gate.RequiresHostApproval ||
		gate.FailureClass != FailureApprovalRequired ||
		gate.NextHostAction != "request_host_approval" {
		t.Fatalf("L4 without host approval = %#v", gate)
	}
	if !intensityGateMissingInputContains(gate.MissingInputs, "host:intensity_approval") {
		t.Fatalf("missing host approval = %#v", gate.MissingInputs)
	}

	input.HostApproved = true
	input.ApprovalRefs = []DisplaySafeRef{"approval:l4"}
	approved := BuildExecutionIntensityPreGate(input)
	if !approved.Allowed ||
		approved.ApprovedIntensity != IntensityL4DurableLongRun ||
		approved.FailureClass != FailureNone {
		t.Fatalf("approved L4 gate = %#v", approved)
	}
}

func TestExecutionIntensityPreGateStopsWhenHostDeniesL4(t *testing.T) {
	gate := BuildExecutionIntensityPreGate(IntensityGateInput{
		Activation:           ActivationManaged,
		Policy:               objectiveLoopIntensityPolicy(),
		RequestedControlMode: ControlModeOperations,
		RequestedIntensity:   IntensityL4DurableLongRun,
		UserConfirmed:        true,
		HostDenied:           true,
		Budget:               ObjectiveBudgetSnapshot{BudgetRef: "budget:objective", Limit: 3},
	})
	if gate.Allowed ||
		gate.FailureClass != FailurePermissionDenied ||
		gate.NextHostAction != "return_partial_or_stop" ||
		!gate.RequiresHostApproval ||
		!intensityGateBoundaryContains(gate.Boundaries, "host_denied_intensity_upgrade") {
		t.Fatalf("host denied L4 gate = %#v", gate)
	}
}

func TestExecutionIntensityPreGateBlocksDisabledL5(t *testing.T) {
	policy := objectiveLoopIntensityPolicy()
	policy.DisabledIntensities = []ExecutionIntensity{IntensityL5Autonomous}
	policy.MaxAllowedIntensity = IntensityL5Autonomous

	gate := BuildExecutionIntensityPreGate(IntensityGateInput{
		Activation:           ActivationManaged,
		Policy:               policy,
		RequestedControlMode: ControlModeDelegated,
		RequestedIntensity:   IntensityL5Autonomous,
		UserConfirmed:        true,
		HostApproved:         true,
		ApprovalRefs:         []DisplaySafeRef{"approval:l5"},
		Budget:               ObjectiveBudgetSnapshot{BudgetRef: "budget:objective", Limit: 3},
	})
	if gate.Allowed ||
		gate.FailureClass != FailurePolicyBlocked ||
		gate.NextHostAction != "return_partial_or_request_upgrade" ||
		!intensityGateBoundaryContains(gate.Boundaries, "intensity_disabled_by_policy") {
		t.Fatalf("disabled L5 gate = %#v", gate)
	}
}

func TestExecutionIntensityFinalGateRequiresStrategyAndHonorsPreGate(t *testing.T) {
	policy := objectiveLoopIntensityPolicy()
	pre := BuildExecutionIntensityPreGate(IntensityGateInput{
		Activation:           ActivationManaged,
		Policy:               policy,
		RequestedControlMode: ControlModeObjective,
		RequestedIntensity:   IntensityL3ManagedObjective,
		UserConfirmed:        true,
		Budget:               ObjectiveBudgetSnapshot{BudgetRef: "budget:objective", Limit: 3},
	})
	if !pre.Allowed {
		t.Fatalf("pre gate should be allowed: %#v", pre)
	}

	missing := BuildExecutionIntensityFinalGate(IntensityGateInput{
		Activation:           ActivationManaged,
		Policy:               policy,
		RequestedControlMode: ControlModeObjective,
		RequestedIntensity:   IntensityL3ManagedObjective,
		UserConfirmed:        true,
		Budget:               ObjectiveBudgetSnapshot{BudgetRef: "budget:objective", Limit: 3},
	})
	if missing.Allowed ||
		missing.FailureClass != FailureConfigMissing ||
		missing.NextHostAction != "provide_strategy_candidate" ||
		!intensityGateMissingInputContains(missing.MissingInputs, "host:strategy_candidate") {
		t.Fatalf("missing strategy final gate = %#v", missing)
	}

	strategy := StrategyCandidate{
		ID:              "strategy:metrics",
		ControlMode:     ControlModeOperations,
		MinIntensity:    IntensityL3ManagedObjective,
		MaxIntensity:    IntensityL3ManagedObjective,
		Owner:           "host",
		SideEffectClass: "tool_read_only",
	}
	final := BuildExecutionIntensityFinalGate(IntensityGateInput{
		Activation:           ActivationManaged,
		Policy:               policy,
		RequestedControlMode: ControlModeOperations,
		RequestedIntensity:   IntensityL3ManagedObjective,
		UserConfirmed:        true,
		Budget:               ObjectiveBudgetSnapshot{BudgetRef: "budget:objective", Limit: 3},
		Strategy:             strategy,
		PreGate:              pre,
	})
	if !final.Allowed ||
		final.Stage != IntensityGateFinal ||
		final.StrategyRef != "strategy:metrics" ||
		final.ApprovedIntensity != IntensityL3ManagedObjective ||
		final.ApprovedControlMode != ControlModeOperations ||
		final.RunnerEffect != "none" ||
		final.PromptEffect != "none" {
		t.Fatalf("final gate = %#v", final)
	}

	l2Policy := objectiveLoopIntensityPolicy()
	l2Policy.MaxAllowedIntensity = IntensityL2BoundedToolLoop
	l2Pre := BuildExecutionIntensityPreGate(IntensityGateInput{
		Activation:           ActivationManaged,
		Policy:               l2Policy,
		RequestedControlMode: ControlModeTool,
		RequestedIntensity:   IntensityL2BoundedToolLoop,
	})
	if !l2Pre.Allowed {
		t.Fatalf("L2 pre gate should be allowed: %#v", l2Pre)
	}
	blocked := BuildExecutionIntensityFinalGate(IntensityGateInput{
		Activation:           ActivationManaged,
		Policy:               policy,
		RequestedControlMode: ControlModeOperations,
		RequestedIntensity:   IntensityL3ManagedObjective,
		UserConfirmed:        true,
		Budget:               ObjectiveBudgetSnapshot{BudgetRef: "budget:objective", Limit: 3},
		Strategy:             strategy,
		PreGate:              l2Pre,
	})
	if blocked.Allowed ||
		blocked.FailureClass != FailurePolicyBlocked ||
		blocked.NextHostAction != "request_intensity_upgrade_confirmation" ||
		!intensityGateBoundaryContains(blocked.Boundaries, "strategy_intensity_exceeds_pre_gate") {
		t.Fatalf("final gate over pre gate = %#v", blocked)
	}
}

func TestExecutionIntensityFinalGateBlocksDeniedSideEffect(t *testing.T) {
	policy := objectiveLoopIntensityPolicy()
	policy.DeniedSideEffectsByIntensity = map[ExecutionIntensity][]string{
		IntensityL3ManagedObjective: {"install_or_enable_capability"},
	}

	final := BuildExecutionIntensityFinalGate(IntensityGateInput{
		Activation:           ActivationManaged,
		Policy:               policy,
		RequestedControlMode: ControlModeObjective,
		RequestedIntensity:   IntensityL3ManagedObjective,
		UserConfirmed:        true,
		Budget:               ObjectiveBudgetSnapshot{BudgetRef: "budget:objective", Limit: 3},
		Strategy: StrategyCandidate{
			ID:              "strategy:installer",
			ControlMode:     ControlModeCapabilityResolution,
			MinIntensity:    IntensityL3ManagedObjective,
			SideEffectClass: "install_or_enable_capability",
			Owner:           "host",
		},
	})
	if final.Allowed ||
		final.FailureClass != FailurePolicyBlocked ||
		final.NextHostAction != "select_lower_risk_strategy" ||
		!intensityGateBoundaryContains(final.Boundaries, "strategy_side_effect_denied") {
		t.Fatalf("denied side effect final gate = %#v", final)
	}
}

func TestExecutionIntensityGateRejectsMissingPolicyRefs(t *testing.T) {
	gate := BuildExecutionIntensityPreGate(IntensityGateInput{
		Activation:           ActivationManaged,
		Policy:               ExecutionIntensityPolicy{MaxAllowedIntensity: IntensityL3ManagedObjective},
		RequestedControlMode: ControlModeObjective,
		RequestedIntensity:   IntensityL3ManagedObjective,
		UserConfirmed:        true,
		Budget:               ObjectiveBudgetSnapshot{BudgetRef: "budget:objective", Limit: 3},
	})
	if gate.Allowed ||
		gate.FailureClass != FailurePolicyBlocked ||
		!intensityGateMissingInputContains(gate.MissingInputs, "contract:intensity_policy") ||
		!intensityGateMissingInputContains(gate.MissingInputs, "contract:execution_contract") {
		t.Fatalf("missing policy refs gate = %#v", gate)
	}
}

func objectiveLoopIntensityPolicy() ExecutionIntensityPolicy {
	return ExecutionIntensityPolicy{
		PolicyRef:                   "policy:intensity",
		ExecutionContractRef:        "contract:execution",
		Activation:                  ActivationManaged,
		DefaultIntensity:            IntensityL0AnswerOnly,
		MaxDefaultIntensity:         IntensityL2BoundedToolLoop,
		MaxAllowedIntensity:         IntensityL4DurableLongRun,
		AllowModelSuggestedUpgrade:  false,
		RequireUserConfirmationFrom: IntensityL3ManagedObjective,
		RequireHostApprovalFrom:     IntensityL4DurableLongRun,
		AllowedControlModesByIntensity: map[ExecutionIntensity][]ControlMode{
			IntensityL3ManagedObjective: {ControlModeObjective, ControlModeOperations, ControlModeCapabilityResolution},
			IntensityL4DurableLongRun:   {ControlModeOperations, ControlModeWorkflow},
		},
		BudgetRefs: []DisplaySafeRef{"budget:objective"},
		PolicyRefs: []DisplaySafeRef{"policy:intensity", "contract:execution"},
	}
}

func intensityGateBoundaryContains(values []Boundary, want Boundary) bool {
	for _, value := range normalizeBoundaries(values) {
		if value == want {
			return true
		}
	}
	return false
}

func intensityGateMissingInputContains(values []MissingInput, want MissingInput) bool {
	for _, value := range normalizeMissingInputs(values) {
		if value == want {
			return true
		}
	}
	return false
}
