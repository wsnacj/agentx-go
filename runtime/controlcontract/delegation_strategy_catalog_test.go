package controlcontract

import "testing"

func TestDelegationStrategyCatalogCanFeedStrategyPlanner(t *testing.T) {
	policy := objectiveLoopIntensityPolicy()
	policy.AllowedControlModesByIntensity[IntensityL4DurableLongRun] = append(
		policy.AllowedControlModesByIntensity[IntensityL4DurableLongRun],
		ControlModeDelegated,
	)
	preGate := BuildExecutionIntensityPreGate(IntensityGateInput{
		Activation:           ActivationManaged,
		Policy:               policy,
		RequestedControlMode: ControlModeDelegated,
		RequestedIntensity:   IntensityL4DurableLongRun,
		UserConfirmed:        true,
		HostApproved:         true,
		ApprovalRefs:         []DisplaySafeRef{"approval:delegation_strategy"},
		Budget:               ObjectiveBudgetSnapshot{BudgetRef: "budget:delegation_strategy", Limit: 2},
	})
	if !preGate.Allowed {
		t.Fatalf("pre gate should allow explicit L4 delegation: %#v", preGate)
	}
	request := BuildDelegationRequestProjection(delegationReadyRequestInput(IntensityL4DurableLongRun))
	catalog := BuildDelegationStrategyCatalogSnapshot(DelegationStrategyCatalogSnapshotInput{
		CatalogRef: "catalog:delegation_strategy",
		Entries: []DelegationStrategyCatalogEntryInput{{
			RequestRef:     "delegation_request:fixture",
			Request:        request,
			CandidateRef:   "strategy:delegation_worker_fixture",
			CapabilityRefs: []DisplaySafeRef{"capability:delegation_worker_runtime"},
		}},
	})
	plan := BuildStrategyPlanner(StrategyPlannerInput{
		Activation: ActivationManaged,
		Frame: ObjectiveFrame{
			ID:          "objective:delegation_strategy",
			ControlMode: ControlModeDelegated,
			Intensity:   IntensityL4DurableLongRun,
			RequiredEvidence: []EvidenceRef{{
				Kind: "delegation_worker_result",
			}},
		},
		Policy:                  policy,
		PreGate:                 preGate,
		Catalog:                 catalog,
		AvailableCapabilityRefs: []DisplaySafeRef{"capability:delegation_worker_runtime"},
	})
	if plan.Status != VerificationSatisfied ||
		plan.Selected.SourceKind != StrategyCatalogSourceDelegation ||
		plan.Selected.SourceRef != "delegation_request:fixture" ||
		plan.Selected.Candidate.ID != "strategy:delegation_worker_fixture" ||
		plan.Selected.Candidate.ControlMode != ControlModeDelegated ||
		plan.NextHostAction != "run_strategy_final_gate" {
		t.Fatalf("unexpected delegation strategy plan: %#v", plan)
	}
	for _, boundary := range []Boundary{
		"delegation_strategy_metadata_only",
		"delegation_strategy_requires_host_worker_runtime",
		"controlplane_does_not_dispatch_worker",
		"worker_result_requires_verification",
	} {
		if !delegationBoundaryContains(plan.Selected.Boundaries, boundary) {
			t.Fatalf("missing boundary %q in %#v", boundary, plan.Selected.Boundaries)
		}
	}
}

func TestDelegationStrategyCatalogAllowsMetadataOnlyCandidate(t *testing.T) {
	entry := BuildDelegationStrategyCatalogEntry(DelegationStrategyCatalogEntryInput{
		SourceRef:      "delegation_strategy:metadata_only",
		CandidateRef:   "strategy:delegation_metadata_only",
		CapabilityRefs: []DisplaySafeRef{"capability:delegation_worker_runtime"},
	})
	if entry.Status != VerificationSatisfied ||
		entry.SourceKind != StrategyCatalogSourceDelegation ||
		entry.Candidate.ControlMode != ControlModeDelegated ||
		entry.Candidate.MinIntensity != IntensityL4DurableLongRun ||
		entry.Candidate.RequiresApproval != true ||
		len(entry.MissingInputs) != 0 ||
		!delegationBoundaryContains(entry.Boundaries, "delegation_strategy_requires_explicit_request_before_dispatch") {
		t.Fatalf("metadata-only delegation entry should remain selectable metadata: %#v", entry)
	}
}

func TestDelegationStrategyCatalogBlocksUnreadyRequest(t *testing.T) {
	input := delegationReadyRequestInput(IntensityL4DurableLongRun)
	input.HostApproved = false
	request := BuildDelegationRequestProjection(input)
	entry := BuildDelegationStrategyCatalogEntry(DelegationStrategyCatalogEntryInput{
		RequestRef:   "delegation_request:blocked",
		Request:      request,
		CandidateRef: "strategy:delegation_worker_blocked",
	})
	if entry.Status != VerificationBlocked ||
		entry.FailureClass != FailureApprovalRequired ||
		!delegationMissingInputContains(entry.MissingInputs, "host:delegation_approval") ||
		!delegationBoundaryContains(entry.Boundaries, "delegation_request_not_ready_for_strategy_catalog") {
		t.Fatalf("expected blocked delegation catalog entry, got %#v", entry)
	}
}
