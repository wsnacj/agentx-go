package controlcontract

type ProductionAdapterObjectiveCloseoutWriterMode string

const (
	ProductionAdapterObjectiveCloseoutWriterPlanOnly     ProductionAdapterObjectiveCloseoutWriterMode = "plan_only"
	ProductionAdapterObjectiveCloseoutWriterDryRun       ProductionAdapterObjectiveCloseoutWriterMode = "dry_run"
	ProductionAdapterObjectiveCloseoutWriterDurableWrite ProductionAdapterObjectiveCloseoutWriterMode = "durable_write"
)

func NormalizeProductionAdapterObjectiveCloseoutWriterMode(raw string) ProductionAdapterObjectiveCloseoutWriterMode {
	switch normalizeEnumToken(raw) {
	case "", "plan", "plan_only", "preview":
		return ProductionAdapterObjectiveCloseoutWriterPlanOnly
	case "dry_run", "dryrun", "shadow", "shadow_write":
		return ProductionAdapterObjectiveCloseoutWriterDryRun
	case "durable_write", "write", "apply", "host_write":
		return ProductionAdapterObjectiveCloseoutWriterDurableWrite
	default:
		return ""
	}
}

type ProductionAdapterObjectiveCloseoutWriterDescriptor struct {
	ContractVersion            string                                         `json:"contract_version,omitempty"`
	Projected                  bool                                           `json:"projected"`
	Status                     HostActionStatus                               `json:"status,omitempty"`
	ReadyForWriterOptIn        bool                                           `json:"ready_for_writer_opt_in"`
	WriterRef                  DisplaySafeRef                                 `json:"writer_ref,omitempty"`
	Owner                      string                                         `json:"owner,omitempty"`
	OwnerRef                   DisplaySafeRef                                 `json:"owner_ref,omitempty"`
	Version                    string                                         `json:"version,omitempty"`
	SupportedModes             []ProductionAdapterObjectiveCloseoutWriterMode `json:"supported_modes,omitempty"`
	SupportedTargetRefs        []DisplaySafeRef                               `json:"supported_target_refs,omitempty"`
	RequiresCapabilityRefs     []DisplaySafeRef                               `json:"requires_capability_refs,omitempty"`
	RequiredPolicyRefs         []DisplaySafeRef                               `json:"required_policy_refs,omitempty"`
	RequiredApprovalRefs       []DisplaySafeRef                               `json:"required_approval_refs,omitempty"`
	RequiredBudgetRef          DisplaySafeRef                                 `json:"required_budget_ref,omitempty"`
	IdempotencyContractRef     DisplaySafeRef                                 `json:"idempotency_contract_ref,omitempty"`
	InputContractRef           DisplaySafeRef                                 `json:"input_contract_ref,omitempty"`
	OutputContractRef          DisplaySafeRef                                 `json:"output_contract_ref,omitempty"`
	DryRunContractRef          DisplaySafeRef                                 `json:"dry_run_contract_ref,omitempty"`
	ReadbackContractRef        DisplaySafeRef                                 `json:"readback_contract_ref,omitempty"`
	RollbackReviewRef          DisplaySafeRef                                 `json:"rollback_review_ref,omitempty"`
	CompensationReviewRef      DisplaySafeRef                                 `json:"compensation_review_ref,omitempty"`
	RedactionPolicyRef         DisplaySafeRef                                 `json:"redaction_policy_ref,omitempty"`
	TimeoutPolicyRef           DisplaySafeRef                                 `json:"timeout_policy_ref,omitempty"`
	PlanOnlyDefault            bool                                           `json:"plan_only_default"`
	DryRunRequired             bool                                           `json:"dry_run_required"`
	ReadbackRequired           bool                                           `json:"readback_required"`
	RollbackReviewRequired     bool                                           `json:"rollback_review_required"`
	CompensationReviewRequired bool                                           `json:"compensation_review_required"`
	MissingInputs              []MissingInput                                 `json:"missing_inputs,omitempty"`
	BlockedReasons             []string                                       `json:"blocked_reasons,omitempty"`
	Boundaries                 []Boundary                                     `json:"boundaries,omitempty"`
	NextHostAction             NextHostAction                                 `json:"next_host_action,omitempty"`
	RunnerEffect               string                                         `json:"runner_effect,omitempty"`
	PromptEffect               string                                         `json:"prompt_effect,omitempty"`
	RawOutputLoaded            bool                                           `json:"raw_output_loaded"`
}

type ProductionAdapterObjectiveCloseoutWriterOptInInput struct {
	WriterOptInRef                  DisplaySafeRef                                     `json:"writer_opt_in_ref,omitempty"`
	WriterDescriptor                ProductionAdapterObjectiveCloseoutWriterDescriptor `json:"writer_descriptor,omitempty"`
	DurableHandoff                  ProductionAdapterObjectiveCloseoutDurableHandoff   `json:"durable_handoff,omitempty"`
	HostUIHandoff                   ProductionAdapterObjectiveCloseoutHostUIHandoff    `json:"host_ui_handoff,omitempty"`
	RequestedMode                   ProductionAdapterObjectiveCloseoutWriterMode       `json:"requested_mode,omitempty"`
	ExplicitOptIn                   bool                                               `json:"explicit_opt_in"`
	HostWriterBindingRef            DisplaySafeRef                                     `json:"host_writer_binding_ref,omitempty"`
	HostWriterAvailable             bool                                               `json:"host_writer_available"`
	HostReadbackAvailable           bool                                               `json:"host_readback_available"`
	HostRollbackReviewAvailable     bool                                               `json:"host_rollback_review_available"`
	HostCompensationReviewAvailable bool                                               `json:"host_compensation_review_available"`
	AvailableCapabilityRefs         []DisplaySafeRef                                   `json:"available_capability_refs,omitempty"`
	PolicyRefs                      []DisplaySafeRef                                   `json:"policy_refs,omitempty"`
	ApprovalRefs                    []DisplaySafeRef                                   `json:"approval_refs,omitempty"`
	BudgetRef                       DisplaySafeRef                                     `json:"budget_ref,omitempty"`
	IdempotencyRef                  DisplaySafeRef                                     `json:"idempotency_ref,omitempty"`
	DryRunPlanRef                   DisplaySafeRef                                     `json:"dry_run_plan_ref,omitempty"`
	DryRunResultRef                 DisplaySafeRef                                     `json:"dry_run_result_ref,omitempty"`
	ExpectedReadbackRef             DisplaySafeRef                                     `json:"expected_readback_ref,omitempty"`
	RollbackReviewRef               DisplaySafeRef                                     `json:"rollback_review_ref,omitempty"`
	CompensationReviewRef           DisplaySafeRef                                     `json:"compensation_review_ref,omitempty"`
	RawOutputLoaded                 bool                                               `json:"raw_output_loaded"`
}

type ProductionAdapterObjectiveCloseoutWriterOptIn struct {
	ContractVersion                 string                                       `json:"contract_version,omitempty"`
	Projected                       bool                                         `json:"projected"`
	Available                       bool                                         `json:"available"`
	Status                          HostActionStatus                             `json:"status,omitempty"`
	Mode                            string                                       `json:"mode,omitempty"`
	RequestedMode                   ProductionAdapterObjectiveCloseoutWriterMode `json:"requested_mode,omitempty"`
	ReadyForHostWriterPlan          bool                                         `json:"ready_for_host_writer_plan"`
	ReadyForHostWriterDryRun        bool                                         `json:"ready_for_host_writer_dry_run"`
	ReadyForHostDurableWrite        bool                                         `json:"ready_for_host_durable_write"`
	HostMayExecuteDurableWrite      bool                                         `json:"host_may_execute_durable_write"`
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

func BuildProductionAdapterObjectiveCloseoutWriterDescriptor(input ProductionAdapterObjectiveCloseoutWriterDescriptor) ProductionAdapterObjectiveCloseoutWriterDescriptor {
	result := input.Normalize()
	result.Status = HostActionBlocked
	result.ReadyForWriterOptIn = false
	if productionAdapterObjectiveCloseoutWriterDescriptorUnsafe(input) {
		result = productionAdapterObjectiveCloseoutWriterDescriptorBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	for _, check := range productionAdapterObjectiveCloseoutWriterDescriptorChecks(result) {
		if check.ref == "" {
			result = productionAdapterObjectiveCloseoutWriterDescriptorBlock(result, check.failure, check.reason, check.missing, check.next)
		}
	}
	if len(result.SupportedModes) == 0 {
		result = productionAdapterObjectiveCloseoutWriterDescriptorBlock(result, FailureConfigMissing, "writer_supported_modes_missing", "host:writer_supported_modes", "provide_objective_closeout_writer_descriptor")
	}
	if !productionAdapterObjectiveCloseoutWriterModeContains(result.SupportedModes, ProductionAdapterObjectiveCloseoutWriterPlanOnly) {
		result = productionAdapterObjectiveCloseoutWriterDescriptorBlock(result, FailureConfigMissing, "writer_plan_only_mode_missing", "host:writer_plan_only_mode", "provide_objective_closeout_writer_descriptor")
	}
	if len(result.BlockedReasons) == 0 && len(result.MissingInputs) == 0 {
		result.Status = HostActionReady
		result.ReadyForWriterOptIn = true
		result.NextHostAction = "review_objective_closeout_writer_opt_in"
		result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_objective_closeout_writer_opt_in")
	}
	return result.Normalize()
}

func BuildProductionAdapterObjectiveCloseoutWriterOptIn(input ProductionAdapterObjectiveCloseoutWriterOptInInput) ProductionAdapterObjectiveCloseoutWriterOptIn {
	descriptor := input.WriterDescriptor.Normalize()
	handoffProvided := !productionAdapterObjectiveCloseoutDurableHandoffEmpty(input.DurableHandoff)
	handoff := input.DurableHandoff.Normalize()
	uiProvided := !productionAdapterObjectiveCloseoutHostUIHandoffEmpty(input.HostUIHandoff)
	ui := input.HostUIHandoff.Normalize()
	mode := firstProductionAdapterObjectiveCloseoutWriterMode(input.RequestedMode, ProductionAdapterObjectiveCloseoutWriterPlanOnly)
	result := ProductionAdapterObjectiveCloseoutWriterOptIn{
		ContractVersion:                 ContractVersion,
		Projected:                       true,
		Available:                       true,
		Status:                          HostActionBlocked,
		Mode:                            "production_adapter_objective_closeout_writer_opt_in",
		RequestedMode:                   mode,
		ExplicitOptIn:                   input.ExplicitOptIn,
		PlanOnly:                        mode == ProductionAdapterObjectiveCloseoutWriterPlanOnly,
		DryRun:                          mode == ProductionAdapterObjectiveCloseoutWriterDryRun,
		DurableWriteMode:                mode == ProductionAdapterObjectiveCloseoutWriterDurableWrite,
		HostWriterAvailable:             input.HostWriterAvailable,
		HostReadbackAvailable:           input.HostReadbackAvailable,
		HostRollbackReviewAvailable:     input.HostRollbackReviewAvailable,
		HostCompensationReviewAvailable: input.HostCompensationReviewAvailable,
		WriterOptInRef:                  normalizeOneDisplaySafeRef(input.WriterOptInRef),
		WriterRef:                       descriptor.WriterRef,
		OwnerRef:                        descriptor.OwnerRef,
		HostWriterBindingRef:            normalizeOneDisplaySafeRef(input.HostWriterBindingRef),
		ObjectiveCloseoutHandoffRef:     handoff.ObjectiveCloseoutHandoffRef,
		HostUIHandoffRef:                ui.HostUIHandoffRef,
		ObjectiveCloseoutPacketRef:      handoff.ObjectiveCloseoutPacketRef,
		ObjectiveRef:                    handoff.ObjectiveRef,
		HostObjectiveLifecycleRef:       handoff.HostObjectiveLifecycleRef,
		HostRunstoreRef:                 handoff.HostRunstoreRef,
		ExpectedDurableEventRef:         handoff.ExpectedDurableEventRef,
		ExpectedObjectiveStateRef:       handoff.ExpectedObjectiveStateRef,
		HostDurableApplyConfirmationRef: handoff.HostDurableApplyConfirmationRef,
		AvailableCapabilityRefs:         normalizeDisplaySafeRefs(input.AvailableCapabilityRefs),
		RequiredCapabilityRefs:          cloneDisplaySafeRefs(descriptor.RequiresCapabilityRefs),
		PolicyRefs:                      normalizeDisplaySafeRefs(input.PolicyRefs),
		RequiredPolicyRefs:              cloneDisplaySafeRefs(descriptor.RequiredPolicyRefs),
		ApprovalRefs:                    normalizeDisplaySafeRefs(input.ApprovalRefs),
		RequiredApprovalRefs:            cloneDisplaySafeRefs(descriptor.RequiredApprovalRefs),
		BudgetRef:                       normalizeOneDisplaySafeRef(input.BudgetRef),
		RequiredBudgetRef:               descriptor.RequiredBudgetRef,
		IdempotencyRef:                  normalizeOneDisplaySafeRef(input.IdempotencyRef),
		IdempotencyContractRef:          descriptor.IdempotencyContractRef,
		DryRunPlanRef:                   normalizeOneDisplaySafeRef(input.DryRunPlanRef),
		DryRunResultRef:                 normalizeOneDisplaySafeRef(input.DryRunResultRef),
		DryRunContractRef:               descriptor.DryRunContractRef,
		ExpectedReadbackRef:             normalizeOneDisplaySafeRef(input.ExpectedReadbackRef),
		ReadbackContractRef:             descriptor.ReadbackContractRef,
		RollbackReviewRef:               normalizeOneDisplaySafeRef(input.RollbackReviewRef),
		RequiredRollbackReviewRef:       descriptor.RollbackReviewRef,
		CompensationReviewRef:           normalizeOneDisplaySafeRef(input.CompensationReviewRef),
		RequiredCompensationReviewRef:   descriptor.CompensationReviewRef,
		RedactionPolicyRef:              descriptor.RedactionPolicyRef,
		TimeoutPolicyRef:                descriptor.TimeoutPolicyRef,
		FailureClass:                    FailureNone,
		Boundaries:                      productionAdapterObjectiveCloseoutWriterOptInBoundaries(descriptor.Boundaries, handoff.Boundaries, ui.Boundaries, mode),
		RunnerEffect:                    "none",
		PromptEffect:                    "none",
		RawOutputLoaded:                 input.RawOutputLoaded || descriptor.RawOutputLoaded || handoff.RawOutputLoaded || ui.RawOutputLoaded,
	}
	if productionAdapterObjectiveCloseoutWriterOptInUnsafe(input, descriptor, handoff, handoffProvided, ui, uiProvided) {
		result = productionAdapterObjectiveCloseoutWriterOptInBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if result.WriterOptInRef == "" {
		result = productionAdapterObjectiveCloseoutWriterOptInBlock(result, FailureEvidenceMissing, "writer_opt_in_ref_missing", "host:objective_closeout_writer_opt_in_ref", "provide_objective_closeout_writer_opt_in")
	}
	if !descriptor.ReadyForWriterOptIn {
		result = productionAdapterObjectiveCloseoutWriterOptInBlock(result, firstFailureClass(descriptorFailureClass(descriptor), FailureConfigMissing), "objective_closeout_writer_descriptor_not_ready", "host:objective_closeout_writer_descriptor", "provide_objective_closeout_writer_descriptor")
	}
	if !productionAdapterObjectiveCloseoutWriterModeContains(descriptor.SupportedModes, mode) {
		result = productionAdapterObjectiveCloseoutWriterOptInBlock(result, FailureUnsupportedOperation, "objective_closeout_writer_mode_unsupported", "host:objective_closeout_writer_mode", "review_objective_closeout_writer_mode")
	}
	if !handoffProvided || !handoff.ReadyForHostDurableApply {
		result = productionAdapterObjectiveCloseoutWriterOptInBlock(result, firstFailureClass(handoff.FailureClass, FailureEvidenceMissing), "objective_closeout_durable_handoff_not_ready", "host:objective_closeout_durable_handoff", "review_objective_closeout_durable_handoff")
	}
	if !uiProvided || !ui.ReadyForHostDurableApply {
		result = productionAdapterObjectiveCloseoutWriterOptInBlock(result, firstFailureClass(ui.FailureClass, FailureEvidenceMissing), "objective_closeout_host_ui_handoff_not_ready", "host:objective_closeout_host_ui_handoff", "review_objective_closeout_host_ui_handoff")
	}
	for _, mismatch := range productionAdapterObjectiveCloseoutWriterOptInMismatches(handoff, ui) {
		result = productionAdapterObjectiveCloseoutWriterOptInBlock(result, FailureVerificationFailed, mismatch.reason, mismatch.missing, "review_objective_closeout_writer_opt_in")
	}
	if result.PlanOnly {
		if len(result.BlockedReasons) == 0 && len(result.MissingInputs) == 0 {
			result.Status = HostActionReviewRequired
			result.ReadyForHostWriterPlan = true
			result.NextHostAction = "review_objective_closeout_writer_plan"
			result.Boundaries = AppendBoundaries(result.Boundaries, "objective_closeout_writer_plan_only", "durable_write_not_enabled")
		}
		return result.Normalize()
	}
	if !input.ExplicitOptIn {
		result = productionAdapterObjectiveCloseoutWriterOptInBlock(result, FailureApprovalRequired, "objective_closeout_writer_explicit_opt_in_required", "host:objective_closeout_writer_explicit_opt_in", "request_objective_closeout_writer_opt_in")
	}
	if result.HostWriterBindingRef == "" {
		result = productionAdapterObjectiveCloseoutWriterOptInBlock(result, FailureConfigMissing, "host_writer_binding_ref_missing", "host:objective_closeout_writer_binding_ref", "provide_objective_closeout_writer_binding")
	}
	if !input.HostWriterAvailable {
		result = productionAdapterObjectiveCloseoutWriterOptInBlock(result, FailureHostAdapterMissing, "host_writer_unavailable", "host:objective_closeout_writer", "provide_objective_closeout_writer")
	}
	for _, required := range descriptor.RequiresCapabilityRefs {
		if !displaySafeRefSliceContains(input.AvailableCapabilityRefs, required) {
			result = productionAdapterObjectiveCloseoutWriterOptInBlock(result, FailureCapabilityMissing, "writer_capability_missing", MissingInput(required), "request_capability_resolution")
		}
	}
	for _, required := range descriptor.RequiredPolicyRefs {
		if !displaySafeRefSliceContains(input.PolicyRefs, required) {
			result = productionAdapterObjectiveCloseoutWriterOptInBlock(result, FailurePolicyBlocked, "writer_policy_missing", MissingInput(required), "provide_objective_closeout_writer_policy")
		}
	}
	for _, required := range descriptor.RequiredApprovalRefs {
		if !displaySafeRefSliceContains(input.ApprovalRefs, required) {
			result = productionAdapterObjectiveCloseoutWriterOptInBlock(result, FailureApprovalRequired, "writer_approval_missing", MissingInput(required), "request_objective_closeout_writer_approval")
		}
	}
	if descriptor.RequiredBudgetRef != "" {
		switch {
		case result.BudgetRef == "":
			result = productionAdapterObjectiveCloseoutWriterOptInBlock(result, FailureBudgetExhausted, "writer_budget_missing", MissingInput(descriptor.RequiredBudgetRef), "provide_objective_closeout_writer_budget")
		case result.BudgetRef != descriptor.RequiredBudgetRef:
			result = productionAdapterObjectiveCloseoutWriterOptInBlock(result, FailurePolicyBlocked, "writer_budget_ref_mismatch", MissingInput(descriptor.RequiredBudgetRef), "review_objective_closeout_writer_budget")
		}
	}
	if result.IdempotencyRef == "" {
		result = productionAdapterObjectiveCloseoutWriterOptInBlock(result, FailureInvalidInput, "writer_idempotency_ref_missing", "host:objective_closeout_writer_idempotency_ref", "provide_objective_closeout_writer_idempotency")
	}
	if result.DryRunPlanRef == "" {
		result = productionAdapterObjectiveCloseoutWriterOptInBlock(result, FailureEvidenceMissing, "writer_dry_run_plan_ref_missing", "host:objective_closeout_writer_dry_run_plan_ref", "provide_objective_closeout_writer_dry_run_plan")
	}
	if result.DryRun {
		if len(result.BlockedReasons) == 0 && len(result.MissingInputs) == 0 {
			result.Status = HostActionReady
			result.ReadyForHostWriterDryRun = true
			result.NextHostAction = "host_may_run_objective_closeout_writer_dry_run"
			result.Boundaries = AppendBoundaries(result.Boundaries, "objective_closeout_writer_dry_run_ready", "durable_write_not_enabled")
		}
		return result.Normalize()
	}
	if descriptor.DryRunRequired && result.DryRunResultRef == "" {
		result = productionAdapterObjectiveCloseoutWriterOptInBlock(result, FailureEvidenceMissing, "writer_dry_run_result_ref_missing", "host:objective_closeout_writer_dry_run_result_ref", "provide_objective_closeout_writer_dry_run_result")
	}
	if descriptor.ReadbackRequired {
		if !input.HostReadbackAvailable {
			result = productionAdapterObjectiveCloseoutWriterOptInBlock(result, FailureConfigMissing, "writer_readback_unavailable", "host:objective_closeout_writer_readback", "provide_objective_closeout_writer_readback")
		}
		if result.ExpectedReadbackRef == "" {
			result = productionAdapterObjectiveCloseoutWriterOptInBlock(result, FailureEvidenceMissing, "writer_expected_readback_ref_missing", "host:objective_closeout_writer_expected_readback_ref", "provide_objective_closeout_writer_expected_readback")
		}
	}
	if descriptor.RollbackReviewRequired {
		if !input.HostRollbackReviewAvailable {
			result = productionAdapterObjectiveCloseoutWriterOptInBlock(result, FailureConfigMissing, "writer_rollback_review_unavailable", "host:objective_closeout_writer_rollback_review", "provide_objective_closeout_writer_rollback_review")
		}
		if result.RollbackReviewRef == "" {
			result = productionAdapterObjectiveCloseoutWriterOptInBlock(result, FailureEvidenceMissing, "writer_rollback_review_ref_missing", "host:objective_closeout_writer_rollback_review_ref", "provide_objective_closeout_writer_rollback_review")
		}
	}
	if descriptor.CompensationReviewRequired {
		if !input.HostCompensationReviewAvailable {
			result = productionAdapterObjectiveCloseoutWriterOptInBlock(result, FailureConfigMissing, "writer_compensation_review_unavailable", "host:objective_closeout_writer_compensation_review", "provide_objective_closeout_writer_compensation_review")
		}
		if result.CompensationReviewRef == "" {
			result = productionAdapterObjectiveCloseoutWriterOptInBlock(result, FailureEvidenceMissing, "writer_compensation_review_ref_missing", "host:objective_closeout_writer_compensation_review_ref", "provide_objective_closeout_writer_compensation_review")
		}
	}
	if len(result.BlockedReasons) == 0 && len(result.MissingInputs) == 0 {
		result.Status = HostActionReady
		result.ReadyForHostDurableWrite = true
		result.HostMayExecuteDurableWrite = true
		result.NextHostAction = "host_may_execute_objective_closeout_durable_writer"
		result.Boundaries = AppendBoundaries(result.Boundaries, "objective_closeout_writer_explicit_opt_in_confirmed", "ready_for_host_objective_closeout_durable_writer", "host_may_execute_durable_writer", "core_durable_write_not_executed")
	}
	return result.Normalize()
}

func CloneProductionAdapterObjectiveCloseoutWriterDescriptor(in ProductionAdapterObjectiveCloseoutWriterDescriptor) ProductionAdapterObjectiveCloseoutWriterDescriptor {
	out := in
	out.SupportedModes = cloneProductionAdapterObjectiveCloseoutWriterModes(in.SupportedModes)
	out.SupportedTargetRefs = cloneDisplaySafeRefs(in.SupportedTargetRefs)
	out.RequiresCapabilityRefs = cloneDisplaySafeRefs(in.RequiresCapabilityRefs)
	out.RequiredPolicyRefs = cloneDisplaySafeRefs(in.RequiredPolicyRefs)
	out.RequiredApprovalRefs = cloneDisplaySafeRefs(in.RequiredApprovalRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (d ProductionAdapterObjectiveCloseoutWriterDescriptor) Clone() ProductionAdapterObjectiveCloseoutWriterDescriptor {
	return CloneProductionAdapterObjectiveCloseoutWriterDescriptor(d)
}

func (d ProductionAdapterObjectiveCloseoutWriterDescriptor) Normalize() ProductionAdapterObjectiveCloseoutWriterDescriptor {
	out := CloneProductionAdapterObjectiveCloseoutWriterDescriptor(d)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.WriterRef = normalizeOneDisplaySafeRef(out.WriterRef)
	out.Owner = normalizeControlToken(out.Owner)
	out.OwnerRef = normalizeOneDisplaySafeRef(out.OwnerRef)
	out.Version = normalizeVersionToken(out.Version)
	out.SupportedModes = normalizeProductionAdapterObjectiveCloseoutWriterModes(out.SupportedModes)
	out.SupportedTargetRefs = normalizeDisplaySafeRefs(out.SupportedTargetRefs)
	out.RequiresCapabilityRefs = normalizeDisplaySafeRefs(out.RequiresCapabilityRefs)
	out.RequiredPolicyRefs = normalizeDisplaySafeRefs(out.RequiredPolicyRefs)
	out.RequiredApprovalRefs = normalizeDisplaySafeRefs(out.RequiredApprovalRefs)
	out.RequiredBudgetRef = normalizeOneDisplaySafeRef(out.RequiredBudgetRef)
	out.IdempotencyContractRef = normalizeOneDisplaySafeRef(out.IdempotencyContractRef)
	out.InputContractRef = normalizeOneDisplaySafeRef(out.InputContractRef)
	out.OutputContractRef = normalizeOneDisplaySafeRef(out.OutputContractRef)
	out.DryRunContractRef = normalizeOneDisplaySafeRef(out.DryRunContractRef)
	out.ReadbackContractRef = normalizeOneDisplaySafeRef(out.ReadbackContractRef)
	out.RollbackReviewRef = normalizeOneDisplaySafeRef(out.RollbackReviewRef)
	out.CompensationReviewRef = normalizeOneDisplaySafeRef(out.CompensationReviewRef)
	out.RedactionPolicyRef = normalizeOneDisplaySafeRef(out.RedactionPolicyRef)
	out.TimeoutPolicyRef = normalizeOneDisplaySafeRef(out.TimeoutPolicyRef)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.BlockedReasons = normalizeControlTokenList(out.BlockedReasons)
	out.Boundaries = productionAdapterObjectiveCloseoutWriterDescriptorBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	if out.RunnerEffect == "" {
		out.RunnerEffect = "none"
	}
	if out.PromptEffect == "" {
		out.PromptEffect = "none"
	}
	if out.RawOutputLoaded || productionAdapterObjectiveCloseoutWriterDescriptorUnsafe(out) {
		out.RawOutputLoaded = true
		out.Status = HostActionBlocked
		out.ReadyForWriterOptIn = false
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		if out.NextHostAction == "" {
			out.NextHostAction = "provide_display_safe_refs"
		}
	}
	out.ReadyForWriterOptIn = out.ReadyForWriterOptIn &&
		out.Status == HostActionReady &&
		out.WriterRef != "" &&
		out.OwnerRef != "" &&
		len(out.SupportedModes) > 0 &&
		out.InputContractRef != "" &&
		out.OutputContractRef != "" &&
		out.DryRunContractRef != "" &&
		out.ReadbackContractRef != "" &&
		out.IdempotencyContractRef != "" &&
		out.RedactionPolicyRef != "" &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.RawOutputLoaded
	return out
}

func CloneProductionAdapterObjectiveCloseoutWriterOptIn(in ProductionAdapterObjectiveCloseoutWriterOptIn) ProductionAdapterObjectiveCloseoutWriterOptIn {
	out := in
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

func (o ProductionAdapterObjectiveCloseoutWriterOptIn) Clone() ProductionAdapterObjectiveCloseoutWriterOptIn {
	return CloneProductionAdapterObjectiveCloseoutWriterOptIn(o)
}

func (o ProductionAdapterObjectiveCloseoutWriterOptIn) Normalize() ProductionAdapterObjectiveCloseoutWriterOptIn {
	out := CloneProductionAdapterObjectiveCloseoutWriterOptIn(o)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.Mode = normalizeControlToken(out.Mode)
	if out.Mode == "" {
		out.Mode = "production_adapter_objective_closeout_writer_opt_in"
	}
	out.RequestedMode = NormalizeProductionAdapterObjectiveCloseoutWriterMode(string(out.RequestedMode))
	if out.RequestedMode == "" {
		out.RequestedMode = ProductionAdapterObjectiveCloseoutWriterPlanOnly
	}
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
	if out.RawOutputLoaded || productionAdapterObjectiveCloseoutWriterOptInOutputUnsafe(out) {
		out.RawOutputLoaded = true
		out.Status = HostActionBlocked
		out.FailureClass = firstFailureClass(out.FailureClass, FailureEvidenceWeak)
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		if out.NextHostAction == "" {
			out.NextHostAction = "provide_display_safe_refs"
		}
	}
	out.PlanOnly = out.RequestedMode == ProductionAdapterObjectiveCloseoutWriterPlanOnly
	out.DryRun = out.RequestedMode == ProductionAdapterObjectiveCloseoutWriterDryRun
	out.DurableWriteMode = out.RequestedMode == ProductionAdapterObjectiveCloseoutWriterDurableWrite
	out.CoreInvocationExecuted = false
	out.DurableWriteByCore = false
	out.ObjectiveStoreWriteByCore = false
	out.RunstoreWriteByCore = false
	out.ReadyForHostWriterPlan = out.ReadyForHostWriterPlan &&
		out.Status == HostActionReviewRequired &&
		out.PlanOnly &&
		out.WriterOptInRef != "" &&
		out.WriterRef != "" &&
		out.ObjectiveCloseoutHandoffRef != "" &&
		out.HostUIHandoffRef != "" &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.RawOutputLoaded
	out.ReadyForHostWriterDryRun = out.ReadyForHostWriterDryRun &&
		out.Status == HostActionReady &&
		out.DryRun &&
		out.ExplicitOptIn &&
		out.HostWriterBindingRef != "" &&
		out.DryRunPlanRef != "" &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.RawOutputLoaded
	out.ReadyForHostDurableWrite = out.ReadyForHostDurableWrite &&
		out.Status == HostActionReady &&
		out.DurableWriteMode &&
		out.ExplicitOptIn &&
		out.HostWriterBindingRef != "" &&
		out.DryRunPlanRef != "" &&
		out.ExpectedReadbackRef != "" &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.RawOutputLoaded
	out.HostMayExecuteDurableWrite = out.HostMayExecuteDurableWrite &&
		out.ReadyForHostDurableWrite &&
		!out.CoreInvocationExecuted &&
		!out.DurableWriteByCore &&
		!out.ObjectiveStoreWriteByCore &&
		!out.RunstoreWriteByCore
	if !out.ReadyForHostDurableWrite {
		out.HostMayExecuteDurableWrite = false
	}
	if out.Status == HostActionReady && !out.ReadyForHostWriterDryRun && !out.ReadyForHostDurableWrite {
		out.Status = HostActionReviewRequired
	}
	return out
}

func productionAdapterObjectiveCloseoutWriterDescriptorBlock(result ProductionAdapterObjectiveCloseoutWriterDescriptor, failure FailureClass, reason string, missing MissingInput, next NextHostAction) ProductionAdapterObjectiveCloseoutWriterDescriptor {
	result.Status = HostActionBlocked
	result.ReadyForWriterOptIn = false
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.Boundaries = AppendBoundaries(result.Boundaries, "objective_closeout_writer_descriptor_blocked")
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	return result
}

func productionAdapterObjectiveCloseoutWriterOptInBlock(result ProductionAdapterObjectiveCloseoutWriterOptIn, failure FailureClass, reason string, missing MissingInput, next NextHostAction) ProductionAdapterObjectiveCloseoutWriterOptIn {
	result.Status = HostActionBlocked
	result.ReadyForHostWriterPlan = false
	result.ReadyForHostWriterDryRun = false
	result.ReadyForHostDurableWrite = false
	result.HostMayExecuteDurableWrite = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.Boundaries = AppendBoundaries(result.Boundaries, "objective_closeout_writer_opt_in_blocked")
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	return result
}

type productionAdapterObjectiveCloseoutWriterDescriptorCheck struct {
	ref     DisplaySafeRef
	failure FailureClass
	reason  string
	missing MissingInput
	next    NextHostAction
}

func productionAdapterObjectiveCloseoutWriterDescriptorChecks(input ProductionAdapterObjectiveCloseoutWriterDescriptor) []productionAdapterObjectiveCloseoutWriterDescriptorCheck {
	return []productionAdapterObjectiveCloseoutWriterDescriptorCheck{
		{input.WriterRef, FailureConfigMissing, "writer_ref_missing", "host:objective_closeout_writer_ref", "provide_objective_closeout_writer_descriptor"},
		{input.OwnerRef, FailureConfigMissing, "writer_owner_ref_missing", "host:objective_closeout_writer_owner_ref", "provide_objective_closeout_writer_descriptor"},
		{input.InputContractRef, FailureConfigMissing, "writer_input_contract_ref_missing", "host:objective_closeout_writer_input_contract_ref", "provide_objective_closeout_writer_descriptor"},
		{input.OutputContractRef, FailureConfigMissing, "writer_output_contract_ref_missing", "host:objective_closeout_writer_output_contract_ref", "provide_objective_closeout_writer_descriptor"},
		{input.DryRunContractRef, FailureConfigMissing, "writer_dry_run_contract_ref_missing", "host:objective_closeout_writer_dry_run_contract_ref", "provide_objective_closeout_writer_descriptor"},
		{input.ReadbackContractRef, FailureConfigMissing, "writer_readback_contract_ref_missing", "host:objective_closeout_writer_readback_contract_ref", "provide_objective_closeout_writer_descriptor"},
		{input.IdempotencyContractRef, FailureConfigMissing, "writer_idempotency_contract_ref_missing", "host:objective_closeout_writer_idempotency_contract_ref", "provide_objective_closeout_writer_descriptor"},
		{input.RedactionPolicyRef, FailureConfigMissing, "writer_redaction_policy_ref_missing", "host:objective_closeout_writer_redaction_policy_ref", "provide_objective_closeout_writer_descriptor"},
		{input.TimeoutPolicyRef, FailureConfigMissing, "writer_timeout_policy_ref_missing", "host:objective_closeout_writer_timeout_policy_ref", "provide_objective_closeout_writer_descriptor"},
	}
}

type productionAdapterObjectiveCloseoutWriterOptInMismatch struct {
	reason  string
	missing MissingInput
}

func productionAdapterObjectiveCloseoutWriterOptInMismatches(handoff ProductionAdapterObjectiveCloseoutDurableHandoff, ui ProductionAdapterObjectiveCloseoutHostUIHandoff) []productionAdapterObjectiveCloseoutWriterOptInMismatch {
	var out []productionAdapterObjectiveCloseoutWriterOptInMismatch
	out = append(out, productionAdapterObjectiveCloseoutWriterOptInRefMismatch(handoff.ObjectiveCloseoutHandoffRef, ui.ObjectiveCloseoutHandoffRef, "writer_opt_in_handoff_ref_mismatch", "host:objective_closeout_handoff_ref")...)
	out = append(out, productionAdapterObjectiveCloseoutWriterOptInRefMismatch(handoff.ObjectiveCloseoutPacketRef, ui.ObjectiveCloseoutPacketRef, "writer_opt_in_packet_ref_mismatch", "host:objective_closeout_packet_ref")...)
	out = append(out, productionAdapterObjectiveCloseoutWriterOptInRefMismatch(handoff.ObjectiveRef, ui.ObjectiveRef, "writer_opt_in_objective_ref_mismatch", "host:objective_ref")...)
	out = append(out, productionAdapterObjectiveCloseoutWriterOptInRefMismatch(handoff.HostRunstoreRef, ui.HostRunstoreRef, "writer_opt_in_runstore_ref_mismatch", "host:runstore_ref")...)
	return out
}

func productionAdapterObjectiveCloseoutWriterOptInRefMismatch(left DisplaySafeRef, right DisplaySafeRef, reason string, missing MissingInput) []productionAdapterObjectiveCloseoutWriterOptInMismatch {
	left = normalizeOneDisplaySafeRef(left)
	right = normalizeOneDisplaySafeRef(right)
	if left != "" && right != "" && left != right {
		return []productionAdapterObjectiveCloseoutWriterOptInMismatch{{reason: reason, missing: missing}}
	}
	return nil
}

func productionAdapterObjectiveCloseoutWriterDescriptorBoundaries(values []Boundary) []Boundary {
	return MergeBoundaries([]Boundary{
		"production_adapter_objective_closeout_writer_descriptor",
		"objective_closeout_writer_descriptor_projection_only",
		"host_owned_objective_closeout_writer",
		"explicit_opt_in_required",
		"display_safe_refs_only",
		"display_safe_result_refs_only",
		"no_runner_dispatch",
		"no_durable_write_by_core",
		"no_objective_store_write_by_core",
		"no_runstore_write_by_core",
	}, values)
}

func productionAdapterObjectiveCloseoutWriterOptInBoundaries(descriptorBoundaries []Boundary, handoffBoundaries []Boundary, uiBoundaries []Boundary, mode ProductionAdapterObjectiveCloseoutWriterMode) []Boundary {
	out := MergeBoundaries(descriptorBoundaries, handoffBoundaries, uiBoundaries)
	out = AppendBoundaries(out,
		"production_adapter_objective_closeout_writer_opt_in",
		"objective_closeout_writer_opt_in_projection_only",
		"host_owned_objective_closeout_writer",
		"explicit_opt_in_required",
		"display_safe_refs_only",
		"display_safe_result_refs_only",
		"no_runner_dispatch",
		"no_durable_write_by_core",
		"no_objective_store_write_by_core",
		"no_runstore_write_by_core",
	)
	switch mode {
	case ProductionAdapterObjectiveCloseoutWriterDryRun:
		out = AppendBoundaries(out, "objective_closeout_writer_dry_run_mode")
	case ProductionAdapterObjectiveCloseoutWriterDurableWrite:
		out = AppendBoundaries(out, "objective_closeout_writer_durable_write_mode")
	default:
		out = AppendBoundaries(out, "objective_closeout_writer_plan_only_mode")
	}
	return normalizeBoundaries(out)
}

func productionAdapterObjectiveCloseoutWriterDescriptorUnsafe(input ProductionAdapterObjectiveCloseoutWriterDescriptor) bool {
	return input.RawOutputLoaded ||
		displaySafeRefRejected(input.WriterRef) ||
		displaySafeRefRejected(input.OwnerRef) ||
		displaySafeRefSliceRejected(input.SupportedTargetRefs) ||
		displaySafeRefSliceRejected(input.RequiresCapabilityRefs) ||
		displaySafeRefSliceRejected(input.RequiredPolicyRefs) ||
		displaySafeRefSliceRejected(input.RequiredApprovalRefs) ||
		displaySafeRefRejected(input.RequiredBudgetRef) ||
		displaySafeRefRejected(input.IdempotencyContractRef) ||
		displaySafeRefRejected(input.InputContractRef) ||
		displaySafeRefRejected(input.OutputContractRef) ||
		displaySafeRefRejected(input.DryRunContractRef) ||
		displaySafeRefRejected(input.ReadbackContractRef) ||
		displaySafeRefRejected(input.RollbackReviewRef) ||
		displaySafeRefRejected(input.CompensationReviewRef) ||
		displaySafeRefRejected(input.RedactionPolicyRef) ||
		displaySafeRefRejected(input.TimeoutPolicyRef)
}

func productionAdapterObjectiveCloseoutWriterOptInUnsafe(input ProductionAdapterObjectiveCloseoutWriterOptInInput, descriptor ProductionAdapterObjectiveCloseoutWriterDescriptor, handoff ProductionAdapterObjectiveCloseoutDurableHandoff, handoffProvided bool, ui ProductionAdapterObjectiveCloseoutHostUIHandoff, uiProvided bool) bool {
	return input.RawOutputLoaded ||
		displaySafeRefRejected(input.WriterOptInRef) ||
		displaySafeRefRejected(input.HostWriterBindingRef) ||
		displaySafeRefSliceRejected(input.AvailableCapabilityRefs) ||
		displaySafeRefSliceRejected(input.PolicyRefs) ||
		displaySafeRefSliceRejected(input.ApprovalRefs) ||
		displaySafeRefRejected(input.BudgetRef) ||
		displaySafeRefRejected(input.IdempotencyRef) ||
		displaySafeRefRejected(input.DryRunPlanRef) ||
		displaySafeRefRejected(input.DryRunResultRef) ||
		displaySafeRefRejected(input.ExpectedReadbackRef) ||
		displaySafeRefRejected(input.RollbackReviewRef) ||
		displaySafeRefRejected(input.CompensationReviewRef) ||
		productionAdapterObjectiveCloseoutWriterDescriptorUnsafe(descriptor) ||
		(handoffProvided && productionAdapterObjectiveCloseoutDurableHandoffRefsUnsafe(handoff)) ||
		(uiProvided && productionAdapterObjectiveCloseoutHostUIHandoffUnsafeOutput(ui))
}

func productionAdapterObjectiveCloseoutWriterOptInOutputUnsafe(input ProductionAdapterObjectiveCloseoutWriterOptIn) bool {
	return displaySafeRefRejected(input.WriterOptInRef) ||
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

func productionAdapterObjectiveCloseoutHostUIHandoffEmpty(handoff ProductionAdapterObjectiveCloseoutHostUIHandoff) bool {
	return !handoff.Projected &&
		!handoff.Available &&
		handoff.Status == "" &&
		handoff.Mode == "" &&
		handoff.HostUIHandoffRef == "" &&
		handoff.HostViewRef == "" &&
		handoff.ObjectiveCloseoutPacketRef == "" &&
		handoff.ObjectiveRef == "" &&
		len(handoff.MissingInputs) == 0 &&
		len(handoff.BlockedReasons) == 0 &&
		len(handoff.Boundaries) == 0 &&
		handoff.NextHostAction == "" &&
		!handoff.RawOutputLoaded
}

func productionAdapterObjectiveCloseoutWriterDescriptorEmpty(descriptor ProductionAdapterObjectiveCloseoutWriterDescriptor) bool {
	return !descriptor.Projected &&
		descriptor.Status == "" &&
		descriptor.WriterRef == "" &&
		descriptor.OwnerRef == "" &&
		len(descriptor.MissingInputs) == 0 &&
		len(descriptor.BlockedReasons) == 0 &&
		len(descriptor.Boundaries) == 0 &&
		descriptor.NextHostAction == "" &&
		!descriptor.RawOutputLoaded
}

func descriptorFailureClass(descriptor ProductionAdapterObjectiveCloseoutWriterDescriptor) FailureClass {
	if descriptor.RawOutputLoaded {
		return FailureEvidenceWeak
	}
	if len(descriptor.MissingInputs) > 0 || len(descriptor.BlockedReasons) > 0 {
		return FailureConfigMissing
	}
	return FailureNone
}

func cloneProductionAdapterObjectiveCloseoutWriterModes(values []ProductionAdapterObjectiveCloseoutWriterMode) []ProductionAdapterObjectiveCloseoutWriterMode {
	if len(values) == 0 {
		return nil
	}
	return append([]ProductionAdapterObjectiveCloseoutWriterMode(nil), values...)
}

func normalizeProductionAdapterObjectiveCloseoutWriterModes(values []ProductionAdapterObjectiveCloseoutWriterMode) []ProductionAdapterObjectiveCloseoutWriterMode {
	out := make([]ProductionAdapterObjectiveCloseoutWriterMode, 0, len(values))
	seen := map[ProductionAdapterObjectiveCloseoutWriterMode]bool{}
	for _, value := range values {
		mode := NormalizeProductionAdapterObjectiveCloseoutWriterMode(string(value))
		if mode == "" || seen[mode] {
			continue
		}
		seen[mode] = true
		out = append(out, mode)
	}
	return out
}

func productionAdapterObjectiveCloseoutWriterModeContains(values []ProductionAdapterObjectiveCloseoutWriterMode, want ProductionAdapterObjectiveCloseoutWriterMode) bool {
	want = NormalizeProductionAdapterObjectiveCloseoutWriterMode(string(want))
	if want == "" {
		return false
	}
	for _, value := range normalizeProductionAdapterObjectiveCloseoutWriterModes(values) {
		if value == want {
			return true
		}
	}
	return false
}

func firstProductionAdapterObjectiveCloseoutWriterMode(values ...ProductionAdapterObjectiveCloseoutWriterMode) ProductionAdapterObjectiveCloseoutWriterMode {
	for _, value := range values {
		if mode := NormalizeProductionAdapterObjectiveCloseoutWriterMode(string(value)); mode != "" {
			return mode
		}
	}
	return ""
}
