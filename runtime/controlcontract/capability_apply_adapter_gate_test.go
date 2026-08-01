package controlcontract

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestHostOwnedCapabilityApplyAdapterGateBuildsReadinessAndInvocation(t *testing.T) {
	request := capabilityApplyReadyRequest()
	readiness := BuildHostOwnedCapabilityApplyAdapterReadiness(hostOwnedCapabilityApplyAdapterReadinessInput(request))
	if readiness.Status != HostActionReady ||
		!readiness.ReadyForHostCapabilityAdapterInvocation ||
		!readiness.HostCapabilityAdapterInvocationAuthorized ||
		!readiness.HostMayInvokeCapabilityAdapter ||
		readiness.NextHostAction != "host_may_invoke_capability_apply_adapter" ||
		readiness.AdapterRef != request.CapabilityAdapterRef ||
		readiness.AdapterDryRunRef != request.CapabilityDryRunProofRef ||
		readiness.ResultBindingRef != request.ExpectedCapabilityResultRef ||
		readiness.ReadbackBindingRef != request.ExpectedReadbackRef ||
		readiness.IdempotencyRef != request.IdempotencyRef ||
		readiness.RollbackRef != request.RollbackPathRef ||
		readiness.FailureBindingRef == "" ||
		readiness.CompensationBindingRef == "" ||
		readiness.CoreInvocationExecuted ||
		readiness.InstallerExecutedByCore ||
		readiness.InstallExecutedByCore ||
		readiness.EnableExecutedByCore ||
		readiness.PackageManagerExecutedByCore ||
		readiness.SkillWriteByCore ||
		readiness.RuntimeReloadByCore ||
		readiness.RunnerDispatched ||
		readiness.RuntimeAdapterExecuted ||
		readiness.WorkflowDispatched ||
		readiness.WorkerDispatched ||
		readiness.StoreMutationExecuted ||
		readiness.CompensationExecuted {
		t.Fatalf("unexpected capability apply adapter readiness: %#v", readiness)
	}
	assertHostOwnedProjectionOnly(t, testProjection[Boundary]{
		Name:         "capability apply adapter readiness",
		RunnerEffect: readiness.RunnerEffect,
		PromptEffect: readiness.PromptEffect,
		Boundaries:   readiness.Boundaries,
		Payload:      readiness,
	}, "host_owned_capability_apply_adapter_gate", "capability_apply_adapter_invocation_gate", "explicit_host_confirmation_required", "capability_apply_adapter_dry_run_required", "no_capability_apply_by_core", "no_package_manager_execution_by_core", "no_skill_write_by_core", "no_runtime_reload_by_core")

	invocation := BuildHostOwnedCapabilityApplyAdapterInvocation(hostOwnedCapabilityApplyAdapterInvocationInput(readiness))
	if invocation.Status != HostActionRecorded ||
		!invocation.ReadyForCapabilityApplyResult ||
		invocation.ReadyForFailureReview ||
		!invocation.HostAdapterInvocationReported ||
		!invocation.HostAdapterInvocationCompleted ||
		invocation.HostAdapterInvocationFailed ||
		invocation.NextHostAction != "build_capability_apply_result" ||
		invocation.ObservedInvocationRef != readiness.InvocationRef ||
		invocation.CapabilityApplyResultRef != request.ExpectedCapabilityResultRef ||
		invocation.CapabilityReadbackRef != request.ExpectedReadbackRef ||
		invocation.AppliedCapabilityRef != request.ExpectedCapabilityRef ||
		invocation.AppliedCapabilityStateRef != request.ExpectedCapabilityStateRef ||
		invocation.CoreInvocationExecuted ||
		invocation.InstallerExecutedByCore ||
		invocation.InstallExecutedByCore ||
		invocation.EnableExecutedByCore ||
		invocation.PackageManagerExecutedByCore ||
		invocation.SkillWriteByCore ||
		invocation.RuntimeReloadByCore ||
		invocation.RunnerDispatched ||
		invocation.RuntimeAdapterExecuted ||
		invocation.WorkflowDispatched ||
		invocation.WorkerDispatched ||
		invocation.StoreMutationExecuted ||
		invocation.CompensationExecuted {
		t.Fatalf("unexpected capability apply adapter invocation: %#v", invocation)
	}
	assertHostOwnedProjectionOnly(t, testProjection[Boundary]{
		Name:         "capability apply adapter invocation",
		RunnerEffect: invocation.RunnerEffect,
		PromptEffect: invocation.PromptEffect,
		Boundaries:   invocation.Boundaries,
		Payload:      invocation,
	}, "host_owned_capability_apply_adapter_gate", "host_capability_adapter_invocation_report_only", "host_adapter_capability_mutation_reported_only", "capability_apply_result_requires_readback", "no_capability_apply_by_core", "no_package_manager_execution_by_core", "no_skill_write_by_core", "no_runtime_reload_by_core")

	applyResult := BuildHostOwnedCapabilityApplyResult(HostOwnedCapabilityApplyResultInput{
		Request:                      request,
		CapabilityApplyResultRef:     invocation.CapabilityApplyResultRef,
		HostCapabilityRunRef:         invocation.HostCapabilityAdapterRunRef,
		HostCapabilityApplyReported:  true,
		HostCapabilityApplySucceeded: true,
		AppliedCapabilityRef:         invocation.AppliedCapabilityRef,
		AppliedCapabilityStateRef:    invocation.AppliedCapabilityStateRef,
		CapabilityEvidenceRefs:       invocation.CapabilityEvidenceRefs,
	})
	if applyResult.Status != HostActionRecorded ||
		!applyResult.ReadyForCapabilityReadback ||
		!applyResult.HostCapabilityInstalled ||
		applyResult.CoreInvocationExecuted ||
		applyResult.PackageManagerExecutedByCore ||
		applyResult.SkillWriteByCore ||
		applyResult.RuntimeReloadByCore {
		t.Fatalf("adapter invocation should feed capability apply result, got %#v", applyResult)
	}
	readback := BuildHostOwnedCapabilityApplyReadback(HostOwnedCapabilityApplyReadbackInput{
		Result:                     applyResult,
		CapabilityReadbackRef:      invocation.CapabilityReadbackRef,
		ObservedCapabilityRef:      applyResult.AppliedCapabilityRef,
		ObservedCapabilityStateRef: applyResult.AppliedCapabilityStateRef,
		ObservedRollbackPathRef:    applyResult.RollbackPathRef,
		ReadbackEvidenceRefs:       []DisplaySafeRef{"evidence:capability_apply_adapter_readback"},
	})
	if readback.Status != HostActionRecorded ||
		!readback.ReadyForRuntimeLoopContinuation {
		t.Fatalf("adapter-fed capability apply readback should continue runtime loop, got %#v", readback)
	}
}

func TestHostOwnedCapabilityApplyAdapterGateBlocksMissingConfirmationAndMismatches(t *testing.T) {
	request := capabilityApplyReadyRequest()
	missingInput := hostOwnedCapabilityApplyAdapterReadinessInput(request)
	missingInput.HostConfirmationRef = ""
	missing := BuildHostOwnedCapabilityApplyAdapterReadiness(missingInput)
	if missing.Status != HostActionBlocked ||
		missing.ReadyForHostCapabilityAdapterInvocation ||
		missing.FailureClass != FailureConfigMissing ||
		!controlTokenListContains(missing.BlockedReasons, "host_confirmation_ref_missing") ||
		!missingInputContains(missing.MissingInputs, "host:capability_apply_adapter_confirmation") {
		t.Fatalf("expected missing confirmation to block, got %#v", missing)
	}

	mismatchInput := hostOwnedCapabilityApplyAdapterReadinessInput(request)
	mismatchInput.AdapterDryRunRef = "dry_run:wrong"
	mismatch := BuildHostOwnedCapabilityApplyAdapterReadiness(mismatchInput)
	if mismatch.Status != HostActionBlocked ||
		mismatch.ReadyForHostCapabilityAdapterInvocation ||
		mismatch.FailureClass != FailureVerificationFailed ||
		!controlTokenListContains(mismatch.BlockedReasons, "adapter_dry_run_ref_mismatch") {
		t.Fatalf("expected dry-run mismatch to block, got %#v", mismatch)
	}

	readiness := BuildHostOwnedCapabilityApplyAdapterReadiness(hostOwnedCapabilityApplyAdapterReadinessInput(request))
	invocationInput := hostOwnedCapabilityApplyAdapterInvocationInput(readiness)
	invocationInput.ObservedInvocationRef = "invocation:wrong"
	invocationInput.CapabilityApplyResultRef = "capability_result:wrong"
	invocationInput.CapabilityReadbackRef = "capability_readback:wrong"
	invocation := BuildHostOwnedCapabilityApplyAdapterInvocation(invocationInput)
	if invocation.Status != HostActionBlocked ||
		invocation.ReadyForCapabilityApplyResult ||
		invocation.FailureClass != FailureVerificationFailed ||
		!controlTokenListContains(invocation.BlockedReasons, "observed_invocation_ref_mismatch") ||
		!controlTokenListContains(invocation.BlockedReasons, "capability_apply_result_ref_mismatch") ||
		!controlTokenListContains(invocation.BlockedReasons, "capability_readback_ref_mismatch") {
		t.Fatalf("expected invocation mismatches to block, got %#v", invocation)
	}
}

func TestHostOwnedCapabilityApplyAdapterGateFailureReportAndUnsafeRefs(t *testing.T) {
	request := capabilityApplyReadyRequest()
	readiness := BuildHostOwnedCapabilityApplyAdapterReadiness(hostOwnedCapabilityApplyAdapterReadinessInput(request))
	failedInput := hostOwnedCapabilityApplyAdapterInvocationInput(readiness)
	failedInput.HostAdapterInvocationCompleted = false
	failedInput.HostAdapterInvocationFailed = true
	failedInput.CapabilityApplyResultRef = ""
	failedInput.CapabilityReadbackRef = ""
	failedInput.AppliedCapabilityRef = ""
	failedInput.AppliedCapabilityStateRef = ""
	failedInput.FailureRef = "failure:capability_apply_adapter"
	failedInput.CompensationRef = "compensation:capability_apply_adapter"
	failed := BuildHostOwnedCapabilityApplyAdapterInvocation(failedInput)
	if failed.Status != HostActionRecorded ||
		failed.ReadyForCapabilityApplyResult ||
		!failed.ReadyForFailureReview ||
		failed.NextHostAction != "review_capability_apply_adapter_failure" ||
		failed.FailureClass != FailureVerificationFailed ||
		!objectiveLoopBoundaryContains(failed.Boundaries, "host_owned_capability_apply_adapter_failure_recorded") {
		t.Fatalf("unexpected failed capability apply adapter invocation: %#v", failed)
	}

	rawRef := "https://installer.example/raw-adapter"
	unsafeInput := hostOwnedCapabilityApplyAdapterReadinessInput(request)
	unsafeInput.AdapterRef = DisplaySafeRef(rawRef)
	unsafe := BuildHostOwnedCapabilityApplyAdapterReadiness(unsafeInput)
	if unsafe.Status != HostActionReviewRequired ||
		unsafe.ReadyForHostCapabilityAdapterInvocation ||
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

func hostOwnedCapabilityApplyAdapterReadinessInput(request HostOwnedCapabilityApplyRequest) HostOwnedCapabilityApplyAdapterReadinessInput {
	return HostOwnedCapabilityApplyAdapterReadinessInput{
		Request:                request,
		AdapterRef:             request.CapabilityAdapterRef,
		AdapterVersionRef:      "adapter_version:capability_apply_v1",
		AdapterCapabilityRef:   "capability:host_capability_apply",
		AdapterContractRef:     "contract:capability_apply_adapter",
		HostConfirmationRef:    request.HostCapabilityConfirmationRef,
		AdapterDryRunRef:       request.CapabilityDryRunProofRef,
		InvocationRef:          "invocation:capability_apply_missing_tool",
		ResultBindingRef:       request.ExpectedCapabilityResultRef,
		ReadbackBindingRef:     request.ExpectedReadbackRef,
		IdempotencyRef:         request.IdempotencyRef,
		RollbackRef:            request.RollbackPathRef,
		FailureBindingRef:      "failure:capability_apply_adapter",
		CompensationBindingRef: "compensation:capability_apply_adapter",
		PolicyRefs:             []DisplaySafeRef{"policy:capability_apply_adapter"},
		EvidenceRefs: []EvidenceRef{{
			Ref:      "evidence:capability_apply_adapter_binding",
			Kind:     "capability_apply_adapter_binding",
			Strength: EvidenceStrong,
			Source:   "host:capability_apply_adapter",
		}},
	}
}

func hostOwnedCapabilityApplyAdapterInvocationInput(readiness HostOwnedCapabilityApplyAdapterReadiness) HostOwnedCapabilityApplyAdapterInvocationInput {
	return HostOwnedCapabilityApplyAdapterInvocationInput{
		Readiness:                      readiness,
		InvocationReportRef:            "invocation_report:capability_apply_missing_tool",
		ObservedInvocationRef:          readiness.InvocationRef,
		HostCapabilityAdapterRunRef:    "adapter_run:capability_apply_missing_tool",
		CapabilityApplyResultRef:       readiness.ExpectedCapabilityResultRef,
		CapabilityReadbackRef:          readiness.ExpectedReadbackRef,
		AppliedCapabilityRef:           readiness.ExpectedCapabilityRef,
		AppliedCapabilityStateRef:      readiness.ExpectedCapabilityStateRef,
		HostAdapterInvocationReported:  true,
		HostAdapterInvocationCompleted: true,
		CapabilityEvidenceRefs:         []DisplaySafeRef{"evidence:capability_apply_adapter_invocation"},
	}
}
