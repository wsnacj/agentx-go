package controlcontract

import "testing"

func TestAutoDelegationPlanReviewReadyManagedReadOnly(t *testing.T) {
	policyReview := BuildAutoDelegationPolicyReview(AutoDelegationPolicy{
		Mode:        AutoDelegationManagedReadOnly,
		MaxChildren: 2,
	})
	review := BuildAutoDelegationPlanReview(policyReview, validAutoDelegationPlan())

	if review.Status != VerificationSatisfied ||
		!review.Ready ||
		!review.HostMayDispatch ||
		review.PlanOnly ||
		review.FailureClass != FailureNone ||
		review.NextHostAction != "host_may_dispatch_auto_delegation_children" {
		t.Fatalf("unexpected managed-readonly plan review: %+v", review)
	}
	if len(review.AcceptedChildRefs) != 1 || review.AcceptedChildRefs[0] != "child:collect_public_sources" {
		t.Fatalf("accepted child refs not recorded: %+v", review.AcceptedChildRefs)
	}
	if !review.WorkerAsToolDefault ||
		!autoDelegationBoundaryContains(review.Boundaries, "delegation_plan_contract_only") ||
		!autoDelegationBoundaryContains(review.Boundaries, "no_subagent_dispatch") ||
		!autoDelegationBoundaryContains(review.Boundaries, "auto_delegation_plan_ready_for_host_dispatch") {
		t.Fatalf("plan review did not preserve expected boundaries: %+v", review.Boundaries)
	}
}

func TestAutoDelegationPlanReviewProposeIsPlanOnly(t *testing.T) {
	policyReview := BuildAutoDelegationPolicyReview(AutoDelegationPolicy{
		Mode: AutoDelegationPropose,
	})
	review := BuildAutoDelegationPlanReview(policyReview, validAutoDelegationPlan())

	if review.Status != VerificationSatisfied ||
		!review.Ready ||
		review.HostMayDispatch ||
		!review.PlanOnly ||
		review.NextHostAction != "review_auto_delegation_proposal" {
		t.Fatalf("unexpected proposal plan review: %+v", review)
	}
	if !autoDelegationBoundaryContains(review.Boundaries, "auto_delegation_plan_ready_for_review") {
		t.Fatalf("proposal plan should be review-only: %+v", review.Boundaries)
	}
}

func TestAutoDelegationPlanReviewBlocksHandoffOwnership(t *testing.T) {
	policyReview := BuildAutoDelegationPolicyReview(AutoDelegationPolicy{
		Mode: AutoDelegationManagedReadOnly,
	})
	plan := validAutoDelegationPlan()
	plan.Children[0].ConversationOwnership = AutoDelegationConversationHandoff

	review := BuildAutoDelegationPlanReview(policyReview, plan)

	if review.Status != VerificationBlocked ||
		review.Ready ||
		review.HostMayDispatch ||
		review.FailureClass != FailureUnsupportedOperation ||
		!autoDelegationMissingInputContains(review.MissingInputs, "host:auto_delegation_handoff_policy") {
		t.Fatalf("handoff should be blocked by default: %+v", review)
	}
	if !autoDelegationBoundaryContains(review.Boundaries, "handoff_not_enabled") {
		t.Fatalf("handoff boundary missing: %+v", review.Boundaries)
	}
}

func TestAutoDelegationPlanReviewRejectsOverBudgetChildren(t *testing.T) {
	policyReview := BuildAutoDelegationPolicyReview(AutoDelegationPolicy{
		Mode:        AutoDelegationManagedReadOnly,
		MaxChildren: 1,
	})
	plan := validAutoDelegationPlan()
	second := plan.Children[0]
	second.ChildRef = "child:cross_check_sources"
	second.Goal = "Cross-check the first child result against an independent source."
	plan.Children = append(plan.Children, second)

	review := BuildAutoDelegationPlanReview(policyReview, plan)

	if review.Status != VerificationBlocked ||
		review.FailureClass != FailurePolicyBlocked ||
		review.HostMayDispatch ||
		!autoDelegationMissingInputContains(review.MissingInputs, "host:auto_delegation_child_budget") ||
		review.NextHostAction != "reduce_auto_delegation_children" {
		t.Fatalf("over-budget children should be blocked: %+v", review)
	}
}

func TestAutoDelegationPlanReviewRequiresSelfContainedChildInputs(t *testing.T) {
	policyReview := BuildAutoDelegationPolicyReview(AutoDelegationPolicy{
		Mode: AutoDelegationManagedReadOnly,
	})
	plan := validAutoDelegationPlan()
	plan.Children[0].Goal = ""
	plan.Children[0].CapabilityRefs = nil
	plan.Children[0].AllowedToolRefs = nil
	plan.Children[0].ExpectedEvidence = nil
	plan.Children[0].ExpectedOutput = ""

	review := BuildAutoDelegationPlanReview(policyReview, plan)

	if review.Status != VerificationBlocked ||
		review.Ready ||
		review.HostMayDispatch ||
		review.FailureClass == FailureNone {
		t.Fatalf("malformed child should block plan: %+v", review)
	}
	for _, want := range []MissingInput{
		"host:auto_delegation_child_goal",
		"host:auto_delegation_child_capability_refs",
		"host:auto_delegation_child_expected_evidence",
		"host:auto_delegation_child_expected_output",
	} {
		if !autoDelegationMissingInputContains(review.MissingInputs, want) {
			t.Fatalf("missing expected input %q in %+v", want, review.MissingInputs)
		}
	}
}

func TestAutoDelegationPlanReviewRejectsDependencyCycle(t *testing.T) {
	policyReview := BuildAutoDelegationPolicyReview(AutoDelegationPolicy{
		Mode:        AutoDelegationManagedReadOnly,
		MaxChildren: 2,
	})
	plan := validAutoDelegationPlan()
	second := plan.Children[0]
	second.ChildRef = "child:cross_check_sources"
	second.Goal = "Cross-check the first child result against an independent source."
	second.Dependencies = []DisplaySafeRef{"child:collect_public_sources"}
	plan.Children[0].Dependencies = []DisplaySafeRef{"child:cross_check_sources"}
	plan.Children = append(plan.Children, second)

	review := BuildAutoDelegationPlanReview(policyReview, plan)

	if review.Status != VerificationBlocked ||
		review.FailureClass != FailureInvalidInput ||
		!autoDelegationMissingInputContains(review.MissingInputs, "host:auto_delegation_dependency_order") ||
		!autoDelegationBoundaryContains(review.Boundaries, "auto_delegation_dependency_cycle") {
		t.Fatalf("dependency cycle should be blocked: %+v", review)
	}
}

func TestAutoDelegationPlanReviewBlocksRecursiveOrchestratorWithoutDepthPolicy(t *testing.T) {
	policyReview := BuildAutoDelegationPolicyReview(AutoDelegationPolicy{
		Mode: AutoDelegationManagedReadOnly,
	})
	plan := validAutoDelegationPlan()
	plan.MaxDepth = 1
	plan.Children[0].Role = AutoDelegationChildRoleOrchestrator
	plan.Children[0].Depth = 0

	review := BuildAutoDelegationPlanReview(policyReview, plan)

	if review.Status != VerificationBlocked ||
		review.FailureClass != FailurePolicyBlocked ||
		!autoDelegationMissingInputContains(review.MissingInputs, "host:auto_delegation_depth_policy") ||
		!autoDelegationBoundaryContains(review.Boundaries, "auto_delegation_depth_policy_blocked") {
		t.Fatalf("orchestrator child should require explicit depth policy: %+v", review)
	}
}

func TestAutoDelegationPlanReviewAcceptsOrchestratorWithDepthPolicy(t *testing.T) {
	policyReview := BuildAutoDelegationPolicyReview(AutoDelegationPolicy{
		Mode: AutoDelegationManagedReadOnly,
	})
	plan := validAutoDelegationPlan()
	plan.MaxDepth = 2
	plan.Children[0].Role = AutoDelegationChildRoleOrchestrator
	plan.Children[0].Depth = 0

	review := BuildAutoDelegationPlanReview(policyReview, plan)

	if review.Status != VerificationSatisfied || !review.Ready || !review.HostMayDispatch {
		t.Fatalf("orchestrator child with explicit depth should be accepted: %+v", review)
	}
}

func TestAutoDelegationPlanReviewRejectsUnsafeRefs(t *testing.T) {
	policyReview := BuildAutoDelegationPolicyReview(AutoDelegationPolicy{
		Mode: AutoDelegationManagedReadOnly,
	})
	plan := validAutoDelegationPlan()
	plan.PlanRef = DisplaySafeRef("https://example.invalid/plan")

	review := BuildAutoDelegationPlanReview(policyReview, plan)

	if review.Status != VerificationReviewRequired ||
		review.FailureClass != FailureEvidenceWeak ||
		review.Ready ||
		review.HostMayDispatch ||
		review.NextHostAction != "provide_display_safe_refs" {
		t.Fatalf("unsafe plan ref should force display-safe review: %+v", review)
	}
	if !autoDelegationMissingInputContains(review.MissingInputs, "host:display_safe_refs") ||
		!autoDelegationBoundaryContains(review.Boundaries, "raw_output_not_allowed") {
		t.Fatalf("unsafe review did not report display-safe boundary: %+v", review)
	}
}

func validAutoDelegationPlan() AutoDelegationPlan {
	return AutoDelegationPlan{
		PlanRef:            "plan:public_source_summary",
		ParentObjectiveRef: "objective:root",
		Children: []AutoDelegationChildTask{
			{
				ChildRef:           "child:collect_public_sources",
				ParentObjectiveRef: "objective:root",
				Goal:               "Collect public-source evidence needed to answer the parent objective.",
				ContextRefs:        []DisplaySafeRef{"context:user_goal"},
				RelevantFindings:   []string{"The parent objective needs independent source evidence."},
				Constraints:        []string{"Use read-only tools only."},
				CapabilityRefs:     []DisplaySafeRef{"capability:public_source"},
				ExpectedEvidence: []EvidenceRef{
					{
						Ref:      "evidence:public_source_summary",
						Kind:     "summary",
						Strength: EvidenceAdequate,
						Source:   "capability:public_source",
					},
				},
				ExpectedOutput:   "A bounded summary with display-safe evidence references.",
				Role:             AutoDelegationChildRoleLeaf,
				SideEffectPolicy: ObjectiveSpecSideEffectReadOnly,
			},
		},
		RequiredEvidence: []EvidenceRef{
			{
				Ref:      "evidence:public_source_summary",
				Kind:     "summary",
				Strength: EvidenceAdequate,
				Source:   "capability:public_source",
			},
		},
	}
}
