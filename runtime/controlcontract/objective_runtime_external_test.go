package controlcontract_test

import (
	"testing"

	controlcontract "github.com/wsnacj/agentx-go/runtime/controlcontract"
)

func TestObjectiveRuntimeContractsFailClosedWithoutHostInputs(t *testing.T) {
	policy := controlcontract.BuildAutoDelegationPolicyReview(controlcontract.AutoDelegationPolicy{})
	if !policy.Projected || policy.AutoDelegationAllowed || policy.RunnerEffect != "none" {
		t.Fatalf("unexpected default policy review: %#v", policy)
	}

	step := controlcontract.BuildObjectiveRuntimeLoopStep(controlcontract.ObjectiveRuntimeLoopInput{})
	if !step.Projected || step.Available || step.Status != "inactive" || step.RunnerEffect != "none" {
		t.Fatalf("unexpected hostless runtime step: %#v", step)
	}

	request := controlcontract.BuildHostOwnedObjectiveExecutorStepRequest(controlcontract.HostOwnedObjectiveExecutorStepRequestInput{
		RuntimeLoop: step,
	})
	if !request.Projected || !request.RequestOnly || request.ReadyForHostExecution || request.RunnerEffect != "none" {
		t.Fatalf("unexpected hostless executor request: %#v", request)
	}

	report := controlcontract.BuildObjectiveRuntimeProductization(controlcontract.ObjectiveRuntimeProductizationInput{
		RuntimeLoop: step,
	})
	if !report.Projected || report.Available || report.ReadyForHostProductization || report.RunnerEffect != "none" {
		t.Fatalf("unexpected hostless productization report: %#v", report)
	}
}

func TestStructuredObservationNormalizationRejectsMissingObservation(t *testing.T) {
	result := controlcontract.BuildStructuredObservationNormalization(controlcontract.StructuredObservationNormalizationInput{
		SourceKind: "delegation_worker_result",
		SourceRef:  "result:worker-1",
	})
	if result.ReadyForVerification || result.Status != controlcontract.VerificationBlocked || result.RunnerEffect != "none" {
		t.Fatalf("unexpected empty normalization result: %#v", result)
	}
}
