package controlcontract_test

import (
	"testing"

	controlcontract "github.com/wsnacj/agentx-go/runtime/controlcontract"
)

func TestHostEffectGateIsUsableFromExternalPackage(t *testing.T) {
	gate := controlcontract.BuildProductionAdapterIndependentEffectGate(controlcontract.ProductionAdapterIndependentEffectGateSpec{
		Kind:                  controlcontract.ProductionAdapterEffectGateRuntimeExecutor,
		GateRef:               "gate:runtime_executor",
		AdapterRef:            "adapter:runtime_executor",
		ContractRef:           "contract:runtime_executor",
		PolicyRef:             "policy:runtime_executor",
		ApprovalRef:           "approval:runtime_executor",
		BudgetRef:             "budget:runtime_executor",
		IdempotencyRef:        "idempotency:runtime_executor",
		ReadbackRef:           "readback:runtime_executor",
		EvalRef:               "eval:runtime_executor",
		FailureReviewRef:      "review:runtime_executor_failure",
		CompensationReviewRef: "review:runtime_executor_compensation",
	})
	if !gate.ReadyForIndependentGatePlan || gate.Status != controlcontract.HostActionReady {
		t.Fatalf("external gate must be ready: %#v", gate)
	}

	blocked := controlcontract.BuildProductionAdapterIndependentEffectGatePlan(controlcontract.ProductionAdapterIndependentEffectGatePlanInput{
		PlanRef:                    "plan:host_effects",
		AggregateExecutorRequested: true,
		AggregateExecutorRef:       "executor:aggregate",
	})
	if blocked.ReadyForIndependentGatePlan || !blocked.AggregateExecutorBlocked || blocked.Status != controlcontract.HostActionBlocked {
		t.Fatalf("aggregate executor must fail closed: %#v", blocked)
	}
}
