package controlcontract

type ProductionAdapterPreflightReviewPacketInput struct {
	PreflightReviewPacketRef DisplaySafeRef              `json:"preflight_review_packet_ref,omitempty"`
	Resolution               ProductionAdapterResolution `json:"resolution,omitempty"`
	Preflight                ProductionAdapterPreflight  `json:"preflight,omitempty"`
	RawOutputLoaded          bool                        `json:"raw_output_loaded"`
}

type ProductionAdapterPreflightReviewPacket struct {
	ContractVersion              string              `json:"contract_version,omitempty"`
	Projected                    bool                `json:"projected"`
	Available                    bool                `json:"available"`
	Status                       string              `json:"status,omitempty"`
	Mode                         string              `json:"mode,omitempty"`
	RunnerEffect                 string              `json:"runner_effect,omitempty"`
	PromptEffect                 string              `json:"prompt_effect,omitempty"`
	ReadyForHostDisplay          bool                `json:"ready_for_host_display"`
	ReadyForHostPreflight        bool                `json:"ready_for_host_preflight"`
	PreflightReported            bool                `json:"preflight_reported"`
	ReadyForHostInvocationReview bool                `json:"ready_for_host_invocation_review"`
	CoreInvocationExecuted       bool                `json:"core_invocation_executed"`
	DurableWriteByCore           bool                `json:"durable_write_by_core"`
	PreflightReviewPacketRef     DisplaySafeRef      `json:"preflight_review_packet_ref,omitempty"`
	DisplaySections              []string            `json:"display_sections,omitempty"`
	AdapterRef                   DisplaySafeRef      `json:"adapter_ref,omitempty"`
	DescriptorRef                DisplaySafeRef      `json:"descriptor_ref,omitempty"`
	ApplyEnvelopeRef             DisplaySafeRef      `json:"apply_envelope_ref,omitempty"`
	SelectedSourceKind           ReplannerSourceKind `json:"selected_source_kind,omitempty"`
	SelectedSourceRef            DisplaySafeRef      `json:"selected_source_ref,omitempty"`
	SelectedCandidateRef         DisplaySafeRef      `json:"selected_candidate_ref,omitempty"`
	CatalogSnapshotRef           DisplaySafeRef      `json:"catalog_snapshot_ref,omitempty"`
	CatalogSelectionRef          DisplaySafeRef      `json:"catalog_selection_ref,omitempty"`
	HostPolicyRef                DisplaySafeRef      `json:"host_policy_ref,omitempty"`
	ApprovalContextRefs          []DisplaySafeRef    `json:"approval_context_refs,omitempty"`
	BudgetRef                    DisplaySafeRef      `json:"budget_ref,omitempty"`
	IdempotencyRef               DisplaySafeRef      `json:"idempotency_ref,omitempty"`
	IdempotencyContractRef       DisplaySafeRef      `json:"idempotency_contract_ref,omitempty"`
	RiskRef                      DisplaySafeRef      `json:"risk_ref,omitempty"`
	SideEffectClass              string              `json:"side_effect_class,omitempty"`
	TimeoutPolicyRef             DisplaySafeRef      `json:"timeout_policy_ref,omitempty"`
	CompensationHandoffRef       DisplaySafeRef      `json:"compensation_handoff_ref,omitempty"`
	RedactionPolicyRef           DisplaySafeRef      `json:"redaction_policy_ref,omitempty"`
	RequiredCapabilityRefs       []DisplaySafeRef    `json:"required_capability_refs,omitempty"`
	RequiredPolicyRefs           []DisplaySafeRef    `json:"required_policy_refs,omitempty"`
	RequiredApprovalRefs         []DisplaySafeRef    `json:"required_approval_refs,omitempty"`
	RequiredBudgetRef            DisplaySafeRef      `json:"required_budget_ref,omitempty"`
	RequiredPreflightRefs        []DisplaySafeRef    `json:"required_preflight_refs,omitempty"`
	PreflightRequiredInputs      []MissingInput      `json:"preflight_required_inputs,omitempty"`
	PreflightStatus              HostActionStatus    `json:"preflight_status,omitempty"`
	PreflightResultRefs          []DisplaySafeRef    `json:"preflight_result_refs,omitempty"`
	PreflightMissingInputs       []MissingInput      `json:"preflight_missing_inputs,omitempty"`
	PreflightBlockedReasons      []string            `json:"preflight_blocked_reasons,omitempty"`
	MissingInputs                []MissingInput      `json:"missing_inputs,omitempty"`
	BlockedReasons               []string            `json:"blocked_reasons,omitempty"`
	FailureClass                 FailureClass        `json:"failure_class,omitempty"`
	Boundaries                   []Boundary          `json:"boundaries,omitempty"`
	NextHostAction               NextHostAction      `json:"next_host_action,omitempty"`
	RawOutputLoaded              bool                `json:"raw_output_loaded"`
}

func BuildProductionAdapterPreflightReviewPacket(input ProductionAdapterPreflightReviewPacketInput) ProductionAdapterPreflightReviewPacket {
	if productionAdapterResolutionEmpty(input.Resolution) {
		return unavailableProductionAdapterPreflightReviewPacket()
	}
	resolution := input.Resolution.Normalize()
	preflightProvided := !productionAdapterPreflightEmpty(input.Preflight)
	preflight := ProductionAdapterPreflight{}
	if preflightProvided {
		preflight = input.Preflight.Normalize()
	}
	descriptor := resolution.Descriptor.Normalize()
	result := ProductionAdapterPreflightReviewPacket{
		ContractVersion:          ContractVersion,
		Projected:                true,
		Available:                true,
		Status:                   "blocked",
		Mode:                     "production_adapter_preflight_review_packet",
		RunnerEffect:             "none",
		PromptEffect:             "none",
		PreflightReviewPacketRef: normalizeOneDisplaySafeRef(input.PreflightReviewPacketRef),
		DisplaySections:          productionAdapterPreflightReviewPacketDisplaySections(),
		AdapterRef:               resolution.AdapterRef,
		DescriptorRef:            resolution.DescriptorRef,
		ApplyEnvelopeRef:         resolution.ApplyEnvelopeRef,
		SelectedSourceKind:       resolution.SelectedSourceKind,
		SelectedSourceRef:        resolution.SelectedSourceRef,
		SelectedCandidateRef:     resolution.SelectedCandidateRef,
		CatalogSnapshotRef:       resolution.CatalogSnapshotRef,
		CatalogSelectionRef:      resolution.CatalogSelectionRef,
		HostPolicyRef:            resolution.HostPolicyRef,
		ApprovalContextRefs:      cloneDisplaySafeRefs(resolution.ApprovalContextRefs),
		BudgetRef:                resolution.BudgetRef,
		IdempotencyRef:           resolution.IdempotencyRef,
		IdempotencyContractRef:   descriptor.IdempotencyContractRef,
		RiskRef:                  descriptor.RiskRef,
		SideEffectClass:          descriptor.SideEffectClass,
		TimeoutPolicyRef:         descriptor.TimeoutPolicyRef,
		CompensationHandoffRef:   descriptor.CompensationHandoffRef,
		RedactionPolicyRef:       descriptor.RedactionPolicyRef,
		RequiredCapabilityRefs:   cloneDisplaySafeRefs(resolution.RequiredCapabilityRefs),
		RequiredPolicyRefs:       cloneDisplaySafeRefs(resolution.RequiredPolicyRefs),
		RequiredApprovalRefs:     cloneDisplaySafeRefs(resolution.RequiredApprovalRefs),
		RequiredBudgetRef:        descriptor.RequiredBudgetRef,
		RequiredPreflightRefs:    cloneDisplaySafeRefs(resolution.RequiredPreflightRefs),
		PreflightRequiredInputs:  productionAdapterPreflightReviewRequiredInputs(),
		PreflightReported:        preflightProvided,
		PreflightStatus:          preflight.Status,
		PreflightResultRefs:      cloneDisplaySafeRefs(preflight.PreflightResultRefs),
		PreflightMissingInputs:   cloneMissingInputs(preflight.MissingInputs),
		PreflightBlockedReasons:  cloneStringSlice(preflight.BlockedReasons),
		FailureClass:             FailureNone,
		Boundaries:               productionAdapterPreflightReviewPacketBoundaries(resolution.Boundaries, preflight.Boundaries),
		RawOutputLoaded:          input.RawOutputLoaded || resolution.RawOutputLoaded || preflight.RawOutputLoaded,
	}
	if productionAdapterPreflightReviewPacketUnsafe(input, resolution, preflight, preflightProvided) {
		result = productionAdapterPreflightReviewPacketBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if result.PreflightReviewPacketRef == "" {
		result = productionAdapterPreflightReviewPacketBlock(result, FailureEvidenceMissing, "adapter_preflight_review_packet_ref_missing", "host:adapter_preflight_review_packet_ref", "provide_adapter_preflight_review_packet")
	}
	if !resolution.ReadyForHostPreflight {
		result = productionAdapterPreflightReviewPacketBlock(result, firstFailureClass(resolution.FailureClass, FailureConfigMissing), "adapter_resolution_not_ready", "host:adapter_resolution", firstNextHostAction(resolution.NextHostAction, "provide_adapter_resolution"))
	}
	if result.PreflightReviewPacketRef != "" && resolution.ReadyForHostPreflight && !result.RawOutputLoaded {
		result.ReadyForHostDisplay = true
		result.ReadyForHostPreflight = true
	}
	if !preflightProvided {
		if len(result.BlockedReasons) == 0 && len(result.MissingInputs) == 0 {
			result.Status = "ready_for_adapter_preflight_review"
			result.NextHostAction = "host_may_run_adapter_preflight"
			result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_adapter_preflight_review")
		}
		return result.Normalize()
	}
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, preflight.MissingInputs...)
	for _, reason := range preflight.BlockedReasons {
		result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	}
	if !preflight.ReadyForHostInvocation {
		result = productionAdapterPreflightReviewPacketBlock(result, firstFailureClass(preflight.FailureClass, FailureConfigMissing), "adapter_preflight_not_ready", "host:adapter_preflight", firstNextHostAction(preflight.NextHostAction, "resolve_adapter_preflight"))
	}
	if len(preflight.PreflightResultRefs) == 0 {
		result = productionAdapterPreflightReviewPacketBlock(result, FailureEvidenceMissing, "preflight_result_ref_missing", "host:adapter_preflight_result_ref", "provide_adapter_preflight_result")
	}
	for _, check := range productionAdapterPreflightReviewPacketConsistencyChecks(resolution, preflight) {
		if check.mismatch {
			result = productionAdapterPreflightReviewPacketBlock(result, FailureInvalidInput, check.reason, check.input, "review_adapter_preflight")
		}
	}
	if len(result.BlockedReasons) == 0 && len(result.MissingInputs) == 0 {
		result.Status = "ready_for_adapter_invocation_review"
		result.ReadyForHostDisplay = true
		result.ReadyForHostPreflight = true
		result.ReadyForHostInvocationReview = true
		result.NextHostAction = "review_adapter_invocation"
		result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_adapter_invocation_review")
	}
	return result.Normalize()
}

func CloneProductionAdapterPreflightReviewPacket(in ProductionAdapterPreflightReviewPacket) ProductionAdapterPreflightReviewPacket {
	out := in
	out.DisplaySections = cloneStringSlice(in.DisplaySections)
	out.ApprovalContextRefs = cloneDisplaySafeRefs(in.ApprovalContextRefs)
	out.RequiredCapabilityRefs = cloneDisplaySafeRefs(in.RequiredCapabilityRefs)
	out.RequiredPolicyRefs = cloneDisplaySafeRefs(in.RequiredPolicyRefs)
	out.RequiredApprovalRefs = cloneDisplaySafeRefs(in.RequiredApprovalRefs)
	out.RequiredPreflightRefs = cloneDisplaySafeRefs(in.RequiredPreflightRefs)
	out.PreflightRequiredInputs = cloneMissingInputs(in.PreflightRequiredInputs)
	out.PreflightResultRefs = cloneDisplaySafeRefs(in.PreflightResultRefs)
	out.PreflightMissingInputs = cloneMissingInputs(in.PreflightMissingInputs)
	out.PreflightBlockedReasons = cloneStringSlice(in.PreflightBlockedReasons)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (p ProductionAdapterPreflightReviewPacket) Clone() ProductionAdapterPreflightReviewPacket {
	return CloneProductionAdapterPreflightReviewPacket(p)
}

func (p ProductionAdapterPreflightReviewPacket) Normalize() ProductionAdapterPreflightReviewPacket {
	out := CloneProductionAdapterPreflightReviewPacket(p)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = normalizeControlToken(out.Status)
	out.Mode = normalizeControlToken(out.Mode)
	if out.Mode == "" {
		out.Mode = "production_adapter_preflight_review_packet"
	}
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	if out.RunnerEffect == "" {
		out.RunnerEffect = "none"
	}
	if out.PromptEffect == "" {
		out.PromptEffect = "none"
	}
	out.CoreInvocationExecuted = false
	out.DurableWriteByCore = false
	out.PreflightReviewPacketRef = normalizeOneDisplaySafeRef(out.PreflightReviewPacketRef)
	out.DisplaySections = normalizeControlTokenList(out.DisplaySections)
	out.AdapterRef = normalizeOneDisplaySafeRef(out.AdapterRef)
	out.DescriptorRef = normalizeOneDisplaySafeRef(out.DescriptorRef)
	out.ApplyEnvelopeRef = normalizeOneDisplaySafeRef(out.ApplyEnvelopeRef)
	out.SelectedSourceKind = NormalizeReplannerSourceKind(string(out.SelectedSourceKind))
	out.SelectedSourceRef = normalizeOneDisplaySafeRef(out.SelectedSourceRef)
	out.SelectedCandidateRef = normalizeOneDisplaySafeRef(out.SelectedCandidateRef)
	out.CatalogSnapshotRef = normalizeOneDisplaySafeRef(out.CatalogSnapshotRef)
	out.CatalogSelectionRef = normalizeOneDisplaySafeRef(out.CatalogSelectionRef)
	out.HostPolicyRef = normalizeOneDisplaySafeRef(out.HostPolicyRef)
	out.ApprovalContextRefs = normalizeDisplaySafeRefs(out.ApprovalContextRefs)
	out.BudgetRef = normalizeOneDisplaySafeRef(out.BudgetRef)
	out.IdempotencyRef = normalizeOneDisplaySafeRef(out.IdempotencyRef)
	out.IdempotencyContractRef = normalizeOneDisplaySafeRef(out.IdempotencyContractRef)
	out.RiskRef = normalizeOneDisplaySafeRef(out.RiskRef)
	out.SideEffectClass = normalizeControlToken(out.SideEffectClass)
	out.TimeoutPolicyRef = normalizeOneDisplaySafeRef(out.TimeoutPolicyRef)
	out.CompensationHandoffRef = normalizeOneDisplaySafeRef(out.CompensationHandoffRef)
	out.RedactionPolicyRef = normalizeOneDisplaySafeRef(out.RedactionPolicyRef)
	out.RequiredCapabilityRefs = normalizeDisplaySafeRefs(out.RequiredCapabilityRefs)
	out.RequiredPolicyRefs = normalizeDisplaySafeRefs(out.RequiredPolicyRefs)
	out.RequiredApprovalRefs = normalizeDisplaySafeRefs(out.RequiredApprovalRefs)
	out.RequiredBudgetRef = normalizeOneDisplaySafeRef(out.RequiredBudgetRef)
	out.RequiredPreflightRefs = normalizeDisplaySafeRefs(out.RequiredPreflightRefs)
	out.PreflightRequiredInputs = normalizeMissingInputs(out.PreflightRequiredInputs)
	out.PreflightStatus = NormalizeHostActionStatus(string(out.PreflightStatus))
	out.PreflightResultRefs = normalizeDisplaySafeRefs(out.PreflightResultRefs)
	out.PreflightMissingInputs = normalizeMissingInputs(out.PreflightMissingInputs)
	out.PreflightBlockedReasons = normalizeControlTokenList(out.PreflightBlockedReasons)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.BlockedReasons = normalizeControlTokenList(out.BlockedReasons)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	if !out.Available {
		out.Status = "unavailable"
		out.ReadyForHostDisplay = false
		out.ReadyForHostPreflight = false
		out.ReadyForHostInvocationReview = false
	}
	if out.Status == "" {
		out.Status = "blocked"
	}
	if out.RawOutputLoaded {
		out.Status = "blocked"
		out.ReadyForHostDisplay = false
		out.ReadyForHostPreflight = false
		out.ReadyForHostInvocationReview = false
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
	out.ReadyForHostDisplay = out.ReadyForHostDisplay &&
		out.Available &&
		out.PreflightReviewPacketRef != "" &&
		out.AdapterRef != "" &&
		out.DescriptorRef != "" &&
		!out.RawOutputLoaded
	out.ReadyForHostPreflight = out.ReadyForHostPreflight &&
		out.ReadyForHostDisplay &&
		(out.Status == "ready_for_adapter_preflight_review" || out.Status == "ready_for_adapter_invocation_review")
	out.ReadyForHostInvocationReview = out.ReadyForHostInvocationReview &&
		out.ReadyForHostDisplay &&
		out.PreflightReported &&
		out.Status == "ready_for_adapter_invocation_review" &&
		len(out.PreflightResultRefs) > 0 &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.CoreInvocationExecuted &&
		!out.DurableWriteByCore
	return out
}

func productionAdapterPreflightReviewPacketBlock(result ProductionAdapterPreflightReviewPacket, failure FailureClass, reason string, missing MissingInput, next NextHostAction) ProductionAdapterPreflightReviewPacket {
	result.Status = "blocked"
	result.ReadyForHostInvocationReview = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	return result
}

type productionAdapterPreflightReviewPacketConsistencyCheck struct {
	mismatch bool
	reason   string
	input    MissingInput
}

func productionAdapterPreflightReviewPacketConsistencyChecks(resolution ProductionAdapterResolution, preflight ProductionAdapterPreflight) []productionAdapterPreflightReviewPacketConsistencyCheck {
	return []productionAdapterPreflightReviewPacketConsistencyCheck{
		{resolution.AdapterRef != "" && preflight.AdapterRef != "" && resolution.AdapterRef != preflight.AdapterRef, "preflight_review_adapter_ref_mismatch", "host:adapter_preflight"},
		{resolution.DescriptorRef != "" && preflight.DescriptorRef != "" && resolution.DescriptorRef != preflight.DescriptorRef, "preflight_review_descriptor_ref_mismatch", "host:adapter_preflight"},
		{resolution.IdempotencyRef != "" && preflight.IdempotencyRef != "" && resolution.IdempotencyRef != preflight.IdempotencyRef, "preflight_review_idempotency_ref_mismatch", "host:adapter_preflight"},
		{resolution.BudgetRef != "" && preflight.BudgetRef != "" && resolution.BudgetRef != preflight.BudgetRef, "preflight_review_budget_ref_mismatch", "host:adapter_preflight"},
		{len(preflight.PolicyRefs) > 0 && !displaySafeRefsContainAll(append(cloneDisplaySafeRefs(resolution.RequiredPolicyRefs), resolution.HostPolicyRef), preflight.PolicyRefs), "preflight_review_policy_ref_mismatch", "host:adapter_preflight"},
		{len(preflight.ApprovalRefs) > 0 && !displaySafeRefsContainAll(append(cloneDisplaySafeRefs(resolution.RequiredApprovalRefs), resolution.ApprovalContextRefs...), preflight.ApprovalRefs), "preflight_review_approval_ref_mismatch", "host:adapter_preflight"},
	}
}

func productionAdapterPreflightReviewPacketUnsafe(input ProductionAdapterPreflightReviewPacketInput, resolution ProductionAdapterResolution, preflight ProductionAdapterPreflight, preflightProvided bool) bool {
	return input.RawOutputLoaded ||
		displaySafeRefRejected(input.PreflightReviewPacketRef) ||
		productionAdapterResolutionOutputUnsafe(resolution) ||
		(preflightProvided && productionAdapterPreflightOutputUnsafe(input.Preflight)) ||
		(preflightProvided && productionAdapterPreflightOutputUnsafe(preflight))
}

func productionAdapterPreflightOutputUnsafe(preflight ProductionAdapterPreflight) bool {
	return preflight.RawOutputLoaded ||
		displaySafeRefRejected(preflight.AdapterRef) ||
		displaySafeRefRejected(preflight.DescriptorRef) ||
		displaySafeRefRejected(preflight.IdempotencyRef) ||
		displaySafeRefSliceRejected(preflight.ApprovalRefs) ||
		displaySafeRefSliceRejected(preflight.PolicyRefs) ||
		displaySafeRefRejected(preflight.BudgetRef) ||
		displaySafeRefSliceRejected(preflight.PreflightResultRefs)
}

func productionAdapterResolutionEmpty(resolution ProductionAdapterResolution) bool {
	return !resolution.Projected &&
		resolution.Status == "" &&
		!resolution.ReadyForHostPreflight &&
		resolution.AdapterRef == "" &&
		resolution.DescriptorRef == "" &&
		resolution.ApplyEnvelopeRef == "" &&
		resolution.SelectedSourceKind == "" &&
		resolution.SelectedSourceRef == "" &&
		resolution.SelectedCandidateRef == "" &&
		resolution.RequestedAdapterRef == "" &&
		resolution.CatalogSnapshotRef == "" &&
		resolution.CatalogSelectionRef == "" &&
		resolution.HostPolicyRef == "" &&
		len(resolution.ApprovalContextRefs) == 0 &&
		resolution.BudgetRef == "" &&
		resolution.IdempotencyRef == "" &&
		len(resolution.MissingInputs) == 0 &&
		len(resolution.BlockedReasons) == 0 &&
		len(resolution.Boundaries) == 0 &&
		resolution.NextHostAction == "" &&
		!resolution.RawOutputLoaded
}

func productionAdapterPreflightEmpty(preflight ProductionAdapterPreflight) bool {
	return !preflight.Projected &&
		preflight.Status == "" &&
		!preflight.ReadyForHostInvocation &&
		preflight.AdapterRef == "" &&
		preflight.DescriptorRef == "" &&
		preflight.IdempotencyRef == "" &&
		len(preflight.ApprovalRefs) == 0 &&
		len(preflight.PolicyRefs) == 0 &&
		preflight.BudgetRef == "" &&
		len(preflight.PreflightResultRefs) == 0 &&
		len(preflight.MissingInputs) == 0 &&
		len(preflight.BlockedReasons) == 0 &&
		preflight.FailureClass == "" &&
		len(preflight.Boundaries) == 0 &&
		preflight.NextHostAction == "" &&
		!preflight.RawOutputLoaded
}

func productionAdapterPreflightReviewRequiredInputs() []MissingInput {
	return []MissingInput{
		"host:adapter_available",
		"host:adapter_version",
		"host:capability_refs",
		"host:credential_ref",
		"host:authorization_ref",
		"host:service_ref",
		"host:policy_ref",
		"host:approval_ref",
		"host:budget_ref",
		"host:idempotency_ref",
		"host:timeout_policy_ref",
		"host:compensation_handoff_ref",
		"host:adapter_preflight_result_ref",
	}
}

func productionAdapterPreflightReviewPacketDisplaySections() []string {
	return []string{
		"adapter_preflight_provenance",
		"adapter_preflight_requirements",
		"adapter_preflight_host_checks",
		"adapter_invocation_review_handoff",
	}
}

func productionAdapterPreflightReviewPacketBoundaries(groups ...[]Boundary) []Boundary {
	return MergeBoundaries(
		[]Boundary{
			"production_adapter_preflight_review_packet",
			"preflight_review_packet_projection_only",
			"host_owned_adapter_preflight_review",
			"host_cli_preflight_review",
			"display_safe_refs_only",
			"no_adapter_invocation",
			"no_runner_dispatch",
		},
		MergeBoundaries(groups...),
	)
}

func unavailableProductionAdapterPreflightReviewPacket() ProductionAdapterPreflightReviewPacket {
	return ProductionAdapterPreflightReviewPacket{
		ContractVersion: ContractVersion,
		Projected:       true,
		Available:       false,
		Status:          "unavailable",
		Mode:            "production_adapter_preflight_review_packet",
		RunnerEffect:    "none",
		PromptEffect:    "none",
		DisplaySections: productionAdapterPreflightReviewPacketDisplaySections(),
		Boundaries: []Boundary{
			"production_adapter_preflight_review_packet",
			"preflight_review_packet_projection_only",
			"host_cli_preflight_review",
			"display_safe_refs_only",
			"no_adapter_invocation",
			"no_runner_dispatch",
		},
		NextHostAction: "provide_adapter_resolution",
	}
}
