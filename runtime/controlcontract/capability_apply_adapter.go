package controlcontract

type CapabilityApplyAction string

const (
	CapabilityApplyInstall    CapabilityApplyAction = "install"
	CapabilityApplyEnable     CapabilityApplyAction = "enable"
	CapabilityApplySkillApply CapabilityApplyAction = "skill_apply"
	CapabilityApplyAuthorize  CapabilityApplyAction = "authorize"
	CapabilityApplyReload     CapabilityApplyAction = "reload"
	CapabilityApplyRollback   CapabilityApplyAction = "rollback"
)

func KnownCapabilityApplyActions() []CapabilityApplyAction {
	return []CapabilityApplyAction{
		CapabilityApplyInstall,
		CapabilityApplyEnable,
		CapabilityApplySkillApply,
		CapabilityApplyAuthorize,
		CapabilityApplyReload,
		CapabilityApplyRollback,
	}
}

func NormalizeCapabilityApplyAction(raw string) CapabilityApplyAction {
	switch normalizeEnumToken(raw) {
	case "install", "tool_install", "capability_install":
		return CapabilityApplyInstall
	case "enable", "activate", "tool_enable", "capability_enable":
		return CapabilityApplyEnable
	case "skill_apply", "apply_skill", "skill_write", "skill_install":
		return CapabilityApplySkillApply
	case "authorize", "authorization", "auth", "connector_authorize":
		return CapabilityApplyAuthorize
	case "reload", "runtime_reload", "refresh_runtime":
		return CapabilityApplyReload
	case "rollback", "revert", "undo":
		return CapabilityApplyRollback
	default:
		return ""
	}
}

type HostOwnedCapabilityApplyDescriptor struct {
	ContractVersion              string                  `json:"contract_version,omitempty"`
	Projected                    bool                    `json:"projected"`
	Available                    bool                    `json:"available"`
	Status                       HostActionStatus        `json:"status,omitempty"`
	Mode                         string                  `json:"mode,omitempty"`
	ReadyForCapabilityApply      bool                    `json:"ready_for_capability_apply"`
	CapabilityDescriptorRef      DisplaySafeRef          `json:"capability_descriptor_ref,omitempty"`
	CapabilityAdapterRef         DisplaySafeRef          `json:"capability_adapter_ref,omitempty"`
	OwnerRef                     DisplaySafeRef          `json:"owner_ref,omitempty"`
	SupportsInstall              bool                    `json:"supports_install"`
	SupportsEnable               bool                    `json:"supports_enable"`
	SupportsSkillApply           bool                    `json:"supports_skill_apply"`
	SupportsAuthorize            bool                    `json:"supports_authorize"`
	SupportsReload               bool                    `json:"supports_reload"`
	SupportsRollback             bool                    `json:"supports_rollback"`
	SupportedActions             []CapabilityApplyAction `json:"supported_actions,omitempty"`
	ProposalContractRef          DisplaySafeRef          `json:"proposal_contract_ref,omitempty"`
	CandidateContractRef         DisplaySafeRef          `json:"candidate_contract_ref,omitempty"`
	GuardContractRef             DisplaySafeRef          `json:"guard_contract_ref,omitempty"`
	DryRunContractRef            DisplaySafeRef          `json:"dry_run_contract_ref,omitempty"`
	IdempotencyContractRef       DisplaySafeRef          `json:"idempotency_contract_ref,omitempty"`
	ReadbackContractRef          DisplaySafeRef          `json:"readback_contract_ref,omitempty"`
	RollbackContractRef          DisplaySafeRef          `json:"rollback_contract_ref,omitempty"`
	ApprovalPolicyRef            DisplaySafeRef          `json:"approval_policy_ref,omitempty"`
	RedactionPolicyRef           DisplaySafeRef          `json:"redaction_policy_ref,omitempty"`
	TimeoutPolicyRef             DisplaySafeRef          `json:"timeout_policy_ref,omitempty"`
	PolicyRefs                   []DisplaySafeRef        `json:"policy_refs,omitempty"`
	RequiredApprovalRefs         []DisplaySafeRef        `json:"required_approval_refs,omitempty"`
	MissingInputs                []MissingInput          `json:"missing_inputs,omitempty"`
	BlockedReasons               []string                `json:"blocked_reasons,omitempty"`
	FailureClass                 FailureClass            `json:"failure_class,omitempty"`
	Boundaries                   []Boundary              `json:"boundaries,omitempty"`
	NextHostAction               NextHostAction          `json:"next_host_action,omitempty"`
	RunnerEffect                 string                  `json:"runner_effect,omitempty"`
	PromptEffect                 string                  `json:"prompt_effect,omitempty"`
	RuntimeEffect                string                  `json:"runtime_effect,omitempty"`
	CoreInvocationExecuted       bool                    `json:"core_invocation_executed"`
	InstallerExecutedByCore      bool                    `json:"installer_executed_by_core"`
	InstallExecutedByCore        bool                    `json:"install_executed_by_core"`
	EnableExecutedByCore         bool                    `json:"enable_executed_by_core"`
	PackageManagerExecutedByCore bool                    `json:"package_manager_executed_by_core"`
	SkillWriteByCore             bool                    `json:"skill_write_by_core"`
	RuntimeReloadByCore          bool                    `json:"runtime_reload_by_core"`
	RawOutputLoaded              bool                    `json:"raw_output_loaded"`
}

type HostOwnedCapabilityApplyRequestInput struct {
	Descriptor                    HostOwnedCapabilityApplyDescriptor     `json:"descriptor,omitempty"`
	IndependentGate               ProductionAdapterIndependentEffectGate `json:"independent_gate,omitempty"`
	FinalGate                     IntensityGateResult                    `json:"final_gate,omitempty"`
	Action                        CapabilityApplyAction                  `json:"action,omitempty"`
	CapabilityApplyRequestRef     DisplaySafeRef                         `json:"capability_apply_request_ref,omitempty"`
	CapabilityProposalRef         DisplaySafeRef                         `json:"capability_proposal_ref,omitempty"`
	CapabilityCandidateRef        DisplaySafeRef                         `json:"capability_candidate_ref,omitempty"`
	CapabilityGuardRef            DisplaySafeRef                         `json:"capability_guard_ref,omitempty"`
	CapabilityDryRunProofRef      DisplaySafeRef                         `json:"capability_dry_run_proof_ref,omitempty"`
	StrategyRef                   DisplaySafeRef                         `json:"strategy_ref,omitempty"`
	ObjectiveRunRef               DisplaySafeRef                         `json:"objective_run_ref,omitempty"`
	TargetCapabilityRef           DisplaySafeRef                         `json:"target_capability_ref,omitempty"`
	ExpectedCapabilityRef         DisplaySafeRef                         `json:"expected_capability_ref,omitempty"`
	ExpectedCapabilityStateRef    DisplaySafeRef                         `json:"expected_capability_state_ref,omitempty"`
	HostCapabilityConfirmationRef DisplaySafeRef                         `json:"host_capability_confirmation_ref,omitempty"`
	IdempotencyRef                DisplaySafeRef                         `json:"idempotency_ref,omitempty"`
	ExpectedCapabilityResultRef   DisplaySafeRef                         `json:"expected_capability_result_ref,omitempty"`
	ExpectedReadbackRef           DisplaySafeRef                         `json:"expected_readback_ref,omitempty"`
	RollbackPathRef               DisplaySafeRef                         `json:"rollback_path_ref,omitempty"`
	ApprovalRefs                  []DisplaySafeRef                       `json:"approval_refs,omitempty"`
	EvidenceRefs                  []EvidenceRef                          `json:"evidence_refs,omitempty"`
	Boundaries                    []Boundary                             `json:"boundaries,omitempty"`
	RawOutputLoaded               bool                                   `json:"raw_output_loaded"`
}

type HostOwnedCapabilityApplyRequest struct {
	ContractVersion                string                                 `json:"contract_version,omitempty"`
	Projected                      bool                                   `json:"projected"`
	Available                      bool                                   `json:"available"`
	Status                         HostActionStatus                       `json:"status,omitempty"`
	Mode                           string                                 `json:"mode,omitempty"`
	ReadyForHostCapabilityApply    bool                                   `json:"ready_for_host_capability_apply"`
	HostCapabilityApplyAuthorized  bool                                   `json:"host_capability_apply_authorized"`
	HostMayApplyCapabilityMutation bool                                   `json:"host_may_apply_capability_mutation"`
	Descriptor                     HostOwnedCapabilityApplyDescriptor     `json:"descriptor,omitempty"`
	IndependentGate                ProductionAdapterIndependentEffectGate `json:"independent_gate,omitempty"`
	FinalGate                      IntensityGateResult                    `json:"final_gate,omitempty"`
	Action                         CapabilityApplyAction                  `json:"action,omitempty"`
	CapabilityApplyRequestRef      DisplaySafeRef                         `json:"capability_apply_request_ref,omitempty"`
	CapabilityDescriptorRef        DisplaySafeRef                         `json:"capability_descriptor_ref,omitempty"`
	CapabilityAdapterRef           DisplaySafeRef                         `json:"capability_adapter_ref,omitempty"`
	OwnerRef                       DisplaySafeRef                         `json:"owner_ref,omitempty"`
	GateRef                        DisplaySafeRef                         `json:"gate_ref,omitempty"`
	PolicyRef                      DisplaySafeRef                         `json:"policy_ref,omitempty"`
	CapabilityProposalRef          DisplaySafeRef                         `json:"capability_proposal_ref,omitempty"`
	CapabilityCandidateRef         DisplaySafeRef                         `json:"capability_candidate_ref,omitempty"`
	CapabilityGuardRef             DisplaySafeRef                         `json:"capability_guard_ref,omitempty"`
	CapabilityDryRunProofRef       DisplaySafeRef                         `json:"capability_dry_run_proof_ref,omitempty"`
	StrategyRef                    DisplaySafeRef                         `json:"strategy_ref,omitempty"`
	ObjectiveRunRef                DisplaySafeRef                         `json:"objective_run_ref,omitempty"`
	TargetCapabilityRef            DisplaySafeRef                         `json:"target_capability_ref,omitempty"`
	ExpectedCapabilityRef          DisplaySafeRef                         `json:"expected_capability_ref,omitempty"`
	ExpectedCapabilityStateRef     DisplaySafeRef                         `json:"expected_capability_state_ref,omitempty"`
	HostCapabilityConfirmationRef  DisplaySafeRef                         `json:"host_capability_confirmation_ref,omitempty"`
	IdempotencyRef                 DisplaySafeRef                         `json:"idempotency_ref,omitempty"`
	IdempotencyContractRef         DisplaySafeRef                         `json:"idempotency_contract_ref,omitempty"`
	ExpectedCapabilityResultRef    DisplaySafeRef                         `json:"expected_capability_result_ref,omitempty"`
	ExpectedReadbackRef            DisplaySafeRef                         `json:"expected_readback_ref,omitempty"`
	RollbackPathRef                DisplaySafeRef                         `json:"rollback_path_ref,omitempty"`
	ApprovalRefs                   []DisplaySafeRef                       `json:"approval_refs,omitempty"`
	EvidenceRefs                   []EvidenceRef                          `json:"evidence_refs,omitempty"`
	MissingInputs                  []MissingInput                         `json:"missing_inputs,omitempty"`
	BlockedReasons                 []string                               `json:"blocked_reasons,omitempty"`
	FailureClass                   FailureClass                           `json:"failure_class,omitempty"`
	Boundaries                     []Boundary                             `json:"boundaries,omitempty"`
	NextHostAction                 NextHostAction                         `json:"next_host_action,omitempty"`
	RunnerEffect                   string                                 `json:"runner_effect,omitempty"`
	PromptEffect                   string                                 `json:"prompt_effect,omitempty"`
	RuntimeEffect                  string                                 `json:"runtime_effect,omitempty"`
	CoreInvocationExecuted         bool                                   `json:"core_invocation_executed"`
	InstallerExecutedByCore        bool                                   `json:"installer_executed_by_core"`
	InstallExecutedByCore          bool                                   `json:"install_executed_by_core"`
	EnableExecutedByCore           bool                                   `json:"enable_executed_by_core"`
	PackageManagerExecutedByCore   bool                                   `json:"package_manager_executed_by_core"`
	SkillWriteByCore               bool                                   `json:"skill_write_by_core"`
	RuntimeReloadByCore            bool                                   `json:"runtime_reload_by_core"`
	RawOutputLoaded                bool                                   `json:"raw_output_loaded"`
}

type HostOwnedCapabilityApplyResultInput struct {
	Request                      HostOwnedCapabilityApplyRequest `json:"request,omitempty"`
	CapabilityApplyResultRef     DisplaySafeRef                  `json:"capability_apply_result_ref,omitempty"`
	HostCapabilityRunRef         DisplaySafeRef                  `json:"host_capability_run_ref,omitempty"`
	HostCapabilityApplyReported  bool                            `json:"host_capability_apply_reported"`
	HostCapabilityApplySucceeded bool                            `json:"host_capability_apply_succeeded"`
	HostCapabilityApplyFailed    bool                            `json:"host_capability_apply_failed"`
	AppliedCapabilityRef         DisplaySafeRef                  `json:"applied_capability_ref,omitempty"`
	AppliedCapabilityStateRef    DisplaySafeRef                  `json:"applied_capability_state_ref,omitempty"`
	FailureRef                   DisplaySafeRef                  `json:"failure_ref,omitempty"`
	CompensationRef              DisplaySafeRef                  `json:"compensation_ref,omitempty"`
	CapabilityEvidenceRefs       []DisplaySafeRef                `json:"capability_evidence_refs,omitempty"`
	RawOutputLoaded              bool                            `json:"raw_output_loaded"`
}

type HostOwnedCapabilityApplyResult struct {
	ContractVersion              string                          `json:"contract_version,omitempty"`
	Projected                    bool                            `json:"projected"`
	Available                    bool                            `json:"available"`
	Status                       HostActionStatus                `json:"status,omitempty"`
	Mode                         string                          `json:"mode,omitempty"`
	ReadyForCapabilityReadback   bool                            `json:"ready_for_capability_readback"`
	HostCapabilityApplyReported  bool                            `json:"host_capability_apply_reported"`
	HostCapabilityApplySucceeded bool                            `json:"host_capability_apply_succeeded"`
	HostCapabilityApplyFailed    bool                            `json:"host_capability_apply_failed"`
	HostCapabilityApplyRecorded  bool                            `json:"host_capability_apply_recorded"`
	HostCapabilityInstalled      bool                            `json:"host_capability_installed"`
	HostCapabilityEnabled        bool                            `json:"host_capability_enabled"`
	HostSkillApplied             bool                            `json:"host_skill_applied"`
	HostCapabilityAuthorized     bool                            `json:"host_capability_authorized"`
	HostRuntimeReloaded          bool                            `json:"host_runtime_reloaded"`
	HostCapabilityRolledBack     bool                            `json:"host_capability_rolled_back"`
	Request                      HostOwnedCapabilityApplyRequest `json:"request,omitempty"`
	Action                       CapabilityApplyAction           `json:"action,omitempty"`
	CapabilityApplyResultRef     DisplaySafeRef                  `json:"capability_apply_result_ref,omitempty"`
	ExpectedCapabilityResultRef  DisplaySafeRef                  `json:"expected_capability_result_ref,omitempty"`
	CapabilityApplyRequestRef    DisplaySafeRef                  `json:"capability_apply_request_ref,omitempty"`
	CapabilityAdapterRef         DisplaySafeRef                  `json:"capability_adapter_ref,omitempty"`
	HostCapabilityRunRef         DisplaySafeRef                  `json:"host_capability_run_ref,omitempty"`
	ExpectedCapabilityRef        DisplaySafeRef                  `json:"expected_capability_ref,omitempty"`
	ExpectedCapabilityStateRef   DisplaySafeRef                  `json:"expected_capability_state_ref,omitempty"`
	ExpectedReadbackRef          DisplaySafeRef                  `json:"expected_readback_ref,omitempty"`
	AppliedCapabilityRef         DisplaySafeRef                  `json:"applied_capability_ref,omitempty"`
	AppliedCapabilityStateRef    DisplaySafeRef                  `json:"applied_capability_state_ref,omitempty"`
	RollbackPathRef              DisplaySafeRef                  `json:"rollback_path_ref,omitempty"`
	FailureRef                   DisplaySafeRef                  `json:"failure_ref,omitempty"`
	CompensationRef              DisplaySafeRef                  `json:"compensation_ref,omitempty"`
	CapabilityEvidenceRefs       []DisplaySafeRef                `json:"capability_evidence_refs,omitempty"`
	MissingInputs                []MissingInput                  `json:"missing_inputs,omitempty"`
	BlockedReasons               []string                        `json:"blocked_reasons,omitempty"`
	FailureClass                 FailureClass                    `json:"failure_class,omitempty"`
	Boundaries                   []Boundary                      `json:"boundaries,omitempty"`
	NextHostAction               NextHostAction                  `json:"next_host_action,omitempty"`
	RunnerEffect                 string                          `json:"runner_effect,omitempty"`
	PromptEffect                 string                          `json:"prompt_effect,omitempty"`
	RuntimeEffect                string                          `json:"runtime_effect,omitempty"`
	CoreInvocationExecuted       bool                            `json:"core_invocation_executed"`
	InstallerExecutedByCore      bool                            `json:"installer_executed_by_core"`
	InstallExecutedByCore        bool                            `json:"install_executed_by_core"`
	EnableExecutedByCore         bool                            `json:"enable_executed_by_core"`
	PackageManagerExecutedByCore bool                            `json:"package_manager_executed_by_core"`
	SkillWriteByCore             bool                            `json:"skill_write_by_core"`
	RuntimeReloadByCore          bool                            `json:"runtime_reload_by_core"`
	RawOutputLoaded              bool                            `json:"raw_output_loaded"`
}

type HostOwnedCapabilityApplyReadbackInput struct {
	CapabilityReadbackRef      DisplaySafeRef                 `json:"capability_readback_ref,omitempty"`
	Result                     HostOwnedCapabilityApplyResult `json:"result,omitempty"`
	ObservedCapabilityRef      DisplaySafeRef                 `json:"observed_capability_ref,omitempty"`
	ObservedCapabilityStateRef DisplaySafeRef                 `json:"observed_capability_state_ref,omitempty"`
	ObservedRollbackPathRef    DisplaySafeRef                 `json:"observed_rollback_path_ref,omitempty"`
	ReadbackEvidenceRefs       []DisplaySafeRef               `json:"readback_evidence_refs,omitempty"`
	RawOutputLoaded            bool                           `json:"raw_output_loaded"`
}

type HostOwnedCapabilityApplyReadback struct {
	ContractVersion                 string                         `json:"contract_version,omitempty"`
	Projected                       bool                           `json:"projected"`
	Available                       bool                           `json:"available"`
	Status                          HostActionStatus               `json:"status,omitempty"`
	Mode                            string                         `json:"mode,omitempty"`
	CapabilityReadbackBound         bool                           `json:"capability_readback_bound"`
	RollbackPathVerified            bool                           `json:"rollback_path_verified"`
	ReadyForRuntimeLoopContinuation bool                           `json:"ready_for_runtime_loop_continuation"`
	Result                          HostOwnedCapabilityApplyResult `json:"result,omitempty"`
	Action                          CapabilityApplyAction          `json:"action,omitempty"`
	CapabilityReadbackRef           DisplaySafeRef                 `json:"capability_readback_ref,omitempty"`
	CapabilityApplyResultRef        DisplaySafeRef                 `json:"capability_apply_result_ref,omitempty"`
	CapabilityApplyRequestRef       DisplaySafeRef                 `json:"capability_apply_request_ref,omitempty"`
	CapabilityAdapterRef            DisplaySafeRef                 `json:"capability_adapter_ref,omitempty"`
	HostCapabilityRunRef            DisplaySafeRef                 `json:"host_capability_run_ref,omitempty"`
	ExpectedCapabilityRef           DisplaySafeRef                 `json:"expected_capability_ref,omitempty"`
	ExpectedCapabilityStateRef      DisplaySafeRef                 `json:"expected_capability_state_ref,omitempty"`
	ExpectedReadbackRef             DisplaySafeRef                 `json:"expected_readback_ref,omitempty"`
	AppliedCapabilityRef            DisplaySafeRef                 `json:"applied_capability_ref,omitempty"`
	AppliedCapabilityStateRef       DisplaySafeRef                 `json:"applied_capability_state_ref,omitempty"`
	ObservedCapabilityRef           DisplaySafeRef                 `json:"observed_capability_ref,omitempty"`
	ObservedCapabilityStateRef      DisplaySafeRef                 `json:"observed_capability_state_ref,omitempty"`
	RollbackPathRef                 DisplaySafeRef                 `json:"rollback_path_ref,omitempty"`
	ObservedRollbackPathRef         DisplaySafeRef                 `json:"observed_rollback_path_ref,omitempty"`
	ReadbackEvidenceRefs            []DisplaySafeRef               `json:"readback_evidence_refs,omitempty"`
	MissingInputs                   []MissingInput                 `json:"missing_inputs,omitempty"`
	BlockedReasons                  []string                       `json:"blocked_reasons,omitempty"`
	FailureClass                    FailureClass                   `json:"failure_class,omitempty"`
	Boundaries                      []Boundary                     `json:"boundaries,omitempty"`
	NextHostAction                  NextHostAction                 `json:"next_host_action,omitempty"`
	RunnerEffect                    string                         `json:"runner_effect,omitempty"`
	PromptEffect                    string                         `json:"prompt_effect,omitempty"`
	RuntimeEffect                   string                         `json:"runtime_effect,omitempty"`
	CoreInvocationExecuted          bool                           `json:"core_invocation_executed"`
	InstallerExecutedByCore         bool                           `json:"installer_executed_by_core"`
	InstallExecutedByCore           bool                           `json:"install_executed_by_core"`
	EnableExecutedByCore            bool                           `json:"enable_executed_by_core"`
	PackageManagerExecutedByCore    bool                           `json:"package_manager_executed_by_core"`
	SkillWriteByCore                bool                           `json:"skill_write_by_core"`
	RuntimeReloadByCore             bool                           `json:"runtime_reload_by_core"`
	RawOutputLoaded                 bool                           `json:"raw_output_loaded"`
}

func BuildHostOwnedCapabilityApplyDescriptor(input HostOwnedCapabilityApplyDescriptor) HostOwnedCapabilityApplyDescriptor {
	unsafeInput := hostOwnedCapabilityApplyDescriptorOutputUnsafe(input)
	result := input.Normalize()
	result.Status = HostActionBlocked
	result.ReadyForCapabilityApply = false
	result.FailureClass = firstFailureClass(result.FailureClass, FailureNone)
	result.Boundaries = hostOwnedCapabilityApplyDescriptorBoundaries(result.Boundaries)
	if unsafeInput {
		result = hostOwnedCapabilityApplyDescriptorBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		result.RawOutputLoaded = true
		return result.Normalize()
	}
	if !result.Available {
		result.Status = HostActionNotReady
		result.FailureClass = firstFailureClass(result.FailureClass, FailureConfigMissing)
		result.NextHostAction = firstNextHostAction(result.NextHostAction, "configure_capability_apply_adapter")
		return result.Normalize()
	}
	if result.CapabilityDescriptorRef == "" {
		result = hostOwnedCapabilityApplyDescriptorBlock(result, FailureEvidenceMissing, "capability_descriptor_ref_missing", "host:capability_descriptor_ref", "provide_capability_descriptor_ref")
	}
	if result.CapabilityAdapterRef == "" {
		result = hostOwnedCapabilityApplyDescriptorBlock(result, FailureHostAdapterMissing, "capability_adapter_ref_missing", "host:capability_adapter_ref", "configure_capability_adapter")
	}
	if result.OwnerRef == "" {
		result = hostOwnedCapabilityApplyDescriptorBlock(result, FailureConfigMissing, "capability_owner_ref_missing", "host:capability_owner_ref", "provide_capability_owner_ref")
	}
	for _, check := range []struct {
		ok      bool
		reason  string
		missing MissingInput
	}{
		{len(result.SupportedActions) > 0, "capability_actions_missing", "host:capability_supported_actions"},
		{result.SupportsRollback, "capability_rollback_path_not_supported", "host:capability_rollback_path"},
	} {
		if !check.ok {
			result = hostOwnedCapabilityApplyDescriptorBlock(result, FailureConfigMissing, check.reason, check.missing, "configure_capability_lifecycle_controls")
		}
	}
	for _, check := range []struct {
		ref     DisplaySafeRef
		reason  string
		missing MissingInput
		next    NextHostAction
	}{
		{result.ProposalContractRef, "capability_proposal_contract_ref_missing", "contract:capability_proposal", "provide_capability_proposal_contract"},
		{result.CandidateContractRef, "capability_candidate_contract_ref_missing", "contract:capability_candidate", "provide_capability_candidate_contract"},
		{result.GuardContractRef, "capability_guard_contract_ref_missing", "contract:capability_guard", "provide_capability_guard_contract"},
		{result.DryRunContractRef, "capability_dry_run_contract_ref_missing", "contract:capability_dry_run", "provide_capability_dry_run_contract"},
		{result.IdempotencyContractRef, "capability_idempotency_contract_ref_missing", "contract:capability_idempotency", "provide_capability_idempotency_contract"},
		{result.ReadbackContractRef, "capability_readback_contract_ref_missing", "contract:capability_readback", "provide_capability_readback_contract"},
		{result.RollbackContractRef, "capability_rollback_contract_ref_missing", "contract:capability_rollback", "provide_capability_rollback_contract"},
		{result.ApprovalPolicyRef, "capability_approval_policy_ref_missing", "policy:capability_apply_approval", "provide_capability_apply_approval_policy"},
	} {
		if check.ref == "" {
			result = hostOwnedCapabilityApplyDescriptorBlock(result, FailureConfigMissing, check.reason, check.missing, check.next)
		}
	}
	if len(result.MissingInputs) == 0 && len(result.BlockedReasons) == 0 {
		result.Status = HostActionReady
		result.ReadyForCapabilityApply = true
		result.NextHostAction = "host_may_prepare_capability_apply_request"
		result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_capability_apply_request")
	}
	return result.Normalize()
}

func BuildHostOwnedCapabilityApplyRequest(input HostOwnedCapabilityApplyRequestInput) HostOwnedCapabilityApplyRequest {
	if hostOwnedCapabilityApplyDescriptorEmpty(input.Descriptor) {
		return unavailableHostOwnedCapabilityApplyRequest()
	}
	descriptor := BuildHostOwnedCapabilityApplyDescriptor(input.Descriptor)
	gate := input.IndependentGate.Normalize()
	finalGate := input.FinalGate.Normalize()
	action := NormalizeCapabilityApplyAction(string(input.Action))
	approvalRefs := normalizeDisplaySafeRefs(append(append(cloneDisplaySafeRefs(input.ApprovalRefs), gate.ApprovalRef), finalGate.ApprovalRefs...))
	result := HostOwnedCapabilityApplyRequest{
		ContractVersion:               ContractVersion,
		Projected:                     true,
		Available:                     descriptor.Available,
		Status:                        HostActionBlocked,
		Mode:                          "host_owned_capability_apply_request",
		Descriptor:                    descriptor,
		IndependentGate:               gate,
		FinalGate:                     finalGate,
		Action:                        action,
		CapabilityApplyRequestRef:     normalizeOneDisplaySafeRef(input.CapabilityApplyRequestRef),
		CapabilityDescriptorRef:       descriptor.CapabilityDescriptorRef,
		CapabilityAdapterRef:          descriptor.CapabilityAdapterRef,
		OwnerRef:                      descriptor.OwnerRef,
		GateRef:                       gate.GateRef,
		PolicyRef:                     gate.PolicyRef,
		CapabilityProposalRef:         normalizeOneDisplaySafeRef(input.CapabilityProposalRef),
		CapabilityCandidateRef:        normalizeOneDisplaySafeRef(input.CapabilityCandidateRef),
		CapabilityGuardRef:            normalizeOneDisplaySafeRef(input.CapabilityGuardRef),
		CapabilityDryRunProofRef:      normalizeOneDisplaySafeRef(input.CapabilityDryRunProofRef),
		StrategyRef:                   firstDisplaySafeRef(input.StrategyRef, finalGate.StrategyRef),
		ObjectiveRunRef:               normalizeOneDisplaySafeRef(input.ObjectiveRunRef),
		TargetCapabilityRef:           normalizeOneDisplaySafeRef(input.TargetCapabilityRef),
		ExpectedCapabilityRef:         normalizeOneDisplaySafeRef(input.ExpectedCapabilityRef),
		ExpectedCapabilityStateRef:    normalizeOneDisplaySafeRef(input.ExpectedCapabilityStateRef),
		HostCapabilityConfirmationRef: normalizeOneDisplaySafeRef(input.HostCapabilityConfirmationRef),
		IdempotencyRef:                firstDisplaySafeRef(input.IdempotencyRef, gate.IdempotencyRef),
		IdempotencyContractRef:        descriptor.IdempotencyContractRef,
		ExpectedCapabilityResultRef:   normalizeOneDisplaySafeRef(input.ExpectedCapabilityResultRef),
		ExpectedReadbackRef:           firstDisplaySafeRef(input.ExpectedReadbackRef, gate.ReadbackRef),
		RollbackPathRef:               normalizeOneDisplaySafeRef(input.RollbackPathRef),
		ApprovalRefs:                  approvalRefs,
		EvidenceRefs:                  MergeEvidenceRefs(input.EvidenceRefs, finalGate.EvidenceRefs),
		FailureClass:                  FailureNone,
		Boundaries:                    hostOwnedCapabilityApplyRequestBoundaries(descriptor.Boundaries, gate.Boundaries, finalGate.Boundaries, input.Boundaries),
		NextHostAction:                "prepare_capability_apply_request",
		RunnerEffect:                  "none",
		PromptEffect:                  "none",
		RuntimeEffect:                 "none",
		RawOutputLoaded:               input.RawOutputLoaded || descriptor.RawOutputLoaded || gate.RawOutputLoaded || finalGate.RawOutputLoaded,
	}
	if hostOwnedCapabilityApplyRequestUnsafe(input, descriptor, gate, finalGate) {
		result = hostOwnedCapabilityApplyRequestBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if !descriptor.ReadyForCapabilityApply {
		result = hostOwnedCapabilityApplyRequestBlock(result, firstFailureClass(descriptor.FailureClass, FailureConfigMissing), "capability_descriptor_not_ready", "host:capability_descriptor", firstNextHostAction(descriptor.NextHostAction, "review_capability_descriptor"))
	}
	if gate.Kind != ProductionAdapterEffectGateInstallerApply || !gate.ReadyForIndependentGatePlan {
		result = hostOwnedCapabilityApplyRequestBlock(result, firstFailureClass(gate.FailureClass, FailurePolicyBlocked), "capability_independent_gate_not_ready", "host:installer_apply_independent_gate", firstNextHostAction(gate.NextHostAction, "review_installer_apply_independent_gate"))
	}
	if gate.AdapterRef != "" && descriptor.CapabilityAdapterRef != "" && gate.AdapterRef != descriptor.CapabilityAdapterRef {
		result = hostOwnedCapabilityApplyRequestBlock(result, FailureVerificationFailed, "capability_gate_adapter_ref_mismatch", "host:capability_adapter_ref", "review_capability_apply_gate")
	}
	if gate.IdempotencyRef != "" && result.IdempotencyRef != "" && gate.IdempotencyRef != result.IdempotencyRef {
		result = hostOwnedCapabilityApplyRequestBlock(result, FailureVerificationFailed, "capability_gate_idempotency_ref_mismatch", "host:capability_idempotency_ref", "review_capability_apply_gate")
	}
	if gate.ReadbackRef != "" && result.ExpectedReadbackRef != "" && gate.ReadbackRef != result.ExpectedReadbackRef {
		result = hostOwnedCapabilityApplyRequestBlock(result, FailureVerificationFailed, "capability_gate_readback_ref_mismatch", "host:capability_readback_ref", "review_capability_apply_gate")
	}
	if finalGate.Stage != IntensityGateFinal || !finalGate.Allowed {
		result = hostOwnedCapabilityApplyRequestBlock(result, firstFailureClass(finalGate.FailureClass, FailurePolicyBlocked), "final_gate_not_satisfied", "host:execution_intensity_final_gate", firstNextHostAction(finalGate.NextHostAction, "run_strategy_final_gate"))
	} else if executionIntensityRank(finalGate.ApprovedIntensity) < executionIntensityRank(IntensityL4DurableLongRun) {
		result = hostOwnedCapabilityApplyRequestBlock(result, FailurePolicyBlocked, "capability_apply_requires_l4", "contract:l4_durable_long_run", "request_l4_capability_approval")
	}
	if finalGate.StrategyRef != "" && result.StrategyRef != "" && finalGate.StrategyRef != result.StrategyRef {
		result = hostOwnedCapabilityApplyRequestBlock(result, FailureVerificationFailed, "final_gate_strategy_ref_mismatch", "host:execution_intensity_final_gate", "run_strategy_final_gate")
	}
	if action == "" {
		result = hostOwnedCapabilityApplyRequestBlock(result, FailureInvalidInput, "capability_apply_action_missing", "host:capability_apply_action", "provide_capability_apply_action")
	} else if !capabilityApplyActionContains(descriptor.SupportedActions, action) {
		result = hostOwnedCapabilityApplyRequestBlock(result, FailureUnsupportedOperation, "capability_apply_action_not_supported", "host:capability_supported_actions", "select_supported_capability_action")
	}
	if gate.ApprovalRef != "" &&
		result.HostCapabilityConfirmationRef != gate.ApprovalRef &&
		!capabilityApplyDisplaySafeRefContains(result.ApprovalRefs, gate.ApprovalRef) {
		result = hostOwnedCapabilityApplyRequestBlock(result, FailureApprovalRequired, "capability_gate_approval_ref_missing", "host:capability_gate_approval_ref", "provide_capability_apply_approval")
	}
	for _, check := range []struct {
		ref     DisplaySafeRef
		reason  string
		missing MissingInput
		next    NextHostAction
		failure FailureClass
	}{
		{result.CapabilityApplyRequestRef, "capability_apply_request_ref_missing", "host:capability_apply_request_ref", "provide_capability_apply_request_ref", FailureEvidenceMissing},
		{result.CapabilityProposalRef, "capability_proposal_ref_missing", "host:capability_proposal_ref", "provide_capability_proposal_ref", FailureEvidenceMissing},
		{result.CapabilityCandidateRef, "capability_candidate_ref_missing", "host:capability_candidate_ref", "provide_capability_candidate_ref", FailureEvidenceMissing},
		{result.CapabilityGuardRef, "capability_guard_ref_missing", "host:capability_guard_ref", "provide_capability_guard_ref", FailureEvidenceMissing},
		{result.CapabilityDryRunProofRef, "capability_dry_run_proof_ref_missing", "host:capability_dry_run_proof_ref", "provide_capability_dry_run_proof_ref", FailureEvidenceMissing},
		{result.StrategyRef, "capability_strategy_ref_missing", "host:strategy_ref", "provide_capability_strategy_ref", FailureEvidenceMissing},
		{result.ObjectiveRunRef, "capability_objective_run_ref_missing", "host:objective_run_ref", "provide_capability_objective_run_ref", FailureEvidenceMissing},
		{result.ExpectedCapabilityRef, "expected_capability_ref_missing", "host:expected_capability_ref", "provide_expected_capability_ref", FailureEvidenceMissing},
		{result.ExpectedCapabilityStateRef, "expected_capability_state_ref_missing", "host:expected_capability_state_ref", "provide_expected_capability_state_ref", FailureEvidenceMissing},
		{result.HostCapabilityConfirmationRef, "host_capability_confirmation_ref_missing", "host:capability_confirmation_ref", "request_capability_apply_confirmation", FailureApprovalRequired},
		{result.IdempotencyRef, "capability_idempotency_ref_missing", "host:capability_idempotency_ref", "provide_capability_idempotency_ref", FailureConfigMissing},
		{result.ExpectedCapabilityResultRef, "expected_capability_result_ref_missing", "host:expected_capability_result_ref", "provide_expected_capability_result_ref", FailureEvidenceMissing},
		{result.ExpectedReadbackRef, "expected_capability_readback_ref_missing", "host:expected_capability_readback_ref", "provide_expected_capability_readback_ref", FailureEvidenceMissing},
		{result.RollbackPathRef, "capability_rollback_path_ref_missing", "host:capability_rollback_path_ref", "provide_capability_rollback_path_ref", FailureConfigMissing},
	} {
		if check.ref == "" {
			result = hostOwnedCapabilityApplyRequestBlock(result, check.failure, check.reason, check.missing, check.next)
		}
	}
	if action == CapabilityApplyEnable || action == CapabilityApplyReload || action == CapabilityApplyRollback {
		if result.TargetCapabilityRef == "" {
			result = hostOwnedCapabilityApplyRequestBlock(result, FailureEvidenceMissing, "target_capability_ref_missing", "host:target_capability_ref", "provide_target_capability_ref")
		} else if result.ExpectedCapabilityRef != "" && result.TargetCapabilityRef != result.ExpectedCapabilityRef {
			result = hostOwnedCapabilityApplyRequestBlock(result, FailureVerificationFailed, "target_capability_ref_mismatch", "host:target_capability_ref", "review_capability_apply_request")
		}
	}
	if len(result.MissingInputs) == 0 && len(result.BlockedReasons) == 0 {
		result.Status = HostActionReady
		result.ReadyForHostCapabilityApply = true
		result.HostCapabilityApplyAuthorized = true
		result.HostMayApplyCapabilityMutation = true
		result.NextHostAction = "host_may_apply_capability_mutation"
		result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_host_capability_apply", "host_may_apply_capability_mutation", "core_capability_apply_not_executed")
	}
	return result.Normalize()
}

func BuildHostOwnedCapabilityApplyResult(input HostOwnedCapabilityApplyResultInput) HostOwnedCapabilityApplyResult {
	if hostOwnedCapabilityApplyRequestEmpty(input.Request) {
		return unavailableHostOwnedCapabilityApplyResult()
	}
	request := input.Request.Normalize()
	result := HostOwnedCapabilityApplyResult{
		ContractVersion:              ContractVersion,
		Projected:                    true,
		Available:                    request.Available,
		Status:                       HostActionBlocked,
		Mode:                         "host_owned_capability_apply_result",
		HostCapabilityApplyReported:  input.HostCapabilityApplyReported,
		HostCapabilityApplySucceeded: input.HostCapabilityApplySucceeded,
		HostCapabilityApplyFailed:    input.HostCapabilityApplyFailed,
		Request:                      request,
		Action:                       request.Action,
		CapabilityApplyResultRef:     normalizeOneDisplaySafeRef(input.CapabilityApplyResultRef),
		ExpectedCapabilityResultRef:  request.ExpectedCapabilityResultRef,
		CapabilityApplyRequestRef:    request.CapabilityApplyRequestRef,
		CapabilityAdapterRef:         request.CapabilityAdapterRef,
		HostCapabilityRunRef:         normalizeOneDisplaySafeRef(input.HostCapabilityRunRef),
		ExpectedCapabilityRef:        request.ExpectedCapabilityRef,
		ExpectedCapabilityStateRef:   request.ExpectedCapabilityStateRef,
		ExpectedReadbackRef:          request.ExpectedReadbackRef,
		AppliedCapabilityRef:         normalizeOneDisplaySafeRef(input.AppliedCapabilityRef),
		AppliedCapabilityStateRef:    normalizeOneDisplaySafeRef(input.AppliedCapabilityStateRef),
		RollbackPathRef:              request.RollbackPathRef,
		FailureRef:                   normalizeOneDisplaySafeRef(input.FailureRef),
		CompensationRef:              normalizeOneDisplaySafeRef(input.CompensationRef),
		CapabilityEvidenceRefs:       normalizeDisplaySafeRefs(input.CapabilityEvidenceRefs),
		FailureClass:                 FailureNone,
		Boundaries:                   hostOwnedCapabilityApplyResultBoundaries(request.Boundaries),
		NextHostAction:               "provide_capability_apply_result",
		RunnerEffect:                 "none",
		PromptEffect:                 "none",
		RuntimeEffect:                "none",
		RawOutputLoaded:              input.RawOutputLoaded || request.RawOutputLoaded,
	}
	if hostOwnedCapabilityApplyResultUnsafe(input, request) {
		result = hostOwnedCapabilityApplyResultBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if !request.ReadyForHostCapabilityApply || !request.HostMayApplyCapabilityMutation {
		result = hostOwnedCapabilityApplyResultBlock(result, firstFailureClass(request.FailureClass, FailureConfigMissing), "capability_apply_request_not_ready", "host:capability_apply_request", firstNextHostAction(request.NextHostAction, "review_capability_apply_request"))
	}
	if !input.HostCapabilityApplyReported {
		result = hostOwnedCapabilityApplyResultBlock(result, FailureEvidenceMissing, "capability_apply_not_reported", "host:capability_apply_report", "provide_capability_apply_report")
	}
	if input.HostCapabilityApplySucceeded && input.HostCapabilityApplyFailed {
		result = hostOwnedCapabilityApplyResultBlock(result, FailureVerificationFailed, "capability_apply_status_conflict", "host:capability_apply_status", "review_capability_apply_result")
	}
	if result.HostCapabilityRunRef == "" {
		result = hostOwnedCapabilityApplyResultBlock(result, FailureEvidenceMissing, "host_capability_run_ref_missing", "host:capability_run_ref", "provide_capability_apply_report")
	}
	if result.CapabilityApplyResultRef == "" {
		result = hostOwnedCapabilityApplyResultBlock(result, FailureEvidenceMissing, "capability_apply_result_ref_missing", "host:capability_apply_result_ref", "provide_capability_apply_result_ref")
	} else if result.ExpectedCapabilityResultRef != "" && result.CapabilityApplyResultRef != result.ExpectedCapabilityResultRef {
		result = hostOwnedCapabilityApplyResultBlock(result, FailureVerificationFailed, "capability_apply_result_ref_mismatch", "host:capability_apply_result_ref", "review_capability_apply_result")
	}
	if len(result.MissingInputs) > 0 || len(result.BlockedReasons) > 0 {
		return result.Normalize()
	}
	if input.HostCapabilityApplyFailed {
		if result.FailureRef == "" {
			result = hostOwnedCapabilityApplyResultBlock(result, FailureEvidenceMissing, "capability_apply_failure_ref_missing", "host:capability_apply_failure_ref", "provide_capability_apply_failure_ref")
			return result.Normalize()
		}
		result.Status = HostActionReviewRequired
		result.HostCapabilityApplyRecorded = true
		result.FailureClass = FailureVerificationFailed
		result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, "capability_apply_failed")
		result.NextHostAction = "review_capability_apply_failure"
		result.Boundaries = AppendBoundaries(result.Boundaries, "capability_apply_failed", "compensation_not_executed")
		return result.Normalize()
	}
	if !input.HostCapabilityApplySucceeded {
		result = hostOwnedCapabilityApplyResultBlock(result, FailureEvidenceMissing, "capability_apply_status_missing", "host:capability_apply_status", "provide_capability_apply_report")
		return result.Normalize()
	}
	if result.AppliedCapabilityRef == "" {
		result = hostOwnedCapabilityApplyResultBlock(result, FailureEvidenceMissing, "applied_capability_ref_missing", "host:applied_capability_ref", "provide_capability_apply_result")
	} else if result.ExpectedCapabilityRef != "" && result.AppliedCapabilityRef != result.ExpectedCapabilityRef {
		result = hostOwnedCapabilityApplyResultBlock(result, FailureVerificationFailed, "capability_applied_ref_mismatch", "host:applied_capability_ref", "review_capability_apply_result")
	}
	if result.AppliedCapabilityStateRef == "" {
		result = hostOwnedCapabilityApplyResultBlock(result, FailureEvidenceMissing, "applied_capability_state_ref_missing", "host:applied_capability_state_ref", "provide_capability_apply_result")
	} else if result.ExpectedCapabilityStateRef != "" && result.AppliedCapabilityStateRef != result.ExpectedCapabilityStateRef {
		result = hostOwnedCapabilityApplyResultBlock(result, FailureVerificationFailed, "capability_state_ref_mismatch", "host:applied_capability_state_ref", "review_capability_apply_result")
	}
	if len(result.MissingInputs) == 0 && len(result.BlockedReasons) == 0 {
		result.Status = HostActionRecorded
		result.HostCapabilityApplyRecorded = true
		result.ReadyForCapabilityReadback = true
		result = hostOwnedCapabilityApplyResultMarkAction(result)
		result.NextHostAction = "bind_capability_apply_readback"
		result.Boundaries = AppendBoundaries(result.Boundaries, "host_capability_apply_recorded", "ready_for_capability_readback")
	}
	return result.Normalize()
}

func BuildHostOwnedCapabilityApplyReadback(input HostOwnedCapabilityApplyReadbackInput) HostOwnedCapabilityApplyReadback {
	if hostOwnedCapabilityApplyResultEmpty(input.Result) {
		return unavailableHostOwnedCapabilityApplyReadback()
	}
	applyResult := input.Result.Normalize()
	result := HostOwnedCapabilityApplyReadback{
		ContractVersion:            ContractVersion,
		Projected:                  true,
		Available:                  applyResult.Available,
		Status:                     HostActionBlocked,
		Mode:                       "host_owned_capability_apply_readback",
		Result:                     applyResult,
		Action:                     applyResult.Action,
		CapabilityReadbackRef:      normalizeOneDisplaySafeRef(input.CapabilityReadbackRef),
		CapabilityApplyResultRef:   applyResult.CapabilityApplyResultRef,
		CapabilityApplyRequestRef:  applyResult.CapabilityApplyRequestRef,
		CapabilityAdapterRef:       applyResult.CapabilityAdapterRef,
		HostCapabilityRunRef:       applyResult.HostCapabilityRunRef,
		ExpectedCapabilityRef:      applyResult.ExpectedCapabilityRef,
		ExpectedCapabilityStateRef: applyResult.ExpectedCapabilityStateRef,
		ExpectedReadbackRef:        applyResult.ExpectedReadbackRef,
		AppliedCapabilityRef:       applyResult.AppliedCapabilityRef,
		AppliedCapabilityStateRef:  applyResult.AppliedCapabilityStateRef,
		ObservedCapabilityRef:      normalizeOneDisplaySafeRef(input.ObservedCapabilityRef),
		ObservedCapabilityStateRef: normalizeOneDisplaySafeRef(input.ObservedCapabilityStateRef),
		RollbackPathRef:            applyResult.RollbackPathRef,
		ObservedRollbackPathRef:    normalizeOneDisplaySafeRef(input.ObservedRollbackPathRef),
		ReadbackEvidenceRefs:       normalizeDisplaySafeRefs(input.ReadbackEvidenceRefs),
		FailureClass:               FailureNone,
		Boundaries:                 hostOwnedCapabilityApplyReadbackBoundaries(applyResult.Boundaries),
		NextHostAction:             "provide_capability_apply_readback",
		RunnerEffect:               "none",
		PromptEffect:               "none",
		RuntimeEffect:              "none",
		RawOutputLoaded:            input.RawOutputLoaded || applyResult.RawOutputLoaded,
	}
	if hostOwnedCapabilityApplyReadbackUnsafe(input, applyResult) {
		result = hostOwnedCapabilityApplyReadbackBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if !applyResult.ReadyForCapabilityReadback {
		result = hostOwnedCapabilityApplyReadbackBlock(result, firstFailureClass(applyResult.FailureClass, FailureEvidenceMissing), "capability_apply_result_not_ready", "host:capability_apply_result", firstNextHostAction(applyResult.NextHostAction, "review_capability_apply_result"))
	}
	if result.CapabilityReadbackRef == "" {
		result = hostOwnedCapabilityApplyReadbackBlock(result, FailureEvidenceMissing, "capability_readback_ref_missing", "host:capability_readback_ref", "provide_capability_readback_ref")
	} else if result.ExpectedReadbackRef != "" && result.CapabilityReadbackRef != result.ExpectedReadbackRef {
		result = hostOwnedCapabilityApplyReadbackBlock(result, FailureVerificationFailed, "capability_readback_ref_mismatch", "host:capability_readback_ref", "review_capability_apply_readback")
	}
	if result.ObservedCapabilityRef == "" {
		result = hostOwnedCapabilityApplyReadbackBlock(result, FailureEvidenceMissing, "observed_capability_ref_missing", "host:observed_capability_ref", "provide_capability_apply_readback")
	} else if result.AppliedCapabilityRef != "" && result.ObservedCapabilityRef != result.AppliedCapabilityRef {
		result = hostOwnedCapabilityApplyReadbackBlock(result, FailureVerificationFailed, "observed_capability_ref_mismatch", "host:observed_capability_ref", "review_capability_apply_readback")
	}
	if result.ObservedCapabilityStateRef == "" {
		result = hostOwnedCapabilityApplyReadbackBlock(result, FailureEvidenceMissing, "observed_capability_state_ref_missing", "host:observed_capability_state_ref", "provide_capability_apply_readback")
	} else if result.AppliedCapabilityStateRef != "" && result.ObservedCapabilityStateRef != result.AppliedCapabilityStateRef {
		result = hostOwnedCapabilityApplyReadbackBlock(result, FailureVerificationFailed, "observed_capability_state_ref_mismatch", "host:observed_capability_state_ref", "review_capability_apply_readback")
	}
	if result.ObservedRollbackPathRef == "" {
		result = hostOwnedCapabilityApplyReadbackBlock(result, FailureEvidenceMissing, "capability_rollback_path_readback_missing", "host:observed_capability_rollback_path_ref", "provide_capability_rollback_readback")
	} else if result.RollbackPathRef != "" && result.ObservedRollbackPathRef != result.RollbackPathRef {
		result = hostOwnedCapabilityApplyReadbackBlock(result, FailureVerificationFailed, "capability_rollback_path_readback_mismatch", "host:observed_capability_rollback_path_ref", "review_capability_rollback_readback")
	}
	if len(result.MissingInputs) == 0 && len(result.BlockedReasons) == 0 {
		result.Status = HostActionRecorded
		result.CapabilityReadbackBound = true
		result.RollbackPathVerified = true
		result.ReadyForRuntimeLoopContinuation = true
		result.NextHostAction = "continue_objective_runtime_loop"
		result.Boundaries = AppendBoundaries(result.Boundaries, "capability_apply_readback_bound", "capability_rollback_path_verified", "ready_for_runtime_loop_continuation")
	}
	return result.Normalize()
}

func CloneHostOwnedCapabilityApplyDescriptor(in HostOwnedCapabilityApplyDescriptor) HostOwnedCapabilityApplyDescriptor {
	out := in
	out.SupportedActions = cloneCapabilityApplyActions(in.SupportedActions)
	out.PolicyRefs = cloneDisplaySafeRefs(in.PolicyRefs)
	out.RequiredApprovalRefs = cloneDisplaySafeRefs(in.RequiredApprovalRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (d HostOwnedCapabilityApplyDescriptor) Clone() HostOwnedCapabilityApplyDescriptor {
	return CloneHostOwnedCapabilityApplyDescriptor(d)
}

func (d HostOwnedCapabilityApplyDescriptor) Normalize() HostOwnedCapabilityApplyDescriptor {
	out := CloneHostOwnedCapabilityApplyDescriptor(d)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.Mode = normalizeControlToken(out.Mode)
	if out.Mode == "" {
		out.Mode = "host_owned_capability_apply_descriptor"
	}
	out.CapabilityDescriptorRef = normalizeOneDisplaySafeRef(out.CapabilityDescriptorRef)
	out.CapabilityAdapterRef = normalizeOneDisplaySafeRef(out.CapabilityAdapterRef)
	out.OwnerRef = normalizeOneDisplaySafeRef(out.OwnerRef)
	out.SupportedActions = normalizeCapabilityApplyActions(out.SupportedActions)
	out.SupportedActions = capabilityApplyActionsFromSupport(out.SupportedActions, out.SupportsInstall, out.SupportsEnable, out.SupportsSkillApply, out.SupportsAuthorize, out.SupportsReload, out.SupportsRollback)
	out.SupportsInstall = capabilityApplyActionContains(out.SupportedActions, CapabilityApplyInstall)
	out.SupportsEnable = capabilityApplyActionContains(out.SupportedActions, CapabilityApplyEnable)
	out.SupportsSkillApply = capabilityApplyActionContains(out.SupportedActions, CapabilityApplySkillApply)
	out.SupportsAuthorize = capabilityApplyActionContains(out.SupportedActions, CapabilityApplyAuthorize)
	out.SupportsReload = capabilityApplyActionContains(out.SupportedActions, CapabilityApplyReload)
	out.SupportsRollback = capabilityApplyActionContains(out.SupportedActions, CapabilityApplyRollback)
	out.ProposalContractRef = normalizeOneDisplaySafeRef(out.ProposalContractRef)
	out.CandidateContractRef = normalizeOneDisplaySafeRef(out.CandidateContractRef)
	out.GuardContractRef = normalizeOneDisplaySafeRef(out.GuardContractRef)
	out.DryRunContractRef = normalizeOneDisplaySafeRef(out.DryRunContractRef)
	out.IdempotencyContractRef = normalizeOneDisplaySafeRef(out.IdempotencyContractRef)
	out.ReadbackContractRef = normalizeOneDisplaySafeRef(out.ReadbackContractRef)
	out.RollbackContractRef = normalizeOneDisplaySafeRef(out.RollbackContractRef)
	out.ApprovalPolicyRef = normalizeOneDisplaySafeRef(out.ApprovalPolicyRef)
	out.RedactionPolicyRef = normalizeOneDisplaySafeRef(out.RedactionPolicyRef)
	out.TimeoutPolicyRef = normalizeOneDisplaySafeRef(out.TimeoutPolicyRef)
	out.PolicyRefs = normalizeDisplaySafeRefs(out.PolicyRefs)
	out.RequiredApprovalRefs = normalizeDisplaySafeRefs(out.RequiredApprovalRefs)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.BlockedReasons = normalizeControlTokenList(out.BlockedReasons)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.Boundaries = hostOwnedCapabilityApplyDescriptorBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	out.RuntimeEffect = normalizeControlToken(out.RuntimeEffect)
	if out.RunnerEffect == "" {
		out.RunnerEffect = "none"
	}
	if out.PromptEffect == "" {
		out.PromptEffect = "none"
	}
	if out.RuntimeEffect == "" {
		out.RuntimeEffect = "none"
	}
	if !out.Available {
		out.Status = HostActionNotReady
		out.ReadyForCapabilityApply = false
	}
	if out.RawOutputLoaded || hostOwnedCapabilityApplyDescriptorOutputUnsafe(out) {
		out.RawOutputLoaded = true
		out.Status = HostActionBlocked
		out.ReadyForCapabilityApply = false
		out.FailureClass = firstFailureClass(out.FailureClass, FailureEvidenceWeak)
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		if out.NextHostAction == "" {
			out.NextHostAction = "provide_display_safe_refs"
		}
	}
	out = hostOwnedCapabilityApplyDescriptorResetCoreEffects(out)
	out.ReadyForCapabilityApply = out.ReadyForCapabilityApply &&
		out.Status == HostActionReady &&
		out.Available &&
		out.CapabilityDescriptorRef != "" &&
		out.CapabilityAdapterRef != "" &&
		out.OwnerRef != "" &&
		len(out.SupportedActions) > 0 &&
		out.SupportsRollback &&
		out.ProposalContractRef != "" &&
		out.CandidateContractRef != "" &&
		out.GuardContractRef != "" &&
		out.DryRunContractRef != "" &&
		out.IdempotencyContractRef != "" &&
		out.ReadbackContractRef != "" &&
		out.RollbackContractRef != "" &&
		out.ApprovalPolicyRef != "" &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.RawOutputLoaded
	return out
}

func CloneHostOwnedCapabilityApplyRequest(in HostOwnedCapabilityApplyRequest) HostOwnedCapabilityApplyRequest {
	out := in
	out.Descriptor = in.Descriptor.Clone()
	out.IndependentGate = in.IndependentGate.Clone()
	out.FinalGate = in.FinalGate.Clone()
	out.ApprovalRefs = cloneDisplaySafeRefs(in.ApprovalRefs)
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (r HostOwnedCapabilityApplyRequest) Clone() HostOwnedCapabilityApplyRequest {
	return CloneHostOwnedCapabilityApplyRequest(r)
}

func (r HostOwnedCapabilityApplyRequest) Normalize() HostOwnedCapabilityApplyRequest {
	out := CloneHostOwnedCapabilityApplyRequest(r)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.Mode = normalizeControlToken(out.Mode)
	if out.Mode == "" {
		out.Mode = "host_owned_capability_apply_request"
	}
	out.Descriptor = out.Descriptor.Normalize()
	out.IndependentGate = out.IndependentGate.Normalize()
	out.FinalGate = out.FinalGate.Normalize()
	out.Action = NormalizeCapabilityApplyAction(string(out.Action))
	out.CapabilityApplyRequestRef = normalizeOneDisplaySafeRef(out.CapabilityApplyRequestRef)
	out.CapabilityDescriptorRef = normalizeOneDisplaySafeRef(out.CapabilityDescriptorRef)
	out.CapabilityAdapterRef = normalizeOneDisplaySafeRef(out.CapabilityAdapterRef)
	out.OwnerRef = normalizeOneDisplaySafeRef(out.OwnerRef)
	out.GateRef = normalizeOneDisplaySafeRef(out.GateRef)
	out.PolicyRef = normalizeOneDisplaySafeRef(out.PolicyRef)
	out.CapabilityProposalRef = normalizeOneDisplaySafeRef(out.CapabilityProposalRef)
	out.CapabilityCandidateRef = normalizeOneDisplaySafeRef(out.CapabilityCandidateRef)
	out.CapabilityGuardRef = normalizeOneDisplaySafeRef(out.CapabilityGuardRef)
	out.CapabilityDryRunProofRef = normalizeOneDisplaySafeRef(out.CapabilityDryRunProofRef)
	out.StrategyRef = normalizeOneDisplaySafeRef(out.StrategyRef)
	out.ObjectiveRunRef = normalizeOneDisplaySafeRef(out.ObjectiveRunRef)
	out.TargetCapabilityRef = normalizeOneDisplaySafeRef(out.TargetCapabilityRef)
	out.ExpectedCapabilityRef = normalizeOneDisplaySafeRef(out.ExpectedCapabilityRef)
	out.ExpectedCapabilityStateRef = normalizeOneDisplaySafeRef(out.ExpectedCapabilityStateRef)
	out.HostCapabilityConfirmationRef = normalizeOneDisplaySafeRef(out.HostCapabilityConfirmationRef)
	out.IdempotencyRef = normalizeOneDisplaySafeRef(out.IdempotencyRef)
	out.IdempotencyContractRef = normalizeOneDisplaySafeRef(out.IdempotencyContractRef)
	out.ExpectedCapabilityResultRef = normalizeOneDisplaySafeRef(out.ExpectedCapabilityResultRef)
	out.ExpectedReadbackRef = normalizeOneDisplaySafeRef(out.ExpectedReadbackRef)
	out.RollbackPathRef = normalizeOneDisplaySafeRef(out.RollbackPathRef)
	out.ApprovalRefs = normalizeDisplaySafeRefs(out.ApprovalRefs)
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.BlockedReasons = normalizeControlTokenList(out.BlockedReasons)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.Boundaries = hostOwnedCapabilityApplyRequestBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	out.RuntimeEffect = normalizeControlToken(out.RuntimeEffect)
	if out.RunnerEffect == "" {
		out.RunnerEffect = "none"
	}
	if out.PromptEffect == "" {
		out.PromptEffect = "none"
	}
	if out.RuntimeEffect == "" {
		out.RuntimeEffect = "none"
	}
	if !out.Available {
		out.Status = HostActionNotReady
		out.ReadyForHostCapabilityApply = false
		out.HostCapabilityApplyAuthorized = false
		out.HostMayApplyCapabilityMutation = false
	}
	if out.RawOutputLoaded || hostOwnedCapabilityApplyRequestOutputUnsafe(out) {
		out.RawOutputLoaded = true
		out.Status = HostActionBlocked
		out.ReadyForHostCapabilityApply = false
		out.HostCapabilityApplyAuthorized = false
		out.HostMayApplyCapabilityMutation = false
		out.FailureClass = firstFailureClass(out.FailureClass, FailureEvidenceWeak)
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		if out.NextHostAction == "" {
			out.NextHostAction = "provide_display_safe_refs"
		}
	}
	out = hostOwnedCapabilityApplyRequestResetCoreEffects(out)
	out.ReadyForHostCapabilityApply = out.ReadyForHostCapabilityApply &&
		out.Status == HostActionReady &&
		out.Descriptor.ReadyForCapabilityApply &&
		out.IndependentGate.Kind == ProductionAdapterEffectGateInstallerApply &&
		out.IndependentGate.ReadyForIndependentGatePlan &&
		out.FinalGate.Stage == IntensityGateFinal &&
		out.FinalGate.Allowed &&
		executionIntensityRank(out.FinalGate.ApprovedIntensity) >= executionIntensityRank(IntensityL4DurableLongRun) &&
		out.Action != "" &&
		capabilityApplyActionContains(out.Descriptor.SupportedActions, out.Action) &&
		out.CapabilityApplyRequestRef != "" &&
		out.CapabilityProposalRef != "" &&
		out.CapabilityCandidateRef != "" &&
		out.CapabilityGuardRef != "" &&
		out.CapabilityDryRunProofRef != "" &&
		out.StrategyRef != "" &&
		out.ObjectiveRunRef != "" &&
		out.ExpectedCapabilityRef != "" &&
		out.ExpectedCapabilityStateRef != "" &&
		out.HostCapabilityConfirmationRef != "" &&
		out.IdempotencyRef != "" &&
		out.ExpectedCapabilityResultRef != "" &&
		out.ExpectedReadbackRef != "" &&
		out.RollbackPathRef != "" &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.RawOutputLoaded
	if (out.Action == CapabilityApplyEnable || out.Action == CapabilityApplyReload || out.Action == CapabilityApplyRollback) && out.TargetCapabilityRef == "" {
		out.ReadyForHostCapabilityApply = false
	}
	out.HostCapabilityApplyAuthorized = out.HostCapabilityApplyAuthorized && out.ReadyForHostCapabilityApply
	out.HostMayApplyCapabilityMutation = out.HostMayApplyCapabilityMutation &&
		out.ReadyForHostCapabilityApply &&
		!out.CoreInvocationExecuted &&
		!out.InstallerExecutedByCore &&
		!out.InstallExecutedByCore &&
		!out.EnableExecutedByCore &&
		!out.PackageManagerExecutedByCore &&
		!out.SkillWriteByCore &&
		!out.RuntimeReloadByCore
	return out
}

func CloneHostOwnedCapabilityApplyResult(in HostOwnedCapabilityApplyResult) HostOwnedCapabilityApplyResult {
	out := in
	out.Request = in.Request.Clone()
	out.CapabilityEvidenceRefs = cloneDisplaySafeRefs(in.CapabilityEvidenceRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (r HostOwnedCapabilityApplyResult) Clone() HostOwnedCapabilityApplyResult {
	return CloneHostOwnedCapabilityApplyResult(r)
}

func (r HostOwnedCapabilityApplyResult) Normalize() HostOwnedCapabilityApplyResult {
	out := CloneHostOwnedCapabilityApplyResult(r)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.Mode = normalizeControlToken(out.Mode)
	if out.Mode == "" {
		out.Mode = "host_owned_capability_apply_result"
	}
	out.Request = out.Request.Normalize()
	out.Action = NormalizeCapabilityApplyAction(string(out.Action))
	out.CapabilityApplyResultRef = normalizeOneDisplaySafeRef(out.CapabilityApplyResultRef)
	out.ExpectedCapabilityResultRef = normalizeOneDisplaySafeRef(out.ExpectedCapabilityResultRef)
	out.CapabilityApplyRequestRef = normalizeOneDisplaySafeRef(out.CapabilityApplyRequestRef)
	out.CapabilityAdapterRef = normalizeOneDisplaySafeRef(out.CapabilityAdapterRef)
	out.HostCapabilityRunRef = normalizeOneDisplaySafeRef(out.HostCapabilityRunRef)
	out.ExpectedCapabilityRef = normalizeOneDisplaySafeRef(out.ExpectedCapabilityRef)
	out.ExpectedCapabilityStateRef = normalizeOneDisplaySafeRef(out.ExpectedCapabilityStateRef)
	out.ExpectedReadbackRef = normalizeOneDisplaySafeRef(out.ExpectedReadbackRef)
	out.AppliedCapabilityRef = normalizeOneDisplaySafeRef(out.AppliedCapabilityRef)
	out.AppliedCapabilityStateRef = normalizeOneDisplaySafeRef(out.AppliedCapabilityStateRef)
	out.RollbackPathRef = normalizeOneDisplaySafeRef(out.RollbackPathRef)
	out.FailureRef = normalizeOneDisplaySafeRef(out.FailureRef)
	out.CompensationRef = normalizeOneDisplaySafeRef(out.CompensationRef)
	out.CapabilityEvidenceRefs = normalizeDisplaySafeRefs(out.CapabilityEvidenceRefs)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.BlockedReasons = normalizeControlTokenList(out.BlockedReasons)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.Boundaries = hostOwnedCapabilityApplyResultBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	out.RuntimeEffect = normalizeControlToken(out.RuntimeEffect)
	if out.RunnerEffect == "" {
		out.RunnerEffect = "none"
	}
	if out.PromptEffect == "" {
		out.PromptEffect = "none"
	}
	if out.RuntimeEffect == "" {
		out.RuntimeEffect = "none"
	}
	if !out.Available {
		out.Status = HostActionNotReady
		out.ReadyForCapabilityReadback = false
		out.HostCapabilityApplyRecorded = false
	}
	if out.RawOutputLoaded || hostOwnedCapabilityApplyResultOutputUnsafe(out) {
		out.RawOutputLoaded = true
		out.Status = HostActionBlocked
		out.ReadyForCapabilityReadback = false
		out.HostCapabilityApplyRecorded = false
		out.FailureClass = firstFailureClass(out.FailureClass, FailureEvidenceWeak)
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		if out.NextHostAction == "" {
			out.NextHostAction = "provide_display_safe_refs"
		}
	}
	out = hostOwnedCapabilityApplyResultResetCoreEffects(out)
	out.HostCapabilityApplyRecorded = out.HostCapabilityApplyRecorded &&
		(out.Status == HostActionRecorded || out.Status == HostActionReviewRequired) &&
		out.HostCapabilityApplyReported &&
		out.CapabilityApplyResultRef != "" &&
		out.HostCapabilityRunRef != "" &&
		len(out.MissingInputs) == 0 &&
		!out.RawOutputLoaded
	if out.Status != HostActionRecorded {
		out.ReadyForCapabilityReadback = false
	}
	out.ReadyForCapabilityReadback = out.ReadyForCapabilityReadback &&
		out.Status == HostActionRecorded &&
		out.HostCapabilityApplyRecorded &&
		out.HostCapabilityApplySucceeded &&
		!out.HostCapabilityApplyFailed &&
		out.AppliedCapabilityRef != "" &&
		out.AppliedCapabilityStateRef != "" &&
		!out.CoreInvocationExecuted &&
		!out.InstallerExecutedByCore &&
		!out.InstallExecutedByCore &&
		!out.EnableExecutedByCore &&
		!out.PackageManagerExecutedByCore &&
		!out.SkillWriteByCore &&
		!out.RuntimeReloadByCore
	return out
}

func CloneHostOwnedCapabilityApplyReadback(in HostOwnedCapabilityApplyReadback) HostOwnedCapabilityApplyReadback {
	out := in
	out.Result = in.Result.Clone()
	out.ReadbackEvidenceRefs = cloneDisplaySafeRefs(in.ReadbackEvidenceRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (r HostOwnedCapabilityApplyReadback) Clone() HostOwnedCapabilityApplyReadback {
	return CloneHostOwnedCapabilityApplyReadback(r)
}

func (r HostOwnedCapabilityApplyReadback) Normalize() HostOwnedCapabilityApplyReadback {
	out := CloneHostOwnedCapabilityApplyReadback(r)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.Mode = normalizeControlToken(out.Mode)
	if out.Mode == "" {
		out.Mode = "host_owned_capability_apply_readback"
	}
	out.Result = out.Result.Normalize()
	out.Action = NormalizeCapabilityApplyAction(string(out.Action))
	out.CapabilityReadbackRef = normalizeOneDisplaySafeRef(out.CapabilityReadbackRef)
	out.CapabilityApplyResultRef = normalizeOneDisplaySafeRef(out.CapabilityApplyResultRef)
	out.CapabilityApplyRequestRef = normalizeOneDisplaySafeRef(out.CapabilityApplyRequestRef)
	out.CapabilityAdapterRef = normalizeOneDisplaySafeRef(out.CapabilityAdapterRef)
	out.HostCapabilityRunRef = normalizeOneDisplaySafeRef(out.HostCapabilityRunRef)
	out.ExpectedCapabilityRef = normalizeOneDisplaySafeRef(out.ExpectedCapabilityRef)
	out.ExpectedCapabilityStateRef = normalizeOneDisplaySafeRef(out.ExpectedCapabilityStateRef)
	out.ExpectedReadbackRef = normalizeOneDisplaySafeRef(out.ExpectedReadbackRef)
	out.AppliedCapabilityRef = normalizeOneDisplaySafeRef(out.AppliedCapabilityRef)
	out.AppliedCapabilityStateRef = normalizeOneDisplaySafeRef(out.AppliedCapabilityStateRef)
	out.ObservedCapabilityRef = normalizeOneDisplaySafeRef(out.ObservedCapabilityRef)
	out.ObservedCapabilityStateRef = normalizeOneDisplaySafeRef(out.ObservedCapabilityStateRef)
	out.RollbackPathRef = normalizeOneDisplaySafeRef(out.RollbackPathRef)
	out.ObservedRollbackPathRef = normalizeOneDisplaySafeRef(out.ObservedRollbackPathRef)
	out.ReadbackEvidenceRefs = normalizeDisplaySafeRefs(out.ReadbackEvidenceRefs)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.BlockedReasons = normalizeControlTokenList(out.BlockedReasons)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.Boundaries = hostOwnedCapabilityApplyReadbackBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	out.RuntimeEffect = normalizeControlToken(out.RuntimeEffect)
	if out.RunnerEffect == "" {
		out.RunnerEffect = "none"
	}
	if out.PromptEffect == "" {
		out.PromptEffect = "none"
	}
	if out.RuntimeEffect == "" {
		out.RuntimeEffect = "none"
	}
	if !out.Available {
		out.Status = HostActionNotReady
		out.CapabilityReadbackBound = false
		out.RollbackPathVerified = false
		out.ReadyForRuntimeLoopContinuation = false
	}
	if out.RawOutputLoaded || hostOwnedCapabilityApplyReadbackOutputUnsafe(out) {
		out.RawOutputLoaded = true
		out.Status = HostActionBlocked
		out.CapabilityReadbackBound = false
		out.RollbackPathVerified = false
		out.ReadyForRuntimeLoopContinuation = false
		out.FailureClass = firstFailureClass(out.FailureClass, FailureEvidenceWeak)
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		if out.NextHostAction == "" {
			out.NextHostAction = "provide_display_safe_refs"
		}
	}
	out = hostOwnedCapabilityApplyReadbackResetCoreEffects(out)
	out.CapabilityReadbackBound = out.CapabilityReadbackBound &&
		out.Status == HostActionRecorded &&
		out.CapabilityReadbackRef != "" &&
		out.CapabilityApplyResultRef != "" &&
		out.ObservedCapabilityRef != "" &&
		out.ObservedCapabilityStateRef != "" &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.RawOutputLoaded
	out.RollbackPathVerified = out.RollbackPathVerified &&
		out.CapabilityReadbackBound &&
		out.RollbackPathRef != "" &&
		out.ObservedRollbackPathRef == out.RollbackPathRef
	out.ReadyForRuntimeLoopContinuation = out.ReadyForRuntimeLoopContinuation &&
		out.CapabilityReadbackBound &&
		out.RollbackPathVerified &&
		!out.CoreInvocationExecuted &&
		!out.InstallerExecutedByCore &&
		!out.InstallExecutedByCore &&
		!out.EnableExecutedByCore &&
		!out.PackageManagerExecutedByCore &&
		!out.SkillWriteByCore &&
		!out.RuntimeReloadByCore
	return out
}

func hostOwnedCapabilityApplyDescriptorBlock(result HostOwnedCapabilityApplyDescriptor, failure FailureClass, reason string, missing MissingInput, next NextHostAction) HostOwnedCapabilityApplyDescriptor {
	result.Status = HostActionBlocked
	result.ReadyForCapabilityApply = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.Boundaries = AppendBoundaries(result.Boundaries, "capability_apply_descriptor_blocked")
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	return result
}

func hostOwnedCapabilityApplyRequestBlock(result HostOwnedCapabilityApplyRequest, failure FailureClass, reason string, missing MissingInput, next NextHostAction) HostOwnedCapabilityApplyRequest {
	result.Status = HostActionBlocked
	result.ReadyForHostCapabilityApply = false
	result.HostCapabilityApplyAuthorized = false
	result.HostMayApplyCapabilityMutation = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.Boundaries = AppendBoundaries(result.Boundaries, "capability_apply_request_blocked")
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	return result
}

func hostOwnedCapabilityApplyResultBlock(result HostOwnedCapabilityApplyResult, failure FailureClass, reason string, missing MissingInput, next NextHostAction) HostOwnedCapabilityApplyResult {
	result.Status = HostActionBlocked
	result.ReadyForCapabilityReadback = false
	result.HostCapabilityApplyRecorded = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.Boundaries = AppendBoundaries(result.Boundaries, "capability_apply_result_blocked")
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	return result
}

func hostOwnedCapabilityApplyReadbackBlock(result HostOwnedCapabilityApplyReadback, failure FailureClass, reason string, missing MissingInput, next NextHostAction) HostOwnedCapabilityApplyReadback {
	result.Status = HostActionBlocked
	result.CapabilityReadbackBound = false
	result.RollbackPathVerified = false
	result.ReadyForRuntimeLoopContinuation = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.Boundaries = AppendBoundaries(result.Boundaries, "capability_apply_readback_blocked")
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	return result
}

func hostOwnedCapabilityApplyDescriptorBoundaries(groups ...[]Boundary) []Boundary {
	return MergeBoundaries(
		[]Boundary{
			"host_owned_capability_apply_descriptor",
			"capability_apply_adapter_contract",
			"host_owned_capability_apply",
			"capability_descriptor_projection_only",
			"display_safe_refs_only",
			"rollback_path_required",
			"no_capability_apply_by_core",
			"no_install_apply_by_core",
			"no_package_manager_execution_by_core",
			"no_skill_write_by_core",
			"no_runtime_reload_by_core",
			"no_runner_dispatch",
		},
		MergeBoundaries(groups...),
	)
}

func hostOwnedCapabilityApplyRequestBoundaries(groups ...[]Boundary) []Boundary {
	return MergeBoundaries(
		[]Boundary{
			"host_owned_capability_apply_request",
			"capability_apply_adapter_contract",
			"host_owned_capability_apply",
			"explicit_host_approval_required",
			"l4_capability_apply_required",
			"capability_install_proposal_not_apply",
			"capability_guard_required",
			"capability_apply_dry_run_required",
			"idempotency_required",
			"readback_required",
			"rollback_path_required",
			"display_safe_refs_only",
			"no_capability_apply_by_core",
			"no_install_apply_by_core",
			"no_package_manager_execution_by_core",
			"no_skill_write_by_core",
			"no_runtime_reload_by_core",
			"no_runner_dispatch",
		},
		MergeBoundaries(groups...),
	)
}

func hostOwnedCapabilityApplyResultBoundaries(groups ...[]Boundary) []Boundary {
	return MergeBoundaries(
		[]Boundary{
			"host_owned_capability_apply_result",
			"capability_apply_adapter_contract",
			"host_owned_capability_apply",
			"host_capability_report_required",
			"capability_install_proposal_not_apply",
			"display_safe_refs_only",
			"no_capability_apply_by_core",
			"no_install_apply_by_core",
			"no_package_manager_execution_by_core",
			"no_skill_write_by_core",
			"no_runtime_reload_by_core",
			"no_runner_dispatch",
		},
		MergeBoundaries(groups...),
	)
}

func hostOwnedCapabilityApplyReadbackBoundaries(groups ...[]Boundary) []Boundary {
	return MergeBoundaries(
		[]Boundary{
			"host_owned_capability_apply_readback",
			"capability_apply_adapter_contract",
			"host_owned_capability_apply",
			"capability_apply_readback_bound",
			"rollback_path_verified",
			"display_safe_refs_only",
			"no_capability_apply_by_core",
			"no_install_apply_by_core",
			"no_package_manager_execution_by_core",
			"no_skill_write_by_core",
			"no_runtime_reload_by_core",
			"no_runner_dispatch",
		},
		MergeBoundaries(groups...),
	)
}

func hostOwnedCapabilityApplyRequestUnsafe(input HostOwnedCapabilityApplyRequestInput, descriptor HostOwnedCapabilityApplyDescriptor, gate ProductionAdapterIndependentEffectGate, finalGate IntensityGateResult) bool {
	return input.RawOutputLoaded ||
		descriptor.RawOutputLoaded ||
		gate.RawOutputLoaded ||
		finalGate.RawOutputLoaded ||
		displaySafeRefRejected(input.CapabilityApplyRequestRef) ||
		displaySafeRefRejected(input.CapabilityProposalRef) ||
		displaySafeRefRejected(input.CapabilityCandidateRef) ||
		displaySafeRefRejected(input.CapabilityGuardRef) ||
		displaySafeRefRejected(input.CapabilityDryRunProofRef) ||
		displaySafeRefRejected(input.StrategyRef) ||
		displaySafeRefRejected(input.ObjectiveRunRef) ||
		displaySafeRefRejected(input.TargetCapabilityRef) ||
		displaySafeRefRejected(input.ExpectedCapabilityRef) ||
		displaySafeRefRejected(input.ExpectedCapabilityStateRef) ||
		displaySafeRefRejected(input.HostCapabilityConfirmationRef) ||
		displaySafeRefRejected(input.IdempotencyRef) ||
		displaySafeRefRejected(input.ExpectedCapabilityResultRef) ||
		displaySafeRefRejected(input.ExpectedReadbackRef) ||
		displaySafeRefRejected(input.RollbackPathRef) ||
		displaySafeRefSliceRejected(input.ApprovalRefs) ||
		evidenceRefRejected(input.EvidenceRefs) ||
		hostOwnedCapabilityApplyDescriptorOutputUnsafe(descriptor) ||
		productionAdapterIndependentEffectGateOutputUnsafe(gate) ||
		capabilityApplyFinalGateOutputUnsafe(finalGate)
}

func hostOwnedCapabilityApplyResultUnsafe(input HostOwnedCapabilityApplyResultInput, request HostOwnedCapabilityApplyRequest) bool {
	return input.RawOutputLoaded ||
		request.RawOutputLoaded ||
		displaySafeRefRejected(input.CapabilityApplyResultRef) ||
		displaySafeRefRejected(input.HostCapabilityRunRef) ||
		displaySafeRefRejected(input.AppliedCapabilityRef) ||
		displaySafeRefRejected(input.AppliedCapabilityStateRef) ||
		displaySafeRefRejected(input.FailureRef) ||
		displaySafeRefRejected(input.CompensationRef) ||
		displaySafeRefSliceRejected(input.CapabilityEvidenceRefs) ||
		hostOwnedCapabilityApplyRequestOutputUnsafe(request)
}

func hostOwnedCapabilityApplyReadbackUnsafe(input HostOwnedCapabilityApplyReadbackInput, result HostOwnedCapabilityApplyResult) bool {
	return input.RawOutputLoaded ||
		result.RawOutputLoaded ||
		displaySafeRefRejected(input.CapabilityReadbackRef) ||
		displaySafeRefRejected(input.ObservedCapabilityRef) ||
		displaySafeRefRejected(input.ObservedCapabilityStateRef) ||
		displaySafeRefRejected(input.ObservedRollbackPathRef) ||
		displaySafeRefSliceRejected(input.ReadbackEvidenceRefs) ||
		hostOwnedCapabilityApplyResultOutputUnsafe(result)
}

func hostOwnedCapabilityApplyDescriptorOutputUnsafe(input HostOwnedCapabilityApplyDescriptor) bool {
	return displaySafeRefRejected(input.CapabilityDescriptorRef) ||
		displaySafeRefRejected(input.CapabilityAdapterRef) ||
		displaySafeRefRejected(input.OwnerRef) ||
		displaySafeRefRejected(input.ProposalContractRef) ||
		displaySafeRefRejected(input.CandidateContractRef) ||
		displaySafeRefRejected(input.GuardContractRef) ||
		displaySafeRefRejected(input.DryRunContractRef) ||
		displaySafeRefRejected(input.IdempotencyContractRef) ||
		displaySafeRefRejected(input.ReadbackContractRef) ||
		displaySafeRefRejected(input.RollbackContractRef) ||
		displaySafeRefRejected(input.ApprovalPolicyRef) ||
		displaySafeRefRejected(input.RedactionPolicyRef) ||
		displaySafeRefRejected(input.TimeoutPolicyRef) ||
		displaySafeRefSliceRejected(input.PolicyRefs) ||
		displaySafeRefSliceRejected(input.RequiredApprovalRefs) ||
		input.RawOutputLoaded
}

func hostOwnedCapabilityApplyRequestOutputUnsafe(input HostOwnedCapabilityApplyRequest) bool {
	return displaySafeRefRejected(input.CapabilityApplyRequestRef) ||
		displaySafeRefRejected(input.CapabilityDescriptorRef) ||
		displaySafeRefRejected(input.CapabilityAdapterRef) ||
		displaySafeRefRejected(input.OwnerRef) ||
		displaySafeRefRejected(input.GateRef) ||
		displaySafeRefRejected(input.PolicyRef) ||
		displaySafeRefRejected(input.CapabilityProposalRef) ||
		displaySafeRefRejected(input.CapabilityCandidateRef) ||
		displaySafeRefRejected(input.CapabilityGuardRef) ||
		displaySafeRefRejected(input.CapabilityDryRunProofRef) ||
		displaySafeRefRejected(input.StrategyRef) ||
		displaySafeRefRejected(input.ObjectiveRunRef) ||
		displaySafeRefRejected(input.TargetCapabilityRef) ||
		displaySafeRefRejected(input.ExpectedCapabilityRef) ||
		displaySafeRefRejected(input.ExpectedCapabilityStateRef) ||
		displaySafeRefRejected(input.HostCapabilityConfirmationRef) ||
		displaySafeRefRejected(input.IdempotencyRef) ||
		displaySafeRefRejected(input.IdempotencyContractRef) ||
		displaySafeRefRejected(input.ExpectedCapabilityResultRef) ||
		displaySafeRefRejected(input.ExpectedReadbackRef) ||
		displaySafeRefRejected(input.RollbackPathRef) ||
		displaySafeRefSliceRejected(input.ApprovalRefs) ||
		evidenceRefRejected(input.EvidenceRefs) ||
		hostOwnedCapabilityApplyDescriptorOutputUnsafe(input.Descriptor) ||
		productionAdapterIndependentEffectGateOutputUnsafe(input.IndependentGate) ||
		capabilityApplyFinalGateOutputUnsafe(input.FinalGate) ||
		input.RawOutputLoaded
}

func hostOwnedCapabilityApplyResultOutputUnsafe(input HostOwnedCapabilityApplyResult) bool {
	return displaySafeRefRejected(input.CapabilityApplyResultRef) ||
		displaySafeRefRejected(input.ExpectedCapabilityResultRef) ||
		displaySafeRefRejected(input.CapabilityApplyRequestRef) ||
		displaySafeRefRejected(input.CapabilityAdapterRef) ||
		displaySafeRefRejected(input.HostCapabilityRunRef) ||
		displaySafeRefRejected(input.ExpectedCapabilityRef) ||
		displaySafeRefRejected(input.ExpectedCapabilityStateRef) ||
		displaySafeRefRejected(input.ExpectedReadbackRef) ||
		displaySafeRefRejected(input.AppliedCapabilityRef) ||
		displaySafeRefRejected(input.AppliedCapabilityStateRef) ||
		displaySafeRefRejected(input.RollbackPathRef) ||
		displaySafeRefRejected(input.FailureRef) ||
		displaySafeRefRejected(input.CompensationRef) ||
		displaySafeRefSliceRejected(input.CapabilityEvidenceRefs) ||
		hostOwnedCapabilityApplyRequestOutputUnsafe(input.Request) ||
		input.RawOutputLoaded
}

func hostOwnedCapabilityApplyReadbackOutputUnsafe(input HostOwnedCapabilityApplyReadback) bool {
	return displaySafeRefRejected(input.CapabilityReadbackRef) ||
		displaySafeRefRejected(input.CapabilityApplyResultRef) ||
		displaySafeRefRejected(input.CapabilityApplyRequestRef) ||
		displaySafeRefRejected(input.CapabilityAdapterRef) ||
		displaySafeRefRejected(input.HostCapabilityRunRef) ||
		displaySafeRefRejected(input.ExpectedCapabilityRef) ||
		displaySafeRefRejected(input.ExpectedCapabilityStateRef) ||
		displaySafeRefRejected(input.ExpectedReadbackRef) ||
		displaySafeRefRejected(input.AppliedCapabilityRef) ||
		displaySafeRefRejected(input.AppliedCapabilityStateRef) ||
		displaySafeRefRejected(input.ObservedCapabilityRef) ||
		displaySafeRefRejected(input.ObservedCapabilityStateRef) ||
		displaySafeRefRejected(input.RollbackPathRef) ||
		displaySafeRefRejected(input.ObservedRollbackPathRef) ||
		displaySafeRefSliceRejected(input.ReadbackEvidenceRefs) ||
		hostOwnedCapabilityApplyResultOutputUnsafe(input.Result) ||
		input.RawOutputLoaded
}

func capabilityApplyFinalGateOutputUnsafe(input IntensityGateResult) bool {
	return input.RawOutputLoaded ||
		displaySafeRefRejected(input.StrategyRef) ||
		displaySafeRefSliceRejected(input.ApprovalRefs) ||
		displaySafeRefSliceRejected(input.PolicyRefs) ||
		evidenceRefRejected(input.EvidenceRefs)
}

func hostOwnedCapabilityApplyResultMarkAction(result HostOwnedCapabilityApplyResult) HostOwnedCapabilityApplyResult {
	switch result.Action {
	case CapabilityApplyInstall:
		result.HostCapabilityInstalled = true
	case CapabilityApplyEnable:
		result.HostCapabilityEnabled = true
	case CapabilityApplySkillApply:
		result.HostSkillApplied = true
	case CapabilityApplyAuthorize:
		result.HostCapabilityAuthorized = true
	case CapabilityApplyReload:
		result.HostRuntimeReloaded = true
	case CapabilityApplyRollback:
		result.HostCapabilityRolledBack = true
	}
	return result
}

func unavailableHostOwnedCapabilityApplyRequest() HostOwnedCapabilityApplyRequest {
	return HostOwnedCapabilityApplyRequest{
		ContractVersion: ContractVersion,
		Projected:       true,
		Status:          HostActionNotReady,
		FailureClass:    FailureHostAdapterMissing,
		MissingInputs:   []MissingInput{"host:capability_apply_descriptor"},
		BlockedReasons:  []string{"capability_apply_descriptor_missing"},
		Boundaries:      hostOwnedCapabilityApplyRequestBoundaries([]Boundary{"capability_apply_descriptor_missing"}),
		NextHostAction:  "provide_capability_apply_descriptor",
		RunnerEffect:    "none",
		PromptEffect:    "none",
		RuntimeEffect:   "none",
	}
}

func unavailableHostOwnedCapabilityApplyResult() HostOwnedCapabilityApplyResult {
	return HostOwnedCapabilityApplyResult{
		ContractVersion: ContractVersion,
		Projected:       true,
		Status:          HostActionNotReady,
		FailureClass:    FailureEvidenceMissing,
		MissingInputs:   []MissingInput{"host:capability_apply_request"},
		BlockedReasons:  []string{"capability_apply_request_missing"},
		Boundaries:      hostOwnedCapabilityApplyResultBoundaries([]Boundary{"capability_apply_request_missing"}),
		NextHostAction:  "provide_capability_apply_request",
		RunnerEffect:    "none",
		PromptEffect:    "none",
		RuntimeEffect:   "none",
	}
}

func unavailableHostOwnedCapabilityApplyReadback() HostOwnedCapabilityApplyReadback {
	return HostOwnedCapabilityApplyReadback{
		ContractVersion: ContractVersion,
		Projected:       true,
		Status:          HostActionNotReady,
		FailureClass:    FailureEvidenceMissing,
		MissingInputs:   []MissingInput{"host:capability_apply_result"},
		BlockedReasons:  []string{"capability_apply_result_missing"},
		Boundaries:      hostOwnedCapabilityApplyReadbackBoundaries([]Boundary{"capability_apply_result_missing"}),
		NextHostAction:  "provide_capability_apply_result",
		RunnerEffect:    "none",
		PromptEffect:    "none",
		RuntimeEffect:   "none",
	}
}

func hostOwnedCapabilityApplyDescriptorEmpty(descriptor HostOwnedCapabilityApplyDescriptor) bool {
	return !descriptor.Projected &&
		!descriptor.Available &&
		descriptor.CapabilityDescriptorRef == "" &&
		descriptor.CapabilityAdapterRef == "" &&
		descriptor.OwnerRef == "" &&
		len(descriptor.SupportedActions) == 0 &&
		descriptor.ProposalContractRef == ""
}

func hostOwnedCapabilityApplyRequestEmpty(request HostOwnedCapabilityApplyRequest) bool {
	return !request.Projected &&
		!request.Available &&
		request.CapabilityApplyRequestRef == "" &&
		request.CapabilityAdapterRef == "" &&
		request.CapabilityProposalRef == ""
}

func hostOwnedCapabilityApplyResultEmpty(result HostOwnedCapabilityApplyResult) bool {
	return !result.Projected &&
		!result.Available &&
		result.CapabilityApplyResultRef == "" &&
		result.CapabilityApplyRequestRef == "" &&
		result.HostCapabilityRunRef == ""
}

func cloneCapabilityApplyActions(in []CapabilityApplyAction) []CapabilityApplyAction {
	return append([]CapabilityApplyAction(nil), in...)
}

func normalizeCapabilityApplyActions(in []CapabilityApplyAction) []CapabilityApplyAction {
	out := make([]CapabilityApplyAction, 0, len(in))
	seen := map[CapabilityApplyAction]struct{}{}
	for _, value := range in {
		action := NormalizeCapabilityApplyAction(string(value))
		if action == "" {
			continue
		}
		if _, ok := seen[action]; ok {
			continue
		}
		seen[action] = struct{}{}
		out = append(out, action)
	}
	return out
}

func capabilityApplyActionsFromSupport(in []CapabilityApplyAction, supportsInstall, supportsEnable, supportsSkillApply, supportsAuthorize, supportsReload, supportsRollback bool) []CapabilityApplyAction {
	out := normalizeCapabilityApplyActions(in)
	if supportsInstall {
		out = appendCapabilityApplyActionIfMissing(out, CapabilityApplyInstall)
	}
	if supportsEnable {
		out = appendCapabilityApplyActionIfMissing(out, CapabilityApplyEnable)
	}
	if supportsSkillApply {
		out = appendCapabilityApplyActionIfMissing(out, CapabilityApplySkillApply)
	}
	if supportsAuthorize {
		out = appendCapabilityApplyActionIfMissing(out, CapabilityApplyAuthorize)
	}
	if supportsReload {
		out = appendCapabilityApplyActionIfMissing(out, CapabilityApplyReload)
	}
	if supportsRollback {
		out = appendCapabilityApplyActionIfMissing(out, CapabilityApplyRollback)
	}
	return out
}

func appendCapabilityApplyActionIfMissing(in []CapabilityApplyAction, action CapabilityApplyAction) []CapabilityApplyAction {
	action = NormalizeCapabilityApplyAction(string(action))
	if action == "" || capabilityApplyActionContains(in, action) {
		return in
	}
	return append(in, action)
}

func capabilityApplyActionContains(values []CapabilityApplyAction, needle CapabilityApplyAction) bool {
	normalized := NormalizeCapabilityApplyAction(string(needle))
	if normalized == "" {
		return false
	}
	for _, value := range normalizeCapabilityApplyActions(values) {
		if value == normalized {
			return true
		}
	}
	return false
}

func capabilityApplyDisplaySafeRefContains(values []DisplaySafeRef, needle DisplaySafeRef) bool {
	needle = normalizeOneDisplaySafeRef(needle)
	for _, value := range normalizeDisplaySafeRefs(values) {
		if value == needle {
			return true
		}
	}
	return false
}

func hostOwnedCapabilityApplyDescriptorResetCoreEffects(out HostOwnedCapabilityApplyDescriptor) HostOwnedCapabilityApplyDescriptor {
	out.CoreInvocationExecuted = false
	out.InstallerExecutedByCore = false
	out.InstallExecutedByCore = false
	out.EnableExecutedByCore = false
	out.PackageManagerExecutedByCore = false
	out.SkillWriteByCore = false
	out.RuntimeReloadByCore = false
	return out
}

func hostOwnedCapabilityApplyRequestResetCoreEffects(out HostOwnedCapabilityApplyRequest) HostOwnedCapabilityApplyRequest {
	out.CoreInvocationExecuted = false
	out.InstallerExecutedByCore = false
	out.InstallExecutedByCore = false
	out.EnableExecutedByCore = false
	out.PackageManagerExecutedByCore = false
	out.SkillWriteByCore = false
	out.RuntimeReloadByCore = false
	return out
}

func hostOwnedCapabilityApplyResultResetCoreEffects(out HostOwnedCapabilityApplyResult) HostOwnedCapabilityApplyResult {
	out.CoreInvocationExecuted = false
	out.InstallerExecutedByCore = false
	out.InstallExecutedByCore = false
	out.EnableExecutedByCore = false
	out.PackageManagerExecutedByCore = false
	out.SkillWriteByCore = false
	out.RuntimeReloadByCore = false
	return out
}

func hostOwnedCapabilityApplyReadbackResetCoreEffects(out HostOwnedCapabilityApplyReadback) HostOwnedCapabilityApplyReadback {
	out.CoreInvocationExecuted = false
	out.InstallerExecutedByCore = false
	out.InstallExecutedByCore = false
	out.EnableExecutedByCore = false
	out.PackageManagerExecutedByCore = false
	out.SkillWriteByCore = false
	out.RuntimeReloadByCore = false
	return out
}
