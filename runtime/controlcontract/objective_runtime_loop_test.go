package controlcontract

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestBuildObjectiveRuntimeLoopStepConsumesSatisfiedLedgerPatch(t *testing.T) {
	input := objectiveRuntimeLoopReadyInput()
	input.Ledger.Attempts = []AttemptSummary{{
		Ref:         "attempt:completion",
		ObjectiveID: "objective:runtime_loop",
		Status:      VerificationPartial,
	}}
	input.LedgerPatch = objectiveRuntimeLoopLedgerPatch(VerificationSatisfied, FailureNone)
	input.Verification = objectiveRuntimeLoopVerificationGate(VerificationSatisfied, FailureNone, "return_satisfied")

	step := BuildObjectiveRuntimeLoopStep(input)
	if step.Status != "ready_for_host_persist" ||
		!step.ReadyForControllerDecision ||
		!step.ReadyForHostPersist ||
		step.ReadyForNextRuntimeAction ||
		step.Run.State != ObjectiveControllerSatisfied ||
		step.ControllerDecision.Action != ObjectiveActionReturnSatisfied ||
		step.RunnerEffect != "none" ||
		step.PromptEffect != "none" ||
		step.RuntimeEffect != "none" ||
		step.LoopEffect != "state_projection_only" {
		t.Fatalf("unexpected satisfied runtime loop step: %#v", step)
	}
	if step.AutoDelegationPolicyReview != nil ||
		step.AutoDelegationInstructionProfile != nil ||
		step.AutoDelegationPlannerCandidate != nil ||
		step.AutoDelegationPlanReview != nil {
		t.Fatalf("default runtime loop should not project auto delegation: %#v", step)
	}
	if len(step.Run.Ledger.Attempts) != 1 ||
		step.Run.Ledger.Attempts[0].Ref != "attempt:completion" ||
		step.Run.Ledger.Attempts[0].Status != VerificationSatisfied {
		t.Fatalf("expected ledger patch to replace matching attempt, got %#v", step.Run.Ledger.Attempts)
	}
	for _, want := range []Boundary{
		"objective_runtime_loop_step",
		"objective_run_state_projection",
		"host_must_persist_objective_run",
		"objective_runtime_loop_ledger_patch_consumed",
		"no_runtime_adapter_execution",
		"no_tool_execution",
	} {
		if !objectiveLoopBoundaryContains(step.Boundaries, want) {
			t.Fatalf("expected boundary %q, got %#v", want, step.Boundaries)
		}
	}
}

func TestBuildObjectiveRuntimeLoopStepAutoDelegationOffDoesNotChangeMainProjection(t *testing.T) {
	baseInput := objectiveRuntimeLoopReadyInput()
	baseInput.LedgerPatch = objectiveRuntimeLoopLedgerPatch(VerificationSatisfied, FailureNone)
	baseInput.Verification = objectiveRuntimeLoopVerificationGate(VerificationSatisfied, FailureNone, "return_satisfied")
	base := BuildObjectiveRuntimeLoopStep(baseInput)

	input := baseInput
	input.AutoDelegationPolicy = AutoDelegationPolicy{Mode: AutoDelegationOff}
	input.AutoDelegationAvailableToolRefs = []DisplaySafeRef{"tool:web_search"}
	got := BuildObjectiveRuntimeLoopStep(input)

	if got.Status != base.Status ||
		got.ReadyForControllerDecision != base.ReadyForControllerDecision ||
		got.ReadyForHostPersist != base.ReadyForHostPersist ||
		got.ReadyForNextRuntimeAction != base.ReadyForNextRuntimeAction ||
		got.NextHostAction != base.NextHostAction ||
		got.FailureClass != base.FailureClass ||
		got.RunnerEffect != "none" ||
		got.PromptEffect != "none" ||
		got.RuntimeEffect != "none" {
		t.Fatalf("auto delegation off changed main projection: base=%#v got=%#v", base, got)
	}
	if got.AutoDelegationPolicyReview == nil ||
		got.AutoDelegationPolicyReview.Status != VerificationNotApplicable ||
		got.AutoDelegationInstructionProfile == nil ||
		got.AutoDelegationInstructionProfile.RequestPlanCandidate ||
		got.AutoDelegationInstructionProfile.ProactiveDelegationInstructionsExposed ||
		len(got.AutoDelegationInstructionProfile.AvailableToolRefs) != 1 {
		t.Fatalf("off policy should project hidden instruction profile: %#v", got)
	}
}

func TestBuildObjectiveRuntimeLoopStepAutoDelegationManagedExposesBoundedInstructionOnly(t *testing.T) {
	input := objectiveRuntimeLoopReadyInput()
	input.LedgerPatch = objectiveRuntimeLoopLedgerPatch(VerificationPartial, FailureEvidenceMissing)
	input.Verification = objectiveRuntimeLoopVerificationGate(VerificationPartial, FailureEvidenceMissing, "request_replan_or_return_partial")
	input.AutoDelegationPolicy = AutoDelegationPolicy{Mode: AutoDelegationManagedReadOnly}
	input.AutoDelegationAvailableToolRefs = []DisplaySafeRef{"tool:web_search", "tool:file_read"}

	step := BuildObjectiveRuntimeLoopStep(input)

	if step.AutoDelegationPolicyReview == nil ||
		!step.AutoDelegationPolicyReview.ManagedExecutionAllowed ||
		step.AutoDelegationInstructionProfile == nil ||
		!step.AutoDelegationInstructionProfile.RequestPlanCandidate ||
		!step.AutoDelegationInstructionProfile.BoundedExecutionInstructionsExposed ||
		step.AutoDelegationPlanReview != nil ||
		step.AutoDelegationPlannerCandidate != nil {
		t.Fatalf("managed policy should expose bounded planner instruction only: %#v", step)
	}
	if step.RunnerEffect != "none" ||
		step.RuntimeEffect != "none" ||
		!objectiveLoopBoundaryContains(step.Boundaries, "no_subagent_dispatch") ||
		!objectiveLoopBoundaryContains(step.Boundaries, "host_review_required_before_dispatch") {
		t.Fatalf("managed instruction must remain projection-only: %#v", step)
	}
}

func TestBuildObjectiveRuntimeLoopStepAutoDelegationCandidateReviewedAfterFrame(t *testing.T) {
	input := objectiveRuntimeLoopReadyInput()
	input.LedgerPatch = objectiveRuntimeLoopLedgerPatch(VerificationPartial, FailureEvidenceMissing)
	input.Verification = objectiveRuntimeLoopVerificationGate(VerificationPartial, FailureEvidenceMissing, "request_replan_or_return_partial")
	input.AutoDelegationPolicy = AutoDelegationPolicy{
		Mode:        AutoDelegationManagedReadOnly,
		MaxChildren: 2,
	}
	input.AutoDelegationPlannerCandidateJSON = validAutoDelegationPlannerPlanJSON()

	step := BuildObjectiveRuntimeLoopStep(input)

	if step.AutoDelegationPlannerCandidate == nil ||
		!step.AutoDelegationPlannerCandidate.Decoded ||
		step.AutoDelegationPlanReview == nil ||
		!step.AutoDelegationPlanReview.Ready ||
		!step.AutoDelegationPlanReview.HostMayDispatch ||
		step.AutoDelegationPlanReview.Plan.ParentObjectiveRef != "objective:runtime_loop" {
		t.Fatalf("planner candidate should be decoded and reviewed after frame: %#v", step)
	}
	if step.RunnerEffect != "none" ||
		step.RuntimeEffect != "none" ||
		!objectiveLoopBoundaryContains(step.Boundaries, "auto_delegation_planner_candidate_validated") ||
		!objectiveLoopBoundaryContains(step.Boundaries, "no_tool_execution") {
		t.Fatalf("candidate review should not execute tools or workers: %#v", step)
	}
}

func TestBuildObjectiveRuntimeLoopStepInvalidAutoDelegationCandidateDoesNotBreakMainLoop(t *testing.T) {
	input := objectiveRuntimeLoopReadyInput()
	input.LedgerPatch = objectiveRuntimeLoopLedgerPatch(VerificationSatisfied, FailureNone)
	input.Verification = objectiveRuntimeLoopVerificationGate(VerificationSatisfied, FailureNone, "return_satisfied")
	input.AutoDelegationPolicy = AutoDelegationPolicy{Mode: AutoDelegationManagedReadOnly}
	input.AutoDelegationPlannerCandidateJSON = "split this into child tasks"

	step := BuildObjectiveRuntimeLoopStep(input)

	if step.Status != "ready_for_host_persist" ||
		!step.ReadyForHostPersist ||
		step.AutoDelegationPlannerCandidate == nil ||
		step.AutoDelegationPlannerCandidate.Status != VerificationBlocked ||
		step.AutoDelegationPlanReview != nil {
		t.Fatalf("invalid delegation candidate should be nested-blocked without breaking main loop: %#v", step)
	}
}

func TestBuildObjectiveRuntimeLoopStepConsumesAutoDelegationParentMerge(t *testing.T) {
	bridge := autoDelegationParentMergeReadyBridge(t)
	invocation := autoDelegationParentMergeCompletedInvocation(t, bridge.Children[0].Readiness)
	parentMerge := BuildAutoDelegationParentMerge(AutoDelegationParentMergeInput{
		HostBridge:       bridge,
		ParentFrame:      autoDelegationParentMergeFrameFixture(),
		ParentLedgerRef:  "ledger:auto_delegation_parent",
		FailureReviewRef: "failure_review:auto_delegation_parent",
		FailureRef:       "failure:auto_delegation_parent",
		ChildResults: []AutoDelegationChildMergeInput{
			autoDelegationChildMergeInputFixture("child:collect_public_sources", invocation, "value:one", EvidenceAdequate),
		},
	})
	if !parentMerge.ReadyForParentMerge {
		t.Fatalf("parent merge fixture should be ready: %+v", parentMerge)
	}
	input := objectiveRuntimeLoopReadyInput()
	input.Frame.ID = "objective:root"
	input.Ledger.ObjectiveID = "objective:root"
	input.Ledger.LedgerRef = "ledger:auto_delegation_parent"
	input.Verification = objectiveRuntimeLoopVerificationGate(VerificationSatisfied, FailureNone, "return_satisfied")
	input.Verification.Frame.ID = "objective:root"
	input.AutoDelegationParentMerge = parentMerge

	step := BuildObjectiveRuntimeLoopStep(input)

	if step.AutoDelegationParentMerge == nil ||
		!step.AutoDelegationParentMerge.ReadyForParentMerge ||
		len(step.LedgerPatch.Attempts) != 1 ||
		step.LedgerPatch.Attempts[0].Ref != "attempt:child:collect_public_sources" ||
		len(step.Run.Ledger.Attempts) != 1 ||
		step.Run.Ledger.Attempts[0].ControlMode != ControlModeDelegated ||
		!objectiveLoopEvidenceContains(step.EvidenceRefs, "evidence:public_source_summary") ||
		!objectiveLoopObservationKindContains(step.Observations, "auto_delegation_child_result") ||
		!objectiveLoopBoundaryContains(step.Boundaries, "auto_delegation_parent_merge_ready") {
		t.Fatalf("runtime loop should consume AD6 parent merge: %+v", step)
	}
}

func TestBuildObjectiveRuntimeLoopStepMapsPartialToReplan(t *testing.T) {
	input := objectiveRuntimeLoopReadyInput()
	input.LedgerPatch = objectiveRuntimeLoopLedgerPatch(VerificationPartial, FailureEvidenceMissing)
	input.Verification = objectiveRuntimeLoopVerificationGate(VerificationPartial, FailureEvidenceMissing, "request_replan_or_return_partial")

	step := BuildObjectiveRuntimeLoopStep(input)
	if step.Status != "ready_for_host_persist" ||
		step.Run.State != ObjectiveControllerPartial ||
		step.ControllerDecision.Action != ObjectiveActionRequestReplanDecision ||
		step.FailureClass != FailureEvidenceMissing ||
		step.NextHostAction != "request_host_replanner_decision" ||
		!step.ReadyForNextRuntimeAction ||
		len(step.LedgerPatch.Attempts) != 1 ||
		step.LedgerPatch.Attempts[0].Status != VerificationPartial {
		t.Fatalf("unexpected partial runtime loop step: %#v", step)
	}
}

func TestBuildObjectiveRuntimeLoopStepDoesNotActivateOutsideManaged(t *testing.T) {
	input := objectiveRuntimeLoopReadyInput()
	input.Activation = ActivationOff
	input.LedgerPatch = objectiveRuntimeLoopLedgerPatch(VerificationSatisfied, FailureNone)
	input.Verification = objectiveRuntimeLoopVerificationGate(VerificationSatisfied, FailureNone, "return_satisfied")

	step := BuildObjectiveRuntimeLoopStep(input)
	if step.Available ||
		step.Status != "inactive" ||
		step.Run.FullRun ||
		step.Run.State != ObjectiveControllerInactive ||
		step.ControllerDecision.Action != ObjectiveActionNone ||
		step.ReadyForControllerDecision ||
		step.ReadyForHostPersist ||
		step.ReadyForNextRuntimeAction ||
		step.NextHostAction != "enable_managed_objective" {
		t.Fatalf("A0/off should not activate runtime loop: %#v", step)
	}
}

func TestBuildObjectiveRuntimeLoopStepRawOutputForcesReviewWithoutLeak(t *testing.T) {
	input := objectiveRuntimeLoopReadyInput()
	unsafePath := "/" + "Users/example/raw-runtime-loop-secret"
	input.RawOutputLoaded = true
	input.LedgerPatch = objectiveRuntimeLoopLedgerPatch(VerificationSatisfied, FailureNone)
	input.LedgerPatch.LedgerRef = DisplaySafeRef(unsafePath)
	input.Verification = objectiveRuntimeLoopVerificationGate(VerificationSatisfied, FailureNone, "return_satisfied")
	input.Verification.EvidenceRefs = []EvidenceRef{{Ref: DisplaySafeRef(unsafePath), Kind: "metric", Strength: EvidenceStrong}}
	input.Observations = []Observation{{
		Kind:            "metric",
		DisplaySafeRefs: []DisplaySafeRef{DisplaySafeRef(unsafePath)},
		RawOutputLoaded: true,
	}}

	step := BuildObjectiveRuntimeLoopStep(input)
	if step.Status != "review_required" ||
		step.ReadyForControllerDecision ||
		step.ReadyForHostPersist ||
		step.ReadyForNextRuntimeAction ||
		step.FailureClass != FailureEvidenceWeak ||
		step.NextHostAction != "provide_display_safe_refs" ||
		!objectiveLoopMissingInputContains(step.MissingInputs, "host:display_safe_refs") ||
		!objectiveLoopBoundaryContains(step.Boundaries, "raw_output_not_allowed") {
		t.Fatalf("raw output should force review, got %#v", step)
	}
	payload, err := json.Marshal(step)
	if err != nil {
		t.Fatalf("marshal runtime loop step: %v", err)
	}
	if strings.Contains(string(payload), unsafePath) {
		t.Fatalf("runtime loop leaked unsafe ref %q in %s", unsafePath, payload)
	}
}

func objectiveRuntimeLoopReadyInput() ObjectiveRuntimeLoopInput {
	return ObjectiveRuntimeLoopInput{
		Activation: ActivationManaged,
		Frame: ObjectiveFrame{
			ID:              "objective:runtime_loop",
			UserGoalDigest:  "goal:runtime_loop",
			ControlMode:     ControlModeObjective,
			Intensity:       IntensityL3ManagedObjective,
			SuccessCriteria: []string{"objective runtime loop consumes verified ledger patch"},
		},
		Ledger: AttemptLedgerPatch{
			ObjectiveID: "objective:runtime_loop",
			LedgerRef:   "ledger:runtime_loop",
		},
		Budget: ObjectiveBudgetSnapshot{
			BudgetRef: "budget:runtime_loop",
			Limit:     3,
		},
		Approval: ObjectiveApprovalState{
			Required: false,
		},
		Strategies: []StrategyCandidate{{
			ID:           "strategy:runtime_loop",
			Kind:         "host_adapter",
			ControlMode:  ControlModeObjective,
			MinIntensity: IntensityL3ManagedObjective,
			MaxIntensity: IntensityL3ManagedObjective,
			Owner:        "host",
			ExpectedEvidence: []EvidenceRef{{
				Ref:      "evidence:runtime_loop",
				Kind:     "metric",
				Strength: EvidenceAdequate,
			}},
		}},
	}
}

func objectiveRuntimeLoopLedgerPatch(status VerificationStatus, failure FailureClass) AttemptLedgerPatch {
	return AttemptLedgerPatch{
		ObjectiveID: "objective:runtime_loop",
		LedgerRef:   "ledger:runtime_loop",
		EvidenceRefs: []EvidenceRef{{
			Ref:      "evidence:runtime_loop",
			Kind:     "metric",
			Strength: EvidenceStrong,
			Source:   "source:runtime_loop",
		}},
		Attempts: []AttemptSummary{{
			Ref:              "attempt:completion",
			ObjectiveID:      "objective:runtime_loop",
			StrategyID:       "strategy:runtime_loop",
			ControlMode:      ControlModeObjective,
			Intensity:        IntensityL3ManagedObjective,
			Status:           status,
			EvidenceRefs:     []EvidenceRef{{Ref: "evidence:runtime_loop", Kind: "metric", Strength: EvidenceStrong}},
			ObservationCount: 1,
			FailureClass:     failure,
		}},
		NextHostAction: "host_may_update_objective_controller",
	}
}

func objectiveRuntimeLoopVerificationGate(status VerificationStatus, failure FailureClass, next NextHostAction) ObjectiveVerificationGateResult {
	verification := VerificationResult{
		Status:         status,
		FailureClass:   failure,
		EvidenceRefs:   []EvidenceRef{{Ref: "evidence:runtime_loop", Kind: "metric", Strength: EvidenceStrong}},
		NextHostAction: next,
	}.Normalize()
	return ObjectiveVerificationGateResult{
		ContractVersion: ContractVersion,
		Projected:       true,
		Status:          verification.Status,
		Satisfied:       verification.Satisfied,
		Frame: ObjectiveFrame{
			ID:             "objective:runtime_loop",
			UserGoalDigest: "goal:runtime_loop",
		},
		Observations: []Observation{{
			Kind:     "metric",
			Source:   "source:runtime_loop",
			Name:     "runtime_loop_result",
			Value:    string(status),
			Strength: EvidenceStrong,
			EvidenceRefs: []EvidenceRef{{
				Ref:      "evidence:runtime_loop",
				Kind:     "metric",
				Strength: EvidenceStrong,
			}},
		}},
		EvidenceRefs:   verification.EvidenceRefs,
		Verification:   verification,
		FailureClass:   failure,
		NextHostAction: next,
		RunnerEffect:   "none",
		PromptEffect:   "none",
	}.Normalize()
}

func objectiveLoopEvidenceContains(values []EvidenceRef, ref DisplaySafeRef) bool {
	for _, value := range values {
		if value.Ref == ref {
			return true
		}
	}
	return false
}

func objectiveLoopObservationKindContains(values []Observation, kind string) bool {
	for _, value := range values {
		if value.Kind == kind {
			return true
		}
	}
	return false
}
