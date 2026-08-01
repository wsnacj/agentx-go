package controlcontract

type ProductionAdapterObjectiveCloseoutFailureReviewPacketInput struct {
	FailureReviewPacketRef DisplaySafeRef                             `json:"failure_review_packet_ref,omitempty"`
	HostView               ProductionAdapterObjectiveCloseoutHostView `json:"host_view,omitempty"`
	RawOutputLoaded        bool                                       `json:"raw_output_loaded"`
}

type ProductionAdapterObjectiveCloseoutFailureReviewPacket struct {
	ContractVersion                 string         `json:"contract_version,omitempty"`
	Projected                       bool           `json:"projected"`
	Available                       bool           `json:"available"`
	Status                          string         `json:"status,omitempty"`
	Mode                            string         `json:"mode,omitempty"`
	ReadyForHostDisplay             bool           `json:"ready_for_host_display"`
	ReadyForFailureReview           bool           `json:"ready_for_failure_review"`
	ReadyForCompensationReview      bool           `json:"ready_for_compensation_review"`
	ObjectiveLifecycleClosed        bool           `json:"objective_lifecycle_closed"`
	ObjectiveSatisfied              bool           `json:"objective_satisfied"`
	HostDurableWriteReported        bool           `json:"host_durable_write_reported"`
	HostDurableWriteSucceeded       bool           `json:"host_durable_write_succeeded"`
	HostDurableWriteFailed          bool           `json:"host_durable_write_failed"`
	CoreInvocationExecuted          bool           `json:"core_invocation_executed"`
	DurableWriteByCore              bool           `json:"durable_write_by_core"`
	ObjectiveStoreWriteByCore       bool           `json:"objective_store_write_by_core"`
	RunstoreWriteByCore             bool           `json:"runstore_write_by_core"`
	FailureReviewPacketRef          DisplaySafeRef `json:"failure_review_packet_ref,omitempty"`
	HostViewRef                     DisplaySafeRef `json:"host_view_ref,omitempty"`
	DisplaySections                 []string       `json:"display_sections,omitempty"`
	ObjectiveCloseoutPacketRef      DisplaySafeRef `json:"objective_closeout_packet_ref,omitempty"`
	ObjectiveCloseoutHandoffRef     DisplaySafeRef `json:"objective_closeout_handoff_ref,omitempty"`
	ObjectiveCloseoutReadbackRef    DisplaySafeRef `json:"objective_closeout_readback_ref,omitempty"`
	ObjectiveRef                    DisplaySafeRef `json:"objective_ref,omitempty"`
	HostObjectiveLifecycleRef       DisplaySafeRef `json:"host_objective_lifecycle_ref,omitempty"`
	HostRunstoreRef                 DisplaySafeRef `json:"host_runstore_ref,omitempty"`
	ExpectedDurableEventRef         DisplaySafeRef `json:"expected_durable_event_ref,omitempty"`
	ExpectedObjectiveStateRef       DisplaySafeRef `json:"expected_objective_state_ref,omitempty"`
	FailureRef                      DisplaySafeRef `json:"failure_ref,omitempty"`
	CompensationRef                 DisplaySafeRef `json:"compensation_ref,omitempty"`
	HostDurableApplyConfirmationRef DisplaySafeRef `json:"host_durable_apply_confirmation_ref,omitempty"`
	CompletionAuditResultBindingRef DisplaySafeRef `json:"completion_audit_result_binding_ref,omitempty"`
	AdapterRef                      DisplaySafeRef `json:"adapter_ref,omitempty"`
	EvidenceRefs                    []EvidenceRef  `json:"evidence_refs,omitempty"`
	MissingInputs                   []MissingInput `json:"missing_inputs,omitempty"`
	BlockedReasons                  []string       `json:"blocked_reasons,omitempty"`
	FailureClass                    FailureClass   `json:"failure_class,omitempty"`
	Boundaries                      []Boundary     `json:"boundaries,omitempty"`
	NextHostAction                  NextHostAction `json:"next_host_action,omitempty"`
	RunnerEffect                    string         `json:"runner_effect,omitempty"`
	PromptEffect                    string         `json:"prompt_effect,omitempty"`
	RawOutputLoaded                 bool           `json:"raw_output_loaded"`
}

// agentx-api: internal_candidate
type ProductionAdapterObjectiveCloseoutFailureReviewBlackboxFixtureInput struct {
	FixtureRef          DisplaySafeRef                                        `json:"fixture_ref,omitempty"`
	FailureReviewPacket ProductionAdapterObjectiveCloseoutFailureReviewPacket `json:"failure_review_packet,omitempty"`
	RawOutputLoaded     bool                                                  `json:"raw_output_loaded"`
}

// agentx-api: internal_candidate
type ProductionAdapterObjectiveCloseoutFailureReviewBlackboxFixture struct {
	ContractVersion                 string         `json:"contract_version,omitempty"`
	Projected                       bool           `json:"projected"`
	Available                       bool           `json:"available"`
	Status                          string         `json:"status,omitempty"`
	Mode                            string         `json:"mode,omitempty"`
	ReadyForHostDisplay             bool           `json:"ready_for_host_display"`
	ReadyForFailureReview           bool           `json:"ready_for_failure_review"`
	ReadyForCompensationReview      bool           `json:"ready_for_compensation_review"`
	ObjectiveLifecycleClosed        bool           `json:"objective_lifecycle_closed"`
	ObjectiveSatisfied              bool           `json:"objective_satisfied"`
	HostDurableWriteReported        bool           `json:"host_durable_write_reported"`
	HostDurableWriteSucceeded       bool           `json:"host_durable_write_succeeded"`
	HostDurableWriteFailed          bool           `json:"host_durable_write_failed"`
	CoreInvocationExecuted          bool           `json:"core_invocation_executed"`
	DurableWriteByCore              bool           `json:"durable_write_by_core"`
	ObjectiveStoreWriteByCore       bool           `json:"objective_store_write_by_core"`
	RunstoreWriteByCore             bool           `json:"runstore_write_by_core"`
	FixtureRef                      DisplaySafeRef `json:"fixture_ref,omitempty"`
	FailureReviewPacketRef          DisplaySafeRef `json:"failure_review_packet_ref,omitempty"`
	HostViewRef                     DisplaySafeRef `json:"host_view_ref,omitempty"`
	DisplaySections                 []string       `json:"display_sections,omitempty"`
	ObjectiveCloseoutPacketRef      DisplaySafeRef `json:"objective_closeout_packet_ref,omitempty"`
	ObjectiveCloseoutHandoffRef     DisplaySafeRef `json:"objective_closeout_handoff_ref,omitempty"`
	ObjectiveCloseoutReadbackRef    DisplaySafeRef `json:"objective_closeout_readback_ref,omitempty"`
	ObjectiveRef                    DisplaySafeRef `json:"objective_ref,omitempty"`
	HostObjectiveLifecycleRef       DisplaySafeRef `json:"host_objective_lifecycle_ref,omitempty"`
	HostRunstoreRef                 DisplaySafeRef `json:"host_runstore_ref,omitempty"`
	ExpectedDurableEventRef         DisplaySafeRef `json:"expected_durable_event_ref,omitempty"`
	ExpectedObjectiveStateRef       DisplaySafeRef `json:"expected_objective_state_ref,omitempty"`
	FailureRef                      DisplaySafeRef `json:"failure_ref,omitempty"`
	CompensationRef                 DisplaySafeRef `json:"compensation_ref,omitempty"`
	HostDurableApplyConfirmationRef DisplaySafeRef `json:"host_durable_apply_confirmation_ref,omitempty"`
	CompletionAuditResultBindingRef DisplaySafeRef `json:"completion_audit_result_binding_ref,omitempty"`
	AdapterRef                      DisplaySafeRef `json:"adapter_ref,omitempty"`
	MissingInputs                   []MissingInput `json:"missing_inputs,omitempty"`
	BlockedReasons                  []string       `json:"blocked_reasons,omitempty"`
	FailureClass                    FailureClass   `json:"failure_class,omitempty"`
	Boundaries                      []Boundary     `json:"boundaries,omitempty"`
	NextHostAction                  NextHostAction `json:"next_host_action,omitempty"`
	RunnerEffect                    string         `json:"runner_effect,omitempty"`
	PromptEffect                    string         `json:"prompt_effect,omitempty"`
	RawOutputLoaded                 bool           `json:"raw_output_loaded"`
}

func BuildProductionAdapterObjectiveCloseoutFailureReviewPacket(input ProductionAdapterObjectiveCloseoutFailureReviewPacketInput) ProductionAdapterObjectiveCloseoutFailureReviewPacket {
	if productionAdapterObjectiveCloseoutHostViewEmpty(input.HostView) {
		return unavailableProductionAdapterObjectiveCloseoutFailureReviewPacket()
	}
	view := input.HostView.Normalize()
	result := ProductionAdapterObjectiveCloseoutFailureReviewPacket{
		ContractVersion:                 ContractVersion,
		Projected:                       true,
		Available:                       view.Available,
		Status:                          "blocked",
		Mode:                            "production_adapter_objective_closeout_failure_review_packet",
		FailureReviewPacketRef:          normalizeOneDisplaySafeRef(input.FailureReviewPacketRef),
		HostViewRef:                     view.HostViewRef,
		DisplaySections:                 productionAdapterObjectiveCloseoutFailureReviewDisplaySections(),
		ObjectiveLifecycleClosed:        false,
		ObjectiveSatisfied:              false,
		HostDurableWriteReported:        view.HostDurableWriteReported,
		HostDurableWriteSucceeded:       view.HostDurableWriteSucceeded,
		HostDurableWriteFailed:          view.HostDurableWriteFailed,
		ObjectiveCloseoutPacketRef:      view.ObjectiveCloseoutPacketRef,
		ObjectiveCloseoutHandoffRef:     view.ObjectiveCloseoutHandoffRef,
		ObjectiveCloseoutReadbackRef:    view.ObjectiveCloseoutReadbackRef,
		ObjectiveRef:                    view.ObjectiveRef,
		HostObjectiveLifecycleRef:       view.HostObjectiveLifecycleRef,
		HostRunstoreRef:                 view.HostRunstoreRef,
		ExpectedDurableEventRef:         view.ExpectedDurableEventRef,
		ExpectedObjectiveStateRef:       view.ExpectedObjectiveStateRef,
		FailureRef:                      view.FailureRef,
		CompensationRef:                 view.CompensationRef,
		HostDurableApplyConfirmationRef: view.HostDurableApplyConfirmationRef,
		CompletionAuditResultBindingRef: view.CompletionAuditResultBindingRef,
		AdapterRef:                      view.AdapterRef,
		EvidenceRefs:                    cloneEvidenceRefs(view.EvidenceRefs),
		MissingInputs:                   cloneMissingInputs(view.MissingInputs),
		BlockedReasons:                  cloneStringSlice(view.BlockedReasons),
		FailureClass:                    firstFailureClass(view.FailureClass, FailureVerificationFailed),
		Boundaries:                      productionAdapterObjectiveCloseoutFailureReviewPacketBoundaries(view.Boundaries),
		NextHostAction:                  firstNextHostAction(view.NextHostAction, "review_objective_closeout_durable_failure"),
		RunnerEffect:                    "none",
		PromptEffect:                    "none",
		RawOutputLoaded:                 input.RawOutputLoaded || view.RawOutputLoaded,
	}
	if input.RawOutputLoaded || displaySafeRefRejected(input.FailureReviewPacketRef) || productionAdapterObjectiveCloseoutHostViewUnsafeOutput(view) {
		result = productionAdapterObjectiveCloseoutFailureReviewPacketBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if result.FailureReviewPacketRef == "" {
		result = productionAdapterObjectiveCloseoutFailureReviewPacketBlock(result, FailureEvidenceMissing, "objective_closeout_failure_review_packet_ref_missing", "host:objective_closeout_failure_review_ref", "provide_objective_closeout_failure_review_ref")
	}
	if !view.ReadyForFailureReview {
		result = productionAdapterObjectiveCloseoutFailureReviewPacketBlock(result, firstFailureClass(view.FailureClass, FailureEvidenceMissing), "objective_closeout_failure_review_not_ready", "host:objective_closeout_failure_review", firstNextHostAction(view.NextHostAction, "review_objective_closeout_durable_failure"))
	}
	if result.FailureRef == "" {
		result = productionAdapterObjectiveCloseoutFailureReviewPacketBlock(result, FailureEvidenceMissing, "objective_closeout_failure_ref_missing", "host:objective_closeout_failure_ref", "provide_objective_closeout_failure_ref")
	}
	if len(result.MissingInputs) == 0 && len(result.BlockedReasons) == 0 {
		result.Status = "ready_for_objective_closeout_failure_review"
		result.ReadyForHostDisplay = true
		result.ReadyForFailureReview = true
		result.ReadyForCompensationReview = result.CompensationRef != ""
		result.NextHostAction = "review_objective_closeout_durable_failure"
		result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_objective_closeout_failure_review")
	}
	return result.Normalize()
}

// agentx-api: internal_candidate
func BuildProductionAdapterObjectiveCloseoutFailureReviewBlackboxFixture(input ProductionAdapterObjectiveCloseoutFailureReviewBlackboxFixtureInput) ProductionAdapterObjectiveCloseoutFailureReviewBlackboxFixture {
	if productionAdapterObjectiveCloseoutFailureReviewPacketEmpty(input.FailureReviewPacket) {
		return unavailableProductionAdapterObjectiveCloseoutFailureReviewBlackboxFixture()
	}
	packet := input.FailureReviewPacket.Normalize()
	result := ProductionAdapterObjectiveCloseoutFailureReviewBlackboxFixture{
		ContractVersion:                 ContractVersion,
		Projected:                       true,
		Available:                       packet.Available,
		Status:                          "blocked",
		Mode:                            "production_adapter_objective_closeout_failure_review_blackbox_fixture",
		FixtureRef:                      normalizeOneDisplaySafeRef(input.FixtureRef),
		FailureReviewPacketRef:          packet.FailureReviewPacketRef,
		HostViewRef:                     packet.HostViewRef,
		DisplaySections:                 productionAdapterObjectiveCloseoutFailureReviewDisplaySections(),
		ObjectiveLifecycleClosed:        false,
		ObjectiveSatisfied:              false,
		HostDurableWriteReported:        packet.HostDurableWriteReported,
		HostDurableWriteSucceeded:       packet.HostDurableWriteSucceeded,
		HostDurableWriteFailed:          packet.HostDurableWriteFailed,
		ObjectiveCloseoutPacketRef:      packet.ObjectiveCloseoutPacketRef,
		ObjectiveCloseoutHandoffRef:     packet.ObjectiveCloseoutHandoffRef,
		ObjectiveCloseoutReadbackRef:    packet.ObjectiveCloseoutReadbackRef,
		ObjectiveRef:                    packet.ObjectiveRef,
		HostObjectiveLifecycleRef:       packet.HostObjectiveLifecycleRef,
		HostRunstoreRef:                 packet.HostRunstoreRef,
		ExpectedDurableEventRef:         packet.ExpectedDurableEventRef,
		ExpectedObjectiveStateRef:       packet.ExpectedObjectiveStateRef,
		FailureRef:                      packet.FailureRef,
		CompensationRef:                 packet.CompensationRef,
		HostDurableApplyConfirmationRef: packet.HostDurableApplyConfirmationRef,
		CompletionAuditResultBindingRef: packet.CompletionAuditResultBindingRef,
		AdapterRef:                      packet.AdapterRef,
		MissingInputs:                   cloneMissingInputs(packet.MissingInputs),
		BlockedReasons:                  cloneStringSlice(packet.BlockedReasons),
		FailureClass:                    packet.FailureClass,
		Boundaries:                      productionAdapterObjectiveCloseoutFailureReviewBlackboxFixtureBoundaries(packet.Boundaries),
		NextHostAction:                  packet.NextHostAction,
		RunnerEffect:                    "none",
		PromptEffect:                    "none",
		RawOutputLoaded:                 input.RawOutputLoaded || packet.RawOutputLoaded,
	}
	if input.RawOutputLoaded || displaySafeRefRejected(input.FixtureRef) || productionAdapterObjectiveCloseoutFailureReviewPacketUnsafeOutput(packet) {
		result = productionAdapterObjectiveCloseoutFailureReviewBlackboxFixtureBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if result.FixtureRef == "" {
		result = productionAdapterObjectiveCloseoutFailureReviewBlackboxFixtureBlock(result, FailureEvidenceMissing, "objective_closeout_failure_fixture_ref_missing", "host:objective_closeout_failure_fixture_ref", "provide_objective_closeout_failure_fixture")
	}
	if !packet.ReadyForFailureReview {
		result = productionAdapterObjectiveCloseoutFailureReviewBlackboxFixtureBlock(result, firstFailureClass(packet.FailureClass, FailureEvidenceMissing), "objective_closeout_failure_review_packet_not_ready", "host:objective_closeout_failure_review_packet", firstNextHostAction(packet.NextHostAction, "review_objective_closeout_durable_failure"))
	}
	if len(result.MissingInputs) == 0 && len(result.BlockedReasons) == 0 {
		result.Status = "ready_for_objective_closeout_failure_display"
		result.ReadyForHostDisplay = true
		result.ReadyForFailureReview = true
		result.ReadyForCompensationReview = packet.ReadyForCompensationReview
		result.NextHostAction = "review_objective_closeout_durable_failure"
		result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_objective_closeout_failure_display", "host_cli_objective_closeout_failure_display_ready")
	}
	return result.Normalize()
}

func CloneProductionAdapterObjectiveCloseoutFailureReviewPacket(in ProductionAdapterObjectiveCloseoutFailureReviewPacket) ProductionAdapterObjectiveCloseoutFailureReviewPacket {
	out := in
	out.DisplaySections = cloneStringSlice(in.DisplaySections)
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (p ProductionAdapterObjectiveCloseoutFailureReviewPacket) Clone() ProductionAdapterObjectiveCloseoutFailureReviewPacket {
	return CloneProductionAdapterObjectiveCloseoutFailureReviewPacket(p)
}

func (p ProductionAdapterObjectiveCloseoutFailureReviewPacket) Normalize() ProductionAdapterObjectiveCloseoutFailureReviewPacket {
	out := CloneProductionAdapterObjectiveCloseoutFailureReviewPacket(p)
	unsafe := productionAdapterObjectiveCloseoutFailureReviewPacketUnsafeOutput(out)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = normalizeControlToken(out.Status)
	out.Mode = normalizeControlToken(out.Mode)
	if out.Mode == "" {
		out.Mode = "production_adapter_objective_closeout_failure_review_packet"
	}
	out.FailureReviewPacketRef = normalizeOneDisplaySafeRef(out.FailureReviewPacketRef)
	out.HostViewRef = normalizeOneDisplaySafeRef(out.HostViewRef)
	out.DisplaySections = normalizeControlTokenList(out.DisplaySections)
	out.ObjectiveCloseoutPacketRef = normalizeOneDisplaySafeRef(out.ObjectiveCloseoutPacketRef)
	out.ObjectiveCloseoutHandoffRef = normalizeOneDisplaySafeRef(out.ObjectiveCloseoutHandoffRef)
	out.ObjectiveCloseoutReadbackRef = normalizeOneDisplaySafeRef(out.ObjectiveCloseoutReadbackRef)
	out.ObjectiveRef = normalizeOneDisplaySafeRef(out.ObjectiveRef)
	out.HostObjectiveLifecycleRef = normalizeOneDisplaySafeRef(out.HostObjectiveLifecycleRef)
	out.HostRunstoreRef = normalizeOneDisplaySafeRef(out.HostRunstoreRef)
	out.ExpectedDurableEventRef = normalizeOneDisplaySafeRef(out.ExpectedDurableEventRef)
	out.ExpectedObjectiveStateRef = normalizeOneDisplaySafeRef(out.ExpectedObjectiveStateRef)
	out.FailureRef = normalizeOneDisplaySafeRef(out.FailureRef)
	out.CompensationRef = normalizeOneDisplaySafeRef(out.CompensationRef)
	out.HostDurableApplyConfirmationRef = normalizeOneDisplaySafeRef(out.HostDurableApplyConfirmationRef)
	out.CompletionAuditResultBindingRef = normalizeOneDisplaySafeRef(out.CompletionAuditResultBindingRef)
	out.AdapterRef = normalizeOneDisplaySafeRef(out.AdapterRef)
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
	if !out.Available {
		out.Status = "unavailable"
		out.ReadyForHostDisplay = false
		out.ReadyForFailureReview = false
		out.ReadyForCompensationReview = false
	}
	if out.Status == "" {
		out.Status = "blocked"
	}
	if unsafe || out.RawOutputLoaded {
		out.RawOutputLoaded = true
		out.Status = "blocked"
		out.ReadyForHostDisplay = false
		out.ReadyForFailureReview = false
		out.ReadyForCompensationReview = false
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
	out.ObjectiveLifecycleClosed = false
	out.ObjectiveSatisfied = false
	out.CoreInvocationExecuted = false
	out.DurableWriteByCore = false
	out.ObjectiveStoreWriteByCore = false
	out.RunstoreWriteByCore = false
	out.ReadyForHostDisplay = out.ReadyForHostDisplay &&
		out.Available &&
		out.FailureReviewPacketRef != "" &&
		out.HostViewRef != "" &&
		out.ObjectiveCloseoutPacketRef != "" &&
		out.ObjectiveCloseoutReadbackRef != "" &&
		out.ObjectiveRef != "" &&
		!out.RawOutputLoaded
	out.ReadyForFailureReview = out.ReadyForFailureReview &&
		out.ReadyForHostDisplay &&
		out.Status == "ready_for_objective_closeout_failure_review" &&
		out.HostDurableWriteReported &&
		out.HostDurableWriteFailed &&
		!out.HostDurableWriteSucceeded &&
		out.FailureRef != "" &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.RawOutputLoaded
	out.ReadyForCompensationReview = out.ReadyForCompensationReview &&
		out.ReadyForFailureReview &&
		out.CompensationRef != "" &&
		!out.RawOutputLoaded
	return out
}

func CloneProductionAdapterObjectiveCloseoutFailureReviewBlackboxFixture(in ProductionAdapterObjectiveCloseoutFailureReviewBlackboxFixture) ProductionAdapterObjectiveCloseoutFailureReviewBlackboxFixture {
	out := in
	out.DisplaySections = cloneStringSlice(in.DisplaySections)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (f ProductionAdapterObjectiveCloseoutFailureReviewBlackboxFixture) Clone() ProductionAdapterObjectiveCloseoutFailureReviewBlackboxFixture {
	return CloneProductionAdapterObjectiveCloseoutFailureReviewBlackboxFixture(f)
}

func (f ProductionAdapterObjectiveCloseoutFailureReviewBlackboxFixture) Normalize() ProductionAdapterObjectiveCloseoutFailureReviewBlackboxFixture {
	out := CloneProductionAdapterObjectiveCloseoutFailureReviewBlackboxFixture(f)
	unsafe := productionAdapterObjectiveCloseoutFailureReviewBlackboxFixtureUnsafeOutput(out)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = normalizeControlToken(out.Status)
	out.Mode = normalizeControlToken(out.Mode)
	if out.Mode == "" {
		out.Mode = "production_adapter_objective_closeout_failure_review_blackbox_fixture"
	}
	out.FixtureRef = normalizeOneDisplaySafeRef(out.FixtureRef)
	out.FailureReviewPacketRef = normalizeOneDisplaySafeRef(out.FailureReviewPacketRef)
	out.HostViewRef = normalizeOneDisplaySafeRef(out.HostViewRef)
	out.DisplaySections = normalizeControlTokenList(out.DisplaySections)
	out.ObjectiveCloseoutPacketRef = normalizeOneDisplaySafeRef(out.ObjectiveCloseoutPacketRef)
	out.ObjectiveCloseoutHandoffRef = normalizeOneDisplaySafeRef(out.ObjectiveCloseoutHandoffRef)
	out.ObjectiveCloseoutReadbackRef = normalizeOneDisplaySafeRef(out.ObjectiveCloseoutReadbackRef)
	out.ObjectiveRef = normalizeOneDisplaySafeRef(out.ObjectiveRef)
	out.HostObjectiveLifecycleRef = normalizeOneDisplaySafeRef(out.HostObjectiveLifecycleRef)
	out.HostRunstoreRef = normalizeOneDisplaySafeRef(out.HostRunstoreRef)
	out.ExpectedDurableEventRef = normalizeOneDisplaySafeRef(out.ExpectedDurableEventRef)
	out.ExpectedObjectiveStateRef = normalizeOneDisplaySafeRef(out.ExpectedObjectiveStateRef)
	out.FailureRef = normalizeOneDisplaySafeRef(out.FailureRef)
	out.CompensationRef = normalizeOneDisplaySafeRef(out.CompensationRef)
	out.HostDurableApplyConfirmationRef = normalizeOneDisplaySafeRef(out.HostDurableApplyConfirmationRef)
	out.CompletionAuditResultBindingRef = normalizeOneDisplaySafeRef(out.CompletionAuditResultBindingRef)
	out.AdapterRef = normalizeOneDisplaySafeRef(out.AdapterRef)
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
		out.ReadyForFailureReview = false
		out.ReadyForCompensationReview = false
	}
	if out.Status == "" {
		out.Status = "blocked"
	}
	if unsafe || out.RawOutputLoaded {
		out.RawOutputLoaded = true
		out.Status = "blocked"
		out.ReadyForHostDisplay = false
		out.ReadyForFailureReview = false
		out.ReadyForCompensationReview = false
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
	out.ObjectiveLifecycleClosed = false
	out.ObjectiveSatisfied = false
	out.CoreInvocationExecuted = false
	out.DurableWriteByCore = false
	out.ObjectiveStoreWriteByCore = false
	out.RunstoreWriteByCore = false
	out.ReadyForHostDisplay = out.ReadyForHostDisplay &&
		out.Available &&
		out.FixtureRef != "" &&
		out.FailureReviewPacketRef != "" &&
		out.HostViewRef != "" &&
		out.ObjectiveCloseoutPacketRef != "" &&
		out.ObjectiveCloseoutReadbackRef != "" &&
		out.ObjectiveRef != "" &&
		!out.RawOutputLoaded
	out.ReadyForFailureReview = out.ReadyForFailureReview &&
		out.ReadyForHostDisplay &&
		out.Status == "ready_for_objective_closeout_failure_display" &&
		out.HostDurableWriteReported &&
		out.HostDurableWriteFailed &&
		!out.HostDurableWriteSucceeded &&
		out.FailureRef != "" &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.RawOutputLoaded
	out.ReadyForCompensationReview = out.ReadyForCompensationReview &&
		out.ReadyForFailureReview &&
		out.CompensationRef != "" &&
		!out.RawOutputLoaded
	return out
}

func productionAdapterObjectiveCloseoutFailureReviewPacketBlock(result ProductionAdapterObjectiveCloseoutFailureReviewPacket, failure FailureClass, reason string, missing MissingInput, next NextHostAction) ProductionAdapterObjectiveCloseoutFailureReviewPacket {
	result.Status = "blocked"
	result.ReadyForHostDisplay = false
	result.ReadyForFailureReview = false
	result.ReadyForCompensationReview = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = NormalizeNextHostAction(string(next))
	if result.NextHostAction == "" {
		result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	}
	result.Boundaries = AppendBoundaries(result.Boundaries, "objective_closeout_failure_review_packet_blocked")
	return result
}

func productionAdapterObjectiveCloseoutFailureReviewBlackboxFixtureBlock(result ProductionAdapterObjectiveCloseoutFailureReviewBlackboxFixture, failure FailureClass, reason string, missing MissingInput, next NextHostAction) ProductionAdapterObjectiveCloseoutFailureReviewBlackboxFixture {
	result.Status = "blocked"
	result.ReadyForHostDisplay = false
	result.ReadyForFailureReview = false
	result.ReadyForCompensationReview = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = NormalizeNextHostAction(string(next))
	if result.NextHostAction == "" {
		result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	}
	result.Boundaries = AppendBoundaries(result.Boundaries, "objective_closeout_failure_review_blackbox_fixture_blocked")
	return result
}

func productionAdapterObjectiveCloseoutFailureReviewDisplaySections() []string {
	return append(productionAdapterObjectiveCloseoutDisplaySections(), "objective_closeout_failure_review", "objective_closeout_compensation_review")
}

func productionAdapterObjectiveCloseoutFailureReviewPacketBoundaries(viewBoundaries []Boundary) []Boundary {
	return MergeBoundaries(
		[]Boundary{
			"production_adapter_objective_closeout_failure_review_packet",
			"objective_closeout_failure_review_projection_only",
			"host_owned_objective_closeout_failure_review",
			"host_cli_objective_closeout_failure_display",
			"failure_compensation_review_only",
			"compensation_not_executed",
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

func productionAdapterObjectiveCloseoutFailureReviewBlackboxFixtureBoundaries(packetBoundaries []Boundary) []Boundary {
	return MergeBoundaries(
		[]Boundary{
			"production_adapter_objective_closeout_failure_review_blackbox_fixture",
			"objective_closeout_failure_review_fixture_projection_only",
			"host_cli_objective_closeout_failure_display",
			"host_owned_objective_closeout_failure_review",
			"failure_compensation_review_only",
			"compensation_not_executed",
			"display_safe_refs_only",
			"display_safe_result_refs_only",
			"no_runner_dispatch",
			"no_durable_write_by_core",
			"no_objective_store_write_by_core",
			"no_runstore_write_by_core",
		},
		packetBoundaries,
	)
}

func productionAdapterObjectiveCloseoutFailureReviewPacketUnsafeOutput(input ProductionAdapterObjectiveCloseoutFailureReviewPacket) bool {
	return displaySafeRefRejected(input.FailureReviewPacketRef) ||
		displaySafeRefRejected(input.HostViewRef) ||
		displaySafeRefRejected(input.ObjectiveCloseoutPacketRef) ||
		displaySafeRefRejected(input.ObjectiveCloseoutHandoffRef) ||
		displaySafeRefRejected(input.ObjectiveCloseoutReadbackRef) ||
		displaySafeRefRejected(input.ObjectiveRef) ||
		displaySafeRefRejected(input.HostObjectiveLifecycleRef) ||
		displaySafeRefRejected(input.HostRunstoreRef) ||
		displaySafeRefRejected(input.ExpectedDurableEventRef) ||
		displaySafeRefRejected(input.ExpectedObjectiveStateRef) ||
		displaySafeRefRejected(input.FailureRef) ||
		displaySafeRefRejected(input.CompensationRef) ||
		displaySafeRefRejected(input.HostDurableApplyConfirmationRef) ||
		displaySafeRefRejected(input.CompletionAuditResultBindingRef) ||
		displaySafeRefRejected(input.AdapterRef) ||
		evidenceRefsRejected(input.EvidenceRefs) ||
		input.RawOutputLoaded
}

func productionAdapterObjectiveCloseoutFailureReviewBlackboxFixtureUnsafeOutput(input ProductionAdapterObjectiveCloseoutFailureReviewBlackboxFixture) bool {
	return displaySafeRefRejected(input.FixtureRef) ||
		displaySafeRefRejected(input.FailureReviewPacketRef) ||
		displaySafeRefRejected(input.HostViewRef) ||
		displaySafeRefRejected(input.ObjectiveCloseoutPacketRef) ||
		displaySafeRefRejected(input.ObjectiveCloseoutHandoffRef) ||
		displaySafeRefRejected(input.ObjectiveCloseoutReadbackRef) ||
		displaySafeRefRejected(input.ObjectiveRef) ||
		displaySafeRefRejected(input.HostObjectiveLifecycleRef) ||
		displaySafeRefRejected(input.HostRunstoreRef) ||
		displaySafeRefRejected(input.ExpectedDurableEventRef) ||
		displaySafeRefRejected(input.ExpectedObjectiveStateRef) ||
		displaySafeRefRejected(input.FailureRef) ||
		displaySafeRefRejected(input.CompensationRef) ||
		displaySafeRefRejected(input.HostDurableApplyConfirmationRef) ||
		displaySafeRefRejected(input.CompletionAuditResultBindingRef) ||
		displaySafeRefRejected(input.AdapterRef) ||
		input.RawOutputLoaded
}

func productionAdapterObjectiveCloseoutFailureReviewPacketEmpty(packet ProductionAdapterObjectiveCloseoutFailureReviewPacket) bool {
	return !packet.Projected &&
		!packet.Available &&
		packet.Status == "" &&
		packet.Mode == "" &&
		packet.FailureReviewPacketRef == "" &&
		packet.HostViewRef == "" &&
		packet.ObjectiveCloseoutPacketRef == "" &&
		packet.ObjectiveRef == "" &&
		len(packet.MissingInputs) == 0 &&
		len(packet.BlockedReasons) == 0 &&
		len(packet.Boundaries) == 0 &&
		packet.NextHostAction == "" &&
		!packet.RawOutputLoaded
}

func unavailableProductionAdapterObjectiveCloseoutFailureReviewPacket() ProductionAdapterObjectiveCloseoutFailureReviewPacket {
	return ProductionAdapterObjectiveCloseoutFailureReviewPacket{
		ContractVersion: ContractVersion,
		Projected:       true,
		Available:       false,
		Status:          "unavailable",
		Mode:            "production_adapter_objective_closeout_failure_review_packet",
		DisplaySections: productionAdapterObjectiveCloseoutFailureReviewDisplaySections(),
		Boundaries: []Boundary{
			"production_adapter_objective_closeout_failure_review_packet",
			"objective_closeout_failure_review_projection_only",
			"host_cli_objective_closeout_failure_display",
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

func unavailableProductionAdapterObjectiveCloseoutFailureReviewBlackboxFixture() ProductionAdapterObjectiveCloseoutFailureReviewBlackboxFixture {
	return ProductionAdapterObjectiveCloseoutFailureReviewBlackboxFixture{
		ContractVersion: ContractVersion,
		Projected:       true,
		Available:       false,
		Status:          "unavailable",
		Mode:            "production_adapter_objective_closeout_failure_review_blackbox_fixture",
		DisplaySections: productionAdapterObjectiveCloseoutFailureReviewDisplaySections(),
		Boundaries: []Boundary{
			"production_adapter_objective_closeout_failure_review_blackbox_fixture",
			"objective_closeout_failure_review_fixture_projection_only",
			"host_cli_objective_closeout_failure_display",
			"display_safe_refs_only",
			"no_runner_dispatch",
			"no_durable_write_by_core",
			"no_objective_store_write_by_core",
			"no_runstore_write_by_core",
		},
		RunnerEffect:   "none",
		PromptEffect:   "none",
		NextHostAction: "provide_objective_closeout_failure_review_packet",
	}
}
