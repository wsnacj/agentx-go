package controlcontract

import (
	"encoding/json"
	"testing"
)

func TestProductionAdapterObjectiveCloseoutWriterInvocationEnvelopeReady(t *testing.T) {
	request := productionAdapterReadyObjectiveCloseoutWriterDurableRequest()
	packet := productionAdapterReadyObjectiveCloseoutWriterDurableWriteReviewPacket(request)
	envelope := BuildProductionAdapterObjectiveCloseoutWriterInvocationEnvelope(productionAdapterObjectiveCloseoutWriterInvocationEnvelopeInput(packet, "run:metrics_objective_closeout_writer_durable"))
	if envelope.Status != "ready_for_objective_closeout_writer_host_adapter_invocation" ||
		!envelope.ReadyForHostDisplay ||
		!envelope.ReadyForHostAdapterInvocation ||
		!envelope.HostAdapterInvocationAuthorized ||
		!envelope.HostMayInvokeWriterAdapter ||
		!envelope.HostMayExecuteDurableWrite ||
		envelope.NextHostAction != "host_may_invoke_objective_closeout_writer_adapter" ||
		envelope.ReviewPacketRef != packet.ReviewPacketRef ||
		envelope.DurableRequestRef != request.DurableRequestRef ||
		envelope.IdempotencyKeyRef != packet.IdempotencyRef ||
		envelope.DryRunProofRef != packet.DryRunSmokeRef ||
		envelope.HostDurableWriteConfirmationRef != packet.HostDurableWriteConfirmationRef ||
		envelope.ExpectedDurableResultRef != packet.ExpectedDurableResultRef ||
		envelope.ExpectedReadbackRef != packet.ExpectedReadbackRef ||
		envelope.ExpectedFailureRef == "" ||
		envelope.ExpectedCompensationRef == "" ||
		len(envelope.CapabilityProofRefs) == 0 ||
		len(envelope.ApprovalBindingRefs) == 0 ||
		envelope.CoreInvocationExecuted ||
		envelope.DryRunByCore ||
		envelope.DurableWriteByCore ||
		envelope.ObjectiveStoreWriteByCore ||
		envelope.RunstoreWriteByCore {
		t.Fatalf("unexpected writer invocation envelope: %#v", envelope)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "objective closeout writer invocation envelope",
		RunnerEffect: envelope.RunnerEffect,
		PromptEffect: envelope.PromptEffect,
		Boundaries:   envelope.Boundaries,
		Payload:      envelope,
	}, "production_adapter_objective_closeout_writer_invocation_envelope", "objective_closeout_writer_invocation_envelope_projection_only", "host_writer_adapter_invocation_authorization", "host_may_invoke_writer_adapter", "no_adapter_invocation", "no_durable_write_by_core")
	AssertNoCoreMutation(t, "objective closeout writer invocation envelope", envelope.CoreInvocationExecuted, envelope.DurableWriteByCore)
}

func TestProductionAdapterObjectiveCloseoutWriterInvocationEnvelopeBlocksMissingProofAndMismatch(t *testing.T) {
	request := productionAdapterReadyObjectiveCloseoutWriterDurableRequest()
	packet := productionAdapterReadyObjectiveCloseoutWriterDurableWriteReviewPacket(request)
	missingInput := productionAdapterObjectiveCloseoutWriterInvocationEnvelopeInput(packet, "run:metrics_objective_closeout_writer_durable")
	missingInput.HostAdapterVersionRef = ""
	missingInput.CapabilityProofRefs = nil
	missing := BuildProductionAdapterObjectiveCloseoutWriterInvocationEnvelope(missingInput)
	if missing.Status != "blocked" ||
		missing.ReadyForHostAdapterInvocation ||
		missing.HostMayInvokeWriterAdapter ||
		missing.HostMayExecuteDurableWrite ||
		missing.FailureClass != FailureConfigMissing ||
		!productionAdapterStringContains(missing.BlockedReasons, "host_adapter_version_ref_missing") ||
		!productionAdapterStringContains(missing.BlockedReasons, "capability_proof_ref_missing") ||
		!productionAdapterMissingContains(missing.MissingInputs, "host:objective_closeout_writer_adapter_version_ref") ||
		!productionAdapterMissingContains(missing.MissingInputs, "host:objective_closeout_writer_capability_proof_ref") {
		t.Fatalf("expected missing invocation envelope proofs to block, got %#v", missing)
	}

	mismatchInput := productionAdapterObjectiveCloseoutWriterInvocationEnvelopeInput(packet, "run:metrics_objective_closeout_writer_durable")
	mismatchInput.DryRunProofRef = "smoke:wrong_metrics_objective_closeout_writer"
	mismatch := BuildProductionAdapterObjectiveCloseoutWriterInvocationEnvelope(mismatchInput)
	if mismatch.Status != "blocked" ||
		mismatch.ReadyForHostAdapterInvocation ||
		mismatch.FailureClass != FailureVerificationFailed ||
		!productionAdapterStringContains(mismatch.BlockedReasons, "writer_invocation_dry_run_proof_ref_mismatch") ||
		!productionAdapterMissingContains(mismatch.MissingInputs, "host:objective_closeout_writer_dry_run_proof_ref") {
		t.Fatalf("expected dry-run proof mismatch to block, got %#v", mismatch)
	}
}

func TestProductionAdapterObjectiveCloseoutWriterInvocationResultEnvelopeReadyForReadbackReview(t *testing.T) {
	request := productionAdapterReadyObjectiveCloseoutWriterDurableRequest()
	packet := productionAdapterReadyObjectiveCloseoutWriterDurableWriteReviewPacket(request)
	envelope := BuildProductionAdapterObjectiveCloseoutWriterInvocationEnvelope(productionAdapterObjectiveCloseoutWriterInvocationEnvelopeInput(packet, "run:metrics_objective_closeout_writer_durable"))
	durableResult := productionAdapterReadyObjectiveCloseoutWriterDurableResult(request)
	resultEnvelope := BuildProductionAdapterObjectiveCloseoutWriterInvocationResultEnvelope(ProductionAdapterObjectiveCloseoutWriterInvocationResultEnvelopeInput{
		ResultEnvelopeRef:  "result_envelope:metrics_objective_closeout_writer",
		InvocationEnvelope: envelope,
		DurableResult:      durableResult,
	})
	if resultEnvelope.Status != "ready_for_objective_closeout_writer_invocation_result_readback_review" ||
		!resultEnvelope.ReadyForHostDisplay ||
		!resultEnvelope.ReadyForDurableReadbackReview ||
		resultEnvelope.ReadyForFailureReview ||
		!resultEnvelope.HostAdapterInvocationBound ||
		resultEnvelope.NextHostAction != "bind_objective_closeout_writer_durable_readback" ||
		resultEnvelope.HostAdapterRunRef != envelope.ExpectedHostAdapterRunRef ||
		resultEnvelope.DurableResultRef != envelope.ExpectedDurableResultRef ||
		resultEnvelope.AppliedDurableEventRef != envelope.ExpectedDurableEventRef ||
		resultEnvelope.AppliedRunstoreRef != envelope.HostRunstoreRef ||
		resultEnvelope.AppliedObjectiveStateRef != envelope.ExpectedObjectiveStateRef ||
		resultEnvelope.CoreInvocationExecuted ||
		resultEnvelope.DryRunByCore ||
		resultEnvelope.DurableWriteByCore ||
		resultEnvelope.ObjectiveStoreWriteByCore ||
		resultEnvelope.RunstoreWriteByCore {
		t.Fatalf("unexpected writer invocation result envelope: %#v", resultEnvelope)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "objective closeout writer invocation result envelope",
		RunnerEffect: resultEnvelope.RunnerEffect,
		PromptEffect: resultEnvelope.PromptEffect,
		Boundaries:   resultEnvelope.Boundaries,
		Payload:      resultEnvelope,
	}, "production_adapter_objective_closeout_writer_invocation_result_envelope", "objective_closeout_writer_invocation_result_envelope_projection_only", "host_writer_adapter_invocation_result_binding", "ready_for_objective_closeout_writer_invocation_result_readback_review", "no_adapter_invocation", "no_durable_write_by_core")
	AssertNoCoreMutation(t, "objective closeout writer invocation result envelope", resultEnvelope.CoreInvocationExecuted, resultEnvelope.DurableWriteByCore)
}

func TestProductionAdapterObjectiveCloseoutWriterInvocationResultEnvelopeReadyForFailureReview(t *testing.T) {
	request := productionAdapterReadyObjectiveCloseoutWriterDurableRequest()
	packet := productionAdapterReadyObjectiveCloseoutWriterDurableWriteReviewPacket(request)
	envelope := BuildProductionAdapterObjectiveCloseoutWriterInvocationEnvelope(productionAdapterObjectiveCloseoutWriterInvocationEnvelopeInput(packet, "run:metrics_objective_closeout_writer_durable_failed"))
	durableResult := productionAdapterFailedObjectiveCloseoutWriterDurableResult(request)
	resultEnvelope := BuildProductionAdapterObjectiveCloseoutWriterInvocationResultEnvelope(ProductionAdapterObjectiveCloseoutWriterInvocationResultEnvelopeInput{
		ResultEnvelopeRef:  "result_envelope:metrics_objective_closeout_writer_failure",
		InvocationEnvelope: envelope,
		DurableResult:      durableResult,
	})
	if resultEnvelope.Status != "ready_for_objective_closeout_writer_invocation_result_failure_review" ||
		!resultEnvelope.ReadyForHostDisplay ||
		resultEnvelope.ReadyForDurableReadbackReview ||
		!resultEnvelope.ReadyForFailureReview ||
		!resultEnvelope.ReadyForCompensationReview ||
		!resultEnvelope.HostAdapterInvocationBound ||
		resultEnvelope.FailureClass != FailureVerificationFailed ||
		resultEnvelope.NextHostAction != "review_objective_closeout_writer_durable_failure" ||
		resultEnvelope.HostAdapterRunRef != envelope.ExpectedHostAdapterRunRef ||
		resultEnvelope.FailureRef != envelope.ExpectedFailureRef ||
		resultEnvelope.CompensationRef != envelope.ExpectedCompensationRef ||
		resultEnvelope.DurableWriteByCore ||
		resultEnvelope.ObjectiveStoreWriteByCore ||
		resultEnvelope.RunstoreWriteByCore {
		t.Fatalf("unexpected writer invocation failure result envelope: %#v", resultEnvelope)
	}
}

func TestProductionAdapterObjectiveCloseoutWriterInvocationResultEnvelopeBlocksMismatchAndUnsafe(t *testing.T) {
	request := productionAdapterReadyObjectiveCloseoutWriterDurableRequest()
	packet := productionAdapterReadyObjectiveCloseoutWriterDurableWriteReviewPacket(request)
	envelope := BuildProductionAdapterObjectiveCloseoutWriterInvocationEnvelope(productionAdapterObjectiveCloseoutWriterInvocationEnvelopeInput(packet, "run:metrics_objective_closeout_writer_durable"))
	durableResult := productionAdapterReadyObjectiveCloseoutWriterDurableResult(request)
	durableResult.HostAdapterRunRef = "run:wrong_metrics_objective_closeout_writer_durable"
	mismatch := BuildProductionAdapterObjectiveCloseoutWriterInvocationResultEnvelope(ProductionAdapterObjectiveCloseoutWriterInvocationResultEnvelopeInput{
		ResultEnvelopeRef:  "result_envelope:metrics_objective_closeout_writer",
		InvocationEnvelope: envelope,
		DurableResult:      durableResult,
	})
	if mismatch.Status != "blocked" ||
		mismatch.ReadyForDurableReadbackReview ||
		mismatch.HostAdapterInvocationBound ||
		mismatch.FailureClass != FailureVerificationFailed ||
		!productionAdapterStringContains(mismatch.BlockedReasons, "writer_invocation_result_adapter_run_ref_mismatch") ||
		!productionAdapterMissingContains(mismatch.MissingInputs, "host:objective_closeout_writer_host_adapter_run_ref") {
		t.Fatalf("expected writer invocation result mismatch to block, got %#v", mismatch)
	}

	readbackMismatchResult := productionAdapterReadyObjectiveCloseoutWriterDurableResult(request)
	readbackMismatchResult.ExpectedReadbackRef = "readback:wrong_metrics_objective_closeout_writer"
	readbackMismatch := BuildProductionAdapterObjectiveCloseoutWriterInvocationResultEnvelope(ProductionAdapterObjectiveCloseoutWriterInvocationResultEnvelopeInput{
		ResultEnvelopeRef:  "result_envelope:metrics_objective_closeout_writer",
		InvocationEnvelope: envelope,
		DurableResult:      readbackMismatchResult,
	})
	if readbackMismatch.Status != "blocked" ||
		readbackMismatch.ReadyForDurableReadbackReview ||
		readbackMismatch.HostAdapterInvocationBound ||
		readbackMismatch.FailureClass != FailureVerificationFailed ||
		!productionAdapterStringContains(readbackMismatch.BlockedReasons, "writer_invocation_result_expected_readback_ref_mismatch") ||
		!productionAdapterMissingContains(readbackMismatch.MissingInputs, "host:objective_closeout_writer_expected_readback_ref") {
		t.Fatalf("expected writer invocation result readback mismatch to block, got %#v", readbackMismatch)
	}

	missingReadbackResult := productionAdapterReadyObjectiveCloseoutWriterDurableResult(request)
	missingReadbackResult.ExpectedReadbackRef = ""
	missingReadback := BuildProductionAdapterObjectiveCloseoutWriterInvocationResultEnvelope(ProductionAdapterObjectiveCloseoutWriterInvocationResultEnvelopeInput{
		ResultEnvelopeRef:  "result_envelope:metrics_objective_closeout_writer",
		InvocationEnvelope: envelope,
		DurableResult:      missingReadbackResult,
	})
	if missingReadback.Status != "blocked" ||
		missingReadback.ReadyForDurableReadbackReview ||
		missingReadback.HostAdapterInvocationBound ||
		missingReadback.FailureClass != FailureEvidenceMissing ||
		!productionAdapterStringContains(missingReadback.BlockedReasons, "writer_invocation_result_expected_readback_ref_missing") ||
		!productionAdapterMissingContains(missingReadback.MissingInputs, "host:objective_closeout_writer_expected_readback_ref") {
		t.Fatalf("expected writer invocation result missing readback to block, got %#v", missingReadback)
	}

	unsafe := BuildProductionAdapterObjectiveCloseoutWriterInvocationResultEnvelope(ProductionAdapterObjectiveCloseoutWriterInvocationResultEnvelopeInput{
		ResultEnvelopeRef:  "/tmp/raw-writer-result-envelope.json",
		InvocationEnvelope: envelope,
		DurableResult:      productionAdapterReadyObjectiveCloseoutWriterDurableResult(request),
	})
	if unsafe.Status != "blocked" ||
		unsafe.ReadyForHostDisplay ||
		unsafe.FailureClass != FailureEvidenceWeak ||
		!productionAdapterStringContains(unsafe.BlockedReasons, "unsafe_input_ref") ||
		!productionAdapterMissingContains(unsafe.MissingInputs, "host:display_safe_refs") {
		t.Fatalf("expected unsafe writer invocation result envelope to block, got %#v", unsafe)
	}
	AssertNoRawPayload(t, "unsafe writer invocation result envelope", unsafe, "/tmp/raw-writer-result-envelope.json")
}

func TestProductionAdapterObjectiveCloseoutWriterInvocationEnvelopeJSONCompatibility(t *testing.T) {
	request := productionAdapterReadyObjectiveCloseoutWriterDurableRequest()
	packet := productionAdapterReadyObjectiveCloseoutWriterDurableWriteReviewPacket(request)
	envelope := BuildProductionAdapterObjectiveCloseoutWriterInvocationEnvelope(productionAdapterObjectiveCloseoutWriterInvocationEnvelopeInput(packet, "run:metrics_objective_closeout_writer_durable"))
	resultEnvelope := BuildProductionAdapterObjectiveCloseoutWriterInvocationResultEnvelope(ProductionAdapterObjectiveCloseoutWriterInvocationResultEnvelopeInput{
		ResultEnvelopeRef:  "result_envelope:metrics_objective_closeout_writer",
		InvocationEnvelope: envelope,
		DurableResult:      productionAdapterReadyObjectiveCloseoutWriterDurableResult(request),
	})
	raw, err := json.Marshal(struct {
		Envelope       ProductionAdapterObjectiveCloseoutWriterInvocationEnvelope       `json:"envelope"`
		ResultEnvelope ProductionAdapterObjectiveCloseoutWriterInvocationResultEnvelope `json:"result_envelope"`
	}{Envelope: envelope, ResultEnvelope: resultEnvelope})
	if err != nil {
		t.Fatalf("marshal writer invocation envelope contracts: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("unmarshal writer invocation envelope contracts: %v", err)
	}
	for _, token := range []string{
		"ready_for_objective_closeout_writer_host_adapter_invocation",
		"ready_for_objective_closeout_writer_invocation_result_readback_review",
		"host_adapter_version_ref",
		"capability_proof_refs",
		"approval_binding_refs",
		"dry_run_proof_ref",
		"objective_closeout_writer_invocation_result_envelope_projection_only",
	} {
		if !jsonPayloadContains(raw, token) {
			t.Fatalf("expected writer invocation envelope JSON token %q in %s", token, raw)
		}
	}
	AssertNoRawPayload(t, "objective closeout writer invocation envelope JSON", raw, "/Users/mason", "postgresql://secret", "raw local host task")
}

func productionAdapterReadyObjectiveCloseoutWriterDurableWriteReviewPacket(request ProductionAdapterObjectiveCloseoutWriterDurableRequest) ProductionAdapterObjectiveCloseoutWriterDurableReviewPacket {
	return BuildProductionAdapterObjectiveCloseoutWriterDurableReviewPacket(ProductionAdapterObjectiveCloseoutWriterDurableReviewPacketInput{
		ReviewPacketRef: "review:metrics_objective_closeout_writer_durable",
		DurableRequest:  request,
	})
}

func productionAdapterObjectiveCloseoutWriterInvocationEnvelopeInput(packet ProductionAdapterObjectiveCloseoutWriterDurableReviewPacket, expectedRunRef DisplaySafeRef) ProductionAdapterObjectiveCloseoutWriterInvocationEnvelopeInput {
	return ProductionAdapterObjectiveCloseoutWriterInvocationEnvelopeInput{
		InvocationEnvelopeRef:           "invocation_envelope:metrics_objective_closeout_writer",
		DurableReviewPacket:             packet,
		WriterInvocationRef:             "invocation:metrics_objective_closeout_writer",
		HostAdapterVersionRef:           "version:metrics_objective_closeout_writer_v1",
		CapabilityProofRefs:             []DisplaySafeRef{"capability_proof:metrics_objective_closeout_writer"},
		ApprovalBindingRefs:             []DisplaySafeRef{"approval_binding:metrics_objective_closeout_writer"},
		IdempotencyKeyRef:               packet.IdempotencyRef,
		DryRunProofRef:                  packet.DryRunSmokeRef,
		HostDurableWriteConfirmationRef: packet.HostDurableWriteConfirmationRef,
		ExpectedHostAdapterRunRef:       expectedRunRef,
		ExpectedDurableResultRef:        packet.ExpectedDurableResultRef,
		ExpectedReadbackRef:             packet.ExpectedReadbackRef,
		ExpectedFailureRef:              "failure:metrics_objective_closeout_writer_durable",
		ExpectedCompensationRef:         "compensation:metrics_objective_closeout_writer_review",
		TimeoutPolicyRef:                "timeout:metrics_objective_closeout_writer",
	}
}
