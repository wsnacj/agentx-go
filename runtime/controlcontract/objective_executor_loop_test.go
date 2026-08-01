package controlcontract

import "testing"

func TestHostOwnedObjectiveExecutorLoopBuildsRequestResultReadback(t *testing.T) {
	step := hostOwnedObjectiveExecutorReadyRuntimeStep(t)
	strategy := step.Run.Strategies[0]
	request := BuildHostOwnedObjectiveExecutorStepRequest(hostOwnedObjectiveExecutorRequestInput(step, strategy))
	if request.Status != HostActionReady ||
		!request.ReadyForHostExecution ||
		!request.RequestOnly ||
		request.HostExecutionReported ||
		request.CoreExecutionExecuted ||
		request.RunnerDispatched ||
		request.RuntimeAdapterExecuted ||
		request.SchedulerApplied ||
		request.InstallerExecuted ||
		request.WorkflowDispatched ||
		request.WorkerDispatched ||
		request.StoreMutationExecuted ||
		request.CompensationExecuted ||
		request.NextHostAction != "host_may_execute_objective_step" {
		t.Fatalf("unexpected executor request: %#v", request)
	}
	for _, boundary := range []Boundary{
		"host_owned_objective_executor_loop",
		"host_owned_executor_step_request",
		"request_only",
		"host_must_execute_objective_step",
		"no_core_execution",
		"no_runner_dispatch",
		"no_runtime_adapter_execution",
		"no_scheduler_apply",
		"no_install_apply",
		"no_worker_dispatch",
		"no_store_mutation_by_core",
		"no_compensation_execution",
	} {
		if !objectiveLoopBoundaryContains(request.Boundaries, boundary) {
			t.Fatalf("request missing boundary %q in %#v", boundary, request.Boundaries)
		}
	}

	result := BuildHostOwnedObjectiveExecutorStepResult(HostOwnedObjectiveExecutorStepResultInput{
		Request:               request,
		ExecutorStepResultRef: request.ExpectedResultRef,
		HostExecutorRunRef:    "executor_run:objective_step_1",
		Attempt:               hostOwnedObjectiveExecutorAttempt(VerificationPartial, FailureEvidenceMissing),
		Observations:          hostOwnedObjectiveExecutorObservations(),
		EvidenceRefs: []EvidenceRef{{
			Ref:      "evidence:executor_step",
			Kind:     "metric",
			Strength: EvidenceStrong,
			Source:   "executor:objective",
		}},
		HostExecutionReported: true,
	})
	if result.Status != HostActionRecorded ||
		!result.ReadyForExecutorReadback ||
		!result.ReadyForObservationNormalization ||
		result.AttemptLedgerPatch.LedgerRef != request.Run.Ledger.LedgerRef ||
		len(result.AttemptLedgerPatch.Attempts) != 1 ||
		result.AttemptLedgerPatch.Attempts[0].Ref != request.ExpectedAttemptRef ||
		result.CoreExecutionExecuted ||
		result.RunnerDispatched ||
		result.RuntimeAdapterExecuted ||
		result.SchedulerApplied ||
		result.InstallerExecuted ||
		result.WorkflowDispatched ||
		result.WorkerDispatched ||
		result.StoreMutationExecuted ||
		result.CompensationExecuted {
		t.Fatalf("unexpected executor result: %#v", result)
	}

	verification := objectiveRuntimeLoopVerificationGate(VerificationPartial, FailureEvidenceMissing, "request_replan_or_return_partial")
	verification.Observations = result.Observations
	verification.EvidenceRefs = result.EvidenceRefs
	readback := BuildHostOwnedObjectiveExecutorStepReadback(HostOwnedObjectiveExecutorStepReadbackInput{
		Request:                 request,
		Result:                  result,
		ExecutorStepReadbackRef: request.ExpectedReadbackRef,
		ObservedResultRef:       result.ExecutorStepResultRef,
		ObservedAttemptRef:      result.Attempt.Ref,
		Verification:            verification,
	})
	if readback.Status != HostActionRecorded ||
		!readback.ReadyForRuntimeLoopInput ||
		readback.NextHostAction != "run_objective_runtime_loop_step" ||
		readback.CoreExecutionExecuted ||
		readback.RunnerDispatched ||
		readback.RuntimeAdapterExecuted ||
		readback.SchedulerApplied ||
		readback.InstallerExecuted ||
		readback.WorkflowDispatched ||
		readback.WorkerDispatched ||
		readback.StoreMutationExecuted ||
		readback.CompensationExecuted {
		t.Fatalf("unexpected executor readback: %#v", readback)
	}

	nextStep := BuildObjectiveRuntimeLoopStep(ObjectiveRuntimeLoopInput{
		Run:          request.Run,
		LedgerPatch:  readback.AttemptLedgerPatch,
		Verification: readback.Verification,
		Observations: readback.Observations,
		EvidenceRefs: readback.EvidenceRefs,
	})
	if nextStep.Status != "ready_for_host_persist" ||
		!nextStep.ReadyForNextRuntimeAction ||
		nextStep.ControllerDecision.Action != ObjectiveActionRequestReplanDecision {
		t.Fatalf("executor readback should enter runtime loop, got %#v", nextStep)
	}
}

func TestHostOwnedObjectiveExecutorLoopBlocksNonExecutableControllerAction(t *testing.T) {
	input := objectiveRuntimeLoopReadyInput()
	input.LedgerPatch = objectiveRuntimeLoopLedgerPatch(VerificationSatisfied, FailureNone)
	input.Verification = objectiveRuntimeLoopVerificationGate(VerificationSatisfied, FailureNone, "return_satisfied")
	step := BuildObjectiveRuntimeLoopStep(input)
	request := BuildHostOwnedObjectiveExecutorStepRequest(hostOwnedObjectiveExecutorRequestInput(step, step.Run.Strategies[0]))
	if request.Status != HostActionBlocked ||
		request.ReadyForHostExecution ||
		!objectiveLoopStringContains(request.BlockedReasons, "runtime_loop_not_ready_for_executor_step") ||
		!objectiveLoopStringContains(request.BlockedReasons, "controller_action_not_executable") ||
		request.CoreExecutionExecuted ||
		request.RunnerDispatched ||
		request.RuntimeAdapterExecuted {
		t.Fatalf("satisfied action should not create executor request: %#v", request)
	}
}

func TestHostOwnedObjectiveExecutorLoopRequiresFinalGateAndReadbackBinding(t *testing.T) {
	step := hostOwnedObjectiveExecutorReadyRuntimeStep(t)
	strategy := step.Run.Strategies[0]
	requestInput := hostOwnedObjectiveExecutorRequestInput(step, strategy)
	requestInput.FinalGate.Status = VerificationBlocked
	requestInput.FinalGate.Allowed = false
	requestInput.FinalGate.FailureClass = FailurePolicyBlocked
	blocked := BuildHostOwnedObjectiveExecutorStepRequest(requestInput)
	if blocked.Status != HostActionBlocked ||
		blocked.ReadyForHostExecution ||
		!objectiveLoopStringContains(blocked.BlockedReasons, "final_gate_not_satisfied") {
		t.Fatalf("final gate blocked request = %#v", blocked)
	}

	request := BuildHostOwnedObjectiveExecutorStepRequest(hostOwnedObjectiveExecutorRequestInput(step, strategy))
	result := BuildHostOwnedObjectiveExecutorStepResult(HostOwnedObjectiveExecutorStepResultInput{
		Request:               request,
		ExecutorStepResultRef: request.ExpectedResultRef,
		HostExecutorRunRef:    "executor_run:objective_step_1",
		Attempt:               hostOwnedObjectiveExecutorAttempt(VerificationPartial, FailureEvidenceMissing),
		Observations:          hostOwnedObjectiveExecutorObservations(),
		HostExecutionReported: true,
	})
	readback := BuildHostOwnedObjectiveExecutorStepReadback(HostOwnedObjectiveExecutorStepReadbackInput{
		Request:                 request,
		Result:                  result,
		ExecutorStepReadbackRef: "executor_readback:wrong",
		ObservedResultRef:       result.ExecutorStepResultRef,
		ObservedAttemptRef:      result.Attempt.Ref,
		Verification:            objectiveRuntimeLoopVerificationGate(VerificationPartial, FailureEvidenceMissing, "request_replan_or_return_partial"),
	})
	if readback.Status != HostActionBlocked ||
		readback.ReadyForRuntimeLoopInput ||
		!objectiveLoopStringContains(readback.BlockedReasons, "executor_step_readback_ref_mismatch") {
		t.Fatalf("readback mismatch should block: %#v", readback)
	}
}

func TestHostOwnedObjectiveExecutorLoopRawOutputForcesReview(t *testing.T) {
	step := hostOwnedObjectiveExecutorReadyRuntimeStep(t)
	strategy := step.Run.Strategies[0]
	input := hostOwnedObjectiveExecutorRequestInput(step, strategy)
	input.ExecutorStepRef = "/Users/example/raw-executor-step"
	request := BuildHostOwnedObjectiveExecutorStepRequest(input)
	if request.Status != HostActionReviewRequired ||
		request.ReadyForHostExecution ||
		request.FailureClass != FailureEvidenceWeak ||
		!objectiveLoopMissingInputContains(request.MissingInputs, "host:display_safe_refs") ||
		!objectiveLoopBoundaryContains(request.Boundaries, "raw_output_not_allowed") {
		t.Fatalf("unsafe executor request should force review: %#v", request)
	}
}

func hostOwnedObjectiveExecutorReadyRuntimeStep(t *testing.T) ObjectiveRuntimeLoopStep {
	t.Helper()
	input := objectiveRuntimeLoopReadyInput()
	input.LedgerPatch = objectiveRuntimeLoopLedgerPatch(VerificationPartial, FailureEvidenceMissing)
	input.Verification = objectiveRuntimeLoopVerificationGate(VerificationPartial, FailureEvidenceMissing, "request_replan_or_return_partial")
	step := BuildObjectiveRuntimeLoopStep(input)
	if !step.ReadyForNextRuntimeAction ||
		step.ControllerDecision.Action != ObjectiveActionRequestReplanDecision {
		t.Fatalf("expected ready runtime action step, got %#v", step)
	}
	return step
}

func hostOwnedObjectiveExecutorRequestInput(step ObjectiveRuntimeLoopStep, strategy StrategyCandidate) HostOwnedObjectiveExecutorStepRequestInput {
	return HostOwnedObjectiveExecutorStepRequestInput{
		RuntimeLoop:          step,
		SelectedStrategy:     strategy,
		FinalGate:            hostOwnedObjectiveExecutorFinalGate(strategy, step.Run.Budget),
		HostExecutorRef:      "executor:host_objective",
		ExecutorStepRef:      "executor_step:objective_step_1",
		ExecutionContractRef: "contract:objective_executor",
		IdempotencyRef:       "idempotency:objective_step_1",
		ExpectedAttemptRef:   "attempt:executor_step_1",
		ExpectedResultRef:    "executor_result:objective_step_1",
		ExpectedReadbackRef:  "executor_readback:objective_step_1",
	}
}

func hostOwnedObjectiveExecutorFinalGate(strategy StrategyCandidate, budget ObjectiveBudgetSnapshot) IntensityGateResult {
	return IntensityGateResult{
		Stage:                IntensityGateFinal,
		Activation:           ActivationManaged,
		Status:               VerificationSatisfied,
		Allowed:              true,
		RequestedControlMode: strategy.ControlMode,
		ApprovedControlMode:  strategy.ControlMode,
		RequestedIntensity:   strategy.MinIntensity,
		ApprovedIntensity:    strategy.MinIntensity,
		MaxAllowedIntensity:  IntensityL3ManagedObjective,
		StrategyRef:          DisplaySafeRef(strategy.ID),
		HostApproved:         true,
		Budget:               budget,
		PolicyRefs:           []DisplaySafeRef{"policy:objective_executor"},
		NextHostAction:       "host_may_execute_objective_step",
		RunnerEffect:         "none",
		PromptEffect:         "none",
	}
}

func hostOwnedObjectiveExecutorAttempt(status VerificationStatus, failure FailureClass) AttemptSummary {
	return AttemptSummary{
		Ref:              "attempt:executor_step_1",
		ObjectiveID:      "objective:runtime_loop",
		StrategyID:       "strategy:runtime_loop",
		Index:            2,
		ControlMode:      ControlModeObjective,
		Intensity:        IntensityL3ManagedObjective,
		Status:           status,
		ObservationCount: 1,
		EvidenceRefs: []EvidenceRef{{
			Ref:      "evidence:executor_step",
			Kind:     "metric",
			Strength: EvidenceStrong,
			Source:   "executor:objective",
		}},
		FailureClass: failure,
	}
}

func hostOwnedObjectiveExecutorObservations() []Observation {
	return []Observation{{
		Kind:     "metric",
		Source:   "executor:objective",
		Subject:  "objective:runtime_loop",
		Name:     "progress",
		Value:    "partial",
		Strength: EvidenceStrong,
		EvidenceRefs: []EvidenceRef{{
			Ref:      "evidence:executor_step",
			Kind:     "metric",
			Strength: EvidenceStrong,
			Source:   "executor:objective",
		}},
	}}
}

func objectiveLoopStringContains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
