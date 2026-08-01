package controlcontract

type AdapterMetadataStrategyContribution struct {
	ContractVersion    string                    `json:"contract_version,omitempty"`
	ContributionRef    DisplaySafeRef            `json:"contribution_ref,omitempty"`
	OwnerRef           DisplaySafeRef            `json:"owner_ref,omitempty"`
	ProviderRef        DisplaySafeRef            `json:"provider_ref,omitempty"`
	StrategyVersionRef DisplaySafeRef            `json:"strategy_version_ref,omitempty"`
	StrategyDigestRef  DisplaySafeRef            `json:"strategy_digest_ref,omitempty"`
	SourceKind         StrategyCatalogSourceKind `json:"source_kind,omitempty"`
	SourceRef          DisplaySafeRef            `json:"source_ref,omitempty"`
	Candidate          StrategyCandidate         `json:"candidate,omitempty"`
	ProvenanceRefs     []DisplaySafeRef          `json:"provenance_refs,omitempty"`
	EvidenceRefs       []EvidenceRef             `json:"evidence_refs,omitempty"`
	PolicyRefs         []DisplaySafeRef          `json:"policy_refs,omitempty"`
	MissingInputs      []MissingInput            `json:"missing_inputs,omitempty"`
	Boundaries         []Boundary                `json:"boundaries,omitempty"`
	RawOutputLoaded    bool                      `json:"raw_output_loaded"`
}

type AdapterMetadataRegistrySnapshotInput struct {
	RegistrySnapshotRef  DisplaySafeRef                        `json:"registry_snapshot_ref,omitempty"`
	StrategyCatalogRef   DisplaySafeRef                        `json:"strategy_catalog_ref,omitempty"`
	OwnerRef             DisplaySafeRef                        `json:"owner_ref,omitempty"`
	ProviderRef          DisplaySafeRef                        `json:"provider_ref,omitempty"`
	RegistryVersionRef   DisplaySafeRef                        `json:"registry_version_ref,omitempty"`
	RegistryDigestRef    DisplaySafeRef                        `json:"registry_digest_ref,omitempty"`
	MaxContributionCount int                                   `json:"max_contribution_count,omitempty"`
	HostPolicyRefs       []DisplaySafeRef                      `json:"host_policy_refs,omitempty"`
	Contributions        []AdapterMetadataStrategyContribution `json:"contributions,omitempty"`
	Boundaries           []Boundary                            `json:"boundaries,omitempty"`
	RawOutputLoaded      bool                                  `json:"raw_output_loaded"`
}

type AdapterMetadataRegistrySnapshot struct {
	ContractVersion         string                                `json:"contract_version,omitempty"`
	Projected               bool                                  `json:"projected"`
	Status                  HostActionStatus                      `json:"status,omitempty"`
	ReadyForStrategyCatalog bool                                  `json:"ready_for_strategy_catalog"`
	RegistrySnapshotRef     DisplaySafeRef                        `json:"registry_snapshot_ref,omitempty"`
	StrategyCatalogRef      DisplaySafeRef                        `json:"strategy_catalog_ref,omitempty"`
	OwnerRef                DisplaySafeRef                        `json:"owner_ref,omitempty"`
	ProviderRef             DisplaySafeRef                        `json:"provider_ref,omitempty"`
	RegistryVersionRef      DisplaySafeRef                        `json:"registry_version_ref,omitempty"`
	RegistryDigestRef       DisplaySafeRef                        `json:"registry_digest_ref,omitempty"`
	MaxContributionCount    int                                   `json:"max_contribution_count,omitempty"`
	ContributionCount       int                                   `json:"contribution_count,omitempty"`
	ReadyContributionCount  int                                   `json:"ready_contribution_count,omitempty"`
	ContributionRefs        []DisplaySafeRef                      `json:"contribution_refs,omitempty"`
	OwnerRefs               []DisplaySafeRef                      `json:"owner_refs,omitempty"`
	ProviderRefs            []DisplaySafeRef                      `json:"provider_refs,omitempty"`
	SourceRefs              []DisplaySafeRef                      `json:"source_refs,omitempty"`
	CandidateRefs           []DisplaySafeRef                      `json:"candidate_refs,omitempty"`
	CapabilityRefs          []DisplaySafeRef                      `json:"capability_refs,omitempty"`
	ProvenanceRefs          []DisplaySafeRef                      `json:"provenance_refs,omitempty"`
	EvidenceRefs            []EvidenceRef                         `json:"evidence_refs,omitempty"`
	PolicyRefs              []DisplaySafeRef                      `json:"policy_refs,omitempty"`
	Contributions           []AdapterMetadataStrategyContribution `json:"contributions,omitempty"`
	StrategyCatalog         StrategyCatalogSnapshot               `json:"strategy_catalog,omitempty"`
	MissingInputs           []MissingInput                        `json:"missing_inputs,omitempty"`
	BlockedReasons          []string                              `json:"blocked_reasons,omitempty"`
	FailureClass            FailureClass                          `json:"failure_class,omitempty"`
	Boundaries              []Boundary                            `json:"boundaries,omitempty"`
	NextHostAction          NextHostAction                        `json:"next_host_action,omitempty"`
	RunnerEffect            string                                `json:"runner_effect,omitempty"`
	PromptEffect            string                                `json:"prompt_effect,omitempty"`
	RawOutputLoaded         bool                                  `json:"raw_output_loaded"`
}

func BuildAdapterMetadataRegistrySnapshot(input AdapterMetadataRegistrySnapshotInput) AdapterMetadataRegistrySnapshot {
	contributions := normalizeAdapterMetadataStrategyContributions(input.Contributions)
	result := AdapterMetadataRegistrySnapshot{
		ContractVersion:      ContractVersion,
		Projected:            true,
		Status:               HostActionBlocked,
		RegistrySnapshotRef:  normalizeOneDisplaySafeRef(input.RegistrySnapshotRef),
		StrategyCatalogRef:   normalizeOneDisplaySafeRef(input.StrategyCatalogRef),
		OwnerRef:             normalizeOneDisplaySafeRef(input.OwnerRef),
		ProviderRef:          normalizeOneDisplaySafeRef(input.ProviderRef),
		RegistryVersionRef:   normalizeOneDisplaySafeRef(input.RegistryVersionRef),
		RegistryDigestRef:    normalizeOneDisplaySafeRef(input.RegistryDigestRef),
		MaxContributionCount: maxNonNegativeInt(input.MaxContributionCount),
		Contributions:        contributions,
		FailureClass:         FailureNone,
		PolicyRefs:           normalizeDisplaySafeRefs(input.HostPolicyRefs),
		Boundaries:           adapterMetadataRegistryBoundaries(input.Boundaries),
		RunnerEffect:         "none",
		PromptEffect:         "none",
		RawOutputLoaded:      input.RawOutputLoaded,
	}
	result = adapterMetadataRegistryAggregate(result, contributions)
	if adapterMetadataRegistryInputUnsafe(input) {
		result = adapterMetadataRegistryBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if result.RegistrySnapshotRef == "" {
		result = adapterMetadataRegistryBlock(result, FailureEvidenceMissing, "adapter_metadata_registry_ref_missing", "host:adapter_metadata_registry_snapshot_ref", "provide_adapter_metadata_registry")
	}
	if result.StrategyCatalogRef == "" {
		result = adapterMetadataRegistryBlock(result, FailureEvidenceMissing, "strategy_catalog_ref_missing", "host:strategy_catalog_ref", "provide_strategy_catalog_ref")
	}
	if result.OwnerRef == "" {
		result = adapterMetadataRegistryBlock(result, FailureConfigMissing, "adapter_metadata_registry_owner_missing", "host:adapter_metadata_registry_owner", "provide_adapter_metadata_registry_owner")
	}
	if result.ProviderRef == "" {
		result = adapterMetadataRegistryBlock(result, FailureConfigMissing, "adapter_metadata_registry_provider_missing", "host:adapter_metadata_registry_provider", "provide_adapter_metadata_registry_provider")
	}
	if result.RegistryVersionRef == "" {
		result = adapterMetadataRegistryBlock(result, FailureEvidenceMissing, "adapter_metadata_registry_version_missing", "host:adapter_metadata_registry_version_ref", "provide_adapter_metadata_registry_version")
	}
	if result.RegistryDigestRef == "" {
		result = adapterMetadataRegistryBlock(result, FailureEvidenceMissing, "adapter_metadata_registry_digest_missing", "host:adapter_metadata_registry_digest_ref", "provide_adapter_metadata_registry_digest")
	}
	if result.MaxContributionCount <= 0 {
		result = adapterMetadataRegistryBlock(result, FailureConfigMissing, "adapter_metadata_contribution_limit_missing", "host:adapter_metadata_contribution_limit", "provide_adapter_metadata_contribution_limit")
	} else if len(contributions) > result.MaxContributionCount {
		result = adapterMetadataRegistryBlock(result, FailurePolicyBlocked, "adapter_metadata_contribution_limit_exceeded", "host:adapter_metadata_contribution_limit", "review_adapter_metadata_registry")
	}
	if len(contributions) == 0 {
		result = adapterMetadataRegistryBlock(result, FailureHostAdapterMissing, "adapter_metadata_contributions_empty", "host:adapter_metadata_contributions", "provide_adapter_metadata_contributions")
	}
	seenContributionRefs := map[DisplaySafeRef]struct{}{}
	seenCandidateRefs := map[DisplaySafeRef]struct{}{}
	for _, contribution := range contributions {
		if contribution.ContributionRef != "" {
			if _, exists := seenContributionRefs[contribution.ContributionRef]; exists {
				result = adapterMetadataRegistryBlock(result, FailureInvalidInput, "adapter_metadata_duplicate_contribution_ref", "host:adapter_metadata_unique_contribution_ref", "review_adapter_metadata_registry")
			}
			seenContributionRefs[contribution.ContributionRef] = struct{}{}
		}
		candidateRef, candidateRefOK := NormalizeDisplaySafeRef(contribution.Candidate.ID)
		if candidateRefOK {
			if _, exists := seenCandidateRefs[candidateRef]; exists {
				result = adapterMetadataRegistryBlock(result, FailureInvalidInput, "adapter_metadata_duplicate_strategy_ref", "host:adapter_metadata_unique_strategy_ref", "review_adapter_metadata_registry")
			}
			seenCandidateRefs[candidateRef] = struct{}{}
		}
		for _, check := range adapterMetadataContributionCompletenessChecks(contribution) {
			if check.missing {
				result = adapterMetadataRegistryBlock(result, check.failure, check.reason, check.input, "provide_adapter_metadata_contribution")
			}
		}
	}
	if len(result.BlockedReasons) == 0 && len(result.MissingInputs) == 0 {
		result.Status = HostActionReady
		result.ReadyForStrategyCatalog = true
		result.NextHostAction = "host_may_review_strategy_catalog"
		result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_strategy_catalog_projection")
	}
	result.StrategyCatalog = BuildStrategyCatalogSnapshotFromAdapterMetadataRegistry(result)
	return result.Normalize()
}

func BuildStrategyCatalogSnapshotFromAdapterMetadataRegistry(snapshot AdapterMetadataRegistrySnapshot) StrategyCatalogSnapshot {
	normalized := snapshot.Normalize()
	entries := []StrategyCatalogEntry{}
	if normalized.ReadyForStrategyCatalog {
		entries = make([]StrategyCatalogEntry, 0, len(normalized.Contributions))
		for _, contribution := range normalized.Contributions {
			entries = append(entries, StrategyCatalogEntry{
				SourceKind:      contribution.SourceKind,
				SourceRef:       contribution.SourceRef,
				Candidate:       contribution.Candidate,
				Status:          VerificationSatisfied,
				EvidenceRefs:    MergeEvidenceRefs(contribution.EvidenceRefs, contribution.Candidate.ExpectedEvidence),
				MissingInputs:   contribution.MissingInputs,
				Boundaries:      adapterMetadataStrategyCatalogEntryBoundaries(contribution.Boundaries),
				RawOutputLoaded: contribution.RawOutputLoaded,
			})
		}
	}
	return StrategyCatalogSnapshot{
		CatalogRef:      normalized.StrategyCatalogRef,
		Entries:         entries,
		SourceRefs:      cloneDisplaySafeRefs(normalized.SourceRefs),
		PolicyRefs:      cloneDisplaySafeRefs(normalized.PolicyRefs),
		EvidenceRefs:    cloneEvidenceRefs(normalized.EvidenceRefs),
		MissingInputs:   cloneMissingInputs(normalized.MissingInputs),
		Boundaries:      adapterMetadataStrategyCatalogBoundaries(normalized.Boundaries),
		RawOutputLoaded: normalized.RawOutputLoaded,
	}.Normalize()
}

func CloneAdapterMetadataStrategyContribution(in AdapterMetadataStrategyContribution) AdapterMetadataStrategyContribution {
	out := in
	out.Candidate = in.Candidate.Clone()
	out.ProvenanceRefs = cloneDisplaySafeRefs(in.ProvenanceRefs)
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.PolicyRefs = cloneDisplaySafeRefs(in.PolicyRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (c AdapterMetadataStrategyContribution) Clone() AdapterMetadataStrategyContribution {
	return CloneAdapterMetadataStrategyContribution(c)
}

func (c AdapterMetadataStrategyContribution) Normalize() AdapterMetadataStrategyContribution {
	out := CloneAdapterMetadataStrategyContribution(c)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.ContributionRef = normalizeOneDisplaySafeRef(out.ContributionRef)
	out.OwnerRef = normalizeOneDisplaySafeRef(out.OwnerRef)
	out.ProviderRef = normalizeOneDisplaySafeRef(out.ProviderRef)
	out.StrategyVersionRef = normalizeOneDisplaySafeRef(out.StrategyVersionRef)
	out.StrategyDigestRef = normalizeOneDisplaySafeRef(out.StrategyDigestRef)
	out.SourceKind = NormalizeStrategyCatalogSourceKind(string(out.SourceKind))
	out.SourceRef = normalizeOneDisplaySafeRef(out.SourceRef)
	out.Candidate = out.Candidate.Normalize()
	out.ProvenanceRefs = normalizeDisplaySafeRefs(out.ProvenanceRefs)
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
	out.PolicyRefs = normalizeDisplaySafeRefs(out.PolicyRefs)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.Boundaries = adapterMetadataContributionBoundaries(out.Boundaries)
	return out
}

func CloneAdapterMetadataRegistrySnapshot(in AdapterMetadataRegistrySnapshot) AdapterMetadataRegistrySnapshot {
	out := in
	out.ContributionRefs = cloneDisplaySafeRefs(in.ContributionRefs)
	out.OwnerRefs = cloneDisplaySafeRefs(in.OwnerRefs)
	out.ProviderRefs = cloneDisplaySafeRefs(in.ProviderRefs)
	out.SourceRefs = cloneDisplaySafeRefs(in.SourceRefs)
	out.CandidateRefs = cloneDisplaySafeRefs(in.CandidateRefs)
	out.CapabilityRefs = cloneDisplaySafeRefs(in.CapabilityRefs)
	out.ProvenanceRefs = cloneDisplaySafeRefs(in.ProvenanceRefs)
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.PolicyRefs = cloneDisplaySafeRefs(in.PolicyRefs)
	out.Contributions = cloneAdapterMetadataStrategyContributions(in.Contributions)
	out.StrategyCatalog = in.StrategyCatalog.Clone()
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (s AdapterMetadataRegistrySnapshot) Clone() AdapterMetadataRegistrySnapshot {
	return CloneAdapterMetadataRegistrySnapshot(s)
}

func (s AdapterMetadataRegistrySnapshot) Normalize() AdapterMetadataRegistrySnapshot {
	out := CloneAdapterMetadataRegistrySnapshot(s)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.RegistrySnapshotRef = normalizeOneDisplaySafeRef(out.RegistrySnapshotRef)
	out.StrategyCatalogRef = normalizeOneDisplaySafeRef(out.StrategyCatalogRef)
	out.OwnerRef = normalizeOneDisplaySafeRef(out.OwnerRef)
	out.ProviderRef = normalizeOneDisplaySafeRef(out.ProviderRef)
	out.RegistryVersionRef = normalizeOneDisplaySafeRef(out.RegistryVersionRef)
	out.RegistryDigestRef = normalizeOneDisplaySafeRef(out.RegistryDigestRef)
	out.MaxContributionCount = maxNonNegativeInt(out.MaxContributionCount)
	out.Contributions = normalizeAdapterMetadataStrategyContributions(out.Contributions)
	out.ContributionCount = len(out.Contributions)
	out.ReadyContributionCount = adapterMetadataReadyContributionCount(out.Contributions)
	out.ContributionRefs = normalizeDisplaySafeRefs(out.ContributionRefs)
	out.OwnerRefs = normalizeDisplaySafeRefs(out.OwnerRefs)
	out.ProviderRefs = normalizeDisplaySafeRefs(out.ProviderRefs)
	out.SourceRefs = normalizeDisplaySafeRefs(out.SourceRefs)
	out.CandidateRefs = normalizeDisplaySafeRefs(out.CandidateRefs)
	out.CapabilityRefs = normalizeDisplaySafeRefs(out.CapabilityRefs)
	out.ProvenanceRefs = normalizeDisplaySafeRefs(out.ProvenanceRefs)
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
	out.PolicyRefs = normalizeDisplaySafeRefs(out.PolicyRefs)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.BlockedReasons = normalizeControlTokenList(out.BlockedReasons)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	if out.RunnerEffect == "" {
		out.RunnerEffect = "none"
	}
	if out.PromptEffect == "" {
		out.PromptEffect = "none"
	}
	if out.RawOutputLoaded {
		out.Status = HostActionReviewRequired
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
	out.ReadyForStrategyCatalog = out.Status == HostActionReady &&
		out.RegistrySnapshotRef != "" &&
		out.StrategyCatalogRef != "" &&
		out.OwnerRef != "" &&
		out.ProviderRef != "" &&
		out.RegistryVersionRef != "" &&
		out.RegistryDigestRef != "" &&
		out.MaxContributionCount > 0 &&
		out.ContributionCount > 0 &&
		out.ReadyContributionCount == out.ContributionCount &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.RawOutputLoaded
	if !strategyCatalogSnapshotEmpty(out.StrategyCatalog) {
		out.StrategyCatalog = out.StrategyCatalog.Normalize()
	}
	return out
}

func adapterMetadataRegistryAggregate(result AdapterMetadataRegistrySnapshot, contributions []AdapterMetadataStrategyContribution) AdapterMetadataRegistrySnapshot {
	result.ContributionCount = len(contributions)
	for _, contribution := range contributions {
		result.ContributionRefs = appendUniqueDisplaySafeRef(result.ContributionRefs, contribution.ContributionRef)
		result.OwnerRefs = appendUniqueDisplaySafeRef(result.OwnerRefs, contribution.OwnerRef)
		result.ProviderRefs = appendUniqueDisplaySafeRef(result.ProviderRefs, contribution.ProviderRef)
		result.SourceRefs = appendUniqueDisplaySafeRef(result.SourceRefs, contribution.SourceRef)
		if candidateRef, ok := NormalizeDisplaySafeRef(contribution.Candidate.ID); ok {
			result.CandidateRefs = appendUniqueDisplaySafeRef(result.CandidateRefs, candidateRef)
		}
		for _, ref := range contribution.Candidate.CapabilityRefs {
			result.CapabilityRefs = appendUniqueDisplaySafeRef(result.CapabilityRefs, ref)
		}
		for _, ref := range contribution.ProvenanceRefs {
			result.ProvenanceRefs = appendUniqueDisplaySafeRef(result.ProvenanceRefs, ref)
		}
		result.EvidenceRefs = MergeEvidenceRefs(result.EvidenceRefs, contribution.EvidenceRefs, contribution.Candidate.ExpectedEvidence)
		for _, ref := range contribution.PolicyRefs {
			result.PolicyRefs = appendUniqueDisplaySafeRef(result.PolicyRefs, ref)
		}
		result.MissingInputs = AppendMissingInputs(result.MissingInputs, contribution.MissingInputs...)
		result.Boundaries = MergeBoundaries(result.Boundaries, contribution.Boundaries)
		result.RawOutputLoaded = result.RawOutputLoaded || contribution.RawOutputLoaded
	}
	result.ProviderRefs = appendUniqueDisplaySafeRef(result.ProviderRefs, result.ProviderRef)
	result.OwnerRefs = appendUniqueDisplaySafeRef(result.OwnerRefs, result.OwnerRef)
	result.PolicyRefs = normalizeDisplaySafeRefs(result.PolicyRefs)
	result.ContributionRefs = normalizeDisplaySafeRefs(result.ContributionRefs)
	result.SourceRefs = normalizeDisplaySafeRefs(result.SourceRefs)
	result.CandidateRefs = normalizeDisplaySafeRefs(result.CandidateRefs)
	result.CapabilityRefs = normalizeDisplaySafeRefs(result.CapabilityRefs)
	result.ProvenanceRefs = normalizeDisplaySafeRefs(result.ProvenanceRefs)
	result.EvidenceRefs = normalizeEvidenceRefs(result.EvidenceRefs)
	return result
}

func adapterMetadataRegistryBlock(result AdapterMetadataRegistrySnapshot, failure FailureClass, reason string, missing MissingInput, next NextHostAction) AdapterMetadataRegistrySnapshot {
	result.Status = HostActionBlocked
	result.ReadyForStrategyCatalog = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	return result
}

type adapterMetadataContributionCompletenessCheck struct {
	missing bool
	failure FailureClass
	reason  string
	input   MissingInput
}

func adapterMetadataContributionCompletenessChecks(contribution AdapterMetadataStrategyContribution) []adapterMetadataContributionCompletenessCheck {
	candidateRef, candidateRefOK := NormalizeDisplaySafeRef(contribution.Candidate.ID)
	return []adapterMetadataContributionCompletenessCheck{
		{contribution.ContributionRef == "", FailureEvidenceMissing, "adapter_metadata_contribution_ref_missing", "host:adapter_metadata_contribution_ref"},
		{contribution.OwnerRef == "", FailureConfigMissing, "adapter_metadata_contribution_owner_missing", "host:adapter_metadata_contribution_owner"},
		{contribution.StrategyVersionRef == "", FailureEvidenceMissing, "adapter_metadata_strategy_version_missing", "host:adapter_metadata_strategy_version_ref"},
		{contribution.StrategyDigestRef == "", FailureEvidenceMissing, "adapter_metadata_strategy_digest_missing", "host:adapter_metadata_strategy_digest_ref"},
		{contribution.SourceKind == "", FailureInvalidInput, "adapter_metadata_source_kind_missing", "host:adapter_metadata_source_kind"},
		{contribution.SourceRef == "", FailureEvidenceMissing, "adapter_metadata_source_ref_missing", "host:adapter_metadata_source_ref"},
		{contribution.Candidate.ID == "", FailureEvidenceMissing, "adapter_metadata_strategy_ref_missing", "host:adapter_metadata_strategy_ref"},
		{contribution.Candidate.ID != "" && !candidateRefOK, FailureEvidenceWeak, "adapter_metadata_strategy_ref_unsafe", "host:adapter_metadata_strategy_ref"},
		{candidateRefOK && candidateRef == "", FailureEvidenceWeak, "adapter_metadata_strategy_ref_unsafe", "host:adapter_metadata_strategy_ref"},
		{contribution.Candidate.ControlMode == "", FailureInvalidInput, "adapter_metadata_control_mode_missing", "host:adapter_metadata_control_mode"},
		{contribution.Candidate.MinIntensity == "", FailureInvalidInput, "adapter_metadata_min_intensity_missing", "host:adapter_metadata_min_intensity"},
		{contribution.Candidate.Owner == "", FailureConfigMissing, "adapter_metadata_strategy_owner_missing", "host:adapter_metadata_strategy_owner"},
	}
}

func adapterMetadataRegistryInputUnsafe(input AdapterMetadataRegistrySnapshotInput) bool {
	if input.RawOutputLoaded ||
		displaySafeRefRejected(input.RegistrySnapshotRef) ||
		displaySafeRefRejected(input.StrategyCatalogRef) ||
		displaySafeRefRejected(input.OwnerRef) ||
		displaySafeRefRejected(input.ProviderRef) ||
		displaySafeRefRejected(input.RegistryVersionRef) ||
		displaySafeRefRejected(input.RegistryDigestRef) ||
		displaySafeRefSliceRejected(input.HostPolicyRefs) {
		return true
	}
	for _, contribution := range input.Contributions {
		if adapterMetadataContributionUnsafe(contribution) {
			return true
		}
	}
	return false
}

func adapterMetadataContributionUnsafe(contribution AdapterMetadataStrategyContribution) bool {
	return contribution.RawOutputLoaded ||
		displaySafeRefRejected(contribution.ContributionRef) ||
		displaySafeRefRejected(contribution.OwnerRef) ||
		displaySafeRefRejected(contribution.ProviderRef) ||
		displaySafeRefRejected(contribution.StrategyVersionRef) ||
		displaySafeRefRejected(contribution.StrategyDigestRef) ||
		displaySafeRefRejected(contribution.SourceRef) ||
		(contribution.Candidate.ID != "" && displaySafeRefRejected(DisplaySafeRef(contribution.Candidate.ID))) ||
		displaySafeRefSliceRejected(contribution.Candidate.CapabilityRefs) ||
		displaySafeRefSliceRejected(contribution.ProvenanceRefs) ||
		displaySafeRefSliceRejected(contribution.PolicyRefs) ||
		evidenceRefRejected(contribution.Candidate.ExpectedEvidence) ||
		evidenceRefRejected(contribution.EvidenceRefs)
}

func adapterMetadataReadyContributionCount(contributions []AdapterMetadataStrategyContribution) int {
	count := 0
	for _, contribution := range contributions {
		if len(contribution.MissingInputs) == 0 && !contribution.RawOutputLoaded {
			ready := true
			for _, check := range adapterMetadataContributionCompletenessChecks(contribution) {
				if check.missing {
					ready = false
					break
				}
			}
			if ready {
				count++
			}
		}
	}
	return count
}

func normalizeAdapterMetadataStrategyContributions(in []AdapterMetadataStrategyContribution) []AdapterMetadataStrategyContribution {
	out := make([]AdapterMetadataStrategyContribution, 0, len(in))
	for _, contribution := range in {
		normalized := contribution.Normalize()
		out = append(out, normalized)
	}
	return out
}

func cloneAdapterMetadataStrategyContributions(in []AdapterMetadataStrategyContribution) []AdapterMetadataStrategyContribution {
	if len(in) == 0 {
		return nil
	}
	out := make([]AdapterMetadataStrategyContribution, len(in))
	for i, contribution := range in {
		out[i] = contribution.Clone()
	}
	return out
}

func adapterMetadataRegistryBoundaries(extra []Boundary) []Boundary {
	return MergeBoundaries([]Boundary{
		"adapter_metadata_registry_snapshot",
		"adapter_metadata_registry_projection_only",
		"host_owned_adapter_metadata",
		"project_owned_strategy_metadata_allowed",
		"metadata_presence_not_capability",
		"strategy_catalog_projection_only",
		"display_safe_refs_only",
		"no_adapter_invocation",
		"no_runner_dispatch",
		"no_skill_write",
		"no_workflow_write",
		"no_core_mutation",
	}, extra)
}

func adapterMetadataContributionBoundaries(extra []Boundary) []Boundary {
	return MergeBoundaries([]Boundary{
		"adapter_metadata_strategy_contribution",
		"strategy_metadata_only",
		"metadata_presence_not_capability",
		"display_safe_refs_only",
		"no_adapter_invocation",
	}, extra)
}

func adapterMetadataStrategyCatalogBoundaries(extra []Boundary) []Boundary {
	return MergeBoundaries([]Boundary{
		"adapter_metadata_to_strategy_catalog_projection",
		"strategy_catalog_projection_only",
		"metadata_presence_not_capability",
		"no_adapter_invocation",
		"no_runner_dispatch",
		"no_skill_write",
		"no_workflow_write",
	}, extra)
}

func adapterMetadataStrategyCatalogEntryBoundaries(extra []Boundary) []Boundary {
	return MergeBoundaries([]Boundary{
		"adapter_metadata_strategy_catalog_entry",
		"strategy_metadata_only",
		"metadata_presence_not_capability",
		"no_adapter_invocation",
	}, extra)
}

func strategyCatalogSnapshotEmpty(snapshot StrategyCatalogSnapshot) bool {
	return snapshot.CatalogRef == "" &&
		len(snapshot.Entries) == 0 &&
		len(snapshot.SourceRefs) == 0 &&
		len(snapshot.PolicyRefs) == 0 &&
		len(snapshot.EvidenceRefs) == 0 &&
		len(snapshot.MissingInputs) == 0 &&
		len(snapshot.Boundaries) == 0 &&
		!snapshot.RawOutputLoaded
}
