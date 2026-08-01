package controlcontract

import (
	"encoding/json"
	"testing"
)

func TestProductionAdapterObjectiveCloseoutWriterDurableRequestReady(t *testing.T) {
	request := productionAdapterReadyObjectiveCloseoutWriterDurableRequest()
	if request.Status != HostActionReady ||
		!request.ReadyForHostDurableWrite ||
		!request.HostDurableWriteAuthorized ||
		!request.HostMayExecuteDurableWrite ||
		request.RequestedMode != ProductionAdapterObjectiveCloseoutWriterDurableWrite ||
		request.DurableRequestRef == "" ||
		request.HostDurableWriteConfirmationRef == "" ||
		request.ExpectedDurableResultRef == "" ||
		request.DryRunSmokeRef == "" ||
		request.DryRunResultRef == "" ||
		request.CoreInvocationExecuted ||
		request.DryRunByCore ||
		request.DurableWriteByCore ||
		request.ObjectiveStoreWriteByCore ||
		request.RunstoreWriteByCore ||
		request.NextHostAction != "host_may_execute_objective_closeout_durable_writer_adapter" {
		t.Fatalf("unexpected objective closeout writer durable request: %#v", request)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "objective closeout writer durable request",
		RunnerEffect: request.RunnerEffect,
		PromptEffect: request.PromptEffect,
		Boundaries:   request.Boundaries,
		Payload:      request,
	}, "production_adapter_objective_closeout_writer_durable_request", "objective_closeout_writer_durable_request_projection_only", "host_owned_objective_closeout_writer_adapter", "dry_run_smoke_required", "ready_for_host_objective_closeout_writer_durable_write", "host_may_execute_durable_writer", "core_durable_write_not_executed")
}

func TestProductionAdapterObjectiveCloseoutWriterDurableResultRecorded(t *testing.T) {
	request := productionAdapterReadyObjectiveCloseoutWriterDurableRequest()
	adapter := productionAdapterObjectiveCloseoutWriterFakeAdapter{descriptor: productionAdapterReadyObjectiveCloseoutWriterDescriptor()}
	var _ ProductionAdapterObjectiveCloseoutWriterHostAdapter = adapter

	result := adapter.ExecuteObjectiveCloseoutDurableWriter(request)
	if result.Status != HostActionRecorded ||
		!result.HostDurableWriteReported ||
		!result.HostDurableWriteSucceeded ||
		result.HostDurableWriteFailed ||
		!result.HostDurableWriteRecorded ||
		!result.ReadyForWriterDurableReadback ||
		result.DurableResultRef != request.ExpectedDurableResultRef ||
		result.AppliedDurableEventRef != request.ExpectedDurableEventRef ||
		result.AppliedRunstoreRef != request.HostRunstoreRef ||
		result.AppliedObjectiveStateRef != request.ExpectedObjectiveStateRef ||
		result.CoreInvocationExecuted ||
		result.DryRunByCore ||
		result.DurableWriteByCore ||
		result.ObjectiveStoreWriteByCore ||
		result.RunstoreWriteByCore ||
		result.NextHostAction != "bind_objective_closeout_writer_durable_readback" {
		t.Fatalf("unexpected objective closeout writer durable result: %#v", result)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "objective closeout writer durable result",
		RunnerEffect: result.RunnerEffect,
		PromptEffect: result.PromptEffect,
		Boundaries:   result.Boundaries,
		Payload:      result,
	}, "production_adapter_objective_closeout_writer_durable_result", "objective_closeout_writer_durable_result_projection_only", "host_owned_objective_closeout_writer_adapter", "host_objective_closeout_writer_durable_write_recorded", "ready_for_objective_closeout_writer_durable_readback")
}

func TestProductionAdapterObjectiveCloseoutWriterDurableReadbackBound(t *testing.T) {
	request := productionAdapterReadyObjectiveCloseoutWriterDurableRequest()
	result := productionAdapterReadyObjectiveCloseoutWriterDurableResult(request)
	closeoutReadback := productionAdapterReadyObjectiveCloseoutWriterCloseoutReadback(result)
	readback := BuildProductionAdapterObjectiveCloseoutWriterDurableReadback(ProductionAdapterObjectiveCloseoutWriterDurableReadbackInput{
		DurableReadbackRef:        "readback_binding:metrics_objective_closeout_writer",
		DurableResult:             result,
		ObjectiveCloseoutReadback: closeoutReadback,
	})
	if readback.Status != HostActionRecorded ||
		!readback.WriterDurableReadbackBound ||
		!readback.ReadyForObjectiveReturn ||
		!readback.ObjectiveLifecycleClosed ||
		!readback.ObjectiveSatisfied ||
		readback.DurableResultRef != result.DurableResultRef ||
		readback.ObjectiveCloseoutReadbackRef != result.ExpectedReadbackRef ||
		readback.AppliedDurableEventRef != result.AppliedDurableEventRef ||
		readback.AppliedRunstoreRef != result.AppliedRunstoreRef ||
		readback.AppliedObjectiveStateRef != result.AppliedObjectiveStateRef ||
		readback.CoreInvocationExecuted ||
		readback.DryRunByCore ||
		readback.DurableWriteByCore ||
		readback.ObjectiveStoreWriteByCore ||
		readback.RunstoreWriteByCore ||
		readback.NextHostAction != "return_objective_closed_lifecycle" {
		t.Fatalf("unexpected objective closeout writer durable readback: %#v", readback)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "objective closeout writer durable readback",
		RunnerEffect: readback.RunnerEffect,
		PromptEffect: readback.PromptEffect,
		Boundaries:   readback.Boundaries,
		Payload:      readback,
	}, "production_adapter_objective_closeout_writer_durable_readback", "objective_closeout_writer_durable_readback_projection_only", "host_owned_objective_closeout_writer_adapter", "authorization_bound_objective_closeout_readback", "objective_closeout_writer_durable_readback_bound", "ready_for_objective_return")
}

func TestProductionAdapterObjectiveCloseoutWriterDurableRequestBlocksMissingInputsAndSmokeMismatch(t *testing.T) {
	optIn, smoke := productionAdapterReadyObjectiveCloseoutWriterDurableOptInAndSmoke()
	tests := []struct {
		name        string
		input       ProductionAdapterObjectiveCloseoutWriterDurableRequestInput
		wantReason  string
		wantMissing MissingInput
		wantFailure FailureClass
	}{
		{
			name: "missing host durable confirmation",
			input: ProductionAdapterObjectiveCloseoutWriterDurableRequestInput{
				DurableRequestRef:        "durable_request:metrics_objective_closeout_writer",
				WriterOptIn:              optIn,
				DryRunSmoke:              smoke,
				ExpectedDurableResultRef: "durable_result:metrics_objective_closeout_writer",
			},
			wantReason:  "host_durable_write_confirmation_ref_missing",
			wantMissing: "host:objective_closeout_writer_durable_write_confirmation_ref",
			wantFailure: FailureAuthorizationMissing,
		},
		{
			name: "dry-run smoke mismatch",
			input: func() ProductionAdapterObjectiveCloseoutWriterDurableRequestInput {
				badSmoke := smoke
				badSmoke.DryRunResultRef = "dryrun_result:wrong_metrics_objective_closeout_writer"
				return ProductionAdapterObjectiveCloseoutWriterDurableRequestInput{
					DurableRequestRef:               "durable_request:metrics_objective_closeout_writer",
					WriterOptIn:                     optIn,
					DryRunSmoke:                     badSmoke,
					HostDurableWriteConfirmationRef: "confirmation:metrics_objective_closeout_writer_durable_write",
					ExpectedDurableResultRef:        "durable_result:metrics_objective_closeout_writer",
				}
			}(),
			wantReason:  "durable_request_dry_run_result_ref_mismatch",
			wantMissing: "host:objective_closeout_writer_dry_run_result_ref",
			wantFailure: FailureVerificationFailed,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			request := BuildProductionAdapterObjectiveCloseoutWriterDurableRequest(tc.input)
			if request.Status != HostActionBlocked ||
				request.ReadyForHostDurableWrite ||
				request.HostDurableWriteAuthorized ||
				request.HostMayExecuteDurableWrite ||
				request.FailureClass != tc.wantFailure ||
				!productionAdapterStringContains(request.BlockedReasons, tc.wantReason) ||
				!productionAdapterMissingContains(request.MissingInputs, tc.wantMissing) ||
				request.CoreInvocationExecuted ||
				request.DryRunByCore ||
				request.DurableWriteByCore {
				t.Fatalf("unexpected blocked durable request %s: %#v", tc.name, request)
			}
		})
	}
}

func TestProductionAdapterObjectiveCloseoutWriterDurableResultBlocksMismatchesAndFailure(t *testing.T) {
	request := productionAdapterReadyObjectiveCloseoutWriterDurableRequest()
	tests := []struct {
		name        string
		input       ProductionAdapterObjectiveCloseoutWriterDurableResultInput
		wantStatus  HostActionStatus
		wantReason  string
		wantMissing MissingInput
		wantFailure FailureClass
	}{
		{
			name: "not reported",
			input: ProductionAdapterObjectiveCloseoutWriterDurableResultInput{
				DurableResultRef:         request.ExpectedDurableResultRef,
				DurableRequest:           request,
				HostAdapterRunRef:        "run:metrics_objective_closeout_writer_durable",
				AppliedDurableEventRef:   request.ExpectedDurableEventRef,
				AppliedRunstoreRef:       request.HostRunstoreRef,
				AppliedObjectiveStateRef: request.ExpectedObjectiveStateRef,
			},
			wantStatus:  HostActionBlocked,
			wantReason:  "writer_durable_write_not_reported",
			wantMissing: "host:objective_closeout_writer_durable_report",
			wantFailure: FailureEvidenceMissing,
		},
		{
			name: "event ref mismatch",
			input: ProductionAdapterObjectiveCloseoutWriterDurableResultInput{
				DurableResultRef:          request.ExpectedDurableResultRef,
				DurableRequest:            request,
				HostAdapterRunRef:         "run:metrics_objective_closeout_writer_durable",
				HostDurableWriteReported:  true,
				HostDurableWriteSucceeded: true,
				AppliedDurableEventRef:    "event:wrong_metrics_objective_closed",
				AppliedRunstoreRef:        request.HostRunstoreRef,
				AppliedObjectiveStateRef:  request.ExpectedObjectiveStateRef,
			},
			wantStatus:  HostActionBlocked,
			wantReason:  "writer_durable_event_ref_mismatch",
			wantMissing: "host:durable_event_ref",
			wantFailure: FailureVerificationFailed,
		},
		{
			name: "failure reported",
			input: ProductionAdapterObjectiveCloseoutWriterDurableResultInput{
				DurableResultRef:         request.ExpectedDurableResultRef,
				DurableRequest:           request,
				HostAdapterRunRef:        "run:metrics_objective_closeout_writer_durable",
				HostDurableWriteReported: true,
				HostDurableWriteFailed:   true,
				FailureRef:               "failure:metrics_objective_closeout_writer_durable",
				CompensationRef:          "compensation:metrics_objective_closeout_writer_review",
			},
			wantStatus:  HostActionReviewRequired,
			wantReason:  "objective_closeout_writer_durable_write_failed",
			wantMissing: "",
			wantFailure: FailureVerificationFailed,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := BuildProductionAdapterObjectiveCloseoutWriterDurableResult(tc.input)
			if result.Status != tc.wantStatus ||
				result.ReadyForWriterDurableReadback ||
				(result.Status != HostActionReviewRequired && result.HostDurableWriteRecorded) ||
				result.FailureClass != tc.wantFailure ||
				!productionAdapterStringContains(result.BlockedReasons, tc.wantReason) ||
				(tc.wantMissing != "" && !productionAdapterMissingContains(result.MissingInputs, tc.wantMissing)) ||
				result.CoreInvocationExecuted ||
				result.DryRunByCore ||
				result.DurableWriteByCore {
				t.Fatalf("unexpected durable result %s: %#v", tc.name, result)
			}
		})
	}
}

func TestProductionAdapterObjectiveCloseoutWriterDurableReadbackBlocksMismatch(t *testing.T) {
	request := productionAdapterReadyObjectiveCloseoutWriterDurableRequest()
	result := productionAdapterReadyObjectiveCloseoutWriterDurableResult(request)
	closeoutReadback := productionAdapterReadyObjectiveCloseoutWriterCloseoutReadback(result)
	closeoutReadback.ObjectiveCloseoutReadbackRef = "readback:wrong_metrics_objective_closeout_writer"
	readback := BuildProductionAdapterObjectiveCloseoutWriterDurableReadback(ProductionAdapterObjectiveCloseoutWriterDurableReadbackInput{
		DurableReadbackRef:        "readback_binding:metrics_objective_closeout_writer",
		DurableResult:             result,
		ObjectiveCloseoutReadback: closeoutReadback,
	})
	if readback.Status != HostActionBlocked ||
		readback.WriterDurableReadbackBound ||
		readback.ReadyForObjectiveReturn ||
		readback.ObjectiveLifecycleClosed ||
		readback.ObjectiveSatisfied ||
		!productionAdapterStringContains(readback.BlockedReasons, "writer_readback_ref_mismatch") ||
		!productionAdapterMissingContains(readback.MissingInputs, "host:objective_closeout_readback_ref") ||
		readback.CoreInvocationExecuted ||
		readback.DryRunByCore ||
		readback.DurableWriteByCore {
		t.Fatalf("expected durable readback mismatch to block, got %#v", readback)
	}
}

func TestProductionAdapterObjectiveCloseoutWriterDurableContractsJSONCompatibility(t *testing.T) {
	request := productionAdapterReadyObjectiveCloseoutWriterDurableRequest()
	result := productionAdapterReadyObjectiveCloseoutWriterDurableResult(request)
	readback := BuildProductionAdapterObjectiveCloseoutWriterDurableReadback(ProductionAdapterObjectiveCloseoutWriterDurableReadbackInput{
		DurableReadbackRef:        "readback_binding:metrics_objective_closeout_writer",
		DurableResult:             result,
		ObjectiveCloseoutReadback: productionAdapterReadyObjectiveCloseoutWriterCloseoutReadback(result),
	})
	raw, err := json.Marshal(struct {
		Request  ProductionAdapterObjectiveCloseoutWriterDurableRequest  `json:"request"`
		Result   ProductionAdapterObjectiveCloseoutWriterDurableResult   `json:"result"`
		Readback ProductionAdapterObjectiveCloseoutWriterDurableReadback `json:"readback"`
	}{Request: request, Result: result, Readback: readback})
	if err != nil {
		t.Fatalf("marshal durable contracts: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("unmarshal durable contracts: %v", err)
	}
	for _, token := range []string{
		"durable_request_ref",
		"expected_durable_result_ref",
		"host_durable_write_confirmation_ref",
		"ready_for_writer_durable_readback",
		"writer_durable_readback_bound",
		"objective_closeout_writer_durable_write_only",
	} {
		if !jsonPayloadContains(raw, token) {
			t.Fatalf("expected durable JSON token %q in %s", token, raw)
		}
	}
	AssertNoRawPayload(t, "objective closeout writer durable contracts JSON", raw, "/Users/mason", "postgresql://secret", "raw local host task")
}

func productionAdapterReadyObjectiveCloseoutWriterDryRunSmoke() ProductionAdapterObjectiveCloseoutWriterDryRunSmokeHarness {
	request := productionAdapterReadyObjectiveCloseoutWriterDryRunRequest()
	result := productionAdapterReadyObjectiveCloseoutWriterDryRunResult(request)
	return BuildProductionAdapterObjectiveCloseoutWriterDryRunSmokeHarness(ProductionAdapterObjectiveCloseoutWriterDryRunSmokeHarnessInput{
		SmokeRef:      "smoke:metrics_objective_closeout_writer_dry_run",
		DryRunRequest: request,
		DryRunResult:  result,
	})
}

func productionAdapterReadyObjectiveCloseoutWriterDurableOptInAndSmoke() (ProductionAdapterObjectiveCloseoutWriterOptIn, ProductionAdapterObjectiveCloseoutWriterDryRunSmokeHarness) {
	smoke := productionAdapterReadyObjectiveCloseoutWriterDryRunSmoke()
	handoff, uiHandoff := productionAdapterReadyObjectiveCloseoutWriterInputs()
	descriptor := productionAdapterReadyObjectiveCloseoutWriterDescriptor()
	input := productionAdapterObjectiveCloseoutWriterOptInInputForMode(
		descriptor,
		handoff,
		uiHandoff,
		"optin:metrics_objective_closeout_writer_durable",
		ProductionAdapterObjectiveCloseoutWriterDurableWrite,
	)
	input.DryRunResultRef = smoke.DryRunResultRef
	input.ExpectedReadbackRef = smoke.ExpectedReadbackRef
	return BuildProductionAdapterObjectiveCloseoutWriterOptIn(input), smoke
}

func productionAdapterReadyObjectiveCloseoutWriterDurableRequest() ProductionAdapterObjectiveCloseoutWriterDurableRequest {
	optIn, smoke := productionAdapterReadyObjectiveCloseoutWriterDurableOptInAndSmoke()
	return BuildProductionAdapterObjectiveCloseoutWriterDurableRequest(ProductionAdapterObjectiveCloseoutWriterDurableRequestInput{
		DurableRequestRef:               "durable_request:metrics_objective_closeout_writer",
		WriterOptIn:                     optIn,
		DryRunSmoke:                     smoke,
		HostDurableWriteConfirmationRef: "confirmation:metrics_objective_closeout_writer_durable_write",
		ExpectedDurableResultRef:        "durable_result:metrics_objective_closeout_writer",
	})
}

func productionAdapterReadyObjectiveCloseoutWriterDurableResult(request ProductionAdapterObjectiveCloseoutWriterDurableRequest) ProductionAdapterObjectiveCloseoutWriterDurableResult {
	return BuildProductionAdapterObjectiveCloseoutWriterDurableResult(ProductionAdapterObjectiveCloseoutWriterDurableResultInput{
		DurableResultRef:          request.ExpectedDurableResultRef,
		DurableRequest:            request,
		HostAdapterRunRef:         "run:metrics_objective_closeout_writer_durable",
		HostDurableWriteReported:  true,
		HostDurableWriteSucceeded: true,
		AppliedDurableEventRef:    request.ExpectedDurableEventRef,
		AppliedRunstoreRef:        request.HostRunstoreRef,
		AppliedObjectiveStateRef:  request.ExpectedObjectiveStateRef,
		DurableEvidenceRefs:       []DisplaySafeRef{"evidence:metrics_objective_closeout_writer_durable"},
	})
}

func productionAdapterReadyObjectiveCloseoutWriterCloseoutReadback(result ProductionAdapterObjectiveCloseoutWriterDurableResult) ProductionAdapterObjectiveCloseoutReadback {
	return BuildProductionAdapterObjectiveCloseoutReadback(ProductionAdapterObjectiveCloseoutReadbackInput{
		ObjectiveCloseoutReadbackRef: result.ExpectedReadbackRef,
		DurableHandoff:               productionAdapterReadyObjectiveCloseoutDurableHandoff(),
		AppliedDurableEventRef:       result.AppliedDurableEventRef,
		AppliedRunstoreRef:           result.AppliedRunstoreRef,
		AppliedObjectiveStateRef:     result.AppliedObjectiveStateRef,
		HostDurableWriteReported:     true,
		HostDurableWriteSucceeded:    true,
	})
}
