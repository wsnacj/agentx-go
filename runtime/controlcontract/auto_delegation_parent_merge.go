package controlcontract

type AutoDelegationParentMergeDecision string

const (
	AutoDelegationParentMergeAccept        AutoDelegationParentMergeDecision = "accept"
	AutoDelegationParentMergePartial       AutoDelegationParentMergeDecision = "partial"
	AutoDelegationParentMergeRetry         AutoDelegationParentMergeDecision = "retry"
	AutoDelegationParentMergeAlternatePath AutoDelegationParentMergeDecision = "alternate_path"
	AutoDelegationParentMergePrune         AutoDelegationParentMergeDecision = "prune"
	AutoDelegationParentMergeBlock         AutoDelegationParentMergeDecision = "block"
)

func NormalizeAutoDelegationParentMergeDecision(raw string) AutoDelegationParentMergeDecision {
	switch normalizeEnumToken(raw) {
	case "accept", "accepted", "merge":
		return AutoDelegationParentMergeAccept
	case "partial", "partially_accept", "partial_merge":
		return AutoDelegationParentMergePartial
	case "retry", "try_again":
		return AutoDelegationParentMergeRetry
	case "alternate_path", "alternate", "fallback":
		return AutoDelegationParentMergeAlternatePath
	case "prune", "drop", "discard":
		return AutoDelegationParentMergePrune
	case "block", "blocked", "manual_review", "review":
		return AutoDelegationParentMergeBlock
	default:
		return ""
	}
}

type AutoDelegationParentMergeInput struct {
	HostBridge          AutoDelegationHostBridge        `json:"host_bridge,omitempty"`
	ParentFrame         ObjectiveFrame                  `json:"parent_frame,omitempty"`
	ParentLedgerRef     DisplaySafeRef                  `json:"parent_ledger_ref,omitempty"`
	RequiredEvidence    []EvidenceRef                   `json:"required_evidence,omitempty"`
	ChildResults        []AutoDelegationChildMergeInput `json:"child_results,omitempty"`
	FailureReviewRef    DisplaySafeRef                  `json:"failure_review_ref,omitempty"`
	FailureRef          DisplaySafeRef                  `json:"failure_ref,omitempty"`
	CompensationRef     DisplaySafeRef                  `json:"compensation_ref,omitempty"`
	ConflictRefs        []DisplaySafeRef                `json:"conflict_refs,omitempty"`
	NoProgressThreshold int                             `json:"no_progress_threshold,omitempty"`
	Boundaries          []Boundary                      `json:"boundaries,omitempty"`
	RawOutputLoaded     bool                            `json:"raw_output_loaded"`
}

type AutoDelegationChildMergeInput struct {
	ChildRef                 DisplaySafeRef                             `json:"child_ref,omitempty"`
	Invocation               HostOwnedDelegationWorkerRuntimeInvocation `json:"invocation,omitempty"`
	ParentLedgerRef          DisplaySafeRef                             `json:"parent_ledger_ref,omitempty"`
	WorkerAttemptRef         AttemptRef                                 `json:"worker_attempt_ref,omitempty"`
	MergeRef                 DisplaySafeRef                             `json:"merge_ref,omitempty"`
	MergePolicyRef           DisplaySafeRef                             `json:"merge_policy_ref,omitempty"`
	WorkerObservations       []Observation                              `json:"worker_observations,omitempty"`
	EvidenceRefs             []EvidenceRef                              `json:"evidence_refs,omitempty"`
	RequiredEvidence         []EvidenceRef                              `json:"required_evidence,omitempty"`
	ExpectedObservationKinds []string                                   `json:"expected_observation_kinds,omitempty"`
	AlternatePathRefs        []DisplaySafeRef                           `json:"alternate_path_refs,omitempty"`
	StaleEvidenceRefs        []DisplaySafeRef                           `json:"stale_evidence_refs,omitempty"`
	StaleEvidenceDetected    bool                                       `json:"stale_evidence_detected"`
	PruneRequested           bool                                       `json:"prune_requested"`
	DecisionHint             AutoDelegationParentMergeDecision          `json:"decision_hint,omitempty"`
	Boundaries               []Boundary                                 `json:"boundaries,omitempty"`
	RawOutputLoaded          bool                                       `json:"raw_output_loaded"`
}

type AutoDelegationChildParentMerge struct {
	ContractVersion                  string                                     `json:"contract_version,omitempty"`
	Projected                        bool                                       `json:"projected"`
	Status                           VerificationStatus                         `json:"status,omitempty"`
	Decision                         AutoDelegationParentMergeDecision          `json:"decision,omitempty"`
	ReadyForParentMerge              bool                                       `json:"ready_for_parent_merge"`
	ParentAnswerMayUseChildEvidence  bool                                       `json:"parent_answer_may_use_child_evidence"`
	WorkerResultRequiresVerification bool                                       `json:"worker_result_requires_verification"`
	WorkerOutputAcceptedAsFact       bool                                       `json:"worker_output_accepted_as_fact"`
	ChildRef                         DisplaySafeRef                             `json:"child_ref,omitempty"`
	Child                            AutoDelegationChildTask                    `json:"child,omitempty"`
	Runtime                          AutoDelegationHostChildRuntime             `json:"runtime,omitempty"`
	Input                            AutoDelegationChildMergeInput              `json:"input,omitempty"`
	Invocation                       HostOwnedDelegationWorkerRuntimeInvocation `json:"invocation,omitempty"`
	ParentMerge                      DelegationWorkerParentMerge                `json:"parent_merge,omitempty"`
	AlternatePathRefs                []DisplaySafeRef                           `json:"alternate_path_refs,omitempty"`
	StaleEvidenceRefs                []DisplaySafeRef                           `json:"stale_evidence_refs,omitempty"`
	StaleEvidenceDetected            bool                                       `json:"stale_evidence_detected"`
	WeakEvidenceDetected             bool                                       `json:"weak_evidence_detected"`
	MissingEvidenceDetected          bool                                       `json:"missing_evidence_detected"`
	EvidenceRefs                     []EvidenceRef                              `json:"evidence_refs,omitempty"`
	Observations                     []Observation                              `json:"observations,omitempty"`
	ParentLedgerPatch                AttemptLedgerPatch                         `json:"parent_ledger_patch,omitempty"`
	MissingInputs                    []MissingInput                             `json:"missing_inputs,omitempty"`
	BlockedReasons                   []string                                   `json:"blocked_reasons,omitempty"`
	FailureClass                     FailureClass                               `json:"failure_class,omitempty"`
	Boundaries                       []Boundary                                 `json:"boundaries,omitempty"`
	NextHostAction                   NextHostAction                             `json:"next_host_action,omitempty"`
	RunnerEffect                     string                                     `json:"runner_effect,omitempty"`
	PromptEffect                     string                                     `json:"prompt_effect,omitempty"`
	RuntimeEffect                    string                                     `json:"runtime_effect,omitempty"`
	RawOutputLoaded                  bool                                       `json:"raw_output_loaded"`
}

type AutoDelegationParentMerge struct {
	ContractVersion                  string                            `json:"contract_version,omitempty"`
	Projected                        bool                              `json:"projected"`
	Status                           VerificationStatus                `json:"status,omitempty"`
	Decision                         AutoDelegationParentMergeDecision `json:"decision,omitempty"`
	ReadyForParentMerge              bool                              `json:"ready_for_parent_merge"`
	ParentAnswerMayUseChildEvidence  bool                              `json:"parent_answer_may_use_child_evidence"`
	WorkerResultRequiresVerification bool                              `json:"worker_result_requires_verification"`
	WorkerOutputAcceptedAsFact       bool                              `json:"worker_output_accepted_as_fact"`
	HostBridge                       AutoDelegationHostBridge          `json:"host_bridge,omitempty"`
	Children                         []AutoDelegationChildParentMerge  `json:"children,omitempty"`
	FailureReview                    *DelegationWorkerFailureReview    `json:"failure_review,omitempty"`
	ParentFrame                      ObjectiveFrame                    `json:"parent_frame,omitempty"`
	ParentLedgerRef                  DisplaySafeRef                    `json:"parent_ledger_ref,omitempty"`
	RequiredEvidence                 []EvidenceRef                     `json:"required_evidence,omitempty"`
	ParentLedgerPatch                AttemptLedgerPatch                `json:"parent_ledger_patch,omitempty"`
	AcceptedChildRefs                []DisplaySafeRef                  `json:"accepted_child_refs,omitempty"`
	MergedChildRefs                  []DisplaySafeRef                  `json:"merged_child_refs,omitempty"`
	PartialChildRefs                 []DisplaySafeRef                  `json:"partial_child_refs,omitempty"`
	RetryChildRefs                   []DisplaySafeRef                  `json:"retry_child_refs,omitempty"`
	AlternatePathChildRefs           []DisplaySafeRef                  `json:"alternate_path_child_refs,omitempty"`
	PrunedChildRefs                  []DisplaySafeRef                  `json:"pruned_child_refs,omitempty"`
	BlockedChildRefs                 []DisplaySafeRef                  `json:"blocked_child_refs,omitempty"`
	ConflictRefs                     []DisplaySafeRef                  `json:"conflict_refs,omitempty"`
	StaleEvidenceRefs                []DisplaySafeRef                  `json:"stale_evidence_refs,omitempty"`
	StaleEvidenceDetected            bool                              `json:"stale_evidence_detected"`
	WeakEvidenceDetected             bool                              `json:"weak_evidence_detected"`
	MissingEvidenceDetected          bool                              `json:"missing_evidence_detected"`
	ConflictDetected                 bool                              `json:"conflict_detected"`
	EvidenceRefs                     []EvidenceRef                     `json:"evidence_refs,omitempty"`
	Observations                     []Observation                     `json:"observations,omitempty"`
	MissingInputs                    []MissingInput                    `json:"missing_inputs,omitempty"`
	BlockedReasons                   []string                          `json:"blocked_reasons,omitempty"`
	FailureClass                     FailureClass                      `json:"failure_class,omitempty"`
	Boundaries                       []Boundary                        `json:"boundaries,omitempty"`
	NextHostAction                   NextHostAction                    `json:"next_host_action,omitempty"`
	RunnerEffect                     string                            `json:"runner_effect,omitempty"`
	PromptEffect                     string                            `json:"prompt_effect,omitempty"`
	RuntimeEffect                    string                            `json:"runtime_effect,omitempty"`
	RawOutputLoaded                  bool                              `json:"raw_output_loaded"`
}

func BuildAutoDelegationParentMerge(input AutoDelegationParentMergeInput) AutoDelegationParentMerge {
	hostBridge := input.HostBridge.Normalize()
	parentFrame := autoDelegationParentMergeFrame(input.ParentFrame, hostBridge)
	parentLedgerRef := normalizeOneDisplaySafeRef(input.ParentLedgerRef)
	requiredEvidence := MergeEvidenceRefs(input.RequiredEvidence, hostBridge.PlanReview.Plan.RequiredEvidence)
	result := AutoDelegationParentMerge{
		ContractVersion:   ContractVersion,
		Projected:         true,
		Status:            VerificationBlocked,
		Decision:          AutoDelegationParentMergeBlock,
		HostBridge:        hostBridge,
		ParentFrame:       parentFrame,
		ParentLedgerRef:   parentLedgerRef,
		RequiredEvidence:  requiredEvidence,
		AcceptedChildRefs: cloneDisplaySafeRefs(hostBridge.AcceptedChildRefs),
		ConflictRefs:      normalizeDisplaySafeRefs(input.ConflictRefs),
		FailureClass:      FailureNone,
		Boundaries: MergeBoundaries(
			[]Boundary{
				"auto_delegation_parent_merge",
				"parent_owned_child_merge",
				"worker_as_tool_default",
				"parent_verification_required",
				"child_output_not_fact",
				"projection_only",
				"display_safe_refs_only",
				"no_child_task_spawn_by_core",
				"no_subagent_dispatch_by_core",
				"no_runner_dispatch",
				"no_store_mutation_by_core",
			},
			input.Boundaries,
			hostBridge.Boundaries,
		),
		NextHostAction:                   "review_auto_delegation_parent_merge",
		RunnerEffect:                     "none",
		PromptEffect:                     "none",
		RuntimeEffect:                    "none",
		WorkerResultRequiresVerification: true,
		WorkerOutputAcceptedAsFact:       false,
		RawOutputLoaded:                  input.RawOutputLoaded || hostBridge.RawOutputLoaded,
	}
	if autoDelegationParentMergeUnsafe(input) {
		result.RawOutputLoaded = true
		return autoDelegationParentMergeBlock(result, VerificationReviewRequired, AutoDelegationParentMergeBlock, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs", "raw_output_not_allowed").Normalize()
	}
	if len(result.AcceptedChildRefs) == 0 {
		result = autoDelegationParentMergeBlock(result, VerificationBlocked, AutoDelegationParentMergeBlock, FailureInsufficientInformation, "auto_delegation_accepted_children_missing", "host:auto_delegation_accepted_children", "provide_auto_delegation_host_bridge", "auto_delegation_accepted_children_missing")
	}
	if parentFrame.ID == "" {
		result = autoDelegationParentMergeBlock(result, VerificationBlocked, AutoDelegationParentMergeBlock, FailureConfigMissing, "parent_objective_frame_missing", "host:parent_objective_frame", "provide_parent_objective_frame", "parent_objective_frame_missing")
	}

	runtimesByChild := autoDelegationParentMergeRuntimesByChild(hostBridge.Children)
	resultsByChild := autoDelegationParentMergeChildResultsByChild(input.ChildResults)
	childInputs := autoDelegationParentMergeOrderedChildInputs(result.AcceptedChildRefs, resultsByChild)
	parentLedgerPatch := AttemptLedgerPatch{
		ObjectiveID: parentFrame.ID,
		LedgerRef:   parentLedgerRef,
		Boundaries:  []Boundary{"auto_delegation_parent_merge_ledger_patch"},
	}
	invocations := []HostOwnedDelegationWorkerRuntimeInvocation{}
	parentMerges := []DelegationWorkerParentMerge{}
	workerAttempts := []AttemptSummary{}
	for _, childRef := range result.AcceptedChildRefs {
		childResult, exists := childInputs[childRef]
		runtime := runtimesByChild[childRef]
		childMerge := buildAutoDelegationChildParentMerge(autoDelegationChildParentMergeBuildInput{
			ParentFrame:      parentFrame,
			ParentLedgerRef:  parentLedgerRef,
			RequiredEvidence: requiredEvidence,
			Runtime:          runtime,
			ChildResult:      childResult,
			ResultProvided:   exists,
		})
		result.Children = append(result.Children, childMerge)
		result.EvidenceRefs = MergeEvidenceRefs(result.EvidenceRefs, childMerge.EvidenceRefs)
		result.Observations = append(result.Observations, childMerge.Observations...)
		result.StaleEvidenceRefs = mergeDisplaySafeRefs(result.StaleEvidenceRefs, childMerge.StaleEvidenceRefs)
		result.MissingInputs = AppendMissingInputs(result.MissingInputs, childMerge.MissingInputs...)
		result.BlockedReasons = appendUniqueControlTokens(result.BlockedReasons, childMerge.BlockedReasons)
		result.Boundaries = AppendBoundaries(result.Boundaries, childMerge.Boundaries...)
		result.FailureClass = firstFailureClass(result.FailureClass, childMerge.FailureClass)
		if childMerge.Invocation.WorkerRunRef != "" || childMerge.Invocation.InvocationRef != "" {
			invocations = append(invocations, childMerge.Invocation)
		}
		if childMerge.ParentMerge.Projected {
			parentMerges = append(parentMerges, childMerge.ParentMerge)
		}
		if childMerge.ParentMerge.WorkerAttempt.Ref != "" {
			workerAttempts = append(workerAttempts, childMerge.ParentMerge.WorkerAttempt)
		}
		parentLedgerPatch = objectiveRuntimeLoopMergeLedger(parentLedgerPatch, childMerge.ParentLedgerPatch)
		switch childMerge.Decision {
		case AutoDelegationParentMergeAccept:
			result.MergedChildRefs = appendDisplaySafeRefIfPresent(result.MergedChildRefs, childRef)
		case AutoDelegationParentMergePartial:
			result.PartialChildRefs = appendDisplaySafeRefIfPresent(result.PartialChildRefs, childRef)
		case AutoDelegationParentMergeRetry:
			result.RetryChildRefs = appendDisplaySafeRefIfPresent(result.RetryChildRefs, childRef)
		case AutoDelegationParentMergeAlternatePath:
			result.AlternatePathChildRefs = appendDisplaySafeRefIfPresent(result.AlternatePathChildRefs, childRef)
		case AutoDelegationParentMergePrune:
			result.PrunedChildRefs = appendDisplaySafeRefIfPresent(result.PrunedChildRefs, childRef)
		default:
			result.BlockedChildRefs = appendDisplaySafeRefIfPresent(result.BlockedChildRefs, childRef)
		}
		result.StaleEvidenceDetected = result.StaleEvidenceDetected || childMerge.StaleEvidenceDetected
		result.WeakEvidenceDetected = result.WeakEvidenceDetected || childMerge.WeakEvidenceDetected
		result.MissingEvidenceDetected = result.MissingEvidenceDetected || childMerge.MissingEvidenceDetected
	}
	result.ParentLedgerPatch = parentLedgerPatch.Normalize()
	result.Observations = normalizeObservations(result.Observations)
	result.EvidenceRefs = MergeEvidenceRefs(result.EvidenceRefs, result.ParentLedgerPatch.EvidenceRefs)
	conflictDetected, conflictRefs := delegationWorkerConflictRefs(parentMerges, input.ConflictRefs)
	result.ConflictDetected = conflictDetected
	result.ConflictRefs = mergeDisplaySafeRefs(result.ConflictRefs, conflictRefs)
	if autoDelegationParentMergeNeedsFailureReview(result, invocations, parentMerges, workerAttempts) {
		failureReview := BuildDelegationWorkerFailureReview(DelegationWorkerFailureReviewInput{
			Invocations:         invocations,
			ParentMerges:        parentMerges,
			WorkerAttempts:      workerAttempts,
			FailureReviewRef:    input.FailureReviewRef,
			FailureRef:          input.FailureRef,
			CompensationRef:     input.CompensationRef,
			ConflictRefs:        result.ConflictRefs,
			NoProgressThreshold: input.NoProgressThreshold,
			EvidenceRefs:        result.EvidenceRefs,
			Boundaries: []Boundary{
				"auto_delegation_parent_merge_failure_review",
			},
			RawOutputLoaded: input.RawOutputLoaded,
		})
		result.FailureReview = &failureReview
		result.MissingInputs = AppendMissingInputs(result.MissingInputs, failureReview.MissingInputs...)
		result.BlockedReasons = appendUniqueControlTokens(result.BlockedReasons, failureReview.BlockedReasons)
		result.Boundaries = AppendBoundaries(result.Boundaries, failureReview.Boundaries...)
		result.FailureClass = firstFailureClass(result.FailureClass, failureReview.FailureClass)
		result.EvidenceRefs = MergeEvidenceRefs(result.EvidenceRefs, failureReview.EvidenceRefs)
	}
	result = autoDelegationParentMergeDecide(result)
	return result.Normalize()
}

type autoDelegationChildParentMergeBuildInput struct {
	ParentFrame      ObjectiveFrame
	ParentLedgerRef  DisplaySafeRef
	RequiredEvidence []EvidenceRef
	Runtime          AutoDelegationHostChildRuntime
	ChildResult      AutoDelegationChildMergeInput
	ResultProvided   bool
}

func buildAutoDelegationChildParentMerge(input autoDelegationChildParentMergeBuildInput) AutoDelegationChildParentMerge {
	runtime := input.Runtime.Normalize()
	child := runtime.Child.Normalize()
	childResult := input.ChildResult.Normalize()
	childRef := firstDisplaySafeRef(childResult.ChildRef, runtime.Child.ChildRef)
	result := AutoDelegationChildParentMerge{
		ContractVersion:       ContractVersion,
		Projected:             true,
		Status:                VerificationBlocked,
		Decision:              AutoDelegationParentMergeBlock,
		ChildRef:              childRef,
		Child:                 child,
		Runtime:               runtime,
		Input:                 childResult,
		Invocation:            childResult.Invocation.Normalize(),
		AlternatePathRefs:     mergeDisplaySafeRefs(child.AlternatePathRefs, childResult.AlternatePathRefs),
		StaleEvidenceRefs:     normalizeDisplaySafeRefs(childResult.StaleEvidenceRefs),
		StaleEvidenceDetected: childResult.StaleEvidenceDetected || len(childResult.StaleEvidenceRefs) > 0,
		FailureClass:          FailureNone,
		Boundaries: MergeBoundaries(
			[]Boundary{
				"auto_delegation_child_parent_merge",
				"parent_verification_required",
				"child_output_not_fact",
				"projection_only",
				"display_safe_refs_only",
				"no_worker_dispatch",
				"no_runner_dispatch",
				"no_store_mutation_by_core",
			},
			runtime.Boundaries,
			childResult.Boundaries,
		),
		NextHostAction:                   "review_auto_delegation_child_parent_merge",
		RunnerEffect:                     "none",
		PromptEffect:                     "none",
		RuntimeEffect:                    "none",
		WorkerResultRequiresVerification: true,
		WorkerOutputAcceptedAsFact:       false,
		RawOutputLoaded:                  runtime.RawOutputLoaded || childResult.RawOutputLoaded || childResult.Invocation.RawOutputLoaded,
	}
	if !input.ResultProvided {
		return autoDelegationChildParentMergeBlock(result, VerificationBlocked, AutoDelegationParentMergeRetry, FailureEvidenceMissing, "auto_delegation_child_result_missing", "host:auto_delegation_child_result", "provide_auto_delegation_child_result", "auto_delegation_child_result_missing").Normalize()
	}
	if childResult.PruneRequested || childResult.DecisionHint == AutoDelegationParentMergePrune {
		result.Status = VerificationNotApplicable
		result.Decision = AutoDelegationParentMergePrune
		result.NextHostAction = "prune_auto_delegation_child_result"
		result.Boundaries = AppendBoundaries(result.Boundaries, "auto_delegation_child_pruned")
		return result.Normalize()
	}
	if result.RawOutputLoaded || autoDelegationChildMergeInputUnsafe(childResult) {
		result.RawOutputLoaded = true
		return autoDelegationChildParentMergeBlock(result, VerificationReviewRequired, AutoDelegationParentMergeBlock, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs", "raw_output_not_allowed").Normalize()
	}
	if result.StaleEvidenceDetected {
		decision := AutoDelegationParentMergeRetry
		next := NextHostAction("retry_auto_delegation_child")
		if len(result.AlternatePathRefs) > 0 {
			decision = AutoDelegationParentMergeAlternatePath
			next = "try_auto_delegation_alternate_path"
		}
		return autoDelegationChildParentMergeBlock(result, VerificationPartial, decision, FailureEvidenceWeak, "auto_delegation_child_evidence_stale", "host:fresh_child_evidence", next, "auto_delegation_child_evidence_stale").Normalize()
	}
	if result.Invocation.HostInvocationFailed || result.Invocation.ReadyForFailureReview {
		decision := AutoDelegationParentMergeRetry
		next := NextHostAction("retry_auto_delegation_child")
		if len(result.AlternatePathRefs) > 0 {
			decision = AutoDelegationParentMergeAlternatePath
			next = "try_auto_delegation_alternate_path"
		}
		return autoDelegationChildParentMergeBlock(result, VerificationFailed, decision, firstFailureClass(result.Invocation.FailureClass, FailureVerificationFailed), "auto_delegation_child_invocation_failed", "host:auto_delegation_child_failure", next, "auto_delegation_child_invocation_failed").Normalize()
	}
	requiredEvidence := MergeEvidenceRefs(childResult.RequiredEvidence, child.ExpectedEvidence, input.RequiredEvidence)
	merge := BuildDelegationWorkerParentMerge(DelegationWorkerParentMergeInput{
		Invocation:               result.Invocation,
		ParentFrame:              input.ParentFrame,
		ParentLedgerRef:          firstDisplaySafeRef(childResult.ParentLedgerRef, input.ParentLedgerRef),
		WorkerAttemptRef:         childResult.WorkerAttemptRef,
		MergeRef:                 childResult.MergeRef,
		MergePolicyRef:           firstDisplaySafeRef(childResult.MergePolicyRef, runtime.Request.MergePolicyRef),
		WorkerObservations:       childResult.WorkerObservations,
		EvidenceRefs:             childResult.EvidenceRefs,
		RequiredEvidence:         requiredEvidence,
		ExpectedObservationKinds: childResult.ExpectedObservationKinds,
		Boundaries: []Boundary{
			"auto_delegation_child_parent_merge_delegation_worker_merge",
		},
		RawOutputLoaded: childResult.RawOutputLoaded,
	})
	result.ParentMerge = merge
	result.EvidenceRefs = MergeEvidenceRefs(childResult.EvidenceRefs, merge.EvidenceRefs)
	result.Observations = normalizeObservations(merge.Observations)
	result.ParentLedgerPatch = merge.ParentLedgerPatch.Normalize()
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, merge.MissingInputs...)
	result.BlockedReasons = appendUniqueControlTokens(result.BlockedReasons, merge.BlockedReasons)
	result.Boundaries = AppendBoundaries(result.Boundaries, merge.Boundaries...)
	result.FailureClass = firstFailureClass(result.FailureClass, merge.FailureClass)
	result.WeakEvidenceDetected = autoDelegationMergeWeakEvidence(merge)
	result.MissingEvidenceDetected = autoDelegationMergeMissingEvidence(merge)
	if merge.ReadyForParentMerge {
		result.Status = VerificationSatisfied
		result.Decision = AutoDelegationParentMergeAccept
		result.ReadyForParentMerge = true
		result.ParentAnswerMayUseChildEvidence = true
		result.FailureClass = FailureNone
		result.NextHostAction = "merge_auto_delegation_child_result"
		result.Boundaries = AppendBoundaries(result.Boundaries, "auto_delegation_child_result_verified", "auto_delegation_child_result_ready_for_parent_merge")
		return result.Normalize()
	}
	result.Decision = autoDelegationChildMergeDecision(result)
	result.Status = autoDelegationChildMergeStatus(result)
	result.NextHostAction = autoDelegationChildMergeNextHostAction(result.Decision)
	result.Boundaries = AppendBoundaries(result.Boundaries, Boundary("auto_delegation_child_merge_decision_"+string(result.Decision)))
	return result.Normalize()
}

func (input AutoDelegationChildMergeInput) Normalize() AutoDelegationChildMergeInput {
	out := input
	out.ChildRef = normalizeOneDisplaySafeRef(out.ChildRef)
	out.Invocation = out.Invocation.Normalize()
	out.ParentLedgerRef = normalizeOneDisplaySafeRef(out.ParentLedgerRef)
	out.WorkerAttemptRef = normalizeOneAttemptRef(out.WorkerAttemptRef)
	out.MergeRef = normalizeOneDisplaySafeRef(out.MergeRef)
	out.MergePolicyRef = normalizeOneDisplaySafeRef(out.MergePolicyRef)
	out.WorkerObservations = normalizeObservations(out.WorkerObservations)
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
	out.RequiredEvidence = normalizeEvidenceRefs(out.RequiredEvidence)
	out.ExpectedObservationKinds = normalizeControlTokenList(out.ExpectedObservationKinds)
	out.AlternatePathRefs = normalizeDisplaySafeRefs(out.AlternatePathRefs)
	out.StaleEvidenceRefs = normalizeDisplaySafeRefs(out.StaleEvidenceRefs)
	out.DecisionHint = NormalizeAutoDelegationParentMergeDecision(string(out.DecisionHint))
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	return out
}

func (result AutoDelegationParentMerge) Normalize() AutoDelegationParentMerge {
	out := result
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeVerificationStatus(string(out.Status))
	out.Decision = NormalizeAutoDelegationParentMergeDecision(string(out.Decision))
	out.HostBridge = out.HostBridge.Normalize()
	for i := range out.Children {
		out.Children[i] = out.Children[i].Normalize()
	}
	if out.FailureReview != nil {
		review := out.FailureReview.Normalize()
		out.FailureReview = &review
	}
	out.ParentFrame = out.ParentFrame.Normalize()
	out.ParentLedgerRef = normalizeOneDisplaySafeRef(out.ParentLedgerRef)
	out.RequiredEvidence = normalizeEvidenceRefs(out.RequiredEvidence)
	out.ParentLedgerPatch = out.ParentLedgerPatch.Normalize()
	out.AcceptedChildRefs = normalizeDisplaySafeRefs(out.AcceptedChildRefs)
	out.MergedChildRefs = normalizeDisplaySafeRefs(out.MergedChildRefs)
	out.PartialChildRefs = normalizeDisplaySafeRefs(out.PartialChildRefs)
	out.RetryChildRefs = normalizeDisplaySafeRefs(out.RetryChildRefs)
	out.AlternatePathChildRefs = normalizeDisplaySafeRefs(out.AlternatePathChildRefs)
	out.PrunedChildRefs = normalizeDisplaySafeRefs(out.PrunedChildRefs)
	out.BlockedChildRefs = normalizeDisplaySafeRefs(out.BlockedChildRefs)
	out.ConflictRefs = normalizeDisplaySafeRefs(out.ConflictRefs)
	out.StaleEvidenceRefs = normalizeDisplaySafeRefs(out.StaleEvidenceRefs)
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
	out.Observations = normalizeObservations(out.Observations)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.BlockedReasons = normalizeControlTokenList(out.BlockedReasons)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	out.RuntimeEffect = normalizeControlToken(out.RuntimeEffect)
	if out.Status == VerificationNotEvaluated {
		out.Status = VerificationBlocked
	}
	if out.Decision == "" {
		out.Decision = AutoDelegationParentMergeBlock
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
	out.WorkerResultRequiresVerification = true
	out.WorkerOutputAcceptedAsFact = false
	if out.RawOutputLoaded {
		out.Status = VerificationReviewRequired
		out.ReadyForParentMerge = false
		out.ParentAnswerMayUseChildEvidence = false
		out.FailureClass = firstFailureClass(out.FailureClass, FailureEvidenceWeak)
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		out.NextHostAction = "provide_display_safe_refs"
	}
	return out
}

func (result AutoDelegationChildParentMerge) Normalize() AutoDelegationChildParentMerge {
	out := result
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeVerificationStatus(string(out.Status))
	out.Decision = NormalizeAutoDelegationParentMergeDecision(string(out.Decision))
	out.ChildRef = normalizeOneDisplaySafeRef(out.ChildRef)
	out.Child = out.Child.Normalize()
	out.Runtime = out.Runtime.Normalize()
	out.Input = out.Input.Normalize()
	out.Invocation = out.Invocation.Normalize()
	out.ParentMerge = out.ParentMerge.Normalize()
	out.AlternatePathRefs = normalizeDisplaySafeRefs(out.AlternatePathRefs)
	out.StaleEvidenceRefs = normalizeDisplaySafeRefs(out.StaleEvidenceRefs)
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
	out.Observations = normalizeObservations(out.Observations)
	out.ParentLedgerPatch = out.ParentLedgerPatch.Normalize()
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.BlockedReasons = normalizeControlTokenList(out.BlockedReasons)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	out.RuntimeEffect = normalizeControlToken(out.RuntimeEffect)
	if out.Status == VerificationNotEvaluated {
		out.Status = VerificationBlocked
	}
	if out.Decision == "" {
		out.Decision = AutoDelegationParentMergeBlock
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
	out.WorkerResultRequiresVerification = true
	out.WorkerOutputAcceptedAsFact = false
	if out.RawOutputLoaded {
		out.Status = VerificationReviewRequired
		out.ReadyForParentMerge = false
		out.ParentAnswerMayUseChildEvidence = false
		out.FailureClass = firstFailureClass(out.FailureClass, FailureEvidenceWeak)
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		out.NextHostAction = "provide_display_safe_refs"
	}
	return out
}

func autoDelegationParentMergeDecide(result AutoDelegationParentMerge) AutoDelegationParentMerge {
	if result.ConflictDetected {
		return autoDelegationParentMergeBlock(result, VerificationBlocked, AutoDelegationParentMergeBlock, FailureVerificationFailed, "auto_delegation_child_result_conflict", "host:auto_delegation_conflict_review", "review_auto_delegation_child_conflict", "auto_delegation_child_result_conflict")
	}
	if len(result.MergedChildRefs) == len(result.AcceptedChildRefs) && len(result.AcceptedChildRefs) > 0 && len(result.MissingInputs) == 0 && len(result.BlockedReasons) == 0 {
		result.Status = VerificationSatisfied
		result.Decision = AutoDelegationParentMergeAccept
		result.ReadyForParentMerge = true
		result.ParentAnswerMayUseChildEvidence = true
		result.FailureClass = FailureNone
		result.NextHostAction = "update_objective_controller"
		result.Boundaries = AppendBoundaries(result.Boundaries, "auto_delegation_parent_merge_ready", "auto_delegation_child_evidence_accepted")
		return result
	}
	if len(result.MergedChildRefs) > 0 {
		result.Status = VerificationPartial
		result.Decision = AutoDelegationParentMergePartial
		result.ReadyForParentMerge = true
		result.ParentAnswerMayUseChildEvidence = true
		result.NextHostAction = "update_objective_controller_with_partial_child_evidence"
		result.FailureClass = firstFailureClass(result.FailureClass, FailureEvidenceMissing)
		result.Boundaries = AppendBoundaries(result.Boundaries, "auto_delegation_parent_merge_partial", "partial_child_evidence_ready_for_parent_merge")
		return result
	}
	if len(result.AlternatePathChildRefs) > 0 {
		result.Status = VerificationBlocked
		result.Decision = AutoDelegationParentMergeAlternatePath
		result.NextHostAction = "try_auto_delegation_alternate_path"
		result.FailureClass = firstFailureClass(result.FailureClass, FailureEvidenceMissing)
		result.Boundaries = AppendBoundaries(result.Boundaries, "auto_delegation_alternate_path_required")
		return result
	}
	if len(result.RetryChildRefs) > 0 {
		result.Status = VerificationBlocked
		result.Decision = AutoDelegationParentMergeRetry
		result.NextHostAction = "retry_auto_delegation_children"
		result.FailureClass = firstFailureClass(result.FailureClass, FailureEvidenceMissing)
		result.Boundaries = AppendBoundaries(result.Boundaries, "auto_delegation_retry_required")
		return result
	}
	if len(result.PrunedChildRefs) > 0 && len(result.BlockedChildRefs) == 0 {
		result.Status = VerificationPartial
		result.Decision = AutoDelegationParentMergePrune
		result.NextHostAction = "review_auto_delegation_pruned_children"
		result.FailureClass = firstFailureClass(result.FailureClass, FailureEvidenceMissing)
		result.Boundaries = AppendBoundaries(result.Boundaries, "auto_delegation_children_pruned")
		return result
	}
	result.Status = VerificationBlocked
	result.Decision = AutoDelegationParentMergeBlock
	result.NextHostAction = firstNextHostAction(result.NextHostAction, "review_auto_delegation_parent_merge")
	result.FailureClass = firstFailureClass(result.FailureClass, FailureEvidenceMissing)
	return result
}

func autoDelegationParentMergeBlock(result AutoDelegationParentMerge, status VerificationStatus, decision AutoDelegationParentMergeDecision, failure FailureClass, reason string, missing MissingInput, next NextHostAction, boundary Boundary) AutoDelegationParentMerge {
	result.Status = status
	result.Decision = decision
	result.ReadyForParentMerge = false
	result.ParentAnswerMayUseChildEvidence = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = firstNextHostAction(next, result.NextHostAction)
	result.Boundaries = AppendBoundaries(result.Boundaries, boundary)
	return result
}

func autoDelegationChildParentMergeBlock(result AutoDelegationChildParentMerge, status VerificationStatus, decision AutoDelegationParentMergeDecision, failure FailureClass, reason string, missing MissingInput, next NextHostAction, boundary Boundary) AutoDelegationChildParentMerge {
	result.Status = status
	result.Decision = decision
	result.ReadyForParentMerge = false
	result.ParentAnswerMayUseChildEvidence = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = firstNextHostAction(next, result.NextHostAction)
	result.Boundaries = AppendBoundaries(result.Boundaries, boundary)
	result.MissingEvidenceDetected = failure == FailureEvidenceMissing
	result.WeakEvidenceDetected = failure == FailureEvidenceWeak
	return result
}

func autoDelegationParentMergeFrame(frame ObjectiveFrame, hostBridge AutoDelegationHostBridge) ObjectiveFrame {
	out := frame.Normalize()
	if out.ID == "" {
		out.ID = string(hostBridge.PlanReview.Plan.ParentObjectiveRef)
	}
	if out.ID == "" && len(hostBridge.Children) > 0 {
		out.ID = string(hostBridge.Children[0].Child.ParentObjectiveRef)
	}
	if out.Intensity == "" {
		out.Intensity = IntensityL4DurableLongRun
	}
	if len(out.RequiredEvidence) == 0 {
		out.RequiredEvidence = hostBridge.PlanReview.Plan.RequiredEvidence
	}
	return out.Normalize()
}

func autoDelegationParentMergeRuntimesByChild(children []AutoDelegationHostChildRuntime) map[DisplaySafeRef]AutoDelegationHostChildRuntime {
	out := map[DisplaySafeRef]AutoDelegationHostChildRuntime{}
	for _, child := range children {
		normalized := child.Normalize()
		ref := normalized.Child.ChildRef
		if ref == "" {
			ref = normalized.Binding.ChildRef
		}
		if ref == "" {
			continue
		}
		if _, exists := out[ref]; exists {
			continue
		}
		out[ref] = normalized
	}
	return out
}

func autoDelegationParentMergeChildResultsByChild(results []AutoDelegationChildMergeInput) map[DisplaySafeRef]AutoDelegationChildMergeInput {
	out := map[DisplaySafeRef]AutoDelegationChildMergeInput{}
	for _, result := range results {
		normalized := result.Normalize()
		ref := firstDisplaySafeRef(normalized.ChildRef, normalized.Invocation.SubgoalRef)
		if ref == "" {
			continue
		}
		if _, exists := out[ref]; exists {
			continue
		}
		out[ref] = normalized
	}
	return out
}

func autoDelegationParentMergeOrderedChildInputs(acceptedRefs []DisplaySafeRef, byChild map[DisplaySafeRef]AutoDelegationChildMergeInput) map[DisplaySafeRef]AutoDelegationChildMergeInput {
	out := map[DisplaySafeRef]AutoDelegationChildMergeInput{}
	for _, ref := range normalizeDisplaySafeRefs(acceptedRefs) {
		if result, ok := byChild[ref]; ok {
			out[ref] = result
		}
	}
	return out
}

func autoDelegationParentMergeNeedsFailureReview(result AutoDelegationParentMerge, invocations []HostOwnedDelegationWorkerRuntimeInvocation, parentMerges []DelegationWorkerParentMerge, attempts []AttemptSummary) bool {
	if result.ConflictDetected || len(result.ConflictRefs) > 0 || len(result.BlockedChildRefs) > 0 {
		return true
	}
	for _, invocation := range normalizeDelegationWorkerRuntimeInvocations(invocations) {
		if invocation.HostInvocationFailed || invocation.ReadyForFailureReview {
			return true
		}
	}
	for _, merge := range normalizeDelegationWorkerParentMerges(parentMerges) {
		if merge.FailureClass == FailureVerificationFailed {
			return true
		}
	}
	for _, attempt := range normalizeAttemptSummaries(attempts) {
		if attempt.FailureClass == FailureRepeatedNoProgress {
			return true
		}
	}
	return false
}

func autoDelegationChildMergeDecision(child AutoDelegationChildParentMerge) AutoDelegationParentMergeDecision {
	if child.Input.DecisionHint != "" {
		return child.Input.DecisionHint
	}
	if child.StaleEvidenceDetected {
		if len(child.AlternatePathRefs) > 0 {
			return AutoDelegationParentMergeAlternatePath
		}
		return AutoDelegationParentMergeRetry
	}
	switch child.FailureClass {
	case FailureEvidenceWeak, FailureVerificationFailed:
		if len(child.AlternatePathRefs) > 0 {
			return AutoDelegationParentMergeAlternatePath
		}
		return AutoDelegationParentMergeRetry
	case FailureEvidenceMissing:
		if len(child.AlternatePathRefs) > 0 {
			return AutoDelegationParentMergeAlternatePath
		}
		return AutoDelegationParentMergeRetry
	default:
		return AutoDelegationParentMergeBlock
	}
}

func autoDelegationChildMergeStatus(child AutoDelegationChildParentMerge) VerificationStatus {
	switch child.Decision {
	case AutoDelegationParentMergeRetry, AutoDelegationParentMergeAlternatePath, AutoDelegationParentMergePartial:
		return VerificationPartial
	case AutoDelegationParentMergePrune:
		return VerificationNotApplicable
	default:
		return VerificationBlocked
	}
}

func autoDelegationChildMergeNextHostAction(decision AutoDelegationParentMergeDecision) NextHostAction {
	switch decision {
	case AutoDelegationParentMergeRetry:
		return "retry_auto_delegation_child"
	case AutoDelegationParentMergeAlternatePath:
		return "try_auto_delegation_alternate_path"
	case AutoDelegationParentMergePrune:
		return "prune_auto_delegation_child_result"
	case AutoDelegationParentMergePartial:
		return "review_partial_auto_delegation_child_result"
	default:
		return "review_auto_delegation_child_parent_merge"
	}
}

func autoDelegationMergeWeakEvidence(merge DelegationWorkerParentMerge) bool {
	merge = merge.Normalize()
	return merge.FailureClass == FailureEvidenceWeak ||
		merge.ParentVerification.FailureClass == FailureEvidenceWeak ||
		merge.ResultReview.FailureClass == FailureEvidenceWeak ||
		observationsHaveStrength(merge.Observations, EvidenceWeak)
}

func autoDelegationMergeMissingEvidence(merge DelegationWorkerParentMerge) bool {
	merge = merge.Normalize()
	return merge.FailureClass == FailureEvidenceMissing ||
		merge.ParentVerification.FailureClass == FailureEvidenceMissing ||
		merge.ResultReview.FailureClass == FailureEvidenceMissing ||
		len(merge.MissingInputs) > 0
}

func observationsHaveStrength(observations []Observation, strength EvidenceStrength) bool {
	for _, observation := range normalizeObservations(observations) {
		if observation.Strength == strength {
			return true
		}
		for _, evidence := range observation.EvidenceRefs {
			if evidence.Normalize().Strength == strength {
				return true
			}
		}
	}
	return false
}

func autoDelegationParentMergeUnsafe(input AutoDelegationParentMergeInput) bool {
	if input.RawOutputLoaded ||
		input.HostBridge.RawOutputLoaded ||
		displaySafeRefRejected(input.ParentLedgerRef) ||
		evidenceRefRejected(input.RequiredEvidence) ||
		displaySafeRefRejected(input.FailureReviewRef) ||
		displaySafeRefRejected(input.FailureRef) ||
		displaySafeRefRejected(input.CompensationRef) ||
		displaySafeRefSliceRejected(input.ConflictRefs) {
		return true
	}
	for _, child := range input.ChildResults {
		if autoDelegationChildMergeInputUnsafe(child) {
			return true
		}
	}
	return false
}

func autoDelegationChildMergeInputUnsafe(input AutoDelegationChildMergeInput) bool {
	return input.RawOutputLoaded ||
		input.Invocation.RawOutputLoaded ||
		displaySafeRefRejected(input.ChildRef) ||
		displaySafeRefRejected(input.ParentLedgerRef) ||
		attemptRefRejected(input.WorkerAttemptRef) ||
		displaySafeRefRejected(input.MergeRef) ||
		displaySafeRefRejected(input.MergePolicyRef) ||
		observationSliceUnsafePayload(input.WorkerObservations) ||
		evidenceRefRejected(input.EvidenceRefs) ||
		evidenceRefRejected(input.RequiredEvidence) ||
		displaySafeRefSliceRejected(input.AlternatePathRefs) ||
		displaySafeRefSliceRejected(input.StaleEvidenceRefs)
}
