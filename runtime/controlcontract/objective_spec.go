package controlcontract

import "strings"

type ObjectiveSpecSideEffectPolicy string

const (
	ObjectiveSpecSideEffectUnspecified      ObjectiveSpecSideEffectPolicy = "unspecified"
	ObjectiveSpecSideEffectReadOnly         ObjectiveSpecSideEffectPolicy = "read_only"
	ObjectiveSpecSideEffectRequiresApproval ObjectiveSpecSideEffectPolicy = "requires_approval"
	ObjectiveSpecSideEffectForbidden        ObjectiveSpecSideEffectPolicy = "forbidden"
	ObjectiveSpecSideEffectAllowed          ObjectiveSpecSideEffectPolicy = "allowed"
)

func NormalizeObjectiveSpecSideEffectPolicy(raw string) ObjectiveSpecSideEffectPolicy {
	switch normalizeEnumToken(raw) {
	case "read_only", "readonly", "query_only", "no_write", "external_readonly":
		return ObjectiveSpecSideEffectReadOnly
	case "requires_approval", "approval_required", "needs_approval", "confirm_before_action":
		return ObjectiveSpecSideEffectRequiresApproval
	case "forbidden", "blocked", "deny", "no_side_effect", "no_external_action":
		return ObjectiveSpecSideEffectForbidden
	case "allowed", "approved", "host_allowed":
		return ObjectiveSpecSideEffectAllowed
	default:
		return ObjectiveSpecSideEffectUnspecified
	}
}

type ObjectiveSpecMissingInfoPolicy string

const (
	ObjectiveSpecMissingInfoUnspecified ObjectiveSpecMissingInfoPolicy = "unspecified"
	ObjectiveSpecMissingInfoDefaultSafe ObjectiveSpecMissingInfoPolicy = "default_safe"
	ObjectiveSpecMissingInfoQuerySafe   ObjectiveSpecMissingInfoPolicy = "query_safe"
	ObjectiveSpecMissingInfoAskUser     ObjectiveSpecMissingInfoPolicy = "ask_user"
	ObjectiveSpecMissingInfoBlock       ObjectiveSpecMissingInfoPolicy = "block"
)

func NormalizeObjectiveSpecMissingInfoPolicy(raw string) ObjectiveSpecMissingInfoPolicy {
	switch normalizeEnumToken(raw) {
	case "default_safe", "safe_default", "default":
		return ObjectiveSpecMissingInfoDefaultSafe
	case "query_safe", "safe_query", "retrieve":
		return ObjectiveSpecMissingInfoQuerySafe
	case "ask_user", "ask", "clarify", "clarification":
		return ObjectiveSpecMissingInfoAskUser
	case "block", "blocked", "fail_closed":
		return ObjectiveSpecMissingInfoBlock
	default:
		return ObjectiveSpecMissingInfoUnspecified
	}
}

type ObjectiveSuccessCriterion struct {
	CriteriaRef      DisplaySafeRef `json:"criteria_ref,omitempty"`
	Text             string         `json:"text,omitempty"`
	RequiredEvidence []EvidenceRef  `json:"required_evidence,omitempty"`
	AcceptsPartial   bool           `json:"accepts_partial"`
	Boundaries       []Boundary     `json:"boundaries,omitempty"`
	MissingInputs    []MissingInput `json:"missing_inputs,omitempty"`
}

func CloneObjectiveSuccessCriterion(in ObjectiveSuccessCriterion) ObjectiveSuccessCriterion {
	out := in
	out.RequiredEvidence = cloneEvidenceRefs(in.RequiredEvidence)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	return out
}

func (c ObjectiveSuccessCriterion) Clone() ObjectiveSuccessCriterion {
	return CloneObjectiveSuccessCriterion(c)
}

func (c ObjectiveSuccessCriterion) Normalize() ObjectiveSuccessCriterion {
	out := CloneObjectiveSuccessCriterion(c)
	out.CriteriaRef = normalizeOneDisplaySafeRef(out.CriteriaRef)
	out.Text = objectiveSpecSafeText(out.Text)
	out.RequiredEvidence = normalizeEvidenceRefs(out.RequiredEvidence)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	return out
}

type ObjectiveConstraint struct {
	ConstraintRef DisplaySafeRef `json:"constraint_ref,omitempty"`
	Kind          string         `json:"kind,omitempty"`
	Text          string         `json:"text,omitempty"`
	Boundaries    []Boundary     `json:"boundaries,omitempty"`
}

func CloneObjectiveConstraint(in ObjectiveConstraint) ObjectiveConstraint {
	out := in
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (c ObjectiveConstraint) Clone() ObjectiveConstraint {
	return CloneObjectiveConstraint(c)
}

func (c ObjectiveConstraint) Normalize() ObjectiveConstraint {
	out := CloneObjectiveConstraint(c)
	out.ConstraintRef = normalizeOneDisplaySafeRef(out.ConstraintRef)
	out.Kind = normalizeControlToken(out.Kind)
	out.Text = objectiveSpecSafeText(out.Text)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	return out
}

type ObjectiveSpecBudget struct {
	BudgetRef          DisplaySafeRef   `json:"budget_ref,omitempty"`
	MaxNodes           int              `json:"max_nodes,omitempty"`
	MaxAttempts        int              `json:"max_attempts,omitempty"`
	MaxDurationSeconds int              `json:"max_duration_seconds,omitempty"`
	MaxCostUnits       int              `json:"max_cost_units,omitempty"`
	PolicyRefs         []DisplaySafeRef `json:"policy_refs,omitempty"`
}

func CloneObjectiveSpecBudget(in ObjectiveSpecBudget) ObjectiveSpecBudget {
	out := in
	out.PolicyRefs = cloneDisplaySafeRefs(in.PolicyRefs)
	return out
}

func (b ObjectiveSpecBudget) Clone() ObjectiveSpecBudget {
	return CloneObjectiveSpecBudget(b)
}

func (b ObjectiveSpecBudget) Normalize() ObjectiveSpecBudget {
	out := CloneObjectiveSpecBudget(b)
	out.BudgetRef = normalizeOneDisplaySafeRef(out.BudgetRef)
	if out.MaxNodes < 0 {
		out.MaxNodes = 0
	}
	if out.MaxAttempts < 0 {
		out.MaxAttempts = 0
	}
	if out.MaxDurationSeconds < 0 {
		out.MaxDurationSeconds = 0
	}
	if out.MaxCostUnits < 0 {
		out.MaxCostUnits = 0
	}
	out.PolicyRefs = normalizeDisplaySafeRefs(out.PolicyRefs)
	return out
}

type ObjectiveSpec struct {
	ContractVersion       string                         `json:"contract_version,omitempty"`
	SpecRef               DisplaySafeRef                 `json:"spec_ref,omitempty"`
	ObjectiveID           string                         `json:"objective_id,omitempty"`
	UserGoalDigest        string                         `json:"user_goal_digest,omitempty"`
	RawGoalRef            DisplaySafeRef                 `json:"raw_goal_ref,omitempty"`
	GoalSummary           string                         `json:"goal_summary,omitempty"`
	ControlMode           ControlMode                    `json:"control_mode,omitempty"`
	Intensity             ExecutionIntensity             `json:"intensity,omitempty"`
	SuccessCriteria       []ObjectiveSuccessCriterion    `json:"success_criteria,omitempty"`
	Constraints           []ObjectiveConstraint          `json:"constraints,omitempty"`
	RequiredEvidence      []EvidenceRef                  `json:"required_evidence,omitempty"`
	CandidateCapabilities []DisplaySafeRef               `json:"candidate_capabilities,omitempty"`
	SourceContext         []DisplaySafeRef               `json:"source_context,omitempty"`
	SideEffectPolicy      ObjectiveSpecSideEffectPolicy  `json:"side_effect_policy,omitempty"`
	MissingInfoPolicy     ObjectiveSpecMissingInfoPolicy `json:"missing_info_policy,omitempty"`
	AcceptablePartial     []string                       `json:"acceptable_partial,omitempty"`
	Budget                ObjectiveSpecBudget            `json:"budget,omitempty"`
	ApprovalRefs          []DisplaySafeRef               `json:"approval_refs,omitempty"`
	PolicyRefs            []DisplaySafeRef               `json:"policy_refs,omitempty"`
	Boundaries            []Boundary                     `json:"boundaries,omitempty"`
	MissingInputs         []MissingInput                 `json:"missing_inputs,omitempty"`
	RawOutputLoaded       bool                           `json:"raw_output_loaded"`
}

func CloneObjectiveSpec(in ObjectiveSpec) ObjectiveSpec {
	out := in
	out.SuccessCriteria = cloneObjectiveSuccessCriteria(in.SuccessCriteria)
	out.Constraints = cloneObjectiveConstraints(in.Constraints)
	out.RequiredEvidence = cloneEvidenceRefs(in.RequiredEvidence)
	out.CandidateCapabilities = cloneDisplaySafeRefs(in.CandidateCapabilities)
	out.SourceContext = cloneDisplaySafeRefs(in.SourceContext)
	out.AcceptablePartial = cloneStringSlice(in.AcceptablePartial)
	out.Budget = in.Budget.Clone()
	out.ApprovalRefs = cloneDisplaySafeRefs(in.ApprovalRefs)
	out.PolicyRefs = cloneDisplaySafeRefs(in.PolicyRefs)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	return out
}

func (s ObjectiveSpec) Clone() ObjectiveSpec {
	return CloneObjectiveSpec(s)
}

func (s ObjectiveSpec) Normalize() ObjectiveSpec {
	out := CloneObjectiveSpec(s)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.SpecRef = normalizeOneDisplaySafeRef(out.SpecRef)
	out.ObjectiveID = strings.TrimSpace(out.ObjectiveID)
	if ContainsUnsafeRawOutput(out.ObjectiveID) {
		out.ObjectiveID = ""
	}
	out.UserGoalDigest = normalizeFingerprint(out.UserGoalDigest)
	out.RawGoalRef = normalizeOneDisplaySafeRef(out.RawGoalRef)
	out.GoalSummary = objectiveSpecSafeText(out.GoalSummary)
	out.ControlMode = NormalizeControlMode(string(out.ControlMode))
	if out.ControlMode == "" {
		out.ControlMode = ControlModeObjective
	}
	out.Intensity = NormalizeExecutionIntensity(string(out.Intensity))
	if out.Intensity == "" {
		out.Intensity = IntensityL3ManagedObjective
	}
	out.SuccessCriteria = normalizeObjectiveSuccessCriteria(out.SuccessCriteria)
	out.Constraints = normalizeObjectiveConstraints(out.Constraints)
	out.RequiredEvidence = normalizeEvidenceRefs(out.RequiredEvidence)
	out.CandidateCapabilities = normalizeDisplaySafeRefs(out.CandidateCapabilities)
	out.SourceContext = normalizeDisplaySafeRefs(out.SourceContext)
	out.AcceptablePartial = normalizeStringList(out.AcceptablePartial)
	out.SideEffectPolicy = NormalizeObjectiveSpecSideEffectPolicy(string(out.SideEffectPolicy))
	out.MissingInfoPolicy = NormalizeObjectiveSpecMissingInfoPolicy(string(out.MissingInfoPolicy))
	out.Budget = out.Budget.Normalize()
	out.ApprovalRefs = normalizeDisplaySafeRefs(out.ApprovalRefs)
	out.PolicyRefs = normalizeDisplaySafeRefs(out.PolicyRefs)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	return out
}

type ObjectiveSpecFrameProjectionInput struct {
	Spec            ObjectiveSpec  `json:"spec,omitempty"`
	ProjectionRef   DisplaySafeRef `json:"projection_ref,omitempty"`
	SourceRef       DisplaySafeRef `json:"source_ref,omitempty"`
	Boundaries      []Boundary     `json:"boundaries,omitempty"`
	RawOutputLoaded bool           `json:"raw_output_loaded"`
}

type ObjectiveSpecFrameProjection struct {
	ContractVersion string             `json:"contract_version,omitempty"`
	Projected       bool               `json:"projected"`
	Available       bool               `json:"available"`
	Status          VerificationStatus `json:"status,omitempty"`
	Mode            string             `json:"mode,omitempty"`
	RunnerEffect    string             `json:"runner_effect,omitempty"`
	PromptEffect    string             `json:"prompt_effect,omitempty"`
	ProjectionRef   DisplaySafeRef     `json:"projection_ref,omitempty"`
	SourceRef       DisplaySafeRef     `json:"source_ref,omitempty"`
	Spec            ObjectiveSpec      `json:"spec,omitempty"`
	Frame           ObjectiveFrame     `json:"frame,omitempty"`
	FrameMapped     bool               `json:"frame_mapped"`
	MappingWarnings []string           `json:"mapping_warnings,omitempty"`
	FailureClass    FailureClass       `json:"failure_class,omitempty"`
	MissingInputs   []MissingInput     `json:"missing_inputs,omitempty"`
	Boundaries      []Boundary         `json:"boundaries,omitempty"`
	NextHostAction  NextHostAction     `json:"next_host_action,omitempty"`
	RawOutputLoaded bool               `json:"raw_output_loaded"`
}

func BuildObjectiveSpecFrameProjection(input ObjectiveSpecFrameProjectionInput) ObjectiveSpecFrameProjection {
	result := baseObjectiveSpecFrameProjection(input)
	if objectiveSpecFrameProjectionInputUnsafe(input) {
		return objectiveSpecFrameProjectionBlock(result, VerificationReviewRequired, FailureEvidenceWeak, "host:display_safe_refs", "provide_display_safe_refs", "raw_output_not_allowed")
	}

	result.Frame = objectiveSpecFrameFromSpec(result.Spec)
	result.MappingWarnings = objectiveSpecFrameMappingWarnings(result.Spec)
	result.MissingInputs = MergeMissingInputs(result.MissingInputs, objectiveSpecFrameProjectionMissingInputs(result.Spec, result.Frame))
	if len(result.MissingInputs) > 0 {
		result.FailureClass = objectiveSpecFrameProjectionFailure(result.MissingInputs)
		result.NextHostAction = objectiveSpecFrameProjectionNextAction(result.MissingInputs)
		result.Boundaries = AppendBoundaries(result.Boundaries, "objective_spec_frame_projection_incomplete")
		return result.Normalize()
	}

	result.Status = VerificationSatisfied
	result.FrameMapped = true
	result.FailureClass = FailureNone
	result.NextHostAction = "run_objective_graph_planner"
	result.Boundaries = AppendBoundaries(result.Boundaries, "objective_spec_frame_projection_ready")
	return result.Normalize()
}

func (p ObjectiveSpecFrameProjection) Clone() ObjectiveSpecFrameProjection {
	out := p
	out.Spec = p.Spec.Clone()
	out.Frame = p.Frame.Clone()
	out.MappingWarnings = cloneStringSlice(p.MappingWarnings)
	out.MissingInputs = cloneMissingInputs(p.MissingInputs)
	out.Boundaries = cloneBoundaries(p.Boundaries)
	return out
}

func (p ObjectiveSpecFrameProjection) Normalize() ObjectiveSpecFrameProjection {
	out := p.Clone()
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Available = true
	out.Status = NormalizeVerificationStatus(string(out.Status))
	if out.Status == VerificationNotEvaluated {
		out.Status = VerificationBlocked
	}
	out.Mode = normalizeControlToken(out.Mode)
	if out.Mode == "" {
		out.Mode = "objective_spec_frame_projection"
	}
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	if out.RunnerEffect == "" {
		out.RunnerEffect = "none"
	}
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	if out.PromptEffect == "" {
		out.PromptEffect = "none"
	}
	out.ProjectionRef = normalizeOneDisplaySafeRef(out.ProjectionRef)
	out.SourceRef = normalizeOneDisplaySafeRef(out.SourceRef)
	out.Spec = out.Spec.Normalize()
	out.Frame = out.Frame.Normalize()
	out.MappingWarnings = normalizeStringList(out.MappingWarnings)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	if out.RawOutputLoaded || out.Spec.RawOutputLoaded {
		out.Status = VerificationReviewRequired
		out.FrameMapped = false
		out.FailureClass = firstFailureClass(out.FailureClass, FailureEvidenceWeak)
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		out.NextHostAction = "provide_display_safe_refs"
	}
	if out.Status != VerificationSatisfied {
		out.FrameMapped = false
	}
	return out
}

func baseObjectiveSpecFrameProjection(input ObjectiveSpecFrameProjectionInput) ObjectiveSpecFrameProjection {
	spec := input.Spec.Normalize()
	return ObjectiveSpecFrameProjection{
		ContractVersion: ContractVersion,
		Projected:       true,
		Available:       true,
		Status:          VerificationBlocked,
		Mode:            "objective_spec_frame_projection",
		RunnerEffect:    "none",
		PromptEffect:    "none",
		ProjectionRef:   normalizeOneDisplaySafeRef(input.ProjectionRef),
		SourceRef:       normalizeOneDisplaySafeRef(input.SourceRef),
		Spec:            spec,
		Frame:           objectiveSpecFrameFromSpec(spec),
		FailureClass:    FailureInsufficientInformation,
		Boundaries: AppendBoundaries(
			[]Boundary{
				"objective_spec_frame_projection",
				"objective_spec_normalized_to_objective_frame",
				"no_second_runner",
				"no_parallel_catalog",
				"no_workflow_engine_replacement",
				"no_llm_call",
				"no_runner_dispatch",
				"no_backend_execution",
			},
			input.Boundaries...,
		),
		NextHostAction:  "provide_objective_spec",
		RawOutputLoaded: input.RawOutputLoaded || spec.RawOutputLoaded,
	}
}

func objectiveSpecFrameProjectionBlock(result ObjectiveSpecFrameProjection, status VerificationStatus, failure FailureClass, missing MissingInput, next NextHostAction, boundary Boundary) ObjectiveSpecFrameProjection {
	result.Status = status
	result.FailureClass = failure
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = next
	result.Boundaries = AppendBoundaries(result.Boundaries, boundary)
	return result.Normalize()
}

func objectiveSpecFrameFromSpec(spec ObjectiveSpec) ObjectiveFrame {
	normalized := spec.Normalize()
	criteria := make([]string, 0, len(normalized.SuccessCriteria))
	requiredEvidence := cloneEvidenceRefs(normalized.RequiredEvidence)
	for _, criterion := range normalized.SuccessCriteria {
		if criterion.Text != "" {
			criteria = append(criteria, criterion.Text)
		}
		requiredEvidence = MergeEvidenceRefs(requiredEvidence, criterion.RequiredEvidence)
	}
	constraints := objectiveSpecFrameConstraints(normalized)
	sourceContext := cloneDisplaySafeRefs(normalized.SourceContext)
	if normalized.RawGoalRef != "" {
		sourceContext = appendUniqueDisplaySafeRef(sourceContext, normalized.RawGoalRef)
	}
	if normalized.SpecRef != "" {
		sourceContext = appendUniqueDisplaySafeRef(sourceContext, normalized.SpecRef)
	}
	return ObjectiveFrame{
		ID:                    normalized.ObjectiveID,
		UserGoalDigest:        normalized.UserGoalDigest,
		ControlMode:           normalized.ControlMode,
		Intensity:             normalized.Intensity,
		SuccessCriteria:       criteria,
		Constraints:           constraints,
		RequiredEvidence:      requiredEvidence,
		CandidateCapabilities: normalized.CandidateCapabilities,
		SourceContext:         sourceContext,
		Boundaries:            objectiveSpecFrameBoundaries(normalized),
		MissingInputs:         normalized.MissingInputs,
	}.Normalize()
}

func objectiveSpecFrameConstraints(spec ObjectiveSpec) []string {
	values := make([]string, 0, len(spec.Constraints)+4+len(spec.AcceptablePartial))
	if spec.GoalSummary != "" {
		values = append(values, "goal_summary:"+spec.GoalSummary)
	}
	for _, constraint := range spec.Constraints {
		text := constraint.Text
		if constraint.Kind != "" && text != "" {
			text = constraint.Kind + ":" + text
		}
		values = append(values, text)
	}
	if spec.SideEffectPolicy != ObjectiveSpecSideEffectUnspecified {
		values = append(values, "side_effect_policy:"+string(spec.SideEffectPolicy))
	}
	if spec.MissingInfoPolicy != ObjectiveSpecMissingInfoUnspecified {
		values = append(values, "missing_info_policy:"+string(spec.MissingInfoPolicy))
	}
	for _, partial := range spec.AcceptablePartial {
		values = append(values, "acceptable_partial:"+partial)
	}
	return normalizeStringList(values)
}

func objectiveSpecFrameBoundaries(spec ObjectiveSpec) []Boundary {
	boundaries := MergeBoundaries(
		[]Boundary{
			"objective_spec_to_objective_frame",
			"objective_spec_projection_not_parallel_runtime",
		},
		spec.Boundaries,
	)
	if spec.SideEffectPolicy != ObjectiveSpecSideEffectUnspecified {
		boundaries = AppendBoundaries(boundaries, Boundary("objective_spec_side_effect_"+string(spec.SideEffectPolicy)))
	}
	if spec.MissingInfoPolicy != ObjectiveSpecMissingInfoUnspecified {
		boundaries = AppendBoundaries(boundaries, Boundary("objective_spec_missing_info_"+string(spec.MissingInfoPolicy)))
	}
	if spec.Budget.BudgetRef != "" || spec.Budget.MaxNodes > 0 || spec.Budget.MaxAttempts > 0 || spec.Budget.MaxDurationSeconds > 0 || spec.Budget.MaxCostUnits > 0 {
		boundaries = AppendBoundaries(boundaries, "objective_spec_budget_carried_by_spec")
	}
	if len(spec.ApprovalRefs) > 0 {
		boundaries = AppendBoundaries(boundaries, "objective_spec_approval_refs_carried_by_spec")
	}
	return boundaries
}

func objectiveSpecFrameProjectionMissingInputs(spec ObjectiveSpec, frame ObjectiveFrame) []MissingInput {
	var missing []MissingInput
	if spec.GoalSummary == "" {
		missing = AppendMissingInputs(missing, "host:objective_goal_summary")
	}
	if spec.RawGoalRef == "" && spec.UserGoalDigest == "" {
		missing = AppendMissingInputs(missing, "host:objective_raw_goal_ref")
	}
	if len(frame.SuccessCriteria) == 0 {
		missing = AppendMissingInputs(missing, "host:success_criteria")
	}
	if len(frame.RequiredEvidence) == 0 {
		missing = AppendMissingInputs(missing, "host:required_evidence_contract")
	}
	if spec.SideEffectPolicy == ObjectiveSpecSideEffectUnspecified {
		missing = AppendMissingInputs(missing, "host:objective_side_effect_policy")
	}
	if spec.SideEffectPolicy == ObjectiveSpecSideEffectRequiresApproval && len(spec.ApprovalRefs) == 0 {
		missing = AppendMissingInputs(missing, "host:objective_approval_ref")
	}
	missing = AppendMissingInputs(missing, spec.MissingInputs...)
	return missing
}

func objectiveSpecFrameProjectionFailure(missing []MissingInput) FailureClass {
	for _, value := range missing {
		switch value {
		case "host:display_safe_refs":
			return FailureEvidenceWeak
		case "host:required_evidence_contract":
			return FailureEvidenceMissing
		case "host:objective_approval_ref":
			return FailureApprovalRequired
		case "host:objective_side_effect_policy":
			return FailurePolicyBlocked
		}
	}
	return FailureInsufficientInformation
}

func objectiveSpecFrameProjectionNextAction(missing []MissingInput) NextHostAction {
	for _, value := range missing {
		switch value {
		case "host:required_evidence_contract":
			return "provide_required_evidence_contract"
		case "host:objective_approval_ref":
			return "request_host_approval"
		case "host:objective_side_effect_policy":
			return "provide_objective_contract"
		}
	}
	return "provide_objective_spec"
}

func objectiveSpecFrameMappingWarnings(spec ObjectiveSpec) []string {
	var warnings []string
	if spec.Budget.BudgetRef != "" || spec.Budget.MaxNodes > 0 || spec.Budget.MaxAttempts > 0 || spec.Budget.MaxDurationSeconds > 0 || spec.Budget.MaxCostUnits > 0 {
		warnings = append(warnings, "budget_stays_on_objective_spec")
	}
	if len(spec.ApprovalRefs) > 0 {
		warnings = append(warnings, "approval_refs_stay_on_objective_spec")
	}
	if len(spec.PolicyRefs) > 0 {
		warnings = append(warnings, "policy_refs_stay_on_objective_spec")
	}
	return normalizeStringList(warnings)
}

func objectiveSpecFrameProjectionInputUnsafe(input ObjectiveSpecFrameProjectionInput) bool {
	return input.RawOutputLoaded ||
		displaySafeRefRejected(input.ProjectionRef) ||
		displaySafeRefRejected(input.SourceRef) ||
		objectiveSpecUnsafe(input.Spec)
}

func objectiveSpecUnsafe(spec ObjectiveSpec) bool {
	if spec.RawOutputLoaded ||
		displaySafeRefRejected(spec.SpecRef) ||
		displaySafeRefRejected(spec.RawGoalRef) ||
		displaySafeRefSliceRejected(spec.CandidateCapabilities) ||
		displaySafeRefSliceRejected(spec.SourceContext) ||
		displaySafeRefSliceRejected(spec.ApprovalRefs) ||
		displaySafeRefSliceRejected(spec.PolicyRefs) ||
		evidenceRefRejected(spec.RequiredEvidence) ||
		ContainsUnsafeRawOutput(spec.ObjectiveID, spec.UserGoalDigest, spec.GoalSummary) {
		return true
	}
	if displaySafeRefRejected(spec.Budget.BudgetRef) || displaySafeRefSliceRejected(spec.Budget.PolicyRefs) {
		return true
	}
	for _, criterion := range spec.SuccessCriteria {
		if displaySafeRefRejected(criterion.CriteriaRef) ||
			evidenceRefRejected(criterion.RequiredEvidence) ||
			ContainsUnsafeRawOutput(criterion.Text) {
			return true
		}
	}
	for _, constraint := range spec.Constraints {
		if displaySafeRefRejected(constraint.ConstraintRef) ||
			ContainsUnsafeRawOutput(constraint.Kind, constraint.Text) {
			return true
		}
	}
	for _, partial := range spec.AcceptablePartial {
		if ContainsUnsafeRawOutput(partial) {
			return true
		}
	}
	return false
}

func normalizeObjectiveSuccessCriteria(in []ObjectiveSuccessCriterion) []ObjectiveSuccessCriterion {
	out := make([]ObjectiveSuccessCriterion, 0, len(in))
	for _, criterion := range in {
		normalized := criterion.Normalize()
		if normalized.Text == "" && normalized.CriteriaRef == "" && len(normalized.RequiredEvidence) == 0 {
			continue
		}
		out = append(out, normalized)
	}
	return out
}

func cloneObjectiveSuccessCriteria(in []ObjectiveSuccessCriterion) []ObjectiveSuccessCriterion {
	if len(in) == 0 {
		return nil
	}
	out := make([]ObjectiveSuccessCriterion, 0, len(in))
	for _, criterion := range in {
		out = append(out, criterion.Clone())
	}
	return out
}

func normalizeObjectiveConstraints(in []ObjectiveConstraint) []ObjectiveConstraint {
	out := make([]ObjectiveConstraint, 0, len(in))
	for _, constraint := range in {
		normalized := constraint.Normalize()
		if normalized.Text == "" && normalized.ConstraintRef == "" && normalized.Kind == "" {
			continue
		}
		out = append(out, normalized)
	}
	return out
}

func cloneObjectiveConstraints(in []ObjectiveConstraint) []ObjectiveConstraint {
	if len(in) == 0 {
		return nil
	}
	out := make([]ObjectiveConstraint, 0, len(in))
	for _, constraint := range in {
		out = append(out, constraint.Clone())
	}
	return out
}

func objectiveSpecSafeText(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || ContainsUnsafeRawOutput(trimmed) {
		return ""
	}
	return trimmed
}
