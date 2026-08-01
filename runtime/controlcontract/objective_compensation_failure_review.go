package controlcontract

type ObjectiveCompensationFailureReviewPacketInput struct {
	CompensationFailureReviewPacketRef DisplaySafeRef                         `json:"compensation_failure_review_packet_ref,omitempty"`
	HostReviewRef                      DisplaySafeRef                         `json:"host_review_ref,omitempty"`
	Result                             ObjectiveCompensationExecutionResult   `json:"result,omitempty"`
	Readback                           ObjectiveCompensationExecutionReadback `json:"readback,omitempty"`
	EvidenceRefs                       []EvidenceRef                          `json:"evidence_refs,omitempty"`
	Boundaries                         []Boundary                             `json:"boundaries,omitempty"`
	RawOutputLoaded                    bool                                   `json:"raw_output_loaded"`
}

type ObjectiveCompensationFailureReviewPacket struct {
	ContractVersion                    string           `json:"contract_version,omitempty"`
	Projected                          bool             `json:"projected"`
	Available                          bool             `json:"available"`
	Status                             HostActionStatus `json:"status,omitempty"`
	Mode                               string           `json:"mode,omitempty"`
	ReadyForHostDisplay                bool             `json:"ready_for_host_display"`
	ReadyForCloseoutFailureReview      bool             `json:"ready_for_closeout_failure_review"`
	ReadyForCompensationFailureReview  bool             `json:"ready_for_compensation_failure_review"`
	CompensationFailureRecorded        bool             `json:"compensation_failure_recorded"`
	ResidualRiskRecorded               bool             `json:"residual_risk_recorded"`
	CompensationFailureReviewPacketRef DisplaySafeRef   `json:"compensation_failure_review_packet_ref,omitempty"`
	HostReviewRef                      DisplaySafeRef   `json:"host_review_ref,omitempty"`
	FailureReviewPacketRef             DisplaySafeRef   `json:"failure_review_packet_ref,omitempty"`
	CompensationResultRef              DisplaySafeRef   `json:"compensation_result_ref,omitempty"`
	CompensationReadbackRef            DisplaySafeRef   `json:"compensation_readback_ref,omitempty"`
	CompensationRequestRef             DisplaySafeRef   `json:"compensation_request_ref,omitempty"`
	ObjectiveRef                       DisplaySafeRef   `json:"objective_ref,omitempty"`
	FailureRef                         DisplaySafeRef   `json:"failure_ref,omitempty"`
	CompensationRef                    DisplaySafeRef   `json:"compensation_ref,omitempty"`
	AppliedCompensationRef             DisplaySafeRef   `json:"applied_compensation_ref,omitempty"`
	ResidualRiskRef                    DisplaySafeRef   `json:"residual_risk_ref,omitempty"`
	ObservedResidualRiskRef            DisplaySafeRef   `json:"observed_residual_risk_ref,omitempty"`
	ExecutorRef                        DisplaySafeRef   `json:"executor_ref,omitempty"`
	HostCompensationRunRef             DisplaySafeRef   `json:"host_compensation_run_ref,omitempty"`
	ObservedHostRunRef                 DisplaySafeRef   `json:"observed_host_run_ref,omitempty"`
	EvidenceRefs                       []EvidenceRef    `json:"evidence_refs,omitempty"`
	MissingInputs                      []MissingInput   `json:"missing_inputs,omitempty"`
	BlockedReasons                     []string         `json:"blocked_reasons,omitempty"`
	FailureClass                       FailureClass     `json:"failure_class,omitempty"`
	Boundaries                         []Boundary       `json:"boundaries,omitempty"`
	NextHostAction                     NextHostAction   `json:"next_host_action,omitempty"`
	RunnerEffect                       string           `json:"runner_effect,omitempty"`
	PromptEffect                       string           `json:"prompt_effect,omitempty"`
	CoreExecutionExecuted              bool             `json:"core_execution_executed"`
	RunnerDispatched                   bool             `json:"runner_dispatched"`
	ToolExecuted                       bool             `json:"tool_executed"`
	WorkflowDispatched                 bool             `json:"workflow_dispatched"`
	SchedulerApplied                   bool             `json:"scheduler_applied"`
	InstallerExecuted                  bool             `json:"installer_executed"`
	StoreMutationExecuted              bool             `json:"store_mutation_executed"`
	CompensationExecutedByCore         bool             `json:"compensation_executed_by_core"`
	RawOutputLoaded                    bool             `json:"raw_output_loaded"`
}

func BuildObjectiveCompensationFailureReviewPacket(input ObjectiveCompensationFailureReviewPacketInput) ObjectiveCompensationFailureReviewPacket {
	rawResultEmpty := objectiveCompensationExecutionResultEmpty(input.Result)
	rawReadbackEmpty := objectiveCompensationExecutionReadbackEmpty(input.Readback)
	if rawResultEmpty && rawReadbackEmpty {
		return unavailableObjectiveCompensationFailureReviewPacket()
	}
	result := input.Result.Normalize()
	readback := input.Readback.Normalize()
	if rawResultEmpty && !rawReadbackEmpty {
		result = readback.Result
	}
	resultReadyForFailureReview := result.ReadyForCloseoutReview && result.HostCompensationFailed && result.ResidualRiskRef != ""
	readbackReadyForFailureReview := !rawReadbackEmpty && readback.ReadyForCloseoutReview && readback.ResidualRiskRecorded && readback.ObservedResidualRiskRef != ""
	packet := ObjectiveCompensationFailureReviewPacket{
		ContractVersion:                    ContractVersion,
		Projected:                          true,
		Available:                          result.Available || readback.Available,
		Status:                             HostActionBlocked,
		Mode:                               "objective_compensation_failure_review_packet",
		CompensationFailureReviewPacketRef: normalizeOneDisplaySafeRef(input.CompensationFailureReviewPacketRef),
		HostReviewRef:                      normalizeOneDisplaySafeRef(input.HostReviewRef),
		FailureReviewPacketRef:             firstDisplaySafeRef(readback.FailureReviewPacketRef, result.FailureReviewPacketRef),
		CompensationResultRef:              firstDisplaySafeRef(readback.CompensationResultRef, result.CompensationResultRef),
		CompensationReadbackRef:            readback.CompensationReadbackRef,
		CompensationRequestRef:             firstDisplaySafeRef(readback.CompensationRequestRef, result.CompensationRequestRef),
		ObjectiveRef:                       firstDisplaySafeRef(readback.ObjectiveRef, result.ObjectiveRef),
		FailureRef:                         firstDisplaySafeRef(readback.FailureRef, result.FailureRef),
		CompensationRef:                    firstDisplaySafeRef(readback.CompensationRef, result.CompensationRef),
		AppliedCompensationRef:             firstDisplaySafeRef(readback.AppliedCompensationRef, result.AppliedCompensationRef),
		ResidualRiskRef:                    firstDisplaySafeRef(readback.ResidualRiskRef, result.ResidualRiskRef),
		ObservedResidualRiskRef:            readback.ObservedResidualRiskRef,
		ExecutorRef:                        result.ExecutorRef,
		HostCompensationRunRef:             firstDisplaySafeRef(readback.HostCompensationRunRef, result.HostCompensationRunRef),
		ObservedHostRunRef:                 readback.ObservedHostRunRef,
		CompensationFailureRecorded:        result.HostCompensationFailed,
		ResidualRiskRecorded:               result.ResidualRiskRef != "" || readback.ResidualRiskRecorded,
		EvidenceRefs:                       MergeEvidenceRefs(result.CompensationEvidenceRefs, readback.ReadbackEvidenceRefs, input.EvidenceRefs),
		FailureClass:                       firstFailureClass(result.FailureClass, readback.FailureClass, FailureNone),
		Boundaries:                         objectiveCompensationFailureReviewPacketBoundaries(input.Boundaries, result.Boundaries, readback.Boundaries),
		NextHostAction:                     "review_compensation_failure_closeout",
		RunnerEffect:                       "none",
		PromptEffect:                       "none",
		RawOutputLoaded:                    input.RawOutputLoaded || result.RawOutputLoaded || readback.RawOutputLoaded,
	}
	if objectiveCompensationFailureReviewPacketInputUnsafe(input) {
		packet = objectiveCompensationFailureReviewPacketBlock(packet, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		packet.RawOutputLoaded = true
		return packet.Normalize()
	}
	if packet.CompensationFailureReviewPacketRef == "" {
		packet = objectiveCompensationFailureReviewPacketBlock(packet, FailureEvidenceMissing, "compensation_failure_review_packet_ref_missing", "host:compensation_failure_review_packet_ref", "provide_compensation_failure_review_refs")
	}
	if packet.HostReviewRef == "" {
		packet = objectiveCompensationFailureReviewPacketBlock(packet, FailureEvidenceMissing, "compensation_failure_host_review_ref_missing", "host:compensation_failure_review_ref", "provide_compensation_failure_review_refs")
	}
	if packet.FailureReviewPacketRef == "" {
		packet = objectiveCompensationFailureReviewPacketBlock(packet, FailureEvidenceMissing, "source_failure_review_packet_ref_missing", "host:source_failure_review_packet_ref", "provide_compensation_failure_review_refs")
	}
	if packet.CompensationResultRef == "" {
		packet = objectiveCompensationFailureReviewPacketBlock(packet, FailureEvidenceMissing, "compensation_result_ref_missing", "host:compensation_result_ref", "provide_compensation_result_ref")
	}
	if !rawReadbackEmpty && !readback.ReadyForCloseoutReview {
		packet = objectiveCompensationFailureReviewPacketBlock(packet, firstFailureClass(readback.FailureClass, FailureEvidenceMissing), "compensation_readback_not_ready_for_failure_review", "host:compensation_readback", firstNextHostAction(readback.NextHostAction, "review_compensation_readback"))
	}
	if !result.ReadyForCloseoutReview && !readbackReadyForFailureReview {
		packet = objectiveCompensationFailureReviewPacketBlock(packet, firstFailureClass(result.FailureClass, FailureEvidenceMissing), "compensation_result_not_ready_for_failure_review", "host:compensation_result", firstNextHostAction(result.NextHostAction, "review_compensation_execution_result"))
	}
	if !result.HostCompensationFailed && !readback.ResidualRiskRecorded {
		packet = objectiveCompensationFailureReviewPacketBlock(packet, FailureVerificationFailed, "compensation_failure_not_present", "host:compensation_failure_or_residual_risk", "continue_closeout_review")
	}
	compensationFailurePresent := result.HostCompensationFailed || readback.ResidualRiskRecorded
	if compensationFailurePresent && packet.FailureRef == "" {
		packet = objectiveCompensationFailureReviewPacketBlock(packet, FailureEvidenceMissing, "compensation_failure_ref_missing", "host:compensation_failure_ref", "record_residual_risk")
	}
	if compensationFailurePresent && packet.ResidualRiskRef == "" {
		packet = objectiveCompensationFailureReviewPacketBlock(packet, FailureEvidenceMissing, "compensation_residual_risk_ref_missing", "host:residual_risk_ref", "record_residual_risk")
	}
	if compensationFailurePresent && !rawReadbackEmpty && readback.ResidualRiskRecorded && packet.ObservedResidualRiskRef == "" {
		packet = objectiveCompensationFailureReviewPacketBlock(packet, FailureEvidenceMissing, "observed_residual_risk_ref_missing", "host:observed_residual_risk_ref", "provide_compensation_readback")
	}
	if len(packet.MissingInputs) == 0 && len(packet.BlockedReasons) == 0 && (resultReadyForFailureReview || readbackReadyForFailureReview) {
		packet.Status = HostActionReviewRequired
		packet.ReadyForHostDisplay = true
		packet.ReadyForCloseoutFailureReview = true
		packet.ReadyForCompensationFailureReview = true
		packet.NextHostAction = "review_compensation_failure_closeout"
		packet.Boundaries = AppendBoundaries(packet.Boundaries, "ready_for_compensation_failure_closeout_review")
	}
	return packet.Normalize()
}

func CloneObjectiveCompensationFailureReviewPacket(in ObjectiveCompensationFailureReviewPacket) ObjectiveCompensationFailureReviewPacket {
	out := in
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (p ObjectiveCompensationFailureReviewPacket) Clone() ObjectiveCompensationFailureReviewPacket {
	return CloneObjectiveCompensationFailureReviewPacket(p)
}

func (p ObjectiveCompensationFailureReviewPacket) Normalize() ObjectiveCompensationFailureReviewPacket {
	out := CloneObjectiveCompensationFailureReviewPacket(p)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.Mode = normalizeControlToken(out.Mode)
	if out.Mode == "" {
		out.Mode = "objective_compensation_failure_review_packet"
	}
	out.CompensationFailureReviewPacketRef = normalizeOneDisplaySafeRef(out.CompensationFailureReviewPacketRef)
	out.HostReviewRef = normalizeOneDisplaySafeRef(out.HostReviewRef)
	out.FailureReviewPacketRef = normalizeOneDisplaySafeRef(out.FailureReviewPacketRef)
	out.CompensationResultRef = normalizeOneDisplaySafeRef(out.CompensationResultRef)
	out.CompensationReadbackRef = normalizeOneDisplaySafeRef(out.CompensationReadbackRef)
	out.CompensationRequestRef = normalizeOneDisplaySafeRef(out.CompensationRequestRef)
	out.ObjectiveRef = normalizeOneDisplaySafeRef(out.ObjectiveRef)
	out.FailureRef = normalizeOneDisplaySafeRef(out.FailureRef)
	out.CompensationRef = normalizeOneDisplaySafeRef(out.CompensationRef)
	out.AppliedCompensationRef = normalizeOneDisplaySafeRef(out.AppliedCompensationRef)
	out.ResidualRiskRef = normalizeOneDisplaySafeRef(out.ResidualRiskRef)
	out.ObservedResidualRiskRef = normalizeOneDisplaySafeRef(out.ObservedResidualRiskRef)
	out.ExecutorRef = normalizeOneDisplaySafeRef(out.ExecutorRef)
	out.HostCompensationRunRef = normalizeOneDisplaySafeRef(out.HostCompensationRunRef)
	out.ObservedHostRunRef = normalizeOneDisplaySafeRef(out.ObservedHostRunRef)
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
	objectiveCompensationClearCoreEffects(&out.CoreExecutionExecuted, &out.RunnerDispatched, &out.ToolExecuted, &out.WorkflowDispatched, &out.SchedulerApplied, &out.InstallerExecuted, &out.StoreMutationExecuted, &out.CompensationExecutedByCore)
	if !out.Available {
		out.Status = HostActionNotReady
		out.ReadyForHostDisplay = false
		out.ReadyForCloseoutFailureReview = false
		out.ReadyForCompensationFailureReview = false
	}
	if out.RawOutputLoaded {
		out.Status = HostActionBlocked
		out.ReadyForHostDisplay = false
		out.ReadyForCloseoutFailureReview = false
		out.ReadyForCompensationFailureReview = false
		if out.FailureClass == FailureNone {
			out.FailureClass = FailureEvidenceWeak
		}
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		if out.NextHostAction == "" {
			out.NextHostAction = "provide_display_safe_refs"
		}
	}
	out.CompensationFailureRecorded = out.CompensationFailureRecorded && !out.RawOutputLoaded
	out.ResidualRiskRecorded = out.ResidualRiskRecorded && out.ResidualRiskRef != "" && !out.RawOutputLoaded
	out.ReadyForHostDisplay = out.ReadyForHostDisplay &&
		out.Status == HostActionReviewRequired &&
		out.Available &&
		out.CompensationFailureReviewPacketRef != "" &&
		out.HostReviewRef != "" &&
		out.FailureReviewPacketRef != "" &&
		out.CompensationResultRef != "" &&
		out.ObjectiveRef != "" &&
		out.FailureRef != "" &&
		out.CompensationRef != "" &&
		out.ResidualRiskRef != "" &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.RawOutputLoaded
	out.ReadyForCloseoutFailureReview = out.ReadyForCloseoutFailureReview &&
		out.ReadyForHostDisplay &&
		out.CompensationFailureRecorded &&
		out.ResidualRiskRecorded
	out.ReadyForCompensationFailureReview = out.ReadyForCompensationFailureReview && out.ReadyForCloseoutFailureReview
	return out
}

func objectiveCompensationFailureReviewPacketBlock(result ObjectiveCompensationFailureReviewPacket, failure FailureClass, reason string, missing MissingInput, next NextHostAction) ObjectiveCompensationFailureReviewPacket {
	result.Status = HostActionBlocked
	result.ReadyForHostDisplay = false
	result.ReadyForCloseoutFailureReview = false
	result.ReadyForCompensationFailureReview = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = firstNextHostAction(next, result.NextHostAction)
	result.Boundaries = AppendBoundaries(result.Boundaries, "objective_compensation_failure_review_packet_blocked")
	return result
}

func objectiveCompensationFailureReviewPacketBoundaries(groups ...[]Boundary) []Boundary {
	all := append([][]Boundary{{
		"objective_compensation_failure_review_packet",
		"host_owned_compensation_failure_review",
		"compensation_failure_to_closeout_failure_review",
		"residual_risk_review_required",
		"display_safe_refs_only",
		"no_compensation_execution_by_core",
		"no_runner_dispatch",
		"no_tool_execution_by_core",
		"no_workflow_dispatch_by_core",
		"no_scheduler_apply_by_core",
		"no_install_apply_by_core",
		"no_store_mutation_by_core",
	}}, groups...)
	return MergeBoundaries(all...)
}

func objectiveCompensationFailureReviewPacketInputUnsafe(input ObjectiveCompensationFailureReviewPacketInput) bool {
	return displaySafeRefRejected(input.CompensationFailureReviewPacketRef) ||
		displaySafeRefRejected(input.HostReviewRef) ||
		objectiveCompensationExecutionResultOutputUnsafe(input.Result) ||
		objectiveCompensationExecutionReadbackOutputUnsafe(input.Readback) ||
		evidenceRefsRejected(input.EvidenceRefs) ||
		input.RawOutputLoaded
}

func objectiveCompensationExecutionResultEmpty(result ObjectiveCompensationExecutionResult) bool {
	return !result.Projected &&
		!result.Available &&
		result.Status == "" &&
		result.CompensationResultRef == "" &&
		result.CompensationRequestRef == "" &&
		result.FailureReviewPacketRef == "" &&
		result.ObjectiveRef == "" &&
		result.FailureRef == "" &&
		result.CompensationRef == "" &&
		result.ResidualRiskRef == "" &&
		len(result.MissingInputs) == 0 &&
		len(result.BlockedReasons) == 0 &&
		len(result.Boundaries) == 0 &&
		result.NextHostAction == "" &&
		!result.RawOutputLoaded
}

func objectiveCompensationExecutionReadbackEmpty(readback ObjectiveCompensationExecutionReadback) bool {
	return !readback.Projected &&
		!readback.Available &&
		readback.Status == "" &&
		readback.CompensationReadbackRef == "" &&
		readback.CompensationResultRef == "" &&
		readback.CompensationRequestRef == "" &&
		readback.FailureReviewPacketRef == "" &&
		readback.ObjectiveRef == "" &&
		readback.FailureRef == "" &&
		readback.CompensationRef == "" &&
		readback.ResidualRiskRef == "" &&
		readback.ObservedResidualRiskRef == "" &&
		len(readback.MissingInputs) == 0 &&
		len(readback.BlockedReasons) == 0 &&
		len(readback.Boundaries) == 0 &&
		readback.NextHostAction == "" &&
		!readback.RawOutputLoaded &&
		objectiveCompensationExecutionResultEmpty(readback.Result)
}

func objectiveCompensationExecutionReadbackOutputUnsafe(input ObjectiveCompensationExecutionReadback) bool {
	return objectiveCompensationExecutionResultOutputUnsafe(input.Result) ||
		displaySafeRefRejected(input.CompensationReadbackRef) ||
		displaySafeRefRejected(input.CompensationResultRef) ||
		displaySafeRefRejected(input.CompensationRequestRef) ||
		displaySafeRefRejected(input.FailureReviewPacketRef) ||
		displaySafeRefRejected(input.ObjectiveRef) ||
		displaySafeRefRejected(input.FailureRef) ||
		displaySafeRefRejected(input.CompensationRef) ||
		displaySafeRefRejected(input.AppliedCompensationRef) ||
		displaySafeRefRejected(input.ObservedCompensationRef) ||
		displaySafeRefRejected(input.HostCompensationRunRef) ||
		displaySafeRefRejected(input.ObservedHostRunRef) ||
		displaySafeRefRejected(input.ResidualRiskRef) ||
		displaySafeRefRejected(input.ObservedResidualRiskRef) ||
		evidenceRefsRejected(input.ReadbackEvidenceRefs) ||
		input.RawOutputLoaded
}

func unavailableObjectiveCompensationFailureReviewPacket() ObjectiveCompensationFailureReviewPacket {
	return ObjectiveCompensationFailureReviewPacket{
		ContractVersion: ContractVersion,
		Projected:       true,
		Available:       false,
		Status:          HostActionNotReady,
		Mode:            "objective_compensation_failure_review_packet",
		Boundaries: objectiveCompensationFailureReviewPacketBoundaries([]Boundary{
			"objective_compensation_failure_review_unavailable",
		}),
		RunnerEffect:   "none",
		PromptEffect:   "none",
		NextHostAction: "provide_compensation_result_or_readback",
	}
}
