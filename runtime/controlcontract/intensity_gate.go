package controlcontract

type IntensityGateStage string

const (
	IntensityGatePre   IntensityGateStage = "pre_gate"
	IntensityGateFinal IntensityGateStage = "final_gate"
)

func NormalizeIntensityGateStage(raw string) IntensityGateStage {
	switch normalizeEnumToken(raw) {
	case "pre", "pre_gate":
		return IntensityGatePre
	case "final", "final_gate":
		return IntensityGateFinal
	default:
		return IntensityGatePre
	}
}

type ExecutionIntensityPolicy struct {
	ContractVersion                string                               `json:"contract_version,omitempty"`
	PolicyRef                      DisplaySafeRef                       `json:"policy_ref,omitempty"`
	ExecutionContractRef           DisplaySafeRef                       `json:"execution_contract_ref,omitempty"`
	Activation                     Activation                           `json:"activation,omitempty"`
	DefaultIntensity               ExecutionIntensity                   `json:"default_intensity,omitempty"`
	MaxDefaultIntensity            ExecutionIntensity                   `json:"max_default_intensity,omitempty"`
	MaxAllowedIntensity            ExecutionIntensity                   `json:"max_allowed_intensity,omitempty"`
	DisabledIntensities            []ExecutionIntensity                 `json:"disabled_intensities,omitempty"`
	AllowModelSuggestedUpgrade     bool                                 `json:"allow_model_suggested_upgrade"`
	RequireUserConfirmationFrom    ExecutionIntensity                   `json:"require_user_confirmation_from,omitempty"`
	RequireHostApprovalFrom        ExecutionIntensity                   `json:"require_host_approval_from,omitempty"`
	AllowedControlModesByIntensity map[ExecutionIntensity][]ControlMode `json:"allowed_control_modes_by_intensity,omitempty"`
	DeniedControlModesByIntensity  map[ExecutionIntensity][]ControlMode `json:"denied_control_modes_by_intensity,omitempty"`
	DeniedSideEffectsByIntensity   map[ExecutionIntensity][]string      `json:"denied_side_effects_by_intensity,omitempty"`
	BudgetRefs                     []DisplaySafeRef                     `json:"budget_refs,omitempty"`
	PolicyRefs                     []DisplaySafeRef                     `json:"policy_refs,omitempty"`
	MissingInputs                  []MissingInput                       `json:"missing_inputs,omitempty"`
	Boundaries                     []Boundary                           `json:"boundaries,omitempty"`
}

func CloneExecutionIntensityPolicy(in ExecutionIntensityPolicy) ExecutionIntensityPolicy {
	out := in
	out.DisabledIntensities = cloneExecutionIntensities(in.DisabledIntensities)
	out.AllowedControlModesByIntensity = cloneControlModeByIntensity(in.AllowedControlModesByIntensity)
	out.DeniedControlModesByIntensity = cloneControlModeByIntensity(in.DeniedControlModesByIntensity)
	out.DeniedSideEffectsByIntensity = cloneStringByIntensity(in.DeniedSideEffectsByIntensity)
	out.BudgetRefs = cloneDisplaySafeRefs(in.BudgetRefs)
	out.PolicyRefs = cloneDisplaySafeRefs(in.PolicyRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (p ExecutionIntensityPolicy) Clone() ExecutionIntensityPolicy {
	return CloneExecutionIntensityPolicy(p)
}

func (p ExecutionIntensityPolicy) Normalize() ExecutionIntensityPolicy {
	out := CloneExecutionIntensityPolicy(p)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.PolicyRef = normalizeOneDisplaySafeRef(out.PolicyRef)
	out.ExecutionContractRef = normalizeOneDisplaySafeRef(out.ExecutionContractRef)
	out.Activation = NormalizeActivation(string(out.Activation))
	out.DefaultIntensity = normalizeIntensityOr(out.DefaultIntensity, IntensityL0AnswerOnly)
	out.MaxDefaultIntensity = normalizeIntensityOr(out.MaxDefaultIntensity, IntensityL2BoundedToolLoop)
	out.MaxAllowedIntensity = normalizeIntensityOr(out.MaxAllowedIntensity, out.MaxDefaultIntensity)
	out.DisabledIntensities = normalizeExecutionIntensities(out.DisabledIntensities)
	out.RequireUserConfirmationFrom = normalizeIntensityOr(out.RequireUserConfirmationFrom, IntensityL3ManagedObjective)
	out.RequireHostApprovalFrom = normalizeIntensityOr(out.RequireHostApprovalFrom, IntensityL4DurableLongRun)
	out.AllowedControlModesByIntensity = normalizeControlModeByIntensity(out.AllowedControlModesByIntensity)
	out.DeniedControlModesByIntensity = normalizeControlModeByIntensity(out.DeniedControlModesByIntensity)
	out.DeniedSideEffectsByIntensity = normalizeStringByIntensity(out.DeniedSideEffectsByIntensity)
	out.BudgetRefs = normalizeDisplaySafeRefs(out.BudgetRefs)
	out.PolicyRefs = normalizeDisplaySafeRefs(out.PolicyRefs)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	if out.PolicyRef == "" {
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "contract:intensity_policy")
		out.Boundaries = AppendBoundaries(out.Boundaries, "intensity_policy_ref_missing")
	}
	if out.ExecutionContractRef == "" {
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "contract:execution_contract")
		out.Boundaries = AppendBoundaries(out.Boundaries, "execution_contract_ref_missing")
	}
	if !displaySafeRefSliceContains(out.PolicyRefs, out.PolicyRef) && out.PolicyRef != "" {
		out.PolicyRefs = appendUniqueDisplaySafeRef(out.PolicyRefs, out.PolicyRef)
	}
	if !displaySafeRefSliceContains(out.PolicyRefs, out.ExecutionContractRef) && out.ExecutionContractRef != "" {
		out.PolicyRefs = appendUniqueDisplaySafeRef(out.PolicyRefs, out.ExecutionContractRef)
	}
	return out
}

type IntensityRouteSuggestion struct {
	ControlMode       ControlMode        `json:"control_mode,omitempty"`
	Intensity         ExecutionIntensity `json:"intensity,omitempty"`
	ModelSuggested    bool               `json:"model_suggested"`
	UserExplicit      bool               `json:"user_explicit"`
	HostRouteExplicit bool               `json:"host_route_explicit"`
	ReasonRef         DisplaySafeRef     `json:"reason_ref,omitempty"`
}

func (s IntensityRouteSuggestion) Normalize() IntensityRouteSuggestion {
	out := s
	out.ControlMode = NormalizeControlMode(string(out.ControlMode))
	out.Intensity = NormalizeExecutionIntensity(string(out.Intensity))
	out.ReasonRef = normalizeOneDisplaySafeRef(out.ReasonRef)
	return out
}

type IntensityUpgrade struct {
	ContractVersion       string                    `json:"contract_version,omitempty"`
	FromIntensity         ExecutionIntensity        `json:"from_intensity,omitempty"`
	ToIntensity           ExecutionIntensity        `json:"to_intensity,omitempty"`
	RequestedBy           string                    `json:"requested_by,omitempty"`
	Reason                string                    `json:"reason,omitempty"`
	ApprovalStatus        HostActionStatus          `json:"approval_status,omitempty"`
	BudgetStatus          VerificationStatus        `json:"budget_status,omitempty"`
	AllowedPartialAction  ObjectiveControllerAction `json:"allowed_partial_action,omitempty"`
	SuggestedUserQuestion string                    `json:"suggested_user_question,omitempty"`
	EvidenceRefs          []EvidenceRef             `json:"evidence_refs,omitempty"`
	MissingInputs         []MissingInput            `json:"missing_inputs,omitempty"`
	Boundaries            []Boundary                `json:"boundaries,omitempty"`
}

func CloneIntensityUpgrade(in IntensityUpgrade) IntensityUpgrade {
	out := in
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (u IntensityUpgrade) Clone() IntensityUpgrade {
	return CloneIntensityUpgrade(u)
}

func (u IntensityUpgrade) Normalize() IntensityUpgrade {
	out := CloneIntensityUpgrade(u)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.FromIntensity = NormalizeExecutionIntensity(string(out.FromIntensity))
	out.ToIntensity = NormalizeExecutionIntensity(string(out.ToIntensity))
	out.RequestedBy = normalizeControlToken(out.RequestedBy)
	out.Reason = managedObjectiveReplannerSafeReason(out.Reason)
	out.ApprovalStatus = NormalizeHostActionStatus(string(out.ApprovalStatus))
	out.BudgetStatus = NormalizeVerificationStatus(string(out.BudgetStatus))
	out.AllowedPartialAction = NormalizeObjectiveControllerAction(string(out.AllowedPartialAction))
	out.SuggestedUserQuestion = firstNonEmptyContractString(out.SuggestedUserQuestion)
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	return out
}

type IntensityGateInput struct {
	Stage                IntensityGateStage       `json:"stage,omitempty"`
	Activation           Activation               `json:"activation,omitempty"`
	Policy               ExecutionIntensityPolicy `json:"policy,omitempty"`
	RequestedControlMode ControlMode              `json:"requested_control_mode,omitempty"`
	RequestedIntensity   ExecutionIntensity       `json:"requested_intensity,omitempty"`
	RouteSuggestion      IntensityRouteSuggestion `json:"route_suggestion,omitempty"`
	UserConfirmed        bool                     `json:"user_confirmed"`
	UserDenied           bool                     `json:"user_denied"`
	HostApproved         bool                     `json:"host_approved"`
	HostDenied           bool                     `json:"host_denied"`
	ApprovalRefs         []DisplaySafeRef         `json:"approval_refs,omitempty"`
	Budget               ObjectiveBudgetSnapshot  `json:"budget,omitempty"`
	Strategy             StrategyCandidate        `json:"strategy,omitempty"`
	PreGate              IntensityGateResult      `json:"pre_gate,omitempty"`
	EvidenceRefs         []EvidenceRef            `json:"evidence_refs,omitempty"`
	DecisionBasis        []DisplaySafeRef         `json:"decision_basis,omitempty"`
	Boundaries           []Boundary               `json:"boundaries,omitempty"`
	RawOutputLoaded      bool                     `json:"raw_output_loaded"`
}

type IntensityGateResult struct {
	ContractVersion          string                  `json:"contract_version,omitempty"`
	Projected                bool                    `json:"projected"`
	Stage                    IntensityGateStage      `json:"stage,omitempty"`
	Activation               Activation              `json:"activation,omitempty"`
	Status                   VerificationStatus      `json:"status,omitempty"`
	Allowed                  bool                    `json:"allowed"`
	RequestedControlMode     ControlMode             `json:"requested_control_mode,omitempty"`
	ApprovedControlMode      ControlMode             `json:"approved_control_mode,omitempty"`
	SuggestedControlMode     ControlMode             `json:"suggested_control_mode,omitempty"`
	RequestedIntensity       ExecutionIntensity      `json:"requested_intensity,omitempty"`
	SuggestedIntensity       ExecutionIntensity      `json:"suggested_intensity,omitempty"`
	MaxAllowedIntensity      ExecutionIntensity      `json:"max_allowed_intensity,omitempty"`
	ApprovedIntensity        ExecutionIntensity      `json:"approved_intensity,omitempty"`
	StrategyRef              DisplaySafeRef          `json:"strategy_ref,omitempty"`
	RequiresUserConfirmation bool                    `json:"requires_user_confirmation"`
	RequiresHostApproval     bool                    `json:"requires_host_approval"`
	UserConfirmed            bool                    `json:"user_confirmed"`
	UserDenied               bool                    `json:"user_denied"`
	HostApproved             bool                    `json:"host_approved"`
	HostDenied               bool                    `json:"host_denied"`
	ApprovalRefs             []DisplaySafeRef        `json:"approval_refs,omitempty"`
	Budget                   ObjectiveBudgetSnapshot `json:"budget,omitempty"`
	PolicyRefs               []DisplaySafeRef        `json:"policy_refs,omitempty"`
	EvidenceRefs             []EvidenceRef           `json:"evidence_refs,omitempty"`
	FailureClass             FailureClass            `json:"failure_class,omitempty"`
	BlockedReason            string                  `json:"blocked_reason,omitempty"`
	MissingInputs            []MissingInput          `json:"missing_inputs,omitempty"`
	DecisionBasis            []DisplaySafeRef        `json:"decision_basis,omitempty"`
	Boundaries               []Boundary              `json:"boundaries,omitempty"`
	Upgrade                  IntensityUpgrade        `json:"upgrade,omitempty"`
	NextHostAction           NextHostAction          `json:"next_host_action,omitempty"`
	RunnerEffect             string                  `json:"runner_effect,omitempty"`
	PromptEffect             string                  `json:"prompt_effect,omitempty"`
	RawOutputLoaded          bool                    `json:"raw_output_loaded"`
}

func CloneIntensityGateResult(in IntensityGateResult) IntensityGateResult {
	out := in
	out.ApprovalRefs = cloneDisplaySafeRefs(in.ApprovalRefs)
	out.Budget = in.Budget.Clone()
	out.PolicyRefs = cloneDisplaySafeRefs(in.PolicyRefs)
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.DecisionBasis = cloneDisplaySafeRefs(in.DecisionBasis)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	out.Upgrade = in.Upgrade.Clone()
	return out
}

func (r IntensityGateResult) Clone() IntensityGateResult {
	return CloneIntensityGateResult(r)
}

func (r IntensityGateResult) Normalize() IntensityGateResult {
	out := CloneIntensityGateResult(r)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Stage = NormalizeIntensityGateStage(string(out.Stage))
	out.Activation = NormalizeActivation(string(out.Activation))
	out.Status = NormalizeVerificationStatus(string(out.Status))
	out.RequestedControlMode = NormalizeControlMode(string(out.RequestedControlMode))
	out.ApprovedControlMode = NormalizeControlMode(string(out.ApprovedControlMode))
	out.SuggestedControlMode = NormalizeControlMode(string(out.SuggestedControlMode))
	out.RequestedIntensity = NormalizeExecutionIntensity(string(out.RequestedIntensity))
	out.SuggestedIntensity = NormalizeExecutionIntensity(string(out.SuggestedIntensity))
	out.MaxAllowedIntensity = NormalizeExecutionIntensity(string(out.MaxAllowedIntensity))
	out.ApprovedIntensity = NormalizeExecutionIntensity(string(out.ApprovedIntensity))
	out.StrategyRef = normalizeOneDisplaySafeRef(out.StrategyRef)
	out.ApprovalRefs = normalizeDisplaySafeRefs(out.ApprovalRefs)
	out.Budget = out.Budget.Normalize()
	out.PolicyRefs = normalizeDisplaySafeRefs(out.PolicyRefs)
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.BlockedReason = managedObjectiveReplannerSafeReason(out.BlockedReason)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.DecisionBasis = normalizeDisplaySafeRefs(out.DecisionBasis)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.Upgrade = out.Upgrade.Normalize()
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
	out.Allowed = out.Status == VerificationSatisfied && len(out.MissingInputs) == 0 && !out.RawOutputLoaded
	if out.RawOutputLoaded && out.Status != VerificationBlocked {
		out.Status = VerificationReviewRequired
		out.Allowed = false
		if out.FailureClass == FailureNone {
			out.FailureClass = FailureEvidenceWeak
		}
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		out.NextHostAction = "provide_display_safe_refs"
	}
	return out
}

func BuildExecutionIntensityPreGate(input IntensityGateInput) IntensityGateResult {
	return buildExecutionIntensityGate(input, IntensityGatePre)
}

func BuildExecutionIntensityFinalGate(input IntensityGateInput) IntensityGateResult {
	return buildExecutionIntensityGate(input, IntensityGateFinal)
}

func buildExecutionIntensityGate(input IntensityGateInput, stage IntensityGateStage) IntensityGateResult {
	policy := input.Policy.Normalize()
	activation := intensityGateActivation(input.Activation, policy.Activation)
	route := input.RouteSuggestion.Normalize()
	budget := input.Budget.Normalize()
	strategy := input.Strategy.Normalize()
	requestedMode := firstControlMode(input.RequestedControlMode, route.ControlMode, strategy.ControlMode, ControlModeObjective)
	requestedIntensity := firstIntensity(input.RequestedIntensity, route.Intensity, strategy.MinIntensity, policy.DefaultIntensity)
	if stage == IntensityGateFinal && strategy.MinIntensity != "" {
		requestedIntensity = strategy.MinIntensity
	}
	result := IntensityGateResult{
		ContractVersion:      ContractVersion,
		Projected:            true,
		Stage:                stage,
		Activation:           activation,
		Status:               VerificationSatisfied,
		RequestedControlMode: requestedMode,
		ApprovedControlMode:  requestedMode,
		SuggestedControlMode: route.ControlMode,
		RequestedIntensity:   requestedIntensity,
		SuggestedIntensity:   route.Intensity,
		MaxAllowedIntensity:  policy.MaxAllowedIntensity,
		ApprovedIntensity:    minExecutionIntensity(requestedIntensity, policy.MaxAllowedIntensity),
		StrategyRef:          strategyRef(strategy),
		UserConfirmed:        input.UserConfirmed,
		UserDenied:           input.UserDenied,
		HostApproved:         input.HostApproved,
		HostDenied:           input.HostDenied,
		ApprovalRefs:         normalizeDisplaySafeRefs(input.ApprovalRefs),
		Budget:               budget,
		PolicyRefs:           normalizeDisplaySafeRefs(policy.PolicyRefs),
		EvidenceRefs:         normalizeEvidenceRefs(input.EvidenceRefs),
		FailureClass:         FailureNone,
		DecisionBasis:        intensityGateDecisionBasis(stage, activation, requestedMode, requestedIntensity, route, strategy, input.DecisionBasis),
		Boundaries: MergeBoundaries(
			[]Boundary{
				"execution_intensity_gate",
				Boundary(string(stage)),
				"projection_only",
				"no_runner_dispatch",
				"no_strategy_dispatch",
				"model_route_is_not_authorization",
			},
			policy.Boundaries,
			input.Boundaries,
		),
		RunnerEffect:    "none",
		PromptEffect:    "none",
		RawOutputLoaded: input.RawOutputLoaded,
	}
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, policy.MissingInputs...)
	if activation == ActivationOff {
		return intensityGateBlock(result, FailurePolicyBlocked, "control plane activation is off", "control_plane:activation", "enable_control_plane", "intensity_gate_activation_off")
	}
	if len(policy.MissingInputs) > 0 {
		return intensityGateBlock(result, FailurePolicyBlocked, "execution intensity policy is incomplete", "contract:intensity_policy", "provide_intensity_policy", "intensity_policy_incomplete")
	}
	if activation == ActivationObserveOnly {
		result.Status = VerificationNotApplicable
		result.Allowed = false
		result.FailureClass = FailurePolicyBlocked
		result.BlockedReason = "observe_only does not authorize execution intensity"
		result.NextHostAction = "none"
		result.Boundaries = AppendBoundaries(result.Boundaries, "observe_only_no_decision_effect")
		return result.Normalize()
	}
	if activation == ActivationAdvisory && executionIntensityRank(requestedIntensity) >= executionIntensityRank(IntensityL3ManagedObjective) {
		result.Upgrade = buildIntensityUpgrade(policy.MaxDefaultIntensity, requestedIntensity, route, result.EvidenceRefs, "advisory activation cannot approve L3+")
		return intensityGateBlock(result, FailureApprovalRequired, "managed activation is required for L3+", "control_plane:managed_activation", "request_intensity_upgrade_confirmation", "advisory_only_no_execution_effect")
	}
	if intensityDisabled(policy, requestedIntensity) {
		return intensityGateBlock(result, FailurePolicyBlocked, "requested intensity is disabled by policy", MissingInput("contract:intensity:"+string(requestedIntensity)), "return_partial_or_request_upgrade", "intensity_disabled_by_policy")
	}
	if executionIntensityRank(requestedIntensity) > executionIntensityRank(policy.MaxAllowedIntensity) {
		result.Upgrade = buildIntensityUpgrade(policy.MaxAllowedIntensity, requestedIntensity, route, result.EvidenceRefs, "requested intensity exceeds policy")
		return intensityGateBlock(result, FailurePolicyBlocked, "requested intensity exceeds max allowed intensity", "contract:max_allowed_intensity", "request_intensity_upgrade_confirmation", "intensity_exceeds_policy")
	}
	if denied, blocked := applyExplicitIntensityDenialGate(result, policy, input); blocked {
		return denied
	}
	if route.ModelSuggested && executionIntensityRank(route.Intensity) > executionIntensityRank(policy.MaxDefaultIntensity) && !policy.AllowModelSuggestedUpgrade {
		result.Upgrade = buildIntensityUpgrade(policy.MaxDefaultIntensity, route.Intensity, route, result.EvidenceRefs, "model suggestion cannot authorize upgrade")
		return intensityGateBlock(result, FailureApprovalRequired, "model route suggestion cannot authorize intensity upgrade", "host:intensity_upgrade_confirmation", "request_intensity_upgrade_confirmation", "model_route_is_not_authorization")
	}
	if !controlModeAllowedForIntensity(policy, requestedIntensity, requestedMode) {
		return intensityGateBlock(result, FailurePolicyBlocked, "control mode is not allowed for requested intensity", MissingInput("contract:control_mode:"+string(requestedMode)), "return_partial_or_request_upgrade", "control_mode_denied_by_intensity_policy")
	}
	if stage == IntensityGateFinal {
		if !intensityGateResultInputEmpty(input.PreGate) {
			preGate := input.PreGate.Normalize()
			result = mergePreGateResult(result, preGate)
			if !preGate.Allowed {
				result.Upgrade = preGate.Upgrade
				return intensityGateBlock(result, firstFailureClass(preGate.FailureClass, FailurePolicyBlocked), "pre-gate is not satisfied", "control_plane:pre_gate", "satisfy_intensity_pre_gate", "pre_gate_not_satisfied")
			}
			if executionIntensityRank(requestedIntensity) > executionIntensityRank(result.MaxAllowedIntensity) {
				result.Upgrade = buildIntensityUpgrade(result.MaxAllowedIntensity, requestedIntensity, route, result.EvidenceRefs, "strategy exceeds pre-gate allowance")
				return intensityGateBlock(result, FailurePolicyBlocked, "strategy intensity exceeds pre-gate allowance", "contract:pre_gate_intensity", "request_intensity_upgrade_confirmation", "strategy_intensity_exceeds_pre_gate")
			}
		}
		result = applyStrategyFinalGate(result, policy, strategy, input)
		if result.Status != VerificationSatisfied {
			return result.Normalize()
		}
	} else if executionIntensityRank(requestedIntensity) >= executionIntensityRank(IntensityL3ManagedObjective) {
		result = applyConfirmationAndBudgetGate(result, policy, input)
		if result.Status != VerificationSatisfied {
			return result.Normalize()
		}
	}
	result.Boundaries = AppendBoundaries(result.Boundaries, "intensity_gate_satisfied")
	result.NextHostAction = firstNextHostAction(result.NextHostAction, "host_may_plan_strategy")
	return result.Normalize()
}

func applyStrategyFinalGate(result IntensityGateResult, policy ExecutionIntensityPolicy, strategy StrategyCandidate, input IntensityGateInput) IntensityGateResult {
	if strategy.ID == "" {
		return intensityGateBlock(result, FailureConfigMissing, "strategy candidate is required for final gate", "host:strategy_candidate", "provide_strategy_candidate", "strategy_candidate_missing")
	}
	if executionIntensityRank(strategy.MinIntensity) > executionIntensityRank(result.MaxAllowedIntensity) {
		result.Upgrade = buildIntensityUpgrade(result.MaxAllowedIntensity, strategy.MinIntensity, input.RouteSuggestion.Normalize(), result.EvidenceRefs, "strategy requires higher intensity")
		return intensityGateBlock(result, FailurePolicyBlocked, "strategy minimum intensity exceeds pre-gate allowance", "contract:strategy_intensity", "request_intensity_upgrade_confirmation", "strategy_intensity_exceeds_pre_gate")
	}
	if strategySideEffectDenied(policy, strategy.MinIntensity, strategy.SideEffectClass) {
		return intensityGateBlock(result, FailurePolicyBlocked, "strategy side effect class is denied for intensity", MissingInput("contract:side_effect:"+strategy.SideEffectClass), "select_lower_risk_strategy", "strategy_side_effect_denied")
	}
	result = applyConfirmationAndBudgetGate(result, policy, input)
	if result.Status != VerificationSatisfied {
		return result
	}
	if strategy.RequiresApproval && !input.HostApproved {
		result.RequiresHostApproval = true
		return intensityGateBlock(result, FailureApprovalRequired, "strategy requires host approval", "host:strategy_approval", "request_host_approval", "strategy_requires_host_approval")
	}
	if strategy.RequiresApproval && len(result.ApprovalRefs) == 0 {
		result.RequiresHostApproval = true
		result.Status = VerificationReviewRequired
		result.FailureClass = FailureEvidenceMissing
		result.MissingInputs = AppendMissingInputs(result.MissingInputs, "host:approval_ref")
		result.Boundaries = AppendBoundaries(result.Boundaries, "strategy_approval_ref_missing")
		result.NextHostAction = "provide_host_approval_ref"
		return result
	}
	result.ApprovedIntensity = strategy.MinIntensity
	result.ApprovedControlMode = strategy.ControlMode
	return result
}

func applyConfirmationAndBudgetGate(result IntensityGateResult, policy ExecutionIntensityPolicy, input IntensityGateInput) IntensityGateResult {
	requested := result.RequestedIntensity
	if executionIntensityRank(requested) >= executionIntensityRank(policy.RequireUserConfirmationFrom) {
		result.RequiresUserConfirmation = true
		if !input.UserConfirmed {
			result.Upgrade = buildIntensityUpgrade(policy.MaxDefaultIntensity, requested, input.RouteSuggestion.Normalize(), result.EvidenceRefs, "user confirmation required")
			return intensityGateBlock(result, FailureApprovalRequired, "user confirmation is required for requested intensity", "user:intensity_confirmation", "request_user_confirmation", "user_confirmation_required")
		}
	}
	if executionIntensityRank(requested) >= executionIntensityRank(policy.RequireHostApprovalFrom) {
		result.RequiresHostApproval = true
		if !input.HostApproved {
			result.Upgrade = buildIntensityUpgrade(policy.MaxDefaultIntensity, requested, input.RouteSuggestion.Normalize(), result.EvidenceRefs, "host approval required")
			return intensityGateBlock(result, FailureApprovalRequired, "host approval is required for requested intensity", "host:intensity_approval", "request_host_approval", "host_approval_required")
		}
		if len(result.ApprovalRefs) == 0 {
			result.Status = VerificationReviewRequired
			result.FailureClass = FailureEvidenceMissing
			result.MissingInputs = AppendMissingInputs(result.MissingInputs, "host:approval_ref")
			result.Boundaries = AppendBoundaries(result.Boundaries, "host_approval_ref_missing")
			result.NextHostAction = "provide_host_approval_ref"
			return result
		}
	}
	if executionIntensityRank(requested) >= executionIntensityRank(IntensityL3ManagedObjective) {
		budget := input.Budget.Normalize()
		result.Budget = budget
		if budget.BudgetRef == "" || budget.Limit <= 0 {
			return intensityGateBlock(result, FailureConfigMissing, "budget policy is required for L3+", "contract:budget", "provide_budget_policy", "intensity_budget_missing")
		}
		if budget.Exhausted {
			return intensityGateBlock(result, FailureBudgetExhausted, "budget is exhausted for requested intensity", "contract:budget", "return_partial_or_request_budget", "intensity_budget_exhausted")
		}
	}
	return result
}

func applyExplicitIntensityDenialGate(result IntensityGateResult, policy ExecutionIntensityPolicy, input IntensityGateInput) (IntensityGateResult, bool) {
	requested := result.RequestedIntensity
	if executionIntensityRank(requested) >= executionIntensityRank(policy.RequireUserConfirmationFrom) {
		result.RequiresUserConfirmation = true
		if input.UserDenied {
			return intensityGateBlock(result, FailurePermissionDenied, "user denied requested intensity", "user:intensity_confirmation", "return_partial_or_stop", "user_denied_intensity_upgrade"), true
		}
	}
	if executionIntensityRank(requested) >= executionIntensityRank(policy.RequireHostApprovalFrom) {
		result.RequiresHostApproval = true
		if input.HostDenied {
			return intensityGateBlock(result, FailurePermissionDenied, "host denied requested intensity", "host:intensity_approval", "return_partial_or_stop", "host_denied_intensity_upgrade"), true
		}
	}
	return result, false
}

func intensityGateBlock(result IntensityGateResult, failure FailureClass, reason string, missing MissingInput, next NextHostAction, boundary Boundary) IntensityGateResult {
	result.Status = VerificationBlocked
	result.Allowed = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReason = firstNonEmptyContractString(result.BlockedReason, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.Boundaries = AppendBoundaries(result.Boundaries, boundary)
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	return result.Normalize()
}

func buildIntensityUpgrade(from ExecutionIntensity, to ExecutionIntensity, route IntensityRouteSuggestion, evidence []EvidenceRef, reason string) IntensityUpgrade {
	requestedBy := "host"
	if route.ModelSuggested {
		requestedBy = "model"
	}
	if route.UserExplicit {
		requestedBy = "user"
	}
	if route.HostRouteExplicit {
		requestedBy = "host"
	}
	return IntensityUpgrade{
		FromIntensity:        normalizeIntensityOr(from, IntensityL0AnswerOnly),
		ToIntensity:          NormalizeExecutionIntensity(string(to)),
		RequestedBy:          requestedBy,
		Reason:               reason,
		ApprovalStatus:       HostActionRequiresApproval,
		BudgetStatus:         VerificationNotEvaluated,
		AllowedPartialAction: ObjectiveActionReturnPartial,
		EvidenceRefs:         normalizeEvidenceRefs(evidence),
		Boundaries:           []Boundary{"upgrade_proposal_only", "ask_before_upgrade"},
	}.Normalize()
}

func intensityGateActivation(input Activation, policy Activation) Activation {
	if input != "" {
		return NormalizeActivation(string(input))
	}
	return NormalizeActivation(string(policy))
}

func firstControlMode(values ...ControlMode) ControlMode {
	for _, value := range values {
		normalized := NormalizeControlMode(string(value))
		if normalized != "" {
			return normalized
		}
	}
	return ""
}

func firstIntensity(values ...ExecutionIntensity) ExecutionIntensity {
	for _, value := range values {
		normalized := NormalizeExecutionIntensity(string(value))
		if normalized != "" {
			return normalized
		}
	}
	return IntensityL0AnswerOnly
}

func normalizeIntensityOr(value ExecutionIntensity, fallback ExecutionIntensity) ExecutionIntensity {
	normalized := NormalizeExecutionIntensity(string(value))
	if normalized != "" {
		return normalized
	}
	return NormalizeExecutionIntensity(string(fallback))
}

func executionIntensityRank(value ExecutionIntensity) int {
	switch NormalizeExecutionIntensity(string(value)) {
	case IntensityL0AnswerOnly:
		return 0
	case IntensityL1ToolOnce:
		return 1
	case IntensityL2BoundedToolLoop:
		return 2
	case IntensityL3ManagedObjective:
		return 3
	case IntensityL4DurableLongRun:
		return 4
	case IntensityL5Autonomous:
		return 5
	default:
		return -1
	}
}

func minExecutionIntensity(left ExecutionIntensity, right ExecutionIntensity) ExecutionIntensity {
	if executionIntensityRank(left) <= executionIntensityRank(right) {
		return NormalizeExecutionIntensity(string(left))
	}
	return NormalizeExecutionIntensity(string(right))
}

func intensityDisabled(policy ExecutionIntensityPolicy, intensity ExecutionIntensity) bool {
	for _, disabled := range policy.DisabledIntensities {
		if NormalizeExecutionIntensity(string(disabled)) == NormalizeExecutionIntensity(string(intensity)) {
			return true
		}
	}
	return false
}

func controlModeAllowedForIntensity(policy ExecutionIntensityPolicy, intensity ExecutionIntensity, mode ControlMode) bool {
	normalizedIntensity := NormalizeExecutionIntensity(string(intensity))
	normalizedMode := NormalizeControlMode(string(mode))
	for _, denied := range policy.DeniedControlModesByIntensity[normalizedIntensity] {
		if denied == normalizedMode {
			return false
		}
	}
	allowed := policy.AllowedControlModesByIntensity[normalizedIntensity]
	if len(allowed) == 0 {
		return true
	}
	for _, item := range allowed {
		if item == normalizedMode {
			return true
		}
	}
	return false
}

func strategySideEffectDenied(policy ExecutionIntensityPolicy, intensity ExecutionIntensity, sideEffect string) bool {
	token := normalizeControlToken(sideEffect)
	if token == "" {
		return false
	}
	for _, denied := range policy.DeniedSideEffectsByIntensity[NormalizeExecutionIntensity(string(intensity))] {
		if denied == token {
			return true
		}
	}
	return false
}

func strategyRef(strategy StrategyCandidate) DisplaySafeRef {
	ref, ok := NormalizeDisplaySafeRef(strategy.ID)
	if !ok {
		return ""
	}
	return ref
}

func intensityGateDecisionBasis(stage IntensityGateStage, activation Activation, mode ControlMode, intensity ExecutionIntensity, route IntensityRouteSuggestion, strategy StrategyCandidate, extra []DisplaySafeRef) []DisplaySafeRef {
	basis := []DisplaySafeRef{
		DisplaySafeRef("intensity_gate:" + string(stage)),
		DisplaySafeRef("activation:" + string(activation)),
		DisplaySafeRef("control_mode:" + string(mode)),
		DisplaySafeRef("requested_intensity:" + string(intensity)),
	}
	if route.ModelSuggested {
		basis = append(basis, "route:model_suggested")
	}
	if route.UserExplicit {
		basis = append(basis, "route:user_explicit")
	}
	if route.HostRouteExplicit {
		basis = append(basis, "route:host_explicit")
	}
	if route.ReasonRef != "" {
		basis = append(basis, route.ReasonRef)
	}
	if ref := strategyRef(strategy); ref != "" {
		basis = append(basis, ref)
	}
	basis = append(basis, extra...)
	return normalizeDisplaySafeRefs(basis)
}

func mergePreGateResult(result IntensityGateResult, preGate IntensityGateResult) IntensityGateResult {
	result.MaxAllowedIntensity = firstIntensity(preGate.ApprovedIntensity, preGate.MaxAllowedIntensity, result.MaxAllowedIntensity)
	result.PolicyRefs = normalizeDisplaySafeRefs(append(result.PolicyRefs, preGate.PolicyRefs...))
	result.EvidenceRefs = MergeEvidenceRefs(result.EvidenceRefs, preGate.EvidenceRefs)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, preGate.MissingInputs...)
	result.DecisionBasis = normalizeDisplaySafeRefs(append(result.DecisionBasis, preGate.DecisionBasis...))
	result.Boundaries = AppendBoundaries(MergeBoundaries(result.Boundaries, preGate.Boundaries), "final_gate_uses_pre_gate")
	result.RequiresUserConfirmation = result.RequiresUserConfirmation || preGate.RequiresUserConfirmation
	result.RequiresHostApproval = result.RequiresHostApproval || preGate.RequiresHostApproval
	return result
}

func intensityGateResultInputEmpty(result IntensityGateResult) bool {
	return result.ContractVersion == "" &&
		!result.Projected &&
		result.Stage == "" &&
		result.Activation == "" &&
		result.Status == "" &&
		result.RequestedIntensity == "" &&
		result.MaxAllowedIntensity == "" &&
		result.ApprovedIntensity == "" &&
		len(result.MissingInputs) == 0 &&
		len(result.Boundaries) == 0
}

func cloneExecutionIntensities(in []ExecutionIntensity) []ExecutionIntensity {
	if len(in) == 0 {
		return nil
	}
	return append([]ExecutionIntensity(nil), in...)
}

func normalizeExecutionIntensities(in []ExecutionIntensity) []ExecutionIntensity {
	out := make([]ExecutionIntensity, 0, len(in))
	seen := map[ExecutionIntensity]struct{}{}
	for _, value := range in {
		normalized := NormalizeExecutionIntensity(string(value))
		if normalized == "" {
			continue
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}
	return out
}

func cloneControlModeByIntensity(in map[ExecutionIntensity][]ControlMode) map[ExecutionIntensity][]ControlMode {
	if len(in) == 0 {
		return nil
	}
	out := make(map[ExecutionIntensity][]ControlMode, len(in))
	for key, values := range in {
		out[key] = append([]ControlMode(nil), values...)
	}
	return out
}

func normalizeControlModeByIntensity(in map[ExecutionIntensity][]ControlMode) map[ExecutionIntensity][]ControlMode {
	if len(in) == 0 {
		return nil
	}
	out := map[ExecutionIntensity][]ControlMode{}
	for rawKey, rawValues := range in {
		key := NormalizeExecutionIntensity(string(rawKey))
		if key == "" {
			continue
		}
		values := []ControlMode{}
		seen := map[ControlMode]struct{}{}
		for _, rawValue := range rawValues {
			value := NormalizeControlMode(string(rawValue))
			if value == "" {
				continue
			}
			if _, exists := seen[value]; exists {
				continue
			}
			seen[value] = struct{}{}
			values = append(values, value)
		}
		if len(values) > 0 {
			out[key] = values
		}
	}
	return out
}

func cloneStringByIntensity(in map[ExecutionIntensity][]string) map[ExecutionIntensity][]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[ExecutionIntensity][]string, len(in))
	for key, values := range in {
		out[key] = append([]string(nil), values...)
	}
	return out
}

func normalizeStringByIntensity(in map[ExecutionIntensity][]string) map[ExecutionIntensity][]string {
	if len(in) == 0 {
		return nil
	}
	out := map[ExecutionIntensity][]string{}
	for rawKey, rawValues := range in {
		key := NormalizeExecutionIntensity(string(rawKey))
		if key == "" {
			continue
		}
		values := normalizeStringList(rawValues)
		if len(values) > 0 {
			out[key] = values
		}
	}
	return out
}
