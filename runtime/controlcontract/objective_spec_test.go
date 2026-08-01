package controlcontract

import "testing"

func TestBuildObjectiveSpecFrameProjectionMapsToExistingFrame(t *testing.T) {
	spec := ObjectiveSpec{
		SpecRef:        "spec:orl1",
		ObjectiveID:    "objective:orl1",
		UserGoalDigest: "sha256:abc",
		RawGoalRef:     "goal:orl1_raw",
		GoalSummary:    "summarize public website in Chinese",
		SuccessCriteria: []ObjectiveSuccessCriterion{
			{
				CriteriaRef: "criteria:fetch",
				Text:        "fetch public page evidence",
				RequiredEvidence: []EvidenceRef{{
					Ref:      "evidence:webpage_fetch",
					Kind:     "webpage_fetch",
					Strength: EvidenceAdequate,
					Source:   "scene:public_source",
				}},
			},
			{
				CriteriaRef: "criteria:summary",
				Text:        "produce Chinese summary",
				RequiredEvidence: []EvidenceRef{{
					Ref:      "evidence:summary",
					Kind:     "summary",
					Strength: EvidenceAdequate,
					Source:   "scene:public_source",
				}},
			},
		},
		Constraints: []ObjectiveConstraint{{
			ConstraintRef: "constraint:readonly",
			Kind:          "output_language",
			Text:          "Chinese",
		}},
		RequiredEvidence: []EvidenceRef{{
			Ref:      "evidence:source_ref",
			Kind:     "source",
			Strength: EvidenceAdequate,
			Source:   "scene:public_source",
		}},
		CandidateCapabilities: []DisplaySafeRef{"capability:public_source_fetch"},
		SourceContext:         []DisplaySafeRef{"catalog:objective_capabilities"},
		SideEffectPolicy:      ObjectiveSpecSideEffectReadOnly,
		MissingInfoPolicy:     ObjectiveSpecMissingInfoAskUser,
		AcceptablePartial:     []string{"source has title only"},
		Budget: ObjectiveSpecBudget{
			BudgetRef:   "budget:orl1",
			MaxNodes:    4,
			MaxAttempts: 3,
			PolicyRefs:  []DisplaySafeRef{"policy:orl1_budget"},
		},
		PolicyRefs: []DisplaySafeRef{"policy:display_safe"},
	}

	got := BuildObjectiveSpecFrameProjection(ObjectiveSpecFrameProjectionInput{
		Spec:          spec,
		ProjectionRef: "projection:orl1",
		SourceRef:     "host:objective_spec_builder",
	})
	if got.Status != VerificationSatisfied || !got.FrameMapped || got.FailureClass != FailureNone || got.NextHostAction != "run_objective_graph_planner" {
		t.Fatalf("unexpected projection = %#v", got)
	}
	if got.Frame.ID != "objective:orl1" ||
		got.Frame.UserGoalDigest != "sha256:abc" ||
		got.Frame.ControlMode != ControlModeObjective ||
		got.Frame.Intensity != IntensityL3ManagedObjective {
		t.Fatalf("frame identity/mode = %#v", got.Frame)
	}
	if len(got.Frame.SuccessCriteria) != 2 ||
		got.Frame.SuccessCriteria[0] != "fetch public page evidence" ||
		got.Frame.SuccessCriteria[1] != "produce Chinese summary" {
		t.Fatalf("success criteria = %#v", got.Frame.SuccessCriteria)
	}
	if len(got.Frame.RequiredEvidence) != 3 {
		t.Fatalf("required evidence = %#v", got.Frame.RequiredEvidence)
	}
	if len(got.Frame.CandidateCapabilities) != 1 || got.Frame.CandidateCapabilities[0] != "capability:public_source_fetch" {
		t.Fatalf("candidate capabilities = %#v", got.Frame.CandidateCapabilities)
	}
	if !objectiveSpecTestStringContains(got.Frame.Constraints, "goal_summary:summarize public website in Chinese") ||
		!objectiveSpecTestStringContains(got.Frame.Constraints, "side_effect_policy:read_only") ||
		!objectiveSpecTestStringContains(got.Frame.Constraints, "missing_info_policy:ask_user") ||
		!objectiveSpecTestStringContains(got.Frame.Constraints, "acceptable_partial:source has title only") {
		t.Fatalf("constraints = %#v", got.Frame.Constraints)
	}
	if !objectiveSpecTestBoundaryContains(got.Frame.Boundaries, "objective_spec_to_objective_frame") ||
		!objectiveSpecTestBoundaryContains(got.Boundaries, "no_second_runner") ||
		!objectiveSpecTestBoundaryContains(got.Boundaries, "no_parallel_catalog") {
		t.Fatalf("boundaries projection=%#v frame=%#v", got.Boundaries, got.Frame.Boundaries)
	}
	if !objectiveSpecTestStringContains(got.MappingWarnings, "budget_stays_on_objective_spec") ||
		!objectiveSpecTestStringContains(got.MappingWarnings, "policy_refs_stay_on_objective_spec") {
		t.Fatalf("mapping warnings = %#v", got.MappingWarnings)
	}

	clone := got.Clone()
	clone.Spec.SuccessCriteria[0].Text = "changed"
	clone.Frame.SuccessCriteria[0] = "changed"
	clone.Frame.RequiredEvidence[0].Ref = "evidence:changed"
	if got.Spec.SuccessCriteria[0].Text != "fetch public page evidence" ||
		got.Frame.SuccessCriteria[0] != "fetch public page evidence" ||
		got.Frame.RequiredEvidence[0].Ref != "evidence:source_ref" {
		t.Fatalf("clone mutated original = %#v", got)
	}
}

func TestBuildObjectiveSpecFrameProjectionBlocksMissingRequiredEvidence(t *testing.T) {
	got := BuildObjectiveSpecFrameProjection(ObjectiveSpecFrameProjectionInput{
		Spec: ObjectiveSpec{
			SpecRef:          "spec:missing_evidence",
			RawGoalRef:       "goal:missing_evidence",
			GoalSummary:      "inspect local runtime",
			SideEffectPolicy: ObjectiveSpecSideEffectReadOnly,
			SuccessCriteria: []ObjectiveSuccessCriterion{{
				Text: "inspect runtime",
			}},
		},
	})
	if got.Status != VerificationBlocked ||
		got.FrameMapped ||
		got.FailureClass != FailureEvidenceMissing ||
		got.NextHostAction != "provide_required_evidence_contract" ||
		!objectiveSpecTestMissingContains(got.MissingInputs, "host:required_evidence_contract") {
		t.Fatalf("unexpected missing evidence projection = %#v", got)
	}
}

func TestBuildObjectiveSpecFrameProjectionRejectsUnsafeRawText(t *testing.T) {
	got := BuildObjectiveSpecFrameProjection(ObjectiveSpecFrameProjectionInput{
		Spec: ObjectiveSpec{
			SpecRef:          "spec:unsafe",
			RawGoalRef:       "goal:unsafe",
			GoalSummary:      "summarize https://example.com",
			SideEffectPolicy: ObjectiveSpecSideEffectReadOnly,
			SuccessCriteria: []ObjectiveSuccessCriterion{{
				Text: "produce summary",
				RequiredEvidence: []EvidenceRef{{
					Ref:      "evidence:summary",
					Kind:     "summary",
					Strength: EvidenceAdequate,
				}},
			}},
		},
	})
	if got.Status != VerificationReviewRequired ||
		got.FrameMapped ||
		got.FailureClass != FailureEvidenceWeak ||
		got.NextHostAction != "provide_display_safe_refs" ||
		!objectiveSpecTestMissingContains(got.MissingInputs, "host:display_safe_refs") ||
		!objectiveSpecTestBoundaryContains(got.Boundaries, "raw_output_not_allowed") {
		t.Fatalf("unexpected unsafe projection = %#v", got)
	}
}

func TestBuildObjectiveSpecFrameProjectionRequiresApprovalRefWhenPolicyRequiresApproval(t *testing.T) {
	got := BuildObjectiveSpecFrameProjection(ObjectiveSpecFrameProjectionInput{
		Spec: ObjectiveSpec{
			SpecRef:          "spec:approval",
			RawGoalRef:       "goal:approval",
			GoalSummary:      "prepare side effect proposal",
			SideEffectPolicy: ObjectiveSpecSideEffectRequiresApproval,
			SuccessCriteria: []ObjectiveSuccessCriterion{{
				Text: "prepare proposal",
				RequiredEvidence: []EvidenceRef{{
					Ref:      "evidence:proposal",
					Kind:     "proposal",
					Strength: EvidenceAdequate,
				}},
			}},
		},
	})
	if got.Status != VerificationBlocked ||
		got.FailureClass != FailureApprovalRequired ||
		got.NextHostAction != "request_host_approval" ||
		!objectiveSpecTestMissingContains(got.MissingInputs, "host:objective_approval_ref") {
		t.Fatalf("unexpected approval projection = %#v", got)
	}
}

func TestBuildObjectiveSpecFrameProjectionBlocksMissingSideEffectPolicy(t *testing.T) {
	got := BuildObjectiveSpecFrameProjection(ObjectiveSpecFrameProjectionInput{
		Spec: ObjectiveSpec{
			SpecRef:     "spec:side_effect",
			RawGoalRef:  "goal:side_effect",
			GoalSummary: "answer goal",
			SuccessCriteria: []ObjectiveSuccessCriterion{{
				Text: "answer goal",
				RequiredEvidence: []EvidenceRef{{
					Ref:      "evidence:answer",
					Kind:     "answer",
					Strength: EvidenceAdequate,
				}},
			}},
		},
	})
	if got.Status != VerificationBlocked ||
		got.FailureClass != FailurePolicyBlocked ||
		got.NextHostAction != "provide_objective_contract" ||
		!objectiveSpecTestMissingContains(got.MissingInputs, "host:objective_side_effect_policy") {
		t.Fatalf("unexpected side effect projection = %#v", got)
	}
}

func objectiveSpecTestStringContains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func objectiveSpecTestBoundaryContains(values []Boundary, want Boundary) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func objectiveSpecTestMissingContains(values []MissingInput, want MissingInput) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
