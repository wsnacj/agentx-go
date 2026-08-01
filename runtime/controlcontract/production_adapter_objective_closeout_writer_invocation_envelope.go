package controlcontract

type ProductionAdapterObjectiveCloseoutWriterInvocationEnvelopeInput struct {
	InvocationEnvelopeRef           DisplaySafeRef                                              `json:"invocation_envelope_ref,omitempty"`
	DurableReviewPacket             ProductionAdapterObjectiveCloseoutWriterDurableReviewPacket `json:"durable_review_packet,omitempty"`
	WriterInvocationRef             DisplaySafeRef                                              `json:"writer_invocation_ref,omitempty"`
	HostAdapterVersionRef           DisplaySafeRef                                              `json:"host_adapter_version_ref,omitempty"`
	CapabilityProofRefs             []DisplaySafeRef                                            `json:"capability_proof_refs,omitempty"`
	ApprovalBindingRefs             []DisplaySafeRef                                            `json:"approval_binding_refs,omitempty"`
	IdempotencyKeyRef               DisplaySafeRef                                              `json:"idempotency_key_ref,omitempty"`
	DryRunProofRef                  DisplaySafeRef                                              `json:"dry_run_proof_ref,omitempty"`
	HostDurableWriteConfirmationRef DisplaySafeRef                                              `json:"host_durable_write_confirmation_ref,omitempty"`
	ExpectedHostAdapterRunRef       DisplaySafeRef                                              `json:"expected_host_adapter_run_ref,omitempty"`
	ExpectedDurableResultRef        DisplaySafeRef                                              `json:"expected_durable_result_ref,omitempty"`
	ExpectedReadbackRef             DisplaySafeRef                                              `json:"expected_readback_ref,omitempty"`
	ExpectedFailureRef              DisplaySafeRef                                              `json:"expected_failure_ref,omitempty"`
	ExpectedCompensationRef         DisplaySafeRef                                              `json:"expected_compensation_ref,omitempty"`
	TimeoutPolicyRef                DisplaySafeRef                                              `json:"timeout_policy_ref,omitempty"`
	RawOutputLoaded                 bool                                                        `json:"raw_output_loaded"`
}

type ProductionAdapterObjectiveCloseoutWriterInvocationEnvelope struct {
	ContractVersion                 string           `json:"contract_version,omitempty"`
	Projected                       bool             `json:"projected"`
	Available                       bool             `json:"available"`
	Status                          string           `json:"status,omitempty"`
	Mode                            string           `json:"mode,omitempty"`
	ReadyForHostDisplay             bool             `json:"ready_for_host_display"`
	ReadyForHostAdapterInvocation   bool             `json:"ready_for_host_adapter_invocation"`
	HostAdapterInvocationAuthorized bool             `json:"host_adapter_invocation_authorized"`
	HostMayInvokeWriterAdapter      bool             `json:"host_may_invoke_writer_adapter"`
	HostMayExecuteDurableWrite      bool             `json:"host_may_execute_durable_write"`
	CoreInvocationExecuted          bool             `json:"core_invocation_executed"`
	DryRunByCore                    bool             `json:"dry_run_by_core"`
	DurableWriteByCore              bool             `json:"durable_write_by_core"`
	ObjectiveStoreWriteByCore       bool             `json:"objective_store_write_by_core"`
	RunstoreWriteByCore             bool             `json:"runstore_write_by_core"`
	InvocationEnvelopeRef           DisplaySafeRef   `json:"invocation_envelope_ref,omitempty"`
	ReviewPacketRef                 DisplaySafeRef   `json:"review_packet_ref,omitempty"`
	DurableRequestRef               DisplaySafeRef   `json:"durable_request_ref,omitempty"`
	WriterInvocationRef             DisplaySafeRef   `json:"writer_invocation_ref,omitempty"`
	WriterOptInRef                  DisplaySafeRef   `json:"writer_opt_in_ref,omitempty"`
	WriterRef                       DisplaySafeRef   `json:"writer_ref,omitempty"`
	OwnerRef                        DisplaySafeRef   `json:"owner_ref,omitempty"`
	HostWriterBindingRef            DisplaySafeRef   `json:"host_writer_binding_ref,omitempty"`
	HostAdapterVersionRef           DisplaySafeRef   `json:"host_adapter_version_ref,omitempty"`
	ExpectedHostAdapterRunRef       DisplaySafeRef   `json:"expected_host_adapter_run_ref,omitempty"`
	DryRunSmokeRef                  DisplaySafeRef   `json:"dry_run_smoke_ref,omitempty"`
	DryRunResultRef                 DisplaySafeRef   `json:"dry_run_result_ref,omitempty"`
	DryRunProofRef                  DisplaySafeRef   `json:"dry_run_proof_ref,omitempty"`
	ExpectedDurableResultRef        DisplaySafeRef   `json:"expected_durable_result_ref,omitempty"`
	ExpectedReadbackRef             DisplaySafeRef   `json:"expected_readback_ref,omitempty"`
	ExpectedFailureRef              DisplaySafeRef   `json:"expected_failure_ref,omitempty"`
	ExpectedCompensationRef         DisplaySafeRef   `json:"expected_compensation_ref,omitempty"`
	ObjectiveCloseoutHandoffRef     DisplaySafeRef   `json:"objective_closeout_handoff_ref,omitempty"`
	HostUIHandoffRef                DisplaySafeRef   `json:"host_ui_handoff_ref,omitempty"`
	ObjectiveCloseoutPacketRef      DisplaySafeRef   `json:"objective_closeout_packet_ref,omitempty"`
	ObjectiveRef                    DisplaySafeRef   `json:"objective_ref,omitempty"`
	HostObjectiveLifecycleRef       DisplaySafeRef   `json:"host_objective_lifecycle_ref,omitempty"`
	HostRunstoreRef                 DisplaySafeRef   `json:"host_runstore_ref,omitempty"`
	ExpectedDurableEventRef         DisplaySafeRef   `json:"expected_durable_event_ref,omitempty"`
	ExpectedObjectiveStateRef       DisplaySafeRef   `json:"expected_objective_state_ref,omitempty"`
	HostDurableApplyConfirmationRef DisplaySafeRef   `json:"host_durable_apply_confirmation_ref,omitempty"`
	HostDurableWriteConfirmationRef DisplaySafeRef   `json:"host_durable_write_confirmation_ref,omitempty"`
	CapabilityProofRefs             []DisplaySafeRef `json:"capability_proof_refs,omitempty"`
	ApprovalBindingRefs             []DisplaySafeRef `json:"approval_binding_refs,omitempty"`
	IdempotencyKeyRef               DisplaySafeRef   `json:"idempotency_key_ref,omitempty"`
	IdempotencyContractRef          DisplaySafeRef   `json:"idempotency_contract_ref,omitempty"`
	ReadbackContractRef             DisplaySafeRef   `json:"readback_contract_ref,omitempty"`
	RollbackReviewRef               DisplaySafeRef   `json:"rollback_review_ref,omitempty"`
	CompensationReviewRef           DisplaySafeRef   `json:"compensation_review_ref,omitempty"`
	TimeoutPolicyRef                DisplaySafeRef   `json:"timeout_policy_ref,omitempty"`
	RequiredInputs                  []MissingInput   `json:"required_inputs,omitempty"`
	MissingInputs                   []MissingInput   `json:"missing_inputs,omitempty"`
	BlockedReasons                  []string         `json:"blocked_reasons,omitempty"`
	FailureClass                    FailureClass     `json:"failure_class,omitempty"`
	Boundaries                      []Boundary       `json:"boundaries,omitempty"`
	NextHostAction                  NextHostAction   `json:"next_host_action,omitempty"`
	RunnerEffect                    string           `json:"runner_effect,omitempty"`
	PromptEffect                    string           `json:"prompt_effect,omitempty"`
	RawOutputLoaded                 bool             `json:"raw_output_loaded"`
}

type ProductionAdapterObjectiveCloseoutWriterInvocationResultEnvelopeInput struct {
	ResultEnvelopeRef  DisplaySafeRef                                             `json:"result_envelope_ref,omitempty"`
	InvocationEnvelope ProductionAdapterObjectiveCloseoutWriterInvocationEnvelope `json:"invocation_envelope,omitempty"`
	DurableResult      ProductionAdapterObjectiveCloseoutWriterDurableResult      `json:"durable_result,omitempty"`
	RawOutputLoaded    bool                                                       `json:"raw_output_loaded"`
}

type ProductionAdapterObjectiveCloseoutWriterInvocationResultEnvelope struct {
	ContractVersion                 string           `json:"contract_version,omitempty"`
	Projected                       bool             `json:"projected"`
	Available                       bool             `json:"available"`
	Status                          string           `json:"status,omitempty"`
	Mode                            string           `json:"mode,omitempty"`
	ReadyForHostDisplay             bool             `json:"ready_for_host_display"`
	ReadyForDurableReadbackReview   bool             `json:"ready_for_durable_readback_review"`
	ReadyForFailureReview           bool             `json:"ready_for_failure_review"`
	ReadyForCompensationReview      bool             `json:"ready_for_compensation_review"`
	HostAdapterInvocationBound      bool             `json:"host_adapter_invocation_bound"`
	HostAdapterInvocationAuthorized bool             `json:"host_adapter_invocation_authorized"`
	HostDurableWriteReported        bool             `json:"host_durable_write_reported"`
	HostDurableWriteSucceeded       bool             `json:"host_durable_write_succeeded"`
	HostDurableWriteFailed          bool             `json:"host_durable_write_failed"`
	HostDurableWriteRecorded        bool             `json:"host_durable_write_recorded"`
	CoreInvocationExecuted          bool             `json:"core_invocation_executed"`
	DryRunByCore                    bool             `json:"dry_run_by_core"`
	DurableWriteByCore              bool             `json:"durable_write_by_core"`
	ObjectiveStoreWriteByCore       bool             `json:"objective_store_write_by_core"`
	RunstoreWriteByCore             bool             `json:"runstore_write_by_core"`
	ResultEnvelopeRef               DisplaySafeRef   `json:"result_envelope_ref,omitempty"`
	InvocationEnvelopeRef           DisplaySafeRef   `json:"invocation_envelope_ref,omitempty"`
	ReviewPacketRef                 DisplaySafeRef   `json:"review_packet_ref,omitempty"`
	DurableRequestRef               DisplaySafeRef   `json:"durable_request_ref,omitempty"`
	DurableResultRef                DisplaySafeRef   `json:"durable_result_ref,omitempty"`
	ExpectedDurableResultRef        DisplaySafeRef   `json:"expected_durable_result_ref,omitempty"`
	WriterInvocationRef             DisplaySafeRef   `json:"writer_invocation_ref,omitempty"`
	WriterOptInRef                  DisplaySafeRef   `json:"writer_opt_in_ref,omitempty"`
	WriterRef                       DisplaySafeRef   `json:"writer_ref,omitempty"`
	HostWriterBindingRef            DisplaySafeRef   `json:"host_writer_binding_ref,omitempty"`
	HostAdapterVersionRef           DisplaySafeRef   `json:"host_adapter_version_ref,omitempty"`
	ExpectedHostAdapterRunRef       DisplaySafeRef   `json:"expected_host_adapter_run_ref,omitempty"`
	HostAdapterRunRef               DisplaySafeRef   `json:"host_adapter_run_ref,omitempty"`
	DryRunSmokeRef                  DisplaySafeRef   `json:"dry_run_smoke_ref,omitempty"`
	DryRunResultRef                 DisplaySafeRef   `json:"dry_run_result_ref,omitempty"`
	ExpectedReadbackRef             DisplaySafeRef   `json:"expected_readback_ref,omitempty"`
	ExpectedFailureRef              DisplaySafeRef   `json:"expected_failure_ref,omitempty"`
	ExpectedCompensationRef         DisplaySafeRef   `json:"expected_compensation_ref,omitempty"`
	ObjectiveCloseoutHandoffRef     DisplaySafeRef   `json:"objective_closeout_handoff_ref,omitempty"`
	ObjectiveCloseoutPacketRef      DisplaySafeRef   `json:"objective_closeout_packet_ref,omitempty"`
	ObjectiveRef                    DisplaySafeRef   `json:"objective_ref,omitempty"`
	HostRunstoreRef                 DisplaySafeRef   `json:"host_runstore_ref,omitempty"`
	ExpectedDurableEventRef         DisplaySafeRef   `json:"expected_durable_event_ref,omitempty"`
	ExpectedObjectiveStateRef       DisplaySafeRef   `json:"expected_objective_state_ref,omitempty"`
	AppliedDurableEventRef          DisplaySafeRef   `json:"applied_durable_event_ref,omitempty"`
	AppliedRunstoreRef              DisplaySafeRef   `json:"applied_runstore_ref,omitempty"`
	AppliedObjectiveStateRef        DisplaySafeRef   `json:"applied_objective_state_ref,omitempty"`
	FailureRef                      DisplaySafeRef   `json:"failure_ref,omitempty"`
	CompensationRef                 DisplaySafeRef   `json:"compensation_ref,omitempty"`
	HostDurableWriteConfirmationRef DisplaySafeRef   `json:"host_durable_write_confirmation_ref,omitempty"`
	CapabilityProofRefs             []DisplaySafeRef `json:"capability_proof_refs,omitempty"`
	ApprovalBindingRefs             []DisplaySafeRef `json:"approval_binding_refs,omitempty"`
	IdempotencyKeyRef               DisplaySafeRef   `json:"idempotency_key_ref,omitempty"`
	IdempotencyContractRef          DisplaySafeRef   `json:"idempotency_contract_ref,omitempty"`
	ReadbackContractRef             DisplaySafeRef   `json:"readback_contract_ref,omitempty"`
	RollbackReviewRef               DisplaySafeRef   `json:"rollback_review_ref,omitempty"`
	CompensationReviewRef           DisplaySafeRef   `json:"compensation_review_ref,omitempty"`
	DurableEvidenceRefs             []DisplaySafeRef `json:"durable_evidence_refs,omitempty"`
	RequiredInputs                  []MissingInput   `json:"required_inputs,omitempty"`
	MissingInputs                   []MissingInput   `json:"missing_inputs,omitempty"`
	BlockedReasons                  []string         `json:"blocked_reasons,omitempty"`
	FailureClass                    FailureClass     `json:"failure_class,omitempty"`
	Boundaries                      []Boundary       `json:"boundaries,omitempty"`
	NextHostAction                  NextHostAction   `json:"next_host_action,omitempty"`
	RunnerEffect                    string           `json:"runner_effect,omitempty"`
	PromptEffect                    string           `json:"prompt_effect,omitempty"`
	RawOutputLoaded                 bool             `json:"raw_output_loaded"`
}

func BuildProductionAdapterObjectiveCloseoutWriterInvocationEnvelope(input ProductionAdapterObjectiveCloseoutWriterInvocationEnvelopeInput) ProductionAdapterObjectiveCloseoutWriterInvocationEnvelope {
	if productionAdapterObjectiveCloseoutWriterDurableReviewPacketEmpty(input.DurableReviewPacket) {
		return unavailableProductionAdapterObjectiveCloseoutWriterInvocationEnvelope()
	}
	rawReview := input.DurableReviewPacket
	review := rawReview.Normalize()
	result := ProductionAdapterObjectiveCloseoutWriterInvocationEnvelope{
		ContractVersion:                 ContractVersion,
		Projected:                       true,
		Available:                       true,
		Status:                          "blocked",
		Mode:                            "production_adapter_objective_closeout_writer_invocation_envelope",
		InvocationEnvelopeRef:           normalizeOneDisplaySafeRef(input.InvocationEnvelopeRef),
		ReviewPacketRef:                 review.ReviewPacketRef,
		DurableRequestRef:               review.DurableRequestRef,
		WriterInvocationRef:             normalizeOneDisplaySafeRef(input.WriterInvocationRef),
		WriterOptInRef:                  review.WriterOptInRef,
		WriterRef:                       review.WriterRef,
		OwnerRef:                        review.OwnerRef,
		HostWriterBindingRef:            review.HostWriterBindingRef,
		HostAdapterVersionRef:           normalizeOneDisplaySafeRef(input.HostAdapterVersionRef),
		ExpectedHostAdapterRunRef:       normalizeOneDisplaySafeRef(input.ExpectedHostAdapterRunRef),
		DryRunSmokeRef:                  review.DryRunSmokeRef,
		DryRunResultRef:                 review.DryRunResultRef,
		DryRunProofRef:                  normalizeOneDisplaySafeRef(input.DryRunProofRef),
		ExpectedDurableResultRef:        normalizeOneDisplaySafeRef(input.ExpectedDurableResultRef),
		ExpectedReadbackRef:             normalizeOneDisplaySafeRef(input.ExpectedReadbackRef),
		ExpectedFailureRef:              normalizeOneDisplaySafeRef(input.ExpectedFailureRef),
		ExpectedCompensationRef:         normalizeOneDisplaySafeRef(input.ExpectedCompensationRef),
		ObjectiveCloseoutHandoffRef:     review.ObjectiveCloseoutHandoffRef,
		HostUIHandoffRef:                review.HostUIHandoffRef,
		ObjectiveCloseoutPacketRef:      review.ObjectiveCloseoutPacketRef,
		ObjectiveRef:                    review.ObjectiveRef,
		HostObjectiveLifecycleRef:       review.HostObjectiveLifecycleRef,
		HostRunstoreRef:                 review.HostRunstoreRef,
		ExpectedDurableEventRef:         review.ExpectedDurableEventRef,
		ExpectedObjectiveStateRef:       review.ExpectedObjectiveStateRef,
		HostDurableWriteConfirmationRef: normalizeOneDisplaySafeRef(input.HostDurableWriteConfirmationRef),
		CapabilityProofRefs:             normalizeDisplaySafeRefs(input.CapabilityProofRefs),
		ApprovalBindingRefs:             normalizeDisplaySafeRefs(input.ApprovalBindingRefs),
		IdempotencyKeyRef:               normalizeOneDisplaySafeRef(input.IdempotencyKeyRef),
		IdempotencyContractRef:          review.IdempotencyContractRef,
		ReadbackContractRef:             review.ReadbackContractRef,
		RollbackReviewRef:               review.RollbackReviewRef,
		CompensationReviewRef:           review.CompensationReviewRef,
		TimeoutPolicyRef:                normalizeOneDisplaySafeRef(input.TimeoutPolicyRef),
		RequiredInputs:                  productionAdapterObjectiveCloseoutWriterInvocationEnvelopeRequiredInputs(),
		FailureClass:                    FailureNone,
		Boundaries:                      productionAdapterObjectiveCloseoutWriterInvocationEnvelopeBoundaries(review.Boundaries),
		RunnerEffect:                    "none",
		PromptEffect:                    "none",
		RawOutputLoaded:                 input.RawOutputLoaded || review.RawOutputLoaded,
	}
	if productionAdapterObjectiveCloseoutWriterInvocationEnvelopeUnsafe(input) || productionAdapterObjectiveCloseoutWriterDurableReviewPacketUnsafeOutput(rawReview) {
		result = productionAdapterObjectiveCloseoutWriterInvocationEnvelopeBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if !review.ReadyForHostDurableWrite || !review.HostMayExecuteDurableWrite || !review.HostDurableWriteAuthorized {
		result = productionAdapterObjectiveCloseoutWriterInvocationEnvelopeBlock(result, firstFailureClass(review.FailureClass, FailureAuthorizationMissing), "durable_writer_review_not_ready", "host:objective_closeout_writer_durable_review_packet", firstNextHostAction(review.NextHostAction, "review_objective_closeout_writer_durable_review"))
	}
	if result.InvocationEnvelopeRef != "" && review.ReadyForHostDisplay && !result.RawOutputLoaded {
		result.ReadyForHostDisplay = true
	}
	for _, check := range productionAdapterObjectiveCloseoutWriterInvocationEnvelopeRequiredRefChecks(result) {
		if check.ref == "" {
			result = productionAdapterObjectiveCloseoutWriterInvocationEnvelopeBlock(result, check.failure, check.reason, check.missing, check.next)
		}
	}
	if len(result.CapabilityProofRefs) == 0 {
		result = productionAdapterObjectiveCloseoutWriterInvocationEnvelopeBlock(result, FailureCapabilityMissing, "capability_proof_ref_missing", "host:objective_closeout_writer_capability_proof_ref", "provide_objective_closeout_writer_capability_proof")
	}
	if len(result.ApprovalBindingRefs) == 0 {
		result = productionAdapterObjectiveCloseoutWriterInvocationEnvelopeBlock(result, FailureApprovalRequired, "approval_binding_ref_missing", "host:objective_closeout_writer_approval_binding_ref", "request_objective_closeout_writer_approval")
	}
	for _, mismatch := range productionAdapterObjectiveCloseoutWriterInvocationEnvelopeMismatches(result, review) {
		result = productionAdapterObjectiveCloseoutWriterInvocationEnvelopeBlock(result, mismatch.failure, mismatch.reason, mismatch.missing, "review_objective_closeout_writer_invocation_envelope")
	}
	if len(result.BlockedReasons) == 0 && len(result.MissingInputs) == 0 {
		result.Status = "ready_for_objective_closeout_writer_host_adapter_invocation"
		result.ReadyForHostDisplay = true
		result.ReadyForHostAdapterInvocation = true
		result.HostAdapterInvocationAuthorized = true
		result.HostMayInvokeWriterAdapter = true
		result.HostMayExecuteDurableWrite = true
		result.NextHostAction = "host_may_invoke_objective_closeout_writer_adapter"
		result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_objective_closeout_writer_host_adapter_invocation", "host_may_invoke_writer_adapter")
	}
	return result.Normalize()
}

func BuildProductionAdapterObjectiveCloseoutWriterInvocationResultEnvelope(input ProductionAdapterObjectiveCloseoutWriterInvocationResultEnvelopeInput) ProductionAdapterObjectiveCloseoutWriterInvocationResultEnvelope {
	if productionAdapterObjectiveCloseoutWriterInvocationEnvelopeEmpty(input.InvocationEnvelope) {
		return unavailableProductionAdapterObjectiveCloseoutWriterInvocationResultEnvelope()
	}
	rawEnvelope := input.InvocationEnvelope
	envelope := rawEnvelope.Normalize()
	resultProvided := !productionAdapterObjectiveCloseoutWriterDurableResultEmpty(input.DurableResult)
	durableResult := ProductionAdapterObjectiveCloseoutWriterDurableResult{}
	if resultProvided {
		durableResult = input.DurableResult.Normalize()
	}
	result := ProductionAdapterObjectiveCloseoutWriterInvocationResultEnvelope{
		ContractVersion:                 ContractVersion,
		Projected:                       true,
		Available:                       true,
		Status:                          "blocked",
		Mode:                            "production_adapter_objective_closeout_writer_invocation_result_envelope",
		ResultEnvelopeRef:               normalizeOneDisplaySafeRef(input.ResultEnvelopeRef),
		InvocationEnvelopeRef:           envelope.InvocationEnvelopeRef,
		ReviewPacketRef:                 envelope.ReviewPacketRef,
		DurableRequestRef:               firstDisplaySafeRef(durableResult.DurableRequestRef, envelope.DurableRequestRef),
		DurableResultRef:                durableResult.DurableResultRef,
		ExpectedDurableResultRef:        envelope.ExpectedDurableResultRef,
		WriterInvocationRef:             envelope.WriterInvocationRef,
		WriterOptInRef:                  firstDisplaySafeRef(durableResult.WriterOptInRef, envelope.WriterOptInRef),
		WriterRef:                       firstDisplaySafeRef(durableResult.WriterRef, envelope.WriterRef),
		HostWriterBindingRef:            firstDisplaySafeRef(durableResult.HostWriterBindingRef, envelope.HostWriterBindingRef),
		HostAdapterVersionRef:           envelope.HostAdapterVersionRef,
		ExpectedHostAdapterRunRef:       envelope.ExpectedHostAdapterRunRef,
		HostAdapterRunRef:               durableResult.HostAdapterRunRef,
		DryRunSmokeRef:                  firstDisplaySafeRef(durableResult.DryRunSmokeRef, envelope.DryRunSmokeRef),
		DryRunResultRef:                 firstDisplaySafeRef(durableResult.DryRunResultRef, envelope.DryRunResultRef),
		ExpectedReadbackRef:             envelope.ExpectedReadbackRef,
		ExpectedFailureRef:              envelope.ExpectedFailureRef,
		ExpectedCompensationRef:         envelope.ExpectedCompensationRef,
		ObjectiveCloseoutHandoffRef:     firstDisplaySafeRef(durableResult.ObjectiveCloseoutHandoffRef, envelope.ObjectiveCloseoutHandoffRef),
		ObjectiveCloseoutPacketRef:      firstDisplaySafeRef(durableResult.ObjectiveCloseoutPacketRef, envelope.ObjectiveCloseoutPacketRef),
		ObjectiveRef:                    firstDisplaySafeRef(durableResult.ObjectiveRef, envelope.ObjectiveRef),
		HostRunstoreRef:                 firstDisplaySafeRef(durableResult.HostRunstoreRef, envelope.HostRunstoreRef),
		ExpectedDurableEventRef:         envelope.ExpectedDurableEventRef,
		ExpectedObjectiveStateRef:       envelope.ExpectedObjectiveStateRef,
		AppliedDurableEventRef:          durableResult.AppliedDurableEventRef,
		AppliedRunstoreRef:              durableResult.AppliedRunstoreRef,
		AppliedObjectiveStateRef:        durableResult.AppliedObjectiveStateRef,
		FailureRef:                      durableResult.FailureRef,
		CompensationRef:                 durableResult.CompensationRef,
		HostDurableWriteConfirmationRef: envelope.HostDurableWriteConfirmationRef,
		CapabilityProofRefs:             cloneDisplaySafeRefs(envelope.CapabilityProofRefs),
		ApprovalBindingRefs:             cloneDisplaySafeRefs(envelope.ApprovalBindingRefs),
		IdempotencyKeyRef:               firstDisplaySafeRef(durableResult.IdempotencyRef, envelope.IdempotencyKeyRef),
		IdempotencyContractRef:          firstDisplaySafeRef(durableResult.IdempotencyContractRef, envelope.IdempotencyContractRef),
		ReadbackContractRef:             firstDisplaySafeRef(durableResult.ReadbackContractRef, envelope.ReadbackContractRef),
		RollbackReviewRef:               firstDisplaySafeRef(durableResult.RollbackReviewRef, envelope.RollbackReviewRef),
		CompensationReviewRef:           firstDisplaySafeRef(durableResult.CompensationReviewRef, envelope.CompensationReviewRef),
		DurableEvidenceRefs:             cloneDisplaySafeRefs(durableResult.DurableEvidenceRefs),
		RequiredInputs:                  productionAdapterObjectiveCloseoutWriterInvocationResultEnvelopeRequiredInputs(),
		HostAdapterInvocationAuthorized: envelope.HostAdapterInvocationAuthorized,
		HostDurableWriteReported:        durableResult.HostDurableWriteReported,
		HostDurableWriteSucceeded:       durableResult.HostDurableWriteSucceeded,
		HostDurableWriteFailed:          durableResult.HostDurableWriteFailed,
		HostDurableWriteRecorded:        durableResult.HostDurableWriteRecorded,
		FailureClass:                    FailureNone,
		Boundaries:                      productionAdapterObjectiveCloseoutWriterInvocationResultEnvelopeBoundaries(envelope.Boundaries, durableResult.Boundaries),
		RunnerEffect:                    "none",
		PromptEffect:                    "none",
		RawOutputLoaded:                 input.RawOutputLoaded || envelope.RawOutputLoaded || durableResult.RawOutputLoaded,
	}
	if productionAdapterObjectiveCloseoutWriterInvocationResultEnvelopeUnsafe(input, durableResult, resultProvided) || productionAdapterObjectiveCloseoutWriterInvocationEnvelopeUnsafeOutput(rawEnvelope) {
		result = productionAdapterObjectiveCloseoutWriterInvocationResultEnvelopeBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if !envelope.ReadyForHostAdapterInvocation || !envelope.HostAdapterInvocationAuthorized {
		result = productionAdapterObjectiveCloseoutWriterInvocationResultEnvelopeBlock(result, firstFailureClass(envelope.FailureClass, FailureAuthorizationMissing), "writer_invocation_envelope_not_ready", "host:objective_closeout_writer_invocation_envelope", firstNextHostAction(envelope.NextHostAction, "review_objective_closeout_writer_invocation_envelope"))
	}
	if result.ResultEnvelopeRef != "" && envelope.ReadyForHostDisplay && !result.RawOutputLoaded {
		result.ReadyForHostDisplay = true
	}
	if result.ResultEnvelopeRef == "" {
		result = productionAdapterObjectiveCloseoutWriterInvocationResultEnvelopeBlock(result, FailureEvidenceMissing, "writer_invocation_result_envelope_ref_missing", "host:objective_closeout_writer_invocation_result_envelope_ref", "provide_objective_closeout_writer_invocation_result_envelope")
	}
	if !resultProvided {
		result = productionAdapterObjectiveCloseoutWriterInvocationResultEnvelopeBlock(result, FailureEvidenceMissing, "writer_durable_result_missing", "host:objective_closeout_writer_durable_result", "provide_objective_closeout_writer_durable_result")
		return result.Normalize()
	}
	if !durableResult.HostDurableWriteReported || !durableResult.HostDurableWriteRecorded {
		result = productionAdapterObjectiveCloseoutWriterInvocationResultEnvelopeBlock(result, firstFailureClass(durableResult.FailureClass, FailureEvidenceMissing), "writer_durable_result_not_recorded", "host:objective_closeout_writer_durable_result", firstNextHostAction(durableResult.NextHostAction, "provide_objective_closeout_writer_durable_result"))
		return result.Normalize()
	}
	if durableResult.HostDurableWriteSucceeded && durableResult.HostDurableWriteFailed {
		result = productionAdapterObjectiveCloseoutWriterInvocationResultEnvelopeBlock(result, FailureVerificationFailed, "writer_durable_result_conflict", "host:objective_closeout_writer_durable_result", "review_objective_closeout_writer_durable_result")
		return result.Normalize()
	}
	for _, mismatch := range productionAdapterObjectiveCloseoutWriterInvocationResultEnvelopeCommonMismatches(result, envelope, durableResult) {
		result = productionAdapterObjectiveCloseoutWriterInvocationResultEnvelopeBlock(result, mismatch.failure, mismatch.reason, mismatch.missing, "review_objective_closeout_writer_invocation_result")
	}
	switch {
	case durableResult.HostDurableWriteSucceeded:
		for _, mismatch := range productionAdapterObjectiveCloseoutWriterInvocationResultEnvelopeSuccessMismatches(result) {
			result = productionAdapterObjectiveCloseoutWriterInvocationResultEnvelopeBlock(result, mismatch.failure, mismatch.reason, mismatch.missing, "review_objective_closeout_writer_invocation_result")
		}
		if len(result.BlockedReasons) == 0 && len(result.MissingInputs) == 0 {
			result.Status = "ready_for_objective_closeout_writer_invocation_result_readback_review"
			result.ReadyForHostDisplay = true
			result.ReadyForDurableReadbackReview = true
			result.HostAdapterInvocationBound = true
			result.NextHostAction = "bind_objective_closeout_writer_durable_readback"
			result.Boundaries = AppendBoundaries(result.Boundaries, "objective_closeout_writer_invocation_result_bound", "ready_for_objective_closeout_writer_invocation_result_readback_review")
		}
	case durableResult.HostDurableWriteFailed:
		for _, mismatch := range productionAdapterObjectiveCloseoutWriterInvocationResultEnvelopeFailureMismatches(result) {
			result = productionAdapterObjectiveCloseoutWriterInvocationResultEnvelopeBlock(result, mismatch.failure, mismatch.reason, mismatch.missing, "review_objective_closeout_writer_invocation_result")
		}
		if len(result.BlockedReasons) == 0 && len(result.MissingInputs) == 0 {
			result.Status = "ready_for_objective_closeout_writer_invocation_result_failure_review"
			result.ReadyForHostDisplay = true
			result.ReadyForFailureReview = true
			result.ReadyForCompensationReview = result.CompensationRef != ""
			result.HostAdapterInvocationBound = true
			result.FailureClass = FailureVerificationFailed
			result.NextHostAction = "review_objective_closeout_writer_durable_failure"
			result.Boundaries = AppendBoundaries(result.Boundaries, "objective_closeout_writer_invocation_failure_bound", "ready_for_objective_closeout_writer_invocation_result_failure_review", "compensation_not_executed")
		}
	default:
		result = productionAdapterObjectiveCloseoutWriterInvocationResultEnvelopeBlock(result, FailureEvidenceMissing, "writer_durable_result_status_missing", "host:objective_closeout_writer_durable_result", "provide_objective_closeout_writer_durable_result")
	}
	return result.Normalize()
}

func CloneProductionAdapterObjectiveCloseoutWriterInvocationEnvelope(in ProductionAdapterObjectiveCloseoutWriterInvocationEnvelope) ProductionAdapterObjectiveCloseoutWriterInvocationEnvelope {
	out := in
	out.CapabilityProofRefs = cloneDisplaySafeRefs(in.CapabilityProofRefs)
	out.ApprovalBindingRefs = cloneDisplaySafeRefs(in.ApprovalBindingRefs)
	out.RequiredInputs = cloneMissingInputs(in.RequiredInputs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (p ProductionAdapterObjectiveCloseoutWriterInvocationEnvelope) Clone() ProductionAdapterObjectiveCloseoutWriterInvocationEnvelope {
	return CloneProductionAdapterObjectiveCloseoutWriterInvocationEnvelope(p)
}

func (p ProductionAdapterObjectiveCloseoutWriterInvocationEnvelope) Normalize() ProductionAdapterObjectiveCloseoutWriterInvocationEnvelope {
	out := CloneProductionAdapterObjectiveCloseoutWriterInvocationEnvelope(p)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = normalizeControlToken(out.Status)
	out.Mode = normalizeControlToken(out.Mode)
	if out.Mode == "" {
		out.Mode = "production_adapter_objective_closeout_writer_invocation_envelope"
	}
	out.InvocationEnvelopeRef = normalizeOneDisplaySafeRef(out.InvocationEnvelopeRef)
	out.ReviewPacketRef = normalizeOneDisplaySafeRef(out.ReviewPacketRef)
	out.DurableRequestRef = normalizeOneDisplaySafeRef(out.DurableRequestRef)
	out.WriterInvocationRef = normalizeOneDisplaySafeRef(out.WriterInvocationRef)
	out.WriterOptInRef = normalizeOneDisplaySafeRef(out.WriterOptInRef)
	out.WriterRef = normalizeOneDisplaySafeRef(out.WriterRef)
	out.OwnerRef = normalizeOneDisplaySafeRef(out.OwnerRef)
	out.HostWriterBindingRef = normalizeOneDisplaySafeRef(out.HostWriterBindingRef)
	out.HostAdapterVersionRef = normalizeOneDisplaySafeRef(out.HostAdapterVersionRef)
	out.ExpectedHostAdapterRunRef = normalizeOneDisplaySafeRef(out.ExpectedHostAdapterRunRef)
	out.DryRunSmokeRef = normalizeOneDisplaySafeRef(out.DryRunSmokeRef)
	out.DryRunResultRef = normalizeOneDisplaySafeRef(out.DryRunResultRef)
	out.DryRunProofRef = normalizeOneDisplaySafeRef(out.DryRunProofRef)
	out.ExpectedDurableResultRef = normalizeOneDisplaySafeRef(out.ExpectedDurableResultRef)
	out.ExpectedReadbackRef = normalizeOneDisplaySafeRef(out.ExpectedReadbackRef)
	out.ExpectedFailureRef = normalizeOneDisplaySafeRef(out.ExpectedFailureRef)
	out.ExpectedCompensationRef = normalizeOneDisplaySafeRef(out.ExpectedCompensationRef)
	out.ObjectiveCloseoutHandoffRef = normalizeOneDisplaySafeRef(out.ObjectiveCloseoutHandoffRef)
	out.HostUIHandoffRef = normalizeOneDisplaySafeRef(out.HostUIHandoffRef)
	out.ObjectiveCloseoutPacketRef = normalizeOneDisplaySafeRef(out.ObjectiveCloseoutPacketRef)
	out.ObjectiveRef = normalizeOneDisplaySafeRef(out.ObjectiveRef)
	out.HostObjectiveLifecycleRef = normalizeOneDisplaySafeRef(out.HostObjectiveLifecycleRef)
	out.HostRunstoreRef = normalizeOneDisplaySafeRef(out.HostRunstoreRef)
	out.ExpectedDurableEventRef = normalizeOneDisplaySafeRef(out.ExpectedDurableEventRef)
	out.ExpectedObjectiveStateRef = normalizeOneDisplaySafeRef(out.ExpectedObjectiveStateRef)
	out.HostDurableApplyConfirmationRef = normalizeOneDisplaySafeRef(out.HostDurableApplyConfirmationRef)
	out.HostDurableWriteConfirmationRef = normalizeOneDisplaySafeRef(out.HostDurableWriteConfirmationRef)
	out.CapabilityProofRefs = normalizeDisplaySafeRefs(out.CapabilityProofRefs)
	out.ApprovalBindingRefs = normalizeDisplaySafeRefs(out.ApprovalBindingRefs)
	out.IdempotencyKeyRef = normalizeOneDisplaySafeRef(out.IdempotencyKeyRef)
	out.IdempotencyContractRef = normalizeOneDisplaySafeRef(out.IdempotencyContractRef)
	out.ReadbackContractRef = normalizeOneDisplaySafeRef(out.ReadbackContractRef)
	out.RollbackReviewRef = normalizeOneDisplaySafeRef(out.RollbackReviewRef)
	out.CompensationReviewRef = normalizeOneDisplaySafeRef(out.CompensationReviewRef)
	out.TimeoutPolicyRef = normalizeOneDisplaySafeRef(out.TimeoutPolicyRef)
	out.RequiredInputs = normalizeMissingInputs(out.RequiredInputs)
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
	out.CoreInvocationExecuted = false
	out.DryRunByCore = false
	out.DurableWriteByCore = false
	out.ObjectiveStoreWriteByCore = false
	out.RunstoreWriteByCore = false
	if !out.Available {
		out.Status = "unavailable"
		out.ReadyForHostDisplay = false
		out.ReadyForHostAdapterInvocation = false
		out.HostAdapterInvocationAuthorized = false
		out.HostMayInvokeWriterAdapter = false
		out.HostMayExecuteDurableWrite = false
	}
	if out.Status == "" {
		out.Status = "blocked"
	}
	if out.RawOutputLoaded || productionAdapterObjectiveCloseoutWriterInvocationEnvelopeUnsafeOutput(out) {
		out.RawOutputLoaded = true
		out.Status = "blocked"
		out.ReadyForHostDisplay = false
		out.ReadyForHostAdapterInvocation = false
		out.HostAdapterInvocationAuthorized = false
		out.HostMayInvokeWriterAdapter = false
		out.HostMayExecuteDurableWrite = false
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
		out.InvocationEnvelopeRef != "" &&
		out.ReviewPacketRef != "" &&
		out.WriterRef != "" &&
		!out.RawOutputLoaded
	ready := out.Status == "ready_for_objective_closeout_writer_host_adapter_invocation" &&
		out.ReadyForHostDisplay &&
		out.InvocationEnvelopeRef != "" &&
		out.DurableRequestRef != "" &&
		out.WriterInvocationRef != "" &&
		out.WriterRef != "" &&
		out.HostWriterBindingRef != "" &&
		out.HostAdapterVersionRef != "" &&
		out.ExpectedHostAdapterRunRef != "" &&
		out.DryRunProofRef != "" &&
		out.ExpectedDurableResultRef != "" &&
		out.ExpectedReadbackRef != "" &&
		out.ExpectedFailureRef != "" &&
		out.ExpectedCompensationRef != "" &&
		out.HostDurableWriteConfirmationRef != "" &&
		out.IdempotencyKeyRef != "" &&
		out.TimeoutPolicyRef != "" &&
		len(out.CapabilityProofRefs) > 0 &&
		len(out.ApprovalBindingRefs) > 0 &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.RawOutputLoaded &&
		!out.CoreInvocationExecuted &&
		!out.DryRunByCore &&
		!out.DurableWriteByCore &&
		!out.ObjectiveStoreWriteByCore &&
		!out.RunstoreWriteByCore
	out.ReadyForHostAdapterInvocation = out.ReadyForHostAdapterInvocation && ready
	out.HostAdapterInvocationAuthorized = out.HostAdapterInvocationAuthorized && ready
	out.HostMayInvokeWriterAdapter = out.HostMayInvokeWriterAdapter && ready
	out.HostMayExecuteDurableWrite = out.HostMayExecuteDurableWrite && ready
	return out
}

func CloneProductionAdapterObjectiveCloseoutWriterInvocationResultEnvelope(in ProductionAdapterObjectiveCloseoutWriterInvocationResultEnvelope) ProductionAdapterObjectiveCloseoutWriterInvocationResultEnvelope {
	out := in
	out.CapabilityProofRefs = cloneDisplaySafeRefs(in.CapabilityProofRefs)
	out.ApprovalBindingRefs = cloneDisplaySafeRefs(in.ApprovalBindingRefs)
	out.DurableEvidenceRefs = cloneDisplaySafeRefs(in.DurableEvidenceRefs)
	out.RequiredInputs = cloneMissingInputs(in.RequiredInputs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (p ProductionAdapterObjectiveCloseoutWriterInvocationResultEnvelope) Clone() ProductionAdapterObjectiveCloseoutWriterInvocationResultEnvelope {
	return CloneProductionAdapterObjectiveCloseoutWriterInvocationResultEnvelope(p)
}

func (p ProductionAdapterObjectiveCloseoutWriterInvocationResultEnvelope) Normalize() ProductionAdapterObjectiveCloseoutWriterInvocationResultEnvelope {
	out := CloneProductionAdapterObjectiveCloseoutWriterInvocationResultEnvelope(p)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = normalizeControlToken(out.Status)
	out.Mode = normalizeControlToken(out.Mode)
	if out.Mode == "" {
		out.Mode = "production_adapter_objective_closeout_writer_invocation_result_envelope"
	}
	out.ResultEnvelopeRef = normalizeOneDisplaySafeRef(out.ResultEnvelopeRef)
	out.InvocationEnvelopeRef = normalizeOneDisplaySafeRef(out.InvocationEnvelopeRef)
	out.ReviewPacketRef = normalizeOneDisplaySafeRef(out.ReviewPacketRef)
	out.DurableRequestRef = normalizeOneDisplaySafeRef(out.DurableRequestRef)
	out.DurableResultRef = normalizeOneDisplaySafeRef(out.DurableResultRef)
	out.ExpectedDurableResultRef = normalizeOneDisplaySafeRef(out.ExpectedDurableResultRef)
	out.WriterInvocationRef = normalizeOneDisplaySafeRef(out.WriterInvocationRef)
	out.WriterOptInRef = normalizeOneDisplaySafeRef(out.WriterOptInRef)
	out.WriterRef = normalizeOneDisplaySafeRef(out.WriterRef)
	out.HostWriterBindingRef = normalizeOneDisplaySafeRef(out.HostWriterBindingRef)
	out.HostAdapterVersionRef = normalizeOneDisplaySafeRef(out.HostAdapterVersionRef)
	out.ExpectedHostAdapterRunRef = normalizeOneDisplaySafeRef(out.ExpectedHostAdapterRunRef)
	out.HostAdapterRunRef = normalizeOneDisplaySafeRef(out.HostAdapterRunRef)
	out.DryRunSmokeRef = normalizeOneDisplaySafeRef(out.DryRunSmokeRef)
	out.DryRunResultRef = normalizeOneDisplaySafeRef(out.DryRunResultRef)
	out.ExpectedReadbackRef = normalizeOneDisplaySafeRef(out.ExpectedReadbackRef)
	out.ExpectedFailureRef = normalizeOneDisplaySafeRef(out.ExpectedFailureRef)
	out.ExpectedCompensationRef = normalizeOneDisplaySafeRef(out.ExpectedCompensationRef)
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
	out.HostDurableWriteConfirmationRef = normalizeOneDisplaySafeRef(out.HostDurableWriteConfirmationRef)
	out.CapabilityProofRefs = normalizeDisplaySafeRefs(out.CapabilityProofRefs)
	out.ApprovalBindingRefs = normalizeDisplaySafeRefs(out.ApprovalBindingRefs)
	out.IdempotencyKeyRef = normalizeOneDisplaySafeRef(out.IdempotencyKeyRef)
	out.IdempotencyContractRef = normalizeOneDisplaySafeRef(out.IdempotencyContractRef)
	out.ReadbackContractRef = normalizeOneDisplaySafeRef(out.ReadbackContractRef)
	out.RollbackReviewRef = normalizeOneDisplaySafeRef(out.RollbackReviewRef)
	out.CompensationReviewRef = normalizeOneDisplaySafeRef(out.CompensationReviewRef)
	out.DurableEvidenceRefs = normalizeDisplaySafeRefs(out.DurableEvidenceRefs)
	out.RequiredInputs = normalizeMissingInputs(out.RequiredInputs)
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
	out.CoreInvocationExecuted = false
	out.DryRunByCore = false
	out.DurableWriteByCore = false
	out.ObjectiveStoreWriteByCore = false
	out.RunstoreWriteByCore = false
	if !out.Available {
		out.Status = "unavailable"
		out.ReadyForHostDisplay = false
		out.ReadyForDurableReadbackReview = false
		out.ReadyForFailureReview = false
		out.ReadyForCompensationReview = false
		out.HostAdapterInvocationBound = false
	}
	if out.Status == "" {
		out.Status = "blocked"
	}
	if out.RawOutputLoaded || productionAdapterObjectiveCloseoutWriterInvocationResultEnvelopeUnsafeOutput(out) {
		out.RawOutputLoaded = true
		out.Status = "blocked"
		out.ReadyForHostDisplay = false
		out.ReadyForDurableReadbackReview = false
		out.ReadyForFailureReview = false
		out.ReadyForCompensationReview = false
		out.HostAdapterInvocationBound = false
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
		out.ResultEnvelopeRef != "" &&
		out.InvocationEnvelopeRef != "" &&
		out.WriterRef != "" &&
		!out.RawOutputLoaded
	successReady := out.Status == "ready_for_objective_closeout_writer_invocation_result_readback_review" &&
		out.ReadyForHostDisplay &&
		out.HostAdapterInvocationAuthorized &&
		out.HostDurableWriteReported &&
		out.HostDurableWriteSucceeded &&
		out.HostDurableWriteRecorded &&
		!out.HostDurableWriteFailed &&
		out.ExpectedHostAdapterRunRef != "" &&
		out.HostAdapterRunRef == out.ExpectedHostAdapterRunRef &&
		out.DurableResultRef == out.ExpectedDurableResultRef &&
		out.AppliedDurableEventRef == out.ExpectedDurableEventRef &&
		out.AppliedRunstoreRef == out.HostRunstoreRef &&
		out.AppliedObjectiveStateRef == out.ExpectedObjectiveStateRef &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.CoreInvocationExecuted &&
		!out.DryRunByCore &&
		!out.DurableWriteByCore &&
		!out.ObjectiveStoreWriteByCore &&
		!out.RunstoreWriteByCore
	failureReady := out.Status == "ready_for_objective_closeout_writer_invocation_result_failure_review" &&
		out.ReadyForHostDisplay &&
		out.HostAdapterInvocationAuthorized &&
		out.HostDurableWriteReported &&
		out.HostDurableWriteFailed &&
		out.HostDurableWriteRecorded &&
		!out.HostDurableWriteSucceeded &&
		out.ExpectedHostAdapterRunRef != "" &&
		out.HostAdapterRunRef == out.ExpectedHostAdapterRunRef &&
		out.DurableResultRef == out.ExpectedDurableResultRef &&
		out.FailureRef == out.ExpectedFailureRef &&
		out.CompensationRef == out.ExpectedCompensationRef &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.CoreInvocationExecuted &&
		!out.DryRunByCore &&
		!out.DurableWriteByCore &&
		!out.ObjectiveStoreWriteByCore &&
		!out.RunstoreWriteByCore
	out.ReadyForDurableReadbackReview = out.ReadyForDurableReadbackReview && successReady
	out.ReadyForFailureReview = out.ReadyForFailureReview && failureReady
	out.ReadyForCompensationReview = out.ReadyForCompensationReview && failureReady && out.CompensationRef != ""
	out.HostAdapterInvocationBound = out.HostAdapterInvocationBound && (successReady || failureReady)
	return out
}

func productionAdapterObjectiveCloseoutWriterInvocationEnvelopeBlock(result ProductionAdapterObjectiveCloseoutWriterInvocationEnvelope, failure FailureClass, reason string, missing MissingInput, next NextHostAction) ProductionAdapterObjectiveCloseoutWriterInvocationEnvelope {
	result.Status = "blocked"
	result.ReadyForHostAdapterInvocation = false
	result.HostAdapterInvocationAuthorized = false
	result.HostMayInvokeWriterAdapter = false
	result.HostMayExecuteDurableWrite = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	return result
}

func productionAdapterObjectiveCloseoutWriterInvocationResultEnvelopeBlock(result ProductionAdapterObjectiveCloseoutWriterInvocationResultEnvelope, failure FailureClass, reason string, missing MissingInput, next NextHostAction) ProductionAdapterObjectiveCloseoutWriterInvocationResultEnvelope {
	result.Status = "blocked"
	result.ReadyForDurableReadbackReview = false
	result.ReadyForFailureReview = false
	result.ReadyForCompensationReview = false
	result.HostAdapterInvocationBound = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	return result
}

type productionAdapterObjectiveCloseoutWriterEnvelopeRefCheck struct {
	ref     DisplaySafeRef
	reason  string
	missing MissingInput
	next    NextHostAction
	failure FailureClass
}

func productionAdapterObjectiveCloseoutWriterInvocationEnvelopeRequiredRefChecks(result ProductionAdapterObjectiveCloseoutWriterInvocationEnvelope) []productionAdapterObjectiveCloseoutWriterEnvelopeRefCheck {
	return []productionAdapterObjectiveCloseoutWriterEnvelopeRefCheck{
		{result.InvocationEnvelopeRef, "writer_invocation_envelope_ref_missing", "host:objective_closeout_writer_invocation_envelope_ref", "provide_objective_closeout_writer_invocation_envelope", FailureEvidenceMissing},
		{result.WriterInvocationRef, "writer_invocation_ref_missing", "host:objective_closeout_writer_invocation_ref", "provide_objective_closeout_writer_invocation_ref", FailureEvidenceMissing},
		{result.HostAdapterVersionRef, "host_adapter_version_ref_missing", "host:objective_closeout_writer_adapter_version_ref", "provide_objective_closeout_writer_adapter_version", FailureConfigMissing},
		{result.ExpectedHostAdapterRunRef, "expected_host_adapter_run_ref_missing", "host:objective_closeout_writer_expected_adapter_run_ref", "provide_objective_closeout_writer_expected_refs", FailureEvidenceMissing},
		{result.DryRunProofRef, "dry_run_proof_ref_missing", "host:objective_closeout_writer_dry_run_proof_ref", "provide_objective_closeout_writer_dry_run_proof", FailureEvidenceMissing},
		{result.HostDurableWriteConfirmationRef, "host_durable_write_confirmation_ref_missing", "host:objective_closeout_writer_durable_write_confirmation_ref", "request_objective_closeout_writer_durable_write_confirmation", FailureAuthorizationMissing},
		{result.ExpectedDurableResultRef, "expected_durable_result_ref_missing", "host:objective_closeout_writer_expected_durable_result_ref", "provide_objective_closeout_writer_expected_refs", FailureEvidenceMissing},
		{result.ExpectedReadbackRef, "expected_readback_ref_missing", "host:objective_closeout_writer_expected_readback_ref", "provide_objective_closeout_writer_expected_refs", FailureEvidenceMissing},
		{result.ExpectedFailureRef, "expected_failure_ref_missing", "host:objective_closeout_writer_expected_failure_ref", "provide_objective_closeout_writer_expected_refs", FailureEvidenceMissing},
		{result.ExpectedCompensationRef, "expected_compensation_ref_missing", "host:objective_closeout_writer_expected_compensation_ref", "provide_objective_closeout_writer_expected_refs", FailureEvidenceMissing},
		{result.IdempotencyKeyRef, "idempotency_key_ref_missing", "host:objective_closeout_writer_idempotency_key_ref", "provide_objective_closeout_writer_idempotency_key", FailureInvalidInput},
		{result.TimeoutPolicyRef, "timeout_policy_ref_missing", "host:objective_closeout_writer_timeout_policy_ref", "provide_objective_closeout_writer_timeout_policy", FailureTimeout},
	}
}

type productionAdapterObjectiveCloseoutWriterEnvelopeMismatch struct {
	reason  string
	missing MissingInput
	failure FailureClass
}

func productionAdapterObjectiveCloseoutWriterInvocationEnvelopeMismatches(result ProductionAdapterObjectiveCloseoutWriterInvocationEnvelope, review ProductionAdapterObjectiveCloseoutWriterDurableReviewPacket) []productionAdapterObjectiveCloseoutWriterEnvelopeMismatch {
	var out []productionAdapterObjectiveCloseoutWriterEnvelopeMismatch
	if review.IdempotencyRef != "" && result.IdempotencyKeyRef != "" && result.IdempotencyKeyRef != review.IdempotencyRef {
		out = append(out, productionAdapterObjectiveCloseoutWriterEnvelopeMismatch{reason: "writer_invocation_idempotency_ref_mismatch", missing: "host:objective_closeout_writer_idempotency_key_ref", failure: FailureInvalidInput})
	}
	if review.DryRunSmokeRef != "" && result.DryRunProofRef != "" && result.DryRunProofRef != review.DryRunSmokeRef {
		out = append(out, productionAdapterObjectiveCloseoutWriterEnvelopeMismatch{reason: "writer_invocation_dry_run_proof_ref_mismatch", missing: "host:objective_closeout_writer_dry_run_proof_ref", failure: FailureVerificationFailed})
	}
	if review.HostDurableWriteConfirmationRef != "" && result.HostDurableWriteConfirmationRef != "" && result.HostDurableWriteConfirmationRef != review.HostDurableWriteConfirmationRef {
		out = append(out, productionAdapterObjectiveCloseoutWriterEnvelopeMismatch{reason: "writer_invocation_durable_confirmation_ref_mismatch", missing: "host:objective_closeout_writer_durable_write_confirmation_ref", failure: FailureAuthorizationMissing})
	}
	if review.ExpectedDurableResultRef != "" && result.ExpectedDurableResultRef != "" && result.ExpectedDurableResultRef != review.ExpectedDurableResultRef {
		out = append(out, productionAdapterObjectiveCloseoutWriterEnvelopeMismatch{reason: "writer_invocation_expected_result_ref_mismatch", missing: "host:objective_closeout_writer_expected_durable_result_ref", failure: FailureVerificationFailed})
	}
	if review.ExpectedReadbackRef != "" && result.ExpectedReadbackRef != "" && result.ExpectedReadbackRef != review.ExpectedReadbackRef {
		out = append(out, productionAdapterObjectiveCloseoutWriterEnvelopeMismatch{reason: "writer_invocation_expected_readback_ref_mismatch", missing: "host:objective_closeout_writer_expected_readback_ref", failure: FailureVerificationFailed})
	}
	return out
}

func productionAdapterObjectiveCloseoutWriterInvocationResultEnvelopeCommonMismatches(result ProductionAdapterObjectiveCloseoutWriterInvocationResultEnvelope, envelope ProductionAdapterObjectiveCloseoutWriterInvocationEnvelope, durableResult ProductionAdapterObjectiveCloseoutWriterDurableResult) []productionAdapterObjectiveCloseoutWriterEnvelopeMismatch {
	var out []productionAdapterObjectiveCloseoutWriterEnvelopeMismatch
	if envelope.DurableRequestRef != "" && durableResult.DurableRequestRef != "" && envelope.DurableRequestRef != durableResult.DurableRequestRef {
		out = append(out, productionAdapterObjectiveCloseoutWriterEnvelopeMismatch{reason: "writer_invocation_result_request_ref_mismatch", missing: "host:objective_closeout_writer_durable_request_ref", failure: FailureVerificationFailed})
	}
	if envelope.WriterRef != "" && durableResult.WriterRef != "" && envelope.WriterRef != durableResult.WriterRef {
		out = append(out, productionAdapterObjectiveCloseoutWriterEnvelopeMismatch{reason: "writer_invocation_result_writer_ref_mismatch", missing: "host:objective_closeout_writer_ref", failure: FailureVerificationFailed})
	}
	if envelope.HostWriterBindingRef != "" && durableResult.HostWriterBindingRef != "" && envelope.HostWriterBindingRef != durableResult.HostWriterBindingRef {
		out = append(out, productionAdapterObjectiveCloseoutWriterEnvelopeMismatch{reason: "writer_invocation_result_binding_ref_mismatch", missing: "host:objective_closeout_writer_binding_ref", failure: FailureVerificationFailed})
	}
	if envelope.IdempotencyKeyRef != "" && durableResult.IdempotencyRef != "" && envelope.IdempotencyKeyRef != durableResult.IdempotencyRef {
		out = append(out, productionAdapterObjectiveCloseoutWriterEnvelopeMismatch{reason: "writer_invocation_result_idempotency_ref_mismatch", missing: "host:objective_closeout_writer_idempotency_key_ref", failure: FailureInvalidInput})
	}
	if envelope.ExpectedHostAdapterRunRef != "" && durableResult.HostAdapterRunRef != "" && envelope.ExpectedHostAdapterRunRef != durableResult.HostAdapterRunRef {
		out = append(out, productionAdapterObjectiveCloseoutWriterEnvelopeMismatch{reason: "writer_invocation_result_adapter_run_ref_mismatch", missing: "host:objective_closeout_writer_host_adapter_run_ref", failure: FailureVerificationFailed})
	}
	if envelope.ExpectedDurableResultRef != "" && durableResult.DurableResultRef != "" && envelope.ExpectedDurableResultRef != durableResult.DurableResultRef {
		out = append(out, productionAdapterObjectiveCloseoutWriterEnvelopeMismatch{reason: "writer_invocation_result_durable_result_ref_mismatch", missing: "host:objective_closeout_writer_durable_result_ref", failure: FailureVerificationFailed})
	}
	if durableResult.ExpectedReadbackRef == "" {
		out = append(out, productionAdapterObjectiveCloseoutWriterEnvelopeMismatch{reason: "writer_invocation_result_expected_readback_ref_missing", missing: "host:objective_closeout_writer_expected_readback_ref", failure: FailureEvidenceMissing})
	} else if envelope.ExpectedReadbackRef != "" && envelope.ExpectedReadbackRef != durableResult.ExpectedReadbackRef {
		out = append(out, productionAdapterObjectiveCloseoutWriterEnvelopeMismatch{reason: "writer_invocation_result_expected_readback_ref_mismatch", missing: "host:objective_closeout_writer_expected_readback_ref", failure: FailureVerificationFailed})
	}
	if result.HostAdapterRunRef == "" {
		out = append(out, productionAdapterObjectiveCloseoutWriterEnvelopeMismatch{reason: "writer_invocation_result_adapter_run_ref_missing", missing: "host:objective_closeout_writer_host_adapter_run_ref", failure: FailureEvidenceMissing})
	}
	if result.DurableResultRef == "" {
		out = append(out, productionAdapterObjectiveCloseoutWriterEnvelopeMismatch{reason: "writer_invocation_result_durable_result_ref_missing", missing: "host:objective_closeout_writer_durable_result_ref", failure: FailureEvidenceMissing})
	}
	return out
}

func productionAdapterObjectiveCloseoutWriterInvocationResultEnvelopeSuccessMismatches(result ProductionAdapterObjectiveCloseoutWriterInvocationResultEnvelope) []productionAdapterObjectiveCloseoutWriterEnvelopeMismatch {
	var out []productionAdapterObjectiveCloseoutWriterEnvelopeMismatch
	if result.AppliedDurableEventRef == "" {
		out = append(out, productionAdapterObjectiveCloseoutWriterEnvelopeMismatch{reason: "writer_invocation_result_applied_durable_event_ref_missing", missing: "host:applied_durable_event_ref", failure: FailureEvidenceMissing})
	} else if result.ExpectedDurableEventRef != "" && result.AppliedDurableEventRef != result.ExpectedDurableEventRef {
		out = append(out, productionAdapterObjectiveCloseoutWriterEnvelopeMismatch{reason: "writer_invocation_result_durable_event_ref_mismatch", missing: "host:durable_event_ref", failure: FailureVerificationFailed})
	}
	if result.AppliedRunstoreRef == "" {
		out = append(out, productionAdapterObjectiveCloseoutWriterEnvelopeMismatch{reason: "writer_invocation_result_applied_runstore_ref_missing", missing: "host:applied_runstore_ref", failure: FailureEvidenceMissing})
	} else if result.HostRunstoreRef != "" && result.AppliedRunstoreRef != result.HostRunstoreRef {
		out = append(out, productionAdapterObjectiveCloseoutWriterEnvelopeMismatch{reason: "writer_invocation_result_runstore_ref_mismatch", missing: "host:runstore_ref", failure: FailureVerificationFailed})
	}
	if result.AppliedObjectiveStateRef == "" {
		out = append(out, productionAdapterObjectiveCloseoutWriterEnvelopeMismatch{reason: "writer_invocation_result_applied_objective_state_ref_missing", missing: "host:applied_objective_state_ref", failure: FailureEvidenceMissing})
	} else if result.ExpectedObjectiveStateRef != "" && result.AppliedObjectiveStateRef != result.ExpectedObjectiveStateRef {
		out = append(out, productionAdapterObjectiveCloseoutWriterEnvelopeMismatch{reason: "writer_invocation_result_objective_state_ref_mismatch", missing: "host:objective_state_ref", failure: FailureVerificationFailed})
	}
	return out
}

func productionAdapterObjectiveCloseoutWriterInvocationResultEnvelopeFailureMismatches(result ProductionAdapterObjectiveCloseoutWriterInvocationResultEnvelope) []productionAdapterObjectiveCloseoutWriterEnvelopeMismatch {
	var out []productionAdapterObjectiveCloseoutWriterEnvelopeMismatch
	if result.FailureRef == "" {
		out = append(out, productionAdapterObjectiveCloseoutWriterEnvelopeMismatch{reason: "writer_invocation_result_failure_ref_missing", missing: "host:objective_closeout_writer_durable_failure_ref", failure: FailureEvidenceMissing})
	} else if result.ExpectedFailureRef != "" && result.FailureRef != result.ExpectedFailureRef {
		out = append(out, productionAdapterObjectiveCloseoutWriterEnvelopeMismatch{reason: "writer_invocation_result_failure_ref_mismatch", missing: "host:objective_closeout_writer_durable_failure_ref", failure: FailureVerificationFailed})
	}
	if result.CompensationRef == "" {
		out = append(out, productionAdapterObjectiveCloseoutWriterEnvelopeMismatch{reason: "writer_invocation_result_compensation_ref_missing", missing: "host:objective_closeout_writer_compensation_ref", failure: FailureEvidenceMissing})
	} else if result.ExpectedCompensationRef != "" && result.CompensationRef != result.ExpectedCompensationRef {
		out = append(out, productionAdapterObjectiveCloseoutWriterEnvelopeMismatch{reason: "writer_invocation_result_compensation_ref_mismatch", missing: "host:objective_closeout_writer_compensation_ref", failure: FailureVerificationFailed})
	}
	return out
}

func productionAdapterObjectiveCloseoutWriterInvocationEnvelopeUnsafe(input ProductionAdapterObjectiveCloseoutWriterInvocationEnvelopeInput) bool {
	return input.RawOutputLoaded ||
		displaySafeRefRejected(input.InvocationEnvelopeRef) ||
		displaySafeRefRejected(input.WriterInvocationRef) ||
		displaySafeRefRejected(input.HostAdapterVersionRef) ||
		displaySafeRefSliceRejected(input.CapabilityProofRefs) ||
		displaySafeRefSliceRejected(input.ApprovalBindingRefs) ||
		displaySafeRefRejected(input.IdempotencyKeyRef) ||
		displaySafeRefRejected(input.DryRunProofRef) ||
		displaySafeRefRejected(input.HostDurableWriteConfirmationRef) ||
		displaySafeRefRejected(input.ExpectedHostAdapterRunRef) ||
		displaySafeRefRejected(input.ExpectedDurableResultRef) ||
		displaySafeRefRejected(input.ExpectedReadbackRef) ||
		displaySafeRefRejected(input.ExpectedFailureRef) ||
		displaySafeRefRejected(input.ExpectedCompensationRef) ||
		displaySafeRefRejected(input.TimeoutPolicyRef)
}

func productionAdapterObjectiveCloseoutWriterInvocationEnvelopeUnsafeOutput(input ProductionAdapterObjectiveCloseoutWriterInvocationEnvelope) bool {
	return displaySafeRefRejected(input.InvocationEnvelopeRef) ||
		displaySafeRefRejected(input.ReviewPacketRef) ||
		displaySafeRefRejected(input.DurableRequestRef) ||
		displaySafeRefRejected(input.WriterInvocationRef) ||
		displaySafeRefRejected(input.WriterOptInRef) ||
		displaySafeRefRejected(input.WriterRef) ||
		displaySafeRefRejected(input.OwnerRef) ||
		displaySafeRefRejected(input.HostWriterBindingRef) ||
		displaySafeRefRejected(input.HostAdapterVersionRef) ||
		displaySafeRefRejected(input.ExpectedHostAdapterRunRef) ||
		displaySafeRefRejected(input.DryRunSmokeRef) ||
		displaySafeRefRejected(input.DryRunResultRef) ||
		displaySafeRefRejected(input.DryRunProofRef) ||
		displaySafeRefRejected(input.ExpectedDurableResultRef) ||
		displaySafeRefRejected(input.ExpectedReadbackRef) ||
		displaySafeRefRejected(input.ExpectedFailureRef) ||
		displaySafeRefRejected(input.ExpectedCompensationRef) ||
		displaySafeRefRejected(input.ObjectiveCloseoutHandoffRef) ||
		displaySafeRefRejected(input.HostUIHandoffRef) ||
		displaySafeRefRejected(input.ObjectiveCloseoutPacketRef) ||
		displaySafeRefRejected(input.ObjectiveRef) ||
		displaySafeRefRejected(input.HostObjectiveLifecycleRef) ||
		displaySafeRefRejected(input.HostRunstoreRef) ||
		displaySafeRefRejected(input.ExpectedDurableEventRef) ||
		displaySafeRefRejected(input.ExpectedObjectiveStateRef) ||
		displaySafeRefRejected(input.HostDurableApplyConfirmationRef) ||
		displaySafeRefRejected(input.HostDurableWriteConfirmationRef) ||
		displaySafeRefSliceRejected(input.CapabilityProofRefs) ||
		displaySafeRefSliceRejected(input.ApprovalBindingRefs) ||
		displaySafeRefRejected(input.IdempotencyKeyRef) ||
		displaySafeRefRejected(input.IdempotencyContractRef) ||
		displaySafeRefRejected(input.ReadbackContractRef) ||
		displaySafeRefRejected(input.RollbackReviewRef) ||
		displaySafeRefRejected(input.CompensationReviewRef) ||
		displaySafeRefRejected(input.TimeoutPolicyRef) ||
		input.RawOutputLoaded
}

func productionAdapterObjectiveCloseoutWriterInvocationResultEnvelopeUnsafe(input ProductionAdapterObjectiveCloseoutWriterInvocationResultEnvelopeInput, durableResult ProductionAdapterObjectiveCloseoutWriterDurableResult, resultProvided bool) bool {
	return input.RawOutputLoaded ||
		displaySafeRefRejected(input.ResultEnvelopeRef) ||
		(resultProvided && productionAdapterObjectiveCloseoutWriterDurableResultOutputUnsafe(input.DurableResult)) ||
		(resultProvided && productionAdapterObjectiveCloseoutWriterDurableResultOutputUnsafe(durableResult))
}

func productionAdapterObjectiveCloseoutWriterInvocationResultEnvelopeUnsafeOutput(input ProductionAdapterObjectiveCloseoutWriterInvocationResultEnvelope) bool {
	return displaySafeRefRejected(input.ResultEnvelopeRef) ||
		displaySafeRefRejected(input.InvocationEnvelopeRef) ||
		displaySafeRefRejected(input.ReviewPacketRef) ||
		displaySafeRefRejected(input.DurableRequestRef) ||
		displaySafeRefRejected(input.DurableResultRef) ||
		displaySafeRefRejected(input.ExpectedDurableResultRef) ||
		displaySafeRefRejected(input.WriterInvocationRef) ||
		displaySafeRefRejected(input.WriterOptInRef) ||
		displaySafeRefRejected(input.WriterRef) ||
		displaySafeRefRejected(input.HostWriterBindingRef) ||
		displaySafeRefRejected(input.HostAdapterVersionRef) ||
		displaySafeRefRejected(input.ExpectedHostAdapterRunRef) ||
		displaySafeRefRejected(input.HostAdapterRunRef) ||
		displaySafeRefRejected(input.DryRunSmokeRef) ||
		displaySafeRefRejected(input.DryRunResultRef) ||
		displaySafeRefRejected(input.ExpectedReadbackRef) ||
		displaySafeRefRejected(input.ExpectedFailureRef) ||
		displaySafeRefRejected(input.ExpectedCompensationRef) ||
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
		displaySafeRefRejected(input.HostDurableWriteConfirmationRef) ||
		displaySafeRefSliceRejected(input.CapabilityProofRefs) ||
		displaySafeRefSliceRejected(input.ApprovalBindingRefs) ||
		displaySafeRefRejected(input.IdempotencyKeyRef) ||
		displaySafeRefRejected(input.IdempotencyContractRef) ||
		displaySafeRefRejected(input.ReadbackContractRef) ||
		displaySafeRefRejected(input.RollbackReviewRef) ||
		displaySafeRefRejected(input.CompensationReviewRef) ||
		displaySafeRefSliceRejected(input.DurableEvidenceRefs) ||
		input.RawOutputLoaded
}

func productionAdapterObjectiveCloseoutWriterInvocationEnvelopeEmpty(envelope ProductionAdapterObjectiveCloseoutWriterInvocationEnvelope) bool {
	return !envelope.Projected &&
		!envelope.Available &&
		envelope.Status == "" &&
		envelope.Mode == "" &&
		!envelope.ReadyForHostDisplay &&
		!envelope.ReadyForHostAdapterInvocation &&
		!envelope.HostAdapterInvocationAuthorized &&
		envelope.InvocationEnvelopeRef == "" &&
		envelope.ReviewPacketRef == "" &&
		envelope.WriterInvocationRef == "" &&
		envelope.WriterRef == "" &&
		len(envelope.MissingInputs) == 0 &&
		len(envelope.BlockedReasons) == 0 &&
		len(envelope.Boundaries) == 0 &&
		envelope.NextHostAction == "" &&
		!envelope.RawOutputLoaded
}

func productionAdapterObjectiveCloseoutWriterInvocationEnvelopeRequiredInputs() []MissingInput {
	return []MissingInput{
		"host:objective_closeout_writer_invocation_envelope_ref",
		"host:objective_closeout_writer_durable_review_packet",
		"host:objective_closeout_writer_invocation_ref",
		"host:objective_closeout_writer_adapter_version_ref",
		"host:objective_closeout_writer_capability_proof_ref",
		"host:objective_closeout_writer_approval_binding_ref",
		"host:objective_closeout_writer_idempotency_key_ref",
		"host:objective_closeout_writer_dry_run_proof_ref",
		"host:objective_closeout_writer_durable_write_confirmation_ref",
		"host:objective_closeout_writer_expected_adapter_run_ref",
		"host:objective_closeout_writer_expected_durable_result_ref",
		"host:objective_closeout_writer_expected_readback_ref",
		"host:objective_closeout_writer_expected_failure_ref",
		"host:objective_closeout_writer_expected_compensation_ref",
		"host:objective_closeout_writer_timeout_policy_ref",
	}
}

func productionAdapterObjectiveCloseoutWriterInvocationResultEnvelopeRequiredInputs() []MissingInput {
	return []MissingInput{
		"host:objective_closeout_writer_invocation_result_envelope_ref",
		"host:objective_closeout_writer_invocation_envelope",
		"host:objective_closeout_writer_durable_result",
		"host:objective_closeout_writer_host_adapter_run_ref",
		"host:objective_closeout_writer_durable_result_ref",
		"host:applied_durable_event_ref",
		"host:applied_runstore_ref",
		"host:applied_objective_state_ref",
		"host:objective_closeout_writer_durable_failure_ref",
		"host:objective_closeout_writer_compensation_ref",
	}
}

func productionAdapterObjectiveCloseoutWriterInvocationEnvelopeBoundaries(groups ...[]Boundary) []Boundary {
	return MergeBoundaries(
		[]Boundary{
			"production_adapter_objective_closeout_writer_invocation_envelope",
			"objective_closeout_writer_invocation_envelope_projection_only",
			"host_owned_objective_closeout_writer_adapter",
			"host_writer_adapter_invocation_authorization",
			"display_safe_refs_only",
			"display_safe_result_refs_only",
			"no_runner_dispatch",
			"no_adapter_invocation",
			"no_dry_run_by_core",
			"no_durable_write_by_core",
			"no_objective_store_write_by_core",
			"no_runstore_write_by_core",
		},
		MergeBoundaries(groups...),
	)
}

func productionAdapterObjectiveCloseoutWriterInvocationResultEnvelopeBoundaries(groups ...[]Boundary) []Boundary {
	return MergeBoundaries(
		[]Boundary{
			"production_adapter_objective_closeout_writer_invocation_result_envelope",
			"objective_closeout_writer_invocation_result_envelope_projection_only",
			"host_owned_objective_closeout_writer_adapter",
			"host_writer_adapter_invocation_result_binding",
			"display_safe_refs_only",
			"display_safe_result_refs_only",
			"no_runner_dispatch",
			"no_adapter_invocation",
			"no_dry_run_by_core",
			"no_durable_write_by_core",
			"no_objective_store_write_by_core",
			"no_runstore_write_by_core",
		},
		MergeBoundaries(groups...),
	)
}

func unavailableProductionAdapterObjectiveCloseoutWriterInvocationEnvelope() ProductionAdapterObjectiveCloseoutWriterInvocationEnvelope {
	return ProductionAdapterObjectiveCloseoutWriterInvocationEnvelope{
		ContractVersion: ContractVersion,
		Projected:       true,
		Available:       false,
		Status:          "unavailable",
		Mode:            "production_adapter_objective_closeout_writer_invocation_envelope",
		RequiredInputs:  productionAdapterObjectiveCloseoutWriterInvocationEnvelopeRequiredInputs(),
		Boundaries: []Boundary{
			"production_adapter_objective_closeout_writer_invocation_envelope",
			"objective_closeout_writer_invocation_envelope_projection_only",
			"host_owned_objective_closeout_writer_adapter",
			"display_safe_refs_only",
			"no_adapter_invocation",
			"no_runner_dispatch",
			"no_durable_write_by_core",
		},
		NextHostAction: "provide_objective_closeout_writer_durable_review_packet",
		RunnerEffect:   "none",
		PromptEffect:   "none",
	}
}

func unavailableProductionAdapterObjectiveCloseoutWriterInvocationResultEnvelope() ProductionAdapterObjectiveCloseoutWriterInvocationResultEnvelope {
	return ProductionAdapterObjectiveCloseoutWriterInvocationResultEnvelope{
		ContractVersion: ContractVersion,
		Projected:       true,
		Available:       false,
		Status:          "unavailable",
		Mode:            "production_adapter_objective_closeout_writer_invocation_result_envelope",
		RequiredInputs:  productionAdapterObjectiveCloseoutWriterInvocationResultEnvelopeRequiredInputs(),
		Boundaries: []Boundary{
			"production_adapter_objective_closeout_writer_invocation_result_envelope",
			"objective_closeout_writer_invocation_result_envelope_projection_only",
			"host_owned_objective_closeout_writer_adapter",
			"display_safe_refs_only",
			"no_adapter_invocation",
			"no_runner_dispatch",
			"no_durable_write_by_core",
		},
		NextHostAction: "provide_objective_closeout_writer_invocation_envelope",
		RunnerEffect:   "none",
		PromptEffect:   "none",
	}
}
