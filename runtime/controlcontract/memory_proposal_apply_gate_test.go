package controlcontract

import "testing"

func TestMemoryProposalReviewPacketAndApplyInvocationReady(t *testing.T) {
	review := memoryProposalReviewPacketReadyFixture(t)
	if review.Status != HostActionReady ||
		!review.ReadyForHostMemoryApply ||
		!review.HostMemoryApplyAuthorized ||
		!review.HostMayApplyMemoryMutation ||
		!review.ProposalOnly ||
		len(review.SelectedProposalRefs) != 3 ||
		len(review.ExpectedMemoryArtifactRefs) != 3 ||
		review.SkillWriteByCore ||
		review.WorkflowWriteByCore ||
		review.TemplateWriteByCore ||
		review.InstallExecutedByCore ||
		review.RuntimeReloadByCore ||
		review.CoreMutationExecuted ||
		review.RunnerDispatched {
		t.Fatalf("ready memory proposal review = %#v", review)
	}
	for _, boundary := range []Boundary{
		"memory_proposal_review_packet",
		"proposal_not_apply",
		"host_owned_memory_apply",
		"no_skill_write_by_core",
		"no_workflow_write_by_core",
		"no_template_write_by_core",
		"ready_for_host_memory_apply",
	} {
		if !adapterMetadataBoundaryContains(review.Boundaries, boundary) {
			t.Fatalf("missing review boundary %q in %#v", boundary, review.Boundaries)
		}
	}

	readiness := memoryProposalApplyReadinessReadyFixture(t, review)
	if readiness.Status != HostActionReady ||
		!readiness.ReadyForHostMemoryAdapterInvocation ||
		!readiness.HostMemoryAdapterInvocationAuthorized ||
		!readiness.HostMayInvokeMemoryAdapter ||
		readiness.SkillWriteByCore ||
		readiness.WorkflowWriteByCore ||
		readiness.TemplateWriteByCore ||
		readiness.RuntimeAdapterExecuted ||
		readiness.StoreMutationExecuted {
		t.Fatalf("ready memory proposal apply adapter = %#v", readiness)
	}
	for _, boundary := range []Boundary{
		"host_owned_memory_apply_adapter_gate",
		"memory_apply_independent_gate_required",
		"ready_for_host_owned_memory_apply_adapter_invocation",
		"no_skill_write_by_core",
		"no_store_mutation_by_core",
	} {
		if !adapterMetadataBoundaryContains(readiness.Boundaries, boundary) {
			t.Fatalf("missing readiness boundary %q in %#v", boundary, readiness.Boundaries)
		}
	}

	invocation := BuildHostOwnedMemoryProposalApplyInvocation(HostOwnedMemoryProposalApplyInvocationInput{
		Readiness:                      readiness,
		InvocationReportRef:            "invocation_report:memory_apply",
		ObservedInvocationRef:          readiness.InvocationRef,
		HostMemoryAdapterRunRef:        "memory_adapter_run:memory_apply",
		MemoryApplyResultRef:           readiness.ExpectedMemoryApplyResultRef,
		MemoryReadbackRef:              readiness.ExpectedReadbackRef,
		AppliedMemoryArtifactRefs:      readiness.ExpectedMemoryArtifactRefs,
		ReadbackMemoryArtifactRefs:     readiness.ExpectedMemoryArtifactRefs,
		HostAdapterInvocationReported:  true,
		HostAdapterInvocationCompleted: true,
		MemoryEvidenceRefs:             []DisplaySafeRef{"evidence:memory_apply_invocation"},
	})
	if invocation.Status != HostActionRecorded ||
		!invocation.ReadyForMemoryApplyReadback ||
		invocation.ReadyForFailureReview ||
		!invocation.HostAdapterInvocationReported ||
		!invocation.HostAdapterInvocationCompleted ||
		invocation.SkillWriteByCore ||
		invocation.WorkflowWriteByCore ||
		invocation.TemplateWriteByCore ||
		invocation.RuntimeAdapterExecuted ||
		invocation.StoreMutationExecuted {
		t.Fatalf("recorded memory proposal apply invocation = %#v", invocation)
	}
	for _, boundary := range []Boundary{
		"host_owned_memory_apply_adapter_invocation_report",
		"host_adapter_memory_mutation_reported_only",
		"memory_apply_readback_bound",
		"no_skill_write_by_core",
		"no_core_mutation",
	} {
		if !adapterMetadataBoundaryContains(invocation.Boundaries, boundary) {
			t.Fatalf("missing invocation boundary %q in %#v", boundary, invocation.Boundaries)
		}
	}
}

func TestMemoryProposalReviewPacketBlocksUnreviewedOrInvalidSelection(t *testing.T) {
	registry := BuildAdapterMetadataRegistrySnapshot(adapterMetadataRegistryReadyInput())
	proposalSet := BuildRepeatedSuccessMemoryProposal(repeatedSuccessMemoryProposalReadyInput(registry.StrategyCatalog))
	input := memoryProposalReviewPacketReadyInput(proposalSet)
	input.HostReviewCompleted = false
	unreviewed := BuildMemoryProposalReviewPacket(input)
	if unreviewed.Status != HostActionBlocked ||
		unreviewed.ReadyForHostMemoryApply ||
		!adapterMetadataMissingInputContains(unreviewed.MissingInputs, "host:memory_review_completed") ||
		unreviewed.SkillWriteByCore ||
		unreviewed.CoreMutationExecuted {
		t.Fatalf("unreviewed packet = %#v", unreviewed)
	}

	input = memoryProposalReviewPacketReadyInput(proposalSet)
	input.SelectedProposalRefs = []DisplaySafeRef{"proposal:skill:not_in_set"}
	invalid := BuildMemoryProposalReviewPacket(input)
	if invalid.Status != HostActionBlocked ||
		invalid.ReadyForHostMemoryApply ||
		!adapterMetadataStringContains(invalid.BlockedReasons, "selected_proposal_ref_not_in_set") ||
		!adapterMetadataMissingInputContains(invalid.MissingInputs, "host:selected_memory_proposal_refs") {
		t.Fatalf("invalid selected proposal = %#v", invalid)
	}
}

func TestMemoryProposalApplyReadinessRequiresMemoryGateAndMatchingBindings(t *testing.T) {
	review := memoryProposalReviewPacketReadyFixture(t)
	wrongGate := memoryProposalApplyGateFixture()
	wrongGate.Kind = ProductionAdapterEffectGateInstallerApply
	blocked := BuildHostOwnedMemoryProposalApplyReadiness(HostOwnedMemoryProposalApplyReadinessInput{
		ReviewPacket:         review,
		MemoryApplyGate:      wrongGate,
		AdapterRef:           review.MemoryApplyAdapterRef,
		AdapterVersionRef:    "adapter_version:memory_apply_v1",
		AdapterCapabilityRef: "capability:memory_apply",
		AdapterContractRef:   "contract:memory_apply_adapter",
		HostConfirmationRef:  firstDisplaySafeRef(review.ApprovalRefs...),
		AdapterDryRunRef:     "dry_run:memory_apply_adapter",
		InvocationRef:        "invocation:memory_apply_adapter",
		ResultBindingRef:     review.ExpectedMemoryApplyResultRef,
		ReadbackBindingRef:   review.ExpectedReadbackRef,
		IdempotencyRef:       review.IdempotencyRef,
		RollbackRef:          review.RollbackPathRef,
		FailureBindingRef:    "failure:memory_apply_adapter",
		CompensationRef:      "compensation:memory_apply_adapter",
	})
	if blocked.Status == HostActionReady ||
		blocked.ReadyForHostMemoryAdapterInvocation ||
		!adapterMetadataStringContains(blocked.BlockedReasons, "memory_apply_independent_gate_not_ready") {
		t.Fatalf("wrong memory apply gate should block: %#v", blocked)
	}

	readiness := memoryProposalApplyReadinessReadyFixture(t, review)
	input := HostOwnedMemoryProposalApplyReadinessInput{
		ReviewPacket:         review,
		MemoryApplyGate:      memoryProposalApplyGateFixture(),
		AdapterRef:           "adapter:other",
		AdapterVersionRef:    readiness.AdapterVersionRef,
		AdapterCapabilityRef: readiness.AdapterCapabilityRef,
		AdapterContractRef:   readiness.AdapterContractRef,
		HostConfirmationRef:  readiness.HostConfirmationRef,
		AdapterDryRunRef:     readiness.AdapterDryRunRef,
		InvocationRef:        readiness.InvocationRef,
		ResultBindingRef:     readiness.ResultBindingRef,
		ReadbackBindingRef:   readiness.ReadbackBindingRef,
		IdempotencyRef:       readiness.IdempotencyRef,
		RollbackRef:          readiness.RollbackRef,
		FailureBindingRef:    readiness.FailureBindingRef,
		CompensationRef:      readiness.CompensationRef,
	}
	mismatch := BuildHostOwnedMemoryProposalApplyReadiness(input)
	if mismatch.Status == HostActionReady ||
		!adapterMetadataStringContains(mismatch.BlockedReasons, "adapter_ref_mismatch") ||
		!adapterMetadataMissingInputContains(mismatch.MissingInputs, "host:memory_apply_adapter_ref") {
		t.Fatalf("adapter mismatch should block: %#v", mismatch)
	}
}

func TestMemoryProposalApplyInvocationBlocksMismatchFailureAndUnsafeRefs(t *testing.T) {
	review := memoryProposalReviewPacketReadyFixture(t)
	readiness := memoryProposalApplyReadinessReadyFixture(t, review)
	mismatch := BuildHostOwnedMemoryProposalApplyInvocation(HostOwnedMemoryProposalApplyInvocationInput{
		Readiness:                      readiness,
		InvocationReportRef:            "invocation_report:memory_apply",
		ObservedInvocationRef:          "invocation:other",
		HostMemoryAdapterRunRef:        "memory_adapter_run:memory_apply",
		MemoryApplyResultRef:           readiness.ExpectedMemoryApplyResultRef,
		MemoryReadbackRef:              readiness.ExpectedReadbackRef,
		AppliedMemoryArtifactRefs:      readiness.ExpectedMemoryArtifactRefs,
		ReadbackMemoryArtifactRefs:     readiness.ExpectedMemoryArtifactRefs,
		HostAdapterInvocationReported:  true,
		HostAdapterInvocationCompleted: true,
	})
	if mismatch.Status == HostActionRecorded ||
		!adapterMetadataStringContains(mismatch.BlockedReasons, "observed_invocation_ref_mismatch") ||
		!adapterMetadataMissingInputContains(mismatch.MissingInputs, "host:memory_apply_adapter_invocation_ref") {
		t.Fatalf("invocation mismatch should block: %#v", mismatch)
	}

	failed := BuildHostOwnedMemoryProposalApplyInvocation(HostOwnedMemoryProposalApplyInvocationInput{
		Readiness:                     readiness,
		InvocationReportRef:           "invocation_report:memory_apply_failure",
		ObservedInvocationRef:         readiness.InvocationRef,
		HostMemoryAdapterRunRef:       "memory_adapter_run:memory_apply_failure",
		FailureRef:                    "failure:memory_apply",
		CompensationRef:               "compensation:memory_apply",
		HostAdapterInvocationReported: true,
		HostAdapterInvocationFailed:   true,
	})
	if failed.Status != HostActionRecorded ||
		!failed.ReadyForFailureReview ||
		failed.ReadyForMemoryApplyReadback ||
		failed.SkillWriteByCore ||
		failed.CoreMutationExecuted {
		t.Fatalf("failed memory apply invocation = %#v", failed)
	}

	unsafe := BuildHostOwnedMemoryProposalApplyInvocation(HostOwnedMemoryProposalApplyInvocationInput{
		Readiness:             readiness,
		InvocationReportRef:   "/tmp/raw-memory-apply-report",
		ObservedInvocationRef: readiness.InvocationRef,
	})
	if unsafe.Status != HostActionReviewRequired ||
		!adapterMetadataMissingInputContains(unsafe.MissingInputs, "host:display_safe_refs") ||
		!adapterMetadataBoundaryContains(unsafe.Boundaries, "raw_output_not_allowed") {
		t.Fatalf("unsafe memory apply invocation = %#v", unsafe)
	}
}

func memoryProposalReviewPacketReadyFixture(t *testing.T) MemoryProposalReviewPacket {
	t.Helper()
	registry := BuildAdapterMetadataRegistrySnapshot(adapterMetadataRegistryReadyInput())
	proposalSet := BuildRepeatedSuccessMemoryProposal(repeatedSuccessMemoryProposalReadyInput(registry.StrategyCatalog))
	return BuildMemoryProposalReviewPacket(memoryProposalReviewPacketReadyInput(proposalSet))
}

func memoryProposalReviewPacketReadyInput(proposalSet RepeatedSuccessMemoryProposalSet) MemoryProposalReviewPacketInput {
	selected := []DisplaySafeRef{}
	for _, proposal := range proposalSet.Proposals {
		selected = append(selected, proposal.ProposalRef)
	}
	return MemoryProposalReviewPacketInput{
		ProposalSet:           proposalSet,
		ReviewPacketRef:       "memory_review_packet:project_schema_meaning",
		ReviewerRef:           "reviewer:host",
		HostReviewRef:         "review:memory_project_schema_meaning",
		MemoryApplyPolicyRef:  "policy:memory_apply_review_required",
		MemoryApplyAdapterRef: "adapter:memory_apply_project_schema_meaning",
		SelectedProposalRefs:  selected,
		ExpectedMemoryArtifactRefs: []DisplaySafeRef{
			"memory_artifact:project_schema_skill",
			"memory_artifact:project_schema_workflow",
			"memory_artifact:project_schema_template",
		},
		ExpectedMemoryApplyResultRef: "memory_apply_result:project_schema_meaning",
		ExpectedReadbackRef:          "memory_apply_readback:project_schema_meaning",
		IdempotencyRef:               "idempotency:memory_apply_project_schema_meaning",
		RollbackPathRef:              "rollback:memory_apply_project_schema_meaning",
		ApprovalRefs:                 []DisplaySafeRef{"approval:memory_apply_project_schema_meaning"},
		EvidenceRefs: []EvidenceRef{{
			Ref:      "evidence:memory_apply_review",
			Kind:     "memory_review",
			Source:   "host:memory_reviewer",
			Strength: EvidenceStrong,
		}},
		HostReviewCompleted: true,
		HostReviewApproved:  true,
	}
}

func memoryProposalApplyReadinessReadyFixture(t *testing.T, review MemoryProposalReviewPacket) HostOwnedMemoryProposalApplyReadiness {
	t.Helper()
	return BuildHostOwnedMemoryProposalApplyReadiness(HostOwnedMemoryProposalApplyReadinessInput{
		ReviewPacket:         review,
		MemoryApplyGate:      memoryProposalApplyGateFixture(),
		AdapterRef:           review.MemoryApplyAdapterRef,
		AdapterVersionRef:    "adapter_version:memory_apply_v1",
		AdapterCapabilityRef: "capability:memory_apply",
		AdapterContractRef:   "contract:memory_apply_adapter",
		HostConfirmationRef:  firstDisplaySafeRef(review.ApprovalRefs...),
		AdapterDryRunRef:     "dry_run:memory_apply_adapter",
		InvocationRef:        "invocation:memory_apply_adapter",
		ResultBindingRef:     review.ExpectedMemoryApplyResultRef,
		ReadbackBindingRef:   review.ExpectedReadbackRef,
		IdempotencyRef:       review.IdempotencyRef,
		RollbackRef:          review.RollbackPathRef,
		FailureBindingRef:    "failure:memory_apply_adapter",
		CompensationRef:      "compensation:memory_apply_adapter",
		EvidenceRefs: []EvidenceRef{{
			Ref:      "evidence:memory_apply_adapter_binding",
			Kind:     "memory_apply_adapter_binding",
			Source:   "host:memory_apply_adapter",
			Strength: EvidenceStrong,
		}},
	})
}

func memoryProposalApplyGateFixture() ProductionAdapterIndependentEffectGate {
	return BuildProductionAdapterIndependentEffectGate(ProductionAdapterIndependentEffectGateSpec{
		Kind:                  ProductionAdapterEffectGateMemoryApply,
		GateRef:               "gate:memory_apply",
		AdapterRef:            "adapter:memory_apply_project_schema_meaning",
		ContractRef:           "contract:memory_apply",
		PolicyRef:             "policy:memory_apply",
		ApprovalRef:           "approval:memory_apply_project_schema_meaning",
		BudgetRef:             "budget:memory_apply",
		IdempotencyRef:        "idempotency:memory_apply_project_schema_meaning",
		ReadbackRef:           "memory_apply_readback:project_schema_meaning",
		EvalRef:               "eval:memory_apply",
		FailureReviewRef:      "failure_review:memory_apply",
		CompensationReviewRef: "compensation_review:memory_apply",
		EvidenceRefs:          []DisplaySafeRef{"evidence:memory_apply_gate"},
	})
}
