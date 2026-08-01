package controlcontract

type HostAdapterRegistryInput struct {
	RegistryRef        DisplaySafeRef                    `json:"registry_ref,omitempty"`
	ProductionRegistry ProductionAdapterRegistrySnapshot `json:"production_registry,omitempty"`
	Descriptors        []ProductionAdapterDescriptor     `json:"descriptors,omitempty"`
	Entries            []HostAdapterRegistryEntry        `json:"entries,omitempty"`
	PolicyRefs         []DisplaySafeRef                  `json:"policy_refs,omitempty"`
	EvidenceRefs       []EvidenceRef                     `json:"evidence_refs,omitempty"`
	Boundaries         []Boundary                        `json:"boundaries,omitempty"`
	RawOutputLoaded    bool                              `json:"raw_output_loaded"`
}

type HostAdapterRegistryEntry struct {
	ContractVersion          string                      `json:"contract_version,omitempty"`
	Projected                bool                        `json:"projected"`
	AdapterRef               DisplaySafeRef              `json:"adapter_ref,omitempty"`
	Descriptor               ProductionAdapterDescriptor `json:"descriptor,omitempty"`
	SupportedStrategyRefs    []DisplaySafeRef            `json:"supported_strategy_refs,omitempty"`
	SupportedControlModes    []ControlMode               `json:"supported_control_modes,omitempty"`
	SupportedIntensities     []ExecutionIntensity        `json:"supported_intensities,omitempty"`
	ProvidesCapabilityRefs   []DisplaySafeRef            `json:"provides_capability_refs,omitempty"`
	RequiresCapabilityRefs   []DisplaySafeRef            `json:"requires_capability_refs,omitempty"`
	RequiredPolicyRefs       []DisplaySafeRef            `json:"required_policy_refs,omitempty"`
	RequiredApprovalRefs     []DisplaySafeRef            `json:"required_approval_refs,omitempty"`
	RequiredBudgetRef        DisplaySafeRef              `json:"required_budget_ref,omitempty"`
	IdempotencyContractRef   DisplaySafeRef              `json:"idempotency_contract_ref,omitempty"`
	SideEffectClass          string                      `json:"side_effect_class,omitempty"`
	ReadOnly                 bool                        `json:"read_only"`
	ExpectedObservationKinds []string                    `json:"expected_observation_kinds,omitempty"`
	MissingInputs            []MissingInput              `json:"missing_inputs,omitempty"`
	Boundaries               []Boundary                  `json:"boundaries,omitempty"`
	RunnerEffect             string                      `json:"runner_effect,omitempty"`
	PromptEffect             string                      `json:"prompt_effect,omitempty"`
	RawOutputLoaded          bool                        `json:"raw_output_loaded"`
}

type HostAdapterRegistrySnapshot struct {
	ContractVersion        string                     `json:"contract_version,omitempty"`
	Projected              bool                       `json:"projected"`
	Available              bool                       `json:"available"`
	Status                 HostActionStatus           `json:"status,omitempty"`
	ReadyForRuntimeRequest bool                       `json:"ready_for_runtime_request"`
	RegistryRef            DisplaySafeRef             `json:"registry_ref,omitempty"`
	AdapterRefs            []DisplaySafeRef           `json:"adapter_refs,omitempty"`
	StrategyRefs           []DisplaySafeRef           `json:"strategy_refs,omitempty"`
	PolicyRefs             []DisplaySafeRef           `json:"policy_refs,omitempty"`
	EvidenceRefs           []EvidenceRef              `json:"evidence_refs,omitempty"`
	Entries                []HostAdapterRegistryEntry `json:"entries,omitempty"`
	MissingInputs          []MissingInput             `json:"missing_inputs,omitempty"`
	FailureClass           FailureClass               `json:"failure_class,omitempty"`
	Boundaries             []Boundary                 `json:"boundaries,omitempty"`
	NextHostAction         NextHostAction             `json:"next_host_action,omitempty"`
	RunnerEffect           string                     `json:"runner_effect,omitempty"`
	PromptEffect           string                     `json:"prompt_effect,omitempty"`
	RawOutputLoaded        bool                       `json:"raw_output_loaded"`
}

func BuildHostAdapterRegistry(input HostAdapterRegistryInput) HostAdapterRegistrySnapshot {
	productionRegistry := input.ProductionRegistry.Normalize()
	registryRef := firstDisplaySafeRef(input.RegistryRef, productionRegistry.RegistrySnapshotRef)
	entries := cloneHostAdapterRegistryEntries(input.Entries)
	for _, descriptor := range productionRegistry.CatalogSnapshot.Descriptors {
		entries = append(entries, hostAdapterRegistryEntryFromDescriptor(descriptor))
	}
	for _, descriptor := range input.Descriptors {
		entries = append(entries, hostAdapterRegistryEntryFromDescriptor(descriptor))
	}
	result := HostAdapterRegistrySnapshot{
		ContractVersion: ContractVersion,
		Projected:       true,
		Available:       true,
		Status:          HostActionBlocked,
		RegistryRef:     normalizeOneDisplaySafeRef(registryRef),
		Entries:         normalizeHostAdapterRegistryEntries(entries),
		PolicyRefs: normalizeDisplaySafeRefs(append(
			cloneDisplaySafeRefs(input.PolicyRefs),
			productionRegistry.CatalogSnapshot.PolicyRefs...,
		)),
		EvidenceRefs: normalizeEvidenceRefs(input.EvidenceRefs),
		FailureClass: FailureNone,
		Boundaries: MergeBoundaries(
			[]Boundary{
				"host_adapter_registry",
				"objective_loop_runtime_adapter_registry",
				"host_owned_adapter_metadata",
				"projection_only",
				"no_adapter_invocation",
				"no_runner_dispatch",
			},
			input.Boundaries,
			productionRegistry.Boundaries,
		),
		RunnerEffect:    "none",
		PromptEffect:    "none",
		RawOutputLoaded: input.RawOutputLoaded || productionRegistry.RawOutputLoaded,
	}
	if hostAdapterRegistryInputUnsafe(input) {
		result = hostAdapterRegistryBlock(result, FailureEvidenceWeak, "host:display_safe_refs", "provide_display_safe_refs", "raw_output_not_allowed")
		return result.Normalize()
	}
	if result.RegistryRef == "" {
		result = hostAdapterRegistryBlock(result, FailureEvidenceMissing, "host:adapter_registry_ref", "provide_adapter_registry", "host_adapter_registry_ref_missing")
	}
	if len(result.Entries) == 0 {
		result = hostAdapterRegistryBlock(result, FailureHostAdapterMissing, "host:adapter_registry_entries", "provide_adapter_registry", "host_adapter_registry_empty")
	}
	for _, entry := range result.Entries {
		result.AdapterRefs = appendUniqueDisplaySafeRef(result.AdapterRefs, entry.AdapterRef)
		result.StrategyRefs = normalizeDisplaySafeRefs(append(result.StrategyRefs, entry.SupportedStrategyRefs...))
		result.PolicyRefs = normalizeDisplaySafeRefs(append(result.PolicyRefs, entry.RequiredPolicyRefs...))
		result.EvidenceRefs = MergeEvidenceRefs(result.EvidenceRefs)
		result.MissingInputs = AppendMissingInputs(result.MissingInputs, entry.MissingInputs...)
		result.RawOutputLoaded = result.RawOutputLoaded || entry.RawOutputLoaded
	}
	if len(result.MissingInputs) > 0 && result.FailureClass == FailureNone {
		result.FailureClass = FailureConfigMissing
	}
	if len(result.MissingInputs) == 0 && len(result.Entries) > 0 && result.RegistryRef != "" {
		result.Status = HostActionReady
		result.ReadyForRuntimeRequest = true
		result.NextHostAction = "host_may_build_runtime_adapter_request"
		result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_runtime_adapter_request")
	}
	return result.Normalize()
}

func CloneHostAdapterRegistryEntry(in HostAdapterRegistryEntry) HostAdapterRegistryEntry {
	out := in
	out.Descriptor = in.Descriptor.Clone()
	out.SupportedStrategyRefs = cloneDisplaySafeRefs(in.SupportedStrategyRefs)
	out.SupportedControlModes = cloneControlModes(in.SupportedControlModes)
	out.SupportedIntensities = cloneExecutionIntensities(in.SupportedIntensities)
	out.ProvidesCapabilityRefs = cloneDisplaySafeRefs(in.ProvidesCapabilityRefs)
	out.RequiresCapabilityRefs = cloneDisplaySafeRefs(in.RequiresCapabilityRefs)
	out.RequiredPolicyRefs = cloneDisplaySafeRefs(in.RequiredPolicyRefs)
	out.RequiredApprovalRefs = cloneDisplaySafeRefs(in.RequiredApprovalRefs)
	out.ExpectedObservationKinds = cloneStringSlice(in.ExpectedObservationKinds)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (e HostAdapterRegistryEntry) Clone() HostAdapterRegistryEntry {
	return CloneHostAdapterRegistryEntry(e)
}

func (e HostAdapterRegistryEntry) Normalize() HostAdapterRegistryEntry {
	out := CloneHostAdapterRegistryEntry(e)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Descriptor = out.Descriptor.Normalize()
	out.AdapterRef = firstDisplaySafeRef(out.AdapterRef, out.Descriptor.AdapterRef)
	out.SupportedStrategyRefs = normalizeDisplaySafeRefs(append(out.SupportedStrategyRefs, out.Descriptor.SupportedCandidateRefs...))
	out.SupportedControlModes = normalizeControlModes(out.SupportedControlModes)
	out.SupportedIntensities = normalizeExecutionIntensities(out.SupportedIntensities)
	out.ProvidesCapabilityRefs = normalizeDisplaySafeRefs(append(out.ProvidesCapabilityRefs, out.Descriptor.ProvidesCapabilityRefs...))
	out.RequiresCapabilityRefs = normalizeDisplaySafeRefs(append(out.RequiresCapabilityRefs, out.Descriptor.RequiresCapabilityRefs...))
	out.RequiredPolicyRefs = normalizeDisplaySafeRefs(append(out.RequiredPolicyRefs, out.Descriptor.RequiredPolicyRefs...))
	out.RequiredApprovalRefs = normalizeDisplaySafeRefs(append(out.RequiredApprovalRefs, out.Descriptor.RequiredApprovalRefs...))
	out.RequiredBudgetRef = firstDisplaySafeRef(out.RequiredBudgetRef, out.Descriptor.RequiredBudgetRef)
	out.IdempotencyContractRef = firstDisplaySafeRef(out.IdempotencyContractRef, out.Descriptor.IdempotencyContractRef)
	out.SideEffectClass = firstNonEmptyControlToken(out.SideEffectClass, out.Descriptor.SideEffectClass)
	out.ReadOnly = out.ReadOnly || runtimeAdapterReadOnlyKind(out.Descriptor.Kind) || runtimeAdapterReadOnlySideEffectClass(out.SideEffectClass)
	out.ExpectedObservationKinds = normalizeControlTokenList(out.ExpectedObservationKinds)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.Boundaries = MergeBoundaries(
		[]Boundary{
			"host_adapter_registry_entry",
			"host_owned_adapter_metadata",
			"no_adapter_invocation",
		},
		out.Descriptor.Boundaries,
		out.Boundaries,
	)
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	if out.RunnerEffect == "" {
		out.RunnerEffect = "none"
	}
	if out.PromptEffect == "" {
		out.PromptEffect = "none"
	}
	if out.AdapterRef == "" {
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:adapter_ref")
	}
	if len(out.SupportedStrategyRefs) == 0 {
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:adapter_supported_strategy_ref")
	}
	if out.SideEffectClass == "" {
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:adapter_side_effect_class")
	}
	if out.RawOutputLoaded || out.Descriptor.RawOutputLoaded {
		out.RawOutputLoaded = true
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
	}
	return out
}

func CloneHostAdapterRegistrySnapshot(in HostAdapterRegistrySnapshot) HostAdapterRegistrySnapshot {
	out := in
	out.AdapterRefs = cloneDisplaySafeRefs(in.AdapterRefs)
	out.StrategyRefs = cloneDisplaySafeRefs(in.StrategyRefs)
	out.PolicyRefs = cloneDisplaySafeRefs(in.PolicyRefs)
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.Entries = cloneHostAdapterRegistryEntries(in.Entries)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (s HostAdapterRegistrySnapshot) Clone() HostAdapterRegistrySnapshot {
	return CloneHostAdapterRegistrySnapshot(s)
}

func (s HostAdapterRegistrySnapshot) Normalize() HostAdapterRegistrySnapshot {
	out := CloneHostAdapterRegistrySnapshot(s)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.RegistryRef = normalizeOneDisplaySafeRef(out.RegistryRef)
	out.AdapterRefs = normalizeDisplaySafeRefs(out.AdapterRefs)
	out.StrategyRefs = normalizeDisplaySafeRefs(out.StrategyRefs)
	out.PolicyRefs = normalizeDisplaySafeRefs(out.PolicyRefs)
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
	out.Entries = normalizeHostAdapterRegistryEntries(out.Entries)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
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
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		if out.NextHostAction == "" {
			out.NextHostAction = "provide_display_safe_refs"
		}
	}
	out.ReadyForRuntimeRequest = out.Status == HostActionReady &&
		out.RegistryRef != "" &&
		len(out.Entries) > 0 &&
		len(out.MissingInputs) == 0 &&
		!out.RawOutputLoaded
	return out
}

type RuntimeAdapterExecutionRequestInput struct {
	Activation                 Activation                  `json:"activation,omitempty"`
	Frame                      ObjectiveFrame              `json:"frame,omitempty"`
	Selected                   StrategyPlanCandidate       `json:"selected,omitempty"`
	FinalGate                  IntensityGateResult         `json:"final_gate,omitempty"`
	Registry                   HostAdapterRegistrySnapshot `json:"registry,omitempty"`
	RequestedAdapterRef        DisplaySafeRef              `json:"requested_adapter_ref,omitempty"`
	Budget                     ObjectiveBudgetSnapshot     `json:"budget,omitempty"`
	ApprovalRefs               []DisplaySafeRef            `json:"approval_refs,omitempty"`
	AllowHostSideEffectAdapter bool                        `json:"allow_host_side_effect_adapter"`
	HostSideEffectApprovalRefs []DisplaySafeRef            `json:"host_side_effect_approval_refs,omitempty"`
	PolicyRefs                 []DisplaySafeRef            `json:"policy_refs,omitempty"`
	AvailableCapabilityRefs    []DisplaySafeRef            `json:"available_capability_refs,omitempty"`
	IdempotencyRef             DisplaySafeRef              `json:"idempotency_ref,omitempty"`
	InputRefs                  []DisplaySafeRef            `json:"input_refs,omitempty"`
	ExpectedObservationKinds   []string                    `json:"expected_observation_kinds,omitempty"`
	Boundaries                 []Boundary                  `json:"boundaries,omitempty"`
	RawOutputLoaded            bool                        `json:"raw_output_loaded"`
}

type RuntimeAdapterExecutionRequest struct {
	ContractVersion              string                      `json:"contract_version,omitempty"`
	Projected                    bool                        `json:"projected"`
	Activation                   Activation                  `json:"activation,omitempty"`
	Status                       HostActionStatus            `json:"status,omitempty"`
	ReadyForHostExecution        bool                        `json:"ready_for_host_execution"`
	RegistryRef                  DisplaySafeRef              `json:"registry_ref,omitempty"`
	AdapterRef                   DisplaySafeRef              `json:"adapter_ref,omitempty"`
	Descriptor                   ProductionAdapterDescriptor `json:"descriptor,omitempty"`
	StrategyRef                  DisplaySafeRef              `json:"strategy_ref,omitempty"`
	Frame                        ObjectiveFrame              `json:"frame,omitempty"`
	Strategy                     StrategyCandidate           `json:"strategy,omitempty"`
	FinalGate                    IntensityGateResult         `json:"final_gate,omitempty"`
	Budget                       ObjectiveBudgetSnapshot     `json:"budget,omitempty"`
	ApprovalRefs                 []DisplaySafeRef            `json:"approval_refs,omitempty"`
	HostSideEffectAdapterAllowed bool                        `json:"host_side_effect_adapter_allowed"`
	HostSideEffectApprovalRefs   []DisplaySafeRef            `json:"host_side_effect_approval_refs,omitempty"`
	PolicyRefs                   []DisplaySafeRef            `json:"policy_refs,omitempty"`
	RequiredCapabilityRefs       []DisplaySafeRef            `json:"required_capability_refs,omitempty"`
	RequiredPolicyRefs           []DisplaySafeRef            `json:"required_policy_refs,omitempty"`
	RequiredApprovalRefs         []DisplaySafeRef            `json:"required_approval_refs,omitempty"`
	RequiredBudgetRef            DisplaySafeRef              `json:"required_budget_ref,omitempty"`
	IdempotencyRef               DisplaySafeRef              `json:"idempotency_ref,omitempty"`
	IdempotencyContractRef       DisplaySafeRef              `json:"idempotency_contract_ref,omitempty"`
	InputRefs                    []DisplaySafeRef            `json:"input_refs,omitempty"`
	ExpectedObservationKinds     []string                    `json:"expected_observation_kinds,omitempty"`
	FailureClass                 FailureClass                `json:"failure_class,omitempty"`
	MissingInputs                []MissingInput              `json:"missing_inputs,omitempty"`
	Boundaries                   []Boundary                  `json:"boundaries,omitempty"`
	NextHostAction               NextHostAction              `json:"next_host_action,omitempty"`
	RunnerEffect                 string                      `json:"runner_effect,omitempty"`
	PromptEffect                 string                      `json:"prompt_effect,omitempty"`
	RawOutputLoaded              bool                        `json:"raw_output_loaded"`
}

func BuildRuntimeAdapterExecutionRequest(input RuntimeAdapterExecutionRequestInput) RuntimeAdapterExecutionRequest {
	registry := input.Registry.Normalize()
	selected := input.Selected.Normalize()
	finalGate := input.FinalGate.Normalize()
	frame := input.Frame.Normalize()
	budget := input.Budget.Normalize()
	activation := intensityGateActivation(input.Activation, finalGate.Activation)
	strategyRef := normalizeOneDisplaySafeRef(DisplaySafeRef(selected.Candidate.ID))
	entry, hasEntry := selectRuntimeAdapterRegistryEntry(registry, input.RequestedAdapterRef, strategyRef)
	result := RuntimeAdapterExecutionRequest{
		ContractVersion:              ContractVersion,
		Projected:                    true,
		Activation:                   activation,
		Status:                       HostActionBlocked,
		RegistryRef:                  registry.RegistryRef,
		AdapterRef:                   entry.AdapterRef,
		Descriptor:                   entry.Descriptor,
		StrategyRef:                  strategyRef,
		Frame:                        frame,
		Strategy:                     selected.Candidate,
		FinalGate:                    finalGate,
		Budget:                       budget,
		ApprovalRefs:                 normalizeDisplaySafeRefs(append(cloneDisplaySafeRefs(input.ApprovalRefs), finalGate.ApprovalRefs...)),
		HostSideEffectAdapterAllowed: input.AllowHostSideEffectAdapter,
		HostSideEffectApprovalRefs:   normalizeDisplaySafeRefs(input.HostSideEffectApprovalRefs),
		PolicyRefs:                   normalizeDisplaySafeRefs(append(cloneDisplaySafeRefs(input.PolicyRefs), finalGate.PolicyRefs...)),
		RequiredCapabilityRefs:       cloneDisplaySafeRefs(entry.RequiresCapabilityRefs),
		RequiredPolicyRefs:           cloneDisplaySafeRefs(entry.RequiredPolicyRefs),
		RequiredApprovalRefs:         cloneDisplaySafeRefs(entry.RequiredApprovalRefs),
		RequiredBudgetRef:            entry.RequiredBudgetRef,
		IdempotencyRef:               normalizeOneDisplaySafeRef(input.IdempotencyRef),
		IdempotencyContractRef:       entry.IdempotencyContractRef,
		InputRefs:                    normalizeDisplaySafeRefs(input.InputRefs),
		ExpectedObservationKinds: normalizeControlTokenList(append(
			cloneStringSlice(input.ExpectedObservationKinds),
			entry.ExpectedObservationKinds...,
		)),
		FailureClass: FailureNone,
		Boundaries: MergeBoundaries(
			[]Boundary{
				"runtime_adapter_execution_request",
				"host_owned_runtime_adapter",
				"final_gate_bound_adapter_request",
				"projection_only",
				"no_adapter_invocation",
				"no_runner_dispatch",
				"core_does_not_execute_adapter",
			},
			input.Boundaries,
			registry.Boundaries,
			finalGate.Boundaries,
			entry.Boundaries,
		),
		RunnerEffect:    "none",
		PromptEffect:    "none",
		RawOutputLoaded: input.RawOutputLoaded || registry.RawOutputLoaded || finalGate.RawOutputLoaded || entry.RawOutputLoaded,
	}
	if runtimeAdapterExecutionRequestUnsafe(input) {
		return runtimeAdapterRequestBlock(result, FailureEvidenceWeak, "host:display_safe_refs", "provide_display_safe_refs", "raw_output_not_allowed")
	}
	if activation != ActivationManaged {
		return runtimeAdapterRequestBlock(result, FailurePolicyBlocked, "control_plane:managed_activation", "enable_managed_objective", "runtime_adapter_requires_managed_activation")
	}
	if !finalGate.Allowed {
		return runtimeAdapterRequestBlock(result, firstFailureClass(finalGate.FailureClass, FailurePolicyBlocked), "control_plane:strategy_final_gate", "run_strategy_final_gate", "runtime_adapter_requires_final_gate")
	}
	if !registry.ReadyForRuntimeRequest {
		return runtimeAdapterRequestBlock(result, firstFailureClass(registry.FailureClass, FailureHostAdapterMissing), "host:adapter_registry", "provide_adapter_registry", "runtime_adapter_registry_not_ready")
	}
	if strategyRef == "" {
		return runtimeAdapterRequestBlock(result, FailureConfigMissing, "host:strategy_candidate", "provide_strategy_candidate", "runtime_adapter_strategy_missing")
	}
	if finalGate.StrategyRef != "" && finalGate.StrategyRef != strategyRef {
		return runtimeAdapterRequestBlock(result, FailureVerificationFailed, "host:strategy_final_gate", "run_strategy_final_gate", "runtime_adapter_strategy_gate_mismatch")
	}
	if !hasEntry {
		return runtimeAdapterRequestBlock(result, FailureHostAdapterMissing, "host:runtime_adapter", "provide_runtime_adapter", "runtime_adapter_missing")
	}
	if !displaySafeRefSliceContains(entry.SupportedStrategyRefs, strategyRef) {
		return runtimeAdapterRequestBlock(result, FailureUnsupportedOperation, "host:runtime_adapter_strategy", "select_matching_runtime_adapter", "runtime_adapter_strategy_mismatch")
	}
	if !entry.ReadOnly {
		if !input.AllowHostSideEffectAdapter {
			return runtimeAdapterRequestBlock(result, FailurePolicyBlocked, "contract:read_only_runtime_adapter", "request_host_approval", "runtime_adapter_non_read_only_not_enabled")
		}
		result.Boundaries = AppendBoundaries(result.Boundaries, "host_side_effect_adapter_explicitly_allowed")
		if len(result.HostSideEffectApprovalRefs) == 0 {
			result = runtimeAdapterRequestBlock(result, FailureApprovalRequired, "host:side_effect_adapter_approval_ref", "request_host_approval", "runtime_adapter_side_effect_approval_ref_missing")
		}
		for _, approvalRef := range result.HostSideEffectApprovalRefs {
			if !displaySafeRefSliceContains(result.ApprovalRefs, approvalRef) {
				result = runtimeAdapterRequestBlock(result, FailureApprovalRequired, MissingInput(approvalRef), "request_host_approval", "runtime_adapter_side_effect_approval_ref_not_confirmed")
			}
		}
	}
	for _, required := range entry.RequiresCapabilityRefs {
		if !displaySafeRefSliceContains(input.AvailableCapabilityRefs, required) {
			result.Boundaries = AppendBoundaries(result.Boundaries, "runtime_adapter_capability_missing")
			result = runtimeAdapterRequestBlock(result, FailureCapabilityMissing, MissingInput(required), "enter_capability_resolution", "capability_gap_proposal_only")
		}
	}
	for _, required := range entry.RequiredPolicyRefs {
		if !displaySafeRefSliceContains(result.PolicyRefs, required) {
			result = runtimeAdapterRequestBlock(result, FailurePolicyBlocked, MissingInput(required), "provide_adapter_policy", "runtime_adapter_policy_missing")
		}
	}
	for _, required := range entry.RequiredApprovalRefs {
		if !displaySafeRefSliceContains(result.ApprovalRefs, required) {
			result = runtimeAdapterRequestBlock(result, FailureApprovalRequired, MissingInput(required), "request_host_approval", "runtime_adapter_approval_missing")
		}
	}
	if entry.RequiredBudgetRef != "" {
		switch {
		case budget.BudgetRef == "":
			result = runtimeAdapterRequestBlock(result, FailureBudgetExhausted, MissingInput(entry.RequiredBudgetRef), "provide_adapter_budget", "runtime_adapter_budget_missing")
		case budget.BudgetRef != entry.RequiredBudgetRef:
			result = runtimeAdapterRequestBlock(result, FailurePolicyBlocked, MissingInput(entry.RequiredBudgetRef), "provide_adapter_budget", "runtime_adapter_budget_mismatch")
		}
	}
	if result.IdempotencyRef == "" {
		result = runtimeAdapterRequestBlock(result, FailureInvalidInput, "host:idempotency_ref", "provide_idempotency_ref", "runtime_adapter_idempotency_missing")
	}
	if len(result.InputRefs) == 0 {
		result = runtimeAdapterRequestBlock(result, FailureEvidenceMissing, "host:runtime_adapter_input_ref", "provide_runtime_adapter_input_refs", "runtime_adapter_input_ref_missing")
	}
	if len(result.MissingInputs) == 0 {
		result.Status = HostActionReady
		result.ReadyForHostExecution = true
		result.NextHostAction = "host_may_execute_runtime_adapter"
		result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_host_runtime_adapter_execution")
	}
	return result.Normalize()
}

func CloneRuntimeAdapterExecutionRequest(in RuntimeAdapterExecutionRequest) RuntimeAdapterExecutionRequest {
	out := in
	out.Descriptor = in.Descriptor.Clone()
	out.Frame = in.Frame.Clone()
	out.Strategy = in.Strategy.Clone()
	out.FinalGate = in.FinalGate.Clone()
	out.Budget = in.Budget.Clone()
	out.ApprovalRefs = cloneDisplaySafeRefs(in.ApprovalRefs)
	out.HostSideEffectApprovalRefs = cloneDisplaySafeRefs(in.HostSideEffectApprovalRefs)
	out.PolicyRefs = cloneDisplaySafeRefs(in.PolicyRefs)
	out.RequiredCapabilityRefs = cloneDisplaySafeRefs(in.RequiredCapabilityRefs)
	out.RequiredPolicyRefs = cloneDisplaySafeRefs(in.RequiredPolicyRefs)
	out.RequiredApprovalRefs = cloneDisplaySafeRefs(in.RequiredApprovalRefs)
	out.InputRefs = cloneDisplaySafeRefs(in.InputRefs)
	out.ExpectedObservationKinds = cloneStringSlice(in.ExpectedObservationKinds)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (r RuntimeAdapterExecutionRequest) Clone() RuntimeAdapterExecutionRequest {
	return CloneRuntimeAdapterExecutionRequest(r)
}

func (r RuntimeAdapterExecutionRequest) Normalize() RuntimeAdapterExecutionRequest {
	out := CloneRuntimeAdapterExecutionRequest(r)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Activation = NormalizeActivation(string(out.Activation))
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.RegistryRef = normalizeOneDisplaySafeRef(out.RegistryRef)
	out.AdapterRef = normalizeOneDisplaySafeRef(out.AdapterRef)
	out.Descriptor = out.Descriptor.Normalize()
	out.StrategyRef = normalizeOneDisplaySafeRef(out.StrategyRef)
	out.Frame = out.Frame.Normalize()
	out.Strategy = out.Strategy.Normalize()
	out.FinalGate = out.FinalGate.Normalize()
	out.Budget = out.Budget.Normalize()
	out.ApprovalRefs = normalizeDisplaySafeRefs(out.ApprovalRefs)
	out.HostSideEffectApprovalRefs = normalizeDisplaySafeRefs(out.HostSideEffectApprovalRefs)
	out.PolicyRefs = normalizeDisplaySafeRefs(out.PolicyRefs)
	out.RequiredCapabilityRefs = normalizeDisplaySafeRefs(out.RequiredCapabilityRefs)
	out.RequiredPolicyRefs = normalizeDisplaySafeRefs(out.RequiredPolicyRefs)
	out.RequiredApprovalRefs = normalizeDisplaySafeRefs(out.RequiredApprovalRefs)
	out.RequiredBudgetRef = normalizeOneDisplaySafeRef(out.RequiredBudgetRef)
	out.IdempotencyRef = normalizeOneDisplaySafeRef(out.IdempotencyRef)
	out.IdempotencyContractRef = normalizeOneDisplaySafeRef(out.IdempotencyContractRef)
	out.InputRefs = normalizeDisplaySafeRefs(out.InputRefs)
	out.ExpectedObservationKinds = normalizeControlTokenList(out.ExpectedObservationKinds)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
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
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		if out.NextHostAction == "" {
			out.NextHostAction = "provide_display_safe_refs"
		}
	}
	out.ReadyForHostExecution = out.Status == HostActionReady &&
		out.AdapterRef != "" &&
		out.StrategyRef != "" &&
		out.IdempotencyRef != "" &&
		len(out.InputRefs) > 0 &&
		len(out.MissingInputs) == 0 &&
		!out.RawOutputLoaded
	return out
}

type RuntimeAdapterExecutionResultInput struct {
	Request               RuntimeAdapterExecutionRequest `json:"request,omitempty"`
	AdapterRef            DisplaySafeRef                 `json:"adapter_ref,omitempty"`
	StrategyRef           DisplaySafeRef                 `json:"strategy_ref,omitempty"`
	HostAdapterRunRef     DisplaySafeRef                 `json:"host_adapter_run_ref,omitempty"`
	Status                VerificationStatus             `json:"status,omitempty"`
	FailureClass          FailureClass                   `json:"failure_class,omitempty"`
	FailureReason         string                         `json:"failure_reason,omitempty"`
	NextHostAction        NextHostAction                 `json:"next_host_action,omitempty"`
	Observations          []Observation                  `json:"observations,omitempty"`
	EvidenceRefs          []EvidenceRef                  `json:"evidence_refs,omitempty"`
	OutputRefs            []DisplaySafeRef               `json:"output_refs,omitempty"`
	MissingCapabilityRefs []DisplaySafeRef               `json:"missing_capability_refs,omitempty"`
	MissingInputs         []MissingInput                 `json:"missing_inputs,omitempty"`
	Boundaries            []Boundary                     `json:"boundaries,omitempty"`
	RawOutputLoaded       bool                           `json:"raw_output_loaded"`
}

type RuntimeAdapterExecutionResult struct {
	ContractVersion                  string                         `json:"contract_version,omitempty"`
	Projected                        bool                           `json:"projected"`
	Status                           VerificationStatus             `json:"status,omitempty"`
	Satisfied                        bool                           `json:"satisfied"`
	HostExecutionReported            bool                           `json:"host_execution_reported"`
	ReadyForObservationNormalization bool                           `json:"ready_for_observation_normalization"`
	Request                          RuntimeAdapterExecutionRequest `json:"request,omitempty"`
	AdapterRef                       DisplaySafeRef                 `json:"adapter_ref,omitempty"`
	StrategyRef                      DisplaySafeRef                 `json:"strategy_ref,omitempty"`
	HostAdapterRunRef                DisplaySafeRef                 `json:"host_adapter_run_ref,omitempty"`
	Observations                     []Observation                  `json:"observations,omitempty"`
	EvidenceRefs                     []EvidenceRef                  `json:"evidence_refs,omitempty"`
	OutputRefs                       []DisplaySafeRef               `json:"output_refs,omitempty"`
	MissingCapabilityRefs            []DisplaySafeRef               `json:"missing_capability_refs,omitempty"`
	FailureClass                     FailureClass                   `json:"failure_class,omitempty"`
	FailureReason                    string                         `json:"failure_reason,omitempty"`
	MissingInputs                    []MissingInput                 `json:"missing_inputs,omitempty"`
	Boundaries                       []Boundary                     `json:"boundaries,omitempty"`
	NextHostAction                   NextHostAction                 `json:"next_host_action,omitempty"`
	RunnerEffect                     string                         `json:"runner_effect,omitempty"`
	PromptEffect                     string                         `json:"prompt_effect,omitempty"`
	RawOutputLoaded                  bool                           `json:"raw_output_loaded"`
}

func BuildRuntimeAdapterExecutionResult(input RuntimeAdapterExecutionResultInput) RuntimeAdapterExecutionResult {
	request := input.Request.Normalize()
	status := NormalizeVerificationStatus(string(input.Status))
	if status == VerificationNotEvaluated {
		status = VerificationSatisfied
	}
	observations := normalizeObservations(input.Observations)
	result := RuntimeAdapterExecutionResult{
		ContractVersion:       ContractVersion,
		Projected:             true,
		Status:                status,
		Request:               request,
		AdapterRef:            firstDisplaySafeRef(input.AdapterRef, request.AdapterRef),
		StrategyRef:           firstDisplaySafeRef(input.StrategyRef, request.StrategyRef),
		HostAdapterRunRef:     normalizeOneDisplaySafeRef(input.HostAdapterRunRef),
		Observations:          observations,
		EvidenceRefs:          MergeEvidenceRefs(input.EvidenceRefs, runtimeAdapterObservationEvidenceRefs(observations)),
		OutputRefs:            normalizeDisplaySafeRefs(input.OutputRefs),
		MissingCapabilityRefs: normalizeDisplaySafeRefs(input.MissingCapabilityRefs),
		FailureClass:          NormalizeFailureClass(string(input.FailureClass)),
		FailureReason:         managedObjectiveReplannerSafeReason(input.FailureReason),
		NextHostAction:        NormalizeNextHostAction(string(input.NextHostAction)),
		MissingInputs:         normalizeMissingInputs(input.MissingInputs),
		Boundaries: MergeBoundaries(
			[]Boundary{
				"runtime_adapter_execution_result",
				"host_reported_adapter_result",
				"adapter_result_not_objective_completion",
				"projection_only",
				"no_runner_dispatch",
				"core_does_not_execute_adapter",
			},
			input.Boundaries,
			request.Boundaries,
		),
		RunnerEffect:    "none",
		PromptEffect:    "none",
		RawOutputLoaded: input.RawOutputLoaded || request.RawOutputLoaded || runtimeAdapterObservationRawOutputLoaded(observations),
	}
	if runtimeAdapterExecutionResultUnsafe(input) {
		return runtimeAdapterResultBlock(result, FailureEvidenceWeak, "host:display_safe_refs", "provide_display_safe_refs", "raw_output_not_allowed")
	}
	if !request.ReadyForHostExecution {
		return runtimeAdapterResultBlock(result, firstFailureClass(request.FailureClass, FailureConfigMissing), "host:runtime_adapter_execution_request", "provide_runtime_adapter_execution_request", "runtime_adapter_request_not_ready")
	}
	if result.AdapterRef != request.AdapterRef {
		return runtimeAdapterResultBlock(result, FailureVerificationFailed, "host:runtime_adapter_result", "review_runtime_adapter_result", "runtime_adapter_result_adapter_mismatch")
	}
	if result.StrategyRef != request.StrategyRef {
		return runtimeAdapterResultBlock(result, FailureVerificationFailed, "host:runtime_adapter_result", "review_runtime_adapter_result", "runtime_adapter_result_strategy_mismatch")
	}
	if result.HostAdapterRunRef == "" {
		result = runtimeAdapterResultBlock(result, FailureEvidenceMissing, "host:adapter_run_ref", "provide_runtime_adapter_result", "runtime_adapter_run_ref_missing")
	}
	if len(result.MissingCapabilityRefs) > 0 {
		result.FailureClass = FailureCapabilityMissing
		for _, ref := range result.MissingCapabilityRefs {
			result.MissingInputs = AppendMissingInputs(result.MissingInputs, MissingInput(ref))
		}
		result.Status = VerificationBlocked
		result.NextHostAction = "enter_capability_resolution"
		result.Boundaries = AppendBoundaries(result.Boundaries, "runtime_adapter_reported_capability_gap", "capability_gap_proposal_only")
		return result.Normalize()
	}
	if status == VerificationSatisfied && len(observations) == 0 {
		result = runtimeAdapterResultBlock(result, FailureEvidenceMissing, "host:adapter_observation", "provide_runtime_adapter_observations", "runtime_adapter_observation_missing")
	}
	if runtimeAdapterObservationsWeak(observations) && result.Status == VerificationSatisfied {
		result.Status = VerificationReviewRequired
		result.FailureClass = FailureEvidenceWeak
		result.MissingInputs = AppendMissingInputs(result.MissingInputs, "host:stronger_adapter_evidence")
		result.Boundaries = AppendBoundaries(result.Boundaries, "runtime_adapter_observation_evidence_weak")
		result.NextHostAction = "review_runtime_adapter_evidence"
		return result.Normalize()
	}
	if result.Status == VerificationFailed || result.Status == VerificationBlocked {
		result.FailureClass = firstFailureClass(result.FailureClass, FailureVerificationFailed)
		result.NextHostAction = firstNextHostAction(result.NextHostAction, "request_replan_or_return_partial")
		result.Boundaries = AppendBoundaries(result.Boundaries, "runtime_adapter_execution_not_satisfied")
		return result.Normalize()
	}
	if len(result.MissingInputs) == 0 && result.Status == VerificationSatisfied {
		result.Satisfied = true
		result.HostExecutionReported = true
		result.ReadyForObservationNormalization = true
		result.NextHostAction = "normalize_adapter_observations"
		result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_observation_normalization")
	}
	return result.Normalize()
}

func CloneRuntimeAdapterExecutionResult(in RuntimeAdapterExecutionResult) RuntimeAdapterExecutionResult {
	out := in
	out.Request = in.Request.Clone()
	out.Observations = cloneObservations(in.Observations)
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.OutputRefs = cloneDisplaySafeRefs(in.OutputRefs)
	out.MissingCapabilityRefs = cloneDisplaySafeRefs(in.MissingCapabilityRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (r RuntimeAdapterExecutionResult) Clone() RuntimeAdapterExecutionResult {
	return CloneRuntimeAdapterExecutionResult(r)
}

func (r RuntimeAdapterExecutionResult) Normalize() RuntimeAdapterExecutionResult {
	out := CloneRuntimeAdapterExecutionResult(r)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeVerificationStatus(string(out.Status))
	out.Request = out.Request.Normalize()
	out.AdapterRef = normalizeOneDisplaySafeRef(out.AdapterRef)
	out.StrategyRef = normalizeOneDisplaySafeRef(out.StrategyRef)
	out.HostAdapterRunRef = normalizeOneDisplaySafeRef(out.HostAdapterRunRef)
	out.Observations = normalizeObservations(out.Observations)
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
	out.OutputRefs = normalizeDisplaySafeRefs(out.OutputRefs)
	out.MissingCapabilityRefs = normalizeDisplaySafeRefs(out.MissingCapabilityRefs)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.FailureReason = managedObjectiveReplannerSafeReason(out.FailureReason)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	if out.Status == VerificationNotEvaluated {
		out.Status = VerificationBlocked
	}
	if out.RunnerEffect == "" {
		out.RunnerEffect = "none"
	}
	if out.PromptEffect == "" {
		out.PromptEffect = "none"
	}
	if out.RawOutputLoaded {
		out.Status = VerificationReviewRequired
		if out.FailureClass == FailureNone {
			out.FailureClass = FailureEvidenceWeak
		}
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		if out.NextHostAction == "" {
			out.NextHostAction = "provide_display_safe_refs"
		}
	}
	out.Satisfied = out.Status == VerificationSatisfied && len(out.MissingInputs) == 0 && len(out.MissingCapabilityRefs) == 0 && !out.RawOutputLoaded
	out.HostExecutionReported = out.HostAdapterRunRef != "" && (out.Satisfied || out.Status == VerificationFailed || out.Status == VerificationBlocked)
	out.ReadyForObservationNormalization = out.Satisfied && len(out.Observations) > 0
	return out
}

func hostAdapterRegistryEntryFromDescriptor(descriptor ProductionAdapterDescriptor) HostAdapterRegistryEntry {
	return HostAdapterRegistryEntry{
		AdapterRef:             descriptor.AdapterRef,
		Descriptor:             descriptor,
		SupportedStrategyRefs:  descriptor.SupportedCandidateRefs,
		ProvidesCapabilityRefs: descriptor.ProvidesCapabilityRefs,
		RequiresCapabilityRefs: descriptor.RequiresCapabilityRefs,
		RequiredPolicyRefs:     descriptor.RequiredPolicyRefs,
		RequiredApprovalRefs:   descriptor.RequiredApprovalRefs,
		RequiredBudgetRef:      descriptor.RequiredBudgetRef,
		IdempotencyContractRef: descriptor.IdempotencyContractRef,
		SideEffectClass:        descriptor.SideEffectClass,
	}.Normalize()
}

func selectRuntimeAdapterRegistryEntry(registry HostAdapterRegistrySnapshot, requested DisplaySafeRef, strategyRef DisplaySafeRef) (HostAdapterRegistryEntry, bool) {
	normalizedRequested := normalizeOneDisplaySafeRef(requested)
	normalizedStrategy := normalizeOneDisplaySafeRef(strategyRef)
	for _, entry := range registry.Entries {
		normalized := entry.Normalize()
		if normalizedRequested != "" && normalized.AdapterRef != normalizedRequested {
			continue
		}
		if normalizedStrategy != "" && !displaySafeRefSliceContains(normalized.SupportedStrategyRefs, normalizedStrategy) {
			continue
		}
		return normalized, true
	}
	if normalizedRequested != "" {
		for _, entry := range registry.Entries {
			normalized := entry.Normalize()
			if normalized.AdapterRef == normalizedRequested {
				return normalized, true
			}
		}
	}
	return HostAdapterRegistryEntry{}, false
}

func hostAdapterRegistryBlock(result HostAdapterRegistrySnapshot, failure FailureClass, missing MissingInput, next NextHostAction, boundary Boundary) HostAdapterRegistrySnapshot {
	result.Status = HostActionBlocked
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.Boundaries = AppendBoundaries(result.Boundaries, boundary)
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	return result
}

func runtimeAdapterRequestBlock(result RuntimeAdapterExecutionRequest, failure FailureClass, missing MissingInput, next NextHostAction, boundary Boundary) RuntimeAdapterExecutionRequest {
	result.Status = HostActionBlocked
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.Boundaries = AppendBoundaries(result.Boundaries, boundary)
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	return result.Normalize()
}

func runtimeAdapterResultBlock(result RuntimeAdapterExecutionResult, failure FailureClass, missing MissingInput, next NextHostAction, boundary Boundary) RuntimeAdapterExecutionResult {
	result.Status = VerificationBlocked
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.Boundaries = AppendBoundaries(result.Boundaries, boundary)
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	return result.Normalize()
}

func runtimeAdapterReadOnlyKind(kind ProductionAdapterKind) bool {
	switch NormalizeProductionAdapterKind(string(kind)) {
	case ProductionAdapterOperationsMetricCollect, ProductionAdapterSourceReadback:
		return true
	default:
		return false
	}
}

func runtimeAdapterReadOnlySideEffectClass(value string) bool {
	switch normalizeControlToken(value) {
	case "read_only", "tool_read_only", "metric_read_only", "metrics_read_only", "no_side_effect", "none":
		return true
	default:
		return false
	}
}

func firstNonEmptyControlToken(values ...string) string {
	for _, value := range values {
		token := normalizeControlToken(value)
		if token != "" {
			return token
		}
	}
	return ""
}

func normalizeControlModes(in []ControlMode) []ControlMode {
	out := make([]ControlMode, 0, len(in))
	seen := map[ControlMode]struct{}{}
	for _, value := range in {
		mode := NormalizeControlMode(string(value))
		if mode == "" {
			continue
		}
		if _, exists := seen[mode]; exists {
			continue
		}
		seen[mode] = struct{}{}
		out = append(out, mode)
	}
	return out
}

func cloneControlModes(in []ControlMode) []ControlMode {
	if len(in) == 0 {
		return nil
	}
	return append([]ControlMode(nil), in...)
}

func cloneHostAdapterRegistryEntries(in []HostAdapterRegistryEntry) []HostAdapterRegistryEntry {
	if len(in) == 0 {
		return nil
	}
	out := make([]HostAdapterRegistryEntry, len(in))
	for i, value := range in {
		out[i] = value.Clone()
	}
	return out
}

func normalizeHostAdapterRegistryEntries(in []HostAdapterRegistryEntry) []HostAdapterRegistryEntry {
	out := make([]HostAdapterRegistryEntry, 0, len(in))
	seen := map[DisplaySafeRef]struct{}{}
	for _, value := range in {
		normalized := value.Normalize()
		if normalized.AdapterRef == "" {
			continue
		}
		if _, exists := seen[normalized.AdapterRef]; exists {
			continue
		}
		seen[normalized.AdapterRef] = struct{}{}
		out = append(out, normalized)
	}
	return out
}

func hostAdapterRegistryInputUnsafe(input HostAdapterRegistryInput) bool {
	return displaySafeRefRejected(input.RegistryRef) ||
		productionAdapterRegistrySnapshotUnsafe(input.ProductionRegistry) ||
		productionAdapterDescriptorSliceUnsafe(input.Descriptors) ||
		hostAdapterRegistryEntrySliceUnsafe(input.Entries) ||
		displaySafeRefSliceRejected(input.PolicyRefs) ||
		evidenceRefRejected(input.EvidenceRefs) ||
		input.RawOutputLoaded
}

func runtimeAdapterExecutionRequestUnsafe(input RuntimeAdapterExecutionRequestInput) bool {
	return displaySafeRefRejected(input.RequestedAdapterRef) ||
		displaySafeRefSliceRejected(input.ApprovalRefs) ||
		displaySafeRefSliceRejected(input.PolicyRefs) ||
		displaySafeRefSliceRejected(input.AvailableCapabilityRefs) ||
		displaySafeRefRejected(input.IdempotencyRef) ||
		displaySafeRefSliceRejected(input.InputRefs) ||
		input.RawOutputLoaded
}

func runtimeAdapterExecutionResultUnsafe(input RuntimeAdapterExecutionResultInput) bool {
	return displaySafeRefRejected(input.AdapterRef) ||
		displaySafeRefRejected(input.StrategyRef) ||
		displaySafeRefRejected(input.HostAdapterRunRef) ||
		displaySafeRefSliceRejected(input.OutputRefs) ||
		displaySafeRefSliceRejected(input.MissingCapabilityRefs) ||
		evidenceRefRejected(input.EvidenceRefs) ||
		runtimeAdapterObservationRejected(input.Observations) ||
		input.RawOutputLoaded
}

func productionAdapterDescriptorSliceUnsafe(values []ProductionAdapterDescriptor) bool {
	for _, value := range values {
		if productionAdapterDescriptorUnsafe(value) {
			return true
		}
	}
	return false
}

func hostAdapterRegistryEntrySliceUnsafe(values []HostAdapterRegistryEntry) bool {
	for _, value := range values {
		normalized := value.Normalize()
		if displaySafeRefRejected(normalized.AdapterRef) ||
			productionAdapterDescriptorUnsafe(normalized.Descriptor) ||
			displaySafeRefSliceRejected(normalized.SupportedStrategyRefs) ||
			displaySafeRefSliceRejected(normalized.ProvidesCapabilityRefs) ||
			displaySafeRefSliceRejected(normalized.RequiresCapabilityRefs) ||
			displaySafeRefSliceRejected(normalized.RequiredPolicyRefs) ||
			displaySafeRefSliceRejected(normalized.RequiredApprovalRefs) ||
			displaySafeRefRejected(normalized.RequiredBudgetRef) ||
			displaySafeRefRejected(normalized.IdempotencyContractRef) ||
			normalized.RawOutputLoaded {
			return true
		}
	}
	return false
}

func runtimeAdapterObservationRejected(values []Observation) bool {
	for _, value := range values {
		if displaySafeRefRejected(value.Source) ||
			displaySafeRefRejected(value.Subject) ||
			displaySafeRefSliceRejected(value.DisplaySafeRefs) ||
			evidenceRefRejected(value.EvidenceRefs) ||
			value.RawOutputLoaded {
			return true
		}
	}
	return false
}

func runtimeAdapterObservationRawOutputLoaded(values []Observation) bool {
	for _, observation := range values {
		if observation.RawOutputLoaded {
			return true
		}
	}
	return false
}

func runtimeAdapterObservationsWeak(values []Observation) bool {
	for _, observation := range normalizeObservations(values) {
		if observation.Strength == EvidenceWeak || observation.Strength == EvidenceMissing {
			return true
		}
		if len(observation.EvidenceRefs) == 0 && len(observation.DisplaySafeRefs) == 0 {
			return true
		}
	}
	return false
}
