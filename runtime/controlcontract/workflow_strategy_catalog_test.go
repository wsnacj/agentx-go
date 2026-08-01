package controlcontract

import "testing"

func TestWorkflowStrategyCatalogEntryBuildsMetadataCandidate(t *testing.T) {
	entry := BuildWorkflowStrategyCatalogEntry(WorkflowStrategyCatalogEntryInput{
		WorkflowRef:    "workflow:runtime_objective",
		CandidateRef:   "strategy:workflow_runtime",
		CapabilityRefs: []DisplaySafeRef{"capability:workflow_runtime_backend"},
	})
	if entry.SourceKind != StrategyCatalogSourceWorkflow ||
		entry.SourceRef != "workflow:runtime_objective" ||
		entry.Status != VerificationSatisfied ||
		entry.Candidate.ID != "strategy:workflow_runtime" ||
		entry.Candidate.ControlMode != ControlModeWorkflow ||
		entry.Candidate.MinIntensity != IntensityL3ManagedObjective ||
		entry.Candidate.MaxIntensity != IntensityL3ManagedObjective ||
		!entry.Candidate.RequiresApproval ||
		entry.Candidate.Owner != "host" ||
		len(entry.Candidate.ExpectedEvidence) != 1 ||
		entry.Candidate.ExpectedEvidence[0].Kind != "workflow_node_result" {
		t.Fatalf("unexpected workflow strategy catalog entry: %#v", entry)
	}
	for _, want := range []Boundary{
		"workflow_strategy_catalog_entry",
		"workflow_strategy_metadata_only",
		"workflow_strategy_requires_host_runtime_backend",
		"controlplane_does_not_execute_workflow",
		"no_workflow_dispatch",
		"no_runner_dispatch",
	} {
		if !intensityGateBoundaryContains(entry.Boundaries, want) {
			t.Fatalf("expected boundary %q, got %#v", want, entry.Boundaries)
		}
	}
}

func TestWorkflowStrategyCatalogCanFeedStrategyPlanner(t *testing.T) {
	policy := workflowStrategyPlannerPolicy()
	preGate := strategyPlannerPreGate(t, policy, ControlModeWorkflow, IntensityL3ManagedObjective)
	catalog := BuildWorkflowStrategyCatalogSnapshot(WorkflowStrategyCatalogSnapshotInput{
		CatalogRef: "catalog:workflow_strategy",
		Entries: []WorkflowStrategyCatalogEntryInput{{
			WorkflowSpecRef: "workflow_spec:runtime_objective",
			CandidateRef:    "strategy:workflow_runtime",
			CapabilityRefs:  []DisplaySafeRef{"capability:workflow_runtime_backend"},
		}},
	})
	plan := BuildStrategyPlanner(StrategyPlannerInput{
		Activation: ActivationManaged,
		Frame: ObjectiveFrame{
			ID:          "objective:workflow_runtime",
			ControlMode: ControlModeWorkflow,
			Intensity:   IntensityL3ManagedObjective,
			RequiredEvidence: []EvidenceRef{{
				Kind: "workflow_node_result",
			}},
		},
		Policy:                  policy,
		PreGate:                 preGate,
		Catalog:                 catalog,
		AvailableCapabilityRefs: []DisplaySafeRef{"capability:workflow_runtime_backend"},
	})
	if plan.Status != VerificationSatisfied ||
		plan.Selected.SourceKind != StrategyCatalogSourceWorkflow ||
		plan.Selected.SourceRef != "workflow_spec:runtime_objective" ||
		plan.Selected.Candidate.ID != "strategy:workflow_runtime" ||
		plan.Selected.Candidate.ControlMode != ControlModeWorkflow ||
		plan.NextHostAction != "run_strategy_final_gate" {
		t.Fatalf("unexpected workflow strategy plan: %#v", plan)
	}
	if !intensityGateBoundaryContains(plan.Selected.Boundaries, "strategy_requires_approval") ||
		!intensityGateBoundaryContains(plan.Selected.Boundaries, "controlplane_does_not_execute_workflow") {
		t.Fatalf("workflow selected boundaries = %#v", plan.Selected.Boundaries)
	}
}

func TestWorkflowStrategyCatalogBlocksUnsafeRefs(t *testing.T) {
	entry := BuildWorkflowStrategyCatalogEntry(WorkflowStrategyCatalogEntryInput{
		WorkflowRef:  "/tmp/raw-workflow.json",
		CandidateRef: "strategy:workflow_runtime",
	})
	if entry.Status != VerificationBlocked ||
		entry.FailureClass != FailureEvidenceWeak ||
		!entry.RawOutputLoaded ||
		!intensityGateMissingInputContains(entry.MissingInputs, "host:display_safe_refs") ||
		!intensityGateBoundaryContains(entry.Boundaries, "raw_output_not_allowed") {
		t.Fatalf("expected unsafe workflow strategy block, got %#v", entry)
	}
}

func TestStrategyPlannerWorkflowRetryAndReplanPath(t *testing.T) {
	policy := workflowStrategyPlannerPolicy()
	preGate := strategyPlannerPreGate(t, policy, ControlModeWorkflow, IntensityL3ManagedObjective)
	catalog := BuildWorkflowStrategyCatalogSnapshot(WorkflowStrategyCatalogSnapshotInput{
		CatalogRef: "catalog:workflow_strategy",
		Entries: []WorkflowStrategyCatalogEntryInput{
			{
				WorkflowRef:    "workflow:primary",
				CandidateRef:   "strategy:workflow_primary",
				CapabilityRefs: []DisplaySafeRef{"capability:workflow_runtime_backend"},
			},
			{
				WorkflowRef:    "workflow:alternate",
				CandidateRef:   "strategy:workflow_alternate",
				CapabilityRefs: []DisplaySafeRef{"capability:workflow_runtime_backend"},
			},
		},
	})
	frame := ObjectiveFrame{
		ID:          "objective:workflow_retry",
		ControlMode: ControlModeWorkflow,
		Intensity:   IntensityL3ManagedObjective,
	}
	partialRetry := BuildStrategyPlanner(StrategyPlannerInput{
		Activation:         ActivationManaged,
		Frame:              frame,
		Policy:             policy,
		PreGate:            preGate,
		Catalog:            catalog,
		CurrentStrategyRef: "strategy:workflow_primary",
		AvailableCapabilityRefs: []DisplaySafeRef{
			"capability:workflow_runtime_backend",
		},
		Attempts: []AttemptSummary{{
			StrategyID:       "strategy:workflow_primary",
			ControlMode:      ControlModeWorkflow,
			Intensity:        IntensityL3ManagedObjective,
			Status:           VerificationPartial,
			ObservationCount: 1,
			FailureClass:     FailureVerificationFailed,
		}},
	})
	if partialRetry.Status != VerificationSatisfied ||
		partialRetry.Selected.Candidate.ID != "strategy:workflow_primary" ||
		!strategyPlannerRankedCandidateExists(partialRetry.RankedCandidates, "strategy:workflow_alternate") {
		t.Fatalf("expected same-strategy workflow retry, got %#v", partialRetry)
	}

	replan := BuildStrategyPlanner(StrategyPlannerInput{
		Activation:         ActivationManaged,
		Frame:              frame,
		Policy:             policy,
		PreGate:            preGate,
		Catalog:            catalog,
		CurrentStrategyRef: "strategy:workflow_primary",
		AvailableCapabilityRefs: []DisplaySafeRef{
			"capability:workflow_runtime_backend",
		},
		Attempts: []AttemptSummary{
			{
				StrategyID:   "strategy:workflow_primary",
				ControlMode:  ControlModeWorkflow,
				Intensity:    IntensityL3ManagedObjective,
				Status:       VerificationFailed,
				FailureClass: FailureVerificationFailed,
			},
			{
				StrategyID:   "strategy:workflow_primary",
				ControlMode:  ControlModeWorkflow,
				Intensity:    IntensityL3ManagedObjective,
				Status:       VerificationFailed,
				FailureClass: FailureVerificationFailed,
			},
		},
	})
	if replan.Status != VerificationSatisfied ||
		replan.Selected.Candidate.ID != "strategy:workflow_alternate" ||
		!strategyPlannerCandidateContains(replan.RejectedCandidates, "strategy:workflow_primary", FailureRepeatedNoProgress, "strategy_repeated_no_progress_dedupe") {
		t.Fatalf("expected workflow cross-strategy replan, got %#v", replan)
	}
}

func workflowStrategyPlannerPolicy() ExecutionIntensityPolicy {
	policy := objectiveLoopIntensityPolicy()
	policy.AllowedControlModesByIntensity[IntensityL3ManagedObjective] = append(
		policy.AllowedControlModesByIntensity[IntensityL3ManagedObjective],
		ControlModeWorkflow,
	)
	return policy
}
