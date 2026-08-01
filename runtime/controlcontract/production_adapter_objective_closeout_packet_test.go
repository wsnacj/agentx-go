package controlcontract

import (
	"testing"
)

func TestProductionAdapterObjectiveCloseoutPacketRequiresHostConfirmation(t *testing.T) {
	auditBinding := productionAdapterReadyCompletionAuditResultBinding()

	packet := BuildProductionAdapterObjectiveCloseoutPacket(ProductionAdapterObjectiveCloseoutPacketInput{
		ObjectiveCloseoutPacketRef:   "closeout:metrics_objective",
		ObjectiveRef:                 "objective:metrics_monitoring",
		CompletionAuditResultBinding: auditBinding,
	})
	if packet.Status != VerificationReviewRequired ||
		!packet.ReadyForHostCloseoutReview ||
		packet.ReadyForObjectiveCompletion ||
		packet.ObjectiveSatisfied ||
		!packet.VerificationSatisfied ||
		packet.HostCloseoutConfirmed ||
		packet.FailureClass != FailureApprovalRequired ||
		packet.NextHostAction != "confirm_objective_closeout" ||
		!productionAdapterMissingContains(packet.MissingInputs, "host:objective_closeout_confirmation_ref") ||
		!productionAdapterStringContains(packet.BlockedReasons, "objective_closeout_confirmation_required") ||
		!productionAdapterBoundaryContains(packet.Boundaries, "host_objective_closeout_confirmation_required") ||
		packet.CoreInvocationExecuted ||
		packet.DurableWriteByCore {
		t.Fatalf("objective closeout packet requiring confirmation = %#v", packet)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "adapter objective closeout packet",
		RunnerEffect: packet.RunnerEffect,
		PromptEffect: packet.PromptEffect,
		Boundaries:   packet.Boundaries,
		Payload:      packet,
	}, "production_adapter_objective_closeout_packet", "objective_closeout_projection_only", "host_owned_objective_closeout", "ready_for_host_objective_closeout_review", "host_objective_closeout_confirmation_required")
	AssertNoCoreMutation(t, "adapter objective closeout packet", packet.CoreInvocationExecuted, packet.DurableWriteByCore)
}

func TestProductionAdapterObjectiveCloseoutPacketConfirmsObjectiveSatisfied(t *testing.T) {
	auditBinding := productionAdapterReadyCompletionAuditResultBinding()

	packet := BuildProductionAdapterObjectiveCloseoutPacket(ProductionAdapterObjectiveCloseoutPacketInput{
		ObjectiveCloseoutPacketRef:   "closeout:metrics_objective",
		ObjectiveRef:                 "objective:metrics_monitoring",
		HostCloseoutConfirmationRef:  "confirmation:metrics_objective_closeout",
		CompletionAuditResultBinding: auditBinding,
	})
	if packet.Status != VerificationSatisfied ||
		!packet.ReadyForHostCloseoutReview ||
		!packet.ReadyForObjectiveCompletion ||
		!packet.ObjectiveSatisfied ||
		!packet.VerificationSatisfied ||
		!packet.HostCloseoutConfirmed ||
		!packet.AuthorizationBound ||
		!packet.CompletionAuditBound ||
		packet.FailureClass != FailureNone ||
		packet.NextHostAction != "return_objective_closeout" ||
		packet.ObjectiveCloseoutPacketRef != "closeout:metrics_objective" ||
		packet.ObjectiveRef != "objective:metrics_monitoring" ||
		packet.HostCloseoutConfirmationRef != "confirmation:metrics_objective_closeout" ||
		packet.CompletionAuditResultBindingRef != auditBinding.CompletionAuditResultBindingRef ||
		packet.CoreInvocationExecuted ||
		packet.DurableWriteByCore ||
		!productionAdapterEvidenceContains(packet.EvidenceRefs, "confirmation:metrics_objective_closeout", "host_objective_closeout_confirmation") ||
		!productionAdapterBoundaryContains(packet.Boundaries, "objective_closeout_confirmed") ||
		!productionAdapterBoundaryContains(packet.Boundaries, "objective_satisfied_by_host_closeout") ||
		!productionAdapterBoundaryContains(packet.Boundaries, "ready_for_objective_completion") {
		t.Fatalf("confirmed objective closeout packet = %#v", packet)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "adapter confirmed objective closeout packet",
		RunnerEffect: packet.RunnerEffect,
		PromptEffect: packet.PromptEffect,
		Boundaries:   packet.Boundaries,
		Payload:      packet,
	}, "production_adapter_objective_closeout_packet", "objective_closeout_projection_only", "host_owned_objective_closeout", "objective_closeout_confirmed", "objective_satisfied_by_host_closeout")
	AssertNoCoreMutation(t, "adapter confirmed objective closeout packet", packet.CoreInvocationExecuted, packet.DurableWriteByCore)
}

func TestProductionAdapterObjectiveCloseoutPacketRequiresReadyAuditBinding(t *testing.T) {
	review := productionAdapterBoundSuccessReadbackReview()
	handoff := BuildProductionAdapterCompletionHandoff(review)
	hostView := BuildProductionAdapterReadbackHostView(review, handoff)
	auditBinding := BuildProductionAdapterCompletionAuditResultBinding(ProductionAdapterCompletionAuditResultBindingInput{
		CompletionAuditResultBindingRef: "binding:metrics_completion_audit",
		CompletionAuditResultRef:        "audit:metrics_completion",
		AuthorizedHostView:              hostView,
		CompletionHandoff:               handoff,
		CompletionAuditResult: VerificationResult{
			Status:       VerificationFailed,
			FailureClass: FailureVerificationFailed,
			EvidenceRefs: []EvidenceRef{{
				Ref:      "audit:metrics_completion",
				Kind:     "completion_audit_result",
				Strength: EvidenceStrong,
				Source:   "handoff:metrics_completion",
			}},
		},
	})

	packet := BuildProductionAdapterObjectiveCloseoutPacket(ProductionAdapterObjectiveCloseoutPacketInput{
		ObjectiveCloseoutPacketRef:   "closeout:metrics_objective",
		ObjectiveRef:                 "objective:metrics_monitoring",
		HostCloseoutConfirmationRef:  "confirmation:metrics_objective_closeout",
		CompletionAuditResultBinding: auditBinding,
	})
	if packet.Status != VerificationBlocked ||
		packet.ReadyForHostCloseoutReview ||
		packet.ReadyForObjectiveCompletion ||
		packet.ObjectiveSatisfied ||
		packet.VerificationSatisfied ||
		packet.HostCloseoutConfirmed ||
		packet.FailureClass != FailureVerificationFailed ||
		!productionAdapterStringContains(packet.BlockedReasons, "completion_audit_result_binding_not_ready") ||
		!productionAdapterMissingContains(packet.MissingInputs, "host:completion_audit_result_binding") {
		t.Fatalf("expected not-ready audit binding to block objective closeout, got %#v", packet)
	}
}

func TestProductionAdapterObjectiveCloseoutPacketRejectsUnsafeRefs(t *testing.T) {
	auditBinding := productionAdapterReadyCompletionAuditResultBinding()

	packet := BuildProductionAdapterObjectiveCloseoutPacket(ProductionAdapterObjectiveCloseoutPacketInput{
		ObjectiveCloseoutPacketRef:   "closeout:metrics_objective",
		ObjectiveRef:                 "objective:metrics_monitoring",
		HostCloseoutConfirmationRef:  "/tmp/raw-closeout.json",
		CompletionAuditResultBinding: auditBinding,
	})
	if packet.Status != VerificationBlocked ||
		packet.ReadyForHostCloseoutReview ||
		packet.ReadyForObjectiveCompletion ||
		packet.ObjectiveSatisfied ||
		packet.VerificationSatisfied ||
		packet.HostCloseoutConfirmed ||
		packet.FailureClass != FailureEvidenceWeak ||
		!productionAdapterStringContains(packet.BlockedReasons, "unsafe_input_ref") ||
		!productionAdapterMissingContains(packet.MissingInputs, "host:display_safe_refs") {
		t.Fatalf("unsafe objective closeout packet = %#v", packet)
	}
	AssertNoRawPayload(t, "unsafe objective closeout packet", packet, "/tmp/raw-closeout.json")
}

func productionAdapterReadyCompletionAuditResultBinding() ProductionAdapterCompletionAuditResultBinding {
	review := productionAdapterBoundSuccessReadbackReview()
	handoff := BuildProductionAdapterCompletionHandoff(review)
	hostView := BuildProductionAdapterReadbackHostView(review, handoff)
	return BuildProductionAdapterCompletionAuditResultBinding(ProductionAdapterCompletionAuditResultBindingInput{
		CompletionAuditResultBindingRef: "binding:metrics_completion_audit",
		CompletionAuditResultRef:        "audit:metrics_completion",
		AuthorizedHostView:              hostView,
		CompletionHandoff:               handoff,
		CompletionAuditResult:           productionAdapterSatisfiedCompletionAuditResult(),
	})
}
