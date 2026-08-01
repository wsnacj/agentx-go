package controlcontract

type ProductionAdapterInvocationAuthorizationPacketInput struct {
	AuthorizationPacketRef       DisplaySafeRef                         `json:"authorization_packet_ref,omitempty"`
	PreflightReviewPacket        ProductionAdapterPreflightReviewPacket `json:"preflight_review_packet,omitempty"`
	InvocationRef                DisplaySafeRef                         `json:"invocation_ref,omitempty"`
	HostConfirmationRef          DisplaySafeRef                         `json:"host_confirmation_ref,omitempty"`
	ApprovalRefs                 []DisplaySafeRef                       `json:"approval_refs,omitempty"`
	IdempotencyRef               DisplaySafeRef                         `json:"idempotency_ref,omitempty"`
	ExpectedStartedEventRef      DisplaySafeRef                         `json:"expected_started_event_ref,omitempty"`
	ExpectedCompletedEventRef    DisplaySafeRef                         `json:"expected_completed_event_ref,omitempty"`
	ExpectedResultRef            DisplaySafeRef                         `json:"expected_result_ref,omitempty"`
	ExpectedReadbackRef          DisplaySafeRef                         `json:"expected_readback_ref,omitempty"`
	ExpectedFailureRef           DisplaySafeRef                         `json:"expected_failure_ref,omitempty"`
	ExpectedCompensationRef      DisplaySafeRef                         `json:"expected_compensation_ref,omitempty"`
	ExpectedCompletionHandoffRef DisplaySafeRef                         `json:"expected_completion_handoff_ref,omitempty"`
	TimeoutPolicyRef             DisplaySafeRef                         `json:"timeout_policy_ref,omitempty"`
	CancellationPolicyRef        DisplaySafeRef                         `json:"cancellation_policy_ref,omitempty"`
	HostPolicyRef                DisplaySafeRef                         `json:"host_policy_ref,omitempty"`
	BudgetRef                    DisplaySafeRef                         `json:"budget_ref,omitempty"`
	RawOutputLoaded              bool                                   `json:"raw_output_loaded"`
}

type ProductionAdapterInvocationAuthorizationPacket struct {
	ContractVersion              string              `json:"contract_version,omitempty"`
	Projected                    bool                `json:"projected"`
	Available                    bool                `json:"available"`
	Status                       string              `json:"status,omitempty"`
	Mode                         string              `json:"mode,omitempty"`
	RunnerEffect                 string              `json:"runner_effect,omitempty"`
	PromptEffect                 string              `json:"prompt_effect,omitempty"`
	ReadyForHostDisplay          bool                `json:"ready_for_host_display"`
	ReadyForHostInvocation       bool                `json:"ready_for_host_invocation"`
	ReadyForHostAuthorization    bool                `json:"ready_for_host_authorization"`
	HostInvocationAuthorized     bool                `json:"host_invocation_authorized"`
	CoreInvocationExecuted       bool                `json:"core_invocation_executed"`
	DurableWriteByCore           bool                `json:"durable_write_by_core"`
	AuthorizationPacketRef       DisplaySafeRef      `json:"authorization_packet_ref,omitempty"`
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
	InvocationRef                DisplaySafeRef      `json:"invocation_ref,omitempty"`
	HostConfirmationRef          DisplaySafeRef      `json:"host_confirmation_ref,omitempty"`
	HostPolicyRef                DisplaySafeRef      `json:"host_policy_ref,omitempty"`
	ApprovalContextRefs          []DisplaySafeRef    `json:"approval_context_refs,omitempty"`
	ApprovalRefs                 []DisplaySafeRef    `json:"approval_refs,omitempty"`
	RequiredApprovalRefs         []DisplaySafeRef    `json:"required_approval_refs,omitempty"`
	BudgetRef                    DisplaySafeRef      `json:"budget_ref,omitempty"`
	IdempotencyRef               DisplaySafeRef      `json:"idempotency_ref,omitempty"`
	IdempotencyContractRef       DisplaySafeRef      `json:"idempotency_contract_ref,omitempty"`
	RiskRef                      DisplaySafeRef      `json:"risk_ref,omitempty"`
	SideEffectClass              string              `json:"side_effect_class,omitempty"`
	TimeoutPolicyRef             DisplaySafeRef      `json:"timeout_policy_ref,omitempty"`
	CancellationPolicyRef        DisplaySafeRef      `json:"cancellation_policy_ref,omitempty"`
	CompensationHandoffRef       DisplaySafeRef      `json:"compensation_handoff_ref,omitempty"`
	RedactionPolicyRef           DisplaySafeRef      `json:"redaction_policy_ref,omitempty"`
	ExpectedStartedEventRef      DisplaySafeRef      `json:"expected_started_event_ref,omitempty"`
	ExpectedCompletedEventRef    DisplaySafeRef      `json:"expected_completed_event_ref,omitempty"`
	ExpectedResultRef            DisplaySafeRef      `json:"expected_result_ref,omitempty"`
	ExpectedReadbackRef          DisplaySafeRef      `json:"expected_readback_ref,omitempty"`
	ExpectedFailureRef           DisplaySafeRef      `json:"expected_failure_ref,omitempty"`
	ExpectedCompensationRef      DisplaySafeRef      `json:"expected_compensation_ref,omitempty"`
	ExpectedCompletionHandoffRef DisplaySafeRef      `json:"expected_completion_handoff_ref,omitempty"`
	PreflightResultRefs          []DisplaySafeRef    `json:"preflight_result_refs,omitempty"`
	AuthorizationRequiredInputs  []MissingInput      `json:"authorization_required_inputs,omitempty"`
	MissingInputs                []MissingInput      `json:"missing_inputs,omitempty"`
	BlockedReasons               []string            `json:"blocked_reasons,omitempty"`
	FailureClass                 FailureClass        `json:"failure_class,omitempty"`
	Boundaries                   []Boundary          `json:"boundaries,omitempty"`
	NextHostAction               NextHostAction      `json:"next_host_action,omitempty"`
	RawOutputLoaded              bool                `json:"raw_output_loaded"`
}

func BuildProductionAdapterInvocationAuthorizationPacket(input ProductionAdapterInvocationAuthorizationPacketInput) ProductionAdapterInvocationAuthorizationPacket {
	if productionAdapterPreflightReviewPacketEmpty(input.PreflightReviewPacket) {
		return unavailableProductionAdapterInvocationAuthorizationPacket()
	}
	rawReview := input.PreflightReviewPacket
	review := rawReview.Normalize()
	result := ProductionAdapterInvocationAuthorizationPacket{
		ContractVersion:              ContractVersion,
		Projected:                    true,
		Available:                    true,
		Status:                       "blocked",
		Mode:                         "production_adapter_invocation_authorization_packet",
		RunnerEffect:                 "none",
		PromptEffect:                 "none",
		AuthorizationPacketRef:       normalizeOneDisplaySafeRef(input.AuthorizationPacketRef),
		PreflightReviewPacketRef:     review.PreflightReviewPacketRef,
		DisplaySections:              productionAdapterInvocationAuthorizationPacketDisplaySections(),
		AdapterRef:                   review.AdapterRef,
		DescriptorRef:                review.DescriptorRef,
		ApplyEnvelopeRef:             review.ApplyEnvelopeRef,
		SelectedSourceKind:           review.SelectedSourceKind,
		SelectedSourceRef:            review.SelectedSourceRef,
		SelectedCandidateRef:         review.SelectedCandidateRef,
		CatalogSnapshotRef:           review.CatalogSnapshotRef,
		CatalogSelectionRef:          review.CatalogSelectionRef,
		InvocationRef:                normalizeOneDisplaySafeRef(input.InvocationRef),
		HostConfirmationRef:          normalizeOneDisplaySafeRef(input.HostConfirmationRef),
		HostPolicyRef:                normalizeOneDisplaySafeRef(input.HostPolicyRef),
		ApprovalContextRefs:          cloneDisplaySafeRefs(review.ApprovalContextRefs),
		ApprovalRefs:                 normalizeDisplaySafeRefs(input.ApprovalRefs),
		RequiredApprovalRefs:         cloneDisplaySafeRefs(review.RequiredApprovalRefs),
		BudgetRef:                    normalizeOneDisplaySafeRef(input.BudgetRef),
		IdempotencyRef:               normalizeOneDisplaySafeRef(input.IdempotencyRef),
		IdempotencyContractRef:       review.IdempotencyContractRef,
		RiskRef:                      review.RiskRef,
		SideEffectClass:              review.SideEffectClass,
		TimeoutPolicyRef:             normalizeOneDisplaySafeRef(input.TimeoutPolicyRef),
		CancellationPolicyRef:        normalizeOneDisplaySafeRef(input.CancellationPolicyRef),
		CompensationHandoffRef:       review.CompensationHandoffRef,
		RedactionPolicyRef:           review.RedactionPolicyRef,
		ExpectedStartedEventRef:      normalizeOneDisplaySafeRef(input.ExpectedStartedEventRef),
		ExpectedCompletedEventRef:    normalizeOneDisplaySafeRef(input.ExpectedCompletedEventRef),
		ExpectedResultRef:            normalizeOneDisplaySafeRef(input.ExpectedResultRef),
		ExpectedReadbackRef:          normalizeOneDisplaySafeRef(input.ExpectedReadbackRef),
		ExpectedFailureRef:           normalizeOneDisplaySafeRef(input.ExpectedFailureRef),
		ExpectedCompensationRef:      normalizeOneDisplaySafeRef(input.ExpectedCompensationRef),
		ExpectedCompletionHandoffRef: normalizeOneDisplaySafeRef(input.ExpectedCompletionHandoffRef),
		PreflightResultRefs:          cloneDisplaySafeRefs(review.PreflightResultRefs),
		AuthorizationRequiredInputs:  productionAdapterInvocationAuthorizationRequiredInputs(),
		FailureClass:                 FailureNone,
		Boundaries:                   productionAdapterInvocationAuthorizationPacketBoundaries(review.Boundaries),
		RawOutputLoaded:              input.RawOutputLoaded || review.RawOutputLoaded,
	}
	if productionAdapterInvocationAuthorizationPacketUnsafe(input) || productionAdapterPreflightReviewPacketOutputUnsafe(rawReview) {
		result = productionAdapterInvocationAuthorizationPacketBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if !review.ReadyForHostInvocationReview {
		result = productionAdapterInvocationAuthorizationPacketBlock(result, firstFailureClass(review.FailureClass, FailureConfigMissing), "adapter_preflight_review_not_ready", "host:adapter_preflight_review_packet", firstNextHostAction(review.NextHostAction, "review_adapter_preflight"))
	}
	if result.AuthorizationPacketRef != "" && review.ReadyForHostInvocationReview && !result.RawOutputLoaded {
		result.ReadyForHostDisplay = true
	}
	if result.AuthorizationPacketRef == "" {
		result = productionAdapterInvocationAuthorizationPacketBlock(result, FailureEvidenceMissing, "invocation_authorization_packet_ref_missing", "host:invocation_authorization_packet_ref", "provide_invocation_authorization_packet")
	}
	if result.HostConfirmationRef == "" {
		result = productionAdapterInvocationAuthorizationPacketBlock(result, FailureAuthorizationMissing, "host_confirmation_ref_missing", "host:host_confirmation_ref", "request_host_invocation_confirmation")
	}
	if result.InvocationRef == "" {
		result = productionAdapterInvocationAuthorizationPacketBlock(result, FailureEvidenceMissing, "invocation_ref_missing", "host:invocation_ref", "provide_invocation_ref")
	}
	if result.IdempotencyRef == "" {
		result = productionAdapterInvocationAuthorizationPacketBlock(result, FailureInvalidInput, "idempotency_missing", "host:idempotency_ref", "provide_idempotency_ref")
	}
	if input.IdempotencyRef != "" && review.IdempotencyRef != "" && input.IdempotencyRef != review.IdempotencyRef {
		result = productionAdapterInvocationAuthorizationPacketBlock(result, FailureInvalidInput, "authorization_idempotency_ref_mismatch", "host:idempotency_ref", "review_adapter_invocation")
	}
	if !displaySafeRefsContainAll(input.ApprovalRefs, productionAdapterInvocationAuthorizationApprovalRefs(review)) {
		result = productionAdapterInvocationAuthorizationPacketBlock(result, FailureApprovalRequired, "approval_ref_missing", "host:approval_ref", "request_host_approval")
	}
	if result.HostPolicyRef == "" {
		result = productionAdapterInvocationAuthorizationPacketBlock(result, FailurePolicyBlocked, "policy_ref_missing", "host:policy_ref", "request_host_policy_or_budget_review")
	}
	if input.HostPolicyRef != "" && review.HostPolicyRef != "" && input.HostPolicyRef != review.HostPolicyRef {
		result = productionAdapterInvocationAuthorizationPacketBlock(result, FailurePolicyBlocked, "authorization_policy_ref_mismatch", "host:policy_ref", "request_host_policy_or_budget_review")
	}
	if result.BudgetRef == "" {
		result = productionAdapterInvocationAuthorizationPacketBlock(result, FailureBudgetExhausted, "budget_ref_missing", "host:budget_ref", "request_host_policy_or_budget_review")
	}
	if input.BudgetRef != "" && review.BudgetRef != "" && input.BudgetRef != review.BudgetRef {
		result = productionAdapterInvocationAuthorizationPacketBlock(result, FailurePolicyBlocked, "authorization_budget_ref_mismatch", "host:budget_ref", "request_host_policy_or_budget_review")
	}
	if result.TimeoutPolicyRef == "" {
		result = productionAdapterInvocationAuthorizationPacketBlock(result, FailureTimeout, "timeout_policy_missing", "host:timeout_policy_ref", "provide_timeout_policy_ref")
	}
	if input.TimeoutPolicyRef != "" && review.TimeoutPolicyRef != "" && input.TimeoutPolicyRef != review.TimeoutPolicyRef {
		result = productionAdapterInvocationAuthorizationPacketBlock(result, FailureTimeout, "authorization_timeout_policy_ref_mismatch", "host:timeout_policy_ref", "review_adapter_invocation")
	}
	if result.CancellationPolicyRef == "" {
		result = productionAdapterInvocationAuthorizationPacketBlock(result, FailureConfigMissing, "cancellation_policy_missing", "host:cancellation_policy_ref", "provide_cancellation_policy_ref")
	}
	for _, check := range productionAdapterInvocationAuthorizationExpectedRefChecks(result) {
		if check.ref == "" {
			result = productionAdapterInvocationAuthorizationPacketBlock(result, FailureEvidenceMissing, check.reason, check.missing, check.next)
		}
	}
	if len(result.BlockedReasons) == 0 && len(result.MissingInputs) == 0 {
		result.Status = "ready_for_host_adapter_invocation_authorization"
		result.ReadyForHostDisplay = true
		result.ReadyForHostAuthorization = true
		result.ReadyForHostInvocation = true
		result.HostInvocationAuthorized = true
		result.NextHostAction = "host_may_invoke_adapter"
		result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_host_adapter_invocation_authorization")
	}
	return result.Normalize()
}

func CloneProductionAdapterInvocationAuthorizationPacket(in ProductionAdapterInvocationAuthorizationPacket) ProductionAdapterInvocationAuthorizationPacket {
	out := in
	out.DisplaySections = cloneStringSlice(in.DisplaySections)
	out.ApprovalContextRefs = cloneDisplaySafeRefs(in.ApprovalContextRefs)
	out.ApprovalRefs = cloneDisplaySafeRefs(in.ApprovalRefs)
	out.RequiredApprovalRefs = cloneDisplaySafeRefs(in.RequiredApprovalRefs)
	out.PreflightResultRefs = cloneDisplaySafeRefs(in.PreflightResultRefs)
	out.AuthorizationRequiredInputs = cloneMissingInputs(in.AuthorizationRequiredInputs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (p ProductionAdapterInvocationAuthorizationPacket) Clone() ProductionAdapterInvocationAuthorizationPacket {
	return CloneProductionAdapterInvocationAuthorizationPacket(p)
}

func (p ProductionAdapterInvocationAuthorizationPacket) Normalize() ProductionAdapterInvocationAuthorizationPacket {
	out := CloneProductionAdapterInvocationAuthorizationPacket(p)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = normalizeControlToken(out.Status)
	out.Mode = normalizeControlToken(out.Mode)
	if out.Mode == "" {
		out.Mode = "production_adapter_invocation_authorization_packet"
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
	out.AuthorizationPacketRef = normalizeOneDisplaySafeRef(out.AuthorizationPacketRef)
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
	out.InvocationRef = normalizeOneDisplaySafeRef(out.InvocationRef)
	out.HostConfirmationRef = normalizeOneDisplaySafeRef(out.HostConfirmationRef)
	out.HostPolicyRef = normalizeOneDisplaySafeRef(out.HostPolicyRef)
	out.ApprovalContextRefs = normalizeDisplaySafeRefs(out.ApprovalContextRefs)
	out.ApprovalRefs = normalizeDisplaySafeRefs(out.ApprovalRefs)
	out.RequiredApprovalRefs = normalizeDisplaySafeRefs(out.RequiredApprovalRefs)
	out.BudgetRef = normalizeOneDisplaySafeRef(out.BudgetRef)
	out.IdempotencyRef = normalizeOneDisplaySafeRef(out.IdempotencyRef)
	out.IdempotencyContractRef = normalizeOneDisplaySafeRef(out.IdempotencyContractRef)
	out.RiskRef = normalizeOneDisplaySafeRef(out.RiskRef)
	out.SideEffectClass = normalizeControlToken(out.SideEffectClass)
	out.TimeoutPolicyRef = normalizeOneDisplaySafeRef(out.TimeoutPolicyRef)
	out.CancellationPolicyRef = normalizeOneDisplaySafeRef(out.CancellationPolicyRef)
	out.CompensationHandoffRef = normalizeOneDisplaySafeRef(out.CompensationHandoffRef)
	out.RedactionPolicyRef = normalizeOneDisplaySafeRef(out.RedactionPolicyRef)
	out.ExpectedStartedEventRef = normalizeOneDisplaySafeRef(out.ExpectedStartedEventRef)
	out.ExpectedCompletedEventRef = normalizeOneDisplaySafeRef(out.ExpectedCompletedEventRef)
	out.ExpectedResultRef = normalizeOneDisplaySafeRef(out.ExpectedResultRef)
	out.ExpectedReadbackRef = normalizeOneDisplaySafeRef(out.ExpectedReadbackRef)
	out.ExpectedFailureRef = normalizeOneDisplaySafeRef(out.ExpectedFailureRef)
	out.ExpectedCompensationRef = normalizeOneDisplaySafeRef(out.ExpectedCompensationRef)
	out.ExpectedCompletionHandoffRef = normalizeOneDisplaySafeRef(out.ExpectedCompletionHandoffRef)
	out.PreflightResultRefs = normalizeDisplaySafeRefs(out.PreflightResultRefs)
	out.AuthorizationRequiredInputs = normalizeMissingInputs(out.AuthorizationRequiredInputs)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.BlockedReasons = normalizeControlTokenList(out.BlockedReasons)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	if !out.Available {
		out.Status = "unavailable"
		out.ReadyForHostDisplay = false
		out.ReadyForHostAuthorization = false
		out.ReadyForHostInvocation = false
		out.HostInvocationAuthorized = false
	}
	if out.Status == "" {
		out.Status = "blocked"
	}
	if out.RawOutputLoaded {
		out.Status = "blocked"
		out.ReadyForHostDisplay = false
		out.ReadyForHostAuthorization = false
		out.ReadyForHostInvocation = false
		out.HostInvocationAuthorized = false
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
		out.AuthorizationPacketRef != "" &&
		out.PreflightReviewPacketRef != "" &&
		out.AdapterRef != "" &&
		out.DescriptorRef != "" &&
		!out.RawOutputLoaded
	ready := out.Status == "ready_for_host_adapter_invocation_authorization" &&
		out.ReadyForHostDisplay &&
		out.HostConfirmationRef != "" &&
		out.InvocationRef != "" &&
		out.IdempotencyRef != "" &&
		out.TimeoutPolicyRef != "" &&
		out.CancellationPolicyRef != "" &&
		out.ExpectedStartedEventRef != "" &&
		out.ExpectedCompletedEventRef != "" &&
		out.ExpectedResultRef != "" &&
		out.ExpectedReadbackRef != "" &&
		out.ExpectedFailureRef != "" &&
		out.ExpectedCompensationRef != "" &&
		out.ExpectedCompletionHandoffRef != "" &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.CoreInvocationExecuted &&
		!out.DurableWriteByCore
	out.ReadyForHostAuthorization = out.ReadyForHostAuthorization && ready
	out.ReadyForHostInvocation = out.ReadyForHostInvocation && ready
	out.HostInvocationAuthorized = out.HostInvocationAuthorized && ready
	return out
}

func productionAdapterInvocationAuthorizationPacketBlock(result ProductionAdapterInvocationAuthorizationPacket, failure FailureClass, reason string, missing MissingInput, next NextHostAction) ProductionAdapterInvocationAuthorizationPacket {
	result.Status = "blocked"
	result.ReadyForHostAuthorization = false
	result.ReadyForHostInvocation = false
	result.HostInvocationAuthorized = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	return result
}

type productionAdapterInvocationAuthorizationExpectedRefCheck struct {
	ref     DisplaySafeRef
	reason  string
	missing MissingInput
	next    NextHostAction
}

func productionAdapterInvocationAuthorizationExpectedRefChecks(result ProductionAdapterInvocationAuthorizationPacket) []productionAdapterInvocationAuthorizationExpectedRefCheck {
	return []productionAdapterInvocationAuthorizationExpectedRefCheck{
		{result.ExpectedStartedEventRef, "expected_started_event_ref_missing", "host:expected_started_event_ref", "provide_expected_invocation_refs"},
		{result.ExpectedCompletedEventRef, "expected_completed_event_ref_missing", "host:expected_completed_event_ref", "provide_expected_invocation_refs"},
		{result.ExpectedResultRef, "expected_result_ref_missing", "host:expected_result_ref", "provide_expected_invocation_refs"},
		{result.ExpectedReadbackRef, "expected_readback_ref_missing", "host:expected_readback_ref", "provide_expected_invocation_refs"},
		{result.ExpectedFailureRef, "expected_failure_ref_missing", "host:expected_failure_ref", "provide_expected_invocation_refs"},
		{result.ExpectedCompensationRef, "expected_compensation_ref_missing", "host:expected_compensation_ref", "provide_expected_invocation_refs"},
		{result.ExpectedCompletionHandoffRef, "expected_completion_handoff_ref_missing", "host:expected_completion_handoff_ref", "provide_expected_invocation_refs"},
	}
}

func productionAdapterInvocationAuthorizationApprovalRefs(review ProductionAdapterPreflightReviewPacket) []DisplaySafeRef {
	return normalizeDisplaySafeRefs(append(cloneDisplaySafeRefs(review.RequiredApprovalRefs), review.ApprovalContextRefs...))
}

func productionAdapterInvocationAuthorizationPacketUnsafe(input ProductionAdapterInvocationAuthorizationPacketInput) bool {
	return input.RawOutputLoaded ||
		displaySafeRefRejected(input.AuthorizationPacketRef) ||
		displaySafeRefRejected(input.InvocationRef) ||
		displaySafeRefRejected(input.HostConfirmationRef) ||
		displaySafeRefSliceRejected(input.ApprovalRefs) ||
		displaySafeRefRejected(input.IdempotencyRef) ||
		displaySafeRefRejected(input.ExpectedStartedEventRef) ||
		displaySafeRefRejected(input.ExpectedCompletedEventRef) ||
		displaySafeRefRejected(input.ExpectedResultRef) ||
		displaySafeRefRejected(input.ExpectedReadbackRef) ||
		displaySafeRefRejected(input.ExpectedFailureRef) ||
		displaySafeRefRejected(input.ExpectedCompensationRef) ||
		displaySafeRefRejected(input.ExpectedCompletionHandoffRef) ||
		displaySafeRefRejected(input.TimeoutPolicyRef) ||
		displaySafeRefRejected(input.CancellationPolicyRef) ||
		displaySafeRefRejected(input.HostPolicyRef) ||
		displaySafeRefRejected(input.BudgetRef)
}

func productionAdapterPreflightReviewPacketOutputUnsafe(packet ProductionAdapterPreflightReviewPacket) bool {
	return packet.RawOutputLoaded ||
		displaySafeRefRejected(packet.PreflightReviewPacketRef) ||
		displaySafeRefRejected(packet.AdapterRef) ||
		displaySafeRefRejected(packet.DescriptorRef) ||
		displaySafeRefRejected(packet.ApplyEnvelopeRef) ||
		displaySafeRefRejected(packet.SelectedSourceRef) ||
		displaySafeRefRejected(packet.SelectedCandidateRef) ||
		displaySafeRefRejected(packet.CatalogSnapshotRef) ||
		displaySafeRefRejected(packet.CatalogSelectionRef) ||
		displaySafeRefRejected(packet.HostPolicyRef) ||
		displaySafeRefSliceRejected(packet.ApprovalContextRefs) ||
		displaySafeRefRejected(packet.BudgetRef) ||
		displaySafeRefRejected(packet.IdempotencyRef) ||
		displaySafeRefRejected(packet.IdempotencyContractRef) ||
		displaySafeRefRejected(packet.RiskRef) ||
		displaySafeRefRejected(packet.TimeoutPolicyRef) ||
		displaySafeRefRejected(packet.CompensationHandoffRef) ||
		displaySafeRefRejected(packet.RedactionPolicyRef) ||
		displaySafeRefSliceRejected(packet.RequiredCapabilityRefs) ||
		displaySafeRefSliceRejected(packet.RequiredPolicyRefs) ||
		displaySafeRefSliceRejected(packet.RequiredApprovalRefs) ||
		displaySafeRefRejected(packet.RequiredBudgetRef) ||
		displaySafeRefSliceRejected(packet.RequiredPreflightRefs) ||
		displaySafeRefSliceRejected(packet.PreflightResultRefs)
}

func productionAdapterPreflightReviewPacketEmpty(packet ProductionAdapterPreflightReviewPacket) bool {
	return !packet.Projected &&
		!packet.Available &&
		packet.Status == "" &&
		packet.Mode == "" &&
		!packet.ReadyForHostDisplay &&
		!packet.ReadyForHostPreflight &&
		!packet.PreflightReported &&
		!packet.ReadyForHostInvocationReview &&
		packet.PreflightReviewPacketRef == "" &&
		packet.AdapterRef == "" &&
		packet.DescriptorRef == "" &&
		packet.ApplyEnvelopeRef == "" &&
		packet.SelectedSourceKind == "" &&
		packet.SelectedSourceRef == "" &&
		packet.SelectedCandidateRef == "" &&
		packet.CatalogSnapshotRef == "" &&
		packet.CatalogSelectionRef == "" &&
		packet.HostPolicyRef == "" &&
		len(packet.ApprovalContextRefs) == 0 &&
		packet.BudgetRef == "" &&
		packet.IdempotencyRef == "" &&
		len(packet.PreflightResultRefs) == 0 &&
		len(packet.MissingInputs) == 0 &&
		len(packet.BlockedReasons) == 0 &&
		len(packet.Boundaries) == 0 &&
		packet.NextHostAction == "" &&
		!packet.RawOutputLoaded
}

func productionAdapterInvocationAuthorizationRequiredInputs() []MissingInput {
	return []MissingInput{
		"host:invocation_authorization_packet_ref",
		"host:adapter_preflight_review_packet",
		"host:host_confirmation_ref",
		"host:invocation_ref",
		"host:approval_ref",
		"host:idempotency_ref",
		"host:expected_started_event_ref",
		"host:expected_completed_event_ref",
		"host:expected_result_ref",
		"host:expected_readback_ref",
		"host:expected_failure_ref",
		"host:expected_compensation_ref",
		"host:expected_completion_handoff_ref",
		"host:timeout_policy_ref",
		"host:cancellation_policy_ref",
		"host:policy_ref",
		"host:budget_ref",
	}
}

func productionAdapterInvocationAuthorizationPacketDisplaySections() []string {
	return []string{
		"adapter_invocation_authorization",
		"adapter_invocation_expected_events",
		"adapter_invocation_readback_handoff",
		"adapter_invocation_failure_handoff",
	}
}

func productionAdapterInvocationAuthorizationPacketBoundaries(groups ...[]Boundary) []Boundary {
	return MergeBoundaries(
		[]Boundary{
			"production_adapter_invocation_authorization_packet",
			"invocation_authorization_packet_projection_only",
			"host_owned_invocation_authorization",
			"host_final_confirmation_required",
			"display_safe_refs_only",
			"no_adapter_invocation",
			"no_runner_dispatch",
			"no_durable_write_by_core",
		},
		MergeBoundaries(groups...),
	)
}

func unavailableProductionAdapterInvocationAuthorizationPacket() ProductionAdapterInvocationAuthorizationPacket {
	return ProductionAdapterInvocationAuthorizationPacket{
		ContractVersion:             ContractVersion,
		Projected:                   true,
		Available:                   false,
		Status:                      "unavailable",
		Mode:                        "production_adapter_invocation_authorization_packet",
		RunnerEffect:                "none",
		PromptEffect:                "none",
		DisplaySections:             productionAdapterInvocationAuthorizationPacketDisplaySections(),
		AuthorizationRequiredInputs: productionAdapterInvocationAuthorizationRequiredInputs(),
		Boundaries: []Boundary{
			"production_adapter_invocation_authorization_packet",
			"invocation_authorization_packet_projection_only",
			"host_final_confirmation_required",
			"display_safe_refs_only",
			"no_adapter_invocation",
			"no_runner_dispatch",
			"no_durable_write_by_core",
		},
		NextHostAction: "provide_adapter_preflight_review_packet",
	}
}
