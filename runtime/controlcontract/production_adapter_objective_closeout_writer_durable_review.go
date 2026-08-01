package controlcontract

type ProductionAdapterObjectiveCloseoutWriterDurableReviewPacketInput struct {
	ReviewPacketRef DisplaySafeRef                                          `json:"review_packet_ref,omitempty"`
	DurableRequest  ProductionAdapterObjectiveCloseoutWriterDurableRequest  `json:"durable_request,omitempty"`
	DurableResult   ProductionAdapterObjectiveCloseoutWriterDurableResult   `json:"durable_result,omitempty"`
	DurableReadback ProductionAdapterObjectiveCloseoutWriterDurableReadback `json:"durable_readback,omitempty"`
	RawOutputLoaded bool                                                    `json:"raw_output_loaded"`
}

type ProductionAdapterObjectiveCloseoutWriterDurableReviewPacket struct {
	ContractVersion                 string           `json:"contract_version,omitempty"`
	Projected                       bool             `json:"projected"`
	Available                       bool             `json:"available"`
	Status                          string           `json:"status,omitempty"`
	Mode                            string           `json:"mode,omitempty"`
	DisplayState                    string           `json:"display_state,omitempty"`
	DisplayStage                    string           `json:"display_stage,omitempty"`
	DisplaySections                 []string         `json:"display_sections,omitempty"`
	ReadyForHostDisplay             bool             `json:"ready_for_host_display"`
	ReadyForHostDurableWrite        bool             `json:"ready_for_host_durable_write"`
	ReadyForDurableResultReview     bool             `json:"ready_for_durable_result_review"`
	ReadyForDurableReadbackReview   bool             `json:"ready_for_durable_readback_review"`
	ReadyForFailureReview           bool             `json:"ready_for_failure_review"`
	ReadyForCompensationReview      bool             `json:"ready_for_compensation_review"`
	ReadyForBlockedReview           bool             `json:"ready_for_blocked_review"`
	ReadyForObjectiveReturn         bool             `json:"ready_for_objective_return"`
	HostMayExecuteDurableWrite      bool             `json:"host_may_execute_durable_write"`
	HostDurableWriteAuthorized      bool             `json:"host_durable_write_authorized"`
	HostDurableWriteReported        bool             `json:"host_durable_write_reported"`
	HostDurableWriteSucceeded       bool             `json:"host_durable_write_succeeded"`
	HostDurableWriteFailed          bool             `json:"host_durable_write_failed"`
	HostDurableWriteRecorded        bool             `json:"host_durable_write_recorded"`
	WriterDurableReadbackBound      bool             `json:"writer_durable_readback_bound"`
	ObjectiveLifecycleClosed        bool             `json:"objective_lifecycle_closed"`
	ObjectiveSatisfied              bool             `json:"objective_satisfied"`
	CoreInvocationExecuted          bool             `json:"core_invocation_executed"`
	DryRunByCore                    bool             `json:"dry_run_by_core"`
	DurableWriteByCore              bool             `json:"durable_write_by_core"`
	ObjectiveStoreWriteByCore       bool             `json:"objective_store_write_by_core"`
	RunstoreWriteByCore             bool             `json:"runstore_write_by_core"`
	ReviewPacketRef                 DisplaySafeRef   `json:"review_packet_ref,omitempty"`
	DurableRequestRef               DisplaySafeRef   `json:"durable_request_ref,omitempty"`
	DurableResultRef                DisplaySafeRef   `json:"durable_result_ref,omitempty"`
	DurableReadbackRef              DisplaySafeRef   `json:"durable_readback_ref,omitempty"`
	WriterOptInRef                  DisplaySafeRef   `json:"writer_opt_in_ref,omitempty"`
	WriterRef                       DisplaySafeRef   `json:"writer_ref,omitempty"`
	OwnerRef                        DisplaySafeRef   `json:"owner_ref,omitempty"`
	HostWriterBindingRef            DisplaySafeRef   `json:"host_writer_binding_ref,omitempty"`
	HostAdapterRunRef               DisplaySafeRef   `json:"host_adapter_run_ref,omitempty"`
	DryRunSmokeRef                  DisplaySafeRef   `json:"dry_run_smoke_ref,omitempty"`
	DryRunResultRef                 DisplaySafeRef   `json:"dry_run_result_ref,omitempty"`
	ExpectedDurableResultRef        DisplaySafeRef   `json:"expected_durable_result_ref,omitempty"`
	ExpectedReadbackRef             DisplaySafeRef   `json:"expected_readback_ref,omitempty"`
	ObjectiveCloseoutReadbackRef    DisplaySafeRef   `json:"objective_closeout_readback_ref,omitempty"`
	ObjectiveCloseoutHandoffRef     DisplaySafeRef   `json:"objective_closeout_handoff_ref,omitempty"`
	HostUIHandoffRef                DisplaySafeRef   `json:"host_ui_handoff_ref,omitempty"`
	ObjectiveCloseoutPacketRef      DisplaySafeRef   `json:"objective_closeout_packet_ref,omitempty"`
	ObjectiveRef                    DisplaySafeRef   `json:"objective_ref,omitempty"`
	HostObjectiveLifecycleRef       DisplaySafeRef   `json:"host_objective_lifecycle_ref,omitempty"`
	HostRunstoreRef                 DisplaySafeRef   `json:"host_runstore_ref,omitempty"`
	ExpectedDurableEventRef         DisplaySafeRef   `json:"expected_durable_event_ref,omitempty"`
	ExpectedObjectiveStateRef       DisplaySafeRef   `json:"expected_objective_state_ref,omitempty"`
	AppliedDurableEventRef          DisplaySafeRef   `json:"applied_durable_event_ref,omitempty"`
	AppliedRunstoreRef              DisplaySafeRef   `json:"applied_runstore_ref,omitempty"`
	AppliedObjectiveStateRef        DisplaySafeRef   `json:"applied_objective_state_ref,omitempty"`
	FailureRef                      DisplaySafeRef   `json:"failure_ref,omitempty"`
	CompensationRef                 DisplaySafeRef   `json:"compensation_ref,omitempty"`
	HostDurableWriteConfirmationRef DisplaySafeRef   `json:"host_durable_write_confirmation_ref,omitempty"`
	IdempotencyRef                  DisplaySafeRef   `json:"idempotency_ref,omitempty"`
	IdempotencyContractRef          DisplaySafeRef   `json:"idempotency_contract_ref,omitempty"`
	ReadbackContractRef             DisplaySafeRef   `json:"readback_contract_ref,omitempty"`
	RollbackReviewRef               DisplaySafeRef   `json:"rollback_review_ref,omitempty"`
	CompensationReviewRef           DisplaySafeRef   `json:"compensation_review_ref,omitempty"`
	DurableEvidenceRefs             []DisplaySafeRef `json:"durable_evidence_refs,omitempty"`
	MissingInputs                   []MissingInput   `json:"missing_inputs,omitempty"`
	BlockedReasons                  []string         `json:"blocked_reasons,omitempty"`
	FailureClass                    FailureClass     `json:"failure_class,omitempty"`
	Boundaries                      []Boundary       `json:"boundaries,omitempty"`
	NextHostAction                  NextHostAction   `json:"next_host_action,omitempty"`
	RunnerEffect                    string           `json:"runner_effect,omitempty"`
	PromptEffect                    string           `json:"prompt_effect,omitempty"`
	RawOutputLoaded                 bool             `json:"raw_output_loaded"`
}

// agentx-api: internal_candidate
type ProductionAdapterObjectiveCloseoutWriterDurableReviewBlackboxFixtureInput struct {
	FixtureRef      DisplaySafeRef                                              `json:"fixture_ref,omitempty"`
	ReviewPacket    ProductionAdapterObjectiveCloseoutWriterDurableReviewPacket `json:"review_packet,omitempty"`
	RawOutputLoaded bool                                                        `json:"raw_output_loaded"`
}

// agentx-api: internal_candidate
type ProductionAdapterObjectiveCloseoutWriterDurableReviewBlackboxFixture struct {
	ContractVersion               string         `json:"contract_version,omitempty"`
	Projected                     bool           `json:"projected"`
	Available                     bool           `json:"available"`
	Status                        string         `json:"status,omitempty"`
	Mode                          string         `json:"mode,omitempty"`
	DisplayState                  string         `json:"display_state,omitempty"`
	DisplayStage                  string         `json:"display_stage,omitempty"`
	DisplaySections               []string       `json:"display_sections,omitempty"`
	ReadyForHostDisplay           bool           `json:"ready_for_host_display"`
	ReadyForHostDurableWrite      bool           `json:"ready_for_host_durable_write"`
	ReadyForDurableResultReview   bool           `json:"ready_for_durable_result_review"`
	ReadyForDurableReadbackReview bool           `json:"ready_for_durable_readback_review"`
	ReadyForFailureReview         bool           `json:"ready_for_failure_review"`
	ReadyForCompensationReview    bool           `json:"ready_for_compensation_review"`
	ReadyForBlockedReview         bool           `json:"ready_for_blocked_review"`
	ReadyForObjectiveReturn       bool           `json:"ready_for_objective_return"`
	HostMayExecuteDurableWrite    bool           `json:"host_may_execute_durable_write"`
	HostDurableWriteAuthorized    bool           `json:"host_durable_write_authorized"`
	HostDurableWriteReported      bool           `json:"host_durable_write_reported"`
	HostDurableWriteSucceeded     bool           `json:"host_durable_write_succeeded"`
	HostDurableWriteFailed        bool           `json:"host_durable_write_failed"`
	HostDurableWriteRecorded      bool           `json:"host_durable_write_recorded"`
	WriterDurableReadbackBound    bool           `json:"writer_durable_readback_bound"`
	ObjectiveLifecycleClosed      bool           `json:"objective_lifecycle_closed"`
	ObjectiveSatisfied            bool           `json:"objective_satisfied"`
	CoreInvocationExecuted        bool           `json:"core_invocation_executed"`
	DryRunByCore                  bool           `json:"dry_run_by_core"`
	DurableWriteByCore            bool           `json:"durable_write_by_core"`
	ObjectiveStoreWriteByCore     bool           `json:"objective_store_write_by_core"`
	RunstoreWriteByCore           bool           `json:"runstore_write_by_core"`
	FixtureRef                    DisplaySafeRef `json:"fixture_ref,omitempty"`
	ReviewPacketRef               DisplaySafeRef `json:"review_packet_ref,omitempty"`
	DurableRequestRef             DisplaySafeRef `json:"durable_request_ref,omitempty"`
	DurableResultRef              DisplaySafeRef `json:"durable_result_ref,omitempty"`
	DurableReadbackRef            DisplaySafeRef `json:"durable_readback_ref,omitempty"`
	WriterOptInRef                DisplaySafeRef `json:"writer_opt_in_ref,omitempty"`
	WriterRef                     DisplaySafeRef `json:"writer_ref,omitempty"`
	HostWriterBindingRef          DisplaySafeRef `json:"host_writer_binding_ref,omitempty"`
	HostAdapterRunRef             DisplaySafeRef `json:"host_adapter_run_ref,omitempty"`
	ExpectedReadbackRef           DisplaySafeRef `json:"expected_readback_ref,omitempty"`
	ObjectiveCloseoutReadbackRef  DisplaySafeRef `json:"objective_closeout_readback_ref,omitempty"`
	ObjectiveCloseoutHandoffRef   DisplaySafeRef `json:"objective_closeout_handoff_ref,omitempty"`
	ObjectiveCloseoutPacketRef    DisplaySafeRef `json:"objective_closeout_packet_ref,omitempty"`
	ObjectiveRef                  DisplaySafeRef `json:"objective_ref,omitempty"`
	HostRunstoreRef               DisplaySafeRef `json:"host_runstore_ref,omitempty"`
	ExpectedDurableEventRef       DisplaySafeRef `json:"expected_durable_event_ref,omitempty"`
	ExpectedObjectiveStateRef     DisplaySafeRef `json:"expected_objective_state_ref,omitempty"`
	AppliedDurableEventRef        DisplaySafeRef `json:"applied_durable_event_ref,omitempty"`
	AppliedRunstoreRef            DisplaySafeRef `json:"applied_runstore_ref,omitempty"`
	AppliedObjectiveStateRef      DisplaySafeRef `json:"applied_objective_state_ref,omitempty"`
	FailureRef                    DisplaySafeRef `json:"failure_ref,omitempty"`
	CompensationRef               DisplaySafeRef `json:"compensation_ref,omitempty"`
	MissingInputs                 []MissingInput `json:"missing_inputs,omitempty"`
	BlockedReasons                []string       `json:"blocked_reasons,omitempty"`
	FailureClass                  FailureClass   `json:"failure_class,omitempty"`
	Boundaries                    []Boundary     `json:"boundaries,omitempty"`
	NextHostAction                NextHostAction `json:"next_host_action,omitempty"`
	RunnerEffect                  string         `json:"runner_effect,omitempty"`
	PromptEffect                  string         `json:"prompt_effect,omitempty"`
	RawOutputLoaded               bool           `json:"raw_output_loaded"`
}

func BuildProductionAdapterObjectiveCloseoutWriterDurableReviewPacket(input ProductionAdapterObjectiveCloseoutWriterDurableReviewPacketInput) ProductionAdapterObjectiveCloseoutWriterDurableReviewPacket {
	requestProvided := !productionAdapterObjectiveCloseoutWriterDurableRequestEmpty(input.DurableRequest)
	resultProvided := !productionAdapterObjectiveCloseoutWriterDurableResultEmpty(input.DurableResult)
	readbackProvided := !productionAdapterObjectiveCloseoutWriterDurableReadbackEmpty(input.DurableReadback)
	if !requestProvided && !resultProvided && !readbackProvided {
		return unavailableProductionAdapterObjectiveCloseoutWriterDurableReviewPacket()
	}
	request := ProductionAdapterObjectiveCloseoutWriterDurableRequest{}
	if requestProvided {
		request = input.DurableRequest.Normalize()
	}
	resultProjection := ProductionAdapterObjectiveCloseoutWriterDurableResult{}
	if resultProvided {
		resultProjection = input.DurableResult.Normalize()
	}
	readback := ProductionAdapterObjectiveCloseoutWriterDurableReadback{}
	if readbackProvided {
		readback = input.DurableReadback.Normalize()
	}
	result := ProductionAdapterObjectiveCloseoutWriterDurableReviewPacket{
		ContractVersion:                 ContractVersion,
		Projected:                       true,
		Available:                       productionAdapterObjectiveCloseoutWriterDurableReviewAvailable(request, requestProvided, resultProjection, resultProvided, readback, readbackProvided),
		Status:                          "blocked",
		Mode:                            "production_adapter_objective_closeout_writer_durable_review_packet",
		DisplayState:                    "blocked",
		DisplayStage:                    "blocked",
		DisplaySections:                 productionAdapterObjectiveCloseoutWriterDurableReviewDisplaySections(),
		ReviewPacketRef:                 normalizeOneDisplaySafeRef(input.ReviewPacketRef),
		DurableRequestRef:               firstDisplaySafeRef(readback.DurableRequestRef, resultProjection.DurableRequestRef, request.DurableRequestRef),
		DurableResultRef:                firstDisplaySafeRef(readback.DurableResultRef, resultProjection.DurableResultRef),
		DurableReadbackRef:              readback.DurableReadbackRef,
		WriterOptInRef:                  firstDisplaySafeRef(readback.WriterOptInRef, resultProjection.WriterOptInRef, request.WriterOptInRef),
		WriterRef:                       firstDisplaySafeRef(readback.WriterRef, resultProjection.WriterRef, request.WriterRef),
		OwnerRef:                        request.OwnerRef,
		HostWriterBindingRef:            firstDisplaySafeRef(resultProjection.HostWriterBindingRef, request.HostWriterBindingRef),
		HostAdapterRunRef:               firstDisplaySafeRef(readback.HostAdapterRunRef, resultProjection.HostAdapterRunRef),
		DryRunSmokeRef:                  firstDisplaySafeRef(readback.DryRunSmokeRef, resultProjection.DryRunSmokeRef, request.DryRunSmokeRef),
		DryRunResultRef:                 firstDisplaySafeRef(readback.DryRunResultRef, resultProjection.DryRunResultRef, request.DryRunResultRef),
		ExpectedDurableResultRef:        firstDisplaySafeRef(resultProjection.ExpectedDurableResultRef, request.ExpectedDurableResultRef),
		ExpectedReadbackRef:             firstDisplaySafeRef(readback.ExpectedReadbackRef, resultProjection.ExpectedReadbackRef, request.ExpectedReadbackRef),
		ObjectiveCloseoutReadbackRef:    readback.ObjectiveCloseoutReadbackRef,
		ObjectiveCloseoutHandoffRef:     firstDisplaySafeRef(readback.ObjectiveCloseoutHandoffRef, resultProjection.ObjectiveCloseoutHandoffRef, request.ObjectiveCloseoutHandoffRef),
		HostUIHandoffRef:                request.HostUIHandoffRef,
		ObjectiveCloseoutPacketRef:      firstDisplaySafeRef(readback.ObjectiveCloseoutPacketRef, resultProjection.ObjectiveCloseoutPacketRef, request.ObjectiveCloseoutPacketRef),
		ObjectiveRef:                    firstDisplaySafeRef(readback.ObjectiveRef, resultProjection.ObjectiveRef, request.ObjectiveRef),
		HostObjectiveLifecycleRef:       request.HostObjectiveLifecycleRef,
		HostRunstoreRef:                 firstDisplaySafeRef(readback.HostRunstoreRef, resultProjection.HostRunstoreRef, request.HostRunstoreRef),
		ExpectedDurableEventRef:         firstDisplaySafeRef(readback.ExpectedDurableEventRef, resultProjection.ExpectedDurableEventRef, request.ExpectedDurableEventRef),
		ExpectedObjectiveStateRef:       firstDisplaySafeRef(readback.ExpectedObjectiveStateRef, resultProjection.ExpectedObjectiveStateRef, request.ExpectedObjectiveStateRef),
		AppliedDurableEventRef:          firstDisplaySafeRef(readback.AppliedDurableEventRef, resultProjection.AppliedDurableEventRef),
		AppliedRunstoreRef:              firstDisplaySafeRef(readback.AppliedRunstoreRef, resultProjection.AppliedRunstoreRef),
		AppliedObjectiveStateRef:        firstDisplaySafeRef(readback.AppliedObjectiveStateRef, resultProjection.AppliedObjectiveStateRef),
		FailureRef:                      resultProjection.FailureRef,
		CompensationRef:                 resultProjection.CompensationRef,
		HostDurableWriteConfirmationRef: request.HostDurableWriteConfirmationRef,
		IdempotencyRef:                  firstDisplaySafeRef(resultProjection.IdempotencyRef, request.IdempotencyRef),
		IdempotencyContractRef:          firstDisplaySafeRef(resultProjection.IdempotencyContractRef, request.IdempotencyContractRef),
		ReadbackContractRef:             firstDisplaySafeRef(resultProjection.ReadbackContractRef, request.ReadbackContractRef),
		RollbackReviewRef:               firstDisplaySafeRef(resultProjection.RollbackReviewRef, request.RollbackReviewRef),
		CompensationReviewRef:           firstDisplaySafeRef(resultProjection.CompensationReviewRef, request.CompensationReviewRef),
		DurableEvidenceRefs:             cloneDisplaySafeRefs(resultProjection.DurableEvidenceRefs),
		MissingInputs:                   productionAdapterObjectiveCloseoutWriterDurableReviewMissingInputs(request, requestProvided, resultProjection, resultProvided, readback, readbackProvided),
		BlockedReasons:                  productionAdapterObjectiveCloseoutWriterDurableReviewBlockedReasons(request, requestProvided, resultProjection, resultProvided, readback, readbackProvided),
		FailureClass:                    productionAdapterObjectiveCloseoutWriterDurableReviewFailureClass(request, requestProvided, resultProjection, resultProvided, readback, readbackProvided),
		Boundaries:                      productionAdapterObjectiveCloseoutWriterDurableReviewPacketBoundaries(request.Boundaries, resultProjection.Boundaries, readback.Boundaries),
		NextHostAction:                  productionAdapterObjectiveCloseoutWriterDurableReviewNextAction(request, requestProvided, resultProjection, resultProvided, readback, readbackProvided),
		RunnerEffect:                    "none",
		PromptEffect:                    "none",
		RawOutputLoaded:                 input.RawOutputLoaded || request.RawOutputLoaded || resultProjection.RawOutputLoaded || readback.RawOutputLoaded,
		HostDurableWriteAuthorized:      request.HostDurableWriteAuthorized,
		HostMayExecuteDurableWrite:      request.HostMayExecuteDurableWrite,
		HostDurableWriteReported:        resultProjection.HostDurableWriteReported,
		HostDurableWriteSucceeded:       resultProjection.HostDurableWriteSucceeded,
		HostDurableWriteFailed:          resultProjection.HostDurableWriteFailed,
		HostDurableWriteRecorded:        resultProjection.HostDurableWriteRecorded,
		WriterDurableReadbackBound:      readback.WriterDurableReadbackBound,
		ObjectiveLifecycleClosed:        readback.ObjectiveLifecycleClosed,
		ObjectiveSatisfied:              readback.ObjectiveSatisfied,
	}
	if productionAdapterObjectiveCloseoutWriterDurableReviewPacketUnsafe(input, request, requestProvided, resultProjection, resultProvided, readback, readbackProvided) {
		result = productionAdapterObjectiveCloseoutWriterDurableReviewPacketBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if result.ReviewPacketRef == "" {
		result = productionAdapterObjectiveCloseoutWriterDurableReviewPacketBlock(result, FailureEvidenceMissing, "writer_durable_review_packet_ref_missing", "host:objective_closeout_writer_durable_review_packet_ref", "provide_objective_closeout_writer_durable_review_packet")
		return result.Normalize()
	}
	for _, mismatch := range productionAdapterObjectiveCloseoutWriterDurableReviewMismatches(request, requestProvided, resultProjection, resultProvided, readback, readbackProvided) {
		result = productionAdapterObjectiveCloseoutWriterDurableReviewPacketBlock(result, FailureVerificationFailed, mismatch.reason, mismatch.missing, "review_objective_closeout_writer_durable_review")
		return result.Normalize()
	}
	switch {
	case readbackProvided && readback.WriterDurableReadbackBound && readback.ReadyForObjectiveReturn:
		result.Status = "ready_for_objective_closeout_writer_durable_return_review"
		result.DisplayState = "durable_readback_bound"
		result.DisplayStage = "final"
		result.ReadyForHostDisplay = true
		result.ReadyForObjectiveReturn = true
		result.NextHostAction = "return_objective_closed_lifecycle"
		result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_objective_closeout_writer_durable_return_review", "objective_closeout_writer_durable_review_final")
	case resultProvided && resultProjection.ReadyForWriterDurableReadback:
		result.Status = "ready_for_objective_closeout_writer_durable_readback_review"
		result.DisplayState = "durable_write_recorded"
		result.DisplayStage = "readback_review"
		result.ReadyForHostDisplay = true
		result.ReadyForDurableResultReview = true
		result.ReadyForDurableReadbackReview = true
		result.NextHostAction = "bind_objective_closeout_writer_durable_readback"
		result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_objective_closeout_writer_durable_readback_review")
	case resultProvided && resultProjection.HostDurableWriteFailed:
		if result.FailureRef == "" {
			result = productionAdapterObjectiveCloseoutWriterDurableReviewPacketBlock(result, FailureEvidenceMissing, "writer_durable_failure_ref_missing", "host:objective_closeout_writer_durable_failure_ref", "provide_objective_closeout_writer_durable_failure")
			return result.Normalize()
		}
		result.Status = "ready_for_objective_closeout_writer_durable_failure_review"
		result.DisplayState = "durable_write_failed"
		result.DisplayStage = "failure_review"
		result.ReadyForHostDisplay = true
		result.ReadyForFailureReview = true
		result.ReadyForCompensationReview = result.CompensationRef != ""
		result.NextHostAction = "review_objective_closeout_writer_durable_failure"
		result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_objective_closeout_writer_durable_failure_review", "failure_compensation_review_only", "compensation_not_executed")
	case requestProvided && request.ReadyForHostDurableWrite && request.HostMayExecuteDurableWrite:
		result.Status = "ready_for_objective_closeout_writer_durable_write_review"
		result.DisplayState = "durable_write_ready"
		result.DisplayStage = "durable_write_review"
		result.ReadyForHostDisplay = true
		result.ReadyForHostDurableWrite = true
		result.HostDurableWriteAuthorized = true
		result.HostMayExecuteDurableWrite = true
		result.NextHostAction = "host_may_execute_objective_closeout_durable_writer_adapter"
		result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_objective_closeout_writer_durable_write_review", "host_may_execute_durable_writer", "core_durable_write_not_executed")
	default:
		result.Status = "ready_for_objective_closeout_writer_durable_blocked_review"
		result.DisplayState = productionAdapterObjectiveCloseoutWriterDurableReviewBlockedState(result)
		result.DisplayStage = "blocked_review"
		result.ReadyForHostDisplay = true
		result.ReadyForBlockedReview = true
		result.HostMayExecuteDurableWrite = false
		result.ReadyForHostDurableWrite = false
		result.NextHostAction = firstNextHostAction(result.NextHostAction, "review_objective_closeout_writer_durable_blocked")
		result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_objective_closeout_writer_durable_blocked_review")
		if len(result.BlockedReasons) == 0 {
			result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, "objective_closeout_writer_durable_not_ready")
		}
		if result.FailureClass == FailureNone {
			result.FailureClass = FailureEvidenceMissing
		}
	}
	return result.Normalize()
}

// agentx-api: internal_candidate
func BuildProductionAdapterObjectiveCloseoutWriterDurableReviewBlackboxFixture(input ProductionAdapterObjectiveCloseoutWriterDurableReviewBlackboxFixtureInput) ProductionAdapterObjectiveCloseoutWriterDurableReviewBlackboxFixture {
	if productionAdapterObjectiveCloseoutWriterDurableReviewPacketEmpty(input.ReviewPacket) {
		return unavailableProductionAdapterObjectiveCloseoutWriterDurableReviewBlackboxFixture()
	}
	packet := input.ReviewPacket.Normalize()
	result := ProductionAdapterObjectiveCloseoutWriterDurableReviewBlackboxFixture{
		ContractVersion:               ContractVersion,
		Projected:                     true,
		Available:                     packet.Available,
		Status:                        "blocked",
		Mode:                          "production_adapter_objective_closeout_writer_durable_review_blackbox_fixture",
		DisplayState:                  packet.DisplayState,
		DisplayStage:                  packet.DisplayStage,
		DisplaySections:               productionAdapterObjectiveCloseoutWriterDurableReviewDisplaySections(),
		FixtureRef:                    normalizeOneDisplaySafeRef(input.FixtureRef),
		ReviewPacketRef:               packet.ReviewPacketRef,
		DurableRequestRef:             packet.DurableRequestRef,
		DurableResultRef:              packet.DurableResultRef,
		DurableReadbackRef:            packet.DurableReadbackRef,
		WriterOptInRef:                packet.WriterOptInRef,
		WriterRef:                     packet.WriterRef,
		HostWriterBindingRef:          packet.HostWriterBindingRef,
		HostAdapterRunRef:             packet.HostAdapterRunRef,
		ExpectedReadbackRef:           packet.ExpectedReadbackRef,
		ObjectiveCloseoutReadbackRef:  packet.ObjectiveCloseoutReadbackRef,
		ObjectiveCloseoutHandoffRef:   packet.ObjectiveCloseoutHandoffRef,
		ObjectiveCloseoutPacketRef:    packet.ObjectiveCloseoutPacketRef,
		ObjectiveRef:                  packet.ObjectiveRef,
		HostRunstoreRef:               packet.HostRunstoreRef,
		ExpectedDurableEventRef:       packet.ExpectedDurableEventRef,
		ExpectedObjectiveStateRef:     packet.ExpectedObjectiveStateRef,
		AppliedDurableEventRef:        packet.AppliedDurableEventRef,
		AppliedRunstoreRef:            packet.AppliedRunstoreRef,
		AppliedObjectiveStateRef:      packet.AppliedObjectiveStateRef,
		FailureRef:                    packet.FailureRef,
		CompensationRef:               packet.CompensationRef,
		MissingInputs:                 cloneMissingInputs(packet.MissingInputs),
		BlockedReasons:                cloneStringSlice(packet.BlockedReasons),
		FailureClass:                  packet.FailureClass,
		Boundaries:                    productionAdapterObjectiveCloseoutWriterDurableReviewBlackboxFixtureBoundaries(packet.Boundaries),
		NextHostAction:                packet.NextHostAction,
		RunnerEffect:                  "none",
		PromptEffect:                  "none",
		RawOutputLoaded:               input.RawOutputLoaded || packet.RawOutputLoaded,
		ReadyForHostDurableWrite:      packet.ReadyForHostDurableWrite,
		ReadyForDurableResultReview:   packet.ReadyForDurableResultReview,
		ReadyForDurableReadbackReview: packet.ReadyForDurableReadbackReview,
		ReadyForFailureReview:         packet.ReadyForFailureReview,
		ReadyForCompensationReview:    packet.ReadyForCompensationReview,
		ReadyForBlockedReview:         packet.ReadyForBlockedReview,
		ReadyForObjectiveReturn:       packet.ReadyForObjectiveReturn,
		HostMayExecuteDurableWrite:    packet.HostMayExecuteDurableWrite,
		HostDurableWriteAuthorized:    packet.HostDurableWriteAuthorized,
		HostDurableWriteReported:      packet.HostDurableWriteReported,
		HostDurableWriteSucceeded:     packet.HostDurableWriteSucceeded,
		HostDurableWriteFailed:        packet.HostDurableWriteFailed,
		HostDurableWriteRecorded:      packet.HostDurableWriteRecorded,
		WriterDurableReadbackBound:    packet.WriterDurableReadbackBound,
		ObjectiveLifecycleClosed:      packet.ObjectiveLifecycleClosed,
		ObjectiveSatisfied:            packet.ObjectiveSatisfied,
	}
	if input.RawOutputLoaded || displaySafeRefRejected(input.FixtureRef) || productionAdapterObjectiveCloseoutWriterDurableReviewPacketUnsafeOutput(packet) {
		result = productionAdapterObjectiveCloseoutWriterDurableReviewBlackboxFixtureBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if result.FixtureRef == "" {
		result = productionAdapterObjectiveCloseoutWriterDurableReviewBlackboxFixtureBlock(result, FailureEvidenceMissing, "writer_durable_review_fixture_ref_missing", "host:objective_closeout_writer_durable_review_fixture_ref", "provide_objective_closeout_writer_durable_review_fixture")
		return result.Normalize()
	}
	if !packet.ReadyForHostDisplay {
		result = productionAdapterObjectiveCloseoutWriterDurableReviewBlackboxFixtureBlock(result, firstFailureClass(packet.FailureClass, FailureEvidenceMissing), "writer_durable_review_packet_not_ready", "host:objective_closeout_writer_durable_review_packet", firstNextHostAction(packet.NextHostAction, "review_objective_closeout_writer_durable_review"))
		return result.Normalize()
	}
	result.Status = productionAdapterObjectiveCloseoutWriterDurableReviewFixtureStatus(packet)
	result.ReadyForHostDisplay = true
	result.Boundaries = AppendBoundaries(result.Boundaries, "host_cli_objective_closeout_writer_durable_display_ready")
	return result.Normalize()
}

func CloneProductionAdapterObjectiveCloseoutWriterDurableReviewPacket(in ProductionAdapterObjectiveCloseoutWriterDurableReviewPacket) ProductionAdapterObjectiveCloseoutWriterDurableReviewPacket {
	out := in
	out.DisplaySections = cloneStringSlice(in.DisplaySections)
	out.DurableEvidenceRefs = cloneDisplaySafeRefs(in.DurableEvidenceRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (p ProductionAdapterObjectiveCloseoutWriterDurableReviewPacket) Clone() ProductionAdapterObjectiveCloseoutWriterDurableReviewPacket {
	return CloneProductionAdapterObjectiveCloseoutWriterDurableReviewPacket(p)
}

func (p ProductionAdapterObjectiveCloseoutWriterDurableReviewPacket) Normalize() ProductionAdapterObjectiveCloseoutWriterDurableReviewPacket {
	out := CloneProductionAdapterObjectiveCloseoutWriterDurableReviewPacket(p)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = normalizeControlToken(out.Status)
	out.Mode = normalizeControlToken(out.Mode)
	if out.Mode == "" {
		out.Mode = "production_adapter_objective_closeout_writer_durable_review_packet"
	}
	out.DisplayState = normalizeControlToken(out.DisplayState)
	if out.DisplayState == "" {
		out.DisplayState = "blocked"
	}
	out.DisplayStage = normalizeControlToken(out.DisplayStage)
	if out.DisplayStage == "" {
		out.DisplayStage = "blocked"
	}
	out.DisplaySections = normalizeControlTokenList(out.DisplaySections)
	out.ReviewPacketRef = normalizeOneDisplaySafeRef(out.ReviewPacketRef)
	out.DurableRequestRef = normalizeOneDisplaySafeRef(out.DurableRequestRef)
	out.DurableResultRef = normalizeOneDisplaySafeRef(out.DurableResultRef)
	out.DurableReadbackRef = normalizeOneDisplaySafeRef(out.DurableReadbackRef)
	out.WriterOptInRef = normalizeOneDisplaySafeRef(out.WriterOptInRef)
	out.WriterRef = normalizeOneDisplaySafeRef(out.WriterRef)
	out.OwnerRef = normalizeOneDisplaySafeRef(out.OwnerRef)
	out.HostWriterBindingRef = normalizeOneDisplaySafeRef(out.HostWriterBindingRef)
	out.HostAdapterRunRef = normalizeOneDisplaySafeRef(out.HostAdapterRunRef)
	out.DryRunSmokeRef = normalizeOneDisplaySafeRef(out.DryRunSmokeRef)
	out.DryRunResultRef = normalizeOneDisplaySafeRef(out.DryRunResultRef)
	out.ExpectedDurableResultRef = normalizeOneDisplaySafeRef(out.ExpectedDurableResultRef)
	out.ExpectedReadbackRef = normalizeOneDisplaySafeRef(out.ExpectedReadbackRef)
	out.ObjectiveCloseoutReadbackRef = normalizeOneDisplaySafeRef(out.ObjectiveCloseoutReadbackRef)
	out.ObjectiveCloseoutHandoffRef = normalizeOneDisplaySafeRef(out.ObjectiveCloseoutHandoffRef)
	out.HostUIHandoffRef = normalizeOneDisplaySafeRef(out.HostUIHandoffRef)
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
	out.HostDurableWriteConfirmationRef = normalizeOneDisplaySafeRef(out.HostDurableWriteConfirmationRef)
	out.IdempotencyRef = normalizeOneDisplaySafeRef(out.IdempotencyRef)
	out.IdempotencyContractRef = normalizeOneDisplaySafeRef(out.IdempotencyContractRef)
	out.ReadbackContractRef = normalizeOneDisplaySafeRef(out.ReadbackContractRef)
	out.RollbackReviewRef = normalizeOneDisplaySafeRef(out.RollbackReviewRef)
	out.CompensationReviewRef = normalizeOneDisplaySafeRef(out.CompensationReviewRef)
	out.DurableEvidenceRefs = normalizeDisplaySafeRefs(out.DurableEvidenceRefs)
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
	}
	if out.RawOutputLoaded || productionAdapterObjectiveCloseoutWriterDurableReviewPacketUnsafeOutput(out) {
		out.RawOutputLoaded = true
		out.Status = "blocked"
		out.DisplayState = "blocked_unsafe_refs"
		out.DisplayStage = "blocked"
		out.ReadyForHostDisplay = false
		out.ReadyForHostDurableWrite = false
		out.ReadyForDurableResultReview = false
		out.ReadyForDurableReadbackReview = false
		out.ReadyForFailureReview = false
		out.ReadyForCompensationReview = false
		out.ReadyForBlockedReview = false
		out.ReadyForObjectiveReturn = false
		out.HostMayExecuteDurableWrite = false
		out.FailureClass = firstFailureClass(out.FailureClass, FailureEvidenceWeak)
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		if out.NextHostAction == "" {
			out.NextHostAction = "provide_display_safe_refs"
		}
	}
	out.CoreInvocationExecuted = false
	out.DryRunByCore = false
	out.DurableWriteByCore = false
	out.ObjectiveStoreWriteByCore = false
	out.RunstoreWriteByCore = false
	out.ReadyForHostDisplay = out.ReadyForHostDisplay &&
		out.Available &&
		out.ReviewPacketRef != "" &&
		out.WriterRef != "" &&
		!out.RawOutputLoaded
	out.ReadyForHostDurableWrite = out.ReadyForHostDurableWrite &&
		out.ReadyForHostDisplay &&
		out.Status == "ready_for_objective_closeout_writer_durable_write_review" &&
		out.HostDurableWriteAuthorized &&
		out.HostMayExecuteDurableWrite &&
		out.DurableRequestRef != "" &&
		out.HostDurableWriteConfirmationRef != "" &&
		len(out.BlockedReasons) == 0 &&
		!out.RawOutputLoaded
	out.ReadyForDurableResultReview = out.ReadyForDurableResultReview &&
		out.ReadyForHostDisplay &&
		out.DurableResultRef != "" &&
		out.HostDurableWriteRecorded &&
		out.HostDurableWriteSucceeded &&
		!out.HostDurableWriteFailed &&
		len(out.BlockedReasons) == 0 &&
		!out.RawOutputLoaded
	out.ReadyForDurableReadbackReview = out.ReadyForDurableReadbackReview &&
		out.ReadyForDurableResultReview &&
		out.Status == "ready_for_objective_closeout_writer_durable_readback_review" &&
		!out.RawOutputLoaded
	out.ReadyForFailureReview = out.ReadyForFailureReview &&
		out.ReadyForHostDisplay &&
		out.Status == "ready_for_objective_closeout_writer_durable_failure_review" &&
		out.HostDurableWriteReported &&
		out.HostDurableWriteFailed &&
		!out.HostDurableWriteSucceeded &&
		out.FailureRef != "" &&
		len(out.BlockedReasons) > 0 &&
		!out.RawOutputLoaded
	out.ReadyForCompensationReview = out.ReadyForCompensationReview &&
		out.ReadyForFailureReview &&
		out.CompensationRef != "" &&
		!out.RawOutputLoaded
	out.ReadyForBlockedReview = out.ReadyForBlockedReview &&
		out.ReadyForHostDisplay &&
		out.Status == "ready_for_objective_closeout_writer_durable_blocked_review" &&
		len(out.BlockedReasons) > 0 &&
		!out.RawOutputLoaded
	out.ReadyForObjectiveReturn = out.ReadyForObjectiveReturn &&
		out.ReadyForHostDisplay &&
		out.Status == "ready_for_objective_closeout_writer_durable_return_review" &&
		out.WriterDurableReadbackBound &&
		out.ObjectiveLifecycleClosed &&
		out.ObjectiveSatisfied &&
		out.DurableReadbackRef != "" &&
		out.ObjectiveCloseoutReadbackRef != "" &&
		len(out.BlockedReasons) == 0 &&
		!out.RawOutputLoaded
	if !out.ReadyForHostDurableWrite {
		out.HostMayExecuteDurableWrite = false
	}
	return out
}

func CloneProductionAdapterObjectiveCloseoutWriterDurableReviewBlackboxFixture(in ProductionAdapterObjectiveCloseoutWriterDurableReviewBlackboxFixture) ProductionAdapterObjectiveCloseoutWriterDurableReviewBlackboxFixture {
	out := in
	out.DisplaySections = cloneStringSlice(in.DisplaySections)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (f ProductionAdapterObjectiveCloseoutWriterDurableReviewBlackboxFixture) Clone() ProductionAdapterObjectiveCloseoutWriterDurableReviewBlackboxFixture {
	return CloneProductionAdapterObjectiveCloseoutWriterDurableReviewBlackboxFixture(f)
}

func (f ProductionAdapterObjectiveCloseoutWriterDurableReviewBlackboxFixture) Normalize() ProductionAdapterObjectiveCloseoutWriterDurableReviewBlackboxFixture {
	out := CloneProductionAdapterObjectiveCloseoutWriterDurableReviewBlackboxFixture(f)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = normalizeControlToken(out.Status)
	out.Mode = normalizeControlToken(out.Mode)
	if out.Mode == "" {
		out.Mode = "production_adapter_objective_closeout_writer_durable_review_blackbox_fixture"
	}
	out.DisplayState = normalizeControlToken(out.DisplayState)
	if out.DisplayState == "" {
		out.DisplayState = "blocked"
	}
	out.DisplayStage = normalizeControlToken(out.DisplayStage)
	if out.DisplayStage == "" {
		out.DisplayStage = "blocked"
	}
	out.DisplaySections = normalizeControlTokenList(out.DisplaySections)
	out.FixtureRef = normalizeOneDisplaySafeRef(out.FixtureRef)
	out.ReviewPacketRef = normalizeOneDisplaySafeRef(out.ReviewPacketRef)
	out.DurableRequestRef = normalizeOneDisplaySafeRef(out.DurableRequestRef)
	out.DurableResultRef = normalizeOneDisplaySafeRef(out.DurableResultRef)
	out.DurableReadbackRef = normalizeOneDisplaySafeRef(out.DurableReadbackRef)
	out.WriterOptInRef = normalizeOneDisplaySafeRef(out.WriterOptInRef)
	out.WriterRef = normalizeOneDisplaySafeRef(out.WriterRef)
	out.HostWriterBindingRef = normalizeOneDisplaySafeRef(out.HostWriterBindingRef)
	out.HostAdapterRunRef = normalizeOneDisplaySafeRef(out.HostAdapterRunRef)
	out.ExpectedReadbackRef = normalizeOneDisplaySafeRef(out.ExpectedReadbackRef)
	out.ObjectiveCloseoutReadbackRef = normalizeOneDisplaySafeRef(out.ObjectiveCloseoutReadbackRef)
	out.ObjectiveCloseoutHandoffRef = normalizeOneDisplaySafeRef(out.ObjectiveCloseoutHandoffRef)
	out.ObjectiveCloseoutPacketRef = normalizeOneDisplaySafeRef(out.ObjectiveCloseoutPacketRef)
	out.ObjectiveRef = normalizeOneDisplaySafeRef(out.ObjectiveRef)
	out.HostRunstoreRef = normalizeOneDisplaySafeRef(out.HostRunstoreRef)
	out.ExpectedDurableEventRef = normalizeOneDisplaySafeRef(out.ExpectedDurableEventRef)
	out.ExpectedObjectiveStateRef = normalizeOneDisplaySafeRef(out.ExpectedObjectiveStateRef)
	out.AppliedDurableEventRef = normalizeOneDisplaySafeRef(out.AppliedDurableEventRef)
	out.AppliedRunstoreRef = normalizeOneDisplaySafeRef(out.AppliedRunstoreRef)
	out.AppliedObjectiveStateRef = normalizeOneDisplaySafeRef(out.AppliedObjectiveStateRef)
	out.FailureRef = normalizeOneDisplaySafeRef(out.FailureRef)
	out.CompensationRef = normalizeOneDisplaySafeRef(out.CompensationRef)
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
	}
	if out.RawOutputLoaded || productionAdapterObjectiveCloseoutWriterDurableReviewBlackboxFixtureUnsafeOutput(out) {
		out.RawOutputLoaded = true
		out.Status = "blocked"
		out.DisplayState = "blocked_unsafe_refs"
		out.DisplayStage = "blocked"
		out.ReadyForHostDisplay = false
		out.ReadyForHostDurableWrite = false
		out.ReadyForDurableResultReview = false
		out.ReadyForDurableReadbackReview = false
		out.ReadyForFailureReview = false
		out.ReadyForCompensationReview = false
		out.ReadyForBlockedReview = false
		out.ReadyForObjectiveReturn = false
		out.HostMayExecuteDurableWrite = false
		out.FailureClass = firstFailureClass(out.FailureClass, FailureEvidenceWeak)
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		if out.NextHostAction == "" {
			out.NextHostAction = "provide_display_safe_refs"
		}
	}
	out.CoreInvocationExecuted = false
	out.DryRunByCore = false
	out.DurableWriteByCore = false
	out.ObjectiveStoreWriteByCore = false
	out.RunstoreWriteByCore = false
	out.ReadyForHostDisplay = out.ReadyForHostDisplay &&
		out.Available &&
		out.FixtureRef != "" &&
		out.ReviewPacketRef != "" &&
		out.WriterRef != "" &&
		!out.RawOutputLoaded
	out.ReadyForHostDurableWrite = out.ReadyForHostDurableWrite && out.ReadyForHostDisplay && out.Status == "ready_for_objective_closeout_writer_durable_write_display"
	out.ReadyForDurableResultReview = out.ReadyForDurableResultReview && out.ReadyForHostDisplay && out.DurableResultRef != "" && out.HostDurableWriteRecorded
	out.ReadyForDurableReadbackReview = out.ReadyForDurableReadbackReview && out.ReadyForDurableResultReview && out.Status == "ready_for_objective_closeout_writer_durable_readback_display"
	out.ReadyForFailureReview = out.ReadyForFailureReview && out.ReadyForHostDisplay && out.Status == "ready_for_objective_closeout_writer_durable_failure_display" && out.FailureRef != ""
	out.ReadyForCompensationReview = out.ReadyForCompensationReview && out.ReadyForFailureReview && out.CompensationRef != ""
	out.ReadyForBlockedReview = out.ReadyForBlockedReview && out.ReadyForHostDisplay && out.Status == "ready_for_objective_closeout_writer_durable_blocked_display"
	out.ReadyForObjectiveReturn = out.ReadyForObjectiveReturn && out.ReadyForHostDisplay && out.Status == "ready_for_objective_closeout_writer_durable_return_display" && out.WriterDurableReadbackBound
	if !out.ReadyForHostDurableWrite {
		out.HostMayExecuteDurableWrite = false
	}
	return out
}

func productionAdapterObjectiveCloseoutWriterDurableReviewPacketBlock(result ProductionAdapterObjectiveCloseoutWriterDurableReviewPacket, failure FailureClass, reason string, missing MissingInput, next NextHostAction) ProductionAdapterObjectiveCloseoutWriterDurableReviewPacket {
	result.Status = "blocked"
	result.DisplayState = "blocked"
	result.DisplayStage = "blocked"
	result.ReadyForHostDisplay = false
	result.ReadyForHostDurableWrite = false
	result.ReadyForDurableResultReview = false
	result.ReadyForDurableReadbackReview = false
	result.ReadyForFailureReview = false
	result.ReadyForCompensationReview = false
	result.ReadyForBlockedReview = false
	result.ReadyForObjectiveReturn = false
	result.HostMayExecuteDurableWrite = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	result.Boundaries = AppendBoundaries(result.Boundaries, "objective_closeout_writer_durable_review_packet_blocked")
	return result
}

func productionAdapterObjectiveCloseoutWriterDurableReviewBlackboxFixtureBlock(result ProductionAdapterObjectiveCloseoutWriterDurableReviewBlackboxFixture, failure FailureClass, reason string, missing MissingInput, next NextHostAction) ProductionAdapterObjectiveCloseoutWriterDurableReviewBlackboxFixture {
	result.Status = "blocked"
	result.DisplayState = "blocked"
	result.DisplayStage = "blocked"
	result.ReadyForHostDisplay = false
	result.ReadyForHostDurableWrite = false
	result.ReadyForDurableResultReview = false
	result.ReadyForDurableReadbackReview = false
	result.ReadyForFailureReview = false
	result.ReadyForCompensationReview = false
	result.ReadyForBlockedReview = false
	result.ReadyForObjectiveReturn = false
	result.HostMayExecuteDurableWrite = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	result.Boundaries = AppendBoundaries(result.Boundaries, "objective_closeout_writer_durable_review_fixture_blocked")
	return result
}

func productionAdapterObjectiveCloseoutWriterDurableReviewDisplaySections() []string {
	return []string{
		"objective_closeout_writer_durable_summary",
		"objective_closeout_writer_durable_request_refs",
		"objective_closeout_writer_durable_result_refs",
		"objective_closeout_writer_durable_readback_refs",
		"objective_closeout_writer_durable_failure_review",
		"objective_closeout_writer_durable_next_action",
	}
}

func productionAdapterObjectiveCloseoutWriterDurableReviewPacketBoundaries(groups ...[]Boundary) []Boundary {
	return MergeBoundaries(
		[]Boundary{
			"production_adapter_objective_closeout_writer_durable_review_packet",
			"objective_closeout_writer_durable_review_packet_projection_only",
			"host_owned_objective_closeout_writer_adapter",
			"host_cli_objective_closeout_writer_durable_display",
			"durable_writer_review_only",
			"display_safe_refs_only",
			"display_safe_result_refs_only",
			"no_runner_dispatch",
			"no_dry_run_by_core",
			"no_durable_write_by_core",
			"no_objective_store_write_by_core",
			"no_runstore_write_by_core",
		},
		MergeBoundaries(groups...),
	)
}

func productionAdapterObjectiveCloseoutWriterDurableReviewBlackboxFixtureBoundaries(groups ...[]Boundary) []Boundary {
	return MergeBoundaries(
		[]Boundary{
			"production_adapter_objective_closeout_writer_durable_review_blackbox_fixture",
			"objective_closeout_writer_durable_review_fixture_projection_only",
			"host_owned_objective_closeout_writer_adapter",
			"host_cli_objective_closeout_writer_durable_display",
			"durable_writer_review_only",
			"display_safe_refs_only",
			"display_safe_result_refs_only",
			"no_runner_dispatch",
			"no_dry_run_by_core",
			"no_durable_write_by_core",
			"no_objective_store_write_by_core",
			"no_runstore_write_by_core",
		},
		MergeBoundaries(groups...),
	)
}

func productionAdapterObjectiveCloseoutWriterDurableReviewAvailable(request ProductionAdapterObjectiveCloseoutWriterDurableRequest, requestProvided bool, result ProductionAdapterObjectiveCloseoutWriterDurableResult, resultProvided bool, readback ProductionAdapterObjectiveCloseoutWriterDurableReadback, readbackProvided bool) bool {
	available := false
	if requestProvided {
		available = available || request.Available
	}
	if resultProvided {
		available = available || result.Available
	}
	if readbackProvided {
		available = available || readback.Available
	}
	return available
}

func productionAdapterObjectiveCloseoutWriterDurableReviewMissingInputs(request ProductionAdapterObjectiveCloseoutWriterDurableRequest, requestProvided bool, result ProductionAdapterObjectiveCloseoutWriterDurableResult, resultProvided bool, readback ProductionAdapterObjectiveCloseoutWriterDurableReadback, readbackProvided bool) []MissingInput {
	var out []MissingInput
	if requestProvided {
		out = AppendMissingInputs(out, request.MissingInputs...)
	}
	if resultProvided {
		out = AppendMissingInputs(out, result.MissingInputs...)
	}
	if readbackProvided {
		out = AppendMissingInputs(out, readback.MissingInputs...)
	}
	return out
}

func productionAdapterObjectiveCloseoutWriterDurableReviewBlockedReasons(request ProductionAdapterObjectiveCloseoutWriterDurableRequest, requestProvided bool, result ProductionAdapterObjectiveCloseoutWriterDurableResult, resultProvided bool, readback ProductionAdapterObjectiveCloseoutWriterDurableReadback, readbackProvided bool) []string {
	var out []string
	if requestProvided {
		out = append(out, request.BlockedReasons...)
	}
	if resultProvided {
		out = append(out, result.BlockedReasons...)
	}
	if readbackProvided {
		out = append(out, readback.BlockedReasons...)
	}
	return normalizeControlTokenList(out)
}

func productionAdapterObjectiveCloseoutWriterDurableReviewFailureClass(request ProductionAdapterObjectiveCloseoutWriterDurableRequest, requestProvided bool, result ProductionAdapterObjectiveCloseoutWriterDurableResult, resultProvided bool, readback ProductionAdapterObjectiveCloseoutWriterDurableReadback, readbackProvided bool) FailureClass {
	var values []FailureClass
	if requestProvided {
		values = append(values, request.FailureClass)
	}
	if resultProvided {
		values = append(values, result.FailureClass)
	}
	if readbackProvided {
		values = append(values, readback.FailureClass)
	}
	return firstFailureClass(values...)
}

func productionAdapterObjectiveCloseoutWriterDurableReviewNextAction(request ProductionAdapterObjectiveCloseoutWriterDurableRequest, requestProvided bool, result ProductionAdapterObjectiveCloseoutWriterDurableResult, resultProvided bool, readback ProductionAdapterObjectiveCloseoutWriterDurableReadback, readbackProvided bool) NextHostAction {
	if readbackProvided && readback.NextHostAction != "" {
		return readback.NextHostAction
	}
	if resultProvided && result.NextHostAction != "" {
		return result.NextHostAction
	}
	if requestProvided {
		return request.NextHostAction
	}
	return ""
}

func productionAdapterObjectiveCloseoutWriterDurableReviewBlockedState(packet ProductionAdapterObjectiveCloseoutWriterDurableReviewPacket) string {
	switch {
	case len(packet.MissingInputs) > 0:
		return "blocked_missing_inputs"
	case productionAdapterObjectiveCloseoutWriterDurableReviewHasReason(packet.BlockedReasons, "unsafe_input_ref"):
		return "blocked_unsafe_refs"
	case productionAdapterObjectiveCloseoutWriterDurableReviewHasReason(packet.BlockedReasons, "objective_closeout_writer_durable_request_not_ready"):
		return "blocked_request_not_ready"
	case productionAdapterObjectiveCloseoutWriterDurableReviewHasReason(packet.BlockedReasons, "objective_closeout_writer_durable_result_not_ready"):
		return "blocked_result_not_ready"
	case productionAdapterObjectiveCloseoutWriterDurableReviewHasReason(packet.BlockedReasons, "objective_closeout_readback_not_ready"):
		return "blocked_readback_not_ready"
	default:
		return "blocked"
	}
}

func productionAdapterObjectiveCloseoutWriterDurableReviewHasReason(values []string, want string) bool {
	want = normalizeControlToken(want)
	if want == "" {
		return false
	}
	for _, value := range normalizeControlTokenList(values) {
		if value == want {
			return true
		}
	}
	return false
}

func productionAdapterObjectiveCloseoutWriterDurableReviewFixtureStatus(packet ProductionAdapterObjectiveCloseoutWriterDurableReviewPacket) string {
	switch packet.Status {
	case "ready_for_objective_closeout_writer_durable_write_review":
		return "ready_for_objective_closeout_writer_durable_write_display"
	case "ready_for_objective_closeout_writer_durable_readback_review":
		return "ready_for_objective_closeout_writer_durable_readback_display"
	case "ready_for_objective_closeout_writer_durable_failure_review":
		return "ready_for_objective_closeout_writer_durable_failure_display"
	case "ready_for_objective_closeout_writer_durable_return_review":
		return "ready_for_objective_closeout_writer_durable_return_display"
	case "ready_for_objective_closeout_writer_durable_blocked_review":
		return "ready_for_objective_closeout_writer_durable_blocked_display"
	default:
		return "blocked"
	}
}

func productionAdapterObjectiveCloseoutWriterDurableReviewPacketUnsafe(input ProductionAdapterObjectiveCloseoutWriterDurableReviewPacketInput, request ProductionAdapterObjectiveCloseoutWriterDurableRequest, requestProvided bool, result ProductionAdapterObjectiveCloseoutWriterDurableResult, resultProvided bool, readback ProductionAdapterObjectiveCloseoutWriterDurableReadback, readbackProvided bool) bool {
	return input.RawOutputLoaded ||
		displaySafeRefRejected(input.ReviewPacketRef) ||
		(requestProvided && productionAdapterObjectiveCloseoutWriterDurableRequestOutputUnsafe(request)) ||
		(resultProvided && productionAdapterObjectiveCloseoutWriterDurableResultOutputUnsafe(result)) ||
		(readbackProvided && productionAdapterObjectiveCloseoutWriterDurableReadbackOutputUnsafe(readback))
}

func productionAdapterObjectiveCloseoutWriterDurableReviewPacketUnsafeOutput(input ProductionAdapterObjectiveCloseoutWriterDurableReviewPacket) bool {
	return displaySafeRefRejected(input.ReviewPacketRef) ||
		displaySafeRefRejected(input.DurableRequestRef) ||
		displaySafeRefRejected(input.DurableResultRef) ||
		displaySafeRefRejected(input.DurableReadbackRef) ||
		displaySafeRefRejected(input.WriterOptInRef) ||
		displaySafeRefRejected(input.WriterRef) ||
		displaySafeRefRejected(input.OwnerRef) ||
		displaySafeRefRejected(input.HostWriterBindingRef) ||
		displaySafeRefRejected(input.HostAdapterRunRef) ||
		displaySafeRefRejected(input.DryRunSmokeRef) ||
		displaySafeRefRejected(input.DryRunResultRef) ||
		displaySafeRefRejected(input.ExpectedDurableResultRef) ||
		displaySafeRefRejected(input.ExpectedReadbackRef) ||
		displaySafeRefRejected(input.ObjectiveCloseoutReadbackRef) ||
		displaySafeRefRejected(input.ObjectiveCloseoutHandoffRef) ||
		displaySafeRefRejected(input.HostUIHandoffRef) ||
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
		displaySafeRefRejected(input.HostDurableWriteConfirmationRef) ||
		displaySafeRefRejected(input.IdempotencyRef) ||
		displaySafeRefRejected(input.IdempotencyContractRef) ||
		displaySafeRefRejected(input.ReadbackContractRef) ||
		displaySafeRefRejected(input.RollbackReviewRef) ||
		displaySafeRefRejected(input.CompensationReviewRef) ||
		displaySafeRefSliceRejected(input.DurableEvidenceRefs) ||
		input.RawOutputLoaded
}

func productionAdapterObjectiveCloseoutWriterDurableReviewBlackboxFixtureUnsafeOutput(input ProductionAdapterObjectiveCloseoutWriterDurableReviewBlackboxFixture) bool {
	return displaySafeRefRejected(input.FixtureRef) ||
		displaySafeRefRejected(input.ReviewPacketRef) ||
		displaySafeRefRejected(input.DurableRequestRef) ||
		displaySafeRefRejected(input.DurableResultRef) ||
		displaySafeRefRejected(input.DurableReadbackRef) ||
		displaySafeRefRejected(input.WriterOptInRef) ||
		displaySafeRefRejected(input.WriterRef) ||
		displaySafeRefRejected(input.HostWriterBindingRef) ||
		displaySafeRefRejected(input.HostAdapterRunRef) ||
		displaySafeRefRejected(input.ExpectedReadbackRef) ||
		displaySafeRefRejected(input.ObjectiveCloseoutReadbackRef) ||
		displaySafeRefRejected(input.ObjectiveCloseoutHandoffRef) ||
		displaySafeRefRejected(input.ObjectiveCloseoutPacketRef) ||
		displaySafeRefRejected(input.ObjectiveRef) ||
		displaySafeRefRejected(input.HostRunstoreRef) ||
		displaySafeRefRejected(input.ExpectedDurableEventRef) ||
		displaySafeRefRejected(input.ExpectedObjectiveStateRef) ||
		displaySafeRefRejected(input.AppliedDurableEventRef) ||
		displaySafeRefRejected(input.AppliedRunstoreRef) ||
		displaySafeRefRejected(input.AppliedObjectiveStateRef) ||
		displaySafeRefRejected(input.FailureRef) ||
		displaySafeRefRejected(input.CompensationRef) ||
		input.RawOutputLoaded
}

func productionAdapterObjectiveCloseoutWriterDurableReviewMismatches(request ProductionAdapterObjectiveCloseoutWriterDurableRequest, requestProvided bool, result ProductionAdapterObjectiveCloseoutWriterDurableResult, resultProvided bool, readback ProductionAdapterObjectiveCloseoutWriterDurableReadback, readbackProvided bool) []productionAdapterObjectiveCloseoutWriterDurableMismatch {
	var out []productionAdapterObjectiveCloseoutWriterDurableMismatch
	if requestProvided && resultProvided {
		out = append(out, productionAdapterObjectiveCloseoutWriterDurableRefMismatch(request.DurableRequestRef, result.DurableRequestRef, "writer_durable_review_request_ref_mismatch", "host:objective_closeout_writer_durable_request_ref")...)
		out = append(out, productionAdapterObjectiveCloseoutWriterDurableRefMismatch(request.WriterOptInRef, result.WriterOptInRef, "writer_durable_review_opt_in_ref_mismatch", "host:objective_closeout_writer_opt_in_ref")...)
		out = append(out, productionAdapterObjectiveCloseoutWriterDurableRefMismatch(request.WriterRef, result.WriterRef, "writer_durable_review_writer_ref_mismatch", "host:objective_closeout_writer_ref")...)
		out = append(out, productionAdapterObjectiveCloseoutWriterDurableRefMismatch(request.ExpectedDurableResultRef, result.DurableResultRef, "writer_durable_review_result_ref_mismatch", "host:objective_closeout_writer_durable_result_ref")...)
		out = append(out, productionAdapterObjectiveCloseoutWriterDurableRefMismatch(request.ExpectedReadbackRef, result.ExpectedReadbackRef, "writer_durable_review_readback_ref_mismatch", "host:objective_closeout_writer_expected_readback_ref")...)
	}
	if resultProvided && readbackProvided {
		out = append(out, productionAdapterObjectiveCloseoutWriterDurableRefMismatch(result.DurableResultRef, readback.DurableResultRef, "writer_durable_review_readback_result_ref_mismatch", "host:objective_closeout_writer_durable_result_ref")...)
		out = append(out, productionAdapterObjectiveCloseoutWriterDurableRefMismatch(result.ExpectedReadbackRef, readback.ObjectiveCloseoutReadbackRef, "writer_durable_review_objective_readback_ref_mismatch", "host:objective_closeout_readback_ref")...)
		out = append(out, productionAdapterObjectiveCloseoutWriterDurableRefMismatch(result.ObjectiveCloseoutHandoffRef, readback.ObjectiveCloseoutHandoffRef, "writer_durable_review_handoff_ref_mismatch", "host:objective_closeout_handoff_ref")...)
		out = append(out, productionAdapterObjectiveCloseoutWriterDurableRefMismatch(result.ObjectiveCloseoutPacketRef, readback.ObjectiveCloseoutPacketRef, "writer_durable_review_packet_ref_mismatch", "host:objective_closeout_packet_ref")...)
		out = append(out, productionAdapterObjectiveCloseoutWriterDurableRefMismatch(result.ObjectiveRef, readback.ObjectiveRef, "writer_durable_review_objective_ref_mismatch", "host:objective_ref")...)
		out = append(out, productionAdapterObjectiveCloseoutWriterDurableRefMismatch(result.HostRunstoreRef, readback.AppliedRunstoreRef, "writer_durable_review_runstore_ref_mismatch", "host:runstore_ref")...)
		out = append(out, productionAdapterObjectiveCloseoutWriterDurableRefMismatch(result.ExpectedDurableEventRef, readback.AppliedDurableEventRef, "writer_durable_review_durable_event_ref_mismatch", "host:durable_event_ref")...)
		out = append(out, productionAdapterObjectiveCloseoutWriterDurableRefMismatch(result.ExpectedObjectiveStateRef, readback.AppliedObjectiveStateRef, "writer_durable_review_objective_state_ref_mismatch", "host:objective_state_ref")...)
	}
	return out
}

func productionAdapterObjectiveCloseoutWriterDurableReadbackEmpty(readback ProductionAdapterObjectiveCloseoutWriterDurableReadback) bool {
	return !readback.Projected &&
		!readback.Available &&
		readback.Status == "" &&
		readback.Mode == "" &&
		readback.DurableReadbackRef == "" &&
		readback.DurableResultRef == "" &&
		readback.DurableRequestRef == "" &&
		readback.WriterOptInRef == "" &&
		len(readback.MissingInputs) == 0 &&
		len(readback.BlockedReasons) == 0 &&
		len(readback.Boundaries) == 0 &&
		readback.NextHostAction == "" &&
		!readback.RawOutputLoaded
}

func productionAdapterObjectiveCloseoutWriterDurableReviewPacketEmpty(packet ProductionAdapterObjectiveCloseoutWriterDurableReviewPacket) bool {
	return !packet.Projected &&
		!packet.Available &&
		packet.Status == "" &&
		packet.Mode == "" &&
		packet.ReviewPacketRef == "" &&
		packet.DurableRequestRef == "" &&
		packet.DurableResultRef == "" &&
		packet.WriterRef == "" &&
		len(packet.MissingInputs) == 0 &&
		len(packet.BlockedReasons) == 0 &&
		len(packet.Boundaries) == 0 &&
		packet.NextHostAction == "" &&
		!packet.RawOutputLoaded
}

func unavailableProductionAdapterObjectiveCloseoutWriterDurableReviewPacket() ProductionAdapterObjectiveCloseoutWriterDurableReviewPacket {
	return ProductionAdapterObjectiveCloseoutWriterDurableReviewPacket{
		ContractVersion: ContractVersion,
		Projected:       true,
		Available:       false,
		Status:          "unavailable",
		Mode:            "production_adapter_objective_closeout_writer_durable_review_packet",
		DisplayState:    "unavailable",
		DisplayStage:    "unavailable",
		DisplaySections: productionAdapterObjectiveCloseoutWriterDurableReviewDisplaySections(),
		Boundaries: []Boundary{
			"production_adapter_objective_closeout_writer_durable_review_packet",
			"objective_closeout_writer_durable_review_packet_projection_only",
			"host_owned_objective_closeout_writer_adapter",
			"host_cli_objective_closeout_writer_durable_display",
			"no_runner_dispatch",
			"no_dry_run_by_core",
			"no_durable_write_by_core",
			"no_objective_store_write_by_core",
			"no_runstore_write_by_core",
		},
		RunnerEffect:   "none",
		PromptEffect:   "none",
		NextHostAction: "provide_objective_closeout_writer_durable_projection",
	}
}

func unavailableProductionAdapterObjectiveCloseoutWriterDurableReviewBlackboxFixture() ProductionAdapterObjectiveCloseoutWriterDurableReviewBlackboxFixture {
	return ProductionAdapterObjectiveCloseoutWriterDurableReviewBlackboxFixture{
		ContractVersion: ContractVersion,
		Projected:       true,
		Available:       false,
		Status:          "unavailable",
		Mode:            "production_adapter_objective_closeout_writer_durable_review_blackbox_fixture",
		DisplayState:    "unavailable",
		DisplayStage:    "unavailable",
		DisplaySections: productionAdapterObjectiveCloseoutWriterDurableReviewDisplaySections(),
		Boundaries: []Boundary{
			"production_adapter_objective_closeout_writer_durable_review_blackbox_fixture",
			"objective_closeout_writer_durable_review_fixture_projection_only",
			"host_owned_objective_closeout_writer_adapter",
			"host_cli_objective_closeout_writer_durable_display",
			"no_runner_dispatch",
			"no_dry_run_by_core",
			"no_durable_write_by_core",
			"no_objective_store_write_by_core",
			"no_runstore_write_by_core",
		},
		RunnerEffect:   "none",
		PromptEffect:   "none",
		NextHostAction: "provide_objective_closeout_writer_durable_review_packet",
	}
}
