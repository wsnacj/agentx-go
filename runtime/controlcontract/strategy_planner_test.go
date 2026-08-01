package controlcontract

import "testing"

func TestStrategyPlannerRanksMetadataCandidates(t *testing.T) {
	policy := objectiveLoopIntensityPolicy()
	preGate := strategyPlannerPreGate(t, policy, ControlModeOperations, IntensityL3ManagedObjective)
	catalog := StrategyCatalogSnapshot{
		CatalogRef: "catalog:strategy",
		Entries: []StrategyCatalogEntry{
			strategyPlannerCatalogEntry(StrategyCatalogSourceTool, "source:tool_metric", StrategyCandidate{
				ID:             "strategy:tool_metric",
				ControlMode:    ControlModeOperations,
				MinIntensity:   IntensityL3ManagedObjective,
				CapabilityRefs: []DisplaySafeRef{"capability:metric_reader"},
				ExpectedEvidence: []EvidenceRef{{
					Ref:      "evidence:metric",
					Kind:     "metric",
					Strength: EvidenceAdequate,
					Source:   "source:tool_metric",
				}},
				Risk:  "low",
				Owner: "tool",
			}),
			strategyPlannerCatalogEntry(StrategyCatalogSourceHostAdapter, "source:host_metric", StrategyCandidate{
				ID:             "strategy:host_metric",
				ControlMode:    ControlModeOperations,
				MinIntensity:   IntensityL3ManagedObjective,
				CapabilityRefs: []DisplaySafeRef{"capability:metric_reader"},
				ExpectedEvidence: []EvidenceRef{{
					Ref:      "evidence:metric",
					Kind:     "metric",
					Strength: EvidenceStrong,
					Source:   "source:host_metric",
				}},
				Risk:  "read_only",
				Owner: "host",
			}),
		},
	}

	plan := BuildStrategyPlanner(StrategyPlannerInput{
		Activation: ActivationManaged,
		Frame: ObjectiveFrame{
			ID:             "objective:metrics",
			UserGoalDigest: "prefer a lower ranked candidate from text",
			ControlMode:    ControlModeOperations,
			Intensity:      IntensityL3ManagedObjective,
			SuccessCriteria: []string{
				"collect metric evidence",
			},
			RequiredEvidence: []EvidenceRef{{
				Ref:    "evidence:metric",
				Kind:   "metric",
				Source: "source:host_metric",
			}},
		},
		Policy:                  policy,
		PreGate:                 preGate,
		Catalog:                 catalog,
		AvailableCapabilityRefs: []DisplaySafeRef{"capability:metric_reader"},
	})
	if plan.Status != VerificationSatisfied ||
		plan.Selected.Candidate.ID != "strategy:host_metric" ||
		plan.NextHostAction != "run_strategy_final_gate" ||
		plan.RunnerEffect != "none" ||
		plan.PromptEffect != "none" {
		t.Fatalf("ranked strategy plan = %#v", plan)
	}
	if len(plan.RankedCandidates) != 2 ||
		plan.RankedCandidates[0].Rank != 1 ||
		plan.RankedCandidates[1].Rank != 2 {
		t.Fatalf("ranked candidates = %#v", plan.RankedCandidates)
	}
	if !intensityGateBoundaryContains(plan.Boundaries, "metadata_only_planner") ||
		!intensityGateBoundaryContains(plan.Boundaries, "planner_does_not_authorize_execution") ||
		!intensityGateBoundaryContains(plan.Boundaries, "core_must_not_parse_goal_text") {
		t.Fatalf("planner boundaries = %#v", plan.Boundaries)
	}
}

func TestStrategyPlannerRejectsMissingCapabilityAndPolicyBlocked(t *testing.T) {
	policy := objectiveLoopIntensityPolicy()
	preGate := strategyPlannerPreGate(t, policy, ControlModeOperations, IntensityL3ManagedObjective)
	plan := BuildStrategyPlanner(StrategyPlannerInput{
		Activation: ActivationManaged,
		Frame: ObjectiveFrame{
			ID:             "objective:blocked",
			UserGoalDigest: "metadata only",
			ControlMode:    ControlModeOperations,
			Intensity:      IntensityL3ManagedObjective,
		},
		Policy:  policy,
		PreGate: preGate,
		Catalog: StrategyCatalogSnapshot{
			CatalogRef: "catalog:strategy",
			Entries: []StrategyCatalogEntry{
				strategyPlannerCatalogEntry(StrategyCatalogSourceScene, "source:scene_missing", StrategyCandidate{
					ID:             "strategy:missing_capability",
					ControlMode:    ControlModeOperations,
					MinIntensity:   IntensityL3ManagedObjective,
					CapabilityRefs: []DisplaySafeRef{"capability:missing_reader"},
					Owner:          "scene",
				}),
				strategyPlannerCatalogEntry(StrategyCatalogSourceHostAdapter, "source:host_l4", StrategyCandidate{
					ID:           "strategy:l4_candidate",
					ControlMode:  ControlModeOperations,
					MinIntensity: IntensityL4DurableLongRun,
					Owner:        "host",
				}),
			},
		},
	})
	if plan.Status != VerificationBlocked ||
		len(plan.RankedCandidates) != 0 ||
		len(plan.RejectedCandidates) != 2 ||
		plan.NextHostAction != "provide_strategy_catalog_or_upgrade" {
		t.Fatalf("blocked plan = %#v", plan)
	}
	if !strategyPlannerCandidateContains(plan.RejectedCandidates, "strategy:missing_capability", FailureCapabilityMissing, "strategy_capability_not_proven_available") ||
		!strategyPlannerCandidateContains(plan.RejectedCandidates, "strategy:l4_candidate", FailurePolicyBlocked, "strategy_intensity_exceeds_pre_gate") {
		t.Fatalf("rejected candidates = %#v", plan.RejectedCandidates)
	}
	if !intensityGateMissingInputContains(plan.MissingInputs, "host:available_capability:missing_reader") {
		t.Fatalf("missing inputs = %#v", plan.MissingInputs)
	}
}

func TestStrategyPlannerApprovalAndWeakEvidenceSorting(t *testing.T) {
	policy := objectiveLoopIntensityPolicy()
	preGate := strategyPlannerPreGate(t, policy, ControlModeOperations, IntensityL3ManagedObjective)
	plan := BuildStrategyPlanner(StrategyPlannerInput{
		Activation: ActivationManaged,
		Frame: ObjectiveFrame{
			ID:          "objective:evidence",
			ControlMode: ControlModeOperations,
			Intensity:   IntensityL3ManagedObjective,
			RequiredEvidence: []EvidenceRef{{
				Ref:  "evidence:metric",
				Kind: "metric",
			}},
		},
		Policy:  policy,
		PreGate: preGate,
		Catalog: StrategyCatalogSnapshot{
			CatalogRef: "catalog:strategy",
			Entries: []StrategyCatalogEntry{
				strategyPlannerCatalogEntry(StrategyCatalogSourceScene, "source:scene_clean", StrategyCandidate{
					ID:           "strategy:clean",
					ControlMode:  ControlModeOperations,
					MinIntensity: IntensityL3ManagedObjective,
					ExpectedEvidence: []EvidenceRef{{
						Ref:      "evidence:metric",
						Kind:     "metric",
						Strength: EvidenceStrong,
					}},
					Owner: "scene",
				}),
				strategyPlannerCatalogEntry(StrategyCatalogSourceScene, "source:scene_approval", StrategyCandidate{
					ID:               "strategy:approval",
					ControlMode:      ControlModeOperations,
					MinIntensity:     IntensityL3ManagedObjective,
					RequiresApproval: true,
					ExpectedEvidence: []EvidenceRef{{
						Ref:      "evidence:metric",
						Kind:     "metric",
						Strength: EvidenceStrong,
					}},
					Owner: "scene",
				}),
				strategyPlannerCatalogEntry(StrategyCatalogSourceScene, "source:scene_weak", StrategyCandidate{
					ID:           "strategy:weak",
					ControlMode:  ControlModeOperations,
					MinIntensity: IntensityL3ManagedObjective,
					ExpectedEvidence: []EvidenceRef{{
						Ref:      "evidence:metric",
						Kind:     "metric",
						Strength: EvidenceWeak,
					}},
					Owner: "scene",
				}),
			},
		},
	})
	if plan.Status != VerificationSatisfied ||
		len(plan.RankedCandidates) != 3 ||
		plan.RankedCandidates[0].Candidate.ID != "strategy:clean" ||
		plan.RankedCandidates[1].Candidate.ID != "strategy:approval" ||
		plan.RankedCandidates[2].Candidate.ID != "strategy:weak" {
		t.Fatalf("approval/evidence ranking = %#v", plan)
	}
	if !intensityGateBoundaryContains(plan.RankedCandidates[1].Boundaries, "strategy_requires_approval") {
		t.Fatalf("approval candidate boundaries = %#v", plan.RankedCandidates[1].Boundaries)
	}
	if plan.RankedCandidates[2].Status != VerificationReviewRequired ||
		plan.RankedCandidates[2].FailureClass != FailureEvidenceWeak ||
		!intensityGateBoundaryContains(plan.RankedCandidates[2].Boundaries, "strategy_expected_evidence_weak") {
		t.Fatalf("weak evidence candidate = %#v", plan.RankedCandidates[2])
	}
}

func TestStrategyPlannerL2SameStrategyBoundaryAndL3CrossStrategy(t *testing.T) {
	l2Policy := objectiveLoopIntensityPolicy()
	l2Policy.MaxAllowedIntensity = IntensityL2BoundedToolLoop
	l2Policy.AllowedControlModesByIntensity = map[ExecutionIntensity][]ControlMode{
		IntensityL2BoundedToolLoop: {ControlModeTool},
	}
	l2PreGate := strategyPlannerPreGate(t, l2Policy, ControlModeTool, IntensityL2BoundedToolLoop)
	catalog := StrategyCatalogSnapshot{
		CatalogRef: "catalog:strategy",
		Entries: []StrategyCatalogEntry{
			strategyPlannerCatalogEntry(StrategyCatalogSourceTool, "source:tool_current", StrategyCandidate{
				ID:           "strategy:current",
				ControlMode:  ControlModeTool,
				MinIntensity: IntensityL2BoundedToolLoop,
				Owner:        "tool",
			}),
			strategyPlannerCatalogEntry(StrategyCatalogSourceTool, "source:tool_alternate", StrategyCandidate{
				ID:           "strategy:alternate",
				ControlMode:  ControlModeTool,
				MinIntensity: IntensityL2BoundedToolLoop,
				Owner:        "tool",
			}),
		},
	}
	l2Plan := BuildStrategyPlanner(StrategyPlannerInput{
		Activation:         ActivationManaged,
		Frame:              strategyPlannerFrame(ControlModeTool, IntensityL2BoundedToolLoop),
		Policy:             l2Policy,
		PreGate:            l2PreGate,
		Catalog:            catalog,
		CurrentStrategyRef: "strategy:current",
	})
	if l2Plan.Status != VerificationSatisfied ||
		l2Plan.Selected.Candidate.ID != "strategy:current" ||
		!strategyPlannerCandidateContains(l2Plan.RejectedCandidates, "strategy:alternate", FailurePolicyBlocked, "l2_cross_strategy_blocked") {
		t.Fatalf("L2 plan = %#v", l2Plan)
	}

	l3Policy := objectiveLoopIntensityPolicy()
	l3Policy.AllowedControlModesByIntensity = map[ExecutionIntensity][]ControlMode{
		IntensityL2BoundedToolLoop:  {ControlModeTool},
		IntensityL3ManagedObjective: {ControlModeTool},
	}
	l3PreGate := strategyPlannerPreGate(t, l3Policy, ControlModeTool, IntensityL3ManagedObjective)
	l3Plan := BuildStrategyPlanner(StrategyPlannerInput{
		Activation:         ActivationManaged,
		Frame:              strategyPlannerFrame(ControlModeTool, IntensityL3ManagedObjective),
		Policy:             l3Policy,
		PreGate:            l3PreGate,
		Catalog:            catalog,
		CurrentStrategyRef: "strategy:current",
		Attempts: []AttemptSummary{{
			StrategyID:   "strategy:current",
			Status:       VerificationFailed,
			FailureClass: FailureRepeatedNoProgress,
		}},
	})
	if l3Plan.Status != VerificationSatisfied ||
		l3Plan.Selected.Candidate.ID != "strategy:alternate" ||
		!strategyPlannerRankedCandidateExists(l3Plan.RankedCandidates, "strategy:alternate") ||
		!strategyPlannerCandidateContains(l3Plan.RejectedCandidates, "strategy:current", FailureRepeatedNoProgress, "strategy_repeated_no_progress_dedupe") {
		t.Fatalf("L3 cross strategy plan = %#v", l3Plan)
	}
}

func TestStrategyPlannerRequiresManagedActivationAndSatisfiedPreGate(t *testing.T) {
	policy := objectiveLoopIntensityPolicy()
	preGate := strategyPlannerPreGate(t, policy, ControlModeOperations, IntensityL3ManagedObjective)
	catalog := strategyPlannerCatalog("strategy:ready", ControlModeOperations, IntensityL3ManagedObjective)

	inactive := BuildStrategyPlanner(StrategyPlannerInput{
		Activation: ActivationAdvisory,
		Frame:      strategyPlannerFrame(ControlModeOperations, IntensityL3ManagedObjective),
		Policy:     policy,
		PreGate:    preGate,
		Catalog:    catalog,
	})
	if inactive.Status != VerificationBlocked ||
		inactive.FailureClass != FailurePolicyBlocked ||
		inactive.NextHostAction != "enable_managed_objective" {
		t.Fatalf("inactive planner = %#v", inactive)
	}

	blockedPreGate := preGate
	blockedPreGate.Status = VerificationBlocked
	blockedPreGate.FailureClass = FailureApprovalRequired
	blockedPreGate.Allowed = false
	blocked := BuildStrategyPlanner(StrategyPlannerInput{
		Activation: ActivationManaged,
		Frame:      strategyPlannerFrame(ControlModeOperations, IntensityL3ManagedObjective),
		Policy:     policy,
		PreGate:    blockedPreGate,
		Catalog:    catalog,
	})
	if blocked.Status != VerificationBlocked ||
		blocked.FailureClass != FailureApprovalRequired ||
		blocked.NextHostAction != "satisfy_intensity_pre_gate" {
		t.Fatalf("blocked pre-gate planner = %#v", blocked)
	}
}

func strategyPlannerPreGate(t *testing.T, policy ExecutionIntensityPolicy, mode ControlMode, intensity ExecutionIntensity) IntensityGateResult {
	t.Helper()
	gate := BuildExecutionIntensityPreGate(IntensityGateInput{
		Activation:           ActivationManaged,
		Policy:               policy,
		RequestedControlMode: mode,
		RequestedIntensity:   intensity,
		UserConfirmed:        true,
		Budget:               ObjectiveBudgetSnapshot{BudgetRef: "budget:objective", Limit: 3},
	})
	if !gate.Allowed {
		t.Fatalf("pre gate should be allowed: %#v", gate)
	}
	return gate
}

func strategyPlannerFrame(mode ControlMode, intensity ExecutionIntensity) ObjectiveFrame {
	return ObjectiveFrame{
		ID:             "objective:strategy",
		UserGoalDigest: "metadata driven objective",
		ControlMode:    mode,
		Intensity:      intensity,
		SuccessCriteria: []string{
			"select a strategy candidate",
		},
	}
}

func strategyPlannerCatalog(id string, mode ControlMode, intensity ExecutionIntensity) StrategyCatalogSnapshot {
	return StrategyCatalogSnapshot{
		CatalogRef: "catalog:strategy",
		Entries: []StrategyCatalogEntry{
			strategyPlannerCatalogEntry(StrategyCatalogSourceScene, "source:scene_ready", StrategyCandidate{
				ID:           id,
				ControlMode:  mode,
				MinIntensity: intensity,
				Owner:        "scene",
			}),
		},
	}
}

func strategyPlannerCatalogEntry(kind StrategyCatalogSourceKind, sourceRef DisplaySafeRef, candidate StrategyCandidate) StrategyCatalogEntry {
	return StrategyCatalogEntry{
		SourceKind: kind,
		SourceRef:  sourceRef,
		Candidate:  candidate,
		Status:     VerificationSatisfied,
	}
}

func strategyPlannerCandidateContains(candidates []StrategyPlanCandidate, id string, failure FailureClass, boundary Boundary) bool {
	for _, candidate := range candidates {
		if candidate.Candidate.ID != id {
			continue
		}
		return candidate.FailureClass == failure &&
			intensityGateBoundaryContains(candidate.Boundaries, boundary)
	}
	return false
}

func strategyPlannerRankedCandidateExists(candidates []StrategyPlanCandidate, id string) bool {
	for _, candidate := range candidates {
		if candidate.Candidate.ID == id {
			return true
		}
	}
	return false
}
