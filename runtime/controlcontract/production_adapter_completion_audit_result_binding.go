package controlcontract

type ProductionAdapterCompletionAuditResultBindingInput struct {
	CompletionAuditResultBindingRef DisplaySafeRef                     `json:"completion_audit_result_binding_ref,omitempty"`
	CompletionAuditResultRef        DisplaySafeRef                     `json:"completion_audit_result_ref,omitempty"`
	AuthorizedHostView              ProductionAdapterReadbackHostView  `json:"authorized_host_view,omitempty"`
	CompletionHandoff               ProductionAdapterCompletionHandoff `json:"completion_handoff,omitempty"`
	CompletionAuditResult           VerificationResult                 `json:"completion_audit_result,omitempty"`
	RawOutputLoaded                 bool                               `json:"raw_output_loaded"`
}

type ProductionAdapterCompletionAuditResultBinding struct {
	ContractVersion                 string             `json:"contract_version,omitempty"`
	Projected                       bool               `json:"projected"`
	Available                       bool               `json:"available"`
	Status                          VerificationStatus `json:"status,omitempty"`
	Mode                            string             `json:"mode,omitempty"`
	ReadyForObjectiveCloseout       bool               `json:"ready_for_objective_closeout"`
	ObjectiveSatisfied              bool               `json:"objective_satisfied"`
	VerificationSatisfied           bool               `json:"verification_satisfied"`
	AuthorizationBound              bool               `json:"authorization_bound"`
	CompletionAuditBound            bool               `json:"completion_audit_bound"`
	HostInvocationReported          bool               `json:"host_invocation_reported"`
	CoreInvocationExecuted          bool               `json:"core_invocation_executed"`
	DurableWriteByCore              bool               `json:"durable_write_by_core"`
	AuthorizedHostViewStatus        string             `json:"authorized_host_view_status,omitempty"`
	CompletionHandoffStatus         VerificationStatus `json:"completion_handoff_status,omitempty"`
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

func BuildProductionAdapterCompletionAuditResultBinding(input ProductionAdapterCompletionAuditResultBindingInput) ProductionAdapterCompletionAuditResultBinding {
	unsafe := input.RawOutputLoaded ||
		displaySafeRefRejected(input.CompletionAuditResultBindingRef) ||
		displaySafeRefRejected(input.CompletionAuditResultRef) ||
		productionAdapterReadbackHostViewUnsafe(input.AuthorizedHostView) ||
		productionAdapterCompletionHandoffUnsafe(input.CompletionHandoff) ||
		verificationResultUnsafe(input.CompletionAuditResult)
	hostView := input.AuthorizedHostView.Normalize()
	handoff := input.CompletionHandoff.Normalize()
	audit := input.CompletionAuditResult.Normalize()
	if unsafe {
		audit = VerificationResult{}
	}
	result := ProductionAdapterCompletionAuditResultBinding{
		ContractVersion:                 ContractVersion,
		Projected:                       true,
		Available:                       true,
		Status:                          VerificationBlocked,
		Mode:                            "production_adapter_completion_audit_result_binding",
		HostInvocationReported:          hostView.HostInvocationReported || handoff.HostInvocationReported,
		AuthorizationBound:              hostView.AuthorizationBound && handoff.AuthorizationBound,
		AuthorizedHostViewStatus:        hostView.Status,
		CompletionHandoffStatus:         handoff.Status,
		CompletionAuditResultBindingRef: normalizeOneDisplaySafeRef(input.CompletionAuditResultBindingRef),
		CompletionAuditResultRef:        normalizeOneDisplaySafeRef(input.CompletionAuditResultRef),
		InvocationReportBindingRef:      firstDisplaySafeRef(hostView.InvocationReportBindingRef, handoff.InvocationReportBindingRef),
		AuthorizationPacketRef:          firstDisplaySafeRef(hostView.AuthorizationPacketRef, handoff.AuthorizationPacketRef),
		PreflightReviewPacketRef:        firstDisplaySafeRef(hostView.PreflightReviewPacketRef, handoff.PreflightReviewPacketRef),
		AdapterRef:                      firstDisplaySafeRef(hostView.AdapterRef, handoff.AdapterRef),
		DescriptorRef:                   firstDisplaySafeRef(hostView.DescriptorRef, handoff.DescriptorRef),
		InvocationRef:                   firstDisplaySafeRef(hostView.InvocationRef, handoff.InvocationRef),
		ResultRef:                       firstDisplaySafeRef(hostView.ResultRef, handoff.ResultRef),
		ReadbackRef:                     firstDisplaySafeRef(hostView.ReadbackRef, handoff.ReadbackRef),
		CompletionHandoffRef:            firstDisplaySafeRef(hostView.CompletionHandoffRef, handoff.CompletionHandoffRef),
		EvidenceRefs:                    productionAdapterCompletionAuditResultBindingEvidenceRefs(hostView, handoff, audit, normalizeOneDisplaySafeRef(input.CompletionAuditResultRef)),
		Verification:                    audit,
		MissingInputs:                   productionAdapterCompletionAuditResultBindingMissingInputs(hostView, handoff, audit),
		BlockedReasons:                  productionAdapterCompletionAuditResultBindingBlockedReasons(hostView, handoff),
		FailureClass:                    FailureNone,
		Boundaries:                      productionAdapterCompletionAuditResultBindingBoundaries(hostView, handoff, audit),
		NextHostAction:                  firstNextHostAction(audit.NextHostAction, firstNextHostAction(handoff.NextHostAction, hostView.NextHostAction)),
		RunnerEffect:                    "none",
		PromptEffect:                    "none",
		RawOutputLoaded:                 input.RawOutputLoaded || hostView.RawOutputLoaded || handoff.RawOutputLoaded || audit.RawOutputLoaded,
	}
	if unsafe {
		result = productionAdapterCompletionAuditResultBindingBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if result.CompletionAuditResultBindingRef == "" {
		result = productionAdapterCompletionAuditResultBindingBlock(result, FailureEvidenceMissing, "completion_audit_result_binding_ref_missing", "host:completion_audit_result_binding_ref", "provide_completion_audit_result_binding")
	}
	if !hostView.Available || !hostView.ReadyForCompletionAudit || !hostView.AuthorizationBound {
		result = productionAdapterCompletionAuditResultBindingBlock(result, firstFailureClass(hostView.FailureClass, FailureAuthorizationMissing), "authorized_host_view_not_ready", "host:authorized_adapter_host_view", "review_authorized_host_view")
	}
	if !handoff.ReadyForCompletionAudit || !handoff.AuthorizationBound {
		result = productionAdapterCompletionAuditResultBindingBlock(result, firstFailureClass(handoff.FailureClass, FailureAuthorizationMissing), "completion_handoff_not_ready", "host:production_adapter_completion_handoff", firstNextHostAction(handoff.NextHostAction, "run_completion_audit"))
	}
	if result.CompletionAuditResultRef == "" {
		result = productionAdapterCompletionAuditResultBindingBlock(result, FailureEvidenceMissing, "completion_audit_result_ref_missing", "host:completion_audit_result_ref", "provide_completion_audit_result")
	}
	for _, mismatch := range productionAdapterCompletionAuditResultBindingMismatches(hostView, handoff) {
		result = productionAdapterCompletionAuditResultBindingBlock(result, FailureVerificationFailed, mismatch.reason, mismatch.missing, "review_completion_audit_binding")
	}
	if verificationResultEmpty(input.CompletionAuditResult) {
		result = productionAdapterCompletionAuditResultBindingBlock(result, FailureEvidenceMissing, "completion_audit_result_missing", "host:completion_audit_result", "provide_completion_audit_result")
	}
	if len(result.MissingInputs) > 0 || len(result.BlockedReasons) > 0 {
		return result.Normalize()
	}
	result.CompletionAuditBound = true
	result.Boundaries = AppendBoundaries(result.Boundaries, "completion_audit_result_bound")
	if len(audit.MissingInputs) > 0 {
		result.Status = productionAdapterCompletionAuditResultBindingReviewStatus(audit.Status)
		result.FailureClass = firstFailureClass(audit.FailureClass, FailureEvidenceMissing)
		result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, "completion_audit_inputs_missing")
		result.NextHostAction = firstNextHostAction(audit.NextHostAction, "provide_completion_audit_result")
		return result.Normalize()
	}
	if audit.Satisfied && audit.FailureClass != FailureNone {
		result.Status = VerificationReviewRequired
		result.FailureClass = firstFailureClass(audit.FailureClass, FailureVerificationFailed)
		result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, "completion_audit_failure_class_conflict")
		result.NextHostAction = "review_completion_audit_result"
		return result.Normalize()
	}
	if !audit.Satisfied {
		result.Status = productionAdapterCompletionAuditResultBindingReviewStatus(audit.Status)
		result.FailureClass = firstFailureClass(audit.FailureClass, FailureVerificationFailed)
		result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, "completion_audit_not_satisfied")
		result.Boundaries = AppendBoundaries(result.Boundaries, "completion_audit_result_not_satisfied")
		result.NextHostAction = firstNextHostAction(audit.NextHostAction, "review_completion_audit_result")
		return result.Normalize()
	}
	result.Status = VerificationSatisfied
	result.FailureClass = FailureNone
	result.Boundaries = AppendBoundaries(result.Boundaries, "host_completion_audit_satisfied", "ready_for_objective_closeout")
	result.NextHostAction = "review_objective_closeout"
	return result.Normalize()
}

func CloneProductionAdapterCompletionAuditResultBinding(in ProductionAdapterCompletionAuditResultBinding) ProductionAdapterCompletionAuditResultBinding {
	out := in
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.Verification = in.Verification.Clone()
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (b ProductionAdapterCompletionAuditResultBinding) Clone() ProductionAdapterCompletionAuditResultBinding {
	return CloneProductionAdapterCompletionAuditResultBinding(b)
}

func (b ProductionAdapterCompletionAuditResultBinding) Normalize() ProductionAdapterCompletionAuditResultBinding {
	out := CloneProductionAdapterCompletionAuditResultBinding(b)
	unsafe := productionAdapterCompletionAuditResultBindingUnsafe(out)
	if unsafe {
		out.Verification = VerificationResult{}
	}
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Mode = normalizeControlToken(out.Mode)
	if out.Mode == "" {
		out.Mode = "production_adapter_completion_audit_result_binding"
	}
	out.Status = NormalizeVerificationStatus(string(out.Status))
	out.AuthorizedHostViewStatus = normalizeControlToken(out.AuthorizedHostViewStatus)
	out.CompletionHandoffStatus = NormalizeVerificationStatus(string(out.CompletionHandoffStatus))
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
		out.ReadyForObjectiveCloseout = false
		out.CompletionAuditBound = false
		out.VerificationSatisfied = false
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
	out.ObjectiveSatisfied = false
	out.AuthorizationBound = out.AuthorizationBound &&
		out.InvocationReportBindingRef != "" &&
		out.AuthorizationPacketRef != "" &&
		out.PreflightReviewPacketRef != "" &&
		!containsMissingInput(out.MissingInputs, "host:authorized_adapter_host_view") &&
		!containsMissingInput(out.MissingInputs, "host:production_adapter_completion_handoff") &&
		!containsMissingInput(out.MissingInputs, "host:display_safe_refs")
	out.CompletionAuditBound = out.CompletionAuditBound &&
		out.AuthorizationBound &&
		out.CompletionAuditResultBindingRef != "" &&
		out.CompletionAuditResultRef != "" &&
		out.CompletionHandoffRef != "" &&
		!containsMissingInput(out.MissingInputs, "host:completion_audit_result") &&
		!containsMissingInput(out.MissingInputs, "host:completion_audit_result_ref") &&
		!containsMissingInput(out.MissingInputs, "host:completion_audit_result_binding_ref") &&
		!containsMissingInput(out.MissingInputs, "host:display_safe_refs") &&
		!out.RawOutputLoaded
	out.VerificationSatisfied = out.CompletionAuditBound &&
		out.Verification.Status == VerificationSatisfied &&
		out.Verification.Satisfied &&
		out.Verification.FailureClass == FailureNone &&
		len(out.Verification.MissingInputs) == 0
	if !out.CompletionAuditBound {
		out.Verification.Satisfied = false
		if out.Verification.Status == VerificationSatisfied {
			out.Verification.Status = VerificationReviewRequired
		}
	}
	out.ReadyForObjectiveCloseout = out.Available &&
		out.Status == VerificationSatisfied &&
		out.CompletionAuditBound &&
		out.VerificationSatisfied &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.RawOutputLoaded
	if !out.ReadyForObjectiveCloseout && out.Status == VerificationSatisfied {
		out.Status = VerificationReviewRequired
	}
	return out
}

func productionAdapterCompletionAuditResultBindingBlock(result ProductionAdapterCompletionAuditResultBinding, failure FailureClass, reason string, missing MissingInput, next NextHostAction) ProductionAdapterCompletionAuditResultBinding {
	result.Status = VerificationBlocked
	result.ReadyForObjectiveCloseout = false
	result.CompletionAuditBound = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	result.Boundaries = AppendBoundaries(result.Boundaries, "completion_audit_result_binding_blocked")
	return result
}

func productionAdapterCompletionAuditResultBindingReviewStatus(status VerificationStatus) VerificationStatus {
	switch NormalizeVerificationStatus(string(status)) {
	case VerificationPartial, VerificationBlocked, VerificationReviewRequired, VerificationFailed:
		return NormalizeVerificationStatus(string(status))
	default:
		return VerificationReviewRequired
	}
}

func productionAdapterCompletionAuditResultBindingEvidenceRefs(hostView ProductionAdapterReadbackHostView, handoff ProductionAdapterCompletionHandoff, audit VerificationResult, auditResultRef DisplaySafeRef) []EvidenceRef {
	evidence := append(cloneEvidenceRefs(hostView.EvidenceRefs), handoff.EvidenceRefs...)
	evidence = append(evidence, audit.EvidenceRefs...)
	if auditResultRef != "" {
		evidence = append(evidence, EvidenceRef{
			Ref:      auditResultRef,
			Kind:     "completion_audit_result",
			Strength: EvidenceStrong,
			Source:   firstDisplaySafeRef(hostView.CompletionHandoffRef, handoff.CompletionHandoffRef, hostView.InvocationRef, handoff.InvocationRef),
		})
	}
	return normalizeEvidenceRefs(evidence)
}

func productionAdapterCompletionAuditResultBindingMissingInputs(hostView ProductionAdapterReadbackHostView, handoff ProductionAdapterCompletionHandoff, audit VerificationResult) []MissingInput {
	var missing []MissingInput
	for _, input := range append(cloneMissingInputs(hostView.MissingInputs), handoff.MissingInputs...) {
		if input == "host:completion_audit_result" {
			continue
		}
		missing = AppendMissingInputs(missing, input)
	}
	missing = append(missing, audit.MissingInputs...)
	return normalizeMissingInputs(missing)
}

func productionAdapterCompletionAuditResultBindingBlockedReasons(hostView ProductionAdapterReadbackHostView, handoff ProductionAdapterCompletionHandoff) []string {
	var reasons []string
	for _, reason := range append(cloneStringSlice(hostView.BlockedReasons), handoff.BlockedReasons...) {
		if reason == "completion_audit_required" {
			continue
		}
		reasons = append(reasons, reason)
	}
	return normalizeControlTokenList(reasons)
}

func productionAdapterCompletionAuditResultBindingBoundaries(hostView ProductionAdapterReadbackHostView, handoff ProductionAdapterCompletionHandoff, audit VerificationResult) []Boundary {
	var out []Boundary
	out = append(out, hostView.Boundaries...)
	out = append(out, handoff.Boundaries...)
	out = append(out, audit.Boundaries...)
	for _, item := range []Boundary{
		"production_adapter_completion_audit_result_binding",
		"completion_audit_result_binding_projection_only",
		"host_owned_completion_audit",
		"objective_closeout_input_only",
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

type productionAdapterCompletionAuditResultBindingMismatch struct {
	reason  string
	missing MissingInput
}

func productionAdapterCompletionAuditResultBindingMismatches(hostView ProductionAdapterReadbackHostView, handoff ProductionAdapterCompletionHandoff) []productionAdapterCompletionAuditResultBindingMismatch {
	checks := []struct {
		host    DisplaySafeRef
		handoff DisplaySafeRef
		reason  string
		missing MissingInput
	}{
		{hostView.InvocationReportBindingRef, handoff.InvocationReportBindingRef, "completion_audit_invocation_report_binding_ref_mismatch", "host:invocation_report_binding"},
		{hostView.AuthorizationPacketRef, handoff.AuthorizationPacketRef, "completion_audit_authorization_packet_ref_mismatch", "host:invocation_authorization_packet"},
		{hostView.PreflightReviewPacketRef, handoff.PreflightReviewPacketRef, "completion_audit_preflight_review_packet_ref_mismatch", "host:adapter_preflight_review_packet"},
		{hostView.AdapterRef, handoff.AdapterRef, "completion_audit_adapter_ref_mismatch", "host:adapter_ref"},
		{hostView.DescriptorRef, handoff.DescriptorRef, "completion_audit_descriptor_ref_mismatch", "host:adapter_descriptor"},
		{hostView.InvocationRef, handoff.InvocationRef, "completion_audit_invocation_ref_mismatch", "host:invocation_ref"},
		{hostView.ResultRef, handoff.ResultRef, "completion_audit_result_ref_mismatch", "host:result_ref"},
		{hostView.ReadbackRef, handoff.ReadbackRef, "completion_audit_readback_ref_mismatch", "host:readback_ref"},
		{hostView.CompletionHandoffRef, handoff.CompletionHandoffRef, "completion_audit_handoff_ref_mismatch", "host:completion_handoff_ref"},
	}
	var out []productionAdapterCompletionAuditResultBindingMismatch
	for _, check := range checks {
		host := normalizeOneDisplaySafeRef(check.host)
		handoffRef := normalizeOneDisplaySafeRef(check.handoff)
		if host != "" && handoffRef != "" && host != handoffRef {
			out = append(out, productionAdapterCompletionAuditResultBindingMismatch{reason: check.reason, missing: check.missing})
		}
	}
	return out
}

func productionAdapterCompletionAuditResultBindingUnsafe(input ProductionAdapterCompletionAuditResultBinding) bool {
	return displaySafeRefRejected(input.CompletionAuditResultBindingRef) ||
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

func productionAdapterReadbackHostViewUnsafe(input ProductionAdapterReadbackHostView) bool {
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
		verificationResultUnsafe(input.Verification) ||
		input.RawOutputLoaded
}

func verificationResultUnsafe(input VerificationResult) bool {
	if input.RawOutputLoaded ||
		ContainsUnsafeRawOutput(input.FailureReason) ||
		evidenceRefsRejected(input.EvidenceRefs) {
		return true
	}
	for _, finding := range input.Findings {
		if ContainsUnsafeRawOutput(finding) {
			return true
		}
	}
	return false
}

func verificationResultEmpty(input VerificationResult) bool {
	return input.ContractVersion == "" &&
		input.Status == "" &&
		!input.Satisfied &&
		input.FailureClass == "" &&
		input.FailureReason == "" &&
		len(input.EvidenceRefs) == 0 &&
		len(input.MissingInputs) == 0 &&
		len(input.Boundaries) == 0 &&
		len(input.Findings) == 0 &&
		input.NextHostAction == "" &&
		!input.RawOutputLoaded
}
