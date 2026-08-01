package controlcontract

import "testing"

func TestAdapterMetadataRegistryProjectsHostAndProjectStrategies(t *testing.T) {
	registry := BuildAdapterMetadataRegistrySnapshot(adapterMetadataRegistryReadyInput())
	if registry.Status != HostActionReady ||
		!registry.ReadyForStrategyCatalog ||
		registry.RunnerEffect != "none" ||
		registry.PromptEffect != "none" ||
		registry.StrategyCatalog.CatalogRef != "catalog:adapter_metadata_registry" ||
		len(registry.StrategyCatalog.Entries) != 2 {
		t.Fatalf("ready registry = %#v", registry)
	}
	for _, boundary := range []Boundary{
		"adapter_metadata_registry_projection_only",
		"project_owned_strategy_metadata_allowed",
		"metadata_presence_not_capability",
		"no_adapter_invocation",
		"no_skill_write",
		"no_core_mutation",
	} {
		if !adapterMetadataBoundaryContains(registry.Boundaries, boundary) {
			t.Fatalf("missing boundary %q in %#v", boundary, registry.Boundaries)
		}
	}
	if registry.ReadyContributionCount != 2 ||
		!adapterMetadataDisplaySafeRefContains(registry.CandidateRefs, "strategy:project_schema_meaning") ||
		!adapterMetadataDisplaySafeRefContains(registry.SourceRefs, "project:database_domain_adapter") {
		t.Fatalf("registry aggregate = %#v", registry)
	}

	policy := objectiveLoopIntensityPolicy()
	preGate := strategyPlannerPreGate(t, policy, ControlModeObjective, IntensityL3ManagedObjective)
	plan := BuildStrategyPlanner(StrategyPlannerInput{
		Activation: ActivationManaged,
		Frame: ObjectiveFrame{
			ID:          "objective:adapter_metadata_registry",
			ControlMode: ControlModeObjective,
			Intensity:   IntensityL3ManagedObjective,
			RequiredEvidence: []EvidenceRef{{
				Ref:    "evidence:project_schema_meaning",
				Kind:   "schema_summary",
				Source: "project:database_domain_adapter",
			}},
		},
		Policy:                  policy,
		PreGate:                 preGate,
		Catalog:                 registry.StrategyCatalog,
		AvailableCapabilityRefs: []DisplaySafeRef{"capability:read_only_schema", "capability:local_metrics"},
	})
	if plan.Status != VerificationSatisfied ||
		plan.Selected.Candidate.ID != "strategy:project_schema_meaning" ||
		plan.Selected.SourceKind != StrategyCatalogSourceProject ||
		plan.NextHostAction != "run_strategy_final_gate" {
		t.Fatalf("plan from adapter metadata registry = %#v", plan)
	}
	if !adapterMetadataBoundaryContains(plan.Boundaries, "metadata_only_planner") ||
		!adapterMetadataBoundaryContains(plan.Boundaries, "planner_does_not_authorize_execution") {
		t.Fatalf("planner boundaries = %#v", plan.Boundaries)
	}
}

func TestAdapterMetadataRegistryBlocksMissingProvenanceAndUnsafeInput(t *testing.T) {
	input := adapterMetadataRegistryReadyInput()
	input.RegistryVersionRef = ""
	input.RegistryDigestRef = ""
	input.MaxContributionCount = 0
	input.Contributions[0].StrategyVersionRef = ""
	input.Contributions[0].StrategyDigestRef = ""
	input.Contributions[0].Candidate.ID = ""
	blocked := BuildAdapterMetadataRegistrySnapshot(input)
	if blocked.Status != HostActionBlocked ||
		blocked.ReadyForStrategyCatalog ||
		len(blocked.StrategyCatalog.Entries) != 0 ||
		!adapterMetadataMissingInputContains(blocked.MissingInputs, "host:adapter_metadata_registry_version_ref") ||
		!adapterMetadataMissingInputContains(blocked.MissingInputs, "host:adapter_metadata_registry_digest_ref") ||
		!adapterMetadataMissingInputContains(blocked.MissingInputs, "host:adapter_metadata_contribution_limit") ||
		!adapterMetadataMissingInputContains(blocked.MissingInputs, "host:adapter_metadata_strategy_version_ref") ||
		!adapterMetadataMissingInputContains(blocked.MissingInputs, "host:adapter_metadata_strategy_digest_ref") ||
		!adapterMetadataMissingInputContains(blocked.MissingInputs, "host:adapter_metadata_strategy_ref") {
		t.Fatalf("blocked registry = %#v", blocked)
	}

	unsafeInput := adapterMetadataRegistryReadyInput()
	unsafeInput.Contributions[0].RawOutputLoaded = true
	unsafe := BuildAdapterMetadataRegistrySnapshot(unsafeInput)
	if unsafe.Status != HostActionReviewRequired ||
		unsafe.ReadyForStrategyCatalog ||
		unsafe.FailureClass != FailureEvidenceWeak ||
		!adapterMetadataMissingInputContains(unsafe.MissingInputs, "host:display_safe_refs") ||
		!adapterMetadataBoundaryContains(unsafe.Boundaries, "raw_output_not_allowed") {
		t.Fatalf("unsafe registry = %#v", unsafe)
	}
}

func TestAdapterMetadataRegistryDoesNotTreatMetadataAsCapabilityAvailable(t *testing.T) {
	registry := BuildAdapterMetadataRegistrySnapshot(adapterMetadataRegistryReadyInput())
	policy := objectiveLoopIntensityPolicy()
	preGate := strategyPlannerPreGate(t, policy, ControlModeObjective, IntensityL3ManagedObjective)
	plan := BuildStrategyPlanner(StrategyPlannerInput{
		Activation: ActivationManaged,
		Frame: ObjectiveFrame{
			ID:          "objective:metadata_presence",
			ControlMode: ControlModeObjective,
			Intensity:   IntensityL3ManagedObjective,
		},
		Policy:  policy,
		PreGate: preGate,
		Catalog: registry.StrategyCatalog,
	})
	if plan.Status != VerificationBlocked ||
		len(plan.RankedCandidates) != 0 ||
		len(plan.RejectedCandidates) != 2 ||
		!strategyPlannerCandidateContains(plan.RejectedCandidates, "strategy:project_schema_meaning", FailureCapabilityMissing, "strategy_capability_not_proven_available") {
		t.Fatalf("metadata-only capabilities should not be available: %#v", plan)
	}
}

func adapterMetadataRegistryReadyInput() AdapterMetadataRegistrySnapshotInput {
	return AdapterMetadataRegistrySnapshotInput{
		RegistrySnapshotRef:  "registry:adapter_metadata",
		StrategyCatalogRef:   "catalog:adapter_metadata_registry",
		OwnerRef:             "owner:host",
		ProviderRef:          "provider:local_host",
		RegistryVersionRef:   "version:adapter_metadata_v1",
		RegistryDigestRef:    "digest:adapter_metadata_fixture",
		MaxContributionCount: 4,
		HostPolicyRefs:       []DisplaySafeRef{"policy:adapter_metadata_registry"},
		Contributions: []AdapterMetadataStrategyContribution{
			{
				ContributionRef:    "contribution:host_metrics_strategy",
				OwnerRef:           "owner:host",
				ProviderRef:        "provider:local_host",
				StrategyVersionRef: "version:host_metrics_strategy_v1",
				StrategyDigestRef:  "digest:host_metrics_strategy",
				SourceKind:         StrategyCatalogSourceHostAdapter,
				SourceRef:          "host:local_metrics_adapter",
				Candidate: StrategyCandidate{
					ID:             "strategy:host_metrics",
					Kind:           "runtime_adapter",
					ControlMode:    ControlModeObjective,
					MinIntensity:   IntensityL3ManagedObjective,
					MaxIntensity:   IntensityL3ManagedObjective,
					CapabilityRefs: []DisplaySafeRef{"capability:local_metrics"},
					ExpectedEvidence: []EvidenceRef{{
						Ref:      "evidence:host_metric",
						Kind:     "metric",
						Source:   "host:local_metrics_adapter",
						Strength: EvidenceAdequate,
					}},
					Risk:            "read_only",
					SideEffectClass: "read_only",
					Owner:           "host",
				},
				ProvenanceRefs: []DisplaySafeRef{"provenance:host_metrics_strategy"},
				PolicyRefs:     []DisplaySafeRef{"policy:read_only_host_adapter"},
			},
			{
				ContributionRef:    "contribution:project_schema_meaning",
				OwnerRef:           "owner:project",
				ProviderRef:        "provider:project_database_domain",
				StrategyVersionRef: "version:project_schema_meaning_v1",
				StrategyDigestRef:  "digest:project_schema_meaning",
				SourceKind:         StrategyCatalogSourceProject,
				SourceRef:          "project:database_domain_adapter",
				Candidate: StrategyCandidate{
					ID:             "strategy:project_schema_meaning",
					Kind:           "runtime_adapter",
					ControlMode:    ControlModeObjective,
					MinIntensity:   IntensityL3ManagedObjective,
					MaxIntensity:   IntensityL3ManagedObjective,
					CapabilityRefs: []DisplaySafeRef{"capability:read_only_schema"},
					ExpectedEvidence: []EvidenceRef{{
						Ref:      "evidence:project_schema_meaning",
						Kind:     "schema_summary",
						Source:   "project:database_domain_adapter",
						Strength: EvidenceStrong,
					}},
					Risk:            "read_only",
					SideEffectClass: "read_only",
					Owner:           "project",
				},
				ProvenanceRefs: []DisplaySafeRef{"provenance:project_schema_meaning"},
				PolicyRefs:     []DisplaySafeRef{"policy:project_read_only_schema"},
			},
		},
	}
}

func adapterMetadataBoundaryContains(values []Boundary, want Boundary) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func adapterMetadataMissingInputContains(values []MissingInput, want MissingInput) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func adapterMetadataDisplaySafeRefContains(values []DisplaySafeRef, want DisplaySafeRef) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
