package controlcontract

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestProductionAdapterObjectiveCloseoutWriterDryRunRequestReady(t *testing.T) {
	request := productionAdapterReadyObjectiveCloseoutWriterDryRunRequest()
	if request.Status != HostActionReady ||
		!request.ReadyForHostDryRun ||
		!request.HostDryRunAuthorized ||
		request.RequestedMode != ProductionAdapterObjectiveCloseoutWriterDryRun ||
		request.DryRunRequestRef == "" ||
		request.HostDryRunConfirmationRef == "" ||
		request.ExpectedDryRunResultRef == "" ||
		request.ExpectedReadbackRef == "" ||
		request.CoreInvocationExecuted ||
		request.DryRunByCore ||
		request.DurableWriteByCore ||
		request.ObjectiveStoreWriteByCore ||
		request.RunstoreWriteByCore ||
		request.NextHostAction != "host_may_run_objective_closeout_writer_dry_run_adapter" {
		t.Fatalf("unexpected objective closeout writer dry-run request: %#v", request)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "objective closeout writer dry-run request",
		RunnerEffect: request.RunnerEffect,
		PromptEffect: request.PromptEffect,
		Boundaries:   request.Boundaries,
		Payload:      request,
	}, "production_adapter_objective_closeout_writer_dry_run_request", "objective_closeout_writer_dry_run_request_projection_only", "host_owned_objective_closeout_writer_adapter", "objective_closeout_writer_dry_run_only", "ready_for_host_objective_closeout_writer_dry_run", "durable_write_not_enabled")
}

func TestProductionAdapterObjectiveCloseoutWriterDryRunResultRecorded(t *testing.T) {
	request := productionAdapterReadyObjectiveCloseoutWriterDryRunRequest()
	adapter := productionAdapterObjectiveCloseoutWriterFakeAdapter{descriptor: productionAdapterReadyObjectiveCloseoutWriterDescriptor()}
	var _ ProductionAdapterObjectiveCloseoutWriterHostAdapter = adapter

	result := adapter.DryRunObjectiveCloseoutWriter(request)
	if result.Status != HostActionRecorded ||
		!result.HostDryRunReported ||
		!result.HostDryRunSucceeded ||
		!result.HostDryRunRecorded ||
		!result.ReadyForDurableWriteOptIn ||
		result.DryRunResultRef != request.ExpectedDryRunResultRef ||
		result.ExpectedReadbackRef != request.ExpectedReadbackRef ||
		result.CoreInvocationExecuted ||
		result.DryRunByCore ||
		result.DurableWriteByCore ||
		result.ObjectiveStoreWriteByCore ||
		result.RunstoreWriteByCore ||
		result.NextHostAction != "review_objective_closeout_writer_durable_opt_in" {
		t.Fatalf("unexpected objective closeout writer dry-run result: %#v", result)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "objective closeout writer dry-run result",
		RunnerEffect: result.RunnerEffect,
		PromptEffect: result.PromptEffect,
		Boundaries:   result.Boundaries,
		Payload:      result,
	}, "production_adapter_objective_closeout_writer_dry_run_result", "objective_closeout_writer_dry_run_result_projection_only", "host_owned_objective_closeout_writer_adapter", "host_objective_closeout_writer_dry_run_recorded", "ready_for_objective_closeout_writer_durable_opt_in", "durable_write_not_enabled")
}

func TestProductionAdapterObjectiveCloseoutWriterDryRunSmokeHarnessPassed(t *testing.T) {
	request := productionAdapterReadyObjectiveCloseoutWriterDryRunRequest()
	result := productionAdapterReadyObjectiveCloseoutWriterDryRunResult(request)
	smoke := BuildProductionAdapterObjectiveCloseoutWriterDryRunSmokeHarness(ProductionAdapterObjectiveCloseoutWriterDryRunSmokeHarnessInput{
		SmokeRef:      "smoke:metrics_objective_closeout_writer_dry_run",
		DryRunRequest: request,
		DryRunResult:  result,
	})
	if smoke.Status != "dry_run_smoke_passed" ||
		!smoke.ReadyForHostDisplay ||
		!smoke.SmokePassed ||
		!smoke.ReadyForDurableWriteOptIn ||
		smoke.DryRunRequestRef != request.DryRunRequestRef ||
		smoke.DryRunResultRef != result.DryRunResultRef ||
		smoke.ExpectedReadbackRef != request.ExpectedReadbackRef ||
		smoke.CoreInvocationExecuted ||
		smoke.DryRunByCore ||
		smoke.DurableWriteByCore ||
		smoke.ObjectiveStoreWriteByCore ||
		smoke.RunstoreWriteByCore ||
		smoke.NextHostAction != "review_objective_closeout_writer_durable_opt_in" {
		t.Fatalf("unexpected objective closeout writer dry-run smoke harness: %#v", smoke)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "objective closeout writer dry-run smoke harness",
		RunnerEffect: smoke.RunnerEffect,
		PromptEffect: smoke.PromptEffect,
		Boundaries:   smoke.Boundaries,
		Payload:      smoke,
	}, "production_adapter_objective_closeout_writer_dry_run_smoke_harness", "objective_closeout_writer_dry_run_smoke_harness_projection_only", "host_owned_objective_closeout_writer_adapter", "dry_run_smoke_passed", "ready_for_objective_closeout_writer_durable_opt_in", "durable_write_not_enabled")
}

func TestProductionAdapterObjectiveCloseoutWriterDryRunRequestBlocksMissingInputsAndUnsafeRefs(t *testing.T) {
	fixture := productionAdapterReadyObjectiveCloseoutWriterDryRunFixture()
	tests := []struct {
		name        string
		input       ProductionAdapterObjectiveCloseoutWriterDryRunRequestInput
		wantReason  string
		wantMissing MissingInput
		wantFailure FailureClass
		rejected    []string
	}{
		{
			name: "missing host dry-run confirmation",
			input: ProductionAdapterObjectiveCloseoutWriterDryRunRequestInput{
				DryRunRequestRef:        "dryrun_request:metrics_objective_closeout_writer",
				WriterFixture:           fixture,
				ExpectedDryRunResultRef: "dryrun_result:metrics_objective_closeout_writer",
				ExpectedReadbackRef:     "readback:metrics_objective_closeout_writer_expected",
			},
			wantReason:  "host_dry_run_confirmation_ref_missing",
			wantMissing: "host:objective_closeout_writer_dry_run_confirmation_ref",
			wantFailure: FailureAuthorizationMissing,
		},
		{
			name: "missing expected dry-run result",
			input: ProductionAdapterObjectiveCloseoutWriterDryRunRequestInput{
				DryRunRequestRef:          "dryrun_request:metrics_objective_closeout_writer",
				WriterFixture:             fixture,
				HostDryRunConfirmationRef: "confirmation:metrics_objective_closeout_writer_dry_run",
				ExpectedReadbackRef:       "readback:metrics_objective_closeout_writer_expected",
			},
			wantReason:  "expected_dry_run_result_ref_missing",
			wantMissing: "host:objective_closeout_writer_expected_dry_run_result_ref",
			wantFailure: FailureEvidenceMissing,
		},
		{
			name: "unsafe expected readback ref",
			input: ProductionAdapterObjectiveCloseoutWriterDryRunRequestInput{
				DryRunRequestRef:          "dryrun_request:metrics_objective_closeout_writer",
				WriterFixture:             fixture,
				HostDryRunConfirmationRef: "confirmation:metrics_objective_closeout_writer_dry_run",
				ExpectedDryRunResultRef:   "dryrun_result:metrics_objective_closeout_writer",
				ExpectedReadbackRef:       "postgresql://secret@example.invalid/db",
			},
			wantReason:  "unsafe_input_ref",
			wantMissing: "host:display_safe_refs",
			wantFailure: FailureEvidenceWeak,
			rejected:    []string{"postgresql://secret", "example.invalid"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			request := BuildProductionAdapterObjectiveCloseoutWriterDryRunRequest(tc.input)
			if request.Status != HostActionBlocked ||
				request.ReadyForHostDryRun ||
				request.HostDryRunAuthorized ||
				request.FailureClass != tc.wantFailure ||
				!productionAdapterStringContains(request.BlockedReasons, tc.wantReason) ||
				!productionAdapterMissingContains(request.MissingInputs, tc.wantMissing) ||
				request.CoreInvocationExecuted ||
				request.DryRunByCore ||
				request.DurableWriteByCore {
				t.Fatalf("unexpected blocked dry-run request %s: %#v", tc.name, request)
			}
			AssertNoRawPayload(t, "objective closeout writer dry-run request "+tc.name, request, append([]string{"/Users/mason", "raw local host task"}, tc.rejected...)...)
		})
	}
}

func TestProductionAdapterObjectiveCloseoutWriterDryRunResultBlocksMismatches(t *testing.T) {
	request := productionAdapterReadyObjectiveCloseoutWriterDryRunRequest()
	tests := []struct {
		name        string
		input       ProductionAdapterObjectiveCloseoutWriterDryRunResultInput
		wantReason  string
		wantMissing MissingInput
	}{
		{
			name: "not reported",
			input: ProductionAdapterObjectiveCloseoutWriterDryRunResultInput{
				DryRunResultRef:     request.ExpectedDryRunResultRef,
				DryRunRequest:       request,
				HostAdapterRunRef:   "run:metrics_objective_closeout_writer_dry_run",
				ExpectedReadbackRef: request.ExpectedReadbackRef,
			},
			wantReason:  "writer_dry_run_not_reported",
			wantMissing: "host:objective_closeout_writer_dry_run_report",
		},
		{
			name: "result ref mismatch",
			input: ProductionAdapterObjectiveCloseoutWriterDryRunResultInput{
				DryRunResultRef:     "dryrun_result:wrong_metrics_objective_closeout_writer",
				DryRunRequest:       request,
				HostAdapterRunRef:   "run:metrics_objective_closeout_writer_dry_run",
				HostDryRunReported:  true,
				HostDryRunSucceeded: true,
				ExpectedReadbackRef: request.ExpectedReadbackRef,
			},
			wantReason:  "dry_run_result_ref_mismatch",
			wantMissing: "host:objective_closeout_writer_dry_run_result_ref",
		},
		{
			name: "expected readback mismatch",
			input: ProductionAdapterObjectiveCloseoutWriterDryRunResultInput{
				DryRunResultRef:     request.ExpectedDryRunResultRef,
				DryRunRequest:       request,
				HostAdapterRunRef:   "run:metrics_objective_closeout_writer_dry_run",
				HostDryRunReported:  true,
				HostDryRunSucceeded: true,
				ExpectedReadbackRef: "readback:wrong_metrics_objective_closeout_writer",
			},
			wantReason:  "expected_readback_ref_mismatch",
			wantMissing: "host:objective_closeout_writer_expected_readback_ref",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := BuildProductionAdapterObjectiveCloseoutWriterDryRunResult(tc.input)
			if result.Status != HostActionBlocked ||
				result.ReadyForDurableWriteOptIn ||
				result.HostDryRunRecorded ||
				!productionAdapterStringContains(result.BlockedReasons, tc.wantReason) ||
				!productionAdapterMissingContains(result.MissingInputs, tc.wantMissing) ||
				result.CoreInvocationExecuted ||
				result.DryRunByCore ||
				result.DurableWriteByCore {
				t.Fatalf("unexpected blocked dry-run result %s: %#v", tc.name, result)
			}
		})
	}
}

func TestProductionAdapterObjectiveCloseoutWriterDryRunSmokeHarnessBlocksMismatch(t *testing.T) {
	request := productionAdapterReadyObjectiveCloseoutWriterDryRunRequest()
	result := productionAdapterReadyObjectiveCloseoutWriterDryRunResult(request)
	result.DryRunRequestRef = "dryrun_request:wrong_metrics_objective_closeout_writer"
	smoke := BuildProductionAdapterObjectiveCloseoutWriterDryRunSmokeHarness(ProductionAdapterObjectiveCloseoutWriterDryRunSmokeHarnessInput{
		SmokeRef:      "smoke:metrics_objective_closeout_writer_dry_run",
		DryRunRequest: request,
		DryRunResult:  result,
	})
	if smoke.Status != "blocked" ||
		smoke.SmokePassed ||
		smoke.ReadyForDurableWriteOptIn ||
		!productionAdapterStringContains(smoke.BlockedReasons, "dry_run_request_ref_mismatch") ||
		!productionAdapterMissingContains(smoke.MissingInputs, "host:objective_closeout_writer_dry_run_request_ref") ||
		smoke.CoreInvocationExecuted ||
		smoke.DryRunByCore ||
		smoke.DurableWriteByCore {
		t.Fatalf("expected dry-run smoke mismatch to block, got %#v", smoke)
	}
}

func TestProductionAdapterObjectiveCloseoutWriterDryRunContractsJSONCompatibility(t *testing.T) {
	request := productionAdapterReadyObjectiveCloseoutWriterDryRunRequest()
	result := productionAdapterReadyObjectiveCloseoutWriterDryRunResult(request)
	smoke := BuildProductionAdapterObjectiveCloseoutWriterDryRunSmokeHarness(ProductionAdapterObjectiveCloseoutWriterDryRunSmokeHarnessInput{
		SmokeRef:      "smoke:metrics_objective_closeout_writer_dry_run",
		DryRunRequest: request,
		DryRunResult:  result,
	})
	raw, err := json.Marshal(struct {
		Request ProductionAdapterObjectiveCloseoutWriterDryRunRequest      `json:"request"`
		Result  ProductionAdapterObjectiveCloseoutWriterDryRunResult       `json:"result"`
		Smoke   ProductionAdapterObjectiveCloseoutWriterDryRunSmokeHarness `json:"smoke"`
	}{Request: request, Result: result, Smoke: smoke})
	if err != nil {
		t.Fatalf("marshal dry-run contracts: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("unmarshal dry-run contracts: %v", err)
	}
	for _, token := range []string{
		"dry_run_request_ref",
		"expected_dry_run_result_ref",
		"host_adapter_run_ref",
		"ready_for_durable_write_opt_in",
		"smoke_passed",
		"objective_closeout_writer_dry_run_only",
	} {
		if !jsonPayloadContains(raw, token) {
			t.Fatalf("expected dry-run JSON token %q in %s", token, raw)
		}
	}
	AssertNoRawPayload(t, "objective closeout writer dry-run contracts JSON", raw, "/Users/mason", "postgresql://secret", "raw local host task")
}

type productionAdapterObjectiveCloseoutWriterFakeAdapter struct {
	descriptor ProductionAdapterObjectiveCloseoutWriterDescriptor
}

func (a productionAdapterObjectiveCloseoutWriterFakeAdapter) ObjectiveCloseoutWriterDescriptor() ProductionAdapterObjectiveCloseoutWriterDescriptor {
	return a.descriptor
}

func (a productionAdapterObjectiveCloseoutWriterFakeAdapter) DryRunObjectiveCloseoutWriter(request ProductionAdapterObjectiveCloseoutWriterDryRunRequest) ProductionAdapterObjectiveCloseoutWriterDryRunResult {
	return productionAdapterReadyObjectiveCloseoutWriterDryRunResult(request)
}

func (a productionAdapterObjectiveCloseoutWriterFakeAdapter) ExecuteObjectiveCloseoutDurableWriter(request ProductionAdapterObjectiveCloseoutWriterDurableRequest) ProductionAdapterObjectiveCloseoutWriterDurableResult {
	return productionAdapterReadyObjectiveCloseoutWriterDurableResult(request)
}

func productionAdapterReadyObjectiveCloseoutWriterDryRunFixture() ProductionAdapterObjectiveCloseoutWriterBlackboxFixture {
	handoff, uiHandoff := productionAdapterReadyObjectiveCloseoutWriterInputs()
	descriptor := productionAdapterReadyObjectiveCloseoutWriterDescriptor()
	optIn := BuildProductionAdapterObjectiveCloseoutWriterOptIn(productionAdapterObjectiveCloseoutWriterOptInInputForMode(
		descriptor,
		handoff,
		uiHandoff,
		"optin:metrics_objective_closeout_writer_dry_run",
		ProductionAdapterObjectiveCloseoutWriterDryRun,
	))
	return BuildProductionAdapterObjectiveCloseoutWriterBlackboxFixture(ProductionAdapterObjectiveCloseoutWriterBlackboxFixtureInput{
		FixtureRef:  "fixture:metrics_objective_closeout_writer_dry_run",
		WriterOptIn: optIn,
	})
}

func productionAdapterReadyObjectiveCloseoutWriterDryRunRequest() ProductionAdapterObjectiveCloseoutWriterDryRunRequest {
	return BuildProductionAdapterObjectiveCloseoutWriterDryRunRequest(ProductionAdapterObjectiveCloseoutWriterDryRunRequestInput{
		DryRunRequestRef:          "dryrun_request:metrics_objective_closeout_writer",
		WriterFixture:             productionAdapterReadyObjectiveCloseoutWriterDryRunFixture(),
		HostDryRunConfirmationRef: "confirmation:metrics_objective_closeout_writer_dry_run",
		ExpectedDryRunResultRef:   "dryrun_result:metrics_objective_closeout_writer",
		ExpectedReadbackRef:       "readback:metrics_objective_closeout_writer_expected",
	})
}

func productionAdapterReadyObjectiveCloseoutWriterDryRunResult(request ProductionAdapterObjectiveCloseoutWriterDryRunRequest) ProductionAdapterObjectiveCloseoutWriterDryRunResult {
	return BuildProductionAdapterObjectiveCloseoutWriterDryRunResult(ProductionAdapterObjectiveCloseoutWriterDryRunResultInput{
		DryRunResultRef:     request.ExpectedDryRunResultRef,
		DryRunRequest:       request,
		HostAdapterRunRef:   "run:metrics_objective_closeout_writer_dry_run",
		HostDryRunReported:  true,
		HostDryRunSucceeded: true,
		ExpectedReadbackRef: request.ExpectedReadbackRef,
		DryRunEvidenceRefs:  []DisplaySafeRef{"evidence:metrics_objective_closeout_writer_dry_run"},
	})
}

func jsonPayloadContains(raw []byte, token string) bool {
	return strings.Contains(string(raw), token)
}
