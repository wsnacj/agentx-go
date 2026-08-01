package controlcontract

import (
	"encoding/json"
	"testing"
)

func TestProductionAdapterObjectiveCloseoutWriterDescriptorReady(t *testing.T) {
	descriptor := productionAdapterReadyObjectiveCloseoutWriterDescriptor()
	if descriptor.Status != HostActionReady ||
		!descriptor.ReadyForWriterOptIn ||
		descriptor.WriterRef != "writer:metrics_objective_closeout" ||
		!productionAdapterObjectiveCloseoutWriterModeContains(descriptor.SupportedModes, ProductionAdapterObjectiveCloseoutWriterPlanOnly) ||
		!productionAdapterObjectiveCloseoutWriterModeContains(descriptor.SupportedModes, ProductionAdapterObjectiveCloseoutWriterDryRun) ||
		!productionAdapterObjectiveCloseoutWriterModeContains(descriptor.SupportedModes, ProductionAdapterObjectiveCloseoutWriterDurableWrite) ||
		descriptor.RunnerEffect != "none" ||
		descriptor.PromptEffect != "none" {
		t.Fatalf("unexpected objective closeout writer descriptor: %#v", descriptor)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "objective closeout writer descriptor",
		RunnerEffect: descriptor.RunnerEffect,
		PromptEffect: descriptor.PromptEffect,
		Boundaries:   descriptor.Boundaries,
		Payload:      descriptor,
	}, "production_adapter_objective_closeout_writer_descriptor", "objective_closeout_writer_descriptor_projection_only", "host_owned_objective_closeout_writer", "explicit_opt_in_required", "no_durable_write_by_core", "no_runstore_write_by_core")
}

func TestProductionAdapterObjectiveCloseoutWriterOptInPlanOnly(t *testing.T) {
	handoff, uiHandoff := productionAdapterReadyObjectiveCloseoutWriterInputs()
	descriptor := productionAdapterReadyObjectiveCloseoutWriterDescriptor()

	optIn := BuildProductionAdapterObjectiveCloseoutWriterOptIn(ProductionAdapterObjectiveCloseoutWriterOptInInput{
		WriterOptInRef:   "optin:metrics_objective_closeout_plan",
		WriterDescriptor: descriptor,
		DurableHandoff:   handoff,
		HostUIHandoff:    uiHandoff,
		RequestedMode:    ProductionAdapterObjectiveCloseoutWriterPlanOnly,
	})
	if optIn.Status != HostActionReviewRequired ||
		!optIn.ReadyForHostWriterPlan ||
		optIn.ReadyForHostWriterDryRun ||
		optIn.ReadyForHostDurableWrite ||
		optIn.HostMayExecuteDurableWrite ||
		!optIn.PlanOnly ||
		optIn.DryRun ||
		optIn.DurableWriteMode ||
		optIn.CoreInvocationExecuted ||
		optIn.DurableWriteByCore ||
		optIn.ObjectiveStoreWriteByCore ||
		optIn.RunstoreWriteByCore ||
		optIn.NextHostAction != "review_objective_closeout_writer_plan" {
		t.Fatalf("unexpected plan-only objective closeout writer opt-in: %#v", optIn)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "objective closeout writer plan-only opt-in",
		RunnerEffect: optIn.RunnerEffect,
		PromptEffect: optIn.PromptEffect,
		Boundaries:   optIn.Boundaries,
		Payload:      optIn,
	}, "production_adapter_objective_closeout_writer_opt_in", "objective_closeout_writer_opt_in_projection_only", "objective_closeout_writer_plan_only", "durable_write_not_enabled")
	AssertNoCoreMutation(t, "objective closeout writer plan-only opt-in", optIn.CoreInvocationExecuted, optIn.DurableWriteByCore)
}

func TestProductionAdapterObjectiveCloseoutWriterOptInDryRun(t *testing.T) {
	handoff, uiHandoff := productionAdapterReadyObjectiveCloseoutWriterInputs()
	descriptor := productionAdapterReadyObjectiveCloseoutWriterDescriptor()

	optIn := BuildProductionAdapterObjectiveCloseoutWriterOptIn(ProductionAdapterObjectiveCloseoutWriterOptInInput{
		WriterOptInRef:          "optin:metrics_objective_closeout_dry_run",
		WriterDescriptor:        descriptor,
		DurableHandoff:          handoff,
		HostUIHandoff:           uiHandoff,
		RequestedMode:           ProductionAdapterObjectiveCloseoutWriterDryRun,
		ExplicitOptIn:           true,
		HostWriterBindingRef:    "binding:metrics_objective_closeout_writer",
		HostWriterAvailable:     true,
		AvailableCapabilityRefs: descriptor.RequiresCapabilityRefs,
		PolicyRefs:              descriptor.RequiredPolicyRefs,
		ApprovalRefs:            descriptor.RequiredApprovalRefs,
		BudgetRef:               descriptor.RequiredBudgetRef,
		IdempotencyRef:          "idempotency:metrics_objective_closeout_request",
		DryRunPlanRef:           "dryrun_plan:metrics_objective_closeout",
	})
	if optIn.Status != HostActionReady ||
		!optIn.ReadyForHostWriterDryRun ||
		optIn.ReadyForHostDurableWrite ||
		optIn.HostMayExecuteDurableWrite ||
		!optIn.DryRun ||
		optIn.PlanOnly ||
		optIn.DurableWriteMode ||
		optIn.NextHostAction != "host_may_run_objective_closeout_writer_dry_run" {
		t.Fatalf("unexpected dry-run objective closeout writer opt-in: %#v", optIn)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "objective closeout writer dry-run opt-in",
		RunnerEffect: optIn.RunnerEffect,
		PromptEffect: optIn.PromptEffect,
		Boundaries:   optIn.Boundaries,
		Payload:      optIn,
	}, "production_adapter_objective_closeout_writer_opt_in", "objective_closeout_writer_dry_run_mode", "objective_closeout_writer_dry_run_ready", "durable_write_not_enabled")
	AssertNoCoreMutation(t, "objective closeout writer dry-run opt-in", optIn.CoreInvocationExecuted, optIn.DurableWriteByCore)
}

func TestProductionAdapterObjectiveCloseoutWriterOptInDurableWriteReady(t *testing.T) {
	handoff, uiHandoff := productionAdapterReadyObjectiveCloseoutWriterInputs()
	descriptor := productionAdapterReadyObjectiveCloseoutWriterDescriptor()

	optIn := productionAdapterReadyObjectiveCloseoutWriterOptIn(descriptor, handoff, uiHandoff)
	if optIn.Status != HostActionReady ||
		!optIn.ReadyForHostDurableWrite ||
		!optIn.HostMayExecuteDurableWrite ||
		optIn.ReadyForHostWriterPlan ||
		optIn.ReadyForHostWriterDryRun ||
		!optIn.DurableWriteMode ||
		optIn.PlanOnly ||
		optIn.DryRun ||
		!optIn.ExplicitOptIn ||
		optIn.HostWriterBindingRef == "" ||
		optIn.ExpectedReadbackRef == "" ||
		optIn.RollbackReviewRef != descriptor.RollbackReviewRef ||
		optIn.CompensationReviewRef != descriptor.CompensationReviewRef ||
		optIn.CoreInvocationExecuted ||
		optIn.DurableWriteByCore ||
		optIn.ObjectiveStoreWriteByCore ||
		optIn.RunstoreWriteByCore ||
		optIn.NextHostAction != "host_may_execute_objective_closeout_durable_writer" {
		t.Fatalf("unexpected durable-write objective closeout writer opt-in: %#v", optIn)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "objective closeout writer durable-write opt-in",
		RunnerEffect: optIn.RunnerEffect,
		PromptEffect: optIn.PromptEffect,
		Boundaries:   optIn.Boundaries,
		Payload:      optIn,
	}, "production_adapter_objective_closeout_writer_opt_in", "objective_closeout_writer_durable_write_mode", "objective_closeout_writer_explicit_opt_in_confirmed", "ready_for_host_objective_closeout_durable_writer", "host_may_execute_durable_writer", "core_durable_write_not_executed")
	AssertNoCoreMutation(t, "objective closeout writer durable-write opt-in", optIn.CoreInvocationExecuted, optIn.DurableWriteByCore)
}

func TestProductionAdapterObjectiveCloseoutWriterOptInRequiresExplicitOptIn(t *testing.T) {
	handoff, uiHandoff := productionAdapterReadyObjectiveCloseoutWriterInputs()
	descriptor := productionAdapterReadyObjectiveCloseoutWriterDescriptor()

	optIn := productionAdapterReadyObjectiveCloseoutWriterOptIn(descriptor, handoff, uiHandoff)
	optIn = BuildProductionAdapterObjectiveCloseoutWriterOptIn(ProductionAdapterObjectiveCloseoutWriterOptInInput{
		WriterOptInRef:                  optIn.WriterOptInRef,
		WriterDescriptor:                descriptor,
		DurableHandoff:                  handoff,
		HostUIHandoff:                   uiHandoff,
		RequestedMode:                   ProductionAdapterObjectiveCloseoutWriterDurableWrite,
		HostWriterBindingRef:            optIn.HostWriterBindingRef,
		HostWriterAvailable:             true,
		HostReadbackAvailable:           true,
		HostRollbackReviewAvailable:     true,
		HostCompensationReviewAvailable: true,
		AvailableCapabilityRefs:         descriptor.RequiresCapabilityRefs,
		PolicyRefs:                      descriptor.RequiredPolicyRefs,
		ApprovalRefs:                    descriptor.RequiredApprovalRefs,
		BudgetRef:                       descriptor.RequiredBudgetRef,
		IdempotencyRef:                  optIn.IdempotencyRef,
		DryRunPlanRef:                   optIn.DryRunPlanRef,
		DryRunResultRef:                 optIn.DryRunResultRef,
		ExpectedReadbackRef:             optIn.ExpectedReadbackRef,
		RollbackReviewRef:               descriptor.RollbackReviewRef,
		CompensationReviewRef:           descriptor.CompensationReviewRef,
	})
	if optIn.Status != HostActionBlocked ||
		optIn.ReadyForHostDurableWrite ||
		optIn.HostMayExecuteDurableWrite ||
		!productionAdapterStringContains(optIn.BlockedReasons, "objective_closeout_writer_explicit_opt_in_required") ||
		!productionAdapterMissingContains(optIn.MissingInputs, "host:objective_closeout_writer_explicit_opt_in") {
		t.Fatalf("expected explicit opt-in to block durable writer, got %#v", optIn)
	}
}

func TestProductionAdapterObjectiveCloseoutWriterOptInRequiresDryRunAndReadbackForDurableWrite(t *testing.T) {
	handoff, uiHandoff := productionAdapterReadyObjectiveCloseoutWriterInputs()
	descriptor := productionAdapterReadyObjectiveCloseoutWriterDescriptor()

	optIn := productionAdapterReadyObjectiveCloseoutWriterOptIn(descriptor, handoff, uiHandoff)
	optIn = BuildProductionAdapterObjectiveCloseoutWriterOptIn(ProductionAdapterObjectiveCloseoutWriterOptInInput{
		WriterOptInRef:                  optIn.WriterOptInRef,
		WriterDescriptor:                descriptor,
		DurableHandoff:                  handoff,
		HostUIHandoff:                   uiHandoff,
		RequestedMode:                   ProductionAdapterObjectiveCloseoutWriterDurableWrite,
		ExplicitOptIn:                   true,
		HostWriterBindingRef:            optIn.HostWriterBindingRef,
		HostWriterAvailable:             true,
		HostReadbackAvailable:           true,
		HostRollbackReviewAvailable:     true,
		HostCompensationReviewAvailable: true,
		AvailableCapabilityRefs:         descriptor.RequiresCapabilityRefs,
		PolicyRefs:                      descriptor.RequiredPolicyRefs,
		ApprovalRefs:                    descriptor.RequiredApprovalRefs,
		BudgetRef:                       descriptor.RequiredBudgetRef,
		IdempotencyRef:                  optIn.IdempotencyRef,
		DryRunPlanRef:                   optIn.DryRunPlanRef,
		RollbackReviewRef:               descriptor.RollbackReviewRef,
		CompensationReviewRef:           descriptor.CompensationReviewRef,
	})
	if optIn.Status != HostActionBlocked ||
		optIn.ReadyForHostDurableWrite ||
		!productionAdapterStringContains(optIn.BlockedReasons, "writer_dry_run_result_ref_missing") ||
		!productionAdapterStringContains(optIn.BlockedReasons, "writer_expected_readback_ref_missing") ||
		!productionAdapterMissingContains(optIn.MissingInputs, "host:objective_closeout_writer_dry_run_result_ref") ||
		!productionAdapterMissingContains(optIn.MissingInputs, "host:objective_closeout_writer_expected_readback_ref") {
		t.Fatalf("expected durable writer dry-run/readback gates to block, got %#v", optIn)
	}
}

func TestProductionAdapterObjectiveCloseoutWriterJSONCompatibility(t *testing.T) {
	handoff, uiHandoff := productionAdapterReadyObjectiveCloseoutWriterInputs()
	descriptor := productionAdapterReadyObjectiveCloseoutWriterDescriptor()
	optIn := productionAdapterReadyObjectiveCloseoutWriterOptIn(descriptor, handoff, uiHandoff)

	raw, err := json.Marshal(optIn)
	if err != nil {
		t.Fatalf("marshal objective closeout writer opt-in: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("unmarshal objective closeout writer opt-in: %v", err)
	}
	for _, key := range []string{
		"writer_opt_in_ref",
		"writer_ref",
		"requested_mode",
		"ready_for_host_durable_write",
		"host_may_execute_durable_write",
		"dry_run_plan_ref",
		"dry_run_result_ref",
		"expected_readback_ref",
		"rollback_review_ref",
		"compensation_review_ref",
	} {
		if _, ok := payload[key]; !ok {
			t.Fatalf("expected stable JSON key %q in %s", key, string(raw))
		}
	}
	AssertNoRawPayload(t, "objective closeout writer opt-in JSON", optIn, "/Users/mason", "postgresql://secret", "raw local host task")
}

func productionAdapterReadyObjectiveCloseoutWriterDescriptor() ProductionAdapterObjectiveCloseoutWriterDescriptor {
	return BuildProductionAdapterObjectiveCloseoutWriterDescriptor(ProductionAdapterObjectiveCloseoutWriterDescriptor{
		WriterRef:                  "writer:metrics_objective_closeout",
		Owner:                      "host",
		OwnerRef:                   "host:metrics_runtime",
		Version:                    "v1",
		SupportedModes:             []ProductionAdapterObjectiveCloseoutWriterMode{ProductionAdapterObjectiveCloseoutWriterPlanOnly, ProductionAdapterObjectiveCloseoutWriterDryRun, ProductionAdapterObjectiveCloseoutWriterDurableWrite},
		SupportedTargetRefs:        []DisplaySafeRef{"target:objective_lifecycle", "target:runstore", "target:durable_event"},
		RequiresCapabilityRefs:     []DisplaySafeRef{"capability:objective_closeout_writer"},
		RequiredPolicyRefs:         []DisplaySafeRef{"policy:objective_closeout_durable_write"},
		RequiredApprovalRefs:       []DisplaySafeRef{"approval:objective_closeout_durable_write"},
		RequiredBudgetRef:          "budget:objective_closeout_writer",
		IdempotencyContractRef:     "idempotency:objective_closeout_writer",
		InputContractRef:           "contract:objective_closeout_writer_input",
		OutputContractRef:          "contract:objective_closeout_writer_output",
		DryRunContractRef:          "contract:objective_closeout_writer_dry_run",
		ReadbackContractRef:        "contract:objective_closeout_writer_readback",
		RollbackReviewRef:          "rollback:objective_closeout_writer_manual_review",
		CompensationReviewRef:      "compensation:objective_closeout_writer_manual_review",
		RedactionPolicyRef:         "redaction:objective_closeout_writer_display_safe",
		TimeoutPolicyRef:           "timeout:objective_closeout_writer_short",
		PlanOnlyDefault:            true,
		DryRunRequired:             true,
		ReadbackRequired:           true,
		RollbackReviewRequired:     true,
		CompensationReviewRequired: true,
		Boundaries:                 []Boundary{"metrics_objective_closeout_writer_descriptor"},
	})
}

func productionAdapterReadyObjectiveCloseoutWriterInputs() (ProductionAdapterObjectiveCloseoutDurableHandoff, ProductionAdapterObjectiveCloseoutHostUIHandoff) {
	handoff := productionAdapterReadyObjectiveCloseoutDurableHandoff()
	view := BuildProductionAdapterObjectiveCloseoutHostView(ProductionAdapterObjectiveCloseoutHostViewInput{
		HostViewRef:             "view:metrics_objective_closeout_pending",
		ObjectiveCloseoutPacket: handoffObjectiveCloseoutPacket(handoff),
		DurableHandoff:          handoff,
	})
	fixture := BuildProductionAdapterObjectiveCloseoutBlackboxFixture(ProductionAdapterObjectiveCloseoutBlackboxFixtureInput{
		FixtureRef: "fixture:metrics_objective_closeout_pending",
		HostView:   view,
	})
	uiHandoff := BuildProductionAdapterObjectiveCloseoutHostUIHandoff(ProductionAdapterObjectiveCloseoutHostUIHandoffInput{
		HostUIHandoffRef: "ui:metrics_objective_closeout_pending",
		HostView:         view,
		DisplayFixture:   fixture,
	})
	return handoff, uiHandoff
}

func productionAdapterReadyObjectiveCloseoutWriterOptIn(descriptor ProductionAdapterObjectiveCloseoutWriterDescriptor, handoff ProductionAdapterObjectiveCloseoutDurableHandoff, uiHandoff ProductionAdapterObjectiveCloseoutHostUIHandoff) ProductionAdapterObjectiveCloseoutWriterOptIn {
	return BuildProductionAdapterObjectiveCloseoutWriterOptIn(ProductionAdapterObjectiveCloseoutWriterOptInInput{
		WriterOptInRef:                  "optin:metrics_objective_closeout_writer",
		WriterDescriptor:                descriptor,
		DurableHandoff:                  handoff,
		HostUIHandoff:                   uiHandoff,
		RequestedMode:                   ProductionAdapterObjectiveCloseoutWriterDurableWrite,
		ExplicitOptIn:                   true,
		HostWriterBindingRef:            "binding:metrics_objective_closeout_writer",
		HostWriterAvailable:             true,
		HostReadbackAvailable:           true,
		HostRollbackReviewAvailable:     true,
		HostCompensationReviewAvailable: true,
		AvailableCapabilityRefs:         descriptor.RequiresCapabilityRefs,
		PolicyRefs:                      descriptor.RequiredPolicyRefs,
		ApprovalRefs:                    descriptor.RequiredApprovalRefs,
		BudgetRef:                       descriptor.RequiredBudgetRef,
		IdempotencyRef:                  "idempotency:metrics_objective_closeout_request",
		DryRunPlanRef:                   "dryrun_plan:metrics_objective_closeout",
		DryRunResultRef:                 "dryrun_result:metrics_objective_closeout",
		ExpectedReadbackRef:             "readback:metrics_objective_closeout",
		RollbackReviewRef:               descriptor.RollbackReviewRef,
		CompensationReviewRef:           descriptor.CompensationReviewRef,
	})
}
