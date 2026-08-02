package hostkit

import (
	"strings"

	agentxcontrolplane "github.com/wsnacj/agentx-go/runtime/session"
)

type DelegationWorkerObjectiveRuntimeHandoffInput struct {
	Run                      agentxcontrolplane.ObjectiveRun
	Invocation               agentxcontrolplane.HostOwnedDelegationWorkerRuntimeInvocation
	ParentFrame              agentxcontrolplane.ObjectiveFrame
	ParentLedgerRef          agentxcontrolplane.DisplaySafeRef
	WorkerAttemptRef         agentxcontrolplane.AttemptRef
	MergeRef                 agentxcontrolplane.DisplaySafeRef
	MergePolicyRef           agentxcontrolplane.DisplaySafeRef
	HandoffRef               agentxcontrolplane.DisplaySafeRef
	WorkerObservations       []agentxcontrolplane.Observation
	RequiredEvidence         []agentxcontrolplane.EvidenceRef
	ExpectedObservationKinds []string
	EvidenceRefs             []agentxcontrolplane.EvidenceRef
	DecisionBasis            []agentxcontrolplane.DisplaySafeRef
	Boundaries               []agentxcontrolplane.Boundary
	RawOutputLoaded          bool
}

type DelegationWorkerObjectiveRuntimeHandoffReport struct {
	Available                        bool     `json:"available"`
	Status                           string   `json:"status,omitempty"`
	ParentMergeStatus                string   `json:"parent_merge_status,omitempty"`
	RuntimeHandoffStatus             string   `json:"runtime_handoff_status,omitempty"`
	RuntimeLoopStatus                string   `json:"runtime_loop_status,omitempty"`
	RuntimeLoopState                 string   `json:"runtime_loop_state,omitempty"`
	RuntimeLoopAction                string   `json:"runtime_loop_action,omitempty"`
	ReadyForParentMerge              bool     `json:"ready_for_parent_merge"`
	ReadyForRuntimeLoopInput         bool     `json:"ready_for_runtime_loop_input"`
	RuntimeLoopReady                 bool     `json:"runtime_loop_ready"`
	RuntimeLoopHostPersistReady      bool     `json:"runtime_loop_host_persist_ready"`
	ReadyForNextRuntimeAction        bool     `json:"ready_for_next_runtime_action"`
	ReadyForFailureReview            bool     `json:"ready_for_failure_review"`
	WorkerResultRequiresVerification bool     `json:"worker_result_requires_verification"`
	WorkerOutputAcceptedAsFact       bool     `json:"worker_output_accepted_as_fact"`
	WorkerDispatchedByCore           bool     `json:"worker_dispatched_by_core"`
	RunnerDispatchedByCore           bool     `json:"runner_dispatched_by_core"`
	RuntimeAdapterExecutedByCore     bool     `json:"runtime_adapter_executed_by_core"`
	StoreMutationExecutedByCore      bool     `json:"store_mutation_executed_by_core"`
	CoreExecutionExecuted            bool     `json:"core_execution_executed"`
	RawOutputLoaded                  bool     `json:"raw_output_loaded"`
	ParentLedgerRef                  string   `json:"parent_ledger_ref,omitempty"`
	WorkerAttemptRef                 string   `json:"worker_attempt_ref,omitempty"`
	MergeRef                         string   `json:"merge_ref,omitempty"`
	HandoffRef                       string   `json:"handoff_ref,omitempty"`
	WorkerRunRef                     string   `json:"worker_run_ref,omitempty"`
	WorkerResultRef                  string   `json:"worker_result_ref,omitempty"`
	WorkerReadbackRef                string   `json:"worker_readback_ref,omitempty"`
	ObservationRef                   string   `json:"observation_ref,omitempty"`
	EvidenceRefs                     []string `json:"evidence_refs,omitempty"`
	MissingInputs                    []string `json:"missing_inputs,omitempty"`
	BlockedReasons                   []string `json:"blocked_reasons,omitempty"`
	Boundaries                       []string `json:"boundaries,omitempty"`
	NextHostAction                   string   `json:"next_host_action,omitempty"`
	ParentMerge                      agentxcontrolplane.DelegationWorkerParentMerge
	RuntimeHandoff                   agentxcontrolplane.DelegationObjectiveRuntimeHandoff
	RuntimeLoopStep                  agentxcontrolplane.ObjectiveRuntimeLoopStep
}

func BuildDelegationWorkerObjectiveRuntimeHandoff(input DelegationWorkerObjectiveRuntimeHandoffInput) DelegationWorkerObjectiveRuntimeHandoffReport {
	run := input.Run.Normalize()
	invocation := input.Invocation.Normalize()
	parentFrame := input.ParentFrame.Normalize()
	if parentFrame.ID == "" {
		parentFrame = run.Frame.Normalize()
	}
	parentLedgerRef := delegationWorkerObjectiveFirstDisplaySafeRef(input.ParentLedgerRef, run.Ledger.LedgerRef)
	requiredEvidence := agentxcontrolplane.MergeEvidenceRefs(input.RequiredEvidence, parentFrame.RequiredEvidence, run.Frame.RequiredEvidence)
	merge := agentxcontrolplane.BuildDelegationWorkerParentMerge(agentxcontrolplane.DelegationWorkerParentMergeInput{
		Invocation:               invocation,
		ParentFrame:              parentFrame,
		ParentLedgerRef:          parentLedgerRef,
		WorkerAttemptRef:         input.WorkerAttemptRef,
		MergeRef:                 input.MergeRef,
		MergePolicyRef:           input.MergePolicyRef,
		WorkerObservations:       input.WorkerObservations,
		EvidenceRefs:             input.EvidenceRefs,
		RequiredEvidence:         requiredEvidence,
		ExpectedObservationKinds: input.ExpectedObservationKinds,
		Boundaries: agentxcontrolplane.AppendBoundaries(input.Boundaries,
			"hostruntime_delegation_worker_parent_merge",
			"hostruntime_delegation_worker_objective_handoff",
		),
		RawOutputLoaded: input.RawOutputLoaded,
	})
	handoff := agentxcontrolplane.BuildDelegationObjectiveRuntimeHandoff(agentxcontrolplane.DelegationObjectiveRuntimeHandoffInput{
		HandoffRef:    input.HandoffRef,
		Run:           run,
		ParentMerge:   merge,
		EvidenceRefs:  input.EvidenceRefs,
		DecisionBasis: input.DecisionBasis,
		Boundaries: agentxcontrolplane.AppendBoundaries(input.Boundaries,
			"hostruntime_delegation_worker_objective_handoff",
		),
		RawOutputLoaded: input.RawOutputLoaded,
	})
	var step agentxcontrolplane.ObjectiveRuntimeLoopStep
	if handoff.ReadyForRuntimeLoopInput {
		step = agentxcontrolplane.BuildObjectiveRuntimeLoopStep(handoff.RuntimeLoopInput)
	}
	report := DelegationWorkerObjectiveRuntimeHandoffReport{
		Available:                        true,
		Status:                           "blocked",
		ParentMergeStatus:                string(merge.Status),
		RuntimeHandoffStatus:             string(handoff.Status),
		RuntimeLoopStatus:                strings.TrimSpace(step.Status),
		RuntimeLoopState:                 string(step.Run.State),
		RuntimeLoopAction:                string(step.ControllerDecision.Action),
		ReadyForParentMerge:              merge.ReadyForParentMerge,
		ReadyForRuntimeLoopInput:         handoff.ReadyForRuntimeLoopInput,
		RuntimeLoopReady:                 step.ReadyForControllerDecision,
		RuntimeLoopHostPersistReady:      step.ReadyForHostPersist,
		ReadyForNextRuntimeAction:        step.ReadyForNextRuntimeAction,
		ReadyForFailureReview:            handoff.ReadyForFailureReview,
		WorkerResultRequiresVerification: true,
		WorkerOutputAcceptedAsFact:       false,
		ParentLedgerRef:                  string(merge.ParentLedgerRef),
		WorkerAttemptRef:                 string(merge.WorkerAttemptRef),
		MergeRef:                         string(merge.MergeRef),
		HandoffRef:                       string(handoff.HandoffRef),
		WorkerRunRef:                     string(invocation.WorkerRunRef),
		WorkerResultRef:                  string(invocation.WorkerResultRef),
		WorkerReadbackRef:                string(invocation.WorkerReadbackRef),
		ObservationRef:                   string(invocation.ObservationRef),
		EvidenceRefs:                     evidenceRefsToStrings(agentxcontrolplane.MergeEvidenceRefs(input.EvidenceRefs, merge.EvidenceRefs, handoff.EvidenceRefs, step.EvidenceRefs)),
		MissingInputs:                    appendUniqueStrings(missingInputsToStrings(merge.MissingInputs), missingInputsToStrings(handoff.MissingInputs)...),
		BlockedReasons:                   appendUniqueStrings(append([]string(nil), merge.BlockedReasons...), handoff.BlockedReasons...),
		Boundaries: appendUniqueStrings(
			appendUniqueStrings(boundariesToStrings(merge.Boundaries), boundariesToStrings(handoff.Boundaries)...),
			boundariesToStrings(step.Boundaries)...,
		),
		NextHostAction:  firstNonEmpty(string(step.NextHostAction), string(handoff.NextHostAction), string(merge.NextHostAction)),
		RawOutputLoaded: input.RawOutputLoaded || merge.RawOutputLoaded || handoff.RawOutputLoaded || step.RawOutputLoaded,
		ParentMerge:     merge,
		RuntimeHandoff:  handoff,
		RuntimeLoopStep: step,
	}
	if report.ReadyForRuntimeLoopInput && report.RuntimeLoopHostPersistReady {
		report.Status = "ready_for_host_persist"
		report.NextHostAction = "persist_objective_run"
		report.Boundaries = appendUniqueStrings(report.Boundaries,
			"delegation_worker_parent_merge_ready_for_runtime_loop",
			"delegation_worker_runtime_loop_ready_for_host_persist",
		)
	} else if report.ReadyForRuntimeLoopInput {
		report.Status = firstNonEmpty(report.RuntimeLoopStatus, "runtime_loop_projected")
	} else if report.ReadyForFailureReview {
		report.Status = "failure_review_ready"
	}
	return report.Normalize()
}

func (report DelegationWorkerObjectiveRuntimeHandoffReport) Normalize() DelegationWorkerObjectiveRuntimeHandoffReport {
	out := report
	out.Status = strings.TrimSpace(out.Status)
	out.ParentMergeStatus = strings.TrimSpace(out.ParentMergeStatus)
	out.RuntimeHandoffStatus = strings.TrimSpace(out.RuntimeHandoffStatus)
	out.RuntimeLoopStatus = strings.TrimSpace(out.RuntimeLoopStatus)
	out.RuntimeLoopState = strings.TrimSpace(out.RuntimeLoopState)
	out.RuntimeLoopAction = strings.TrimSpace(out.RuntimeLoopAction)
	out.ParentMerge = out.ParentMerge.Normalize()
	out.RuntimeHandoff = out.RuntimeHandoff.Normalize()
	out.RuntimeLoopStep = out.RuntimeLoopStep.Normalize()
	for _, ref := range []*string{
		&out.ParentLedgerRef,
		&out.MergeRef,
		&out.HandoffRef,
		&out.WorkerRunRef,
		&out.WorkerResultRef,
		&out.WorkerReadbackRef,
		&out.ObservationRef,
	} {
		delegationWorkerObjectiveNormalizeDisplayRef(ref, &out.RawOutputLoaded)
	}
	if strings.TrimSpace(out.WorkerAttemptRef) != "" {
		if ref, ok := agentxcontrolplane.NormalizeAttemptRef(out.WorkerAttemptRef); ok {
			out.WorkerAttemptRef = string(ref)
		} else {
			out.WorkerAttemptRef = ""
			out.RawOutputLoaded = true
		}
	}
	out.EvidenceRefs = appendUniqueStrings(nil, out.EvidenceRefs...)
	out.MissingInputs = appendUniqueStrings(nil, out.MissingInputs...)
	out.BlockedReasons = appendUniqueStrings(nil, out.BlockedReasons...)
	out.Boundaries = appendUniqueStrings(nil, out.Boundaries...)
	out.NextHostAction = strings.TrimSpace(out.NextHostAction)
	out.WorkerResultRequiresVerification = true
	out.WorkerOutputAcceptedAsFact = false
	out.WorkerDispatchedByCore = false
	out.RunnerDispatchedByCore = false
	out.RuntimeAdapterExecutedByCore = false
	out.StoreMutationExecutedByCore = false
	out.CoreExecutionExecuted = false
	if out.RawOutputLoaded {
		out.Status = "review_required"
		out.ReadyForParentMerge = false
		out.ReadyForRuntimeLoopInput = false
		out.RuntimeLoopReady = false
		out.RuntimeLoopHostPersistReady = false
		out.ReadyForNextRuntimeAction = false
		out.MissingInputs = appendUniqueStrings(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueStrings(out.BlockedReasons, "unsafe_input_ref")
		out.Boundaries = appendUniqueStrings(out.Boundaries, "raw_output_not_allowed")
		out.NextHostAction = "provide_display_safe_refs"
	}
	return out
}

func evidenceRefsToStrings(refs []agentxcontrolplane.EvidenceRef) []string {
	out := make([]string, 0, len(refs))
	for _, ref := range agentxcontrolplane.MergeEvidenceRefs(refs) {
		if safe := delegationWorkerObjectiveSafeRef(ref.Ref); safe != "" {
			out = append(out, string(safe))
		}
	}
	return out
}

func delegationWorkerObjectiveFirstDisplaySafeRef(values ...agentxcontrolplane.DisplaySafeRef) agentxcontrolplane.DisplaySafeRef {
	for _, value := range values {
		if safe := delegationWorkerObjectiveSafeRef(value); safe != "" {
			return safe
		}
	}
	return ""
}

func delegationWorkerObjectiveNormalizeDisplayRef(value *string, raw *bool) {
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		*value = ""
		return
	}
	ref, ok := agentxcontrolplane.NormalizeDisplaySafeRef(trimmed)
	if !ok {
		*value = ""
		*raw = true
		return
	}
	*value = string(ref)
}

func delegationWorkerObjectiveSafeRef(ref agentxcontrolplane.DisplaySafeRef) agentxcontrolplane.DisplaySafeRef {
	raw := strings.TrimSpace(string(ref))
	if raw == "" {
		return ""
	}
	normalized, ok := agentxcontrolplane.NormalizeDisplaySafeRef(raw)
	if !ok {
		return ""
	}
	return normalized
}
