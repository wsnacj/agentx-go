package controlcontract

import (
	"testing"
)

func TestProductionAdapterObjectiveCloseoutDurableHandoffRequiresApplyConfirmation(t *testing.T) {
	packet := productionAdapterConfirmedObjectiveCloseoutPacket()

	handoff := BuildProductionAdapterObjectiveCloseoutDurableHandoff(ProductionAdapterObjectiveCloseoutDurableHandoffInput{
		ObjectiveCloseoutHandoffRef: "handoff:metrics_objective_closeout",
		HostObjectiveLifecycleRef:   "lifecycle:metrics_objective",
		HostRunstoreRef:             "runstore:metrics_objective",
		ExpectedDurableEventRef:     "event:metrics_objective_closed",
		ExpectedObjectiveStateRef:   "state:metrics_objective_closed",
		ObjectiveCloseoutPacket:     packet,
	})
	if handoff.Status != HostActionReviewRequired ||
		handoff.ReadyForHostDurableApply ||
		!handoff.ObjectiveSatisfied ||
		!handoff.VerificationSatisfied ||
		!handoff.HostCloseoutConfirmed ||
		handoff.HostDurableApplyConfirmed ||
		handoff.FailureClass != FailureApprovalRequired ||
		handoff.NextHostAction != "confirm_objective_closeout_durable_apply" ||
		!productionAdapterMissingContains(handoff.MissingInputs, "host:objective_closeout_durable_apply_confirmation_ref") ||
		!productionAdapterStringContains(handoff.BlockedReasons, "objective_closeout_durable_apply_confirmation_required") ||
		!productionAdapterBoundaryContains(handoff.Boundaries, "host_objective_closeout_durable_apply_confirmation_required") ||
		handoff.CoreInvocationExecuted ||
		handoff.DurableWriteByCore ||
		handoff.ObjectiveStoreWriteByCore ||
		handoff.RunstoreWriteByCore {
		t.Fatalf("durable handoff requiring host apply confirmation = %#v", handoff)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "adapter objective closeout durable handoff",
		RunnerEffect: handoff.RunnerEffect,
		PromptEffect: handoff.PromptEffect,
		Boundaries:   handoff.Boundaries,
		Payload:      handoff,
	}, "production_adapter_objective_closeout_durable_handoff", "objective_closeout_durable_handoff_projection_only", "host_owned_objective_closeout_durable_handoff", "host_owned_durable_closeout", "host_objective_closeout_durable_apply_confirmation_required")
	AssertNoCoreMutation(t, "adapter objective closeout durable handoff", handoff.CoreInvocationExecuted, handoff.DurableWriteByCore)
}

func TestProductionAdapterObjectiveCloseoutDurableHandoffReady(t *testing.T) {
	handoff := productionAdapterReadyObjectiveCloseoutDurableHandoff()

	if handoff.Status != HostActionReady ||
		!handoff.ReadyForHostDurableApply ||
		!handoff.ObjectiveSatisfied ||
		!handoff.VerificationSatisfied ||
		!handoff.HostCloseoutConfirmed ||
		!handoff.HostDurableApplyConfirmed ||
		handoff.CurrentLifecycleStage != LifecycleStageAudit ||
		handoff.TargetLifecycleStage != LifecycleStageClosed ||
		!handoff.LifecycleTransition.Allowed ||
		handoff.FailureClass != FailureNone ||
		handoff.NextHostAction != "host_may_apply_objective_closeout" ||
		handoff.CoreInvocationExecuted ||
		handoff.DurableWriteByCore ||
		handoff.ObjectiveStoreWriteByCore ||
		handoff.RunstoreWriteByCore ||
		!productionAdapterEvidenceContains(handoff.EvidenceRefs, "confirmation:metrics_objective_durable_apply", "host_objective_closeout_durable_apply_confirmation") ||
		!productionAdapterBoundaryContains(handoff.Boundaries, "host_objective_closeout_durable_apply_confirmed") ||
		!productionAdapterBoundaryContains(handoff.Boundaries, "ready_for_host_objective_closeout_durable_apply") {
		t.Fatalf("ready objective closeout durable handoff = %#v", handoff)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "adapter ready objective closeout durable handoff",
		RunnerEffect: handoff.RunnerEffect,
		PromptEffect: handoff.PromptEffect,
		Boundaries:   handoff.Boundaries,
		Payload:      handoff,
	}, "production_adapter_objective_closeout_durable_handoff", "host_owned_durable_closeout", "no_durable_write_by_core", "no_objective_store_write_by_core", "no_runstore_write_by_core", "ready_for_host_objective_closeout_durable_apply")
	AssertNoCoreMutation(t, "adapter ready objective closeout durable handoff", handoff.CoreInvocationExecuted, handoff.DurableWriteByCore)
}

func TestProductionAdapterObjectiveCloseoutReadbackRecordsHostDurableCloseout(t *testing.T) {
	handoff := productionAdapterReadyObjectiveCloseoutDurableHandoff()

	readback := BuildProductionAdapterObjectiveCloseoutReadback(ProductionAdapterObjectiveCloseoutReadbackInput{
		ObjectiveCloseoutReadbackRef: "readback:metrics_objective_closeout",
		DurableHandoff:               handoff,
		AppliedDurableEventRef:       handoff.ExpectedDurableEventRef,
		AppliedRunstoreRef:           handoff.HostRunstoreRef,
		AppliedObjectiveStateRef:     handoff.ExpectedObjectiveStateRef,
		HostDurableWriteReported:     true,
		HostDurableWriteSucceeded:    true,
	})
	if readback.Status != HostActionRecorded ||
		!readback.ReadyForObjectiveCloseoutReadback ||
		!readback.ObjectiveLifecycleClosed ||
		!readback.ObjectiveSatisfied ||
		!readback.VerificationSatisfied ||
		!readback.HostDurableApplyConfirmed ||
		!readback.HostDurableWriteReported ||
		!readback.HostDurableWriteSucceeded ||
		readback.HostDurableWriteFailed ||
		readback.FailureClass != FailureNone ||
		readback.NextHostAction != "return_objective_closed_lifecycle" ||
		readback.CoreInvocationExecuted ||
		readback.DurableWriteByCore ||
		readback.ObjectiveStoreWriteByCore ||
		readback.RunstoreWriteByCore ||
		!productionAdapterEvidenceContains(readback.EvidenceRefs, "event:metrics_objective_closed", "objective_closeout_durable_event") ||
		!productionAdapterBoundaryContains(readback.Boundaries, "host_objective_closeout_durable_write_recorded") ||
		!productionAdapterBoundaryContains(readback.Boundaries, "objective_lifecycle_closeout_readback_recorded") ||
		!productionAdapterBoundaryContains(readback.Boundaries, "objective_lifecycle_closed_by_host") {
		t.Fatalf("objective closeout durable readback = %#v", readback)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "adapter objective closeout durable readback",
		RunnerEffect: readback.RunnerEffect,
		PromptEffect: readback.PromptEffect,
		Boundaries:   readback.Boundaries,
		Payload:      readback,
	}, "production_adapter_objective_closeout_readback", "objective_closeout_readback_projection_only", "host_owned_objective_closeout_readback", "host_owned_durable_closeout_readback", "objective_lifecycle_closed_by_host")
	AssertNoCoreMutation(t, "adapter objective closeout durable readback", readback.CoreInvocationExecuted, readback.DurableWriteByCore)
}

func TestProductionAdapterObjectiveCloseoutReadbackRejectsMismatchedRefs(t *testing.T) {
	handoff := productionAdapterReadyObjectiveCloseoutDurableHandoff()

	readback := BuildProductionAdapterObjectiveCloseoutReadback(ProductionAdapterObjectiveCloseoutReadbackInput{
		ObjectiveCloseoutReadbackRef: "readback:metrics_objective_closeout",
		DurableHandoff:               handoff,
		AppliedDurableEventRef:       handoff.ExpectedDurableEventRef,
		AppliedRunstoreRef:           handoff.HostRunstoreRef,
		AppliedObjectiveStateRef:     "state:other_objective",
		HostDurableWriteReported:     true,
		HostDurableWriteSucceeded:    true,
	})
	if readback.Status != HostActionBlocked ||
		readback.ReadyForObjectiveCloseoutReadback ||
		readback.ObjectiveLifecycleClosed ||
		readback.ObjectiveSatisfied ||
		readback.FailureClass != FailureVerificationFailed ||
		!productionAdapterStringContains(readback.BlockedReasons, "objective_closeout_state_ref_mismatch") ||
		!productionAdapterMissingContains(readback.MissingInputs, "host:objective_state_ref") {
		t.Fatalf("expected mismatched objective state ref to block readback, got %#v", readback)
	}
}

func TestProductionAdapterObjectiveCloseoutDurableRejectsUnsafeRefs(t *testing.T) {
	packet := productionAdapterConfirmedObjectiveCloseoutPacket()

	handoff := BuildProductionAdapterObjectiveCloseoutDurableHandoff(ProductionAdapterObjectiveCloseoutDurableHandoffInput{
		ObjectiveCloseoutHandoffRef:     "handoff:metrics_objective_closeout",
		HostObjectiveLifecycleRef:       "lifecycle:metrics_objective",
		HostRunstoreRef:                 "/Users/mason/private/runstore.json",
		ExpectedDurableEventRef:         "event:metrics_objective_closed",
		ExpectedObjectiveStateRef:       "state:metrics_objective_closed",
		HostDurableApplyConfirmationRef: "confirmation:metrics_objective_durable_apply",
		ObjectiveCloseoutPacket:         packet,
	})
	if handoff.Status != HostActionBlocked ||
		handoff.ReadyForHostDurableApply ||
		handoff.FailureClass != FailureEvidenceWeak ||
		!productionAdapterStringContains(handoff.BlockedReasons, "unsafe_input_ref") ||
		!productionAdapterMissingContains(handoff.MissingInputs, "host:display_safe_refs") {
		t.Fatalf("unsafe durable handoff = %#v", handoff)
	}
	AssertNoRawPayload(t, "unsafe objective closeout durable handoff", handoff, "/Users/mason/private/runstore.json")
}

func productionAdapterConfirmedObjectiveCloseoutPacket() ProductionAdapterObjectiveCloseoutPacket {
	return BuildProductionAdapterObjectiveCloseoutPacket(ProductionAdapterObjectiveCloseoutPacketInput{
		ObjectiveCloseoutPacketRef:   "closeout:metrics_objective",
		ObjectiveRef:                 "objective:metrics_monitoring",
		HostCloseoutConfirmationRef:  "confirmation:metrics_objective_closeout",
		CompletionAuditResultBinding: productionAdapterReadyCompletionAuditResultBinding(),
	})
}

func productionAdapterReadyObjectiveCloseoutDurableHandoff() ProductionAdapterObjectiveCloseoutDurableHandoff {
	return BuildProductionAdapterObjectiveCloseoutDurableHandoff(ProductionAdapterObjectiveCloseoutDurableHandoffInput{
		ObjectiveCloseoutHandoffRef:     "handoff:metrics_objective_closeout",
		HostObjectiveLifecycleRef:       "lifecycle:metrics_objective",
		HostRunstoreRef:                 "runstore:metrics_objective",
		ExpectedDurableEventRef:         "event:metrics_objective_closed",
		ExpectedObjectiveStateRef:       "state:metrics_objective_closed",
		HostDurableApplyConfirmationRef: "confirmation:metrics_objective_durable_apply",
		ObjectiveCloseoutPacket:         productionAdapterConfirmedObjectiveCloseoutPacket(),
	})
}
