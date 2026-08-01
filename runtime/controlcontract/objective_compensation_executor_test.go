package controlcontract

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestObjectiveCompensationExecutionRequestReady(t *testing.T) {
	request := objectiveCompensationReadyRequest()
	if request.Status != HostActionReady ||
		!request.ReadyForHostCompensation ||
		!request.HostCompensationAuthorized ||
		!request.HostMayExecuteCompensation ||
		request.CompensationRef != "compensation:closeout_store_revert" ||
		request.ExecutorRef != "compensation_executor:store_revert" ||
		request.FailureClass != FailureNone ||
		request.NextHostAction != "host_may_execute_compensation" {
		t.Fatalf("unexpected compensation request: %#v", request)
	}
	if request.CoreExecutionExecuted ||
		request.CompensationExecutedByCore ||
		request.RunnerDispatched ||
		request.ToolExecuted ||
		request.WorkflowDispatched ||
		request.SchedulerApplied ||
		request.InstallerExecuted ||
		request.StoreMutationExecuted {
		t.Fatalf("core must not execute compensation side effects: %#v", request)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "host owned compensation request",
		RunnerEffect: request.RunnerEffect,
		PromptEffect: request.PromptEffect,
		Boundaries:   request.Boundaries,
		Payload:      request,
	}, "objective_compensation_execution_request", "host_owned_compensation_execution_gate", "compensation_request_not_execution", "no_compensation_execution_by_core")
}

func TestObjectiveCompensationExecutionResultAndReadbackRecorded(t *testing.T) {
	request := objectiveCompensationReadyRequest()
	result := objectiveCompensationSuccessResult(request)
	if result.Status != HostActionRecorded ||
		!result.HostCompensationReported ||
		!result.HostCompensationSucceeded ||
		result.HostCompensationFailed ||
		!result.HostCompensationExecuted ||
		!result.ReadyForCompensationReadback ||
		!result.ReadyForCloseoutReview ||
		result.CompensationResultRef != request.ExpectedResultRef ||
		result.AppliedCompensationRef != request.CompensationRef ||
		result.NextHostAction != "bind_compensation_readback" {
		t.Fatalf("unexpected compensation result: %#v", result)
	}
	readback := BuildObjectiveCompensationExecutionReadback(ObjectiveCompensationExecutionReadbackInput{
		Result:                  result,
		CompensationReadbackRef: result.ExpectedReadbackRef,
		ObservedCompensationRef: result.AppliedCompensationRef,
		ObservedHostRunRef:      result.HostCompensationRunRef,
		ReadbackEvidenceRefs: []EvidenceRef{{
			Ref:      "evidence:compensation_readback",
			Kind:     "host_readback",
			Strength: EvidenceAdequate,
		}},
	})
	if readback.Status != HostActionRecorded ||
		!readback.CompensationReadbackBound ||
		!readback.ReadyForCloseoutReview ||
		!readback.CompensationSucceeded ||
		readback.ResidualRiskRecorded ||
		readback.CompensationReadbackRef != result.ExpectedReadbackRef ||
		readback.NextHostAction != "continue_closeout_failure_review" {
		t.Fatalf("unexpected compensation readback: %#v", readback)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "host owned compensation readback",
		RunnerEffect: readback.RunnerEffect,
		PromptEffect: readback.PromptEffect,
		Boundaries:   readback.Boundaries,
		Payload:      readback,
	}, "objective_compensation_execution_readback", "host_owned_compensation_readback", "compensation_readback_bound", "no_compensation_execution_by_core")
}

func TestObjectiveCompensationExecutionBlocksUnsupportedCompensation(t *testing.T) {
	input := objectiveCompensationReadyRequestInput()
	input.ExecutorDescriptor.SupportedCompensationRefs = []DisplaySafeRef{"compensation:other"}
	request := BuildObjectiveCompensationExecutionRequest(input)
	if request.Status != HostActionBlocked ||
		request.ReadyForHostCompensation ||
		request.HostMayExecuteCompensation ||
		request.FailureClass != FailureUnsupportedOperation ||
		!controlTokenListContains(request.BlockedReasons, "compensation_ref_not_supported") ||
		!missingInputContains(request.MissingInputs, "compensation:closeout_store_revert") ||
		request.NextHostAction != "record_residual_risk" {
		t.Fatalf("expected unsupported compensation block, got %#v", request)
	}
}

func TestObjectiveCompensationExecutionFailureRequiresResidualRisk(t *testing.T) {
	request := objectiveCompensationReadyRequest()
	result := BuildObjectiveCompensationExecutionResult(ObjectiveCompensationExecutionResultInput{
		Request:                  request,
		CompensationResultRef:    request.ExpectedResultRef,
		HostCompensationRunRef:   "compensation_run:store_revert_failed",
		HostCompensationReported: true,
		HostCompensationFailed:   true,
		FailureRef:               "failure:compensation_store_revert",
		CompensationEvidenceRefs: []EvidenceRef{{Ref: "evidence:compensation_failure", Kind: "host_report", Strength: EvidenceAdequate}},
	})
	if result.Status != HostActionBlocked ||
		result.ReadyForCompensationReadback ||
		result.ReadyForCloseoutReview ||
		result.FailureClass != FailureEvidenceMissing ||
		!missingInputContains(result.MissingInputs, "host:residual_risk_ref") ||
		!controlTokenListContains(result.BlockedReasons, "compensation_residual_risk_ref_missing") ||
		result.NextHostAction != "record_residual_risk" {
		t.Fatalf("expected residual risk block, got %#v", result)
	}
}

func TestObjectiveCompensationExecutionFailureReadbackRecordsResidualRisk(t *testing.T) {
	request := objectiveCompensationReadyRequest()
	result := BuildObjectiveCompensationExecutionResult(ObjectiveCompensationExecutionResultInput{
		Request:                  request,
		CompensationResultRef:    request.ExpectedResultRef,
		HostCompensationRunRef:   "compensation_run:store_revert_failed",
		HostCompensationReported: true,
		HostCompensationFailed:   true,
		FailureRef:               "failure:compensation_store_revert",
		ResidualRiskRef:          "residual_risk:manual_store_review",
		CompensationEvidenceRefs: []EvidenceRef{{Ref: "evidence:compensation_failure", Kind: "host_report", Strength: EvidenceAdequate}},
	})
	readback := BuildObjectiveCompensationExecutionReadback(ObjectiveCompensationExecutionReadbackInput{
		Result:                  result,
		CompensationReadbackRef: result.ExpectedReadbackRef,
		ObservedHostRunRef:      result.HostCompensationRunRef,
		ObservedResidualRiskRef: result.ResidualRiskRef,
		ReadbackEvidenceRefs:    []EvidenceRef{{Ref: "evidence:residual_risk_readback", Kind: "host_readback", Strength: EvidenceAdequate}},
	})
	if readback.Status != HostActionRecorded ||
		!readback.CompensationReadbackBound ||
		!readback.ReadyForCloseoutReview ||
		readback.CompensationSucceeded ||
		!readback.ResidualRiskRecorded ||
		readback.NextHostAction != "continue_closeout_failure_review" {
		t.Fatalf("unexpected residual risk readback: %#v", readback)
	}
}

func TestObjectiveCompensationExecutionRejectsUnsafeRefWithoutLeak(t *testing.T) {
	input := objectiveCompensationReadyRequestInput()
	rawRef := "https://example.invalid/raw/compensation"
	input.CompensationRequestRef = DisplaySafeRef(rawRef)
	request := BuildObjectiveCompensationExecutionRequest(input)
	if request.Status != HostActionBlocked ||
		request.ReadyForHostCompensation ||
		request.HostMayExecuteCompensation ||
		request.FailureClass != FailureEvidenceWeak ||
		!request.RawOutputLoaded ||
		!missingInputContains(request.MissingInputs, "host:display_safe_refs") ||
		!controlTokenListContains(request.BlockedReasons, "unsafe_input_ref") {
		t.Fatalf("expected unsafe-ref block, got %#v", request)
	}
	payload, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	if strings.Contains(string(payload), rawRef) {
		t.Fatalf("request leaked unsafe ref %q in %s", rawRef, payload)
	}
}

func TestObjectiveCompensationExecutionResultRejectsUnsafeRequestRefWithoutLeak(t *testing.T) {
	request := objectiveCompensationReadyRequest()
	rawRef := "https://example.invalid/raw/executor"
	request.ExecutorRef = DisplaySafeRef(rawRef)
	result := BuildObjectiveCompensationExecutionResult(ObjectiveCompensationExecutionResultInput{
		Request:                   request,
		CompensationResultRef:     request.ExpectedResultRef,
		HostCompensationRunRef:    "compensation_run:store_revert",
		HostCompensationReported:  true,
		HostCompensationSucceeded: true,
		AppliedCompensationRef:    request.CompensationRef,
	})
	if result.Status != HostActionBlocked ||
		result.ReadyForCompensationReadback ||
		result.ReadyForCloseoutReview ||
		result.HostCompensationExecuted ||
		result.FailureClass != FailureEvidenceWeak ||
		!result.RawOutputLoaded ||
		!missingInputContains(result.MissingInputs, "host:display_safe_refs") ||
		!controlTokenListContains(result.BlockedReasons, "unsafe_input_ref") {
		t.Fatalf("expected unsafe request ref block, got %#v", result)
	}
	payload, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal result: %v", err)
	}
	if strings.Contains(string(payload), rawRef) {
		t.Fatalf("result leaked unsafe ref %q in %s", rawRef, payload)
	}
}

func TestObjectiveCompensationExecutionReadbackRejectsUnsafeResultRefWithoutLeak(t *testing.T) {
	result := objectiveCompensationSuccessResult(objectiveCompensationReadyRequest())
	rawRef := "https://example.invalid/raw/compensation-run"
	result.HostCompensationRunRef = DisplaySafeRef(rawRef)
	readback := BuildObjectiveCompensationExecutionReadback(ObjectiveCompensationExecutionReadbackInput{
		Result:                  result,
		CompensationReadbackRef: result.ExpectedReadbackRef,
		ObservedCompensationRef: result.AppliedCompensationRef,
		ObservedHostRunRef:      "compensation_run:store_revert",
	})
	if readback.Status != HostActionBlocked ||
		readback.CompensationReadbackBound ||
		readback.ReadyForCloseoutReview ||
		readback.FailureClass != FailureEvidenceWeak ||
		!readback.RawOutputLoaded ||
		!missingInputContains(readback.MissingInputs, "host:display_safe_refs") ||
		!controlTokenListContains(readback.BlockedReasons, "unsafe_input_ref") {
		t.Fatalf("expected unsafe result ref block, got %#v", readback)
	}
	payload, err := json.Marshal(readback)
	if err != nil {
		t.Fatalf("marshal readback: %v", err)
	}
	if strings.Contains(string(payload), rawRef) {
		t.Fatalf("readback leaked unsafe ref %q in %s", rawRef, payload)
	}
}

func objectiveCompensationReadyRequestInput() ObjectiveCompensationExecutionRequestInput {
	return ObjectiveCompensationExecutionRequestInput{
		CompensationRequestRef:      "compensation_request:closeout_store_revert",
		FailureReview:               objectiveCompensationFailureReview(),
		ExecutorDescriptor:          objectiveCompensationExecutorDescriptor(),
		HostCompensationApprovalRef: "approval:host_compensation",
		IdempotencyRef:              "idempotency:compensation_store_revert",
		ExpectedResultRef:           "compensation_result:closeout_store_revert",
		ExpectedReadbackRef:         "compensation_readback:closeout_store_revert",
		RollbackPlanRef:             "rollback_plan:closeout_store_revert",
		AuditRef:                    "audit:compensation_store_revert",
		PolicyRefs:                  []DisplaySafeRef{"policy:compensation_allowed"},
		EvidenceRefs: []EvidenceRef{{
			Ref:      "evidence:closeout_failure",
			Kind:     "failure_review",
			Strength: EvidenceAdequate,
		}},
	}
}

func objectiveCompensationReadyRequest() ObjectiveCompensationExecutionRequest {
	return BuildObjectiveCompensationExecutionRequest(objectiveCompensationReadyRequestInput())
}

func objectiveCompensationSuccessResult(request ObjectiveCompensationExecutionRequest) ObjectiveCompensationExecutionResult {
	return BuildObjectiveCompensationExecutionResult(ObjectiveCompensationExecutionResultInput{
		Request:                   request,
		CompensationResultRef:     request.ExpectedResultRef,
		HostCompensationRunRef:    "compensation_run:store_revert",
		HostCompensationReported:  true,
		HostCompensationSucceeded: true,
		AppliedCompensationRef:    request.CompensationRef,
		CompensationEvidenceRefs: []EvidenceRef{{
			Ref:      "evidence:compensation_success",
			Kind:     "host_report",
			Strength: EvidenceAdequate,
		}},
	})
}

func objectiveCompensationFailureReview() ProductionAdapterObjectiveCloseoutFailureReviewPacket {
	return ProductionAdapterObjectiveCloseoutFailureReviewPacket{
		ContractVersion:              ContractVersion,
		Projected:                    true,
		Available:                    true,
		Status:                       "ready_for_objective_closeout_failure_review",
		Mode:                         "production_adapter_objective_closeout_failure_review_packet",
		ReadyForHostDisplay:          true,
		ReadyForFailureReview:        true,
		ReadyForCompensationReview:   true,
		HostDurableWriteReported:     true,
		HostDurableWriteFailed:       true,
		FailureReviewPacketRef:       "failure_review:closeout_store",
		HostViewRef:                  "host_view:closeout_store",
		ObjectiveCloseoutPacketRef:   "closeout_packet:store",
		ObjectiveCloseoutReadbackRef: "closeout_readback:store",
		ObjectiveRef:                 "objective:store_closeout",
		HostObjectiveLifecycleRef:    "objective_lifecycle:store",
		HostRunstoreRef:              "runstore:objective",
		ExpectedDurableEventRef:      "event:closeout_store_failed",
		ExpectedObjectiveStateRef:    "objective_state:closeout_failed",
		FailureRef:                   "failure:closeout_store_write",
		CompensationRef:              "compensation:closeout_store_revert",
		AdapterRef:                   "adapter:objective_closeout_writer",
		EvidenceRefs: []EvidenceRef{{
			Ref:      "evidence:closeout_failure",
			Kind:     "failure_review",
			Strength: EvidenceAdequate,
		}},
		Boundaries:   []Boundary{"test_failure_review_ready"},
		RunnerEffect: "none",
		PromptEffect: "none",
	}.Normalize()
}

func objectiveCompensationExecutorDescriptor() ObjectiveCompensationExecutorDescriptor {
	return ObjectiveCompensationExecutorDescriptor{
		Available:                 true,
		DescriptorRef:             "compensation_descriptor:store_revert",
		ExecutorRef:               "compensation_executor:store_revert",
		OwnerRef:                  "owner:product_host",
		SupportedCompensationRefs: []DisplaySafeRef{"compensation:closeout_store_revert"},
		IdempotencyContractRef:    "contract:compensation_idempotency",
		ReadbackContractRef:       "contract:compensation_readback",
		RollbackContractRef:       "contract:compensation_rollback",
		TimeoutPolicyRef:          "policy:compensation_timeout",
		PolicyRefs:                []DisplaySafeRef{"policy:compensation_allowed"},
		RequiredPolicyRefs:        []DisplaySafeRef{"policy:compensation_allowed"},
		ApprovalRefs:              []DisplaySafeRef{"approval:host_compensation"},
		RequiredApprovalRefs:      []DisplaySafeRef{"approval:host_compensation"},
		Boundaries:                []Boundary{"test_compensation_executor"},
		RunnerEffect:              "none",
		PromptEffect:              "none",
	}
}
