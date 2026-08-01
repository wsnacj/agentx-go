package controlcontract

import "strings"

type MemoryProposalKind string

const (
	MemoryProposalKindSkill    MemoryProposalKind = "skill"
	MemoryProposalKindWorkflow MemoryProposalKind = "workflow"
	MemoryProposalKindTemplate MemoryProposalKind = "template"
)

func KnownMemoryProposalKinds() []MemoryProposalKind {
	return []MemoryProposalKind{
		MemoryProposalKindSkill,
		MemoryProposalKindWorkflow,
		MemoryProposalKindTemplate,
	}
}

func NormalizeMemoryProposalKind(raw string) MemoryProposalKind {
	switch normalizeEnumToken(raw) {
	case "skill", "skills":
		return MemoryProposalKindSkill
	case "workflow", "workflows", "flow":
		return MemoryProposalKindWorkflow
	case "template", "templates", "prompt_template":
		return MemoryProposalKindTemplate
	default:
		return ""
	}
}

type RepeatedSuccessMemoryProposalInput struct {
	Activation            Activation              `json:"activation,omitempty"`
	Frame                 ObjectiveFrame          `json:"frame,omitempty"`
	ProposalSetRef        DisplaySafeRef          `json:"proposal_set_ref,omitempty"`
	ProposalOwnerRef      DisplaySafeRef          `json:"proposal_owner_ref,omitempty"`
	MemoryPolicyRef       DisplaySafeRef          `json:"memory_policy_ref,omitempty"`
	StrategyCatalog       StrategyCatalogSnapshot `json:"strategy_catalog,omitempty"`
	Attempts              []AttemptSummary        `json:"attempts,omitempty"`
	ProposalKinds         []MemoryProposalKind    `json:"proposal_kinds,omitempty"`
	MinSuccessfulAttempts int                     `json:"min_successful_attempts,omitempty"`
	MaxProposalCount      int                     `json:"max_proposal_count,omitempty"`
	EvidenceRefs          []EvidenceRef           `json:"evidence_refs,omitempty"`
	ProvenanceRefs        []DisplaySafeRef        `json:"provenance_refs,omitempty"`
	DecisionBasis         []DisplaySafeRef        `json:"decision_basis,omitempty"`
	Boundaries            []Boundary              `json:"boundaries,omitempty"`
	RawOutputLoaded       bool                    `json:"raw_output_loaded"`
}

type MemoryProposal struct {
	ContractVersion    string                    `json:"contract_version,omitempty"`
	Kind               MemoryProposalKind        `json:"kind,omitempty"`
	ProposalRef        DisplaySafeRef            `json:"proposal_ref,omitempty"`
	SourceStrategyRef  DisplaySafeRef            `json:"source_strategy_ref,omitempty"`
	SourceKind         StrategyCatalogSourceKind `json:"source_kind,omitempty"`
	SourceRef          DisplaySafeRef            `json:"source_ref,omitempty"`
	Candidate          StrategyCandidate         `json:"candidate,omitempty"`
	SupportCount       int                       `json:"support_count,omitempty"`
	SupportAttemptRefs []AttemptRef              `json:"support_attempt_refs,omitempty"`
	EvidenceRefs       []EvidenceRef             `json:"evidence_refs,omitempty"`
	ProvenanceRefs     []DisplaySafeRef          `json:"provenance_refs,omitempty"`
	Boundaries         []Boundary                `json:"boundaries,omitempty"`
}

type RepeatedSuccessMemoryProposalSet struct {
	ContractVersion        string               `json:"contract_version,omitempty"`
	Projected              bool                 `json:"projected"`
	Status                 HostActionStatus     `json:"status,omitempty"`
	ReadyForHostReview     bool                 `json:"ready_for_host_review"`
	ProposalOnly           bool                 `json:"proposal_only"`
	Activation             Activation           `json:"activation,omitempty"`
	Frame                  ObjectiveFrame       `json:"frame,omitempty"`
	ProposalSetRef         DisplaySafeRef       `json:"proposal_set_ref,omitempty"`
	ProposalOwnerRef       DisplaySafeRef       `json:"proposal_owner_ref,omitempty"`
	MemoryPolicyRef        DisplaySafeRef       `json:"memory_policy_ref,omitempty"`
	StrategyCatalogRef     DisplaySafeRef       `json:"strategy_catalog_ref,omitempty"`
	MinSuccessfulAttempts  int                  `json:"min_successful_attempts,omitempty"`
	MaxProposalCount       int                  `json:"max_proposal_count,omitempty"`
	AttemptCount           int                  `json:"attempt_count,omitempty"`
	SuccessfulAttemptCount int                  `json:"successful_attempt_count,omitempty"`
	RepeatedStrategyRefs   []DisplaySafeRef     `json:"repeated_strategy_refs,omitempty"`
	ProposalKinds          []MemoryProposalKind `json:"proposal_kinds,omitempty"`
	Proposals              []MemoryProposal     `json:"proposals,omitempty"`
	EvidenceRefs           []EvidenceRef        `json:"evidence_refs,omitempty"`
	ProvenanceRefs         []DisplaySafeRef     `json:"provenance_refs,omitempty"`
	MissingInputs          []MissingInput       `json:"missing_inputs,omitempty"`
	BlockedReasons         []string             `json:"blocked_reasons,omitempty"`
	FailureClass           FailureClass         `json:"failure_class,omitempty"`
	DecisionBasis          []DisplaySafeRef     `json:"decision_basis,omitempty"`
	Boundaries             []Boundary           `json:"boundaries,omitempty"`
	NextHostAction         NextHostAction       `json:"next_host_action,omitempty"`
	SkillWriteExecuted     bool                 `json:"skill_write_executed"`
	WorkflowWriteExecuted  bool                 `json:"workflow_write_executed"`
	TemplateWriteExecuted  bool                 `json:"template_write_executed"`
	InstallExecuted        bool                 `json:"install_executed"`
	RuntimeReloadExecuted  bool                 `json:"runtime_reload_executed"`
	CoreMutationExecuted   bool                 `json:"core_mutation_executed"`
	RunnerEffect           string               `json:"runner_effect,omitempty"`
	PromptEffect           string               `json:"prompt_effect,omitempty"`
	RawOutputLoaded        bool                 `json:"raw_output_loaded"`
}

func BuildRepeatedSuccessMemoryProposal(input RepeatedSuccessMemoryProposalInput) RepeatedSuccessMemoryProposalSet {
	activation := NormalizeActivation(string(input.Activation))
	frame := input.Frame.Normalize()
	catalog := input.StrategyCatalog.Normalize()
	attempts := normalizeAttemptSummaries(input.Attempts)
	proposalKinds := normalizeMemoryProposalKinds(input.ProposalKinds)
	result := RepeatedSuccessMemoryProposalSet{
		ContractVersion:       ContractVersion,
		Projected:             true,
		Status:                HostActionBlocked,
		ProposalOnly:          true,
		Activation:            activation,
		Frame:                 frame,
		ProposalSetRef:        normalizeOneDisplaySafeRef(input.ProposalSetRef),
		ProposalOwnerRef:      normalizeOneDisplaySafeRef(input.ProposalOwnerRef),
		MemoryPolicyRef:       normalizeOneDisplaySafeRef(input.MemoryPolicyRef),
		StrategyCatalogRef:    catalog.CatalogRef,
		MinSuccessfulAttempts: input.MinSuccessfulAttempts,
		MaxProposalCount:      maxNonNegativeInt(input.MaxProposalCount),
		AttemptCount:          len(attempts),
		ProposalKinds:         proposalKinds,
		EvidenceRefs:          normalizeEvidenceRefs(input.EvidenceRefs),
		ProvenanceRefs:        normalizeDisplaySafeRefs(input.ProvenanceRefs),
		FailureClass:          FailureNone,
		DecisionBasis: normalizeDisplaySafeRefs(append(
			[]DisplaySafeRef{
				"memory_proposal:repeated_success",
				"memory_proposal:proposal_only",
			},
			input.DecisionBasis...,
		)),
		Boundaries:      repeatedSuccessMemoryProposalBoundaries(input.Boundaries),
		RunnerEffect:    "none",
		PromptEffect:    "none",
		RawOutputLoaded: input.RawOutputLoaded || catalog.RawOutputLoaded,
	}
	if repeatedSuccessMemoryProposalInputUnsafe(input) {
		result.RawOutputLoaded = true
		result = repeatedSuccessMemoryProposalBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if activation != ActivationManaged {
		result = repeatedSuccessMemoryProposalBlock(result, FailurePolicyBlocked, "managed_activation_required", "control_plane:managed_activation", "enable_managed_objective")
	}
	if frame.ID == "" {
		result = repeatedSuccessMemoryProposalBlock(result, FailureConfigMissing, "objective_frame_missing", "host:objective_frame", "provide_objective_frame")
	}
	if result.ProposalSetRef == "" {
		result = repeatedSuccessMemoryProposalBlock(result, FailureEvidenceMissing, "memory_proposal_set_ref_missing", "host:memory_proposal_set_ref", "provide_memory_proposal_set_ref")
	}
	if result.ProposalOwnerRef == "" {
		result = repeatedSuccessMemoryProposalBlock(result, FailureConfigMissing, "memory_proposal_owner_missing", "host:memory_proposal_owner", "provide_memory_proposal_owner")
	}
	if result.MemoryPolicyRef == "" {
		result = repeatedSuccessMemoryProposalBlock(result, FailurePolicyBlocked, "memory_policy_ref_missing", "host:memory_policy_ref", "provide_memory_policy")
	}
	if result.MinSuccessfulAttempts < 2 {
		result = repeatedSuccessMemoryProposalBlock(result, FailureConfigMissing, "memory_min_successful_attempts_missing", "host:memory_min_successful_attempts", "provide_memory_policy")
	}
	if result.MaxProposalCount <= 0 {
		result = repeatedSuccessMemoryProposalBlock(result, FailureConfigMissing, "memory_max_proposal_count_missing", "host:memory_max_proposal_count", "provide_memory_policy")
	}
	if len(proposalKinds) == 0 {
		result = repeatedSuccessMemoryProposalBlock(result, FailureConfigMissing, "memory_proposal_kinds_missing", "host:memory_proposal_kinds", "provide_memory_policy")
	}
	if strategyCatalogSnapshotEmpty(catalog) || catalog.CatalogRef == "" || len(catalog.Entries) == 0 {
		result = repeatedSuccessMemoryProposalBlock(result, FailureHostAdapterMissing, "strategy_catalog_missing", "host:strategy_catalog", "provide_strategy_catalog")
	}
	if len(attempts) == 0 {
		result = repeatedSuccessMemoryProposalBlock(result, FailureEvidenceMissing, "attempt_ledger_missing", "host:attempt_ledger", "provide_attempt_ledger")
	}
	support := repeatedSuccessAttemptSupport(attempts)
	result.SuccessfulAttemptCount = support.total
	result.RepeatedStrategyRefs = support.repeatedRefs(result.MinSuccessfulAttempts)
	if len(result.BlockedReasons) == 0 && len(result.MissingInputs) == 0 {
		result.Proposals = buildRepeatedSuccessMemoryProposals(catalog, support, proposalKinds, result.MinSuccessfulAttempts, result.MaxProposalCount, result.ProvenanceRefs, result.EvidenceRefs)
		if len(result.Proposals) == 0 {
			result = repeatedSuccessMemoryProposalBlock(result, FailureEvidenceMissing, "repeated_success_path_missing", "host:repeated_success_attempts", "continue_collecting_success_attempts")
		}
	}
	if len(result.BlockedReasons) == 0 && len(result.MissingInputs) == 0 {
		result.Status = HostActionReady
		result.ReadyForHostReview = true
		result.NextHostAction = "review_memory_proposals"
		result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_memory_proposal_review")
	}
	return result.Normalize()
}

func CloneMemoryProposal(in MemoryProposal) MemoryProposal {
	out := in
	out.Candidate = in.Candidate.Clone()
	out.SupportAttemptRefs = cloneAttemptRefs(in.SupportAttemptRefs)
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.ProvenanceRefs = cloneDisplaySafeRefs(in.ProvenanceRefs)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (p MemoryProposal) Clone() MemoryProposal {
	return CloneMemoryProposal(p)
}

func (p MemoryProposal) Normalize() MemoryProposal {
	out := CloneMemoryProposal(p)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Kind = NormalizeMemoryProposalKind(string(out.Kind))
	out.ProposalRef = normalizeOneDisplaySafeRef(out.ProposalRef)
	out.SourceStrategyRef = normalizeOneDisplaySafeRef(out.SourceStrategyRef)
	out.SourceKind = NormalizeStrategyCatalogSourceKind(string(out.SourceKind))
	out.SourceRef = normalizeOneDisplaySafeRef(out.SourceRef)
	out.Candidate = out.Candidate.Normalize()
	out.SupportAttemptRefs = normalizeAttemptRefs(out.SupportAttemptRefs)
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
	out.ProvenanceRefs = normalizeDisplaySafeRefs(out.ProvenanceRefs)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	if out.SupportCount < 0 {
		out.SupportCount = 0
	}
	return out
}

func CloneRepeatedSuccessMemoryProposalSet(in RepeatedSuccessMemoryProposalSet) RepeatedSuccessMemoryProposalSet {
	out := in
	out.Frame = in.Frame.Clone()
	out.RepeatedStrategyRefs = cloneDisplaySafeRefs(in.RepeatedStrategyRefs)
	out.ProposalKinds = cloneMemoryProposalKinds(in.ProposalKinds)
	out.Proposals = cloneMemoryProposals(in.Proposals)
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.ProvenanceRefs = cloneDisplaySafeRefs(in.ProvenanceRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.DecisionBasis = cloneDisplaySafeRefs(in.DecisionBasis)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (s RepeatedSuccessMemoryProposalSet) Clone() RepeatedSuccessMemoryProposalSet {
	return CloneRepeatedSuccessMemoryProposalSet(s)
}

func (s RepeatedSuccessMemoryProposalSet) Normalize() RepeatedSuccessMemoryProposalSet {
	out := CloneRepeatedSuccessMemoryProposalSet(s)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.Activation = NormalizeActivation(string(out.Activation))
	out.Frame = out.Frame.Normalize()
	out.ProposalSetRef = normalizeOneDisplaySafeRef(out.ProposalSetRef)
	out.ProposalOwnerRef = normalizeOneDisplaySafeRef(out.ProposalOwnerRef)
	out.MemoryPolicyRef = normalizeOneDisplaySafeRef(out.MemoryPolicyRef)
	out.StrategyCatalogRef = normalizeOneDisplaySafeRef(out.StrategyCatalogRef)
	out.RepeatedStrategyRefs = normalizeDisplaySafeRefs(out.RepeatedStrategyRefs)
	out.ProposalKinds = normalizeMemoryProposalKinds(out.ProposalKinds)
	out.Proposals = normalizeMemoryProposals(out.Proposals)
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
	out.ProvenanceRefs = normalizeDisplaySafeRefs(out.ProvenanceRefs)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.BlockedReasons = normalizeControlTokenList(out.BlockedReasons)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.DecisionBasis = normalizeDisplaySafeRefs(out.DecisionBasis)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	if out.MinSuccessfulAttempts < 0 {
		out.MinSuccessfulAttempts = 0
	}
	if out.MaxProposalCount < 0 {
		out.MaxProposalCount = 0
	}
	if out.AttemptCount < 0 {
		out.AttemptCount = 0
	}
	if out.SuccessfulAttemptCount < 0 {
		out.SuccessfulAttemptCount = 0
	}
	if out.RunnerEffect == "" {
		out.RunnerEffect = "none"
	}
	if out.PromptEffect == "" {
		out.PromptEffect = "none"
	}
	if out.RawOutputLoaded {
		out.Status = HostActionReviewRequired
		out.ReadyForHostReview = false
		if out.FailureClass == FailureNone {
			out.FailureClass = FailureEvidenceWeak
		}
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		if out.NextHostAction == "" {
			out.NextHostAction = "provide_display_safe_refs"
		}
	}
	out.ProposalOnly = true
	out.SkillWriteExecuted = false
	out.WorkflowWriteExecuted = false
	out.TemplateWriteExecuted = false
	out.InstallExecuted = false
	out.RuntimeReloadExecuted = false
	out.CoreMutationExecuted = false
	out.ReadyForHostReview = out.Status == HostActionReady &&
		len(out.Proposals) > 0 &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.RawOutputLoaded
	return out
}

type repeatedSuccessSupport struct {
	byStrategy map[DisplaySafeRef][]AttemptSummary
	total      int
}

func repeatedSuccessAttemptSupport(attempts []AttemptSummary) repeatedSuccessSupport {
	support := repeatedSuccessSupport{byStrategy: map[DisplaySafeRef][]AttemptSummary{}}
	for _, attempt := range normalizeAttemptSummaries(attempts) {
		strategyRef, ok := NormalizeDisplaySafeRef(attempt.StrategyID)
		if !ok || attempt.Status != VerificationSatisfied || attempt.RawOutputLoaded || attempt.ObservationCount <= 0 || len(attempt.EvidenceRefs) == 0 {
			continue
		}
		support.total++
		support.byStrategy[strategyRef] = append(support.byStrategy[strategyRef], attempt)
	}
	return support
}

func (s repeatedSuccessSupport) repeatedRefs(minSuccessful int) []DisplaySafeRef {
	out := []DisplaySafeRef{}
	for strategyRef, attempts := range s.byStrategy {
		if len(attempts) >= minSuccessful {
			out = appendUniqueDisplaySafeRef(out, strategyRef)
		}
	}
	return normalizeDisplaySafeRefs(out)
}

func buildRepeatedSuccessMemoryProposals(catalog StrategyCatalogSnapshot, support repeatedSuccessSupport, kinds []MemoryProposalKind, minSuccessful int, maxProposalCount int, provenanceRefs []DisplaySafeRef, evidenceRefs []EvidenceRef) []MemoryProposal {
	if maxProposalCount <= 0 {
		return nil
	}
	out := []MemoryProposal{}
	for _, entry := range catalog.Normalize().Entries {
		strategyRef, ok := NormalizeDisplaySafeRef(entry.Candidate.ID)
		if !ok {
			continue
		}
		attempts := support.byStrategy[strategyRef]
		if len(attempts) < minSuccessful {
			continue
		}
		for _, kind := range kinds {
			if len(out) >= maxProposalCount {
				return normalizeMemoryProposals(out)
			}
			proposal := MemoryProposal{
				Kind:               kind,
				ProposalRef:        repeatedSuccessMemoryProposalRef(kind, strategyRef),
				SourceStrategyRef:  strategyRef,
				SourceKind:         entry.SourceKind,
				SourceRef:          entry.SourceRef,
				Candidate:          entry.Candidate,
				SupportCount:       len(attempts),
				SupportAttemptRefs: repeatedSuccessAttemptRefs(attempts),
				EvidenceRefs:       repeatedSuccessEvidenceRefs(attempts, entry, evidenceRefs),
				ProvenanceRefs:     normalizeDisplaySafeRefs(append(cloneDisplaySafeRefs(provenanceRefs), entry.SourceRef)),
				Boundaries:         memoryProposalItemBoundaries(entry.Boundaries),
			}
			out = append(out, proposal.Normalize())
		}
	}
	return normalizeMemoryProposals(out)
}

func repeatedSuccessMemoryProposalRef(kind MemoryProposalKind, strategyRef DisplaySafeRef) DisplaySafeRef {
	token := normalizeControlToken(strings.ReplaceAll(string(strategyRef), ":", "_"))
	ref, ok := NormalizeDisplaySafeRef("proposal:" + string(kind) + ":" + token)
	if !ok {
		return ""
	}
	return ref
}

func repeatedSuccessAttemptRefs(attempts []AttemptSummary) []AttemptRef {
	out := []AttemptRef{}
	for _, attempt := range attempts {
		ref, ok := NormalizeAttemptRef(string(attempt.Ref))
		if ok {
			out = append(out, ref)
		}
	}
	return normalizeAttemptRefs(out)
}

func repeatedSuccessEvidenceRefs(attempts []AttemptSummary, entry StrategyCatalogEntry, extra []EvidenceRef) []EvidenceRef {
	evidence := MergeEvidenceRefs(entry.EvidenceRefs, entry.Candidate.ExpectedEvidence, extra)
	for _, attempt := range attempts {
		evidence = MergeEvidenceRefs(evidence, attempt.EvidenceRefs)
	}
	return normalizeEvidenceRefs(evidence)
}

func repeatedSuccessMemoryProposalBlock(result RepeatedSuccessMemoryProposalSet, failure FailureClass, reason string, missing MissingInput, next NextHostAction) RepeatedSuccessMemoryProposalSet {
	result.Status = HostActionBlocked
	result.ReadyForHostReview = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	return result
}

func repeatedSuccessMemoryProposalInputUnsafe(input RepeatedSuccessMemoryProposalInput) bool {
	if input.RawOutputLoaded ||
		displaySafeRefRejected(input.ProposalSetRef) ||
		displaySafeRefRejected(input.ProposalOwnerRef) ||
		displaySafeRefRejected(input.MemoryPolicyRef) ||
		displaySafeRefSliceRejected(input.ProvenanceRefs) ||
		displaySafeRefSliceRejected(input.DecisionBasis) ||
		evidenceRefRejected(input.EvidenceRefs) {
		return true
	}
	if input.StrategyCatalog.RawOutputLoaded {
		return true
	}
	for _, attempt := range input.Attempts {
		if attempt.RawOutputLoaded || evidenceRefRejected(attempt.EvidenceRefs) {
			return true
		}
		if attempt.Ref != "" && displaySafeRefRejected(DisplaySafeRef(attempt.Ref)) {
			return true
		}
		if attempt.StrategyID != "" && displaySafeRefRejected(DisplaySafeRef(attempt.StrategyID)) {
			return true
		}
	}
	return false
}

func normalizeMemoryProposalKinds(in []MemoryProposalKind) []MemoryProposalKind {
	out := make([]MemoryProposalKind, 0, len(in))
	seen := map[MemoryProposalKind]struct{}{}
	for _, value := range in {
		normalized := NormalizeMemoryProposalKind(string(value))
		if normalized == "" {
			continue
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}
	return out
}

func cloneMemoryProposalKinds(in []MemoryProposalKind) []MemoryProposalKind {
	if len(in) == 0 {
		return nil
	}
	out := make([]MemoryProposalKind, len(in))
	copy(out, in)
	return out
}

func normalizeMemoryProposals(in []MemoryProposal) []MemoryProposal {
	out := make([]MemoryProposal, 0, len(in))
	for _, proposal := range in {
		normalized := proposal.Normalize()
		if normalized.ProposalRef == "" && normalized.SourceStrategyRef == "" {
			continue
		}
		out = append(out, normalized)
	}
	return out
}

func cloneMemoryProposals(in []MemoryProposal) []MemoryProposal {
	if len(in) == 0 {
		return nil
	}
	out := make([]MemoryProposal, len(in))
	for i, proposal := range in {
		out[i] = proposal.Clone()
	}
	return out
}

func repeatedSuccessMemoryProposalBoundaries(extra []Boundary) []Boundary {
	return MergeBoundaries([]Boundary{
		"repeated_success_memory_proposal",
		"proposal_only",
		"memory_proposal_projection_only",
		"host_must_review_memory_proposal",
		"metadata_presence_not_capability",
		"display_safe_refs_only",
		"no_skill_write",
		"no_workflow_write",
		"no_template_write",
		"no_install_or_reload",
		"no_core_mutation",
		"no_runner_dispatch",
	}, extra)
}

func memoryProposalItemBoundaries(extra []Boundary) []Boundary {
	return MergeBoundaries([]Boundary{
		"memory_proposal_item",
		"proposal_only",
		"host_must_review_memory_proposal",
		"no_skill_write",
		"no_workflow_write",
		"no_template_write",
	}, extra)
}
