package controlcontract

import "strings"

type ObjectiveRuntimeLoopInput struct {
	Run                                   ObjectiveRun                    `json:"run,omitempty"`
	Activation                            Activation                      `json:"activation,omitempty"`
	Frame                                 ObjectiveFrame                  `json:"frame,omitempty"`
	Ledger                                AttemptLedgerPatch              `json:"ledger,omitempty"`
	LedgerPatch                           AttemptLedgerPatch              `json:"ledger_patch,omitempty"`
	Budget                                ObjectiveBudgetSnapshot         `json:"budget,omitempty"`
	Approval                              ObjectiveApprovalState          `json:"approval,omitempty"`
	PolicyRefs                            []DisplaySafeRef                `json:"policy_refs,omitempty"`
	Strategies                            []StrategyCandidate             `json:"strategies,omitempty"`
	CurrentStrategyRef                    DisplaySafeRef                  `json:"current_strategy_ref,omitempty"`
	Verification                          ObjectiveVerificationGateResult `json:"verification,omitempty"`
	VerificationResult                    VerificationResult              `json:"verification_result,omitempty"`
	Observations                          []Observation                   `json:"observations,omitempty"`
	EvidenceRefs                          []EvidenceRef                   `json:"evidence_refs,omitempty"`
	AutoDelegationPolicy                  AutoDelegationPolicy            `json:"auto_delegation_policy,omitempty"`
	AutoDelegationAvailableToolRefs       []DisplaySafeRef                `json:"auto_delegation_available_tool_refs,omitempty"`
	AutoDelegationAvailableCapabilityRefs []DisplaySafeRef                `json:"auto_delegation_available_capability_refs,omitempty"`
	AutoDelegationPlannerCandidateJSON    string                          `json:"-"`
	AutoDelegationPlan                    AutoDelegationPlan              `json:"auto_delegation_plan,omitempty"`
	AutoDelegationParentMerge             AutoDelegationParentMerge       `json:"auto_delegation_parent_merge,omitempty"`
	Boundaries                            []Boundary                      `json:"boundaries,omitempty"`
	RawOutputLoaded                       bool                            `json:"raw_output_loaded"`
}

type ObjectiveRuntimeLoopStep struct {
	ContractVersion                  string                                          `json:"contract_version,omitempty"`
	Projected                        bool                                            `json:"projected"`
	Available                        bool                                            `json:"available"`
	Status                           string                                          `json:"status,omitempty"`
	LoopEffect                       string                                          `json:"loop_effect,omitempty"`
	RunnerEffect                     string                                          `json:"runner_effect,omitempty"`
	PromptEffect                     string                                          `json:"prompt_effect,omitempty"`
	RuntimeEffect                    string                                          `json:"runtime_effect,omitempty"`
	Run                              ObjectiveRun                                    `json:"run,omitempty"`
	LedgerPatch                      AttemptLedgerPatch                              `json:"ledger_patch,omitempty"`
	ControllerDecision               ObjectiveControllerDecision                     `json:"controller_decision,omitempty"`
	Verification                     ObjectiveVerificationGateResult                 `json:"verification,omitempty"`
	Observations                     []Observation                                   `json:"observations,omitempty"`
	EvidenceRefs                     []EvidenceRef                                   `json:"evidence_refs,omitempty"`
	AutoDelegationPolicyReview       *AutoDelegationPolicyReview                     `json:"auto_delegation_policy_review,omitempty"`
	AutoDelegationInstructionProfile *AutoDelegationPlannerInstructionProfile        `json:"auto_delegation_instruction_profile,omitempty"`
	AutoDelegationPlannerCandidate   *AutoDelegationPlannerCandidateJSONDecodeReport `json:"auto_delegation_planner_candidate,omitempty"`
	AutoDelegationPlanReview         *AutoDelegationPlanReview                       `json:"auto_delegation_plan_review,omitempty"`
	AutoDelegationParentMerge        *AutoDelegationParentMerge                      `json:"auto_delegation_parent_merge,omitempty"`
	FailureClass                     FailureClass                                    `json:"failure_class,omitempty"`
	MissingInputs                    []MissingInput                                  `json:"missing_inputs,omitempty"`
	Boundaries                       []Boundary                                      `json:"boundaries,omitempty"`
	NextHostAction                   NextHostAction                                  `json:"next_host_action,omitempty"`
	ReadyForControllerDecision       bool                                            `json:"ready_for_controller_decision"`
	ReadyForHostPersist              bool                                            `json:"ready_for_host_persist"`
	ReadyForNextRuntimeAction        bool                                            `json:"ready_for_next_runtime_action"`
	RawOutputLoaded                  bool                                            `json:"raw_output_loaded"`
}

func BuildObjectiveRuntimeLoopStep(input ObjectiveRuntimeLoopInput) ObjectiveRuntimeLoopStep {
	step := ObjectiveRuntimeLoopStep{
		ContractVersion: ContractVersion,
		Projected:       true,
		Available:       true,
		Status:          "blocked",
		LoopEffect:      "state_projection_only",
		RunnerEffect:    "none",
		PromptEffect:    "none",
		RuntimeEffect:   "none",
		Boundaries: MergeBoundaries(
			[]Boundary{
				"objective_runtime_loop_step",
				"objective_run_state_projection",
				"host_must_persist_objective_run",
				"no_runner_dispatch",
				"no_runtime_adapter_execution",
				"no_tool_execution",
				"no_workflow_dispatch",
				"no_install_or_schedule_apply",
				"projection_only",
			},
			input.Boundaries,
		),
		NextHostAction: "provide_objective_runtime_loop_inputs",
	}

	currentRun := input.Run.Normalize()
	hasRun := !objectiveRunInputEmpty(input.Run)
	activation := objectiveRuntimeLoopActivation(input, currentRun, hasRun)
	rawOutputLoaded := input.RawOutputLoaded || currentRun.RawOutputLoaded || input.Ledger.RawOutputLoaded || input.LedgerPatch.RawOutputLoaded

	frame := objectiveRuntimeLoopFrame(input, currentRun, hasRun)
	autoDelegationPolicyReview, autoDelegationInstructionProfile, autoDelegationPlannerCandidate, autoDelegationPlanReview := objectiveRuntimeLoopAutoDelegation(input, frame)
	autoDelegationParentMerge := objectiveRuntimeLoopAutoDelegationParentMerge(input)
	budget := objectiveRuntimeLoopBudget(input, currentRun, hasRun)
	approval := objectiveRuntimeLoopApproval(input, currentRun, hasRun)
	policyRefs := objectiveRuntimeLoopPolicyRefs(input, currentRun, hasRun)
	strategies := objectiveRuntimeLoopStrategies(input, currentRun, hasRun)
	currentStrategyRef := objectiveRuntimeLoopCurrentStrategyRef(input, currentRun, hasRun)
	verificationGate, verificationResult, hasVerification := objectiveRuntimeLoopVerification(input, currentRun, hasRun)
	observations := normalizeObservations(append(cloneObservations(currentRun.LatestObservations), input.Observations...))
	observations = normalizeObservations(append(observations, verificationGate.Observations...))
	for _, observation := range observations {
		rawOutputLoaded = rawOutputLoaded || observation.RawOutputLoaded
	}
	rawOutputLoaded = rawOutputLoaded ||
		objectiveRuntimeLoopAutoDelegationRawOutput(autoDelegationPolicyReview, autoDelegationInstructionProfile, autoDelegationPlannerCandidate, autoDelegationPlanReview, autoDelegationParentMerge)

	ledgerPatch := objectiveRuntimeLoopAutoDelegationLedgerPatch(input.LedgerPatch, autoDelegationParentMerge)
	ledger := objectiveRuntimeLoopMergeLedger(objectiveRuntimeLoopLedger(input, currentRun, hasRun), ledgerPatch)
	observations = append(observations, objectiveRuntimeLoopAutoDelegationParentMergeObservations(autoDelegationParentMerge)...)
	evidenceRefs := MergeEvidenceRefs(input.EvidenceRefs, currentRun.EvidenceRefs, ledger.EvidenceRefs, verificationResult.EvidenceRefs, verificationGate.EvidenceRefs, objectiveObservationEvidenceRefs(observations), objectiveRuntimeLoopAutoDelegationParentMergeEvidenceRefs(autoDelegationParentMerge))
	run := BuildObjectiveRun(ObjectiveRunInput{
		Activation:         activation,
		Frame:              frame,
		Ledger:             ledger,
		Budget:             budget,
		Approval:           approval,
		PolicyRefs:         policyRefs,
		Strategies:         strategies,
		CurrentStrategyRef: currentStrategyRef,
		Verification:       verificationResult,
		Observations:       observations,
		EvidenceRefs:       evidenceRefs,
		Boundaries: MergeBoundaries(
			currentRun.Boundaries,
			ledger.Boundaries,
			verificationResult.Boundaries,
			verificationGate.Boundaries,
			objectiveRuntimeLoopAutoDelegationBoundaries(autoDelegationPolicyReview, autoDelegationInstructionProfile, autoDelegationPlannerCandidate, autoDelegationPlanReview, autoDelegationParentMerge),
			step.Boundaries,
		),
		RawOutputLoaded: rawOutputLoaded,
	})
	if !hasVerification {
		run.MissingInputs = AppendMissingInputs(run.MissingInputs, "host:objective_verification")
		run.Boundaries = AppendBoundaries(run.Boundaries, "objective_runtime_loop_verification_missing")
		if run.FailureClass == FailureNone {
			run.FailureClass = FailureEvidenceMissing
		}
	}
	if objectiveRuntimeLoopLedgerEmpty(ledgerPatch) {
		run.MissingInputs = AppendMissingInputs(run.MissingInputs, "host:objective_run_ledger_patch")
		run.Boundaries = AppendBoundaries(run.Boundaries, "objective_runtime_loop_ledger_patch_missing")
		if run.FailureClass == FailureNone {
			run.FailureClass = FailureEvidenceMissing
		}
	}
	run = run.Normalize()
	decision := BuildObjectiveControllerDecision(ObjectiveControllerInput{Run: run, RawOutputLoaded: rawOutputLoaded})

	step.Run = run
	step.LedgerPatch = ledgerPatch.Normalize()
	step.ControllerDecision = decision
	step.Verification = verificationGate.Normalize()
	step.Observations = observations
	step.EvidenceRefs = MergeEvidenceRefs(evidenceRefs, decision.EvidenceRefs)
	step.AutoDelegationPolicyReview = autoDelegationPolicyReview
	step.AutoDelegationInstructionProfile = autoDelegationInstructionProfile
	step.AutoDelegationPlannerCandidate = autoDelegationPlannerCandidate
	step.AutoDelegationPlanReview = autoDelegationPlanReview
	step.AutoDelegationParentMerge = autoDelegationParentMerge
	step.FailureClass = firstFailureClass(run.FailureClass, decision.FailureClass)
	step.MissingInputs = MergeMissingInputs(run.MissingInputs, decision.MissingInputs)
	step.Boundaries = MergeBoundaries(
		step.Boundaries,
		run.Boundaries,
		decision.Boundaries,
		objectiveRuntimeLoopAutoDelegationBoundaries(autoDelegationPolicyReview, autoDelegationInstructionProfile, autoDelegationPlannerCandidate, autoDelegationPlanReview, autoDelegationParentMerge),
	)
	step.NextHostAction = firstNextHostAction(decision.NextHostAction, run.NextHostAction)
	step.RawOutputLoaded = rawOutputLoaded || run.RawOutputLoaded || decision.RawOutputLoaded
	step.ReadyForControllerDecision = objectiveRuntimeLoopReadyForControllerDecision(run, decision, hasVerification, step.RawOutputLoaded)
	step.ReadyForHostPersist = step.ReadyForControllerDecision && !objectiveRuntimeLoopLedgerEmpty(ledgerPatch) && run.FullRun
	step.ReadyForNextRuntimeAction = step.ReadyForControllerDecision && objectiveRuntimeLoopReadyForNextRuntimeAction(decision)
	step.Status = objectiveRuntimeLoopStatus(step, run, decision, activation, hasVerification)
	if step.RawOutputLoaded {
		step.ReadyForControllerDecision = false
		step.ReadyForHostPersist = false
		step.ReadyForNextRuntimeAction = false
		step.Status = "review_required"
		step.FailureClass = firstFailureClass(step.FailureClass, FailureEvidenceWeak)
		step.MissingInputs = AppendMissingInputs(step.MissingInputs, "host:display_safe_refs")
		step.Boundaries = AppendBoundaries(step.Boundaries, "raw_output_not_allowed")
		step.NextHostAction = "provide_display_safe_refs"
	}
	if activation != ActivationManaged {
		step.Available = false
		step.ReadyForControllerDecision = false
		step.ReadyForHostPersist = false
		step.ReadyForNextRuntimeAction = false
		step.Status = "inactive"
	}
	return step.Normalize()
}

func CloneObjectiveRuntimeLoopStep(in ObjectiveRuntimeLoopStep) ObjectiveRuntimeLoopStep {
	out := in
	out.Run = in.Run.Clone()
	out.LedgerPatch = in.LedgerPatch.Clone()
	out.ControllerDecision = in.ControllerDecision.Clone()
	out.Verification = in.Verification.Clone()
	out.Observations = cloneObservations(in.Observations)
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.AutoDelegationPolicyReview = cloneAutoDelegationPolicyReviewPtr(in.AutoDelegationPolicyReview)
	out.AutoDelegationInstructionProfile = cloneAutoDelegationPlannerInstructionProfilePtr(in.AutoDelegationInstructionProfile)
	out.AutoDelegationPlannerCandidate = cloneAutoDelegationPlannerCandidateJSONDecodeReportPtr(in.AutoDelegationPlannerCandidate)
	out.AutoDelegationPlanReview = cloneAutoDelegationPlanReviewPtr(in.AutoDelegationPlanReview)
	out.AutoDelegationParentMerge = cloneAutoDelegationParentMergePtr(in.AutoDelegationParentMerge)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (s ObjectiveRuntimeLoopStep) Clone() ObjectiveRuntimeLoopStep {
	return CloneObjectiveRuntimeLoopStep(s)
}

func (s ObjectiveRuntimeLoopStep) Normalize() ObjectiveRuntimeLoopStep {
	out := CloneObjectiveRuntimeLoopStep(s)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = normalizeControlToken(out.Status)
	out.LoopEffect = normalizeControlToken(out.LoopEffect)
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	out.RuntimeEffect = normalizeControlToken(out.RuntimeEffect)
	out.Run = out.Run.Normalize()
	out.LedgerPatch = out.LedgerPatch.Normalize()
	out.ControllerDecision = out.ControllerDecision.Normalize()
	out.Verification = out.Verification.Normalize()
	out.Observations = normalizeObservations(out.Observations)
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
	out.AutoDelegationPolicyReview = normalizeAutoDelegationPolicyReviewPtr(out.AutoDelegationPolicyReview)
	out.AutoDelegationInstructionProfile = normalizeAutoDelegationPlannerInstructionProfilePtr(out.AutoDelegationInstructionProfile)
	out.AutoDelegationPlannerCandidate = normalizeAutoDelegationPlannerCandidateJSONDecodeReportPtr(out.AutoDelegationPlannerCandidate)
	out.AutoDelegationPlanReview = normalizeAutoDelegationPlanReviewPtr(out.AutoDelegationPlanReview)
	out.AutoDelegationParentMerge = normalizeAutoDelegationParentMergePtr(out.AutoDelegationParentMerge)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	if out.Status == "" {
		out.Status = "blocked"
	}
	if out.LoopEffect == "" {
		out.LoopEffect = "state_projection_only"
	}
	if out.RunnerEffect == "" {
		out.RunnerEffect = "none"
	}
	if out.PromptEffect == "" {
		out.PromptEffect = "none"
	}
	if out.RuntimeEffect == "" {
		out.RuntimeEffect = "none"
	}
	if out.RawOutputLoaded {
		out.ReadyForControllerDecision = false
		out.ReadyForHostPersist = false
		out.ReadyForNextRuntimeAction = false
		out.Status = "review_required"
		if out.FailureClass == FailureNone {
			out.FailureClass = FailureEvidenceWeak
		}
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		out.NextHostAction = "provide_display_safe_refs"
	}
	return out
}

func objectiveRuntimeLoopAutoDelegation(input ObjectiveRuntimeLoopInput, frame ObjectiveFrame) (*AutoDelegationPolicyReview, *AutoDelegationPlannerInstructionProfile, *AutoDelegationPlannerCandidateJSONDecodeReport, *AutoDelegationPlanReview) {
	if !objectiveRuntimeLoopAutoDelegationPresent(input) {
		return nil, nil, nil, nil
	}
	policyReview := BuildAutoDelegationPolicyReview(input.AutoDelegationPolicy)
	instructionProfile := BuildAutoDelegationPlannerInstructionProfile(AutoDelegationPlannerInstructionInput{
		PolicyReview:            policyReview,
		AvailableToolRefs:       input.AutoDelegationAvailableToolRefs,
		AvailableCapabilityRefs: input.AutoDelegationAvailableCapabilityRefs,
		Boundaries:              []Boundary{"objective_runtime_loop_auto_delegation_instruction_profile"},
		RawOutputLoaded:         input.RawOutputLoaded,
	})
	var candidateReport *AutoDelegationPlannerCandidateJSONDecodeReport
	var planReview *AutoDelegationPlanReview
	if stringsTrimmedNotEmpty(input.AutoDelegationPlannerCandidateJSON) {
		report := BuildAutoDelegationPlannerCandidateFromJSON(AutoDelegationPlannerCandidateJSONDecodeInput{
			RawJSON:            input.AutoDelegationPlannerCandidateJSON,
			PolicyReview:       policyReview,
			SourceRef:          "source:auto_delegation_planner",
			ParentObjectiveRef: DisplaySafeRef(frame.ID),
			Boundaries:         []Boundary{"objective_runtime_loop_auto_delegation_candidate_decode"},
			RawOutputLoaded:    input.RawOutputLoaded,
		})
		candidateReport = &report
		if report.Decoded {
			review := report.PlanReview.Normalize()
			if !autoDelegationPlanReviewEmpty(review) {
				planReview = &review
			}
		}
	} else if !autoDelegationPlanEmpty(input.AutoDelegationPlan) {
		plan := input.AutoDelegationPlan.Clone()
		if plan.ParentObjectiveRef == "" {
			plan.ParentObjectiveRef = DisplaySafeRef(frame.ID)
		}
		if autoDelegationPolicyEmpty(plan.Policy) {
			plan.Policy = policyReview.Policy
		}
		review := BuildAutoDelegationPlanReview(policyReview, plan)
		planReview = &review
	}
	return &policyReview, &instructionProfile, candidateReport, planReview
}

func objectiveRuntimeLoopAutoDelegationParentMerge(input ObjectiveRuntimeLoopInput) *AutoDelegationParentMerge {
	if autoDelegationParentMergeEmpty(input.AutoDelegationParentMerge) {
		return nil
	}
	parentMerge := input.AutoDelegationParentMerge.Normalize()
	return &parentMerge
}

func objectiveRuntimeLoopAutoDelegationPresent(input ObjectiveRuntimeLoopInput) bool {
	return !autoDelegationPolicyEmpty(input.AutoDelegationPolicy) ||
		stringsTrimmedNotEmpty(input.AutoDelegationPlannerCandidateJSON) ||
		!autoDelegationPlanEmpty(input.AutoDelegationPlan)
}

func objectiveRuntimeLoopAutoDelegationRawOutput(policyReview *AutoDelegationPolicyReview, instructionProfile *AutoDelegationPlannerInstructionProfile, candidateReport *AutoDelegationPlannerCandidateJSONDecodeReport, planReview *AutoDelegationPlanReview, parentMerge *AutoDelegationParentMerge) bool {
	return (policyReview != nil && policyReview.RawOutputLoaded) ||
		(instructionProfile != nil && instructionProfile.RawOutputLoaded) ||
		(candidateReport != nil && candidateReport.RawOutputLoaded) ||
		(planReview != nil && planReview.RawOutputLoaded) ||
		(parentMerge != nil && parentMerge.RawOutputLoaded)
}

func objectiveRuntimeLoopAutoDelegationBoundaries(policyReview *AutoDelegationPolicyReview, instructionProfile *AutoDelegationPlannerInstructionProfile, candidateReport *AutoDelegationPlannerCandidateJSONDecodeReport, planReview *AutoDelegationPlanReview, parentMerge *AutoDelegationParentMerge) []Boundary {
	var groups [][]Boundary
	if policyReview != nil {
		groups = append(groups, policyReview.Boundaries)
	}
	if instructionProfile != nil {
		groups = append(groups, instructionProfile.Boundaries)
	}
	if candidateReport != nil {
		groups = append(groups, candidateReport.Boundaries)
	}
	if planReview != nil {
		groups = append(groups, planReview.Boundaries)
	}
	if parentMerge != nil {
		groups = append(groups, parentMerge.Boundaries)
	}
	return MergeBoundaries(groups...)
}

func objectiveRuntimeLoopAutoDelegationLedgerPatch(inputPatch AttemptLedgerPatch, parentMerge *AutoDelegationParentMerge) AttemptLedgerPatch {
	patch := inputPatch.Normalize()
	if parentMerge == nil || !parentMerge.ReadyForParentMerge {
		return patch
	}
	return objectiveRuntimeLoopMergeLedger(patch, parentMerge.ParentLedgerPatch)
}

func objectiveRuntimeLoopAutoDelegationParentMergeEvidenceRefs(parentMerge *AutoDelegationParentMerge) []EvidenceRef {
	if parentMerge == nil || !parentMerge.ReadyForParentMerge {
		return nil
	}
	return cloneEvidenceRefs(parentMerge.EvidenceRefs)
}

func objectiveRuntimeLoopAutoDelegationParentMergeObservations(parentMerge *AutoDelegationParentMerge) []Observation {
	if parentMerge == nil || !parentMerge.ReadyForParentMerge {
		return nil
	}
	return cloneObservations(parentMerge.Observations)
}

func cloneAutoDelegationPolicyReviewPtr(in *AutoDelegationPolicyReview) *AutoDelegationPolicyReview {
	if in == nil {
		return nil
	}
	out := in.Clone()
	return &out
}

func cloneAutoDelegationPlannerInstructionProfilePtr(in *AutoDelegationPlannerInstructionProfile) *AutoDelegationPlannerInstructionProfile {
	if in == nil {
		return nil
	}
	out := in.Clone()
	return &out
}

func cloneAutoDelegationPlannerCandidateJSONDecodeReportPtr(in *AutoDelegationPlannerCandidateJSONDecodeReport) *AutoDelegationPlannerCandidateJSONDecodeReport {
	if in == nil {
		return nil
	}
	out := in.Clone()
	return &out
}

func cloneAutoDelegationPlanReviewPtr(in *AutoDelegationPlanReview) *AutoDelegationPlanReview {
	if in == nil {
		return nil
	}
	out := in.Clone()
	return &out
}

func cloneAutoDelegationParentMergePtr(in *AutoDelegationParentMerge) *AutoDelegationParentMerge {
	if in == nil {
		return nil
	}
	out := in.Normalize()
	return &out
}

func normalizeAutoDelegationPolicyReviewPtr(in *AutoDelegationPolicyReview) *AutoDelegationPolicyReview {
	if in == nil {
		return nil
	}
	out := in.Normalize()
	return &out
}

func normalizeAutoDelegationPlannerInstructionProfilePtr(in *AutoDelegationPlannerInstructionProfile) *AutoDelegationPlannerInstructionProfile {
	if in == nil {
		return nil
	}
	out := in.Normalize()
	return &out
}

func normalizeAutoDelegationPlannerCandidateJSONDecodeReportPtr(in *AutoDelegationPlannerCandidateJSONDecodeReport) *AutoDelegationPlannerCandidateJSONDecodeReport {
	if in == nil {
		return nil
	}
	out := in.Normalize()
	return &out
}

func normalizeAutoDelegationPlanReviewPtr(in *AutoDelegationPlanReview) *AutoDelegationPlanReview {
	if in == nil {
		return nil
	}
	out := in.Normalize()
	return &out
}

func normalizeAutoDelegationParentMergePtr(in *AutoDelegationParentMerge) *AutoDelegationParentMerge {
	if in == nil {
		return nil
	}
	out := in.Normalize()
	return &out
}

func autoDelegationPlanReviewEmpty(review AutoDelegationPlanReview) bool {
	return review.ContractVersion == "" &&
		!review.Projected &&
		review.Status == "" &&
		!review.Ready &&
		!review.PlanOnly &&
		!review.HostMayDispatch &&
		!review.WorkerAsToolDefault &&
		review.PolicyReview.ContractVersion == "" &&
		autoDelegationPlanEmpty(review.Plan) &&
		len(review.AcceptedChildRefs) == 0 &&
		len(review.RejectedChildRefs) == 0 &&
		len(review.MissingInputs) == 0 &&
		len(review.BlockedReasons) == 0 &&
		review.FailureClass == "" &&
		len(review.Boundaries) == 0 &&
		review.NextHostAction == "" &&
		review.RunnerEffect == "" &&
		review.PromptEffect == "" &&
		!review.RawOutputLoaded
}

func autoDelegationParentMergeEmpty(parentMerge AutoDelegationParentMerge) bool {
	return parentMerge.ContractVersion == "" &&
		!parentMerge.Projected &&
		parentMerge.Status == "" &&
		parentMerge.Decision == "" &&
		!parentMerge.ReadyForParentMerge &&
		!parentMerge.ParentAnswerMayUseChildEvidence &&
		parentMerge.HostBridge.ContractVersion == "" &&
		len(parentMerge.Children) == 0 &&
		parentMerge.FailureReview == nil &&
		parentMerge.ParentFrame.ID == "" &&
		parentMerge.ParentLedgerRef == "" &&
		len(parentMerge.RequiredEvidence) == 0 &&
		objectiveRuntimeLoopLedgerEmpty(parentMerge.ParentLedgerPatch) &&
		len(parentMerge.AcceptedChildRefs) == 0 &&
		len(parentMerge.MergedChildRefs) == 0 &&
		len(parentMerge.PartialChildRefs) == 0 &&
		len(parentMerge.RetryChildRefs) == 0 &&
		len(parentMerge.AlternatePathChildRefs) == 0 &&
		len(parentMerge.PrunedChildRefs) == 0 &&
		len(parentMerge.BlockedChildRefs) == 0 &&
		len(parentMerge.ConflictRefs) == 0 &&
		len(parentMerge.StaleEvidenceRefs) == 0 &&
		!parentMerge.StaleEvidenceDetected &&
		!parentMerge.WeakEvidenceDetected &&
		!parentMerge.MissingEvidenceDetected &&
		!parentMerge.ConflictDetected &&
		len(parentMerge.EvidenceRefs) == 0 &&
		len(parentMerge.Observations) == 0 &&
		len(parentMerge.MissingInputs) == 0 &&
		len(parentMerge.BlockedReasons) == 0 &&
		parentMerge.FailureClass == "" &&
		len(parentMerge.Boundaries) == 0 &&
		parentMerge.NextHostAction == "" &&
		parentMerge.RunnerEffect == "" &&
		parentMerge.PromptEffect == "" &&
		parentMerge.RuntimeEffect == "" &&
		!parentMerge.RawOutputLoaded
}

func stringsTrimmedNotEmpty(value string) bool {
	return strings.TrimSpace(value) != ""
}

func objectiveRuntimeLoopActivation(input ObjectiveRuntimeLoopInput, run ObjectiveRun, hasRun bool) Activation {
	if input.Activation != "" {
		return NormalizeActivation(string(input.Activation))
	}
	if hasRun {
		return run.Activation
	}
	return ActivationOff
}

func objectiveRuntimeLoopFrame(input ObjectiveRuntimeLoopInput, run ObjectiveRun, hasRun bool) ObjectiveFrame {
	frame := input.Frame.Normalize()
	if hasRun {
		base := run.Frame.Normalize()
		if frame.ID == "" {
			frame.ID = base.ID
		}
		if frame.UserGoalDigest == "" {
			frame.UserGoalDigest = base.UserGoalDigest
		}
		if frame.ControlMode == "" {
			frame.ControlMode = base.ControlMode
		}
		if frame.Intensity == "" {
			frame.Intensity = base.Intensity
		}
		if len(frame.SuccessCriteria) == 0 {
			frame.SuccessCriteria = base.SuccessCriteria
		}
		if len(frame.Constraints) == 0 {
			frame.Constraints = base.Constraints
		}
		frame.RequiredEvidence = MergeEvidenceRefs(base.RequiredEvidence, frame.RequiredEvidence)
		frame.CandidateCapabilities = normalizeDisplaySafeRefs(append(cloneDisplaySafeRefs(base.CandidateCapabilities), frame.CandidateCapabilities...))
		frame.SourceContext = normalizeDisplaySafeRefs(append(cloneDisplaySafeRefs(base.SourceContext), frame.SourceContext...))
		frame.Boundaries = MergeBoundaries(base.Boundaries, frame.Boundaries)
		frame.MissingInputs = MergeMissingInputs(base.MissingInputs, frame.MissingInputs)
	}
	return frame.Normalize()
}

func objectiveRuntimeLoopLedger(input ObjectiveRuntimeLoopInput, run ObjectiveRun, hasRun bool) AttemptLedgerPatch {
	if !objectiveRuntimeLoopLedgerEmpty(input.Ledger) {
		return input.Ledger.Normalize()
	}
	if hasRun {
		return run.Ledger.Normalize()
	}
	return AttemptLedgerPatch{}
}

func objectiveRuntimeLoopBudget(input ObjectiveRuntimeLoopInput, run ObjectiveRun, hasRun bool) ObjectiveBudgetSnapshot {
	budget := input.Budget.Normalize()
	if budget.BudgetRef == "" && hasRun {
		return run.Budget.Normalize()
	}
	return budget
}

func objectiveRuntimeLoopApproval(input ObjectiveRuntimeLoopInput, run ObjectiveRun, hasRun bool) ObjectiveApprovalState {
	approval := input.Approval.Normalize()
	if len(approval.ApprovalRefs) == 0 && len(approval.PolicyRefs) == 0 && len(approval.MissingInputs) == 0 && !approval.Required && !approval.Approved && hasRun {
		return run.Approval.Normalize()
	}
	return approval
}

func objectiveRuntimeLoopPolicyRefs(input ObjectiveRuntimeLoopInput, run ObjectiveRun, hasRun bool) []DisplaySafeRef {
	if hasRun {
		return normalizeDisplaySafeRefs(append(cloneDisplaySafeRefs(run.PolicyRefs), input.PolicyRefs...))
	}
	return normalizeDisplaySafeRefs(input.PolicyRefs)
}

func objectiveRuntimeLoopStrategies(input ObjectiveRuntimeLoopInput, run ObjectiveRun, hasRun bool) []StrategyCandidate {
	if hasRun {
		return normalizeStrategyCandidates(append(cloneStrategyCandidates(run.Strategies), input.Strategies...))
	}
	return normalizeStrategyCandidates(input.Strategies)
}

func objectiveRuntimeLoopCurrentStrategyRef(input ObjectiveRuntimeLoopInput, run ObjectiveRun, hasRun bool) DisplaySafeRef {
	if input.CurrentStrategyRef != "" {
		return normalizeOneDisplaySafeRef(input.CurrentStrategyRef)
	}
	if hasRun {
		return normalizeOneDisplaySafeRef(run.CurrentStrategyRef)
	}
	return ""
}

func objectiveRuntimeLoopVerification(input ObjectiveRuntimeLoopInput, run ObjectiveRun, hasRun bool) (ObjectiveVerificationGateResult, VerificationResult, bool) {
	if !objectiveRuntimeLoopVerificationGateEmpty(input.Verification) {
		gate := input.Verification.Normalize()
		return gate, gate.Verification.Normalize(), true
	}
	if !objectiveRuntimeLoopVerificationResultEmpty(input.VerificationResult) {
		verification := input.VerificationResult.Normalize()
		return ObjectiveVerificationGateResult{
			ContractVersion: ContractVersion,
			Projected:       true,
			Status:          verification.Status,
			Satisfied:       verification.Satisfied,
			Verification:    verification,
			EvidenceRefs:    verification.EvidenceRefs,
			FailureClass:    verification.FailureClass,
			FailureReason:   verification.FailureReason,
			MissingInputs:   verification.MissingInputs,
			Boundaries: AppendBoundaries(
				verification.Boundaries,
				"objective_runtime_loop_verification_result",
			),
			NextHostAction:  verification.NextHostAction,
			RunnerEffect:    "none",
			PromptEffect:    "none",
			RawOutputLoaded: verification.RawOutputLoaded,
		}.Normalize(), verification, true
	}
	if hasRun && !objectiveRuntimeLoopVerificationResultEmpty(run.LatestVerification) {
		verification := run.LatestVerification.Normalize()
		return ObjectiveVerificationGateResult{
			ContractVersion: ContractVersion,
			Projected:       true,
			Status:          verification.Status,
			Satisfied:       verification.Satisfied,
			Frame:           run.Frame,
			Observations:    run.LatestObservations,
			EvidenceRefs:    verification.EvidenceRefs,
			Verification:    verification,
			FailureClass:    verification.FailureClass,
			FailureReason:   verification.FailureReason,
			MissingInputs:   verification.MissingInputs,
			Boundaries: AppendBoundaries(
				verification.Boundaries,
				"objective_runtime_loop_existing_verification",
			),
			NextHostAction:  verification.NextHostAction,
			RunnerEffect:    "none",
			PromptEffect:    "none",
			RawOutputLoaded: verification.RawOutputLoaded,
		}.Normalize(), verification, true
	}
	return ObjectiveVerificationGateResult{}, VerificationResult{}, false
}

func objectiveRuntimeLoopMergeLedger(base, patch AttemptLedgerPatch) AttemptLedgerPatch {
	base = base.Normalize()
	patch = patch.Normalize()
	out := base.Clone()
	if out.ObjectiveID == "" {
		out.ObjectiveID = patch.ObjectiveID
	}
	if out.LedgerRef == "" {
		out.LedgerRef = patch.LedgerRef
	}
	out.Attempts = objectiveRuntimeLoopMergeAttemptSummaries(out.Attempts, patch.Attempts)
	if patch.RetryBudgetUsed > out.RetryBudgetUsed {
		out.RetryBudgetUsed = patch.RetryBudgetUsed
	}
	if patch.RetryBudgetRemaining > 0 || out.RetryBudgetRemaining == 0 {
		out.RetryBudgetRemaining = patch.RetryBudgetRemaining
	}
	out.EvidenceRefs = MergeEvidenceRefs(out.EvidenceRefs, patch.EvidenceRefs)
	out.Boundaries = AppendBoundaries(MergeBoundaries(out.Boundaries, patch.Boundaries), "objective_runtime_loop_ledger_patch_consumed")
	out.MissingInputs = MergeMissingInputs(out.MissingInputs, patch.MissingInputs)
	if patch.NextHostAction != "" {
		out.NextHostAction = patch.NextHostAction
	}
	out.RawOutputLoaded = out.RawOutputLoaded || patch.RawOutputLoaded
	return out.Normalize()
}

func objectiveRuntimeLoopMergeAttemptSummaries(base, patch []AttemptSummary) []AttemptSummary {
	out := normalizeAttemptSummaries(base)
	indexByRef := map[AttemptRef]int{}
	for i, attempt := range out {
		if attempt.Ref != "" {
			indexByRef[attempt.Ref] = i
		}
	}
	for _, attempt := range normalizeAttemptSummaries(patch) {
		if attempt.Ref != "" {
			if idx, exists := indexByRef[attempt.Ref]; exists {
				out[idx] = objectiveRuntimeLoopMergeAttemptSummary(out[idx], attempt)
				continue
			}
			indexByRef[attempt.Ref] = len(out)
		}
		out = append(out, attempt)
	}
	return normalizeAttemptSummaries(out)
}

func objectiveRuntimeLoopMergeAttemptSummary(base, patch AttemptSummary) AttemptSummary {
	out := base.Normalize()
	patch = patch.Normalize()
	if patch.ObjectiveID != "" {
		out.ObjectiveID = patch.ObjectiveID
	}
	if patch.StrategyID != "" {
		out.StrategyID = patch.StrategyID
	}
	if patch.Index != 0 {
		out.Index = patch.Index
	}
	if patch.ControlMode != "" {
		out.ControlMode = patch.ControlMode
	}
	if patch.Intensity != "" {
		out.Intensity = patch.Intensity
	}
	if patch.Status != VerificationNotEvaluated {
		out.Status = patch.Status
	}
	out.EvidenceRefs = MergeEvidenceRefs(out.EvidenceRefs, patch.EvidenceRefs)
	if patch.ObservationCount > out.ObservationCount {
		out.ObservationCount = patch.ObservationCount
	}
	out.FailureClass = firstFailureClass(patch.FailureClass, out.FailureClass)
	if patch.FailureReason != "" {
		out.FailureReason = patch.FailureReason
	}
	if patch.NextHostAction != "" {
		out.NextHostAction = patch.NextHostAction
	}
	out.Boundaries = MergeBoundaries(out.Boundaries, patch.Boundaries)
	out.MissingInputs = MergeMissingInputs(out.MissingInputs, patch.MissingInputs)
	out.RawOutputLoaded = out.RawOutputLoaded || patch.RawOutputLoaded
	return out.Normalize()
}

func objectiveRuntimeLoopReadyForControllerDecision(run ObjectiveRun, decision ObjectiveControllerDecision, hasVerification bool, rawOutputLoaded bool) bool {
	if rawOutputLoaded || !hasVerification {
		return false
	}
	if !run.FullRun || run.Activation != ActivationManaged {
		return false
	}
	switch decision.Action {
	case ObjectiveActionReturnSatisfied, ObjectiveActionRequestReplanDecision, ObjectiveActionReturnBlocked, ObjectiveActionReturnFailed, ObjectiveActionPlanStrategy:
		return true
	default:
		return false
	}
}

func objectiveRuntimeLoopReadyForNextRuntimeAction(decision ObjectiveControllerDecision) bool {
	switch decision.Action {
	case ObjectiveActionPlanStrategy, ObjectiveActionRequestReplanDecision:
		return true
	default:
		return false
	}
}

func objectiveRuntimeLoopStatus(step ObjectiveRuntimeLoopStep, run ObjectiveRun, decision ObjectiveControllerDecision, activation Activation, hasVerification bool) string {
	if activation != ActivationManaged {
		return "inactive"
	}
	if step.RawOutputLoaded || run.State == ObjectiveControllerReviewRequired {
		return "review_required"
	}
	if !hasVerification {
		return "blocked"
	}
	if step.ReadyForHostPersist {
		return "ready_for_host_persist"
	}
	if step.ReadyForControllerDecision {
		return "ready_for_controller_decision"
	}
	if decision.Action == ObjectiveActionProvideObjectiveContract ||
		decision.Action == ObjectiveActionProvideBudgetPolicy ||
		decision.Action == ObjectiveActionProvideStrategyScope ||
		decision.Action == ObjectiveActionRequestHostApproval ||
		decision.Action == ObjectiveActionProvideApprovalRef {
		return "blocked"
	}
	return string(run.State)
}

func objectiveRuntimeLoopLedgerEmpty(ledger AttemptLedgerPatch) bool {
	ledger = ledger.Normalize()
	return ledger.ObjectiveID == "" &&
		ledger.LedgerRef == "" &&
		len(ledger.Attempts) == 0 &&
		len(ledger.EvidenceRefs) == 0 &&
		len(ledger.MissingInputs) == 0 &&
		len(ledger.Boundaries) == 0
}

func objectiveRuntimeLoopVerificationGateEmpty(gate ObjectiveVerificationGateResult) bool {
	return gate.ContractVersion == "" &&
		!gate.Projected &&
		gate.Status == "" &&
		!gate.Satisfied &&
		gate.Frame.ID == "" &&
		gate.Normalization.Status == "" &&
		len(gate.Requirements) == 0 &&
		len(gate.Observations) == 0 &&
		len(gate.EvidenceRefs) == 0 &&
		gate.Verification.Status == "" &&
		len(gate.MissingInputs) == 0 &&
		len(gate.Boundaries) == 0 &&
		!gate.RawOutputLoaded
}

func objectiveRuntimeLoopVerificationResultEmpty(result VerificationResult) bool {
	return result.ContractVersion == "" &&
		result.Status == "" &&
		!result.Satisfied &&
		result.FailureClass == "" &&
		result.FailureReason == "" &&
		len(result.EvidenceRefs) == 0 &&
		len(result.MissingInputs) == 0 &&
		len(result.Boundaries) == 0 &&
		len(result.Findings) == 0 &&
		result.NextHostAction == "" &&
		!result.RawOutputLoaded
}
