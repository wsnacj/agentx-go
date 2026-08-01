package controlcontract

import (
	"testing"
)

func TestProductionAdapterRegistrySnapshotBuildsCatalogSnapshot(t *testing.T) {
	descriptor := productionAdapterTestDescriptor()
	registry := BuildProductionAdapterRegistrySnapshot(ProductionAdapterRegistrySnapshotInput{
		RegistrySnapshotRef:       "registry:host_adapters",
		ProviderRef:               "provider:host_adapter_registry",
		CatalogSnapshotRef:        "catalog:host_adapters",
		CatalogVersionRef:         "version:host_adapters_v1",
		CatalogDigestRef:          "digest:host_adapters_v1",
		ExpectedCatalogVersionRef: "version:host_adapters_v1",
		ExpectedCatalogDigestRef:  "digest:host_adapters_v1",
		MaxDescriptorCount:        4,
		HostPolicyRefs:            []DisplaySafeRef{"policy:host_adapter_catalog"},
		Descriptors:               []ProductionAdapterDescriptor{descriptor},
	})
	if registry.Status != HostActionReady ||
		!registry.ReadyForCatalogSnapshot ||
		registry.NextHostAction != "host_may_review_adapter_catalog" ||
		registry.DescriptorCount != 1 ||
		registry.ReadyDescriptorCount != 1 ||
		registry.ProviderRef != "provider:host_adapter_registry" ||
		registry.CatalogVersionRef != "version:host_adapters_v1" ||
		registry.CatalogDigestRef != "digest:host_adapters_v1" ||
		!registry.CatalogSnapshot.ReadyForHostSelection ||
		registry.CatalogSnapshot.ProviderRef != registry.ProviderRef ||
		registry.CatalogSnapshot.CatalogVersionRef != registry.CatalogVersionRef ||
		registry.CatalogSnapshot.CatalogDigestRef != registry.CatalogDigestRef ||
		registry.CatalogSnapshot.MaxDescriptorCount != 4 {
		t.Fatalf("unexpected registry snapshot: %#v", registry)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "production adapter registry snapshot",
		RunnerEffect: registry.RunnerEffect,
		PromptEffect: registry.PromptEffect,
		Boundaries:   registry.Boundaries,
		Payload:      registry,
	}, "production_adapter_registry_snapshot", "registry_snapshot_projection_only", "host_owned_adapter_registry", "adapter_catalog_provenance_gate", "ready_for_adapter_catalog_review")

	view := BuildProductionAdapterCatalogDiscoveryView(registry.CatalogSnapshot)
	if !view.Available ||
		view.Status != "ready_for_host_adapter_selection" ||
		view.ProviderRef != "provider:host_adapter_registry" ||
		view.CatalogVersionRef != "version:host_adapters_v1" ||
		view.CatalogDigestRef != "digest:host_adapters_v1" ||
		view.MaxDescriptorCount != 4 ||
		!productionAdapterDisplaySafeRefContains(view.DescriptorRefs, descriptor.AdapterRef) {
		t.Fatalf("unexpected registry-derived catalog discovery view: %#v", view)
	}
}

func TestProductionAdapterRegistrySnapshotBlocksMissingProvenanceAndLimit(t *testing.T) {
	descriptor := productionAdapterTestDescriptor()
	missing := BuildProductionAdapterRegistrySnapshot(ProductionAdapterRegistrySnapshotInput{
		RegistrySnapshotRef: "registry:host_adapters",
		CatalogSnapshotRef:  "catalog:host_adapters",
		Descriptors:         []ProductionAdapterDescriptor{descriptor},
	})
	if missing.Status != HostActionBlocked ||
		missing.ReadyForCatalogSnapshot ||
		missing.FailureClass != FailureConfigMissing ||
		!productionAdapterStringContains(missing.BlockedReasons, "adapter_registry_provider_ref_missing") ||
		!productionAdapterStringContains(missing.BlockedReasons, "adapter_catalog_version_ref_missing") ||
		!productionAdapterStringContains(missing.BlockedReasons, "adapter_catalog_digest_ref_missing") ||
		!productionAdapterStringContains(missing.BlockedReasons, "adapter_catalog_descriptor_limit_missing") ||
		!productionAdapterMissingContains(missing.MissingInputs, "host:adapter_registry_provider_ref") ||
		!productionAdapterMissingContains(missing.MissingInputs, "host:adapter_catalog_version_ref") ||
		!productionAdapterMissingContains(missing.MissingInputs, "host:adapter_catalog_digest_ref") ||
		!productionAdapterMissingContains(missing.MissingInputs, "host:adapter_catalog_descriptor_limit") ||
		missing.CatalogSnapshot.ReadyForHostSelection {
		t.Fatalf("expected missing provenance block, got %#v", missing)
	}
	if !productionAdapterStringContains(missing.CatalogSnapshot.BlockedReasons, "adapter_registry_snapshot_not_ready") {
		t.Fatalf("expected embedded catalog to be downgraded, got %#v", missing.CatalogSnapshot)
	}

	schedule := productionAdapterTestDescriptor()
	schedule.AdapterRef = "adapter:operations_schedule"
	tooMany := BuildProductionAdapterRegistrySnapshot(ProductionAdapterRegistrySnapshotInput{
		RegistrySnapshotRef: "registry:host_adapters",
		ProviderRef:         "provider:host_adapter_registry",
		CatalogSnapshotRef:  "catalog:host_adapters",
		CatalogVersionRef:   "version:host_adapters_v1",
		CatalogDigestRef:    "digest:host_adapters_v1",
		MaxDescriptorCount:  1,
		Descriptors:         []ProductionAdapterDescriptor{descriptor, schedule},
	})
	if tooMany.Status != HostActionBlocked ||
		tooMany.ReadyForCatalogSnapshot ||
		tooMany.FailureClass != FailurePolicyBlocked ||
		!productionAdapterStringContains(tooMany.BlockedReasons, "adapter_registry_descriptor_count_exceeded") ||
		!productionAdapterStringContains(tooMany.CatalogSnapshot.BlockedReasons, "adapter_catalog_descriptor_count_exceeded") ||
		!productionAdapterMissingContains(tooMany.MissingInputs, "host:adapter_catalog_descriptor_limit") ||
		tooMany.CatalogSnapshot.ReadyForHostSelection {
		t.Fatalf("expected descriptor count limit block, got %#v", tooMany)
	}
}

func TestProductionAdapterRegistrySnapshotBlocksStaleAndConflict(t *testing.T) {
	descriptor := productionAdapterTestDescriptor()
	registry := BuildProductionAdapterRegistrySnapshot(ProductionAdapterRegistrySnapshotInput{
		RegistrySnapshotRef:       "registry:host_adapters",
		ProviderRef:               "provider:host_adapter_registry",
		CatalogSnapshotRef:        "catalog:host_adapters",
		CatalogVersionRef:         "version:host_adapters_v1",
		CatalogDigestRef:          "digest:host_adapters_v1",
		ExpectedCatalogVersionRef: "version:host_adapters_v2",
		ExpectedCatalogDigestRef:  "digest:host_adapters_v2",
		MaxDescriptorCount:        4,
		Descriptors:               []ProductionAdapterDescriptor{descriptor},
	})
	if registry.Status != HostActionBlocked ||
		registry.ReadyForCatalogSnapshot ||
		registry.FailureClass != FailureVerificationFailed ||
		registry.NextHostAction != "refresh_adapter_registry_snapshot" ||
		!productionAdapterStringContains(registry.BlockedReasons, "adapter_registry_snapshot_stale") ||
		!productionAdapterStringContains(registry.BlockedReasons, "adapter_registry_snapshot_conflict") ||
		!productionAdapterMissingContains(registry.MissingInputs, "host:adapter_registry_refresh") ||
		!productionAdapterMissingContains(registry.MissingInputs, "host:adapter_registry_review") ||
		registry.CatalogSnapshot.ReadyForHostSelection {
		t.Fatalf("expected stale/conflict block, got %#v", registry)
	}
	if !productionAdapterStringContains(registry.CatalogSnapshot.BlockedReasons, "adapter_registry_snapshot_not_ready") {
		t.Fatalf("expected embedded catalog to be blocked after stale/conflict, got %#v", registry.CatalogSnapshot)
	}
}

func TestProductionAdapterRegistrySnapshotUnsafeRefsDoNotLeak(t *testing.T) {
	descriptor := productionAdapterTestDescriptor()
	descriptor.DisplaySafeOutputRefs = append(descriptor.DisplaySafeOutputRefs, "/tmp/raw-adapter-output.json")
	registry := BuildProductionAdapterRegistrySnapshot(ProductionAdapterRegistrySnapshotInput{
		RegistrySnapshotRef: "/tmp/raw-registry.json",
		ProviderRef:         "provider:host_adapter_registry",
		CatalogSnapshotRef:  "catalog:host_adapters",
		CatalogVersionRef:   "version:host_adapters_v1",
		CatalogDigestRef:    "digest:host_adapters_v1",
		MaxDescriptorCount:  4,
		Descriptors:         []ProductionAdapterDescriptor{descriptor},
	})
	if registry.Status != HostActionBlocked ||
		registry.ReadyForCatalogSnapshot ||
		registry.FailureClass != FailureEvidenceWeak ||
		!productionAdapterStringContains(registry.BlockedReasons, "unsafe_input_ref") ||
		!productionAdapterMissingContains(registry.MissingInputs, "host:display_safe_refs") {
		t.Fatalf("expected unsafe registry block, got %#v", registry)
	}
	AssertNoRawPayload(t, "unsafe adapter registry snapshot", registry, "/tmp/raw-registry.json", "/tmp/raw-adapter-output.json")
}
