package controlcontract

import "strings"

type WorkflowStrategyCatalogEntryInput struct {
	SourceRef        DisplaySafeRef     `json:"source_ref,omitempty"`
	WorkflowRef      DisplaySafeRef     `json:"workflow_ref,omitempty"`
	WorkflowSpecRef  DisplaySafeRef     `json:"workflow_spec_ref,omitempty"`
	CandidateRef     DisplaySafeRef     `json:"candidate_ref,omitempty"`
	CapabilityRefs   []DisplaySafeRef   `json:"capability_refs,omitempty"`
	ExpectedEvidence []EvidenceRef      `json:"expected_evidence,omitempty"`
	Preconditions    []MissingInput     `json:"preconditions,omitempty"`
	EvidenceRefs     []EvidenceRef      `json:"evidence_refs,omitempty"`
	MinIntensity     ExecutionIntensity `json:"min_intensity,omitempty"`
	MaxIntensity     ExecutionIntensity `json:"max_intensity,omitempty"`
	ApprovalOptional bool               `json:"approval_optional"`
	Risk             string             `json:"risk,omitempty"`
	SideEffectClass  string             `json:"side_effect_class,omitempty"`
	Owner            string             `json:"owner,omitempty"`
	Status           VerificationStatus `json:"status,omitempty"`
	FailureClass     FailureClass       `json:"failure_class,omitempty"`
	MissingInputs    []MissingInput     `json:"missing_inputs,omitempty"`
	Boundaries       []Boundary         `json:"boundaries,omitempty"`
	RawOutputLoaded  bool               `json:"raw_output_loaded"`
}

type WorkflowStrategyCatalogSnapshotInput struct {
	CatalogRef      DisplaySafeRef                      `json:"catalog_ref,omitempty"`
	Entries         []WorkflowStrategyCatalogEntryInput `json:"entries,omitempty"`
	PolicyRefs      []DisplaySafeRef                    `json:"policy_refs,omitempty"`
	EvidenceRefs    []EvidenceRef                       `json:"evidence_refs,omitempty"`
	MissingInputs   []MissingInput                      `json:"missing_inputs,omitempty"`
	Boundaries      []Boundary                          `json:"boundaries,omitempty"`
	RawOutputLoaded bool                                `json:"raw_output_loaded"`
}

func BuildWorkflowStrategyCatalogEntry(input WorkflowStrategyCatalogEntryInput) StrategyCatalogEntry {
	sourceRef := firstDisplaySafeRef(input.SourceRef, input.WorkflowRef, input.WorkflowSpecRef)
	candidateRef := firstDisplaySafeRef(input.CandidateRef, workflowStrategyCatalogDefaultCandidateRef(sourceRef))
	minIntensity := normalizeIntensityOr(input.MinIntensity, IntensityL3ManagedObjective)
	maxIntensity := normalizeIntensityOr(input.MaxIntensity, minIntensity)
	if executionIntensityRank(maxIntensity) < executionIntensityRank(minIntensity) {
		maxIntensity = minIntensity
	}
	expectedEvidence := normalizeEvidenceRefs(input.ExpectedEvidence)
	if len(expectedEvidence) == 0 && sourceRef != "" {
		expectedEvidence = []EvidenceRef{workflowStrategyCatalogEvidence(sourceRef)}
	}
	status := NormalizeVerificationStatus(string(input.Status))
	if status == VerificationNotEvaluated {
		status = VerificationSatisfied
	}
	failure := NormalizeFailureClass(string(input.FailureClass))
	missingInputs := normalizeMissingInputs(input.MissingInputs)
	boundaries := MergeBoundaries([]Boundary{
		"workflow_strategy_catalog_entry",
		"workflow_strategy_metadata_only",
		"workflow_strategy_requires_host_runtime_backend",
		"workflow_strategy_candidate_not_executed",
		"controlplane_does_not_execute_workflow",
		"no_workflow_dispatch",
		"no_runner_dispatch",
	}, input.Boundaries)
	rawOutputLoaded := input.RawOutputLoaded || workflowStrategyCatalogEntryInputUnsafe(input)
	if rawOutputLoaded {
		status = VerificationBlocked
		failure = firstFailureClass(failure, FailureEvidenceWeak)
		missingInputs = AppendMissingInputs(missingInputs, "host:display_safe_refs")
		boundaries = AppendBoundaries(boundaries, "raw_output_not_allowed")
	}
	if sourceRef == "" {
		status = VerificationBlocked
		failure = firstFailureClass(failure, FailureConfigMissing)
		missingInputs = AppendMissingInputs(missingInputs, "host:workflow_strategy_source_ref")
		boundaries = AppendBoundaries(boundaries, "workflow_strategy_source_ref_missing")
	}
	if candidateRef == "" {
		status = VerificationBlocked
		failure = firstFailureClass(failure, FailureConfigMissing)
		missingInputs = AppendMissingInputs(missingInputs, "host:workflow_strategy_candidate_ref")
		boundaries = AppendBoundaries(boundaries, "workflow_strategy_candidate_ref_missing")
	}
	return StrategyCatalogEntry{
		SourceKind:   StrategyCatalogSourceWorkflow,
		SourceRef:    sourceRef,
		Status:       status,
		FailureClass: failure,
		Candidate: StrategyCandidate{
			ID:               string(candidateRef),
			Kind:             "workflow_runtime_strategy",
			ControlMode:      ControlModeWorkflow,
			MinIntensity:     minIntensity,
			MaxIntensity:     maxIntensity,
			CapabilityRefs:   normalizeDisplaySafeRefs(input.CapabilityRefs),
			ExpectedEvidence: expectedEvidence,
			Preconditions:    normalizeMissingInputs(input.Preconditions),
			Boundaries: []Boundary{
				"workflow_strategy_candidate_metadata",
				"workflow_strategy_requires_host_runtime_backend",
				"controlplane_does_not_execute_workflow",
			},
			Risk:             firstNonEmptyControlToken(input.Risk, "requires_review"),
			SideEffectClass:  firstNonEmptyControlToken(input.SideEffectClass, "host_workflow_runtime"),
			RequiresApproval: !input.ApprovalOptional,
			Owner:            firstNonEmptyControlToken(input.Owner, "host"),
		},
		EvidenceRefs:    MergeEvidenceRefs(input.EvidenceRefs, expectedEvidence),
		MissingInputs:   missingInputs,
		Boundaries:      boundaries,
		RawOutputLoaded: rawOutputLoaded,
	}.Normalize()
}

func BuildWorkflowStrategyCatalogSnapshot(input WorkflowStrategyCatalogSnapshotInput) StrategyCatalogSnapshot {
	entries := make([]StrategyCatalogEntry, 0, len(input.Entries))
	for _, entry := range input.Entries {
		entries = append(entries, BuildWorkflowStrategyCatalogEntry(entry))
	}
	rawOutputLoaded := input.RawOutputLoaded || workflowStrategyCatalogSnapshotInputUnsafe(input)
	missingInputs := normalizeMissingInputs(input.MissingInputs)
	boundaries := MergeBoundaries([]Boundary{"workflow_strategy_catalog_snapshot", "workflow_strategy_catalog_projection_only"}, input.Boundaries)
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

func workflowStrategyCatalogDefaultCandidateRef(sourceRef DisplaySafeRef) DisplaySafeRef {
	if sourceRef == "" {
		return ""
	}
	return DisplaySafeRef("strategy:" + workflowStrategyCatalogSuffix(sourceRef) + "_workflow_runtime")
}

func workflowStrategyCatalogEvidence(sourceRef DisplaySafeRef) EvidenceRef {
	suffix := workflowStrategyCatalogSuffix(sourceRef)
	return EvidenceRef{
		Ref:      DisplaySafeRef("evidence:" + suffix + "_workflow_node_result"),
		Kind:     "workflow_node_result",
		Strength: EvidenceStrong,
		Source:   sourceRef,
	}.Normalize()
}

func workflowStrategyCatalogSuffix(ref DisplaySafeRef) string {
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
		return "workflow_runtime"
	}
	return token
}

func workflowStrategyCatalogEntryInputUnsafe(input WorkflowStrategyCatalogEntryInput) bool {
	if input.RawOutputLoaded {
		return true
	}
	for _, value := range []DisplaySafeRef{
		input.SourceRef,
		input.WorkflowRef,
		input.WorkflowSpecRef,
		input.CandidateRef,
	} {
		if workflowStrategyCatalogRefUnsafe(value) {
			return true
		}
	}
	return workflowStrategyCatalogEvidenceUnsafe(input.ExpectedEvidence) || workflowStrategyCatalogEvidenceUnsafe(input.EvidenceRefs)
}

func workflowStrategyCatalogSnapshotInputUnsafe(input WorkflowStrategyCatalogSnapshotInput) bool {
	if input.RawOutputLoaded || workflowStrategyCatalogRefUnsafe(input.CatalogRef) {
		return true
	}
	return workflowStrategyCatalogEvidenceUnsafe(input.EvidenceRefs)
}

func workflowStrategyCatalogRefUnsafe(value DisplaySafeRef) bool {
	raw := strings.TrimSpace(string(value))
	return raw != "" && ContainsUnsafeRawOutput(raw)
}

func workflowStrategyCatalogEvidenceUnsafe(values []EvidenceRef) bool {
	for _, value := range values {
		if workflowStrategyCatalogRefUnsafe(value.Ref) || workflowStrategyCatalogRefUnsafe(value.Source) {
			return true
		}
	}
	return false
}
