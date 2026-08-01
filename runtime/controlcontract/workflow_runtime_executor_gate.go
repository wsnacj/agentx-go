package controlcontract

type HostOwnedWorkflowRuntimeExecutorReadinessInput struct {
	SourceProjection     ReplannerSourceProjection              `json:"source_projection,omitempty"`
	Verification         ObjectiveVerificationGateResult        `json:"verification,omitempty"`
	ReplannerDecision    ObjectiveReplannerDecision             `json:"replanner_decision,omitempty"`
	WorkflowRetryGate    ProductionAdapterIndependentEffectGate `json:"workflow_retry_gate,omitempty"`
	RuntimeExecutorGate  ProductionAdapterIndependentEffectGate `json:"runtime_executor_gate,omitempty"`
	AdapterRef           DisplaySafeRef                         `json:"adapter_ref,omitempty"`
	AdapterVersionRef    DisplaySafeRef                         `json:"adapter_version_ref,omitempty"`
	AdapterCapabilityRef DisplaySafeRef                         `json:"adapter_capability_ref,omitempty"`
	AdapterContractRef   DisplaySafeRef                         `json:"adapter_contract_ref,omitempty"`
	HostConfirmationRef  DisplaySafeRef                         `json:"host_confirmation_ref,omitempty"`
	WorkflowRunRef       DisplaySafeRef                         `json:"workflow_run_ref,omitempty"`
	RetryRequestRef      DisplaySafeRef                         `json:"retry_request_ref,omitempty"`
	InvocationRef        DisplaySafeRef                         `json:"invocation_ref,omitempty"`
	ResultBindingRef     DisplaySafeRef                         `json:"result_binding_ref,omitempty"`
	ReadbackBindingRef   DisplaySafeRef                         `json:"readback_binding_ref,omitempty"`
	IdempotencyRef       DisplaySafeRef                         `json:"idempotency_ref,omitempty"`
	BudgetRef            DisplaySafeRef                         `json:"budget_ref,omitempty"`
	RetryPolicyRef       DisplaySafeRef                         `json:"retry_policy_ref,omitempty"`
	FallbackPolicyRef    DisplaySafeRef                         `json:"fallback_policy_ref,omitempty"`
	FailureBindingRef    DisplaySafeRef                         `json:"failure_binding_ref,omitempty"`
	CompensationRef      DisplaySafeRef                         `json:"compensation_ref,omitempty"`
	EvidenceRefs         []EvidenceRef                          `json:"evidence_refs,omitempty"`
	DecisionBasis        []DisplaySafeRef                       `json:"decision_basis,omitempty"`
	Boundaries           []Boundary                             `json:"boundaries,omitempty"`
	RawOutputLoaded      bool                                   `json:"raw_output_loaded"`
}

type HostOwnedWorkflowRuntimeExecutorReadiness struct {
	ContractVersion                         string                                 `json:"contract_version,omitempty"`
	Projected                               bool                                   `json:"projected"`
	Status                                  HostActionStatus                       `json:"status,omitempty"`
	ReadyForHostWorkflowRuntimeInvocation   bool                                   `json:"ready_for_host_workflow_runtime_invocation"`
	ReadyForSameStrategyRetry               bool                                   `json:"ready_for_same_strategy_retry"`
	ReadyForL3Fallback                      bool                                   `json:"ready_for_l3_fallback"`
	HostRuntimeExecutorInvocationAuthorized bool                                   `json:"host_runtime_executor_invocation_authorized"`
	HostMayInvokeRuntimeExecutor            bool                                   `json:"host_may_invoke_runtime_executor"`
	SourceProjection                        ReplannerSourceProjection              `json:"source_projection,omitempty"`
	Verification                            ObjectiveVerificationGateResult        `json:"verification,omitempty"`
	ReplannerDecision                       ObjectiveReplannerDecision             `json:"replanner_decision,omitempty"`
	WorkflowRetryGate                       ProductionAdapterIndependentEffectGate `json:"workflow_retry_gate,omitempty"`
	RuntimeExecutorGate                     ProductionAdapterIndependentEffectGate `json:"runtime_executor_gate,omitempty"`
	SourceRef                               DisplaySafeRef                         `json:"source_ref,omitempty"`
	WorkflowRunRef                          DisplaySafeRef                         `json:"workflow_run_ref,omitempty"`
	RetryRequestRef                         DisplaySafeRef                         `json:"retry_request_ref,omitempty"`
	CurrentStrategyRef                      DisplaySafeRef                         `json:"current_strategy_ref,omitempty"`
	NextStrategyRef                         DisplaySafeRef                         `json:"next_strategy_ref,omitempty"`
	AdapterRef                              DisplaySafeRef                         `json:"adapter_ref,omitempty"`
	AdapterVersionRef                       DisplaySafeRef                         `json:"adapter_version_ref,omitempty"`
	AdapterCapabilityRef                    DisplaySafeRef                         `json:"adapter_capability_ref,omitempty"`
	AdapterContractRef                      DisplaySafeRef                         `json:"adapter_contract_ref,omitempty"`
	HostConfirmationRef                     DisplaySafeRef                         `json:"host_confirmation_ref,omitempty"`
	InvocationRef                           DisplaySafeRef                         `json:"invocation_ref,omitempty"`
	ResultBindingRef                        DisplaySafeRef                         `json:"result_binding_ref,omitempty"`
	ReadbackBindingRef                      DisplaySafeRef                         `json:"readback_binding_ref,omitempty"`
	IdempotencyRef                          DisplaySafeRef                         `json:"idempotency_ref,omitempty"`
	BudgetRef                               DisplaySafeRef                         `json:"budget_ref,omitempty"`
	RetryPolicyRef                          DisplaySafeRef                         `json:"retry_policy_ref,omitempty"`
	FallbackPolicyRef                       DisplaySafeRef                         `json:"fallback_policy_ref,omitempty"`
	FailureBindingRef                       DisplaySafeRef                         `json:"failure_binding_ref,omitempty"`
	CompensationRef                         DisplaySafeRef                         `json:"compensation_ref,omitempty"`
	EvidenceRefs                            []EvidenceRef                          `json:"evidence_refs,omitempty"`
	FailureClass                            FailureClass                           `json:"failure_class,omitempty"`
	BlockedReasons                          []string                               `json:"blocked_reasons,omitempty"`
	MissingInputs                           []MissingInput                         `json:"missing_inputs,omitempty"`
	DecisionBasis                           []DisplaySafeRef                       `json:"decision_basis,omitempty"`
	Boundaries                              []Boundary                             `json:"boundaries,omitempty"`
	NextHostAction                          NextHostAction                         `json:"next_host_action,omitempty"`
	ExecutorEffect                          string                                 `json:"executor_effect,omitempty"`
	RunnerEffect                            string                                 `json:"runner_effect,omitempty"`
	PromptEffect                            string                                 `json:"prompt_effect,omitempty"`
	RuntimeEffect                           string                                 `json:"runtime_effect,omitempty"`
	CoreWorkflowRetryApplied                bool                                   `json:"core_workflow_retry_applied"`
	CoreRuntimeExecutorInvoked              bool                                   `json:"core_runtime_executor_invoked"`
	RunnerDispatched                        bool                                   `json:"runner_dispatched"`
	RuntimeAdapterExecuted                  bool                                   `json:"runtime_adapter_executed"`
	WorkflowDispatched                      bool                                   `json:"workflow_dispatched"`
	WorkerDispatched                        bool                                   `json:"worker_dispatched"`
	StoreMutationExecuted                   bool                                   `json:"store_mutation_executed"`
	CompensationExecuted                    bool                                   `json:"compensation_executed"`
	RawOutputLoaded                         bool                                   `json:"raw_output_loaded"`
}

type HostOwnedWorkflowRuntimeExecutorInvocationInput struct {
	Readiness                 HostOwnedWorkflowRuntimeExecutorReadiness `json:"readiness,omitempty"`
	InvocationReportRef       DisplaySafeRef                            `json:"invocation_report_ref,omitempty"`
	ObservedInvocationRef     DisplaySafeRef                            `json:"observed_invocation_ref,omitempty"`
	HostRuntimeExecutorRunRef DisplaySafeRef                            `json:"host_runtime_executor_run_ref,omitempty"`
	WorkflowResultRef         DisplaySafeRef                            `json:"workflow_result_ref,omitempty"`
	WorkflowReadbackRef       DisplaySafeRef                            `json:"workflow_readback_ref,omitempty"`
	ObservedWorkflowRunRef    DisplaySafeRef                            `json:"observed_workflow_run_ref,omitempty"`
	ObservationRef            DisplaySafeRef                            `json:"observation_ref,omitempty"`
	FailureRef                DisplaySafeRef                            `json:"failure_ref,omitempty"`
	CompensationRef           DisplaySafeRef                            `json:"compensation_ref,omitempty"`
	HostInvocationReported    bool                                      `json:"host_invocation_reported"`
	HostInvocationCompleted   bool                                      `json:"host_invocation_completed"`
	HostInvocationFailed      bool                                      `json:"host_invocation_failed"`
	EvidenceRefs              []EvidenceRef                             `json:"evidence_refs,omitempty"`
	Boundaries                []Boundary                                `json:"boundaries,omitempty"`
	RawOutputLoaded           bool                                      `json:"raw_output_loaded"`
}

type HostOwnedWorkflowRuntimeExecutorInvocation struct {
	ContractVersion             string                                    `json:"contract_version,omitempty"`
	Projected                   bool                                      `json:"projected"`
	Status                      HostActionStatus                          `json:"status,omitempty"`
	ReadyForWorkflowObservation bool                                      `json:"ready_for_workflow_observation"`
	ReadyForFailureReview       bool                                      `json:"ready_for_failure_review"`
	HostInvocationReported      bool                                      `json:"host_invocation_reported"`
	HostInvocationCompleted     bool                                      `json:"host_invocation_completed"`
	HostInvocationFailed        bool                                      `json:"host_invocation_failed"`
	Readiness                   HostOwnedWorkflowRuntimeExecutorReadiness `json:"readiness,omitempty"`
	SourceRef                   DisplaySafeRef                            `json:"source_ref,omitempty"`
	WorkflowRunRef              DisplaySafeRef                            `json:"workflow_run_ref,omitempty"`
	RetryRequestRef             DisplaySafeRef                            `json:"retry_request_ref,omitempty"`
	CurrentStrategyRef          DisplaySafeRef                            `json:"current_strategy_ref,omitempty"`
	NextStrategyRef             DisplaySafeRef                            `json:"next_strategy_ref,omitempty"`
	AdapterRef                  DisplaySafeRef                            `json:"adapter_ref,omitempty"`
	AdapterVersionRef           DisplaySafeRef                            `json:"adapter_version_ref,omitempty"`
	InvocationRef               DisplaySafeRef                            `json:"invocation_ref,omitempty"`
	InvocationReportRef         DisplaySafeRef                            `json:"invocation_report_ref,omitempty"`
	ObservedInvocationRef       DisplaySafeRef                            `json:"observed_invocation_ref,omitempty"`
	HostRuntimeExecutorRunRef   DisplaySafeRef                            `json:"host_runtime_executor_run_ref,omitempty"`
	WorkflowResultRef           DisplaySafeRef                            `json:"workflow_result_ref,omitempty"`
	WorkflowReadbackRef         DisplaySafeRef                            `json:"workflow_readback_ref,omitempty"`
	ObservedWorkflowRunRef      DisplaySafeRef                            `json:"observed_workflow_run_ref,omitempty"`
	ObservationRef              DisplaySafeRef                            `json:"observation_ref,omitempty"`
	FailureRef                  DisplaySafeRef                            `json:"failure_ref,omitempty"`
	CompensationRef             DisplaySafeRef                            `json:"compensation_ref,omitempty"`
	EvidenceRefs                []EvidenceRef                             `json:"evidence_refs,omitempty"`
	FailureClass                FailureClass                              `json:"failure_class,omitempty"`
	BlockedReasons              []string                                  `json:"blocked_reasons,omitempty"`
	MissingInputs               []MissingInput                            `json:"missing_inputs,omitempty"`
	Boundaries                  []Boundary                                `json:"boundaries,omitempty"`
	NextHostAction              NextHostAction                            `json:"next_host_action,omitempty"`
	ExecutorEffect              string                                    `json:"executor_effect,omitempty"`
	RunnerEffect                string                                    `json:"runner_effect,omitempty"`
	PromptEffect                string                                    `json:"prompt_effect,omitempty"`
	RuntimeEffect               string                                    `json:"runtime_effect,omitempty"`
	CoreWorkflowRetryApplied    bool                                      `json:"core_workflow_retry_applied"`
	CoreRuntimeExecutorInvoked  bool                                      `json:"core_runtime_executor_invoked"`
	RunnerDispatched            bool                                      `json:"runner_dispatched"`
	RuntimeAdapterExecuted      bool                                      `json:"runtime_adapter_executed"`
	WorkflowDispatched          bool                                      `json:"workflow_dispatched"`
	WorkerDispatched            bool                                      `json:"worker_dispatched"`
	StoreMutationExecuted       bool                                      `json:"store_mutation_executed"`
	CompensationExecuted        bool                                      `json:"compensation_executed"`
	RawOutputLoaded             bool                                      `json:"raw_output_loaded"`
}

func BuildHostOwnedWorkflowRuntimeExecutorReadiness(input HostOwnedWorkflowRuntimeExecutorReadinessInput) HostOwnedWorkflowRuntimeExecutorReadiness {
	source := input.SourceProjection.Normalize()
	verification := input.Verification.Normalize()
	decision := input.ReplannerDecision.Normalize()
	workflowGate := input.WorkflowRetryGate.Normalize()
	runtimeGate := input.RuntimeExecutorGate.Normalize()
	currentStrategyRef := firstDisplaySafeRef(decision.CurrentStrategyRef, DisplaySafeRef(source.Candidate.ID))
	nextStrategyRef := firstDisplaySafeRef(decision.NextStrategyRef, DisplaySafeRef(source.Candidate.ID), currentStrategyRef)
	result := HostOwnedWorkflowRuntimeExecutorReadiness{
		ContractVersion:      ContractVersion,
		Projected:            true,
		Status:               HostActionBlocked,
		SourceProjection:     source,
		Verification:         verification,
		ReplannerDecision:    decision,
		WorkflowRetryGate:    workflowGate,
		RuntimeExecutorGate:  runtimeGate,
		SourceRef:            source.SourceRef,
		WorkflowRunRef:       normalizeOneDisplaySafeRef(input.WorkflowRunRef),
		RetryRequestRef:      normalizeOneDisplaySafeRef(input.RetryRequestRef),
		CurrentStrategyRef:   currentStrategyRef,
		NextStrategyRef:      nextStrategyRef,
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
		RetryPolicyRef:       normalizeOneDisplaySafeRef(input.RetryPolicyRef),
		FallbackPolicyRef:    normalizeOneDisplaySafeRef(input.FallbackPolicyRef),
		FailureBindingRef:    normalizeOneDisplaySafeRef(input.FailureBindingRef),
		CompensationRef:      normalizeOneDisplaySafeRef(input.CompensationRef),
		EvidenceRefs:         MergeEvidenceRefs(input.EvidenceRefs, source.EvidenceRefs, verification.EvidenceRefs, decision.EvidenceRefs),
		FailureClass:         firstFailureClass(decision.FailureClass, verification.FailureClass, source.FailureClass, FailureNone),
		DecisionBasis: normalizeDisplaySafeRefs(append(
			[]DisplaySafeRef{
				"workflow_runtime_executor:host_owned",
				"workflow_runtime_executor:readiness",
			},
			input.DecisionBasis...,
		)),
		Boundaries:     hostOwnedWorkflowRuntimeExecutorReadinessBoundaries(source.Boundaries, verification.Boundaries, decision.Boundaries, workflowGate.Boundaries, runtimeGate.Boundaries, input.Boundaries),
		NextHostAction: "provide_workflow_runtime_executor_binding",
		ExecutorEffect: "none",
		RunnerEffect:   "none",
		PromptEffect:   "none",
		RuntimeEffect:  "none",
		RawOutputLoaded: input.RawOutputLoaded ||
			source.RawOutputLoaded ||
			verification.RawOutputLoaded ||
			decision.RawOutputLoaded ||
			workflowGate.RawOutputLoaded ||
			runtimeGate.RawOutputLoaded,
	}
	if hostOwnedWorkflowRuntimeExecutorReadinessUnsafe(input, source, verification, decision, workflowGate, runtimeGate) {
		result.RawOutputLoaded = true
		result = hostOwnedWorkflowRuntimeExecutorReadinessBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if source.SourceKind != ReplannerSourceWorkflow {
		result = hostOwnedWorkflowRuntimeExecutorReadinessBlock(result, FailureInvalidInput, "workflow_source_projection_required", "source:workflow_projection", "provide_workflow_source_projection")
	}
	if source.Status != VerificationFailed && source.Status != VerificationPartial {
		result = hostOwnedWorkflowRuntimeExecutorReadinessBlock(result, FailureVerificationFailed, "workflow_source_not_failed_or_partial", "source:workflow_failed_or_partial", "provide_failed_or_partial_workflow_observation")
	}
	if verification.Verification.Status != "" && verification.Verification.Status != VerificationFailed && verification.Verification.Status != VerificationPartial {
		result = hostOwnedWorkflowRuntimeExecutorReadinessBlock(result, FailureVerificationFailed, "verification_not_failed_or_partial", "verification:workflow_failed_or_partial", "provide_workflow_verification_failure")
	}
	if decision.Action != ObjectiveReplannerActionRetrySameStrategy && decision.Action != ObjectiveReplannerActionSwitchStrategy {
		result = hostOwnedWorkflowRuntimeExecutorReadinessBlock(result, firstFailureClass(decision.FailureClass, FailureVerificationFailed), "replanner_decision_not_retry_or_switch", "replanner:retry_or_switch_decision", "request_host_replanner_decision")
	}
	if decision.Action == ObjectiveReplannerActionRetrySameStrategy {
		if result.CurrentStrategyRef == "" {
			result = hostOwnedWorkflowRuntimeExecutorReadinessBlock(result, FailureConfigMissing, "current_strategy_ref_missing", "strategy:current", "provide_current_strategy_ref")
		}
		if result.NextStrategyRef != "" && result.CurrentStrategyRef != "" && result.NextStrategyRef != result.CurrentStrategyRef {
			result = hostOwnedWorkflowRuntimeExecutorReadinessBlock(result, FailurePolicyBlocked, "same_strategy_retry_ref_mismatch", "strategy:same_strategy_retry", "review_workflow_retry_decision")
		}
		if !hostOwnedWorkflowRuntimeExecutorGateReady(workflowGate, ProductionAdapterEffectGateWorkflowRetryApply) {
			result = hostOwnedWorkflowRuntimeExecutorReadinessBlock(result, firstFailureClass(workflowGate.FailureClass, FailureConfigMissing), "workflow_retry_gate_not_ready", "host:workflow_retry_apply_gate", "provide_workflow_retry_apply_gate")
		}
	}
	if decision.Action == ObjectiveReplannerActionSwitchStrategy {
		if result.CurrentStrategyRef == "" || result.NextStrategyRef == "" || result.CurrentStrategyRef == result.NextStrategyRef {
			result = hostOwnedWorkflowRuntimeExecutorReadinessBlock(result, FailureConfigMissing, "l3_fallback_strategy_ref_missing", "strategy:l3_fallback", "provide_l3_fallback_strategy_ref")
		}
	}
	if !hostOwnedWorkflowRuntimeExecutorGateReady(runtimeGate, ProductionAdapterEffectGateRuntimeExecutor) {
		result = hostOwnedWorkflowRuntimeExecutorReadinessBlock(result, firstFailureClass(runtimeGate.FailureClass, FailureConfigMissing), "runtime_executor_gate_not_ready", "host:runtime_executor_gate", "provide_runtime_executor_gate")
	}
	for _, check := range []struct {
		ok      bool
		reason  string
		missing MissingInput
		next    NextHostAction
	}{
		{result.AdapterRef != "", "adapter_ref_missing", "host:workflow_runtime_executor_adapter_ref", "provide_workflow_runtime_executor_binding"},
		{result.AdapterVersionRef != "", "adapter_version_ref_missing", "host:workflow_runtime_executor_adapter_version_ref", "provide_workflow_runtime_executor_binding"},
		{result.AdapterCapabilityRef != "", "adapter_capability_ref_missing", "host:workflow_runtime_executor_adapter_capability_ref", "provide_workflow_runtime_executor_capability"},
		{result.AdapterContractRef != "", "adapter_contract_ref_missing", "contract:workflow_runtime_executor_adapter", "provide_workflow_runtime_executor_contract"},
		{result.HostConfirmationRef != "", "host_confirmation_ref_missing", "host:workflow_runtime_executor_confirmation", "request_workflow_runtime_executor_confirmation"},
		{result.WorkflowRunRef != "", "workflow_run_ref_missing", "host:workflow_run_ref", "provide_workflow_run_ref"},
		{result.RetryRequestRef != "", "retry_request_ref_missing", "host:workflow_retry_request_ref", "provide_workflow_retry_request_ref"},
		{result.InvocationRef != "", "invocation_ref_missing", "host:workflow_runtime_executor_invocation_ref", "provide_workflow_runtime_executor_invocation_ref"},
		{result.ResultBindingRef != "", "result_binding_ref_missing", "host:workflow_result_binding", "provide_workflow_result_binding"},
		{result.ReadbackBindingRef != "", "readback_binding_ref_missing", "host:workflow_readback_binding", "provide_workflow_readback_binding"},
		{result.IdempotencyRef != "", "idempotency_ref_missing", "host:workflow_retry_idempotency", "provide_workflow_retry_idempotency"},
		{result.BudgetRef != "", "budget_ref_missing", "host:workflow_retry_budget", "provide_workflow_retry_budget"},
		{result.FailureBindingRef != "", "failure_binding_ref_missing", "host:workflow_failure_binding", "provide_workflow_failure_binding"},
	} {
		if !check.ok {
			result = hostOwnedWorkflowRuntimeExecutorReadinessBlock(result, FailureConfigMissing, check.reason, check.missing, check.next)
		}
	}
	if len(result.MissingInputs) == 0 && len(result.BlockedReasons) == 0 {
		result.Status = HostActionReady
		result.ReadyForHostWorkflowRuntimeInvocation = true
		result.ReadyForSameStrategyRetry = decision.Action == ObjectiveReplannerActionRetrySameStrategy
		result.ReadyForL3Fallback = decision.Action == ObjectiveReplannerActionSwitchStrategy
		result.HostRuntimeExecutorInvocationAuthorized = true
		result.HostMayInvokeRuntimeExecutor = true
		result.NextHostAction = "host_may_invoke_workflow_runtime_executor"
		result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_host_owned_workflow_runtime_executor_invocation", "host_may_invoke_workflow_runtime_executor")
		if result.ReadyForSameStrategyRetry {
			result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_host_same_strategy_workflow_retry")
		}
		if result.ReadyForL3Fallback {
			result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_host_l3_workflow_fallback")
		}
	}
	return result.Normalize()
}

func BuildHostOwnedWorkflowRuntimeExecutorInvocation(input HostOwnedWorkflowRuntimeExecutorInvocationInput) HostOwnedWorkflowRuntimeExecutorInvocation {
	readiness := input.Readiness.Normalize()
	result := HostOwnedWorkflowRuntimeExecutorInvocation{
		ContractVersion:           ContractVersion,
		Projected:                 true,
		Status:                    HostActionBlocked,
		Readiness:                 readiness,
		SourceRef:                 readiness.SourceRef,
		WorkflowRunRef:            readiness.WorkflowRunRef,
		RetryRequestRef:           readiness.RetryRequestRef,
		CurrentStrategyRef:        readiness.CurrentStrategyRef,
		NextStrategyRef:           readiness.NextStrategyRef,
		AdapterRef:                readiness.AdapterRef,
		AdapterVersionRef:         readiness.AdapterVersionRef,
		InvocationRef:             readiness.InvocationRef,
		InvocationReportRef:       normalizeOneDisplaySafeRef(input.InvocationReportRef),
		ObservedInvocationRef:     normalizeOneDisplaySafeRef(input.ObservedInvocationRef),
		HostRuntimeExecutorRunRef: normalizeOneDisplaySafeRef(input.HostRuntimeExecutorRunRef),
		WorkflowResultRef:         normalizeOneDisplaySafeRef(input.WorkflowResultRef),
		WorkflowReadbackRef:       normalizeOneDisplaySafeRef(input.WorkflowReadbackRef),
		ObservedWorkflowRunRef:    normalizeOneDisplaySafeRef(input.ObservedWorkflowRunRef),
		ObservationRef:            normalizeOneDisplaySafeRef(input.ObservationRef),
		FailureRef:                normalizeOneDisplaySafeRef(input.FailureRef),
		CompensationRef:           normalizeOneDisplaySafeRef(input.CompensationRef),
		HostInvocationReported:    input.HostInvocationReported,
		HostInvocationCompleted:   input.HostInvocationCompleted,
		HostInvocationFailed:      input.HostInvocationFailed,
		EvidenceRefs:              MergeEvidenceRefs(input.EvidenceRefs, readiness.EvidenceRefs),
		FailureClass:              readiness.FailureClass,
		Boundaries:                hostOwnedWorkflowRuntimeExecutorInvocationBoundaries(readiness.Boundaries, input.Boundaries),
		NextHostAction:            "provide_workflow_runtime_executor_invocation_report",
		ExecutorEffect:            "none",
		RunnerEffect:              "none",
		PromptEffect:              "none",
		RuntimeEffect:             "none",
		RawOutputLoaded:           input.RawOutputLoaded || readiness.RawOutputLoaded,
	}
	if hostOwnedWorkflowRuntimeExecutorInvocationUnsafe(input, readiness) {
		result.RawOutputLoaded = true
		result = hostOwnedWorkflowRuntimeExecutorInvocationBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if !readiness.ReadyForHostWorkflowRuntimeInvocation {
		result = hostOwnedWorkflowRuntimeExecutorInvocationBlock(result, firstFailureClass(readiness.FailureClass, FailureConfigMissing), "workflow_runtime_executor_readiness_not_ready", "host:workflow_runtime_executor_readiness", firstNextHostAction(readiness.NextHostAction, "review_workflow_runtime_executor_readiness"))
	}
	for _, check := range []struct {
		ok      bool
		reason  string
		missing MissingInput
		next    NextHostAction
	}{
		{result.InvocationReportRef != "", "invocation_report_ref_missing", "host:workflow_runtime_executor_invocation_report", "provide_workflow_runtime_executor_invocation_report"},
		{result.ObservedInvocationRef != "", "observed_invocation_ref_missing", "host:workflow_runtime_executor_observed_invocation", "provide_workflow_runtime_executor_invocation_report"},
		{result.HostRuntimeExecutorRunRef != "", "host_runtime_executor_run_ref_missing", "host:workflow_runtime_executor_run_ref", "provide_workflow_runtime_executor_run_ref"},
		{input.HostInvocationReported, "host_invocation_not_reported", "host:workflow_runtime_executor_invocation_report", "provide_workflow_runtime_executor_invocation_report"},
	} {
		if !check.ok {
			result = hostOwnedWorkflowRuntimeExecutorInvocationBlock(result, FailureConfigMissing, check.reason, check.missing, check.next)
		}
	}
	if result.ObservedInvocationRef != "" && result.InvocationRef != "" && result.ObservedInvocationRef != result.InvocationRef {
		result = hostOwnedWorkflowRuntimeExecutorInvocationBlock(result, FailureVerificationFailed, "observed_invocation_ref_mismatch", "host:workflow_runtime_executor_invocation_report", "review_workflow_runtime_executor_invocation_report")
	}
	if result.ObservedWorkflowRunRef != "" && result.WorkflowRunRef != "" && result.ObservedWorkflowRunRef != result.WorkflowRunRef {
		result = hostOwnedWorkflowRuntimeExecutorInvocationBlock(result, FailureVerificationFailed, "observed_workflow_run_ref_mismatch", "host:workflow_run_ref", "review_workflow_runtime_executor_invocation_report")
	}
	if input.HostInvocationFailed {
		if result.FailureRef == "" {
			result = hostOwnedWorkflowRuntimeExecutorInvocationBlock(result, FailureVerificationFailed, "failure_ref_missing", "host:workflow_failure_ref", "provide_workflow_failure_review")
		}
		if result.CompensationRef == "" {
			result = hostOwnedWorkflowRuntimeExecutorInvocationBlock(result, FailureConfigMissing, "compensation_ref_missing", "host:workflow_compensation_ref", "provide_workflow_compensation_ref")
		}
		result.ReadyForFailureReview = result.FailureRef != "" && result.CompensationRef != ""
		result.NextHostAction = "review_workflow_runtime_executor_failure"
		result.Boundaries = AppendBoundaries(result.Boundaries, "workflow_runtime_executor_failure_reported")
	}
	if input.HostInvocationCompleted {
		for _, check := range []struct {
			ok      bool
			reason  string
			missing MissingInput
			next    NextHostAction
		}{
			{result.WorkflowResultRef != "", "workflow_result_ref_missing", "host:workflow_result_ref", "provide_workflow_result_ref"},
			{result.WorkflowReadbackRef != "", "workflow_readback_ref_missing", "host:workflow_readback_ref", "provide_workflow_readback_ref"},
			{result.ObservedWorkflowRunRef != "", "observed_workflow_run_ref_missing", "host:workflow_run_ref", "provide_workflow_run_ref"},
			{result.ObservationRef != "", "observation_ref_missing", "host:workflow_result_observation_ref", "provide_workflow_result_observation_ref"},
		} {
			if !check.ok {
				result = hostOwnedWorkflowRuntimeExecutorInvocationBlock(result, FailureConfigMissing, check.reason, check.missing, check.next)
			}
		}
	}
	if len(result.MissingInputs) == 0 && len(result.BlockedReasons) == 0 {
		result.Status = HostActionRecorded
		result.ReadyForWorkflowObservation = input.HostInvocationCompleted
		result.ReadyForFailureReview = input.HostInvocationFailed
		result.NextHostAction = "host_may_project_workflow_runtime_observation"
		result.Boundaries = AppendBoundaries(result.Boundaries, "host_owned_workflow_runtime_executor_invocation_recorded")
		if result.ReadyForWorkflowObservation {
			result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_workflow_result_observation")
		}
	}
	return result.Normalize()
}

func CloneHostOwnedWorkflowRuntimeExecutorReadiness(in HostOwnedWorkflowRuntimeExecutorReadiness) HostOwnedWorkflowRuntimeExecutorReadiness {
	out := in
	out.SourceProjection = in.SourceProjection.Clone()
	out.Verification = in.Verification.Clone()
	out.ReplannerDecision = in.ReplannerDecision.Clone()
	out.WorkflowRetryGate = in.WorkflowRetryGate.Clone()
	out.RuntimeExecutorGate = in.RuntimeExecutorGate.Clone()
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.DecisionBasis = cloneDisplaySafeRefs(in.DecisionBasis)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (r HostOwnedWorkflowRuntimeExecutorReadiness) Clone() HostOwnedWorkflowRuntimeExecutorReadiness {
	return CloneHostOwnedWorkflowRuntimeExecutorReadiness(r)
}

func (r HostOwnedWorkflowRuntimeExecutorReadiness) Normalize() HostOwnedWorkflowRuntimeExecutorReadiness {
	out := CloneHostOwnedWorkflowRuntimeExecutorReadiness(r)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.SourceProjection = out.SourceProjection.Normalize()
	out.Verification = out.Verification.Normalize()
	out.ReplannerDecision = out.ReplannerDecision.Normalize()
	out.WorkflowRetryGate = out.WorkflowRetryGate.Normalize()
	out.RuntimeExecutorGate = out.RuntimeExecutorGate.Normalize()
	out.SourceRef = normalizeOneDisplaySafeRef(out.SourceRef)
	out.WorkflowRunRef = normalizeOneDisplaySafeRef(out.WorkflowRunRef)
	out.RetryRequestRef = normalizeOneDisplaySafeRef(out.RetryRequestRef)
	out.CurrentStrategyRef = normalizeOneDisplaySafeRef(out.CurrentStrategyRef)
	out.NextStrategyRef = normalizeOneDisplaySafeRef(out.NextStrategyRef)
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
	out.RetryPolicyRef = normalizeOneDisplaySafeRef(out.RetryPolicyRef)
	out.FallbackPolicyRef = normalizeOneDisplaySafeRef(out.FallbackPolicyRef)
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
	out = hostOwnedWorkflowRuntimeExecutorReadinessResetEffects(out)
	if out.RawOutputLoaded {
		out.Status = HostActionReviewRequired
		out.ReadyForHostWorkflowRuntimeInvocation = false
		out.ReadyForSameStrategyRetry = false
		out.ReadyForL3Fallback = false
		out.HostRuntimeExecutorInvocationAuthorized = false
		out.HostMayInvokeRuntimeExecutor = false
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
		out.ReadyForHostWorkflowRuntimeInvocation = false
		out.ReadyForSameStrategyRetry = false
		out.ReadyForL3Fallback = false
		out.HostRuntimeExecutorInvocationAuthorized = false
		out.HostMayInvokeRuntimeExecutor = false
		if out.Status == HostActionReady {
			out.Status = HostActionBlocked
		}
	}
	return out
}

func CloneHostOwnedWorkflowRuntimeExecutorInvocation(in HostOwnedWorkflowRuntimeExecutorInvocation) HostOwnedWorkflowRuntimeExecutorInvocation {
	out := in
	out.Readiness = in.Readiness.Clone()
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (i HostOwnedWorkflowRuntimeExecutorInvocation) Clone() HostOwnedWorkflowRuntimeExecutorInvocation {
	return CloneHostOwnedWorkflowRuntimeExecutorInvocation(i)
}

func (i HostOwnedWorkflowRuntimeExecutorInvocation) Normalize() HostOwnedWorkflowRuntimeExecutorInvocation {
	out := CloneHostOwnedWorkflowRuntimeExecutorInvocation(i)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.Readiness = out.Readiness.Normalize()
	out.SourceRef = normalizeOneDisplaySafeRef(out.SourceRef)
	out.WorkflowRunRef = normalizeOneDisplaySafeRef(out.WorkflowRunRef)
	out.RetryRequestRef = normalizeOneDisplaySafeRef(out.RetryRequestRef)
	out.CurrentStrategyRef = normalizeOneDisplaySafeRef(out.CurrentStrategyRef)
	out.NextStrategyRef = normalizeOneDisplaySafeRef(out.NextStrategyRef)
	out.AdapterRef = normalizeOneDisplaySafeRef(out.AdapterRef)
	out.AdapterVersionRef = normalizeOneDisplaySafeRef(out.AdapterVersionRef)
	out.InvocationRef = normalizeOneDisplaySafeRef(out.InvocationRef)
	out.InvocationReportRef = normalizeOneDisplaySafeRef(out.InvocationReportRef)
	out.ObservedInvocationRef = normalizeOneDisplaySafeRef(out.ObservedInvocationRef)
	out.HostRuntimeExecutorRunRef = normalizeOneDisplaySafeRef(out.HostRuntimeExecutorRunRef)
	out.WorkflowResultRef = normalizeOneDisplaySafeRef(out.WorkflowResultRef)
	out.WorkflowReadbackRef = normalizeOneDisplaySafeRef(out.WorkflowReadbackRef)
	out.ObservedWorkflowRunRef = normalizeOneDisplaySafeRef(out.ObservedWorkflowRunRef)
	out.ObservationRef = normalizeOneDisplaySafeRef(out.ObservationRef)
	out.FailureRef = normalizeOneDisplaySafeRef(out.FailureRef)
	out.CompensationRef = normalizeOneDisplaySafeRef(out.CompensationRef)
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
	out = hostOwnedWorkflowRuntimeExecutorInvocationResetEffects(out)
	if out.RawOutputLoaded {
		out.Status = HostActionReviewRequired
		out.ReadyForWorkflowObservation = false
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
		out.ReadyForWorkflowObservation = false
		out.ReadyForFailureReview = false
		if out.Status == HostActionRecorded {
			out.Status = HostActionBlocked
		}
	}
	return out
}

func hostOwnedWorkflowRuntimeExecutorGateReady(gate ProductionAdapterIndependentEffectGate, kind ProductionAdapterEffectGateKind) bool {
	normalized := gate.Normalize()
	return normalized.Kind == kind &&
		normalized.ReadyForIndependentGatePlan &&
		normalized.Status == HostActionReady &&
		len(normalized.MissingInputs) == 0 &&
		len(normalized.BlockedReasons) == 0 &&
		!normalized.RawOutputLoaded
}

func hostOwnedWorkflowRuntimeExecutorReadinessBlock(result HostOwnedWorkflowRuntimeExecutorReadiness, failure FailureClass, reason string, missing MissingInput, next NextHostAction) HostOwnedWorkflowRuntimeExecutorReadiness {
	result.Status = HostActionBlocked
	result.ReadyForHostWorkflowRuntimeInvocation = false
	result.ReadyForSameStrategyRetry = false
	result.ReadyForL3Fallback = false
	result.HostRuntimeExecutorInvocationAuthorized = false
	result.HostMayInvokeRuntimeExecutor = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = firstNextHostAction(next, result.NextHostAction)
	if result.NextHostAction == "" {
		result.NextHostAction = "review_workflow_runtime_executor_readiness"
	}
	return result
}

func hostOwnedWorkflowRuntimeExecutorInvocationBlock(result HostOwnedWorkflowRuntimeExecutorInvocation, failure FailureClass, reason string, missing MissingInput, next NextHostAction) HostOwnedWorkflowRuntimeExecutorInvocation {
	result.Status = HostActionBlocked
	result.ReadyForWorkflowObservation = false
	result.ReadyForFailureReview = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = firstNextHostAction(next, result.NextHostAction)
	if result.NextHostAction == "" {
		result.NextHostAction = "review_workflow_runtime_executor_invocation"
	}
	return result
}

func hostOwnedWorkflowRuntimeExecutorReadinessBoundaries(groups ...[]Boundary) []Boundary {
	base := []Boundary{
		"host_owned_workflow_runtime_executor_gate",
		"host_owned_workflow_retry_runtime_executor_readiness",
		"workflow_retry_runtime_executor_projection_only",
		"workflow_failed_partial_observation_required",
		"workflow_retry_from_replanner_decision_only",
		"host_owned_runtime_executor",
		"display_safe_refs_only",
		"explicit_host_confirmation_required",
		"no_second_workflow_engine_in_controlplane",
		"no_workflow_dispatch",
		"no_workflow_retry_apply_by_core",
		"no_runtime_executor_by_core",
		"no_runner_dispatch",
		"no_store_mutation_by_core",
		"no_compensation_executor_by_core",
	}
	for _, group := range groups {
		base = AppendBoundaries(base, group...)
	}
	return base
}

func hostOwnedWorkflowRuntimeExecutorInvocationBoundaries(groups ...[]Boundary) []Boundary {
	base := []Boundary{
		"host_owned_workflow_runtime_executor_invocation_report",
		"workflow_runtime_executor_invocation_report_only",
		"host_workflow_runtime_executor_reported_only",
		"display_safe_refs_only",
		"no_second_workflow_engine_in_controlplane",
		"no_workflow_dispatch",
		"no_workflow_retry_apply_by_core",
		"no_runtime_executor_by_core",
		"no_runner_dispatch",
		"no_store_mutation_by_core",
		"no_compensation_executor_by_core",
	}
	for _, group := range groups {
		base = AppendBoundaries(base, group...)
	}
	return base
}

func hostOwnedWorkflowRuntimeExecutorReadinessUnsafe(input HostOwnedWorkflowRuntimeExecutorReadinessInput, source ReplannerSourceProjection, verification ObjectiveVerificationGateResult, decision ObjectiveReplannerDecision, workflowGate ProductionAdapterIndependentEffectGate, runtimeGate ProductionAdapterIndependentEffectGate) bool {
	if input.RawOutputLoaded ||
		source.RawOutputLoaded ||
		verification.RawOutputLoaded ||
		decision.RawOutputLoaded ||
		workflowGate.RawOutputLoaded ||
		runtimeGate.RawOutputLoaded {
		return true
	}
	if displaySafeRefRejected(input.AdapterRef) ||
		displaySafeRefRejected(input.AdapterVersionRef) ||
		displaySafeRefRejected(input.AdapterCapabilityRef) ||
		displaySafeRefRejected(input.AdapterContractRef) ||
		displaySafeRefRejected(input.HostConfirmationRef) ||
		displaySafeRefRejected(input.WorkflowRunRef) ||
		displaySafeRefRejected(input.RetryRequestRef) ||
		displaySafeRefRejected(input.InvocationRef) ||
		displaySafeRefRejected(input.ResultBindingRef) ||
		displaySafeRefRejected(input.ReadbackBindingRef) ||
		displaySafeRefRejected(input.IdempotencyRef) ||
		displaySafeRefRejected(input.BudgetRef) ||
		displaySafeRefRejected(input.RetryPolicyRef) ||
		displaySafeRefRejected(input.FallbackPolicyRef) ||
		displaySafeRefRejected(input.FailureBindingRef) ||
		displaySafeRefRejected(input.CompensationRef) {
		return true
	}
	return false
}

func hostOwnedWorkflowRuntimeExecutorInvocationUnsafe(input HostOwnedWorkflowRuntimeExecutorInvocationInput, readiness HostOwnedWorkflowRuntimeExecutorReadiness) bool {
	if input.RawOutputLoaded || readiness.RawOutputLoaded {
		return true
	}
	return displaySafeRefRejected(input.InvocationReportRef) ||
		displaySafeRefRejected(input.ObservedInvocationRef) ||
		displaySafeRefRejected(input.HostRuntimeExecutorRunRef) ||
		displaySafeRefRejected(input.WorkflowResultRef) ||
		displaySafeRefRejected(input.WorkflowReadbackRef) ||
		displaySafeRefRejected(input.ObservedWorkflowRunRef) ||
		displaySafeRefRejected(input.ObservationRef) ||
		displaySafeRefRejected(input.FailureRef) ||
		displaySafeRefRejected(input.CompensationRef)
}

func hostOwnedWorkflowRuntimeExecutorReadinessResetEffects(in HostOwnedWorkflowRuntimeExecutorReadiness) HostOwnedWorkflowRuntimeExecutorReadiness {
	in.ExecutorEffect = "none"
	in.RunnerEffect = "none"
	in.PromptEffect = "none"
	in.RuntimeEffect = "none"
	in.CoreWorkflowRetryApplied = false
	in.CoreRuntimeExecutorInvoked = false
	in.RunnerDispatched = false
	in.RuntimeAdapterExecuted = false
	in.WorkflowDispatched = false
	in.WorkerDispatched = false
	in.StoreMutationExecuted = false
	in.CompensationExecuted = false
	return in
}

func hostOwnedWorkflowRuntimeExecutorInvocationResetEffects(in HostOwnedWorkflowRuntimeExecutorInvocation) HostOwnedWorkflowRuntimeExecutorInvocation {
	in.ExecutorEffect = "none"
	in.RunnerEffect = "none"
	in.PromptEffect = "none"
	in.RuntimeEffect = "none"
	in.CoreWorkflowRetryApplied = false
	in.CoreRuntimeExecutorInvoked = false
	in.RunnerDispatched = false
	in.RuntimeAdapterExecuted = false
	in.WorkflowDispatched = false
	in.WorkerDispatched = false
	in.StoreMutationExecuted = false
	in.CompensationExecuted = false
	return in
}
