package controlcontract

import "strings"

type HostOwnedCapabilityApplyAdapterReadinessInput struct {
	Request                HostOwnedCapabilityApplyRequest `json:"request,omitempty"`
	AdapterRef             DisplaySafeRef                  `json:"adapter_ref,omitempty"`
	AdapterVersionRef      DisplaySafeRef                  `json:"adapter_version_ref,omitempty"`
	AdapterCapabilityRef   DisplaySafeRef                  `json:"adapter_capability_ref,omitempty"`
	AdapterContractRef     DisplaySafeRef                  `json:"adapter_contract_ref,omitempty"`
	HostConfirmationRef    DisplaySafeRef                  `json:"host_confirmation_ref,omitempty"`
	AdapterDryRunRef       DisplaySafeRef                  `json:"adapter_dry_run_ref,omitempty"`
	InvocationRef          DisplaySafeRef                  `json:"invocation_ref,omitempty"`
	ResultBindingRef       DisplaySafeRef                  `json:"result_binding_ref,omitempty"`
	ReadbackBindingRef     DisplaySafeRef                  `json:"readback_binding_ref,omitempty"`
	IdempotencyRef         DisplaySafeRef                  `json:"idempotency_ref,omitempty"`
	RollbackRef            DisplaySafeRef                  `json:"rollback_ref,omitempty"`
	FailureBindingRef      DisplaySafeRef                  `json:"failure_binding_ref,omitempty"`
	CompensationBindingRef DisplaySafeRef                  `json:"compensation_binding_ref,omitempty"`
	ApprovalRefs           []DisplaySafeRef                `json:"approval_refs,omitempty"`
	PolicyRefs             []DisplaySafeRef                `json:"policy_refs,omitempty"`
	EvidenceRefs           []EvidenceRef                   `json:"evidence_refs,omitempty"`
	DecisionBasis          []DisplaySafeRef                `json:"decision_basis,omitempty"`
	Boundaries             []Boundary                      `json:"boundaries,omitempty"`
	RawOutputLoaded        bool                            `json:"raw_output_loaded"`
}

type HostOwnedCapabilityApplyAdapterReadiness struct {
	ContractVersion                           string                          `json:"contract_version,omitempty"`
	Projected                                 bool                            `json:"projected"`
	Status                                    HostActionStatus                `json:"status,omitempty"`
	ReadyForHostCapabilityAdapterInvocation   bool                            `json:"ready_for_host_capability_adapter_invocation"`
	HostCapabilityAdapterInvocationAuthorized bool                            `json:"host_capability_adapter_invocation_authorized"`
	HostMayInvokeCapabilityAdapter            bool                            `json:"host_may_invoke_capability_adapter"`
	Request                                   HostOwnedCapabilityApplyRequest `json:"request,omitempty"`
	Action                                    CapabilityApplyAction           `json:"action,omitempty"`
	CapabilityApplyRequestRef                 DisplaySafeRef                  `json:"capability_apply_request_ref,omitempty"`
	CapabilityAdapterRef                      DisplaySafeRef                  `json:"capability_adapter_ref,omitempty"`
	ExpectedCapabilityRef                     DisplaySafeRef                  `json:"expected_capability_ref,omitempty"`
	ExpectedCapabilityStateRef                DisplaySafeRef                  `json:"expected_capability_state_ref,omitempty"`
	ExpectedCapabilityResultRef               DisplaySafeRef                  `json:"expected_capability_result_ref,omitempty"`
	ExpectedReadbackRef                       DisplaySafeRef                  `json:"expected_readback_ref,omitempty"`
	RollbackPathRef                           DisplaySafeRef                  `json:"rollback_path_ref,omitempty"`
	AdapterRef                                DisplaySafeRef                  `json:"adapter_ref,omitempty"`
	AdapterVersionRef                         DisplaySafeRef                  `json:"adapter_version_ref,omitempty"`
	AdapterCapabilityRef                      DisplaySafeRef                  `json:"adapter_capability_ref,omitempty"`
	AdapterContractRef                        DisplaySafeRef                  `json:"adapter_contract_ref,omitempty"`
	HostConfirmationRef                       DisplaySafeRef                  `json:"host_confirmation_ref,omitempty"`
	AdapterDryRunRef                          DisplaySafeRef                  `json:"adapter_dry_run_ref,omitempty"`
	InvocationRef                             DisplaySafeRef                  `json:"invocation_ref,omitempty"`
	ResultBindingRef                          DisplaySafeRef                  `json:"result_binding_ref,omitempty"`
	ReadbackBindingRef                        DisplaySafeRef                  `json:"readback_binding_ref,omitempty"`
	IdempotencyRef                            DisplaySafeRef                  `json:"idempotency_ref,omitempty"`
	RollbackRef                               DisplaySafeRef                  `json:"rollback_ref,omitempty"`
	FailureBindingRef                         DisplaySafeRef                  `json:"failure_binding_ref,omitempty"`
	CompensationBindingRef                    DisplaySafeRef                  `json:"compensation_binding_ref,omitempty"`
	ApprovalRefs                              []DisplaySafeRef                `json:"approval_refs,omitempty"`
	PolicyRefs                                []DisplaySafeRef                `json:"policy_refs,omitempty"`
	EvidenceRefs                              []EvidenceRef                   `json:"evidence_refs,omitempty"`
	FailureClass                              FailureClass                    `json:"failure_class,omitempty"`
	BlockedReasons                            []string                        `json:"blocked_reasons,omitempty"`
	MissingInputs                             []MissingInput                  `json:"missing_inputs,omitempty"`
	DecisionBasis                             []DisplaySafeRef                `json:"decision_basis,omitempty"`
	Boundaries                                []Boundary                      `json:"boundaries,omitempty"`
	NextHostAction                            NextHostAction                  `json:"next_host_action,omitempty"`
	RunnerEffect                              string                          `json:"runner_effect,omitempty"`
	PromptEffect                              string                          `json:"prompt_effect,omitempty"`
	RuntimeEffect                             string                          `json:"runtime_effect,omitempty"`
	CoreInvocationExecuted                    bool                            `json:"core_invocation_executed"`
	InstallerExecutedByCore                   bool                            `json:"installer_executed_by_core"`
	InstallExecutedByCore                     bool                            `json:"install_executed_by_core"`
	EnableExecutedByCore                      bool                            `json:"enable_executed_by_core"`
	PackageManagerExecutedByCore              bool                            `json:"package_manager_executed_by_core"`
	SkillWriteByCore                          bool                            `json:"skill_write_by_core"`
	RuntimeReloadByCore                       bool                            `json:"runtime_reload_by_core"`
	RunnerDispatched                          bool                            `json:"runner_dispatched"`
	RuntimeAdapterExecuted                    bool                            `json:"runtime_adapter_executed"`
	WorkflowDispatched                        bool                            `json:"workflow_dispatched"`
	WorkerDispatched                          bool                            `json:"worker_dispatched"`
	StoreMutationExecuted                     bool                            `json:"store_mutation_executed"`
	CompensationExecuted                      bool                            `json:"compensation_executed"`
	RawOutputLoaded                           bool                            `json:"raw_output_loaded"`
}

type HostOwnedCapabilityApplyAdapterInvocationInput struct {
	Readiness                      HostOwnedCapabilityApplyAdapterReadiness `json:"readiness,omitempty"`
	InvocationReportRef            DisplaySafeRef                           `json:"invocation_report_ref,omitempty"`
	ObservedInvocationRef          DisplaySafeRef                           `json:"observed_invocation_ref,omitempty"`
	HostCapabilityAdapterRunRef    DisplaySafeRef                           `json:"host_capability_adapter_run_ref,omitempty"`
	CapabilityApplyResultRef       DisplaySafeRef                           `json:"capability_apply_result_ref,omitempty"`
	CapabilityReadbackRef          DisplaySafeRef                           `json:"capability_readback_ref,omitempty"`
	AppliedCapabilityRef           DisplaySafeRef                           `json:"applied_capability_ref,omitempty"`
	AppliedCapabilityStateRef      DisplaySafeRef                           `json:"applied_capability_state_ref,omitempty"`
	FailureRef                     DisplaySafeRef                           `json:"failure_ref,omitempty"`
	CompensationRef                DisplaySafeRef                           `json:"compensation_ref,omitempty"`
	HostAdapterInvocationReported  bool                                     `json:"host_adapter_invocation_reported"`
	HostAdapterInvocationCompleted bool                                     `json:"host_adapter_invocation_completed"`
	HostAdapterInvocationFailed    bool                                     `json:"host_adapter_invocation_failed"`
	CapabilityEvidenceRefs         []DisplaySafeRef                         `json:"capability_evidence_refs,omitempty"`
	Boundaries                     []Boundary                               `json:"boundaries,omitempty"`
	RawOutputLoaded                bool                                     `json:"raw_output_loaded"`
}

type HostOwnedCapabilityApplyAdapterInvocation struct {
	ContractVersion                string                                   `json:"contract_version,omitempty"`
	Projected                      bool                                     `json:"projected"`
	Status                         HostActionStatus                         `json:"status,omitempty"`
	ReadyForCapabilityApplyResult  bool                                     `json:"ready_for_capability_apply_result"`
	ReadyForFailureReview          bool                                     `json:"ready_for_failure_review"`
	HostAdapterInvocationReported  bool                                     `json:"host_adapter_invocation_reported"`
	HostAdapterInvocationCompleted bool                                     `json:"host_adapter_invocation_completed"`
	HostAdapterInvocationFailed    bool                                     `json:"host_adapter_invocation_failed"`
	Readiness                      HostOwnedCapabilityApplyAdapterReadiness `json:"readiness,omitempty"`
	Action                         CapabilityApplyAction                    `json:"action,omitempty"`
	CapabilityApplyRequestRef      DisplaySafeRef                           `json:"capability_apply_request_ref,omitempty"`
	CapabilityAdapterRef           DisplaySafeRef                           `json:"capability_adapter_ref,omitempty"`
	ExpectedCapabilityRef          DisplaySafeRef                           `json:"expected_capability_ref,omitempty"`
	ExpectedCapabilityStateRef     DisplaySafeRef                           `json:"expected_capability_state_ref,omitempty"`
	ExpectedCapabilityResultRef    DisplaySafeRef                           `json:"expected_capability_result_ref,omitempty"`
	ExpectedReadbackRef            DisplaySafeRef                           `json:"expected_readback_ref,omitempty"`
	RollbackPathRef                DisplaySafeRef                           `json:"rollback_path_ref,omitempty"`
	AdapterRef                     DisplaySafeRef                           `json:"adapter_ref,omitempty"`
	AdapterVersionRef              DisplaySafeRef                           `json:"adapter_version_ref,omitempty"`
	InvocationRef                  DisplaySafeRef                           `json:"invocation_ref,omitempty"`
	InvocationReportRef            DisplaySafeRef                           `json:"invocation_report_ref,omitempty"`
	ObservedInvocationRef          DisplaySafeRef                           `json:"observed_invocation_ref,omitempty"`
	HostCapabilityAdapterRunRef    DisplaySafeRef                           `json:"host_capability_adapter_run_ref,omitempty"`
	CapabilityApplyResultRef       DisplaySafeRef                           `json:"capability_apply_result_ref,omitempty"`
	CapabilityReadbackRef          DisplaySafeRef                           `json:"capability_readback_ref,omitempty"`
	AppliedCapabilityRef           DisplaySafeRef                           `json:"applied_capability_ref,omitempty"`
	AppliedCapabilityStateRef      DisplaySafeRef                           `json:"applied_capability_state_ref,omitempty"`
	FailureRef                     DisplaySafeRef                           `json:"failure_ref,omitempty"`
	CompensationRef                DisplaySafeRef                           `json:"compensation_ref,omitempty"`
	CapabilityEvidenceRefs         []DisplaySafeRef                         `json:"capability_evidence_refs,omitempty"`
	FailureClass                   FailureClass                             `json:"failure_class,omitempty"`
	BlockedReasons                 []string                                 `json:"blocked_reasons,omitempty"`
	MissingInputs                  []MissingInput                           `json:"missing_inputs,omitempty"`
	Boundaries                     []Boundary                               `json:"boundaries,omitempty"`
	NextHostAction                 NextHostAction                           `json:"next_host_action,omitempty"`
	RunnerEffect                   string                                   `json:"runner_effect,omitempty"`
	PromptEffect                   string                                   `json:"prompt_effect,omitempty"`
	RuntimeEffect                  string                                   `json:"runtime_effect,omitempty"`
	CoreInvocationExecuted         bool                                     `json:"core_invocation_executed"`
	InstallerExecutedByCore        bool                                     `json:"installer_executed_by_core"`
	InstallExecutedByCore          bool                                     `json:"install_executed_by_core"`
	EnableExecutedByCore           bool                                     `json:"enable_executed_by_core"`
	PackageManagerExecutedByCore   bool                                     `json:"package_manager_executed_by_core"`
	SkillWriteByCore               bool                                     `json:"skill_write_by_core"`
	RuntimeReloadByCore            bool                                     `json:"runtime_reload_by_core"`
	RunnerDispatched               bool                                     `json:"runner_dispatched"`
	RuntimeAdapterExecuted         bool                                     `json:"runtime_adapter_executed"`
	WorkflowDispatched             bool                                     `json:"workflow_dispatched"`
	WorkerDispatched               bool                                     `json:"worker_dispatched"`
	StoreMutationExecuted          bool                                     `json:"store_mutation_executed"`
	CompensationExecuted           bool                                     `json:"compensation_executed"`
	RawOutputLoaded                bool                                     `json:"raw_output_loaded"`
}

func BuildHostOwnedCapabilityApplyAdapterReadiness(input HostOwnedCapabilityApplyAdapterReadinessInput) HostOwnedCapabilityApplyAdapterReadiness {
	request := input.Request.Normalize()
	result := HostOwnedCapabilityApplyAdapterReadiness{
		ContractVersion:             ContractVersion,
		Projected:                   true,
		Status:                      HostActionBlocked,
		Request:                     request,
		Action:                      request.Action,
		CapabilityApplyRequestRef:   request.CapabilityApplyRequestRef,
		CapabilityAdapterRef:        request.CapabilityAdapterRef,
		ExpectedCapabilityRef:       request.ExpectedCapabilityRef,
		ExpectedCapabilityStateRef:  request.ExpectedCapabilityStateRef,
		ExpectedCapabilityResultRef: request.ExpectedCapabilityResultRef,
		ExpectedReadbackRef:         request.ExpectedReadbackRef,
		RollbackPathRef:             request.RollbackPathRef,
		AdapterRef:                  normalizeOneDisplaySafeRef(input.AdapterRef),
		AdapterVersionRef:           normalizeOneDisplaySafeRef(input.AdapterVersionRef),
		AdapterCapabilityRef:        normalizeOneDisplaySafeRef(input.AdapterCapabilityRef),
		AdapterContractRef:          normalizeOneDisplaySafeRef(input.AdapterContractRef),
		HostConfirmationRef:         normalizeOneDisplaySafeRef(input.HostConfirmationRef),
		AdapterDryRunRef:            normalizeOneDisplaySafeRef(input.AdapterDryRunRef),
		InvocationRef:               normalizeOneDisplaySafeRef(input.InvocationRef),
		ResultBindingRef:            normalizeOneDisplaySafeRef(input.ResultBindingRef),
		ReadbackBindingRef:          normalizeOneDisplaySafeRef(input.ReadbackBindingRef),
		IdempotencyRef:              normalizeOneDisplaySafeRef(input.IdempotencyRef),
		RollbackRef:                 normalizeOneDisplaySafeRef(input.RollbackRef),
		FailureBindingRef:           normalizeOneDisplaySafeRef(input.FailureBindingRef),
		CompensationBindingRef:      normalizeOneDisplaySafeRef(input.CompensationBindingRef),
		ApprovalRefs:                normalizeDisplaySafeRefs(append(cloneDisplaySafeRefs(request.ApprovalRefs), input.ApprovalRefs...)),
		PolicyRefs:                  normalizeDisplaySafeRefs(append(cloneDisplaySafeRefs(request.Descriptor.PolicyRefs), input.PolicyRefs...)),
		EvidenceRefs:                MergeEvidenceRefs(input.EvidenceRefs, request.EvidenceRefs),
		FailureClass:                FailureNone,
		DecisionBasis: normalizeDisplaySafeRefs(append(
			[]DisplaySafeRef{
				"capability_apply:host_owned",
				"capability_apply:adapter_readiness",
			},
			input.DecisionBasis...,
		)),
		Boundaries:     hostOwnedCapabilityApplyAdapterReadinessBoundaries(request.Boundaries, input.Boundaries),
		NextHostAction: "provide_capability_apply_adapter_binding",
		RunnerEffect:   "none",
		PromptEffect:   "none",
		RuntimeEffect:  "none",
		RawOutputLoaded: input.RawOutputLoaded ||
			request.RawOutputLoaded,
	}
	if hostOwnedCapabilityApplyAdapterReadinessUnsafe(input, request) {
		result.RawOutputLoaded = true
		result = hostOwnedCapabilityApplyAdapterReadinessBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if !request.ReadyForHostCapabilityApply {
		result = hostOwnedCapabilityApplyAdapterReadinessBlock(result, firstFailureClass(request.FailureClass, FailureConfigMissing), "capability_apply_request_not_ready", "host:capability_apply_request", firstNextHostAction(request.NextHostAction, "review_capability_apply_request"))
	}
	for _, check := range []struct {
		ok      bool
		reason  string
		missing MissingInput
		next    NextHostAction
	}{
		{result.AdapterRef != "", "adapter_ref_missing", "host:capability_apply_adapter_ref", "provide_capability_apply_adapter_binding"},
		{result.AdapterVersionRef != "", "adapter_version_ref_missing", "host:capability_apply_adapter_version_ref", "provide_capability_apply_adapter_binding"},
		{result.AdapterCapabilityRef != "", "adapter_capability_ref_missing", "host:capability_apply_adapter_capability_ref", "provide_capability_apply_adapter_capability"},
		{result.AdapterContractRef != "", "adapter_contract_ref_missing", "contract:capability_apply_adapter", "provide_capability_apply_adapter_contract"},
		{result.HostConfirmationRef != "", "host_confirmation_ref_missing", "host:capability_apply_adapter_confirmation", "request_capability_apply_adapter_confirmation"},
		{result.AdapterDryRunRef != "", "adapter_dry_run_ref_missing", "host:capability_apply_adapter_dry_run_ref", "provide_capability_apply_adapter_dry_run"},
		{result.InvocationRef != "", "invocation_ref_missing", "host:capability_apply_adapter_invocation_ref", "provide_capability_apply_adapter_invocation_ref"},
		{result.ResultBindingRef != "", "result_binding_ref_missing", "host:capability_apply_result_binding", "provide_capability_apply_result_binding"},
		{result.ReadbackBindingRef != "", "readback_binding_ref_missing", "host:capability_apply_readback_binding", "provide_capability_apply_readback_binding"},
		{result.IdempotencyRef != "", "idempotency_ref_missing", "host:capability_apply_idempotency_ref", "provide_capability_apply_idempotency_ref"},
		{result.RollbackRef != "", "rollback_ref_missing", "host:capability_apply_rollback_path_ref", "provide_capability_apply_rollback_path"},
		{result.FailureBindingRef != "", "failure_binding_ref_missing", "host:capability_apply_failure_ref", "provide_capability_apply_failure_binding"},
		{result.CompensationBindingRef != "", "compensation_binding_ref_missing", "host:capability_apply_compensation_ref", "provide_capability_apply_compensation_binding"},
	} {
		if !check.ok {
			result = hostOwnedCapabilityApplyAdapterReadinessBlock(result, FailureConfigMissing, check.reason, check.missing, check.next)
		}
	}
	for _, check := range []struct {
		got     DisplaySafeRef
		want    DisplaySafeRef
		reason  string
		missing MissingInput
		next    NextHostAction
	}{
		{result.AdapterRef, request.CapabilityAdapterRef, "adapter_ref_mismatch", "host:capability_apply_adapter_ref", "review_capability_apply_adapter_binding"},
		{result.HostConfirmationRef, request.HostCapabilityConfirmationRef, "host_confirmation_ref_mismatch", "host:capability_apply_adapter_confirmation", "review_capability_apply_adapter_binding"},
		{result.AdapterDryRunRef, request.CapabilityDryRunProofRef, "adapter_dry_run_ref_mismatch", "host:capability_apply_adapter_dry_run_ref", "review_capability_apply_adapter_binding"},
		{result.ResultBindingRef, request.ExpectedCapabilityResultRef, "result_binding_ref_mismatch", "host:capability_apply_result_binding", "review_capability_apply_adapter_binding"},
		{result.ReadbackBindingRef, request.ExpectedReadbackRef, "readback_binding_ref_mismatch", "host:capability_apply_readback_binding", "review_capability_apply_adapter_binding"},
		{result.IdempotencyRef, request.IdempotencyRef, "idempotency_ref_mismatch", "host:capability_apply_idempotency_ref", "review_capability_apply_adapter_binding"},
		{result.RollbackRef, request.RollbackPathRef, "rollback_ref_mismatch", "host:capability_apply_rollback_path_ref", "review_capability_apply_adapter_binding"},
	} {
		if check.got != "" && check.want != "" && check.got != check.want {
			result = hostOwnedCapabilityApplyAdapterReadinessBlock(result, FailureVerificationFailed, check.reason, check.missing, check.next)
		}
	}
	if len(result.MissingInputs) == 0 && len(result.BlockedReasons) == 0 {
		result.Status = HostActionReady
		result.ReadyForHostCapabilityAdapterInvocation = true
		result.HostCapabilityAdapterInvocationAuthorized = true
		result.HostMayInvokeCapabilityAdapter = true
		result.NextHostAction = "host_may_invoke_capability_apply_adapter"
		result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_host_owned_capability_apply_adapter_invocation", "host_may_invoke_capability_apply_adapter")
	}
	return result.Normalize()
}

func BuildHostOwnedCapabilityApplyAdapterInvocation(input HostOwnedCapabilityApplyAdapterInvocationInput) HostOwnedCapabilityApplyAdapterInvocation {
	readiness := input.Readiness.Normalize()
	result := HostOwnedCapabilityApplyAdapterInvocation{
		ContractVersion:                ContractVersion,
		Projected:                      true,
		Status:                         HostActionBlocked,
		Readiness:                      readiness,
		Action:                         readiness.Action,
		CapabilityApplyRequestRef:      readiness.CapabilityApplyRequestRef,
		CapabilityAdapterRef:           readiness.CapabilityAdapterRef,
		ExpectedCapabilityRef:          readiness.ExpectedCapabilityRef,
		ExpectedCapabilityStateRef:     readiness.ExpectedCapabilityStateRef,
		ExpectedCapabilityResultRef:    readiness.ExpectedCapabilityResultRef,
		ExpectedReadbackRef:            readiness.ExpectedReadbackRef,
		RollbackPathRef:                readiness.RollbackPathRef,
		AdapterRef:                     readiness.AdapterRef,
		AdapterVersionRef:              readiness.AdapterVersionRef,
		InvocationRef:                  readiness.InvocationRef,
		InvocationReportRef:            normalizeOneDisplaySafeRef(input.InvocationReportRef),
		ObservedInvocationRef:          normalizeOneDisplaySafeRef(input.ObservedInvocationRef),
		HostCapabilityAdapterRunRef:    normalizeOneDisplaySafeRef(input.HostCapabilityAdapterRunRef),
		CapabilityApplyResultRef:       normalizeOneDisplaySafeRef(input.CapabilityApplyResultRef),
		CapabilityReadbackRef:          normalizeOneDisplaySafeRef(input.CapabilityReadbackRef),
		AppliedCapabilityRef:           normalizeOneDisplaySafeRef(input.AppliedCapabilityRef),
		AppliedCapabilityStateRef:      normalizeOneDisplaySafeRef(input.AppliedCapabilityStateRef),
		FailureRef:                     normalizeOneDisplaySafeRef(input.FailureRef),
		CompensationRef:                normalizeOneDisplaySafeRef(input.CompensationRef),
		CapabilityEvidenceRefs:         normalizeDisplaySafeRefs(input.CapabilityEvidenceRefs),
		FailureClass:                   FailureNone,
		HostAdapterInvocationReported:  input.HostAdapterInvocationReported,
		HostAdapterInvocationCompleted: input.HostAdapterInvocationCompleted,
		HostAdapterInvocationFailed:    input.HostAdapterInvocationFailed,
		Boundaries:                     hostOwnedCapabilityApplyAdapterInvocationBoundaries(readiness.Boundaries, input.Boundaries),
		NextHostAction:                 "provide_capability_apply_adapter_invocation_report",
		RunnerEffect:                   "none",
		PromptEffect:                   "none",
		RuntimeEffect:                  "none",
		RawOutputLoaded:                input.RawOutputLoaded || readiness.RawOutputLoaded,
	}
	if hostOwnedCapabilityApplyAdapterInvocationUnsafe(input, readiness) {
		result.RawOutputLoaded = true
		result = hostOwnedCapabilityApplyAdapterInvocationBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if !readiness.ReadyForHostCapabilityAdapterInvocation {
		result = hostOwnedCapabilityApplyAdapterInvocationBlock(result, firstFailureClass(readiness.FailureClass, FailureConfigMissing), "adapter_readiness_not_ready", "host:capability_apply_adapter_readiness", firstNextHostAction(readiness.NextHostAction, "review_capability_apply_adapter_readiness"))
	}
	if result.InvocationReportRef == "" {
		result = hostOwnedCapabilityApplyAdapterInvocationBlock(result, FailureEvidenceMissing, "invocation_report_ref_missing", "host:capability_apply_adapter_invocation_report_ref", "provide_capability_apply_adapter_invocation_report")
	}
	if result.ObservedInvocationRef == "" {
		result = hostOwnedCapabilityApplyAdapterInvocationBlock(result, FailureEvidenceMissing, "observed_invocation_ref_missing", "host:capability_apply_adapter_invocation_ref", "provide_capability_apply_adapter_invocation_report")
	} else if result.InvocationRef != "" && result.ObservedInvocationRef != result.InvocationRef {
		result = hostOwnedCapabilityApplyAdapterInvocationBlock(result, FailureVerificationFailed, "observed_invocation_ref_mismatch", "host:capability_apply_adapter_invocation_ref", "review_capability_apply_adapter_invocation_report")
	}
	if result.HostCapabilityAdapterRunRef == "" {
		result = hostOwnedCapabilityApplyAdapterInvocationBlock(result, FailureEvidenceMissing, "host_capability_adapter_run_ref_missing", "host:capability_apply_adapter_run_ref", "provide_capability_apply_adapter_invocation_report")
	}
	if !result.HostAdapterInvocationReported {
		result = hostOwnedCapabilityApplyAdapterInvocationBlock(result, FailureEvidenceMissing, "host_adapter_invocation_not_reported", "host:capability_apply_adapter_invocation_report", "provide_capability_apply_adapter_invocation_report")
	}
	if !result.HostAdapterInvocationCompleted && !result.HostAdapterInvocationFailed {
		result = hostOwnedCapabilityApplyAdapterInvocationBlock(result, FailureEvidenceMissing, "host_adapter_invocation_status_missing", "host:capability_apply_adapter_invocation_status", "provide_capability_apply_adapter_invocation_report")
	}
	if result.HostAdapterInvocationCompleted && result.HostAdapterInvocationFailed {
		result = hostOwnedCapabilityApplyAdapterInvocationBlock(result, FailureInvalidInput, "host_adapter_invocation_status_conflict", "host:capability_apply_adapter_invocation_status", "review_capability_apply_adapter_invocation_report")
	}
	if result.HostAdapterInvocationFailed {
		if result.FailureRef == "" {
			result = hostOwnedCapabilityApplyAdapterInvocationBlock(result, FailureEvidenceMissing, "failure_ref_missing", "host:capability_apply_adapter_failure_ref", "provide_capability_apply_adapter_failure_ref")
		}
		if result.CompensationRef == "" {
			result = hostOwnedCapabilityApplyAdapterInvocationBlock(result, FailureEvidenceMissing, "compensation_ref_missing", "host:capability_apply_adapter_compensation_ref", "provide_capability_apply_adapter_compensation_ref")
		}
	} else {
		for _, check := range []struct {
			got     DisplaySafeRef
			want    DisplaySafeRef
			reason  string
			missing MissingInput
			next    NextHostAction
		}{
			{result.CapabilityApplyResultRef, readiness.ExpectedCapabilityResultRef, "capability_apply_result_ref_mismatch", "host:capability_apply_result_ref", "review_capability_apply_adapter_invocation_report"},
			{result.CapabilityReadbackRef, readiness.ExpectedReadbackRef, "capability_readback_ref_mismatch", "host:capability_readback_ref", "review_capability_apply_adapter_invocation_report"},
			{result.AppliedCapabilityRef, readiness.ExpectedCapabilityRef, "applied_capability_ref_mismatch", "host:applied_capability_ref", "review_capability_apply_adapter_invocation_report"},
			{result.AppliedCapabilityStateRef, readiness.ExpectedCapabilityStateRef, "applied_capability_state_ref_mismatch", "host:applied_capability_state_ref", "review_capability_apply_adapter_invocation_report"},
		} {
			if check.got == "" {
				result = hostOwnedCapabilityApplyAdapterInvocationBlock(result, FailureEvidenceMissing, strings.TrimSuffix(check.reason, "_mismatch")+"_missing", check.missing, "provide_capability_apply_adapter_invocation_report")
			} else if check.want != "" && check.got != check.want {
				result = hostOwnedCapabilityApplyAdapterInvocationBlock(result, FailureVerificationFailed, check.reason, check.missing, check.next)
			}
		}
	}
	if len(result.MissingInputs) == 0 && len(result.BlockedReasons) == 0 {
		result.Status = HostActionRecorded
		if result.HostAdapterInvocationFailed {
			result.ReadyForFailureReview = true
			result.FailureClass = firstFailureClass(result.FailureClass, FailureVerificationFailed)
			result.NextHostAction = "review_capability_apply_adapter_failure"
			result.Boundaries = AppendBoundaries(result.Boundaries, "host_owned_capability_apply_adapter_failure_recorded")
		} else {
			result.ReadyForCapabilityApplyResult = true
			result.NextHostAction = "build_capability_apply_result"
			result.Boundaries = AppendBoundaries(result.Boundaries, "host_owned_capability_apply_adapter_invocation_recorded")
		}
	}
	return result.Normalize()
}

func CloneHostOwnedCapabilityApplyAdapterReadiness(in HostOwnedCapabilityApplyAdapterReadiness) HostOwnedCapabilityApplyAdapterReadiness {
	out := in
	out.Request = in.Request.Clone()
	out.ApprovalRefs = cloneDisplaySafeRefs(in.ApprovalRefs)
	out.PolicyRefs = cloneDisplaySafeRefs(in.PolicyRefs)
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.DecisionBasis = cloneDisplaySafeRefs(in.DecisionBasis)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (r HostOwnedCapabilityApplyAdapterReadiness) Clone() HostOwnedCapabilityApplyAdapterReadiness {
	return CloneHostOwnedCapabilityApplyAdapterReadiness(r)
}

func (r HostOwnedCapabilityApplyAdapterReadiness) Normalize() HostOwnedCapabilityApplyAdapterReadiness {
	out := CloneHostOwnedCapabilityApplyAdapterReadiness(r)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.Request = out.Request.Normalize()
	out.Action = NormalizeCapabilityApplyAction(string(out.Action))
	out.CapabilityApplyRequestRef = normalizeOneDisplaySafeRef(out.CapabilityApplyRequestRef)
	out.CapabilityAdapterRef = normalizeOneDisplaySafeRef(out.CapabilityAdapterRef)
	out.ExpectedCapabilityRef = normalizeOneDisplaySafeRef(out.ExpectedCapabilityRef)
	out.ExpectedCapabilityStateRef = normalizeOneDisplaySafeRef(out.ExpectedCapabilityStateRef)
	out.ExpectedCapabilityResultRef = normalizeOneDisplaySafeRef(out.ExpectedCapabilityResultRef)
	out.ExpectedReadbackRef = normalizeOneDisplaySafeRef(out.ExpectedReadbackRef)
	out.RollbackPathRef = normalizeOneDisplaySafeRef(out.RollbackPathRef)
	out.AdapterRef = normalizeOneDisplaySafeRef(out.AdapterRef)
	out.AdapterVersionRef = normalizeOneDisplaySafeRef(out.AdapterVersionRef)
	out.AdapterCapabilityRef = normalizeOneDisplaySafeRef(out.AdapterCapabilityRef)
	out.AdapterContractRef = normalizeOneDisplaySafeRef(out.AdapterContractRef)
	out.HostConfirmationRef = normalizeOneDisplaySafeRef(out.HostConfirmationRef)
	out.AdapterDryRunRef = normalizeOneDisplaySafeRef(out.AdapterDryRunRef)
	out.InvocationRef = normalizeOneDisplaySafeRef(out.InvocationRef)
	out.ResultBindingRef = normalizeOneDisplaySafeRef(out.ResultBindingRef)
	out.ReadbackBindingRef = normalizeOneDisplaySafeRef(out.ReadbackBindingRef)
	out.IdempotencyRef = normalizeOneDisplaySafeRef(out.IdempotencyRef)
	out.RollbackRef = normalizeOneDisplaySafeRef(out.RollbackRef)
	out.FailureBindingRef = normalizeOneDisplaySafeRef(out.FailureBindingRef)
	out.CompensationBindingRef = normalizeOneDisplaySafeRef(out.CompensationBindingRef)
	out.ApprovalRefs = normalizeDisplaySafeRefs(out.ApprovalRefs)
	out.PolicyRefs = normalizeDisplaySafeRefs(out.PolicyRefs)
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.BlockedReasons = normalizeControlTokenList(out.BlockedReasons)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.DecisionBasis = normalizeDisplaySafeRefs(out.DecisionBasis)
	out.Boundaries = hostOwnedCapabilityApplyAdapterReadinessBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	out.RuntimeEffect = normalizeControlToken(out.RuntimeEffect)
	out = hostOwnedCapabilityApplyAdapterNormalizeReadinessEffects(out)
	out.ReadyForHostCapabilityAdapterInvocation = out.Status == HostActionReady && len(out.MissingInputs) == 0 && len(out.BlockedReasons) == 0 && !out.RawOutputLoaded
	out.HostCapabilityAdapterInvocationAuthorized = out.ReadyForHostCapabilityAdapterInvocation
	out.HostMayInvokeCapabilityAdapter = out.ReadyForHostCapabilityAdapterInvocation
	if out.RawOutputLoaded {
		out.Status = HostActionReviewRequired
		out.ReadyForHostCapabilityAdapterInvocation = false
		out.HostCapabilityAdapterInvocationAuthorized = false
		out.HostMayInvokeCapabilityAdapter = false
		if out.FailureClass == FailureNone {
			out.FailureClass = FailureEvidenceWeak
		}
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		out.NextHostAction = "provide_display_safe_refs"
	}
	return out
}

func CloneHostOwnedCapabilityApplyAdapterInvocation(in HostOwnedCapabilityApplyAdapterInvocation) HostOwnedCapabilityApplyAdapterInvocation {
	out := in
	out.Readiness = in.Readiness.Clone()
	out.CapabilityEvidenceRefs = cloneDisplaySafeRefs(in.CapabilityEvidenceRefs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (r HostOwnedCapabilityApplyAdapterInvocation) Clone() HostOwnedCapabilityApplyAdapterInvocation {
	return CloneHostOwnedCapabilityApplyAdapterInvocation(r)
}

func (r HostOwnedCapabilityApplyAdapterInvocation) Normalize() HostOwnedCapabilityApplyAdapterInvocation {
	out := CloneHostOwnedCapabilityApplyAdapterInvocation(r)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.Readiness = out.Readiness.Normalize()
	out.Action = NormalizeCapabilityApplyAction(string(out.Action))
	out.CapabilityApplyRequestRef = normalizeOneDisplaySafeRef(out.CapabilityApplyRequestRef)
	out.CapabilityAdapterRef = normalizeOneDisplaySafeRef(out.CapabilityAdapterRef)
	out.ExpectedCapabilityRef = normalizeOneDisplaySafeRef(out.ExpectedCapabilityRef)
	out.ExpectedCapabilityStateRef = normalizeOneDisplaySafeRef(out.ExpectedCapabilityStateRef)
	out.ExpectedCapabilityResultRef = normalizeOneDisplaySafeRef(out.ExpectedCapabilityResultRef)
	out.ExpectedReadbackRef = normalizeOneDisplaySafeRef(out.ExpectedReadbackRef)
	out.RollbackPathRef = normalizeOneDisplaySafeRef(out.RollbackPathRef)
	out.AdapterRef = normalizeOneDisplaySafeRef(out.AdapterRef)
	out.AdapterVersionRef = normalizeOneDisplaySafeRef(out.AdapterVersionRef)
	out.InvocationRef = normalizeOneDisplaySafeRef(out.InvocationRef)
	out.InvocationReportRef = normalizeOneDisplaySafeRef(out.InvocationReportRef)
	out.ObservedInvocationRef = normalizeOneDisplaySafeRef(out.ObservedInvocationRef)
	out.HostCapabilityAdapterRunRef = normalizeOneDisplaySafeRef(out.HostCapabilityAdapterRunRef)
	out.CapabilityApplyResultRef = normalizeOneDisplaySafeRef(out.CapabilityApplyResultRef)
	out.CapabilityReadbackRef = normalizeOneDisplaySafeRef(out.CapabilityReadbackRef)
	out.AppliedCapabilityRef = normalizeOneDisplaySafeRef(out.AppliedCapabilityRef)
	out.AppliedCapabilityStateRef = normalizeOneDisplaySafeRef(out.AppliedCapabilityStateRef)
	out.FailureRef = normalizeOneDisplaySafeRef(out.FailureRef)
	out.CompensationRef = normalizeOneDisplaySafeRef(out.CompensationRef)
	out.CapabilityEvidenceRefs = normalizeDisplaySafeRefs(out.CapabilityEvidenceRefs)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.BlockedReasons = normalizeControlTokenList(out.BlockedReasons)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.Boundaries = hostOwnedCapabilityApplyAdapterInvocationBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	out.RuntimeEffect = normalizeControlToken(out.RuntimeEffect)
	out = hostOwnedCapabilityApplyAdapterNormalizeInvocationEffects(out)
	out.ReadyForCapabilityApplyResult = out.Status == HostActionRecorded && out.HostAdapterInvocationCompleted && !out.HostAdapterInvocationFailed && len(out.MissingInputs) == 0 && len(out.BlockedReasons) == 0 && !out.RawOutputLoaded
	out.ReadyForFailureReview = out.Status == HostActionRecorded && out.HostAdapterInvocationFailed && len(out.MissingInputs) == 0 && len(out.BlockedReasons) == 0 && !out.RawOutputLoaded
	if out.RawOutputLoaded {
		out.Status = HostActionReviewRequired
		out.ReadyForCapabilityApplyResult = false
		out.ReadyForFailureReview = false
		if out.FailureClass == FailureNone {
			out.FailureClass = FailureEvidenceWeak
		}
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		out.NextHostAction = "provide_display_safe_refs"
	}
	return out
}

func hostOwnedCapabilityApplyAdapterReadinessBlock(result HostOwnedCapabilityApplyAdapterReadiness, failure FailureClass, reason string, missing MissingInput, next NextHostAction) HostOwnedCapabilityApplyAdapterReadiness {
	result.Status = HostActionBlocked
	result.ReadyForHostCapabilityAdapterInvocation = false
	result.HostCapabilityAdapterInvocationAuthorized = false
	result.HostMayInvokeCapabilityAdapter = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	return result
}

func hostOwnedCapabilityApplyAdapterInvocationBlock(result HostOwnedCapabilityApplyAdapterInvocation, failure FailureClass, reason string, missing MissingInput, next NextHostAction) HostOwnedCapabilityApplyAdapterInvocation {
	result.Status = HostActionBlocked
	result.ReadyForCapabilityApplyResult = false
	result.ReadyForFailureReview = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	return result
}

func hostOwnedCapabilityApplyAdapterReadinessUnsafe(input HostOwnedCapabilityApplyAdapterReadinessInput, request HostOwnedCapabilityApplyRequest) bool {
	return input.RawOutputLoaded ||
		request.RawOutputLoaded ||
		displaySafeRefRejected(input.AdapterRef) ||
		displaySafeRefRejected(input.AdapterVersionRef) ||
		displaySafeRefRejected(input.AdapterCapabilityRef) ||
		displaySafeRefRejected(input.AdapterContractRef) ||
		displaySafeRefRejected(input.HostConfirmationRef) ||
		displaySafeRefRejected(input.AdapterDryRunRef) ||
		displaySafeRefRejected(input.InvocationRef) ||
		displaySafeRefRejected(input.ResultBindingRef) ||
		displaySafeRefRejected(input.ReadbackBindingRef) ||
		displaySafeRefRejected(input.IdempotencyRef) ||
		displaySafeRefRejected(input.RollbackRef) ||
		displaySafeRefRejected(input.FailureBindingRef) ||
		displaySafeRefRejected(input.CompensationBindingRef) ||
		displaySafeRefSliceRejected(input.ApprovalRefs) ||
		displaySafeRefSliceRejected(input.PolicyRefs) ||
		displaySafeRefSliceRejected(input.DecisionBasis) ||
		evidenceRefRejected(input.EvidenceRefs) ||
		hostOwnedCapabilityApplyRequestOutputUnsafe(request)
}

func hostOwnedCapabilityApplyAdapterInvocationUnsafe(input HostOwnedCapabilityApplyAdapterInvocationInput, readiness HostOwnedCapabilityApplyAdapterReadiness) bool {
	return input.RawOutputLoaded ||
		readiness.RawOutputLoaded ||
		displaySafeRefRejected(input.InvocationReportRef) ||
		displaySafeRefRejected(input.ObservedInvocationRef) ||
		displaySafeRefRejected(input.HostCapabilityAdapterRunRef) ||
		displaySafeRefRejected(input.CapabilityApplyResultRef) ||
		displaySafeRefRejected(input.CapabilityReadbackRef) ||
		displaySafeRefRejected(input.AppliedCapabilityRef) ||
		displaySafeRefRejected(input.AppliedCapabilityStateRef) ||
		displaySafeRefRejected(input.FailureRef) ||
		displaySafeRefRejected(input.CompensationRef) ||
		displaySafeRefSliceRejected(input.CapabilityEvidenceRefs) ||
		hostOwnedCapabilityApplyAdapterReadinessOutputUnsafe(readiness)
}

func hostOwnedCapabilityApplyAdapterReadinessBoundaries(groups ...[]Boundary) []Boundary {
	all := append([][]Boundary{{
		"host_owned_capability_apply_adapter_gate",
		"host_owned_capability_apply_adapter_readiness",
		"capability_apply_adapter_invocation_gate",
		"explicit_host_confirmation_required",
		"capability_apply_adapter_dry_run_required",
		"host_owned_capability_apply_request_required",
		"host_adapter_may_apply_capability_after_approval",
		"display_safe_refs_only",
		"no_capability_apply_by_core",
		"no_install_apply_by_core",
		"no_package_manager_execution_by_core",
		"no_skill_write_by_core",
		"no_runtime_reload_by_core",
		"no_core_execution",
		"no_runner_dispatch",
		"no_runtime_adapter_execution",
		"no_workflow_dispatch",
		"no_worker_dispatch",
		"no_store_mutation_by_core",
		"no_compensation_execution",
		"projection_only",
	}}, groups...)
	return MergeBoundaries(all...)
}

func hostOwnedCapabilityApplyAdapterInvocationBoundaries(groups ...[]Boundary) []Boundary {
	all := append([][]Boundary{{
		"host_owned_capability_apply_adapter_gate",
		"host_owned_capability_apply_adapter_invocation_report",
		"host_capability_adapter_invocation_report_only",
		"host_adapter_capability_mutation_reported_only",
		"capability_apply_result_requires_readback",
		"capability_install_proposal_not_apply",
		"display_safe_refs_only",
		"no_capability_apply_by_core",
		"no_install_apply_by_core",
		"no_package_manager_execution_by_core",
		"no_skill_write_by_core",
		"no_runtime_reload_by_core",
		"no_core_execution",
		"no_runner_dispatch",
		"no_runtime_adapter_execution",
		"no_workflow_dispatch",
		"no_worker_dispatch",
		"no_store_mutation_by_core",
		"no_compensation_execution",
		"projection_only",
	}}, groups...)
	return MergeBoundaries(all...)
}

func hostOwnedCapabilityApplyAdapterNormalizeReadinessEffects(value HostOwnedCapabilityApplyAdapterReadiness) HostOwnedCapabilityApplyAdapterReadiness {
	if value.RunnerEffect == "" {
		value.RunnerEffect = "none"
	}
	if value.PromptEffect == "" {
		value.PromptEffect = "none"
	}
	if value.RuntimeEffect == "" {
		value.RuntimeEffect = "none"
	}
	value.CoreInvocationExecuted = false
	value.InstallerExecutedByCore = false
	value.InstallExecutedByCore = false
	value.EnableExecutedByCore = false
	value.PackageManagerExecutedByCore = false
	value.SkillWriteByCore = false
	value.RuntimeReloadByCore = false
	value.RunnerDispatched = false
	value.RuntimeAdapterExecuted = false
	value.WorkflowDispatched = false
	value.WorkerDispatched = false
	value.StoreMutationExecuted = false
	value.CompensationExecuted = false
	return value
}

func hostOwnedCapabilityApplyAdapterNormalizeInvocationEffects(value HostOwnedCapabilityApplyAdapterInvocation) HostOwnedCapabilityApplyAdapterInvocation {
	if value.RunnerEffect == "" {
		value.RunnerEffect = "none"
	}
	if value.PromptEffect == "" {
		value.PromptEffect = "none"
	}
	if value.RuntimeEffect == "" {
		value.RuntimeEffect = "none"
	}
	value.CoreInvocationExecuted = false
	value.InstallerExecutedByCore = false
	value.InstallExecutedByCore = false
	value.EnableExecutedByCore = false
	value.PackageManagerExecutedByCore = false
	value.SkillWriteByCore = false
	value.RuntimeReloadByCore = false
	value.RunnerDispatched = false
	value.RuntimeAdapterExecuted = false
	value.WorkflowDispatched = false
	value.WorkerDispatched = false
	value.StoreMutationExecuted = false
	value.CompensationExecuted = false
	return value
}

func hostOwnedCapabilityApplyAdapterReadinessOutputUnsafe(input HostOwnedCapabilityApplyAdapterReadiness) bool {
	return displaySafeRefRejected(input.CapabilityApplyRequestRef) ||
		displaySafeRefRejected(input.CapabilityAdapterRef) ||
		displaySafeRefRejected(input.ExpectedCapabilityRef) ||
		displaySafeRefRejected(input.ExpectedCapabilityStateRef) ||
		displaySafeRefRejected(input.ExpectedCapabilityResultRef) ||
		displaySafeRefRejected(input.ExpectedReadbackRef) ||
		displaySafeRefRejected(input.RollbackPathRef) ||
		displaySafeRefRejected(input.AdapterRef) ||
		displaySafeRefRejected(input.AdapterVersionRef) ||
		displaySafeRefRejected(input.AdapterCapabilityRef) ||
		displaySafeRefRejected(input.AdapterContractRef) ||
		displaySafeRefRejected(input.HostConfirmationRef) ||
		displaySafeRefRejected(input.AdapterDryRunRef) ||
		displaySafeRefRejected(input.InvocationRef) ||
		displaySafeRefRejected(input.ResultBindingRef) ||
		displaySafeRefRejected(input.ReadbackBindingRef) ||
		displaySafeRefRejected(input.IdempotencyRef) ||
		displaySafeRefRejected(input.RollbackRef) ||
		displaySafeRefRejected(input.FailureBindingRef) ||
		displaySafeRefRejected(input.CompensationBindingRef) ||
		displaySafeRefSliceRejected(input.ApprovalRefs) ||
		displaySafeRefSliceRejected(input.PolicyRefs) ||
		displaySafeRefSliceRejected(input.DecisionBasis) ||
		evidenceRefRejected(input.EvidenceRefs) ||
		hostOwnedCapabilityApplyRequestOutputUnsafe(input.Request)
}
