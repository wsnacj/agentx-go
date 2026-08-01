package controlcontract

type ProductionAdapterObjectiveCloseoutHostUIHandoffInput struct {
	HostUIHandoffRef     DisplaySafeRef                                                 `json:"host_ui_handoff_ref,omitempty"`
	HostView             ProductionAdapterObjectiveCloseoutHostView                     `json:"host_view,omitempty"`
	DisplayFixture       ProductionAdapterObjectiveCloseoutBlackboxFixture              `json:"display_fixture,omitempty"`
	FailureReview        ProductionAdapterObjectiveCloseoutFailureReviewPacket          `json:"failure_review,omitempty"`
	FailureReviewFixture ProductionAdapterObjectiveCloseoutFailureReviewBlackboxFixture `json:"failure_review_fixture,omitempty"`
	RawOutputLoaded      bool                                                           `json:"raw_output_loaded"`
}

type ProductionAdapterObjectiveCloseoutHostUIHandoff struct {
	ContractVersion                          string         `json:"contract_version,omitempty"`
	Projected                                bool           `json:"projected"`
	Available                                bool           `json:"available"`
	Status                                   string         `json:"status,omitempty"`
	Mode                                     string         `json:"mode,omitempty"`
	DisplayState                             string         `json:"display_state,omitempty"`
	DisplayStage                             string         `json:"display_stage,omitempty"`
	DisplaySteps                             []string       `json:"display_steps,omitempty"`
	ReadyForHostDisplay                      bool           `json:"ready_for_host_display"`
	ReadyForHostDurableApply                 bool           `json:"ready_for_host_durable_apply"`
	ReadyForFailureReview                    bool           `json:"ready_for_failure_review"`
	ReadyForCompensationReview               bool           `json:"ready_for_compensation_review"`
	ReadyForObjectiveReturn                  bool           `json:"ready_for_objective_return"`
	IntermediateDisplay                      bool           `json:"intermediate_display"`
	FinalDisplay                             bool           `json:"final_display"`
	FailureReviewDisplay                     bool           `json:"failure_review_display"`
	ObjectiveLifecycleClosed                 bool           `json:"objective_lifecycle_closed"`
	ObjectiveSatisfied                       bool           `json:"objective_satisfied"`
	CoreInvocationExecuted                   bool           `json:"core_invocation_executed"`
	DurableWriteByCore                       bool           `json:"durable_write_by_core"`
	ObjectiveStoreWriteByCore                bool           `json:"objective_store_write_by_core"`
	RunstoreWriteByCore                      bool           `json:"runstore_write_by_core"`
	HostUIHandoffRef                         DisplaySafeRef `json:"host_ui_handoff_ref,omitempty"`
	PrimaryDisplayRef                        DisplaySafeRef `json:"primary_display_ref,omitempty"`
	HostViewRef                              DisplaySafeRef `json:"host_view_ref,omitempty"`
	DisplayFixtureRef                        DisplaySafeRef `json:"display_fixture_ref,omitempty"`
	FailureReviewPacketRef                   DisplaySafeRef `json:"failure_review_packet_ref,omitempty"`
	FailureReviewFixtureRef                  DisplaySafeRef `json:"failure_review_fixture_ref,omitempty"`
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
	FailureRef                               DisplaySafeRef `json:"failure_ref,omitempty"`
	CompensationRef                          DisplaySafeRef `json:"compensation_ref,omitempty"`
	HostDurableApplyConfirmationRef          DisplaySafeRef `json:"host_durable_apply_confirmation_ref,omitempty"`
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

func BuildProductionAdapterObjectiveCloseoutHostUIHandoff(input ProductionAdapterObjectiveCloseoutHostUIHandoffInput) ProductionAdapterObjectiveCloseoutHostUIHandoff {
	if productionAdapterObjectiveCloseoutHostViewEmpty(input.HostView) {
		return unavailableProductionAdapterObjectiveCloseoutHostUIHandoff()
	}
	view := input.HostView.Normalize()
	fixtureProvided := !productionAdapterObjectiveCloseoutBlackboxFixtureEmpty(input.DisplayFixture)
	fixture := ProductionAdapterObjectiveCloseoutBlackboxFixture{}
	if fixtureProvided {
		fixture = input.DisplayFixture.Normalize()
	}
	failureReviewProvided := !productionAdapterObjectiveCloseoutFailureReviewPacketEmpty(input.FailureReview)
	failureReview := ProductionAdapterObjectiveCloseoutFailureReviewPacket{}
	if failureReviewProvided {
		failureReview = input.FailureReview.Normalize()
	}
	failureFixtureProvided := !productionAdapterObjectiveCloseoutFailureReviewBlackboxFixtureEmpty(input.FailureReviewFixture)
	failureFixture := ProductionAdapterObjectiveCloseoutFailureReviewBlackboxFixture{}
	if failureFixtureProvided {
		failureFixture = input.FailureReviewFixture.Normalize()
	}
	stage := productionAdapterObjectiveCloseoutHostUIHandoffStage(view)
	result := ProductionAdapterObjectiveCloseoutHostUIHandoff{
		ContractVersion:                          ContractVersion,
		Projected:                                true,
		Available:                                view.Available,
		Status:                                   "blocked",
		Mode:                                     "production_adapter_objective_closeout_host_ui_handoff",
		DisplayState:                             view.DisplayState,
		DisplayStage:                             stage,
		DisplaySteps:                             productionAdapterObjectiveCloseoutHostUIHandoffSteps(stage, view.ReadyForCompensationReview),
		HostUIHandoffRef:                         normalizeOneDisplaySafeRef(input.HostUIHandoffRef),
		PrimaryDisplayRef:                        productionAdapterObjectiveCloseoutHostUIHandoffPrimaryDisplayRef(view, fixture, fixtureProvided, failureReview, failureReviewProvided, failureFixture, failureFixtureProvided),
		HostViewRef:                              view.HostViewRef,
		DisplayFixtureRef:                        fixture.FixtureRef,
		FailureReviewPacketRef:                   failureReview.FailureReviewPacketRef,
		FailureReviewFixtureRef:                  failureFixture.FixtureRef,
		DisplaySections:                          productionAdapterObjectiveCloseoutHostUIHandoffDisplaySections(stage),
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
		FailureRef:                               firstDisplaySafeRef(failureReview.FailureRef, failureFixture.FailureRef, view.FailureRef),
		CompensationRef:                          firstDisplaySafeRef(failureReview.CompensationRef, failureFixture.CompensationRef, view.CompensationRef),
		HostDurableApplyConfirmationRef:          view.HostDurableApplyConfirmationRef,
		CompletionAuditResultBindingRef:          view.CompletionAuditResultBindingRef,
		AdapterRef:                               view.AdapterRef,
		MissingInputs:                            productionAdapterObjectiveCloseoutHostUIHandoffMissingInputs(view, fixture, fixtureProvided, failureReview, failureReviewProvided, failureFixture, failureFixtureProvided),
		BlockedReasons:                           productionAdapterObjectiveCloseoutHostUIHandoffBlockedReasons(view, fixture, fixtureProvided, failureReview, failureReviewProvided, failureFixture, failureFixtureProvided),
		DisplayProgressMissingInputs:             productionAdapterObjectiveCloseoutHostUIHandoffProgressMissingInputs(view, fixture, fixtureProvided),
		DisplayProgressBlockedReasons:            productionAdapterObjectiveCloseoutHostUIHandoffProgressBlockedReasons(view, fixture, fixtureProvided),
		FailureClass:                             firstFailureClass(firstFailureClass(view.FailureClass, fixture.FailureClass), firstFailureClass(failureReview.FailureClass, failureFixture.FailureClass)),
		Boundaries:                               productionAdapterObjectiveCloseoutHostUIHandoffBoundaries(view.Boundaries, fixture.Boundaries, failureReview.Boundaries, failureFixture.Boundaries),
		NextHostAction:                           view.NextHostAction,
		RunnerEffect:                             "none",
		PromptEffect:                             "none",
		RawOutputLoaded:                          input.RawOutputLoaded || view.RawOutputLoaded || fixture.RawOutputLoaded || failureReview.RawOutputLoaded || failureFixture.RawOutputLoaded,
	}
	if productionAdapterObjectiveCloseoutHostUIHandoffUnsafe(input, view, fixture, fixtureProvided, failureReview, failureReviewProvided, failureFixture, failureFixtureProvided) {
		result = productionAdapterObjectiveCloseoutHostUIHandoffBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if result.HostUIHandoffRef == "" {
		result = productionAdapterObjectiveCloseoutHostUIHandoffBlock(result, FailureEvidenceMissing, "objective_closeout_host_ui_handoff_ref_missing", "host:objective_closeout_host_ui_handoff_ref", "provide_objective_closeout_host_ui_handoff_ref")
	}
	if !view.ReadyForHostDisplay {
		result = productionAdapterObjectiveCloseoutHostUIHandoffBlock(result, firstFailureClass(view.FailureClass, FailureEvidenceMissing), "objective_closeout_host_view_not_ready", "host:objective_closeout_host_view", firstNextHostAction(view.NextHostAction, "review_objective_closeout_host_view"))
	}
	for _, mismatch := range productionAdapterObjectiveCloseoutHostUIHandoffMismatches(view, fixture, fixtureProvided, failureReview, failureReviewProvided, failureFixture, failureFixtureProvided) {
		result = productionAdapterObjectiveCloseoutHostUIHandoffBlock(result, FailureVerificationFailed, mismatch.reason, mismatch.missing, "review_objective_closeout_host_ui_handoff")
	}
	if view.ReadyForFailureReview && !failureReviewProvided {
		result = productionAdapterObjectiveCloseoutHostUIHandoffBlock(result, FailureEvidenceMissing, "objective_closeout_failure_review_packet_not_ready", "host:objective_closeout_failure_review_packet", "review_objective_closeout_durable_failure")
	} else if failureReviewProvided && !failureReview.ReadyForFailureReview {
		result = productionAdapterObjectiveCloseoutHostUIHandoffBlock(result, firstFailureClass(failureReview.FailureClass, FailureEvidenceMissing), "objective_closeout_failure_review_packet_not_ready", "host:objective_closeout_failure_review_packet", firstNextHostAction(failureReview.NextHostAction, "review_objective_closeout_durable_failure"))
	}
	if failureFixtureProvided && !failureFixture.ReadyForFailureReview {
		result = productionAdapterObjectiveCloseoutHostUIHandoffBlock(result, firstFailureClass(failureFixture.FailureClass, FailureEvidenceMissing), "objective_closeout_failure_review_fixture_not_ready", "host:objective_closeout_failure_review_fixture", firstNextHostAction(failureFixture.NextHostAction, "review_objective_closeout_durable_failure"))
	}
	if len(result.MissingInputs) == 0 && len(result.BlockedReasons) == 0 && result.HostUIHandoffRef != "" && view.ReadyForHostDisplay {
		result.ReadyForHostDisplay = true
		switch {
		case view.ReadyForObjectiveReturn:
			result.Status = "ready_for_objective_return_handoff"
			result.DisplayStage = "final"
			result.DisplaySteps = productionAdapterObjectiveCloseoutHostUIHandoffSteps(result.DisplayStage, false)
			result.ReadyForObjectiveReturn = true
			result.FinalDisplay = true
			result.IntermediateDisplay = false
			result.FailureReviewDisplay = false
			result.NextHostAction = "return_objective_closed_lifecycle"
			result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_objective_return_handoff", "host_ui_objective_closeout_final_handoff")
		case view.ReadyForFailureReview:
			result.Status = "ready_for_objective_closeout_failure_review_handoff"
			result.DisplayStage = "failure_review"
			result.DisplaySteps = productionAdapterObjectiveCloseoutHostUIHandoffSteps(result.DisplayStage, view.ReadyForCompensationReview)
			result.ReadyForFailureReview = true
			result.ReadyForCompensationReview = view.ReadyForCompensationReview
			result.IntermediateDisplay = true
			result.FinalDisplay = false
			result.FailureReviewDisplay = true
			result.ObjectiveLifecycleClosed = false
			result.ObjectiveSatisfied = false
			result.NextHostAction = "review_objective_closeout_durable_failure"
			result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_objective_closeout_failure_review_handoff", "host_ui_objective_closeout_failure_handoff", "compensation_not_executed")
		case view.ReadyForHostDurableApply:
			result.Status = "ready_for_host_durable_apply_handoff"
			result.DisplayStage = "intermediate"
			result.DisplaySteps = productionAdapterObjectiveCloseoutHostUIHandoffSteps(result.DisplayStage, false)
			result.ReadyForHostDurableApply = true
			result.IntermediateDisplay = true
			result.FinalDisplay = false
			result.FailureReviewDisplay = false
			result.NextHostAction = "host_may_apply_objective_closeout"
			result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_host_durable_apply_handoff", "host_ui_objective_closeout_intermediate_handoff")
		default:
			result.Status = "ready_for_objective_closeout_display_handoff"
			result.DisplayStage = firstControlToken(result.DisplayStage, "review")
			result.DisplaySteps = productionAdapterObjectiveCloseoutHostUIHandoffSteps(result.DisplayStage, false)
			result.IntermediateDisplay = true
			result.NextHostAction = firstNextHostAction(view.NextHostAction, "review_objective_closeout_display")
			result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_objective_closeout_display_handoff", "host_ui_objective_closeout_review_handoff")
		}
	}
	return result.Normalize()
}

func CloneProductionAdapterObjectiveCloseoutHostUIHandoff(in ProductionAdapterObjectiveCloseoutHostUIHandoff) ProductionAdapterObjectiveCloseoutHostUIHandoff {
	out := in
	out.DisplaySteps = cloneStringSlice(in.DisplaySteps)
	out.DisplaySections = cloneStringSlice(in.DisplaySections)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.DisplayProgressMissingInputs = cloneMissingInputs(in.DisplayProgressMissingInputs)
	out.DisplayProgressBlockedReasons = cloneStringSlice(in.DisplayProgressBlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (h ProductionAdapterObjectiveCloseoutHostUIHandoff) Clone() ProductionAdapterObjectiveCloseoutHostUIHandoff {
	return CloneProductionAdapterObjectiveCloseoutHostUIHandoff(h)
}

func (h ProductionAdapterObjectiveCloseoutHostUIHandoff) Normalize() ProductionAdapterObjectiveCloseoutHostUIHandoff {
	out := CloneProductionAdapterObjectiveCloseoutHostUIHandoff(h)
	unsafe := productionAdapterObjectiveCloseoutHostUIHandoffUnsafeOutput(out)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = normalizeControlToken(out.Status)
	out.Mode = normalizeControlToken(out.Mode)
	out.DisplayState = normalizeControlToken(out.DisplayState)
	out.DisplayStage = normalizeControlToken(out.DisplayStage)
	if out.Mode == "" {
		out.Mode = "production_adapter_objective_closeout_host_ui_handoff"
	}
	if out.DisplayState == "" {
		out.DisplayState = "blocked"
	}
	if out.DisplayStage == "" {
		out.DisplayStage = "blocked"
	}
	out.DisplaySteps = normalizeControlTokenList(out.DisplaySteps)
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	if out.RunnerEffect == "" {
		out.RunnerEffect = "none"
	}
	if out.PromptEffect == "" {
		out.PromptEffect = "none"
	}
	out.HostUIHandoffRef = normalizeOneDisplaySafeRef(out.HostUIHandoffRef)
	out.PrimaryDisplayRef = normalizeOneDisplaySafeRef(out.PrimaryDisplayRef)
	out.HostViewRef = normalizeOneDisplaySafeRef(out.HostViewRef)
	out.DisplayFixtureRef = normalizeOneDisplaySafeRef(out.DisplayFixtureRef)
	out.FailureReviewPacketRef = normalizeOneDisplaySafeRef(out.FailureReviewPacketRef)
	out.FailureReviewFixtureRef = normalizeOneDisplaySafeRef(out.FailureReviewFixtureRef)
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
	out.FailureRef = normalizeOneDisplaySafeRef(out.FailureRef)
	out.CompensationRef = normalizeOneDisplaySafeRef(out.CompensationRef)
	out.HostDurableApplyConfirmationRef = normalizeOneDisplaySafeRef(out.HostDurableApplyConfirmationRef)
	out.CompletionAuditResultBindingRef = normalizeOneDisplaySafeRef(out.CompletionAuditResultBindingRef)
	out.AdapterRef = normalizeOneDisplaySafeRef(out.AdapterRef)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.BlockedReasons = normalizeControlTokenList(out.BlockedReasons)
	out.DisplayProgressMissingInputs = normalizeMissingInputs(out.DisplayProgressMissingInputs)
	out.DisplayProgressBlockedReasons = normalizeControlTokenList(out.DisplayProgressBlockedReasons)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	if !out.Available {
		out.Status = "unavailable"
		out.DisplayStage = "blocked"
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
	if unsafe || out.RawOutputLoaded {
		out.RawOutputLoaded = true
		out.Status = "blocked"
		out.DisplayStage = "blocked"
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
		out.HostUIHandoffRef != "" &&
		out.PrimaryDisplayRef != "" &&
		out.HostViewRef != "" &&
		out.ObjectiveCloseoutPacketRef != "" &&
		out.ObjectiveRef != "" &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.RawOutputLoaded
	out.ReadyForHostDurableApply = out.ReadyForHostDurableApply &&
		out.ReadyForHostDisplay &&
		out.Status == "ready_for_host_durable_apply_handoff" &&
		out.DisplayState == "durable_apply_pending" &&
		out.DisplayStage == "intermediate" &&
		out.ObjectiveCloseoutHandoffRef != "" &&
		out.HostDurableApplyConfirmationRef != "" &&
		!out.FinalDisplay &&
		!out.RawOutputLoaded
	out.ReadyForFailureReview = out.ReadyForFailureReview &&
		out.ReadyForHostDisplay &&
		out.Status == "ready_for_objective_closeout_failure_review_handoff" &&
		out.DisplayState == "durable_failure_review" &&
		out.DisplayStage == "failure_review" &&
		out.FailureReviewPacketRef != "" &&
		out.FailureRef != "" &&
		!out.FinalDisplay &&
		!out.RawOutputLoaded
	out.ReadyForCompensationReview = out.ReadyForCompensationReview &&
		out.ReadyForFailureReview &&
		out.CompensationRef != "" &&
		!out.RawOutputLoaded
	out.ReadyForObjectiveReturn = out.ReadyForObjectiveReturn &&
		out.ReadyForHostDisplay &&
		out.Status == "ready_for_objective_return_handoff" &&
		out.DisplayState == "objective_return_final" &&
		out.DisplayStage == "final" &&
		out.ObjectiveLifecycleClosed &&
		out.ObjectiveSatisfied &&
		out.ObjectiveCloseoutReadbackRef != "" &&
		out.ObservedAppliedDurableEventRef != "" &&
		out.ObservedAppliedObjectiveStateRef != "" &&
		!out.IntermediateDisplay &&
		!out.RawOutputLoaded
	out.IntermediateDisplay = out.IntermediateDisplay &&
		out.ReadyForHostDisplay &&
		(out.DisplayStage == "intermediate" || out.DisplayStage == "failure_review" || out.DisplayStage == "review") &&
		!out.FinalDisplay &&
		!out.RawOutputLoaded
	out.FinalDisplay = out.FinalDisplay &&
		out.ReadyForObjectiveReturn &&
		out.DisplayStage == "final" &&
		!out.IntermediateDisplay &&
		!out.RawOutputLoaded
	out.FailureReviewDisplay = out.FailureReviewDisplay &&
		out.ReadyForFailureReview &&
		out.DisplayStage == "failure_review" &&
		!out.FinalDisplay &&
		!out.RawOutputLoaded
	return out
}

func productionAdapterObjectiveCloseoutHostUIHandoffBlock(result ProductionAdapterObjectiveCloseoutHostUIHandoff, failure FailureClass, reason string, missing MissingInput, next NextHostAction) ProductionAdapterObjectiveCloseoutHostUIHandoff {
	result.Status = "blocked"
	result.DisplayStage = "blocked"
	result.ReadyForHostDisplay = false
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
	result.NextHostAction = firstNextHostAction(next, result.NextHostAction)
	return result
}

func productionAdapterObjectiveCloseoutHostUIHandoffStage(view ProductionAdapterObjectiveCloseoutHostView) string {
	switch {
	case view.FinalDisplay || view.ReadyForObjectiveReturn || view.DisplayState == "objective_return_final":
		return "final"
	case view.FailureReviewDisplay || view.ReadyForFailureReview || view.DisplayState == "durable_failure_review":
		return "failure_review"
	case view.IntermediateDisplay || view.ReadyForHostDurableApply || view.DisplayState == "durable_apply_pending":
		return "intermediate"
	case view.ReadyForHostDisplay:
		return "review"
	default:
		return "blocked"
	}
}

func productionAdapterObjectiveCloseoutHostUIHandoffSteps(stage string, includeCompensation bool) []string {
	switch normalizeControlToken(stage) {
	case "final":
		return []string{"objective_closeout_summary", "durable_readback", "objective_return"}
	case "failure_review":
		steps := []string{"objective_closeout_summary", "durable_readback", "failure_review"}
		if includeCompensation {
			steps = append(steps, "compensation_review")
		}
		return steps
	case "intermediate":
		return []string{"objective_closeout_summary", "durable_handoff", "host_durable_apply"}
	case "review":
		return []string{"objective_closeout_summary", "closeout_review"}
	default:
		return []string{"objective_closeout_summary"}
	}
}

func productionAdapterObjectiveCloseoutHostUIHandoffDisplaySections(stage string) []string {
	return append(productionAdapterObjectiveCloseoutDisplaySections(), "host_ui_handoff", normalizeControlToken(stage))
}

func productionAdapterObjectiveCloseoutHostUIHandoffPrimaryDisplayRef(
	view ProductionAdapterObjectiveCloseoutHostView,
	fixture ProductionAdapterObjectiveCloseoutBlackboxFixture,
	fixtureProvided bool,
	failureReview ProductionAdapterObjectiveCloseoutFailureReviewPacket,
	failureReviewProvided bool,
	failureFixture ProductionAdapterObjectiveCloseoutFailureReviewBlackboxFixture,
	failureFixtureProvided bool,
) DisplaySafeRef {
	if view.ReadyForFailureReview || view.DisplayState == "durable_failure_review" {
		if failureFixtureProvided && failureFixture.FixtureRef != "" {
			return failureFixture.FixtureRef
		}
		if failureReviewProvided && failureReview.FailureReviewPacketRef != "" {
			return failureReview.FailureReviewPacketRef
		}
	}
	if fixtureProvided && fixture.FixtureRef != "" {
		return fixture.FixtureRef
	}
	return view.HostViewRef
}

func productionAdapterObjectiveCloseoutHostUIHandoffMissingInputs(
	view ProductionAdapterObjectiveCloseoutHostView,
	fixture ProductionAdapterObjectiveCloseoutBlackboxFixture,
	fixtureProvided bool,
	failureReview ProductionAdapterObjectiveCloseoutFailureReviewPacket,
	failureReviewProvided bool,
	failureFixture ProductionAdapterObjectiveCloseoutFailureReviewBlackboxFixture,
	failureFixtureProvided bool,
) []MissingInput {
	switch {
	case view.ReadyForFailureReview && failureFixtureProvided:
		return cloneMissingInputs(failureFixture.MissingInputs)
	case view.ReadyForFailureReview && failureReviewProvided:
		return cloneMissingInputs(failureReview.MissingInputs)
	case fixtureProvided:
		return cloneMissingInputs(fixture.MissingInputs)
	case view.ReadyForHostDurableApply && view.DisplayState == "durable_apply_pending":
		return productionAdapterObjectiveCloseoutMissingWithout(view.MissingInputs, "host:objective_closeout_readback")
	default:
		return cloneMissingInputs(view.MissingInputs)
	}
}

func productionAdapterObjectiveCloseoutHostUIHandoffBlockedReasons(
	view ProductionAdapterObjectiveCloseoutHostView,
	fixture ProductionAdapterObjectiveCloseoutBlackboxFixture,
	fixtureProvided bool,
	failureReview ProductionAdapterObjectiveCloseoutFailureReviewPacket,
	failureReviewProvided bool,
	failureFixture ProductionAdapterObjectiveCloseoutFailureReviewBlackboxFixture,
	failureFixtureProvided bool,
) []string {
	switch {
	case view.ReadyForFailureReview && failureFixtureProvided:
		return cloneStringSlice(failureFixture.BlockedReasons)
	case view.ReadyForFailureReview && failureReviewProvided:
		return cloneStringSlice(failureReview.BlockedReasons)
	case fixtureProvided:
		return cloneStringSlice(fixture.BlockedReasons)
	case view.ReadyForHostDurableApply && view.DisplayState == "durable_apply_pending":
		return productionAdapterObjectiveCloseoutControlTokensWithout(view.BlockedReasons, "objective_closeout_readback_not_ready")
	default:
		return cloneStringSlice(view.BlockedReasons)
	}
}

func productionAdapterObjectiveCloseoutHostUIHandoffProgressMissingInputs(view ProductionAdapterObjectiveCloseoutHostView, fixture ProductionAdapterObjectiveCloseoutBlackboxFixture, fixtureProvided bool) []MissingInput {
	if fixtureProvided && len(fixture.DisplayProgressMissingInputs) > 0 {
		return cloneMissingInputs(fixture.DisplayProgressMissingInputs)
	}
	return cloneMissingInputs(view.DisplayProgressMissingInputs)
}

func productionAdapterObjectiveCloseoutHostUIHandoffProgressBlockedReasons(view ProductionAdapterObjectiveCloseoutHostView, fixture ProductionAdapterObjectiveCloseoutBlackboxFixture, fixtureProvided bool) []string {
	if fixtureProvided && len(fixture.DisplayProgressBlockedReasons) > 0 {
		return cloneStringSlice(fixture.DisplayProgressBlockedReasons)
	}
	return cloneStringSlice(view.DisplayProgressBlockedReasons)
}

type productionAdapterObjectiveCloseoutHostUIHandoffMismatch struct {
	reason  string
	missing MissingInput
}

func productionAdapterObjectiveCloseoutHostUIHandoffMismatches(
	view ProductionAdapterObjectiveCloseoutHostView,
	fixture ProductionAdapterObjectiveCloseoutBlackboxFixture,
	fixtureProvided bool,
	failureReview ProductionAdapterObjectiveCloseoutFailureReviewPacket,
	failureReviewProvided bool,
	failureFixture ProductionAdapterObjectiveCloseoutFailureReviewBlackboxFixture,
	failureFixtureProvided bool,
) []productionAdapterObjectiveCloseoutHostUIHandoffMismatch {
	var out []productionAdapterObjectiveCloseoutHostUIHandoffMismatch
	if fixtureProvided {
		out = append(out, productionAdapterObjectiveCloseoutHostUIHandoffRefMismatch(view.HostViewRef, fixture.HostViewRef, "closeout_ui_handoff_fixture_host_view_ref_mismatch", "host:objective_closeout_host_view")...)
		out = append(out, productionAdapterObjectiveCloseoutHostUIHandoffRefMismatch(view.ObjectiveCloseoutPacketRef, fixture.ObjectiveCloseoutPacketRef, "closeout_ui_handoff_fixture_packet_ref_mismatch", "host:objective_closeout_packet")...)
		out = append(out, productionAdapterObjectiveCloseoutHostUIHandoffRefMismatch(view.ObjectiveCloseoutReadbackRef, fixture.ObjectiveCloseoutReadbackRef, "closeout_ui_handoff_fixture_readback_ref_mismatch", "host:objective_closeout_readback")...)
		if view.DisplayState != "" && fixture.DisplayState != "" && view.DisplayState != fixture.DisplayState {
			out = append(out, productionAdapterObjectiveCloseoutHostUIHandoffMismatch{reason: "closeout_ui_handoff_fixture_display_state_mismatch", missing: "host:objective_closeout_display_state"})
		}
	}
	if failureReviewProvided {
		out = append(out, productionAdapterObjectiveCloseoutHostUIHandoffRefMismatch(view.HostViewRef, failureReview.HostViewRef, "closeout_ui_handoff_failure_review_host_view_ref_mismatch", "host:objective_closeout_host_view")...)
		out = append(out, productionAdapterObjectiveCloseoutHostUIHandoffRefMismatch(view.ObjectiveCloseoutPacketRef, failureReview.ObjectiveCloseoutPacketRef, "closeout_ui_handoff_failure_review_packet_ref_mismatch", "host:objective_closeout_packet")...)
		out = append(out, productionAdapterObjectiveCloseoutHostUIHandoffRefMismatch(view.FailureRef, failureReview.FailureRef, "closeout_ui_handoff_failure_ref_mismatch", "host:objective_closeout_failure_ref")...)
	}
	if failureFixtureProvided {
		out = append(out, productionAdapterObjectiveCloseoutHostUIHandoffRefMismatch(view.HostViewRef, failureFixture.HostViewRef, "closeout_ui_handoff_failure_fixture_host_view_ref_mismatch", "host:objective_closeout_host_view")...)
		out = append(out, productionAdapterObjectiveCloseoutHostUIHandoffRefMismatch(failureReview.FailureReviewPacketRef, failureFixture.FailureReviewPacketRef, "closeout_ui_handoff_failure_fixture_packet_ref_mismatch", "host:objective_closeout_failure_review_packet")...)
		out = append(out, productionAdapterObjectiveCloseoutHostUIHandoffRefMismatch(view.FailureRef, failureFixture.FailureRef, "closeout_ui_handoff_failure_fixture_ref_mismatch", "host:objective_closeout_failure_ref")...)
	}
	return out
}

func productionAdapterObjectiveCloseoutHostUIHandoffRefMismatch(left DisplaySafeRef, right DisplaySafeRef, reason string, missing MissingInput) []productionAdapterObjectiveCloseoutHostUIHandoffMismatch {
	left = normalizeOneDisplaySafeRef(left)
	right = normalizeOneDisplaySafeRef(right)
	if left != "" && right != "" && left != right {
		return []productionAdapterObjectiveCloseoutHostUIHandoffMismatch{{reason: reason, missing: missing}}
	}
	return nil
}

func productionAdapterObjectiveCloseoutHostUIHandoffBoundaries(groups ...[]Boundary) []Boundary {
	return MergeBoundaries(
		[]Boundary{
			"production_adapter_objective_closeout_host_ui_handoff",
			"objective_closeout_host_ui_handoff_projection_only",
			"host_ui_closeout_handoff",
			"host_cli_objective_closeout_display",
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

func productionAdapterObjectiveCloseoutHostUIHandoffUnsafe(
	input ProductionAdapterObjectiveCloseoutHostUIHandoffInput,
	view ProductionAdapterObjectiveCloseoutHostView,
	fixture ProductionAdapterObjectiveCloseoutBlackboxFixture,
	fixtureProvided bool,
	failureReview ProductionAdapterObjectiveCloseoutFailureReviewPacket,
	failureReviewProvided bool,
	failureFixture ProductionAdapterObjectiveCloseoutFailureReviewBlackboxFixture,
	failureFixtureProvided bool,
) bool {
	return input.RawOutputLoaded ||
		displaySafeRefRejected(input.HostUIHandoffRef) ||
		productionAdapterObjectiveCloseoutHostViewUnsafeOutput(view) ||
		(fixtureProvided && productionAdapterObjectiveCloseoutBlackboxFixtureUnsafeOutput(fixture)) ||
		(failureReviewProvided && productionAdapterObjectiveCloseoutFailureReviewPacketUnsafeOutput(failureReview)) ||
		(failureFixtureProvided && productionAdapterObjectiveCloseoutFailureReviewBlackboxFixtureUnsafeOutput(failureFixture))
}

func productionAdapterObjectiveCloseoutHostUIHandoffUnsafeOutput(input ProductionAdapterObjectiveCloseoutHostUIHandoff) bool {
	return displaySafeRefRejected(input.HostUIHandoffRef) ||
		displaySafeRefRejected(input.PrimaryDisplayRef) ||
		displaySafeRefRejected(input.HostViewRef) ||
		displaySafeRefRejected(input.DisplayFixtureRef) ||
		displaySafeRefRejected(input.FailureReviewPacketRef) ||
		displaySafeRefRejected(input.FailureReviewFixtureRef) ||
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
		displaySafeRefRejected(input.FailureRef) ||
		displaySafeRefRejected(input.CompensationRef) ||
		displaySafeRefRejected(input.HostDurableApplyConfirmationRef) ||
		displaySafeRefRejected(input.CompletionAuditResultBindingRef) ||
		displaySafeRefRejected(input.AdapterRef) ||
		input.RawOutputLoaded
}

func productionAdapterObjectiveCloseoutBlackboxFixtureEmpty(fixture ProductionAdapterObjectiveCloseoutBlackboxFixture) bool {
	return !fixture.Projected &&
		!fixture.Available &&
		fixture.Status == "" &&
		fixture.Mode == "" &&
		fixture.FixtureRef == "" &&
		fixture.HostViewRef == "" &&
		fixture.ObjectiveCloseoutPacketRef == "" &&
		fixture.ObjectiveRef == "" &&
		len(fixture.MissingInputs) == 0 &&
		len(fixture.BlockedReasons) == 0 &&
		len(fixture.Boundaries) == 0 &&
		fixture.NextHostAction == "" &&
		!fixture.RawOutputLoaded
}

func productionAdapterObjectiveCloseoutFailureReviewBlackboxFixtureEmpty(fixture ProductionAdapterObjectiveCloseoutFailureReviewBlackboxFixture) bool {
	return !fixture.Projected &&
		!fixture.Available &&
		fixture.Status == "" &&
		fixture.Mode == "" &&
		fixture.FixtureRef == "" &&
		fixture.FailureReviewPacketRef == "" &&
		fixture.HostViewRef == "" &&
		fixture.ObjectiveCloseoutPacketRef == "" &&
		fixture.ObjectiveRef == "" &&
		len(fixture.MissingInputs) == 0 &&
		len(fixture.BlockedReasons) == 0 &&
		len(fixture.Boundaries) == 0 &&
		fixture.NextHostAction == "" &&
		!fixture.RawOutputLoaded
}

func unavailableProductionAdapterObjectiveCloseoutHostUIHandoff() ProductionAdapterObjectiveCloseoutHostUIHandoff {
	return ProductionAdapterObjectiveCloseoutHostUIHandoff{
		ContractVersion: ContractVersion,
		Projected:       true,
		Available:       false,
		Status:          "unavailable",
		Mode:            "production_adapter_objective_closeout_host_ui_handoff",
		DisplayState:    "blocked",
		DisplayStage:    "blocked",
		DisplaySteps:    []string{"objective_closeout_summary"},
		DisplaySections: productionAdapterObjectiveCloseoutHostUIHandoffDisplaySections("blocked"),
		Boundaries: []Boundary{
			"production_adapter_objective_closeout_host_ui_handoff",
			"objective_closeout_host_ui_handoff_projection_only",
			"host_ui_closeout_handoff",
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
