package controlcontract

type ProductionAdapterObjectiveCloseoutDurableHandoffInput struct {
	ObjectiveCloseoutHandoffRef     DisplaySafeRef                           `json:"objective_closeout_handoff_ref,omitempty"`
	HostObjectiveLifecycleRef       DisplaySafeRef                           `json:"host_objective_lifecycle_ref,omitempty"`
	HostRunstoreRef                 DisplaySafeRef                           `json:"host_runstore_ref,omitempty"`
	ExpectedDurableEventRef         DisplaySafeRef                           `json:"expected_durable_event_ref,omitempty"`
	ExpectedObjectiveStateRef       DisplaySafeRef                           `json:"expected_objective_state_ref,omitempty"`
	HostDurableApplyConfirmationRef DisplaySafeRef                           `json:"host_durable_apply_confirmation_ref,omitempty"`
	CurrentLifecycleStage           LifecycleStage                           `json:"current_lifecycle_stage,omitempty"`
	TargetLifecycleStage            LifecycleStage                           `json:"target_lifecycle_stage,omitempty"`
	ObjectiveCloseoutPacket         ProductionAdapterObjectiveCloseoutPacket `json:"objective_closeout_packet,omitempty"`
	AdditionalMissingInputs         []MissingInput                           `json:"additional_missing_inputs,omitempty"`
	AdditionalBlockedReasons        []string                                 `json:"additional_blocked_reasons,omitempty"`
	RawOutputLoaded                 bool                                     `json:"raw_output_loaded"`
}

type ProductionAdapterObjectiveCloseoutDurableHandoff struct {
	ContractVersion                 string                    `json:"contract_version,omitempty"`
	Projected                       bool                      `json:"projected"`
	Available                       bool                      `json:"available"`
	Status                          HostActionStatus          `json:"status,omitempty"`
	Mode                            string                    `json:"mode,omitempty"`
	ReadyForHostDurableApply        bool                      `json:"ready_for_host_durable_apply"`
	ObjectiveSatisfied              bool                      `json:"objective_satisfied"`
	VerificationSatisfied           bool                      `json:"verification_satisfied"`
	HostCloseoutConfirmed           bool                      `json:"host_closeout_confirmed"`
	HostDurableApplyConfirmed       bool                      `json:"host_durable_apply_confirmed"`
	AuthorizationBound              bool                      `json:"authorization_bound"`
	CompletionAuditBound            bool                      `json:"completion_audit_bound"`
	CoreInvocationExecuted          bool                      `json:"core_invocation_executed"`
	DurableWriteByCore              bool                      `json:"durable_write_by_core"`
	ObjectiveStoreWriteByCore       bool                      `json:"objective_store_write_by_core"`
	RunstoreWriteByCore             bool                      `json:"runstore_write_by_core"`
	ObjectiveCloseoutHandoffRef     DisplaySafeRef            `json:"objective_closeout_handoff_ref,omitempty"`
	ObjectiveCloseoutPacketRef      DisplaySafeRef            `json:"objective_closeout_packet_ref,omitempty"`
	ObjectiveRef                    DisplaySafeRef            `json:"objective_ref,omitempty"`
	HostCloseoutConfirmationRef     DisplaySafeRef            `json:"host_closeout_confirmation_ref,omitempty"`
	HostObjectiveLifecycleRef       DisplaySafeRef            `json:"host_objective_lifecycle_ref,omitempty"`
	HostRunstoreRef                 DisplaySafeRef            `json:"host_runstore_ref,omitempty"`
	ExpectedDurableEventRef         DisplaySafeRef            `json:"expected_durable_event_ref,omitempty"`
	ExpectedObjectiveStateRef       DisplaySafeRef            `json:"expected_objective_state_ref,omitempty"`
	HostDurableApplyConfirmationRef DisplaySafeRef            `json:"host_durable_apply_confirmation_ref,omitempty"`
	CurrentLifecycleStage           LifecycleStage            `json:"current_lifecycle_stage,omitempty"`
	TargetLifecycleStage            LifecycleStage            `json:"target_lifecycle_stage,omitempty"`
	LifecycleTransition             LifecycleTransitionResult `json:"lifecycle_transition,omitempty"`
	CompletionAuditResultBindingRef DisplaySafeRef            `json:"completion_audit_result_binding_ref,omitempty"`
	CompletionAuditResultRef        DisplaySafeRef            `json:"completion_audit_result_ref,omitempty"`
	InvocationReportBindingRef      DisplaySafeRef            `json:"invocation_report_binding_ref,omitempty"`
	AuthorizationPacketRef          DisplaySafeRef            `json:"authorization_packet_ref,omitempty"`
	PreflightReviewPacketRef        DisplaySafeRef            `json:"preflight_review_packet_ref,omitempty"`
	AdapterRef                      DisplaySafeRef            `json:"adapter_ref,omitempty"`
	DescriptorRef                   DisplaySafeRef            `json:"descriptor_ref,omitempty"`
	InvocationRef                   DisplaySafeRef            `json:"invocation_ref,omitempty"`
	ResultRef                       DisplaySafeRef            `json:"result_ref,omitempty"`
	ReadbackRef                     DisplaySafeRef            `json:"readback_ref,omitempty"`
	CompletionHandoffRef            DisplaySafeRef            `json:"completion_handoff_ref,omitempty"`
	EvidenceRefs                    []EvidenceRef             `json:"evidence_refs,omitempty"`
	Verification                    VerificationResult        `json:"verification,omitempty"`
	MissingInputs                   []MissingInput            `json:"missing_inputs,omitempty"`
	BlockedReasons                  []string                  `json:"blocked_reasons,omitempty"`
	FailureClass                    FailureClass              `json:"failure_class,omitempty"`
	Boundaries                      []Boundary                `json:"boundaries,omitempty"`
	NextHostAction                  NextHostAction            `json:"next_host_action,omitempty"`
	RunnerEffect                    string                    `json:"runner_effect,omitempty"`
	PromptEffect                    string                    `json:"prompt_effect,omitempty"`
	RawOutputLoaded                 bool                      `json:"raw_output_loaded"`
}

type ProductionAdapterObjectiveCloseoutReadbackInput struct {
	ObjectiveCloseoutReadbackRef DisplaySafeRef                                   `json:"objective_closeout_readback_ref,omitempty"`
	DurableHandoff               ProductionAdapterObjectiveCloseoutDurableHandoff `json:"durable_handoff,omitempty"`
	AppliedDurableEventRef       DisplaySafeRef                                   `json:"applied_durable_event_ref,omitempty"`
	AppliedRunstoreRef           DisplaySafeRef                                   `json:"applied_runstore_ref,omitempty"`
	AppliedObjectiveStateRef     DisplaySafeRef                                   `json:"applied_objective_state_ref,omitempty"`
	FailureRef                   DisplaySafeRef                                   `json:"failure_ref,omitempty"`
	CompensationRef              DisplaySafeRef                                   `json:"compensation_ref,omitempty"`
	HostDurableWriteReported     bool                                             `json:"host_durable_write_reported"`
	HostDurableWriteSucceeded    bool                                             `json:"host_durable_write_succeeded"`
	HostDurableWriteFailed       bool                                             `json:"host_durable_write_failed"`
	RawOutputLoaded              bool                                             `json:"raw_output_loaded"`
}

type ProductionAdapterObjectiveCloseoutReadback struct {
	ContractVersion                   string             `json:"contract_version,omitempty"`
	Projected                         bool               `json:"projected"`
	Available                         bool               `json:"available"`
	Status                            HostActionStatus   `json:"status,omitempty"`
	Mode                              string             `json:"mode,omitempty"`
	ReadyForObjectiveCloseoutReadback bool               `json:"ready_for_objective_closeout_readback"`
	ObjectiveLifecycleClosed          bool               `json:"objective_lifecycle_closed"`
	ObjectiveSatisfied                bool               `json:"objective_satisfied"`
	VerificationSatisfied             bool               `json:"verification_satisfied"`
	HostCloseoutConfirmed             bool               `json:"host_closeout_confirmed"`
	HostDurableApplyConfirmed         bool               `json:"host_durable_apply_confirmed"`
	HostDurableWriteReported          bool               `json:"host_durable_write_reported"`
	HostDurableWriteSucceeded         bool               `json:"host_durable_write_succeeded"`
	HostDurableWriteFailed            bool               `json:"host_durable_write_failed"`
	CoreInvocationExecuted            bool               `json:"core_invocation_executed"`
	DurableWriteByCore                bool               `json:"durable_write_by_core"`
	ObjectiveStoreWriteByCore         bool               `json:"objective_store_write_by_core"`
	RunstoreWriteByCore               bool               `json:"runstore_write_by_core"`
	ObjectiveCloseoutReadbackRef      DisplaySafeRef     `json:"objective_closeout_readback_ref,omitempty"`
	ObjectiveCloseoutHandoffRef       DisplaySafeRef     `json:"objective_closeout_handoff_ref,omitempty"`
	ObjectiveCloseoutPacketRef        DisplaySafeRef     `json:"objective_closeout_packet_ref,omitempty"`
	ObjectiveRef                      DisplaySafeRef     `json:"objective_ref,omitempty"`
	HostObjectiveLifecycleRef         DisplaySafeRef     `json:"host_objective_lifecycle_ref,omitempty"`
	HostRunstoreRef                   DisplaySafeRef     `json:"host_runstore_ref,omitempty"`
	ExpectedDurableEventRef           DisplaySafeRef     `json:"expected_durable_event_ref,omitempty"`
	ExpectedObjectiveStateRef         DisplaySafeRef     `json:"expected_objective_state_ref,omitempty"`
	AppliedDurableEventRef            DisplaySafeRef     `json:"applied_durable_event_ref,omitempty"`
	AppliedRunstoreRef                DisplaySafeRef     `json:"applied_runstore_ref,omitempty"`
	AppliedObjectiveStateRef          DisplaySafeRef     `json:"applied_objective_state_ref,omitempty"`
	FailureRef                        DisplaySafeRef     `json:"failure_ref,omitempty"`
	CompensationRef                   DisplaySafeRef     `json:"compensation_ref,omitempty"`
	HostDurableApplyConfirmationRef   DisplaySafeRef     `json:"host_durable_apply_confirmation_ref,omitempty"`
	CurrentLifecycleStage             LifecycleStage     `json:"current_lifecycle_stage,omitempty"`
	TargetLifecycleStage              LifecycleStage     `json:"target_lifecycle_stage,omitempty"`
	CompletionAuditResultBindingRef   DisplaySafeRef     `json:"completion_audit_result_binding_ref,omitempty"`
	CompletionAuditResultRef          DisplaySafeRef     `json:"completion_audit_result_ref,omitempty"`
	InvocationReportBindingRef        DisplaySafeRef     `json:"invocation_report_binding_ref,omitempty"`
	AuthorizationPacketRef            DisplaySafeRef     `json:"authorization_packet_ref,omitempty"`
	PreflightReviewPacketRef          DisplaySafeRef     `json:"preflight_review_packet_ref,omitempty"`
	AdapterRef                        DisplaySafeRef     `json:"adapter_ref,omitempty"`
	DescriptorRef                     DisplaySafeRef     `json:"descriptor_ref,omitempty"`
	InvocationRef                     DisplaySafeRef     `json:"invocation_ref,omitempty"`
	ResultRef                         DisplaySafeRef     `json:"result_ref,omitempty"`
	ReadbackRef                       DisplaySafeRef     `json:"readback_ref,omitempty"`
	CompletionHandoffRef              DisplaySafeRef     `json:"completion_handoff_ref,omitempty"`
	EvidenceRefs                      []EvidenceRef      `json:"evidence_refs,omitempty"`
	Verification                      VerificationResult `json:"verification,omitempty"`
	MissingInputs                     []MissingInput     `json:"missing_inputs,omitempty"`
	BlockedReasons                    []string           `json:"blocked_reasons,omitempty"`
	FailureClass                      FailureClass       `json:"failure_class,omitempty"`
	Boundaries                        []Boundary         `json:"boundaries,omitempty"`
	NextHostAction                    NextHostAction     `json:"next_host_action,omitempty"`
	RunnerEffect                      string             `json:"runner_effect,omitempty"`
	PromptEffect                      string             `json:"prompt_effect,omitempty"`
	RawOutputLoaded                   bool               `json:"raw_output_loaded"`
}

func BuildProductionAdapterObjectiveCloseoutDurableHandoff(input ProductionAdapterObjectiveCloseoutDurableHandoffInput) ProductionAdapterObjectiveCloseoutDurableHandoff {
	packet := input.ObjectiveCloseoutPacket.Normalize()
	from, to := productionAdapterObjectiveCloseoutLifecycleStages(input.CurrentLifecycleStage, input.TargetLifecycleStage)
	transition := checkLifecycleTransition(from, to)
	unsafe := input.RawOutputLoaded ||
		displaySafeRefRejected(input.ObjectiveCloseoutHandoffRef) ||
		displaySafeRefRejected(input.HostObjectiveLifecycleRef) ||
		displaySafeRefRejected(input.HostRunstoreRef) ||
		displaySafeRefRejected(input.ExpectedDurableEventRef) ||
		displaySafeRefRejected(input.ExpectedObjectiveStateRef) ||
		displaySafeRefRejected(input.HostDurableApplyConfirmationRef) ||
		productionAdapterObjectiveCloseoutPacketUnsafe(packet)
	verification := packet.Verification.Normalize()
	if unsafe {
		verification = VerificationResult{}
	}
	result := ProductionAdapterObjectiveCloseoutDurableHandoff{
		ContractVersion:                 ContractVersion,
		Projected:                       true,
		Available:                       true,
		Status:                          HostActionBlocked,
		Mode:                            "production_adapter_objective_closeout_durable_handoff",
		ObjectiveSatisfied:              packet.ObjectiveSatisfied,
		VerificationSatisfied:           packet.VerificationSatisfied,
		HostCloseoutConfirmed:           packet.HostCloseoutConfirmed,
		AuthorizationBound:              packet.AuthorizationBound,
		CompletionAuditBound:            packet.CompletionAuditBound,
		ObjectiveCloseoutHandoffRef:     normalizeOneDisplaySafeRef(input.ObjectiveCloseoutHandoffRef),
		ObjectiveCloseoutPacketRef:      packet.ObjectiveCloseoutPacketRef,
		ObjectiveRef:                    packet.ObjectiveRef,
		HostCloseoutConfirmationRef:     packet.HostCloseoutConfirmationRef,
		HostObjectiveLifecycleRef:       normalizeOneDisplaySafeRef(input.HostObjectiveLifecycleRef),
		HostRunstoreRef:                 normalizeOneDisplaySafeRef(input.HostRunstoreRef),
		ExpectedDurableEventRef:         normalizeOneDisplaySafeRef(input.ExpectedDurableEventRef),
		ExpectedObjectiveStateRef:       normalizeOneDisplaySafeRef(input.ExpectedObjectiveStateRef),
		HostDurableApplyConfirmationRef: normalizeOneDisplaySafeRef(input.HostDurableApplyConfirmationRef),
		CurrentLifecycleStage:           from,
		TargetLifecycleStage:            to,
		LifecycleTransition:             transition,
		CompletionAuditResultBindingRef: packet.CompletionAuditResultBindingRef,
		CompletionAuditResultRef:        packet.CompletionAuditResultRef,
		InvocationReportBindingRef:      packet.InvocationReportBindingRef,
		AuthorizationPacketRef:          packet.AuthorizationPacketRef,
		PreflightReviewPacketRef:        packet.PreflightReviewPacketRef,
		AdapterRef:                      packet.AdapterRef,
		DescriptorRef:                   packet.DescriptorRef,
		InvocationRef:                   packet.InvocationRef,
		ResultRef:                       packet.ResultRef,
		ReadbackRef:                     packet.ReadbackRef,
		CompletionHandoffRef:            packet.CompletionHandoffRef,
		EvidenceRefs:                    productionAdapterObjectiveCloseoutDurableHandoffEvidenceRefs(packet, normalizeOneDisplaySafeRef(input.ObjectiveCloseoutHandoffRef), normalizeOneDisplaySafeRef(input.HostObjectiveLifecycleRef), normalizeOneDisplaySafeRef(input.HostRunstoreRef), normalizeOneDisplaySafeRef(input.HostDurableApplyConfirmationRef)),
		Verification:                    verification,
		MissingInputs:                   appendProductionAdapterCloseoutAdditionalMissing(cloneMissingInputs(packet.MissingInputs), input.AdditionalMissingInputs),
		BlockedReasons:                  appendProductionAdapterCloseoutAdditionalBlocked(cloneStringSlice(packet.BlockedReasons), input.AdditionalBlockedReasons),
		FailureClass:                    firstFailureClass(packet.FailureClass, verification.FailureClass),
		Boundaries:                      productionAdapterObjectiveCloseoutDurableHandoffBoundaries(packet.Boundaries, transition.Boundaries),
		NextHostAction:                  firstNextHostAction(packet.NextHostAction, "review_objective_closeout"),
		RunnerEffect:                    "none",
		PromptEffect:                    "none",
		RawOutputLoaded:                 input.RawOutputLoaded || packet.RawOutputLoaded || verification.RawOutputLoaded,
	}
	if unsafe {
		result = productionAdapterObjectiveCloseoutDurableHandoffBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if !packet.ReadyForObjectiveCompletion || !packet.ObjectiveSatisfied || !packet.HostCloseoutConfirmed || !packet.VerificationSatisfied {
		result = productionAdapterObjectiveCloseoutDurableHandoffBlock(result, firstFailureClass(packet.FailureClass, FailureEvidenceMissing), "objective_closeout_packet_not_ready", "host:objective_closeout_packet", firstNextHostAction(packet.NextHostAction, "review_objective_closeout"))
	}
	if result.ObjectiveCloseoutHandoffRef == "" {
		result = productionAdapterObjectiveCloseoutDurableHandoffBlock(result, FailureEvidenceMissing, "objective_closeout_handoff_ref_missing", "host:objective_closeout_handoff_ref", "provide_objective_closeout_handoff")
	}
	if result.HostObjectiveLifecycleRef == "" {
		result = productionAdapterObjectiveCloseoutDurableHandoffBlock(result, FailureEvidenceMissing, "host_objective_lifecycle_ref_missing", "host:objective_lifecycle_ref", "provide_objective_lifecycle_ref")
	}
	if result.HostRunstoreRef == "" {
		result = productionAdapterObjectiveCloseoutDurableHandoffBlock(result, FailureEvidenceMissing, "host_runstore_ref_missing", "host:runstore_ref", "provide_runstore_ref")
	}
	if result.ExpectedDurableEventRef == "" {
		result = productionAdapterObjectiveCloseoutDurableHandoffBlock(result, FailureEvidenceMissing, "expected_durable_event_ref_missing", "host:expected_durable_event_ref", "provide_expected_durable_event_ref")
	}
	if result.ExpectedObjectiveStateRef == "" {
		result = productionAdapterObjectiveCloseoutDurableHandoffBlock(result, FailureEvidenceMissing, "expected_objective_state_ref_missing", "host:expected_objective_state_ref", "provide_expected_objective_state_ref")
	}
	if !transition.Allowed {
		for _, missing := range transition.MissingInputs {
			result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
		}
		result = productionAdapterObjectiveCloseoutDurableHandoffBlock(result, firstFailureClass(transition.FailureClass, FailureVerificationFailed), "objective_closeout_lifecycle_transition_not_ready", "host:lifecycle_transition_review", firstNextHostAction(transition.NextHostAction, "review_lifecycle_transition"))
	}
	if len(result.MissingInputs) > 0 || len(result.BlockedReasons) > 0 {
		return result.Normalize()
	}
	if result.HostDurableApplyConfirmationRef == "" {
		result.Status = HostActionReviewRequired
		result.FailureClass = FailureApprovalRequired
		result.MissingInputs = AppendMissingInputs(result.MissingInputs, "host:objective_closeout_durable_apply_confirmation_ref")
		result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, "objective_closeout_durable_apply_confirmation_required")
		result.Boundaries = AppendBoundaries(result.Boundaries, "host_objective_closeout_durable_apply_confirmation_required")
		result.NextHostAction = "confirm_objective_closeout_durable_apply"
		return result.Normalize()
	}
	result.Status = HostActionReady
	result.FailureClass = FailureNone
	result.HostDurableApplyConfirmed = true
	result.Boundaries = AppendBoundaries(result.Boundaries, "host_objective_closeout_durable_apply_confirmed", "ready_for_host_objective_closeout_durable_apply")
	result.NextHostAction = "host_may_apply_objective_closeout"
	return result.Normalize()
}

func BuildProductionAdapterObjectiveCloseoutReadback(input ProductionAdapterObjectiveCloseoutReadbackInput) ProductionAdapterObjectiveCloseoutReadback {
	handoff := input.DurableHandoff.Normalize()
	unsafe := input.RawOutputLoaded ||
		displaySafeRefRejected(input.ObjectiveCloseoutReadbackRef) ||
		displaySafeRefRejected(input.AppliedDurableEventRef) ||
		displaySafeRefRejected(input.AppliedRunstoreRef) ||
		displaySafeRefRejected(input.AppliedObjectiveStateRef) ||
		displaySafeRefRejected(input.FailureRef) ||
		displaySafeRefRejected(input.CompensationRef) ||
		productionAdapterObjectiveCloseoutDurableHandoffUnsafe(handoff)
	verification := handoff.Verification.Normalize()
	if unsafe {
		verification = VerificationResult{}
	}
	result := ProductionAdapterObjectiveCloseoutReadback{
		ContractVersion:                 ContractVersion,
		Projected:                       true,
		Available:                       true,
		Status:                          HostActionBlocked,
		Mode:                            "production_adapter_objective_closeout_readback",
		ObjectiveSatisfied:              handoff.ObjectiveSatisfied,
		VerificationSatisfied:           handoff.VerificationSatisfied,
		HostCloseoutConfirmed:           handoff.HostCloseoutConfirmed,
		HostDurableApplyConfirmed:       handoff.HostDurableApplyConfirmed,
		HostDurableWriteReported:        input.HostDurableWriteReported,
		HostDurableWriteSucceeded:       input.HostDurableWriteSucceeded,
		HostDurableWriteFailed:          input.HostDurableWriteFailed,
		ObjectiveCloseoutReadbackRef:    normalizeOneDisplaySafeRef(input.ObjectiveCloseoutReadbackRef),
		ObjectiveCloseoutHandoffRef:     handoff.ObjectiveCloseoutHandoffRef,
		ObjectiveCloseoutPacketRef:      handoff.ObjectiveCloseoutPacketRef,
		ObjectiveRef:                    handoff.ObjectiveRef,
		HostObjectiveLifecycleRef:       handoff.HostObjectiveLifecycleRef,
		HostRunstoreRef:                 handoff.HostRunstoreRef,
		ExpectedDurableEventRef:         handoff.ExpectedDurableEventRef,
		ExpectedObjectiveStateRef:       handoff.ExpectedObjectiveStateRef,
		AppliedDurableEventRef:          normalizeOneDisplaySafeRef(input.AppliedDurableEventRef),
		AppliedRunstoreRef:              normalizeOneDisplaySafeRef(input.AppliedRunstoreRef),
		AppliedObjectiveStateRef:        normalizeOneDisplaySafeRef(input.AppliedObjectiveStateRef),
		FailureRef:                      normalizeOneDisplaySafeRef(input.FailureRef),
		CompensationRef:                 normalizeOneDisplaySafeRef(input.CompensationRef),
		HostDurableApplyConfirmationRef: handoff.HostDurableApplyConfirmationRef,
		CurrentLifecycleStage:           handoff.CurrentLifecycleStage,
		TargetLifecycleStage:            handoff.TargetLifecycleStage,
		CompletionAuditResultBindingRef: handoff.CompletionAuditResultBindingRef,
		CompletionAuditResultRef:        handoff.CompletionAuditResultRef,
		InvocationReportBindingRef:      handoff.InvocationReportBindingRef,
		AuthorizationPacketRef:          handoff.AuthorizationPacketRef,
		PreflightReviewPacketRef:        handoff.PreflightReviewPacketRef,
		AdapterRef:                      handoff.AdapterRef,
		DescriptorRef:                   handoff.DescriptorRef,
		InvocationRef:                   handoff.InvocationRef,
		ResultRef:                       handoff.ResultRef,
		ReadbackRef:                     handoff.ReadbackRef,
		CompletionHandoffRef:            handoff.CompletionHandoffRef,
		EvidenceRefs:                    productionAdapterObjectiveCloseoutReadbackEvidenceRefs(handoff, normalizeOneDisplaySafeRef(input.ObjectiveCloseoutReadbackRef), normalizeOneDisplaySafeRef(input.AppliedDurableEventRef), normalizeOneDisplaySafeRef(input.AppliedRunstoreRef), normalizeOneDisplaySafeRef(input.AppliedObjectiveStateRef), normalizeOneDisplaySafeRef(input.FailureRef), normalizeOneDisplaySafeRef(input.CompensationRef)),
		Verification:                    verification,
		MissingInputs:                   cloneMissingInputs(handoff.MissingInputs),
		BlockedReasons:                  cloneStringSlice(handoff.BlockedReasons),
		FailureClass:                    firstFailureClass(handoff.FailureClass, verification.FailureClass),
		Boundaries:                      productionAdapterObjectiveCloseoutReadbackBoundaries(handoff.Boundaries),
		NextHostAction:                  firstNextHostAction(handoff.NextHostAction, "host_may_apply_objective_closeout"),
		RunnerEffect:                    "none",
		PromptEffect:                    "none",
		RawOutputLoaded:                 input.RawOutputLoaded || handoff.RawOutputLoaded || verification.RawOutputLoaded,
	}
	if unsafe {
		result = productionAdapterObjectiveCloseoutReadbackBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if !handoff.ReadyForHostDurableApply || !handoff.HostDurableApplyConfirmed {
		result = productionAdapterObjectiveCloseoutReadbackBlock(result, firstFailureClass(handoff.FailureClass, FailureEvidenceMissing), "objective_closeout_durable_handoff_not_ready", "host:objective_closeout_durable_handoff", firstNextHostAction(handoff.NextHostAction, "review_objective_closeout_durable_handoff"))
	}
	if result.ObjectiveCloseoutReadbackRef == "" {
		result = productionAdapterObjectiveCloseoutReadbackBlock(result, FailureEvidenceMissing, "objective_closeout_readback_ref_missing", "host:objective_closeout_readback_ref", "provide_objective_closeout_readback")
	}
	if !input.HostDurableWriteReported {
		result = productionAdapterObjectiveCloseoutReadbackBlock(result, FailureEvidenceMissing, "objective_closeout_durable_write_not_reported", "host:objective_closeout_durable_readback", "provide_objective_closeout_durable_readback")
	}
	if input.HostDurableWriteSucceeded && input.HostDurableWriteFailed {
		result = productionAdapterObjectiveCloseoutReadbackBlock(result, FailureVerificationFailed, "objective_closeout_durable_write_conflict", "host:objective_closeout_durable_readback_review", "review_objective_closeout_durable_readback")
	}
	if len(result.MissingInputs) > 0 || len(result.BlockedReasons) > 0 {
		return result.Normalize()
	}
	if input.HostDurableWriteFailed {
		if result.FailureRef == "" {
			result = productionAdapterObjectiveCloseoutReadbackBlock(result, FailureEvidenceMissing, "objective_closeout_failure_ref_missing", "host:objective_closeout_failure_ref", "provide_objective_closeout_failure_ref")
			return result.Normalize()
		}
		result.Status = HostActionReviewRequired
		result.FailureClass = FailureVerificationFailed
		result.Boundaries = AppendBoundaries(result.Boundaries, "objective_closeout_durable_write_failed")
		result.NextHostAction = "review_objective_closeout_durable_failure"
		return result.Normalize()
	}
	if !input.HostDurableWriteSucceeded {
		result = productionAdapterObjectiveCloseoutReadbackBlock(result, FailureEvidenceMissing, "objective_closeout_durable_write_status_missing", "host:objective_closeout_durable_write_status", "provide_objective_closeout_durable_readback")
		return result.Normalize()
	}
	for _, mismatch := range productionAdapterObjectiveCloseoutReadbackMismatches(result) {
		result = productionAdapterObjectiveCloseoutReadbackBlock(result, FailureVerificationFailed, mismatch.reason, mismatch.missing, "review_objective_closeout_durable_readback")
	}
	if result.AppliedDurableEventRef == "" {
		result = productionAdapterObjectiveCloseoutReadbackBlock(result, FailureEvidenceMissing, "applied_durable_event_ref_missing", "host:applied_durable_event_ref", "provide_objective_closeout_durable_readback")
	}
	if result.AppliedRunstoreRef == "" {
		result = productionAdapterObjectiveCloseoutReadbackBlock(result, FailureEvidenceMissing, "applied_runstore_ref_missing", "host:applied_runstore_ref", "provide_objective_closeout_durable_readback")
	}
	if result.AppliedObjectiveStateRef == "" {
		result = productionAdapterObjectiveCloseoutReadbackBlock(result, FailureEvidenceMissing, "applied_objective_state_ref_missing", "host:applied_objective_state_ref", "provide_objective_closeout_durable_readback")
	}
	if len(result.MissingInputs) > 0 || len(result.BlockedReasons) > 0 {
		return result.Normalize()
	}
	result.Status = HostActionRecorded
	result.FailureClass = FailureNone
	result.Boundaries = AppendBoundaries(result.Boundaries, "host_objective_closeout_durable_write_recorded", "objective_lifecycle_closeout_readback_recorded", "objective_lifecycle_closed_by_host")
	result.NextHostAction = "return_objective_closed_lifecycle"
	return result.Normalize()
}

func CloneProductionAdapterObjectiveCloseoutDurableHandoff(in ProductionAdapterObjectiveCloseoutDurableHandoff) ProductionAdapterObjectiveCloseoutDurableHandoff {
	out := in
	out.LifecycleTransition = in.LifecycleTransition.Normalize()
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.Verification = in.Verification.Clone()
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (h ProductionAdapterObjectiveCloseoutDurableHandoff) Clone() ProductionAdapterObjectiveCloseoutDurableHandoff {
	return CloneProductionAdapterObjectiveCloseoutDurableHandoff(h)
}

func (h ProductionAdapterObjectiveCloseoutDurableHandoff) Normalize() ProductionAdapterObjectiveCloseoutDurableHandoff {
	out := CloneProductionAdapterObjectiveCloseoutDurableHandoff(h)
	unsafe := productionAdapterObjectiveCloseoutDurableHandoffRefsUnsafe(out)
	if unsafe {
		out.Verification = VerificationResult{}
	}
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Mode = normalizeControlToken(out.Mode)
	if out.Mode == "" {
		out.Mode = "production_adapter_objective_closeout_durable_handoff"
	}
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.ObjectiveCloseoutHandoffRef = normalizeOneDisplaySafeRef(out.ObjectiveCloseoutHandoffRef)
	out.ObjectiveCloseoutPacketRef = normalizeOneDisplaySafeRef(out.ObjectiveCloseoutPacketRef)
	out.ObjectiveRef = normalizeOneDisplaySafeRef(out.ObjectiveRef)
	out.HostCloseoutConfirmationRef = normalizeOneDisplaySafeRef(out.HostCloseoutConfirmationRef)
	out.HostObjectiveLifecycleRef = normalizeOneDisplaySafeRef(out.HostObjectiveLifecycleRef)
	out.HostRunstoreRef = normalizeOneDisplaySafeRef(out.HostRunstoreRef)
	out.ExpectedDurableEventRef = normalizeOneDisplaySafeRef(out.ExpectedDurableEventRef)
	out.ExpectedObjectiveStateRef = normalizeOneDisplaySafeRef(out.ExpectedObjectiveStateRef)
	out.HostDurableApplyConfirmationRef = normalizeOneDisplaySafeRef(out.HostDurableApplyConfirmationRef)
	out.CurrentLifecycleStage = NormalizeLifecycleStage(string(out.CurrentLifecycleStage))
	out.TargetLifecycleStage = NormalizeLifecycleStage(string(out.TargetLifecycleStage))
	out.LifecycleTransition = out.LifecycleTransition.Normalize()
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
		out.Status = HostActionBlocked
		out.ReadyForHostDurableApply = false
		out.HostDurableApplyConfirmed = false
	}
	if unsafe || out.RawOutputLoaded {
		out.RawOutputLoaded = true
		out.Status = HostActionBlocked
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
	out.ObjectiveStoreWriteByCore = false
	out.RunstoreWriteByCore = false
	out.HostDurableApplyConfirmed = out.HostDurableApplyConfirmed &&
		out.HostDurableApplyConfirmationRef != "" &&
		!containsMissingInput(out.MissingInputs, "host:objective_closeout_durable_apply_confirmation_ref") &&
		!out.RawOutputLoaded
	out.VerificationSatisfied = out.VerificationSatisfied &&
		out.Verification.Status == VerificationSatisfied &&
		out.Verification.Satisfied &&
		out.Verification.FailureClass == FailureNone &&
		len(out.Verification.MissingInputs) == 0
	out.ObjectiveSatisfied = out.ObjectiveSatisfied &&
		out.VerificationSatisfied &&
		out.HostCloseoutConfirmed &&
		!containsMissingInput(out.MissingInputs, "host:objective_closeout_packet") &&
		!out.RawOutputLoaded
	out.ReadyForHostDurableApply = out.Status == HostActionReady &&
		out.Available &&
		out.ObjectiveSatisfied &&
		out.HostDurableApplyConfirmed &&
		out.ObjectiveCloseoutHandoffRef != "" &&
		out.ObjectiveCloseoutPacketRef != "" &&
		out.ObjectiveRef != "" &&
		out.HostObjectiveLifecycleRef != "" &&
		out.HostRunstoreRef != "" &&
		out.ExpectedDurableEventRef != "" &&
		out.ExpectedObjectiveStateRef != "" &&
		out.LifecycleTransition.Allowed &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.RawOutputLoaded
	if !out.ReadyForHostDurableApply && out.Status == HostActionReady {
		out.Status = HostActionReviewRequired
	}
	return out
}

func CloneProductionAdapterObjectiveCloseoutReadback(in ProductionAdapterObjectiveCloseoutReadback) ProductionAdapterObjectiveCloseoutReadback {
	out := in
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.Verification = in.Verification.Clone()
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (r ProductionAdapterObjectiveCloseoutReadback) Clone() ProductionAdapterObjectiveCloseoutReadback {
	return CloneProductionAdapterObjectiveCloseoutReadback(r)
}

func (r ProductionAdapterObjectiveCloseoutReadback) Normalize() ProductionAdapterObjectiveCloseoutReadback {
	out := CloneProductionAdapterObjectiveCloseoutReadback(r)
	unsafe := productionAdapterObjectiveCloseoutReadbackRefsUnsafe(out)
	if unsafe {
		out.Verification = VerificationResult{}
	}
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Mode = normalizeControlToken(out.Mode)
	if out.Mode == "" {
		out.Mode = "production_adapter_objective_closeout_readback"
	}
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.ObjectiveCloseoutReadbackRef = normalizeOneDisplaySafeRef(out.ObjectiveCloseoutReadbackRef)
	out.ObjectiveCloseoutHandoffRef = normalizeOneDisplaySafeRef(out.ObjectiveCloseoutHandoffRef)
	out.ObjectiveCloseoutPacketRef = normalizeOneDisplaySafeRef(out.ObjectiveCloseoutPacketRef)
	out.ObjectiveRef = normalizeOneDisplaySafeRef(out.ObjectiveRef)
	out.HostObjectiveLifecycleRef = normalizeOneDisplaySafeRef(out.HostObjectiveLifecycleRef)
	out.HostRunstoreRef = normalizeOneDisplaySafeRef(out.HostRunstoreRef)
	out.ExpectedDurableEventRef = normalizeOneDisplaySafeRef(out.ExpectedDurableEventRef)
	out.ExpectedObjectiveStateRef = normalizeOneDisplaySafeRef(out.ExpectedObjectiveStateRef)
	out.AppliedDurableEventRef = normalizeOneDisplaySafeRef(out.AppliedDurableEventRef)
	out.AppliedRunstoreRef = normalizeOneDisplaySafeRef(out.AppliedRunstoreRef)
	out.AppliedObjectiveStateRef = normalizeOneDisplaySafeRef(out.AppliedObjectiveStateRef)
	out.FailureRef = normalizeOneDisplaySafeRef(out.FailureRef)
	out.CompensationRef = normalizeOneDisplaySafeRef(out.CompensationRef)
	out.HostDurableApplyConfirmationRef = normalizeOneDisplaySafeRef(out.HostDurableApplyConfirmationRef)
	out.CurrentLifecycleStage = NormalizeLifecycleStage(string(out.CurrentLifecycleStage))
	out.TargetLifecycleStage = NormalizeLifecycleStage(string(out.TargetLifecycleStage))
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
		out.Status = HostActionBlocked
		out.ReadyForObjectiveCloseoutReadback = false
		out.ObjectiveLifecycleClosed = false
	}
	if unsafe || out.RawOutputLoaded {
		out.RawOutputLoaded = true
		out.Status = HostActionBlocked
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
	out.ObjectiveStoreWriteByCore = false
	out.RunstoreWriteByCore = false
	out.ReadyForObjectiveCloseoutReadback = out.Status == HostActionRecorded &&
		out.HostDurableWriteReported &&
		out.HostDurableWriteSucceeded &&
		!out.HostDurableWriteFailed &&
		out.ObjectiveCloseoutReadbackRef != "" &&
		out.ObjectiveCloseoutHandoffRef != "" &&
		out.AppliedDurableEventRef != "" &&
		out.AppliedRunstoreRef != "" &&
		out.AppliedObjectiveStateRef != "" &&
		out.AppliedDurableEventRef == out.ExpectedDurableEventRef &&
		out.AppliedRunstoreRef == out.HostRunstoreRef &&
		out.AppliedObjectiveStateRef == out.ExpectedObjectiveStateRef &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.RawOutputLoaded
	out.ObjectiveLifecycleClosed = out.ReadyForObjectiveCloseoutReadback
	out.ObjectiveSatisfied = out.ReadyForObjectiveCloseoutReadback && out.VerificationSatisfied && out.HostCloseoutConfirmed
	if !out.ReadyForObjectiveCloseoutReadback && out.Status == HostActionRecorded {
		out.Status = HostActionReviewRequired
	}
	return out
}

func productionAdapterObjectiveCloseoutDurableHandoffBlock(result ProductionAdapterObjectiveCloseoutDurableHandoff, failure FailureClass, reason string, missing MissingInput, next NextHostAction) ProductionAdapterObjectiveCloseoutDurableHandoff {
	result.Status = HostActionBlocked
	result.ReadyForHostDurableApply = false
	result.HostDurableApplyConfirmed = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	result.Boundaries = AppendBoundaries(result.Boundaries, "objective_closeout_durable_handoff_blocked")
	return result
}

func productionAdapterObjectiveCloseoutReadbackBlock(result ProductionAdapterObjectiveCloseoutReadback, failure FailureClass, reason string, missing MissingInput, next NextHostAction) ProductionAdapterObjectiveCloseoutReadback {
	result.Status = HostActionBlocked
	result.ReadyForObjectiveCloseoutReadback = false
	result.ObjectiveLifecycleClosed = false
	result.ObjectiveSatisfied = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	result.Boundaries = AppendBoundaries(result.Boundaries, "objective_closeout_readback_blocked")
	return result
}

func productionAdapterObjectiveCloseoutLifecycleStages(from LifecycleStage, to LifecycleStage) (LifecycleStage, LifecycleStage) {
	normalizedFrom := NormalizeLifecycleStage(string(from))
	normalizedTo := NormalizeLifecycleStage(string(to))
	if normalizedFrom == LifecycleStageNotReady {
		normalizedFrom = LifecycleStageAudit
	}
	if normalizedTo == LifecycleStageNotReady {
		normalizedTo = LifecycleStageClosed
	}
	return normalizedFrom, normalizedTo
}

func productionAdapterObjectiveCloseoutDurableHandoffEvidenceRefs(packet ProductionAdapterObjectiveCloseoutPacket, handoffRef DisplaySafeRef, lifecycleRef DisplaySafeRef, runstoreRef DisplaySafeRef, confirmationRef DisplaySafeRef) []EvidenceRef {
	evidence := cloneEvidenceRefs(packet.EvidenceRefs)
	source := firstDisplaySafeRef(packet.ObjectiveCloseoutPacketRef, packet.ObjectiveRef, packet.CompletionAuditResultBindingRef)
	for _, item := range []struct {
		ref      DisplaySafeRef
		kind     string
		strength EvidenceStrength
	}{
		{handoffRef, "objective_closeout_durable_handoff", EvidenceAdequate},
		{lifecycleRef, "host_objective_lifecycle", EvidenceAdequate},
		{runstoreRef, "host_runstore", EvidenceAdequate},
		{confirmationRef, "host_objective_closeout_durable_apply_confirmation", EvidenceStrong},
	} {
		if item.ref != "" {
			evidence = append(evidence, EvidenceRef{
				Ref:      item.ref,
				Kind:     item.kind,
				Strength: item.strength,
				Source:   source,
			})
		}
	}
	return normalizeEvidenceRefs(evidence)
}

func productionAdapterObjectiveCloseoutReadbackEvidenceRefs(handoff ProductionAdapterObjectiveCloseoutDurableHandoff, readbackRef DisplaySafeRef, durableEventRef DisplaySafeRef, runstoreRef DisplaySafeRef, objectiveStateRef DisplaySafeRef, failureRef DisplaySafeRef, compensationRef DisplaySafeRef) []EvidenceRef {
	evidence := cloneEvidenceRefs(handoff.EvidenceRefs)
	source := firstDisplaySafeRef(handoff.ObjectiveCloseoutHandoffRef, handoff.ObjectiveCloseoutPacketRef, handoff.ObjectiveRef)
	for _, item := range []struct {
		ref      DisplaySafeRef
		kind     string
		strength EvidenceStrength
	}{
		{readbackRef, "objective_closeout_readback", EvidenceAdequate},
		{durableEventRef, "objective_closeout_durable_event", EvidenceStrong},
		{runstoreRef, "objective_closeout_runstore_readback", EvidenceStrong},
		{objectiveStateRef, "objective_state_readback", EvidenceStrong},
		{failureRef, "objective_closeout_failure", EvidenceAdequate},
		{compensationRef, "objective_closeout_compensation", EvidenceAdequate},
	} {
		if item.ref != "" {
			evidence = append(evidence, EvidenceRef{
				Ref:      item.ref,
				Kind:     item.kind,
				Strength: item.strength,
				Source:   source,
			})
		}
	}
	return normalizeEvidenceRefs(evidence)
}

func productionAdapterObjectiveCloseoutDurableHandoffBoundaries(packetBoundaries []Boundary, transitionBoundaries []Boundary) []Boundary {
	out := cloneBoundaries(packetBoundaries)
	out = append(out, transitionBoundaries...)
	for _, item := range []Boundary{
		"production_adapter_objective_closeout_durable_handoff",
		"objective_closeout_durable_handoff_projection_only",
		"host_owned_objective_closeout_durable_handoff",
		"host_owned_durable_closeout",
		"display_safe_refs_only",
		"display_safe_result_refs_only",
		"no_runner_dispatch",
		"no_durable_write_by_core",
		"no_objective_store_write_by_core",
		"no_runstore_write_by_core",
	} {
		out = AppendBoundaries(out, item)
	}
	return normalizeBoundaries(out)
}

func productionAdapterObjectiveCloseoutReadbackBoundaries(handoffBoundaries []Boundary) []Boundary {
	out := cloneBoundaries(handoffBoundaries)
	for _, item := range []Boundary{
		"production_adapter_objective_closeout_readback",
		"objective_closeout_readback_projection_only",
		"host_owned_objective_closeout_readback",
		"host_owned_durable_closeout_readback",
		"display_safe_refs_only",
		"display_safe_result_refs_only",
		"no_runner_dispatch",
		"no_durable_write_by_core",
		"no_objective_store_write_by_core",
		"no_runstore_write_by_core",
	} {
		out = AppendBoundaries(out, item)
	}
	return normalizeBoundaries(out)
}

type productionAdapterObjectiveCloseoutReadbackMismatch struct {
	reason  string
	missing MissingInput
}

func productionAdapterObjectiveCloseoutReadbackMismatches(result ProductionAdapterObjectiveCloseoutReadback) []productionAdapterObjectiveCloseoutReadbackMismatch {
	checks := []struct {
		expected DisplaySafeRef
		applied  DisplaySafeRef
		reason   string
		missing  MissingInput
	}{
		{result.ExpectedDurableEventRef, result.AppliedDurableEventRef, "objective_closeout_durable_event_ref_mismatch", "host:durable_event_ref"},
		{result.HostRunstoreRef, result.AppliedRunstoreRef, "objective_closeout_runstore_ref_mismatch", "host:runstore_ref"},
		{result.ExpectedObjectiveStateRef, result.AppliedObjectiveStateRef, "objective_closeout_state_ref_mismatch", "host:objective_state_ref"},
	}
	var out []productionAdapterObjectiveCloseoutReadbackMismatch
	for _, check := range checks {
		expected := normalizeOneDisplaySafeRef(check.expected)
		applied := normalizeOneDisplaySafeRef(check.applied)
		if expected != "" && applied != "" && expected != applied {
			out = append(out, productionAdapterObjectiveCloseoutReadbackMismatch{reason: check.reason, missing: check.missing})
		}
	}
	return out
}

func appendProductionAdapterCloseoutAdditionalMissing(base []MissingInput, additional []MissingInput) []MissingInput {
	out := cloneMissingInputs(base)
	for _, missing := range additional {
		out = AppendMissingInputs(out, missing)
	}
	return normalizeMissingInputs(out)
}

func appendProductionAdapterCloseoutAdditionalBlocked(base []string, additional []string) []string {
	out := cloneStringSlice(base)
	for _, reason := range additional {
		out = appendUniqueControlToken(out, reason)
	}
	return normalizeControlTokenList(out)
}

func productionAdapterObjectiveCloseoutDurableHandoffUnsafe(input ProductionAdapterObjectiveCloseoutDurableHandoff) bool {
	return productionAdapterObjectiveCloseoutDurableHandoffRefsUnsafe(input) ||
		evidenceRefsRejected(input.EvidenceRefs) ||
		verificationResultUnsafe(input.Verification) ||
		input.RawOutputLoaded
}

func productionAdapterObjectiveCloseoutDurableHandoffRefsUnsafe(input ProductionAdapterObjectiveCloseoutDurableHandoff) bool {
	return displaySafeRefRejected(input.ObjectiveCloseoutHandoffRef) ||
		displaySafeRefRejected(input.ObjectiveCloseoutPacketRef) ||
		displaySafeRefRejected(input.ObjectiveRef) ||
		displaySafeRefRejected(input.HostCloseoutConfirmationRef) ||
		displaySafeRefRejected(input.HostObjectiveLifecycleRef) ||
		displaySafeRefRejected(input.HostRunstoreRef) ||
		displaySafeRefRejected(input.ExpectedDurableEventRef) ||
		displaySafeRefRejected(input.ExpectedObjectiveStateRef) ||
		displaySafeRefRejected(input.HostDurableApplyConfirmationRef) ||
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
		displaySafeRefRejected(input.CompletionHandoffRef)
}

func productionAdapterObjectiveCloseoutReadbackUnsafe(input ProductionAdapterObjectiveCloseoutReadback) bool {
	return productionAdapterObjectiveCloseoutReadbackRefsUnsafe(input) ||
		evidenceRefsRejected(input.EvidenceRefs) ||
		verificationResultUnsafe(input.Verification) ||
		input.RawOutputLoaded
}

func productionAdapterObjectiveCloseoutReadbackRefsUnsafe(input ProductionAdapterObjectiveCloseoutReadback) bool {
	return displaySafeRefRejected(input.ObjectiveCloseoutReadbackRef) ||
		displaySafeRefRejected(input.ObjectiveCloseoutHandoffRef) ||
		displaySafeRefRejected(input.ObjectiveCloseoutPacketRef) ||
		displaySafeRefRejected(input.ObjectiveRef) ||
		displaySafeRefRejected(input.HostObjectiveLifecycleRef) ||
		displaySafeRefRejected(input.HostRunstoreRef) ||
		displaySafeRefRejected(input.ExpectedDurableEventRef) ||
		displaySafeRefRejected(input.ExpectedObjectiveStateRef) ||
		displaySafeRefRejected(input.AppliedDurableEventRef) ||
		displaySafeRefRejected(input.AppliedRunstoreRef) ||
		displaySafeRefRejected(input.AppliedObjectiveStateRef) ||
		displaySafeRefRejected(input.FailureRef) ||
		displaySafeRefRejected(input.CompensationRef) ||
		displaySafeRefRejected(input.HostDurableApplyConfirmationRef) ||
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
		displaySafeRefRejected(input.CompletionHandoffRef)
}
