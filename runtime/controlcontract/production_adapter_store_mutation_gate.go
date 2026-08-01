package controlcontract

type ProductionAdapterStoreMutationDescriptor struct {
	ContractVersion                string           `json:"contract_version,omitempty"`
	Projected                      bool             `json:"projected"`
	Available                      bool             `json:"available"`
	Status                         HostActionStatus `json:"status,omitempty"`
	Mode                           string           `json:"mode,omitempty"`
	ReadyForStoreMutationRequest   bool             `json:"ready_for_store_mutation_request"`
	CoreInvocationExecuted         bool             `json:"core_invocation_executed"`
	DurableWriteByCore             bool             `json:"durable_write_by_core"`
	ObjectiveStoreWriteByCore      bool             `json:"objective_store_write_by_core"`
	RunstoreWriteByCore            bool             `json:"runstore_write_by_core"`
	StoreMutationDescriptorRef     DisplaySafeRef   `json:"store_mutation_descriptor_ref,omitempty"`
	StoreAdapterRef                DisplaySafeRef   `json:"store_adapter_ref,omitempty"`
	OwnerRef                       DisplaySafeRef   `json:"owner_ref,omitempty"`
	HostRunstoreRef                DisplaySafeRef   `json:"host_runstore_ref,omitempty"`
	HostObjectiveStoreRef          DisplaySafeRef   `json:"host_objective_store_ref,omitempty"`
	SupportsRunstoreMutation       bool             `json:"supports_runstore_mutation"`
	SupportsObjectiveStoreMutation bool             `json:"supports_objective_store_mutation"`
	SupportsTransactionReplay      bool             `json:"supports_transaction_replay"`
	SupportedMutationKinds         []string         `json:"supported_mutation_kinds,omitempty"`
	TransactionContractRef         DisplaySafeRef   `json:"transaction_contract_ref,omitempty"`
	IdempotencyRef                 DisplaySafeRef   `json:"idempotency_ref,omitempty"`
	IdempotencyContractRef         DisplaySafeRef   `json:"idempotency_contract_ref,omitempty"`
	ReplayContractRef              DisplaySafeRef   `json:"replay_contract_ref,omitempty"`
	ReadbackContractRef            DisplaySafeRef   `json:"readback_contract_ref,omitempty"`
	RedactionPolicyRef             DisplaySafeRef   `json:"redaction_policy_ref,omitempty"`
	TimeoutPolicyRef               DisplaySafeRef   `json:"timeout_policy_ref,omitempty"`
	PolicyRefs                     []DisplaySafeRef `json:"policy_refs,omitempty"`
	RequiredPolicyRefs             []DisplaySafeRef `json:"required_policy_refs,omitempty"`
	ApprovalRefs                   []DisplaySafeRef `json:"approval_refs,omitempty"`
	RequiredApprovalRefs           []DisplaySafeRef `json:"required_approval_refs,omitempty"`
	MissingInputs                  []MissingInput   `json:"missing_inputs,omitempty"`
	BlockedReasons                 []string         `json:"blocked_reasons,omitempty"`
	FailureClass                   FailureClass     `json:"failure_class,omitempty"`
	Boundaries                     []Boundary       `json:"boundaries,omitempty"`
	NextHostAction                 NextHostAction   `json:"next_host_action,omitempty"`
	RunnerEffect                   string           `json:"runner_effect,omitempty"`
	PromptEffect                   string           `json:"prompt_effect,omitempty"`
	RawOutputLoaded                bool             `json:"raw_output_loaded"`
}

type ProductionAdapterStoreMutationRequestInput struct {
	StoreMutationRequestRef          DisplaySafeRef                           `json:"store_mutation_request_ref,omitempty"`
	StoreMutationDescriptor          ProductionAdapterStoreMutationDescriptor `json:"store_mutation_descriptor,omitempty"`
	SourceDurableResultRef           DisplaySafeRef                           `json:"source_durable_result_ref,omitempty"`
	SourceDurableEventRef            DisplaySafeRef                           `json:"source_durable_event_ref,omitempty"`
	SourceRunstoreRef                DisplaySafeRef                           `json:"source_runstore_ref,omitempty"`
	SourceObjectiveStateRef          DisplaySafeRef                           `json:"source_objective_state_ref,omitempty"`
	HostStoreMutationConfirmationRef DisplaySafeRef                           `json:"host_store_mutation_confirmation_ref,omitempty"`
	ExpectedMutationResultRef        DisplaySafeRef                           `json:"expected_mutation_result_ref,omitempty"`
	ExpectedTransactionRef           DisplaySafeRef                           `json:"expected_transaction_ref,omitempty"`
	ExpectedReadbackRef              DisplaySafeRef                           `json:"expected_readback_ref,omitempty"`
	RawOutputLoaded                  bool                                     `json:"raw_output_loaded"`
}

type ProductionAdapterStoreMutationRequest struct {
	ContractVersion                  string           `json:"contract_version,omitempty"`
	Projected                        bool             `json:"projected"`
	Available                        bool             `json:"available"`
	Status                           HostActionStatus `json:"status,omitempty"`
	Mode                             string           `json:"mode,omitempty"`
	ReadyForHostStoreMutation        bool             `json:"ready_for_host_store_mutation"`
	HostStoreMutationAuthorized      bool             `json:"host_store_mutation_authorized"`
	HostMayExecuteStoreMutation      bool             `json:"host_may_execute_store_mutation"`
	CoreInvocationExecuted           bool             `json:"core_invocation_executed"`
	DurableWriteByCore               bool             `json:"durable_write_by_core"`
	ObjectiveStoreWriteByCore        bool             `json:"objective_store_write_by_core"`
	RunstoreWriteByCore              bool             `json:"runstore_write_by_core"`
	StoreMutationRequestRef          DisplaySafeRef   `json:"store_mutation_request_ref,omitempty"`
	StoreMutationDescriptorRef       DisplaySafeRef   `json:"store_mutation_descriptor_ref,omitempty"`
	StoreAdapterRef                  DisplaySafeRef   `json:"store_adapter_ref,omitempty"`
	OwnerRef                         DisplaySafeRef   `json:"owner_ref,omitempty"`
	HostRunstoreRef                  DisplaySafeRef   `json:"host_runstore_ref,omitempty"`
	HostObjectiveStoreRef            DisplaySafeRef   `json:"host_objective_store_ref,omitempty"`
	SupportsRunstoreMutation         bool             `json:"supports_runstore_mutation"`
	SupportsObjectiveStoreMutation   bool             `json:"supports_objective_store_mutation"`
	SupportsTransactionReplay        bool             `json:"supports_transaction_replay"`
	SupportedMutationKinds           []string         `json:"supported_mutation_kinds,omitempty"`
	SourceDurableResultRef           DisplaySafeRef   `json:"source_durable_result_ref,omitempty"`
	SourceDurableEventRef            DisplaySafeRef   `json:"source_durable_event_ref,omitempty"`
	SourceRunstoreRef                DisplaySafeRef   `json:"source_runstore_ref,omitempty"`
	SourceObjectiveStateRef          DisplaySafeRef   `json:"source_objective_state_ref,omitempty"`
	HostStoreMutationConfirmationRef DisplaySafeRef   `json:"host_store_mutation_confirmation_ref,omitempty"`
	ExpectedMutationResultRef        DisplaySafeRef   `json:"expected_mutation_result_ref,omitempty"`
	ExpectedTransactionRef           DisplaySafeRef   `json:"expected_transaction_ref,omitempty"`
	ExpectedReadbackRef              DisplaySafeRef   `json:"expected_readback_ref,omitempty"`
	TransactionContractRef           DisplaySafeRef   `json:"transaction_contract_ref,omitempty"`
	IdempotencyRef                   DisplaySafeRef   `json:"idempotency_ref,omitempty"`
	IdempotencyContractRef           DisplaySafeRef   `json:"idempotency_contract_ref,omitempty"`
	ReplayContractRef                DisplaySafeRef   `json:"replay_contract_ref,omitempty"`
	ReadbackContractRef              DisplaySafeRef   `json:"readback_contract_ref,omitempty"`
	RedactionPolicyRef               DisplaySafeRef   `json:"redaction_policy_ref,omitempty"`
	TimeoutPolicyRef                 DisplaySafeRef   `json:"timeout_policy_ref,omitempty"`
	PolicyRefs                       []DisplaySafeRef `json:"policy_refs,omitempty"`
	RequiredPolicyRefs               []DisplaySafeRef `json:"required_policy_refs,omitempty"`
	ApprovalRefs                     []DisplaySafeRef `json:"approval_refs,omitempty"`
	RequiredApprovalRefs             []DisplaySafeRef `json:"required_approval_refs,omitempty"`
	MissingInputs                    []MissingInput   `json:"missing_inputs,omitempty"`
	BlockedReasons                   []string         `json:"blocked_reasons,omitempty"`
	FailureClass                     FailureClass     `json:"failure_class,omitempty"`
	Boundaries                       []Boundary       `json:"boundaries,omitempty"`
	NextHostAction                   NextHostAction   `json:"next_host_action,omitempty"`
	RunnerEffect                     string           `json:"runner_effect,omitempty"`
	PromptEffect                     string           `json:"prompt_effect,omitempty"`
	RawOutputLoaded                  bool             `json:"raw_output_loaded"`
}

type ProductionAdapterStoreMutationResultInput struct {
	StoreMutationResultRef     DisplaySafeRef                        `json:"store_mutation_result_ref,omitempty"`
	StoreMutationRequest       ProductionAdapterStoreMutationRequest `json:"store_mutation_request,omitempty"`
	HostStoreAdapterRunRef     DisplaySafeRef                        `json:"host_store_adapter_run_ref,omitempty"`
	HostStoreMutationReported  bool                                  `json:"host_store_mutation_reported"`
	HostStoreMutationSucceeded bool                                  `json:"host_store_mutation_succeeded"`
	HostStoreMutationFailed    bool                                  `json:"host_store_mutation_failed"`
	AppliedTransactionRef      DisplaySafeRef                        `json:"applied_transaction_ref,omitempty"`
	AppliedRunstoreRef         DisplaySafeRef                        `json:"applied_runstore_ref,omitempty"`
	AppliedObjectiveStateRef   DisplaySafeRef                        `json:"applied_objective_state_ref,omitempty"`
	FailureRef                 DisplaySafeRef                        `json:"failure_ref,omitempty"`
	CompensationRef            DisplaySafeRef                        `json:"compensation_ref,omitempty"`
	StoreMutationEvidenceRefs  []DisplaySafeRef                      `json:"store_mutation_evidence_refs,omitempty"`
	RawOutputLoaded            bool                                  `json:"raw_output_loaded"`
}

type ProductionAdapterStoreMutationResult struct {
	ContractVersion                string           `json:"contract_version,omitempty"`
	Projected                      bool             `json:"projected"`
	Available                      bool             `json:"available"`
	Status                         HostActionStatus `json:"status,omitempty"`
	Mode                           string           `json:"mode,omitempty"`
	ReadyForStoreMutationReadback  bool             `json:"ready_for_store_mutation_readback"`
	HostStoreMutationReported      bool             `json:"host_store_mutation_reported"`
	HostStoreMutationSucceeded     bool             `json:"host_store_mutation_succeeded"`
	HostStoreMutationFailed        bool             `json:"host_store_mutation_failed"`
	HostStoreMutationRecorded      bool             `json:"host_store_mutation_recorded"`
	CoreInvocationExecuted         bool             `json:"core_invocation_executed"`
	DurableWriteByCore             bool             `json:"durable_write_by_core"`
	ObjectiveStoreWriteByCore      bool             `json:"objective_store_write_by_core"`
	RunstoreWriteByCore            bool             `json:"runstore_write_by_core"`
	StoreMutationResultRef         DisplaySafeRef   `json:"store_mutation_result_ref,omitempty"`
	ExpectedMutationResultRef      DisplaySafeRef   `json:"expected_mutation_result_ref,omitempty"`
	StoreMutationRequestRef        DisplaySafeRef   `json:"store_mutation_request_ref,omitempty"`
	StoreMutationDescriptorRef     DisplaySafeRef   `json:"store_mutation_descriptor_ref,omitempty"`
	StoreAdapterRef                DisplaySafeRef   `json:"store_adapter_ref,omitempty"`
	HostStoreAdapterRunRef         DisplaySafeRef   `json:"host_store_adapter_run_ref,omitempty"`
	HostRunstoreRef                DisplaySafeRef   `json:"host_runstore_ref,omitempty"`
	HostObjectiveStoreRef          DisplaySafeRef   `json:"host_objective_store_ref,omitempty"`
	SupportsRunstoreMutation       bool             `json:"supports_runstore_mutation"`
	SupportsObjectiveStoreMutation bool             `json:"supports_objective_store_mutation"`
	SupportsTransactionReplay      bool             `json:"supports_transaction_replay"`
	SourceDurableResultRef         DisplaySafeRef   `json:"source_durable_result_ref,omitempty"`
	SourceDurableEventRef          DisplaySafeRef   `json:"source_durable_event_ref,omitempty"`
	SourceRunstoreRef              DisplaySafeRef   `json:"source_runstore_ref,omitempty"`
	SourceObjectiveStateRef        DisplaySafeRef   `json:"source_objective_state_ref,omitempty"`
	ExpectedTransactionRef         DisplaySafeRef   `json:"expected_transaction_ref,omitempty"`
	ExpectedReadbackRef            DisplaySafeRef   `json:"expected_readback_ref,omitempty"`
	AppliedTransactionRef          DisplaySafeRef   `json:"applied_transaction_ref,omitempty"`
	AppliedRunstoreRef             DisplaySafeRef   `json:"applied_runstore_ref,omitempty"`
	AppliedObjectiveStateRef       DisplaySafeRef   `json:"applied_objective_state_ref,omitempty"`
	FailureRef                     DisplaySafeRef   `json:"failure_ref,omitempty"`
	CompensationRef                DisplaySafeRef   `json:"compensation_ref,omitempty"`
	TransactionContractRef         DisplaySafeRef   `json:"transaction_contract_ref,omitempty"`
	IdempotencyRef                 DisplaySafeRef   `json:"idempotency_ref,omitempty"`
	IdempotencyContractRef         DisplaySafeRef   `json:"idempotency_contract_ref,omitempty"`
	ReplayContractRef              DisplaySafeRef   `json:"replay_contract_ref,omitempty"`
	ReadbackContractRef            DisplaySafeRef   `json:"readback_contract_ref,omitempty"`
	StoreMutationEvidenceRefs      []DisplaySafeRef `json:"store_mutation_evidence_refs,omitempty"`
	MissingInputs                  []MissingInput   `json:"missing_inputs,omitempty"`
	BlockedReasons                 []string         `json:"blocked_reasons,omitempty"`
	FailureClass                   FailureClass     `json:"failure_class,omitempty"`
	Boundaries                     []Boundary       `json:"boundaries,omitempty"`
	NextHostAction                 NextHostAction   `json:"next_host_action,omitempty"`
	RunnerEffect                   string           `json:"runner_effect,omitempty"`
	PromptEffect                   string           `json:"prompt_effect,omitempty"`
	RawOutputLoaded                bool             `json:"raw_output_loaded"`
}

type ProductionAdapterStoreMutationReadbackInput struct {
	StoreMutationReadbackRef  DisplaySafeRef                       `json:"store_mutation_readback_ref,omitempty"`
	StoreMutationResult       ProductionAdapterStoreMutationResult `json:"store_mutation_result,omitempty"`
	ObservedTransactionRef    DisplaySafeRef                       `json:"observed_transaction_ref,omitempty"`
	ObservedRunstoreRef       DisplaySafeRef                       `json:"observed_runstore_ref,omitempty"`
	ObservedObjectiveStateRef DisplaySafeRef                       `json:"observed_objective_state_ref,omitempty"`
	ReplayRef                 DisplaySafeRef                       `json:"replay_ref,omitempty"`
	RawOutputLoaded           bool                                 `json:"raw_output_loaded"`
}

type ProductionAdapterStoreMutationReadback struct {
	ContractVersion                string           `json:"contract_version,omitempty"`
	Projected                      bool             `json:"projected"`
	Available                      bool             `json:"available"`
	Status                         HostActionStatus `json:"status,omitempty"`
	Mode                           string           `json:"mode,omitempty"`
	StoreMutationReadbackBound     bool             `json:"store_mutation_readback_bound"`
	TransactionReplayVerified      bool             `json:"transaction_replay_verified"`
	ReadyForDownstreamReadback     bool             `json:"ready_for_downstream_readback"`
	CoreInvocationExecuted         bool             `json:"core_invocation_executed"`
	DurableWriteByCore             bool             `json:"durable_write_by_core"`
	ObjectiveStoreWriteByCore      bool             `json:"objective_store_write_by_core"`
	RunstoreWriteByCore            bool             `json:"runstore_write_by_core"`
	StoreMutationReadbackRef       DisplaySafeRef   `json:"store_mutation_readback_ref,omitempty"`
	StoreMutationResultRef         DisplaySafeRef   `json:"store_mutation_result_ref,omitempty"`
	StoreMutationRequestRef        DisplaySafeRef   `json:"store_mutation_request_ref,omitempty"`
	StoreMutationDescriptorRef     DisplaySafeRef   `json:"store_mutation_descriptor_ref,omitempty"`
	StoreAdapterRef                DisplaySafeRef   `json:"store_adapter_ref,omitempty"`
	HostStoreAdapterRunRef         DisplaySafeRef   `json:"host_store_adapter_run_ref,omitempty"`
	HostRunstoreRef                DisplaySafeRef   `json:"host_runstore_ref,omitempty"`
	HostObjectiveStoreRef          DisplaySafeRef   `json:"host_objective_store_ref,omitempty"`
	SupportsRunstoreMutation       bool             `json:"supports_runstore_mutation"`
	SupportsObjectiveStoreMutation bool             `json:"supports_objective_store_mutation"`
	SupportsTransactionReplay      bool             `json:"supports_transaction_replay"`
	ExpectedTransactionRef         DisplaySafeRef   `json:"expected_transaction_ref,omitempty"`
	ExpectedReadbackRef            DisplaySafeRef   `json:"expected_readback_ref,omitempty"`
	AppliedTransactionRef          DisplaySafeRef   `json:"applied_transaction_ref,omitempty"`
	AppliedRunstoreRef             DisplaySafeRef   `json:"applied_runstore_ref,omitempty"`
	AppliedObjectiveStateRef       DisplaySafeRef   `json:"applied_objective_state_ref,omitempty"`
	ObservedTransactionRef         DisplaySafeRef   `json:"observed_transaction_ref,omitempty"`
	ObservedRunstoreRef            DisplaySafeRef   `json:"observed_runstore_ref,omitempty"`
	ObservedObjectiveStateRef      DisplaySafeRef   `json:"observed_objective_state_ref,omitempty"`
	ReplayRef                      DisplaySafeRef   `json:"replay_ref,omitempty"`
	MissingInputs                  []MissingInput   `json:"missing_inputs,omitempty"`
	BlockedReasons                 []string         `json:"blocked_reasons,omitempty"`
	FailureClass                   FailureClass     `json:"failure_class,omitempty"`
	Boundaries                     []Boundary       `json:"boundaries,omitempty"`
	NextHostAction                 NextHostAction   `json:"next_host_action,omitempty"`
	RunnerEffect                   string           `json:"runner_effect,omitempty"`
	PromptEffect                   string           `json:"prompt_effect,omitempty"`
	RawOutputLoaded                bool             `json:"raw_output_loaded"`
}

func BuildProductionAdapterStoreMutationDescriptor(input ProductionAdapterStoreMutationDescriptor) ProductionAdapterStoreMutationDescriptor {
	unsafeInput := productionAdapterStoreMutationDescriptorOutputUnsafe(input)
	result := input.Normalize()
	result.Status = HostActionBlocked
	result.ReadyForStoreMutationRequest = false
	result.FailureClass = firstFailureClass(result.FailureClass, FailureNone)
	result.Boundaries = productionAdapterStoreMutationDescriptorBoundaries(result.Boundaries)
	if unsafeInput {
		result = productionAdapterStoreMutationDescriptorBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		result.RawOutputLoaded = true
		result.NextHostAction = "provide_display_safe_refs"
		return result.Normalize()
	}
	if !result.Available {
		result.Status = HostActionNotReady
		result.FailureClass = firstFailureClass(result.FailureClass, FailureConfigMissing)
		result.NextHostAction = firstNextHostAction(result.NextHostAction, "configure_store_mutation_adapter")
		return result.Normalize()
	}
	if result.StoreMutationDescriptorRef == "" {
		result = productionAdapterStoreMutationDescriptorBlock(result, FailureEvidenceMissing, "store_mutation_descriptor_ref_missing", "host:store_mutation_descriptor_ref", "provide_store_mutation_descriptor_ref")
	}
	if result.StoreAdapterRef == "" {
		result = productionAdapterStoreMutationDescriptorBlock(result, FailureHostAdapterMissing, "store_adapter_ref_missing", "host:store_adapter_ref", "configure_store_mutation_adapter")
	}
	if result.OwnerRef == "" {
		result = productionAdapterStoreMutationDescriptorBlock(result, FailureConfigMissing, "store_mutation_owner_ref_missing", "host:store_mutation_owner_ref", "provide_store_mutation_owner_ref")
	}
	if !result.SupportsRunstoreMutation && !result.SupportsObjectiveStoreMutation {
		result = productionAdapterStoreMutationDescriptorBlock(result, FailureUnsupportedOperation, "store_mutation_scope_missing", "host:store_mutation_scope", "configure_store_mutation_scope")
	}
	if !result.SupportsTransactionReplay {
		result = productionAdapterStoreMutationDescriptorBlock(result, FailureConfigMissing, "transaction_replay_not_supported", "host:store_mutation_transaction_replay", "configure_store_mutation_replay")
	}
	if result.SupportsRunstoreMutation && result.HostRunstoreRef == "" {
		result = productionAdapterStoreMutationDescriptorBlock(result, FailureConfigMissing, "host_runstore_ref_missing", "host:runstore_ref", "configure_host_runstore_ref")
	}
	if result.SupportsObjectiveStoreMutation && result.HostObjectiveStoreRef == "" {
		result = productionAdapterStoreMutationDescriptorBlock(result, FailureConfigMissing, "host_objective_store_ref_missing", "host:objective_store_ref", "configure_host_objective_store_ref")
	}
	if len(result.SupportedMutationKinds) == 0 {
		result = productionAdapterStoreMutationDescriptorBlock(result, FailureConfigMissing, "supported_mutation_kinds_missing", "host:supported_store_mutation_kinds", "configure_supported_store_mutation_kinds")
	}
	for _, check := range []struct {
		ref     DisplaySafeRef
		reason  string
		missing MissingInput
		next    NextHostAction
	}{
		{result.TransactionContractRef, "transaction_contract_ref_missing", "host:store_mutation_transaction_contract_ref", "provide_store_mutation_transaction_contract"},
		{result.IdempotencyRef, "idempotency_ref_missing", "host:store_mutation_idempotency_ref", "provide_store_mutation_idempotency_ref"},
		{result.IdempotencyContractRef, "idempotency_contract_ref_missing", "host:store_mutation_idempotency_contract_ref", "provide_store_mutation_idempotency_contract"},
		{result.ReplayContractRef, "replay_contract_ref_missing", "host:store_mutation_replay_contract_ref", "provide_store_mutation_replay_contract"},
		{result.ReadbackContractRef, "readback_contract_ref_missing", "host:store_mutation_readback_contract_ref", "provide_store_mutation_readback_contract"},
	} {
		if check.ref == "" {
			result = productionAdapterStoreMutationDescriptorBlock(result, FailureConfigMissing, check.reason, check.missing, check.next)
		}
	}
	if len(result.MissingInputs) == 0 && len(result.BlockedReasons) == 0 {
		result.Status = HostActionReady
		result.ReadyForStoreMutationRequest = true
		result.NextHostAction = "host_may_prepare_store_mutation_request"
		result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_host_store_mutation_request")
	}
	return result.Normalize()
}

func BuildProductionAdapterStoreMutationRequest(input ProductionAdapterStoreMutationRequestInput) ProductionAdapterStoreMutationRequest {
	if productionAdapterStoreMutationDescriptorEmpty(input.StoreMutationDescriptor) {
		return unavailableProductionAdapterStoreMutationRequest()
	}
	descriptor := input.StoreMutationDescriptor.Normalize()
	result := ProductionAdapterStoreMutationRequest{
		ContractVersion:                  ContractVersion,
		Projected:                        true,
		Available:                        descriptor.Available,
		Status:                           HostActionBlocked,
		Mode:                             "production_adapter_store_mutation_request",
		StoreMutationRequestRef:          normalizeOneDisplaySafeRef(input.StoreMutationRequestRef),
		StoreMutationDescriptorRef:       descriptor.StoreMutationDescriptorRef,
		StoreAdapterRef:                  descriptor.StoreAdapterRef,
		OwnerRef:                         descriptor.OwnerRef,
		HostRunstoreRef:                  descriptor.HostRunstoreRef,
		HostObjectiveStoreRef:            descriptor.HostObjectiveStoreRef,
		SupportsRunstoreMutation:         descriptor.SupportsRunstoreMutation,
		SupportsObjectiveStoreMutation:   descriptor.SupportsObjectiveStoreMutation,
		SupportsTransactionReplay:        descriptor.SupportsTransactionReplay,
		SupportedMutationKinds:           cloneStringSlice(descriptor.SupportedMutationKinds),
		SourceDurableResultRef:           normalizeOneDisplaySafeRef(input.SourceDurableResultRef),
		SourceDurableEventRef:            normalizeOneDisplaySafeRef(input.SourceDurableEventRef),
		SourceRunstoreRef:                normalizeOneDisplaySafeRef(input.SourceRunstoreRef),
		SourceObjectiveStateRef:          normalizeOneDisplaySafeRef(input.SourceObjectiveStateRef),
		HostStoreMutationConfirmationRef: normalizeOneDisplaySafeRef(input.HostStoreMutationConfirmationRef),
		ExpectedMutationResultRef:        normalizeOneDisplaySafeRef(input.ExpectedMutationResultRef),
		ExpectedTransactionRef:           normalizeOneDisplaySafeRef(input.ExpectedTransactionRef),
		ExpectedReadbackRef:              normalizeOneDisplaySafeRef(input.ExpectedReadbackRef),
		TransactionContractRef:           descriptor.TransactionContractRef,
		IdempotencyRef:                   descriptor.IdempotencyRef,
		IdempotencyContractRef:           descriptor.IdempotencyContractRef,
		ReplayContractRef:                descriptor.ReplayContractRef,
		ReadbackContractRef:              descriptor.ReadbackContractRef,
		RedactionPolicyRef:               descriptor.RedactionPolicyRef,
		TimeoutPolicyRef:                 descriptor.TimeoutPolicyRef,
		PolicyRefs:                       cloneDisplaySafeRefs(descriptor.PolicyRefs),
		RequiredPolicyRefs:               cloneDisplaySafeRefs(descriptor.RequiredPolicyRefs),
		ApprovalRefs:                     cloneDisplaySafeRefs(descriptor.ApprovalRefs),
		RequiredApprovalRefs:             cloneDisplaySafeRefs(descriptor.RequiredApprovalRefs),
		FailureClass:                     FailureNone,
		Boundaries:                       productionAdapterStoreMutationRequestBoundaries(descriptor.Boundaries),
		RunnerEffect:                     "none",
		PromptEffect:                     "none",
		RawOutputLoaded:                  input.RawOutputLoaded || descriptor.RawOutputLoaded,
	}
	if productionAdapterStoreMutationRequestUnsafe(input, descriptor) {
		result = productionAdapterStoreMutationRequestBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if !descriptor.ReadyForStoreMutationRequest {
		result = productionAdapterStoreMutationRequestBlock(result, firstFailureClass(descriptor.FailureClass, FailureConfigMissing), "store_mutation_descriptor_not_ready", "host:store_mutation_descriptor", firstNextHostAction(descriptor.NextHostAction, "review_store_mutation_descriptor"))
	}
	for _, check := range []struct {
		ref     DisplaySafeRef
		reason  string
		missing MissingInput
		next    NextHostAction
		failure FailureClass
	}{
		{result.StoreMutationRequestRef, "store_mutation_request_ref_missing", "host:store_mutation_request_ref", "provide_store_mutation_request_ref", FailureEvidenceMissing},
		{result.SourceDurableResultRef, "source_durable_result_ref_missing", "host:source_durable_result_ref", "provide_source_durable_result_ref", FailureEvidenceMissing},
		{result.SourceDurableEventRef, "source_durable_event_ref_missing", "host:source_durable_event_ref", "provide_source_durable_event_ref", FailureEvidenceMissing},
		{result.HostStoreMutationConfirmationRef, "host_store_mutation_confirmation_ref_missing", "host:store_mutation_confirmation_ref", "request_store_mutation_confirmation", FailureAuthorizationMissing},
		{result.ExpectedMutationResultRef, "expected_mutation_result_ref_missing", "host:expected_store_mutation_result_ref", "provide_expected_store_mutation_result_ref", FailureEvidenceMissing},
		{result.ExpectedTransactionRef, "expected_transaction_ref_missing", "host:expected_store_transaction_ref", "provide_expected_store_transaction_ref", FailureEvidenceMissing},
		{result.ExpectedReadbackRef, "expected_readback_ref_missing", "host:store_mutation_expected_readback_ref", "provide_store_mutation_expected_readback_ref", FailureEvidenceMissing},
	} {
		if check.ref == "" {
			result = productionAdapterStoreMutationRequestBlock(result, check.failure, check.reason, check.missing, check.next)
		}
	}
	if result.SupportsRunstoreMutation {
		if result.SourceRunstoreRef == "" {
			result = productionAdapterStoreMutationRequestBlock(result, FailureEvidenceMissing, "source_runstore_ref_missing", "host:source_runstore_ref", "provide_source_runstore_ref")
		} else if result.HostRunstoreRef != "" && result.SourceRunstoreRef != result.HostRunstoreRef {
			result = productionAdapterStoreMutationRequestBlock(result, FailureVerificationFailed, "store_mutation_runstore_ref_mismatch", "host:runstore_ref", "review_store_mutation_request")
		}
	}
	if result.SupportsObjectiveStoreMutation && result.SourceObjectiveStateRef == "" {
		result = productionAdapterStoreMutationRequestBlock(result, FailureEvidenceMissing, "source_objective_state_ref_missing", "host:source_objective_state_ref", "provide_source_objective_state_ref")
	}
	if len(result.MissingInputs) == 0 && len(result.BlockedReasons) == 0 {
		result.Status = HostActionReady
		result.ReadyForHostStoreMutation = true
		result.HostStoreMutationAuthorized = true
		result.HostMayExecuteStoreMutation = true
		result.NextHostAction = "host_may_execute_store_mutation_adapter"
		result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_host_store_mutation", "host_may_execute_store_adapter", "core_store_mutation_not_executed")
	}
	return result.Normalize()
}

func BuildProductionAdapterStoreMutationResult(input ProductionAdapterStoreMutationResultInput) ProductionAdapterStoreMutationResult {
	if productionAdapterStoreMutationRequestEmpty(input.StoreMutationRequest) {
		return unavailableProductionAdapterStoreMutationResult()
	}
	request := input.StoreMutationRequest.Normalize()
	result := ProductionAdapterStoreMutationResult{
		ContractVersion:                ContractVersion,
		Projected:                      true,
		Available:                      request.Available,
		Status:                         HostActionBlocked,
		Mode:                           "production_adapter_store_mutation_result",
		HostStoreMutationReported:      input.HostStoreMutationReported,
		HostStoreMutationSucceeded:     input.HostStoreMutationSucceeded,
		HostStoreMutationFailed:        input.HostStoreMutationFailed,
		StoreMutationResultRef:         normalizeOneDisplaySafeRef(input.StoreMutationResultRef),
		ExpectedMutationResultRef:      request.ExpectedMutationResultRef,
		StoreMutationRequestRef:        request.StoreMutationRequestRef,
		StoreMutationDescriptorRef:     request.StoreMutationDescriptorRef,
		StoreAdapterRef:                request.StoreAdapterRef,
		HostStoreAdapterRunRef:         normalizeOneDisplaySafeRef(input.HostStoreAdapterRunRef),
		HostRunstoreRef:                request.HostRunstoreRef,
		HostObjectiveStoreRef:          request.HostObjectiveStoreRef,
		SupportsRunstoreMutation:       request.SupportsRunstoreMutation,
		SupportsObjectiveStoreMutation: request.SupportsObjectiveStoreMutation,
		SupportsTransactionReplay:      request.SupportsTransactionReplay,
		SourceDurableResultRef:         request.SourceDurableResultRef,
		SourceDurableEventRef:          request.SourceDurableEventRef,
		SourceRunstoreRef:              request.SourceRunstoreRef,
		SourceObjectiveStateRef:        request.SourceObjectiveStateRef,
		ExpectedTransactionRef:         request.ExpectedTransactionRef,
		ExpectedReadbackRef:            request.ExpectedReadbackRef,
		AppliedTransactionRef:          normalizeOneDisplaySafeRef(input.AppliedTransactionRef),
		AppliedRunstoreRef:             normalizeOneDisplaySafeRef(input.AppliedRunstoreRef),
		AppliedObjectiveStateRef:       normalizeOneDisplaySafeRef(input.AppliedObjectiveStateRef),
		FailureRef:                     normalizeOneDisplaySafeRef(input.FailureRef),
		CompensationRef:                normalizeOneDisplaySafeRef(input.CompensationRef),
		TransactionContractRef:         request.TransactionContractRef,
		IdempotencyRef:                 request.IdempotencyRef,
		IdempotencyContractRef:         request.IdempotencyContractRef,
		ReplayContractRef:              request.ReplayContractRef,
		ReadbackContractRef:            request.ReadbackContractRef,
		StoreMutationEvidenceRefs:      normalizeDisplaySafeRefs(input.StoreMutationEvidenceRefs),
		FailureClass:                   FailureNone,
		Boundaries:                     productionAdapterStoreMutationResultBoundaries(request.Boundaries),
		RunnerEffect:                   "none",
		PromptEffect:                   "none",
		RawOutputLoaded:                input.RawOutputLoaded || request.RawOutputLoaded,
	}
	if productionAdapterStoreMutationResultUnsafe(input, request) {
		result = productionAdapterStoreMutationResultBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if !request.ReadyForHostStoreMutation || !request.HostMayExecuteStoreMutation {
		result = productionAdapterStoreMutationResultBlock(result, firstFailureClass(request.FailureClass, FailureConfigMissing), "store_mutation_request_not_ready", "host:store_mutation_request", firstNextHostAction(request.NextHostAction, "review_store_mutation_request"))
	}
	if !input.HostStoreMutationReported {
		result = productionAdapterStoreMutationResultBlock(result, FailureEvidenceMissing, "store_mutation_not_reported", "host:store_mutation_report", "provide_store_mutation_report")
	}
	if input.HostStoreMutationSucceeded && input.HostStoreMutationFailed {
		result = productionAdapterStoreMutationResultBlock(result, FailureVerificationFailed, "store_mutation_status_conflict", "host:store_mutation_status", "review_store_mutation_result")
	}
	if result.HostStoreAdapterRunRef == "" {
		result = productionAdapterStoreMutationResultBlock(result, FailureEvidenceMissing, "host_store_adapter_run_ref_missing", "host:store_adapter_run_ref", "provide_store_mutation_report")
	}
	if result.StoreMutationResultRef == "" {
		result = productionAdapterStoreMutationResultBlock(result, FailureEvidenceMissing, "store_mutation_result_ref_missing", "host:store_mutation_result_ref", "provide_store_mutation_result_ref")
	} else if result.ExpectedMutationResultRef != "" && result.StoreMutationResultRef != result.ExpectedMutationResultRef {
		result = productionAdapterStoreMutationResultBlock(result, FailureVerificationFailed, "store_mutation_result_ref_mismatch", "host:store_mutation_result_ref", "review_store_mutation_result")
	}
	if len(result.MissingInputs) > 0 || len(result.BlockedReasons) > 0 {
		return result.Normalize()
	}
	if input.HostStoreMutationFailed {
		if result.FailureRef == "" {
			result = productionAdapterStoreMutationResultBlock(result, FailureEvidenceMissing, "store_mutation_failure_ref_missing", "host:store_mutation_failure_ref", "provide_store_mutation_failure_ref")
			return result.Normalize()
		}
		result.Status = HostActionReviewRequired
		result.HostStoreMutationRecorded = true
		result.FailureClass = FailureVerificationFailed
		result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, "store_mutation_failed")
		result.NextHostAction = "review_store_mutation_failure"
		result.Boundaries = AppendBoundaries(result.Boundaries, "store_mutation_failed", "compensation_not_executed")
		return result.Normalize()
	}
	if !input.HostStoreMutationSucceeded {
		result = productionAdapterStoreMutationResultBlock(result, FailureEvidenceMissing, "store_mutation_status_missing", "host:store_mutation_status", "provide_store_mutation_report")
		return result.Normalize()
	}
	if result.AppliedTransactionRef == "" {
		result = productionAdapterStoreMutationResultBlock(result, FailureEvidenceMissing, "applied_transaction_ref_missing", "host:applied_store_transaction_ref", "provide_store_mutation_result")
	} else if result.ExpectedTransactionRef != "" && result.AppliedTransactionRef != result.ExpectedTransactionRef {
		result = productionAdapterStoreMutationResultBlock(result, FailureVerificationFailed, "store_transaction_ref_mismatch", "host:store_transaction_ref", "review_store_mutation_result")
	}
	if result.SupportsRunstoreMutation {
		if result.AppliedRunstoreRef == "" {
			result = productionAdapterStoreMutationResultBlock(result, FailureEvidenceMissing, "applied_runstore_ref_missing", "host:applied_runstore_ref", "provide_store_mutation_result")
		} else if result.HostRunstoreRef != "" && result.AppliedRunstoreRef != result.HostRunstoreRef {
			result = productionAdapterStoreMutationResultBlock(result, FailureVerificationFailed, "store_mutation_applied_runstore_ref_mismatch", "host:applied_runstore_ref", "review_store_mutation_result")
		}
	}
	if result.SupportsObjectiveStoreMutation {
		if result.AppliedObjectiveStateRef == "" {
			result = productionAdapterStoreMutationResultBlock(result, FailureEvidenceMissing, "applied_objective_state_ref_missing", "host:applied_objective_state_ref", "provide_store_mutation_result")
		} else if result.SourceObjectiveStateRef != "" && result.AppliedObjectiveStateRef != result.SourceObjectiveStateRef {
			result = productionAdapterStoreMutationResultBlock(result, FailureVerificationFailed, "store_mutation_objective_state_ref_mismatch", "host:applied_objective_state_ref", "review_store_mutation_result")
		}
	}
	if len(result.MissingInputs) == 0 && len(result.BlockedReasons) == 0 {
		result.Status = HostActionRecorded
		result.HostStoreMutationRecorded = true
		result.ReadyForStoreMutationReadback = true
		result.NextHostAction = "bind_store_mutation_readback"
		result.Boundaries = AppendBoundaries(result.Boundaries, "host_store_mutation_recorded", "ready_for_store_mutation_readback")
	}
	return result.Normalize()
}

func BuildProductionAdapterStoreMutationReadback(input ProductionAdapterStoreMutationReadbackInput) ProductionAdapterStoreMutationReadback {
	if productionAdapterStoreMutationResultEmpty(input.StoreMutationResult) {
		return unavailableProductionAdapterStoreMutationReadback()
	}
	mutationResult := input.StoreMutationResult.Normalize()
	result := ProductionAdapterStoreMutationReadback{
		ContractVersion:                ContractVersion,
		Projected:                      true,
		Available:                      mutationResult.Available,
		Status:                         HostActionBlocked,
		Mode:                           "production_adapter_store_mutation_readback",
		StoreMutationReadbackRef:       normalizeOneDisplaySafeRef(input.StoreMutationReadbackRef),
		StoreMutationResultRef:         mutationResult.StoreMutationResultRef,
		StoreMutationRequestRef:        mutationResult.StoreMutationRequestRef,
		StoreMutationDescriptorRef:     mutationResult.StoreMutationDescriptorRef,
		StoreAdapterRef:                mutationResult.StoreAdapterRef,
		HostStoreAdapterRunRef:         mutationResult.HostStoreAdapterRunRef,
		HostRunstoreRef:                mutationResult.HostRunstoreRef,
		HostObjectiveStoreRef:          mutationResult.HostObjectiveStoreRef,
		SupportsRunstoreMutation:       mutationResult.SupportsRunstoreMutation,
		SupportsObjectiveStoreMutation: mutationResult.SupportsObjectiveStoreMutation,
		SupportsTransactionReplay:      mutationResult.SupportsTransactionReplay,
		ExpectedTransactionRef:         mutationResult.ExpectedTransactionRef,
		ExpectedReadbackRef:            mutationResult.ExpectedReadbackRef,
		AppliedTransactionRef:          mutationResult.AppliedTransactionRef,
		AppliedRunstoreRef:             mutationResult.AppliedRunstoreRef,
		AppliedObjectiveStateRef:       mutationResult.AppliedObjectiveStateRef,
		ObservedTransactionRef:         normalizeOneDisplaySafeRef(input.ObservedTransactionRef),
		ObservedRunstoreRef:            normalizeOneDisplaySafeRef(input.ObservedRunstoreRef),
		ObservedObjectiveStateRef:      normalizeOneDisplaySafeRef(input.ObservedObjectiveStateRef),
		ReplayRef:                      normalizeOneDisplaySafeRef(input.ReplayRef),
		FailureClass:                   FailureNone,
		Boundaries:                     productionAdapterStoreMutationReadbackBoundaries(mutationResult.Boundaries),
		RunnerEffect:                   "none",
		PromptEffect:                   "none",
		RawOutputLoaded:                input.RawOutputLoaded || mutationResult.RawOutputLoaded,
	}
	if productionAdapterStoreMutationReadbackUnsafe(input, mutationResult) {
		result = productionAdapterStoreMutationReadbackBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if !mutationResult.ReadyForStoreMutationReadback {
		result = productionAdapterStoreMutationReadbackBlock(result, firstFailureClass(mutationResult.FailureClass, FailureEvidenceMissing), "store_mutation_result_not_ready", "host:store_mutation_result", firstNextHostAction(mutationResult.NextHostAction, "review_store_mutation_result"))
	}
	if result.StoreMutationReadbackRef == "" {
		result = productionAdapterStoreMutationReadbackBlock(result, FailureEvidenceMissing, "store_mutation_readback_ref_missing", "host:store_mutation_readback_ref", "provide_store_mutation_readback_ref")
	}
	if result.ReplayRef == "" {
		result = productionAdapterStoreMutationReadbackBlock(result, FailureEvidenceMissing, "store_mutation_replay_ref_missing", "host:store_mutation_replay_ref", "provide_store_mutation_replay_ref")
	}
	if result.ObservedTransactionRef == "" {
		result = productionAdapterStoreMutationReadbackBlock(result, FailureEvidenceMissing, "observed_transaction_ref_missing", "host:observed_store_transaction_ref", "provide_store_mutation_readback")
	} else if result.AppliedTransactionRef != "" && result.ObservedTransactionRef != result.AppliedTransactionRef {
		result = productionAdapterStoreMutationReadbackBlock(result, FailureVerificationFailed, "store_mutation_observed_transaction_ref_mismatch", "host:observed_store_transaction_ref", "review_store_mutation_readback")
	}
	if result.SupportsRunstoreMutation {
		if result.ObservedRunstoreRef == "" {
			result = productionAdapterStoreMutationReadbackBlock(result, FailureEvidenceMissing, "observed_runstore_ref_missing", "host:observed_runstore_ref", "provide_store_mutation_readback")
		} else if result.AppliedRunstoreRef != "" && result.ObservedRunstoreRef != result.AppliedRunstoreRef {
			result = productionAdapterStoreMutationReadbackBlock(result, FailureVerificationFailed, "store_mutation_observed_runstore_ref_mismatch", "host:observed_runstore_ref", "review_store_mutation_readback")
		}
	}
	if result.SupportsObjectiveStoreMutation {
		if result.ObservedObjectiveStateRef == "" {
			result = productionAdapterStoreMutationReadbackBlock(result, FailureEvidenceMissing, "observed_objective_state_ref_missing", "host:observed_objective_state_ref", "provide_store_mutation_readback")
		} else if result.AppliedObjectiveStateRef != "" && result.ObservedObjectiveStateRef != result.AppliedObjectiveStateRef {
			result = productionAdapterStoreMutationReadbackBlock(result, FailureVerificationFailed, "store_mutation_observed_objective_state_ref_mismatch", "host:observed_objective_state_ref", "review_store_mutation_readback")
		}
	}
	if len(result.MissingInputs) == 0 && len(result.BlockedReasons) == 0 {
		result.Status = HostActionRecorded
		result.StoreMutationReadbackBound = true
		result.TransactionReplayVerified = true
		result.ReadyForDownstreamReadback = true
		result.NextHostAction = "bind_downstream_readback"
		result.Boundaries = AppendBoundaries(result.Boundaries, "store_mutation_readback_bound", "transaction_replay_verified", "ready_for_downstream_readback")
	}
	return result.Normalize()
}

func CloneProductionAdapterStoreMutationDescriptor(in ProductionAdapterStoreMutationDescriptor) ProductionAdapterStoreMutationDescriptor {
	out := in
	out.SupportedMutationKinds = cloneStringSlice(in.SupportedMutationKinds)
	out.PolicyRefs = cloneDisplaySafeRefs(in.PolicyRefs)
	out.RequiredPolicyRefs = cloneDisplaySafeRefs(in.RequiredPolicyRefs)
	out.ApprovalRefs = cloneDisplaySafeRefs(in.ApprovalRefs)
	out.RequiredApprovalRefs = cloneDisplaySafeRefs(in.RequiredApprovalRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (d ProductionAdapterStoreMutationDescriptor) Clone() ProductionAdapterStoreMutationDescriptor {
	return CloneProductionAdapterStoreMutationDescriptor(d)
}

func (d ProductionAdapterStoreMutationDescriptor) Normalize() ProductionAdapterStoreMutationDescriptor {
	out := CloneProductionAdapterStoreMutationDescriptor(d)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.Mode = normalizeControlToken(out.Mode)
	if out.Mode == "" {
		out.Mode = "production_adapter_store_mutation_descriptor"
	}
	out.StoreMutationDescriptorRef = normalizeOneDisplaySafeRef(out.StoreMutationDescriptorRef)
	out.StoreAdapterRef = normalizeOneDisplaySafeRef(out.StoreAdapterRef)
	out.OwnerRef = normalizeOneDisplaySafeRef(out.OwnerRef)
	out.HostRunstoreRef = normalizeOneDisplaySafeRef(out.HostRunstoreRef)
	out.HostObjectiveStoreRef = normalizeOneDisplaySafeRef(out.HostObjectiveStoreRef)
	out.SupportedMutationKinds = normalizeStringList(out.SupportedMutationKinds)
	out.TransactionContractRef = normalizeOneDisplaySafeRef(out.TransactionContractRef)
	out.IdempotencyRef = normalizeOneDisplaySafeRef(out.IdempotencyRef)
	out.IdempotencyContractRef = normalizeOneDisplaySafeRef(out.IdempotencyContractRef)
	out.ReplayContractRef = normalizeOneDisplaySafeRef(out.ReplayContractRef)
	out.ReadbackContractRef = normalizeOneDisplaySafeRef(out.ReadbackContractRef)
	out.RedactionPolicyRef = normalizeOneDisplaySafeRef(out.RedactionPolicyRef)
	out.TimeoutPolicyRef = normalizeOneDisplaySafeRef(out.TimeoutPolicyRef)
	out.PolicyRefs = normalizeDisplaySafeRefs(out.PolicyRefs)
	out.RequiredPolicyRefs = normalizeDisplaySafeRefs(out.RequiredPolicyRefs)
	out.ApprovalRefs = normalizeDisplaySafeRefs(out.ApprovalRefs)
	out.RequiredApprovalRefs = normalizeDisplaySafeRefs(out.RequiredApprovalRefs)
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
	if !out.Available {
		out.Status = HostActionNotReady
		out.ReadyForStoreMutationRequest = false
	}
	if out.RawOutputLoaded || productionAdapterStoreMutationDescriptorOutputUnsafe(out) {
		out.RawOutputLoaded = true
		out.Status = HostActionBlocked
		out.ReadyForStoreMutationRequest = false
		out.FailureClass = firstFailureClass(out.FailureClass, FailureEvidenceWeak)
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		out.NextHostAction = "provide_display_safe_refs"
	}
	out.CoreInvocationExecuted = false
	out.DurableWriteByCore = false
	out.ObjectiveStoreWriteByCore = false
	out.RunstoreWriteByCore = false
	out.ReadyForStoreMutationRequest = out.ReadyForStoreMutationRequest &&
		out.Available &&
		out.Status == HostActionReady &&
		out.StoreMutationDescriptorRef != "" &&
		out.StoreAdapterRef != "" &&
		out.OwnerRef != "" &&
		(out.SupportsRunstoreMutation || out.SupportsObjectiveStoreMutation) &&
		(!out.SupportsRunstoreMutation || out.HostRunstoreRef != "") &&
		(!out.SupportsObjectiveStoreMutation || out.HostObjectiveStoreRef != "") &&
		out.SupportsTransactionReplay &&
		len(out.SupportedMutationKinds) > 0 &&
		out.TransactionContractRef != "" &&
		out.IdempotencyRef != "" &&
		out.IdempotencyContractRef != "" &&
		out.ReplayContractRef != "" &&
		out.ReadbackContractRef != "" &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.RawOutputLoaded
	return out
}

func CloneProductionAdapterStoreMutationRequest(in ProductionAdapterStoreMutationRequest) ProductionAdapterStoreMutationRequest {
	out := in
	out.SupportedMutationKinds = cloneStringSlice(in.SupportedMutationKinds)
	out.PolicyRefs = cloneDisplaySafeRefs(in.PolicyRefs)
	out.RequiredPolicyRefs = cloneDisplaySafeRefs(in.RequiredPolicyRefs)
	out.ApprovalRefs = cloneDisplaySafeRefs(in.ApprovalRefs)
	out.RequiredApprovalRefs = cloneDisplaySafeRefs(in.RequiredApprovalRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (r ProductionAdapterStoreMutationRequest) Clone() ProductionAdapterStoreMutationRequest {
	return CloneProductionAdapterStoreMutationRequest(r)
}

func (r ProductionAdapterStoreMutationRequest) Normalize() ProductionAdapterStoreMutationRequest {
	out := CloneProductionAdapterStoreMutationRequest(r)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.Mode = normalizeControlToken(out.Mode)
	if out.Mode == "" {
		out.Mode = "production_adapter_store_mutation_request"
	}
	out.StoreMutationRequestRef = normalizeOneDisplaySafeRef(out.StoreMutationRequestRef)
	out.StoreMutationDescriptorRef = normalizeOneDisplaySafeRef(out.StoreMutationDescriptorRef)
	out.StoreAdapterRef = normalizeOneDisplaySafeRef(out.StoreAdapterRef)
	out.OwnerRef = normalizeOneDisplaySafeRef(out.OwnerRef)
	out.HostRunstoreRef = normalizeOneDisplaySafeRef(out.HostRunstoreRef)
	out.HostObjectiveStoreRef = normalizeOneDisplaySafeRef(out.HostObjectiveStoreRef)
	out.SupportedMutationKinds = normalizeStringList(out.SupportedMutationKinds)
	out.SourceDurableResultRef = normalizeOneDisplaySafeRef(out.SourceDurableResultRef)
	out.SourceDurableEventRef = normalizeOneDisplaySafeRef(out.SourceDurableEventRef)
	out.SourceRunstoreRef = normalizeOneDisplaySafeRef(out.SourceRunstoreRef)
	out.SourceObjectiveStateRef = normalizeOneDisplaySafeRef(out.SourceObjectiveStateRef)
	out.HostStoreMutationConfirmationRef = normalizeOneDisplaySafeRef(out.HostStoreMutationConfirmationRef)
	out.ExpectedMutationResultRef = normalizeOneDisplaySafeRef(out.ExpectedMutationResultRef)
	out.ExpectedTransactionRef = normalizeOneDisplaySafeRef(out.ExpectedTransactionRef)
	out.ExpectedReadbackRef = normalizeOneDisplaySafeRef(out.ExpectedReadbackRef)
	out.TransactionContractRef = normalizeOneDisplaySafeRef(out.TransactionContractRef)
	out.IdempotencyRef = normalizeOneDisplaySafeRef(out.IdempotencyRef)
	out.IdempotencyContractRef = normalizeOneDisplaySafeRef(out.IdempotencyContractRef)
	out.ReplayContractRef = normalizeOneDisplaySafeRef(out.ReplayContractRef)
	out.ReadbackContractRef = normalizeOneDisplaySafeRef(out.ReadbackContractRef)
	out.RedactionPolicyRef = normalizeOneDisplaySafeRef(out.RedactionPolicyRef)
	out.TimeoutPolicyRef = normalizeOneDisplaySafeRef(out.TimeoutPolicyRef)
	out.PolicyRefs = normalizeDisplaySafeRefs(out.PolicyRefs)
	out.RequiredPolicyRefs = normalizeDisplaySafeRefs(out.RequiredPolicyRefs)
	out.ApprovalRefs = normalizeDisplaySafeRefs(out.ApprovalRefs)
	out.RequiredApprovalRefs = normalizeDisplaySafeRefs(out.RequiredApprovalRefs)
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
	if !out.Available {
		out.Status = HostActionNotReady
		out.ReadyForHostStoreMutation = false
		out.HostStoreMutationAuthorized = false
		out.HostMayExecuteStoreMutation = false
	}
	if out.RawOutputLoaded || productionAdapterStoreMutationRequestOutputUnsafe(out) {
		out.RawOutputLoaded = true
		out.Status = HostActionBlocked
		out.ReadyForHostStoreMutation = false
		out.HostStoreMutationAuthorized = false
		out.HostMayExecuteStoreMutation = false
		out.FailureClass = firstFailureClass(out.FailureClass, FailureEvidenceWeak)
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		if out.NextHostAction == "" {
			out.NextHostAction = "provide_display_safe_refs"
		}
	}
	out.CoreInvocationExecuted = false
	out.DurableWriteByCore = false
	out.ObjectiveStoreWriteByCore = false
	out.RunstoreWriteByCore = false
	out.ReadyForHostStoreMutation = out.ReadyForHostStoreMutation &&
		out.Status == HostActionReady &&
		out.StoreMutationRequestRef != "" &&
		out.StoreMutationDescriptorRef != "" &&
		out.StoreAdapterRef != "" &&
		out.SourceDurableResultRef != "" &&
		out.SourceDurableEventRef != "" &&
		out.SupportsTransactionReplay &&
		(!out.SupportsRunstoreMutation || (out.SourceRunstoreRef != "" && out.SourceRunstoreRef == out.HostRunstoreRef)) &&
		(!out.SupportsObjectiveStoreMutation || out.SourceObjectiveStateRef != "") &&
		out.HostStoreMutationConfirmationRef != "" &&
		out.ExpectedMutationResultRef != "" &&
		out.ExpectedTransactionRef != "" &&
		out.ExpectedReadbackRef != "" &&
		out.IdempotencyRef != "" &&
		out.IdempotencyContractRef != "" &&
		out.TransactionContractRef != "" &&
		out.ReplayContractRef != "" &&
		out.ReadbackContractRef != "" &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.RawOutputLoaded
	out.HostStoreMutationAuthorized = out.HostStoreMutationAuthorized && out.ReadyForHostStoreMutation
	out.HostMayExecuteStoreMutation = out.HostMayExecuteStoreMutation &&
		out.ReadyForHostStoreMutation &&
		!out.CoreInvocationExecuted &&
		!out.DurableWriteByCore &&
		!out.ObjectiveStoreWriteByCore &&
		!out.RunstoreWriteByCore
	return out
}

func CloneProductionAdapterStoreMutationResult(in ProductionAdapterStoreMutationResult) ProductionAdapterStoreMutationResult {
	out := in
	out.StoreMutationEvidenceRefs = cloneDisplaySafeRefs(in.StoreMutationEvidenceRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (r ProductionAdapterStoreMutationResult) Clone() ProductionAdapterStoreMutationResult {
	return CloneProductionAdapterStoreMutationResult(r)
}

func (r ProductionAdapterStoreMutationResult) Normalize() ProductionAdapterStoreMutationResult {
	out := CloneProductionAdapterStoreMutationResult(r)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.Mode = normalizeControlToken(out.Mode)
	if out.Mode == "" {
		out.Mode = "production_adapter_store_mutation_result"
	}
	out.StoreMutationResultRef = normalizeOneDisplaySafeRef(out.StoreMutationResultRef)
	out.ExpectedMutationResultRef = normalizeOneDisplaySafeRef(out.ExpectedMutationResultRef)
	out.StoreMutationRequestRef = normalizeOneDisplaySafeRef(out.StoreMutationRequestRef)
	out.StoreMutationDescriptorRef = normalizeOneDisplaySafeRef(out.StoreMutationDescriptorRef)
	out.StoreAdapterRef = normalizeOneDisplaySafeRef(out.StoreAdapterRef)
	out.HostStoreAdapterRunRef = normalizeOneDisplaySafeRef(out.HostStoreAdapterRunRef)
	out.HostRunstoreRef = normalizeOneDisplaySafeRef(out.HostRunstoreRef)
	out.HostObjectiveStoreRef = normalizeOneDisplaySafeRef(out.HostObjectiveStoreRef)
	out.SourceDurableResultRef = normalizeOneDisplaySafeRef(out.SourceDurableResultRef)
	out.SourceDurableEventRef = normalizeOneDisplaySafeRef(out.SourceDurableEventRef)
	out.SourceRunstoreRef = normalizeOneDisplaySafeRef(out.SourceRunstoreRef)
	out.SourceObjectiveStateRef = normalizeOneDisplaySafeRef(out.SourceObjectiveStateRef)
	out.ExpectedTransactionRef = normalizeOneDisplaySafeRef(out.ExpectedTransactionRef)
	out.ExpectedReadbackRef = normalizeOneDisplaySafeRef(out.ExpectedReadbackRef)
	out.AppliedTransactionRef = normalizeOneDisplaySafeRef(out.AppliedTransactionRef)
	out.AppliedRunstoreRef = normalizeOneDisplaySafeRef(out.AppliedRunstoreRef)
	out.AppliedObjectiveStateRef = normalizeOneDisplaySafeRef(out.AppliedObjectiveStateRef)
	out.FailureRef = normalizeOneDisplaySafeRef(out.FailureRef)
	out.CompensationRef = normalizeOneDisplaySafeRef(out.CompensationRef)
	out.TransactionContractRef = normalizeOneDisplaySafeRef(out.TransactionContractRef)
	out.IdempotencyRef = normalizeOneDisplaySafeRef(out.IdempotencyRef)
	out.IdempotencyContractRef = normalizeOneDisplaySafeRef(out.IdempotencyContractRef)
	out.ReplayContractRef = normalizeOneDisplaySafeRef(out.ReplayContractRef)
	out.ReadbackContractRef = normalizeOneDisplaySafeRef(out.ReadbackContractRef)
	out.StoreMutationEvidenceRefs = normalizeDisplaySafeRefs(out.StoreMutationEvidenceRefs)
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
	if !out.Available {
		out.Status = HostActionNotReady
		out.HostStoreMutationRecorded = false
		out.ReadyForStoreMutationReadback = false
	}
	if out.RawOutputLoaded || productionAdapterStoreMutationResultOutputUnsafe(out) {
		out.RawOutputLoaded = true
		out.Status = HostActionBlocked
		out.HostStoreMutationRecorded = false
		out.ReadyForStoreMutationReadback = false
		out.FailureClass = firstFailureClass(out.FailureClass, FailureEvidenceWeak)
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		if out.NextHostAction == "" {
			out.NextHostAction = "provide_display_safe_refs"
		}
	}
	out.CoreInvocationExecuted = false
	out.DurableWriteByCore = false
	out.ObjectiveStoreWriteByCore = false
	out.RunstoreWriteByCore = false
	out.HostStoreMutationRecorded = out.HostStoreMutationRecorded &&
		(out.Status == HostActionRecorded || out.Status == HostActionReviewRequired) &&
		out.HostStoreMutationReported &&
		out.StoreMutationRequestRef != "" &&
		out.StoreMutationResultRef != "" &&
		out.HostStoreAdapterRunRef != "" &&
		!out.RawOutputLoaded
	out.ReadyForStoreMutationReadback = out.ReadyForStoreMutationReadback &&
		out.Status == HostActionRecorded &&
		out.HostStoreMutationRecorded &&
		out.HostStoreMutationSucceeded &&
		!out.HostStoreMutationFailed &&
		out.SupportsTransactionReplay &&
		out.AppliedTransactionRef != "" &&
		out.AppliedTransactionRef == out.ExpectedTransactionRef &&
		(!out.SupportsRunstoreMutation || (out.AppliedRunstoreRef != "" && out.AppliedRunstoreRef == out.HostRunstoreRef)) &&
		(!out.SupportsObjectiveStoreMutation || (out.AppliedObjectiveStateRef != "" && out.AppliedObjectiveStateRef == out.SourceObjectiveStateRef)) &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.RawOutputLoaded
	return out
}

func CloneProductionAdapterStoreMutationReadback(in ProductionAdapterStoreMutationReadback) ProductionAdapterStoreMutationReadback {
	out := in
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (r ProductionAdapterStoreMutationReadback) Clone() ProductionAdapterStoreMutationReadback {
	return CloneProductionAdapterStoreMutationReadback(r)
}

func (r ProductionAdapterStoreMutationReadback) Normalize() ProductionAdapterStoreMutationReadback {
	out := CloneProductionAdapterStoreMutationReadback(r)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.Mode = normalizeControlToken(out.Mode)
	if out.Mode == "" {
		out.Mode = "production_adapter_store_mutation_readback"
	}
	out.StoreMutationReadbackRef = normalizeOneDisplaySafeRef(out.StoreMutationReadbackRef)
	out.StoreMutationResultRef = normalizeOneDisplaySafeRef(out.StoreMutationResultRef)
	out.StoreMutationRequestRef = normalizeOneDisplaySafeRef(out.StoreMutationRequestRef)
	out.StoreMutationDescriptorRef = normalizeOneDisplaySafeRef(out.StoreMutationDescriptorRef)
	out.StoreAdapterRef = normalizeOneDisplaySafeRef(out.StoreAdapterRef)
	out.HostStoreAdapterRunRef = normalizeOneDisplaySafeRef(out.HostStoreAdapterRunRef)
	out.HostRunstoreRef = normalizeOneDisplaySafeRef(out.HostRunstoreRef)
	out.HostObjectiveStoreRef = normalizeOneDisplaySafeRef(out.HostObjectiveStoreRef)
	out.ExpectedTransactionRef = normalizeOneDisplaySafeRef(out.ExpectedTransactionRef)
	out.ExpectedReadbackRef = normalizeOneDisplaySafeRef(out.ExpectedReadbackRef)
	out.AppliedTransactionRef = normalizeOneDisplaySafeRef(out.AppliedTransactionRef)
	out.AppliedRunstoreRef = normalizeOneDisplaySafeRef(out.AppliedRunstoreRef)
	out.AppliedObjectiveStateRef = normalizeOneDisplaySafeRef(out.AppliedObjectiveStateRef)
	out.ObservedTransactionRef = normalizeOneDisplaySafeRef(out.ObservedTransactionRef)
	out.ObservedRunstoreRef = normalizeOneDisplaySafeRef(out.ObservedRunstoreRef)
	out.ObservedObjectiveStateRef = normalizeOneDisplaySafeRef(out.ObservedObjectiveStateRef)
	out.ReplayRef = normalizeOneDisplaySafeRef(out.ReplayRef)
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
	if !out.Available {
		out.Status = HostActionNotReady
		out.StoreMutationReadbackBound = false
		out.TransactionReplayVerified = false
		out.ReadyForDownstreamReadback = false
	}
	if out.RawOutputLoaded || productionAdapterStoreMutationReadbackOutputUnsafe(out) {
		out.RawOutputLoaded = true
		out.Status = HostActionBlocked
		out.StoreMutationReadbackBound = false
		out.TransactionReplayVerified = false
		out.ReadyForDownstreamReadback = false
		out.FailureClass = firstFailureClass(out.FailureClass, FailureEvidenceWeak)
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		if out.NextHostAction == "" {
			out.NextHostAction = "provide_display_safe_refs"
		}
	}
	out.CoreInvocationExecuted = false
	out.DurableWriteByCore = false
	out.ObjectiveStoreWriteByCore = false
	out.RunstoreWriteByCore = false
	out.StoreMutationReadbackBound = out.StoreMutationReadbackBound &&
		out.Status == HostActionRecorded &&
		out.StoreMutationReadbackRef != "" &&
		out.StoreMutationResultRef != "" &&
		out.StoreMutationRequestRef != "" &&
		out.SupportsTransactionReplay &&
		out.ObservedTransactionRef != "" &&
		out.ObservedTransactionRef == out.AppliedTransactionRef &&
		(!out.SupportsRunstoreMutation || (out.ObservedRunstoreRef != "" && out.ObservedRunstoreRef == out.AppliedRunstoreRef)) &&
		(!out.SupportsObjectiveStoreMutation || (out.ObservedObjectiveStateRef != "" && out.ObservedObjectiveStateRef == out.AppliedObjectiveStateRef)) &&
		out.ReplayRef != "" &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.RawOutputLoaded
	out.TransactionReplayVerified = out.TransactionReplayVerified && out.StoreMutationReadbackBound
	out.ReadyForDownstreamReadback = out.ReadyForDownstreamReadback && out.StoreMutationReadbackBound
	return out
}

func productionAdapterStoreMutationDescriptorBlock(result ProductionAdapterStoreMutationDescriptor, failure FailureClass, reason string, missing MissingInput, next NextHostAction) ProductionAdapterStoreMutationDescriptor {
	result.Status = HostActionBlocked
	result.ReadyForStoreMutationRequest = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.Boundaries = AppendBoundaries(result.Boundaries, "store_mutation_descriptor_blocked")
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	return result
}

func productionAdapterStoreMutationRequestBlock(result ProductionAdapterStoreMutationRequest, failure FailureClass, reason string, missing MissingInput, next NextHostAction) ProductionAdapterStoreMutationRequest {
	result.Status = HostActionBlocked
	result.ReadyForHostStoreMutation = false
	result.HostStoreMutationAuthorized = false
	result.HostMayExecuteStoreMutation = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.Boundaries = AppendBoundaries(result.Boundaries, "store_mutation_request_blocked")
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	return result
}

func productionAdapterStoreMutationResultBlock(result ProductionAdapterStoreMutationResult, failure FailureClass, reason string, missing MissingInput, next NextHostAction) ProductionAdapterStoreMutationResult {
	result.Status = HostActionBlocked
	result.HostStoreMutationRecorded = false
	result.ReadyForStoreMutationReadback = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.Boundaries = AppendBoundaries(result.Boundaries, "store_mutation_result_blocked")
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	return result
}

func productionAdapterStoreMutationReadbackBlock(result ProductionAdapterStoreMutationReadback, failure FailureClass, reason string, missing MissingInput, next NextHostAction) ProductionAdapterStoreMutationReadback {
	result.Status = HostActionBlocked
	result.StoreMutationReadbackBound = false
	result.TransactionReplayVerified = false
	result.ReadyForDownstreamReadback = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.Boundaries = AppendBoundaries(result.Boundaries, "store_mutation_readback_blocked")
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	return result
}

func productionAdapterStoreMutationDescriptorBoundaries(groups ...[]Boundary) []Boundary {
	return MergeBoundaries(
		[]Boundary{
			"production_adapter_store_mutation_descriptor",
			"store_mutation_gate_projection_only",
			"host_owned_store_adapter",
			"display_safe_refs_only",
			"no_runner_dispatch",
			"no_durable_write_by_core",
			"no_objective_store_write_by_core",
			"no_runstore_write_by_core",
		},
		MergeBoundaries(groups...),
	)
}

func productionAdapterStoreMutationRequestBoundaries(groups ...[]Boundary) []Boundary {
	return MergeBoundaries(
		[]Boundary{
			"production_adapter_store_mutation_request",
			"store_mutation_request_projection_only",
			"host_owned_store_adapter",
			"explicit_store_mutation_confirmation_required",
			"idempotency_required",
			"transaction_replay_required",
			"display_safe_refs_only",
			"no_runner_dispatch",
			"no_durable_write_by_core",
			"no_objective_store_write_by_core",
			"no_runstore_write_by_core",
		},
		MergeBoundaries(groups...),
	)
}

func productionAdapterStoreMutationResultBoundaries(groups ...[]Boundary) []Boundary {
	return MergeBoundaries(
		[]Boundary{
			"production_adapter_store_mutation_result",
			"store_mutation_result_projection_only",
			"host_owned_store_adapter",
			"display_safe_refs_only",
			"no_runner_dispatch",
			"no_durable_write_by_core",
			"no_objective_store_write_by_core",
			"no_runstore_write_by_core",
		},
		MergeBoundaries(groups...),
	)
}

func productionAdapterStoreMutationReadbackBoundaries(groups ...[]Boundary) []Boundary {
	return MergeBoundaries(
		[]Boundary{
			"production_adapter_store_mutation_readback",
			"store_mutation_readback_projection_only",
			"host_owned_store_adapter",
			"transaction_replay_verified_by_host",
			"display_safe_refs_only",
			"no_runner_dispatch",
			"no_durable_write_by_core",
			"no_objective_store_write_by_core",
			"no_runstore_write_by_core",
		},
		MergeBoundaries(groups...),
	)
}

func productionAdapterStoreMutationDescriptorOutputUnsafe(input ProductionAdapterStoreMutationDescriptor) bool {
	return displaySafeRefRejected(input.StoreMutationDescriptorRef) ||
		displaySafeRefRejected(input.StoreAdapterRef) ||
		displaySafeRefRejected(input.OwnerRef) ||
		displaySafeRefRejected(input.HostRunstoreRef) ||
		displaySafeRefRejected(input.HostObjectiveStoreRef) ||
		displaySafeRefRejected(input.TransactionContractRef) ||
		displaySafeRefRejected(input.IdempotencyRef) ||
		displaySafeRefRejected(input.IdempotencyContractRef) ||
		displaySafeRefRejected(input.ReplayContractRef) ||
		displaySafeRefRejected(input.ReadbackContractRef) ||
		displaySafeRefRejected(input.RedactionPolicyRef) ||
		displaySafeRefRejected(input.TimeoutPolicyRef) ||
		displaySafeRefSliceRejected(input.PolicyRefs) ||
		displaySafeRefSliceRejected(input.RequiredPolicyRefs) ||
		displaySafeRefSliceRejected(input.ApprovalRefs) ||
		displaySafeRefSliceRejected(input.RequiredApprovalRefs) ||
		input.RawOutputLoaded
}

func productionAdapterStoreMutationRequestUnsafe(input ProductionAdapterStoreMutationRequestInput, descriptor ProductionAdapterStoreMutationDescriptor) bool {
	return input.RawOutputLoaded ||
		displaySafeRefRejected(input.StoreMutationRequestRef) ||
		displaySafeRefRejected(input.SourceDurableResultRef) ||
		displaySafeRefRejected(input.SourceDurableEventRef) ||
		displaySafeRefRejected(input.SourceRunstoreRef) ||
		displaySafeRefRejected(input.SourceObjectiveStateRef) ||
		displaySafeRefRejected(input.HostStoreMutationConfirmationRef) ||
		displaySafeRefRejected(input.ExpectedMutationResultRef) ||
		displaySafeRefRejected(input.ExpectedTransactionRef) ||
		displaySafeRefRejected(input.ExpectedReadbackRef) ||
		productionAdapterStoreMutationDescriptorOutputUnsafe(descriptor)
}

func productionAdapterStoreMutationRequestOutputUnsafe(input ProductionAdapterStoreMutationRequest) bool {
	return displaySafeRefRejected(input.StoreMutationRequestRef) ||
		displaySafeRefRejected(input.StoreMutationDescriptorRef) ||
		displaySafeRefRejected(input.StoreAdapterRef) ||
		displaySafeRefRejected(input.OwnerRef) ||
		displaySafeRefRejected(input.HostRunstoreRef) ||
		displaySafeRefRejected(input.HostObjectiveStoreRef) ||
		displaySafeRefRejected(input.SourceDurableResultRef) ||
		displaySafeRefRejected(input.SourceDurableEventRef) ||
		displaySafeRefRejected(input.SourceRunstoreRef) ||
		displaySafeRefRejected(input.SourceObjectiveStateRef) ||
		displaySafeRefRejected(input.HostStoreMutationConfirmationRef) ||
		displaySafeRefRejected(input.ExpectedMutationResultRef) ||
		displaySafeRefRejected(input.ExpectedTransactionRef) ||
		displaySafeRefRejected(input.ExpectedReadbackRef) ||
		displaySafeRefRejected(input.TransactionContractRef) ||
		displaySafeRefRejected(input.IdempotencyRef) ||
		displaySafeRefRejected(input.IdempotencyContractRef) ||
		displaySafeRefRejected(input.ReplayContractRef) ||
		displaySafeRefRejected(input.ReadbackContractRef) ||
		displaySafeRefRejected(input.RedactionPolicyRef) ||
		displaySafeRefRejected(input.TimeoutPolicyRef) ||
		displaySafeRefSliceRejected(input.PolicyRefs) ||
		displaySafeRefSliceRejected(input.RequiredPolicyRefs) ||
		displaySafeRefSliceRejected(input.ApprovalRefs) ||
		displaySafeRefSliceRejected(input.RequiredApprovalRefs) ||
		input.RawOutputLoaded
}

func productionAdapterStoreMutationResultUnsafe(input ProductionAdapterStoreMutationResultInput, request ProductionAdapterStoreMutationRequest) bool {
	return input.RawOutputLoaded ||
		displaySafeRefRejected(input.StoreMutationResultRef) ||
		displaySafeRefRejected(input.HostStoreAdapterRunRef) ||
		displaySafeRefRejected(input.AppliedTransactionRef) ||
		displaySafeRefRejected(input.AppliedRunstoreRef) ||
		displaySafeRefRejected(input.AppliedObjectiveStateRef) ||
		displaySafeRefRejected(input.FailureRef) ||
		displaySafeRefRejected(input.CompensationRef) ||
		displaySafeRefSliceRejected(input.StoreMutationEvidenceRefs) ||
		productionAdapterStoreMutationRequestOutputUnsafe(request)
}

func productionAdapterStoreMutationResultOutputUnsafe(input ProductionAdapterStoreMutationResult) bool {
	return displaySafeRefRejected(input.StoreMutationResultRef) ||
		displaySafeRefRejected(input.ExpectedMutationResultRef) ||
		displaySafeRefRejected(input.StoreMutationRequestRef) ||
		displaySafeRefRejected(input.StoreMutationDescriptorRef) ||
		displaySafeRefRejected(input.StoreAdapterRef) ||
		displaySafeRefRejected(input.HostStoreAdapterRunRef) ||
		displaySafeRefRejected(input.HostRunstoreRef) ||
		displaySafeRefRejected(input.HostObjectiveStoreRef) ||
		displaySafeRefRejected(input.SourceDurableResultRef) ||
		displaySafeRefRejected(input.SourceDurableEventRef) ||
		displaySafeRefRejected(input.SourceRunstoreRef) ||
		displaySafeRefRejected(input.SourceObjectiveStateRef) ||
		displaySafeRefRejected(input.ExpectedTransactionRef) ||
		displaySafeRefRejected(input.ExpectedReadbackRef) ||
		displaySafeRefRejected(input.AppliedTransactionRef) ||
		displaySafeRefRejected(input.AppliedRunstoreRef) ||
		displaySafeRefRejected(input.AppliedObjectiveStateRef) ||
		displaySafeRefRejected(input.FailureRef) ||
		displaySafeRefRejected(input.CompensationRef) ||
		displaySafeRefRejected(input.TransactionContractRef) ||
		displaySafeRefRejected(input.IdempotencyRef) ||
		displaySafeRefRejected(input.IdempotencyContractRef) ||
		displaySafeRefRejected(input.ReplayContractRef) ||
		displaySafeRefRejected(input.ReadbackContractRef) ||
		displaySafeRefSliceRejected(input.StoreMutationEvidenceRefs) ||
		input.RawOutputLoaded
}

func productionAdapterStoreMutationReadbackUnsafe(input ProductionAdapterStoreMutationReadbackInput, result ProductionAdapterStoreMutationResult) bool {
	return input.RawOutputLoaded ||
		displaySafeRefRejected(input.StoreMutationReadbackRef) ||
		displaySafeRefRejected(input.ObservedTransactionRef) ||
		displaySafeRefRejected(input.ObservedRunstoreRef) ||
		displaySafeRefRejected(input.ObservedObjectiveStateRef) ||
		displaySafeRefRejected(input.ReplayRef) ||
		productionAdapterStoreMutationResultOutputUnsafe(result)
}

func productionAdapterStoreMutationReadbackOutputUnsafe(input ProductionAdapterStoreMutationReadback) bool {
	return displaySafeRefRejected(input.StoreMutationReadbackRef) ||
		displaySafeRefRejected(input.StoreMutationResultRef) ||
		displaySafeRefRejected(input.StoreMutationRequestRef) ||
		displaySafeRefRejected(input.StoreMutationDescriptorRef) ||
		displaySafeRefRejected(input.StoreAdapterRef) ||
		displaySafeRefRejected(input.HostStoreAdapterRunRef) ||
		displaySafeRefRejected(input.HostRunstoreRef) ||
		displaySafeRefRejected(input.HostObjectiveStoreRef) ||
		displaySafeRefRejected(input.ExpectedTransactionRef) ||
		displaySafeRefRejected(input.ExpectedReadbackRef) ||
		displaySafeRefRejected(input.AppliedTransactionRef) ||
		displaySafeRefRejected(input.AppliedRunstoreRef) ||
		displaySafeRefRejected(input.AppliedObjectiveStateRef) ||
		displaySafeRefRejected(input.ObservedTransactionRef) ||
		displaySafeRefRejected(input.ObservedRunstoreRef) ||
		displaySafeRefRejected(input.ObservedObjectiveStateRef) ||
		displaySafeRefRejected(input.ReplayRef) ||
		input.RawOutputLoaded
}

func productionAdapterStoreMutationDescriptorEmpty(descriptor ProductionAdapterStoreMutationDescriptor) bool {
	return !descriptor.Projected &&
		!descriptor.Available &&
		descriptor.Status == "" &&
		descriptor.Mode == "" &&
		descriptor.StoreMutationDescriptorRef == "" &&
		descriptor.StoreAdapterRef == "" &&
		len(descriptor.MissingInputs) == 0 &&
		len(descriptor.BlockedReasons) == 0 &&
		len(descriptor.Boundaries) == 0 &&
		descriptor.NextHostAction == "" &&
		!descriptor.RawOutputLoaded
}

func productionAdapterStoreMutationRequestEmpty(request ProductionAdapterStoreMutationRequest) bool {
	return !request.Projected &&
		!request.Available &&
		request.Status == "" &&
		request.Mode == "" &&
		request.StoreMutationRequestRef == "" &&
		request.StoreMutationDescriptorRef == "" &&
		request.StoreAdapterRef == "" &&
		len(request.MissingInputs) == 0 &&
		len(request.BlockedReasons) == 0 &&
		len(request.Boundaries) == 0 &&
		request.NextHostAction == "" &&
		!request.RawOutputLoaded
}

func productionAdapterStoreMutationResultEmpty(result ProductionAdapterStoreMutationResult) bool {
	return !result.Projected &&
		!result.Available &&
		result.Status == "" &&
		result.Mode == "" &&
		result.StoreMutationResultRef == "" &&
		result.StoreMutationRequestRef == "" &&
		len(result.MissingInputs) == 0 &&
		len(result.BlockedReasons) == 0 &&
		len(result.Boundaries) == 0 &&
		result.NextHostAction == "" &&
		!result.RawOutputLoaded
}

func unavailableProductionAdapterStoreMutationRequest() ProductionAdapterStoreMutationRequest {
	return ProductionAdapterStoreMutationRequest{
		ContractVersion: ContractVersion,
		Projected:       true,
		Available:       false,
		Status:          HostActionNotReady,
		Mode:            "production_adapter_store_mutation_request",
		Boundaries: []Boundary{
			"production_adapter_store_mutation_request",
			"store_mutation_request_projection_only",
			"host_owned_store_adapter",
			"no_runner_dispatch",
			"no_durable_write_by_core",
			"no_objective_store_write_by_core",
			"no_runstore_write_by_core",
		},
		RunnerEffect:   "none",
		PromptEffect:   "none",
		NextHostAction: "provide_store_mutation_descriptor",
	}
}

func unavailableProductionAdapterStoreMutationResult() ProductionAdapterStoreMutationResult {
	return ProductionAdapterStoreMutationResult{
		ContractVersion: ContractVersion,
		Projected:       true,
		Available:       false,
		Status:          HostActionNotReady,
		Mode:            "production_adapter_store_mutation_result",
		Boundaries: []Boundary{
			"production_adapter_store_mutation_result",
			"store_mutation_result_projection_only",
			"host_owned_store_adapter",
			"no_runner_dispatch",
			"no_durable_write_by_core",
			"no_objective_store_write_by_core",
			"no_runstore_write_by_core",
		},
		RunnerEffect:   "none",
		PromptEffect:   "none",
		NextHostAction: "provide_store_mutation_request",
	}
}

func unavailableProductionAdapterStoreMutationReadback() ProductionAdapterStoreMutationReadback {
	return ProductionAdapterStoreMutationReadback{
		ContractVersion: ContractVersion,
		Projected:       true,
		Available:       false,
		Status:          HostActionNotReady,
		Mode:            "production_adapter_store_mutation_readback",
		Boundaries: []Boundary{
			"production_adapter_store_mutation_readback",
			"store_mutation_readback_projection_only",
			"host_owned_store_adapter",
			"no_runner_dispatch",
			"no_durable_write_by_core",
			"no_objective_store_write_by_core",
			"no_runstore_write_by_core",
		},
		RunnerEffect:   "none",
		PromptEffect:   "none",
		NextHostAction: "provide_store_mutation_result",
	}
}
