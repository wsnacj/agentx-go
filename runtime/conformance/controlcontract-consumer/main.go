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

	return fmt.Sprintf(
		"agentx-controlcontract-ok:%s:%d:%s:%s:%s",
		projection.Status,
		budget.RetryBudgetRemaining,
		lifecycle.To,
		unsafe.FailureClass,
		graphStatus,
	), nil
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
