package controlcontract

type AutoDelegationControllerAction string

const (
	AutoDelegationControllerActionNone            AutoDelegationControllerAction = "none"
	AutoDelegationControllerActionSpawnOnce       AutoDelegationControllerAction = "spawn_once"
	AutoDelegationControllerActionCollectExisting AutoDelegationControllerAction = "collect_existing"
	AutoDelegationControllerActionReplayOnce      AutoDelegationControllerAction = "replay_once"
	AutoDelegationControllerActionPartialCloseout AutoDelegationControllerAction = "partial_closeout"
	AutoDelegationControllerActionCancel          AutoDelegationControllerAction = "cancel"
	AutoDelegationControllerActionBlock           AutoDelegationControllerAction = "block"
	AutoDelegationControllerActionSatisfied       AutoDelegationControllerAction = "satisfied"
)

func NormalizeAutoDelegationControllerAction(raw string) AutoDelegationControllerAction {
	switch normalizeEnumToken(raw) {
	case "", "none":
		return AutoDelegationControllerActionNone
	case "spawn_once", "dispatch_ready_children", "dispatch", "invoke", "invoke_children", "spawn", "fanout":
		return AutoDelegationControllerActionSpawnOnce
	case "collect_existing", "collect_children", "collect", "wait", "wait_children", "wait_for_children":
		return AutoDelegationControllerActionCollectExisting
	case "replay_once", "replay_children", "replay", "retry", "retry_children":
		return AutoDelegationControllerActionReplayOnce
	case "merge_partial", "partial", "partial_closeout", "close_partial":
		return AutoDelegationControllerActionPartialCloseout
	case "cancel", "cancel_children":
		return AutoDelegationControllerActionCancel
	case "complete", "satisfied", "accept":
		return AutoDelegationControllerActionSatisfied
	case "block", "blocked":
		return AutoDelegationControllerActionBlock
	default:
		return AutoDelegationControllerActionBlock
	}
}

type AutoDelegationControllerInput struct {
	HostBridge          AutoDelegationHostBridge                `json:"host_bridge,omitempty"`
	ParentMerge         AutoDelegationParentMerge               `json:"parent_merge,omitempty"`
	AsyncCompletion     AutoDelegationAsyncCompletionProjection `json:"async_completion,omitempty"`
	RequestedDispatch   bool                                    `json:"requested_dispatch"`
	MaxAttemptsPerChild int                                     `json:"max_attempts_per_child,omitempty"`
	ChildAttempts       []AutoDelegationControllerChildState    `json:"child_attempts,omitempty"`
	DecisionBasis       []DisplaySafeRef                        `json:"decision_basis,omitempty"`
	Boundaries          []Boundary                              `json:"boundaries,omitempty"`
	RawOutputLoaded     bool                                    `json:"raw_output_loaded"`
}

type AutoDelegationControllerChildState struct {
	ChildRef         DisplaySafeRef `json:"child_ref,omitempty"`
	Attempts         int            `json:"attempts,omitempty"`
	LastAttemptRef   AttemptRef     `json:"last_attempt_ref,omitempty"`
	LastFailureClass FailureClass   `json:"last_failure_class,omitempty"`
	MissingInputs    []MissingInput `json:"missing_inputs,omitempty"`
	Boundaries       []Boundary     `json:"boundaries,omitempty"`
	RawOutputLoaded  bool           `json:"raw_output_loaded"`
}

type AutoDelegationControllerDecision struct {
	ContractVersion    string                           `json:"contract_version,omitempty"`
	Projected          bool                             `json:"projected"`
	Status             VerificationStatus               `json:"status,omitempty"`
	Action             AutoDelegationControllerAction   `json:"action,omitempty"`
	Ready              bool                             `json:"ready"`
	HostMayDispatch    bool                             `json:"host_may_dispatch"`
	HostMayCollect     bool                             `json:"host_may_collect"`
	HostMayReplay      bool                             `json:"host_may_replay"`
	ReadyForCloseout   bool                             `json:"ready_for_closeout"`
	RequestedDispatch  bool                             `json:"requested_dispatch"`
	RejectedActions    []AutoDelegationControllerAction `json:"rejected_actions,omitempty"`
	HostBridge         AutoDelegationHostBridge         `json:"host_bridge,omitempty"`
	ParentMerge        AutoDelegationParentMerge        `json:"parent_merge,omitempty"`
	AcceptedChildRefs  []DisplaySafeRef                 `json:"accepted_child_refs,omitempty"`
	InvokableChildRefs []DisplaySafeRef                 `json:"invokable_child_refs,omitempty"`
	OpenChildRefs      []DisplaySafeRef                 `json:"open_child_refs,omitempty"`
	CompletedChildRefs []DisplaySafeRef                 `json:"completed_child_refs,omitempty"`
	FailedChildRefs    []DisplaySafeRef                 `json:"failed_child_refs,omitempty"`
	CancelledChildRefs []DisplaySafeRef                 `json:"cancelled_child_refs,omitempty"`
	MergedChildRefs    []DisplaySafeRef                 `json:"merged_child_refs,omitempty"`
	PartialChildRefs   []DisplaySafeRef                 `json:"partial_child_refs,omitempty"`
	RetryChildRefs     []DisplaySafeRef                 `json:"retry_child_refs,omitempty"`
	ReplayChildRefs    []DisplaySafeRef                 `json:"replay_child_refs,omitempty"`
	ExhaustedChildRefs []DisplaySafeRef                 `json:"exhausted_child_refs,omitempty"`
	MissingInputs      []MissingInput                   `json:"missing_inputs,omitempty"`
	BlockedReasons     []string                         `json:"blocked_reasons,omitempty"`
	FailureClass       FailureClass                     `json:"failure_class,omitempty"`
	DecisionBasis      []DisplaySafeRef                 `json:"decision_basis,omitempty"`
	Boundaries         []Boundary                       `json:"boundaries,omitempty"`
	NextHostAction     NextHostAction                   `json:"next_host_action,omitempty"`
	RunnerEffect       string                           `json:"runner_effect,omitempty"`
	PromptEffect       string                           `json:"prompt_effect,omitempty"`
	RuntimeEffect      string                           `json:"runtime_effect,omitempty"`
	RawOutputLoaded    bool                             `json:"raw_output_loaded"`
}

func BuildAutoDelegationControllerDecision(input AutoDelegationControllerInput) AutoDelegationControllerDecision {
	hostBridgeProvided := autoDelegationControllerHostBridgeProvided(input.HostBridge)
	parentMergeProvided := autoDelegationControllerParentMergeProvided(input.ParentMerge)
	asyncCompletionProvided := autoDelegationControllerAsyncCompletionProvided(input.AsyncCompletion)
	hostBridgeInput := input.HostBridge
	if !hostBridgeProvided && asyncCompletionProvided && autoDelegationControllerHostBridgeProvided(input.AsyncCompletion.HostBridge) {
		hostBridgeInput = input.AsyncCompletion.HostBridge
		hostBridgeProvided = true
	}
	hostBridge := hostBridgeInput.Normalize()
	asyncCompletion := input.AsyncCompletion.Normalize()
	if asyncCompletionProvided {
		hostBridge = autoDelegationControllerHostBridgeWithAsyncCompletion(hostBridge, asyncCompletion)
	}
	parentMerge := input.ParentMerge.Normalize()
	childAttempts := autoDelegationControllerNormalizeChildStates(input.ChildAttempts)
	maxAttempts := firstPositiveInt(input.MaxAttemptsPerChild, hostBridge.MaxAttemptsPerChild, DefaultAutoDelegationMaxAttemptsPerChild)
	result := AutoDelegationControllerDecision{
		ContractVersion:    ContractVersion,
		Projected:          true,
		Status:             VerificationBlocked,
		Action:             AutoDelegationControllerActionBlock,
		RequestedDispatch:  input.RequestedDispatch,
		HostBridge:         hostBridge,
		ParentMerge:        parentMerge,
		AcceptedChildRefs:  cloneDisplaySafeRefs(hostBridge.AcceptedChildRefs),
		InvokableChildRefs: mergeDisplaySafeRefs(hostBridge.InvokableChildRefs),
		OpenChildRefs: mergeDisplaySafeRefs(
			hostBridge.ActiveChildRefs,
			hostBridge.QueuedChildRefs,
		),
		CompletedChildRefs: mergeDisplaySafeRefs(hostBridge.CompletedChildRefs),
		FailedChildRefs:    mergeDisplaySafeRefs(hostBridge.FailedChildRefs),
		CancelledChildRefs: mergeDisplaySafeRefs(hostBridge.CancelledChildRefs),
		MergedChildRefs:    mergeDisplaySafeRefs(parentMerge.MergedChildRefs),
		PartialChildRefs:   mergeDisplaySafeRefs(parentMerge.PartialChildRefs),
		RetryChildRefs: mergeDisplaySafeRefs(
			parentMerge.RetryChildRefs,
			parentMerge.AlternatePathChildRefs,
			hostBridge.FailedChildRefs,
		),
		MissingInputs: MergeMissingInputs(
			hostBridge.MissingInputs,
			parentMerge.MissingInputs,
			autoDelegationControllerChildStateMissingInputs(childAttempts),
		),
		BlockedReasons: appendUniqueControlTokens(
			appendUniqueControlTokens(nil, hostBridge.BlockedReasons),
			parentMerge.BlockedReasons,
			autoDelegationControllerChildStateBlockedReasons(childAttempts),
		),
		FailureClass: firstFailureClass(hostBridge.FailureClass, parentMerge.FailureClass, FailureNone),
		DecisionBasis: normalizeDisplaySafeRefs(append(
			[]DisplaySafeRef{"auto_delegation:controller", "auto_delegation:runtime_managed_delegation_controller"},
			input.DecisionBasis...,
		)),
		Boundaries: MergeBoundaries(
			[]Boundary{
				"auto_delegation_controller",
				"runtime_managed_delegation_controller",
				"deterministic_child_lifecycle_reducer",
				"host_owned_child_runtime",
				"parent_verification_required",
				"child_output_not_fact",
				"projection_only",
				"display_safe_refs_only",
				"no_child_task_spawn_by_core",
				"no_subagent_dispatch_by_core",
				"no_runner_dispatch",
				"no_tool_execution",
				"no_store_mutation_by_core",
			},
			input.Boundaries,
			hostBridge.Boundaries,
			parentMerge.Boundaries,
			autoDelegationControllerAsyncCompletionBoundaries(asyncCompletionProvided, asyncCompletion),
			autoDelegationControllerChildStateBoundaries(childAttempts),
		),
		NextHostAction: "review_auto_delegation_controller",
		RunnerEffect:   "none",
		PromptEffect:   "none",
		RuntimeEffect:  "none",
		RawOutputLoaded: input.RawOutputLoaded ||
			hostBridge.RawOutputLoaded ||
			parentMerge.RawOutputLoaded ||
			autoDelegationControllerAsyncCompletionRawOutput(asyncCompletionProvided, asyncCompletion) ||
			autoDelegationControllerChildStateRawOutput(childAttempts),
	}
	if autoDelegationControllerUnsafe(input, childAttempts) || result.RawOutputLoaded {
		result.RawOutputLoaded = true
		return autoDelegationControllerBlock(result, VerificationReviewRequired, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs", "raw_output_not_allowed").Normalize()
	}
	if !hostBridgeProvided {
		return autoDelegationControllerBlock(result, VerificationBlocked, FailureInsufficientInformation, "auto_delegation_host_bridge_missing", "host:auto_delegation_host_bridge", "provide_auto_delegation_host_bridge", "auto_delegation_host_bridge_missing").Normalize()
	}
	if !hostBridge.PlanReview.Ready || !hostBridge.PlanReview.HostMayDispatch {
		return autoDelegationControllerBlock(result, VerificationBlocked, firstFailureClass(hostBridge.PlanReview.FailureClass, FailurePolicyBlocked), "auto_delegation_plan_not_dispatchable", "host:auto_delegation_plan_review", firstNextHostAction(hostBridge.PlanReview.NextHostAction, "review_auto_delegation_plan"), "auto_delegation_plan_not_dispatchable").Normalize()
	}
	if len(result.AcceptedChildRefs) == 0 {
		return autoDelegationControllerBlock(result, VerificationBlocked, FailureInsufficientInformation, "auto_delegation_accepted_children_missing", "host:auto_delegation_accepted_children", "provide_auto_delegation_plan", "auto_delegation_accepted_children_missing").Normalize()
	}

	if parentMergeProvided && parentMerge.ReadyForParentMerge && parentMerge.Decision == AutoDelegationParentMergeAccept && len(parentMerge.RetryChildRefs) == 0 && len(parentMerge.AlternatePathChildRefs) == 0 {
		result = autoDelegationControllerSelect(result, VerificationSatisfied, AutoDelegationControllerActionSatisfied, FailureNone, "update_objective_controller", "auto_delegation_controller_satisfied")
		result.ReadyForCloseout = true
		result.HostMayCollect = true
		return result.Normalize()
	}

	replayable, exhausted := autoDelegationControllerReplayBudget(result.RetryChildRefs, childAttempts, maxAttempts)
	result.ReplayChildRefs = replayable
	result.ExhaustedChildRefs = exhausted
	if len(result.ReplayChildRefs) > 0 {
		if input.RequestedDispatch && len(result.OpenChildRefs) > 0 {
			result.RejectedActions = append(result.RejectedActions, AutoDelegationControllerActionSpawnOnce)
			result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, "auto_delegation_repeated_fanout_rejected")
			result.Boundaries = AppendBoundaries(result.Boundaries, "auto_delegation_repeated_fanout_rejected")
		}
		result = autoDelegationControllerSelect(result, VerificationBlocked, AutoDelegationControllerActionReplayOnce, FailureEvidenceMissing, "retry_auto_delegation_children", "auto_delegation_controller_replay_once")
		result.HostMayReplay = true
		return result.Normalize()
	}

	if len(result.OpenChildRefs) > 0 {
		if input.RequestedDispatch {
			result.RejectedActions = append(result.RejectedActions, AutoDelegationControllerActionSpawnOnce)
			result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, "auto_delegation_repeated_fanout_rejected")
			result.Boundaries = AppendBoundaries(result.Boundaries, "auto_delegation_repeated_fanout_rejected")
		}
		result = autoDelegationControllerSelect(result, VerificationBlocked, AutoDelegationControllerActionCollectExisting, FailureNone, "collect_auto_delegation_child_results", "auto_delegation_controller_collect_existing")
		result.HostMayCollect = true
		return result.Normalize()
	}

	if len(result.ExhaustedChildRefs) > 0 {
		if len(result.CompletedChildRefs) > 0 || len(result.MergedChildRefs) > 0 || len(result.PartialChildRefs) > 0 {
			result = autoDelegationControllerSelect(result, VerificationPartial, AutoDelegationControllerActionPartialCloseout, FailureEvidenceMissing, "update_objective_controller_with_partial_child_evidence", "auto_delegation_controller_retry_budget_exhausted_partial")
			result.ReadyForCloseout = true
			result.HostMayCollect = true
			return result.Normalize()
		}
		return autoDelegationControllerBlock(result, VerificationBlocked, FailureRepeatedNoProgress, "auto_delegation_retry_budget_exhausted", "host:auto_delegation_retry_budget", "review_auto_delegation_failure", "auto_delegation_controller_retry_budget_exhausted").Normalize()
	}

	if parentMergeProvided && parentMerge.ReadyForParentMerge {
		result = autoDelegationControllerSelect(result, VerificationPartial, AutoDelegationControllerActionPartialCloseout, FailureEvidenceMissing, "update_objective_controller_with_partial_child_evidence", "auto_delegation_controller_partial_closeout_ready")
		result.ReadyForCloseout = true
		result.HostMayCollect = true
		return result.Normalize()
	}

	if len(result.CompletedChildRefs) > 0 || len(result.FailedChildRefs) > 0 || len(result.CancelledChildRefs) > 0 {
		result = autoDelegationControllerSelect(result, VerificationBlocked, AutoDelegationControllerActionCollectExisting, FailureEvidenceMissing, "provide_auto_delegation_parent_merge", "auto_delegation_controller_collect_terminal_children")
		result.HostMayCollect = true
		if !parentMergeProvided {
			result.MissingInputs = AppendMissingInputs(result.MissingInputs, "host:auto_delegation_parent_merge")
		}
		return result.Normalize()
	}

	if len(result.InvokableChildRefs) > 0 && len(result.MissingInputs) == 0 && len(result.BlockedReasons) == 0 {
		result = autoDelegationControllerSelect(result, VerificationSatisfied, AutoDelegationControllerActionSpawnOnce, FailureNone, "host_may_invoke_auto_delegation_children", "auto_delegation_controller_spawn_once")
		result.HostMayDispatch = true
		return result.Normalize()
	}

	return autoDelegationControllerBlock(result, VerificationBlocked, firstFailureClass(result.FailureClass, FailureEvidenceMissing), "auto_delegation_no_progress_surface_missing", "host:auto_delegation_runtime_progress", "provide_auto_delegation_runtime_progress", "auto_delegation_controller_no_progress_surface").Normalize()
}

func (d AutoDelegationControllerDecision) Normalize() AutoDelegationControllerDecision {
	out := d
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeVerificationStatus(string(out.Status))
	if out.Status == VerificationNotEvaluated {
		out.Status = VerificationBlocked
	}
	out.Action = NormalizeAutoDelegationControllerAction(string(out.Action))
	out.HostBridge = out.HostBridge.Normalize()
	out.ParentMerge = out.ParentMerge.Normalize()
	out.AcceptedChildRefs = normalizeDisplaySafeRefs(out.AcceptedChildRefs)
	out.InvokableChildRefs = normalizeDisplaySafeRefs(out.InvokableChildRefs)
	out.OpenChildRefs = normalizeDisplaySafeRefs(out.OpenChildRefs)
	out.CompletedChildRefs = normalizeDisplaySafeRefs(out.CompletedChildRefs)
	out.FailedChildRefs = normalizeDisplaySafeRefs(out.FailedChildRefs)
	out.CancelledChildRefs = normalizeDisplaySafeRefs(out.CancelledChildRefs)
	out.MergedChildRefs = normalizeDisplaySafeRefs(out.MergedChildRefs)
	out.PartialChildRefs = normalizeDisplaySafeRefs(out.PartialChildRefs)
	out.RetryChildRefs = normalizeDisplaySafeRefs(out.RetryChildRefs)
	out.ReplayChildRefs = normalizeDisplaySafeRefs(out.ReplayChildRefs)
	out.ExhaustedChildRefs = normalizeDisplaySafeRefs(out.ExhaustedChildRefs)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.BlockedReasons = normalizeControlTokenList(out.BlockedReasons)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.DecisionBasis = normalizeDisplaySafeRefs(out.DecisionBasis)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	out.RuntimeEffect = normalizeControlToken(out.RuntimeEffect)
	if out.RunnerEffect == "" {
		out.RunnerEffect = "none"
	}
	if out.PromptEffect == "" {
		out.PromptEffect = "none"
	}
	if out.RuntimeEffect == "" {
		out.RuntimeEffect = "none"
	}
	actions := make([]AutoDelegationControllerAction, 0, len(out.RejectedActions))
	seen := map[AutoDelegationControllerAction]struct{}{}
	for _, action := range out.RejectedActions {
		normalized := NormalizeAutoDelegationControllerAction(string(action))
		if normalized == AutoDelegationControllerActionNone {
			continue
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		actions = append(actions, normalized)
	}
	out.RejectedActions = actions
	if out.RawOutputLoaded {
		out.Status = VerificationReviewRequired
		out.Ready = false
		out.HostMayDispatch = false
		out.HostMayCollect = false
		out.HostMayReplay = false
		out.ReadyForCloseout = false
		out.Action = AutoDelegationControllerActionBlock
		out.FailureClass = firstFailureClass(out.FailureClass, FailureEvidenceWeak)
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		out.NextHostAction = "provide_display_safe_refs"
	}
	if out.Status != VerificationSatisfied && out.Status != VerificationPartial {
		out.Ready = false
	}
	return out
}

func autoDelegationControllerSelect(result AutoDelegationControllerDecision, status VerificationStatus, action AutoDelegationControllerAction, failure FailureClass, next NextHostAction, boundary Boundary) AutoDelegationControllerDecision {
	result.Status = status
	result.Action = action
	result.Ready = true
	result.FailureClass = firstFailureClass(failure, result.FailureClass)
	result.NextHostAction = firstNextHostAction(next, result.NextHostAction)
	result.Boundaries = AppendBoundaries(result.Boundaries, boundary)
	return result
}

func autoDelegationControllerBlock(result AutoDelegationControllerDecision, status VerificationStatus, failure FailureClass, reason string, missing MissingInput, next NextHostAction, boundary Boundary) AutoDelegationControllerDecision {
	result.Status = status
	result.Action = AutoDelegationControllerActionBlock
	result.Ready = false
	result.HostMayDispatch = false
	result.HostMayCollect = false
	result.HostMayReplay = false
	result.ReadyForCloseout = false
	result.FailureClass = firstFailureClass(failure, result.FailureClass)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = firstNextHostAction(next, result.NextHostAction)
	result.Boundaries = AppendBoundaries(result.Boundaries, boundary)
	return result
}

func autoDelegationControllerReplayBudget(childRefs []DisplaySafeRef, states []AutoDelegationControllerChildState, maxAttempts int) ([]DisplaySafeRef, []DisplaySafeRef) {
	if maxAttempts <= 0 {
		maxAttempts = DefaultAutoDelegationMaxAttemptsPerChild
	}
	statesByChild := map[DisplaySafeRef]AutoDelegationControllerChildState{}
	for _, state := range states {
		if state.ChildRef == "" {
			continue
		}
		if _, exists := statesByChild[state.ChildRef]; exists {
			continue
		}
		statesByChild[state.ChildRef] = state
	}
	replay := []DisplaySafeRef{}
	exhausted := []DisplaySafeRef{}
	for _, childRef := range normalizeDisplaySafeRefs(childRefs) {
		state := statesByChild[childRef]
		if state.Attempts < maxAttempts {
			replay = appendDisplaySafeRefIfPresent(replay, childRef)
		} else {
			exhausted = appendDisplaySafeRefIfPresent(exhausted, childRef)
		}
	}
	return replay, exhausted
}

func autoDelegationControllerNormalizeChildStates(states []AutoDelegationControllerChildState) []AutoDelegationControllerChildState {
	out := make([]AutoDelegationControllerChildState, 0, len(states))
	seen := map[DisplaySafeRef]struct{}{}
	for _, state := range states {
		normalized := state.Normalize()
		if normalized.ChildRef == "" {
			continue
		}
		if _, exists := seen[normalized.ChildRef]; exists {
			continue
		}
		seen[normalized.ChildRef] = struct{}{}
		out = append(out, normalized)
	}
	return out
}

func (s AutoDelegationControllerChildState) Normalize() AutoDelegationControllerChildState {
	out := s
	out.ChildRef = normalizeOneDisplaySafeRef(out.ChildRef)
	if out.Attempts < 0 {
		out.Attempts = 0
	}
	out.LastAttemptRef = normalizeOneAttemptRef(out.LastAttemptRef)
	out.LastFailureClass = NormalizeFailureClass(string(out.LastFailureClass))
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	if out.RawOutputLoaded {
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
	}
	return out
}

func autoDelegationControllerHostBridgeProvided(bridge AutoDelegationHostBridge) bool {
	return bridge.ContractVersion != "" ||
		bridge.Status != "" ||
		bridge.PlanReview.ContractVersion != "" ||
		len(bridge.Children) > 0 ||
		len(bridge.AcceptedChildRefs) > 0 ||
		len(bridge.InvokableChildRefs) > 0 ||
		len(bridge.ActiveChildRefs) > 0 ||
		len(bridge.CompletedChildRefs) > 0 ||
		len(bridge.FailedChildRefs) > 0 ||
		len(bridge.CancelledChildRefs) > 0
}

func autoDelegationControllerParentMergeProvided(merge AutoDelegationParentMerge) bool {
	return merge.ContractVersion != "" ||
		merge.Status != "" ||
		merge.Decision != "" ||
		len(merge.Children) > 0 ||
		len(merge.AcceptedChildRefs) > 0 ||
		len(merge.MergedChildRefs) > 0 ||
		len(merge.PartialChildRefs) > 0 ||
		len(merge.RetryChildRefs) > 0 ||
		len(merge.AlternatePathChildRefs) > 0 ||
		len(merge.BlockedChildRefs) > 0
}

func autoDelegationControllerChildStateMissingInputs(states []AutoDelegationControllerChildState) []MissingInput {
	out := []MissingInput{}
	for _, state := range states {
		out = AppendMissingInputs(out, state.MissingInputs...)
	}
	return out
}

func autoDelegationControllerChildStateBlockedReasons(states []AutoDelegationControllerChildState) []string {
	out := []string{}
	for _, state := range states {
		if state.LastFailureClass != "" && state.LastFailureClass != FailureNone {
			out = appendUniqueControlToken(out, "auto_delegation_child_last_failure_"+string(state.LastFailureClass))
		}
	}
	return out
}

func autoDelegationControllerChildStateBoundaries(states []AutoDelegationControllerChildState) []Boundary {
	out := []Boundary{}
	for _, state := range states {
		out = AppendBoundaries(out, state.Boundaries...)
	}
	return out
}

func autoDelegationControllerChildStateRawOutput(states []AutoDelegationControllerChildState) bool {
	for _, state := range states {
		if state.RawOutputLoaded {
			return true
		}
	}
	return false
}

func autoDelegationControllerUnsafe(input AutoDelegationControllerInput, states []AutoDelegationControllerChildState) bool {
	if displaySafeRefSliceRejected(input.DecisionBasis) ||
		autoDelegationControllerHostBridgeUnsafe(input.HostBridge) ||
		autoDelegationControllerParentMergeUnsafeRefs(input.ParentMerge) ||
		autoDelegationControllerAsyncCompletionUnsafeRefs(input.AsyncCompletion) {
		return true
	}
	for _, state := range states {
		if displaySafeRefRejected(state.ChildRef) || state.RawOutputLoaded {
			return true
		}
	}
	return false
}

func autoDelegationControllerHostBridgeUnsafe(bridge AutoDelegationHostBridge) bool {
	return bridge.RawOutputLoaded ||
		displaySafeRefSliceRejected(bridge.AcceptedChildRefs) ||
		displaySafeRefSliceRejected(bridge.InvokableChildRefs) ||
		displaySafeRefSliceRejected(bridge.QueuedChildRefs) ||
		displaySafeRefSliceRejected(bridge.ActiveChildRefs) ||
		displaySafeRefSliceRejected(bridge.CompletedChildRefs) ||
		displaySafeRefSliceRejected(bridge.FailedChildRefs) ||
		displaySafeRefSliceRejected(bridge.CancelledChildRefs)
}

func autoDelegationControllerParentMergeUnsafeRefs(merge AutoDelegationParentMerge) bool {
	return merge.RawOutputLoaded ||
		displaySafeRefSliceRejected(merge.AcceptedChildRefs) ||
		displaySafeRefSliceRejected(merge.MergedChildRefs) ||
		displaySafeRefSliceRejected(merge.PartialChildRefs) ||
		displaySafeRefSliceRejected(merge.RetryChildRefs) ||
		displaySafeRefSliceRejected(merge.AlternatePathChildRefs) ||
		displaySafeRefSliceRejected(merge.PrunedChildRefs) ||
		displaySafeRefSliceRejected(merge.BlockedChildRefs)
}

func autoDelegationControllerAsyncCompletionProvided(projection AutoDelegationAsyncCompletionProjection) bool {
	return projection.ContractVersion != "" ||
		projection.Status != "" ||
		projection.BackendKind != "" ||
		projection.BackendRef != "" ||
		projection.ParentObjectiveRunRef != "" ||
		projection.ParentResumeRef != "" ||
		len(projection.Children) > 0 ||
		len(projection.QueuedChildRefs) > 0 ||
		len(projection.ActiveChildRefs) > 0 ||
		len(projection.CompletedChildRefs) > 0 ||
		len(projection.FailedChildRefs) > 0 ||
		len(projection.CancelledChildRefs) > 0 ||
		len(projection.InterruptedChildRefs) > 0 ||
		len(projection.CompletionEnvelopes) > 0 ||
		len(projection.ResumeRequest.ChildRefs) > 0
}

func autoDelegationControllerHostBridgeWithAsyncCompletion(hostBridge AutoDelegationHostBridge, async AutoDelegationAsyncCompletionProjection) AutoDelegationHostBridge {
	out := hostBridge.Normalize()
	async = async.Normalize()
	out.QueuedChildRefs = autoDelegationControllerMergeAsyncChildRefs(out, out.QueuedChildRefs, async.QueuedChildRefs)
	out.ActiveChildRefs = autoDelegationControllerMergeAsyncChildRefs(out, out.ActiveChildRefs, async.ActiveChildRefs)
	out.CompletedChildRefs = autoDelegationControllerMergeAsyncChildRefs(out, out.CompletedChildRefs, async.CompletedChildRefs)
	out.FailedChildRefs = autoDelegationControllerMergeAsyncChildRefs(out, out.FailedChildRefs, async.FailedChildRefs)
	out.CancelledChildRefs = autoDelegationControllerMergeAsyncChildRefs(out, out.CancelledChildRefs, async.CancelledChildRefs)
	out.CancelledChildRefs = autoDelegationControllerMergeAsyncChildRefs(out, out.CancelledChildRefs, async.InterruptedChildRefs)
	out.MissingInputs = MergeMissingInputs(out.MissingInputs, async.MissingInputs)
	out.BlockedReasons = appendUniqueControlTokens(out.BlockedReasons, async.BlockedReasons)
	out.FailureClass = firstFailureClass(out.FailureClass, async.FailureClass)
	out.DecisionBasis = mergeDisplaySafeRefs(out.DecisionBasis, async.DecisionBasis)
	out.Boundaries = MergeBoundaries(out.Boundaries, async.Boundaries, []Boundary{
		"auto_delegation_controller_consumed_async_child_readback",
	})
	out.RawOutputLoaded = out.RawOutputLoaded || async.RawOutputLoaded
	if autoDelegationControllerAsyncHasUnacceptedChildRefs(out, async) {
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:auto_delegation_accepted_children")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "auto_delegation_async_child_ref_not_accepted")
		out.FailureClass = firstFailureClass(out.FailureClass, FailurePolicyBlocked)
		out.Boundaries = AppendBoundaries(out.Boundaries, "auto_delegation_async_child_ref_not_accepted")
		out.NextHostAction = firstNextHostAction(out.NextHostAction, "review_auto_delegation_async_child_refs")
	}
	return out.Normalize()
}

func autoDelegationControllerMergeAsyncChildRefs(hostBridge AutoDelegationHostBridge, existing []DisplaySafeRef, asyncRefs []DisplaySafeRef) []DisplaySafeRef {
	accepted := hostBridge.AcceptedChildRefs
	if len(accepted) == 0 {
		return mergeDisplaySafeRefs(existing, asyncRefs)
	}
	return mergeDisplaySafeRefs(existing, intersectDisplaySafeRefs(asyncRefs, accepted))
}

func autoDelegationControllerAsyncHasUnacceptedChildRefs(hostBridge AutoDelegationHostBridge, async AutoDelegationAsyncCompletionProjection) bool {
	accepted := hostBridge.AcceptedChildRefs
	if len(accepted) == 0 {
		return false
	}
	for _, group := range [][]DisplaySafeRef{
		async.QueuedChildRefs,
		async.ActiveChildRefs,
		async.CompletedChildRefs,
		async.FailedChildRefs,
		async.CancelledChildRefs,
		async.InterruptedChildRefs,
	} {
		for _, ref := range normalizeDisplaySafeRefs(group) {
			if len(intersectDisplaySafeRefs([]DisplaySafeRef{ref}, accepted)) == 0 {
				return true
			}
		}
	}
	for _, child := range async.Children {
		if child.ChildRef != "" && len(intersectDisplaySafeRefs([]DisplaySafeRef{child.ChildRef}, accepted)) == 0 {
			return true
		}
	}
	return false
}

func autoDelegationControllerAsyncCompletionBoundaries(provided bool, projection AutoDelegationAsyncCompletionProjection) []Boundary {
	if !provided {
		return nil
	}
	return projection.Boundaries
}

func autoDelegationControllerAsyncCompletionRawOutput(provided bool, projection AutoDelegationAsyncCompletionProjection) bool {
	return provided && projection.RawOutputLoaded
}

func autoDelegationControllerAsyncCompletionUnsafeRefs(projection AutoDelegationAsyncCompletionProjection) bool {
	if !autoDelegationControllerAsyncCompletionProvided(projection) {
		return false
	}
	if projection.RawOutputLoaded ||
		autoDelegationControllerHostBridgeUnsafe(projection.HostBridge) ||
		displaySafeRefRejected(projection.BackendRef) ||
		displaySafeRefRejected(projection.ParentObjectiveRef) ||
		displaySafeRefRejected(projection.ParentObjectiveRunRef) ||
		displaySafeRefRejected(projection.ParentLedgerRef) ||
		displaySafeRefRejected(projection.ParentResumeRef) ||
		displaySafeRefRejected(projection.ResumeRequest.ParentObjectiveRef) ||
		displaySafeRefRejected(projection.ResumeRequest.ParentObjectiveRunRef) ||
		displaySafeRefRejected(projection.ResumeRequest.ParentLedgerRef) ||
		displaySafeRefRejected(projection.ResumeRequest.ParentResumeRef) ||
		displaySafeRefSliceRejected(projection.QueuedChildRefs) ||
		displaySafeRefSliceRejected(projection.ActiveChildRefs) ||
		displaySafeRefSliceRejected(projection.CompletedChildRefs) ||
		displaySafeRefSliceRejected(projection.FailedChildRefs) ||
		displaySafeRefSliceRejected(projection.CancelledChildRefs) ||
		displaySafeRefSliceRejected(projection.InterruptedChildRefs) ||
		displaySafeRefSliceRejected(projection.DecisionBasis) {
		return true
	}
	for _, child := range projection.Children {
		if displaySafeRefRejected(child.ChildRef) ||
			displaySafeRefSliceRejected(child.CapabilityRefs) ||
			displaySafeRefSliceRejected(child.AllowedToolRefs) ||
			displaySafeRefSliceRejected(child.BoundCapabilityRefs) ||
			displaySafeRefSliceRejected(child.BoundAllowedToolRefs) ||
			displaySafeRefRejected(child.WorkerRunRef) ||
			displaySafeRefRejected(child.WorkerResultRef) ||
			displaySafeRefRejected(child.WorkerReadbackRef) ||
			displaySafeRefRejected(child.ObservationRef) ||
			displaySafeRefRejected(child.FailureRef) ||
			displaySafeRefRejected(child.FailureReviewRef) ||
			displaySafeRefRejected(child.CancellationRef) ||
			displaySafeRefRejected(child.InterruptionRef) ||
			displaySafeRefRejected(child.CompletionEnvelopeRef) ||
			evidenceRefRejected(child.EvidenceRefs) {
			return true
		}
	}
	for _, envelope := range projection.CompletionEnvelopes {
		if displaySafeRefRejected(envelope.EnvelopeRef) ||
			displaySafeRefRejected(envelope.ChildRef) ||
			displaySafeRefRejected(envelope.WorkerRunRef) ||
			displaySafeRefRejected(envelope.WorkerResultRef) ||
			displaySafeRefRejected(envelope.WorkerReadbackRef) ||
			displaySafeRefRejected(envelope.ObservationRef) ||
			displaySafeRefSliceRejected(envelope.BoundCapabilityRefs) ||
			displaySafeRefSliceRejected(envelope.BoundAllowedToolRefs) ||
			evidenceRefRejected(envelope.EvidenceRefs) {
			return true
		}
	}
	return displaySafeRefSliceRejected(projection.ResumeRequest.ChildRefs) ||
		displaySafeRefSliceRejected(projection.ResumeRequest.CompletionEnvelopeRefs) ||
		displaySafeRefSliceRejected(projection.ResumeRequest.WorkerResultRefs) ||
		displaySafeRefSliceRejected(projection.ResumeRequest.WorkerReadbackRefs)
}
