package controlcontract

// ManagedObjectiveIngressInput is the generic host-facing entry point for
// turning a user request into a managed objective contract. It only consumes
// display-safe host inputs such as a goal digest, refs, policy, budget, and
// catalog metadata; it never parses raw goal text or dispatches execution.
type ManagedObjectiveIngressInput struct {
	Activation                 Activation                  `json:"activation,omitempty"`
	ObjectiveID                string                      `json:"objective_id,omitempty"`
	UserGoalDigest             string                      `json:"user_goal_digest,omitempty"`
	SourceRef                  DisplaySafeRef              `json:"source_ref,omitempty"`
	SuccessCriteria            []string                    `json:"success_criteria,omitempty"`
	Constraints                []string                    `json:"constraints,omitempty"`
	RequiredEvidence           []EvidenceRef               `json:"required_evidence,omitempty"`
	LedgerRef                  DisplaySafeRef              `json:"ledger_ref,omitempty"`
	Policy                     ExecutionIntensityPolicy    `json:"policy,omitempty"`
	Budget                     ObjectiveBudgetSnapshot     `json:"budget,omitempty"`
	UserConfirmed              bool                        `json:"user_confirmed"`
	HostApproved               bool                        `json:"host_approved"`
	ApprovalRefs               []DisplaySafeRef            `json:"approval_refs,omitempty"`
	AllowedStrategyRefs        []DisplaySafeRef            `json:"allowed_strategy_refs,omitempty"`
	StrategyCatalog            StrategyCatalogSnapshot     `json:"strategy_catalog,omitempty"`
	CurrentStrategyRef         DisplaySafeRef              `json:"current_strategy_ref,omitempty"`
	AvailableCapabilityRefs    []DisplaySafeRef            `json:"available_capability_refs,omitempty"`
	AdapterRegistry            HostAdapterRegistrySnapshot `json:"adapter_registry,omitempty"`
	RequestedAdapterRef        DisplaySafeRef              `json:"requested_adapter_ref,omitempty"`
	AllowHostSideEffectAdapter bool                        `json:"allow_host_side_effect_adapter"`
	HostSideEffectApprovalRefs []DisplaySafeRef            `json:"host_side_effect_approval_refs,omitempty"`
	IdempotencyRef             DisplaySafeRef              `json:"idempotency_ref,omitempty"`
	RuntimeInputRefs           []DisplaySafeRef            `json:"runtime_input_refs,omitempty"`
	ExpectedObservationKinds   []string                    `json:"expected_observation_kinds,omitempty"`
	SatisfiedPreconditions     []MissingInput              `json:"satisfied_preconditions,omitempty"`
	Attempts                   []AttemptSummary            `json:"attempts,omitempty"`
	PolicyRefs                 []DisplaySafeRef            `json:"policy_refs,omitempty"`
	DecisionBasis              []DisplaySafeRef            `json:"decision_basis,omitempty"`
	Boundaries                 []Boundary                  `json:"boundaries,omitempty"`
	RawOutputLoaded            bool                        `json:"raw_output_loaded"`
}

type ManagedObjectiveIngressProjection struct {
	ContractVersion           string                         `json:"contract_version,omitempty"`
	Projected                 bool                           `json:"projected"`
	Activation                Activation                     `json:"activation,omitempty"`
	Status                    VerificationStatus             `json:"status,omitempty"`
	ReadyForObjectiveLoop     bool                           `json:"ready_for_objective_loop"`
	ReadyForStrategyPlanning  bool                           `json:"ready_for_strategy_planning"`
	ReadyForStrategyFinalGate bool                           `json:"ready_for_strategy_final_gate"`
	ReadyForRuntimeAdapter    bool                           `json:"ready_for_runtime_adapter"`
	Frame                     ObjectiveFrame                 `json:"frame,omitempty"`
	ManagedObjective          ManagedObjectiveProjection     `json:"managed_objective,omitempty"`
	ObjectiveRun              ObjectiveRun                   `json:"objective_run,omitempty"`
	ControllerDecision        ObjectiveControllerDecision    `json:"controller_decision,omitempty"`
	PreGate                   IntensityGateResult            `json:"pre_gate,omitempty"`
	StrategyFinalGate         IntensityGateResult            `json:"strategy_final_gate,omitempty"`
	StrategyPlan              StrategyPlannerResult          `json:"strategy_plan,omitempty"`
	AdapterRegistry           HostAdapterRegistrySnapshot    `json:"adapter_registry,omitempty"`
	RuntimeAdapterRequest     RuntimeAdapterExecutionRequest `json:"runtime_adapter_request,omitempty"`
	CatalogRef                DisplaySafeRef                 `json:"catalog_ref,omitempty"`
	SelectedStrategyRef       DisplaySafeRef                 `json:"selected_strategy_ref,omitempty"`
	EvidenceRefs              []EvidenceRef                  `json:"evidence_refs,omitempty"`
	FailureClass              FailureClass                   `json:"failure_class,omitempty"`
	MissingInputs             []MissingInput                 `json:"missing_inputs,omitempty"`
	DecisionBasis             []DisplaySafeRef               `json:"decision_basis,omitempty"`
	Boundaries                []Boundary                     `json:"boundaries,omitempty"`
	NextHostAction            NextHostAction                 `json:"next_host_action,omitempty"`
	RunnerEffect              string                         `json:"runner_effect,omitempty"`
	PromptEffect              string                         `json:"prompt_effect,omitempty"`
	RawOutputLoaded           bool                           `json:"raw_output_loaded"`
}

func BuildManagedObjectiveIngress(input ManagedObjectiveIngressInput) ManagedObjectiveIngressProjection {
	activation := NormalizeActivation(string(input.Activation))
	catalog := input.StrategyCatalog.Normalize()
	allowedStrategyRefs := managedObjectiveIngressAllowedStrategyRefs(input.AllowedStrategyRefs, catalog)
	policy := input.Policy.Normalize()
	budget := input.Budget.Normalize()
	policyRefs := managedObjectiveIngressPolicyRefs(input.PolicyRefs, policy)
	evidenceRefs := normalizeEvidenceRefs(input.RequiredEvidence)
	sourceContext := normalizeDisplaySafeRefs([]DisplaySafeRef{input.SourceRef})
	frame := ObjectiveFrame{
		ID:                    managedObjectiveIngressObjectiveID(input.ObjectiveID, input.UserGoalDigest),
		UserGoalDigest:        input.UserGoalDigest,
		ControlMode:           ControlModeObjective,
		Intensity:             IntensityL3ManagedObjective,
		SuccessCriteria:       cloneStringSlice(input.SuccessCriteria),
		Constraints:           cloneStringSlice(input.Constraints),
		RequiredEvidence:      cloneEvidenceRefs(input.RequiredEvidence),
		CandidateCapabilities: cloneDisplaySafeRefs(allowedStrategyRefs),
		SourceContext:         sourceContext,
		Boundaries: []Boundary{
			"managed_objective_ingress_frame",
			"host_supplied_goal_digest",
			"raw_goal_text_not_loaded",
		},
	}.Normalize()
	managed := BuildManagedObjectiveProjection(ManagedObjectiveProjectionInput{
		Activation:          activation,
		Frame:               frame,
		LedgerRef:           input.LedgerRef,
		Attempts:            input.Attempts,
		Approved:            input.HostApproved,
		ApprovalRefs:        input.ApprovalRefs,
		PolicyRefs:          policyRefs,
		AllowedStrategyRefs: allowedStrategyRefs,
		EvidenceRefs:        evidenceRefs,
		Boundaries: MergeBoundaries(
			[]Boundary{
				"managed_objective_ingress",
				"host_owned_objective_entry",
			},
			input.Boundaries,
		),
		RawOutputLoaded: input.RawOutputLoaded,
	})
	preGate := BuildExecutionIntensityPreGate(IntensityGateInput{
		Activation:           activation,
		Policy:               policy,
		RequestedControlMode: ControlModeObjective,
		RequestedIntensity:   IntensityL3ManagedObjective,
		RouteSuggestion: IntensityRouteSuggestion{
			ControlMode:       ControlModeObjective,
			Intensity:         IntensityL3ManagedObjective,
			UserExplicit:      input.UserConfirmed,
			HostRouteExplicit: true,
			ReasonRef:         firstDisplaySafeRef(input.SourceRef, "host:managed_objective_ingress"),
		},
		UserConfirmed:   input.UserConfirmed,
		HostApproved:    input.HostApproved,
		ApprovalRefs:    input.ApprovalRefs,
		Budget:          budget,
		EvidenceRefs:    evidenceRefs,
		DecisionBasis:   input.DecisionBasis,
		Boundaries:      []Boundary{"managed_objective_ingress_pre_gate"},
		RawOutputLoaded: input.RawOutputLoaded,
	})
	strategies := managedObjectiveIngressStrategyCandidates(allowedStrategyRefs, frame.RequiredEvidence)
	run := BuildObjectiveRun(ObjectiveRunInput{
		Activation: activation,
		Frame:      frame,
		Ledger: AttemptLedgerPatch{
			ObjectiveID:     frame.ID,
			LedgerRef:       input.LedgerRef,
			Attempts:        cloneAttemptSummaries(input.Attempts),
			EvidenceRefs:    evidenceRefs,
			Boundaries:      []Boundary{"managed_objective_ingress_ledger"},
			RawOutputLoaded: input.RawOutputLoaded,
		},
		Budget: budget,
		Approval: ObjectiveApprovalState{
			Required:     true,
			Approved:     input.HostApproved,
			ApprovalRefs: input.ApprovalRefs,
			PolicyRefs:   policyRefs,
			Boundaries:   []Boundary{"managed_objective_ingress_approval"},
		},
		PolicyRefs:         policyRefs,
		Strategies:         strategies,
		CurrentStrategyRef: input.CurrentStrategyRef,
		Verification: VerificationResult{
			Status:         VerificationNotEvaluated,
			EvidenceRefs:   evidenceRefs,
			NextHostAction: "host_may_select_strategy",
		},
		EvidenceRefs:    evidenceRefs,
		Boundaries:      []Boundary{"managed_objective_ingress_objective_run"},
		RawOutputLoaded: input.RawOutputLoaded,
	})
	controller := BuildObjectiveControllerDecision(ObjectiveControllerInput{Run: run})
	plan := BuildStrategyPlanner(StrategyPlannerInput{
		Activation:              activation,
		Frame:                   frame,
		Policy:                  policy,
		PreGate:                 preGate,
		Catalog:                 catalog,
		CurrentStrategyRef:      input.CurrentStrategyRef,
		AvailableCapabilityRefs: input.AvailableCapabilityRefs,
		SatisfiedPreconditions:  input.SatisfiedPreconditions,
		Attempts:                input.Attempts,
		Verification:            VerificationResult{Status: VerificationNotEvaluated},
		EvidenceRefs:            evidenceRefs,
		PolicyRefs:              policyRefs,
		DecisionBasis:           input.DecisionBasis,
		Boundaries:              []Boundary{"managed_objective_ingress_strategy_planner"},
		RawOutputLoaded:         input.RawOutputLoaded,
	})
	finalGate := managedObjectiveIngressFinalGate(input, activation, policy, budget, preGate, plan, evidenceRefs)
	adapterRegistry := input.AdapterRegistry.Normalize()
	runtimeRequest := managedObjectiveIngressRuntimeAdapterRequest(input, activation, frame, budget, policyRefs, finalGate, plan, adapterRegistry)
	result := ManagedObjectiveIngressProjection{
		ContractVersion:       ContractVersion,
		Projected:             true,
		Activation:            activation,
		Status:                VerificationBlocked,
		Frame:                 frame,
		ManagedObjective:      managed,
		ObjectiveRun:          run,
		ControllerDecision:    controller,
		PreGate:               preGate,
		StrategyFinalGate:     finalGate,
		StrategyPlan:          plan,
		AdapterRegistry:       adapterRegistry,
		RuntimeAdapterRequest: runtimeRequest,
		CatalogRef:            catalog.CatalogRef,
		EvidenceRefs:          MergeEvidenceRefs(evidenceRefs, managed.EvidenceRefs, preGate.EvidenceRefs, finalGate.EvidenceRefs, plan.EvidenceRefs),
		FailureClass:          FailureNone,
		DecisionBasis: normalizeDisplaySafeRefs(append(
			cloneDisplaySafeRefs(input.DecisionBasis),
			"managed_objective_ingress",
			"host_supplied_goal_digest",
		)),
		Boundaries: MergeBoundaries(
			[]Boundary{
				"managed_objective_ingress",
				"projection_only",
				"no_prompt_effect",
				"no_runner_dispatch",
				"no_strategy_dispatch",
				"no_runtime_adapter_execution",
				"no_tool_execution",
				"no_workflow_dispatch",
				"no_install_or_schedule_apply",
				"core_does_not_parse_goal_text",
			},
			input.Boundaries,
			managed.Boundaries,
			preGate.Boundaries,
			finalGate.Boundaries,
			plan.Boundaries,
			adapterRegistry.Boundaries,
			runtimeRequest.Boundaries,
		),
		RunnerEffect:    "none",
		PromptEffect:    "none",
		RawOutputLoaded: input.RawOutputLoaded || managed.RawOutputLoaded || preGate.RawOutputLoaded || finalGate.RawOutputLoaded || plan.RawOutputLoaded || adapterRegistry.RawOutputLoaded || runtimeRequest.RawOutputLoaded,
	}
	result.ReadyForObjectiveLoop = managed.Ready && preGate.Allowed
	result.ReadyForStrategyPlanning = result.ReadyForObjectiveLoop
	result.ReadyForStrategyFinalGate = result.ReadyForStrategyPlanning &&
		(plan.Status == VerificationSatisfied || plan.Status == VerificationReviewRequired) &&
		plan.Selected.Candidate.ID != "" &&
		finalGate.Allowed
	result.ReadyForRuntimeAdapter = result.ReadyForStrategyFinalGate && runtimeRequest.ReadyForHostExecution
	result.SelectedStrategyRef = normalizeOneDisplaySafeRef(DisplaySafeRef(plan.Selected.Candidate.ID))
	switch {
	case !managed.Ready:
		result.Status = VerificationBlocked
		result.FailureClass = firstFailureClass(managed.FailureClass, FailureConfigMissing)
		result.MissingInputs = AppendMissingInputs(result.MissingInputs, managed.MissingInputs...)
		result.NextHostAction = firstNextHostAction(managed.NextHostAction, "provide_managed_objective_contract")
	case !preGate.Allowed:
		result.Status = VerificationBlocked
		result.FailureClass = firstFailureClass(preGate.FailureClass, FailurePolicyBlocked)
		result.MissingInputs = AppendMissingInputs(result.MissingInputs, preGate.MissingInputs...)
		result.NextHostAction = firstNextHostAction(preGate.NextHostAction, "satisfy_intensity_pre_gate")
	case !managedObjectiveIngressPlanSelected(plan):
		result.Status = plan.Status
		result.FailureClass = firstFailureClass(plan.FailureClass, FailureNone)
		result.MissingInputs = AppendMissingInputs(result.MissingInputs, plan.MissingInputs...)
		result.NextHostAction = firstNextHostAction(plan.NextHostAction, "run_strategy_final_gate")
	case !finalGate.Allowed:
		result.Status = VerificationBlocked
		result.FailureClass = firstFailureClass(finalGate.FailureClass, FailurePolicyBlocked)
		result.MissingInputs = AppendMissingInputs(result.MissingInputs, finalGate.MissingInputs...)
		result.NextHostAction = firstNextHostAction(finalGate.NextHostAction, "run_strategy_final_gate")
	case !runtimeRequest.ReadyForHostExecution:
		result.Status = VerificationBlocked
		result.FailureClass = firstFailureClass(runtimeRequest.FailureClass, FailureHostAdapterMissing)
		result.MissingInputs = AppendMissingInputs(result.MissingInputs, runtimeRequest.MissingInputs...)
		result.NextHostAction = firstNextHostAction(runtimeRequest.NextHostAction, "provide_runtime_adapter_execution_request")
	default:
		result.Status = VerificationSatisfied
		result.FailureClass = FailureNone
		result.NextHostAction = "host_may_execute_runtime_adapter"
		result.Boundaries = AppendBoundaries(result.Boundaries, "adapter_request_ready_not_objective_satisfied")
	}
	return result.Normalize()
}

func CloneManagedObjectiveIngressProjection(in ManagedObjectiveIngressProjection) ManagedObjectiveIngressProjection {
	out := in
	out.Frame = in.Frame.Clone()
	out.ManagedObjective = in.ManagedObjective.Clone()
	out.ObjectiveRun = in.ObjectiveRun.Clone()
	out.ControllerDecision = in.ControllerDecision.Clone()
	out.PreGate = in.PreGate.Clone()
	out.StrategyFinalGate = in.StrategyFinalGate.Clone()
	out.StrategyPlan = in.StrategyPlan.Clone()
	out.AdapterRegistry = in.AdapterRegistry.Clone()
	out.RuntimeAdapterRequest = in.RuntimeAdapterRequest.Clone()
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.DecisionBasis = cloneDisplaySafeRefs(in.DecisionBasis)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (p ManagedObjectiveIngressProjection) Clone() ManagedObjectiveIngressProjection {
	return CloneManagedObjectiveIngressProjection(p)
}

func (p ManagedObjectiveIngressProjection) Normalize() ManagedObjectiveIngressProjection {
	out := CloneManagedObjectiveIngressProjection(p)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Activation = NormalizeActivation(string(out.Activation))
	out.Status = NormalizeVerificationStatus(string(out.Status))
	out.Frame = out.Frame.Normalize()
	out.ManagedObjective = out.ManagedObjective.Normalize()
	out.ObjectiveRun = out.ObjectiveRun.Normalize()
	out.ControllerDecision = out.ControllerDecision.Normalize()
	out.PreGate = out.PreGate.Normalize()
	out.StrategyFinalGate = out.StrategyFinalGate.Normalize()
	out.StrategyPlan = out.StrategyPlan.Normalize()
	out.AdapterRegistry = out.AdapterRegistry.Normalize()
	out.RuntimeAdapterRequest = out.RuntimeAdapterRequest.Normalize()
	out.CatalogRef = normalizeOneDisplaySafeRef(out.CatalogRef)
	out.SelectedStrategyRef = normalizeOneDisplaySafeRef(out.SelectedStrategyRef)
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.DecisionBasis = normalizeDisplaySafeRefs(out.DecisionBasis)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	if out.Status == VerificationNotEvaluated {
		out.Status = VerificationBlocked
	}
	if out.RunnerEffect == "" {
		out.RunnerEffect = "none"
	}
	if out.PromptEffect == "" {
		out.PromptEffect = "none"
	}
	if out.RawOutputLoaded && out.Status != VerificationBlocked {
		out.Status = VerificationReviewRequired
		out.ReadyForObjectiveLoop = false
		out.ReadyForStrategyPlanning = false
		out.ReadyForStrategyFinalGate = false
		out.ReadyForRuntimeAdapter = false
		if out.FailureClass == FailureNone {
			out.FailureClass = FailureEvidenceWeak
		}
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		out.NextHostAction = "provide_display_safe_refs"
	}
	return out
}

func managedObjectiveIngressFinalGate(input ManagedObjectiveIngressInput, activation Activation, policy ExecutionIntensityPolicy, budget ObjectiveBudgetSnapshot, preGate IntensityGateResult, plan StrategyPlannerResult, evidenceRefs []EvidenceRef) IntensityGateResult {
	if !managedObjectiveIngressPlanSelected(plan) {
		return IntensityGateResult{}
	}
	strategy := plan.Selected.Candidate.Normalize()
	return BuildExecutionIntensityFinalGate(IntensityGateInput{
		Activation:           activation,
		Policy:               policy,
		RequestedControlMode: strategy.ControlMode,
		RequestedIntensity:   strategy.MinIntensity,
		RouteSuggestion: IntensityRouteSuggestion{
			ControlMode:       strategy.ControlMode,
			Intensity:         strategy.MinIntensity,
			UserExplicit:      input.UserConfirmed,
			HostRouteExplicit: true,
			ReasonRef:         firstDisplaySafeRef(plan.Selected.SourceRef, input.SourceRef, "host:managed_objective_strategy"),
		},
		UserConfirmed:   input.UserConfirmed,
		HostApproved:    input.HostApproved,
		ApprovalRefs:    input.ApprovalRefs,
		Budget:          budget,
		Strategy:        strategy,
		PreGate:         preGate,
		EvidenceRefs:    evidenceRefs,
		DecisionBasis:   input.DecisionBasis,
		Boundaries:      []Boundary{"managed_objective_ingress_strategy_final_gate"},
		RawOutputLoaded: input.RawOutputLoaded || plan.RawOutputLoaded,
	})
}

func managedObjectiveIngressRuntimeAdapterRequest(input ManagedObjectiveIngressInput, activation Activation, frame ObjectiveFrame, budget ObjectiveBudgetSnapshot, policyRefs []DisplaySafeRef, finalGate IntensityGateResult, plan StrategyPlannerResult, registry HostAdapterRegistrySnapshot) RuntimeAdapterExecutionRequest {
	if !managedObjectiveIngressPlanSelected(plan) {
		return RuntimeAdapterExecutionRequest{}
	}
	return BuildRuntimeAdapterExecutionRequest(RuntimeAdapterExecutionRequestInput{
		Activation:                 activation,
		Frame:                      frame,
		Selected:                   plan.Selected,
		FinalGate:                  finalGate,
		Registry:                   registry,
		RequestedAdapterRef:        input.RequestedAdapterRef,
		Budget:                     budget,
		ApprovalRefs:               input.ApprovalRefs,
		AllowHostSideEffectAdapter: input.AllowHostSideEffectAdapter,
		HostSideEffectApprovalRefs: normalizeDisplaySafeRefs(input.HostSideEffectApprovalRefs),
		PolicyRefs:                 policyRefs,
		AvailableCapabilityRefs:    input.AvailableCapabilityRefs,
		IdempotencyRef:             input.IdempotencyRef,
		InputRefs:                  input.RuntimeInputRefs,
		ExpectedObservationKinds:   input.ExpectedObservationKinds,
		Boundaries: []Boundary{
			"managed_objective_ingress_runtime_adapter_request",
			"adapter_request_not_adapter_execution",
		},
		RawOutputLoaded: input.RawOutputLoaded,
	})
}

func managedObjectiveIngressPlanSelected(plan StrategyPlannerResult) bool {
	normalized := plan.Normalize()
	return normalized.Selected.Candidate.ID != "" &&
		(normalized.Status == VerificationSatisfied || normalized.Status == VerificationReviewRequired)
}

func managedObjectiveIngressObjectiveID(raw string, digest string) string {
	if ref, ok := NormalizeDisplaySafeRef(raw); ok {
		return string(ref)
	}
	if ref, ok := NormalizeDisplaySafeRef("objective:" + normalizeControlToken(digest)); ok {
		return string(ref)
	}
	return ""
}

func managedObjectiveIngressAllowedStrategyRefs(input []DisplaySafeRef, catalog StrategyCatalogSnapshot) []DisplaySafeRef {
	out := normalizeDisplaySafeRefs(input)
	for _, entry := range catalog.Normalize().Entries {
		if ref, ok := NormalizeDisplaySafeRef(entry.Candidate.ID); ok {
			out = appendUniqueDisplaySafeRef(out, ref)
		}
	}
	return normalizeDisplaySafeRefs(out)
}

func managedObjectiveIngressPolicyRefs(input []DisplaySafeRef, policy ExecutionIntensityPolicy) []DisplaySafeRef {
	out := normalizeDisplaySafeRefs(input)
	for _, ref := range []DisplaySafeRef{policy.PolicyRef, policy.ExecutionContractRef} {
		out = appendUniqueDisplaySafeRef(out, ref)
	}
	out = normalizeDisplaySafeRefs(append(out, policy.PolicyRefs...))
	return out
}

func managedObjectiveIngressStrategyCandidates(refs []DisplaySafeRef, evidence []EvidenceRef) []StrategyCandidate {
	out := []StrategyCandidate{}
	for _, ref := range normalizeDisplaySafeRefs(refs) {
		out = append(out, StrategyCandidate{
			ID:               string(ref),
			Kind:             "host_declared_strategy",
			ControlMode:      ControlModeObjective,
			MinIntensity:     IntensityL3ManagedObjective,
			MaxIntensity:     IntensityL3ManagedObjective,
			ExpectedEvidence: cloneEvidenceRefs(evidence),
			Boundaries: []Boundary{
				"host_declared_strategy_scope",
				"host_must_dispatch_strategy",
			},
			RequiresApproval: true,
			Owner:            "host",
		}.Normalize())
	}
	return out
}
