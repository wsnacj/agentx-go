package controlcontract

import (
	"testing"
)

func TestHostOwnedObjectiveExecutorAdapterGateBuildsReadinessAndInvocation(t *testing.T) {
	step := hostOwnedObjectiveExecutorReadyRuntimeStep(t)
	request := BuildHostOwnedObjectiveExecutorStepRequest(hostOwnedObjectiveExecutorRequestInput(step, step.Run.Strategies[0]))
	readiness := BuildHostOwnedObjectiveExecutorAdapterReadiness(hostOwnedObjectiveExecutorAdapterReadinessInput(request))
	if readiness.Status != HostActionReady ||
		!readiness.ReadyForHostAdapterInvocation ||
		!readiness.HostAdapterInvocationAuthorized ||
		!readiness.HostMayInvokeAdapter ||
		readiness.NextHostAction != "host_may_invoke_objective_executor_adapter" ||
		readiness.ResultBindingRef != request.ExpectedResultRef ||
		readiness.ReadbackBindingRef != request.ExpectedReadbackRef ||
		readiness.CoreExecutionExecuted ||
		readiness.RunnerDispatched ||
		readiness.RuntimeAdapterExecuted ||
		readiness.SchedulerApplied ||
		readiness.InstallerExecuted ||
		readiness.WorkflowDispatched ||
		readiness.WorkerDispatched ||
		readiness.StoreMutationExecuted ||
		readiness.CompensationExecuted {
		t.Fatalf("unexpected adapter readiness: %#v", readiness)
	}
	assertHostOwnedProjectionOnly(t, testProjection[Boundary]{
		Name:         "objective executor adapter readiness",
		RunnerEffect: readiness.RunnerEffect,
		PromptEffect: readiness.PromptEffect,
		Boundaries:   readiness.Boundaries,
		Payload:      readiness,
	}, "host_owned_objective_executor_adapter_gate", "objective_executor_adapter_invocation_gate", "explicit_host_confirmation_required", "adapter_invocation_not_executed_by_core", "no_runtime_adapter_execution")

	invocation := BuildHostOwnedObjectiveExecutorAdapterInvocation(hostOwnedObjectiveExecutorAdapterInvocationInput(readiness))
	if invocation.Status != HostActionRecorded ||
		!invocation.ReadyForExecutorStepResult ||
		invocation.ReadyForFailureReview ||
		!invocation.HostInvocationReported ||
		!invocation.HostInvocationCompleted ||
		invocation.HostInvocationFailed ||
		invocation.NextHostAction != "build_objective_executor_step_result" ||
		invocation.ObservedInvocationRef != readiness.InvocationRef ||
		invocation.ExecutorStepResultRef != request.ExpectedResultRef ||
		invocation.ExecutorStepReadbackRef != request.ExpectedReadbackRef ||
		invocation.AttemptRef != request.ExpectedAttemptRef ||
		invocation.CoreExecutionExecuted ||
		invocation.RunnerDispatched ||
		invocation.RuntimeAdapterExecuted ||
		invocation.SchedulerApplied ||
		invocation.InstallerExecuted ||
		invocation.WorkflowDispatched ||
		invocation.WorkerDispatched ||
		invocation.StoreMutationExecuted ||
		invocation.CompensationExecuted {
		t.Fatalf("unexpected adapter invocation: %#v", invocation)
	}
	assertHostOwnedProjectionOnly(t, testProjection[Boundary]{
		Name:         "objective executor adapter invocation",
		RunnerEffect: invocation.RunnerEffect,
		PromptEffect: invocation.PromptEffect,
		Boundaries:   invocation.Boundaries,
		Payload:      invocation,
	}, "host_owned_objective_executor_adapter_gate", "host_adapter_invocation_report_only", "host_invocation_report_not_objective_completion", "objective_completion_requires_verification_gate", "no_runtime_adapter_execution")
}

func TestHostOwnedObjectiveExecutorAdapterGateBlocksMissingConfirmationAndMismatches(t *testing.T) {
	step := hostOwnedObjectiveExecutorReadyRuntimeStep(t)
	request := BuildHostOwnedObjectiveExecutorStepRequest(hostOwnedObjectiveExecutorRequestInput(step, step.Run.Strategies[0]))
	missingInput := hostOwnedObjectiveExecutorAdapterReadinessInput(request)
	missingInput.HostConfirmationRef = ""
	missing := BuildHostOwnedObjectiveExecutorAdapterReadiness(missingInput)
	if missing.Status != HostActionBlocked ||
		missing.ReadyForHostAdapterInvocation ||
		missing.HostMayInvokeAdapter ||
		missing.FailureClass != FailureConfigMissing ||
		!objectiveLoopStringContains(missing.BlockedReasons, "host_confirmation_ref_missing") ||
		!objectiveLoopMissingInputContains(missing.MissingInputs, "host:objective_executor_adapter_confirmation") {
		t.Fatalf("expected missing confirmation to block, got %#v", missing)
	}

	readiness := BuildHostOwnedObjectiveExecutorAdapterReadiness(hostOwnedObjectiveExecutorAdapterReadinessInput(request))
	mismatchInput := hostOwnedObjectiveExecutorAdapterInvocationInput(readiness)
	mismatchInput.AttemptRef = "attempt:wrong"
	mismatchInput.ExecutorStepResultRef = "executor_result:wrong"
	mismatchInput.ExecutorStepReadbackRef = "executor_readback:wrong"
	mismatch := BuildHostOwnedObjectiveExecutorAdapterInvocation(mismatchInput)
	if mismatch.Status != HostActionBlocked ||
		mismatch.ReadyForExecutorStepResult ||
		mismatch.FailureClass != FailureVerificationFailed ||
		!objectiveLoopStringContains(mismatch.BlockedReasons, "attempt_ref_mismatch") ||
		!objectiveLoopStringContains(mismatch.BlockedReasons, "executor_step_result_ref_mismatch") ||
		!objectiveLoopStringContains(mismatch.BlockedReasons, "executor_step_readback_ref_mismatch") {
		t.Fatalf("expected invocation mismatch to block, got %#v", mismatch)
	}
}

func TestHostOwnedObjectiveExecutorAdapterGateFailureReportAndUnsafeRefs(t *testing.T) {
	step := hostOwnedObjectiveExecutorReadyRuntimeStep(t)
	request := BuildHostOwnedObjectiveExecutorStepRequest(hostOwnedObjectiveExecutorRequestInput(step, step.Run.Strategies[0]))
	readiness := BuildHostOwnedObjectiveExecutorAdapterReadiness(hostOwnedObjectiveExecutorAdapterReadinessInput(request))
	failedInput := hostOwnedObjectiveExecutorAdapterInvocationInput(readiness)
	failedInput.HostInvocationCompleted = false
	failedInput.HostInvocationFailed = true
	failedInput.FailureRef = "failure:objective_executor_adapter"
	failedInput.FailureClass = FailureVerificationFailed
	failed := BuildHostOwnedObjectiveExecutorAdapterInvocation(failedInput)
	if failed.Status != HostActionRecorded ||
		failed.ReadyForExecutorStepResult ||
		!failed.ReadyForFailureReview ||
		failed.NextHostAction != "review_objective_executor_adapter_failure" ||
		failed.FailureClass != FailureVerificationFailed ||
		!objectiveLoopBoundaryContains(failed.Boundaries, "host_owned_objective_executor_adapter_failure_recorded") {
		t.Fatalf("unexpected failed invocation report: %#v", failed)
	}

	unsafeInput := hostOwnedObjectiveExecutorAdapterReadinessInput(request)
	unsafeInput.AdapterRef = "/Users/example/raw-adapter"
	unsafe := BuildHostOwnedObjectiveExecutorAdapterReadiness(unsafeInput)
	if unsafe.Status != HostActionReviewRequired ||
		unsafe.ReadyForHostAdapterInvocation ||
		unsafe.FailureClass != FailureEvidenceWeak ||
		!objectiveLoopStringContains(unsafe.BlockedReasons, "unsafe_input_ref") ||
		!objectiveLoopMissingInputContains(unsafe.MissingInputs, "host:display_safe_refs") ||
		!objectiveLoopBoundaryContains(unsafe.Boundaries, "raw_output_not_allowed") {
		t.Fatalf("expected unsafe adapter ref to force review, got %#v", unsafe)
	}
	assertNoRawPayload(t, "objective executor adapter readiness", unsafe, "/Users/example/raw-adapter")
}

func hostOwnedObjectiveExecutorAdapterReadinessInput(request HostOwnedObjectiveExecutorStepRequest) HostOwnedObjectiveExecutorAdapterReadinessInput {
	return HostOwnedObjectiveExecutorAdapterReadinessInput{
		Request:              request,
		AdapterRef:           "adapter:host_objective_executor",
		AdapterVersionRef:    "adapter_version:host_objective_executor_v1",
		AdapterCapabilityRef: "capability:objective_executor_host_adapter",
		AdapterContractRef:   "contract:objective_executor_adapter",
		HostConfirmationRef:  "confirmation:objective_executor_adapter",
		InvocationRef:        "invocation:objective_executor_step_1",
		ResultBindingRef:     request.ExpectedResultRef,
		ReadbackBindingRef:   request.ExpectedReadbackRef,
		CancellationRef:      "cancel:objective_executor_step_1",
		PolicyRefs:           []DisplaySafeRef{"policy:objective_executor_adapter"},
		EvidenceRefs: []EvidenceRef{{
			Ref:      "evidence:objective_executor_adapter_binding",
			Kind:     "adapter_binding",
			Strength: EvidenceStrong,
			Source:   "host:objective_executor",
		}},
	}
}

func hostOwnedObjectiveExecutorAdapterInvocationInput(readiness HostOwnedObjectiveExecutorAdapterReadiness) HostOwnedObjectiveExecutorAdapterInvocationInput {
	return HostOwnedObjectiveExecutorAdapterInvocationInput{
		Readiness:               readiness,
		InvocationReportRef:     "invocation_report:objective_executor_step_1",
		ObservedInvocationRef:   readiness.InvocationRef,
		HostAdapterRunRef:       "adapter_run:objective_executor_step_1",
		ExecutorStepResultRef:   readiness.ExpectedResultRef,
		ExecutorStepReadbackRef: readiness.ExpectedReadbackRef,
		AttemptRef:              readiness.ExpectedAttemptRef,
		HostInvocationReported:  true,
		HostInvocationCompleted: true,
		EvidenceRefs: []EvidenceRef{{
			Ref:      "evidence:objective_executor_adapter_invocation",
			Kind:     "adapter_invocation",
			Strength: EvidenceStrong,
			Source:   "host:objective_executor",
		}},
	}
}
