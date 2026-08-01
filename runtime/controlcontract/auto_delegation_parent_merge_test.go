package controlcontract

import "testing"

func TestAutoDelegationParentMergeAcceptsVerifiedChildEvidence(t *testing.T) {
	bridge := autoDelegationParentMergeReadyBridge(t)
	invocation := autoDelegationParentMergeCompletedInvocation(t, bridge.Children[0].Readiness)
	merge := BuildAutoDelegationParentMerge(AutoDelegationParentMergeInput{
		HostBridge:       bridge,
		ParentFrame:      autoDelegationParentMergeFrameFixture(),
		ParentLedgerRef:  "ledger:auto_delegation_parent",
		FailureReviewRef: "failure_review:auto_delegation_parent",
		FailureRef:       "failure:auto_delegation_parent",
		ChildResults: []AutoDelegationChildMergeInput{
			autoDelegationChildMergeInputFixture("child:collect_public_sources", invocation, "value:one", EvidenceAdequate),
		},
	})

	if merge.Status != VerificationSatisfied ||
		merge.Decision != AutoDelegationParentMergeAccept ||
		!merge.ReadyForParentMerge ||
		!merge.ParentAnswerMayUseChildEvidence ||
		merge.WorkerOutputAcceptedAsFact ||
		!merge.WorkerResultRequiresVerification ||
		len(merge.MergedChildRefs) != 1 ||
		merge.MergedChildRefs[0] != "child:collect_public_sources" ||
		len(merge.ParentLedgerPatch.Attempts) != 1 ||
		merge.ParentLedgerPatch.Attempts[0].Status != VerificationSatisfied {
		t.Fatalf("unexpected accepted parent merge: %+v", merge)
	}
	for _, boundary := range []Boundary{
		"auto_delegation_parent_merge_ready",
		"auto_delegation_child_result_verified",
		"no_child_task_spawn_by_core",
		"no_subagent_dispatch_by_core",
		"no_store_mutation_by_core",
	} {
		if !autoDelegationBoundaryContains(merge.Boundaries, boundary) {
			t.Fatalf("missing boundary %q in %+v", boundary, merge.Boundaries)
		}
	}
}

func TestAutoDelegationParentMergeReturnsPartialWhenOneChildIsMissing(t *testing.T) {
	input := validAutoDelegationHostBridgeInput()
	input.PlanReview = twoChildAutoDelegationPlanReview(t)
	input.MaxParallelism = 2
	input.ChildBindings = append(input.ChildBindings, autoDelegationHostBridgeChildBinding("child:cross_check_sources"))
	bridge := BuildAutoDelegationHostBridge(input)
	if len(bridge.Children) != 2 || !bridge.Ready {
		t.Fatalf("expected ready two-child bridge: %+v", bridge)
	}
	invocation := autoDelegationParentMergeCompletedInvocation(t, bridge.Children[0].Readiness)
	merge := BuildAutoDelegationParentMerge(AutoDelegationParentMergeInput{
		HostBridge:       bridge,
		ParentFrame:      autoDelegationParentMergeFrameFixture(),
		ParentLedgerRef:  "ledger:auto_delegation_parent",
		FailureReviewRef: "failure_review:auto_delegation_parent",
		FailureRef:       "failure:auto_delegation_parent",
		ChildResults: []AutoDelegationChildMergeInput{
			autoDelegationChildMergeInputFixture("child:collect_public_sources", invocation, "value:one", EvidenceAdequate),
		},
	})

	if merge.Status != VerificationPartial ||
		merge.Decision != AutoDelegationParentMergePartial ||
		!merge.ReadyForParentMerge ||
		!merge.ParentAnswerMayUseChildEvidence ||
		len(merge.MergedChildRefs) != 1 ||
		len(merge.RetryChildRefs) != 1 ||
		merge.RetryChildRefs[0] != "child:cross_check_sources" ||
		!merge.MissingEvidenceDetected {
		t.Fatalf("expected partial parent merge with one retry child: %+v", merge)
	}
}

func TestAutoDelegationParentMergeDetectsWeakAndStaleEvidence(t *testing.T) {
	bridge := autoDelegationParentMergeReadyBridge(t)
	invocation := autoDelegationParentMergeCompletedInvocation(t, bridge.Children[0].Readiness)
	weak := autoDelegationChildMergeInputFixture("child:collect_public_sources", invocation, "value:one", EvidenceWeak)
	weak.StaleEvidenceDetected = true
	weak.StaleEvidenceRefs = []DisplaySafeRef{"evidence:auto_delegation_child"}
	weak.AlternatePathRefs = []DisplaySafeRef{"path:auto_delegation_child_alternate"}
	merge := BuildAutoDelegationParentMerge(AutoDelegationParentMergeInput{
		HostBridge:       bridge,
		ParentFrame:      autoDelegationParentMergeFrameFixture(),
		ParentLedgerRef:  "ledger:auto_delegation_parent",
		FailureReviewRef: "failure_review:auto_delegation_parent",
		FailureRef:       "failure:auto_delegation_parent",
		ChildResults:     []AutoDelegationChildMergeInput{weak},
	})

	if merge.Decision != AutoDelegationParentMergeAlternatePath ||
		merge.ReadyForParentMerge ||
		!merge.StaleEvidenceDetected ||
		!merge.WeakEvidenceDetected ||
		len(merge.AlternatePathChildRefs) != 1 ||
		merge.AlternatePathChildRefs[0] != "child:collect_public_sources" ||
		!autoDelegationBoundaryContains(merge.Boundaries, "auto_delegation_child_evidence_stale") {
		t.Fatalf("expected stale weak evidence to require alternate path: %+v", merge)
	}
}

func TestAutoDelegationParentMergeDetectsConflictingChildResults(t *testing.T) {
	input := validAutoDelegationHostBridgeInput()
	input.PlanReview = twoChildAutoDelegationPlanReview(t)
	input.MaxParallelism = 2
	input.ChildBindings = append(input.ChildBindings, autoDelegationHostBridgeChildBinding("child:cross_check_sources"))
	bridge := BuildAutoDelegationHostBridge(input)
	firstInvocation := autoDelegationParentMergeCompletedInvocation(t, bridge.Children[0].Readiness)
	secondInvocation := autoDelegationParentMergeCompletedInvocation(t, bridge.Children[1].Readiness)
	merge := BuildAutoDelegationParentMerge(AutoDelegationParentMergeInput{
		HostBridge:       bridge,
		ParentFrame:      autoDelegationParentMergeFrameFixture(),
		ParentLedgerRef:  "ledger:auto_delegation_parent",
		FailureReviewRef: "failure_review:auto_delegation_parent",
		FailureRef:       "failure:auto_delegation_parent",
		ChildResults: []AutoDelegationChildMergeInput{
			autoDelegationChildMergeInputFixture("child:collect_public_sources", firstInvocation, "42", EvidenceAdequate),
			autoDelegationChildMergeInputFixture("child:cross_check_sources", secondInvocation, "7", EvidenceAdequate),
		},
	})

	if merge.Status == VerificationSatisfied ||
		merge.Decision != AutoDelegationParentMergeBlock ||
		!merge.ConflictDetected ||
		len(merge.ConflictRefs) == 0 ||
		merge.FailureReview == nil ||
		!merge.FailureReview.ConflictingResultsDetected ||
		!autoDelegationBoundaryContains(merge.Boundaries, "auto_delegation_child_result_conflict") {
		t.Fatalf("expected conflicting child results to block parent merge: %+v", merge)
	}
}

func TestAutoDelegationParentMergeRejectsUnsafeRefs(t *testing.T) {
	bridge := autoDelegationParentMergeReadyBridge(t)
	invocation := autoDelegationParentMergeCompletedInvocation(t, bridge.Children[0].Readiness)
	child := autoDelegationChildMergeInputFixture("child:collect_public_sources", invocation, "value:one", EvidenceAdequate)
	child.MergeRef = "https://example.invalid/raw-merge"
	merge := BuildAutoDelegationParentMerge(AutoDelegationParentMergeInput{
		HostBridge:      bridge,
		ParentFrame:     autoDelegationParentMergeFrameFixture(),
		ParentLedgerRef: "ledger:auto_delegation_parent",
		ChildResults:    []AutoDelegationChildMergeInput{child},
	})

	if merge.Status != VerificationReviewRequired ||
		merge.Decision != AutoDelegationParentMergeBlock ||
		merge.ReadyForParentMerge ||
		merge.ParentAnswerMayUseChildEvidence ||
		!autoDelegationMissingInputContains(merge.MissingInputs, "host:display_safe_refs") ||
		!autoDelegationBoundaryContains(merge.Boundaries, "raw_output_not_allowed") {
		t.Fatalf("unsafe merge refs should require review: %+v", merge)
	}
}

func autoDelegationParentMergeReadyBridge(t *testing.T) AutoDelegationHostBridge {
	t.Helper()
	bridge := BuildAutoDelegationHostBridge(validAutoDelegationHostBridgeInput())
	if bridge.Status != HostActionReady || len(bridge.Children) != 1 {
		t.Fatalf("expected ready bridge fixture: %+v", bridge)
	}
	return bridge
}

func autoDelegationParentMergeFrameFixture() ObjectiveFrame {
	return ObjectiveFrame{
		ID:        "objective:root",
		Intensity: IntensityL4DurableLongRun,
		RequiredEvidence: []EvidenceRef{{
			Ref:      "evidence:public_source_summary",
			Kind:     "summary",
			Strength: EvidenceAdequate,
			Source:   "capability:public_source",
		}},
	}
}

func autoDelegationParentMergeCompletedInvocation(t *testing.T, readiness HostOwnedDelegationWorkerRuntimeReadiness) HostOwnedDelegationWorkerRuntimeInvocation {
	t.Helper()
	readiness = readiness.Normalize()
	invocation := BuildHostOwnedDelegationWorkerRuntimeInvocation(HostOwnedDelegationWorkerRuntimeInvocationInput{
		Readiness:               readiness,
		InvocationReportRef:     DisplaySafeRef("invocation_report:" + string(readiness.SubgoalRef)),
		ObservedInvocationRef:   readiness.InvocationRef,
		HostWorkerRuntimeRunRef: DisplaySafeRef("worker_runtime_run:" + string(readiness.SubgoalRef)),
		ObservedWorkerRunRef:    readiness.WorkerRunRef,
		WorkerResultRef:         DisplaySafeRef("worker_result:" + string(readiness.SubgoalRef)),
		WorkerReadbackRef:       DisplaySafeRef("worker_readback:" + string(readiness.SubgoalRef)),
		ObservationRef:          DisplaySafeRef("observation:" + string(readiness.SubgoalRef)),
		HostInvocationReported:  true,
		HostInvocationCompleted: true,
		EvidenceRefs: []EvidenceRef{{
			Ref:      "evidence:public_source_summary",
			Kind:     "summary",
			Strength: EvidenceAdequate,
			Source:   "capability:public_source",
		}},
	})
	if !invocation.ReadyForWorkerResultReview {
		t.Fatalf("expected invocation ready for parent review: %+v", invocation)
	}
	return invocation
}

func autoDelegationChildMergeInputFixture(childRef DisplaySafeRef, invocation HostOwnedDelegationWorkerRuntimeInvocation, value string, strength EvidenceStrength) AutoDelegationChildMergeInput {
	return AutoDelegationChildMergeInput{
		ChildRef:          childRef,
		Invocation:        invocation,
		ParentLedgerRef:   "ledger:auto_delegation_parent",
		WorkerAttemptRef:  AttemptRef("attempt:" + string(childRef)),
		MergeRef:          DisplaySafeRef("merge:" + string(childRef)),
		MergePolicyRef:    "merge:auto_delegation_child",
		RequiredEvidence:  autoDelegationParentMergeFrameFixture().RequiredEvidence,
		EvidenceRefs:      []EvidenceRef{autoDelegationChildEvidence(strength)},
		AlternatePathRefs: []DisplaySafeRef{"path:auto_delegation_child_retry"},
		WorkerObservations: []Observation{{
			Kind:     "auto_delegation_child_result",
			Source:   "worker:auto_delegation_runtime",
			Subject:  "objective:root",
			Name:     "child_result",
			Value:    value,
			Strength: strength,
			DisplaySafeRefs: []DisplaySafeRef{
				DisplaySafeRef("observation:" + string(childRef)),
			},
			EvidenceRefs: []EvidenceRef{autoDelegationChildEvidence(strength)},
		}},
		ExpectedObservationKinds: []string{"auto_delegation_child_result"},
	}
}

func autoDelegationChildEvidence(strength EvidenceStrength) EvidenceRef {
	return EvidenceRef{
		Ref:      "evidence:public_source_summary",
		Kind:     "summary",
		Strength: strength,
		Source:   "capability:public_source",
	}
}
