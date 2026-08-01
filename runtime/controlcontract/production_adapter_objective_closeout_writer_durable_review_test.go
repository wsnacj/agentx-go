package controlcontract

import (
	"encoding/json"
	"testing"
)

func TestProductionAdapterObjectiveCloseoutWriterDurableReviewPacketReadyForWrite(t *testing.T) {
	request := productionAdapterReadyObjectiveCloseoutWriterDurableRequest()
	packet := BuildProductionAdapterObjectiveCloseoutWriterDurableReviewPacket(ProductionAdapterObjectiveCloseoutWriterDurableReviewPacketInput{
		ReviewPacketRef: "review:metrics_objective_closeout_writer_durable",
		DurableRequest:  request,
	})
	if packet.Status != "ready_for_objective_closeout_writer_durable_write_review" ||
		!packet.ReadyForHostDisplay ||
		!packet.ReadyForHostDurableWrite ||
		packet.ReadyForDurableReadbackReview ||
		packet.ReadyForFailureReview ||
		packet.ReadyForObjectiveReturn ||
		!packet.HostDurableWriteAuthorized ||
		!packet.HostMayExecuteDurableWrite ||
		packet.DurableRequestRef != request.DurableRequestRef ||
		packet.ReviewPacketRef == "" ||
		packet.CoreInvocationExecuted ||
		packet.DryRunByCore ||
		packet.DurableWriteByCore ||
		packet.ObjectiveStoreWriteByCore ||
		packet.RunstoreWriteByCore ||
		packet.NextHostAction != "host_may_execute_objective_closeout_durable_writer_adapter" {
		t.Fatalf("unexpected durable writer write review packet: %#v", packet)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "objective closeout writer durable review packet",
		RunnerEffect: packet.RunnerEffect,
		PromptEffect: packet.PromptEffect,
		Boundaries:   packet.Boundaries,
		Payload:      packet,
	}, "production_adapter_objective_closeout_writer_durable_review_packet", "objective_closeout_writer_durable_review_packet_projection_only", "host_owned_objective_closeout_writer_adapter", "durable_writer_review_only", "ready_for_objective_closeout_writer_durable_write_review", "host_may_execute_durable_writer", "core_durable_write_not_executed")
}

func TestProductionAdapterObjectiveCloseoutWriterDurableReviewPacketReadyForReadback(t *testing.T) {
	request := productionAdapterReadyObjectiveCloseoutWriterDurableRequest()
	result := productionAdapterReadyObjectiveCloseoutWriterDurableResult(request)
	packet := BuildProductionAdapterObjectiveCloseoutWriterDurableReviewPacket(ProductionAdapterObjectiveCloseoutWriterDurableReviewPacketInput{
		ReviewPacketRef: "review:metrics_objective_closeout_writer_durable",
		DurableRequest:  request,
		DurableResult:   result,
	})
	if packet.Status != "ready_for_objective_closeout_writer_durable_readback_review" ||
		!packet.ReadyForHostDisplay ||
		packet.ReadyForHostDurableWrite ||
		!packet.ReadyForDurableResultReview ||
		!packet.ReadyForDurableReadbackReview ||
		packet.ReadyForFailureReview ||
		packet.ReadyForObjectiveReturn ||
		packet.HostMayExecuteDurableWrite ||
		!packet.HostDurableWriteReported ||
		!packet.HostDurableWriteSucceeded ||
		packet.HostDurableWriteFailed ||
		!packet.HostDurableWriteRecorded ||
		packet.DurableResultRef != result.DurableResultRef ||
		packet.AppliedDurableEventRef != result.AppliedDurableEventRef ||
		packet.NextHostAction != "bind_objective_closeout_writer_durable_readback" {
		t.Fatalf("unexpected durable writer readback review packet: %#v", packet)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "objective closeout writer durable readback review packet",
		RunnerEffect: packet.RunnerEffect,
		PromptEffect: packet.PromptEffect,
		Boundaries:   packet.Boundaries,
		Payload:      packet,
	}, "production_adapter_objective_closeout_writer_durable_review_packet", "durable_writer_review_only", "ready_for_objective_closeout_writer_durable_readback_review", "host_objective_closeout_writer_durable_write_recorded")
}

func TestProductionAdapterObjectiveCloseoutWriterDurableReviewPacketFailureReview(t *testing.T) {
	request := productionAdapterReadyObjectiveCloseoutWriterDurableRequest()
	result := productionAdapterFailedObjectiveCloseoutWriterDurableResult(request)
	packet := BuildProductionAdapterObjectiveCloseoutWriterDurableReviewPacket(ProductionAdapterObjectiveCloseoutWriterDurableReviewPacketInput{
		ReviewPacketRef: "review:metrics_objective_closeout_writer_durable_failure",
		DurableRequest:  request,
		DurableResult:   result,
	})
	if packet.Status != "ready_for_objective_closeout_writer_durable_failure_review" ||
		!packet.ReadyForHostDisplay ||
		packet.ReadyForHostDurableWrite ||
		packet.ReadyForDurableReadbackReview ||
		!packet.ReadyForFailureReview ||
		!packet.ReadyForCompensationReview ||
		packet.ReadyForObjectiveReturn ||
		packet.HostMayExecuteDurableWrite ||
		!packet.HostDurableWriteReported ||
		packet.HostDurableWriteSucceeded ||
		!packet.HostDurableWriteFailed ||
		!packet.HostDurableWriteRecorded ||
		packet.FailureRef == "" ||
		packet.CompensationRef == "" ||
		packet.NextHostAction != "review_objective_closeout_writer_durable_failure" {
		t.Fatalf("unexpected durable writer failure review packet: %#v", packet)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "objective closeout writer durable failure review packet",
		RunnerEffect: packet.RunnerEffect,
		PromptEffect: packet.PromptEffect,
		Boundaries:   packet.Boundaries,
		Payload:      packet,
	}, "production_adapter_objective_closeout_writer_durable_review_packet", "durable_writer_review_only", "ready_for_objective_closeout_writer_durable_failure_review", "compensation_not_executed")
}

func TestProductionAdapterObjectiveCloseoutWriterDurableReviewBlackboxFixtureReturnDisplay(t *testing.T) {
	request := productionAdapterReadyObjectiveCloseoutWriterDurableRequest()
	result := productionAdapterReadyObjectiveCloseoutWriterDurableResult(request)
	readback := BuildProductionAdapterObjectiveCloseoutWriterDurableReadback(ProductionAdapterObjectiveCloseoutWriterDurableReadbackInput{
		DurableReadbackRef:        "readback_binding:metrics_objective_closeout_writer",
		DurableResult:             result,
		ObjectiveCloseoutReadback: productionAdapterReadyObjectiveCloseoutWriterCloseoutReadback(result),
	})
	packet := BuildProductionAdapterObjectiveCloseoutWriterDurableReviewPacket(ProductionAdapterObjectiveCloseoutWriterDurableReviewPacketInput{
		ReviewPacketRef: "review:metrics_objective_closeout_writer_durable_final",
		DurableRequest:  request,
		DurableResult:   result,
		DurableReadback: readback,
	})
	fixture := BuildProductionAdapterObjectiveCloseoutWriterDurableReviewBlackboxFixture(ProductionAdapterObjectiveCloseoutWriterDurableReviewBlackboxFixtureInput{
		FixtureRef:   "fixture:metrics_objective_closeout_writer_durable_final",
		ReviewPacket: packet,
	})
	if fixture.Status != "ready_for_objective_closeout_writer_durable_return_display" ||
		!fixture.ReadyForHostDisplay ||
		fixture.ReadyForHostDurableWrite ||
		fixture.ReadyForDurableReadbackReview ||
		fixture.ReadyForFailureReview ||
		!fixture.ReadyForObjectiveReturn ||
		fixture.HostMayExecuteDurableWrite ||
		!fixture.WriterDurableReadbackBound ||
		!fixture.ObjectiveLifecycleClosed ||
		!fixture.ObjectiveSatisfied ||
		fixture.DurableReadbackRef != readback.DurableReadbackRef ||
		fixture.ObjectiveCloseoutReadbackRef != readback.ObjectiveCloseoutReadbackRef ||
		fixture.CoreInvocationExecuted ||
		fixture.DryRunByCore ||
		fixture.DurableWriteByCore ||
		fixture.ObjectiveStoreWriteByCore ||
		fixture.RunstoreWriteByCore ||
		fixture.NextHostAction != "return_objective_closed_lifecycle" {
		t.Fatalf("unexpected durable writer final fixture: %#v", fixture)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "objective closeout writer durable return fixture",
		RunnerEffect: fixture.RunnerEffect,
		PromptEffect: fixture.PromptEffect,
		Boundaries:   fixture.Boundaries,
		Payload:      fixture,
	}, "production_adapter_objective_closeout_writer_durable_review_blackbox_fixture", "objective_closeout_writer_durable_review_fixture_projection_only", "host_cli_objective_closeout_writer_durable_display_ready")
}

func TestProductionAdapterObjectiveCloseoutWriterDurableReviewPacketBlockedDisplayAndUnsafe(t *testing.T) {
	optIn, smoke := productionAdapterReadyObjectiveCloseoutWriterDurableOptInAndSmoke()
	blockedRequest := BuildProductionAdapterObjectiveCloseoutWriterDurableRequest(ProductionAdapterObjectiveCloseoutWriterDurableRequestInput{
		DurableRequestRef:        "durable_request:metrics_objective_closeout_writer_blocked",
		WriterOptIn:              optIn,
		DryRunSmoke:              smoke,
		ExpectedDurableResultRef: "durable_result:metrics_objective_closeout_writer_blocked",
	})
	packet := BuildProductionAdapterObjectiveCloseoutWriterDurableReviewPacket(ProductionAdapterObjectiveCloseoutWriterDurableReviewPacketInput{
		ReviewPacketRef: "review:metrics_objective_closeout_writer_durable_blocked",
		DurableRequest:  blockedRequest,
	})
	if packet.Status != "ready_for_objective_closeout_writer_durable_blocked_review" ||
		!packet.ReadyForHostDisplay ||
		!packet.ReadyForBlockedReview ||
		packet.ReadyForHostDurableWrite ||
		packet.HostMayExecuteDurableWrite ||
		!productionAdapterStringContains(packet.BlockedReasons, "host_durable_write_confirmation_ref_missing") ||
		!productionAdapterMissingContains(packet.MissingInputs, "host:objective_closeout_writer_durable_write_confirmation_ref") {
		t.Fatalf("unexpected blocked durable writer review packet: %#v", packet)
	}
	fixture := BuildProductionAdapterObjectiveCloseoutWriterDurableReviewBlackboxFixture(ProductionAdapterObjectiveCloseoutWriterDurableReviewBlackboxFixtureInput{
		FixtureRef:   "fixture:metrics_objective_closeout_writer_durable_blocked",
		ReviewPacket: packet,
	})
	if fixture.Status != "ready_for_objective_closeout_writer_durable_blocked_display" ||
		!fixture.ReadyForHostDisplay ||
		!fixture.ReadyForBlockedReview ||
		fixture.HostMayExecuteDurableWrite {
		t.Fatalf("unexpected blocked durable writer review fixture: %#v", fixture)
	}
	unsafe := BuildProductionAdapterObjectiveCloseoutWriterDurableReviewPacket(ProductionAdapterObjectiveCloseoutWriterDurableReviewPacketInput{
		ReviewPacketRef: "postgresql://secret@example.invalid/db",
		DurableRequest:  blockedRequest,
	})
	if unsafe.ReadyForHostDisplay ||
		unsafe.ReadyForBlockedReview ||
		!productionAdapterStringContains(unsafe.BlockedReasons, "unsafe_input_ref") {
		t.Fatalf("expected unsafe durable writer review packet to block: %#v", unsafe)
	}
	AssertNoRawPayload(t, "unsafe durable writer review packet", unsafe, "postgresql://secret", "example.invalid")
}

func TestProductionAdapterObjectiveCloseoutWriterDurableReviewJSONCompatibility(t *testing.T) {
	request := productionAdapterReadyObjectiveCloseoutWriterDurableRequest()
	result := productionAdapterReadyObjectiveCloseoutWriterDurableResult(request)
	readback := BuildProductionAdapterObjectiveCloseoutWriterDurableReadback(ProductionAdapterObjectiveCloseoutWriterDurableReadbackInput{
		DurableReadbackRef:        "readback_binding:metrics_objective_closeout_writer",
		DurableResult:             result,
		ObjectiveCloseoutReadback: productionAdapterReadyObjectiveCloseoutWriterCloseoutReadback(result),
	})
	packet := BuildProductionAdapterObjectiveCloseoutWriterDurableReviewPacket(ProductionAdapterObjectiveCloseoutWriterDurableReviewPacketInput{
		ReviewPacketRef: "review:metrics_objective_closeout_writer_durable_final",
		DurableRequest:  request,
		DurableResult:   result,
		DurableReadback: readback,
	})
	fixture := BuildProductionAdapterObjectiveCloseoutWriterDurableReviewBlackboxFixture(ProductionAdapterObjectiveCloseoutWriterDurableReviewBlackboxFixtureInput{
		FixtureRef:   "fixture:metrics_objective_closeout_writer_durable_final",
		ReviewPacket: packet,
	})
	raw, err := json.Marshal(struct {
		Packet  ProductionAdapterObjectiveCloseoutWriterDurableReviewPacket          `json:"packet"`
		Fixture ProductionAdapterObjectiveCloseoutWriterDurableReviewBlackboxFixture `json:"fixture"`
	}{Packet: packet, Fixture: fixture})
	if err != nil {
		t.Fatalf("marshal durable review contracts: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("unmarshal durable review contracts: %v", err)
	}
	for _, token := range []string{
		"review_packet_ref",
		"ready_for_objective_closeout_writer_durable_return_review",
		"ready_for_objective_closeout_writer_durable_return_display",
		"host_cli_objective_closeout_writer_durable_display",
		"durable_writer_review_only",
	} {
		if !jsonPayloadContains(raw, token) {
			t.Fatalf("expected durable review JSON token %q in %s", token, raw)
		}
	}
	AssertNoRawPayload(t, "objective closeout writer durable review JSON", raw, "/Users/mason", "postgresql://secret", "raw local host task")
}

func productionAdapterFailedObjectiveCloseoutWriterDurableResult(request ProductionAdapterObjectiveCloseoutWriterDurableRequest) ProductionAdapterObjectiveCloseoutWriterDurableResult {
	return BuildProductionAdapterObjectiveCloseoutWriterDurableResult(ProductionAdapterObjectiveCloseoutWriterDurableResultInput{
		DurableResultRef:         request.ExpectedDurableResultRef,
		DurableRequest:           request,
		HostAdapterRunRef:        "run:metrics_objective_closeout_writer_durable_failed",
		HostDurableWriteReported: true,
		HostDurableWriteFailed:   true,
		FailureRef:               "failure:metrics_objective_closeout_writer_durable",
		CompensationRef:          "compensation:metrics_objective_closeout_writer_review",
	})
}
