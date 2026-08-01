package controlcontract

import "testing"

func TestRepeatedSuccessMemoryProposalBuildsProposalOnlyItems(t *testing.T) {
	registry := BuildAdapterMetadataRegistrySnapshot(adapterMetadataRegistryReadyInput())
	proposals := BuildRepeatedSuccessMemoryProposal(repeatedSuccessMemoryProposalReadyInput(registry.StrategyCatalog))
	if proposals.Status != HostActionReady ||
		!proposals.ReadyForHostReview ||
		!proposals.ProposalOnly ||
		proposals.RunnerEffect != "none" ||
		proposals.PromptEffect != "none" ||
		len(proposals.Proposals) != 3 ||
		proposals.SkillWriteExecuted ||
		proposals.WorkflowWriteExecuted ||
		proposals.TemplateWriteExecuted ||
		proposals.InstallExecuted ||
		proposals.RuntimeReloadExecuted ||
		proposals.CoreMutationExecuted {
		t.Fatalf("ready memory proposal = %#v", proposals)
	}
	for _, boundary := range []Boundary{
		"proposal_only",
		"memory_proposal_projection_only",
		"host_must_review_memory_proposal",
		"no_skill_write",
		"no_workflow_write",
		"no_template_write",
		"no_install_or_reload",
		"no_core_mutation",
		"no_runner_dispatch",
	} {
		if !adapterMetadataBoundaryContains(proposals.Boundaries, boundary) {
			t.Fatalf("missing boundary %q in %#v", boundary, proposals.Boundaries)
		}
	}
	if !adapterMetadataDisplaySafeRefContains(proposals.RepeatedStrategyRefs, "strategy:project_schema_meaning") {
		t.Fatalf("repeated strategies = %#v", proposals.RepeatedStrategyRefs)
	}
	for _, proposal := range proposals.Proposals {
		if proposal.SourceStrategyRef != "strategy:project_schema_meaning" ||
			proposal.SupportCount != 2 ||
			len(proposal.SupportAttemptRefs) != 2 ||
			len(proposal.EvidenceRefs) == 0 ||
			!adapterMetadataBoundaryContains(proposal.Boundaries, "proposal_only") ||
			!adapterMetadataBoundaryContains(proposal.Boundaries, "no_skill_write") {
			t.Fatalf("proposal item = %#v", proposal)
		}
	}
}

func TestRepeatedSuccessMemoryProposalBlocksWithoutRepeatedEvidence(t *testing.T) {
	registry := BuildAdapterMetadataRegistrySnapshot(adapterMetadataRegistryReadyInput())
	input := repeatedSuccessMemoryProposalReadyInput(registry.StrategyCatalog)
	input.Attempts = input.Attempts[:1]
	proposals := BuildRepeatedSuccessMemoryProposal(input)
	if proposals.Status != HostActionBlocked ||
		proposals.ReadyForHostReview ||
		len(proposals.Proposals) != 0 ||
		!adapterMetadataMissingInputContains(proposals.MissingInputs, "host:repeated_success_attempts") ||
		!adapterMetadataStringContains(proposals.BlockedReasons, "repeated_success_path_missing") ||
		proposals.SkillWriteExecuted ||
		proposals.WorkflowWriteExecuted ||
		proposals.CoreMutationExecuted {
		t.Fatalf("single success proposal = %#v", proposals)
	}
}

func TestRepeatedSuccessMemoryProposalRequiresPolicyAndBoundedCount(t *testing.T) {
	registry := BuildAdapterMetadataRegistrySnapshot(adapterMetadataRegistryReadyInput())
	input := repeatedSuccessMemoryProposalReadyInput(registry.StrategyCatalog)
	input.MemoryPolicyRef = ""
	input.MinSuccessfulAttempts = 0
	input.MaxProposalCount = 0
	blocked := BuildRepeatedSuccessMemoryProposal(input)
	if blocked.Status != HostActionBlocked ||
		blocked.ReadyForHostReview ||
		!blocked.ProposalOnly ||
		!adapterMetadataMissingInputContains(blocked.MissingInputs, "host:memory_policy_ref") ||
		!adapterMetadataMissingInputContains(blocked.MissingInputs, "host:memory_min_successful_attempts") ||
		!adapterMetadataMissingInputContains(blocked.MissingInputs, "host:memory_max_proposal_count") {
		t.Fatalf("policy blocked proposal = %#v", blocked)
	}

	unsafeInput := repeatedSuccessMemoryProposalReadyInput(registry.StrategyCatalog)
	unsafeInput.Attempts[0].RawOutputLoaded = true
	unsafe := BuildRepeatedSuccessMemoryProposal(unsafeInput)
	if unsafe.Status != HostActionReviewRequired ||
		unsafe.ReadyForHostReview ||
		unsafe.FailureClass != FailureEvidenceWeak ||
		!adapterMetadataMissingInputContains(unsafe.MissingInputs, "host:display_safe_refs") ||
		!adapterMetadataBoundaryContains(unsafe.Boundaries, "raw_output_not_allowed") {
		t.Fatalf("unsafe memory proposal = %#v", unsafe)
	}
}

func repeatedSuccessMemoryProposalReadyInput(catalog StrategyCatalogSnapshot) RepeatedSuccessMemoryProposalInput {
	return RepeatedSuccessMemoryProposalInput{
		Activation:            ActivationManaged,
		Frame:                 strategyMemoryProposalFrame(),
		ProposalSetRef:        "memory_proposal:set:project_schema_meaning",
		ProposalOwnerRef:      "owner:host",
		MemoryPolicyRef:       "policy:memory_proposal_review_required",
		StrategyCatalog:       catalog,
		ProposalKinds:         []MemoryProposalKind{MemoryProposalKindSkill, MemoryProposalKindWorkflow, MemoryProposalKindTemplate},
		MinSuccessfulAttempts: 2,
		MaxProposalCount:      3,
		ProvenanceRefs:        []DisplaySafeRef{"provenance:objective_ledger_review"},
		EvidenceRefs: []EvidenceRef{{
			Ref:      "evidence:memory_policy_review",
			Kind:     "memory_policy",
			Source:   "host:memory_policy",
			Strength: EvidenceAdequate,
		}},
		Attempts: []AttemptSummary{
			strategyMemoryProposalAttempt("attempt:project_schema_meaning_1", 1),
			strategyMemoryProposalAttempt("attempt:project_schema_meaning_2", 2),
			{
				Ref:              "attempt:failed_path",
				ObjectiveID:      "objective:memory_proposal",
				StrategyID:       "strategy:project_schema_meaning",
				Index:            3,
				ControlMode:      ControlModeObjective,
				Intensity:        IntensityL3ManagedObjective,
				Status:           VerificationBlocked,
				ObservationCount: 1,
				FailureClass:     FailureVerificationFailed,
			},
		},
	}
}

func strategyMemoryProposalFrame() ObjectiveFrame {
	return ObjectiveFrame{
		ID:             "objective:memory_proposal",
		UserGoalDigest: "project schema meaning objective",
		ControlMode:    ControlModeObjective,
		Intensity:      IntensityL3ManagedObjective,
		RequiredEvidence: []EvidenceRef{{
			Ref:      "evidence:project_schema_meaning",
			Kind:     "schema_summary",
			Source:   "project:database_domain_adapter",
			Strength: EvidenceStrong,
		}},
	}
}

func strategyMemoryProposalAttempt(ref AttemptRef, index int) AttemptSummary {
	return AttemptSummary{
		Ref:              ref,
		ObjectiveID:      "objective:memory_proposal",
		StrategyID:       "strategy:project_schema_meaning",
		Index:            index,
		ControlMode:      ControlModeObjective,
		Intensity:        IntensityL3ManagedObjective,
		Status:           VerificationSatisfied,
		ObservationCount: 2,
		EvidenceRefs: []EvidenceRef{{
			Ref:      "evidence:project_schema_meaning",
			Kind:     "schema_summary",
			Source:   "project:database_domain_adapter",
			Strength: EvidenceStrong,
		}},
	}
}

func adapterMetadataStringContains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
