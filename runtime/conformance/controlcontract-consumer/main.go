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
	objectiveRuntimeStatus, err := validateObjectiveRuntimeProjection()
	if err != nil {
		return "", err
	}
	hostAdapterStatus, err := validateHostAdapterIngressProjection()
	if err != nil {
		return "", err
	}
	objectiveCompletionStatus, err := validateObjectiveCompletionContract()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(
		"agentx-controlcontract-ok:%s:%d:%s:%s:%s:%s:%s:%s:%s:%s",
		projection.Status,
		budget.RetryBudgetRemaining,
		lifecycle.To,
		unsafe.FailureClass,
		graphStatus,
		verificationStatus,
		hostEffectStatus,
		objectiveRuntimeStatus,
		hostAdapterStatus,
		objectiveCompletionStatus,
	), nil
}

func validateObjectiveCompletionContract() (string, error) {
	durable := controlcontract.BuildObjectiveRunStoreDurableRequest(controlcontract.ObjectiveRunStoreDurableRequestInput{})
	if durable.ReadyForHostObjectiveRunStore || durable.HostMayPersistObjectiveRun || durable.DurableWriteByCore || durable.RunnerEffect != "none" {
		return "", fmt.Errorf("hostless durable projection did not fail closed: %#v", durable)
	}
	answer := controlcontract.BuildObjectiveFinalAnswer(nil, controlcontract.ObjectiveFinalAnswerInput{EnableSynthesizer: true})
	if answer.ReadyForUser || answer.FailureClass != controlcontract.FailureConfigMissing || answer.NextHostAction != "provide_objective_final_answer_synthesizer" {
		return "", fmt.Errorf("hostless final answer did not require a synthesizer: %#v", answer)
	}
	completion := controlcontract.BuildAutoDelegationAsyncCompletionProjection(controlcontract.AutoDelegationAsyncCompletionInput{})
	if completion.Ready || completion.ReadyForResume || completion.RunnerEffect != "none" {
		return "", fmt.Errorf("hostless async completion did not fail closed: %#v", completion)
	}
	controller := controlcontract.BuildAutoDelegationControllerDecision(controlcontract.AutoDelegationControllerInput{})
	if controller.Ready || controller.HostMayDispatch || controller.RunnerEffect != "none" || controller.RuntimeEffect != "none" {
		return "", fmt.Errorf("hostless delegation controller did not fail closed: %#v", controller)
	}
	strategy := controlcontract.BuildWorkflowStrategyCatalogEntry(controlcontract.WorkflowStrategyCatalogEntryInput{
		WorkflowRef:    "workflow:fixed-consumer",
		CandidateRef:   "strategy:fixed-consumer-workflow",
		CapabilityRefs: []controlcontract.DisplaySafeRef{"capability:fixed-consumer-workflow-runtime"},
	})
	if strategy.Status != controlcontract.VerificationSatisfied || strategy.SourceKind != controlcontract.StrategyCatalogSourceWorkflow {
		return "", fmt.Errorf("workflow strategy metadata is not ready: %#v", strategy)
	}
	return "objective_completion_contract_ready", nil
}

func validateHostAdapterIngressProjection() (string, error) {
	descriptor := controlcontract.ProductionAdapterDescriptor{
		AdapterRef:             "adapter:fixed-consumer-readonly",
		Owner:                  "host",
		OwnerRef:               "host:fixed-consumer",
		Version:                "v1",
		Kind:                   controlcontract.ProductionAdapterSourceReadback,
		SupportedSourceKinds:   []controlcontract.ReplannerSourceKind{controlcontract.ReplannerSourceOperations},
		SupportedCandidateRefs: []controlcontract.DisplaySafeRef{"strategy:fixed-consumer-readonly"},
		ProvidesCapabilityRefs: []controlcontract.DisplaySafeRef{"capability:fixed-consumer-observation"},
		RequiresCapabilityRefs: []controlcontract.DisplaySafeRef{"capability:fixed-consumer-host"},
		InputContractRef:       "contract:fixed-consumer-input",
		OutputContractRef:      "contract:fixed-consumer-output",
		ReadbackContractRef:    "contract:fixed-consumer-readback",
		RequiredPolicyRefs:     []controlcontract.DisplaySafeRef{"policy:fixed-consumer-readonly"},
		RequiredApprovalRefs:   []controlcontract.DisplaySafeRef{"approval:fixed-consumer-readonly"},
		RequiredBudgetRef:      "budget:fixed-consumer-readonly",
		IdempotencyContractRef: "idempotency:fixed-consumer-readonly",
		RiskRef:                "risk:fixed-consumer-readonly",
		SideEffectClass:        "read_only",
		TimeoutPolicyRef:       "timeout:fixed-consumer-short",
		CompensationHandoffRef: "compensation:fixed-consumer-review",
		RedactionPolicyRef:     "redaction:fixed-consumer-display-safe",
		PreflightCheckRefs:     []controlcontract.DisplaySafeRef{"preflight:fixed-consumer-ready"},
		DisplaySafeInputRefs:   []controlcontract.DisplaySafeRef{"input:fixed-consumer-scope"},
		DisplaySafeOutputRefs:  []controlcontract.DisplaySafeRef{"output:fixed-consumer-summary"},
	}
	catalog := controlcontract.BuildProductionAdapterCatalogSnapshot(controlcontract.ProductionAdapterCatalogSnapshotInput{
		CatalogSnapshotRef: "catalog:fixed-consumer-adapters",
		Producer:           "host:fixed-consumer-catalog",
		ProviderRef:        "provider:fixed-consumer-adapters",
		CatalogVersionRef:  "version:fixed-consumer-v1",
		CatalogDigestRef:   "digest:fixed-consumer-v1",
		MaxDescriptorCount: 2,
		HostPolicyRefs:     []controlcontract.DisplaySafeRef{"policy:fixed-consumer-catalog"},
		Descriptors:        []controlcontract.ProductionAdapterDescriptor{descriptor},
	})
	if !catalog.ReadyForHostSelection || catalog.Status != controlcontract.HostActionReady || catalog.RunnerEffect != "none" {
		return "", fmt.Errorf("Host adapter catalog is not ready: %#v", catalog)
	}
	registry := controlcontract.BuildHostAdapterRegistry(controlcontract.HostAdapterRegistryInput{
		RegistryRef: "registry:fixed-consumer-adapters",
		Descriptors: []controlcontract.ProductionAdapterDescriptor{descriptor},
	})
	if !registry.ReadyForRuntimeRequest || registry.Status != controlcontract.HostActionReady || registry.RunnerEffect != "none" {
		return "", fmt.Errorf("Host adapter registry is not ready: %#v", registry)
	}
	ingress := controlcontract.BuildManagedObjectiveIngress(controlcontract.ManagedObjectiveIngressInput{})
	if !ingress.Projected || ingress.ReadyForRuntimeAdapter || ingress.RunnerEffect != "none" {
		return "", fmt.Errorf("hostless managed objective ingress did not fail closed: %#v", ingress)
	}
	return "host_adapter_ingress_contract_ready", nil
}

func validateObjectiveRuntimeProjection() (string, error) {
	policy := controlcontract.BuildAutoDelegationPolicyReview(controlcontract.AutoDelegationPolicy{
		PolicyRef: "policy:fixed-consumer-auto-delegation",
		Enabled:   true,
		Mode:      controlcontract.AutoDelegationObserve,
	})
	if !policy.Ready || policy.Status != controlcontract.VerificationSatisfied || policy.RunnerEffect != "none" {
		return "", fmt.Errorf("auto-delegation policy is not ready: %#v", policy)
	}

	loop := controlcontract.BuildObjectiveRuntimeLoopStep(controlcontract.ObjectiveRuntimeLoopInput{})
	if loop.Available || loop.Status != "inactive" || loop.RunnerEffect != "none" {
		return "", fmt.Errorf("hostless runtime loop did not fail closed: %#v", loop)
	}
	request := controlcontract.BuildHostOwnedObjectiveExecutorStepRequest(controlcontract.HostOwnedObjectiveExecutorStepRequestInput{RuntimeLoop: loop})
	if !request.RequestOnly || request.ReadyForHostExecution || request.Status != controlcontract.HostActionBlocked {
		return "", fmt.Errorf("hostless executor request did not fail closed: %#v", request)
	}
	report := controlcontract.BuildObjectiveRuntimeProductization(controlcontract.ObjectiveRuntimeProductizationInput{RuntimeLoop: loop})
	if report.Available || report.ReadyForHostProductization || report.Status != "inactive" {
		return "", fmt.Errorf("hostless productization did not fail closed: %#v", report)
	}

	normalization := controlcontract.BuildStructuredObservationNormalization(controlcontract.StructuredObservationNormalizationInput{
		Frame:      controlcontract.ObjectiveFrame{ID: "objective:fixed-consumer-runtime"},
		SourceKind: "delegation_worker_result",
		SourceRef:  "result:fixed-consumer-worker",
		Observations: []controlcontract.Observation{{
			Kind:     "metric",
			Source:   "adapter:fixed-consumer-worker",
			Subject:  "objective:fixed-consumer-runtime",
			Strength: controlcontract.EvidenceStrong,
			EvidenceRefs: []controlcontract.EvidenceRef{{
				Ref: "evidence:fixed-consumer-runtime", Kind: "metric", Strength: controlcontract.EvidenceStrong,
				Source: "adapter:fixed-consumer-worker",
			}},
		}},
	})
	if !normalization.ReadyForVerification || normalization.Status != controlcontract.VerificationSatisfied {
		return "", fmt.Errorf("structured observation normalization is not ready: %#v", normalization)
	}
	return "objective_runtime_contract_ready", nil
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
