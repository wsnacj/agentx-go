package controlcontract

// agentx-api: internal_candidate
type ProductionAdapterObjectiveCloseoutWriterBlackboxFixtureInput struct {
	FixtureRef      DisplaySafeRef                                `json:"fixture_ref,omitempty"`
	WriterOptIn     ProductionAdapterObjectiveCloseoutWriterOptIn `json:"writer_opt_in,omitempty"`
	RawOutputLoaded bool                                          `json:"raw_output_loaded"`
}

type ProductionAdapterObjectiveCloseoutWriterBlackboxFixture struct {
	ContractVersion                 string                                       `json:"contract_version,omitempty"`
	Projected                       bool                                         `json:"projected"`
	Available                       bool                                         `json:"available"`
	Status                          string                                       `json:"status,omitempty"`
	Mode                            string                                       `json:"mode,omitempty"`
	DisplayState                    string                                       `json:"display_state,omitempty"`
	DisplaySections                 []string                                     `json:"display_sections,omitempty"`
	ReadyForHostDisplay             bool                                         `json:"ready_for_host_display"`
	ReadyForWriterPlanDisplay       bool                                         `json:"ready_for_writer_plan_display"`
	ReadyForWriterDryRunDisplay     bool                                         `json:"ready_for_writer_dry_run_display"`
	ReadyForDurableWriteDisplay     bool                                         `json:"ready_for_durable_write_display"`
	BlockedDisplay                  bool                                         `json:"blocked_display"`
	HostMayExecuteDurableWrite      bool                                         `json:"host_may_execute_durable_write"`
	RequestedMode                   ProductionAdapterObjectiveCloseoutWriterMode `json:"requested_mode,omitempty"`
	ExplicitOptIn                   bool                                         `json:"explicit_opt_in"`
	PlanOnly                        bool                                         `json:"plan_only"`
	DryRun                          bool                                         `json:"dry_run"`
	DurableWriteMode                bool                                         `json:"durable_write_mode"`
	HostWriterAvailable             bool                                         `json:"host_writer_available"`
	HostReadbackAvailable           bool                                         `json:"host_readback_available"`
	HostRollbackReviewAvailable     bool                                         `json:"host_rollback_review_available"`
	HostCompensationReviewAvailable bool                                         `json:"host_compensation_review_available"`
	CoreInvocationExecuted          bool                                         `json:"core_invocation_executed"`
	DurableWriteByCore              bool                                         `json:"durable_write_by_core"`
	ObjectiveStoreWriteByCore       bool                                         `json:"objective_store_write_by_core"`
	RunstoreWriteByCore             bool                                         `json:"runstore_write_by_core"`
	FixtureRef                      DisplaySafeRef                               `json:"fixture_ref,omitempty"`
	WriterOptInRef                  DisplaySafeRef                               `json:"writer_opt_in_ref,omitempty"`
	WriterRef                       DisplaySafeRef                               `json:"writer_ref,omitempty"`
	OwnerRef                        DisplaySafeRef                               `json:"owner_ref,omitempty"`
	HostWriterBindingRef            DisplaySafeRef                               `json:"host_writer_binding_ref,omitempty"`
	ObjectiveCloseoutHandoffRef     DisplaySafeRef                               `json:"objective_closeout_handoff_ref,omitempty"`
	HostUIHandoffRef                DisplaySafeRef                               `json:"host_ui_handoff_ref,omitempty"`
	ObjectiveCloseoutPacketRef      DisplaySafeRef                               `json:"objective_closeout_packet_ref,omitempty"`
	ObjectiveRef                    DisplaySafeRef                               `json:"objective_ref,omitempty"`
	HostObjectiveLifecycleRef       DisplaySafeRef                               `json:"host_objective_lifecycle_ref,omitempty"`
	HostRunstoreRef                 DisplaySafeRef                               `json:"host_runstore_ref,omitempty"`
	ExpectedDurableEventRef         DisplaySafeRef                               `json:"expected_durable_event_ref,omitempty"`
	ExpectedObjectiveStateRef       DisplaySafeRef                               `json:"expected_objective_state_ref,omitempty"`
	HostDurableApplyConfirmationRef DisplaySafeRef                               `json:"host_durable_apply_confirmation_ref,omitempty"`
	AvailableCapabilityRefs         []DisplaySafeRef                             `json:"available_capability_refs,omitempty"`
	RequiredCapabilityRefs          []DisplaySafeRef                             `json:"required_capability_refs,omitempty"`
	PolicyRefs                      []DisplaySafeRef                             `json:"policy_refs,omitempty"`
	RequiredPolicyRefs              []DisplaySafeRef                             `json:"required_policy_refs,omitempty"`
	ApprovalRefs                    []DisplaySafeRef                             `json:"approval_refs,omitempty"`
	RequiredApprovalRefs            []DisplaySafeRef                             `json:"required_approval_refs,omitempty"`
	BudgetRef                       DisplaySafeRef                               `json:"budget_ref,omitempty"`
	RequiredBudgetRef               DisplaySafeRef                               `json:"required_budget_ref,omitempty"`
	IdempotencyRef                  DisplaySafeRef                               `json:"idempotency_ref,omitempty"`
	IdempotencyContractRef          DisplaySafeRef                               `json:"idempotency_contract_ref,omitempty"`
	DryRunPlanRef                   DisplaySafeRef                               `json:"dry_run_plan_ref,omitempty"`
	DryRunResultRef                 DisplaySafeRef                               `json:"dry_run_result_ref,omitempty"`
	DryRunContractRef               DisplaySafeRef                               `json:"dry_run_contract_ref,omitempty"`
	ExpectedReadbackRef             DisplaySafeRef                               `json:"expected_readback_ref,omitempty"`
	ReadbackContractRef             DisplaySafeRef                               `json:"readback_contract_ref,omitempty"`
	RollbackReviewRef               DisplaySafeRef                               `json:"rollback_review_ref,omitempty"`
	RequiredRollbackReviewRef       DisplaySafeRef                               `json:"required_rollback_review_ref,omitempty"`
	CompensationReviewRef           DisplaySafeRef                               `json:"compensation_review_ref,omitempty"`
	RequiredCompensationReviewRef   DisplaySafeRef                               `json:"required_compensation_review_ref,omitempty"`
	RedactionPolicyRef              DisplaySafeRef                               `json:"redaction_policy_ref,omitempty"`
	TimeoutPolicyRef                DisplaySafeRef                               `json:"timeout_policy_ref,omitempty"`
	MissingInputs                   []MissingInput                               `json:"missing_inputs,omitempty"`
	BlockedReasons                  []string                                     `json:"blocked_reasons,omitempty"`
	FailureClass                    FailureClass                                 `json:"failure_class,omitempty"`
	Boundaries                      []Boundary                                   `json:"boundaries,omitempty"`
	NextHostAction                  NextHostAction                               `json:"next_host_action,omitempty"`
	RunnerEffect                    string                                       `json:"runner_effect,omitempty"`
	PromptEffect                    string                                       `json:"prompt_effect,omitempty"`
	RawOutputLoaded                 bool                                         `json:"raw_output_loaded"`
}

// agentx-api: internal_candidate
func BuildProductionAdapterObjectiveCloseoutWriterBlackboxFixture(input ProductionAdapterObjectiveCloseoutWriterBlackboxFixtureInput) ProductionAdapterObjectiveCloseoutWriterBlackboxFixture {
	if productionAdapterObjectiveCloseoutWriterOptInEmpty(input.WriterOptIn) {
		return unavailableProductionAdapterObjectiveCloseoutWriterBlackboxFixture()
	}
	optIn := input.WriterOptIn.Normalize()
	result := ProductionAdapterObjectiveCloseoutWriterBlackboxFixture{
		ContractVersion:                 ContractVersion,
		Projected:                       true,
		Available:                       optIn.Available,
		Status:                          "blocked",
		Mode:                            "production_adapter_objective_closeout_writer_blackbox_fixture",
		DisplayState:                    productionAdapterObjectiveCloseoutWriterFixtureDisplayState(optIn),
		DisplaySections:                 productionAdapterObjectiveCloseoutWriterFixtureDisplaySections(),
		RequestedMode:                   optIn.RequestedMode,
		ExplicitOptIn:                   optIn.ExplicitOptIn,
		PlanOnly:                        optIn.PlanOnly,
		DryRun:                          optIn.DryRun,
		DurableWriteMode:                optIn.DurableWriteMode,
		HostWriterAvailable:             optIn.HostWriterAvailable,
		HostReadbackAvailable:           optIn.HostReadbackAvailable,
		HostRollbackReviewAvailable:     optIn.HostRollbackReviewAvailable,
		HostCompensationReviewAvailable: optIn.HostCompensationReviewAvailable,
		HostMayExecuteDurableWrite:      optIn.HostMayExecuteDurableWrite,
		FixtureRef:                      normalizeOneDisplaySafeRef(input.FixtureRef),
		WriterOptInRef:                  optIn.WriterOptInRef,
		WriterRef:                       optIn.WriterRef,
		OwnerRef:                        optIn.OwnerRef,
		HostWriterBindingRef:            optIn.HostWriterBindingRef,
		ObjectiveCloseoutHandoffRef:     optIn.ObjectiveCloseoutHandoffRef,
		HostUIHandoffRef:                optIn.HostUIHandoffRef,
		ObjectiveCloseoutPacketRef:      optIn.ObjectiveCloseoutPacketRef,
		ObjectiveRef:                    optIn.ObjectiveRef,
		HostObjectiveLifecycleRef:       optIn.HostObjectiveLifecycleRef,
		HostRunstoreRef:                 optIn.HostRunstoreRef,
		ExpectedDurableEventRef:         optIn.ExpectedDurableEventRef,
		ExpectedObjectiveStateRef:       optIn.ExpectedObjectiveStateRef,
		HostDurableApplyConfirmationRef: optIn.HostDurableApplyConfirmationRef,
		AvailableCapabilityRefs:         cloneDisplaySafeRefs(optIn.AvailableCapabilityRefs),
		RequiredCapabilityRefs:          cloneDisplaySafeRefs(optIn.RequiredCapabilityRefs),
		PolicyRefs:                      cloneDisplaySafeRefs(optIn.PolicyRefs),
		RequiredPolicyRefs:              cloneDisplaySafeRefs(optIn.RequiredPolicyRefs),
		ApprovalRefs:                    cloneDisplaySafeRefs(optIn.ApprovalRefs),
		RequiredApprovalRefs:            cloneDisplaySafeRefs(optIn.RequiredApprovalRefs),
		BudgetRef:                       optIn.BudgetRef,
		RequiredBudgetRef:               optIn.RequiredBudgetRef,
		IdempotencyRef:                  optIn.IdempotencyRef,
		IdempotencyContractRef:          optIn.IdempotencyContractRef,
		DryRunPlanRef:                   optIn.DryRunPlanRef,
		DryRunResultRef:                 optIn.DryRunResultRef,
		DryRunContractRef:               optIn.DryRunContractRef,
		ExpectedReadbackRef:             optIn.ExpectedReadbackRef,
		ReadbackContractRef:             optIn.ReadbackContractRef,
		RollbackReviewRef:               optIn.RollbackReviewRef,
		RequiredRollbackReviewRef:       optIn.RequiredRollbackReviewRef,
		CompensationReviewRef:           optIn.CompensationReviewRef,
		RequiredCompensationReviewRef:   optIn.RequiredCompensationReviewRef,
		RedactionPolicyRef:              optIn.RedactionPolicyRef,
		TimeoutPolicyRef:                optIn.TimeoutPolicyRef,
		MissingInputs:                   cloneMissingInputs(optIn.MissingInputs),
		BlockedReasons:                  cloneStringSlice(optIn.BlockedReasons),
		FailureClass:                    optIn.FailureClass,
		Boundaries:                      productionAdapterObjectiveCloseoutWriterBlackboxFixtureBoundaries(optIn.Boundaries),
		NextHostAction:                  optIn.NextHostAction,
		RunnerEffect:                    "none",
		PromptEffect:                    "none",
		RawOutputLoaded:                 input.RawOutputLoaded || optIn.RawOutputLoaded,
	}
	if input.RawOutputLoaded || displaySafeRefRejected(input.FixtureRef) || productionAdapterObjectiveCloseoutWriterOptInOutputUnsafe(optIn) {
		result = productionAdapterObjectiveCloseoutWriterBlackboxFixtureBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if result.FixtureRef == "" {
		result = productionAdapterObjectiveCloseoutWriterBlackboxFixtureBlock(result, FailureEvidenceMissing, "objective_closeout_writer_fixture_ref_missing", "host:objective_closeout_writer_fixture_ref", "provide_objective_closeout_writer_fixture")
		return result.Normalize()
	}
	if !optIn.Available || optIn.WriterOptInRef == "" {
		result = productionAdapterObjectiveCloseoutWriterBlackboxFixtureBlock(result, firstFailureClass(optIn.FailureClass, FailureEvidenceMissing), "objective_closeout_writer_opt_in_not_ready", "host:objective_closeout_writer_opt_in", firstNextHostAction(optIn.NextHostAction, "review_objective_closeout_writer_opt_in"))
		return result.Normalize()
	}
	if optIn.WriterRef == "" || optIn.ObjectiveCloseoutHandoffRef == "" || optIn.HostUIHandoffRef == "" {
		result = productionAdapterObjectiveCloseoutWriterBlackboxFixtureBlock(result, FailureEvidenceMissing, "objective_closeout_writer_display_refs_missing", "host:objective_closeout_writer_display_refs", "review_objective_closeout_writer_opt_in")
		return result.Normalize()
	}
	if optIn.ReadyForHostWriterPlan {
		result.Status = "ready_for_writer_plan_display"
		result.DisplayState = "plan_only"
		result.ReadyForHostDisplay = true
		result.ReadyForWriterPlanDisplay = true
		result.HostMayExecuteDurableWrite = false
		result.NextHostAction = "review_objective_closeout_writer_plan"
		result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_writer_plan_display", "durable_write_not_enabled", "host_cli_objective_closeout_writer_display_ready")
		return result.Normalize()
	}
	if optIn.ReadyForHostWriterDryRun {
		result.Status = "ready_for_writer_dry_run_display"
		result.DisplayState = "dry_run_ready"
		result.ReadyForHostDisplay = true
		result.ReadyForWriterDryRunDisplay = true
		result.HostMayExecuteDurableWrite = false
		result.NextHostAction = "host_may_run_objective_closeout_writer_dry_run"
		result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_writer_dry_run_display", "durable_write_not_enabled", "host_cli_objective_closeout_writer_display_ready")
		return result.Normalize()
	}
	if optIn.ReadyForHostDurableWrite {
		result.Status = "ready_for_durable_write_display"
		result.DisplayState = "durable_write_ready"
		result.ReadyForHostDisplay = true
		result.ReadyForDurableWriteDisplay = true
		result.HostMayExecuteDurableWrite = optIn.HostMayExecuteDurableWrite
		result.NextHostAction = "host_may_execute_objective_closeout_durable_writer"
		result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_durable_write_display", "host_may_execute_durable_writer", "core_durable_write_not_executed", "host_cli_objective_closeout_writer_display_ready")
		return result.Normalize()
	}
	result.Status = "writer_opt_in_blocked_display"
	result.DisplayState = productionAdapterObjectiveCloseoutWriterFixtureBlockedState(optIn)
	result.ReadyForHostDisplay = true
	result.BlockedDisplay = true
	result.HostMayExecuteDurableWrite = false
	result.NextHostAction = firstNextHostAction(optIn.NextHostAction, "review_objective_closeout_writer_opt_in")
	result.Boundaries = AppendBoundaries(result.Boundaries, "writer_opt_in_blocked_display", "host_cli_objective_closeout_writer_display_ready")
	return result.Normalize()
}

func CloneProductionAdapterObjectiveCloseoutWriterBlackboxFixture(in ProductionAdapterObjectiveCloseoutWriterBlackboxFixture) ProductionAdapterObjectiveCloseoutWriterBlackboxFixture {
	out := in
	out.DisplaySections = cloneStringSlice(in.DisplaySections)
	out.AvailableCapabilityRefs = cloneDisplaySafeRefs(in.AvailableCapabilityRefs)
	out.RequiredCapabilityRefs = cloneDisplaySafeRefs(in.RequiredCapabilityRefs)
	out.PolicyRefs = cloneDisplaySafeRefs(in.PolicyRefs)
	out.RequiredPolicyRefs = cloneDisplaySafeRefs(in.RequiredPolicyRefs)
	out.ApprovalRefs = cloneDisplaySafeRefs(in.ApprovalRefs)
	out.RequiredApprovalRefs = cloneDisplaySafeRefs(in.RequiredApprovalRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (f ProductionAdapterObjectiveCloseoutWriterBlackboxFixture) Clone() ProductionAdapterObjectiveCloseoutWriterBlackboxFixture {
	return CloneProductionAdapterObjectiveCloseoutWriterBlackboxFixture(f)
}

func (f ProductionAdapterObjectiveCloseoutWriterBlackboxFixture) Normalize() ProductionAdapterObjectiveCloseoutWriterBlackboxFixture {
	out := CloneProductionAdapterObjectiveCloseoutWriterBlackboxFixture(f)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = normalizeControlToken(out.Status)
	out.Mode = normalizeControlToken(out.Mode)
	if out.Mode == "" {
		out.Mode = "production_adapter_objective_closeout_writer_blackbox_fixture"
	}
	out.DisplayState = normalizeControlToken(out.DisplayState)
	if out.DisplayState == "" {
		out.DisplayState = "blocked"
	}
	out.DisplaySections = normalizeControlTokenList(out.DisplaySections)
	out.RequestedMode = NormalizeProductionAdapterObjectiveCloseoutWriterMode(string(out.RequestedMode))
	out.FixtureRef = normalizeOneDisplaySafeRef(out.FixtureRef)
	out.WriterOptInRef = normalizeOneDisplaySafeRef(out.WriterOptInRef)
	out.WriterRef = normalizeOneDisplaySafeRef(out.WriterRef)
	out.OwnerRef = normalizeOneDisplaySafeRef(out.OwnerRef)
	out.HostWriterBindingRef = normalizeOneDisplaySafeRef(out.HostWriterBindingRef)
	out.ObjectiveCloseoutHandoffRef = normalizeOneDisplaySafeRef(out.ObjectiveCloseoutHandoffRef)
	out.HostUIHandoffRef = normalizeOneDisplaySafeRef(out.HostUIHandoffRef)
	out.ObjectiveCloseoutPacketRef = normalizeOneDisplaySafeRef(out.ObjectiveCloseoutPacketRef)
	out.ObjectiveRef = normalizeOneDisplaySafeRef(out.ObjectiveRef)
	out.HostObjectiveLifecycleRef = normalizeOneDisplaySafeRef(out.HostObjectiveLifecycleRef)
	out.HostRunstoreRef = normalizeOneDisplaySafeRef(out.HostRunstoreRef)
	out.ExpectedDurableEventRef = normalizeOneDisplaySafeRef(out.ExpectedDurableEventRef)
	out.ExpectedObjectiveStateRef = normalizeOneDisplaySafeRef(out.ExpectedObjectiveStateRef)
	out.HostDurableApplyConfirmationRef = normalizeOneDisplaySafeRef(out.HostDurableApplyConfirmationRef)
	out.AvailableCapabilityRefs = normalizeDisplaySafeRefs(out.AvailableCapabilityRefs)
	out.RequiredCapabilityRefs = normalizeDisplaySafeRefs(out.RequiredCapabilityRefs)
	out.PolicyRefs = normalizeDisplaySafeRefs(out.PolicyRefs)
	out.RequiredPolicyRefs = normalizeDisplaySafeRefs(out.RequiredPolicyRefs)
	out.ApprovalRefs = normalizeDisplaySafeRefs(out.ApprovalRefs)
	out.RequiredApprovalRefs = normalizeDisplaySafeRefs(out.RequiredApprovalRefs)
	out.BudgetRef = normalizeOneDisplaySafeRef(out.BudgetRef)
	out.RequiredBudgetRef = normalizeOneDisplaySafeRef(out.RequiredBudgetRef)
	out.IdempotencyRef = normalizeOneDisplaySafeRef(out.IdempotencyRef)
	out.IdempotencyContractRef = normalizeOneDisplaySafeRef(out.IdempotencyContractRef)
	out.DryRunPlanRef = normalizeOneDisplaySafeRef(out.DryRunPlanRef)
	out.DryRunResultRef = normalizeOneDisplaySafeRef(out.DryRunResultRef)
	out.DryRunContractRef = normalizeOneDisplaySafeRef(out.DryRunContractRef)
	out.ExpectedReadbackRef = normalizeOneDisplaySafeRef(out.ExpectedReadbackRef)
	out.ReadbackContractRef = normalizeOneDisplaySafeRef(out.ReadbackContractRef)
	out.RollbackReviewRef = normalizeOneDisplaySafeRef(out.RollbackReviewRef)
	out.RequiredRollbackReviewRef = normalizeOneDisplaySafeRef(out.RequiredRollbackReviewRef)
	out.CompensationReviewRef = normalizeOneDisplaySafeRef(out.CompensationReviewRef)
	out.RequiredCompensationReviewRef = normalizeOneDisplaySafeRef(out.RequiredCompensationReviewRef)
	out.RedactionPolicyRef = normalizeOneDisplaySafeRef(out.RedactionPolicyRef)
	out.TimeoutPolicyRef = normalizeOneDisplaySafeRef(out.TimeoutPolicyRef)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.BlockedReasons = normalizeControlTokenList(out.BlockedReasons)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	if out.RunnerEffect == "" {
		out.RunnerEffect = "none"
	}
	if out.PromptEffect == "" {
		out.PromptEffect = "none"
	}
	if !out.Available {
		out.Status = "unavailable"
		out.ReadyForHostDisplay = false
		out.ReadyForWriterPlanDisplay = false
		out.ReadyForWriterDryRunDisplay = false
		out.ReadyForDurableWriteDisplay = false
		out.BlockedDisplay = false
		out.HostMayExecuteDurableWrite = false
	}
	if out.Status == "" {
		out.Status = "blocked"
	}
	if out.RawOutputLoaded || productionAdapterObjectiveCloseoutWriterBlackboxFixtureUnsafeOutput(out) {
		out.RawOutputLoaded = true
		out.Status = "blocked"
		out.DisplayState = "blocked_unsafe_refs"
		out.ReadyForHostDisplay = false
		out.ReadyForWriterPlanDisplay = false
		out.ReadyForWriterDryRunDisplay = false
		out.ReadyForDurableWriteDisplay = false
		out.BlockedDisplay = false
		out.HostMayExecuteDurableWrite = false
		out.FailureClass = firstFailureClass(out.FailureClass, FailureEvidenceWeak)
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		if out.NextHostAction == "" {
			out.NextHostAction = "provide_display_safe_refs"
		}
	}
	out.CoreInvocationExecuted = false
	out.DurableWriteByCore = false
	out.ObjectiveStoreWriteByCore = false
	out.RunstoreWriteByCore = false
	out.ReadyForHostDisplay = out.ReadyForHostDisplay &&
		out.Available &&
		out.FixtureRef != "" &&
		out.WriterOptInRef != "" &&
		out.WriterRef != "" &&
		out.ObjectiveCloseoutHandoffRef != "" &&
		out.HostUIHandoffRef != "" &&
		!out.RawOutputLoaded
	out.ReadyForWriterPlanDisplay = out.ReadyForWriterPlanDisplay &&
		out.ReadyForHostDisplay &&
		out.Status == "ready_for_writer_plan_display" &&
		out.DisplayState == "plan_only" &&
		out.PlanOnly &&
		!out.HostMayExecuteDurableWrite
	out.ReadyForWriterDryRunDisplay = out.ReadyForWriterDryRunDisplay &&
		out.ReadyForHostDisplay &&
		out.Status == "ready_for_writer_dry_run_display" &&
		out.DisplayState == "dry_run_ready" &&
		out.DryRun &&
		!out.HostMayExecuteDurableWrite
	out.ReadyForDurableWriteDisplay = out.ReadyForDurableWriteDisplay &&
		out.ReadyForHostDisplay &&
		out.Status == "ready_for_durable_write_display" &&
		out.DisplayState == "durable_write_ready" &&
		out.DurableWriteMode &&
		out.HostMayExecuteDurableWrite
	out.BlockedDisplay = out.BlockedDisplay &&
		out.ReadyForHostDisplay &&
		out.Status == "writer_opt_in_blocked_display" &&
		len(out.BlockedReasons) > 0 &&
		!out.HostMayExecuteDurableWrite
	if !out.ReadyForDurableWriteDisplay {
		out.HostMayExecuteDurableWrite = false
	}
	return out
}

func productionAdapterObjectiveCloseoutWriterBlackboxFixtureBlock(result ProductionAdapterObjectiveCloseoutWriterBlackboxFixture, failure FailureClass, reason string, missing MissingInput, next NextHostAction) ProductionAdapterObjectiveCloseoutWriterBlackboxFixture {
	result.Status = "blocked"
	result.ReadyForHostDisplay = false
	result.ReadyForWriterPlanDisplay = false
	result.ReadyForWriterDryRunDisplay = false
	result.ReadyForDurableWriteDisplay = false
	result.BlockedDisplay = false
	result.HostMayExecuteDurableWrite = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.Boundaries = AppendBoundaries(result.Boundaries, "objective_closeout_writer_blackbox_fixture_blocked")
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	return result
}

func productionAdapterObjectiveCloseoutWriterFixtureDisplayState(optIn ProductionAdapterObjectiveCloseoutWriterOptIn) string {
	switch {
	case optIn.ReadyForHostWriterPlan:
		return "plan_only"
	case optIn.ReadyForHostWriterDryRun:
		return "dry_run_ready"
	case optIn.ReadyForHostDurableWrite:
		return "durable_write_ready"
	default:
		return productionAdapterObjectiveCloseoutWriterFixtureBlockedState(optIn)
	}
}

func productionAdapterObjectiveCloseoutWriterFixtureBlockedState(optIn ProductionAdapterObjectiveCloseoutWriterOptIn) string {
	switch {
	case productionAdapterObjectiveCloseoutWriterHasReason(optIn, "unsafe_input_ref") || containsMissingInput(optIn.MissingInputs, "host:display_safe_refs"):
		return "blocked_unsafe_refs"
	case productionAdapterObjectiveCloseoutWriterHasReason(optIn, "writer_approval_missing") || productionAdapterObjectiveCloseoutWriterHasReason(optIn, "objective_closeout_writer_explicit_opt_in_required"):
		return "blocked_missing_approval"
	case productionAdapterObjectiveCloseoutWriterHasReason(optIn, "writer_dry_run_result_ref_missing"):
		return "blocked_missing_dry_run_result"
	case productionAdapterObjectiveCloseoutWriterHasReason(optIn, "writer_expected_readback_ref_missing") || productionAdapterObjectiveCloseoutWriterHasReason(optIn, "writer_readback_unavailable"):
		return "blocked_missing_expected_readback"
	case productionAdapterObjectiveCloseoutWriterHasReason(optIn, "writer_rollback_review_ref_missing") || productionAdapterObjectiveCloseoutWriterHasReason(optIn, "writer_rollback_review_unavailable"):
		return "blocked_missing_rollback_review"
	case productionAdapterObjectiveCloseoutWriterHasReason(optIn, "writer_compensation_review_ref_missing") || productionAdapterObjectiveCloseoutWriterHasReason(optIn, "writer_compensation_review_unavailable"):
		return "blocked_missing_compensation_review"
	case productionAdapterObjectiveCloseoutWriterHasReason(optIn, "writer_policy_missing"):
		return "blocked_missing_policy"
	case productionAdapterObjectiveCloseoutWriterHasReason(optIn, "writer_capability_missing"):
		return "blocked_missing_capability"
	case productionAdapterObjectiveCloseoutWriterHasReason(optIn, "writer_budget_missing") || productionAdapterObjectiveCloseoutWriterHasReason(optIn, "writer_budget_ref_mismatch"):
		return "blocked_missing_budget"
	case productionAdapterObjectiveCloseoutWriterHasReason(optIn, "writer_idempotency_ref_missing"):
		return "blocked_missing_idempotency"
	default:
		return "blocked"
	}
}

func productionAdapterObjectiveCloseoutWriterHasReason(optIn ProductionAdapterObjectiveCloseoutWriterOptIn, reason string) bool {
	want := normalizeControlToken(reason)
	if want == "" {
		return false
	}
	for _, value := range normalizeControlTokenList(optIn.BlockedReasons) {
		if value == want {
			return true
		}
	}
	return false
}

func productionAdapterObjectiveCloseoutWriterFixtureDisplaySections() []string {
	return []string{
		"objective_closeout_writer_summary",
		"objective_closeout_writer_opt_in",
		"objective_closeout_writer_mode",
		"objective_closeout_writer_dry_run",
		"objective_closeout_writer_readback",
		"objective_closeout_writer_rollback_review",
		"objective_closeout_writer_compensation_review",
		"host_cli_objective_closeout_writer_display",
	}
}

func productionAdapterObjectiveCloseoutWriterBlackboxFixtureBoundaries(groups ...[]Boundary) []Boundary {
	return MergeBoundaries(
		[]Boundary{
			"production_adapter_objective_closeout_writer_blackbox_fixture",
			"objective_closeout_writer_blackbox_fixture_projection_only",
			"host_cli_objective_closeout_writer_display",
			"host_owned_objective_closeout_writer",
			"display_safe_refs_only",
			"display_safe_result_refs_only",
			"no_runner_dispatch",
			"no_durable_write_by_core",
			"no_objective_store_write_by_core",
			"no_runstore_write_by_core",
		},
		MergeBoundaries(groups...),
	)
}

func productionAdapterObjectiveCloseoutWriterBlackboxFixtureUnsafeOutput(input ProductionAdapterObjectiveCloseoutWriterBlackboxFixture) bool {
	return displaySafeRefRejected(input.FixtureRef) ||
		displaySafeRefRejected(input.WriterOptInRef) ||
		displaySafeRefRejected(input.WriterRef) ||
		displaySafeRefRejected(input.OwnerRef) ||
		displaySafeRefRejected(input.HostWriterBindingRef) ||
		displaySafeRefRejected(input.ObjectiveCloseoutHandoffRef) ||
		displaySafeRefRejected(input.HostUIHandoffRef) ||
		displaySafeRefRejected(input.ObjectiveCloseoutPacketRef) ||
		displaySafeRefRejected(input.ObjectiveRef) ||
		displaySafeRefRejected(input.HostObjectiveLifecycleRef) ||
		displaySafeRefRejected(input.HostRunstoreRef) ||
		displaySafeRefRejected(input.ExpectedDurableEventRef) ||
		displaySafeRefRejected(input.ExpectedObjectiveStateRef) ||
		displaySafeRefRejected(input.HostDurableApplyConfirmationRef) ||
		displaySafeRefSliceRejected(input.AvailableCapabilityRefs) ||
		displaySafeRefSliceRejected(input.RequiredCapabilityRefs) ||
		displaySafeRefSliceRejected(input.PolicyRefs) ||
		displaySafeRefSliceRejected(input.RequiredPolicyRefs) ||
		displaySafeRefSliceRejected(input.ApprovalRefs) ||
		displaySafeRefSliceRejected(input.RequiredApprovalRefs) ||
		displaySafeRefRejected(input.BudgetRef) ||
		displaySafeRefRejected(input.RequiredBudgetRef) ||
		displaySafeRefRejected(input.IdempotencyRef) ||
		displaySafeRefRejected(input.IdempotencyContractRef) ||
		displaySafeRefRejected(input.DryRunPlanRef) ||
		displaySafeRefRejected(input.DryRunResultRef) ||
		displaySafeRefRejected(input.DryRunContractRef) ||
		displaySafeRefRejected(input.ExpectedReadbackRef) ||
		displaySafeRefRejected(input.ReadbackContractRef) ||
		displaySafeRefRejected(input.RollbackReviewRef) ||
		displaySafeRefRejected(input.RequiredRollbackReviewRef) ||
		displaySafeRefRejected(input.CompensationReviewRef) ||
		displaySafeRefRejected(input.RequiredCompensationReviewRef) ||
		displaySafeRefRejected(input.RedactionPolicyRef) ||
		displaySafeRefRejected(input.TimeoutPolicyRef) ||
		input.RawOutputLoaded
}

func productionAdapterObjectiveCloseoutWriterOptInEmpty(optIn ProductionAdapterObjectiveCloseoutWriterOptIn) bool {
	return !optIn.Projected &&
		!optIn.Available &&
		optIn.Status == "" &&
		optIn.Mode == "" &&
		optIn.WriterOptInRef == "" &&
		optIn.WriterRef == "" &&
		optIn.ObjectiveCloseoutHandoffRef == "" &&
		optIn.HostUIHandoffRef == "" &&
		len(optIn.MissingInputs) == 0 &&
		len(optIn.BlockedReasons) == 0 &&
		len(optIn.Boundaries) == 0 &&
		optIn.NextHostAction == "" &&
		!optIn.RawOutputLoaded
}

func unavailableProductionAdapterObjectiveCloseoutWriterBlackboxFixture() ProductionAdapterObjectiveCloseoutWriterBlackboxFixture {
	return ProductionAdapterObjectiveCloseoutWriterBlackboxFixture{
		ContractVersion: ContractVersion,
		Projected:       true,
		Available:       false,
		Status:          "unavailable",
		Mode:            "production_adapter_objective_closeout_writer_blackbox_fixture",
		DisplayState:    "blocked",
		DisplaySections: productionAdapterObjectiveCloseoutWriterFixtureDisplaySections(),
		Boundaries: []Boundary{
			"production_adapter_objective_closeout_writer_blackbox_fixture",
			"objective_closeout_writer_blackbox_fixture_projection_only",
			"host_cli_objective_closeout_writer_display",
			"display_safe_refs_only",
			"no_runner_dispatch",
			"no_durable_write_by_core",
			"no_objective_store_write_by_core",
			"no_runstore_write_by_core",
		},
		RunnerEffect:   "none",
		PromptEffect:   "none",
		NextHostAction: "provide_objective_closeout_writer_opt_in",
	}
}
