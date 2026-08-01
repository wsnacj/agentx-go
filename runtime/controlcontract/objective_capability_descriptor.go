package controlcontract

type ObjectiveCapabilitySideEffectClass string

const (
	ObjectiveCapabilitySideEffectUnspecified   ObjectiveCapabilitySideEffectClass = "unspecified"
	ObjectiveCapabilitySideEffectReadOnly      ObjectiveCapabilitySideEffectClass = "read_only"
	ObjectiveCapabilitySideEffectLocalWrite    ObjectiveCapabilitySideEffectClass = "local_write"
	ObjectiveCapabilitySideEffectExternalWrite ObjectiveCapabilitySideEffectClass = "external_write"
	ObjectiveCapabilitySideEffectInstall       ObjectiveCapabilitySideEffectClass = "install"
	ObjectiveCapabilitySideEffectSchedule      ObjectiveCapabilitySideEffectClass = "schedule"
	ObjectiveCapabilitySideEffectPayment       ObjectiveCapabilitySideEffectClass = "payment"
)

func NormalizeObjectiveCapabilitySideEffectClass(raw string) ObjectiveCapabilitySideEffectClass {
	switch normalizeEnumToken(raw) {
	case "read_only", "readonly", "query_only", "inspect", "fetch":
		return ObjectiveCapabilitySideEffectReadOnly
	case "local_write", "write_local", "filesystem_write", "store_write":
		return ObjectiveCapabilitySideEffectLocalWrite
	case "external_write", "write_external", "external_action", "remote_write":
		return ObjectiveCapabilitySideEffectExternalWrite
	case "install", "tool_install", "skill_install", "capability_install":
		return ObjectiveCapabilitySideEffectInstall
	case "schedule", "scheduled", "scheduler", "automation":
		return ObjectiveCapabilitySideEffectSchedule
	case "payment", "purchase", "booking", "order":
		return ObjectiveCapabilitySideEffectPayment
	default:
		return ObjectiveCapabilitySideEffectUnspecified
	}
}

type ObjectiveCapabilityDescriptor struct {
	ContractVersion           string                             `json:"contract_version,omitempty"`
	DescriptorRef             DisplaySafeRef                     `json:"descriptor_ref,omitempty"`
	CapabilityRef             DisplaySafeRef                     `json:"capability_ref,omitempty"`
	StrategyRef               DisplaySafeRef                     `json:"strategy_ref,omitempty"`
	SourceKind                StrategyCatalogSourceKind          `json:"source_kind,omitempty"`
	SourceRef                 DisplaySafeRef                     `json:"source_ref,omitempty"`
	OwnerRef                  DisplaySafeRef                     `json:"owner_ref,omitempty"`
	ProviderRef               DisplaySafeRef                     `json:"provider_ref,omitempty"`
	StrategyKind              string                             `json:"strategy_kind,omitempty"`
	ControlMode               ControlMode                        `json:"control_mode,omitempty"`
	MinIntensity              ExecutionIntensity                 `json:"min_intensity,omitempty"`
	MaxIntensity              ExecutionIntensity                 `json:"max_intensity,omitempty"`
	InputSchemaRef            DisplaySafeRef                     `json:"input_schema_ref,omitempty"`
	OutputSchemaRef           DisplaySafeRef                     `json:"output_schema_ref,omitempty"`
	EvidenceContractRef       DisplaySafeRef                     `json:"evidence_contract_ref,omitempty"`
	RequiredEvidence          []EvidenceRef                      `json:"required_evidence,omitempty"`
	SideEffectClass           ObjectiveCapabilitySideEffectClass `json:"side_effect_class,omitempty"`
	RequiresApproval          bool                               `json:"requires_approval"`
	CredentialRequirementRefs []DisplaySafeRef                   `json:"credential_requirement_refs,omitempty"`
	ConfigRequirementRefs     []DisplaySafeRef                   `json:"config_requirement_refs,omitempty"`
	FailureClasses            []FailureClass                     `json:"failure_classes,omitempty"`
	ExampleRefs               []DisplaySafeRef                   `json:"example_refs,omitempty"`
	VerificationHintRefs      []DisplaySafeRef                   `json:"verification_hint_refs,omitempty"`
	DomainHintRefs            []DisplaySafeRef                   `json:"domain_hint_refs,omitempty"`
	PolicyRefs                []DisplaySafeRef                   `json:"policy_refs,omitempty"`
	Boundaries                []Boundary                         `json:"boundaries,omitempty"`
	MissingInputs             []MissingInput                     `json:"missing_inputs,omitempty"`
	RawOutputLoaded           bool                               `json:"raw_output_loaded"`
}

func CloneObjectiveCapabilityDescriptor(in ObjectiveCapabilityDescriptor) ObjectiveCapabilityDescriptor {
	out := in
	out.RequiredEvidence = cloneEvidenceRefs(in.RequiredEvidence)
	out.CredentialRequirementRefs = cloneDisplaySafeRefs(in.CredentialRequirementRefs)
	out.ConfigRequirementRefs = cloneDisplaySafeRefs(in.ConfigRequirementRefs)
	out.FailureClasses = cloneFailureClasses(in.FailureClasses)
	out.ExampleRefs = cloneDisplaySafeRefs(in.ExampleRefs)
	out.VerificationHintRefs = cloneDisplaySafeRefs(in.VerificationHintRefs)
	out.DomainHintRefs = cloneDisplaySafeRefs(in.DomainHintRefs)
	out.PolicyRefs = cloneDisplaySafeRefs(in.PolicyRefs)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	return out
}

func (d ObjectiveCapabilityDescriptor) Clone() ObjectiveCapabilityDescriptor {
	return CloneObjectiveCapabilityDescriptor(d)
}

func (d ObjectiveCapabilityDescriptor) Normalize() ObjectiveCapabilityDescriptor {
	out := CloneObjectiveCapabilityDescriptor(d)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.DescriptorRef = normalizeOneDisplaySafeRef(out.DescriptorRef)
	out.CapabilityRef = normalizeOneDisplaySafeRef(out.CapabilityRef)
	out.StrategyRef = normalizeOneDisplaySafeRef(out.StrategyRef)
	out.SourceKind = NormalizeStrategyCatalogSourceKind(string(out.SourceKind))
	out.SourceRef = normalizeOneDisplaySafeRef(out.SourceRef)
	out.OwnerRef = normalizeOneDisplaySafeRef(out.OwnerRef)
	out.ProviderRef = normalizeOneDisplaySafeRef(out.ProviderRef)
	out.StrategyKind = normalizeControlToken(out.StrategyKind)
	out.ControlMode = NormalizeControlMode(string(out.ControlMode))
	if out.ControlMode == "" {
		out.ControlMode = ControlModeObjective
	}
	out.MinIntensity = NormalizeExecutionIntensity(string(out.MinIntensity))
	if out.MinIntensity == "" {
		out.MinIntensity = IntensityL3ManagedObjective
	}
	out.MaxIntensity = NormalizeExecutionIntensity(string(out.MaxIntensity))
	if out.MaxIntensity == "" {
		out.MaxIntensity = out.MinIntensity
	}
	out.InputSchemaRef = normalizeOneDisplaySafeRef(out.InputSchemaRef)
	out.OutputSchemaRef = normalizeOneDisplaySafeRef(out.OutputSchemaRef)
	out.EvidenceContractRef = normalizeOneDisplaySafeRef(out.EvidenceContractRef)
	out.RequiredEvidence = normalizeEvidenceRefs(out.RequiredEvidence)
	out.SideEffectClass = NormalizeObjectiveCapabilitySideEffectClass(string(out.SideEffectClass))
	out.CredentialRequirementRefs = normalizeDisplaySafeRefs(out.CredentialRequirementRefs)
	out.ConfigRequirementRefs = normalizeDisplaySafeRefs(out.ConfigRequirementRefs)
	out.FailureClasses = normalizeFailureClasses(out.FailureClasses)
	out.ExampleRefs = normalizeDisplaySafeRefs(out.ExampleRefs)
	out.VerificationHintRefs = normalizeDisplaySafeRefs(out.VerificationHintRefs)
	out.DomainHintRefs = normalizeDisplaySafeRefs(out.DomainHintRefs)
	out.PolicyRefs = normalizeDisplaySafeRefs(out.PolicyRefs)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	return out
}

type ObjectiveCapabilityDescriptorProjectionInput struct {
	Descriptor         ObjectiveCapabilityDescriptor `json:"descriptor,omitempty"`
	ContributionRef    DisplaySafeRef                `json:"contribution_ref,omitempty"`
	StrategyVersionRef DisplaySafeRef                `json:"strategy_version_ref,omitempty"`
	StrategyDigestRef  DisplaySafeRef                `json:"strategy_digest_ref,omitempty"`
	ProvenanceRefs     []DisplaySafeRef              `json:"provenance_refs,omitempty"`
	Boundaries         []Boundary                    `json:"boundaries,omitempty"`
	RawOutputLoaded    bool                          `json:"raw_output_loaded"`
}

type ObjectiveCapabilityDescriptorProjection struct {
	ContractVersion      string                              `json:"contract_version,omitempty"`
	Projected            bool                                `json:"projected"`
	Available            bool                                `json:"available"`
	Status               VerificationStatus                  `json:"status,omitempty"`
	Mode                 string                              `json:"mode,omitempty"`
	RunnerEffect         string                              `json:"runner_effect,omitempty"`
	PromptEffect         string                              `json:"prompt_effect,omitempty"`
	Descriptor           ObjectiveCapabilityDescriptor       `json:"descriptor,omitempty"`
	StrategyCandidate    StrategyCandidate                   `json:"strategy_candidate,omitempty"`
	StrategyContribution AdapterMetadataStrategyContribution `json:"strategy_contribution,omitempty"`
	ReadyForCatalog      bool                                `json:"ready_for_catalog"`
	FailureClass         FailureClass                        `json:"failure_class,omitempty"`
	MissingInputs        []MissingInput                      `json:"missing_inputs,omitempty"`
	BlockedReasons       []string                            `json:"blocked_reasons,omitempty"`
	Boundaries           []Boundary                          `json:"boundaries,omitempty"`
	NextHostAction       NextHostAction                      `json:"next_host_action,omitempty"`
	RawOutputLoaded      bool                                `json:"raw_output_loaded"`
}

func BuildObjectiveCapabilityDescriptorProjection(input ObjectiveCapabilityDescriptorProjectionInput) ObjectiveCapabilityDescriptorProjection {
	result := baseObjectiveCapabilityDescriptorProjection(input)
	if objectiveCapabilityDescriptorProjectionInputUnsafe(input) {
		return objectiveCapabilityDescriptorProjectionBlock(result, VerificationReviewRequired, FailureEvidenceWeak, "objective_capability_descriptor_unsafe_input", "host:display_safe_refs", "provide_display_safe_refs", "raw_output_not_allowed")
	}
	missing := objectiveCapabilityDescriptorMissingInputs(result.Descriptor)
	result.MissingInputs = MergeMissingInputs(result.MissingInputs, missing)
	if len(result.MissingInputs) > 0 {
		result.FailureClass = objectiveCapabilityDescriptorFailure(result.MissingInputs)
		result.BlockedReasons = normalizeControlTokenList(append(result.BlockedReasons, "objective_capability_descriptor_incomplete"))
		result.Boundaries = AppendBoundaries(result.Boundaries, "objective_capability_descriptor_incomplete")
		result.NextHostAction = objectiveCapabilityDescriptorNextAction(result.MissingInputs)
		return result.Normalize()
	}
	if result.Descriptor.SideEffectClass != ObjectiveCapabilitySideEffectReadOnly && !result.Descriptor.RequiresApproval {
		return objectiveCapabilityDescriptorProjectionBlock(result, VerificationBlocked, FailureApprovalRequired, "objective_capability_descriptor_requires_approval_missing", "host:objective_capability_approval_policy", "provide_objective_capability_approval_policy", "objective_capability_side_effect_requires_approval")
	}

	result.StrategyCandidate = objectiveCapabilityDescriptorStrategyCandidate(result.Descriptor)
	result.StrategyContribution = objectiveCapabilityDescriptorStrategyContribution(result.Descriptor, input, result.StrategyCandidate)
	result.Status = VerificationSatisfied
	result.ReadyForCatalog = true
	result.FailureClass = FailureNone
	result.NextHostAction = "include_in_strategy_catalog"
	result.Boundaries = AppendBoundaries(result.Boundaries, "objective_capability_descriptor_ready")
	return result.Normalize()
}

func (p ObjectiveCapabilityDescriptorProjection) Clone() ObjectiveCapabilityDescriptorProjection {
	out := p
	out.Descriptor = p.Descriptor.Clone()
	out.StrategyCandidate = p.StrategyCandidate.Clone()
	out.StrategyContribution = p.StrategyContribution.Clone()
	out.MissingInputs = cloneMissingInputs(p.MissingInputs)
	out.BlockedReasons = cloneStringSlice(p.BlockedReasons)
	out.Boundaries = cloneBoundaries(p.Boundaries)
	return out
}

func (p ObjectiveCapabilityDescriptorProjection) Normalize() ObjectiveCapabilityDescriptorProjection {
	out := p.Clone()
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Available = true
	out.Status = NormalizeVerificationStatus(string(out.Status))
	if out.Status == VerificationNotEvaluated {
		out.Status = VerificationBlocked
	}
	out.Mode = normalizeControlToken(out.Mode)
	if out.Mode == "" {
		out.Mode = "objective_capability_descriptor_projection"
	}
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	if out.RunnerEffect == "" {
		out.RunnerEffect = "none"
	}
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	if out.PromptEffect == "" {
		out.PromptEffect = "none"
	}
	out.Descriptor = out.Descriptor.Normalize()
	out.StrategyCandidate = out.StrategyCandidate.Normalize()
	out.StrategyContribution = out.StrategyContribution.Normalize()
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.BlockedReasons = normalizeControlTokenList(out.BlockedReasons)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	if out.RawOutputLoaded || out.Descriptor.RawOutputLoaded || out.StrategyContribution.RawOutputLoaded {
		out.Status = VerificationReviewRequired
		out.ReadyForCatalog = false
		out.FailureClass = firstFailureClass(out.FailureClass, FailureEvidenceWeak)
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "objective_capability_descriptor_unsafe_input")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		out.NextHostAction = "provide_display_safe_refs"
	}
	if out.Status != VerificationSatisfied {
		out.ReadyForCatalog = false
	}
	return out
}

func baseObjectiveCapabilityDescriptorProjection(input ObjectiveCapabilityDescriptorProjectionInput) ObjectiveCapabilityDescriptorProjection {
	descriptor := input.Descriptor.Normalize()
	return ObjectiveCapabilityDescriptorProjection{
		ContractVersion: ContractVersion,
		Projected:       true,
		Available:       true,
		Status:          VerificationBlocked,
		Mode:            "objective_capability_descriptor_projection",
		RunnerEffect:    "none",
		PromptEffect:    "none",
		Descriptor:      descriptor,
		FailureClass:    FailureInsufficientInformation,
		MissingInputs: MergeMissingInputs(
			descriptor.MissingInputs,
			objectiveCapabilityDescriptorProjectionMissingInputs(input),
		),
		Boundaries: MergeBoundaries(
			[]Boundary{
				"objective_capability_descriptor",
				"objective_capability_descriptor_validator",
				"strategy_catalog_contribution_projection",
				"no_parallel_catalog",
				"no_prompt_parsing",
				"no_llm_call",
				"no_runner_dispatch",
				"no_backend_execution",
			},
			descriptor.Boundaries,
			input.Boundaries,
		),
		NextHostAction:  "provide_objective_capability_descriptor",
		RawOutputLoaded: input.RawOutputLoaded || descriptor.RawOutputLoaded,
	}
}

func objectiveCapabilityDescriptorProjectionBlock(result ObjectiveCapabilityDescriptorProjection, status VerificationStatus, failure FailureClass, reason string, missing MissingInput, next NextHostAction, boundary Boundary) ObjectiveCapabilityDescriptorProjection {
	result.Status = status
	result.FailureClass = failure
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = next
	result.Boundaries = AppendBoundaries(result.Boundaries, boundary)
	return result.Normalize()
}

func objectiveCapabilityDescriptorStrategyCandidate(descriptor ObjectiveCapabilityDescriptor) StrategyCandidate {
	normalized := descriptor.Normalize()
	return StrategyCandidate{
		ID:               string(normalized.StrategyRef),
		Kind:             normalized.StrategyKind,
		ControlMode:      normalized.ControlMode,
		MinIntensity:     normalized.MinIntensity,
		MaxIntensity:     normalized.MaxIntensity,
		CapabilityRefs:   []DisplaySafeRef{normalized.CapabilityRef},
		ExpectedEvidence: cloneEvidenceRefs(normalized.RequiredEvidence),
		Preconditions:    objectiveCapabilityDescriptorPreconditions(normalized),
		Boundaries:       objectiveCapabilityDescriptorStrategyBoundaries(normalized),
		Risk:             objectiveCapabilityDescriptorRisk(normalized.SideEffectClass),
		SideEffectClass:  string(normalized.SideEffectClass),
		RequiresApproval: normalized.RequiresApproval,
		Owner:            string(normalized.OwnerRef),
	}.Normalize()
}

func objectiveCapabilityDescriptorStrategyContribution(descriptor ObjectiveCapabilityDescriptor, input ObjectiveCapabilityDescriptorProjectionInput, candidate StrategyCandidate) AdapterMetadataStrategyContribution {
	normalized := descriptor.Normalize()
	return AdapterMetadataStrategyContribution{
		ContributionRef:    normalizeOneDisplaySafeRef(input.ContributionRef),
		OwnerRef:           normalized.OwnerRef,
		ProviderRef:        normalized.ProviderRef,
		StrategyVersionRef: normalizeOneDisplaySafeRef(input.StrategyVersionRef),
		StrategyDigestRef:  normalizeOneDisplaySafeRef(input.StrategyDigestRef),
		SourceKind:         normalized.SourceKind,
		SourceRef:          normalized.SourceRef,
		Candidate:          candidate,
		ProvenanceRefs:     normalizeDisplaySafeRefs(append(cloneDisplaySafeRefs(input.ProvenanceRefs), normalized.DescriptorRef)),
		EvidenceRefs:       cloneEvidenceRefs(normalized.RequiredEvidence),
		PolicyRefs:         cloneDisplaySafeRefs(normalized.PolicyRefs),
		Boundaries: MergeBoundaries(
			normalized.Boundaries,
			input.Boundaries,
			[]Boundary{"objective_capability_descriptor_projected_to_strategy_contribution"},
		),
	}.Normalize()
}

func objectiveCapabilityDescriptorMissingInputs(descriptor ObjectiveCapabilityDescriptor) []MissingInput {
	var missing []MissingInput
	checks := []struct {
		missing bool
		input   MissingInput
	}{
		{descriptor.DescriptorRef == "", "host:objective_capability_descriptor_ref"},
		{descriptor.CapabilityRef == "", "host:objective_capability_ref"},
		{descriptor.StrategyRef == "", "host:objective_capability_strategy_ref"},
		{descriptor.SourceKind == "", "host:objective_capability_source_kind"},
		{descriptor.SourceRef == "", "host:objective_capability_source_ref"},
		{descriptor.OwnerRef == "", "host:objective_capability_owner_ref"},
		{descriptor.ProviderRef == "", "host:objective_capability_provider_ref"},
		{descriptor.StrategyKind == "", "host:objective_capability_strategy_kind"},
		{descriptor.InputSchemaRef == "", "host:objective_capability_input_schema_ref"},
		{descriptor.OutputSchemaRef == "", "host:objective_capability_output_schema_ref"},
		{descriptor.EvidenceContractRef == "", "host:objective_capability_evidence_contract_ref"},
		{len(descriptor.RequiredEvidence) == 0, "host:objective_capability_required_evidence"},
		{descriptor.SideEffectClass == ObjectiveCapabilitySideEffectUnspecified, "host:objective_capability_side_effect_class"},
		{len(descriptor.CredentialRequirementRefs) == 0, "host:objective_capability_credential_requirement_ref"},
		{len(descriptor.ConfigRequirementRefs) == 0, "host:objective_capability_config_requirement_ref"},
		{len(descriptor.FailureClasses) == 0, "host:objective_capability_failure_classes"},
		{len(descriptor.ExampleRefs) == 0, "host:objective_capability_example_refs"},
		{len(descriptor.VerificationHintRefs) == 0, "host:objective_capability_verification_hint_refs"},
	}
	for _, check := range checks {
		if check.missing {
			missing = AppendMissingInputs(missing, check.input)
		}
	}
	return missing
}

func objectiveCapabilityDescriptorProjectionMissingInputs(input ObjectiveCapabilityDescriptorProjectionInput) []MissingInput {
	var missing []MissingInput
	if normalizeOneDisplaySafeRef(input.ContributionRef) == "" {
		missing = AppendMissingInputs(missing, "host:objective_capability_contribution_ref")
	}
	if normalizeOneDisplaySafeRef(input.StrategyVersionRef) == "" {
		missing = AppendMissingInputs(missing, "host:objective_capability_strategy_version_ref")
	}
	if normalizeOneDisplaySafeRef(input.StrategyDigestRef) == "" {
		missing = AppendMissingInputs(missing, "host:objective_capability_strategy_digest_ref")
	}
	return missing
}

func objectiveCapabilityDescriptorFailure(missing []MissingInput) FailureClass {
	for _, value := range missing {
		switch value {
		case "host:objective_capability_side_effect_class":
			return FailurePolicyBlocked
		case "host:objective_capability_credential_requirement_ref", "host:objective_capability_config_requirement_ref":
			return FailureConfigMissing
		case "host:objective_capability_required_evidence", "host:objective_capability_input_schema_ref", "host:objective_capability_output_schema_ref", "host:objective_capability_evidence_contract_ref":
			return FailureEvidenceMissing
		}
	}
	return FailureInsufficientInformation
}

func objectiveCapabilityDescriptorNextAction(missing []MissingInput) NextHostAction {
	for _, value := range missing {
		switch value {
		case "host:objective_capability_credential_requirement_ref", "host:objective_capability_config_requirement_ref":
			return "provide_objective_capability_readiness_refs"
		case "host:objective_capability_required_evidence", "host:objective_capability_input_schema_ref", "host:objective_capability_output_schema_ref", "host:objective_capability_evidence_contract_ref":
			return "provide_objective_capability_contract"
		}
	}
	return "provide_objective_capability_descriptor"
}

func objectiveCapabilityDescriptorPreconditions(descriptor ObjectiveCapabilityDescriptor) []MissingInput {
	var out []MissingInput
	for _, ref := range descriptor.CredentialRequirementRefs {
		out = AppendMissingInputs(out, MissingInput("host:"+string(ref)))
	}
	for _, ref := range descriptor.ConfigRequirementRefs {
		out = AppendMissingInputs(out, MissingInput("host:"+string(ref)))
	}
	return out
}

func objectiveCapabilityDescriptorStrategyBoundaries(descriptor ObjectiveCapabilityDescriptor) []Boundary {
	return MergeBoundaries(
		[]Boundary{
			"objective_capability_descriptor_strategy",
			Boundary("objective_capability_side_effect_" + string(descriptor.SideEffectClass)),
			Boundary("input_schema_ref:" + string(descriptor.InputSchemaRef)),
			Boundary("output_schema_ref:" + string(descriptor.OutputSchemaRef)),
			Boundary("evidence_contract_ref:" + string(descriptor.EvidenceContractRef)),
		},
		descriptor.Boundaries,
	)
}

func objectiveCapabilityDescriptorRisk(sideEffect ObjectiveCapabilitySideEffectClass) string {
	switch sideEffect {
	case ObjectiveCapabilitySideEffectReadOnly:
		return "read_only"
	case ObjectiveCapabilitySideEffectLocalWrite:
		return "local_write"
	case ObjectiveCapabilitySideEffectExternalWrite, ObjectiveCapabilitySideEffectInstall, ObjectiveCapabilitySideEffectSchedule, ObjectiveCapabilitySideEffectPayment:
		return "approval_required"
	default:
		return "unknown"
	}
}

func objectiveCapabilityDescriptorProjectionInputUnsafe(input ObjectiveCapabilityDescriptorProjectionInput) bool {
	return input.RawOutputLoaded ||
		displaySafeRefRejected(input.ContributionRef) ||
		displaySafeRefRejected(input.StrategyVersionRef) ||
		displaySafeRefRejected(input.StrategyDigestRef) ||
		displaySafeRefSliceRejected(input.ProvenanceRefs) ||
		objectiveCapabilityDescriptorUnsafe(input.Descriptor)
}

func objectiveCapabilityDescriptorUnsafe(descriptor ObjectiveCapabilityDescriptor) bool {
	return descriptor.RawOutputLoaded ||
		displaySafeRefRejected(descriptor.DescriptorRef) ||
		displaySafeRefRejected(descriptor.CapabilityRef) ||
		displaySafeRefRejected(descriptor.StrategyRef) ||
		displaySafeRefRejected(descriptor.SourceRef) ||
		displaySafeRefRejected(descriptor.OwnerRef) ||
		displaySafeRefRejected(descriptor.ProviderRef) ||
		displaySafeRefRejected(descriptor.InputSchemaRef) ||
		displaySafeRefRejected(descriptor.OutputSchemaRef) ||
		displaySafeRefRejected(descriptor.EvidenceContractRef) ||
		displaySafeRefSliceRejected(descriptor.CredentialRequirementRefs) ||
		displaySafeRefSliceRejected(descriptor.ConfigRequirementRefs) ||
		displaySafeRefSliceRejected(descriptor.ExampleRefs) ||
		displaySafeRefSliceRejected(descriptor.VerificationHintRefs) ||
		displaySafeRefSliceRejected(descriptor.DomainHintRefs) ||
		displaySafeRefSliceRejected(descriptor.PolicyRefs) ||
		evidenceRefRejected(descriptor.RequiredEvidence) ||
		ContainsUnsafeRawOutput(descriptor.StrategyKind)
}

func normalizeFailureClasses(in []FailureClass) []FailureClass {
	out := make([]FailureClass, 0, len(in))
	seen := map[FailureClass]struct{}{}
	for _, value := range in {
		normalized := NormalizeFailureClass(string(value))
		if normalized == FailureNone {
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

func cloneFailureClasses(in []FailureClass) []FailureClass {
	if len(in) == 0 {
		return nil
	}
	return append([]FailureClass(nil), in...)
}
