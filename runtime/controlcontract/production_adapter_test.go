package controlcontract

import (
	"testing"
)

func TestProductionAdapterDescriptorNormalizeAndClone(t *testing.T) {
	descriptor := productionAdapterTestDescriptor()
	descriptor.SupportedCandidateRefs = append(descriptor.SupportedCandidateRefs, "https://example.invalid/raw")
	descriptor.DisplaySafeInputRefs = append(descriptor.DisplaySafeInputRefs, "/tmp/raw-env")

	normalized := descriptor.Normalize()
	if normalized.ContractVersion != ContractVersion || !normalized.Projected {
		t.Fatalf("unexpected descriptor header: %#v", normalized)
	}
	if normalized.Kind != ProductionAdapterSourceApply ||
		normalized.AdapterRef != "adapter:operations_local_metrics" ||
		normalized.Owner != "scene" ||
		normalized.RunnerEffect != "none" ||
		normalized.PromptEffect != "none" {
		t.Fatalf("unexpected normalized descriptor: %#v", normalized)
	}
	if len(normalized.SupportedCandidateRefs) != 1 || normalized.SupportedCandidateRefs[0] != "strategy:operations_metric_collect" {
		t.Fatalf("candidate refs were not display-safe filtered: %#v", normalized.SupportedCandidateRefs)
	}
	if len(normalized.DisplaySafeInputRefs) != 1 || normalized.DisplaySafeInputRefs[0] != "input:metric_scope" {
		t.Fatalf("input refs were not display-safe filtered: %#v", normalized.DisplaySafeInputRefs)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "descriptor",
		RunnerEffect: normalized.RunnerEffect,
		PromptEffect: normalized.PromptEffect,
		Boundaries:   normalized.Boundaries,
		Payload:      normalized,
	}, "production_adapter_descriptor", "no_adapter_invocation")
	AssertNoRawPayload(t, "descriptor", normalized, "example.invalid", "/tmp/raw-env")

	clone := normalized.Clone()
	clone.SupportedCandidateRefs[0] = "strategy:changed"
	clone.DisplaySafeInputRefs[0] = "input:changed"
	if normalized.SupportedCandidateRefs[0] != "strategy:operations_metric_collect" ||
		normalized.DisplaySafeInputRefs[0] != "input:metric_scope" {
		t.Fatalf("clone mutated original descriptor: %#v", normalized)
	}
}

func TestProductionAdapterResolutionReadyAndMissingCapability(t *testing.T) {
	ready := BuildProductionAdapterResolution(productionAdapterResolutionInput(productionAdapterTestDescriptor()))
	if ready.Status != HostActionReady ||
		!ready.ReadyForHostPreflight ||
		ready.FailureClass != FailureNone ||
		ready.NextHostAction != "host_may_run_adapter_preflight" {
		t.Fatalf("ready resolution = %#v", ready)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "adapter resolution",
		RunnerEffect: ready.RunnerEffect,
		PromptEffect: ready.PromptEffect,
		Boundaries:   ready.Boundaries,
		Payload:      ready,
	}, "production_adapter_resolution", "no_adapter_invocation", "ready_for_host_preflight")

	missingCapabilityInput := productionAdapterResolutionInput(productionAdapterTestDescriptor())
	missingCapabilityInput.AvailableCapabilityRefs = nil
	missingCapability := BuildProductionAdapterResolution(missingCapabilityInput)
	if missingCapability.Status != HostActionBlocked ||
		missingCapability.ReadyForHostPreflight ||
		missingCapability.FailureClass != FailureCapabilityMissing ||
		!productionAdapterStringContains(missingCapability.BlockedReasons, "capability_missing") ||
		!productionAdapterMissingContains(missingCapability.MissingInputs, "capability:local_metrics") ||
		!productionAdapterBoundaryContains(missingCapability.Boundaries, "capability_gap_proposal_only") {
		t.Fatalf("missing capability resolution = %#v", missingCapability)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "missing capability resolution",
		RunnerEffect: missingCapability.RunnerEffect,
		PromptEffect: missingCapability.PromptEffect,
		Boundaries:   missingCapability.Boundaries,
		Payload:      missingCapability,
	}, "production_adapter_resolution", "capability_gap_proposal_only")
}

func TestProductionAdapterResolutionRequiresCompleteDescriptorAndMatchingBudget(t *testing.T) {
	incompleteDescriptor := productionAdapterTestDescriptor()
	incompleteDescriptor.SupportedSourceKinds = nil
	incompleteDescriptor.InputContractRef = ""
	incompleteDescriptor.RedactionPolicyRef = ""
	incompleteDescriptor.PreflightCheckRefs = nil
	incompleteDescriptor.IdempotencyContractRef = ""
	incompleteDescriptor.TimeoutPolicyRef = ""
	incompleteDescriptor.CompensationHandoffRef = ""
	input := productionAdapterResolutionInput(incompleteDescriptor)
	incomplete := BuildProductionAdapterResolution(input)
	if incomplete.Status != HostActionBlocked ||
		incomplete.ReadyForHostPreflight ||
		incomplete.FailureClass != FailureConfigMissing ||
		!productionAdapterStringContains(incomplete.BlockedReasons, "adapter_supported_source_kind_missing") ||
		!productionAdapterStringContains(incomplete.BlockedReasons, "adapter_input_contract_ref_missing") ||
		!productionAdapterStringContains(incomplete.BlockedReasons, "adapter_redaction_policy_ref_missing") ||
		!productionAdapterStringContains(incomplete.BlockedReasons, "adapter_preflight_check_ref_missing") ||
		!productionAdapterStringContains(incomplete.BlockedReasons, "adapter_idempotency_contract_ref_missing") ||
		!productionAdapterStringContains(incomplete.BlockedReasons, "adapter_timeout_policy_ref_missing") ||
		!productionAdapterStringContains(incomplete.BlockedReasons, "adapter_compensation_handoff_ref_missing") ||
		productionAdapterStringContains(incomplete.BlockedReasons, "adapter_source_mismatch") ||
		productionAdapterStringContains(incomplete.BlockedReasons, "idempotency_missing") ||
		productionAdapterStringContains(incomplete.BlockedReasons, "timeout_policy_missing") ||
		productionAdapterStringContains(incomplete.BlockedReasons, "compensation_handoff_missing") ||
		!productionAdapterMissingContains(incomplete.MissingInputs, "host:adapter_supported_source_kind") ||
		!productionAdapterMissingContains(incomplete.MissingInputs, "host:adapter_input_contract_ref") ||
		!productionAdapterMissingContains(incomplete.MissingInputs, "host:adapter_redaction_policy_ref") ||
		!productionAdapterMissingContains(incomplete.MissingInputs, "host:adapter_preflight_check_ref") ||
		!productionAdapterMissingContains(incomplete.MissingInputs, "host:adapter_idempotency_contract_ref") ||
		!productionAdapterMissingContains(incomplete.MissingInputs, "host:adapter_timeout_policy_ref") ||
		!productionAdapterMissingContains(incomplete.MissingInputs, "host:adapter_compensation_handoff_ref") {
		t.Fatalf("incomplete descriptor resolution = %#v", incomplete)
	}

	budgetMismatchInput := productionAdapterResolutionInput(productionAdapterTestDescriptor())
	budgetMismatchInput.BudgetRef = "budget:other"
	budgetMismatch := BuildProductionAdapterResolution(budgetMismatchInput)
	if budgetMismatch.Status != HostActionBlocked ||
		budgetMismatch.ReadyForHostPreflight ||
		budgetMismatch.FailureClass != FailurePolicyBlocked ||
		!productionAdapterStringContains(budgetMismatch.BlockedReasons, "budget_ref_mismatch") ||
		!productionAdapterMissingContains(budgetMismatch.MissingInputs, "budget:local_probe") {
		t.Fatalf("budget mismatch resolution = %#v", budgetMismatch)
	}
}

func TestProductionAdapterResolutionUnsafeRefsDoNotLeak(t *testing.T) {
	input := productionAdapterResolutionInput(productionAdapterTestDescriptor())
	input.RequestedAdapterRef = "service://user:pass@example.invalid/db"
	input.Descriptor.DisplaySafeOutputRefs = []DisplaySafeRef{"output:summary", "api_key=abc123"}

	resolution := BuildProductionAdapterResolution(input)
	if resolution.ReadyForHostPreflight ||
		resolution.FailureClass != FailureEvidenceWeak ||
		!productionAdapterStringContains(resolution.BlockedReasons, "unsafe_input_ref") ||
		!productionAdapterMissingContains(resolution.MissingInputs, "host:display_safe_refs") {
		t.Fatalf("unsafe resolution = %#v", resolution)
	}
	AssertNoRawPayload(t, "unsafe resolution", resolution, "service://", "api_key=abc123", "example.invalid")
}

func TestProductionAdapterPreflightAndInvocationProjection(t *testing.T) {
	preflight := productionAdapterReadyPreflight()
	if preflight.Status != HostActionReady ||
		!preflight.ReadyForHostInvocation ||
		preflight.NextHostAction != "host_may_invoke_adapter" {
		t.Fatalf("ready preflight = %#v", preflight)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "adapter preflight",
		RunnerEffect: preflight.RunnerEffect,
		PromptEffect: preflight.PromptEffect,
		Boundaries:   preflight.Boundaries,
		Payload:      preflight,
	}, "production_adapter_preflight", "no_adapter_invocation", "ready_for_host_invocation")

	invocation := BuildHostAdapterInvocationProjection(HostAdapterInvocationInput{
		Preflight:               preflight,
		InvocationRef:           "invocation:metrics_1",
		StartedEventRef:         "event:metrics_started",
		CompletedEventRef:       "event:metrics_completed",
		ResultRef:               "result:metrics_summary",
		ReadbackRef:             "readback:metrics_summary",
		CompletionHandoffRef:    "handoff:metrics_completion",
		HostInvocationCompleted: true,
	})
	if invocation.Status != HostActionRecorded ||
		!invocation.HostInvocationReported ||
		!invocation.HostInvocationCompleted ||
		invocation.HostInvocationFailed ||
		invocation.NextHostAction != "review_adapter_readback" {
		t.Fatalf("success invocation projection = %#v", invocation)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "adapter invocation",
		RunnerEffect: invocation.RunnerEffect,
		PromptEffect: invocation.PromptEffect,
		Boundaries:   invocation.Boundaries,
		Payload:      invocation,
	}, "host_adapter_invocation_projection", "host_owned_invocation", "core_invocation_not_executed", "no_durable_write_by_core")
	AssertNoCoreMutation(t, "adapter invocation", invocation.CoreInvocationExecuted, invocation.DurableWriteByCore)
}

func TestProductionAdapterReadbackReviewAndCompletionHandoff(t *testing.T) {
	review := productionAdapterBoundSuccessReadbackReview()
	if review.Status != HostActionReady ||
		!review.ReadyForReadbackReview ||
		!review.ReadyForCompletionAudit ||
		review.ReadyForFailureReview ||
		!review.AuthorizationBound ||
		review.ObjectiveSatisfied ||
		review.CoreInvocationExecuted ||
		review.DurableWriteByCore ||
		review.NextHostAction != "review_completion_handoff" ||
		!productionAdapterEvidenceContains(review.EvidenceRefs, review.InvocationReportBindingRef, "invocation_report_binding") ||
		!productionAdapterEvidenceContains(review.EvidenceRefs, "result:metrics_summary", "adapter_result") ||
		!productionAdapterEvidenceContains(review.EvidenceRefs, "readback:metrics_summary", "adapter_readback") ||
		!productionAdapterEvidenceContains(review.EvidenceRefs, "handoff:metrics_completion", "completion_handoff") {
		t.Fatalf("readback review = %#v", review)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "adapter readback review",
		RunnerEffect: review.RunnerEffect,
		PromptEffect: review.PromptEffect,
		Boundaries:   review.Boundaries,
		Payload:      review,
	}, "production_adapter_readback_review", "readback_review_from_invocation_report_binding", "authorization_bound_readback_review", "completion_handoff_projection_only", "objective_not_satisfied_by_adapter", "no_runner_dispatch")

	handoff := BuildProductionAdapterCompletionHandoff(review)
	if handoff.Status != VerificationReviewRequired ||
		!handoff.ReadyForCompletionAudit ||
		!handoff.AuthorizationBound ||
		handoff.InvocationReportBindingRef != review.InvocationReportBindingRef ||
		handoff.AuthorizationPacketRef != review.AuthorizationPacketRef ||
		handoff.PreflightReviewPacketRef != review.PreflightReviewPacketRef ||
		handoff.ObjectiveSatisfied ||
		handoff.Verification.Satisfied ||
		handoff.Verification.Status != VerificationReviewRequired ||
		handoff.FailureClass != FailureEvidenceMissing ||
		handoff.NextHostAction != "run_completion_audit" ||
		!productionAdapterMissingContains(handoff.MissingInputs, "host:completion_audit_result") ||
		!productionAdapterMissingContains(handoff.Verification.MissingInputs, "host:completion_audit_result") ||
		!productionAdapterEvidenceContains(handoff.EvidenceRefs, "handoff:metrics_completion", "completion_handoff") {
		t.Fatalf("completion handoff = %#v", handoff)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "adapter completion handoff",
		RunnerEffect: handoff.RunnerEffect,
		PromptEffect: handoff.PromptEffect,
		Boundaries:   handoff.Boundaries,
		Payload:      handoff,
	}, "production_adapter_completion_handoff", "completion_handoff_from_authorization_bound_readback", "authorization_bound_completion_handoff", "completion_audit_input_only", "objective_not_satisfied_by_adapter", "no_runner_dispatch", "no_durable_write_by_core")
	AssertHostOwnedAuditInputOnly(t, Projection[Boundary]{
		Name:         "adapter completion handoff",
		RunnerEffect: handoff.RunnerEffect,
		PromptEffect: handoff.PromptEffect,
		Boundaries:   handoff.Boundaries,
		Payload:      handoff,
	}, handoff.ObjectiveSatisfied, handoff.Verification.Satisfied, false, false, "production_adapter_completion_handoff", "completion_handoff_from_authorization_bound_readback", "authorization_bound_completion_handoff", "completion_audit_input_only", "objective_not_satisfied_by_adapter")
}

func TestProductionAdapterCompletionHandoffRequiresAuthorizationBoundReadback(t *testing.T) {
	review := BuildProductionAdapterReadbackReview(productionAdapterSuccessInvocation())
	if !review.ReadyForCompletionAudit || review.AuthorizationBound {
		t.Fatalf("expected direct readback projection to be completion-ready but not authorization-bound, got %#v", review)
	}

	handoff := BuildProductionAdapterCompletionHandoff(review)
	if handoff.Status != VerificationBlocked ||
		handoff.ReadyForCompletionAudit ||
		handoff.AuthorizationBound ||
		handoff.ObjectiveSatisfied ||
		handoff.Verification.Satisfied ||
		handoff.FailureClass != FailureAuthorizationMissing ||
		handoff.NextHostAction != "review_adapter_readback_authorization" ||
		!productionAdapterStringContains(handoff.BlockedReasons, "adapter_readback_not_authorization_bound") ||
		!productionAdapterMissingContains(handoff.MissingInputs, "host:authorization_bound_readback_review") {
		t.Fatalf("expected unbound readback to block completion handoff, got %#v", handoff)
	}
}

func TestProductionAdapterReadbackHostViewShowsCompletionAuditInput(t *testing.T) {
	review := productionAdapterBoundSuccessReadbackReview()
	handoff := BuildProductionAdapterCompletionHandoff(review)
	view := BuildProductionAdapterReadbackHostView(review, handoff)
	if !view.Available ||
		view.Status != "ready_for_completion_audit" ||
		view.Mode != "production_adapter_readback_host_view" ||
		view.ReadbackStatus != HostActionReady ||
		view.HandoffStatus != VerificationReviewRequired ||
		!view.ReadyForHostReview ||
		!view.ReadyForReadbackReview ||
		view.ReadyForFailureReview ||
		!view.ReadyForCompletionAudit ||
		!view.AuthorizationBound ||
		view.AuthorizationChainStatus != "authorization_bound" ||
		view.InvocationReportBindingRef != review.InvocationReportBindingRef ||
		view.AuthorizationPacketRef != review.AuthorizationPacketRef ||
		view.PreflightReviewPacketRef != review.PreflightReviewPacketRef ||
		len(view.AuthorizationMissingInputs) != 0 ||
		view.ObjectiveSatisfied ||
		view.VerificationSatisfied ||
		view.Verification.Satisfied ||
		view.CoreInvocationExecuted ||
		view.DurableWriteByCore ||
		view.NextHostAction != "run_completion_audit" ||
		view.FailureClass != FailureEvidenceMissing ||
		!productionAdapterMissingContains(view.MissingInputs, "host:completion_audit_result") ||
		!productionAdapterEvidenceContains(view.EvidenceRefs, "handoff:metrics_completion", "completion_handoff") {
		t.Fatalf("completion audit host view = %#v", view)
	}
	AssertHostOwnedAuditInputOnly(t, Projection[Boundary]{
		Name:         "adapter readback host view",
		RunnerEffect: view.RunnerEffect,
		PromptEffect: view.PromptEffect,
		Boundaries:   view.Boundaries,
		Payload:      view,
	}, view.ObjectiveSatisfied, view.VerificationSatisfied, view.CoreInvocationExecuted, view.DurableWriteByCore, "production_adapter_readback_host_view", "host_adapter_result_view_only", "authorization_bound_host_view", "authorized_lifecycle_closeout_view", "completion_audit_input_only", "objective_not_satisfied_by_adapter")
}

func TestProductionAdapterReadbackHostViewShowsAuthorizationMissingInputs(t *testing.T) {
	review := BuildProductionAdapterReadbackReview(productionAdapterSuccessInvocation())
	handoff := BuildProductionAdapterCompletionHandoff(review)
	view := BuildProductionAdapterReadbackHostView(review, handoff)

	if !view.Available ||
		view.Status != "blocked" ||
		view.ReadyForHostReview ||
		view.ReadyForCompletionAudit ||
		view.AuthorizationBound ||
		view.AuthorizationChainStatus != "authorization_required" ||
		view.NextHostAction != "review_adapter_readback_authorization" ||
		view.FailureClass != FailureAuthorizationMissing ||
		!productionAdapterMissingContains(view.MissingInputs, "host:authorization_bound_readback_review") ||
		!productionAdapterMissingContains(view.AuthorizationMissingInputs, "host:authorization_bound_readback_review") ||
		!productionAdapterMissingContains(view.AuthorizationMissingInputs, "host:invocation_report_binding") ||
		!productionAdapterMissingContains(view.AuthorizationMissingInputs, "host:invocation_authorization_packet") ||
		!productionAdapterMissingContains(view.AuthorizationMissingInputs, "host:adapter_preflight_review_packet") {
		t.Fatalf("expected host view to expose authorization missing inputs, got %#v", view)
	}
}

func TestProductionAdapterReadbackHostViewShowsFailureReview(t *testing.T) {
	review := BuildProductionAdapterReadbackReview(productionAdapterFailureInvocation())
	handoff := BuildProductionAdapterCompletionHandoff(review)
	view := BuildProductionAdapterReadbackHostView(review, handoff)
	if !view.Available ||
		view.Status != "ready_for_failure_review" ||
		!view.ReadyForHostReview ||
		view.ReadyForReadbackReview ||
		!view.ReadyForFailureReview ||
		view.ReadyForCompletionAudit ||
		view.ObjectiveSatisfied ||
		view.VerificationSatisfied ||
		view.FailureClass != FailureVerificationFailed ||
		view.NextHostAction != "review_adapter_failure" ||
		view.FailureRef != "failure:metrics_failed" ||
		view.CompensationRef != "compensation:metrics_review" ||
		view.ResultRef != "" ||
		view.ReadbackRef != "" ||
		view.CompletionHandoffRef != "" {
		t.Fatalf("failure host view = %#v", view)
	}
	AssertHostOwnedAuditInputOnly(t, Projection[Boundary]{
		Name:         "adapter failure host view",
		RunnerEffect: view.RunnerEffect,
		PromptEffect: view.PromptEffect,
		Boundaries:   view.Boundaries,
		Payload:      view,
	}, view.ObjectiveSatisfied, view.VerificationSatisfied, view.CoreInvocationExecuted, view.DurableWriteByCore, "production_adapter_readback_host_view", "host_adapter_result_view_only", "readback_projection_only", "objective_not_satisfied_by_adapter")
}

func TestProductionAdapterReadbackHostViewUnsafeRefsDoNotLeak(t *testing.T) {
	review := productionAdapterBoundSuccessReadbackReview()
	handoff := BuildProductionAdapterCompletionHandoff(review)
	handoff.CompletionHandoffRef = "/tmp/raw-completion-handoff.json"
	view := BuildProductionAdapterReadbackHostView(review, handoff)
	if !view.Available ||
		view.Status != "blocked" ||
		view.ReadyForHostReview ||
		view.ReadyForCompletionAudit ||
		view.HostInvocationReported ||
		view.FailureClass != FailureEvidenceWeak ||
		!productionAdapterStringContains(view.BlockedReasons, "unsafe_input_ref") ||
		!productionAdapterMissingContains(view.MissingInputs, "host:display_safe_refs") {
		t.Fatalf("unsafe host view = %#v", view)
	}
	AssertNoRawPayload(t, "unsafe adapter readback host view", view, "/tmp/raw-completion-handoff.json")
}

func TestProductionAdapterReadbackReviewFailureDoesNotEnterCompletionAudit(t *testing.T) {
	invocation := productionAdapterFailureInvocation()
	review := BuildProductionAdapterReadbackReview(invocation)
	if review.Status != HostActionReviewRequired ||
		review.ReadyForReadbackReview ||
		review.ReadyForCompletionAudit ||
		!review.ReadyForFailureReview ||
		review.ObjectiveSatisfied ||
		review.FailureClass != FailureVerificationFailed ||
		review.NextHostAction != "review_adapter_failure" ||
		!productionAdapterEvidenceContains(review.EvidenceRefs, "failure:metrics_failed", "adapter_failure") ||
		!productionAdapterEvidenceContains(review.EvidenceRefs, "compensation:metrics_review", "adapter_compensation") {
		t.Fatalf("failure readback review = %#v", review)
	}

	handoff := BuildProductionAdapterCompletionHandoff(review)
	if handoff.Status != VerificationBlocked ||
		handoff.ReadyForCompletionAudit ||
		handoff.ObjectiveSatisfied ||
		handoff.Verification.Satisfied ||
		handoff.FailureClass != FailureVerificationFailed ||
		!productionAdapterStringContains(handoff.BlockedReasons, "adapter_readback_not_ready") ||
		!productionAdapterMissingContains(handoff.MissingInputs, "host:production_adapter_readback_review") {
		t.Fatalf("failure completion handoff = %#v", handoff)
	}
}

func TestProductionAdapterReadbackReviewUnsafeRefsDoNotLeak(t *testing.T) {
	invocation := productionAdapterSuccessInvocation()
	invocation.ReadbackRef = "/tmp/raw-adapter-readback.json"
	review := BuildProductionAdapterReadbackReview(invocation)
	if review.Status != HostActionBlocked ||
		review.ReadyForReadbackReview ||
		review.ReadyForCompletionAudit ||
		review.FailureClass != FailureEvidenceWeak ||
		!productionAdapterStringContains(review.BlockedReasons, "unsafe_input_ref") ||
		!productionAdapterMissingContains(review.MissingInputs, "host:display_safe_refs") {
		t.Fatalf("unsafe readback review = %#v", review)
	}
	AssertNoRawPayload(t, "unsafe adapter readback review", review, "/tmp/raw-adapter-readback.json")
}

func TestProductionAdapterCompletionHandoffUnsafeRefsDoNotLeak(t *testing.T) {
	review := productionAdapterBoundSuccessReadbackReview()
	review.CompletionHandoffRef = "/tmp/raw-completion-handoff.json"
	handoff := BuildProductionAdapterCompletionHandoff(review)
	if handoff.Status != VerificationBlocked ||
		handoff.ReadyForCompletionAudit ||
		handoff.FailureClass != FailureEvidenceWeak ||
		!productionAdapterStringContains(handoff.BlockedReasons, "unsafe_input_ref") ||
		!productionAdapterMissingContains(handoff.MissingInputs, "host:display_safe_refs") {
		t.Fatalf("unsafe completion handoff = %#v", handoff)
	}
	AssertNoRawPayload(t, "unsafe adapter completion handoff", handoff, "/tmp/raw-completion-handoff.json")
}

func TestProductionAdapterPreflightAndInvocationShortCircuitUpstreamBlocked(t *testing.T) {
	blockedResolutionInput := productionAdapterResolutionInput(productionAdapterTestDescriptor())
	blockedResolutionInput.AvailableCapabilityRefs = nil
	blockedResolution := BuildProductionAdapterResolution(blockedResolutionInput)
	preflight := BuildProductionAdapterPreflight(ProductionAdapterPreflightInput{
		Resolution: blockedResolution,
	})
	if preflight.Status != HostActionBlocked ||
		preflight.ReadyForHostInvocation ||
		preflight.FailureClass != FailureConfigMissing ||
		!productionAdapterStringContains(preflight.BlockedReasons, "adapter_resolution_not_ready") ||
		productionAdapterStringContains(preflight.BlockedReasons, "adapter_unavailable") ||
		!productionAdapterMissingContains(preflight.MissingInputs, "host:adapter_resolution") {
		t.Fatalf("blocked upstream preflight = %#v", preflight)
	}

	blockedPreflight := BuildProductionAdapterPreflight(ProductionAdapterPreflightInput{
		Resolution:             BuildProductionAdapterResolution(productionAdapterResolutionInput(productionAdapterTestDescriptor())),
		AdapterAvailable:       false,
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
	invocation := BuildHostAdapterInvocationProjection(HostAdapterInvocationInput{
		Preflight: blockedPreflight,
	})
	if invocation.Status != HostActionBlocked ||
		invocation.HostInvocationReported ||
		invocation.CoreInvocationExecuted ||
		invocation.DurableWriteByCore ||
		invocation.FailureClass != FailureConfigMissing ||
		!productionAdapterStringContains(invocation.BlockedReasons, "adapter_preflight_not_ready") ||
		productionAdapterStringContains(invocation.BlockedReasons, "invocation_ref_missing") ||
		!productionAdapterMissingContains(invocation.MissingInputs, "host:adapter_preflight") {
		t.Fatalf("blocked upstream invocation = %#v", invocation)
	}
	AssertNoCoreMutation(t, "blocked upstream invocation", invocation.CoreInvocationExecuted, invocation.DurableWriteByCore)
}

func TestHostAdapterInvocationConflictBlocked(t *testing.T) {
	preflight := productionAdapterReadyPreflight()
	conflict := BuildHostAdapterInvocationProjection(HostAdapterInvocationInput{
		Preflight:               preflight,
		InvocationRef:           "invocation:metrics_1",
		StartedEventRef:         "event:metrics_started",
		CompletedEventRef:       "event:metrics_completed",
		ResultRef:               "result:metrics_summary",
		FailureRef:              "failure:metrics_failed",
		ReadbackRef:             "readback:metrics_summary",
		CompensationRef:         "compensation:metrics_review",
		CompletionHandoffRef:    "handoff:metrics_completion",
		HostInvocationCompleted: true,
		HostInvocationFailed:    true,
	})
	if conflict.Status != HostActionBlocked ||
		conflict.HostInvocationReported ||
		conflict.CoreInvocationExecuted ||
		conflict.DurableWriteByCore ||
		conflict.FailureClass != FailureVerificationFailed ||
		!productionAdapterStringContains(conflict.BlockedReasons, "invocation_result_conflict") {
		t.Fatalf("conflict invocation projection = %#v", conflict)
	}
	AssertNoCoreMutation(t, "conflict invocation", conflict.CoreInvocationExecuted, conflict.DurableWriteByCore)
}

func productionAdapterReadyPreflight() ProductionAdapterPreflight {
	return BuildProductionAdapterPreflight(ProductionAdapterPreflightInput{
		Resolution:             BuildProductionAdapterResolution(productionAdapterResolutionInput(productionAdapterTestDescriptor())),
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
}

func productionAdapterSuccessInvocation() HostAdapterInvocationProjection {
	return BuildHostAdapterInvocationProjection(HostAdapterInvocationInput{
		Preflight:               productionAdapterReadyPreflight(),
		InvocationRef:           "invocation:metrics_1",
		StartedEventRef:         "event:metrics_started",
		CompletedEventRef:       "event:metrics_completed",
		ResultRef:               "result:metrics_summary",
		ReadbackRef:             "readback:metrics_summary",
		CompletionHandoffRef:    "handoff:metrics_completion",
		HostInvocationCompleted: true,
	})
}

func productionAdapterBoundSuccessReadbackReview() ProductionAdapterReadbackReview {
	authorization := productionAdapterReadyInvocationAuthorizationPacket()
	binding := BuildProductionAdapterInvocationReportBinding(ProductionAdapterInvocationReportBindingInput{
		InvocationReportBindingRef: "binding:metrics_invocation_report",
		AuthorizationPacket:        authorization,
		Invocation:                 productionAdapterAuthorizedSuccessInvocation(authorization),
	})
	return BuildProductionAdapterReadbackReviewFromBinding(binding)
}

func productionAdapterFailureInvocation() HostAdapterInvocationProjection {
	return BuildHostAdapterInvocationProjection(HostAdapterInvocationInput{
		Preflight:            productionAdapterReadyPreflight(),
		InvocationRef:        "invocation:metrics_1",
		StartedEventRef:      "event:metrics_started",
		CompletedEventRef:    "event:metrics_failed",
		FailureRef:           "failure:metrics_failed",
		CompensationRef:      "compensation:metrics_review",
		HostInvocationFailed: true,
	})
}

func productionAdapterTestDescriptor() ProductionAdapterDescriptor {
	return ProductionAdapterDescriptor{
		AdapterRef:             "adapter:operations_local_metrics",
		Owner:                  "scene",
		OwnerRef:               "scene:agentx_operations",
		Version:                "v1",
		Kind:                   ProductionAdapterSourceApply,
		SupportedSourceKinds:   []ReplannerSourceKind{ReplannerSourceOperations},
		SupportedCandidateRefs: []DisplaySafeRef{"strategy:operations_metric_collect"},
		ProvidesCapabilityRefs: []DisplaySafeRef{"capability:metric_observation"},
		RequiresCapabilityRefs: []DisplaySafeRef{"capability:local_metrics"},
		InputContractRef:       "contract:metrics_input",
		OutputContractRef:      "contract:metrics_output",
		ReadbackContractRef:    "contract:metrics_readback",
		RequiredPolicyRefs:     []DisplaySafeRef{"policy:local_readonly"},
		RequiredApprovalRefs:   []DisplaySafeRef{"approval:local_readonly"},
		RequiredBudgetRef:      "budget:local_probe",
		IdempotencyContractRef: "idempotency:metrics_probe",
		RiskRef:                "risk:local_readonly",
		SideEffectClass:        "read_only",
		TimeoutPolicyRef:       "timeout:short",
		CompensationHandoffRef: "compensation:manual_review",
		RedactionPolicyRef:     "redaction:display_safe",
		PreflightCheckRefs:     []DisplaySafeRef{"preflight:metrics_ready"},
		DisplaySafeInputRefs:   []DisplaySafeRef{"input:metric_scope"},
		DisplaySafeOutputRefs:  []DisplaySafeRef{"output:summary"},
	}
}

func productionAdapterResolutionInput(descriptor ProductionAdapterDescriptor) ProductionAdapterResolutionInput {
	return ProductionAdapterResolutionInput{
		ApplyEnvelopeReady:      true,
		ApplyEnvelopeRef:        "envelope:metrics_apply",
		SelectedSourceKind:      ReplannerSourceOperations,
		SelectedSourceRef:       "source:operations_metrics",
		SelectedCandidateRef:    "strategy:operations_metric_collect",
		RequestedAdapterRef:     "adapter:operations_local_metrics",
		CatalogSnapshotRef:      "catalog:host_adapters",
		HostPolicyRef:           "policy:local_readonly",
		ApprovalContextRefs:     []DisplaySafeRef{"approval:local_readonly"},
		BudgetRef:               "budget:local_probe",
		IdempotencyRef:          "idempotency:metrics_probe_1",
		AvailableCapabilityRefs: []DisplaySafeRef{"capability:local_metrics"},
		ConfirmedPolicyRefs:     []DisplaySafeRef{"policy:local_readonly"},
		ConfirmedApprovalRefs:   []DisplaySafeRef{"approval:local_readonly"},
		Descriptor:              descriptor,
	}
}

func productionAdapterEvidenceContains(values []EvidenceRef, wantRef DisplaySafeRef, wantKind string) bool {
	for _, value := range values {
		if value.Ref == wantRef && value.Kind == wantKind {
			return true
		}
	}
	return false
}
