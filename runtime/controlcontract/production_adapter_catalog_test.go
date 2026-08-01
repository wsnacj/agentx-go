package controlcontract

import (
	"testing"
)

func TestProductionAdapterCatalogSnapshotAndDiscoveryView(t *testing.T) {
	descriptor := productionAdapterTestDescriptor()
	schedule := productionAdapterTestDescriptor()
	schedule.AdapterRef = "adapter:operations_schedule"
	schedule.Kind = ProductionAdapterOperationsSchedule
	schedule.SupportedCandidateRefs = []DisplaySafeRef{"strategy:operations_schedule_review"}
	schedule.ProvidesCapabilityRefs = []DisplaySafeRef{"capability:operations_schedule_proposal"}
	schedule.RequiresCapabilityRefs = []DisplaySafeRef{"capability:operations_scheduler_host"}

	snapshot := BuildProductionAdapterCatalogSnapshot(ProductionAdapterCatalogSnapshotInput{
		CatalogSnapshotRef: "catalog:scene_adapters",
		Producer:           "host:adapter_catalog",
		ProviderRef:        "provider:scene_adapters",
		CatalogVersionRef:  "version:scene_adapters_v1",
		CatalogDigestRef:   "digest:scene_adapters_v1",
		MaxDescriptorCount: 4,
		HostPolicyRefs:     []DisplaySafeRef{"policy:host_adapter_catalog"},
		Descriptors:        []ProductionAdapterDescriptor{descriptor, schedule},
	})
	if snapshot.Status != HostActionReady ||
		!snapshot.ReadyForHostSelection ||
		snapshot.DescriptorCount != 2 ||
		snapshot.ReadyDescriptorCount != 2 ||
		snapshot.ProviderRef != "provider:scene_adapters" ||
		snapshot.CatalogVersionRef != "version:scene_adapters_v1" ||
		snapshot.CatalogDigestRef != "digest:scene_adapters_v1" ||
		snapshot.MaxDescriptorCount != 4 ||
		snapshot.NextHostAction != "host_may_select_adapter" ||
		!productionAdapterDisplaySafeRefContains(snapshot.DescriptorRefs, "adapter:operations_local_metrics") ||
		!productionAdapterDisplaySafeRefContains(snapshot.DescriptorRefs, "adapter:operations_schedule") ||
		!productionAdapterDisplaySafeRefContains(snapshot.PolicyRefs, "policy:host_adapter_catalog") ||
		!productionAdapterKindContains(snapshot.Kinds, ProductionAdapterOperationsSchedule) ||
		!replannerSourceKindContains(snapshot.SourceKinds, ReplannerSourceOperations) {
		t.Fatalf("unexpected catalog snapshot: %#v", snapshot)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "adapter catalog snapshot",
		RunnerEffect: snapshot.RunnerEffect,
		PromptEffect: snapshot.PromptEffect,
		Boundaries:   snapshot.Boundaries,
		Payload:      snapshot,
	}, "production_adapter_catalog_snapshot", "catalog_snapshot_projection_only", "host_owned_adapter_catalog", "ready_for_host_adapter_selection")

	view := BuildProductionAdapterCatalogDiscoveryView(snapshot)
	if !view.Available ||
		view.Status != "ready_for_host_adapter_selection" ||
		view.Mode != "production_adapter_catalog_discovery_view" ||
		!view.ReadyForHostSelection ||
		view.DescriptorCount != 2 ||
		view.ReadyDescriptorCount != 2 ||
		view.ProviderRef != "provider:scene_adapters" ||
		view.CatalogVersionRef != "version:scene_adapters_v1" ||
		view.CatalogDigestRef != "digest:scene_adapters_v1" ||
		view.MaxDescriptorCount != 4 ||
		view.NextHostAction != "host_may_select_adapter" ||
		!productionAdapterDisplaySafeRefContains(view.DescriptorRefs, "adapter:operations_schedule") {
		t.Fatalf("unexpected catalog discovery view: %#v", view)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "adapter catalog discovery view",
		RunnerEffect: view.RunnerEffect,
		PromptEffect: view.PromptEffect,
		Boundaries:   view.Boundaries,
		Payload:      view,
	}, "production_adapter_catalog_discovery_view", "catalog_discovery_view_only", "host_owned_adapter_catalog", "ready_for_host_adapter_selection")
	AssertNoRawPayload(t, "adapter catalog discovery view", view, "/Users/mason", "postgresql://secret", "raw adapter command")
}

func TestProductionAdapterCatalogSnapshotBlocksEmptyIncompleteAndDuplicateCatalog(t *testing.T) {
	empty := BuildProductionAdapterCatalogSnapshot(ProductionAdapterCatalogSnapshotInput{
		CatalogSnapshotRef: "catalog:empty",
		Producer:           "host:adapter_catalog",
	})
	if empty.Status != HostActionBlocked ||
		empty.ReadyForHostSelection ||
		empty.FailureClass != FailureHostAdapterMissing ||
		!productionAdapterStringContains(empty.BlockedReasons, "adapter_catalog_empty") ||
		!productionAdapterMissingContains(empty.MissingInputs, "host:adapter_catalog_descriptors") {
		t.Fatalf("expected empty catalog block, got %#v", empty)
	}

	incompleteDescriptor := productionAdapterTestDescriptor()
	incompleteDescriptor.InputContractRef = ""
	incomplete := BuildProductionAdapterCatalogSnapshot(ProductionAdapterCatalogSnapshotInput{
		CatalogSnapshotRef: "catalog:incomplete",
		Producer:           "host:adapter_catalog",
		Descriptors:        []ProductionAdapterDescriptor{incompleteDescriptor},
	})
	if incomplete.Status != HostActionBlocked ||
		incomplete.ReadyDescriptorCount != 0 ||
		incomplete.FailureClass != FailureConfigMissing ||
		!productionAdapterStringContains(incomplete.BlockedReasons, "adapter_input_contract_ref_missing") ||
		!productionAdapterMissingContains(incomplete.MissingInputs, "host:adapter_input_contract_ref") {
		t.Fatalf("expected incomplete descriptor block, got %#v", incomplete)
	}

	duplicate := BuildProductionAdapterCatalogSnapshot(ProductionAdapterCatalogSnapshotInput{
		CatalogSnapshotRef: "catalog:duplicate",
		Producer:           "host:adapter_catalog",
		Descriptors:        []ProductionAdapterDescriptor{productionAdapterTestDescriptor(), productionAdapterTestDescriptor()},
	})
	if duplicate.Status != HostActionBlocked ||
		duplicate.FailureClass != FailureInvalidInput ||
		!productionAdapterStringContains(duplicate.BlockedReasons, "adapter_catalog_duplicate_ref") ||
		!productionAdapterMissingContains(duplicate.MissingInputs, "host:adapter_catalog_unique_ref") {
		t.Fatalf("expected duplicate descriptor block, got %#v", duplicate)
	}

	descriptor := productionAdapterTestDescriptor()
	schedule := productionAdapterTestDescriptor()
	schedule.AdapterRef = "adapter:operations_schedule"
	schedule.Kind = ProductionAdapterOperationsSchedule
	tooMany := BuildProductionAdapterCatalogSnapshot(ProductionAdapterCatalogSnapshotInput{
		CatalogSnapshotRef: "catalog:too_many",
		Producer:           "host:adapter_catalog",
		MaxDescriptorCount: 1,
		Descriptors:        []ProductionAdapterDescriptor{descriptor, schedule},
	})
	if tooMany.Status != HostActionBlocked ||
		tooMany.ReadyForHostSelection ||
		tooMany.FailureClass != FailurePolicyBlocked ||
		!productionAdapterStringContains(tooMany.BlockedReasons, "adapter_catalog_descriptor_count_exceeded") ||
		!productionAdapterMissingContains(tooMany.MissingInputs, "host:adapter_catalog_descriptor_limit") {
		t.Fatalf("expected descriptor count limit block, got %#v", tooMany)
	}
}

func TestProductionAdapterCatalogSnapshotUnsafeRefsDoNotLeak(t *testing.T) {
	descriptor := productionAdapterTestDescriptor()
	descriptor.DisplaySafeOutputRefs = append(descriptor.DisplaySafeOutputRefs, "/tmp/raw-adapter-output.json")
	snapshot := BuildProductionAdapterCatalogSnapshot(ProductionAdapterCatalogSnapshotInput{
		CatalogSnapshotRef: "/tmp/raw-catalog.json",
		Producer:           "host:adapter_catalog",
		Descriptors:        []ProductionAdapterDescriptor{descriptor},
	})
	if snapshot.Status != HostActionBlocked ||
		snapshot.ReadyForHostSelection ||
		snapshot.FailureClass != FailureEvidenceWeak ||
		!productionAdapterStringContains(snapshot.BlockedReasons, "unsafe_input_ref") ||
		!productionAdapterMissingContains(snapshot.MissingInputs, "host:display_safe_refs") {
		t.Fatalf("expected unsafe catalog block, got %#v", snapshot)
	}
	view := BuildProductionAdapterCatalogDiscoveryView(snapshot)
	if view.Status != "blocked" ||
		view.ReadyForHostSelection ||
		view.FailureClass != FailureEvidenceWeak ||
		!productionAdapterStringContains(view.BlockedReasons, "unsafe_input_ref") ||
		!productionAdapterMissingContains(view.MissingInputs, "host:display_safe_refs") {
		t.Fatalf("expected unsafe catalog discovery view block, got %#v", view)
	}
	AssertNoRawPayload(t, "unsafe adapter catalog snapshot", snapshot, "/tmp/raw-catalog.json", "/tmp/raw-adapter-output.json")
	AssertNoRawPayload(t, "unsafe adapter catalog discovery view", view, "/tmp/raw-catalog.json", "/tmp/raw-adapter-output.json")
}

func TestProductionAdapterCatalogDiscoveryViewUnavailableForEmptyInput(t *testing.T) {
	view := BuildProductionAdapterCatalogDiscoveryView(ProductionAdapterCatalogSnapshot{})
	if view.Available ||
		view.Status != "unavailable" ||
		view.Mode != "production_adapter_catalog_discovery_view" ||
		view.RunnerEffect != "none" ||
		view.PromptEffect != "none" {
		t.Fatalf("unexpected unavailable catalog view: %#v", view)
	}
}

func productionAdapterDisplaySafeRefContains(values []DisplaySafeRef, want DisplaySafeRef) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func productionAdapterKindContains(values []ProductionAdapterKind, want ProductionAdapterKind) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
