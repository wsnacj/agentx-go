package controlcontract

import (
	"testing"
)

func TestProductionAdapterCatalogReviewBlackboxFixtureReadyForResolutionHandoff(t *testing.T) {
	descriptor := productionAdapterTestDescriptor()
	registry := productionAdapterReadyRegistrySnapshot(descriptor)
	view := BuildProductionAdapterCatalogDiscoveryView(registry.CatalogSnapshot)
	packet := BuildProductionAdapterCatalogReviewPacket(ProductionAdapterCatalogReviewPacketInput{
		ReviewPacketRef:      "review:host_adapter_catalog",
		RegistrySnapshot:     registry,
		CatalogDiscoveryView: view,
	})
	selection := BuildProductionAdapterCatalogSelection(ProductionAdapterCatalogSelectionInput{
		CatalogSnapshot:      registry.CatalogSnapshot,
		SelectionRef:         "selection:host_metrics_adapter",
		SelectedAdapterRef:   descriptor.AdapterRef,
		SelectedSourceKind:   ReplannerSourceOperations,
		SelectedSourceRef:    "source:operations_metrics",
		SelectedCandidateRef: descriptor.SupportedCandidateRefs[0],
	})
	resolution := BuildProductionAdapterResolution(BuildProductionAdapterResolutionInputFromCatalogSelection(ProductionAdapterCatalogSelectionResolutionInput{
		CatalogSelection:        selection,
		ApplyEnvelopeReady:      true,
		ApplyEnvelopeRef:        "envelope:metrics_apply",
		HostPolicyRef:           "policy:local_readonly",
		ApprovalContextRefs:     []DisplaySafeRef{"approval:local_readonly"},
		BudgetRef:               "budget:local_probe",
		IdempotencyRef:          "idempotency:metrics_probe_1",
		AvailableCapabilityRefs: []DisplaySafeRef{"capability:local_metrics"},
		ConfirmedPolicyRefs:     []DisplaySafeRef{"policy:local_readonly"},
		ConfirmedApprovalRefs:   []DisplaySafeRef{"approval:local_readonly"},
	}))
	fixture := BuildProductionAdapterCatalogReviewBlackboxFixture(ProductionAdapterCatalogReviewBlackboxFixtureInput{
		FixtureRef:   "fixture:host_adapter_catalog_review",
		ReviewPacket: packet,
		Selection:    selection,
		Resolution:   resolution,
	})
	if fixture.Status != "ready_for_adapter_resolution_handoff" ||
		fixture.Mode != "production_adapter_catalog_review_blackbox_fixture" ||
		!fixture.ReadyForHostDisplay ||
		!fixture.ReadyForSelectionHandoff ||
		!fixture.ReadyForResolutionHandoff ||
		!fixture.ReadyForHostPreflight ||
		fixture.NextHostAction != "host_may_run_adapter_preflight" ||
		fixture.ReviewPacketRef != "review:host_adapter_catalog" ||
		fixture.SelectionRef != "selection:host_metrics_adapter" ||
		fixture.SelectedAdapterRef != descriptor.AdapterRef ||
		fixture.SelectedCandidateRef != descriptor.SupportedCandidateRefs[0] ||
		fixture.ResolutionStatus != HostActionReady ||
		!productionAdapterStringContains(fixture.DisplaySections, "adapter_catalog_provenance") ||
		!productionAdapterStringContains(fixture.DisplaySections, "adapter_resolution_handoff") {
		t.Fatalf("unexpected catalog review blackbox fixture: %#v", fixture)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "adapter catalog review blackbox fixture",
		RunnerEffect: fixture.RunnerEffect,
		PromptEffect: fixture.PromptEffect,
		Boundaries:   fixture.Boundaries,
		Payload:      fixture,
	}, "production_adapter_catalog_review_blackbox_fixture", "catalog_review_fixture_projection_only", "host_cli_display_fixture", "ready_for_host_adapter_catalog_display", "ready_for_adapter_resolution_handoff")
	AssertNoRawPayload(t, "adapter catalog review blackbox fixture", fixture, "/tmp/raw", "postgresql://secret")
}

func TestProductionAdapterCatalogReviewBlackboxFixtureShowsDisplayBeforeSelection(t *testing.T) {
	registry := productionAdapterReadyRegistrySnapshot(productionAdapterTestDescriptor())
	packet := BuildProductionAdapterCatalogReviewPacket(ProductionAdapterCatalogReviewPacketInput{
		ReviewPacketRef:  "review:host_adapter_catalog",
		RegistrySnapshot: registry,
	})
	fixture := BuildProductionAdapterCatalogReviewBlackboxFixture(ProductionAdapterCatalogReviewBlackboxFixtureInput{
		FixtureRef:   "fixture:host_adapter_catalog_review",
		ReviewPacket: packet,
	})
	if fixture.Status != "blocked" ||
		!fixture.ReadyForHostDisplay ||
		fixture.ReadyForSelectionHandoff ||
		fixture.ReadyForResolutionHandoff ||
		fixture.FailureClass != FailureConfigMissing ||
		!productionAdapterStringContains(fixture.BlockedReasons, "adapter_catalog_selection_not_ready") ||
		!productionAdapterStringContains(fixture.BlockedReasons, "adapter_resolution_not_ready") ||
		!productionAdapterMissingContains(fixture.MissingInputs, "host:adapter_catalog_selection") ||
		!productionAdapterMissingContains(fixture.MissingInputs, "host:adapter_resolution") {
		t.Fatalf("expected display-ready fixture blocked on selection/resolution, got %#v", fixture)
	}
}

func TestProductionAdapterCatalogReviewBlackboxFixtureBlocksMismatchAndUnsafeRefs(t *testing.T) {
	descriptor := productionAdapterTestDescriptor()
	registry := productionAdapterReadyRegistrySnapshot(descriptor)
	packet := BuildProductionAdapterCatalogReviewPacket(ProductionAdapterCatalogReviewPacketInput{
		ReviewPacketRef:  "review:host_adapter_catalog",
		RegistrySnapshot: registry,
	})
	selection := BuildProductionAdapterCatalogSelection(ProductionAdapterCatalogSelectionInput{
		CatalogSnapshot:      registry.CatalogSnapshot,
		SelectionRef:         "selection:host_metrics_adapter",
		SelectedAdapterRef:   descriptor.AdapterRef,
		SelectedSourceKind:   ReplannerSourceOperations,
		SelectedSourceRef:    "source:operations_metrics",
		SelectedCandidateRef: descriptor.SupportedCandidateRefs[0],
	})
	resolution := BuildProductionAdapterResolution(BuildProductionAdapterResolutionInputFromCatalogSelection(ProductionAdapterCatalogSelectionResolutionInput{
		CatalogSelection:        selection,
		ApplyEnvelopeReady:      true,
		ApplyEnvelopeRef:        "envelope:metrics_apply",
		HostPolicyRef:           "policy:local_readonly",
		ApprovalContextRefs:     []DisplaySafeRef{"approval:local_readonly"},
		BudgetRef:               "budget:local_probe",
		IdempotencyRef:          "idempotency:metrics_probe_1",
		AvailableCapabilityRefs: []DisplaySafeRef{"capability:local_metrics"},
		ConfirmedPolicyRefs:     []DisplaySafeRef{"policy:local_readonly"},
		ConfirmedApprovalRefs:   []DisplaySafeRef{"approval:local_readonly"},
	}))
	resolution.CatalogSelectionRef = "selection:other"
	mismatch := BuildProductionAdapterCatalogReviewBlackboxFixture(ProductionAdapterCatalogReviewBlackboxFixtureInput{
		FixtureRef:   "fixture:host_adapter_catalog_review",
		ReviewPacket: packet,
		Selection:    selection,
		Resolution:   resolution,
	})
	if mismatch.Status != "blocked" ||
		mismatch.ReadyForResolutionHandoff ||
		mismatch.FailureClass != FailureInvalidInput ||
		!productionAdapterStringContains(mismatch.BlockedReasons, "fixture_resolution_selection_ref_mismatch") ||
		!productionAdapterMissingContains(mismatch.MissingInputs, "host:adapter_resolution") {
		t.Fatalf("expected fixture mismatch block, got %#v", mismatch)
	}

	unsafe := BuildProductionAdapterCatalogReviewBlackboxFixture(ProductionAdapterCatalogReviewBlackboxFixtureInput{
		FixtureRef:   "/tmp/raw-fixture.json",
		ReviewPacket: packet,
		Selection:    selection,
		Resolution:   resolution,
	})
	if unsafe.Status != "blocked" ||
		unsafe.ReadyForHostDisplay ||
		unsafe.FailureClass != FailureEvidenceWeak ||
		!productionAdapterStringContains(unsafe.BlockedReasons, "unsafe_input_ref") ||
		!productionAdapterMissingContains(unsafe.MissingInputs, "host:display_safe_refs") {
		t.Fatalf("expected unsafe fixture block, got %#v", unsafe)
	}
	AssertNoRawPayload(t, "unsafe adapter catalog review fixture", unsafe, "/tmp/raw-fixture.json")
}

func TestProductionAdapterCatalogReviewBlackboxFixtureUnavailableWithoutReviewPacket(t *testing.T) {
	fixture := BuildProductionAdapterCatalogReviewBlackboxFixture(ProductionAdapterCatalogReviewBlackboxFixtureInput{})
	if fixture.Available ||
		fixture.Status != "unavailable" ||
		fixture.ReadyForHostDisplay ||
		fixture.NextHostAction != "provide_adapter_catalog_review_packet" {
		t.Fatalf("unexpected unavailable catalog review fixture: %#v", fixture)
	}
}
