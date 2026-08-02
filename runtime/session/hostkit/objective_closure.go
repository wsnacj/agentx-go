package hostkit

import (
	"context"
	"strings"

	agentxcontrolplane "github.com/wsnacj/agentx-go/runtime/session"
)

// ObjectiveRuntimeClosureInput carries the parent objective facts needed after
// a host-owned delegation worker has executed and read back its result.
type ObjectiveRuntimeClosureInput struct {
	Run                      agentxcontrolplane.ObjectiveRun
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

// ObjectiveRuntimeClosureProfile composes a host-owned delegation worker
// backend with the reusable objective-runtime handoff. It does not choose a
// worker, execute from core, or persist the ObjectiveRun; those remain host
// responsibilities.
type ObjectiveRuntimeClosureProfile struct {
	Enabled      bool
	Backend      Backend
	BackendInput BackendInput
	ClosureInput ObjectiveRuntimeClosureInput
	Boundaries   []agentxcontrolplane.Boundary
}

type ObjectiveRuntimeClosureProfileReport struct {
	Available                        bool                                                       `json:"available"`
	Enabled                          bool                                                       `json:"enabled"`
	Ready                            bool                                                       `json:"ready"`
	Status                           string                                                     `json:"status,omitempty"`
	BackendReady                     bool                                                       `json:"backend_ready"`
	RuntimeLoopReady                 bool                                                       `json:"runtime_loop_ready"`
	RuntimeLoopHostPersistReady      bool                                                       `json:"runtime_loop_host_persist_ready"`
	HostWorkerRuntimeExecuted        bool                                                       `json:"host_worker_runtime_executed"`
	WorkerResultRecorded             bool                                                       `json:"worker_result_recorded"`
	WorkerResultReadbackReady        bool                                                       `json:"worker_result_readback_ready"`
	ReadyForRuntimeLoopInput         bool                                                       `json:"ready_for_runtime_loop_input"`
	ReadyForFailureReview            bool                                                       `json:"ready_for_failure_review"`
	WorkerResultRequiresVerification bool                                                       `json:"worker_result_requires_verification"`
	WorkerOutputAcceptedAsFact       bool                                                       `json:"worker_output_accepted_as_fact"`
	WorkerDispatchedByCore           bool                                                       `json:"worker_dispatched_by_core"`
	RunnerDispatchedByCore           bool                                                       `json:"runner_dispatched_by_core"`
	RuntimeAdapterExecutedByCore     bool                                                       `json:"runtime_adapter_executed_by_core"`
	StoreMutationExecutedByCore      bool                                                       `json:"store_mutation_executed_by_core"`
	CoreExecutionExecuted            bool                                                       `json:"core_execution_executed"`
	Backend                          BackendReport                                              `json:"backend,omitempty"`
	Closure                          DelegationWorkerObjectiveRuntimeHandoffReport              `json:"closure,omitempty"`
	AsyncCompletion                  agentxcontrolplane.AutoDelegationAsyncCompletionProjection `json:"async_completion,omitempty"`
	MissingInputs                    []string                                                   `json:"missing_inputs,omitempty"`
	BlockedReasons                   []string                                                   `json:"blocked_reasons,omitempty"`
	Boundaries                       []string                                                   `json:"boundaries,omitempty"`
	NextHostAction                   string                                                     `json:"next_host_action,omitempty"`
}

func RunObjectiveRuntimeClosureProfile(ctx context.Context, profile ObjectiveRuntimeClosureProfile) (ObjectiveRuntimeClosureProfileReport, error) {
	report := ObjectiveRuntimeClosureProfileReport{
		Enabled:                          profile.Enabled,
		Status:                           "blocked",
		WorkerResultRequiresVerification: true,
		WorkerOutputAcceptedAsFact:       false,
		Boundaries: appendUniqueStrings(boundariesToStrings(profile.Boundaries),
			"delegation_worker_objective_runtime_closure_profile",
			"host_owned_delegation_worker_runtime_profile",
			"delegation_worker_result_requires_parent_verification",
			"worker_output_not_fact",
			"display_safe_refs_only",
			"no_delegation_worker_runtime_by_core",
			"no_runner_dispatch_by_core",
			"no_runtime_adapter_execution_by_core",
			"no_store_mutation_by_core",
		),
		NextHostAction: "review_delegation_worker_objective_runtime_closure_profile",
	}
	if !profile.Enabled {
		report.Status = "disabled"
		report.MissingInputs = appendUniqueStrings(report.MissingInputs, "host:delegation_worker_objective_runtime_closure_profile_enabled")
		report.BlockedReasons = appendUniqueStrings(report.BlockedReasons, "delegation_worker_objective_runtime_closure_profile_disabled")
		report.Boundaries = appendUniqueStrings(report.Boundaries, "delegation_worker_objective_runtime_closure_profile_default_off")
		report.NextHostAction = "enable_delegation_worker_objective_runtime_closure_profile"
		return report.Normalize(), nil
	}
	if ctx == nil {
		ctx = context.Background()
	}

	backendInput := profile.BackendInput
	backendInput.Enabled = true
	backendInput.Boundaries = agentxcontrolplane.AppendBoundaries(backendInput.Boundaries,
		"profile_delegation_worker_runtime_backend",
	)
	backendReport, err := profile.Backend.RunDelegationWorkerRuntime(ctx, backendInput)
	if err != nil {
		return report.Normalize(), err
	}
	report = objectiveRuntimeClosureProfileBindBackend(report, backendReport)
	report = objectiveRuntimeClosureProfileBindAsyncCompletion(report, profile, backendReport)
	if !backendReport.WorkerResultReadbackReady || !backendReport.Invocation.ReadyForWorkerResultReview {
		report.Status = "delegation_worker_objective_runtime_closure_backend_blocked"
		report.BlockedReasons = appendUniqueStrings(report.BlockedReasons, "delegation_worker_runtime_backend_not_ready_for_objective_closure")
		report.NextHostAction = firstNonEmpty(string(backendReport.NextHostAction), "complete_delegation_worker_runtime_backend")
		return report.Normalize(), nil
	}

	closureInput := profile.ClosureInput
	closureInput.EvidenceRefs = agentxcontrolplane.MergeEvidenceRefs(closureInput.EvidenceRefs, backendReport.EvidenceRefs)
	closureInput.Boundaries = agentxcontrolplane.AppendBoundaries(closureInput.Boundaries,
		"profile_delegation_worker_objective_runtime_handoff",
	)
	closureReport := BuildDelegationWorkerObjectiveRuntimeHandoff(DelegationWorkerObjectiveRuntimeHandoffInput{
		Run:                      closureInput.Run,
		Invocation:               backendReport.Invocation,
		ParentFrame:              closureInput.ParentFrame,
		ParentLedgerRef:          closureInput.ParentLedgerRef,
		WorkerAttemptRef:         closureInput.WorkerAttemptRef,
		MergeRef:                 closureInput.MergeRef,
		MergePolicyRef:           closureInput.MergePolicyRef,
		HandoffRef:               closureInput.HandoffRef,
		WorkerObservations:       closureInput.WorkerObservations,
		RequiredEvidence:         closureInput.RequiredEvidence,
		ExpectedObservationKinds: closureInput.ExpectedObservationKinds,
		EvidenceRefs:             closureInput.EvidenceRefs,
		DecisionBasis:            closureInput.DecisionBasis,
		Boundaries:               closureInput.Boundaries,
		RawOutputLoaded:          closureInput.RawOutputLoaded,
	})
	report = objectiveRuntimeClosureProfileBindClosure(report, closureReport)
	if closureReport.ReadyForRuntimeLoopInput && closureReport.RuntimeLoopHostPersistReady {
		report.Available = true
		report.Ready = true
		report.Status = "delegation_worker_objective_runtime_closure_ready_for_host_persist"
		report.NextHostAction = "persist_objective_run"
		report.Boundaries = appendUniqueStrings(report.Boundaries,
			"delegation_worker_objective_runtime_closure_ready",
			"delegation_worker_objective_runtime_loop_ready_for_host_persist",
		)
		return report.Normalize(), nil
	}
	if closureReport.ReadyForFailureReview {
		report.Status = "delegation_worker_objective_runtime_closure_failure_review_ready"
		report.NextHostAction = firstNonEmpty(closureReport.NextHostAction, report.NextHostAction)
		return report.Normalize(), nil
	}
	report.Status = firstNonEmpty(closureReport.Status, "delegation_worker_objective_runtime_closure_blocked")
	report.NextHostAction = firstNonEmpty(closureReport.NextHostAction, report.NextHostAction)
	return report.Normalize(), nil
}

func (report ObjectiveRuntimeClosureProfileReport) Normalize() ObjectiveRuntimeClosureProfileReport {
	out := report
	out.Status = strings.TrimSpace(out.Status)
	out.NextHostAction = strings.TrimSpace(out.NextHostAction)
	out.Backend = out.Backend.Normalize()
	out.Closure = out.Closure.Normalize()
	if objectiveRuntimeClosureProfileAsyncCompletionPresent(out.AsyncCompletion) {
		out.AsyncCompletion = out.AsyncCompletion.Normalize()
	}
	out.MissingInputs = appendUniqueStrings(nil, out.MissingInputs...)
	out.BlockedReasons = appendUniqueStrings(nil, out.BlockedReasons...)
	out.Boundaries = appendUniqueStrings(nil, out.Boundaries...)
	out.WorkerResultRequiresVerification = true
	out.WorkerOutputAcceptedAsFact = false
	out.WorkerDispatchedByCore = false
	out.RunnerDispatchedByCore = false
	out.RuntimeAdapterExecutedByCore = false
	out.StoreMutationExecutedByCore = false
	out.CoreExecutionExecuted = false
	if out.Status == "" {
		out.Status = "blocked"
	}
	if len(out.MissingInputs) > 0 || len(out.BlockedReasons) > 0 {
		out.Ready = false
	}
	if out.NextHostAction == "" {
		out.NextHostAction = "review_delegation_worker_objective_runtime_closure_profile"
	}
	return out
}

func objectiveRuntimeClosureProfileBindBackend(report ObjectiveRuntimeClosureProfileReport, backend BackendReport) ObjectiveRuntimeClosureProfileReport {
	backend = backend.Normalize()
	report.Backend = backend
	report.Available = backend.Available
	report.BackendReady = backend.WorkerResultReadbackReady && backend.Invocation.ReadyForWorkerResultReview
	report.HostWorkerRuntimeExecuted = backend.WorkerRunAttempted && backend.WorkerReadbackAttempted
	report.WorkerResultRecorded = backend.WorkerResultRecorded
	report.WorkerResultReadbackReady = backend.WorkerResultReadbackReady
	report.ReadyForFailureReview = backend.ReadyForFailureReview
	report.MissingInputs = appendUniqueStrings(report.MissingInputs, missingInputsToStrings(backend.MissingInputs)...)
	report.BlockedReasons = appendUniqueStrings(report.BlockedReasons, backend.BlockedReasons...)
	report.Boundaries = appendUniqueStrings(report.Boundaries, boundariesToStrings(backend.Boundaries)...)
	report.NextHostAction = firstNonEmpty(string(backend.NextHostAction), report.NextHostAction)
	return report
}

func objectiveRuntimeClosureProfileBindClosure(report ObjectiveRuntimeClosureProfileReport, closure DelegationWorkerObjectiveRuntimeHandoffReport) ObjectiveRuntimeClosureProfileReport {
	closure = closure.Normalize()
	report.Closure = closure
	report.RuntimeLoopReady = closure.RuntimeLoopReady
	report.RuntimeLoopHostPersistReady = closure.RuntimeLoopHostPersistReady
	report.ReadyForRuntimeLoopInput = closure.ReadyForRuntimeLoopInput
	report.ReadyForFailureReview = report.ReadyForFailureReview || closure.ReadyForFailureReview
	report.MissingInputs = appendUniqueStrings(report.MissingInputs, closure.MissingInputs...)
	report.BlockedReasons = appendUniqueStrings(report.BlockedReasons, closure.BlockedReasons...)
	report.Boundaries = appendUniqueStrings(report.Boundaries, closure.Boundaries...)
	report.NextHostAction = firstNonEmpty(closure.NextHostAction, report.NextHostAction)
	return report
}

func objectiveRuntimeClosureProfileBindAsyncCompletion(report ObjectiveRuntimeClosureProfileReport, profile ObjectiveRuntimeClosureProfile, backend BackendReport) ObjectiveRuntimeClosureProfileReport {
	backend = backend.Normalize()
	if !objectiveRuntimeClosureProfileBackendHasAsyncSurface(backend) {
		return report
	}
	backendKind := agentxcontrolplane.AutoDelegationAsyncBackendProcessLocal
	if profile.Backend.Durable {
		backendKind = agentxcontrolplane.AutoDelegationAsyncBackendDurable
	}
	projection := BuildDelegationWorkerAsyncCompletionProjection(DelegationWorkerAsyncCompletionInput{
		Run:             profile.ClosureInput.Run,
		BackendKind:     backendKind,
		BackendRef:      backend.BackendRef,
		RequireDurable:  profile.Backend.Durable,
		Invocation:      backend.Invocation,
		EvidenceRefs:    backend.EvidenceRefs,
		MissingInputs:   backend.MissingInputs,
		BlockedReasons:  backend.BlockedReasons,
		FailureClass:    backend.FailureClass,
		DecisionBasis:   []agentxcontrolplane.DisplaySafeRef{"delegationruntime:objective_runtime_closure_profile_async_completion"},
		Boundaries:      []agentxcontrolplane.Boundary{"profile_delegation_worker_async_completion"},
		RawOutputLoaded: backend.RawOutputLoaded,
	})
	report.AsyncCompletion = projection
	report.Boundaries = appendUniqueStrings(report.Boundaries, boundariesToStrings(projection.Boundaries)...)
	if projection.RawOutputLoaded {
		report.MissingInputs = appendUniqueStrings(report.MissingInputs, "host:display_safe_refs")
		report.BlockedReasons = appendUniqueStrings(report.BlockedReasons, "unsafe_input_ref")
	}
	return report
}

func objectiveRuntimeClosureProfileBackendHasAsyncSurface(backend BackendReport) bool {
	return backend.Invocation.Projected ||
		backend.WorkerRunAttempted ||
		backend.WorkerResultRecorded ||
		backend.WorkerReadbackAttempted ||
		backend.WorkerResultReadbackReady
}

func objectiveRuntimeClosureProfileAsyncCompletionPresent(projection agentxcontrolplane.AutoDelegationAsyncCompletionProjection) bool {
	return projection.ContractVersion != "" ||
		projection.Status != "" ||
		projection.BackendKind != "" ||
		projection.BackendRef != "" ||
		len(projection.Children) > 0 ||
		len(projection.ActiveChildRefs) > 0 ||
		len(projection.CompletedChildRefs) > 0 ||
		len(projection.ResumeRequest.ChildRefs) > 0
}

func missingInputsToStrings(values []agentxcontrolplane.MissingInput) []string {
	out := make([]string, 0, len(values))
	for _, value := range agentxcontrolplane.AppendMissingInputs(nil, values...) {
		if trimmed := strings.TrimSpace(string(value)); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return appendUniqueStrings(nil, out...)
}

func boundariesToStrings(values []agentxcontrolplane.Boundary) []string {
	out := make([]string, 0, len(values))
	for _, value := range agentxcontrolplane.AppendBoundaries(nil, values...) {
		if trimmed := strings.TrimSpace(string(value)); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return appendUniqueStrings(nil, out...)
}
