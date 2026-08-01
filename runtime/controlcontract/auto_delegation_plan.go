package controlcontract

import "strings"

const DefaultAutoDelegationMaxDepth = 1

type AutoDelegationChildRole string

const (
	AutoDelegationChildRoleLeaf         AutoDelegationChildRole = "leaf"
	AutoDelegationChildRoleOrchestrator AutoDelegationChildRole = "orchestrator"
)

func NormalizeAutoDelegationChildRole(raw string) AutoDelegationChildRole {
	switch normalizeEnumToken(raw) {
	case "", "leaf", "worker", "task":
		return AutoDelegationChildRoleLeaf
	case "orchestrator", "planner", "coordinator":
		return AutoDelegationChildRoleOrchestrator
	default:
		return AutoDelegationChildRoleLeaf
	}
}

type AutoDelegationConversationOwnership string

const (
	AutoDelegationConversationParentOwned AutoDelegationConversationOwnership = "parent_owned"
	AutoDelegationConversationHandoff     AutoDelegationConversationOwnership = "handoff"
)

func NormalizeAutoDelegationConversationOwnership(raw string) AutoDelegationConversationOwnership {
	switch normalizeEnumToken(raw) {
	case "", "parent_owned", "parent", "worker_as_tool", "tool":
		return AutoDelegationConversationParentOwned
	case "handoff", "child_owned", "transfer":
		return AutoDelegationConversationHandoff
	default:
		return AutoDelegationConversationParentOwned
	}
}

type AutoDelegationMergeStrategy string

const (
	AutoDelegationMergeSummarize      AutoDelegationMergeStrategy = "summarize"
	AutoDelegationMergeAppendEvidence AutoDelegationMergeStrategy = "append_evidence"
	AutoDelegationMergeCompare        AutoDelegationMergeStrategy = "compare"
	AutoDelegationMergeFirstSuccess   AutoDelegationMergeStrategy = "first_success"
	AutoDelegationMergeManualReview   AutoDelegationMergeStrategy = "manual_review"
)

func NormalizeAutoDelegationMergeStrategy(raw string) AutoDelegationMergeStrategy {
	switch normalizeEnumToken(raw) {
	case "", "summarize", "summary":
		return AutoDelegationMergeSummarize
	case "append_evidence", "evidence", "ledger":
		return AutoDelegationMergeAppendEvidence
	case "compare", "cross_check", "crosscheck":
		return AutoDelegationMergeCompare
	case "first_success", "first":
		return AutoDelegationMergeFirstSuccess
	case "manual_review", "review":
		return AutoDelegationMergeManualReview
	default:
		return AutoDelegationMergeSummarize
	}
}

type AutoDelegationChildTask struct {
	ContractVersion       string                              `json:"contract_version,omitempty"`
	ChildRef              DisplaySafeRef                      `json:"child_ref,omitempty"`
	ParentObjectiveRef    DisplaySafeRef                      `json:"parent_objective_ref,omitempty"`
	Goal                  string                              `json:"goal,omitempty"`
	ContextRefs           []DisplaySafeRef                    `json:"context_refs,omitempty"`
	RelevantFindings      []string                            `json:"relevant_findings,omitempty"`
	Constraints           []string                            `json:"constraints,omitempty"`
	CapabilityRefs        []DisplaySafeRef                    `json:"capability_refs,omitempty"`
	AllowedToolRefs       []DisplaySafeRef                    `json:"allowed_tool_refs,omitempty"`
	DeniedToolRefs        []DisplaySafeRef                    `json:"denied_tool_refs,omitempty"`
	ExpectedEvidence      []EvidenceRef                       `json:"expected_evidence,omitempty"`
	ExpectedOutput        string                              `json:"expected_output,omitempty"`
	Role                  AutoDelegationChildRole             `json:"role,omitempty"`
	Depth                 int                                 `json:"depth,omitempty"`
	MaxAttempts           int                                 `json:"max_attempts,omitempty"`
	MaxDurationSeconds    int                                 `json:"max_duration_seconds,omitempty"`
	SideEffectPolicy      ObjectiveSpecSideEffectPolicy       `json:"side_effect_policy,omitempty"`
	ConversationOwnership AutoDelegationConversationOwnership `json:"conversation_ownership,omitempty"`
	MergeStrategy         AutoDelegationMergeStrategy         `json:"merge_strategy,omitempty"`
	Dependencies          []DisplaySafeRef                    `json:"dependencies,omitempty"`
	RetryPolicyRef        DisplaySafeRef                      `json:"retry_policy_ref,omitempty"`
	AlternatePathRefs     []DisplaySafeRef                    `json:"alternate_path_refs,omitempty"`
	Boundaries            []Boundary                          `json:"boundaries,omitempty"`
	MissingInputs         []MissingInput                      `json:"missing_inputs,omitempty"`
	RawOutputLoaded       bool                                `json:"raw_output_loaded"`
}

func CloneAutoDelegationChildTask(in AutoDelegationChildTask) AutoDelegationChildTask {
	out := in
	out.ContextRefs = cloneDisplaySafeRefs(in.ContextRefs)
	out.RelevantFindings = cloneStringSlice(in.RelevantFindings)
	out.Constraints = cloneStringSlice(in.Constraints)
	out.CapabilityRefs = cloneDisplaySafeRefs(in.CapabilityRefs)
	out.AllowedToolRefs = cloneDisplaySafeRefs(in.AllowedToolRefs)
	out.DeniedToolRefs = cloneDisplaySafeRefs(in.DeniedToolRefs)
	out.ExpectedEvidence = cloneEvidenceRefs(in.ExpectedEvidence)
	out.Dependencies = cloneDisplaySafeRefs(in.Dependencies)
	out.AlternatePathRefs = cloneDisplaySafeRefs(in.AlternatePathRefs)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	return out
}

func (t AutoDelegationChildTask) Clone() AutoDelegationChildTask {
	return CloneAutoDelegationChildTask(t)
}

func (t AutoDelegationChildTask) Normalize() AutoDelegationChildTask {
	out := t.Clone()
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.ChildRef = normalizeOneDisplaySafeRef(out.ChildRef)
	out.ParentObjectiveRef = normalizeOneDisplaySafeRef(out.ParentObjectiveRef)
	out.Goal = autoDelegationPlanSafeText(out.Goal)
	out.ContextRefs = normalizeDisplaySafeRefs(out.ContextRefs)
	out.RelevantFindings = normalizeStringList(out.RelevantFindings)
	out.Constraints = normalizeStringList(out.Constraints)
	out.CapabilityRefs = normalizeDisplaySafeRefs(out.CapabilityRefs)
	out.AllowedToolRefs = normalizeDisplaySafeRefs(out.AllowedToolRefs)
	out.DeniedToolRefs = normalizeDisplaySafeRefs(out.DeniedToolRefs)
	out.ExpectedEvidence = normalizeEvidenceRefs(out.ExpectedEvidence)
	out.ExpectedOutput = autoDelegationPlanSafeText(out.ExpectedOutput)
	out.Role = NormalizeAutoDelegationChildRole(string(out.Role))
	if out.Depth < 0 {
		out.Depth = 0
	}
	if out.MaxAttempts < 0 {
		out.MaxAttempts = 0
	}
	if out.MaxDurationSeconds < 0 {
		out.MaxDurationSeconds = 0
	}
	out.SideEffectPolicy = NormalizeObjectiveSpecSideEffectPolicy(string(out.SideEffectPolicy))
	out.ConversationOwnership = NormalizeAutoDelegationConversationOwnership(string(out.ConversationOwnership))
	out.MergeStrategy = NormalizeAutoDelegationMergeStrategy(string(out.MergeStrategy))
	out.Dependencies = normalizeDisplaySafeRefs(out.Dependencies)
	out.RetryPolicyRef = normalizeOneDisplaySafeRef(out.RetryPolicyRef)
	out.AlternatePathRefs = normalizeDisplaySafeRefs(out.AlternatePathRefs)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	return out
}

type AutoDelegationPlan struct {
	ContractVersion       string                              `json:"contract_version,omitempty"`
	PlanRef               DisplaySafeRef                      `json:"plan_ref,omitempty"`
	PolicyRef             DisplaySafeRef                      `json:"policy_ref,omitempty"`
	ParentObjectiveRef    DisplaySafeRef                      `json:"parent_objective_ref,omitempty"`
	Policy                AutoDelegationPolicy                `json:"policy,omitempty"`
	Children              []AutoDelegationChildTask           `json:"children,omitempty"`
	MaxChildren           int                                 `json:"max_children,omitempty"`
	MaxParallelism        int                                 `json:"max_parallelism,omitempty"`
	MaxDepth              int                                 `json:"max_depth,omitempty"`
	MergeStrategy         AutoDelegationMergeStrategy         `json:"merge_strategy,omitempty"`
	ConversationOwnership AutoDelegationConversationOwnership `json:"conversation_ownership,omitempty"`
	RequiredEvidence      []EvidenceRef                       `json:"required_evidence,omitempty"`
	Boundaries            []Boundary                          `json:"boundaries,omitempty"`
	MissingInputs         []MissingInput                      `json:"missing_inputs,omitempty"`
	RawOutputLoaded       bool                                `json:"raw_output_loaded"`
}

func CloneAutoDelegationPlan(in AutoDelegationPlan) AutoDelegationPlan {
	out := in
	out.Policy = in.Policy.Clone()
	if len(in.Children) > 0 {
		out.Children = make([]AutoDelegationChildTask, 0, len(in.Children))
		for _, child := range in.Children {
			out.Children = append(out.Children, child.Clone())
		}
	}
	out.RequiredEvidence = cloneEvidenceRefs(in.RequiredEvidence)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	return out
}

func (p AutoDelegationPlan) Clone() AutoDelegationPlan {
	return CloneAutoDelegationPlan(p)
}

func (p AutoDelegationPlan) Normalize() AutoDelegationPlan {
	out := p.Clone()
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.PlanRef = normalizeOneDisplaySafeRef(out.PlanRef)
	out.PolicyRef = normalizeOneDisplaySafeRef(out.PolicyRef)
	out.ParentObjectiveRef = normalizeOneDisplaySafeRef(out.ParentObjectiveRef)
	out.Policy = out.Policy.Normalize()
	if len(out.Children) > 0 {
		children := make([]AutoDelegationChildTask, 0, len(out.Children))
		for _, child := range out.Children {
			children = append(children, child.Normalize())
		}
		out.Children = children
	}
	if out.MaxChildren < 0 {
		out.MaxChildren = 0
	}
	if out.MaxParallelism < 0 {
		out.MaxParallelism = 0
	}
	if out.MaxDepth <= 0 {
		out.MaxDepth = DefaultAutoDelegationMaxDepth
	}
	if out.Policy.Enabled {
		if out.MaxChildren == 0 {
			out.MaxChildren = out.Policy.MaxChildren
		}
		if out.MaxParallelism == 0 {
			out.MaxParallelism = out.Policy.MaxParallelism
		}
		if out.MaxParallelism > out.MaxChildren {
			out.MaxParallelism = out.MaxChildren
		}
	}
	out.MergeStrategy = NormalizeAutoDelegationMergeStrategy(string(out.MergeStrategy))
	out.ConversationOwnership = NormalizeAutoDelegationConversationOwnership(string(out.ConversationOwnership))
	out.RequiredEvidence = normalizeEvidenceRefs(out.RequiredEvidence)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	return out
}

type AutoDelegationPlanReview struct {
	ContractVersion     string                     `json:"contract_version,omitempty"`
	Projected           bool                       `json:"projected"`
	Status              VerificationStatus         `json:"status,omitempty"`
	Ready               bool                       `json:"ready"`
	PlanOnly            bool                       `json:"plan_only"`
	HostMayDispatch     bool                       `json:"host_may_dispatch"`
	WorkerAsToolDefault bool                       `json:"worker_as_tool_default"`
	PolicyReview        AutoDelegationPolicyReview `json:"policy_review,omitempty"`
	Plan                AutoDelegationPlan         `json:"plan,omitempty"`
	AcceptedChildRefs   []DisplaySafeRef           `json:"accepted_child_refs,omitempty"`
	RejectedChildRefs   []DisplaySafeRef           `json:"rejected_child_refs,omitempty"`
	MissingInputs       []MissingInput             `json:"missing_inputs,omitempty"`
	BlockedReasons      []string                   `json:"blocked_reasons,omitempty"`
	FailureClass        FailureClass               `json:"failure_class,omitempty"`
	Boundaries          []Boundary                 `json:"boundaries,omitempty"`
	NextHostAction      NextHostAction             `json:"next_host_action,omitempty"`
	RunnerEffect        string                     `json:"runner_effect,omitempty"`
	PromptEffect        string                     `json:"prompt_effect,omitempty"`
	RawOutputLoaded     bool                       `json:"raw_output_loaded"`
}

func BuildAutoDelegationPlanReview(policyReview AutoDelegationPolicyReview, plan AutoDelegationPlan) AutoDelegationPlanReview {
	unsafe := autoDelegationPlanUnsafe(plan)
	policyReview = policyReview.Normalize()
	if autoDelegationPolicyEmpty(plan.Policy) {
		plan.Policy = policyReview.Policy
	}
	if plan.PolicyRef == "" {
		plan.PolicyRef = policyReview.Policy.PolicyRef
	}
	normalized := plan.Normalize()
	result := baseAutoDelegationPlanReview(policyReview, normalized)
	if unsafe || policyReview.RawOutputLoaded || normalized.RawOutputLoaded {
		return autoDelegationPlanReviewBlock(result, VerificationReviewRequired, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs", "raw_output_not_allowed")
	}

	switch normalized.Policy.Mode {
	case AutoDelegationObserve:
		result.Status = VerificationSatisfied
		result.Ready = true
		result.PlanOnly = true
		result.FailureClass = FailureNone
		result.NextHostAction = "record_auto_delegation_observation"
		result.PromptEffect = "observe_only"
		result.Boundaries = AppendBoundaries(result.Boundaries, "auto_delegation_observed_plan_only")
		return result.Normalize()
	case AutoDelegationOff:
		return autoDelegationPlanReviewBlock(result, policyReview.Status, firstFailureClass(policyReview.FailureClass, FailureNone), "auto_delegation_policy_not_ready", "host:auto_delegation_policy_review", firstNextHostAction(policyReview.NextHostAction, "review_auto_delegation_policy"), "auto_delegation_policy_not_ready")
	}
	if !policyReview.Ready || !policyReview.AutoDelegationAllowed {
		result.MissingInputs = AppendMissingInputs(result.MissingInputs, policyReview.MissingInputs...)
		return autoDelegationPlanReviewBlock(result, firstNonApplicableStatus(policyReview.Status, VerificationBlocked), firstFailureClass(policyReview.FailureClass, FailurePolicyBlocked), "auto_delegation_policy_not_ready", "host:auto_delegation_policy_review", firstNextHostAction(policyReview.NextHostAction, "review_auto_delegation_policy"), "auto_delegation_policy_not_ready")
	}

	if normalized.ParentObjectiveRef == "" {
		result = autoDelegationPlanReviewAccumulate(result, FailureConfigMissing, "parent_objective_ref_missing", "host:auto_delegation_parent_objective_ref", "provide_auto_delegation_plan", "auto_delegation_parent_objective_ref_missing")
	}
	if normalized.ConversationOwnership == AutoDelegationConversationHandoff {
		result = autoDelegationPlanReviewAccumulate(result, FailureUnsupportedOperation, "auto_delegation_handoff_not_enabled", "host:auto_delegation_handoff_policy", "review_auto_delegation_handoff_policy", "handoff_not_enabled")
	}
	if len(normalized.Children) == 0 {
		result = autoDelegationPlanReviewAccumulate(result, FailureInsufficientInformation, "auto_delegation_children_missing", "host:auto_delegation_children", "provide_auto_delegation_plan", "auto_delegation_children_missing")
	}
	if autoDelegationPlanChildLimitExceeded(normalized) {
		result = autoDelegationPlanReviewAccumulate(result, FailurePolicyBlocked, "auto_delegation_child_budget_exceeded", "host:auto_delegation_child_budget", "reduce_auto_delegation_children", "auto_delegation_child_budget_exceeded")
	}

	childRefs := map[DisplaySafeRef]struct{}{}
	for _, child := range normalized.Children {
		if child.ChildRef != "" {
			childRefs[child.ChildRef] = struct{}{}
		}
	}
	seenChildRefs := map[DisplaySafeRef]struct{}{}
	for _, child := range normalized.Children {
		childValid := true
		if child.ChildRef == "" {
			result = autoDelegationPlanReviewAccumulate(result, FailureConfigMissing, "auto_delegation_child_ref_missing", "host:auto_delegation_child_ref", "provide_auto_delegation_plan", "auto_delegation_child_ref_missing")
			childValid = false
		} else if _, exists := seenChildRefs[child.ChildRef]; exists {
			result = autoDelegationPlanReviewAccumulate(result, FailureInvalidInput, "auto_delegation_child_ref_duplicate", "host:auto_delegation_child_ref", "deduplicate_auto_delegation_children", "auto_delegation_child_ref_duplicate")
			childValid = false
		} else {
			seenChildRefs[child.ChildRef] = struct{}{}
		}
		if child.ParentObjectiveRef == "" {
			result = autoDelegationPlanReviewAccumulate(result, FailureConfigMissing, "auto_delegation_child_parent_objective_ref_missing", "host:auto_delegation_child_parent_objective_ref", "provide_auto_delegation_plan", "auto_delegation_child_parent_objective_ref_missing")
			childValid = false
		}
		if child.Goal == "" {
			result = autoDelegationPlanReviewAccumulate(result, FailureInsufficientInformation, "auto_delegation_child_goal_missing", "host:auto_delegation_child_goal", "provide_auto_delegation_plan", "auto_delegation_child_goal_missing")
			childValid = false
		}
		if len(child.CapabilityRefs) == 0 && len(child.AllowedToolRefs) == 0 {
			result = autoDelegationPlanReviewAccumulate(result, FailureCapabilityMissing, "auto_delegation_child_capability_constraints_missing", "host:auto_delegation_child_capability_refs", "provide_auto_delegation_plan", "auto_delegation_child_capability_constraints_missing")
			childValid = false
		}
		if child.ExpectedOutput == "" {
			result = autoDelegationPlanReviewAccumulate(result, FailureInsufficientInformation, "auto_delegation_child_expected_output_missing", "host:auto_delegation_child_expected_output", "provide_auto_delegation_plan", "auto_delegation_child_expected_output_missing")
			childValid = false
		}
		if normalized.Policy.RequireEvidence && len(child.ExpectedEvidence) == 0 {
			result = autoDelegationPlanReviewAccumulate(result, FailureEvidenceMissing, "auto_delegation_child_expected_evidence_missing", "host:auto_delegation_child_expected_evidence", "provide_auto_delegation_plan", "auto_delegation_child_expected_evidence_missing")
			childValid = false
		}
		if child.ConversationOwnership == AutoDelegationConversationHandoff {
			result = autoDelegationPlanReviewAccumulate(result, FailureUnsupportedOperation, "auto_delegation_handoff_not_enabled", "host:auto_delegation_handoff_policy", "review_auto_delegation_handoff_policy", "handoff_not_enabled")
			childValid = false
		}
		if !autoDelegationChildDepthAllowed(child, normalized.MaxDepth) {
			result = autoDelegationPlanReviewAccumulate(result, FailurePolicyBlocked, "auto_delegation_depth_policy_blocked", "host:auto_delegation_depth_policy", "review_auto_delegation_depth_policy", "auto_delegation_depth_policy_blocked")
			childValid = false
		}
		if !autoDelegationChildSideEffectAllowed(child.SideEffectPolicy, normalized.Policy.AllowedSideEffectPolicy) {
			result = autoDelegationPlanReviewAccumulate(result, FailurePolicyBlocked, "auto_delegation_child_side_effect_policy_blocked", "host:auto_delegation_child_side_effect_policy", "review_auto_delegation_policy", "auto_delegation_child_side_effect_policy_blocked")
			childValid = false
		}
		for _, dependency := range child.Dependencies {
			if _, exists := childRefs[dependency]; !exists {
				result = autoDelegationPlanReviewAccumulate(result, FailureInvalidInput, "auto_delegation_dependency_ref_unknown", "host:auto_delegation_dependency_ref", "repair_auto_delegation_dependencies", "auto_delegation_dependency_ref_unknown")
				childValid = false
			}
		}
		if childValid && child.ChildRef != "" {
			result.AcceptedChildRefs = append(result.AcceptedChildRefs, child.ChildRef)
		} else if child.ChildRef != "" {
			result.RejectedChildRefs = append(result.RejectedChildRefs, child.ChildRef)
		}
	}
	if autoDelegationPlanHasDependencyCycle(normalized.Children) {
		result = autoDelegationPlanReviewAccumulate(result, FailureInvalidInput, "auto_delegation_dependency_cycle", "host:auto_delegation_dependency_order", "repair_auto_delegation_dependencies", "auto_delegation_dependency_cycle")
	}

	if len(result.MissingInputs) > 0 || len(result.BlockedReasons) > 0 {
		return result.Normalize()
	}
	result.Status = VerificationSatisfied
	result.Ready = true
	result.PlanOnly = policyReview.PlanOnly
	result.HostMayDispatch = policyReview.ManagedExecutionAllowed && !policyReview.PlanOnly
	result.FailureClass = FailureNone
	if result.PlanOnly {
		result.NextHostAction = "review_auto_delegation_proposal"
		result.PromptEffect = "proposal_only"
		result.Boundaries = AppendBoundaries(result.Boundaries, "auto_delegation_plan_ready_for_review")
	} else {
		result.NextHostAction = "host_may_dispatch_auto_delegation_children"
		result.PromptEffect = "managed_plan_ready"
		result.Boundaries = AppendBoundaries(result.Boundaries, "auto_delegation_plan_ready_for_host_dispatch")
	}
	return result.Normalize()
}

func (r AutoDelegationPlanReview) Clone() AutoDelegationPlanReview {
	out := r
	out.PolicyReview = r.PolicyReview.Clone()
	out.Plan = r.Plan.Clone()
	out.AcceptedChildRefs = cloneDisplaySafeRefs(r.AcceptedChildRefs)
	out.RejectedChildRefs = cloneDisplaySafeRefs(r.RejectedChildRefs)
	out.MissingInputs = cloneMissingInputs(r.MissingInputs)
	out.BlockedReasons = cloneStringSlice(r.BlockedReasons)
	out.Boundaries = cloneBoundaries(r.Boundaries)
	return out
}

func (r AutoDelegationPlanReview) Normalize() AutoDelegationPlanReview {
	out := r.Clone()
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.PolicyReview = out.PolicyReview.Normalize()
	out.Plan = out.Plan.Normalize()
	out.AcceptedChildRefs = normalizeDisplaySafeRefs(out.AcceptedChildRefs)
	out.RejectedChildRefs = normalizeDisplaySafeRefs(out.RejectedChildRefs)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.BlockedReasons = normalizeControlTokenList(out.BlockedReasons)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	if out.RunnerEffect == "" {
		out.RunnerEffect = "none"
	}
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	if out.PromptEffect == "" {
		out.PromptEffect = "none"
	}
	out.Status = NormalizeVerificationStatus(string(out.Status))
	if out.Status == VerificationNotEvaluated {
		out.Status = VerificationBlocked
	}
	out.WorkerAsToolDefault = true
	if out.RawOutputLoaded || out.PolicyReview.RawOutputLoaded || out.Plan.RawOutputLoaded {
		out.Status = VerificationReviewRequired
		out.Ready = false
		out.HostMayDispatch = false
		out.FailureClass = firstFailureClass(out.FailureClass, FailureEvidenceWeak)
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		out.NextHostAction = "provide_display_safe_refs"
	}
	if out.Status != VerificationSatisfied {
		out.Ready = false
		out.HostMayDispatch = false
	}
	if out.PlanOnly {
		out.HostMayDispatch = false
	}
	return out
}

func baseAutoDelegationPlanReview(policyReview AutoDelegationPolicyReview, plan AutoDelegationPlan) AutoDelegationPlanReview {
	return AutoDelegationPlanReview{
		ContractVersion:     ContractVersion,
		Projected:           true,
		Status:              VerificationBlocked,
		PlanOnly:            policyReview.PlanOnly,
		WorkerAsToolDefault: true,
		PolicyReview:        policyReview,
		Plan:                plan,
		FailureClass:        FailureNone,
		Boundaries: MergeBoundaries(
			[]Boundary{
				"auto_delegation_plan_review",
				"delegation_plan_contract_only",
				"no_child_task_spawn",
				"no_subagent_dispatch",
				"no_scheduler_enqueue",
				"child_output_not_fact",
				"parent_verification_required",
				"worker_as_tool_default",
				"no_handoff_by_default",
				"no_llm_call",
				"no_backend_execution",
			},
			policyReview.Boundaries,
			plan.Boundaries,
		),
		MissingInputs:   MergeMissingInputs(policyReview.MissingInputs, plan.MissingInputs),
		NextHostAction:  "review_auto_delegation_plan",
		RunnerEffect:    "none",
		PromptEffect:    "none",
		RawOutputLoaded: policyReview.RawOutputLoaded || plan.RawOutputLoaded,
	}
}

func autoDelegationPlanReviewBlock(result AutoDelegationPlanReview, status VerificationStatus, failure FailureClass, reason string, missing MissingInput, next NextHostAction, boundary Boundary) AutoDelegationPlanReview {
	result.Status = status
	result.FailureClass = failure
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = next
	result.Boundaries = AppendBoundaries(result.Boundaries, boundary)
	return result.Normalize()
}

func autoDelegationPlanReviewAccumulate(result AutoDelegationPlanReview, failure FailureClass, reason string, missing MissingInput, next NextHostAction, boundary Boundary) AutoDelegationPlanReview {
	result.Status = VerificationBlocked
	result.Ready = false
	result.HostMayDispatch = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = firstNextHostAction(next, result.NextHostAction)
	result.Boundaries = AppendBoundaries(result.Boundaries, boundary)
	return result
}

func autoDelegationPolicyEmpty(policy AutoDelegationPolicy) bool {
	return policy.ContractVersion == "" &&
		policy.PolicyRef == "" &&
		!policy.Enabled &&
		policy.Mode == "" &&
		policy.MaxChildren == 0 &&
		policy.MaxParallelism == 0 &&
		policy.MaxAttemptsPerChild == 0 &&
		policy.MaxDurationSeconds == 0 &&
		policy.MaxBudgetTokens == 0 &&
		policy.MaxCostUnits == 0 &&
		policy.AllowedSideEffectPolicy == "" &&
		len(policy.AllowedToolRefs) == 0 &&
		len(policy.DeniedToolRefs) == 0 &&
		len(policy.RequiredApprovalRefs) == 0 &&
		!policy.RequireEvidence &&
		!policy.RequireVerification &&
		!policy.AllowBackgroundReadOnly &&
		len(policy.Boundaries) == 0 &&
		len(policy.MissingInputs) == 0 &&
		!policy.RawOutputLoaded
}

func autoDelegationPlanChildLimitExceeded(plan AutoDelegationPlan) bool {
	limit := plan.MaxChildren
	if plan.Policy.MaxChildren > 0 && (limit == 0 || plan.Policy.MaxChildren < limit) {
		limit = plan.Policy.MaxChildren
	}
	return limit > 0 && len(plan.Children) > limit
}

func autoDelegationPlanHasDependencyCycle(children []AutoDelegationChildTask) bool {
	graph := map[DisplaySafeRef][]DisplaySafeRef{}
	for _, child := range children {
		if child.ChildRef == "" {
			continue
		}
		graph[child.ChildRef] = child.Dependencies
	}
	visiting := map[DisplaySafeRef]bool{}
	visited := map[DisplaySafeRef]bool{}
	var visit func(DisplaySafeRef) bool
	visit = func(ref DisplaySafeRef) bool {
		if ref == "" {
			return false
		}
		if visiting[ref] {
			return true
		}
		if visited[ref] {
			return false
		}
		visiting[ref] = true
		for _, dependency := range graph[ref] {
			if _, exists := graph[dependency]; !exists {
				continue
			}
			if visit(dependency) {
				return true
			}
		}
		visiting[ref] = false
		visited[ref] = true
		return false
	}
	for ref := range graph {
		if visit(ref) {
			return true
		}
	}
	return false
}

func autoDelegationPlanUnsafe(plan AutoDelegationPlan) bool {
	if plan.RawOutputLoaded ||
		displaySafeRefRejected(plan.PlanRef) ||
		displaySafeRefRejected(plan.PolicyRef) ||
		displaySafeRefRejected(plan.ParentObjectiveRef) ||
		autoDelegationPolicyUnsafe(plan.Policy) ||
		evidenceRefRejected(plan.RequiredEvidence) ||
		ContainsUnsafeRawOutput(string(plan.MergeStrategy), string(plan.ConversationOwnership)) {
		return true
	}
	for _, child := range plan.Children {
		if autoDelegationChildTaskUnsafe(child) {
			return true
		}
	}
	return false
}

func autoDelegationChildTaskUnsafe(child AutoDelegationChildTask) bool {
	return child.RawOutputLoaded ||
		displaySafeRefRejected(child.ChildRef) ||
		displaySafeRefRejected(child.ParentObjectiveRef) ||
		displaySafeRefSliceRejected(child.ContextRefs) ||
		displaySafeRefSliceRejected(child.CapabilityRefs) ||
		displaySafeRefSliceRejected(child.AllowedToolRefs) ||
		displaySafeRefSliceRejected(child.DeniedToolRefs) ||
		displaySafeRefSliceRejected(child.Dependencies) ||
		displaySafeRefRejected(child.RetryPolicyRef) ||
		displaySafeRefSliceRejected(child.AlternatePathRefs) ||
		evidenceRefRejected(child.ExpectedEvidence) ||
		ContainsUnsafeRawOutput(
			child.Goal,
			child.ExpectedOutput,
			string(child.Role),
			string(child.SideEffectPolicy),
			string(child.ConversationOwnership),
			string(child.MergeStrategy),
		) ||
		containsUnsafeStringSlice(child.RelevantFindings) ||
		containsUnsafeStringSlice(child.Constraints)
}

func containsUnsafeStringSlice(values []string) bool {
	for _, value := range values {
		if ContainsUnsafeRawOutput(value) {
			return true
		}
	}
	return false
}

func autoDelegationPlanSafeText(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" || ContainsUnsafeRawOutput(value) {
		return ""
	}
	return value
}

func autoDelegationChildDepthAllowed(child AutoDelegationChildTask, maxDepth int) bool {
	if maxDepth <= 0 {
		maxDepth = DefaultAutoDelegationMaxDepth
	}
	if child.Depth > maxDepth {
		return false
	}
	if child.Role != AutoDelegationChildRoleOrchestrator {
		return true
	}
	return maxDepth > child.Depth+1
}

func autoDelegationChildSideEffectAllowed(childPolicy ObjectiveSpecSideEffectPolicy, allowed ObjectiveSpecSideEffectPolicy) bool {
	childPolicy = NormalizeObjectiveSpecSideEffectPolicy(string(childPolicy))
	allowed = NormalizeObjectiveSpecSideEffectPolicy(string(allowed))
	if childPolicy == ObjectiveSpecSideEffectUnspecified {
		childPolicy = ObjectiveSpecSideEffectReadOnly
	}
	switch allowed {
	case ObjectiveSpecSideEffectReadOnly:
		return childPolicy == ObjectiveSpecSideEffectReadOnly
	case ObjectiveSpecSideEffectRequiresApproval:
		return childPolicy == ObjectiveSpecSideEffectReadOnly || childPolicy == ObjectiveSpecSideEffectRequiresApproval
	case ObjectiveSpecSideEffectAllowed:
		return childPolicy != ObjectiveSpecSideEffectForbidden
	case ObjectiveSpecSideEffectForbidden:
		return childPolicy == ObjectiveSpecSideEffectForbidden
	default:
		return childPolicy == ObjectiveSpecSideEffectReadOnly
	}
}

func firstNonApplicableStatus(value VerificationStatus, fallback VerificationStatus) VerificationStatus {
	normalized := NormalizeVerificationStatus(string(value))
	if normalized != VerificationNotEvaluated {
		return normalized
	}
	return NormalizeVerificationStatus(string(fallback))
}
