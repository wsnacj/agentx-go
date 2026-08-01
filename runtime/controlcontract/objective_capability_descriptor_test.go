package controlcontract

import "testing"

func TestBuildObjectiveCapabilityDescriptorProjectionReady(t *testing.T) {
	descriptor := objectiveCapabilityDescriptorTestReady()

	got := BuildObjectiveCapabilityDescriptorProjection(ObjectiveCapabilityDescriptorProjectionInput{
		Descriptor:         descriptor,
		ContributionRef:    "contribution:public_source_summary",
		StrategyVersionRef: "version:public_source_summary_v1",
		StrategyDigestRef:  "digest:public_source_summary",
		ProvenanceRefs:     []DisplaySafeRef{"provenance:objective_capability_descriptor_test"},
	})
	if got.Status != VerificationSatisfied ||
		!got.ReadyForCatalog ||
		got.FailureClass != FailureNone ||
		got.NextHostAction != "include_in_strategy_catalog" {
		t.Fatalf("unexpected ready projection = %#v", got)
	}
	if got.StrategyCandidate.ID != "strategy:public_source_summary" ||
		got.StrategyCandidate.Kind != "public_source_summary" ||
		got.StrategyCandidate.SideEffectClass != "read_only" ||
		got.StrategyCandidate.RequiresApproval {
		t.Fatalf("strategy candidate = %#v", got.StrategyCandidate)
	}
	if len(got.StrategyCandidate.CapabilityRefs) != 1 || got.StrategyCandidate.CapabilityRefs[0] != "capability:public_source_summary" {
		t.Fatalf("capability refs = %#v", got.StrategyCandidate.CapabilityRefs)
	}
	if len(got.StrategyCandidate.ExpectedEvidence) != 2 {
		t.Fatalf("expected evidence = %#v", got.StrategyCandidate.ExpectedEvidence)
	}
	if got.StrategyContribution.ContributionRef != "contribution:public_source_summary" ||
		got.StrategyContribution.StrategyVersionRef != "version:public_source_summary_v1" ||
		got.StrategyContribution.SourceKind != StrategyCatalogSourceScene ||
		got.StrategyContribution.SourceRef != "scene:public_source" ||
		got.StrategyContribution.Candidate.ID != "strategy:public_source_summary" {
		t.Fatalf("strategy contribution = %#v", got.StrategyContribution)
	}
	if !objectiveSpecTestBoundaryContains(got.Boundaries, "no_parallel_catalog") ||
		!objectiveSpecTestBoundaryContains(got.StrategyContribution.Boundaries, "objective_capability_descriptor_projected_to_strategy_contribution") {
		t.Fatalf("boundaries projection=%#v contribution=%#v", got.Boundaries, got.StrategyContribution.Boundaries)
	}

	clone := got.Clone()
	clone.Descriptor.RequiredEvidence[0].Ref = "evidence:changed"
	clone.StrategyCandidate.ExpectedEvidence[0].Ref = "evidence:changed"
	clone.StrategyContribution.Candidate.ExpectedEvidence[0].Ref = "evidence:changed"
	if got.Descriptor.RequiredEvidence[0].Ref != "evidence:source" ||
		got.StrategyCandidate.ExpectedEvidence[0].Ref != "evidence:source" ||
		got.StrategyContribution.Candidate.ExpectedEvidence[0].Ref != "evidence:source" {
		t.Fatalf("clone mutated original = %#v", got)
	}
}

func TestBuildObjectiveCapabilityDescriptorProjectionBlocksMissingRequiredFields(t *testing.T) {
	descriptor := objectiveCapabilityDescriptorTestReady()
	descriptor.InputSchemaRef = ""
	descriptor.RequiredEvidence = nil

	got := BuildObjectiveCapabilityDescriptorProjection(ObjectiveCapabilityDescriptorProjectionInput{
		Descriptor:         descriptor,
		ContributionRef:    "contribution:missing_contract",
		StrategyVersionRef: "version:missing_contract_v1",
		StrategyDigestRef:  "digest:missing_contract",
	})
	if got.Status != VerificationBlocked ||
		got.ReadyForCatalog ||
		got.FailureClass != FailureEvidenceMissing ||
		got.NextHostAction != "provide_objective_capability_contract" ||
		!objectiveSpecTestMissingContains(got.MissingInputs, "host:objective_capability_input_schema_ref") ||
		!objectiveSpecTestMissingContains(got.MissingInputs, "host:objective_capability_required_evidence") ||
		!objectiveSpecTestBoundaryContains(got.Boundaries, "objective_capability_descriptor_incomplete") {
		t.Fatalf("unexpected missing fields projection = %#v", got)
	}
}

func TestBuildObjectiveCapabilityDescriptorProjectionRequiresReadinessRefs(t *testing.T) {
	descriptor := objectiveCapabilityDescriptorTestReady()
	descriptor.CredentialRequirementRefs = nil
	descriptor.ConfigRequirementRefs = nil

	got := BuildObjectiveCapabilityDescriptorProjection(ObjectiveCapabilityDescriptorProjectionInput{
		Descriptor:         descriptor,
		ContributionRef:    "contribution:missing_readiness",
		StrategyVersionRef: "version:missing_readiness_v1",
		StrategyDigestRef:  "digest:missing_readiness",
	})
	if got.Status != VerificationBlocked ||
		got.FailureClass != FailureConfigMissing ||
		got.NextHostAction != "provide_objective_capability_readiness_refs" ||
		!objectiveSpecTestMissingContains(got.MissingInputs, "host:objective_capability_credential_requirement_ref") ||
		!objectiveSpecTestMissingContains(got.MissingInputs, "host:objective_capability_config_requirement_ref") {
		t.Fatalf("unexpected readiness projection = %#v", got)
	}
}

func TestBuildObjectiveCapabilityDescriptorProjectionBlocksSideEffectWithoutApproval(t *testing.T) {
	descriptor := objectiveCapabilityDescriptorTestReady()
	descriptor.CapabilityRef = "capability:external_write_action"
	descriptor.StrategyRef = "strategy:external_write_action"
	descriptor.SideEffectClass = ObjectiveCapabilitySideEffectExternalWrite
	descriptor.RequiresApproval = false

	got := BuildObjectiveCapabilityDescriptorProjection(ObjectiveCapabilityDescriptorProjectionInput{
		Descriptor:         descriptor,
		ContributionRef:    "contribution:external_write_action",
		StrategyVersionRef: "version:external_write_action_v1",
		StrategyDigestRef:  "digest:external_write_action",
	})
	if got.Status != VerificationBlocked ||
		got.ReadyForCatalog ||
		got.FailureClass != FailureApprovalRequired ||
		got.NextHostAction != "provide_objective_capability_approval_policy" ||
		!objectiveSpecTestMissingContains(got.MissingInputs, "host:objective_capability_approval_policy") ||
		!objectiveSpecTestBoundaryContains(got.Boundaries, "objective_capability_side_effect_requires_approval") {
		t.Fatalf("unexpected side effect projection = %#v", got)
	}
}

func TestBuildObjectiveCapabilityDescriptorProjectionRejectsUnsafeRefs(t *testing.T) {
	descriptor := objectiveCapabilityDescriptorTestReady()
	descriptor.CapabilityRef = "https://example.com/capability"

	got := BuildObjectiveCapabilityDescriptorProjection(ObjectiveCapabilityDescriptorProjectionInput{
		Descriptor:         descriptor,
		ContributionRef:    "contribution:unsafe",
		StrategyVersionRef: "version:unsafe_v1",
		StrategyDigestRef:  "digest:unsafe",
	})
	if got.Status != VerificationReviewRequired ||
		got.ReadyForCatalog ||
		got.FailureClass != FailureEvidenceWeak ||
		got.NextHostAction != "provide_display_safe_refs" ||
		!objectiveSpecTestMissingContains(got.MissingInputs, "host:display_safe_refs") ||
		!objectiveSpecTestBoundaryContains(got.Boundaries, "raw_output_not_allowed") {
		t.Fatalf("unexpected unsafe projection = %#v", got)
	}
}

func TestBuildObjectiveCapabilityDescriptorProjectionBlocksMissingContributionRefs(t *testing.T) {
	got := BuildObjectiveCapabilityDescriptorProjection(ObjectiveCapabilityDescriptorProjectionInput{
		Descriptor: objectiveCapabilityDescriptorTestReady(),
	})
	if got.Status != VerificationBlocked ||
		got.ReadyForCatalog ||
		got.FailureClass != FailureInsufficientInformation ||
		!objectiveSpecTestMissingContains(got.MissingInputs, "host:objective_capability_contribution_ref") ||
		!objectiveSpecTestMissingContains(got.MissingInputs, "host:objective_capability_strategy_version_ref") ||
		!objectiveSpecTestMissingContains(got.MissingInputs, "host:objective_capability_strategy_digest_ref") {
		t.Fatalf("unexpected missing contribution refs projection = %#v", got)
	}
}

func objectiveCapabilityDescriptorTestReady() ObjectiveCapabilityDescriptor {
	return ObjectiveCapabilityDescriptor{
		DescriptorRef:       "descriptor:public_source_summary",
		CapabilityRef:       "capability:public_source_summary",
		StrategyRef:         "strategy:public_source_summary",
		SourceKind:          StrategyCatalogSourceScene,
		SourceRef:           "scene:public_source",
		OwnerRef:            "owner:host",
		ProviderRef:         "provider:scene_public_source",
		StrategyKind:        "public_source_summary",
		ControlMode:         ControlModeObjective,
		MinIntensity:        IntensityL3ManagedObjective,
		MaxIntensity:        IntensityL3ManagedObjective,
		InputSchemaRef:      "schema:public_source.summary_request.v1",
		OutputSchemaRef:     "schema:public_source.summary_report.v1",
		EvidenceContractRef: "evidence_contract:public_source.summary.v1",
		RequiredEvidence: []EvidenceRef{
			{
				Ref:      "evidence:source",
				Kind:     "source",
				Strength: EvidenceAdequate,
				Source:   "scene:public_source",
			},
			{
				Ref:      "evidence:summary",
				Kind:     "summary",
				Strength: EvidenceAdequate,
				Source:   "scene:public_source",
			},
		},
		SideEffectClass:           ObjectiveCapabilitySideEffectReadOnly,
		CredentialRequirementRefs: []DisplaySafeRef{"readiness:public_source.no_credentials_required"},
		ConfigRequirementRefs:     []DisplaySafeRef{"config:public_source.runtime_adapter"},
		FailureClasses:            []FailureClass{FailureExternalDependencyUnavailable, FailureEvidenceMissing},
		ExampleRefs:               []DisplaySafeRef{"example:public_source.summary"},
		VerificationHintRefs:      []DisplaySafeRef{"verification:public_source.source_and_summary_required"},
		DomainHintRefs:            []DisplaySafeRef{"domain:public_source"},
		PolicyRefs:                []DisplaySafeRef{"policy:display_safe_refs_only"},
		Boundaries:                []Boundary{"metadata_only"},
	}
}
