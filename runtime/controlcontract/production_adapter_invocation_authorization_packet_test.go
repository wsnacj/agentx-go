package controlcontract

import (
	"testing"
)

func TestProductionAdapterInvocationAuthorizationPacketReadyForHostInvocation(t *testing.T) {
	review := productionAdapterReadyPreflightReviewPacket()
	packet := BuildProductionAdapterInvocationAuthorizationPacket(productionAdapterInvocationAuthorizationInput(review))
	if packet.Status != "ready_for_host_adapter_invocation_authorization" ||
		!packet.ReadyForHostDisplay ||
		!packet.ReadyForHostAuthorization ||
		!packet.ReadyForHostInvocation ||
		!packet.HostInvocationAuthorized ||
		packet.NextHostAction != "host_may_invoke_adapter" ||
		packet.PreflightReviewPacketRef != "review:metrics_preflight" ||
		packet.AuthorizationPacketRef != "authorization:metrics_invocation" ||
		packet.InvocationRef != "invocation:metrics_1" ||
		packet.IdempotencyRef != review.IdempotencyRef ||
		packet.TimeoutPolicyRef != review.TimeoutPolicyRef ||
		packet.CancellationPolicyRef != "cancellation:metrics_abort" ||
		packet.CoreInvocationExecuted ||
		packet.DurableWriteByCore ||
		!productionAdapterDisplaySafeRefContains(packet.PreflightResultRefs, "preflight:metrics_ready") ||
		!productionAdapterMissingContains(packet.AuthorizationRequiredInputs, "host:cancellation_policy_ref") {
		t.Fatalf("unexpected invocation authorization packet: %#v", packet)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "adapter invocation authorization packet",
		RunnerEffect: packet.RunnerEffect,
		PromptEffect: packet.PromptEffect,
		Boundaries:   packet.Boundaries,
		Payload:      packet,
	}, "production_adapter_invocation_authorization_packet", "invocation_authorization_packet_projection_only", "host_owned_invocation_authorization", "host_final_confirmation_required", "no_adapter_invocation", "ready_for_host_adapter_invocation_authorization")
	AssertNoCoreMutation(t, "adapter invocation authorization packet", packet.CoreInvocationExecuted, packet.DurableWriteByCore)
}

func TestProductionAdapterInvocationAuthorizationPacketBlocksMissingConfirmationAndExpectedRefs(t *testing.T) {
	review := productionAdapterReadyPreflightReviewPacket()
	input := productionAdapterInvocationAuthorizationInput(review)
	input.HostConfirmationRef = ""
	input.ExpectedResultRef = ""
	input.ExpectedReadbackRef = ""
	packet := BuildProductionAdapterInvocationAuthorizationPacket(input)
	if packet.Status != "blocked" ||
		!packet.ReadyForHostDisplay ||
		packet.ReadyForHostAuthorization ||
		packet.ReadyForHostInvocation ||
		packet.HostInvocationAuthorized ||
		packet.FailureClass != FailureAuthorizationMissing ||
		packet.NextHostAction != "request_host_invocation_confirmation" ||
		!productionAdapterStringContains(packet.BlockedReasons, "host_confirmation_ref_missing") ||
		!productionAdapterStringContains(packet.BlockedReasons, "expected_result_ref_missing") ||
		!productionAdapterStringContains(packet.BlockedReasons, "expected_readback_ref_missing") ||
		!productionAdapterMissingContains(packet.MissingInputs, "host:host_confirmation_ref") ||
		!productionAdapterMissingContains(packet.MissingInputs, "host:expected_result_ref") ||
		!productionAdapterMissingContains(packet.MissingInputs, "host:expected_readback_ref") {
		t.Fatalf("expected missing authorization inputs block, got %#v", packet)
	}
}

func TestProductionAdapterInvocationAuthorizationPacketBlocksPreflightReviewNotReady(t *testing.T) {
	resolution := BuildProductionAdapterResolution(productionAdapterResolutionInput(productionAdapterTestDescriptor()))
	review := BuildProductionAdapterPreflightReviewPacket(ProductionAdapterPreflightReviewPacketInput{
		PreflightReviewPacketRef: "review:metrics_preflight",
		Resolution:               resolution,
	})
	input := productionAdapterInvocationAuthorizationInput(review)
	packet := BuildProductionAdapterInvocationAuthorizationPacket(input)
	if packet.Status != "blocked" ||
		packet.ReadyForHostAuthorization ||
		packet.ReadyForHostInvocation ||
		packet.HostInvocationAuthorized ||
		packet.FailureClass != FailureConfigMissing ||
		!productionAdapterStringContains(packet.BlockedReasons, "adapter_preflight_review_not_ready") ||
		!productionAdapterMissingContains(packet.MissingInputs, "host:adapter_preflight_review_packet") {
		t.Fatalf("expected preflight review not-ready block, got %#v", packet)
	}
}

func TestProductionAdapterInvocationAuthorizationPacketBlocksMismatchAndUnsafeRefs(t *testing.T) {
	review := productionAdapterReadyPreflightReviewPacket()
	mismatchInput := productionAdapterInvocationAuthorizationInput(review)
	mismatchInput.IdempotencyRef = "idempotency:other"
	mismatch := BuildProductionAdapterInvocationAuthorizationPacket(mismatchInput)
	if mismatch.Status != "blocked" ||
		mismatch.ReadyForHostAuthorization ||
		mismatch.FailureClass != FailureInvalidInput ||
		!productionAdapterStringContains(mismatch.BlockedReasons, "authorization_idempotency_ref_mismatch") ||
		!productionAdapterMissingContains(mismatch.MissingInputs, "host:idempotency_ref") {
		t.Fatalf("expected invocation authorization mismatch block, got %#v", mismatch)
	}

	unsafeInput := productionAdapterInvocationAuthorizationInput(review)
	unsafeInput.AuthorizationPacketRef = "/tmp/raw-authorization.json"
	unsafe := BuildProductionAdapterInvocationAuthorizationPacket(unsafeInput)
	if unsafe.Status != "blocked" ||
		unsafe.ReadyForHostDisplay ||
		unsafe.FailureClass != FailureEvidenceWeak ||
		!productionAdapterStringContains(unsafe.BlockedReasons, "unsafe_input_ref") ||
		!productionAdapterMissingContains(unsafe.MissingInputs, "host:display_safe_refs") {
		t.Fatalf("expected unsafe invocation authorization packet block, got %#v", unsafe)
	}
	AssertNoRawPayload(t, "unsafe adapter invocation authorization packet", unsafe, "/tmp/raw-authorization.json")
}

func TestProductionAdapterInvocationAuthorizationPacketUnavailableWithoutReview(t *testing.T) {
	packet := BuildProductionAdapterInvocationAuthorizationPacket(ProductionAdapterInvocationAuthorizationPacketInput{})
	if packet.Available ||
		packet.Status != "unavailable" ||
		packet.ReadyForHostDisplay ||
		packet.ReadyForHostAuthorization ||
		packet.ReadyForHostInvocation ||
		packet.NextHostAction != "provide_adapter_preflight_review_packet" {
		t.Fatalf("unexpected unavailable invocation authorization packet: %#v", packet)
	}
}

func productionAdapterReadyPreflightReviewPacket() ProductionAdapterPreflightReviewPacket {
	resolution := BuildProductionAdapterResolution(productionAdapterResolutionInput(productionAdapterTestDescriptor()))
	return BuildProductionAdapterPreflightReviewPacket(ProductionAdapterPreflightReviewPacketInput{
		PreflightReviewPacketRef: "review:metrics_preflight",
		Resolution:               resolution,
		Preflight:                productionAdapterReadyPreflight(),
	})
}

func productionAdapterInvocationAuthorizationInput(review ProductionAdapterPreflightReviewPacket) ProductionAdapterInvocationAuthorizationPacketInput {
	return ProductionAdapterInvocationAuthorizationPacketInput{
		AuthorizationPacketRef:       "authorization:metrics_invocation",
		PreflightReviewPacket:        review,
		InvocationRef:                "invocation:metrics_1",
		HostConfirmationRef:          "confirmation:metrics_invocation",
		ApprovalRefs:                 append([]DisplaySafeRef(nil), review.RequiredApprovalRefs...),
		IdempotencyRef:               review.IdempotencyRef,
		ExpectedStartedEventRef:      "event:metrics_started",
		ExpectedCompletedEventRef:    "event:metrics_completed",
		ExpectedResultRef:            "result:metrics_summary",
		ExpectedReadbackRef:          "readback:metrics_summary",
		ExpectedFailureRef:           "failure:metrics_failed",
		ExpectedCompensationRef:      "compensation:metrics_review",
		ExpectedCompletionHandoffRef: "handoff:metrics_completion",
		TimeoutPolicyRef:             review.TimeoutPolicyRef,
		CancellationPolicyRef:        "cancellation:metrics_abort",
		HostPolicyRef:                review.HostPolicyRef,
		BudgetRef:                    review.BudgetRef,
	}
}
