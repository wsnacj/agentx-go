package controlcontract

import (
	"encoding/json"
	"testing"
)

func TestProductionAdapterObjectiveCloseoutWriterBlackboxFixtureReadyMatrix(t *testing.T) {
	handoff, uiHandoff := productionAdapterReadyObjectiveCloseoutWriterInputs()
	descriptor := productionAdapterReadyObjectiveCloseoutWriterDescriptor()
	tests := []struct {
		name               string
		optIn              ProductionAdapterObjectiveCloseoutWriterOptIn
		wantStatus         string
		wantDisplayState   string
		wantPlan           bool
		wantDryRun         bool
		wantDurableWrite   bool
		wantHostMayWrite   bool
		requiredBoundary   string
		wantNextHostAction NextHostAction
	}{
		{
			name: "plan only",
			optIn: BuildProductionAdapterObjectiveCloseoutWriterOptIn(ProductionAdapterObjectiveCloseoutWriterOptInInput{
				WriterOptInRef:   "optin:metrics_objective_closeout_plan",
				WriterDescriptor: descriptor,
				DurableHandoff:   handoff,
				HostUIHandoff:    uiHandoff,
				RequestedMode:    ProductionAdapterObjectiveCloseoutWriterPlanOnly,
			}),
			wantStatus:         "ready_for_writer_plan_display",
			wantDisplayState:   "plan_only",
			wantPlan:           true,
			requiredBoundary:   "ready_for_writer_plan_display",
			wantNextHostAction: "review_objective_closeout_writer_plan",
		},
		{
			name: "dry run",
			optIn: BuildProductionAdapterObjectiveCloseoutWriterOptIn(productionAdapterObjectiveCloseoutWriterOptInInputForMode(
				descriptor,
				handoff,
				uiHandoff,
				"optin:metrics_objective_closeout_dry_run",
				ProductionAdapterObjectiveCloseoutWriterDryRun,
			)),
			wantStatus:         "ready_for_writer_dry_run_display",
			wantDisplayState:   "dry_run_ready",
			wantDryRun:         true,
			requiredBoundary:   "ready_for_writer_dry_run_display",
			wantNextHostAction: "host_may_run_objective_closeout_writer_dry_run",
		},
		{
			name:               "durable write",
			optIn:              productionAdapterReadyObjectiveCloseoutWriterOptIn(descriptor, handoff, uiHandoff),
			wantStatus:         "ready_for_durable_write_display",
			wantDisplayState:   "durable_write_ready",
			wantDurableWrite:   true,
			wantHostMayWrite:   true,
			requiredBoundary:   "ready_for_durable_write_display",
			wantNextHostAction: "host_may_execute_objective_closeout_durable_writer",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fixture := BuildProductionAdapterObjectiveCloseoutWriterBlackboxFixture(ProductionAdapterObjectiveCloseoutWriterBlackboxFixtureInput{
				FixtureRef:  "fixture:" + DisplaySafeRef(tc.optIn.WriterOptInRef),
				WriterOptIn: tc.optIn,
			})
			if fixture.Status != tc.wantStatus ||
				fixture.DisplayState != tc.wantDisplayState ||
				!fixture.ReadyForHostDisplay ||
				fixture.ReadyForWriterPlanDisplay != tc.wantPlan ||
				fixture.ReadyForWriterDryRunDisplay != tc.wantDryRun ||
				fixture.ReadyForDurableWriteDisplay != tc.wantDurableWrite ||
				fixture.HostMayExecuteDurableWrite != tc.wantHostMayWrite ||
				fixture.CoreInvocationExecuted ||
				fixture.DurableWriteByCore ||
				fixture.ObjectiveStoreWriteByCore ||
				fixture.RunstoreWriteByCore ||
				fixture.NextHostAction != tc.wantNextHostAction {
				t.Fatalf("unexpected writer fixture %s: %#v", tc.name, fixture)
			}
			AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
				Name:         "objective closeout writer fixture " + tc.name,
				RunnerEffect: fixture.RunnerEffect,
				PromptEffect: fixture.PromptEffect,
				Boundaries:   fixture.Boundaries,
				Payload:      fixture,
			}, "production_adapter_objective_closeout_writer_blackbox_fixture", "objective_closeout_writer_blackbox_fixture_projection_only", "host_cli_objective_closeout_writer_display", tc.requiredBoundary)
			AssertNoCoreMutation(t, "objective closeout writer fixture "+tc.name, fixture.CoreInvocationExecuted, fixture.DurableWriteByCore)
		})
	}
}

func TestProductionAdapterObjectiveCloseoutWriterBlackboxFixtureBlockedMatrix(t *testing.T) {
	handoff, uiHandoff := productionAdapterReadyObjectiveCloseoutWriterInputs()
	descriptor := productionAdapterReadyObjectiveCloseoutWriterDescriptor()
	tests := []struct {
		name             string
		input            ProductionAdapterObjectiveCloseoutWriterOptInInput
		wantDisplayState string
		wantReason       string
		wantMissing      MissingInput
	}{
		{
			name: "missing approval",
			input: func() ProductionAdapterObjectiveCloseoutWriterOptInInput {
				input := productionAdapterObjectiveCloseoutWriterOptInInputForMode(descriptor, handoff, uiHandoff, "optin:metrics_objective_closeout_missing_approval", ProductionAdapterObjectiveCloseoutWriterDurableWrite)
				input.ApprovalRefs = nil
				return input
			}(),
			wantDisplayState: "blocked_missing_approval",
			wantReason:       "writer_approval_missing",
			wantMissing:      "approval:objective_closeout_durable_write",
		},
		{
			name: "missing dry-run result",
			input: func() ProductionAdapterObjectiveCloseoutWriterOptInInput {
				input := productionAdapterObjectiveCloseoutWriterOptInInputForMode(descriptor, handoff, uiHandoff, "optin:metrics_objective_closeout_missing_dry_run", ProductionAdapterObjectiveCloseoutWriterDurableWrite)
				input.DryRunResultRef = ""
				return input
			}(),
			wantDisplayState: "blocked_missing_dry_run_result",
			wantReason:       "writer_dry_run_result_ref_missing",
			wantMissing:      "host:objective_closeout_writer_dry_run_result_ref",
		},
		{
			name: "missing expected readback",
			input: func() ProductionAdapterObjectiveCloseoutWriterOptInInput {
				input := productionAdapterObjectiveCloseoutWriterOptInInputForMode(descriptor, handoff, uiHandoff, "optin:metrics_objective_closeout_missing_readback", ProductionAdapterObjectiveCloseoutWriterDurableWrite)
				input.ExpectedReadbackRef = ""
				return input
			}(),
			wantDisplayState: "blocked_missing_expected_readback",
			wantReason:       "writer_expected_readback_ref_missing",
			wantMissing:      "host:objective_closeout_writer_expected_readback_ref",
		},
		{
			name: "missing rollback review",
			input: func() ProductionAdapterObjectiveCloseoutWriterOptInInput {
				input := productionAdapterObjectiveCloseoutWriterOptInInputForMode(descriptor, handoff, uiHandoff, "optin:metrics_objective_closeout_missing_rollback", ProductionAdapterObjectiveCloseoutWriterDurableWrite)
				input.RollbackReviewRef = ""
				return input
			}(),
			wantDisplayState: "blocked_missing_rollback_review",
			wantReason:       "writer_rollback_review_ref_missing",
			wantMissing:      "host:objective_closeout_writer_rollback_review_ref",
		},
		{
			name: "missing compensation review",
			input: func() ProductionAdapterObjectiveCloseoutWriterOptInInput {
				input := productionAdapterObjectiveCloseoutWriterOptInInputForMode(descriptor, handoff, uiHandoff, "optin:metrics_objective_closeout_missing_compensation", ProductionAdapterObjectiveCloseoutWriterDurableWrite)
				input.CompensationReviewRef = ""
				return input
			}(),
			wantDisplayState: "blocked_missing_compensation_review",
			wantReason:       "writer_compensation_review_ref_missing",
			wantMissing:      "host:objective_closeout_writer_compensation_review_ref",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			optIn := BuildProductionAdapterObjectiveCloseoutWriterOptIn(tc.input)
			fixture := BuildProductionAdapterObjectiveCloseoutWriterBlackboxFixture(ProductionAdapterObjectiveCloseoutWriterBlackboxFixtureInput{
				FixtureRef:  "fixture:" + DisplaySafeRef(tc.input.WriterOptInRef),
				WriterOptIn: optIn,
			})
			if fixture.Status != "writer_opt_in_blocked_display" ||
				fixture.DisplayState != tc.wantDisplayState ||
				!fixture.ReadyForHostDisplay ||
				!fixture.BlockedDisplay ||
				fixture.ReadyForWriterPlanDisplay ||
				fixture.ReadyForWriterDryRunDisplay ||
				fixture.ReadyForDurableWriteDisplay ||
				fixture.HostMayExecuteDurableWrite ||
				!productionAdapterStringContains(fixture.BlockedReasons, tc.wantReason) ||
				!productionAdapterMissingContains(fixture.MissingInputs, tc.wantMissing) ||
				fixture.CoreInvocationExecuted ||
				fixture.DurableWriteByCore ||
				fixture.ObjectiveStoreWriteByCore ||
				fixture.RunstoreWriteByCore {
				t.Fatalf("unexpected blocked writer fixture %s: %#v", tc.name, fixture)
			}
			AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
				Name:         "objective closeout writer blocked fixture " + tc.name,
				RunnerEffect: fixture.RunnerEffect,
				PromptEffect: fixture.PromptEffect,
				Boundaries:   fixture.Boundaries,
				Payload:      fixture,
			}, "production_adapter_objective_closeout_writer_blackbox_fixture", "host_cli_objective_closeout_writer_display", "writer_opt_in_blocked_display")
			AssertNoRawPayload(t, "objective closeout writer blocked fixture "+tc.name, fixture, "/Users/mason", "raw local host task")
		})
	}
}

func TestProductionAdapterObjectiveCloseoutWriterBlackboxFixtureUnsafeRefsDisplaySafeBlock(t *testing.T) {
	handoff, uiHandoff := productionAdapterReadyObjectiveCloseoutWriterInputs()
	descriptor := productionAdapterReadyObjectiveCloseoutWriterDescriptor()
	input := productionAdapterObjectiveCloseoutWriterOptInInputForMode(
		descriptor,
		handoff,
		uiHandoff,
		"optin:metrics_objective_closeout_unsafe",
		ProductionAdapterObjectiveCloseoutWriterDurableWrite,
	)
	input.ApprovalRefs = []DisplaySafeRef{"postgresql://secret@example.invalid/db"}

	optIn := BuildProductionAdapterObjectiveCloseoutWriterOptIn(input)
	fixture := BuildProductionAdapterObjectiveCloseoutWriterBlackboxFixture(ProductionAdapterObjectiveCloseoutWriterBlackboxFixtureInput{
		FixtureRef:  "fixture:metrics_objective_closeout_unsafe",
		WriterOptIn: optIn,
	})
	if fixture.Status != "writer_opt_in_blocked_display" ||
		fixture.DisplayState != "blocked_unsafe_refs" ||
		!fixture.ReadyForHostDisplay ||
		!fixture.BlockedDisplay ||
		fixture.ReadyForWriterPlanDisplay ||
		fixture.ReadyForWriterDryRunDisplay ||
		fixture.ReadyForDurableWriteDisplay ||
		fixture.HostMayExecuteDurableWrite ||
		fixture.FailureClass != FailureEvidenceWeak ||
		!productionAdapterStringContains(fixture.BlockedReasons, "unsafe_input_ref") ||
		!productionAdapterMissingContains(fixture.MissingInputs, "host:display_safe_refs") ||
		fixture.CoreInvocationExecuted ||
		fixture.DurableWriteByCore ||
		fixture.ObjectiveStoreWriteByCore ||
		fixture.RunstoreWriteByCore {
		t.Fatalf("expected unsafe writer fixture to expose a display-safe block, got %#v", fixture)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "objective closeout writer unsafe fixture",
		RunnerEffect: fixture.RunnerEffect,
		PromptEffect: fixture.PromptEffect,
		Boundaries:   fixture.Boundaries,
		Payload:      fixture,
	}, "production_adapter_objective_closeout_writer_blackbox_fixture", "objective_closeout_writer_blackbox_fixture_projection_only", "host_cli_objective_closeout_writer_display", "writer_opt_in_blocked_display")
	AssertNoRawPayload(t, "objective closeout writer unsafe fixture", fixture, "postgresql://secret", "example.invalid", "/Users/mason", "raw local host task")
}

func TestProductionAdapterObjectiveCloseoutWriterBlackboxFixtureRejectsUnsafeFixtureInput(t *testing.T) {
	handoff, uiHandoff := productionAdapterReadyObjectiveCloseoutWriterInputs()
	descriptor := productionAdapterReadyObjectiveCloseoutWriterDescriptor()
	optIn := productionAdapterReadyObjectiveCloseoutWriterOptIn(descriptor, handoff, uiHandoff)

	fixture := BuildProductionAdapterObjectiveCloseoutWriterBlackboxFixture(ProductionAdapterObjectiveCloseoutWriterBlackboxFixtureInput{
		FixtureRef:  "postgresql://secret@example.invalid/db",
		WriterOptIn: optIn,
	})
	if fixture.Status != "blocked" ||
		fixture.ReadyForHostDisplay ||
		fixture.BlockedDisplay ||
		fixture.ReadyForDurableWriteDisplay ||
		fixture.HostMayExecuteDurableWrite ||
		fixture.FailureClass != FailureEvidenceWeak ||
		!productionAdapterStringContains(fixture.BlockedReasons, "unsafe_input_ref") ||
		!productionAdapterMissingContains(fixture.MissingInputs, "host:display_safe_refs") {
		t.Fatalf("expected unsafe fixture input to block display, got %#v", fixture)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "objective closeout writer unsafe fixture input",
		RunnerEffect: fixture.RunnerEffect,
		PromptEffect: fixture.PromptEffect,
		Boundaries:   fixture.Boundaries,
		Payload:      fixture,
	}, "production_adapter_objective_closeout_writer_blackbox_fixture", "objective_closeout_writer_blackbox_fixture_projection_only", "host_cli_objective_closeout_writer_display", "objective_closeout_writer_blackbox_fixture_blocked")
	AssertNoRawPayload(t, "objective closeout writer unsafe fixture input", fixture, "postgresql://secret", "example.invalid")
}

func TestProductionAdapterObjectiveCloseoutWriterBlackboxFixtureRequiresDisplayRefs(t *testing.T) {
	handoff, uiHandoff := productionAdapterReadyObjectiveCloseoutWriterInputs()
	descriptor := productionAdapterReadyObjectiveCloseoutWriterDescriptor()
	descriptor.WriterRef = ""
	descriptor = BuildProductionAdapterObjectiveCloseoutWriterDescriptor(descriptor)
	optIn := BuildProductionAdapterObjectiveCloseoutWriterOptIn(ProductionAdapterObjectiveCloseoutWriterOptInInput{
		WriterOptInRef:   "optin:metrics_objective_closeout_missing_display_refs",
		WriterDescriptor: descriptor,
		DurableHandoff:   handoff,
		HostUIHandoff:    uiHandoff,
		RequestedMode:    ProductionAdapterObjectiveCloseoutWriterPlanOnly,
	})

	fixture := BuildProductionAdapterObjectiveCloseoutWriterBlackboxFixture(ProductionAdapterObjectiveCloseoutWriterBlackboxFixtureInput{
		FixtureRef:  "fixture:metrics_objective_closeout_missing_display_refs",
		WriterOptIn: optIn,
	})
	if fixture.Status != "blocked" ||
		fixture.ReadyForHostDisplay ||
		fixture.BlockedDisplay ||
		!productionAdapterStringContains(fixture.BlockedReasons, "objective_closeout_writer_display_refs_missing") ||
		!productionAdapterMissingContains(fixture.MissingInputs, "host:objective_closeout_writer_display_refs") {
		t.Fatalf("expected missing display refs to block fixture display, got %#v", fixture)
	}
	AssertNoCoreMutation(t, "objective closeout writer missing display refs fixture", fixture.CoreInvocationExecuted, fixture.DurableWriteByCore)
}

func TestProductionAdapterObjectiveCloseoutWriterBlackboxFixtureJSONCompatibility(t *testing.T) {
	handoff, uiHandoff := productionAdapterReadyObjectiveCloseoutWriterInputs()
	descriptor := productionAdapterReadyObjectiveCloseoutWriterDescriptor()
	fixture := BuildProductionAdapterObjectiveCloseoutWriterBlackboxFixture(ProductionAdapterObjectiveCloseoutWriterBlackboxFixtureInput{
		FixtureRef:  "fixture:metrics_objective_closeout_writer",
		WriterOptIn: productionAdapterReadyObjectiveCloseoutWriterOptIn(descriptor, handoff, uiHandoff),
	})
	raw, err := json.Marshal(fixture)
	if err != nil {
		t.Fatalf("marshal objective closeout writer fixture: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("unmarshal objective closeout writer fixture: %v", err)
	}
	for _, key := range []string{
		"fixture_ref",
		"display_state",
		"display_sections",
		"writer_opt_in_ref",
		"writer_ref",
		"requested_mode",
		"ready_for_host_display",
		"ready_for_durable_write_display",
		"host_may_execute_durable_write",
		"dry_run_result_ref",
		"expected_readback_ref",
	} {
		if _, ok := payload[key]; !ok {
			t.Fatalf("expected stable JSON key %q in %s", key, string(raw))
		}
	}
	AssertNoRawPayload(t, "objective closeout writer fixture JSON", fixture, "/Users/mason", "postgresql://secret", "raw local host task")
}

func productionAdapterObjectiveCloseoutWriterOptInInputForMode(
	descriptor ProductionAdapterObjectiveCloseoutWriterDescriptor,
	handoff ProductionAdapterObjectiveCloseoutDurableHandoff,
	uiHandoff ProductionAdapterObjectiveCloseoutHostUIHandoff,
	optInRef DisplaySafeRef,
	mode ProductionAdapterObjectiveCloseoutWriterMode,
) ProductionAdapterObjectiveCloseoutWriterOptInInput {
	return ProductionAdapterObjectiveCloseoutWriterOptInInput{
		WriterOptInRef:                  optInRef,
		WriterDescriptor:                descriptor,
		DurableHandoff:                  handoff,
		HostUIHandoff:                   uiHandoff,
		RequestedMode:                   mode,
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
	}
}
