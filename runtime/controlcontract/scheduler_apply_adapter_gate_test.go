package controlcontract

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestHostOwnedSchedulerApplyAdapterGateBuildsReadinessAndInvocation(t *testing.T) {
	request := schedulerApplyReadyRequest()
	readiness := BuildHostOwnedSchedulerApplyAdapterReadiness(hostOwnedSchedulerApplyAdapterReadinessInput(request))
	if readiness.Status != HostActionReady ||
		!readiness.ReadyForHostSchedulerAdapterInvocation ||
		!readiness.HostSchedulerAdapterInvocationAuthorized ||
		!readiness.HostMayInvokeSchedulerAdapter ||
		readiness.NextHostAction != "host_may_invoke_scheduler_apply_adapter" ||
		readiness.AdapterRef != request.SchedulerAdapterRef ||
		readiness.AdapterDryRunRef != request.ScheduleDryRunProofRef ||
		readiness.ResultBindingRef != request.ExpectedSchedulerResultRef ||
		readiness.ReadbackBindingRef != request.ExpectedReadbackRef ||
		readiness.IdempotencyRef != request.IdempotencyRef ||
		readiness.CancellationRef != request.CancelPathRef ||
		readiness.DeleteRef != request.DeletePathRef ||
		readiness.DisableRef != request.DisablePathRef ||
		readiness.CoreInvocationExecuted ||
		readiness.SchedulerMutationByCore ||
		readiness.CoreScheduleCreated ||
		readiness.AutomationCreatedByCore ||
		readiness.RunnerDispatched ||
		readiness.RuntimeAdapterExecuted ||
		readiness.InstallerExecuted ||
		readiness.WorkflowDispatched ||
		readiness.WorkerDispatched ||
		readiness.StoreMutationExecuted ||
		readiness.CompensationExecuted {
		t.Fatalf("unexpected scheduler apply adapter readiness: %#v", readiness)
	}
	assertHostOwnedProjectionOnly(t, testProjection[Boundary]{
		Name:         "scheduler apply adapter readiness",
		RunnerEffect: readiness.RunnerEffect,
		PromptEffect: readiness.PromptEffect,
		Boundaries:   readiness.Boundaries,
		Payload:      readiness,
	}, "host_owned_scheduler_apply_adapter_gate", "scheduler_apply_adapter_invocation_gate", "explicit_host_confirmation_required", "scheduler_apply_adapter_dry_run_required", "no_scheduler_mutation_by_core", "no_core_schedule_created")

	invocation := BuildHostOwnedSchedulerApplyAdapterInvocation(hostOwnedSchedulerApplyAdapterInvocationInput(readiness))
	if invocation.Status != HostActionRecorded ||
		!invocation.ReadyForSchedulerApplyResult ||
		invocation.ReadyForFailureReview ||
		!invocation.HostAdapterInvocationReported ||
		!invocation.HostAdapterInvocationCompleted ||
		invocation.HostAdapterInvocationFailed ||
		invocation.NextHostAction != "build_scheduler_apply_result" ||
		invocation.ObservedInvocationRef != readiness.InvocationRef ||
		invocation.SchedulerApplyResultRef != request.ExpectedSchedulerResultRef ||
		invocation.SchedulerReadbackRef != request.ExpectedReadbackRef ||
		invocation.AppliedScheduleRef != request.ExpectedScheduleRef ||
		invocation.AppliedLifecycleStateRef != request.ExpectedLifecycleStateRef ||
		invocation.CoreInvocationExecuted ||
		invocation.SchedulerMutationByCore ||
		invocation.CoreScheduleCreated ||
		invocation.AutomationCreatedByCore ||
		invocation.RunnerDispatched ||
		invocation.RuntimeAdapterExecuted ||
		invocation.InstallerExecuted ||
		invocation.WorkflowDispatched ||
		invocation.WorkerDispatched ||
		invocation.StoreMutationExecuted ||
		invocation.CompensationExecuted {
		t.Fatalf("unexpected scheduler apply adapter invocation: %#v", invocation)
	}
	assertHostOwnedProjectionOnly(t, testProjection[Boundary]{
		Name:         "scheduler apply adapter invocation",
		RunnerEffect: invocation.RunnerEffect,
		PromptEffect: invocation.PromptEffect,
		Boundaries:   invocation.Boundaries,
		Payload:      invocation,
	}, "host_owned_scheduler_apply_adapter_gate", "host_scheduler_adapter_invocation_report_only", "host_adapter_scheduler_mutation_reported_only", "scheduler_apply_result_requires_readback", "no_scheduler_mutation_by_core", "no_core_schedule_created")

	applyResult := BuildHostOwnedSchedulerApplyResult(HostOwnedSchedulerApplyResultInput{
		Request:                     request,
		SchedulerApplyResultRef:     invocation.SchedulerApplyResultRef,
		HostSchedulerRunRef:         invocation.HostSchedulerAdapterRunRef,
		HostSchedulerApplyReported:  true,
		HostSchedulerApplySucceeded: true,
		AppliedScheduleRef:          invocation.AppliedScheduleRef,
		AppliedLifecycleStateRef:    invocation.AppliedLifecycleStateRef,
		SchedulerEvidenceRefs:       invocation.SchedulerEvidenceRefs,
	})
	if applyResult.Status != HostActionRecorded ||
		!applyResult.ReadyForSchedulerReadback ||
		!applyResult.HostScheduleCreated ||
		applyResult.CoreInvocationExecuted ||
		applyResult.SchedulerMutationByCore ||
		applyResult.CoreScheduleCreated ||
		applyResult.AutomationCreatedByCore {
		t.Fatalf("adapter invocation should feed scheduler apply result, got %#v", applyResult)
	}
	readback := BuildHostOwnedSchedulerApplyReadback(HostOwnedSchedulerApplyReadbackInput{
		Result:                    applyResult,
		SchedulerReadbackRef:      invocation.SchedulerReadbackRef,
		ObservedScheduleRef:       applyResult.AppliedScheduleRef,
		ObservedLifecycleStateRef: applyResult.AppliedLifecycleStateRef,
		ObservedCancelPathRef:     applyResult.CancelPathRef,
		ObservedDeletePathRef:     applyResult.DeletePathRef,
		ObservedDisablePathRef:    applyResult.DisablePathRef,
		ReadbackEvidenceRefs:      []DisplaySafeRef{"evidence:scheduler_apply_adapter_readback"},
	})
	if readback.Status != HostActionRecorded ||
		!readback.ReadyForRuntimeLoopContinuation {
		t.Fatalf("adapter-fed scheduler apply readback should continue runtime loop, got %#v", readback)
	}
}

func TestHostOwnedSchedulerApplyAdapterGateBlocksMissingConfirmationAndMismatches(t *testing.T) {
	request := schedulerApplyReadyRequest()
	missingInput := hostOwnedSchedulerApplyAdapterReadinessInput(request)
	missingInput.HostConfirmationRef = ""
	missing := BuildHostOwnedSchedulerApplyAdapterReadiness(missingInput)
	if missing.Status != HostActionBlocked ||
		missing.ReadyForHostSchedulerAdapterInvocation ||
		missing.FailureClass != FailureConfigMissing ||
		!controlTokenListContains(missing.BlockedReasons, "host_confirmation_ref_missing") ||
		!missingInputContains(missing.MissingInputs, "host:scheduler_apply_adapter_confirmation") {
		t.Fatalf("expected missing confirmation to block, got %#v", missing)
	}

	mismatchInput := hostOwnedSchedulerApplyAdapterReadinessInput(request)
	mismatchInput.AdapterDryRunRef = "dry_run:wrong"
	mismatch := BuildHostOwnedSchedulerApplyAdapterReadiness(mismatchInput)
	if mismatch.Status != HostActionBlocked ||
		mismatch.ReadyForHostSchedulerAdapterInvocation ||
		mismatch.FailureClass != FailureVerificationFailed ||
		!controlTokenListContains(mismatch.BlockedReasons, "adapter_dry_run_ref_mismatch") {
		t.Fatalf("expected dry-run mismatch to block, got %#v", mismatch)
	}

	readiness := BuildHostOwnedSchedulerApplyAdapterReadiness(hostOwnedSchedulerApplyAdapterReadinessInput(request))
	invocationInput := hostOwnedSchedulerApplyAdapterInvocationInput(readiness)
	invocationInput.ObservedInvocationRef = "invocation:wrong"
	invocationInput.SchedulerApplyResultRef = "scheduler_result:wrong"
	invocationInput.SchedulerReadbackRef = "scheduler_readback:wrong"
	invocation := BuildHostOwnedSchedulerApplyAdapterInvocation(invocationInput)
	if invocation.Status != HostActionBlocked ||
		invocation.ReadyForSchedulerApplyResult ||
		invocation.FailureClass != FailureVerificationFailed ||
		!controlTokenListContains(invocation.BlockedReasons, "observed_invocation_ref_mismatch") ||
		!controlTokenListContains(invocation.BlockedReasons, "scheduler_apply_result_ref_mismatch") ||
		!controlTokenListContains(invocation.BlockedReasons, "scheduler_readback_ref_mismatch") {
		t.Fatalf("expected invocation mismatches to block, got %#v", invocation)
	}
}

func TestHostOwnedSchedulerApplyAdapterGateFailureReportAndUnsafeRefs(t *testing.T) {
	request := schedulerApplyReadyRequest()
	readiness := BuildHostOwnedSchedulerApplyAdapterReadiness(hostOwnedSchedulerApplyAdapterReadinessInput(request))
	failedInput := hostOwnedSchedulerApplyAdapterInvocationInput(readiness)
	failedInput.HostAdapterInvocationCompleted = false
	failedInput.HostAdapterInvocationFailed = true
	failedInput.SchedulerApplyResultRef = ""
	failedInput.SchedulerReadbackRef = ""
	failedInput.AppliedScheduleRef = ""
	failedInput.AppliedLifecycleStateRef = ""
	failedInput.FailureRef = "failure:scheduler_apply_adapter"
	failed := BuildHostOwnedSchedulerApplyAdapterInvocation(failedInput)
	if failed.Status != HostActionRecorded ||
		failed.ReadyForSchedulerApplyResult ||
		!failed.ReadyForFailureReview ||
		failed.NextHostAction != "review_scheduler_apply_adapter_failure" ||
		failed.FailureClass != FailureVerificationFailed ||
		!objectiveLoopBoundaryContains(failed.Boundaries, "host_owned_scheduler_apply_adapter_failure_recorded") {
		t.Fatalf("unexpected failed scheduler apply adapter invocation: %#v", failed)
	}

	rawRef := "https://scheduler.example/raw-adapter"
	unsafeInput := hostOwnedSchedulerApplyAdapterReadinessInput(request)
	unsafeInput.AdapterRef = DisplaySafeRef(rawRef)
	unsafe := BuildHostOwnedSchedulerApplyAdapterReadiness(unsafeInput)
	if unsafe.Status != HostActionReviewRequired ||
		unsafe.ReadyForHostSchedulerAdapterInvocation ||
		unsafe.FailureClass != FailureEvidenceWeak ||
		!controlTokenListContains(unsafe.BlockedReasons, "unsafe_input_ref") ||
		!missingInputContains(unsafe.MissingInputs, "host:display_safe_refs") {
		t.Fatalf("expected unsafe adapter ref to force review, got %#v", unsafe)
	}
	payload, err := json.Marshal(unsafe)
	if err != nil {
		t.Fatalf("marshal unsafe readiness: %v", err)
	}
	if strings.Contains(string(payload), rawRef) {
		t.Fatalf("unsafe readiness leaked raw ref %q in %s", rawRef, payload)
	}
}

func hostOwnedSchedulerApplyAdapterReadinessInput(request HostOwnedSchedulerApplyRequest) HostOwnedSchedulerApplyAdapterReadinessInput {
	return HostOwnedSchedulerApplyAdapterReadinessInput{
		Request:              request,
		AdapterRef:           request.SchedulerAdapterRef,
		AdapterVersionRef:    "adapter_version:operations_scheduler_v1",
		AdapterCapabilityRef: "capability:operations_scheduler_apply",
		AdapterContractRef:   "contract:scheduler_apply_adapter",
		HostConfirmationRef:  request.HostSchedulerConfirmationRef,
		AdapterDryRunRef:     request.ScheduleDryRunProofRef,
		InvocationRef:        "invocation:scheduler_apply_agentx_release_notes_weekday_9",
		ResultBindingRef:     request.ExpectedSchedulerResultRef,
		ReadbackBindingRef:   request.ExpectedReadbackRef,
		IdempotencyRef:       request.IdempotencyRef,
		CancellationRef:      request.CancelPathRef,
		DeleteRef:            request.DeletePathRef,
		DisableRef:           request.DisablePathRef,
		PolicyRefs:           []DisplaySafeRef{"policy:scheduler_apply_adapter"},
		EvidenceRefs: []EvidenceRef{{
			Ref:      "evidence:scheduler_apply_adapter_binding",
			Kind:     "scheduler_apply_adapter_binding",
			Strength: EvidenceStrong,
			Source:   "host:scheduler_apply_adapter",
		}},
	}
}

func hostOwnedSchedulerApplyAdapterInvocationInput(readiness HostOwnedSchedulerApplyAdapterReadiness) HostOwnedSchedulerApplyAdapterInvocationInput {
	return HostOwnedSchedulerApplyAdapterInvocationInput{
		Readiness:                      readiness,
		InvocationReportRef:            "invocation_report:scheduler_apply_agentx_release_notes_weekday_9",
		ObservedInvocationRef:          readiness.InvocationRef,
		HostSchedulerAdapterRunRef:     "adapter_run:scheduler_apply_agentx_release_notes_weekday_9",
		SchedulerApplyResultRef:        readiness.ExpectedSchedulerResultRef,
		SchedulerReadbackRef:           readiness.ExpectedReadbackRef,
		AppliedScheduleRef:             readiness.ExpectedScheduleRef,
		AppliedLifecycleStateRef:       readiness.ExpectedLifecycleStateRef,
		HostAdapterInvocationReported:  true,
		HostAdapterInvocationCompleted: true,
		SchedulerEvidenceRefs:          []DisplaySafeRef{"evidence:scheduler_apply_adapter_invocation"},
	}
}
