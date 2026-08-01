package controlcontract

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestObjectiveRunStoreDurableRequestReady(t *testing.T) {
	request := objectiveRunStoreReadyDurableRequest()
	if request.Status != HostActionReady ||
		!request.ReadyForHostObjectiveRunStore ||
		!request.HostObjectiveRunStoreAuthorized ||
		!request.HostMayPersistObjectiveRun ||
		request.ObjectiveRunRef == "" ||
		request.LedgerRef == "" ||
		request.VerificationRef == "" ||
		request.ReplannerDecisionRef == "" ||
		request.ExpectedDurableEventRef == "" ||
		request.ExpectedObjectiveStateRef == "" ||
		!request.StoreMutationRequest.ReadyForHostStoreMutation ||
		!request.StoreMutationRequest.HostMayExecuteStoreMutation ||
		request.CoreInvocationExecuted ||
		request.DurableWriteByCore ||
		request.ObjectiveStoreWriteByCore ||
		request.RunstoreWriteByCore ||
		request.NextHostAction != "host_may_persist_objective_run" {
		t.Fatalf("unexpected objective run store request: %#v", request)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "objective run store durable request",
		RunnerEffect: request.RunnerEffect,
		PromptEffect: request.PromptEffect,
		Boundaries:   request.Boundaries,
		Payload:      request,
	}, "objective_run_store_durable_request", "host_owned_objective_run_store", "store_mutation_gate_reused", "transaction_replay_required", "ready_for_host_objective_run_store", "core_objective_run_store_write_not_executed")
}

func TestObjectiveRunStoreDurableResultAndReadbackRecorded(t *testing.T) {
	request := objectiveRunStoreReadyDurableRequest()
	result := objectiveRunStoreReadyDurableResult(request)
	if result.Status != HostActionRecorded ||
		!result.HostObjectiveRunStoreReported ||
		!result.HostObjectiveRunStoreSucceeded ||
		result.HostObjectiveRunStoreFailed ||
		!result.HostObjectiveRunStoreRecorded ||
		!result.ReadyForObjectiveRunStoreReadback ||
		result.ObjectiveRunStoreResultRef != request.ExpectedStoreMutationResultRef ||
		result.AppliedTransactionRef != request.ExpectedStoreTransactionRef ||
		result.AppliedRunstoreRef != request.HostRunstoreRef ||
		result.AppliedObjectiveStateRef != request.ExpectedObjectiveStateRef ||
		result.CoreInvocationExecuted ||
		result.DurableWriteByCore ||
		result.ObjectiveStoreWriteByCore ||
		result.RunstoreWriteByCore ||
		result.NextHostAction != "bind_objective_run_store_readback" {
		t.Fatalf("unexpected objective run store result: %#v", result)
	}

	readback := BuildObjectiveRunStoreDurableReadback(ObjectiveRunStoreDurableReadbackInput{
		ObjectiveRunStoreReadbackRef: "readback:objective_run_store",
		DurableResult:                result,
		ObservedTransactionRef:       result.AppliedTransactionRef,
		ObservedRunstoreRef:          result.AppliedRunstoreRef,
		ObservedObjectiveStateRef:    result.AppliedObjectiveStateRef,
		ReplayRef:                    "replay:objective_run_store",
	})
	if readback.Status != HostActionRecorded ||
		!readback.ObjectiveRunStoreReadbackBound ||
		!readback.TransactionReplayVerified ||
		!readback.ReadyForRuntimeLoopContinuation ||
		readback.ObjectiveRunStoreResultRef != result.ObjectiveRunStoreResultRef ||
		readback.ObservedTransactionRef != result.AppliedTransactionRef ||
		readback.ObservedRunstoreRef != result.AppliedRunstoreRef ||
		readback.ObservedObjectiveStateRef != result.AppliedObjectiveStateRef ||
		readback.CoreInvocationExecuted ||
		readback.DurableWriteByCore ||
		readback.ObjectiveStoreWriteByCore ||
		readback.RunstoreWriteByCore ||
		readback.NextHostAction != "continue_objective_runtime_loop" {
		t.Fatalf("unexpected objective run store readback: %#v", readback)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "objective run store durable readback",
		RunnerEffect: readback.RunnerEffect,
		PromptEffect: readback.PromptEffect,
		Boundaries:   readback.Boundaries,
		Payload:      readback,
	}, "objective_run_store_durable_readback", "host_owned_objective_run_store", "transaction_replay_verified_by_host", "objective_run_store_readback_bound", "ready_for_runtime_loop_continuation")
}

func TestObjectiveRunStoreDurableRequestRequiresReplannerDecisionRef(t *testing.T) {
	input := objectiveRunStoreReadyDurableRequestInput()
	input.ReplannerDecisionRef = ""
	request := BuildObjectiveRunStoreDurableRequest(input)
	if request.Status != HostActionBlocked ||
		request.ReadyForHostObjectiveRunStore ||
		request.HostMayPersistObjectiveRun ||
		request.FailureClass != FailureEvidenceMissing ||
		!missingInputContains(request.MissingInputs, "host:objective_replanner_decision_ref") ||
		!controlTokenListContains(request.BlockedReasons, "objective_run_replanner_decision_ref_missing") {
		t.Fatalf("expected replanner decision ref block, got %#v", request)
	}
}

func TestObjectiveRunStoreDurableRequestRejectsUnsafeRefWithoutLeak(t *testing.T) {
	input := objectiveRunStoreReadyDurableRequestInput()
	rawRef := "https://example.invalid/raw/objective"
	input.ObjectiveRunRef = DisplaySafeRef(rawRef)
	request := BuildObjectiveRunStoreDurableRequest(input)
	if request.Status != HostActionBlocked ||
		request.ReadyForHostObjectiveRunStore ||
		request.HostMayPersistObjectiveRun ||
		request.FailureClass != FailureEvidenceWeak ||
		request.RawOutputLoaded ||
		request.ObjectiveRunRef != "" ||
		!missingInputContains(request.MissingInputs, "host:display_safe_refs") ||
		!controlTokenListContains(request.BlockedReasons, "unsafe_input_ref") {
		t.Fatalf("expected unsafe-ref block without raw output retention, got %#v", request)
	}
	payload, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	if strings.Contains(string(payload), rawRef) {
		t.Fatalf("request leaked unsafe ref %q in %s", rawRef, payload)
	}
}

func TestObjectiveRunStoreDurableReadbackBlocksMismatch(t *testing.T) {
	result := objectiveRunStoreReadyDurableResult(objectiveRunStoreReadyDurableRequest())
	readback := BuildObjectiveRunStoreDurableReadback(ObjectiveRunStoreDurableReadbackInput{
		ObjectiveRunStoreReadbackRef: "readback:objective_run_store",
		DurableResult:                result,
		ObservedTransactionRef:       "store_txn:wrong",
		ObservedRunstoreRef:          result.AppliedRunstoreRef,
		ObservedObjectiveStateRef:    result.AppliedObjectiveStateRef,
		ReplayRef:                    "replay:objective_run_store",
	})
	if readback.Status != HostActionBlocked ||
		readback.ObjectiveRunStoreReadbackBound ||
		readback.TransactionReplayVerified ||
		readback.ReadyForRuntimeLoopContinuation ||
		readback.FailureClass != FailureVerificationFailed ||
		!controlTokenListContains(readback.BlockedReasons, "store_mutation_observed_transaction_ref_mismatch") {
		t.Fatalf("expected transaction mismatch block, got %#v", readback)
	}
}

func objectiveRunStoreReadyDurableRequestInput() ObjectiveRunStoreDurableRequestInput {
	loopInput := objectiveRuntimeLoopReadyInput()
	loopInput.LedgerPatch = objectiveRuntimeLoopLedgerPatch(VerificationPartial, FailureNone)
	loopInput.Verification = objectiveRuntimeLoopVerificationGate(VerificationPartial, FailureNone, "request_host_replanner_decision")
	loopInput.Observations = []Observation{{
		Kind:     "metric",
		Source:   "source:objective_run_store",
		Name:     "progress",
		Value:    "partial",
		Strength: EvidenceAdequate,
		EvidenceRefs: []EvidenceRef{{
			Ref:      "evidence:objective_run_store",
			Kind:     "metric",
			Strength: EvidenceAdequate,
		}},
	}}
	loopInput.EvidenceRefs = []EvidenceRef{{Ref: "evidence:objective_run_store", Kind: "metric", Strength: EvidenceAdequate}}
	loopInput.Boundaries = []Boundary{"test_objective_run_store_runtime_loop"}
	loop := BuildObjectiveRuntimeLoopStep(loopInput)
	return ObjectiveRunStoreDurableRequestInput{
		RuntimeLoop:                      loop,
		StoreMutationDescriptor:          objectiveRunStoreReadyMutationDescriptor(),
		ObjectiveRunStoreRequestRef:      "store_request:objective_run",
		ObjectiveRunRef:                  "objective_run:runtime_loop",
		LedgerRef:                        "ledger:runtime_loop",
		VerificationRef:                  "verification:runtime_loop_partial",
		ReplannerDecisionRef:             "replanner:runtime_loop_partial",
		ExpectedDurableEventRef:          "event:objective_run_persisted",
		ExpectedObjectiveStateRef:        "objective_state:runtime_loop_partial",
		HostStoreMutationConfirmationRef: "confirmation:objective_run_store",
		ExpectedStoreMutationResultRef:   "store_result:objective_run",
		ExpectedStoreTransactionRef:      "store_txn:objective_run",
		ExpectedStoreReadbackRef:         "readback:objective_run_store",
	}
}

func objectiveRunStoreReadyDurableRequest() ObjectiveRunStoreDurableRequest {
	return BuildObjectiveRunStoreDurableRequest(objectiveRunStoreReadyDurableRequestInput())
}

func objectiveRunStoreReadyDurableResult(request ObjectiveRunStoreDurableRequest) ObjectiveRunStoreDurableResult {
	return BuildObjectiveRunStoreDurableResult(ObjectiveRunStoreDurableResultInput{
		ObjectiveRunStoreResultRef:     request.ExpectedStoreMutationResultRef,
		DurableRequest:                 request,
		HostStoreAdapterRunRef:         "run:objective_run_store",
		HostObjectiveRunStoreReported:  true,
		HostObjectiveRunStoreSucceeded: true,
		AppliedTransactionRef:          request.ExpectedStoreTransactionRef,
		AppliedRunstoreRef:             request.HostRunstoreRef,
		AppliedObjectiveStateRef:       request.ExpectedObjectiveStateRef,
		StoreMutationEvidenceRefs:      []DisplaySafeRef{"evidence:objective_run_store"},
	})
}

func objectiveRunStoreReadyMutationDescriptor() ProductionAdapterStoreMutationDescriptor {
	return BuildProductionAdapterStoreMutationDescriptor(ProductionAdapterStoreMutationDescriptor{
		Available:                      true,
		StoreMutationDescriptorRef:     "store_descriptor:objective_run",
		StoreAdapterRef:                "store_adapter:objective_run",
		OwnerRef:                       "owner:host_reference",
		HostRunstoreRef:                "runstore:agentx_project",
		HostObjectiveStoreRef:          "objective_store:agentx_project",
		SupportsRunstoreMutation:       true,
		SupportsObjectiveStoreMutation: true,
		SupportsTransactionReplay:      true,
		SupportedMutationKinds: []string{
			"objective_run_upsert",
			"attempt_ledger_append",
			"verification_snapshot_upsert",
			"replanner_decision_append",
		},
		TransactionContractRef: "contract:objective_run_store_transaction",
		IdempotencyRef:         "idempotency:objective_run_store",
		IdempotencyContractRef: "contract:objective_run_store_idempotency",
		ReplayContractRef:      "contract:objective_run_store_replay",
		ReadbackContractRef:    "contract:objective_run_store_readback",
		RedactionPolicyRef:     "policy:display_safe_refs",
		TimeoutPolicyRef:       "policy:bounded_objective_run_store",
		RequiredPolicyRefs:     []DisplaySafeRef{"policy:objective_run_store_confirmation"},
	})
}
