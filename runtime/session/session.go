// Package session defines the portable identities and verification contracts
// used by Task, Session, and Subagent Host Kits.
//
// It is an Experimental naming boundary over the existing control kernel. It
// does not dispatch a worker, create a process, schedule a task, or persist a
// session. Use hostkit for the first recommended child-worker lifecycle.
package session

import controlcontract "github.com/wsnacj/agentx-go/runtime/controlcontract"

type (
	DisplaySafeRef          = controlcontract.DisplaySafeRef
	AttemptRef              = controlcontract.AttemptRef
	Boundary                = controlcontract.Boundary
	MissingInput            = controlcontract.MissingInput
	NextHostAction          = controlcontract.NextHostAction
	FailureClass            = controlcontract.FailureClass
	EvidenceStrength        = controlcontract.EvidenceStrength
	EvidenceRef             = controlcontract.EvidenceRef
	Observation             = controlcontract.Observation
	Activation              = controlcontract.Activation
	ExecutionIntensity      = controlcontract.ExecutionIntensity
	ControlMode             = controlcontract.ControlMode
	ObjectiveFrame          = controlcontract.ObjectiveFrame
	ObjectiveRun            = controlcontract.ObjectiveRun
	ObjectiveBudgetSnapshot = controlcontract.ObjectiveBudgetSnapshot

	HostActionStatus = controlcontract.HostActionStatus

	DelegationRequestInput                          = controlcontract.DelegationRequestInput
	DelegationRequestProjection                     = controlcontract.DelegationRequestProjection
	ProductionAdapterEffectGateKind                 = controlcontract.ProductionAdapterEffectGateKind
	ProductionAdapterIndependentEffectGateSpec      = controlcontract.ProductionAdapterIndependentEffectGateSpec
	ProductionAdapterIndependentEffectGate          = controlcontract.ProductionAdapterIndependentEffectGate
	HostOwnedDelegationWorkerRuntimeReadinessInput  = controlcontract.HostOwnedDelegationWorkerRuntimeReadinessInput
	HostOwnedDelegationWorkerRuntimeReadiness       = controlcontract.HostOwnedDelegationWorkerRuntimeReadiness
	HostOwnedDelegationWorkerRuntimeInvocationInput = controlcontract.HostOwnedDelegationWorkerRuntimeInvocationInput
	HostOwnedDelegationWorkerRuntimeInvocation      = controlcontract.HostOwnedDelegationWorkerRuntimeInvocation

	DelegationWorkerParentMergeInput        = controlcontract.DelegationWorkerParentMergeInput
	DelegationWorkerParentMerge             = controlcontract.DelegationWorkerParentMerge
	DelegationObjectiveRuntimeHandoffInput  = controlcontract.DelegationObjectiveRuntimeHandoffInput
	DelegationObjectiveRuntimeHandoff       = controlcontract.DelegationObjectiveRuntimeHandoff
	ObjectiveRuntimeLoopInput               = controlcontract.ObjectiveRuntimeLoopInput
	ObjectiveRuntimeLoopStep                = controlcontract.ObjectiveRuntimeLoopStep
	AutoDelegationHostBridge                = controlcontract.AutoDelegationHostBridge
	AutoDelegationAsyncBackendKind          = controlcontract.AutoDelegationAsyncBackendKind
	AutoDelegationAsyncChildStatus          = controlcontract.AutoDelegationAsyncChildStatus
	AutoDelegationChildRole                 = controlcontract.AutoDelegationChildRole
	AutoDelegationAsyncChildReadback        = controlcontract.AutoDelegationAsyncChildReadback
	AutoDelegationAsyncCompletionInput      = controlcontract.AutoDelegationAsyncCompletionInput
	AutoDelegationAsyncCompletionProjection = controlcontract.AutoDelegationAsyncCompletionProjection
)

const (
	FailureNone               = controlcontract.FailureNone
	FailureConfigMissing      = controlcontract.FailureConfigMissing
	FailurePolicyBlocked      = controlcontract.FailurePolicyBlocked
	FailureApprovalRequired   = controlcontract.FailureApprovalRequired
	FailureHostAdapterMissing = controlcontract.FailureHostAdapterMissing
	FailureEvidenceMissing    = controlcontract.FailureEvidenceMissing
	FailureEvidenceWeak       = controlcontract.FailureEvidenceWeak
	FailureVerificationFailed = controlcontract.FailureVerificationFailed
	FailureInternalError      = controlcontract.FailureInternalError
	EvidenceAdequate          = controlcontract.EvidenceAdequate

	ActivationManaged                           = controlcontract.ActivationManaged
	IntensityL4DurableLongRun                   = controlcontract.IntensityL4DurableLongRun
	ControlModeDelegated                        = controlcontract.ControlModeDelegated
	ProductionAdapterEffectGateDelegationWorker = controlcontract.ProductionAdapterEffectGateDelegationWorker

	HostActionBlocked  = controlcontract.HostActionBlocked
	HostActionRecorded = controlcontract.HostActionRecorded

	AutoDelegationAsyncBackendUnsupported  = controlcontract.AutoDelegationAsyncBackendUnsupported
	AutoDelegationAsyncBackendProcessLocal = controlcontract.AutoDelegationAsyncBackendProcessLocal
	AutoDelegationAsyncBackendDurable      = controlcontract.AutoDelegationAsyncBackendDurable

	AutoDelegationAsyncChildStatusUnknown   = controlcontract.AutoDelegationAsyncChildStatusUnknown
	AutoDelegationAsyncChildStatusQueued    = controlcontract.AutoDelegationAsyncChildStatusQueued
	AutoDelegationAsyncChildStatusActive    = controlcontract.AutoDelegationAsyncChildStatusActive
	AutoDelegationAsyncChildStatusCompleted = controlcontract.AutoDelegationAsyncChildStatusCompleted
	AutoDelegationAsyncChildStatusFailed    = controlcontract.AutoDelegationAsyncChildStatusFailed

	AutoDelegationChildRoleLeaf = controlcontract.AutoDelegationChildRoleLeaf
)

func NormalizeDisplaySafeRef(raw string) (DisplaySafeRef, bool) {
	return controlcontract.NormalizeDisplaySafeRef(raw)
}

func NormalizeAttemptRef(raw string) (AttemptRef, bool) {
	return controlcontract.NormalizeAttemptRef(raw)
}

func NormalizeFailureClass(raw string) FailureClass {
	return controlcontract.NormalizeFailureClass(raw)
}

func NormalizeAutoDelegationAsyncBackendKind(raw string) AutoDelegationAsyncBackendKind {
	return controlcontract.NormalizeAutoDelegationAsyncBackendKind(raw)
}

func NormalizeAutoDelegationAsyncChildStatus(raw string) AutoDelegationAsyncChildStatus {
	return controlcontract.NormalizeAutoDelegationAsyncChildStatus(raw)
}

func NormalizeAutoDelegationChildRole(raw string) AutoDelegationChildRole {
	return controlcontract.NormalizeAutoDelegationChildRole(raw)
}

func AppendBoundaries(base []Boundary, values ...Boundary) []Boundary {
	return controlcontract.AppendBoundaries(base, values...)
}

func AppendMissingInputs(base []MissingInput, values ...MissingInput) []MissingInput {
	return controlcontract.AppendMissingInputs(base, values...)
}

func MergeEvidenceRefs(groups ...[]EvidenceRef) []EvidenceRef {
	return controlcontract.MergeEvidenceRefs(groups...)
}

func BuildDelegationRequestProjection(input DelegationRequestInput) DelegationRequestProjection {
	return controlcontract.BuildDelegationRequestProjection(input)
}

func BuildProductionAdapterIndependentEffectGate(spec ProductionAdapterIndependentEffectGateSpec) ProductionAdapterIndependentEffectGate {
	return controlcontract.BuildProductionAdapterIndependentEffectGate(spec)
}

func BuildHostOwnedDelegationWorkerRuntimeReadiness(input HostOwnedDelegationWorkerRuntimeReadinessInput) HostOwnedDelegationWorkerRuntimeReadiness {
	return controlcontract.BuildHostOwnedDelegationWorkerRuntimeReadiness(input)
}

func BuildHostOwnedDelegationWorkerRuntimeInvocation(input HostOwnedDelegationWorkerRuntimeInvocationInput) HostOwnedDelegationWorkerRuntimeInvocation {
	return controlcontract.BuildHostOwnedDelegationWorkerRuntimeInvocation(input)
}

func BuildDelegationWorkerParentMerge(input DelegationWorkerParentMergeInput) DelegationWorkerParentMerge {
	return controlcontract.BuildDelegationWorkerParentMerge(input)
}

func BuildDelegationObjectiveRuntimeHandoff(input DelegationObjectiveRuntimeHandoffInput) DelegationObjectiveRuntimeHandoff {
	return controlcontract.BuildDelegationObjectiveRuntimeHandoff(input)
}

func BuildObjectiveRuntimeLoopStep(input ObjectiveRuntimeLoopInput) ObjectiveRuntimeLoopStep {
	return controlcontract.BuildObjectiveRuntimeLoopStep(input)
}

func BuildAutoDelegationAsyncCompletionProjection(input AutoDelegationAsyncCompletionInput) AutoDelegationAsyncCompletionProjection {
	return controlcontract.BuildAutoDelegationAsyncCompletionProjection(input)
}
