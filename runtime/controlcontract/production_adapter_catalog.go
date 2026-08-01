package controlcontract

type ProductionAdapterCatalogSnapshotInput struct {
	CatalogSnapshotRef DisplaySafeRef                `json:"catalog_snapshot_ref,omitempty"`
	Producer           DisplaySafeRef                `json:"producer,omitempty"`
	ProviderRef        DisplaySafeRef                `json:"provider_ref,omitempty"`
	CatalogVersionRef  DisplaySafeRef                `json:"catalog_version_ref,omitempty"`
	CatalogDigestRef   DisplaySafeRef                `json:"catalog_digest_ref,omitempty"`
	MaxDescriptorCount int                           `json:"max_descriptor_count,omitempty"`
	HostPolicyRefs     []DisplaySafeRef              `json:"host_policy_refs,omitempty"`
	Descriptors        []ProductionAdapterDescriptor `json:"descriptors,omitempty"`
	Boundaries         []Boundary                    `json:"boundaries,omitempty"`
	RawOutputLoaded    bool                          `json:"raw_output_loaded"`
}

type ProductionAdapterCatalogSnapshot struct {
	ContractVersion        string                        `json:"contract_version,omitempty"`
	Projected              bool                          `json:"projected"`
	Available              bool                          `json:"available"`
	Status                 HostActionStatus              `json:"status,omitempty"`
	ReadyForHostSelection  bool                          `json:"ready_for_host_selection"`
	CatalogSnapshotRef     DisplaySafeRef                `json:"catalog_snapshot_ref,omitempty"`
	Producer               DisplaySafeRef                `json:"producer,omitempty"`
	ProviderRef            DisplaySafeRef                `json:"provider_ref,omitempty"`
	CatalogVersionRef      DisplaySafeRef                `json:"catalog_version_ref,omitempty"`
	CatalogDigestRef       DisplaySafeRef                `json:"catalog_digest_ref,omitempty"`
	MaxDescriptorCount     int                           `json:"max_descriptor_count,omitempty"`
	DescriptorCount        int                           `json:"descriptor_count,omitempty"`
	ReadyDescriptorCount   int                           `json:"ready_descriptor_count,omitempty"`
	DescriptorRefs         []DisplaySafeRef              `json:"descriptor_refs,omitempty"`
	OwnerRefs              []DisplaySafeRef              `json:"owner_refs,omitempty"`
	Kinds                  []ProductionAdapterKind       `json:"kinds,omitempty"`
	SourceKinds            []ReplannerSourceKind         `json:"source_kinds,omitempty"`
	CandidateRefs          []DisplaySafeRef              `json:"candidate_refs,omitempty"`
	ProvidesCapabilityRefs []DisplaySafeRef              `json:"provides_capability_refs,omitempty"`
	RequiresCapabilityRefs []DisplaySafeRef              `json:"requires_capability_refs,omitempty"`
	PolicyRefs             []DisplaySafeRef              `json:"policy_refs,omitempty"`
	ApprovalRefs           []DisplaySafeRef              `json:"approval_refs,omitempty"`
	BudgetRefs             []DisplaySafeRef              `json:"budget_refs,omitempty"`
	PreflightRefs          []DisplaySafeRef              `json:"preflight_refs,omitempty"`
	Descriptors            []ProductionAdapterDescriptor `json:"descriptors,omitempty"`
	MissingInputs          []MissingInput                `json:"missing_inputs,omitempty"`
	BlockedReasons         []string                      `json:"blocked_reasons,omitempty"`
	FailureClass           FailureClass                  `json:"failure_class,omitempty"`
	Boundaries             []Boundary                    `json:"boundaries,omitempty"`
	NextHostAction         NextHostAction                `json:"next_host_action,omitempty"`
	RunnerEffect           string                        `json:"runner_effect,omitempty"`
	PromptEffect           string                        `json:"prompt_effect,omitempty"`
	RawOutputLoaded        bool                          `json:"raw_output_loaded"`
}

type ProductionAdapterCatalogDiscoveryView struct {
	ContractVersion        string                  `json:"contract_version,omitempty"`
	Projected              bool                    `json:"projected"`
	Available              bool                    `json:"available"`
	Status                 string                  `json:"status,omitempty"`
	Mode                   string                  `json:"mode,omitempty"`
	RunnerEffect           string                  `json:"runner_effect,omitempty"`
	PromptEffect           string                  `json:"prompt_effect,omitempty"`
	ReadyForHostSelection  bool                    `json:"ready_for_host_selection"`
	CatalogSnapshotRef     DisplaySafeRef          `json:"catalog_snapshot_ref,omitempty"`
	Producer               DisplaySafeRef          `json:"producer,omitempty"`
	ProviderRef            DisplaySafeRef          `json:"provider_ref,omitempty"`
	CatalogVersionRef      DisplaySafeRef          `json:"catalog_version_ref,omitempty"`
	CatalogDigestRef       DisplaySafeRef          `json:"catalog_digest_ref,omitempty"`
	MaxDescriptorCount     int                     `json:"max_descriptor_count,omitempty"`
	DescriptorCount        int                     `json:"descriptor_count,omitempty"`
	ReadyDescriptorCount   int                     `json:"ready_descriptor_count,omitempty"`
	DescriptorRefs         []DisplaySafeRef        `json:"descriptor_refs,omitempty"`
	OwnerRefs              []DisplaySafeRef        `json:"owner_refs,omitempty"`
	Kinds                  []ProductionAdapterKind `json:"kinds,omitempty"`
	SourceKinds            []ReplannerSourceKind   `json:"source_kinds,omitempty"`
	CandidateRefs          []DisplaySafeRef        `json:"candidate_refs,omitempty"`
	ProvidesCapabilityRefs []DisplaySafeRef        `json:"provides_capability_refs,omitempty"`
	RequiresCapabilityRefs []DisplaySafeRef        `json:"requires_capability_refs,omitempty"`
	PolicyRefs             []DisplaySafeRef        `json:"policy_refs,omitempty"`
	ApprovalRefs           []DisplaySafeRef        `json:"approval_refs,omitempty"`
	BudgetRefs             []DisplaySafeRef        `json:"budget_refs,omitempty"`
	PreflightRefs          []DisplaySafeRef        `json:"preflight_refs,omitempty"`
	MissingInputs          []MissingInput          `json:"missing_inputs,omitempty"`
	BlockedReasons         []string                `json:"blocked_reasons,omitempty"`
	FailureClass           FailureClass            `json:"failure_class,omitempty"`
	Boundaries             []Boundary              `json:"boundaries,omitempty"`
	NextHostAction         NextHostAction          `json:"next_host_action,omitempty"`
	RawOutputLoaded        bool                    `json:"raw_output_loaded"`
}

func BuildProductionAdapterCatalogSnapshot(input ProductionAdapterCatalogSnapshotInput) ProductionAdapterCatalogSnapshot {
	descriptors := normalizeProductionAdapterCatalogDescriptors(input.Descriptors)
	result := ProductionAdapterCatalogSnapshot{
		ContractVersion:    ContractVersion,
		Projected:          true,
		Available:          true,
		Status:             HostActionBlocked,
		CatalogSnapshotRef: normalizeOneDisplaySafeRef(input.CatalogSnapshotRef),
		Producer:           normalizeOneDisplaySafeRef(input.Producer),
		ProviderRef:        normalizeOneDisplaySafeRef(input.ProviderRef),
		CatalogVersionRef:  normalizeOneDisplaySafeRef(input.CatalogVersionRef),
		CatalogDigestRef:   normalizeOneDisplaySafeRef(input.CatalogDigestRef),
		MaxDescriptorCount: maxNonNegativeInt(input.MaxDescriptorCount),
		Descriptors:        descriptors,
		FailureClass:       FailureNone,
		Boundaries:         productionAdapterCatalogSnapshotBoundaries(input.Boundaries),
		RunnerEffect:       "none",
		PromptEffect:       "none",
		RawOutputLoaded:    input.RawOutputLoaded,
	}
	result = productionAdapterCatalogSnapshotAggregate(result, descriptors, input.HostPolicyRefs)
	if productionAdapterCatalogSnapshotInputUnsafe(input) {
		result = productionAdapterCatalogSnapshotBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if result.CatalogSnapshotRef == "" {
		result = productionAdapterCatalogSnapshotBlock(result, FailureEvidenceMissing, "catalog_snapshot_ref_missing", "host:adapter_catalog_snapshot_ref", "provide_adapter_catalog_snapshot")
	}
	if len(descriptors) == 0 {
		result = productionAdapterCatalogSnapshotBlock(result, FailureHostAdapterMissing, "adapter_catalog_empty", "host:adapter_catalog_descriptors", "provide_adapter_catalog_descriptors")
	}
	if result.MaxDescriptorCount > 0 && len(descriptors) > result.MaxDescriptorCount {
		result = productionAdapterCatalogSnapshotBlock(result, FailurePolicyBlocked, "adapter_catalog_descriptor_count_exceeded", "host:adapter_catalog_descriptor_limit", "review_adapter_catalog")
	}
	seen := map[DisplaySafeRef]struct{}{}
	for _, descriptor := range descriptors {
		if descriptor.AdapterRef != "" {
			if _, exists := seen[descriptor.AdapterRef]; exists {
				result = productionAdapterCatalogSnapshotBlock(result, FailureInvalidInput, "adapter_catalog_duplicate_ref", "host:adapter_catalog_unique_ref", "review_adapter_catalog")
			}
			seen[descriptor.AdapterRef] = struct{}{}
		}
		for _, check := range productionAdapterDescriptorCompletenessChecks(descriptor) {
			if check.missing {
				result = productionAdapterCatalogSnapshotBlock(result, check.failure, check.reason, check.input, "provide_adapter_descriptor")
			}
		}
	}
	if len(result.BlockedReasons) == 0 && len(result.MissingInputs) == 0 {
		result.Status = HostActionReady
		result.ReadyForHostSelection = true
		result.NextHostAction = "host_may_select_adapter"
		result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_host_adapter_selection")
	}
	return result.Normalize()
}

func CloneProductionAdapterCatalogSnapshot(in ProductionAdapterCatalogSnapshot) ProductionAdapterCatalogSnapshot {
	out := in
	out.DescriptorRefs = cloneDisplaySafeRefs(in.DescriptorRefs)
	out.OwnerRefs = cloneDisplaySafeRefs(in.OwnerRefs)
	out.Kinds = cloneProductionAdapterKinds(in.Kinds)
	out.SourceKinds = cloneReplannerSourceKinds(in.SourceKinds)
	out.CandidateRefs = cloneDisplaySafeRefs(in.CandidateRefs)
	out.ProvidesCapabilityRefs = cloneDisplaySafeRefs(in.ProvidesCapabilityRefs)
	out.RequiresCapabilityRefs = cloneDisplaySafeRefs(in.RequiresCapabilityRefs)
	out.PolicyRefs = cloneDisplaySafeRefs(in.PolicyRefs)
	out.ApprovalRefs = cloneDisplaySafeRefs(in.ApprovalRefs)
	out.BudgetRefs = cloneDisplaySafeRefs(in.BudgetRefs)
	out.PreflightRefs = cloneDisplaySafeRefs(in.PreflightRefs)
	out.Descriptors = cloneProductionAdapterDescriptors(in.Descriptors)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (s ProductionAdapterCatalogSnapshot) Clone() ProductionAdapterCatalogSnapshot {
	return CloneProductionAdapterCatalogSnapshot(s)
}

func (s ProductionAdapterCatalogSnapshot) Normalize() ProductionAdapterCatalogSnapshot {
	out := CloneProductionAdapterCatalogSnapshot(s)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.CatalogSnapshotRef = normalizeOneDisplaySafeRef(out.CatalogSnapshotRef)
	out.Producer = normalizeOneDisplaySafeRef(out.Producer)
	out.ProviderRef = normalizeOneDisplaySafeRef(out.ProviderRef)
	out.CatalogVersionRef = normalizeOneDisplaySafeRef(out.CatalogVersionRef)
	out.CatalogDigestRef = normalizeOneDisplaySafeRef(out.CatalogDigestRef)
	out.MaxDescriptorCount = maxNonNegativeInt(out.MaxDescriptorCount)
	out.Descriptors = normalizeProductionAdapterCatalogDescriptors(out.Descriptors)
	out.DescriptorCount = len(out.Descriptors)
	out.ReadyDescriptorCount = productionAdapterReadyDescriptorCount(out.Descriptors)
	out.DescriptorRefs = normalizeDisplaySafeRefs(out.DescriptorRefs)
	out.OwnerRefs = normalizeDisplaySafeRefs(out.OwnerRefs)
	out.Kinds = normalizeProductionAdapterKinds(out.Kinds)
	out.SourceKinds = normalizeReplannerSourceKinds(out.SourceKinds)
	out.CandidateRefs = normalizeDisplaySafeRefs(out.CandidateRefs)
	out.ProvidesCapabilityRefs = normalizeDisplaySafeRefs(out.ProvidesCapabilityRefs)
	out.RequiresCapabilityRefs = normalizeDisplaySafeRefs(out.RequiresCapabilityRefs)
	out.PolicyRefs = normalizeDisplaySafeRefs(out.PolicyRefs)
	out.ApprovalRefs = normalizeDisplaySafeRefs(out.ApprovalRefs)
	out.BudgetRefs = normalizeDisplaySafeRefs(out.BudgetRefs)
	out.PreflightRefs = normalizeDisplaySafeRefs(out.PreflightRefs)
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
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		if out.NextHostAction == "" {
			out.NextHostAction = "provide_display_safe_refs"
		}
	}
	out.ReadyForHostSelection = out.Status == HostActionReady && len(out.MissingInputs) == 0 && len(out.BlockedReasons) == 0 && out.DescriptorCount > 0 && !out.RawOutputLoaded
	return out
}

// agentx-api: internal_candidate
func BuildProductionAdapterCatalogDiscoveryView(snapshot ProductionAdapterCatalogSnapshot) ProductionAdapterCatalogDiscoveryView {
	if productionAdapterCatalogSnapshotEmpty(snapshot) {
		return unavailableProductionAdapterCatalogDiscoveryView()
	}
	unsafe := productionAdapterCatalogSnapshotUnsafe(snapshot)
	normalized := snapshot.Normalize()
	result := ProductionAdapterCatalogDiscoveryView{
		ContractVersion:        ContractVersion,
		Projected:              true,
		Available:              true,
		Status:                 "blocked",
		Mode:                   "production_adapter_catalog_discovery_view",
		RunnerEffect:           "none",
		PromptEffect:           "none",
		CatalogSnapshotRef:     normalized.CatalogSnapshotRef,
		Producer:               normalized.Producer,
		ProviderRef:            normalized.ProviderRef,
		CatalogVersionRef:      normalized.CatalogVersionRef,
		CatalogDigestRef:       normalized.CatalogDigestRef,
		MaxDescriptorCount:     normalized.MaxDescriptorCount,
		DescriptorCount:        normalized.DescriptorCount,
		ReadyDescriptorCount:   normalized.ReadyDescriptorCount,
		DescriptorRefs:         cloneDisplaySafeRefs(normalized.DescriptorRefs),
		OwnerRefs:              cloneDisplaySafeRefs(normalized.OwnerRefs),
		Kinds:                  cloneProductionAdapterKinds(normalized.Kinds),
		SourceKinds:            cloneReplannerSourceKinds(normalized.SourceKinds),
		CandidateRefs:          cloneDisplaySafeRefs(normalized.CandidateRefs),
		ProvidesCapabilityRefs: cloneDisplaySafeRefs(normalized.ProvidesCapabilityRefs),
		RequiresCapabilityRefs: cloneDisplaySafeRefs(normalized.RequiresCapabilityRefs),
		PolicyRefs:             cloneDisplaySafeRefs(normalized.PolicyRefs),
		ApprovalRefs:           cloneDisplaySafeRefs(normalized.ApprovalRefs),
		BudgetRefs:             cloneDisplaySafeRefs(normalized.BudgetRefs),
		PreflightRefs:          cloneDisplaySafeRefs(normalized.PreflightRefs),
		MissingInputs:          cloneMissingInputs(normalized.MissingInputs),
		BlockedReasons:         cloneStringSlice(normalized.BlockedReasons),
		FailureClass:           normalized.FailureClass,
		Boundaries:             productionAdapterCatalogDiscoveryViewBoundaries(normalized.Boundaries),
		NextHostAction:         normalized.NextHostAction,
		RawOutputLoaded:        normalized.RawOutputLoaded,
	}
	if unsafe {
		result.FailureClass = FailureEvidenceWeak
		result.MissingInputs = AppendMissingInputs(result.MissingInputs, "host:display_safe_refs")
		result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, "unsafe_input_ref")
		result.NextHostAction = "provide_display_safe_refs"
		return result.Normalize()
	}
	if normalized.ReadyForHostSelection {
		result.Status = "ready_for_host_adapter_selection"
		result.ReadyForHostSelection = true
		result.NextHostAction = firstNextHostAction(normalized.NextHostAction, "host_may_select_adapter")
		result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_host_adapter_selection")
	}
	return result.Normalize()
}

func CloneProductionAdapterCatalogDiscoveryView(in ProductionAdapterCatalogDiscoveryView) ProductionAdapterCatalogDiscoveryView {
	out := in
	out.DescriptorRefs = cloneDisplaySafeRefs(in.DescriptorRefs)
	out.OwnerRefs = cloneDisplaySafeRefs(in.OwnerRefs)
	out.Kinds = cloneProductionAdapterKinds(in.Kinds)
	out.SourceKinds = cloneReplannerSourceKinds(in.SourceKinds)
	out.CandidateRefs = cloneDisplaySafeRefs(in.CandidateRefs)
	out.ProvidesCapabilityRefs = cloneDisplaySafeRefs(in.ProvidesCapabilityRefs)
	out.RequiresCapabilityRefs = cloneDisplaySafeRefs(in.RequiresCapabilityRefs)
	out.PolicyRefs = cloneDisplaySafeRefs(in.PolicyRefs)
	out.ApprovalRefs = cloneDisplaySafeRefs(in.ApprovalRefs)
	out.BudgetRefs = cloneDisplaySafeRefs(in.BudgetRefs)
	out.PreflightRefs = cloneDisplaySafeRefs(in.PreflightRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (v ProductionAdapterCatalogDiscoveryView) Clone() ProductionAdapterCatalogDiscoveryView {
	return CloneProductionAdapterCatalogDiscoveryView(v)
}

func (v ProductionAdapterCatalogDiscoveryView) Normalize() ProductionAdapterCatalogDiscoveryView {
	out := CloneProductionAdapterCatalogDiscoveryView(v)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = normalizeControlToken(out.Status)
	out.Mode = normalizeControlToken(out.Mode)
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	if out.RunnerEffect == "" {
		out.RunnerEffect = "none"
	}
	if out.PromptEffect == "" {
		out.PromptEffect = "none"
	}
	out.CatalogSnapshotRef = normalizeOneDisplaySafeRef(out.CatalogSnapshotRef)
	out.Producer = normalizeOneDisplaySafeRef(out.Producer)
	out.ProviderRef = normalizeOneDisplaySafeRef(out.ProviderRef)
	out.CatalogVersionRef = normalizeOneDisplaySafeRef(out.CatalogVersionRef)
	out.CatalogDigestRef = normalizeOneDisplaySafeRef(out.CatalogDigestRef)
	out.MaxDescriptorCount = maxNonNegativeInt(out.MaxDescriptorCount)
	out.DescriptorRefs = normalizeDisplaySafeRefs(out.DescriptorRefs)
	out.OwnerRefs = normalizeDisplaySafeRefs(out.OwnerRefs)
	out.Kinds = normalizeProductionAdapterKinds(out.Kinds)
	out.SourceKinds = normalizeReplannerSourceKinds(out.SourceKinds)
	out.CandidateRefs = normalizeDisplaySafeRefs(out.CandidateRefs)
	out.ProvidesCapabilityRefs = normalizeDisplaySafeRefs(out.ProvidesCapabilityRefs)
	out.RequiresCapabilityRefs = normalizeDisplaySafeRefs(out.RequiresCapabilityRefs)
	out.PolicyRefs = normalizeDisplaySafeRefs(out.PolicyRefs)
	out.ApprovalRefs = normalizeDisplaySafeRefs(out.ApprovalRefs)
	out.BudgetRefs = normalizeDisplaySafeRefs(out.BudgetRefs)
	out.PreflightRefs = normalizeDisplaySafeRefs(out.PreflightRefs)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.BlockedReasons = normalizeControlTokenList(out.BlockedReasons)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	if out.RawOutputLoaded {
		out.Status = "review_required"
		if out.FailureClass == FailureNone {
			out.FailureClass = FailureEvidenceWeak
		}
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		if out.NextHostAction == "" {
			out.NextHostAction = "provide_display_safe_refs"
		}
	}
	out.ReadyForHostSelection = out.Status == "ready_for_host_adapter_selection" && len(out.MissingInputs) == 0 && len(out.BlockedReasons) == 0 && out.DescriptorCount > 0 && !out.RawOutputLoaded
	return out
}

func productionAdapterCatalogSnapshotBlock(result ProductionAdapterCatalogSnapshot, failure FailureClass, reason string, missing MissingInput, next NextHostAction) ProductionAdapterCatalogSnapshot {
	result.Status = HostActionBlocked
	result.ReadyForHostSelection = false
	result.FailureClass = firstFailureClass(failure, result.FailureClass)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	return result
}

func productionAdapterCatalogSnapshotAggregate(result ProductionAdapterCatalogSnapshot, descriptors []ProductionAdapterDescriptor, hostPolicyRefs []DisplaySafeRef) ProductionAdapterCatalogSnapshot {
	result.PolicyRefs = normalizeDisplaySafeRefs(hostPolicyRefs)
	for _, descriptor := range descriptors {
		result.DescriptorRefs = append(result.DescriptorRefs, descriptor.AdapterRef)
		result.OwnerRefs = append(result.OwnerRefs, descriptor.OwnerRef)
		result.Kinds = append(result.Kinds, descriptor.Kind)
		result.SourceKinds = append(result.SourceKinds, descriptor.SupportedSourceKinds...)
		result.CandidateRefs = append(result.CandidateRefs, descriptor.SupportedCandidateRefs...)
		result.ProvidesCapabilityRefs = append(result.ProvidesCapabilityRefs, descriptor.ProvidesCapabilityRefs...)
		result.RequiresCapabilityRefs = append(result.RequiresCapabilityRefs, descriptor.RequiresCapabilityRefs...)
		result.PolicyRefs = append(result.PolicyRefs, descriptor.RequiredPolicyRefs...)
		result.ApprovalRefs = append(result.ApprovalRefs, descriptor.RequiredApprovalRefs...)
		result.BudgetRefs = append(result.BudgetRefs, descriptor.RequiredBudgetRef)
		result.PreflightRefs = append(result.PreflightRefs, descriptor.PreflightCheckRefs...)
	}
	return result.Normalize()
}

func productionAdapterCatalogSnapshotInputUnsafe(input ProductionAdapterCatalogSnapshotInput) bool {
	if input.RawOutputLoaded ||
		displaySafeRefRejected(input.CatalogSnapshotRef) ||
		displaySafeRefRejected(input.Producer) ||
		displaySafeRefRejected(input.ProviderRef) ||
		displaySafeRefRejected(input.CatalogVersionRef) ||
		displaySafeRefRejected(input.CatalogDigestRef) ||
		displaySafeRefSliceRejected(input.HostPolicyRefs) {
		return true
	}
	for _, descriptor := range input.Descriptors {
		if productionAdapterDescriptorUnsafe(descriptor) {
			return true
		}
	}
	return false
}

func productionAdapterCatalogSnapshotUnsafe(snapshot ProductionAdapterCatalogSnapshot) bool {
	if snapshot.RawOutputLoaded ||
		displaySafeRefRejected(snapshot.CatalogSnapshotRef) ||
		displaySafeRefRejected(snapshot.Producer) ||
		displaySafeRefRejected(snapshot.ProviderRef) ||
		displaySafeRefRejected(snapshot.CatalogVersionRef) ||
		displaySafeRefRejected(snapshot.CatalogDigestRef) ||
		displaySafeRefSliceRejected(snapshot.DescriptorRefs) ||
		displaySafeRefSliceRejected(snapshot.OwnerRefs) ||
		displaySafeRefSliceRejected(snapshot.CandidateRefs) ||
		displaySafeRefSliceRejected(snapshot.ProvidesCapabilityRefs) ||
		displaySafeRefSliceRejected(snapshot.RequiresCapabilityRefs) ||
		displaySafeRefSliceRejected(snapshot.PolicyRefs) ||
		displaySafeRefSliceRejected(snapshot.ApprovalRefs) ||
		displaySafeRefSliceRejected(snapshot.BudgetRefs) ||
		displaySafeRefSliceRejected(snapshot.PreflightRefs) {
		return true
	}
	for _, descriptor := range snapshot.Descriptors {
		if productionAdapterDescriptorUnsafe(descriptor) {
			return true
		}
	}
	return false
}

func productionAdapterCatalogSnapshotEmpty(snapshot ProductionAdapterCatalogSnapshot) bool {
	return !snapshot.Projected &&
		!snapshot.Available &&
		snapshot.Status == "" &&
		snapshot.CatalogSnapshotRef == "" &&
		snapshot.Producer == "" &&
		snapshot.ProviderRef == "" &&
		snapshot.CatalogVersionRef == "" &&
		snapshot.CatalogDigestRef == "" &&
		snapshot.MaxDescriptorCount == 0 &&
		len(snapshot.Descriptors) == 0 &&
		len(snapshot.DescriptorRefs) == 0 &&
		len(snapshot.MissingInputs) == 0 &&
		len(snapshot.BlockedReasons) == 0 &&
		len(snapshot.Boundaries) == 0 &&
		snapshot.NextHostAction == "" &&
		!snapshot.RawOutputLoaded
}

func unavailableProductionAdapterCatalogDiscoveryView() ProductionAdapterCatalogDiscoveryView {
	return ProductionAdapterCatalogDiscoveryView{
		ContractVersion: ContractVersion,
		Projected:       true,
		Available:       false,
		Status:          "unavailable",
		Mode:            "production_adapter_catalog_discovery_view",
		RunnerEffect:    "none",
		PromptEffect:    "none",
		Boundaries: []Boundary{
			"production_adapter_catalog_discovery_view",
			"catalog_discovery_view_only",
			"display_safe_refs_only",
			"no_adapter_invocation",
			"no_runner_dispatch",
		},
	}.Normalize()
}

func productionAdapterCatalogSnapshotBoundaries(extra []Boundary) []Boundary {
	return MergeBoundaries([]Boundary{
		"production_adapter_catalog_snapshot",
		"catalog_snapshot_projection_only",
		"host_owned_adapter_catalog",
		"display_safe_refs_only",
		"no_adapter_invocation",
		"no_runner_dispatch",
	}, extra)
}

func productionAdapterCatalogDiscoveryViewBoundaries(extra []Boundary) []Boundary {
	return MergeBoundaries([]Boundary{
		"production_adapter_catalog_discovery_view",
		"catalog_discovery_view_only",
		"host_owned_adapter_catalog",
		"display_safe_refs_only",
		"no_adapter_invocation",
		"no_runner_dispatch",
	}, extra)
}

func normalizeProductionAdapterCatalogDescriptors(in []ProductionAdapterDescriptor) []ProductionAdapterDescriptor {
	out := make([]ProductionAdapterDescriptor, 0, len(in))
	for _, descriptor := range in {
		out = append(out, descriptor.Normalize())
	}
	return out
}

func cloneProductionAdapterDescriptors(in []ProductionAdapterDescriptor) []ProductionAdapterDescriptor {
	if len(in) == 0 {
		return nil
	}
	out := make([]ProductionAdapterDescriptor, 0, len(in))
	for _, descriptor := range in {
		out = append(out, descriptor.Clone())
	}
	return out
}

func productionAdapterReadyDescriptorCount(descriptors []ProductionAdapterDescriptor) int {
	count := 0
	for _, descriptor := range descriptors {
		ready := true
		for _, check := range productionAdapterDescriptorCompletenessChecks(descriptor) {
			if check.missing {
				ready = false
				break
			}
		}
		if ready {
			count++
		}
	}
	return count
}

func normalizeProductionAdapterKinds(in []ProductionAdapterKind) []ProductionAdapterKind {
	out := make([]ProductionAdapterKind, 0, len(in))
	seen := map[ProductionAdapterKind]struct{}{}
	for _, value := range in {
		kind := NormalizeProductionAdapterKind(string(value))
		if kind == "" {
			continue
		}
		if _, exists := seen[kind]; exists {
			continue
		}
		seen[kind] = struct{}{}
		out = append(out, kind)
	}
	return out
}

func cloneProductionAdapterKinds(in []ProductionAdapterKind) []ProductionAdapterKind {
	if len(in) == 0 {
		return nil
	}
	return append([]ProductionAdapterKind(nil), in...)
}
