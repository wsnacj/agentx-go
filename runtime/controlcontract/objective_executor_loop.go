package controlcontract

type HostOwnedObjectiveExecutorStepRequestInput struct {
	RuntimeLoop          ObjectiveRuntimeLoopStep    `json:"runtime_loop,omitempty"`
	Run                  ObjectiveRun                `json:"run,omitempty"`
	ControllerDecision   ObjectiveControllerDecision `json:"controller_decision,omitempty"`
	SelectedStrategy     StrategyCandidate           `json:"selected_strategy,omitempty"`
	FinalGate            IntensityGateResult         `json:"final_gate,omitempty"`
	HostExecutorRef      DisplaySafeRef              `json:"host_executor_ref,omitempty"`
	ExecutorStepRef      DisplaySafeRef              `json:"executor_step_ref,omitempty"`
	ExecutionContractRef DisplaySafeRef              `json:"execution_contract_ref,omitempty"`
	IdempotencyRef       DisplaySafeRef              `json:"idempotency_ref,omitempty"`
	ExpectedAttemptRef   AttemptRef                  `json:"expected_attempt_ref,omitempty"`
	ExpectedResultRef    DisplaySafeRef              `json:"expected_result_ref,omitempty"`
	ExpectedReadbackRef  DisplaySafeRef              `json:"expected_readback_ref,omitempty"`
	ApprovalRefs         []DisplaySafeRef            `json:"approval_refs,omitempty"`
	PolicyRefs           []DisplaySafeRef            `json:"policy_refs,omitempty"`
	EvidenceRefs         []EvidenceRef               `json:"evidence_refs,omitempty"`
	DecisionBasis        []DisplaySafeRef            `json:"decision_basis,omitempty"`
	Boundaries           []Boundary                  `json:"boundaries,omitempty"`
	RawOutputLoaded      bool                        `json:"raw_output_loaded"`
}

type HostOwnedObjectiveExecutorStepRequest struct {
	ContractVersion        string                      `json:"contract_version,omitempty"`
	Projected              bool                        `json:"projected"`
	Status                 HostActionStatus            `json:"status,omitempty"`
	ReadyForHostExecution  bool                        `json:"ready_for_host_execution"`
	RequestOnly            bool                        `json:"request_only"`
	Activation             Activation                  `json:"activation,omitempty"`
	Frame                  ObjectiveFrame              `json:"frame,omitempty"`
	Run                    ObjectiveRun                `json:"run,omitempty"`
	ControllerDecision     ObjectiveControllerDecision `json:"controller_decision,omitempty"`
	SelectedStrategy       StrategyCandidate           `json:"selected_strategy,omitempty"`
	FinalGate              IntensityGateResult         `json:"final_gate,omitempty"`
	HostExecutorRef        DisplaySafeRef              `json:"host_executor_ref,omitempty"`
	ExecutorStepRef        DisplaySafeRef              `json:"executor_step_ref,omitempty"`
	ExecutionContractRef   DisplaySafeRef              `json:"execution_contract_ref,omitempty"`
	IdempotencyRef         DisplaySafeRef              `json:"idempotency_ref,omitempty"`
	ExpectedAttemptRef     AttemptRef                  `json:"expected_attempt_ref,omitempty"`
	ExpectedResultRef      DisplaySafeRef              `json:"expected_result_ref,omitempty"`
	ExpectedReadbackRef    DisplaySafeRef              `json:"expected_readback_ref,omitempty"`
	ApprovalRefs           []DisplaySafeRef            `json:"approval_refs,omitempty"`
	PolicyRefs             []DisplaySafeRef            `json:"policy_refs,omitempty"`
	EvidenceRefs           []EvidenceRef               `json:"evidence_refs,omitempty"`
	FailureClass           FailureClass                `json:"failure_class,omitempty"`
	BlockedReasons         []string                    `json:"blocked_reasons,omitempty"`
	MissingInputs          []MissingInput              `json:"missing_inputs,omitempty"`
	DecisionBasis          []DisplaySafeRef            `json:"decision_basis,omitempty"`
	Boundaries             []Boundary                  `json:"boundaries,omitempty"`
	NextHostAction         NextHostAction              `json:"next_host_action,omitempty"`
	ExecutorEffect         string                      `json:"executor_effect,omitempty"`
	RunnerEffect           string                      `json:"runner_effect,omitempty"`
	PromptEffect           string                      `json:"prompt_effect,omitempty"`
	RuntimeEffect          string                      `json:"runtime_effect,omitempty"`
	HostExecutionReported  bool                        `json:"host_execution_reported"`
	CoreExecutionExecuted  bool                        `json:"core_execution_executed"`
	RunnerDispatched       bool                        `json:"runner_dispatched"`
	RuntimeAdapterExecuted bool                        `json:"runtime_adapter_executed"`
	SchedulerApplied       bool                        `json:"scheduler_applied"`
	InstallerExecuted      bool                        `json:"installer_executed"`
	WorkflowDispatched     bool                        `json:"workflow_dispatched"`
	WorkerDispatched       bool                        `json:"worker_dispatched"`
	StoreMutationExecuted  bool                        `json:"store_mutation_executed"`
	CompensationExecuted   bool                        `json:"compensation_executed"`
	RawOutputLoaded        bool                        `json:"raw_output_loaded"`
}

type HostOwnedObjectiveExecutorStepResultInput struct {
	Request               HostOwnedObjectiveExecutorStepRequest `json:"request,omitempty"`
	ExecutorStepResultRef DisplaySafeRef                        `json:"executor_step_result_ref,omitempty"`
	HostExecutorRunRef    DisplaySafeRef                        `json:"host_executor_run_ref,omitempty"`
	Attempt               AttemptSummary                        `json:"attempt,omitempty"`
	Observations          []Observation                         `json:"observations,omitempty"`
	EvidenceRefs          []EvidenceRef                         `json:"evidence_refs,omitempty"`
	FailureClass          FailureClass                          `json:"failure_class,omitempty"`
	FailureReason         string                                `json:"failure_reason,omitempty"`
	HostExecutionReported bool                                  `json:"host_execution_reported"`
	Boundaries            []Boundary                            `json:"boundaries,omitempty"`
	RawOutputLoaded       bool                                  `json:"raw_output_loaded"`
}

type HostOwnedObjectiveExecutorStepResult struct {
	ContractVersion                  string                                `json:"contract_version,omitempty"`
	Projected                        bool                                  `json:"projected"`
	Status                           HostActionStatus                      `json:"status,omitempty"`
	ReadyForExecutorReadback         bool                                  `json:"ready_for_executor_readback"`
	ReadyForObservationNormalization bool                                  `json:"ready_for_observation_normalization"`
	Request                          HostOwnedObjectiveExecutorStepRequest `json:"request,omitempty"`
	ExecutorStepRef                  DisplaySafeRef                        `json:"executor_step_ref,omitempty"`
	ExecutorStepResultRef            DisplaySafeRef                        `json:"executor_step_result_ref,omitempty"`
	HostExecutorRef                  DisplaySafeRef                        `json:"host_executor_ref,omitempty"`
	HostExecutorRunRef               DisplaySafeRef                        `json:"host_executor_run_ref,omitempty"`
	ExpectedAttemptRef               AttemptRef                            `json:"expected_attempt_ref,omitempty"`
	Attempt                          AttemptSummary                        `json:"attempt,omitempty"`
	ExpectedResultRef                DisplaySafeRef                        `json:"expected_result_ref,omitempty"`
	ExpectedReadbackRef              DisplaySafeRef                        `json:"expected_readback_ref,omitempty"`
	Observations                     []Observation                         `json:"observations,omitempty"`
	AttemptLedgerPatch               AttemptLedgerPatch                    `json:"attempt_ledger_patch,omitempty"`
	EvidenceRefs                     []EvidenceRef                         `json:"evidence_refs,omitempty"`
	FailureClass                     FailureClass                          `json:"failure_class,omitempty"`
	FailureReason                    string                                `json:"failure_reason,omitempty"`
	BlockedReasons                   []string                              `json:"blocked_reasons,omitempty"`
	MissingInputs                    []MissingInput                        `json:"missing_inputs,omitempty"`
	Boundaries                       []Boundary                            `json:"boundaries,omitempty"`
	NextHostAction                   NextHostAction                        `json:"next_host_action,omitempty"`
	ExecutorEffect                   string                                `json:"executor_effect,omitempty"`
	RunnerEffect                     string                                `json:"runner_effect,omitempty"`
	PromptEffect                     string                                `json:"prompt_effect,omitempty"`
	RuntimeEffect                    string                                `json:"runtime_effect,omitempty"`
	HostExecutionReported            bool                                  `json:"host_execution_reported"`
	CoreExecutionExecuted            bool                                  `json:"core_execution_executed"`
	RunnerDispatched                 bool                                  `json:"runner_dispatched"`
	RuntimeAdapterExecuted           bool                                  `json:"runtime_adapter_executed"`
	SchedulerApplied                 bool                                  `json:"scheduler_applied"`
	InstallerExecuted                bool                                  `json:"installer_executed"`
	WorkflowDispatched               bool                                  `json:"workflow_dispatched"`
	WorkerDispatched                 bool                                  `json:"worker_dispatched"`
	StoreMutationExecuted            bool                                  `json:"store_mutation_executed"`
	CompensationExecuted             bool                                  `json:"compensation_executed"`
	RawOutputLoaded                  bool                                  `json:"raw_output_loaded"`
}

type HostOwnedObjectiveExecutorStepReadbackInput struct {
	Request                 HostOwnedObjectiveExecutorStepRequest `json:"request,omitempty"`
	Result                  HostOwnedObjectiveExecutorStepResult  `json:"result,omitempty"`
	ExecutorStepReadbackRef DisplaySafeRef                        `json:"executor_step_readback_ref,omitempty"`
	ObservedResultRef       DisplaySafeRef                        `json:"observed_result_ref,omitempty"`
	ObservedAttemptRef      AttemptRef                            `json:"observed_attempt_ref,omitempty"`
	Verification            ObjectiveVerificationGateResult       `json:"verification,omitempty"`
	EvidenceRefs            []EvidenceRef                         `json:"evidence_refs,omitempty"`
	Boundaries              []Boundary                            `json:"boundaries,omitempty"`
	RawOutputLoaded         bool                                  `json:"raw_output_loaded"`
}

type HostOwnedObjectiveExecutorStepReadback struct {
	ContractVersion          string                                `json:"contract_version,omitempty"`
	Projected                bool                                  `json:"projected"`
	Status                   HostActionStatus                      `json:"status,omitempty"`
	ReadyForRuntimeLoopInput bool                                  `json:"ready_for_runtime_loop_input"`
	Request                  HostOwnedObjectiveExecutorStepRequest `json:"request,omitempty"`
	Result                   HostOwnedObjectiveExecutorStepResult  `json:"result,omitempty"`
	ExecutorStepRef          DisplaySafeRef                        `json:"executor_step_ref,omitempty"`
	ExecutorStepResultRef    DisplaySafeRef                        `json:"executor_step_result_ref,omitempty"`
	ExecutorStepReadbackRef  DisplaySafeRef                        `json:"executor_step_readback_ref,omitempty"`
	ExpectedReadbackRef      DisplaySafeRef                        `json:"expected_readback_ref,omitempty"`
	ObservedResultRef        DisplaySafeRef                        `json:"observed_result_ref,omitempty"`
	ObservedAttemptRef       AttemptRef                            `json:"observed_attempt_ref,omitempty"`
	AttemptLedgerPatch       AttemptLedgerPatch                    `json:"attempt_ledger_patch,omitempty"`
	Observations             []Observation                         `json:"observations,omitempty"`
	Verification             ObjectiveVerificationGateResult       `json:"verification,omitempty"`
	EvidenceRefs             []EvidenceRef                         `json:"evidence_refs,omitempty"`
	FailureClass             FailureClass                          `json:"failure_class,omitempty"`
	BlockedReasons           []string                              `json:"blocked_reasons,omitempty"`
	MissingInputs            []MissingInput                        `json:"missing_inputs,omitempty"`
	Boundaries               []Boundary                            `json:"boundaries,omitempty"`
	NextHostAction           NextHostAction                        `json:"next_host_action,omitempty"`
	ExecutorEffect           string                                `json:"executor_effect,omitempty"`
	RunnerEffect             string                                `json:"runner_effect,omitempty"`
	PromptEffect             string                                `json:"prompt_effect,omitempty"`
	RuntimeEffect            string                                `json:"runtime_effect,omitempty"`
	CoreExecutionExecuted    bool                                  `json:"core_execution_executed"`
	RunnerDispatched         bool                                  `json:"runner_dispatched"`
	RuntimeAdapterExecuted   bool                                  `json:"runtime_adapter_executed"`
	SchedulerApplied         bool                                  `json:"scheduler_applied"`
	InstallerExecuted        bool                                  `json:"installer_executed"`
	WorkflowDispatched       bool                                  `json:"workflow_dispatched"`
	WorkerDispatched         bool                                  `json:"worker_dispatched"`
	StoreMutationExecuted    bool                                  `json:"store_mutation_executed"`
	CompensationExecuted     bool                                  `json:"compensation_executed"`
	RawOutputLoaded          bool                                  `json:"raw_output_loaded"`
}

func BuildHostOwnedObjectiveExecutorStepRequest(input HostOwnedObjectiveExecutorStepRequestInput) HostOwnedObjectiveExecutorStepRequest {
	run := hostOwnedObjectiveExecutorRun(input)
	decision := hostOwnedObjectiveExecutorDecision(input, run)
	strategy := input.SelectedStrategy.Normalize()
	finalGate := input.FinalGate.Normalize()
	result := HostOwnedObjectiveExecutorStepRequest{
		ContractVersion:      ContractVersion,
		Projected:            true,
		Status:               HostActionBlocked,
		RequestOnly:          true,
		Activation:           run.Activation,
		Frame:                run.Frame,
		Run:                  run,
		ControllerDecision:   decision,
		SelectedStrategy:     strategy,
		FinalGate:            finalGate,
		HostExecutorRef:      normalizeOneDisplaySafeRef(input.HostExecutorRef),
		ExecutorStepRef:      normalizeOneDisplaySafeRef(input.ExecutorStepRef),
		ExecutionContractRef: normalizeOneDisplaySafeRef(input.ExecutionContractRef),
		IdempotencyRef:       normalizeOneDisplaySafeRef(input.IdempotencyRef),
		ExpectedAttemptRef:   normalizeOneAttemptRef(input.ExpectedAttemptRef),
		ExpectedResultRef:    normalizeOneDisplaySafeRef(input.ExpectedResultRef),
		ExpectedReadbackRef:  normalizeOneDisplaySafeRef(input.ExpectedReadbackRef),
		ApprovalRefs:         normalizeDisplaySafeRefs(append(cloneDisplaySafeRefs(run.Approval.ApprovalRefs), input.ApprovalRefs...)),
		PolicyRefs:           normalizeDisplaySafeRefs(append(append(cloneDisplaySafeRefs(run.PolicyRefs), finalGate.PolicyRefs...), input.PolicyRefs...)),
		EvidenceRefs:         MergeEvidenceRefs(input.EvidenceRefs, run.EvidenceRefs, decision.EvidenceRefs, finalGate.EvidenceRefs, strategy.ExpectedEvidence),
		FailureClass:         FailureNone,
		DecisionBasis: normalizeDisplaySafeRefs(append(
			[]DisplaySafeRef{
				"objective_executor:host_owned",
				"objective_executor:step_request",
			},
			input.DecisionBasis...,
		)),
		Boundaries:     hostOwnedObjectiveExecutorBoundaries(input.Boundaries),
		NextHostAction: "provide_host_owned_objective_executor_inputs",
		ExecutorEffect: "none",
		RunnerEffect:   "none",
		PromptEffect:   "none",
		RuntimeEffect:  "none",
		RawOutputLoaded: input.RawOutputLoaded ||
			input.RuntimeLoop.RawOutputLoaded ||
			run.RawOutputLoaded ||
			decision.RawOutputLoaded ||
			finalGate.RawOutputLoaded,
	}
	if hostOwnedObjectiveExecutorRequestUnsafe(input, run, decision, strategy, finalGate) {
		result.RawOutputLoaded = true
		result = hostOwnedObjectiveExecutorRequestBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if run.Activation != ActivationManaged || !run.FullRun {
		result = hostOwnedObjectiveExecutorRequestBlock(result, FailurePolicyBlocked, "managed_objective_required", "control_plane:managed_activation", "enable_managed_objective")
	}
	if !hostOwnedObjectiveExecutorRuntimeLoopReady(input.RuntimeLoop, decision) {
		result = hostOwnedObjectiveExecutorRequestBlock(result, FailureEvidenceMissing, "runtime_loop_not_ready_for_executor_step", "host:objective_runtime_loop_step", "provide_ready_objective_runtime_loop_step")
	}
	if !hostOwnedObjectiveExecutorActionExecutable(decision.Action) {
		result = hostOwnedObjectiveExecutorRequestBlock(result, FailurePolicyBlocked, "controller_action_not_executable", "host:controller_decision", "return_current_objective_state")
	}
	if strategyRef := hostOwnedObjectiveExecutorStrategyRef(strategy); strategyRef == "" {
		result = hostOwnedObjectiveExecutorRequestBlock(result, FailureConfigMissing, "selected_strategy_missing", "host:selected_strategy", "select_strategy")
	} else if !hostOwnedObjectiveExecutorStrategyAllowed(strategyRef, decision.AllowedStrategies, run.Strategies) {
		result = hostOwnedObjectiveExecutorRequestBlock(result, FailurePolicyBlocked, "selected_strategy_not_in_scope", "host:selected_strategy", "select_allowed_strategy")
	}
	if finalGate.Stage != IntensityGateFinal || !finalGate.Allowed {
		result = hostOwnedObjectiveExecutorRequestBlock(result, firstFailureClass(finalGate.FailureClass, FailurePolicyBlocked), "final_gate_not_satisfied", "host:execution_intensity_final_gate", firstNextHostAction(finalGate.NextHostAction, "run_strategy_final_gate"))
	}
	if gateStrategyRef := normalizeOneDisplaySafeRef(finalGate.StrategyRef); gateStrategyRef != "" && gateStrategyRef != hostOwnedObjectiveExecutorStrategyRef(strategy) {
		result = hostOwnedObjectiveExecutorRequestBlock(result, FailureVerificationFailed, "final_gate_strategy_ref_mismatch", "host:execution_intensity_final_gate", "run_strategy_final_gate")
	}
	if run.Budget.BudgetRef == "" || run.Budget.Exhausted {
		result = hostOwnedObjectiveExecutorRequestBlock(result, FailureBudgetExhausted, "executor_budget_not_available", "contract:budget", "provide_budget_policy")
	}
	if run.Approval.Required && !run.Approval.Approved {
		result = hostOwnedObjectiveExecutorRequestBlock(result, FailureApprovalRequired, "objective_approval_required", "host:objective_approval", "request_host_approval")
	}
	if run.Approval.Required && len(result.ApprovalRefs) == 0 {
		result = hostOwnedObjectiveExecutorRequestBlock(result, FailureEvidenceMissing, "objective_approval_ref_missing", "host:approval_ref", "provide_host_approval_ref")
	}
	for _, check := range []struct {
		ok      bool
		reason  string
		missing MissingInput
		next    NextHostAction
	}{
		{result.HostExecutorRef != "", "host_executor_ref_missing", "host:objective_executor_ref", "provide_objective_executor_ref"},
		{result.ExecutorStepRef != "", "executor_step_ref_missing", "host:objective_executor_step_ref", "provide_objective_executor_step_refs"},
		{result.ExecutionContractRef != "", "execution_contract_ref_missing", "contract:execution", "provide_execution_contract_ref"},
		{result.IdempotencyRef != "", "idempotency_ref_missing", "host:idempotency_ref", "provide_idempotency_ref"},
		{result.ExpectedAttemptRef != "", "expected_attempt_ref_missing", "host:expected_attempt_ref", "provide_objective_executor_step_refs"},
		{result.ExpectedResultRef != "", "expected_result_ref_missing", "host:expected_executor_result_ref", "provide_objective_executor_step_refs"},
		{result.ExpectedReadbackRef != "", "expected_readback_ref_missing", "host:expected_executor_readback_ref", "provide_objective_executor_step_refs"},
	} {
		if !check.ok {
			result = hostOwnedObjectiveExecutorRequestBlock(result, FailureConfigMissing, check.reason, check.missing, check.next)
		}
	}
	if len(result.MissingInputs) == 0 && len(result.BlockedReasons) == 0 {
		result.Status = HostActionReady
		result.ReadyForHostExecution = true
		result.NextHostAction = "host_may_execute_objective_step"
		result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_host_owned_executor_step")
	}
	return result.Normalize()
}

func BuildHostOwnedObjectiveExecutorStepResult(input HostOwnedObjectiveExecutorStepResultInput) HostOwnedObjectiveExecutorStepResult {
	request := input.Request.Normalize()
	attempt := input.Attempt.Normalize()
	observations := normalizeObservations(input.Observations)
	result := HostOwnedObjectiveExecutorStepResult{
		ContractVersion:       ContractVersion,
		Projected:             true,
		Status:                HostActionBlocked,
		Request:               request,
		ExecutorStepRef:       request.ExecutorStepRef,
		ExecutorStepResultRef: normalizeOneDisplaySafeRef(input.ExecutorStepResultRef),
		HostExecutorRef:       request.HostExecutorRef,
		HostExecutorRunRef:    normalizeOneDisplaySafeRef(input.HostExecutorRunRef),
		ExpectedAttemptRef:    request.ExpectedAttemptRef,
		Attempt:               attempt,
		ExpectedResultRef:     request.ExpectedResultRef,
		ExpectedReadbackRef:   request.ExpectedReadbackRef,
		Observations:          observations,
		EvidenceRefs:          MergeEvidenceRefs(input.EvidenceRefs, request.EvidenceRefs, attempt.EvidenceRefs, objectiveObservationEvidenceRefs(observations)),
		FailureClass:          firstFailureClass(input.FailureClass, attempt.FailureClass, FailureNone),
		FailureReason:         firstNonEmptyContractString(input.FailureReason, attempt.FailureReason),
		Boundaries:            hostOwnedObjectiveExecutorResultBoundaries(request.Boundaries, input.Boundaries),
		NextHostAction:        "provide_host_owned_objective_executor_result",
		ExecutorEffect:        "none",
		RunnerEffect:          "none",
		PromptEffect:          "none",
		RuntimeEffect:         "none",
		HostExecutionReported: input.HostExecutionReported,
		RawOutputLoaded:       input.RawOutputLoaded || request.RawOutputLoaded || attempt.RawOutputLoaded || observationSliceUnsafePayload(observations),
	}
	if hostOwnedObjectiveExecutorResultUnsafe(input, request, attempt, observations) {
		result.RawOutputLoaded = true
		result = hostOwnedObjectiveExecutorResultBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if !request.ReadyForHostExecution {
		result = hostOwnedObjectiveExecutorResultBlock(result, firstFailureClass(request.FailureClass, FailureConfigMissing), "executor_step_request_not_ready", "host:objective_executor_step_request", firstNextHostAction(request.NextHostAction, "review_objective_executor_step_request"))
	}
	if result.ExecutorStepResultRef == "" {
		result = hostOwnedObjectiveExecutorResultBlock(result, FailureEvidenceMissing, "executor_step_result_ref_missing", "host:objective_executor_step_result_ref", "provide_objective_executor_step_result")
	} else if result.ExpectedResultRef != "" && result.ExecutorStepResultRef != result.ExpectedResultRef {
		result = hostOwnedObjectiveExecutorResultBlock(result, FailureVerificationFailed, "executor_step_result_ref_mismatch", "host:objective_executor_step_result_ref", "review_objective_executor_step_result")
	}
	if result.HostExecutorRunRef == "" {
		result = hostOwnedObjectiveExecutorResultBlock(result, FailureEvidenceMissing, "host_executor_run_ref_missing", "host:objective_executor_run_ref", "provide_objective_executor_run_ref")
	}
	if !result.HostExecutionReported {
		result = hostOwnedObjectiveExecutorResultBlock(result, FailureEvidenceMissing, "host_execution_not_reported", "host:objective_executor_execution_report", "provide_objective_executor_execution_report")
	}
	if attempt.Ref == "" {
		result = hostOwnedObjectiveExecutorResultBlock(result, FailureEvidenceMissing, "attempt_ref_missing", "host:attempt_ref", "provide_attempt_summary")
	} else if result.ExpectedAttemptRef != "" && attempt.Ref != result.ExpectedAttemptRef {
		result = hostOwnedObjectiveExecutorResultBlock(result, FailureVerificationFailed, "attempt_ref_mismatch", "host:attempt_ref", "review_objective_executor_step_result")
	}
	if attempt.Status == VerificationNotEvaluated {
		result = hostOwnedObjectiveExecutorResultBlock(result, FailureEvidenceMissing, "attempt_status_missing", "host:attempt_status", "provide_attempt_summary")
	}
	if attempt.Status == VerificationSatisfied || attempt.Status == VerificationPartial {
		if len(observations) == 0 || len(result.EvidenceRefs) == 0 {
			result = hostOwnedObjectiveExecutorResultBlock(result, FailureEvidenceMissing, "executor_step_evidence_missing", "host:executor_step_evidence", "provide_executor_step_evidence")
		}
	}
	if len(result.MissingInputs) == 0 && len(result.BlockedReasons) == 0 {
		result.Status = HostActionRecorded
		result.ReadyForExecutorReadback = true
		result.ReadyForObservationNormalization = len(observations) > 0
		result.AttemptLedgerPatch = hostOwnedObjectiveExecutorLedgerPatch(request, attempt, result.EvidenceRefs)
		result.NextHostAction = "bind_objective_executor_step_readback"
		result.Boundaries = AppendBoundaries(result.Boundaries, "host_owned_executor_step_result_recorded")
	}
	return result.Normalize()
}

func BuildHostOwnedObjectiveExecutorStepReadback(input HostOwnedObjectiveExecutorStepReadbackInput) HostOwnedObjectiveExecutorStepReadback {
	request := input.Request.Normalize()
	resultProjection := input.Result.Normalize()
	verification := input.Verification.Normalize()
	readback := HostOwnedObjectiveExecutorStepReadback{
		ContractVersion:         ContractVersion,
		Projected:               true,
		Status:                  HostActionBlocked,
		Request:                 request,
		Result:                  resultProjection,
		ExecutorStepRef:         request.ExecutorStepRef,
		ExecutorStepResultRef:   resultProjection.ExecutorStepResultRef,
		ExecutorStepReadbackRef: normalizeOneDisplaySafeRef(input.ExecutorStepReadbackRef),
		ExpectedReadbackRef:     request.ExpectedReadbackRef,
		ObservedResultRef:       normalizeOneDisplaySafeRef(input.ObservedResultRef),
		ObservedAttemptRef:      normalizeOneAttemptRef(input.ObservedAttemptRef),
		AttemptLedgerPatch:      resultProjection.AttemptLedgerPatch,
		Observations:            cloneObservations(resultProjection.Observations),
		Verification:            verification,
		EvidenceRefs:            MergeEvidenceRefs(input.EvidenceRefs, resultProjection.EvidenceRefs, verification.EvidenceRefs),
		FailureClass:            firstFailureClass(resultProjection.FailureClass, verification.FailureClass, FailureNone),
		Boundaries:              hostOwnedObjectiveExecutorReadbackBoundaries(resultProjection.Boundaries, verification.Boundaries, input.Boundaries),
		NextHostAction:          "provide_host_owned_objective_executor_readback",
		ExecutorEffect:          "none",
		RunnerEffect:            "none",
		PromptEffect:            "none",
		RuntimeEffect:           "none",
		RawOutputLoaded:         input.RawOutputLoaded || request.RawOutputLoaded || resultProjection.RawOutputLoaded || verification.RawOutputLoaded,
	}
	if hostOwnedObjectiveExecutorReadbackUnsafe(input, request, resultProjection, verification) {
		readback.RawOutputLoaded = true
		readback = hostOwnedObjectiveExecutorReadbackBlock(readback, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return readback.Normalize()
	}
	if !resultProjection.ReadyForExecutorReadback {
		readback = hostOwnedObjectiveExecutorReadbackBlock(readback, firstFailureClass(resultProjection.FailureClass, FailureEvidenceMissing), "executor_step_result_not_ready", "host:objective_executor_step_result", firstNextHostAction(resultProjection.NextHostAction, "review_objective_executor_step_result"))
	}
	if readback.ExecutorStepReadbackRef == "" {
		readback = hostOwnedObjectiveExecutorReadbackBlock(readback, FailureEvidenceMissing, "executor_step_readback_ref_missing", "host:objective_executor_step_readback_ref", "provide_objective_executor_step_readback")
	} else if readback.ExpectedReadbackRef != "" && readback.ExecutorStepReadbackRef != readback.ExpectedReadbackRef {
		readback = hostOwnedObjectiveExecutorReadbackBlock(readback, FailureVerificationFailed, "executor_step_readback_ref_mismatch", "host:objective_executor_step_readback_ref", "review_objective_executor_step_readback")
	}
	if readback.ObservedResultRef == "" {
		readback = hostOwnedObjectiveExecutorReadbackBlock(readback, FailureEvidenceMissing, "observed_result_ref_missing", "host:observed_executor_result_ref", "provide_objective_executor_step_readback")
	} else if readback.ObservedResultRef != resultProjection.ExecutorStepResultRef {
		readback = hostOwnedObjectiveExecutorReadbackBlock(readback, FailureVerificationFailed, "observed_result_ref_mismatch", "host:observed_executor_result_ref", "review_objective_executor_step_readback")
	}
	if readback.ObservedAttemptRef == "" {
		readback = hostOwnedObjectiveExecutorReadbackBlock(readback, FailureEvidenceMissing, "observed_attempt_ref_missing", "host:observed_attempt_ref", "provide_objective_executor_step_readback")
	} else if readback.ObservedAttemptRef != resultProjection.Attempt.Ref {
		readback = hostOwnedObjectiveExecutorReadbackBlock(readback, FailureVerificationFailed, "observed_attempt_ref_mismatch", "host:observed_attempt_ref", "review_objective_executor_step_readback")
	}
	if objectiveRuntimeLoopVerificationGateEmpty(verification) || verification.Status == VerificationNotEvaluated {
		readback = hostOwnedObjectiveExecutorReadbackBlock(readback, FailureEvidenceMissing, "executor_step_verification_missing", "host:objective_verification", "run_objective_verification_gate")
	}
	if len(readback.MissingInputs) == 0 && len(readback.BlockedReasons) == 0 {
		readback.Status = HostActionRecorded
		readback.ReadyForRuntimeLoopInput = true
		readback.NextHostAction = "run_objective_runtime_loop_step"
		readback.Boundaries = AppendBoundaries(readback.Boundaries, "ready_for_objective_runtime_loop")
	}
	return readback.Normalize()
}

func CloneHostOwnedObjectiveExecutorStepRequest(in HostOwnedObjectiveExecutorStepRequest) HostOwnedObjectiveExecutorStepRequest {
	out := in
	out.Frame = in.Frame.Clone()
	out.Run = in.Run.Clone()
	out.ControllerDecision = in.ControllerDecision.Clone()
	out.SelectedStrategy = in.SelectedStrategy.Clone()
	out.FinalGate = in.FinalGate.Clone()
	out.ApprovalRefs = cloneDisplaySafeRefs(in.ApprovalRefs)
	out.PolicyRefs = cloneDisplaySafeRefs(in.PolicyRefs)
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.DecisionBasis = cloneDisplaySafeRefs(in.DecisionBasis)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (r HostOwnedObjectiveExecutorStepRequest) Clone() HostOwnedObjectiveExecutorStepRequest {
	return CloneHostOwnedObjectiveExecutorStepRequest(r)
}

func (r HostOwnedObjectiveExecutorStepRequest) Normalize() HostOwnedObjectiveExecutorStepRequest {
	out := CloneHostOwnedObjectiveExecutorStepRequest(r)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.Activation = NormalizeActivation(string(out.Activation))
	out.Frame = out.Frame.Normalize()
	out.Run = out.Run.Normalize()
	out.ControllerDecision = out.ControllerDecision.Normalize()
	out.SelectedStrategy = out.SelectedStrategy.Normalize()
	out.FinalGate = out.FinalGate.Normalize()
	out.HostExecutorRef = normalizeOneDisplaySafeRef(out.HostExecutorRef)
	out.ExecutorStepRef = normalizeOneDisplaySafeRef(out.ExecutorStepRef)
	out.ExecutionContractRef = normalizeOneDisplaySafeRef(out.ExecutionContractRef)
	out.IdempotencyRef = normalizeOneDisplaySafeRef(out.IdempotencyRef)
	out.ExpectedAttemptRef = normalizeOneAttemptRef(out.ExpectedAttemptRef)
	out.ExpectedResultRef = normalizeOneDisplaySafeRef(out.ExpectedResultRef)
	out.ExpectedReadbackRef = normalizeOneDisplaySafeRef(out.ExpectedReadbackRef)
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
	out = hostOwnedObjectiveExecutorNormalizeEffects(out)
	out.RequestOnly = true
	out.ReadyForHostExecution = out.Status == HostActionReady && len(out.MissingInputs) == 0 && len(out.BlockedReasons) == 0 && !out.RawOutputLoaded
	if out.RawOutputLoaded {
		out.Status = HostActionReviewRequired
		out.ReadyForHostExecution = false
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

func CloneHostOwnedObjectiveExecutorStepResult(in HostOwnedObjectiveExecutorStepResult) HostOwnedObjectiveExecutorStepResult {
	out := in
	out.Request = in.Request.Clone()
	out.Attempt = in.Attempt.Clone()
	out.Observations = cloneObservations(in.Observations)
	out.AttemptLedgerPatch = in.AttemptLedgerPatch.Clone()
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (r HostOwnedObjectiveExecutorStepResult) Clone() HostOwnedObjectiveExecutorStepResult {
	return CloneHostOwnedObjectiveExecutorStepResult(r)
}

func (r HostOwnedObjectiveExecutorStepResult) Normalize() HostOwnedObjectiveExecutorStepResult {
	out := CloneHostOwnedObjectiveExecutorStepResult(r)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.Request = out.Request.Normalize()
	out.ExecutorStepRef = normalizeOneDisplaySafeRef(out.ExecutorStepRef)
	out.ExecutorStepResultRef = normalizeOneDisplaySafeRef(out.ExecutorStepResultRef)
	out.HostExecutorRef = normalizeOneDisplaySafeRef(out.HostExecutorRef)
	out.HostExecutorRunRef = normalizeOneDisplaySafeRef(out.HostExecutorRunRef)
	out.ExpectedAttemptRef = normalizeOneAttemptRef(out.ExpectedAttemptRef)
	out.Attempt = out.Attempt.Normalize()
	out.ExpectedResultRef = normalizeOneDisplaySafeRef(out.ExpectedResultRef)
	out.ExpectedReadbackRef = normalizeOneDisplaySafeRef(out.ExpectedReadbackRef)
	out.Observations = normalizeObservations(out.Observations)
	out.AttemptLedgerPatch = out.AttemptLedgerPatch.Normalize()
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
	out = hostOwnedObjectiveExecutorNormalizeEffects(out)
	out.ReadyForExecutorReadback = out.Status == HostActionRecorded && len(out.MissingInputs) == 0 && len(out.BlockedReasons) == 0 && !out.RawOutputLoaded
	out.ReadyForObservationNormalization = out.ReadyForExecutorReadback && len(out.Observations) > 0
	if out.RawOutputLoaded {
		out.Status = HostActionReviewRequired
		out.ReadyForExecutorReadback = false
		out.ReadyForObservationNormalization = false
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

func CloneHostOwnedObjectiveExecutorStepReadback(in HostOwnedObjectiveExecutorStepReadback) HostOwnedObjectiveExecutorStepReadback {
	out := in
	out.Request = in.Request.Clone()
	out.Result = in.Result.Clone()
	out.AttemptLedgerPatch = in.AttemptLedgerPatch.Clone()
	out.Observations = cloneObservations(in.Observations)
	out.Verification = in.Verification.Clone()
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (r HostOwnedObjectiveExecutorStepReadback) Clone() HostOwnedObjectiveExecutorStepReadback {
	return CloneHostOwnedObjectiveExecutorStepReadback(r)
}

func (r HostOwnedObjectiveExecutorStepReadback) Normalize() HostOwnedObjectiveExecutorStepReadback {
	out := CloneHostOwnedObjectiveExecutorStepReadback(r)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.Request = out.Request.Normalize()
	out.Result = out.Result.Normalize()
	out.ExecutorStepRef = normalizeOneDisplaySafeRef(out.ExecutorStepRef)
	out.ExecutorStepResultRef = normalizeOneDisplaySafeRef(out.ExecutorStepResultRef)
	out.ExecutorStepReadbackRef = normalizeOneDisplaySafeRef(out.ExecutorStepReadbackRef)
	out.ExpectedReadbackRef = normalizeOneDisplaySafeRef(out.ExpectedReadbackRef)
	out.ObservedResultRef = normalizeOneDisplaySafeRef(out.ObservedResultRef)
	out.ObservedAttemptRef = normalizeOneAttemptRef(out.ObservedAttemptRef)
	out.AttemptLedgerPatch = out.AttemptLedgerPatch.Normalize()
	out.Observations = normalizeObservations(out.Observations)
	out.Verification = out.Verification.Normalize()
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
	out = hostOwnedObjectiveExecutorNormalizeEffects(out)
	out.ReadyForRuntimeLoopInput = out.Status == HostActionRecorded && len(out.MissingInputs) == 0 && len(out.BlockedReasons) == 0 && !out.RawOutputLoaded
	if out.RawOutputLoaded {
		out.Status = HostActionReviewRequired
		out.ReadyForRuntimeLoopInput = false
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

func hostOwnedObjectiveExecutorRun(input HostOwnedObjectiveExecutorStepRequestInput) ObjectiveRun {
	if !objectiveRunInputEmpty(input.RuntimeLoop.Run) {
		return input.RuntimeLoop.Run.Normalize()
	}
	if !objectiveRunInputEmpty(input.Run) {
		return input.Run.Normalize()
	}
	return ObjectiveRun{}
}

func hostOwnedObjectiveExecutorDecision(input HostOwnedObjectiveExecutorStepRequestInput, run ObjectiveRun) ObjectiveControllerDecision {
	if !hostOwnedObjectiveExecutorDecisionEmpty(input.RuntimeLoop.ControllerDecision) {
		return input.RuntimeLoop.ControllerDecision.Normalize()
	}
	if !hostOwnedObjectiveExecutorDecisionEmpty(input.ControllerDecision) {
		return input.ControllerDecision.Normalize()
	}
	if !objectiveRunInputEmpty(run) {
		return BuildObjectiveControllerDecision(ObjectiveControllerInput{Run: run})
	}
	return ObjectiveControllerDecision{}
}

func hostOwnedObjectiveExecutorRuntimeLoopReady(step ObjectiveRuntimeLoopStep, decision ObjectiveControllerDecision) bool {
	if hostOwnedObjectiveExecutorRuntimeLoopEmpty(step) {
		return true
	}
	return step.Normalize().ReadyForNextRuntimeAction && step.ControllerDecision.Normalize().Action == decision.Normalize().Action
}

func hostOwnedObjectiveExecutorActionExecutable(action ObjectiveControllerAction) bool {
	switch NormalizeObjectiveControllerAction(string(action)) {
	case ObjectiveActionPlanStrategy, ObjectiveActionRequestReplanDecision:
		return true
	default:
		return false
	}
}

func hostOwnedObjectiveExecutorStrategyRef(strategy StrategyCandidate) DisplaySafeRef {
	ref, _ := NormalizeDisplaySafeRef(strategy.Normalize().ID)
	return ref
}

func hostOwnedObjectiveExecutorStrategyAllowed(ref DisplaySafeRef, groups ...[]StrategyCandidate) bool {
	ref = normalizeOneDisplaySafeRef(ref)
	if ref == "" {
		return false
	}
	for _, group := range groups {
		for _, candidate := range normalizeStrategyCandidates(group) {
			if hostOwnedObjectiveExecutorStrategyRef(candidate) == ref {
				return true
			}
		}
	}
	return false
}

func hostOwnedObjectiveExecutorLedgerPatch(request HostOwnedObjectiveExecutorStepRequest, attempt AttemptSummary, evidenceRefs []EvidenceRef) AttemptLedgerPatch {
	return AttemptLedgerPatch{
		ObjectiveID:  request.Frame.ID,
		LedgerRef:    request.Run.Ledger.LedgerRef,
		Attempts:     []AttemptSummary{attempt},
		EvidenceRefs: MergeEvidenceRefs(evidenceRefs, attempt.EvidenceRefs),
		Boundaries: AppendBoundaries(
			request.Boundaries,
			"host_owned_executor_step_result",
			"host_must_persist_objective_run",
			"core_runtime_loop_does_not_write_store",
		),
		NextHostAction: "run_objective_runtime_loop_step",
	}.Normalize()
}

func hostOwnedObjectiveExecutorRequestBlock(result HostOwnedObjectiveExecutorStepRequest, failure FailureClass, reason string, missing MissingInput, next NextHostAction) HostOwnedObjectiveExecutorStepRequest {
	result.Status = HostActionBlocked
	result.ReadyForHostExecution = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	return result
}

func hostOwnedObjectiveExecutorResultBlock(result HostOwnedObjectiveExecutorStepResult, failure FailureClass, reason string, missing MissingInput, next NextHostAction) HostOwnedObjectiveExecutorStepResult {
	result.Status = HostActionBlocked
	result.ReadyForExecutorReadback = false
	result.ReadyForObservationNormalization = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	return result
}

func hostOwnedObjectiveExecutorReadbackBlock(result HostOwnedObjectiveExecutorStepReadback, failure FailureClass, reason string, missing MissingInput, next NextHostAction) HostOwnedObjectiveExecutorStepReadback {
	result.Status = HostActionBlocked
	result.ReadyForRuntimeLoopInput = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	return result
}

func hostOwnedObjectiveExecutorRequestUnsafe(input HostOwnedObjectiveExecutorStepRequestInput, run ObjectiveRun, decision ObjectiveControllerDecision, strategy StrategyCandidate, gate IntensityGateResult) bool {
	return input.RawOutputLoaded ||
		displaySafeRefRejected(input.HostExecutorRef) ||
		displaySafeRefRejected(input.ExecutorStepRef) ||
		displaySafeRefRejected(input.ExecutionContractRef) ||
		displaySafeRefRejected(input.IdempotencyRef) ||
		displaySafeRefRejected(DisplaySafeRef(input.ExpectedAttemptRef)) ||
		displaySafeRefRejected(input.ExpectedResultRef) ||
		displaySafeRefRejected(input.ExpectedReadbackRef) ||
		displaySafeRefSliceRejected(input.ApprovalRefs) ||
		displaySafeRefSliceRejected(input.PolicyRefs) ||
		displaySafeRefSliceRejected(input.DecisionBasis) ||
		evidenceRefRejected(input.EvidenceRefs) ||
		run.RawOutputLoaded ||
		decision.RawOutputLoaded ||
		gate.RawOutputLoaded ||
		(strategy.ID != "" && displaySafeRefRejected(DisplaySafeRef(strategy.ID))) ||
		evidenceRefRejected(strategy.ExpectedEvidence)
}

func hostOwnedObjectiveExecutorResultUnsafe(input HostOwnedObjectiveExecutorStepResultInput, request HostOwnedObjectiveExecutorStepRequest, attempt AttemptSummary, observations []Observation) bool {
	return input.RawOutputLoaded ||
		request.RawOutputLoaded ||
		displaySafeRefRejected(input.ExecutorStepResultRef) ||
		displaySafeRefRejected(input.HostExecutorRunRef) ||
		displaySafeRefRejected(DisplaySafeRef(attempt.Ref)) ||
		displaySafeRefRejected(DisplaySafeRef(attempt.StrategyID)) ||
		evidenceRefRejected(input.EvidenceRefs) ||
		evidenceRefRejected(attempt.EvidenceRefs) ||
		observationSliceUnsafePayload(observations) ||
		attempt.RawOutputLoaded
}

func hostOwnedObjectiveExecutorReadbackUnsafe(input HostOwnedObjectiveExecutorStepReadbackInput, request HostOwnedObjectiveExecutorStepRequest, result HostOwnedObjectiveExecutorStepResult, verification ObjectiveVerificationGateResult) bool {
	return input.RawOutputLoaded ||
		request.RawOutputLoaded ||
		result.RawOutputLoaded ||
		verification.RawOutputLoaded ||
		displaySafeRefRejected(input.ExecutorStepReadbackRef) ||
		displaySafeRefRejected(input.ObservedResultRef) ||
		displaySafeRefRejected(DisplaySafeRef(input.ObservedAttemptRef)) ||
		evidenceRefRejected(input.EvidenceRefs) ||
		evidenceRefRejected(verification.EvidenceRefs) ||
		observationSliceUnsafePayload(verification.Observations)
}

func hostOwnedObjectiveExecutorBoundaries(extra []Boundary) []Boundary {
	return MergeBoundaries(
		[]Boundary{
			"host_owned_objective_executor_loop",
			"host_owned_executor_step_request",
			"request_only",
			"host_must_execute_objective_step",
			"host_must_report_step_result",
			"host_must_bind_step_readback",
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
		},
		extra,
	)
}

func hostOwnedObjectiveExecutorResultBoundaries(groups ...[]Boundary) []Boundary {
	all := append([][]Boundary{{
		"host_owned_objective_executor_loop",
		"host_owned_executor_step_result",
		"host_execution_report_only",
		"host_result_not_objective_completion",
		"objective_completion_requires_verification_gate",
		"no_core_execution",
		"no_runner_dispatch_by_core",
		"no_runtime_adapter_execution_by_core",
		"no_scheduler_apply",
		"no_install_apply",
		"no_worker_dispatch",
		"no_store_mutation_by_core",
		"no_compensation_execution",
		"projection_only",
	}}, groups...)
	return MergeBoundaries(all...)
}

func hostOwnedObjectiveExecutorReadbackBoundaries(groups ...[]Boundary) []Boundary {
	all := append([][]Boundary{{
		"host_owned_objective_executor_loop",
		"host_owned_executor_step_readback",
		"executor_readback_projection_only",
		"ready_readback_not_runtime_execution",
		"host_must_persist_objective_run",
		"no_core_execution",
		"no_runner_dispatch_by_core",
		"no_runtime_adapter_execution_by_core",
		"no_scheduler_apply",
		"no_install_apply",
		"no_worker_dispatch",
		"no_store_mutation_by_core",
		"no_compensation_execution",
		"projection_only",
	}}, groups...)
	return MergeBoundaries(all...)
}

func hostOwnedObjectiveExecutorDecisionEmpty(decision ObjectiveControllerDecision) bool {
	return decision.ContractVersion == "" &&
		!decision.Projected &&
		decision.Activation == "" &&
		decision.State == "" &&
		decision.Action == "" &&
		decision.ObjectiveID == "" &&
		decision.CurrentStrategyRef == "" &&
		len(decision.AllowedStrategies) == 0 &&
		decision.Verification.Status == "" &&
		len(decision.EvidenceRefs) == 0 &&
		decision.FailureClass == "" &&
		decision.Reason == "" &&
		len(decision.MissingInputs) == 0 &&
		len(decision.DecisionBasis) == 0 &&
		len(decision.Boundaries) == 0 &&
		decision.NextHostAction == "" &&
		!decision.RawOutputLoaded
}

func hostOwnedObjectiveExecutorRuntimeLoopEmpty(step ObjectiveRuntimeLoopStep) bool {
	return step.ContractVersion == "" &&
		!step.Projected &&
		!step.Available &&
		step.Status == "" &&
		step.Run.Frame.ID == "" &&
		hostOwnedObjectiveExecutorDecisionEmpty(step.ControllerDecision) &&
		len(step.MissingInputs) == 0 &&
		len(step.Boundaries) == 0 &&
		!step.RawOutputLoaded
}

func normalizeOneAttemptRef(value AttemptRef) AttemptRef {
	ref, _ := NormalizeAttemptRef(string(value))
	return ref
}

type hostOwnedObjectiveExecutorEffectCarrier interface {
	HostOwnedObjectiveExecutorStepRequest | HostOwnedObjectiveExecutorStepResult | HostOwnedObjectiveExecutorStepReadback
}

func hostOwnedObjectiveExecutorNormalizeEffects[T hostOwnedObjectiveExecutorEffectCarrier](in T) T {
	switch value := any(in).(type) {
	case HostOwnedObjectiveExecutorStepRequest:
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
		value.HostExecutionReported = false
		value.CoreExecutionExecuted = false
		value.RunnerDispatched = false
		value.RuntimeAdapterExecuted = false
		value.SchedulerApplied = false
		value.InstallerExecuted = false
		value.WorkflowDispatched = false
		value.WorkerDispatched = false
		value.StoreMutationExecuted = false
		value.CompensationExecuted = false
		return any(value).(T)
	case HostOwnedObjectiveExecutorStepResult:
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
		return any(value).(T)
	case HostOwnedObjectiveExecutorStepReadback:
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
		return any(value).(T)
	default:
		return in
	}
}
