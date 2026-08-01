package controlcontract

type ProductionAdapterKind string

const (
	ProductionAdapterSourceApply             ProductionAdapterKind = "source_apply"
	ProductionAdapterSourceReadback          ProductionAdapterKind = "source_readback"
	ProductionAdapterCapabilityApply         ProductionAdapterKind = "capability_apply"
	ProductionAdapterOperationsSchedule      ProductionAdapterKind = "operations_schedule"
	ProductionAdapterOperationsMetricCollect ProductionAdapterKind = "operations_metric_collect"
	ProductionAdapterWorkflowDispatch        ProductionAdapterKind = "workflow_dispatch"
)

func NormalizeProductionAdapterKind(raw string) ProductionAdapterKind {
	switch normalizeEnumToken(raw) {
	case "source_apply", "apply_source":
		return ProductionAdapterSourceApply
	case "source_readback", "readback_source":
		return ProductionAdapterSourceReadback
	case "capability_apply", "capability_install", "capability_enable", "capability_authorize":
		return ProductionAdapterCapabilityApply
	case "operations_schedule", "operation_schedule", "schedule":
		return ProductionAdapterOperationsSchedule
	case "operations_metric_collect", "operation_metric_collect", "metric_collect", "metrics_collect":
		return ProductionAdapterOperationsMetricCollect
	case "workflow_dispatch", "dispatch_workflow":
		return ProductionAdapterWorkflowDispatch
	default:
		return ""
	}
}

type ProductionAdapterDescriptor struct {
	ContractVersion        string                `json:"contract_version,omitempty"`
	Projected              bool                  `json:"projected"`
	AdapterRef             DisplaySafeRef        `json:"adapter_ref,omitempty"`
	Owner                  string                `json:"owner,omitempty"`
	OwnerRef               DisplaySafeRef        `json:"owner_ref,omitempty"`
	Version                string                `json:"version,omitempty"`
	Kind                   ProductionAdapterKind `json:"kind,omitempty"`
	SupportedSourceKinds   []ReplannerSourceKind `json:"supported_source_kinds,omitempty"`
	SupportedCandidateRefs []DisplaySafeRef      `json:"supported_candidate_refs,omitempty"`
	ProvidesCapabilityRefs []DisplaySafeRef      `json:"provides_capability_refs,omitempty"`
	RequiresCapabilityRefs []DisplaySafeRef      `json:"requires_capability_refs,omitempty"`
	InputContractRef       DisplaySafeRef        `json:"input_contract_ref,omitempty"`
	OutputContractRef      DisplaySafeRef        `json:"output_contract_ref,omitempty"`
	ReadbackContractRef    DisplaySafeRef        `json:"readback_contract_ref,omitempty"`
	RequiredPolicyRefs     []DisplaySafeRef      `json:"required_policy_refs,omitempty"`
	RequiredApprovalRefs   []DisplaySafeRef      `json:"required_approval_refs,omitempty"`
	RequiredBudgetRef      DisplaySafeRef        `json:"required_budget_ref,omitempty"`
	IdempotencyContractRef DisplaySafeRef        `json:"idempotency_contract_ref,omitempty"`
	RiskRef                DisplaySafeRef        `json:"risk_ref,omitempty"`
	SideEffectClass        string                `json:"side_effect_class,omitempty"`
	TimeoutPolicyRef       DisplaySafeRef        `json:"timeout_policy_ref,omitempty"`
	CompensationHandoffRef DisplaySafeRef        `json:"compensation_handoff_ref,omitempty"`
	RedactionPolicyRef     DisplaySafeRef        `json:"redaction_policy_ref,omitempty"`
	PreflightCheckRefs     []DisplaySafeRef      `json:"preflight_check_refs,omitempty"`
	DisplaySafeInputRefs   []DisplaySafeRef      `json:"display_safe_input_refs,omitempty"`
	DisplaySafeOutputRefs  []DisplaySafeRef      `json:"display_safe_output_refs,omitempty"`
	MissingInputs          []MissingInput        `json:"missing_inputs,omitempty"`
	Boundaries             []Boundary            `json:"boundaries,omitempty"`
	RunnerEffect           string                `json:"runner_effect,omitempty"`
	PromptEffect           string                `json:"prompt_effect,omitempty"`
	RawOutputLoaded        bool                  `json:"raw_output_loaded"`
}

func CloneProductionAdapterDescriptor(in ProductionAdapterDescriptor) ProductionAdapterDescriptor {
	out := in
	out.SupportedSourceKinds = cloneReplannerSourceKinds(in.SupportedSourceKinds)
	out.SupportedCandidateRefs = cloneDisplaySafeRefs(in.SupportedCandidateRefs)
	out.ProvidesCapabilityRefs = cloneDisplaySafeRefs(in.ProvidesCapabilityRefs)
	out.RequiresCapabilityRefs = cloneDisplaySafeRefs(in.RequiresCapabilityRefs)
	out.RequiredPolicyRefs = cloneDisplaySafeRefs(in.RequiredPolicyRefs)
	out.RequiredApprovalRefs = cloneDisplaySafeRefs(in.RequiredApprovalRefs)
	out.PreflightCheckRefs = cloneDisplaySafeRefs(in.PreflightCheckRefs)
	out.DisplaySafeInputRefs = cloneDisplaySafeRefs(in.DisplaySafeInputRefs)
	out.DisplaySafeOutputRefs = cloneDisplaySafeRefs(in.DisplaySafeOutputRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (d ProductionAdapterDescriptor) Clone() ProductionAdapterDescriptor {
	return CloneProductionAdapterDescriptor(d)
}

func (d ProductionAdapterDescriptor) Normalize() ProductionAdapterDescriptor {
	out := CloneProductionAdapterDescriptor(d)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.AdapterRef, _ = NormalizeDisplaySafeRef(string(out.AdapterRef))
	out.Owner = normalizeControlToken(out.Owner)
	out.OwnerRef, _ = NormalizeDisplaySafeRef(string(out.OwnerRef))
	out.Version = normalizeVersionToken(out.Version)
	out.Kind = NormalizeProductionAdapterKind(string(out.Kind))
	out.SupportedSourceKinds = normalizeReplannerSourceKinds(out.SupportedSourceKinds)
	out.SupportedCandidateRefs = normalizeDisplaySafeRefs(out.SupportedCandidateRefs)
	out.ProvidesCapabilityRefs = normalizeDisplaySafeRefs(out.ProvidesCapabilityRefs)
	out.RequiresCapabilityRefs = normalizeDisplaySafeRefs(out.RequiresCapabilityRefs)
	out.InputContractRef, _ = NormalizeDisplaySafeRef(string(out.InputContractRef))
	out.OutputContractRef, _ = NormalizeDisplaySafeRef(string(out.OutputContractRef))
	out.ReadbackContractRef, _ = NormalizeDisplaySafeRef(string(out.ReadbackContractRef))
	out.RequiredPolicyRefs = normalizeDisplaySafeRefs(out.RequiredPolicyRefs)
	out.RequiredApprovalRefs = normalizeDisplaySafeRefs(out.RequiredApprovalRefs)
	out.RequiredBudgetRef, _ = NormalizeDisplaySafeRef(string(out.RequiredBudgetRef))
	out.IdempotencyContractRef, _ = NormalizeDisplaySafeRef(string(out.IdempotencyContractRef))
	out.RiskRef, _ = NormalizeDisplaySafeRef(string(out.RiskRef))
	out.SideEffectClass = normalizeControlToken(out.SideEffectClass)
	out.TimeoutPolicyRef, _ = NormalizeDisplaySafeRef(string(out.TimeoutPolicyRef))
	out.CompensationHandoffRef, _ = NormalizeDisplaySafeRef(string(out.CompensationHandoffRef))
	out.RedactionPolicyRef, _ = NormalizeDisplaySafeRef(string(out.RedactionPolicyRef))
	out.PreflightCheckRefs = normalizeDisplaySafeRefs(out.PreflightCheckRefs)
	out.DisplaySafeInputRefs = normalizeDisplaySafeRefs(out.DisplaySafeInputRefs)
	out.DisplaySafeOutputRefs = normalizeDisplaySafeRefs(out.DisplaySafeOutputRefs)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.Boundaries = MergeBoundaries([]Boundary{
		"production_adapter_descriptor",
		"descriptor_projection_only",
		"host_owned_adapter",
		"display_safe_refs_only",
		"no_adapter_invocation",
		"no_runner_dispatch",
	}, out.Boundaries)
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	if out.RunnerEffect == "" {
		out.RunnerEffect = "none"
	}
	if out.PromptEffect == "" {
		out.PromptEffect = "none"
	}
	if out.RawOutputLoaded {
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
	}
	return out
}

type ProductionAdapterResolutionInput struct {
	ApplyEnvelopeReady       bool                              `json:"apply_envelope_ready"`
	ApplyEnvelopeRef         DisplaySafeRef                    `json:"apply_envelope_ref,omitempty"`
	SelectedSourceKind       ReplannerSourceKind               `json:"selected_source_kind,omitempty"`
	SelectedSourceRef        DisplaySafeRef                    `json:"selected_source_ref,omitempty"`
	SelectedCandidateRef     DisplaySafeRef                    `json:"selected_candidate_ref,omitempty"`
	RequestedAdapterRef      DisplaySafeRef                    `json:"requested_adapter_ref,omitempty"`
	CatalogSnapshotRef       DisplaySafeRef                    `json:"catalog_snapshot_ref,omitempty"`
	CatalogSelectionRef      DisplaySafeRef                    `json:"catalog_selection_ref,omitempty"`
	CatalogSelection         ProductionAdapterCatalogSelection `json:"catalog_selection,omitempty"`
	HostPolicyRef            DisplaySafeRef                    `json:"host_policy_ref,omitempty"`
	ApprovalContextRefs      []DisplaySafeRef                  `json:"approval_context_refs,omitempty"`
	BudgetRef                DisplaySafeRef                    `json:"budget_ref,omitempty"`
	IdempotencyRef           DisplaySafeRef                    `json:"idempotency_ref,omitempty"`
	AvailableCapabilityRefs  []DisplaySafeRef                  `json:"available_capability_refs,omitempty"`
	ConfirmedPolicyRefs      []DisplaySafeRef                  `json:"confirmed_policy_refs,omitempty"`
	ConfirmedApprovalRefs    []DisplaySafeRef                  `json:"confirmed_approval_refs,omitempty"`
	Descriptor               ProductionAdapterDescriptor       `json:"descriptor,omitempty"`
	AllowAdapterSubstitution bool                              `json:"allow_adapter_substitution"`
	RawOutputLoaded          bool                              `json:"raw_output_loaded"`
}

type ProductionAdapterResolution struct {
	ContractVersion        string                      `json:"contract_version,omitempty"`
	Projected              bool                        `json:"projected"`
	Status                 HostActionStatus            `json:"status,omitempty"`
	ReadyForHostPreflight  bool                        `json:"ready_for_host_preflight"`
	AdapterRef             DisplaySafeRef              `json:"adapter_ref,omitempty"`
	DescriptorRef          DisplaySafeRef              `json:"descriptor_ref,omitempty"`
	Descriptor             ProductionAdapterDescriptor `json:"descriptor,omitempty"`
	ApplyEnvelopeRef       DisplaySafeRef              `json:"apply_envelope_ref,omitempty"`
	SelectedSourceKind     ReplannerSourceKind         `json:"selected_source_kind,omitempty"`
	SelectedSourceRef      DisplaySafeRef              `json:"selected_source_ref,omitempty"`
	SelectedCandidateRef   DisplaySafeRef              `json:"selected_candidate_ref,omitempty"`
	RequestedAdapterRef    DisplaySafeRef              `json:"requested_adapter_ref,omitempty"`
	CatalogSnapshotRef     DisplaySafeRef              `json:"catalog_snapshot_ref,omitempty"`
	CatalogSelectionRef    DisplaySafeRef              `json:"catalog_selection_ref,omitempty"`
	HostPolicyRef          DisplaySafeRef              `json:"host_policy_ref,omitempty"`
	ApprovalContextRefs    []DisplaySafeRef            `json:"approval_context_refs,omitempty"`
	BudgetRef              DisplaySafeRef              `json:"budget_ref,omitempty"`
	IdempotencyRef         DisplaySafeRef              `json:"idempotency_ref,omitempty"`
	MissingInputs          []MissingInput              `json:"missing_inputs,omitempty"`
	BlockedReasons         []string                    `json:"blocked_reasons,omitempty"`
	RequiredCapabilityRefs []DisplaySafeRef            `json:"required_capability_refs,omitempty"`
	RequiredPolicyRefs     []DisplaySafeRef            `json:"required_policy_refs,omitempty"`
	RequiredApprovalRefs   []DisplaySafeRef            `json:"required_approval_refs,omitempty"`
	RequiredPreflightRefs  []DisplaySafeRef            `json:"required_preflight_refs,omitempty"`
	FailureClass           FailureClass                `json:"failure_class,omitempty"`
	Boundaries             []Boundary                  `json:"boundaries,omitempty"`
	NextHostAction         NextHostAction              `json:"next_host_action,omitempty"`
	RunnerEffect           string                      `json:"runner_effect,omitempty"`
	PromptEffect           string                      `json:"prompt_effect,omitempty"`
	RawOutputLoaded        bool                        `json:"raw_output_loaded"`
}

func BuildProductionAdapterResolution(input ProductionAdapterResolutionInput) ProductionAdapterResolution {
	selectionProvided := !productionAdapterCatalogSelectionEmpty(input.CatalogSelection)
	selection := input.CatalogSelection.Normalize()
	descriptor := input.Descriptor.Normalize()
	if selectionProvided && descriptor.AdapterRef == "" {
		descriptor = selection.SelectedDescriptor.Normalize()
	}
	result := ProductionAdapterResolution{
		ContractVersion:        ContractVersion,
		Projected:              true,
		Status:                 HostActionBlocked,
		AdapterRef:             descriptor.AdapterRef,
		DescriptorRef:          descriptor.AdapterRef,
		Descriptor:             descriptor,
		ApplyEnvelopeRef:       normalizeOneDisplaySafeRef(input.ApplyEnvelopeRef),
		SelectedSourceKind:     firstReplannerSourceKind(input.SelectedSourceKind, selection.SelectedSourceKind),
		SelectedSourceRef:      firstDisplaySafeRef(input.SelectedSourceRef, selection.SelectedSourceRef),
		SelectedCandidateRef:   firstDisplaySafeRef(input.SelectedCandidateRef, selection.SelectedCandidateRef),
		RequestedAdapterRef:    firstDisplaySafeRef(input.RequestedAdapterRef, selection.SelectedAdapterRef),
		CatalogSnapshotRef:     firstDisplaySafeRef(input.CatalogSnapshotRef, selection.CatalogSnapshotRef),
		CatalogSelectionRef:    firstDisplaySafeRef(input.CatalogSelectionRef, selection.SelectionRef),
		HostPolicyRef:          normalizeOneDisplaySafeRef(input.HostPolicyRef),
		ApprovalContextRefs:    normalizeDisplaySafeRefs(input.ApprovalContextRefs),
		BudgetRef:              normalizeOneDisplaySafeRef(input.BudgetRef),
		IdempotencyRef:         normalizeOneDisplaySafeRef(input.IdempotencyRef),
		RequiredCapabilityRefs: cloneDisplaySafeRefs(descriptor.RequiresCapabilityRefs),
		RequiredPolicyRefs:     cloneDisplaySafeRefs(descriptor.RequiredPolicyRefs),
		RequiredApprovalRefs:   cloneDisplaySafeRefs(descriptor.RequiredApprovalRefs),
		RequiredPreflightRefs:  cloneDisplaySafeRefs(descriptor.PreflightCheckRefs),
		FailureClass:           FailureNone,
		Boundaries: []Boundary{
			"production_adapter_resolution",
			"resolution_projection_only",
			"host_owned_adapter_resolution",
			"display_safe_refs_only",
			"no_adapter_invocation",
			"no_runner_dispatch",
		},
		RunnerEffect:    "none",
		PromptEffect:    "none",
		RawOutputLoaded: input.RawOutputLoaded || descriptor.RawOutputLoaded,
	}
	if productionAdapterResolutionUnsafe(input) {
		result = productionAdapterBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if selectionProvided {
		result = applyProductionAdapterCatalogSelectionBinding(result, selection, descriptor)
	}
	result = applyProductionAdapterDescriptorCompleteness(result, descriptor)
	if !input.ApplyEnvelopeReady {
		result = productionAdapterBlock(result, FailureConfigMissing, "apply_envelope_not_ready", "host:source_apply_envelope", "provide_source_apply_envelope")
	}
	if result.ApplyEnvelopeRef == "" {
		result = productionAdapterBlock(result, FailureEvidenceMissing, "apply_envelope_ref_missing", "host:source_apply_envelope_ref", "provide_source_apply_envelope")
	}
	if result.RequestedAdapterRef == "" {
		result = productionAdapterBlock(result, FailureHostAdapterMissing, "requested_adapter_ref_missing", "host:apply_adapter_ref", "provide_apply_adapter_ref")
	} else if descriptor.AdapterRef != "" && result.RequestedAdapterRef != descriptor.AdapterRef && !input.AllowAdapterSubstitution {
		result = productionAdapterBlock(result, FailurePolicyBlocked, "adapter_ref_mismatch", "host:adapter_selection_review", "review_adapter_selection")
	}
	if result.SelectedSourceKind == "" {
		result = productionAdapterBlock(result, FailureInvalidInput, "source_kind_missing", "host:selected_source_kind", "provide_selected_source")
	} else if len(descriptor.SupportedSourceKinds) > 0 && !replannerSourceKindContains(descriptor.SupportedSourceKinds, result.SelectedSourceKind) {
		result = productionAdapterBlock(result, FailureUnsupportedOperation, "adapter_source_mismatch", "host:adapter_source_review", "review_adapter_source")
	}
	if result.SelectedSourceRef == "" {
		result = productionAdapterBlock(result, FailureEvidenceMissing, "source_ref_missing", "host:selected_source_ref", "provide_selected_source")
	}
	if result.SelectedCandidateRef == "" {
		result = productionAdapterBlock(result, FailureEvidenceMissing, "candidate_ref_missing", "host:selected_candidate_strategy_ref", "provide_selected_candidate")
	} else if len(descriptor.SupportedCandidateRefs) > 0 && !displaySafeRefSliceContains(descriptor.SupportedCandidateRefs, result.SelectedCandidateRef) {
		result = productionAdapterBlock(result, FailureUnsupportedOperation, "adapter_candidate_mismatch", "host:adapter_candidate_review", "review_adapter_candidate")
	}
	for _, required := range descriptor.RequiresCapabilityRefs {
		if !displaySafeRefSliceContains(input.AvailableCapabilityRefs, required) {
			result = productionAdapterBlock(result, FailureCapabilityMissing, "capability_missing", MissingInput(required), "request_capability_resolution")
			result.Boundaries = AppendBoundaries(result.Boundaries, "capability_gap_proposal_only")
		}
	}
	for _, required := range descriptor.RequiredPolicyRefs {
		if !displaySafeRefSliceContains(input.ConfirmedPolicyRefs, required) {
			result = productionAdapterBlock(result, FailurePolicyBlocked, "policy_blocked", MissingInput(required), "provide_adapter_policy")
		}
	}
	for _, required := range descriptor.RequiredApprovalRefs {
		if !displaySafeRefSliceContains(append(cloneDisplaySafeRefs(input.ConfirmedApprovalRefs), input.ApprovalContextRefs...), required) {
			result = productionAdapterBlock(result, FailureApprovalRequired, "approval_missing", MissingInput(required), "request_host_approval")
		}
	}
	if descriptor.RequiredBudgetRef != "" {
		switch {
		case result.BudgetRef == "":
			result = productionAdapterBlock(result, FailureBudgetExhausted, "budget_missing", MissingInput(descriptor.RequiredBudgetRef), "provide_adapter_budget")
		case result.BudgetRef != descriptor.RequiredBudgetRef:
			result = productionAdapterBlock(result, FailurePolicyBlocked, "budget_ref_mismatch", MissingInput(descriptor.RequiredBudgetRef), "provide_adapter_budget")
		}
	}
	if descriptor.IdempotencyContractRef != "" && result.IdempotencyRef == "" {
		result = productionAdapterBlock(result, FailureInvalidInput, "idempotency_missing", "host:idempotency_ref", "provide_idempotency_ref")
	}
	if len(result.BlockedReasons) == 0 {
		result.Status = HostActionReady
		result.ReadyForHostPreflight = true
		result.NextHostAction = "host_may_run_adapter_preflight"
		result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_host_preflight")
	}
	return result.Normalize()
}

func CloneProductionAdapterResolution(in ProductionAdapterResolution) ProductionAdapterResolution {
	out := in
	out.Descriptor = in.Descriptor.Clone()
	out.ApprovalContextRefs = cloneDisplaySafeRefs(in.ApprovalContextRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.RequiredCapabilityRefs = cloneDisplaySafeRefs(in.RequiredCapabilityRefs)
	out.RequiredPolicyRefs = cloneDisplaySafeRefs(in.RequiredPolicyRefs)
	out.RequiredApprovalRefs = cloneDisplaySafeRefs(in.RequiredApprovalRefs)
	out.RequiredPreflightRefs = cloneDisplaySafeRefs(in.RequiredPreflightRefs)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (r ProductionAdapterResolution) Clone() ProductionAdapterResolution {
	return CloneProductionAdapterResolution(r)
}

func (r ProductionAdapterResolution) Normalize() ProductionAdapterResolution {
	out := CloneProductionAdapterResolution(r)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.AdapterRef = normalizeOneDisplaySafeRef(out.AdapterRef)
	out.DescriptorRef = normalizeOneDisplaySafeRef(out.DescriptorRef)
	out.Descriptor = out.Descriptor.Normalize()
	out.ApplyEnvelopeRef = normalizeOneDisplaySafeRef(out.ApplyEnvelopeRef)
	out.SelectedSourceKind = NormalizeReplannerSourceKind(string(out.SelectedSourceKind))
	out.SelectedSourceRef = normalizeOneDisplaySafeRef(out.SelectedSourceRef)
	out.SelectedCandidateRef = normalizeOneDisplaySafeRef(out.SelectedCandidateRef)
	out.RequestedAdapterRef = normalizeOneDisplaySafeRef(out.RequestedAdapterRef)
	out.CatalogSnapshotRef = normalizeOneDisplaySafeRef(out.CatalogSnapshotRef)
	out.CatalogSelectionRef = normalizeOneDisplaySafeRef(out.CatalogSelectionRef)
	out.HostPolicyRef = normalizeOneDisplaySafeRef(out.HostPolicyRef)
	out.ApprovalContextRefs = normalizeDisplaySafeRefs(out.ApprovalContextRefs)
	out.BudgetRef = normalizeOneDisplaySafeRef(out.BudgetRef)
	out.IdempotencyRef = normalizeOneDisplaySafeRef(out.IdempotencyRef)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.BlockedReasons = normalizeControlTokenList(out.BlockedReasons)
	out.RequiredCapabilityRefs = normalizeDisplaySafeRefs(out.RequiredCapabilityRefs)
	out.RequiredPolicyRefs = normalizeDisplaySafeRefs(out.RequiredPolicyRefs)
	out.RequiredApprovalRefs = normalizeDisplaySafeRefs(out.RequiredApprovalRefs)
	out.RequiredPreflightRefs = normalizeDisplaySafeRefs(out.RequiredPreflightRefs)
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
	if out.RawOutputLoaded {
		out.Status = HostActionReviewRequired
		if out.FailureClass == FailureNone {
			out.FailureClass = FailureEvidenceWeak
		}
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		if out.NextHostAction == "" {
			out.NextHostAction = "provide_display_safe_refs"
		}
	}
	out.ReadyForHostPreflight = out.Status == HostActionReady && len(out.MissingInputs) == 0 && len(out.BlockedReasons) == 0 && !out.RawOutputLoaded
	return out
}

type ProductionAdapterPreflightInput struct {
	Resolution               ProductionAdapterResolution `json:"resolution,omitempty"`
	AdapterAvailable         bool                        `json:"adapter_available"`
	VersionStable            bool                        `json:"version_stable"`
	CapabilitiesSatisfied    bool                        `json:"capabilities_satisfied"`
	CredentialsAvailable     bool                        `json:"credentials_available"`
	AuthorizationAvailable   bool                        `json:"authorization_available"`
	HostServiceAvailable     bool                        `json:"host_service_available"`
	PolicyAllowed            bool                        `json:"policy_allowed"`
	ApprovalValid            bool                        `json:"approval_valid"`
	BudgetAvailable          bool                        `json:"budget_available"`
	IdempotencyReady         bool                        `json:"idempotency_ready"`
	TimeoutReady             bool                        `json:"timeout_ready"`
	CompensationReady        bool                        `json:"compensation_ready"`
	PreflightResultRefs      []DisplaySafeRef            `json:"preflight_result_refs,omitempty"`
	AdditionalMissingInputs  []MissingInput              `json:"additional_missing_inputs,omitempty"`
	AdditionalBlockedReasons []string                    `json:"additional_blocked_reasons,omitempty"`
	RawOutputLoaded          bool                        `json:"raw_output_loaded"`
}

type ProductionAdapterPreflight struct {
	ContractVersion        string           `json:"contract_version,omitempty"`
	Projected              bool             `json:"projected"`
	Status                 HostActionStatus `json:"status,omitempty"`
	ReadyForHostInvocation bool             `json:"ready_for_host_invocation"`
	AdapterRef             DisplaySafeRef   `json:"adapter_ref,omitempty"`
	DescriptorRef          DisplaySafeRef   `json:"descriptor_ref,omitempty"`
	IdempotencyRef         DisplaySafeRef   `json:"idempotency_ref,omitempty"`
	ApprovalRefs           []DisplaySafeRef `json:"approval_refs,omitempty"`
	PolicyRefs             []DisplaySafeRef `json:"policy_refs,omitempty"`
	BudgetRef              DisplaySafeRef   `json:"budget_ref,omitempty"`
	PreflightResultRefs    []DisplaySafeRef `json:"preflight_result_refs,omitempty"`
	MissingInputs          []MissingInput   `json:"missing_inputs,omitempty"`
	BlockedReasons         []string         `json:"blocked_reasons,omitempty"`
	FailureClass           FailureClass     `json:"failure_class,omitempty"`
	Boundaries             []Boundary       `json:"boundaries,omitempty"`
	NextHostAction         NextHostAction   `json:"next_host_action,omitempty"`
	RunnerEffect           string           `json:"runner_effect,omitempty"`
	PromptEffect           string           `json:"prompt_effect,omitempty"`
	RawOutputLoaded        bool             `json:"raw_output_loaded"`
}

func BuildProductionAdapterPreflight(input ProductionAdapterPreflightInput) ProductionAdapterPreflight {
	resolution := input.Resolution.Normalize()
	result := ProductionAdapterPreflight{
		ContractVersion:     ContractVersion,
		Projected:           true,
		Status:              HostActionBlocked,
		AdapterRef:          resolution.AdapterRef,
		DescriptorRef:       resolution.DescriptorRef,
		IdempotencyRef:      resolution.IdempotencyRef,
		ApprovalRefs:        normalizeDisplaySafeRefs(resolution.ApprovalContextRefs),
		PolicyRefs:          normalizeDisplaySafeRefs(append(cloneDisplaySafeRefs(resolution.RequiredPolicyRefs), resolution.HostPolicyRef)),
		BudgetRef:           resolution.BudgetRef,
		PreflightResultRefs: normalizeDisplaySafeRefs(input.PreflightResultRefs),
		FailureClass:        FailureNone,
		Boundaries: []Boundary{
			"production_adapter_preflight",
			"preflight_projection_only",
			"host_owned_preflight",
			"display_safe_refs_only",
			"no_adapter_invocation",
			"no_runner_dispatch",
		},
		RunnerEffect:    "none",
		PromptEffect:    "none",
		RawOutputLoaded: input.RawOutputLoaded || resolution.RawOutputLoaded,
	}
	if !resolution.ReadyForHostPreflight {
		result = productionAdapterPreflightBlock(result, FailureConfigMissing, "adapter_resolution_not_ready", "host:adapter_resolution", "provide_adapter_resolution")
		return result.Normalize()
	}
	if productionAdapterPreflightUnsafe(input) {
		result = productionAdapterPreflightBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	for _, check := range []struct {
		ok      bool
		reason  string
		missing MissingInput
		failure FailureClass
	}{
		{input.AdapterAvailable, "adapter_unavailable", "host:adapter_available", FailureHostAdapterMissing},
		{input.VersionStable, "adapter_version_drift", "host:adapter_version", FailureVerificationFailed},
		{input.CapabilitiesSatisfied, "capability_missing", "host:capability_refs", FailureCapabilityMissing},
		{input.CredentialsAvailable, "credential_missing", "host:credential_ref", FailureCredentialMissing},
		{input.AuthorizationAvailable, "authorization_missing", "host:authorization_ref", FailureAuthorizationMissing},
		{input.HostServiceAvailable, "host_service_unavailable", "host:service_ref", FailureTargetUnavailable},
		{input.PolicyAllowed, "policy_blocked", "host:policy_ref", FailurePolicyBlocked},
		{input.ApprovalValid, "approval_missing", "host:approval_ref", FailureApprovalRequired},
		{input.BudgetAvailable, "budget_missing", "host:budget_ref", FailureBudgetExhausted},
		{input.IdempotencyReady, "idempotency_missing", "host:idempotency_ref", FailureInvalidInput},
		{input.TimeoutReady, "timeout_policy_missing", "host:timeout_policy_ref", FailureTimeout},
		{input.CompensationReady, "compensation_handoff_missing", "host:compensation_handoff_ref", FailureConfigMissing},
	} {
		if !check.ok {
			result = productionAdapterPreflightBlock(result, check.failure, check.reason, check.missing, "resolve_adapter_preflight")
		}
	}
	for _, missing := range input.AdditionalMissingInputs {
		result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	}
	for _, reason := range input.AdditionalBlockedReasons {
		result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	}
	if len(result.BlockedReasons) == 0 && len(result.MissingInputs) == 0 {
		result.Status = HostActionReady
		result.ReadyForHostInvocation = true
		result.NextHostAction = "host_may_invoke_adapter"
		result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_host_invocation")
	}
	return result.Normalize()
}

func CloneProductionAdapterPreflight(in ProductionAdapterPreflight) ProductionAdapterPreflight {
	out := in
	out.ApprovalRefs = cloneDisplaySafeRefs(in.ApprovalRefs)
	out.PolicyRefs = cloneDisplaySafeRefs(in.PolicyRefs)
	out.PreflightResultRefs = cloneDisplaySafeRefs(in.PreflightResultRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (p ProductionAdapterPreflight) Clone() ProductionAdapterPreflight {
	return CloneProductionAdapterPreflight(p)
}

func (p ProductionAdapterPreflight) Normalize() ProductionAdapterPreflight {
	out := CloneProductionAdapterPreflight(p)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.AdapterRef = normalizeOneDisplaySafeRef(out.AdapterRef)
	out.DescriptorRef = normalizeOneDisplaySafeRef(out.DescriptorRef)
	out.IdempotencyRef = normalizeOneDisplaySafeRef(out.IdempotencyRef)
	out.ApprovalRefs = normalizeDisplaySafeRefs(out.ApprovalRefs)
	out.PolicyRefs = normalizeDisplaySafeRefs(out.PolicyRefs)
	out.BudgetRef = normalizeOneDisplaySafeRef(out.BudgetRef)
	out.PreflightResultRefs = normalizeDisplaySafeRefs(out.PreflightResultRefs)
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
	if out.RawOutputLoaded {
		out.Status = HostActionReviewRequired
		if out.FailureClass == FailureNone {
			out.FailureClass = FailureEvidenceWeak
		}
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		if out.NextHostAction == "" {
			out.NextHostAction = "provide_display_safe_refs"
		}
	}
	out.ReadyForHostInvocation = out.Status == HostActionReady && len(out.MissingInputs) == 0 && len(out.BlockedReasons) == 0 && !out.RawOutputLoaded
	return out
}

type HostAdapterInvocationInput struct {
	Preflight               ProductionAdapterPreflight `json:"preflight,omitempty"`
	InvocationRef           DisplaySafeRef             `json:"invocation_ref,omitempty"`
	AdapterRef              DisplaySafeRef             `json:"adapter_ref,omitempty"`
	DescriptorRef           DisplaySafeRef             `json:"descriptor_ref,omitempty"`
	IdempotencyRef          DisplaySafeRef             `json:"idempotency_ref,omitempty"`
	ApprovalRefs            []DisplaySafeRef           `json:"approval_refs,omitempty"`
	StartedEventRef         DisplaySafeRef             `json:"started_event_ref,omitempty"`
	CompletedEventRef       DisplaySafeRef             `json:"completed_event_ref,omitempty"`
	ResultRef               DisplaySafeRef             `json:"result_ref,omitempty"`
	FailureRef              DisplaySafeRef             `json:"failure_ref,omitempty"`
	ReadbackRef             DisplaySafeRef             `json:"readback_ref,omitempty"`
	CompensationRef         DisplaySafeRef             `json:"compensation_ref,omitempty"`
	CompletionHandoffRef    DisplaySafeRef             `json:"completion_handoff_ref,omitempty"`
	HostInvocationCompleted bool                       `json:"host_invocation_completed"`
	HostInvocationFailed    bool                       `json:"host_invocation_failed"`
	RawOutputLoaded         bool                       `json:"raw_output_loaded"`
}

type HostAdapterInvocationProjection struct {
	ContractVersion         string           `json:"contract_version,omitempty"`
	Projected               bool             `json:"projected"`
	Status                  HostActionStatus `json:"status,omitempty"`
	HostInvocationReported  bool             `json:"host_invocation_reported"`
	HostInvocationCompleted bool             `json:"host_invocation_completed"`
	HostInvocationFailed    bool             `json:"host_invocation_failed"`
	CoreInvocationExecuted  bool             `json:"core_invocation_executed"`
	DurableWriteByCore      bool             `json:"durable_write_by_core"`
	AdapterRef              DisplaySafeRef   `json:"adapter_ref,omitempty"`
	DescriptorRef           DisplaySafeRef   `json:"descriptor_ref,omitempty"`
	InvocationRef           DisplaySafeRef   `json:"invocation_ref,omitempty"`
	IdempotencyRef          DisplaySafeRef   `json:"idempotency_ref,omitempty"`
	ApprovalRefs            []DisplaySafeRef `json:"approval_refs,omitempty"`
	StartedEventRef         DisplaySafeRef   `json:"started_event_ref,omitempty"`
	CompletedEventRef       DisplaySafeRef   `json:"completed_event_ref,omitempty"`
	ResultRef               DisplaySafeRef   `json:"result_ref,omitempty"`
	FailureRef              DisplaySafeRef   `json:"failure_ref,omitempty"`
	ReadbackRef             DisplaySafeRef   `json:"readback_ref,omitempty"`
	CompensationRef         DisplaySafeRef   `json:"compensation_ref,omitempty"`
	CompletionHandoffRef    DisplaySafeRef   `json:"completion_handoff_ref,omitempty"`
	MissingInputs           []MissingInput   `json:"missing_inputs,omitempty"`
	BlockedReasons          []string         `json:"blocked_reasons,omitempty"`
	FailureClass            FailureClass     `json:"failure_class,omitempty"`
	Boundaries              []Boundary       `json:"boundaries,omitempty"`
	NextHostAction          NextHostAction   `json:"next_host_action,omitempty"`
	RunnerEffect            string           `json:"runner_effect,omitempty"`
	PromptEffect            string           `json:"prompt_effect,omitempty"`
	RawOutputLoaded         bool             `json:"raw_output_loaded"`
}

// BuildHostAdapterInvocationProjection reduces host-owned invocation evidence
// into the display-safe control-plane projection.
// agentx-api: internal_candidate
func BuildHostAdapterInvocationProjection(input HostAdapterInvocationInput) HostAdapterInvocationProjection {
	preflight := input.Preflight.Normalize()
	result := HostAdapterInvocationProjection{
		ContractVersion:      ContractVersion,
		Projected:            true,
		Status:               HostActionBlocked,
		AdapterRef:           firstDisplaySafeRef(input.AdapterRef, preflight.AdapterRef),
		DescriptorRef:        firstDisplaySafeRef(input.DescriptorRef, preflight.DescriptorRef),
		InvocationRef:        normalizeOneDisplaySafeRef(input.InvocationRef),
		IdempotencyRef:       firstDisplaySafeRef(input.IdempotencyRef, preflight.IdempotencyRef),
		ApprovalRefs:         normalizeDisplaySafeRefs(append(cloneDisplaySafeRefs(input.ApprovalRefs), preflight.ApprovalRefs...)),
		StartedEventRef:      normalizeOneDisplaySafeRef(input.StartedEventRef),
		CompletedEventRef:    normalizeOneDisplaySafeRef(input.CompletedEventRef),
		ResultRef:            normalizeOneDisplaySafeRef(input.ResultRef),
		FailureRef:           normalizeOneDisplaySafeRef(input.FailureRef),
		ReadbackRef:          normalizeOneDisplaySafeRef(input.ReadbackRef),
		CompensationRef:      normalizeOneDisplaySafeRef(input.CompensationRef),
		CompletionHandoffRef: normalizeOneDisplaySafeRef(input.CompletionHandoffRef),
		FailureClass:         FailureNone,
		Boundaries: []Boundary{
			"host_adapter_invocation_projection",
			"host_owned_invocation",
			"core_invocation_not_executed",
			"no_durable_write_by_core",
			"display_safe_refs_only",
			"no_runner_dispatch",
		},
		RunnerEffect:    "none",
		PromptEffect:    "none",
		RawOutputLoaded: input.RawOutputLoaded || preflight.RawOutputLoaded,
	}
	if !preflight.ReadyForHostInvocation {
		result = hostAdapterInvocationBlock(result, FailureConfigMissing, "adapter_preflight_not_ready", "host:adapter_preflight", "provide_adapter_preflight")
		return result.Normalize()
	}
	if hostAdapterInvocationUnsafe(input) {
		result = hostAdapterInvocationBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if result.InvocationRef == "" {
		result = hostAdapterInvocationBlock(result, FailureEvidenceMissing, "invocation_ref_missing", "host:invocation_ref", "provide_invocation_ref")
	}
	if result.IdempotencyRef == "" {
		result = hostAdapterInvocationBlock(result, FailureInvalidInput, "idempotency_missing", "host:idempotency_ref", "provide_idempotency_ref")
	}
	if result.StartedEventRef == "" {
		result = hostAdapterInvocationBlock(result, FailureEvidenceMissing, "started_event_ref_missing", "host:started_event_ref", "provide_started_event_ref")
	}
	if input.HostInvocationCompleted && input.HostInvocationFailed {
		result = hostAdapterInvocationBlock(result, FailureVerificationFailed, "invocation_result_conflict", "host:invocation_result_review", "review_invocation_result")
		return result.Normalize()
	}
	switch {
	case input.HostInvocationCompleted:
		if result.CompletedEventRef == "" {
			result = hostAdapterInvocationBlock(result, FailureEvidenceMissing, "completed_event_ref_missing", "host:completed_event_ref", "provide_completed_event_ref")
		}
		if result.ResultRef == "" {
			result = hostAdapterInvocationBlock(result, FailureEvidenceMissing, "result_ref_missing", "host:result_ref", "provide_result_ref")
		}
		if result.ReadbackRef == "" {
			result = hostAdapterInvocationBlock(result, FailureEvidenceMissing, "readback_ref_missing", "host:readback_ref", "provide_readback_ref")
		}
		if result.CompletionHandoffRef == "" {
			result = hostAdapterInvocationBlock(result, FailureEvidenceMissing, "completion_handoff_ref_missing", "host:completion_handoff_ref", "provide_completion_handoff_ref")
		}
		if len(result.BlockedReasons) == 0 && len(result.MissingInputs) == 0 {
			result.Status = HostActionRecorded
			result.HostInvocationReported = true
			result.HostInvocationCompleted = true
			result.NextHostAction = "review_adapter_readback"
			result.Boundaries = AppendBoundaries(result.Boundaries, "host_reported_invocation_success")
		}
	case input.HostInvocationFailed:
		if result.CompletedEventRef == "" {
			result = hostAdapterInvocationBlock(result, FailureEvidenceMissing, "completed_event_ref_missing", "host:completed_event_ref", "provide_completed_event_ref")
		}
		if result.FailureRef == "" {
			result = hostAdapterInvocationBlock(result, FailureEvidenceMissing, "failure_ref_missing", "host:failure_ref", "provide_failure_ref")
		}
		if result.CompensationRef == "" {
			result = hostAdapterInvocationBlock(result, FailureEvidenceMissing, "compensation_ref_missing", "host:compensation_ref", "provide_compensation_ref")
		}
		if len(result.BlockedReasons) == 0 && len(result.MissingInputs) == 0 {
			result.Status = HostActionRecorded
			result.HostInvocationReported = true
			result.HostInvocationFailed = true
			result.FailureClass = FailureVerificationFailed
			result.NextHostAction = "review_adapter_failure"
			result.Boundaries = AppendBoundaries(result.Boundaries, "host_reported_invocation_failure")
		}
	default:
		result = hostAdapterInvocationBlock(result, FailureEvidenceMissing, "invocation_result_missing", "host:invocation_result", "provide_invocation_result")
	}
	return result.Normalize()
}

func CloneHostAdapterInvocationProjection(in HostAdapterInvocationProjection) HostAdapterInvocationProjection {
	out := in
	out.ApprovalRefs = cloneDisplaySafeRefs(in.ApprovalRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (p HostAdapterInvocationProjection) Clone() HostAdapterInvocationProjection {
	return CloneHostAdapterInvocationProjection(p)
}

func (p HostAdapterInvocationProjection) Normalize() HostAdapterInvocationProjection {
	out := CloneHostAdapterInvocationProjection(p)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.AdapterRef = normalizeOneDisplaySafeRef(out.AdapterRef)
	out.DescriptorRef = normalizeOneDisplaySafeRef(out.DescriptorRef)
	out.InvocationRef = normalizeOneDisplaySafeRef(out.InvocationRef)
	out.IdempotencyRef = normalizeOneDisplaySafeRef(out.IdempotencyRef)
	out.ApprovalRefs = normalizeDisplaySafeRefs(out.ApprovalRefs)
	out.StartedEventRef = normalizeOneDisplaySafeRef(out.StartedEventRef)
	out.CompletedEventRef = normalizeOneDisplaySafeRef(out.CompletedEventRef)
	out.ResultRef = normalizeOneDisplaySafeRef(out.ResultRef)
	out.FailureRef = normalizeOneDisplaySafeRef(out.FailureRef)
	out.ReadbackRef = normalizeOneDisplaySafeRef(out.ReadbackRef)
	out.CompensationRef = normalizeOneDisplaySafeRef(out.CompensationRef)
	out.CompletionHandoffRef = normalizeOneDisplaySafeRef(out.CompletionHandoffRef)
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
	if out.RawOutputLoaded {
		out.Status = HostActionReviewRequired
		if out.FailureClass == FailureNone {
			out.FailureClass = FailureEvidenceWeak
		}
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		if out.NextHostAction == "" {
			out.NextHostAction = "provide_display_safe_refs"
		}
	}
	out.CoreInvocationExecuted = false
	out.DurableWriteByCore = false
	out.HostInvocationReported = out.Status == HostActionRecorded && (out.HostInvocationCompleted || out.HostInvocationFailed)
	return out
}

type ProductionAdapterReadbackReview struct {
	ContractVersion            string           `json:"contract_version,omitempty"`
	Projected                  bool             `json:"projected"`
	Status                     HostActionStatus `json:"status,omitempty"`
	ReadyForReadbackReview     bool             `json:"ready_for_readback_review"`
	ReadyForFailureReview      bool             `json:"ready_for_failure_review"`
	ReadyForCompletionAudit    bool             `json:"ready_for_completion_audit"`
	AuthorizationBound         bool             `json:"authorization_bound"`
	ObjectiveSatisfied         bool             `json:"objective_satisfied"`
	HostInvocationReported     bool             `json:"host_invocation_reported"`
	HostInvocationCompleted    bool             `json:"host_invocation_completed"`
	HostInvocationFailed       bool             `json:"host_invocation_failed"`
	CoreInvocationExecuted     bool             `json:"core_invocation_executed"`
	DurableWriteByCore         bool             `json:"durable_write_by_core"`
	InvocationReportBindingRef DisplaySafeRef   `json:"invocation_report_binding_ref,omitempty"`
	AuthorizationPacketRef     DisplaySafeRef   `json:"authorization_packet_ref,omitempty"`
	PreflightReviewPacketRef   DisplaySafeRef   `json:"preflight_review_packet_ref,omitempty"`
	AdapterRef                 DisplaySafeRef   `json:"adapter_ref,omitempty"`
	DescriptorRef              DisplaySafeRef   `json:"descriptor_ref,omitempty"`
	InvocationRef              DisplaySafeRef   `json:"invocation_ref,omitempty"`
	IdempotencyRef             DisplaySafeRef   `json:"idempotency_ref,omitempty"`
	ResultRef                  DisplaySafeRef   `json:"result_ref,omitempty"`
	FailureRef                 DisplaySafeRef   `json:"failure_ref,omitempty"`
	ReadbackRef                DisplaySafeRef   `json:"readback_ref,omitempty"`
	CompensationRef            DisplaySafeRef   `json:"compensation_ref,omitempty"`
	CompletionHandoffRef       DisplaySafeRef   `json:"completion_handoff_ref,omitempty"`
	EvidenceRefs               []EvidenceRef    `json:"evidence_refs,omitempty"`
	MissingInputs              []MissingInput   `json:"missing_inputs,omitempty"`
	BlockedReasons             []string         `json:"blocked_reasons,omitempty"`
	FailureClass               FailureClass     `json:"failure_class,omitempty"`
	Boundaries                 []Boundary       `json:"boundaries,omitempty"`
	NextHostAction             NextHostAction   `json:"next_host_action,omitempty"`
	RunnerEffect               string           `json:"runner_effect,omitempty"`
	PromptEffect               string           `json:"prompt_effect,omitempty"`
	RawOutputLoaded            bool             `json:"raw_output_loaded"`
}

func BuildProductionAdapterReadbackReview(input HostAdapterInvocationProjection) ProductionAdapterReadbackReview {
	unsafe := productionAdapterInvocationProjectionUnsafe(input)
	invocation := input.Normalize()
	result := ProductionAdapterReadbackReview{
		ContractVersion:         ContractVersion,
		Projected:               true,
		Status:                  HostActionBlocked,
		HostInvocationReported:  invocation.HostInvocationReported,
		HostInvocationCompleted: invocation.HostInvocationCompleted,
		HostInvocationFailed:    invocation.HostInvocationFailed,
		AdapterRef:              invocation.AdapterRef,
		DescriptorRef:           invocation.DescriptorRef,
		InvocationRef:           invocation.InvocationRef,
		IdempotencyRef:          invocation.IdempotencyRef,
		ResultRef:               invocation.ResultRef,
		FailureRef:              invocation.FailureRef,
		ReadbackRef:             invocation.ReadbackRef,
		CompensationRef:         invocation.CompensationRef,
		CompletionHandoffRef:    invocation.CompletionHandoffRef,
		EvidenceRefs:            productionAdapterReadbackEvidenceRefs(invocation),
		FailureClass:            FailureNone,
		Boundaries: []Boundary{
			"production_adapter_readback_review",
			"host_invocation_result_view_only",
			"readback_projection_only",
			"completion_handoff_projection_only",
			"objective_not_satisfied_by_adapter",
			"display_safe_refs_only",
			"display_safe_result_refs_only",
			"no_runner_dispatch",
			"no_durable_write_by_core",
		},
		RunnerEffect:    "none",
		PromptEffect:    "none",
		RawOutputLoaded: input.RawOutputLoaded || invocation.RawOutputLoaded,
	}
	if unsafe {
		result = productionAdapterReadbackReviewBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if invocation.Status != HostActionRecorded || !invocation.HostInvocationReported {
		result = productionAdapterReadbackReviewBlock(result, FailureConfigMissing, "adapter_invocation_not_recorded", "host:adapter_invocation", "provide_adapter_invocation")
		return result.Normalize()
	}
	if invocation.HostInvocationCompleted && invocation.HostInvocationFailed {
		result = productionAdapterReadbackReviewBlock(result, FailureVerificationFailed, "invocation_result_conflict", "host:invocation_result_review", "review_invocation_result")
		return result.Normalize()
	}
	switch {
	case invocation.HostInvocationCompleted:
		if result.ResultRef == "" {
			result = productionAdapterReadbackReviewBlock(result, FailureEvidenceMissing, "result_ref_missing", "host:result_ref", "provide_result_ref")
		}
		if result.ReadbackRef == "" {
			result = productionAdapterReadbackReviewBlock(result, FailureEvidenceMissing, "readback_ref_missing", "host:readback_ref", "provide_readback_ref")
		}
		if result.CompletionHandoffRef == "" {
			result = productionAdapterReadbackReviewBlock(result, FailureEvidenceMissing, "completion_handoff_ref_missing", "host:completion_handoff_ref", "provide_completion_handoff_ref")
		}
		if len(result.BlockedReasons) == 0 && len(result.MissingInputs) == 0 {
			result.Status = HostActionReady
			result.ReadyForReadbackReview = true
			result.ReadyForCompletionAudit = true
			result.NextHostAction = "review_completion_handoff"
			result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_completion_audit")
		}
	case invocation.HostInvocationFailed:
		if result.FailureRef == "" {
			result = productionAdapterReadbackReviewBlock(result, FailureEvidenceMissing, "failure_ref_missing", "host:failure_ref", "provide_failure_ref")
		}
		if result.CompensationRef == "" {
			result = productionAdapterReadbackReviewBlock(result, FailureEvidenceMissing, "compensation_ref_missing", "host:compensation_ref", "provide_compensation_ref")
		}
		if len(result.BlockedReasons) == 0 && len(result.MissingInputs) == 0 {
			result.Status = HostActionReviewRequired
			result.ReadyForFailureReview = true
			result.FailureClass = FailureVerificationFailed
			result.NextHostAction = "review_adapter_failure"
			result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_failure_review")
		}
	default:
		result = productionAdapterReadbackReviewBlock(result, FailureEvidenceMissing, "invocation_result_missing", "host:invocation_result", "provide_invocation_result")
	}
	return result.Normalize()
}

func BuildProductionAdapterReadbackReviewFromBinding(input ProductionAdapterInvocationReportBinding) ProductionAdapterReadbackReview {
	bindingEmpty := productionAdapterInvocationReportBindingEmpty(input)
	bindingUnsafe := productionAdapterInvocationReportBindingOutputUnsafe(input)
	binding := input.Normalize()
	invocation := productionAdapterInvocationFromReportBinding(binding)
	result := BuildProductionAdapterReadbackReview(invocation)
	result.InvocationReportBindingRef = binding.InvocationReportBindingRef
	result.AuthorizationPacketRef = binding.AuthorizationPacketRef
	result.PreflightReviewPacketRef = binding.PreflightReviewPacketRef
	result.AuthorizationBound = binding.AuthorizationBound
	result.HostInvocationReported = binding.HostInvocationReported
	result.HostInvocationCompleted = binding.HostInvocationCompleted
	result.HostInvocationFailed = binding.HostInvocationFailed
	result.Boundaries = AppendBoundaries(result.Boundaries,
		"readback_review_from_invocation_report_binding",
		"authorization_bound_readback_review",
	)
	if binding.InvocationReportBindingRef != "" {
		result.EvidenceRefs = normalizeEvidenceRefs(append(result.EvidenceRefs, EvidenceRef{
			Ref:      binding.InvocationReportBindingRef,
			Kind:     "invocation_report_binding",
			Strength: EvidenceAdequate,
			Source:   firstDisplaySafeRef(binding.AuthorizationPacketRef, binding.InvocationRef),
		}))
	}
	result.RawOutputLoaded = result.RawOutputLoaded || binding.RawOutputLoaded
	if bindingEmpty {
		result.MissingInputs = nil
		result.BlockedReasons = nil
		result.FailureClass = FailureNone
		result.AuthorizationBound = false
		result = productionAdapterReadbackReviewBlock(result, FailureEvidenceMissing, "invocation_report_binding_missing", "host:invocation_report_binding", "provide_invocation_report_binding")
		return result.Normalize()
	}
	if bindingUnsafe {
		result.MissingInputs = nil
		result.BlockedReasons = nil
		result.FailureClass = FailureNone
		result.AuthorizationBound = false
		result = productionAdapterReadbackReviewBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if !binding.AuthorizationBound || (!binding.ReadyForReadbackReview && !binding.ReadyForFailureReview) {
		result.MissingInputs = nil
		result.BlockedReasons = nil
		result.FailureClass = FailureNone
		result.AuthorizationBound = false
		result = productionAdapterReadbackReviewBlock(result, firstFailureClass(binding.FailureClass, FailureAuthorizationMissing), "invocation_report_binding_not_ready", "host:invocation_report_binding", firstNextHostAction(binding.NextHostAction, "review_adapter_invocation_report"))
		return result.Normalize()
	}
	return result.Normalize()
}

func CloneProductionAdapterReadbackReview(in ProductionAdapterReadbackReview) ProductionAdapterReadbackReview {
	out := in
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (r ProductionAdapterReadbackReview) Clone() ProductionAdapterReadbackReview {
	return CloneProductionAdapterReadbackReview(r)
}

func (r ProductionAdapterReadbackReview) Normalize() ProductionAdapterReadbackReview {
	out := CloneProductionAdapterReadbackReview(r)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.AdapterRef = normalizeOneDisplaySafeRef(out.AdapterRef)
	out.DescriptorRef = normalizeOneDisplaySafeRef(out.DescriptorRef)
	out.InvocationRef = normalizeOneDisplaySafeRef(out.InvocationRef)
	out.IdempotencyRef = normalizeOneDisplaySafeRef(out.IdempotencyRef)
	out.ResultRef = normalizeOneDisplaySafeRef(out.ResultRef)
	out.FailureRef = normalizeOneDisplaySafeRef(out.FailureRef)
	out.ReadbackRef = normalizeOneDisplaySafeRef(out.ReadbackRef)
	out.CompensationRef = normalizeOneDisplaySafeRef(out.CompensationRef)
	out.CompletionHandoffRef = normalizeOneDisplaySafeRef(out.CompletionHandoffRef)
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
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
	if out.RawOutputLoaded {
		out.Status = HostActionReviewRequired
		if out.FailureClass == FailureNone {
			out.FailureClass = FailureEvidenceWeak
		}
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		if out.NextHostAction == "" {
			out.NextHostAction = "provide_display_safe_refs"
		}
	}
	out.ObjectiveSatisfied = false
	out.CoreInvocationExecuted = false
	out.DurableWriteByCore = false
	out.InvocationReportBindingRef = normalizeOneDisplaySafeRef(out.InvocationReportBindingRef)
	out.AuthorizationPacketRef = normalizeOneDisplaySafeRef(out.AuthorizationPacketRef)
	out.PreflightReviewPacketRef = normalizeOneDisplaySafeRef(out.PreflightReviewPacketRef)
	out.ReadyForReadbackReview = out.Status == HostActionReady && out.HostInvocationCompleted && len(out.MissingInputs) == 0 && len(out.BlockedReasons) == 0 && !out.RawOutputLoaded
	out.ReadyForFailureReview = out.Status == HostActionReviewRequired && out.HostInvocationFailed && out.FailureRef != "" && out.CompensationRef != "" && len(out.MissingInputs) == 0 && len(out.BlockedReasons) == 0 && !out.RawOutputLoaded
	out.ReadyForCompletionAudit = out.ReadyForReadbackReview && out.CompletionHandoffRef != ""
	out.AuthorizationBound = out.AuthorizationBound &&
		out.InvocationReportBindingRef != "" &&
		out.AuthorizationPacketRef != "" &&
		out.PreflightReviewPacketRef != "" &&
		(out.ReadyForReadbackReview || out.ReadyForFailureReview)
	return out
}

type ProductionAdapterCompletionHandoff struct {
	ContractVersion            string             `json:"contract_version,omitempty"`
	Projected                  bool               `json:"projected"`
	Status                     VerificationStatus `json:"status,omitempty"`
	ReadyForCompletionAudit    bool               `json:"ready_for_completion_audit"`
	ObjectiveSatisfied         bool               `json:"objective_satisfied"`
	HostInvocationReported     bool               `json:"host_invocation_reported"`
	AuthorizationBound         bool               `json:"authorization_bound"`
	InvocationReportBindingRef DisplaySafeRef     `json:"invocation_report_binding_ref,omitempty"`
	AuthorizationPacketRef     DisplaySafeRef     `json:"authorization_packet_ref,omitempty"`
	PreflightReviewPacketRef   DisplaySafeRef     `json:"preflight_review_packet_ref,omitempty"`
	AdapterRef                 DisplaySafeRef     `json:"adapter_ref,omitempty"`
	DescriptorRef              DisplaySafeRef     `json:"descriptor_ref,omitempty"`
	InvocationRef              DisplaySafeRef     `json:"invocation_ref,omitempty"`
	ResultRef                  DisplaySafeRef     `json:"result_ref,omitempty"`
	ReadbackRef                DisplaySafeRef     `json:"readback_ref,omitempty"`
	CompletionHandoffRef       DisplaySafeRef     `json:"completion_handoff_ref,omitempty"`
	EvidenceRefs               []EvidenceRef      `json:"evidence_refs,omitempty"`
	Verification               VerificationResult `json:"verification,omitempty"`
	MissingInputs              []MissingInput     `json:"missing_inputs,omitempty"`
	BlockedReasons             []string           `json:"blocked_reasons,omitempty"`
	FailureClass               FailureClass       `json:"failure_class,omitempty"`
	Boundaries                 []Boundary         `json:"boundaries,omitempty"`
	NextHostAction             NextHostAction     `json:"next_host_action,omitempty"`
	RunnerEffect               string             `json:"runner_effect,omitempty"`
	PromptEffect               string             `json:"prompt_effect,omitempty"`
	RawOutputLoaded            bool               `json:"raw_output_loaded"`
}

func BuildProductionAdapterCompletionHandoff(input ProductionAdapterReadbackReview) ProductionAdapterCompletionHandoff {
	unsafe := productionAdapterReadbackReviewUnsafe(input)
	review := input.Normalize()
	result := ProductionAdapterCompletionHandoff{
		ContractVersion:            ContractVersion,
		Projected:                  true,
		Status:                     VerificationBlocked,
		HostInvocationReported:     review.HostInvocationReported,
		AuthorizationBound:         review.AuthorizationBound,
		InvocationReportBindingRef: review.InvocationReportBindingRef,
		AuthorizationPacketRef:     review.AuthorizationPacketRef,
		PreflightReviewPacketRef:   review.PreflightReviewPacketRef,
		AdapterRef:                 review.AdapterRef,
		DescriptorRef:              review.DescriptorRef,
		InvocationRef:              review.InvocationRef,
		ResultRef:                  review.ResultRef,
		ReadbackRef:                review.ReadbackRef,
		CompletionHandoffRef:       review.CompletionHandoffRef,
		EvidenceRefs:               productionAdapterCompletionEvidenceRefs(review),
		FailureClass:               FailureNone,
		Boundaries: []Boundary{
			"production_adapter_completion_handoff",
			"completion_handoff_projection_only",
			"completion_audit_input_only",
			"objective_not_satisfied_by_adapter",
			"display_safe_refs_only",
			"display_safe_result_refs_only",
			"no_runner_dispatch",
			"no_durable_write_by_core",
		},
		RunnerEffect:    "none",
		PromptEffect:    "none",
		RawOutputLoaded: input.RawOutputLoaded || review.RawOutputLoaded,
	}
	if unsafe {
		result = productionAdapterCompletionHandoffBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if !review.ReadyForCompletionAudit {
		result = productionAdapterCompletionHandoffBlock(result, firstFailureClass(review.FailureClass, FailureEvidenceMissing), "adapter_readback_not_ready", "host:production_adapter_readback_review", firstNextHostAction(review.NextHostAction, "review_adapter_readback"))
		return result.Normalize()
	}
	if !review.AuthorizationBound {
		result = productionAdapterCompletionHandoffBlock(result, firstFailureClass(review.FailureClass, FailureAuthorizationMissing), "adapter_readback_not_authorization_bound", "host:authorization_bound_readback_review", "review_adapter_readback_authorization")
		return result.Normalize()
	}
	result.Status = VerificationReviewRequired
	result.ReadyForCompletionAudit = true
	result.FailureClass = FailureEvidenceMissing
	result.Boundaries = AppendBoundaries(result.Boundaries, "completion_handoff_from_authorization_bound_readback", "authorization_bound_completion_handoff")
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, "host:completion_audit_result")
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, "completion_audit_required")
	result.NextHostAction = "run_completion_audit"
	result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_completion_audit")
	result.Verification = VerificationResult{
		ContractVersion: ContractVersion,
		Status:          VerificationReviewRequired,
		FailureClass:    FailureEvidenceMissing,
		EvidenceRefs:    cloneEvidenceRefs(result.EvidenceRefs),
		MissingInputs:   []MissingInput{"host:completion_audit_result"},
		Boundaries:      AppendBoundaries(result.Boundaries, "host_reported_success_not_objective_success"),
		NextHostAction:  "run_completion_audit",
	}.Normalize()
	return result.Normalize()
}

func CloneProductionAdapterCompletionHandoff(in ProductionAdapterCompletionHandoff) ProductionAdapterCompletionHandoff {
	out := in
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.Verification = in.Verification.Clone()
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (h ProductionAdapterCompletionHandoff) Clone() ProductionAdapterCompletionHandoff {
	return CloneProductionAdapterCompletionHandoff(h)
}

func (h ProductionAdapterCompletionHandoff) Normalize() ProductionAdapterCompletionHandoff {
	out := CloneProductionAdapterCompletionHandoff(h)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeVerificationStatus(string(out.Status))
	out.AdapterRef = normalizeOneDisplaySafeRef(out.AdapterRef)
	out.DescriptorRef = normalizeOneDisplaySafeRef(out.DescriptorRef)
	out.InvocationRef = normalizeOneDisplaySafeRef(out.InvocationRef)
	out.ResultRef = normalizeOneDisplaySafeRef(out.ResultRef)
	out.ReadbackRef = normalizeOneDisplaySafeRef(out.ReadbackRef)
	out.CompletionHandoffRef = normalizeOneDisplaySafeRef(out.CompletionHandoffRef)
	out.InvocationReportBindingRef = normalizeOneDisplaySafeRef(out.InvocationReportBindingRef)
	out.AuthorizationPacketRef = normalizeOneDisplaySafeRef(out.AuthorizationPacketRef)
	out.PreflightReviewPacketRef = normalizeOneDisplaySafeRef(out.PreflightReviewPacketRef)
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
	out.Verification = out.Verification.Normalize()
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
	if out.RawOutputLoaded {
		out.Status = VerificationReviewRequired
		if out.FailureClass == FailureNone {
			out.FailureClass = FailureEvidenceWeak
		}
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		if out.NextHostAction == "" {
			out.NextHostAction = "provide_display_safe_refs"
		}
	}
	out.ObjectiveSatisfied = false
	out.AuthorizationBound = out.AuthorizationBound &&
		out.InvocationReportBindingRef != "" &&
		out.AuthorizationPacketRef != "" &&
		out.PreflightReviewPacketRef != ""
	out.ReadyForCompletionAudit = out.Status == VerificationReviewRequired && out.AuthorizationBound && out.CompletionHandoffRef != "" && !out.RawOutputLoaded && !containsMissingInput(out.MissingInputs, "host:production_adapter_readback_review") && !containsMissingInput(out.MissingInputs, "host:authorization_bound_readback_review") && !containsMissingInput(out.MissingInputs, "host:display_safe_refs")
	out.Verification.Satisfied = false
	return out
}

type ProductionAdapterReadbackHostView struct {
	ContractVersion            string             `json:"contract_version,omitempty"`
	Projected                  bool               `json:"projected"`
	Available                  bool               `json:"available"`
	Status                     string             `json:"status,omitempty"`
	Mode                       string             `json:"mode,omitempty"`
	RunnerEffect               string             `json:"runner_effect,omitempty"`
	PromptEffect               string             `json:"prompt_effect,omitempty"`
	ReadbackStatus             HostActionStatus   `json:"readback_status,omitempty"`
	HandoffStatus              VerificationStatus `json:"handoff_status,omitempty"`
	ReadyForHostReview         bool               `json:"ready_for_host_review"`
	ReadyForReadbackReview     bool               `json:"ready_for_readback_review"`
	ReadyForFailureReview      bool               `json:"ready_for_failure_review"`
	ReadyForCompletionAudit    bool               `json:"ready_for_completion_audit"`
	ObjectiveSatisfied         bool               `json:"objective_satisfied"`
	VerificationSatisfied      bool               `json:"verification_satisfied"`
	HostInvocationReported     bool               `json:"host_invocation_reported"`
	HostInvocationCompleted    bool               `json:"host_invocation_completed"`
	HostInvocationFailed       bool               `json:"host_invocation_failed"`
	CoreInvocationExecuted     bool               `json:"core_invocation_executed"`
	DurableWriteByCore         bool               `json:"durable_write_by_core"`
	AuthorizationBound         bool               `json:"authorization_bound"`
	AuthorizationChainStatus   string             `json:"authorization_chain_status,omitempty"`
	InvocationReportBindingRef DisplaySafeRef     `json:"invocation_report_binding_ref,omitempty"`
	AuthorizationPacketRef     DisplaySafeRef     `json:"authorization_packet_ref,omitempty"`
	PreflightReviewPacketRef   DisplaySafeRef     `json:"preflight_review_packet_ref,omitempty"`
	AuthorizationMissingInputs []MissingInput     `json:"authorization_missing_inputs,omitempty"`
	AdapterRef                 DisplaySafeRef     `json:"adapter_ref,omitempty"`
	DescriptorRef              DisplaySafeRef     `json:"descriptor_ref,omitempty"`
	InvocationRef              DisplaySafeRef     `json:"invocation_ref,omitempty"`
	IdempotencyRef             DisplaySafeRef     `json:"idempotency_ref,omitempty"`
	ResultRef                  DisplaySafeRef     `json:"result_ref,omitempty"`
	FailureRef                 DisplaySafeRef     `json:"failure_ref,omitempty"`
	ReadbackRef                DisplaySafeRef     `json:"readback_ref,omitempty"`
	CompensationRef            DisplaySafeRef     `json:"compensation_ref,omitempty"`
	CompletionHandoffRef       DisplaySafeRef     `json:"completion_handoff_ref,omitempty"`
	EvidenceRefs               []EvidenceRef      `json:"evidence_refs,omitempty"`
	Verification               VerificationResult `json:"verification,omitempty"`
	MissingInputs              []MissingInput     `json:"missing_inputs,omitempty"`
	BlockedReasons             []string           `json:"blocked_reasons,omitempty"`
	FailureClass               FailureClass       `json:"failure_class,omitempty"`
	Boundaries                 []Boundary         `json:"boundaries,omitempty"`
	NextHostAction             NextHostAction     `json:"next_host_action,omitempty"`
	RawOutputLoaded            bool               `json:"raw_output_loaded"`
}

func BuildProductionAdapterReadbackHostView(review ProductionAdapterReadbackReview, handoff ProductionAdapterCompletionHandoff) ProductionAdapterReadbackHostView {
	if productionAdapterReadbackReviewEmpty(review) {
		return unavailableProductionAdapterReadbackHostView()
	}
	unsafe := productionAdapterReadbackReviewUnsafe(review) || productionAdapterCompletionHandoffUnsafe(handoff)
	normalizedReview := review.Normalize()
	normalizedHandoff := handoff.Normalize()
	result := ProductionAdapterReadbackHostView{
		ContractVersion:            ContractVersion,
		Projected:                  true,
		Available:                  true,
		Status:                     "blocked",
		Mode:                       "production_adapter_readback_host_view",
		RunnerEffect:               "none",
		PromptEffect:               "none",
		ReadbackStatus:             normalizedReview.Status,
		HandoffStatus:              normalizedHandoff.Status,
		HostInvocationReported:     normalizedReview.HostInvocationReported,
		HostInvocationCompleted:    normalizedReview.HostInvocationCompleted,
		HostInvocationFailed:       normalizedReview.HostInvocationFailed,
		AuthorizationBound:         productionAdapterHostViewAuthorizationBound(normalizedReview, normalizedHandoff),
		AuthorizationChainStatus:   productionAdapterAuthorizationChainStatus(normalizedReview, normalizedHandoff),
		InvocationReportBindingRef: firstDisplaySafeRef(normalizedHandoff.InvocationReportBindingRef, normalizedReview.InvocationReportBindingRef),
		AuthorizationPacketRef:     firstDisplaySafeRef(normalizedHandoff.AuthorizationPacketRef, normalizedReview.AuthorizationPacketRef),
		PreflightReviewPacketRef:   firstDisplaySafeRef(normalizedHandoff.PreflightReviewPacketRef, normalizedReview.PreflightReviewPacketRef),
		AuthorizationMissingInputs: productionAdapterAuthorizationMissingInputs(normalizedReview, normalizedHandoff),
		AdapterRef:                 normalizedReview.AdapterRef,
		DescriptorRef:              normalizedReview.DescriptorRef,
		InvocationRef:              normalizedReview.InvocationRef,
		IdempotencyRef:             normalizedReview.IdempotencyRef,
		ResultRef:                  normalizedReview.ResultRef,
		FailureRef:                 normalizedReview.FailureRef,
		ReadbackRef:                normalizedReview.ReadbackRef,
		CompensationRef:            normalizedReview.CompensationRef,
		CompletionHandoffRef:       normalizedReview.CompletionHandoffRef,
		EvidenceRefs:               productionAdapterHostViewEvidenceRefs(normalizedReview, normalizedHandoff),
		Verification:               normalizedHandoff.Verification,
		MissingInputs:              productionAdapterHostViewMissingInputs(normalizedReview, normalizedHandoff),
		BlockedReasons:             productionAdapterHostViewBlockedReasons(normalizedReview, normalizedHandoff),
		FailureClass:               firstFailureClass(normalizedHandoff.FailureClass, normalizedReview.FailureClass),
		Boundaries:                 productionAdapterReadbackHostViewBoundaries(normalizedReview.Boundaries, normalizedHandoff.Boundaries),
		NextHostAction:             firstNextHostAction(normalizedHandoff.NextHostAction, normalizedReview.NextHostAction),
		RawOutputLoaded:            review.RawOutputLoaded || handoff.RawOutputLoaded || normalizedReview.RawOutputLoaded || normalizedHandoff.RawOutputLoaded,
	}
	if unsafe {
		result.HostInvocationReported = false
		result.HostInvocationCompleted = false
		result.HostInvocationFailed = false
		result.AuthorizationBound = false
		result.AuthorizationChainStatus = "authorization_blocked"
		result.ReadyForHostReview = false
		result.ReadyForReadbackReview = false
		result.ReadyForFailureReview = false
		result.ReadyForCompletionAudit = false
		result.FailureClass = FailureEvidenceWeak
		result.MissingInputs = AppendMissingInputs(result.MissingInputs, "host:display_safe_refs")
		result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, "unsafe_input_ref")
		result.NextHostAction = "provide_display_safe_refs"
		return result.Normalize()
	}
	switch {
	case normalizedReview.ReadyForFailureReview:
		result.Status = "ready_for_failure_review"
		result.ReadyForHostReview = true
		result.ReadyForFailureReview = true
		result.FailureClass = normalizedReview.FailureClass
		result.NextHostAction = firstNextHostAction(normalizedReview.NextHostAction, "review_adapter_failure")
		result.MissingInputs = nil
		result.BlockedReasons = nil
		result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_failure_review")
	case normalizedReview.ReadyForCompletionAudit:
		if !normalizedHandoff.ReadyForCompletionAudit {
			result.FailureClass = firstFailureClass(normalizedHandoff.FailureClass, FailureEvidenceMissing)
			result.MissingInputs = AppendMissingInputs(result.MissingInputs, "host:production_adapter_completion_handoff")
			result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, "completion_handoff_not_ready")
			result.NextHostAction = firstNextHostAction(normalizedHandoff.NextHostAction, "run_completion_audit")
			return result.Normalize()
		}
		result.Status = "ready_for_completion_audit"
		result.ReadyForHostReview = true
		result.ReadyForReadbackReview = true
		result.ReadyForCompletionAudit = true
		result.FailureClass = firstFailureClass(normalizedHandoff.FailureClass, FailureEvidenceMissing)
		result.MissingInputs = cloneMissingInputs(normalizedHandoff.MissingInputs)
		result.BlockedReasons = cloneStringSlice(normalizedHandoff.BlockedReasons)
		result.NextHostAction = firstNextHostAction(normalizedHandoff.NextHostAction, "run_completion_audit")
		result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_completion_audit")
	default:
		result.FailureClass = firstFailureClass(normalizedReview.FailureClass, FailureEvidenceMissing)
		if len(result.MissingInputs) == 0 {
			result.MissingInputs = AppendMissingInputs(result.MissingInputs, "host:production_adapter_readback_review")
		}
		if len(result.BlockedReasons) == 0 {
			result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, "adapter_readback_not_ready")
		}
		result.NextHostAction = firstNextHostAction(result.NextHostAction, "review_adapter_readback")
	}
	if result.AuthorizationBound {
		result.Boundaries = AppendBoundaries(result.Boundaries, "authorization_bound_host_view", "authorized_lifecycle_closeout_view")
	}
	return result.Normalize()
}

func CloneProductionAdapterReadbackHostView(in ProductionAdapterReadbackHostView) ProductionAdapterReadbackHostView {
	out := in
	out.AuthorizationMissingInputs = cloneMissingInputs(in.AuthorizationMissingInputs)
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.Verification = in.Verification.Clone()
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (v ProductionAdapterReadbackHostView) Clone() ProductionAdapterReadbackHostView {
	return CloneProductionAdapterReadbackHostView(v)
}

func (v ProductionAdapterReadbackHostView) Normalize() ProductionAdapterReadbackHostView {
	out := CloneProductionAdapterReadbackHostView(v)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = normalizeControlToken(out.Status)
	out.Mode = normalizeControlToken(out.Mode)
	if out.Mode == "" {
		out.Mode = "production_adapter_readback_host_view"
	}
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	if out.RunnerEffect == "" {
		out.RunnerEffect = "none"
	}
	if out.PromptEffect == "" {
		out.PromptEffect = "none"
	}
	out.ReadbackStatus = NormalizeHostActionStatus(string(out.ReadbackStatus))
	out.HandoffStatus = NormalizeVerificationStatus(string(out.HandoffStatus))
	out.AuthorizationChainStatus = normalizeControlToken(out.AuthorizationChainStatus)
	out.InvocationReportBindingRef = normalizeOneDisplaySafeRef(out.InvocationReportBindingRef)
	out.AuthorizationPacketRef = normalizeOneDisplaySafeRef(out.AuthorizationPacketRef)
	out.PreflightReviewPacketRef = normalizeOneDisplaySafeRef(out.PreflightReviewPacketRef)
	out.AuthorizationMissingInputs = normalizeMissingInputs(out.AuthorizationMissingInputs)
	out.AdapterRef = normalizeOneDisplaySafeRef(out.AdapterRef)
	out.DescriptorRef = normalizeOneDisplaySafeRef(out.DescriptorRef)
	out.InvocationRef = normalizeOneDisplaySafeRef(out.InvocationRef)
	out.IdempotencyRef = normalizeOneDisplaySafeRef(out.IdempotencyRef)
	out.ResultRef = normalizeOneDisplaySafeRef(out.ResultRef)
	out.FailureRef = normalizeOneDisplaySafeRef(out.FailureRef)
	out.ReadbackRef = normalizeOneDisplaySafeRef(out.ReadbackRef)
	out.CompensationRef = normalizeOneDisplaySafeRef(out.CompensationRef)
	out.CompletionHandoffRef = normalizeOneDisplaySafeRef(out.CompletionHandoffRef)
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
	out.Verification = out.Verification.Normalize()
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.BlockedReasons = normalizeControlTokenList(out.BlockedReasons)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	if !out.Available {
		out.Status = "unavailable"
		out.AuthorizationBound = false
		out.AuthorizationChainStatus = "unavailable"
		out.AuthorizationMissingInputs = nil
		out.ReadyForHostReview = false
		out.ReadyForReadbackReview = false
		out.ReadyForFailureReview = false
		out.ReadyForCompletionAudit = false
	}
	if out.Status == "" {
		out.Status = "blocked"
	}
	if out.RawOutputLoaded {
		out.Status = "blocked"
		if out.FailureClass == FailureNone {
			out.FailureClass = FailureEvidenceWeak
		}
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		if out.NextHostAction == "" {
			out.NextHostAction = "provide_display_safe_refs"
		}
	}
	out.ObjectiveSatisfied = false
	out.VerificationSatisfied = false
	out.Verification.Satisfied = false
	out.CoreInvocationExecuted = false
	out.DurableWriteByCore = false
	out.AuthorizationBound = out.AuthorizationBound &&
		out.InvocationReportBindingRef != "" &&
		out.AuthorizationPacketRef != "" &&
		out.PreflightReviewPacketRef != "" &&
		!containsMissingInput(out.AuthorizationMissingInputs, "host:authorization_bound_readback_review")
	if out.AuthorizationChainStatus == "" {
		if out.AuthorizationBound {
			out.AuthorizationChainStatus = "authorization_bound"
		} else {
			out.AuthorizationChainStatus = "authorization_required"
		}
	}
	if !out.AuthorizationBound && out.AuthorizationChainStatus == "authorization_bound" {
		out.AuthorizationChainStatus = "authorization_required"
	}
	out.ReadyForCompletionAudit = out.Status == "ready_for_completion_audit" && out.AuthorizationBound && !containsMissingInput(out.MissingInputs, "host:authorization_bound_readback_review") && !containsMissingInput(out.MissingInputs, "host:display_safe_refs") && !out.RawOutputLoaded
	out.ReadyForReadbackReview = out.ReadyForCompletionAudit
	out.ReadyForFailureReview = out.Status == "ready_for_failure_review" && !out.RawOutputLoaded
	out.ReadyForHostReview = (out.ReadyForReadbackReview || out.ReadyForFailureReview || out.ReadyForCompletionAudit) && out.Available
	return out
}

func productionAdapterBlock(result ProductionAdapterResolution, failure FailureClass, reason string, missing MissingInput, next NextHostAction) ProductionAdapterResolution {
	result.Status = HostActionBlocked
	result.ReadyForHostPreflight = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	return result
}

func productionAdapterPreflightBlock(result ProductionAdapterPreflight, failure FailureClass, reason string, missing MissingInput, next NextHostAction) ProductionAdapterPreflight {
	result.Status = HostActionBlocked
	result.ReadyForHostInvocation = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	return result
}

func hostAdapterInvocationBlock(result HostAdapterInvocationProjection, failure FailureClass, reason string, missing MissingInput, next NextHostAction) HostAdapterInvocationProjection {
	result.Status = HostActionBlocked
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	return result
}

func productionAdapterReadbackReviewBlock(result ProductionAdapterReadbackReview, failure FailureClass, reason string, missing MissingInput, next NextHostAction) ProductionAdapterReadbackReview {
	result.Status = HostActionBlocked
	result.ReadyForReadbackReview = false
	result.ReadyForFailureReview = false
	result.ReadyForCompletionAudit = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	return result
}

func productionAdapterCompletionHandoffBlock(result ProductionAdapterCompletionHandoff, failure FailureClass, reason string, missing MissingInput, next NextHostAction) ProductionAdapterCompletionHandoff {
	result.Status = VerificationBlocked
	result.ReadyForCompletionAudit = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	result.Verification = VerificationResult{
		ContractVersion: ContractVersion,
		Status:          VerificationBlocked,
		FailureClass:    result.FailureClass,
		EvidenceRefs:    cloneEvidenceRefs(result.EvidenceRefs),
		MissingInputs:   cloneMissingInputs(result.MissingInputs),
		Boundaries:      cloneBoundaries(result.Boundaries),
		NextHostAction:  result.NextHostAction,
	}.Normalize()
	return result
}

func applyProductionAdapterDescriptorCompleteness(result ProductionAdapterResolution, descriptor ProductionAdapterDescriptor) ProductionAdapterResolution {
	for _, check := range productionAdapterDescriptorCompletenessChecks(descriptor) {
		if check.missing {
			result = productionAdapterBlock(result, check.failure, check.reason, check.input, "provide_adapter_descriptor")
		}
	}
	return result
}

func applyProductionAdapterCatalogSelectionBinding(result ProductionAdapterResolution, selection ProductionAdapterCatalogSelection, descriptor ProductionAdapterDescriptor) ProductionAdapterResolution {
	selection = selection.Normalize()
	descriptor = descriptor.Normalize()
	if !selection.ReadyForResolution {
		result = productionAdapterBlock(result, firstFailureClass(selection.FailureClass, FailureConfigMissing), "adapter_catalog_selection_not_ready", "host:adapter_catalog_selection", "select_adapter_from_catalog")
		return result
	}
	if result.CatalogSelectionRef == "" {
		result = productionAdapterBlock(result, FailureEvidenceMissing, "adapter_selection_ref_missing", "host:adapter_selection_ref", "provide_adapter_selection")
	}
	for _, check := range []struct {
		actual  DisplaySafeRef
		want    DisplaySafeRef
		reason  string
		missing MissingInput
	}{
		{result.CatalogSelectionRef, selection.SelectionRef, "adapter_selection_ref_mismatch", "host:adapter_catalog_selection"},
		{result.CatalogSnapshotRef, selection.CatalogSnapshotRef, "catalog_snapshot_ref_mismatch", "host:adapter_catalog_snapshot_ref"},
		{result.RequestedAdapterRef, selection.SelectedAdapterRef, "selected_adapter_ref_mismatch", "host:selected_adapter_ref"},
		{result.SelectedSourceRef, selection.SelectedSourceRef, "selected_source_ref_mismatch", "host:selected_source_ref"},
		{result.SelectedCandidateRef, selection.SelectedCandidateRef, "selected_candidate_ref_mismatch", "host:selected_candidate_strategy_ref"},
	} {
		if check.want != "" && check.actual != check.want {
			result = productionAdapterBlock(result, FailureInvalidInput, check.reason, check.missing, "review_adapter_selection")
		}
	}
	if selection.SelectedSourceKind != "" && result.SelectedSourceKind != selection.SelectedSourceKind {
		result = productionAdapterBlock(result, FailureInvalidInput, "selected_source_kind_mismatch", "host:selected_source_kind", "review_adapter_selection")
	}
	if selection.SelectedDescriptor.AdapterRef != "" && descriptor.AdapterRef != "" && descriptor.AdapterRef != selection.SelectedDescriptor.AdapterRef {
		result = productionAdapterBlock(result, FailureInvalidInput, "adapter_descriptor_selection_mismatch", "host:adapter_descriptor", "review_adapter_selection")
	}
	return result
}

type productionAdapterDescriptorCompletenessCheck struct {
	missing bool
	reason  string
	input   MissingInput
	failure FailureClass
}

func productionAdapterDescriptorCompletenessChecks(descriptor ProductionAdapterDescriptor) []productionAdapterDescriptorCompletenessCheck {
	return []productionAdapterDescriptorCompletenessCheck{
		{descriptor.AdapterRef == "", "adapter_descriptor_missing", "host:adapter_ref", FailureHostAdapterMissing},
		{descriptor.Owner == "", "adapter_owner_missing", "host:adapter_owner", FailureConfigMissing},
		{descriptor.OwnerRef == "", "adapter_owner_ref_missing", "host:adapter_owner_ref", FailureConfigMissing},
		{descriptor.Kind == "", "adapter_kind_missing", "host:adapter_kind", FailureInvalidInput},
		{len(descriptor.SupportedSourceKinds) == 0, "adapter_supported_source_kind_missing", "host:adapter_supported_source_kind", FailureConfigMissing},
		{len(descriptor.SupportedCandidateRefs) == 0, "adapter_supported_candidate_ref_missing", "host:adapter_supported_candidate_ref", FailureConfigMissing},
		{descriptor.InputContractRef == "", "adapter_input_contract_ref_missing", "host:adapter_input_contract_ref", FailureConfigMissing},
		{descriptor.OutputContractRef == "", "adapter_output_contract_ref_missing", "host:adapter_output_contract_ref", FailureConfigMissing},
		{descriptor.ReadbackContractRef == "", "adapter_readback_contract_ref_missing", "host:adapter_readback_contract_ref", FailureConfigMissing},
		{len(descriptor.RequiredPolicyRefs) == 0, "adapter_required_policy_ref_missing", "host:adapter_required_policy_ref", FailurePolicyBlocked},
		{len(descriptor.RequiredApprovalRefs) == 0, "adapter_required_approval_ref_missing", "host:adapter_required_approval_ref", FailureApprovalRequired},
		{descriptor.RequiredBudgetRef == "", "adapter_required_budget_ref_missing", "host:adapter_required_budget_ref", FailureBudgetExhausted},
		{descriptor.IdempotencyContractRef == "", "adapter_idempotency_contract_ref_missing", "host:adapter_idempotency_contract_ref", FailureInvalidInput},
		{descriptor.RiskRef == "", "adapter_risk_ref_missing", "host:adapter_risk_ref", FailureConfigMissing},
		{descriptor.SideEffectClass == "", "adapter_side_effect_class_missing", "host:adapter_side_effect_class", FailureConfigMissing},
		{descriptor.TimeoutPolicyRef == "", "adapter_timeout_policy_ref_missing", "host:adapter_timeout_policy_ref", FailureTimeout},
		{descriptor.CompensationHandoffRef == "", "adapter_compensation_handoff_ref_missing", "host:adapter_compensation_handoff_ref", FailureConfigMissing},
		{descriptor.RedactionPolicyRef == "", "adapter_redaction_policy_ref_missing", "host:adapter_redaction_policy_ref", FailureConfigMissing},
		{len(descriptor.PreflightCheckRefs) == 0, "adapter_preflight_check_ref_missing", "host:adapter_preflight_check_ref", FailureConfigMissing},
		{len(descriptor.DisplaySafeInputRefs) == 0, "adapter_display_safe_input_ref_missing", "host:adapter_display_safe_input_ref", FailureEvidenceMissing},
		{len(descriptor.DisplaySafeOutputRefs) == 0, "adapter_display_safe_output_ref_missing", "host:adapter_display_safe_output_ref", FailureEvidenceMissing},
	}
}

func productionAdapterResolutionUnsafe(input ProductionAdapterResolutionInput) bool {
	return displaySafeRefRejected(input.ApplyEnvelopeRef) ||
		displaySafeRefRejected(input.SelectedSourceRef) ||
		displaySafeRefRejected(input.SelectedCandidateRef) ||
		displaySafeRefRejected(input.RequestedAdapterRef) ||
		displaySafeRefRejected(input.CatalogSnapshotRef) ||
		displaySafeRefRejected(input.CatalogSelectionRef) ||
		displaySafeRefRejected(input.HostPolicyRef) ||
		displaySafeRefRejected(input.BudgetRef) ||
		displaySafeRefRejected(input.IdempotencyRef) ||
		displaySafeRefSliceRejected(input.ApprovalContextRefs) ||
		displaySafeRefSliceRejected(input.AvailableCapabilityRefs) ||
		displaySafeRefSliceRejected(input.ConfirmedPolicyRefs) ||
		displaySafeRefSliceRejected(input.ConfirmedApprovalRefs) ||
		productionAdapterCatalogSelectionUnsafe(input.CatalogSelection) ||
		productionAdapterDescriptorUnsafe(input.Descriptor)
}

func productionAdapterDescriptorUnsafe(descriptor ProductionAdapterDescriptor) bool {
	return displaySafeRefRejected(descriptor.AdapterRef) ||
		displaySafeRefRejected(descriptor.OwnerRef) ||
		displaySafeRefRejected(descriptor.InputContractRef) ||
		displaySafeRefRejected(descriptor.OutputContractRef) ||
		displaySafeRefRejected(descriptor.ReadbackContractRef) ||
		displaySafeRefRejected(descriptor.RequiredBudgetRef) ||
		displaySafeRefRejected(descriptor.IdempotencyContractRef) ||
		displaySafeRefRejected(descriptor.RiskRef) ||
		displaySafeRefRejected(descriptor.TimeoutPolicyRef) ||
		displaySafeRefRejected(descriptor.CompensationHandoffRef) ||
		displaySafeRefRejected(descriptor.RedactionPolicyRef) ||
		displaySafeRefSliceRejected(descriptor.SupportedCandidateRefs) ||
		displaySafeRefSliceRejected(descriptor.ProvidesCapabilityRefs) ||
		displaySafeRefSliceRejected(descriptor.RequiresCapabilityRefs) ||
		displaySafeRefSliceRejected(descriptor.RequiredPolicyRefs) ||
		displaySafeRefSliceRejected(descriptor.RequiredApprovalRefs) ||
		displaySafeRefSliceRejected(descriptor.PreflightCheckRefs) ||
		displaySafeRefSliceRejected(descriptor.DisplaySafeInputRefs) ||
		displaySafeRefSliceRejected(descriptor.DisplaySafeOutputRefs) ||
		descriptor.RawOutputLoaded
}

func productionAdapterPreflightUnsafe(input ProductionAdapterPreflightInput) bool {
	return displaySafeRefSliceRejected(input.PreflightResultRefs) || input.RawOutputLoaded
}

func hostAdapterInvocationUnsafe(input HostAdapterInvocationInput) bool {
	return displaySafeRefRejected(input.InvocationRef) ||
		displaySafeRefRejected(input.AdapterRef) ||
		displaySafeRefRejected(input.DescriptorRef) ||
		displaySafeRefRejected(input.IdempotencyRef) ||
		displaySafeRefRejected(input.StartedEventRef) ||
		displaySafeRefRejected(input.CompletedEventRef) ||
		displaySafeRefRejected(input.ResultRef) ||
		displaySafeRefRejected(input.FailureRef) ||
		displaySafeRefRejected(input.ReadbackRef) ||
		displaySafeRefRejected(input.CompensationRef) ||
		displaySafeRefRejected(input.CompletionHandoffRef) ||
		displaySafeRefSliceRejected(input.ApprovalRefs) ||
		input.RawOutputLoaded
}

func productionAdapterInvocationProjectionUnsafe(input HostAdapterInvocationProjection) bool {
	return displaySafeRefRejected(input.AdapterRef) ||
		displaySafeRefRejected(input.DescriptorRef) ||
		displaySafeRefRejected(input.InvocationRef) ||
		displaySafeRefRejected(input.IdempotencyRef) ||
		displaySafeRefRejected(input.StartedEventRef) ||
		displaySafeRefRejected(input.CompletedEventRef) ||
		displaySafeRefRejected(input.ResultRef) ||
		displaySafeRefRejected(input.FailureRef) ||
		displaySafeRefRejected(input.ReadbackRef) ||
		displaySafeRefRejected(input.CompensationRef) ||
		displaySafeRefRejected(input.CompletionHandoffRef) ||
		displaySafeRefSliceRejected(input.ApprovalRefs) ||
		input.RawOutputLoaded
}

func productionAdapterInvocationFromReportBinding(binding ProductionAdapterInvocationReportBinding) HostAdapterInvocationProjection {
	completed := binding.AuthorizationBound && binding.ReadyForReadbackReview && binding.HostInvocationCompleted
	failed := binding.AuthorizationBound && binding.ReadyForFailureReview && binding.HostInvocationFailed
	status := HostActionBlocked
	if completed || failed {
		status = HostActionRecorded
	}
	return HostAdapterInvocationProjection{
		ContractVersion:         ContractVersion,
		Projected:               true,
		Status:                  status,
		HostInvocationReported:  completed || failed,
		HostInvocationCompleted: completed,
		HostInvocationFailed:    failed,
		AdapterRef:              binding.AdapterRef,
		DescriptorRef:           binding.DescriptorRef,
		InvocationRef:           binding.InvocationRef,
		IdempotencyRef:          binding.IdempotencyRef,
		StartedEventRef:         binding.StartedEventRef,
		CompletedEventRef:       binding.CompletedEventRef,
		ResultRef:               binding.ResultRef,
		FailureRef:              binding.FailureRef,
		ReadbackRef:             binding.ReadbackRef,
		CompensationRef:         binding.CompensationRef,
		CompletionHandoffRef:    binding.CompletionHandoffRef,
		FailureClass:            binding.FailureClass,
		Boundaries:              binding.Boundaries,
		NextHostAction:          binding.NextHostAction,
		RunnerEffect:            "none",
		PromptEffect:            "none",
		RawOutputLoaded:         binding.RawOutputLoaded,
	}.Normalize()
}

func productionAdapterInvocationReportBindingOutputUnsafe(binding ProductionAdapterInvocationReportBinding) bool {
	return binding.RawOutputLoaded ||
		displaySafeRefRejected(binding.InvocationReportBindingRef) ||
		displaySafeRefRejected(binding.AuthorizationPacketRef) ||
		displaySafeRefRejected(binding.PreflightReviewPacketRef) ||
		displaySafeRefRejected(binding.AdapterRef) ||
		displaySafeRefRejected(binding.DescriptorRef) ||
		displaySafeRefRejected(binding.ApplyEnvelopeRef) ||
		displaySafeRefRejected(binding.SelectedSourceRef) ||
		displaySafeRefRejected(binding.SelectedCandidateRef) ||
		displaySafeRefRejected(binding.CatalogSnapshotRef) ||
		displaySafeRefRejected(binding.CatalogSelectionRef) ||
		displaySafeRefRejected(binding.InvocationRef) ||
		displaySafeRefRejected(binding.HostConfirmationRef) ||
		displaySafeRefRejected(binding.IdempotencyRef) ||
		displaySafeRefRejected(binding.ExpectedStartedEventRef) ||
		displaySafeRefRejected(binding.StartedEventRef) ||
		displaySafeRefRejected(binding.ExpectedCompletedEventRef) ||
		displaySafeRefRejected(binding.CompletedEventRef) ||
		displaySafeRefRejected(binding.ExpectedResultRef) ||
		displaySafeRefRejected(binding.ResultRef) ||
		displaySafeRefRejected(binding.ExpectedReadbackRef) ||
		displaySafeRefRejected(binding.ReadbackRef) ||
		displaySafeRefRejected(binding.ExpectedFailureRef) ||
		displaySafeRefRejected(binding.FailureRef) ||
		displaySafeRefRejected(binding.ExpectedCompensationRef) ||
		displaySafeRefRejected(binding.CompensationRef) ||
		displaySafeRefRejected(binding.ExpectedCompletionHandoffRef) ||
		displaySafeRefRejected(binding.CompletionHandoffRef) ||
		displaySafeRefSliceRejected(binding.ApprovalRefs) ||
		displaySafeRefSliceRejected(binding.RequiredApprovalRefs)
}

func productionAdapterInvocationReportBindingEmpty(binding ProductionAdapterInvocationReportBinding) bool {
	return !binding.Projected &&
		!binding.Available &&
		binding.Status == "" &&
		binding.Mode == "" &&
		!binding.ReadyForHostDisplay &&
		!binding.ReadyForReadbackReview &&
		!binding.ReadyForFailureReview &&
		!binding.AuthorizationBound &&
		!binding.HostInvocationAuthorized &&
		!binding.HostInvocationReported &&
		!binding.HostInvocationCompleted &&
		!binding.HostInvocationFailed &&
		binding.InvocationReportBindingRef == "" &&
		binding.AuthorizationPacketRef == "" &&
		binding.PreflightReviewPacketRef == "" &&
		binding.AdapterRef == "" &&
		binding.DescriptorRef == "" &&
		binding.InvocationRef == "" &&
		binding.HostConfirmationRef == "" &&
		binding.IdempotencyRef == "" &&
		len(binding.ApprovalRefs) == 0 &&
		len(binding.RequiredApprovalRefs) == 0 &&
		binding.StartedEventRef == "" &&
		binding.CompletedEventRef == "" &&
		binding.ResultRef == "" &&
		binding.ReadbackRef == "" &&
		binding.FailureRef == "" &&
		binding.CompensationRef == "" &&
		binding.CompletionHandoffRef == "" &&
		len(binding.MissingInputs) == 0 &&
		len(binding.BlockedReasons) == 0 &&
		len(binding.Boundaries) == 0 &&
		binding.NextHostAction == "" &&
		!binding.RawOutputLoaded
}

func productionAdapterReadbackReviewUnsafe(input ProductionAdapterReadbackReview) bool {
	return displaySafeRefRejected(input.InvocationReportBindingRef) ||
		displaySafeRefRejected(input.AuthorizationPacketRef) ||
		displaySafeRefRejected(input.PreflightReviewPacketRef) ||
		displaySafeRefRejected(input.AdapterRef) ||
		displaySafeRefRejected(input.DescriptorRef) ||
		displaySafeRefRejected(input.InvocationRef) ||
		displaySafeRefRejected(input.IdempotencyRef) ||
		displaySafeRefRejected(input.ResultRef) ||
		displaySafeRefRejected(input.FailureRef) ||
		displaySafeRefRejected(input.ReadbackRef) ||
		displaySafeRefRejected(input.CompensationRef) ||
		displaySafeRefRejected(input.CompletionHandoffRef) ||
		evidenceRefsRejected(input.EvidenceRefs) ||
		input.RawOutputLoaded
}

func productionAdapterCompletionHandoffUnsafe(input ProductionAdapterCompletionHandoff) bool {
	return displaySafeRefRejected(input.InvocationReportBindingRef) ||
		displaySafeRefRejected(input.AuthorizationPacketRef) ||
		displaySafeRefRejected(input.PreflightReviewPacketRef) ||
		displaySafeRefRejected(input.AdapterRef) ||
		displaySafeRefRejected(input.DescriptorRef) ||
		displaySafeRefRejected(input.InvocationRef) ||
		displaySafeRefRejected(input.ResultRef) ||
		displaySafeRefRejected(input.ReadbackRef) ||
		displaySafeRefRejected(input.CompletionHandoffRef) ||
		evidenceRefsRejected(input.EvidenceRefs) ||
		evidenceRefsRejected(input.Verification.EvidenceRefs) ||
		input.RawOutputLoaded ||
		input.Verification.RawOutputLoaded
}

func productionAdapterReadbackReviewEmpty(input ProductionAdapterReadbackReview) bool {
	return !input.Projected &&
		input.Status == "" &&
		!input.AuthorizationBound &&
		!input.HostInvocationReported &&
		!input.HostInvocationCompleted &&
		!input.HostInvocationFailed &&
		input.InvocationReportBindingRef == "" &&
		input.AuthorizationPacketRef == "" &&
		input.PreflightReviewPacketRef == "" &&
		input.AdapterRef == "" &&
		input.DescriptorRef == "" &&
		input.InvocationRef == "" &&
		input.IdempotencyRef == "" &&
		input.ResultRef == "" &&
		input.FailureRef == "" &&
		input.ReadbackRef == "" &&
		input.CompensationRef == "" &&
		input.CompletionHandoffRef == "" &&
		len(input.EvidenceRefs) == 0 &&
		len(input.MissingInputs) == 0 &&
		len(input.BlockedReasons) == 0 &&
		len(input.Boundaries) == 0 &&
		input.NextHostAction == "" &&
		!input.RawOutputLoaded
}

func unavailableProductionAdapterReadbackHostView() ProductionAdapterReadbackHostView {
	return ProductionAdapterReadbackHostView{
		ContractVersion: ContractVersion,
		Projected:       true,
		Available:       false,
		Status:          "unavailable",
		Mode:            "production_adapter_readback_host_view",
		RunnerEffect:    "none",
		PromptEffect:    "none",
		Boundaries: []Boundary{
			"production_adapter_readback_host_view",
			"host_adapter_result_view_only",
			"display_safe_refs_only",
			"no_runner_dispatch",
			"no_durable_write_by_core",
		},
	}.Normalize()
}

func productionAdapterHostViewEvidenceRefs(review ProductionAdapterReadbackReview, handoff ProductionAdapterCompletionHandoff) []EvidenceRef {
	return normalizeEvidenceRefs(append(cloneEvidenceRefs(review.EvidenceRefs), handoff.EvidenceRefs...))
}

func productionAdapterHostViewAuthorizationBound(review ProductionAdapterReadbackReview, handoff ProductionAdapterCompletionHandoff) bool {
	if review.ReadyForCompletionAudit {
		return review.AuthorizationBound && handoff.AuthorizationBound
	}
	if review.ReadyForFailureReview {
		return review.AuthorizationBound
	}
	return review.AuthorizationBound
}

func productionAdapterAuthorizationChainStatus(review ProductionAdapterReadbackReview, handoff ProductionAdapterCompletionHandoff) string {
	if productionAdapterHostViewAuthorizationBound(review, handoff) {
		return "authorization_bound"
	}
	if review.AuthorizationBound && review.ReadyForCompletionAudit && !handoff.AuthorizationBound {
		return "completion_handoff_authorization_required"
	}
	return "authorization_required"
}

func productionAdapterAuthorizationMissingInputs(review ProductionAdapterReadbackReview, handoff ProductionAdapterCompletionHandoff) []MissingInput {
	var out []MissingInput
	if firstDisplaySafeRef(review.InvocationReportBindingRef, handoff.InvocationReportBindingRef) == "" {
		out = AppendMissingInputs(out, "host:invocation_report_binding")
	}
	if firstDisplaySafeRef(review.AuthorizationPacketRef, handoff.AuthorizationPacketRef) == "" {
		out = AppendMissingInputs(out, "host:invocation_authorization_packet")
	}
	if firstDisplaySafeRef(review.PreflightReviewPacketRef, handoff.PreflightReviewPacketRef) == "" {
		out = AppendMissingInputs(out, "host:adapter_preflight_review_packet")
	}
	if !productionAdapterHostViewAuthorizationBound(review, handoff) {
		out = AppendMissingInputs(out, "host:authorization_bound_readback_review")
	}
	for _, input := range append(cloneMissingInputs(review.MissingInputs), handoff.MissingInputs...) {
		if productionAdapterAuthorizationMissingInput(input) {
			out = AppendMissingInputs(out, input)
		}
	}
	return normalizeMissingInputs(out)
}

func productionAdapterAuthorizationMissingInput(input MissingInput) bool {
	switch input {
	case "host:authorization_bound_readback_review",
		"host:invocation_report_binding",
		"host:invocation_report_binding_ref",
		"host:invocation_authorization_packet",
		"host:invocation_authorization_packet_ref",
		"host:adapter_preflight_review_packet",
		"host:adapter_preflight_review_packet_ref":
		return true
	default:
		return false
	}
}

func productionAdapterHostViewMissingInputs(review ProductionAdapterReadbackReview, handoff ProductionAdapterCompletionHandoff) []MissingInput {
	return normalizeMissingInputs(append(cloneMissingInputs(review.MissingInputs), handoff.MissingInputs...))
}

func productionAdapterHostViewBlockedReasons(review ProductionAdapterReadbackReview, handoff ProductionAdapterCompletionHandoff) []string {
	return normalizeControlTokenList(append(cloneStringSlice(review.BlockedReasons), handoff.BlockedReasons...))
}

func productionAdapterReadbackHostViewBoundaries(groups ...[]Boundary) []Boundary {
	var out []Boundary
	for _, group := range groups {
		out = append(out, group...)
	}
	for _, item := range []Boundary{
		"production_adapter_readback_host_view",
		"host_adapter_result_view_only",
		"readback_projection_only",
		"completion_handoff_projection_only",
		"completion_audit_input_only",
		"objective_not_satisfied_by_adapter",
		"display_safe_refs_only",
		"display_safe_result_refs_only",
		"no_runner_dispatch",
		"no_durable_write_by_core",
	} {
		out = AppendBoundaries(out, item)
	}
	return normalizeBoundaries(out)
}

func productionAdapterReadbackEvidenceRefs(invocation HostAdapterInvocationProjection) []EvidenceRef {
	source := firstDisplaySafeRef(invocation.AdapterRef, invocation.InvocationRef)
	var evidence []EvidenceRef
	if invocation.ResultRef != "" {
		evidence = append(evidence, EvidenceRef{Ref: invocation.ResultRef, Kind: "adapter_result", Strength: EvidenceAdequate, Source: source})
	}
	if invocation.ReadbackRef != "" {
		evidence = append(evidence, EvidenceRef{Ref: invocation.ReadbackRef, Kind: "adapter_readback", Strength: EvidenceAdequate, Source: source})
	}
	if invocation.CompletionHandoffRef != "" {
		evidence = append(evidence, EvidenceRef{Ref: invocation.CompletionHandoffRef, Kind: "completion_handoff", Strength: EvidenceWeak, Source: source})
	}
	if invocation.FailureRef != "" {
		evidence = append(evidence, EvidenceRef{Ref: invocation.FailureRef, Kind: "adapter_failure", Strength: EvidenceAdequate, Source: source})
	}
	if invocation.CompensationRef != "" {
		evidence = append(evidence, EvidenceRef{Ref: invocation.CompensationRef, Kind: "adapter_compensation", Strength: EvidenceWeak, Source: source})
	}
	return normalizeEvidenceRefs(evidence)
}

func productionAdapterCompletionEvidenceRefs(review ProductionAdapterReadbackReview) []EvidenceRef {
	source := firstDisplaySafeRef(review.AdapterRef, review.InvocationRef)
	var evidence []EvidenceRef
	if review.ResultRef != "" {
		evidence = append(evidence, EvidenceRef{Ref: review.ResultRef, Kind: "adapter_result", Strength: EvidenceAdequate, Source: source})
	}
	if review.ReadbackRef != "" {
		evidence = append(evidence, EvidenceRef{Ref: review.ReadbackRef, Kind: "adapter_readback", Strength: EvidenceAdequate, Source: source})
	}
	if review.CompletionHandoffRef != "" {
		evidence = append(evidence, EvidenceRef{Ref: review.CompletionHandoffRef, Kind: "completion_handoff", Strength: EvidenceWeak, Source: source})
	}
	return normalizeEvidenceRefs(append(evidence, review.EvidenceRefs...))
}

func evidenceRefsRejected(values []EvidenceRef) bool {
	for _, value := range values {
		if displaySafeRefRejected(value.Ref) || displaySafeRefRejected(value.Source) {
			return true
		}
	}
	return false
}

func containsMissingInput(values []MissingInput, want MissingInput) bool {
	needle := firstMissingInput(want, "")
	if needle == "" {
		return false
	}
	for _, value := range normalizeMissingInputs(values) {
		if value == needle {
			return true
		}
	}
	return false
}

func firstReplannerSourceKind(values ...ReplannerSourceKind) ReplannerSourceKind {
	for _, value := range values {
		if kind := NormalizeReplannerSourceKind(string(value)); kind != "" {
			return kind
		}
	}
	return ""
}

func cloneReplannerSourceKinds(in []ReplannerSourceKind) []ReplannerSourceKind {
	if len(in) == 0 {
		return nil
	}
	return append([]ReplannerSourceKind(nil), in...)
}

func normalizeReplannerSourceKinds(in []ReplannerSourceKind) []ReplannerSourceKind {
	out := make([]ReplannerSourceKind, 0, len(in))
	seen := map[ReplannerSourceKind]struct{}{}
	for _, value := range in {
		kind := NormalizeReplannerSourceKind(string(value))
		if kind == "" {
			continue
		}
		if _, exists := seen[kind]; exists {
			continue
		}
		seen[kind] = struct{}{}
		out = append(out, kind)
	}
	return out
}

func replannerSourceKindContains(values []ReplannerSourceKind, needle ReplannerSourceKind) bool {
	needle = NormalizeReplannerSourceKind(string(needle))
	if needle == "" {
		return false
	}
	for _, value := range normalizeReplannerSourceKinds(values) {
		if value == needle {
			return true
		}
	}
	return false
}

func normalizeVersionToken(raw string) string {
	value := normalizeControlToken(raw)
	if value != "" {
		return value
	}
	return ""
}
