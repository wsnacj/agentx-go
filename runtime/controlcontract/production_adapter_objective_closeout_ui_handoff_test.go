package controlcontract

import (
	"encoding/json"
	"testing"
)

func TestProductionAdapterObjectiveCloseoutHostUIHandoffPendingDurableApply(t *testing.T) {
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
	if uiHandoff.Status != "ready_for_host_durable_apply_handoff" ||
		uiHandoff.DisplayState != "durable_apply_pending" ||
		uiHandoff.DisplayStage != "intermediate" ||
		!productionAdapterStringContains(uiHandoff.DisplaySteps, "host_durable_apply") ||
		!uiHandoff.ReadyForHostDisplay ||
		!uiHandoff.ReadyForHostDurableApply ||
		uiHandoff.ReadyForObjectiveReturn ||
		!uiHandoff.IntermediateDisplay ||
		uiHandoff.FinalDisplay ||
		uiHandoff.PrimaryDisplayRef != fixture.FixtureRef ||
		uiHandoff.DisplayFixtureRef != fixture.FixtureRef ||
		uiHandoff.AuthoritativeObjectiveCloseoutPacketRef != handoff.ObjectiveCloseoutPacketRef ||
		uiHandoff.ObservedObjectiveCloseoutPacketRef != "" ||
		!productionAdapterMissingContains(uiHandoff.DisplayProgressMissingInputs, "host:objective_closeout_readback") ||
		!productionAdapterStringContains(uiHandoff.DisplayProgressBlockedReasons, "objective_closeout_readback_not_ready") ||
		len(uiHandoff.MissingInputs) != 0 ||
		len(uiHandoff.BlockedReasons) != 0 ||
		uiHandoff.CoreInvocationExecuted ||
		uiHandoff.DurableWriteByCore ||
		uiHandoff.ObjectiveStoreWriteByCore ||
		uiHandoff.RunstoreWriteByCore ||
		uiHandoff.NextHostAction != "host_may_apply_objective_closeout" {
		t.Fatalf("unexpected pending objective closeout host UI handoff: %#v", uiHandoff)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "adapter objective closeout pending host UI handoff",
		RunnerEffect: uiHandoff.RunnerEffect,
		PromptEffect: uiHandoff.PromptEffect,
		Boundaries:   uiHandoff.Boundaries,
		Payload:      uiHandoff,
	}, "production_adapter_objective_closeout_host_ui_handoff", "objective_closeout_host_ui_handoff_projection_only", "host_ui_closeout_handoff", "ready_for_host_durable_apply_handoff")
	AssertNoCoreMutation(t, "adapter objective closeout pending host UI handoff", uiHandoff.CoreInvocationExecuted, uiHandoff.DurableWriteByCore)
}

func TestProductionAdapterObjectiveCloseoutHostUIHandoffFinalReturn(t *testing.T) {
	view := productionAdapterReadyObjectiveCloseoutHostView()
	fixture := BuildProductionAdapterObjectiveCloseoutBlackboxFixture(ProductionAdapterObjectiveCloseoutBlackboxFixtureInput{
		FixtureRef: "fixture:metrics_objective_closeout",
		HostView:   view,
	})

	uiHandoff := BuildProductionAdapterObjectiveCloseoutHostUIHandoff(ProductionAdapterObjectiveCloseoutHostUIHandoffInput{
		HostUIHandoffRef: "ui:metrics_objective_closeout",
		HostView:         view,
		DisplayFixture:   fixture,
	})
	if uiHandoff.Status != "ready_for_objective_return_handoff" ||
		uiHandoff.DisplayState != "objective_return_final" ||
		uiHandoff.DisplayStage != "final" ||
		!productionAdapterStringContains(uiHandoff.DisplaySteps, "objective_return") ||
		!uiHandoff.ReadyForHostDisplay ||
		uiHandoff.ReadyForHostDurableApply ||
		!uiHandoff.ReadyForObjectiveReturn ||
		uiHandoff.IntermediateDisplay ||
		!uiHandoff.FinalDisplay ||
		uiHandoff.FailureReviewDisplay ||
		!uiHandoff.ObjectiveLifecycleClosed ||
		!uiHandoff.ObjectiveSatisfied ||
		uiHandoff.PrimaryDisplayRef != fixture.FixtureRef ||
		uiHandoff.DisplayFixtureRef != fixture.FixtureRef ||
		uiHandoff.ObservedObjectiveCloseoutPacketRef != view.ObservedObjectiveCloseoutPacketRef ||
		uiHandoff.ObservedAppliedDurableEventRef != view.ObservedAppliedDurableEventRef ||
		uiHandoff.NextHostAction != "return_objective_closed_lifecycle" {
		t.Fatalf("unexpected final objective closeout host UI handoff: %#v", uiHandoff)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "adapter objective closeout final host UI handoff",
		RunnerEffect: uiHandoff.RunnerEffect,
		PromptEffect: uiHandoff.PromptEffect,
		Boundaries:   uiHandoff.Boundaries,
		Payload:      uiHandoff,
	}, "production_adapter_objective_closeout_host_ui_handoff", "objective_closeout_host_ui_handoff_projection_only", "host_ui_closeout_handoff", "ready_for_objective_return_handoff")
	AssertNoCoreMutation(t, "adapter objective closeout final host UI handoff", uiHandoff.CoreInvocationExecuted, uiHandoff.DurableWriteByCore)
}

func TestProductionAdapterObjectiveCloseoutHostUIHandoffFailureReview(t *testing.T) {
	handoff := productionAdapterReadyObjectiveCloseoutDurableHandoff()
	view := BuildProductionAdapterObjectiveCloseoutHostView(ProductionAdapterObjectiveCloseoutHostViewInput{
		HostViewRef:             "view:metrics_objective_closeout_failure",
		ObjectiveCloseoutPacket: handoffObjectiveCloseoutPacket(handoff),
		DurableHandoff:          handoff,
		Readback:                productionAdapterFailedObjectiveCloseoutReadback(handoff),
	})
	review := BuildProductionAdapterObjectiveCloseoutFailureReviewPacket(ProductionAdapterObjectiveCloseoutFailureReviewPacketInput{
		FailureReviewPacketRef: "review:metrics_objective_closeout_failure",
		HostView:               view,
	})
	fixture := BuildProductionAdapterObjectiveCloseoutFailureReviewBlackboxFixture(ProductionAdapterObjectiveCloseoutFailureReviewBlackboxFixtureInput{
		FixtureRef:          "fixture:metrics_objective_closeout_failure",
		FailureReviewPacket: review,
	})

	uiHandoff := BuildProductionAdapterObjectiveCloseoutHostUIHandoff(ProductionAdapterObjectiveCloseoutHostUIHandoffInput{
		HostUIHandoffRef:     "ui:metrics_objective_closeout_failure",
		HostView:             view,
		FailureReview:        review,
		FailureReviewFixture: fixture,
	})
	if uiHandoff.Status != "ready_for_objective_closeout_failure_review_handoff" ||
		uiHandoff.DisplayState != "durable_failure_review" ||
		uiHandoff.DisplayStage != "failure_review" ||
		!productionAdapterStringContains(uiHandoff.DisplaySteps, "failure_review") ||
		!productionAdapterStringContains(uiHandoff.DisplaySteps, "compensation_review") ||
		!uiHandoff.ReadyForHostDisplay ||
		uiHandoff.ReadyForHostDurableApply ||
		!uiHandoff.ReadyForFailureReview ||
		!uiHandoff.ReadyForCompensationReview ||
		uiHandoff.ReadyForObjectiveReturn ||
		!uiHandoff.IntermediateDisplay ||
		uiHandoff.FinalDisplay ||
		!uiHandoff.FailureReviewDisplay ||
		uiHandoff.ObjectiveLifecycleClosed ||
		uiHandoff.ObjectiveSatisfied ||
		uiHandoff.PrimaryDisplayRef != fixture.FixtureRef ||
		uiHandoff.FailureReviewPacketRef != review.FailureReviewPacketRef ||
		uiHandoff.FailureReviewFixtureRef != fixture.FixtureRef ||
		uiHandoff.FailureRef != view.FailureRef ||
		uiHandoff.CompensationRef != view.CompensationRef ||
		uiHandoff.NextHostAction != "review_objective_closeout_durable_failure" {
		t.Fatalf("unexpected failure objective closeout host UI handoff: %#v", uiHandoff)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "adapter objective closeout failure host UI handoff",
		RunnerEffect: uiHandoff.RunnerEffect,
		PromptEffect: uiHandoff.PromptEffect,
		Boundaries:   uiHandoff.Boundaries,
		Payload:      uiHandoff,
	}, "production_adapter_objective_closeout_host_ui_handoff", "objective_closeout_host_ui_handoff_projection_only", "host_ui_closeout_handoff", "ready_for_objective_closeout_failure_review_handoff", "compensation_not_executed")
	AssertNoCoreMutation(t, "adapter objective closeout failure host UI handoff", uiHandoff.CoreInvocationExecuted, uiHandoff.DurableWriteByCore)
}

func TestProductionAdapterObjectiveCloseoutHostUIHandoffRequiresFailureReviewPacket(t *testing.T) {
	handoff := productionAdapterReadyObjectiveCloseoutDurableHandoff()
	view := BuildProductionAdapterObjectiveCloseoutHostView(ProductionAdapterObjectiveCloseoutHostViewInput{
		HostViewRef:             "view:metrics_objective_closeout_failure",
		ObjectiveCloseoutPacket: handoffObjectiveCloseoutPacket(handoff),
		DurableHandoff:          handoff,
		Readback:                productionAdapterFailedObjectiveCloseoutReadback(handoff),
	})

	uiHandoff := BuildProductionAdapterObjectiveCloseoutHostUIHandoff(ProductionAdapterObjectiveCloseoutHostUIHandoffInput{
		HostUIHandoffRef: "ui:metrics_objective_closeout_failure",
		HostView:         view,
	})
	if uiHandoff.Status != "blocked" ||
		uiHandoff.ReadyForFailureReview ||
		!productionAdapterStringContains(uiHandoff.BlockedReasons, "objective_closeout_failure_review_packet_not_ready") ||
		!productionAdapterMissingContains(uiHandoff.MissingInputs, "host:objective_closeout_failure_review_packet") {
		t.Fatalf("expected missing failure review packet to block UI handoff, got %#v", uiHandoff)
	}
}

func TestProductionAdapterObjectiveCloseoutHostUIHandoffRejectsMismatchedFixture(t *testing.T) {
	view := productionAdapterReadyObjectiveCloseoutHostView()
	fixture := BuildProductionAdapterObjectiveCloseoutBlackboxFixture(ProductionAdapterObjectiveCloseoutBlackboxFixtureInput{
		FixtureRef: "fixture:metrics_objective_closeout",
		HostView:   view,
	})
	fixture.HostViewRef = "view:other_objective_closeout"

	uiHandoff := BuildProductionAdapterObjectiveCloseoutHostUIHandoff(ProductionAdapterObjectiveCloseoutHostUIHandoffInput{
		HostUIHandoffRef: "ui:metrics_objective_closeout",
		HostView:         view,
		DisplayFixture:   fixture,
	})
	if uiHandoff.Status != "blocked" ||
		uiHandoff.ReadyForObjectiveReturn ||
		!productionAdapterStringContains(uiHandoff.BlockedReasons, "closeout_ui_handoff_fixture_host_view_ref_mismatch") ||
		!productionAdapterMissingContains(uiHandoff.MissingInputs, "host:objective_closeout_host_view") {
		t.Fatalf("expected fixture mismatch to block UI handoff, got %#v", uiHandoff)
	}
}

func TestProductionAdapterObjectiveCloseoutHostUIHandoffJSONCompatibility(t *testing.T) {
	view := productionAdapterReadyObjectiveCloseoutHostView()
	fixture := BuildProductionAdapterObjectiveCloseoutBlackboxFixture(ProductionAdapterObjectiveCloseoutBlackboxFixtureInput{
		FixtureRef: "fixture:metrics_objective_closeout",
		HostView:   view,
	})
	uiHandoff := BuildProductionAdapterObjectiveCloseoutHostUIHandoff(ProductionAdapterObjectiveCloseoutHostUIHandoffInput{
		HostUIHandoffRef: "ui:metrics_objective_closeout",
		HostView:         view,
		DisplayFixture:   fixture,
	})

	raw, err := json.Marshal(uiHandoff)
	if err != nil {
		t.Fatalf("marshal objective closeout host UI handoff: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("unmarshal objective closeout host UI handoff: %v", err)
	}
	for _, key := range []string{
		"display_state",
		"display_stage",
		"display_steps",
		"primary_display_ref",
		"authoritative_objective_closeout_packet_ref",
		"observed_objective_closeout_packet_ref",
		"ready_for_objective_return",
		"next_host_action",
	} {
		if _, ok := payload[key]; !ok {
			t.Fatalf("expected stable JSON key %q in %s", key, string(raw))
		}
	}
	if payload["display_state"] != "objective_return_final" ||
		payload["display_stage"] != "final" ||
		payload["next_host_action"] != "return_objective_closed_lifecycle" {
		t.Fatalf("unexpected stable JSON payload: %s", string(raw))
	}
	AssertNoRawPayload(t, "objective closeout host UI handoff JSON", uiHandoff, "/Users/mason", "postgresql://secret", "raw local host task")
}
