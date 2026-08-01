package controlcontract

import "testing"

func TestHostOwnedWorkflowRuntimeExecutorGateBuildsRetryAndInvocation(t *testing.T) {
	readiness := workflowRuntimeExecutorTestReadiness(t, ObjectiveReplannerActionRetrySameStrategy, "strategy:workflow", "strategy:workflow")
	if readiness.Status != HostActionReady ||
		!readiness.ReadyForHostWorkflowRuntimeInvocation ||
		!readiness.ReadyForSameStrategyRetry ||
		readiness.ReadyForL3Fallback ||
		!readiness.HostRuntimeExecutorInvocationAuthorized ||
		!readiness.HostMayInvokeRuntimeExecutor {
		t.Fatalf("unexpected retry readiness: %#v", readiness)
	}
	if readiness.CoreWorkflowRetryApplied ||
		readiness.CoreRuntimeExecutorInvoked ||
		readiness.WorkflowDispatched ||
		readiness.RunnerDispatched ||
		readiness.RuntimeAdapterExecuted ||
		readiness.StoreMutationExecuted ||
		readiness.CompensationExecuted {
		t.Fatalf("readiness must not execute side effects: %#v", readiness)
	}
	if !workflowRuntimeExecutorBoundaryContains(readiness.Boundaries, "ready_for_host_same_strategy_workflow_retry") ||
		!workflowRuntimeExecutorBoundaryContains(readiness.Boundaries, "no_second_workflow_engine_in_controlplane") ||
		!workflowRuntimeExecutorBoundaryContains(readiness.Boundaries, "no_workflow_dispatch") {
		t.Fatalf("missing readiness boundaries: %#v", readiness.Boundaries)
	}

	invocation := BuildHostOwnedWorkflowRuntimeExecutorInvocation(HostOwnedWorkflowRuntimeExecutorInvocationInput{
		Readiness:                 readiness,
		InvocationReportRef:       "invocation_report:workflow_retry",
		ObservedInvocationRef:     readiness.InvocationRef,
		HostRuntimeExecutorRunRef: "runtime_executor_run:workflow_retry",
		WorkflowResultRef:         "workflow_result:retry",
		WorkflowReadbackRef:       "workflow_readback:retry",
		ObservedWorkflowRunRef:    readiness.WorkflowRunRef,
		ObservationRef:            "observation:workflow_retry_result",
		HostInvocationReported:    true,
		HostInvocationCompleted:   true,
	})
	if invocation.Status != HostActionRecorded ||
		!invocation.ReadyForWorkflowObservation ||
		invocation.ReadyForFailureReview ||
		!invocation.HostInvocationReported ||
		!invocation.HostInvocationCompleted ||
		invocation.HostInvocationFailed {
		t.Fatalf("unexpected invocation: %#v", invocation)
	}
	if invocation.CoreWorkflowRetryApplied ||
		invocation.CoreRuntimeExecutorInvoked ||
		invocation.WorkflowDispatched ||
		invocation.RunnerDispatched ||
		invocation.RuntimeAdapterExecuted ||
		invocation.StoreMutationExecuted ||
		invocation.CompensationExecuted {
		t.Fatalf("invocation must not execute side effects: %#v", invocation)
	}
	if !workflowRuntimeExecutorBoundaryContains(invocation.Boundaries, "host_owned_workflow_runtime_executor_invocation_recorded") ||
		!workflowRuntimeExecutorBoundaryContains(invocation.Boundaries, "ready_for_workflow_result_observation") {
		t.Fatalf("missing invocation boundaries: %#v", invocation.Boundaries)
	}
}

func TestHostOwnedWorkflowRuntimeExecutorGateBuildsL3Fallback(t *testing.T) {
	readiness := workflowRuntimeExecutorTestReadiness(t, ObjectiveReplannerActionSwitchStrategy, "strategy:workflow_primary", "strategy:workflow_fallback")
	if readiness.Status != HostActionReady ||
		!readiness.ReadyForHostWorkflowRuntimeInvocation ||
		readiness.ReadyForSameStrategyRetry ||
		!readiness.ReadyForL3Fallback ||
		readiness.CurrentStrategyRef != "strategy:workflow_primary" ||
		readiness.NextStrategyRef != "strategy:workflow_fallback" {
		t.Fatalf("unexpected fallback readiness: %#v", readiness)
	}
	if !workflowRuntimeExecutorBoundaryContains(readiness.Boundaries, "ready_for_host_l3_workflow_fallback") {
		t.Fatalf("missing fallback boundary: %#v", readiness.Boundaries)
	}
}

func TestHostOwnedWorkflowRuntimeExecutorGateBlocksMissingConfirmationAndMismatches(t *testing.T) {
	input := workflowRuntimeExecutorTestReadinessInput(ObjectiveReplannerActionRetrySameStrategy, "strategy:workflow", "strategy:workflow")
	input.HostConfirmationRef = ""
	blocked := BuildHostOwnedWorkflowRuntimeExecutorReadiness(input)
	if blocked.Status == HostActionReady ||
		blocked.ReadyForHostWorkflowRuntimeInvocation ||
		!workflowRuntimeExecutorStringContains(blocked.BlockedReasons, "host_confirmation_ref_missing") ||
		!workflowRuntimeExecutorMissingContains(blocked.MissingInputs, "host:workflow_runtime_executor_confirmation") {
		t.Fatalf("expected missing confirmation block, got %#v", blocked)
	}

	input = workflowRuntimeExecutorTestReadinessInput(ObjectiveReplannerActionRetrySameStrategy, "strategy:workflow", "strategy:workflow_fallback")
	mismatch := BuildHostOwnedWorkflowRuntimeExecutorReadiness(input)
	if mismatch.Status == HostActionReady ||
		!workflowRuntimeExecutorStringContains(mismatch.BlockedReasons, "same_strategy_retry_ref_mismatch") {
		t.Fatalf("expected same-strategy mismatch block, got %#v", mismatch)
	}

	input = workflowRuntimeExecutorTestReadinessInput(ObjectiveReplannerActionRetrySameStrategy, "strategy:workflow", "strategy:workflow")
	input.SourceProjection.SourceKind = ReplannerSourceOperations
	wrongSource := BuildHostOwnedWorkflowRuntimeExecutorReadiness(input)
	if wrongSource.Status == HostActionReady ||
		!workflowRuntimeExecutorStringContains(wrongSource.BlockedReasons, "workflow_source_projection_required") {
		t.Fatalf("expected workflow source block, got %#v", wrongSource)
	}
}

func TestHostOwnedWorkflowRuntimeExecutorInvocationFailureAndUnsafeRefs(t *testing.T) {
	readiness := workflowRuntimeExecutorTestReadiness(t, ObjectiveReplannerActionRetrySameStrategy, "strategy:workflow", "strategy:workflow")
	failure := BuildHostOwnedWorkflowRuntimeExecutorInvocation(HostOwnedWorkflowRuntimeExecutorInvocationInput{
		Readiness:                 readiness,
		InvocationReportRef:       "invocation_report:workflow_retry_failure",
		ObservedInvocationRef:     readiness.InvocationRef,
		HostRuntimeExecutorRunRef: "runtime_executor_run:workflow_retry_failure",
		ObservedWorkflowRunRef:    readiness.WorkflowRunRef,
		FailureRef:                "failure:workflow_retry_failure",
		CompensationRef:           "compensation:workflow_retry_failure",
		HostInvocationReported:    true,
		HostInvocationFailed:      true,
	})
	if failure.Status != HostActionRecorded ||
		failure.ReadyForWorkflowObservation ||
		!failure.ReadyForFailureReview ||
		!failure.HostInvocationFailed ||
		!workflowRuntimeExecutorBoundaryContains(failure.Boundaries, "workflow_runtime_executor_failure_reported") {
		t.Fatalf("unexpected failure invocation: %#v", failure)
	}

	mismatch := BuildHostOwnedWorkflowRuntimeExecutorInvocation(HostOwnedWorkflowRuntimeExecutorInvocationInput{
		Readiness:                 readiness,
		InvocationReportRef:       "invocation_report:workflow_retry",
		ObservedInvocationRef:     "invocation:other",
		HostRuntimeExecutorRunRef: "runtime_executor_run:workflow_retry",
		WorkflowResultRef:         "workflow_result:retry",
		WorkflowReadbackRef:       "workflow_readback:retry",
		ObservedWorkflowRunRef:    readiness.WorkflowRunRef,
		ObservationRef:            "observation:workflow_retry_result",
		HostInvocationReported:    true,
		HostInvocationCompleted:   true,
	})
	if mismatch.Status == HostActionRecorded ||
		!workflowRuntimeExecutorStringContains(mismatch.BlockedReasons, "observed_invocation_ref_mismatch") {
		t.Fatalf("expected invocation mismatch block, got %#v", mismatch)
	}

	input := workflowRuntimeExecutorTestReadinessInput(ObjectiveReplannerActionRetrySameStrategy, "strategy:workflow", "strategy:workflow")
	input.AdapterRef = "http://localhost/raw-workflow-run"
	unsafe := BuildHostOwnedWorkflowRuntimeExecutorReadiness(input)
	if unsafe.Status != HostActionReviewRequired ||
		unsafe.AdapterRef != "" ||
		!unsafe.RawOutputLoaded ||
		!workflowRuntimeExecutorMissingContains(unsafe.MissingInputs, "host:display_safe_refs") ||
		!workflowRuntimeExecutorBoundaryContains(unsafe.Boundaries, "raw_output_not_allowed") {
		t.Fatalf("expected unsafe ref review, got %#v", unsafe)
	}
}

func workflowRuntimeExecutorTestReadiness(t *testing.T, action ObjectiveReplannerAction, current, next DisplaySafeRef) HostOwnedWorkflowRuntimeExecutorReadiness {
	t.Helper()
	readiness := BuildHostOwnedWorkflowRuntimeExecutorReadiness(workflowRuntimeExecutorTestReadinessInput(action, current, next))
	if readiness.Status != HostActionReady {
		t.Fatalf("test readiness should be ready: %#v", readiness)
	}
	return readiness
}

func workflowRuntimeExecutorTestReadinessInput(action ObjectiveReplannerAction, current, next DisplaySafeRef) HostOwnedWorkflowRuntimeExecutorReadinessInput {
	return HostOwnedWorkflowRuntimeExecutorReadinessInput{
		SourceProjection: ReplannerSourceProjection{
			SourceKind:   ReplannerSourceWorkflow,
			SourceRef:    "workflow:failed_run",
			Producer:     "core:agentx_workflow",
			ControlMode:  ControlModeWorkflow,
			Status:       VerificationFailed,
			FailureClass: FailureVerificationFailed,
			Candidate: StrategyCandidate{
				ID:           "strategy:workflow",
				ControlMode:  ControlModeWorkflow,
				MinIntensity: IntensityL2BoundedToolLoop,
				MaxIntensity: IntensityL3ManagedObjective,
			},
			EvidenceRefs: []EvidenceRef{{
				Ref:      "evidence:workflow_failed",
				Kind:     "workflow_node_result",
				Strength: EvidenceWeak,
				Source:   "workflow:failed_run",
			}},
			Boundaries: []Boundary{"workflow_source_projection"},
		},
		Verification: ObjectiveVerificationGateResult{
			Status:       VerificationFailed,
			FailureClass: FailureVerificationFailed,
			Frame: ObjectiveFrame{
				ID:        "objective:workflow_retry",
				Intensity: IntensityL3ManagedObjective,
			},
			EvidenceRefs: []EvidenceRef{{
				Ref:      "evidence:workflow_failed",
				Kind:     "workflow_node_result",
				Strength: EvidenceWeak,
			}},
			Verification: VerificationResult{
				Status:       VerificationFailed,
				FailureClass: FailureVerificationFailed,
				EvidenceRefs: []EvidenceRef{{
					Ref:      "evidence:workflow_failed",
					Kind:     "workflow_node_result",
					Strength: EvidenceWeak,
				}},
			},
		},
		ReplannerDecision: ObjectiveReplannerDecision{
			Status:             VerificationPartial,
			Action:             action,
			CurrentStrategyRef: current,
			NextStrategyRef:    next,
			FailureClass:       FailureVerificationFailed,
			EvidenceRefs: []EvidenceRef{{
				Ref:      "evidence:workflow_failed",
				Kind:     "workflow_node_result",
				Strength: EvidenceWeak,
			}},
			Boundaries: []Boundary{"objective_replanner_decision"},
		},
		WorkflowRetryGate: BuildProductionAdapterIndependentEffectGate(ProductionAdapterIndependentEffectGateSpec{
			Kind:                  ProductionAdapterEffectGateWorkflowRetryApply,
			GateRef:               "gate:workflow_retry",
			AdapterRef:            "adapter:workflow_retry",
			ContractRef:           "contract:workflow_retry",
			PolicyRef:             "policy:workflow_retry",
			ApprovalRef:           "approval:workflow_retry",
			BudgetRef:             "budget:workflow_retry",
			IdempotencyRef:        "idempotency:workflow_retry",
			ReadbackRef:           "readback:workflow_retry",
			EvalRef:               "eval:workflow_retry",
			FailureReviewRef:      "failure_review:workflow_retry",
			CompensationReviewRef: "compensation_review:workflow_retry",
			EvidenceRefs:          []DisplaySafeRef{"evidence:workflow_failed"},
		}),
		RuntimeExecutorGate: BuildProductionAdapterIndependentEffectGate(ProductionAdapterIndependentEffectGateSpec{
			Kind:                  ProductionAdapterEffectGateRuntimeExecutor,
			GateRef:               "gate:runtime_executor",
			AdapterRef:            "adapter:runtime_executor",
			ContractRef:           "contract:runtime_executor",
			PolicyRef:             "policy:runtime_executor",
			ApprovalRef:           "approval:runtime_executor",
			BudgetRef:             "budget:runtime_executor",
			IdempotencyRef:        "idempotency:runtime_executor",
			ReadbackRef:           "readback:runtime_executor",
			EvalRef:               "eval:runtime_executor",
			FailureReviewRef:      "failure_review:runtime_executor",
			CompensationReviewRef: "compensation_review:runtime_executor",
			EvidenceRefs:          []DisplaySafeRef{"evidence:workflow_failed"},
		}),
		AdapterRef:           "adapter:workflow_runtime_executor",
		AdapterVersionRef:    "adapter_version:workflow_runtime_executor_v1",
		AdapterCapabilityRef: "capability:workflow_runtime_executor",
		AdapterContractRef:   "contract:workflow_runtime_executor",
		HostConfirmationRef:  "confirmation:workflow_runtime_executor",
		WorkflowRunRef:       "workflow_run:failed_run",
		RetryRequestRef:      "workflow_retry_request:failed_run",
		InvocationRef:        "invocation:workflow_runtime_executor",
		ResultBindingRef:     "workflow_result:retry",
		ReadbackBindingRef:   "workflow_readback:retry",
		IdempotencyRef:       "idempotency:workflow_runtime_executor",
		BudgetRef:            "budget:workflow_runtime_executor",
		RetryPolicyRef:       "policy:workflow_retry",
		FallbackPolicyRef:    "policy:workflow_fallback",
		FailureBindingRef:    "failure:workflow_runtime_executor",
		CompensationRef:      "compensation:workflow_runtime_executor",
	}
}

func workflowRuntimeExecutorBoundaryContains(values []Boundary, needle Boundary) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

func workflowRuntimeExecutorStringContains(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

func workflowRuntimeExecutorMissingContains(values []MissingInput, needle MissingInput) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}
