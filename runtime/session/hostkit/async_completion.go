package hostkit

import (
	"strings"

	agentxcontrolplane "github.com/wsnacj/agentx-go/runtime/session"
)

type DelegationWorkerAsyncCompletionInput struct {
	Run                   agentxcontrolplane.ObjectiveRun
	HostBridge            agentxcontrolplane.AutoDelegationHostBridge
	BackendKind           agentxcontrolplane.AutoDelegationAsyncBackendKind
	BackendRef            agentxcontrolplane.DisplaySafeRef
	ParentObjectiveRef    agentxcontrolplane.DisplaySafeRef
	ParentObjectiveRunRef agentxcontrolplane.DisplaySafeRef
	ParentLedgerRef       agentxcontrolplane.DisplaySafeRef
	ParentResumeRef       agentxcontrolplane.DisplaySafeRef
	RequireDurable        bool
	ChildRef              agentxcontrolplane.DisplaySafeRef
	Role                  agentxcontrolplane.AutoDelegationChildRole
	Status                agentxcontrolplane.AutoDelegationAsyncChildStatus
	AgeSeconds            int
	CurrentAction         string
	CapabilityRefs        []agentxcontrolplane.DisplaySafeRef
	AllowedToolRefs       []agentxcontrolplane.DisplaySafeRef
	BoundCapabilityRefs   []agentxcontrolplane.DisplaySafeRef
	BoundAllowedToolRefs  []agentxcontrolplane.DisplaySafeRef
	Invocation            agentxcontrolplane.HostOwnedDelegationWorkerRuntimeInvocation
	ObservationRef        agentxcontrolplane.DisplaySafeRef
	FailureRef            agentxcontrolplane.DisplaySafeRef
	FailureReviewRef      agentxcontrolplane.DisplaySafeRef
	CancellationRef       agentxcontrolplane.DisplaySafeRef
	InterruptionRef       agentxcontrolplane.DisplaySafeRef
	CompletionEnvelopeRef agentxcontrolplane.DisplaySafeRef
	EvidenceRefs          []agentxcontrolplane.EvidenceRef
	ReadyForFailureReview bool
	MissingInputs         []agentxcontrolplane.MissingInput
	BlockedReasons        []string
	FailureClass          agentxcontrolplane.FailureClass
	DecisionBasis         []agentxcontrolplane.DisplaySafeRef
	Boundaries            []agentxcontrolplane.Boundary
	RawOutputLoaded       bool
}

func BuildDelegationWorkerAsyncCompletionProjection(input DelegationWorkerAsyncCompletionInput) agentxcontrolplane.AutoDelegationAsyncCompletionProjection {
	run := input.Run.Normalize()
	invocation := input.Invocation.Normalize()
	parentObjectiveRef := delegationWorkerAsyncFirstRef(input.ParentObjectiveRef, agentxcontrolplane.DisplaySafeRef(run.Frame.ID))
	parentObjectiveRunRef := delegationWorkerAsyncFirstRef(input.ParentObjectiveRunRef, delegationWorkerAsyncDerivedRef("objective_run", parentObjectiveRef))
	parentLedgerRef := delegationWorkerAsyncFirstRef(input.ParentLedgerRef, run.Ledger.LedgerRef)
	parentResumeRef := delegationWorkerAsyncFirstRef(input.ParentResumeRef, delegationWorkerAsyncDerivedRef("resume", parentObjectiveRef))
	childRef := delegationWorkerAsyncFirstRef(input.ChildRef, invocation.SubgoalRef, invocation.Readiness.SubgoalRef, invocation.WorkerRunRef)
	status := delegationWorkerAsyncStatus(input.Status, invocation)
	capabilityRefs := delegationWorkerAsyncRefs(input.CapabilityRefs, []agentxcontrolplane.DisplaySafeRef{invocation.Readiness.AdapterCapabilityRef, invocation.WorkerRef})
	allowedToolRefs := delegationWorkerAsyncRefs(input.AllowedToolRefs, invocation.Readiness.Request.AllowedToolRefs)
	boundCapabilityRefs := delegationWorkerAsyncRefs(input.BoundCapabilityRefs, capabilityRefs)
	boundAllowedToolRefs := delegationWorkerAsyncRefs(input.BoundAllowedToolRefs, allowedToolRefs)
	workerRunRef := delegationWorkerAsyncFirstRef(invocation.WorkerRunRef, invocation.ObservedWorkerRunRef)
	workerResultRef := delegationWorkerAsyncFirstRef(invocation.WorkerResultRef)
	workerReadbackRef := delegationWorkerAsyncFirstRef(invocation.WorkerReadbackRef)
	observationRef := delegationWorkerAsyncFirstRef(input.ObservationRef, invocation.ObservationRef)
	failureRef := delegationWorkerAsyncFirstRef(input.FailureRef, invocation.FailureRef)
	completionEnvelopeRef := delegationWorkerAsyncFirstRef(input.CompletionEnvelopeRef, delegationWorkerAsyncDerivedRef("envelope", workerResultRef), delegationWorkerAsyncDerivedRef("envelope", childRef))

	child := agentxcontrolplane.AutoDelegationAsyncChildReadback{
		ChildRef:                         childRef,
		Role:                             delegationWorkerAsyncRole(input.Role),
		Status:                           status,
		AgeSeconds:                       input.AgeSeconds,
		CurrentAction:                    strings.TrimSpace(input.CurrentAction),
		CapabilityRefs:                   capabilityRefs,
		AllowedToolRefs:                  allowedToolRefs,
		BoundCapabilityRefs:              boundCapabilityRefs,
		BoundAllowedToolRefs:             boundAllowedToolRefs,
		WorkerRunRef:                     workerRunRef,
		WorkerResultRef:                  workerResultRef,
		WorkerReadbackRef:                workerReadbackRef,
		ObservationRef:                   observationRef,
		FailureRef:                       failureRef,
		FailureReviewRef:                 delegationWorkerAsyncFirstRef(input.FailureReviewRef, invocation.Readiness.VerificationRef),
		CancellationRef:                  delegationWorkerAsyncFirstRef(input.CancellationRef),
		InterruptionRef:                  delegationWorkerAsyncFirstRef(input.InterruptionRef),
		CompletionEnvelopeRef:            completionEnvelopeRef,
		EvidenceRefs:                     agentxcontrolplane.MergeEvidenceRefs(input.EvidenceRefs, invocation.EvidenceRefs),
		ReadyForWorkerResultReview:       invocation.ReadyForWorkerResultReview,
		ReadyForFailureReview:            input.ReadyForFailureReview || invocation.ReadyForFailureReview,
		CancelAvailable:                  input.CancellationRef != "",
		InterruptAvailable:               input.InterruptionRef != "",
		WorkerResultRequiresVerification: true,
		WorkerOutputAcceptedAsFact:       false,
		MissingInputs:                    agentxcontrolplane.AppendMissingInputs(nil, input.MissingInputs...),
		BlockedReasons:                   delegationWorkerAsyncStrings(input.BlockedReasons),
		FailureClass:                     input.FailureClass,
		Boundaries: agentxcontrolplane.AppendBoundaries(input.Boundaries,
			"hostruntime_delegation_worker_async_child_readback",
			"completion_envelope_only",
			"raw_child_tool_logs_not_consumed",
		),
		RawOutputLoaded: input.RawOutputLoaded || invocation.RawOutputLoaded || run.RawOutputLoaded,
	}

	return agentxcontrolplane.BuildAutoDelegationAsyncCompletionProjection(agentxcontrolplane.AutoDelegationAsyncCompletionInput{
		HostBridge:            input.HostBridge,
		BackendKind:           delegationWorkerAsyncBackendKind(input.BackendKind),
		BackendRef:            delegationWorkerAsyncFirstRef(input.BackendRef),
		ParentObjectiveRef:    parentObjectiveRef,
		ParentObjectiveRunRef: parentObjectiveRunRef,
		ParentLedgerRef:       parentLedgerRef,
		ParentResumeRef:       parentResumeRef,
		RequireDurable:        input.RequireDurable,
		Children:              []agentxcontrolplane.AutoDelegationAsyncChildReadback{child},
		DecisionBasis: delegationWorkerAsyncRefs(
			[]agentxcontrolplane.DisplaySafeRef{
				"hostruntime:delegation_worker_async_completion",
				"hostruntime:host_owned_child_readback",
			},
			input.DecisionBasis,
		),
		Boundaries: agentxcontrolplane.AppendBoundaries(input.Boundaries,
			"hostruntime_delegation_worker_async_completion_projection",
			"host_owned_async_child_runtime",
			"no_child_task_spawn_by_core",
			"no_background_process_by_core",
		),
		RawOutputLoaded: child.RawOutputLoaded,
	}).Normalize()
}

func delegationWorkerAsyncStatus(status agentxcontrolplane.AutoDelegationAsyncChildStatus, invocation agentxcontrolplane.HostOwnedDelegationWorkerRuntimeInvocation) agentxcontrolplane.AutoDelegationAsyncChildStatus {
	if normalized := agentxcontrolplane.NormalizeAutoDelegationAsyncChildStatus(string(status)); normalized != agentxcontrolplane.AutoDelegationAsyncChildStatusUnknown {
		return normalized
	}
	invocation = invocation.Normalize()
	if invocation.HostInvocationCompleted && invocation.ReadyForWorkerResultReview {
		return agentxcontrolplane.AutoDelegationAsyncChildStatusCompleted
	}
	if invocation.HostInvocationFailed || invocation.ReadyForFailureReview {
		return agentxcontrolplane.AutoDelegationAsyncChildStatusFailed
	}
	if invocation.HostInvocationReported || invocation.WorkerRunRef != "" || invocation.ObservedWorkerRunRef != "" {
		return agentxcontrolplane.AutoDelegationAsyncChildStatusActive
	}
	return agentxcontrolplane.AutoDelegationAsyncChildStatusQueued
}

func delegationWorkerAsyncBackendKind(kind agentxcontrolplane.AutoDelegationAsyncBackendKind) agentxcontrolplane.AutoDelegationAsyncBackendKind {
	if normalized := agentxcontrolplane.NormalizeAutoDelegationAsyncBackendKind(string(kind)); normalized != agentxcontrolplane.AutoDelegationAsyncBackendUnsupported {
		return normalized
	}
	return agentxcontrolplane.AutoDelegationAsyncBackendProcessLocal
}

func delegationWorkerAsyncRole(role agentxcontrolplane.AutoDelegationChildRole) agentxcontrolplane.AutoDelegationChildRole {
	if normalized := agentxcontrolplane.NormalizeAutoDelegationChildRole(string(role)); normalized != "" {
		return normalized
	}
	return agentxcontrolplane.AutoDelegationChildRoleLeaf
}

func delegationWorkerAsyncFirstRef(groups ...interface{}) agentxcontrolplane.DisplaySafeRef {
	for _, group := range groups {
		switch values := group.(type) {
		case agentxcontrolplane.DisplaySafeRef:
			if ref := delegationWorkerAsyncSafeRef(values); ref != "" {
				return ref
			}
		case []agentxcontrolplane.DisplaySafeRef:
			for _, value := range values {
				if ref := delegationWorkerAsyncSafeRef(value); ref != "" {
					return ref
				}
			}
		}
	}
	return ""
}

func delegationWorkerAsyncRefs(groups ...[]agentxcontrolplane.DisplaySafeRef) []agentxcontrolplane.DisplaySafeRef {
	var out []agentxcontrolplane.DisplaySafeRef
	for _, values := range groups {
		for _, value := range values {
			if ref := delegationWorkerAsyncSafeRef(value); ref != "" {
				out = appendDisplaySafeRef(out, ref)
			}
		}
	}
	return out
}

func appendDisplaySafeRef(items []agentxcontrolplane.DisplaySafeRef, value agentxcontrolplane.DisplaySafeRef) []agentxcontrolplane.DisplaySafeRef {
	value = delegationWorkerAsyncSafeRef(value)
	if value == "" {
		return items
	}
	for _, item := range items {
		if item == value {
			return items
		}
	}
	return append(items, value)
}

func delegationWorkerAsyncSafeRef(value agentxcontrolplane.DisplaySafeRef) agentxcontrolplane.DisplaySafeRef {
	ref, ok := agentxcontrolplane.NormalizeDisplaySafeRef(strings.TrimSpace(string(value)))
	if !ok {
		return ""
	}
	return ref
}

func delegationWorkerAsyncDerivedRef(kind string, source agentxcontrolplane.DisplaySafeRef) agentxcontrolplane.DisplaySafeRef {
	source = delegationWorkerAsyncSafeRef(source)
	if source == "" {
		return ""
	}
	suffix := string(source)
	if idx := strings.LastIndex(suffix, ":"); idx >= 0 && idx+1 < len(suffix) {
		suffix = suffix[idx+1:]
	}
	suffix = strings.Trim(strings.ToLower(suffix), " ._:-")
	if suffix == "" {
		return ""
	}
	ref, ok := agentxcontrolplane.NormalizeDisplaySafeRef(kind + ":" + suffix)
	if !ok {
		return ""
	}
	return ref
}

func delegationWorkerAsyncStrings(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return appendUniqueStrings(nil, out...)
}
