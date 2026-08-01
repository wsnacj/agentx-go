package controlcontract

type ProductionAdapterObjectiveCloseoutPacketInput struct {
	ObjectiveCloseoutPacketRef   DisplaySafeRef                                `json:"objective_closeout_packet_ref,omitempty"`
	ObjectiveRef                 DisplaySafeRef                                `json:"objective_ref,omitempty"`
	HostCloseoutConfirmationRef  DisplaySafeRef                                `json:"host_closeout_confirmation_ref,omitempty"`
	CompletionAuditResultBinding ProductionAdapterCompletionAuditResultBinding `json:"completion_audit_result_binding,omitempty"`
	RawOutputLoaded              bool                                          `json:"raw_output_loaded"`
}

type ProductionAdapterObjectiveCloseoutPacket struct {
	ContractVersion                 string             `json:"contract_version,omitempty"`
	Projected                       bool               `json:"projected"`
	Available                       bool               `json:"available"`
	Status                          VerificationStatus `json:"status,omitempty"`
	Mode                            string             `json:"mode,omitempty"`
	ReadyForHostCloseoutReview      bool               `json:"ready_for_host_closeout_review"`
	ReadyForObjectiveCompletion     bool               `json:"ready_for_objective_completion"`
	ObjectiveSatisfied              bool               `json:"objective_satisfied"`
	VerificationSatisfied           bool               `json:"verification_satisfied"`
	HostCloseoutConfirmed           bool               `json:"host_closeout_confirmed"`
	AuthorizationBound              bool               `json:"authorization_bound"`
	CompletionAuditBound            bool               `json:"completion_audit_bound"`
	CoreInvocationExecuted          bool               `json:"core_invocation_executed"`
	DurableWriteByCore              bool               `json:"durable_write_by_core"`
	ObjectiveCloseoutPacketRef      DisplaySafeRef     `json:"objective_closeout_packet_ref,omitempty"`
	ObjectiveRef                    DisplaySafeRef     `json:"objective_ref,omitempty"`
	HostCloseoutConfirmationRef     DisplaySafeRef     `json:"host_closeout_confirmation_ref,omitempty"`
	CompletionAuditResultBindingRef DisplaySafeRef     `json:"completion_audit_result_binding_ref,omitempty"`
	CompletionAuditResultRef        DisplaySafeRef     `json:"completion_audit_result_ref,omitempty"`
	InvocationReportBindingRef      DisplaySafeRef     `json:"invocation_report_binding_ref,omitempty"`
	AuthorizationPacketRef          DisplaySafeRef     `json:"authorization_packet_ref,omitempty"`
	PreflightReviewPacketRef        DisplaySafeRef     `json:"preflight_review_packet_ref,omitempty"`
	AdapterRef                      DisplaySafeRef     `json:"adapter_ref,omitempty"`
	DescriptorRef                   DisplaySafeRef     `json:"descriptor_ref,omitempty"`
	InvocationRef                   DisplaySafeRef     `json:"invocation_ref,omitempty"`
	ResultRef                       DisplaySafeRef     `json:"result_ref,omitempty"`
	ReadbackRef                     DisplaySafeRef     `json:"readback_ref,omitempty"`
	CompletionHandoffRef            DisplaySafeRef     `json:"completion_handoff_ref,omitempty"`
	EvidenceRefs                    []EvidenceRef      `json:"evidence_refs,omitempty"`
	Verification                    VerificationResult `json:"verification,omitempty"`
	MissingInputs                   []MissingInput     `json:"missing_inputs,omitempty"`
	BlockedReasons                  []string           `json:"blocked_reasons,omitempty"`
	FailureClass                    FailureClass       `json:"failure_class,omitempty"`
	Boundaries                      []Boundary         `json:"boundaries,omitempty"`
	NextHostAction                  NextHostAction     `json:"next_host_action,omitempty"`
	RunnerEffect                    string             `json:"runner_effect,omitempty"`
	PromptEffect                    string             `json:"prompt_effect,omitempty"`
	RawOutputLoaded                 bool               `json:"raw_output_loaded"`
}

func BuildProductionAdapterObjectiveCloseoutPacket(input ProductionAdapterObjectiveCloseoutPacketInput) ProductionAdapterObjectiveCloseoutPacket {
	unsafe := input.RawOutputLoaded ||
		displaySafeRefRejected(input.ObjectiveCloseoutPacketRef) ||
		displaySafeRefRejected(input.ObjectiveRef) ||
		displaySafeRefRejected(input.HostCloseoutConfirmationRef) ||
		productionAdapterCompletionAuditResultBindingUnsafe(input.CompletionAuditResultBinding)
	binding := input.CompletionAuditResultBinding.Normalize()
	verification := binding.Verification.Normalize()
	if unsafe {
		verification = VerificationResult{}
	}
	result := ProductionAdapterObjectiveCloseoutPacket{
		ContractVersion:                 ContractVersion,
		Projected:                       true,
		Available:                       true,
		Status:                          VerificationBlocked,
		Mode:                            "production_adapter_objective_closeout_packet",
		AuthorizationBound:              binding.AuthorizationBound,
		CompletionAuditBound:            binding.CompletionAuditBound,
		ObjectiveCloseoutPacketRef:      normalizeOneDisplaySafeRef(input.ObjectiveCloseoutPacketRef),
		ObjectiveRef:                    normalizeOneDisplaySafeRef(input.ObjectiveRef),
		HostCloseoutConfirmationRef:     normalizeOneDisplaySafeRef(input.HostCloseoutConfirmationRef),
		CompletionAuditResultBindingRef: binding.CompletionAuditResultBindingRef,
		CompletionAuditResultRef:        binding.CompletionAuditResultRef,
		InvocationReportBindingRef:      binding.InvocationReportBindingRef,
		AuthorizationPacketRef:          binding.AuthorizationPacketRef,
		PreflightReviewPacketRef:        binding.PreflightReviewPacketRef,
		AdapterRef:                      binding.AdapterRef,
		DescriptorRef:                   binding.DescriptorRef,
		InvocationRef:                   binding.InvocationRef,
		ResultRef:                       binding.ResultRef,
		ReadbackRef:                     binding.ReadbackRef,
		CompletionHandoffRef:            binding.CompletionHandoffRef,
		EvidenceRefs:                    productionAdapterObjectiveCloseoutEvidenceRefs(binding, normalizeOneDisplaySafeRef(input.ObjectiveCloseoutPacketRef), normalizeOneDisplaySafeRef(input.ObjectiveRef), normalizeOneDisplaySafeRef(input.HostCloseoutConfirmationRef)),
		Verification:                    verification,
		MissingInputs:                   cloneMissingInputs(binding.MissingInputs),
		BlockedReasons:                  cloneStringSlice(binding.BlockedReasons),
		FailureClass:                    firstFailureClass(binding.FailureClass, verification.FailureClass),
		Boundaries:                      productionAdapterObjectiveCloseoutBoundaries(binding.Boundaries),
		NextHostAction:                  firstNextHostAction(binding.NextHostAction, "review_objective_closeout"),
		RunnerEffect:                    "none",
		PromptEffect:                    "none",
		RawOutputLoaded:                 input.RawOutputLoaded || binding.RawOutputLoaded || verification.RawOutputLoaded,
	}
	if unsafe {
		result = productionAdapterObjectiveCloseoutBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if result.ObjectiveCloseoutPacketRef == "" {
		result = productionAdapterObjectiveCloseoutBlock(result, FailureEvidenceMissing, "objective_closeout_packet_ref_missing", "host:objective_closeout_packet_ref", "provide_objective_closeout_packet")
	}
	if result.ObjectiveRef == "" {
		result = productionAdapterObjectiveCloseoutBlock(result, FailureEvidenceMissing, "objective_ref_missing", "host:objective_ref", "provide_objective_ref")
	}
	if !binding.ReadyForObjectiveCloseout || !binding.VerificationSatisfied || !binding.CompletionAuditBound {
		result = productionAdapterObjectiveCloseoutBlock(result, firstFailureClass(binding.FailureClass, FailureEvidenceMissing), "completion_audit_result_binding_not_ready", "host:completion_audit_result_binding", firstNextHostAction(binding.NextHostAction, "review_completion_audit_result"))
	}
	if len(result.MissingInputs) > 0 || len(result.BlockedReasons) > 0 {
		return result.Normalize()
	}
	result.ReadyForHostCloseoutReview = true
	result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_host_objective_closeout_review")
	if result.HostCloseoutConfirmationRef == "" {
		result.Status = VerificationReviewRequired
		result.FailureClass = FailureApprovalRequired
		result.MissingInputs = AppendMissingInputs(result.MissingInputs, "host:objective_closeout_confirmation_ref")
		result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, "objective_closeout_confirmation_required")
		result.Boundaries = AppendBoundaries(result.Boundaries, "host_objective_closeout_confirmation_required")
		result.NextHostAction = "confirm_objective_closeout"
		return result.Normalize()
	}
	result.Status = VerificationSatisfied
	result.FailureClass = FailureNone
	result.HostCloseoutConfirmed = true
	result.Boundaries = AppendBoundaries(result.Boundaries, "objective_closeout_confirmed", "objective_satisfied_by_host_closeout", "ready_for_objective_completion")
	result.NextHostAction = "return_objective_closeout"
	return result.Normalize()
}

func CloneProductionAdapterObjectiveCloseoutPacket(in ProductionAdapterObjectiveCloseoutPacket) ProductionAdapterObjectiveCloseoutPacket {
	out := in
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.Verification = in.Verification.Clone()
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (p ProductionAdapterObjectiveCloseoutPacket) Clone() ProductionAdapterObjectiveCloseoutPacket {
	return CloneProductionAdapterObjectiveCloseoutPacket(p)
}

func (p ProductionAdapterObjectiveCloseoutPacket) Normalize() ProductionAdapterObjectiveCloseoutPacket {
	out := CloneProductionAdapterObjectiveCloseoutPacket(p)
	unsafe := productionAdapterObjectiveCloseoutPacketUnsafe(out)
	if unsafe {
		out.Verification = VerificationResult{}
	}
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Mode = normalizeControlToken(out.Mode)
	if out.Mode == "" {
		out.Mode = "production_adapter_objective_closeout_packet"
	}
	out.Status = NormalizeVerificationStatus(string(out.Status))
	out.ObjectiveCloseoutPacketRef = normalizeOneDisplaySafeRef(out.ObjectiveCloseoutPacketRef)
	out.ObjectiveRef = normalizeOneDisplaySafeRef(out.ObjectiveRef)
	out.HostCloseoutConfirmationRef = normalizeOneDisplaySafeRef(out.HostCloseoutConfirmationRef)
	out.CompletionAuditResultBindingRef = normalizeOneDisplaySafeRef(out.CompletionAuditResultBindingRef)
	out.CompletionAuditResultRef = normalizeOneDisplaySafeRef(out.CompletionAuditResultRef)
	out.InvocationReportBindingRef = normalizeOneDisplaySafeRef(out.InvocationReportBindingRef)
	out.AuthorizationPacketRef = normalizeOneDisplaySafeRef(out.AuthorizationPacketRef)
	out.PreflightReviewPacketRef = normalizeOneDisplaySafeRef(out.PreflightReviewPacketRef)
	out.AdapterRef = normalizeOneDisplaySafeRef(out.AdapterRef)
	out.DescriptorRef = normalizeOneDisplaySafeRef(out.DescriptorRef)
	out.InvocationRef = normalizeOneDisplaySafeRef(out.InvocationRef)
	out.ResultRef = normalizeOneDisplaySafeRef(out.ResultRef)
	out.ReadbackRef = normalizeOneDisplaySafeRef(out.ReadbackRef)
	out.CompletionHandoffRef = normalizeOneDisplaySafeRef(out.CompletionHandoffRef)
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
	if !out.Available {
		out.Status = VerificationBlocked
		out.ReadyForHostCloseoutReview = false
		out.ReadyForObjectiveCompletion = false
		out.HostCloseoutConfirmed = false
	}
	if unsafe || out.RawOutputLoaded {
		out.RawOutputLoaded = true
		out.Status = VerificationBlocked
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
	out.AuthorizationBound = out.AuthorizationBound &&
		out.InvocationReportBindingRef != "" &&
		out.AuthorizationPacketRef != "" &&
		out.PreflightReviewPacketRef != "" &&
		!containsMissingInput(out.MissingInputs, "host:display_safe_refs")
	out.CompletionAuditBound = out.CompletionAuditBound &&
		out.AuthorizationBound &&
		out.CompletionAuditResultBindingRef != "" &&
		out.CompletionAuditResultRef != "" &&
		out.CompletionHandoffRef != "" &&
		!containsMissingInput(out.MissingInputs, "host:completion_audit_result_binding") &&
		!containsMissingInput(out.MissingInputs, "host:display_safe_refs")
	out.VerificationSatisfied = out.CompletionAuditBound &&
		out.Verification.Status == VerificationSatisfied &&
		out.Verification.Satisfied &&
		out.Verification.FailureClass == FailureNone &&
		len(out.Verification.MissingInputs) == 0
	out.HostCloseoutConfirmed = out.HostCloseoutConfirmed &&
		out.HostCloseoutConfirmationRef != "" &&
		!containsMissingInput(out.MissingInputs, "host:objective_closeout_confirmation_ref") &&
		!out.RawOutputLoaded
	out.ReadyForHostCloseoutReview = out.ReadyForHostCloseoutReview &&
		out.Available &&
		out.ObjectiveCloseoutPacketRef != "" &&
		out.ObjectiveRef != "" &&
		out.CompletionAuditBound &&
		out.VerificationSatisfied &&
		!containsMissingInput(out.MissingInputs, "host:objective_closeout_packet_ref") &&
		!containsMissingInput(out.MissingInputs, "host:objective_ref") &&
		!containsMissingInput(out.MissingInputs, "host:completion_audit_result_binding") &&
		!containsMissingInput(out.MissingInputs, "host:display_safe_refs") &&
		!out.RawOutputLoaded
	out.ReadyForObjectiveCompletion = out.Status == VerificationSatisfied &&
		out.ReadyForHostCloseoutReview &&
		out.HostCloseoutConfirmed &&
		out.VerificationSatisfied &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.RawOutputLoaded
	out.ObjectiveSatisfied = out.ReadyForObjectiveCompletion
	if !out.ReadyForObjectiveCompletion && out.Status == VerificationSatisfied {
		out.Status = VerificationReviewRequired
	}
	return out
}

func productionAdapterObjectiveCloseoutBlock(result ProductionAdapterObjectiveCloseoutPacket, failure FailureClass, reason string, missing MissingInput, next NextHostAction) ProductionAdapterObjectiveCloseoutPacket {
	result.Status = VerificationBlocked
	result.ReadyForHostCloseoutReview = false
	result.ReadyForObjectiveCompletion = false
	result.ObjectiveSatisfied = false
	result.HostCloseoutConfirmed = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	result.Boundaries = AppendBoundaries(result.Boundaries, "objective_closeout_packet_blocked")
	return result
}

func productionAdapterObjectiveCloseoutEvidenceRefs(binding ProductionAdapterCompletionAuditResultBinding, packetRef DisplaySafeRef, objectiveRef DisplaySafeRef, confirmationRef DisplaySafeRef) []EvidenceRef {
	evidence := cloneEvidenceRefs(binding.EvidenceRefs)
	source := firstDisplaySafeRef(binding.CompletionAuditResultBindingRef, binding.CompletionAuditResultRef, binding.CompletionHandoffRef)
	if packetRef != "" {
		evidence = append(evidence, EvidenceRef{
			Ref:      packetRef,
			Kind:     "objective_closeout_packet",
			Strength: EvidenceAdequate,
			Source:   source,
		})
	}
	if objectiveRef != "" {
		evidence = append(evidence, EvidenceRef{
			Ref:      objectiveRef,
			Kind:     "objective",
			Strength: EvidenceAdequate,
			Source:   source,
		})
	}
	if confirmationRef != "" {
		evidence = append(evidence, EvidenceRef{
			Ref:      confirmationRef,
			Kind:     "host_objective_closeout_confirmation",
			Strength: EvidenceStrong,
			Source:   firstDisplaySafeRef(packetRef, objectiveRef, source),
		})
	}
	return normalizeEvidenceRefs(evidence)
}

func productionAdapterObjectiveCloseoutBoundaries(bindingBoundaries []Boundary) []Boundary {
	out := cloneBoundaries(bindingBoundaries)
	for _, item := range []Boundary{
		"production_adapter_objective_closeout_packet",
		"objective_closeout_projection_only",
		"host_owned_objective_closeout",
		"display_safe_refs_only",
		"display_safe_result_refs_only",
		"no_runner_dispatch",
		"no_durable_write_by_core",
	} {
		out = AppendBoundaries(out, item)
	}
	return normalizeBoundaries(out)
}

func productionAdapterObjectiveCloseoutPacketUnsafe(input ProductionAdapterObjectiveCloseoutPacket) bool {
	return displaySafeRefRejected(input.ObjectiveCloseoutPacketRef) ||
		displaySafeRefRejected(input.ObjectiveRef) ||
		displaySafeRefRejected(input.HostCloseoutConfirmationRef) ||
		displaySafeRefRejected(input.CompletionAuditResultBindingRef) ||
		displaySafeRefRejected(input.CompletionAuditResultRef) ||
		displaySafeRefRejected(input.InvocationReportBindingRef) ||
		displaySafeRefRejected(input.AuthorizationPacketRef) ||
		displaySafeRefRejected(input.PreflightReviewPacketRef) ||
		displaySafeRefRejected(input.AdapterRef) ||
		displaySafeRefRejected(input.DescriptorRef) ||
		displaySafeRefRejected(input.InvocationRef) ||
		displaySafeRefRejected(input.ResultRef) ||
		displaySafeRefRejected(input.ReadbackRef) ||
		displaySafeRefRejected(input.CompletionHandoffRef) ||
		evidenceRefsRejected(input.EvidenceRefs) ||
		verificationResultUnsafe(input.Verification) ||
		input.RawOutputLoaded
}
