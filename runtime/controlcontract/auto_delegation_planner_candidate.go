package controlcontract

import (
	"encoding/json"
	"io"
	"strings"
)

type AutoDelegationPlannerInstructionMode string

const (
	AutoDelegationPlannerInstructionNone    AutoDelegationPlannerInstructionMode = "none"
	AutoDelegationPlannerInstructionObserve AutoDelegationPlannerInstructionMode = "observe"
	AutoDelegationPlannerInstructionPropose AutoDelegationPlannerInstructionMode = "propose"
	AutoDelegationPlannerInstructionManaged AutoDelegationPlannerInstructionMode = "managed"
)

func NormalizeAutoDelegationPlannerInstructionMode(raw string) AutoDelegationPlannerInstructionMode {
	switch normalizeEnumToken(raw) {
	case "", "none", "off", "disabled":
		return AutoDelegationPlannerInstructionNone
	case "observe", "observe_only":
		return AutoDelegationPlannerInstructionObserve
	case "propose", "proposal", "proposal_only":
		return AutoDelegationPlannerInstructionPropose
	case "managed", "bounded", "bounded_execution":
		return AutoDelegationPlannerInstructionManaged
	default:
		return AutoDelegationPlannerInstructionNone
	}
}

type AutoDelegationPlannerInstructionInput struct {
	PolicyReview            AutoDelegationPolicyReview `json:"policy_review,omitempty"`
	AvailableToolRefs       []DisplaySafeRef           `json:"available_tool_refs,omitempty"`
	AvailableCapabilityRefs []DisplaySafeRef           `json:"available_capability_refs,omitempty"`
	Boundaries              []Boundary                 `json:"boundaries,omitempty"`
	MissingInputs           []MissingInput             `json:"missing_inputs,omitempty"`
	RawOutputLoaded         bool                       `json:"raw_output_loaded"`
}

type AutoDelegationPlannerInstructionProfile struct {
	ContractVersion                        string                               `json:"contract_version,omitempty"`
	Projected                              bool                                 `json:"projected"`
	Status                                 VerificationStatus                   `json:"status,omitempty"`
	Mode                                   AutoDelegationPlannerInstructionMode `json:"mode,omitempty"`
	RequestPlanCandidate                   bool                                 `json:"request_plan_candidate"`
	ProactiveDelegationInstructionsExposed bool                                 `json:"proactive_delegation_instructions_exposed"`
	BoundedExecutionInstructionsExposed    bool                                 `json:"bounded_execution_instructions_exposed"`
	PlanOnly                               bool                                 `json:"plan_only"`
	WorkerAsToolDefault                    bool                                 `json:"worker_as_tool_default"`
	AvailableToolRefs                      []DisplaySafeRef                     `json:"available_tool_refs,omitempty"`
	AvailableCapabilityRefs                []DisplaySafeRef                     `json:"available_capability_refs,omitempty"`
	RequiredOutputFields                   []string                             `json:"required_output_fields,omitempty"`
	ForbiddenInstructionRefs               []DisplaySafeRef                     `json:"forbidden_instruction_refs,omitempty"`
	MissingInputs                          []MissingInput                       `json:"missing_inputs,omitempty"`
	BlockedReasons                         []string                             `json:"blocked_reasons,omitempty"`
	FailureClass                           FailureClass                         `json:"failure_class,omitempty"`
	Boundaries                             []Boundary                           `json:"boundaries,omitempty"`
	NextHostAction                         NextHostAction                       `json:"next_host_action,omitempty"`
	RunnerEffect                           string                               `json:"runner_effect,omitempty"`
	PromptEffect                           string                               `json:"prompt_effect,omitempty"`
	RawOutputLoaded                        bool                                 `json:"raw_output_loaded"`
}

func BuildAutoDelegationPlannerInstructionProfile(input AutoDelegationPlannerInstructionInput) AutoDelegationPlannerInstructionProfile {
	unsafe := autoDelegationPlannerInstructionInputUnsafe(input)
	policyReview := input.PolicyReview.Normalize()
	result := baseAutoDelegationPlannerInstructionProfile(input, policyReview)
	if unsafe || input.RawOutputLoaded || policyReview.RawOutputLoaded {
		return autoDelegationPlannerInstructionProfileBlock(result, VerificationReviewRequired, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs", "raw_output_not_allowed")
	}
	if !policyReview.Ready || (!policyReview.AutoDelegationAllowed && policyReview.Policy.Mode != AutoDelegationObserve) {
		result.Status = firstNonApplicableStatus(policyReview.Status, VerificationNotApplicable)
		result.Mode = AutoDelegationPlannerInstructionNone
		result.FailureClass = firstFailureClass(policyReview.FailureClass, FailureNone)
		result.NextHostAction = "do_not_request_auto_delegation_plan"
		result.PromptEffect = "auto_delegation_instruction_absent"
		result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, "auto_delegation_policy_not_ready")
		result.Boundaries = AppendBoundaries(result.Boundaries, "auto_delegation_instruction_absent", "tools_may_be_available_without_delegation_instruction")
		return result.Normalize()
	}

	switch policyReview.Policy.Mode {
	case AutoDelegationObserve:
		result.Status = VerificationSatisfied
		result.Mode = AutoDelegationPlannerInstructionObserve
		result.FailureClass = FailureNone
		result.NextHostAction = "record_auto_delegation_observation"
		result.PromptEffect = "auto_delegation_instruction_absent"
		result.Boundaries = AppendBoundaries(result.Boundaries, "auto_delegation_observe_without_spawn_instruction", "tools_may_be_available_without_delegation_instruction")
	case AutoDelegationPropose:
		result.Status = VerificationSatisfied
		result.Mode = AutoDelegationPlannerInstructionPropose
		result.RequestPlanCandidate = true
		result.ProactiveDelegationInstructionsExposed = true
		result.PlanOnly = true
		result.FailureClass = FailureNone
		result.NextHostAction = "request_auto_delegation_proposal_json"
		result.PromptEffect = "auto_delegation_proposal_instruction"
		result.RequiredOutputFields = autoDelegationPlannerRequiredOutputFields()
		result.ForbiddenInstructionRefs = []DisplaySafeRef{"instruction:spawn_child_task", "instruction:dispatch_subagent"}
		result.Boundaries = AppendBoundaries(result.Boundaries, "auto_delegation_proposal_instruction_only", "no_child_task_spawn", "no_subagent_dispatch")
	case AutoDelegationManagedReadOnly, AutoDelegationManaged:
		result.Status = VerificationSatisfied
		result.Mode = AutoDelegationPlannerInstructionManaged
		result.RequestPlanCandidate = true
		result.ProactiveDelegationInstructionsExposed = true
		result.BoundedExecutionInstructionsExposed = true
		result.PlanOnly = false
		result.FailureClass = FailureNone
		result.NextHostAction = "request_bounded_auto_delegation_plan_json"
		result.PromptEffect = "auto_delegation_bounded_plan_instruction"
		result.RequiredOutputFields = autoDelegationPlannerRequiredOutputFields()
		result.ForbiddenInstructionRefs = []DisplaySafeRef{"instruction:direct_child_spawn", "instruction:unbounded_subagent_dispatch"}
		result.Boundaries = AppendBoundaries(result.Boundaries, "auto_delegation_bounded_instruction", "host_review_required_before_dispatch")
	default:
		result.Status = VerificationNotApplicable
		result.Mode = AutoDelegationPlannerInstructionNone
		result.FailureClass = FailureNone
		result.NextHostAction = "do_not_request_auto_delegation_plan"
		result.PromptEffect = "auto_delegation_instruction_absent"
		result.Boundaries = AppendBoundaries(result.Boundaries, "auto_delegation_instruction_absent", "tools_may_be_available_without_delegation_instruction")
	}
	return result.Normalize()
}

func (p AutoDelegationPlannerInstructionProfile) Clone() AutoDelegationPlannerInstructionProfile {
	out := p
	out.AvailableToolRefs = cloneDisplaySafeRefs(p.AvailableToolRefs)
	out.AvailableCapabilityRefs = cloneDisplaySafeRefs(p.AvailableCapabilityRefs)
	out.RequiredOutputFields = cloneStringSlice(p.RequiredOutputFields)
	out.ForbiddenInstructionRefs = cloneDisplaySafeRefs(p.ForbiddenInstructionRefs)
	out.MissingInputs = cloneMissingInputs(p.MissingInputs)
	out.BlockedReasons = cloneStringSlice(p.BlockedReasons)
	out.Boundaries = cloneBoundaries(p.Boundaries)
	return out
}

func (p AutoDelegationPlannerInstructionProfile) Normalize() AutoDelegationPlannerInstructionProfile {
	out := p.Clone()
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeVerificationStatus(string(out.Status))
	if out.Status == VerificationNotEvaluated {
		out.Status = VerificationNotApplicable
	}
	out.Mode = NormalizeAutoDelegationPlannerInstructionMode(string(out.Mode))
	out.WorkerAsToolDefault = true
	out.AvailableToolRefs = normalizeDisplaySafeRefs(out.AvailableToolRefs)
	out.AvailableCapabilityRefs = normalizeDisplaySafeRefs(out.AvailableCapabilityRefs)
	out.RequiredOutputFields = normalizeControlTokenList(out.RequiredOutputFields)
	out.ForbiddenInstructionRefs = normalizeDisplaySafeRefs(out.ForbiddenInstructionRefs)
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
	if out.RawOutputLoaded {
		out.Status = VerificationReviewRequired
		out.RequestPlanCandidate = false
		out.ProactiveDelegationInstructionsExposed = false
		out.BoundedExecutionInstructionsExposed = false
		out.FailureClass = firstFailureClass(out.FailureClass, FailureEvidenceWeak)
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		out.NextHostAction = "provide_display_safe_refs"
	}
	if out.Status != VerificationSatisfied {
		out.RequestPlanCandidate = false
		out.ProactiveDelegationInstructionsExposed = false
		out.BoundedExecutionInstructionsExposed = false
	}
	if !out.ProactiveDelegationInstructionsExposed {
		out.BoundedExecutionInstructionsExposed = false
	}
	if out.PlanOnly {
		out.BoundedExecutionInstructionsExposed = false
	}
	return out
}

type AutoDelegationPlannerCandidate struct {
	ContractVersion string             `json:"contract_version,omitempty"`
	CandidateRef    DisplaySafeRef     `json:"candidate_ref,omitempty"`
	SourceRef       DisplaySafeRef     `json:"source_ref,omitempty"`
	PolicyRef       DisplaySafeRef     `json:"policy_ref,omitempty"`
	Plan            AutoDelegationPlan `json:"plan,omitempty"`
	Boundaries      []Boundary         `json:"boundaries,omitempty"`
	MissingInputs   []MissingInput     `json:"missing_inputs,omitempty"`
	RawOutputLoaded bool               `json:"raw_output_loaded"`
}

func CloneAutoDelegationPlannerCandidate(in AutoDelegationPlannerCandidate) AutoDelegationPlannerCandidate {
	out := in
	out.Plan = in.Plan.Clone()
	out.Boundaries = cloneBoundaries(in.Boundaries)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	return out
}

func (c AutoDelegationPlannerCandidate) Clone() AutoDelegationPlannerCandidate {
	return CloneAutoDelegationPlannerCandidate(c)
}

func (c AutoDelegationPlannerCandidate) Normalize() AutoDelegationPlannerCandidate {
	out := c.Clone()
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.CandidateRef = normalizeOneDisplaySafeRef(out.CandidateRef)
	out.SourceRef = normalizeOneDisplaySafeRef(out.SourceRef)
	out.PolicyRef = normalizeOneDisplaySafeRef(out.PolicyRef)
	out.Plan = out.Plan.Normalize()
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	return out
}

type AutoDelegationPlannerCandidateJSONDecodeInput struct {
	RawJSON            string                     `json:"-"`
	PolicyReview       AutoDelegationPolicyReview `json:"policy_review,omitempty"`
	SourceRef          DisplaySafeRef             `json:"source_ref,omitempty"`
	ParentObjectiveRef DisplaySafeRef             `json:"parent_objective_ref,omitempty"`
	Boundaries         []Boundary                 `json:"boundaries,omitempty"`
	MissingInputs      []MissingInput             `json:"missing_inputs,omitempty"`
	RawOutputLoaded    bool                       `json:"raw_output_loaded"`
}

type AutoDelegationPlannerCandidateJSONDecodeReport struct {
	ContractVersion string                         `json:"contract_version,omitempty"`
	Projected       bool                           `json:"projected"`
	Status          VerificationStatus             `json:"status,omitempty"`
	Decoded         bool                           `json:"decoded"`
	Candidate       AutoDelegationPlannerCandidate `json:"candidate,omitempty"`
	PlanReview      AutoDelegationPlanReview       `json:"plan_review,omitempty"`
	MissingInputs   []MissingInput                 `json:"missing_inputs,omitempty"`
	BlockedReasons  []string                       `json:"blocked_reasons,omitempty"`
	FailureClass    FailureClass                   `json:"failure_class,omitempty"`
	Boundaries      []Boundary                     `json:"boundaries,omitempty"`
	NextHostAction  NextHostAction                 `json:"next_host_action,omitempty"`
	RunnerEffect    string                         `json:"runner_effect,omitempty"`
	PromptEffect    string                         `json:"prompt_effect,omitempty"`
	RawOutputLoaded bool                           `json:"raw_output_loaded"`
}

func BuildAutoDelegationPlannerCandidateFromJSON(input AutoDelegationPlannerCandidateJSONDecodeInput) AutoDelegationPlannerCandidateJSONDecodeReport {
	result := baseAutoDelegationPlannerCandidateJSONDecodeReport(input)
	if autoDelegationPlannerCandidateJSONDecodeInputUnsafe(input) || input.RawOutputLoaded {
		return autoDelegationPlannerCandidateJSONDecodeBlock(result, VerificationReviewRequired, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs", "raw_output_not_allowed")
	}
	if strings.TrimSpace(input.RawJSON) == "" {
		return autoDelegationPlannerCandidateJSONDecodeBlock(result, VerificationBlocked, FailureInvalidInput, "auto_delegation_planner_json_missing", "host:auto_delegation_planner_json", "provide_auto_delegation_planner_json", "auto_delegation_planner_json_missing")
	}

	candidate, ok, boundary := decodeAutoDelegationPlannerCandidateJSON(input.RawJSON)
	if !ok {
		result = autoDelegationPlannerCandidateJSONDecodeBlock(result, VerificationBlocked, FailureInvalidInput, "auto_delegation_planner_json_invalid", "host:auto_delegation_planner_json", "provide_auto_delegation_planner_json", boundary)
		result.Boundaries = AppendBoundaries(result.Boundaries, "deterministic_blocked_fallback", "no_prompt_heuristic_fallback")
		return result.Normalize()
	}
	result.Decoded = true
	candidate = autoDelegationPlannerCandidateApplyDefaults(candidate, input)
	planReview := BuildAutoDelegationPlanReview(input.PolicyReview, candidate.Plan)
	result.Candidate = candidate.Normalize()
	result.PlanReview = planReview
	result.Status = planReview.Status
	result.FailureClass = planReview.FailureClass
	result.MissingInputs = MergeMissingInputs(result.MissingInputs, candidate.MissingInputs, planReview.MissingInputs)
	result.BlockedReasons = appendUniqueControlTokens(result.BlockedReasons, candidateBoundedReasons(candidate), planReview.BlockedReasons)
	result.Boundaries = MergeBoundaries(result.Boundaries, candidate.Boundaries, planReview.Boundaries)
	result.NextHostAction = planReview.NextHostAction
	result.RawOutputLoaded = result.RawOutputLoaded || candidate.RawOutputLoaded || planReview.RawOutputLoaded
	if result.Status == VerificationSatisfied && result.FailureClass == FailureNone {
		result.Boundaries = AppendBoundaries(result.Boundaries, "auto_delegation_planner_candidate_validated")
	}
	return result.Normalize()
}

func (r AutoDelegationPlannerCandidateJSONDecodeReport) Clone() AutoDelegationPlannerCandidateJSONDecodeReport {
	out := r
	out.Candidate = r.Candidate.Clone()
	out.PlanReview = r.PlanReview.Clone()
	out.MissingInputs = cloneMissingInputs(r.MissingInputs)
	out.BlockedReasons = cloneStringSlice(r.BlockedReasons)
	out.Boundaries = cloneBoundaries(r.Boundaries)
	return out
}

func (r AutoDelegationPlannerCandidateJSONDecodeReport) Normalize() AutoDelegationPlannerCandidateJSONDecodeReport {
	out := r.Clone()
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeVerificationStatus(string(out.Status))
	if out.Status == VerificationNotEvaluated {
		out.Status = VerificationBlocked
	}
	out.Candidate = out.Candidate.Normalize()
	out.PlanReview = out.PlanReview.Normalize()
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
	if out.RawOutputLoaded || out.Candidate.RawOutputLoaded || out.PlanReview.RawOutputLoaded {
		out.Status = VerificationReviewRequired
		out.FailureClass = firstFailureClass(out.FailureClass, FailureEvidenceWeak)
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		out.NextHostAction = "provide_display_safe_refs"
	}
	return out
}

func baseAutoDelegationPlannerInstructionProfile(input AutoDelegationPlannerInstructionInput, policyReview AutoDelegationPolicyReview) AutoDelegationPlannerInstructionProfile {
	return AutoDelegationPlannerInstructionProfile{
		ContractVersion:         ContractVersion,
		Projected:               true,
		Status:                  VerificationNotApplicable,
		Mode:                    AutoDelegationPlannerInstructionNone,
		WorkerAsToolDefault:     true,
		AvailableToolRefs:       cloneDisplaySafeRefs(input.AvailableToolRefs),
		AvailableCapabilityRefs: cloneDisplaySafeRefs(input.AvailableCapabilityRefs),
		MissingInputs:           MergeMissingInputs(input.MissingInputs, policyReview.MissingInputs),
		FailureClass:            FailureNone,
		Boundaries: MergeBoundaries(
			[]Boundary{
				"auto_delegation_planner_instruction_profile",
				"instruction_projection_only",
				"no_child_task_spawn",
				"no_subagent_dispatch",
				"no_scheduler_enqueue",
				"worker_as_tool_default",
				"no_backend_execution",
			},
			input.Boundaries,
			policyReview.Boundaries,
		),
		NextHostAction:  "do_not_request_auto_delegation_plan",
		RunnerEffect:    "none",
		PromptEffect:    "none",
		RawOutputLoaded: input.RawOutputLoaded || policyReview.RawOutputLoaded,
	}
}

func autoDelegationPlannerInstructionProfileBlock(result AutoDelegationPlannerInstructionProfile, status VerificationStatus, failure FailureClass, reason string, missing MissingInput, next NextHostAction, boundary Boundary) AutoDelegationPlannerInstructionProfile {
	result.Status = status
	result.FailureClass = failure
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = next
	result.Boundaries = AppendBoundaries(result.Boundaries, boundary)
	return result.Normalize()
}

func baseAutoDelegationPlannerCandidateJSONDecodeReport(input AutoDelegationPlannerCandidateJSONDecodeInput) AutoDelegationPlannerCandidateJSONDecodeReport {
	return AutoDelegationPlannerCandidateJSONDecodeReport{
		ContractVersion: ContractVersion,
		Projected:       true,
		Status:          VerificationBlocked,
		FailureClass:    FailureInvalidInput,
		MissingInputs:   cloneMissingInputs(input.MissingInputs),
		Boundaries: MergeBoundaries(
			[]Boundary{
				"auto_delegation_planner_candidate_json_decode",
				"strict_json_candidate",
				"code_fence_json_cleanup_allowed",
				"deterministic_validation_before_execution",
				"no_child_task_spawn",
				"no_subagent_dispatch",
				"no_scheduler_enqueue",
				"no_backend_execution",
			},
			input.Boundaries,
		),
		NextHostAction:  "provide_auto_delegation_planner_json",
		RunnerEffect:    "none",
		PromptEffect:    "none",
		RawOutputLoaded: input.RawOutputLoaded,
	}
}

func autoDelegationPlannerCandidateJSONDecodeBlock(result AutoDelegationPlannerCandidateJSONDecodeReport, status VerificationStatus, failure FailureClass, reason string, missing MissingInput, next NextHostAction, boundary Boundary) AutoDelegationPlannerCandidateJSONDecodeReport {
	result.Status = status
	result.FailureClass = failure
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = next
	result.Boundaries = AppendBoundaries(result.Boundaries, boundary)
	return result.Normalize()
}

func decodeAutoDelegationPlannerCandidateJSON(raw string) (AutoDelegationPlannerCandidate, bool, Boundary) {
	cleaned := strings.TrimSpace(stripAutoDelegationJSONFence(raw))
	if cleaned == "" {
		return AutoDelegationPlannerCandidate{}, false, "auto_delegation_planner_json_missing"
	}
	var candidate AutoDelegationPlannerCandidate
	if strictJSONDecode(cleaned, &candidate) == nil && !autoDelegationPlannerCandidatePlanEmpty(candidate.Plan) {
		return candidate, true, ""
	}
	var plan AutoDelegationPlan
	if strictJSONDecode(cleaned, &plan) == nil && !autoDelegationPlanEmpty(plan) {
		return AutoDelegationPlannerCandidate{Plan: plan}, true, ""
	}
	return AutoDelegationPlannerCandidate{}, false, "auto_delegation_planner_json_invalid"
}

func strictJSONDecode(raw string, target any) error {
	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return err
	}
	return nil
}

func autoDelegationPlannerCandidateApplyDefaults(candidate AutoDelegationPlannerCandidate, input AutoDelegationPlannerCandidateJSONDecodeInput) AutoDelegationPlannerCandidate {
	out := candidate.Clone()
	if out.SourceRef == "" {
		out.SourceRef = input.SourceRef
	}
	if out.PolicyRef == "" {
		out.PolicyRef = input.PolicyReview.Policy.PolicyRef
	}
	if out.Plan.PolicyRef == "" {
		out.Plan.PolicyRef = out.PolicyRef
	}
	if autoDelegationPolicyEmpty(out.Plan.Policy) {
		out.Plan.Policy = input.PolicyReview.Policy
	}
	if input.ParentObjectiveRef != "" {
		out.Plan.ParentObjectiveRef = input.ParentObjectiveRef
		for i := range out.Plan.Children {
			out.Plan.Children[i].ParentObjectiveRef = input.ParentObjectiveRef
		}
	} else if out.Plan.ParentObjectiveRef == "" {
		out.Plan.ParentObjectiveRef = input.ParentObjectiveRef
	}
	return out
}

func autoDelegationPlannerCandidatePlanEmpty(plan AutoDelegationPlan) bool {
	return autoDelegationPlanEmpty(plan)
}

func autoDelegationPlanEmpty(plan AutoDelegationPlan) bool {
	return plan.ContractVersion == "" &&
		plan.PlanRef == "" &&
		plan.PolicyRef == "" &&
		plan.ParentObjectiveRef == "" &&
		autoDelegationPolicyEmpty(plan.Policy) &&
		len(plan.Children) == 0 &&
		plan.MaxChildren == 0 &&
		plan.MaxParallelism == 0 &&
		plan.MaxDepth == 0 &&
		plan.MergeStrategy == "" &&
		plan.ConversationOwnership == "" &&
		len(plan.RequiredEvidence) == 0 &&
		len(plan.Boundaries) == 0 &&
		len(plan.MissingInputs) == 0 &&
		!plan.RawOutputLoaded
}

func autoDelegationPlannerInstructionInputUnsafe(input AutoDelegationPlannerInstructionInput) bool {
	return input.RawOutputLoaded ||
		input.PolicyReview.RawOutputLoaded ||
		displaySafeRefSliceRejected(input.AvailableToolRefs) ||
		displaySafeRefSliceRejected(input.AvailableCapabilityRefs)
}

func autoDelegationPlannerCandidateJSONDecodeInputUnsafe(input AutoDelegationPlannerCandidateJSONDecodeInput) bool {
	return input.RawOutputLoaded ||
		input.PolicyReview.RawOutputLoaded ||
		displaySafeRefRejected(input.SourceRef) ||
		displaySafeRefRejected(input.ParentObjectiveRef)
}

func autoDelegationPlannerRequiredOutputFields() []string {
	return []string{
		"plan_ref",
		"parent_objective_ref",
		"children",
		"child_ref",
		"goal",
		"capability_refs_or_allowed_tool_refs",
		"expected_output",
		"expected_evidence",
	}
}

func candidateBoundedReasons(candidate AutoDelegationPlannerCandidate) []string {
	if len(candidate.MissingInputs) == 0 {
		return nil
	}
	out := make([]string, 0, len(candidate.MissingInputs))
	for _, missing := range candidate.MissingInputs {
		token := normalizeControlToken("candidate_missing_" + string(missing))
		if token != "" {
			out = append(out, token)
		}
	}
	return out
}
