package controlcontract

// agentx-api: internal_candidate
type ProductionAdapterCatalogReviewPacketInput struct {
	ReviewPacketRef      DisplaySafeRef                        `json:"review_packet_ref,omitempty"`
	RegistrySnapshot     ProductionAdapterRegistrySnapshot     `json:"registry_snapshot,omitempty"`
	CatalogDiscoveryView ProductionAdapterCatalogDiscoveryView `json:"catalog_discovery_view,omitempty"`
	RawOutputLoaded      bool                                  `json:"raw_output_loaded"`
}

type ProductionAdapterCatalogReviewPacket struct {
	ContractVersion          string                  `json:"contract_version,omitempty"`
	Projected                bool                    `json:"projected"`
	Available                bool                    `json:"available"`
	Status                   string                  `json:"status,omitempty"`
	Mode                     string                  `json:"mode,omitempty"`
	RunnerEffect             string                  `json:"runner_effect,omitempty"`
	PromptEffect             string                  `json:"prompt_effect,omitempty"`
	ReadyForHostReview       bool                    `json:"ready_for_host_review"`
	ReadyForCatalogSelection bool                    `json:"ready_for_catalog_selection"`
	ReviewPacketRef          DisplaySafeRef          `json:"review_packet_ref,omitempty"`
	RegistrySnapshotRef      DisplaySafeRef          `json:"registry_snapshot_ref,omitempty"`
	RegistryStatus           HostActionStatus        `json:"registry_status,omitempty"`
	CatalogStatus            string                  `json:"catalog_status,omitempty"`
	ProviderRef              DisplaySafeRef          `json:"provider_ref,omitempty"`
	CatalogSnapshotRef       DisplaySafeRef          `json:"catalog_snapshot_ref,omitempty"`
	CatalogVersionRef        DisplaySafeRef          `json:"catalog_version_ref,omitempty"`
	CatalogDigestRef         DisplaySafeRef          `json:"catalog_digest_ref,omitempty"`
	MaxDescriptorCount       int                     `json:"max_descriptor_count,omitempty"`
	DescriptorCount          int                     `json:"descriptor_count,omitempty"`
	ReadyDescriptorCount     int                     `json:"ready_descriptor_count,omitempty"`
	DescriptorRefs           []DisplaySafeRef        `json:"descriptor_refs,omitempty"`
	OwnerRefs                []DisplaySafeRef        `json:"owner_refs,omitempty"`
	Kinds                    []ProductionAdapterKind `json:"kinds,omitempty"`
	SourceKinds              []ReplannerSourceKind   `json:"source_kinds,omitempty"`
	CandidateRefs            []DisplaySafeRef        `json:"candidate_refs,omitempty"`
	ProvidesCapabilityRefs   []DisplaySafeRef        `json:"provides_capability_refs,omitempty"`
	RequiresCapabilityRefs   []DisplaySafeRef        `json:"requires_capability_refs,omitempty"`
	PolicyRefs               []DisplaySafeRef        `json:"policy_refs,omitempty"`
	ApprovalRefs             []DisplaySafeRef        `json:"approval_refs,omitempty"`
	BudgetRefs               []DisplaySafeRef        `json:"budget_refs,omitempty"`
	PreflightRefs            []DisplaySafeRef        `json:"preflight_refs,omitempty"`
	SelectionRequiredInputs  []MissingInput          `json:"selection_required_inputs,omitempty"`
	MissingInputs            []MissingInput          `json:"missing_inputs,omitempty"`
	BlockedReasons           []string                `json:"blocked_reasons,omitempty"`
	FailureClass             FailureClass            `json:"failure_class,omitempty"`
	Boundaries               []Boundary              `json:"boundaries,omitempty"`
	NextHostAction           NextHostAction          `json:"next_host_action,omitempty"`
	RawOutputLoaded          bool                    `json:"raw_output_loaded"`
}

// agentx-api: internal_candidate
func BuildProductionAdapterCatalogReviewPacket(input ProductionAdapterCatalogReviewPacketInput) ProductionAdapterCatalogReviewPacket {
	if productionAdapterRegistrySnapshotEmpty(input.RegistrySnapshot) {
		return unavailableProductionAdapterCatalogReviewPacket()
	}
	registry := input.RegistrySnapshot.Normalize()
	view := input.CatalogDiscoveryView.Normalize()
	if productionAdapterCatalogDiscoveryViewEmpty(input.CatalogDiscoveryView) {
		view = BuildProductionAdapterCatalogDiscoveryView(registry.CatalogSnapshot)
	}
	result := ProductionAdapterCatalogReviewPacket{
		ContractVersion:        ContractVersion,
		Projected:              true,
		Available:              registry.Available && view.Available,
		Status:                 "blocked",
		Mode:                   "production_adapter_catalog_review_packet",
		RunnerEffect:           "none",
		PromptEffect:           "none",
		ReviewPacketRef:        normalizeOneDisplaySafeRef(input.ReviewPacketRef),
		RegistrySnapshotRef:    registry.RegistrySnapshotRef,
		RegistryStatus:         registry.Status,
		CatalogStatus:          view.Status,
		ProviderRef:            registry.ProviderRef,
		CatalogSnapshotRef:     registry.CatalogSnapshotRef,
		CatalogVersionRef:      registry.CatalogVersionRef,
		CatalogDigestRef:       registry.CatalogDigestRef,
		MaxDescriptorCount:     registry.MaxDescriptorCount,
		DescriptorCount:        view.DescriptorCount,
		ReadyDescriptorCount:   view.ReadyDescriptorCount,
		DescriptorRefs:         cloneDisplaySafeRefs(view.DescriptorRefs),
		OwnerRefs:              cloneDisplaySafeRefs(view.OwnerRefs),
		Kinds:                  cloneProductionAdapterKinds(view.Kinds),
		SourceKinds:            cloneReplannerSourceKinds(view.SourceKinds),
		CandidateRefs:          cloneDisplaySafeRefs(view.CandidateRefs),
		ProvidesCapabilityRefs: cloneDisplaySafeRefs(view.ProvidesCapabilityRefs),
		RequiresCapabilityRefs: cloneDisplaySafeRefs(view.RequiresCapabilityRefs),
		PolicyRefs:             cloneDisplaySafeRefs(view.PolicyRefs),
		ApprovalRefs:           cloneDisplaySafeRefs(view.ApprovalRefs),
		BudgetRefs:             cloneDisplaySafeRefs(view.BudgetRefs),
		PreflightRefs:          cloneDisplaySafeRefs(view.PreflightRefs),
		FailureClass:           FailureNone,
		Boundaries:             productionAdapterCatalogReviewPacketBoundaries(registry.Boundaries, view.Boundaries),
		NextHostAction:         firstNextHostAction(registry.NextHostAction, view.NextHostAction),
		RawOutputLoaded:        input.RawOutputLoaded || registry.RawOutputLoaded || view.RawOutputLoaded,
	}
	if productionAdapterCatalogReviewPacketInputUnsafe(input, view) {
		result = productionAdapterCatalogReviewPacketBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if result.ReviewPacketRef == "" {
		result = productionAdapterCatalogReviewPacketBlock(result, FailureEvidenceMissing, "adapter_catalog_review_packet_ref_missing", "host:adapter_catalog_review_packet_ref", "provide_adapter_catalog_review_packet")
	}
	if !registry.ReadyForCatalogSnapshot {
		result = productionAdapterCatalogReviewPacketBlock(result, firstFailureClass(registry.FailureClass, FailureConfigMissing), "adapter_registry_snapshot_not_ready", "host:adapter_registry_snapshot", firstNextHostAction(registry.NextHostAction, "provide_adapter_registry_snapshot"))
	}
	if !view.ReadyForHostSelection {
		result = productionAdapterCatalogReviewPacketBlock(result, firstFailureClass(view.FailureClass, FailureConfigMissing), "adapter_catalog_discovery_not_ready", "host:adapter_catalog_discovery_view", firstNextHostAction(view.NextHostAction, "provide_adapter_catalog"))
	}
	for _, check := range productionAdapterCatalogReviewPacketConsistencyChecks(registry, view) {
		if check.mismatch {
			result = productionAdapterCatalogReviewPacketBlock(result, FailureInvalidInput, check.reason, check.input, "review_adapter_catalog")
		}
	}
	if len(result.BlockedReasons) == 0 && len(result.MissingInputs) == 0 {
		result.Status = "ready_for_adapter_catalog_review"
		result.ReadyForHostReview = true
		result.ReadyForCatalogSelection = true
		result.SelectionRequiredInputs = productionAdapterCatalogSelectionRequiredInputs()
		result.NextHostAction = "host_may_select_adapter"
		result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_adapter_catalog_review", "ready_for_catalog_selection")
	}
	return result.Normalize()
}

func CloneProductionAdapterCatalogReviewPacket(in ProductionAdapterCatalogReviewPacket) ProductionAdapterCatalogReviewPacket {
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
	out.SelectionRequiredInputs = cloneMissingInputs(in.SelectionRequiredInputs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (p ProductionAdapterCatalogReviewPacket) Clone() ProductionAdapterCatalogReviewPacket {
	return CloneProductionAdapterCatalogReviewPacket(p)
}

func (p ProductionAdapterCatalogReviewPacket) Normalize() ProductionAdapterCatalogReviewPacket {
	out := CloneProductionAdapterCatalogReviewPacket(p)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = normalizeControlToken(out.Status)
	out.Mode = normalizeControlToken(out.Mode)
	if out.Mode == "" {
		out.Mode = "production_adapter_catalog_review_packet"
	}
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	if out.RunnerEffect == "" {
		out.RunnerEffect = "none"
	}
	if out.PromptEffect == "" {
		out.PromptEffect = "none"
	}
	out.ReviewPacketRef = normalizeOneDisplaySafeRef(out.ReviewPacketRef)
	out.RegistrySnapshotRef = normalizeOneDisplaySafeRef(out.RegistrySnapshotRef)
	out.RegistryStatus = NormalizeHostActionStatus(string(out.RegistryStatus))
	out.CatalogStatus = normalizeControlToken(out.CatalogStatus)
	out.ProviderRef = normalizeOneDisplaySafeRef(out.ProviderRef)
	out.CatalogSnapshotRef = normalizeOneDisplaySafeRef(out.CatalogSnapshotRef)
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
	out.SelectionRequiredInputs = normalizeMissingInputs(out.SelectionRequiredInputs)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.BlockedReasons = normalizeControlTokenList(out.BlockedReasons)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	if !out.Available {
		out.Status = "unavailable"
		out.ReadyForHostReview = false
		out.ReadyForCatalogSelection = false
	}
	if out.Status == "" {
		out.Status = "blocked"
	}
	if out.RawOutputLoaded {
		out.Status = "blocked"
		out.ReadyForHostReview = false
		out.ReadyForCatalogSelection = false
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
	out.ReadyForHostReview = out.Available &&
		out.Status == "ready_for_adapter_catalog_review" &&
		out.ReviewPacketRef != "" &&
		out.RegistrySnapshotRef != "" &&
		out.ProviderRef != "" &&
		out.CatalogSnapshotRef != "" &&
		out.CatalogVersionRef != "" &&
		out.CatalogDigestRef != "" &&
		out.DescriptorCount > 0 &&
		out.ReadyDescriptorCount > 0 &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.RawOutputLoaded
	out.ReadyForCatalogSelection = out.ReadyForHostReview && len(out.SelectionRequiredInputs) > 0
	return out
}

func productionAdapterCatalogReviewPacketBlock(result ProductionAdapterCatalogReviewPacket, failure FailureClass, reason string, missing MissingInput, next NextHostAction) ProductionAdapterCatalogReviewPacket {
	result.Status = "blocked"
	result.ReadyForHostReview = false
	result.ReadyForCatalogSelection = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	return result
}

type productionAdapterCatalogReviewPacketConsistencyCheck struct {
	mismatch bool
	reason   string
	input    MissingInput
}

func productionAdapterCatalogReviewPacketConsistencyChecks(registry ProductionAdapterRegistrySnapshot, view ProductionAdapterCatalogDiscoveryView) []productionAdapterCatalogReviewPacketConsistencyCheck {
	return []productionAdapterCatalogReviewPacketConsistencyCheck{
		{registry.CatalogSnapshotRef != "" && view.CatalogSnapshotRef != "" && registry.CatalogSnapshotRef != view.CatalogSnapshotRef, "catalog_review_snapshot_ref_mismatch", "host:adapter_catalog_discovery_view"},
		{registry.ProviderRef != "" && view.ProviderRef != "" && registry.ProviderRef != view.ProviderRef, "catalog_review_provider_ref_mismatch", "host:adapter_catalog_discovery_view"},
		{registry.CatalogVersionRef != "" && view.CatalogVersionRef != "" && registry.CatalogVersionRef != view.CatalogVersionRef, "catalog_review_version_ref_mismatch", "host:adapter_catalog_discovery_view"},
		{registry.CatalogDigestRef != "" && view.CatalogDigestRef != "" && registry.CatalogDigestRef != view.CatalogDigestRef, "catalog_review_digest_ref_mismatch", "host:adapter_catalog_discovery_view"},
		{registry.DescriptorCount != 0 && view.DescriptorCount != 0 && registry.DescriptorCount != view.DescriptorCount, "catalog_review_descriptor_count_mismatch", "host:adapter_catalog_discovery_view"},
	}
}

func productionAdapterCatalogReviewPacketInputUnsafe(input ProductionAdapterCatalogReviewPacketInput, view ProductionAdapterCatalogDiscoveryView) bool {
	return input.RawOutputLoaded ||
		displaySafeRefRejected(input.ReviewPacketRef) ||
		productionAdapterRegistrySnapshotUnsafe(input.RegistrySnapshot) ||
		productionAdapterCatalogDiscoveryViewUnsafe(view)
}

func productionAdapterCatalogReviewPacketUnsafe(input ProductionAdapterCatalogReviewPacket) bool {
	return input.RawOutputLoaded ||
		displaySafeRefRejected(input.ReviewPacketRef) ||
		displaySafeRefRejected(input.RegistrySnapshotRef) ||
		displaySafeRefRejected(input.ProviderRef) ||
		displaySafeRefRejected(input.CatalogSnapshotRef) ||
		displaySafeRefRejected(input.CatalogVersionRef) ||
		displaySafeRefRejected(input.CatalogDigestRef) ||
		displaySafeRefSliceRejected(input.DescriptorRefs) ||
		displaySafeRefSliceRejected(input.OwnerRefs) ||
		displaySafeRefSliceRejected(input.CandidateRefs) ||
		displaySafeRefSliceRejected(input.ProvidesCapabilityRefs) ||
		displaySafeRefSliceRejected(input.RequiresCapabilityRefs) ||
		displaySafeRefSliceRejected(input.PolicyRefs) ||
		displaySafeRefSliceRejected(input.ApprovalRefs) ||
		displaySafeRefSliceRejected(input.BudgetRefs) ||
		displaySafeRefSliceRejected(input.PreflightRefs)
}

func productionAdapterCatalogSelectionRequiredInputs() []MissingInput {
	return []MissingInput{
		"host:adapter_selection_ref",
		"host:selected_adapter_ref",
		"host:selected_source_kind",
		"host:selected_source_ref",
		"host:selected_candidate_strategy_ref",
	}
}

func productionAdapterCatalogReviewPacketBoundaries(groups ...[]Boundary) []Boundary {
	return MergeBoundaries(
		[]Boundary{
			"production_adapter_catalog_review_packet",
			"catalog_review_packet_projection_only",
			"host_owned_adapter_catalog_review",
			"display_safe_refs_only",
			"no_adapter_invocation",
			"no_runner_dispatch",
		},
		MergeBoundaries(groups...),
	)
}

func unavailableProductionAdapterCatalogReviewPacket() ProductionAdapterCatalogReviewPacket {
	return ProductionAdapterCatalogReviewPacket{
		ContractVersion: ContractVersion,
		Projected:       true,
		Available:       false,
		Status:          "unavailable",
		Mode:            "production_adapter_catalog_review_packet",
		RunnerEffect:    "none",
		PromptEffect:    "none",
		Boundaries: []Boundary{
			"production_adapter_catalog_review_packet",
			"catalog_review_packet_projection_only",
			"display_safe_refs_only",
			"no_adapter_invocation",
			"no_runner_dispatch",
		},
		NextHostAction: "provide_adapter_registry_snapshot",
	}
}

func productionAdapterRegistrySnapshotUnsafe(snapshot ProductionAdapterRegistrySnapshot) bool {
	return snapshot.RawOutputLoaded ||
		displaySafeRefRejected(snapshot.RegistrySnapshotRef) ||
		displaySafeRefRejected(snapshot.ProviderRef) ||
		displaySafeRefRejected(snapshot.CatalogSnapshotRef) ||
		displaySafeRefRejected(snapshot.CatalogVersionRef) ||
		displaySafeRefRejected(snapshot.CatalogDigestRef) ||
		displaySafeRefRejected(snapshot.ExpectedCatalogVersionRef) ||
		displaySafeRefRejected(snapshot.ExpectedCatalogDigestRef) ||
		productionAdapterCatalogSnapshotUnsafe(snapshot.CatalogSnapshot)
}

func productionAdapterRegistrySnapshotEmpty(snapshot ProductionAdapterRegistrySnapshot) bool {
	return !snapshot.Projected &&
		!snapshot.Available &&
		snapshot.Status == "" &&
		snapshot.RegistrySnapshotRef == "" &&
		snapshot.ProviderRef == "" &&
		snapshot.CatalogSnapshotRef == "" &&
		snapshot.CatalogVersionRef == "" &&
		snapshot.CatalogDigestRef == "" &&
		snapshot.MaxDescriptorCount == 0 &&
		productionAdapterCatalogSnapshotEmpty(snapshot.CatalogSnapshot) &&
		len(snapshot.MissingInputs) == 0 &&
		len(snapshot.BlockedReasons) == 0 &&
		len(snapshot.Boundaries) == 0 &&
		snapshot.NextHostAction == "" &&
		!snapshot.RawOutputLoaded
}

func productionAdapterCatalogDiscoveryViewUnsafe(view ProductionAdapterCatalogDiscoveryView) bool {
	return view.RawOutputLoaded ||
		displaySafeRefRejected(view.CatalogSnapshotRef) ||
		displaySafeRefRejected(view.Producer) ||
		displaySafeRefRejected(view.ProviderRef) ||
		displaySafeRefRejected(view.CatalogVersionRef) ||
		displaySafeRefRejected(view.CatalogDigestRef) ||
		displaySafeRefSliceRejected(view.DescriptorRefs) ||
		displaySafeRefSliceRejected(view.OwnerRefs) ||
		displaySafeRefSliceRejected(view.CandidateRefs) ||
		displaySafeRefSliceRejected(view.ProvidesCapabilityRefs) ||
		displaySafeRefSliceRejected(view.RequiresCapabilityRefs) ||
		displaySafeRefSliceRejected(view.PolicyRefs) ||
		displaySafeRefSliceRejected(view.ApprovalRefs) ||
		displaySafeRefSliceRejected(view.BudgetRefs) ||
		displaySafeRefSliceRejected(view.PreflightRefs)
}

func productionAdapterCatalogDiscoveryViewEmpty(view ProductionAdapterCatalogDiscoveryView) bool {
	return !view.Projected &&
		!view.Available &&
		view.Status == "" &&
		view.Mode == "" &&
		view.CatalogSnapshotRef == "" &&
		view.Producer == "" &&
		view.ProviderRef == "" &&
		view.CatalogVersionRef == "" &&
		view.CatalogDigestRef == "" &&
		view.MaxDescriptorCount == 0 &&
		len(view.DescriptorRefs) == 0 &&
		len(view.MissingInputs) == 0 &&
		len(view.BlockedReasons) == 0 &&
		len(view.Boundaries) == 0 &&
		view.NextHostAction == "" &&
		!view.RawOutputLoaded
}
