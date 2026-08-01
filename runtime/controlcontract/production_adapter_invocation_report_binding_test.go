package controlcontract

import (
	"testing"
)

func TestProductionAdapterInvocationReportBindingReadyForReadbackReview(t *testing.T) {
	authorization := productionAdapterReadyInvocationAuthorizationPacket()
	invocation := productionAdapterAuthorizedSuccessInvocation(authorization)
	binding := BuildProductionAdapterInvocationReportBinding(ProductionAdapterInvocationReportBindingInput{
		InvocationReportBindingRef: "binding:metrics_invocation_report",
		AuthorizationPacket:        authorization,
		Invocation:                 invocation,
	})
	if binding.Status != "ready_for_authorized_invocation_readback_review" ||
		!binding.ReadyForHostDisplay ||
		!binding.ReadyForReadbackReview ||
		binding.ReadyForFailureReview ||
		!binding.AuthorizationBound ||
		binding.NextHostAction != "review_adapter_readback" ||
		binding.InvocationRef != authorization.InvocationRef ||
		binding.StartedEventRef != authorization.ExpectedStartedEventRef ||
		binding.CompletedEventRef != authorization.ExpectedCompletedEventRef ||
		binding.ResultRef != authorization.ExpectedResultRef ||
		binding.ReadbackRef != authorization.ExpectedReadbackRef ||
		binding.CompletionHandoffRef != authorization.ExpectedCompletionHandoffRef ||
		binding.CoreInvocationExecuted ||
		binding.DurableWriteByCore {
		t.Fatalf("unexpected invocation report binding: %#v", binding)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "adapter invocation report binding",
		RunnerEffect: binding.RunnerEffect,
		PromptEffect: binding.PromptEffect,
		Boundaries:   binding.Boundaries,
		Payload:      binding,
	}, "production_adapter_invocation_report_binding", "authorization_bound_invocation_report", "host_owned_invocation_report", "core_invocation_not_executed", "no_adapter_invocation", "ready_for_authorized_invocation_readback_review")
	AssertNoCoreMutation(t, "adapter invocation report binding", binding.CoreInvocationExecuted, binding.DurableWriteByCore)
}

func TestProductionAdapterInvocationReportBindingReadyForFailureReview(t *testing.T) {
	authorization := productionAdapterReadyInvocationAuthorizationPacket()
	invocation := productionAdapterAuthorizedFailureInvocation(authorization)
	binding := BuildProductionAdapterInvocationReportBinding(ProductionAdapterInvocationReportBindingInput{
		InvocationReportBindingRef: "binding:metrics_invocation_report",
		AuthorizationPacket:        authorization,
		Invocation:                 invocation,
	})
	if binding.Status != "ready_for_authorized_invocation_failure_review" ||
		!binding.ReadyForHostDisplay ||
		binding.ReadyForReadbackReview ||
		!binding.ReadyForFailureReview ||
		!binding.AuthorizationBound ||
		binding.FailureClass != FailureVerificationFailed ||
		binding.NextHostAction != "review_adapter_failure" ||
		binding.FailureRef != authorization.ExpectedFailureRef ||
		binding.CompensationRef != authorization.ExpectedCompensationRef ||
		binding.CoreInvocationExecuted ||
		binding.DurableWriteByCore {
		t.Fatalf("unexpected failure invocation report binding: %#v", binding)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "adapter failure invocation report binding",
		RunnerEffect: binding.RunnerEffect,
		PromptEffect: binding.PromptEffect,
		Boundaries:   binding.Boundaries,
		Payload:      binding,
	}, "production_adapter_invocation_report_binding", "authorization_bound_invocation_report", "host_owned_invocation_report", "authorized_invocation_failure_report")
	AssertNoCoreMutation(t, "adapter failure invocation report binding", binding.CoreInvocationExecuted, binding.DurableWriteByCore)
}

func TestProductionAdapterInvocationReportBindingBlocksMismatch(t *testing.T) {
	authorization := productionAdapterReadyInvocationAuthorizationPacket()
	invocation := productionAdapterAuthorizedSuccessInvocation(authorization)
	invocation.ResultRef = "result:other"
	mismatch := BuildProductionAdapterInvocationReportBinding(ProductionAdapterInvocationReportBindingInput{
		InvocationReportBindingRef: "binding:metrics_invocation_report",
		AuthorizationPacket:        authorization,
		Invocation:                 invocation,
	})
	if mismatch.Status != "blocked" ||
		mismatch.ReadyForReadbackReview ||
		mismatch.AuthorizationBound ||
		mismatch.FailureClass != FailureVerificationFailed ||
		!productionAdapterStringContains(mismatch.BlockedReasons, "invocation_report_result_ref_mismatch") ||
		!productionAdapterMissingContains(mismatch.MissingInputs, "host:result_ref") {
		t.Fatalf("expected invocation report mismatch block, got %#v", mismatch)
	}
}

func TestProductionAdapterInvocationReportBindingBlocksAuthorizationNotReadyAndMissingInvocation(t *testing.T) {
	review := productionAdapterReadyPreflightReviewPacket()
	authInput := productionAdapterInvocationAuthorizationInput(review)
	authInput.HostConfirmationRef = ""
	authorization := BuildProductionAdapterInvocationAuthorizationPacket(authInput)
	notReady := BuildProductionAdapterInvocationReportBinding(ProductionAdapterInvocationReportBindingInput{
		InvocationReportBindingRef: "binding:metrics_invocation_report",
		AuthorizationPacket:        authorization,
		Invocation:                 productionAdapterAuthorizedSuccessInvocation(productionAdapterReadyInvocationAuthorizationPacket()),
	})
	if notReady.Status != "blocked" ||
		notReady.ReadyForReadbackReview ||
		notReady.AuthorizationBound ||
		notReady.FailureClass != FailureAuthorizationMissing ||
		!productionAdapterStringContains(notReady.BlockedReasons, "adapter_invocation_authorization_not_ready") ||
		!productionAdapterMissingContains(notReady.MissingInputs, "host:invocation_authorization_packet") {
		t.Fatalf("expected authorization not-ready block, got %#v", notReady)
	}

	missingInvocation := BuildProductionAdapterInvocationReportBinding(ProductionAdapterInvocationReportBindingInput{
		InvocationReportBindingRef: "binding:metrics_invocation_report",
		AuthorizationPacket:        productionAdapterReadyInvocationAuthorizationPacket(),
	})
	if missingInvocation.Status != "blocked" ||
		!missingInvocation.ReadyForHostDisplay ||
		missingInvocation.AuthorizationBound ||
		missingInvocation.NextHostAction != "provide_adapter_invocation" ||
		!productionAdapterStringContains(missingInvocation.BlockedReasons, "adapter_invocation_report_missing") ||
		!productionAdapterMissingContains(missingInvocation.MissingInputs, "host:adapter_invocation") {
		t.Fatalf("expected missing invocation block, got %#v", missingInvocation)
	}
}

func TestProductionAdapterInvocationReportBindingBlocksUnsafeAndUnavailable(t *testing.T) {
	authorization := productionAdapterReadyInvocationAuthorizationPacket()
	unsafe := BuildProductionAdapterInvocationReportBinding(ProductionAdapterInvocationReportBindingInput{
		InvocationReportBindingRef: "/tmp/raw-binding.json",
		AuthorizationPacket:        authorization,
		Invocation:                 productionAdapterAuthorizedSuccessInvocation(authorization),
	})
	if unsafe.Status != "blocked" ||
		unsafe.ReadyForHostDisplay ||
		unsafe.FailureClass != FailureEvidenceWeak ||
		!productionAdapterStringContains(unsafe.BlockedReasons, "unsafe_input_ref") ||
		!productionAdapterMissingContains(unsafe.MissingInputs, "host:display_safe_refs") {
		t.Fatalf("expected unsafe invocation report binding block, got %#v", unsafe)
	}
	AssertNoRawPayload(t, "unsafe adapter invocation report binding", unsafe, "/tmp/raw-binding.json")

	unavailable := BuildProductionAdapterInvocationReportBinding(ProductionAdapterInvocationReportBindingInput{})
	if unavailable.Available ||
		unavailable.Status != "unavailable" ||
		unavailable.ReadyForHostDisplay ||
		unavailable.NextHostAction != "provide_invocation_authorization_packet" {
		t.Fatalf("unexpected unavailable invocation report binding: %#v", unavailable)
	}
}

func productionAdapterReadyInvocationAuthorizationPacket() ProductionAdapterInvocationAuthorizationPacket {
	review := productionAdapterReadyPreflightReviewPacket()
	return BuildProductionAdapterInvocationAuthorizationPacket(productionAdapterInvocationAuthorizationInput(review))
}

func productionAdapterAuthorizedSuccessInvocation(authorization ProductionAdapterInvocationAuthorizationPacket) HostAdapterInvocationProjection {
	return BuildHostAdapterInvocationProjection(HostAdapterInvocationInput{
		Preflight:               productionAdapterReadyPreflight(),
		InvocationRef:           authorization.InvocationRef,
		IdempotencyRef:          authorization.IdempotencyRef,
		StartedEventRef:         authorization.ExpectedStartedEventRef,
		CompletedEventRef:       authorization.ExpectedCompletedEventRef,
		ResultRef:               authorization.ExpectedResultRef,
		ReadbackRef:             authorization.ExpectedReadbackRef,
		CompletionHandoffRef:    authorization.ExpectedCompletionHandoffRef,
		HostInvocationCompleted: true,
	})
}

func productionAdapterAuthorizedFailureInvocation(authorization ProductionAdapterInvocationAuthorizationPacket) HostAdapterInvocationProjection {
	return BuildHostAdapterInvocationProjection(HostAdapterInvocationInput{
		Preflight:            productionAdapterReadyPreflight(),
		InvocationRef:        authorization.InvocationRef,
		IdempotencyRef:       authorization.IdempotencyRef,
		StartedEventRef:      authorization.ExpectedStartedEventRef,
		CompletedEventRef:    authorization.ExpectedCompletedEventRef,
		FailureRef:           authorization.ExpectedFailureRef,
		CompensationRef:      authorization.ExpectedCompensationRef,
		HostInvocationFailed: true,
	})
}
