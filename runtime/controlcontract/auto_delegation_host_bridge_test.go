package controlcontract

import "testing"

func TestAutoDelegationHostBridgeBuildsReadyHostOwnedWorkerRuntime(t *testing.T) {
	bridge := BuildAutoDelegationHostBridge(validAutoDelegationHostBridgeInput())

	if bridge.Status != HostActionReady ||
		!bridge.Ready ||
		!bridge.HostMayInvokeWorkerRuntime ||
		bridge.FailureClass != FailureNone ||
		len(bridge.InvokableChildRefs) != 1 ||
		bridge.InvokableChildRefs[0] != "child:collect_public_sources" {
		t.Fatalf("unexpected ready bridge: %+v", bridge)
	}
	if len(bridge.Children) != 1 ||
		!bridge.Children[0].Ready ||
		bridge.Children[0].Queued ||
		!bridge.Children[0].Readiness.HostMayInvokeWorkerRuntime ||
		!bridge.Children[0].Request.ReadyForWorkerDispatch {
		t.Fatalf("unexpected child runtime readiness: %+v", bridge.Children)
	}
	if bridge.Children[0].WorkerOutputAcceptedAsFact ||
		!bridge.Children[0].WorkerResultRequiresVerification ||
		bridge.WorkerOutputAcceptedAsFact ||
		!bridge.WorkerResultRequiresVerification {
		t.Fatalf("child output must remain parent-verified: %+v", bridge)
	}
	if !autoDelegationBoundaryContains(bridge.Boundaries, "no_child_task_spawn_by_core") ||
		!autoDelegationBoundaryContains(bridge.Boundaries, "no_subagent_dispatch_by_core") ||
		!autoDelegationBoundaryContains(bridge.Children[0].Boundaries, "child_leaf_no_recursive_delegation") ||
		!autoDelegationBoundaryContains(bridge.Children[0].Boundaries, "auto_delegation_child_ready_for_host_runtime") {
		t.Fatalf("expected host-owned/no-core-dispatch boundaries: %+v child=%+v", bridge.Boundaries, bridge.Children[0].Boundaries)
	}
}

func TestAutoDelegationHostBridgeQueuesBeyondParallelismWindow(t *testing.T) {
	input := validAutoDelegationHostBridgeInput()
	input.PlanReview = twoChildAutoDelegationPlanReview(t)
	input.MaxParallelism = 1
	input.ChildBindings = append(input.ChildBindings, autoDelegationHostBridgeChildBinding("child:cross_check_sources"))

	bridge := BuildAutoDelegationHostBridge(input)

	if bridge.Status != HostActionReady ||
		!bridge.HostMayInvokeWorkerRuntime ||
		len(bridge.InvokableChildRefs) != 1 ||
		len(bridge.QueuedChildRefs) != 1 ||
		bridge.QueuedChildRefs[0] != "child:cross_check_sources" {
		t.Fatalf("expected one invokable child and one queued child: %+v", bridge)
	}
	if len(bridge.Children) != 2 ||
		!bridge.Children[1].Queued ||
		!autoDelegationBoundaryContains(bridge.Children[1].Boundaries, "auto_delegation_parallelism_window_queued") {
		t.Fatalf("second child should be queued by parallelism: %+v", bridge.Children)
	}
}

func TestAutoDelegationHostBridgeBlocksChildAttemptAndTimeoutLimit(t *testing.T) {
	input := validAutoDelegationHostBridgeInput()
	plan := validAutoDelegationPlan()
	plan.Children[0].AllowedToolRefs = []DisplaySafeRef{"tool:public_source_read"}
	plan.Children[0].MaxAttempts = 2
	plan.Children[0].MaxDurationSeconds = 121
	policyReview := BuildAutoDelegationPolicyReview(AutoDelegationPolicy{
		Mode:                    AutoDelegationManagedReadOnly,
		MaxChildren:             1,
		MaxParallelism:          1,
		MaxAttemptsPerChild:     1,
		MaxDurationSeconds:      120,
		AllowedToolRefs:         []DisplaySafeRef{"tool:public_source_read"},
		AllowBackgroundReadOnly: true,
	})
	input.PlanReview = BuildAutoDelegationPlanReview(policyReview, plan)

	bridge := BuildAutoDelegationHostBridge(input)

	if bridge.Status == HostActionReady ||
		bridge.HostMayInvokeWorkerRuntime ||
		!autoDelegationHostBridgeStringContains(bridge.BlockedReasons, "auto_delegation_child_attempt_limit_exceeded") ||
		!autoDelegationHostBridgeStringContains(bridge.BlockedReasons, "auto_delegation_child_timeout_exceeded") {
		t.Fatalf("expected child attempt and timeout limits to block: %+v", bridge)
	}
}

func TestAutoDelegationHostBridgeFiltersChildToolsAndBlockedTools(t *testing.T) {
	input := validAutoDelegationHostBridgeInput()
	input.AllowedToolRefs = []DisplaySafeRef{"tool:public_source_read", "tool:browser_read"}
	input.DeniedToolRefs = []DisplaySafeRef{"tool:browser_read"}
	input.ChildBindings[0].AllowedToolRefs = []DisplaySafeRef{"tool:public_source_read", "tool:browser_read"}
	input.ChildBindings[0].DeniedToolRefs = []DisplaySafeRef{"tool:secret_write"}

	bridge := BuildAutoDelegationHostBridge(input)

	if bridge.Status != HostActionReady || len(bridge.Children) != 1 {
		t.Fatalf("expected ready bridge: %+v", bridge)
	}
	got := bridge.Children[0].EffectiveAllowedToolRefs
	if len(got) != 1 || got[0] != "tool:public_source_read" {
		t.Fatalf("expected denied tool to be filtered from child tool boundary, got %+v", got)
	}
	if !autoDelegationHostBridgeRefContains(bridge.Children[0].EffectiveDeniedToolRefs, "tool:browser_read") ||
		!autoDelegationHostBridgeRefContains(bridge.Children[0].EffectiveDeniedToolRefs, "tool:secret_write") {
		t.Fatalf("expected blocked tools to be recorded: %+v", bridge.Children[0].EffectiveDeniedToolRefs)
	}
}

func TestAutoDelegationHostBridgeReportsCancellationAndFailureSurfaces(t *testing.T) {
	cancelledInput := validAutoDelegationHostBridgeInput()
	cancelledInput.ChildBindings[0].Cancelled = true
	cancelledInput.ChildBindings[0].CancellationRef = "cancel:child_collect_sources"
	cancelled := BuildAutoDelegationHostBridge(cancelledInput)
	if len(cancelled.CancelledChildRefs) != 1 ||
		cancelled.CancelledChildRefs[0] != "child:collect_public_sources" ||
		!autoDelegationBoundaryContains(cancelled.Children[0].Boundaries, "auto_delegation_child_cancelled") {
		t.Fatalf("expected cancellation surface: %+v", cancelled)
	}

	failedInput := validAutoDelegationHostBridgeInput()
	failedInput.ChildBindings[0].Failed = true
	failedInput.ChildBindings[0].FailureBindingRef = "failure:child_collect_sources"
	failed := BuildAutoDelegationHostBridge(failedInput)
	if len(failed.FailedChildRefs) != 1 ||
		failed.FailedChildRefs[0] != "child:collect_public_sources" ||
		!autoDelegationBoundaryContains(failed.Children[0].Boundaries, "auto_delegation_child_failed") {
		t.Fatalf("expected failure surface: %+v", failed)
	}
}

func TestAutoDelegationHostBridgeRequiresBoundRefsAfterDispatch(t *testing.T) {
	input := validAutoDelegationHostBridgeInput()
	input.ChildBindings[0].Active = true
	input.ChildBindings[0].BoundAllowedToolRefs = nil
	input.ChildBindings[0].BoundCapabilityRefs = nil

	bridge := BuildAutoDelegationHostBridge(input)

	if bridge.Status == HostActionReady ||
		bridge.HostMayInvokeWorkerRuntime ||
		!autoDelegationHostBridgeStringContains(bridge.BlockedReasons, "auto_delegation_child_bound_capability_refs_missing") ||
		!autoDelegationHostBridgeStringContains(bridge.BlockedReasons, "auto_delegation_child_bound_tool_refs_missing") ||
		!autoDelegationMissingInputContains(bridge.MissingInputs, "host:auto_delegation_bound_capability_refs") ||
		!autoDelegationMissingInputContains(bridge.MissingInputs, "host:auto_delegation_bound_allowed_tool_refs") {
		t.Fatalf("expected dispatched child to require immutable bound refs: %+v", bridge)
	}
}

func TestAutoDelegationHostBridgeBlocksBoundRefWideningAfterDispatch(t *testing.T) {
	input := validAutoDelegationHostBridgeInput()
	input.ChildBindings[0].Active = true
	input.ChildBindings[0].BoundAllowedToolRefs = []DisplaySafeRef{"tool:public_source_read", "tool:browser_read", "tool:secret_write"}
	input.ChildBindings[0].BoundCapabilityRefs = []DisplaySafeRef{"capability:public_source", "capability:write_access"}

	bridge := BuildAutoDelegationHostBridge(input)

	if bridge.Status == HostActionReady ||
		bridge.HostMayInvokeWorkerRuntime ||
		!autoDelegationHostBridgeStringContains(bridge.BlockedReasons, "auto_delegation_child_capability_refs_widened_after_dispatch") ||
		!autoDelegationHostBridgeStringContains(bridge.BlockedReasons, "auto_delegation_child_tool_refs_widened_after_dispatch") ||
		!autoDelegationBoundaryContains(bridge.Boundaries, "auto_delegation_child_capability_refs_immutable_after_dispatch") ||
		!autoDelegationBoundaryContains(bridge.Boundaries, "auto_delegation_child_tool_refs_immutable_after_dispatch") {
		t.Fatalf("expected widened bound refs to block dispatched child: %+v", bridge)
	}
}

func TestAutoDelegationHostBridgeRejectsUnsafeRefs(t *testing.T) {
	input := validAutoDelegationHostBridgeInput()
	input.ChildBindings[0].WorkerRunRef = "https://example.invalid/raw-worker-run"

	bridge := BuildAutoDelegationHostBridge(input)

	if bridge.Status != HostActionReviewRequired ||
		bridge.Ready ||
		bridge.HostMayInvokeWorkerRuntime ||
		bridge.FailureClass != FailureEvidenceWeak ||
		!autoDelegationMissingInputContains(bridge.MissingInputs, "host:display_safe_refs") ||
		!autoDelegationBoundaryContains(bridge.Boundaries, "raw_output_not_allowed") {
		t.Fatalf("unsafe refs should force review: %+v", bridge)
	}
}

func validAutoDelegationHostBridgeInput() AutoDelegationHostBridgeInput {
	plan := validAutoDelegationPlan()
	plan.Children[0].AllowedToolRefs = []DisplaySafeRef{"tool:public_source_read", "tool:browser_read"}
	policyReview := BuildAutoDelegationPolicyReview(AutoDelegationPolicy{
		Mode:                    AutoDelegationManagedReadOnly,
		MaxChildren:             2,
		MaxParallelism:          1,
		MaxAttemptsPerChild:     1,
		MaxDurationSeconds:      120,
		MaxBudgetTokens:         100,
		AllowedToolRefs:         []DisplaySafeRef{"tool:public_source_read", "tool:browser_read"},
		DeniedToolRefs:          []DisplaySafeRef{"tool:blocked_write"},
		AllowBackgroundReadOnly: true,
	})
	return AutoDelegationHostBridgeInput{
		PlanReview: BuildAutoDelegationPlanReview(policyReview, plan),
		Activation: ActivationManaged,
		ParentFrame: ObjectiveFrame{
			ID:        "objective:root",
			Intensity: IntensityL4DurableLongRun,
		},
		Budget: ObjectiveBudgetSnapshot{
			BudgetRef: "budget:auto_delegation",
			Limit:     100,
			Remaining: 100,
		},
		WorkerRuntimeGate: BuildProductionAdapterIndependentEffectGate(ProductionAdapterIndependentEffectGateSpec{
			Kind:                  ProductionAdapterEffectGateDelegationWorker,
			GateRef:               "gate:auto_delegation_worker_runtime",
			AdapterRef:            "adapter:auto_delegation_worker_runtime",
			ContractRef:           "contract:auto_delegation_worker_runtime",
			PolicyRef:             "policy:auto_delegation_worker_runtime",
			ApprovalRef:           "approval:auto_delegation_worker_runtime",
			BudgetRef:             "budget:auto_delegation_worker_runtime",
			IdempotencyRef:        "idempotency:auto_delegation_worker_runtime",
			ReadbackRef:           "readback:auto_delegation_worker_runtime",
			EvalRef:               "eval:auto_delegation_worker_runtime",
			FailureReviewRef:      "review:auto_delegation_worker_failure",
			CompensationReviewRef: "review:auto_delegation_worker_compensation",
			EvidenceRefs:          []DisplaySafeRef{"evidence:auto_delegation_worker_runtime"},
		}),
		RequestedIntensity:                IntensityL4DurableLongRun,
		ExecutionContractAllowsDelegation: true,
		HostAllowsL4Delegation:            true,
		UserConfirmed:                     true,
		HostApproved:                      true,
		AllowedToolRefs:                   []DisplaySafeRef{"tool:public_source_read", "tool:browser_read"},
		ApprovalRefs:                      []DisplaySafeRef{"approval:auto_delegation"},
		PolicyRefs:                        []DisplaySafeRef{"policy:auto_delegation"},
		StopConditionRefs:                 []DisplaySafeRef{"stop:auto_delegation_child"},
		RedactionPolicyRef:                "redaction:auto_delegation_child",
		MergePolicyRef:                    "merge:auto_delegation_child",
		AdapterRef:                        "adapter:auto_delegation_worker_runtime",
		AdapterVersionRef:                 "adapter_version:auto_delegation_worker_runtime_v1",
		AdapterCapabilityRef:              "capability:auto_delegation_worker_runtime",
		AdapterContractRef:                "contract:auto_delegation_worker_runtime",
		HostConfirmationRef:               "confirmation:auto_delegation_worker_runtime",
		WorkerRuntimeRef:                  "worker:auto_delegation_runtime",
		IdempotencyRef:                    "idempotency:auto_delegation_child",
		BudgetRef:                         "budget:auto_delegation_child",
		VerificationRef:                   "verification:auto_delegation_parent",
		FailureReviewRef:                  "review:auto_delegation_child_failure",
		FailureBindingRef:                 "failure:auto_delegation_child",
		CompensationRef:                   "compensation:auto_delegation_child",
		CancellationRef:                   "cancel:auto_delegation_child",
		ChildBindings: []AutoDelegationHostChildBinding{
			autoDelegationHostBridgeChildBinding("child:collect_public_sources"),
		},
	}
}

func twoChildAutoDelegationPlanReview(t *testing.T) AutoDelegationPlanReview {
	t.Helper()
	plan := validAutoDelegationPlan()
	plan.Children[0].AllowedToolRefs = []DisplaySafeRef{"tool:public_source_read"}
	second := plan.Children[0]
	second.ChildRef = "child:cross_check_sources"
	second.Goal = "Cross-check the first child result against an independent read-only source."
	second.ExpectedOutput = "A bounded cross-check summary with display-safe evidence references."
	second.AllowedToolRefs = []DisplaySafeRef{"tool:public_source_read"}
	plan.Children = append(plan.Children, second)
	policyReview := BuildAutoDelegationPolicyReview(AutoDelegationPolicy{
		Mode:                    AutoDelegationManagedReadOnly,
		MaxChildren:             2,
		MaxParallelism:          1,
		MaxAttemptsPerChild:     1,
		MaxDurationSeconds:      120,
		MaxBudgetTokens:         100,
		AllowedToolRefs:         []DisplaySafeRef{"tool:public_source_read"},
		AllowBackgroundReadOnly: true,
	})
	review := BuildAutoDelegationPlanReview(policyReview, plan)
	if !review.Ready || !review.HostMayDispatch {
		t.Fatalf("two-child plan review should be dispatchable: %+v", review)
	}
	return review
}

func autoDelegationHostBridgeChildBinding(childRef DisplaySafeRef) AutoDelegationHostChildBinding {
	return AutoDelegationHostChildBinding{
		ChildRef:             childRef,
		WorkerRunRef:         DisplaySafeRef("worker_run:" + string(childRef)),
		WorkerRequestRef:     DisplaySafeRef("worker_request:" + string(childRef)),
		InvocationRef:        DisplaySafeRef("invocation:" + string(childRef)),
		ResultBindingRef:     DisplaySafeRef("worker_result:" + string(childRef)),
		ReadbackBindingRef:   DisplaySafeRef("worker_readback:" + string(childRef)),
		IdempotencyRef:       DisplaySafeRef("idempotency:" + string(childRef)),
		BudgetRef:            DisplaySafeRef("budget:" + string(childRef)),
		VerificationRef:      DisplaySafeRef("verification:" + string(childRef)),
		FailureBindingRef:    DisplaySafeRef("failure:" + string(childRef)),
		CompensationRef:      DisplaySafeRef("compensation:" + string(childRef)),
		CancellationRef:      DisplaySafeRef("cancel:" + string(childRef)),
		ProgressSummary:      "not started",
		LineageRefs:          []DisplaySafeRef{"objective:root", childRef},
		ContextRefs:          []DisplaySafeRef{"context:user_goal"},
		AllowedToolRefs:      []DisplaySafeRef{"tool:public_source_read", "tool:browser_read"},
		BoundAllowedToolRefs: []DisplaySafeRef{"tool:public_source_read", "tool:browser_read"},
		BoundCapabilityRefs:  []DisplaySafeRef{"capability:public_source"},
		StopConditionRefs:    []DisplaySafeRef{"stop:auto_delegation_child"},
		RedactionPolicyRef:   "redaction:auto_delegation_child",
		MergePolicyRef:       "merge:auto_delegation_child",
		ApprovalRefs:         []DisplaySafeRef{"approval:auto_delegation"},
		PolicyRefs:           []DisplaySafeRef{"policy:auto_delegation"},
		AdapterRef:           "adapter:auto_delegation_worker_runtime",
		AdapterVersionRef:    "adapter_version:auto_delegation_worker_runtime_v1",
		AdapterCapabilityRef: "capability:auto_delegation_worker_runtime",
		AdapterContractRef:   "contract:auto_delegation_worker_runtime",
		HostConfirmationRef:  "confirmation:auto_delegation_worker_runtime",
	}
}

func autoDelegationHostBridgeStringContains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func autoDelegationHostBridgeRefContains(values []DisplaySafeRef, want DisplaySafeRef) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
