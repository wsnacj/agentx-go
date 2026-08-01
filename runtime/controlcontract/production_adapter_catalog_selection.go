package controlcontract

// agentx-api: internal_candidate
type ProductionAdapterCatalogSelectionInput struct {
	CatalogSnapshot      ProductionAdapterCatalogSnapshot `json:"catalog_snapshot,omitempty"`
	SelectionRef         DisplaySafeRef                   `json:"selection_ref,omitempty"`
	SelectedAdapterRef   DisplaySafeRef                   `json:"selected_adapter_ref,omitempty"`
	SelectedSourceKind   ReplannerSourceKind              `json:"selected_source_kind,omitempty"`
	SelectedSourceRef    DisplaySafeRef                   `json:"selected_source_ref,omitempty"`
	SelectedCandidateRef DisplaySafeRef                   `json:"selected_candidate_ref,omitempty"`
	RawOutputLoaded      bool                             `json:"raw_output_loaded"`
}

type ProductionAdapterCatalogSelection struct {
	ContractVersion        string                      `json:"contract_version,omitempty"`
	Projected              bool                        `json:"projected"`
	Available              bool                        `json:"available"`
	Status                 HostActionStatus            `json:"status,omitempty"`
	ReadyForResolution     bool                        `json:"ready_for_resolution"`
	SelectionRef           DisplaySafeRef              `json:"selection_ref,omitempty"`
	CatalogSnapshotRef     DisplaySafeRef              `json:"catalog_snapshot_ref,omitempty"`
	Producer               DisplaySafeRef              `json:"producer,omitempty"`
	SelectedAdapterRef     DisplaySafeRef              `json:"selected_adapter_ref,omitempty"`
	SelectedDescriptorRef  DisplaySafeRef              `json:"selected_descriptor_ref,omitempty"`
	SelectedDescriptor     ProductionAdapterDescriptor `json:"selected_descriptor,omitempty"`
	SelectedSourceKind     ReplannerSourceKind         `json:"selected_source_kind,omitempty"`
	SelectedSourceRef      DisplaySafeRef              `json:"selected_source_ref,omitempty"`
	SelectedCandidateRef   DisplaySafeRef              `json:"selected_candidate_ref,omitempty"`
	DescriptorRefs         []DisplaySafeRef            `json:"descriptor_refs,omitempty"`
	RequiredCapabilityRefs []DisplaySafeRef            `json:"required_capability_refs,omitempty"`
	RequiredPolicyRefs     []DisplaySafeRef            `json:"required_policy_refs,omitempty"`
	RequiredApprovalRefs   []DisplaySafeRef            `json:"required_approval_refs,omitempty"`
	RequiredBudgetRef      DisplaySafeRef              `json:"required_budget_ref,omitempty"`
	RequiredPreflightRefs  []DisplaySafeRef            `json:"required_preflight_refs,omitempty"`
	MissingInputs          []MissingInput              `json:"missing_inputs,omitempty"`
	BlockedReasons         []string                    `json:"blocked_reasons,omitempty"`
	FailureClass           FailureClass                `json:"failure_class,omitempty"`
	Boundaries             []Boundary                  `json:"boundaries,omitempty"`
	NextHostAction         NextHostAction              `json:"next_host_action,omitempty"`
	RunnerEffect           string                      `json:"runner_effect,omitempty"`
	PromptEffect           string                      `json:"prompt_effect,omitempty"`
	RawOutputLoaded        bool                        `json:"raw_output_loaded"`
}

// agentx-api: internal_candidate
type ProductionAdapterCatalogSelectionResolutionInput struct {
	CatalogSelection         ProductionAdapterCatalogSelection `json:"catalog_selection,omitempty"`
	ApplyEnvelopeReady       bool                              `json:"apply_envelope_ready"`
	ApplyEnvelopeRef         DisplaySafeRef                    `json:"apply_envelope_ref,omitempty"`
	HostPolicyRef            DisplaySafeRef                    `json:"host_policy_ref,omitempty"`
	ApprovalContextRefs      []DisplaySafeRef                  `json:"approval_context_refs,omitempty"`
	BudgetRef                DisplaySafeRef                    `json:"budget_ref,omitempty"`
	IdempotencyRef           DisplaySafeRef                    `json:"idempotency_ref,omitempty"`
	AvailableCapabilityRefs  []DisplaySafeRef                  `json:"available_capability_refs,omitempty"`
	ConfirmedPolicyRefs      []DisplaySafeRef                  `json:"confirmed_policy_refs,omitempty"`
	ConfirmedApprovalRefs    []DisplaySafeRef                  `json:"confirmed_approval_refs,omitempty"`
	AllowAdapterSubstitution bool                              `json:"allow_adapter_substitution"`
	RawOutputLoaded          bool                              `json:"raw_output_loaded"`
}

// agentx-api: internal_candidate
func BuildProductionAdapterCatalogSelection(input ProductionAdapterCatalogSelectionInput) ProductionAdapterCatalogSelection {
	snapshot := input.CatalogSnapshot.Normalize()
	selectedAdapterRef := normalizeOneDisplaySafeRef(input.SelectedAdapterRef)
	selectedDescriptor, selectedDescriptorFound := productionAdapterCatalogSelectionDescriptor(snapshot, selectedAdapterRef)
	result := ProductionAdapterCatalogSelection{
		ContractVersion:        ContractVersion,
		Projected:              true,
		Available:              snapshot.Available,
		Status:                 HostActionBlocked,
		SelectionRef:           normalizeOneDisplaySafeRef(input.SelectionRef),
		CatalogSnapshotRef:     snapshot.CatalogSnapshotRef,
		Producer:               snapshot.Producer,
		SelectedAdapterRef:     selectedAdapterRef,
		SelectedDescriptorRef:  selectedDescriptor.AdapterRef,
		SelectedDescriptor:     selectedDescriptor,
		SelectedSourceKind:     NormalizeReplannerSourceKind(string(input.SelectedSourceKind)),
		SelectedSourceRef:      normalizeOneDisplaySafeRef(input.SelectedSourceRef),
		SelectedCandidateRef:   normalizeOneDisplaySafeRef(input.SelectedCandidateRef),
		DescriptorRefs:         cloneDisplaySafeRefs(snapshot.DescriptorRefs),
		RequiredCapabilityRefs: cloneDisplaySafeRefs(selectedDescriptor.RequiresCapabilityRefs),
		RequiredPolicyRefs:     cloneDisplaySafeRefs(selectedDescriptor.RequiredPolicyRefs),
		RequiredApprovalRefs:   cloneDisplaySafeRefs(selectedDescriptor.RequiredApprovalRefs),
		RequiredBudgetRef:      selectedDescriptor.RequiredBudgetRef,
		RequiredPreflightRefs:  cloneDisplaySafeRefs(selectedDescriptor.PreflightCheckRefs),
		FailureClass:           FailureNone,
		Boundaries:             productionAdapterCatalogSelectionBoundaries(snapshot.Boundaries),
		RunnerEffect:           "none",
		PromptEffect:           "none",
		RawOutputLoaded:        input.RawOutputLoaded || snapshot.RawOutputLoaded,
	}
	if productionAdapterCatalogSelectionInputUnsafe(input) {
		result = productionAdapterCatalogSelectionBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if productionAdapterCatalogSnapshotEmpty(input.CatalogSnapshot) {
		result.Available = false
		result = productionAdapterCatalogSelectionBlock(result, FailureHostAdapterMissing, "adapter_catalog_unavailable", "host:adapter_catalog_snapshot", "provide_adapter_catalog")
		return result.Normalize()
	}
	if !snapshot.ReadyForHostSelection {
		result = productionAdapterCatalogSelectionBlock(result, firstFailureClass(snapshot.FailureClass, FailureConfigMissing), "adapter_catalog_not_ready", "host:adapter_catalog_snapshot", "provide_adapter_catalog")
	}
	if result.SelectionRef == "" {
		result = productionAdapterCatalogSelectionBlock(result, FailureEvidenceMissing, "adapter_selection_ref_missing", "host:adapter_selection_ref", "provide_adapter_selection")
	}
	if result.CatalogSnapshotRef == "" {
		result = productionAdapterCatalogSelectionBlock(result, FailureEvidenceMissing, "catalog_snapshot_ref_missing", "host:adapter_catalog_snapshot_ref", "provide_adapter_catalog")
	}
	if result.Producer == "" {
		result = productionAdapterCatalogSelectionBlock(result, FailureConfigMissing, "adapter_catalog_producer_missing", "host:adapter_catalog_producer", "provide_adapter_catalog")
	}
	if result.SelectedAdapterRef == "" {
		result = productionAdapterCatalogSelectionBlock(result, FailureHostAdapterMissing, "selected_adapter_ref_missing", "host:selected_adapter_ref", "select_adapter_from_catalog")
	} else if !selectedDescriptorFound {
		result = productionAdapterCatalogSelectionBlock(result, FailureHostAdapterMissing, "adapter_not_in_catalog", "host:selected_adapter_ref", "select_adapter_from_catalog")
	}
	for _, check := range productionAdapterDescriptorCompletenessChecks(selectedDescriptor) {
		if selectedDescriptorFound && check.missing {
			result = productionAdapterCatalogSelectionBlock(result, check.failure, check.reason, check.input, "provide_adapter_descriptor")
		}
	}
	if result.SelectedSourceKind == "" {
		result = productionAdapterCatalogSelectionBlock(result, FailureInvalidInput, "source_kind_missing", "host:selected_source_kind", "provide_selected_source")
	} else if selectedDescriptorFound && len(selectedDescriptor.SupportedSourceKinds) > 0 && !replannerSourceKindContains(selectedDescriptor.SupportedSourceKinds, result.SelectedSourceKind) {
		result = productionAdapterCatalogSelectionBlock(result, FailureUnsupportedOperation, "adapter_source_mismatch", "host:adapter_source_review", "review_adapter_source")
	}
	if result.SelectedSourceRef == "" {
		result = productionAdapterCatalogSelectionBlock(result, FailureEvidenceMissing, "source_ref_missing", "host:selected_source_ref", "provide_selected_source")
	}
	if result.SelectedCandidateRef == "" {
		result = productionAdapterCatalogSelectionBlock(result, FailureEvidenceMissing, "candidate_ref_missing", "host:selected_candidate_strategy_ref", "provide_selected_candidate")
	} else if selectedDescriptorFound && len(selectedDescriptor.SupportedCandidateRefs) > 0 && !displaySafeRefSliceContains(selectedDescriptor.SupportedCandidateRefs, result.SelectedCandidateRef) {
		result = productionAdapterCatalogSelectionBlock(result, FailureUnsupportedOperation, "adapter_candidate_mismatch", "host:adapter_candidate_review", "review_adapter_candidate")
	}
	if len(result.BlockedReasons) == 0 && len(result.MissingInputs) == 0 {
		result.Status = HostActionReady
		result.ReadyForResolution = true
		result.NextHostAction = "build_adapter_resolution"
		result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_adapter_resolution")
	}
	return result.Normalize()
}

// agentx-api: internal_candidate
func BuildProductionAdapterResolutionInputFromCatalogSelection(input ProductionAdapterCatalogSelectionResolutionInput) ProductionAdapterResolutionInput {
	selection := input.CatalogSelection.Normalize()
	return ProductionAdapterResolutionInput{
		ApplyEnvelopeReady:       input.ApplyEnvelopeReady && selection.ReadyForResolution,
		ApplyEnvelopeRef:         input.ApplyEnvelopeRef,
		SelectedSourceKind:       selection.SelectedSourceKind,
		SelectedSourceRef:        selection.SelectedSourceRef,
		SelectedCandidateRef:     selection.SelectedCandidateRef,
		RequestedAdapterRef:      selection.SelectedAdapterRef,
		CatalogSnapshotRef:       selection.CatalogSnapshotRef,
		CatalogSelectionRef:      selection.SelectionRef,
		CatalogSelection:         selection,
		HostPolicyRef:            input.HostPolicyRef,
		ApprovalContextRefs:      cloneDisplaySafeRefs(input.ApprovalContextRefs),
		BudgetRef:                input.BudgetRef,
		IdempotencyRef:           input.IdempotencyRef,
		AvailableCapabilityRefs:  cloneDisplaySafeRefs(input.AvailableCapabilityRefs),
		ConfirmedPolicyRefs:      cloneDisplaySafeRefs(input.ConfirmedPolicyRefs),
		ConfirmedApprovalRefs:    cloneDisplaySafeRefs(input.ConfirmedApprovalRefs),
		Descriptor:               selection.SelectedDescriptor,
		AllowAdapterSubstitution: input.AllowAdapterSubstitution,
		RawOutputLoaded:          input.RawOutputLoaded || selection.RawOutputLoaded,
	}
}

func CloneProductionAdapterCatalogSelection(in ProductionAdapterCatalogSelection) ProductionAdapterCatalogSelection {
	out := in
	out.SelectedDescriptor = in.SelectedDescriptor.Clone()
	out.DescriptorRefs = cloneDisplaySafeRefs(in.DescriptorRefs)
	out.RequiredCapabilityRefs = cloneDisplaySafeRefs(in.RequiredCapabilityRefs)
	out.RequiredPolicyRefs = cloneDisplaySafeRefs(in.RequiredPolicyRefs)
	out.RequiredApprovalRefs = cloneDisplaySafeRefs(in.RequiredApprovalRefs)
	out.RequiredPreflightRefs = cloneDisplaySafeRefs(in.RequiredPreflightRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (s ProductionAdapterCatalogSelection) Clone() ProductionAdapterCatalogSelection {
	return CloneProductionAdapterCatalogSelection(s)
}

func (s ProductionAdapterCatalogSelection) Normalize() ProductionAdapterCatalogSelection {
	out := CloneProductionAdapterCatalogSelection(s)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.SelectionRef = normalizeOneDisplaySafeRef(out.SelectionRef)
	out.CatalogSnapshotRef = normalizeOneDisplaySafeRef(out.CatalogSnapshotRef)
	out.Producer = normalizeOneDisplaySafeRef(out.Producer)
	out.SelectedAdapterRef = normalizeOneDisplaySafeRef(out.SelectedAdapterRef)
	out.SelectedDescriptorRef = normalizeOneDisplaySafeRef(out.SelectedDescriptorRef)
	out.SelectedDescriptor = out.SelectedDescriptor.Normalize()
	out.SelectedSourceKind = NormalizeReplannerSourceKind(string(out.SelectedSourceKind))
	out.SelectedSourceRef = normalizeOneDisplaySafeRef(out.SelectedSourceRef)
	out.SelectedCandidateRef = normalizeOneDisplaySafeRef(out.SelectedCandidateRef)
	out.DescriptorRefs = normalizeDisplaySafeRefs(out.DescriptorRefs)
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
	out.ReadyForResolution = out.Available &&
		out.Status == HostActionReady &&
		out.SelectionRef != "" &&
		out.CatalogSnapshotRef != "" &&
		out.Producer != "" &&
		out.SelectedAdapterRef != "" &&
		out.SelectedDescriptor.AdapterRef == out.SelectedAdapterRef &&
		out.SelectedSourceKind != "" &&
		out.SelectedSourceRef != "" &&
		out.SelectedCandidateRef != "" &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.RawOutputLoaded
	return out
}

func productionAdapterCatalogSelectionBlock(result ProductionAdapterCatalogSelection, failure FailureClass, reason string, missing MissingInput, next NextHostAction) ProductionAdapterCatalogSelection {
	result.Status = HostActionBlocked
	result.ReadyForResolution = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	return result
}

func productionAdapterCatalogSelectionDescriptor(snapshot ProductionAdapterCatalogSnapshot, adapterRef DisplaySafeRef) (ProductionAdapterDescriptor, bool) {
	if adapterRef == "" {
		return ProductionAdapterDescriptor{}, false
	}
	for _, descriptor := range snapshot.Descriptors {
		normalized := descriptor.Normalize()
		if normalized.AdapterRef == adapterRef {
			return normalized, true
		}
	}
	return ProductionAdapterDescriptor{}, false
}

func productionAdapterCatalogSelectionInputUnsafe(input ProductionAdapterCatalogSelectionInput) bool {
	return input.RawOutputLoaded ||
		displaySafeRefRejected(input.SelectionRef) ||
		displaySafeRefRejected(input.SelectedAdapterRef) ||
		displaySafeRefRejected(input.SelectedSourceRef) ||
		displaySafeRefRejected(input.SelectedCandidateRef) ||
		productionAdapterCatalogSnapshotUnsafe(input.CatalogSnapshot)
}

func productionAdapterCatalogSelectionUnsafe(input ProductionAdapterCatalogSelection) bool {
	return input.RawOutputLoaded ||
		displaySafeRefRejected(input.SelectionRef) ||
		displaySafeRefRejected(input.CatalogSnapshotRef) ||
		displaySafeRefRejected(input.Producer) ||
		displaySafeRefRejected(input.SelectedAdapterRef) ||
		displaySafeRefRejected(input.SelectedDescriptorRef) ||
		productionAdapterDescriptorUnsafe(input.SelectedDescriptor) ||
		displaySafeRefRejected(input.SelectedSourceRef) ||
		displaySafeRefRejected(input.SelectedCandidateRef) ||
		displaySafeRefSliceRejected(input.DescriptorRefs) ||
		displaySafeRefSliceRejected(input.RequiredCapabilityRefs) ||
		displaySafeRefSliceRejected(input.RequiredPolicyRefs) ||
		displaySafeRefSliceRejected(input.RequiredApprovalRefs) ||
		displaySafeRefRejected(input.RequiredBudgetRef) ||
		displaySafeRefSliceRejected(input.RequiredPreflightRefs)
}

func productionAdapterCatalogSelectionEmpty(input ProductionAdapterCatalogSelection) bool {
	return !input.Projected &&
		!input.Available &&
		input.Status == "" &&
		input.SelectionRef == "" &&
		input.CatalogSnapshotRef == "" &&
		input.SelectedAdapterRef == "" &&
		input.SelectedDescriptorRef == "" &&
		input.SelectedDescriptor.AdapterRef == "" &&
		input.SelectedSourceKind == "" &&
		input.SelectedSourceRef == "" &&
		input.SelectedCandidateRef == "" &&
		len(input.MissingInputs) == 0 &&
		len(input.BlockedReasons) == 0 &&
		len(input.Boundaries) == 0 &&
		input.NextHostAction == "" &&
		!input.RawOutputLoaded
}

func productionAdapterCatalogSelectionBoundaries(extra []Boundary) []Boundary {
	return MergeBoundaries([]Boundary{
		"production_adapter_catalog_selection",
		"catalog_selection_projection_only",
		"host_owned_adapter_selection",
		"catalog_bound_adapter_selection",
		"display_safe_refs_only",
		"no_adapter_invocation",
		"no_runner_dispatch",
	}, extra)
}
