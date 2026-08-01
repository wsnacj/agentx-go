// Package controlcontract defines the generic AgentX execution-control contract.
//
// The package intentionally contains only data shapes and pure normalization
// helpers. It must not import engine, runstore, scene, workflow, or host
// adapters, and it must not execute tools, runners, or storage mutations.
//
// 当前成熟度为 Experimental / private validation。该包固定可移植的状态、证据、
// 阻断、下一步动作与 display-safe 数据合同；具体调度、持久化和产品策略由 Host 拥有。
package controlcontract

import (
	"regexp"
	"sort"
	"strings"
)

const ContractVersion = "agentx.controlplane.contract.v1"

type Activation string

const (
	ActivationOff         Activation = "off"
	ActivationObserveOnly Activation = "observe_only"
	ActivationAdvisory    Activation = "advisory"
	ActivationManaged     Activation = "managed"
)

func KnownActivations() []Activation {
	return []Activation{
		ActivationOff,
		ActivationObserveOnly,
		ActivationAdvisory,
		ActivationManaged,
	}
}

func NormalizeActivation(raw string) Activation {
	switch normalizeEnumToken(raw) {
	case "", "off", "a0", "disabled", "disable":
		return ActivationOff
	case "observe", "observe_only", "observeonly", "diagnostics", "diagnostics_only", "a1":
		return ActivationObserveOnly
	case "advisory", "advise", "host_advisory", "a2":
		return ActivationAdvisory
	case "managed", "manage", "a3":
		return ActivationManaged
	default:
		return ActivationOff
	}
}

type ControlMode string

const (
	ControlModeAnswer               ControlMode = "answer"
	ControlModeTool                 ControlMode = "tool"
	ControlModeWorkflow             ControlMode = "workflow"
	ControlModeOperations           ControlMode = "operations"
	ControlModeCapabilityResolution ControlMode = "capability_resolution"
	ControlModeObjective            ControlMode = "objective"
	ControlModeDelegated            ControlMode = "delegated"
)

func KnownControlModes() []ControlMode {
	return []ControlMode{
		ControlModeAnswer,
		ControlModeTool,
		ControlModeWorkflow,
		ControlModeOperations,
		ControlModeCapabilityResolution,
		ControlModeObjective,
		ControlModeDelegated,
	}
}

func NormalizeControlMode(raw string) ControlMode {
	switch normalizeEnumToken(raw) {
	case "answer", "chat", "reply":
		return ControlModeAnswer
	case "tool", "tools", "tool_once", "tool_loop":
		return ControlModeTool
	case "workflow", "flow":
		return ControlModeWorkflow
	case "operations", "operation", "ops":
		return ControlModeOperations
	case "capability_resolution", "capability", "capabilities", "capability_install":
		return ControlModeCapabilityResolution
	case "objective", "goal", "task_goal":
		return ControlModeObjective
	case "delegated", "delegate", "subtask", "sub_agent":
		return ControlModeDelegated
	default:
		return ""
	}
}

type ExecutionIntensity string

const (
	IntensityL0AnswerOnly       ExecutionIntensity = "l0_answer_only"
	IntensityL1ToolOnce         ExecutionIntensity = "l1_tool_once"
	IntensityL2BoundedToolLoop  ExecutionIntensity = "l2_bounded_tool_loop"
	IntensityL3ManagedObjective ExecutionIntensity = "l3_managed_objective"
	IntensityL4DurableLongRun   ExecutionIntensity = "l4_durable_long_run"
	IntensityL5Autonomous       ExecutionIntensity = "l5_autonomous_delegated"
)

func KnownExecutionIntensities() []ExecutionIntensity {
	return []ExecutionIntensity{
		IntensityL0AnswerOnly,
		IntensityL1ToolOnce,
		IntensityL2BoundedToolLoop,
		IntensityL3ManagedObjective,
		IntensityL4DurableLongRun,
		IntensityL5Autonomous,
	}
}

func NormalizeExecutionIntensity(raw string) ExecutionIntensity {
	switch normalizeEnumToken(raw) {
	case "l0", "answer", "answer_only", "l0_answer_only":
		return IntensityL0AnswerOnly
	case "l1", "tool_once", "one_tool", "single_tool", "l1_tool_once":
		return IntensityL1ToolOnce
	case "l2", "bounded_tool_loop", "tool_loop", "bounded_retry", "l2_bounded_tool_loop":
		return IntensityL2BoundedToolLoop
	case "l3", "managed_objective", "objective", "l3_managed_objective":
		return IntensityL3ManagedObjective
	case "l4", "durable_long_run", "long_run", "durable", "l4_durable_long_run":
		return IntensityL4DurableLongRun
	case "l5", "autonomous_delegated", "delegated", "autonomous", "l5_autonomous_delegated":
		return IntensityL5Autonomous
	default:
		return ""
	}
}

type EvidenceStrength string

const (
	EvidenceStrong   EvidenceStrength = "strong"
	EvidenceAdequate EvidenceStrength = "adequate"
	EvidenceWeak     EvidenceStrength = "weak"
	EvidenceMissing  EvidenceStrength = "missing"
)

func NormalizeEvidenceStrength(raw string) EvidenceStrength {
	switch normalizeEnumToken(raw) {
	case "strong", "verified", "direct":
		return EvidenceStrong
	case "adequate", "ok", "sufficient":
		return EvidenceAdequate
	case "", "missing", "none", "absent":
		return EvidenceMissing
	case "weak", "partial", "indirect":
		return EvidenceWeak
	default:
		return EvidenceWeak
	}
}

type VerificationStatus string

const (
	VerificationNotEvaluated   VerificationStatus = "not_evaluated"
	VerificationNotApplicable  VerificationStatus = "not_applicable"
	VerificationSatisfied      VerificationStatus = "satisfied"
	VerificationPartial        VerificationStatus = "partial"
	VerificationBlocked        VerificationStatus = "blocked"
	VerificationReviewRequired VerificationStatus = "review_required"
	VerificationFailed         VerificationStatus = "failed"
)

func NormalizeVerificationStatus(raw string) VerificationStatus {
	switch normalizeEnumToken(raw) {
	case "", "not_evaluated", "unknown", "pending":
		return VerificationNotEvaluated
	case "not_applicable", "not_app", "notapplicable", "na", "n_a":
		return VerificationNotApplicable
	case "satisfied", "success", "succeeded", "complete", "completed":
		return VerificationSatisfied
	case "partial", "partially_satisfied":
		return VerificationPartial
	case "blocked", "cannot_continue":
		return VerificationBlocked
	case "review_required", "needs_review", "manual_review":
		return VerificationReviewRequired
	case "failed", "failure", "error":
		return VerificationFailed
	default:
		return VerificationReviewRequired
	}
}

type HostActionStatus string

const (
	HostActionNotReady         HostActionStatus = "not_ready"
	HostActionRequiresApproval HostActionStatus = "requires_approval"
	HostActionReady            HostActionStatus = "ready_for_host_action"
	HostActionRecorded         HostActionStatus = "host_action_recorded"
	HostActionReviewRequired   HostActionStatus = "review_required"
	HostActionBlocked          HostActionStatus = "blocked"
)

func NormalizeHostActionStatus(raw string) HostActionStatus {
	switch normalizeEnumToken(raw) {
	case "", "not_ready", "pending":
		return HostActionNotReady
	case "requires_approval", "approval_required", "needs_approval":
		return HostActionRequiresApproval
	case "ready", "ready_for_host_action", "ready_for_action":
		return HostActionReady
	case "host_action_recorded", "recorded", "applied":
		return HostActionRecorded
	case "review_required", "needs_review", "manual_review":
		return HostActionReviewRequired
	case "blocked":
		return HostActionBlocked
	default:
		return HostActionReviewRequired
	}
}

type FailureClass string

const (
	FailureNone                          FailureClass = "none"
	FailureInvalidInput                  FailureClass = "invalid_input"
	FailureAmbiguousGoal                 FailureClass = "ambiguous_goal"
	FailureInsufficientInformation       FailureClass = "insufficient_information"
	FailureToolUnavailable               FailureClass = "tool_unavailable"
	FailureSkillUnavailable              FailureClass = "skill_unavailable"
	FailureConnectorUnavailable          FailureClass = "connector_unavailable"
	FailureHostAdapterMissing            FailureClass = "host_adapter_missing"
	FailureCapabilityMissing             FailureClass = "capability_missing"
	FailureConfigMissing                 FailureClass = "config_missing"
	FailureCredentialMissing             FailureClass = "credential_missing"
	FailureAuthorizationMissing          FailureClass = "authorization_missing"
	FailureApprovalRequired              FailureClass = "approval_required"
	FailurePermissionDenied              FailureClass = "permission_denied"
	FailurePolicyBlocked                 FailureClass = "policy_blocked"
	FailureSandboxBlocked                FailureClass = "sandbox_blocked"
	FailureBudgetExhausted               FailureClass = "budget_exhausted"
	FailureTimeout                       FailureClass = "timeout"
	FailureRepeatedNoProgress            FailureClass = "repeated_no_progress"
	FailureTargetUnavailable             FailureClass = "target_unavailable"
	FailureTargetNotFound                FailureClass = "target_not_found"
	FailureExternalDependencyUnavailable FailureClass = "external_dependency_unavailable"
	FailureEvidenceMissing               FailureClass = "evidence_missing"
	FailureEvidenceWeak                  FailureClass = "evidence_weak"
	FailureVerificationFailed            FailureClass = "verification_failed"
	FailureUnsupportedOperation          FailureClass = "unsupported_operation"
	FailureInternalError                 FailureClass = "internal_error"
)

func KnownFailureClasses() []FailureClass {
	return []FailureClass{
		FailureNone,
		FailureInvalidInput,
		FailureAmbiguousGoal,
		FailureInsufficientInformation,
		FailureToolUnavailable,
		FailureSkillUnavailable,
		FailureConnectorUnavailable,
		FailureHostAdapterMissing,
		FailureCapabilityMissing,
		FailureConfigMissing,
		FailureCredentialMissing,
		FailureAuthorizationMissing,
		FailureApprovalRequired,
		FailurePermissionDenied,
		FailurePolicyBlocked,
		FailureSandboxBlocked,
		FailureBudgetExhausted,
		FailureTimeout,
		FailureRepeatedNoProgress,
		FailureTargetUnavailable,
		FailureTargetNotFound,
		FailureExternalDependencyUnavailable,
		FailureEvidenceMissing,
		FailureEvidenceWeak,
		FailureVerificationFailed,
		FailureUnsupportedOperation,
		FailureInternalError,
	}
}

func NormalizeFailureClass(raw string) FailureClass {
	token := normalizeEnumToken(raw)
	if token == "" {
		return FailureNone
	}
	for _, known := range KnownFailureClasses() {
		if token == string(known) {
			return known
		}
	}
	return FailureInternalError
}

type Boundary string
type MissingInput string
type NextHostAction string
type DisplaySafeRef string
type AttemptRef string

type RedactionPolicy struct {
	DisplaySafeOnly bool     `json:"display_safe_only"`
	BlockLocalPaths bool     `json:"block_local_paths"`
	BlockURLs       bool     `json:"block_urls"`
	BlockSecrets    bool     `json:"block_secrets"`
	AllowedPrefixes []string `json:"allowed_prefixes,omitempty"`
}

func DefaultRedactionPolicy() RedactionPolicy {
	return RedactionPolicy{
		DisplaySafeOnly: true,
		BlockLocalPaths: true,
		BlockURLs:       true,
		BlockSecrets:    true,
	}
}

func (p RedactionPolicy) Normalize() RedactionPolicy {
	out := p
	out.AllowedPrefixes = normalizeStringList(out.AllowedPrefixes)
	sort.Strings(out.AllowedPrefixes)
	return out
}

type EvidenceRef struct {
	Ref        DisplaySafeRef   `json:"ref,omitempty"`
	Kind       string           `json:"kind,omitempty"`
	Strength   EvidenceStrength `json:"strength,omitempty"`
	Source     DisplaySafeRef   `json:"source,omitempty"`
	ObservedAt string           `json:"observed_at,omitempty"`
}

func (r EvidenceRef) Normalize() EvidenceRef {
	out := r
	out.Kind = normalizeControlToken(out.Kind)
	out.Strength = NormalizeEvidenceStrength(string(out.Strength))
	out.ObservedAt = strings.TrimSpace(out.ObservedAt)
	out.Ref, _ = NormalizeDisplaySafeRef(string(out.Ref))
	out.Source, _ = NormalizeDisplaySafeRef(string(out.Source))
	return out
}

type ObjectiveFrame struct {
	ContractVersion       string             `json:"contract_version,omitempty"`
	ID                    string             `json:"id,omitempty"`
	UserGoalDigest        string             `json:"user_goal_digest,omitempty"`
	ControlMode           ControlMode        `json:"control_mode,omitempty"`
	Intensity             ExecutionIntensity `json:"intensity,omitempty"`
	SuccessCriteria       []string           `json:"success_criteria,omitempty"`
	Constraints           []string           `json:"constraints,omitempty"`
	RequiredEvidence      []EvidenceRef      `json:"required_evidence,omitempty"`
	CandidateCapabilities []DisplaySafeRef   `json:"candidate_capabilities,omitempty"`
	SourceContext         []DisplaySafeRef   `json:"source_context,omitempty"`
	Boundaries            []Boundary         `json:"boundaries,omitempty"`
	MissingInputs         []MissingInput     `json:"missing_inputs,omitempty"`
}

func CloneObjectiveFrame(in ObjectiveFrame) ObjectiveFrame {
	out := in
	out.SuccessCriteria = cloneStringSlice(in.SuccessCriteria)
	out.Constraints = cloneStringSlice(in.Constraints)
	out.RequiredEvidence = cloneEvidenceRefs(in.RequiredEvidence)
	out.CandidateCapabilities = cloneDisplaySafeRefs(in.CandidateCapabilities)
	out.SourceContext = cloneDisplaySafeRefs(in.SourceContext)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	return out
}

func (f ObjectiveFrame) Clone() ObjectiveFrame {
	return CloneObjectiveFrame(f)
}

func (f ObjectiveFrame) Normalize() ObjectiveFrame {
	out := CloneObjectiveFrame(f)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.ID = strings.TrimSpace(out.ID)
	out.UserGoalDigest = strings.TrimSpace(out.UserGoalDigest)
	out.ControlMode = NormalizeControlMode(string(out.ControlMode))
	out.Intensity = NormalizeExecutionIntensity(string(out.Intensity))
	out.SuccessCriteria = normalizeStringList(out.SuccessCriteria)
	out.Constraints = normalizeStringList(out.Constraints)
	out.RequiredEvidence = normalizeEvidenceRefs(out.RequiredEvidence)
	out.CandidateCapabilities = normalizeDisplaySafeRefs(out.CandidateCapabilities)
	out.SourceContext = normalizeDisplaySafeRefs(out.SourceContext)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	return out
}

type StrategyCandidate struct {
	ContractVersion  string             `json:"contract_version,omitempty"`
	ID               string             `json:"id,omitempty"`
	Kind             string             `json:"kind,omitempty"`
	ControlMode      ControlMode        `json:"control_mode,omitempty"`
	MinIntensity     ExecutionIntensity `json:"min_intensity,omitempty"`
	MaxIntensity     ExecutionIntensity `json:"max_intensity,omitempty"`
	CapabilityRefs   []DisplaySafeRef   `json:"capability_refs,omitempty"`
	ExpectedEvidence []EvidenceRef      `json:"expected_evidence,omitempty"`
	Preconditions    []MissingInput     `json:"preconditions,omitempty"`
	Boundaries       []Boundary         `json:"boundaries,omitempty"`
	Risk             string             `json:"risk,omitempty"`
	SideEffectClass  string             `json:"side_effect_class,omitempty"`
	RequiresApproval bool               `json:"requires_approval"`
	FallbackOf       DisplaySafeRef     `json:"fallback_of,omitempty"`
	Owner            string             `json:"owner,omitempty"`
}

func CloneStrategyCandidate(in StrategyCandidate) StrategyCandidate {
	out := in
	out.CapabilityRefs = cloneDisplaySafeRefs(in.CapabilityRefs)
	out.ExpectedEvidence = cloneEvidenceRefs(in.ExpectedEvidence)
	out.Preconditions = cloneMissingInputs(in.Preconditions)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (s StrategyCandidate) Clone() StrategyCandidate {
	return CloneStrategyCandidate(s)
}

func (s StrategyCandidate) Normalize() StrategyCandidate {
	out := CloneStrategyCandidate(s)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.ID = strings.TrimSpace(out.ID)
	out.Kind = normalizeControlToken(out.Kind)
	out.ControlMode = NormalizeControlMode(string(out.ControlMode))
	out.MinIntensity = NormalizeExecutionIntensity(string(out.MinIntensity))
	out.MaxIntensity = NormalizeExecutionIntensity(string(out.MaxIntensity))
	out.CapabilityRefs = normalizeDisplaySafeRefs(out.CapabilityRefs)
	out.ExpectedEvidence = normalizeEvidenceRefs(out.ExpectedEvidence)
	out.Preconditions = normalizeMissingInputs(out.Preconditions)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.Risk = normalizeControlToken(out.Risk)
	out.SideEffectClass = normalizeControlToken(out.SideEffectClass)
	out.FallbackOf, _ = NormalizeDisplaySafeRef(string(out.FallbackOf))
	out.Owner = normalizeControlToken(out.Owner)
	return out
}

type Observation struct {
	ContractVersion   string           `json:"contract_version,omitempty"`
	Kind              string           `json:"kind,omitempty"`
	Source            DisplaySafeRef   `json:"source,omitempty"`
	Subject           DisplaySafeRef   `json:"subject,omitempty"`
	Name              string           `json:"name,omitempty"`
	Value             string           `json:"value,omitempty"`
	Unit              string           `json:"unit,omitempty"`
	Strength          EvidenceStrength `json:"strength,omitempty"`
	ObservedAt        string           `json:"observed_at,omitempty"`
	EvidenceRefs      []EvidenceRef    `json:"evidence_refs,omitempty"`
	DisplaySafeRefs   []DisplaySafeRef `json:"display_safe_refs,omitempty"`
	DegradationReason string           `json:"degradation_reason,omitempty"`
	RawOutputLoaded   bool             `json:"raw_output_loaded"`
}

func CloneObservation(in Observation) Observation {
	out := in
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.DisplaySafeRefs = cloneDisplaySafeRefs(in.DisplaySafeRefs)
	return out
}

func (o Observation) Clone() Observation {
	return CloneObservation(o)
}

func (o Observation) Normalize() Observation {
	out := CloneObservation(o)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Kind = normalizeControlToken(out.Kind)
	out.Source, _ = NormalizeDisplaySafeRef(string(out.Source))
	out.Subject, _ = NormalizeDisplaySafeRef(string(out.Subject))
	out.Name = strings.TrimSpace(out.Name)
	out.Value = strings.TrimSpace(out.Value)
	out.Unit = strings.TrimSpace(out.Unit)
	out.Strength = NormalizeEvidenceStrength(string(out.Strength))
	out.ObservedAt = strings.TrimSpace(out.ObservedAt)
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
	out.DisplaySafeRefs = normalizeDisplaySafeRefs(out.DisplaySafeRefs)
	out.DegradationReason = normalizeControlToken(out.DegradationReason)
	return out
}

type AttemptSummary struct {
	ContractVersion  string             `json:"contract_version,omitempty"`
	Ref              AttemptRef         `json:"ref,omitempty"`
	ObjectiveID      string             `json:"objective_id,omitempty"`
	StrategyID       string             `json:"strategy_id,omitempty"`
	Index            int                `json:"index,omitempty"`
	ControlMode      ControlMode        `json:"control_mode,omitempty"`
	Intensity        ExecutionIntensity `json:"intensity,omitempty"`
	Status           VerificationStatus `json:"status,omitempty"`
	EvidenceRefs     []EvidenceRef      `json:"evidence_refs,omitempty"`
	ObservationCount int                `json:"observation_count,omitempty"`
	FailureClass     FailureClass       `json:"failure_class,omitempty"`
	FailureReason    string             `json:"failure_reason,omitempty"`
	NextHostAction   NextHostAction     `json:"next_host_action,omitempty"`
	Boundaries       []Boundary         `json:"boundaries,omitempty"`
	MissingInputs    []MissingInput     `json:"missing_inputs,omitempty"`
	RawOutputLoaded  bool               `json:"raw_output_loaded"`
}

func CloneAttemptSummary(in AttemptSummary) AttemptSummary {
	out := in
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	return out
}

func (a AttemptSummary) Clone() AttemptSummary {
	return CloneAttemptSummary(a)
}

func (a AttemptSummary) Normalize() AttemptSummary {
	out := CloneAttemptSummary(a)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Ref, _ = NormalizeAttemptRef(string(out.Ref))
	out.ObjectiveID = strings.TrimSpace(out.ObjectiveID)
	out.StrategyID = strings.TrimSpace(out.StrategyID)
	out.ControlMode = NormalizeControlMode(string(out.ControlMode))
	out.Intensity = NormalizeExecutionIntensity(string(out.Intensity))
	out.Status = NormalizeVerificationStatus(string(out.Status))
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.FailureReason = strings.TrimSpace(out.FailureReason)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	if out.ObservationCount < 0 {
		out.ObservationCount = 0
	}
	return out
}

type Attempt struct {
	AttemptSummary
	Observations []Observation `json:"observations,omitempty"`
}

func CloneAttempt(in Attempt) Attempt {
	out := in
	out.AttemptSummary = CloneAttemptSummary(in.AttemptSummary)
	out.Observations = cloneObservations(in.Observations)
	return out
}

func (a Attempt) Clone() Attempt {
	return CloneAttempt(a)
}

func (a Attempt) Normalize() Attempt {
	out := CloneAttempt(a)
	out.AttemptSummary = out.AttemptSummary.Normalize()
	out.Observations = normalizeObservations(out.Observations)
	if out.ObservationCount == 0 {
		out.ObservationCount = len(out.Observations)
	}
	return out
}

type AttemptLedgerPatch struct {
	ContractVersion      string           `json:"contract_version,omitempty"`
	ObjectiveID          string           `json:"objective_id,omitempty"`
	LedgerRef            DisplaySafeRef   `json:"ledger_ref,omitempty"`
	Attempts             []AttemptSummary `json:"attempts,omitempty"`
	RetryBudgetUsed      int              `json:"retry_budget_used,omitempty"`
	RetryBudgetRemaining int              `json:"retry_budget_remaining,omitempty"`
	EvidenceRefs         []EvidenceRef    `json:"evidence_refs,omitempty"`
	Boundaries           []Boundary       `json:"boundaries,omitempty"`
	MissingInputs        []MissingInput   `json:"missing_inputs,omitempty"`
	NextHostAction       NextHostAction   `json:"next_host_action,omitempty"`
	RawOutputLoaded      bool             `json:"raw_output_loaded"`
}

func CloneAttemptLedgerPatch(in AttemptLedgerPatch) AttemptLedgerPatch {
	out := in
	out.Attempts = cloneAttemptSummaries(in.Attempts)
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	return out
}

func (p AttemptLedgerPatch) Clone() AttemptLedgerPatch {
	return CloneAttemptLedgerPatch(p)
}

func (p AttemptLedgerPatch) Normalize() AttemptLedgerPatch {
	out := CloneAttemptLedgerPatch(p)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.ObjectiveID = strings.TrimSpace(out.ObjectiveID)
	out.LedgerRef, _ = NormalizeDisplaySafeRef(string(out.LedgerRef))
	out.Attempts = normalizeAttemptSummaries(out.Attempts)
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	if out.RetryBudgetUsed < 0 {
		out.RetryBudgetUsed = 0
	}
	if out.RetryBudgetRemaining < 0 {
		out.RetryBudgetRemaining = 0
	}
	return out
}

// ManagedObjectiveProjection is the generic L3 managed-objective contract
// surface. It is a projection only: it does not dispatch strategies, write
// durable ledgers, or mutate host state.
type ManagedObjectiveProjection struct {
	ContractVersion     string             `json:"contract_version,omitempty"`
	Projected           bool               `json:"projected"`
	Activation          Activation         `json:"activation,omitempty"`
	Status              HostActionStatus   `json:"status,omitempty"`
	Ready               bool               `json:"ready"`
	Frame               ObjectiveFrame     `json:"frame,omitempty"`
	Ledger              AttemptLedgerPatch `json:"ledger,omitempty"`
	RequiresApproval    bool               `json:"requires_approval"`
	ApprovalRefs        []DisplaySafeRef   `json:"approval_refs,omitempty"`
	PolicyRefs          []DisplaySafeRef   `json:"policy_refs,omitempty"`
	AllowedStrategyRefs []DisplaySafeRef   `json:"allowed_strategy_refs,omitempty"`
	EvidenceRefs        []EvidenceRef      `json:"evidence_refs,omitempty"`
	FailureClass        FailureClass       `json:"failure_class,omitempty"`
	MissingInputs       []MissingInput     `json:"missing_inputs,omitempty"`
	Boundaries          []Boundary         `json:"boundaries,omitempty"`
	NextHostAction      NextHostAction     `json:"next_host_action,omitempty"`
	RunnerEffect        string             `json:"runner_effect,omitempty"`
	PromptEffect        string             `json:"prompt_effect,omitempty"`
	RawOutputLoaded     bool               `json:"raw_output_loaded"`
}

// ManagedObjectiveReplannerProjection is a proposal-only L3 replanner surface.
// It can describe candidate host actions, but it does not dispatch strategies
// or authorize durable state changes.
type ManagedObjectiveReplannerProjection struct {
	ContractVersion string              `json:"contract_version,omitempty"`
	Projected       bool                `json:"projected"`
	Activation      Activation          `json:"activation,omitempty"`
	Status          string              `json:"status,omitempty"`
	Mode            string              `json:"mode,omitempty"`
	ObjectiveID     string              `json:"objective_id,omitempty"`
	Verification    VerificationResult  `json:"verification,omitempty"`
	Candidates      []StrategyCandidate `json:"candidates,omitempty"`
	Proposal        HostActionProposal  `json:"proposal,omitempty"`
	EvidenceRefs    []EvidenceRef       `json:"evidence_refs,omitempty"`
	FailureClass    FailureClass        `json:"failure_class,omitempty"`
	BlockedReason   string              `json:"blocked_reason,omitempty"`
	MissingInputs   []MissingInput      `json:"missing_inputs,omitempty"`
	DecisionBasis   []DisplaySafeRef    `json:"decision_basis,omitempty"`
	Boundaries      []Boundary          `json:"boundaries,omitempty"`
	NextHostAction  NextHostAction      `json:"next_host_action,omitempty"`
	RunnerEffect    string              `json:"runner_effect,omitempty"`
	PromptEffect    string              `json:"prompt_effect,omitempty"`
	RawOutputLoaded bool                `json:"raw_output_loaded"`
}

type ReplannerSourceKind string

const (
	ReplannerSourceOperations ReplannerSourceKind = "operations"
	ReplannerSourceCapability ReplannerSourceKind = "capability"
	ReplannerSourceWorkflow   ReplannerSourceKind = "workflow"
	ReplannerSourceRecovery   ReplannerSourceKind = "recovery"
)

func KnownReplannerSourceKinds() []ReplannerSourceKind {
	return []ReplannerSourceKind{
		ReplannerSourceOperations,
		ReplannerSourceCapability,
		ReplannerSourceWorkflow,
		ReplannerSourceRecovery,
	}
}

func NormalizeReplannerSourceKind(raw string) ReplannerSourceKind {
	switch normalizeEnumToken(raw) {
	case "operations", "operation", "ops":
		return ReplannerSourceOperations
	case "capability", "capabilities", "capability_resolution", "capability_install":
		return ReplannerSourceCapability
	case "workflow", "workflow_node":
		return ReplannerSourceWorkflow
	case "recovery", "objective_recovery", "recovery_contract":
		return ReplannerSourceRecovery
	default:
		return ""
	}
}

// ReplannerSourceProjection is the generic proposal-only handoff from a
// structured source such as operations, capability resolution, or workflow into
// the managed-objective replanner. It carries only display-safe refs and
// metadata; it does not apply proposals, dispatch workflows, install tools, or
// mutate durable host state.
type ReplannerSourceProjection struct {
	ContractVersion string              `json:"contract_version,omitempty"`
	Projected       bool                `json:"projected"`
	SourceKind      ReplannerSourceKind `json:"source_kind,omitempty"`
	SourceRef       DisplaySafeRef      `json:"source_ref,omitempty"`
	Producer        DisplaySafeRef      `json:"producer,omitempty"`
	ControlMode     ControlMode         `json:"control_mode,omitempty"`
	Status          VerificationStatus  `json:"status,omitempty"`
	FailureClass    FailureClass        `json:"failure_class,omitempty"`
	Observation     Observation         `json:"observation,omitempty"`
	Verification    VerificationResult  `json:"verification,omitempty"`
	Candidate       StrategyCandidate   `json:"candidate,omitempty"`
	Proposal        HostActionProposal  `json:"proposal,omitempty"`
	EvidenceRefs    []EvidenceRef       `json:"evidence_refs,omitempty"`
	SourceRefs      []DisplaySafeRef    `json:"source_refs,omitempty"`
	ProposalRefs    []DisplaySafeRef    `json:"proposal_refs,omitempty"`
	MissingInputs   []MissingInput      `json:"missing_inputs,omitempty"`
	Boundaries      []Boundary          `json:"boundaries,omitempty"`
	NextHostAction  NextHostAction      `json:"next_host_action,omitempty"`
	RunnerEffect    string              `json:"runner_effect,omitempty"`
	PromptEffect    string              `json:"prompt_effect,omitempty"`
	RawOutputLoaded bool                `json:"raw_output_loaded"`
}

func CloneManagedObjectiveProjection(in ManagedObjectiveProjection) ManagedObjectiveProjection {
	out := in
	out.Frame = in.Frame.Clone()
	out.Ledger = in.Ledger.Clone()
	out.ApprovalRefs = cloneDisplaySafeRefs(in.ApprovalRefs)
	out.PolicyRefs = cloneDisplaySafeRefs(in.PolicyRefs)
	out.AllowedStrategyRefs = cloneDisplaySafeRefs(in.AllowedStrategyRefs)
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (p ManagedObjectiveProjection) Clone() ManagedObjectiveProjection {
	return CloneManagedObjectiveProjection(p)
}

func (p ManagedObjectiveProjection) Normalize() ManagedObjectiveProjection {
	out := CloneManagedObjectiveProjection(p)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Activation = NormalizeActivation(string(out.Activation))
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.Frame = out.Frame.Normalize()
	out.Ledger = out.Ledger.Normalize()
	out.RequiresApproval = true
	out.ApprovalRefs = normalizeDisplaySafeRefs(out.ApprovalRefs)
	out.PolicyRefs = normalizeDisplaySafeRefs(out.PolicyRefs)
	out.AllowedStrategyRefs = normalizeDisplaySafeRefs(out.AllowedStrategyRefs)
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	if out.RunnerEffect == "" {
		out.RunnerEffect = "none"
	}
	if out.PromptEffect == "" {
		out.PromptEffect = "none"
	}
	if out.RawOutputLoaded && out.Status != HostActionBlocked {
		out.Status = HostActionReviewRequired
		if out.FailureClass == FailureNone {
			out.FailureClass = FailureEvidenceWeak
		}
		out.Boundaries = appendUniqueBoundary(out.Boundaries, "raw_output_not_allowed")
		out.MissingInputs = appendUniqueMissingInput(out.MissingInputs, "host:display_safe_refs")
		if out.NextHostAction == "" {
			out.NextHostAction = "provide_display_safe_refs"
		}
	}
	out.Ready = out.Activation == ActivationManaged &&
		out.Status == HostActionReady &&
		len(out.MissingInputs) == 0 &&
		!out.RawOutputLoaded
	return out
}

func CloneManagedObjectiveReplannerProjection(in ManagedObjectiveReplannerProjection) ManagedObjectiveReplannerProjection {
	out := in
	out.Verification = in.Verification.Clone()
	out.Candidates = cloneStrategyCandidates(in.Candidates)
	out.Proposal = in.Proposal.Clone()
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.DecisionBasis = cloneDisplaySafeRefs(in.DecisionBasis)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (p ManagedObjectiveReplannerProjection) Clone() ManagedObjectiveReplannerProjection {
	return CloneManagedObjectiveReplannerProjection(p)
}

func (p ManagedObjectiveReplannerProjection) Normalize() ManagedObjectiveReplannerProjection {
	out := CloneManagedObjectiveReplannerProjection(p)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Activation = NormalizeActivation(string(out.Activation))
	out.Status = normalizeControlToken(out.Status)
	out.Mode = normalizeControlToken(out.Mode)
	out.ObjectiveID = strings.TrimSpace(out.ObjectiveID)
	out.Verification = out.Verification.Normalize()
	out.Candidates = normalizeStrategyCandidates(out.Candidates)
	out.Proposal = out.Proposal.Normalize()
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.BlockedReason = strings.TrimSpace(out.BlockedReason)
	if ContainsUnsafeRawOutput(out.BlockedReason) {
		out.BlockedReason = ""
	}
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.DecisionBasis = normalizeDisplaySafeRefs(out.DecisionBasis)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	if out.Status == "" {
		out.Status = "not_candidate"
	}
	if out.Mode == "" {
		out.Mode = "l3_managed_objective_replanner"
	}
	if out.RunnerEffect == "" {
		out.RunnerEffect = "none"
	}
	if out.PromptEffect == "" {
		out.PromptEffect = "none"
	}
	if out.RawOutputLoaded {
		out.Status = "review_required"
		if out.FailureClass == FailureNone {
			out.FailureClass = FailureEvidenceWeak
		}
		out.Boundaries = appendUniqueBoundary(out.Boundaries, "raw_output_not_allowed")
		out.MissingInputs = appendUniqueMissingInput(out.MissingInputs, "host:display_safe_refs")
		if out.NextHostAction == "" {
			out.NextHostAction = "provide_display_safe_refs"
		}
	}
	return out
}

func CloneReplannerSourceProjection(in ReplannerSourceProjection) ReplannerSourceProjection {
	out := in
	out.Observation = in.Observation.Clone()
	out.Verification = in.Verification.Clone()
	out.Candidate = in.Candidate.Clone()
	out.Proposal = in.Proposal.Clone()
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.SourceRefs = cloneDisplaySafeRefs(in.SourceRefs)
	out.ProposalRefs = cloneDisplaySafeRefs(in.ProposalRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (p ReplannerSourceProjection) Clone() ReplannerSourceProjection {
	return CloneReplannerSourceProjection(p)
}

func (p ReplannerSourceProjection) Normalize() ReplannerSourceProjection {
	out := CloneReplannerSourceProjection(p)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.SourceKind = NormalizeReplannerSourceKind(string(out.SourceKind))
	out.SourceRef, _ = NormalizeDisplaySafeRef(string(out.SourceRef))
	out.Producer, _ = NormalizeDisplaySafeRef(string(out.Producer))
	out.ControlMode = NormalizeControlMode(string(out.ControlMode))
	out.Status = NormalizeVerificationStatus(string(out.Status))
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.Observation = out.Observation.Normalize()
	out.Verification = out.Verification.Normalize()
	out.Candidate = out.Candidate.Normalize()
	out.Proposal = out.Proposal.Normalize()
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
	out.SourceRefs = normalizeDisplaySafeRefs(out.SourceRefs)
	out.ProposalRefs = normalizeDisplaySafeRefs(out.ProposalRefs)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	if out.RunnerEffect == "" {
		out.RunnerEffect = "none"
	}
	if out.PromptEffect == "" {
		out.PromptEffect = "none"
	}
	if out.RawOutputLoaded {
		out.Status = VerificationReviewRequired
		if out.FailureClass == FailureNone {
			out.FailureClass = FailureEvidenceWeak
		}
		out.Boundaries = appendUniqueBoundary(out.Boundaries, "raw_output_not_allowed")
		out.MissingInputs = appendUniqueMissingInput(out.MissingInputs, "host:display_safe_refs")
		if out.NextHostAction == "" {
			out.NextHostAction = "provide_display_safe_refs"
		}
	}
	return out
}

type VerificationResult struct {
	ContractVersion string             `json:"contract_version,omitempty"`
	Status          VerificationStatus `json:"status,omitempty"`
	Satisfied       bool               `json:"satisfied"`
	FailureClass    FailureClass       `json:"failure_class,omitempty"`
	FailureReason   string             `json:"failure_reason,omitempty"`
	EvidenceRefs    []EvidenceRef      `json:"evidence_refs,omitempty"`
	MissingInputs   []MissingInput     `json:"missing_inputs,omitempty"`
	Boundaries      []Boundary         `json:"boundaries,omitempty"`
	Findings        []string           `json:"findings,omitempty"`
	NextHostAction  NextHostAction     `json:"next_host_action,omitempty"`
	RawOutputLoaded bool               `json:"raw_output_loaded"`
}

func CloneVerificationResult(in VerificationResult) VerificationResult {
	out := in
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	out.Findings = cloneStringSlice(in.Findings)
	return out
}

func (r VerificationResult) Clone() VerificationResult {
	return CloneVerificationResult(r)
}

func (r VerificationResult) Normalize() VerificationResult {
	out := CloneVerificationResult(r)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Status = NormalizeVerificationStatus(string(out.Status))
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.FailureReason = strings.TrimSpace(out.FailureReason)
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.Findings = normalizeStringList(out.Findings)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	if out.RawOutputLoaded {
		if out.Status == VerificationSatisfied {
			out.Status = VerificationReviewRequired
		}
		out.Satisfied = false
		if out.FailureClass == FailureNone {
			out.FailureClass = FailureEvidenceWeak
		}
		out.Boundaries = appendUniqueBoundary(out.Boundaries, "raw_output_not_allowed")
		out.MissingInputs = appendUniqueMissingInput(out.MissingInputs, "host:display_safe_refs")
		if out.NextHostAction == "" {
			out.NextHostAction = "provide_display_safe_refs"
		}
		return out
	}
	out.Satisfied = out.Status == VerificationSatisfied
	return out
}

type HostActionProposal struct {
	ContractVersion  string           `json:"contract_version,omitempty"`
	Kind             string           `json:"kind,omitempty"`
	Status           HostActionStatus `json:"status,omitempty"`
	RequiresApproval bool             `json:"requires_approval"`
	ApprovalRefs     []DisplaySafeRef `json:"approval_refs,omitempty"`
	ActionRefs       []DisplaySafeRef `json:"action_refs,omitempty"`
	EvidenceRefs     []EvidenceRef    `json:"evidence_refs,omitempty"`
	FailureClass     FailureClass     `json:"failure_class,omitempty"`
	FailureReason    string           `json:"failure_reason,omitempty"`
	MissingInputs    []MissingInput   `json:"missing_inputs,omitempty"`
	Boundaries       []Boundary       `json:"boundaries,omitempty"`
	NextHostAction   NextHostAction   `json:"next_host_action,omitempty"`
	RawOutputLoaded  bool             `json:"raw_output_loaded"`
}

func CloneHostActionProposal(in HostActionProposal) HostActionProposal {
	out := in
	out.ApprovalRefs = cloneDisplaySafeRefs(in.ApprovalRefs)
	out.ActionRefs = cloneDisplaySafeRefs(in.ActionRefs)
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (p HostActionProposal) Clone() HostActionProposal {
	return CloneHostActionProposal(p)
}

func (p HostActionProposal) Normalize() HostActionProposal {
	out := CloneHostActionProposal(p)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Kind = normalizeControlToken(out.Kind)
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.ApprovalRefs = normalizeDisplaySafeRefs(out.ApprovalRefs)
	out.ActionRefs = normalizeDisplaySafeRefs(out.ActionRefs)
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.FailureReason = strings.TrimSpace(out.FailureReason)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	if out.RawOutputLoaded && out.Status != HostActionBlocked {
		out.Status = HostActionReviewRequired
		out.FailureClass = FailureEvidenceWeak
		out.Boundaries = appendUniqueBoundary(out.Boundaries, "raw_output_not_allowed")
		out.MissingInputs = appendUniqueMissingInput(out.MissingInputs, "host:display_safe_refs")
	}
	return out
}

var (
	displaySafeRefPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_.:-]{0,127}$`)
	urlPattern            = regexp.MustCompile(`(?i)\b[a-z][a-z0-9+.-]*://`)
	unixPathPattern       = regexp.MustCompile(`(^|[\s"'(])/(Users|home|var|etc|tmp|private|opt|usr|Volumes)/[^\s"' )]+`)
	windowsPathPattern    = regexp.MustCompile(`(?i)\b[A-Z]:\\[^\s]+`)
	secretPattern         = regexp.MustCompile(`(?i)\b(api[_-]?key|access[_-]?token|refresh[_-]?token|auth[_-]?token|password|passwd|secret|credential)\s*[:=]\s*\S+`)
	pemPattern            = regexp.MustCompile(`-----BEGIN [A-Z ]+PRIVATE KEY-----`)
)

func NormalizeDisplaySafeRef(raw string) (DisplaySafeRef, bool) {
	ref := strings.TrimSpace(raw)
	if ref == "" {
		return "", false
	}
	if ContainsUnsafeRawOutput(ref) {
		return "", false
	}
	if !displaySafeRefPattern.MatchString(ref) {
		return "", false
	}
	return DisplaySafeRef(ref), true
}

func NormalizeAttemptRef(raw string) (AttemptRef, bool) {
	ref, ok := NormalizeDisplaySafeRef(raw)
	if !ok {
		return "", false
	}
	return AttemptRef(ref), true
}

func DisplaySafeRefs(raw []string) []DisplaySafeRef {
	out := make([]DisplaySafeRef, 0, len(raw))
	seen := map[DisplaySafeRef]struct{}{}
	for _, value := range raw {
		ref, ok := NormalizeDisplaySafeRef(value)
		if !ok {
			continue
		}
		if _, exists := seen[ref]; exists {
			continue
		}
		seen[ref] = struct{}{}
		out = append(out, ref)
	}
	return out
}

func ContainsUnsafeRawOutput(values ...string) bool {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if urlPattern.MatchString(trimmed) ||
			unixPathPattern.MatchString(trimmed) ||
			windowsPathPattern.MatchString(trimmed) ||
			secretPattern.MatchString(trimmed) ||
			pemPattern.MatchString(trimmed) {
			return true
		}
	}
	return false
}

func VerifyDisplaySafeOnly(rawOutputLoaded bool, refs []string) VerificationResult {
	result := VerificationResult{
		ContractVersion: ContractVersion,
		Status:          VerificationSatisfied,
		Satisfied:       true,
		FailureClass:    FailureNone,
		RawOutputLoaded: rawOutputLoaded,
	}
	for _, raw := range refs {
		ref, ok := NormalizeDisplaySafeRef(raw)
		if !ok {
			result.Status = VerificationBlocked
			result.Satisfied = false
			result.FailureClass = FailureEvidenceWeak
			result.Boundaries = appendUniqueBoundary(result.Boundaries, "raw_output_not_allowed")
			result.MissingInputs = appendUniqueMissingInput(result.MissingInputs, "host:display_safe_refs")
			result.NextHostAction = "provide_display_safe_refs"
			continue
		}
		result.EvidenceRefs = append(result.EvidenceRefs, EvidenceRef{
			Ref:      ref,
			Kind:     "display_safe_ref",
			Strength: EvidenceAdequate,
			Source:   "host:control_plane",
		})
	}
	if rawOutputLoaded {
		result.Status = VerificationBlocked
		result.Satisfied = false
		result.FailureClass = FailureEvidenceWeak
		result.Boundaries = appendUniqueBoundary(result.Boundaries, "raw_output_not_allowed")
		result.MissingInputs = appendUniqueMissingInput(result.MissingInputs, "host:display_safe_refs")
		result.NextHostAction = "provide_display_safe_refs"
	}
	return result.Normalize()
}

func NormalizeNextHostAction(raw string) NextHostAction {
	return NextHostAction(normalizeControlToken(raw))
}

func defaultContractVersion(value string) string {
	if strings.TrimSpace(value) == "" {
		return ContractVersion
	}
	return strings.TrimSpace(value)
}

func normalizeEnumToken(raw string) string {
	token := strings.ToLower(strings.TrimSpace(raw))
	token = strings.ReplaceAll(token, "-", "_")
	token = strings.ReplaceAll(token, " ", "_")
	for strings.Contains(token, "__") {
		token = strings.ReplaceAll(token, "__", "_")
	}
	return strings.Trim(token, "_")
}

func normalizeControlToken(raw string) string {
	token := normalizeEnumToken(raw)
	if token == "" || ContainsUnsafeRawOutput(token) {
		return ""
	}
	if !displaySafeRefPattern.MatchString(token) {
		return ""
	}
	return token
}

func normalizeStringList(in []string) []string {
	out := make([]string, 0, len(in))
	seen := map[string]struct{}{}
	for _, value := range in {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" || ContainsUnsafeRawOutput(trimmed) {
			continue
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	return out
}

func normalizeDisplaySafeRefs(in []DisplaySafeRef) []DisplaySafeRef {
	out := make([]DisplaySafeRef, 0, len(in))
	seen := map[DisplaySafeRef]struct{}{}
	for _, value := range in {
		ref, ok := NormalizeDisplaySafeRef(string(value))
		if !ok {
			continue
		}
		if _, exists := seen[ref]; exists {
			continue
		}
		seen[ref] = struct{}{}
		out = append(out, ref)
	}
	return out
}

func normalizeEvidenceRefs(in []EvidenceRef) []EvidenceRef {
	out := make([]EvidenceRef, 0, len(in))
	seen := map[string]struct{}{}
	for _, value := range in {
		ref := value.Normalize()
		if ref.Ref == "" {
			continue
		}
		key := string(ref.Ref) + "|" + ref.Kind + "|" + string(ref.Source)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, ref)
	}
	return out
}

func normalizeObservations(in []Observation) []Observation {
	out := make([]Observation, 0, len(in))
	for _, value := range in {
		normalized := value.Normalize()
		if normalized.Kind == "" && len(normalized.EvidenceRefs) == 0 && len(normalized.DisplaySafeRefs) == 0 {
			continue
		}
		out = append(out, normalized)
	}
	return out
}

func normalizeAttemptSummaries(in []AttemptSummary) []AttemptSummary {
	out := make([]AttemptSummary, 0, len(in))
	for _, value := range in {
		normalized := value.Normalize()
		if normalized.Ref == "" && normalized.Status == VerificationNotEvaluated && len(normalized.EvidenceRefs) == 0 {
			continue
		}
		out = append(out, normalized)
	}
	return out
}

func normalizeStrategyCandidates(in []StrategyCandidate) []StrategyCandidate {
	out := make([]StrategyCandidate, 0, len(in))
	seen := map[string]struct{}{}
	for _, value := range in {
		normalized := value.Normalize()
		if normalized.ID == "" {
			continue
		}
		if _, exists := seen[normalized.ID]; exists {
			continue
		}
		seen[normalized.ID] = struct{}{}
		out = append(out, normalized)
	}
	return out
}

func normalizeBoundaries(in []Boundary) []Boundary {
	out := make([]Boundary, 0, len(in))
	seen := map[Boundary]struct{}{}
	for _, value := range in {
		token := normalizeControlToken(string(value))
		if token == "" {
			continue
		}
		boundary := Boundary(token)
		if _, exists := seen[boundary]; exists {
			continue
		}
		seen[boundary] = struct{}{}
		out = append(out, boundary)
	}
	return out
}

func normalizeMissingInputs(in []MissingInput) []MissingInput {
	out := make([]MissingInput, 0, len(in))
	seen := map[MissingInput]struct{}{}
	for _, value := range in {
		token := normalizeControlToken(string(value))
		if token == "" {
			continue
		}
		missing := MissingInput(token)
		if _, exists := seen[missing]; exists {
			continue
		}
		seen[missing] = struct{}{}
		out = append(out, missing)
	}
	return out
}

func appendUniqueBoundary(in []Boundary, raw string) []Boundary {
	normalized := normalizeBoundaries(append(cloneBoundaries(in), Boundary(raw)))
	return normalized
}

func appendUniqueMissingInput(in []MissingInput, raw string) []MissingInput {
	normalized := normalizeMissingInputs(append(cloneMissingInputs(in), MissingInput(raw)))
	return normalized
}

func cloneStringSlice(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	return append([]string(nil), in...)
}

func cloneDisplaySafeRefs(in []DisplaySafeRef) []DisplaySafeRef {
	if len(in) == 0 {
		return nil
	}
	return append([]DisplaySafeRef(nil), in...)
}

func cloneEvidenceRefs(in []EvidenceRef) []EvidenceRef {
	if len(in) == 0 {
		return nil
	}
	return append([]EvidenceRef(nil), in...)
}

func cloneBoundaries(in []Boundary) []Boundary {
	if len(in) == 0 {
		return nil
	}
	return append([]Boundary(nil), in...)
}

func cloneMissingInputs(in []MissingInput) []MissingInput {
	if len(in) == 0 {
		return nil
	}
	return append([]MissingInput(nil), in...)
}

func cloneObservations(in []Observation) []Observation {
	if len(in) == 0 {
		return nil
	}
	out := make([]Observation, len(in))
	for i := range in {
		out[i] = CloneObservation(in[i])
	}
	return out
}

func cloneAttemptSummaries(in []AttemptSummary) []AttemptSummary {
	if len(in) == 0 {
		return nil
	}
	out := make([]AttemptSummary, len(in))
	for i := range in {
		out[i] = CloneAttemptSummary(in[i])
	}
	return out
}

func cloneStrategyCandidates(in []StrategyCandidate) []StrategyCandidate {
	if len(in) == 0 {
		return nil
	}
	out := make([]StrategyCandidate, len(in))
	for i, value := range in {
		out[i] = value.Clone()
	}
	return out
}
