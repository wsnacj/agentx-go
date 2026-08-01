package controlcontract

import "testing"

func TestHostOwnedDelegationWorkerRuntimeGateBuildsL4ReadinessInvocationAndReview(t *testing.T) {
	readiness := delegationWorkerRuntimeTestReadiness(t, IntensityL4DurableLongRun)
	if readiness.Status != HostActionReady ||
		!readiness.ReadyForHostWorkerRuntimeInvocation ||
		!readiness.ReadyForL4HostApprovedWorker ||
		readiness.ReadyForL5ExplicitWorker ||
		!readiness.HostWorkerRuntimeInvocationAuthorized ||
		!readiness.HostMayInvokeWorkerRuntime ||
		readiness.WorkerOutputAcceptedAsFact ||
		!readiness.WorkerResultRequiresVerification {
		t.Fatalf("unexpected L4 readiness: %#v", readiness)
	}
	if readiness.CoreWorkerRuntimeInvoked ||
		readiness.WorkerDispatched ||
		readiness.RunnerDispatched ||
		readiness.RuntimeAdapterExecuted ||
		readiness.WorkflowDispatched ||
		readiness.SchedulerApplied ||
		readiness.InstallerExecuted ||
		readiness.StoreMutationExecuted ||
		readiness.CompensationExecuted {
		t.Fatalf("readiness must not execute side effects: %#v", readiness)
	}
	if !delegationWorkerRuntimeBoundaryContains(readiness.Boundaries, "ready_for_l4_host_approved_delegation_worker") ||
		!delegationWorkerRuntimeBoundaryContains(readiness.Boundaries, "no_delegation_worker_runtime_by_core") ||
		!delegationWorkerRuntimeBoundaryContains(readiness.Boundaries, "worker_result_requires_verification") {
		t.Fatalf("missing readiness boundaries: %#v", readiness.Boundaries)
	}

	invocation := BuildHostOwnedDelegationWorkerRuntimeInvocation(HostOwnedDelegationWorkerRuntimeInvocationInput{
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
	})
	if invocation.Status != HostActionRecorded ||
		!invocation.ReadyForWorkerResultReview ||
		invocation.ReadyForFailureReview ||
		!invocation.HostInvocationReported ||
		!invocation.HostInvocationCompleted ||
		invocation.HostInvocationFailed ||
		invocation.ResultReview.Status != VerificationBlocked ||
		invocation.ResultReview.ReadyForParentMerge ||
		!delegationWorkerRuntimeMissingContains(invocation.ResultReview.MissingInputs, "host:parent_verification") ||
		invocation.WorkerOutputAcceptedAsFact ||
		!invocation.WorkerResultRequiresVerification {
		t.Fatalf("unexpected invocation: %#v", invocation)
	}
	if invocation.CoreWorkerRuntimeInvoked ||
		invocation.WorkerDispatched ||
		invocation.RunnerDispatched ||
		invocation.RuntimeAdapterExecuted ||
		invocation.WorkflowDispatched ||
		invocation.SchedulerApplied ||
		invocation.InstallerExecuted ||
		invocation.StoreMutationExecuted ||
		invocation.CompensationExecuted {
		t.Fatalf("invocation must not execute side effects: %#v", invocation)
	}
	if !delegationWorkerRuntimeBoundaryContains(invocation.Boundaries, "host_owned_delegation_worker_runtime_invocation_recorded") ||
		!delegationWorkerRuntimeBoundaryContains(invocation.Boundaries, "ready_for_delegation_worker_result_review") ||
		!delegationWorkerRuntimeBoundaryContains(invocation.Boundaries, "worker_result_requires_parent_verification") {
		t.Fatalf("missing invocation boundaries: %#v", invocation.Boundaries)
	}
}

func TestHostOwnedDelegationWorkerRuntimeGateAllowsExplicitL5(t *testing.T) {
	readiness := delegationWorkerRuntimeTestReadiness(t, IntensityL5Autonomous)
	if readiness.Status != HostActionReady ||
		!readiness.ReadyForHostWorkerRuntimeInvocation ||
		readiness.ReadyForL4HostApprovedWorker ||
		!readiness.ReadyForL5ExplicitWorker {
		t.Fatalf("unexpected L5 readiness: %#v", readiness)
	}
	if !delegationWorkerRuntimeBoundaryContains(readiness.Boundaries, "ready_for_l5_explicit_delegation_worker") {
		t.Fatalf("missing explicit L5 boundary: %#v", readiness.Boundaries)
	}
}

func TestHostOwnedDelegationWorkerRuntimeGateBlocksL5DefaultOffAndMissingConfirmation(t *testing.T) {
	input := delegationWorkerRuntimeTestReadinessInput(IntensityL5Autonomous)
	input.Request.L5Enabled = false
	defaultOff := BuildHostOwnedDelegationWorkerRuntimeReadiness(input)
	if defaultOff.Status == HostActionReady ||
		defaultOff.ReadyForHostWorkerRuntimeInvocation ||
		!delegationWorkerRuntimeStringContains(defaultOff.BlockedReasons, "l5_delegation_not_explicitly_enabled") ||
		!delegationWorkerRuntimeMissingContains(defaultOff.MissingInputs, "host:l5_delegation_policy") {
		t.Fatalf("expected L5 default-off block, got %#v", defaultOff)
	}

	input = delegationWorkerRuntimeTestReadinessInput(IntensityL4DurableLongRun)
	input.HostConfirmationRef = ""
	missingConfirmation := BuildHostOwnedDelegationWorkerRuntimeReadiness(input)
	if missingConfirmation.Status == HostActionReady ||
		missingConfirmation.ReadyForHostWorkerRuntimeInvocation ||
		!delegationWorkerRuntimeStringContains(missingConfirmation.BlockedReasons, "host_confirmation_ref_missing") ||
		!delegationWorkerRuntimeMissingContains(missingConfirmation.MissingInputs, "host:delegation_worker_runtime_confirmation") {
		t.Fatalf("expected missing confirmation block, got %#v", missingConfirmation)
	}
}

func TestHostOwnedDelegationWorkerRuntimeInvocationFailureAndUnsafeRefs(t *testing.T) {
	readiness := delegationWorkerRuntimeTestReadiness(t, IntensityL4DurableLongRun)
	failure := BuildHostOwnedDelegationWorkerRuntimeInvocation(HostOwnedDelegationWorkerRuntimeInvocationInput{
		Readiness:               readiness,
		InvocationReportRef:     "invocation_report:delegation_worker_runtime_failure",
		ObservedInvocationRef:   readiness.InvocationRef,
		HostWorkerRuntimeRunRef: "worker_runtime_run:delegation_failure",
		ObservedWorkerRunRef:    readiness.WorkerRunRef,
		FailureRef:              "failure:delegation_worker_runtime",
		CompensationRef:         "compensation:delegation_worker_runtime",
		HostInvocationReported:  true,
		HostInvocationFailed:    true,
	})
	if failure.Status != HostActionRecorded ||
		failure.ReadyForWorkerResultReview ||
		!failure.ReadyForFailureReview ||
		!failure.HostInvocationFailed ||
		!delegationWorkerRuntimeBoundaryContains(failure.Boundaries, "delegation_worker_runtime_failure_reported") {
		t.Fatalf("unexpected failure invocation: %#v", failure)
	}

	mismatch := BuildHostOwnedDelegationWorkerRuntimeInvocation(HostOwnedDelegationWorkerRuntimeInvocationInput{
		Readiness:               readiness,
		InvocationReportRef:     "invocation_report:delegation_worker_runtime",
		ObservedInvocationRef:   "invocation:other",
		HostWorkerRuntimeRunRef: "worker_runtime_run:delegation_fixture",
		ObservedWorkerRunRef:    readiness.WorkerRunRef,
		WorkerResultRef:         "worker_result:delegation_fixture",
		WorkerReadbackRef:       "worker_readback:delegation_fixture",
		ObservationRef:          "observation:delegation_worker_result",
		HostInvocationReported:  true,
		HostInvocationCompleted: true,
	})
	if mismatch.Status == HostActionRecorded ||
		!delegationWorkerRuntimeStringContains(mismatch.BlockedReasons, "observed_invocation_ref_mismatch") {
		t.Fatalf("expected invocation mismatch block, got %#v", mismatch)
	}

	input := delegationWorkerRuntimeTestReadinessInput(IntensityL4DurableLongRun)
	input.AdapterRef = "http://localhost/raw-worker"
	unsafe := BuildHostOwnedDelegationWorkerRuntimeReadiness(input)
	if unsafe.Status != HostActionReviewRequired ||
		unsafe.ReadyForHostWorkerRuntimeInvocation ||
		unsafe.AdapterRef != "" ||
		!unsafe.RawOutputLoaded ||
		!delegationWorkerRuntimeMissingContains(unsafe.MissingInputs, "host:display_safe_refs") ||
		!delegationWorkerRuntimeBoundaryContains(unsafe.Boundaries, "raw_output_not_allowed") {
		t.Fatalf("expected unsafe ref review, got %#v", unsafe)
	}
}

func delegationWorkerRuntimeTestReadiness(t *testing.T, intensity ExecutionIntensity) HostOwnedDelegationWorkerRuntimeReadiness {
	t.Helper()
	readiness := BuildHostOwnedDelegationWorkerRuntimeReadiness(delegationWorkerRuntimeTestReadinessInput(intensity))
	if readiness.Status != HostActionReady {
		t.Fatalf("test readiness should be ready: %#v", readiness)
	}
	return readiness
}

func delegationWorkerRuntimeTestReadinessInput(intensity ExecutionIntensity) HostOwnedDelegationWorkerRuntimeReadinessInput {
	requestInput := delegationReadyRequestInput(intensity)
	if intensity == IntensityL5Autonomous {
		requestInput.L5Enabled = true
	}
	return HostOwnedDelegationWorkerRuntimeReadinessInput{
		Request: BuildDelegationRequestProjection(requestInput),
		WorkerRuntimeGate: BuildProductionAdapterIndependentEffectGate(ProductionAdapterIndependentEffectGateSpec{
			Kind:                  ProductionAdapterEffectGateDelegationWorker,
			GateRef:               "gate:delegation_worker_runtime",
			AdapterRef:            "adapter:delegation_worker_runtime",
			ContractRef:           "contract:delegation_worker_runtime",
			PolicyRef:             "policy:delegation_worker_runtime",
			ApprovalRef:           "approval:delegation_worker_runtime",
			BudgetRef:             "budget:delegation_worker_runtime",
			IdempotencyRef:        "idempotency:delegation_worker_runtime",
			ReadbackRef:           "readback:delegation_worker_runtime",
			EvalRef:               "eval:delegation_worker_runtime",
			FailureReviewRef:      "review:delegation_worker_runtime_failure",
			CompensationReviewRef: "review:delegation_worker_runtime_compensation",
			EvidenceRefs:          []DisplaySafeRef{"evidence:delegation_worker_verified"},
		}),
		AdapterRef:           "adapter:delegation_worker_runtime",
		AdapterVersionRef:    "adapter_version:delegation_worker_runtime_v1",
		AdapterCapabilityRef: "capability:delegation_worker_runtime",
		AdapterContractRef:   "contract:delegation_worker_runtime",
		HostConfirmationRef:  "confirmation:delegation_worker_runtime",
		WorkerRunRef:         "worker_run:delegation_fixture",
		WorkerRequestRef:     "worker_request:delegation_fixture",
		InvocationRef:        "invocation:delegation_worker_runtime",
		ResultBindingRef:     "worker_result:delegation_fixture",
		ReadbackBindingRef:   "worker_readback:delegation_fixture",
		IdempotencyRef:       "idempotency:delegation_worker_runtime",
		BudgetRef:            "budget:delegation_worker_runtime",
		VerificationRef:      "verification:delegation_worker_parent",
		FailureBindingRef:    "failure:delegation_worker_runtime",
		CompensationRef:      "compensation:delegation_worker_runtime",
	}
}

func delegationWorkerRuntimeBoundaryContains(values []Boundary, needle Boundary) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

func delegationWorkerRuntimeStringContains(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

func delegationWorkerRuntimeMissingContains(values []MissingInput, needle MissingInput) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}
