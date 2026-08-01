package controlcontract

import "testing"

func TestDelegationRequestBlocksL5ByDefault(t *testing.T) {
	request := BuildDelegationRequestProjection(delegationReadyRequestInput(IntensityL5Autonomous))
	if request.Status != VerificationBlocked ||
		request.ReadyForWorkerDispatch ||
		request.DelegationAllowed ||
		request.FailureClass != FailureApprovalRequired ||
		!delegationMissingInputContains(request.MissingInputs, "host:l5_delegation_policy") ||
		!delegationBoundaryContains(request.Boundaries, "l5_delegation_default_off") ||
		request.WorkerOutputAcceptedAsFact {
		t.Fatalf("L5 default-off request = %#v", request)
	}
	if request.RunnerEffect != "none" || request.PromptEffect != "none" {
		t.Fatalf("delegation request should be projection-only: %#v", request)
	}
}

func TestDelegationRequestBlocksL4WithoutHostPolicy(t *testing.T) {
	input := delegationReadyRequestInput(IntensityL4DurableLongRun)
	input.HostAllowsL4Delegation = false
	request := BuildDelegationRequestProjection(input)
	if request.Status != VerificationBlocked ||
		request.ReadyForWorkerDispatch ||
		request.DelegationAllowed ||
		request.FailureClass != FailureApprovalRequired ||
		!delegationMissingInputContains(request.MissingInputs, "host:l4_delegation_policy") ||
		!delegationBoundaryContains(request.Boundaries, "l4_delegation_not_explicitly_allowed") {
		t.Fatalf("L4 without host policy = %#v", request)
	}
}

func TestDelegationRequestAllowsExplicitL4LimitedWorkerBoundary(t *testing.T) {
	request := BuildDelegationRequestProjection(delegationReadyRequestInput(IntensityL4DurableLongRun))
	if request.Status != VerificationSatisfied ||
		!request.ReadyForWorkerDispatch ||
		!request.DelegationAllowed ||
		request.RequestedIntensity != IntensityL4DurableLongRun ||
		!request.HostAllowsL4Delegation ||
		!request.HostApproved ||
		!request.UserConfirmed ||
		request.WorkerOutputAcceptedAsFact ||
		!request.WorkerResultRequiresVerification {
		t.Fatalf("explicit L4 delegation request = %#v", request)
	}
	if len(request.AllowedToolRefs) != 2 ||
		!delegationDisplayRefContains(request.AllowedToolRefs, "tool:read") ||
		!delegationDisplayRefContains(request.AllowedToolRefs, "tool:search") {
		t.Fatalf("expected worker tool boundary, got %#v", request.AllowedToolRefs)
	}
	for _, boundary := range []Boundary{
		"delegation_request_projection",
		"delegation_worker_boundary",
		"no_worker_dispatch",
		"worker_output_not_fact",
		"worker_result_requires_verification",
		"delegation_request_ready",
	} {
		if !delegationBoundaryContains(request.Boundaries, boundary) {
			t.Fatalf("missing boundary %q in %#v", boundary, request.Boundaries)
		}
	}
}

func TestDelegationRequestRequiresWorkerBoundaryBudgetStopAndEvidence(t *testing.T) {
	input := delegationReadyRequestInput(IntensityL4DurableLongRun)
	input.AllowedToolRefs = nil
	input.Budget = ObjectiveBudgetSnapshot{}
	input.EvidenceRequirements = nil
	input.StopConditionRefs = nil
	input.RedactionPolicyRef = ""
	input.MergePolicyRef = ""
	request := BuildDelegationRequestProjection(input)
	if request.Status != VerificationBlocked || request.ReadyForWorkerDispatch {
		t.Fatalf("expected blocked incomplete delegation request, got %#v", request)
	}
	for _, missing := range []MissingInput{
		"host:worker_allowed_tools",
		"host:delegation_budget",
		"host:delegation_evidence_requirements",
		"host:delegation_stop_condition",
		"host:redaction_policy",
		"host:merge_policy",
	} {
		if !delegationMissingInputContains(request.MissingInputs, missing) {
			t.Fatalf("missing %q in %#v", missing, request.MissingInputs)
		}
	}
}

func TestDelegationWorkerResultReviewRequiresParentVerification(t *testing.T) {
	request := BuildDelegationRequestProjection(delegationReadyRequestInput(IntensityL4DurableLongRun))
	review := BuildDelegationWorkerResultReview(DelegationWorkerResultReviewInput{
		Request:         request,
		WorkerRunRef:    "worker_run:delegation_fixture",
		WorkerResultRef: "worker_result:delegation_fixture",
	})
	if review.Status != VerificationBlocked ||
		review.ReadyForParentMerge ||
		review.WorkerOutputAcceptedAsFact ||
		!review.WorkerResultRequiresVerification ||
		!delegationMissingInputContains(review.MissingInputs, "host:parent_verification") ||
		!delegationStringContains(review.BlockedReasons, "parent_verification_missing") {
		t.Fatalf("unverified worker result review = %#v", review)
	}

	verified := BuildDelegationWorkerResultReview(DelegationWorkerResultReviewInput{
		Request:         request,
		WorkerRunRef:    "worker_run:delegation_fixture",
		WorkerResultRef: "worker_result:delegation_fixture",
		Verification: ObjectiveVerificationGateResult{
			Status:    VerificationSatisfied,
			Satisfied: true,
			EvidenceRefs: []EvidenceRef{{
				Ref:      "evidence:delegation_worker_verified",
				Kind:     "delegation_worker_result",
				Strength: EvidenceAdequate,
				Source:   "worker:delegation_fixture",
			}},
			Boundaries: []Boundary{"parent_objective_verification_gate"},
		},
	})
	if verified.Status != VerificationSatisfied ||
		!verified.ReadyForParentMerge ||
		verified.WorkerOutputAcceptedAsFact ||
		!verified.WorkerResultRequiresVerification ||
		!delegationBoundaryContains(verified.Boundaries, "verified_worker_result_ready_for_parent_merge") {
		t.Fatalf("verified worker result review = %#v", verified)
	}
}

func TestDelegationWorkerParentMergeReverifiesAndProjectsLedgerPatch(t *testing.T) {
	invocation := delegationReadyWorkerInvocation()
	merge := BuildDelegationWorkerParentMerge(DelegationWorkerParentMergeInput{
		Invocation:       invocation,
		ParentLedgerRef:  "ledger:delegation_parent",
		WorkerAttemptRef: "attempt:delegation_worker_1",
		MergeRef:         "merge:delegation_worker_1",
		WorkerObservations: []Observation{{
			Kind:     "delegation_worker_result",
			Source:   "worker:delegation_fixture",
			Subject:  "objective:delegation_fixture",
			Name:     "worker_result_verified",
			Value:    "true",
			Strength: EvidenceAdequate,
			EvidenceRefs: []EvidenceRef{{
				Ref:      "evidence:delegation_worker_verified",
				Kind:     "delegation_worker_result",
				Strength: EvidenceAdequate,
				Source:   "worker:delegation_fixture",
			}},
		}},
		ExpectedObservationKinds: []string{"delegation_worker_result"},
	})
	if merge.Status != VerificationSatisfied ||
		!merge.ReadyForParentMerge ||
		merge.WorkerOutputAcceptedAsFact ||
		!merge.WorkerResultRequiresVerification ||
		!merge.ParentVerification.Satisfied ||
		!merge.ResultReview.ReadyForParentMerge ||
		merge.ParentLedgerPatch.LedgerRef != "ledger:delegation_parent" ||
		len(merge.ParentLedgerPatch.Attempts) != 1 ||
		merge.ParentLedgerPatch.Attempts[0].Ref != "attempt:delegation_worker_1" ||
		merge.ParentLedgerPatch.Attempts[0].Status != VerificationSatisfied ||
		merge.ParentLedgerPatch.Attempts[0].ObservationCount != 1 {
		t.Fatalf("delegation worker parent merge = %#v", merge)
	}
	for _, boundary := range []Boundary{
		"parent_objective_reverified_before_merge",
		"verified_worker_result_ready_for_parent_merge",
		"verified_worker_result_merged_into_parent_ledger",
		"ledger_patch_projection_only",
		"no_store_mutation_by_core",
	} {
		if !delegationBoundaryContains(merge.Boundaries, boundary) {
			t.Fatalf("missing boundary %q in %#v", boundary, merge.Boundaries)
		}
	}
}

func TestDelegationWorkerParentMergeBlocksWhenParentVerificationMissingEvidence(t *testing.T) {
	invocation := delegationReadyWorkerInvocation()
	merge := BuildDelegationWorkerParentMerge(DelegationWorkerParentMergeInput{
		Invocation:       invocation,
		ParentLedgerRef:  "ledger:delegation_parent",
		WorkerAttemptRef: "attempt:delegation_worker_1",
		MergeRef:         "merge:delegation_worker_1",
		WorkerObservations: []Observation{{
			Kind:     "other_evidence",
			Source:   "worker:delegation_fixture",
			Subject:  "objective:delegation_fixture",
			Name:     "worker_partial_result",
			Value:    "partial",
			Strength: EvidenceAdequate,
			EvidenceRefs: []EvidenceRef{{
				Ref:      "evidence:other",
				Kind:     "other_evidence",
				Strength: EvidenceAdequate,
				Source:   "worker:delegation_fixture",
			}},
		}},
	})
	if merge.ReadyForParentMerge ||
		merge.Status == VerificationSatisfied ||
		merge.ParentVerification.Satisfied ||
		!delegationStringContains(merge.BlockedReasons, "parent_verification_not_satisfied") ||
		!delegationMissingInputContains(merge.MissingInputs, "evidence:delegation_worker_verified") ||
		len(merge.ParentLedgerPatch.Attempts) != 0 {
		t.Fatalf("expected parent verification block, got %#v", merge)
	}
}

func TestDelegationWorkerParentMergeBlocksUnreadyInvocation(t *testing.T) {
	invocation := delegationReadyWorkerInvocation()
	invocation.HostInvocationCompleted = false
	invocation.ReadyForWorkerResultReview = false
	merge := BuildDelegationWorkerParentMerge(DelegationWorkerParentMergeInput{
		Invocation:       invocation,
		ParentLedgerRef:  "ledger:delegation_parent",
		WorkerAttemptRef: "attempt:delegation_worker_1",
		MergeRef:         "merge:delegation_worker_1",
		WorkerObservations: []Observation{{
			Kind:     "delegation_worker_result",
			Strength: EvidenceAdequate,
			EvidenceRefs: []EvidenceRef{{
				Ref:      "evidence:delegation_worker_verified",
				Kind:     "delegation_worker_result",
				Strength: EvidenceAdequate,
				Source:   "worker:delegation_fixture",
			}},
		}},
	})
	if merge.ReadyForParentMerge ||
		!delegationStringContains(merge.BlockedReasons, "delegation_worker_runtime_invocation_not_ready_for_parent_merge") ||
		!delegationMissingInputContains(merge.MissingInputs, "host:delegation_worker_runtime_invocation") {
		t.Fatalf("expected unready invocation block, got %#v", merge)
	}
}

func TestDelegationWorkerFailureReviewReadyForNoProgress(t *testing.T) {
	review := BuildDelegationWorkerFailureReview(DelegationWorkerFailureReviewInput{
		Request:          BuildDelegationRequestProjection(delegationReadyRequestInput(IntensityL4DurableLongRun)),
		FailureReviewRef: "failure_review:delegation_worker_no_progress",
		FailureRef:       "failure:delegation_worker_no_progress",
		CompensationRef:  "compensation:delegation_worker_no_progress",
		WorkerAttempts: []AttemptSummary{
			{
				Ref:          "attempt:delegation_worker_1",
				ObjectiveID:  "objective:delegation_fixture",
				StrategyID:   "worker:delegation_fixture_readonly",
				ControlMode:  ControlModeDelegated,
				Intensity:    IntensityL4DurableLongRun,
				Status:       VerificationPartial,
				FailureClass: FailureEvidenceMissing,
			},
			{
				Ref:          "attempt:delegation_worker_2",
				ObjectiveID:  "objective:delegation_fixture",
				StrategyID:   "worker:delegation_fixture_readonly",
				ControlMode:  ControlModeDelegated,
				Intensity:    IntensityL4DurableLongRun,
				Status:       VerificationBlocked,
				FailureClass: FailureEvidenceMissing,
			},
		},
	})
	if review.Status != "ready_for_delegation_worker_failure_review" ||
		!review.ReadyForFailureReview ||
		!review.ReadyForCompensationReview ||
		!review.NoProgressDetected ||
		review.ConflictingResultsDetected ||
		review.WorkerFailureReported ||
		review.NoProgressAttemptCount != 2 ||
		review.FailureClass != FailureRepeatedNoProgress ||
		len(review.NoProgressAttemptRefs) != 2 ||
		review.WorkerOutputAcceptedAsFact ||
		!review.WorkerResultRequiresVerification {
		t.Fatalf("no-progress failure review = %#v", review)
	}
	for _, boundary := range []Boundary{
		"delegation_worker_no_progress_detected",
		"ready_for_delegation_worker_failure_review",
		"compensation_not_executed",
		"no_store_mutation_by_core",
	} {
		if !delegationBoundaryContains(review.Boundaries, boundary) {
			t.Fatalf("missing boundary %q in %#v", boundary, review.Boundaries)
		}
	}
}

func TestDelegationWorkerFailureReviewReadyForConflictingResults(t *testing.T) {
	first := delegationWorkerParentMergeForConflict("merge:delegation_worker_a", "evidence:delegation_worker_conflict_a", "42")
	second := delegationWorkerParentMergeForConflict("merge:delegation_worker_b", "evidence:delegation_worker_conflict_b", "7")
	review := BuildDelegationWorkerFailureReview(DelegationWorkerFailureReviewInput{
		ParentMerges:     []DelegationWorkerParentMerge{first, second},
		FailureReviewRef: "failure_review:delegation_worker_conflict",
		FailureRef:       "failure:delegation_worker_conflict",
	})
	if review.Status != "ready_for_delegation_worker_failure_review" ||
		!review.ReadyForFailureReview ||
		review.NoProgressDetected ||
		!review.ConflictingResultsDetected ||
		review.FailureClass != FailureVerificationFailed ||
		len(review.ConflictRefs) != 2 ||
		!delegationBoundaryContains(review.Boundaries, "delegation_worker_conflict_detected") {
		t.Fatalf("conflict failure review = %#v", review)
	}
}

func TestDelegationWorkerFailureReviewReadyForReportedWorkerFailure(t *testing.T) {
	readiness := BuildHostOwnedDelegationWorkerRuntimeReadiness(delegationWorkerRuntimeTestReadinessInput(IntensityL4DurableLongRun))
	invocation := BuildHostOwnedDelegationWorkerRuntimeInvocation(HostOwnedDelegationWorkerRuntimeInvocationInput{
		Readiness:               readiness,
		InvocationReportRef:     "invocation_report:delegation_worker_failure",
		ObservedInvocationRef:   readiness.InvocationRef,
		HostWorkerRuntimeRunRef: "worker_runtime_run:delegation_failure",
		ObservedWorkerRunRef:    readiness.WorkerRunRef,
		FailureRef:              "failure:delegation_worker_runtime",
		CompensationRef:         "compensation:delegation_worker_runtime",
		HostInvocationReported:  true,
		HostInvocationFailed:    true,
	})
	review := BuildDelegationWorkerFailureReview(DelegationWorkerFailureReviewInput{
		Invocations:      []HostOwnedDelegationWorkerRuntimeInvocation{invocation},
		FailureReviewRef: "failure_review:delegation_worker_runtime",
		FailureRef:       "failure:delegation_worker_runtime",
		CompensationRef:  "compensation:delegation_worker_runtime",
	})
	if !review.ReadyForFailureReview ||
		!review.ReadyForCompensationReview ||
		!review.WorkerFailureReported ||
		len(review.FailedWorkerRunRefs) != 1 ||
		!delegationBoundaryContains(review.Boundaries, "delegation_worker_failure_reported") {
		t.Fatalf("worker failure review = %#v", review)
	}
}

func TestDelegationWorkerFailureReviewBlocksWithoutSignalAndUnsafeRefs(t *testing.T) {
	empty := BuildDelegationWorkerFailureReview(DelegationWorkerFailureReviewInput{
		FailureReviewRef: "failure_review:delegation_worker",
		FailureRef:       "failure:delegation_worker",
	})
	if empty.ReadyForFailureReview ||
		!delegationMissingInputContains(empty.MissingInputs, "host:delegation_worker_failure_signal") ||
		!delegationStringContains(empty.BlockedReasons, "delegation_worker_failure_signal_missing") {
		t.Fatalf("expected missing signal block, got %#v", empty)
	}

	unsafe := BuildDelegationWorkerFailureReview(DelegationWorkerFailureReviewInput{
		FailureReviewRef: "http://localhost/failure-review",
		FailureRef:       "failure:delegation_worker",
		ConflictRefs:     []DisplaySafeRef{"conflict:delegation_worker"},
	})
	if unsafe.ReadyForFailureReview ||
		!unsafe.RawOutputLoaded ||
		!delegationMissingInputContains(unsafe.MissingInputs, "host:display_safe_refs") ||
		!delegationBoundaryContains(unsafe.Boundaries, "raw_output_not_allowed") {
		t.Fatalf("expected unsafe input block, got %#v", unsafe)
	}
}

func delegationReadyWorkerInvocation() HostOwnedDelegationWorkerRuntimeInvocation {
	readiness := BuildHostOwnedDelegationWorkerRuntimeReadiness(delegationWorkerRuntimeTestReadinessInput(IntensityL4DurableLongRun))
	return BuildHostOwnedDelegationWorkerRuntimeInvocation(HostOwnedDelegationWorkerRuntimeInvocationInput{
		Readiness:               readiness,
		InvocationReportRef:     "invocation_report:delegation_worker_runtime",
		ObservedInvocationRef:   readiness.InvocationRef,
		HostWorkerRuntimeRunRef: "worker_runtime_run:delegation_fixture",
		ObservedWorkerRunRef:    readiness.WorkerRunRef,
		WorkerResultRef:         "worker_result:delegation_fixture",
		WorkerReadbackRef:       "worker_readback:delegation_fixture",
		ObservationRef:          "observation:delegation_worker_result",
		HostInvocationReported:  true,
		HostInvocationCompleted: true,
		EvidenceRefs: []EvidenceRef{{
			Ref:      "evidence:delegation_worker_verified",
			Kind:     "delegation_worker_result",
			Strength: EvidenceAdequate,
			Source:   "worker:delegation_fixture",
		}},
	})
}

func delegationWorkerParentMergeForConflict(mergeRef DisplaySafeRef, evidenceRef DisplaySafeRef, value string) DelegationWorkerParentMerge {
	return BuildDelegationWorkerParentMerge(DelegationWorkerParentMergeInput{
		Invocation:       delegationReadyWorkerInvocation(),
		ParentLedgerRef:  DisplaySafeRef("ledger:" + string(mergeRef)),
		WorkerAttemptRef: AttemptRef("attempt:" + string(mergeRef)),
		MergeRef:         mergeRef,
		WorkerObservations: []Observation{{
			Kind:     "delegation_worker_result",
			Source:   "worker:delegation_fixture",
			Subject:  "objective:delegation_fixture",
			Name:     "worker_answer",
			Value:    value,
			Strength: EvidenceAdequate,
			DisplaySafeRefs: []DisplaySafeRef{
				DisplaySafeRef("observation:" + string(mergeRef)),
			},
			EvidenceRefs: []EvidenceRef{{
				Ref:      evidenceRef,
				Kind:     "delegation_worker_result",
				Strength: EvidenceAdequate,
				Source:   "worker:delegation_fixture",
			}},
		}},
		RequiredEvidence: []EvidenceRef{{
			Ref:      evidenceRef,
			Kind:     "delegation_worker_result",
			Strength: EvidenceAdequate,
			Source:   "worker:delegation_fixture",
		}},
	})
}

func delegationReadyRequestInput(intensity ExecutionIntensity) DelegationRequestInput {
	return DelegationRequestInput{
		Activation:         ActivationManaged,
		RequestedIntensity: intensity,
		Frame: ObjectiveFrame{
			ID:             "objective:delegation_fixture",
			UserGoalDigest: "delegation fixture objective",
			ControlMode:    ControlModeDelegated,
			Intensity:      intensity,
			SuccessCriteria: []string{
				"worker result is verified before parent merge",
			},
			RequiredEvidence: []EvidenceRef{{
				Ref:      "evidence:delegation_worker_verified",
				Kind:     "delegation_worker_result",
				Strength: EvidenceAdequate,
				Source:   "worker:delegation_fixture",
			}},
		},
		SubgoalRef:                        "subgoal:delegation_fixture_research",
		WorkerRef:                         "worker:delegation_fixture_readonly",
		AllowedToolRefs:                   []DisplaySafeRef{"tool:read", "tool:search"},
		DeniedToolRefs:                    []DisplaySafeRef{"tool:exec"},
		Budget:                            ObjectiveBudgetSnapshot{BudgetRef: "budget:delegation_fixture", Limit: 2, Remaining: 2},
		EvidenceRequirements:              []EvidenceRef{{Ref: "evidence:delegation_worker_verified", Kind: "delegation_worker_result", Strength: EvidenceAdequate, Source: "worker:delegation_fixture"}},
		StopConditionRefs:                 []DisplaySafeRef{"stop:delegation_fixture_max_2_workers", "stop:delegation_fixture_parent_verified"},
		RedactionPolicyRef:                "redaction:delegation_fixture",
		MergePolicyRef:                    "merge:delegation_fixture_verify_before_merge",
		ExecutionContractAllowsDelegation: true,
		HostAllowsL4Delegation:            true,
		L5Enabled:                         false,
		UserConfirmed:                     true,
		HostApproved:                      true,
		ApprovalRefs:                      []DisplaySafeRef{"approval:delegation_fixture"},
		PolicyRefs:                        []DisplaySafeRef{"policy:delegation_fixture"},
		Boundaries:                        []Boundary{"delegation_fixture"},
	}
}

func delegationBoundaryContains(values []Boundary, want Boundary) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func delegationMissingInputContains(values []MissingInput, want MissingInput) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func delegationDisplayRefContains(values []DisplaySafeRef, want DisplaySafeRef) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func delegationStringContains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
