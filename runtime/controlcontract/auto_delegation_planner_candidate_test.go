package controlcontract

import "testing"

func TestAutoDelegationPlannerInstructionProfileOffKeepsToolsButNoProactiveInstruction(t *testing.T) {
	profile := BuildAutoDelegationPlannerInstructionProfile(AutoDelegationPlannerInstructionInput{
		PolicyReview: BuildAutoDelegationPolicyReview(AutoDelegationPolicy{}),
		AvailableToolRefs: []DisplaySafeRef{
			"tool:web_search",
			"tool:file_read",
		},
	})

	if profile.Status != VerificationNotApplicable ||
		profile.RequestPlanCandidate ||
		profile.ProactiveDelegationInstructionsExposed ||
		profile.BoundedExecutionInstructionsExposed ||
		len(profile.AvailableToolRefs) != 2 ||
		profile.PromptEffect != "auto_delegation_instruction_absent" {
		t.Fatalf("off mode should keep tools but hide delegation instructions: %+v", profile)
	}
	if !autoDelegationBoundaryContains(profile.Boundaries, "tools_may_be_available_without_delegation_instruction") ||
		!autoDelegationBoundaryContains(profile.Boundaries, "no_subagent_dispatch") {
		t.Fatalf("off mode boundaries missing: %+v", profile.Boundaries)
	}
}

func TestAutoDelegationPlannerInstructionProfileObserveDoesNotExposeSpawnInstruction(t *testing.T) {
	profile := BuildAutoDelegationPlannerInstructionProfile(AutoDelegationPlannerInstructionInput{
		PolicyReview: BuildAutoDelegationPolicyReview(AutoDelegationPolicy{
			Mode: AutoDelegationObserve,
		}),
		AvailableToolRefs: []DisplaySafeRef{"tool:web_search"},
	})

	if profile.Status != VerificationSatisfied ||
		profile.Mode != AutoDelegationPlannerInstructionObserve ||
		profile.RequestPlanCandidate ||
		profile.ProactiveDelegationInstructionsExposed ||
		profile.BoundedExecutionInstructionsExposed ||
		profile.NextHostAction != "record_auto_delegation_observation" {
		t.Fatalf("observe mode should not expose spawn instructions: %+v", profile)
	}
}

func TestAutoDelegationPlannerInstructionProfileProposeRequestsReviewablePlanOnly(t *testing.T) {
	profile := BuildAutoDelegationPlannerInstructionProfile(AutoDelegationPlannerInstructionInput{
		PolicyReview: BuildAutoDelegationPolicyReview(AutoDelegationPolicy{
			Mode: AutoDelegationPropose,
		}),
	})

	if profile.Status != VerificationSatisfied ||
		profile.Mode != AutoDelegationPlannerInstructionPropose ||
		!profile.RequestPlanCandidate ||
		!profile.ProactiveDelegationInstructionsExposed ||
		!profile.PlanOnly ||
		profile.BoundedExecutionInstructionsExposed ||
		profile.NextHostAction != "request_auto_delegation_proposal_json" {
		t.Fatalf("proposal mode should request review-only JSON: %+v", profile)
	}
	if !autoDelegationBoundaryContains(profile.Boundaries, "auto_delegation_proposal_instruction_only") {
		t.Fatalf("proposal boundary missing: %+v", profile.Boundaries)
	}
}

func TestAutoDelegationPlannerInstructionProfileManagedRequestsBoundedPlan(t *testing.T) {
	profile := BuildAutoDelegationPlannerInstructionProfile(AutoDelegationPlannerInstructionInput{
		PolicyReview: BuildAutoDelegationPolicyReview(AutoDelegationPolicy{
			Mode: AutoDelegationManagedReadOnly,
		}),
	})

	if profile.Status != VerificationSatisfied ||
		profile.Mode != AutoDelegationPlannerInstructionManaged ||
		!profile.RequestPlanCandidate ||
		!profile.ProactiveDelegationInstructionsExposed ||
		!profile.BoundedExecutionInstructionsExposed ||
		profile.PlanOnly ||
		profile.NextHostAction != "request_bounded_auto_delegation_plan_json" {
		t.Fatalf("managed mode should request bounded plan JSON: %+v", profile)
	}
}

func TestAutoDelegationPlannerCandidateDecodeFencedPlanAndReview(t *testing.T) {
	report := BuildAutoDelegationPlannerCandidateFromJSON(AutoDelegationPlannerCandidateJSONDecodeInput{
		RawJSON: "```json\n" + validAutoDelegationPlannerPlanJSON() + "\n```",
		PolicyReview: BuildAutoDelegationPolicyReview(AutoDelegationPolicy{
			Mode:        AutoDelegationManagedReadOnly,
			MaxChildren: 2,
		}),
	})

	if report.Status != VerificationSatisfied ||
		!report.Decoded ||
		report.FailureClass != FailureNone ||
		!report.PlanReview.HostMayDispatch ||
		report.NextHostAction != "host_may_dispatch_auto_delegation_children" {
		t.Fatalf("valid fenced planner JSON should decode and review: %+v", report)
	}
	if !autoDelegationBoundaryContains(report.Boundaries, "code_fence_json_cleanup_allowed") ||
		!autoDelegationBoundaryContains(report.Boundaries, "auto_delegation_planner_candidate_validated") {
		t.Fatalf("decode boundaries missing: %+v", report.Boundaries)
	}
}

func TestAutoDelegationPlannerCandidateDecodeWrapper(t *testing.T) {
	report := BuildAutoDelegationPlannerCandidateFromJSON(AutoDelegationPlannerCandidateJSONDecodeInput{
		RawJSON: `{
			"candidate_ref":"candidate:auto_delegation_public_source",
			"source_ref":"planner:llm",
			"plan":` + validAutoDelegationPlannerPlanJSON() + `
		}`,
		PolicyReview: BuildAutoDelegationPolicyReview(AutoDelegationPolicy{
			Mode: AutoDelegationPropose,
		}),
	})

	if report.Status != VerificationSatisfied ||
		!report.Decoded ||
		report.Candidate.CandidateRef != "candidate:auto_delegation_public_source" ||
		!report.PlanReview.PlanOnly ||
		report.PlanReview.HostMayDispatch {
		t.Fatalf("wrapped planner candidate should decode as proposal-only: %+v", report)
	}
}

func TestAutoDelegationPlannerCandidateRejectsFreeFormOutput(t *testing.T) {
	report := BuildAutoDelegationPlannerCandidateFromJSON(AutoDelegationPlannerCandidateJSONDecodeInput{
		RawJSON: "I would split this into two child tasks.",
		PolicyReview: BuildAutoDelegationPolicyReview(AutoDelegationPolicy{
			Mode: AutoDelegationManagedReadOnly,
		}),
	})

	if report.Status != VerificationBlocked ||
		report.Decoded ||
		report.FailureClass != FailureInvalidInput ||
		!autoDelegationMissingInputContains(report.MissingInputs, "host:auto_delegation_planner_json") ||
		!autoDelegationBoundaryContains(report.Boundaries, "deterministic_blocked_fallback") {
		t.Fatalf("free-form output should be deterministically blocked: %+v", report)
	}
}

func TestAutoDelegationPlannerCandidateRejectsUnknownFields(t *testing.T) {
	report := BuildAutoDelegationPlannerCandidateFromJSON(AutoDelegationPlannerCandidateJSONDecodeInput{
		RawJSON: `{
			"plan_ref":"plan:unknown_field",
			"parent_objective_ref":"objective:root",
			"children":[],
			"unstructured_notes":"please trust me"
		}`,
		PolicyReview: BuildAutoDelegationPolicyReview(AutoDelegationPolicy{
			Mode: AutoDelegationManagedReadOnly,
		}),
	})

	if report.Status != VerificationBlocked ||
		report.Decoded ||
		report.FailureClass != FailureInvalidInput ||
		!autoDelegationBoundaryContains(report.Boundaries, "auto_delegation_planner_json_invalid") {
		t.Fatalf("unknown fields should fail strict JSON decode: %+v", report)
	}
}

func TestAutoDelegationPlannerCandidatePartialOutputRunsPlanReview(t *testing.T) {
	report := BuildAutoDelegationPlannerCandidateFromJSON(AutoDelegationPlannerCandidateJSONDecodeInput{
		RawJSON: `{
			"plan_ref":"plan:partial",
			"parent_objective_ref":"objective:root",
			"children":[]
		}`,
		PolicyReview: BuildAutoDelegationPolicyReview(AutoDelegationPolicy{
			Mode: AutoDelegationManagedReadOnly,
		}),
	})

	if report.Status != VerificationBlocked ||
		!report.Decoded ||
		report.FailureClass != FailureInsufficientInformation ||
		!autoDelegationMissingInputContains(report.MissingInputs, "host:auto_delegation_children") {
		t.Fatalf("partial planner output should decode but fail plan review: %+v", report)
	}
}

func TestAutoDelegationPlannerCandidateUnsafePlanNeedsDisplaySafeRefs(t *testing.T) {
	report := BuildAutoDelegationPlannerCandidateFromJSON(AutoDelegationPlannerCandidateJSONDecodeInput{
		RawJSON: `{
			"plan_ref":"https://example.invalid/plan",
			"parent_objective_ref":"objective:root",
			"children":[]
		}`,
		PolicyReview: BuildAutoDelegationPolicyReview(AutoDelegationPolicy{
			Mode: AutoDelegationManagedReadOnly,
		}),
	})

	if report.Status != VerificationReviewRequired ||
		!report.Decoded ||
		report.FailureClass != FailureEvidenceWeak ||
		!autoDelegationMissingInputContains(report.MissingInputs, "host:display_safe_refs") ||
		!autoDelegationBoundaryContains(report.Boundaries, "raw_output_not_allowed") {
		t.Fatalf("unsafe planner output should require display-safe refs: %+v", report)
	}
}

func validAutoDelegationPlannerPlanJSON() string {
	return `{
		"plan_ref":"plan:public_source_summary",
		"parent_objective_ref":"objective:root",
		"children":[{
			"child_ref":"child:collect_public_sources",
			"parent_objective_ref":"objective:root",
			"goal":"Collect public-source evidence needed to answer the parent objective.",
			"context_refs":["context:user_goal"],
			"relevant_findings":["The parent objective needs independent source evidence."],
			"constraints":["Use read-only tools only."],
			"capability_refs":["capability:public_source"],
			"expected_evidence":[{
				"ref":"evidence:public_source_summary",
				"kind":"summary",
				"strength":"adequate",
				"source":"capability:public_source"
			}],
			"expected_output":"A bounded summary with display-safe evidence references.",
			"role":"leaf",
			"side_effect_policy":"read_only"
		}],
		"required_evidence":[{
			"ref":"evidence:public_source_summary",
			"kind":"summary",
			"strength":"adequate",
			"source":"capability:public_source"
		}]
	}`
}
