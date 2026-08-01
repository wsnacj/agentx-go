package controlcontract

import "strings"

type HostOwnedSchedulerApplyAdapterReadinessInput struct {
	Request              HostOwnedSchedulerApplyRequest `json:"request,omitempty"`
	AdapterRef           DisplaySafeRef                 `json:"adapter_ref,omitempty"`
	AdapterVersionRef    DisplaySafeRef                 `json:"adapter_version_ref,omitempty"`
	AdapterCapabilityRef DisplaySafeRef                 `json:"adapter_capability_ref,omitempty"`
	AdapterContractRef   DisplaySafeRef                 `json:"adapter_contract_ref,omitempty"`
	HostConfirmationRef  DisplaySafeRef                 `json:"host_confirmation_ref,omitempty"`
	AdapterDryRunRef     DisplaySafeRef                 `json:"adapter_dry_run_ref,omitempty"`
	InvocationRef        DisplaySafeRef                 `json:"invocation_ref,omitempty"`
	ResultBindingRef     DisplaySafeRef                 `json:"result_binding_ref,omitempty"`
	ReadbackBindingRef   DisplaySafeRef                 `json:"readback_binding_ref,omitempty"`
	IdempotencyRef       DisplaySafeRef                 `json:"idempotency_ref,omitempty"`
	CancellationRef      DisplaySafeRef                 `json:"cancellation_ref,omitempty"`
	DeleteRef            DisplaySafeRef                 `json:"delete_ref,omitempty"`
	DisableRef           DisplaySafeRef                 `json:"disable_ref,omitempty"`
	ApprovalRefs         []DisplaySafeRef               `json:"approval_refs,omitempty"`
	PolicyRefs           []DisplaySafeRef               `json:"policy_refs,omitempty"`
	EvidenceRefs         []EvidenceRef                  `json:"evidence_refs,omitempty"`
	DecisionBasis        []DisplaySafeRef               `json:"decision_basis,omitempty"`
	Boundaries           []Boundary                     `json:"boundaries,omitempty"`
	RawOutputLoaded      bool                           `json:"raw_output_loaded"`
}

type HostOwnedSchedulerApplyAdapterReadiness struct {
	ContractVersion                          string                         `json:"contract_version,omitempty"`
	Projected                                bool                           `json:"projected"`
	Status                                   HostActionStatus               `json:"status,omitempty"`
	ReadyForHostSchedulerAdapterInvocation   bool                           `json:"ready_for_host_scheduler_adapter_invocation"`
	HostSchedulerAdapterInvocationAuthorized bool                           `json:"host_scheduler_adapter_invocation_authorized"`
	HostMayInvokeSchedulerAdapter            bool                           `json:"host_may_invoke_scheduler_adapter"`
	Request                                  HostOwnedSchedulerApplyRequest `json:"request,omitempty"`
	Action                                   SchedulerApplyAction           `json:"action,omitempty"`
	SchedulerApplyRequestRef                 DisplaySafeRef                 `json:"scheduler_apply_request_ref,omitempty"`
	SchedulerAdapterRef                      DisplaySafeRef                 `json:"scheduler_adapter_ref,omitempty"`
	ExpectedScheduleRef                      DisplaySafeRef                 `json:"expected_schedule_ref,omitempty"`
	ExpectedLifecycleStateRef                DisplaySafeRef                 `json:"expected_lifecycle_state_ref,omitempty"`
	ExpectedSchedulerResultRef               DisplaySafeRef                 `json:"expected_scheduler_result_ref,omitempty"`
	ExpectedReadbackRef                      DisplaySafeRef                 `json:"expected_readback_ref,omitempty"`
	AdapterRef                               DisplaySafeRef                 `json:"adapter_ref,omitempty"`
	AdapterVersionRef                        DisplaySafeRef                 `json:"adapter_version_ref,omitempty"`
	AdapterCapabilityRef                     DisplaySafeRef                 `json:"adapter_capability_ref,omitempty"`
	AdapterContractRef                       DisplaySafeRef                 `json:"adapter_contract_ref,omitempty"`
	HostConfirmationRef                      DisplaySafeRef                 `json:"host_confirmation_ref,omitempty"`
	AdapterDryRunRef                         DisplaySafeRef                 `json:"adapter_dry_run_ref,omitempty"`
	InvocationRef                            DisplaySafeRef                 `json:"invocation_ref,omitempty"`
	ResultBindingRef                         DisplaySafeRef                 `json:"result_binding_ref,omitempty"`
	ReadbackBindingRef                       DisplaySafeRef                 `json:"readback_binding_ref,omitempty"`
	IdempotencyRef                           DisplaySafeRef                 `json:"idempotency_ref,omitempty"`
	CancellationRef                          DisplaySafeRef                 `json:"cancellation_ref,omitempty"`
	DeleteRef                                DisplaySafeRef                 `json:"delete_ref,omitempty"`
	DisableRef                               DisplaySafeRef                 `json:"disable_ref,omitempty"`
	ApprovalRefs                             []DisplaySafeRef               `json:"approval_refs,omitempty"`
	PolicyRefs                               []DisplaySafeRef               `json:"policy_refs,omitempty"`
	EvidenceRefs                             []EvidenceRef                  `json:"evidence_refs,omitempty"`
	FailureClass                             FailureClass                   `json:"failure_class,omitempty"`
	BlockedReasons                           []string                       `json:"blocked_reasons,omitempty"`
	MissingInputs                            []MissingInput                 `json:"missing_inputs,omitempty"`
	DecisionBasis                            []DisplaySafeRef               `json:"decision_basis,omitempty"`
	Boundaries                               []Boundary                     `json:"boundaries,omitempty"`
	NextHostAction                           NextHostAction                 `json:"next_host_action,omitempty"`
	RunnerEffect                             string                         `json:"runner_effect,omitempty"`
	PromptEffect                             string                         `json:"prompt_effect,omitempty"`
	RuntimeEffect                            string                         `json:"runtime_effect,omitempty"`
	CoreInvocationExecuted                   bool                           `json:"core_invocation_executed"`
	SchedulerMutationByCore                  bool                           `json:"scheduler_mutation_by_core"`
	CoreScheduleCreated                      bool                           `json:"core_schedule_created"`
	AutomationCreatedByCore                  bool                           `json:"automation_created_by_core"`
	RunnerDispatched                         bool                           `json:"runner_dispatched"`
	RuntimeAdapterExecuted                   bool                           `json:"runtime_adapter_executed"`
	InstallerExecuted                        bool                           `json:"installer_executed"`
	WorkflowDispatched                       bool                           `json:"workflow_dispatched"`
	WorkerDispatched                         bool                           `json:"worker_dispatched"`
	StoreMutationExecuted                    bool                           `json:"store_mutation_executed"`
	CompensationExecuted                     bool                           `json:"compensation_executed"`
	RawOutputLoaded                          bool                           `json:"raw_output_loaded"`
}

type HostOwnedSchedulerApplyAdapterInvocationInput struct {
	Readiness                      HostOwnedSchedulerApplyAdapterReadiness `json:"readiness,omitempty"`
	InvocationReportRef            DisplaySafeRef                          `json:"invocation_report_ref,omitempty"`
	ObservedInvocationRef          DisplaySafeRef                          `json:"observed_invocation_ref,omitempty"`
	HostSchedulerAdapterRunRef     DisplaySafeRef                          `json:"host_scheduler_adapter_run_ref,omitempty"`
	SchedulerApplyResultRef        DisplaySafeRef                          `json:"scheduler_apply_result_ref,omitempty"`
	SchedulerReadbackRef           DisplaySafeRef                          `json:"scheduler_readback_ref,omitempty"`
	AppliedScheduleRef             DisplaySafeRef                          `json:"applied_schedule_ref,omitempty"`
	AppliedLifecycleStateRef       DisplaySafeRef                          `json:"applied_lifecycle_state_ref,omitempty"`
	FailureRef                     DisplaySafeRef                          `json:"failure_ref,omitempty"`
	CompensationRef                DisplaySafeRef                          `json:"compensation_ref,omitempty"`
	HostAdapterInvocationReported  bool                                    `json:"host_adapter_invocation_reported"`
	HostAdapterInvocationCompleted bool                                    `json:"host_adapter_invocation_completed"`
	HostAdapterInvocationFailed    bool                                    `json:"host_adapter_invocation_failed"`
	SchedulerEvidenceRefs          []DisplaySafeRef                        `json:"scheduler_evidence_refs,omitempty"`
	Boundaries                     []Boundary                              `json:"boundaries,omitempty"`
	RawOutputLoaded                bool                                    `json:"raw_output_loaded"`
}

type HostOwnedSchedulerApplyAdapterInvocation struct {
	ContractVersion                string                                  `json:"contract_version,omitempty"`
	Projected                      bool                                    `json:"projected"`
	Status                         HostActionStatus                        `json:"status,omitempty"`
	ReadyForSchedulerApplyResult   bool                                    `json:"ready_for_scheduler_apply_result"`
	ReadyForFailureReview          bool                                    `json:"ready_for_failure_review"`
	HostAdapterInvocationReported  bool                                    `json:"host_adapter_invocation_reported"`
	HostAdapterInvocationCompleted bool                                    `json:"host_adapter_invocation_completed"`
	HostAdapterInvocationFailed    bool                                    `json:"host_adapter_invocation_failed"`
	Readiness                      HostOwnedSchedulerApplyAdapterReadiness `json:"readiness,omitempty"`
	Action                         SchedulerApplyAction                    `json:"action,omitempty"`
	SchedulerApplyRequestRef       DisplaySafeRef                          `json:"scheduler_apply_request_ref,omitempty"`
	SchedulerAdapterRef            DisplaySafeRef                          `json:"scheduler_adapter_ref,omitempty"`
	ExpectedScheduleRef            DisplaySafeRef                          `json:"expected_schedule_ref,omitempty"`
	ExpectedLifecycleStateRef      DisplaySafeRef                          `json:"expected_lifecycle_state_ref,omitempty"`
	ExpectedSchedulerResultRef     DisplaySafeRef                          `json:"expected_scheduler_result_ref,omitempty"`
	ExpectedReadbackRef            DisplaySafeRef                          `json:"expected_readback_ref,omitempty"`
	AdapterRef                     DisplaySafeRef                          `json:"adapter_ref,omitempty"`
	AdapterVersionRef              DisplaySafeRef                          `json:"adapter_version_ref,omitempty"`
	InvocationRef                  DisplaySafeRef                          `json:"invocation_ref,omitempty"`
	InvocationReportRef            DisplaySafeRef                          `json:"invocation_report_ref,omitempty"`
	ObservedInvocationRef          DisplaySafeRef                          `json:"observed_invocation_ref,omitempty"`
	HostSchedulerAdapterRunRef     DisplaySafeRef                          `json:"host_scheduler_adapter_run_ref,omitempty"`
	SchedulerApplyResultRef        DisplaySafeRef                          `json:"scheduler_apply_result_ref,omitempty"`
	SchedulerReadbackRef           DisplaySafeRef                          `json:"scheduler_readback_ref,omitempty"`
	AppliedScheduleRef             DisplaySafeRef                          `json:"applied_schedule_ref,omitempty"`
	AppliedLifecycleStateRef       DisplaySafeRef                          `json:"applied_lifecycle_state_ref,omitempty"`
	FailureRef                     DisplaySafeRef                          `json:"failure_ref,omitempty"`
	CompensationRef                DisplaySafeRef                          `json:"compensation_ref,omitempty"`
	SchedulerEvidenceRefs          []DisplaySafeRef                        `json:"scheduler_evidence_refs,omitempty"`
	FailureClass                   FailureClass                            `json:"failure_class,omitempty"`
	BlockedReasons                 []string                                `json:"blocked_reasons,omitempty"`
	MissingInputs                  []MissingInput                          `json:"missing_inputs,omitempty"`
	Boundaries                     []Boundary                              `json:"boundaries,omitempty"`
	NextHostAction                 NextHostAction                          `json:"next_host_action,omitempty"`
	RunnerEffect                   string                                  `json:"runner_effect,omitempty"`
	PromptEffect                   string                                  `json:"prompt_effect,omitempty"`
	RuntimeEffect                  string                                  `json:"runtime_effect,omitempty"`
	CoreInvocationExecuted         bool                                    `json:"core_invocation_executed"`
	SchedulerMutationByCore        bool                                    `json:"scheduler_mutation_by_core"`
	CoreScheduleCreated            bool                                    `json:"core_schedule_created"`
	AutomationCreatedByCore        bool                                    `json:"automation_created_by_core"`
	RunnerDispatched               bool                                    `json:"runner_dispatched"`
	RuntimeAdapterExecuted         bool                                    `json:"runtime_adapter_executed"`
	InstallerExecuted              bool                                    `json:"installer_executed"`
	WorkflowDispatched             bool                                    `json:"workflow_dispatched"`
	WorkerDispatched               bool                                    `json:"worker_dispatched"`
	StoreMutationExecuted          bool                                    `json:"store_mutation_executed"`
	CompensationExecuted           bool                                    `json:"compensation_executed"`
	RawOutputLoaded                bool                                    `json:"raw_output_loaded"`
}

func BuildHostOwnedSchedulerApplyAdapterReadiness(input HostOwnedSchedulerApplyAdapterReadinessInput) HostOwnedSchedulerApplyAdapterReadiness {
	request := input.Request.Normalize()
	result := HostOwnedSchedulerApplyAdapterReadiness{
		ContractVersion:            ContractVersion,
		Projected:                  true,
		Status:                     HostActionBlocked,
		Request:                    request,
		Action:                     request.Action,
		SchedulerApplyRequestRef:   request.SchedulerApplyRequestRef,
		SchedulerAdapterRef:        request.SchedulerAdapterRef,
		ExpectedScheduleRef:        request.ExpectedScheduleRef,
		ExpectedLifecycleStateRef:  request.ExpectedLifecycleStateRef,
		ExpectedSchedulerResultRef: request.ExpectedSchedulerResultRef,
		ExpectedReadbackRef:        request.ExpectedReadbackRef,
		AdapterRef:                 normalizeOneDisplaySafeRef(input.AdapterRef),
		AdapterVersionRef:          normalizeOneDisplaySafeRef(input.AdapterVersionRef),
		AdapterCapabilityRef:       normalizeOneDisplaySafeRef(input.AdapterCapabilityRef),
		AdapterContractRef:         normalizeOneDisplaySafeRef(input.AdapterContractRef),
		HostConfirmationRef:        normalizeOneDisplaySafeRef(input.HostConfirmationRef),
		AdapterDryRunRef:           normalizeOneDisplaySafeRef(input.AdapterDryRunRef),
		InvocationRef:              normalizeOneDisplaySafeRef(input.InvocationRef),
		ResultBindingRef:           normalizeOneDisplaySafeRef(input.ResultBindingRef),
		ReadbackBindingRef:         normalizeOneDisplaySafeRef(input.ReadbackBindingRef),
		IdempotencyRef:             normalizeOneDisplaySafeRef(input.IdempotencyRef),
		CancellationRef:            normalizeOneDisplaySafeRef(input.CancellationRef),
		DeleteRef:                  normalizeOneDisplaySafeRef(input.DeleteRef),
		DisableRef:                 normalizeOneDisplaySafeRef(input.DisableRef),
		ApprovalRefs:               normalizeDisplaySafeRefs(append(cloneDisplaySafeRefs(request.ApprovalRefs), input.ApprovalRefs...)),
		PolicyRefs:                 normalizeDisplaySafeRefs(append(cloneDisplaySafeRefs(request.Descriptor.PolicyRefs), input.PolicyRefs...)),
		EvidenceRefs:               MergeEvidenceRefs(input.EvidenceRefs, request.EvidenceRefs),
		FailureClass:               FailureNone,
		DecisionBasis: normalizeDisplaySafeRefs(append(
			[]DisplaySafeRef{
				"scheduler_apply:host_owned",
				"scheduler_apply:adapter_readiness",
			},
			input.DecisionBasis...,
		)),
		Boundaries:     hostOwnedSchedulerApplyAdapterReadinessBoundaries(request.Boundaries, input.Boundaries),
		NextHostAction: "provide_scheduler_apply_adapter_binding",
		RunnerEffect:   "none",
		PromptEffect:   "none",
		RuntimeEffect:  "none",
		RawOutputLoaded: input.RawOutputLoaded ||
			request.RawOutputLoaded,
	}
	if hostOwnedSchedulerApplyAdapterReadinessUnsafe(input, request) {
		result.RawOutputLoaded = true
		result = hostOwnedSchedulerApplyAdapterReadinessBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if !request.ReadyForHostSchedulerApply {
		result = hostOwnedSchedulerApplyAdapterReadinessBlock(result, firstFailureClass(request.FailureClass, FailureConfigMissing), "scheduler_apply_request_not_ready", "host:scheduler_apply_request", firstNextHostAction(request.NextHostAction, "review_scheduler_apply_request"))
	}
	for _, check := range []struct {
		ok      bool
		reason  string
		missing MissingInput
		next    NextHostAction
	}{
		{result.AdapterRef != "", "adapter_ref_missing", "host:scheduler_apply_adapter_ref", "provide_scheduler_apply_adapter_binding"},
		{result.AdapterVersionRef != "", "adapter_version_ref_missing", "host:scheduler_apply_adapter_version_ref", "provide_scheduler_apply_adapter_binding"},
		{result.AdapterCapabilityRef != "", "adapter_capability_ref_missing", "host:scheduler_apply_adapter_capability_ref", "provide_scheduler_apply_adapter_capability"},
		{result.AdapterContractRef != "", "adapter_contract_ref_missing", "contract:scheduler_apply_adapter", "provide_scheduler_apply_adapter_contract"},
		{result.HostConfirmationRef != "", "host_confirmation_ref_missing", "host:scheduler_apply_adapter_confirmation", "request_scheduler_apply_adapter_confirmation"},
		{result.AdapterDryRunRef != "", "adapter_dry_run_ref_missing", "host:scheduler_apply_adapter_dry_run_ref", "provide_scheduler_apply_adapter_dry_run"},
		{result.InvocationRef != "", "invocation_ref_missing", "host:scheduler_apply_adapter_invocation_ref", "provide_scheduler_apply_adapter_invocation_ref"},
		{result.ResultBindingRef != "", "result_binding_ref_missing", "host:scheduler_apply_result_binding", "provide_scheduler_apply_result_binding"},
		{result.ReadbackBindingRef != "", "readback_binding_ref_missing", "host:scheduler_apply_readback_binding", "provide_scheduler_apply_readback_binding"},
		{result.IdempotencyRef != "", "idempotency_ref_missing", "host:scheduler_apply_idempotency_ref", "provide_scheduler_apply_idempotency_ref"},
		{result.CancellationRef != "", "cancellation_ref_missing", "host:scheduler_apply_cancel_path_ref", "provide_scheduler_apply_cancel_path"},
		{result.DeleteRef != "", "delete_ref_missing", "host:scheduler_apply_delete_path_ref", "provide_scheduler_apply_delete_path"},
		{result.DisableRef != "", "disable_ref_missing", "host:scheduler_apply_disable_path_ref", "provide_scheduler_apply_disable_path"},
	} {
		if !check.ok {
			result = hostOwnedSchedulerApplyAdapterReadinessBlock(result, FailureConfigMissing, check.reason, check.missing, check.next)
		}
	}
	for _, check := range []struct {
		got     DisplaySafeRef
		want    DisplaySafeRef
		reason  string
		missing MissingInput
		next    NextHostAction
	}{
		{result.AdapterRef, request.SchedulerAdapterRef, "adapter_ref_mismatch", "host:scheduler_apply_adapter_ref", "review_scheduler_apply_adapter_binding"},
		{result.HostConfirmationRef, request.HostSchedulerConfirmationRef, "host_confirmation_ref_mismatch", "host:scheduler_apply_adapter_confirmation", "review_scheduler_apply_adapter_binding"},
		{result.AdapterDryRunRef, request.ScheduleDryRunProofRef, "adapter_dry_run_ref_mismatch", "host:scheduler_apply_adapter_dry_run_ref", "review_scheduler_apply_adapter_binding"},
		{result.ResultBindingRef, request.ExpectedSchedulerResultRef, "result_binding_ref_mismatch", "host:scheduler_apply_result_binding", "review_scheduler_apply_adapter_binding"},
		{result.ReadbackBindingRef, request.ExpectedReadbackRef, "readback_binding_ref_mismatch", "host:scheduler_apply_readback_binding", "review_scheduler_apply_adapter_binding"},
		{result.IdempotencyRef, request.IdempotencyRef, "idempotency_ref_mismatch", "host:scheduler_apply_idempotency_ref", "review_scheduler_apply_adapter_binding"},
		{result.CancellationRef, request.CancelPathRef, "cancellation_ref_mismatch", "host:scheduler_apply_cancel_path_ref", "review_scheduler_apply_adapter_binding"},
		{result.DeleteRef, request.DeletePathRef, "delete_ref_mismatch", "host:scheduler_apply_delete_path_ref", "review_scheduler_apply_adapter_binding"},
		{result.DisableRef, request.DisablePathRef, "disable_ref_mismatch", "host:scheduler_apply_disable_path_ref", "review_scheduler_apply_adapter_binding"},
	} {
		if check.got != "" && check.want != "" && check.got != check.want {
			result = hostOwnedSchedulerApplyAdapterReadinessBlock(result, FailureVerificationFailed, check.reason, check.missing, check.next)
		}
	}
	if len(result.MissingInputs) == 0 && len(result.BlockedReasons) == 0 {
		result.Status = HostActionReady
		result.ReadyForHostSchedulerAdapterInvocation = true
		result.HostSchedulerAdapterInvocationAuthorized = true
		result.HostMayInvokeSchedulerAdapter = true
		result.NextHostAction = "host_may_invoke_scheduler_apply_adapter"
		result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_host_owned_scheduler_apply_adapter_invocation", "host_may_invoke_scheduler_apply_adapter")
	}
	return result.Normalize()
}

func BuildHostOwnedSchedulerApplyAdapterInvocation(input HostOwnedSchedulerApplyAdapterInvocationInput) HostOwnedSchedulerApplyAdapterInvocation {
	readiness := input.Readiness.Normalize()
	result := HostOwnedSchedulerApplyAdapterInvocation{
		ContractVersion:                ContractVersion,
		Projected:                      true,
		Status:                         HostActionBlocked,
		Readiness:                      readiness,
		Action:                         readiness.Action,
		SchedulerApplyRequestRef:       readiness.SchedulerApplyRequestRef,
		SchedulerAdapterRef:            readiness.SchedulerAdapterRef,
		ExpectedScheduleRef:            readiness.ExpectedScheduleRef,
		ExpectedLifecycleStateRef:      readiness.ExpectedLifecycleStateRef,
		ExpectedSchedulerResultRef:     readiness.ExpectedSchedulerResultRef,
		ExpectedReadbackRef:            readiness.ExpectedReadbackRef,
		AdapterRef:                     readiness.AdapterRef,
		AdapterVersionRef:              readiness.AdapterVersionRef,
		InvocationRef:                  readiness.InvocationRef,
		InvocationReportRef:            normalizeOneDisplaySafeRef(input.InvocationReportRef),
		ObservedInvocationRef:          normalizeOneDisplaySafeRef(input.ObservedInvocationRef),
		HostSchedulerAdapterRunRef:     normalizeOneDisplaySafeRef(input.HostSchedulerAdapterRunRef),
		SchedulerApplyResultRef:        normalizeOneDisplaySafeRef(input.SchedulerApplyResultRef),
		SchedulerReadbackRef:           normalizeOneDisplaySafeRef(input.SchedulerReadbackRef),
		AppliedScheduleRef:             normalizeOneDisplaySafeRef(input.AppliedScheduleRef),
		AppliedLifecycleStateRef:       normalizeOneDisplaySafeRef(input.AppliedLifecycleStateRef),
		FailureRef:                     normalizeOneDisplaySafeRef(input.FailureRef),
		CompensationRef:                normalizeOneDisplaySafeRef(input.CompensationRef),
		SchedulerEvidenceRefs:          normalizeDisplaySafeRefs(input.SchedulerEvidenceRefs),
		FailureClass:                   FailureNone,
		HostAdapterInvocationReported:  input.HostAdapterInvocationReported,
		HostAdapterInvocationCompleted: input.HostAdapterInvocationCompleted,
		HostAdapterInvocationFailed:    input.HostAdapterInvocationFailed,
		Boundaries:                     hostOwnedSchedulerApplyAdapterInvocationBoundaries(readiness.Boundaries, input.Boundaries),
		NextHostAction:                 "provide_scheduler_apply_adapter_invocation_report",
		RunnerEffect:                   "none",
		PromptEffect:                   "none",
		RuntimeEffect:                  "none",
		RawOutputLoaded:                input.RawOutputLoaded || readiness.RawOutputLoaded,
	}
	if hostOwnedSchedulerApplyAdapterInvocationUnsafe(input, readiness) {
		result.RawOutputLoaded = true
		result = hostOwnedSchedulerApplyAdapterInvocationBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if !readiness.ReadyForHostSchedulerAdapterInvocation {
		result = hostOwnedSchedulerApplyAdapterInvocationBlock(result, firstFailureClass(readiness.FailureClass, FailureConfigMissing), "adapter_readiness_not_ready", "host:scheduler_apply_adapter_readiness", firstNextHostAction(readiness.NextHostAction, "review_scheduler_apply_adapter_readiness"))
	}
	if result.InvocationReportRef == "" {
		result = hostOwnedSchedulerApplyAdapterInvocationBlock(result, FailureEvidenceMissing, "invocation_report_ref_missing", "host:scheduler_apply_adapter_invocation_report_ref", "provide_scheduler_apply_adapter_invocation_report")
	}
	if result.ObservedInvocationRef == "" {
		result = hostOwnedSchedulerApplyAdapterInvocationBlock(result, FailureEvidenceMissing, "observed_invocation_ref_missing", "host:scheduler_apply_adapter_invocation_ref", "provide_scheduler_apply_adapter_invocation_report")
	} else if result.InvocationRef != "" && result.ObservedInvocationRef != result.InvocationRef {
		result = hostOwnedSchedulerApplyAdapterInvocationBlock(result, FailureVerificationFailed, "observed_invocation_ref_mismatch", "host:scheduler_apply_adapter_invocation_ref", "review_scheduler_apply_adapter_invocation_report")
	}
	if result.HostSchedulerAdapterRunRef == "" {
		result = hostOwnedSchedulerApplyAdapterInvocationBlock(result, FailureEvidenceMissing, "host_scheduler_adapter_run_ref_missing", "host:scheduler_apply_adapter_run_ref", "provide_scheduler_apply_adapter_invocation_report")
	}
	if !result.HostAdapterInvocationReported {
		result = hostOwnedSchedulerApplyAdapterInvocationBlock(result, FailureEvidenceMissing, "host_adapter_invocation_not_reported", "host:scheduler_apply_adapter_invocation_report", "provide_scheduler_apply_adapter_invocation_report")
	}
	if !result.HostAdapterInvocationCompleted && !result.HostAdapterInvocationFailed {
		result = hostOwnedSchedulerApplyAdapterInvocationBlock(result, FailureEvidenceMissing, "host_adapter_invocation_status_missing", "host:scheduler_apply_adapter_invocation_status", "provide_scheduler_apply_adapter_invocation_report")
	}
	if result.HostAdapterInvocationCompleted && result.HostAdapterInvocationFailed {
		result = hostOwnedSchedulerApplyAdapterInvocationBlock(result, FailureInvalidInput, "host_adapter_invocation_status_conflict", "host:scheduler_apply_adapter_invocation_status", "review_scheduler_apply_adapter_invocation_report")
	}
	if result.HostAdapterInvocationFailed {
		if result.FailureRef == "" {
			result = hostOwnedSchedulerApplyAdapterInvocationBlock(result, FailureEvidenceMissing, "failure_ref_missing", "host:scheduler_apply_adapter_failure_ref", "provide_scheduler_apply_adapter_failure_ref")
		}
	} else {
		for _, check := range []struct {
			got     DisplaySafeRef
			want    DisplaySafeRef
			reason  string
			missing MissingInput
			next    NextHostAction
		}{
			{result.SchedulerApplyResultRef, readiness.ExpectedSchedulerResultRef, "scheduler_apply_result_ref_mismatch", "host:scheduler_apply_result_ref", "review_scheduler_apply_adapter_invocation_report"},
			{result.SchedulerReadbackRef, readiness.ExpectedReadbackRef, "scheduler_readback_ref_mismatch", "host:scheduler_readback_ref", "review_scheduler_apply_adapter_invocation_report"},
			{result.AppliedScheduleRef, readiness.ExpectedScheduleRef, "applied_schedule_ref_mismatch", "host:applied_schedule_ref", "review_scheduler_apply_adapter_invocation_report"},
			{result.AppliedLifecycleStateRef, readiness.ExpectedLifecycleStateRef, "applied_lifecycle_state_ref_mismatch", "host:applied_schedule_lifecycle_state_ref", "review_scheduler_apply_adapter_invocation_report"},
		} {
			if check.got == "" {
				result = hostOwnedSchedulerApplyAdapterInvocationBlock(result, FailureEvidenceMissing, strings.TrimSuffix(check.reason, "_mismatch")+"_missing", check.missing, "provide_scheduler_apply_adapter_invocation_report")
			} else if check.want != "" && check.got != check.want {
				result = hostOwnedSchedulerApplyAdapterInvocationBlock(result, FailureVerificationFailed, check.reason, check.missing, check.next)
			}
		}
	}
	if len(result.MissingInputs) == 0 && len(result.BlockedReasons) == 0 {
		result.Status = HostActionRecorded
		if result.HostAdapterInvocationFailed {
			result.ReadyForFailureReview = true
			result.FailureClass = firstFailureClass(result.FailureClass, FailureVerificationFailed)
			result.NextHostAction = "review_scheduler_apply_adapter_failure"
			result.Boundaries = AppendBoundaries(result.Boundaries, "host_owned_scheduler_apply_adapter_failure_recorded")
		} else {
			result.ReadyForSchedulerApplyResult = true
			result.NextHostAction = "build_scheduler_apply_result"
			result.Boundaries = AppendBoundaries(result.Boundaries, "host_owned_scheduler_apply_adapter_invocation_recorded")
		}
	}
	return result.Normalize()
}

func CloneHostOwnedSchedulerApplyAdapterReadiness(in HostOwnedSchedulerApplyAdapterReadiness) HostOwnedSchedulerApplyAdapterReadiness {
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

func (r HostOwnedSchedulerApplyAdapterReadiness) Clone() HostOwnedSchedulerApplyAdapterReadiness {
	return CloneHostOwnedSchedulerApplyAdapterReadiness(r)
}

func (r HostOwnedSchedulerApplyAdapterReadiness) Normalize() HostOwnedSchedulerApplyAdapterReadiness {
	out := CloneHostOwnedSchedulerApplyAdapterReadiness(r)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.Request = out.Request.Normalize()
	out.Action = NormalizeSchedulerApplyAction(string(out.Action))
	out.SchedulerApplyRequestRef = normalizeOneDisplaySafeRef(out.SchedulerApplyRequestRef)
	out.SchedulerAdapterRef = normalizeOneDisplaySafeRef(out.SchedulerAdapterRef)
	out.ExpectedScheduleRef = normalizeOneDisplaySafeRef(out.ExpectedScheduleRef)
	out.ExpectedLifecycleStateRef = normalizeOneDisplaySafeRef(out.ExpectedLifecycleStateRef)
	out.ExpectedSchedulerResultRef = normalizeOneDisplaySafeRef(out.ExpectedSchedulerResultRef)
	out.ExpectedReadbackRef = normalizeOneDisplaySafeRef(out.ExpectedReadbackRef)
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
	out.CancellationRef = normalizeOneDisplaySafeRef(out.CancellationRef)
	out.DeleteRef = normalizeOneDisplaySafeRef(out.DeleteRef)
	out.DisableRef = normalizeOneDisplaySafeRef(out.DisableRef)
	out.ApprovalRefs = normalizeDisplaySafeRefs(out.ApprovalRefs)
	out.PolicyRefs = normalizeDisplaySafeRefs(out.PolicyRefs)
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.BlockedReasons = normalizeControlTokenList(out.BlockedReasons)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.DecisionBasis = normalizeDisplaySafeRefs(out.DecisionBasis)
	out.Boundaries = hostOwnedSchedulerApplyAdapterReadinessBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	out.RuntimeEffect = normalizeControlToken(out.RuntimeEffect)
	out = hostOwnedSchedulerApplyAdapterNormalizeReadinessEffects(out)
	out.ReadyForHostSchedulerAdapterInvocation = out.Status == HostActionReady && len(out.MissingInputs) == 0 && len(out.BlockedReasons) == 0 && !out.RawOutputLoaded
	out.HostSchedulerAdapterInvocationAuthorized = out.ReadyForHostSchedulerAdapterInvocation
	out.HostMayInvokeSchedulerAdapter = out.ReadyForHostSchedulerAdapterInvocation
	if out.RawOutputLoaded {
		out.Status = HostActionReviewRequired
		out.ReadyForHostSchedulerAdapterInvocation = false
		out.HostSchedulerAdapterInvocationAuthorized = false
		out.HostMayInvokeSchedulerAdapter = false
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

func CloneHostOwnedSchedulerApplyAdapterInvocation(in HostOwnedSchedulerApplyAdapterInvocation) HostOwnedSchedulerApplyAdapterInvocation {
	out := in
	out.Readiness = in.Readiness.Clone()
	out.SchedulerEvidenceRefs = cloneDisplaySafeRefs(in.SchedulerEvidenceRefs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (r HostOwnedSchedulerApplyAdapterInvocation) Clone() HostOwnedSchedulerApplyAdapterInvocation {
	return CloneHostOwnedSchedulerApplyAdapterInvocation(r)
}

func (r HostOwnedSchedulerApplyAdapterInvocation) Normalize() HostOwnedSchedulerApplyAdapterInvocation {
	out := CloneHostOwnedSchedulerApplyAdapterInvocation(r)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.Readiness = out.Readiness.Normalize()
	out.Action = NormalizeSchedulerApplyAction(string(out.Action))
	out.SchedulerApplyRequestRef = normalizeOneDisplaySafeRef(out.SchedulerApplyRequestRef)
	out.SchedulerAdapterRef = normalizeOneDisplaySafeRef(out.SchedulerAdapterRef)
	out.ExpectedScheduleRef = normalizeOneDisplaySafeRef(out.ExpectedScheduleRef)
	out.ExpectedLifecycleStateRef = normalizeOneDisplaySafeRef(out.ExpectedLifecycleStateRef)
	out.ExpectedSchedulerResultRef = normalizeOneDisplaySafeRef(out.ExpectedSchedulerResultRef)
	out.ExpectedReadbackRef = normalizeOneDisplaySafeRef(out.ExpectedReadbackRef)
	out.AdapterRef = normalizeOneDisplaySafeRef(out.AdapterRef)
	out.AdapterVersionRef = normalizeOneDisplaySafeRef(out.AdapterVersionRef)
	out.InvocationRef = normalizeOneDisplaySafeRef(out.InvocationRef)
	out.InvocationReportRef = normalizeOneDisplaySafeRef(out.InvocationReportRef)
	out.ObservedInvocationRef = normalizeOneDisplaySafeRef(out.ObservedInvocationRef)
	out.HostSchedulerAdapterRunRef = normalizeOneDisplaySafeRef(out.HostSchedulerAdapterRunRef)
	out.SchedulerApplyResultRef = normalizeOneDisplaySafeRef(out.SchedulerApplyResultRef)
	out.SchedulerReadbackRef = normalizeOneDisplaySafeRef(out.SchedulerReadbackRef)
	out.AppliedScheduleRef = normalizeOneDisplaySafeRef(out.AppliedScheduleRef)
	out.AppliedLifecycleStateRef = normalizeOneDisplaySafeRef(out.AppliedLifecycleStateRef)
	out.FailureRef = normalizeOneDisplaySafeRef(out.FailureRef)
	out.CompensationRef = normalizeOneDisplaySafeRef(out.CompensationRef)
	out.SchedulerEvidenceRefs = normalizeDisplaySafeRefs(out.SchedulerEvidenceRefs)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.BlockedReasons = normalizeControlTokenList(out.BlockedReasons)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.Boundaries = hostOwnedSchedulerApplyAdapterInvocationBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	out.RuntimeEffect = normalizeControlToken(out.RuntimeEffect)
	out = hostOwnedSchedulerApplyAdapterNormalizeInvocationEffects(out)
	out.ReadyForSchedulerApplyResult = out.Status == HostActionRecorded && out.HostAdapterInvocationCompleted && !out.HostAdapterInvocationFailed && len(out.MissingInputs) == 0 && len(out.BlockedReasons) == 0 && !out.RawOutputLoaded
	out.ReadyForFailureReview = out.Status == HostActionRecorded && out.HostAdapterInvocationFailed && len(out.MissingInputs) == 0 && len(out.BlockedReasons) == 0 && !out.RawOutputLoaded
	if out.RawOutputLoaded {
		out.Status = HostActionReviewRequired
		out.ReadyForSchedulerApplyResult = false
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

func hostOwnedSchedulerApplyAdapterReadinessBlock(result HostOwnedSchedulerApplyAdapterReadiness, failure FailureClass, reason string, missing MissingInput, next NextHostAction) HostOwnedSchedulerApplyAdapterReadiness {
	result.Status = HostActionBlocked
	result.ReadyForHostSchedulerAdapterInvocation = false
	result.HostSchedulerAdapterInvocationAuthorized = false
	result.HostMayInvokeSchedulerAdapter = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	return result
}

func hostOwnedSchedulerApplyAdapterInvocationBlock(result HostOwnedSchedulerApplyAdapterInvocation, failure FailureClass, reason string, missing MissingInput, next NextHostAction) HostOwnedSchedulerApplyAdapterInvocation {
	result.Status = HostActionBlocked
	result.ReadyForSchedulerApplyResult = false
	result.ReadyForFailureReview = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	return result
}

func hostOwnedSchedulerApplyAdapterReadinessUnsafe(input HostOwnedSchedulerApplyAdapterReadinessInput, request HostOwnedSchedulerApplyRequest) bool {
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
		displaySafeRefRejected(input.CancellationRef) ||
		displaySafeRefRejected(input.DeleteRef) ||
		displaySafeRefRejected(input.DisableRef) ||
		displaySafeRefSliceRejected(input.ApprovalRefs) ||
		displaySafeRefSliceRejected(input.PolicyRefs) ||
		displaySafeRefSliceRejected(input.DecisionBasis) ||
		evidenceRefRejected(input.EvidenceRefs) ||
		hostOwnedSchedulerApplyRequestOutputUnsafe(request)
}

func hostOwnedSchedulerApplyAdapterInvocationUnsafe(input HostOwnedSchedulerApplyAdapterInvocationInput, readiness HostOwnedSchedulerApplyAdapterReadiness) bool {
	return input.RawOutputLoaded ||
		readiness.RawOutputLoaded ||
		displaySafeRefRejected(input.InvocationReportRef) ||
		displaySafeRefRejected(input.ObservedInvocationRef) ||
		displaySafeRefRejected(input.HostSchedulerAdapterRunRef) ||
		displaySafeRefRejected(input.SchedulerApplyResultRef) ||
		displaySafeRefRejected(input.SchedulerReadbackRef) ||
		displaySafeRefRejected(input.AppliedScheduleRef) ||
		displaySafeRefRejected(input.AppliedLifecycleStateRef) ||
		displaySafeRefRejected(input.FailureRef) ||
		displaySafeRefRejected(input.CompensationRef) ||
		displaySafeRefSliceRejected(input.SchedulerEvidenceRefs) ||
		hostOwnedSchedulerApplyAdapterReadinessOutputUnsafe(readiness)
}

func hostOwnedSchedulerApplyAdapterReadinessBoundaries(groups ...[]Boundary) []Boundary {
	all := append([][]Boundary{{
		"host_owned_scheduler_apply_adapter_gate",
		"host_owned_scheduler_apply_adapter_readiness",
		"scheduler_apply_adapter_invocation_gate",
		"explicit_host_confirmation_required",
		"scheduler_apply_adapter_dry_run_required",
		"host_owned_scheduler_apply_request_required",
		"host_adapter_may_apply_scheduler_after_approval",
		"display_safe_refs_only",
		"no_scheduler_mutation_by_core",
		"no_core_schedule_created",
		"no_automation_created_by_core",
		"no_core_execution",
		"no_runner_dispatch",
		"no_runtime_adapter_execution",
		"no_install_apply",
		"no_workflow_dispatch",
		"no_worker_dispatch",
		"no_store_mutation_by_core",
		"no_compensation_execution",
		"projection_only",
	}}, groups...)
	return MergeBoundaries(all...)
}

func hostOwnedSchedulerApplyAdapterInvocationBoundaries(groups ...[]Boundary) []Boundary {
	all := append([][]Boundary{{
		"host_owned_scheduler_apply_adapter_gate",
		"host_owned_scheduler_apply_adapter_invocation_report",
		"host_scheduler_adapter_invocation_report_only",
		"host_adapter_scheduler_mutation_reported_only",
		"scheduler_apply_result_requires_readback",
		"schedule_proposal_not_schedule_created_by_core",
		"display_safe_refs_only",
		"no_scheduler_mutation_by_core",
		"no_core_schedule_created",
		"no_automation_created_by_core",
		"no_core_execution",
		"no_runner_dispatch",
		"no_runtime_adapter_execution",
		"no_install_apply",
		"no_workflow_dispatch",
		"no_worker_dispatch",
		"no_store_mutation_by_core",
		"no_compensation_execution",
		"projection_only",
	}}, groups...)
	return MergeBoundaries(all...)
}

func hostOwnedSchedulerApplyAdapterNormalizeReadinessEffects(value HostOwnedSchedulerApplyAdapterReadiness) HostOwnedSchedulerApplyAdapterReadiness {
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
	value.SchedulerMutationByCore = false
	value.CoreScheduleCreated = false
	value.AutomationCreatedByCore = false
	value.RunnerDispatched = false
	value.RuntimeAdapterExecuted = false
	value.InstallerExecuted = false
	value.WorkflowDispatched = false
	value.WorkerDispatched = false
	value.StoreMutationExecuted = false
	value.CompensationExecuted = false
	return value
}

func hostOwnedSchedulerApplyAdapterNormalizeInvocationEffects(value HostOwnedSchedulerApplyAdapterInvocation) HostOwnedSchedulerApplyAdapterInvocation {
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
	value.SchedulerMutationByCore = false
	value.CoreScheduleCreated = false
	value.AutomationCreatedByCore = false
	value.RunnerDispatched = false
	value.RuntimeAdapterExecuted = false
	value.InstallerExecuted = false
	value.WorkflowDispatched = false
	value.WorkerDispatched = false
	value.StoreMutationExecuted = false
	value.CompensationExecuted = false
	return value
}

func hostOwnedSchedulerApplyAdapterReadinessOutputUnsafe(input HostOwnedSchedulerApplyAdapterReadiness) bool {
	return displaySafeRefRejected(input.SchedulerApplyRequestRef) ||
		displaySafeRefRejected(input.SchedulerAdapterRef) ||
		displaySafeRefRejected(input.ExpectedScheduleRef) ||
		displaySafeRefRejected(input.ExpectedLifecycleStateRef) ||
		displaySafeRefRejected(input.ExpectedSchedulerResultRef) ||
		displaySafeRefRejected(input.ExpectedReadbackRef) ||
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
		displaySafeRefRejected(input.CancellationRef) ||
		displaySafeRefRejected(input.DeleteRef) ||
		displaySafeRefRejected(input.DisableRef) ||
		displaySafeRefSliceRejected(input.ApprovalRefs) ||
		displaySafeRefSliceRejected(input.PolicyRefs) ||
		displaySafeRefSliceRejected(input.DecisionBasis) ||
		evidenceRefRejected(input.EvidenceRefs) ||
		hostOwnedSchedulerApplyRequestOutputUnsafe(input.Request)
}
