package controlcontract

import "strings"

const (
	DefaultAutoDelegationMaxChildren           = 3
	DefaultAutoDelegationMaxParallelism        = 2
	DefaultAutoDelegationMaxAttemptsPerChild   = 1
	DefaultAutoDelegationMaxDurationSeconds    = 120
	DefaultAutoDelegationAllowedSideEffect     = ObjectiveSpecSideEffectReadOnly
	DefaultAutoDelegationManagedSideEffect     = ObjectiveSpecSideEffectRequiresApproval
	DefaultAutoDelegationManagedReadOnlyAction = "review_auto_delegation_plan"
)

type AutoDelegationMode string

const (
	AutoDelegationOff             AutoDelegationMode = "off"
	AutoDelegationObserve         AutoDelegationMode = "observe"
	AutoDelegationPropose         AutoDelegationMode = "propose"
	AutoDelegationManagedReadOnly AutoDelegationMode = "managed_readonly"
	AutoDelegationManaged         AutoDelegationMode = "managed"
)

func NormalizeAutoDelegationMode(raw string) AutoDelegationMode {
	switch normalizeEnumToken(raw) {
	case "", "off", "disabled", "disable", "none":
		return AutoDelegationOff
	case "observe", "observe_only", "observeonly", "diagnostic", "diagnostics":
		return AutoDelegationObserve
	case "propose", "proposal", "advisory", "suggest":
		return AutoDelegationPropose
	case "managed_readonly", "managed_read_only", "readonly", "read_only", "safe_auto", "read_only_auto":
		return AutoDelegationManagedReadOnly
	case "managed", "auto", "automatic", "enabled":
		return AutoDelegationManaged
	default:
		return AutoDelegationOff
	}
}

type AutoDelegationPolicy struct {
	ContractVersion         string                        `json:"contract_version,omitempty"`
	PolicyRef               DisplaySafeRef                `json:"policy_ref,omitempty"`
	Enabled                 bool                          `json:"enabled"`
	Mode                    AutoDelegationMode            `json:"mode,omitempty"`
	MaxChildren             int                           `json:"max_children,omitempty"`
	MaxParallelism          int                           `json:"max_parallelism,omitempty"`
	MaxAttemptsPerChild     int                           `json:"max_attempts_per_child,omitempty"`
	MaxDurationSeconds      int                           `json:"max_duration_seconds,omitempty"`
	MaxBudgetTokens         int                           `json:"max_budget_tokens,omitempty"`
	MaxCostUnits            int                           `json:"max_cost_units,omitempty"`
	AllowedSideEffectPolicy ObjectiveSpecSideEffectPolicy `json:"allowed_side_effect_policy,omitempty"`
	AllowedToolRefs         []DisplaySafeRef              `json:"allowed_tool_refs,omitempty"`
	DeniedToolRefs          []DisplaySafeRef              `json:"denied_tool_refs,omitempty"`
	RequiredApprovalRefs    []DisplaySafeRef              `json:"required_approval_refs,omitempty"`
	RequireEvidence         bool                          `json:"require_evidence"`
	RequireVerification     bool                          `json:"require_verification"`
	AllowBackgroundReadOnly bool                          `json:"allow_background_read_only"`
	Boundaries              []Boundary                    `json:"boundaries,omitempty"`
	MissingInputs           []MissingInput                `json:"missing_inputs,omitempty"`
	RawOutputLoaded         bool                          `json:"raw_output_loaded"`
}

func CloneAutoDelegationPolicy(in AutoDelegationPolicy) AutoDelegationPolicy {
	out := in
	out.AllowedToolRefs = cloneDisplaySafeRefs(in.AllowedToolRefs)
	out.DeniedToolRefs = cloneDisplaySafeRefs(in.DeniedToolRefs)
	out.RequiredApprovalRefs = cloneDisplaySafeRefs(in.RequiredApprovalRefs)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	return out
}

func (p AutoDelegationPolicy) Clone() AutoDelegationPolicy {
	return CloneAutoDelegationPolicy(p)
}

func (p AutoDelegationPolicy) Normalize() AutoDelegationPolicy {
	out := CloneAutoDelegationPolicy(p)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.PolicyRef = normalizeOneDisplaySafeRef(out.PolicyRef)

	rawMode := strings.TrimSpace(string(out.Mode))
	if out.Enabled && rawMode == "" {
		out.Mode = AutoDelegationManagedReadOnly
	} else {
		out.Mode = NormalizeAutoDelegationMode(rawMode)
	}
	if out.Mode == AutoDelegationOff {
		out.Enabled = false
	} else {
		out.Enabled = true
	}

	if out.MaxChildren < 0 {
		out.MaxChildren = 0
	}
	if out.MaxParallelism < 0 {
		out.MaxParallelism = 0
	}
	if out.MaxAttemptsPerChild < 0 {
		out.MaxAttemptsPerChild = 0
	}
	if out.MaxDurationSeconds < 0 {
		out.MaxDurationSeconds = 0
	}
	if out.MaxBudgetTokens < 0 {
		out.MaxBudgetTokens = 0
	}
	if out.MaxCostUnits < 0 {
		out.MaxCostUnits = 0
	}
	if out.Enabled {
		if out.MaxChildren == 0 {
			out.MaxChildren = DefaultAutoDelegationMaxChildren
		}
		if out.MaxParallelism == 0 {
			out.MaxParallelism = DefaultAutoDelegationMaxParallelism
		}
		if out.MaxParallelism > out.MaxChildren {
			out.MaxParallelism = out.MaxChildren
		}
		if out.MaxAttemptsPerChild == 0 {
			out.MaxAttemptsPerChild = DefaultAutoDelegationMaxAttemptsPerChild
		}
		if out.MaxDurationSeconds == 0 {
			out.MaxDurationSeconds = DefaultAutoDelegationMaxDurationSeconds
		}
		out.RequireEvidence = true
		out.RequireVerification = true
	}

	out.AllowedSideEffectPolicy = NormalizeObjectiveSpecSideEffectPolicy(string(out.AllowedSideEffectPolicy))
	switch {
	case out.Mode == AutoDelegationManagedReadOnly:
		out.AllowedSideEffectPolicy = DefaultAutoDelegationAllowedSideEffect
	case out.Mode == AutoDelegationManaged && out.AllowedSideEffectPolicy == ObjectiveSpecSideEffectUnspecified:
		out.AllowedSideEffectPolicy = DefaultAutoDelegationManagedSideEffect
	}

	out.AllowedToolRefs = normalizeDisplaySafeRefs(out.AllowedToolRefs)
	out.DeniedToolRefs = normalizeDisplaySafeRefs(out.DeniedToolRefs)
	out.RequiredApprovalRefs = normalizeDisplaySafeRefs(out.RequiredApprovalRefs)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	return out
}

type AutoDelegationPolicyReview struct {
	ContractVersion         string               `json:"contract_version,omitempty"`
	Projected               bool                 `json:"projected"`
	Status                  VerificationStatus   `json:"status,omitempty"`
	Ready                   bool                 `json:"ready"`
	AutoDelegationAllowed   bool                 `json:"auto_delegation_allowed"`
	PlanOnly                bool                 `json:"plan_only"`
	ManagedExecutionAllowed bool                 `json:"managed_execution_allowed"`
	ReadOnlyOnly            bool                 `json:"read_only_only"`
	Policy                  AutoDelegationPolicy `json:"policy,omitempty"`
	MissingInputs           []MissingInput       `json:"missing_inputs,omitempty"`
	BlockedReasons          []string             `json:"blocked_reasons,omitempty"`
	FailureClass            FailureClass         `json:"failure_class,omitempty"`
	Boundaries              []Boundary           `json:"boundaries,omitempty"`
	NextHostAction          NextHostAction       `json:"next_host_action,omitempty"`
	RunnerEffect            string               `json:"runner_effect,omitempty"`
	PromptEffect            string               `json:"prompt_effect,omitempty"`
	RawOutputLoaded         bool                 `json:"raw_output_loaded"`
}

func BuildAutoDelegationPolicyReview(policy AutoDelegationPolicy) AutoDelegationPolicyReview {
	unsafe := autoDelegationPolicyUnsafe(policy)
	normalized := policy.Normalize()
	result := baseAutoDelegationPolicyReview(normalized)
	if unsafe || normalized.RawOutputLoaded {
		return autoDelegationPolicyReviewBlock(result, VerificationReviewRequired, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs", "raw_output_not_allowed")
	}

	switch normalized.Mode {
	case AutoDelegationOff:
		result.Status = VerificationNotApplicable
		result.FailureClass = FailureNone
		result.NextHostAction = "enable_auto_delegation_policy"
		result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, "auto_delegation_disabled")
		result.Boundaries = AppendBoundaries(result.Boundaries, "auto_delegation_default_off")
	case AutoDelegationObserve:
		result.Status = VerificationSatisfied
		result.Ready = true
		result.PlanOnly = true
		result.FailureClass = FailureNone
		result.NextHostAction = "record_auto_delegation_observation"
		result.PromptEffect = "observe_only"
		result.Boundaries = AppendBoundaries(result.Boundaries, "auto_delegation_observe_only")
	case AutoDelegationPropose:
		result.Status = VerificationSatisfied
		result.Ready = true
		result.AutoDelegationAllowed = true
		result.PlanOnly = true
		result.FailureClass = FailureNone
		result.NextHostAction = "review_auto_delegation_proposal"
		result.PromptEffect = "proposal_only"
		result.Boundaries = AppendBoundaries(result.Boundaries, "auto_delegation_proposal_only")
	case AutoDelegationManagedReadOnly:
		result.Status = VerificationSatisfied
		result.Ready = true
		result.AutoDelegationAllowed = true
		result.ManagedExecutionAllowed = true
		result.ReadOnlyOnly = true
		result.FailureClass = FailureNone
		result.NextHostAction = DefaultAutoDelegationManagedReadOnlyAction
		result.PromptEffect = "managed_readonly_allowed"
		result.Boundaries = AppendBoundaries(result.Boundaries, "auto_delegation_managed_readonly", "read_only_child_work_only")
	case AutoDelegationManaged:
		if normalized.AllowedSideEffectPolicy == ObjectiveSpecSideEffectForbidden {
			return autoDelegationPolicyReviewBlock(result, VerificationBlocked, FailurePolicyBlocked, "auto_delegation_side_effect_forbidden", "host:auto_delegation_side_effect_policy", "review_auto_delegation_policy", "auto_delegation_policy_blocked")
		}
		if len(normalized.RequiredApprovalRefs) == 0 {
			return autoDelegationPolicyReviewBlock(result, VerificationReviewRequired, FailureApprovalRequired, "auto_delegation_approval_ref_missing", "host:auto_delegation_approval_ref", "request_auto_delegation_approval", "auto_delegation_approval_required")
		}
		result.Status = VerificationSatisfied
		result.Ready = true
		result.AutoDelegationAllowed = true
		result.ManagedExecutionAllowed = true
		result.FailureClass = FailureNone
		result.NextHostAction = "review_auto_delegation_plan"
		result.PromptEffect = "managed_allowed_with_approval"
		result.Boundaries = AppendBoundaries(result.Boundaries, "auto_delegation_managed_with_approval")
	default:
		return autoDelegationPolicyReviewBlock(result, VerificationBlocked, FailurePolicyBlocked, "auto_delegation_mode_not_allowed", "host:auto_delegation_policy", "review_auto_delegation_policy", "auto_delegation_policy_blocked")
	}

	return result.Normalize()
}

func (r AutoDelegationPolicyReview) Clone() AutoDelegationPolicyReview {
	out := r
	out.Policy = r.Policy.Clone()
	out.MissingInputs = cloneMissingInputs(r.MissingInputs)
	out.BlockedReasons = cloneStringSlice(r.BlockedReasons)
	out.Boundaries = cloneBoundaries(r.Boundaries)
	return out
}

func (r AutoDelegationPolicyReview) Normalize() AutoDelegationPolicyReview {
	out := r.Clone()
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeVerificationStatus(string(out.Status))
	if out.Status == VerificationNotEvaluated {
		out.Status = VerificationBlocked
	}
	out.Policy = out.Policy.Normalize()
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
	if out.RawOutputLoaded || out.Policy.RawOutputLoaded {
		out.Status = VerificationReviewRequired
		out.Ready = false
		out.AutoDelegationAllowed = false
		out.ManagedExecutionAllowed = false
		out.FailureClass = firstFailureClass(out.FailureClass, FailureEvidenceWeak)
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		out.NextHostAction = "provide_display_safe_refs"
	}
	if out.Status != VerificationSatisfied {
		out.Ready = false
		if out.Status != VerificationNotApplicable {
			out.AutoDelegationAllowed = false
			out.ManagedExecutionAllowed = false
		}
	}
	return out
}

func baseAutoDelegationPolicyReview(policy AutoDelegationPolicy) AutoDelegationPolicyReview {
	return AutoDelegationPolicyReview{
		ContractVersion: ContractVersion,
		Projected:       true,
		Status:          VerificationBlocked,
		Policy:          policy,
		FailureClass:    FailurePolicyBlocked,
		Boundaries: AppendBoundaries(
			[]Boundary{
				"auto_delegation_policy_review",
				"policy_projection_only",
				"no_child_task_spawn",
				"no_subagent_dispatch",
				"no_scheduler_enqueue",
				"child_output_not_fact",
				"parent_verification_required",
				"no_llm_call",
				"no_backend_execution",
			},
			policy.Boundaries...,
		),
		MissingInputs:   cloneMissingInputs(policy.MissingInputs),
		NextHostAction:  "review_auto_delegation_policy",
		RunnerEffect:    "none",
		PromptEffect:    "none",
		RawOutputLoaded: policy.RawOutputLoaded,
	}
}

func autoDelegationPolicyReviewBlock(result AutoDelegationPolicyReview, status VerificationStatus, failure FailureClass, reason string, missing MissingInput, next NextHostAction, boundary Boundary) AutoDelegationPolicyReview {
	result.Status = status
	result.FailureClass = failure
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = next
	result.Boundaries = AppendBoundaries(result.Boundaries, boundary)
	return result.Normalize()
}

func autoDelegationPolicyUnsafe(policy AutoDelegationPolicy) bool {
	return policy.RawOutputLoaded ||
		displaySafeRefRejected(policy.PolicyRef) ||
		displaySafeRefSliceRejected(policy.AllowedToolRefs) ||
		displaySafeRefSliceRejected(policy.DeniedToolRefs) ||
		displaySafeRefSliceRejected(policy.RequiredApprovalRefs) ||
		ContainsUnsafeRawOutput(string(policy.Mode), string(policy.AllowedSideEffectPolicy))
}
