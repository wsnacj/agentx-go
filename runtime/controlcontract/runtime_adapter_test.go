package controlcontract

import "testing"

func TestHostAdapterRegistryFromProductionDescriptorReady(t *testing.T) {
	registry := runtimeAdapterTestRegistry(runtimeAdapterTestDescriptor())

	if registry.Status != HostActionReady ||
		!registry.ReadyForRuntimeRequest ||
		registry.NextHostAction != "host_may_build_runtime_adapter_request" ||
		registry.RunnerEffect != "none" ||
		registry.PromptEffect != "none" {
		t.Fatalf("registry = %#v", registry)
	}
	if !displaySafeRefSliceContains(registry.AdapterRefs, "adapter:metrics_readonly") ||
		!displaySafeRefSliceContains(registry.StrategyRefs, "strategy:host_metric") {
		t.Fatalf("registry refs = %#v / %#v", registry.AdapterRefs, registry.StrategyRefs)
	}
	if !intensityGateBoundaryContains(registry.Boundaries, "host_adapter_registry") ||
		!intensityGateBoundaryContains(registry.Boundaries, "objective_loop_runtime_adapter_registry") ||
		!intensityGateBoundaryContains(registry.Boundaries, "no_adapter_invocation") {
		t.Fatalf("registry boundaries = %#v", registry.Boundaries)
	}
}

func TestRuntimeAdapterExecutionRequestReadyOnlyAfterFinalGate(t *testing.T) {
	request := runtimeAdapterTestRequest(t, runtimeAdapterTestRequestOptions{})

	if request.Status != HostActionReady ||
		!request.ReadyForHostExecution ||
		request.NextHostAction != "host_may_execute_runtime_adapter" ||
		request.AdapterRef != "adapter:metrics_readonly" ||
		request.StrategyRef != "strategy:host_metric" ||
		request.RunnerEffect != "none" ||
		request.PromptEffect != "none" {
		t.Fatalf("request = %#v", request)
	}
	if !displaySafeRefSliceContains(request.RequiredCapabilityRefs, "capability:metric_reader") ||
		!displaySafeRefSliceContains(request.RequiredPolicyRefs, "policy:runtime_adapter") ||
		!displaySafeRefSliceContains(request.RequiredApprovalRefs, "approval:runtime_adapter") ||
		request.RequiredBudgetRef != "budget:objective" ||
		request.IdempotencyContractRef != "contract:idempotency" {
		t.Fatalf("request requirements = %#v", request)
	}
	if !intensityGateBoundaryContains(request.Boundaries, "final_gate_bound_adapter_request") ||
		!intensityGateBoundaryContains(request.Boundaries, "core_does_not_execute_adapter") ||
		!intensityGateBoundaryContains(request.Boundaries, "ready_for_host_runtime_adapter_execution") {
		t.Fatalf("request boundaries = %#v", request.Boundaries)
	}
}

func TestRuntimeAdapterExecutionRequestBlocksFinalGateAndUnavailableAdapter(t *testing.T) {
	blockedFinalGate := runtimeAdapterTestFinalGate(t, runtimeAdapterTestStrategy())
	blockedFinalGate.Allowed = false
	blockedFinalGate.Status = VerificationBlocked
	blockedFinalGate.FailureClass = FailureApprovalRequired

	blocked := BuildRuntimeAdapterExecutionRequest(RuntimeAdapterExecutionRequestInput{
		Activation:              ActivationManaged,
		Frame:                   runtimeAdapterTestFrame(),
		Selected:                runtimeAdapterTestSelected(),
		FinalGate:               blockedFinalGate,
		Registry:                runtimeAdapterTestRegistry(runtimeAdapterTestDescriptor()),
		RequestedAdapterRef:     "adapter:metrics_readonly",
		Budget:                  runtimeAdapterTestBudget(),
		ApprovalRefs:            []DisplaySafeRef{"approval:runtime_adapter"},
		PolicyRefs:              []DisplaySafeRef{"policy:runtime_adapter"},
		AvailableCapabilityRefs: []DisplaySafeRef{"capability:metric_reader"},
		IdempotencyRef:          "idempotency:metrics_1",
		InputRefs:               []DisplaySafeRef{"input:metrics_scope"},
	})
	if blocked.ReadyForHostExecution ||
		blocked.FailureClass != FailureApprovalRequired ||
		blocked.NextHostAction != "run_strategy_final_gate" ||
		!intensityGateBoundaryContains(blocked.Boundaries, "runtime_adapter_requires_final_gate") {
		t.Fatalf("blocked final gate request = %#v", blocked)
	}

	missing := runtimeAdapterTestRequest(t, runtimeAdapterTestRequestOptions{RequestedAdapterRef: "adapter:missing"})
	if missing.ReadyForHostExecution ||
		missing.FailureClass != FailureHostAdapterMissing ||
		missing.NextHostAction != "provide_runtime_adapter" ||
		!intensityGateBoundaryContains(missing.Boundaries, "runtime_adapter_missing") {
		t.Fatalf("missing adapter request = %#v", missing)
	}
}

func TestRuntimeAdapterExecutionRequestBlocksPolicyCapabilityAndNonReadOnly(t *testing.T) {
	missingCapability := runtimeAdapterTestRequest(t, runtimeAdapterTestRequestOptions{AvailableCapabilityRefs: []DisplaySafeRef{}})
	if missingCapability.ReadyForHostExecution ||
		missingCapability.FailureClass != FailureCapabilityMissing ||
		missingCapability.NextHostAction != "enter_capability_resolution" ||
		!intensityGateBoundaryContains(missingCapability.Boundaries, "capability_gap_proposal_only") {
		t.Fatalf("missing capability request = %#v", missingCapability)
	}

	missingPolicy := runtimeAdapterTestRequest(t, runtimeAdapterTestRequestOptions{PolicyRefs: []DisplaySafeRef{}})
	if missingPolicy.ReadyForHostExecution ||
		missingPolicy.FailureClass != FailurePolicyBlocked ||
		missingPolicy.NextHostAction != "provide_adapter_policy" ||
		!intensityGateBoundaryContains(missingPolicy.Boundaries, "runtime_adapter_policy_missing") {
		t.Fatalf("missing policy request = %#v", missingPolicy)
	}

	nonReadOnlyDescriptor := runtimeAdapterTestDescriptor()
	nonReadOnlyDescriptor.Kind = ProductionAdapterWorkflowDispatch
	nonReadOnlyDescriptor.SideEffectClass = "workflow_dispatch"
	nonReadOnly := runtimeAdapterTestRequest(t, runtimeAdapterTestRequestOptions{Descriptor: nonReadOnlyDescriptor})
	if nonReadOnly.ReadyForHostExecution ||
		nonReadOnly.FailureClass != FailurePolicyBlocked ||
		nonReadOnly.NextHostAction != "request_host_approval" ||
		!intensityGateBoundaryContains(nonReadOnly.Boundaries, "runtime_adapter_non_read_only_not_enabled") {
		t.Fatalf("non-read-only request = %#v", nonReadOnly)
	}

	allowedWithoutSideEffectApproval := runtimeAdapterTestRequest(t, runtimeAdapterTestRequestOptions{
		Descriptor:                 nonReadOnlyDescriptor,
		AllowHostSideEffectAdapter: true,
	})
	if allowedWithoutSideEffectApproval.ReadyForHostExecution ||
		allowedWithoutSideEffectApproval.FailureClass != FailureApprovalRequired ||
		allowedWithoutSideEffectApproval.NextHostAction != "request_host_approval" ||
		!intensityGateBoundaryContains(allowedWithoutSideEffectApproval.Boundaries, "runtime_adapter_side_effect_approval_ref_missing") {
		t.Fatalf("expected side-effect approval ref block, got %#v", allowedWithoutSideEffectApproval)
	}
}

func TestRuntimeAdapterExecutionRequestAllowsExplicitHostSideEffectAdapter(t *testing.T) {
	descriptor := runtimeAdapterTestDescriptor()
	descriptor.Kind = ProductionAdapterWorkflowDispatch
	descriptor.SideEffectClass = "workflow_dispatch"
	descriptor.RequiredApprovalRefs = []DisplaySafeRef{"approval:runtime_adapter", "approval:workflow_dispatch"}
	request := runtimeAdapterTestRequest(t, runtimeAdapterTestRequestOptions{
		Descriptor:                 descriptor,
		ApprovalRefs:               descriptor.RequiredApprovalRefs,
		AllowHostSideEffectAdapter: true,
		HostSideEffectApprovalRefs: []DisplaySafeRef{"approval:workflow_dispatch"},
	})

	if !request.ReadyForHostExecution ||
		request.FailureClass != FailureNone ||
		!request.HostSideEffectAdapterAllowed ||
		!displaySafeRefSliceContains(request.HostSideEffectApprovalRefs, "approval:workflow_dispatch") ||
		!intensityGateBoundaryContains(request.Boundaries, "host_side_effect_adapter_explicitly_allowed") ||
		!intensityGateBoundaryContains(request.Boundaries, "ready_for_host_runtime_adapter_execution") {
		t.Fatalf("explicit side-effect request = %#v", request)
	}
}

func TestRuntimeAdapterExecutionResultBindsStructuredObservations(t *testing.T) {
	request := runtimeAdapterTestRequest(t, runtimeAdapterTestRequestOptions{})
	result := runtimeAdapterTestResult(request, []Observation{runtimeAdapterTestObservation(EvidenceStrong)})

	if result.Status != VerificationSatisfied ||
		!result.Satisfied ||
		!result.HostExecutionReported ||
		!result.ReadyForObservationNormalization ||
		result.NextHostAction != "normalize_adapter_observations" ||
		result.RunnerEffect != "none" ||
		result.PromptEffect != "none" {
		t.Fatalf("result = %#v", result)
	}
	if len(result.Observations) != 1 ||
		result.Observations[0].Kind != "metric" ||
		result.Observations[0].Name != "cpu_usage" ||
		len(result.EvidenceRefs) == 0 ||
		!displaySafeRefSliceContains(result.OutputRefs, "output:metrics_summary") {
		t.Fatalf("result observations/evidence = %#v", result)
	}
	if !intensityGateBoundaryContains(result.Boundaries, "adapter_result_not_objective_completion") ||
		!intensityGateBoundaryContains(result.Boundaries, "ready_for_observation_normalization") {
		t.Fatalf("result boundaries = %#v", result.Boundaries)
	}
}

func TestRuntimeAdapterExecutionResultBlocksMismatchUnsafeWeakAndCapabilityGap(t *testing.T) {
	request := runtimeAdapterTestRequest(t, runtimeAdapterTestRequestOptions{})

	mismatch := runtimeAdapterTestResultInput(request, []Observation{runtimeAdapterTestObservation(EvidenceStrong)})
	mismatch.AdapterRef = "adapter:other"
	mismatchResult := BuildRuntimeAdapterExecutionResult(mismatch)
	if mismatchResult.Satisfied ||
		mismatchResult.FailureClass != FailureVerificationFailed ||
		!intensityGateBoundaryContains(mismatchResult.Boundaries, "runtime_adapter_result_adapter_mismatch") {
		t.Fatalf("mismatch result = %#v", mismatchResult)
	}

	unsafe := runtimeAdapterTestResultInput(request, []Observation{runtimeAdapterTestObservation(EvidenceStrong)})
	unsafe.RawOutputLoaded = true
	unsafeResult := BuildRuntimeAdapterExecutionResult(unsafe)
	if unsafeResult.Satisfied ||
		unsafeResult.Status != VerificationReviewRequired ||
		unsafeResult.FailureClass != FailureEvidenceWeak ||
		!intensityGateBoundaryContains(unsafeResult.Boundaries, "raw_output_not_allowed") {
		t.Fatalf("unsafe result = %#v", unsafeResult)
	}

	weak := runtimeAdapterTestResult(request, []Observation{runtimeAdapterTestObservation(EvidenceWeak)})
	if weak.Satisfied ||
		weak.Status != VerificationReviewRequired ||
		weak.FailureClass != FailureEvidenceWeak ||
		weak.NextHostAction != "review_runtime_adapter_evidence" ||
		!intensityGateBoundaryContains(weak.Boundaries, "runtime_adapter_observation_evidence_weak") {
		t.Fatalf("weak result = %#v", weak)
	}

	gapInput := runtimeAdapterTestResultInput(request, []Observation{runtimeAdapterTestObservation(EvidenceStrong)})
	gapInput.MissingCapabilityRefs = []DisplaySafeRef{"capability:docker_socket"}
	gap := BuildRuntimeAdapterExecutionResult(gapInput)
	if gap.Satisfied ||
		gap.Status != VerificationBlocked ||
		gap.FailureClass != FailureCapabilityMissing ||
		gap.NextHostAction != "enter_capability_resolution" ||
		!intensityGateBoundaryContains(gap.Boundaries, "runtime_adapter_reported_capability_gap") {
		t.Fatalf("capability gap result = %#v", gap)
	}
}

func TestRuntimeAdapterExecutionResultPreservesBlockedNextHostAction(t *testing.T) {
	request := runtimeAdapterTestRequest(t, runtimeAdapterTestRequestOptions{})
	input := runtimeAdapterTestResultInput(request, []Observation{runtimeAdapterTestObservation(EvidenceStrong)})
	input.Status = VerificationBlocked
	input.FailureClass = FailureCredentialMissing
	input.FailureReason = "credential invalid"
	input.NextHostAction = "refresh_runtime_adapter_credential"
	input.MissingInputs = []MissingInput{"host:runtime_adapter_credential"}
	input.Boundaries = []Boundary{"runtime_adapter_credential_refresh_required"}

	blocked := BuildRuntimeAdapterExecutionResult(input)
	if blocked.Satisfied ||
		blocked.Status != VerificationBlocked ||
		blocked.FailureClass != FailureCredentialMissing ||
		blocked.NextHostAction != "refresh_runtime_adapter_credential" ||
		!intensityGateBoundaryContains(blocked.Boundaries, "runtime_adapter_execution_not_satisfied") ||
		!intensityGateBoundaryContains(blocked.Boundaries, "runtime_adapter_credential_refresh_required") {
		t.Fatalf("blocked result = %#v", blocked)
	}
}

type runtimeAdapterTestRequestOptions struct {
	Descriptor                 ProductionAdapterDescriptor
	RequestedAdapterRef        DisplaySafeRef
	AvailableCapabilityRefs    []DisplaySafeRef
	PolicyRefs                 []DisplaySafeRef
	ApprovalRefs               []DisplaySafeRef
	AllowHostSideEffectAdapter bool
	HostSideEffectApprovalRefs []DisplaySafeRef
}

func runtimeAdapterTestRequest(t *testing.T, opts runtimeAdapterTestRequestOptions) RuntimeAdapterExecutionRequest {
	t.Helper()
	descriptor := opts.Descriptor
	if descriptor.AdapterRef == "" {
		descriptor = runtimeAdapterTestDescriptor()
	}
	requested := opts.RequestedAdapterRef
	if requested == "" {
		requested = "adapter:metrics_readonly"
	}
	capabilityRefs := opts.AvailableCapabilityRefs
	if capabilityRefs == nil {
		capabilityRefs = []DisplaySafeRef{"capability:metric_reader"}
	}
	policyRefs := opts.PolicyRefs
	if policyRefs == nil {
		policyRefs = []DisplaySafeRef{"policy:runtime_adapter"}
	}
	approvalRefs := opts.ApprovalRefs
	if approvalRefs == nil {
		approvalRefs = []DisplaySafeRef{"approval:runtime_adapter"}
	}
	strategy := runtimeAdapterTestStrategy()
	return BuildRuntimeAdapterExecutionRequest(RuntimeAdapterExecutionRequestInput{
		Activation:                 ActivationManaged,
		Frame:                      runtimeAdapterTestFrame(),
		Selected:                   runtimeAdapterTestSelected(),
		FinalGate:                  runtimeAdapterTestFinalGate(t, strategy),
		Registry:                   runtimeAdapterTestRegistry(descriptor),
		RequestedAdapterRef:        requested,
		Budget:                     runtimeAdapterTestBudget(),
		ApprovalRefs:               approvalRefs,
		AllowHostSideEffectAdapter: opts.AllowHostSideEffectAdapter,
		HostSideEffectApprovalRefs: opts.HostSideEffectApprovalRefs,
		PolicyRefs:                 policyRefs,
		AvailableCapabilityRefs:    capabilityRefs,
		IdempotencyRef:             "idempotency:metrics_1",
		InputRefs:                  []DisplaySafeRef{"input:metrics_scope"},
		ExpectedObservationKinds:   []string{"metric"},
	})
}

func runtimeAdapterTestRegistry(descriptor ProductionAdapterDescriptor) HostAdapterRegistrySnapshot {
	return BuildHostAdapterRegistry(HostAdapterRegistryInput{
		RegistryRef: "registry:runtime",
		Descriptors: []ProductionAdapterDescriptor{
			descriptor,
		},
		PolicyRefs: []DisplaySafeRef{"policy:intensity", "contract:execution"},
	})
}

func runtimeAdapterTestDescriptor() ProductionAdapterDescriptor {
	return ProductionAdapterDescriptor{
		AdapterRef:             "adapter:metrics_readonly",
		Owner:                  "host",
		OwnerRef:               "host:runtime",
		Version:                "v1",
		Kind:                   ProductionAdapterOperationsMetricCollect,
		SupportedSourceKinds:   []ReplannerSourceKind{ReplannerSourceOperations},
		SupportedCandidateRefs: []DisplaySafeRef{"strategy:host_metric"},
		ProvidesCapabilityRefs: []DisplaySafeRef{"capability:metric_reader"},
		RequiresCapabilityRefs: []DisplaySafeRef{"capability:metric_reader"},
		InputContractRef:       "contract:metrics_input",
		OutputContractRef:      "contract:metrics_output",
		ReadbackContractRef:    "contract:metrics_readback",
		RequiredPolicyRefs:     []DisplaySafeRef{"policy:runtime_adapter"},
		RequiredApprovalRefs:   []DisplaySafeRef{"approval:runtime_adapter"},
		RequiredBudgetRef:      "budget:objective",
		IdempotencyContractRef: "contract:idempotency",
		RiskRef:                "risk:read_only",
		SideEffectClass:        "tool_read_only",
		TimeoutPolicyRef:       "policy:timeout",
		CompensationHandoffRef: "handoff:compensation",
		RedactionPolicyRef:     "policy:redaction",
		PreflightCheckRefs:     []DisplaySafeRef{"preflight:metrics"},
		DisplaySafeInputRefs:   []DisplaySafeRef{"input:metrics_scope"},
		DisplaySafeOutputRefs:  []DisplaySafeRef{"output:metrics_summary"},
	}
}

func runtimeAdapterTestSelected() StrategyPlanCandidate {
	return StrategyPlanCandidate{
		Rank:       1,
		Score:      10,
		SourceKind: StrategyCatalogSourceHostAdapter,
		SourceRef:  "source:host_metric",
		Candidate:  runtimeAdapterTestStrategy(),
		Status:     VerificationSatisfied,
	}
}

func runtimeAdapterTestStrategy() StrategyCandidate {
	return StrategyCandidate{
		ID:              "strategy:host_metric",
		ControlMode:     ControlModeOperations,
		MinIntensity:    IntensityL3ManagedObjective,
		MaxIntensity:    IntensityL3ManagedObjective,
		CapabilityRefs:  []DisplaySafeRef{"capability:metric_reader"},
		Risk:            "read_only",
		SideEffectClass: "tool_read_only",
		Owner:           "host",
		ExpectedEvidence: []EvidenceRef{{
			Ref:      "evidence:metric",
			Kind:     "metric",
			Strength: EvidenceStrong,
			Source:   "source:host_metric",
		}},
	}
}

func runtimeAdapterTestFrame() ObjectiveFrame {
	return ObjectiveFrame{
		ID:             "objective:metrics",
		UserGoalDigest: "collect host metrics through a selected host adapter",
		ControlMode:    ControlModeOperations,
		Intensity:      IntensityL3ManagedObjective,
		SuccessCriteria: []string{
			"collect CPU and memory metric evidence",
		},
		RequiredEvidence: []EvidenceRef{{
			Ref:      "evidence:metric",
			Kind:     "metric",
			Strength: EvidenceStrong,
			Source:   "adapter:metrics_readonly",
		}},
	}
}

func runtimeAdapterTestBudget() ObjectiveBudgetSnapshot {
	return ObjectiveBudgetSnapshot{
		BudgetRef: "budget:objective",
		Limit:     3,
		Remaining: 3,
	}
}

func runtimeAdapterTestFinalGate(t *testing.T, strategy StrategyCandidate) IntensityGateResult {
	t.Helper()
	policy := objectiveLoopIntensityPolicy()
	pre := strategyPlannerPreGate(t, policy, ControlModeOperations, IntensityL3ManagedObjective)
	final := BuildExecutionIntensityFinalGate(IntensityGateInput{
		Activation:           ActivationManaged,
		Policy:               policy,
		RequestedControlMode: ControlModeOperations,
		RequestedIntensity:   IntensityL3ManagedObjective,
		UserConfirmed:        true,
		Budget:               runtimeAdapterTestBudget(),
		Strategy:             strategy,
		PreGate:              pre,
	})
	if !final.Allowed {
		t.Fatalf("final gate should be allowed: %#v", final)
	}
	return final
}

func runtimeAdapterTestResult(request RuntimeAdapterExecutionRequest, observations []Observation) RuntimeAdapterExecutionResult {
	return BuildRuntimeAdapterExecutionResult(runtimeAdapterTestResultInput(request, observations))
}

func runtimeAdapterTestResultInput(request RuntimeAdapterExecutionRequest, observations []Observation) RuntimeAdapterExecutionResultInput {
	return RuntimeAdapterExecutionResultInput{
		Request:           request,
		AdapterRef:        request.AdapterRef,
		StrategyRef:       request.StrategyRef,
		HostAdapterRunRef: "adapter_run:metrics_1",
		Status:            VerificationSatisfied,
		Observations:      observations,
		OutputRefs:        []DisplaySafeRef{"output:metrics_summary"},
	}
}

func runtimeAdapterTestObservation(strength EvidenceStrength) Observation {
	return Observation{
		Kind:       "metric",
		Source:     "adapter:metrics_readonly",
		Subject:    "objective:metrics",
		Name:       "cpu_usage",
		Value:      "10.5",
		Unit:       "percent",
		Strength:   strength,
		ObservedAt: "2026-06-02T10:00:00Z",
		EvidenceRefs: []EvidenceRef{{
			Ref:      "evidence:metric",
			Kind:     "metric",
			Strength: strength,
			Source:   "adapter:metrics_readonly",
		}},
		DisplaySafeRefs: []DisplaySafeRef{"output:metrics_summary"},
	}
}
