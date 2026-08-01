package controlcontract

import (
	"encoding/json"
	"testing"
)

func TestProductionAdapterObjectiveCloseoutWriterExecutionBridgeReady(t *testing.T) {
	request, handoff := productionAdapterReadyObjectiveCloseoutWriterExecutionBridgeRequestAndHandoff(t, "run:metrics_objective_closeout_writer_durable")
	bridge := BuildProductionAdapterObjectiveCloseoutWriterExecutionBridge(ProductionAdapterObjectiveCloseoutWriterExecutionBridgeInput{
		BridgeRef:         "bridge:metrics_objective_closeout_writer_execution",
		InvocationHandoff: handoff,
		DurableRequest:    request,
	})
	if bridge.Status != "ready_for_objective_closeout_writer_host_adapter_execution_bridge" ||
		!bridge.ReadyForHostDisplay ||
		!bridge.ReadyForHostAdapterExecution ||
		bridge.ReadyForInvocationResultEnvelope ||
		!bridge.HostAdapterExecutionAuthorized ||
		bridge.HostAdapterExecutionBound ||
		!bridge.HostMayInvokeWriterAdapter ||
		!bridge.HostMayExecuteDurableWrite ||
		bridge.NextHostAction != "host_may_execute_objective_closeout_writer_adapter" ||
		bridge.HostUIHandoffRef != handoff.HostUIHandoffRef ||
		bridge.InvocationEnvelopeRef != handoff.InvocationEnvelopeRef ||
		bridge.DurableRequestRef != request.DurableRequestRef ||
		bridge.ExpectedHostAdapterRunRef != handoff.ExpectedHostAdapterRunRef ||
		bridge.ExpectedDurableResultRef != request.ExpectedDurableResultRef ||
		bridge.ExpectedReadbackRef != request.ExpectedReadbackRef ||
		bridge.CoreInvocationExecuted ||
		bridge.DryRunByCore ||
		bridge.DurableWriteByCore ||
		bridge.ObjectiveStoreWriteByCore ||
		bridge.RunstoreWriteByCore {
		t.Fatalf("unexpected execution bridge: %#v", bridge)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "objective closeout writer execution bridge",
		RunnerEffect: bridge.RunnerEffect,
		PromptEffect: bridge.PromptEffect,
		Boundaries:   bridge.Boundaries,
		Payload:      bridge,
	}, "production_adapter_objective_closeout_writer_execution_bridge", "objective_closeout_writer_execution_bridge_projection_only", "host_owned_writer_adapter_execution", "ready_for_objective_closeout_writer_host_adapter_execution_bridge", "no_adapter_invocation_by_core", "no_durable_write_by_core")
	AssertNoCoreMutation(t, "objective closeout writer execution bridge", bridge.CoreInvocationExecuted, bridge.DurableWriteByCore)
}

func TestProductionAdapterObjectiveCloseoutWriterExecutionBridgeCanonicalizesSuccess(t *testing.T) {
	request, handoff := productionAdapterReadyObjectiveCloseoutWriterExecutionBridgeRequestAndHandoff(t, "run:metrics_objective_closeout_writer_durable")
	durableResult := productionAdapterReadyObjectiveCloseoutWriterDurableResult(request)
	bridge := BuildProductionAdapterObjectiveCloseoutWriterExecutionBridge(ProductionAdapterObjectiveCloseoutWriterExecutionBridgeInput{
		BridgeRef:         "bridge:metrics_objective_closeout_writer_execution_result",
		ResultEnvelopeRef: "result_envelope:metrics_objective_closeout_writer_execution",
		InvocationHandoff: handoff,
		DurableRequest:    request,
		DurableResult:     durableResult,
	})
	if bridge.Status != "ready_for_objective_closeout_writer_host_adapter_result_bridge" ||
		!bridge.ReadyForHostDisplay ||
		bridge.ReadyForHostAdapterExecution ||
		!bridge.ReadyForInvocationResultEnvelope ||
		!bridge.ReadyForDurableReadbackReview ||
		bridge.ReadyForFailureReview ||
		!bridge.HostAdapterExecutionBound ||
		!bridge.ResultCanonicalized ||
		!bridge.HostAdapterExecutionReported ||
		!bridge.HostAdapterExecutionRecorded ||
		!bridge.HostAdapterExecutionSucceeded ||
		bridge.HostAdapterExecutionFailed ||
		bridge.HostMayInvokeWriterAdapter ||
		bridge.HostMayExecuteDurableWrite ||
		bridge.NextHostAction != "build_objective_closeout_writer_invocation_result_envelope" ||
		bridge.ResultEnvelopeRef == "" ||
		bridge.DurableResultRef != request.ExpectedDurableResultRef ||
		bridge.HostAdapterRunRef != handoff.ExpectedHostAdapterRunRef ||
		bridge.AppliedDurableEventRef != request.ExpectedDurableEventRef ||
		bridge.AppliedRunstoreRef != request.HostRunstoreRef ||
		bridge.AppliedObjectiveStateRef != request.ExpectedObjectiveStateRef ||
		bridge.CoreInvocationExecuted ||
		bridge.DurableWriteByCore {
		t.Fatalf("unexpected success execution bridge: %#v", bridge)
	}
}

func TestProductionAdapterObjectiveCloseoutWriterExecutionBridgeCanonicalizesFailure(t *testing.T) {
	request, handoff := productionAdapterReadyObjectiveCloseoutWriterExecutionBridgeRequestAndHandoff(t, "run:metrics_objective_closeout_writer_durable_failed")
	durableResult := productionAdapterFailedObjectiveCloseoutWriterDurableResult(request)
	bridge := BuildProductionAdapterObjectiveCloseoutWriterExecutionBridge(ProductionAdapterObjectiveCloseoutWriterExecutionBridgeInput{
		BridgeRef:         "bridge:metrics_objective_closeout_writer_execution_failure",
		ResultEnvelopeRef: "result_envelope:metrics_objective_closeout_writer_execution_failure",
		InvocationHandoff: handoff,
		DurableRequest:    request,
		DurableResult:     durableResult,
	})
	if bridge.Status != "ready_for_objective_closeout_writer_host_adapter_failure_bridge" ||
		!bridge.ReadyForHostDisplay ||
		bridge.ReadyForHostAdapterExecution ||
		!bridge.ReadyForInvocationResultEnvelope ||
		bridge.ReadyForDurableReadbackReview ||
		!bridge.ReadyForFailureReview ||
		!bridge.ReadyForCompensationReview ||
		!bridge.HostAdapterExecutionBound ||
		!bridge.ResultCanonicalized ||
		!bridge.HostAdapterExecutionReported ||
		!bridge.HostAdapterExecutionRecorded ||
		bridge.HostAdapterExecutionSucceeded ||
		!bridge.HostAdapterExecutionFailed ||
		bridge.FailureClass != FailureVerificationFailed ||
		bridge.NextHostAction != "build_objective_closeout_writer_invocation_failure_envelope" ||
		bridge.HostAdapterRunRef != handoff.ExpectedHostAdapterRunRef ||
		bridge.FailureRef != handoff.ExpectedFailureRef ||
		bridge.CompensationRef != handoff.ExpectedCompensationRef ||
		bridge.HostMayExecuteDurableWrite ||
		bridge.DurableWriteByCore ||
		bridge.ObjectiveStoreWriteByCore ||
		bridge.RunstoreWriteByCore {
		t.Fatalf("unexpected failure execution bridge: %#v", bridge)
	}
}

func TestProductionAdapterObjectiveCloseoutWriterExecutionBridgeBlocksMismatchesAndUnsafe(t *testing.T) {
	request, handoff := productionAdapterReadyObjectiveCloseoutWriterExecutionBridgeRequestAndHandoff(t, "run:metrics_objective_closeout_writer_durable")
	mismatch := productionAdapterReadyObjectiveCloseoutWriterDurableResult(request)
	mismatch.HostAdapterRunRef = "run:wrong_metrics_objective_closeout_writer_durable"
	blocked := BuildProductionAdapterObjectiveCloseoutWriterExecutionBridge(ProductionAdapterObjectiveCloseoutWriterExecutionBridgeInput{
		BridgeRef:         "bridge:metrics_objective_closeout_writer_execution_mismatch",
		ResultEnvelopeRef: "result_envelope:metrics_objective_closeout_writer_execution_mismatch",
		InvocationHandoff: handoff,
		DurableRequest:    request,
		DurableResult:     mismatch,
	})
	if blocked.Status != "blocked" ||
		blocked.ReadyForInvocationResultEnvelope ||
		blocked.HostAdapterExecutionBound ||
		blocked.ResultCanonicalized ||
		blocked.FailureClass != FailureVerificationFailed ||
		!productionAdapterStringContains(blocked.BlockedReasons, "writer_execution_bridge_adapter_run_ref_mismatch") ||
		!productionAdapterMissingContains(blocked.MissingInputs, "host:objective_closeout_writer_host_adapter_run_ref") {
		t.Fatalf("expected adapter run ref mismatch to block, got %#v", blocked)
	}

	missingRef := BuildProductionAdapterObjectiveCloseoutWriterExecutionBridge(ProductionAdapterObjectiveCloseoutWriterExecutionBridgeInput{
		InvocationHandoff: handoff,
		DurableRequest:    request,
	})
	if missingRef.Status != "blocked" ||
		missingRef.ReadyForHostAdapterExecution ||
		missingRef.FailureClass != FailureEvidenceMissing ||
		!productionAdapterStringContains(missingRef.BlockedReasons, "writer_execution_bridge_ref_missing") ||
		!productionAdapterMissingContains(missingRef.MissingInputs, "host:objective_closeout_writer_execution_bridge_ref") {
		t.Fatalf("expected missing bridge ref to block, got %#v", missingRef)
	}

	unsafe := BuildProductionAdapterObjectiveCloseoutWriterExecutionBridge(ProductionAdapterObjectiveCloseoutWriterExecutionBridgeInput{
		BridgeRef:         "/tmp/raw-execution-bridge.json",
		InvocationHandoff: handoff,
		DurableRequest:    request,
	})
	if unsafe.Status != "blocked" ||
		unsafe.ReadyForHostDisplay ||
		unsafe.FailureClass != FailureEvidenceWeak ||
		!productionAdapterStringContains(unsafe.BlockedReasons, "unsafe_input_ref") ||
		!productionAdapterMissingContains(unsafe.MissingInputs, "host:display_safe_refs") {
		t.Fatalf("expected unsafe bridge to block, got %#v", unsafe)
	}
	AssertNoRawPayload(t, "unsafe objective closeout writer execution bridge", unsafe, "/tmp/raw-execution-bridge.json")
}

func TestProductionAdapterObjectiveCloseoutWriterExecutionBridgeJSONCompatibility(t *testing.T) {
	request, handoff := productionAdapterReadyObjectiveCloseoutWriterExecutionBridgeRequestAndHandoff(t, "run:metrics_objective_closeout_writer_durable")
	bridge := BuildProductionAdapterObjectiveCloseoutWriterExecutionBridge(ProductionAdapterObjectiveCloseoutWriterExecutionBridgeInput{
		BridgeRef:         "bridge:metrics_objective_closeout_writer_execution",
		InvocationHandoff: handoff,
		DurableRequest:    request,
	})
	raw, err := json.Marshal(bridge)
	if err != nil {
		t.Fatalf("marshal writer execution bridge: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("unmarshal writer execution bridge: %v", err)
	}
	for _, key := range []string{
		"bridge_ref",
		"host_ui_handoff_ref",
		"ready_for_host_adapter_execution",
		"ready_for_invocation_result_envelope",
		"host_adapter_execution_authorized",
		"next_host_action",
	} {
		if _, ok := payload[key]; !ok {
			t.Fatalf("expected stable JSON key %q in %s", key, string(raw))
		}
	}
	for _, token := range []string{
		"ready_for_objective_closeout_writer_host_adapter_execution_bridge",
		"objective_closeout_writer_execution_bridge_projection_only",
		"host_owned_writer_adapter_execution",
		"no_adapter_invocation_by_core",
	} {
		if !jsonPayloadContains(raw, token) {
			t.Fatalf("expected execution bridge JSON token %q in %s", token, raw)
		}
	}
	AssertNoRawPayload(t, "objective closeout writer execution bridge JSON", raw, "/Users/mason", "postgresql://secret", "raw local host task")
}

func productionAdapterReadyObjectiveCloseoutWriterExecutionBridgeRequestAndHandoff(t *testing.T, expectedRunRef DisplaySafeRef) (ProductionAdapterObjectiveCloseoutWriterDurableRequest, ProductionAdapterObjectiveCloseoutWriterInvocationHostUIHandoff) {
	t.Helper()
	request := productionAdapterReadyObjectiveCloseoutWriterDurableRequest()
	packet := productionAdapterReadyObjectiveCloseoutWriterDurableWriteReviewPacket(request)
	envelope := BuildProductionAdapterObjectiveCloseoutWriterInvocationEnvelope(productionAdapterObjectiveCloseoutWriterInvocationEnvelopeInput(packet, expectedRunRef))
	review := BuildProductionAdapterObjectiveCloseoutWriterInvocationReview(ProductionAdapterObjectiveCloseoutWriterInvocationReviewInput{
		ReviewRef:  "review:metrics_objective_closeout_writer_execution",
		Invocation: envelope,
	})
	fixture := BuildProductionAdapterObjectiveCloseoutWriterInvocationReviewBlackboxFixture(ProductionAdapterObjectiveCloseoutWriterInvocationReviewBlackboxFixtureInput{
		FixtureRef: "fixture:metrics_objective_closeout_writer_execution",
		Review:     review,
	})
	handoff := BuildProductionAdapterObjectiveCloseoutWriterInvocationHostUIHandoff(ProductionAdapterObjectiveCloseoutWriterInvocationHostUIHandoffInput{
		HostUIHandoffRef: "ui:metrics_objective_closeout_writer_execution",
		Review:           review,
		DisplayFixture:   fixture,
	})
	if request.Status != HostActionReady ||
		!request.ReadyForHostDurableWrite ||
		handoff.Status != "ready_for_objective_closeout_writer_invocation_handoff" ||
		!handoff.ReadyForHostAdapterInvocation {
		t.Fatalf("test helper expected ready request and handoff, got request=%#v handoff=%#v", request, handoff)
	}
	return request, handoff
}
