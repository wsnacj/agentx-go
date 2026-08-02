package controlcontract

// ObservationNormalizationInput carries structured source results into the
// portable observation schema. It is projection-only and never executes a
// tool, adapter, workflow, schedule, or install action.
type ObservationNormalizationInput struct {
	Frame                    ObjectiveFrame                `json:"frame,omitempty"`
	SourceKind               string                        `json:"source_kind,omitempty"`
	SourceRef                DisplaySafeRef                `json:"source_ref,omitempty"`
	RuntimeAdapterResult     RuntimeAdapterExecutionResult `json:"runtime_adapter_result,omitempty"`
	Observations             []Observation                 `json:"observations,omitempty"`
	EvidenceRefs             []EvidenceRef                 `json:"evidence_refs,omitempty"`
	ExpectedObservationKinds []string                      `json:"expected_observation_kinds,omitempty"`
	Boundaries               []Boundary                    `json:"boundaries,omitempty"`
	RawOutputLoaded          bool                          `json:"raw_output_loaded"`
}

// BuildObservationNormalization combines a structured host result with the
// canonical observation normalizer while preserving adapter identity.
func BuildObservationNormalization(input ObservationNormalizationInput) ObservationNormalizationResult {
	frame := input.Frame.Normalize()
	adapterResult := input.RuntimeAdapterResult.Normalize()
	hasAdapterResult := observationNormalizationHasRuntimeAdapterResult(input.RuntimeAdapterResult)
	sourceKind := normalizeControlToken(input.SourceKind)
	if sourceKind == "" && hasAdapterResult {
		sourceKind = "host_adapter_result"
	}
	sourceRef := firstDisplaySafeRef(input.SourceRef, adapterResult.AdapterRef, adapterResult.HostAdapterRunRef)
	result := ObservationNormalizationResult{
		ContractVersion:      ContractVersion,
		Projected:            true,
		Status:               VerificationBlocked,
		Frame:                frame,
		SourceKind:           sourceKind,
		SourceRef:            sourceRef,
		RuntimeAdapterRunRef: adapterResult.HostAdapterRunRef,
		AdapterRef:           adapterResult.AdapterRef,
		StrategyRef:          adapterResult.StrategyRef,
		EvidenceRefs:         MergeEvidenceRefs(input.EvidenceRefs, adapterResult.EvidenceRefs),
		OutputRefs:           normalizeDisplaySafeRefs(adapterResult.OutputRefs),
		FailureClass:         FailureNone,
		FailureReason:        managedObjectiveReplannerSafeReason(adapterResult.FailureReason),
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
			adapterResult.Boundaries,
		),
		NextHostAction:  "normalize_observations",
		RunnerEffect:    "none",
		PromptEffect:    "none",
		RawOutputLoaded: input.RawOutputLoaded || adapterResult.RawOutputLoaded,
	}
	if observationNormalizationInputUnsafe(input) {
		return observationNormalizationBlock(result, VerificationReviewRequired, FailureEvidenceWeak, "host:display_safe_refs", "provide_display_safe_refs", "raw_output_not_allowed")
	}
	if hasAdapterResult && !adapterResult.ReadyForObservationNormalization {
		status := adapterResult.Status
		if status == VerificationSatisfied || status == VerificationNotEvaluated || status == VerificationNotApplicable {
			status = VerificationBlocked
		}
		return observationNormalizationBlock(
			result,
			status,
			firstFailureClass(adapterResult.FailureClass, FailureEvidenceMissing),
			"host:runtime_adapter_result",
			firstNextHostAction(adapterResult.NextHostAction, "review_runtime_adapter_result"),
			"runtime_adapter_result_not_ready_for_normalization",
		)
	}

	result = BuildStructuredObservationNormalization(StructuredObservationNormalizationInput{
		Frame:                    frame,
		SourceKind:               sourceKind,
		SourceRef:                sourceRef,
		Observations:             append(cloneObservations(adapterResult.Observations), input.Observations...),
		EvidenceRefs:             MergeEvidenceRefs(input.EvidenceRefs, adapterResult.EvidenceRefs),
		ExpectedObservationKinds: append(cloneStringSlice(input.ExpectedObservationKinds), adapterResult.Request.ExpectedObservationKinds...),
		Boundaries:               MergeBoundaries(input.Boundaries, adapterResult.Boundaries),
		RawOutputLoaded:          input.RawOutputLoaded || adapterResult.RawOutputLoaded,
	})
	result.RuntimeAdapterRunRef = adapterResult.HostAdapterRunRef
	result.AdapterRef = adapterResult.AdapterRef
	result.StrategyRef = adapterResult.StrategyRef
	result.OutputRefs = normalizeDisplaySafeRefs(adapterResult.OutputRefs)
	result.FailureReason = managedObjectiveReplannerSafeReason(adapterResult.FailureReason)
	return result.Normalize()
}

func observationNormalizationBlock(result ObservationNormalizationResult, status VerificationStatus, failure FailureClass, missing MissingInput, next NextHostAction, boundary Boundary) ObservationNormalizationResult {
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

func observationNormalizationHasRuntimeAdapterResult(result RuntimeAdapterExecutionResult) bool {
	return result.ContractVersion != "" || result.Projected || result.HostAdapterRunRef != "" || result.AdapterRef != "" || result.StrategyRef != "" || len(result.Observations) > 0 || len(result.EvidenceRefs) > 0 || len(result.OutputRefs) > 0 || len(result.MissingCapabilityRefs) > 0 || len(result.MissingInputs) > 0 || result.RawOutputLoaded
}

func observationNormalizationInputUnsafe(input ObservationNormalizationInput) bool {
	if displaySafeRefRejected(input.SourceRef) || evidenceRefRejected(input.EvidenceRefs) || observationSliceUnsafePayload(input.Observations) || input.RawOutputLoaded {
		return true
	}
	result := input.RuntimeAdapterResult
	return runtimeAdapterExecutionResultUnsafe(RuntimeAdapterExecutionResultInput{
		Request:               result.Request,
		AdapterRef:            result.AdapterRef,
		StrategyRef:           result.StrategyRef,
		HostAdapterRunRef:     result.HostAdapterRunRef,
		Observations:          result.Observations,
		EvidenceRefs:          result.EvidenceRefs,
		OutputRefs:            result.OutputRefs,
		MissingCapabilityRefs: result.MissingCapabilityRefs,
		RawOutputLoaded:       result.RawOutputLoaded,
	}) || observationSliceUnsafePayload(result.Observations)
}

// ObservationsContainUnsafePayload reports whether observations contain raw or
// non-display-safe values that must not enter a portable Objective contract.
func ObservationsContainUnsafePayload(values []Observation) bool {
	return observationSliceUnsafePayload(values)
}
