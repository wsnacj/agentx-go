package controlcontract

import (
	"testing"
)

func TestProductionAdapterCompletionAuditResultBindingReadyForObjectiveCloseout(t *testing.T) {
	review := productionAdapterBoundSuccessReadbackReview()
	handoff := BuildProductionAdapterCompletionHandoff(review)
	hostView := BuildProductionAdapterReadbackHostView(review, handoff)
	audit := productionAdapterSatisfiedCompletionAuditResult()

	binding := BuildProductionAdapterCompletionAuditResultBinding(ProductionAdapterCompletionAuditResultBindingInput{
		CompletionAuditResultBindingRef: "binding:metrics_completion_audit",
		CompletionAuditResultRef:        "audit:metrics_completion",
		AuthorizedHostView:              hostView,
		CompletionHandoff:               handoff,
		CompletionAuditResult:           audit,
	})
	if binding.Status != VerificationSatisfied ||
		!binding.ReadyForObjectiveCloseout ||
		binding.ObjectiveSatisfied ||
		!binding.VerificationSatisfied ||
		!binding.AuthorizationBound ||
		!binding.CompletionAuditBound ||
		binding.CoreInvocationExecuted ||
		binding.DurableWriteByCore ||
		binding.CompletionAuditResultBindingRef != "binding:metrics_completion_audit" ||
		binding.CompletionAuditResultRef != "audit:metrics_completion" ||
		binding.CompletionHandoffRef != handoff.CompletionHandoffRef ||
		binding.InvocationReportBindingRef != handoff.InvocationReportBindingRef ||
		binding.AuthorizationPacketRef != handoff.AuthorizationPacketRef ||
		binding.PreflightReviewPacketRef != handoff.PreflightReviewPacketRef ||
		binding.NextHostAction != "review_objective_closeout" ||
		binding.FailureClass != FailureNone ||
		len(binding.MissingInputs) != 0 ||
		len(binding.BlockedReasons) != 0 ||
		!binding.Verification.Satisfied ||
		!productionAdapterEvidenceContains(binding.EvidenceRefs, "audit:metrics_completion", "completion_audit_result") ||
		!productionAdapterBoundaryContains(binding.Boundaries, "completion_audit_result_bound") ||
		!productionAdapterBoundaryContains(binding.Boundaries, "ready_for_objective_closeout") ||
		!productionAdapterBoundaryContains(binding.Boundaries, "host_owned_completion_audit") {
		t.Fatalf("completion audit result binding = %#v", binding)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "adapter completion audit result binding",
		RunnerEffect: binding.RunnerEffect,
		PromptEffect: binding.PromptEffect,
		Boundaries:   binding.Boundaries,
		Payload:      binding,
	}, "production_adapter_completion_audit_result_binding", "completion_audit_result_binding_projection_only", "host_owned_completion_audit", "objective_closeout_input_only", "ready_for_objective_closeout")
	AssertNoCoreMutation(t, "adapter completion audit result binding", binding.CoreInvocationExecuted, binding.DurableWriteByCore)
	AssertNoObjectiveSatisfied(t, "adapter completion audit result binding", binding.ObjectiveSatisfied)
}

func TestProductionAdapterCompletionAuditResultBindingRequiresAuthorizedHostView(t *testing.T) {
	review := BuildProductionAdapterReadbackReview(productionAdapterSuccessInvocation())
	handoff := BuildProductionAdapterCompletionHandoff(review)
	hostView := BuildProductionAdapterReadbackHostView(review, handoff)

	binding := BuildProductionAdapterCompletionAuditResultBinding(ProductionAdapterCompletionAuditResultBindingInput{
		CompletionAuditResultBindingRef: "binding:metrics_completion_audit",
		CompletionAuditResultRef:        "audit:metrics_completion",
		AuthorizedHostView:              hostView,
		CompletionHandoff:               handoff,
		CompletionAuditResult:           productionAdapterSatisfiedCompletionAuditResult(),
	})
	if binding.Status != VerificationBlocked ||
		binding.ReadyForObjectiveCloseout ||
		binding.ObjectiveSatisfied ||
		binding.VerificationSatisfied ||
		binding.AuthorizationBound ||
		binding.CompletionAuditBound ||
		binding.FailureClass != FailureAuthorizationMissing ||
		!productionAdapterStringContains(binding.BlockedReasons, "authorized_host_view_not_ready") ||
		!productionAdapterStringContains(binding.BlockedReasons, "completion_handoff_not_ready") ||
		!productionAdapterMissingContains(binding.MissingInputs, "host:authorized_adapter_host_view") ||
		!productionAdapterMissingContains(binding.MissingInputs, "host:production_adapter_completion_handoff") {
		t.Fatalf("expected unbound host view to block completion audit binding, got %#v", binding)
	}
}

func TestProductionAdapterCompletionAuditResultBindingRejectsUnsatisfiedAudit(t *testing.T) {
	review := productionAdapterBoundSuccessReadbackReview()
	handoff := BuildProductionAdapterCompletionHandoff(review)
	hostView := BuildProductionAdapterReadbackHostView(review, handoff)
	audit := productionAdapterSatisfiedCompletionAuditResult()
	audit.Status = VerificationFailed
	audit.FailureClass = FailureVerificationFailed
	audit.Findings = []string{"completion audit failed"}

	binding := BuildProductionAdapterCompletionAuditResultBinding(ProductionAdapterCompletionAuditResultBindingInput{
		CompletionAuditResultBindingRef: "binding:metrics_completion_audit",
		CompletionAuditResultRef:        "audit:metrics_completion",
		AuthorizedHostView:              hostView,
		CompletionHandoff:               handoff,
		CompletionAuditResult:           audit,
	})
	if binding.Status != VerificationFailed ||
		binding.ReadyForObjectiveCloseout ||
		binding.ObjectiveSatisfied ||
		binding.VerificationSatisfied ||
		!binding.CompletionAuditBound ||
		binding.FailureClass != FailureVerificationFailed ||
		binding.NextHostAction != "review_completion_audit_result" ||
		!productionAdapterStringContains(binding.BlockedReasons, "completion_audit_not_satisfied") ||
		!productionAdapterBoundaryContains(binding.Boundaries, "completion_audit_result_not_satisfied") ||
		!productionAdapterBoundaryContains(binding.Boundaries, "completion_audit_result_bound") {
		t.Fatalf("unsatisfied audit binding = %#v", binding)
	}
}

func TestProductionAdapterCompletionAuditResultBindingRejectsUnsafeAuditRefs(t *testing.T) {
	review := productionAdapterBoundSuccessReadbackReview()
	handoff := BuildProductionAdapterCompletionHandoff(review)
	hostView := BuildProductionAdapterReadbackHostView(review, handoff)
	audit := productionAdapterSatisfiedCompletionAuditResult()
	audit.FailureReason = "/tmp/raw-audit.json"

	binding := BuildProductionAdapterCompletionAuditResultBinding(ProductionAdapterCompletionAuditResultBindingInput{
		CompletionAuditResultBindingRef: "binding:metrics_completion_audit",
		CompletionAuditResultRef:        "audit:metrics_completion",
		AuthorizedHostView:              hostView,
		CompletionHandoff:               handoff,
		CompletionAuditResult:           audit,
	})
	if binding.Status != VerificationBlocked ||
		binding.ReadyForObjectiveCloseout ||
		binding.ObjectiveSatisfied ||
		binding.VerificationSatisfied ||
		binding.CompletionAuditBound ||
		binding.FailureClass != FailureEvidenceWeak ||
		!productionAdapterStringContains(binding.BlockedReasons, "unsafe_input_ref") ||
		!productionAdapterMissingContains(binding.MissingInputs, "host:display_safe_refs") {
		t.Fatalf("unsafe audit binding = %#v", binding)
	}
	AssertNoRawPayload(t, "unsafe completion audit result binding", binding, "/tmp/raw-audit.json")
}

func productionAdapterSatisfiedCompletionAuditResult() VerificationResult {
	return VerificationResult{
		ContractVersion: ContractVersion,
		Status:          VerificationSatisfied,
		FailureClass:    FailureNone,
		EvidenceRefs: []EvidenceRef{{
			Ref:      "audit:metrics_completion",
			Kind:     "completion_audit_result",
			Strength: EvidenceStrong,
			Source:   "handoff:metrics_completion",
		}},
		Boundaries: []Boundary{"host_completion_audit"},
		Findings:   []string{"completion audit satisfied"},
	}.Normalize()
}
