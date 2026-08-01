package controlcontract

type ProductionAdapterObjectiveCloseoutWriterExecutionBridgeInput struct {
	BridgeRef         DisplaySafeRef                                                  `json:"bridge_ref,omitempty"`
	ResultEnvelopeRef DisplaySafeRef                                                  `json:"result_envelope_ref,omitempty"`
	InvocationHandoff ProductionAdapterObjectiveCloseoutWriterInvocationHostUIHandoff `json:"invocation_handoff,omitempty"`
	DurableRequest    ProductionAdapterObjectiveCloseoutWriterDurableRequest          `json:"durable_request,omitempty"`
	DurableResult     ProductionAdapterObjectiveCloseoutWriterDurableResult           `json:"durable_result,omitempty"`
	RawOutputLoaded   bool                                                            `json:"raw_output_loaded"`
}

type ProductionAdapterObjectiveCloseoutWriterExecutionBridge struct {
	ContractVersion                  string           `json:"contract_version,omitempty"`
	Projected                        bool             `json:"projected"`
	Available                        bool             `json:"available"`
	Status                           string           `json:"status,omitempty"`
	Mode                             string           `json:"mode,omitempty"`
	ReadyForHostDisplay              bool             `json:"ready_for_host_display"`
	ReadyForHostAdapterExecution     bool             `json:"ready_for_host_adapter_execution"`
	ReadyForInvocationResultEnvelope bool             `json:"ready_for_invocation_result_envelope"`
	ReadyForDurableReadbackReview    bool             `json:"ready_for_durable_readback_review"`
	ReadyForFailureReview            bool             `json:"ready_for_failure_review"`
	ReadyForCompensationReview       bool             `json:"ready_for_compensation_review"`
	HostAdapterExecutionAuthorized   bool             `json:"host_adapter_execution_authorized"`
	HostAdapterExecutionBound        bool             `json:"host_adapter_execution_bound"`
	HostAdapterExecutionReported     bool             `json:"host_adapter_execution_reported"`
	HostAdapterExecutionSucceeded    bool             `json:"host_adapter_execution_succeeded"`
	HostAdapterExecutionFailed       bool             `json:"host_adapter_execution_failed"`
	HostAdapterExecutionRecorded     bool             `json:"host_adapter_execution_recorded"`
	ResultCanonicalized              bool             `json:"result_canonicalized"`
	HostMayInvokeWriterAdapter       bool             `json:"host_may_invoke_writer_adapter"`
	HostMayExecuteDurableWrite       bool             `json:"host_may_execute_durable_write"`
	CoreInvocationExecuted           bool             `json:"core_invocation_executed"`
	DryRunByCore                     bool             `json:"dry_run_by_core"`
	DurableWriteByCore               bool             `json:"durable_write_by_core"`
	ObjectiveStoreWriteByCore        bool             `json:"objective_store_write_by_core"`
	RunstoreWriteByCore              bool             `json:"runstore_write_by_core"`
	BridgeRef                        DisplaySafeRef   `json:"bridge_ref,omitempty"`
	HostUIHandoffRef                 DisplaySafeRef   `json:"host_ui_handoff_ref,omitempty"`
	PrimaryDisplayRef                DisplaySafeRef   `json:"primary_display_ref,omitempty"`
	ReviewRef                        DisplaySafeRef   `json:"review_ref,omitempty"`
	FixtureRef                       DisplaySafeRef   `json:"fixture_ref,omitempty"`
	InvocationEnvelopeRef            DisplaySafeRef   `json:"invocation_envelope_ref,omitempty"`
	ResultEnvelopeRef                DisplaySafeRef   `json:"result_envelope_ref,omitempty"`
	ReviewPacketRef                  DisplaySafeRef   `json:"review_packet_ref,omitempty"`
	DurableRequestRef                DisplaySafeRef   `json:"durable_request_ref,omitempty"`
	DurableResultRef                 DisplaySafeRef   `json:"durable_result_ref,omitempty"`
	ExpectedDurableResultRef         DisplaySafeRef   `json:"expected_durable_result_ref,omitempty"`
	WriterInvocationRef              DisplaySafeRef   `json:"writer_invocation_ref,omitempty"`
	WriterOptInRef                   DisplaySafeRef   `json:"writer_opt_in_ref,omitempty"`
	WriterRef                        DisplaySafeRef   `json:"writer_ref,omitempty"`
	HostWriterBindingRef             DisplaySafeRef   `json:"host_writer_binding_ref,omitempty"`
	HostAdapterVersionRef            DisplaySafeRef   `json:"host_adapter_version_ref,omitempty"`
	ExpectedHostAdapterRunRef        DisplaySafeRef   `json:"expected_host_adapter_run_ref,omitempty"`
	HostAdapterRunRef                DisplaySafeRef   `json:"host_adapter_run_ref,omitempty"`
	DryRunSmokeRef                   DisplaySafeRef   `json:"dry_run_smoke_ref,omitempty"`
	DryRunResultRef                  DisplaySafeRef   `json:"dry_run_result_ref,omitempty"`
	ExpectedReadbackRef              DisplaySafeRef   `json:"expected_readback_ref,omitempty"`
	ExpectedFailureRef               DisplaySafeRef   `json:"expected_failure_ref,omitempty"`
	ExpectedCompensationRef          DisplaySafeRef   `json:"expected_compensation_ref,omitempty"`
	ObjectiveCloseoutHandoffRef      DisplaySafeRef   `json:"objective_closeout_handoff_ref,omitempty"`
	ObjectiveCloseoutPacketRef       DisplaySafeRef   `json:"objective_closeout_packet_ref,omitempty"`
	ObjectiveRef                     DisplaySafeRef   `json:"objective_ref,omitempty"`
	HostRunstoreRef                  DisplaySafeRef   `json:"host_runstore_ref,omitempty"`
	ExpectedDurableEventRef          DisplaySafeRef   `json:"expected_durable_event_ref,omitempty"`
	ExpectedObjectiveStateRef        DisplaySafeRef   `json:"expected_objective_state_ref,omitempty"`
	AppliedDurableEventRef           DisplaySafeRef   `json:"applied_durable_event_ref,omitempty"`
	AppliedRunstoreRef               DisplaySafeRef   `json:"applied_runstore_ref,omitempty"`
	AppliedObjectiveStateRef         DisplaySafeRef   `json:"applied_objective_state_ref,omitempty"`
	FailureRef                       DisplaySafeRef   `json:"failure_ref,omitempty"`
	CompensationRef                  DisplaySafeRef   `json:"compensation_ref,omitempty"`
	HostDurableWriteConfirmationRef  DisplaySafeRef   `json:"host_durable_write_confirmation_ref,omitempty"`
	CapabilityProofRefs              []DisplaySafeRef `json:"capability_proof_refs,omitempty"`
	ApprovalBindingRefs              []DisplaySafeRef `json:"approval_binding_refs,omitempty"`
	PolicyRefs                       []DisplaySafeRef `json:"policy_refs,omitempty"`
	RequiredPolicyRefs               []DisplaySafeRef `json:"required_policy_refs,omitempty"`
	ApprovalRefs                     []DisplaySafeRef `json:"approval_refs,omitempty"`
	RequiredApprovalRefs             []DisplaySafeRef `json:"required_approval_refs,omitempty"`
	BudgetRef                        DisplaySafeRef   `json:"budget_ref,omitempty"`
	RequiredBudgetRef                DisplaySafeRef   `json:"required_budget_ref,omitempty"`
	IdempotencyRef                   DisplaySafeRef   `json:"idempotency_ref,omitempty"`
	IdempotencyContractRef           DisplaySafeRef   `json:"idempotency_contract_ref,omitempty"`
	TimeoutPolicyRef                 DisplaySafeRef   `json:"timeout_policy_ref,omitempty"`
	DurableEvidenceRefs              []DisplaySafeRef `json:"durable_evidence_refs,omitempty"`
	RequiredInputs                   []MissingInput   `json:"required_inputs,omitempty"`
	MissingInputs                    []MissingInput   `json:"missing_inputs,omitempty"`
	BlockedReasons                   []string         `json:"blocked_reasons,omitempty"`
	FailureClass                     FailureClass     `json:"failure_class,omitempty"`
	Boundaries                       []Boundary       `json:"boundaries,omitempty"`
	NextHostAction                   NextHostAction   `json:"next_host_action,omitempty"`
	RunnerEffect                     string           `json:"runner_effect,omitempty"`
	PromptEffect                     string           `json:"prompt_effect,omitempty"`
	RawOutputLoaded                  bool             `json:"raw_output_loaded"`
}

func BuildProductionAdapterObjectiveCloseoutWriterExecutionBridge(input ProductionAdapterObjectiveCloseoutWriterExecutionBridgeInput) ProductionAdapterObjectiveCloseoutWriterExecutionBridge {
	if productionAdapterObjectiveCloseoutWriterInvocationHostUIHandoffEmpty(input.InvocationHandoff) || productionAdapterObjectiveCloseoutWriterDurableRequestEmpty(input.DurableRequest) {
		return unavailableProductionAdapterObjectiveCloseoutWriterExecutionBridge()
	}
	handoff := input.InvocationHandoff.Normalize()
	request := input.DurableRequest.Normalize()
	resultProvided := !productionAdapterObjectiveCloseoutWriterDurableResultEmpty(input.DurableResult)
	durableResult := ProductionAdapterObjectiveCloseoutWriterDurableResult{}
	if resultProvided {
		durableResult = input.DurableResult.Normalize()
	}
	result := ProductionAdapterObjectiveCloseoutWriterExecutionBridge{
		ContractVersion:                 ContractVersion,
		Projected:                       true,
		Available:                       handoff.Available && request.Available,
		Status:                          "blocked",
		Mode:                            "production_adapter_objective_closeout_writer_execution_bridge",
		BridgeRef:                       normalizeOneDisplaySafeRef(input.BridgeRef),
		HostUIHandoffRef:                handoff.HostUIHandoffRef,
		PrimaryDisplayRef:               handoff.PrimaryDisplayRef,
		ReviewRef:                       handoff.ReviewRef,
		FixtureRef:                      handoff.FixtureRef,
		InvocationEnvelopeRef:           handoff.InvocationEnvelopeRef,
		ResultEnvelopeRef:               normalizeOneDisplaySafeRef(input.ResultEnvelopeRef),
		ReviewPacketRef:                 handoff.ReviewPacketRef,
		DurableRequestRef:               firstDisplaySafeRef(durableResult.DurableRequestRef, request.DurableRequestRef, handoff.DurableRequestRef),
		DurableResultRef:                firstDisplaySafeRef(durableResult.DurableResultRef, handoff.DurableResultRef),
		ExpectedDurableResultRef:        firstDisplaySafeRef(handoff.ExpectedDurableResultRef, request.ExpectedDurableResultRef, durableResult.ExpectedDurableResultRef),
		WriterInvocationRef:             handoff.WriterInvocationRef,
		WriterOptInRef:                  request.WriterOptInRef,
		WriterRef:                       firstDisplaySafeRef(handoff.WriterRef, request.WriterRef, durableResult.WriterRef),
		HostWriterBindingRef:            firstDisplaySafeRef(handoff.HostWriterBindingRef, request.HostWriterBindingRef, durableResult.HostWriterBindingRef),
		HostAdapterVersionRef:           handoff.HostAdapterVersionRef,
		ExpectedHostAdapterRunRef:       handoff.ExpectedHostAdapterRunRef,
		HostAdapterRunRef:               durableResult.HostAdapterRunRef,
		DryRunSmokeRef:                  firstDisplaySafeRef(request.DryRunSmokeRef, durableResult.DryRunSmokeRef),
		DryRunResultRef:                 firstDisplaySafeRef(request.DryRunResultRef, durableResult.DryRunResultRef),
		ExpectedReadbackRef:             firstDisplaySafeRef(handoff.ExpectedReadbackRef, request.ExpectedReadbackRef, durableResult.ExpectedReadbackRef),
		ExpectedFailureRef:              handoff.ExpectedFailureRef,
		ExpectedCompensationRef:         handoff.ExpectedCompensationRef,
		ObjectiveCloseoutHandoffRef:     firstDisplaySafeRef(request.ObjectiveCloseoutHandoffRef, durableResult.ObjectiveCloseoutHandoffRef),
		ObjectiveCloseoutPacketRef:      firstDisplaySafeRef(request.ObjectiveCloseoutPacketRef, durableResult.ObjectiveCloseoutPacketRef),
		ObjectiveRef:                    firstDisplaySafeRef(request.ObjectiveRef, durableResult.ObjectiveRef),
		HostRunstoreRef:                 firstDisplaySafeRef(request.HostRunstoreRef, durableResult.HostRunstoreRef),
		ExpectedDurableEventRef:         firstDisplaySafeRef(request.ExpectedDurableEventRef, durableResult.ExpectedDurableEventRef),
		ExpectedObjectiveStateRef:       firstDisplaySafeRef(request.ExpectedObjectiveStateRef, durableResult.ExpectedObjectiveStateRef),
		AppliedDurableEventRef:          durableResult.AppliedDurableEventRef,
		AppliedRunstoreRef:              durableResult.AppliedRunstoreRef,
		AppliedObjectiveStateRef:        durableResult.AppliedObjectiveStateRef,
		FailureRef:                      durableResult.FailureRef,
		CompensationRef:                 durableResult.CompensationRef,
		HostDurableWriteConfirmationRef: firstDisplaySafeRef(handoff.HostDurableWriteConfirmationRef, request.HostDurableWriteConfirmationRef),
		CapabilityProofRefs:             cloneDisplaySafeRefs(handoff.CapabilityProofRefs),
		ApprovalBindingRefs:             cloneDisplaySafeRefs(handoff.ApprovalBindingRefs),
		PolicyRefs:                      cloneDisplaySafeRefs(request.PolicyRefs),
		RequiredPolicyRefs:              cloneDisplaySafeRefs(request.RequiredPolicyRefs),
		ApprovalRefs:                    cloneDisplaySafeRefs(request.ApprovalRefs),
		RequiredApprovalRefs:            cloneDisplaySafeRefs(request.RequiredApprovalRefs),
		BudgetRef:                       request.BudgetRef,
		RequiredBudgetRef:               request.RequiredBudgetRef,
		IdempotencyRef:                  request.IdempotencyRef,
		IdempotencyContractRef:          request.IdempotencyContractRef,
		TimeoutPolicyRef:                request.TimeoutPolicyRef,
		DurableEvidenceRefs:             cloneDisplaySafeRefs(durableResult.DurableEvidenceRefs),
		RequiredInputs:                  productionAdapterObjectiveCloseoutWriterExecutionBridgeRequiredInputs(),
		HostAdapterExecutionAuthorized:  handoff.HostAdapterInvocationAuthorized && request.HostDurableWriteAuthorized,
		HostAdapterExecutionReported:    durableResult.HostDurableWriteReported,
		HostAdapterExecutionSucceeded:   durableResult.HostDurableWriteSucceeded,
		HostAdapterExecutionFailed:      durableResult.HostDurableWriteFailed,
		HostAdapterExecutionRecorded:    durableResult.HostDurableWriteRecorded,
		HostMayInvokeWriterAdapter:      handoff.HostMayInvokeWriterAdapter,
		HostMayExecuteDurableWrite:      handoff.HostMayExecuteDurableWrite && request.HostMayExecuteDurableWrite,
		FailureClass:                    FailureNone,
		Boundaries:                      productionAdapterObjectiveCloseoutWriterExecutionBridgeBoundaries(handoff.Boundaries, request.Boundaries, durableResult.Boundaries),
		RunnerEffect:                    "none",
		PromptEffect:                    "none",
		RawOutputLoaded:                 input.RawOutputLoaded || handoff.RawOutputLoaded || request.RawOutputLoaded || durableResult.RawOutputLoaded,
	}
	if productionAdapterObjectiveCloseoutWriterExecutionBridgeUnsafe(input, handoff, request, durableResult, resultProvided) {
		result = productionAdapterObjectiveCloseoutWriterExecutionBridgeBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if result.BridgeRef == "" {
		result = productionAdapterObjectiveCloseoutWriterExecutionBridgeBlock(result, FailureEvidenceMissing, "writer_execution_bridge_ref_missing", "host:objective_closeout_writer_execution_bridge_ref", "provide_objective_closeout_writer_execution_bridge_ref")
	}
	if !handoff.ReadyForHostAdapterInvocation || !handoff.HostMayInvokeWriterAdapter || !handoff.HostMayExecuteDurableWrite || !handoff.HostAdapterInvocationAuthorized {
		result = productionAdapterObjectiveCloseoutWriterExecutionBridgeBlock(result, firstFailureClass(handoff.FailureClass, FailureAuthorizationMissing), "writer_invocation_handoff_not_ready", "host:objective_closeout_writer_invocation_handoff", firstNextHostAction(handoff.NextHostAction, "review_objective_closeout_writer_invocation_handoff"))
	}
	if !request.ReadyForHostDurableWrite || !request.HostMayExecuteDurableWrite || !request.HostDurableWriteAuthorized {
		result = productionAdapterObjectiveCloseoutWriterExecutionBridgeBlock(result, firstFailureClass(request.FailureClass, FailureAuthorizationMissing), "writer_durable_request_not_ready", "host:objective_closeout_writer_durable_request", firstNextHostAction(request.NextHostAction, "review_objective_closeout_writer_durable_request"))
	}
	for _, mismatch := range productionAdapterObjectiveCloseoutWriterExecutionBridgeHandoffRequestMismatches(handoff, request) {
		result = productionAdapterObjectiveCloseoutWriterExecutionBridgeBlock(result, mismatch.failure, mismatch.reason, mismatch.missing, "review_objective_closeout_writer_execution_bridge")
	}
	if len(result.BlockedReasons) == 0 && len(result.MissingInputs) == 0 && !resultProvided {
		result.Status = "ready_for_objective_closeout_writer_host_adapter_execution_bridge"
		result.ReadyForHostDisplay = true
		result.ReadyForHostAdapterExecution = true
		result.HostAdapterExecutionAuthorized = true
		result.HostMayInvokeWriterAdapter = true
		result.HostMayExecuteDurableWrite = true
		result.NextHostAction = "host_may_execute_objective_closeout_writer_adapter"
		result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_objective_closeout_writer_host_adapter_execution_bridge", "host_may_execute_writer_adapter")
		return result.Normalize()
	}
	if !resultProvided {
		return result.Normalize()
	}
	if result.ResultEnvelopeRef == "" {
		result = productionAdapterObjectiveCloseoutWriterExecutionBridgeBlock(result, FailureEvidenceMissing, "writer_invocation_result_envelope_ref_missing", "host:objective_closeout_writer_invocation_result_envelope_ref", "provide_objective_closeout_writer_invocation_result_envelope_ref")
	}
	if !durableResult.HostDurableWriteReported || !durableResult.HostDurableWriteRecorded {
		result = productionAdapterObjectiveCloseoutWriterExecutionBridgeBlock(result, firstFailureClass(durableResult.FailureClass, FailureEvidenceMissing), "writer_execution_result_not_recorded", "host:objective_closeout_writer_durable_result", firstNextHostAction(durableResult.NextHostAction, "provide_objective_closeout_writer_durable_result"))
	}
	if durableResult.HostDurableWriteSucceeded && durableResult.HostDurableWriteFailed {
		result = productionAdapterObjectiveCloseoutWriterExecutionBridgeBlock(result, FailureVerificationFailed, "writer_execution_result_conflict", "host:objective_closeout_writer_durable_result", "review_objective_closeout_writer_durable_result")
	}
	for _, mismatch := range productionAdapterObjectiveCloseoutWriterExecutionBridgeDurableResultMismatches(handoff, request, durableResult) {
		result = productionAdapterObjectiveCloseoutWriterExecutionBridgeBlock(result, mismatch.failure, mismatch.reason, mismatch.missing, "review_objective_closeout_writer_execution_bridge")
	}
	switch {
	case durableResult.HostDurableWriteSucceeded:
		if !durableResult.ReadyForWriterDurableReadback {
			result = productionAdapterObjectiveCloseoutWriterExecutionBridgeBlock(result, firstFailureClass(durableResult.FailureClass, FailureEvidenceMissing), "writer_execution_success_not_ready_for_readback", "host:objective_closeout_writer_durable_result", "review_objective_closeout_writer_durable_result")
		}
		if len(result.BlockedReasons) == 0 && len(result.MissingInputs) == 0 {
			result.Status = "ready_for_objective_closeout_writer_host_adapter_result_bridge"
			result.ReadyForHostDisplay = true
			result.ReadyForInvocationResultEnvelope = true
			result.ReadyForDurableReadbackReview = true
			result.HostAdapterExecutionBound = true
			result.ResultCanonicalized = true
			result.NextHostAction = "build_objective_closeout_writer_invocation_result_envelope"
			result.Boundaries = AppendBoundaries(result.Boundaries, "objective_closeout_writer_host_adapter_execution_result_canonicalized", "ready_for_objective_closeout_writer_invocation_result_envelope")
		}
	case durableResult.HostDurableWriteFailed:
		if durableResult.FailureRef == "" {
			result = productionAdapterObjectiveCloseoutWriterExecutionBridgeBlock(result, FailureEvidenceMissing, "writer_execution_failure_ref_missing", "host:objective_closeout_writer_durable_failure_ref", "provide_objective_closeout_writer_durable_failure")
		}
		if len(result.BlockedReasons) == 0 && len(result.MissingInputs) == 0 {
			result.Status = "ready_for_objective_closeout_writer_host_adapter_failure_bridge"
			result.ReadyForHostDisplay = true
			result.ReadyForInvocationResultEnvelope = true
			result.ReadyForFailureReview = true
			result.ReadyForCompensationReview = durableResult.CompensationRef != ""
			result.HostAdapterExecutionBound = true
			result.ResultCanonicalized = true
			result.FailureClass = FailureVerificationFailed
			result.NextHostAction = "build_objective_closeout_writer_invocation_failure_envelope"
			result.Boundaries = AppendBoundaries(result.Boundaries, "objective_closeout_writer_host_adapter_execution_failure_canonicalized", "ready_for_objective_closeout_writer_invocation_failure_envelope", "compensation_not_executed")
		}
	default:
		result = productionAdapterObjectiveCloseoutWriterExecutionBridgeBlock(result, FailureEvidenceMissing, "writer_execution_result_status_missing", "host:objective_closeout_writer_durable_result", "provide_objective_closeout_writer_durable_result")
	}
	return result.Normalize()
}

func CloneProductionAdapterObjectiveCloseoutWriterExecutionBridge(in ProductionAdapterObjectiveCloseoutWriterExecutionBridge) ProductionAdapterObjectiveCloseoutWriterExecutionBridge {
	out := in
	out.CapabilityProofRefs = cloneDisplaySafeRefs(in.CapabilityProofRefs)
	out.ApprovalBindingRefs = cloneDisplaySafeRefs(in.ApprovalBindingRefs)
	out.PolicyRefs = cloneDisplaySafeRefs(in.PolicyRefs)
	out.RequiredPolicyRefs = cloneDisplaySafeRefs(in.RequiredPolicyRefs)
	out.ApprovalRefs = cloneDisplaySafeRefs(in.ApprovalRefs)
	out.RequiredApprovalRefs = cloneDisplaySafeRefs(in.RequiredApprovalRefs)
	out.DurableEvidenceRefs = cloneDisplaySafeRefs(in.DurableEvidenceRefs)
	out.RequiredInputs = cloneMissingInputs(in.RequiredInputs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (b ProductionAdapterObjectiveCloseoutWriterExecutionBridge) Clone() ProductionAdapterObjectiveCloseoutWriterExecutionBridge {
	return CloneProductionAdapterObjectiveCloseoutWriterExecutionBridge(b)
}

func (b ProductionAdapterObjectiveCloseoutWriterExecutionBridge) Normalize() ProductionAdapterObjectiveCloseoutWriterExecutionBridge {
	out := CloneProductionAdapterObjectiveCloseoutWriterExecutionBridge(b)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = normalizeControlToken(out.Status)
	out.Mode = normalizeControlToken(out.Mode)
	if out.Mode == "" {
		out.Mode = "production_adapter_objective_closeout_writer_execution_bridge"
	}
	out.BridgeRef = normalizeOneDisplaySafeRef(out.BridgeRef)
	out.HostUIHandoffRef = normalizeOneDisplaySafeRef(out.HostUIHandoffRef)
	out.PrimaryDisplayRef = normalizeOneDisplaySafeRef(out.PrimaryDisplayRef)
	out.ReviewRef = normalizeOneDisplaySafeRef(out.ReviewRef)
	out.FixtureRef = normalizeOneDisplaySafeRef(out.FixtureRef)
	out.InvocationEnvelopeRef = normalizeOneDisplaySafeRef(out.InvocationEnvelopeRef)
	out.ResultEnvelopeRef = normalizeOneDisplaySafeRef(out.ResultEnvelopeRef)
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
	out.PolicyRefs = normalizeDisplaySafeRefs(out.PolicyRefs)
	out.RequiredPolicyRefs = normalizeDisplaySafeRefs(out.RequiredPolicyRefs)
	out.ApprovalRefs = normalizeDisplaySafeRefs(out.ApprovalRefs)
	out.RequiredApprovalRefs = normalizeDisplaySafeRefs(out.RequiredApprovalRefs)
	out.BudgetRef = normalizeOneDisplaySafeRef(out.BudgetRef)
	out.RequiredBudgetRef = normalizeOneDisplaySafeRef(out.RequiredBudgetRef)
	out.IdempotencyRef = normalizeOneDisplaySafeRef(out.IdempotencyRef)
	out.IdempotencyContractRef = normalizeOneDisplaySafeRef(out.IdempotencyContractRef)
	out.TimeoutPolicyRef = normalizeOneDisplaySafeRef(out.TimeoutPolicyRef)
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
		out.ReadyForHostAdapterExecution = false
		out.ReadyForInvocationResultEnvelope = false
		out.ReadyForDurableReadbackReview = false
		out.ReadyForFailureReview = false
		out.ReadyForCompensationReview = false
		out.HostAdapterExecutionAuthorized = false
		out.HostAdapterExecutionBound = false
		out.HostMayInvokeWriterAdapter = false
		out.HostMayExecuteDurableWrite = false
	}
	if out.Status == "" {
		out.Status = "blocked"
	}
	if out.RawOutputLoaded || productionAdapterObjectiveCloseoutWriterExecutionBridgeUnsafeOutput(out) {
		out.RawOutputLoaded = true
		out = productionAdapterObjectiveCloseoutWriterExecutionBridgeBlock(out, firstFailureClass(out.FailureClass, FailureEvidenceWeak), "unsafe_input_ref", "host:display_safe_refs", firstNextHostAction(out.NextHostAction, "provide_display_safe_refs"))
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
	}
	out.ReadyForHostDisplay = out.ReadyForHostDisplay &&
		out.Available &&
		out.BridgeRef != "" &&
		out.HostUIHandoffRef != "" &&
		out.ReviewRef != "" &&
		out.InvocationEnvelopeRef != "" &&
		out.DurableRequestRef != "" &&
		out.WriterRef != "" &&
		!out.RawOutputLoaded
	executionReady := out.Status == "ready_for_objective_closeout_writer_host_adapter_execution_bridge" &&
		out.ReadyForHostDisplay &&
		out.HostAdapterExecutionAuthorized &&
		out.BridgeRef != "" &&
		out.HostUIHandoffRef != "" &&
		out.InvocationEnvelopeRef != "" &&
		out.DurableRequestRef != "" &&
		out.WriterInvocationRef != "" &&
		out.WriterRef != "" &&
		out.HostWriterBindingRef != "" &&
		out.HostAdapterVersionRef != "" &&
		out.ExpectedHostAdapterRunRef != "" &&
		out.ExpectedDurableResultRef != "" &&
		out.ExpectedReadbackRef != "" &&
		out.ExpectedFailureRef != "" &&
		out.ExpectedCompensationRef != "" &&
		out.HostDurableWriteConfirmationRef != "" &&
		len(out.CapabilityProofRefs) > 0 &&
		len(out.ApprovalBindingRefs) > 0 &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.RawOutputLoaded
	successReady := out.Status == "ready_for_objective_closeout_writer_host_adapter_result_bridge" &&
		out.ReadyForHostDisplay &&
		out.ResultCanonicalized &&
		out.HostAdapterExecutionBound &&
		out.HostAdapterExecutionReported &&
		out.HostAdapterExecutionRecorded &&
		out.HostAdapterExecutionSucceeded &&
		!out.HostAdapterExecutionFailed &&
		out.ResultEnvelopeRef != "" &&
		out.DurableResultRef != "" &&
		out.ExpectedDurableResultRef == out.DurableResultRef &&
		out.HostAdapterRunRef != "" &&
		out.ExpectedHostAdapterRunRef == out.HostAdapterRunRef &&
		out.AppliedDurableEventRef != "" &&
		out.AppliedDurableEventRef == out.ExpectedDurableEventRef &&
		out.AppliedRunstoreRef != "" &&
		out.AppliedRunstoreRef == out.HostRunstoreRef &&
		out.AppliedObjectiveStateRef != "" &&
		out.AppliedObjectiveStateRef == out.ExpectedObjectiveStateRef &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.RawOutputLoaded
	failureReady := out.Status == "ready_for_objective_closeout_writer_host_adapter_failure_bridge" &&
		out.ReadyForHostDisplay &&
		out.ResultCanonicalized &&
		out.HostAdapterExecutionBound &&
		out.HostAdapterExecutionReported &&
		out.HostAdapterExecutionRecorded &&
		out.HostAdapterExecutionFailed &&
		!out.HostAdapterExecutionSucceeded &&
		out.ResultEnvelopeRef != "" &&
		out.DurableResultRef != "" &&
		out.ExpectedDurableResultRef == out.DurableResultRef &&
		out.HostAdapterRunRef != "" &&
		out.ExpectedHostAdapterRunRef == out.HostAdapterRunRef &&
		out.FailureRef != "" &&
		out.ExpectedFailureRef == out.FailureRef &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.RawOutputLoaded
	out.ReadyForHostAdapterExecution = out.ReadyForHostAdapterExecution && executionReady
	out.ReadyForInvocationResultEnvelope = out.ReadyForInvocationResultEnvelope && (successReady || failureReady)
	out.ReadyForDurableReadbackReview = out.ReadyForDurableReadbackReview && successReady
	out.ReadyForFailureReview = out.ReadyForFailureReview && failureReady
	out.ReadyForCompensationReview = out.ReadyForCompensationReview && failureReady && out.CompensationRef != ""
	out.HostAdapterExecutionAuthorized = out.HostAdapterExecutionAuthorized && (executionReady || successReady || failureReady)
	out.HostAdapterExecutionBound = out.HostAdapterExecutionBound && (successReady || failureReady)
	out.ResultCanonicalized = out.ResultCanonicalized && (successReady || failureReady)
	out.HostMayInvokeWriterAdapter = out.HostMayInvokeWriterAdapter && executionReady
	out.HostMayExecuteDurableWrite = out.HostMayExecuteDurableWrite && executionReady
	return out
}

func productionAdapterObjectiveCloseoutWriterExecutionBridgeBlock(result ProductionAdapterObjectiveCloseoutWriterExecutionBridge, failure FailureClass, reason string, missing MissingInput, next NextHostAction) ProductionAdapterObjectiveCloseoutWriterExecutionBridge {
	result.Status = "blocked"
	result.ReadyForHostAdapterExecution = false
	result.ReadyForInvocationResultEnvelope = false
	result.ReadyForDurableReadbackReview = false
	result.ReadyForFailureReview = false
	result.ReadyForCompensationReview = false
	result.HostAdapterExecutionBound = false
	result.ResultCanonicalized = false
	result.HostMayInvokeWriterAdapter = false
	result.HostMayExecuteDurableWrite = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	result.Boundaries = AppendBoundaries(result.Boundaries, "objective_closeout_writer_execution_bridge_blocked")
	return result
}

type productionAdapterObjectiveCloseoutWriterExecutionBridgeMismatch struct {
	failure FailureClass
	reason  string
	missing MissingInput
}

func productionAdapterObjectiveCloseoutWriterExecutionBridgeHandoffRequestMismatches(handoff ProductionAdapterObjectiveCloseoutWriterInvocationHostUIHandoff, request ProductionAdapterObjectiveCloseoutWriterDurableRequest) []productionAdapterObjectiveCloseoutWriterExecutionBridgeMismatch {
	var out []productionAdapterObjectiveCloseoutWriterExecutionBridgeMismatch
	out = append(out, productionAdapterObjectiveCloseoutWriterExecutionBridgeRefMismatch(handoff.DurableRequestRef, request.DurableRequestRef, "writer_execution_bridge_durable_request_ref_mismatch", "host:objective_closeout_writer_durable_request_ref")...)
	out = append(out, productionAdapterObjectiveCloseoutWriterExecutionBridgeRefMismatch(handoff.ExpectedDurableResultRef, request.ExpectedDurableResultRef, "writer_execution_bridge_expected_durable_result_ref_mismatch", "host:objective_closeout_writer_expected_durable_result_ref")...)
	out = append(out, productionAdapterObjectiveCloseoutWriterExecutionBridgeRefMismatch(handoff.ExpectedReadbackRef, request.ExpectedReadbackRef, "writer_execution_bridge_expected_readback_ref_mismatch", "host:objective_closeout_writer_expected_readback_ref")...)
	out = append(out, productionAdapterObjectiveCloseoutWriterExecutionBridgeRefMismatch(handoff.WriterRef, request.WriterRef, "writer_execution_bridge_writer_ref_mismatch", "host:objective_closeout_writer_ref")...)
	out = append(out, productionAdapterObjectiveCloseoutWriterExecutionBridgeRefMismatch(handoff.HostWriterBindingRef, request.HostWriterBindingRef, "writer_execution_bridge_host_writer_binding_ref_mismatch", "host:objective_closeout_writer_binding_ref")...)
	out = append(out, productionAdapterObjectiveCloseoutWriterExecutionBridgeRefMismatch(handoff.HostDurableWriteConfirmationRef, request.HostDurableWriteConfirmationRef, "writer_execution_bridge_confirmation_ref_mismatch", "host:objective_closeout_writer_durable_write_confirmation_ref")...)
	return out
}

func productionAdapterObjectiveCloseoutWriterExecutionBridgeDurableResultMismatches(handoff ProductionAdapterObjectiveCloseoutWriterInvocationHostUIHandoff, request ProductionAdapterObjectiveCloseoutWriterDurableRequest, result ProductionAdapterObjectiveCloseoutWriterDurableResult) []productionAdapterObjectiveCloseoutWriterExecutionBridgeMismatch {
	var out []productionAdapterObjectiveCloseoutWriterExecutionBridgeMismatch
	out = append(out, productionAdapterObjectiveCloseoutWriterExecutionBridgeRefMismatch(request.DurableRequestRef, result.DurableRequestRef, "writer_execution_bridge_result_durable_request_ref_mismatch", "host:objective_closeout_writer_durable_request_ref")...)
	out = append(out, productionAdapterObjectiveCloseoutWriterExecutionBridgeRefMismatch(request.ExpectedDurableResultRef, result.DurableResultRef, "writer_execution_bridge_result_durable_result_ref_mismatch", "host:objective_closeout_writer_durable_result_ref")...)
	out = append(out, productionAdapterObjectiveCloseoutWriterExecutionBridgeRefMismatch(handoff.ExpectedHostAdapterRunRef, result.HostAdapterRunRef, "writer_execution_bridge_adapter_run_ref_mismatch", "host:objective_closeout_writer_host_adapter_run_ref")...)
	out = append(out, productionAdapterObjectiveCloseoutWriterExecutionBridgeRefMismatch(handoff.ExpectedReadbackRef, result.ExpectedReadbackRef, "writer_execution_bridge_result_expected_readback_ref_mismatch", "host:objective_closeout_writer_expected_readback_ref")...)
	out = append(out, productionAdapterObjectiveCloseoutWriterExecutionBridgeRefMismatch(handoff.WriterRef, result.WriterRef, "writer_execution_bridge_result_writer_ref_mismatch", "host:objective_closeout_writer_ref")...)
	if result.HostDurableWriteSucceeded {
		out = append(out, productionAdapterObjectiveCloseoutWriterExecutionBridgeRefMismatch(request.ExpectedDurableEventRef, result.AppliedDurableEventRef, "writer_execution_bridge_applied_durable_event_ref_mismatch", "host:applied_durable_event_ref")...)
		out = append(out, productionAdapterObjectiveCloseoutWriterExecutionBridgeRefMismatch(request.HostRunstoreRef, result.AppliedRunstoreRef, "writer_execution_bridge_applied_runstore_ref_mismatch", "host:applied_runstore_ref")...)
		out = append(out, productionAdapterObjectiveCloseoutWriterExecutionBridgeRefMismatch(request.ExpectedObjectiveStateRef, result.AppliedObjectiveStateRef, "writer_execution_bridge_applied_objective_state_ref_mismatch", "host:applied_objective_state_ref")...)
	}
	if result.HostDurableWriteFailed {
		out = append(out, productionAdapterObjectiveCloseoutWriterExecutionBridgeRefMismatch(handoff.ExpectedFailureRef, result.FailureRef, "writer_execution_bridge_failure_ref_mismatch", "host:objective_closeout_writer_durable_failure_ref")...)
		if result.CompensationRef != "" {
			out = append(out, productionAdapterObjectiveCloseoutWriterExecutionBridgeRefMismatch(handoff.ExpectedCompensationRef, result.CompensationRef, "writer_execution_bridge_compensation_ref_mismatch", "host:objective_closeout_writer_compensation_ref")...)
		}
	}
	return out
}

func productionAdapterObjectiveCloseoutWriterExecutionBridgeRefMismatch(left DisplaySafeRef, right DisplaySafeRef, reason string, missing MissingInput) []productionAdapterObjectiveCloseoutWriterExecutionBridgeMismatch {
	left = normalizeOneDisplaySafeRef(left)
	right = normalizeOneDisplaySafeRef(right)
	if left != "" && right != "" && left != right {
		return []productionAdapterObjectiveCloseoutWriterExecutionBridgeMismatch{{
			failure: FailureVerificationFailed,
			reason:  reason,
			missing: missing,
		}}
	}
	return nil
}

func productionAdapterObjectiveCloseoutWriterExecutionBridgeRequiredInputs() []MissingInput {
	return []MissingInput{
		"host:objective_closeout_writer_execution_bridge_ref",
		"host:objective_closeout_writer_invocation_handoff",
		"host:objective_closeout_writer_durable_request",
		"host:objective_closeout_writer_host_adapter_run_ref",
		"host:objective_closeout_writer_durable_result",
		"host:objective_closeout_writer_invocation_result_envelope_ref",
	}
}

func productionAdapterObjectiveCloseoutWriterExecutionBridgeBoundaries(groups ...[]Boundary) []Boundary {
	return MergeBoundaries(
		[]Boundary{
			"production_adapter_objective_closeout_writer_execution_bridge",
			"objective_closeout_writer_execution_bridge_projection_only",
			"host_owned_objective_closeout_writer_adapter",
			"host_owned_writer_adapter_execution",
			"display_safe_refs_only",
			"display_safe_result_refs_only",
			"no_runner_dispatch",
			"no_adapter_invocation_by_core",
			"no_dry_run_by_core",
			"no_durable_write_by_core",
			"no_objective_store_write_by_core",
			"no_runstore_write_by_core",
			"no_compensation_execution_by_core",
		},
		MergeBoundaries(groups...),
	)
}

func productionAdapterObjectiveCloseoutWriterExecutionBridgeUnsafe(input ProductionAdapterObjectiveCloseoutWriterExecutionBridgeInput, handoff ProductionAdapterObjectiveCloseoutWriterInvocationHostUIHandoff, request ProductionAdapterObjectiveCloseoutWriterDurableRequest, result ProductionAdapterObjectiveCloseoutWriterDurableResult, resultProvided bool) bool {
	return input.RawOutputLoaded ||
		displaySafeRefRejected(input.BridgeRef) ||
		displaySafeRefRejected(input.ResultEnvelopeRef) ||
		productionAdapterObjectiveCloseoutWriterInvocationHostUIHandoffUnsafeOutput(handoff) ||
		productionAdapterObjectiveCloseoutWriterDurableRequestOutputUnsafe(request) ||
		(resultProvided && productionAdapterObjectiveCloseoutWriterDurableResultOutputUnsafe(result))
}

func productionAdapterObjectiveCloseoutWriterExecutionBridgeUnsafeOutput(input ProductionAdapterObjectiveCloseoutWriterExecutionBridge) bool {
	return displaySafeRefRejected(input.BridgeRef) ||
		displaySafeRefRejected(input.HostUIHandoffRef) ||
		displaySafeRefRejected(input.PrimaryDisplayRef) ||
		displaySafeRefRejected(input.ReviewRef) ||
		displaySafeRefRejected(input.FixtureRef) ||
		displaySafeRefRejected(input.InvocationEnvelopeRef) ||
		displaySafeRefRejected(input.ResultEnvelopeRef) ||
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
		displaySafeRefSliceRejected(input.PolicyRefs) ||
		displaySafeRefSliceRejected(input.RequiredPolicyRefs) ||
		displaySafeRefSliceRejected(input.ApprovalRefs) ||
		displaySafeRefSliceRejected(input.RequiredApprovalRefs) ||
		displaySafeRefRejected(input.BudgetRef) ||
		displaySafeRefRejected(input.RequiredBudgetRef) ||
		displaySafeRefRejected(input.IdempotencyRef) ||
		displaySafeRefRejected(input.IdempotencyContractRef) ||
		displaySafeRefRejected(input.TimeoutPolicyRef) ||
		displaySafeRefSliceRejected(input.DurableEvidenceRefs) ||
		input.RawOutputLoaded
}

func unavailableProductionAdapterObjectiveCloseoutWriterExecutionBridge() ProductionAdapterObjectiveCloseoutWriterExecutionBridge {
	return ProductionAdapterObjectiveCloseoutWriterExecutionBridge{
		ContractVersion: ContractVersion,
		Projected:       true,
		Available:       false,
		Status:          "unavailable",
		Mode:            "production_adapter_objective_closeout_writer_execution_bridge",
		RequiredInputs:  productionAdapterObjectiveCloseoutWriterExecutionBridgeRequiredInputs(),
		Boundaries: []Boundary{
			"production_adapter_objective_closeout_writer_execution_bridge",
			"objective_closeout_writer_execution_bridge_projection_only",
			"host_owned_objective_closeout_writer_adapter",
			"display_safe_refs_only",
			"no_adapter_invocation_by_core",
			"no_runner_dispatch",
			"no_durable_write_by_core",
			"no_objective_store_write_by_core",
			"no_runstore_write_by_core",
		},
		NextHostAction: "provide_objective_closeout_writer_invocation_handoff",
		RunnerEffect:   "none",
		PromptEffect:   "none",
	}
}
