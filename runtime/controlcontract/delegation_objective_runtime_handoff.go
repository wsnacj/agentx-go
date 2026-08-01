package controlcontract

type DelegationObjectiveRuntimeHandoffInput struct {
	HandoffRef      DisplaySafeRef                `json:"handoff_ref,omitempty"`
	Run             ObjectiveRun                  `json:"run,omitempty"`
	ParentMerge     DelegationWorkerParentMerge   `json:"parent_merge,omitempty"`
	FailureReview   DelegationWorkerFailureReview `json:"failure_review,omitempty"`
	EvidenceRefs    []EvidenceRef                 `json:"evidence_refs,omitempty"`
	DecisionBasis   []DisplaySafeRef              `json:"decision_basis,omitempty"`
	Boundaries      []Boundary                    `json:"boundaries,omitempty"`
	RawOutputLoaded bool                          `json:"raw_output_loaded"`
}

type DelegationObjectiveRuntimeHandoff struct {
	ContractVersion          string                          `json:"contract_version,omitempty"`
	Projected                bool                            `json:"projected"`
	Status                   HostActionStatus                `json:"status,omitempty"`
	ReadyForRuntimeLoopInput bool                            `json:"ready_for_runtime_loop_input"`
	ReadyForFailureReview    bool                            `json:"ready_for_failure_review"`
	HandoffRef               DisplaySafeRef                  `json:"handoff_ref,omitempty"`
	Run                      ObjectiveRun                    `json:"run,omitempty"`
	ParentMerge              DelegationWorkerParentMerge     `json:"parent_merge,omitempty"`
	FailureReview            DelegationWorkerFailureReview   `json:"failure_review,omitempty"`
	RuntimeLoopInput         ObjectiveRuntimeLoopInput       `json:"runtime_loop_input,omitempty"`
	LedgerPatch              AttemptLedgerPatch              `json:"ledger_patch,omitempty"`
	Verification             ObjectiveVerificationGateResult `json:"verification,omitempty"`
	Observations             []Observation                   `json:"observations,omitempty"`
	EvidenceRefs             []EvidenceRef                   `json:"evidence_refs,omitempty"`
	FailureClass             FailureClass                    `json:"failure_class,omitempty"`
	BlockedReasons           []string                        `json:"blocked_reasons,omitempty"`
	MissingInputs            []MissingInput                  `json:"missing_inputs,omitempty"`
	DecisionBasis            []DisplaySafeRef                `json:"decision_basis,omitempty"`
	Boundaries               []Boundary                      `json:"boundaries,omitempty"`
	NextHostAction           NextHostAction                  `json:"next_host_action,omitempty"`
	LoopEffect               string                          `json:"loop_effect,omitempty"`
	RunnerEffect             string                          `json:"runner_effect,omitempty"`
	PromptEffect             string                          `json:"prompt_effect,omitempty"`
	RuntimeEffect            string                          `json:"runtime_effect,omitempty"`
	CoreExecutionExecuted    bool                            `json:"core_execution_executed"`
	RunnerDispatched         bool                            `json:"runner_dispatched"`
	WorkerDispatched         bool                            `json:"worker_dispatched"`
	StoreMutationExecuted    bool                            `json:"store_mutation_executed"`
	RawOutputLoaded          bool                            `json:"raw_output_loaded"`
}

func BuildDelegationObjectiveRuntimeHandoff(input DelegationObjectiveRuntimeHandoffInput) DelegationObjectiveRuntimeHandoff {
	run := input.Run.Normalize()
	parentMerge := input.ParentMerge.Normalize()
	failureReview := input.FailureReview.Normalize()
	result := DelegationObjectiveRuntimeHandoff{
		ContractVersion: ContractVersion,
		Projected:       true,
		Status:          HostActionBlocked,
		HandoffRef:      normalizeOneDisplaySafeRef(input.HandoffRef),
		Run:             run,
		ParentMerge:     parentMerge,
		FailureReview:   failureReview,
		EvidenceRefs: MergeEvidenceRefs(
			input.EvidenceRefs,
			parentMerge.EvidenceRefs,
			parentMerge.ParentLedgerPatch.EvidenceRefs,
			parentMerge.ParentVerification.EvidenceRefs,
			failureReview.EvidenceRefs,
		),
		FailureClass: FailureNone,
		DecisionBasis: normalizeDisplaySafeRefs(append(
			[]DisplaySafeRef{
				"delegation:objective_runtime_handoff",
				"delegation:parent_merge_consumer",
			},
			input.DecisionBasis...,
		)),
		Boundaries: MergeBoundaries(
			[]Boundary{
				"delegation_objective_runtime_handoff",
				"delegation_worker_parent_merge_consumer",
				"runtime_loop_input_projection_only",
				"worker_output_not_fact",
				"worker_result_requires_verification",
				"no_worker_dispatch",
				"no_runner_dispatch",
				"no_runtime_adapter_execution",
				"no_tool_execution",
				"no_workflow_dispatch",
				"no_store_mutation_by_core",
				"projection_only",
			},
			input.Boundaries,
			parentMerge.Boundaries,
		),
		NextHostAction: "provide_delegation_objective_runtime_handoff",
		LoopEffect:     "state_projection_only",
		RunnerEffect:   "none",
		PromptEffect:   "none",
		RuntimeEffect:  "none",
		RawOutputLoaded: input.RawOutputLoaded ||
			run.RawOutputLoaded ||
			parentMerge.RawOutputLoaded ||
			failureReview.RawOutputLoaded ||
			delegationObjectiveRuntimeHandoffUnsafe(input, run, parentMerge, failureReview),
	}
	if result.RawOutputLoaded {
		return delegationObjectiveRuntimeHandoffBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs", "raw_output_not_allowed")
	}
	if result.HandoffRef == "" {
		result = delegationObjectiveRuntimeHandoffAccumulate(result, FailureConfigMissing, "handoff_ref_missing", "host:delegation_objective_runtime_handoff_ref", "provide_delegation_objective_runtime_handoff_ref", "delegation_objective_runtime_handoff_ref_missing")
	}
	if failureReview.ReadyForFailureReview {
		result.Status = HostActionRecorded
		result.ReadyForFailureReview = true
		result.FailureClass = firstFailureClass(failureReview.FailureClass, FailureVerificationFailed)
		result.NextHostAction = "review_delegation_worker_failure"
		result.Boundaries = AppendBoundaries(result.Boundaries, "delegation_worker_failure_review_handoff_ready")
		return result.Normalize()
	}
	if !delegationObjectiveRuntimeRunReady(run) {
		result = delegationObjectiveRuntimeHandoffAccumulate(result, FailureConfigMissing, "objective_run_missing_or_not_managed", "host:objective_run", "provide_objective_run", "delegation_objective_run_required")
	}
	if !parentMerge.ReadyForParentMerge {
		result = delegationObjectiveRuntimeHandoffAccumulate(result, firstFailureClass(parentMerge.FailureClass, FailureEvidenceMissing), "delegation_parent_merge_not_ready", "host:delegation_worker_parent_merge", firstNextHostAction(parentMerge.NextHostAction, "review_delegation_worker_parent_merge"), "delegation_parent_merge_not_ready_for_runtime_loop")
	}
	if parentMerge.ParentLedgerPatch.LedgerRef == "" {
		result = delegationObjectiveRuntimeHandoffAccumulate(result, FailureEvidenceMissing, "parent_ledger_patch_missing", "host:delegation_parent_ledger_patch", "provide_delegation_parent_ledger_patch", "delegation_parent_ledger_patch_missing")
	}
	if run.Ledger.LedgerRef != "" && parentMerge.ParentLedgerPatch.LedgerRef != "" && run.Ledger.LedgerRef != parentMerge.ParentLedgerPatch.LedgerRef {
		result = delegationObjectiveRuntimeHandoffAccumulate(result, FailureVerificationFailed, "parent_ledger_ref_mismatch", "host:objective_ledger_ref", "review_delegation_parent_merge_binding", "delegation_parent_ledger_ref_mismatch")
	}
	if run.Frame.ID != "" && parentMerge.ParentFrame.ID != "" && run.Frame.ID != parentMerge.ParentFrame.ID {
		result = delegationObjectiveRuntimeHandoffAccumulate(result, FailureVerificationFailed, "parent_objective_ref_mismatch", "host:objective_frame", "review_delegation_parent_merge_binding", "delegation_parent_objective_ref_mismatch")
	}
	if len(result.MissingInputs) > 0 || len(result.BlockedReasons) > 0 {
		return result.Normalize()
	}
	result.Status = HostActionRecorded
	result.ReadyForRuntimeLoopInput = true
	result.LedgerPatch = parentMerge.ParentLedgerPatch
	result.Verification = parentMerge.ParentVerification
	result.Observations = cloneObservations(parentMerge.Observations)
	result.EvidenceRefs = MergeEvidenceRefs(result.EvidenceRefs, result.Verification.EvidenceRefs, objectiveObservationEvidenceRefs(result.Observations))
	result.RuntimeLoopInput = ObjectiveRuntimeLoopInput{
		Run:          run,
		LedgerPatch:  result.LedgerPatch,
		Verification: result.Verification,
		Observations: cloneObservations(result.Observations),
		EvidenceRefs: cloneEvidenceRefs(result.EvidenceRefs),
		Boundaries: AppendBoundaries(
			result.Boundaries,
			"delegation_objective_runtime_handoff_consumed",
			"ready_for_objective_runtime_loop",
		),
	}
	result.NextHostAction = "run_objective_runtime_loop_step"
	result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_objective_runtime_loop", "delegation_parent_merge_ready_for_runtime_loop")
	return result.Normalize()
}

func CloneDelegationObjectiveRuntimeHandoff(in DelegationObjectiveRuntimeHandoff) DelegationObjectiveRuntimeHandoff {
	out := in
	out.Run = in.Run.Clone()
	out.ParentMerge = in.ParentMerge.Clone()
	out.FailureReview = in.FailureReview.Clone()
	out.RuntimeLoopInput = cloneDelegationObjectiveRuntimeLoopInput(in.RuntimeLoopInput)
	out.LedgerPatch = in.LedgerPatch.Clone()
	out.Verification = in.Verification.Clone()
	out.Observations = cloneObservations(in.Observations)
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.DecisionBasis = cloneDisplaySafeRefs(in.DecisionBasis)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (h DelegationObjectiveRuntimeHandoff) Clone() DelegationObjectiveRuntimeHandoff {
	return CloneDelegationObjectiveRuntimeHandoff(h)
}

func (h DelegationObjectiveRuntimeHandoff) Normalize() DelegationObjectiveRuntimeHandoff {
	out := CloneDelegationObjectiveRuntimeHandoff(h)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.HandoffRef = normalizeOneDisplaySafeRef(out.HandoffRef)
	out.Run = out.Run.Normalize()
	out.ParentMerge = out.ParentMerge.Normalize()
	out.FailureReview = out.FailureReview.Normalize()
	out.LedgerPatch = out.LedgerPatch.Normalize()
	out.Verification = out.Verification.Normalize()
	out.Observations = normalizeObservations(out.Observations)
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.BlockedReasons = normalizeControlTokenList(out.BlockedReasons)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.DecisionBasis = normalizeDisplaySafeRefs(out.DecisionBasis)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	out.LoopEffect = normalizeControlToken(out.LoopEffect)
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	out.RuntimeEffect = normalizeControlToken(out.RuntimeEffect)
	if out.LoopEffect == "" {
		out.LoopEffect = "state_projection_only"
	}
	if out.RunnerEffect == "" {
		out.RunnerEffect = "none"
	}
	if out.PromptEffect == "" {
		out.PromptEffect = "none"
	}
	if out.RuntimeEffect == "" {
		out.RuntimeEffect = "none"
	}
	if out.RawOutputLoaded {
		out.Status = HostActionReviewRequired
		out.ReadyForRuntimeLoopInput = false
		out.ReadyForFailureReview = false
		if out.FailureClass == FailureNone {
			out.FailureClass = FailureEvidenceWeak
		}
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		out.NextHostAction = "provide_display_safe_refs"
	}
	out.ReadyForRuntimeLoopInput = out.ReadyForRuntimeLoopInput &&
		out.Status == HostActionRecorded &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.RawOutputLoaded &&
		out.RuntimeLoopInput.Run.FullRun &&
		!objectiveRuntimeLoopLedgerEmpty(out.RuntimeLoopInput.LedgerPatch) &&
		!objectiveRuntimeLoopVerificationGateEmpty(out.RuntimeLoopInput.Verification)
	out.ReadyForFailureReview = out.ReadyForFailureReview &&
		out.Status == HostActionRecorded &&
		out.FailureReview.ReadyForFailureReview &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.RawOutputLoaded
	return out
}

func delegationObjectiveRuntimeHandoffBlock(result DelegationObjectiveRuntimeHandoff, failure FailureClass, reason string, missing MissingInput, next NextHostAction, boundary Boundary) DelegationObjectiveRuntimeHandoff {
	result = delegationObjectiveRuntimeHandoffAccumulate(result, failure, reason, missing, next, boundary)
	result.Status = HostActionReviewRequired
	result.ReadyForRuntimeLoopInput = false
	result.ReadyForFailureReview = false
	return result.Normalize()
}

func delegationObjectiveRuntimeHandoffAccumulate(result DelegationObjectiveRuntimeHandoff, failure FailureClass, reason string, missing MissingInput, next NextHostAction, boundary Boundary) DelegationObjectiveRuntimeHandoff {
	result.Status = HostActionBlocked
	result.ReadyForRuntimeLoopInput = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	result.Boundaries = AppendBoundaries(result.Boundaries, boundary)
	return result
}

func delegationObjectiveRuntimeRunReady(run ObjectiveRun) bool {
	run = run.Normalize()
	return !objectiveRunInputEmpty(run) &&
		run.FullRun &&
		run.Activation == ActivationManaged &&
		run.Frame.ID != "" &&
		run.Ledger.LedgerRef != "" &&
		run.Budget.BudgetRef != "" &&
		run.Budget.Limit > 0
}

func delegationObjectiveRuntimeHandoffUnsafe(input DelegationObjectiveRuntimeHandoffInput, run ObjectiveRun, parentMerge DelegationWorkerParentMerge, failureReview DelegationWorkerFailureReview) bool {
	return input.RawOutputLoaded ||
		displaySafeRefRejected(input.HandoffRef) ||
		displaySafeRefSliceRejected(input.DecisionBasis) ||
		evidenceRefRejected(input.EvidenceRefs) ||
		run.RawOutputLoaded ||
		parentMerge.RawOutputLoaded ||
		failureReview.RawOutputLoaded ||
		delegationWorkerParentMergeOutputUnsafe(parentMerge) ||
		delegationWorkerFailureReviewUnsafe(DelegationWorkerFailureReviewInput{
			Request:          failureReview.Request,
			Invocations:      failureReview.Invocations,
			ParentMerges:     failureReview.ParentMerges,
			WorkerAttempts:   failureReview.WorkerAttempts,
			FailureReviewRef: failureReview.FailureReviewRef,
			FailureRef:       failureReview.FailureRef,
			CompensationRef:  failureReview.CompensationRef,
			ConflictRefs:     failureReview.ConflictRefs,
			EvidenceRefs:     failureReview.EvidenceRefs,
			RawOutputLoaded:  failureReview.RawOutputLoaded,
		}, failureReview.Invocations, failureReview.ParentMerges, failureReview.WorkerAttempts)
}

func cloneDelegationObjectiveRuntimeLoopInput(in ObjectiveRuntimeLoopInput) ObjectiveRuntimeLoopInput {
	out := in
	out.Run = in.Run.Clone()
	out.Frame = in.Frame.Clone()
	out.Ledger = in.Ledger.Clone()
	out.LedgerPatch = in.LedgerPatch.Clone()
	out.Budget = in.Budget.Clone()
	out.Approval = in.Approval.Clone()
	out.PolicyRefs = cloneDisplaySafeRefs(in.PolicyRefs)
	out.Strategies = cloneStrategyCandidates(in.Strategies)
	out.Verification = in.Verification.Clone()
	out.VerificationResult = in.VerificationResult.Clone()
	out.Observations = cloneObservations(in.Observations)
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}
