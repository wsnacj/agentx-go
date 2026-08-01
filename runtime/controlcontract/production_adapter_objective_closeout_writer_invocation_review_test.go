package controlcontract

import (
	"encoding/json"
	"testing"
)

func TestProductionAdapterObjectiveCloseoutWriterInvocationReviewReadyForInvocationDisplay(t *testing.T) {
	request := productionAdapterReadyObjectiveCloseoutWriterDurableRequest()
	packet := productionAdapterReadyObjectiveCloseoutWriterDurableWriteReviewPacket(request)
	envelope := BuildProductionAdapterObjectiveCloseoutWriterInvocationEnvelope(productionAdapterObjectiveCloseoutWriterInvocationEnvelopeInput(packet, "run:metrics_objective_closeout_writer_durable"))
	review := BuildProductionAdapterObjectiveCloseoutWriterInvocationReview(ProductionAdapterObjectiveCloseoutWriterInvocationReviewInput{
		ReviewRef:  "review:metrics_objective_closeout_writer_invocation",
		Invocation: envelope,
	})
	if review.Status != "ready_for_objective_closeout_writer_invocation_review" ||
		review.DisplayState != "host_adapter_invocation_ready" ||
		review.DisplayStage != "invocation_review" ||
		!review.ReadyForHostDisplay ||
		!review.ReadyForHostAdapterInvocation ||
		!review.HostMayInvokeWriterAdapter ||
		!review.HostMayExecuteDurableWrite ||
		!review.HostAdapterInvocationAuthorized ||
		review.HostAdapterInvocationBound ||
		review.CoreInvocationExecuted ||
		review.DryRunByCore ||
		review.DurableWriteByCore ||
		review.ObjectiveStoreWriteByCore ||
		review.RunstoreWriteByCore {
		t.Fatalf("unexpected writer invocation review: %#v", review)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "objective closeout writer invocation review",
		RunnerEffect: review.RunnerEffect,
		PromptEffect: review.PromptEffect,
		Boundaries:   review.Boundaries,
		Payload:      review,
	}, "production_adapter_objective_closeout_writer_invocation_review", "objective_closeout_writer_invocation_review_projection_only", "host_cli_objective_closeout_writer_invocation_display", "no_adapter_invocation", "no_durable_write_by_core")

	fixture := BuildProductionAdapterObjectiveCloseoutWriterInvocationReviewBlackboxFixture(ProductionAdapterObjectiveCloseoutWriterInvocationReviewBlackboxFixtureInput{
		FixtureRef: "fixture:metrics_objective_closeout_writer_invocation",
		Review:     review,
	})
	if fixture.Status != "ready_for_objective_closeout_writer_invocation_display" ||
		fixture.DisplayState != "host_adapter_invocation_ready" ||
		!fixture.ReadyForHostDisplay ||
		!fixture.ReadyForHostAdapterInvocation ||
		!fixture.HostMayInvokeWriterAdapter ||
		!fixture.HostMayExecuteDurableWrite ||
		fixture.HostAdapterInvocationBound ||
		fixture.DurableWriteByCore {
		t.Fatalf("unexpected writer invocation fixture: %#v", fixture)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "objective closeout writer invocation review fixture",
		RunnerEffect: fixture.RunnerEffect,
		PromptEffect: fixture.PromptEffect,
		Boundaries:   fixture.Boundaries,
		Payload:      fixture,
	}, "production_adapter_objective_closeout_writer_invocation_review_blackbox_fixture", "objective_closeout_writer_invocation_review_fixture_projection_only", "host_cli_objective_closeout_writer_invocation_display", "no_adapter_invocation", "no_durable_write_by_core")
}

func TestProductionAdapterObjectiveCloseoutWriterInvocationReviewResultDisplays(t *testing.T) {
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
	if readbackReview.Status != "ready_for_objective_closeout_writer_invocation_result_readback_review" ||
		readbackReview.DisplayState != "invocation_result_readback_bound" ||
		readbackReview.DisplayStage != "result_readback_review" ||
		!readbackReview.ReadyForDurableReadbackReview ||
		!readbackReview.HostAdapterInvocationBound ||
		readbackReview.HostMayInvokeWriterAdapter ||
		readbackReview.HostMayExecuteDurableWrite ||
		readbackReview.ReadyForFailureReview {
		t.Fatalf("unexpected writer invocation readback review: %#v", readbackReview)
	}
	readbackFixture := BuildProductionAdapterObjectiveCloseoutWriterInvocationReviewBlackboxFixture(ProductionAdapterObjectiveCloseoutWriterInvocationReviewBlackboxFixtureInput{
		FixtureRef: "fixture:metrics_objective_closeout_writer_invocation_result",
		Review:     readbackReview,
	})
	if readbackFixture.Status != "ready_for_objective_closeout_writer_invocation_result_readback_display" ||
		!readbackFixture.ReadyForDurableReadbackReview ||
		!readbackFixture.HostAdapterInvocationBound ||
		readbackFixture.HostMayInvokeWriterAdapter {
		t.Fatalf("unexpected writer invocation readback fixture: %#v", readbackFixture)
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
	if failureReview.Status != "ready_for_objective_closeout_writer_invocation_result_failure_review" ||
		failureReview.DisplayState != "invocation_result_failed" ||
		failureReview.DisplayStage != "failure_review" ||
		!failureReview.ReadyForFailureReview ||
		!failureReview.ReadyForCompensationReview ||
		!failureReview.HostAdapterInvocationBound ||
		failureReview.HostMayInvokeWriterAdapter ||
		failureReview.HostMayExecuteDurableWrite ||
		failureReview.ReadyForDurableReadbackReview {
		t.Fatalf("unexpected writer invocation failure review: %#v", failureReview)
	}
	failureFixture := BuildProductionAdapterObjectiveCloseoutWriterInvocationReviewBlackboxFixture(ProductionAdapterObjectiveCloseoutWriterInvocationReviewBlackboxFixtureInput{
		FixtureRef: "fixture:metrics_objective_closeout_writer_invocation_failure",
		Review:     failureReview,
	})
	if failureFixture.Status != "ready_for_objective_closeout_writer_invocation_result_failure_display" ||
		!failureFixture.ReadyForFailureReview ||
		!failureFixture.ReadyForCompensationReview ||
		failureFixture.HostMayExecuteDurableWrite {
		t.Fatalf("unexpected writer invocation failure fixture: %#v", failureFixture)
	}
}

func TestProductionAdapterObjectiveCloseoutWriterInvocationReviewBlockedAndUnsafe(t *testing.T) {
	request := productionAdapterReadyObjectiveCloseoutWriterDurableRequest()
	packet := productionAdapterReadyObjectiveCloseoutWriterDurableWriteReviewPacket(request)
	envelope := BuildProductionAdapterObjectiveCloseoutWriterInvocationEnvelope(productionAdapterObjectiveCloseoutWriterInvocationEnvelopeInput(packet, "run:metrics_objective_closeout_writer_durable"))
	readbackMismatchResult := productionAdapterReadyObjectiveCloseoutWriterDurableResult(request)
	readbackMismatchResult.ExpectedReadbackRef = "readback:wrong_metrics_objective_closeout_writer"
	mismatchEnvelope := BuildProductionAdapterObjectiveCloseoutWriterInvocationResultEnvelope(ProductionAdapterObjectiveCloseoutWriterInvocationResultEnvelopeInput{
		ResultEnvelopeRef:  "result_envelope:metrics_objective_closeout_writer",
		InvocationEnvelope: envelope,
		DurableResult:      readbackMismatchResult,
	})
	mismatchReview := BuildProductionAdapterObjectiveCloseoutWriterInvocationReview(ProductionAdapterObjectiveCloseoutWriterInvocationReviewInput{
		ReviewRef:        "review:metrics_objective_closeout_writer_invocation_mismatch",
		Invocation:       envelope,
		InvocationResult: mismatchEnvelope,
	})
	if mismatchReview.Status != "ready_for_objective_closeout_writer_invocation_blocked_review" ||
		mismatchReview.DisplayState != "blocked_result_readback_mismatch" ||
		!mismatchReview.ReadyForBlockedReview ||
		mismatchReview.HostMayInvokeWriterAdapter ||
		mismatchReview.HostMayExecuteDurableWrite ||
		mismatchReview.HostAdapterInvocationBound ||
		!productionAdapterStringContains(mismatchReview.BlockedReasons, "writer_invocation_result_expected_readback_ref_mismatch") {
		t.Fatalf("expected writer invocation mismatch review to block, got %#v", mismatchReview)
	}
	mismatchFixture := BuildProductionAdapterObjectiveCloseoutWriterInvocationReviewBlackboxFixture(ProductionAdapterObjectiveCloseoutWriterInvocationReviewBlackboxFixtureInput{
		FixtureRef: "fixture:metrics_objective_closeout_writer_invocation_mismatch",
		Review:     mismatchReview,
	})
	if mismatchFixture.Status != "ready_for_objective_closeout_writer_invocation_blocked_display" ||
		mismatchFixture.DisplayState != "blocked_result_readback_mismatch" ||
		!mismatchFixture.ReadyForBlockedReview {
		t.Fatalf("expected writer invocation mismatch fixture to block, got %#v", mismatchFixture)
	}

	unsafe := BuildProductionAdapterObjectiveCloseoutWriterInvocationReview(ProductionAdapterObjectiveCloseoutWriterInvocationReviewInput{
		ReviewRef:  "/tmp/raw-writer-invocation-review.json",
		Invocation: envelope,
	})
	if unsafe.Status != "ready_for_objective_closeout_writer_invocation_blocked_review" ||
		unsafe.ReadyForHostDisplay ||
		unsafe.HostMayInvokeWriterAdapter ||
		unsafe.FailureClass != FailureEvidenceWeak ||
		!productionAdapterStringContains(unsafe.BlockedReasons, "unsafe_input_ref") ||
		!productionAdapterMissingContains(unsafe.MissingInputs, "host:display_safe_refs") {
		t.Fatalf("expected unsafe writer invocation review to block, got %#v", unsafe)
	}
	AssertNoRawPayload(t, "unsafe writer invocation review", unsafe, "/tmp/raw-writer-invocation-review.json")
}

func TestProductionAdapterObjectiveCloseoutWriterInvocationReviewJSONCompatibility(t *testing.T) {
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
	raw, err := json.Marshal(struct {
		Review  ProductionAdapterObjectiveCloseoutWriterInvocationReview                `json:"review"`
		Fixture ProductionAdapterObjectiveCloseoutWriterInvocationReviewBlackboxFixture `json:"fixture"`
	}{Review: review, Fixture: fixture})
	if err != nil {
		t.Fatalf("marshal writer invocation review contracts: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("unmarshal writer invocation review contracts: %v", err)
	}
	for _, token := range []string{
		"ready_for_objective_closeout_writer_invocation_review",
		"ready_for_objective_closeout_writer_invocation_display",
		"host_cli_objective_closeout_writer_invocation_display",
		"host_adapter_version_ref",
		"capability_proof_refs",
		"approval_binding_refs",
	} {
		if !jsonPayloadContains(raw, token) {
			t.Fatalf("expected writer invocation review JSON token %q in %s", token, raw)
		}
	}
	AssertNoRawPayload(t, "objective closeout writer invocation review JSON", raw, "/Users/mason", "postgresql://secret", "raw local host task")
}
