package controlcontract

type ObjectiveRequiredEvidenceBinding struct {
	CriteriaIndex int            `json:"criteria_index,omitempty"`
	CriteriaRef   DisplaySafeRef `json:"criteria_ref,omitempty"`
	Evidence      EvidenceRef    `json:"evidence,omitempty"`
}

type ObjectiveRequiredEvidenceContractInput struct {
	Frame            ObjectiveFrame                     `json:"frame,omitempty"`
	ContractRef      DisplaySafeRef                     `json:"contract_ref,omitempty"`
	SourceRef        DisplaySafeRef                     `json:"source_ref,omitempty"`
	Bindings         []ObjectiveRequiredEvidenceBinding `json:"bindings,omitempty"`
	RequiredEvidence []EvidenceRef                      `json:"required_evidence,omitempty"`
	Boundaries       []Boundary                         `json:"boundaries,omitempty"`
	RawOutputLoaded  bool                               `json:"raw_output_loaded"`
}

type ObjectiveRequiredEvidenceContract struct {
	ContractVersion      string                             `json:"contract_version,omitempty"`
	Projected            bool                               `json:"projected"`
	Available            bool                               `json:"available"`
	Status               VerificationStatus                 `json:"status,omitempty"`
	Mode                 string                             `json:"mode,omitempty"`
	RunnerEffect         string                             `json:"runner_effect,omitempty"`
	PromptEffect         string                             `json:"prompt_effect,omitempty"`
	ContractRef          DisplaySafeRef                     `json:"contract_ref,omitempty"`
	SourceRef            DisplaySafeRef                     `json:"source_ref,omitempty"`
	Frame                ObjectiveFrame                     `json:"frame,omitempty"`
	SuccessCriteriaCount int                                `json:"success_criteria_count"`
	BindingCount         int                                `json:"binding_count"`
	RequiredEvidence     []EvidenceRef                      `json:"required_evidence,omitempty"`
	Bindings             []ObjectiveRequiredEvidenceBinding `json:"bindings,omitempty"`
	FailureClass         FailureClass                       `json:"failure_class,omitempty"`
	MissingInputs        []MissingInput                     `json:"missing_inputs,omitempty"`
	Boundaries           []Boundary                         `json:"boundaries,omitempty"`
	NextHostAction       NextHostAction                     `json:"next_host_action,omitempty"`
	RawOutputLoaded      bool                               `json:"raw_output_loaded"`
}

func BuildObjectiveRequiredEvidenceContract(input ObjectiveRequiredEvidenceContractInput) ObjectiveRequiredEvidenceContract {
	result := baseObjectiveRequiredEvidenceContract(input)
	if objectiveRequiredEvidenceContractInputUnsafe(input) {
		return objectiveRequiredEvidenceContractBlock(result, VerificationReviewRequired, FailureEvidenceWeak, "host:display_safe_refs", "provide_display_safe_refs", "raw_output_not_allowed")
	}
	if len(result.Frame.SuccessCriteria) == 0 {
		return objectiveRequiredEvidenceContractBlock(result, VerificationBlocked, FailureEvidenceMissing, "host:success_criteria", "provide_objective_contract", "objective_success_criteria_missing")
	}
	if len(input.Bindings) == 0 && len(input.RequiredEvidence) == 0 {
		return objectiveRequiredEvidenceContractBlock(result, VerificationBlocked, FailureEvidenceMissing, "host:required_evidence_contract", "provide_required_evidence_contract", "success_criteria_requires_required_evidence")
	}

	covered := map[int]struct{}{}
	for _, binding := range input.Bindings {
		normalized := normalizeObjectiveRequiredEvidenceBinding(binding)
		result.Bindings = append(result.Bindings, normalized)
		if normalized.CriteriaIndex <= 0 || normalized.CriteriaIndex > len(result.Frame.SuccessCriteria) {
			result.MissingInputs = AppendMissingInputs(result.MissingInputs, "host:required_evidence_criteria_index")
			continue
		}
		if normalized.Evidence.Ref == "" {
			result.MissingInputs = AppendMissingInputs(result.MissingInputs, "host:required_evidence_ref")
			continue
		}
		if normalized.Evidence.Kind == "" {
			result.MissingInputs = AppendMissingInputs(result.MissingInputs, "host:required_evidence_kind")
			continue
		}
		covered[normalized.CriteriaIndex] = struct{}{}
		result.RequiredEvidence = MergeEvidenceRefs(result.RequiredEvidence, []EvidenceRef{normalized.Evidence})
	}
	result.RequiredEvidence = MergeEvidenceRefs(result.RequiredEvidence, input.RequiredEvidence)
	for _, evidence := range normalizeEvidenceRefs(input.RequiredEvidence) {
		if evidence.Kind == "" {
			result.MissingInputs = AppendMissingInputs(result.MissingInputs, "host:required_evidence_kind")
		}
	}
	for idx := 1; idx <= len(result.Frame.SuccessCriteria); idx++ {
		if _, ok := covered[idx]; !ok && len(input.RequiredEvidence) == 0 {
			result.MissingInputs = AppendMissingInputs(result.MissingInputs, MissingInput("host:required_evidence_criteria_"+itoaSmall(idx)))
		}
	}
	if len(result.RequiredEvidence) == 0 {
		result.MissingInputs = AppendMissingInputs(result.MissingInputs, "host:required_evidence_contract")
	}
	if len(result.MissingInputs) > 0 {
		result.FailureClass = FailureEvidenceMissing
		result.NextHostAction = "provide_required_evidence_contract"
		result.Boundaries = AppendBoundaries(result.Boundaries, "required_evidence_contract_incomplete")
		return result.Normalize()
	}
	result.Status = VerificationSatisfied
	result.FailureClass = FailureNone
	result.NextHostAction = "run_objective_verification_gate"
	result.Frame.RequiredEvidence = MergeEvidenceRefs(result.Frame.RequiredEvidence, result.RequiredEvidence)
	result.Boundaries = AppendBoundaries(result.Boundaries, "required_evidence_contract_ready")
	return result.Normalize()
}

func baseObjectiveRequiredEvidenceContract(input ObjectiveRequiredEvidenceContractInput) ObjectiveRequiredEvidenceContract {
	frame := input.Frame.Normalize()
	return ObjectiveRequiredEvidenceContract{
		ContractVersion:      ContractVersion,
		Projected:            true,
		Available:            true,
		Status:               VerificationBlocked,
		Mode:                 "objective_required_evidence_contract",
		RunnerEffect:         "none",
		PromptEffect:         "none",
		ContractRef:          normalizeOneDisplaySafeRef(input.ContractRef),
		SourceRef:            normalizeOneDisplaySafeRef(input.SourceRef),
		Frame:                frame,
		SuccessCriteriaCount: len(frame.SuccessCriteria),
		BindingCount:         len(input.Bindings),
		RequiredEvidence:     normalizeEvidenceRefs(input.RequiredEvidence),
		Boundaries: AppendBoundaries(
			[]Boundary{
				"objective_required_evidence_contract",
				"host_scene_explicit_evidence_mapping",
				"no_success_criteria_text_inference",
				"projection_only",
				"no_runner_dispatch",
			},
			input.Boundaries...,
		),
		NextHostAction:  "provide_required_evidence_contract",
		RawOutputLoaded: input.RawOutputLoaded,
	}
}

func (c ObjectiveRequiredEvidenceContract) Normalize() ObjectiveRequiredEvidenceContract {
	out := c
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeVerificationStatus(string(out.Status))
	if out.Status == VerificationNotEvaluated {
		out.Status = VerificationBlocked
	}
	out.Mode = normalizeControlToken(out.Mode)
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	if out.RunnerEffect == "" {
		out.RunnerEffect = "none"
	}
	if out.PromptEffect == "" {
		out.PromptEffect = "none"
	}
	out.ContractRef = normalizeOneDisplaySafeRef(out.ContractRef)
	out.SourceRef = normalizeOneDisplaySafeRef(out.SourceRef)
	out.Frame = out.Frame.Normalize()
	out.SuccessCriteriaCount = len(out.Frame.SuccessCriteria)
	out.BindingCount = len(out.Bindings)
	out.RequiredEvidence = normalizeEvidenceRefs(out.RequiredEvidence)
	out.Bindings = normalizeObjectiveRequiredEvidenceBindings(out.Bindings)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	if out.RawOutputLoaded {
		out.Status = VerificationReviewRequired
		out.FailureClass = firstFailureClass(out.FailureClass, FailureEvidenceWeak)
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		if out.NextHostAction == "" {
			out.NextHostAction = "provide_display_safe_refs"
		}
	}
	return out
}

func objectiveRequiredEvidenceContractBlock(result ObjectiveRequiredEvidenceContract, status VerificationStatus, failure FailureClass, missing MissingInput, next NextHostAction, boundary Boundary) ObjectiveRequiredEvidenceContract {
	result.Status = status
	result.FailureClass = failure
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = next
	result.Boundaries = AppendBoundaries(result.Boundaries, boundary)
	return result.Normalize()
}

func normalizeObjectiveRequiredEvidenceBindings(in []ObjectiveRequiredEvidenceBinding) []ObjectiveRequiredEvidenceBinding {
	out := make([]ObjectiveRequiredEvidenceBinding, 0, len(in))
	for _, binding := range in {
		normalized := normalizeObjectiveRequiredEvidenceBinding(binding)
		if normalized.CriteriaIndex == 0 && normalized.CriteriaRef == "" && normalized.Evidence.Ref == "" {
			continue
		}
		out = append(out, normalized)
	}
	return out
}

func normalizeObjectiveRequiredEvidenceBinding(binding ObjectiveRequiredEvidenceBinding) ObjectiveRequiredEvidenceBinding {
	out := binding
	if out.CriteriaIndex < 0 {
		out.CriteriaIndex = 0
	}
	out.CriteriaRef = normalizeOneDisplaySafeRef(out.CriteriaRef)
	out.Evidence = out.Evidence.Normalize()
	return out
}

func objectiveRequiredEvidenceContractInputUnsafe(input ObjectiveRequiredEvidenceContractInput) bool {
	if input.RawOutputLoaded ||
		displaySafeRefRejected(input.ContractRef) ||
		displaySafeRefRejected(input.SourceRef) ||
		evidenceRefRejected(input.RequiredEvidence) {
		return true
	}
	for _, binding := range input.Bindings {
		if displaySafeRefRejected(binding.CriteriaRef) ||
			evidenceRefRejected([]EvidenceRef{binding.Evidence}) {
			return true
		}
	}
	return false
}

func itoaSmall(value int) string {
	if value <= 0 {
		return "0"
	}
	digits := [20]byte{}
	pos := len(digits)
	for value > 0 {
		pos--
		digits[pos] = byte('0' + value%10)
		value /= 10
	}
	return string(digits[pos:])
}
