package controlcontract

// StructuredObservationNormalizationInput carries already structured host
// observations into the portable control-plane observation contract.
type StructuredObservationNormalizationInput struct {
	Frame                    ObjectiveFrame `json:"frame,omitempty"`
	SourceKind               string         `json:"source_kind,omitempty"`
	SourceRef                DisplaySafeRef `json:"source_ref,omitempty"`
	Observations             []Observation  `json:"observations,omitempty"`
	EvidenceRefs             []EvidenceRef  `json:"evidence_refs,omitempty"`
	ExpectedObservationKinds []string       `json:"expected_observation_kinds,omitempty"`
	Boundaries               []Boundary     `json:"boundaries,omitempty"`
	RawOutputLoaded          bool           `json:"raw_output_loaded"`
}

// BuildStructuredObservationNormalization normalizes structured observations
// without translating or executing a concrete host adapter.
func BuildStructuredObservationNormalization(input StructuredObservationNormalizationInput) ObservationNormalizationResult {
	frame := input.Frame.Normalize()
	sourceKind := normalizeControlToken(input.SourceKind)
	sourceRef := normalizeOneDisplaySafeRef(input.SourceRef)
	result := ObservationNormalizationResult{
		ContractVersion: ContractVersion,
		Projected:       true,
		Status:          VerificationBlocked,
		Frame:           frame,
		SourceKind:      sourceKind,
		SourceRef:       sourceRef,
		EvidenceRefs:    MergeEvidenceRefs(input.EvidenceRefs),
		FailureClass:    FailureNone,
		Boundaries: MergeBoundaries(
			[]Boundary{
				"observation_normalization",
				"normalized_observation_contract",
				"projection_only",
				"no_runner_dispatch",
				"no_runtime_adapter_execution",
				"no_transcript_fact_inference",
			},
			input.Boundaries,
		),
		NextHostAction:  "normalize_observations",
		RunnerEffect:    "none",
		PromptEffect:    "none",
		RawOutputLoaded: input.RawOutputLoaded,
	}
	if structuredObservationNormalizationInputUnsafe(input) {
		return structuredObservationNormalizationBlock(result, VerificationReviewRequired, FailureEvidenceWeak, "host:display_safe_refs", "provide_display_safe_refs", "raw_output_not_allowed")
	}
	result.Observations = structuredObservationNormalizationObservations(frame, sourceRef, input.Observations)
	result.ObservationKinds = structuredObservationKinds(result.Observations)
	result.EvidenceRefs = MergeEvidenceRefs(result.EvidenceRefs, structuredObservationEvidenceRefs(result.Observations))
	expectedKinds := normalizeControlTokenList(input.ExpectedObservationKinds)
	result.MissingObservationKinds = structuredMissingObservationKinds(expectedKinds, result.ObservationKinds)
	switch {
	case len(result.Observations) == 0:
		return structuredObservationNormalizationBlock(result, VerificationBlocked, FailureEvidenceMissing, "host:normalized_observation", "provide_runtime_adapter_observations", "normalized_observation_missing")
	case len(result.MissingObservationKinds) > 0:
		for _, kind := range result.MissingObservationKinds {
			result.MissingInputs = AppendMissingInputs(result.MissingInputs, MissingInput("observation_kind:"+kind))
		}
		return structuredObservationNormalizationBlock(result, VerificationBlocked, FailureEvidenceMissing, "host:expected_observation_kind", "provide_runtime_adapter_observations", "expected_observation_kind_missing")
	case structuredObservationsWeak(result.Observations):
		return structuredObservationNormalizationBlock(result, VerificationReviewRequired, FailureEvidenceWeak, "host:stronger_observation_evidence", "review_normalized_observation_evidence", "normalized_observation_evidence_weak")
	case result.RawOutputLoaded:
		return structuredObservationNormalizationBlock(result, VerificationReviewRequired, FailureEvidenceWeak, "host:display_safe_refs", "provide_display_safe_refs", "raw_output_not_allowed")
	default:
		result.Status = VerificationSatisfied
		result.ReadyForVerification = true
		result.NextHostAction = "run_objective_verification_gate"
		result.Boundaries = AppendBoundaries(result.Boundaries, "normalized_observations_ready", "ready_for_objective_verification_gate")
	}
	return result.Normalize()
}

func structuredObservationNormalizationBlock(result ObservationNormalizationResult, status VerificationStatus, failure FailureClass, missing MissingInput, next NextHostAction, boundary Boundary) ObservationNormalizationResult {
	result.Status = NormalizeVerificationStatus(string(status))
	if result.Status == VerificationSatisfied || result.Status == VerificationNotEvaluated {
		result.Status = VerificationBlocked
	}
	result.ReadyForVerification = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.Boundaries = AppendBoundaries(result.Boundaries, boundary)
	result.NextHostAction = firstNextHostAction(next, result.NextHostAction)
	return result.Normalize()
}

func structuredObservationNormalizationInputUnsafe(input StructuredObservationNormalizationInput) bool {
	return displaySafeRefRejected(input.SourceRef) ||
		evidenceRefRejected(input.EvidenceRefs) ||
		structuredObservationSliceUnsafe(input.Observations) ||
		input.RawOutputLoaded
}

func structuredObservationSliceUnsafe(values []Observation) bool {
	for _, value := range values {
		if value.RawOutputLoaded ||
			displaySafeRefRejected(value.Source) ||
			displaySafeRefRejected(value.Subject) ||
			displaySafeRefSliceRejected(value.DisplaySafeRefs) ||
			evidenceRefRejected(value.EvidenceRefs) ||
			ContainsUnsafeRawOutput(value.Name, value.Value, value.Unit, value.ObservedAt, value.DegradationReason) {
			return true
		}
	}
	return false
}

func structuredObservationNormalizationObservations(frame ObjectiveFrame, sourceRef DisplaySafeRef, values []Observation) []Observation {
	subjectRef := normalizeOneDisplaySafeRef(DisplaySafeRef(frame.ID))
	out := make([]Observation, 0, len(values))
	for _, value := range values {
		normalized := value.Normalize()
		if normalized.Source == "" {
			normalized.Source = sourceRef
		}
		if normalized.Subject == "" {
			normalized.Subject = subjectRef
		}
		normalized = normalized.Normalize()
		if normalized.Kind == "" && len(normalized.EvidenceRefs) == 0 && len(normalized.DisplaySafeRefs) == 0 {
			continue
		}
		out = append(out, normalized)
	}
	return normalizeObservations(out)
}

func structuredObservationKinds(values []Observation) []string {
	out := []string{}
	for _, observation := range normalizeObservations(values) {
		if observation.Kind != "" {
			out = append(out, observation.Kind)
		}
	}
	return normalizeControlTokenList(out)
}

func structuredMissingObservationKinds(expected []string, actual []string) []string {
	actualSet := map[string]struct{}{}
	for _, kind := range normalizeControlTokenList(actual) {
		actualSet[kind] = struct{}{}
	}
	missing := []string{}
	for _, kind := range normalizeControlTokenList(expected) {
		if _, ok := actualSet[kind]; !ok {
			missing = append(missing, kind)
		}
	}
	return normalizeControlTokenList(missing)
}

func structuredObservationsWeak(values []Observation) bool {
	for _, observation := range normalizeObservations(values) {
		if observation.Strength == EvidenceWeak || observation.Strength == EvidenceMissing {
			return true
		}
		if len(observation.EvidenceRefs) == 0 && len(observation.DisplaySafeRefs) == 0 {
			return true
		}
	}
	return false
}

func structuredObservationEvidenceRefs(values []Observation) []EvidenceRef {
	out := []EvidenceRef{}
	for _, observation := range normalizeObservations(values) {
		out = append(out, observation.EvidenceRefs...)
	}
	return normalizeEvidenceRefs(out)
}
