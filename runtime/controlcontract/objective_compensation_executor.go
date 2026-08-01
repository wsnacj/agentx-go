package controlcontract

type ObjectiveCompensationExecutorDescriptor struct {
	ContractVersion           string           `json:"contract_version,omitempty"`
	Projected                 bool             `json:"projected"`
	Available                 bool             `json:"available"`
	Status                    HostActionStatus `json:"status,omitempty"`
	DescriptorRef             DisplaySafeRef   `json:"descriptor_ref,omitempty"`
	ExecutorRef               DisplaySafeRef   `json:"executor_ref,omitempty"`
	OwnerRef                  DisplaySafeRef   `json:"owner_ref,omitempty"`
	SupportedCompensationRefs []DisplaySafeRef `json:"supported_compensation_refs,omitempty"`
	IdempotencyContractRef    DisplaySafeRef   `json:"idempotency_contract_ref,omitempty"`
	ReadbackContractRef       DisplaySafeRef   `json:"readback_contract_ref,omitempty"`
	RollbackContractRef       DisplaySafeRef   `json:"rollback_contract_ref,omitempty"`
	TimeoutPolicyRef          DisplaySafeRef   `json:"timeout_policy_ref,omitempty"`
	PolicyRefs                []DisplaySafeRef `json:"policy_refs,omitempty"`
	RequiredPolicyRefs        []DisplaySafeRef `json:"required_policy_refs,omitempty"`
	ApprovalRefs              []DisplaySafeRef `json:"approval_refs,omitempty"`
	RequiredApprovalRefs      []DisplaySafeRef `json:"required_approval_refs,omitempty"`
	MissingInputs             []MissingInput   `json:"missing_inputs,omitempty"`
	BlockedReasons            []string         `json:"blocked_reasons,omitempty"`
	FailureClass              FailureClass     `json:"failure_class,omitempty"`
	Boundaries                []Boundary       `json:"boundaries,omitempty"`
	NextHostAction            NextHostAction   `json:"next_host_action,omitempty"`
	RunnerEffect              string           `json:"runner_effect,omitempty"`
	PromptEffect              string           `json:"prompt_effect,omitempty"`
	RawOutputLoaded           bool             `json:"raw_output_loaded"`
}

type ObjectiveCompensationExecutionRequestInput struct {
	CompensationRequestRef      DisplaySafeRef                                        `json:"compensation_request_ref,omitempty"`
	FailureReview               ProductionAdapterObjectiveCloseoutFailureReviewPacket `json:"failure_review,omitempty"`
	ExecutorDescriptor          ObjectiveCompensationExecutorDescriptor               `json:"executor_descriptor,omitempty"`
	HostCompensationApprovalRef DisplaySafeRef                                        `json:"host_compensation_approval_ref,omitempty"`
	IdempotencyRef              DisplaySafeRef                                        `json:"idempotency_ref,omitempty"`
	ExpectedResultRef           DisplaySafeRef                                        `json:"expected_result_ref,omitempty"`
	ExpectedReadbackRef         DisplaySafeRef                                        `json:"expected_readback_ref,omitempty"`
	RollbackPlanRef             DisplaySafeRef                                        `json:"rollback_plan_ref,omitempty"`
	AuditRef                    DisplaySafeRef                                        `json:"audit_ref,omitempty"`
	PolicyRefs                  []DisplaySafeRef                                      `json:"policy_refs,omitempty"`
	ApprovalRefs                []DisplaySafeRef                                      `json:"approval_refs,omitempty"`
	EvidenceRefs                []EvidenceRef                                         `json:"evidence_refs,omitempty"`
	Boundaries                  []Boundary                                            `json:"boundaries,omitempty"`
	RawOutputLoaded             bool                                                  `json:"raw_output_loaded"`
}

type ObjectiveCompensationExecutionRequest struct {
	ContractVersion             string           `json:"contract_version,omitempty"`
	Projected                   bool             `json:"projected"`
	Available                   bool             `json:"available"`
	Status                      HostActionStatus `json:"status,omitempty"`
	ReadyForHostCompensation    bool             `json:"ready_for_host_compensation"`
	HostCompensationAuthorized  bool             `json:"host_compensation_authorized"`
	HostMayExecuteCompensation  bool             `json:"host_may_execute_compensation"`
	CompensationRequestRef      DisplaySafeRef   `json:"compensation_request_ref,omitempty"`
	FailureReviewPacketRef      DisplaySafeRef   `json:"failure_review_packet_ref,omitempty"`
	ObjectiveRef                DisplaySafeRef   `json:"objective_ref,omitempty"`
	FailureRef                  DisplaySafeRef   `json:"failure_ref,omitempty"`
	CompensationRef             DisplaySafeRef   `json:"compensation_ref,omitempty"`
	AdapterRef                  DisplaySafeRef   `json:"adapter_ref,omitempty"`
	DescriptorRef               DisplaySafeRef   `json:"descriptor_ref,omitempty"`
	ExecutorRef                 DisplaySafeRef   `json:"executor_ref,omitempty"`
	OwnerRef                    DisplaySafeRef   `json:"owner_ref,omitempty"`
	HostCompensationApprovalRef DisplaySafeRef   `json:"host_compensation_approval_ref,omitempty"`
	IdempotencyRef              DisplaySafeRef   `json:"idempotency_ref,omitempty"`
	IdempotencyContractRef      DisplaySafeRef   `json:"idempotency_contract_ref,omitempty"`
	ExpectedResultRef           DisplaySafeRef   `json:"expected_result_ref,omitempty"`
	ExpectedReadbackRef         DisplaySafeRef   `json:"expected_readback_ref,omitempty"`
	RollbackPlanRef             DisplaySafeRef   `json:"rollback_plan_ref,omitempty"`
	RollbackContractRef         DisplaySafeRef   `json:"rollback_contract_ref,omitempty"`
	ReadbackContractRef         DisplaySafeRef   `json:"readback_contract_ref,omitempty"`
	AuditRef                    DisplaySafeRef   `json:"audit_ref,omitempty"`
	TimeoutPolicyRef            DisplaySafeRef   `json:"timeout_policy_ref,omitempty"`
	PolicyRefs                  []DisplaySafeRef `json:"policy_refs,omitempty"`
	RequiredPolicyRefs          []DisplaySafeRef `json:"required_policy_refs,omitempty"`
	ApprovalRefs                []DisplaySafeRef `json:"approval_refs,omitempty"`
	RequiredApprovalRefs        []DisplaySafeRef `json:"required_approval_refs,omitempty"`
	EvidenceRefs                []EvidenceRef    `json:"evidence_refs,omitempty"`
	MissingInputs               []MissingInput   `json:"missing_inputs,omitempty"`
	BlockedReasons              []string         `json:"blocked_reasons,omitempty"`
	FailureClass                FailureClass     `json:"failure_class,omitempty"`
	Boundaries                  []Boundary       `json:"boundaries,omitempty"`
	NextHostAction              NextHostAction   `json:"next_host_action,omitempty"`
	RunnerEffect                string           `json:"runner_effect,omitempty"`
	PromptEffect                string           `json:"prompt_effect,omitempty"`
	CoreExecutionExecuted       bool             `json:"core_execution_executed"`
	RunnerDispatched            bool             `json:"runner_dispatched"`
	ToolExecuted                bool             `json:"tool_executed"`
	WorkflowDispatched          bool             `json:"workflow_dispatched"`
	SchedulerApplied            bool             `json:"scheduler_applied"`
	InstallerExecuted           bool             `json:"installer_executed"`
	StoreMutationExecuted       bool             `json:"store_mutation_executed"`
	CompensationExecutedByCore  bool             `json:"compensation_executed_by_core"`
	RawOutputLoaded             bool             `json:"raw_output_loaded"`
}

type ObjectiveCompensationExecutionResultInput struct {
	Request                   ObjectiveCompensationExecutionRequest `json:"request,omitempty"`
	CompensationResultRef     DisplaySafeRef                        `json:"compensation_result_ref,omitempty"`
	HostCompensationRunRef    DisplaySafeRef                        `json:"host_compensation_run_ref,omitempty"`
	HostCompensationReported  bool                                  `json:"host_compensation_reported"`
	HostCompensationSucceeded bool                                  `json:"host_compensation_succeeded"`
	HostCompensationFailed    bool                                  `json:"host_compensation_failed"`
	AppliedCompensationRef    DisplaySafeRef                        `json:"applied_compensation_ref,omitempty"`
	FailureRef                DisplaySafeRef                        `json:"failure_ref,omitempty"`
	ResidualRiskRef           DisplaySafeRef                        `json:"residual_risk_ref,omitempty"`
	CompensationEvidenceRefs  []EvidenceRef                         `json:"compensation_evidence_refs,omitempty"`
	Boundaries                []Boundary                            `json:"boundaries,omitempty"`
	RawOutputLoaded           bool                                  `json:"raw_output_loaded"`
}

type ObjectiveCompensationExecutionResult struct {
	ContractVersion              string                                `json:"contract_version,omitempty"`
	Projected                    bool                                  `json:"projected"`
	Available                    bool                                  `json:"available"`
	Status                       HostActionStatus                      `json:"status,omitempty"`
	ReadyForCompensationReadback bool                                  `json:"ready_for_compensation_readback"`
	ReadyForCloseoutReview       bool                                  `json:"ready_for_closeout_review"`
	HostCompensationReported     bool                                  `json:"host_compensation_reported"`
	HostCompensationSucceeded    bool                                  `json:"host_compensation_succeeded"`
	HostCompensationFailed       bool                                  `json:"host_compensation_failed"`
	HostCompensationExecuted     bool                                  `json:"host_compensation_executed"`
	Request                      ObjectiveCompensationExecutionRequest `json:"request,omitempty"`
	CompensationResultRef        DisplaySafeRef                        `json:"compensation_result_ref,omitempty"`
	ExpectedResultRef            DisplaySafeRef                        `json:"expected_result_ref,omitempty"`
	ExpectedReadbackRef          DisplaySafeRef                        `json:"expected_readback_ref,omitempty"`
	CompensationRequestRef       DisplaySafeRef                        `json:"compensation_request_ref,omitempty"`
	FailureReviewPacketRef       DisplaySafeRef                        `json:"failure_review_packet_ref,omitempty"`
	ObjectiveRef                 DisplaySafeRef                        `json:"objective_ref,omitempty"`
	FailureRef                   DisplaySafeRef                        `json:"failure_ref,omitempty"`
	CompensationRef              DisplaySafeRef                        `json:"compensation_ref,omitempty"`
	AppliedCompensationRef       DisplaySafeRef                        `json:"applied_compensation_ref,omitempty"`
	ResidualRiskRef              DisplaySafeRef                        `json:"residual_risk_ref,omitempty"`
	ExecutorRef                  DisplaySafeRef                        `json:"executor_ref,omitempty"`
	HostCompensationRunRef       DisplaySafeRef                        `json:"host_compensation_run_ref,omitempty"`
	CompensationEvidenceRefs     []EvidenceRef                         `json:"compensation_evidence_refs,omitempty"`
	MissingInputs                []MissingInput                        `json:"missing_inputs,omitempty"`
	BlockedReasons               []string                              `json:"blocked_reasons,omitempty"`
	FailureClass                 FailureClass                          `json:"failure_class,omitempty"`
	Boundaries                   []Boundary                            `json:"boundaries,omitempty"`
	NextHostAction               NextHostAction                        `json:"next_host_action,omitempty"`
	RunnerEffect                 string                                `json:"runner_effect,omitempty"`
	PromptEffect                 string                                `json:"prompt_effect,omitempty"`
	CoreExecutionExecuted        bool                                  `json:"core_execution_executed"`
	RunnerDispatched             bool                                  `json:"runner_dispatched"`
	ToolExecuted                 bool                                  `json:"tool_executed"`
	WorkflowDispatched           bool                                  `json:"workflow_dispatched"`
	SchedulerApplied             bool                                  `json:"scheduler_applied"`
	InstallerExecuted            bool                                  `json:"installer_executed"`
	StoreMutationExecuted        bool                                  `json:"store_mutation_executed"`
	CompensationExecutedByCore   bool                                  `json:"compensation_executed_by_core"`
	RawOutputLoaded              bool                                  `json:"raw_output_loaded"`
}

type ObjectiveCompensationExecutionReadbackInput struct {
	Result                  ObjectiveCompensationExecutionResult `json:"result,omitempty"`
	CompensationReadbackRef DisplaySafeRef                       `json:"compensation_readback_ref,omitempty"`
	ObservedCompensationRef DisplaySafeRef                       `json:"observed_compensation_ref,omitempty"`
	ObservedHostRunRef      DisplaySafeRef                       `json:"observed_host_run_ref,omitempty"`
	ObservedResidualRiskRef DisplaySafeRef                       `json:"observed_residual_risk_ref,omitempty"`
	ReadbackEvidenceRefs    []EvidenceRef                        `json:"readback_evidence_refs,omitempty"`
	Boundaries              []Boundary                           `json:"boundaries,omitempty"`
	RawOutputLoaded         bool                                 `json:"raw_output_loaded"`
}

type ObjectiveCompensationExecutionReadback struct {
	ContractVersion            string                               `json:"contract_version,omitempty"`
	Projected                  bool                                 `json:"projected"`
	Available                  bool                                 `json:"available"`
	Status                     HostActionStatus                     `json:"status,omitempty"`
	CompensationReadbackBound  bool                                 `json:"compensation_readback_bound"`
	ReadyForCloseoutReview     bool                                 `json:"ready_for_closeout_review"`
	CompensationSucceeded      bool                                 `json:"compensation_succeeded"`
	ResidualRiskRecorded       bool                                 `json:"residual_risk_recorded"`
	Result                     ObjectiveCompensationExecutionResult `json:"result,omitempty"`
	CompensationReadbackRef    DisplaySafeRef                       `json:"compensation_readback_ref,omitempty"`
	CompensationResultRef      DisplaySafeRef                       `json:"compensation_result_ref,omitempty"`
	CompensationRequestRef     DisplaySafeRef                       `json:"compensation_request_ref,omitempty"`
	FailureReviewPacketRef     DisplaySafeRef                       `json:"failure_review_packet_ref,omitempty"`
	ObjectiveRef               DisplaySafeRef                       `json:"objective_ref,omitempty"`
	FailureRef                 DisplaySafeRef                       `json:"failure_ref,omitempty"`
	CompensationRef            DisplaySafeRef                       `json:"compensation_ref,omitempty"`
	AppliedCompensationRef     DisplaySafeRef                       `json:"applied_compensation_ref,omitempty"`
	ObservedCompensationRef    DisplaySafeRef                       `json:"observed_compensation_ref,omitempty"`
	HostCompensationRunRef     DisplaySafeRef                       `json:"host_compensation_run_ref,omitempty"`
	ObservedHostRunRef         DisplaySafeRef                       `json:"observed_host_run_ref,omitempty"`
	ResidualRiskRef            DisplaySafeRef                       `json:"residual_risk_ref,omitempty"`
	ObservedResidualRiskRef    DisplaySafeRef                       `json:"observed_residual_risk_ref,omitempty"`
	ReadbackEvidenceRefs       []EvidenceRef                        `json:"readback_evidence_refs,omitempty"`
	MissingInputs              []MissingInput                       `json:"missing_inputs,omitempty"`
	BlockedReasons             []string                             `json:"blocked_reasons,omitempty"`
	FailureClass               FailureClass                         `json:"failure_class,omitempty"`
	Boundaries                 []Boundary                           `json:"boundaries,omitempty"`
	NextHostAction             NextHostAction                       `json:"next_host_action,omitempty"`
	RunnerEffect               string                               `json:"runner_effect,omitempty"`
	PromptEffect               string                               `json:"prompt_effect,omitempty"`
	CoreExecutionExecuted      bool                                 `json:"core_execution_executed"`
	RunnerDispatched           bool                                 `json:"runner_dispatched"`
	ToolExecuted               bool                                 `json:"tool_executed"`
	WorkflowDispatched         bool                                 `json:"workflow_dispatched"`
	SchedulerApplied           bool                                 `json:"scheduler_applied"`
	InstallerExecuted          bool                                 `json:"installer_executed"`
	StoreMutationExecuted      bool                                 `json:"store_mutation_executed"`
	CompensationExecutedByCore bool                                 `json:"compensation_executed_by_core"`
	RawOutputLoaded            bool                                 `json:"raw_output_loaded"`
}

func BuildObjectiveCompensationExecutorDescriptor(input ObjectiveCompensationExecutorDescriptor) ObjectiveCompensationExecutorDescriptor {
	result := input.Normalize()
	result.Status = HostActionBlocked
	result.Boundaries = MergeBoundaries(
		[]Boundary{
			"objective_compensation_executor_descriptor",
			"host_owned_compensation_executor",
			"compensation_executor_metadata_only",
			"no_compensation_execution_by_core",
		},
		result.Boundaries,
	)
	if objectiveCompensationExecutorDescriptorUnsafe(input) {
		result = objectiveCompensationExecutorDescriptorBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		result.RawOutputLoaded = true
		return result.Normalize()
	}
	if !result.Available {
		result.Status = HostActionNotReady
		result.FailureClass = firstFailureClass(result.FailureClass, FailureHostAdapterMissing)
		result.NextHostAction = firstNextHostAction(result.NextHostAction, "configure_compensation_executor")
		return result.Normalize()
	}
	if result.DescriptorRef == "" {
		result = objectiveCompensationExecutorDescriptorBlock(result, FailureEvidenceMissing, "compensation_executor_descriptor_ref_missing", "host:compensation_executor_descriptor_ref", "provide_compensation_executor_descriptor_ref")
	}
	if result.ExecutorRef == "" {
		result = objectiveCompensationExecutorDescriptorBlock(result, FailureHostAdapterMissing, "compensation_executor_ref_missing", "host:compensation_executor_ref", "configure_compensation_executor")
	}
	if result.OwnerRef == "" {
		result = objectiveCompensationExecutorDescriptorBlock(result, FailureConfigMissing, "compensation_executor_owner_ref_missing", "host:compensation_executor_owner_ref", "provide_compensation_executor_owner_ref")
	}
	if len(result.SupportedCompensationRefs) == 0 {
		result = objectiveCompensationExecutorDescriptorBlock(result, FailureUnsupportedOperation, "supported_compensation_ref_missing", "host:supported_compensation_ref", "declare_compensation_capability")
	}
	if result.IdempotencyContractRef == "" {
		result = objectiveCompensationExecutorDescriptorBlock(result, FailureConfigMissing, "compensation_idempotency_contract_missing", "contract:compensation_idempotency", "provide_compensation_idempotency_contract")
	}
	if result.ReadbackContractRef == "" {
		result = objectiveCompensationExecutorDescriptorBlock(result, FailureConfigMissing, "compensation_readback_contract_missing", "contract:compensation_readback", "provide_compensation_readback_contract")
	}
	if result.RollbackContractRef == "" {
		result = objectiveCompensationExecutorDescriptorBlock(result, FailureConfigMissing, "compensation_rollback_contract_missing", "contract:compensation_rollback", "provide_compensation_rollback_contract")
	}
	if len(result.MissingInputs) == 0 && len(result.BlockedReasons) == 0 {
		result.Status = HostActionReady
		result.NextHostAction = "host_may_build_compensation_execution_request"
		result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_compensation_execution_request")
	}
	return result.Normalize()
}

func BuildObjectiveCompensationExecutionRequest(input ObjectiveCompensationExecutionRequestInput) ObjectiveCompensationExecutionRequest {
	review := input.FailureReview.Normalize()
	descriptor := BuildObjectiveCompensationExecutorDescriptor(input.ExecutorDescriptor)
	result := ObjectiveCompensationExecutionRequest{
		ContractVersion:             ContractVersion,
		Projected:                   true,
		Available:                   review.Available && descriptor.Available,
		Status:                      HostActionBlocked,
		CompensationRequestRef:      normalizeOneDisplaySafeRef(input.CompensationRequestRef),
		FailureReviewPacketRef:      review.FailureReviewPacketRef,
		ObjectiveRef:                review.ObjectiveRef,
		FailureRef:                  review.FailureRef,
		CompensationRef:             review.CompensationRef,
		AdapterRef:                  review.AdapterRef,
		DescriptorRef:               descriptor.DescriptorRef,
		ExecutorRef:                 descriptor.ExecutorRef,
		OwnerRef:                    descriptor.OwnerRef,
		HostCompensationApprovalRef: normalizeOneDisplaySafeRef(input.HostCompensationApprovalRef),
		IdempotencyRef:              normalizeOneDisplaySafeRef(input.IdempotencyRef),
		IdempotencyContractRef:      descriptor.IdempotencyContractRef,
		ExpectedResultRef:           normalizeOneDisplaySafeRef(input.ExpectedResultRef),
		ExpectedReadbackRef:         normalizeOneDisplaySafeRef(input.ExpectedReadbackRef),
		RollbackPlanRef:             normalizeOneDisplaySafeRef(input.RollbackPlanRef),
		RollbackContractRef:         descriptor.RollbackContractRef,
		ReadbackContractRef:         descriptor.ReadbackContractRef,
		AuditRef:                    normalizeOneDisplaySafeRef(input.AuditRef),
		TimeoutPolicyRef:            descriptor.TimeoutPolicyRef,
		PolicyRefs:                  normalizeDisplaySafeRefs(append(cloneDisplaySafeRefs(descriptor.PolicyRefs), input.PolicyRefs...)),
		RequiredPolicyRefs:          cloneDisplaySafeRefs(descriptor.RequiredPolicyRefs),
		ApprovalRefs:                normalizeDisplaySafeRefs(append(append(cloneDisplaySafeRefs(descriptor.ApprovalRefs), input.ApprovalRefs...), input.HostCompensationApprovalRef)),
		RequiredApprovalRefs:        cloneDisplaySafeRefs(descriptor.RequiredApprovalRefs),
		EvidenceRefs:                MergeEvidenceRefs(review.EvidenceRefs, input.EvidenceRefs),
		FailureClass:                FailureNone,
		Boundaries: MergeBoundaries(
			[]Boundary{
				"objective_compensation_execution_request",
				"host_owned_compensation_execution_gate",
				"compensation_request_not_execution",
				"display_safe_refs_only",
				"no_compensation_execution_by_core",
				"no_runner_dispatch",
				"no_tool_execution_by_core",
				"no_workflow_dispatch_by_core",
				"no_scheduler_apply_by_core",
				"no_install_apply_by_core",
				"no_store_mutation_by_core",
			},
			input.Boundaries,
			review.Boundaries,
			descriptor.Boundaries,
		),
		NextHostAction:  "review_compensation_execution_request",
		RunnerEffect:    "none",
		PromptEffect:    "none",
		RawOutputLoaded: input.RawOutputLoaded || review.RawOutputLoaded || descriptor.RawOutputLoaded,
	}
	if objectiveCompensationExecutionRequestInputUnsafe(input) {
		result = objectiveCompensationExecutionRequestBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		result.RawOutputLoaded = true
		return result.Normalize()
	}
	if !review.ReadyForCompensationReview {
		result = objectiveCompensationExecutionRequestBlock(result, firstFailureClass(review.FailureClass, FailureEvidenceMissing), "failure_review_not_ready_for_compensation", "host:objective_failure_compensation_review", firstNextHostAction(review.NextHostAction, "review_objective_failure"))
	}
	if result.CompensationRef == "" {
		result = objectiveCompensationExecutionRequestBlock(result, FailureEvidenceMissing, "compensation_ref_missing", "host:compensation_ref", "record_residual_risk")
	}
	if descriptor.Status != HostActionReady {
		result = objectiveCompensationExecutionRequestBlock(result, firstFailureClass(descriptor.FailureClass, FailureHostAdapterMissing), "compensation_executor_not_ready", "host:compensation_executor", firstNextHostAction(descriptor.NextHostAction, "configure_compensation_executor"))
	}
	if result.CompensationRef != "" && !displaySafeRefSliceContains(descriptor.SupportedCompensationRefs, result.CompensationRef) {
		result = objectiveCompensationExecutionRequestBlock(result, FailureUnsupportedOperation, "compensation_ref_not_supported", MissingInput(result.CompensationRef), "record_residual_risk")
	}
	for _, required := range descriptor.RequiredPolicyRefs {
		if !displaySafeRefSliceContains(result.PolicyRefs, required) {
			result = objectiveCompensationExecutionRequestBlock(result, FailurePolicyBlocked, "compensation_policy_missing", MissingInput(required), "provide_compensation_policy")
		}
	}
	for _, required := range descriptor.RequiredApprovalRefs {
		if !displaySafeRefSliceContains(result.ApprovalRefs, required) {
			result = objectiveCompensationExecutionRequestBlock(result, FailureApprovalRequired, "compensation_approval_missing", MissingInput(required), "request_host_compensation_approval")
		}
	}
	for _, check := range []struct {
		ok      bool
		reason  string
		missing MissingInput
		next    NextHostAction
	}{
		{result.CompensationRequestRef != "", "compensation_request_ref_missing", "host:compensation_request_ref", "provide_compensation_execution_refs"},
		{result.HostCompensationApprovalRef != "", "host_compensation_approval_ref_missing", "host:compensation_approval_ref", "request_host_compensation_approval"},
		{result.IdempotencyRef != "", "compensation_idempotency_ref_missing", "host:compensation_idempotency_ref", "provide_compensation_idempotency_ref"},
		{result.ExpectedResultRef != "", "compensation_expected_result_ref_missing", "host:expected_compensation_result_ref", "provide_compensation_execution_refs"},
		{result.ExpectedReadbackRef != "", "compensation_expected_readback_ref_missing", "host:expected_compensation_readback_ref", "provide_compensation_execution_refs"},
		{result.RollbackPlanRef != "", "compensation_rollback_plan_ref_missing", "host:compensation_rollback_plan_ref", "provide_compensation_rollback_plan"},
		{result.AuditRef != "", "compensation_audit_ref_missing", "host:compensation_audit_ref", "provide_compensation_audit_ref"},
	} {
		if !check.ok {
			result = objectiveCompensationExecutionRequestBlock(result, FailureConfigMissing, check.reason, check.missing, check.next)
		}
	}
	if len(result.MissingInputs) == 0 && len(result.BlockedReasons) == 0 {
		result.Status = HostActionReady
		result.ReadyForHostCompensation = true
		result.HostCompensationAuthorized = true
		result.HostMayExecuteCompensation = true
		result.NextHostAction = "host_may_execute_compensation"
		result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_host_compensation_execution")
	}
	return result.Normalize()
}

func BuildObjectiveCompensationExecutionResult(input ObjectiveCompensationExecutionResultInput) ObjectiveCompensationExecutionResult {
	rawRequest := input.Request
	request := rawRequest.Normalize()
	result := ObjectiveCompensationExecutionResult{
		ContractVersion:           ContractVersion,
		Projected:                 true,
		Available:                 request.Available,
		Status:                    HostActionBlocked,
		Request:                   request,
		CompensationResultRef:     normalizeOneDisplaySafeRef(input.CompensationResultRef),
		ExpectedResultRef:         request.ExpectedResultRef,
		ExpectedReadbackRef:       request.ExpectedReadbackRef,
		CompensationRequestRef:    request.CompensationRequestRef,
		FailureReviewPacketRef:    request.FailureReviewPacketRef,
		ObjectiveRef:              request.ObjectiveRef,
		FailureRef:                firstDisplaySafeRef(input.FailureRef, request.FailureRef),
		CompensationRef:           request.CompensationRef,
		AppliedCompensationRef:    normalizeOneDisplaySafeRef(input.AppliedCompensationRef),
		ResidualRiskRef:           normalizeOneDisplaySafeRef(input.ResidualRiskRef),
		ExecutorRef:               request.ExecutorRef,
		HostCompensationRunRef:    normalizeOneDisplaySafeRef(input.HostCompensationRunRef),
		HostCompensationReported:  input.HostCompensationReported,
		HostCompensationSucceeded: input.HostCompensationSucceeded,
		HostCompensationFailed:    input.HostCompensationFailed,
		CompensationEvidenceRefs:  normalizeEvidenceRefs(input.CompensationEvidenceRefs),
		FailureClass:              FailureNone,
		Boundaries: MergeBoundaries(
			[]Boundary{
				"objective_compensation_execution_result",
				"host_owned_compensation_execution_result",
				"display_safe_refs_only",
				"no_compensation_execution_by_core",
			},
			input.Boundaries,
			request.Boundaries,
		),
		NextHostAction:  "review_compensation_execution_result",
		RunnerEffect:    "none",
		PromptEffect:    "none",
		RawOutputLoaded: input.RawOutputLoaded || request.RawOutputLoaded,
	}
	if objectiveCompensationExecutionResultUnsafe(input, rawRequest) {
		result = objectiveCompensationExecutionResultBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		result.RawOutputLoaded = true
		return result.Normalize()
	}
	if !request.ReadyForHostCompensation {
		result = objectiveCompensationExecutionResultBlock(result, firstFailureClass(request.FailureClass, FailurePolicyBlocked), "compensation_request_not_ready", "host:compensation_execution_request", firstNextHostAction(request.NextHostAction, "review_compensation_execution_request"))
	}
	if result.CompensationResultRef == "" {
		result = objectiveCompensationExecutionResultBlock(result, FailureEvidenceMissing, "compensation_result_ref_missing", "host:compensation_result_ref", "provide_compensation_result_ref")
	}
	if result.CompensationResultRef != "" && result.ExpectedResultRef != "" && result.CompensationResultRef != result.ExpectedResultRef {
		result = objectiveCompensationExecutionResultBlock(result, FailureVerificationFailed, "compensation_result_ref_mismatch", "host:compensation_result_ref", "review_compensation_execution_result")
	}
	if result.HostCompensationRunRef == "" {
		result = objectiveCompensationExecutionResultBlock(result, FailureEvidenceMissing, "host_compensation_run_ref_missing", "host:compensation_run_ref", "provide_compensation_run_ref")
	}
	if !input.HostCompensationReported {
		result = objectiveCompensationExecutionResultBlock(result, FailureEvidenceMissing, "host_compensation_result_not_reported", "host:compensation_result", "provide_compensation_result")
	}
	if input.HostCompensationSucceeded && input.HostCompensationFailed {
		result = objectiveCompensationExecutionResultBlock(result, FailureVerificationFailed, "host_compensation_result_conflict", "host:compensation_result", "review_compensation_execution_result")
	}
	switch {
	case input.HostCompensationSucceeded:
		if result.AppliedCompensationRef == "" {
			result = objectiveCompensationExecutionResultBlock(result, FailureEvidenceMissing, "applied_compensation_ref_missing", "host:applied_compensation_ref", "provide_compensation_readback_refs")
		}
		if result.AppliedCompensationRef != "" && result.AppliedCompensationRef != result.CompensationRef {
			result = objectiveCompensationExecutionResultBlock(result, FailureVerificationFailed, "applied_compensation_ref_mismatch", "host:applied_compensation_ref", "review_compensation_execution_result")
		}
	case input.HostCompensationFailed:
		if result.FailureRef == "" {
			result = objectiveCompensationExecutionResultBlock(result, FailureEvidenceMissing, "compensation_failure_ref_missing", "host:compensation_failure_ref", "record_residual_risk")
		}
		if result.ResidualRiskRef == "" {
			result = objectiveCompensationExecutionResultBlock(result, FailureEvidenceMissing, "compensation_residual_risk_ref_missing", "host:residual_risk_ref", "record_residual_risk")
		}
	default:
		result = objectiveCompensationExecutionResultBlock(result, FailureEvidenceMissing, "host_compensation_status_missing", "host:compensation_result_status", "provide_compensation_result")
	}
	if len(result.MissingInputs) == 0 && len(result.BlockedReasons) == 0 {
		result.Status = HostActionRecorded
		result.ReadyForCompensationReadback = true
		result.ReadyForCloseoutReview = true
		result.HostCompensationExecuted = true
		result.NextHostAction = "bind_compensation_readback"
		result.Boundaries = AppendBoundaries(result.Boundaries, "host_compensation_execution_recorded")
	}
	return result.Normalize()
}

func BuildObjectiveCompensationExecutionReadback(input ObjectiveCompensationExecutionReadbackInput) ObjectiveCompensationExecutionReadback {
	rawResult := input.Result
	result := rawResult.Normalize()
	readback := ObjectiveCompensationExecutionReadback{
		ContractVersion:         ContractVersion,
		Projected:               true,
		Available:               result.Available,
		Status:                  HostActionBlocked,
		Result:                  result,
		CompensationReadbackRef: normalizeOneDisplaySafeRef(input.CompensationReadbackRef),
		CompensationResultRef:   result.CompensationResultRef,
		CompensationRequestRef:  result.CompensationRequestRef,
		FailureReviewPacketRef:  result.FailureReviewPacketRef,
		ObjectiveRef:            result.ObjectiveRef,
		FailureRef:              result.FailureRef,
		CompensationRef:         result.CompensationRef,
		AppliedCompensationRef:  result.AppliedCompensationRef,
		ObservedCompensationRef: normalizeOneDisplaySafeRef(input.ObservedCompensationRef),
		HostCompensationRunRef:  result.HostCompensationRunRef,
		ObservedHostRunRef:      normalizeOneDisplaySafeRef(input.ObservedHostRunRef),
		ResidualRiskRef:         result.ResidualRiskRef,
		ObservedResidualRiskRef: normalizeOneDisplaySafeRef(input.ObservedResidualRiskRef),
		ReadbackEvidenceRefs:    normalizeEvidenceRefs(input.ReadbackEvidenceRefs),
		FailureClass:            FailureNone,
		Boundaries: MergeBoundaries(
			[]Boundary{
				"objective_compensation_execution_readback",
				"host_owned_compensation_readback",
				"display_safe_refs_only",
				"no_compensation_execution_by_core",
			},
			input.Boundaries,
			result.Boundaries,
		),
		NextHostAction:  "review_compensation_readback",
		RunnerEffect:    "none",
		PromptEffect:    "none",
		RawOutputLoaded: input.RawOutputLoaded || result.RawOutputLoaded,
	}
	if objectiveCompensationExecutionReadbackUnsafe(input, rawResult) {
		readback = objectiveCompensationExecutionReadbackBlock(readback, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		readback.RawOutputLoaded = true
		return readback.Normalize()
	}
	if !result.ReadyForCompensationReadback {
		readback = objectiveCompensationExecutionReadbackBlock(readback, firstFailureClass(result.FailureClass, FailureEvidenceMissing), "compensation_result_not_ready_for_readback", "host:compensation_result", firstNextHostAction(result.NextHostAction, "review_compensation_execution_result"))
	}
	if readback.CompensationReadbackRef == "" {
		readback = objectiveCompensationExecutionReadbackBlock(readback, FailureEvidenceMissing, "compensation_readback_ref_missing", "host:compensation_readback_ref", "provide_compensation_readback")
	}
	if result.ExpectedReadbackRef != "" && readback.CompensationReadbackRef != "" && readback.CompensationReadbackRef != result.ExpectedReadbackRef {
		readback = objectiveCompensationExecutionReadbackBlock(readback, FailureVerificationFailed, "compensation_readback_ref_mismatch", "host:compensation_readback_ref", "review_compensation_readback")
	}
	if result.HostCompensationSucceeded {
		if readback.ObservedCompensationRef == "" {
			readback = objectiveCompensationExecutionReadbackBlock(readback, FailureEvidenceMissing, "observed_compensation_ref_missing", "host:observed_compensation_ref", "provide_compensation_readback")
		}
		if readback.ObservedCompensationRef != "" && readback.ObservedCompensationRef != readback.AppliedCompensationRef {
			readback = objectiveCompensationExecutionReadbackBlock(readback, FailureVerificationFailed, "observed_compensation_ref_mismatch", "host:observed_compensation_ref", "review_compensation_readback")
		}
	}
	if result.HostCompensationFailed {
		if readback.ObservedResidualRiskRef == "" {
			readback = objectiveCompensationExecutionReadbackBlock(readback, FailureEvidenceMissing, "observed_residual_risk_ref_missing", "host:observed_residual_risk_ref", "record_residual_risk")
		}
		if readback.ObservedResidualRiskRef != "" && readback.ObservedResidualRiskRef != readback.ResidualRiskRef {
			readback = objectiveCompensationExecutionReadbackBlock(readback, FailureVerificationFailed, "observed_residual_risk_ref_mismatch", "host:observed_residual_risk_ref", "review_compensation_readback")
		}
	}
	if readback.ObservedHostRunRef == "" {
		readback = objectiveCompensationExecutionReadbackBlock(readback, FailureEvidenceMissing, "observed_compensation_run_ref_missing", "host:observed_compensation_run_ref", "provide_compensation_readback")
	}
	if readback.ObservedHostRunRef != "" && readback.ObservedHostRunRef != readback.HostCompensationRunRef {
		readback = objectiveCompensationExecutionReadbackBlock(readback, FailureVerificationFailed, "observed_compensation_run_ref_mismatch", "host:observed_compensation_run_ref", "review_compensation_readback")
	}
	if len(readback.MissingInputs) == 0 && len(readback.BlockedReasons) == 0 {
		readback.Status = HostActionRecorded
		readback.CompensationReadbackBound = true
		readback.ReadyForCloseoutReview = true
		readback.CompensationSucceeded = result.HostCompensationSucceeded
		readback.ResidualRiskRecorded = result.HostCompensationFailed && readback.ObservedResidualRiskRef != ""
		readback.NextHostAction = "continue_closeout_failure_review"
		readback.Boundaries = AppendBoundaries(readback.Boundaries, "compensation_readback_bound")
	}
	return readback.Normalize()
}

func CloneObjectiveCompensationExecutorDescriptor(in ObjectiveCompensationExecutorDescriptor) ObjectiveCompensationExecutorDescriptor {
	out := in
	out.SupportedCompensationRefs = cloneDisplaySafeRefs(in.SupportedCompensationRefs)
	out.PolicyRefs = cloneDisplaySafeRefs(in.PolicyRefs)
	out.RequiredPolicyRefs = cloneDisplaySafeRefs(in.RequiredPolicyRefs)
	out.ApprovalRefs = cloneDisplaySafeRefs(in.ApprovalRefs)
	out.RequiredApprovalRefs = cloneDisplaySafeRefs(in.RequiredApprovalRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (d ObjectiveCompensationExecutorDescriptor) Clone() ObjectiveCompensationExecutorDescriptor {
	return CloneObjectiveCompensationExecutorDescriptor(d)
}

func (d ObjectiveCompensationExecutorDescriptor) Normalize() ObjectiveCompensationExecutorDescriptor {
	out := CloneObjectiveCompensationExecutorDescriptor(d)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.DescriptorRef = normalizeOneDisplaySafeRef(out.DescriptorRef)
	out.ExecutorRef = normalizeOneDisplaySafeRef(out.ExecutorRef)
	out.OwnerRef = normalizeOneDisplaySafeRef(out.OwnerRef)
	out.SupportedCompensationRefs = normalizeDisplaySafeRefs(out.SupportedCompensationRefs)
	out.IdempotencyContractRef = normalizeOneDisplaySafeRef(out.IdempotencyContractRef)
	out.ReadbackContractRef = normalizeOneDisplaySafeRef(out.ReadbackContractRef)
	out.RollbackContractRef = normalizeOneDisplaySafeRef(out.RollbackContractRef)
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
	if out.RawOutputLoaded {
		out.Status = HostActionBlocked
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		if out.FailureClass == FailureNone {
			out.FailureClass = FailureEvidenceWeak
		}
		if out.NextHostAction == "" {
			out.NextHostAction = "provide_display_safe_refs"
		}
	}
	return out
}

func CloneObjectiveCompensationExecutionRequest(in ObjectiveCompensationExecutionRequest) ObjectiveCompensationExecutionRequest {
	out := in
	out.PolicyRefs = cloneDisplaySafeRefs(in.PolicyRefs)
	out.RequiredPolicyRefs = cloneDisplaySafeRefs(in.RequiredPolicyRefs)
	out.ApprovalRefs = cloneDisplaySafeRefs(in.ApprovalRefs)
	out.RequiredApprovalRefs = cloneDisplaySafeRefs(in.RequiredApprovalRefs)
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (r ObjectiveCompensationExecutionRequest) Clone() ObjectiveCompensationExecutionRequest {
	return CloneObjectiveCompensationExecutionRequest(r)
}

func (r ObjectiveCompensationExecutionRequest) Normalize() ObjectiveCompensationExecutionRequest {
	out := CloneObjectiveCompensationExecutionRequest(r)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.CompensationRequestRef = normalizeOneDisplaySafeRef(out.CompensationRequestRef)
	out.FailureReviewPacketRef = normalizeOneDisplaySafeRef(out.FailureReviewPacketRef)
	out.ObjectiveRef = normalizeOneDisplaySafeRef(out.ObjectiveRef)
	out.FailureRef = normalizeOneDisplaySafeRef(out.FailureRef)
	out.CompensationRef = normalizeOneDisplaySafeRef(out.CompensationRef)
	out.AdapterRef = normalizeOneDisplaySafeRef(out.AdapterRef)
	out.DescriptorRef = normalizeOneDisplaySafeRef(out.DescriptorRef)
	out.ExecutorRef = normalizeOneDisplaySafeRef(out.ExecutorRef)
	out.OwnerRef = normalizeOneDisplaySafeRef(out.OwnerRef)
	out.HostCompensationApprovalRef = normalizeOneDisplaySafeRef(out.HostCompensationApprovalRef)
	out.IdempotencyRef = normalizeOneDisplaySafeRef(out.IdempotencyRef)
	out.IdempotencyContractRef = normalizeOneDisplaySafeRef(out.IdempotencyContractRef)
	out.ExpectedResultRef = normalizeOneDisplaySafeRef(out.ExpectedResultRef)
	out.ExpectedReadbackRef = normalizeOneDisplaySafeRef(out.ExpectedReadbackRef)
	out.RollbackPlanRef = normalizeOneDisplaySafeRef(out.RollbackPlanRef)
	out.RollbackContractRef = normalizeOneDisplaySafeRef(out.RollbackContractRef)
	out.ReadbackContractRef = normalizeOneDisplaySafeRef(out.ReadbackContractRef)
	out.AuditRef = normalizeOneDisplaySafeRef(out.AuditRef)
	out.TimeoutPolicyRef = normalizeOneDisplaySafeRef(out.TimeoutPolicyRef)
	out.PolicyRefs = normalizeDisplaySafeRefs(out.PolicyRefs)
	out.RequiredPolicyRefs = normalizeDisplaySafeRefs(out.RequiredPolicyRefs)
	out.ApprovalRefs = normalizeDisplaySafeRefs(out.ApprovalRefs)
	out.RequiredApprovalRefs = normalizeDisplaySafeRefs(out.RequiredApprovalRefs)
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
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
	objectiveCompensationClearCoreEffects(&out.CoreExecutionExecuted, &out.RunnerDispatched, &out.ToolExecuted, &out.WorkflowDispatched, &out.SchedulerApplied, &out.InstallerExecuted, &out.StoreMutationExecuted, &out.CompensationExecutedByCore)
	if out.RawOutputLoaded {
		out.Status = HostActionBlocked
		out.ReadyForHostCompensation = false
		out.HostCompensationAuthorized = false
		out.HostMayExecuteCompensation = false
		if out.FailureClass == FailureNone {
			out.FailureClass = FailureEvidenceWeak
		}
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		if out.NextHostAction == "" {
			out.NextHostAction = "provide_display_safe_refs"
		}
	}
	out.ReadyForHostCompensation = out.ReadyForHostCompensation &&
		out.Status == HostActionReady &&
		out.CompensationRequestRef != "" &&
		out.FailureReviewPacketRef != "" &&
		out.FailureRef != "" &&
		out.CompensationRef != "" &&
		out.ExecutorRef != "" &&
		out.HostCompensationApprovalRef != "" &&
		out.IdempotencyRef != "" &&
		out.ExpectedResultRef != "" &&
		out.ExpectedReadbackRef != "" &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.RawOutputLoaded
	out.HostCompensationAuthorized = out.HostCompensationAuthorized && out.ReadyForHostCompensation
	out.HostMayExecuteCompensation = out.HostMayExecuteCompensation && out.ReadyForHostCompensation
	return out
}

func CloneObjectiveCompensationExecutionResult(in ObjectiveCompensationExecutionResult) ObjectiveCompensationExecutionResult {
	out := in
	out.Request = in.Request.Clone()
	out.CompensationEvidenceRefs = cloneEvidenceRefs(in.CompensationEvidenceRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (r ObjectiveCompensationExecutionResult) Clone() ObjectiveCompensationExecutionResult {
	return CloneObjectiveCompensationExecutionResult(r)
}

func (r ObjectiveCompensationExecutionResult) Normalize() ObjectiveCompensationExecutionResult {
	out := CloneObjectiveCompensationExecutionResult(r)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.Request = out.Request.Normalize()
	out.CompensationResultRef = normalizeOneDisplaySafeRef(out.CompensationResultRef)
	out.ExpectedResultRef = normalizeOneDisplaySafeRef(out.ExpectedResultRef)
	out.ExpectedReadbackRef = normalizeOneDisplaySafeRef(out.ExpectedReadbackRef)
	out.CompensationRequestRef = normalizeOneDisplaySafeRef(out.CompensationRequestRef)
	out.FailureReviewPacketRef = normalizeOneDisplaySafeRef(out.FailureReviewPacketRef)
	out.ObjectiveRef = normalizeOneDisplaySafeRef(out.ObjectiveRef)
	out.FailureRef = normalizeOneDisplaySafeRef(out.FailureRef)
	out.CompensationRef = normalizeOneDisplaySafeRef(out.CompensationRef)
	out.AppliedCompensationRef = normalizeOneDisplaySafeRef(out.AppliedCompensationRef)
	out.ResidualRiskRef = normalizeOneDisplaySafeRef(out.ResidualRiskRef)
	out.ExecutorRef = normalizeOneDisplaySafeRef(out.ExecutorRef)
	out.HostCompensationRunRef = normalizeOneDisplaySafeRef(out.HostCompensationRunRef)
	out.CompensationEvidenceRefs = normalizeEvidenceRefs(out.CompensationEvidenceRefs)
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
	objectiveCompensationClearCoreEffects(&out.CoreExecutionExecuted, &out.RunnerDispatched, &out.ToolExecuted, &out.WorkflowDispatched, &out.SchedulerApplied, &out.InstallerExecuted, &out.StoreMutationExecuted, &out.CompensationExecutedByCore)
	if out.RawOutputLoaded || out.Request.RawOutputLoaded {
		out.RawOutputLoaded = true
		out.Status = HostActionBlocked
		out.ReadyForCompensationReadback = false
		out.ReadyForCloseoutReview = false
		if out.FailureClass == FailureNone {
			out.FailureClass = FailureEvidenceWeak
		}
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		if out.NextHostAction == "" {
			out.NextHostAction = "provide_display_safe_refs"
		}
	}
	out.HostCompensationExecuted = !out.RawOutputLoaded && out.HostCompensationReported && (out.HostCompensationSucceeded || out.HostCompensationFailed)
	out.ReadyForCompensationReadback = out.ReadyForCompensationReadback &&
		out.Status == HostActionRecorded &&
		out.Request.ReadyForHostCompensation &&
		out.HostCompensationExecuted &&
		out.CompensationResultRef != "" &&
		out.HostCompensationRunRef != "" &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.RawOutputLoaded
	out.ReadyForCloseoutReview = out.ReadyForCloseoutReview && out.ReadyForCompensationReadback
	return out
}

func CloneObjectiveCompensationExecutionReadback(in ObjectiveCompensationExecutionReadback) ObjectiveCompensationExecutionReadback {
	out := in
	out.Result = in.Result.Clone()
	out.ReadbackEvidenceRefs = cloneEvidenceRefs(in.ReadbackEvidenceRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (r ObjectiveCompensationExecutionReadback) Clone() ObjectiveCompensationExecutionReadback {
	return CloneObjectiveCompensationExecutionReadback(r)
}

func (r ObjectiveCompensationExecutionReadback) Normalize() ObjectiveCompensationExecutionReadback {
	out := CloneObjectiveCompensationExecutionReadback(r)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.Result = out.Result.Normalize()
	out.CompensationReadbackRef = normalizeOneDisplaySafeRef(out.CompensationReadbackRef)
	out.CompensationResultRef = normalizeOneDisplaySafeRef(out.CompensationResultRef)
	out.CompensationRequestRef = normalizeOneDisplaySafeRef(out.CompensationRequestRef)
	out.FailureReviewPacketRef = normalizeOneDisplaySafeRef(out.FailureReviewPacketRef)
	out.ObjectiveRef = normalizeOneDisplaySafeRef(out.ObjectiveRef)
	out.FailureRef = normalizeOneDisplaySafeRef(out.FailureRef)
	out.CompensationRef = normalizeOneDisplaySafeRef(out.CompensationRef)
	out.AppliedCompensationRef = normalizeOneDisplaySafeRef(out.AppliedCompensationRef)
	out.ObservedCompensationRef = normalizeOneDisplaySafeRef(out.ObservedCompensationRef)
	out.HostCompensationRunRef = normalizeOneDisplaySafeRef(out.HostCompensationRunRef)
	out.ObservedHostRunRef = normalizeOneDisplaySafeRef(out.ObservedHostRunRef)
	out.ResidualRiskRef = normalizeOneDisplaySafeRef(out.ResidualRiskRef)
	out.ObservedResidualRiskRef = normalizeOneDisplaySafeRef(out.ObservedResidualRiskRef)
	out.ReadbackEvidenceRefs = normalizeEvidenceRefs(out.ReadbackEvidenceRefs)
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
	objectiveCompensationClearCoreEffects(&out.CoreExecutionExecuted, &out.RunnerDispatched, &out.ToolExecuted, &out.WorkflowDispatched, &out.SchedulerApplied, &out.InstallerExecuted, &out.StoreMutationExecuted, &out.CompensationExecutedByCore)
	if out.RawOutputLoaded || out.Result.RawOutputLoaded {
		out.RawOutputLoaded = true
		out.Status = HostActionBlocked
		out.CompensationReadbackBound = false
		out.ReadyForCloseoutReview = false
		if out.FailureClass == FailureNone {
			out.FailureClass = FailureEvidenceWeak
		}
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		if out.NextHostAction == "" {
			out.NextHostAction = "provide_display_safe_refs"
		}
	}
	out.CompensationReadbackBound = out.CompensationReadbackBound &&
		out.Status == HostActionRecorded &&
		out.Result.ReadyForCompensationReadback &&
		out.CompensationReadbackRef != "" &&
		out.CompensationResultRef != "" &&
		out.ObservedHostRunRef != "" &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.RawOutputLoaded
	out.ReadyForCloseoutReview = out.ReadyForCloseoutReview && out.CompensationReadbackBound
	out.CompensationSucceeded = out.CompensationSucceeded && out.CompensationReadbackBound
	out.ResidualRiskRecorded = out.ResidualRiskRecorded && out.CompensationReadbackBound
	return out
}

func objectiveCompensationExecutorDescriptorBlock(result ObjectiveCompensationExecutorDescriptor, failure FailureClass, reason string, missing MissingInput, next NextHostAction) ObjectiveCompensationExecutorDescriptor {
	result.Status = HostActionBlocked
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = firstNextHostAction(next, result.NextHostAction)
	result.Boundaries = AppendBoundaries(result.Boundaries, "objective_compensation_executor_descriptor_blocked")
	return result
}

func objectiveCompensationExecutionRequestBlock(result ObjectiveCompensationExecutionRequest, failure FailureClass, reason string, missing MissingInput, next NextHostAction) ObjectiveCompensationExecutionRequest {
	result.Status = HostActionBlocked
	result.ReadyForHostCompensation = false
	result.HostCompensationAuthorized = false
	result.HostMayExecuteCompensation = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = firstNextHostAction(next, result.NextHostAction)
	result.Boundaries = AppendBoundaries(result.Boundaries, "objective_compensation_execution_request_blocked")
	return result
}

func objectiveCompensationExecutionResultBlock(result ObjectiveCompensationExecutionResult, failure FailureClass, reason string, missing MissingInput, next NextHostAction) ObjectiveCompensationExecutionResult {
	result.Status = HostActionBlocked
	result.ReadyForCompensationReadback = false
	result.ReadyForCloseoutReview = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = firstNextHostAction(next, result.NextHostAction)
	result.Boundaries = AppendBoundaries(result.Boundaries, "objective_compensation_execution_result_blocked")
	return result
}

func objectiveCompensationExecutionReadbackBlock(result ObjectiveCompensationExecutionReadback, failure FailureClass, reason string, missing MissingInput, next NextHostAction) ObjectiveCompensationExecutionReadback {
	result.Status = HostActionBlocked
	result.CompensationReadbackBound = false
	result.ReadyForCloseoutReview = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = firstNextHostAction(next, result.NextHostAction)
	result.Boundaries = AppendBoundaries(result.Boundaries, "objective_compensation_execution_readback_blocked")
	return result
}

func objectiveCompensationExecutorDescriptorUnsafe(input ObjectiveCompensationExecutorDescriptor) bool {
	return displaySafeRefRejected(input.DescriptorRef) ||
		displaySafeRefRejected(input.ExecutorRef) ||
		displaySafeRefRejected(input.OwnerRef) ||
		objectiveCompensationDisplaySafeRefsRejected(input.SupportedCompensationRefs) ||
		displaySafeRefRejected(input.IdempotencyContractRef) ||
		displaySafeRefRejected(input.ReadbackContractRef) ||
		displaySafeRefRejected(input.RollbackContractRef) ||
		displaySafeRefRejected(input.TimeoutPolicyRef) ||
		objectiveCompensationDisplaySafeRefsRejected(input.PolicyRefs) ||
		objectiveCompensationDisplaySafeRefsRejected(input.RequiredPolicyRefs) ||
		objectiveCompensationDisplaySafeRefsRejected(input.ApprovalRefs) ||
		objectiveCompensationDisplaySafeRefsRejected(input.RequiredApprovalRefs) ||
		input.RawOutputLoaded
}

func objectiveCompensationExecutionRequestInputUnsafe(input ObjectiveCompensationExecutionRequestInput) bool {
	return displaySafeRefRejected(input.CompensationRequestRef) ||
		displaySafeRefRejected(input.HostCompensationApprovalRef) ||
		displaySafeRefRejected(input.IdempotencyRef) ||
		displaySafeRefRejected(input.ExpectedResultRef) ||
		displaySafeRefRejected(input.ExpectedReadbackRef) ||
		displaySafeRefRejected(input.RollbackPlanRef) ||
		displaySafeRefRejected(input.AuditRef) ||
		objectiveCompensationDisplaySafeRefsRejected(input.PolicyRefs) ||
		objectiveCompensationDisplaySafeRefsRejected(input.ApprovalRefs) ||
		evidenceRefsRejected(input.EvidenceRefs) ||
		productionAdapterObjectiveCloseoutFailureReviewPacketUnsafeOutput(input.FailureReview) ||
		objectiveCompensationExecutorDescriptorUnsafe(input.ExecutorDescriptor) ||
		input.RawOutputLoaded
}

func objectiveCompensationExecutionRequestOutputUnsafe(input ObjectiveCompensationExecutionRequest) bool {
	return displaySafeRefRejected(input.CompensationRequestRef) ||
		displaySafeRefRejected(input.FailureReviewPacketRef) ||
		displaySafeRefRejected(input.ObjectiveRef) ||
		displaySafeRefRejected(input.FailureRef) ||
		displaySafeRefRejected(input.CompensationRef) ||
		displaySafeRefRejected(input.AdapterRef) ||
		displaySafeRefRejected(input.DescriptorRef) ||
		displaySafeRefRejected(input.ExecutorRef) ||
		displaySafeRefRejected(input.OwnerRef) ||
		displaySafeRefRejected(input.HostCompensationApprovalRef) ||
		displaySafeRefRejected(input.IdempotencyRef) ||
		displaySafeRefRejected(input.IdempotencyContractRef) ||
		displaySafeRefRejected(input.ExpectedResultRef) ||
		displaySafeRefRejected(input.ExpectedReadbackRef) ||
		displaySafeRefRejected(input.RollbackPlanRef) ||
		displaySafeRefRejected(input.RollbackContractRef) ||
		displaySafeRefRejected(input.ReadbackContractRef) ||
		displaySafeRefRejected(input.AuditRef) ||
		displaySafeRefRejected(input.TimeoutPolicyRef) ||
		objectiveCompensationDisplaySafeRefsRejected(input.PolicyRefs) ||
		objectiveCompensationDisplaySafeRefsRejected(input.RequiredPolicyRefs) ||
		objectiveCompensationDisplaySafeRefsRejected(input.ApprovalRefs) ||
		objectiveCompensationDisplaySafeRefsRejected(input.RequiredApprovalRefs) ||
		evidenceRefsRejected(input.EvidenceRefs) ||
		input.RawOutputLoaded
}

func objectiveCompensationExecutionResultUnsafe(input ObjectiveCompensationExecutionResultInput, request ObjectiveCompensationExecutionRequest) bool {
	return displaySafeRefRejected(input.CompensationResultRef) ||
		displaySafeRefRejected(input.HostCompensationRunRef) ||
		displaySafeRefRejected(input.AppliedCompensationRef) ||
		displaySafeRefRejected(input.FailureRef) ||
		displaySafeRefRejected(input.ResidualRiskRef) ||
		evidenceRefsRejected(input.CompensationEvidenceRefs) ||
		objectiveCompensationExecutionRequestOutputUnsafe(request) ||
		input.RawOutputLoaded
}

func objectiveCompensationExecutionResultOutputUnsafe(input ObjectiveCompensationExecutionResult) bool {
	return objectiveCompensationExecutionRequestOutputUnsafe(input.Request) ||
		displaySafeRefRejected(input.CompensationResultRef) ||
		displaySafeRefRejected(input.ExpectedResultRef) ||
		displaySafeRefRejected(input.ExpectedReadbackRef) ||
		displaySafeRefRejected(input.CompensationRequestRef) ||
		displaySafeRefRejected(input.FailureReviewPacketRef) ||
		displaySafeRefRejected(input.ObjectiveRef) ||
		displaySafeRefRejected(input.FailureRef) ||
		displaySafeRefRejected(input.CompensationRef) ||
		displaySafeRefRejected(input.AppliedCompensationRef) ||
		displaySafeRefRejected(input.ResidualRiskRef) ||
		displaySafeRefRejected(input.ExecutorRef) ||
		displaySafeRefRejected(input.HostCompensationRunRef) ||
		evidenceRefsRejected(input.CompensationEvidenceRefs) ||
		input.RawOutputLoaded
}

func objectiveCompensationExecutionReadbackUnsafe(input ObjectiveCompensationExecutionReadbackInput, result ObjectiveCompensationExecutionResult) bool {
	return displaySafeRefRejected(input.CompensationReadbackRef) ||
		displaySafeRefRejected(input.ObservedCompensationRef) ||
		displaySafeRefRejected(input.ObservedHostRunRef) ||
		displaySafeRefRejected(input.ObservedResidualRiskRef) ||
		evidenceRefsRejected(input.ReadbackEvidenceRefs) ||
		objectiveCompensationExecutionResultOutputUnsafe(result) ||
		input.RawOutputLoaded
}

func objectiveCompensationClearCoreEffects(flags ...*bool) {
	for _, flag := range flags {
		if flag != nil {
			*flag = false
		}
	}
}

func objectiveCompensationDisplaySafeRefsRejected(refs []DisplaySafeRef) bool {
	for _, ref := range refs {
		if displaySafeRefRejected(ref) {
			return true
		}
	}
	return false
}
