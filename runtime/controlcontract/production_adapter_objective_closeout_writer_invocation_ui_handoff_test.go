package controlcontract

import (
	"encoding/json"
	"testing"
)

func TestProductionAdapterObjectiveCloseoutWriterInvocationHostUIHandoffReady(t *testing.T) {
	review, fixture := productionAdapterReadyObjectiveCloseoutWriterInvocationReviewAndFixture(t)
	handoff := BuildProductionAdapterObjectiveCloseoutWriterInvocationHostUIHandoff(ProductionAdapterObjectiveCloseoutWriterInvocationHostUIHandoffInput{
		HostUIHandoffRef: "ui:metrics_objective_closeout_writer_invocation",
		Review:           review,
		DisplayFixture:   fixture,
	})
	if handoff.Status != "ready_for_objective_closeout_writer_invocation_handoff" ||
		handoff.DisplayState != "host_adapter_invocation_ready" ||
		handoff.DisplayStage != "invocation_ready" ||
		!productionAdapterStringContains(handoff.DisplaySteps, "writer_adapter_invocation") ||
		!handoff.ReadyForHostDisplay ||
		!handoff.ReadyForHostAdapterInvocation ||
		!handoff.InvocationReadyDisplay ||
		handoff.ResultReadbackDisplay ||
		handoff.FailureReviewDisplay ||
		handoff.BlockedDisplay ||
		!handoff.HostMayInvokeWriterAdapter ||
		!handoff.HostMayExecuteDurableWrite ||
		!handoff.HostAdapterInvocationAuthorized ||
		handoff.HostAdapterInvocationBound ||
		handoff.PrimaryDisplayRef != fixture.FixtureRef ||
		handoff.ReviewRef != review.ReviewRef ||
		handoff.FixtureRef != fixture.FixtureRef ||
		handoff.NextHostAction != "host_may_invoke_objective_closeout_writer_adapter" ||
		handoff.CoreInvocationExecuted ||
		handoff.DryRunByCore ||
		handoff.DurableWriteByCore ||
		handoff.ObjectiveStoreWriteByCore ||
		handoff.RunstoreWriteByCore {
		t.Fatalf("unexpected writer invocation host UI handoff: %#v", handoff)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "objective closeout writer invocation host UI handoff",
		RunnerEffect: handoff.RunnerEffect,
		PromptEffect: handoff.PromptEffect,
		Boundaries:   handoff.Boundaries,
		Payload:      handoff,
	}, "production_adapter_objective_closeout_writer_invocation_host_ui_handoff", "objective_closeout_writer_invocation_host_ui_handoff_projection_only", "host_ui_objective_closeout_writer_invocation_handoff", "ready_for_objective_closeout_writer_invocation_handoff", "no_adapter_invocation", "no_durable_write_by_core")
	AssertNoCoreMutation(t, "objective closeout writer invocation host UI handoff", handoff.CoreInvocationExecuted, handoff.DurableWriteByCore)
}

func TestProductionAdapterObjectiveCloseoutWriterInvocationHostUIHandoffResultDisplays(t *testing.T) {
	request := productionAdapterReadyObjectiveCloseoutWriterDurableRequest()
	packet := productionAdapterReadyObjectiveCloseoutWriterDurableWriteReviewPacket(request)
	envelope := BuildProductionAdapterObjectiveCloseoutWriterInvocationEnvelope(productionAdapterObjectiveCloseoutWriterInvocationEnvelopeInput(packet, "run:metrics_objective_closeout_writer_durable"))
	resultEnvelope := BuildProductionAdapterObjectiveCloseoutWriterInvocationResultEnvelope(ProductionAdapterObjectiveCloseoutWriterInvocationResultEnvelopeInput{
		ResultEnvelopeRef:  "result_envelope:metrics_objective_closeout_writer",
		InvocationEnvelope: envelope,
		DurableResult:      productionAdapterReadyObjectiveCloseoutWriterDurableResult(request),
	})
	readbackReview := BuildProductionAdapterObjectiveCloseoutWriterInvocationReview(ProductionAdapterObjectiveCloseoutWriterInvocationReviewInput{
		ReviewRef:        "review:metrics_objective_closeout_writer_invocation_result",
		Invocation:       envelope,
		InvocationResult: resultEnvelope,
	})
	readbackFixture := BuildProductionAdapterObjectiveCloseoutWriterInvocationReviewBlackboxFixture(ProductionAdapterObjectiveCloseoutWriterInvocationReviewBlackboxFixtureInput{
		FixtureRef: "fixture:metrics_objective_closeout_writer_invocation_result",
		Review:     readbackReview,
	})
	readbackHandoff := BuildProductionAdapterObjectiveCloseoutWriterInvocationHostUIHandoff(ProductionAdapterObjectiveCloseoutWriterInvocationHostUIHandoffInput{
		HostUIHandoffRef: "ui:metrics_objective_closeout_writer_invocation_result",
		Review:           readbackReview,
		DisplayFixture:   readbackFixture,
	})
	if readbackHandoff.Status != "ready_for_objective_closeout_writer_invocation_result_readback_handoff" ||
		readbackHandoff.DisplayStage != "result_readback" ||
		!readbackHandoff.ResultReadbackDisplay ||
		!readbackHandoff.ReadyForDurableReadbackReview ||
		!readbackHandoff.HostAdapterInvocationBound ||
		readbackHandoff.HostMayInvokeWriterAdapter ||
		readbackHandoff.HostMayExecuteDurableWrite ||
		readbackHandoff.ReadyForFailureReview ||
		readbackHandoff.NextHostAction != "bind_objective_closeout_writer_durable_readback" {
		t.Fatalf("unexpected writer invocation readback UI handoff: %#v", readbackHandoff)
	}

	failureEnvelope := BuildProductionAdapterObjectiveCloseoutWriterInvocationEnvelope(productionAdapterObjectiveCloseoutWriterInvocationEnvelopeInput(packet, "run:metrics_objective_closeout_writer_durable_failed"))
	failureResult := BuildProductionAdapterObjectiveCloseoutWriterInvocationResultEnvelope(ProductionAdapterObjectiveCloseoutWriterInvocationResultEnvelopeInput{
		ResultEnvelopeRef:  "result_envelope:metrics_objective_closeout_writer_failure",
		InvocationEnvelope: failureEnvelope,
		DurableResult:      productionAdapterFailedObjectiveCloseoutWriterDurableResult(request),
	})
	failureReview := BuildProductionAdapterObjectiveCloseoutWriterInvocationReview(ProductionAdapterObjectiveCloseoutWriterInvocationReviewInput{
		ReviewRef:        "review:metrics_objective_closeout_writer_invocation_failure",
		Invocation:       failureEnvelope,
		InvocationResult: failureResult,
	})
	failureFixture := BuildProductionAdapterObjectiveCloseoutWriterInvocationReviewBlackboxFixture(ProductionAdapterObjectiveCloseoutWriterInvocationReviewBlackboxFixtureInput{
		FixtureRef: "fixture:metrics_objective_closeout_writer_invocation_failure",
		Review:     failureReview,
	})
	failureHandoff := BuildProductionAdapterObjectiveCloseoutWriterInvocationHostUIHandoff(ProductionAdapterObjectiveCloseoutWriterInvocationHostUIHandoffInput{
		HostUIHandoffRef: "ui:metrics_objective_closeout_writer_invocation_failure",
		Review:           failureReview,
		DisplayFixture:   failureFixture,
	})
	if failureHandoff.Status != "ready_for_objective_closeout_writer_invocation_result_failure_handoff" ||
		failureHandoff.DisplayState != "invocation_result_failed" ||
		failureHandoff.DisplayStage != "failure_review" ||
		!failureHandoff.FailureReviewDisplay ||
		!failureHandoff.ReadyForFailureReview ||
		!failureHandoff.ReadyForCompensationReview ||
		!failureHandoff.HostAdapterInvocationBound ||
		failureHandoff.HostMayInvokeWriterAdapter ||
		failureHandoff.HostMayExecuteDurableWrite ||
		failureHandoff.NextHostAction != "review_objective_closeout_writer_durable_failure" {
		t.Fatalf("unexpected writer invocation failure UI handoff: %#v", failureHandoff)
	}
	AssertNoCoreMutation(t, "objective closeout writer invocation failure host UI handoff", failureHandoff.CoreInvocationExecuted, failureHandoff.DurableWriteByCore)
}

func TestProductionAdapterObjectiveCloseoutWriterInvocationHostUIHandoffBlockedDisplay(t *testing.T) {
	request := productionAdapterReadyObjectiveCloseoutWriterDurableRequest()
	packet := productionAdapterReadyObjectiveCloseoutWriterDurableWriteReviewPacket(request)
	envelope := BuildProductionAdapterObjectiveCloseoutWriterInvocationEnvelope(productionAdapterObjectiveCloseoutWriterInvocationEnvelopeInput(packet, "run:metrics_objective_closeout_writer_durable"))
	mismatchResult := productionAdapterReadyObjectiveCloseoutWriterDurableResult(request)
	mismatchResult.ExpectedReadbackRef = "readback:wrong_metrics_objective_closeout_writer"
	mismatchEnvelope := BuildProductionAdapterObjectiveCloseoutWriterInvocationResultEnvelope(ProductionAdapterObjectiveCloseoutWriterInvocationResultEnvelopeInput{
		ResultEnvelopeRef:  "result_envelope:metrics_objective_closeout_writer_mismatch",
		InvocationEnvelope: envelope,
		DurableResult:      mismatchResult,
	})
	review := BuildProductionAdapterObjectiveCloseoutWriterInvocationReview(ProductionAdapterObjectiveCloseoutWriterInvocationReviewInput{
		ReviewRef:        "review:metrics_objective_closeout_writer_invocation_mismatch",
		Invocation:       envelope,
		InvocationResult: mismatchEnvelope,
	})
	fixture := BuildProductionAdapterObjectiveCloseoutWriterInvocationReviewBlackboxFixture(ProductionAdapterObjectiveCloseoutWriterInvocationReviewBlackboxFixtureInput{
		FixtureRef: "fixture:metrics_objective_closeout_writer_invocation_mismatch",
		Review:     review,
	})
	handoff := BuildProductionAdapterObjectiveCloseoutWriterInvocationHostUIHandoff(ProductionAdapterObjectiveCloseoutWriterInvocationHostUIHandoffInput{
		HostUIHandoffRef: "ui:metrics_objective_closeout_writer_invocation_mismatch",
		Review:           review,
		DisplayFixture:   fixture,
	})
	if handoff.Status != "ready_for_objective_closeout_writer_invocation_blocked_handoff" ||
		handoff.DisplayState != "blocked_result_readback_mismatch" ||
		handoff.DisplayStage != "blocked_review" ||
		!handoff.BlockedDisplay ||
		!handoff.ReadyForBlockedReview ||
		handoff.HostMayInvokeWriterAdapter ||
		handoff.HostMayExecuteDurableWrite ||
		!productionAdapterStringContains(handoff.BlockedReasons, "writer_invocation_result_expected_readback_ref_mismatch") ||
		!productionAdapterMissingContains(handoff.MissingInputs, "host:objective_closeout_writer_expected_readback_ref") {
		t.Fatalf("unexpected writer invocation blocked UI handoff: %#v", handoff)
	}
}

func TestProductionAdapterObjectiveCloseoutWriterInvocationHostUIHandoffRejectsMismatchedFixture(t *testing.T) {
	review, fixture := productionAdapterReadyObjectiveCloseoutWriterInvocationReviewAndFixture(t)
	fixture.ReviewRef = "review:other_metrics_objective_closeout_writer_invocation"
	handoff := BuildProductionAdapterObjectiveCloseoutWriterInvocationHostUIHandoff(ProductionAdapterObjectiveCloseoutWriterInvocationHostUIHandoffInput{
		HostUIHandoffRef: "ui:metrics_objective_closeout_writer_invocation",
		Review:           review,
		DisplayFixture:   fixture,
	})
	if handoff.Status != "blocked" ||
		handoff.ReadyForHostAdapterInvocation ||
		handoff.InvocationReadyDisplay ||
		!productionAdapterStringContains(handoff.BlockedReasons, "writer_invocation_handoff_fixture_review_ref_mismatch") ||
		!productionAdapterMissingContains(handoff.MissingInputs, "host:objective_closeout_writer_invocation_review_ref") {
		t.Fatalf("expected writer invocation fixture mismatch to block UI handoff, got %#v", handoff)
	}
}

func TestProductionAdapterObjectiveCloseoutWriterInvocationHostUIHandoffJSONCompatibility(t *testing.T) {
	review, fixture := productionAdapterReadyObjectiveCloseoutWriterInvocationReviewAndFixture(t)
	handoff := BuildProductionAdapterObjectiveCloseoutWriterInvocationHostUIHandoff(ProductionAdapterObjectiveCloseoutWriterInvocationHostUIHandoffInput{
		HostUIHandoffRef: "ui:metrics_objective_closeout_writer_invocation",
		Review:           review,
		DisplayFixture:   fixture,
	})
	raw, err := json.Marshal(handoff)
	if err != nil {
		t.Fatalf("marshal writer invocation host UI handoff: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("unmarshal writer invocation host UI handoff: %v", err)
	}
	for _, key := range []string{
		"display_state",
		"display_stage",
		"display_steps",
		"primary_display_ref",
		"review_ref",
		"fixture_ref",
		"ready_for_host_adapter_invocation",
		"next_host_action",
	} {
		if _, ok := payload[key]; !ok {
			t.Fatalf("expected stable JSON key %q in %s", key, string(raw))
		}
	}
	if payload["display_state"] != "host_adapter_invocation_ready" ||
		payload["display_stage"] != "invocation_ready" ||
		payload["next_host_action"] != "host_may_invoke_objective_closeout_writer_adapter" {
		t.Fatalf("unexpected stable writer invocation UI handoff JSON: %s", string(raw))
	}
	AssertNoRawPayload(t, "objective closeout writer invocation host UI handoff JSON", handoff, "/Users/mason", "postgresql://secret", "raw local host task")
}

func productionAdapterReadyObjectiveCloseoutWriterInvocationReviewAndFixture(t *testing.T) (ProductionAdapterObjectiveCloseoutWriterInvocationReview, ProductionAdapterObjectiveCloseoutWriterInvocationReviewBlackboxFixture) {
	t.Helper()
	request := productionAdapterReadyObjectiveCloseoutWriterDurableRequest()
	packet := productionAdapterReadyObjectiveCloseoutWriterDurableWriteReviewPacket(request)
	envelope := BuildProductionAdapterObjectiveCloseoutWriterInvocationEnvelope(productionAdapterObjectiveCloseoutWriterInvocationEnvelopeInput(packet, "run:metrics_objective_closeout_writer_durable"))
	review := BuildProductionAdapterObjectiveCloseoutWriterInvocationReview(ProductionAdapterObjectiveCloseoutWriterInvocationReviewInput{
		ReviewRef:  "review:metrics_objective_closeout_writer_invocation",
		Invocation: envelope,
	})
	fixture := BuildProductionAdapterObjectiveCloseoutWriterInvocationReviewBlackboxFixture(ProductionAdapterObjectiveCloseoutWriterInvocationReviewBlackboxFixtureInput{
		FixtureRef: "fixture:metrics_objective_closeout_writer_invocation",
		Review:     review,
	})
	if review.Status != "ready_for_objective_closeout_writer_invocation_review" ||
		fixture.Status != "ready_for_objective_closeout_writer_invocation_display" {
		t.Fatalf("test helper expected ready invocation review/fixture, got review=%#v fixture=%#v", review, fixture)
	}
	return review, fixture
}
