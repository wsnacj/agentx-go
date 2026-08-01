package controlcontract

type SchedulerApplyAction string

const (
	SchedulerApplyCreate  SchedulerApplyAction = "create"
	SchedulerApplyUpdate  SchedulerApplyAction = "update"
	SchedulerApplyDelete  SchedulerApplyAction = "delete"
	SchedulerApplyDisable SchedulerApplyAction = "disable"
	SchedulerApplyCancel  SchedulerApplyAction = "cancel"
)

func KnownSchedulerApplyActions() []SchedulerApplyAction {
	return []SchedulerApplyAction{
		SchedulerApplyCreate,
		SchedulerApplyUpdate,
		SchedulerApplyDelete,
		SchedulerApplyDisable,
		SchedulerApplyCancel,
	}
}

func NormalizeSchedulerApplyAction(raw string) SchedulerApplyAction {
	switch normalizeEnumToken(raw) {
	case "create", "schedule_create", "scheduler_create":
		return SchedulerApplyCreate
	case "update", "upsert", "schedule_update", "scheduler_update":
		return SchedulerApplyUpdate
	case "delete", "remove", "schedule_delete", "scheduler_delete":
		return SchedulerApplyDelete
	case "disable", "pause", "schedule_disable", "scheduler_disable":
		return SchedulerApplyDisable
	case "cancel", "stop", "schedule_cancel", "scheduler_cancel":
		return SchedulerApplyCancel
	default:
		return ""
	}
}

type HostOwnedSchedulerApplyDescriptor struct {
	ContractVersion         string                 `json:"contract_version,omitempty"`
	Projected               bool                   `json:"projected"`
	Available               bool                   `json:"available"`
	Status                  HostActionStatus       `json:"status,omitempty"`
	Mode                    string                 `json:"mode,omitempty"`
	ReadyForSchedulerApply  bool                   `json:"ready_for_scheduler_apply"`
	SchedulerDescriptorRef  DisplaySafeRef         `json:"scheduler_descriptor_ref,omitempty"`
	SchedulerAdapterRef     DisplaySafeRef         `json:"scheduler_adapter_ref,omitempty"`
	OwnerRef                DisplaySafeRef         `json:"owner_ref,omitempty"`
	SupportsCreate          bool                   `json:"supports_create"`
	SupportsUpdate          bool                   `json:"supports_update"`
	SupportsDelete          bool                   `json:"supports_delete"`
	SupportsDisable         bool                   `json:"supports_disable"`
	SupportsCancel          bool                   `json:"supports_cancel"`
	SupportedActions        []SchedulerApplyAction `json:"supported_actions,omitempty"`
	ScheduleContractRef     DisplaySafeRef         `json:"schedule_contract_ref,omitempty"`
	IdempotencyContractRef  DisplaySafeRef         `json:"idempotency_contract_ref,omitempty"`
	ReadbackContractRef     DisplaySafeRef         `json:"readback_contract_ref,omitempty"`
	CancellationContractRef DisplaySafeRef         `json:"cancellation_contract_ref,omitempty"`
	DeleteContractRef       DisplaySafeRef         `json:"delete_contract_ref,omitempty"`
	DisableContractRef      DisplaySafeRef         `json:"disable_contract_ref,omitempty"`
	ApprovalPolicyRef       DisplaySafeRef         `json:"approval_policy_ref,omitempty"`
	RedactionPolicyRef      DisplaySafeRef         `json:"redaction_policy_ref,omitempty"`
	TimeoutPolicyRef        DisplaySafeRef         `json:"timeout_policy_ref,omitempty"`
	PolicyRefs              []DisplaySafeRef       `json:"policy_refs,omitempty"`
	RequiredApprovalRefs    []DisplaySafeRef       `json:"required_approval_refs,omitempty"`
	MissingInputs           []MissingInput         `json:"missing_inputs,omitempty"`
	BlockedReasons          []string               `json:"blocked_reasons,omitempty"`
	FailureClass            FailureClass           `json:"failure_class,omitempty"`
	Boundaries              []Boundary             `json:"boundaries,omitempty"`
	NextHostAction          NextHostAction         `json:"next_host_action,omitempty"`
	RunnerEffect            string                 `json:"runner_effect,omitempty"`
	PromptEffect            string                 `json:"prompt_effect,omitempty"`
	CoreInvocationExecuted  bool                   `json:"core_invocation_executed"`
	SchedulerMutationByCore bool                   `json:"scheduler_mutation_by_core"`
	CoreScheduleCreated     bool                   `json:"core_schedule_created"`
	AutomationCreatedByCore bool                   `json:"automation_created_by_core"`
	RawOutputLoaded         bool                   `json:"raw_output_loaded"`
}

type HostOwnedSchedulerApplyRequestInput struct {
	Descriptor                   HostOwnedSchedulerApplyDescriptor      `json:"descriptor,omitempty"`
	IndependentGate              ProductionAdapterIndependentEffectGate `json:"independent_gate,omitempty"`
	FinalGate                    IntensityGateResult                    `json:"final_gate,omitempty"`
	Action                       SchedulerApplyAction                   `json:"action,omitempty"`
	SchedulerApplyRequestRef     DisplaySafeRef                         `json:"scheduler_apply_request_ref,omitempty"`
	ScheduleProposalRef          DisplaySafeRef                         `json:"schedule_proposal_ref,omitempty"`
	ScheduleDryRunProofRef       DisplaySafeRef                         `json:"schedule_dry_run_proof_ref,omitempty"`
	StrategyRef                  DisplaySafeRef                         `json:"strategy_ref,omitempty"`
	ObjectiveRunRef              DisplaySafeRef                         `json:"objective_run_ref,omitempty"`
	TargetScheduleRef            DisplaySafeRef                         `json:"target_schedule_ref,omitempty"`
	ExpectedScheduleRef          DisplaySafeRef                         `json:"expected_schedule_ref,omitempty"`
	HostSchedulerConfirmationRef DisplaySafeRef                         `json:"host_scheduler_confirmation_ref,omitempty"`
	IdempotencyRef               DisplaySafeRef                         `json:"idempotency_ref,omitempty"`
	ExpectedSchedulerResultRef   DisplaySafeRef                         `json:"expected_scheduler_result_ref,omitempty"`
	ExpectedLifecycleStateRef    DisplaySafeRef                         `json:"expected_lifecycle_state_ref,omitempty"`
	ExpectedReadbackRef          DisplaySafeRef                         `json:"expected_readback_ref,omitempty"`
	CancelPathRef                DisplaySafeRef                         `json:"cancel_path_ref,omitempty"`
	DeletePathRef                DisplaySafeRef                         `json:"delete_path_ref,omitempty"`
	DisablePathRef               DisplaySafeRef                         `json:"disable_path_ref,omitempty"`
	ApprovalRefs                 []DisplaySafeRef                       `json:"approval_refs,omitempty"`
	EvidenceRefs                 []EvidenceRef                          `json:"evidence_refs,omitempty"`
	Boundaries                   []Boundary                             `json:"boundaries,omitempty"`
	RawOutputLoaded              bool                                   `json:"raw_output_loaded"`
}

type HostOwnedSchedulerApplyRequest struct {
	ContractVersion               string                                 `json:"contract_version,omitempty"`
	Projected                     bool                                   `json:"projected"`
	Available                     bool                                   `json:"available"`
	Status                        HostActionStatus                       `json:"status,omitempty"`
	Mode                          string                                 `json:"mode,omitempty"`
	ReadyForHostSchedulerApply    bool                                   `json:"ready_for_host_scheduler_apply"`
	HostSchedulerApplyAuthorized  bool                                   `json:"host_scheduler_apply_authorized"`
	HostMayApplySchedulerMutation bool                                   `json:"host_may_apply_scheduler_mutation"`
	Descriptor                    HostOwnedSchedulerApplyDescriptor      `json:"descriptor,omitempty"`
	IndependentGate               ProductionAdapterIndependentEffectGate `json:"independent_gate,omitempty"`
	FinalGate                     IntensityGateResult                    `json:"final_gate,omitempty"`
	Action                        SchedulerApplyAction                   `json:"action,omitempty"`
	SchedulerApplyRequestRef      DisplaySafeRef                         `json:"scheduler_apply_request_ref,omitempty"`
	SchedulerDescriptorRef        DisplaySafeRef                         `json:"scheduler_descriptor_ref,omitempty"`
	SchedulerAdapterRef           DisplaySafeRef                         `json:"scheduler_adapter_ref,omitempty"`
	OwnerRef                      DisplaySafeRef                         `json:"owner_ref,omitempty"`
	GateRef                       DisplaySafeRef                         `json:"gate_ref,omitempty"`
	PolicyRef                     DisplaySafeRef                         `json:"policy_ref,omitempty"`
	ScheduleProposalRef           DisplaySafeRef                         `json:"schedule_proposal_ref,omitempty"`
	ScheduleDryRunProofRef        DisplaySafeRef                         `json:"schedule_dry_run_proof_ref,omitempty"`
	StrategyRef                   DisplaySafeRef                         `json:"strategy_ref,omitempty"`
	ObjectiveRunRef               DisplaySafeRef                         `json:"objective_run_ref,omitempty"`
	TargetScheduleRef             DisplaySafeRef                         `json:"target_schedule_ref,omitempty"`
	ExpectedScheduleRef           DisplaySafeRef                         `json:"expected_schedule_ref,omitempty"`
	HostSchedulerConfirmationRef  DisplaySafeRef                         `json:"host_scheduler_confirmation_ref,omitempty"`
	IdempotencyRef                DisplaySafeRef                         `json:"idempotency_ref,omitempty"`
	IdempotencyContractRef        DisplaySafeRef                         `json:"idempotency_contract_ref,omitempty"`
	ExpectedSchedulerResultRef    DisplaySafeRef                         `json:"expected_scheduler_result_ref,omitempty"`
	ExpectedLifecycleStateRef     DisplaySafeRef                         `json:"expected_lifecycle_state_ref,omitempty"`
	ExpectedReadbackRef           DisplaySafeRef                         `json:"expected_readback_ref,omitempty"`
	CancelPathRef                 DisplaySafeRef                         `json:"cancel_path_ref,omitempty"`
	DeletePathRef                 DisplaySafeRef                         `json:"delete_path_ref,omitempty"`
	DisablePathRef                DisplaySafeRef                         `json:"disable_path_ref,omitempty"`
	ApprovalRefs                  []DisplaySafeRef                       `json:"approval_refs,omitempty"`
	EvidenceRefs                  []EvidenceRef                          `json:"evidence_refs,omitempty"`
	MissingInputs                 []MissingInput                         `json:"missing_inputs,omitempty"`
	BlockedReasons                []string                               `json:"blocked_reasons,omitempty"`
	FailureClass                  FailureClass                           `json:"failure_class,omitempty"`
	Boundaries                    []Boundary                             `json:"boundaries,omitempty"`
	NextHostAction                NextHostAction                         `json:"next_host_action,omitempty"`
	RunnerEffect                  string                                 `json:"runner_effect,omitempty"`
	PromptEffect                  string                                 `json:"prompt_effect,omitempty"`
	CoreInvocationExecuted        bool                                   `json:"core_invocation_executed"`
	SchedulerMutationByCore       bool                                   `json:"scheduler_mutation_by_core"`
	CoreScheduleCreated           bool                                   `json:"core_schedule_created"`
	AutomationCreatedByCore       bool                                   `json:"automation_created_by_core"`
	RawOutputLoaded               bool                                   `json:"raw_output_loaded"`
}

type HostOwnedSchedulerApplyResultInput struct {
	Request                     HostOwnedSchedulerApplyRequest `json:"request,omitempty"`
	SchedulerApplyResultRef     DisplaySafeRef                 `json:"scheduler_apply_result_ref,omitempty"`
	HostSchedulerRunRef         DisplaySafeRef                 `json:"host_scheduler_run_ref,omitempty"`
	HostSchedulerApplyReported  bool                           `json:"host_scheduler_apply_reported"`
	HostSchedulerApplySucceeded bool                           `json:"host_scheduler_apply_succeeded"`
	HostSchedulerApplyFailed    bool                           `json:"host_scheduler_apply_failed"`
	AppliedScheduleRef          DisplaySafeRef                 `json:"applied_schedule_ref,omitempty"`
	AppliedLifecycleStateRef    DisplaySafeRef                 `json:"applied_lifecycle_state_ref,omitempty"`
	FailureRef                  DisplaySafeRef                 `json:"failure_ref,omitempty"`
	CompensationRef             DisplaySafeRef                 `json:"compensation_ref,omitempty"`
	SchedulerEvidenceRefs       []DisplaySafeRef               `json:"scheduler_evidence_refs,omitempty"`
	RawOutputLoaded             bool                           `json:"raw_output_loaded"`
}

type HostOwnedSchedulerApplyResult struct {
	ContractVersion             string                         `json:"contract_version,omitempty"`
	Projected                   bool                           `json:"projected"`
	Available                   bool                           `json:"available"`
	Status                      HostActionStatus               `json:"status,omitempty"`
	Mode                        string                         `json:"mode,omitempty"`
	ReadyForSchedulerReadback   bool                           `json:"ready_for_scheduler_readback"`
	HostSchedulerApplyReported  bool                           `json:"host_scheduler_apply_reported"`
	HostSchedulerApplySucceeded bool                           `json:"host_scheduler_apply_succeeded"`
	HostSchedulerApplyFailed    bool                           `json:"host_scheduler_apply_failed"`
	HostSchedulerApplyRecorded  bool                           `json:"host_scheduler_apply_recorded"`
	HostScheduleCreated         bool                           `json:"host_schedule_created"`
	HostScheduleUpdated         bool                           `json:"host_schedule_updated"`
	HostScheduleDeleted         bool                           `json:"host_schedule_deleted"`
	HostScheduleDisabled        bool                           `json:"host_schedule_disabled"`
	HostScheduleCanceled        bool                           `json:"host_schedule_canceled"`
	Request                     HostOwnedSchedulerApplyRequest `json:"request,omitempty"`
	Action                      SchedulerApplyAction           `json:"action,omitempty"`
	SchedulerApplyResultRef     DisplaySafeRef                 `json:"scheduler_apply_result_ref,omitempty"`
	ExpectedSchedulerResultRef  DisplaySafeRef                 `json:"expected_scheduler_result_ref,omitempty"`
	SchedulerApplyRequestRef    DisplaySafeRef                 `json:"scheduler_apply_request_ref,omitempty"`
	SchedulerAdapterRef         DisplaySafeRef                 `json:"scheduler_adapter_ref,omitempty"`
	HostSchedulerRunRef         DisplaySafeRef                 `json:"host_scheduler_run_ref,omitempty"`
	ExpectedScheduleRef         DisplaySafeRef                 `json:"expected_schedule_ref,omitempty"`
	ExpectedLifecycleStateRef   DisplaySafeRef                 `json:"expected_lifecycle_state_ref,omitempty"`
	ExpectedReadbackRef         DisplaySafeRef                 `json:"expected_readback_ref,omitempty"`
	AppliedScheduleRef          DisplaySafeRef                 `json:"applied_schedule_ref,omitempty"`
	AppliedLifecycleStateRef    DisplaySafeRef                 `json:"applied_lifecycle_state_ref,omitempty"`
	CancelPathRef               DisplaySafeRef                 `json:"cancel_path_ref,omitempty"`
	DeletePathRef               DisplaySafeRef                 `json:"delete_path_ref,omitempty"`
	DisablePathRef              DisplaySafeRef                 `json:"disable_path_ref,omitempty"`
	FailureRef                  DisplaySafeRef                 `json:"failure_ref,omitempty"`
	CompensationRef             DisplaySafeRef                 `json:"compensation_ref,omitempty"`
	SchedulerEvidenceRefs       []DisplaySafeRef               `json:"scheduler_evidence_refs,omitempty"`
	MissingInputs               []MissingInput                 `json:"missing_inputs,omitempty"`
	BlockedReasons              []string                       `json:"blocked_reasons,omitempty"`
	FailureClass                FailureClass                   `json:"failure_class,omitempty"`
	Boundaries                  []Boundary                     `json:"boundaries,omitempty"`
	NextHostAction              NextHostAction                 `json:"next_host_action,omitempty"`
	RunnerEffect                string                         `json:"runner_effect,omitempty"`
	PromptEffect                string                         `json:"prompt_effect,omitempty"`
	CoreInvocationExecuted      bool                           `json:"core_invocation_executed"`
	SchedulerMutationByCore     bool                           `json:"scheduler_mutation_by_core"`
	CoreScheduleCreated         bool                           `json:"core_schedule_created"`
	AutomationCreatedByCore     bool                           `json:"automation_created_by_core"`
	RawOutputLoaded             bool                           `json:"raw_output_loaded"`
}

type HostOwnedSchedulerApplyReadbackInput struct {
	SchedulerReadbackRef      DisplaySafeRef                `json:"scheduler_readback_ref,omitempty"`
	Result                    HostOwnedSchedulerApplyResult `json:"result,omitempty"`
	ObservedScheduleRef       DisplaySafeRef                `json:"observed_schedule_ref,omitempty"`
	ObservedLifecycleStateRef DisplaySafeRef                `json:"observed_lifecycle_state_ref,omitempty"`
	ObservedCancelPathRef     DisplaySafeRef                `json:"observed_cancel_path_ref,omitempty"`
	ObservedDeletePathRef     DisplaySafeRef                `json:"observed_delete_path_ref,omitempty"`
	ObservedDisablePathRef    DisplaySafeRef                `json:"observed_disable_path_ref,omitempty"`
	ReadbackEvidenceRefs      []DisplaySafeRef              `json:"readback_evidence_refs,omitempty"`
	RawOutputLoaded           bool                          `json:"raw_output_loaded"`
}

type HostOwnedSchedulerApplyReadback struct {
	ContractVersion                 string                        `json:"contract_version,omitempty"`
	Projected                       bool                          `json:"projected"`
	Available                       bool                          `json:"available"`
	Status                          HostActionStatus              `json:"status,omitempty"`
	Mode                            string                        `json:"mode,omitempty"`
	SchedulerReadbackBound          bool                          `json:"scheduler_readback_bound"`
	LifecyclePathVerified           bool                          `json:"lifecycle_path_verified"`
	ReadyForRuntimeLoopContinuation bool                          `json:"ready_for_runtime_loop_continuation"`
	Result                          HostOwnedSchedulerApplyResult `json:"result,omitempty"`
	Action                          SchedulerApplyAction          `json:"action,omitempty"`
	SchedulerReadbackRef            DisplaySafeRef                `json:"scheduler_readback_ref,omitempty"`
	SchedulerApplyResultRef         DisplaySafeRef                `json:"scheduler_apply_result_ref,omitempty"`
	SchedulerApplyRequestRef        DisplaySafeRef                `json:"scheduler_apply_request_ref,omitempty"`
	SchedulerAdapterRef             DisplaySafeRef                `json:"scheduler_adapter_ref,omitempty"`
	HostSchedulerRunRef             DisplaySafeRef                `json:"host_scheduler_run_ref,omitempty"`
	ExpectedScheduleRef             DisplaySafeRef                `json:"expected_schedule_ref,omitempty"`
	ExpectedLifecycleStateRef       DisplaySafeRef                `json:"expected_lifecycle_state_ref,omitempty"`
	ExpectedReadbackRef             DisplaySafeRef                `json:"expected_readback_ref,omitempty"`
	AppliedScheduleRef              DisplaySafeRef                `json:"applied_schedule_ref,omitempty"`
	AppliedLifecycleStateRef        DisplaySafeRef                `json:"applied_lifecycle_state_ref,omitempty"`
	ObservedScheduleRef             DisplaySafeRef                `json:"observed_schedule_ref,omitempty"`
	ObservedLifecycleStateRef       DisplaySafeRef                `json:"observed_lifecycle_state_ref,omitempty"`
	CancelPathRef                   DisplaySafeRef                `json:"cancel_path_ref,omitempty"`
	DeletePathRef                   DisplaySafeRef                `json:"delete_path_ref,omitempty"`
	DisablePathRef                  DisplaySafeRef                `json:"disable_path_ref,omitempty"`
	ObservedCancelPathRef           DisplaySafeRef                `json:"observed_cancel_path_ref,omitempty"`
	ObservedDeletePathRef           DisplaySafeRef                `json:"observed_delete_path_ref,omitempty"`
	ObservedDisablePathRef          DisplaySafeRef                `json:"observed_disable_path_ref,omitempty"`
	ReadbackEvidenceRefs            []DisplaySafeRef              `json:"readback_evidence_refs,omitempty"`
	MissingInputs                   []MissingInput                `json:"missing_inputs,omitempty"`
	BlockedReasons                  []string                      `json:"blocked_reasons,omitempty"`
	FailureClass                    FailureClass                  `json:"failure_class,omitempty"`
	Boundaries                      []Boundary                    `json:"boundaries,omitempty"`
	NextHostAction                  NextHostAction                `json:"next_host_action,omitempty"`
	RunnerEffect                    string                        `json:"runner_effect,omitempty"`
	PromptEffect                    string                        `json:"prompt_effect,omitempty"`
	CoreInvocationExecuted          bool                          `json:"core_invocation_executed"`
	SchedulerMutationByCore         bool                          `json:"scheduler_mutation_by_core"`
	CoreScheduleCreated             bool                          `json:"core_schedule_created"`
	AutomationCreatedByCore         bool                          `json:"automation_created_by_core"`
	RawOutputLoaded                 bool                          `json:"raw_output_loaded"`
}

func BuildHostOwnedSchedulerApplyDescriptor(input HostOwnedSchedulerApplyDescriptor) HostOwnedSchedulerApplyDescriptor {
	unsafeInput := hostOwnedSchedulerApplyDescriptorOutputUnsafe(input)
	result := input.Normalize()
	result.Status = HostActionBlocked
	result.ReadyForSchedulerApply = false
	result.FailureClass = firstFailureClass(result.FailureClass, FailureNone)
	result.Boundaries = hostOwnedSchedulerApplyDescriptorBoundaries(result.Boundaries)
	if unsafeInput {
		result = hostOwnedSchedulerApplyDescriptorBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		result.RawOutputLoaded = true
		return result.Normalize()
	}
	if !result.Available {
		result.Status = HostActionNotReady
		result.FailureClass = firstFailureClass(result.FailureClass, FailureConfigMissing)
		result.NextHostAction = firstNextHostAction(result.NextHostAction, "configure_scheduler_apply_adapter")
		return result.Normalize()
	}
	if result.SchedulerDescriptorRef == "" {
		result = hostOwnedSchedulerApplyDescriptorBlock(result, FailureEvidenceMissing, "scheduler_descriptor_ref_missing", "host:scheduler_descriptor_ref", "provide_scheduler_descriptor_ref")
	}
	if result.SchedulerAdapterRef == "" {
		result = hostOwnedSchedulerApplyDescriptorBlock(result, FailureHostAdapterMissing, "scheduler_adapter_ref_missing", "host:scheduler_adapter_ref", "configure_scheduler_adapter")
	}
	if result.OwnerRef == "" {
		result = hostOwnedSchedulerApplyDescriptorBlock(result, FailureConfigMissing, "scheduler_owner_ref_missing", "host:scheduler_owner_ref", "provide_scheduler_owner_ref")
	}
	for _, check := range []struct {
		ok      bool
		reason  string
		missing MissingInput
	}{
		{len(result.SupportedActions) > 0, "scheduler_actions_missing", "host:scheduler_supported_actions"},
		{result.SupportsDelete, "scheduler_delete_path_not_supported", "host:scheduler_delete_path"},
		{result.SupportsDisable, "scheduler_disable_path_not_supported", "host:scheduler_disable_path"},
		{result.SupportsCancel, "scheduler_cancel_path_not_supported", "host:scheduler_cancel_path"},
	} {
		if !check.ok {
			result = hostOwnedSchedulerApplyDescriptorBlock(result, FailureConfigMissing, check.reason, check.missing, "configure_scheduler_lifecycle_controls")
		}
	}
	for _, check := range []struct {
		ref     DisplaySafeRef
		reason  string
		missing MissingInput
		next    NextHostAction
	}{
		{result.ScheduleContractRef, "schedule_contract_ref_missing", "contract:scheduler_schedule", "provide_schedule_contract_ref"},
		{result.IdempotencyContractRef, "scheduler_idempotency_contract_ref_missing", "contract:scheduler_idempotency", "provide_scheduler_idempotency_contract"},
		{result.ReadbackContractRef, "scheduler_readback_contract_ref_missing", "contract:scheduler_readback", "provide_scheduler_readback_contract"},
		{result.CancellationContractRef, "scheduler_cancellation_contract_ref_missing", "contract:scheduler_cancellation", "provide_scheduler_cancellation_contract"},
		{result.DeleteContractRef, "scheduler_delete_contract_ref_missing", "contract:scheduler_delete", "provide_scheduler_delete_contract"},
		{result.DisableContractRef, "scheduler_disable_contract_ref_missing", "contract:scheduler_disable", "provide_scheduler_disable_contract"},
		{result.ApprovalPolicyRef, "scheduler_approval_policy_ref_missing", "policy:scheduler_approval", "provide_scheduler_approval_policy"},
	} {
		if check.ref == "" {
			result = hostOwnedSchedulerApplyDescriptorBlock(result, FailureConfigMissing, check.reason, check.missing, check.next)
		}
	}
	if len(result.MissingInputs) == 0 && len(result.BlockedReasons) == 0 {
		result.Status = HostActionReady
		result.ReadyForSchedulerApply = true
		result.NextHostAction = "host_may_prepare_scheduler_apply_request"
		result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_scheduler_apply_request")
	}
	return result.Normalize()
}

func BuildHostOwnedSchedulerApplyRequest(input HostOwnedSchedulerApplyRequestInput) HostOwnedSchedulerApplyRequest {
	if hostOwnedSchedulerApplyDescriptorEmpty(input.Descriptor) {
		return unavailableHostOwnedSchedulerApplyRequest()
	}
	descriptor := BuildHostOwnedSchedulerApplyDescriptor(input.Descriptor)
	gate := input.IndependentGate.Normalize()
	finalGate := input.FinalGate.Normalize()
	action := NormalizeSchedulerApplyAction(string(input.Action))
	approvalRefs := normalizeDisplaySafeRefs(append(append(cloneDisplaySafeRefs(input.ApprovalRefs), gate.ApprovalRef), finalGate.ApprovalRefs...))
	result := HostOwnedSchedulerApplyRequest{
		ContractVersion:              ContractVersion,
		Projected:                    true,
		Available:                    descriptor.Available,
		Status:                       HostActionBlocked,
		Mode:                         "host_owned_scheduler_apply_request",
		Descriptor:                   descriptor,
		IndependentGate:              gate,
		FinalGate:                    finalGate,
		Action:                       action,
		SchedulerApplyRequestRef:     normalizeOneDisplaySafeRef(input.SchedulerApplyRequestRef),
		SchedulerDescriptorRef:       descriptor.SchedulerDescriptorRef,
		SchedulerAdapterRef:          descriptor.SchedulerAdapterRef,
		OwnerRef:                     descriptor.OwnerRef,
		GateRef:                      gate.GateRef,
		PolicyRef:                    gate.PolicyRef,
		ScheduleProposalRef:          normalizeOneDisplaySafeRef(input.ScheduleProposalRef),
		ScheduleDryRunProofRef:       normalizeOneDisplaySafeRef(input.ScheduleDryRunProofRef),
		StrategyRef:                  firstDisplaySafeRef(input.StrategyRef, finalGate.StrategyRef),
		ObjectiveRunRef:              normalizeOneDisplaySafeRef(input.ObjectiveRunRef),
		TargetScheduleRef:            normalizeOneDisplaySafeRef(input.TargetScheduleRef),
		ExpectedScheduleRef:          normalizeOneDisplaySafeRef(input.ExpectedScheduleRef),
		HostSchedulerConfirmationRef: normalizeOneDisplaySafeRef(input.HostSchedulerConfirmationRef),
		IdempotencyRef:               firstDisplaySafeRef(input.IdempotencyRef, gate.IdempotencyRef),
		IdempotencyContractRef:       descriptor.IdempotencyContractRef,
		ExpectedSchedulerResultRef:   normalizeOneDisplaySafeRef(input.ExpectedSchedulerResultRef),
		ExpectedLifecycleStateRef:    normalizeOneDisplaySafeRef(input.ExpectedLifecycleStateRef),
		ExpectedReadbackRef:          firstDisplaySafeRef(input.ExpectedReadbackRef, gate.ReadbackRef),
		CancelPathRef:                normalizeOneDisplaySafeRef(input.CancelPathRef),
		DeletePathRef:                normalizeOneDisplaySafeRef(input.DeletePathRef),
		DisablePathRef:               normalizeOneDisplaySafeRef(input.DisablePathRef),
		ApprovalRefs:                 approvalRefs,
		EvidenceRefs:                 MergeEvidenceRefs(input.EvidenceRefs, finalGate.EvidenceRefs),
		FailureClass:                 FailureNone,
		Boundaries:                   hostOwnedSchedulerApplyRequestBoundaries(descriptor.Boundaries, gate.Boundaries, finalGate.Boundaries, input.Boundaries),
		NextHostAction:               "prepare_scheduler_apply_request",
		RunnerEffect:                 "none",
		PromptEffect:                 "none",
		RawOutputLoaded:              input.RawOutputLoaded || descriptor.RawOutputLoaded || gate.RawOutputLoaded || finalGate.RawOutputLoaded,
	}
	if hostOwnedSchedulerApplyRequestUnsafe(input, descriptor, gate, finalGate) {
		result = hostOwnedSchedulerApplyRequestBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if !descriptor.ReadyForSchedulerApply {
		result = hostOwnedSchedulerApplyRequestBlock(result, firstFailureClass(descriptor.FailureClass, FailureConfigMissing), "scheduler_descriptor_not_ready", "host:scheduler_descriptor", firstNextHostAction(descriptor.NextHostAction, "review_scheduler_descriptor"))
	}
	if gate.Kind != ProductionAdapterEffectGateSchedulerApply || !gate.ReadyForIndependentGatePlan {
		result = hostOwnedSchedulerApplyRequestBlock(result, firstFailureClass(gate.FailureClass, FailurePolicyBlocked), "scheduler_independent_gate_not_ready", "host:scheduler_apply_independent_gate", firstNextHostAction(gate.NextHostAction, "review_scheduler_apply_independent_gate"))
	}
	if gate.AdapterRef != "" && descriptor.SchedulerAdapterRef != "" && gate.AdapterRef != descriptor.SchedulerAdapterRef {
		result = hostOwnedSchedulerApplyRequestBlock(result, FailureVerificationFailed, "scheduler_gate_adapter_ref_mismatch", "host:scheduler_adapter_ref", "review_scheduler_apply_gate")
	}
	if gate.IdempotencyRef != "" && result.IdempotencyRef != "" && gate.IdempotencyRef != result.IdempotencyRef {
		result = hostOwnedSchedulerApplyRequestBlock(result, FailureVerificationFailed, "scheduler_gate_idempotency_ref_mismatch", "host:scheduler_idempotency_ref", "review_scheduler_apply_gate")
	}
	if gate.ReadbackRef != "" && result.ExpectedReadbackRef != "" && gate.ReadbackRef != result.ExpectedReadbackRef {
		result = hostOwnedSchedulerApplyRequestBlock(result, FailureVerificationFailed, "scheduler_gate_readback_ref_mismatch", "host:scheduler_readback_ref", "review_scheduler_apply_gate")
	}
	if finalGate.Stage != IntensityGateFinal || !finalGate.Allowed {
		result = hostOwnedSchedulerApplyRequestBlock(result, firstFailureClass(finalGate.FailureClass, FailurePolicyBlocked), "final_gate_not_satisfied", "host:execution_intensity_final_gate", firstNextHostAction(finalGate.NextHostAction, "run_strategy_final_gate"))
	} else if executionIntensityRank(finalGate.ApprovedIntensity) < executionIntensityRank(IntensityL4DurableLongRun) {
		result = hostOwnedSchedulerApplyRequestBlock(result, FailurePolicyBlocked, "scheduler_apply_requires_l4", "contract:l4_durable_long_run", "request_l4_scheduler_approval")
	}
	if finalGate.StrategyRef != "" && result.StrategyRef != "" && finalGate.StrategyRef != result.StrategyRef {
		result = hostOwnedSchedulerApplyRequestBlock(result, FailureVerificationFailed, "final_gate_strategy_ref_mismatch", "host:execution_intensity_final_gate", "run_strategy_final_gate")
	}
	if action == "" {
		result = hostOwnedSchedulerApplyRequestBlock(result, FailureInvalidInput, "scheduler_apply_action_missing", "host:scheduler_apply_action", "provide_scheduler_apply_action")
	} else if !schedulerApplyActionContains(descriptor.SupportedActions, action) {
		result = hostOwnedSchedulerApplyRequestBlock(result, FailureUnsupportedOperation, "scheduler_apply_action_not_supported", "host:scheduler_supported_actions", "select_supported_scheduler_action")
	}
	if gate.ApprovalRef != "" &&
		result.HostSchedulerConfirmationRef != gate.ApprovalRef &&
		!schedulerApplyDisplaySafeRefContains(result.ApprovalRefs, gate.ApprovalRef) {
		result = hostOwnedSchedulerApplyRequestBlock(result, FailureApprovalRequired, "scheduler_gate_approval_ref_missing", "host:scheduler_gate_approval_ref", "provide_scheduler_apply_approval")
	}
	for _, check := range []struct {
		ref     DisplaySafeRef
		reason  string
		missing MissingInput
		next    NextHostAction
		failure FailureClass
	}{
		{result.SchedulerApplyRequestRef, "scheduler_apply_request_ref_missing", "host:scheduler_apply_request_ref", "provide_scheduler_apply_request_ref", FailureEvidenceMissing},
		{result.ScheduleProposalRef, "schedule_proposal_ref_missing", "host:schedule_proposal_ref", "provide_schedule_proposal_ref", FailureEvidenceMissing},
		{result.ScheduleDryRunProofRef, "schedule_dry_run_proof_ref_missing", "host:schedule_dry_run_proof_ref", "provide_schedule_dry_run_proof_ref", FailureEvidenceMissing},
		{result.StrategyRef, "scheduler_strategy_ref_missing", "host:strategy_ref", "provide_scheduler_strategy_ref", FailureEvidenceMissing},
		{result.ObjectiveRunRef, "scheduler_objective_run_ref_missing", "host:objective_run_ref", "provide_scheduler_objective_run_ref", FailureEvidenceMissing},
		{result.HostSchedulerConfirmationRef, "host_scheduler_confirmation_ref_missing", "host:scheduler_confirmation_ref", "request_scheduler_apply_confirmation", FailureApprovalRequired},
		{result.IdempotencyRef, "scheduler_idempotency_ref_missing", "host:scheduler_idempotency_ref", "provide_scheduler_idempotency_ref", FailureConfigMissing},
		{result.ExpectedSchedulerResultRef, "expected_scheduler_result_ref_missing", "host:expected_scheduler_result_ref", "provide_expected_scheduler_result_ref", FailureEvidenceMissing},
		{result.ExpectedScheduleRef, "expected_schedule_ref_missing", "host:expected_schedule_ref", "provide_expected_schedule_ref", FailureEvidenceMissing},
		{result.ExpectedLifecycleStateRef, "expected_lifecycle_state_ref_missing", "host:expected_schedule_lifecycle_state_ref", "provide_expected_schedule_lifecycle_state_ref", FailureEvidenceMissing},
		{result.ExpectedReadbackRef, "expected_scheduler_readback_ref_missing", "host:expected_scheduler_readback_ref", "provide_expected_scheduler_readback_ref", FailureEvidenceMissing},
		{result.CancelPathRef, "scheduler_cancel_path_ref_missing", "host:scheduler_cancel_path_ref", "provide_scheduler_cancel_path_ref", FailureConfigMissing},
		{result.DeletePathRef, "scheduler_delete_path_ref_missing", "host:scheduler_delete_path_ref", "provide_scheduler_delete_path_ref", FailureConfigMissing},
		{result.DisablePathRef, "scheduler_disable_path_ref_missing", "host:scheduler_disable_path_ref", "provide_scheduler_disable_path_ref", FailureConfigMissing},
	} {
		if check.ref == "" {
			result = hostOwnedSchedulerApplyRequestBlock(result, check.failure, check.reason, check.missing, check.next)
		}
	}
	if action != SchedulerApplyCreate {
		if result.TargetScheduleRef == "" {
			result = hostOwnedSchedulerApplyRequestBlock(result, FailureEvidenceMissing, "target_schedule_ref_missing", "host:target_schedule_ref", "provide_target_schedule_ref")
		} else if result.ExpectedScheduleRef != "" && result.TargetScheduleRef != result.ExpectedScheduleRef {
			result = hostOwnedSchedulerApplyRequestBlock(result, FailureVerificationFailed, "target_schedule_ref_mismatch", "host:target_schedule_ref", "review_scheduler_apply_request")
		}
	}
	if len(result.MissingInputs) == 0 && len(result.BlockedReasons) == 0 {
		result.Status = HostActionReady
		result.ReadyForHostSchedulerApply = true
		result.HostSchedulerApplyAuthorized = true
		result.HostMayApplySchedulerMutation = true
		result.NextHostAction = "host_may_apply_scheduler_mutation"
		result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_host_scheduler_apply", "host_may_apply_scheduler_mutation", "core_scheduler_apply_not_executed")
	}
	return result.Normalize()
}

func BuildHostOwnedSchedulerApplyResult(input HostOwnedSchedulerApplyResultInput) HostOwnedSchedulerApplyResult {
	if hostOwnedSchedulerApplyRequestEmpty(input.Request) {
		return unavailableHostOwnedSchedulerApplyResult()
	}
	request := input.Request.Normalize()
	result := HostOwnedSchedulerApplyResult{
		ContractVersion:             ContractVersion,
		Projected:                   true,
		Available:                   request.Available,
		Status:                      HostActionBlocked,
		Mode:                        "host_owned_scheduler_apply_result",
		HostSchedulerApplyReported:  input.HostSchedulerApplyReported,
		HostSchedulerApplySucceeded: input.HostSchedulerApplySucceeded,
		HostSchedulerApplyFailed:    input.HostSchedulerApplyFailed,
		Request:                     request,
		Action:                      request.Action,
		SchedulerApplyResultRef:     normalizeOneDisplaySafeRef(input.SchedulerApplyResultRef),
		ExpectedSchedulerResultRef:  request.ExpectedSchedulerResultRef,
		SchedulerApplyRequestRef:    request.SchedulerApplyRequestRef,
		SchedulerAdapterRef:         request.SchedulerAdapterRef,
		HostSchedulerRunRef:         normalizeOneDisplaySafeRef(input.HostSchedulerRunRef),
		ExpectedScheduleRef:         request.ExpectedScheduleRef,
		ExpectedLifecycleStateRef:   request.ExpectedLifecycleStateRef,
		ExpectedReadbackRef:         request.ExpectedReadbackRef,
		AppliedScheduleRef:          normalizeOneDisplaySafeRef(input.AppliedScheduleRef),
		AppliedLifecycleStateRef:    normalizeOneDisplaySafeRef(input.AppliedLifecycleStateRef),
		CancelPathRef:               request.CancelPathRef,
		DeletePathRef:               request.DeletePathRef,
		DisablePathRef:              request.DisablePathRef,
		FailureRef:                  normalizeOneDisplaySafeRef(input.FailureRef),
		CompensationRef:             normalizeOneDisplaySafeRef(input.CompensationRef),
		SchedulerEvidenceRefs:       normalizeDisplaySafeRefs(input.SchedulerEvidenceRefs),
		FailureClass:                FailureNone,
		Boundaries:                  hostOwnedSchedulerApplyResultBoundaries(request.Boundaries),
		NextHostAction:              "provide_scheduler_apply_result",
		RunnerEffect:                "none",
		PromptEffect:                "none",
		RawOutputLoaded:             input.RawOutputLoaded || request.RawOutputLoaded,
	}
	if hostOwnedSchedulerApplyResultUnsafe(input, request) {
		result = hostOwnedSchedulerApplyResultBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if !request.ReadyForHostSchedulerApply || !request.HostMayApplySchedulerMutation {
		result = hostOwnedSchedulerApplyResultBlock(result, firstFailureClass(request.FailureClass, FailureConfigMissing), "scheduler_apply_request_not_ready", "host:scheduler_apply_request", firstNextHostAction(request.NextHostAction, "review_scheduler_apply_request"))
	}
	if !input.HostSchedulerApplyReported {
		result = hostOwnedSchedulerApplyResultBlock(result, FailureEvidenceMissing, "scheduler_apply_not_reported", "host:scheduler_apply_report", "provide_scheduler_apply_report")
	}
	if input.HostSchedulerApplySucceeded && input.HostSchedulerApplyFailed {
		result = hostOwnedSchedulerApplyResultBlock(result, FailureVerificationFailed, "scheduler_apply_status_conflict", "host:scheduler_apply_status", "review_scheduler_apply_result")
	}
	if result.HostSchedulerRunRef == "" {
		result = hostOwnedSchedulerApplyResultBlock(result, FailureEvidenceMissing, "host_scheduler_run_ref_missing", "host:scheduler_run_ref", "provide_scheduler_apply_report")
	}
	if result.SchedulerApplyResultRef == "" {
		result = hostOwnedSchedulerApplyResultBlock(result, FailureEvidenceMissing, "scheduler_apply_result_ref_missing", "host:scheduler_apply_result_ref", "provide_scheduler_apply_result_ref")
	} else if result.ExpectedSchedulerResultRef != "" && result.SchedulerApplyResultRef != result.ExpectedSchedulerResultRef {
		result = hostOwnedSchedulerApplyResultBlock(result, FailureVerificationFailed, "scheduler_apply_result_ref_mismatch", "host:scheduler_apply_result_ref", "review_scheduler_apply_result")
	}
	if len(result.MissingInputs) > 0 || len(result.BlockedReasons) > 0 {
		return result.Normalize()
	}
	if input.HostSchedulerApplyFailed {
		if result.FailureRef == "" {
			result = hostOwnedSchedulerApplyResultBlock(result, FailureEvidenceMissing, "scheduler_apply_failure_ref_missing", "host:scheduler_apply_failure_ref", "provide_scheduler_apply_failure_ref")
			return result.Normalize()
		}
		result.Status = HostActionReviewRequired
		result.HostSchedulerApplyRecorded = true
		result.FailureClass = FailureVerificationFailed
		result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, "scheduler_apply_failed")
		result.NextHostAction = "review_scheduler_apply_failure"
		result.Boundaries = AppendBoundaries(result.Boundaries, "scheduler_apply_failed", "compensation_not_executed")
		return result.Normalize()
	}
	if !input.HostSchedulerApplySucceeded {
		result = hostOwnedSchedulerApplyResultBlock(result, FailureEvidenceMissing, "scheduler_apply_status_missing", "host:scheduler_apply_status", "provide_scheduler_apply_report")
		return result.Normalize()
	}
	if result.AppliedScheduleRef == "" {
		result = hostOwnedSchedulerApplyResultBlock(result, FailureEvidenceMissing, "applied_schedule_ref_missing", "host:applied_schedule_ref", "provide_scheduler_apply_result")
	} else if result.ExpectedScheduleRef != "" && result.AppliedScheduleRef != result.ExpectedScheduleRef {
		result = hostOwnedSchedulerApplyResultBlock(result, FailureVerificationFailed, "scheduler_applied_schedule_ref_mismatch", "host:applied_schedule_ref", "review_scheduler_apply_result")
	}
	if result.AppliedLifecycleStateRef == "" {
		result = hostOwnedSchedulerApplyResultBlock(result, FailureEvidenceMissing, "applied_lifecycle_state_ref_missing", "host:applied_schedule_lifecycle_state_ref", "provide_scheduler_apply_result")
	} else if result.ExpectedLifecycleStateRef != "" && result.AppliedLifecycleStateRef != result.ExpectedLifecycleStateRef {
		result = hostOwnedSchedulerApplyResultBlock(result, FailureVerificationFailed, "scheduler_lifecycle_state_ref_mismatch", "host:applied_schedule_lifecycle_state_ref", "review_scheduler_apply_result")
	}
	if len(result.MissingInputs) == 0 && len(result.BlockedReasons) == 0 {
		result.Status = HostActionRecorded
		result.HostSchedulerApplyRecorded = true
		result.ReadyForSchedulerReadback = true
		result = hostOwnedSchedulerApplyResultMarkAction(result)
		result.NextHostAction = "bind_scheduler_apply_readback"
		result.Boundaries = AppendBoundaries(result.Boundaries, "host_scheduler_apply_recorded", "ready_for_scheduler_readback")
	}
	return result.Normalize()
}

func BuildHostOwnedSchedulerApplyReadback(input HostOwnedSchedulerApplyReadbackInput) HostOwnedSchedulerApplyReadback {
	if hostOwnedSchedulerApplyResultEmpty(input.Result) {
		return unavailableHostOwnedSchedulerApplyReadback()
	}
	applyResult := input.Result.Normalize()
	result := HostOwnedSchedulerApplyReadback{
		ContractVersion:           ContractVersion,
		Projected:                 true,
		Available:                 applyResult.Available,
		Status:                    HostActionBlocked,
		Mode:                      "host_owned_scheduler_apply_readback",
		Result:                    applyResult,
		Action:                    applyResult.Action,
		SchedulerReadbackRef:      normalizeOneDisplaySafeRef(input.SchedulerReadbackRef),
		SchedulerApplyResultRef:   applyResult.SchedulerApplyResultRef,
		SchedulerApplyRequestRef:  applyResult.SchedulerApplyRequestRef,
		SchedulerAdapterRef:       applyResult.SchedulerAdapterRef,
		HostSchedulerRunRef:       applyResult.HostSchedulerRunRef,
		ExpectedScheduleRef:       applyResult.ExpectedScheduleRef,
		ExpectedLifecycleStateRef: applyResult.ExpectedLifecycleStateRef,
		ExpectedReadbackRef:       applyResult.ExpectedReadbackRef,
		AppliedScheduleRef:        applyResult.AppliedScheduleRef,
		AppliedLifecycleStateRef:  applyResult.AppliedLifecycleStateRef,
		ObservedScheduleRef:       normalizeOneDisplaySafeRef(input.ObservedScheduleRef),
		ObservedLifecycleStateRef: normalizeOneDisplaySafeRef(input.ObservedLifecycleStateRef),
		CancelPathRef:             applyResult.CancelPathRef,
		DeletePathRef:             applyResult.DeletePathRef,
		DisablePathRef:            applyResult.DisablePathRef,
		ObservedCancelPathRef:     normalizeOneDisplaySafeRef(input.ObservedCancelPathRef),
		ObservedDeletePathRef:     normalizeOneDisplaySafeRef(input.ObservedDeletePathRef),
		ObservedDisablePathRef:    normalizeOneDisplaySafeRef(input.ObservedDisablePathRef),
		ReadbackEvidenceRefs:      normalizeDisplaySafeRefs(input.ReadbackEvidenceRefs),
		FailureClass:              FailureNone,
		Boundaries:                hostOwnedSchedulerApplyReadbackBoundaries(applyResult.Boundaries),
		NextHostAction:            "provide_scheduler_apply_readback",
		RunnerEffect:              "none",
		PromptEffect:              "none",
		RawOutputLoaded:           input.RawOutputLoaded || applyResult.RawOutputLoaded,
	}
	if hostOwnedSchedulerApplyReadbackUnsafe(input, applyResult) {
		result = hostOwnedSchedulerApplyReadbackBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if !applyResult.ReadyForSchedulerReadback {
		result = hostOwnedSchedulerApplyReadbackBlock(result, firstFailureClass(applyResult.FailureClass, FailureEvidenceMissing), "scheduler_apply_result_not_ready", "host:scheduler_apply_result", firstNextHostAction(applyResult.NextHostAction, "review_scheduler_apply_result"))
	}
	if result.SchedulerReadbackRef == "" {
		result = hostOwnedSchedulerApplyReadbackBlock(result, FailureEvidenceMissing, "scheduler_readback_ref_missing", "host:scheduler_readback_ref", "provide_scheduler_readback_ref")
	} else if result.ExpectedReadbackRef != "" && result.SchedulerReadbackRef != result.ExpectedReadbackRef {
		result = hostOwnedSchedulerApplyReadbackBlock(result, FailureVerificationFailed, "scheduler_readback_ref_mismatch", "host:scheduler_readback_ref", "review_scheduler_apply_readback")
	}
	if result.ObservedScheduleRef == "" {
		result = hostOwnedSchedulerApplyReadbackBlock(result, FailureEvidenceMissing, "observed_schedule_ref_missing", "host:observed_schedule_ref", "provide_scheduler_apply_readback")
	} else if result.AppliedScheduleRef != "" && result.ObservedScheduleRef != result.AppliedScheduleRef {
		result = hostOwnedSchedulerApplyReadbackBlock(result, FailureVerificationFailed, "scheduler_observed_schedule_ref_mismatch", "host:observed_schedule_ref", "review_scheduler_apply_readback")
	}
	if result.ObservedLifecycleStateRef == "" {
		result = hostOwnedSchedulerApplyReadbackBlock(result, FailureEvidenceMissing, "observed_lifecycle_state_ref_missing", "host:observed_schedule_lifecycle_state_ref", "provide_scheduler_apply_readback")
	} else if result.AppliedLifecycleStateRef != "" && result.ObservedLifecycleStateRef != result.AppliedLifecycleStateRef {
		result = hostOwnedSchedulerApplyReadbackBlock(result, FailureVerificationFailed, "scheduler_observed_lifecycle_state_ref_mismatch", "host:observed_schedule_lifecycle_state_ref", "review_scheduler_apply_readback")
	}
	for _, check := range []struct {
		observed DisplaySafeRef
		expected DisplaySafeRef
		reason   string
		missing  MissingInput
	}{
		{result.ObservedCancelPathRef, result.CancelPathRef, "scheduler_cancel_path_readback_mismatch", "host:observed_scheduler_cancel_path_ref"},
		{result.ObservedDeletePathRef, result.DeletePathRef, "scheduler_delete_path_readback_mismatch", "host:observed_scheduler_delete_path_ref"},
		{result.ObservedDisablePathRef, result.DisablePathRef, "scheduler_disable_path_readback_mismatch", "host:observed_scheduler_disable_path_ref"},
	} {
		if check.observed == "" {
			result = hostOwnedSchedulerApplyReadbackBlock(result, FailureEvidenceMissing, check.reason, check.missing, "provide_scheduler_lifecycle_readback")
		} else if check.expected != "" && check.observed != check.expected {
			result = hostOwnedSchedulerApplyReadbackBlock(result, FailureVerificationFailed, check.reason, check.missing, "review_scheduler_lifecycle_readback")
		}
	}
	if len(result.MissingInputs) == 0 && len(result.BlockedReasons) == 0 {
		result.Status = HostActionRecorded
		result.SchedulerReadbackBound = true
		result.LifecyclePathVerified = true
		result.ReadyForRuntimeLoopContinuation = true
		result.NextHostAction = "continue_objective_runtime_loop"
		result.Boundaries = AppendBoundaries(result.Boundaries, "scheduler_apply_readback_bound", "scheduler_lifecycle_path_verified", "ready_for_runtime_loop_continuation")
	}
	return result.Normalize()
}

func CloneHostOwnedSchedulerApplyDescriptor(in HostOwnedSchedulerApplyDescriptor) HostOwnedSchedulerApplyDescriptor {
	out := in
	out.SupportedActions = cloneSchedulerApplyActions(in.SupportedActions)
	out.PolicyRefs = cloneDisplaySafeRefs(in.PolicyRefs)
	out.RequiredApprovalRefs = cloneDisplaySafeRefs(in.RequiredApprovalRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (d HostOwnedSchedulerApplyDescriptor) Clone() HostOwnedSchedulerApplyDescriptor {
	return CloneHostOwnedSchedulerApplyDescriptor(d)
}

func (d HostOwnedSchedulerApplyDescriptor) Normalize() HostOwnedSchedulerApplyDescriptor {
	out := CloneHostOwnedSchedulerApplyDescriptor(d)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.Mode = normalizeControlToken(out.Mode)
	if out.Mode == "" {
		out.Mode = "host_owned_scheduler_apply_descriptor"
	}
	out.SchedulerDescriptorRef = normalizeOneDisplaySafeRef(out.SchedulerDescriptorRef)
	out.SchedulerAdapterRef = normalizeOneDisplaySafeRef(out.SchedulerAdapterRef)
	out.OwnerRef = normalizeOneDisplaySafeRef(out.OwnerRef)
	out.SupportedActions = normalizeSchedulerApplyActions(out.SupportedActions)
	out.SupportedActions = schedulerApplyActionsFromSupport(out.SupportedActions, out.SupportsCreate, out.SupportsUpdate, out.SupportsDelete, out.SupportsDisable, out.SupportsCancel)
	out.SupportsCreate = schedulerApplyActionContains(out.SupportedActions, SchedulerApplyCreate)
	out.SupportsUpdate = schedulerApplyActionContains(out.SupportedActions, SchedulerApplyUpdate)
	out.SupportsDelete = schedulerApplyActionContains(out.SupportedActions, SchedulerApplyDelete)
	out.SupportsDisable = schedulerApplyActionContains(out.SupportedActions, SchedulerApplyDisable)
	out.SupportsCancel = schedulerApplyActionContains(out.SupportedActions, SchedulerApplyCancel)
	out.ScheduleContractRef = normalizeOneDisplaySafeRef(out.ScheduleContractRef)
	out.IdempotencyContractRef = normalizeOneDisplaySafeRef(out.IdempotencyContractRef)
	out.ReadbackContractRef = normalizeOneDisplaySafeRef(out.ReadbackContractRef)
	out.CancellationContractRef = normalizeOneDisplaySafeRef(out.CancellationContractRef)
	out.DeleteContractRef = normalizeOneDisplaySafeRef(out.DeleteContractRef)
	out.DisableContractRef = normalizeOneDisplaySafeRef(out.DisableContractRef)
	out.ApprovalPolicyRef = normalizeOneDisplaySafeRef(out.ApprovalPolicyRef)
	out.RedactionPolicyRef = normalizeOneDisplaySafeRef(out.RedactionPolicyRef)
	out.TimeoutPolicyRef = normalizeOneDisplaySafeRef(out.TimeoutPolicyRef)
	out.PolicyRefs = normalizeDisplaySafeRefs(out.PolicyRefs)
	out.RequiredApprovalRefs = normalizeDisplaySafeRefs(out.RequiredApprovalRefs)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.BlockedReasons = normalizeControlTokenList(out.BlockedReasons)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.Boundaries = hostOwnedSchedulerApplyDescriptorBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	if out.RunnerEffect == "" {
		out.RunnerEffect = "none"
	}
	if out.PromptEffect == "" {
		out.PromptEffect = "none"
	}
	if !out.Available {
		out.Status = HostActionNotReady
		out.ReadyForSchedulerApply = false
	}
	if out.RawOutputLoaded || hostOwnedSchedulerApplyDescriptorOutputUnsafe(out) {
		out.RawOutputLoaded = true
		out.Status = HostActionBlocked
		out.ReadyForSchedulerApply = false
		out.FailureClass = firstFailureClass(out.FailureClass, FailureEvidenceWeak)
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		if out.NextHostAction == "" {
			out.NextHostAction = "provide_display_safe_refs"
		}
	}
	out.CoreInvocationExecuted = false
	out.SchedulerMutationByCore = false
	out.CoreScheduleCreated = false
	out.AutomationCreatedByCore = false
	out.ReadyForSchedulerApply = out.ReadyForSchedulerApply &&
		out.Status == HostActionReady &&
		out.Available &&
		out.SchedulerDescriptorRef != "" &&
		out.SchedulerAdapterRef != "" &&
		out.OwnerRef != "" &&
		len(out.SupportedActions) > 0 &&
		out.SupportsDelete &&
		out.SupportsDisable &&
		out.SupportsCancel &&
		out.ScheduleContractRef != "" &&
		out.IdempotencyContractRef != "" &&
		out.ReadbackContractRef != "" &&
		out.CancellationContractRef != "" &&
		out.DeleteContractRef != "" &&
		out.DisableContractRef != "" &&
		out.ApprovalPolicyRef != "" &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.RawOutputLoaded
	return out
}

func CloneHostOwnedSchedulerApplyRequest(in HostOwnedSchedulerApplyRequest) HostOwnedSchedulerApplyRequest {
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

func (r HostOwnedSchedulerApplyRequest) Clone() HostOwnedSchedulerApplyRequest {
	return CloneHostOwnedSchedulerApplyRequest(r)
}

func (r HostOwnedSchedulerApplyRequest) Normalize() HostOwnedSchedulerApplyRequest {
	out := CloneHostOwnedSchedulerApplyRequest(r)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.Mode = normalizeControlToken(out.Mode)
	if out.Mode == "" {
		out.Mode = "host_owned_scheduler_apply_request"
	}
	out.Descriptor = out.Descriptor.Normalize()
	out.IndependentGate = out.IndependentGate.Normalize()
	out.FinalGate = out.FinalGate.Normalize()
	out.Action = NormalizeSchedulerApplyAction(string(out.Action))
	out.SchedulerApplyRequestRef = normalizeOneDisplaySafeRef(out.SchedulerApplyRequestRef)
	out.SchedulerDescriptorRef = normalizeOneDisplaySafeRef(out.SchedulerDescriptorRef)
	out.SchedulerAdapterRef = normalizeOneDisplaySafeRef(out.SchedulerAdapterRef)
	out.OwnerRef = normalizeOneDisplaySafeRef(out.OwnerRef)
	out.GateRef = normalizeOneDisplaySafeRef(out.GateRef)
	out.PolicyRef = normalizeOneDisplaySafeRef(out.PolicyRef)
	out.ScheduleProposalRef = normalizeOneDisplaySafeRef(out.ScheduleProposalRef)
	out.ScheduleDryRunProofRef = normalizeOneDisplaySafeRef(out.ScheduleDryRunProofRef)
	out.StrategyRef = normalizeOneDisplaySafeRef(out.StrategyRef)
	out.ObjectiveRunRef = normalizeOneDisplaySafeRef(out.ObjectiveRunRef)
	out.TargetScheduleRef = normalizeOneDisplaySafeRef(out.TargetScheduleRef)
	out.ExpectedScheduleRef = normalizeOneDisplaySafeRef(out.ExpectedScheduleRef)
	out.HostSchedulerConfirmationRef = normalizeOneDisplaySafeRef(out.HostSchedulerConfirmationRef)
	out.IdempotencyRef = normalizeOneDisplaySafeRef(out.IdempotencyRef)
	out.IdempotencyContractRef = normalizeOneDisplaySafeRef(out.IdempotencyContractRef)
	out.ExpectedSchedulerResultRef = normalizeOneDisplaySafeRef(out.ExpectedSchedulerResultRef)
	out.ExpectedLifecycleStateRef = normalizeOneDisplaySafeRef(out.ExpectedLifecycleStateRef)
	out.ExpectedReadbackRef = normalizeOneDisplaySafeRef(out.ExpectedReadbackRef)
	out.CancelPathRef = normalizeOneDisplaySafeRef(out.CancelPathRef)
	out.DeletePathRef = normalizeOneDisplaySafeRef(out.DeletePathRef)
	out.DisablePathRef = normalizeOneDisplaySafeRef(out.DisablePathRef)
	out.ApprovalRefs = normalizeDisplaySafeRefs(out.ApprovalRefs)
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.BlockedReasons = normalizeControlTokenList(out.BlockedReasons)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.Boundaries = hostOwnedSchedulerApplyRequestBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	if out.RunnerEffect == "" {
		out.RunnerEffect = "none"
	}
	if out.PromptEffect == "" {
		out.PromptEffect = "none"
	}
	if !out.Available {
		out.Status = HostActionNotReady
		out.ReadyForHostSchedulerApply = false
		out.HostSchedulerApplyAuthorized = false
		out.HostMayApplySchedulerMutation = false
	}
	if out.RawOutputLoaded || hostOwnedSchedulerApplyRequestOutputUnsafe(out) {
		out.RawOutputLoaded = true
		out.Status = HostActionBlocked
		out.ReadyForHostSchedulerApply = false
		out.HostSchedulerApplyAuthorized = false
		out.HostMayApplySchedulerMutation = false
		out.FailureClass = firstFailureClass(out.FailureClass, FailureEvidenceWeak)
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		if out.NextHostAction == "" {
			out.NextHostAction = "provide_display_safe_refs"
		}
	}
	out.CoreInvocationExecuted = false
	out.SchedulerMutationByCore = false
	out.CoreScheduleCreated = false
	out.AutomationCreatedByCore = false
	out.ReadyForHostSchedulerApply = out.ReadyForHostSchedulerApply &&
		out.Status == HostActionReady &&
		out.Descriptor.ReadyForSchedulerApply &&
		out.IndependentGate.Kind == ProductionAdapterEffectGateSchedulerApply &&
		out.IndependentGate.ReadyForIndependentGatePlan &&
		out.FinalGate.Stage == IntensityGateFinal &&
		out.FinalGate.Allowed &&
		executionIntensityRank(out.FinalGate.ApprovedIntensity) >= executionIntensityRank(IntensityL4DurableLongRun) &&
		out.Action != "" &&
		schedulerApplyActionContains(out.Descriptor.SupportedActions, out.Action) &&
		out.SchedulerApplyRequestRef != "" &&
		out.ScheduleProposalRef != "" &&
		out.ScheduleDryRunProofRef != "" &&
		out.StrategyRef != "" &&
		out.ObjectiveRunRef != "" &&
		out.ExpectedScheduleRef != "" &&
		out.HostSchedulerConfirmationRef != "" &&
		out.IdempotencyRef != "" &&
		out.ExpectedSchedulerResultRef != "" &&
		out.ExpectedLifecycleStateRef != "" &&
		out.ExpectedReadbackRef != "" &&
		out.CancelPathRef != "" &&
		out.DeletePathRef != "" &&
		out.DisablePathRef != "" &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.RawOutputLoaded
	if out.Action != SchedulerApplyCreate && out.TargetScheduleRef == "" {
		out.ReadyForHostSchedulerApply = false
	}
	out.HostSchedulerApplyAuthorized = out.HostSchedulerApplyAuthorized && out.ReadyForHostSchedulerApply
	out.HostMayApplySchedulerMutation = out.HostMayApplySchedulerMutation &&
		out.ReadyForHostSchedulerApply &&
		!out.CoreInvocationExecuted &&
		!out.SchedulerMutationByCore &&
		!out.CoreScheduleCreated &&
		!out.AutomationCreatedByCore
	return out
}

func CloneHostOwnedSchedulerApplyResult(in HostOwnedSchedulerApplyResult) HostOwnedSchedulerApplyResult {
	out := in
	out.Request = in.Request.Clone()
	out.SchedulerEvidenceRefs = cloneDisplaySafeRefs(in.SchedulerEvidenceRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (r HostOwnedSchedulerApplyResult) Clone() HostOwnedSchedulerApplyResult {
	return CloneHostOwnedSchedulerApplyResult(r)
}

func (r HostOwnedSchedulerApplyResult) Normalize() HostOwnedSchedulerApplyResult {
	out := CloneHostOwnedSchedulerApplyResult(r)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.Mode = normalizeControlToken(out.Mode)
	if out.Mode == "" {
		out.Mode = "host_owned_scheduler_apply_result"
	}
	out.Request = out.Request.Normalize()
	out.Action = NormalizeSchedulerApplyAction(string(out.Action))
	out.SchedulerApplyResultRef = normalizeOneDisplaySafeRef(out.SchedulerApplyResultRef)
	out.ExpectedSchedulerResultRef = normalizeOneDisplaySafeRef(out.ExpectedSchedulerResultRef)
	out.SchedulerApplyRequestRef = normalizeOneDisplaySafeRef(out.SchedulerApplyRequestRef)
	out.SchedulerAdapterRef = normalizeOneDisplaySafeRef(out.SchedulerAdapterRef)
	out.HostSchedulerRunRef = normalizeOneDisplaySafeRef(out.HostSchedulerRunRef)
	out.ExpectedScheduleRef = normalizeOneDisplaySafeRef(out.ExpectedScheduleRef)
	out.ExpectedLifecycleStateRef = normalizeOneDisplaySafeRef(out.ExpectedLifecycleStateRef)
	out.ExpectedReadbackRef = normalizeOneDisplaySafeRef(out.ExpectedReadbackRef)
	out.AppliedScheduleRef = normalizeOneDisplaySafeRef(out.AppliedScheduleRef)
	out.AppliedLifecycleStateRef = normalizeOneDisplaySafeRef(out.AppliedLifecycleStateRef)
	out.CancelPathRef = normalizeOneDisplaySafeRef(out.CancelPathRef)
	out.DeletePathRef = normalizeOneDisplaySafeRef(out.DeletePathRef)
	out.DisablePathRef = normalizeOneDisplaySafeRef(out.DisablePathRef)
	out.FailureRef = normalizeOneDisplaySafeRef(out.FailureRef)
	out.CompensationRef = normalizeOneDisplaySafeRef(out.CompensationRef)
	out.SchedulerEvidenceRefs = normalizeDisplaySafeRefs(out.SchedulerEvidenceRefs)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.BlockedReasons = normalizeControlTokenList(out.BlockedReasons)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.Boundaries = hostOwnedSchedulerApplyResultBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	if out.RunnerEffect == "" {
		out.RunnerEffect = "none"
	}
	if out.PromptEffect == "" {
		out.PromptEffect = "none"
	}
	if !out.Available {
		out.Status = HostActionNotReady
		out.HostSchedulerApplyRecorded = false
		out.ReadyForSchedulerReadback = false
	}
	if out.RawOutputLoaded || hostOwnedSchedulerApplyResultOutputUnsafe(out) {
		out.RawOutputLoaded = true
		out.Status = HostActionBlocked
		out.HostSchedulerApplyRecorded = false
		out.ReadyForSchedulerReadback = false
		out.FailureClass = firstFailureClass(out.FailureClass, FailureEvidenceWeak)
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		if out.NextHostAction == "" {
			out.NextHostAction = "provide_display_safe_refs"
		}
	}
	out.CoreInvocationExecuted = false
	out.SchedulerMutationByCore = false
	out.CoreScheduleCreated = false
	out.AutomationCreatedByCore = false
	out.HostSchedulerApplyRecorded = out.HostSchedulerApplyRecorded &&
		(out.Status == HostActionRecorded || out.Status == HostActionReviewRequired) &&
		out.HostSchedulerApplyReported &&
		out.SchedulerApplyResultRef != "" &&
		out.HostSchedulerRunRef != "" &&
		!out.RawOutputLoaded
	out.ReadyForSchedulerReadback = out.ReadyForSchedulerReadback &&
		out.Status == HostActionRecorded &&
		out.HostSchedulerApplyRecorded &&
		out.AppliedScheduleRef != "" &&
		out.AppliedLifecycleStateRef != "" &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.RawOutputLoaded
	if out.Status != HostActionRecorded {
		out.HostScheduleCreated = false
		out.HostScheduleUpdated = false
		out.HostScheduleDeleted = false
		out.HostScheduleDisabled = false
		out.HostScheduleCanceled = false
	}
	return out
}

func CloneHostOwnedSchedulerApplyReadback(in HostOwnedSchedulerApplyReadback) HostOwnedSchedulerApplyReadback {
	out := in
	out.Result = in.Result.Clone()
	out.ReadbackEvidenceRefs = cloneDisplaySafeRefs(in.ReadbackEvidenceRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (r HostOwnedSchedulerApplyReadback) Clone() HostOwnedSchedulerApplyReadback {
	return CloneHostOwnedSchedulerApplyReadback(r)
}

func (r HostOwnedSchedulerApplyReadback) Normalize() HostOwnedSchedulerApplyReadback {
	out := CloneHostOwnedSchedulerApplyReadback(r)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.Mode = normalizeControlToken(out.Mode)
	if out.Mode == "" {
		out.Mode = "host_owned_scheduler_apply_readback"
	}
	out.Result = out.Result.Normalize()
	out.Action = NormalizeSchedulerApplyAction(string(out.Action))
	out.SchedulerReadbackRef = normalizeOneDisplaySafeRef(out.SchedulerReadbackRef)
	out.SchedulerApplyResultRef = normalizeOneDisplaySafeRef(out.SchedulerApplyResultRef)
	out.SchedulerApplyRequestRef = normalizeOneDisplaySafeRef(out.SchedulerApplyRequestRef)
	out.SchedulerAdapterRef = normalizeOneDisplaySafeRef(out.SchedulerAdapterRef)
	out.HostSchedulerRunRef = normalizeOneDisplaySafeRef(out.HostSchedulerRunRef)
	out.ExpectedScheduleRef = normalizeOneDisplaySafeRef(out.ExpectedScheduleRef)
	out.ExpectedLifecycleStateRef = normalizeOneDisplaySafeRef(out.ExpectedLifecycleStateRef)
	out.ExpectedReadbackRef = normalizeOneDisplaySafeRef(out.ExpectedReadbackRef)
	out.AppliedScheduleRef = normalizeOneDisplaySafeRef(out.AppliedScheduleRef)
	out.AppliedLifecycleStateRef = normalizeOneDisplaySafeRef(out.AppliedLifecycleStateRef)
	out.ObservedScheduleRef = normalizeOneDisplaySafeRef(out.ObservedScheduleRef)
	out.ObservedLifecycleStateRef = normalizeOneDisplaySafeRef(out.ObservedLifecycleStateRef)
	out.CancelPathRef = normalizeOneDisplaySafeRef(out.CancelPathRef)
	out.DeletePathRef = normalizeOneDisplaySafeRef(out.DeletePathRef)
	out.DisablePathRef = normalizeOneDisplaySafeRef(out.DisablePathRef)
	out.ObservedCancelPathRef = normalizeOneDisplaySafeRef(out.ObservedCancelPathRef)
	out.ObservedDeletePathRef = normalizeOneDisplaySafeRef(out.ObservedDeletePathRef)
	out.ObservedDisablePathRef = normalizeOneDisplaySafeRef(out.ObservedDisablePathRef)
	out.ReadbackEvidenceRefs = normalizeDisplaySafeRefs(out.ReadbackEvidenceRefs)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.BlockedReasons = normalizeControlTokenList(out.BlockedReasons)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.Boundaries = hostOwnedSchedulerApplyReadbackBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	if out.RunnerEffect == "" {
		out.RunnerEffect = "none"
	}
	if out.PromptEffect == "" {
		out.PromptEffect = "none"
	}
	if !out.Available {
		out.Status = HostActionNotReady
		out.SchedulerReadbackBound = false
		out.LifecyclePathVerified = false
		out.ReadyForRuntimeLoopContinuation = false
	}
	if out.RawOutputLoaded || hostOwnedSchedulerApplyReadbackOutputUnsafe(out) {
		out.RawOutputLoaded = true
		out.Status = HostActionBlocked
		out.SchedulerReadbackBound = false
		out.LifecyclePathVerified = false
		out.ReadyForRuntimeLoopContinuation = false
		out.FailureClass = firstFailureClass(out.FailureClass, FailureEvidenceWeak)
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		if out.NextHostAction == "" {
			out.NextHostAction = "provide_display_safe_refs"
		}
	}
	out.CoreInvocationExecuted = false
	out.SchedulerMutationByCore = false
	out.CoreScheduleCreated = false
	out.AutomationCreatedByCore = false
	out.SchedulerReadbackBound = out.SchedulerReadbackBound &&
		out.Status == HostActionRecorded &&
		out.Result.ReadyForSchedulerReadback &&
		out.SchedulerReadbackRef != "" &&
		out.ObservedScheduleRef != "" &&
		out.ObservedLifecycleStateRef != "" &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.RawOutputLoaded
	out.LifecyclePathVerified = out.LifecyclePathVerified && out.SchedulerReadbackBound
	out.ReadyForRuntimeLoopContinuation = out.ReadyForRuntimeLoopContinuation && out.SchedulerReadbackBound
	return out
}

func hostOwnedSchedulerApplyDescriptorBlock(result HostOwnedSchedulerApplyDescriptor, failure FailureClass, reason string, missing MissingInput, next NextHostAction) HostOwnedSchedulerApplyDescriptor {
	result.Status = HostActionBlocked
	result.ReadyForSchedulerApply = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.Boundaries = AppendBoundaries(result.Boundaries, "scheduler_apply_descriptor_blocked")
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	return result
}

func hostOwnedSchedulerApplyRequestBlock(result HostOwnedSchedulerApplyRequest, failure FailureClass, reason string, missing MissingInput, next NextHostAction) HostOwnedSchedulerApplyRequest {
	result.Status = HostActionBlocked
	result.ReadyForHostSchedulerApply = false
	result.HostSchedulerApplyAuthorized = false
	result.HostMayApplySchedulerMutation = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.Boundaries = AppendBoundaries(result.Boundaries, "scheduler_apply_request_blocked")
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	return result
}

func hostOwnedSchedulerApplyResultBlock(result HostOwnedSchedulerApplyResult, failure FailureClass, reason string, missing MissingInput, next NextHostAction) HostOwnedSchedulerApplyResult {
	result.Status = HostActionBlocked
	result.HostSchedulerApplyRecorded = false
	result.ReadyForSchedulerReadback = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.Boundaries = AppendBoundaries(result.Boundaries, "scheduler_apply_result_blocked")
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	return result
}

func hostOwnedSchedulerApplyReadbackBlock(result HostOwnedSchedulerApplyReadback, failure FailureClass, reason string, missing MissingInput, next NextHostAction) HostOwnedSchedulerApplyReadback {
	result.Status = HostActionBlocked
	result.SchedulerReadbackBound = false
	result.LifecyclePathVerified = false
	result.ReadyForRuntimeLoopContinuation = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.Boundaries = AppendBoundaries(result.Boundaries, "scheduler_apply_readback_blocked")
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	return result
}

func hostOwnedSchedulerApplyDescriptorBoundaries(groups ...[]Boundary) []Boundary {
	return MergeBoundaries(
		[]Boundary{
			"host_owned_scheduler_apply_descriptor",
			"scheduler_apply_adapter_contract",
			"host_owned_scheduler_apply",
			"scheduler_descriptor_projection_only",
			"display_safe_refs_only",
			"cancel_delete_disable_path_required",
			"no_scheduler_mutation_by_core",
			"no_schedule_created_by_core",
			"no_automation_created_by_core",
			"no_runner_dispatch",
		},
		MergeBoundaries(groups...),
	)
}

func hostOwnedSchedulerApplyRequestBoundaries(groups ...[]Boundary) []Boundary {
	return MergeBoundaries(
		[]Boundary{
			"host_owned_scheduler_apply_request",
			"scheduler_apply_adapter_contract",
			"host_owned_scheduler_apply",
			"explicit_host_approval_required",
			"l4_scheduler_apply_required",
			"schedule_proposal_not_schedule_created",
			"schedule_dry_run_proof_required",
			"idempotency_required",
			"readback_required",
			"cancel_delete_disable_path_required",
			"display_safe_refs_only",
			"no_scheduler_mutation_by_core",
			"no_schedule_created_by_core",
			"no_automation_created_by_core",
			"no_runner_dispatch",
		},
		MergeBoundaries(groups...),
	)
}

func hostOwnedSchedulerApplyResultBoundaries(groups ...[]Boundary) []Boundary {
	return MergeBoundaries(
		[]Boundary{
			"host_owned_scheduler_apply_result",
			"scheduler_apply_adapter_contract",
			"host_owned_scheduler_apply",
			"host_scheduler_report_required",
			"schedule_proposal_not_schedule_created",
			"display_safe_refs_only",
			"no_scheduler_mutation_by_core",
			"no_schedule_created_by_core",
			"no_automation_created_by_core",
			"no_runner_dispatch",
		},
		MergeBoundaries(groups...),
	)
}

func hostOwnedSchedulerApplyReadbackBoundaries(groups ...[]Boundary) []Boundary {
	return MergeBoundaries(
		[]Boundary{
			"host_owned_scheduler_apply_readback",
			"scheduler_apply_adapter_contract",
			"host_owned_scheduler_apply",
			"scheduler_apply_readback_bound",
			"cancel_delete_disable_path_verified",
			"display_safe_refs_only",
			"no_scheduler_mutation_by_core",
			"no_schedule_created_by_core",
			"no_automation_created_by_core",
			"no_runner_dispatch",
		},
		MergeBoundaries(groups...),
	)
}

func hostOwnedSchedulerApplyRequestUnsafe(input HostOwnedSchedulerApplyRequestInput, descriptor HostOwnedSchedulerApplyDescriptor, gate ProductionAdapterIndependentEffectGate, finalGate IntensityGateResult) bool {
	return input.RawOutputLoaded ||
		descriptor.RawOutputLoaded ||
		gate.RawOutputLoaded ||
		finalGate.RawOutputLoaded ||
		displaySafeRefRejected(input.SchedulerApplyRequestRef) ||
		displaySafeRefRejected(input.ScheduleProposalRef) ||
		displaySafeRefRejected(input.ScheduleDryRunProofRef) ||
		displaySafeRefRejected(input.StrategyRef) ||
		displaySafeRefRejected(input.ObjectiveRunRef) ||
		displaySafeRefRejected(input.TargetScheduleRef) ||
		displaySafeRefRejected(input.ExpectedScheduleRef) ||
		displaySafeRefRejected(input.HostSchedulerConfirmationRef) ||
		displaySafeRefRejected(input.IdempotencyRef) ||
		displaySafeRefRejected(input.ExpectedSchedulerResultRef) ||
		displaySafeRefRejected(input.ExpectedLifecycleStateRef) ||
		displaySafeRefRejected(input.ExpectedReadbackRef) ||
		displaySafeRefRejected(input.CancelPathRef) ||
		displaySafeRefRejected(input.DeletePathRef) ||
		displaySafeRefRejected(input.DisablePathRef) ||
		displaySafeRefSliceRejected(input.ApprovalRefs) ||
		evidenceRefRejected(input.EvidenceRefs) ||
		hostOwnedSchedulerApplyDescriptorOutputUnsafe(descriptor) ||
		productionAdapterIndependentEffectGateOutputUnsafe(gate) ||
		schedulerApplyFinalGateOutputUnsafe(finalGate)
}

func hostOwnedSchedulerApplyResultUnsafe(input HostOwnedSchedulerApplyResultInput, request HostOwnedSchedulerApplyRequest) bool {
	return input.RawOutputLoaded ||
		request.RawOutputLoaded ||
		displaySafeRefRejected(input.SchedulerApplyResultRef) ||
		displaySafeRefRejected(input.HostSchedulerRunRef) ||
		displaySafeRefRejected(input.AppliedScheduleRef) ||
		displaySafeRefRejected(input.AppliedLifecycleStateRef) ||
		displaySafeRefRejected(input.FailureRef) ||
		displaySafeRefRejected(input.CompensationRef) ||
		displaySafeRefSliceRejected(input.SchedulerEvidenceRefs) ||
		hostOwnedSchedulerApplyRequestOutputUnsafe(request)
}

func hostOwnedSchedulerApplyReadbackUnsafe(input HostOwnedSchedulerApplyReadbackInput, result HostOwnedSchedulerApplyResult) bool {
	return input.RawOutputLoaded ||
		result.RawOutputLoaded ||
		displaySafeRefRejected(input.SchedulerReadbackRef) ||
		displaySafeRefRejected(input.ObservedScheduleRef) ||
		displaySafeRefRejected(input.ObservedLifecycleStateRef) ||
		displaySafeRefRejected(input.ObservedCancelPathRef) ||
		displaySafeRefRejected(input.ObservedDeletePathRef) ||
		displaySafeRefRejected(input.ObservedDisablePathRef) ||
		displaySafeRefSliceRejected(input.ReadbackEvidenceRefs) ||
		hostOwnedSchedulerApplyResultOutputUnsafe(result)
}

func hostOwnedSchedulerApplyDescriptorOutputUnsafe(input HostOwnedSchedulerApplyDescriptor) bool {
	return displaySafeRefRejected(input.SchedulerDescriptorRef) ||
		displaySafeRefRejected(input.SchedulerAdapterRef) ||
		displaySafeRefRejected(input.OwnerRef) ||
		displaySafeRefRejected(input.ScheduleContractRef) ||
		displaySafeRefRejected(input.IdempotencyContractRef) ||
		displaySafeRefRejected(input.ReadbackContractRef) ||
		displaySafeRefRejected(input.CancellationContractRef) ||
		displaySafeRefRejected(input.DeleteContractRef) ||
		displaySafeRefRejected(input.DisableContractRef) ||
		displaySafeRefRejected(input.ApprovalPolicyRef) ||
		displaySafeRefRejected(input.RedactionPolicyRef) ||
		displaySafeRefRejected(input.TimeoutPolicyRef) ||
		displaySafeRefSliceRejected(input.PolicyRefs) ||
		displaySafeRefSliceRejected(input.RequiredApprovalRefs) ||
		input.RawOutputLoaded
}

func hostOwnedSchedulerApplyRequestOutputUnsafe(input HostOwnedSchedulerApplyRequest) bool {
	return displaySafeRefRejected(input.SchedulerApplyRequestRef) ||
		displaySafeRefRejected(input.SchedulerDescriptorRef) ||
		displaySafeRefRejected(input.SchedulerAdapterRef) ||
		displaySafeRefRejected(input.OwnerRef) ||
		displaySafeRefRejected(input.GateRef) ||
		displaySafeRefRejected(input.PolicyRef) ||
		displaySafeRefRejected(input.ScheduleProposalRef) ||
		displaySafeRefRejected(input.ScheduleDryRunProofRef) ||
		displaySafeRefRejected(input.StrategyRef) ||
		displaySafeRefRejected(input.ObjectiveRunRef) ||
		displaySafeRefRejected(input.TargetScheduleRef) ||
		displaySafeRefRejected(input.ExpectedScheduleRef) ||
		displaySafeRefRejected(input.HostSchedulerConfirmationRef) ||
		displaySafeRefRejected(input.IdempotencyRef) ||
		displaySafeRefRejected(input.IdempotencyContractRef) ||
		displaySafeRefRejected(input.ExpectedSchedulerResultRef) ||
		displaySafeRefRejected(input.ExpectedLifecycleStateRef) ||
		displaySafeRefRejected(input.ExpectedReadbackRef) ||
		displaySafeRefRejected(input.CancelPathRef) ||
		displaySafeRefRejected(input.DeletePathRef) ||
		displaySafeRefRejected(input.DisablePathRef) ||
		displaySafeRefSliceRejected(input.ApprovalRefs) ||
		evidenceRefRejected(input.EvidenceRefs) ||
		hostOwnedSchedulerApplyDescriptorOutputUnsafe(input.Descriptor) ||
		productionAdapterIndependentEffectGateOutputUnsafe(input.IndependentGate) ||
		schedulerApplyFinalGateOutputUnsafe(input.FinalGate) ||
		input.RawOutputLoaded
}

func hostOwnedSchedulerApplyResultOutputUnsafe(input HostOwnedSchedulerApplyResult) bool {
	return displaySafeRefRejected(input.SchedulerApplyResultRef) ||
		displaySafeRefRejected(input.ExpectedSchedulerResultRef) ||
		displaySafeRefRejected(input.SchedulerApplyRequestRef) ||
		displaySafeRefRejected(input.SchedulerAdapterRef) ||
		displaySafeRefRejected(input.HostSchedulerRunRef) ||
		displaySafeRefRejected(input.ExpectedScheduleRef) ||
		displaySafeRefRejected(input.ExpectedLifecycleStateRef) ||
		displaySafeRefRejected(input.ExpectedReadbackRef) ||
		displaySafeRefRejected(input.AppliedScheduleRef) ||
		displaySafeRefRejected(input.AppliedLifecycleStateRef) ||
		displaySafeRefRejected(input.CancelPathRef) ||
		displaySafeRefRejected(input.DeletePathRef) ||
		displaySafeRefRejected(input.DisablePathRef) ||
		displaySafeRefRejected(input.FailureRef) ||
		displaySafeRefRejected(input.CompensationRef) ||
		displaySafeRefSliceRejected(input.SchedulerEvidenceRefs) ||
		hostOwnedSchedulerApplyRequestOutputUnsafe(input.Request) ||
		input.RawOutputLoaded
}

func hostOwnedSchedulerApplyReadbackOutputUnsafe(input HostOwnedSchedulerApplyReadback) bool {
	return displaySafeRefRejected(input.SchedulerReadbackRef) ||
		displaySafeRefRejected(input.SchedulerApplyResultRef) ||
		displaySafeRefRejected(input.SchedulerApplyRequestRef) ||
		displaySafeRefRejected(input.SchedulerAdapterRef) ||
		displaySafeRefRejected(input.HostSchedulerRunRef) ||
		displaySafeRefRejected(input.ExpectedScheduleRef) ||
		displaySafeRefRejected(input.ExpectedLifecycleStateRef) ||
		displaySafeRefRejected(input.ExpectedReadbackRef) ||
		displaySafeRefRejected(input.AppliedScheduleRef) ||
		displaySafeRefRejected(input.AppliedLifecycleStateRef) ||
		displaySafeRefRejected(input.ObservedScheduleRef) ||
		displaySafeRefRejected(input.ObservedLifecycleStateRef) ||
		displaySafeRefRejected(input.CancelPathRef) ||
		displaySafeRefRejected(input.DeletePathRef) ||
		displaySafeRefRejected(input.DisablePathRef) ||
		displaySafeRefRejected(input.ObservedCancelPathRef) ||
		displaySafeRefRejected(input.ObservedDeletePathRef) ||
		displaySafeRefRejected(input.ObservedDisablePathRef) ||
		displaySafeRefSliceRejected(input.ReadbackEvidenceRefs) ||
		hostOwnedSchedulerApplyResultOutputUnsafe(input.Result) ||
		input.RawOutputLoaded
}

func schedulerApplyFinalGateOutputUnsafe(input IntensityGateResult) bool {
	return input.RawOutputLoaded ||
		displaySafeRefRejected(input.StrategyRef) ||
		displaySafeRefSliceRejected(input.ApprovalRefs) ||
		displaySafeRefSliceRejected(input.PolicyRefs) ||
		displaySafeRefSliceRejected(input.DecisionBasis) ||
		evidenceRefRejected(input.EvidenceRefs)
}

func hostOwnedSchedulerApplyResultMarkAction(result HostOwnedSchedulerApplyResult) HostOwnedSchedulerApplyResult {
	switch result.Action {
	case SchedulerApplyCreate:
		result.HostScheduleCreated = true
	case SchedulerApplyUpdate:
		result.HostScheduleUpdated = true
	case SchedulerApplyDelete:
		result.HostScheduleDeleted = true
	case SchedulerApplyDisable:
		result.HostScheduleDisabled = true
	case SchedulerApplyCancel:
		result.HostScheduleCanceled = true
	}
	return result
}

func unavailableHostOwnedSchedulerApplyRequest() HostOwnedSchedulerApplyRequest {
	return HostOwnedSchedulerApplyRequest{
		ContractVersion: ContractVersion,
		Projected:       true,
		Available:       false,
		Status:          HostActionNotReady,
		Mode:            "host_owned_scheduler_apply_request",
		FailureClass:    FailureConfigMissing,
		MissingInputs:   []MissingInput{"host:scheduler_apply_descriptor"},
		Boundaries:      hostOwnedSchedulerApplyRequestBoundaries([]Boundary{"scheduler_apply_descriptor_missing"}),
		NextHostAction:  "configure_scheduler_apply_adapter",
		RunnerEffect:    "none",
		PromptEffect:    "none",
	}.Normalize()
}

func unavailableHostOwnedSchedulerApplyResult() HostOwnedSchedulerApplyResult {
	return HostOwnedSchedulerApplyResult{
		ContractVersion: ContractVersion,
		Projected:       true,
		Available:       false,
		Status:          HostActionNotReady,
		Mode:            "host_owned_scheduler_apply_result",
		FailureClass:    FailureConfigMissing,
		MissingInputs:   []MissingInput{"host:scheduler_apply_request"},
		Boundaries:      hostOwnedSchedulerApplyResultBoundaries([]Boundary{"scheduler_apply_request_missing"}),
		NextHostAction:  "review_scheduler_apply_request",
		RunnerEffect:    "none",
		PromptEffect:    "none",
	}.Normalize()
}

func unavailableHostOwnedSchedulerApplyReadback() HostOwnedSchedulerApplyReadback {
	return HostOwnedSchedulerApplyReadback{
		ContractVersion: ContractVersion,
		Projected:       true,
		Available:       false,
		Status:          HostActionNotReady,
		Mode:            "host_owned_scheduler_apply_readback",
		FailureClass:    FailureConfigMissing,
		MissingInputs:   []MissingInput{"host:scheduler_apply_result"},
		Boundaries:      hostOwnedSchedulerApplyReadbackBoundaries([]Boundary{"scheduler_apply_result_missing"}),
		NextHostAction:  "review_scheduler_apply_result",
		RunnerEffect:    "none",
		PromptEffect:    "none",
	}.Normalize()
}

func hostOwnedSchedulerApplyDescriptorEmpty(descriptor HostOwnedSchedulerApplyDescriptor) bool {
	return !descriptor.Projected &&
		!descriptor.Available &&
		descriptor.Status == "" &&
		descriptor.Mode == "" &&
		descriptor.SchedulerDescriptorRef == "" &&
		descriptor.SchedulerAdapterRef == "" &&
		len(descriptor.SupportedActions) == 0 &&
		len(descriptor.Boundaries) == 0
}

func hostOwnedSchedulerApplyRequestEmpty(request HostOwnedSchedulerApplyRequest) bool {
	return !request.Projected &&
		!request.Available &&
		request.Status == "" &&
		request.Mode == "" &&
		request.SchedulerApplyRequestRef == "" &&
		request.SchedulerAdapterRef == "" &&
		request.Action == "" &&
		len(request.Boundaries) == 0
}

func hostOwnedSchedulerApplyResultEmpty(result HostOwnedSchedulerApplyResult) bool {
	return !result.Projected &&
		!result.Available &&
		result.Status == "" &&
		result.Mode == "" &&
		result.SchedulerApplyResultRef == "" &&
		result.SchedulerApplyRequestRef == "" &&
		len(result.Boundaries) == 0
}

func cloneSchedulerApplyActions(in []SchedulerApplyAction) []SchedulerApplyAction {
	if len(in) == 0 {
		return nil
	}
	return append([]SchedulerApplyAction(nil), in...)
}

func normalizeSchedulerApplyActions(in []SchedulerApplyAction) []SchedulerApplyAction {
	out := make([]SchedulerApplyAction, 0, len(in))
	seen := map[SchedulerApplyAction]struct{}{}
	for _, value := range in {
		action := NormalizeSchedulerApplyAction(string(value))
		if action == "" {
			continue
		}
		if _, exists := seen[action]; exists {
			continue
		}
		seen[action] = struct{}{}
		out = append(out, action)
	}
	return out
}

func schedulerApplyActionsFromSupport(in []SchedulerApplyAction, supportsCreate, supportsUpdate, supportsDelete, supportsDisable, supportsCancel bool) []SchedulerApplyAction {
	out := normalizeSchedulerApplyActions(in)
	if supportsCreate {
		out = appendSchedulerApplyActionIfMissing(out, SchedulerApplyCreate)
	}
	if supportsUpdate {
		out = appendSchedulerApplyActionIfMissing(out, SchedulerApplyUpdate)
	}
	if supportsDelete {
		out = appendSchedulerApplyActionIfMissing(out, SchedulerApplyDelete)
	}
	if supportsDisable {
		out = appendSchedulerApplyActionIfMissing(out, SchedulerApplyDisable)
	}
	if supportsCancel {
		out = appendSchedulerApplyActionIfMissing(out, SchedulerApplyCancel)
	}
	return out
}

func appendSchedulerApplyActionIfMissing(in []SchedulerApplyAction, action SchedulerApplyAction) []SchedulerApplyAction {
	action = NormalizeSchedulerApplyAction(string(action))
	if action == "" {
		return in
	}
	for _, existing := range in {
		if existing == action {
			return in
		}
	}
	return append(in, action)
}

func schedulerApplyActionContains(values []SchedulerApplyAction, needle SchedulerApplyAction) bool {
	normalized := NormalizeSchedulerApplyAction(string(needle))
	if normalized == "" {
		return false
	}
	for _, value := range normalizeSchedulerApplyActions(values) {
		if value == normalized {
			return true
		}
	}
	return false
}

func schedulerApplyDisplaySafeRefContains(values []DisplaySafeRef, needle DisplaySafeRef) bool {
	needle = normalizeOneDisplaySafeRef(needle)
	if needle == "" {
		return false
	}
	for _, value := range normalizeDisplaySafeRefs(values) {
		if value == needle {
			return true
		}
	}
	return false
}
