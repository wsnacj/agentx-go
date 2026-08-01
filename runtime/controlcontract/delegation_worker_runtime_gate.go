package controlcontract

type HostOwnedDelegationWorkerRuntimeReadinessInput struct {
	Request              DelegationRequestProjection            `json:"request,omitempty"`
	WorkerRuntimeGate    ProductionAdapterIndependentEffectGate `json:"worker_runtime_gate,omitempty"`
	AdapterRef           DisplaySafeRef                         `json:"adapter_ref,omitempty"`
	AdapterVersionRef    DisplaySafeRef                         `json:"adapter_version_ref,omitempty"`
	AdapterCapabilityRef DisplaySafeRef                         `json:"adapter_capability_ref,omitempty"`
	AdapterContractRef   DisplaySafeRef                         `json:"adapter_contract_ref,omitempty"`
	HostConfirmationRef  DisplaySafeRef                         `json:"host_confirmation_ref,omitempty"`
	WorkerRunRef         DisplaySafeRef                         `json:"worker_run_ref,omitempty"`
	WorkerRequestRef     DisplaySafeRef                         `json:"worker_request_ref,omitempty"`
	InvocationRef        DisplaySafeRef                         `json:"invocation_ref,omitempty"`
	ResultBindingRef     DisplaySafeRef                         `json:"result_binding_ref,omitempty"`
	ReadbackBindingRef   DisplaySafeRef                         `json:"readback_binding_ref,omitempty"`
	IdempotencyRef       DisplaySafeRef                         `json:"idempotency_ref,omitempty"`
	BudgetRef            DisplaySafeRef                         `json:"budget_ref,omitempty"`
	VerificationRef      DisplaySafeRef                         `json:"verification_ref,omitempty"`
	FailureBindingRef    DisplaySafeRef                         `json:"failure_binding_ref,omitempty"`
	CompensationRef      DisplaySafeRef                         `json:"compensation_ref,omitempty"`
	EvidenceRefs         []EvidenceRef                          `json:"evidence_refs,omitempty"`
	DecisionBasis        []DisplaySafeRef                       `json:"decision_basis,omitempty"`
	Boundaries           []Boundary                             `json:"boundaries,omitempty"`
	RawOutputLoaded      bool                                   `json:"raw_output_loaded"`
}

type HostOwnedDelegationWorkerRuntimeReadiness struct {
	ContractVersion                       string                                 `json:"contract_version,omitempty"`
	Projected                             bool                                   `json:"projected"`
	Status                                HostActionStatus                       `json:"status,omitempty"`
	ReadyForHostWorkerRuntimeInvocation   bool                                   `json:"ready_for_host_worker_runtime_invocation"`
	ReadyForL4HostApprovedWorker          bool                                   `json:"ready_for_l4_host_approved_worker"`
	ReadyForL5ExplicitWorker              bool                                   `json:"ready_for_l5_explicit_worker"`
	HostWorkerRuntimeInvocationAuthorized bool                                   `json:"host_worker_runtime_invocation_authorized"`
	HostMayInvokeWorkerRuntime            bool                                   `json:"host_may_invoke_worker_runtime"`
	Request                               DelegationRequestProjection            `json:"request,omitempty"`
	WorkerRuntimeGate                     ProductionAdapterIndependentEffectGate `json:"worker_runtime_gate,omitempty"`
	SubgoalRef                            DisplaySafeRef                         `json:"subgoal_ref,omitempty"`
	WorkerRef                             DisplaySafeRef                         `json:"worker_ref,omitempty"`
	WorkerRunRef                          DisplaySafeRef                         `json:"worker_run_ref,omitempty"`
	WorkerRequestRef                      DisplaySafeRef                         `json:"worker_request_ref,omitempty"`
	AdapterRef                            DisplaySafeRef                         `json:"adapter_ref,omitempty"`
	AdapterVersionRef                     DisplaySafeRef                         `json:"adapter_version_ref,omitempty"`
	AdapterCapabilityRef                  DisplaySafeRef                         `json:"adapter_capability_ref,omitempty"`
	AdapterContractRef                    DisplaySafeRef                         `json:"adapter_contract_ref,omitempty"`
	HostConfirmationRef                   DisplaySafeRef                         `json:"host_confirmation_ref,omitempty"`
	InvocationRef                         DisplaySafeRef                         `json:"invocation_ref,omitempty"`
	ResultBindingRef                      DisplaySafeRef                         `json:"result_binding_ref,omitempty"`
	ReadbackBindingRef                    DisplaySafeRef                         `json:"readback_binding_ref,omitempty"`
	IdempotencyRef                        DisplaySafeRef                         `json:"idempotency_ref,omitempty"`
	BudgetRef                             DisplaySafeRef                         `json:"budget_ref,omitempty"`
	VerificationRef                       DisplaySafeRef                         `json:"verification_ref,omitempty"`
	FailureBindingRef                     DisplaySafeRef                         `json:"failure_binding_ref,omitempty"`
	CompensationRef                       DisplaySafeRef                         `json:"compensation_ref,omitempty"`
	EvidenceRefs                          []EvidenceRef                          `json:"evidence_refs,omitempty"`
	FailureClass                          FailureClass                           `json:"failure_class,omitempty"`
	BlockedReasons                        []string                               `json:"blocked_reasons,omitempty"`
	MissingInputs                         []MissingInput                         `json:"missing_inputs,omitempty"`
	DecisionBasis                         []DisplaySafeRef                       `json:"decision_basis,omitempty"`
	Boundaries                            []Boundary                             `json:"boundaries,omitempty"`
	NextHostAction                        NextHostAction                         `json:"next_host_action,omitempty"`
	ExecutorEffect                        string                                 `json:"executor_effect,omitempty"`
	RunnerEffect                          string                                 `json:"runner_effect,omitempty"`
	PromptEffect                          string                                 `json:"prompt_effect,omitempty"`
	RuntimeEffect                         string                                 `json:"runtime_effect,omitempty"`
	CoreWorkerRuntimeInvoked              bool                                   `json:"core_worker_runtime_invoked"`
	WorkerDispatched                      bool                                   `json:"worker_dispatched"`
	RunnerDispatched                      bool                                   `json:"runner_dispatched"`
	RuntimeAdapterExecuted                bool                                   `json:"runtime_adapter_executed"`
	WorkflowDispatched                    bool                                   `json:"workflow_dispatched"`
	SchedulerApplied                      bool                                   `json:"scheduler_applied"`
	InstallerExecuted                     bool                                   `json:"installer_executed"`
	StoreMutationExecuted                 bool                                   `json:"store_mutation_executed"`
	CompensationExecuted                  bool                                   `json:"compensation_executed"`
	WorkerResultRequiresVerification      bool                                   `json:"worker_result_requires_verification"`
	WorkerOutputAcceptedAsFact            bool                                   `json:"worker_output_accepted_as_fact"`
	RawOutputLoaded                       bool                                   `json:"raw_output_loaded"`
}

type HostOwnedDelegationWorkerRuntimeInvocationInput struct {
	Readiness               HostOwnedDelegationWorkerRuntimeReadiness `json:"readiness,omitempty"`
	InvocationReportRef     DisplaySafeRef                            `json:"invocation_report_ref,omitempty"`
	ObservedInvocationRef   DisplaySafeRef                            `json:"observed_invocation_ref,omitempty"`
	HostWorkerRuntimeRunRef DisplaySafeRef                            `json:"host_worker_runtime_run_ref,omitempty"`
	ObservedWorkerRunRef    DisplaySafeRef                            `json:"observed_worker_run_ref,omitempty"`
	WorkerResultRef         DisplaySafeRef                            `json:"worker_result_ref,omitempty"`
	WorkerReadbackRef       DisplaySafeRef                            `json:"worker_readback_ref,omitempty"`
	ObservationRef          DisplaySafeRef                            `json:"observation_ref,omitempty"`
	FailureRef              DisplaySafeRef                            `json:"failure_ref,omitempty"`
	CompensationRef         DisplaySafeRef                            `json:"compensation_ref,omitempty"`
	HostInvocationReported  bool                                      `json:"host_invocation_reported"`
	HostInvocationCompleted bool                                      `json:"host_invocation_completed"`
	HostInvocationFailed    bool                                      `json:"host_invocation_failed"`
	EvidenceRefs            []EvidenceRef                             `json:"evidence_refs,omitempty"`
	Boundaries              []Boundary                                `json:"boundaries,omitempty"`
	RawOutputLoaded         bool                                      `json:"raw_output_loaded"`
}

type HostOwnedDelegationWorkerRuntimeInvocation struct {
	ContractVersion                  string                                    `json:"contract_version,omitempty"`
	Projected                        bool                                      `json:"projected"`
	Status                           HostActionStatus                          `json:"status,omitempty"`
	ReadyForWorkerResultReview       bool                                      `json:"ready_for_worker_result_review"`
	ReadyForFailureReview            bool                                      `json:"ready_for_failure_review"`
	HostInvocationReported           bool                                      `json:"host_invocation_reported"`
	HostInvocationCompleted          bool                                      `json:"host_invocation_completed"`
	HostInvocationFailed             bool                                      `json:"host_invocation_failed"`
	Readiness                        HostOwnedDelegationWorkerRuntimeReadiness `json:"readiness,omitempty"`
	SubgoalRef                       DisplaySafeRef                            `json:"subgoal_ref,omitempty"`
	WorkerRef                        DisplaySafeRef                            `json:"worker_ref,omitempty"`
	WorkerRunRef                     DisplaySafeRef                            `json:"worker_run_ref,omitempty"`
	WorkerRequestRef                 DisplaySafeRef                            `json:"worker_request_ref,omitempty"`
	AdapterRef                       DisplaySafeRef                            `json:"adapter_ref,omitempty"`
	AdapterVersionRef                DisplaySafeRef                            `json:"adapter_version_ref,omitempty"`
	InvocationRef                    DisplaySafeRef                            `json:"invocation_ref,omitempty"`
	InvocationReportRef              DisplaySafeRef                            `json:"invocation_report_ref,omitempty"`
	ObservedInvocationRef            DisplaySafeRef                            `json:"observed_invocation_ref,omitempty"`
	HostWorkerRuntimeRunRef          DisplaySafeRef                            `json:"host_worker_runtime_run_ref,omitempty"`
	ObservedWorkerRunRef             DisplaySafeRef                            `json:"observed_worker_run_ref,omitempty"`
	WorkerResultRef                  DisplaySafeRef                            `json:"worker_result_ref,omitempty"`
	WorkerReadbackRef                DisplaySafeRef                            `json:"worker_readback_ref,omitempty"`
	ObservationRef                   DisplaySafeRef                            `json:"observation_ref,omitempty"`
	FailureRef                       DisplaySafeRef                            `json:"failure_ref,omitempty"`
	CompensationRef                  DisplaySafeRef                            `json:"compensation_ref,omitempty"`
	ResultReview                     DelegationWorkerResultReview              `json:"result_review,omitempty"`
	EvidenceRefs                     []EvidenceRef                             `json:"evidence_refs,omitempty"`
	FailureClass                     FailureClass                              `json:"failure_class,omitempty"`
	BlockedReasons                   []string                                  `json:"blocked_reasons,omitempty"`
	MissingInputs                    []MissingInput                            `json:"missing_inputs,omitempty"`
	Boundaries                       []Boundary                                `json:"boundaries,omitempty"`
	NextHostAction                   NextHostAction                            `json:"next_host_action,omitempty"`
	ExecutorEffect                   string                                    `json:"executor_effect,omitempty"`
	RunnerEffect                     string                                    `json:"runner_effect,omitempty"`
	PromptEffect                     string                                    `json:"prompt_effect,omitempty"`
	RuntimeEffect                    string                                    `json:"runtime_effect,omitempty"`
	CoreWorkerRuntimeInvoked         bool                                      `json:"core_worker_runtime_invoked"`
	WorkerDispatched                 bool                                      `json:"worker_dispatched"`
	RunnerDispatched                 bool                                      `json:"runner_dispatched"`
	RuntimeAdapterExecuted           bool                                      `json:"runtime_adapter_executed"`
	WorkflowDispatched               bool                                      `json:"workflow_dispatched"`
	SchedulerApplied                 bool                                      `json:"scheduler_applied"`
	InstallerExecuted                bool                                      `json:"installer_executed"`
	StoreMutationExecuted            bool                                      `json:"store_mutation_executed"`
	CompensationExecuted             bool                                      `json:"compensation_executed"`
	WorkerResultRequiresVerification bool                                      `json:"worker_result_requires_verification"`
	WorkerOutputAcceptedAsFact       bool                                      `json:"worker_output_accepted_as_fact"`
	RawOutputLoaded                  bool                                      `json:"raw_output_loaded"`
}

func BuildHostOwnedDelegationWorkerRuntimeReadiness(input HostOwnedDelegationWorkerRuntimeReadinessInput) HostOwnedDelegationWorkerRuntimeReadiness {
	request := input.Request.Normalize()
	gate := input.WorkerRuntimeGate.Normalize()
	result := HostOwnedDelegationWorkerRuntimeReadiness{
		ContractVersion:      ContractVersion,
		Projected:            true,
		Status:               HostActionBlocked,
		Request:              request,
		WorkerRuntimeGate:    gate,
		SubgoalRef:           request.SubgoalRef,
		WorkerRef:            request.WorkerRef,
		WorkerRunRef:         normalizeOneDisplaySafeRef(input.WorkerRunRef),
		WorkerRequestRef:     normalizeOneDisplaySafeRef(input.WorkerRequestRef),
		AdapterRef:           normalizeOneDisplaySafeRef(input.AdapterRef),
		AdapterVersionRef:    normalizeOneDisplaySafeRef(input.AdapterVersionRef),
		AdapterCapabilityRef: normalizeOneDisplaySafeRef(input.AdapterCapabilityRef),
		AdapterContractRef:   normalizeOneDisplaySafeRef(input.AdapterContractRef),
		HostConfirmationRef:  normalizeOneDisplaySafeRef(input.HostConfirmationRef),
		InvocationRef:        normalizeOneDisplaySafeRef(input.InvocationRef),
		ResultBindingRef:     normalizeOneDisplaySafeRef(input.ResultBindingRef),
		ReadbackBindingRef:   normalizeOneDisplaySafeRef(input.ReadbackBindingRef),
		IdempotencyRef:       normalizeOneDisplaySafeRef(input.IdempotencyRef),
		BudgetRef:            normalizeOneDisplaySafeRef(input.BudgetRef),
		VerificationRef:      normalizeOneDisplaySafeRef(input.VerificationRef),
		FailureBindingRef:    normalizeOneDisplaySafeRef(input.FailureBindingRef),
		CompensationRef:      normalizeOneDisplaySafeRef(input.CompensationRef),
		EvidenceRefs:         MergeEvidenceRefs(input.EvidenceRefs, request.EvidenceRequirements),
		FailureClass:         firstFailureClass(request.FailureClass, gate.FailureClass, FailureNone),
		DecisionBasis: normalizeDisplaySafeRefs(append(
			[]DisplaySafeRef{
				"delegation_worker_runtime:host_owned",
				"delegation_worker_runtime:readiness",
			},
			input.DecisionBasis...,
		)),
		Boundaries:                       hostOwnedDelegationWorkerRuntimeReadinessBoundaries(request.Boundaries, gate.Boundaries, input.Boundaries),
		NextHostAction:                   "provide_delegation_worker_runtime_binding",
		ExecutorEffect:                   "none",
		RunnerEffect:                     "none",
		PromptEffect:                     "none",
		RuntimeEffect:                    "none",
		WorkerResultRequiresVerification: true,
		WorkerOutputAcceptedAsFact:       false,
		RawOutputLoaded: input.RawOutputLoaded ||
			request.RawOutputLoaded ||
			gate.RawOutputLoaded,
	}
	if hostOwnedDelegationWorkerRuntimeReadinessUnsafe(input, request, gate) {
		result.RawOutputLoaded = true
		result = hostOwnedDelegationWorkerRuntimeReadinessBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if !request.ReadyForWorkerDispatch {
		result.MissingInputs = AppendMissingInputs(result.MissingInputs, request.MissingInputs...)
		result.BlockedReasons = appendUniqueControlTokens(result.BlockedReasons, request.BlockedReasons)
		result = hostOwnedDelegationWorkerRuntimeReadinessBlock(result, firstFailureClass(request.FailureClass, FailurePolicyBlocked), "delegation_request_not_ready", "host:delegation_request", firstNextHostAction(request.NextHostAction, "review_delegation_request"))
	}
	if request.RequestedIntensity == IntensityL4DurableLongRun {
		if !request.HostAllowsL4Delegation || !request.HostApproved {
			result = hostOwnedDelegationWorkerRuntimeReadinessBlock(result, FailureApprovalRequired, "l4_delegation_not_host_approved", "host:l4_delegation_approval", "request_host_approval")
		}
	} else if request.RequestedIntensity == IntensityL5Autonomous {
		if !request.L5Enabled || !request.HostApproved {
			result = hostOwnedDelegationWorkerRuntimeReadinessBlock(result, FailureApprovalRequired, "l5_delegation_not_explicitly_enabled", "host:l5_delegation_policy", "request_host_approval")
		}
	} else {
		result = hostOwnedDelegationWorkerRuntimeReadinessBlock(result, FailurePolicyBlocked, "delegation_worker_runtime_requires_l4_or_l5", "contract:intensity_l4_or_l5", "request_intensity_upgrade_confirmation")
	}
	if !hostOwnedDelegationWorkerRuntimeGateReady(gate) {
		result = hostOwnedDelegationWorkerRuntimeReadinessBlock(result, firstFailureClass(gate.FailureClass, FailureConfigMissing), "delegation_worker_runtime_gate_not_ready", "host:delegation_worker_runtime_gate", "provide_delegation_worker_runtime_gate")
	}
	for _, check := range []struct {
		ok      bool
		reason  string
		missing MissingInput
		next    NextHostAction
	}{
		{result.AdapterRef != "", "adapter_ref_missing", "host:delegation_worker_runtime_adapter_ref", "provide_delegation_worker_runtime_binding"},
		{result.AdapterVersionRef != "", "adapter_version_ref_missing", "host:delegation_worker_runtime_adapter_version_ref", "provide_delegation_worker_runtime_binding"},
		{result.AdapterCapabilityRef != "", "adapter_capability_ref_missing", "host:delegation_worker_runtime_adapter_capability_ref", "provide_delegation_worker_runtime_capability"},
		{result.AdapterContractRef != "", "adapter_contract_ref_missing", "contract:delegation_worker_runtime_adapter", "provide_delegation_worker_runtime_contract"},
		{result.HostConfirmationRef != "", "host_confirmation_ref_missing", "host:delegation_worker_runtime_confirmation", "request_delegation_worker_runtime_confirmation"},
		{result.WorkerRunRef != "", "worker_run_ref_missing", "host:delegation_worker_run_ref", "provide_delegation_worker_run_ref"},
		{result.WorkerRequestRef != "", "worker_request_ref_missing", "host:delegation_worker_request_ref", "provide_delegation_worker_request_ref"},
		{result.InvocationRef != "", "invocation_ref_missing", "host:delegation_worker_runtime_invocation_ref", "provide_delegation_worker_runtime_invocation_ref"},
		{result.ResultBindingRef != "", "result_binding_ref_missing", "host:delegation_worker_result_binding", "provide_delegation_worker_result_binding"},
		{result.ReadbackBindingRef != "", "readback_binding_ref_missing", "host:delegation_worker_readback_binding", "provide_delegation_worker_readback_binding"},
		{result.IdempotencyRef != "", "idempotency_ref_missing", "host:delegation_worker_idempotency", "provide_delegation_worker_idempotency"},
		{result.BudgetRef != "", "budget_ref_missing", "host:delegation_worker_budget", "provide_delegation_worker_budget"},
		{result.VerificationRef != "", "verification_ref_missing", "host:delegation_worker_parent_verification_ref", "provide_delegation_worker_parent_verification_ref"},
		{result.FailureBindingRef != "", "failure_binding_ref_missing", "host:delegation_worker_failure_binding", "provide_delegation_worker_failure_binding"},
	} {
		if !check.ok {
			result = hostOwnedDelegationWorkerRuntimeReadinessBlock(result, FailureConfigMissing, check.reason, check.missing, check.next)
		}
	}
	if len(result.MissingInputs) == 0 && len(result.BlockedReasons) == 0 {
		result.Status = HostActionReady
		result.ReadyForHostWorkerRuntimeInvocation = true
		result.ReadyForL4HostApprovedWorker = request.RequestedIntensity == IntensityL4DurableLongRun
		result.ReadyForL5ExplicitWorker = request.RequestedIntensity == IntensityL5Autonomous
		result.HostWorkerRuntimeInvocationAuthorized = true
		result.HostMayInvokeWorkerRuntime = true
		result.NextHostAction = "host_may_invoke_delegation_worker_runtime"
		result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_host_owned_delegation_worker_runtime_invocation", "host_may_invoke_delegation_worker_runtime")
		if result.ReadyForL4HostApprovedWorker {
			result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_l4_host_approved_delegation_worker")
		}
		if result.ReadyForL5ExplicitWorker {
			result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_l5_explicit_delegation_worker")
		}
	}
	return result.Normalize()
}

func BuildHostOwnedDelegationWorkerRuntimeInvocation(input HostOwnedDelegationWorkerRuntimeInvocationInput) HostOwnedDelegationWorkerRuntimeInvocation {
	readiness := input.Readiness.Normalize()
	result := HostOwnedDelegationWorkerRuntimeInvocation{
		ContractVersion:                  ContractVersion,
		Projected:                        true,
		Status:                           HostActionBlocked,
		Readiness:                        readiness,
		SubgoalRef:                       readiness.SubgoalRef,
		WorkerRef:                        readiness.WorkerRef,
		WorkerRunRef:                     readiness.WorkerRunRef,
		WorkerRequestRef:                 readiness.WorkerRequestRef,
		AdapterRef:                       readiness.AdapterRef,
		AdapterVersionRef:                readiness.AdapterVersionRef,
		InvocationRef:                    readiness.InvocationRef,
		InvocationReportRef:              normalizeOneDisplaySafeRef(input.InvocationReportRef),
		ObservedInvocationRef:            normalizeOneDisplaySafeRef(input.ObservedInvocationRef),
		HostWorkerRuntimeRunRef:          normalizeOneDisplaySafeRef(input.HostWorkerRuntimeRunRef),
		ObservedWorkerRunRef:             normalizeOneDisplaySafeRef(input.ObservedWorkerRunRef),
		WorkerResultRef:                  normalizeOneDisplaySafeRef(input.WorkerResultRef),
		WorkerReadbackRef:                normalizeOneDisplaySafeRef(input.WorkerReadbackRef),
		ObservationRef:                   normalizeOneDisplaySafeRef(input.ObservationRef),
		FailureRef:                       normalizeOneDisplaySafeRef(input.FailureRef),
		CompensationRef:                  normalizeOneDisplaySafeRef(input.CompensationRef),
		HostInvocationReported:           input.HostInvocationReported,
		HostInvocationCompleted:          input.HostInvocationCompleted,
		HostInvocationFailed:             input.HostInvocationFailed,
		EvidenceRefs:                     MergeEvidenceRefs(input.EvidenceRefs, readiness.EvidenceRefs),
		FailureClass:                     readiness.FailureClass,
		Boundaries:                       hostOwnedDelegationWorkerRuntimeInvocationBoundaries(readiness.Boundaries, input.Boundaries),
		NextHostAction:                   "provide_delegation_worker_runtime_invocation_report",
		ExecutorEffect:                   "none",
		RunnerEffect:                     "none",
		PromptEffect:                     "none",
		RuntimeEffect:                    "none",
		WorkerResultRequiresVerification: true,
		WorkerOutputAcceptedAsFact:       false,
		RawOutputLoaded:                  input.RawOutputLoaded || readiness.RawOutputLoaded,
	}
	if hostOwnedDelegationWorkerRuntimeInvocationUnsafe(input, readiness) {
		result.RawOutputLoaded = true
		result = hostOwnedDelegationWorkerRuntimeInvocationBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if !readiness.ReadyForHostWorkerRuntimeInvocation {
		result = hostOwnedDelegationWorkerRuntimeInvocationBlock(result, firstFailureClass(readiness.FailureClass, FailureConfigMissing), "delegation_worker_runtime_readiness_not_ready", "host:delegation_worker_runtime_readiness", firstNextHostAction(readiness.NextHostAction, "review_delegation_worker_runtime_readiness"))
	}
	for _, check := range []struct {
		ok      bool
		reason  string
		missing MissingInput
		next    NextHostAction
	}{
		{result.InvocationReportRef != "", "invocation_report_ref_missing", "host:delegation_worker_runtime_invocation_report", "provide_delegation_worker_runtime_invocation_report"},
		{result.ObservedInvocationRef != "", "observed_invocation_ref_missing", "host:delegation_worker_runtime_observed_invocation", "provide_delegation_worker_runtime_invocation_report"},
		{result.HostWorkerRuntimeRunRef != "", "host_worker_runtime_run_ref_missing", "host:delegation_worker_runtime_run_ref", "provide_delegation_worker_runtime_run_ref"},
		{input.HostInvocationReported, "host_invocation_not_reported", "host:delegation_worker_runtime_invocation_report", "provide_delegation_worker_runtime_invocation_report"},
	} {
		if !check.ok {
			result = hostOwnedDelegationWorkerRuntimeInvocationBlock(result, FailureConfigMissing, check.reason, check.missing, check.next)
		}
	}
	if result.ObservedInvocationRef != "" && result.InvocationRef != "" && result.ObservedInvocationRef != result.InvocationRef {
		result = hostOwnedDelegationWorkerRuntimeInvocationBlock(result, FailureVerificationFailed, "observed_invocation_ref_mismatch", "host:delegation_worker_runtime_invocation_report", "review_delegation_worker_runtime_invocation_report")
	}
	if result.ObservedWorkerRunRef != "" && result.WorkerRunRef != "" && result.ObservedWorkerRunRef != result.WorkerRunRef {
		result = hostOwnedDelegationWorkerRuntimeInvocationBlock(result, FailureVerificationFailed, "observed_worker_run_ref_mismatch", "host:delegation_worker_run_ref", "review_delegation_worker_runtime_invocation_report")
	}
	if input.HostInvocationFailed {
		if result.FailureRef == "" {
			result = hostOwnedDelegationWorkerRuntimeInvocationBlock(result, FailureVerificationFailed, "failure_ref_missing", "host:delegation_worker_failure_ref", "provide_delegation_worker_failure_review")
		}
		if result.CompensationRef == "" {
			result = hostOwnedDelegationWorkerRuntimeInvocationBlock(result, FailureConfigMissing, "compensation_ref_missing", "host:delegation_worker_compensation_ref", "provide_delegation_worker_compensation_ref")
		}
		result.ReadyForFailureReview = result.FailureRef != "" && result.CompensationRef != ""
		result.NextHostAction = "review_delegation_worker_runtime_failure"
		result.Boundaries = AppendBoundaries(result.Boundaries, "delegation_worker_runtime_failure_reported")
	}
	if input.HostInvocationCompleted {
		for _, check := range []struct {
			ok      bool
			reason  string
			missing MissingInput
			next    NextHostAction
		}{
			{result.WorkerResultRef != "", "worker_result_ref_missing", "host:delegation_worker_result_ref", "provide_delegation_worker_result_ref"},
			{result.WorkerReadbackRef != "", "worker_readback_ref_missing", "host:delegation_worker_readback_ref", "provide_delegation_worker_readback_ref"},
			{result.ObservedWorkerRunRef != "", "observed_worker_run_ref_missing", "host:delegation_worker_run_ref", "provide_delegation_worker_run_ref"},
			{result.ObservationRef != "", "observation_ref_missing", "host:delegation_worker_observation_ref", "provide_delegation_worker_observation_ref"},
		} {
			if !check.ok {
				result = hostOwnedDelegationWorkerRuntimeInvocationBlock(result, FailureConfigMissing, check.reason, check.missing, check.next)
			}
		}
	}
	if len(result.MissingInputs) == 0 && len(result.BlockedReasons) == 0 {
		result.ResultReview = BuildDelegationWorkerResultReview(DelegationWorkerResultReviewInput{
			Request:         readiness.Request,
			WorkerRunRef:    result.WorkerRunRef,
			WorkerResultRef: result.WorkerResultRef,
			EvidenceRefs:    result.EvidenceRefs,
			Boundaries: []Boundary{
				"delegation_worker_runtime_result_requires_parent_verification",
			},
		})
		result.Status = HostActionRecorded
		result.ReadyForWorkerResultReview = input.HostInvocationCompleted
		result.ReadyForFailureReview = input.HostInvocationFailed
		result.NextHostAction = "run_parent_verification_gate_for_delegation_worker_result"
		result.Boundaries = AppendBoundaries(result.Boundaries, "host_owned_delegation_worker_runtime_invocation_recorded")
		if result.ReadyForWorkerResultReview {
			result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_delegation_worker_result_review", "worker_result_requires_parent_verification")
		}
	}
	return result.Normalize()
}

func CloneHostOwnedDelegationWorkerRuntimeReadiness(in HostOwnedDelegationWorkerRuntimeReadiness) HostOwnedDelegationWorkerRuntimeReadiness {
	out := in
	out.Request = in.Request.Clone()
	out.WorkerRuntimeGate = in.WorkerRuntimeGate.Clone()
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.DecisionBasis = cloneDisplaySafeRefs(in.DecisionBasis)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (r HostOwnedDelegationWorkerRuntimeReadiness) Clone() HostOwnedDelegationWorkerRuntimeReadiness {
	return CloneHostOwnedDelegationWorkerRuntimeReadiness(r)
}

func (r HostOwnedDelegationWorkerRuntimeReadiness) Normalize() HostOwnedDelegationWorkerRuntimeReadiness {
	out := CloneHostOwnedDelegationWorkerRuntimeReadiness(r)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.Request = out.Request.Normalize()
	out.WorkerRuntimeGate = out.WorkerRuntimeGate.Normalize()
	out.SubgoalRef = normalizeOneDisplaySafeRef(out.SubgoalRef)
	out.WorkerRef = normalizeOneDisplaySafeRef(out.WorkerRef)
	out.WorkerRunRef = normalizeOneDisplaySafeRef(out.WorkerRunRef)
	out.WorkerRequestRef = normalizeOneDisplaySafeRef(out.WorkerRequestRef)
	out.AdapterRef = normalizeOneDisplaySafeRef(out.AdapterRef)
	out.AdapterVersionRef = normalizeOneDisplaySafeRef(out.AdapterVersionRef)
	out.AdapterCapabilityRef = normalizeOneDisplaySafeRef(out.AdapterCapabilityRef)
	out.AdapterContractRef = normalizeOneDisplaySafeRef(out.AdapterContractRef)
	out.HostConfirmationRef = normalizeOneDisplaySafeRef(out.HostConfirmationRef)
	out.InvocationRef = normalizeOneDisplaySafeRef(out.InvocationRef)
	out.ResultBindingRef = normalizeOneDisplaySafeRef(out.ResultBindingRef)
	out.ReadbackBindingRef = normalizeOneDisplaySafeRef(out.ReadbackBindingRef)
	out.IdempotencyRef = normalizeOneDisplaySafeRef(out.IdempotencyRef)
	out.BudgetRef = normalizeOneDisplaySafeRef(out.BudgetRef)
	out.VerificationRef = normalizeOneDisplaySafeRef(out.VerificationRef)
	out.FailureBindingRef = normalizeOneDisplaySafeRef(out.FailureBindingRef)
	out.CompensationRef = normalizeOneDisplaySafeRef(out.CompensationRef)
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.BlockedReasons = normalizeControlTokenList(out.BlockedReasons)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.DecisionBasis = normalizeDisplaySafeRefs(out.DecisionBasis)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	out.ExecutorEffect = normalizeControlToken(out.ExecutorEffect)
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	out.RuntimeEffect = normalizeControlToken(out.RuntimeEffect)
	if out.Status == "" || out.Status == HostActionNotReady {
		out.Status = HostActionBlocked
	}
	if out.ExecutorEffect == "" {
		out.ExecutorEffect = "none"
	}
	if out.RunnerEffect == "" {
		out.RunnerEffect = "none"
	}
	if out.PromptEffect == "" {
		out.PromptEffect = "none"
	}
	if out.RuntimeEffect == "" {
		out.RuntimeEffect = "none"
	}
	out = hostOwnedDelegationWorkerRuntimeReadinessResetEffects(out)
	out.WorkerResultRequiresVerification = true
	out.WorkerOutputAcceptedAsFact = false
	if out.RawOutputLoaded {
		out.Status = HostActionReviewRequired
		out.ReadyForHostWorkerRuntimeInvocation = false
		out.ReadyForL4HostApprovedWorker = false
		out.ReadyForL5ExplicitWorker = false
		out.HostWorkerRuntimeInvocationAuthorized = false
		out.HostMayInvokeWorkerRuntime = false
		if out.FailureClass == FailureNone {
			out.FailureClass = FailureEvidenceWeak
		}
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		if out.NextHostAction == "" {
			out.NextHostAction = "provide_display_safe_refs"
		}
	}
	if len(out.MissingInputs) > 0 || len(out.BlockedReasons) > 0 || out.RawOutputLoaded {
		out.ReadyForHostWorkerRuntimeInvocation = false
		out.ReadyForL4HostApprovedWorker = false
		out.ReadyForL5ExplicitWorker = false
		out.HostWorkerRuntimeInvocationAuthorized = false
		out.HostMayInvokeWorkerRuntime = false
		if out.Status == HostActionReady {
			out.Status = HostActionBlocked
		}
	}
	return out
}

func CloneHostOwnedDelegationWorkerRuntimeInvocation(in HostOwnedDelegationWorkerRuntimeInvocation) HostOwnedDelegationWorkerRuntimeInvocation {
	out := in
	out.Readiness = in.Readiness.Clone()
	out.ResultReview = in.ResultReview.Clone()
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (i HostOwnedDelegationWorkerRuntimeInvocation) Clone() HostOwnedDelegationWorkerRuntimeInvocation {
	return CloneHostOwnedDelegationWorkerRuntimeInvocation(i)
}

func (i HostOwnedDelegationWorkerRuntimeInvocation) Normalize() HostOwnedDelegationWorkerRuntimeInvocation {
	out := CloneHostOwnedDelegationWorkerRuntimeInvocation(i)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.Readiness = out.Readiness.Normalize()
	out.SubgoalRef = normalizeOneDisplaySafeRef(out.SubgoalRef)
	out.WorkerRef = normalizeOneDisplaySafeRef(out.WorkerRef)
	out.WorkerRunRef = normalizeOneDisplaySafeRef(out.WorkerRunRef)
	out.WorkerRequestRef = normalizeOneDisplaySafeRef(out.WorkerRequestRef)
	out.AdapterRef = normalizeOneDisplaySafeRef(out.AdapterRef)
	out.AdapterVersionRef = normalizeOneDisplaySafeRef(out.AdapterVersionRef)
	out.InvocationRef = normalizeOneDisplaySafeRef(out.InvocationRef)
	out.InvocationReportRef = normalizeOneDisplaySafeRef(out.InvocationReportRef)
	out.ObservedInvocationRef = normalizeOneDisplaySafeRef(out.ObservedInvocationRef)
	out.HostWorkerRuntimeRunRef = normalizeOneDisplaySafeRef(out.HostWorkerRuntimeRunRef)
	out.ObservedWorkerRunRef = normalizeOneDisplaySafeRef(out.ObservedWorkerRunRef)
	out.WorkerResultRef = normalizeOneDisplaySafeRef(out.WorkerResultRef)
	out.WorkerReadbackRef = normalizeOneDisplaySafeRef(out.WorkerReadbackRef)
	out.ObservationRef = normalizeOneDisplaySafeRef(out.ObservationRef)
	out.FailureRef = normalizeOneDisplaySafeRef(out.FailureRef)
	out.CompensationRef = normalizeOneDisplaySafeRef(out.CompensationRef)
	out.ResultReview = out.ResultReview.Normalize()
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.BlockedReasons = normalizeControlTokenList(out.BlockedReasons)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	out.ExecutorEffect = normalizeControlToken(out.ExecutorEffect)
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	out.RuntimeEffect = normalizeControlToken(out.RuntimeEffect)
	if out.Status == "" || out.Status == HostActionNotReady {
		out.Status = HostActionBlocked
	}
	if out.ExecutorEffect == "" {
		out.ExecutorEffect = "none"
	}
	if out.RunnerEffect == "" {
		out.RunnerEffect = "none"
	}
	if out.PromptEffect == "" {
		out.PromptEffect = "none"
	}
	if out.RuntimeEffect == "" {
		out.RuntimeEffect = "none"
	}
	out = hostOwnedDelegationWorkerRuntimeInvocationResetEffects(out)
	out.WorkerResultRequiresVerification = true
	out.WorkerOutputAcceptedAsFact = false
	if out.RawOutputLoaded {
		out.Status = HostActionReviewRequired
		out.ReadyForWorkerResultReview = false
		out.ReadyForFailureReview = false
		if out.FailureClass == FailureNone {
			out.FailureClass = FailureEvidenceWeak
		}
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		if out.NextHostAction == "" {
			out.NextHostAction = "provide_display_safe_refs"
		}
	}
	if len(out.MissingInputs) > 0 || len(out.BlockedReasons) > 0 || out.RawOutputLoaded {
		out.ReadyForWorkerResultReview = false
		out.ReadyForFailureReview = false
		if out.Status == HostActionRecorded {
			out.Status = HostActionBlocked
		}
	}
	return out
}

func hostOwnedDelegationWorkerRuntimeGateReady(gate ProductionAdapterIndependentEffectGate) bool {
	normalized := gate.Normalize()
	return normalized.Kind == ProductionAdapterEffectGateDelegationWorker &&
		normalized.ReadyForIndependentGatePlan &&
		normalized.Status == HostActionReady &&
		len(normalized.MissingInputs) == 0 &&
		len(normalized.BlockedReasons) == 0 &&
		!normalized.RawOutputLoaded
}

func hostOwnedDelegationWorkerRuntimeReadinessBlock(result HostOwnedDelegationWorkerRuntimeReadiness, failure FailureClass, reason string, missing MissingInput, next NextHostAction) HostOwnedDelegationWorkerRuntimeReadiness {
	result.Status = HostActionBlocked
	result.ReadyForHostWorkerRuntimeInvocation = false
	result.ReadyForL4HostApprovedWorker = false
	result.ReadyForL5ExplicitWorker = false
	result.HostWorkerRuntimeInvocationAuthorized = false
	result.HostMayInvokeWorkerRuntime = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = firstNextHostAction(next, result.NextHostAction)
	if result.NextHostAction == "" {
		result.NextHostAction = "review_delegation_worker_runtime_readiness"
	}
	return result
}

func hostOwnedDelegationWorkerRuntimeInvocationBlock(result HostOwnedDelegationWorkerRuntimeInvocation, failure FailureClass, reason string, missing MissingInput, next NextHostAction) HostOwnedDelegationWorkerRuntimeInvocation {
	result.Status = HostActionBlocked
	result.ReadyForWorkerResultReview = false
	result.ReadyForFailureReview = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = firstNextHostAction(next, result.NextHostAction)
	if result.NextHostAction == "" {
		result.NextHostAction = "review_delegation_worker_runtime_invocation"
	}
	return result
}

func hostOwnedDelegationWorkerRuntimeReadinessBoundaries(groups ...[]Boundary) []Boundary {
	base := []Boundary{
		"host_owned_delegation_worker_runtime_gate",
		"delegation_worker_runtime_readiness",
		"delegation_request_required",
		"l4_host_approved_or_l5_explicit_required",
		"host_owned_worker_runtime",
		"display_safe_refs_only",
		"explicit_host_confirmation_required",
		"worker_output_not_fact",
		"worker_result_requires_verification",
		"no_worker_dispatch_by_core",
		"no_delegation_worker_runtime_by_core",
		"no_runner_dispatch",
		"no_runtime_adapter_by_core",
		"no_workflow_dispatch",
		"no_scheduler_apply",
		"no_installer_apply",
		"no_store_mutation_by_core",
		"no_compensation_executor_by_core",
	}
	for _, group := range groups {
		base = AppendBoundaries(base, group...)
	}
	return base
}

func hostOwnedDelegationWorkerRuntimeInvocationBoundaries(groups ...[]Boundary) []Boundary {
	base := []Boundary{
		"host_owned_delegation_worker_runtime_invocation_report",
		"delegation_worker_runtime_invocation_report_only",
		"host_delegation_worker_runtime_reported_only",
		"display_safe_refs_only",
		"worker_output_not_fact",
		"worker_result_requires_verification",
		"no_worker_dispatch_by_core",
		"no_delegation_worker_runtime_by_core",
		"no_runner_dispatch",
		"no_runtime_adapter_by_core",
		"no_workflow_dispatch",
		"no_scheduler_apply",
		"no_installer_apply",
		"no_store_mutation_by_core",
		"no_compensation_executor_by_core",
	}
	for _, group := range groups {
		base = AppendBoundaries(base, group...)
	}
	return base
}

func hostOwnedDelegationWorkerRuntimeReadinessUnsafe(input HostOwnedDelegationWorkerRuntimeReadinessInput, request DelegationRequestProjection, gate ProductionAdapterIndependentEffectGate) bool {
	if input.RawOutputLoaded ||
		request.RawOutputLoaded ||
		gate.RawOutputLoaded {
		return true
	}
	return displaySafeRefRejected(input.AdapterRef) ||
		displaySafeRefRejected(input.AdapterVersionRef) ||
		displaySafeRefRejected(input.AdapterCapabilityRef) ||
		displaySafeRefRejected(input.AdapterContractRef) ||
		displaySafeRefRejected(input.HostConfirmationRef) ||
		displaySafeRefRejected(input.WorkerRunRef) ||
		displaySafeRefRejected(input.WorkerRequestRef) ||
		displaySafeRefRejected(input.InvocationRef) ||
		displaySafeRefRejected(input.ResultBindingRef) ||
		displaySafeRefRejected(input.ReadbackBindingRef) ||
		displaySafeRefRejected(input.IdempotencyRef) ||
		displaySafeRefRejected(input.BudgetRef) ||
		displaySafeRefRejected(input.VerificationRef) ||
		displaySafeRefRejected(input.FailureBindingRef) ||
		displaySafeRefRejected(input.CompensationRef) ||
		evidenceRefRejected(input.EvidenceRefs) ||
		displaySafeRefSliceRejected(input.DecisionBasis)
}

func hostOwnedDelegationWorkerRuntimeInvocationUnsafe(input HostOwnedDelegationWorkerRuntimeInvocationInput, readiness HostOwnedDelegationWorkerRuntimeReadiness) bool {
	if input.RawOutputLoaded || readiness.RawOutputLoaded {
		return true
	}
	return displaySafeRefRejected(input.InvocationReportRef) ||
		displaySafeRefRejected(input.ObservedInvocationRef) ||
		displaySafeRefRejected(input.HostWorkerRuntimeRunRef) ||
		displaySafeRefRejected(input.ObservedWorkerRunRef) ||
		displaySafeRefRejected(input.WorkerResultRef) ||
		displaySafeRefRejected(input.WorkerReadbackRef) ||
		displaySafeRefRejected(input.ObservationRef) ||
		displaySafeRefRejected(input.FailureRef) ||
		displaySafeRefRejected(input.CompensationRef) ||
		evidenceRefRejected(input.EvidenceRefs)
}

func hostOwnedDelegationWorkerRuntimeReadinessResetEffects(in HostOwnedDelegationWorkerRuntimeReadiness) HostOwnedDelegationWorkerRuntimeReadiness {
	in.ExecutorEffect = "none"
	in.RunnerEffect = "none"
	in.PromptEffect = "none"
	in.RuntimeEffect = "none"
	in.CoreWorkerRuntimeInvoked = false
	in.WorkerDispatched = false
	in.RunnerDispatched = false
	in.RuntimeAdapterExecuted = false
	in.WorkflowDispatched = false
	in.SchedulerApplied = false
	in.InstallerExecuted = false
	in.StoreMutationExecuted = false
	in.CompensationExecuted = false
	return in
}

func hostOwnedDelegationWorkerRuntimeInvocationResetEffects(in HostOwnedDelegationWorkerRuntimeInvocation) HostOwnedDelegationWorkerRuntimeInvocation {
	in.ExecutorEffect = "none"
	in.RunnerEffect = "none"
	in.PromptEffect = "none"
	in.RuntimeEffect = "none"
	in.CoreWorkerRuntimeInvoked = false
	in.WorkerDispatched = false
	in.RunnerDispatched = false
	in.RuntimeAdapterExecuted = false
	in.WorkflowDispatched = false
	in.SchedulerApplied = false
	in.InstallerExecuted = false
	in.StoreMutationExecuted = false
	in.CompensationExecuted = false
	return in
}
