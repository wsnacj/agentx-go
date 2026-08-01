package controlcontract

type ProductionAdapterObjectiveCloseoutWriterInvocationHostUIHandoffInput struct {
	HostUIHandoffRef DisplaySafeRef                                                          `json:"host_ui_handoff_ref,omitempty"`
	Review           ProductionAdapterObjectiveCloseoutWriterInvocationReview                `json:"review,omitempty"`
	DisplayFixture   ProductionAdapterObjectiveCloseoutWriterInvocationReviewBlackboxFixture `json:"display_fixture,omitempty"`
	RawOutputLoaded  bool                                                                    `json:"raw_output_loaded"`
}

type ProductionAdapterObjectiveCloseoutWriterInvocationHostUIHandoff struct {
	ContractVersion                 string           `json:"contract_version,omitempty"`
	Projected                       bool             `json:"projected"`
	Available                       bool             `json:"available"`
	Status                          string           `json:"status,omitempty"`
	Mode                            string           `json:"mode,omitempty"`
	DisplayState                    string           `json:"display_state,omitempty"`
	DisplayStage                    string           `json:"display_stage,omitempty"`
	DisplaySteps                    []string         `json:"display_steps,omitempty"`
	DisplaySections                 []string         `json:"display_sections,omitempty"`
	ReadyForHostDisplay             bool             `json:"ready_for_host_display"`
	ReadyForHostAdapterInvocation   bool             `json:"ready_for_host_adapter_invocation"`
	ReadyForDurableReadbackReview   bool             `json:"ready_for_durable_readback_review"`
	ReadyForFailureReview           bool             `json:"ready_for_failure_review"`
	ReadyForCompensationReview      bool             `json:"ready_for_compensation_review"`
	ReadyForBlockedReview           bool             `json:"ready_for_blocked_review"`
	InvocationReadyDisplay          bool             `json:"invocation_ready_display"`
	ResultReadbackDisplay           bool             `json:"result_readback_display"`
	FailureReviewDisplay            bool             `json:"failure_review_display"`
	BlockedDisplay                  bool             `json:"blocked_display"`
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
	HostUIHandoffRef                DisplaySafeRef   `json:"host_ui_handoff_ref,omitempty"`
	PrimaryDisplayRef               DisplaySafeRef   `json:"primary_display_ref,omitempty"`
	ReviewRef                       DisplaySafeRef   `json:"review_ref,omitempty"`
	FixtureRef                      DisplaySafeRef   `json:"fixture_ref,omitempty"`
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

func BuildProductionAdapterObjectiveCloseoutWriterInvocationHostUIHandoff(input ProductionAdapterObjectiveCloseoutWriterInvocationHostUIHandoffInput) ProductionAdapterObjectiveCloseoutWriterInvocationHostUIHandoff {
	if productionAdapterObjectiveCloseoutWriterInvocationReviewEmpty(input.Review) {
		return unavailableProductionAdapterObjectiveCloseoutWriterInvocationHostUIHandoff()
	}
	review := input.Review.Normalize()
	fixtureProvided := !productionAdapterObjectiveCloseoutWriterInvocationReviewBlackboxFixtureEmpty(input.DisplayFixture)
	fixture := ProductionAdapterObjectiveCloseoutWriterInvocationReviewBlackboxFixture{}
	if fixtureProvided {
		fixture = input.DisplayFixture.Normalize()
	}
	result := ProductionAdapterObjectiveCloseoutWriterInvocationHostUIHandoff{
		ContractVersion:                 ContractVersion,
		Projected:                       true,
		Available:                       review.Available,
		Status:                          "blocked",
		Mode:                            "production_adapter_objective_closeout_writer_invocation_host_ui_handoff",
		DisplayState:                    review.DisplayState,
		DisplayStage:                    productionAdapterObjectiveCloseoutWriterInvocationHostUIHandoffStage(review),
		DisplaySteps:                    productionAdapterObjectiveCloseoutWriterInvocationHostUIHandoffSteps(review),
		DisplaySections:                 productionAdapterObjectiveCloseoutWriterInvocationHostUIHandoffSections(review),
		HostUIHandoffRef:                normalizeOneDisplaySafeRef(input.HostUIHandoffRef),
		PrimaryDisplayRef:               productionAdapterObjectiveCloseoutWriterInvocationHostUIHandoffPrimaryDisplayRef(review, fixture, fixtureProvided),
		ReviewRef:                       review.ReviewRef,
		FixtureRef:                      fixture.FixtureRef,
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
		MissingInputs:                   productionAdapterObjectiveCloseoutWriterInvocationHostUIHandoffMissingInputs(review, fixture, fixtureProvided),
		BlockedReasons:                  productionAdapterObjectiveCloseoutWriterInvocationHostUIHandoffBlockedReasons(review, fixture, fixtureProvided),
		FailureClass:                    firstFailureClass(review.FailureClass, fixture.FailureClass),
		Boundaries:                      productionAdapterObjectiveCloseoutWriterInvocationHostUIHandoffBoundaries(review.Boundaries, fixture.Boundaries),
		NextHostAction:                  review.NextHostAction,
		RunnerEffect:                    "none",
		PromptEffect:                    "none",
		RawOutputLoaded:                 input.RawOutputLoaded || review.RawOutputLoaded || fixture.RawOutputLoaded,
		HostAdapterInvocationAuthorized: review.HostAdapterInvocationAuthorized,
		HostAdapterInvocationBound:      review.HostAdapterInvocationBound,
		HostDurableWriteReported:        review.HostDurableWriteReported,
		HostDurableWriteSucceeded:       review.HostDurableWriteSucceeded,
		HostDurableWriteFailed:          review.HostDurableWriteFailed,
		HostDurableWriteRecorded:        review.HostDurableWriteRecorded,
	}
	if input.RawOutputLoaded || productionAdapterObjectiveCloseoutWriterInvocationHostUIHandoffUnsafe(input, review, fixture, fixtureProvided) {
		result.RawOutputLoaded = true
		result = productionAdapterObjectiveCloseoutWriterInvocationHostUIHandoffBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if result.HostUIHandoffRef == "" {
		result = productionAdapterObjectiveCloseoutWriterInvocationHostUIHandoffBlock(result, FailureEvidenceMissing, "writer_invocation_host_ui_handoff_ref_missing", "host:objective_closeout_writer_invocation_host_ui_handoff_ref", "provide_objective_closeout_writer_invocation_host_ui_handoff_ref")
	}
	if !review.ReadyForHostDisplay {
		result = productionAdapterObjectiveCloseoutWriterInvocationHostUIHandoffBlock(result, firstFailureClass(review.FailureClass, FailureEvidenceMissing), "writer_invocation_review_not_ready", "host:objective_closeout_writer_invocation_review", firstNextHostAction(review.NextHostAction, "review_objective_closeout_writer_invocation"))
	}
	if fixtureProvided && !fixture.ReadyForHostDisplay {
		result = productionAdapterObjectiveCloseoutWriterInvocationHostUIHandoffBlock(result, firstFailureClass(fixture.FailureClass, FailureEvidenceMissing), "writer_invocation_review_fixture_not_ready", "host:objective_closeout_writer_invocation_review_fixture", firstNextHostAction(fixture.NextHostAction, "review_objective_closeout_writer_invocation_display"))
	}
	for _, mismatch := range productionAdapterObjectiveCloseoutWriterInvocationHostUIHandoffMismatches(review, fixture, fixtureProvided) {
		result = productionAdapterObjectiveCloseoutWriterInvocationHostUIHandoffBlock(result, FailureVerificationFailed, mismatch.reason, mismatch.missing, "review_objective_closeout_writer_invocation_host_ui_handoff")
	}
	if result.HostUIHandoffRef != "" && review.ReadyForHostDisplay {
		result.ReadyForHostDisplay = true
		switch review.Status {
		case "ready_for_objective_closeout_writer_invocation_review":
			if review.ReadyForHostAdapterInvocation && len(result.MissingInputs) == 0 && len(result.BlockedReasons) == 0 {
				result.Status = "ready_for_objective_closeout_writer_invocation_handoff"
				result.DisplayStage = "invocation_ready"
				result.ReadyForHostAdapterInvocation = true
				result.InvocationReadyDisplay = true
				result.HostMayInvokeWriterAdapter = true
				result.HostMayExecuteDurableWrite = true
				result.NextHostAction = firstNextHostAction(review.NextHostAction, "host_may_invoke_objective_closeout_writer_adapter")
				result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_objective_closeout_writer_invocation_handoff", "host_ui_objective_closeout_writer_invocation_handoff")
			}
		case "ready_for_objective_closeout_writer_invocation_result_readback_review":
			if review.ReadyForDurableReadbackReview && len(result.MissingInputs) == 0 && len(result.BlockedReasons) == 0 {
				result.Status = "ready_for_objective_closeout_writer_invocation_result_readback_handoff"
				result.DisplayStage = "result_readback"
				result.ReadyForDurableReadbackReview = true
				result.ResultReadbackDisplay = true
				result.HostMayInvokeWriterAdapter = false
				result.HostMayExecuteDurableWrite = false
				result.NextHostAction = firstNextHostAction(review.NextHostAction, "bind_objective_closeout_writer_durable_readback")
				result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_objective_closeout_writer_invocation_result_readback_handoff", "host_ui_objective_closeout_writer_invocation_result_handoff")
			}
		case "ready_for_objective_closeout_writer_invocation_result_failure_review":
			if review.ReadyForFailureReview && len(result.MissingInputs) == 0 && len(result.BlockedReasons) == 0 {
				result.Status = "ready_for_objective_closeout_writer_invocation_result_failure_handoff"
				result.DisplayStage = "failure_review"
				result.ReadyForFailureReview = true
				result.ReadyForCompensationReview = review.ReadyForCompensationReview
				result.FailureReviewDisplay = true
				result.HostMayInvokeWriterAdapter = false
				result.HostMayExecuteDurableWrite = false
				result.NextHostAction = firstNextHostAction(review.NextHostAction, "review_objective_closeout_writer_durable_failure")
				result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_objective_closeout_writer_invocation_result_failure_handoff", "host_ui_objective_closeout_writer_invocation_failure_handoff", "compensation_not_executed")
			}
		case "ready_for_objective_closeout_writer_invocation_blocked_review":
			if review.ReadyForBlockedReview {
				result.Status = "ready_for_objective_closeout_writer_invocation_blocked_handoff"
				result.DisplayStage = "blocked_review"
				result.ReadyForBlockedReview = true
				result.BlockedDisplay = true
				result.HostMayInvokeWriterAdapter = false
				result.HostMayExecuteDurableWrite = false
				result.NextHostAction = firstNextHostAction(review.NextHostAction, "review_objective_closeout_writer_invocation_blocked")
				result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_objective_closeout_writer_invocation_blocked_handoff", "host_ui_objective_closeout_writer_invocation_blocked_handoff")
			}
		default:
			result = productionAdapterObjectiveCloseoutWriterInvocationHostUIHandoffBlock(result, firstFailureClass(review.FailureClass, FailureEvidenceMissing), "writer_invocation_review_not_ready", "host:objective_closeout_writer_invocation_review", firstNextHostAction(review.NextHostAction, "review_objective_closeout_writer_invocation"))
		}
	}
	return result.Normalize()
}

func CloneProductionAdapterObjectiveCloseoutWriterInvocationHostUIHandoff(in ProductionAdapterObjectiveCloseoutWriterInvocationHostUIHandoff) ProductionAdapterObjectiveCloseoutWriterInvocationHostUIHandoff {
	out := in
	out.DisplaySteps = cloneStringSlice(in.DisplaySteps)
	out.DisplaySections = cloneStringSlice(in.DisplaySections)
	out.CapabilityProofRefs = cloneDisplaySafeRefs(in.CapabilityProofRefs)
	out.ApprovalBindingRefs = cloneDisplaySafeRefs(in.ApprovalBindingRefs)
	out.DurableEvidenceRefs = cloneDisplaySafeRefs(in.DurableEvidenceRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (h ProductionAdapterObjectiveCloseoutWriterInvocationHostUIHandoff) Clone() ProductionAdapterObjectiveCloseoutWriterInvocationHostUIHandoff {
	return CloneProductionAdapterObjectiveCloseoutWriterInvocationHostUIHandoff(h)
}

func (h ProductionAdapterObjectiveCloseoutWriterInvocationHostUIHandoff) Normalize() ProductionAdapterObjectiveCloseoutWriterInvocationHostUIHandoff {
	out := CloneProductionAdapterObjectiveCloseoutWriterInvocationHostUIHandoff(h)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = normalizeControlToken(out.Status)
	out.Mode = normalizeControlToken(out.Mode)
	if out.Mode == "" {
		out.Mode = "production_adapter_objective_closeout_writer_invocation_host_ui_handoff"
	}
	out.DisplayState = normalizeControlToken(out.DisplayState)
	if out.DisplayState == "" {
		out.DisplayState = "blocked"
	}
	out.DisplayStage = normalizeControlToken(out.DisplayStage)
	if out.DisplayStage == "" {
		out.DisplayStage = "blocked"
	}
	out.DisplaySteps = normalizeControlTokenList(out.DisplaySteps)
	out.DisplaySections = normalizeControlTokenList(out.DisplaySections)
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
		out.Status = "blocked"
	}
	if out.RawOutputLoaded || productionAdapterObjectiveCloseoutWriterInvocationHostUIHandoffUnsafeOutput(out) {
		out.RawOutputLoaded = true
		out = productionAdapterObjectiveCloseoutWriterInvocationHostUIHandoffBlock(out, firstFailureClass(out.FailureClass, FailureEvidenceWeak), "unsafe_input_ref", "host:display_safe_refs", firstNextHostAction(out.NextHostAction, "provide_display_safe_refs"))
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
		out.InvocationReadyDisplay = false
		out.ResultReadbackDisplay = false
		out.FailureReviewDisplay = false
		out.BlockedDisplay = false
		out.HostMayInvokeWriterAdapter = false
		out.HostMayExecuteDurableWrite = false
		out.HostAdapterInvocationBound = false
		return out
	}
	out.ReadyForHostDisplay = out.ReadyForHostDisplay &&
		out.HostUIHandoffRef != "" &&
		out.PrimaryDisplayRef != "" &&
		out.ReviewRef != "" &&
		out.InvocationEnvelopeRef != "" &&
		out.WriterRef != "" &&
		!out.RawOutputLoaded
	invocationReady := out.Status == "ready_for_objective_closeout_writer_invocation_handoff" &&
		out.ReadyForHostDisplay &&
		out.DisplayState == "host_adapter_invocation_ready" &&
		out.ReadyForHostAdapterInvocation &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0
	readbackReady := out.Status == "ready_for_objective_closeout_writer_invocation_result_readback_handoff" &&
		out.ReadyForHostDisplay &&
		out.DisplayState == "invocation_result_readback_bound" &&
		out.ReadyForDurableReadbackReview &&
		out.HostAdapterInvocationBound &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0
	failureReady := out.Status == "ready_for_objective_closeout_writer_invocation_result_failure_handoff" &&
		out.ReadyForHostDisplay &&
		out.DisplayState == "invocation_result_failed" &&
		out.ReadyForFailureReview &&
		out.HostAdapterInvocationBound &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0
	blockedReady := out.Status == "ready_for_objective_closeout_writer_invocation_blocked_handoff" &&
		out.ReadyForHostDisplay &&
		out.ReadyForBlockedReview &&
		(len(out.MissingInputs) > 0 || len(out.BlockedReasons) > 0)
	out.InvocationReadyDisplay = out.InvocationReadyDisplay && invocationReady
	out.ResultReadbackDisplay = out.ResultReadbackDisplay && readbackReady
	out.FailureReviewDisplay = out.FailureReviewDisplay && failureReady
	out.BlockedDisplay = out.BlockedDisplay && blockedReady
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

func productionAdapterObjectiveCloseoutWriterInvocationHostUIHandoffBlock(result ProductionAdapterObjectiveCloseoutWriterInvocationHostUIHandoff, failure FailureClass, reason string, missing MissingInput, next NextHostAction) ProductionAdapterObjectiveCloseoutWriterInvocationHostUIHandoff {
	result.Status = "blocked"
	result.ReadyForHostAdapterInvocation = false
	result.ReadyForDurableReadbackReview = false
	result.ReadyForFailureReview = false
	result.ReadyForCompensationReview = false
	result.ReadyForBlockedReview = false
	result.InvocationReadyDisplay = false
	result.ResultReadbackDisplay = false
	result.FailureReviewDisplay = false
	result.BlockedDisplay = false
	result.HostMayInvokeWriterAdapter = false
	result.HostMayExecuteDurableWrite = false
	result.HostAdapterInvocationBound = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	result.Boundaries = AppendBoundaries(result.Boundaries, "objective_closeout_writer_invocation_host_ui_handoff_blocked")
	return result
}

func productionAdapterObjectiveCloseoutWriterInvocationHostUIHandoffStage(review ProductionAdapterObjectiveCloseoutWriterInvocationReview) string {
	switch review.Status {
	case "ready_for_objective_closeout_writer_invocation_review":
		return "invocation_ready"
	case "ready_for_objective_closeout_writer_invocation_result_readback_review":
		return "result_readback"
	case "ready_for_objective_closeout_writer_invocation_result_failure_review":
		return "failure_review"
	case "ready_for_objective_closeout_writer_invocation_blocked_review":
		return "blocked_review"
	default:
		return firstControlToken(review.DisplayStage, "blocked")
	}
}

func productionAdapterObjectiveCloseoutWriterInvocationHostUIHandoffSteps(review ProductionAdapterObjectiveCloseoutWriterInvocationReview) []string {
	switch review.Status {
	case "ready_for_objective_closeout_writer_invocation_review":
		return []string{"writer_adapter_invocation", "durable_writer_execution", "result_readback"}
	case "ready_for_objective_closeout_writer_invocation_result_readback_review":
		return []string{"writer_adapter_result", "durable_readback_review"}
	case "ready_for_objective_closeout_writer_invocation_result_failure_review":
		if review.ReadyForCompensationReview {
			return []string{"writer_adapter_result", "failure_review", "compensation_review"}
		}
		return []string{"writer_adapter_result", "failure_review"}
	case "ready_for_objective_closeout_writer_invocation_blocked_review":
		return []string{"writer_adapter_blocked", "required_inputs"}
	default:
		return []string{"writer_adapter_review"}
	}
}

func productionAdapterObjectiveCloseoutWriterInvocationHostUIHandoffSections(review ProductionAdapterObjectiveCloseoutWriterInvocationReview) []string {
	sections := append([]string{"writer_invocation_handoff_summary"}, cloneStringSlice(review.DisplaySections)...)
	return normalizeControlTokenList(sections)
}

func productionAdapterObjectiveCloseoutWriterInvocationHostUIHandoffPrimaryDisplayRef(review ProductionAdapterObjectiveCloseoutWriterInvocationReview, fixture ProductionAdapterObjectiveCloseoutWriterInvocationReviewBlackboxFixture, fixtureProvided bool) DisplaySafeRef {
	if fixtureProvided && fixture.FixtureRef != "" {
		return fixture.FixtureRef
	}
	return review.ReviewRef
}

func productionAdapterObjectiveCloseoutWriterInvocationHostUIHandoffMissingInputs(review ProductionAdapterObjectiveCloseoutWriterInvocationReview, fixture ProductionAdapterObjectiveCloseoutWriterInvocationReviewBlackboxFixture, fixtureProvided bool) []MissingInput {
	out := cloneMissingInputs(review.MissingInputs)
	if fixtureProvided {
		out = append(out, cloneMissingInputs(fixture.MissingInputs)...)
	}
	return normalizeMissingInputs(out)
}

func productionAdapterObjectiveCloseoutWriterInvocationHostUIHandoffBlockedReasons(review ProductionAdapterObjectiveCloseoutWriterInvocationReview, fixture ProductionAdapterObjectiveCloseoutWriterInvocationReviewBlackboxFixture, fixtureProvided bool) []string {
	out := cloneStringSlice(review.BlockedReasons)
	if fixtureProvided {
		out = append(out, cloneStringSlice(fixture.BlockedReasons)...)
	}
	return normalizeControlTokenList(out)
}

type productionAdapterObjectiveCloseoutWriterInvocationHostUIHandoffMismatch struct {
	reason  string
	missing MissingInput
}

func productionAdapterObjectiveCloseoutWriterInvocationHostUIHandoffMismatches(review ProductionAdapterObjectiveCloseoutWriterInvocationReview, fixture ProductionAdapterObjectiveCloseoutWriterInvocationReviewBlackboxFixture, fixtureProvided bool) []productionAdapterObjectiveCloseoutWriterInvocationHostUIHandoffMismatch {
	if !fixtureProvided {
		return nil
	}
	var out []productionAdapterObjectiveCloseoutWriterInvocationHostUIHandoffMismatch
	out = append(out, productionAdapterObjectiveCloseoutWriterInvocationHostUIHandoffRefMismatch(review.ReviewRef, fixture.ReviewRef, "writer_invocation_handoff_fixture_review_ref_mismatch", "host:objective_closeout_writer_invocation_review_ref")...)
	out = append(out, productionAdapterObjectiveCloseoutWriterInvocationHostUIHandoffRefMismatch(review.InvocationEnvelopeRef, fixture.InvocationEnvelopeRef, "writer_invocation_handoff_fixture_invocation_envelope_ref_mismatch", "host:objective_closeout_writer_invocation_envelope_ref")...)
	out = append(out, productionAdapterObjectiveCloseoutWriterInvocationHostUIHandoffRefMismatch(review.ResultEnvelopeRef, fixture.ResultEnvelopeRef, "writer_invocation_handoff_fixture_result_envelope_ref_mismatch", "host:objective_closeout_writer_invocation_result_envelope_ref")...)
	if fixture.Status != productionAdapterObjectiveCloseoutWriterInvocationReviewFixtureStatus(review.Status) {
		out = append(out, productionAdapterObjectiveCloseoutWriterInvocationHostUIHandoffMismatch{
			reason:  "writer_invocation_handoff_fixture_status_mismatch",
			missing: "host:objective_closeout_writer_invocation_review_fixture",
		})
	}
	return out
}

func productionAdapterObjectiveCloseoutWriterInvocationHostUIHandoffRefMismatch(left DisplaySafeRef, right DisplaySafeRef, reason string, missing MissingInput) []productionAdapterObjectiveCloseoutWriterInvocationHostUIHandoffMismatch {
	if left == right {
		return nil
	}
	if left == "" && right == "" {
		return nil
	}
	return []productionAdapterObjectiveCloseoutWriterInvocationHostUIHandoffMismatch{{
		reason:  reason,
		missing: missing,
	}}
}

func productionAdapterObjectiveCloseoutWriterInvocationHostUIHandoffBoundaries(groups ...[]Boundary) []Boundary {
	var out []Boundary
	out = AppendBoundaries(out,
		"production_adapter_objective_closeout_writer_invocation_host_ui_handoff",
		"objective_closeout_writer_invocation_host_ui_handoff_projection_only",
		"host_ui_objective_closeout_writer_invocation_handoff",
		"display_safe_refs_only",
		"no_runner_dispatch",
		"no_adapter_invocation",
		"no_durable_write_by_core",
		"no_objective_store_write_by_core",
		"no_runstore_write_by_core",
	)
	for _, group := range groups {
		out = AppendBoundaries(out, group...)
	}
	return normalizeBoundaries(out)
}

func productionAdapterObjectiveCloseoutWriterInvocationHostUIHandoffUnsafe(input ProductionAdapterObjectiveCloseoutWriterInvocationHostUIHandoffInput, review ProductionAdapterObjectiveCloseoutWriterInvocationReview, fixture ProductionAdapterObjectiveCloseoutWriterInvocationReviewBlackboxFixture, fixtureProvided bool) bool {
	return input.RawOutputLoaded ||
		displaySafeRefRejected(input.HostUIHandoffRef) ||
		productionAdapterObjectiveCloseoutWriterInvocationReviewUnsafeOutput(review) ||
		(fixtureProvided && productionAdapterObjectiveCloseoutWriterInvocationReviewFixtureUnsafeOutput(fixture))
}

func productionAdapterObjectiveCloseoutWriterInvocationHostUIHandoffUnsafeOutput(input ProductionAdapterObjectiveCloseoutWriterInvocationHostUIHandoff) bool {
	return displaySafeRefRejected(input.HostUIHandoffRef) ||
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

func productionAdapterObjectiveCloseoutWriterInvocationReviewBlackboxFixtureEmpty(fixture ProductionAdapterObjectiveCloseoutWriterInvocationReviewBlackboxFixture) bool {
	return !fixture.Projected &&
		!fixture.Available &&
		fixture.Status == "" &&
		fixture.Mode == "" &&
		fixture.FixtureRef == "" &&
		fixture.ReviewRef == "" &&
		fixture.InvocationEnvelopeRef == "" &&
		fixture.ResultEnvelopeRef == "" &&
		fixture.ReviewPacketRef == "" &&
		fixture.DurableRequestRef == "" &&
		fixture.DurableResultRef == "" &&
		fixture.ExpectedDurableResultRef == "" &&
		fixture.WriterInvocationRef == "" &&
		fixture.WriterRef == "" &&
		fixture.HostWriterBindingRef == "" &&
		fixture.HostAdapterVersionRef == "" &&
		fixture.ExpectedHostAdapterRunRef == "" &&
		fixture.HostAdapterRunRef == "" &&
		fixture.ExpectedReadbackRef == "" &&
		fixture.ExpectedFailureRef == "" &&
		fixture.ExpectedCompensationRef == "" &&
		fixture.AppliedDurableEventRef == "" &&
		fixture.AppliedRunstoreRef == "" &&
		fixture.AppliedObjectiveStateRef == "" &&
		fixture.FailureRef == "" &&
		fixture.CompensationRef == "" &&
		fixture.HostDurableWriteConfirmationRef == "" &&
		len(fixture.CapabilityProofRefs) == 0 &&
		len(fixture.ApprovalBindingRefs) == 0 &&
		len(fixture.DurableEvidenceRefs) == 0 &&
		len(fixture.MissingInputs) == 0 &&
		len(fixture.BlockedReasons) == 0 &&
		len(fixture.Boundaries) == 0 &&
		fixture.NextHostAction == "" &&
		!fixture.RawOutputLoaded
}

func productionAdapterObjectiveCloseoutWriterInvocationHostUIHandoffEmpty(handoff ProductionAdapterObjectiveCloseoutWriterInvocationHostUIHandoff) bool {
	return !handoff.Projected &&
		!handoff.Available &&
		handoff.Status == "" &&
		handoff.Mode == "" &&
		handoff.HostUIHandoffRef == "" &&
		handoff.ReviewRef == "" &&
		handoff.PrimaryDisplayRef == "" &&
		handoff.FixtureRef == "" &&
		handoff.InvocationEnvelopeRef == "" &&
		handoff.ResultEnvelopeRef == "" &&
		handoff.ReviewPacketRef == "" &&
		handoff.DurableRequestRef == "" &&
		handoff.DurableResultRef == "" &&
		handoff.ExpectedDurableResultRef == "" &&
		handoff.WriterInvocationRef == "" &&
		handoff.WriterRef == "" &&
		handoff.HostWriterBindingRef == "" &&
		handoff.HostAdapterVersionRef == "" &&
		handoff.ExpectedHostAdapterRunRef == "" &&
		handoff.HostAdapterRunRef == "" &&
		handoff.ExpectedReadbackRef == "" &&
		handoff.ExpectedFailureRef == "" &&
		handoff.ExpectedCompensationRef == "" &&
		handoff.AppliedDurableEventRef == "" &&
		handoff.AppliedRunstoreRef == "" &&
		handoff.AppliedObjectiveStateRef == "" &&
		handoff.FailureRef == "" &&
		handoff.CompensationRef == "" &&
		handoff.HostDurableWriteConfirmationRef == "" &&
		len(handoff.CapabilityProofRefs) == 0 &&
		len(handoff.ApprovalBindingRefs) == 0 &&
		len(handoff.DurableEvidenceRefs) == 0 &&
		len(handoff.MissingInputs) == 0 &&
		len(handoff.BlockedReasons) == 0 &&
		len(handoff.Boundaries) == 0 &&
		handoff.NextHostAction == "" &&
		!handoff.RawOutputLoaded
}

func unavailableProductionAdapterObjectiveCloseoutWriterInvocationHostUIHandoff() ProductionAdapterObjectiveCloseoutWriterInvocationHostUIHandoff {
	return ProductionAdapterObjectiveCloseoutWriterInvocationHostUIHandoff{
		ContractVersion: ContractVersion,
		Projected:       true,
		Available:       false,
		Status:          "unavailable",
		Mode:            "production_adapter_objective_closeout_writer_invocation_host_ui_handoff",
		DisplayState:    "unavailable",
		DisplayStage:    "unavailable",
		RunnerEffect:    "none",
		PromptEffect:    "none",
	}
}
