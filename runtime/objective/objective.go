// Package objective provides the narrow Objective contract names used by the
// recommended Host Kit.
//
// The package is currently Experimental. Its aliases preserve the type and
// JSON identity of the existing controlcontract kernels while the physical
// owners are split behind this stable package boundary.
package objective

import controlcontract "github.com/wsnacj/agentx-go/runtime/controlcontract"

type (
	Activation                     = controlcontract.Activation
	Boundary                       = controlcontract.Boundary
	ControlMode                    = controlcontract.ControlMode
	DisplaySafeRef                 = controlcontract.DisplaySafeRef
	EvidenceRef                    = controlcontract.EvidenceRef
	EvidenceStrength               = controlcontract.EvidenceStrength
	ExecutionIntensity             = controlcontract.ExecutionIntensity
	ExecutionIntensityPolicy       = controlcontract.ExecutionIntensityPolicy
	FailureClass                   = controlcontract.FailureClass
	HostAdapterRegistryInput       = controlcontract.HostAdapterRegistryInput
	HostAdapterRegistrySnapshot    = controlcontract.HostAdapterRegistrySnapshot
	ManagedObjectiveIngressInput   = controlcontract.ManagedObjectiveIngressInput
	ManagedObjectiveIngressResult  = controlcontract.ManagedObjectiveIngressProjection
	MissingInput                   = controlcontract.MissingInput
	NextHostAction                 = controlcontract.NextHostAction
	ObjectiveBudgetSnapshot        = controlcontract.ObjectiveBudgetSnapshot
	ObjectiveVerificationResult    = controlcontract.ObjectiveVerificationGateResult
	Observation                    = controlcontract.Observation
	ObservationNormalizationInput  = controlcontract.ObservationNormalizationInput
	ObservationNormalizationResult = controlcontract.ObservationNormalizationResult
	ObjectiveVerificationInput     = controlcontract.ObjectiveVerificationGateInput
	ProductionAdapterDescriptor    = controlcontract.ProductionAdapterDescriptor
	ProductionAdapterKind          = controlcontract.ProductionAdapterKind
	RuntimeAdapterRequest          = controlcontract.RuntimeAdapterExecutionRequest
	RuntimeAdapterResult           = controlcontract.RuntimeAdapterExecutionResult
	RuntimeAdapterResultInput      = controlcontract.RuntimeAdapterExecutionResultInput
	StrategyCandidate              = controlcontract.StrategyCandidate
	StrategyCatalogEntry           = controlcontract.StrategyCatalogEntry
	StrategyCatalogSnapshot        = controlcontract.StrategyCatalogSnapshot
	StrategyCatalogSourceKind      = controlcontract.StrategyCatalogSourceKind
	VerificationStatus             = controlcontract.VerificationStatus
)

const (
	ActivationOff         = controlcontract.ActivationOff
	ActivationObserveOnly = controlcontract.ActivationObserveOnly
	ActivationAdvisory    = controlcontract.ActivationAdvisory
	ActivationManaged     = controlcontract.ActivationManaged

	ControlModeAnswer     = controlcontract.ControlModeAnswer
	ControlModeTool       = controlcontract.ControlModeTool
	ControlModeWorkflow   = controlcontract.ControlModeWorkflow
	ControlModeOperations = controlcontract.ControlModeOperations
	ControlModeObjective  = controlcontract.ControlModeObjective
	ControlModeDelegated  = controlcontract.ControlModeDelegated

	IntensityL0AnswerOnly       = controlcontract.IntensityL0AnswerOnly
	IntensityL1ToolOnce         = controlcontract.IntensityL1ToolOnce
	IntensityL2BoundedToolLoop  = controlcontract.IntensityL2BoundedToolLoop
	IntensityL3ManagedObjective = controlcontract.IntensityL3ManagedObjective
	IntensityL4DurableLongRun   = controlcontract.IntensityL4DurableLongRun
	IntensityL5Autonomous       = controlcontract.IntensityL5Autonomous

	EvidenceStrong   = controlcontract.EvidenceStrong
	EvidenceAdequate = controlcontract.EvidenceAdequate
	EvidenceWeak     = controlcontract.EvidenceWeak
	EvidenceMissing  = controlcontract.EvidenceMissing

	VerificationNotEvaluated   = controlcontract.VerificationNotEvaluated
	VerificationNotApplicable  = controlcontract.VerificationNotApplicable
	VerificationSatisfied      = controlcontract.VerificationSatisfied
	VerificationPartial        = controlcontract.VerificationPartial
	VerificationBlocked        = controlcontract.VerificationBlocked
	VerificationReviewRequired = controlcontract.VerificationReviewRequired
	VerificationFailed         = controlcontract.VerificationFailed

	FailureNone               = controlcontract.FailureNone
	FailureInvalidInput       = controlcontract.FailureInvalidInput
	FailureHostAdapterMissing = controlcontract.FailureHostAdapterMissing
	FailureCapabilityMissing  = controlcontract.FailureCapabilityMissing
	FailureConfigMissing      = controlcontract.FailureConfigMissing
	FailureApprovalRequired   = controlcontract.FailureApprovalRequired
	FailurePolicyBlocked      = controlcontract.FailurePolicyBlocked
	FailureTargetUnavailable  = controlcontract.FailureTargetUnavailable
	FailureEvidenceMissing    = controlcontract.FailureEvidenceMissing
	FailureEvidenceWeak       = controlcontract.FailureEvidenceWeak
	FailureVerificationFailed = controlcontract.FailureVerificationFailed

	StrategyCatalogSourceTool        = controlcontract.StrategyCatalogSourceTool
	StrategyCatalogSourceSkill       = controlcontract.StrategyCatalogSourceSkill
	StrategyCatalogSourceWorkflow    = controlcontract.StrategyCatalogSourceWorkflow
	StrategyCatalogSourceHostAdapter = controlcontract.StrategyCatalogSourceHostAdapter
	StrategyCatalogSourceOperations  = controlcontract.StrategyCatalogSourceOperations

	ProductionAdapterSourceApply             = controlcontract.ProductionAdapterSourceApply
	ProductionAdapterSourceReadback          = controlcontract.ProductionAdapterSourceReadback
	ProductionAdapterCapabilityApply         = controlcontract.ProductionAdapterCapabilityApply
	ProductionAdapterOperationsSchedule      = controlcontract.ProductionAdapterOperationsSchedule
	ProductionAdapterOperationsMetricCollect = controlcontract.ProductionAdapterOperationsMetricCollect
	ProductionAdapterWorkflowDispatch        = controlcontract.ProductionAdapterWorkflowDispatch
)

// BuildManagedIngress projects host-owned Objective inputs into a ready or
// blocked runtime-adapter request without performing any side effect.
func BuildManagedIngress(input ManagedObjectiveIngressInput) ManagedObjectiveIngressResult {
	return controlcontract.BuildManagedObjectiveIngress(input)
}

// BuildHostAdapterRegistry validates and normalizes a host-owned registry.
func BuildHostAdapterRegistry(input HostAdapterRegistryInput) HostAdapterRegistrySnapshot {
	return controlcontract.BuildHostAdapterRegistry(input)
}

// BuildRuntimeAdapterResult validates a host-reported adapter result.
func BuildRuntimeAdapterResult(input RuntimeAdapterResultInput) RuntimeAdapterResult {
	return controlcontract.BuildRuntimeAdapterExecutionResult(input)
}

// BuildObservationNormalization converts structured host results into the
// canonical portable observation contract.
func BuildObservationNormalization(input ObservationNormalizationInput) ObservationNormalizationResult {
	return controlcontract.BuildObservationNormalization(input)
}

// BuildVerification evaluates normalized observations against the Objective.
func BuildVerification(input ObjectiveVerificationInput) ObjectiveVerificationResult {
	return controlcontract.BuildObjectiveVerificationGate(input)
}

// NormalizeDisplaySafeRef validates a display-safe reference.
func NormalizeDisplaySafeRef(raw string) (DisplaySafeRef, bool) {
	return controlcontract.NormalizeDisplaySafeRef(raw)
}

// AppendBoundaries preserves canonical boundary normalization and ordering.
func AppendBoundaries(existing []Boundary, values ...Boundary) []Boundary {
	return controlcontract.AppendBoundaries(existing, values...)
}

// MergeMissingInputs preserves canonical missing-input normalization.
func MergeMissingInputs(groups ...[]MissingInput) []MissingInput {
	return controlcontract.MergeMissingInputs(groups...)
}

// NormalizeFailureClass preserves canonical failure classification.
func NormalizeFailureClass(raw string) FailureClass {
	return controlcontract.NormalizeFailureClass(raw)
}
