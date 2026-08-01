package controlcontract

import (
	"encoding/json"
	"testing"
)

func TestProductionAdapterStoreMutationDescriptorReady(t *testing.T) {
	descriptor := productionAdapterReadyStoreMutationDescriptor()
	if descriptor.Status != HostActionReady ||
		!descriptor.ReadyForStoreMutationRequest ||
		!descriptor.SupportsRunstoreMutation ||
		!descriptor.SupportsObjectiveStoreMutation ||
		!descriptor.SupportsTransactionReplay ||
		descriptor.StoreMutationDescriptorRef == "" ||
		descriptor.StoreAdapterRef == "" ||
		descriptor.HostRunstoreRef == "" ||
		descriptor.HostObjectiveStoreRef == "" ||
		descriptor.TransactionContractRef == "" ||
		descriptor.IdempotencyRef == "" ||
		descriptor.IdempotencyContractRef == "" ||
		descriptor.ReplayContractRef == "" ||
		descriptor.ReadbackContractRef == "" ||
		descriptor.CoreInvocationExecuted ||
		descriptor.DurableWriteByCore ||
		descriptor.ObjectiveStoreWriteByCore ||
		descriptor.RunstoreWriteByCore ||
		descriptor.NextHostAction != "host_may_prepare_store_mutation_request" {
		t.Fatalf("unexpected store mutation descriptor: %#v", descriptor)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "store mutation descriptor",
		RunnerEffect: descriptor.RunnerEffect,
		PromptEffect: descriptor.PromptEffect,
		Boundaries:   descriptor.Boundaries,
		Payload:      descriptor,
	}, "production_adapter_store_mutation_descriptor", "store_mutation_gate_projection_only", "host_owned_store_adapter", "ready_for_host_store_mutation_request")
}

func TestProductionAdapterStoreMutationRequestReady(t *testing.T) {
	request := productionAdapterReadyStoreMutationRequest()
	if request.Status != HostActionReady ||
		!request.ReadyForHostStoreMutation ||
		!request.HostStoreMutationAuthorized ||
		!request.HostMayExecuteStoreMutation ||
		request.SourceRunstoreRef != request.HostRunstoreRef ||
		request.SourceObjectiveStateRef == "" ||
		request.HostStoreMutationConfirmationRef == "" ||
		request.ExpectedMutationResultRef == "" ||
		request.ExpectedTransactionRef == "" ||
		request.ExpectedReadbackRef == "" ||
		request.CoreInvocationExecuted ||
		request.DurableWriteByCore ||
		request.ObjectiveStoreWriteByCore ||
		request.RunstoreWriteByCore ||
		request.NextHostAction != "host_may_execute_store_mutation_adapter" {
		t.Fatalf("unexpected store mutation request: %#v", request)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "store mutation request",
		RunnerEffect: request.RunnerEffect,
		PromptEffect: request.PromptEffect,
		Boundaries:   request.Boundaries,
		Payload:      request,
	}, "production_adapter_store_mutation_request", "store_mutation_request_projection_only", "explicit_store_mutation_confirmation_required", "idempotency_required", "transaction_replay_required", "ready_for_host_store_mutation", "core_store_mutation_not_executed")
}

func TestProductionAdapterStoreMutationResultRecorded(t *testing.T) {
	request := productionAdapterReadyStoreMutationRequest()
	result := productionAdapterReadyStoreMutationResult(request)
	if result.Status != HostActionRecorded ||
		!result.HostStoreMutationReported ||
		!result.HostStoreMutationSucceeded ||
		result.HostStoreMutationFailed ||
		!result.HostStoreMutationRecorded ||
		!result.ReadyForStoreMutationReadback ||
		result.StoreMutationResultRef != request.ExpectedMutationResultRef ||
		result.AppliedTransactionRef != request.ExpectedTransactionRef ||
		result.AppliedRunstoreRef != request.HostRunstoreRef ||
		result.AppliedObjectiveStateRef != request.SourceObjectiveStateRef ||
		result.CoreInvocationExecuted ||
		result.DurableWriteByCore ||
		result.ObjectiveStoreWriteByCore ||
		result.RunstoreWriteByCore ||
		result.NextHostAction != "bind_store_mutation_readback" {
		t.Fatalf("unexpected store mutation result: %#v", result)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "store mutation result",
		RunnerEffect: result.RunnerEffect,
		PromptEffect: result.PromptEffect,
		Boundaries:   result.Boundaries,
		Payload:      result,
	}, "production_adapter_store_mutation_result", "store_mutation_result_projection_only", "host_store_mutation_recorded", "ready_for_store_mutation_readback")
}

func TestProductionAdapterStoreMutationReadbackBound(t *testing.T) {
	request := productionAdapterReadyStoreMutationRequest()
	result := productionAdapterReadyStoreMutationResult(request)
	readback := BuildProductionAdapterStoreMutationReadback(ProductionAdapterStoreMutationReadbackInput{
		StoreMutationReadbackRef:  "readback:store_mutation_closeout",
		StoreMutationResult:       result,
		ObservedTransactionRef:    result.AppliedTransactionRef,
		ObservedRunstoreRef:       result.AppliedRunstoreRef,
		ObservedObjectiveStateRef: result.AppliedObjectiveStateRef,
		ReplayRef:                 "replay:store_mutation_closeout",
	})
	if readback.Status != HostActionRecorded ||
		!readback.StoreMutationReadbackBound ||
		!readback.TransactionReplayVerified ||
		!readback.ReadyForDownstreamReadback ||
		readback.StoreMutationResultRef != result.StoreMutationResultRef ||
		readback.ObservedTransactionRef != result.AppliedTransactionRef ||
		readback.ObservedRunstoreRef != result.AppliedRunstoreRef ||
		readback.ObservedObjectiveStateRef != result.AppliedObjectiveStateRef ||
		readback.CoreInvocationExecuted ||
		readback.DurableWriteByCore ||
		readback.ObjectiveStoreWriteByCore ||
		readback.RunstoreWriteByCore ||
		readback.NextHostAction != "bind_downstream_readback" {
		t.Fatalf("unexpected store mutation readback: %#v", readback)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "store mutation readback",
		RunnerEffect: readback.RunnerEffect,
		PromptEffect: readback.PromptEffect,
		Boundaries:   readback.Boundaries,
		Payload:      readback,
	}, "production_adapter_store_mutation_readback", "store_mutation_readback_projection_only", "transaction_replay_verified_by_host", "store_mutation_readback_bound", "transaction_replay_verified", "ready_for_downstream_readback")
}

func TestProductionAdapterStoreMutationGateBlocksMissingAndMismatch(t *testing.T) {
	descriptor := productionAdapterReadyStoreMutationDescriptor()
	tests := []struct {
		name        string
		input       ProductionAdapterStoreMutationRequestInput
		wantReason  string
		wantMissing MissingInput
		wantFailure FailureClass
	}{
		{
			name: "missing confirmation",
			input: ProductionAdapterStoreMutationRequestInput{
				StoreMutationRequestRef:   "store_mutation_request:closeout",
				StoreMutationDescriptor:   descriptor,
				SourceDurableResultRef:    "durable_result:closeout",
				SourceDurableEventRef:     "event:objective_closed",
				SourceRunstoreRef:         descriptor.HostRunstoreRef,
				SourceObjectiveStateRef:   "objective_state:closed",
				ExpectedMutationResultRef: "store_mutation_result:closeout",
				ExpectedTransactionRef:    "store_txn:closeout",
				ExpectedReadbackRef:       "readback:store_mutation_closeout",
			},
			wantReason:  "host_store_mutation_confirmation_ref_missing",
			wantMissing: "host:store_mutation_confirmation_ref",
			wantFailure: FailureAuthorizationMissing,
		},
		{
			name: "runstore mismatch",
			input: ProductionAdapterStoreMutationRequestInput{
				StoreMutationRequestRef:          "store_mutation_request:closeout",
				StoreMutationDescriptor:          descriptor,
				SourceDurableResultRef:           "durable_result:closeout",
				SourceDurableEventRef:            "event:objective_closed",
				SourceRunstoreRef:                "runstore:wrong",
				SourceObjectiveStateRef:          "objective_state:closed",
				HostStoreMutationConfirmationRef: "confirmation:store_mutation_closeout",
				ExpectedMutationResultRef:        "store_mutation_result:closeout",
				ExpectedTransactionRef:           "store_txn:closeout",
				ExpectedReadbackRef:              "readback:store_mutation_closeout",
			},
			wantReason:  "store_mutation_runstore_ref_mismatch",
			wantMissing: "host:runstore_ref",
			wantFailure: FailureVerificationFailed,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			request := BuildProductionAdapterStoreMutationRequest(tc.input)
			if request.Status != HostActionBlocked ||
				request.ReadyForHostStoreMutation ||
				request.HostMayExecuteStoreMutation ||
				request.FailureClass != tc.wantFailure ||
				!productionAdapterStringContains(request.BlockedReasons, tc.wantReason) ||
				!productionAdapterMissingContains(request.MissingInputs, tc.wantMissing) ||
				request.CoreInvocationExecuted ||
				request.DurableWriteByCore ||
				request.ObjectiveStoreWriteByCore ||
				request.RunstoreWriteByCore {
				t.Fatalf("unexpected blocked store mutation request %s: %#v", tc.name, request)
			}
		})
	}
}

func TestProductionAdapterStoreMutationResultReviewAndReadbackMismatch(t *testing.T) {
	request := productionAdapterReadyStoreMutationRequest()
	failed := BuildProductionAdapterStoreMutationResult(ProductionAdapterStoreMutationResultInput{
		StoreMutationResultRef:    request.ExpectedMutationResultRef,
		StoreMutationRequest:      request,
		HostStoreAdapterRunRef:    "run:store_mutation_closeout",
		HostStoreMutationReported: true,
		HostStoreMutationFailed:   true,
		FailureRef:                "failure:store_mutation_closeout",
		CompensationRef:           "compensation:store_mutation_review",
	})
	if failed.Status != HostActionReviewRequired ||
		!failed.HostStoreMutationRecorded ||
		failed.ReadyForStoreMutationReadback ||
		failed.FailureClass != FailureVerificationFailed ||
		!productionAdapterStringContains(failed.BlockedReasons, "store_mutation_failed") ||
		failed.NextHostAction != "review_store_mutation_failure" ||
		failed.CoreInvocationExecuted ||
		failed.DurableWriteByCore ||
		failed.ObjectiveStoreWriteByCore ||
		failed.RunstoreWriteByCore {
		t.Fatalf("unexpected failed store mutation result: %#v", failed)
	}

	result := productionAdapterReadyStoreMutationResult(request)
	readback := BuildProductionAdapterStoreMutationReadback(ProductionAdapterStoreMutationReadbackInput{
		StoreMutationReadbackRef:  "readback:store_mutation_closeout",
		StoreMutationResult:       result,
		ObservedTransactionRef:    "store_txn:wrong",
		ObservedRunstoreRef:       result.AppliedRunstoreRef,
		ObservedObjectiveStateRef: result.AppliedObjectiveStateRef,
		ReplayRef:                 "replay:store_mutation_closeout",
	})
	if readback.Status != HostActionBlocked ||
		readback.StoreMutationReadbackBound ||
		readback.TransactionReplayVerified ||
		readback.ReadyForDownstreamReadback ||
		readback.FailureClass != FailureVerificationFailed ||
		!productionAdapterStringContains(readback.BlockedReasons, "store_mutation_observed_transaction_ref_mismatch") ||
		!productionAdapterMissingContains(readback.MissingInputs, "host:observed_store_transaction_ref") {
		t.Fatalf("unexpected store mutation readback mismatch: %#v", readback)
	}
}

func TestProductionAdapterStoreMutationGateRejectsUnsafeRefsAndJSONIsDisplaySafe(t *testing.T) {
	unsafe := productionAdapterReadyStoreMutationDescriptor()
	unsafe.StoreAdapterRef = "postgresql://secret.example.invalid/db"
	unsafeDescriptor := BuildProductionAdapterStoreMutationDescriptor(unsafe)
	if unsafeDescriptor.Status != HostActionBlocked ||
		unsafeDescriptor.ReadyForStoreMutationRequest ||
		unsafeDescriptor.FailureClass != FailureEvidenceWeak ||
		unsafeDescriptor.NextHostAction != "provide_display_safe_refs" ||
		!productionAdapterStringContains(unsafeDescriptor.BlockedReasons, "unsafe_input_ref") ||
		!productionAdapterMissingContains(unsafeDescriptor.MissingInputs, "host:display_safe_refs") {
		t.Fatalf("expected unsafe descriptor to block, got %#v", unsafeDescriptor)
	}

	request := productionAdapterReadyStoreMutationRequest()
	result := productionAdapterReadyStoreMutationResult(request)
	readback := BuildProductionAdapterStoreMutationReadback(ProductionAdapterStoreMutationReadbackInput{
		StoreMutationReadbackRef:  "readback:store_mutation_closeout",
		StoreMutationResult:       result,
		ObservedTransactionRef:    result.AppliedTransactionRef,
		ObservedRunstoreRef:       result.AppliedRunstoreRef,
		ObservedObjectiveStateRef: result.AppliedObjectiveStateRef,
		ReplayRef:                 "replay:store_mutation_closeout",
	})
	raw, err := json.Marshal(struct {
		Descriptor ProductionAdapterStoreMutationDescriptor `json:"descriptor"`
		Request    ProductionAdapterStoreMutationRequest    `json:"request"`
		Result     ProductionAdapterStoreMutationResult     `json:"result"`
		Readback   ProductionAdapterStoreMutationReadback   `json:"readback"`
	}{Descriptor: productionAdapterReadyStoreMutationDescriptor(), Request: request, Result: result, Readback: readback})
	if err != nil {
		t.Fatalf("marshal store mutation contracts: %v", err)
	}
	for _, token := range []string{
		"store_mutation_descriptor_ref",
		"host_store_mutation_confirmation_ref",
		"expected_transaction_ref",
		"transaction_replay_verified",
		"no_runstore_write_by_core",
	} {
		if !jsonPayloadContains(raw, token) {
			t.Fatalf("expected store mutation JSON token %q in %s", token, raw)
		}
	}
	AssertNoRawPayload(t, "store mutation contracts JSON", raw, "/Users/mason", "postgresql://secret", "raw local host task")
}

func productionAdapterReadyStoreMutationDescriptor() ProductionAdapterStoreMutationDescriptor {
	return BuildProductionAdapterStoreMutationDescriptor(ProductionAdapterStoreMutationDescriptor{
		Available:                      true,
		StoreMutationDescriptorRef:     "store_mutation_descriptor:objective_closeout",
		StoreAdapterRef:                "store_adapter:objective_closeout",
		OwnerRef:                       "owner:host_reference",
		HostRunstoreRef:                "runstore:agentx_project",
		HostObjectiveStoreRef:          "objective_store:agentx_project",
		SupportsRunstoreMutation:       true,
		SupportsObjectiveStoreMutation: true,
		SupportsTransactionReplay:      true,
		SupportedMutationKinds: []string{
			"runstore_append_event",
			"objective_state_closeout",
		},
		TransactionContractRef: "contract:store_mutation_transaction",
		IdempotencyRef:         "idempotency:objective_closeout_store_mutation",
		IdempotencyContractRef: "contract:store_mutation_idempotency",
		ReplayContractRef:      "contract:store_mutation_replay",
		ReadbackContractRef:    "contract:store_mutation_readback",
		RedactionPolicyRef:     "policy:display_safe_refs",
		TimeoutPolicyRef:       "policy:bounded_store_mutation",
		RequiredPolicyRefs: []DisplaySafeRef{
			"policy:explicit_store_mutation_confirmation",
		},
	})
}

func productionAdapterReadyStoreMutationRequest() ProductionAdapterStoreMutationRequest {
	descriptor := productionAdapterReadyStoreMutationDescriptor()
	return BuildProductionAdapterStoreMutationRequest(ProductionAdapterStoreMutationRequestInput{
		StoreMutationRequestRef:          "store_mutation_request:closeout",
		StoreMutationDescriptor:          descriptor,
		SourceDurableResultRef:           "durable_result:closeout",
		SourceDurableEventRef:            "event:objective_closed",
		SourceRunstoreRef:                descriptor.HostRunstoreRef,
		SourceObjectiveStateRef:          "objective_state:closed",
		HostStoreMutationConfirmationRef: "confirmation:store_mutation_closeout",
		ExpectedMutationResultRef:        "store_mutation_result:closeout",
		ExpectedTransactionRef:           "store_txn:closeout",
		ExpectedReadbackRef:              "readback:store_mutation_closeout",
	})
}

func productionAdapterReadyStoreMutationResult(request ProductionAdapterStoreMutationRequest) ProductionAdapterStoreMutationResult {
	return BuildProductionAdapterStoreMutationResult(ProductionAdapterStoreMutationResultInput{
		StoreMutationResultRef:     request.ExpectedMutationResultRef,
		StoreMutationRequest:       request,
		HostStoreAdapterRunRef:     "run:store_mutation_closeout",
		HostStoreMutationReported:  true,
		HostStoreMutationSucceeded: true,
		AppliedTransactionRef:      request.ExpectedTransactionRef,
		AppliedRunstoreRef:         request.HostRunstoreRef,
		AppliedObjectiveStateRef:   request.SourceObjectiveStateRef,
		StoreMutationEvidenceRefs:  []DisplaySafeRef{"evidence:store_mutation_closeout"},
	})
}
