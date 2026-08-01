package controlcontract

type ObjectiveReplannerAction string

const (
	ObjectiveReplannerActionNone                      ObjectiveReplannerAction = "none"
	ObjectiveReplannerActionRetrySameStrategy         ObjectiveReplannerAction = "retry_same_strategy"
	ObjectiveReplannerActionSwitchStrategy            ObjectiveReplannerAction = "switch_strategy"
	ObjectiveReplannerActionEnterCapabilityResolution ObjectiveReplannerAction = "enter_capability_resolution"
	ObjectiveReplannerActionRequestApproval           ObjectiveReplannerAction = "request_approval"
	ObjectiveReplannerActionReturnPartial             ObjectiveReplannerAction = "return_partial"
	ObjectiveReplannerActionReturnBlocked             ObjectiveReplannerAction = "return_blocked"
	ObjectiveReplannerActionReturnSatisfied           ObjectiveReplannerAction = "return_satisfied"
	ObjectiveReplannerActionReviewDisplaySafeRefs     ObjectiveReplannerAction = "review_display_safe_refs"
)

func NormalizeObjectiveReplannerAction(raw string) ObjectiveReplannerAction {
	switch normalizeEnumToken(raw) {
	case "", "none":
		return ObjectiveReplannerActionNone
	case "retry_same_strategy", "retry":
		return ObjectiveReplannerActionRetrySameStrategy
	case "switch_strategy", "switch":
		return ObjectiveReplannerActionSwitchStrategy
	case "enter_capability_resolution", "capability_resolution", "capability":
		return ObjectiveReplannerActionEnterCapabilityResolution
	case "request_approval", "request_host_approval":
		return ObjectiveReplannerActionRequestApproval
	case "return_partial", "partial":
		return ObjectiveReplannerActionReturnPartial
	case "return_blocked", "blocked":
		return ObjectiveReplannerActionReturnBlocked
	case "return_satisfied", "satisfied":
		return ObjectiveReplannerActionReturnSatisfied
	case "review_display_safe_refs", "display_safe_refs":
		return ObjectiveReplannerActionReviewDisplaySafeRefs
	default:
		return ObjectiveReplannerActionNone
	}
}

type ObjectiveReplannerDecisionInput struct {
	Activation         Activation                      `json:"activation,omitempty"`
	Controller         ObjectiveControllerDecision     `json:"controller,omitempty"`
	Verification       ObjectiveVerificationGateResult `json:"verification,omitempty"`
	StrategyPlan       StrategyPlannerResult           `json:"strategy_plan,omitempty"`
	CurrentStrategyRef DisplaySafeRef                  `json:"current_strategy_ref,omitempty"`
	Attempts           []AttemptSummary                `json:"attempts,omitempty"`
	Budget             ObjectiveBudgetSnapshot         `json:"budget,omitempty"`
	Approval           ObjectiveApprovalState          `json:"approval,omitempty"`
	EvidenceRefs       []EvidenceRef                   `json:"evidence_refs,omitempty"`
	Boundaries         []Boundary                      `json:"boundaries,omitempty"`
	RawOutputLoaded    bool                            `json:"raw_output_loaded"`
}

type ObjectiveReplannerDecision struct {
	ContractVersion    string                   `json:"contract_version,omitempty"`
	Projected          bool                     `json:"projected"`
	Activation         Activation               `json:"activation,omitempty"`
	Status             VerificationStatus       `json:"status,omitempty"`
	Action             ObjectiveReplannerAction `json:"action,omitempty"`
	ObjectiveID        string                   `json:"objective_id,omitempty"`
	CurrentStrategyRef DisplaySafeRef           `json:"current_strategy_ref,omitempty"`
	NextStrategyRef    DisplaySafeRef           `json:"next_strategy_ref,omitempty"`
	SelectedStrategy   StrategyCandidate        `json:"selected_strategy,omitempty"`
	CandidateRefs      []DisplaySafeRef         `json:"candidate_refs,omitempty"`
	Verification       VerificationResult       `json:"verification,omitempty"`
	EvidenceRefs       []EvidenceRef            `json:"evidence_refs,omitempty"`
	FailureClass       FailureClass             `json:"failure_class,omitempty"`
	FailureReason      string                   `json:"failure_reason,omitempty"`
	MissingInputs      []MissingInput           `json:"missing_inputs,omitempty"`
	DecisionBasis      []DisplaySafeRef         `json:"decision_basis,omitempty"`
	Boundaries         []Boundary               `json:"boundaries,omitempty"`
	NextHostAction     NextHostAction           `json:"next_host_action,omitempty"`
	RunnerEffect       string                   `json:"runner_effect,omitempty"`
	PromptEffect       string                   `json:"prompt_effect,omitempty"`
	RawOutputLoaded    bool                     `json:"raw_output_loaded"`
}

func BuildObjectiveReplannerDecision(input ObjectiveReplannerDecisionInput) ObjectiveReplannerDecision {
	controller := input.Controller.Normalize()
	verificationGate := input.Verification.Normalize()
	strategyPlan := input.StrategyPlan.Normalize()
	budget := input.Budget.Normalize()
	approval := input.Approval.Normalize()
	activation := objectiveReplannerActivation(input.Activation, controller.Activation, strategyPlan.Activation)
	verification := objectiveReplannerVerification(controller.Verification, verificationGate.Verification)
	current := firstDisplaySafeRef(input.CurrentStrategyRef, controller.CurrentStrategyRef, strategyPlan.CurrentStrategyRef)
	candidates := objectiveReplannerCandidates(controller.AllowedStrategies, strategyPlan.RankedCandidates)
	objectiveID := firstNonEmptyContractString(controller.ObjectiveID, verificationGate.Frame.ID, strategyPlan.Frame.ID)
	maxAllowedIntensity := firstIntensity(strategyPlan.MaxAllowedIntensity, verificationGate.Frame.Intensity, IntensityL3ManagedObjective)
	result := ObjectiveReplannerDecision{
		ContractVersion:    ContractVersion,
		Projected:          true,
		Activation:         activation,
		Status:             VerificationBlocked,
		Action:             ObjectiveReplannerActionReturnBlocked,
		ObjectiveID:        objectiveID,
		CurrentStrategyRef: current,
		CandidateRefs:      objectiveReplannerCandidateRefs(candidates),
		Verification:       verification,
		EvidenceRefs:       MergeEvidenceRefs(input.EvidenceRefs, controller.EvidenceRefs, verification.EvidenceRefs, verificationGate.EvidenceRefs, strategyPlan.EvidenceRefs),
		FailureClass:       firstFailureClass(controller.FailureClass, verification.FailureClass, strategyPlan.FailureClass),
		FailureReason:      managedObjectiveReplannerSafeReason(firstNonEmptyContractString(controller.Reason, verification.FailureReason, verificationGate.FailureReason)),
		MissingInputs:      MergeMissingInputs(controller.MissingInputs, verification.MissingInputs, verificationGate.MissingInputs, strategyPlan.MissingInputs),
		DecisionBasis: normalizeDisplaySafeRefs(append(
			append(cloneDisplaySafeRefs(controller.DecisionBasis), strategyPlan.DecisionBasis...),
			DisplaySafeRef("verification_status:"+string(verification.Status)),
			DisplaySafeRef("replanner_decision_only"),
		)),
		Boundaries: MergeBoundaries(
			[]Boundary{
				"objective_replanner_decision",
				"decision_only",
				"projection_only",
				"no_runner_dispatch",
				"no_strategy_dispatch",
				"no_runtime_adapter_execution",
				"no_install_or_schedule_apply",
				"host_must_apply_replanner_decision",
			},
			input.Boundaries,
			controller.Boundaries,
			verificationGate.Boundaries,
			strategyPlan.Boundaries,
		),
		NextHostAction:  "review_replanner_decision",
		RunnerEffect:    "none",
		PromptEffect:    "none",
		RawOutputLoaded: input.RawOutputLoaded || controller.RawOutputLoaded || verificationGate.RawOutputLoaded || strategyPlan.RawOutputLoaded,
	}
	if objectiveReplannerDecisionInputUnsafe(input) {
		result.FailureClass = FailureNone
		return objectiveReplannerDecisionBlock(result, VerificationReviewRequired, ObjectiveReplannerActionReviewDisplaySafeRefs, FailureEvidenceWeak, "host:display_safe_refs", "provide_display_safe_refs", "raw_output_not_allowed")
	}
	if activation != ActivationManaged {
		result.FailureClass = FailureNone
		return objectiveReplannerDecisionBlock(result, VerificationBlocked, ObjectiveReplannerActionNone, FailurePolicyBlocked, "control_plane:managed_activation", "enable_managed_objective", "objective_replanner_requires_managed_activation")
	}
	if verification.Status == VerificationSatisfied || verification.Status == VerificationNotApplicable {
		result.Status = VerificationSatisfied
		result.Action = ObjectiveReplannerActionReturnSatisfied
		result.FailureClass = FailureNone
		result.NextHostAction = "return_satisfied"
		result.Boundaries = AppendBoundaries(result.Boundaries, "objective_already_satisfied")
		return result.Normalize()
	}
	if budget.Exhausted {
		return objectiveReplannerBudgetDecision(result)
	}
	if objectiveReplannerCredentialRequired(verification, controller, approval) {
		return objectiveReplannerDecisionBlock(result, VerificationReviewRequired, ObjectiveReplannerActionRequestApproval, firstFailureClass(result.FailureClass, FailureCredentialMissing), objectiveReplannerCredentialMissingInput(verification, controller, approval), "request_host_credential", "objective_replanner_requires_credential")
	}
	if objectiveReplannerApprovalRequired(verification, controller, approval) {
		return objectiveReplannerDecisionBlock(result, VerificationReviewRequired, ObjectiveReplannerActionRequestApproval, FailureApprovalRequired, objectiveReplannerApprovalMissingInput(verification, controller, approval), "request_host_approval", "objective_replanner_requires_approval")
	}
	if objectiveReplannerCapabilityResolutionRequired(verification, controller, result.MissingInputs) {
		return objectiveReplannerDecisionBlock(result, VerificationBlocked, ObjectiveReplannerActionEnterCapabilityResolution, firstFailureClass(result.FailureClass, FailureCapabilityMissing), objectiveReplannerCapabilityMissingInput(result.MissingInputs), "enter_capability_resolution", "capability_gap_proposal_only")
	}
	if objectiveReplannerPolicyBlocked(verification, controller) {
		return objectiveReplannerPolicyBlockedDecision(result)
	}
	noProgressGate := BuildObjectiveNoProgressSwitchGate(ObjectiveNoProgressSwitchGateInput{
		GateRef:            "gate:objective_replanner_no_progress",
		CurrentStrategyRef: current,
		Attempts:           input.Attempts,
		Verification:       verification,
		Candidates:         candidates,
		Threshold:          2,
	})
	noProgress := noProgressGate.RepeatedNoProgress
	if noProgress && !noProgressGate.ReadyForSwitch {
		result.Boundaries = MergeBoundaries(result.Boundaries, noProgressGate.Boundaries)
		result.DecisionBasis = appendUniqueDisplaySafeRef(result.DecisionBasis, "replanner:no_progress_gate")
		result.FailureClass = FailureNone
		return objectiveReplannerDecisionBlock(result, VerificationBlocked, ObjectiveReplannerActionReturnBlocked, FailureRepeatedNoProgress, "host:new_evidence_or_strategy", "return_blocked", "objective_replanner_repeated_no_progress")
	}
	selected, hasSelected := objectiveReplannerSelectStrategy(current, candidates, verification.Status, firstFailureClass(verification.FailureClass, controller.FailureClass), maxAllowedIntensity, noProgress)
	if !hasSelected {
		return objectiveReplannerNoCandidateDecision(result)
	}
	nextRef := objectiveReplannerCandidateRef(selected)
	if objectiveReplannerCrossStrategyBlocked(maxAllowedIntensity, current, nextRef) {
		result.SelectedStrategy = selected
		result.NextStrategyRef = nextRef
		result.FailureClass = FailureNone
		return objectiveReplannerDecisionBlock(result, VerificationPartial, ObjectiveReplannerActionReturnPartial, FailurePolicyBlocked, "contract:l2_same_strategy", "request_intensity_upgrade_confirmation", "l2_cross_strategy_blocked")
	}
	result.SelectedStrategy = selected
	result.NextStrategyRef = nextRef
	result.Status = VerificationPartial
	result.FailureClass = firstFailureClass(result.FailureClass, FailureVerificationFailed)
	switch {
	case nextRef != "" && current != "" && nextRef != current:
		result.Action = ObjectiveReplannerActionSwitchStrategy
		result.NextHostAction = "host_may_switch_strategy"
		result.DecisionBasis = appendUniqueDisplaySafeRef(result.DecisionBasis, "replanner:switch_strategy")
		result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_host_strategy_switch")
		if noProgress {
			result.FailureClass = FailureRepeatedNoProgress
			result.DecisionBasis = appendUniqueDisplaySafeRef(result.DecisionBasis, "replanner:no_progress_switch_strategy")
			result.DecisionBasis = appendUniqueDisplaySafeRef(result.DecisionBasis, "replanner:no_progress_gate")
			result.Boundaries = MergeBoundaries(result.Boundaries, noProgressGate.Boundaries)
		} else if objectiveReplannerSourceUnavailable(verification, controller) {
			result.DecisionBasis = appendUniqueDisplaySafeRef(result.DecisionBasis, "replanner:source_unavailable_switch_strategy")
			result.Boundaries = AppendBoundaries(result.Boundaries, "objective_replanner_source_unavailable_switch_strategy")
		}
	default:
		result.Action = ObjectiveReplannerActionRetrySameStrategy
		result.NextHostAction = "host_may_retry_same_strategy"
		result.DecisionBasis = appendUniqueDisplaySafeRef(result.DecisionBasis, "replanner:retry_same_strategy")
		result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_host_same_strategy_retry")
	}
	return result.Normalize()
}

func CloneObjectiveReplannerDecision(in ObjectiveReplannerDecision) ObjectiveReplannerDecision {
	out := in
	out.SelectedStrategy = in.SelectedStrategy.Clone()
	out.CandidateRefs = cloneDisplaySafeRefs(in.CandidateRefs)
	out.Verification = in.Verification.Clone()
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.DecisionBasis = cloneDisplaySafeRefs(in.DecisionBasis)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (d ObjectiveReplannerDecision) Clone() ObjectiveReplannerDecision {
	return CloneObjectiveReplannerDecision(d)
}

func (d ObjectiveReplannerDecision) Normalize() ObjectiveReplannerDecision {
	out := CloneObjectiveReplannerDecision(d)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Activation = NormalizeActivation(string(out.Activation))
	out.Status = NormalizeVerificationStatus(string(out.Status))
	out.Action = NormalizeObjectiveReplannerAction(string(out.Action))
	out.ObjectiveID = firstNonEmptyContractString(out.ObjectiveID)
	out.CurrentStrategyRef = normalizeOneDisplaySafeRef(out.CurrentStrategyRef)
	out.NextStrategyRef = normalizeOneDisplaySafeRef(out.NextStrategyRef)
	out.SelectedStrategy = out.SelectedStrategy.Normalize()
	out.CandidateRefs = normalizeDisplaySafeRefs(out.CandidateRefs)
	out.Verification = out.Verification.Normalize()
	out.EvidenceRefs = MergeEvidenceRefs(out.EvidenceRefs, out.Verification.EvidenceRefs)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.FailureReason = managedObjectiveReplannerSafeReason(out.FailureReason)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.DecisionBasis = normalizeDisplaySafeRefs(out.DecisionBasis)
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
		out.Action = ObjectiveReplannerActionReviewDisplaySafeRefs
		if out.FailureClass == FailureNone {
			out.FailureClass = FailureEvidenceWeak
		}
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		if out.NextHostAction == "" {
			out.NextHostAction = "provide_display_safe_refs"
		}
	}
	return out
}

func objectiveReplannerDecisionBlock(result ObjectiveReplannerDecision, status VerificationStatus, action ObjectiveReplannerAction, failure FailureClass, missing MissingInput, next NextHostAction, boundary Boundary) ObjectiveReplannerDecision {
	result.Status = NormalizeVerificationStatus(string(status))
	if result.Status == VerificationSatisfied || result.Status == VerificationNotEvaluated {
		result.Status = VerificationBlocked
	}
	result.Action = action
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.Boundaries = AppendBoundaries(result.Boundaries, boundary)
	result.NextHostAction = firstNextHostAction(next, result.NextHostAction)
	return result.Normalize()
}

func objectiveReplannerBudgetDecision(result ObjectiveReplannerDecision) ObjectiveReplannerDecision {
	result.FailureClass = FailureNone
	if len(result.EvidenceRefs) > 0 {
		return objectiveReplannerDecisionBlock(result, VerificationPartial, ObjectiveReplannerActionReturnPartial, FailureBudgetExhausted, "contract:budget", "return_partial_or_request_budget", "objective_replanner_budget_exhausted")
	}
	return objectiveReplannerDecisionBlock(result, VerificationBlocked, ObjectiveReplannerActionReturnBlocked, FailureBudgetExhausted, "contract:budget", "return_blocked", "objective_replanner_budget_exhausted")
}

func objectiveReplannerNoCandidateDecision(result ObjectiveReplannerDecision) ObjectiveReplannerDecision {
	if len(result.EvidenceRefs) > 0 {
		return objectiveReplannerDecisionBlock(result, VerificationPartial, ObjectiveReplannerActionReturnPartial, FailureConfigMissing, "host:strategy_candidate", "return_partial", "objective_replanner_no_candidate")
	}
	return objectiveReplannerDecisionBlock(result, VerificationBlocked, ObjectiveReplannerActionReturnBlocked, FailureConfigMissing, "host:strategy_candidate", "provide_strategy_scope", "objective_replanner_no_candidate")
}

func objectiveReplannerPolicyBlockedDecision(result ObjectiveReplannerDecision) ObjectiveReplannerDecision {
	if len(result.EvidenceRefs) > 0 {
		return objectiveReplannerDecisionBlock(result, VerificationPartial, ObjectiveReplannerActionReturnPartial, FailurePolicyBlocked, "host:policy_review", "return_partial_or_request_policy_review", "objective_replanner_policy_blocked")
	}
	return objectiveReplannerDecisionBlock(result, VerificationBlocked, ObjectiveReplannerActionReturnBlocked, FailurePolicyBlocked, "host:policy_review", "return_blocked", "objective_replanner_policy_blocked")
}

func objectiveReplannerActivation(values ...Activation) Activation {
	for _, value := range values {
		activation := NormalizeActivation(string(value))
		if activation != ActivationOff {
			return activation
		}
	}
	return ActivationOff
}

func objectiveReplannerVerification(values ...VerificationResult) VerificationResult {
	for _, value := range values {
		verification := value.Normalize()
		if verification.Status != VerificationNotEvaluated {
			return verification
		}
	}
	return VerificationResult{Status: VerificationNotEvaluated}.Normalize()
}

func objectiveReplannerCandidates(controller []StrategyCandidate, ranked []StrategyPlanCandidate) []StrategyCandidate {
	out := cloneStrategyCandidates(controller)
	for _, value := range normalizeStrategyPlanCandidates(ranked) {
		if value.Status != VerificationSatisfied && value.Status != VerificationReviewRequired {
			continue
		}
		out = append(out, value.Candidate)
	}
	return normalizeStrategyCandidates(out)
}

func objectiveReplannerCandidateRefs(candidates []StrategyCandidate) []DisplaySafeRef {
	out := []DisplaySafeRef{}
	for _, candidate := range normalizeStrategyCandidates(candidates) {
		out = appendUniqueDisplaySafeRef(out, objectiveReplannerCandidateRef(candidate))
	}
	return out
}

func objectiveReplannerCandidateRef(candidate StrategyCandidate) DisplaySafeRef {
	ref, _ := NormalizeDisplaySafeRef(candidate.Normalize().ID)
	return ref
}

func objectiveReplannerSelectStrategy(current DisplaySafeRef, candidates []StrategyCandidate, status VerificationStatus, failure FailureClass, maxAllowed ExecutionIntensity, noProgress bool) (StrategyCandidate, bool) {
	normalized := normalizeStrategyCandidates(candidates)
	if len(normalized) == 0 {
		return StrategyCandidate{}, false
	}
	same, hasSame := objectiveReplannerSameStrategy(current, normalized)
	switchCandidates := objectiveReplannerSwitchCandidates(current, normalized)
	if noProgress && len(switchCandidates) > 0 {
		return switchCandidates[0], true
	}
	if executionIntensityRank(maxAllowed) <= executionIntensityRank(IntensityL2BoundedToolLoop) && hasSame {
		return same, true
	}
	if objectiveReplannerFailurePrefersSwitch(failure) && len(switchCandidates) > 0 {
		return switchCandidates[0], true
	}
	switch NormalizeVerificationStatus(string(status)) {
	case VerificationPartial, VerificationReviewRequired:
		if hasSame {
			return same, true
		}
	}
	if current != "" {
		for _, candidate := range switchCandidates {
			return candidate, true
		}
	}
	if hasSame {
		return same, true
	}
	return normalized[0], true
}

func objectiveReplannerSwitchCandidates(current DisplaySafeRef, candidates []StrategyCandidate) []StrategyCandidate {
	out := []StrategyCandidate{}
	for _, candidate := range normalizeStrategyCandidates(candidates) {
		if current != "" && objectiveReplannerCandidateRef(candidate) == current {
			continue
		}
		out = append(out, candidate)
	}
	return normalizeStrategyCandidates(out)
}

func objectiveReplannerSameStrategy(current DisplaySafeRef, candidates []StrategyCandidate) (StrategyCandidate, bool) {
	if current == "" {
		return StrategyCandidate{}, false
	}
	for _, candidate := range normalizeStrategyCandidates(candidates) {
		if objectiveReplannerCandidateRef(candidate) == current {
			return candidate, true
		}
	}
	return StrategyCandidate{}, false
}

func objectiveReplannerCrossStrategyBlocked(maxAllowed ExecutionIntensity, current DisplaySafeRef, next DisplaySafeRef) bool {
	if executionIntensityRank(maxAllowed) > executionIntensityRank(IntensityL2BoundedToolLoop) {
		return false
	}
	return current != "" && next != "" && current != next
}

func objectiveReplannerApprovalRequired(verification VerificationResult, controller ObjectiveControllerDecision, approval ObjectiveApprovalState) bool {
	if approval.Required && !approval.Approved {
		return true
	}
	if controller.Action == ObjectiveActionRequestHostApproval || controller.Action == ObjectiveActionProvideApprovalRef {
		return true
	}
	switch firstFailureClass(verification.FailureClass, controller.FailureClass) {
	case FailureApprovalRequired, FailureAuthorizationMissing, FailurePermissionDenied:
		return true
	default:
		return false
	}
}

func objectiveReplannerCredentialRequired(verification VerificationResult, controller ObjectiveControllerDecision, approval ObjectiveApprovalState) bool {
	if firstFailureClass(verification.FailureClass, controller.FailureClass) == FailureCredentialMissing {
		return true
	}
	for _, missing := range MergeMissingInputs(verification.MissingInputs, controller.MissingInputs, approval.MissingInputs) {
		token := normalizeControlToken(string(missing))
		if stringsHasAnyPrefix(token, "host:credential", "credential:") {
			return true
		}
	}
	return false
}

func objectiveReplannerCredentialMissingInput(verification VerificationResult, controller ObjectiveControllerDecision, approval ObjectiveApprovalState) MissingInput {
	for _, missing := range MergeMissingInputs(verification.MissingInputs, controller.MissingInputs, approval.MissingInputs) {
		token := normalizeControlToken(string(missing))
		if stringsHasAnyPrefix(token, "host:credential", "credential:") {
			return missing
		}
	}
	return "host:credential"
}

func objectiveReplannerApprovalMissingInput(verification VerificationResult, controller ObjectiveControllerDecision, approval ObjectiveApprovalState) MissingInput {
	for _, missing := range MergeMissingInputs(verification.MissingInputs, controller.MissingInputs, approval.MissingInputs) {
		switch missing {
		case "host:objective_approval", "host:approval_ref", "host:credential", "host:authorization":
			return missing
		}
	}
	if approval.Required {
		return "host:objective_approval"
	}
	return "host:approval"
}

func objectiveReplannerCapabilityResolutionRequired(verification VerificationResult, controller ObjectiveControllerDecision, missingInputs []MissingInput) bool {
	switch firstFailureClass(verification.FailureClass, controller.FailureClass) {
	case FailureToolUnavailable, FailureSkillUnavailable, FailureConnectorUnavailable, FailureHostAdapterMissing, FailureCapabilityMissing, FailureConfigMissing:
		return true
	}
	for _, missing := range normalizeMissingInputs(missingInputs) {
		token := normalizeControlToken(string(missing))
		if stringsHasAnyPrefix(token, "capability:", "host:available_capability:", "host:available_capability", "host:strategy_catalog", "host:strategy_scope") {
			return true
		}
	}
	return false
}

func objectiveReplannerCapabilityMissingInput(missingInputs []MissingInput) MissingInput {
	for _, missing := range normalizeMissingInputs(missingInputs) {
		token := normalizeControlToken(string(missing))
		if stringsHasAnyPrefix(token, "capability:", "host:available_capability:", "host:available_capability") {
			return missing
		}
	}
	return "host:capability_resolution"
}

func objectiveReplannerPolicyBlocked(verification VerificationResult, controller ObjectiveControllerDecision) bool {
	switch firstFailureClass(verification.FailureClass, controller.FailureClass) {
	case FailurePolicyBlocked, FailureSandboxBlocked, FailureUnsupportedOperation:
		return true
	default:
		return false
	}
}

func objectiveReplannerSourceUnavailable(verification VerificationResult, controller ObjectiveControllerDecision) bool {
	switch firstFailureClass(verification.FailureClass, controller.FailureClass) {
	case FailureTargetUnavailable, FailureTargetNotFound, FailureExternalDependencyUnavailable, FailureTimeout:
		return true
	default:
		return false
	}
}

func objectiveReplannerFailurePrefersSwitch(failure FailureClass) bool {
	switch NormalizeFailureClass(string(failure)) {
	case FailureTargetUnavailable, FailureTargetNotFound, FailureExternalDependencyUnavailable, FailureTimeout, FailureVerificationFailed:
		return true
	default:
		return false
	}
}

func objectiveReplannerNoProgress(attempts []AttemptSummary, current DisplaySafeRef, verification VerificationResult) bool {
	return BuildObjectiveNoProgressSwitchGate(ObjectiveNoProgressSwitchGateInput{
		CurrentStrategyRef: current,
		Attempts:           attempts,
		Verification:       verification,
		Threshold:          2,
	}).RepeatedNoProgress
}

func objectiveReplannerDecisionInputUnsafe(input ObjectiveReplannerDecisionInput) bool {
	return displaySafeRefRejected(input.CurrentStrategyRef) ||
		evidenceRefRejected(input.EvidenceRefs) ||
		objectiveReplannerControllerUnsafe(input.Controller) ||
		objectiveReplannerVerificationGateUnsafe(input.Verification) ||
		objectiveReplannerStrategyPlanUnsafe(input.StrategyPlan) ||
		input.RawOutputLoaded
}

func objectiveReplannerControllerUnsafe(controller ObjectiveControllerDecision) bool {
	return displaySafeRefRejected(controller.CurrentStrategyRef) ||
		objectiveReplannerStrategyCandidatesUnsafe(controller.AllowedStrategies) ||
		objectiveReplannerVerificationUnsafe(controller.Verification) ||
		evidenceRefRejected(controller.EvidenceRefs) ||
		displaySafeRefSliceRejected(controller.DecisionBasis) ||
		controller.RawOutputLoaded
}

func objectiveReplannerVerificationGateUnsafe(gate ObjectiveVerificationGateResult) bool {
	return objectiveReplannerFrameUnsafe(gate.Frame) ||
		objectiveVerificationNormalizationUnsafe(gate.Normalization) ||
		objectiveReplannerVerificationUnsafe(gate.Verification) ||
		evidenceRefRejected(gate.EvidenceRefs) ||
		observationSliceUnsafePayload(gate.Observations) ||
		gate.RawOutputLoaded
}

func objectiveReplannerStrategyPlanUnsafe(plan StrategyPlannerResult) bool {
	return objectiveReplannerFrameUnsafe(plan.Frame) ||
		displaySafeRefRejected(plan.CatalogRef) ||
		displaySafeRefRejected(plan.CurrentStrategyRef) ||
		objectiveReplannerStrategyPlanCandidateUnsafe(plan.Selected) ||
		objectiveReplannerStrategyPlanCandidatesUnsafe(plan.RankedCandidates) ||
		objectiveReplannerStrategyPlanCandidatesUnsafe(plan.RejectedCandidates) ||
		displaySafeRefSliceRejected(plan.PolicyRefs) ||
		evidenceRefRejected(plan.EvidenceRefs) ||
		displaySafeRefSliceRejected(plan.DecisionBasis) ||
		plan.RawOutputLoaded
}

func objectiveReplannerFrameUnsafe(frame ObjectiveFrame) bool {
	return evidenceRefRejected(frame.RequiredEvidence) ||
		displaySafeRefSliceRejected(frame.CandidateCapabilities) ||
		displaySafeRefSliceRejected(frame.SourceContext)
}

func objectiveReplannerVerificationUnsafe(verification VerificationResult) bool {
	return evidenceRefRejected(verification.EvidenceRefs) ||
		verification.RawOutputLoaded
}

func objectiveReplannerStrategyCandidatesUnsafe(candidates []StrategyCandidate) bool {
	for _, candidate := range candidates {
		if objectiveReplannerStrategyCandidateUnsafe(candidate) {
			return true
		}
	}
	return false
}

func objectiveReplannerStrategyCandidateUnsafe(candidate StrategyCandidate) bool {
	return displaySafeRefRejected(DisplaySafeRef(candidate.ID)) ||
		displaySafeRefSliceRejected(candidate.CapabilityRefs) ||
		evidenceRefRejected(candidate.ExpectedEvidence) ||
		displaySafeRefRejected(candidate.FallbackOf)
}

func objectiveReplannerStrategyPlanCandidatesUnsafe(candidates []StrategyPlanCandidate) bool {
	for _, candidate := range candidates {
		if objectiveReplannerStrategyPlanCandidateUnsafe(candidate) {
			return true
		}
	}
	return false
}

func objectiveReplannerStrategyPlanCandidateUnsafe(candidate StrategyPlanCandidate) bool {
	return displaySafeRefRejected(candidate.SourceRef) ||
		objectiveReplannerStrategyCandidateUnsafe(candidate.Candidate) ||
		evidenceRefRejected(candidate.EvidenceRefs) ||
		displaySafeRefSliceRejected(candidate.DecisionBasis) ||
		candidate.RawOutputLoaded
}

func stringsHasAnyPrefix(value string, prefixes ...string) bool {
	for _, prefix := range prefixes {
		if len(prefix) <= len(value) && value[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}
