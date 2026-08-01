package controlcontract

type ProductionAdapterObjectiveCloseoutWriterInvocationReviewInput struct {
	ReviewRef        DisplaySafeRef                                                   `json:"review_ref,omitempty"`
	Invocation       ProductionAdapterObjectiveCloseoutWriterInvocationEnvelope       `json:"invocation,omitempty"`
	InvocationResult ProductionAdapterObjectiveCloseoutWriterInvocationResultEnvelope `json:"invocation_result,omitempty"`
	RawOutputLoaded  bool                                                             `json:"raw_output_loaded"`
}

type ProductionAdapterObjectiveCloseoutWriterInvocationReview struct {
	ContractVersion                 string           `json:"contract_version,omitempty"`
	Projected                       bool             `json:"projected"`
	Available                       bool             `json:"available"`
	Status                          string           `json:"status,omitempty"`
	Mode                            string           `json:"mode,omitempty"`
	DisplayState                    string           `json:"display_state,omitempty"`
	DisplayStage                    string           `json:"display_stage,omitempty"`
	DisplaySections                 []string         `json:"display_sections,omitempty"`
	ReadyForHostDisplay             bool             `json:"ready_for_host_display"`
	ReadyForHostAdapterInvocation   bool             `json:"ready_for_host_adapter_invocation"`
	ReadyForDurableReadbackReview   bool             `json:"ready_for_durable_readback_review"`
	ReadyForFailureReview           bool             `json:"ready_for_failure_review"`
	ReadyForCompensationReview      bool             `json:"ready_for_compensation_review"`
	ReadyForBlockedReview           bool             `json:"ready_for_blocked_review"`
	HostMayInvokeWriterAdapter      bool             `json:"host_may_invoke_writer_adapter"`
	HostMayExecuteDurableWrite      bool             `json:"host_may_execute_durable_write"`
	HostAdapterInvocationAuthorized bool             `json:"host_adapter_invocation_authorized"`
	HostAdapterInvocationBound      bool             `json:"host_adapter_invocation_bound"`
	HostDurableWriteReported        bool             `json:"host_durable_write_reported"`
	HostDurableWriteSucceeded       bool             `json:"host_durable_write_succeeded"`
	HostDurableWriteFailed          bool             `json:"host_durable_write_failed"`
	HostDurableWriteRecorded        bool             `json:"host_durable_write_recorded"`
	CoreInvocationExecuted          bool             `json:"core_invocation_executed"`
	DryRunByCore                    bool             `json:"dry_run_by_core"`
	DurableWriteByCore              bool             `json:"durable_write_by_core"`
	ObjectiveStoreWriteByCore       bool             `json:"objective_store_write_by_core"`
	RunstoreWriteByCore             bool             `json:"runstore_write_by_core"`
	ReviewRef                       DisplaySafeRef   `json:"review_ref,omitempty"`
	InvocationEnvelopeRef           DisplaySafeRef   `json:"invocation_envelope_ref,omitempty"`
	ResultEnvelopeRef               DisplaySafeRef   `json:"result_envelope_ref,omitempty"`
	ReviewPacketRef                 DisplaySafeRef   `json:"review_packet_ref,omitempty"`
	DurableRequestRef               DisplaySafeRef   `json:"durable_request_ref,omitempty"`
	DurableResultRef                DisplaySafeRef   `json:"durable_result_ref,omitempty"`
	ExpectedDurableResultRef        DisplaySafeRef   `json:"expected_durable_result_ref,omitempty"`
	WriterInvocationRef             DisplaySafeRef   `json:"writer_invocation_ref,omitempty"`
	WriterRef                       DisplaySafeRef   `json:"writer_ref,omitempty"`
	HostWriterBindingRef            DisplaySafeRef   `json:"host_writer_binding_ref,omitempty"`
	HostAdapterVersionRef           DisplaySafeRef   `json:"host_adapter_version_ref,omitempty"`
	ExpectedHostAdapterRunRef       DisplaySafeRef   `json:"expected_host_adapter_run_ref,omitempty"`
	HostAdapterRunRef               DisplaySafeRef   `json:"host_adapter_run_ref,omitempty"`
	ExpectedReadbackRef             DisplaySafeRef   `json:"expected_readback_ref,omitempty"`
	ExpectedFailureRef              DisplaySafeRef   `json:"expected_failure_ref,omitempty"`
	ExpectedCompensationRef         DisplaySafeRef   `json:"expected_compensation_ref,omitempty"`
	AppliedDurableEventRef          DisplaySafeRef   `json:"applied_durable_event_ref,omitempty"`
	AppliedRunstoreRef              DisplaySafeRef   `json:"applied_runstore_ref,omitempty"`
	AppliedObjectiveStateRef        DisplaySafeRef   `json:"applied_objective_state_ref,omitempty"`
	FailureRef                      DisplaySafeRef   `json:"failure_ref,omitempty"`
	CompensationRef                 DisplaySafeRef   `json:"compensation_ref,omitempty"`
	HostDurableWriteConfirmationRef DisplaySafeRef   `json:"host_durable_write_confirmation_ref,omitempty"`
	CapabilityProofRefs             []DisplaySafeRef `json:"capability_proof_refs,omitempty"`
	ApprovalBindingRefs             []DisplaySafeRef `json:"approval_binding_refs,omitempty"`
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
type ProductionAdapterObjectiveCloseoutWriterInvocationReviewBlackboxFixtureInput struct {
	FixtureRef      DisplaySafeRef                                           `json:"fixture_ref,omitempty"`
	Review          ProductionAdapterObjectiveCloseoutWriterInvocationReview `json:"review,omitempty"`
	RawOutputLoaded bool                                                     `json:"raw_output_loaded"`
}

// agentx-api: internal_candidate
type ProductionAdapterObjectiveCloseoutWriterInvocationReviewBlackboxFixture struct {
	ContractVersion                 string           `json:"contract_version,omitempty"`
	Projected                       bool             `json:"projected"`
	Available                       bool             `json:"available"`
	Status                          string           `json:"status,omitempty"`
	Mode                            string           `json:"mode,omitempty"`
	DisplayState                    string           `json:"display_state,omitempty"`
	DisplayStage                    string           `json:"display_stage,omitempty"`
	DisplaySections                 []string         `json:"display_sections,omitempty"`
	ReadyForHostDisplay             bool             `json:"ready_for_host_display"`
	ReadyForHostAdapterInvocation   bool             `json:"ready_for_host_adapter_invocation"`
	ReadyForDurableReadbackReview   bool             `json:"ready_for_durable_readback_review"`
	ReadyForFailureReview           bool             `json:"ready_for_failure_review"`
	ReadyForCompensationReview      bool             `json:"ready_for_compensation_review"`
	ReadyForBlockedReview           bool             `json:"ready_for_blocked_review"`
	HostMayInvokeWriterAdapter      bool             `json:"host_may_invoke_writer_adapter"`
	HostMayExecuteDurableWrite      bool             `json:"host_may_execute_durable_write"`
	HostAdapterInvocationAuthorized bool             `json:"host_adapter_invocation_authorized"`
	HostAdapterInvocationBound      bool             `json:"host_adapter_invocation_bound"`
	HostDurableWriteReported        bool             `json:"host_durable_write_reported"`
	HostDurableWriteSucceeded       bool             `json:"host_durable_write_succeeded"`
	HostDurableWriteFailed          bool             `json:"host_durable_write_failed"`
	HostDurableWriteRecorded        bool             `json:"host_durable_write_recorded"`
	CoreInvocationExecuted          bool             `json:"core_invocation_executed"`
	DryRunByCore                    bool             `json:"dry_run_by_core"`
	DurableWriteByCore              bool             `json:"durable_write_by_core"`
	ObjectiveStoreWriteByCore       bool             `json:"objective_store_write_by_core"`
	RunstoreWriteByCore             bool             `json:"runstore_write_by_core"`
	FixtureRef                      DisplaySafeRef   `json:"fixture_ref,omitempty"`
	ReviewRef                       DisplaySafeRef   `json:"review_ref,omitempty"`
	InvocationEnvelopeRef           DisplaySafeRef   `json:"invocation_envelope_ref,omitempty"`
	ResultEnvelopeRef               DisplaySafeRef   `json:"result_envelope_ref,omitempty"`
	ReviewPacketRef                 DisplaySafeRef   `json:"review_packet_ref,omitempty"`
	DurableRequestRef               DisplaySafeRef   `json:"durable_request_ref,omitempty"`
	DurableResultRef                DisplaySafeRef   `json:"durable_result_ref,omitempty"`
	ExpectedDurableResultRef        DisplaySafeRef   `json:"expected_durable_result_ref,omitempty"`
	WriterInvocationRef             DisplaySafeRef   `json:"writer_invocation_ref,omitempty"`
	WriterRef                       DisplaySafeRef   `json:"writer_ref,omitempty"`
	HostWriterBindingRef            DisplaySafeRef   `json:"host_writer_binding_ref,omitempty"`
	HostAdapterVersionRef           DisplaySafeRef   `json:"host_adapter_version_ref,omitempty"`
	ExpectedHostAdapterRunRef       DisplaySafeRef   `json:"expected_host_adapter_run_ref,omitempty"`
	HostAdapterRunRef               DisplaySafeRef   `json:"host_adapter_run_ref,omitempty"`
	ExpectedReadbackRef             DisplaySafeRef   `json:"expected_readback_ref,omitempty"`
	ExpectedFailureRef              DisplaySafeRef   `json:"expected_failure_ref,omitempty"`
	ExpectedCompensationRef         DisplaySafeRef   `json:"expected_compensation_ref,omitempty"`
	AppliedDurableEventRef          DisplaySafeRef   `json:"applied_durable_event_ref,omitempty"`
	AppliedRunstoreRef              DisplaySafeRef   `json:"applied_runstore_ref,omitempty"`
	AppliedObjectiveStateRef        DisplaySafeRef   `json:"applied_objective_state_ref,omitempty"`
	FailureRef                      DisplaySafeRef   `json:"failure_ref,omitempty"`
	CompensationRef                 DisplaySafeRef   `json:"compensation_ref,omitempty"`
	HostDurableWriteConfirmationRef DisplaySafeRef   `json:"host_durable_write_confirmation_ref,omitempty"`
	CapabilityProofRefs             []DisplaySafeRef `json:"capability_proof_refs,omitempty"`
	ApprovalBindingRefs             []DisplaySafeRef `json:"approval_binding_refs,omitempty"`
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

func BuildProductionAdapterObjectiveCloseoutWriterInvocationReview(input ProductionAdapterObjectiveCloseoutWriterInvocationReviewInput) ProductionAdapterObjectiveCloseoutWriterInvocationReview {
	envelopeProvided := !productionAdapterObjectiveCloseoutWriterInvocationEnvelopeEmpty(input.Invocation)
	resultProvided := !productionAdapterObjectiveCloseoutWriterInvocationResultEnvelopeEmpty(input.InvocationResult)
	if !envelopeProvided && !resultProvided {
		return unavailableProductionAdapterObjectiveCloseoutWriterInvocationReview()
	}
	envelope := ProductionAdapterObjectiveCloseoutWriterInvocationEnvelope{}
	if envelopeProvided {
		envelope = input.Invocation.Normalize()
	}
	resultEnvelope := ProductionAdapterObjectiveCloseoutWriterInvocationResultEnvelope{}
	if resultProvided {
		resultEnvelope = input.InvocationResult.Normalize()
	}
	result := ProductionAdapterObjectiveCloseoutWriterInvocationReview{
		ContractVersion:                 ContractVersion,
		Projected:                       true,
		Available:                       (envelopeProvided && envelope.Available) || (resultProvided && resultEnvelope.Available),
		Status:                          "ready_for_objective_closeout_writer_invocation_blocked_review",
		Mode:                            "production_adapter_objective_closeout_writer_invocation_review",
		DisplayState:                    "blocked",
		DisplayStage:                    "blocked_review",
		DisplaySections:                 productionAdapterObjectiveCloseoutWriterInvocationReviewDisplaySections(),
		ReviewRef:                       normalizeOneDisplaySafeRef(input.ReviewRef),
		InvocationEnvelopeRef:           firstDisplaySafeRef(resultEnvelope.InvocationEnvelopeRef, envelope.InvocationEnvelopeRef),
		ResultEnvelopeRef:               resultEnvelope.ResultEnvelopeRef,
		ReviewPacketRef:                 firstDisplaySafeRef(resultEnvelope.ReviewPacketRef, envelope.ReviewPacketRef),
		DurableRequestRef:               firstDisplaySafeRef(resultEnvelope.DurableRequestRef, envelope.DurableRequestRef),
		DurableResultRef:                resultEnvelope.DurableResultRef,
		ExpectedDurableResultRef:        firstDisplaySafeRef(resultEnvelope.ExpectedDurableResultRef, envelope.ExpectedDurableResultRef),
		WriterInvocationRef:             firstDisplaySafeRef(resultEnvelope.WriterInvocationRef, envelope.WriterInvocationRef),
		WriterRef:                       firstDisplaySafeRef(resultEnvelope.WriterRef, envelope.WriterRef),
		HostWriterBindingRef:            firstDisplaySafeRef(resultEnvelope.HostWriterBindingRef, envelope.HostWriterBindingRef),
		HostAdapterVersionRef:           firstDisplaySafeRef(resultEnvelope.HostAdapterVersionRef, envelope.HostAdapterVersionRef),
		ExpectedHostAdapterRunRef:       firstDisplaySafeRef(resultEnvelope.ExpectedHostAdapterRunRef, envelope.ExpectedHostAdapterRunRef),
		HostAdapterRunRef:               resultEnvelope.HostAdapterRunRef,
		ExpectedReadbackRef:             firstDisplaySafeRef(resultEnvelope.ExpectedReadbackRef, envelope.ExpectedReadbackRef),
		ExpectedFailureRef:              firstDisplaySafeRef(resultEnvelope.ExpectedFailureRef, envelope.ExpectedFailureRef),
		ExpectedCompensationRef:         firstDisplaySafeRef(resultEnvelope.ExpectedCompensationRef, envelope.ExpectedCompensationRef),
		AppliedDurableEventRef:          resultEnvelope.AppliedDurableEventRef,
		AppliedRunstoreRef:              resultEnvelope.AppliedRunstoreRef,
		AppliedObjectiveStateRef:        resultEnvelope.AppliedObjectiveStateRef,
		FailureRef:                      resultEnvelope.FailureRef,
		CompensationRef:                 resultEnvelope.CompensationRef,
		HostDurableWriteConfirmationRef: firstDisplaySafeRef(resultEnvelope.HostDurableWriteConfirmationRef, envelope.HostDurableWriteConfirmationRef),
		CapabilityProofRefs:             productionAdapterObjectiveCloseoutWriterInvocationReviewFirstRefs(resultEnvelope.CapabilityProofRefs, envelope.CapabilityProofRefs),
		ApprovalBindingRefs:             productionAdapterObjectiveCloseoutWriterInvocationReviewFirstRefs(resultEnvelope.ApprovalBindingRefs, envelope.ApprovalBindingRefs),
		DurableEvidenceRefs:             cloneDisplaySafeRefs(resultEnvelope.DurableEvidenceRefs),
		MissingInputs:                   productionAdapterObjectiveCloseoutWriterInvocationReviewMissingInputs(envelope, envelopeProvided, resultEnvelope, resultProvided),
		BlockedReasons:                  productionAdapterObjectiveCloseoutWriterInvocationReviewBlockedReasons(envelope, envelopeProvided, resultEnvelope, resultProvided),
		FailureClass:                    productionAdapterObjectiveCloseoutWriterInvocationReviewFailureClass(envelope, envelopeProvided, resultEnvelope, resultProvided),
		Boundaries:                      productionAdapterObjectiveCloseoutWriterInvocationReviewBoundaries(envelope.Boundaries, resultEnvelope.Boundaries),
		NextHostAction:                  firstNextHostAction(resultEnvelope.NextHostAction, envelope.NextHostAction),
		RunnerEffect:                    "none",
		PromptEffect:                    "none",
		RawOutputLoaded:                 input.RawOutputLoaded || envelope.RawOutputLoaded || resultEnvelope.RawOutputLoaded,
		HostMayInvokeWriterAdapter:      envelope.HostMayInvokeWriterAdapter,
		HostMayExecuteDurableWrite:      envelope.HostMayExecuteDurableWrite,
		HostAdapterInvocationAuthorized: firstBool(resultEnvelope.HostAdapterInvocationAuthorized, envelope.HostAdapterInvocationAuthorized),
		HostAdapterInvocationBound:      resultEnvelope.HostAdapterInvocationBound,
		HostDurableWriteReported:        resultEnvelope.HostDurableWriteReported,
		HostDurableWriteSucceeded:       resultEnvelope.HostDurableWriteSucceeded,
		HostDurableWriteFailed:          resultEnvelope.HostDurableWriteFailed,
		HostDurableWriteRecorded:        resultEnvelope.HostDurableWriteRecorded,
	}
	if input.RawOutputLoaded || displaySafeRefRejected(input.ReviewRef) ||
		(envelopeProvided && productionAdapterObjectiveCloseoutWriterInvocationEnvelopeUnsafeOutput(envelope)) ||
		(resultProvided && productionAdapterObjectiveCloseoutWriterInvocationResultEnvelopeUnsafeOutput(resultEnvelope)) {
		result.RawOutputLoaded = true
		result = productionAdapterObjectiveCloseoutWriterInvocationReviewBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if result.ReviewRef == "" {
		result = productionAdapterObjectiveCloseoutWriterInvocationReviewBlock(result, FailureEvidenceMissing, "writer_invocation_review_ref_missing", "host:objective_closeout_writer_invocation_review_ref", "provide_objective_closeout_writer_invocation_review")
		return result.Normalize()
	}
	if resultProvided {
		productionAdapterObjectiveCloseoutWriterInvocationReviewApplyResult(&result, resultEnvelope)
	} else {
		productionAdapterObjectiveCloseoutWriterInvocationReviewApplyEnvelope(&result, envelope)
	}
	return result.Normalize()
}

// agentx-api: internal_candidate
func BuildProductionAdapterObjectiveCloseoutWriterInvocationReviewBlackboxFixture(input ProductionAdapterObjectiveCloseoutWriterInvocationReviewBlackboxFixtureInput) ProductionAdapterObjectiveCloseoutWriterInvocationReviewBlackboxFixture {
	if productionAdapterObjectiveCloseoutWriterInvocationReviewEmpty(input.Review) {
		return unavailableProductionAdapterObjectiveCloseoutWriterInvocationReviewBlackboxFixture()
	}
	review := input.Review.Normalize()
	result := ProductionAdapterObjectiveCloseoutWriterInvocationReviewBlackboxFixture{
		ContractVersion:                 ContractVersion,
		Projected:                       true,
		Available:                       review.Available,
		Status:                          productionAdapterObjectiveCloseoutWriterInvocationReviewFixtureStatus(review.Status),
		Mode:                            "production_adapter_objective_closeout_writer_invocation_review_blackbox_fixture",
		DisplayState:                    review.DisplayState,
		DisplayStage:                    review.DisplayStage,
		DisplaySections:                 append(productionAdapterObjectiveCloseoutWriterInvocationReviewDisplaySections(), "objective_closeout_writer_invocation_blackbox_assertions"),
		FixtureRef:                      normalizeOneDisplaySafeRef(input.FixtureRef),
		ReviewRef:                       review.ReviewRef,
		InvocationEnvelopeRef:           review.InvocationEnvelopeRef,
		ResultEnvelopeRef:               review.ResultEnvelopeRef,
		ReviewPacketRef:                 review.ReviewPacketRef,
		DurableRequestRef:               review.DurableRequestRef,
		DurableResultRef:                review.DurableResultRef,
		ExpectedDurableResultRef:        review.ExpectedDurableResultRef,
		WriterInvocationRef:             review.WriterInvocationRef,
		WriterRef:                       review.WriterRef,
		HostWriterBindingRef:            review.HostWriterBindingRef,
		HostAdapterVersionRef:           review.HostAdapterVersionRef,
		ExpectedHostAdapterRunRef:       review.ExpectedHostAdapterRunRef,
		HostAdapterRunRef:               review.HostAdapterRunRef,
		ExpectedReadbackRef:             review.ExpectedReadbackRef,
		ExpectedFailureRef:              review.ExpectedFailureRef,
		ExpectedCompensationRef:         review.ExpectedCompensationRef,
		AppliedDurableEventRef:          review.AppliedDurableEventRef,
		AppliedRunstoreRef:              review.AppliedRunstoreRef,
		AppliedObjectiveStateRef:        review.AppliedObjectiveStateRef,
		FailureRef:                      review.FailureRef,
		CompensationRef:                 review.CompensationRef,
		HostDurableWriteConfirmationRef: review.HostDurableWriteConfirmationRef,
		CapabilityProofRefs:             cloneDisplaySafeRefs(review.CapabilityProofRefs),
		ApprovalBindingRefs:             cloneDisplaySafeRefs(review.ApprovalBindingRefs),
		DurableEvidenceRefs:             cloneDisplaySafeRefs(review.DurableEvidenceRefs),
		MissingInputs:                   cloneMissingInputs(review.MissingInputs),
		BlockedReasons:                  cloneStringSlice(review.BlockedReasons),
		FailureClass:                    review.FailureClass,
		Boundaries:                      productionAdapterObjectiveCloseoutWriterInvocationReviewFixtureBoundaries(review.Boundaries),
		NextHostAction:                  review.NextHostAction,
		RunnerEffect:                    "none",
		PromptEffect:                    "none",
		RawOutputLoaded:                 input.RawOutputLoaded || review.RawOutputLoaded,
		ReadyForHostDisplay:             review.ReadyForHostDisplay,
		ReadyForHostAdapterInvocation:   review.ReadyForHostAdapterInvocation,
		ReadyForDurableReadbackReview:   review.ReadyForDurableReadbackReview,
		ReadyForFailureReview:           review.ReadyForFailureReview,
		ReadyForCompensationReview:      review.ReadyForCompensationReview,
		ReadyForBlockedReview:           review.ReadyForBlockedReview,
		HostMayInvokeWriterAdapter:      review.HostMayInvokeWriterAdapter,
		HostMayExecuteDurableWrite:      review.HostMayExecuteDurableWrite,
		HostAdapterInvocationAuthorized: review.HostAdapterInvocationAuthorized,
		HostAdapterInvocationBound:      review.HostAdapterInvocationBound,
		HostDurableWriteReported:        review.HostDurableWriteReported,
		HostDurableWriteSucceeded:       review.HostDurableWriteSucceeded,
		HostDurableWriteFailed:          review.HostDurableWriteFailed,
		HostDurableWriteRecorded:        review.HostDurableWriteRecorded,
	}
	if input.RawOutputLoaded || displaySafeRefRejected(input.FixtureRef) || productionAdapterObjectiveCloseoutWriterInvocationReviewUnsafeOutput(review) {
		result.RawOutputLoaded = true
		result = productionAdapterObjectiveCloseoutWriterInvocationReviewFixtureBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if result.FixtureRef == "" {
		result = productionAdapterObjectiveCloseoutWriterInvocationReviewFixtureBlock(result, FailureEvidenceMissing, "writer_invocation_review_fixture_ref_missing", "host:objective_closeout_writer_invocation_review_fixture_ref", "provide_objective_closeout_writer_invocation_review_fixture")
	}
	return result.Normalize()
}

func CloneProductionAdapterObjectiveCloseoutWriterInvocationReview(in ProductionAdapterObjectiveCloseoutWriterInvocationReview) ProductionAdapterObjectiveCloseoutWriterInvocationReview {
	out := in
	out.DisplaySections = cloneStringSlice(in.DisplaySections)
	out.CapabilityProofRefs = cloneDisplaySafeRefs(in.CapabilityProofRefs)
	out.ApprovalBindingRefs = cloneDisplaySafeRefs(in.ApprovalBindingRefs)
	out.DurableEvidenceRefs = cloneDisplaySafeRefs(in.DurableEvidenceRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (p ProductionAdapterObjectiveCloseoutWriterInvocationReview) Clone() ProductionAdapterObjectiveCloseoutWriterInvocationReview {
	return CloneProductionAdapterObjectiveCloseoutWriterInvocationReview(p)
}

func (p ProductionAdapterObjectiveCloseoutWriterInvocationReview) Normalize() ProductionAdapterObjectiveCloseoutWriterInvocationReview {
	out := CloneProductionAdapterObjectiveCloseoutWriterInvocationReview(p)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = normalizeControlToken(out.Status)
	out.Mode = normalizeControlToken(out.Mode)
	if out.Mode == "" {
		out.Mode = "production_adapter_objective_closeout_writer_invocation_review"
	}
	out.DisplayState = normalizeControlToken(out.DisplayState)
	out.DisplayStage = normalizeControlToken(out.DisplayStage)
	out.DisplaySections = normalizeControlTokenList(out.DisplaySections)
	if len(out.DisplaySections) == 0 {
		out.DisplaySections = productionAdapterObjectiveCloseoutWriterInvocationReviewDisplaySections()
	}
	productionAdapterObjectiveCloseoutWriterInvocationReviewNormalizeRefs(&out)
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
	if out.Status == "" {
		out.Status = "ready_for_objective_closeout_writer_invocation_blocked_review"
	}
	if out.RawOutputLoaded || productionAdapterObjectiveCloseoutWriterInvocationReviewUnsafeOutput(out) {
		out.RawOutputLoaded = true
		out = productionAdapterObjectiveCloseoutWriterInvocationReviewBlock(out, firstFailureClass(out.FailureClass, FailureEvidenceWeak), "unsafe_input_ref", "host:display_safe_refs", firstNextHostAction(out.NextHostAction, "provide_display_safe_refs"))
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
	}
	if !out.Available {
		out.Status = "unavailable"
		out.DisplayState = "unavailable"
		out.DisplayStage = "unavailable"
		out.ReadyForHostDisplay = false
		out.ReadyForHostAdapterInvocation = false
		out.ReadyForDurableReadbackReview = false
		out.ReadyForFailureReview = false
		out.ReadyForCompensationReview = false
		out.ReadyForBlockedReview = false
		out.HostMayInvokeWriterAdapter = false
		out.HostMayExecuteDurableWrite = false
		out.HostAdapterInvocationBound = false
		return out
	}
	out.ReadyForHostDisplay = out.ReadyForHostDisplay &&
		out.ReviewRef != "" &&
		out.InvocationEnvelopeRef != "" &&
		out.WriterRef != "" &&
		!out.RawOutputLoaded
	invocationReady := out.Status == "ready_for_objective_closeout_writer_invocation_review" &&
		out.ReadyForHostDisplay &&
		out.DisplayState == "host_adapter_invocation_ready" &&
		out.InvocationEnvelopeRef != "" &&
		out.WriterInvocationRef != "" &&
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
		len(out.BlockedReasons) == 0
	readbackReady := out.Status == "ready_for_objective_closeout_writer_invocation_result_readback_review" &&
		out.ReadyForHostDisplay &&
		out.ResultEnvelopeRef != "" &&
		out.DisplayState == "invocation_result_readback_bound" &&
		out.HostDurableWriteReported &&
		out.HostDurableWriteSucceeded &&
		out.HostDurableWriteRecorded &&
		!out.HostDurableWriteFailed &&
		out.HostAdapterRunRef == out.ExpectedHostAdapterRunRef &&
		out.DurableResultRef == out.ExpectedDurableResultRef &&
		out.AppliedDurableEventRef != "" &&
		out.AppliedRunstoreRef != "" &&
		out.AppliedObjectiveStateRef != "" &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0
	failureReady := out.Status == "ready_for_objective_closeout_writer_invocation_result_failure_review" &&
		out.ReadyForHostDisplay &&
		out.ResultEnvelopeRef != "" &&
		out.DisplayState == "invocation_result_failed" &&
		out.HostDurableWriteReported &&
		out.HostDurableWriteFailed &&
		out.HostDurableWriteRecorded &&
		!out.HostDurableWriteSucceeded &&
		out.HostAdapterRunRef == out.ExpectedHostAdapterRunRef &&
		out.DurableResultRef == out.ExpectedDurableResultRef &&
		out.FailureRef != "" &&
		out.CompensationRef != "" &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0
	blockedReady := out.Status == "ready_for_objective_closeout_writer_invocation_blocked_review" &&
		out.ReadyForHostDisplay &&
		(len(out.MissingInputs) > 0 || len(out.BlockedReasons) > 0)
	out.ReadyForHostAdapterInvocation = out.ReadyForHostAdapterInvocation && invocationReady
	out.ReadyForDurableReadbackReview = out.ReadyForDurableReadbackReview && readbackReady
	out.ReadyForFailureReview = out.ReadyForFailureReview && failureReady
	out.ReadyForCompensationReview = out.ReadyForCompensationReview && failureReady && out.CompensationRef != ""
	out.ReadyForBlockedReview = out.ReadyForBlockedReview && blockedReady
	out.HostMayInvokeWriterAdapter = out.HostMayInvokeWriterAdapter && invocationReady
	out.HostMayExecuteDurableWrite = out.HostMayExecuteDurableWrite && invocationReady
	out.HostAdapterInvocationAuthorized = out.HostAdapterInvocationAuthorized && (invocationReady || readbackReady || failureReady || blockedReady)
	out.HostAdapterInvocationBound = out.HostAdapterInvocationBound && (readbackReady || failureReady)
	return out
}

func CloneProductionAdapterObjectiveCloseoutWriterInvocationReviewBlackboxFixture(in ProductionAdapterObjectiveCloseoutWriterInvocationReviewBlackboxFixture) ProductionAdapterObjectiveCloseoutWriterInvocationReviewBlackboxFixture {
	out := in
	out.DisplaySections = cloneStringSlice(in.DisplaySections)
	out.CapabilityProofRefs = cloneDisplaySafeRefs(in.CapabilityProofRefs)
	out.ApprovalBindingRefs = cloneDisplaySafeRefs(in.ApprovalBindingRefs)
	out.DurableEvidenceRefs = cloneDisplaySafeRefs(in.DurableEvidenceRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (p ProductionAdapterObjectiveCloseoutWriterInvocationReviewBlackboxFixture) Clone() ProductionAdapterObjectiveCloseoutWriterInvocationReviewBlackboxFixture {
	return CloneProductionAdapterObjectiveCloseoutWriterInvocationReviewBlackboxFixture(p)
}

func (p ProductionAdapterObjectiveCloseoutWriterInvocationReviewBlackboxFixture) Normalize() ProductionAdapterObjectiveCloseoutWriterInvocationReviewBlackboxFixture {
	out := CloneProductionAdapterObjectiveCloseoutWriterInvocationReviewBlackboxFixture(p)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = normalizeControlToken(out.Status)
	out.Mode = normalizeControlToken(out.Mode)
	if out.Mode == "" {
		out.Mode = "production_adapter_objective_closeout_writer_invocation_review_blackbox_fixture"
	}
	out.DisplayState = normalizeControlToken(out.DisplayState)
	out.DisplayStage = normalizeControlToken(out.DisplayStage)
	out.DisplaySections = normalizeControlTokenList(out.DisplaySections)
	if len(out.DisplaySections) == 0 {
		out.DisplaySections = append(productionAdapterObjectiveCloseoutWriterInvocationReviewDisplaySections(), "objective_closeout_writer_invocation_blackbox_assertions")
	}
	productionAdapterObjectiveCloseoutWriterInvocationReviewNormalizeFixtureRefs(&out)
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
	if out.Status == "" {
		out.Status = "ready_for_objective_closeout_writer_invocation_blocked_display"
	}
	if out.RawOutputLoaded || productionAdapterObjectiveCloseoutWriterInvocationReviewFixtureUnsafeOutput(out) {
		out.RawOutputLoaded = true
		out = productionAdapterObjectiveCloseoutWriterInvocationReviewFixtureBlock(out, firstFailureClass(out.FailureClass, FailureEvidenceWeak), "unsafe_input_ref", "host:display_safe_refs", firstNextHostAction(out.NextHostAction, "provide_display_safe_refs"))
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
	}
	if !out.Available {
		out.Status = "unavailable"
		out.DisplayState = "unavailable"
		out.DisplayStage = "unavailable"
		out.ReadyForHostDisplay = false
		out.ReadyForHostAdapterInvocation = false
		out.ReadyForDurableReadbackReview = false
		out.ReadyForFailureReview = false
		out.ReadyForCompensationReview = false
		out.ReadyForBlockedReview = false
		out.HostMayInvokeWriterAdapter = false
		out.HostMayExecuteDurableWrite = false
		out.HostAdapterInvocationBound = false
		return out
	}
	out.ReadyForHostDisplay = out.ReadyForHostDisplay &&
		out.FixtureRef != "" &&
		out.ReviewRef != "" &&
		out.InvocationEnvelopeRef != "" &&
		out.WriterRef != "" &&
		!out.RawOutputLoaded
	invocationReady := out.Status == "ready_for_objective_closeout_writer_invocation_display" &&
		out.ReadyForHostDisplay &&
		out.DisplayState == "host_adapter_invocation_ready" &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0
	readbackReady := out.Status == "ready_for_objective_closeout_writer_invocation_result_readback_display" &&
		out.ReadyForHostDisplay &&
		out.DisplayState == "invocation_result_readback_bound" &&
		out.HostDurableWriteSucceeded &&
		out.HostAdapterInvocationBound &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0
	failureReady := out.Status == "ready_for_objective_closeout_writer_invocation_result_failure_display" &&
		out.ReadyForHostDisplay &&
		out.DisplayState == "invocation_result_failed" &&
		out.HostDurableWriteFailed &&
		out.HostAdapterInvocationBound &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0
	blockedReady := out.Status == "ready_for_objective_closeout_writer_invocation_blocked_display" &&
		out.ReadyForHostDisplay &&
		(len(out.MissingInputs) > 0 || len(out.BlockedReasons) > 0)
	out.ReadyForHostAdapterInvocation = out.ReadyForHostAdapterInvocation && invocationReady
	out.ReadyForDurableReadbackReview = out.ReadyForDurableReadbackReview && readbackReady
	out.ReadyForFailureReview = out.ReadyForFailureReview && failureReady
	out.ReadyForCompensationReview = out.ReadyForCompensationReview && failureReady && out.CompensationRef != ""
	out.ReadyForBlockedReview = out.ReadyForBlockedReview && blockedReady
	out.HostMayInvokeWriterAdapter = out.HostMayInvokeWriterAdapter && invocationReady
	out.HostMayExecuteDurableWrite = out.HostMayExecuteDurableWrite && invocationReady
	out.HostAdapterInvocationAuthorized = out.HostAdapterInvocationAuthorized && (invocationReady || readbackReady || failureReady || blockedReady)
	out.HostAdapterInvocationBound = out.HostAdapterInvocationBound && (readbackReady || failureReady)
	return out
}

func productionAdapterObjectiveCloseoutWriterInvocationReviewApplyEnvelope(result *ProductionAdapterObjectiveCloseoutWriterInvocationReview, envelope ProductionAdapterObjectiveCloseoutWriterInvocationEnvelope) {
	if envelope.Status == "ready_for_objective_closeout_writer_host_adapter_invocation" &&
		envelope.ReadyForHostAdapterInvocation &&
		envelope.HostAdapterInvocationAuthorized &&
		len(envelope.MissingInputs) == 0 &&
		len(envelope.BlockedReasons) == 0 {
		result.Status = "ready_for_objective_closeout_writer_invocation_review"
		result.DisplayState = "host_adapter_invocation_ready"
		result.DisplayStage = "invocation_review"
		result.ReadyForHostDisplay = true
		result.ReadyForHostAdapterInvocation = true
		result.HostMayInvokeWriterAdapter = true
		result.HostMayExecuteDurableWrite = true
		result.HostAdapterInvocationAuthorized = true
		result.NextHostAction = "host_may_invoke_objective_closeout_writer_adapter"
		result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_objective_closeout_writer_invocation_review", "host_adapter_invocation_ready")
		return
	}
	*result = productionAdapterObjectiveCloseoutWriterInvocationReviewBlock(*result, firstFailureClass(envelope.FailureClass, FailureAuthorizationMissing), productionAdapterObjectiveCloseoutWriterInvocationReviewFirstBlockedReason(envelope.BlockedReasons, "durable_writer_review_not_ready"), "host:objective_closeout_writer_invocation_envelope", firstNextHostAction(envelope.NextHostAction, "review_objective_closeout_writer_invocation_envelope"))
}

func productionAdapterObjectiveCloseoutWriterInvocationReviewApplyResult(result *ProductionAdapterObjectiveCloseoutWriterInvocationReview, envelope ProductionAdapterObjectiveCloseoutWriterInvocationResultEnvelope) {
	switch envelope.Status {
	case "ready_for_objective_closeout_writer_invocation_result_readback_review":
		result.Status = envelope.Status
		result.DisplayState = "invocation_result_readback_bound"
		result.DisplayStage = "result_readback_review"
		result.ReadyForHostDisplay = true
		result.ReadyForDurableReadbackReview = true
		result.HostMayInvokeWriterAdapter = false
		result.HostMayExecuteDurableWrite = false
		result.HostAdapterInvocationBound = true
		result.NextHostAction = "bind_objective_closeout_writer_durable_readback"
		result.Boundaries = AppendBoundaries(result.Boundaries, "objective_closeout_writer_invocation_result_readback_display", "invocation_result_readback_bound")
	case "ready_for_objective_closeout_writer_invocation_result_failure_review":
		result.Status = envelope.Status
		result.DisplayState = "invocation_result_failed"
		result.DisplayStage = "failure_review"
		result.ReadyForHostDisplay = true
		result.ReadyForFailureReview = true
		result.ReadyForCompensationReview = envelope.CompensationRef != ""
		result.HostMayInvokeWriterAdapter = false
		result.HostMayExecuteDurableWrite = false
		result.HostAdapterInvocationBound = true
		result.NextHostAction = "review_objective_closeout_writer_durable_failure"
		result.Boundaries = AppendBoundaries(result.Boundaries, "objective_closeout_writer_invocation_result_failure_display", "compensation_not_executed")
	default:
		*result = productionAdapterObjectiveCloseoutWriterInvocationReviewBlock(*result, firstFailureClass(envelope.FailureClass, FailureVerificationFailed), productionAdapterObjectiveCloseoutWriterInvocationReviewFirstBlockedReason(envelope.BlockedReasons, "writer_invocation_result_not_ready"), "host:objective_closeout_writer_invocation_result_envelope", firstNextHostAction(envelope.NextHostAction, "review_objective_closeout_writer_invocation_result"))
	}
}

func productionAdapterObjectiveCloseoutWriterInvocationReviewBlock(result ProductionAdapterObjectiveCloseoutWriterInvocationReview, failure FailureClass, reason string, missing MissingInput, next NextHostAction) ProductionAdapterObjectiveCloseoutWriterInvocationReview {
	result.Status = "ready_for_objective_closeout_writer_invocation_blocked_review"
	result.DisplayState = productionAdapterObjectiveCloseoutWriterInvocationReviewBlockedDisplayState(reason)
	result.DisplayStage = "blocked_review"
	result.ReadyForHostAdapterInvocation = false
	result.ReadyForDurableReadbackReview = false
	result.ReadyForFailureReview = false
	result.ReadyForCompensationReview = false
	result.ReadyForBlockedReview = true
	result.HostMayInvokeWriterAdapter = false
	result.HostMayExecuteDurableWrite = false
	result.HostAdapterInvocationBound = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	result.Boundaries = AppendBoundaries(result.Boundaries, "objective_closeout_writer_invocation_blocked_review")
	if result.ReviewRef != "" && result.InvocationEnvelopeRef != "" && result.WriterRef != "" && !result.RawOutputLoaded {
		result.ReadyForHostDisplay = true
	}
	return result
}

func productionAdapterObjectiveCloseoutWriterInvocationReviewFixtureBlock(result ProductionAdapterObjectiveCloseoutWriterInvocationReviewBlackboxFixture, failure FailureClass, reason string, missing MissingInput, next NextHostAction) ProductionAdapterObjectiveCloseoutWriterInvocationReviewBlackboxFixture {
	result.Status = "ready_for_objective_closeout_writer_invocation_blocked_display"
	result.DisplayState = productionAdapterObjectiveCloseoutWriterInvocationReviewBlockedDisplayState(reason)
	result.DisplayStage = "blocked_review"
	result.ReadyForHostAdapterInvocation = false
	result.ReadyForDurableReadbackReview = false
	result.ReadyForFailureReview = false
	result.ReadyForCompensationReview = false
	result.ReadyForBlockedReview = true
	result.HostMayInvokeWriterAdapter = false
	result.HostMayExecuteDurableWrite = false
	result.HostAdapterInvocationBound = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	result.Boundaries = AppendBoundaries(result.Boundaries, "objective_closeout_writer_invocation_blocked_display")
	if result.FixtureRef != "" && result.ReviewRef != "" && result.InvocationEnvelopeRef != "" && result.WriterRef != "" && !result.RawOutputLoaded {
		result.ReadyForHostDisplay = true
	}
	return result
}

func productionAdapterObjectiveCloseoutWriterInvocationReviewNormalizeRefs(out *ProductionAdapterObjectiveCloseoutWriterInvocationReview) {
	out.ReviewRef = normalizeOneDisplaySafeRef(out.ReviewRef)
	out.InvocationEnvelopeRef = normalizeOneDisplaySafeRef(out.InvocationEnvelopeRef)
	out.ResultEnvelopeRef = normalizeOneDisplaySafeRef(out.ResultEnvelopeRef)
	out.ReviewPacketRef = normalizeOneDisplaySafeRef(out.ReviewPacketRef)
	out.DurableRequestRef = normalizeOneDisplaySafeRef(out.DurableRequestRef)
	out.DurableResultRef = normalizeOneDisplaySafeRef(out.DurableResultRef)
	out.ExpectedDurableResultRef = normalizeOneDisplaySafeRef(out.ExpectedDurableResultRef)
	out.WriterInvocationRef = normalizeOneDisplaySafeRef(out.WriterInvocationRef)
	out.WriterRef = normalizeOneDisplaySafeRef(out.WriterRef)
	out.HostWriterBindingRef = normalizeOneDisplaySafeRef(out.HostWriterBindingRef)
	out.HostAdapterVersionRef = normalizeOneDisplaySafeRef(out.HostAdapterVersionRef)
	out.ExpectedHostAdapterRunRef = normalizeOneDisplaySafeRef(out.ExpectedHostAdapterRunRef)
	out.HostAdapterRunRef = normalizeOneDisplaySafeRef(out.HostAdapterRunRef)
	out.ExpectedReadbackRef = normalizeOneDisplaySafeRef(out.ExpectedReadbackRef)
	out.ExpectedFailureRef = normalizeOneDisplaySafeRef(out.ExpectedFailureRef)
	out.ExpectedCompensationRef = normalizeOneDisplaySafeRef(out.ExpectedCompensationRef)
	out.AppliedDurableEventRef = normalizeOneDisplaySafeRef(out.AppliedDurableEventRef)
	out.AppliedRunstoreRef = normalizeOneDisplaySafeRef(out.AppliedRunstoreRef)
	out.AppliedObjectiveStateRef = normalizeOneDisplaySafeRef(out.AppliedObjectiveStateRef)
	out.FailureRef = normalizeOneDisplaySafeRef(out.FailureRef)
	out.CompensationRef = normalizeOneDisplaySafeRef(out.CompensationRef)
	out.HostDurableWriteConfirmationRef = normalizeOneDisplaySafeRef(out.HostDurableWriteConfirmationRef)
	out.CapabilityProofRefs = normalizeDisplaySafeRefs(out.CapabilityProofRefs)
	out.ApprovalBindingRefs = normalizeDisplaySafeRefs(out.ApprovalBindingRefs)
	out.DurableEvidenceRefs = normalizeDisplaySafeRefs(out.DurableEvidenceRefs)
}

func productionAdapterObjectiveCloseoutWriterInvocationReviewNormalizeFixtureRefs(out *ProductionAdapterObjectiveCloseoutWriterInvocationReviewBlackboxFixture) {
	out.FixtureRef = normalizeOneDisplaySafeRef(out.FixtureRef)
	out.ReviewRef = normalizeOneDisplaySafeRef(out.ReviewRef)
	out.InvocationEnvelopeRef = normalizeOneDisplaySafeRef(out.InvocationEnvelopeRef)
	out.ResultEnvelopeRef = normalizeOneDisplaySafeRef(out.ResultEnvelopeRef)
	out.ReviewPacketRef = normalizeOneDisplaySafeRef(out.ReviewPacketRef)
	out.DurableRequestRef = normalizeOneDisplaySafeRef(out.DurableRequestRef)
	out.DurableResultRef = normalizeOneDisplaySafeRef(out.DurableResultRef)
	out.ExpectedDurableResultRef = normalizeOneDisplaySafeRef(out.ExpectedDurableResultRef)
	out.WriterInvocationRef = normalizeOneDisplaySafeRef(out.WriterInvocationRef)
	out.WriterRef = normalizeOneDisplaySafeRef(out.WriterRef)
	out.HostWriterBindingRef = normalizeOneDisplaySafeRef(out.HostWriterBindingRef)
	out.HostAdapterVersionRef = normalizeOneDisplaySafeRef(out.HostAdapterVersionRef)
	out.ExpectedHostAdapterRunRef = normalizeOneDisplaySafeRef(out.ExpectedHostAdapterRunRef)
	out.HostAdapterRunRef = normalizeOneDisplaySafeRef(out.HostAdapterRunRef)
	out.ExpectedReadbackRef = normalizeOneDisplaySafeRef(out.ExpectedReadbackRef)
	out.ExpectedFailureRef = normalizeOneDisplaySafeRef(out.ExpectedFailureRef)
	out.ExpectedCompensationRef = normalizeOneDisplaySafeRef(out.ExpectedCompensationRef)
	out.AppliedDurableEventRef = normalizeOneDisplaySafeRef(out.AppliedDurableEventRef)
	out.AppliedRunstoreRef = normalizeOneDisplaySafeRef(out.AppliedRunstoreRef)
	out.AppliedObjectiveStateRef = normalizeOneDisplaySafeRef(out.AppliedObjectiveStateRef)
	out.FailureRef = normalizeOneDisplaySafeRef(out.FailureRef)
	out.CompensationRef = normalizeOneDisplaySafeRef(out.CompensationRef)
	out.HostDurableWriteConfirmationRef = normalizeOneDisplaySafeRef(out.HostDurableWriteConfirmationRef)
	out.CapabilityProofRefs = normalizeDisplaySafeRefs(out.CapabilityProofRefs)
	out.ApprovalBindingRefs = normalizeDisplaySafeRefs(out.ApprovalBindingRefs)
	out.DurableEvidenceRefs = normalizeDisplaySafeRefs(out.DurableEvidenceRefs)
}

func productionAdapterObjectiveCloseoutWriterInvocationReviewMissingInputs(envelope ProductionAdapterObjectiveCloseoutWriterInvocationEnvelope, envelopeProvided bool, result ProductionAdapterObjectiveCloseoutWriterInvocationResultEnvelope, resultProvided bool) []MissingInput {
	var out []MissingInput
	if !envelopeProvided && !resultProvided {
		out = AppendMissingInputs(out, "host:objective_closeout_writer_invocation_envelope")
	}
	if envelopeProvided {
		out = append(out, cloneMissingInputs(envelope.MissingInputs)...)
	}
	if resultProvided {
		out = append(out, cloneMissingInputs(result.MissingInputs)...)
	}
	return normalizeMissingInputs(out)
}

func productionAdapterObjectiveCloseoutWriterInvocationReviewBlockedReasons(envelope ProductionAdapterObjectiveCloseoutWriterInvocationEnvelope, envelopeProvided bool, result ProductionAdapterObjectiveCloseoutWriterInvocationResultEnvelope, resultProvided bool) []string {
	var out []string
	if !envelopeProvided && !resultProvided {
		out = appendUniqueControlToken(out, "missing_inputs")
	}
	if envelopeProvided {
		out = append(out, cloneStringSlice(envelope.BlockedReasons)...)
	}
	if resultProvided {
		out = append(out, cloneStringSlice(result.BlockedReasons)...)
	}
	return normalizeControlTokenList(out)
}

func productionAdapterObjectiveCloseoutWriterInvocationReviewFailureClass(envelope ProductionAdapterObjectiveCloseoutWriterInvocationEnvelope, envelopeProvided bool, result ProductionAdapterObjectiveCloseoutWriterInvocationResultEnvelope, resultProvided bool) FailureClass {
	if resultProvided {
		return firstFailureClass(result.FailureClass, envelope.FailureClass, FailureNone)
	}
	if envelopeProvided {
		return firstFailureClass(envelope.FailureClass, FailureNone)
	}
	return FailureEvidenceMissing
}

func productionAdapterObjectiveCloseoutWriterInvocationReviewFirstBlockedReason(values []string, fallback string) string {
	for _, value := range values {
		if normalized := normalizeControlToken(value); normalized != "" {
			return normalized
		}
	}
	return normalizeControlToken(fallback)
}

func productionAdapterObjectiveCloseoutWriterInvocationReviewBlockedDisplayState(reason string) string {
	switch normalizeControlToken(reason) {
	case "unsafe_input_ref":
		return "blocked_unsafe_refs"
	case "writer_invocation_result_expected_readback_ref_mismatch":
		return "blocked_result_readback_mismatch"
	case "writer_invocation_result_adapter_run_ref_mismatch":
		return "blocked_result_adapter_run_mismatch"
	case "durable_writer_review_not_ready":
		return "blocked_durable_review_not_ready"
	case "missing_inputs":
		return "blocked_missing_inputs"
	case "":
		return "blocked"
	default:
		return "blocked"
	}
}

func productionAdapterObjectiveCloseoutWriterInvocationReviewFixtureStatus(status string) string {
	switch normalizeControlToken(status) {
	case "ready_for_objective_closeout_writer_invocation_review":
		return "ready_for_objective_closeout_writer_invocation_display"
	case "ready_for_objective_closeout_writer_invocation_result_readback_review":
		return "ready_for_objective_closeout_writer_invocation_result_readback_display"
	case "ready_for_objective_closeout_writer_invocation_result_failure_review":
		return "ready_for_objective_closeout_writer_invocation_result_failure_display"
	case "ready_for_objective_closeout_writer_invocation_blocked_review":
		return "ready_for_objective_closeout_writer_invocation_blocked_display"
	case "unavailable":
		return "unavailable"
	default:
		return "ready_for_objective_closeout_writer_invocation_blocked_display"
	}
}

func productionAdapterObjectiveCloseoutWriterInvocationReviewFirstRefs(values []DisplaySafeRef, fallback []DisplaySafeRef) []DisplaySafeRef {
	normalized := normalizeDisplaySafeRefs(values)
	if len(normalized) > 0 {
		return normalized
	}
	return normalizeDisplaySafeRefs(fallback)
}

func productionAdapterObjectiveCloseoutWriterInvocationReviewUnsafeOutput(input ProductionAdapterObjectiveCloseoutWriterInvocationReview) bool {
	return displaySafeRefRejected(input.ReviewRef) ||
		displaySafeRefRejected(input.InvocationEnvelopeRef) ||
		displaySafeRefRejected(input.ResultEnvelopeRef) ||
		displaySafeRefRejected(input.ReviewPacketRef) ||
		displaySafeRefRejected(input.DurableRequestRef) ||
		displaySafeRefRejected(input.DurableResultRef) ||
		displaySafeRefRejected(input.ExpectedDurableResultRef) ||
		displaySafeRefRejected(input.WriterInvocationRef) ||
		displaySafeRefRejected(input.WriterRef) ||
		displaySafeRefRejected(input.HostWriterBindingRef) ||
		displaySafeRefRejected(input.HostAdapterVersionRef) ||
		displaySafeRefRejected(input.ExpectedHostAdapterRunRef) ||
		displaySafeRefRejected(input.HostAdapterRunRef) ||
		displaySafeRefRejected(input.ExpectedReadbackRef) ||
		displaySafeRefRejected(input.ExpectedFailureRef) ||
		displaySafeRefRejected(input.ExpectedCompensationRef) ||
		displaySafeRefRejected(input.AppliedDurableEventRef) ||
		displaySafeRefRejected(input.AppliedRunstoreRef) ||
		displaySafeRefRejected(input.AppliedObjectiveStateRef) ||
		displaySafeRefRejected(input.FailureRef) ||
		displaySafeRefRejected(input.CompensationRef) ||
		displaySafeRefRejected(input.HostDurableWriteConfirmationRef) ||
		displaySafeRefSliceRejected(input.CapabilityProofRefs) ||
		displaySafeRefSliceRejected(input.ApprovalBindingRefs) ||
		displaySafeRefSliceRejected(input.DurableEvidenceRefs) ||
		input.RawOutputLoaded
}

func productionAdapterObjectiveCloseoutWriterInvocationReviewFixtureUnsafeOutput(input ProductionAdapterObjectiveCloseoutWriterInvocationReviewBlackboxFixture) bool {
	return displaySafeRefRejected(input.FixtureRef) ||
		displaySafeRefRejected(input.ReviewRef) ||
		displaySafeRefRejected(input.InvocationEnvelopeRef) ||
		displaySafeRefRejected(input.ResultEnvelopeRef) ||
		displaySafeRefRejected(input.ReviewPacketRef) ||
		displaySafeRefRejected(input.DurableRequestRef) ||
		displaySafeRefRejected(input.DurableResultRef) ||
		displaySafeRefRejected(input.ExpectedDurableResultRef) ||
		displaySafeRefRejected(input.WriterInvocationRef) ||
		displaySafeRefRejected(input.WriterRef) ||
		displaySafeRefRejected(input.HostWriterBindingRef) ||
		displaySafeRefRejected(input.HostAdapterVersionRef) ||
		displaySafeRefRejected(input.ExpectedHostAdapterRunRef) ||
		displaySafeRefRejected(input.HostAdapterRunRef) ||
		displaySafeRefRejected(input.ExpectedReadbackRef) ||
		displaySafeRefRejected(input.ExpectedFailureRef) ||
		displaySafeRefRejected(input.ExpectedCompensationRef) ||
		displaySafeRefRejected(input.AppliedDurableEventRef) ||
		displaySafeRefRejected(input.AppliedRunstoreRef) ||
		displaySafeRefRejected(input.AppliedObjectiveStateRef) ||
		displaySafeRefRejected(input.FailureRef) ||
		displaySafeRefRejected(input.CompensationRef) ||
		displaySafeRefRejected(input.HostDurableWriteConfirmationRef) ||
		displaySafeRefSliceRejected(input.CapabilityProofRefs) ||
		displaySafeRefSliceRejected(input.ApprovalBindingRefs) ||
		displaySafeRefSliceRejected(input.DurableEvidenceRefs) ||
		input.RawOutputLoaded
}

func productionAdapterObjectiveCloseoutWriterInvocationResultEnvelopeEmpty(envelope ProductionAdapterObjectiveCloseoutWriterInvocationResultEnvelope) bool {
	return !envelope.Projected &&
		!envelope.Available &&
		envelope.Status == "" &&
		envelope.Mode == "" &&
		!envelope.ReadyForHostDisplay &&
		!envelope.ReadyForDurableReadbackReview &&
		!envelope.ReadyForFailureReview &&
		!envelope.ReadyForCompensationReview &&
		!envelope.HostAdapterInvocationBound &&
		envelope.ResultEnvelopeRef == "" &&
		envelope.InvocationEnvelopeRef == "" &&
		envelope.WriterRef == "" &&
		len(envelope.MissingInputs) == 0 &&
		len(envelope.BlockedReasons) == 0 &&
		len(envelope.Boundaries) == 0 &&
		envelope.NextHostAction == "" &&
		!envelope.RawOutputLoaded
}

func productionAdapterObjectiveCloseoutWriterInvocationReviewEmpty(review ProductionAdapterObjectiveCloseoutWriterInvocationReview) bool {
	return !review.Projected &&
		!review.Available &&
		review.Status == "" &&
		review.Mode == "" &&
		review.ReviewRef == "" &&
		review.InvocationEnvelopeRef == "" &&
		review.WriterRef == "" &&
		len(review.MissingInputs) == 0 &&
		len(review.BlockedReasons) == 0 &&
		len(review.Boundaries) == 0 &&
		review.NextHostAction == "" &&
		!review.RawOutputLoaded
}

func productionAdapterObjectiveCloseoutWriterInvocationReviewDisplaySections() []string {
	return []string{
		"writer_invocation_summary",
		"writer_adapter_authorization",
		"writer_adapter_result",
		"writer_result_readback",
		"writer_failure_compensation",
		"writer_blocked_reasons",
	}
}

func productionAdapterObjectiveCloseoutWriterInvocationReviewBoundaries(groups ...[]Boundary) []Boundary {
	return MergeBoundaries(
		[]Boundary{
			"production_adapter_objective_closeout_writer_invocation_review",
			"objective_closeout_writer_invocation_review_projection_only",
			"host_owned_objective_closeout_writer_adapter",
			"host_cli_objective_closeout_writer_invocation_display",
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

func productionAdapterObjectiveCloseoutWriterInvocationReviewFixtureBoundaries(groups ...[]Boundary) []Boundary {
	return MergeBoundaries(
		[]Boundary{
			"production_adapter_objective_closeout_writer_invocation_review_blackbox_fixture",
			"objective_closeout_writer_invocation_review_fixture_projection_only",
			"host_owned_objective_closeout_writer_adapter",
			"host_cli_objective_closeout_writer_invocation_display",
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

func unavailableProductionAdapterObjectiveCloseoutWriterInvocationReview() ProductionAdapterObjectiveCloseoutWriterInvocationReview {
	return ProductionAdapterObjectiveCloseoutWriterInvocationReview{
		ContractVersion: ContractVersion,
		Projected:       true,
		Available:       false,
		Status:          "unavailable",
		Mode:            "production_adapter_objective_closeout_writer_invocation_review",
		DisplayState:    "unavailable",
		DisplayStage:    "unavailable",
		DisplaySections: productionAdapterObjectiveCloseoutWriterInvocationReviewDisplaySections(),
		MissingInputs:   []MissingInput{"host:objective_closeout_writer_invocation_envelope"},
		Boundaries: []Boundary{
			"production_adapter_objective_closeout_writer_invocation_review",
			"objective_closeout_writer_invocation_review_projection_only",
			"host_owned_objective_closeout_writer_adapter",
			"host_cli_objective_closeout_writer_invocation_display",
			"display_safe_refs_only",
			"no_runner_dispatch",
			"no_adapter_invocation",
			"no_durable_write_by_core",
		},
		RunnerEffect:   "none",
		PromptEffect:   "none",
		NextHostAction: "provide_objective_closeout_writer_invocation_envelope",
	}
}

func unavailableProductionAdapterObjectiveCloseoutWriterInvocationReviewBlackboxFixture() ProductionAdapterObjectiveCloseoutWriterInvocationReviewBlackboxFixture {
	return ProductionAdapterObjectiveCloseoutWriterInvocationReviewBlackboxFixture{
		ContractVersion: ContractVersion,
		Projected:       true,
		Available:       false,
		Status:          "unavailable",
		Mode:            "production_adapter_objective_closeout_writer_invocation_review_blackbox_fixture",
		DisplayState:    "unavailable",
		DisplayStage:    "unavailable",
		DisplaySections: append(productionAdapterObjectiveCloseoutWriterInvocationReviewDisplaySections(), "objective_closeout_writer_invocation_blackbox_assertions"),
		MissingInputs:   []MissingInput{"host:objective_closeout_writer_invocation_review"},
		Boundaries: []Boundary{
			"production_adapter_objective_closeout_writer_invocation_review_blackbox_fixture",
			"objective_closeout_writer_invocation_review_fixture_projection_only",
			"host_owned_objective_closeout_writer_adapter",
			"host_cli_objective_closeout_writer_invocation_display",
			"display_safe_refs_only",
			"no_runner_dispatch",
			"no_adapter_invocation",
			"no_durable_write_by_core",
		},
		RunnerEffect:   "none",
		PromptEffect:   "none",
		NextHostAction: "provide_objective_closeout_writer_invocation_review",
	}
}
