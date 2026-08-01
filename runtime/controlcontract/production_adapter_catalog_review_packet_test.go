package controlcontract

import (
	"testing"
)

func TestProductionAdapterCatalogReviewPacketReadyForSelection(t *testing.T) {
	descriptor := productionAdapterTestDescriptor()
	registry := productionAdapterReadyRegistrySnapshot(descriptor)
	view := BuildProductionAdapterCatalogDiscoveryView(registry.CatalogSnapshot)
	packet := BuildProductionAdapterCatalogReviewPacket(ProductionAdapterCatalogReviewPacketInput{
		ReviewPacketRef:      "review:host_adapter_catalog",
		RegistrySnapshot:     registry,
		CatalogDiscoveryView: view,
	})
	if packet.Status != "ready_for_adapter_catalog_review" ||
		!packet.ReadyForHostReview ||
		!packet.ReadyForCatalogSelection ||
		packet.NextHostAction != "host_may_select_adapter" ||
		packet.RegistrySnapshotRef != "registry:host_adapters" ||
		packet.ProviderRef != "provider:host_adapter_registry" ||
		packet.CatalogSnapshotRef != "catalog:host_adapters" ||
		packet.CatalogVersionRef != "version:host_adapters_v1" ||
		packet.CatalogDigestRef != "digest:host_adapters_v1" ||
		packet.DescriptorCount != 1 ||
		packet.ReadyDescriptorCount != 1 ||
		!productionAdapterDisplaySafeRefContains(packet.DescriptorRefs, descriptor.AdapterRef) ||
		!productionAdapterMissingContains(packet.SelectionRequiredInputs, "host:selected_adapter_ref") ||
		!productionAdapterMissingContains(packet.SelectionRequiredInputs, "host:selected_candidate_strategy_ref") {
		t.Fatalf("unexpected catalog review packet: %#v", packet)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "production adapter catalog review packet",
		RunnerEffect: packet.RunnerEffect,
		PromptEffect: packet.PromptEffect,
		Boundaries:   packet.Boundaries,
		Payload:      packet,
	}, "production_adapter_catalog_review_packet", "catalog_review_packet_projection_only", "host_owned_adapter_catalog_review", "ready_for_adapter_catalog_review", "ready_for_catalog_selection")

	selection := BuildProductionAdapterCatalogSelection(ProductionAdapterCatalogSelectionInput{
		CatalogSnapshot:      registry.CatalogSnapshot,
		SelectionRef:         "selection:host_metrics_adapter",
		SelectedAdapterRef:   descriptor.AdapterRef,
		SelectedSourceKind:   ReplannerSourceOperations,
		SelectedSourceRef:    "source:operations_metrics",
		SelectedCandidateRef: descriptor.SupportedCandidateRefs[0],
	})
	if selection.Status != HostActionReady || !selection.ReadyForResolution {
		t.Fatalf("expected packet-reviewed catalog to remain selectable, got %#v", selection)
	}
}

func TestProductionAdapterCatalogReviewPacketDerivesDiscoveryView(t *testing.T) {
	packet := BuildProductionAdapterCatalogReviewPacket(ProductionAdapterCatalogReviewPacketInput{
		ReviewPacketRef:  "review:host_adapter_catalog",
		RegistrySnapshot: productionAdapterReadyRegistrySnapshot(productionAdapterTestDescriptor()),
	})
	if packet.Status != "ready_for_adapter_catalog_review" ||
		!packet.ReadyForHostReview ||
		!packet.ReadyForCatalogSelection ||
		packet.CatalogStatus != "ready_for_host_adapter_selection" ||
		packet.DescriptorCount != 1 {
		t.Fatalf("expected ready packet with derived view, got %#v", packet)
	}
}

func TestProductionAdapterCatalogReviewPacketBlocksMissingRefAndMismatchedView(t *testing.T) {
	descriptor := productionAdapterTestDescriptor()
	registry := productionAdapterReadyRegistrySnapshot(descriptor)
	missingRef := BuildProductionAdapterCatalogReviewPacket(ProductionAdapterCatalogReviewPacketInput{
		RegistrySnapshot: registry,
	})
	if missingRef.Status != "blocked" ||
		missingRef.ReadyForHostReview ||
		missingRef.FailureClass != FailureEvidenceMissing ||
		!productionAdapterStringContains(missingRef.BlockedReasons, "adapter_catalog_review_packet_ref_missing") ||
		!productionAdapterMissingContains(missingRef.MissingInputs, "host:adapter_catalog_review_packet_ref") {
		t.Fatalf("expected missing review packet ref block, got %#v", missingRef)
	}

	mismatchedView := BuildProductionAdapterCatalogDiscoveryView(registry.CatalogSnapshot)
	mismatchedView.ProviderRef = "provider:other_registry"
	mismatchedView.CatalogVersionRef = "version:other_adapters"
	mismatch := BuildProductionAdapterCatalogReviewPacket(ProductionAdapterCatalogReviewPacketInput{
		ReviewPacketRef:      "review:host_adapter_catalog",
		RegistrySnapshot:     registry,
		CatalogDiscoveryView: mismatchedView,
	})
	if mismatch.Status != "blocked" ||
		mismatch.ReadyForHostReview ||
		mismatch.FailureClass != FailureInvalidInput ||
		!productionAdapterStringContains(mismatch.BlockedReasons, "catalog_review_provider_ref_mismatch") ||
		!productionAdapterStringContains(mismatch.BlockedReasons, "catalog_review_version_ref_mismatch") ||
		!productionAdapterMissingContains(mismatch.MissingInputs, "host:adapter_catalog_discovery_view") {
		t.Fatalf("expected catalog review mismatch block, got %#v", mismatch)
	}
}

func TestProductionAdapterCatalogReviewPacketBlocksStaleRegistry(t *testing.T) {
	registry := BuildProductionAdapterRegistrySnapshot(ProductionAdapterRegistrySnapshotInput{
		RegistrySnapshotRef:       "registry:host_adapters",
		ProviderRef:               "provider:host_adapter_registry",
		CatalogSnapshotRef:        "catalog:host_adapters",
		CatalogVersionRef:         "version:host_adapters_v1",
		CatalogDigestRef:          "digest:host_adapters_v1",
		ExpectedCatalogVersionRef: "version:host_adapters_v2",
		MaxDescriptorCount:        4,
		Descriptors:               []ProductionAdapterDescriptor{productionAdapterTestDescriptor()},
	})
	packet := BuildProductionAdapterCatalogReviewPacket(ProductionAdapterCatalogReviewPacketInput{
		ReviewPacketRef:  "review:host_adapter_catalog",
		RegistrySnapshot: registry,
	})
	if packet.Status != "blocked" ||
		packet.ReadyForHostReview ||
		packet.FailureClass != FailureVerificationFailed ||
		!productionAdapterStringContains(packet.BlockedReasons, "adapter_registry_snapshot_not_ready") ||
		!productionAdapterStringContains(packet.BlockedReasons, "adapter_catalog_discovery_not_ready") ||
		!productionAdapterMissingContains(packet.MissingInputs, "host:adapter_registry_snapshot") {
		t.Fatalf("expected stale registry review block, got %#v", packet)
	}
}

func TestProductionAdapterCatalogReviewPacketUnsafeRefsDoNotLeak(t *testing.T) {
	packet := BuildProductionAdapterCatalogReviewPacket(ProductionAdapterCatalogReviewPacketInput{
		ReviewPacketRef:  "/tmp/raw-review.json",
		RegistrySnapshot: productionAdapterReadyRegistrySnapshot(productionAdapterTestDescriptor()),
	})
	if packet.Status != "blocked" ||
		packet.ReadyForHostReview ||
		packet.FailureClass != FailureEvidenceWeak ||
		!productionAdapterStringContains(packet.BlockedReasons, "unsafe_input_ref") ||
		!productionAdapterMissingContains(packet.MissingInputs, "host:display_safe_refs") {
		t.Fatalf("expected unsafe review packet block, got %#v", packet)
	}
	AssertNoRawPayload(t, "unsafe adapter catalog review packet", packet, "/tmp/raw-review.json")
}

func TestProductionAdapterCatalogReviewPacketUnavailableWithoutRegistry(t *testing.T) {
	packet := BuildProductionAdapterCatalogReviewPacket(ProductionAdapterCatalogReviewPacketInput{})
	if packet.Available ||
		packet.Status != "unavailable" ||
		packet.ReadyForHostReview ||
		packet.NextHostAction != "provide_adapter_registry_snapshot" {
		t.Fatalf("unexpected unavailable catalog review packet: %#v", packet)
	}
}

func productionAdapterReadyRegistrySnapshot(descriptor ProductionAdapterDescriptor) ProductionAdapterRegistrySnapshot {
	return BuildProductionAdapterRegistrySnapshot(ProductionAdapterRegistrySnapshotInput{
		RegistrySnapshotRef: "registry:host_adapters",
		ProviderRef:         "provider:host_adapter_registry",
		CatalogSnapshotRef:  "catalog:host_adapters",
		CatalogVersionRef:   "version:host_adapters_v1",
		CatalogDigestRef:    "digest:host_adapters_v1",
		MaxDescriptorCount:  4,
		HostPolicyRefs:      []DisplaySafeRef{"policy:host_adapter_catalog"},
		Descriptors:         []ProductionAdapterDescriptor{descriptor},
	})
}
