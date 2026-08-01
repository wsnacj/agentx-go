package controlcontract

type ObjectiveReplanProposalAction string

const (
	ObjectiveReplanProposalActionNone              ObjectiveReplanProposalAction = "none"
	ObjectiveReplanProposalActionRetrySameStrategy ObjectiveReplanProposalAction = "retry_same_strategy"
	ObjectiveReplanProposalActionSwitchCapability  ObjectiveReplanProposalAction = "switch_capability"
	ObjectiveReplanProposalActionAddEvidenceNode   ObjectiveReplanProposalAction = "add_evidence_node"
	ObjectiveReplanProposalActionAskUser           ObjectiveReplanProposalAction = "ask_user"
	ObjectiveReplanProposalActionRequestApproval   ObjectiveReplanProposalAction = "request_approval"
	ObjectiveReplanProposalActionCapabilityGap     ObjectiveReplanProposalAction = "capability_gap"
	ObjectiveReplanProposalActionPartialCloseout   ObjectiveReplanProposalAction = "partial_closeout"
	ObjectiveReplanProposalActionBlockedCloseout   ObjectiveReplanProposalAction = "blocked_closeout"
	ObjectiveReplanProposalActionSatisfiedCloseout ObjectiveReplanProposalAction = "satisfied_closeout"
	ObjectiveReplanProposalActionReviewRefs        ObjectiveReplanProposalAction = "review_display_safe_refs"
)

// ObjectiveReplanProposalActionMetadata describes generic ownership and gate
// requirements for a replan proposal action.
type ObjectiveReplanProposalActionMetadata struct {
	Action                      ObjectiveReplanProposalAction `json:"action,omitempty"`
	Owner                       string                        `json:"owner,omitempty"`
	RequiresHostDispatchBinding bool                          `json:"requires_host_dispatch_binding"`
	RequiresUserInput           bool                          `json:"requires_user_input"`
	RequiresApproval            bool                          `json:"requires_approval"`
	Terminal                    bool                          `json:"terminal"`
}

func NormalizeObjectiveReplanProposalAction(raw string) ObjectiveReplanProposalAction {
	switch normalizeEnumToken(raw) {
	case "", "none":
		return ObjectiveReplanProposalActionNone
	case "retry_same_strategy", "retry":
		return ObjectiveReplanProposalActionRetrySameStrategy
	case "switch_capability", "switch_strategy", "switch":
		return ObjectiveReplanProposalActionSwitchCapability
	case "add_evidence_node", "add_evidence":
		return ObjectiveReplanProposalActionAddEvidenceNode
	case "ask_user", "ask_user_clarification", "clarify":
		return ObjectiveReplanProposalActionAskUser
	case "request_approval", "request_host_approval", "request_user_approval":
		return ObjectiveReplanProposalActionRequestApproval
	case "capability_gap", "enter_capability_resolution", "capability_resolution":
		return ObjectiveReplanProposalActionCapabilityGap
	case "partial_closeout", "return_partial", "partial":
		return ObjectiveReplanProposalActionPartialCloseout
	case "blocked_closeout", "return_blocked", "blocked":
		return ObjectiveReplanProposalActionBlockedCloseout
	case "satisfied_closeout", "return_satisfied", "satisfied":
		return ObjectiveReplanProposalActionSatisfiedCloseout
	case "review_display_safe_refs", "display_safe_refs":
		return ObjectiveReplanProposalActionReviewRefs
	default:
		return ObjectiveReplanProposalActionNone
	}
}

// ObjectiveReplanProposalActionMetadataFor returns control-plane metadata for a
// normalized replan proposal action.
func ObjectiveReplanProposalActionMetadataFor(action ObjectiveReplanProposalAction) ObjectiveReplanProposalActionMetadata {
	normalized := NormalizeObjectiveReplanProposalAction(string(action))
	meta := ObjectiveReplanProposalActionMetadata{
		Action: normalized,
		Owner:  "host",
	}
	switch normalized {
	case ObjectiveReplanProposalActionRetrySameStrategy,
		ObjectiveReplanProposalActionSwitchCapability,
		ObjectiveReplanProposalActionAddEvidenceNode:
		meta.RequiresHostDispatchBinding = true
	case ObjectiveReplanProposalActionAskUser:
		meta.Owner = "user"
		meta.RequiresUserInput = true
	case ObjectiveReplanProposalActionRequestApproval:
		meta.RequiresApproval = true
	case ObjectiveReplanProposalActionCapabilityGap,
		ObjectiveReplanProposalActionReviewRefs:
	case ObjectiveReplanProposalActionPartialCloseout,
		ObjectiveReplanProposalActionBlockedCloseout,
		ObjectiveReplanProposalActionSatisfiedCloseout:
		meta.Terminal = true
	case ObjectiveReplanProposalActionNone:
	default:
		meta.RequiresHostDispatchBinding = true
	}
	return meta
}

// ObjectiveReplanProposalRequiresHostDispatchBinding reports whether a proposal
// must be reviewed and bound to a host-prepared child run input before dispatch.
func ObjectiveReplanProposalRequiresHostDispatchBinding(proposal *ObjectiveReplanProposal) bool {
	if proposal == nil {
		return false
	}
	normalized := proposal.Normalize()
	if len(normalized.Steps) == 0 {
		return false
	}
	return ObjectiveReplanProposalActionMetadataFor(normalized.Action).RequiresHostDispatchBinding
}

type ObjectiveReplanProposalInput struct {
	ProposalRef      DisplaySafeRef                  `json:"proposal_ref,omitempty"`
	Decision         ObjectiveReplannerDecision      `json:"decision,omitempty"`
	Verification     ObjectiveVerificationGateResult `json:"verification,omitempty"`
	AttemptLedgerRef DisplaySafeRef                  `json:"attempt_ledger_ref,omitempty"`
	PolicyRefs       []DisplaySafeRef                `json:"policy_refs,omitempty"`
	Boundaries       []Boundary                      `json:"boundaries,omitempty"`
	RawOutputLoaded  bool                            `json:"raw_output_loaded"`
}

type ObjectiveReplanProposalStep struct {
	ContractVersion  string                        `json:"contract_version,omitempty"`
	StepRef          DisplaySafeRef                `json:"step_ref,omitempty"`
	Action           ObjectiveReplanProposalAction `json:"action,omitempty"`
	Owner            string                        `json:"owner,omitempty"`
	CurrentStrategy  DisplaySafeRef                `json:"current_strategy_ref,omitempty"`
	NextStrategy     DisplaySafeRef                `json:"next_strategy_ref,omitempty"`
	CapabilityRefs   []DisplaySafeRef              `json:"capability_refs,omitempty"`
	RequiredEvidence []EvidenceRef                 `json:"required_evidence,omitempty"`
	MissingInputs    []MissingInput                `json:"missing_inputs,omitempty"`
	EvidenceRefs     []EvidenceRef                 `json:"evidence_refs,omitempty"`
	DecisionBasis    []DisplaySafeRef              `json:"decision_basis,omitempty"`
	Boundaries       []Boundary                    `json:"boundaries,omitempty"`
	NextHostAction   NextHostAction                `json:"next_host_action,omitempty"`
}

type ObjectiveReplanProposal struct {
	ContractVersion    string                        `json:"contract_version,omitempty"`
	Projected          bool                          `json:"projected"`
	ProposalRef        DisplaySafeRef                `json:"proposal_ref,omitempty"`
	Status             VerificationStatus            `json:"status,omitempty"`
	Action             ObjectiveReplanProposalAction `json:"action,omitempty"`
	ObjectiveID        string                        `json:"objective_id,omitempty"`
	CurrentStrategyRef DisplaySafeRef                `json:"current_strategy_ref,omitempty"`
	NextStrategyRef    DisplaySafeRef                `json:"next_strategy_ref,omitempty"`
	NextOwner          string                        `json:"next_owner,omitempty"`
	Steps              []ObjectiveReplanProposalStep `json:"steps,omitempty"`
	CapabilityGapRefs  []DisplaySafeRef              `json:"capability_gap_refs,omitempty"`
	EvidenceRefs       []EvidenceRef                 `json:"evidence_refs,omitempty"`
	MissingInputs      []MissingInput                `json:"missing_inputs,omitempty"`
	DecisionBasis      []DisplaySafeRef              `json:"decision_basis,omitempty"`
	PolicyRefs         []DisplaySafeRef              `json:"policy_refs,omitempty"`
	Boundaries         []Boundary                    `json:"boundaries,omitempty"`
	NextHostAction     NextHostAction                `json:"next_host_action,omitempty"`
	RunnerEffect       string                        `json:"runner_effect,omitempty"`
	PromptEffect       string                        `json:"prompt_effect,omitempty"`
	RawOutputLoaded    bool                          `json:"raw_output_loaded"`
}

func BuildObjectiveReplanProposal(input ObjectiveReplanProposalInput) ObjectiveReplanProposal {
	decision := input.Decision.Normalize()
	verification := input.Verification.Normalize()
	result := ObjectiveReplanProposal{
		ContractVersion:    ContractVersion,
		Projected:          true,
		ProposalRef:        firstDisplaySafeRef(input.ProposalRef, objectiveReplanProposalDefaultRef(decision)),
		Status:             decision.Status,
		ObjectiveID:        firstNonEmptyContractString(decision.ObjectiveID, verification.Frame.ID),
		CurrentStrategyRef: decision.CurrentStrategyRef,
		NextStrategyRef:    decision.NextStrategyRef,
		EvidenceRefs:       MergeEvidenceRefs(decision.EvidenceRefs, verification.EvidenceRefs),
		MissingInputs:      MergeMissingInputs(decision.MissingInputs, verification.MissingInputs),
		DecisionBasis: normalizeDisplaySafeRefs(append(
			append(cloneDisplaySafeRefs(decision.DecisionBasis), input.AttemptLedgerRef),
			"objective_replan_proposal",
		)),
		PolicyRefs: normalizeDisplaySafeRefs(input.PolicyRefs),
		Boundaries: MergeBoundaries(
			[]Boundary{
				"objective_replan_proposal",
				"decision_projection_only",
				"no_runner_dispatch",
				"no_strategy_dispatch",
				"no_runtime_adapter_execution",
				"no_install_or_schedule_apply",
				"host_must_apply_replan_proposal",
			},
			input.Boundaries,
			decision.Boundaries,
			verification.Boundaries,
		),
		NextHostAction:  firstNextHostAction(decision.NextHostAction, verification.NextHostAction),
		RunnerEffect:    "none",
		PromptEffect:    "none",
		RawOutputLoaded: input.RawOutputLoaded || decision.RawOutputLoaded || verification.RawOutputLoaded,
	}
	if objectiveReplanProposalInputUnsafe(input) {
		result.Action = ObjectiveReplanProposalActionReviewRefs
		result.Status = VerificationReviewRequired
		result.NextOwner = "host"
		result.MissingInputs = AppendMissingInputs(result.MissingInputs, "host:display_safe_refs")
		result.Boundaries = AppendBoundaries(result.Boundaries, "raw_output_not_allowed")
		result.NextHostAction = "provide_display_safe_refs"
		result.RawOutputLoaded = true
		return result.Normalize()
	}
	steps := objectiveReplanProposalSteps(decision, verification)
	result.Steps = steps
	result.Action = objectiveReplanProposalPrimaryAction(decision, steps)
	result.NextOwner = objectiveReplanProposalOwner(result.Action)
	result.CapabilityGapRefs = objectiveReplanProposalCapabilityGapRefs(result.MissingInputs)
	if objectiveReplanProposalActionOverridesDecisionNext(result.Action) {
		result.NextHostAction = objectiveReplanProposalNextHostAction(result.Action)
	}
	if result.NextHostAction == "" {
		result.NextHostAction = objectiveReplanProposalNextHostAction(result.Action)
	}
	return result.Normalize()
}

func CloneObjectiveReplanProposalStep(in ObjectiveReplanProposalStep) ObjectiveReplanProposalStep {
	out := in
	out.CapabilityRefs = cloneDisplaySafeRefs(in.CapabilityRefs)
	out.RequiredEvidence = cloneEvidenceRefs(in.RequiredEvidence)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.DecisionBasis = cloneDisplaySafeRefs(in.DecisionBasis)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (s ObjectiveReplanProposalStep) Clone() ObjectiveReplanProposalStep {
	return CloneObjectiveReplanProposalStep(s)
}

func (s ObjectiveReplanProposalStep) Normalize() ObjectiveReplanProposalStep {
	out := s.Clone()
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.StepRef = normalizeOneDisplaySafeRef(out.StepRef)
	out.Action = NormalizeObjectiveReplanProposalAction(string(out.Action))
	out.Owner = normalizeControlToken(out.Owner)
	out.CurrentStrategy = normalizeOneDisplaySafeRef(out.CurrentStrategy)
	out.NextStrategy = normalizeOneDisplaySafeRef(out.NextStrategy)
	out.CapabilityRefs = normalizeDisplaySafeRefs(out.CapabilityRefs)
	out.RequiredEvidence = normalizeEvidenceRefs(out.RequiredEvidence)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
	out.DecisionBasis = normalizeDisplaySafeRefs(out.DecisionBasis)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	return out
}

func CloneObjectiveReplanProposal(in ObjectiveReplanProposal) ObjectiveReplanProposal {
	out := in
	out.Steps = cloneObjectiveReplanProposalSteps(in.Steps)
	out.CapabilityGapRefs = cloneDisplaySafeRefs(in.CapabilityGapRefs)
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.DecisionBasis = cloneDisplaySafeRefs(in.DecisionBasis)
	out.PolicyRefs = cloneDisplaySafeRefs(in.PolicyRefs)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (p ObjectiveReplanProposal) Clone() ObjectiveReplanProposal {
	return CloneObjectiveReplanProposal(p)
}

func (p ObjectiveReplanProposal) Normalize() ObjectiveReplanProposal {
	out := p.Clone()
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.ProposalRef = normalizeOneDisplaySafeRef(out.ProposalRef)
	out.Status = NormalizeVerificationStatus(string(out.Status))
	if out.Status == VerificationNotEvaluated {
		out.Status = VerificationBlocked
	}
	out.Action = NormalizeObjectiveReplanProposalAction(string(out.Action))
	out.ObjectiveID = firstNonEmptyContractString(out.ObjectiveID)
	out.CurrentStrategyRef = normalizeOneDisplaySafeRef(out.CurrentStrategyRef)
	out.NextStrategyRef = normalizeOneDisplaySafeRef(out.NextStrategyRef)
	out.NextOwner = normalizeControlToken(out.NextOwner)
	out.Steps = normalizeObjectiveReplanProposalSteps(out.Steps)
	out.CapabilityGapRefs = normalizeDisplaySafeRefs(out.CapabilityGapRefs)
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.DecisionBasis = normalizeDisplaySafeRefs(out.DecisionBasis)
	out.PolicyRefs = normalizeDisplaySafeRefs(out.PolicyRefs)
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
	if out.RawOutputLoaded {
		out.Status = VerificationReviewRequired
		out.Action = ObjectiveReplanProposalActionReviewRefs
		out.NextOwner = "host"
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		if out.NextHostAction == "" {
			out.NextHostAction = "provide_display_safe_refs"
		}
	}
	return out
}

func objectiveReplanProposalSteps(decision ObjectiveReplannerDecision, verification ObjectiveVerificationGateResult) []ObjectiveReplanProposalStep {
	decision = decision.Normalize()
	verification = verification.Normalize()
	steps := []ObjectiveReplanProposalStep{}
	if objectiveReplanProposalNeedsUserClarification(decision, verification) {
		steps = append(steps, objectiveReplanProposalStep(decision, verification, ObjectiveReplanProposalActionAskUser))
	}
	if objectiveReplanProposalNeedsEvidenceNode(decision, verification) {
		steps = append(steps, objectiveReplanProposalStep(decision, verification, ObjectiveReplanProposalActionAddEvidenceNode))
	}
	switch decision.Action {
	case ObjectiveReplannerActionRetrySameStrategy:
		steps = append(steps, objectiveReplanProposalStep(decision, verification, ObjectiveReplanProposalActionRetrySameStrategy))
	case ObjectiveReplannerActionSwitchStrategy:
		steps = append(steps, objectiveReplanProposalStep(decision, verification, ObjectiveReplanProposalActionSwitchCapability))
	case ObjectiveReplannerActionEnterCapabilityResolution:
		steps = append(steps, objectiveReplanProposalStep(decision, verification, ObjectiveReplanProposalActionCapabilityGap))
	case ObjectiveReplannerActionRequestApproval:
		steps = append(steps, objectiveReplanProposalStep(decision, verification, ObjectiveReplanProposalActionRequestApproval))
	case ObjectiveReplannerActionReturnPartial:
		steps = append(steps, objectiveReplanProposalStep(decision, verification, ObjectiveReplanProposalActionPartialCloseout))
	case ObjectiveReplannerActionReturnBlocked:
		steps = append(steps, objectiveReplanProposalStep(decision, verification, ObjectiveReplanProposalActionBlockedCloseout))
	case ObjectiveReplannerActionReturnSatisfied:
		steps = append(steps, objectiveReplanProposalStep(decision, verification, ObjectiveReplanProposalActionSatisfiedCloseout))
	case ObjectiveReplannerActionReviewDisplaySafeRefs:
		steps = append(steps, objectiveReplanProposalStep(decision, verification, ObjectiveReplanProposalActionReviewRefs))
	}
	if len(steps) == 0 {
		steps = append(steps, objectiveReplanProposalStep(decision, verification, ObjectiveReplanProposalActionNone))
	}
	return normalizeObjectiveReplanProposalSteps(steps)
}

func objectiveReplanProposalStep(decision ObjectiveReplannerDecision, verification ObjectiveVerificationGateResult, action ObjectiveReplanProposalAction) ObjectiveReplanProposalStep {
	return ObjectiveReplanProposalStep{
		ContractVersion:  ContractVersion,
		StepRef:          objectiveReplanProposalStepRef(action, decision),
		Action:           action,
		Owner:            objectiveReplanProposalOwner(action),
		CurrentStrategy:  decision.CurrentStrategyRef,
		NextStrategy:     decision.NextStrategyRef,
		CapabilityRefs:   objectiveReplanProposalStepCapabilityRefs(action, decision),
		RequiredEvidence: objectiveReplanProposalStepRequiredEvidence(action, decision, verification),
		MissingInputs:    MergeMissingInputs(decision.MissingInputs, verification.MissingInputs),
		EvidenceRefs:     MergeEvidenceRefs(decision.EvidenceRefs, verification.EvidenceRefs),
		DecisionBasis:    appendUniqueDisplaySafeRef(decision.DecisionBasis, DisplaySafeRef("replan_action:"+string(action))),
		Boundaries: MergeBoundaries(
			[]Boundary{"objective_replan_proposal_step", Boundary("replan_action_" + string(action))},
			decision.Boundaries,
		),
		NextHostAction: objectiveReplanProposalStepNextHostAction(action, decision, verification),
	}.Normalize()
}

func objectiveReplanProposalPrimaryAction(decision ObjectiveReplannerDecision, steps []ObjectiveReplanProposalStep) ObjectiveReplanProposalAction {
	if decision.Action == ObjectiveReplannerActionRetrySameStrategy && len(steps) > 0 && steps[0].Action == ObjectiveReplanProposalActionAddEvidenceNode {
		return ObjectiveReplanProposalActionAddEvidenceNode
	}
	if len(steps) == 0 {
		return ObjectiveReplanProposalActionNone
	}
	return steps[0].Normalize().Action
}

func objectiveReplanProposalNeedsEvidenceNode(decision ObjectiveReplannerDecision, verification ObjectiveVerificationGateResult) bool {
	if decision.Action != ObjectiveReplannerActionRetrySameStrategy && decision.Action != ObjectiveReplannerActionReturnPartial {
		return false
	}
	for _, missing := range MergeMissingInputs(decision.MissingInputs, verification.MissingInputs) {
		token := normalizeControlToken(string(missing))
		if stringsHasAnyPrefix(token, "evidence:", "host:objective_node_readback", "host:required_evidence", "control_plane:normalized_observations") {
			return true
		}
	}
	switch firstFailureClass(decision.FailureClass, verification.FailureClass) {
	case FailureEvidenceMissing, FailureEvidenceWeak:
		return true
	default:
		return false
	}
}

func objectiveReplanProposalNeedsUserClarification(decision ObjectiveReplannerDecision, verification ObjectiveVerificationGateResult) bool {
	switch firstFailureClass(decision.FailureClass, verification.FailureClass) {
	case FailureAmbiguousGoal, FailureInsufficientInformation:
		return true
	}
	for _, missing := range MergeMissingInputs(decision.MissingInputs, verification.MissingInputs) {
		token := normalizeControlToken(string(missing))
		if stringsHasAnyPrefix(token, "user:", "host:user_", "host:success_criteria", "host:objective_contract") {
			return true
		}
	}
	return false
}

func objectiveReplanProposalStepRequiredEvidence(action ObjectiveReplanProposalAction, decision ObjectiveReplannerDecision, verification ObjectiveVerificationGateResult) []EvidenceRef {
	switch action {
	case ObjectiveReplanProposalActionAddEvidenceNode:
		return MergeEvidenceRefs(verification.Frame.RequiredEvidence, decision.Verification.EvidenceRefs, verification.EvidenceRefs)
	default:
		return MergeEvidenceRefs(decision.SelectedStrategy.ExpectedEvidence, verification.Frame.RequiredEvidence)
	}
}

func objectiveReplanProposalStepCapabilityRefs(action ObjectiveReplanProposalAction, decision ObjectiveReplannerDecision) []DisplaySafeRef {
	switch action {
	case ObjectiveReplanProposalActionSwitchCapability, ObjectiveReplanProposalActionRetrySameStrategy:
		return cloneDisplaySafeRefs(decision.SelectedStrategy.CapabilityRefs)
	case ObjectiveReplanProposalActionCapabilityGap:
		return objectiveReplanProposalCapabilityGapRefs(decision.MissingInputs)
	default:
		return nil
	}
}

func objectiveReplanProposalCapabilityGapRefs(inputs []MissingInput) []DisplaySafeRef {
	refs := []DisplaySafeRef{}
	for _, missing := range normalizeMissingInputs(inputs) {
		token := normalizeControlToken(string(missing))
		if !stringsHasAnyPrefix(token, "capability:", "host:available_capability:", "host:available_capability") {
			continue
		}
		if ref, ok := NormalizeDisplaySafeRef(string(missing)); ok {
			refs = appendUniqueDisplaySafeRef(refs, ref)
		}
	}
	return normalizeDisplaySafeRefs(refs)
}

func objectiveReplanProposalOwner(action ObjectiveReplanProposalAction) string {
	return ObjectiveReplanProposalActionMetadataFor(action).Owner
}

func objectiveReplanProposalStepNextHostAction(action ObjectiveReplanProposalAction, decision ObjectiveReplannerDecision, verification ObjectiveVerificationGateResult) NextHostAction {
	if objectiveReplanProposalActionOverridesDecisionNext(action) {
		return objectiveReplanProposalNextHostAction(action)
	}
	return objectiveReplanProposalFirstNextHostAction(decision.NextHostAction, verification.NextHostAction, objectiveReplanProposalNextHostAction(action))
}

func objectiveReplanProposalActionOverridesDecisionNext(action ObjectiveReplanProposalAction) bool {
	switch NormalizeObjectiveReplanProposalAction(string(action)) {
	case ObjectiveReplanProposalActionAddEvidenceNode,
		ObjectiveReplanProposalActionAskUser,
		ObjectiveReplanProposalActionReviewRefs:
		return true
	default:
		return false
	}
}

func objectiveReplanProposalNextHostAction(action ObjectiveReplanProposalAction) NextHostAction {
	switch NormalizeObjectiveReplanProposalAction(string(action)) {
	case ObjectiveReplanProposalActionRetrySameStrategy:
		return "host_may_retry_same_strategy"
	case ObjectiveReplanProposalActionSwitchCapability:
		return "host_may_switch_strategy"
	case ObjectiveReplanProposalActionAddEvidenceNode:
		return "host_may_add_evidence_node"
	case ObjectiveReplanProposalActionAskUser:
		return "ask_user_clarification"
	case ObjectiveReplanProposalActionRequestApproval:
		return "request_host_approval"
	case ObjectiveReplanProposalActionCapabilityGap:
		return "enter_capability_resolution"
	case ObjectiveReplanProposalActionPartialCloseout:
		return "return_partial"
	case ObjectiveReplanProposalActionBlockedCloseout:
		return "return_blocked"
	case ObjectiveReplanProposalActionSatisfiedCloseout:
		return "return_satisfied"
	case ObjectiveReplanProposalActionReviewRefs:
		return "provide_display_safe_refs"
	default:
		return "review_replan_proposal"
	}
}

func objectiveReplanProposalDefaultRef(decision ObjectiveReplannerDecision) DisplaySafeRef {
	action := NormalizeObjectiveReplanProposalAction(string(decision.Action))
	if action == ObjectiveReplanProposalActionNone {
		action = ObjectiveReplanProposalActionBlockedCloseout
	}
	return DisplaySafeRef("proposal:objective_replan_" + normalizeControlToken(string(action)))
}

func objectiveReplanProposalStepRef(action ObjectiveReplanProposalAction, decision ObjectiveReplannerDecision) DisplaySafeRef {
	token := normalizeControlToken(string(action))
	if token == "" {
		token = "none"
	}
	if decision.ObjectiveID != "" {
		token = objectiveGraphSafeID(decision.ObjectiveID) + "_" + token
	}
	return DisplaySafeRef("replan_step:" + token)
}

func objectiveReplanProposalInputUnsafe(input ObjectiveReplanProposalInput) bool {
	decision := input.Decision
	return input.RawOutputLoaded ||
		displaySafeRefRejected(input.ProposalRef) ||
		displaySafeRefRejected(input.AttemptLedgerRef) ||
		displaySafeRefSliceRejected(input.PolicyRefs) ||
		decision.RawOutputLoaded ||
		displaySafeRefRejected(decision.CurrentStrategyRef) ||
		displaySafeRefRejected(decision.NextStrategyRef) ||
		objectiveReplannerStrategyCandidateUnsafe(decision.SelectedStrategy) ||
		displaySafeRefSliceRejected(decision.CandidateRefs) ||
		objectiveReplannerVerificationUnsafe(decision.Verification) ||
		evidenceRefRejected(decision.EvidenceRefs) ||
		displaySafeRefSliceRejected(decision.DecisionBasis) ||
		objectiveReplannerVerificationGateUnsafe(input.Verification)
}

func objectiveReplanProposalFirstNextHostAction(values ...NextHostAction) NextHostAction {
	for _, value := range values {
		if action := NormalizeNextHostAction(string(value)); action != "" {
			return action
		}
	}
	return ""
}

func cloneObjectiveReplanProposalSteps(in []ObjectiveReplanProposalStep) []ObjectiveReplanProposalStep {
	if len(in) == 0 {
		return nil
	}
	out := make([]ObjectiveReplanProposalStep, len(in))
	for i, value := range in {
		out[i] = value.Clone()
	}
	return out
}

func normalizeObjectiveReplanProposalSteps(in []ObjectiveReplanProposalStep) []ObjectiveReplanProposalStep {
	out := make([]ObjectiveReplanProposalStep, 0, len(in))
	seen := map[DisplaySafeRef]bool{}
	for _, value := range in {
		normalized := value.Normalize()
		if normalized.StepRef == "" && normalized.Action == ObjectiveReplanProposalActionNone {
			continue
		}
		if normalized.StepRef != "" {
			if seen[normalized.StepRef] {
				continue
			}
			seen[normalized.StepRef] = true
		}
		out = append(out, normalized)
	}
	return out
}
