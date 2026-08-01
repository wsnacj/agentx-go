package controlcontract

type ProductionAdapterObjectiveCloseoutWriterDurableRequestInput struct {
	DurableRequestRef               DisplaySafeRef                                             `json:"durable_request_ref,omitempty"`
	WriterOptIn                     ProductionAdapterObjectiveCloseoutWriterOptIn              `json:"writer_opt_in,omitempty"`
	DryRunSmoke                     ProductionAdapterObjectiveCloseoutWriterDryRunSmokeHarness `json:"dry_run_smoke,omitempty"`
	HostDurableWriteConfirmationRef DisplaySafeRef                                             `json:"host_durable_write_confirmation_ref,omitempty"`
	ExpectedDurableResultRef        DisplaySafeRef                                             `json:"expected_durable_result_ref,omitempty"`
	RawOutputLoaded                 bool                                                       `json:"raw_output_loaded"`
}

type ProductionAdapterObjectiveCloseoutWriterDurableRequest struct {
	ContractVersion                 string                                       `json:"contract_version,omitempty"`
	Projected                       bool                                         `json:"projected"`
	Available                       bool                                         `json:"available"`
	Status                          HostActionStatus                             `json:"status,omitempty"`
	Mode                            string                                       `json:"mode,omitempty"`
	ReadyForHostDurableWrite        bool                                         `json:"ready_for_host_durable_write"`
	HostDurableWriteAuthorized      bool                                         `json:"host_durable_write_authorized"`
	HostMayExecuteDurableWrite      bool                                         `json:"host_may_execute_durable_write"`
	CoreInvocationExecuted          bool                                         `json:"core_invocation_executed"`
	DryRunByCore                    bool                                         `json:"dry_run_by_core"`
	DurableWriteByCore              bool                                         `json:"durable_write_by_core"`
	ObjectiveStoreWriteByCore       bool                                         `json:"objective_store_write_by_core"`
	RunstoreWriteByCore             bool                                         `json:"runstore_write_by_core"`
	RequestedMode                   ProductionAdapterObjectiveCloseoutWriterMode `json:"requested_mode,omitempty"`
	DurableRequestRef               DisplaySafeRef                               `json:"durable_request_ref,omitempty"`
	WriterOptInRef                  DisplaySafeRef                               `json:"writer_opt_in_ref,omitempty"`
	WriterRef                       DisplaySafeRef                               `json:"writer_ref,omitempty"`
	OwnerRef                        DisplaySafeRef                               `json:"owner_ref,omitempty"`
	HostWriterBindingRef            DisplaySafeRef                               `json:"host_writer_binding_ref,omitempty"`
	DryRunSmokeRef                  DisplaySafeRef                               `json:"dry_run_smoke_ref,omitempty"`
	DryRunResultRef                 DisplaySafeRef                               `json:"dry_run_result_ref,omitempty"`
	ExpectedDryRunResultRef         DisplaySafeRef                               `json:"expected_dry_run_result_ref,omitempty"`
	ExpectedDurableResultRef        DisplaySafeRef                               `json:"expected_durable_result_ref,omitempty"`
	ExpectedReadbackRef             DisplaySafeRef                               `json:"expected_readback_ref,omitempty"`
	ObjectiveCloseoutHandoffRef     DisplaySafeRef                               `json:"objective_closeout_handoff_ref,omitempty"`
	HostUIHandoffRef                DisplaySafeRef                               `json:"host_ui_handoff_ref,omitempty"`
	ObjectiveCloseoutPacketRef      DisplaySafeRef                               `json:"objective_closeout_packet_ref,omitempty"`
	ObjectiveRef                    DisplaySafeRef                               `json:"objective_ref,omitempty"`
	HostObjectiveLifecycleRef       DisplaySafeRef                               `json:"host_objective_lifecycle_ref,omitempty"`
	HostRunstoreRef                 DisplaySafeRef                               `json:"host_runstore_ref,omitempty"`
	ExpectedDurableEventRef         DisplaySafeRef                               `json:"expected_durable_event_ref,omitempty"`
	ExpectedObjectiveStateRef       DisplaySafeRef                               `json:"expected_objective_state_ref,omitempty"`
	HostDurableApplyConfirmationRef DisplaySafeRef                               `json:"host_durable_apply_confirmation_ref,omitempty"`
	HostDurableWriteConfirmationRef DisplaySafeRef                               `json:"host_durable_write_confirmation_ref,omitempty"`
	AvailableCapabilityRefs         []DisplaySafeRef                             `json:"available_capability_refs,omitempty"`
	RequiredCapabilityRefs          []DisplaySafeRef                             `json:"required_capability_refs,omitempty"`
	PolicyRefs                      []DisplaySafeRef                             `json:"policy_refs,omitempty"`
	RequiredPolicyRefs              []DisplaySafeRef                             `json:"required_policy_refs,omitempty"`
	ApprovalRefs                    []DisplaySafeRef                             `json:"approval_refs,omitempty"`
	RequiredApprovalRefs            []DisplaySafeRef                             `json:"required_approval_refs,omitempty"`
	BudgetRef                       DisplaySafeRef                               `json:"budget_ref,omitempty"`
	RequiredBudgetRef               DisplaySafeRef                               `json:"required_budget_ref,omitempty"`
	IdempotencyRef                  DisplaySafeRef                               `json:"idempotency_ref,omitempty"`
	IdempotencyContractRef          DisplaySafeRef                               `json:"idempotency_contract_ref,omitempty"`
	DryRunPlanRef                   DisplaySafeRef                               `json:"dry_run_plan_ref,omitempty"`
	DryRunContractRef               DisplaySafeRef                               `json:"dry_run_contract_ref,omitempty"`
	ReadbackContractRef             DisplaySafeRef                               `json:"readback_contract_ref,omitempty"`
	RollbackReviewRef               DisplaySafeRef                               `json:"rollback_review_ref,omitempty"`
	CompensationReviewRef           DisplaySafeRef                               `json:"compensation_review_ref,omitempty"`
	RedactionPolicyRef              DisplaySafeRef                               `json:"redaction_policy_ref,omitempty"`
	TimeoutPolicyRef                DisplaySafeRef                               `json:"timeout_policy_ref,omitempty"`
	MissingInputs                   []MissingInput                               `json:"missing_inputs,omitempty"`
	BlockedReasons                  []string                                     `json:"blocked_reasons,omitempty"`
	FailureClass                    FailureClass                                 `json:"failure_class,omitempty"`
	Boundaries                      []Boundary                                   `json:"boundaries,omitempty"`
	NextHostAction                  NextHostAction                               `json:"next_host_action,omitempty"`
	RunnerEffect                    string                                       `json:"runner_effect,omitempty"`
	PromptEffect                    string                                       `json:"prompt_effect,omitempty"`
	RawOutputLoaded                 bool                                         `json:"raw_output_loaded"`
}

type ProductionAdapterObjectiveCloseoutWriterDurableResultInput struct {
	DurableResultRef          DisplaySafeRef                                         `json:"durable_result_ref,omitempty"`
	DurableRequest            ProductionAdapterObjectiveCloseoutWriterDurableRequest `json:"durable_request,omitempty"`
	HostAdapterRunRef         DisplaySafeRef                                         `json:"host_adapter_run_ref,omitempty"`
	HostDurableWriteReported  bool                                                   `json:"host_durable_write_reported"`
	HostDurableWriteSucceeded bool                                                   `json:"host_durable_write_succeeded"`
	HostDurableWriteFailed    bool                                                   `json:"host_durable_write_failed"`
	AppliedDurableEventRef    DisplaySafeRef                                         `json:"applied_durable_event_ref,omitempty"`
	AppliedRunstoreRef        DisplaySafeRef                                         `json:"applied_runstore_ref,omitempty"`
	AppliedObjectiveStateRef  DisplaySafeRef                                         `json:"applied_objective_state_ref,omitempty"`
	FailureRef                DisplaySafeRef                                         `json:"failure_ref,omitempty"`
	CompensationRef           DisplaySafeRef                                         `json:"compensation_ref,omitempty"`
	DurableEvidenceRefs       []DisplaySafeRef                                       `json:"durable_evidence_refs,omitempty"`
	RawOutputLoaded           bool                                                   `json:"raw_output_loaded"`
}

type ProductionAdapterObjectiveCloseoutWriterDurableResult struct {
	ContractVersion               string           `json:"contract_version,omitempty"`
	Projected                     bool             `json:"projected"`
	Available                     bool             `json:"available"`
	Status                        HostActionStatus `json:"status,omitempty"`
	Mode                          string           `json:"mode,omitempty"`
	ReadyForWriterDurableReadback bool             `json:"ready_for_writer_durable_readback"`
	HostDurableWriteReported      bool             `json:"host_durable_write_reported"`
	HostDurableWriteSucceeded     bool             `json:"host_durable_write_succeeded"`
	HostDurableWriteFailed        bool             `json:"host_durable_write_failed"`
	HostDurableWriteRecorded      bool             `json:"host_durable_write_recorded"`
	CoreInvocationExecuted        bool             `json:"core_invocation_executed"`
	DryRunByCore                  bool             `json:"dry_run_by_core"`
	DurableWriteByCore            bool             `json:"durable_write_by_core"`
	ObjectiveStoreWriteByCore     bool             `json:"objective_store_write_by_core"`
	RunstoreWriteByCore           bool             `json:"runstore_write_by_core"`
	DurableResultRef              DisplaySafeRef   `json:"durable_result_ref,omitempty"`
	ExpectedDurableResultRef      DisplaySafeRef   `json:"expected_durable_result_ref,omitempty"`
	DurableRequestRef             DisplaySafeRef   `json:"durable_request_ref,omitempty"`
	WriterOptInRef                DisplaySafeRef   `json:"writer_opt_in_ref,omitempty"`
	WriterRef                     DisplaySafeRef   `json:"writer_ref,omitempty"`
	HostWriterBindingRef          DisplaySafeRef   `json:"host_writer_binding_ref,omitempty"`
	HostAdapterRunRef             DisplaySafeRef   `json:"host_adapter_run_ref,omitempty"`
	DryRunSmokeRef                DisplaySafeRef   `json:"dry_run_smoke_ref,omitempty"`
	DryRunResultRef               DisplaySafeRef   `json:"dry_run_result_ref,omitempty"`
	ExpectedReadbackRef           DisplaySafeRef   `json:"expected_readback_ref,omitempty"`
	ObjectiveCloseoutHandoffRef   DisplaySafeRef   `json:"objective_closeout_handoff_ref,omitempty"`
	ObjectiveCloseoutPacketRef    DisplaySafeRef   `json:"objective_closeout_packet_ref,omitempty"`
	ObjectiveRef                  DisplaySafeRef   `json:"objective_ref,omitempty"`
	HostRunstoreRef               DisplaySafeRef   `json:"host_runstore_ref,omitempty"`
	ExpectedDurableEventRef       DisplaySafeRef   `json:"expected_durable_event_ref,omitempty"`
	ExpectedObjectiveStateRef     DisplaySafeRef   `json:"expected_objective_state_ref,omitempty"`
	AppliedDurableEventRef        DisplaySafeRef   `json:"applied_durable_event_ref,omitempty"`
	AppliedRunstoreRef            DisplaySafeRef   `json:"applied_runstore_ref,omitempty"`
	AppliedObjectiveStateRef      DisplaySafeRef   `json:"applied_objective_state_ref,omitempty"`
	FailureRef                    DisplaySafeRef   `json:"failure_ref,omitempty"`
	CompensationRef               DisplaySafeRef   `json:"compensation_ref,omitempty"`
	IdempotencyRef                DisplaySafeRef   `json:"idempotency_ref,omitempty"`
	IdempotencyContractRef        DisplaySafeRef   `json:"idempotency_contract_ref,omitempty"`
	ReadbackContractRef           DisplaySafeRef   `json:"readback_contract_ref,omitempty"`
	RollbackReviewRef             DisplaySafeRef   `json:"rollback_review_ref,omitempty"`
	CompensationReviewRef         DisplaySafeRef   `json:"compensation_review_ref,omitempty"`
	DurableEvidenceRefs           []DisplaySafeRef `json:"durable_evidence_refs,omitempty"`
	MissingInputs                 []MissingInput   `json:"missing_inputs,omitempty"`
	BlockedReasons                []string         `json:"blocked_reasons,omitempty"`
	FailureClass                  FailureClass     `json:"failure_class,omitempty"`
	Boundaries                    []Boundary       `json:"boundaries,omitempty"`
	NextHostAction                NextHostAction   `json:"next_host_action,omitempty"`
	RunnerEffect                  string           `json:"runner_effect,omitempty"`
	PromptEffect                  string           `json:"prompt_effect,omitempty"`
	RawOutputLoaded               bool             `json:"raw_output_loaded"`
}

type ProductionAdapterObjectiveCloseoutWriterDurableReadbackInput struct {
	DurableReadbackRef        DisplaySafeRef                                        `json:"durable_readback_ref,omitempty"`
	DurableResult             ProductionAdapterObjectiveCloseoutWriterDurableResult `json:"durable_result,omitempty"`
	ObjectiveCloseoutReadback ProductionAdapterObjectiveCloseoutReadback            `json:"objective_closeout_readback,omitempty"`
	RawOutputLoaded           bool                                                  `json:"raw_output_loaded"`
}

type ProductionAdapterObjectiveCloseoutWriterDurableReadback struct {
	ContractVersion                    string           `json:"contract_version,omitempty"`
	Projected                          bool             `json:"projected"`
	Available                          bool             `json:"available"`
	Status                             HostActionStatus `json:"status,omitempty"`
	Mode                               string           `json:"mode,omitempty"`
	ReadyForObjectiveReturn            bool             `json:"ready_for_objective_return"`
	WriterDurableReadbackBound         bool             `json:"writer_durable_readback_bound"`
	ObjectiveLifecycleClosed           bool             `json:"objective_lifecycle_closed"`
	ObjectiveSatisfied                 bool             `json:"objective_satisfied"`
	CoreInvocationExecuted             bool             `json:"core_invocation_executed"`
	DryRunByCore                       bool             `json:"dry_run_by_core"`
	DurableWriteByCore                 bool             `json:"durable_write_by_core"`
	ObjectiveStoreWriteByCore          bool             `json:"objective_store_write_by_core"`
	RunstoreWriteByCore                bool             `json:"runstore_write_by_core"`
	DurableReadbackRef                 DisplaySafeRef   `json:"durable_readback_ref,omitempty"`
	DurableResultRef                   DisplaySafeRef   `json:"durable_result_ref,omitempty"`
	DurableRequestRef                  DisplaySafeRef   `json:"durable_request_ref,omitempty"`
	WriterOptInRef                     DisplaySafeRef   `json:"writer_opt_in_ref,omitempty"`
	WriterRef                          DisplaySafeRef   `json:"writer_ref,omitempty"`
	HostAdapterRunRef                  DisplaySafeRef   `json:"host_adapter_run_ref,omitempty"`
	DryRunSmokeRef                     DisplaySafeRef   `json:"dry_run_smoke_ref,omitempty"`
	DryRunResultRef                    DisplaySafeRef   `json:"dry_run_result_ref,omitempty"`
	ExpectedReadbackRef                DisplaySafeRef   `json:"expected_readback_ref,omitempty"`
	ObjectiveCloseoutReadbackRef       DisplaySafeRef   `json:"objective_closeout_readback_ref,omitempty"`
	ObjectiveCloseoutHandoffRef        DisplaySafeRef   `json:"objective_closeout_handoff_ref,omitempty"`
	ObjectiveCloseoutPacketRef         DisplaySafeRef   `json:"objective_closeout_packet_ref,omitempty"`
	ObjectiveRef                       DisplaySafeRef   `json:"objective_ref,omitempty"`
	HostRunstoreRef                    DisplaySafeRef   `json:"host_runstore_ref,omitempty"`
	ExpectedDurableEventRef            DisplaySafeRef   `json:"expected_durable_event_ref,omitempty"`
	ExpectedObjectiveStateRef          DisplaySafeRef   `json:"expected_objective_state_ref,omitempty"`
	AppliedDurableEventRef             DisplaySafeRef   `json:"applied_durable_event_ref,omitempty"`
	AppliedRunstoreRef                 DisplaySafeRef   `json:"applied_runstore_ref,omitempty"`
	AppliedObjectiveStateRef           DisplaySafeRef   `json:"applied_objective_state_ref,omitempty"`
	ObservedObjectiveCloseoutPacketRef DisplaySafeRef   `json:"observed_objective_closeout_packet_ref,omitempty"`
	ObservedObjectiveRef               DisplaySafeRef   `json:"observed_objective_ref,omitempty"`
	MissingInputs                      []MissingInput   `json:"missing_inputs,omitempty"`
	BlockedReasons                     []string         `json:"blocked_reasons,omitempty"`
	FailureClass                       FailureClass     `json:"failure_class,omitempty"`
	Boundaries                         []Boundary       `json:"boundaries,omitempty"`
	NextHostAction                     NextHostAction   `json:"next_host_action,omitempty"`
	RunnerEffect                       string           `json:"runner_effect,omitempty"`
	PromptEffect                       string           `json:"prompt_effect,omitempty"`
	RawOutputLoaded                    bool             `json:"raw_output_loaded"`
}

func BuildProductionAdapterObjectiveCloseoutWriterDurableRequest(input ProductionAdapterObjectiveCloseoutWriterDurableRequestInput) ProductionAdapterObjectiveCloseoutWriterDurableRequest {
	if productionAdapterObjectiveCloseoutWriterOptInEmpty(input.WriterOptIn) || productionAdapterObjectiveCloseoutWriterDryRunSmokeHarnessEmpty(input.DryRunSmoke) {
		return unavailableProductionAdapterObjectiveCloseoutWriterDurableRequest()
	}
	optIn := input.WriterOptIn.Normalize()
	smoke := input.DryRunSmoke.Normalize()
	result := ProductionAdapterObjectiveCloseoutWriterDurableRequest{
		ContractVersion:                 ContractVersion,
		Projected:                       true,
		Available:                       optIn.Available && smoke.Available,
		Status:                          HostActionBlocked,
		Mode:                            "production_adapter_objective_closeout_writer_durable_request",
		RequestedMode:                   optIn.RequestedMode,
		DurableRequestRef:               normalizeOneDisplaySafeRef(input.DurableRequestRef),
		WriterOptInRef:                  optIn.WriterOptInRef,
		WriterRef:                       optIn.WriterRef,
		OwnerRef:                        optIn.OwnerRef,
		HostWriterBindingRef:            optIn.HostWriterBindingRef,
		DryRunSmokeRef:                  smoke.SmokeRef,
		DryRunResultRef:                 smoke.DryRunResultRef,
		ExpectedDryRunResultRef:         smoke.ExpectedDryRunResultRef,
		ExpectedDurableResultRef:        normalizeOneDisplaySafeRef(input.ExpectedDurableResultRef),
		ExpectedReadbackRef:             optIn.ExpectedReadbackRef,
		ObjectiveCloseoutHandoffRef:     optIn.ObjectiveCloseoutHandoffRef,
		HostUIHandoffRef:                optIn.HostUIHandoffRef,
		ObjectiveCloseoutPacketRef:      optIn.ObjectiveCloseoutPacketRef,
		ObjectiveRef:                    optIn.ObjectiveRef,
		HostObjectiveLifecycleRef:       optIn.HostObjectiveLifecycleRef,
		HostRunstoreRef:                 optIn.HostRunstoreRef,
		ExpectedDurableEventRef:         optIn.ExpectedDurableEventRef,
		ExpectedObjectiveStateRef:       optIn.ExpectedObjectiveStateRef,
		HostDurableApplyConfirmationRef: optIn.HostDurableApplyConfirmationRef,
		HostDurableWriteConfirmationRef: normalizeOneDisplaySafeRef(input.HostDurableWriteConfirmationRef),
		AvailableCapabilityRefs:         cloneDisplaySafeRefs(optIn.AvailableCapabilityRefs),
		RequiredCapabilityRefs:          cloneDisplaySafeRefs(optIn.RequiredCapabilityRefs),
		PolicyRefs:                      cloneDisplaySafeRefs(optIn.PolicyRefs),
		RequiredPolicyRefs:              cloneDisplaySafeRefs(optIn.RequiredPolicyRefs),
		ApprovalRefs:                    cloneDisplaySafeRefs(optIn.ApprovalRefs),
		RequiredApprovalRefs:            cloneDisplaySafeRefs(optIn.RequiredApprovalRefs),
		BudgetRef:                       optIn.BudgetRef,
		RequiredBudgetRef:               optIn.RequiredBudgetRef,
		IdempotencyRef:                  optIn.IdempotencyRef,
		IdempotencyContractRef:          optIn.IdempotencyContractRef,
		DryRunPlanRef:                   optIn.DryRunPlanRef,
		DryRunContractRef:               optIn.DryRunContractRef,
		ReadbackContractRef:             optIn.ReadbackContractRef,
		RollbackReviewRef:               optIn.RollbackReviewRef,
		CompensationReviewRef:           optIn.CompensationReviewRef,
		RedactionPolicyRef:              optIn.RedactionPolicyRef,
		TimeoutPolicyRef:                optIn.TimeoutPolicyRef,
		FailureClass:                    FailureNone,
		Boundaries:                      productionAdapterObjectiveCloseoutWriterDurableRequestBoundaries(optIn.Boundaries, smoke.Boundaries),
		RunnerEffect:                    "none",
		PromptEffect:                    "none",
		RawOutputLoaded:                 input.RawOutputLoaded || optIn.RawOutputLoaded || smoke.RawOutputLoaded,
	}
	if productionAdapterObjectiveCloseoutWriterDurableRequestUnsafe(input, optIn, smoke) {
		result = productionAdapterObjectiveCloseoutWriterDurableRequestBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if !optIn.ReadyForHostDurableWrite || !optIn.HostMayExecuteDurableWrite {
		result = productionAdapterObjectiveCloseoutWriterDurableRequestBlock(result, firstFailureClass(optIn.FailureClass, FailureConfigMissing), "objective_closeout_writer_durable_opt_in_not_ready", "host:objective_closeout_writer_durable_opt_in", firstNextHostAction(optIn.NextHostAction, "review_objective_closeout_writer_durable_opt_in"))
	}
	if !smoke.SmokePassed || !smoke.ReadyForDurableWriteOptIn {
		result = productionAdapterObjectiveCloseoutWriterDurableRequestBlock(result, firstFailureClass(smoke.FailureClass, FailureVerificationFailed), "objective_closeout_writer_dry_run_smoke_not_passed", "host:objective_closeout_writer_dry_run_smoke", firstNextHostAction(smoke.NextHostAction, "review_objective_closeout_writer_dry_run_smoke"))
	}
	for _, mismatch := range productionAdapterObjectiveCloseoutWriterDurableRequestMismatches(optIn, smoke) {
		result = productionAdapterObjectiveCloseoutWriterDurableRequestBlock(result, FailureVerificationFailed, mismatch.reason, mismatch.missing, "review_objective_closeout_writer_durable_request")
	}
	if result.DurableRequestRef == "" {
		result = productionAdapterObjectiveCloseoutWriterDurableRequestBlock(result, FailureEvidenceMissing, "writer_durable_request_ref_missing", "host:objective_closeout_writer_durable_request_ref", "provide_objective_closeout_writer_durable_request")
	}
	if result.HostDurableWriteConfirmationRef == "" {
		result = productionAdapterObjectiveCloseoutWriterDurableRequestBlock(result, FailureAuthorizationMissing, "host_durable_write_confirmation_ref_missing", "host:objective_closeout_writer_durable_write_confirmation_ref", "request_objective_closeout_writer_durable_write_confirmation")
	}
	if result.ExpectedDurableResultRef == "" {
		result = productionAdapterObjectiveCloseoutWriterDurableRequestBlock(result, FailureEvidenceMissing, "expected_durable_result_ref_missing", "host:objective_closeout_writer_expected_durable_result_ref", "provide_objective_closeout_writer_expected_durable_result")
	}
	if result.ExpectedDurableEventRef == "" {
		result = productionAdapterObjectiveCloseoutWriterDurableRequestBlock(result, FailureEvidenceMissing, "expected_durable_event_ref_missing", "host:expected_durable_event_ref", "provide_expected_durable_event_ref")
	}
	if result.ExpectedObjectiveStateRef == "" {
		result = productionAdapterObjectiveCloseoutWriterDurableRequestBlock(result, FailureEvidenceMissing, "expected_objective_state_ref_missing", "host:expected_objective_state_ref", "provide_expected_objective_state_ref")
	}
	if result.ExpectedReadbackRef == "" {
		result = productionAdapterObjectiveCloseoutWriterDurableRequestBlock(result, FailureEvidenceMissing, "expected_readback_ref_missing", "host:objective_closeout_writer_expected_readback_ref", "provide_objective_closeout_writer_expected_readback")
	}
	if len(result.BlockedReasons) == 0 && len(result.MissingInputs) == 0 {
		result.Status = HostActionReady
		result.ReadyForHostDurableWrite = true
		result.HostDurableWriteAuthorized = true
		result.HostMayExecuteDurableWrite = true
		result.NextHostAction = "host_may_execute_objective_closeout_durable_writer_adapter"
		result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_host_objective_closeout_writer_durable_write", "host_may_execute_durable_writer", "core_durable_write_not_executed")
	}
	return result.Normalize()
}

func BuildProductionAdapterObjectiveCloseoutWriterDurableResult(input ProductionAdapterObjectiveCloseoutWriterDurableResultInput) ProductionAdapterObjectiveCloseoutWriterDurableResult {
	if productionAdapterObjectiveCloseoutWriterDurableRequestEmpty(input.DurableRequest) {
		return unavailableProductionAdapterObjectiveCloseoutWriterDurableResult()
	}
	request := input.DurableRequest.Normalize()
	result := ProductionAdapterObjectiveCloseoutWriterDurableResult{
		ContractVersion:             ContractVersion,
		Projected:                   true,
		Available:                   request.Available,
		Status:                      HostActionBlocked,
		Mode:                        "production_adapter_objective_closeout_writer_durable_result",
		HostDurableWriteReported:    input.HostDurableWriteReported,
		HostDurableWriteSucceeded:   input.HostDurableWriteSucceeded,
		HostDurableWriteFailed:      input.HostDurableWriteFailed,
		DurableResultRef:            normalizeOneDisplaySafeRef(input.DurableResultRef),
		ExpectedDurableResultRef:    request.ExpectedDurableResultRef,
		DurableRequestRef:           request.DurableRequestRef,
		WriterOptInRef:              request.WriterOptInRef,
		WriterRef:                   request.WriterRef,
		HostWriterBindingRef:        request.HostWriterBindingRef,
		HostAdapterRunRef:           normalizeOneDisplaySafeRef(input.HostAdapterRunRef),
		DryRunSmokeRef:              request.DryRunSmokeRef,
		DryRunResultRef:             request.DryRunResultRef,
		ExpectedReadbackRef:         request.ExpectedReadbackRef,
		ObjectiveCloseoutHandoffRef: request.ObjectiveCloseoutHandoffRef,
		ObjectiveCloseoutPacketRef:  request.ObjectiveCloseoutPacketRef,
		ObjectiveRef:                request.ObjectiveRef,
		HostRunstoreRef:             request.HostRunstoreRef,
		ExpectedDurableEventRef:     request.ExpectedDurableEventRef,
		ExpectedObjectiveStateRef:   request.ExpectedObjectiveStateRef,
		AppliedDurableEventRef:      normalizeOneDisplaySafeRef(input.AppliedDurableEventRef),
		AppliedRunstoreRef:          normalizeOneDisplaySafeRef(input.AppliedRunstoreRef),
		AppliedObjectiveStateRef:    normalizeOneDisplaySafeRef(input.AppliedObjectiveStateRef),
		FailureRef:                  normalizeOneDisplaySafeRef(input.FailureRef),
		CompensationRef:             normalizeOneDisplaySafeRef(input.CompensationRef),
		IdempotencyRef:              request.IdempotencyRef,
		IdempotencyContractRef:      request.IdempotencyContractRef,
		ReadbackContractRef:         request.ReadbackContractRef,
		RollbackReviewRef:           request.RollbackReviewRef,
		CompensationReviewRef:       request.CompensationReviewRef,
		DurableEvidenceRefs:         normalizeDisplaySafeRefs(input.DurableEvidenceRefs),
		FailureClass:                FailureNone,
		Boundaries:                  productionAdapterObjectiveCloseoutWriterDurableResultBoundaries(request.Boundaries),
		RunnerEffect:                "none",
		PromptEffect:                "none",
		RawOutputLoaded:             input.RawOutputLoaded || request.RawOutputLoaded,
	}
	if productionAdapterObjectiveCloseoutWriterDurableResultUnsafe(input, request) {
		result = productionAdapterObjectiveCloseoutWriterDurableResultBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if !request.ReadyForHostDurableWrite || !request.HostMayExecuteDurableWrite {
		result = productionAdapterObjectiveCloseoutWriterDurableResultBlock(result, firstFailureClass(request.FailureClass, FailureConfigMissing), "objective_closeout_writer_durable_request_not_ready", "host:objective_closeout_writer_durable_request", firstNextHostAction(request.NextHostAction, "review_objective_closeout_writer_durable_request"))
	}
	if !input.HostDurableWriteReported {
		result = productionAdapterObjectiveCloseoutWriterDurableResultBlock(result, FailureEvidenceMissing, "writer_durable_write_not_reported", "host:objective_closeout_writer_durable_report", "provide_objective_closeout_writer_durable_report")
	}
	if input.HostDurableWriteSucceeded && input.HostDurableWriteFailed {
		result = productionAdapterObjectiveCloseoutWriterDurableResultBlock(result, FailureVerificationFailed, "writer_durable_write_status_conflict", "host:objective_closeout_writer_durable_status", "review_objective_closeout_writer_durable_result")
	}
	if result.HostAdapterRunRef == "" {
		result = productionAdapterObjectiveCloseoutWriterDurableResultBlock(result, FailureEvidenceMissing, "host_adapter_run_ref_missing", "host:objective_closeout_writer_host_adapter_run_ref", "provide_objective_closeout_writer_durable_report")
	}
	if result.DurableResultRef == "" {
		result = productionAdapterObjectiveCloseoutWriterDurableResultBlock(result, FailureEvidenceMissing, "durable_result_ref_missing", "host:objective_closeout_writer_durable_result_ref", "provide_objective_closeout_writer_durable_result")
	} else if request.ExpectedDurableResultRef != "" && result.DurableResultRef != request.ExpectedDurableResultRef {
		result = productionAdapterObjectiveCloseoutWriterDurableResultBlock(result, FailureVerificationFailed, "durable_result_ref_mismatch", "host:objective_closeout_writer_durable_result_ref", "review_objective_closeout_writer_durable_result")
	}
	if len(result.MissingInputs) > 0 || len(result.BlockedReasons) > 0 {
		return result.Normalize()
	}
	if input.HostDurableWriteFailed {
		if result.FailureRef == "" {
			result = productionAdapterObjectiveCloseoutWriterDurableResultBlock(result, FailureEvidenceMissing, "writer_durable_failure_ref_missing", "host:objective_closeout_writer_durable_failure_ref", "provide_objective_closeout_writer_durable_failure")
			return result.Normalize()
		}
		result.HostDurableWriteRecorded = true
		result.Status = HostActionReviewRequired
		result.FailureClass = FailureVerificationFailed
		result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, "objective_closeout_writer_durable_write_failed")
		result.NextHostAction = "review_objective_closeout_writer_durable_failure"
		result.Boundaries = AppendBoundaries(result.Boundaries, "objective_closeout_writer_durable_write_failed", "compensation_not_executed")
		return result.Normalize()
	}
	if !input.HostDurableWriteSucceeded {
		result = productionAdapterObjectiveCloseoutWriterDurableResultBlock(result, FailureEvidenceMissing, "writer_durable_write_status_missing", "host:objective_closeout_writer_durable_status", "provide_objective_closeout_writer_durable_report")
		return result.Normalize()
	}
	for _, mismatch := range productionAdapterObjectiveCloseoutWriterDurableResultMismatches(result) {
		result = productionAdapterObjectiveCloseoutWriterDurableResultBlock(result, FailureVerificationFailed, mismatch.reason, mismatch.missing, "review_objective_closeout_writer_durable_result")
	}
	if result.AppliedDurableEventRef == "" {
		result = productionAdapterObjectiveCloseoutWriterDurableResultBlock(result, FailureEvidenceMissing, "applied_durable_event_ref_missing", "host:applied_durable_event_ref", "provide_objective_closeout_writer_durable_result")
	}
	if result.AppliedRunstoreRef == "" {
		result = productionAdapterObjectiveCloseoutWriterDurableResultBlock(result, FailureEvidenceMissing, "applied_runstore_ref_missing", "host:applied_runstore_ref", "provide_objective_closeout_writer_durable_result")
	}
	if result.AppliedObjectiveStateRef == "" {
		result = productionAdapterObjectiveCloseoutWriterDurableResultBlock(result, FailureEvidenceMissing, "applied_objective_state_ref_missing", "host:applied_objective_state_ref", "provide_objective_closeout_writer_durable_result")
	}
	if len(result.MissingInputs) == 0 && len(result.BlockedReasons) == 0 {
		result.Status = HostActionRecorded
		result.HostDurableWriteRecorded = true
		result.ReadyForWriterDurableReadback = true
		result.NextHostAction = "bind_objective_closeout_writer_durable_readback"
		result.Boundaries = AppendBoundaries(result.Boundaries, "host_objective_closeout_writer_durable_write_recorded", "ready_for_objective_closeout_writer_durable_readback")
	}
	return result.Normalize()
}

func BuildProductionAdapterObjectiveCloseoutWriterDurableReadback(input ProductionAdapterObjectiveCloseoutWriterDurableReadbackInput) ProductionAdapterObjectiveCloseoutWriterDurableReadback {
	if productionAdapterObjectiveCloseoutWriterDurableResultEmpty(input.DurableResult) || productionAdapterObjectiveCloseoutReadbackEmpty(input.ObjectiveCloseoutReadback) {
		return unavailableProductionAdapterObjectiveCloseoutWriterDurableReadback()
	}
	resultProjection := input.DurableResult.Normalize()
	readback := input.ObjectiveCloseoutReadback.Normalize()
	result := ProductionAdapterObjectiveCloseoutWriterDurableReadback{
		ContractVersion:                    ContractVersion,
		Projected:                          true,
		Available:                          resultProjection.Available && readback.Available,
		Status:                             HostActionBlocked,
		Mode:                               "production_adapter_objective_closeout_writer_durable_readback",
		DurableReadbackRef:                 normalizeOneDisplaySafeRef(input.DurableReadbackRef),
		DurableResultRef:                   resultProjection.DurableResultRef,
		DurableRequestRef:                  resultProjection.DurableRequestRef,
		WriterOptInRef:                     resultProjection.WriterOptInRef,
		WriterRef:                          resultProjection.WriterRef,
		HostAdapterRunRef:                  resultProjection.HostAdapterRunRef,
		DryRunSmokeRef:                     resultProjection.DryRunSmokeRef,
		DryRunResultRef:                    resultProjection.DryRunResultRef,
		ExpectedReadbackRef:                resultProjection.ExpectedReadbackRef,
		ObjectiveCloseoutReadbackRef:       readback.ObjectiveCloseoutReadbackRef,
		ObjectiveCloseoutHandoffRef:        resultProjection.ObjectiveCloseoutHandoffRef,
		ObjectiveCloseoutPacketRef:         resultProjection.ObjectiveCloseoutPacketRef,
		ObjectiveRef:                       resultProjection.ObjectiveRef,
		HostRunstoreRef:                    resultProjection.HostRunstoreRef,
		ExpectedDurableEventRef:            resultProjection.ExpectedDurableEventRef,
		ExpectedObjectiveStateRef:          resultProjection.ExpectedObjectiveStateRef,
		AppliedDurableEventRef:             resultProjection.AppliedDurableEventRef,
		AppliedRunstoreRef:                 resultProjection.AppliedRunstoreRef,
		AppliedObjectiveStateRef:           resultProjection.AppliedObjectiveStateRef,
		ObservedObjectiveCloseoutPacketRef: readback.ObjectiveCloseoutPacketRef,
		ObservedObjectiveRef:               readback.ObjectiveRef,
		ObjectiveLifecycleClosed:           readback.ObjectiveLifecycleClosed,
		ObjectiveSatisfied:                 readback.ObjectiveSatisfied,
		FailureClass:                       FailureNone,
		Boundaries:                         productionAdapterObjectiveCloseoutWriterDurableReadbackBoundaries(resultProjection.Boundaries, readback.Boundaries),
		NextHostAction:                     firstNextHostAction(readback.NextHostAction, resultProjection.NextHostAction),
		RunnerEffect:                       "none",
		PromptEffect:                       "none",
		RawOutputLoaded:                    input.RawOutputLoaded || resultProjection.RawOutputLoaded || readback.RawOutputLoaded,
	}
	if productionAdapterObjectiveCloseoutWriterDurableReadbackUnsafe(input, resultProjection, readback) {
		result = productionAdapterObjectiveCloseoutWriterDurableReadbackBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if result.DurableReadbackRef == "" {
		result = productionAdapterObjectiveCloseoutWriterDurableReadbackBlock(result, FailureEvidenceMissing, "writer_durable_readback_ref_missing", "host:objective_closeout_writer_durable_readback_ref", "provide_objective_closeout_writer_durable_readback")
	}
	if !resultProjection.ReadyForWriterDurableReadback {
		result = productionAdapterObjectiveCloseoutWriterDurableReadbackBlock(result, firstFailureClass(resultProjection.FailureClass, FailureEvidenceMissing), "objective_closeout_writer_durable_result_not_ready", "host:objective_closeout_writer_durable_result", firstNextHostAction(resultProjection.NextHostAction, "review_objective_closeout_writer_durable_result"))
	}
	if !readback.ReadyForObjectiveCloseoutReadback {
		result = productionAdapterObjectiveCloseoutWriterDurableReadbackBlock(result, firstFailureClass(readback.FailureClass, FailureEvidenceMissing), "objective_closeout_readback_not_ready", "host:objective_closeout_readback", firstNextHostAction(readback.NextHostAction, "provide_objective_closeout_readback"))
	}
	for _, mismatch := range productionAdapterObjectiveCloseoutWriterDurableReadbackMismatches(resultProjection, readback) {
		result = productionAdapterObjectiveCloseoutWriterDurableReadbackBlock(result, FailureVerificationFailed, mismatch.reason, mismatch.missing, "review_objective_closeout_writer_durable_readback")
	}
	if len(result.BlockedReasons) == 0 && len(result.MissingInputs) == 0 {
		result.Status = HostActionRecorded
		result.WriterDurableReadbackBound = true
		result.ReadyForObjectiveReturn = true
		result.NextHostAction = "return_objective_closed_lifecycle"
		result.Boundaries = AppendBoundaries(result.Boundaries, "objective_closeout_writer_durable_readback_bound", "ready_for_objective_return", "objective_lifecycle_closed_by_host")
	}
	return result.Normalize()
}

func CloneProductionAdapterObjectiveCloseoutWriterDurableRequest(in ProductionAdapterObjectiveCloseoutWriterDurableRequest) ProductionAdapterObjectiveCloseoutWriterDurableRequest {
	out := in
	out.AvailableCapabilityRefs = cloneDisplaySafeRefs(in.AvailableCapabilityRefs)
	out.RequiredCapabilityRefs = cloneDisplaySafeRefs(in.RequiredCapabilityRefs)
	out.PolicyRefs = cloneDisplaySafeRefs(in.PolicyRefs)
	out.RequiredPolicyRefs = cloneDisplaySafeRefs(in.RequiredPolicyRefs)
	out.ApprovalRefs = cloneDisplaySafeRefs(in.ApprovalRefs)
	out.RequiredApprovalRefs = cloneDisplaySafeRefs(in.RequiredApprovalRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (r ProductionAdapterObjectiveCloseoutWriterDurableRequest) Clone() ProductionAdapterObjectiveCloseoutWriterDurableRequest {
	return CloneProductionAdapterObjectiveCloseoutWriterDurableRequest(r)
}

func (r ProductionAdapterObjectiveCloseoutWriterDurableRequest) Normalize() ProductionAdapterObjectiveCloseoutWriterDurableRequest {
	out := CloneProductionAdapterObjectiveCloseoutWriterDurableRequest(r)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.Mode = normalizeControlToken(out.Mode)
	if out.Mode == "" {
		out.Mode = "production_adapter_objective_closeout_writer_durable_request"
	}
	out.RequestedMode = NormalizeProductionAdapterObjectiveCloseoutWriterMode(string(out.RequestedMode))
	out.DurableRequestRef = normalizeOneDisplaySafeRef(out.DurableRequestRef)
	out.WriterOptInRef = normalizeOneDisplaySafeRef(out.WriterOptInRef)
	out.WriterRef = normalizeOneDisplaySafeRef(out.WriterRef)
	out.OwnerRef = normalizeOneDisplaySafeRef(out.OwnerRef)
	out.HostWriterBindingRef = normalizeOneDisplaySafeRef(out.HostWriterBindingRef)
	out.DryRunSmokeRef = normalizeOneDisplaySafeRef(out.DryRunSmokeRef)
	out.DryRunResultRef = normalizeOneDisplaySafeRef(out.DryRunResultRef)
	out.ExpectedDryRunResultRef = normalizeOneDisplaySafeRef(out.ExpectedDryRunResultRef)
	out.ExpectedDurableResultRef = normalizeOneDisplaySafeRef(out.ExpectedDurableResultRef)
	out.ExpectedReadbackRef = normalizeOneDisplaySafeRef(out.ExpectedReadbackRef)
	out.ObjectiveCloseoutHandoffRef = normalizeOneDisplaySafeRef(out.ObjectiveCloseoutHandoffRef)
	out.HostUIHandoffRef = normalizeOneDisplaySafeRef(out.HostUIHandoffRef)
	out.ObjectiveCloseoutPacketRef = normalizeOneDisplaySafeRef(out.ObjectiveCloseoutPacketRef)
	out.ObjectiveRef = normalizeOneDisplaySafeRef(out.ObjectiveRef)
	out.HostObjectiveLifecycleRef = normalizeOneDisplaySafeRef(out.HostObjectiveLifecycleRef)
	out.HostRunstoreRef = normalizeOneDisplaySafeRef(out.HostRunstoreRef)
	out.ExpectedDurableEventRef = normalizeOneDisplaySafeRef(out.ExpectedDurableEventRef)
	out.ExpectedObjectiveStateRef = normalizeOneDisplaySafeRef(out.ExpectedObjectiveStateRef)
	out.HostDurableApplyConfirmationRef = normalizeOneDisplaySafeRef(out.HostDurableApplyConfirmationRef)
	out.HostDurableWriteConfirmationRef = normalizeOneDisplaySafeRef(out.HostDurableWriteConfirmationRef)
	out.AvailableCapabilityRefs = normalizeDisplaySafeRefs(out.AvailableCapabilityRefs)
	out.RequiredCapabilityRefs = normalizeDisplaySafeRefs(out.RequiredCapabilityRefs)
	out.PolicyRefs = normalizeDisplaySafeRefs(out.PolicyRefs)
	out.RequiredPolicyRefs = normalizeDisplaySafeRefs(out.RequiredPolicyRefs)
	out.ApprovalRefs = normalizeDisplaySafeRefs(out.ApprovalRefs)
	out.RequiredApprovalRefs = normalizeDisplaySafeRefs(out.RequiredApprovalRefs)
	out.BudgetRef = normalizeOneDisplaySafeRef(out.BudgetRef)
	out.RequiredBudgetRef = normalizeOneDisplaySafeRef(out.RequiredBudgetRef)
	out.IdempotencyRef = normalizeOneDisplaySafeRef(out.IdempotencyRef)
	out.IdempotencyContractRef = normalizeOneDisplaySafeRef(out.IdempotencyContractRef)
	out.DryRunPlanRef = normalizeOneDisplaySafeRef(out.DryRunPlanRef)
	out.DryRunContractRef = normalizeOneDisplaySafeRef(out.DryRunContractRef)
	out.ReadbackContractRef = normalizeOneDisplaySafeRef(out.ReadbackContractRef)
	out.RollbackReviewRef = normalizeOneDisplaySafeRef(out.RollbackReviewRef)
	out.CompensationReviewRef = normalizeOneDisplaySafeRef(out.CompensationReviewRef)
	out.RedactionPolicyRef = normalizeOneDisplaySafeRef(out.RedactionPolicyRef)
	out.TimeoutPolicyRef = normalizeOneDisplaySafeRef(out.TimeoutPolicyRef)
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
		out.ReadyForHostDurableWrite = false
		out.HostDurableWriteAuthorized = false
		out.HostMayExecuteDurableWrite = false
	}
	if out.RawOutputLoaded || productionAdapterObjectiveCloseoutWriterDurableRequestOutputUnsafe(out) {
		out.RawOutputLoaded = true
		out.Status = HostActionBlocked
		out.ReadyForHostDurableWrite = false
		out.HostDurableWriteAuthorized = false
		out.HostMayExecuteDurableWrite = false
		out.FailureClass = firstFailureClass(out.FailureClass, FailureEvidenceWeak)
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		if out.NextHostAction == "" {
			out.NextHostAction = "provide_display_safe_refs"
		}
	}
	out.CoreInvocationExecuted = false
	out.DryRunByCore = false
	out.DurableWriteByCore = false
	out.ObjectiveStoreWriteByCore = false
	out.RunstoreWriteByCore = false
	out.ReadyForHostDurableWrite = out.ReadyForHostDurableWrite &&
		out.Status == HostActionReady &&
		out.RequestedMode == ProductionAdapterObjectiveCloseoutWriterDurableWrite &&
		out.DurableRequestRef != "" &&
		out.WriterOptInRef != "" &&
		out.WriterRef != "" &&
		out.HostWriterBindingRef != "" &&
		out.DryRunSmokeRef != "" &&
		out.DryRunResultRef != "" &&
		out.ExpectedDurableResultRef != "" &&
		out.ExpectedReadbackRef != "" &&
		out.ExpectedDurableEventRef != "" &&
		out.ExpectedObjectiveStateRef != "" &&
		out.HostDurableWriteConfirmationRef != "" &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.RawOutputLoaded
	out.HostDurableWriteAuthorized = out.HostDurableWriteAuthorized && out.ReadyForHostDurableWrite
	out.HostMayExecuteDurableWrite = out.HostMayExecuteDurableWrite &&
		out.ReadyForHostDurableWrite &&
		!out.CoreInvocationExecuted &&
		!out.DurableWriteByCore &&
		!out.ObjectiveStoreWriteByCore &&
		!out.RunstoreWriteByCore
	return out
}

func CloneProductionAdapterObjectiveCloseoutWriterDurableResult(in ProductionAdapterObjectiveCloseoutWriterDurableResult) ProductionAdapterObjectiveCloseoutWriterDurableResult {
	out := in
	out.DurableEvidenceRefs = cloneDisplaySafeRefs(in.DurableEvidenceRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (r ProductionAdapterObjectiveCloseoutWriterDurableResult) Clone() ProductionAdapterObjectiveCloseoutWriterDurableResult {
	return CloneProductionAdapterObjectiveCloseoutWriterDurableResult(r)
}

func (r ProductionAdapterObjectiveCloseoutWriterDurableResult) Normalize() ProductionAdapterObjectiveCloseoutWriterDurableResult {
	out := CloneProductionAdapterObjectiveCloseoutWriterDurableResult(r)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.Mode = normalizeControlToken(out.Mode)
	if out.Mode == "" {
		out.Mode = "production_adapter_objective_closeout_writer_durable_result"
	}
	out.DurableResultRef = normalizeOneDisplaySafeRef(out.DurableResultRef)
	out.ExpectedDurableResultRef = normalizeOneDisplaySafeRef(out.ExpectedDurableResultRef)
	out.DurableRequestRef = normalizeOneDisplaySafeRef(out.DurableRequestRef)
	out.WriterOptInRef = normalizeOneDisplaySafeRef(out.WriterOptInRef)
	out.WriterRef = normalizeOneDisplaySafeRef(out.WriterRef)
	out.HostWriterBindingRef = normalizeOneDisplaySafeRef(out.HostWriterBindingRef)
	out.HostAdapterRunRef = normalizeOneDisplaySafeRef(out.HostAdapterRunRef)
	out.DryRunSmokeRef = normalizeOneDisplaySafeRef(out.DryRunSmokeRef)
	out.DryRunResultRef = normalizeOneDisplaySafeRef(out.DryRunResultRef)
	out.ExpectedReadbackRef = normalizeOneDisplaySafeRef(out.ExpectedReadbackRef)
	out.ObjectiveCloseoutHandoffRef = normalizeOneDisplaySafeRef(out.ObjectiveCloseoutHandoffRef)
	out.ObjectiveCloseoutPacketRef = normalizeOneDisplaySafeRef(out.ObjectiveCloseoutPacketRef)
	out.ObjectiveRef = normalizeOneDisplaySafeRef(out.ObjectiveRef)
	out.HostRunstoreRef = normalizeOneDisplaySafeRef(out.HostRunstoreRef)
	out.ExpectedDurableEventRef = normalizeOneDisplaySafeRef(out.ExpectedDurableEventRef)
	out.ExpectedObjectiveStateRef = normalizeOneDisplaySafeRef(out.ExpectedObjectiveStateRef)
	out.AppliedDurableEventRef = normalizeOneDisplaySafeRef(out.AppliedDurableEventRef)
	out.AppliedRunstoreRef = normalizeOneDisplaySafeRef(out.AppliedRunstoreRef)
	out.AppliedObjectiveStateRef = normalizeOneDisplaySafeRef(out.AppliedObjectiveStateRef)
	out.FailureRef = normalizeOneDisplaySafeRef(out.FailureRef)
	out.CompensationRef = normalizeOneDisplaySafeRef(out.CompensationRef)
	out.IdempotencyRef = normalizeOneDisplaySafeRef(out.IdempotencyRef)
	out.IdempotencyContractRef = normalizeOneDisplaySafeRef(out.IdempotencyContractRef)
	out.ReadbackContractRef = normalizeOneDisplaySafeRef(out.ReadbackContractRef)
	out.RollbackReviewRef = normalizeOneDisplaySafeRef(out.RollbackReviewRef)
	out.CompensationReviewRef = normalizeOneDisplaySafeRef(out.CompensationReviewRef)
	out.DurableEvidenceRefs = normalizeDisplaySafeRefs(out.DurableEvidenceRefs)
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
		out.ReadyForWriterDurableReadback = false
		out.HostDurableWriteRecorded = false
	}
	if out.RawOutputLoaded || productionAdapterObjectiveCloseoutWriterDurableResultOutputUnsafe(out) {
		out.RawOutputLoaded = true
		out.Status = HostActionBlocked
		out.ReadyForWriterDurableReadback = false
		out.HostDurableWriteRecorded = false
		out.FailureClass = firstFailureClass(out.FailureClass, FailureEvidenceWeak)
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		if out.NextHostAction == "" {
			out.NextHostAction = "provide_display_safe_refs"
		}
	}
	out.CoreInvocationExecuted = false
	out.DryRunByCore = false
	out.DurableWriteByCore = false
	out.ObjectiveStoreWriteByCore = false
	out.RunstoreWriteByCore = false
	out.HostDurableWriteRecorded = out.HostDurableWriteRecorded &&
		(out.Status == HostActionRecorded || out.Status == HostActionReviewRequired) &&
		out.HostDurableWriteReported &&
		out.DurableRequestRef != "" &&
		out.DurableResultRef != "" &&
		out.HostAdapterRunRef != "" &&
		!out.RawOutputLoaded
	out.ReadyForWriterDurableReadback = out.ReadyForWriterDurableReadback &&
		out.Status == HostActionRecorded &&
		out.HostDurableWriteRecorded &&
		out.HostDurableWriteSucceeded &&
		!out.HostDurableWriteFailed &&
		out.AppliedDurableEventRef != "" &&
		out.AppliedRunstoreRef != "" &&
		out.AppliedObjectiveStateRef != "" &&
		out.AppliedDurableEventRef == out.ExpectedDurableEventRef &&
		out.AppliedRunstoreRef == out.HostRunstoreRef &&
		out.AppliedObjectiveStateRef == out.ExpectedObjectiveStateRef &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.RawOutputLoaded
	return out
}

func CloneProductionAdapterObjectiveCloseoutWriterDurableReadback(in ProductionAdapterObjectiveCloseoutWriterDurableReadback) ProductionAdapterObjectiveCloseoutWriterDurableReadback {
	out := in
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (r ProductionAdapterObjectiveCloseoutWriterDurableReadback) Clone() ProductionAdapterObjectiveCloseoutWriterDurableReadback {
	return CloneProductionAdapterObjectiveCloseoutWriterDurableReadback(r)
}

func (r ProductionAdapterObjectiveCloseoutWriterDurableReadback) Normalize() ProductionAdapterObjectiveCloseoutWriterDurableReadback {
	out := CloneProductionAdapterObjectiveCloseoutWriterDurableReadback(r)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.Mode = normalizeControlToken(out.Mode)
	if out.Mode == "" {
		out.Mode = "production_adapter_objective_closeout_writer_durable_readback"
	}
	out.DurableReadbackRef = normalizeOneDisplaySafeRef(out.DurableReadbackRef)
	out.DurableResultRef = normalizeOneDisplaySafeRef(out.DurableResultRef)
	out.DurableRequestRef = normalizeOneDisplaySafeRef(out.DurableRequestRef)
	out.WriterOptInRef = normalizeOneDisplaySafeRef(out.WriterOptInRef)
	out.WriterRef = normalizeOneDisplaySafeRef(out.WriterRef)
	out.HostAdapterRunRef = normalizeOneDisplaySafeRef(out.HostAdapterRunRef)
	out.DryRunSmokeRef = normalizeOneDisplaySafeRef(out.DryRunSmokeRef)
	out.DryRunResultRef = normalizeOneDisplaySafeRef(out.DryRunResultRef)
	out.ExpectedReadbackRef = normalizeOneDisplaySafeRef(out.ExpectedReadbackRef)
	out.ObjectiveCloseoutReadbackRef = normalizeOneDisplaySafeRef(out.ObjectiveCloseoutReadbackRef)
	out.ObjectiveCloseoutHandoffRef = normalizeOneDisplaySafeRef(out.ObjectiveCloseoutHandoffRef)
	out.ObjectiveCloseoutPacketRef = normalizeOneDisplaySafeRef(out.ObjectiveCloseoutPacketRef)
	out.ObjectiveRef = normalizeOneDisplaySafeRef(out.ObjectiveRef)
	out.HostRunstoreRef = normalizeOneDisplaySafeRef(out.HostRunstoreRef)
	out.ExpectedDurableEventRef = normalizeOneDisplaySafeRef(out.ExpectedDurableEventRef)
	out.ExpectedObjectiveStateRef = normalizeOneDisplaySafeRef(out.ExpectedObjectiveStateRef)
	out.AppliedDurableEventRef = normalizeOneDisplaySafeRef(out.AppliedDurableEventRef)
	out.AppliedRunstoreRef = normalizeOneDisplaySafeRef(out.AppliedRunstoreRef)
	out.AppliedObjectiveStateRef = normalizeOneDisplaySafeRef(out.AppliedObjectiveStateRef)
	out.ObservedObjectiveCloseoutPacketRef = normalizeOneDisplaySafeRef(out.ObservedObjectiveCloseoutPacketRef)
	out.ObservedObjectiveRef = normalizeOneDisplaySafeRef(out.ObservedObjectiveRef)
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
		out.WriterDurableReadbackBound = false
		out.ReadyForObjectiveReturn = false
		out.ObjectiveLifecycleClosed = false
		out.ObjectiveSatisfied = false
	}
	if out.RawOutputLoaded || productionAdapterObjectiveCloseoutWriterDurableReadbackOutputUnsafe(out) {
		out.RawOutputLoaded = true
		out.Status = HostActionBlocked
		out.WriterDurableReadbackBound = false
		out.ReadyForObjectiveReturn = false
		out.ObjectiveLifecycleClosed = false
		out.ObjectiveSatisfied = false
		out.FailureClass = firstFailureClass(out.FailureClass, FailureEvidenceWeak)
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		if out.NextHostAction == "" {
			out.NextHostAction = "provide_display_safe_refs"
		}
	}
	out.CoreInvocationExecuted = false
	out.DryRunByCore = false
	out.DurableWriteByCore = false
	out.ObjectiveStoreWriteByCore = false
	out.RunstoreWriteByCore = false
	out.WriterDurableReadbackBound = out.WriterDurableReadbackBound &&
		out.Status == HostActionRecorded &&
		out.DurableReadbackRef != "" &&
		out.DurableResultRef != "" &&
		out.ObjectiveCloseoutReadbackRef != "" &&
		out.ExpectedReadbackRef == out.ObjectiveCloseoutReadbackRef &&
		out.ObjectiveLifecycleClosed &&
		out.ObjectiveSatisfied &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.RawOutputLoaded
	out.ReadyForObjectiveReturn = out.ReadyForObjectiveReturn && out.WriterDurableReadbackBound
	return out
}

func productionAdapterObjectiveCloseoutWriterDurableRequestBlock(result ProductionAdapterObjectiveCloseoutWriterDurableRequest, failure FailureClass, reason string, missing MissingInput, next NextHostAction) ProductionAdapterObjectiveCloseoutWriterDurableRequest {
	result.Status = HostActionBlocked
	result.ReadyForHostDurableWrite = false
	result.HostDurableWriteAuthorized = false
	result.HostMayExecuteDurableWrite = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.Boundaries = AppendBoundaries(result.Boundaries, "objective_closeout_writer_durable_request_blocked")
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	return result
}

func productionAdapterObjectiveCloseoutWriterDurableResultBlock(result ProductionAdapterObjectiveCloseoutWriterDurableResult, failure FailureClass, reason string, missing MissingInput, next NextHostAction) ProductionAdapterObjectiveCloseoutWriterDurableResult {
	result.Status = HostActionBlocked
	result.ReadyForWriterDurableReadback = false
	result.HostDurableWriteRecorded = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.Boundaries = AppendBoundaries(result.Boundaries, "objective_closeout_writer_durable_result_blocked")
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	return result
}

func productionAdapterObjectiveCloseoutWriterDurableReadbackBlock(result ProductionAdapterObjectiveCloseoutWriterDurableReadback, failure FailureClass, reason string, missing MissingInput, next NextHostAction) ProductionAdapterObjectiveCloseoutWriterDurableReadback {
	result.Status = HostActionBlocked
	result.WriterDurableReadbackBound = false
	result.ReadyForObjectiveReturn = false
	result.ObjectiveLifecycleClosed = false
	result.ObjectiveSatisfied = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.Boundaries = AppendBoundaries(result.Boundaries, "objective_closeout_writer_durable_readback_blocked")
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	return result
}

type productionAdapterObjectiveCloseoutWriterDurableMismatch struct {
	reason  string
	missing MissingInput
}

func productionAdapterObjectiveCloseoutWriterDurableRequestMismatches(optIn ProductionAdapterObjectiveCloseoutWriterOptIn, smoke ProductionAdapterObjectiveCloseoutWriterDryRunSmokeHarness) []productionAdapterObjectiveCloseoutWriterDurableMismatch {
	var out []productionAdapterObjectiveCloseoutWriterDurableMismatch
	out = append(out, productionAdapterObjectiveCloseoutWriterDurableRefMismatch(optIn.WriterRef, smoke.WriterRef, "durable_request_writer_ref_mismatch", "host:objective_closeout_writer_ref")...)
	out = append(out, productionAdapterObjectiveCloseoutWriterDurableRefMismatch(optIn.DryRunResultRef, smoke.DryRunResultRef, "durable_request_dry_run_result_ref_mismatch", "host:objective_closeout_writer_dry_run_result_ref")...)
	out = append(out, productionAdapterObjectiveCloseoutWriterDurableRefMismatch(optIn.ExpectedReadbackRef, smoke.ExpectedReadbackRef, "durable_request_expected_readback_ref_mismatch", "host:objective_closeout_writer_expected_readback_ref")...)
	return out
}

func productionAdapterObjectiveCloseoutWriterDurableResultMismatches(result ProductionAdapterObjectiveCloseoutWriterDurableResult) []productionAdapterObjectiveCloseoutWriterDurableMismatch {
	var out []productionAdapterObjectiveCloseoutWriterDurableMismatch
	out = append(out, productionAdapterObjectiveCloseoutWriterDurableRefMismatch(result.ExpectedDurableEventRef, result.AppliedDurableEventRef, "writer_durable_event_ref_mismatch", "host:durable_event_ref")...)
	out = append(out, productionAdapterObjectiveCloseoutWriterDurableRefMismatch(result.HostRunstoreRef, result.AppliedRunstoreRef, "writer_runstore_ref_mismatch", "host:runstore_ref")...)
	out = append(out, productionAdapterObjectiveCloseoutWriterDurableRefMismatch(result.ExpectedObjectiveStateRef, result.AppliedObjectiveStateRef, "writer_objective_state_ref_mismatch", "host:objective_state_ref")...)
	return out
}

func productionAdapterObjectiveCloseoutWriterDurableReadbackMismatches(result ProductionAdapterObjectiveCloseoutWriterDurableResult, readback ProductionAdapterObjectiveCloseoutReadback) []productionAdapterObjectiveCloseoutWriterDurableMismatch {
	var out []productionAdapterObjectiveCloseoutWriterDurableMismatch
	out = append(out, productionAdapterObjectiveCloseoutWriterDurableRefMismatch(result.ExpectedReadbackRef, readback.ObjectiveCloseoutReadbackRef, "writer_readback_ref_mismatch", "host:objective_closeout_readback_ref")...)
	out = append(out, productionAdapterObjectiveCloseoutWriterDurableRefMismatch(result.ObjectiveCloseoutHandoffRef, readback.ObjectiveCloseoutHandoffRef, "writer_readback_handoff_ref_mismatch", "host:objective_closeout_handoff_ref")...)
	out = append(out, productionAdapterObjectiveCloseoutWriterDurableRefMismatch(result.ObjectiveCloseoutPacketRef, readback.ObjectiveCloseoutPacketRef, "writer_readback_packet_ref_mismatch", "host:objective_closeout_packet_ref")...)
	out = append(out, productionAdapterObjectiveCloseoutWriterDurableRefMismatch(result.ObjectiveRef, readback.ObjectiveRef, "writer_readback_objective_ref_mismatch", "host:objective_ref")...)
	out = append(out, productionAdapterObjectiveCloseoutWriterDurableRefMismatch(result.HostRunstoreRef, readback.AppliedRunstoreRef, "writer_readback_runstore_ref_mismatch", "host:runstore_ref")...)
	out = append(out, productionAdapterObjectiveCloseoutWriterDurableRefMismatch(result.ExpectedDurableEventRef, readback.AppliedDurableEventRef, "writer_readback_durable_event_ref_mismatch", "host:durable_event_ref")...)
	out = append(out, productionAdapterObjectiveCloseoutWriterDurableRefMismatch(result.ExpectedObjectiveStateRef, readback.AppliedObjectiveStateRef, "writer_readback_objective_state_ref_mismatch", "host:objective_state_ref")...)
	return out
}

func productionAdapterObjectiveCloseoutWriterDurableRefMismatch(left DisplaySafeRef, right DisplaySafeRef, reason string, missing MissingInput) []productionAdapterObjectiveCloseoutWriterDurableMismatch {
	left = normalizeOneDisplaySafeRef(left)
	right = normalizeOneDisplaySafeRef(right)
	if left != "" && right != "" && left != right {
		return []productionAdapterObjectiveCloseoutWriterDurableMismatch{{reason: reason, missing: missing}}
	}
	return nil
}

func productionAdapterObjectiveCloseoutWriterDurableRequestBoundaries(groups ...[]Boundary) []Boundary {
	return MergeBoundaries(
		[]Boundary{
			"production_adapter_objective_closeout_writer_durable_request",
			"objective_closeout_writer_durable_request_projection_only",
			"host_owned_objective_closeout_writer_adapter",
			"objective_closeout_writer_durable_write_only",
			"dry_run_smoke_required",
			"display_safe_refs_only",
			"display_safe_result_refs_only",
			"no_runner_dispatch",
			"no_dry_run_by_core",
			"no_durable_write_by_core",
			"no_objective_store_write_by_core",
			"no_runstore_write_by_core",
		},
		MergeBoundaries(groups...),
	)
}

func productionAdapterObjectiveCloseoutWriterDurableResultBoundaries(groups ...[]Boundary) []Boundary {
	return MergeBoundaries(
		[]Boundary{
			"production_adapter_objective_closeout_writer_durable_result",
			"objective_closeout_writer_durable_result_projection_only",
			"host_owned_objective_closeout_writer_adapter",
			"objective_closeout_writer_durable_write_only",
			"display_safe_refs_only",
			"display_safe_result_refs_only",
			"no_runner_dispatch",
			"no_dry_run_by_core",
			"no_durable_write_by_core",
			"no_objective_store_write_by_core",
			"no_runstore_write_by_core",
		},
		MergeBoundaries(groups...),
	)
}

func productionAdapterObjectiveCloseoutWriterDurableReadbackBoundaries(groups ...[]Boundary) []Boundary {
	return MergeBoundaries(
		[]Boundary{
			"production_adapter_objective_closeout_writer_durable_readback",
			"objective_closeout_writer_durable_readback_projection_only",
			"host_owned_objective_closeout_writer_adapter",
			"authorization_bound_objective_closeout_readback",
			"display_safe_refs_only",
			"display_safe_result_refs_only",
			"no_runner_dispatch",
			"no_dry_run_by_core",
			"no_durable_write_by_core",
			"no_objective_store_write_by_core",
			"no_runstore_write_by_core",
		},
		MergeBoundaries(groups...),
	)
}

func productionAdapterObjectiveCloseoutWriterDurableRequestUnsafe(input ProductionAdapterObjectiveCloseoutWriterDurableRequestInput, optIn ProductionAdapterObjectiveCloseoutWriterOptIn, smoke ProductionAdapterObjectiveCloseoutWriterDryRunSmokeHarness) bool {
	return input.RawOutputLoaded ||
		displaySafeRefRejected(input.DurableRequestRef) ||
		displaySafeRefRejected(input.HostDurableWriteConfirmationRef) ||
		displaySafeRefRejected(input.ExpectedDurableResultRef) ||
		productionAdapterObjectiveCloseoutWriterOptInOutputUnsafe(optIn) ||
		productionAdapterObjectiveCloseoutWriterDryRunSmokeHarnessOutputUnsafe(smoke)
}

func productionAdapterObjectiveCloseoutWriterDurableRequestOutputUnsafe(input ProductionAdapterObjectiveCloseoutWriterDurableRequest) bool {
	return displaySafeRefRejected(input.DurableRequestRef) ||
		displaySafeRefRejected(input.WriterOptInRef) ||
		displaySafeRefRejected(input.WriterRef) ||
		displaySafeRefRejected(input.OwnerRef) ||
		displaySafeRefRejected(input.HostWriterBindingRef) ||
		displaySafeRefRejected(input.DryRunSmokeRef) ||
		displaySafeRefRejected(input.DryRunResultRef) ||
		displaySafeRefRejected(input.ExpectedDryRunResultRef) ||
		displaySafeRefRejected(input.ExpectedDurableResultRef) ||
		displaySafeRefRejected(input.ExpectedReadbackRef) ||
		displaySafeRefRejected(input.ObjectiveCloseoutHandoffRef) ||
		displaySafeRefRejected(input.HostUIHandoffRef) ||
		displaySafeRefRejected(input.ObjectiveCloseoutPacketRef) ||
		displaySafeRefRejected(input.ObjectiveRef) ||
		displaySafeRefRejected(input.HostObjectiveLifecycleRef) ||
		displaySafeRefRejected(input.HostRunstoreRef) ||
		displaySafeRefRejected(input.ExpectedDurableEventRef) ||
		displaySafeRefRejected(input.ExpectedObjectiveStateRef) ||
		displaySafeRefRejected(input.HostDurableApplyConfirmationRef) ||
		displaySafeRefRejected(input.HostDurableWriteConfirmationRef) ||
		displaySafeRefSliceRejected(input.AvailableCapabilityRefs) ||
		displaySafeRefSliceRejected(input.RequiredCapabilityRefs) ||
		displaySafeRefSliceRejected(input.PolicyRefs) ||
		displaySafeRefSliceRejected(input.RequiredPolicyRefs) ||
		displaySafeRefSliceRejected(input.ApprovalRefs) ||
		displaySafeRefSliceRejected(input.RequiredApprovalRefs) ||
		displaySafeRefRejected(input.BudgetRef) ||
		displaySafeRefRejected(input.RequiredBudgetRef) ||
		displaySafeRefRejected(input.IdempotencyRef) ||
		displaySafeRefRejected(input.IdempotencyContractRef) ||
		displaySafeRefRejected(input.DryRunPlanRef) ||
		displaySafeRefRejected(input.DryRunContractRef) ||
		displaySafeRefRejected(input.ReadbackContractRef) ||
		displaySafeRefRejected(input.RollbackReviewRef) ||
		displaySafeRefRejected(input.CompensationReviewRef) ||
		displaySafeRefRejected(input.RedactionPolicyRef) ||
		displaySafeRefRejected(input.TimeoutPolicyRef) ||
		input.RawOutputLoaded
}

func productionAdapterObjectiveCloseoutWriterDurableResultUnsafe(input ProductionAdapterObjectiveCloseoutWriterDurableResultInput, request ProductionAdapterObjectiveCloseoutWriterDurableRequest) bool {
	return input.RawOutputLoaded ||
		displaySafeRefRejected(input.DurableResultRef) ||
		displaySafeRefRejected(input.HostAdapterRunRef) ||
		displaySafeRefRejected(input.AppliedDurableEventRef) ||
		displaySafeRefRejected(input.AppliedRunstoreRef) ||
		displaySafeRefRejected(input.AppliedObjectiveStateRef) ||
		displaySafeRefRejected(input.FailureRef) ||
		displaySafeRefRejected(input.CompensationRef) ||
		displaySafeRefSliceRejected(input.DurableEvidenceRefs) ||
		productionAdapterObjectiveCloseoutWriterDurableRequestOutputUnsafe(request)
}

func productionAdapterObjectiveCloseoutWriterDurableResultOutputUnsafe(input ProductionAdapterObjectiveCloseoutWriterDurableResult) bool {
	return displaySafeRefRejected(input.DurableResultRef) ||
		displaySafeRefRejected(input.ExpectedDurableResultRef) ||
		displaySafeRefRejected(input.DurableRequestRef) ||
		displaySafeRefRejected(input.WriterOptInRef) ||
		displaySafeRefRejected(input.WriterRef) ||
		displaySafeRefRejected(input.HostWriterBindingRef) ||
		displaySafeRefRejected(input.HostAdapterRunRef) ||
		displaySafeRefRejected(input.DryRunSmokeRef) ||
		displaySafeRefRejected(input.DryRunResultRef) ||
		displaySafeRefRejected(input.ExpectedReadbackRef) ||
		displaySafeRefRejected(input.ObjectiveCloseoutHandoffRef) ||
		displaySafeRefRejected(input.ObjectiveCloseoutPacketRef) ||
		displaySafeRefRejected(input.ObjectiveRef) ||
		displaySafeRefRejected(input.HostRunstoreRef) ||
		displaySafeRefRejected(input.ExpectedDurableEventRef) ||
		displaySafeRefRejected(input.ExpectedObjectiveStateRef) ||
		displaySafeRefRejected(input.AppliedDurableEventRef) ||
		displaySafeRefRejected(input.AppliedRunstoreRef) ||
		displaySafeRefRejected(input.AppliedObjectiveStateRef) ||
		displaySafeRefRejected(input.FailureRef) ||
		displaySafeRefRejected(input.CompensationRef) ||
		displaySafeRefRejected(input.IdempotencyRef) ||
		displaySafeRefRejected(input.IdempotencyContractRef) ||
		displaySafeRefRejected(input.ReadbackContractRef) ||
		displaySafeRefRejected(input.RollbackReviewRef) ||
		displaySafeRefRejected(input.CompensationReviewRef) ||
		displaySafeRefSliceRejected(input.DurableEvidenceRefs) ||
		input.RawOutputLoaded
}

func productionAdapterObjectiveCloseoutWriterDurableReadbackUnsafe(input ProductionAdapterObjectiveCloseoutWriterDurableReadbackInput, result ProductionAdapterObjectiveCloseoutWriterDurableResult, readback ProductionAdapterObjectiveCloseoutReadback) bool {
	return input.RawOutputLoaded ||
		displaySafeRefRejected(input.DurableReadbackRef) ||
		productionAdapterObjectiveCloseoutWriterDurableResultOutputUnsafe(result) ||
		productionAdapterObjectiveCloseoutReadbackRefsUnsafe(readback)
}

func productionAdapterObjectiveCloseoutWriterDurableReadbackOutputUnsafe(input ProductionAdapterObjectiveCloseoutWriterDurableReadback) bool {
	return displaySafeRefRejected(input.DurableReadbackRef) ||
		displaySafeRefRejected(input.DurableResultRef) ||
		displaySafeRefRejected(input.DurableRequestRef) ||
		displaySafeRefRejected(input.WriterOptInRef) ||
		displaySafeRefRejected(input.WriterRef) ||
		displaySafeRefRejected(input.HostAdapterRunRef) ||
		displaySafeRefRejected(input.DryRunSmokeRef) ||
		displaySafeRefRejected(input.DryRunResultRef) ||
		displaySafeRefRejected(input.ExpectedReadbackRef) ||
		displaySafeRefRejected(input.ObjectiveCloseoutReadbackRef) ||
		displaySafeRefRejected(input.ObjectiveCloseoutHandoffRef) ||
		displaySafeRefRejected(input.ObjectiveCloseoutPacketRef) ||
		displaySafeRefRejected(input.ObjectiveRef) ||
		displaySafeRefRejected(input.HostRunstoreRef) ||
		displaySafeRefRejected(input.ExpectedDurableEventRef) ||
		displaySafeRefRejected(input.ExpectedObjectiveStateRef) ||
		displaySafeRefRejected(input.AppliedDurableEventRef) ||
		displaySafeRefRejected(input.AppliedRunstoreRef) ||
		displaySafeRefRejected(input.AppliedObjectiveStateRef) ||
		displaySafeRefRejected(input.ObservedObjectiveCloseoutPacketRef) ||
		displaySafeRefRejected(input.ObservedObjectiveRef) ||
		input.RawOutputLoaded
}

func productionAdapterObjectiveCloseoutWriterDryRunSmokeHarnessEmpty(smoke ProductionAdapterObjectiveCloseoutWriterDryRunSmokeHarness) bool {
	return !smoke.Projected &&
		!smoke.Available &&
		smoke.Status == "" &&
		smoke.Mode == "" &&
		smoke.SmokeRef == "" &&
		smoke.DryRunRequestRef == "" &&
		smoke.DryRunResultRef == "" &&
		len(smoke.MissingInputs) == 0 &&
		len(smoke.BlockedReasons) == 0 &&
		len(smoke.Boundaries) == 0 &&
		smoke.NextHostAction == "" &&
		!smoke.RawOutputLoaded
}

func productionAdapterObjectiveCloseoutWriterDurableRequestEmpty(request ProductionAdapterObjectiveCloseoutWriterDurableRequest) bool {
	return !request.Projected &&
		!request.Available &&
		request.Status == "" &&
		request.Mode == "" &&
		request.DurableRequestRef == "" &&
		request.WriterOptInRef == "" &&
		request.WriterRef == "" &&
		len(request.MissingInputs) == 0 &&
		len(request.BlockedReasons) == 0 &&
		len(request.Boundaries) == 0 &&
		request.NextHostAction == "" &&
		!request.RawOutputLoaded
}

func productionAdapterObjectiveCloseoutWriterDurableResultEmpty(result ProductionAdapterObjectiveCloseoutWriterDurableResult) bool {
	return !result.Projected &&
		!result.Available &&
		result.Status == "" &&
		result.Mode == "" &&
		result.DurableResultRef == "" &&
		result.DurableRequestRef == "" &&
		result.WriterOptInRef == "" &&
		len(result.MissingInputs) == 0 &&
		len(result.BlockedReasons) == 0 &&
		len(result.Boundaries) == 0 &&
		result.NextHostAction == "" &&
		!result.RawOutputLoaded
}

func unavailableProductionAdapterObjectiveCloseoutWriterDurableRequest() ProductionAdapterObjectiveCloseoutWriterDurableRequest {
	return ProductionAdapterObjectiveCloseoutWriterDurableRequest{
		ContractVersion: ContractVersion,
		Projected:       true,
		Available:       false,
		Status:          HostActionNotReady,
		Mode:            "production_adapter_objective_closeout_writer_durable_request",
		Boundaries: []Boundary{
			"production_adapter_objective_closeout_writer_durable_request",
			"objective_closeout_writer_durable_request_projection_only",
			"host_owned_objective_closeout_writer_adapter",
			"dry_run_smoke_required",
			"no_runner_dispatch",
			"no_dry_run_by_core",
			"no_durable_write_by_core",
			"no_objective_store_write_by_core",
			"no_runstore_write_by_core",
		},
		RunnerEffect:   "none",
		PromptEffect:   "none",
		NextHostAction: "provide_objective_closeout_writer_dry_run_smoke",
	}
}

func unavailableProductionAdapterObjectiveCloseoutWriterDurableResult() ProductionAdapterObjectiveCloseoutWriterDurableResult {
	return ProductionAdapterObjectiveCloseoutWriterDurableResult{
		ContractVersion: ContractVersion,
		Projected:       true,
		Available:       false,
		Status:          HostActionNotReady,
		Mode:            "production_adapter_objective_closeout_writer_durable_result",
		Boundaries: []Boundary{
			"production_adapter_objective_closeout_writer_durable_result",
			"objective_closeout_writer_durable_result_projection_only",
			"host_owned_objective_closeout_writer_adapter",
			"no_runner_dispatch",
			"no_dry_run_by_core",
			"no_durable_write_by_core",
			"no_objective_store_write_by_core",
			"no_runstore_write_by_core",
		},
		RunnerEffect:   "none",
		PromptEffect:   "none",
		NextHostAction: "provide_objective_closeout_writer_durable_request",
	}
}

func unavailableProductionAdapterObjectiveCloseoutWriterDurableReadback() ProductionAdapterObjectiveCloseoutWriterDurableReadback {
	return ProductionAdapterObjectiveCloseoutWriterDurableReadback{
		ContractVersion: ContractVersion,
		Projected:       true,
		Available:       false,
		Status:          HostActionNotReady,
		Mode:            "production_adapter_objective_closeout_writer_durable_readback",
		Boundaries: []Boundary{
			"production_adapter_objective_closeout_writer_durable_readback",
			"objective_closeout_writer_durable_readback_projection_only",
			"host_owned_objective_closeout_writer_adapter",
			"no_runner_dispatch",
			"no_dry_run_by_core",
			"no_durable_write_by_core",
			"no_objective_store_write_by_core",
			"no_runstore_write_by_core",
		},
		RunnerEffect:   "none",
		PromptEffect:   "none",
		NextHostAction: "provide_objective_closeout_writer_durable_result",
	}
}
