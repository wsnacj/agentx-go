package controlcontract

import (
	"testing"
)

func TestProductionAdapterReadbackReviewFromBindingReadyForCompletionAudit(t *testing.T) {
	authorization := productionAdapterReadyInvocationAuthorizationPacket()
	binding := BuildProductionAdapterInvocationReportBinding(ProductionAdapterInvocationReportBindingInput{
		InvocationReportBindingRef: "binding:metrics_invocation_report",
		AuthorizationPacket:        authorization,
		Invocation:                 productionAdapterAuthorizedSuccessInvocation(authorization),
	})

	review := BuildProductionAdapterReadbackReviewFromBinding(binding)
	if review.Status != HostActionReady ||
		!review.ReadyForReadbackReview ||
		review.ReadyForFailureReview ||
		!review.ReadyForCompletionAudit ||
		!review.AuthorizationBound ||
		review.InvocationReportBindingRef != binding.InvocationReportBindingRef ||
		review.AuthorizationPacketRef != binding.AuthorizationPacketRef ||
		review.PreflightReviewPacketRef != binding.PreflightReviewPacketRef ||
		review.NextHostAction != "review_completion_handoff" ||
		review.ObjectiveSatisfied ||
		review.CoreInvocationExecuted ||
		review.DurableWriteByCore ||
		!productionAdapterEvidenceContains(review.EvidenceRefs, binding.InvocationReportBindingRef, "invocation_report_binding") ||
		!productionAdapterBoundaryContains(review.Boundaries, "authorization_bound_readback_review") {
		t.Fatalf("unexpected bound readback review: %#v", review)
	}
	AssertHostOwnedAuditInputOnly(t, Projection[Boundary]{
		Name:         "bound adapter readback review",
		RunnerEffect: review.RunnerEffect,
		PromptEffect: review.PromptEffect,
		Boundaries:   review.Boundaries,
		Payload:      review,
	}, review.ObjectiveSatisfied, false, review.CoreInvocationExecuted, review.DurableWriteByCore, "production_adapter_readback_review", "readback_review_from_invocation_report_binding", "authorization_bound_readback_review", "objective_not_satisfied_by_adapter")
}

func TestProductionAdapterReadbackReviewFromBindingReadyForFailureReview(t *testing.T) {
	authorization := productionAdapterReadyInvocationAuthorizationPacket()
	binding := BuildProductionAdapterInvocationReportBinding(ProductionAdapterInvocationReportBindingInput{
		InvocationReportBindingRef: "binding:metrics_invocation_report",
		AuthorizationPacket:        authorization,
		Invocation:                 productionAdapterAuthorizedFailureInvocation(authorization),
	})

	review := BuildProductionAdapterReadbackReviewFromBinding(binding)
	if review.Status != HostActionReviewRequired ||
		review.ReadyForReadbackReview ||
		!review.ReadyForFailureReview ||
		review.ReadyForCompletionAudit ||
		!review.AuthorizationBound ||
		review.FailureClass != FailureVerificationFailed ||
		review.NextHostAction != "review_adapter_failure" ||
		review.FailureRef != binding.FailureRef ||
		review.CompensationRef != binding.CompensationRef ||
		!productionAdapterBoundaryContains(review.Boundaries, "authorization_bound_readback_review") {
		t.Fatalf("unexpected bound failure review: %#v", review)
	}
}

func TestProductionAdapterReadbackReviewFromBindingBlocksUnboundReport(t *testing.T) {
	authorization := productionAdapterReadyInvocationAuthorizationPacket()
	invocation := productionAdapterAuthorizedSuccessInvocation(authorization)
	invocation.ResultRef = "result:other"
	binding := BuildProductionAdapterInvocationReportBinding(ProductionAdapterInvocationReportBindingInput{
		InvocationReportBindingRef: "binding:metrics_invocation_report",
		AuthorizationPacket:        authorization,
		Invocation:                 invocation,
	})

	review := BuildProductionAdapterReadbackReviewFromBinding(binding)
	if review.Status != HostActionBlocked ||
		review.ReadyForReadbackReview ||
		review.ReadyForFailureReview ||
		review.ReadyForCompletionAudit ||
		review.AuthorizationBound ||
		review.FailureClass != FailureVerificationFailed ||
		!productionAdapterStringContains(review.BlockedReasons, "invocation_report_binding_not_ready") ||
		!productionAdapterMissingContains(review.MissingInputs, "host:invocation_report_binding") {
		t.Fatalf("expected unbound report to block readback review, got %#v", review)
	}
}

func TestProductionAdapterReadbackReviewFromBindingRejectsUnsafeRefs(t *testing.T) {
	authorization := productionAdapterReadyInvocationAuthorizationPacket()
	binding := BuildProductionAdapterInvocationReportBinding(ProductionAdapterInvocationReportBindingInput{
		InvocationReportBindingRef: "binding:metrics_invocation_report",
		AuthorizationPacket:        authorization,
		Invocation:                 productionAdapterAuthorizedSuccessInvocation(authorization),
	})
	binding.ReadbackRef = "/tmp/raw-readback.json"

	review := BuildProductionAdapterReadbackReviewFromBinding(binding)
	if review.Status != HostActionBlocked ||
		review.ReadyForReadbackReview ||
		review.AuthorizationBound ||
		review.FailureClass != FailureEvidenceWeak ||
		!productionAdapterStringContains(review.BlockedReasons, "unsafe_input_ref") ||
		!productionAdapterMissingContains(review.MissingInputs, "host:display_safe_refs") {
		t.Fatalf("expected unsafe binding ref to block readback review, got %#v", review)
	}
	AssertNoRawPayload(t, "unsafe bound adapter readback review", review, "/tmp/raw-readback.json")
}
