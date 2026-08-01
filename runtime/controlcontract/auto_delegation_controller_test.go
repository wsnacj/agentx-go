package controlcontract

import "testing"

func TestAutoDelegationControllerDispatchesReadyChildren(t *testing.T) {
	bridge := BuildAutoDelegationHostBridge(validAutoDelegationHostBridgeInput())

	decision := BuildAutoDelegationControllerDecision(AutoDelegationControllerInput{
		HostBridge:        bridge,
		RequestedDispatch: true,
	})

	if decision.Status != VerificationSatisfied ||
		decision.Action != AutoDelegationControllerActionSpawnOnce ||
		!decision.Ready ||
		!decision.HostMayDispatch ||
		len(decision.InvokableChildRefs) != 1 ||
		decision.InvokableChildRefs[0] != "child:collect_public_sources" ||
		decision.NextHostAction != "host_may_invoke_auto_delegation_children" {
		t.Fatalf("expected dispatch-ready decision: %+v", decision)
	}
	for _, boundary := range []Boundary{
		"auto_delegation_controller",
		"deterministic_child_lifecycle_reducer",
		"no_child_task_spawn_by_core",
		"no_subagent_dispatch_by_core",
		"auto_delegation_controller_spawn_once",
	} {
		if !autoDelegationBoundaryContains(decision.Boundaries, boundary) {
			t.Fatalf("missing boundary %q in %+v", boundary, decision.Boundaries)
		}
	}
}

func TestAutoDelegationControllerCollectsOpenChildrenAndRejectsRepeatedFanout(t *testing.T) {
	input := validAutoDelegationHostBridgeInput()
	input.ChildBindings[0].Active = true
	input.ChildBindings[0].ProgressSummary = "running read-only source check"
	bridge := BuildAutoDelegationHostBridge(input)

	decision := BuildAutoDelegationControllerDecision(AutoDelegationControllerInput{
		HostBridge:        bridge,
		RequestedDispatch: true,
	})

	if decision.Action != AutoDelegationControllerActionCollectExisting ||
		!decision.HostMayCollect ||
		decision.HostMayDispatch ||
		len(decision.OpenChildRefs) != 1 ||
		decision.OpenChildRefs[0] != "child:collect_public_sources" ||
		len(decision.RejectedActions) != 1 ||
		decision.RejectedActions[0] != AutoDelegationControllerActionSpawnOnce ||
		!autoDelegationHostBridgeStringContains(decision.BlockedReasons, "auto_delegation_repeated_fanout_rejected") ||
		!autoDelegationBoundaryContains(decision.Boundaries, "auto_delegation_repeated_fanout_rejected") {
		t.Fatalf("expected open child collection and repeated fanout rejection: %+v", decision)
	}
}

func TestAutoDelegationControllerReplaysFailedChildrenWithinBudget(t *testing.T) {
	input := validAutoDelegationHostBridgeInput()
	input.ChildBindings[0].Failed = true
	input.ChildBindings[0].FailureClass = FailureTimeout
	bridge := BuildAutoDelegationHostBridge(input)

	decision := BuildAutoDelegationControllerDecision(AutoDelegationControllerInput{
		HostBridge: bridge,
		ChildAttempts: []AutoDelegationControllerChildState{{
			ChildRef:         "child:collect_public_sources",
			Attempts:         0,
			LastAttemptRef:   "attempt:child_collect_public_sources_0",
			LastFailureClass: FailureTimeout,
		}},
	})

	if decision.Action != AutoDelegationControllerActionReplayOnce ||
		!decision.HostMayReplay ||
		len(decision.ReplayChildRefs) != 1 ||
		decision.ReplayChildRefs[0] != "child:collect_public_sources" ||
		decision.NextHostAction != "retry_auto_delegation_children" ||
		!autoDelegationBoundaryContains(decision.Boundaries, "auto_delegation_controller_replay_once") {
		t.Fatalf("expected replay decision within budget: %+v", decision)
	}
}

func TestAutoDelegationControllerBlocksWhenRetryBudgetExhaustedWithoutPartialEvidence(t *testing.T) {
	input := validAutoDelegationHostBridgeInput()
	input.ChildBindings[0].Failed = true
	input.ChildBindings[0].FailureClass = FailureTimeout
	bridge := BuildAutoDelegationHostBridge(input)

	decision := BuildAutoDelegationControllerDecision(AutoDelegationControllerInput{
		HostBridge:          bridge,
		MaxAttemptsPerChild: 1,
		ChildAttempts: []AutoDelegationControllerChildState{{
			ChildRef:         "child:collect_public_sources",
			Attempts:         1,
			LastAttemptRef:   "attempt:child_collect_public_sources_1",
			LastFailureClass: FailureTimeout,
		}},
	})

	if decision.Action != AutoDelegationControllerActionBlock ||
		decision.Ready ||
		decision.FailureClass != FailureRepeatedNoProgress ||
		len(decision.ExhaustedChildRefs) != 1 ||
		decision.ExhaustedChildRefs[0] != "child:collect_public_sources" ||
		!autoDelegationMissingInputContains(decision.MissingInputs, "host:auto_delegation_retry_budget") ||
		!autoDelegationBoundaryContains(decision.Boundaries, "auto_delegation_controller_retry_budget_exhausted") {
		t.Fatalf("expected retry budget exhaustion block: %+v", decision)
	}
}

func TestAutoDelegationControllerCollectsTerminalChildrenBeforeParentMerge(t *testing.T) {
	input := validAutoDelegationHostBridgeInput()
	input.ChildBindings[0].Completed = true
	input.ChildBindings[0].CompletionSummary = "child returned display-safe evidence"
	bridge := BuildAutoDelegationHostBridge(input)

	decision := BuildAutoDelegationControllerDecision(AutoDelegationControllerInput{
		HostBridge: bridge,
	})

	if decision.Action != AutoDelegationControllerActionCollectExisting ||
		!decision.HostMayCollect ||
		len(decision.CompletedChildRefs) != 1 ||
		decision.CompletedChildRefs[0] != "child:collect_public_sources" ||
		!autoDelegationMissingInputContains(decision.MissingInputs, "host:auto_delegation_parent_merge") ||
		decision.NextHostAction != "provide_auto_delegation_parent_merge" {
		t.Fatalf("expected terminal child collection before parent merge: %+v", decision)
	}
}

func TestAutoDelegationControllerCompletesAfterAcceptedParentMerge(t *testing.T) {
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

	decision := BuildAutoDelegationControllerDecision(AutoDelegationControllerInput{
		HostBridge:  bridge,
		ParentMerge: merge,
	})

	if decision.Status != VerificationSatisfied ||
		decision.Action != AutoDelegationControllerActionSatisfied ||
		!decision.ReadyForCloseout ||
		!decision.HostMayCollect ||
		len(decision.MergedChildRefs) != 1 ||
		decision.MergedChildRefs[0] != "child:collect_public_sources" ||
		decision.NextHostAction != "update_objective_controller" {
		t.Fatalf("expected complete decision after accepted parent merge: %+v", decision)
	}
}

func TestAutoDelegationControllerReplaysPartialMergeChildrenBeforePartialCloseout(t *testing.T) {
	input := validAutoDelegationHostBridgeInput()
	input.PlanReview = twoChildAutoDelegationPlanReview(t)
	input.MaxParallelism = 2
	input.ChildBindings = append(input.ChildBindings, autoDelegationHostBridgeChildBinding("child:cross_check_sources"))
	bridge := BuildAutoDelegationHostBridge(input)
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

	decision := BuildAutoDelegationControllerDecision(AutoDelegationControllerInput{
		HostBridge:          bridge,
		ParentMerge:         merge,
		MaxAttemptsPerChild: 1,
		ChildAttempts: []AutoDelegationControllerChildState{{
			ChildRef:         "child:cross_check_sources",
			Attempts:         0,
			LastFailureClass: FailureEvidenceMissing,
		}},
	})

	if decision.Action != AutoDelegationControllerActionReplayOnce ||
		!decision.HostMayReplay ||
		decision.ReadyForCloseout ||
		len(decision.ReplayChildRefs) != 1 ||
		decision.ReplayChildRefs[0] != "child:cross_check_sources" ||
		len(decision.MergedChildRefs) != 1 {
		t.Fatalf("expected retry before partial closeout when budget remains: %+v", decision)
	}

	exhausted := BuildAutoDelegationControllerDecision(AutoDelegationControllerInput{
		HostBridge:          bridge,
		ParentMerge:         merge,
		MaxAttemptsPerChild: 1,
		ChildAttempts: []AutoDelegationControllerChildState{{
			ChildRef:         "child:cross_check_sources",
			Attempts:         1,
			LastFailureClass: FailureEvidenceMissing,
		}},
	})
	if exhausted.Action != AutoDelegationControllerActionPartialCloseout ||
		!exhausted.ReadyForCloseout ||
		len(exhausted.ExhaustedChildRefs) != 1 ||
		exhausted.ExhaustedChildRefs[0] != "child:cross_check_sources" {
		t.Fatalf("expected partial closeout after retry budget exhaustion: %+v", exhausted)
	}
}

func TestAutoDelegationControllerRejectsUnsafeRefs(t *testing.T) {
	bridge := BuildAutoDelegationHostBridge(validAutoDelegationHostBridgeInput())
	bridge.ActiveChildRefs = []DisplaySafeRef{"https://example.invalid/raw-child"}

	decision := BuildAutoDelegationControllerDecision(AutoDelegationControllerInput{
		HostBridge: bridge,
	})

	if decision.Status != VerificationReviewRequired ||
		decision.Action != AutoDelegationControllerActionBlock ||
		decision.Ready ||
		!autoDelegationMissingInputContains(decision.MissingInputs, "host:display_safe_refs") ||
		!autoDelegationBoundaryContains(decision.Boundaries, "raw_output_not_allowed") {
		t.Fatalf("unsafe refs should require review: %+v", decision)
	}
}
