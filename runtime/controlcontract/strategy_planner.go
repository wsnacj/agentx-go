package controlcontract

import "sort"

type StrategyCatalogSourceKind string

const (
	StrategyCatalogSourceTool        StrategyCatalogSourceKind = "tool"
	StrategyCatalogSourceSkill       StrategyCatalogSourceKind = "skill"
	StrategyCatalogSourceWorkflow    StrategyCatalogSourceKind = "workflow"
	StrategyCatalogSourceScene       StrategyCatalogSourceKind = "scene"
	StrategyCatalogSourceHostAdapter StrategyCatalogSourceKind = "host_adapter"
	StrategyCatalogSourceProject     StrategyCatalogSourceKind = "project_adapter"
	StrategyCatalogSourceOperations  StrategyCatalogSourceKind = "operations"
	StrategyCatalogSourceCapability  StrategyCatalogSourceKind = "capability"
	StrategyCatalogSourceDelegation  StrategyCatalogSourceKind = "delegation"
)

func KnownStrategyCatalogSourceKinds() []StrategyCatalogSourceKind {
	return []StrategyCatalogSourceKind{
		StrategyCatalogSourceTool,
		StrategyCatalogSourceSkill,
		StrategyCatalogSourceWorkflow,
		StrategyCatalogSourceScene,
		StrategyCatalogSourceHostAdapter,
		StrategyCatalogSourceProject,
		StrategyCatalogSourceOperations,
		StrategyCatalogSourceCapability,
		StrategyCatalogSourceDelegation,
	}
}

func NormalizeStrategyCatalogSourceKind(raw string) StrategyCatalogSourceKind {
	switch normalizeEnumToken(raw) {
	case "tool", "tools":
		return StrategyCatalogSourceTool
	case "skill", "skills":
		return StrategyCatalogSourceSkill
	case "workflow", "workflow_pack":
		return StrategyCatalogSourceWorkflow
	case "scene", "scene_package", "scene_module":
		return StrategyCatalogSourceScene
	case "host_adapter", "adapter", "host":
		return StrategyCatalogSourceHostAdapter
	case "project_adapter", "project", "project_strategy", "project_host_adapter":
		return StrategyCatalogSourceProject
	case "operations", "operation", "ops":
		return StrategyCatalogSourceOperations
	case "capability", "capability_resolution", "capability_install":
		return StrategyCatalogSourceCapability
	case "delegation", "delegated", "delegate", "worker", "worker_runtime", "delegation_worker":
		return StrategyCatalogSourceDelegation
	default:
		return ""
	}
}

type StrategyCatalogEntry struct {
	ContractVersion string                    `json:"contract_version,omitempty"`
	SourceKind      StrategyCatalogSourceKind `json:"source_kind,omitempty"`
	SourceRef       DisplaySafeRef            `json:"source_ref,omitempty"`
	Candidate       StrategyCandidate         `json:"candidate,omitempty"`
	Status          VerificationStatus        `json:"status,omitempty"`
	FailureClass    FailureClass              `json:"failure_class,omitempty"`
	EvidenceRefs    []EvidenceRef             `json:"evidence_refs,omitempty"`
	MissingInputs   []MissingInput            `json:"missing_inputs,omitempty"`
	Boundaries      []Boundary                `json:"boundaries,omitempty"`
	RawOutputLoaded bool                      `json:"raw_output_loaded"`
}

func CloneStrategyCatalogEntry(in StrategyCatalogEntry) StrategyCatalogEntry {
	out := in
	out.Candidate = in.Candidate.Clone()
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (e StrategyCatalogEntry) Clone() StrategyCatalogEntry {
	return CloneStrategyCatalogEntry(e)
}

func (e StrategyCatalogEntry) Normalize() StrategyCatalogEntry {
	out := CloneStrategyCatalogEntry(e)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.SourceKind = NormalizeStrategyCatalogSourceKind(string(out.SourceKind))
	out.SourceRef = normalizeOneDisplaySafeRef(out.SourceRef)
	out.Candidate = out.Candidate.Normalize()
	out.Status = NormalizeVerificationStatus(string(out.Status))
	if out.Status == VerificationNotEvaluated {
		out.Status = VerificationSatisfied
	}
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	return out
}

type StrategyCatalogSnapshot struct {
	ContractVersion string                 `json:"contract_version,omitempty"`
	Projected       bool                   `json:"projected"`
	CatalogRef      DisplaySafeRef         `json:"catalog_ref,omitempty"`
	Entries         []StrategyCatalogEntry `json:"entries,omitempty"`
	SourceRefs      []DisplaySafeRef       `json:"source_refs,omitempty"`
	PolicyRefs      []DisplaySafeRef       `json:"policy_refs,omitempty"`
	EvidenceRefs    []EvidenceRef          `json:"evidence_refs,omitempty"`
	MissingInputs   []MissingInput         `json:"missing_inputs,omitempty"`
	Boundaries      []Boundary             `json:"boundaries,omitempty"`
	RawOutputLoaded bool                   `json:"raw_output_loaded"`
}

func CloneStrategyCatalogSnapshot(in StrategyCatalogSnapshot) StrategyCatalogSnapshot {
	out := in
	out.Entries = cloneStrategyCatalogEntries(in.Entries)
	out.SourceRefs = cloneDisplaySafeRefs(in.SourceRefs)
	out.PolicyRefs = cloneDisplaySafeRefs(in.PolicyRefs)
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (s StrategyCatalogSnapshot) Clone() StrategyCatalogSnapshot {
	return CloneStrategyCatalogSnapshot(s)
}

func (s StrategyCatalogSnapshot) Normalize() StrategyCatalogSnapshot {
	out := CloneStrategyCatalogSnapshot(s)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.CatalogRef = normalizeOneDisplaySafeRef(out.CatalogRef)
	out.Entries = normalizeStrategyCatalogEntries(out.Entries)
	out.SourceRefs = normalizeDisplaySafeRefs(out.SourceRefs)
	for _, entry := range out.Entries {
		out.SourceRefs = appendUniqueDisplaySafeRef(out.SourceRefs, entry.SourceRef)
		out.EvidenceRefs = MergeEvidenceRefs(out.EvidenceRefs, entry.EvidenceRefs, entry.Candidate.ExpectedEvidence)
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, entry.MissingInputs...)
		out.Boundaries = MergeBoundaries(out.Boundaries, entry.Boundaries)
		out.RawOutputLoaded = out.RawOutputLoaded || entry.RawOutputLoaded
	}
	out.PolicyRefs = normalizeDisplaySafeRefs(out.PolicyRefs)
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	if out.CatalogRef == "" {
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:strategy_catalog_ref")
		out.Boundaries = AppendBoundaries(out.Boundaries, "strategy_catalog_ref_missing")
	}
	if len(out.Entries) == 0 {
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:strategy_catalog")
		out.Boundaries = AppendBoundaries(out.Boundaries, "strategy_catalog_empty")
	}
	return out
}

type StrategyPlannerInput struct {
	Activation              Activation               `json:"activation,omitempty"`
	Frame                   ObjectiveFrame           `json:"frame,omitempty"`
	Policy                  ExecutionIntensityPolicy `json:"policy,omitempty"`
	PreGate                 IntensityGateResult      `json:"pre_gate,omitempty"`
	Catalog                 StrategyCatalogSnapshot  `json:"catalog,omitempty"`
	CurrentStrategyRef      DisplaySafeRef           `json:"current_strategy_ref,omitempty"`
	AvailableCapabilityRefs []DisplaySafeRef         `json:"available_capability_refs,omitempty"`
	SatisfiedPreconditions  []MissingInput           `json:"satisfied_preconditions,omitempty"`
	Attempts                []AttemptSummary         `json:"attempts,omitempty"`
	Verification            VerificationResult       `json:"verification,omitempty"`
	EvidenceRefs            []EvidenceRef            `json:"evidence_refs,omitempty"`
	PolicyRefs              []DisplaySafeRef         `json:"policy_refs,omitempty"`
	DecisionBasis           []DisplaySafeRef         `json:"decision_basis,omitempty"`
	Boundaries              []Boundary               `json:"boundaries,omitempty"`
	RawOutputLoaded         bool                     `json:"raw_output_loaded"`
}

type StrategyPlanCandidate struct {
	ContractVersion string                    `json:"contract_version,omitempty"`
	Rank            int                       `json:"rank,omitempty"`
	Score           int                       `json:"score,omitempty"`
	SourceKind      StrategyCatalogSourceKind `json:"source_kind,omitempty"`
	SourceRef       DisplaySafeRef            `json:"source_ref,omitempty"`
	Candidate       StrategyCandidate         `json:"candidate,omitempty"`
	Status          VerificationStatus        `json:"status,omitempty"`
	FailureClass    FailureClass              `json:"failure_class,omitempty"`
	MissingInputs   []MissingInput            `json:"missing_inputs,omitempty"`
	EvidenceRefs    []EvidenceRef             `json:"evidence_refs,omitempty"`
	DecisionBasis   []DisplaySafeRef          `json:"decision_basis,omitempty"`
	Boundaries      []Boundary                `json:"boundaries,omitempty"`
	NextHostAction  NextHostAction            `json:"next_host_action,omitempty"`
	RawOutputLoaded bool                      `json:"raw_output_loaded"`
}

func CloneStrategyPlanCandidate(in StrategyPlanCandidate) StrategyPlanCandidate {
	out := in
	out.Candidate = in.Candidate.Clone()
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.DecisionBasis = cloneDisplaySafeRefs(in.DecisionBasis)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (c StrategyPlanCandidate) Clone() StrategyPlanCandidate {
	return CloneStrategyPlanCandidate(c)
}

func (c StrategyPlanCandidate) Normalize() StrategyPlanCandidate {
	out := CloneStrategyPlanCandidate(c)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	if out.Rank < 0 {
		out.Rank = 0
	}
	out.SourceKind = NormalizeStrategyCatalogSourceKind(string(out.SourceKind))
	out.SourceRef = normalizeOneDisplaySafeRef(out.SourceRef)
	out.Candidate = out.Candidate.Normalize()
	out.Status = NormalizeVerificationStatus(string(out.Status))
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
	out.DecisionBasis = normalizeDisplaySafeRefs(out.DecisionBasis)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	if out.Status == VerificationNotEvaluated {
		out.Status = VerificationBlocked
	}
	return out
}

type StrategyPlannerResult struct {
	ContractVersion     string                   `json:"contract_version,omitempty"`
	Projected           bool                     `json:"projected"`
	Activation          Activation               `json:"activation,omitempty"`
	Status              VerificationStatus       `json:"status,omitempty"`
	Frame               ObjectiveFrame           `json:"frame,omitempty"`
	Policy              ExecutionIntensityPolicy `json:"policy,omitempty"`
	PreGate             IntensityGateResult      `json:"pre_gate,omitempty"`
	CatalogRef          DisplaySafeRef           `json:"catalog_ref,omitempty"`
	MaxAllowedIntensity ExecutionIntensity       `json:"max_allowed_intensity,omitempty"`
	CurrentStrategyRef  DisplaySafeRef           `json:"current_strategy_ref,omitempty"`
	Selected            StrategyPlanCandidate    `json:"selected,omitempty"`
	RankedCandidates    []StrategyPlanCandidate  `json:"ranked_candidates,omitempty"`
	RejectedCandidates  []StrategyPlanCandidate  `json:"rejected_candidates,omitempty"`
	PolicyRefs          []DisplaySafeRef         `json:"policy_refs,omitempty"`
	EvidenceRefs        []EvidenceRef            `json:"evidence_refs,omitempty"`
	FailureClass        FailureClass             `json:"failure_class,omitempty"`
	MissingInputs       []MissingInput           `json:"missing_inputs,omitempty"`
	DecisionBasis       []DisplaySafeRef         `json:"decision_basis,omitempty"`
	Boundaries          []Boundary               `json:"boundaries,omitempty"`
	NextHostAction      NextHostAction           `json:"next_host_action,omitempty"`
	RunnerEffect        string                   `json:"runner_effect,omitempty"`
	PromptEffect        string                   `json:"prompt_effect,omitempty"`
	RawOutputLoaded     bool                     `json:"raw_output_loaded"`
}

func CloneStrategyPlannerResult(in StrategyPlannerResult) StrategyPlannerResult {
	out := in
	out.Frame = in.Frame.Clone()
	out.Policy = in.Policy.Clone()
	out.PreGate = in.PreGate.Clone()
	out.Selected = in.Selected.Clone()
	out.RankedCandidates = cloneStrategyPlanCandidates(in.RankedCandidates)
	out.RejectedCandidates = cloneStrategyPlanCandidates(in.RejectedCandidates)
	out.PolicyRefs = cloneDisplaySafeRefs(in.PolicyRefs)
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.DecisionBasis = cloneDisplaySafeRefs(in.DecisionBasis)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (r StrategyPlannerResult) Clone() StrategyPlannerResult {
	return CloneStrategyPlannerResult(r)
}

func (r StrategyPlannerResult) Normalize() StrategyPlannerResult {
	out := CloneStrategyPlannerResult(r)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Activation = NormalizeActivation(string(out.Activation))
	out.Status = NormalizeVerificationStatus(string(out.Status))
	out.Frame = out.Frame.Normalize()
	out.Policy = out.Policy.Normalize()
	out.PreGate = out.PreGate.Normalize()
	out.CatalogRef = normalizeOneDisplaySafeRef(out.CatalogRef)
	out.MaxAllowedIntensity = NormalizeExecutionIntensity(string(out.MaxAllowedIntensity))
	out.CurrentStrategyRef = normalizeOneDisplaySafeRef(out.CurrentStrategyRef)
	out.Selected = out.Selected.Normalize()
	out.RankedCandidates = normalizeStrategyPlanCandidates(out.RankedCandidates)
	out.RejectedCandidates = normalizeStrategyPlanCandidates(out.RejectedCandidates)
	out.PolicyRefs = normalizeDisplaySafeRefs(out.PolicyRefs)
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
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
	if out.RawOutputLoaded && out.Status != VerificationBlocked {
		out.Status = VerificationReviewRequired
		if out.FailureClass == FailureNone {
			out.FailureClass = FailureEvidenceWeak
		}
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		out.NextHostAction = "provide_display_safe_refs"
	}
	return out
}

func BuildStrategyPlanner(input StrategyPlannerInput) StrategyPlannerResult {
	frame := input.Frame.Normalize()
	policy := input.Policy.Normalize()
	preGate := input.PreGate.Normalize()
	catalog := input.Catalog.Normalize()
	verification := input.Verification.Normalize()
	activation := intensityGateActivation(input.Activation, preGate.Activation)
	maxAllowed := firstIntensity(preGate.ApprovedIntensity, preGate.MaxAllowedIntensity, policy.MaxAllowedIntensity)
	result := StrategyPlannerResult{
		ContractVersion:     ContractVersion,
		Projected:           true,
		Activation:          activation,
		Status:              VerificationSatisfied,
		Frame:               frame,
		Policy:              policy,
		PreGate:             preGate,
		CatalogRef:          catalog.CatalogRef,
		MaxAllowedIntensity: maxAllowed,
		CurrentStrategyRef:  normalizeOneDisplaySafeRef(input.CurrentStrategyRef),
		PolicyRefs: normalizeDisplaySafeRefs(append(
			append(cloneDisplaySafeRefs(input.PolicyRefs), policy.PolicyRefs...),
			catalog.PolicyRefs...,
		)),
		EvidenceRefs: MergeEvidenceRefs(input.EvidenceRefs, catalog.EvidenceRefs, preGate.EvidenceRefs, verification.EvidenceRefs),
		FailureClass: FailureNone,
		MissingInputs: MergeMissingInputs(
			catalog.MissingInputs,
			preGate.MissingInputs,
			policy.MissingInputs,
		),
		DecisionBasis: normalizeDisplaySafeRefs(append(
			cloneDisplaySafeRefs(input.DecisionBasis),
			"strategy_planner:metadata_only",
			DisplaySafeRef("max_allowed_intensity:"+string(maxAllowed)),
		)),
		Boundaries: MergeBoundaries(
			[]Boundary{
				"strategy_candidate_planner",
				"metadata_only_planner",
				"projection_only",
				"no_runner_dispatch",
				"no_strategy_dispatch",
				"no_prompt_effect",
				"planner_does_not_authorize_execution",
				"candidate_presence_is_not_capability_availability",
				"core_must_not_parse_goal_text",
			},
			input.Boundaries,
			catalog.Boundaries,
			preGate.Boundaries,
		),
		RunnerEffect:    "none",
		PromptEffect:    "none",
		RawOutputLoaded: input.RawOutputLoaded || catalog.RawOutputLoaded || preGate.RawOutputLoaded || verification.RawOutputLoaded,
	}
	if activation != ActivationManaged {
		return strategyPlannerBlock(result, FailurePolicyBlocked, "managed activation is required for strategy planning", "control_plane:managed_activation", "enable_managed_objective", "strategy_planner_activation_required")
	}
	if !preGate.Allowed {
		result.FailureClass = firstFailureClass(preGate.FailureClass, FailurePolicyBlocked)
		return strategyPlannerBlock(result, result.FailureClass, "intensity pre-gate must be satisfied before strategy planning", "control_plane:intensity_pre_gate", "satisfy_intensity_pre_gate", "strategy_planner_requires_pre_gate")
	}
	if len(catalog.Entries) == 0 {
		return strategyPlannerBlock(result, FailureConfigMissing, "strategy catalog is empty", "host:strategy_catalog", "provide_strategy_catalog", "strategy_catalog_empty")
	}

	availableCapabilities := normalizeDisplaySafeRefs(input.AvailableCapabilityRefs)
	satisfiedPreconditions := normalizeMissingInputs(input.SatisfiedPreconditions)
	attempts := normalizeAttemptSummaries(input.Attempts)
	entries := catalog.Entries
	evaluated := make([]StrategyPlanCandidate, 0, len(entries))
	rejected := []StrategyPlanCandidate{}
	for _, entry := range entries {
		plan := evaluateStrategyCatalogEntry(strategyPlannerEvaluationInput{
			Entry:                   entry,
			Policy:                  policy,
			PreGate:                 preGate,
			Frame:                   frame,
			CurrentStrategyRef:      result.CurrentStrategyRef,
			AvailableCapabilityRefs: availableCapabilities,
			SatisfiedPreconditions:  satisfiedPreconditions,
			Attempts:                attempts,
			EvidenceRefs:            result.EvidenceRefs,
			MaxAllowedIntensity:     maxAllowed,
		})
		if plan.Status == VerificationSatisfied || plan.Status == VerificationReviewRequired {
			evaluated = append(evaluated, plan)
			continue
		}
		rejected = append(rejected, plan)
	}
	sortStrategyPlanCandidates(evaluated)
	sortStrategyPlanCandidates(rejected)
	for i := range evaluated {
		evaluated[i].Rank = i + 1
	}
	for i := range rejected {
		rejected[i].Rank = i + 1
	}
	result.RankedCandidates = evaluated
	result.RejectedCandidates = rejected
	if len(evaluated) == 0 {
		result.FailureClass = firstStrategyPlanFailureClass(rejected, FailureConfigMissing)
		for _, candidate := range rejected {
			result.MissingInputs = AppendMissingInputs(result.MissingInputs, candidate.MissingInputs...)
		}
		if len(result.MissingInputs) == 0 {
			result.MissingInputs = AppendMissingInputs(result.MissingInputs, firstStrategyPlanMissingInput(rejected, "host:strategy_candidate"))
		}
		result.Boundaries = AppendBoundaries(result.Boundaries, "strategy_planner_no_selectable_candidate")
		result.NextHostAction = "provide_strategy_catalog_or_upgrade"
		result.Status = VerificationBlocked
		return result.Normalize()
	}
	result.Selected = evaluated[0]
	result.FailureClass = firstFailureClass(evaluated[0].FailureClass, FailureNone)
	if evaluated[0].Status == VerificationReviewRequired {
		result.Status = VerificationReviewRequired
		result.MissingInputs = AppendMissingInputs(result.MissingInputs, evaluated[0].MissingInputs...)
		result.NextHostAction = firstNextHostAction(evaluated[0].NextHostAction, "run_strategy_final_gate")
	} else {
		result.Status = VerificationSatisfied
		result.NextHostAction = "run_strategy_final_gate"
	}
	result.DecisionBasis = normalizeDisplaySafeRefs(append(result.DecisionBasis, evaluated[0].DecisionBasis...))
	result.Boundaries = AppendBoundaries(MergeBoundaries(result.Boundaries, evaluated[0].Boundaries), "strategy_candidate_ranked")
	return result.Normalize()
}

type strategyPlannerEvaluationInput struct {
	Entry                   StrategyCatalogEntry
	Policy                  ExecutionIntensityPolicy
	PreGate                 IntensityGateResult
	Frame                   ObjectiveFrame
	CurrentStrategyRef      DisplaySafeRef
	AvailableCapabilityRefs []DisplaySafeRef
	SatisfiedPreconditions  []MissingInput
	Attempts                []AttemptSummary
	EvidenceRefs            []EvidenceRef
	MaxAllowedIntensity     ExecutionIntensity
}

func evaluateStrategyCatalogEntry(input strategyPlannerEvaluationInput) StrategyPlanCandidate {
	entry := input.Entry.Normalize()
	candidate := entry.Candidate.Normalize()
	intensity := normalizeIntensityOr(candidate.MinIntensity, input.Frame.Intensity)
	mode := firstControlMode(candidate.ControlMode, input.Frame.ControlMode, ControlModeObjective)
	score := 1000
	result := StrategyPlanCandidate{
		ContractVersion: ContractVersion,
		SourceKind:      entry.SourceKind,
		SourceRef:       entry.SourceRef,
		Candidate:       candidate,
		Status:          VerificationSatisfied,
		Score:           score,
		FailureClass:    FailureNone,
		EvidenceRefs:    MergeEvidenceRefs(entry.EvidenceRefs, candidate.ExpectedEvidence),
		DecisionBasis: normalizeDisplaySafeRefs([]DisplaySafeRef{
			DisplaySafeRef("strategy:" + candidate.ID),
			DisplaySafeRef("source_kind:" + string(entry.SourceKind)),
			DisplaySafeRef("control_mode:" + string(mode)),
			DisplaySafeRef("min_intensity:" + string(intensity)),
		}),
		Boundaries: MergeBoundaries(
			[]Boundary{
				"strategy_candidate_metadata",
				"strategy_candidate_not_executed",
			},
			entry.Boundaries,
			candidate.Boundaries,
		),
		NextHostAction:  "run_strategy_final_gate",
		RawOutputLoaded: entry.RawOutputLoaded,
	}
	if entry.SourceKind == "" {
		return strategyPlanCandidateBlock(result, FailureConfigMissing, "host:strategy_source_kind", "provide_strategy_catalog", "strategy_source_kind_missing")
	}
	if entry.SourceRef == "" {
		return strategyPlanCandidateBlock(result, FailureConfigMissing, "host:strategy_source_ref", "provide_strategy_catalog", "strategy_source_ref_missing")
	}
	if candidate.ID == "" {
		return strategyPlanCandidateBlock(result, FailureConfigMissing, "host:strategy_candidate_ref", "provide_strategy_catalog", "strategy_candidate_ref_missing")
	}
	if _, ok := NormalizeDisplaySafeRef(candidate.ID); !ok {
		return strategyPlanCandidateBlock(result, FailureEvidenceWeak, "host:display_safe_refs", "provide_display_safe_refs", "raw_output_not_allowed")
	}
	if candidate.ControlMode == "" {
		result.Candidate.ControlMode = mode
	}
	if candidate.MinIntensity == "" {
		result.Candidate.MinIntensity = intensity
	}
	if entry.Status != VerificationSatisfied {
		return strategyPlanCandidateBlock(result, firstFailureClass(entry.FailureClass, FailureCapabilityMissing), "host:strategy_catalog_entry", "provide_strategy_catalog", "strategy_catalog_entry_not_available")
	}
	if executionIntensityRank(intensity) > executionIntensityRank(input.MaxAllowedIntensity) {
		return strategyPlanCandidateBlock(result, FailurePolicyBlocked, "contract:strategy_intensity", "request_intensity_upgrade_confirmation", "strategy_intensity_exceeds_pre_gate")
	}
	if !controlModeAllowedForIntensity(input.Policy, intensity, mode) {
		return strategyPlanCandidateBlock(result, FailurePolicyBlocked, MissingInput("contract:control_mode:"+string(mode)), "return_partial_or_request_upgrade", "control_mode_denied_by_intensity_policy")
	}
	if strategySideEffectDenied(input.Policy, intensity, candidate.SideEffectClass) {
		return strategyPlanCandidateBlock(result, FailurePolicyBlocked, MissingInput("contract:side_effect:"+candidate.SideEffectClass), "select_lower_risk_strategy", "strategy_side_effect_denied")
	}
	if l2SameStrategyBlocked(input.MaxAllowedIntensity, input.CurrentStrategyRef, candidate.ID) {
		return strategyPlanCandidateBlock(result, FailurePolicyBlocked, "contract:l2_same_strategy", "request_intensity_upgrade_confirmation", "l2_cross_strategy_blocked")
	}
	if strategyAttemptNoProgressBlocked(candidate.ID, input.Attempts) {
		return strategyPlanCandidateBlock(result, FailureRepeatedNoProgress, "host:new_evidence_or_strategy", "select_alternate_strategy", "strategy_repeated_no_progress_dedupe")
	}
	for _, missing := range missingStrategyCapabilities(candidate, input.AvailableCapabilityRefs) {
		result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
		result.Boundaries = AppendBoundaries(result.Boundaries, "strategy_capability_not_proven_available")
	}
	for _, missing := range missingStrategyPreconditions(candidate, input.SatisfiedPreconditions) {
		result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
		result.Boundaries = AppendBoundaries(result.Boundaries, "strategy_precondition_not_satisfied")
	}
	if len(result.MissingInputs) > 0 {
		result.Status = VerificationBlocked
		result.FailureClass = firstFailureClass(result.FailureClass, FailureCapabilityMissing)
		result.NextHostAction = "resolve_strategy_preconditions"
		return result.Normalize()
	}
	score += strategySourceScore(entry.SourceKind)
	score += evidenceCoverageScore(candidate.ExpectedEvidence, input.Frame.RequiredEvidence)
	score += evidenceStrengthScore(candidate.ExpectedEvidence)
	score += currentStrategyScore(candidate.ID, input.CurrentStrategyRef)
	score -= strategyAttemptPenalty(candidate.ID, input.Attempts)
	if candidate.RequiresApproval {
		score -= 25
		result.Boundaries = AppendBoundaries(result.Boundaries, "strategy_requires_approval")
	}
	if candidate.Risk != "" {
		score -= strategyRiskPenalty(candidate.Risk)
		result.DecisionBasis = appendUniqueDisplaySafeRef(result.DecisionBasis, DisplaySafeRef("risk:"+candidate.Risk))
	}
	if hasWeakExpectedEvidence(candidate.ExpectedEvidence) {
		score -= 30
		result.Status = VerificationReviewRequired
		result.FailureClass = FailureEvidenceWeak
		result.MissingInputs = AppendMissingInputs(result.MissingInputs, "host:stronger_strategy_evidence")
		result.Boundaries = AppendBoundaries(result.Boundaries, "strategy_expected_evidence_weak")
		result.NextHostAction = "review_strategy_evidence"
	}
	result.Score = score
	return result.Normalize()
}

func strategyPlannerBlock(result StrategyPlannerResult, failure FailureClass, reason string, missing MissingInput, next NextHostAction, boundary Boundary) StrategyPlannerResult {
	result.Status = VerificationBlocked
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.DecisionBasis = appendUniqueDisplaySafeRef(result.DecisionBasis, DisplaySafeRef("blocked:"+normalizeControlToken(reason)))
	result.Boundaries = AppendBoundaries(result.Boundaries, boundary)
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	return result.Normalize()
}

func strategyPlanCandidateBlock(result StrategyPlanCandidate, failure FailureClass, missing MissingInput, next NextHostAction, boundary Boundary) StrategyPlanCandidate {
	result.Status = VerificationBlocked
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.Boundaries = AppendBoundaries(result.Boundaries, boundary)
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	result.Score -= 500
	return result.Normalize()
}

func missingStrategyCapabilities(candidate StrategyCandidate, available []DisplaySafeRef) []MissingInput {
	out := []MissingInput{}
	for _, ref := range normalizeDisplaySafeRefs(candidate.CapabilityRefs) {
		if displaySafeRefSliceContains(available, ref) {
			continue
		}
		out = AppendMissingInputs(out, MissingInput("host:available_"+string(ref)))
	}
	return out
}

func missingStrategyPreconditions(candidate StrategyCandidate, satisfied []MissingInput) []MissingInput {
	out := []MissingInput{}
	for _, required := range normalizeMissingInputs(candidate.Preconditions) {
		if missingInputSliceContains(satisfied, required) {
			continue
		}
		out = AppendMissingInputs(out, required)
	}
	return out
}

func missingInputSliceContains(values []MissingInput, needle MissingInput) bool {
	needle = firstMissingInput(needle, "")
	if needle == "" {
		return false
	}
	for _, value := range normalizeMissingInputs(values) {
		if value == needle {
			return true
		}
	}
	return false
}

func l2SameStrategyBlocked(maxAllowed ExecutionIntensity, current DisplaySafeRef, candidateID string) bool {
	if executionIntensityRank(maxAllowed) > executionIntensityRank(IntensityL2BoundedToolLoop) {
		return false
	}
	if current == "" {
		return false
	}
	ref, ok := NormalizeDisplaySafeRef(candidateID)
	if !ok {
		return false
	}
	return ref != current
}

func evidenceCoverageScore(candidateEvidence []EvidenceRef, requiredEvidence []EvidenceRef) int {
	required := normalizeEvidenceRefs(requiredEvidence)
	if len(required) == 0 {
		return 0
	}
	candidate := normalizeEvidenceRefs(candidateEvidence)
	score := 0
	for _, requiredRef := range required {
		for _, candidateRef := range candidate {
			if candidateEvidenceMatchesRequired(candidateRef, requiredRef) {
				score += 25
				break
			}
		}
	}
	return score
}

func candidateEvidenceMatchesRequired(candidate EvidenceRef, required EvidenceRef) bool {
	if candidate.Ref != "" && required.Ref != "" && candidate.Ref == required.Ref {
		return true
	}
	if candidate.Kind != "" && required.Kind != "" && candidate.Kind == required.Kind {
		return true
	}
	if candidate.Source != "" && required.Source != "" && candidate.Source == required.Source {
		return true
	}
	return false
}

func evidenceStrengthScore(evidence []EvidenceRef) int {
	score := 0
	for _, ref := range normalizeEvidenceRefs(evidence) {
		switch ref.Strength {
		case EvidenceStrong:
			score += 20
		case EvidenceAdequate:
			score += 10
		case EvidenceWeak:
			score -= 10
		case EvidenceMissing:
			score -= 20
		}
	}
	return score
}

func hasWeakExpectedEvidence(evidence []EvidenceRef) bool {
	normalized := normalizeEvidenceRefs(evidence)
	if len(normalized) == 0 {
		return false
	}
	for _, ref := range normalized {
		if ref.Strength == EvidenceWeak || ref.Strength == EvidenceMissing {
			return true
		}
	}
	return false
}

func currentStrategyScore(candidateID string, current DisplaySafeRef) int {
	if current == "" {
		return 0
	}
	ref, ok := NormalizeDisplaySafeRef(candidateID)
	if !ok || ref != current {
		return 0
	}
	return 40
}

func strategyAttemptPenalty(candidateID string, attempts []AttemptSummary) int {
	ref, ok := NormalizeDisplaySafeRef(candidateID)
	if !ok {
		return 0
	}
	penalty := 0
	for _, attempt := range normalizeAttemptSummaries(attempts) {
		if attempt.StrategyID != string(ref) {
			continue
		}
		switch attempt.Status {
		case VerificationFailed, VerificationBlocked:
			penalty += 35
		case VerificationPartial, VerificationReviewRequired:
			penalty += 15
		case VerificationSatisfied:
			penalty -= 20
		}
		if attempt.FailureClass == FailureRepeatedNoProgress {
			penalty += 50
		}
	}
	return penalty
}

func strategyAttemptNoProgressBlocked(candidateID string, attempts []AttemptSummary) bool {
	ref, ok := NormalizeDisplaySafeRef(candidateID)
	if !ok {
		return false
	}
	count := 0
	for _, attempt := range normalizeAttemptSummaries(attempts) {
		if attempt.StrategyID != string(ref) {
			continue
		}
		if attempt.Status == VerificationSatisfied {
			continue
		}
		if len(attempt.EvidenceRefs) > 0 || attempt.ObservationCount > 0 {
			continue
		}
		if attempt.FailureClass == FailureRepeatedNoProgress {
			return true
		}
		count++
		if count >= 2 {
			return true
		}
	}
	return false
}

func strategyRiskPenalty(risk string) int {
	switch normalizeControlToken(risk) {
	case "low", "read_only", "safe":
		return 0
	case "medium", "requires_review":
		return 15
	case "high", "side_effect":
		return 40
	default:
		return 10
	}
}

func strategySourceScore(kind StrategyCatalogSourceKind) int {
	switch NormalizeStrategyCatalogSourceKind(string(kind)) {
	case StrategyCatalogSourceHostAdapter:
		return 30
	case StrategyCatalogSourceScene:
		return 25
	case StrategyCatalogSourceWorkflow:
		return 20
	case StrategyCatalogSourceOperations:
		return 18
	case StrategyCatalogSourceTool:
		return 15
	case StrategyCatalogSourceSkill:
		return 12
	case StrategyCatalogSourceCapability:
		return 5
	case StrategyCatalogSourceDelegation:
		return 4
	default:
		return 0
	}
}

func sortStrategyPlanCandidates(values []StrategyPlanCandidate) {
	sort.SliceStable(values, func(i, j int) bool {
		left := values[i].Normalize()
		right := values[j].Normalize()
		if left.Status != right.Status {
			return strategyPlanStatusRank(left.Status) < strategyPlanStatusRank(right.Status)
		}
		if left.Score != right.Score {
			return left.Score > right.Score
		}
		if left.SourceKind != right.SourceKind {
			return strategyCatalogSourceKindRank(left.SourceKind) < strategyCatalogSourceKindRank(right.SourceKind)
		}
		if left.Candidate.ID != right.Candidate.ID {
			return left.Candidate.ID < right.Candidate.ID
		}
		return left.SourceRef < right.SourceRef
	})
}

func strategyPlanStatusRank(status VerificationStatus) int {
	switch NormalizeVerificationStatus(string(status)) {
	case VerificationSatisfied:
		return 0
	case VerificationReviewRequired:
		return 1
	case VerificationPartial:
		return 2
	case VerificationBlocked:
		return 3
	case VerificationFailed:
		return 4
	default:
		return 5
	}
}

func strategyCatalogSourceKindRank(kind StrategyCatalogSourceKind) int {
	switch NormalizeStrategyCatalogSourceKind(string(kind)) {
	case StrategyCatalogSourceHostAdapter:
		return 0
	case StrategyCatalogSourceScene:
		return 1
	case StrategyCatalogSourceWorkflow:
		return 2
	case StrategyCatalogSourceOperations:
		return 3
	case StrategyCatalogSourceTool:
		return 4
	case StrategyCatalogSourceSkill:
		return 5
	case StrategyCatalogSourceCapability:
		return 6
	case StrategyCatalogSourceDelegation:
		return 7
	default:
		return 8
	}
}

func firstStrategyPlanFailureClass(candidates []StrategyPlanCandidate, fallback FailureClass) FailureClass {
	for _, candidate := range normalizeStrategyPlanCandidates(candidates) {
		if candidate.FailureClass != FailureNone {
			return candidate.FailureClass
		}
	}
	return NormalizeFailureClass(string(fallback))
}

func firstStrategyPlanMissingInput(candidates []StrategyPlanCandidate, fallback MissingInput) MissingInput {
	for _, candidate := range normalizeStrategyPlanCandidates(candidates) {
		if len(candidate.MissingInputs) > 0 {
			return candidate.MissingInputs[0]
		}
	}
	return firstMissingInput(fallback, "host:strategy_candidate")
}

func cloneStrategyCatalogEntries(in []StrategyCatalogEntry) []StrategyCatalogEntry {
	if len(in) == 0 {
		return nil
	}
	out := make([]StrategyCatalogEntry, len(in))
	for i, value := range in {
		out[i] = value.Clone()
	}
	return out
}

func normalizeStrategyCatalogEntries(in []StrategyCatalogEntry) []StrategyCatalogEntry {
	out := make([]StrategyCatalogEntry, 0, len(in))
	seen := map[string]struct{}{}
	for _, value := range in {
		normalized := value.Normalize()
		if normalized.Candidate.ID == "" {
			continue
		}
		key := string(normalized.SourceKind) + "|" + string(normalized.SourceRef) + "|" + normalized.Candidate.ID
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, normalized)
	}
	return out
}

func cloneStrategyPlanCandidates(in []StrategyPlanCandidate) []StrategyPlanCandidate {
	if len(in) == 0 {
		return nil
	}
	out := make([]StrategyPlanCandidate, len(in))
	for i, value := range in {
		out[i] = value.Clone()
	}
	return out
}

func normalizeStrategyPlanCandidates(in []StrategyPlanCandidate) []StrategyPlanCandidate {
	out := make([]StrategyPlanCandidate, 0, len(in))
	seen := map[string]struct{}{}
	for _, value := range in {
		normalized := value.Normalize()
		if normalized.Candidate.ID == "" {
			continue
		}
		key := string(normalized.SourceKind) + "|" + string(normalized.SourceRef) + "|" + normalized.Candidate.ID
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, normalized)
	}
	return out
}
