package controlcontract

// ObservationNormalizationResult is the display-safe output consumed by the
// portable objective verification kernel. Hosts remain responsible for
// translating concrete adapter results into this contract.
type ObservationNormalizationResult struct {
	ContractVersion         string             `json:"contract_version,omitempty"`
	Projected               bool               `json:"projected"`
	Status                  VerificationStatus `json:"status,omitempty"`
	ReadyForVerification    bool               `json:"ready_for_verification"`
	Frame                   ObjectiveFrame     `json:"frame,omitempty"`
	SourceKind              string             `json:"source_kind,omitempty"`
	SourceRef               DisplaySafeRef     `json:"source_ref,omitempty"`
	RuntimeAdapterRunRef    DisplaySafeRef     `json:"runtime_adapter_run_ref,omitempty"`
	AdapterRef              DisplaySafeRef     `json:"adapter_ref,omitempty"`
	StrategyRef             DisplaySafeRef     `json:"strategy_ref,omitempty"`
	Observations            []Observation      `json:"observations,omitempty"`
	ObservationKinds        []string           `json:"observation_kinds,omitempty"`
	MissingObservationKinds []string           `json:"missing_observation_kinds,omitempty"`
	EvidenceRefs            []EvidenceRef      `json:"evidence_refs,omitempty"`
	OutputRefs              []DisplaySafeRef   `json:"output_refs,omitempty"`
	FailureClass            FailureClass       `json:"failure_class,omitempty"`
	FailureReason           string             `json:"failure_reason,omitempty"`
	MissingInputs           []MissingInput     `json:"missing_inputs,omitempty"`
	Boundaries              []Boundary         `json:"boundaries,omitempty"`
	NextHostAction          NextHostAction     `json:"next_host_action,omitempty"`
	RunnerEffect            string             `json:"runner_effect,omitempty"`
	PromptEffect            string             `json:"prompt_effect,omitempty"`
	RawOutputLoaded         bool               `json:"raw_output_loaded"`
}

// CloneObservationNormalizationResult returns a deep copy of in.
func CloneObservationNormalizationResult(in ObservationNormalizationResult) ObservationNormalizationResult {
	out := in
	out.Frame = in.Frame.Clone()
	out.Observations = cloneObservations(in.Observations)
	out.ObservationKinds = cloneStringSlice(in.ObservationKinds)
	out.MissingObservationKinds = cloneStringSlice(in.MissingObservationKinds)
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.OutputRefs = cloneDisplaySafeRefs(in.OutputRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

// Clone returns a deep copy of r.
func (r ObservationNormalizationResult) Clone() ObservationNormalizationResult {
	return CloneObservationNormalizationResult(r)
}

// Normalize applies deterministic defaults and display-safe normalization.
func (r ObservationNormalizationResult) Normalize() ObservationNormalizationResult {
	out := CloneObservationNormalizationResult(r)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeVerificationStatus(string(out.Status))
	out.Frame = out.Frame.Normalize()
	out.SourceKind = normalizeControlToken(out.SourceKind)
	out.SourceRef = normalizeOneDisplaySafeRef(out.SourceRef)
	out.RuntimeAdapterRunRef = normalizeOneDisplaySafeRef(out.RuntimeAdapterRunRef)
	out.AdapterRef = normalizeOneDisplaySafeRef(out.AdapterRef)
	out.StrategyRef = normalizeOneDisplaySafeRef(out.StrategyRef)
	out.Observations = normalizeObservations(out.Observations)
	out.ObservationKinds = normalizeControlTokenList(out.ObservationKinds)
	out.MissingObservationKinds = normalizeControlTokenList(out.MissingObservationKinds)
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
	out.OutputRefs = normalizeDisplaySafeRefs(out.OutputRefs)
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
		out.ReadyForVerification = false
		if out.FailureClass == FailureNone {
			out.FailureClass = FailureEvidenceWeak
		}
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		if out.NextHostAction == "" {
			out.NextHostAction = "provide_display_safe_refs"
		}
		return out
	}
	out.ReadyForVerification = out.Status == VerificationSatisfied &&
		len(out.Observations) > 0 &&
		len(out.MissingObservationKinds) == 0 &&
		len(out.MissingInputs) == 0
	return out
}
