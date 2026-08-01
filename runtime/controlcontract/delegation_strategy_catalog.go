package controlcontract

import "strings"

type DelegationStrategyCatalogEntryInput struct {
	SourceRef        DisplaySafeRef              `json:"source_ref,omitempty"`
	RequestRef       DisplaySafeRef              `json:"request_ref,omitempty"`
	Request          DelegationRequestProjection `json:"request,omitempty"`
	CandidateRef     DisplaySafeRef              `json:"candidate_ref,omitempty"`
	CapabilityRefs   []DisplaySafeRef            `json:"capability_refs,omitempty"`
	ExpectedEvidence []EvidenceRef               `json:"expected_evidence,omitempty"`
	Preconditions    []MissingInput              `json:"preconditions,omitempty"`
	EvidenceRefs     []EvidenceRef               `json:"evidence_refs,omitempty"`
	MinIntensity     ExecutionIntensity          `json:"min_intensity,omitempty"`
	MaxIntensity     ExecutionIntensity          `json:"max_intensity,omitempty"`
	ApprovalOptional bool                        `json:"approval_optional"`
	Risk             string                      `json:"risk,omitempty"`
	SideEffectClass  string                      `json:"side_effect_class,omitempty"`
	Owner            string                      `json:"owner,omitempty"`
	Status           VerificationStatus          `json:"status,omitempty"`
	FailureClass     FailureClass                `json:"failure_class,omitempty"`
	MissingInputs    []MissingInput              `json:"missing_inputs,omitempty"`
	Boundaries       []Boundary                  `json:"boundaries,omitempty"`
	RawOutputLoaded  bool                        `json:"raw_output_loaded"`
}

type DelegationStrategyCatalogSnapshotInput struct {
	CatalogRef      DisplaySafeRef                        `json:"catalog_ref,omitempty"`
	Entries         []DelegationStrategyCatalogEntryInput `json:"entries,omitempty"`
	PolicyRefs      []DisplaySafeRef                      `json:"policy_refs,omitempty"`
	EvidenceRefs    []EvidenceRef                         `json:"evidence_refs,omitempty"`
	MissingInputs   []MissingInput                        `json:"missing_inputs,omitempty"`
	Boundaries      []Boundary                            `json:"boundaries,omitempty"`
	RawOutputLoaded bool                                  `json:"raw_output_loaded"`
}

func BuildDelegationStrategyCatalogEntry(input DelegationStrategyCatalogEntryInput) StrategyCatalogEntry {
	requestProvided := !delegationRequestProjectionEmpty(input.Request)
	request := input.Request.Normalize()
	sourceRef := firstDisplaySafeRef(input.SourceRef, input.RequestRef, request.SubgoalRef, request.WorkerRef)
	candidateRef := firstDisplaySafeRef(input.CandidateRef, delegationStrategyCatalogDefaultCandidateRef(sourceRef))
	minIntensity := normalizeIntensityOr(input.MinIntensity, firstIntensity(request.RequestedIntensity, IntensityL4DurableLongRun))
	maxIntensity := normalizeIntensityOr(input.MaxIntensity, minIntensity)
	if executionIntensityRank(maxIntensity) < executionIntensityRank(minIntensity) {
		maxIntensity = minIntensity
	}
	expectedEvidence := normalizeEvidenceRefs(input.ExpectedEvidence)
	if len(expectedEvidence) == 0 {
		expectedEvidence = normalizeEvidenceRefs(request.EvidenceRequirements)
	}
	if len(expectedEvidence) == 0 && sourceRef != "" {
		expectedEvidence = []EvidenceRef{delegationStrategyCatalogEvidence(sourceRef)}
	}
	status := NormalizeVerificationStatus(string(input.Status))
	if status == VerificationNotEvaluated {
		status = VerificationSatisfied
	}
	failure := NormalizeFailureClass(string(input.FailureClass))
	missingInputs := normalizeMissingInputs(input.MissingInputs)
	boundaries := MergeBoundaries([]Boundary{
		"delegation_strategy_catalog_entry",
		"delegation_strategy_metadata_only",
		"delegation_strategy_requires_host_worker_runtime",
		"delegation_strategy_requires_explicit_request_before_dispatch",
		"delegation_strategy_candidate_not_executed",
		"controlplane_does_not_dispatch_worker",
		"worker_result_requires_verification",
		"worker_output_not_fact",
		"no_worker_dispatch",
		"no_runner_dispatch",
	}, request.Boundaries, input.Boundaries)
	rawOutputLoaded := input.RawOutputLoaded || request.RawOutputLoaded || delegationStrategyCatalogEntryInputUnsafe(input)
	if rawOutputLoaded {
		status = VerificationBlocked
		failure = firstFailureClass(failure, FailureEvidenceWeak)
		missingInputs = AppendMissingInputs(missingInputs, "host:display_safe_refs")
		boundaries = AppendBoundaries(boundaries, "raw_output_not_allowed")
	}
	if sourceRef == "" {
		status = VerificationBlocked
		failure = firstFailureClass(failure, FailureConfigMissing)
		missingInputs = AppendMissingInputs(missingInputs, "host:delegation_strategy_source_ref")
		boundaries = AppendBoundaries(boundaries, "delegation_strategy_source_ref_missing")
	}
	if candidateRef == "" {
		status = VerificationBlocked
		failure = firstFailureClass(failure, FailureConfigMissing)
		missingInputs = AppendMissingInputs(missingInputs, "host:delegation_strategy_candidate_ref")
		boundaries = AppendBoundaries(boundaries, "delegation_strategy_candidate_ref_missing")
	}
	if requestProvided && !request.ReadyForWorkerDispatch {
		status = VerificationBlocked
		failure = firstFailureClass(failure, request.FailureClass, FailurePolicyBlocked)
		missingInputs = AppendMissingInputs(missingInputs, request.MissingInputs...)
		boundaries = AppendBoundaries(boundaries, "delegation_request_not_ready_for_strategy_catalog")
	}
	return StrategyCatalogEntry{
		SourceKind:   StrategyCatalogSourceDelegation,
		SourceRef:    sourceRef,
		Status:       status,
		FailureClass: failure,
		Candidate: StrategyCandidate{
			ID:               string(candidateRef),
			Kind:             "delegation_worker_strategy",
			ControlMode:      ControlModeDelegated,
			MinIntensity:     minIntensity,
			MaxIntensity:     maxIntensity,
			CapabilityRefs:   normalizeDisplaySafeRefs(input.CapabilityRefs),
			ExpectedEvidence: expectedEvidence,
			Preconditions:    normalizeMissingInputs(input.Preconditions),
			Boundaries: []Boundary{
				"delegation_strategy_candidate_metadata",
				"delegation_strategy_requires_host_worker_runtime",
				"controlplane_does_not_dispatch_worker",
				"worker_result_requires_verification",
			},
			Risk:             firstNonEmptyControlToken(input.Risk, "requires_review"),
			SideEffectClass:  firstNonEmptyControlToken(input.SideEffectClass, "delegation_worker_runtime"),
			RequiresApproval: !input.ApprovalOptional,
			Owner:            firstNonEmptyControlToken(input.Owner, "host"),
		},
		EvidenceRefs:    MergeEvidenceRefs(input.EvidenceRefs, request.EvidenceRequirements, expectedEvidence),
		MissingInputs:   missingInputs,
		Boundaries:      boundaries,
		RawOutputLoaded: rawOutputLoaded,
	}.Normalize()
}

func BuildDelegationStrategyCatalogSnapshot(input DelegationStrategyCatalogSnapshotInput) StrategyCatalogSnapshot {
	entries := make([]StrategyCatalogEntry, 0, len(input.Entries))
	for _, entry := range input.Entries {
		entries = append(entries, BuildDelegationStrategyCatalogEntry(entry))
	}
	rawOutputLoaded := input.RawOutputLoaded || delegationStrategyCatalogSnapshotInputUnsafe(input)
	missingInputs := normalizeMissingInputs(input.MissingInputs)
	boundaries := MergeBoundaries([]Boundary{"delegation_strategy_catalog_snapshot", "delegation_strategy_catalog_projection_only"}, input.Boundaries)
	if rawOutputLoaded {
		missingInputs = AppendMissingInputs(missingInputs, "host:display_safe_refs")
		boundaries = AppendBoundaries(boundaries, "raw_output_not_allowed")
	}
	return StrategyCatalogSnapshot{
		CatalogRef:      normalizeOneDisplaySafeRef(input.CatalogRef),
		Entries:         entries,
		PolicyRefs:      normalizeDisplaySafeRefs(input.PolicyRefs),
		EvidenceRefs:    normalizeEvidenceRefs(input.EvidenceRefs),
		MissingInputs:   missingInputs,
		Boundaries:      boundaries,
		RawOutputLoaded: rawOutputLoaded,
	}.Normalize()
}

func delegationStrategyCatalogDefaultCandidateRef(sourceRef DisplaySafeRef) DisplaySafeRef {
	if sourceRef == "" {
		return ""
	}
	return DisplaySafeRef("strategy:" + delegationStrategyCatalogSuffix(sourceRef) + "_delegation_worker")
}

func delegationStrategyCatalogEvidence(sourceRef DisplaySafeRef) EvidenceRef {
	suffix := delegationStrategyCatalogSuffix(sourceRef)
	return EvidenceRef{
		Ref:      DisplaySafeRef("evidence:" + suffix + "_delegation_worker_result"),
		Kind:     "delegation_worker_result",
		Strength: EvidenceAdequate,
		Source:   sourceRef,
	}.Normalize()
}

func delegationStrategyCatalogSuffix(ref DisplaySafeRef) string {
	value := strings.ToLower(strings.TrimSpace(string(ref)))
	if idx := strings.LastIndex(value, ":"); idx >= 0 && idx+1 < len(value) {
		value = value[idx+1:]
	}
	var b strings.Builder
	lastUnderscore := false
	for _, r := range value {
		keep := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
		if keep {
			b.WriteRune(r)
			lastUnderscore = false
			continue
		}
		if !lastUnderscore {
			b.WriteByte('_')
			lastUnderscore = true
		}
	}
	token := strings.Trim(b.String(), "_")
	if token == "" {
		return "delegation_worker"
	}
	return token
}

func delegationStrategyCatalogEntryInputUnsafe(input DelegationStrategyCatalogEntryInput) bool {
	if input.RawOutputLoaded {
		return true
	}
	for _, value := range []DisplaySafeRef{
		input.SourceRef,
		input.RequestRef,
		input.CandidateRef,
	} {
		if delegationStrategyCatalogRefUnsafe(value) {
			return true
		}
	}
	return delegationStrategyCatalogEvidenceUnsafe(input.ExpectedEvidence) || delegationStrategyCatalogEvidenceUnsafe(input.EvidenceRefs)
}

func delegationStrategyCatalogSnapshotInputUnsafe(input DelegationStrategyCatalogSnapshotInput) bool {
	if input.RawOutputLoaded || delegationStrategyCatalogRefUnsafe(input.CatalogRef) {
		return true
	}
	return delegationStrategyCatalogEvidenceUnsafe(input.EvidenceRefs)
}

func delegationStrategyCatalogRefUnsafe(value DisplaySafeRef) bool {
	raw := strings.TrimSpace(string(value))
	return raw != "" && ContainsUnsafeRawOutput(raw)
}

func delegationStrategyCatalogEvidenceUnsafe(values []EvidenceRef) bool {
	for _, value := range values {
		if delegationStrategyCatalogRefUnsafe(value.Ref) || delegationStrategyCatalogRefUnsafe(value.Source) {
			return true
		}
	}
	return false
}

func delegationRequestProjectionEmpty(request DelegationRequestProjection) bool {
	return request.ContractVersion == "" &&
		!request.Projected &&
		request.Status == "" &&
		request.Activation == "" &&
		request.Frame.ID == "" &&
		request.SubgoalRef == "" &&
		request.WorkerRef == "" &&
		len(request.EvidenceRequirements) == 0 &&
		len(request.MissingInputs) == 0 &&
		len(request.BlockedReasons) == 0 &&
		len(request.Boundaries) == 0 &&
		!request.RawOutputLoaded
}
