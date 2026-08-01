package controlcontract

type ObjectiveSafeDefaultCandidate struct {
	MissingInput MissingInput   `json:"missing_input,omitempty"`
	DefaultRef   DisplaySafeRef `json:"default_ref,omitempty"`
	PolicyRef    DisplaySafeRef `json:"policy_ref,omitempty"`
	EvidenceRef  EvidenceRef    `json:"evidence_ref,omitempty"`
	Boundaries   []Boundary     `json:"boundaries,omitempty"`
}

func CloneObjectiveSafeDefaultCandidate(in ObjectiveSafeDefaultCandidate) ObjectiveSafeDefaultCandidate {
	out := in
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (c ObjectiveSafeDefaultCandidate) Clone() ObjectiveSafeDefaultCandidate {
	return CloneObjectiveSafeDefaultCandidate(c)
}

func (c ObjectiveSafeDefaultCandidate) Normalize() ObjectiveSafeDefaultCandidate {
	out := c.Clone()
	out.MissingInput = objectiveReplanPolicyNormalizeMissingInput(out.MissingInput)
	out.DefaultRef = normalizeOneDisplaySafeRef(out.DefaultRef)
	out.PolicyRef = normalizeOneDisplaySafeRef(out.PolicyRef)
	out.EvidenceRef = out.EvidenceRef.Normalize()
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	return out
}

type ObjectiveSafeDefaultProposalInput struct {
	ProposalRef       DisplaySafeRef                  `json:"proposal_ref,omitempty"`
	MissingInfoPolicy ObjectiveSpecMissingInfoPolicy  `json:"missing_info_policy,omitempty"`
	MissingInputs     []MissingInput                  `json:"missing_inputs,omitempty"`
	Candidates        []ObjectiveSafeDefaultCandidate `json:"candidates,omitempty"`
	PolicyRefs        []DisplaySafeRef                `json:"policy_refs,omitempty"`
	Boundaries        []Boundary                      `json:"boundaries,omitempty"`
	RawOutputLoaded   bool                            `json:"raw_output_loaded"`
}

type ObjectiveSafeDefaultProposal struct {
	ContractVersion   string                          `json:"contract_version,omitempty"`
	Projected         bool                            `json:"projected"`
	ProposalRef       DisplaySafeRef                  `json:"proposal_ref,omitempty"`
	Status            VerificationStatus              `json:"status,omitempty"`
	MissingInfoPolicy ObjectiveSpecMissingInfoPolicy  `json:"missing_info_policy,omitempty"`
	ReadyForHostApply bool                            `json:"ready_for_host_apply"`
	Defaults          []ObjectiveSafeDefaultCandidate `json:"defaults,omitempty"`
	FailureClass      FailureClass                    `json:"failure_class,omitempty"`
	MissingInputs     []MissingInput                  `json:"missing_inputs,omitempty"`
	PolicyRefs        []DisplaySafeRef                `json:"policy_refs,omitempty"`
	Boundaries        []Boundary                      `json:"boundaries,omitempty"`
	NextHostAction    NextHostAction                  `json:"next_host_action,omitempty"`
	RunnerEffect      string                          `json:"runner_effect,omitempty"`
	PromptEffect      string                          `json:"prompt_effect,omitempty"`
	RawOutputLoaded   bool                            `json:"raw_output_loaded"`
}

func BuildObjectiveSafeDefaultProposal(input ObjectiveSafeDefaultProposalInput) ObjectiveSafeDefaultProposal {
	result := ObjectiveSafeDefaultProposal{
		ContractVersion:   ContractVersion,
		Projected:         true,
		ProposalRef:       firstDisplaySafeRef(input.ProposalRef, "proposal:objective_safe_defaults"),
		Status:            VerificationBlocked,
		MissingInfoPolicy: NormalizeObjectiveSpecMissingInfoPolicy(string(input.MissingInfoPolicy)),
		MissingInputs:     normalizeMissingInputs(input.MissingInputs),
		PolicyRefs:        normalizeDisplaySafeRefs(input.PolicyRefs),
		Boundaries: AppendBoundaries(
			[]Boundary{
				"objective_safe_default_proposal",
				"projection_only",
				"host_policy_owned_defaults",
				"no_prompt_heuristic_fallback",
				"no_runner_dispatch",
				"no_runtime_adapter_execution",
			},
			input.Boundaries...,
		),
		NextHostAction:  "provide_safe_default_policy",
		RunnerEffect:    "none",
		PromptEffect:    "none",
		RawOutputLoaded: input.RawOutputLoaded,
	}
	if objectiveSafeDefaultProposalInputUnsafe(input) {
		result.Status = VerificationReviewRequired
		result.FailureClass = FailureEvidenceWeak
		result.MissingInputs = AppendMissingInputs(result.MissingInputs, "host:display_safe_refs")
		result.Boundaries = AppendBoundaries(result.Boundaries, "raw_output_not_allowed")
		result.NextHostAction = "provide_display_safe_refs"
		return result.Normalize()
	}
	if len(result.MissingInputs) == 0 {
		result.Status = VerificationNotApplicable
		result.FailureClass = FailureNone
		result.NextHostAction = "continue_objective_runtime_loop"
		result.Boundaries = AppendBoundaries(result.Boundaries, "safe_defaults_not_required")
		return result.Normalize()
	}
	switch result.MissingInfoPolicy {
	case ObjectiveSpecMissingInfoDefaultSafe, ObjectiveSpecMissingInfoQuerySafe:
		defaults, missing := objectiveSafeDefaultMatches(result.MissingInputs, input.Candidates)
		result.Defaults = defaults
		result.MissingInputs = missing
		if len(result.Defaults) > 0 && len(result.MissingInputs) == 0 {
			result.Status = VerificationSatisfied
			result.ReadyForHostApply = true
			result.FailureClass = FailureNone
			result.NextHostAction = "host_may_apply_safe_defaults"
			result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_host_safe_default_apply", "host_must_materialize_default_values")
			return result.Normalize()
		}
		result.FailureClass = FailureInsufficientInformation
		result.NextHostAction = "provide_safe_default_policy"
		result.Boundaries = AppendBoundaries(result.Boundaries, "safe_default_candidate_missing")
	case ObjectiveSpecMissingInfoAskUser:
		result.FailureClass = FailureInsufficientInformation
		result.NextHostAction = "ask_user_clarification"
		result.Boundaries = AppendBoundaries(result.Boundaries, "missing_info_policy_ask_user")
	case ObjectiveSpecMissingInfoBlock:
		result.FailureClass = FailureInsufficientInformation
		result.NextHostAction = "return_blocked"
		result.Boundaries = AppendBoundaries(result.Boundaries, "missing_info_policy_block")
	default:
		result.FailureClass = FailureInsufficientInformation
		result.MissingInputs = AppendMissingInputs(result.MissingInputs, "host:missing_info_policy")
		result.NextHostAction = "provide_missing_info_policy"
		result.Boundaries = AppendBoundaries(result.Boundaries, "missing_info_policy_required")
	}
	return result.Normalize()
}

func (p ObjectiveSafeDefaultProposal) Clone() ObjectiveSafeDefaultProposal {
	out := p
	out.Defaults = cloneObjectiveSafeDefaultCandidates(p.Defaults)
	out.MissingInputs = cloneMissingInputs(p.MissingInputs)
	out.PolicyRefs = cloneDisplaySafeRefs(p.PolicyRefs)
	out.Boundaries = cloneBoundaries(p.Boundaries)
	return out
}

func (p ObjectiveSafeDefaultProposal) Normalize() ObjectiveSafeDefaultProposal {
	out := p.Clone()
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.ProposalRef = normalizeOneDisplaySafeRef(out.ProposalRef)
	out.Status = NormalizeVerificationStatus(string(out.Status))
	out.MissingInfoPolicy = NormalizeObjectiveSpecMissingInfoPolicy(string(out.MissingInfoPolicy))
	out.Defaults = normalizeObjectiveSafeDefaultCandidates(out.Defaults)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.PolicyRefs = normalizeDisplaySafeRefs(out.PolicyRefs)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	if out.Status == VerificationNotEvaluated {
		out.Status = VerificationBlocked
	}
	if out.RunnerEffect == "" {
		out.RunnerEffect = "none"
	}
	if out.PromptEffect == "" {
		out.PromptEffect = "none"
	}
	if out.RawOutputLoaded {
		out.Status = VerificationReviewRequired
		out.ReadyForHostApply = false
		out.FailureClass = firstFailureClass(out.FailureClass, FailureEvidenceWeak)
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		out.NextHostAction = "provide_display_safe_refs"
	}
	if out.Status != VerificationSatisfied {
		out.ReadyForHostApply = false
	}
	return out
}

type ObjectiveSideEffectSplitProposalInput struct {
	ProposalRef     DisplaySafeRef   `json:"proposal_ref,omitempty"`
	Spec            ObjectiveSpec    `json:"spec,omitempty"`
	Graph           ObjectiveGraph   `json:"graph,omitempty"`
	EvidenceRefs    []EvidenceRef    `json:"evidence_refs,omitempty"`
	PolicyRefs      []DisplaySafeRef `json:"policy_refs,omitempty"`
	Boundaries      []Boundary       `json:"boundaries,omitempty"`
	RawOutputLoaded bool             `json:"raw_output_loaded"`
}

type ObjectiveSideEffectSplitProposal struct {
	ContractVersion              string             `json:"contract_version,omitempty"`
	Projected                    bool               `json:"projected"`
	ProposalRef                  DisplaySafeRef     `json:"proposal_ref,omitempty"`
	Status                       VerificationStatus `json:"status,omitempty"`
	ReadyForReadOnlyContinuation bool               `json:"ready_for_read_only_continuation"`
	BlockedSideEffects           bool               `json:"blocked_side_effects"`
	ReadOnlyNodeRefs             []DisplaySafeRef   `json:"read_only_node_refs,omitempty"`
	BlockedNodeRefs              []DisplaySafeRef   `json:"blocked_node_refs,omitempty"`
	BlockedSideEffectClasses     []string           `json:"blocked_side_effect_classes,omitempty"`
	EvidenceRefs                 []EvidenceRef      `json:"evidence_refs,omitempty"`
	FailureClass                 FailureClass       `json:"failure_class,omitempty"`
	MissingInputs                []MissingInput     `json:"missing_inputs,omitempty"`
	PolicyRefs                   []DisplaySafeRef   `json:"policy_refs,omitempty"`
	Boundaries                   []Boundary         `json:"boundaries,omitempty"`
	NextHostAction               NextHostAction     `json:"next_host_action,omitempty"`
	RunnerEffect                 string             `json:"runner_effect,omitempty"`
	PromptEffect                 string             `json:"prompt_effect,omitempty"`
	RawOutputLoaded              bool               `json:"raw_output_loaded"`
}

func BuildObjectiveSideEffectSplitProposal(input ObjectiveSideEffectSplitProposalInput) ObjectiveSideEffectSplitProposal {
	spec := input.Spec.Normalize()
	graph := input.Graph.Normalize()
	result := ObjectiveSideEffectSplitProposal{
		ContractVersion: ContractVersion,
		Projected:       true,
		ProposalRef:     firstDisplaySafeRef(input.ProposalRef, "proposal:objective_side_effect_split"),
		Status:          VerificationNotApplicable,
		EvidenceRefs:    normalizeEvidenceRefs(input.EvidenceRefs),
		PolicyRefs:      normalizeDisplaySafeRefs(input.PolicyRefs),
		Boundaries: AppendBoundaries(
			[]Boundary{
				"objective_side_effect_split_proposal",
				"projection_only",
				"side_effect_policy_gate",
				"no_runner_dispatch",
				"no_runtime_adapter_execution",
			},
			input.Boundaries...,
		),
		NextHostAction:  "continue_objective_runtime_loop",
		RunnerEffect:    "none",
		PromptEffect:    "none",
		RawOutputLoaded: input.RawOutputLoaded || spec.RawOutputLoaded || graph.RawOutputLoaded,
	}
	if objectiveSideEffectSplitProposalInputUnsafe(input) {
		result.Status = VerificationReviewRequired
		result.FailureClass = FailureEvidenceWeak
		result.MissingInputs = AppendMissingInputs(result.MissingInputs, "host:display_safe_refs")
		result.Boundaries = AppendBoundaries(result.Boundaries, "raw_output_not_allowed")
		result.NextHostAction = "provide_display_safe_refs"
		return result.Normalize()
	}
	for _, node := range graph.Nodes {
		node = node.Normalize()
		switch node.SideEffectClass {
		case ObjectiveCapabilitySideEffectReadOnly:
			result.ReadOnlyNodeRefs = appendUniqueDisplaySafeRef(result.ReadOnlyNodeRefs, node.NodeRef)
		case ObjectiveCapabilitySideEffectUnspecified:
			if node.NodeRef != "" {
				result.MissingInputs = AppendMissingInputs(result.MissingInputs, MissingInput("host:objective_graph_node_side_effect_class:"+string(node.NodeRef)))
			} else {
				result.MissingInputs = AppendMissingInputs(result.MissingInputs, "host:objective_graph_node_side_effect_class")
			}
		default:
			result.BlockedSideEffects = true
			result.BlockedNodeRefs = appendUniqueDisplaySafeRef(result.BlockedNodeRefs, node.NodeRef)
			result.BlockedSideEffectClasses = appendUniqueControlToken(result.BlockedSideEffectClasses, string(node.SideEffectClass))
		}
	}
	if len(result.MissingInputs) > 0 {
		result.Status = VerificationBlocked
		result.FailureClass = FailureInvalidInput
		result.NextHostAction = "revise_objective_graph"
		result.Boundaries = AppendBoundaries(result.Boundaries, "side_effect_split_missing_node_side_effect_class")
		return result.Normalize()
	}
	if !result.BlockedSideEffects {
		result.Status = VerificationNotApplicable
		result.FailureClass = FailureNone
		result.Boundaries = AppendBoundaries(result.Boundaries, "side_effect_split_not_required")
		return result.Normalize()
	}
	if spec.SideEffectPolicy == ObjectiveSpecSideEffectAllowed {
		result.Status = VerificationNotApplicable
		result.FailureClass = FailureNone
		result.NextHostAction = "continue_objective_runtime_loop"
		result.Boundaries = AppendBoundaries(result.Boundaries, "side_effect_policy_allows_effect_nodes")
		return result.Normalize()
	}
	result.Status = VerificationPartial
	result.FailureClass = FailurePolicyBlocked
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, "host:side_effect_approval_or_read_only_split")
	result.Boundaries = AppendBoundaries(result.Boundaries, "side_effect_nodes_blocked", "read_only_prefix_allowed", "host_must_not_execute_blocked_side_effect_nodes")
	if len(result.ReadOnlyNodeRefs) > 0 {
		result.ReadyForReadOnlyContinuation = true
		result.NextHostAction = "run_read_only_nodes_then_request_approval"
		return result.Normalize()
	}
	result.NextHostAction = "request_host_approval"
	return result.Normalize()
}

func (p ObjectiveSideEffectSplitProposal) Clone() ObjectiveSideEffectSplitProposal {
	out := p
	out.ReadOnlyNodeRefs = cloneDisplaySafeRefs(p.ReadOnlyNodeRefs)
	out.BlockedNodeRefs = cloneDisplaySafeRefs(p.BlockedNodeRefs)
	out.BlockedSideEffectClasses = cloneStringSlice(p.BlockedSideEffectClasses)
	out.EvidenceRefs = cloneEvidenceRefs(p.EvidenceRefs)
	out.MissingInputs = cloneMissingInputs(p.MissingInputs)
	out.PolicyRefs = cloneDisplaySafeRefs(p.PolicyRefs)
	out.Boundaries = cloneBoundaries(p.Boundaries)
	return out
}

func (p ObjectiveSideEffectSplitProposal) Normalize() ObjectiveSideEffectSplitProposal {
	out := p.Clone()
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.ProposalRef = normalizeOneDisplaySafeRef(out.ProposalRef)
	out.Status = NormalizeVerificationStatus(string(out.Status))
	out.ReadOnlyNodeRefs = normalizeDisplaySafeRefs(out.ReadOnlyNodeRefs)
	out.BlockedNodeRefs = normalizeDisplaySafeRefs(out.BlockedNodeRefs)
	out.BlockedSideEffectClasses = normalizeControlTokenList(out.BlockedSideEffectClasses)
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.PolicyRefs = normalizeDisplaySafeRefs(out.PolicyRefs)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	if out.Status == VerificationNotEvaluated {
		out.Status = VerificationBlocked
	}
	if out.RunnerEffect == "" {
		out.RunnerEffect = "none"
	}
	if out.PromptEffect == "" {
		out.PromptEffect = "none"
	}
	if out.RawOutputLoaded {
		out.Status = VerificationReviewRequired
		out.ReadyForReadOnlyContinuation = false
		out.FailureClass = firstFailureClass(out.FailureClass, FailureEvidenceWeak)
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		out.NextHostAction = "provide_display_safe_refs"
	}
	if out.Status != VerificationPartial {
		out.ReadyForReadOnlyContinuation = false
	}
	out.BlockedSideEffects = len(out.BlockedNodeRefs) > 0 || len(out.BlockedSideEffectClasses) > 0
	return out
}

type ObjectiveNoProgressSwitchGateInput struct {
	GateRef            DisplaySafeRef      `json:"gate_ref,omitempty"`
	CurrentStrategyRef DisplaySafeRef      `json:"current_strategy_ref,omitempty"`
	Attempts           []AttemptSummary    `json:"attempts,omitempty"`
	Verification       VerificationResult  `json:"verification,omitempty"`
	Candidates         []StrategyCandidate `json:"candidates,omitempty"`
	Threshold          int                 `json:"threshold,omitempty"`
	Boundaries         []Boundary          `json:"boundaries,omitempty"`
	RawOutputLoaded    bool                `json:"raw_output_loaded"`
}

type ObjectiveNoProgressSwitchGate struct {
	ContractVersion     string             `json:"contract_version,omitempty"`
	Projected           bool               `json:"projected"`
	GateRef             DisplaySafeRef     `json:"gate_ref,omitempty"`
	Status              VerificationStatus `json:"status,omitempty"`
	RepeatedNoProgress  bool               `json:"repeated_no_progress"`
	ReadyForSwitch      bool               `json:"ready_for_switch"`
	ReadyForCloseout    bool               `json:"ready_for_closeout"`
	CurrentStrategyRef  DisplaySafeRef     `json:"current_strategy_ref,omitempty"`
	SwitchCandidateRefs []DisplaySafeRef   `json:"switch_candidate_refs,omitempty"`
	SupportAttemptRefs  []AttemptRef       `json:"support_attempt_refs,omitempty"`
	FailureSignature    string             `json:"failure_signature,omitempty"`
	Threshold           int                `json:"threshold,omitempty"`
	FailureClass        FailureClass       `json:"failure_class,omitempty"`
	MissingInputs       []MissingInput     `json:"missing_inputs,omitempty"`
	Boundaries          []Boundary         `json:"boundaries,omitempty"`
	NextHostAction      NextHostAction     `json:"next_host_action,omitempty"`
	RunnerEffect        string             `json:"runner_effect,omitempty"`
	PromptEffect        string             `json:"prompt_effect,omitempty"`
	RawOutputLoaded     bool               `json:"raw_output_loaded"`
}

func BuildObjectiveNoProgressSwitchGate(input ObjectiveNoProgressSwitchGateInput) ObjectiveNoProgressSwitchGate {
	verification := input.Verification.Normalize()
	threshold := input.Threshold
	if threshold <= 0 {
		threshold = 2
	}
	result := ObjectiveNoProgressSwitchGate{
		ContractVersion:    ContractVersion,
		Projected:          true,
		GateRef:            firstDisplaySafeRef(input.GateRef, "gate:objective_no_progress_switch"),
		Status:             VerificationNotApplicable,
		CurrentStrategyRef: normalizeOneDisplaySafeRef(input.CurrentStrategyRef),
		Threshold:          threshold,
		Boundaries: AppendBoundaries(
			[]Boundary{
				"objective_no_progress_switch_gate",
				"projection_only",
				"no_strategy_dispatch",
				"no_runtime_adapter_execution",
			},
			input.Boundaries...,
		),
		NextHostAction:  "continue_objective_runtime_loop",
		RunnerEffect:    "none",
		PromptEffect:    "none",
		RawOutputLoaded: input.RawOutputLoaded || verification.RawOutputLoaded,
	}
	if objectiveNoProgressSwitchGateInputUnsafe(input) {
		result.Status = VerificationReviewRequired
		result.FailureClass = FailureEvidenceWeak
		result.MissingInputs = AppendMissingInputs(result.MissingInputs, "host:display_safe_refs")
		result.Boundaries = AppendBoundaries(result.Boundaries, "raw_output_not_allowed")
		result.NextHostAction = "provide_display_safe_refs"
		return result.Normalize()
	}
	result.SwitchCandidateRefs = objectiveNoProgressSwitchCandidateRefs(result.CurrentStrategyRef, input.Candidates)
	result.FailureSignature = objectiveNoProgressFailureSignature(verification.FailureClass, verification.MissingInputs, verification.NextHostAction)
	result.SupportAttemptRefs = objectiveNoProgressSupportAttemptRefs(input.Attempts, result.CurrentStrategyRef, result.FailureSignature)
	if verification.FailureClass == FailureRepeatedNoProgress || len(result.SupportAttemptRefs) >= threshold {
		result.RepeatedNoProgress = true
		result.Status = VerificationPartial
		result.FailureClass = FailureRepeatedNoProgress
		result.Boundaries = AppendBoundaries(result.Boundaries, "objective_replanner_repeated_no_progress")
		if len(result.SwitchCandidateRefs) > 0 {
			result.ReadyForSwitch = true
			result.NextHostAction = "host_may_switch_strategy"
			result.Boundaries = AppendBoundaries(result.Boundaries, "objective_replanner_no_progress_switch_strategy")
		} else {
			result.ReadyForCloseout = true
			result.MissingInputs = AppendMissingInputs(result.MissingInputs, "host:new_evidence_or_strategy")
			result.NextHostAction = "return_blocked"
			result.Boundaries = AppendBoundaries(result.Boundaries, "objective_replanner_no_progress_closeout")
		}
		return result.Normalize()
	}
	result.FailureClass = FailureNone
	result.Boundaries = AppendBoundaries(result.Boundaries, "no_progress_gate_not_triggered")
	return result.Normalize()
}

func (g ObjectiveNoProgressSwitchGate) Clone() ObjectiveNoProgressSwitchGate {
	out := g
	out.SwitchCandidateRefs = cloneDisplaySafeRefs(g.SwitchCandidateRefs)
	out.SupportAttemptRefs = cloneAttemptRefs(g.SupportAttemptRefs)
	out.MissingInputs = cloneMissingInputs(g.MissingInputs)
	out.Boundaries = cloneBoundaries(g.Boundaries)
	return out
}

func (g ObjectiveNoProgressSwitchGate) Normalize() ObjectiveNoProgressSwitchGate {
	out := g.Clone()
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.GateRef = normalizeOneDisplaySafeRef(out.GateRef)
	out.Status = NormalizeVerificationStatus(string(out.Status))
	out.CurrentStrategyRef = normalizeOneDisplaySafeRef(out.CurrentStrategyRef)
	out.SwitchCandidateRefs = normalizeDisplaySafeRefs(out.SwitchCandidateRefs)
	out.SupportAttemptRefs = normalizeAttemptRefs(out.SupportAttemptRefs)
	out.FailureSignature = normalizeControlToken(out.FailureSignature)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	if out.Status == VerificationNotEvaluated {
		out.Status = VerificationBlocked
	}
	if out.Threshold <= 0 {
		out.Threshold = 2
	}
	if out.RunnerEffect == "" {
		out.RunnerEffect = "none"
	}
	if out.PromptEffect == "" {
		out.PromptEffect = "none"
	}
	if out.RawOutputLoaded {
		out.Status = VerificationReviewRequired
		out.ReadyForSwitch = false
		out.ReadyForCloseout = false
		out.FailureClass = firstFailureClass(out.FailureClass, FailureEvidenceWeak)
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		out.NextHostAction = "provide_display_safe_refs"
	}
	if !out.RepeatedNoProgress {
		out.ReadyForSwitch = false
		out.ReadyForCloseout = false
	}
	return out
}

func objectiveSafeDefaultMatches(missing []MissingInput, candidates []ObjectiveSafeDefaultCandidate) ([]ObjectiveSafeDefaultCandidate, []MissingInput) {
	available := map[MissingInput]ObjectiveSafeDefaultCandidate{}
	for _, candidate := range normalizeObjectiveSafeDefaultCandidates(candidates) {
		if candidate.MissingInput == "" || candidate.DefaultRef == "" || candidate.PolicyRef == "" {
			continue
		}
		available[candidate.MissingInput] = candidate
	}
	defaults := []ObjectiveSafeDefaultCandidate{}
	unmatched := []MissingInput{}
	for _, missingInput := range normalizeMissingInputs(missing) {
		if candidate, ok := available[missingInput]; ok {
			defaults = append(defaults, candidate)
			continue
		}
		unmatched = AppendMissingInputs(unmatched, missingInput)
	}
	return normalizeObjectiveSafeDefaultCandidates(defaults), normalizeMissingInputs(unmatched)
}

func objectiveNoProgressSupportAttemptRefs(attempts []AttemptSummary, current DisplaySafeRef, signature string) []AttemptRef {
	if current == "" || signature == "" {
		return nil
	}
	out := []AttemptRef{}
	for _, attempt := range normalizeAttemptSummaries(attempts) {
		if attempt.Ref == "" && attempt.Status == VerificationNotEvaluated {
			continue
		}
		if attempt.StrategyID != string(current) {
			continue
		}
		if attempt.Status == VerificationSatisfied {
			continue
		}
		if len(attempt.EvidenceRefs) > 0 || attempt.ObservationCount > 0 {
			continue
		}
		attemptSignature := objectiveNoProgressFailureSignature(attempt.FailureClass, attempt.MissingInputs, attempt.NextHostAction)
		if attempt.FailureClass == FailureRepeatedNoProgress || attemptSignature == signature {
			out = appendUniqueAttemptRef(out, attempt.Ref)
		}
	}
	return normalizeAttemptRefs(out)
}

func objectiveNoProgressFailureSignature(failure FailureClass, missing []MissingInput, next NextHostAction) string {
	failure = NormalizeFailureClass(string(failure))
	if failure == FailureNone {
		return ""
	}
	token := string(failure)
	for _, value := range normalizeMissingInputs(missing) {
		token += ":" + normalizeControlToken(string(value))
	}
	if action := NormalizeNextHostAction(string(next)); action != "" {
		token += ":" + string(action)
	}
	return normalizeControlToken(token)
}

func objectiveNoProgressSwitchCandidateRefs(current DisplaySafeRef, candidates []StrategyCandidate) []DisplaySafeRef {
	out := []DisplaySafeRef{}
	for _, candidate := range normalizeStrategyCandidates(candidates) {
		ref := objectiveReplannerCandidateRef(candidate)
		if ref == "" || ref == current {
			continue
		}
		out = appendUniqueDisplaySafeRef(out, ref)
	}
	return normalizeDisplaySafeRefs(out)
}

func normalizeObjectiveSafeDefaultCandidates(in []ObjectiveSafeDefaultCandidate) []ObjectiveSafeDefaultCandidate {
	out := make([]ObjectiveSafeDefaultCandidate, 0, len(in))
	seen := map[MissingInput]bool{}
	for _, value := range in {
		normalized := value.Normalize()
		if normalized.MissingInput == "" && normalized.DefaultRef == "" {
			continue
		}
		if normalized.MissingInput != "" {
			if seen[normalized.MissingInput] {
				continue
			}
			seen[normalized.MissingInput] = true
		}
		out = append(out, normalized)
	}
	return out
}

func cloneObjectiveSafeDefaultCandidates(in []ObjectiveSafeDefaultCandidate) []ObjectiveSafeDefaultCandidate {
	if len(in) == 0 {
		return nil
	}
	out := make([]ObjectiveSafeDefaultCandidate, len(in))
	for i := range in {
		out[i] = in[i].Clone()
	}
	return out
}

func appendUniqueAttemptRef(in []AttemptRef, value AttemptRef) []AttemptRef {
	ref, ok := NormalizeAttemptRef(string(value))
	if !ok {
		return in
	}
	for _, existing := range in {
		if existing == ref {
			return in
		}
	}
	return append(in, ref)
}

func objectiveReplanPolicyNormalizeMissingInput(value MissingInput) MissingInput {
	normalized := normalizeMissingInputs([]MissingInput{value})
	if len(normalized) == 0 {
		return ""
	}
	return normalized[0]
}

func objectiveSafeDefaultProposalInputUnsafe(input ObjectiveSafeDefaultProposalInput) bool {
	if input.RawOutputLoaded ||
		displaySafeRefRejected(input.ProposalRef) ||
		displaySafeRefSliceRejected(input.PolicyRefs) {
		return true
	}
	for _, candidate := range input.Candidates {
		if displaySafeRefRejected(candidate.DefaultRef) ||
			displaySafeRefRejected(candidate.PolicyRef) ||
			evidenceRefRejected([]EvidenceRef{candidate.EvidenceRef}) {
			return true
		}
	}
	return false
}

func objectiveSideEffectSplitProposalInputUnsafe(input ObjectiveSideEffectSplitProposalInput) bool {
	return input.RawOutputLoaded ||
		displaySafeRefRejected(input.ProposalRef) ||
		input.Spec.RawOutputLoaded ||
		input.Graph.RawOutputLoaded ||
		evidenceRefRejected(input.EvidenceRefs) ||
		displaySafeRefSliceRejected(input.PolicyRefs)
}

func objectiveNoProgressSwitchGateInputUnsafe(input ObjectiveNoProgressSwitchGateInput) bool {
	if input.RawOutputLoaded ||
		displaySafeRefRejected(input.GateRef) ||
		displaySafeRefRejected(input.CurrentStrategyRef) ||
		objectiveReplannerVerificationUnsafe(input.Verification) ||
		objectiveReplannerStrategyCandidatesUnsafe(input.Candidates) {
		return true
	}
	for _, attempt := range input.Attempts {
		normalized := attempt.Normalize()
		if normalized.RawOutputLoaded ||
			(normalized.Ref != "" && displaySafeRefRejected(DisplaySafeRef(normalized.Ref))) ||
			(normalized.StrategyID != "" && displaySafeRefRejected(DisplaySafeRef(normalized.StrategyID))) ||
			evidenceRefRejected(normalized.EvidenceRefs) {
			return true
		}
	}
	return false
}
