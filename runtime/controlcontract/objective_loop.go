package controlcontract

import "strings"

type ObjectiveControllerState string

const (
	ObjectiveControllerInactive       ObjectiveControllerState = "inactive"
	ObjectiveControllerNeedsContract  ObjectiveControllerState = "needs_contract"
	ObjectiveControllerNeedsAction    ObjectiveControllerState = "needs_action"
	ObjectiveControllerRunning        ObjectiveControllerState = "running"
	ObjectiveControllerPartial        ObjectiveControllerState = "partial"
	ObjectiveControllerBlocked        ObjectiveControllerState = "blocked"
	ObjectiveControllerSatisfied      ObjectiveControllerState = "satisfied"
	ObjectiveControllerFailed         ObjectiveControllerState = "failed"
	ObjectiveControllerReviewRequired ObjectiveControllerState = "review_required"
)

func NormalizeObjectiveControllerState(raw string) ObjectiveControllerState {
	switch normalizeEnumToken(raw) {
	case "inactive", "off", "disabled":
		return ObjectiveControllerInactive
	case "needs_contract", "contract_missing", "missing_contract":
		return ObjectiveControllerNeedsContract
	case "needs_action", "action_required", "needs_host_action":
		return ObjectiveControllerNeedsAction
	case "running", "in_progress", "ready":
		return ObjectiveControllerRunning
	case "partial":
		return ObjectiveControllerPartial
	case "blocked":
		return ObjectiveControllerBlocked
	case "satisfied", "complete", "completed":
		return ObjectiveControllerSatisfied
	case "failed", "failure":
		return ObjectiveControllerFailed
	case "review_required", "needs_review":
		return ObjectiveControllerReviewRequired
	default:
		return ObjectiveControllerInactive
	}
}

type ObjectiveControllerAction string

const (
	ObjectiveActionNone                     ObjectiveControllerAction = "none"
	ObjectiveActionProvideObjectiveContract ObjectiveControllerAction = "provide_objective_contract"
	ObjectiveActionRequestHostApproval      ObjectiveControllerAction = "request_host_approval"
	ObjectiveActionProvideApprovalRef       ObjectiveControllerAction = "provide_approval_ref"
	ObjectiveActionProvideBudgetPolicy      ObjectiveControllerAction = "provide_budget_policy"
	ObjectiveActionProvideStrategyScope     ObjectiveControllerAction = "provide_strategy_scope"
	ObjectiveActionPlanStrategy             ObjectiveControllerAction = "plan_strategy"
	ObjectiveActionRequestReplanDecision    ObjectiveControllerAction = "request_replan_decision"
	ObjectiveActionReturnPartial            ObjectiveControllerAction = "return_partial"
	ObjectiveActionReturnBlocked            ObjectiveControllerAction = "return_blocked"
	ObjectiveActionReturnSatisfied          ObjectiveControllerAction = "return_satisfied"
	ObjectiveActionReturnFailed             ObjectiveControllerAction = "return_failed"
	ObjectiveActionReviewDisplaySafeRefs    ObjectiveControllerAction = "review_display_safe_refs"
)

func NormalizeObjectiveControllerAction(raw string) ObjectiveControllerAction {
	switch normalizeEnumToken(raw) {
	case "", "none":
		return ObjectiveActionNone
	case "provide_objective_contract":
		return ObjectiveActionProvideObjectiveContract
	case "request_host_approval", "request_approval":
		return ObjectiveActionRequestHostApproval
	case "provide_approval_ref":
		return ObjectiveActionProvideApprovalRef
	case "provide_budget_policy":
		return ObjectiveActionProvideBudgetPolicy
	case "provide_strategy_scope":
		return ObjectiveActionProvideStrategyScope
	case "plan_strategy":
		return ObjectiveActionPlanStrategy
	case "request_replan_decision", "replan":
		return ObjectiveActionRequestReplanDecision
	case "return_partial":
		return ObjectiveActionReturnPartial
	case "return_blocked":
		return ObjectiveActionReturnBlocked
	case "return_satisfied":
		return ObjectiveActionReturnSatisfied
	case "return_failed":
		return ObjectiveActionReturnFailed
	case "review_display_safe_refs":
		return ObjectiveActionReviewDisplaySafeRefs
	default:
		return ObjectiveActionNone
	}
}

type ObjectiveBudgetSnapshot struct {
	ContractVersion string           `json:"contract_version,omitempty"`
	BudgetRef       DisplaySafeRef   `json:"budget_ref,omitempty"`
	Limit           int              `json:"limit,omitempty"`
	Used            int              `json:"used,omitempty"`
	Remaining       int              `json:"remaining,omitempty"`
	Exhausted       bool             `json:"exhausted"`
	PolicyRefs      []DisplaySafeRef `json:"policy_refs,omitempty"`
	EvidenceRefs    []EvidenceRef    `json:"evidence_refs,omitempty"`
	MissingInputs   []MissingInput   `json:"missing_inputs,omitempty"`
	Boundaries      []Boundary       `json:"boundaries,omitempty"`
}

func CloneObjectiveBudgetSnapshot(in ObjectiveBudgetSnapshot) ObjectiveBudgetSnapshot {
	out := in
	out.PolicyRefs = cloneDisplaySafeRefs(in.PolicyRefs)
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (b ObjectiveBudgetSnapshot) Clone() ObjectiveBudgetSnapshot {
	return CloneObjectiveBudgetSnapshot(b)
}

func (b ObjectiveBudgetSnapshot) Normalize() ObjectiveBudgetSnapshot {
	out := CloneObjectiveBudgetSnapshot(b)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.BudgetRef = normalizeOneDisplaySafeRef(out.BudgetRef)
	if out.Limit < 0 {
		out.Limit = 0
	}
	if out.Used < 0 {
		out.Used = 0
	}
	if out.Remaining < 0 {
		out.Remaining = 0
	}
	if out.Limit > 0 && out.Remaining == 0 {
		remaining := out.Limit - out.Used
		if remaining < 0 {
			remaining = 0
		}
		out.Remaining = remaining
	}
	out.Exhausted = out.Exhausted || (out.Limit > 0 && out.Used >= out.Limit) || (out.Limit > 0 && out.Remaining == 0)
	out.PolicyRefs = normalizeDisplaySafeRefs(out.PolicyRefs)
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	return out
}

type ObjectiveApprovalState struct {
	ContractVersion string           `json:"contract_version,omitempty"`
	Required        bool             `json:"required"`
	Approved        bool             `json:"approved"`
	ApprovalRefs    []DisplaySafeRef `json:"approval_refs,omitempty"`
	PolicyRefs      []DisplaySafeRef `json:"policy_refs,omitempty"`
	MissingInputs   []MissingInput   `json:"missing_inputs,omitempty"`
	Boundaries      []Boundary       `json:"boundaries,omitempty"`
}

func CloneObjectiveApprovalState(in ObjectiveApprovalState) ObjectiveApprovalState {
	out := in
	out.ApprovalRefs = cloneDisplaySafeRefs(in.ApprovalRefs)
	out.PolicyRefs = cloneDisplaySafeRefs(in.PolicyRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (a ObjectiveApprovalState) Clone() ObjectiveApprovalState {
	return CloneObjectiveApprovalState(a)
}

func (a ObjectiveApprovalState) Normalize() ObjectiveApprovalState {
	out := CloneObjectiveApprovalState(a)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.ApprovalRefs = normalizeDisplaySafeRefs(out.ApprovalRefs)
	out.PolicyRefs = normalizeDisplaySafeRefs(out.PolicyRefs)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	return out
}

type ObjectiveRun struct {
	ContractVersion    string                   `json:"contract_version,omitempty"`
	Projected          bool                     `json:"projected"`
	FullRun            bool                     `json:"full_run"`
	Activation         Activation               `json:"activation,omitempty"`
	State              ObjectiveControllerState `json:"state,omitempty"`
	Frame              ObjectiveFrame           `json:"frame,omitempty"`
	Ledger             AttemptLedgerPatch       `json:"ledger,omitempty"`
	Budget             ObjectiveBudgetSnapshot  `json:"budget,omitempty"`
	Approval           ObjectiveApprovalState   `json:"approval,omitempty"`
	PolicyRefs         []DisplaySafeRef         `json:"policy_refs,omitempty"`
	Strategies         []StrategyCandidate      `json:"strategies,omitempty"`
	CurrentStrategyRef DisplaySafeRef           `json:"current_strategy_ref,omitempty"`
	LatestVerification VerificationResult       `json:"latest_verification,omitempty"`
	LatestObservations []Observation            `json:"latest_observations,omitempty"`
	EvidenceRefs       []EvidenceRef            `json:"evidence_refs,omitempty"`
	FailureClass       FailureClass             `json:"failure_class,omitempty"`
	MissingInputs      []MissingInput           `json:"missing_inputs,omitempty"`
	Boundaries         []Boundary               `json:"boundaries,omitempty"`
	NextHostAction     NextHostAction           `json:"next_host_action,omitempty"`
	RunnerEffect       string                   `json:"runner_effect,omitempty"`
	PromptEffect       string                   `json:"prompt_effect,omitempty"`
	RawOutputLoaded    bool                     `json:"raw_output_loaded"`
}

func CloneObjectiveRun(in ObjectiveRun) ObjectiveRun {
	out := in
	out.Frame = in.Frame.Clone()
	out.Ledger = in.Ledger.Clone()
	out.Budget = in.Budget.Clone()
	out.Approval = in.Approval.Clone()
	out.PolicyRefs = cloneDisplaySafeRefs(in.PolicyRefs)
	out.Strategies = cloneStrategyCandidates(in.Strategies)
	out.LatestVerification = in.LatestVerification.Clone()
	out.LatestObservations = cloneObservations(in.LatestObservations)
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (r ObjectiveRun) Clone() ObjectiveRun {
	return CloneObjectiveRun(r)
}

func (r ObjectiveRun) Normalize() ObjectiveRun {
	out := CloneObjectiveRun(r)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Activation = NormalizeActivation(string(out.Activation))
	out.State = NormalizeObjectiveControllerState(string(out.State))
	out.Frame = normalizeObjectiveRunFrame(out.Frame)
	out.Ledger = out.Ledger.Normalize()
	out.Budget = out.Budget.Normalize()
	out.Approval = out.Approval.Normalize()
	out.PolicyRefs = normalizeDisplaySafeRefs(out.PolicyRefs)
	out.Strategies = normalizeStrategyCandidates(out.Strategies)
	out.CurrentStrategyRef = normalizeOneDisplaySafeRef(out.CurrentStrategyRef)
	out.LatestVerification = out.LatestVerification.Normalize()
	out.LatestObservations = normalizeObservations(out.LatestObservations)
	out.EvidenceRefs = MergeEvidenceRefs(out.EvidenceRefs, out.LatestVerification.EvidenceRefs, out.Ledger.EvidenceRefs, objectiveObservationEvidenceRefs(out.LatestObservations))
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
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
	if out.RawOutputLoaded && out.State != ObjectiveControllerBlocked {
		out.State = ObjectiveControllerReviewRequired
		if out.FailureClass == FailureNone {
			out.FailureClass = FailureEvidenceWeak
		}
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		if out.NextHostAction == "" {
			out.NextHostAction = "provide_display_safe_refs"
		}
	}
	out.FullRun = out.FullRun && out.Activation == ActivationManaged && out.State != ObjectiveControllerInactive
	return out
}

type ObjectiveRunInput struct {
	Activation         Activation              `json:"activation,omitempty"`
	Frame              ObjectiveFrame          `json:"frame,omitempty"`
	Ledger             AttemptLedgerPatch      `json:"ledger,omitempty"`
	Budget             ObjectiveBudgetSnapshot `json:"budget,omitempty"`
	Approval           ObjectiveApprovalState  `json:"approval,omitempty"`
	PolicyRefs         []DisplaySafeRef        `json:"policy_refs,omitempty"`
	Strategies         []StrategyCandidate     `json:"strategies,omitempty"`
	CurrentStrategyRef DisplaySafeRef          `json:"current_strategy_ref,omitempty"`
	Verification       VerificationResult      `json:"verification,omitempty"`
	Observations       []Observation           `json:"observations,omitempty"`
	EvidenceRefs       []EvidenceRef           `json:"evidence_refs,omitempty"`
	Boundaries         []Boundary              `json:"boundaries,omitempty"`
	RawOutputLoaded    bool                    `json:"raw_output_loaded"`
}

type ObjectiveControllerDecision struct {
	ContractVersion    string                    `json:"contract_version,omitempty"`
	Projected          bool                      `json:"projected"`
	Activation         Activation                `json:"activation,omitempty"`
	State              ObjectiveControllerState  `json:"state,omitempty"`
	Action             ObjectiveControllerAction `json:"action,omitempty"`
	ObjectiveID        string                    `json:"objective_id,omitempty"`
	CurrentStrategyRef DisplaySafeRef            `json:"current_strategy_ref,omitempty"`
	AllowedStrategies  []StrategyCandidate       `json:"allowed_strategies,omitempty"`
	Verification       VerificationResult        `json:"verification,omitempty"`
	EvidenceRefs       []EvidenceRef             `json:"evidence_refs,omitempty"`
	FailureClass       FailureClass              `json:"failure_class,omitempty"`
	Reason             string                    `json:"reason,omitempty"`
	MissingInputs      []MissingInput            `json:"missing_inputs,omitempty"`
	DecisionBasis      []DisplaySafeRef          `json:"decision_basis,omitempty"`
	Boundaries         []Boundary                `json:"boundaries,omitempty"`
	NextHostAction     NextHostAction            `json:"next_host_action,omitempty"`
	RunnerEffect       string                    `json:"runner_effect,omitempty"`
	PromptEffect       string                    `json:"prompt_effect,omitempty"`
	RawOutputLoaded    bool                      `json:"raw_output_loaded"`
}

func CloneObjectiveControllerDecision(in ObjectiveControllerDecision) ObjectiveControllerDecision {
	out := in
	out.AllowedStrategies = cloneStrategyCandidates(in.AllowedStrategies)
	out.Verification = in.Verification.Clone()
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.DecisionBasis = cloneDisplaySafeRefs(in.DecisionBasis)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (d ObjectiveControllerDecision) Clone() ObjectiveControllerDecision {
	return CloneObjectiveControllerDecision(d)
}

func (d ObjectiveControllerDecision) Normalize() ObjectiveControllerDecision {
	out := CloneObjectiveControllerDecision(d)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Activation = NormalizeActivation(string(out.Activation))
	out.State = NormalizeObjectiveControllerState(string(out.State))
	out.Action = NormalizeObjectiveControllerAction(string(out.Action))
	out.ObjectiveID = strings.TrimSpace(out.ObjectiveID)
	out.CurrentStrategyRef = normalizeOneDisplaySafeRef(out.CurrentStrategyRef)
	out.AllowedStrategies = normalizeStrategyCandidates(out.AllowedStrategies)
	out.Verification = out.Verification.Normalize()
	out.EvidenceRefs = MergeEvidenceRefs(out.EvidenceRefs, out.Verification.EvidenceRefs)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.Reason = managedObjectiveReplannerSafeReason(out.Reason)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.DecisionBasis = normalizeDisplaySafeRefs(out.DecisionBasis)
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
	if out.RawOutputLoaded && out.State != ObjectiveControllerBlocked {
		out.State = ObjectiveControllerReviewRequired
		out.Action = ObjectiveActionReviewDisplaySafeRefs
		if out.FailureClass == FailureNone {
			out.FailureClass = FailureEvidenceWeak
		}
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		out.NextHostAction = "provide_display_safe_refs"
	}
	return out
}

type ObjectiveControllerInput struct {
	Run                ObjectiveRun            `json:"run,omitempty"`
	Activation         Activation              `json:"activation,omitempty"`
	Frame              ObjectiveFrame          `json:"frame,omitempty"`
	Ledger             AttemptLedgerPatch      `json:"ledger,omitempty"`
	Budget             ObjectiveBudgetSnapshot `json:"budget,omitempty"`
	Approval           ObjectiveApprovalState  `json:"approval,omitempty"`
	PolicyRefs         []DisplaySafeRef        `json:"policy_refs,omitempty"`
	Strategies         []StrategyCandidate     `json:"strategies,omitempty"`
	CurrentStrategyRef DisplaySafeRef          `json:"current_strategy_ref,omitempty"`
	Verification       VerificationResult      `json:"verification,omitempty"`
	Observations       []Observation           `json:"observations,omitempty"`
	EvidenceRefs       []EvidenceRef           `json:"evidence_refs,omitempty"`
	Boundaries         []Boundary              `json:"boundaries,omitempty"`
	RawOutputLoaded    bool                    `json:"raw_output_loaded"`
}

func BuildObjectiveRun(input ObjectiveRunInput) ObjectiveRun {
	activation := NormalizeActivation(string(input.Activation))
	frame := normalizeObjectiveRunFrame(input.Frame)
	ledger := input.Ledger.Normalize()
	if ledger.ObjectiveID == "" {
		ledger.ObjectiveID = frame.ID
		ledger = ledger.Normalize()
	}
	run := ObjectiveRun{
		ContractVersion:    ContractVersion,
		Projected:          true,
		FullRun:            activation == ActivationManaged,
		Activation:         activation,
		State:              ObjectiveControllerRunning,
		Frame:              frame,
		Ledger:             ledger,
		Budget:             input.Budget.Normalize(),
		Approval:           input.Approval.Normalize(),
		PolicyRefs:         normalizeDisplaySafeRefs(input.PolicyRefs),
		Strategies:         normalizeStrategyCandidates(input.Strategies),
		CurrentStrategyRef: normalizeOneDisplaySafeRef(input.CurrentStrategyRef),
		LatestVerification: input.Verification.Normalize(),
		LatestObservations: normalizeObservations(input.Observations),
		EvidenceRefs:       normalizeEvidenceRefs(input.EvidenceRefs),
		FailureClass:       FailureNone,
		Boundaries: MergeBoundaries(
			[]Boundary{
				"objective_run_projection_only",
				"no_runner_dispatch",
				"no_strategy_dispatch",
				"no_runtime_adapter_execution",
				"no_tool_execution",
				"no_workflow_dispatch",
				"no_install_or_schedule_apply",
				"model_route_is_not_authorization",
			},
			input.Boundaries,
		),
		RunnerEffect:    "none",
		PromptEffect:    "none",
		RawOutputLoaded: input.RawOutputLoaded,
	}
	if activation != ActivationManaged {
		run.FullRun = false
		run.State = ObjectiveControllerInactive
		run.FailureClass = FailurePolicyBlocked
		run.MissingInputs = AppendMissingInputs(run.MissingInputs, "control_plane:managed_activation")
		run.Boundaries = AppendBoundaries(run.Boundaries, "objective_controller_activation_required")
		run.NextHostAction = "enable_managed_objective"
		return run.Normalize()
	}
	run.MissingInputs = objectiveRunMissingContractInputs(run)
	if len(run.MissingInputs) > 0 {
		run.State = ObjectiveControllerNeedsContract
		run.FailureClass = FailureConfigMissing
		run.NextHostAction = "provide_objective_contract"
		return run.Normalize()
	}
	if run.Budget.Exhausted {
		run.State = ObjectiveControllerBlocked
		run.FailureClass = FailureBudgetExhausted
		run.Boundaries = AppendBoundaries(run.Boundaries, "objective_budget_exhausted")
		run.NextHostAction = "return_partial_or_request_budget"
		return run.Normalize()
	}
	if run.Approval.Required && !run.Approval.Approved {
		run.State = ObjectiveControllerNeedsAction
		run.FailureClass = FailureApprovalRequired
		run.MissingInputs = AppendMissingInputs(run.MissingInputs, "host:objective_approval")
		run.Boundaries = AppendBoundaries(run.Boundaries, "objective_requires_host_approval")
		run.NextHostAction = "request_host_approval"
		return run.Normalize()
	}
	if run.Approval.Required && len(run.Approval.ApprovalRefs) == 0 {
		run.State = ObjectiveControllerReviewRequired
		run.FailureClass = FailureEvidenceMissing
		run.MissingInputs = AppendMissingInputs(run.MissingInputs, "host:approval_ref")
		run.Boundaries = AppendBoundaries(run.Boundaries, "objective_approval_ref_missing")
		run.NextHostAction = "provide_host_approval_ref"
		return run.Normalize()
	}
	if len(run.Strategies) == 0 {
		run.State = ObjectiveControllerNeedsAction
		run.FailureClass = FailureConfigMissing
		run.MissingInputs = AppendMissingInputs(run.MissingInputs, "host:strategy_scope")
		run.Boundaries = AppendBoundaries(run.Boundaries, "objective_strategy_scope_missing")
		run.NextHostAction = "provide_strategy_scope"
		return run.Normalize()
	}
	run.State = objectiveRunStateFromVerification(run.LatestVerification)
	run.FailureClass = firstFailureClass(run.LatestVerification.FailureClass, FailureNone)
	run.NextHostAction = objectiveRunNextHostAction(run.State)
	return run.Normalize()
}

func BuildObjectiveControllerDecision(input ObjectiveControllerInput) ObjectiveControllerDecision {
	var run ObjectiveRun
	if objectiveRunInputEmpty(input.Run) {
		run = BuildObjectiveRun(ObjectiveRunInput{
			Activation:         input.Activation,
			Frame:              input.Frame,
			Ledger:             input.Ledger,
			Budget:             input.Budget,
			Approval:           input.Approval,
			PolicyRefs:         input.PolicyRefs,
			Strategies:         input.Strategies,
			CurrentStrategyRef: input.CurrentStrategyRef,
			Verification:       input.Verification,
			Observations:       input.Observations,
			EvidenceRefs:       input.EvidenceRefs,
			Boundaries:         input.Boundaries,
			RawOutputLoaded:    input.RawOutputLoaded,
		})
	} else {
		run = input.Run.Normalize()
	}
	decision := ObjectiveControllerDecision{
		ContractVersion:    ContractVersion,
		Projected:          true,
		Activation:         run.Activation,
		State:              run.State,
		Action:             objectiveActionForRun(run),
		ObjectiveID:        run.Frame.ID,
		CurrentStrategyRef: run.CurrentStrategyRef,
		AllowedStrategies:  cloneStrategyCandidates(run.Strategies),
		Verification:       run.LatestVerification,
		EvidenceRefs:       cloneEvidenceRefs(run.EvidenceRefs),
		FailureClass:       run.FailureClass,
		Reason:             objectiveDecisionReason(run),
		MissingInputs:      cloneMissingInputs(run.MissingInputs),
		DecisionBasis:      objectiveDecisionBasis(run),
		Boundaries: AppendBoundaries(
			run.Boundaries,
			"objective_controller_decision",
			"decision_only",
			"host_must_apply_controller_decision",
		),
		NextHostAction:  run.NextHostAction,
		RunnerEffect:    "none",
		PromptEffect:    "none",
		RawOutputLoaded: run.RawOutputLoaded || input.RawOutputLoaded,
	}
	return decision.Normalize()
}

func normalizeObjectiveRunFrame(frame ObjectiveFrame) ObjectiveFrame {
	out := frame.Normalize()
	if out.ControlMode == "" {
		out.ControlMode = ControlModeObjective
	}
	if out.Intensity == "" {
		out.Intensity = IntensityL3ManagedObjective
	}
	out.Boundaries = AppendBoundaries(out.Boundaries, "objective_frame")
	return out.Normalize()
}

func objectiveRunMissingContractInputs(run ObjectiveRun) []MissingInput {
	out := []MissingInput{}
	if strings.TrimSpace(run.Frame.ID) == "" {
		out = AppendMissingInputs(out, "control_plane:objective_frame")
	}
	if strings.TrimSpace(run.Frame.UserGoalDigest) == "" {
		out = AppendMissingInputs(out, "control_plane:user_goal_digest")
	}
	if len(run.Frame.SuccessCriteria) == 0 {
		out = AppendMissingInputs(out, "host:success_criteria")
	}
	if run.Ledger.LedgerRef == "" {
		out = AppendMissingInputs(out, "host:objective_ledger_ref")
	}
	if run.Budget.BudgetRef == "" || run.Budget.Limit <= 0 {
		out = AppendMissingInputs(out, "contract:budget")
	}
	return AppendMissingInputs(out, run.Frame.MissingInputs...)
}

func objectiveRunStateFromVerification(verification VerificationResult) ObjectiveControllerState {
	switch verification.Normalize().Status {
	case VerificationSatisfied, VerificationNotApplicable:
		return ObjectiveControllerSatisfied
	case VerificationPartial:
		return ObjectiveControllerPartial
	case VerificationBlocked:
		return ObjectiveControllerBlocked
	case VerificationFailed:
		return ObjectiveControllerFailed
	case VerificationReviewRequired:
		return ObjectiveControllerReviewRequired
	default:
		return ObjectiveControllerRunning
	}
}

func objectiveRunNextHostAction(state ObjectiveControllerState) NextHostAction {
	switch state {
	case ObjectiveControllerSatisfied:
		return "return_satisfied"
	case ObjectiveControllerPartial:
		return "request_host_replanner_decision"
	case ObjectiveControllerBlocked:
		return "return_blocked"
	case ObjectiveControllerFailed:
		return "return_failed"
	case ObjectiveControllerReviewRequired:
		return "review_objective_evidence"
	case ObjectiveControllerRunning:
		return "host_may_select_strategy"
	default:
		return "none"
	}
}

func objectiveActionForRun(run ObjectiveRun) ObjectiveControllerAction {
	switch run.State {
	case ObjectiveControllerInactive:
		return ObjectiveActionNone
	case ObjectiveControllerNeedsContract:
		if run.Budget.BudgetRef == "" || run.Budget.Limit <= 0 {
			return ObjectiveActionProvideBudgetPolicy
		}
		return ObjectiveActionProvideObjectiveContract
	case ObjectiveControllerNeedsAction:
		if missingInputContains(run.MissingInputs, "host:objective_approval") {
			return ObjectiveActionRequestHostApproval
		}
		if missingInputContains(run.MissingInputs, "host:strategy_scope") {
			return ObjectiveActionProvideStrategyScope
		}
		return ObjectiveActionProvideObjectiveContract
	case ObjectiveControllerRunning:
		return ObjectiveActionPlanStrategy
	case ObjectiveControllerPartial:
		return ObjectiveActionRequestReplanDecision
	case ObjectiveControllerBlocked:
		return ObjectiveActionReturnBlocked
	case ObjectiveControllerSatisfied:
		return ObjectiveActionReturnSatisfied
	case ObjectiveControllerFailed:
		return ObjectiveActionReturnFailed
	case ObjectiveControllerReviewRequired:
		if missingInputContains(run.MissingInputs, "host:approval_ref") {
			return ObjectiveActionProvideApprovalRef
		}
		return ObjectiveActionReviewDisplaySafeRefs
	default:
		return ObjectiveActionNone
	}
}

func objectiveDecisionReason(run ObjectiveRun) string {
	switch run.State {
	case ObjectiveControllerInactive:
		return "managed objective activation is required"
	case ObjectiveControllerNeedsContract:
		return "objective contract inputs are missing"
	case ObjectiveControllerNeedsAction:
		return "host action is required before objective can continue"
	case ObjectiveControllerRunning:
		return "objective is ready for strategy planning"
	case ObjectiveControllerPartial:
		return "verification is partial and requires replanner decision"
	case ObjectiveControllerBlocked:
		return firstNonEmptyContractString(run.LatestVerification.FailureReason, string(run.FailureClass), "objective is blocked")
	case ObjectiveControllerSatisfied:
		return "objective verification is satisfied"
	case ObjectiveControllerFailed:
		return firstNonEmptyContractString(run.LatestVerification.FailureReason, "objective verification failed")
	case ObjectiveControllerReviewRequired:
		return firstNonEmptyContractString(run.LatestVerification.FailureReason, "objective evidence requires review")
	default:
		return ""
	}
}

func objectiveDecisionBasis(run ObjectiveRun) []DisplaySafeRef {
	basis := []DisplaySafeRef{
		DisplaySafeRef("activation:" + string(run.Activation)),
		DisplaySafeRef("state:" + string(run.State)),
		DisplaySafeRef("intensity:" + string(run.Frame.Intensity)),
		DisplaySafeRef("verification_status:" + string(run.LatestVerification.Status)),
	}
	if run.Budget.BudgetRef != "" {
		basis = append(basis, run.Budget.BudgetRef)
	}
	if run.CurrentStrategyRef != "" {
		basis = append(basis, run.CurrentStrategyRef)
	}
	return normalizeDisplaySafeRefs(basis)
}

func objectiveObservationEvidenceRefs(observations []Observation) []EvidenceRef {
	out := []EvidenceRef{}
	for _, observation := range normalizeObservations(observations) {
		out = append(out, observation.EvidenceRefs...)
	}
	return normalizeEvidenceRefs(out)
}

func missingInputContains(values []MissingInput, want MissingInput) bool {
	normalized := normalizeMissingInputs(values)
	target := normalizeMissingInputs([]MissingInput{want})
	if len(target) == 0 {
		return false
	}
	for _, value := range normalized {
		if value == target[0] {
			return true
		}
	}
	return false
}

func objectiveRunInputEmpty(run ObjectiveRun) bool {
	return run.ContractVersion == "" &&
		!run.Projected &&
		!run.FullRun &&
		run.Activation == "" &&
		run.State == "" &&
		run.Frame.ID == "" &&
		run.Ledger.LedgerRef == "" &&
		run.Budget.BudgetRef == "" &&
		len(run.Strategies) == 0 &&
		run.LatestVerification.Status == "" &&
		len(run.LatestObservations) == 0 &&
		len(run.EvidenceRefs) == 0 &&
		len(run.MissingInputs) == 0 &&
		len(run.Boundaries) == 0
}
