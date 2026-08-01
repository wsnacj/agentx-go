package controlcontract

import (
	"encoding/json"
	"testing"
)

func TestProductionAdapterObjectiveCloseoutWriterRunnerGateReady(t *testing.T) {
	bridge := productionAdapterReadyObjectiveCloseoutWriterRunnerGateBridge(t)
	gate := BuildProductionAdapterObjectiveCloseoutWriterRunnerGate(productionAdapterObjectiveCloseoutWriterRunnerGateInput(bridge))
	if gate.Status != "ready_for_objective_closeout_writer_host_adapter_runner_gate" ||
		!gate.ReadyForHostDisplay ||
		!gate.ReadyForHostRunner ||
		!gate.ReadyForHostAdapterExecution ||
		!gate.HostRunnerAuthorized ||
		!gate.HostConfirmationBound ||
		!gate.DryRunFirstSatisfied ||
		!gate.PolicyBound ||
		!gate.ApprovalBound ||
		!gate.BudgetBound ||
		!gate.IdempotencyBound ||
		!gate.TimeoutBound ||
		!gate.HostMayInvokeWriterAdapter ||
		!gate.HostMayExecuteDurableWrite ||
		gate.NextHostAction != "host_may_run_objective_closeout_writer_adapter" ||
		gate.BridgeRef != bridge.BridgeRef ||
		gate.DurableRequestRef != bridge.DurableRequestRef ||
		gate.ExpectedHostAdapterRunRef != bridge.ExpectedHostAdapterRunRef ||
		gate.ExpectedDurableResultRef != bridge.ExpectedDurableResultRef ||
		gate.BudgetRef != bridge.RequiredBudgetRef ||
		gate.IdempotencyRef != bridge.IdempotencyRef ||
		gate.TimeoutPolicyRef != bridge.TimeoutPolicyRef ||
		gate.CoreInvocationExecuted ||
		gate.DryRunByCore ||
		gate.DurableWriteByCore ||
		gate.ObjectiveStoreWriteByCore ||
		gate.RunstoreWriteByCore {
		t.Fatalf("unexpected writer runner gate: %#v", gate)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "objective closeout writer runner gate",
		RunnerEffect: gate.RunnerEffect,
		PromptEffect: gate.PromptEffect,
		Boundaries:   gate.Boundaries,
		Payload:      gate,
	}, "production_adapter_objective_closeout_writer_runner_gate", "objective_closeout_writer_runner_gate_projection_only", "host_owned_writer_adapter_runner_facade", "dry_run_first_runtime_gate", "explicit_host_confirmation_required", "ready_for_objective_closeout_writer_host_adapter_runner_gate", "no_adapter_invocation_by_core", "no_durable_write_by_core")
	AssertNoCoreMutation(t, "objective closeout writer runner gate", gate.CoreInvocationExecuted, gate.DurableWriteByCore)
}

func TestProductionAdapterObjectiveCloseoutWriterRunnerGateBlocksMissingControls(t *testing.T) {
	bridge := productionAdapterReadyObjectiveCloseoutWriterRunnerGateBridge(t)
	input := productionAdapterObjectiveCloseoutWriterRunnerGateInput(bridge)
	input.HostConfirmationRef = ""
	input.ApprovalBindingRefs = nil
	input.BudgetRef = ""
	input.IdempotencyRef = ""
	input.TimeoutPolicyRef = ""
	input.CancellationPolicyRef = ""
	blocked := BuildProductionAdapterObjectiveCloseoutWriterRunnerGate(input)
	if blocked.Status != "blocked" ||
		blocked.ReadyForHostRunner ||
		blocked.HostRunnerAuthorized ||
		blocked.HostMayInvokeWriterAdapter ||
		blocked.FailureClass != FailureAuthorizationMissing ||
		!productionAdapterStringContains(blocked.BlockedReasons, "host_confirmation_ref_missing") ||
		!productionAdapterStringContains(blocked.BlockedReasons, "approval_binding_ref_missing") ||
		!productionAdapterStringContains(blocked.BlockedReasons, "budget_ref_missing") ||
		!productionAdapterStringContains(blocked.BlockedReasons, "idempotency_ref_missing") ||
		!productionAdapterStringContains(blocked.BlockedReasons, "timeout_policy_ref_missing") ||
		!productionAdapterStringContains(blocked.BlockedReasons, "cancellation_policy_ref_missing") {
		t.Fatalf("expected missing runner controls to block, got %#v", blocked)
	}
}

func TestProductionAdapterObjectiveCloseoutWriterRunnerGateBlocksBridgeNotReadyAndMismatches(t *testing.T) {
	request, handoff := productionAdapterReadyObjectiveCloseoutWriterExecutionBridgeRequestAndHandoff(t, "run:metrics_objective_closeout_writer_durable")
	resultBridge := BuildProductionAdapterObjectiveCloseoutWriterExecutionBridge(ProductionAdapterObjectiveCloseoutWriterExecutionBridgeInput{
		BridgeRef:         "bridge:metrics_objective_closeout_writer_execution_result",
		ResultEnvelopeRef: "result_envelope:metrics_objective_closeout_writer_execution",
		InvocationHandoff: handoff,
		DurableRequest:    request,
		DurableResult:     productionAdapterReadyObjectiveCloseoutWriterDurableResult(request),
	})
	notReady := BuildProductionAdapterObjectiveCloseoutWriterRunnerGate(productionAdapterObjectiveCloseoutWriterRunnerGateInput(resultBridge))
	if notReady.Status != "blocked" ||
		notReady.ReadyForHostRunner ||
		notReady.HostMayInvokeWriterAdapter ||
		!productionAdapterStringContains(notReady.BlockedReasons, "writer_execution_bridge_not_ready") {
		t.Fatalf("expected result bridge to be rejected before runner gate, got %#v", notReady)
	}

	bridge := productionAdapterReadyObjectiveCloseoutWriterRunnerGateBridge(t)
	budgetMismatchInput := productionAdapterObjectiveCloseoutWriterRunnerGateInput(bridge)
	budgetMismatchInput.BudgetRef = "budget:wrong_objective_closeout_writer"
	budgetMismatch := BuildProductionAdapterObjectiveCloseoutWriterRunnerGate(budgetMismatchInput)
	if budgetMismatch.Status != "blocked" ||
		budgetMismatch.ReadyForHostRunner ||
		budgetMismatch.FailureClass != FailurePolicyBlocked ||
		!productionAdapterStringContains(budgetMismatch.BlockedReasons, "budget_ref_mismatch") ||
		!productionAdapterMissingContains(budgetMismatch.MissingInputs, "host:budget_ref") {
		t.Fatalf("expected budget mismatch to block, got %#v", budgetMismatch)
	}

	idempotencyMismatchInput := productionAdapterObjectiveCloseoutWriterRunnerGateInput(bridge)
	idempotencyMismatchInput.IdempotencyRef = "idempotency:wrong_metrics_objective_closeout_request"
	idempotencyMismatch := BuildProductionAdapterObjectiveCloseoutWriterRunnerGate(idempotencyMismatchInput)
	if idempotencyMismatch.Status != "blocked" ||
		idempotencyMismatch.ReadyForHostRunner ||
		idempotencyMismatch.FailureClass != FailureInvalidInput ||
		!productionAdapterStringContains(idempotencyMismatch.BlockedReasons, "idempotency_ref_mismatch") ||
		!productionAdapterMissingContains(idempotencyMismatch.MissingInputs, "host:idempotency_ref") {
		t.Fatalf("expected idempotency mismatch to block, got %#v", idempotencyMismatch)
	}
}

func TestProductionAdapterObjectiveCloseoutWriterRunnerGateRejectsUnsafeRefs(t *testing.T) {
	bridge := productionAdapterReadyObjectiveCloseoutWriterRunnerGateBridge(t)
	input := productionAdapterObjectiveCloseoutWriterRunnerGateInput(bridge)
	input.RunnerGateRef = "/tmp/raw-runner-gate.json"
	unsafe := BuildProductionAdapterObjectiveCloseoutWriterRunnerGate(input)
	if unsafe.Status != "blocked" ||
		unsafe.ReadyForHostDisplay ||
		unsafe.FailureClass != FailureEvidenceWeak ||
		!productionAdapterStringContains(unsafe.BlockedReasons, "unsafe_input_ref") ||
		!productionAdapterMissingContains(unsafe.MissingInputs, "host:display_safe_refs") {
		t.Fatalf("expected unsafe runner gate to block, got %#v", unsafe)
	}
	AssertNoRawPayload(t, "unsafe objective closeout writer runner gate", unsafe, "/tmp/raw-runner-gate.json")
}

func TestProductionAdapterObjectiveCloseoutWriterRunnerGateJSONCompatibility(t *testing.T) {
	bridge := productionAdapterReadyObjectiveCloseoutWriterRunnerGateBridge(t)
	gate := BuildProductionAdapterObjectiveCloseoutWriterRunnerGate(productionAdapterObjectiveCloseoutWriterRunnerGateInput(bridge))
	raw, err := json.Marshal(gate)
	if err != nil {
		t.Fatalf("marshal writer runner gate: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("unmarshal writer runner gate: %v", err)
	}
	for _, key := range []string{
		"runner_gate_ref",
		"bridge_ref",
		"host_runner_ref",
		"host_confirmation_ref",
		"ready_for_host_runner",
		"dry_run_first_satisfied",
		"next_host_action",
	} {
		if _, ok := payload[key]; !ok {
			t.Fatalf("expected stable JSON key %q in %s", key, string(raw))
		}
	}
	for _, token := range []string{
		"ready_for_objective_closeout_writer_host_adapter_runner_gate",
		"objective_closeout_writer_runner_gate_projection_only",
		"host_owned_writer_adapter_runner_facade",
		"dry_run_first_runtime_gate",
		"no_adapter_invocation_by_core",
	} {
		if !jsonPayloadContains(raw, token) {
			t.Fatalf("expected runner gate JSON token %q in %s", token, raw)
		}
	}
	AssertNoRawPayload(t, "objective closeout writer runner gate JSON", raw, "/Users/mason", "postgresql://secret", "raw local host task")
}

func productionAdapterReadyObjectiveCloseoutWriterRunnerGateBridge(t *testing.T) ProductionAdapterObjectiveCloseoutWriterExecutionBridge {
	t.Helper()
	request, handoff := productionAdapterReadyObjectiveCloseoutWriterExecutionBridgeRequestAndHandoff(t, "run:metrics_objective_closeout_writer_durable")
	bridge := BuildProductionAdapterObjectiveCloseoutWriterExecutionBridge(ProductionAdapterObjectiveCloseoutWriterExecutionBridgeInput{
		BridgeRef:         "bridge:metrics_objective_closeout_writer_execution",
		InvocationHandoff: handoff,
		DurableRequest:    request,
	})
	if bridge.Status != "ready_for_objective_closeout_writer_host_adapter_execution_bridge" ||
		!bridge.ReadyForHostAdapterExecution {
		t.Fatalf("test helper expected ready execution bridge, got %#v", bridge)
	}
	return bridge
}

func productionAdapterObjectiveCloseoutWriterRunnerGateInput(bridge ProductionAdapterObjectiveCloseoutWriterExecutionBridge) ProductionAdapterObjectiveCloseoutWriterRunnerGateInput {
	return ProductionAdapterObjectiveCloseoutWriterRunnerGateInput{
		RunnerGateRef:         "runner_gate:metrics_objective_closeout_writer",
		ExecutionBridge:       bridge,
		HostRunnerRef:         "runner:metrics_objective_closeout_writer",
		HostRunnerVersionRef:  "version:metrics_objective_closeout_writer_runner_v1",
		RunnerInvocationRef:   "runner_invocation:metrics_objective_closeout_writer",
		HostConfirmationRef:   "confirmation:metrics_objective_closeout_writer_runner",
		PolicyRefs:            cloneDisplaySafeRefs(bridge.RequiredPolicyRefs),
		ApprovalBindingRefs:   cloneDisplaySafeRefs(bridge.ApprovalBindingRefs),
		BudgetRef:             bridge.RequiredBudgetRef,
		IdempotencyRef:        bridge.IdempotencyRef,
		TimeoutPolicyRef:      bridge.TimeoutPolicyRef,
		CancellationPolicyRef: "cancellation:metrics_objective_closeout_writer_runner",
	}
}
