package controlcontract

import (
	"testing"
)

func TestProductionAdapterIndependentEffectGatePlanReady(t *testing.T) {
	plan := BuildProductionAdapterIndependentEffectGatePlan(ProductionAdapterIndependentEffectGatePlanInput{
		PlanRef:   "plan:independent_effect_gates",
		GateSpecs: productionAdapterIndependentEffectGateSpecs(),
	})
	if plan.Status != HostActionReady ||
		!plan.ReadyForIndependentGatePlan ||
		plan.AggregateExecutorBlocked ||
		len(plan.Gates) != len(KnownProductionAdapterEffectGateKinds()) ||
		len(plan.MissingInputs) != 0 ||
		len(plan.BlockedReasons) != 0 ||
		plan.NextHostAction != "host_may_implement_independent_effect_gates" {
		t.Fatalf("expected ready independent effect gate plan: %#v", plan)
	}
	for _, gate := range plan.Gates {
		if !gate.ReadyForIndependentGatePlan ||
			gate.Status != HostActionReady ||
			gate.GateRef == "" ||
			gate.ContractRef == "" ||
			gate.ApprovalRef == "" ||
			gate.BudgetRef == "" ||
			gate.ReadbackRef == "" ||
			gate.EvalRef == "" {
			t.Fatalf("expected ready gate row: %#v in plan %#v", gate, plan)
		}
	}
	for _, boundary := range []Boundary{
		"no_unified_auto_executor",
		"no_scheduler_apply",
		"no_installer_apply",
		"no_workflow_retry_apply",
		"no_runtime_executor",
		"no_compensation_executor",
		"no_memory_apply",
		"independent_effect_gate_plan_ready",
	} {
		if !productionAdapterBoundaryContains(plan.Boundaries, boundary) {
			t.Fatalf("expected boundary %q in %#v", boundary, plan.Boundaries)
		}
	}
	assertHostOwnedProjectionOnly(t, testProjection[Boundary]{
		RunnerEffect: plan.RunnerEffect,
		PromptEffect: plan.PromptEffect,
		Boundaries:   plan.Boundaries,
	})
}

func TestProductionAdapterIndependentEffectGatePlanBlocksAggregateExecutor(t *testing.T) {
	plan := BuildProductionAdapterIndependentEffectGatePlan(ProductionAdapterIndependentEffectGatePlanInput{
		PlanRef:                      "plan:independent_effect_gates",
		GateSpecs:                    productionAdapterIndependentEffectGateSpecs(),
		AggregateExecutorRequested:   true,
		AggregateExecutorRef:         "executor:aggregate_effects",
		AggregateExecutorPolicyRef:   "policy:aggregate_effects",
		AggregateExecutorApprovalRef: "approval:aggregate_effects",
	})
	if plan.Status != HostActionBlocked ||
		plan.ReadyForIndependentGatePlan ||
		!plan.AggregateExecutorBlocked ||
		!productionAdapterStringContains(plan.BlockedReasons, "aggregate_effect_executor_not_allowed") ||
		!productionAdapterMissingContains(plan.MissingInputs, "host:independent_effect_gate") ||
		!productionAdapterBoundaryContains(plan.Boundaries, "aggregate_effect_executor_blocked") ||
		!productionAdapterBoundaryContains(plan.Boundaries, "no_unified_auto_executor") ||
		plan.RunnerEffect != "none" ||
		plan.PromptEffect != "none" {
		t.Fatalf("expected aggregate executor to be blocked: %#v", plan)
	}
}

func TestProductionAdapterIndependentEffectGatePlanRequiresEveryGate(t *testing.T) {
	specs := productionAdapterIndependentEffectGateSpecs()
	plan := BuildProductionAdapterIndependentEffectGatePlan(ProductionAdapterIndependentEffectGatePlanInput{
		PlanRef:   "plan:independent_effect_gates",
		GateSpecs: specs[:1],
	})
	if plan.Status != HostActionBlocked ||
		plan.ReadyForIndependentGatePlan ||
		!productionAdapterStringContains(plan.BlockedReasons, "installer_apply_gate_ref_missing") ||
		!productionAdapterStringContains(plan.BlockedReasons, "workflow_retry_apply_gate_ref_missing") ||
		!productionAdapterStringContains(plan.BlockedReasons, "runtime_executor_gate_ref_missing") ||
		!productionAdapterStringContains(plan.BlockedReasons, "compensation_executor_gate_ref_missing") ||
		!productionAdapterStringContains(plan.BlockedReasons, "memory_apply_gate_ref_missing") ||
		!productionAdapterMissingContains(plan.MissingInputs, "host:installer_apply_gate_ref") ||
		!productionAdapterMissingContains(plan.MissingInputs, "host:memory_apply_gate_ref") ||
		len(plan.Gates) != len(KnownProductionAdapterEffectGateKinds()) {
		t.Fatalf("expected missing independent gates to block: %#v", plan)
	}
	if !plan.Gates[0].ReadyForIndependentGatePlan {
		t.Fatalf("provided scheduler gate should remain ready: %#v", plan.Gates[0])
	}
	for _, gate := range plan.Gates[1:] {
		if gate.ReadyForIndependentGatePlan || gate.GateRef != "" {
			t.Fatalf("missing gate row should stay blocked and empty: %#v", gate)
		}
	}
}

func TestProductionAdapterIndependentEffectGatePlanBlocksUnsafeRefs(t *testing.T) {
	specs := productionAdapterIndependentEffectGateSpecs()
	specs[0].GateRef = "/tmp/scheduler-gate"
	plan := BuildProductionAdapterIndependentEffectGatePlan(ProductionAdapterIndependentEffectGatePlanInput{
		PlanRef:   "plan:independent_effect_gates",
		GateSpecs: specs,
	})
	if plan.Status != HostActionBlocked ||
		plan.ReadyForIndependentGatePlan ||
		!productionAdapterStringContains(plan.BlockedReasons, "unsafe_input_ref") ||
		!productionAdapterMissingContains(plan.MissingInputs, "host:display_safe_refs") ||
		!productionAdapterBoundaryContains(plan.Boundaries, "raw_output_not_allowed") ||
		plan.RunnerEffect != "none" ||
		plan.PromptEffect != "none" {
		t.Fatalf("expected unsafe independent gate refs to block: %#v", plan)
	}
}

func TestProductionAdapterIndependentEffectGatePlanBlocksDuplicateKinds(t *testing.T) {
	specs := productionAdapterIndependentEffectGateSpecs()
	specs = append(specs, specs[0])
	specs[len(specs)-1].GateRef = "gate:scheduler_apply_duplicate"
	plan := BuildProductionAdapterIndependentEffectGatePlan(ProductionAdapterIndependentEffectGatePlanInput{
		PlanRef:   "plan:independent_effect_gates",
		GateSpecs: specs,
	})
	if plan.Status != HostActionBlocked ||
		plan.ReadyForIndependentGatePlan ||
		!productionAdapterStringContains(plan.BlockedReasons, "scheduler_apply_gate_duplicate") ||
		!productionAdapterMissingContains(plan.MissingInputs, "host:scheduler_apply_gate_ref") {
		t.Fatalf("expected duplicate independent gate kind to block: %#v", plan)
	}
}

func productionAdapterIndependentEffectGateSpecs() []ProductionAdapterIndependentEffectGateSpec {
	out := make([]ProductionAdapterIndependentEffectGateSpec, 0, len(KnownProductionAdapterEffectGateKinds()))
	for _, kind := range KnownProductionAdapterEffectGateKinds() {
		prefix := string(kind)
		out = append(out, ProductionAdapterIndependentEffectGateSpec{
			Kind:                  kind,
			GateRef:               DisplaySafeRef("gate:" + prefix),
			AdapterRef:            DisplaySafeRef("adapter:" + prefix),
			ContractRef:           DisplaySafeRef("contract:" + prefix),
			PolicyRef:             DisplaySafeRef("policy:" + prefix),
			ApprovalRef:           DisplaySafeRef("approval:" + prefix),
			BudgetRef:             DisplaySafeRef("budget:" + prefix),
			IdempotencyRef:        DisplaySafeRef("idempotency:" + prefix),
			ReadbackRef:           DisplaySafeRef("readback:" + prefix),
			EvalRef:               DisplaySafeRef("eval:" + prefix),
			FailureReviewRef:      DisplaySafeRef("review:" + prefix + "_failure"),
			CompensationReviewRef: DisplaySafeRef("review:" + prefix + "_compensation"),
			EvidenceRefs: []DisplaySafeRef{
				DisplaySafeRef("evidence:" + prefix),
			},
		})
	}
	return out
}
