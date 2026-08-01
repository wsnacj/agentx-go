package controlcontract

type ProductionAdapterObjectiveCloseoutHostViewInput struct {
	HostViewRef             DisplaySafeRef                                   `json:"host_view_ref,omitempty"`
	ObjectiveCloseoutPacket ProductionAdapterObjectiveCloseoutPacket         `json:"objective_closeout_packet,omitempty"`
	DurableHandoff          ProductionAdapterObjectiveCloseoutDurableHandoff `json:"durable_handoff,omitempty"`
	Readback                ProductionAdapterObjectiveCloseoutReadback       `json:"readback,omitempty"`
	RawOutputLoaded         bool                                             `json:"raw_output_loaded"`
}

type ProductionAdapterObjectiveCloseoutHostView struct {
	ContractVersion                          string             `json:"contract_version,omitempty"`
	Projected                                bool               `json:"projected"`
	Available                                bool               `json:"available"`
	Status                                   string             `json:"status,omitempty"`
	Mode                                     string             `json:"mode,omitempty"`
	DisplayState                             string             `json:"display_state,omitempty"`
	ReadyForHostDisplay                      bool               `json:"ready_for_host_display"`
	ReadyForHostDurableApply                 bool               `json:"ready_for_host_durable_apply"`
	ReadyForFailureReview                    bool               `json:"ready_for_failure_review"`
	ReadyForCompensationReview               bool               `json:"ready_for_compensation_review"`
	ReadyForObjectiveReturn                  bool               `json:"ready_for_objective_return"`
	IntermediateDisplay                      bool               `json:"intermediate_display"`
	FinalDisplay                             bool               `json:"final_display"`
	FailureReviewDisplay                     bool               `json:"failure_review_display"`
	ObjectiveLifecycleClosed                 bool               `json:"objective_lifecycle_closed"`
	ObjectiveSatisfied                       bool               `json:"objective_satisfied"`
	VerificationSatisfied                    bool               `json:"verification_satisfied"`
	HostCloseoutConfirmed                    bool               `json:"host_closeout_confirmed"`
	HostDurableApplyConfirmed                bool               `json:"host_durable_apply_confirmed"`
	HostDurableWriteReported                 bool               `json:"host_durable_write_reported"`
	HostDurableWriteSucceeded                bool               `json:"host_durable_write_succeeded"`
	HostDurableWriteFailed                   bool               `json:"host_durable_write_failed"`
	CoreInvocationExecuted                   bool               `json:"core_invocation_executed"`
	DurableWriteByCore                       bool               `json:"durable_write_by_core"`
	ObjectiveStoreWriteByCore                bool               `json:"objective_store_write_by_core"`
	RunstoreWriteByCore                      bool               `json:"runstore_write_by_core"`
	HostViewRef                              DisplaySafeRef     `json:"host_view_ref,omitempty"`
	DisplaySections                          []string           `json:"display_sections,omitempty"`
	ObjectiveCloseoutPacketStatus            VerificationStatus `json:"objective_closeout_packet_status,omitempty"`
	DurableHandoffStatus                     HostActionStatus   `json:"durable_handoff_status,omitempty"`
	ReadbackStatus                           HostActionStatus   `json:"readback_status,omitempty"`
	ObjectiveCloseoutPacketRef               DisplaySafeRef     `json:"objective_closeout_packet_ref,omitempty"`
	ObjectiveCloseoutHandoffRef              DisplaySafeRef     `json:"objective_closeout_handoff_ref,omitempty"`
	ObjectiveCloseoutReadbackRef             DisplaySafeRef     `json:"objective_closeout_readback_ref,omitempty"`
	ObjectiveRef                             DisplaySafeRef     `json:"objective_ref,omitempty"`
	AuthoritativeObjectiveCloseoutPacketRef  DisplaySafeRef     `json:"authoritative_objective_closeout_packet_ref,omitempty"`
	AuthoritativeObjectiveCloseoutHandoffRef DisplaySafeRef     `json:"authoritative_objective_closeout_handoff_ref,omitempty"`
	AuthoritativeObjectiveRef                DisplaySafeRef     `json:"authoritative_objective_ref,omitempty"`
	AuthoritativeHostRunstoreRef             DisplaySafeRef     `json:"authoritative_host_runstore_ref,omitempty"`
	AuthoritativeDurableEventRef             DisplaySafeRef     `json:"authoritative_durable_event_ref,omitempty"`
	AuthoritativeObjectiveStateRef           DisplaySafeRef     `json:"authoritative_objective_state_ref,omitempty"`
	ObservedObjectiveCloseoutPacketRef       DisplaySafeRef     `json:"observed_objective_closeout_packet_ref,omitempty"`
	ObservedObjectiveCloseoutHandoffRef      DisplaySafeRef     `json:"observed_objective_closeout_handoff_ref,omitempty"`
	ObservedObjectiveRef                     DisplaySafeRef     `json:"observed_objective_ref,omitempty"`
	ObservedHostRunstoreRef                  DisplaySafeRef     `json:"observed_host_runstore_ref,omitempty"`
	ObservedAppliedDurableEventRef           DisplaySafeRef     `json:"observed_applied_durable_event_ref,omitempty"`
	ObservedAppliedRunstoreRef               DisplaySafeRef     `json:"observed_applied_runstore_ref,omitempty"`
	ObservedAppliedObjectiveStateRef         DisplaySafeRef     `json:"observed_applied_objective_state_ref,omitempty"`
	HostCloseoutConfirmationRef              DisplaySafeRef     `json:"host_closeout_confirmation_ref,omitempty"`
	HostObjectiveLifecycleRef                DisplaySafeRef     `json:"host_objective_lifecycle_ref,omitempty"`
	HostRunstoreRef                          DisplaySafeRef     `json:"host_runstore_ref,omitempty"`
	ExpectedDurableEventRef                  DisplaySafeRef     `json:"expected_durable_event_ref,omitempty"`
	ExpectedObjectiveStateRef                DisplaySafeRef     `json:"expected_objective_state_ref,omitempty"`
	AppliedDurableEventRef                   DisplaySafeRef     `json:"applied_durable_event_ref,omitempty"`
	AppliedRunstoreRef                       DisplaySafeRef     `json:"applied_runstore_ref,omitempty"`
	AppliedObjectiveStateRef                 DisplaySafeRef     `json:"applied_objective_state_ref,omitempty"`
	FailureRef                               DisplaySafeRef     `json:"failure_ref,omitempty"`
	CompensationRef                          DisplaySafeRef     `json:"compensation_ref,omitempty"`
	HostDurableApplyConfirmationRef          DisplaySafeRef     `json:"host_durable_apply_confirmation_ref,omitempty"`
	CompletionAuditResultBindingRef          DisplaySafeRef     `json:"completion_audit_result_binding_ref,omitempty"`
	CompletionAuditResultRef                 DisplaySafeRef     `json:"completion_audit_result_ref,omitempty"`
	InvocationReportBindingRef               DisplaySafeRef     `json:"invocation_report_binding_ref,omitempty"`
	AuthorizationPacketRef                   DisplaySafeRef     `json:"authorization_packet_ref,omitempty"`
	PreflightReviewPacketRef                 DisplaySafeRef     `json:"preflight_review_packet_ref,omitempty"`
	AdapterRef                               DisplaySafeRef     `json:"adapter_ref,omitempty"`
	DescriptorRef                            DisplaySafeRef     `json:"descriptor_ref,omitempty"`
	InvocationRef                            DisplaySafeRef     `json:"invocation_ref,omitempty"`
	ResultRef                                DisplaySafeRef     `json:"result_ref,omitempty"`
	ReadbackRef                              DisplaySafeRef     `json:"readback_ref,omitempty"`
	CompletionHandoffRef                     DisplaySafeRef     `json:"completion_handoff_ref,omitempty"`
	EvidenceRefs                             []EvidenceRef      `json:"evidence_refs,omitempty"`
	MissingInputs                            []MissingInput     `json:"missing_inputs,omitempty"`
	BlockedReasons                           []string           `json:"blocked_reasons,omitempty"`
	DisplayProgressMissingInputs             []MissingInput     `json:"display_progress_missing_inputs,omitempty"`
	DisplayProgressBlockedReasons            []string           `json:"display_progress_blocked_reasons,omitempty"`
	FailureClass                             FailureClass       `json:"failure_class,omitempty"`
	Boundaries                               []Boundary         `json:"boundaries,omitempty"`
	NextHostAction                           NextHostAction     `json:"next_host_action,omitempty"`
	RunnerEffect                             string             `json:"runner_effect,omitempty"`
	PromptEffect                             string             `json:"prompt_effect,omitempty"`
	RawOutputLoaded                          bool               `json:"raw_output_loaded"`
}

// agentx-api: internal_candidate
type ProductionAdapterObjectiveCloseoutBlackboxFixtureInput struct {
	FixtureRef      DisplaySafeRef                             `json:"fixture_ref,omitempty"`
	HostView        ProductionAdapterObjectiveCloseoutHostView `json:"host_view,omitempty"`
	RawOutputLoaded bool                                       `json:"raw_output_loaded"`
}

// agentx-api: internal_candidate
type ProductionAdapterObjectiveCloseoutBlackboxFixture struct {
	ContractVersion                          string         `json:"contract_version,omitempty"`
	Projected                                bool           `json:"projected"`
	Available                                bool           `json:"available"`
	Status                                   string         `json:"status,omitempty"`
	Mode                                     string         `json:"mode,omitempty"`
	DisplayState                             string         `json:"display_state,omitempty"`
	ReadyForHostDisplay                      bool           `json:"ready_for_host_display"`
	ReadyForObjectiveReturn                  bool           `json:"ready_for_objective_return"`
	IntermediateDisplay                      bool           `json:"intermediate_display"`
	FinalDisplay                             bool           `json:"final_display"`
	ObjectiveLifecycleClosed                 bool           `json:"objective_lifecycle_closed"`
	ObjectiveSatisfied                       bool           `json:"objective_satisfied"`
	CoreInvocationExecuted                   bool           `json:"core_invocation_executed"`
	DurableWriteByCore                       bool           `json:"durable_write_by_core"`
	ObjectiveStoreWriteByCore                bool           `json:"objective_store_write_by_core"`
	RunstoreWriteByCore                      bool           `json:"runstore_write_by_core"`
	FixtureRef                               DisplaySafeRef `json:"fixture_ref,omitempty"`
	HostViewRef                              DisplaySafeRef `json:"host_view_ref,omitempty"`
	DisplaySections                          []string       `json:"display_sections,omitempty"`
	ObjectiveCloseoutPacketRef               DisplaySafeRef `json:"objective_closeout_packet_ref,omitempty"`
	ObjectiveCloseoutHandoffRef              DisplaySafeRef `json:"objective_closeout_handoff_ref,omitempty"`
	ObjectiveCloseoutReadbackRef             DisplaySafeRef `json:"objective_closeout_readback_ref,omitempty"`
	ObjectiveRef                             DisplaySafeRef `json:"objective_ref,omitempty"`
	AuthoritativeObjectiveCloseoutPacketRef  DisplaySafeRef `json:"authoritative_objective_closeout_packet_ref,omitempty"`
	AuthoritativeObjectiveCloseoutHandoffRef DisplaySafeRef `json:"authoritative_objective_closeout_handoff_ref,omitempty"`
	AuthoritativeObjectiveRef                DisplaySafeRef `json:"authoritative_objective_ref,omitempty"`
	AuthoritativeHostRunstoreRef             DisplaySafeRef `json:"authoritative_host_runstore_ref,omitempty"`
	AuthoritativeDurableEventRef             DisplaySafeRef `json:"authoritative_durable_event_ref,omitempty"`
	AuthoritativeObjectiveStateRef           DisplaySafeRef `json:"authoritative_objective_state_ref,omitempty"`
	ObservedObjectiveCloseoutPacketRef       DisplaySafeRef `json:"observed_objective_closeout_packet_ref,omitempty"`
	ObservedObjectiveCloseoutHandoffRef      DisplaySafeRef `json:"observed_objective_closeout_handoff_ref,omitempty"`
	ObservedObjectiveRef                     DisplaySafeRef `json:"observed_objective_ref,omitempty"`
	ObservedHostRunstoreRef                  DisplaySafeRef `json:"observed_host_runstore_ref,omitempty"`
	ObservedAppliedDurableEventRef           DisplaySafeRef `json:"observed_applied_durable_event_ref,omitempty"`
	ObservedAppliedRunstoreRef               DisplaySafeRef `json:"observed_applied_runstore_ref,omitempty"`
	ObservedAppliedObjectiveStateRef         DisplaySafeRef `json:"observed_applied_objective_state_ref,omitempty"`
	HostObjectiveLifecycleRef                DisplaySafeRef `json:"host_objective_lifecycle_ref,omitempty"`
	HostRunstoreRef                          DisplaySafeRef `json:"host_runstore_ref,omitempty"`
	AppliedDurableEventRef                   DisplaySafeRef `json:"applied_durable_event_ref,omitempty"`
	AppliedObjectiveStateRef                 DisplaySafeRef `json:"applied_objective_state_ref,omitempty"`
	CompletionAuditResultBindingRef          DisplaySafeRef `json:"completion_audit_result_binding_ref,omitempty"`
	AdapterRef                               DisplaySafeRef `json:"adapter_ref,omitempty"`
	MissingInputs                            []MissingInput `json:"missing_inputs,omitempty"`
	BlockedReasons                           []string       `json:"blocked_reasons,omitempty"`
	DisplayProgressMissingInputs             []MissingInput `json:"display_progress_missing_inputs,omitempty"`
	DisplayProgressBlockedReasons            []string       `json:"display_progress_blocked_reasons,omitempty"`
	FailureClass                             FailureClass   `json:"failure_class,omitempty"`
	Boundaries                               []Boundary     `json:"boundaries,omitempty"`
	NextHostAction                           NextHostAction `json:"next_host_action,omitempty"`
	RunnerEffect                             string         `json:"runner_effect,omitempty"`
	PromptEffect                             string         `json:"prompt_effect,omitempty"`
	RawOutputLoaded                          bool           `json:"raw_output_loaded"`
}

func BuildProductionAdapterObjectiveCloseoutHostView(input ProductionAdapterObjectiveCloseoutHostViewInput) ProductionAdapterObjectiveCloseoutHostView {
	if productionAdapterObjectiveCloseoutPacketEmpty(input.ObjectiveCloseoutPacket) &&
		productionAdapterObjectiveCloseoutDurableHandoffEmpty(input.DurableHandoff) &&
		productionAdapterObjectiveCloseoutReadbackEmpty(input.Readback) {
		return unavailableProductionAdapterObjectiveCloseoutHostView()
	}
	packet := input.ObjectiveCloseoutPacket.Normalize()
	handoffProvided := !productionAdapterObjectiveCloseoutDurableHandoffEmpty(input.DurableHandoff)
	readbackProvided := !productionAdapterObjectiveCloseoutReadbackEmpty(input.Readback)
	handoff := input.DurableHandoff.Normalize()
	readback := input.Readback.Normalize()
	if !handoffProvided {
		handoff = ProductionAdapterObjectiveCloseoutDurableHandoff{}
	}
	if !readbackProvided {
		readback = ProductionAdapterObjectiveCloseoutReadback{}
	}
	result := ProductionAdapterObjectiveCloseoutHostView{
		ContractVersion:                          ContractVersion,
		Projected:                                true,
		Available:                                true,
		Status:                                   "blocked",
		Mode:                                     "production_adapter_objective_closeout_host_view",
		DisplayState:                             "blocked",
		HostViewRef:                              normalizeOneDisplaySafeRef(input.HostViewRef),
		DisplaySections:                          productionAdapterObjectiveCloseoutDisplaySections(),
		ObjectiveCloseoutPacketStatus:            packet.Status,
		DurableHandoffStatus:                     handoff.Status,
		ReadbackStatus:                           readback.Status,
		ObjectiveLifecycleClosed:                 readback.ObjectiveLifecycleClosed,
		ObjectiveSatisfied:                       firstBool(readback.ObjectiveSatisfied, handoff.ObjectiveSatisfied, packet.ObjectiveSatisfied),
		VerificationSatisfied:                    firstBool(readback.VerificationSatisfied, handoff.VerificationSatisfied, packet.VerificationSatisfied),
		HostCloseoutConfirmed:                    firstBool(readback.HostCloseoutConfirmed, handoff.HostCloseoutConfirmed, packet.HostCloseoutConfirmed),
		HostDurableApplyConfirmed:                firstBool(readback.HostDurableApplyConfirmed, handoff.HostDurableApplyConfirmed),
		HostDurableWriteReported:                 readback.HostDurableWriteReported,
		HostDurableWriteSucceeded:                readback.HostDurableWriteSucceeded,
		HostDurableWriteFailed:                   readback.HostDurableWriteFailed,
		ObjectiveCloseoutPacketRef:               firstDisplaySafeRef(handoff.ObjectiveCloseoutPacketRef, packet.ObjectiveCloseoutPacketRef, readback.ObjectiveCloseoutPacketRef),
		ObjectiveCloseoutHandoffRef:              firstDisplaySafeRef(handoff.ObjectiveCloseoutHandoffRef, readback.ObjectiveCloseoutHandoffRef),
		ObjectiveCloseoutReadbackRef:             readback.ObjectiveCloseoutReadbackRef,
		ObjectiveRef:                             firstDisplaySafeRef(handoff.ObjectiveRef, packet.ObjectiveRef, readback.ObjectiveRef),
		AuthoritativeObjectiveCloseoutPacketRef:  firstDisplaySafeRef(handoff.ObjectiveCloseoutPacketRef, packet.ObjectiveCloseoutPacketRef),
		AuthoritativeObjectiveCloseoutHandoffRef: handoff.ObjectiveCloseoutHandoffRef,
		AuthoritativeObjectiveRef:                firstDisplaySafeRef(handoff.ObjectiveRef, packet.ObjectiveRef),
		AuthoritativeHostRunstoreRef:             handoff.HostRunstoreRef,
		AuthoritativeDurableEventRef:             handoff.ExpectedDurableEventRef,
		AuthoritativeObjectiveStateRef:           handoff.ExpectedObjectiveStateRef,
		ObservedObjectiveCloseoutPacketRef:       readback.ObjectiveCloseoutPacketRef,
		ObservedObjectiveCloseoutHandoffRef:      readback.ObjectiveCloseoutHandoffRef,
		ObservedObjectiveRef:                     readback.ObjectiveRef,
		ObservedHostRunstoreRef:                  readback.HostRunstoreRef,
		ObservedAppliedDurableEventRef:           readback.AppliedDurableEventRef,
		ObservedAppliedRunstoreRef:               readback.AppliedRunstoreRef,
		ObservedAppliedObjectiveStateRef:         readback.AppliedObjectiveStateRef,
		HostCloseoutConfirmationRef:              firstDisplaySafeRef(handoff.HostCloseoutConfirmationRef, packet.HostCloseoutConfirmationRef),
		HostObjectiveLifecycleRef:                firstDisplaySafeRef(handoff.HostObjectiveLifecycleRef, readback.HostObjectiveLifecycleRef),
		HostRunstoreRef:                          firstDisplaySafeRef(handoff.HostRunstoreRef, readback.HostRunstoreRef),
		ExpectedDurableEventRef:                  firstDisplaySafeRef(handoff.ExpectedDurableEventRef, readback.ExpectedDurableEventRef),
		ExpectedObjectiveStateRef:                firstDisplaySafeRef(handoff.ExpectedObjectiveStateRef, readback.ExpectedObjectiveStateRef),
		AppliedDurableEventRef:                   readback.AppliedDurableEventRef,
		AppliedRunstoreRef:                       readback.AppliedRunstoreRef,
		AppliedObjectiveStateRef:                 readback.AppliedObjectiveStateRef,
		FailureRef:                               readback.FailureRef,
		CompensationRef:                          readback.CompensationRef,
		HostDurableApplyConfirmationRef:          firstDisplaySafeRef(readback.HostDurableApplyConfirmationRef, handoff.HostDurableApplyConfirmationRef),
		CompletionAuditResultBindingRef:          firstDisplaySafeRef(readback.CompletionAuditResultBindingRef, handoff.CompletionAuditResultBindingRef, packet.CompletionAuditResultBindingRef),
		CompletionAuditResultRef:                 firstDisplaySafeRef(readback.CompletionAuditResultRef, handoff.CompletionAuditResultRef, packet.CompletionAuditResultRef),
		InvocationReportBindingRef:               firstDisplaySafeRef(readback.InvocationReportBindingRef, handoff.InvocationReportBindingRef, packet.InvocationReportBindingRef),
		AuthorizationPacketRef:                   firstDisplaySafeRef(readback.AuthorizationPacketRef, handoff.AuthorizationPacketRef, packet.AuthorizationPacketRef),
		PreflightReviewPacketRef:                 firstDisplaySafeRef(readback.PreflightReviewPacketRef, handoff.PreflightReviewPacketRef, packet.PreflightReviewPacketRef),
		AdapterRef:                               firstDisplaySafeRef(readback.AdapterRef, handoff.AdapterRef, packet.AdapterRef),
		DescriptorRef:                            firstDisplaySafeRef(readback.DescriptorRef, handoff.DescriptorRef, packet.DescriptorRef),
		InvocationRef:                            firstDisplaySafeRef(readback.InvocationRef, handoff.InvocationRef, packet.InvocationRef),
		ResultRef:                                firstDisplaySafeRef(readback.ResultRef, handoff.ResultRef, packet.ResultRef),
		ReadbackRef:                              firstDisplaySafeRef(readback.ReadbackRef, handoff.ReadbackRef, packet.ReadbackRef),
		CompletionHandoffRef:                     firstDisplaySafeRef(readback.CompletionHandoffRef, handoff.CompletionHandoffRef, packet.CompletionHandoffRef),
		EvidenceRefs:                             productionAdapterObjectiveCloseoutHostViewEvidenceRefs(packet, handoff, readback),
		MissingInputs:                            productionAdapterObjectiveCloseoutHostViewMissingInputs(packet, handoff, readback, handoffProvided, readbackProvided),
		BlockedReasons:                           productionAdapterObjectiveCloseoutHostViewBlockedReasons(packet, handoff, readback, handoffProvided, readbackProvided),
		DisplayProgressMissingInputs:             productionAdapterObjectiveCloseoutHostViewDisplayProgressMissingInputs(handoff, handoffProvided, readbackProvided),
		DisplayProgressBlockedReasons:            productionAdapterObjectiveCloseoutHostViewDisplayProgressBlockedReasons(handoff, handoffProvided, readbackProvided),
		FailureClass:                             firstFailureClass(firstFailureClass(packet.FailureClass, handoff.FailureClass), readback.FailureClass),
		Boundaries:                               productionAdapterObjectiveCloseoutHostViewBoundaries(packet.Boundaries, handoff.Boundaries, readback.Boundaries),
		NextHostAction:                           firstNextHostAction(firstNextHostAction(readback.NextHostAction, handoff.NextHostAction), packet.NextHostAction),
		RunnerEffect:                             "none",
		PromptEffect:                             "none",
		RawOutputLoaded:                          input.RawOutputLoaded || packet.RawOutputLoaded || handoff.RawOutputLoaded || readback.RawOutputLoaded,
	}
	if productionAdapterObjectiveCloseoutHostViewUnsafe(input, packet, handoff, readback, handoffProvided, readbackProvided) {
		result = productionAdapterObjectiveCloseoutHostViewBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if result.HostViewRef == "" {
		result = productionAdapterObjectiveCloseoutHostViewBlock(result, FailureEvidenceMissing, "objective_closeout_host_view_ref_missing", "host:objective_closeout_host_view_ref", "provide_objective_closeout_host_view")
	}
	if !packet.ReadyForObjectiveCompletion || !packet.ObjectiveSatisfied {
		result = productionAdapterObjectiveCloseoutHostViewBlock(result, firstFailureClass(packet.FailureClass, FailureEvidenceMissing), "objective_closeout_packet_not_ready", "host:objective_closeout_packet", firstNextHostAction(packet.NextHostAction, "review_objective_closeout"))
	}
	if !handoffProvided {
		result = productionAdapterObjectiveCloseoutHostViewBlock(result, FailureEvidenceMissing, "objective_closeout_durable_handoff_not_ready", "host:objective_closeout_durable_handoff", "review_objective_closeout_durable_handoff")
	} else if !handoff.ReadyForHostDurableApply {
		result = productionAdapterObjectiveCloseoutHostViewBlock(result, firstFailureClass(handoff.FailureClass, FailureEvidenceMissing), "objective_closeout_durable_handoff_not_ready", "host:objective_closeout_durable_handoff", firstNextHostAction(handoff.NextHostAction, "review_objective_closeout_durable_handoff"))
	}
	for _, mismatch := range productionAdapterObjectiveCloseoutHostViewMismatches(packet, handoff, readback, handoffProvided, readbackProvided) {
		result = productionAdapterObjectiveCloseoutHostViewBlock(result, FailureVerificationFailed, mismatch.reason, mismatch.missing, "review_objective_closeout_host_view")
	}
	if result.HostViewRef != "" && packet.ReadyForObjectiveCompletion && !result.RawOutputLoaded {
		result.ReadyForHostDisplay = true
	}
	switch {
	case readbackProvided && readback.HostDurableWriteFailed && len(result.MissingInputs) == 0 && len(result.BlockedReasons) == 0:
		result.Status = "objective_closeout_durable_failure_review"
		result.DisplayState = "durable_failure_review"
		result.ReadyForFailureReview = true
		result.ReadyForCompensationReview = readback.CompensationRef != ""
		result.ReadyForHostDurableApply = false
		result.ReadyForObjectiveReturn = false
		result.IntermediateDisplay = true
		result.FinalDisplay = false
		result.FailureReviewDisplay = true
		result.ObjectiveLifecycleClosed = false
		result.ObjectiveSatisfied = false
		result.FailureClass = firstFailureClass(result.FailureClass, FailureVerificationFailed)
		result.Boundaries = AppendBoundaries(result.Boundaries, "objective_closeout_durable_failure_review", "failure_compensation_review_only", "compensation_not_executed")
		result.NextHostAction = "review_objective_closeout_durable_failure"
	case readbackProvided && readback.ReadyForObjectiveCloseoutReadback && readback.ObjectiveLifecycleClosed && len(result.MissingInputs) == 0 && len(result.BlockedReasons) == 0:
		result.Status = "objective_lifecycle_closed"
		result.DisplayState = "objective_return_final"
		result.ReadyForObjectiveReturn = true
		result.IntermediateDisplay = false
		result.FinalDisplay = true
		result.FailureReviewDisplay = false
		result.ObjectiveLifecycleClosed = true
		result.ObjectiveSatisfied = true
		result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_objective_closeout_display", "ready_for_objective_return", "objective_lifecycle_closed_by_host")
		result.NextHostAction = "return_objective_closed_lifecycle"
	case handoffProvided && handoff.ReadyForHostDurableApply:
		if len(result.BlockedReasons) == 0 || onlyDisplayProgressBlocks(result.BlockedReasons, "objective_closeout_readback_not_ready") {
			result.Status = "ready_for_host_durable_apply"
			result.DisplayState = "durable_apply_pending"
			result.ReadyForHostDurableApply = true
			result.IntermediateDisplay = true
			result.FinalDisplay = false
			result.FailureReviewDisplay = false
			result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_objective_closeout_display", "ready_for_host_objective_closeout_durable_apply")
			result.NextHostAction = firstNextHostAction(handoff.NextHostAction, "host_may_apply_objective_closeout")
		}
	default:
		if result.ReadyForHostDisplay && closeoutHostViewControlTokenContains(result.BlockedReasons, "objective_closeout_durable_handoff_not_ready") {
			result.Status = "review_required"
			result.DisplayState = "closeout_review_required"
			result.FailureClass = firstFailureClass(result.FailureClass, FailureEvidenceMissing)
			result.NextHostAction = "review_objective_closeout_durable_handoff"
		}
		if result.ReadyForHostDisplay && len(result.BlockedReasons) == 0 {
			result.Status = "ready_for_objective_closeout_display"
			result.DisplayState = "closeout_review_pending"
			result.IntermediateDisplay = true
			result.NextHostAction = "provide_objective_closeout_durable_handoff"
		}
	}
	return result.Normalize()
}

// agentx-api: internal_candidate
func BuildProductionAdapterObjectiveCloseoutBlackboxFixture(input ProductionAdapterObjectiveCloseoutBlackboxFixtureInput) ProductionAdapterObjectiveCloseoutBlackboxFixture {
	if productionAdapterObjectiveCloseoutHostViewEmpty(input.HostView) {
		return unavailableProductionAdapterObjectiveCloseoutBlackboxFixture()
	}
	view := input.HostView.Normalize()
	result := ProductionAdapterObjectiveCloseoutBlackboxFixture{
		ContractVersion:                          ContractVersion,
		Projected:                                true,
		Available:                                view.Available,
		Status:                                   "blocked",
		Mode:                                     "production_adapter_objective_closeout_blackbox_fixture",
		DisplayState:                             view.DisplayState,
		FixtureRef:                               normalizeOneDisplaySafeRef(input.FixtureRef),
		HostViewRef:                              view.HostViewRef,
		DisplaySections:                          append(productionAdapterObjectiveCloseoutDisplaySections(), "objective_closeout_blackbox_assertions"),
		ObjectiveLifecycleClosed:                 view.ObjectiveLifecycleClosed,
		ObjectiveSatisfied:                       view.ObjectiveSatisfied,
		ObjectiveCloseoutPacketRef:               view.ObjectiveCloseoutPacketRef,
		ObjectiveCloseoutHandoffRef:              view.ObjectiveCloseoutHandoffRef,
		ObjectiveCloseoutReadbackRef:             view.ObjectiveCloseoutReadbackRef,
		ObjectiveRef:                             view.ObjectiveRef,
		AuthoritativeObjectiveCloseoutPacketRef:  view.AuthoritativeObjectiveCloseoutPacketRef,
		AuthoritativeObjectiveCloseoutHandoffRef: view.AuthoritativeObjectiveCloseoutHandoffRef,
		AuthoritativeObjectiveRef:                view.AuthoritativeObjectiveRef,
		AuthoritativeHostRunstoreRef:             view.AuthoritativeHostRunstoreRef,
		AuthoritativeDurableEventRef:             view.AuthoritativeDurableEventRef,
		AuthoritativeObjectiveStateRef:           view.AuthoritativeObjectiveStateRef,
		ObservedObjectiveCloseoutPacketRef:       view.ObservedObjectiveCloseoutPacketRef,
		ObservedObjectiveCloseoutHandoffRef:      view.ObservedObjectiveCloseoutHandoffRef,
		ObservedObjectiveRef:                     view.ObservedObjectiveRef,
		ObservedHostRunstoreRef:                  view.ObservedHostRunstoreRef,
		ObservedAppliedDurableEventRef:           view.ObservedAppliedDurableEventRef,
		ObservedAppliedRunstoreRef:               view.ObservedAppliedRunstoreRef,
		ObservedAppliedObjectiveStateRef:         view.ObservedAppliedObjectiveStateRef,
		HostObjectiveLifecycleRef:                view.HostObjectiveLifecycleRef,
		HostRunstoreRef:                          view.HostRunstoreRef,
		AppliedDurableEventRef:                   view.AppliedDurableEventRef,
		AppliedObjectiveStateRef:                 view.AppliedObjectiveStateRef,
		CompletionAuditResultBindingRef:          view.CompletionAuditResultBindingRef,
		AdapterRef:                               view.AdapterRef,
		MissingInputs:                            productionAdapterObjectiveCloseoutBlackboxFixtureMissingInputs(view),
		BlockedReasons:                           productionAdapterObjectiveCloseoutBlackboxFixtureBlockedReasons(view),
		DisplayProgressMissingInputs:             cloneMissingInputs(view.DisplayProgressMissingInputs),
		DisplayProgressBlockedReasons:            cloneStringSlice(view.DisplayProgressBlockedReasons),
		FailureClass:                             view.FailureClass,
		Boundaries:                               productionAdapterObjectiveCloseoutBlackboxFixtureBoundaries(view.Boundaries),
		NextHostAction:                           view.NextHostAction,
		RunnerEffect:                             "none",
		PromptEffect:                             "none",
		RawOutputLoaded:                          input.RawOutputLoaded || view.RawOutputLoaded,
	}
	if input.RawOutputLoaded || displaySafeRefRejected(input.FixtureRef) || productionAdapterObjectiveCloseoutHostViewUnsafeOutput(view) {
		result = productionAdapterObjectiveCloseoutBlackboxFixtureBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if result.FixtureRef == "" {
		result = productionAdapterObjectiveCloseoutBlackboxFixtureBlock(result, FailureEvidenceMissing, "objective_closeout_fixture_ref_missing", "host:objective_closeout_fixture_ref", "provide_objective_closeout_fixture")
	}
	if !view.ReadyForHostDisplay {
		result = productionAdapterObjectiveCloseoutBlackboxFixtureBlock(result, firstFailureClass(view.FailureClass, FailureEvidenceMissing), "objective_closeout_host_view_not_ready", "host:objective_closeout_host_view", firstNextHostAction(view.NextHostAction, "review_objective_closeout_host_view"))
	}
	if len(result.MissingInputs) == 0 && len(result.BlockedReasons) == 0 {
		result.ReadyForHostDisplay = true
		if view.ReadyForObjectiveReturn {
			result.Status = "ready_for_objective_return"
			result.DisplayState = "objective_return_final"
			result.ReadyForObjectiveReturn = true
			result.IntermediateDisplay = false
			result.FinalDisplay = true
			result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_objective_return", "host_cli_objective_closeout_display_ready")
			result.NextHostAction = "return_objective_closed_lifecycle"
		} else if view.ReadyForHostDurableApply {
			result.Status = "ready_for_host_durable_apply_display"
			result.DisplayState = "durable_apply_pending"
			result.IntermediateDisplay = true
			result.FinalDisplay = false
			result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_host_durable_apply_display", "host_cli_objective_closeout_intermediate_display_ready")
			result.NextHostAction = firstNextHostAction(view.NextHostAction, "host_may_apply_objective_closeout")
		} else {
			result.Status = "ready_for_objective_closeout_display"
			result.DisplayState = firstControlToken(view.DisplayState, "closeout_review_pending")
			result.IntermediateDisplay = true
			result.FinalDisplay = false
			result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_objective_closeout_display", "host_cli_objective_closeout_display_ready")
			result.NextHostAction = firstNextHostAction(view.NextHostAction, "review_objective_closeout_display")
		}
	}
	return result.Normalize()
}

func CloneProductionAdapterObjectiveCloseoutHostView(in ProductionAdapterObjectiveCloseoutHostView) ProductionAdapterObjectiveCloseoutHostView {
	out := in
	out.DisplaySections = cloneStringSlice(in.DisplaySections)
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.DisplayProgressMissingInputs = cloneMissingInputs(in.DisplayProgressMissingInputs)
	out.DisplayProgressBlockedReasons = cloneStringSlice(in.DisplayProgressBlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (v ProductionAdapterObjectiveCloseoutHostView) Clone() ProductionAdapterObjectiveCloseoutHostView {
	return CloneProductionAdapterObjectiveCloseoutHostView(v)
}

func (v ProductionAdapterObjectiveCloseoutHostView) Normalize() ProductionAdapterObjectiveCloseoutHostView {
	out := CloneProductionAdapterObjectiveCloseoutHostView(v)
	unsafe := productionAdapterObjectiveCloseoutHostViewUnsafeOutput(out)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = normalizeControlToken(out.Status)
	out.Mode = normalizeControlToken(out.Mode)
	out.DisplayState = normalizeControlToken(out.DisplayState)
	if out.Mode == "" {
		out.Mode = "production_adapter_objective_closeout_host_view"
	}
	out.HostViewRef = normalizeOneDisplaySafeRef(out.HostViewRef)
	out.DisplaySections = normalizeControlTokenList(out.DisplaySections)
	out.ObjectiveCloseoutPacketStatus = NormalizeVerificationStatus(string(out.ObjectiveCloseoutPacketStatus))
	out.DurableHandoffStatus = NormalizeHostActionStatus(string(out.DurableHandoffStatus))
	out.ReadbackStatus = NormalizeHostActionStatus(string(out.ReadbackStatus))
	out.ObjectiveCloseoutPacketRef = normalizeOneDisplaySafeRef(out.ObjectiveCloseoutPacketRef)
	out.ObjectiveCloseoutHandoffRef = normalizeOneDisplaySafeRef(out.ObjectiveCloseoutHandoffRef)
	out.ObjectiveCloseoutReadbackRef = normalizeOneDisplaySafeRef(out.ObjectiveCloseoutReadbackRef)
	out.ObjectiveRef = normalizeOneDisplaySafeRef(out.ObjectiveRef)
	out.AuthoritativeObjectiveCloseoutPacketRef = normalizeOneDisplaySafeRef(out.AuthoritativeObjectiveCloseoutPacketRef)
	out.AuthoritativeObjectiveCloseoutHandoffRef = normalizeOneDisplaySafeRef(out.AuthoritativeObjectiveCloseoutHandoffRef)
	out.AuthoritativeObjectiveRef = normalizeOneDisplaySafeRef(out.AuthoritativeObjectiveRef)
	out.AuthoritativeHostRunstoreRef = normalizeOneDisplaySafeRef(out.AuthoritativeHostRunstoreRef)
	out.AuthoritativeDurableEventRef = normalizeOneDisplaySafeRef(out.AuthoritativeDurableEventRef)
	out.AuthoritativeObjectiveStateRef = normalizeOneDisplaySafeRef(out.AuthoritativeObjectiveStateRef)
	out.ObservedObjectiveCloseoutPacketRef = normalizeOneDisplaySafeRef(out.ObservedObjectiveCloseoutPacketRef)
	out.ObservedObjectiveCloseoutHandoffRef = normalizeOneDisplaySafeRef(out.ObservedObjectiveCloseoutHandoffRef)
	out.ObservedObjectiveRef = normalizeOneDisplaySafeRef(out.ObservedObjectiveRef)
	out.ObservedHostRunstoreRef = normalizeOneDisplaySafeRef(out.ObservedHostRunstoreRef)
	out.ObservedAppliedDurableEventRef = normalizeOneDisplaySafeRef(out.ObservedAppliedDurableEventRef)
	out.ObservedAppliedRunstoreRef = normalizeOneDisplaySafeRef(out.ObservedAppliedRunstoreRef)
	out.ObservedAppliedObjectiveStateRef = normalizeOneDisplaySafeRef(out.ObservedAppliedObjectiveStateRef)
	out.HostCloseoutConfirmationRef = normalizeOneDisplaySafeRef(out.HostCloseoutConfirmationRef)
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
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.BlockedReasons = normalizeControlTokenList(out.BlockedReasons)
	out.DisplayProgressMissingInputs = normalizeMissingInputs(out.DisplayProgressMissingInputs)
	out.DisplayProgressBlockedReasons = normalizeControlTokenList(out.DisplayProgressBlockedReasons)
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
		out.ReadyForHostDurableApply = false
		out.ReadyForFailureReview = false
		out.ReadyForCompensationReview = false
		out.ReadyForObjectiveReturn = false
		out.IntermediateDisplay = false
		out.FinalDisplay = false
		out.FailureReviewDisplay = false
	}
	if out.Status == "" {
		out.Status = "blocked"
	}
	if out.DisplayState == "" {
		out.DisplayState = "blocked"
	}
	if unsafe || out.RawOutputLoaded {
		out.RawOutputLoaded = true
		out.Status = "blocked"
		out.ReadyForHostDisplay = false
		out.ReadyForHostDurableApply = false
		out.ReadyForFailureReview = false
		out.ReadyForCompensationReview = false
		out.ReadyForObjectiveReturn = false
		out.IntermediateDisplay = false
		out.FinalDisplay = false
		out.FailureReviewDisplay = false
		out.ObjectiveLifecycleClosed = false
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
	out.ReadyForHostDisplay = out.ReadyForHostDisplay &&
		out.Available &&
		out.HostViewRef != "" &&
		out.ObjectiveCloseoutPacketRef != "" &&
		out.ObjectiveRef != "" &&
		!out.RawOutputLoaded
	out.ReadyForHostDurableApply = out.ReadyForHostDurableApply &&
		out.ReadyForHostDisplay &&
		out.Status == "ready_for_host_durable_apply" &&
		out.DisplayState == "durable_apply_pending" &&
		!out.HostDurableWriteFailed &&
		out.ObjectiveCloseoutHandoffRef != "" &&
		out.HostObjectiveLifecycleRef != "" &&
		out.HostRunstoreRef != "" &&
		out.ExpectedDurableEventRef != "" &&
		out.ExpectedObjectiveStateRef != "" &&
		out.HostDurableApplyConfirmationRef != "" &&
		!containsMissingInput(out.MissingInputs, "host:objective_closeout_durable_handoff") &&
		!out.RawOutputLoaded
	out.ReadyForFailureReview = out.ReadyForFailureReview &&
		out.ReadyForHostDisplay &&
		out.Status == "objective_closeout_durable_failure_review" &&
		out.DisplayState == "durable_failure_review" &&
		out.HostDurableWriteReported &&
		out.HostDurableWriteFailed &&
		!out.HostDurableWriteSucceeded &&
		out.ObjectiveCloseoutReadbackRef != "" &&
		out.ObjectiveCloseoutHandoffRef != "" &&
		out.FailureRef != "" &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.RawOutputLoaded
	out.ReadyForCompensationReview = out.ReadyForCompensationReview &&
		out.ReadyForFailureReview &&
		out.CompensationRef != "" &&
		!out.RawOutputLoaded
	out.ReadyForObjectiveReturn = out.ReadyForObjectiveReturn &&
		out.ReadyForHostDisplay &&
		out.Status == "objective_lifecycle_closed" &&
		out.DisplayState == "objective_return_final" &&
		out.ObjectiveLifecycleClosed &&
		out.ObjectiveSatisfied &&
		out.ObjectiveCloseoutReadbackRef != "" &&
		out.AppliedDurableEventRef != "" &&
		out.AppliedObjectiveStateRef != "" &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0
	out.IntermediateDisplay = out.IntermediateDisplay &&
		out.ReadyForHostDisplay &&
		(out.DisplayState == "durable_apply_pending" || out.DisplayState == "durable_failure_review" || out.DisplayState == "closeout_review_pending") &&
		!out.FinalDisplay &&
		!out.RawOutputLoaded
	out.FinalDisplay = out.FinalDisplay &&
		out.ReadyForObjectiveReturn &&
		out.DisplayState == "objective_return_final" &&
		!out.IntermediateDisplay &&
		!out.RawOutputLoaded
	out.FailureReviewDisplay = out.FailureReviewDisplay &&
		out.ReadyForFailureReview &&
		out.DisplayState == "durable_failure_review" &&
		!out.FinalDisplay &&
		!out.RawOutputLoaded
	return out
}

func CloneProductionAdapterObjectiveCloseoutBlackboxFixture(in ProductionAdapterObjectiveCloseoutBlackboxFixture) ProductionAdapterObjectiveCloseoutBlackboxFixture {
	out := in
	out.DisplaySections = cloneStringSlice(in.DisplaySections)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.DisplayProgressMissingInputs = cloneMissingInputs(in.DisplayProgressMissingInputs)
	out.DisplayProgressBlockedReasons = cloneStringSlice(in.DisplayProgressBlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (f ProductionAdapterObjectiveCloseoutBlackboxFixture) Clone() ProductionAdapterObjectiveCloseoutBlackboxFixture {
	return CloneProductionAdapterObjectiveCloseoutBlackboxFixture(f)
}

func (f ProductionAdapterObjectiveCloseoutBlackboxFixture) Normalize() ProductionAdapterObjectiveCloseoutBlackboxFixture {
	out := CloneProductionAdapterObjectiveCloseoutBlackboxFixture(f)
	unsafe := productionAdapterObjectiveCloseoutBlackboxFixtureUnsafeOutput(out)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = normalizeControlToken(out.Status)
	out.Mode = normalizeControlToken(out.Mode)
	out.DisplayState = normalizeControlToken(out.DisplayState)
	if out.Mode == "" {
		out.Mode = "production_adapter_objective_closeout_blackbox_fixture"
	}
	out.FixtureRef = normalizeOneDisplaySafeRef(out.FixtureRef)
	out.HostViewRef = normalizeOneDisplaySafeRef(out.HostViewRef)
	out.DisplaySections = normalizeControlTokenList(out.DisplaySections)
	out.ObjectiveCloseoutPacketRef = normalizeOneDisplaySafeRef(out.ObjectiveCloseoutPacketRef)
	out.ObjectiveCloseoutHandoffRef = normalizeOneDisplaySafeRef(out.ObjectiveCloseoutHandoffRef)
	out.ObjectiveCloseoutReadbackRef = normalizeOneDisplaySafeRef(out.ObjectiveCloseoutReadbackRef)
	out.ObjectiveRef = normalizeOneDisplaySafeRef(out.ObjectiveRef)
	out.AuthoritativeObjectiveCloseoutPacketRef = normalizeOneDisplaySafeRef(out.AuthoritativeObjectiveCloseoutPacketRef)
	out.AuthoritativeObjectiveCloseoutHandoffRef = normalizeOneDisplaySafeRef(out.AuthoritativeObjectiveCloseoutHandoffRef)
	out.AuthoritativeObjectiveRef = normalizeOneDisplaySafeRef(out.AuthoritativeObjectiveRef)
	out.AuthoritativeHostRunstoreRef = normalizeOneDisplaySafeRef(out.AuthoritativeHostRunstoreRef)
	out.AuthoritativeDurableEventRef = normalizeOneDisplaySafeRef(out.AuthoritativeDurableEventRef)
	out.AuthoritativeObjectiveStateRef = normalizeOneDisplaySafeRef(out.AuthoritativeObjectiveStateRef)
	out.ObservedObjectiveCloseoutPacketRef = normalizeOneDisplaySafeRef(out.ObservedObjectiveCloseoutPacketRef)
	out.ObservedObjectiveCloseoutHandoffRef = normalizeOneDisplaySafeRef(out.ObservedObjectiveCloseoutHandoffRef)
	out.ObservedObjectiveRef = normalizeOneDisplaySafeRef(out.ObservedObjectiveRef)
	out.ObservedHostRunstoreRef = normalizeOneDisplaySafeRef(out.ObservedHostRunstoreRef)
	out.ObservedAppliedDurableEventRef = normalizeOneDisplaySafeRef(out.ObservedAppliedDurableEventRef)
	out.ObservedAppliedRunstoreRef = normalizeOneDisplaySafeRef(out.ObservedAppliedRunstoreRef)
	out.ObservedAppliedObjectiveStateRef = normalizeOneDisplaySafeRef(out.ObservedAppliedObjectiveStateRef)
	out.HostObjectiveLifecycleRef = normalizeOneDisplaySafeRef(out.HostObjectiveLifecycleRef)
	out.HostRunstoreRef = normalizeOneDisplaySafeRef(out.HostRunstoreRef)
	out.AppliedDurableEventRef = normalizeOneDisplaySafeRef(out.AppliedDurableEventRef)
	out.AppliedObjectiveStateRef = normalizeOneDisplaySafeRef(out.AppliedObjectiveStateRef)
	out.CompletionAuditResultBindingRef = normalizeOneDisplaySafeRef(out.CompletionAuditResultBindingRef)
	out.AdapterRef = normalizeOneDisplaySafeRef(out.AdapterRef)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.BlockedReasons = normalizeControlTokenList(out.BlockedReasons)
	out.DisplayProgressMissingInputs = normalizeMissingInputs(out.DisplayProgressMissingInputs)
	out.DisplayProgressBlockedReasons = normalizeControlTokenList(out.DisplayProgressBlockedReasons)
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
		out.ReadyForObjectiveReturn = false
		out.IntermediateDisplay = false
		out.FinalDisplay = false
	}
	if out.Status == "" {
		out.Status = "blocked"
	}
	if out.DisplayState == "" {
		out.DisplayState = "blocked"
	}
	if unsafe || out.RawOutputLoaded {
		out.RawOutputLoaded = true
		out.Status = "blocked"
		out.ReadyForHostDisplay = false
		out.ReadyForObjectiveReturn = false
		out.IntermediateDisplay = false
		out.FinalDisplay = false
		out.ObjectiveLifecycleClosed = false
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
	out.ReadyForHostDisplay = out.ReadyForHostDisplay &&
		out.Available &&
		out.FixtureRef != "" &&
		out.HostViewRef != "" &&
		out.ObjectiveCloseoutPacketRef != "" &&
		out.ObjectiveRef != "" &&
		!out.RawOutputLoaded
	out.ReadyForObjectiveReturn = out.ReadyForObjectiveReturn &&
		out.ReadyForHostDisplay &&
		out.Status == "ready_for_objective_return" &&
		out.DisplayState == "objective_return_final" &&
		out.ObjectiveLifecycleClosed &&
		out.ObjectiveSatisfied &&
		out.ObjectiveCloseoutReadbackRef != "" &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0
	out.IntermediateDisplay = out.IntermediateDisplay &&
		out.ReadyForHostDisplay &&
		(out.DisplayState == "durable_apply_pending" || out.DisplayState == "closeout_review_pending") &&
		!out.FinalDisplay &&
		!out.RawOutputLoaded
	out.FinalDisplay = out.FinalDisplay &&
		out.ReadyForObjectiveReturn &&
		out.DisplayState == "objective_return_final" &&
		!out.IntermediateDisplay &&
		!out.RawOutputLoaded
	return out
}

func productionAdapterObjectiveCloseoutHostViewBlock(result ProductionAdapterObjectiveCloseoutHostView, failure FailureClass, reason string, missing MissingInput, next NextHostAction) ProductionAdapterObjectiveCloseoutHostView {
	result.Status = "blocked"
	result.DisplayState = "blocked"
	result.ReadyForHostDurableApply = false
	result.ReadyForFailureReview = false
	result.ReadyForCompensationReview = false
	result.ReadyForObjectiveReturn = false
	result.IntermediateDisplay = false
	result.FinalDisplay = false
	result.FailureReviewDisplay = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = NormalizeNextHostAction(string(next))
	if result.NextHostAction == "" {
		result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	}
	result.Boundaries = AppendBoundaries(result.Boundaries, "objective_closeout_host_view_blocked")
	return result
}

func productionAdapterObjectiveCloseoutBlackboxFixtureBlock(result ProductionAdapterObjectiveCloseoutBlackboxFixture, failure FailureClass, reason string, missing MissingInput, next NextHostAction) ProductionAdapterObjectiveCloseoutBlackboxFixture {
	result.Status = "blocked"
	result.DisplayState = "blocked"
	result.ReadyForHostDisplay = false
	result.ReadyForObjectiveReturn = false
	result.IntermediateDisplay = false
	result.FinalDisplay = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = NormalizeNextHostAction(string(next))
	if result.NextHostAction == "" {
		result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	}
	result.Boundaries = AppendBoundaries(result.Boundaries, "objective_closeout_blackbox_fixture_blocked")
	return result
}

func productionAdapterObjectiveCloseoutDisplaySections() []string {
	return []string{
		"objective_closeout_summary",
		"authorization_chain",
		"completion_audit",
		"durable_handoff",
		"durable_readback",
		"objective_return",
	}
}

func productionAdapterObjectiveCloseoutHostViewEvidenceRefs(packet ProductionAdapterObjectiveCloseoutPacket, handoff ProductionAdapterObjectiveCloseoutDurableHandoff, readback ProductionAdapterObjectiveCloseoutReadback) []EvidenceRef {
	evidence := cloneEvidenceRefs(packet.EvidenceRefs)
	evidence = append(evidence, handoff.EvidenceRefs...)
	evidence = append(evidence, readback.EvidenceRefs...)
	return normalizeEvidenceRefs(evidence)
}

func productionAdapterObjectiveCloseoutHostViewMissingInputs(packet ProductionAdapterObjectiveCloseoutPacket, handoff ProductionAdapterObjectiveCloseoutDurableHandoff, readback ProductionAdapterObjectiveCloseoutReadback, handoffProvided bool, readbackProvided bool) []MissingInput {
	var missing []MissingInput
	for _, item := range packet.MissingInputs {
		missing = AppendMissingInputs(missing, item)
	}
	if handoffProvided {
		for _, item := range handoff.MissingInputs {
			missing = AppendMissingInputs(missing, item)
		}
	} else {
		missing = AppendMissingInputs(missing, "host:objective_closeout_durable_handoff")
	}
	if readbackProvided {
		for _, item := range readback.MissingInputs {
			missing = AppendMissingInputs(missing, item)
		}
	} else if handoffProvided && handoff.ReadyForHostDurableApply {
		missing = AppendMissingInputs(missing, "host:objective_closeout_readback")
	}
	return normalizeMissingInputs(missing)
}

func productionAdapterObjectiveCloseoutHostViewDisplayProgressMissingInputs(handoff ProductionAdapterObjectiveCloseoutDurableHandoff, handoffProvided bool, readbackProvided bool) []MissingInput {
	if handoffProvided && handoff.ReadyForHostDurableApply && !readbackProvided {
		return normalizeMissingInputs([]MissingInput{"host:objective_closeout_readback"})
	}
	return nil
}

func productionAdapterObjectiveCloseoutHostViewBlockedReasons(packet ProductionAdapterObjectiveCloseoutPacket, handoff ProductionAdapterObjectiveCloseoutDurableHandoff, readback ProductionAdapterObjectiveCloseoutReadback, handoffProvided bool, readbackProvided bool) []string {
	var blocked []string
	for _, reason := range packet.BlockedReasons {
		blocked = appendUniqueControlToken(blocked, reason)
	}
	if handoffProvided {
		for _, reason := range handoff.BlockedReasons {
			blocked = appendUniqueControlToken(blocked, reason)
		}
	} else {
		blocked = appendUniqueControlToken(blocked, "objective_closeout_durable_handoff_not_ready")
	}
	if readbackProvided {
		for _, reason := range readback.BlockedReasons {
			blocked = appendUniqueControlToken(blocked, reason)
		}
	} else if handoffProvided && handoff.ReadyForHostDurableApply {
		blocked = appendUniqueControlToken(blocked, "objective_closeout_readback_not_ready")
	}
	return normalizeControlTokenList(blocked)
}

func productionAdapterObjectiveCloseoutHostViewDisplayProgressBlockedReasons(handoff ProductionAdapterObjectiveCloseoutDurableHandoff, handoffProvided bool, readbackProvided bool) []string {
	if handoffProvided && handoff.ReadyForHostDurableApply && !readbackProvided {
		return normalizeControlTokenList([]string{"objective_closeout_readback_not_ready"})
	}
	return nil
}

func productionAdapterObjectiveCloseoutBlackboxFixtureMissingInputs(view ProductionAdapterObjectiveCloseoutHostView) []MissingInput {
	missing := cloneMissingInputs(view.MissingInputs)
	if view.ReadyForHostDurableApply && view.DisplayState == "durable_apply_pending" {
		missing = productionAdapterObjectiveCloseoutMissingWithout(missing, "host:objective_closeout_readback")
	}
	return normalizeMissingInputs(missing)
}

func productionAdapterObjectiveCloseoutBlackboxFixtureBlockedReasons(view ProductionAdapterObjectiveCloseoutHostView) []string {
	blocked := cloneStringSlice(view.BlockedReasons)
	if view.ReadyForHostDurableApply && view.DisplayState == "durable_apply_pending" {
		blocked = productionAdapterObjectiveCloseoutControlTokensWithout(blocked, "objective_closeout_readback_not_ready")
	}
	return normalizeControlTokenList(blocked)
}

func productionAdapterObjectiveCloseoutMissingWithout(values []MissingInput, excluded ...MissingInput) []MissingInput {
	excludedSet := map[MissingInput]bool{}
	for _, item := range excluded {
		excludedSet[MissingInput(normalizeControlToken(string(item)))] = true
	}
	var out []MissingInput
	for _, item := range values {
		normalized := MissingInput(normalizeControlToken(string(item)))
		if normalized == "" || excludedSet[normalized] {
			continue
		}
		out = AppendMissingInputs(out, normalized)
	}
	return out
}

func productionAdapterObjectiveCloseoutControlTokensWithout(values []string, excluded ...string) []string {
	excludedSet := map[string]bool{}
	for _, item := range excluded {
		excludedSet[normalizeControlToken(item)] = true
	}
	var out []string
	for _, item := range values {
		normalized := normalizeControlToken(item)
		if normalized == "" || excludedSet[normalized] {
			continue
		}
		out = appendUniqueControlToken(out, normalized)
	}
	return out
}

func productionAdapterObjectiveCloseoutHostViewBoundaries(groups ...[]Boundary) []Boundary {
	return MergeBoundaries(
		[]Boundary{
			"production_adapter_objective_closeout_host_view",
			"objective_closeout_host_view_projection_only",
			"host_cli_objective_closeout_display",
			"host_owned_objective_closeout_display",
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

func productionAdapterObjectiveCloseoutBlackboxFixtureBoundaries(viewBoundaries []Boundary) []Boundary {
	return MergeBoundaries(
		[]Boundary{
			"production_adapter_objective_closeout_blackbox_fixture",
			"objective_closeout_blackbox_fixture_projection_only",
			"host_cli_objective_closeout_display",
			"host_owned_objective_closeout_fixture",
			"display_safe_refs_only",
			"display_safe_result_refs_only",
			"no_runner_dispatch",
			"no_durable_write_by_core",
			"no_objective_store_write_by_core",
			"no_runstore_write_by_core",
		},
		viewBoundaries,
	)
}

type productionAdapterObjectiveCloseoutHostViewMismatch struct {
	reason  string
	missing MissingInput
}

func productionAdapterObjectiveCloseoutHostViewMismatches(packet ProductionAdapterObjectiveCloseoutPacket, handoff ProductionAdapterObjectiveCloseoutDurableHandoff, readback ProductionAdapterObjectiveCloseoutReadback, handoffProvided bool, readbackProvided bool) []productionAdapterObjectiveCloseoutHostViewMismatch {
	var out []productionAdapterObjectiveCloseoutHostViewMismatch
	if handoffProvided {
		out = append(out, productionAdapterObjectiveCloseoutHostViewRefMismatches(packet.ObjectiveCloseoutPacketRef, handoff.ObjectiveCloseoutPacketRef, "closeout_view_packet_handoff_packet_ref_mismatch", "host:objective_closeout_packet")...)
		out = append(out, productionAdapterObjectiveCloseoutHostViewRefMismatches(packet.ObjectiveRef, handoff.ObjectiveRef, "closeout_view_packet_handoff_objective_ref_mismatch", "host:objective_ref")...)
		out = append(out, productionAdapterObjectiveCloseoutHostViewRefMismatches(packet.CompletionAuditResultBindingRef, handoff.CompletionAuditResultBindingRef, "closeout_view_packet_handoff_audit_binding_ref_mismatch", "host:completion_audit_result_binding")...)
	}
	if readbackProvided {
		out = append(out, productionAdapterObjectiveCloseoutHostViewRefMismatches(packet.ObjectiveCloseoutPacketRef, readback.ObjectiveCloseoutPacketRef, "closeout_view_packet_readback_packet_ref_mismatch", "host:objective_closeout_packet")...)
		out = append(out, productionAdapterObjectiveCloseoutHostViewRefMismatches(packet.ObjectiveRef, readback.ObjectiveRef, "closeout_view_packet_readback_objective_ref_mismatch", "host:objective_ref")...)
		if handoffProvided {
			out = append(out, productionAdapterObjectiveCloseoutHostViewRefMismatches(handoff.ObjectiveCloseoutHandoffRef, readback.ObjectiveCloseoutHandoffRef, "closeout_view_handoff_readback_ref_mismatch", "host:objective_closeout_durable_handoff")...)
			out = append(out, productionAdapterObjectiveCloseoutHostViewRefMismatches(handoff.HostRunstoreRef, readback.HostRunstoreRef, "closeout_view_handoff_readback_runstore_ref_mismatch", "host:runstore_ref")...)
			out = append(out, productionAdapterObjectiveCloseoutHostViewRefMismatches(handoff.ExpectedDurableEventRef, readback.AppliedDurableEventRef, "closeout_view_durable_event_ref_mismatch", "host:durable_event_ref")...)
			out = append(out, productionAdapterObjectiveCloseoutHostViewRefMismatches(handoff.HostRunstoreRef, readback.AppliedRunstoreRef, "closeout_view_runstore_ref_mismatch", "host:runstore_ref")...)
			out = append(out, productionAdapterObjectiveCloseoutHostViewRefMismatches(handoff.ExpectedObjectiveStateRef, readback.AppliedObjectiveStateRef, "closeout_view_objective_state_ref_mismatch", "host:objective_state_ref")...)
		}
	}
	return out
}

func productionAdapterObjectiveCloseoutHostViewRefMismatches(left DisplaySafeRef, right DisplaySafeRef, reason string, missing MissingInput) []productionAdapterObjectiveCloseoutHostViewMismatch {
	left = normalizeOneDisplaySafeRef(left)
	right = normalizeOneDisplaySafeRef(right)
	if left != "" && right != "" && left != right {
		return []productionAdapterObjectiveCloseoutHostViewMismatch{{reason: reason, missing: missing}}
	}
	return nil
}

func productionAdapterObjectiveCloseoutHostViewUnsafe(input ProductionAdapterObjectiveCloseoutHostViewInput, packet ProductionAdapterObjectiveCloseoutPacket, handoff ProductionAdapterObjectiveCloseoutDurableHandoff, readback ProductionAdapterObjectiveCloseoutReadback, handoffProvided bool, readbackProvided bool) bool {
	return input.RawOutputLoaded ||
		displaySafeRefRejected(input.HostViewRef) ||
		productionAdapterObjectiveCloseoutPacketUnsafe(packet) ||
		(handoffProvided && productionAdapterObjectiveCloseoutDurableHandoffUnsafe(handoff)) ||
		(readbackProvided && productionAdapterObjectiveCloseoutReadbackUnsafe(readback))
}

func productionAdapterObjectiveCloseoutHostViewUnsafeOutput(input ProductionAdapterObjectiveCloseoutHostView) bool {
	return displaySafeRefRejected(input.HostViewRef) ||
		displaySafeRefRejected(input.ObjectiveCloseoutPacketRef) ||
		displaySafeRefRejected(input.ObjectiveCloseoutHandoffRef) ||
		displaySafeRefRejected(input.ObjectiveCloseoutReadbackRef) ||
		displaySafeRefRejected(input.ObjectiveRef) ||
		displaySafeRefRejected(input.AuthoritativeObjectiveCloseoutPacketRef) ||
		displaySafeRefRejected(input.AuthoritativeObjectiveCloseoutHandoffRef) ||
		displaySafeRefRejected(input.AuthoritativeObjectiveRef) ||
		displaySafeRefRejected(input.AuthoritativeHostRunstoreRef) ||
		displaySafeRefRejected(input.AuthoritativeDurableEventRef) ||
		displaySafeRefRejected(input.AuthoritativeObjectiveStateRef) ||
		displaySafeRefRejected(input.ObservedObjectiveCloseoutPacketRef) ||
		displaySafeRefRejected(input.ObservedObjectiveCloseoutHandoffRef) ||
		displaySafeRefRejected(input.ObservedObjectiveRef) ||
		displaySafeRefRejected(input.ObservedHostRunstoreRef) ||
		displaySafeRefRejected(input.ObservedAppliedDurableEventRef) ||
		displaySafeRefRejected(input.ObservedAppliedRunstoreRef) ||
		displaySafeRefRejected(input.ObservedAppliedObjectiveStateRef) ||
		displaySafeRefRejected(input.HostCloseoutConfirmationRef) ||
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
		displaySafeRefRejected(input.CompletionHandoffRef) ||
		evidenceRefsRejected(input.EvidenceRefs) ||
		input.RawOutputLoaded
}

func productionAdapterObjectiveCloseoutBlackboxFixtureUnsafeOutput(input ProductionAdapterObjectiveCloseoutBlackboxFixture) bool {
	return displaySafeRefRejected(input.FixtureRef) ||
		displaySafeRefRejected(input.HostViewRef) ||
		displaySafeRefRejected(input.ObjectiveCloseoutPacketRef) ||
		displaySafeRefRejected(input.ObjectiveCloseoutHandoffRef) ||
		displaySafeRefRejected(input.ObjectiveCloseoutReadbackRef) ||
		displaySafeRefRejected(input.ObjectiveRef) ||
		displaySafeRefRejected(input.AuthoritativeObjectiveCloseoutPacketRef) ||
		displaySafeRefRejected(input.AuthoritativeObjectiveCloseoutHandoffRef) ||
		displaySafeRefRejected(input.AuthoritativeObjectiveRef) ||
		displaySafeRefRejected(input.AuthoritativeHostRunstoreRef) ||
		displaySafeRefRejected(input.AuthoritativeDurableEventRef) ||
		displaySafeRefRejected(input.AuthoritativeObjectiveStateRef) ||
		displaySafeRefRejected(input.ObservedObjectiveCloseoutPacketRef) ||
		displaySafeRefRejected(input.ObservedObjectiveCloseoutHandoffRef) ||
		displaySafeRefRejected(input.ObservedObjectiveRef) ||
		displaySafeRefRejected(input.ObservedHostRunstoreRef) ||
		displaySafeRefRejected(input.ObservedAppliedDurableEventRef) ||
		displaySafeRefRejected(input.ObservedAppliedRunstoreRef) ||
		displaySafeRefRejected(input.ObservedAppliedObjectiveStateRef) ||
		displaySafeRefRejected(input.HostObjectiveLifecycleRef) ||
		displaySafeRefRejected(input.HostRunstoreRef) ||
		displaySafeRefRejected(input.AppliedDurableEventRef) ||
		displaySafeRefRejected(input.AppliedObjectiveStateRef) ||
		displaySafeRefRejected(input.CompletionAuditResultBindingRef) ||
		displaySafeRefRejected(input.AdapterRef) ||
		input.RawOutputLoaded
}

func productionAdapterObjectiveCloseoutPacketEmpty(packet ProductionAdapterObjectiveCloseoutPacket) bool {
	return !packet.Projected &&
		!packet.Available &&
		packet.Status == "" &&
		packet.Mode == "" &&
		packet.ObjectiveCloseoutPacketRef == "" &&
		packet.ObjectiveRef == "" &&
		packet.CompletionAuditResultBindingRef == "" &&
		len(packet.MissingInputs) == 0 &&
		len(packet.BlockedReasons) == 0 &&
		len(packet.Boundaries) == 0 &&
		packet.NextHostAction == "" &&
		!packet.RawOutputLoaded
}

func productionAdapterObjectiveCloseoutDurableHandoffEmpty(handoff ProductionAdapterObjectiveCloseoutDurableHandoff) bool {
	return !handoff.Projected &&
		!handoff.Available &&
		handoff.Status == "" &&
		handoff.Mode == "" &&
		handoff.ObjectiveCloseoutHandoffRef == "" &&
		handoff.ObjectiveCloseoutPacketRef == "" &&
		handoff.ObjectiveRef == "" &&
		len(handoff.MissingInputs) == 0 &&
		len(handoff.BlockedReasons) == 0 &&
		len(handoff.Boundaries) == 0 &&
		handoff.NextHostAction == "" &&
		!handoff.RawOutputLoaded
}

func productionAdapterObjectiveCloseoutReadbackEmpty(readback ProductionAdapterObjectiveCloseoutReadback) bool {
	return !readback.Projected &&
		!readback.Available &&
		readback.Status == "" &&
		readback.Mode == "" &&
		readback.ObjectiveCloseoutReadbackRef == "" &&
		readback.ObjectiveCloseoutHandoffRef == "" &&
		readback.ObjectiveCloseoutPacketRef == "" &&
		readback.ObjectiveRef == "" &&
		len(readback.MissingInputs) == 0 &&
		len(readback.BlockedReasons) == 0 &&
		len(readback.Boundaries) == 0 &&
		readback.NextHostAction == "" &&
		!readback.RawOutputLoaded
}

func productionAdapterObjectiveCloseoutHostViewEmpty(view ProductionAdapterObjectiveCloseoutHostView) bool {
	return !view.Projected &&
		!view.Available &&
		view.Status == "" &&
		view.Mode == "" &&
		view.HostViewRef == "" &&
		view.ObjectiveCloseoutPacketRef == "" &&
		view.ObjectiveRef == "" &&
		len(view.MissingInputs) == 0 &&
		len(view.BlockedReasons) == 0 &&
		len(view.Boundaries) == 0 &&
		view.NextHostAction == "" &&
		!view.RawOutputLoaded
}

func unavailableProductionAdapterObjectiveCloseoutHostView() ProductionAdapterObjectiveCloseoutHostView {
	return ProductionAdapterObjectiveCloseoutHostView{
		ContractVersion: ContractVersion,
		Projected:       true,
		Available:       false,
		Status:          "unavailable",
		Mode:            "production_adapter_objective_closeout_host_view",
		DisplaySections: productionAdapterObjectiveCloseoutDisplaySections(),
		Boundaries: []Boundary{
			"production_adapter_objective_closeout_host_view",
			"objective_closeout_host_view_projection_only",
			"host_cli_objective_closeout_display",
			"display_safe_refs_only",
			"no_runner_dispatch",
			"no_durable_write_by_core",
			"no_objective_store_write_by_core",
			"no_runstore_write_by_core",
		},
		RunnerEffect:   "none",
		PromptEffect:   "none",
		NextHostAction: "provide_objective_closeout_packet",
	}
}

func unavailableProductionAdapterObjectiveCloseoutBlackboxFixture() ProductionAdapterObjectiveCloseoutBlackboxFixture {
	return ProductionAdapterObjectiveCloseoutBlackboxFixture{
		ContractVersion: ContractVersion,
		Projected:       true,
		Available:       false,
		Status:          "unavailable",
		Mode:            "production_adapter_objective_closeout_blackbox_fixture",
		DisplaySections: append(productionAdapterObjectiveCloseoutDisplaySections(), "objective_closeout_blackbox_assertions"),
		Boundaries: []Boundary{
			"production_adapter_objective_closeout_blackbox_fixture",
			"objective_closeout_blackbox_fixture_projection_only",
			"host_cli_objective_closeout_display",
			"display_safe_refs_only",
			"no_runner_dispatch",
			"no_durable_write_by_core",
			"no_objective_store_write_by_core",
			"no_runstore_write_by_core",
		},
		RunnerEffect:   "none",
		PromptEffect:   "none",
		NextHostAction: "provide_objective_closeout_host_view",
	}
}

func firstBool(values ...bool) bool {
	for _, value := range values {
		if value {
			return true
		}
	}
	return false
}

func firstControlToken(values ...string) string {
	for _, value := range values {
		normalized := normalizeControlToken(value)
		if normalized != "" {
			return normalized
		}
	}
	return ""
}

func onlyDisplayProgressBlocks(blocked []string, allowed ...string) bool {
	allowedSet := map[string]bool{}
	for _, item := range allowed {
		allowedSet[normalizeControlToken(item)] = true
	}
	for _, item := range blocked {
		if !allowedSet[normalizeControlToken(item)] {
			return false
		}
	}
	return true
}

func closeoutHostViewControlTokenContains(values []string, want string) bool {
	needle := normalizeControlToken(want)
	if needle == "" {
		return false
	}
	for _, value := range values {
		if normalizeControlToken(value) == needle {
			return true
		}
	}
	return false
}
