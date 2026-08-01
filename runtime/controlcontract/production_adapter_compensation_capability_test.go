package controlcontract

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestProductionAdapterCompensationCapabilityPlanReady(t *testing.T) {
	plan := BuildProductionAdapterCompensationCapabilityPlan(productionAdapterCompensationCapabilityPlanInput())
	if plan.Status != HostActionReady ||
		!plan.ReadyForCompensationCapabilities ||
		!plan.ReadyForCompensationExecutorDescriptor ||
		!plan.AllRequiredCapabilitiesDeclared ||
		plan.ResidualRiskDeclared ||
		len(plan.Declarations) != len(KnownProductionAdapterCompensableEffectKinds()) ||
		len(plan.CompensationRefs) != len(KnownProductionAdapterCompensableEffectKinds()) ||
		plan.ExecutorDescriptor.Status != HostActionReady ||
		len(plan.ExecutorDescriptor.SupportedCompensationRefs) != len(KnownProductionAdapterCompensableEffectKinds()) ||
		plan.NextHostAction != "host_may_bind_compensation_executor_descriptor" {
		t.Fatalf("unexpected compensation capability plan: %#v", plan)
	}
	for _, kind := range KnownProductionAdapterCompensableEffectKinds() {
		compensationRef := DisplaySafeRef("compensation:" + string(kind))
		if !displaySafeRefSliceContains(plan.ExecutorDescriptor.SupportedCompensationRefs, compensationRef) {
			t.Fatalf("descriptor missing compensation ref %q in %#v", compensationRef, plan.ExecutorDescriptor.SupportedCompensationRefs)
		}
	}
	for _, declaration := range plan.Declarations {
		if declaration.Status != HostActionReady ||
			!declaration.ReadyForCompensationPlan ||
			!declaration.CompensationCapabilityDeclared ||
			declaration.ResidualRiskDeclared ||
			declaration.CompensationCapabilityRef == "" ||
			declaration.CompensationExecutorRef == "" {
			t.Fatalf("unexpected compensation declaration: %#v", declaration)
		}
		if declaration.CoreExecutionExecuted ||
			declaration.CompensationExecutedByCore ||
			declaration.RunnerDispatched ||
			declaration.ToolExecuted ||
			declaration.WorkflowDispatched ||
			declaration.SchedulerApplied ||
			declaration.InstallerExecuted ||
			declaration.StoreMutationExecuted {
			t.Fatalf("declaration must not execute side effects: %#v", declaration)
		}
	}
	if plan.CoreExecutionExecuted ||
		plan.CompensationExecutedByCore ||
		plan.RunnerDispatched ||
		plan.ToolExecuted ||
		plan.WorkflowDispatched ||
		plan.SchedulerApplied ||
		plan.InstallerExecuted ||
		plan.StoreMutationExecuted {
		t.Fatalf("plan must not execute side effects: %#v", plan)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "compensation capability plan",
		RunnerEffect: plan.RunnerEffect,
		PromptEffect: plan.PromptEffect,
		Boundaries:   plan.Boundaries,
		Payload:      plan,
	}, "production_adapter_compensation_capability_plan", "host_owned_compensation_capability_plan", "compensation_capability_declaration_only", "no_compensation_execution_by_core")
}

func TestProductionAdapterCompensationCapabilityPlanRequiresEveryCompensableEffect(t *testing.T) {
	input := productionAdapterCompensationCapabilityPlanInput()
	input.CapabilitySpecs = input.CapabilitySpecs[:1]
	plan := BuildProductionAdapterCompensationCapabilityPlan(input)
	if plan.Status != HostActionBlocked ||
		plan.ReadyForCompensationCapabilities ||
		plan.AllRequiredCapabilitiesDeclared ||
		!controlTokenListContains(plan.BlockedReasons, "installer_apply_compensation_capability_missing") ||
		!controlTokenListContains(plan.BlockedReasons, "store_mutation_compensation_capability_missing") ||
		!controlTokenListContains(plan.BlockedReasons, "workflow_runtime_compensation_capability_missing") ||
		!controlTokenListContains(plan.BlockedReasons, "delegation_worker_runtime_compensation_capability_missing") ||
		!missingInputContains(plan.MissingInputs, "host:store_mutation_compensation_capability_ref") ||
		!missingInputContains(plan.MissingInputs, "host:workflow_runtime_residual_risk_ref") {
		t.Fatalf("expected missing compensation capabilities to block: %#v", plan)
	}
}

func TestProductionAdapterCompensationCapabilityPlanRecordsResidualRiskForMissingCapability(t *testing.T) {
	input := productionAdapterCompensationCapabilityPlanInput()
	for i := range input.CapabilitySpecs {
		if input.CapabilitySpecs[i].EffectKind == ProductionAdapterCompensableEffectStoreMutation {
			input.CapabilitySpecs[i].CompensationCapabilityRef = ""
			input.CapabilitySpecs[i].CompensationRef = ""
			input.CapabilitySpecs[i].CompensationExecutorRef = ""
			input.CapabilitySpecs[i].CompensationPolicyRef = ""
			input.CapabilitySpecs[i].CompensationApprovalRef = ""
			input.CapabilitySpecs[i].IdempotencyContractRef = ""
			input.CapabilitySpecs[i].ReadbackContractRef = ""
			input.CapabilitySpecs[i].RollbackContractRef = ""
			input.CapabilitySpecs[i].ResidualRiskRef = "residual_risk:store_mutation"
		}
	}
	plan := BuildProductionAdapterCompensationCapabilityPlan(input)
	if plan.Status != HostActionReviewRequired ||
		plan.ReadyForCompensationCapabilities ||
		plan.ReadyForCompensationExecutorDescriptor ||
		plan.AllRequiredCapabilitiesDeclared ||
		!plan.ResidualRiskDeclared ||
		len(plan.ResidualRiskRefs) != 1 ||
		plan.ResidualRiskRefs[0] != "residual_risk:store_mutation" ||
		len(plan.MissingInputs) != 0 ||
		len(plan.BlockedReasons) != 0 ||
		plan.NextHostAction != "review_compensation_residual_risk" {
		t.Fatalf("expected residual-risk review, got %#v", plan)
	}
	var foundStore bool
	for _, declaration := range plan.Declarations {
		if declaration.EffectKind == ProductionAdapterCompensableEffectStoreMutation {
			foundStore = true
			if declaration.Status != HostActionRecorded ||
				declaration.ReadyForCompensationPlan ||
				declaration.CompensationCapabilityDeclared ||
				!declaration.ResidualRiskDeclared ||
				declaration.NextHostAction != "review_store_mutation_compensation_residual_risk" {
				t.Fatalf("unexpected store residual-risk declaration: %#v", declaration)
			}
		}
	}
	if !foundStore {
		t.Fatalf("store mutation declaration missing from %#v", plan.Declarations)
	}
}

func TestProductionAdapterCompensationCapabilityPlanRejectsUnsafeRefsWithoutLeak(t *testing.T) {
	input := productionAdapterCompensationCapabilityPlanInput()
	rawRef := "https://example.invalid/raw/compensation-capability"
	input.CapabilitySpecs[0].CompensationRef = DisplaySafeRef(rawRef)
	plan := BuildProductionAdapterCompensationCapabilityPlan(input)
	if plan.Status != HostActionBlocked ||
		plan.ReadyForCompensationCapabilities ||
		plan.FailureClass != FailureEvidenceWeak ||
		!plan.RawOutputLoaded ||
		!missingInputContains(plan.MissingInputs, "host:display_safe_refs") ||
		!controlTokenListContains(plan.BlockedReasons, "unsafe_input_ref") {
		t.Fatalf("expected unsafe compensation capability block, got %#v", plan)
	}
	payload, err := json.Marshal(plan)
	if err != nil {
		t.Fatalf("marshal plan: %v", err)
	}
	if strings.Contains(string(payload), rawRef) {
		t.Fatalf("plan leaked unsafe ref %q in %s", rawRef, payload)
	}
}

func TestNormalizeProductionAdapterCompensableEffectKindAliases(t *testing.T) {
	cases := map[string]ProductionAdapterCompensableEffectKind{
		"operations_schedule_apply": ProductionAdapterCompensableEffectSchedulerApply,
		"capability_install_apply":  ProductionAdapterCompensableEffectInstallerApply,
		"objective_run_store":       ProductionAdapterCompensableEffectStoreMutation,
		"runtime_executor":          ProductionAdapterCompensableEffectWorkflowRuntime,
		"worker_dispatch":           ProductionAdapterCompensableEffectDelegationWorker,
	}
	for raw, want := range cases {
		if got := NormalizeProductionAdapterCompensableEffectKind(raw); got != want {
			t.Fatalf("NormalizeProductionAdapterCompensableEffectKind(%q)=%q, want %q", raw, got, want)
		}
	}
}

func productionAdapterCompensationCapabilityPlanInput() ProductionAdapterCompensationCapabilityPlanInput {
	return ProductionAdapterCompensationCapabilityPlanInput{
		PlanRef:                "compensation_capability_plan:production_effects",
		ExecutorDescriptorRef:  "compensation_descriptor:production_effects",
		ExecutorRef:            "compensation_executor:production_effects",
		OwnerRef:               "owner:product_host",
		IdempotencyContractRef: "contract:compensation_idempotency",
		ReadbackContractRef:    "contract:compensation_readback",
		RollbackContractRef:    "contract:compensation_rollback",
		TimeoutPolicyRef:       "policy:compensation_timeout",
		RequiredPolicyRefs:     []DisplaySafeRef{"policy:compensation_allowed"},
		RequiredApprovalRefs:   []DisplaySafeRef{"approval:host_compensation"},
		CapabilitySpecs:        productionAdapterCompensationCapabilitySpecs(),
	}
}

func productionAdapterCompensationCapabilitySpecs() []ProductionAdapterCompensationCapabilitySpec {
	out := make([]ProductionAdapterCompensationCapabilitySpec, 0, len(KnownProductionAdapterCompensableEffectKinds()))
	for _, kind := range KnownProductionAdapterCompensableEffectKinds() {
		prefix := string(kind)
		out = append(out, ProductionAdapterCompensationCapabilitySpec{
			EffectKind:                kind,
			EffectRef:                 DisplaySafeRef("effect:" + prefix),
			EffectGateRef:             DisplaySafeRef("gate:" + prefix),
			AdapterRef:                DisplaySafeRef("adapter:" + prefix),
			CompensationCapabilityRef: DisplaySafeRef("compensation_capability:" + prefix),
			CompensationRef:           DisplaySafeRef("compensation:" + prefix),
			CompensationExecutorRef:   "compensation_executor:production_effects",
			CompensationPolicyRef:     "policy:compensation_allowed",
			CompensationApprovalRef:   "approval:host_compensation",
			IdempotencyContractRef:    "contract:compensation_idempotency",
			ReadbackContractRef:       "contract:compensation_readback",
			RollbackContractRef:       "contract:compensation_rollback",
			FailureReviewRef:          DisplaySafeRef("failure_review:" + prefix),
			EvidenceRefs:              []DisplaySafeRef{DisplaySafeRef("evidence:" + prefix)},
		})
	}
	return out
}
