package controlcontract

import (
	"testing"
)

func TestProductionAdapterCatalogSelectionBuildsResolutionHandoff(t *testing.T) {
	descriptor := productionAdapterTestDescriptor()
	snapshot := BuildProductionAdapterCatalogSnapshot(ProductionAdapterCatalogSnapshotInput{
		CatalogSnapshotRef: "catalog:host_adapters",
		Producer:           "host:adapter_catalog",
		Descriptors:        []ProductionAdapterDescriptor{descriptor},
	})
	selection := BuildProductionAdapterCatalogSelection(ProductionAdapterCatalogSelectionInput{
		CatalogSnapshot:      snapshot,
		SelectionRef:         "selection:metrics_adapter",
		SelectedAdapterRef:   descriptor.AdapterRef,
		SelectedSourceKind:   ReplannerSourceOperations,
		SelectedSourceRef:    "source:operations_metrics",
		SelectedCandidateRef: "strategy:operations_metric_collect",
	})
	if selection.Status != HostActionReady ||
		!selection.ReadyForResolution ||
		selection.NextHostAction != "build_adapter_resolution" ||
		selection.SelectedDescriptor.AdapterRef != descriptor.AdapterRef ||
		selection.CatalogSnapshotRef != "catalog:host_adapters" {
		t.Fatalf("unexpected catalog selection: %#v", selection)
	}
	AssertHostOwnedProjectionOnly(t, Projection[Boundary]{
		Name:         "adapter catalog selection",
		RunnerEffect: selection.RunnerEffect,
		PromptEffect: selection.PromptEffect,
		Boundaries:   selection.Boundaries,
		Payload:      selection,
	}, "production_adapter_catalog_selection", "catalog_selection_projection_only", "catalog_bound_adapter_selection", "ready_for_adapter_resolution")

	resolutionInput := BuildProductionAdapterResolutionInputFromCatalogSelection(ProductionAdapterCatalogSelectionResolutionInput{
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
	})
	resolution := BuildProductionAdapterResolution(resolutionInput)
	if resolution.Status != HostActionReady ||
		!resolution.ReadyForHostPreflight ||
		resolution.CatalogSelectionRef != "selection:metrics_adapter" ||
		resolution.CatalogSnapshotRef != "catalog:host_adapters" ||
		resolution.RequestedAdapterRef != descriptor.AdapterRef ||
		resolution.SelectedCandidateRef != "strategy:operations_metric_collect" {
		t.Fatalf("unexpected resolution from catalog selection: %#v", resolution)
	}
}

func TestProductionAdapterCatalogSelectionSupportsWorkflowDispatchCandidate(t *testing.T) {
	descriptor := productionAdapterTestDescriptor()
	descriptor.AdapterRef = "adapter:workflow_runtime_dispatch"
	descriptor.Kind = ProductionAdapterWorkflowDispatch
	descriptor.SupportedSourceKinds = []ReplannerSourceKind{ReplannerSourceWorkflow}
	descriptor.SupportedCandidateRefs = []DisplaySafeRef{"strategy:workflow_runtime"}
	descriptor.ProvidesCapabilityRefs = []DisplaySafeRef{"capability:workflow_runtime_result"}
	descriptor.RequiresCapabilityRefs = []DisplaySafeRef{"capability:workflow_runtime_backend"}
	descriptor.InputContractRef = "contract:workflow_runtime_input"
	descriptor.OutputContractRef = "contract:workflow_runtime_output"
	descriptor.ReadbackContractRef = "contract:workflow_runtime_readback"
	descriptor.RequiredPolicyRefs = []DisplaySafeRef{"policy:workflow_runtime_host"}
	descriptor.RequiredApprovalRefs = []DisplaySafeRef{"approval:workflow_runtime_host"}
	descriptor.RequiredBudgetRef = "budget:workflow_runtime"
	descriptor.IdempotencyContractRef = "idempotency:workflow_runtime"
	descriptor.SideEffectClass = "host_workflow_runtime"
	descriptor.DisplaySafeInputRefs = []DisplaySafeRef{"input:workflow_runtime"}
	descriptor.DisplaySafeOutputRefs = []DisplaySafeRef{"output:workflow_runtime_result"}

	snapshot := BuildProductionAdapterCatalogSnapshot(ProductionAdapterCatalogSnapshotInput{
		CatalogSnapshotRef: "catalog:host_workflow_adapters",
		Producer:           "host:adapter_catalog",
		Descriptors:        []ProductionAdapterDescriptor{descriptor},
	})
	selection := BuildProductionAdapterCatalogSelection(ProductionAdapterCatalogSelectionInput{
		CatalogSnapshot:      snapshot,
		SelectionRef:         "selection:workflow_runtime_adapter",
		SelectedAdapterRef:   descriptor.AdapterRef,
		SelectedSourceKind:   ReplannerSourceWorkflow,
		SelectedSourceRef:    "workflow:runtime_objective",
		SelectedCandidateRef: "strategy:workflow_runtime",
	})
	if selection.Status != HostActionReady ||
		!selection.ReadyForResolution ||
		selection.SelectedDescriptor.Kind != ProductionAdapterWorkflowDispatch ||
		selection.SelectedSourceKind != ReplannerSourceWorkflow ||
		selection.SelectedCandidateRef != "strategy:workflow_runtime" {
		t.Fatalf("unexpected workflow catalog selection: %#v", selection)
	}

	resolutionInput := BuildProductionAdapterResolutionInputFromCatalogSelection(ProductionAdapterCatalogSelectionResolutionInput{
		CatalogSelection:        selection,
		ApplyEnvelopeReady:      true,
		ApplyEnvelopeRef:        "envelope:workflow_runtime_apply",
		HostPolicyRef:           "policy:workflow_runtime_host",
		ApprovalContextRefs:     []DisplaySafeRef{"approval:workflow_runtime_host"},
		BudgetRef:               "budget:workflow_runtime",
		IdempotencyRef:          "idempotency:workflow_runtime_1",
		AvailableCapabilityRefs: []DisplaySafeRef{"capability:workflow_runtime_backend"},
		ConfirmedPolicyRefs:     []DisplaySafeRef{"policy:workflow_runtime_host"},
		ConfirmedApprovalRefs:   []DisplaySafeRef{"approval:workflow_runtime_host"},
	})
	resolution := BuildProductionAdapterResolution(resolutionInput)
	if resolution.Status != HostActionReady ||
		!resolution.ReadyForHostPreflight ||
		resolution.SelectedSourceKind != ReplannerSourceWorkflow ||
		resolution.SelectedCandidateRef != "strategy:workflow_runtime" ||
		resolution.RequestedAdapterRef != descriptor.AdapterRef {
		t.Fatalf("unexpected workflow adapter resolution: %#v", resolution)
	}
}

func TestProductionAdapterCatalogSelectionBlocksMissingProducerAndUnknownAdapter(t *testing.T) {
	descriptor := productionAdapterTestDescriptor()
	missingProducerSnapshot := BuildProductionAdapterCatalogSnapshot(ProductionAdapterCatalogSnapshotInput{
		CatalogSnapshotRef: "catalog:host_adapters",
		Descriptors:        []ProductionAdapterDescriptor{descriptor},
	})
	missingProducer := BuildProductionAdapterCatalogSelection(ProductionAdapterCatalogSelectionInput{
		CatalogSnapshot:      missingProducerSnapshot,
		SelectionRef:         "selection:metrics_adapter",
		SelectedAdapterRef:   descriptor.AdapterRef,
		SelectedSourceKind:   ReplannerSourceOperations,
		SelectedSourceRef:    "source:operations_metrics",
		SelectedCandidateRef: "strategy:operations_metric_collect",
	})
	if missingProducer.Status != HostActionBlocked ||
		missingProducer.ReadyForResolution ||
		missingProducer.FailureClass != FailureConfigMissing ||
		!productionAdapterStringContains(missingProducer.BlockedReasons, "adapter_catalog_producer_missing") ||
		!productionAdapterMissingContains(missingProducer.MissingInputs, "host:adapter_catalog_producer") {
		t.Fatalf("expected missing producer block, got %#v", missingProducer)
	}

	readySnapshot := BuildProductionAdapterCatalogSnapshot(ProductionAdapterCatalogSnapshotInput{
		CatalogSnapshotRef: "catalog:host_adapters",
		Producer:           "host:adapter_catalog",
		Descriptors:        []ProductionAdapterDescriptor{descriptor},
	})
	unknown := BuildProductionAdapterCatalogSelection(ProductionAdapterCatalogSelectionInput{
		CatalogSnapshot:      readySnapshot,
		SelectionRef:         "selection:missing_adapter",
		SelectedAdapterRef:   "adapter:not_in_catalog",
		SelectedSourceKind:   ReplannerSourceOperations,
		SelectedSourceRef:    "source:operations_metrics",
		SelectedCandidateRef: "strategy:operations_metric_collect",
	})
	if unknown.Status != HostActionBlocked ||
		unknown.ReadyForResolution ||
		unknown.FailureClass != FailureHostAdapterMissing ||
		!productionAdapterStringContains(unknown.BlockedReasons, "adapter_not_in_catalog") ||
		!productionAdapterMissingContains(unknown.MissingInputs, "host:selected_adapter_ref") {
		t.Fatalf("expected unknown adapter block, got %#v", unknown)
	}
}

func TestProductionAdapterCatalogSelectionBlocksCandidateMismatchAndUnsafeRefs(t *testing.T) {
	descriptor := productionAdapterTestDescriptor()
	snapshot := BuildProductionAdapterCatalogSnapshot(ProductionAdapterCatalogSnapshotInput{
		CatalogSnapshotRef: "catalog:host_adapters",
		Producer:           "host:adapter_catalog",
		Descriptors:        []ProductionAdapterDescriptor{descriptor},
	})
	mismatch := BuildProductionAdapterCatalogSelection(ProductionAdapterCatalogSelectionInput{
		CatalogSnapshot:      snapshot,
		SelectionRef:         "selection:metrics_adapter",
		SelectedAdapterRef:   descriptor.AdapterRef,
		SelectedSourceKind:   ReplannerSourceOperations,
		SelectedSourceRef:    "source:operations_metrics",
		SelectedCandidateRef: "strategy:other",
	})
	if mismatch.Status != HostActionBlocked ||
		mismatch.FailureClass != FailureUnsupportedOperation ||
		!productionAdapterStringContains(mismatch.BlockedReasons, "adapter_candidate_mismatch") ||
		!productionAdapterMissingContains(mismatch.MissingInputs, "host:adapter_candidate_review") {
		t.Fatalf("expected candidate mismatch block, got %#v", mismatch)
	}

	unsafe := BuildProductionAdapterCatalogSelection(ProductionAdapterCatalogSelectionInput{
		CatalogSnapshot:      snapshot,
		SelectionRef:         "/tmp/raw-selection.json",
		SelectedAdapterRef:   descriptor.AdapterRef,
		SelectedSourceKind:   ReplannerSourceOperations,
		SelectedSourceRef:    "postgresql://secret@example.invalid/db",
		SelectedCandidateRef: "strategy:operations_metric_collect",
	})
	if unsafe.Status != HostActionBlocked ||
		unsafe.ReadyForResolution ||
		unsafe.FailureClass != FailureEvidenceWeak ||
		!productionAdapterStringContains(unsafe.BlockedReasons, "unsafe_input_ref") ||
		!productionAdapterMissingContains(unsafe.MissingInputs, "host:display_safe_refs") {
		t.Fatalf("expected unsafe selection block, got %#v", unsafe)
	}
	AssertNoRawPayload(t, "unsafe adapter catalog selection", unsafe, "/tmp/raw-selection.json", "postgresql://secret")
}

func TestProductionAdapterResolutionRejectsCatalogSelectionMismatch(t *testing.T) {
	descriptor := productionAdapterTestDescriptor()
	snapshot := BuildProductionAdapterCatalogSnapshot(ProductionAdapterCatalogSnapshotInput{
		CatalogSnapshotRef: "catalog:host_adapters",
		Producer:           "host:adapter_catalog",
		Descriptors:        []ProductionAdapterDescriptor{descriptor},
	})
	selection := BuildProductionAdapterCatalogSelection(ProductionAdapterCatalogSelectionInput{
		CatalogSnapshot:      snapshot,
		SelectionRef:         "selection:metrics_adapter",
		SelectedAdapterRef:   descriptor.AdapterRef,
		SelectedSourceKind:   ReplannerSourceOperations,
		SelectedSourceRef:    "source:operations_metrics",
		SelectedCandidateRef: "strategy:operations_metric_collect",
	})
	conflictDescriptor := descriptor
	conflictDescriptor.AdapterRef = "adapter:other"
	input := BuildProductionAdapterResolutionInputFromCatalogSelection(ProductionAdapterCatalogSelectionResolutionInput{
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
	})
	input.Descriptor = conflictDescriptor
	input.RequestedAdapterRef = "adapter:other"

	resolution := BuildProductionAdapterResolution(input)
	if resolution.Status != HostActionBlocked ||
		resolution.ReadyForHostPreflight ||
		resolution.FailureClass != FailureInvalidInput ||
		!productionAdapterStringContains(resolution.BlockedReasons, "selected_adapter_ref_mismatch") ||
		!productionAdapterStringContains(resolution.BlockedReasons, "adapter_descriptor_selection_mismatch") ||
		!productionAdapterMissingContains(resolution.MissingInputs, "host:selected_adapter_ref") ||
		!productionAdapterMissingContains(resolution.MissingInputs, "host:adapter_descriptor") {
		t.Fatalf("expected catalog selection mismatch block, got %#v", resolution)
	}
}
