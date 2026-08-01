package controlcontract

import (
	"testing"
)

func TestProductionAdapterPreflightReviewPacketReadyForPreflightRun(t *testing.T) {
	resolution := BuildProductionAdapterResolution(productionAdapterResolutionInput(productionAdapterTestDescriptor()))
	packet := BuildProductionAdapterPreflightReviewPacket(ProductionAdapterPreflightReviewPacketInput{
		PreflightReviewPacketRef: "review:metrics_preflight",
		Resolution:               resolution,
	})
	if packet.Status != "ready_for_adapter_preflight_review" ||
		!packet.ReadyForHostDisplay ||
		!packet.ReadyForHostPreflight ||
		packet.PreflightReported ||
		packet.ReadyForHostInvocationReview ||
		packet.NextHostAction != "host_may_run_adapter_preflight" ||
		packet.AdapterRef != "adapter:operations_local_metrics" ||
		packet.IdempotencyContractRef != "idempotency:metrics_probe" ||
		packet.TimeoutPolicyRef != "timeout:short" ||
		packet.CompensationHandoffRef != "compensation:manual_review" ||
		!productionAdapterMissingContains(packet.PreflightRequiredInputs, "host:credential_ref") ||
		!productionAdapterMissingContains(packet.PreflightRequiredInputs, "host:adapter_preflight_result_ref") {
		t.Fatalf("unexpected preflight review packet: %#v", packet)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "adapter preflight review packet",
		RunnerEffect: packet.RunnerEffect,
		PromptEffect: packet.PromptEffect,
		Boundaries:   packet.Boundaries,
		Payload:      packet,
	}, "production_adapter_preflight_review_packet", "preflight_review_packet_projection_only", "host_cli_preflight_review", "ready_for_adapter_preflight_review")
	AssertNoCoreMutation(t, "adapter preflight review packet", packet.CoreInvocationExecuted, packet.DurableWriteByCore)
}

func TestProductionAdapterPreflightReviewPacketReadyForInvocationReview(t *testing.T) {
	resolution := BuildProductionAdapterResolution(productionAdapterResolutionInput(productionAdapterTestDescriptor()))
	preflight := BuildProductionAdapterPreflight(ProductionAdapterPreflightInput{
		Resolution:             resolution,
		AdapterAvailable:       true,
		VersionStable:          true,
		CapabilitiesSatisfied:  true,
		CredentialsAvailable:   true,
		AuthorizationAvailable: true,
		HostServiceAvailable:   true,
		PolicyAllowed:          true,
		ApprovalValid:          true,
		BudgetAvailable:        true,
		IdempotencyReady:       true,
		TimeoutReady:           true,
		CompensationReady:      true,
		PreflightResultRefs:    []DisplaySafeRef{"preflight:metrics_ready"},
	})
	packet := BuildProductionAdapterPreflightReviewPacket(ProductionAdapterPreflightReviewPacketInput{
		PreflightReviewPacketRef: "review:metrics_preflight",
		Resolution:               resolution,
		Preflight:                preflight,
	})
	if packet.Status != "ready_for_adapter_invocation_review" ||
		!packet.ReadyForHostDisplay ||
		!packet.ReadyForHostPreflight ||
		!packet.PreflightReported ||
		!packet.ReadyForHostInvocationReview ||
		packet.NextHostAction != "review_adapter_invocation" ||
		packet.PreflightStatus != HostActionReady ||
		!productionAdapterDisplaySafeRefContains(packet.PreflightResultRefs, "preflight:metrics_ready") ||
		packet.CoreInvocationExecuted ||
		packet.DurableWriteByCore {
		t.Fatalf("unexpected invocation review packet: %#v", packet)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "adapter invocation review packet",
		RunnerEffect: packet.RunnerEffect,
		PromptEffect: packet.PromptEffect,
		Boundaries:   packet.Boundaries,
		Payload:      packet,
	}, "production_adapter_preflight_review_packet", "host_owned_adapter_preflight_review", "ready_for_adapter_invocation_review")
	AssertNoCoreMutation(t, "adapter invocation review packet", packet.CoreInvocationExecuted, packet.DurableWriteByCore)
}

func TestProductionAdapterPreflightReviewPacketBlocksMissingPreflightResult(t *testing.T) {
	resolution := BuildProductionAdapterResolution(productionAdapterResolutionInput(productionAdapterTestDescriptor()))
	preflight := BuildProductionAdapterPreflight(ProductionAdapterPreflightInput{
		Resolution:             resolution,
		AdapterAvailable:       true,
		VersionStable:          true,
		CapabilitiesSatisfied:  true,
		CredentialsAvailable:   true,
		AuthorizationAvailable: true,
		HostServiceAvailable:   true,
		PolicyAllowed:          true,
		ApprovalValid:          true,
		BudgetAvailable:        true,
		IdempotencyReady:       true,
		TimeoutReady:           true,
		CompensationReady:      true,
	})
	packet := BuildProductionAdapterPreflightReviewPacket(ProductionAdapterPreflightReviewPacketInput{
		PreflightReviewPacketRef: "review:metrics_preflight",
		Resolution:               resolution,
		Preflight:                preflight,
	})
	if packet.Status != "blocked" ||
		!packet.ReadyForHostDisplay ||
		packet.ReadyForHostInvocationReview ||
		packet.FailureClass != FailureEvidenceMissing ||
		packet.NextHostAction != "provide_adapter_preflight_result" ||
		!productionAdapterStringContains(packet.BlockedReasons, "preflight_result_ref_missing") ||
		!productionAdapterMissingContains(packet.MissingInputs, "host:adapter_preflight_result_ref") {
		t.Fatalf("expected missing preflight result block, got %#v", packet)
	}
}

func TestProductionAdapterPreflightReviewPacketBlocksMismatchAndUnsafeRefs(t *testing.T) {
	resolution := BuildProductionAdapterResolution(productionAdapterResolutionInput(productionAdapterTestDescriptor()))
	preflight := productionAdapterReadyPreflight()
	preflight.IdempotencyRef = "idempotency:other"
	mismatch := BuildProductionAdapterPreflightReviewPacket(ProductionAdapterPreflightReviewPacketInput{
		PreflightReviewPacketRef: "review:metrics_preflight",
		Resolution:               resolution,
		Preflight:                preflight,
	})
	if mismatch.Status != "blocked" ||
		mismatch.ReadyForHostInvocationReview ||
		mismatch.FailureClass != FailureInvalidInput ||
		!productionAdapterStringContains(mismatch.BlockedReasons, "preflight_review_idempotency_ref_mismatch") ||
		!productionAdapterMissingContains(mismatch.MissingInputs, "host:adapter_preflight") {
		t.Fatalf("expected preflight mismatch block, got %#v", mismatch)
	}

	unsafe := BuildProductionAdapterPreflightReviewPacket(ProductionAdapterPreflightReviewPacketInput{
		PreflightReviewPacketRef: "/tmp/raw-preflight-review.json",
		Resolution:               resolution,
		Preflight:                preflight,
	})
	if unsafe.Status != "blocked" ||
		unsafe.ReadyForHostDisplay ||
		unsafe.FailureClass != FailureEvidenceWeak ||
		!productionAdapterStringContains(unsafe.BlockedReasons, "unsafe_input_ref") ||
		!productionAdapterMissingContains(unsafe.MissingInputs, "host:display_safe_refs") {
		t.Fatalf("expected unsafe preflight review packet block, got %#v", unsafe)
	}
	AssertNoRawPayload(t, "unsafe adapter preflight review packet", unsafe, "/tmp/raw-preflight-review.json")
}

func TestProductionAdapterPreflightReviewPacketUnavailableWithoutResolution(t *testing.T) {
	packet := BuildProductionAdapterPreflightReviewPacket(ProductionAdapterPreflightReviewPacketInput{})
	if packet.Available ||
		packet.Status != "unavailable" ||
		packet.ReadyForHostDisplay ||
		packet.NextHostAction != "provide_adapter_resolution" {
		t.Fatalf("unexpected unavailable preflight review packet: %#v", packet)
	}
}
