package controlcontract

type HostOwnedObjectiveExecutorAdapterReadinessInput struct {
	Request              HostOwnedObjectiveExecutorStepRequest `json:"request,omitempty"`
	AdapterRef           DisplaySafeRef                        `json:"adapter_ref,omitempty"`
	AdapterVersionRef    DisplaySafeRef                        `json:"adapter_version_ref,omitempty"`
	AdapterCapabilityRef DisplaySafeRef                        `json:"adapter_capability_ref,omitempty"`
	AdapterContractRef   DisplaySafeRef                        `json:"adapter_contract_ref,omitempty"`
	HostConfirmationRef  DisplaySafeRef                        `json:"host_confirmation_ref,omitempty"`
	InvocationRef        DisplaySafeRef                        `json:"invocation_ref,omitempty"`
	ResultBindingRef     DisplaySafeRef                        `json:"result_binding_ref,omitempty"`
	ReadbackBindingRef   DisplaySafeRef                        `json:"readback_binding_ref,omitempty"`
	CancellationRef      DisplaySafeRef                        `json:"cancellation_ref,omitempty"`
	ApprovalRefs         []DisplaySafeRef                      `json:"approval_refs,omitempty"`
	PolicyRefs           []DisplaySafeRef                      `json:"policy_refs,omitempty"`
	EvidenceRefs         []EvidenceRef                         `json:"evidence_refs,omitempty"`
	DecisionBasis        []DisplaySafeRef                      `json:"decision_basis,omitempty"`
	Boundaries           []Boundary                            `json:"boundaries,omitempty"`
	RawOutputLoaded      bool                                  `json:"raw_output_loaded"`
}

type HostOwnedObjectiveExecutorAdapterReadiness struct {
	ContractVersion                 string                                `json:"contract_version,omitempty"`
	Projected                       bool                                  `json:"projected"`
	Status                          HostActionStatus                      `json:"status,omitempty"`
	ReadyForHostAdapterInvocation   bool                                  `json:"ready_for_host_adapter_invocation"`
	HostAdapterInvocationAuthorized bool                                  `json:"host_adapter_invocation_authorized"`
	HostMayInvokeAdapter            bool                                  `json:"host_may_invoke_adapter"`
	Request                         HostOwnedObjectiveExecutorStepRequest `json:"request,omitempty"`
	HostExecutorRef                 DisplaySafeRef                        `json:"host_executor_ref,omitempty"`
	ExecutorStepRef                 DisplaySafeRef                        `json:"executor_step_ref,omitempty"`
	ExpectedAttemptRef              AttemptRef                            `json:"expected_attempt_ref,omitempty"`
	ExpectedResultRef               DisplaySafeRef                        `json:"expected_result_ref,omitempty"`
	ExpectedReadbackRef             DisplaySafeRef                        `json:"expected_readback_ref,omitempty"`
	AdapterRef                      DisplaySafeRef                        `json:"adapter_ref,omitempty"`
	AdapterVersionRef               DisplaySafeRef                        `json:"adapter_version_ref,omitempty"`
	AdapterCapabilityRef            DisplaySafeRef                        `json:"adapter_capability_ref,omitempty"`
	AdapterContractRef              DisplaySafeRef                        `json:"adapter_contract_ref,omitempty"`
	HostConfirmationRef             DisplaySafeRef                        `json:"host_confirmation_ref,omitempty"`
	InvocationRef                   DisplaySafeRef                        `json:"invocation_ref,omitempty"`
	ResultBindingRef                DisplaySafeRef                        `json:"result_binding_ref,omitempty"`
	ReadbackBindingRef              DisplaySafeRef                        `json:"readback_binding_ref,omitempty"`
	CancellationRef                 DisplaySafeRef                        `json:"cancellation_ref,omitempty"`
	ApprovalRefs                    []DisplaySafeRef                      `json:"approval_refs,omitempty"`
	PolicyRefs                      []DisplaySafeRef                      `json:"policy_refs,omitempty"`
	EvidenceRefs                    []EvidenceRef                         `json:"evidence_refs,omitempty"`
	FailureClass                    FailureClass                          `json:"failure_class,omitempty"`
	BlockedReasons                  []string                              `json:"blocked_reasons,omitempty"`
	MissingInputs                   []MissingInput                        `json:"missing_inputs,omitempty"`
	DecisionBasis                   []DisplaySafeRef                      `json:"decision_basis,omitempty"`
	Boundaries                      []Boundary                            `json:"boundaries,omitempty"`
	NextHostAction                  NextHostAction                        `json:"next_host_action,omitempty"`
	ExecutorEffect                  string                                `json:"executor_effect,omitempty"`
	RunnerEffect                    string                                `json:"runner_effect,omitempty"`
	PromptEffect                    string                                `json:"prompt_effect,omitempty"`
	RuntimeEffect                   string                                `json:"runtime_effect,omitempty"`
	CoreExecutionExecuted           bool                                  `json:"core_execution_executed"`
	RunnerDispatched                bool                                  `json:"runner_dispatched"`
	RuntimeAdapterExecuted          bool                                  `json:"runtime_adapter_executed"`
	SchedulerApplied                bool                                  `json:"scheduler_applied"`
	InstallerExecuted               bool                                  `json:"installer_executed"`
	WorkflowDispatched              bool                                  `json:"workflow_dispatched"`
	WorkerDispatched                bool                                  `json:"worker_dispatched"`
	StoreMutationExecuted           bool                                  `json:"store_mutation_executed"`
	CompensationExecuted            bool                                  `json:"compensation_executed"`
	RawOutputLoaded                 bool                                  `json:"raw_output_loaded"`
}

type HostOwnedObjectiveExecutorAdapterInvocationInput struct {
	Readiness               HostOwnedObjectiveExecutorAdapterReadiness `json:"readiness,omitempty"`
	InvocationReportRef     DisplaySafeRef                             `json:"invocation_report_ref,omitempty"`
	ObservedInvocationRef   DisplaySafeRef                             `json:"observed_invocation_ref,omitempty"`
	HostAdapterRunRef       DisplaySafeRef                             `json:"host_adapter_run_ref,omitempty"`
	ExecutorStepResultRef   DisplaySafeRef                             `json:"executor_step_result_ref,omitempty"`
	ExecutorStepReadbackRef DisplaySafeRef                             `json:"executor_step_readback_ref,omitempty"`
	AttemptRef              AttemptRef                                 `json:"attempt_ref,omitempty"`
	FailureRef              DisplaySafeRef                             `json:"failure_ref,omitempty"`
	FailureClass            FailureClass                               `json:"failure_class,omitempty"`
	FailureReason           string                                     `json:"failure_reason,omitempty"`
	HostInvocationReported  bool                                       `json:"host_invocation_reported"`
	HostInvocationCompleted bool                                       `json:"host_invocation_completed"`
	HostInvocationFailed    bool                                       `json:"host_invocation_failed"`
	EvidenceRefs            []EvidenceRef                              `json:"evidence_refs,omitempty"`
	Boundaries              []Boundary                                 `json:"boundaries,omitempty"`
	RawOutputLoaded         bool                                       `json:"raw_output_loaded"`
}

type HostOwnedObjectiveExecutorAdapterInvocation struct {
	ContractVersion            string                                     `json:"contract_version,omitempty"`
	Projected                  bool                                       `json:"projected"`
	Status                     HostActionStatus                           `json:"status,omitempty"`
	ReadyForExecutorStepResult bool                                       `json:"ready_for_executor_step_result"`
	ReadyForFailureReview      bool                                       `json:"ready_for_failure_review"`
	HostInvocationReported     bool                                       `json:"host_invocation_reported"`
	HostInvocationCompleted    bool                                       `json:"host_invocation_completed"`
	HostInvocationFailed       bool                                       `json:"host_invocation_failed"`
	Readiness                  HostOwnedObjectiveExecutorAdapterReadiness `json:"readiness,omitempty"`
	HostExecutorRef            DisplaySafeRef                             `json:"host_executor_ref,omitempty"`
	ExecutorStepRef            DisplaySafeRef                             `json:"executor_step_ref,omitempty"`
	ExpectedAttemptRef         AttemptRef                                 `json:"expected_attempt_ref,omitempty"`
	ExpectedResultRef          DisplaySafeRef                             `json:"expected_result_ref,omitempty"`
	ExpectedReadbackRef        DisplaySafeRef                             `json:"expected_readback_ref,omitempty"`
	AdapterRef                 DisplaySafeRef                             `json:"adapter_ref,omitempty"`
	AdapterVersionRef          DisplaySafeRef                             `json:"adapter_version_ref,omitempty"`
	InvocationRef              DisplaySafeRef                             `json:"invocation_ref,omitempty"`
	InvocationReportRef        DisplaySafeRef                             `json:"invocation_report_ref,omitempty"`
	ObservedInvocationRef      DisplaySafeRef                             `json:"observed_invocation_ref,omitempty"`
	HostAdapterRunRef          DisplaySafeRef                             `json:"host_adapter_run_ref,omitempty"`
	ExecutorStepResultRef      DisplaySafeRef                             `json:"executor_step_result_ref,omitempty"`
	ExecutorStepReadbackRef    DisplaySafeRef                             `json:"executor_step_readback_ref,omitempty"`
	AttemptRef                 AttemptRef                                 `json:"attempt_ref,omitempty"`
	FailureRef                 DisplaySafeRef                             `json:"failure_ref,omitempty"`
	EvidenceRefs               []EvidenceRef                              `json:"evidence_refs,omitempty"`
	FailureClass               FailureClass                               `json:"failure_class,omitempty"`
	FailureReason              string                                     `json:"failure_reason,omitempty"`
	BlockedReasons             []string                                   `json:"blocked_reasons,omitempty"`
	MissingInputs              []MissingInput                             `json:"missing_inputs,omitempty"`
	Boundaries                 []Boundary                                 `json:"boundaries,omitempty"`
	NextHostAction             NextHostAction                             `json:"next_host_action,omitempty"`
	ExecutorEffect             string                                     `json:"executor_effect,omitempty"`
	RunnerEffect               string                                     `json:"runner_effect,omitempty"`
	PromptEffect               string                                     `json:"prompt_effect,omitempty"`
	RuntimeEffect              string                                     `json:"runtime_effect,omitempty"`
	CoreExecutionExecuted      bool                                       `json:"core_execution_executed"`
	RunnerDispatched           bool                                       `json:"runner_dispatched"`
	RuntimeAdapterExecuted     bool                                       `json:"runtime_adapter_executed"`
	SchedulerApplied           bool                                       `json:"scheduler_applied"`
	InstallerExecuted          bool                                       `json:"installer_executed"`
	WorkflowDispatched         bool                                       `json:"workflow_dispatched"`
	WorkerDispatched           bool                                       `json:"worker_dispatched"`
	StoreMutationExecuted      bool                                       `json:"store_mutation_executed"`
	CompensationExecuted       bool                                       `json:"compensation_executed"`
	RawOutputLoaded            bool                                       `json:"raw_output_loaded"`
}

func BuildHostOwnedObjectiveExecutorAdapterReadiness(input HostOwnedObjectiveExecutorAdapterReadinessInput) HostOwnedObjectiveExecutorAdapterReadiness {
	request := input.Request.Normalize()
	result := HostOwnedObjectiveExecutorAdapterReadiness{
		ContractVersion:      ContractVersion,
		Projected:            true,
		Status:               HostActionBlocked,
		Request:              request,
		HostExecutorRef:      request.HostExecutorRef,
		ExecutorStepRef:      request.ExecutorStepRef,
		ExpectedAttemptRef:   request.ExpectedAttemptRef,
		ExpectedResultRef:    request.ExpectedResultRef,
		ExpectedReadbackRef:  request.ExpectedReadbackRef,
		AdapterRef:           normalizeOneDisplaySafeRef(input.AdapterRef),
		AdapterVersionRef:    normalizeOneDisplaySafeRef(input.AdapterVersionRef),
		AdapterCapabilityRef: normalizeOneDisplaySafeRef(input.AdapterCapabilityRef),
		AdapterContractRef:   normalizeOneDisplaySafeRef(input.AdapterContractRef),
		HostConfirmationRef:  normalizeOneDisplaySafeRef(input.HostConfirmationRef),
		InvocationRef:        normalizeOneDisplaySafeRef(input.InvocationRef),
		ResultBindingRef:     normalizeOneDisplaySafeRef(input.ResultBindingRef),
		ReadbackBindingRef:   normalizeOneDisplaySafeRef(input.ReadbackBindingRef),
		CancellationRef:      normalizeOneDisplaySafeRef(input.CancellationRef),
		ApprovalRefs:         normalizeDisplaySafeRefs(append(cloneDisplaySafeRefs(request.ApprovalRefs), input.ApprovalRefs...)),
		PolicyRefs:           normalizeDisplaySafeRefs(append(cloneDisplaySafeRefs(request.PolicyRefs), input.PolicyRefs...)),
		EvidenceRefs:         MergeEvidenceRefs(input.EvidenceRefs, request.EvidenceRefs),
		FailureClass:         FailureNone,
		DecisionBasis: normalizeDisplaySafeRefs(append(
			[]DisplaySafeRef{
				"objective_executor:host_owned",
				"objective_executor:adapter_readiness",
			},
			input.DecisionBasis...,
		)),
		Boundaries:     hostOwnedObjectiveExecutorAdapterReadinessBoundaries(request.Boundaries, input.Boundaries),
		NextHostAction: "provide_objective_executor_adapter_binding",
		ExecutorEffect: "none",
		RunnerEffect:   "none",
		PromptEffect:   "none",
		RuntimeEffect:  "none",
		RawOutputLoaded: input.RawOutputLoaded ||
			request.RawOutputLoaded,
	}
	if hostOwnedObjectiveExecutorAdapterReadinessUnsafe(input, request) {
		result.RawOutputLoaded = true
		result = hostOwnedObjectiveExecutorAdapterReadinessBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if !request.ReadyForHostExecution {
		result = hostOwnedObjectiveExecutorAdapterReadinessBlock(result, firstFailureClass(request.FailureClass, FailureConfigMissing), "executor_step_request_not_ready", "host:objective_executor_step_request", firstNextHostAction(request.NextHostAction, "review_objective_executor_step_request"))
	}
	for _, check := range []struct {
		ok      bool
		reason  string
		missing MissingInput
		next    NextHostAction
	}{
		{result.AdapterRef != "", "adapter_ref_missing", "host:objective_executor_adapter_ref", "provide_objective_executor_adapter_binding"},
		{result.AdapterVersionRef != "", "adapter_version_ref_missing", "host:objective_executor_adapter_version_ref", "provide_objective_executor_adapter_binding"},
		{result.AdapterCapabilityRef != "", "adapter_capability_ref_missing", "host:objective_executor_adapter_capability_ref", "provide_objective_executor_adapter_capability"},
		{result.AdapterContractRef != "", "adapter_contract_ref_missing", "contract:objective_executor_adapter", "provide_objective_executor_adapter_contract"},
		{result.HostConfirmationRef != "", "host_confirmation_ref_missing", "host:objective_executor_adapter_confirmation", "request_host_adapter_confirmation"},
		{result.InvocationRef != "", "invocation_ref_missing", "host:objective_executor_adapter_invocation_ref", "provide_objective_executor_adapter_invocation_ref"},
		{result.ResultBindingRef != "", "result_binding_ref_missing", "host:objective_executor_result_binding", "provide_objective_executor_result_binding"},
		{result.ReadbackBindingRef != "", "readback_binding_ref_missing", "host:objective_executor_readback_binding", "provide_objective_executor_readback_binding"},
		{result.CancellationRef != "", "cancellation_ref_missing", "host:objective_executor_cancellation_ref", "provide_objective_executor_cancellation_ref"},
	} {
		if !check.ok {
			result = hostOwnedObjectiveExecutorAdapterReadinessBlock(result, FailureConfigMissing, check.reason, check.missing, check.next)
		}
	}
	if result.ResultBindingRef != "" && result.ExpectedResultRef != "" && result.ResultBindingRef != result.ExpectedResultRef {
		result = hostOwnedObjectiveExecutorAdapterReadinessBlock(result, FailureVerificationFailed, "result_binding_ref_mismatch", "host:objective_executor_result_binding", "review_objective_executor_adapter_binding")
	}
	if result.ReadbackBindingRef != "" && result.ExpectedReadbackRef != "" && result.ReadbackBindingRef != result.ExpectedReadbackRef {
		result = hostOwnedObjectiveExecutorAdapterReadinessBlock(result, FailureVerificationFailed, "readback_binding_ref_mismatch", "host:objective_executor_readback_binding", "review_objective_executor_adapter_binding")
	}
	if len(result.MissingInputs) == 0 && len(result.BlockedReasons) == 0 {
		result.Status = HostActionReady
		result.ReadyForHostAdapterInvocation = true
		result.HostAdapterInvocationAuthorized = true
		result.HostMayInvokeAdapter = true
		result.NextHostAction = "host_may_invoke_objective_executor_adapter"
		result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_host_owned_objective_executor_adapter_invocation", "host_may_invoke_objective_executor_adapter")
	}
	return result.Normalize()
}

func BuildHostOwnedObjectiveExecutorAdapterInvocation(input HostOwnedObjectiveExecutorAdapterInvocationInput) HostOwnedObjectiveExecutorAdapterInvocation {
	readiness := input.Readiness.Normalize()
	result := HostOwnedObjectiveExecutorAdapterInvocation{
		ContractVersion:         ContractVersion,
		Projected:               true,
		Status:                  HostActionBlocked,
		Readiness:               readiness,
		HostExecutorRef:         readiness.HostExecutorRef,
		ExecutorStepRef:         readiness.ExecutorStepRef,
		ExpectedAttemptRef:      readiness.ExpectedAttemptRef,
		ExpectedResultRef:       readiness.ExpectedResultRef,
		ExpectedReadbackRef:     readiness.ExpectedReadbackRef,
		AdapterRef:              readiness.AdapterRef,
		AdapterVersionRef:       readiness.AdapterVersionRef,
		InvocationRef:           readiness.InvocationRef,
		InvocationReportRef:     normalizeOneDisplaySafeRef(input.InvocationReportRef),
		ObservedInvocationRef:   normalizeOneDisplaySafeRef(input.ObservedInvocationRef),
		HostAdapterRunRef:       normalizeOneDisplaySafeRef(input.HostAdapterRunRef),
		ExecutorStepResultRef:   normalizeOneDisplaySafeRef(input.ExecutorStepResultRef),
		ExecutorStepReadbackRef: normalizeOneDisplaySafeRef(input.ExecutorStepReadbackRef),
		AttemptRef:              normalizeOneAttemptRef(input.AttemptRef),
		FailureRef:              normalizeOneDisplaySafeRef(input.FailureRef),
		EvidenceRefs:            MergeEvidenceRefs(input.EvidenceRefs, readiness.EvidenceRefs),
		FailureClass:            firstFailureClass(input.FailureClass, FailureNone),
		FailureReason:           firstNonEmptyContractString(input.FailureReason),
		HostInvocationReported:  input.HostInvocationReported,
		HostInvocationCompleted: input.HostInvocationCompleted,
		HostInvocationFailed:    input.HostInvocationFailed,
		Boundaries:              hostOwnedObjectiveExecutorAdapterInvocationBoundaries(readiness.Boundaries, input.Boundaries),
		NextHostAction:          "provide_objective_executor_adapter_invocation_report",
		ExecutorEffect:          "none",
		RunnerEffect:            "none",
		PromptEffect:            "none",
		RuntimeEffect:           "none",
		RawOutputLoaded:         input.RawOutputLoaded || readiness.RawOutputLoaded,
	}
	if hostOwnedObjectiveExecutorAdapterInvocationUnsafe(input, readiness) {
		result.RawOutputLoaded = true
		result = hostOwnedObjectiveExecutorAdapterInvocationBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if !readiness.ReadyForHostAdapterInvocation {
		result = hostOwnedObjectiveExecutorAdapterInvocationBlock(result, firstFailureClass(readiness.FailureClass, FailureConfigMissing), "adapter_readiness_not_ready", "host:objective_executor_adapter_readiness", firstNextHostAction(readiness.NextHostAction, "review_objective_executor_adapter_readiness"))
	}
	if result.InvocationReportRef == "" {
		result = hostOwnedObjectiveExecutorAdapterInvocationBlock(result, FailureEvidenceMissing, "invocation_report_ref_missing", "host:objective_executor_adapter_invocation_report_ref", "provide_objective_executor_adapter_invocation_report")
	}
	if result.ObservedInvocationRef == "" {
		result = hostOwnedObjectiveExecutorAdapterInvocationBlock(result, FailureEvidenceMissing, "observed_invocation_ref_missing", "host:objective_executor_adapter_invocation_ref", "provide_objective_executor_adapter_invocation_report")
	} else if result.InvocationRef != "" && result.ObservedInvocationRef != result.InvocationRef {
		result = hostOwnedObjectiveExecutorAdapterInvocationBlock(result, FailureVerificationFailed, "observed_invocation_ref_mismatch", "host:objective_executor_adapter_invocation_ref", "review_objective_executor_adapter_invocation_report")
	}
	if result.HostAdapterRunRef == "" {
		result = hostOwnedObjectiveExecutorAdapterInvocationBlock(result, FailureEvidenceMissing, "host_adapter_run_ref_missing", "host:objective_executor_adapter_run_ref", "provide_objective_executor_adapter_invocation_report")
	}
	if !result.HostInvocationReported {
		result = hostOwnedObjectiveExecutorAdapterInvocationBlock(result, FailureEvidenceMissing, "host_invocation_not_reported", "host:objective_executor_adapter_invocation_report", "provide_objective_executor_adapter_invocation_report")
	}
	if !result.HostInvocationCompleted && !result.HostInvocationFailed {
		result = hostOwnedObjectiveExecutorAdapterInvocationBlock(result, FailureEvidenceMissing, "host_invocation_status_missing", "host:objective_executor_adapter_invocation_status", "provide_objective_executor_adapter_invocation_report")
	}
	if result.HostInvocationCompleted && result.HostInvocationFailed {
		result = hostOwnedObjectiveExecutorAdapterInvocationBlock(result, FailureInvalidInput, "host_invocation_status_conflict", "host:objective_executor_adapter_invocation_status", "review_objective_executor_adapter_invocation_report")
	}
	if result.AttemptRef == "" {
		result = hostOwnedObjectiveExecutorAdapterInvocationBlock(result, FailureEvidenceMissing, "attempt_ref_missing", "host:attempt_ref", "provide_objective_executor_adapter_invocation_report")
	} else if result.ExpectedAttemptRef != "" && result.AttemptRef != result.ExpectedAttemptRef {
		result = hostOwnedObjectiveExecutorAdapterInvocationBlock(result, FailureVerificationFailed, "attempt_ref_mismatch", "host:attempt_ref", "review_objective_executor_adapter_invocation_report")
	}
	if result.ExecutorStepResultRef == "" {
		result = hostOwnedObjectiveExecutorAdapterInvocationBlock(result, FailureEvidenceMissing, "executor_step_result_ref_missing", "host:objective_executor_step_result_ref", "provide_objective_executor_adapter_invocation_report")
	} else if result.ExpectedResultRef != "" && result.ExecutorStepResultRef != result.ExpectedResultRef {
		result = hostOwnedObjectiveExecutorAdapterInvocationBlock(result, FailureVerificationFailed, "executor_step_result_ref_mismatch", "host:objective_executor_step_result_ref", "review_objective_executor_adapter_invocation_report")
	}
	if result.ExecutorStepReadbackRef == "" {
		result = hostOwnedObjectiveExecutorAdapterInvocationBlock(result, FailureEvidenceMissing, "executor_step_readback_ref_missing", "host:objective_executor_step_readback_ref", "provide_objective_executor_adapter_invocation_report")
	} else if result.ExpectedReadbackRef != "" && result.ExecutorStepReadbackRef != result.ExpectedReadbackRef {
		result = hostOwnedObjectiveExecutorAdapterInvocationBlock(result, FailureVerificationFailed, "executor_step_readback_ref_mismatch", "host:objective_executor_step_readback_ref", "review_objective_executor_adapter_invocation_report")
	}
	if result.HostInvocationFailed && result.FailureRef == "" {
		result = hostOwnedObjectiveExecutorAdapterInvocationBlock(result, FailureEvidenceMissing, "failure_ref_missing", "host:objective_executor_adapter_failure_ref", "provide_objective_executor_adapter_failure_ref")
	}
	if len(result.MissingInputs) == 0 && len(result.BlockedReasons) == 0 {
		result.Status = HostActionRecorded
		if result.HostInvocationFailed {
			result.ReadyForFailureReview = true
			result.FailureClass = firstFailureClass(result.FailureClass, FailureVerificationFailed)
			result.NextHostAction = "review_objective_executor_adapter_failure"
			result.Boundaries = AppendBoundaries(result.Boundaries, "host_owned_objective_executor_adapter_failure_recorded")
		} else {
			result.ReadyForExecutorStepResult = true
			result.NextHostAction = "build_objective_executor_step_result"
			result.Boundaries = AppendBoundaries(result.Boundaries, "host_owned_objective_executor_adapter_invocation_recorded")
		}
	}
	return result.Normalize()
}

func CloneHostOwnedObjectiveExecutorAdapterReadiness(in HostOwnedObjectiveExecutorAdapterReadiness) HostOwnedObjectiveExecutorAdapterReadiness {
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

func (r HostOwnedObjectiveExecutorAdapterReadiness) Clone() HostOwnedObjectiveExecutorAdapterReadiness {
	return CloneHostOwnedObjectiveExecutorAdapterReadiness(r)
}

func (r HostOwnedObjectiveExecutorAdapterReadiness) Normalize() HostOwnedObjectiveExecutorAdapterReadiness {
	out := CloneHostOwnedObjectiveExecutorAdapterReadiness(r)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.Request = out.Request.Normalize()
	out.HostExecutorRef = normalizeOneDisplaySafeRef(out.HostExecutorRef)
	out.ExecutorStepRef = normalizeOneDisplaySafeRef(out.ExecutorStepRef)
	out.ExpectedAttemptRef = normalizeOneAttemptRef(out.ExpectedAttemptRef)
	out.ExpectedResultRef = normalizeOneDisplaySafeRef(out.ExpectedResultRef)
	out.ExpectedReadbackRef = normalizeOneDisplaySafeRef(out.ExpectedReadbackRef)
	out.AdapterRef = normalizeOneDisplaySafeRef(out.AdapterRef)
	out.AdapterVersionRef = normalizeOneDisplaySafeRef(out.AdapterVersionRef)
	out.AdapterCapabilityRef = normalizeOneDisplaySafeRef(out.AdapterCapabilityRef)
	out.AdapterContractRef = normalizeOneDisplaySafeRef(out.AdapterContractRef)
	out.HostConfirmationRef = normalizeOneDisplaySafeRef(out.HostConfirmationRef)
	out.InvocationRef = normalizeOneDisplaySafeRef(out.InvocationRef)
	out.ResultBindingRef = normalizeOneDisplaySafeRef(out.ResultBindingRef)
	out.ReadbackBindingRef = normalizeOneDisplaySafeRef(out.ReadbackBindingRef)
	out.CancellationRef = normalizeOneDisplaySafeRef(out.CancellationRef)
	out.ApprovalRefs = normalizeDisplaySafeRefs(out.ApprovalRefs)
	out.PolicyRefs = normalizeDisplaySafeRefs(out.PolicyRefs)
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
	out = hostOwnedObjectiveExecutorAdapterNormalizeReadinessEffects(out)
	out.ReadyForHostAdapterInvocation = out.Status == HostActionReady && len(out.MissingInputs) == 0 && len(out.BlockedReasons) == 0 && !out.RawOutputLoaded
	out.HostAdapterInvocationAuthorized = out.ReadyForHostAdapterInvocation
	out.HostMayInvokeAdapter = out.ReadyForHostAdapterInvocation
	if out.RawOutputLoaded {
		out.Status = HostActionReviewRequired
		out.ReadyForHostAdapterInvocation = false
		out.HostAdapterInvocationAuthorized = false
		out.HostMayInvokeAdapter = false
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

func CloneHostOwnedObjectiveExecutorAdapterInvocation(in HostOwnedObjectiveExecutorAdapterInvocation) HostOwnedObjectiveExecutorAdapterInvocation {
	out := in
	out.Readiness = in.Readiness.Clone()
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (r HostOwnedObjectiveExecutorAdapterInvocation) Clone() HostOwnedObjectiveExecutorAdapterInvocation {
	return CloneHostOwnedObjectiveExecutorAdapterInvocation(r)
}

func (r HostOwnedObjectiveExecutorAdapterInvocation) Normalize() HostOwnedObjectiveExecutorAdapterInvocation {
	out := CloneHostOwnedObjectiveExecutorAdapterInvocation(r)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.Readiness = out.Readiness.Normalize()
	out.HostExecutorRef = normalizeOneDisplaySafeRef(out.HostExecutorRef)
	out.ExecutorStepRef = normalizeOneDisplaySafeRef(out.ExecutorStepRef)
	out.ExpectedAttemptRef = normalizeOneAttemptRef(out.ExpectedAttemptRef)
	out.ExpectedResultRef = normalizeOneDisplaySafeRef(out.ExpectedResultRef)
	out.ExpectedReadbackRef = normalizeOneDisplaySafeRef(out.ExpectedReadbackRef)
	out.AdapterRef = normalizeOneDisplaySafeRef(out.AdapterRef)
	out.AdapterVersionRef = normalizeOneDisplaySafeRef(out.AdapterVersionRef)
	out.InvocationRef = normalizeOneDisplaySafeRef(out.InvocationRef)
	out.InvocationReportRef = normalizeOneDisplaySafeRef(out.InvocationReportRef)
	out.ObservedInvocationRef = normalizeOneDisplaySafeRef(out.ObservedInvocationRef)
	out.HostAdapterRunRef = normalizeOneDisplaySafeRef(out.HostAdapterRunRef)
	out.ExecutorStepResultRef = normalizeOneDisplaySafeRef(out.ExecutorStepResultRef)
	out.ExecutorStepReadbackRef = normalizeOneDisplaySafeRef(out.ExecutorStepReadbackRef)
	out.AttemptRef = normalizeOneAttemptRef(out.AttemptRef)
	out.FailureRef = normalizeOneDisplaySafeRef(out.FailureRef)
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.FailureReason = firstNonEmptyContractString(out.FailureReason)
	out.BlockedReasons = normalizeControlTokenList(out.BlockedReasons)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	out.ExecutorEffect = normalizeControlToken(out.ExecutorEffect)
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	out.RuntimeEffect = normalizeControlToken(out.RuntimeEffect)
	out = hostOwnedObjectiveExecutorAdapterNormalizeInvocationEffects(out)
	out.ReadyForExecutorStepResult = out.Status == HostActionRecorded && out.HostInvocationCompleted && !out.HostInvocationFailed && len(out.MissingInputs) == 0 && len(out.BlockedReasons) == 0 && !out.RawOutputLoaded
	out.ReadyForFailureReview = out.Status == HostActionRecorded && out.HostInvocationFailed && len(out.MissingInputs) == 0 && len(out.BlockedReasons) == 0 && !out.RawOutputLoaded
	if out.RawOutputLoaded {
		out.Status = HostActionReviewRequired
		out.ReadyForExecutorStepResult = false
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

func hostOwnedObjectiveExecutorAdapterReadinessBlock(result HostOwnedObjectiveExecutorAdapterReadiness, failure FailureClass, reason string, missing MissingInput, next NextHostAction) HostOwnedObjectiveExecutorAdapterReadiness {
	result.Status = HostActionBlocked
	result.ReadyForHostAdapterInvocation = false
	result.HostAdapterInvocationAuthorized = false
	result.HostMayInvokeAdapter = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	return result
}

func hostOwnedObjectiveExecutorAdapterInvocationBlock(result HostOwnedObjectiveExecutorAdapterInvocation, failure FailureClass, reason string, missing MissingInput, next NextHostAction) HostOwnedObjectiveExecutorAdapterInvocation {
	result.Status = HostActionBlocked
	result.ReadyForExecutorStepResult = false
	result.ReadyForFailureReview = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	return result
}

func hostOwnedObjectiveExecutorAdapterReadinessUnsafe(input HostOwnedObjectiveExecutorAdapterReadinessInput, request HostOwnedObjectiveExecutorStepRequest) bool {
	return input.RawOutputLoaded ||
		request.RawOutputLoaded ||
		displaySafeRefRejected(input.AdapterRef) ||
		displaySafeRefRejected(input.AdapterVersionRef) ||
		displaySafeRefRejected(input.AdapterCapabilityRef) ||
		displaySafeRefRejected(input.AdapterContractRef) ||
		displaySafeRefRejected(input.HostConfirmationRef) ||
		displaySafeRefRejected(input.InvocationRef) ||
		displaySafeRefRejected(input.ResultBindingRef) ||
		displaySafeRefRejected(input.ReadbackBindingRef) ||
		displaySafeRefRejected(input.CancellationRef) ||
		displaySafeRefSliceRejected(input.ApprovalRefs) ||
		displaySafeRefSliceRejected(input.PolicyRefs) ||
		displaySafeRefSliceRejected(input.DecisionBasis) ||
		evidenceRefRejected(input.EvidenceRefs)
}

func hostOwnedObjectiveExecutorAdapterInvocationUnsafe(input HostOwnedObjectiveExecutorAdapterInvocationInput, readiness HostOwnedObjectiveExecutorAdapterReadiness) bool {
	return input.RawOutputLoaded ||
		readiness.RawOutputLoaded ||
		displaySafeRefRejected(input.InvocationReportRef) ||
		displaySafeRefRejected(input.ObservedInvocationRef) ||
		displaySafeRefRejected(input.HostAdapterRunRef) ||
		displaySafeRefRejected(input.ExecutorStepResultRef) ||
		displaySafeRefRejected(input.ExecutorStepReadbackRef) ||
		displaySafeRefRejected(DisplaySafeRef(input.AttemptRef)) ||
		displaySafeRefRejected(input.FailureRef) ||
		evidenceRefRejected(input.EvidenceRefs)
}

func hostOwnedObjectiveExecutorAdapterReadinessBoundaries(groups ...[]Boundary) []Boundary {
	all := append([][]Boundary{{
		"host_owned_objective_executor_adapter_gate",
		"host_owned_objective_executor_adapter_readiness",
		"objective_executor_adapter_invocation_gate",
		"explicit_host_confirmation_required",
		"adapter_invocation_not_executed_by_core",
		"display_safe_refs_only",
		"no_core_execution",
		"no_runner_dispatch",
		"no_runtime_adapter_execution",
		"no_tool_execution",
		"no_workflow_dispatch",
		"no_scheduler_apply",
		"no_install_apply",
		"no_worker_dispatch",
		"no_store_mutation_by_core",
		"no_compensation_execution",
		"projection_only",
	}}, groups...)
	return MergeBoundaries(all...)
}

func hostOwnedObjectiveExecutorAdapterInvocationBoundaries(groups ...[]Boundary) []Boundary {
	all := append([][]Boundary{{
		"host_owned_objective_executor_adapter_gate",
		"host_owned_objective_executor_adapter_invocation_report",
		"host_adapter_invocation_report_only",
		"host_invocation_report_not_objective_completion",
		"objective_completion_requires_verification_gate",
		"display_safe_refs_only",
		"no_core_execution",
		"no_runner_dispatch",
		"no_runtime_adapter_execution",
		"no_tool_execution",
		"no_workflow_dispatch",
		"no_scheduler_apply",
		"no_install_apply",
		"no_worker_dispatch",
		"no_store_mutation_by_core",
		"no_compensation_execution",
		"projection_only",
	}}, groups...)
	return MergeBoundaries(all...)
}

func hostOwnedObjectiveExecutorAdapterNormalizeReadinessEffects(value HostOwnedObjectiveExecutorAdapterReadiness) HostOwnedObjectiveExecutorAdapterReadiness {
	if value.ExecutorEffect == "" {
		value.ExecutorEffect = "none"
	}
	if value.RunnerEffect == "" {
		value.RunnerEffect = "none"
	}
	if value.PromptEffect == "" {
		value.PromptEffect = "none"
	}
	if value.RuntimeEffect == "" {
		value.RuntimeEffect = "none"
	}
	value.CoreExecutionExecuted = false
	value.RunnerDispatched = false
	value.RuntimeAdapterExecuted = false
	value.SchedulerApplied = false
	value.InstallerExecuted = false
	value.WorkflowDispatched = false
	value.WorkerDispatched = false
	value.StoreMutationExecuted = false
	value.CompensationExecuted = false
	return value
}

func hostOwnedObjectiveExecutorAdapterNormalizeInvocationEffects(value HostOwnedObjectiveExecutorAdapterInvocation) HostOwnedObjectiveExecutorAdapterInvocation {
	if value.ExecutorEffect == "" {
		value.ExecutorEffect = "none"
	}
	if value.RunnerEffect == "" {
		value.RunnerEffect = "none"
	}
	if value.PromptEffect == "" {
		value.PromptEffect = "none"
	}
	if value.RuntimeEffect == "" {
		value.RuntimeEffect = "none"
	}
	value.CoreExecutionExecuted = false
	value.RunnerDispatched = false
	value.RuntimeAdapterExecuted = false
	value.SchedulerApplied = false
	value.InstallerExecuted = false
	value.WorkflowDispatched = false
	value.WorkerDispatched = false
	value.StoreMutationExecuted = false
	value.CompensationExecuted = false
	return value
}
