package controlcontract

import (
	"testing"
)

func TestProductionAdapterObjectiveCloseoutHostViewDisplaysDurableReadback(t *testing.T) {
	handoff := productionAdapterReadyObjectiveCloseoutDurableHandoff()
	readback := productionAdapterRecordedObjectiveCloseoutReadback(handoff)

	view := BuildProductionAdapterObjectiveCloseoutHostView(ProductionAdapterObjectiveCloseoutHostViewInput{
		HostViewRef:             "view:metrics_objective_closeout",
		ObjectiveCloseoutPacket: handoffObjectiveCloseoutPacket(handoff),
		DurableHandoff:          handoff,
		Readback:                readback,
	})
	if view.Status != "objective_lifecycle_closed" ||
		view.DisplayState != "objective_return_final" ||
		!view.ReadyForHostDisplay ||
		view.ReadyForHostDurableApply ||
		!view.ReadyForObjectiveReturn ||
		view.IntermediateDisplay ||
		!view.FinalDisplay ||
		view.FailureReviewDisplay ||
		!view.ObjectiveLifecycleClosed ||
		!view.ObjectiveSatisfied ||
		!view.VerificationSatisfied ||
		!view.HostCloseoutConfirmed ||
		!view.HostDurableApplyConfirmed ||
		!view.HostDurableWriteReported ||
		!view.HostDurableWriteSucceeded ||
		view.HostDurableWriteFailed ||
		view.ObjectiveCloseoutPacketRef != handoff.ObjectiveCloseoutPacketRef ||
		view.ObjectiveCloseoutHandoffRef != handoff.ObjectiveCloseoutHandoffRef ||
		view.ObjectiveCloseoutReadbackRef != readback.ObjectiveCloseoutReadbackRef ||
		view.AuthoritativeObjectiveCloseoutPacketRef != handoff.ObjectiveCloseoutPacketRef ||
		view.AuthoritativeObjectiveCloseoutHandoffRef != handoff.ObjectiveCloseoutHandoffRef ||
		view.AuthoritativeObjectiveRef != handoff.ObjectiveRef ||
		view.AuthoritativeHostRunstoreRef != handoff.HostRunstoreRef ||
		view.AuthoritativeDurableEventRef != handoff.ExpectedDurableEventRef ||
		view.AuthoritativeObjectiveStateRef != handoff.ExpectedObjectiveStateRef ||
		view.ObservedObjectiveCloseoutPacketRef != readback.ObjectiveCloseoutPacketRef ||
		view.ObservedObjectiveCloseoutHandoffRef != readback.ObjectiveCloseoutHandoffRef ||
		view.ObservedObjectiveRef != readback.ObjectiveRef ||
		view.ObservedHostRunstoreRef != readback.HostRunstoreRef ||
		view.ObservedAppliedDurableEventRef != readback.AppliedDurableEventRef ||
		view.ObservedAppliedRunstoreRef != readback.AppliedRunstoreRef ||
		view.ObservedAppliedObjectiveStateRef != readback.AppliedObjectiveStateRef ||
		view.HostCloseoutConfirmationRef != handoff.HostCloseoutConfirmationRef ||
		view.HostDurableApplyConfirmationRef != handoff.HostDurableApplyConfirmationRef ||
		view.CoreInvocationExecuted ||
		view.DurableWriteByCore ||
		view.ObjectiveStoreWriteByCore ||
		view.RunstoreWriteByCore ||
		view.NextHostAction != "return_objective_closed_lifecycle" {
		t.Fatalf("unexpected objective closeout host view: %#v", view)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "adapter objective closeout host view",
		RunnerEffect: view.RunnerEffect,
		PromptEffect: view.PromptEffect,
		Boundaries:   view.Boundaries,
		Payload:      view,
	}, "production_adapter_objective_closeout_host_view", "objective_closeout_host_view_projection_only", "host_cli_objective_closeout_display", "ready_for_objective_return")
	AssertNoCoreMutation(t, "adapter objective closeout host view", view.CoreInvocationExecuted, view.DurableWriteByCore)
}

func TestProductionAdapterObjectiveCloseoutHostViewReadyBeforeReadback(t *testing.T) {
	handoff := productionAdapterReadyObjectiveCloseoutDurableHandoff()

	view := BuildProductionAdapterObjectiveCloseoutHostView(ProductionAdapterObjectiveCloseoutHostViewInput{
		HostViewRef:             "view:metrics_objective_closeout",
		ObjectiveCloseoutPacket: handoffObjectiveCloseoutPacket(handoff),
		DurableHandoff:          handoff,
	})
	if view.Status != "ready_for_host_durable_apply" ||
		view.DisplayState != "durable_apply_pending" ||
		!view.ReadyForHostDisplay ||
		!view.ReadyForHostDurableApply ||
		view.ReadyForObjectiveReturn ||
		!view.IntermediateDisplay ||
		view.FinalDisplay ||
		view.FailureReviewDisplay ||
		view.ObjectiveLifecycleClosed ||
		!view.ObjectiveSatisfied ||
		view.AuthoritativeObjectiveCloseoutPacketRef != handoff.ObjectiveCloseoutPacketRef ||
		view.AuthoritativeObjectiveCloseoutHandoffRef != handoff.ObjectiveCloseoutHandoffRef ||
		view.AuthoritativeObjectiveRef != handoff.ObjectiveRef ||
		view.ObservedObjectiveCloseoutPacketRef != "" ||
		view.ObservedAppliedDurableEventRef != "" ||
		!productionAdapterMissingContains(view.MissingInputs, "host:objective_closeout_readback") ||
		!productionAdapterStringContains(view.BlockedReasons, "objective_closeout_readback_not_ready") ||
		!productionAdapterMissingContains(view.DisplayProgressMissingInputs, "host:objective_closeout_readback") ||
		!productionAdapterStringContains(view.DisplayProgressBlockedReasons, "objective_closeout_readback_not_ready") ||
		view.NextHostAction != "host_may_apply_objective_closeout" {
		t.Fatalf("expected host durable apply view before readback, got %#v", view)
	}
	AssertNoCoreMutation(t, "adapter objective closeout host view before readback", view.CoreInvocationExecuted, view.DurableWriteByCore)
}

func TestProductionAdapterObjectiveCloseoutBlackboxFixtureReadyBeforeReadback(t *testing.T) {
	handoff := productionAdapterReadyObjectiveCloseoutDurableHandoff()
	view := BuildProductionAdapterObjectiveCloseoutHostView(ProductionAdapterObjectiveCloseoutHostViewInput{
		HostViewRef:             "view:metrics_objective_closeout",
		ObjectiveCloseoutPacket: handoffObjectiveCloseoutPacket(handoff),
		DurableHandoff:          handoff,
	})

	fixture := BuildProductionAdapterObjectiveCloseoutBlackboxFixture(ProductionAdapterObjectiveCloseoutBlackboxFixtureInput{
		FixtureRef: "fixture:metrics_objective_closeout_pending",
		HostView:   view,
	})
	if fixture.Status != "ready_for_host_durable_apply_display" ||
		fixture.DisplayState != "durable_apply_pending" ||
		!fixture.ReadyForHostDisplay ||
		fixture.ReadyForObjectiveReturn ||
		!fixture.IntermediateDisplay ||
		fixture.FinalDisplay ||
		fixture.ObjectiveLifecycleClosed ||
		!fixture.ObjectiveSatisfied ||
		fixture.AuthoritativeObjectiveCloseoutPacketRef != handoff.ObjectiveCloseoutPacketRef ||
		fixture.AuthoritativeObjectiveCloseoutHandoffRef != handoff.ObjectiveCloseoutHandoffRef ||
		fixture.ObservedObjectiveCloseoutPacketRef != "" ||
		fixture.ObservedAppliedDurableEventRef != "" ||
		!productionAdapterMissingContains(fixture.DisplayProgressMissingInputs, "host:objective_closeout_readback") ||
		!productionAdapterStringContains(fixture.DisplayProgressBlockedReasons, "objective_closeout_readback_not_ready") ||
		len(fixture.MissingInputs) != 0 ||
		len(fixture.BlockedReasons) != 0 ||
		fixture.NextHostAction != "host_may_apply_objective_closeout" {
		t.Fatalf("unexpected intermediate objective closeout fixture: %#v", fixture)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "adapter objective closeout intermediate fixture",
		RunnerEffect: fixture.RunnerEffect,
		PromptEffect: fixture.PromptEffect,
		Boundaries:   fixture.Boundaries,
		Payload:      fixture,
	}, "production_adapter_objective_closeout_blackbox_fixture", "objective_closeout_blackbox_fixture_projection_only", "host_cli_objective_closeout_display", "ready_for_host_durable_apply_display")
	AssertNoCoreMutation(t, "adapter objective closeout intermediate fixture", fixture.CoreInvocationExecuted, fixture.DurableWriteByCore)
}

func TestProductionAdapterObjectiveCloseoutFailureReviewReady(t *testing.T) {
	handoff := productionAdapterReadyObjectiveCloseoutDurableHandoff()
	readback := productionAdapterFailedObjectiveCloseoutReadback(handoff)

	view := BuildProductionAdapterObjectiveCloseoutHostView(ProductionAdapterObjectiveCloseoutHostViewInput{
		HostViewRef:             "view:metrics_objective_closeout_failure",
		ObjectiveCloseoutPacket: handoffObjectiveCloseoutPacket(handoff),
		DurableHandoff:          handoff,
		Readback:                readback,
	})
	if view.Status != "objective_closeout_durable_failure_review" ||
		view.DisplayState != "durable_failure_review" ||
		!view.ReadyForHostDisplay ||
		view.ReadyForHostDurableApply ||
		!view.ReadyForFailureReview ||
		!view.ReadyForCompensationReview ||
		view.ReadyForObjectiveReturn ||
		!view.IntermediateDisplay ||
		view.FinalDisplay ||
		!view.FailureReviewDisplay ||
		view.ObjectiveLifecycleClosed ||
		view.ObjectiveSatisfied ||
		!view.HostDurableWriteReported ||
		view.HostDurableWriteSucceeded ||
		!view.HostDurableWriteFailed ||
		view.FailureRef != "failure:metrics_objective_closeout" ||
		view.CompensationRef != "compensation:metrics_objective_closeout_review" ||
		view.FailureClass != FailureVerificationFailed ||
		view.NextHostAction != "review_objective_closeout_durable_failure" {
		t.Fatalf("unexpected objective closeout failure host view: %#v", view)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "adapter objective closeout failure host view",
		RunnerEffect: view.RunnerEffect,
		PromptEffect: view.PromptEffect,
		Boundaries:   view.Boundaries,
		Payload:      view,
	}, "production_adapter_objective_closeout_host_view", "objective_closeout_host_view_projection_only", "host_cli_objective_closeout_display", "objective_closeout_durable_failure_review", "compensation_not_executed")
	AssertNoCoreMutation(t, "adapter objective closeout failure host view", view.CoreInvocationExecuted, view.DurableWriteByCore)

	packet := BuildProductionAdapterObjectiveCloseoutFailureReviewPacket(ProductionAdapterObjectiveCloseoutFailureReviewPacketInput{
		FailureReviewPacketRef: "review:metrics_objective_closeout_failure",
		HostView:               view,
	})
	if packet.Status != "ready_for_objective_closeout_failure_review" ||
		!packet.ReadyForHostDisplay ||
		!packet.ReadyForFailureReview ||
		!packet.ReadyForCompensationReview ||
		packet.ObjectiveLifecycleClosed ||
		packet.ObjectiveSatisfied ||
		packet.FailureRef != view.FailureRef ||
		packet.CompensationRef != view.CompensationRef ||
		packet.CoreInvocationExecuted ||
		packet.DurableWriteByCore ||
		packet.ObjectiveStoreWriteByCore ||
		packet.RunstoreWriteByCore ||
		packet.NextHostAction != "review_objective_closeout_durable_failure" {
		t.Fatalf("unexpected objective closeout failure review packet: %#v", packet)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "adapter objective closeout failure review packet",
		RunnerEffect: packet.RunnerEffect,
		PromptEffect: packet.PromptEffect,
		Boundaries:   packet.Boundaries,
		Payload:      packet,
	}, "production_adapter_objective_closeout_failure_review_packet", "objective_closeout_failure_review_projection_only", "host_owned_objective_closeout_failure_review", "failure_compensation_review_only", "compensation_not_executed")
	AssertNoCoreMutation(t, "adapter objective closeout failure review packet", packet.CoreInvocationExecuted, packet.DurableWriteByCore)

	fixture := BuildProductionAdapterObjectiveCloseoutFailureReviewBlackboxFixture(ProductionAdapterObjectiveCloseoutFailureReviewBlackboxFixtureInput{
		FixtureRef:          "fixture:metrics_objective_closeout_failure",
		FailureReviewPacket: packet,
	})
	if fixture.Status != "ready_for_objective_closeout_failure_display" ||
		!fixture.ReadyForHostDisplay ||
		!fixture.ReadyForFailureReview ||
		!fixture.ReadyForCompensationReview ||
		fixture.ObjectiveLifecycleClosed ||
		fixture.ObjectiveSatisfied ||
		fixture.FailureReviewPacketRef != packet.FailureReviewPacketRef ||
		fixture.FailureRef != packet.FailureRef ||
		fixture.CompensationRef != packet.CompensationRef ||
		fixture.CoreInvocationExecuted ||
		fixture.DurableWriteByCore ||
		fixture.ObjectiveStoreWriteByCore ||
		fixture.RunstoreWriteByCore ||
		fixture.NextHostAction != "review_objective_closeout_durable_failure" {
		t.Fatalf("unexpected objective closeout failure review fixture: %#v", fixture)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "adapter objective closeout failure review fixture",
		RunnerEffect: fixture.RunnerEffect,
		PromptEffect: fixture.PromptEffect,
		Boundaries:   fixture.Boundaries,
		Payload:      fixture,
	}, "production_adapter_objective_closeout_failure_review_blackbox_fixture", "objective_closeout_failure_review_fixture_projection_only", "host_cli_objective_closeout_failure_display", "compensation_not_executed")
	AssertNoCoreMutation(t, "adapter objective closeout failure review fixture", fixture.CoreInvocationExecuted, fixture.DurableWriteByCore)
}

func TestProductionAdapterObjectiveCloseoutBlackboxFixtureReady(t *testing.T) {
	view := productionAdapterReadyObjectiveCloseoutHostView()

	fixture := BuildProductionAdapterObjectiveCloseoutBlackboxFixture(ProductionAdapterObjectiveCloseoutBlackboxFixtureInput{
		FixtureRef: "fixture:metrics_objective_closeout",
		HostView:   view,
	})
	if fixture.Status != "ready_for_objective_return" ||
		fixture.DisplayState != "objective_return_final" ||
		!fixture.ReadyForHostDisplay ||
		!fixture.ReadyForObjectiveReturn ||
		fixture.IntermediateDisplay ||
		!fixture.FinalDisplay ||
		!fixture.ObjectiveLifecycleClosed ||
		!fixture.ObjectiveSatisfied ||
		fixture.HostViewRef != view.HostViewRef ||
		fixture.ObjectiveCloseoutPacketRef != view.ObjectiveCloseoutPacketRef ||
		fixture.ObjectiveCloseoutHandoffRef != view.ObjectiveCloseoutHandoffRef ||
		fixture.ObjectiveCloseoutReadbackRef != view.ObjectiveCloseoutReadbackRef ||
		fixture.CoreInvocationExecuted ||
		fixture.DurableWriteByCore ||
		fixture.ObjectiveStoreWriteByCore ||
		fixture.RunstoreWriteByCore ||
		fixture.NextHostAction != "return_objective_closed_lifecycle" {
		t.Fatalf("unexpected objective closeout blackbox fixture: %#v", fixture)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "adapter objective closeout blackbox fixture",
		RunnerEffect: fixture.RunnerEffect,
		PromptEffect: fixture.PromptEffect,
		Boundaries:   fixture.Boundaries,
		Payload:      fixture,
	}, "production_adapter_objective_closeout_blackbox_fixture", "objective_closeout_blackbox_fixture_projection_only", "host_cli_objective_closeout_display", "ready_for_objective_return")
	AssertNoCoreMutation(t, "adapter objective closeout blackbox fixture", fixture.CoreInvocationExecuted, fixture.DurableWriteByCore)
}

func TestProductionAdapterObjectiveCloseoutHostViewRejectsMismatchedRefs(t *testing.T) {
	handoff := productionAdapterReadyObjectiveCloseoutDurableHandoff()
	readback := productionAdapterRecordedObjectiveCloseoutReadback(handoff)
	readback.ObjectiveCloseoutPacketRef = "closeout:other_objective"

	view := BuildProductionAdapterObjectiveCloseoutHostView(ProductionAdapterObjectiveCloseoutHostViewInput{
		HostViewRef:             "view:metrics_objective_closeout",
		ObjectiveCloseoutPacket: handoffObjectiveCloseoutPacket(handoff),
		DurableHandoff:          handoff,
		Readback:                readback,
	})
	if view.Status != "blocked" ||
		view.ReadyForObjectiveReturn ||
		view.ObjectiveCloseoutPacketRef != handoff.ObjectiveCloseoutPacketRef ||
		view.AuthoritativeObjectiveCloseoutPacketRef != handoff.ObjectiveCloseoutPacketRef ||
		view.ObservedObjectiveCloseoutPacketRef != "closeout:other_objective" ||
		view.FailureClass != FailureVerificationFailed ||
		!productionAdapterStringContains(view.BlockedReasons, "closeout_view_packet_readback_packet_ref_mismatch") ||
		!productionAdapterMissingContains(view.MissingInputs, "host:objective_closeout_packet") {
		t.Fatalf("expected mismatched closeout refs to block host view, got %#v", view)
	}
}

func TestProductionAdapterObjectiveCloseoutHostViewRejectsUnsafeRefs(t *testing.T) {
	handoff := productionAdapterReadyObjectiveCloseoutDurableHandoff()
	readback := productionAdapterRecordedObjectiveCloseoutReadback(handoff)

	view := BuildProductionAdapterObjectiveCloseoutHostView(ProductionAdapterObjectiveCloseoutHostViewInput{
		HostViewRef:             "/tmp/raw-objective-closeout.json",
		ObjectiveCloseoutPacket: handoffObjectiveCloseoutPacket(handoff),
		DurableHandoff:          handoff,
		Readback:                readback,
	})
	if view.Status != "blocked" ||
		view.ReadyForHostDisplay ||
		view.ReadyForObjectiveReturn ||
		view.FailureClass != FailureEvidenceWeak ||
		!productionAdapterStringContains(view.BlockedReasons, "unsafe_input_ref") ||
		!productionAdapterMissingContains(view.MissingInputs, "host:display_safe_refs") {
		t.Fatalf("unsafe objective closeout host view = %#v", view)
	}
	AssertNoRawPayload(t, "unsafe objective closeout host view", view, "/tmp/raw-objective-closeout.json")
}

func productionAdapterReadyObjectiveCloseoutHostView() ProductionAdapterObjectiveCloseoutHostView {
	handoff := productionAdapterReadyObjectiveCloseoutDurableHandoff()
	return BuildProductionAdapterObjectiveCloseoutHostView(ProductionAdapterObjectiveCloseoutHostViewInput{
		HostViewRef:             "view:metrics_objective_closeout",
		ObjectiveCloseoutPacket: handoffObjectiveCloseoutPacket(handoff),
		DurableHandoff:          handoff,
		Readback:                productionAdapterRecordedObjectiveCloseoutReadback(handoff),
	})
}

func productionAdapterRecordedObjectiveCloseoutReadback(handoff ProductionAdapterObjectiveCloseoutDurableHandoff) ProductionAdapterObjectiveCloseoutReadback {
	return BuildProductionAdapterObjectiveCloseoutReadback(ProductionAdapterObjectiveCloseoutReadbackInput{
		ObjectiveCloseoutReadbackRef: "readback:metrics_objective_closeout",
		DurableHandoff:               handoff,
		AppliedDurableEventRef:       handoff.ExpectedDurableEventRef,
		AppliedRunstoreRef:           handoff.HostRunstoreRef,
		AppliedObjectiveStateRef:     handoff.ExpectedObjectiveStateRef,
		HostDurableWriteReported:     true,
		HostDurableWriteSucceeded:    true,
	})
}

func productionAdapterFailedObjectiveCloseoutReadback(handoff ProductionAdapterObjectiveCloseoutDurableHandoff) ProductionAdapterObjectiveCloseoutReadback {
	return BuildProductionAdapterObjectiveCloseoutReadback(ProductionAdapterObjectiveCloseoutReadbackInput{
		ObjectiveCloseoutReadbackRef: "readback:metrics_objective_closeout_failure",
		DurableHandoff:               handoff,
		FailureRef:                   "failure:metrics_objective_closeout",
		CompensationRef:              "compensation:metrics_objective_closeout_review",
		HostDurableWriteReported:     true,
		HostDurableWriteFailed:       true,
	})
}

func handoffObjectiveCloseoutPacket(handoff ProductionAdapterObjectiveCloseoutDurableHandoff) ProductionAdapterObjectiveCloseoutPacket {
	return ProductionAdapterObjectiveCloseoutPacket{
		ContractVersion:                 handoff.ContractVersion,
		Projected:                       true,
		Available:                       true,
		Status:                          VerificationSatisfied,
		Mode:                            "production_adapter_objective_closeout_packet",
		ReadyForHostCloseoutReview:      true,
		ReadyForObjectiveCompletion:     true,
		ObjectiveSatisfied:              handoff.ObjectiveSatisfied,
		VerificationSatisfied:           handoff.VerificationSatisfied,
		HostCloseoutConfirmed:           handoff.HostCloseoutConfirmed,
		AuthorizationBound:              handoff.AuthorizationBound,
		CompletionAuditBound:            handoff.CompletionAuditBound,
		ObjectiveCloseoutPacketRef:      handoff.ObjectiveCloseoutPacketRef,
		ObjectiveRef:                    handoff.ObjectiveRef,
		HostCloseoutConfirmationRef:     handoff.HostCloseoutConfirmationRef,
		CompletionAuditResultBindingRef: handoff.CompletionAuditResultBindingRef,
		CompletionAuditResultRef:        handoff.CompletionAuditResultRef,
		InvocationReportBindingRef:      handoff.InvocationReportBindingRef,
		AuthorizationPacketRef:          handoff.AuthorizationPacketRef,
		PreflightReviewPacketRef:        handoff.PreflightReviewPacketRef,
		AdapterRef:                      handoff.AdapterRef,
		DescriptorRef:                   handoff.DescriptorRef,
		InvocationRef:                   handoff.InvocationRef,
		ResultRef:                       handoff.ResultRef,
		ReadbackRef:                     handoff.ReadbackRef,
		CompletionHandoffRef:            handoff.CompletionHandoffRef,
		EvidenceRefs:                    cloneEvidenceRefs(handoff.EvidenceRefs),
		Verification:                    handoff.Verification.Clone(),
		FailureClass:                    FailureNone,
		Boundaries:                      AppendBoundaries(handoff.Boundaries, "production_adapter_objective_closeout_packet", "objective_closeout_projection_only", "host_owned_objective_closeout", "objective_closeout_confirmed", "ready_for_objective_completion"),
		NextHostAction:                  "return_objective_closeout",
		RunnerEffect:                    "none",
		PromptEffect:                    "none",
	}.Normalize()
}
