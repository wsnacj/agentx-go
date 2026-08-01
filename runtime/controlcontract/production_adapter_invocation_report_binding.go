package controlcontract

type ProductionAdapterInvocationReportBindingInput struct {
	InvocationReportBindingRef DisplaySafeRef                                 `json:"invocation_report_binding_ref,omitempty"`
	AuthorizationPacket        ProductionAdapterInvocationAuthorizationPacket `json:"authorization_packet,omitempty"`
	Invocation                 HostAdapterInvocationProjection                `json:"invocation,omitempty"`
	RawOutputLoaded            bool                                           `json:"raw_output_loaded"`
}

type ProductionAdapterInvocationReportBinding struct {
	ContractVersion              string              `json:"contract_version,omitempty"`
	Projected                    bool                `json:"projected"`
	Available                    bool                `json:"available"`
	Status                       string              `json:"status,omitempty"`
	Mode                         string              `json:"mode,omitempty"`
	RunnerEffect                 string              `json:"runner_effect,omitempty"`
	PromptEffect                 string              `json:"prompt_effect,omitempty"`
	ReadyForHostDisplay          bool                `json:"ready_for_host_display"`
	ReadyForReadbackReview       bool                `json:"ready_for_readback_review"`
	ReadyForFailureReview        bool                `json:"ready_for_failure_review"`
	AuthorizationBound           bool                `json:"authorization_bound"`
	HostInvocationAuthorized     bool                `json:"host_invocation_authorized"`
	HostInvocationReported       bool                `json:"host_invocation_reported"`
	HostInvocationCompleted      bool                `json:"host_invocation_completed"`
	HostInvocationFailed         bool                `json:"host_invocation_failed"`
	CoreInvocationExecuted       bool                `json:"core_invocation_executed"`
	DurableWriteByCore           bool                `json:"durable_write_by_core"`
	InvocationReportBindingRef   DisplaySafeRef      `json:"invocation_report_binding_ref,omitempty"`
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
	IdempotencyRef               DisplaySafeRef      `json:"idempotency_ref,omitempty"`
	ApprovalRefs                 []DisplaySafeRef    `json:"approval_refs,omitempty"`
	RequiredApprovalRefs         []DisplaySafeRef    `json:"required_approval_refs,omitempty"`
	ExpectedStartedEventRef      DisplaySafeRef      `json:"expected_started_event_ref,omitempty"`
	StartedEventRef              DisplaySafeRef      `json:"started_event_ref,omitempty"`
	ExpectedCompletedEventRef    DisplaySafeRef      `json:"expected_completed_event_ref,omitempty"`
	CompletedEventRef            DisplaySafeRef      `json:"completed_event_ref,omitempty"`
	ExpectedResultRef            DisplaySafeRef      `json:"expected_result_ref,omitempty"`
	ResultRef                    DisplaySafeRef      `json:"result_ref,omitempty"`
	ExpectedReadbackRef          DisplaySafeRef      `json:"expected_readback_ref,omitempty"`
	ReadbackRef                  DisplaySafeRef      `json:"readback_ref,omitempty"`
	ExpectedFailureRef           DisplaySafeRef      `json:"expected_failure_ref,omitempty"`
	FailureRef                   DisplaySafeRef      `json:"failure_ref,omitempty"`
	ExpectedCompensationRef      DisplaySafeRef      `json:"expected_compensation_ref,omitempty"`
	CompensationRef              DisplaySafeRef      `json:"compensation_ref,omitempty"`
	ExpectedCompletionHandoffRef DisplaySafeRef      `json:"expected_completion_handoff_ref,omitempty"`
	CompletionHandoffRef         DisplaySafeRef      `json:"completion_handoff_ref,omitempty"`
	ReportRequiredInputs         []MissingInput      `json:"report_required_inputs,omitempty"`
	MissingInputs                []MissingInput      `json:"missing_inputs,omitempty"`
	BlockedReasons               []string            `json:"blocked_reasons,omitempty"`
	FailureClass                 FailureClass        `json:"failure_class,omitempty"`
	Boundaries                   []Boundary          `json:"boundaries,omitempty"`
	NextHostAction               NextHostAction      `json:"next_host_action,omitempty"`
	RawOutputLoaded              bool                `json:"raw_output_loaded"`
}

func BuildProductionAdapterInvocationReportBinding(input ProductionAdapterInvocationReportBindingInput) ProductionAdapterInvocationReportBinding {
	if productionAdapterInvocationAuthorizationPacketEmpty(input.AuthorizationPacket) {
		return unavailableProductionAdapterInvocationReportBinding()
	}
	rawAuthorization := input.AuthorizationPacket
	authorization := rawAuthorization.Normalize()
	invocationProvided := !hostAdapterInvocationProjectionEmpty(input.Invocation)
	invocation := HostAdapterInvocationProjection{}
	if invocationProvided {
		invocation = input.Invocation.Normalize()
	}
	result := ProductionAdapterInvocationReportBinding{
		ContractVersion:              ContractVersion,
		Projected:                    true,
		Available:                    true,
		Status:                       "blocked",
		Mode:                         "production_adapter_invocation_report_binding",
		RunnerEffect:                 "none",
		PromptEffect:                 "none",
		InvocationReportBindingRef:   normalizeOneDisplaySafeRef(input.InvocationReportBindingRef),
		AuthorizationPacketRef:       authorization.AuthorizationPacketRef,
		PreflightReviewPacketRef:     authorization.PreflightReviewPacketRef,
		DisplaySections:              productionAdapterInvocationReportBindingDisplaySections(),
		AdapterRef:                   firstDisplaySafeRef(invocation.AdapterRef, authorization.AdapterRef),
		DescriptorRef:                firstDisplaySafeRef(invocation.DescriptorRef, authorization.DescriptorRef),
		ApplyEnvelopeRef:             authorization.ApplyEnvelopeRef,
		SelectedSourceKind:           authorization.SelectedSourceKind,
		SelectedSourceRef:            authorization.SelectedSourceRef,
		SelectedCandidateRef:         authorization.SelectedCandidateRef,
		CatalogSnapshotRef:           authorization.CatalogSnapshotRef,
		CatalogSelectionRef:          authorization.CatalogSelectionRef,
		InvocationRef:                firstDisplaySafeRef(invocation.InvocationRef, authorization.InvocationRef),
		HostConfirmationRef:          authorization.HostConfirmationRef,
		IdempotencyRef:               firstDisplaySafeRef(invocation.IdempotencyRef, authorization.IdempotencyRef),
		ApprovalRefs:                 normalizeDisplaySafeRefs(append(cloneDisplaySafeRefs(invocation.ApprovalRefs), authorization.ApprovalRefs...)),
		RequiredApprovalRefs:         cloneDisplaySafeRefs(authorization.RequiredApprovalRefs),
		ExpectedStartedEventRef:      authorization.ExpectedStartedEventRef,
		StartedEventRef:              invocation.StartedEventRef,
		ExpectedCompletedEventRef:    authorization.ExpectedCompletedEventRef,
		CompletedEventRef:            invocation.CompletedEventRef,
		ExpectedResultRef:            authorization.ExpectedResultRef,
		ResultRef:                    invocation.ResultRef,
		ExpectedReadbackRef:          authorization.ExpectedReadbackRef,
		ReadbackRef:                  invocation.ReadbackRef,
		ExpectedFailureRef:           authorization.ExpectedFailureRef,
		FailureRef:                   invocation.FailureRef,
		ExpectedCompensationRef:      authorization.ExpectedCompensationRef,
		CompensationRef:              invocation.CompensationRef,
		ExpectedCompletionHandoffRef: authorization.ExpectedCompletionHandoffRef,
		CompletionHandoffRef:         invocation.CompletionHandoffRef,
		ReportRequiredInputs:         productionAdapterInvocationReportBindingRequiredInputs(),
		HostInvocationAuthorized:     authorization.HostInvocationAuthorized,
		HostInvocationReported:       invocation.HostInvocationReported,
		HostInvocationCompleted:      invocation.HostInvocationCompleted,
		HostInvocationFailed:         invocation.HostInvocationFailed,
		FailureClass:                 FailureNone,
		Boundaries:                   productionAdapterInvocationReportBindingBoundaries(authorization.Boundaries, invocation.Boundaries),
		RawOutputLoaded:              input.RawOutputLoaded || authorization.RawOutputLoaded || invocation.RawOutputLoaded,
	}
	if productionAdapterInvocationReportBindingUnsafe(input, invocation, invocationProvided) || productionAdapterInvocationAuthorizationPacketOutputUnsafe(rawAuthorization) {
		result = productionAdapterInvocationReportBindingBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if !authorization.ReadyForHostInvocation || !authorization.HostInvocationAuthorized {
		result = productionAdapterInvocationReportBindingBlock(result, firstFailureClass(authorization.FailureClass, FailureAuthorizationMissing), "adapter_invocation_authorization_not_ready", "host:invocation_authorization_packet", firstNextHostAction(authorization.NextHostAction, "review_adapter_invocation_authorization"))
	}
	if result.InvocationReportBindingRef != "" && authorization.ReadyForHostInvocation && !result.RawOutputLoaded {
		result.ReadyForHostDisplay = true
	}
	if result.InvocationReportBindingRef == "" {
		result = productionAdapterInvocationReportBindingBlock(result, FailureEvidenceMissing, "invocation_report_binding_ref_missing", "host:invocation_report_binding_ref", "provide_invocation_report_binding_ref")
	}
	if !invocationProvided {
		result = productionAdapterInvocationReportBindingBlock(result, FailureEvidenceMissing, "adapter_invocation_report_missing", "host:adapter_invocation", "provide_adapter_invocation")
		return result.Normalize()
	}
	if invocation.Status != HostActionRecorded || !invocation.HostInvocationReported {
		result = productionAdapterInvocationReportBindingBlock(result, firstFailureClass(invocation.FailureClass, FailureEvidenceMissing), "adapter_invocation_not_recorded", "host:adapter_invocation", firstNextHostAction(invocation.NextHostAction, "provide_adapter_invocation"))
		return result.Normalize()
	}
	if invocation.HostInvocationCompleted && invocation.HostInvocationFailed {
		result = productionAdapterInvocationReportBindingBlock(result, FailureVerificationFailed, "invocation_result_conflict", "host:invocation_result_review", "review_invocation_result")
		return result.Normalize()
	}
	for _, check := range productionAdapterInvocationReportBindingCommonChecks(authorization, invocation) {
		if check.mismatch {
			result = productionAdapterInvocationReportBindingBlock(result, check.failure, check.reason, check.missing, "review_adapter_invocation_report")
		}
	}
	switch {
	case invocation.HostInvocationCompleted:
		for _, check := range productionAdapterInvocationReportBindingSuccessChecks(authorization, invocation) {
			if check.mismatch {
				result = productionAdapterInvocationReportBindingBlock(result, check.failure, check.reason, check.missing, "review_adapter_invocation_report")
			}
		}
		if len(result.BlockedReasons) == 0 && len(result.MissingInputs) == 0 {
			result.Status = "ready_for_authorized_invocation_readback_review"
			result.ReadyForHostDisplay = true
			result.ReadyForReadbackReview = true
			result.AuthorizationBound = true
			result.NextHostAction = "review_adapter_readback"
			result.Boundaries = AppendBoundaries(result.Boundaries, "authorized_invocation_success_report", "ready_for_authorized_invocation_readback_review")
		}
	case invocation.HostInvocationFailed:
		for _, check := range productionAdapterInvocationReportBindingFailureChecks(authorization, invocation) {
			if check.mismatch {
				result = productionAdapterInvocationReportBindingBlock(result, check.failure, check.reason, check.missing, "review_adapter_invocation_report")
			}
		}
		if len(result.BlockedReasons) == 0 && len(result.MissingInputs) == 0 {
			result.Status = "ready_for_authorized_invocation_failure_review"
			result.ReadyForHostDisplay = true
			result.ReadyForFailureReview = true
			result.AuthorizationBound = true
			result.FailureClass = FailureVerificationFailed
			result.NextHostAction = "review_adapter_failure"
			result.Boundaries = AppendBoundaries(result.Boundaries, "authorized_invocation_failure_report", "ready_for_authorized_invocation_failure_review")
		}
	default:
		result = productionAdapterInvocationReportBindingBlock(result, FailureEvidenceMissing, "invocation_result_missing", "host:invocation_result", "provide_invocation_result")
	}
	return result.Normalize()
}

func CloneProductionAdapterInvocationReportBinding(in ProductionAdapterInvocationReportBinding) ProductionAdapterInvocationReportBinding {
	out := in
	out.DisplaySections = cloneStringSlice(in.DisplaySections)
	out.ApprovalRefs = cloneDisplaySafeRefs(in.ApprovalRefs)
	out.RequiredApprovalRefs = cloneDisplaySafeRefs(in.RequiredApprovalRefs)
	out.ReportRequiredInputs = cloneMissingInputs(in.ReportRequiredInputs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (p ProductionAdapterInvocationReportBinding) Clone() ProductionAdapterInvocationReportBinding {
	return CloneProductionAdapterInvocationReportBinding(p)
}

func (p ProductionAdapterInvocationReportBinding) Normalize() ProductionAdapterInvocationReportBinding {
	out := CloneProductionAdapterInvocationReportBinding(p)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = normalizeControlToken(out.Status)
	out.Mode = normalizeControlToken(out.Mode)
	if out.Mode == "" {
		out.Mode = "production_adapter_invocation_report_binding"
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
	out.InvocationReportBindingRef = normalizeOneDisplaySafeRef(out.InvocationReportBindingRef)
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
	out.IdempotencyRef = normalizeOneDisplaySafeRef(out.IdempotencyRef)
	out.ApprovalRefs = normalizeDisplaySafeRefs(out.ApprovalRefs)
	out.RequiredApprovalRefs = normalizeDisplaySafeRefs(out.RequiredApprovalRefs)
	out.ExpectedStartedEventRef = normalizeOneDisplaySafeRef(out.ExpectedStartedEventRef)
	out.StartedEventRef = normalizeOneDisplaySafeRef(out.StartedEventRef)
	out.ExpectedCompletedEventRef = normalizeOneDisplaySafeRef(out.ExpectedCompletedEventRef)
	out.CompletedEventRef = normalizeOneDisplaySafeRef(out.CompletedEventRef)
	out.ExpectedResultRef = normalizeOneDisplaySafeRef(out.ExpectedResultRef)
	out.ResultRef = normalizeOneDisplaySafeRef(out.ResultRef)
	out.ExpectedReadbackRef = normalizeOneDisplaySafeRef(out.ExpectedReadbackRef)
	out.ReadbackRef = normalizeOneDisplaySafeRef(out.ReadbackRef)
	out.ExpectedFailureRef = normalizeOneDisplaySafeRef(out.ExpectedFailureRef)
	out.FailureRef = normalizeOneDisplaySafeRef(out.FailureRef)
	out.ExpectedCompensationRef = normalizeOneDisplaySafeRef(out.ExpectedCompensationRef)
	out.CompensationRef = normalizeOneDisplaySafeRef(out.CompensationRef)
	out.ExpectedCompletionHandoffRef = normalizeOneDisplaySafeRef(out.ExpectedCompletionHandoffRef)
	out.CompletionHandoffRef = normalizeOneDisplaySafeRef(out.CompletionHandoffRef)
	out.ReportRequiredInputs = normalizeMissingInputs(out.ReportRequiredInputs)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.BlockedReasons = normalizeControlTokenList(out.BlockedReasons)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	if !out.Available {
		out.Status = "unavailable"
		out.ReadyForHostDisplay = false
		out.ReadyForReadbackReview = false
		out.ReadyForFailureReview = false
		out.AuthorizationBound = false
	}
	if out.Status == "" {
		out.Status = "blocked"
	}
	if out.RawOutputLoaded {
		out.Status = "blocked"
		out.ReadyForHostDisplay = false
		out.ReadyForReadbackReview = false
		out.ReadyForFailureReview = false
		out.AuthorizationBound = false
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
		out.InvocationReportBindingRef != "" &&
		out.AuthorizationPacketRef != "" &&
		out.AdapterRef != "" &&
		out.DescriptorRef != "" &&
		!out.RawOutputLoaded
	successReady := out.Status == "ready_for_authorized_invocation_readback_review" &&
		out.ReadyForHostDisplay &&
		out.HostInvocationAuthorized &&
		out.HostInvocationReported &&
		out.HostInvocationCompleted &&
		!out.HostInvocationFailed &&
		out.InvocationRef != "" &&
		out.IdempotencyRef != "" &&
		out.StartedEventRef == out.ExpectedStartedEventRef &&
		out.CompletedEventRef == out.ExpectedCompletedEventRef &&
		out.ResultRef == out.ExpectedResultRef &&
		out.ReadbackRef == out.ExpectedReadbackRef &&
		out.CompletionHandoffRef == out.ExpectedCompletionHandoffRef &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.CoreInvocationExecuted &&
		!out.DurableWriteByCore
	failureReady := out.Status == "ready_for_authorized_invocation_failure_review" &&
		out.ReadyForHostDisplay &&
		out.HostInvocationAuthorized &&
		out.HostInvocationReported &&
		out.HostInvocationFailed &&
		!out.HostInvocationCompleted &&
		out.InvocationRef != "" &&
		out.IdempotencyRef != "" &&
		out.StartedEventRef == out.ExpectedStartedEventRef &&
		out.CompletedEventRef == out.ExpectedCompletedEventRef &&
		out.FailureRef == out.ExpectedFailureRef &&
		out.CompensationRef == out.ExpectedCompensationRef &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.CoreInvocationExecuted &&
		!out.DurableWriteByCore
	out.ReadyForReadbackReview = out.ReadyForReadbackReview && successReady
	out.ReadyForFailureReview = out.ReadyForFailureReview && failureReady
	out.AuthorizationBound = out.AuthorizationBound && (successReady || failureReady)
	return out
}

func productionAdapterInvocationReportBindingBlock(result ProductionAdapterInvocationReportBinding, failure FailureClass, reason string, missing MissingInput, next NextHostAction) ProductionAdapterInvocationReportBinding {
	result.Status = "blocked"
	result.ReadyForReadbackReview = false
	result.ReadyForFailureReview = false
	result.AuthorizationBound = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	return result
}

type productionAdapterInvocationReportBindingCheck struct {
	mismatch bool
	reason   string
	missing  MissingInput
	failure  FailureClass
}

func productionAdapterInvocationReportBindingCommonChecks(authorization ProductionAdapterInvocationAuthorizationPacket, invocation HostAdapterInvocationProjection) []productionAdapterInvocationReportBindingCheck {
	return []productionAdapterInvocationReportBindingCheck{
		{authorization.AdapterRef != "" && invocation.AdapterRef != "" && authorization.AdapterRef != invocation.AdapterRef, "invocation_report_adapter_ref_mismatch", "host:adapter_invocation", FailureInvalidInput},
		{authorization.DescriptorRef != "" && invocation.DescriptorRef != "" && authorization.DescriptorRef != invocation.DescriptorRef, "invocation_report_descriptor_ref_mismatch", "host:adapter_invocation", FailureInvalidInput},
		{authorization.InvocationRef != "" && invocation.InvocationRef != "" && authorization.InvocationRef != invocation.InvocationRef, "invocation_report_invocation_ref_mismatch", "host:invocation_ref", FailureInvalidInput},
		{authorization.IdempotencyRef != "" && invocation.IdempotencyRef != "" && authorization.IdempotencyRef != invocation.IdempotencyRef, "invocation_report_idempotency_ref_mismatch", "host:idempotency_ref", FailureInvalidInput},
		{len(authorization.RequiredApprovalRefs) > 0 && !displaySafeRefsContainAll(invocation.ApprovalRefs, authorization.RequiredApprovalRefs), "invocation_report_approval_ref_missing", "host:approval_ref", FailureApprovalRequired},
		{authorization.ExpectedStartedEventRef != "" && invocation.StartedEventRef != "" && authorization.ExpectedStartedEventRef != invocation.StartedEventRef, "invocation_report_started_event_ref_mismatch", "host:started_event_ref", FailureVerificationFailed},
		{authorization.ExpectedCompletedEventRef != "" && invocation.CompletedEventRef != "" && authorization.ExpectedCompletedEventRef != invocation.CompletedEventRef, "invocation_report_completed_event_ref_mismatch", "host:completed_event_ref", FailureVerificationFailed},
	}
}

func productionAdapterInvocationReportBindingSuccessChecks(authorization ProductionAdapterInvocationAuthorizationPacket, invocation HostAdapterInvocationProjection) []productionAdapterInvocationReportBindingCheck {
	return []productionAdapterInvocationReportBindingCheck{
		{authorization.ExpectedResultRef != "" && invocation.ResultRef != "" && authorization.ExpectedResultRef != invocation.ResultRef, "invocation_report_result_ref_mismatch", "host:result_ref", FailureVerificationFailed},
		{authorization.ExpectedReadbackRef != "" && invocation.ReadbackRef != "" && authorization.ExpectedReadbackRef != invocation.ReadbackRef, "invocation_report_readback_ref_mismatch", "host:readback_ref", FailureVerificationFailed},
		{authorization.ExpectedCompletionHandoffRef != "" && invocation.CompletionHandoffRef != "" && authorization.ExpectedCompletionHandoffRef != invocation.CompletionHandoffRef, "invocation_report_completion_handoff_ref_mismatch", "host:completion_handoff_ref", FailureVerificationFailed},
	}
}

func productionAdapterInvocationReportBindingFailureChecks(authorization ProductionAdapterInvocationAuthorizationPacket, invocation HostAdapterInvocationProjection) []productionAdapterInvocationReportBindingCheck {
	return []productionAdapterInvocationReportBindingCheck{
		{authorization.ExpectedFailureRef != "" && invocation.FailureRef != "" && authorization.ExpectedFailureRef != invocation.FailureRef, "invocation_report_failure_ref_mismatch", "host:failure_ref", FailureVerificationFailed},
		{authorization.ExpectedCompensationRef != "" && invocation.CompensationRef != "" && authorization.ExpectedCompensationRef != invocation.CompensationRef, "invocation_report_compensation_ref_mismatch", "host:compensation_ref", FailureVerificationFailed},
	}
}

func productionAdapterInvocationReportBindingUnsafe(input ProductionAdapterInvocationReportBindingInput, invocation HostAdapterInvocationProjection, invocationProvided bool) bool {
	return input.RawOutputLoaded ||
		displaySafeRefRejected(input.InvocationReportBindingRef) ||
		(invocationProvided && productionAdapterInvocationProjectionUnsafe(input.Invocation)) ||
		(invocationProvided && productionAdapterInvocationProjectionUnsafe(invocation))
}

func productionAdapterInvocationAuthorizationPacketOutputUnsafe(packet ProductionAdapterInvocationAuthorizationPacket) bool {
	return packet.RawOutputLoaded ||
		displaySafeRefRejected(packet.AuthorizationPacketRef) ||
		displaySafeRefRejected(packet.PreflightReviewPacketRef) ||
		displaySafeRefRejected(packet.AdapterRef) ||
		displaySafeRefRejected(packet.DescriptorRef) ||
		displaySafeRefRejected(packet.ApplyEnvelopeRef) ||
		displaySafeRefRejected(packet.SelectedSourceRef) ||
		displaySafeRefRejected(packet.SelectedCandidateRef) ||
		displaySafeRefRejected(packet.CatalogSnapshotRef) ||
		displaySafeRefRejected(packet.CatalogSelectionRef) ||
		displaySafeRefRejected(packet.InvocationRef) ||
		displaySafeRefRejected(packet.HostConfirmationRef) ||
		displaySafeRefRejected(packet.HostPolicyRef) ||
		displaySafeRefSliceRejected(packet.ApprovalContextRefs) ||
		displaySafeRefSliceRejected(packet.ApprovalRefs) ||
		displaySafeRefSliceRejected(packet.RequiredApprovalRefs) ||
		displaySafeRefRejected(packet.BudgetRef) ||
		displaySafeRefRejected(packet.IdempotencyRef) ||
		displaySafeRefRejected(packet.IdempotencyContractRef) ||
		displaySafeRefRejected(packet.RiskRef) ||
		displaySafeRefRejected(packet.TimeoutPolicyRef) ||
		displaySafeRefRejected(packet.CancellationPolicyRef) ||
		displaySafeRefRejected(packet.CompensationHandoffRef) ||
		displaySafeRefRejected(packet.RedactionPolicyRef) ||
		displaySafeRefRejected(packet.ExpectedStartedEventRef) ||
		displaySafeRefRejected(packet.ExpectedCompletedEventRef) ||
		displaySafeRefRejected(packet.ExpectedResultRef) ||
		displaySafeRefRejected(packet.ExpectedReadbackRef) ||
		displaySafeRefRejected(packet.ExpectedFailureRef) ||
		displaySafeRefRejected(packet.ExpectedCompensationRef) ||
		displaySafeRefRejected(packet.ExpectedCompletionHandoffRef) ||
		displaySafeRefSliceRejected(packet.PreflightResultRefs)
}

func productionAdapterInvocationAuthorizationPacketEmpty(packet ProductionAdapterInvocationAuthorizationPacket) bool {
	return !packet.Projected &&
		!packet.Available &&
		packet.Status == "" &&
		packet.Mode == "" &&
		!packet.ReadyForHostDisplay &&
		!packet.ReadyForHostInvocation &&
		!packet.ReadyForHostAuthorization &&
		!packet.HostInvocationAuthorized &&
		packet.AuthorizationPacketRef == "" &&
		packet.PreflightReviewPacketRef == "" &&
		packet.AdapterRef == "" &&
		packet.DescriptorRef == "" &&
		packet.InvocationRef == "" &&
		packet.HostConfirmationRef == "" &&
		packet.IdempotencyRef == "" &&
		len(packet.ApprovalRefs) == 0 &&
		len(packet.RequiredApprovalRefs) == 0 &&
		packet.ExpectedStartedEventRef == "" &&
		packet.ExpectedCompletedEventRef == "" &&
		packet.ExpectedResultRef == "" &&
		packet.ExpectedReadbackRef == "" &&
		packet.ExpectedFailureRef == "" &&
		packet.ExpectedCompensationRef == "" &&
		packet.ExpectedCompletionHandoffRef == "" &&
		len(packet.MissingInputs) == 0 &&
		len(packet.BlockedReasons) == 0 &&
		len(packet.Boundaries) == 0 &&
		packet.NextHostAction == "" &&
		!packet.RawOutputLoaded
}

func hostAdapterInvocationProjectionEmpty(invocation HostAdapterInvocationProjection) bool {
	return !invocation.Projected &&
		invocation.Status == "" &&
		!invocation.HostInvocationReported &&
		!invocation.HostInvocationCompleted &&
		!invocation.HostInvocationFailed &&
		invocation.AdapterRef == "" &&
		invocation.DescriptorRef == "" &&
		invocation.InvocationRef == "" &&
		invocation.IdempotencyRef == "" &&
		len(invocation.ApprovalRefs) == 0 &&
		invocation.StartedEventRef == "" &&
		invocation.CompletedEventRef == "" &&
		invocation.ResultRef == "" &&
		invocation.FailureRef == "" &&
		invocation.ReadbackRef == "" &&
		invocation.CompensationRef == "" &&
		invocation.CompletionHandoffRef == "" &&
		len(invocation.MissingInputs) == 0 &&
		len(invocation.BlockedReasons) == 0 &&
		len(invocation.Boundaries) == 0 &&
		invocation.NextHostAction == "" &&
		!invocation.RawOutputLoaded
}

func productionAdapterInvocationReportBindingRequiredInputs() []MissingInput {
	return []MissingInput{
		"host:invocation_report_binding_ref",
		"host:invocation_authorization_packet",
		"host:adapter_invocation",
		"host:started_event_ref",
		"host:completed_event_ref",
		"host:result_ref",
		"host:readback_ref",
		"host:failure_ref",
		"host:compensation_ref",
		"host:completion_handoff_ref",
	}
}

func productionAdapterInvocationReportBindingDisplaySections() []string {
	return []string{
		"adapter_invocation_report_binding",
		"adapter_invocation_authorization_trace",
		"adapter_invocation_result_refs",
		"adapter_invocation_failure_refs",
	}
}

func productionAdapterInvocationReportBindingBoundaries(groups ...[]Boundary) []Boundary {
	return MergeBoundaries(
		[]Boundary{
			"production_adapter_invocation_report_binding",
			"invocation_report_binding_projection_only",
			"host_owned_invocation_report",
			"authorization_bound_invocation_report",
			"display_safe_refs_only",
			"core_invocation_not_executed",
			"no_adapter_invocation",
			"no_runner_dispatch",
			"no_durable_write_by_core",
		},
		MergeBoundaries(groups...),
	)
}

func unavailableProductionAdapterInvocationReportBinding() ProductionAdapterInvocationReportBinding {
	return ProductionAdapterInvocationReportBinding{
		ContractVersion:      ContractVersion,
		Projected:            true,
		Available:            false,
		Status:               "unavailable",
		Mode:                 "production_adapter_invocation_report_binding",
		RunnerEffect:         "none",
		PromptEffect:         "none",
		DisplaySections:      productionAdapterInvocationReportBindingDisplaySections(),
		ReportRequiredInputs: productionAdapterInvocationReportBindingRequiredInputs(),
		Boundaries: []Boundary{
			"production_adapter_invocation_report_binding",
			"invocation_report_binding_projection_only",
			"host_owned_invocation_report",
			"authorization_bound_invocation_report",
			"display_safe_refs_only",
			"core_invocation_not_executed",
			"no_adapter_invocation",
			"no_runner_dispatch",
			"no_durable_write_by_core",
		},
		NextHostAction: "provide_invocation_authorization_packet",
	}
}
