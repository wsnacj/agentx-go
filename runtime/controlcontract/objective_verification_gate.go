package controlcontract

type ObjectiveVerificationGateInput struct {
	Frame            ObjectiveFrame                 `json:"frame,omitempty"`
	Normalization    ObservationNormalizationResult `json:"normalization,omitempty"`
	Observations     []Observation                  `json:"observations,omitempty"`
	EvidenceRefs     []EvidenceRef                  `json:"evidence_refs,omitempty"`
	RequiredEvidence []EvidenceRef                  `json:"required_evidence,omitempty"`
	Boundaries       []Boundary                     `json:"boundaries,omitempty"`
	RawOutputLoaded  bool                           `json:"raw_output_loaded"`
}

type ObjectiveEvidenceRequirement struct {
	ContractVersion         string             `json:"contract_version,omitempty"`
	RequirementRef          DisplaySafeRef     `json:"requirement_ref,omitempty"`
	Kind                    string             `json:"kind,omitempty"`
	Source                  DisplaySafeRef     `json:"source,omitempty"`
	MinStrength             EvidenceStrength   `json:"min_strength,omitempty"`
	Status                  VerificationStatus `json:"status,omitempty"`
	Satisfied               bool               `json:"satisfied"`
	MatchedEvidenceRefs     []EvidenceRef      `json:"matched_evidence_refs,omitempty"`
	MatchedObservationKinds []string           `json:"matched_observation_kinds,omitempty"`
	FailureClass            FailureClass       `json:"failure_class,omitempty"`
	MissingInputs           []MissingInput     `json:"missing_inputs,omitempty"`
	Boundaries              []Boundary         `json:"boundaries,omitempty"`
}

type ObjectiveVerificationGateResult struct {
	ContractVersion string                         `json:"contract_version,omitempty"`
	Projected       bool                           `json:"projected"`
	Status          VerificationStatus             `json:"status,omitempty"`
	Satisfied       bool                           `json:"satisfied"`
	Frame           ObjectiveFrame                 `json:"frame,omitempty"`
	Normalization   ObservationNormalizationResult `json:"normalization,omitempty"`
	Requirements    []ObjectiveEvidenceRequirement `json:"requirements,omitempty"`
	Observations    []Observation                  `json:"observations,omitempty"`
	EvidenceRefs    []EvidenceRef                  `json:"evidence_refs,omitempty"`
	Verification    VerificationResult             `json:"verification,omitempty"`
	FailureClass    FailureClass                   `json:"failure_class,omitempty"`
	FailureReason   string                         `json:"failure_reason,omitempty"`
	MissingInputs   []MissingInput                 `json:"missing_inputs,omitempty"`
	Boundaries      []Boundary                     `json:"boundaries,omitempty"`
	NextHostAction  NextHostAction                 `json:"next_host_action,omitempty"`
	RunnerEffect    string                         `json:"runner_effect,omitempty"`
	PromptEffect    string                         `json:"prompt_effect,omitempty"`
	RawOutputLoaded bool                           `json:"raw_output_loaded"`
}

func BuildObjectiveVerificationGate(input ObjectiveVerificationGateInput) ObjectiveVerificationGateResult {
	frame := input.Frame.Normalize()
	normalization := input.Normalization.Normalize()
	if frame.ID == "" {
		frame = normalization.Frame.Normalize()
	}
	observations := normalizeObservations(append(cloneObservations(normalization.Observations), input.Observations...))
	evidenceRefs := MergeEvidenceRefs(input.EvidenceRefs, normalization.EvidenceRefs, runtimeAdapterObservationEvidenceRefs(observations))
	requiredEvidence := normalizeEvidenceRefs(append(cloneEvidenceRefs(frame.RequiredEvidence), input.RequiredEvidence...))
	result := ObjectiveVerificationGateResult{
		ContractVersion: ContractVersion,
		Projected:       true,
		Status:          VerificationBlocked,
		Frame:           frame,
		Normalization:   normalization,
		Observations:    observations,
		EvidenceRefs:    evidenceRefs,
		FailureClass:    FailureNone,
		Boundaries: MergeBoundaries(
			[]Boundary{
				"objective_verification_gate",
				"requirement_by_requirement_verification",
				"normalized_observations_only",
				"projection_only",
				"no_runner_dispatch",
				"no_transcript_fact_inference",
			},
			input.Boundaries,
			normalization.Boundaries,
		),
		NextHostAction:  "run_objective_verification_gate",
		RunnerEffect:    "none",
		PromptEffect:    "none",
		RawOutputLoaded: input.RawOutputLoaded || normalization.RawOutputLoaded,
	}
	if objectiveVerificationGateInputUnsafe(input) {
		return objectiveVerificationGateBlock(result, VerificationReviewRequired, FailureEvidenceWeak, "host:display_safe_refs", "provide_display_safe_refs", "raw_output_not_allowed")
	}
	if !normalization.ReadyForVerification {
		return objectiveVerificationGateBlock(
			result,
			normalization.Status,
			firstFailureClass(normalization.FailureClass, FailureEvidenceMissing),
			"control_plane:normalized_observations",
			firstNextHostAction(normalization.NextHostAction, "normalize_observations"),
			"normalized_observations_not_ready",
		)
	}
	if len(requiredEvidence) == 0 {
		if len(frame.SuccessCriteria) > 0 {
			return objectiveVerificationGateBlock(result, VerificationBlocked, FailureEvidenceMissing, "host:required_evidence_contract", "provide_required_evidence_contract", "success_criteria_requires_required_evidence")
		}
		return objectiveVerificationGateBlock(result, VerificationBlocked, FailureEvidenceMissing, "host:success_criteria", "provide_objective_contract", "objective_success_criteria_missing")
	}

	for _, requirement := range requiredEvidence {
		checked := evaluateObjectiveEvidenceRequirement(requirement, observations, evidenceRefs)
		result.Requirements = append(result.Requirements, checked)
		result.EvidenceRefs = MergeEvidenceRefs(result.EvidenceRefs, checked.MatchedEvidenceRefs)
		result.MissingInputs = AppendMissingInputs(result.MissingInputs, checked.MissingInputs...)
		result.Boundaries = AppendBoundaries(result.Boundaries, checked.Boundaries...)
	}
	result = objectiveVerificationGateAggregate(result)
	if objectiveVerificationContainsProposalOnlyObservation(observations) && !result.Satisfied {
		result.Boundaries = AppendBoundaries(result.Boundaries, "proposal_only_not_objective_completion")
	}
	return result.Normalize()
}

func CloneObjectiveEvidenceRequirement(in ObjectiveEvidenceRequirement) ObjectiveEvidenceRequirement {
	out := in
	out.MatchedEvidenceRefs = cloneEvidenceRefs(in.MatchedEvidenceRefs)
	out.MatchedObservationKinds = cloneStringSlice(in.MatchedObservationKinds)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (r ObjectiveEvidenceRequirement) Clone() ObjectiveEvidenceRequirement {
	return CloneObjectiveEvidenceRequirement(r)
}

func (r ObjectiveEvidenceRequirement) Normalize() ObjectiveEvidenceRequirement {
	out := CloneObjectiveEvidenceRequirement(r)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.RequirementRef = normalizeOneDisplaySafeRef(out.RequirementRef)
	out.Kind = normalizeControlToken(out.Kind)
	out.Source = normalizeOneDisplaySafeRef(out.Source)
	out.MinStrength = objectiveRequirementMinStrength(out.MinStrength)
	out.Status = NormalizeVerificationStatus(string(out.Status))
	out.MatchedEvidenceRefs = normalizeEvidenceRefs(out.MatchedEvidenceRefs)
	out.MatchedObservationKinds = normalizeControlTokenList(out.MatchedObservationKinds)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	if out.Status == VerificationNotEvaluated {
		out.Status = VerificationBlocked
	}
	out.Satisfied = out.Status == VerificationSatisfied && len(out.MissingInputs) == 0
	return out
}

func CloneObjectiveVerificationGateResult(in ObjectiveVerificationGateResult) ObjectiveVerificationGateResult {
	out := in
	out.Frame = in.Frame.Clone()
	out.Normalization = in.Normalization.Clone()
	out.Requirements = cloneObjectiveEvidenceRequirements(in.Requirements)
	out.Observations = cloneObservations(in.Observations)
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.Verification = in.Verification.Clone()
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (r ObjectiveVerificationGateResult) Clone() ObjectiveVerificationGateResult {
	return CloneObjectiveVerificationGateResult(r)
}

func (r ObjectiveVerificationGateResult) Normalize() ObjectiveVerificationGateResult {
	out := CloneObjectiveVerificationGateResult(r)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeVerificationStatus(string(out.Status))
	out.Frame = out.Frame.Normalize()
	out.Normalization = out.Normalization.Normalize()
	out.Requirements = normalizeObjectiveEvidenceRequirements(out.Requirements)
	out.Observations = normalizeObservations(out.Observations)
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
	out.Verification = out.Verification.Normalize()
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.FailureReason = managedObjectiveReplannerSafeReason(out.FailureReason)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
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
		out.Satisfied = false
		if out.FailureClass == FailureNone {
			out.FailureClass = FailureEvidenceWeak
		}
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		if out.NextHostAction == "" {
			out.NextHostAction = "provide_display_safe_refs"
		}
	}
	out.Satisfied = out.Status == VerificationSatisfied && len(out.MissingInputs) == 0 && !out.RawOutputLoaded
	out.Verification = VerificationResult{
		ContractVersion: ContractVersion,
		Status:          out.Status,
		Satisfied:       out.Satisfied,
		FailureClass:    out.FailureClass,
		FailureReason:   out.FailureReason,
		EvidenceRefs:    out.EvidenceRefs,
		MissingInputs:   out.MissingInputs,
		Boundaries:      out.Boundaries,
		Findings:        objectiveVerificationFindings(out.Requirements),
		NextHostAction:  out.NextHostAction,
		RawOutputLoaded: out.RawOutputLoaded,
	}.Normalize()
	return out
}

func objectiveVerificationGateAggregate(result ObjectiveVerificationGateResult) ObjectiveVerificationGateResult {
	total := len(result.Requirements)
	satisfied := 0
	hasWeak := false
	hasMissing := false
	for _, requirement := range normalizeObjectiveEvidenceRequirements(result.Requirements) {
		switch requirement.Status {
		case VerificationSatisfied:
			satisfied++
		case VerificationReviewRequired:
			hasWeak = true
		default:
			hasMissing = true
		}
		result.FailureClass = firstFailureClass(result.FailureClass, requirement.FailureClass)
	}
	switch {
	case total > 0 && satisfied == total:
		result.Status = VerificationSatisfied
		result.FailureClass = FailureNone
		result.NextHostAction = "update_objective_controller"
		result.Boundaries = AppendBoundaries(result.Boundaries, "objective_requirements_satisfied", "ready_for_objective_controller_update")
	case hasMissing:
		result.Status = VerificationPartial
		result.FailureClass = firstFailureClass(result.FailureClass, FailureEvidenceMissing)
		result.NextHostAction = "request_replan_or_return_partial"
		result.Boundaries = AppendBoundaries(result.Boundaries, "objective_requirements_evidence_missing")
	case hasWeak:
		result.Status = VerificationPartial
		result.FailureClass = firstFailureClass(result.FailureClass, FailureEvidenceWeak)
		result.NextHostAction = "request_replan_or_stronger_evidence"
		result.Boundaries = AppendBoundaries(result.Boundaries, "objective_requirements_evidence_weak")
	default:
		result.Status = VerificationBlocked
		result.FailureClass = firstFailureClass(result.FailureClass, FailureEvidenceMissing)
		result.NextHostAction = "provide_required_evidence_contract"
	}
	return result
}

func objectiveVerificationGateBlock(result ObjectiveVerificationGateResult, status VerificationStatus, failure FailureClass, missing MissingInput, next NextHostAction, boundary Boundary) ObjectiveVerificationGateResult {
	result.Status = NormalizeVerificationStatus(string(status))
	if result.Status == VerificationSatisfied || result.Status == VerificationNotEvaluated {
		result.Status = VerificationBlocked
	}
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.Boundaries = AppendBoundaries(result.Boundaries, boundary)
	result.NextHostAction = firstNextHostAction(next, result.NextHostAction)
	return result.Normalize()
}

func evaluateObjectiveEvidenceRequirement(required EvidenceRef, observations []Observation, evidenceRefs []EvidenceRef) ObjectiveEvidenceRequirement {
	required = required.Normalize()
	requirement := ObjectiveEvidenceRequirement{
		ContractVersion: ContractVersion,
		RequirementRef:  required.Ref,
		Kind:            required.Kind,
		Source:          required.Source,
		MinStrength:     objectiveRequirementMinStrength(required.Strength),
		Status:          VerificationBlocked,
		FailureClass:    FailureEvidenceMissing,
		Boundaries:      []Boundary{"objective_evidence_requirement"},
	}
	matches := objectiveRequirementMatches(required, observations, evidenceRefs)
	requirement.MatchedEvidenceRefs = matches.evidenceRefs
	requirement.MatchedObservationKinds = matches.observationKinds
	switch {
	case len(matches.evidenceRefs) == 0:
		requirement.Status = VerificationBlocked
		requirement.FailureClass = FailureEvidenceMissing
		requirement.MissingInputs = objectiveRequirementMissingInputs(required)
		requirement.Boundaries = AppendBoundaries(requirement.Boundaries, "objective_requirement_evidence_missing")
	case objectiveEvidenceStrengthAtLeast(matches.bestStrength, requirement.MinStrength):
		requirement.Status = VerificationSatisfied
		requirement.FailureClass = FailureNone
		requirement.Boundaries = AppendBoundaries(requirement.Boundaries, "objective_requirement_satisfied")
	default:
		requirement.Status = VerificationReviewRequired
		requirement.FailureClass = FailureEvidenceWeak
		requirement.MissingInputs = objectiveRequirementMissingInputs(required)
		requirement.Boundaries = AppendBoundaries(requirement.Boundaries, "objective_requirement_evidence_weak")
	}
	return requirement.Normalize()
}

type objectiveRequirementMatchSet struct {
	evidenceRefs     []EvidenceRef
	observationKinds []string
	bestStrength     EvidenceStrength
}

func objectiveRequirementMatches(required EvidenceRef, observations []Observation, evidenceRefs []EvidenceRef) objectiveRequirementMatchSet {
	result := objectiveRequirementMatchSet{bestStrength: EvidenceMissing}
	for _, evidence := range normalizeEvidenceRefs(evidenceRefs) {
		if objectiveEvidenceRefMatchesRequirement(required, evidence, "") {
			result.evidenceRefs = MergeEvidenceRefs(result.evidenceRefs, []EvidenceRef{evidence})
			result.bestStrength = strongerEvidenceStrength(result.bestStrength, evidence.Strength)
		}
	}
	for _, observation := range normalizeObservations(observations) {
		observationMatched := false
		if objectiveObservationMatchesRequirement(required, observation) {
			observationMatched = true
			result.bestStrength = strongerEvidenceStrength(result.bestStrength, observation.Strength)
		}
		for _, evidence := range normalizeEvidenceRefs(observation.EvidenceRefs) {
			if objectiveEvidenceRefMatchesRequirement(required, evidence, observation.Kind) {
				observationMatched = true
				result.evidenceRefs = MergeEvidenceRefs(result.evidenceRefs, []EvidenceRef{evidence})
				result.bestStrength = strongerEvidenceStrength(result.bestStrength, evidence.Strength)
			}
		}
		if observationMatched && observation.Kind != "" {
			result.observationKinds = append(result.observationKinds, observation.Kind)
		}
	}
	result.observationKinds = normalizeControlTokenList(result.observationKinds)
	return result
}

func objectiveObservationMatchesRequirement(required EvidenceRef, observation Observation) bool {
	required = required.Normalize()
	observation = observation.Normalize()
	if required.Ref != "" && !displaySafeRefSliceContains(observation.DisplaySafeRefs, required.Ref) {
		return false
	}
	if required.Kind != "" && observation.Kind != required.Kind {
		return false
	}
	if required.Source != "" && observation.Source != required.Source {
		return false
	}
	return required.Ref != "" || required.Kind != "" || required.Source != ""
}

func objectiveEvidenceRefMatchesRequirement(required EvidenceRef, evidence EvidenceRef, observationKind string) bool {
	required = required.Normalize()
	evidence = evidence.Normalize()
	if required.Ref != "" && evidence.Ref != required.Ref {
		return false
	}
	if required.Kind != "" && evidence.Kind != required.Kind && normalizeControlToken(observationKind) != required.Kind {
		return false
	}
	if required.Source != "" && evidence.Source != required.Source {
		return false
	}
	return required.Ref != "" || required.Kind != "" || required.Source != ""
}

func objectiveRequirementMissingInputs(required EvidenceRef) []MissingInput {
	required = required.Normalize()
	if required.Ref != "" {
		return []MissingInput{MissingInput(required.Ref)}
	}
	if required.Kind != "" {
		return []MissingInput{MissingInput("evidence_kind:" + required.Kind)}
	}
	return []MissingInput{"host:required_evidence"}
}

func objectiveRequirementMinStrength(value EvidenceStrength) EvidenceStrength {
	normalized := NormalizeEvidenceStrength(string(value))
	if normalized == EvidenceMissing {
		return EvidenceAdequate
	}
	return normalized
}

func objectiveEvidenceStrengthAtLeast(actual EvidenceStrength, required EvidenceStrength) bool {
	return objectiveEvidenceStrengthRank(actual) >= objectiveEvidenceStrengthRank(objectiveRequirementMinStrength(required))
}

func strongerEvidenceStrength(a EvidenceStrength, b EvidenceStrength) EvidenceStrength {
	if objectiveEvidenceStrengthRank(b) > objectiveEvidenceStrengthRank(a) {
		return NormalizeEvidenceStrength(string(b))
	}
	return NormalizeEvidenceStrength(string(a))
}

func objectiveEvidenceStrengthRank(value EvidenceStrength) int {
	switch NormalizeEvidenceStrength(string(value)) {
	case EvidenceStrong:
		return 3
	case EvidenceAdequate:
		return 2
	case EvidenceWeak:
		return 1
	default:
		return 0
	}
}

func objectiveVerificationContainsProposalOnlyObservation(values []Observation) bool {
	for _, observation := range normalizeObservations(values) {
		switch observation.Kind {
		case "proposal", "capability_gap":
			return true
		}
	}
	return false
}

func objectiveVerificationGateInputUnsafe(input ObjectiveVerificationGateInput) bool {
	return evidenceRefRejected(input.EvidenceRefs) ||
		evidenceRefRejected(input.RequiredEvidence) ||
		objectiveVerificationNormalizationUnsafe(input.Normalization) ||
		observationSliceUnsafePayload(input.Observations) ||
		input.RawOutputLoaded
}

func objectiveVerificationNormalizationUnsafe(normalization ObservationNormalizationResult) bool {
	return displaySafeRefRejected(normalization.SourceRef) ||
		displaySafeRefRejected(normalization.RuntimeAdapterRunRef) ||
		displaySafeRefRejected(normalization.AdapterRef) ||
		displaySafeRefRejected(normalization.StrategyRef) ||
		displaySafeRefSliceRejected(normalization.OutputRefs) ||
		evidenceRefRejected(normalization.EvidenceRefs) ||
		observationSliceUnsafePayload(normalization.Observations) ||
		normalization.RawOutputLoaded
}

func cloneObjectiveEvidenceRequirements(in []ObjectiveEvidenceRequirement) []ObjectiveEvidenceRequirement {
	if len(in) == 0 {
		return nil
	}
	out := make([]ObjectiveEvidenceRequirement, len(in))
	for i, value := range in {
		out[i] = value.Clone()
	}
	return out
}

func normalizeObjectiveEvidenceRequirements(in []ObjectiveEvidenceRequirement) []ObjectiveEvidenceRequirement {
	out := make([]ObjectiveEvidenceRequirement, 0, len(in))
	for _, value := range in {
		normalized := value.Normalize()
		if normalized.RequirementRef == "" && normalized.Kind == "" && normalized.Source == "" && len(normalized.MatchedEvidenceRefs) == 0 {
			continue
		}
		out = append(out, normalized)
	}
	return out
}

func objectiveVerificationFindings(requirements []ObjectiveEvidenceRequirement) []string {
	findings := []string{}
	for _, requirement := range normalizeObjectiveEvidenceRequirements(requirements) {
		token := "requirement_" + string(requirement.Status)
		if requirement.Kind != "" {
			token += ":" + requirement.Kind
		}
		findings = append(findings, token)
	}
	return normalizeStringList(findings)
}
