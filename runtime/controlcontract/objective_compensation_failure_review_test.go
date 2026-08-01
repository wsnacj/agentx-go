package controlcontract

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestObjectiveCompensationFailureReviewPacketReadyFromFailedReadback(t *testing.T) {
	result := objectiveCompensationFailureResult(objectiveCompensationReadyRequest())
	readback := objectiveCompensationFailureReadback(result)
	packet := BuildObjectiveCompensationFailureReviewPacket(ObjectiveCompensationFailureReviewPacketInput{
		CompensationFailureReviewPacketRef: "review:compensation_failure_closeout",
		HostReviewRef:                      "host_review:compensation_failure_closeout",
		Result:                             result,
		Readback:                           readback,
		EvidenceRefs: []EvidenceRef{{
			Ref:      "evidence:compensation_failure_review",
			Kind:     "failure_review",
			Strength: EvidenceAdequate,
		}},
	})
	if packet.Status != HostActionReviewRequired ||
		!packet.ReadyForHostDisplay ||
		!packet.ReadyForCloseoutFailureReview ||
		!packet.ReadyForCompensationFailureReview ||
		!packet.CompensationFailureRecorded ||
		!packet.ResidualRiskRecorded ||
		packet.FailureClass != FailureNone ||
		packet.FailureReviewPacketRef != result.FailureReviewPacketRef ||
		packet.CompensationResultRef != result.CompensationResultRef ||
		packet.CompensationReadbackRef != readback.CompensationReadbackRef ||
		packet.ResidualRiskRef != result.ResidualRiskRef ||
		packet.ObservedResidualRiskRef != readback.ObservedResidualRiskRef ||
		packet.NextHostAction != "review_compensation_failure_closeout" {
		t.Fatalf("unexpected compensation failure review packet: %#v", packet)
	}
	if packet.CoreExecutionExecuted ||
		packet.CompensationExecutedByCore ||
		packet.RunnerDispatched ||
		packet.ToolExecuted ||
		packet.WorkflowDispatched ||
		packet.SchedulerApplied ||
		packet.InstallerExecuted ||
		packet.StoreMutationExecuted {
		t.Fatalf("core must not execute compensation failure review side effects: %#v", packet)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "compensation failure review packet",
		RunnerEffect: packet.RunnerEffect,
		PromptEffect: packet.PromptEffect,
		Boundaries:   packet.Boundaries,
		Payload:      packet,
	}, "objective_compensation_failure_review_packet", "host_owned_compensation_failure_review", "compensation_failure_to_closeout_failure_review", "ready_for_compensation_failure_closeout_review", "no_compensation_execution_by_core")
}

func TestObjectiveCompensationFailureReviewPacketReadyFromFailedResult(t *testing.T) {
	result := objectiveCompensationFailureResult(objectiveCompensationReadyRequest())
	packet := BuildObjectiveCompensationFailureReviewPacket(ObjectiveCompensationFailureReviewPacketInput{
		CompensationFailureReviewPacketRef: "review:compensation_failure_closeout",
		HostReviewRef:                      "host_review:compensation_failure_closeout",
		Result:                             result,
	})
	if packet.Status != HostActionReviewRequired ||
		!packet.ReadyForCloseoutFailureReview ||
		packet.CompensationReadbackRef != "" ||
		packet.ResidualRiskRef != "residual_risk:manual_store_review" {
		t.Fatalf("unexpected result-only compensation failure review packet: %#v", packet)
	}
}

func TestObjectiveCompensationFailureReviewPacketBlocksSuccessfulCompensation(t *testing.T) {
	result := objectiveCompensationSuccessResult(objectiveCompensationReadyRequest())
	readback := BuildObjectiveCompensationExecutionReadback(ObjectiveCompensationExecutionReadbackInput{
		Result:                  result,
		CompensationReadbackRef: result.ExpectedReadbackRef,
		ObservedCompensationRef: result.AppliedCompensationRef,
		ObservedHostRunRef:      result.HostCompensationRunRef,
	})
	packet := BuildObjectiveCompensationFailureReviewPacket(ObjectiveCompensationFailureReviewPacketInput{
		CompensationFailureReviewPacketRef: "review:compensation_failure_closeout",
		HostReviewRef:                      "host_review:compensation_failure_closeout",
		Result:                             result,
		Readback:                           readback,
	})
	if packet.Status != HostActionBlocked ||
		packet.ReadyForCloseoutFailureReview ||
		packet.ReadyForCompensationFailureReview ||
		packet.FailureClass != FailureVerificationFailed ||
		!controlTokenListContains(packet.BlockedReasons, "compensation_failure_not_present") ||
		!missingInputContains(packet.MissingInputs, "host:compensation_failure_or_residual_risk") ||
		packet.NextHostAction != "continue_closeout_review" {
		t.Fatalf("expected successful compensation to skip failure review, got %#v", packet)
	}
}

func TestObjectiveCompensationFailureReviewPacketBlocksUnreadyResult(t *testing.T) {
	request := objectiveCompensationReadyRequest()
	result := BuildObjectiveCompensationExecutionResult(ObjectiveCompensationExecutionResultInput{
		Request:                  request,
		CompensationResultRef:    request.ExpectedResultRef,
		HostCompensationRunRef:   "compensation_run:store_revert_failed",
		HostCompensationReported: true,
		HostCompensationFailed:   true,
		FailureRef:               "failure:compensation_store_revert",
	})
	packet := BuildObjectiveCompensationFailureReviewPacket(ObjectiveCompensationFailureReviewPacketInput{
		CompensationFailureReviewPacketRef: "review:compensation_failure_closeout",
		HostReviewRef:                      "host_review:compensation_failure_closeout",
		Result:                             result,
	})
	if packet.Status != HostActionBlocked ||
		packet.ReadyForCloseoutFailureReview ||
		packet.FailureClass != FailureEvidenceMissing ||
		!controlTokenListContains(packet.BlockedReasons, "compensation_result_not_ready_for_failure_review") ||
		!missingInputContains(packet.MissingInputs, "host:compensation_result") {
		t.Fatalf("expected unready result block, got %#v", packet)
	}
}

func TestObjectiveCompensationFailureReviewPacketRejectsUnsafeRefWithoutLeak(t *testing.T) {
	result := objectiveCompensationFailureResult(objectiveCompensationReadyRequest())
	rawRef := "https://example.invalid/raw/compensation-failure-review"
	packet := BuildObjectiveCompensationFailureReviewPacket(ObjectiveCompensationFailureReviewPacketInput{
		CompensationFailureReviewPacketRef: DisplaySafeRef(rawRef),
		HostReviewRef:                      "host_review:compensation_failure_closeout",
		Result:                             result,
	})
	if packet.Status != HostActionBlocked ||
		packet.ReadyForCloseoutFailureReview ||
		packet.FailureClass != FailureEvidenceWeak ||
		!packet.RawOutputLoaded ||
		!missingInputContains(packet.MissingInputs, "host:display_safe_refs") ||
		!controlTokenListContains(packet.BlockedReasons, "unsafe_input_ref") {
		t.Fatalf("expected unsafe compensation failure review block, got %#v", packet)
	}
	payload, err := json.Marshal(packet)
	if err != nil {
		t.Fatalf("marshal packet: %v", err)
	}
	if strings.Contains(string(payload), rawRef) {
		t.Fatalf("packet leaked unsafe ref %q in %s", rawRef, payload)
	}
}

func objectiveCompensationFailureResult(request ObjectiveCompensationExecutionRequest) ObjectiveCompensationExecutionResult {
	return BuildObjectiveCompensationExecutionResult(ObjectiveCompensationExecutionResultInput{
		Request:                  request,
		CompensationResultRef:    request.ExpectedResultRef,
		HostCompensationRunRef:   "compensation_run:store_revert_failed",
		HostCompensationReported: true,
		HostCompensationFailed:   true,
		FailureRef:               "failure:compensation_store_revert",
		ResidualRiskRef:          "residual_risk:manual_store_review",
		CompensationEvidenceRefs: []EvidenceRef{{
			Ref:      "evidence:compensation_failure",
			Kind:     "host_report",
			Strength: EvidenceAdequate,
		}},
	})
}

func objectiveCompensationFailureReadback(result ObjectiveCompensationExecutionResult) ObjectiveCompensationExecutionReadback {
	return BuildObjectiveCompensationExecutionReadback(ObjectiveCompensationExecutionReadbackInput{
		Result:                  result,
		CompensationReadbackRef: result.ExpectedReadbackRef,
		ObservedHostRunRef:      result.HostCompensationRunRef,
		ObservedResidualRiskRef: result.ResidualRiskRef,
		ReadbackEvidenceRefs: []EvidenceRef{{
			Ref:      "evidence:residual_risk_readback",
			Kind:     "host_readback",
			Strength: EvidenceAdequate,
		}},
	})
}
