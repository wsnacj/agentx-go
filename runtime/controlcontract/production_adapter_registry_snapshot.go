package controlcontract

type ProductionAdapterRegistrySnapshotInput struct {
	RegistrySnapshotRef       DisplaySafeRef                `json:"registry_snapshot_ref,omitempty"`
	ProviderRef               DisplaySafeRef                `json:"provider_ref,omitempty"`
	CatalogSnapshotRef        DisplaySafeRef                `json:"catalog_snapshot_ref,omitempty"`
	CatalogVersionRef         DisplaySafeRef                `json:"catalog_version_ref,omitempty"`
	CatalogDigestRef          DisplaySafeRef                `json:"catalog_digest_ref,omitempty"`
	ExpectedCatalogVersionRef DisplaySafeRef                `json:"expected_catalog_version_ref,omitempty"`
	ExpectedCatalogDigestRef  DisplaySafeRef                `json:"expected_catalog_digest_ref,omitempty"`
	MaxDescriptorCount        int                           `json:"max_descriptor_count,omitempty"`
	HostPolicyRefs            []DisplaySafeRef              `json:"host_policy_refs,omitempty"`
	Descriptors               []ProductionAdapterDescriptor `json:"descriptors,omitempty"`
	Boundaries                []Boundary                    `json:"boundaries,omitempty"`
	RawOutputLoaded           bool                          `json:"raw_output_loaded"`
}

type ProductionAdapterRegistrySnapshot struct {
	ContractVersion           string                           `json:"contract_version,omitempty"`
	Projected                 bool                             `json:"projected"`
	Available                 bool                             `json:"available"`
	Status                    HostActionStatus                 `json:"status,omitempty"`
	ReadyForCatalogSnapshot   bool                             `json:"ready_for_catalog_snapshot"`
	RegistrySnapshotRef       DisplaySafeRef                   `json:"registry_snapshot_ref,omitempty"`
	ProviderRef               DisplaySafeRef                   `json:"provider_ref,omitempty"`
	CatalogSnapshotRef        DisplaySafeRef                   `json:"catalog_snapshot_ref,omitempty"`
	CatalogVersionRef         DisplaySafeRef                   `json:"catalog_version_ref,omitempty"`
	CatalogDigestRef          DisplaySafeRef                   `json:"catalog_digest_ref,omitempty"`
	ExpectedCatalogVersionRef DisplaySafeRef                   `json:"expected_catalog_version_ref,omitempty"`
	ExpectedCatalogDigestRef  DisplaySafeRef                   `json:"expected_catalog_digest_ref,omitempty"`
	MaxDescriptorCount        int                              `json:"max_descriptor_count,omitempty"`
	DescriptorCount           int                              `json:"descriptor_count,omitempty"`
	ReadyDescriptorCount      int                              `json:"ready_descriptor_count,omitempty"`
	CatalogSnapshot           ProductionAdapterCatalogSnapshot `json:"catalog_snapshot,omitempty"`
	MissingInputs             []MissingInput                   `json:"missing_inputs,omitempty"`
	BlockedReasons            []string                         `json:"blocked_reasons,omitempty"`
	FailureClass              FailureClass                     `json:"failure_class,omitempty"`
	Boundaries                []Boundary                       `json:"boundaries,omitempty"`
	NextHostAction            NextHostAction                   `json:"next_host_action,omitempty"`
	RunnerEffect              string                           `json:"runner_effect,omitempty"`
	PromptEffect              string                           `json:"prompt_effect,omitempty"`
	RawOutputLoaded           bool                             `json:"raw_output_loaded"`
}

func BuildProductionAdapterRegistrySnapshot(input ProductionAdapterRegistrySnapshotInput) ProductionAdapterRegistrySnapshot {
	providerRef := normalizeOneDisplaySafeRef(input.ProviderRef)
	catalogSnapshot := BuildProductionAdapterCatalogSnapshot(ProductionAdapterCatalogSnapshotInput{
		CatalogSnapshotRef: input.CatalogSnapshotRef,
		Producer:           providerRef,
		ProviderRef:        providerRef,
		CatalogVersionRef:  input.CatalogVersionRef,
		CatalogDigestRef:   input.CatalogDigestRef,
		MaxDescriptorCount: input.MaxDescriptorCount,
		HostPolicyRefs:     input.HostPolicyRefs,
		Descriptors:        input.Descriptors,
		Boundaries:         AppendBoundaries(input.Boundaries, "registry_derived_adapter_catalog"),
		RawOutputLoaded:    input.RawOutputLoaded,
	})
	result := ProductionAdapterRegistrySnapshot{
		ContractVersion:           ContractVersion,
		Projected:                 true,
		Available:                 true,
		Status:                    HostActionBlocked,
		RegistrySnapshotRef:       normalizeOneDisplaySafeRef(input.RegistrySnapshotRef),
		ProviderRef:               providerRef,
		CatalogSnapshotRef:        catalogSnapshot.CatalogSnapshotRef,
		CatalogVersionRef:         normalizeOneDisplaySafeRef(input.CatalogVersionRef),
		CatalogDigestRef:          normalizeOneDisplaySafeRef(input.CatalogDigestRef),
		ExpectedCatalogVersionRef: normalizeOneDisplaySafeRef(input.ExpectedCatalogVersionRef),
		ExpectedCatalogDigestRef:  normalizeOneDisplaySafeRef(input.ExpectedCatalogDigestRef),
		MaxDescriptorCount:        maxNonNegativeInt(input.MaxDescriptorCount),
		DescriptorCount:           catalogSnapshot.DescriptorCount,
		ReadyDescriptorCount:      catalogSnapshot.ReadyDescriptorCount,
		CatalogSnapshot:           catalogSnapshot,
		FailureClass:              FailureNone,
		Boundaries:                productionAdapterRegistrySnapshotBoundaries(input.Boundaries),
		RunnerEffect:              "none",
		PromptEffect:              "none",
		RawOutputLoaded:           input.RawOutputLoaded || catalogSnapshot.RawOutputLoaded,
	}
	if productionAdapterRegistrySnapshotInputUnsafe(input) {
		result = productionAdapterRegistrySnapshotBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if result.RegistrySnapshotRef == "" {
		result = productionAdapterRegistrySnapshotBlock(result, FailureEvidenceMissing, "adapter_registry_snapshot_ref_missing", "host:adapter_registry_snapshot_ref", "provide_adapter_registry_snapshot")
	}
	if result.ProviderRef == "" {
		result = productionAdapterRegistrySnapshotBlock(result, FailureConfigMissing, "adapter_registry_provider_ref_missing", "host:adapter_registry_provider_ref", "provide_adapter_registry_provider_ref")
	}
	if result.CatalogSnapshotRef == "" {
		result = productionAdapterRegistrySnapshotBlock(result, FailureEvidenceMissing, "catalog_snapshot_ref_missing", "host:adapter_catalog_snapshot_ref", "provide_adapter_catalog")
	}
	if result.CatalogVersionRef == "" {
		result = productionAdapterRegistrySnapshotBlock(result, FailureEvidenceMissing, "adapter_catalog_version_ref_missing", "host:adapter_catalog_version_ref", "provide_adapter_catalog_version")
	}
	if result.CatalogDigestRef == "" {
		result = productionAdapterRegistrySnapshotBlock(result, FailureEvidenceMissing, "adapter_catalog_digest_ref_missing", "host:adapter_catalog_digest_ref", "provide_adapter_catalog_digest")
	}
	if result.MaxDescriptorCount <= 0 {
		result = productionAdapterRegistrySnapshotBlock(result, FailureConfigMissing, "adapter_catalog_descriptor_limit_missing", "host:adapter_catalog_descriptor_limit", "provide_adapter_catalog_descriptor_limit")
	} else if result.DescriptorCount > result.MaxDescriptorCount {
		result = productionAdapterRegistrySnapshotBlock(result, FailurePolicyBlocked, "adapter_registry_descriptor_count_exceeded", "host:adapter_catalog_descriptor_limit", "review_adapter_registry_snapshot")
	}
	if !catalogSnapshot.ReadyForHostSelection {
		result = productionAdapterRegistrySnapshotBlock(result, firstFailureClass(catalogSnapshot.FailureClass, FailureConfigMissing), "adapter_catalog_not_ready", "host:adapter_catalog_snapshot", "provide_adapter_catalog")
	}
	if result.ExpectedCatalogVersionRef != "" && result.CatalogVersionRef != "" && result.CatalogVersionRef != result.ExpectedCatalogVersionRef {
		result = productionAdapterRegistrySnapshotBlock(result, FailureVerificationFailed, "adapter_registry_snapshot_stale", "host:adapter_registry_refresh", "refresh_adapter_registry_snapshot")
	}
	if result.ExpectedCatalogDigestRef != "" && result.CatalogDigestRef != "" && result.CatalogDigestRef != result.ExpectedCatalogDigestRef {
		result = productionAdapterRegistrySnapshotBlock(result, FailureVerificationFailed, "adapter_registry_snapshot_conflict", "host:adapter_registry_review", "review_adapter_registry_snapshot")
	}
	if len(result.BlockedReasons) == 0 && len(result.MissingInputs) == 0 {
		result.Status = HostActionReady
		result.ReadyForCatalogSnapshot = true
		result.NextHostAction = "host_may_review_adapter_catalog"
		result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_adapter_catalog_review")
	}
	return result.Normalize()
}

func CloneProductionAdapterRegistrySnapshot(in ProductionAdapterRegistrySnapshot) ProductionAdapterRegistrySnapshot {
	out := in
	out.CatalogSnapshot = in.CatalogSnapshot.Clone()
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (s ProductionAdapterRegistrySnapshot) Clone() ProductionAdapterRegistrySnapshot {
	return CloneProductionAdapterRegistrySnapshot(s)
}

func (s ProductionAdapterRegistrySnapshot) Normalize() ProductionAdapterRegistrySnapshot {
	out := CloneProductionAdapterRegistrySnapshot(s)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.RegistrySnapshotRef = normalizeOneDisplaySafeRef(out.RegistrySnapshotRef)
	out.ProviderRef = normalizeOneDisplaySafeRef(out.ProviderRef)
	out.CatalogSnapshotRef = normalizeOneDisplaySafeRef(out.CatalogSnapshotRef)
	out.CatalogVersionRef = normalizeOneDisplaySafeRef(out.CatalogVersionRef)
	out.CatalogDigestRef = normalizeOneDisplaySafeRef(out.CatalogDigestRef)
	out.ExpectedCatalogVersionRef = normalizeOneDisplaySafeRef(out.ExpectedCatalogVersionRef)
	out.ExpectedCatalogDigestRef = normalizeOneDisplaySafeRef(out.ExpectedCatalogDigestRef)
	out.MaxDescriptorCount = maxNonNegativeInt(out.MaxDescriptorCount)
	out.CatalogSnapshot = out.CatalogSnapshot.Normalize()
	out.DescriptorCount = out.CatalogSnapshot.DescriptorCount
	out.ReadyDescriptorCount = out.CatalogSnapshot.ReadyDescriptorCount
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
	out.ReadyForCatalogSnapshot = out.Status == HostActionReady &&
		out.RegistrySnapshotRef != "" &&
		out.ProviderRef != "" &&
		out.CatalogSnapshotRef != "" &&
		out.CatalogVersionRef != "" &&
		out.CatalogDigestRef != "" &&
		out.MaxDescriptorCount > 0 &&
		out.CatalogSnapshot.ReadyForHostSelection &&
		out.CatalogSnapshot.ProviderRef == out.ProviderRef &&
		out.CatalogSnapshot.CatalogVersionRef == out.CatalogVersionRef &&
		out.CatalogSnapshot.CatalogDigestRef == out.CatalogDigestRef &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.RawOutputLoaded
	if !out.ReadyForCatalogSnapshot && !productionAdapterCatalogSnapshotEmpty(out.CatalogSnapshot) && out.CatalogSnapshot.ReadyForHostSelection {
		out.CatalogSnapshot = productionAdapterCatalogSnapshotBlock(out.CatalogSnapshot, firstFailureClass(out.FailureClass, FailureConfigMissing), "adapter_registry_snapshot_not_ready", "host:adapter_registry_snapshot", firstNextHostAction(out.NextHostAction, "review_adapter_registry_snapshot")).Normalize()
	}
	return out
}

func productionAdapterRegistrySnapshotBlock(result ProductionAdapterRegistrySnapshot, failure FailureClass, reason string, missing MissingInput, next NextHostAction) ProductionAdapterRegistrySnapshot {
	result.Status = HostActionBlocked
	result.ReadyForCatalogSnapshot = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	return result
}

func productionAdapterRegistrySnapshotInputUnsafe(input ProductionAdapterRegistrySnapshotInput) bool {
	if input.RawOutputLoaded ||
		displaySafeRefRejected(input.RegistrySnapshotRef) ||
		displaySafeRefRejected(input.ProviderRef) ||
		displaySafeRefRejected(input.CatalogSnapshotRef) ||
		displaySafeRefRejected(input.CatalogVersionRef) ||
		displaySafeRefRejected(input.CatalogDigestRef) ||
		displaySafeRefRejected(input.ExpectedCatalogVersionRef) ||
		displaySafeRefRejected(input.ExpectedCatalogDigestRef) ||
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

func productionAdapterRegistrySnapshotBoundaries(extra []Boundary) []Boundary {
	return MergeBoundaries([]Boundary{
		"production_adapter_registry_snapshot",
		"registry_snapshot_projection_only",
		"host_owned_adapter_registry",
		"registry_to_catalog_projection",
		"adapter_catalog_provenance_gate",
		"display_safe_refs_only",
		"no_adapter_invocation",
		"no_runner_dispatch",
	}, extra)
}
