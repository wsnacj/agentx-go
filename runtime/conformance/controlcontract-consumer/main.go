package main

import (
	"fmt"

	controlcontract "github.com/wsnacj/agentx-go/runtime/controlcontract"
)

func run() (string, error) {
	projection := controlcontract.BuildManagedObjectiveProjection(controlcontract.ManagedObjectiveProjectionInput{
		Activation: controlcontract.ActivationManaged,
		Frame: controlcontract.ObjectiveFrame{
			ID:              "objective:fixed-consumer",
			UserGoalDigest:  "sha256:fixed-consumer",
			SuccessCriteria: []string{"verified output"},
		},
		LedgerRef:    "ledger:fixed-consumer",
		Approved:     true,
		ApprovalRefs: []controlcontract.DisplaySafeRef{"approval:fixed-consumer"},
		PolicyRefs: []controlcontract.DisplaySafeRef{
			"contract:intensity_gate",
			"contract:budget",
			"contract:approval_policy",
			"contract:strategy_scope",
			"contract:redaction_policy",
		},
		AllowedStrategyRefs: []controlcontract.DisplaySafeRef{"strategy:fixed-consumer"},
	})
	if !projection.Ready || projection.Status != controlcontract.HostActionReady {
		return "", fmt.Errorf("projection is not ready: %#v", projection)
	}

	budget := controlcontract.EvaluateRetryBudgetGate(controlcontract.BudgetGateInput{
		Limit:     3,
		Used:      1,
		Increment: 1,
		Scope:     "objective:fixed-consumer",
	})
	if !budget.Allowed {
		return "", fmt.Errorf("budget gate blocked: %#v", budget)
	}

	lifecycle := controlcontract.CheckLifecycleTransition(
		controlcontract.LifecycleStageReady,
		controlcontract.LifecycleStageApplied,
	)
	if !lifecycle.Allowed {
		return "", fmt.Errorf("lifecycle transition blocked: %#v", lifecycle)
	}

	unsafe := controlcontract.VerifyDisplaySafeOnly(false, []string{"https://example.invalid/raw"})
	if unsafe.Satisfied || unsafe.Status != controlcontract.VerificationBlocked {
		return "", fmt.Errorf("unsafe ref was not rejected: %#v", unsafe)
	}

	graphStatus, err := validateObjectiveGraphKernel()
	if err != nil {
		return "", err
	}
	verificationStatus, err := validateObjectiveVerificationKernel()
	if err != nil {
		return "", err
	}
	hostEffectStatus, err := validateHostEffectKernel()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(
		"agentx-controlcontract-ok:%s:%d:%s:%s:%s:%s:%s",
		projection.Status,
		budget.RetryBudgetRemaining,
		lifecycle.To,
		unsafe.FailureClass,
		graphStatus,
		verificationStatus,
		hostEffectStatus,
	), nil
}

func validateHostEffectKernel() (string, error) {
	gate := controlcontract.BuildProductionAdapterIndependentEffectGate(controlcontract.ProductionAdapterIndependentEffectGateSpec{
		Kind:                  controlcontract.ProductionAdapterEffectGateRuntimeExecutor,
		GateRef:               "gate:fixed-consumer-runtime",
		AdapterRef:            "adapter:fixed-consumer-runtime",
		ContractRef:           "contract:fixed-consumer-runtime",
		PolicyRef:             "policy:fixed-consumer-runtime",
		ApprovalRef:           "approval:fixed-consumer-runtime",
		BudgetRef:             "budget:fixed-consumer-runtime",
		IdempotencyRef:        "idempotency:fixed-consumer-runtime",
		ReadbackRef:           "readback:fixed-consumer-runtime",
		EvalRef:               "eval:fixed-consumer-runtime",
		FailureReviewRef:      "review:fixed-consumer-runtime-failure",
		CompensationReviewRef: "review:fixed-consumer-runtime-compensation",
	})
	if !gate.ReadyForIndependentGatePlan || gate.Status != controlcontract.HostActionReady {
		return "", fmt.Errorf("host effect gate is not ready: %#v", gate)
	}

	blocked := controlcontract.BuildProductionAdapterIndependentEffectGatePlan(controlcontract.ProductionAdapterIndependentEffectGatePlanInput{
		PlanRef:                    "plan:fixed-consumer-host-effects",
		AggregateExecutorRequested: true,
		AggregateExecutorRef:       "executor:fixed-consumer-aggregate",
	})
	if blocked.ReadyForIndependentGatePlan || !blocked.AggregateExecutorBlocked || blocked.Status != controlcontract.HostActionBlocked {
		return "", fmt.Errorf("aggregate host effect executor did not fail closed: %#v", blocked)
	}
	return "host_effect_gate_ready", nil
}

func validateObjectiveVerificationKernel() (string, error) {
	frame := controlcontract.ObjectiveFrame{
		ID: "objective:fixed-consumer-verification",
		RequiredEvidence: []controlcontract.EvidenceRef{{
			Ref:      "evidence:verified-metric",
			Kind:     "metric",
			Strength: controlcontract.EvidenceStrong,
			Source:   "adapter:fixed-consumer",
		}},
	}.Normalize()
	normalized := (controlcontract.ObservationNormalizationResult{
		Status: controlcontract.VerificationSatisfied,
		Frame:  frame,
		Observations: []controlcontract.Observation{{
			Kind:     "metric",
			Source:   "adapter:fixed-consumer",
			Subject:  "objective:fixed-consumer-verification",
			Strength: controlcontract.EvidenceStrong,
			EvidenceRefs: []controlcontract.EvidenceRef{{
				Ref:      "evidence:verified-metric",
				Kind:     "metric",
				Strength: controlcontract.EvidenceStrong,
				Source:   "adapter:fixed-consumer",
			}},
		}},
	}).Normalize()
	verification := controlcontract.BuildObjectiveVerificationGate(controlcontract.ObjectiveVerificationGateInput{
		Frame:         frame,
		Normalization: normalized,
	})
	if !verification.Satisfied || verification.Status != controlcontract.VerificationSatisfied {
		return "", fmt.Errorf("objective verification kernel is not satisfied: %#v", verification)
	}

	recovery := controlcontract.BuildObjectiveRecoveryContractFromJSON(controlcontract.ObjectiveRecoveryContractJSONDecodeInput{
		RawJSON:     []byte(`{"answer_contract":{"recovery_recommended":true,"recovery_targets":[{"missing_input":"evidence:detail","suggested_tools":["capability:detail_lookup"]}]}}`),
		ContractRef: "contract:fixed-consumer-recovery",
		ObjectiveID: "objective:fixed-consumer-verification",
	})
	if !recovery.Decoded || !recovery.Contract.Recommended || recovery.Contract.ReplanProposal.Action != controlcontract.ObjectiveReplanProposalActionAddEvidenceNode {
		return "", fmt.Errorf("objective recovery kernel is not ready: %#v", recovery)
	}
	return "objective_verification_recovery_ready", nil
}

func validateObjectiveGraphKernel() (string, error) {
	const (
		capability = controlcontract.DisplaySafeRef("capability:fixed-consumer")
		strategy   = controlcontract.DisplaySafeRef("strategy:fixed-consumer")
	)
	evidence := []controlcontract.EvidenceRef{{
		Ref:      "evidence:fixed-consumer",
		Kind:     "fixed_consumer",
		Strength: controlcontract.EvidenceAdequate,
		Source:   "source:fixed-consumer",
	}}
	spec := controlcontract.ObjectiveSpec{
		SpecRef:               "spec:fixed-consumer",
		ObjectiveID:           "objective:fixed-consumer",
		UserGoalDigest:        "sha256:fixed-consumer",
		RawGoalRef:            "goal:fixed-consumer",
		GoalSummary:           "validate portable objective graph",
		ControlMode:           controlcontract.ControlModeObjective,
		Intensity:             controlcontract.IntensityL3ManagedObjective,
		CandidateCapabilities: []controlcontract.DisplaySafeRef{capability},
		SuccessCriteria: []controlcontract.ObjectiveSuccessCriterion{{
			CriteriaRef:      "criteria:fixed-consumer",
			Text:             "fixed consumer evidence exists",
			RequiredEvidence: evidence,
		}},
		RequiredEvidence:  evidence,
		SideEffectPolicy:  controlcontract.ObjectiveSpecSideEffectReadOnly,
		MissingInfoPolicy: controlcontract.ObjectiveSpecMissingInfoAskUser,
		Budget: controlcontract.ObjectiveSpecBudget{
			BudgetRef:   "budget:fixed-consumer",
			MaxNodes:    1,
			MaxAttempts: 1,
		},
		PolicyRefs: []controlcontract.DisplaySafeRef{"policy:fixed-consumer"},
	}
	catalog := controlcontract.StrategyCatalogSnapshot{
		CatalogRef: "catalog:fixed-consumer",
		Entries: []controlcontract.StrategyCatalogEntry{{
			SourceKind: controlcontract.StrategyCatalogSourceHostAdapter,
			SourceRef:  "source:fixed-consumer",
			Status:     controlcontract.VerificationSatisfied,
			Candidate: controlcontract.StrategyCandidate{
				ID:               string(strategy),
				Kind:             "host_adapter",
				ControlMode:      controlcontract.ControlModeObjective,
				MinIntensity:     controlcontract.IntensityL3ManagedObjective,
				CapabilityRefs:   []controlcontract.DisplaySafeRef{capability},
				SideEffectClass:  string(controlcontract.ObjectiveCapabilitySideEffectReadOnly),
				ExpectedEvidence: evidence,
				Owner:            "host",
			},
		}},
	}
	result := controlcontract.BuildObjectiveGraphValidation(controlcontract.ObjectiveGraphValidationInput{
		Graph: controlcontract.ObjectiveGraph{
			GraphRef:   "graph:fixed-consumer",
			CatalogRef: catalog.CatalogRef,
			Nodes: []controlcontract.ObjectiveNode{{
				NodeRef:             "node:fixed-consumer",
				Kind:                "host_adapter",
				CapabilityRef:       capability,
				StrategyRef:         strategy,
				DescriptorRef:       "descriptor:fixed-consumer",
				SourceRef:           "source:fixed-consumer",
				InputSchemaRef:      "schema:fixed-consumer.input.v1",
				OutputSchemaRef:     "schema:fixed-consumer.output.v1",
				EvidenceContractRef: "evidence:fixed-consumer.contract.v1",
				RequiredEvidence:    evidence,
				AttemptPolicy: controlcontract.ObjectiveNodeAttemptPolicy{
					MaxAttempts:    1,
					TimeoutSeconds: 30,
					NoProgressGate: true,
				},
				SideEffectClass: controlcontract.ObjectiveCapabilitySideEffectReadOnly,
				PolicyRefs:      []controlcontract.DisplaySafeRef{"policy:fixed-consumer"},
			}},
		},
		Spec:    spec,
		Catalog: catalog,
		Policy: controlcontract.ExecutionIntensityPolicy{
			PolicyRef:            "policy:fixed-consumer",
			ExecutionContractRef: "contract:fixed-consumer",
			Activation:           controlcontract.ActivationManaged,
			DefaultIntensity:     controlcontract.IntensityL3ManagedObjective,
			MaxDefaultIntensity:  controlcontract.IntensityL3ManagedObjective,
			MaxAllowedIntensity:  controlcontract.IntensityL3ManagedObjective,
			AllowedControlModesByIntensity: map[controlcontract.ExecutionIntensity][]controlcontract.ControlMode{
				controlcontract.IntensityL3ManagedObjective: {controlcontract.ControlModeObjective},
			},
			PolicyRefs: []controlcontract.DisplaySafeRef{"policy:fixed-consumer"},
		},
	})
	if !result.Validated || !result.ReadyForRuntimeLoop || result.Status != controlcontract.VerificationSatisfied || result.ReadyNodeCount != 1 {
		return "", fmt.Errorf("objective graph kernel is not ready: %#v", result)
	}
	return "objective_graph_ready", nil
}

func main() {
	result, err := run()
	if err != nil {
		panic(err)
	}
	fmt.Println(result)
}
