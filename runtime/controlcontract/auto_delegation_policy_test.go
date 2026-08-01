package controlcontract

import "testing"

func TestAutoDelegationPolicyReviewDefaultOff(t *testing.T) {
	review := BuildAutoDelegationPolicyReview(AutoDelegationPolicy{})

	if review.Status != VerificationNotApplicable ||
		review.FailureClass != FailureNone ||
		review.Ready ||
		review.AutoDelegationAllowed ||
		review.ManagedExecutionAllowed ||
		review.NextHostAction != "enable_auto_delegation_policy" {
		t.Fatalf("unexpected default-off review: %+v", review)
	}
	if !autoDelegationBoundaryContains(review.Boundaries, "auto_delegation_default_off") ||
		!autoDelegationBoundaryContains(review.Boundaries, "no_subagent_dispatch") {
		t.Fatalf("default review should preserve projection-only boundaries: %+v", review.Boundaries)
	}
}

func TestAutoDelegationPolicyReviewProposeIsPlanOnly(t *testing.T) {
	review := BuildAutoDelegationPolicyReview(AutoDelegationPolicy{
		Mode: AutoDelegationPropose,
	})

	if review.Status != VerificationSatisfied ||
		!review.Ready ||
		!review.AutoDelegationAllowed ||
		!review.PlanOnly ||
		review.ManagedExecutionAllowed ||
		review.NextHostAction != "review_auto_delegation_proposal" {
		t.Fatalf("unexpected proposal review: %+v", review)
	}
	if review.Policy.MaxChildren != DefaultAutoDelegationMaxChildren ||
		review.Policy.MaxAttemptsPerChild != DefaultAutoDelegationMaxAttemptsPerChild ||
		!review.Policy.RequireEvidence ||
		!review.Policy.RequireVerification {
		t.Fatalf("proposal policy defaults were not normalized: %+v", review.Policy)
	}
}

func TestAutoDelegationPolicyReviewManagedReadOnlyDefaults(t *testing.T) {
	review := BuildAutoDelegationPolicyReview(AutoDelegationPolicy{
		Enabled:        true,
		MaxChildren:    1,
		MaxParallelism: 5,
	})

	if review.Status != VerificationSatisfied ||
		!review.Ready ||
		!review.AutoDelegationAllowed ||
		!review.ManagedExecutionAllowed ||
		!review.ReadOnlyOnly ||
		review.PlanOnly ||
		review.Policy.Mode != AutoDelegationManagedReadOnly ||
		review.Policy.MaxParallelism != 1 ||
		review.Policy.AllowedSideEffectPolicy != ObjectiveSpecSideEffectReadOnly {
		t.Fatalf("unexpected managed-readonly review: %+v", review)
	}
	if review.RunnerEffect != "none" || !autoDelegationBoundaryContains(review.Boundaries, "no_child_task_spawn") {
		t.Fatalf("policy review must remain projection-only: %+v", review)
	}
}

func TestAutoDelegationPolicyReviewManagedRequiresApproval(t *testing.T) {
	review := BuildAutoDelegationPolicyReview(AutoDelegationPolicy{
		Mode: AutoDelegationManaged,
	})

	if review.Status != VerificationReviewRequired ||
		review.FailureClass != FailureApprovalRequired ||
		review.Ready ||
		review.AutoDelegationAllowed ||
		review.ManagedExecutionAllowed ||
		review.NextHostAction != "request_auto_delegation_approval" {
		t.Fatalf("managed mode without approval should be review-required: %+v", review)
	}
	if !autoDelegationMissingInputContains(review.MissingInputs, "host:auto_delegation_approval_ref") {
		t.Fatalf("approval ref missing input not reported: %+v", review.MissingInputs)
	}
}

func TestAutoDelegationPolicyReviewManagedWithApproval(t *testing.T) {
	review := BuildAutoDelegationPolicyReview(AutoDelegationPolicy{
		Mode:                 AutoDelegationManaged,
		RequiredApprovalRefs: []DisplaySafeRef{"approval:auto_delegation"},
	})

	if review.Status != VerificationSatisfied ||
		!review.Ready ||
		!review.AutoDelegationAllowed ||
		!review.ManagedExecutionAllowed ||
		review.ReadOnlyOnly ||
		review.Policy.AllowedSideEffectPolicy != ObjectiveSpecSideEffectRequiresApproval {
		t.Fatalf("unexpected managed approval review: %+v", review)
	}
}

func TestAutoDelegationPolicyReviewRejectsUnsafeRefs(t *testing.T) {
	review := BuildAutoDelegationPolicyReview(AutoDelegationPolicy{
		Mode:      AutoDelegationManagedReadOnly,
		PolicyRef: DisplaySafeRef("http://example.invalid/policy"),
	})

	if review.Status != VerificationReviewRequired ||
		review.FailureClass != FailureEvidenceWeak ||
		review.Ready ||
		review.AutoDelegationAllowed ||
		review.NextHostAction != "provide_display_safe_refs" {
		t.Fatalf("unsafe policy ref should force display-safe review: %+v", review)
	}
	if !autoDelegationMissingInputContains(review.MissingInputs, "host:display_safe_refs") ||
		!autoDelegationBoundaryContains(review.Boundaries, "raw_output_not_allowed") {
		t.Fatalf("unsafe review did not report display-safe boundary: %+v", review)
	}
}

func autoDelegationBoundaryContains(values []Boundary, want Boundary) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func autoDelegationMissingInputContains(values []MissingInput, want MissingInput) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
