// agentx-api-default-class: internal_candidate
package controlcontract

type ProductionAdapterCatalogReviewBlackboxFixtureInput struct {
	FixtureRef      DisplaySafeRef                       `json:"fixture_ref,omitempty"`
	ReviewPacket    ProductionAdapterCatalogReviewPacket `json:"review_packet,omitempty"`
	Selection       ProductionAdapterCatalogSelection    `json:"selection,omitempty"`
	Resolution      ProductionAdapterResolution          `json:"resolution,omitempty"`
	RawOutputLoaded bool                                 `json:"raw_output_loaded"`
}

type ProductionAdapterCatalogReviewBlackboxFixture struct {
	ContractVersion           string                  `json:"contract_version,omitempty"`
	Projected                 bool                    `json:"projected"`
	Available                 bool                    `json:"available"`
	Status                    string                  `json:"status,omitempty"`
	Mode                      string                  `json:"mode,omitempty"`
	RunnerEffect              string                  `json:"runner_effect,omitempty"`
	PromptEffect              string                  `json:"prompt_effect,omitempty"`
	ReadyForHostDisplay       bool                    `json:"ready_for_host_display"`
	ReadyForSelectionHandoff  bool                    `json:"ready_for_selection_handoff"`
	ReadyForResolutionHandoff bool                    `json:"ready_for_resolution_handoff"`
	FixtureRef                DisplaySafeRef          `json:"fixture_ref,omitempty"`
	DisplaySections           []string                `json:"display_sections,omitempty"`
	ReviewPacketRef           DisplaySafeRef          `json:"review_packet_ref,omitempty"`
	RegistrySnapshotRef       DisplaySafeRef          `json:"registry_snapshot_ref,omitempty"`
	RegistryStatus            HostActionStatus        `json:"registry_status,omitempty"`
	CatalogStatus             string                  `json:"catalog_status,omitempty"`
	ProviderRef               DisplaySafeRef          `json:"provider_ref,omitempty"`
	CatalogSnapshotRef        DisplaySafeRef          `json:"catalog_snapshot_ref,omitempty"`
	CatalogVersionRef         DisplaySafeRef          `json:"catalog_version_ref,omitempty"`
	CatalogDigestRef          DisplaySafeRef          `json:"catalog_digest_ref,omitempty"`
	DescriptorCount           int                     `json:"descriptor_count,omitempty"`
	ReadyDescriptorCount      int                     `json:"ready_descriptor_count,omitempty"`
	DescriptorRefs            []DisplaySafeRef        `json:"descriptor_refs,omitempty"`
	OwnerRefs                 []DisplaySafeRef        `json:"owner_refs,omitempty"`
	Kinds                     []ProductionAdapterKind `json:"kinds,omitempty"`
	SourceKinds               []ReplannerSourceKind   `json:"source_kinds,omitempty"`
	CandidateRefs             []DisplaySafeRef        `json:"candidate_refs,omitempty"`
	SelectionRequiredInputs   []MissingInput          `json:"selection_required_inputs,omitempty"`
	SelectionRef              DisplaySafeRef          `json:"selection_ref,omitempty"`
	SelectedAdapterRef        DisplaySafeRef          `json:"selected_adapter_ref,omitempty"`
	SelectedDescriptorRef     DisplaySafeRef          `json:"selected_descriptor_ref,omitempty"`
	SelectedSourceKind        ReplannerSourceKind     `json:"selected_source_kind,omitempty"`
	SelectedSourceRef         DisplaySafeRef          `json:"selected_source_ref,omitempty"`
	SelectedCandidateRef      DisplaySafeRef          `json:"selected_candidate_ref,omitempty"`
	ResolutionStatus          HostActionStatus        `json:"resolution_status,omitempty"`
	ReadyForHostPreflight     bool                    `json:"ready_for_host_preflight"`
	RequiredCapabilityRefs    []DisplaySafeRef        `json:"required_capability_refs,omitempty"`
	RequiredPolicyRefs        []DisplaySafeRef        `json:"required_policy_refs,omitempty"`
	RequiredApprovalRefs      []DisplaySafeRef        `json:"required_approval_refs,omitempty"`
	RequiredBudgetRef         DisplaySafeRef          `json:"required_budget_ref,omitempty"`
	RequiredPreflightRefs     []DisplaySafeRef        `json:"required_preflight_refs,omitempty"`
	MissingInputs             []MissingInput          `json:"missing_inputs,omitempty"`
	BlockedReasons            []string                `json:"blocked_reasons,omitempty"`
	FailureClass              FailureClass            `json:"failure_class,omitempty"`
	Boundaries                []Boundary              `json:"boundaries,omitempty"`
	NextHostAction            NextHostAction          `json:"next_host_action,omitempty"`
	RawOutputLoaded           bool                    `json:"raw_output_loaded"`
}

func BuildProductionAdapterCatalogReviewBlackboxFixture(input ProductionAdapterCatalogReviewBlackboxFixtureInput) ProductionAdapterCatalogReviewBlackboxFixture {
	if productionAdapterCatalogReviewPacketEmpty(input.ReviewPacket) {
		return unavailableProductionAdapterCatalogReviewBlackboxFixture()
	}
	review := input.ReviewPacket.Normalize()
	selection := input.Selection.Normalize()
	resolution := input.Resolution.Normalize()
	result := ProductionAdapterCatalogReviewBlackboxFixture{
		ContractVersion:         ContractVersion,
		Projected:               true,
		Available:               review.Available,
		Status:                  "blocked",
		Mode:                    "production_adapter_catalog_review_blackbox_fixture",
		RunnerEffect:            "none",
		PromptEffect:            "none",
		FixtureRef:              normalizeOneDisplaySafeRef(input.FixtureRef),
		DisplaySections:         productionAdapterCatalogReviewFixtureDisplaySections(),
		ReviewPacketRef:         review.ReviewPacketRef,
		RegistrySnapshotRef:     review.RegistrySnapshotRef,
		RegistryStatus:          review.RegistryStatus,
		CatalogStatus:           review.CatalogStatus,
		ProviderRef:             review.ProviderRef,
		CatalogSnapshotRef:      review.CatalogSnapshotRef,
		CatalogVersionRef:       review.CatalogVersionRef,
		CatalogDigestRef:        review.CatalogDigestRef,
		DescriptorCount:         review.DescriptorCount,
		ReadyDescriptorCount:    review.ReadyDescriptorCount,
		DescriptorRefs:          cloneDisplaySafeRefs(review.DescriptorRefs),
		OwnerRefs:               cloneDisplaySafeRefs(review.OwnerRefs),
		Kinds:                   cloneProductionAdapterKinds(review.Kinds),
		SourceKinds:             cloneReplannerSourceKinds(review.SourceKinds),
		CandidateRefs:           cloneDisplaySafeRefs(review.CandidateRefs),
		SelectionRequiredInputs: cloneMissingInputs(review.SelectionRequiredInputs),
		SelectionRef:            selection.SelectionRef,
		SelectedAdapterRef:      selection.SelectedAdapterRef,
		SelectedDescriptorRef:   selection.SelectedDescriptorRef,
		SelectedSourceKind:      selection.SelectedSourceKind,
		SelectedSourceRef:       selection.SelectedSourceRef,
		SelectedCandidateRef:    selection.SelectedCandidateRef,
		ResolutionStatus:        resolution.Status,
		ReadyForHostPreflight:   resolution.ReadyForHostPreflight,
		RequiredCapabilityRefs:  cloneDisplaySafeRefs(resolution.RequiredCapabilityRefs),
		RequiredPolicyRefs:      cloneDisplaySafeRefs(resolution.RequiredPolicyRefs),
		RequiredApprovalRefs:    cloneDisplaySafeRefs(resolution.RequiredApprovalRefs),
		RequiredBudgetRef:       selection.RequiredBudgetRef,
		RequiredPreflightRefs:   cloneDisplaySafeRefs(resolution.RequiredPreflightRefs),
		FailureClass:            FailureNone,
		Boundaries:              productionAdapterCatalogReviewBlackboxFixtureBoundaries(review.Boundaries, selection.Boundaries, resolution.Boundaries),
		NextHostAction:          firstNextHostAction(firstNextHostAction(review.NextHostAction, selection.NextHostAction), resolution.NextHostAction),
		RawOutputLoaded:         input.RawOutputLoaded || review.RawOutputLoaded || selection.RawOutputLoaded || resolution.RawOutputLoaded,
	}
	if productionAdapterCatalogReviewBlackboxFixtureUnsafe(input, review, selection, resolution) {
		result = productionAdapterCatalogReviewBlackboxFixtureBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if result.FixtureRef == "" {
		result = productionAdapterCatalogReviewBlackboxFixtureBlock(result, FailureEvidenceMissing, "adapter_catalog_review_fixture_ref_missing", "host:adapter_catalog_review_fixture_ref", "provide_adapter_catalog_review_fixture")
	}
	if !review.ReadyForHostReview || !review.ReadyForCatalogSelection {
		result = productionAdapterCatalogReviewBlackboxFixtureBlock(result, firstFailureClass(review.FailureClass, FailureConfigMissing), "adapter_catalog_review_packet_not_ready", "host:adapter_catalog_review_packet", firstNextHostAction(review.NextHostAction, "provide_adapter_catalog_review_packet"))
	}
	if result.FixtureRef != "" && review.ReadyForHostReview && review.ReadyForCatalogSelection && !result.RawOutputLoaded {
		result.ReadyForHostDisplay = true
	}
	if !selection.ReadyForResolution {
		result = productionAdapterCatalogReviewBlackboxFixtureBlock(result, firstFailureClass(selection.FailureClass, FailureConfigMissing), "adapter_catalog_selection_not_ready", "host:adapter_catalog_selection", firstNextHostAction(selection.NextHostAction, "host_may_select_adapter"))
	}
	if !resolution.ReadyForHostPreflight {
		result = productionAdapterCatalogReviewBlackboxFixtureBlock(result, firstFailureClass(resolution.FailureClass, FailureConfigMissing), "adapter_resolution_not_ready", "host:adapter_resolution", firstNextHostAction(resolution.NextHostAction, "build_adapter_resolution"))
	}
	for _, check := range productionAdapterCatalogReviewBlackboxFixtureConsistencyChecks(review, selection, resolution) {
		if check.mismatch {
			result = productionAdapterCatalogReviewBlackboxFixtureBlock(result, FailureInvalidInput, check.reason, check.input, "review_adapter_catalog")
		}
	}
	if len(result.BlockedReasons) == 0 && len(result.MissingInputs) == 0 {
		result.Status = "ready_for_adapter_resolution_handoff"
		result.ReadyForHostDisplay = true
		result.ReadyForSelectionHandoff = true
		result.ReadyForResolutionHandoff = true
		result.NextHostAction = "host_may_run_adapter_preflight"
		result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_host_adapter_catalog_display", "ready_for_adapter_resolution_handoff")
	}
	return result.Normalize()
}

func CloneProductionAdapterCatalogReviewBlackboxFixture(in ProductionAdapterCatalogReviewBlackboxFixture) ProductionAdapterCatalogReviewBlackboxFixture {
	out := in
	out.DisplaySections = cloneStringSlice(in.DisplaySections)
	out.DescriptorRefs = cloneDisplaySafeRefs(in.DescriptorRefs)
	out.OwnerRefs = cloneDisplaySafeRefs(in.OwnerRefs)
	out.Kinds = cloneProductionAdapterKinds(in.Kinds)
	out.SourceKinds = cloneReplannerSourceKinds(in.SourceKinds)
	out.CandidateRefs = cloneDisplaySafeRefs(in.CandidateRefs)
	out.SelectionRequiredInputs = cloneMissingInputs(in.SelectionRequiredInputs)
	out.RequiredCapabilityRefs = cloneDisplaySafeRefs(in.RequiredCapabilityRefs)
	out.RequiredPolicyRefs = cloneDisplaySafeRefs(in.RequiredPolicyRefs)
	out.RequiredApprovalRefs = cloneDisplaySafeRefs(in.RequiredApprovalRefs)
	out.RequiredPreflightRefs = cloneDisplaySafeRefs(in.RequiredPreflightRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (f ProductionAdapterCatalogReviewBlackboxFixture) Clone() ProductionAdapterCatalogReviewBlackboxFixture {
	return CloneProductionAdapterCatalogReviewBlackboxFixture(f)
}

func (f ProductionAdapterCatalogReviewBlackboxFixture) Normalize() ProductionAdapterCatalogReviewBlackboxFixture {
	out := CloneProductionAdapterCatalogReviewBlackboxFixture(f)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = normalizeControlToken(out.Status)
	out.Mode = normalizeControlToken(out.Mode)
	if out.Mode == "" {
		out.Mode = "production_adapter_catalog_review_blackbox_fixture"
	}
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	if out.RunnerEffect == "" {
		out.RunnerEffect = "none"
	}
	if out.PromptEffect == "" {
		out.PromptEffect = "none"
	}
	out.FixtureRef = normalizeOneDisplaySafeRef(out.FixtureRef)
	out.DisplaySections = normalizeControlTokenList(out.DisplaySections)
	out.ReviewPacketRef = normalizeOneDisplaySafeRef(out.ReviewPacketRef)
	out.RegistrySnapshotRef = normalizeOneDisplaySafeRef(out.RegistrySnapshotRef)
	out.RegistryStatus = NormalizeHostActionStatus(string(out.RegistryStatus))
	out.CatalogStatus = normalizeControlToken(out.CatalogStatus)
	out.ProviderRef = normalizeOneDisplaySafeRef(out.ProviderRef)
	out.CatalogSnapshotRef = normalizeOneDisplaySafeRef(out.CatalogSnapshotRef)
	out.CatalogVersionRef = normalizeOneDisplaySafeRef(out.CatalogVersionRef)
	out.CatalogDigestRef = normalizeOneDisplaySafeRef(out.CatalogDigestRef)
	out.DescriptorCount = maxNonNegativeInt(out.DescriptorCount)
	out.ReadyDescriptorCount = maxNonNegativeInt(out.ReadyDescriptorCount)
	out.DescriptorRefs = normalizeDisplaySafeRefs(out.DescriptorRefs)
	out.OwnerRefs = normalizeDisplaySafeRefs(out.OwnerRefs)
	out.Kinds = normalizeProductionAdapterKinds(out.Kinds)
	out.SourceKinds = normalizeReplannerSourceKinds(out.SourceKinds)
	out.CandidateRefs = normalizeDisplaySafeRefs(out.CandidateRefs)
	out.SelectionRequiredInputs = normalizeMissingInputs(out.SelectionRequiredInputs)
	out.SelectionRef = normalizeOneDisplaySafeRef(out.SelectionRef)
	out.SelectedAdapterRef = normalizeOneDisplaySafeRef(out.SelectedAdapterRef)
	out.SelectedDescriptorRef = normalizeOneDisplaySafeRef(out.SelectedDescriptorRef)
	out.SelectedSourceKind = NormalizeReplannerSourceKind(string(out.SelectedSourceKind))
	out.SelectedSourceRef = normalizeOneDisplaySafeRef(out.SelectedSourceRef)
	out.SelectedCandidateRef = normalizeOneDisplaySafeRef(out.SelectedCandidateRef)
	out.ResolutionStatus = NormalizeHostActionStatus(string(out.ResolutionStatus))
	out.RequiredCapabilityRefs = normalizeDisplaySafeRefs(out.RequiredCapabilityRefs)
	out.RequiredPolicyRefs = normalizeDisplaySafeRefs(out.RequiredPolicyRefs)
	out.RequiredApprovalRefs = normalizeDisplaySafeRefs(out.RequiredApprovalRefs)
	out.RequiredBudgetRef = normalizeOneDisplaySafeRef(out.RequiredBudgetRef)
	out.RequiredPreflightRefs = normalizeDisplaySafeRefs(out.RequiredPreflightRefs)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.BlockedReasons = normalizeControlTokenList(out.BlockedReasons)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	if !out.Available {
		out.Status = "unavailable"
		out.ReadyForHostDisplay = false
		out.ReadyForSelectionHandoff = false
		out.ReadyForResolutionHandoff = false
	}
	if out.Status == "" {
		out.Status = "blocked"
	}
	if out.RawOutputLoaded {
		out.Status = "blocked"
		out.ReadyForHostDisplay = false
		out.ReadyForSelectionHandoff = false
		out.ReadyForResolutionHandoff = false
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
	out.ReadyForHostPreflight = out.ReadyForHostPreflight && out.ResolutionStatus == HostActionReady
	if out.Status != "ready_for_adapter_resolution_handoff" {
		out.ReadyForSelectionHandoff = false
		out.ReadyForResolutionHandoff = false
	}
	out.ReadyForHostDisplay = out.ReadyForHostDisplay &&
		out.Available &&
		out.FixtureRef != "" &&
		out.ReviewPacketRef != "" &&
		out.RegistrySnapshotRef != "" &&
		out.CatalogSnapshotRef != "" &&
		out.ProviderRef != "" &&
		out.CatalogVersionRef != "" &&
		out.CatalogDigestRef != "" &&
		out.DescriptorCount > 0 &&
		!out.RawOutputLoaded
	out.ReadyForResolutionHandoff = out.ReadyForResolutionHandoff &&
		out.ReadyForHostDisplay &&
		out.ReadyForSelectionHandoff &&
		out.ReadyForHostPreflight &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0
	return out
}

func productionAdapterCatalogReviewBlackboxFixtureBlock(result ProductionAdapterCatalogReviewBlackboxFixture, failure FailureClass, reason string, missing MissingInput, next NextHostAction) ProductionAdapterCatalogReviewBlackboxFixture {
	result.Status = "blocked"
	result.ReadyForSelectionHandoff = false
	result.ReadyForResolutionHandoff = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	return result
}

type productionAdapterCatalogReviewBlackboxFixtureConsistencyCheck struct {
	mismatch bool
	reason   string
	input    MissingInput
}

func productionAdapterCatalogReviewBlackboxFixtureConsistencyChecks(review ProductionAdapterCatalogReviewPacket, selection ProductionAdapterCatalogSelection, resolution ProductionAdapterResolution) []productionAdapterCatalogReviewBlackboxFixtureConsistencyCheck {
	return []productionAdapterCatalogReviewBlackboxFixtureConsistencyCheck{
		{review.CatalogSnapshotRef != "" && selection.CatalogSnapshotRef != "" && review.CatalogSnapshotRef != selection.CatalogSnapshotRef, "fixture_selection_catalog_snapshot_mismatch", "host:adapter_catalog_selection"},
		{selection.SelectedAdapterRef != "" && len(review.DescriptorRefs) > 0 && !displaySafeRefSliceContains(review.DescriptorRefs, selection.SelectedAdapterRef), "fixture_selected_adapter_not_reviewed", "host:selected_adapter_ref"},
		{selection.SelectedCandidateRef != "" && len(review.CandidateRefs) > 0 && !displaySafeRefSliceContains(review.CandidateRefs, selection.SelectedCandidateRef), "fixture_selected_candidate_not_reviewed", "host:selected_candidate_strategy_ref"},
		{selection.SelectedSourceKind != "" && len(review.SourceKinds) > 0 && !replannerSourceKindContains(review.SourceKinds, selection.SelectedSourceKind), "fixture_selected_source_kind_not_reviewed", "host:selected_source_kind"},
		{selection.SelectionRef != "" && resolution.CatalogSelectionRef != "" && selection.SelectionRef != resolution.CatalogSelectionRef, "fixture_resolution_selection_ref_mismatch", "host:adapter_resolution"},
		{selection.CatalogSnapshotRef != "" && resolution.CatalogSnapshotRef != "" && selection.CatalogSnapshotRef != resolution.CatalogSnapshotRef, "fixture_resolution_catalog_snapshot_mismatch", "host:adapter_resolution"},
		{selection.SelectedAdapterRef != "" && resolution.RequestedAdapterRef != "" && selection.SelectedAdapterRef != resolution.RequestedAdapterRef, "fixture_resolution_adapter_ref_mismatch", "host:adapter_resolution"},
		{selection.SelectedAdapterRef != "" && resolution.AdapterRef != "" && selection.SelectedAdapterRef != resolution.AdapterRef, "fixture_resolution_descriptor_ref_mismatch", "host:adapter_resolution"},
		{selection.SelectedSourceKind != "" && resolution.SelectedSourceKind != "" && selection.SelectedSourceKind != resolution.SelectedSourceKind, "fixture_resolution_source_kind_mismatch", "host:adapter_resolution"},
		{selection.SelectedSourceRef != "" && resolution.SelectedSourceRef != "" && selection.SelectedSourceRef != resolution.SelectedSourceRef, "fixture_resolution_source_ref_mismatch", "host:adapter_resolution"},
		{selection.SelectedCandidateRef != "" && resolution.SelectedCandidateRef != "" && selection.SelectedCandidateRef != resolution.SelectedCandidateRef, "fixture_resolution_candidate_ref_mismatch", "host:adapter_resolution"},
		{review.CatalogSnapshotRef != "" && resolution.CatalogSnapshotRef != "" && review.CatalogSnapshotRef != resolution.CatalogSnapshotRef, "fixture_review_resolution_snapshot_mismatch", "host:adapter_resolution"},
	}
}

func productionAdapterCatalogReviewBlackboxFixtureUnsafe(input ProductionAdapterCatalogReviewBlackboxFixtureInput, review ProductionAdapterCatalogReviewPacket, selection ProductionAdapterCatalogSelection, resolution ProductionAdapterResolution) bool {
	return input.RawOutputLoaded ||
		displaySafeRefRejected(input.FixtureRef) ||
		productionAdapterCatalogReviewPacketUnsafe(review) ||
		productionAdapterCatalogSelectionUnsafe(selection) ||
		productionAdapterResolutionOutputUnsafe(resolution)
}

func productionAdapterResolutionOutputUnsafe(resolution ProductionAdapterResolution) bool {
	return resolution.RawOutputLoaded ||
		displaySafeRefRejected(resolution.AdapterRef) ||
		displaySafeRefRejected(resolution.DescriptorRef) ||
		productionAdapterDescriptorUnsafe(resolution.Descriptor) ||
		displaySafeRefRejected(resolution.ApplyEnvelopeRef) ||
		displaySafeRefRejected(resolution.SelectedSourceRef) ||
		displaySafeRefRejected(resolution.SelectedCandidateRef) ||
		displaySafeRefRejected(resolution.RequestedAdapterRef) ||
		displaySafeRefRejected(resolution.CatalogSnapshotRef) ||
		displaySafeRefRejected(resolution.CatalogSelectionRef) ||
		displaySafeRefRejected(resolution.HostPolicyRef) ||
		displaySafeRefSliceRejected(resolution.ApprovalContextRefs) ||
		displaySafeRefRejected(resolution.BudgetRef) ||
		displaySafeRefRejected(resolution.IdempotencyRef) ||
		displaySafeRefSliceRejected(resolution.RequiredCapabilityRefs) ||
		displaySafeRefSliceRejected(resolution.RequiredPolicyRefs) ||
		displaySafeRefSliceRejected(resolution.RequiredApprovalRefs) ||
		displaySafeRefSliceRejected(resolution.RequiredPreflightRefs)
}

func productionAdapterCatalogReviewPacketEmpty(packet ProductionAdapterCatalogReviewPacket) bool {
	return !packet.Projected &&
		!packet.Available &&
		packet.Status == "" &&
		packet.Mode == "" &&
		packet.ReviewPacketRef == "" &&
		packet.RegistrySnapshotRef == "" &&
		packet.CatalogSnapshotRef == "" &&
		packet.ProviderRef == "" &&
		packet.CatalogVersionRef == "" &&
		packet.CatalogDigestRef == "" &&
		len(packet.DescriptorRefs) == 0 &&
		len(packet.MissingInputs) == 0 &&
		len(packet.BlockedReasons) == 0 &&
		len(packet.Boundaries) == 0 &&
		packet.NextHostAction == "" &&
		!packet.RawOutputLoaded
}

func productionAdapterCatalogReviewFixtureDisplaySections() []string {
	return []string{
		"adapter_catalog_provenance",
		"adapter_catalog_descriptors",
		"adapter_catalog_selection_handoff",
		"adapter_resolution_handoff",
	}
}

func productionAdapterCatalogReviewBlackboxFixtureBoundaries(groups ...[]Boundary) []Boundary {
	return MergeBoundaries(
		[]Boundary{
			"production_adapter_catalog_review_blackbox_fixture",
			"catalog_review_fixture_projection_only",
			"host_cli_display_fixture",
			"host_owned_adapter_catalog_review",
			"display_safe_refs_only",
			"no_adapter_invocation",
			"no_runner_dispatch",
		},
		MergeBoundaries(groups...),
	)
}

func unavailableProductionAdapterCatalogReviewBlackboxFixture() ProductionAdapterCatalogReviewBlackboxFixture {
	return ProductionAdapterCatalogReviewBlackboxFixture{
		ContractVersion: ContractVersion,
		Projected:       true,
		Available:       false,
		Status:          "unavailable",
		Mode:            "production_adapter_catalog_review_blackbox_fixture",
		RunnerEffect:    "none",
		PromptEffect:    "none",
		DisplaySections: productionAdapterCatalogReviewFixtureDisplaySections(),
		Boundaries: []Boundary{
			"production_adapter_catalog_review_blackbox_fixture",
			"catalog_review_fixture_projection_only",
			"host_cli_display_fixture",
			"display_safe_refs_only",
			"no_adapter_invocation",
			"no_runner_dispatch",
		},
		NextHostAction: "provide_adapter_catalog_review_packet",
	}
}
