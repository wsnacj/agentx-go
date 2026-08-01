package controlcontract

import (
	"sort"
	"strings"
	"time"
)

const (
	defaultHostActionKind = "host_action"
)

var requiredManagedObjectivePolicyRefs = []DisplaySafeRef{
	"contract:intensity_gate",
	"contract:budget",
	"contract:approval_policy",
	"contract:strategy_scope",
	"contract:redaction_policy",
}

type ManagedObjectiveProjectionInput struct {
	Activation          Activation       `json:"activation,omitempty"`
	Frame               ObjectiveFrame   `json:"frame,omitempty"`
	LedgerRef           DisplaySafeRef   `json:"ledger_ref,omitempty"`
	Attempts            []AttemptSummary `json:"attempts,omitempty"`
	Approved            bool             `json:"approved"`
	ApprovalRefs        []DisplaySafeRef `json:"approval_refs,omitempty"`
	PolicyRefs          []DisplaySafeRef `json:"policy_refs,omitempty"`
	AllowedStrategyRefs []DisplaySafeRef `json:"allowed_strategy_refs,omitempty"`
	EvidenceRefs        []EvidenceRef    `json:"evidence_refs,omitempty"`
	AdditionalInputs    []MissingInput   `json:"additional_inputs,omitempty"`
	Boundaries          []Boundary       `json:"boundaries,omitempty"`
	RawOutputLoaded     bool             `json:"raw_output_loaded"`
}

type ManagedObjectiveReplannerInput struct {
	Activation        Activation                 `json:"activation,omitempty"`
	ManagedObjective  ManagedObjectiveProjection `json:"managed_objective,omitempty"`
	Verification      VerificationResult         `json:"verification,omitempty"`
	AllowedStrategies []StrategyCandidate        `json:"allowed_strategies,omitempty"`
	EvidenceRefs      []EvidenceRef              `json:"evidence_refs,omitempty"`
	DecisionBasis     []DisplaySafeRef           `json:"decision_basis,omitempty"`
	Boundaries        []Boundary                 `json:"boundaries,omitempty"`
	RawOutputLoaded   bool                       `json:"raw_output_loaded"`
}

type ReplannerSourceInput struct {
	SourceKind           ReplannerSourceKind `json:"source_kind,omitempty"`
	SourceRef            DisplaySafeRef      `json:"source_ref,omitempty"`
	Producer             DisplaySafeRef      `json:"producer,omitempty"`
	Status               VerificationStatus  `json:"status,omitempty"`
	FailureClass         FailureClass        `json:"failure_class,omitempty"`
	FailureReason        string              `json:"failure_reason,omitempty"`
	CandidateStrategyRef DisplaySafeRef      `json:"candidate_strategy_ref,omitempty"`
	CapabilityRefs       []DisplaySafeRef    `json:"capability_refs,omitempty"`
	MinIntensity         ExecutionIntensity  `json:"min_intensity,omitempty"`
	MaxIntensity         ExecutionIntensity  `json:"max_intensity,omitempty"`
	ApprovalRefs         []DisplaySafeRef    `json:"approval_refs,omitempty"`
	EvidenceRefs         []EvidenceRef       `json:"evidence_refs,omitempty"`
	SourceRefs           []DisplaySafeRef    `json:"source_refs,omitempty"`
	ProposalRefs         []DisplaySafeRef    `json:"proposal_refs,omitempty"`
	MissingInputs        []MissingInput      `json:"missing_inputs,omitempty"`
	Boundaries           []Boundary          `json:"boundaries,omitempty"`
	NextHostAction       NextHostAction      `json:"next_host_action,omitempty"`
	RawOutputLoaded      bool                `json:"raw_output_loaded"`
}

func BuildManagedObjectiveProjection(input ManagedObjectiveProjectionInput) ManagedObjectiveProjection {
	frame := normalizeManagedObjectiveFrame(input.Frame)
	ledger := AttemptLedgerPatch{
		ObjectiveID:     frame.ID,
		LedgerRef:       input.LedgerRef,
		Attempts:        cloneAttemptSummaries(input.Attempts),
		EvidenceRefs:    cloneEvidenceRefs(input.EvidenceRefs),
		Boundaries:      []Boundary{"managed_objective_ledger_contract", "host_owned_durable_write"},
		RawOutputLoaded: input.RawOutputLoaded,
	}.Normalize()
	policyRefs := normalizeDisplaySafeRefs(input.PolicyRefs)
	allowedStrategyRefs := normalizeDisplaySafeRefs(input.AllowedStrategyRefs)
	if len(allowedStrategyRefs) == 0 {
		allowedStrategyRefs = normalizeDisplaySafeRefs(frame.CandidateCapabilities)
	} else {
		frame.CandidateCapabilities = normalizeDisplaySafeRefs(append(cloneDisplaySafeRefs(frame.CandidateCapabilities), allowedStrategyRefs...))
	}
	projection := ManagedObjectiveProjection{
		ContractVersion:     ContractVersion,
		Projected:           true,
		Activation:          NormalizeActivation(string(input.Activation)),
		Status:              HostActionNotReady,
		Frame:               frame,
		Ledger:              ledger,
		RequiresApproval:    true,
		ApprovalRefs:        normalizeDisplaySafeRefs(input.ApprovalRefs),
		PolicyRefs:          policyRefs,
		AllowedStrategyRefs: allowedStrategyRefs,
		EvidenceRefs:        normalizeEvidenceRefs(input.EvidenceRefs),
		FailureClass:        FailureNone,
		MissingInputs:       normalizeMissingInputs(input.AdditionalInputs),
		Boundaries: MergeBoundaries(
			[]Boundary{
				"managed_objective_projection_only",
				"diagnostics_only",
				"no_strategy_dispatch",
				"host_owned_ledger_write",
				"model_route_is_not_authorization",
			},
			input.Boundaries,
		),
		RunnerEffect:    "none",
		PromptEffect:    "none",
		RawOutputLoaded: input.RawOutputLoaded,
	}
	if projection.Activation != ActivationManaged {
		projection.FailureClass = FailurePolicyBlocked
		projection.MissingInputs = AppendMissingInputs(projection.MissingInputs, "control_plane:managed_activation")
		projection.Boundaries = AppendBoundaries(projection.Boundaries, "managed_objective_activation_required")
		projection.NextHostAction = "enable_managed_objective"
		return projection.Normalize()
	}
	projection.MissingInputs = appendManagedObjectiveFrameMissingInputs(projection.MissingInputs, frame, ledger)
	projection.MissingInputs = appendManagedObjectiveStrategyMissingInputs(projection.MissingInputs, allowedStrategyRefs)
	projection.MissingInputs = appendManagedObjectivePolicyMissingInputs(projection.MissingInputs, policyRefs)
	if len(projection.MissingInputs) > 0 {
		projection.FailureClass = FailureConfigMissing
		projection.NextHostAction = "provide_managed_objective_contract"
		return projection.Normalize()
	}
	if !input.Approved {
		projection.Status = HostActionRequiresApproval
		projection.FailureClass = FailureApprovalRequired
		projection.MissingInputs = AppendMissingInputs(projection.MissingInputs, "host:managed_objective_approval")
		projection.Boundaries = AppendBoundaries(projection.Boundaries, "managed_objective_requires_host_approval")
		projection.NextHostAction = "request_host_approval"
		return projection.Normalize()
	}
	if len(projection.ApprovalRefs) == 0 {
		projection.Status = HostActionReviewRequired
		projection.FailureClass = FailureEvidenceMissing
		projection.MissingInputs = AppendMissingInputs(projection.MissingInputs, "host:approval_ref")
		projection.Boundaries = AppendBoundaries(projection.Boundaries, "managed_objective_approval_ref_missing")
		projection.NextHostAction = "provide_host_approval_ref"
		return projection.Normalize()
	}
	projection.Status = HostActionReady
	projection.Boundaries = AppendBoundaries(projection.Boundaries, "managed_objective_contract_ready")
	projection.NextHostAction = "host_may_plan_managed_objective"
	return projection.Normalize()
}

func BuildManagedObjectiveReplannerProjection(input ManagedObjectiveReplannerInput) ManagedObjectiveReplannerProjection {
	managed := input.ManagedObjective.Normalize()
	verification := input.Verification.Normalize()
	evidenceRefs := normalizeEvidenceRefs(append(cloneEvidenceRefs(input.EvidenceRefs), verification.EvidenceRefs...))
	projection := ManagedObjectiveReplannerProjection{
		ContractVersion: ContractVersion,
		Projected:       true,
		Activation:      NormalizeActivation(string(input.Activation)),
		Status:          "not_candidate",
		Mode:            "l3_managed_objective_replanner",
		ObjectiveID:     managed.Frame.ID,
		Verification:    verification,
		EvidenceRefs:    evidenceRefs,
		FailureClass:    FailureNone,
		DecisionBasis:   normalizeDisplaySafeRefs(input.DecisionBasis),
		Boundaries: MergeBoundaries(
			[]Boundary{
				"diagnostics_only",
				"proposal_only",
				"no_runner_dispatch",
				"no_strategy_dispatch",
				"host_must_apply_strategy",
				"model_route_is_not_authorization",
			},
			input.Boundaries,
		),
		RunnerEffect:    "none",
		PromptEffect:    "none",
		RawOutputLoaded: input.RawOutputLoaded || managed.RawOutputLoaded || verification.RawOutputLoaded,
	}
	if projection.Activation != ActivationManaged || managed.Activation != ActivationManaged {
		projection.Status = "not_candidate"
		projection.FailureClass = FailurePolicyBlocked
		projection.MissingInputs = AppendMissingInputs(projection.MissingInputs, "control_plane:managed_activation")
		projection.DecisionBasis = appendUniqueDisplaySafeRef(projection.DecisionBasis, "managed_objective_activation_required")
		projection.NextHostAction = "enable_managed_objective"
		return projection.Normalize()
	}
	if !managed.Ready {
		projection.Status = "not_candidate"
		projection.FailureClass = firstFailureClass(managed.FailureClass, FailureConfigMissing)
		projection.MissingInputs = AppendMissingInputs(projection.MissingInputs, managed.MissingInputs...)
		projection.DecisionBasis = appendUniqueDisplaySafeRef(projection.DecisionBasis, "managed_objective_not_ready")
		projection.NextHostAction = "provide_managed_objective_contract"
		return projection.Normalize()
	}
	projection.MissingInputs = AppendMissingInputs(projection.MissingInputs, managed.MissingInputs...)
	status := verification.Status
	projection.FailureClass = firstFailureClass(verification.FailureClass, FailureNone)
	projection.BlockedReason = managedObjectiveReplannerSafeReason(verification.FailureReason)
	projection.DecisionBasis = appendUniqueDisplaySafeRef(projection.DecisionBasis, DisplaySafeRef("verification_status:"+string(status)))
	switch status {
	case VerificationSatisfied, VerificationNotApplicable:
		projection.Status = "no_action"
		projection.NextHostAction = "none"
		return projection.Normalize()
	case VerificationNotEvaluated:
		projection.Status = "not_candidate"
		projection.FailureClass = FailureInsufficientInformation
		projection.MissingInputs = AppendMissingInputs(projection.MissingInputs, "host:verification_result")
		projection.NextHostAction = "provide_verification_result"
		return projection.Normalize()
	}
	if managedObjectiveReplannerPolicyBlocked(verification.FailureClass) {
		projection.Status = "blocked"
		projection.FailureClass = firstFailureClass(verification.FailureClass, FailurePolicyBlocked)
		projection.BlockedReason = firstNonEmptyContractString(projection.BlockedReason, string(projection.FailureClass))
		projection.DecisionBasis = appendUniqueDisplaySafeRef(projection.DecisionBasis, "policy_or_budget_blocked")
		projection.NextHostAction = "request_host_policy_or_budget_review"
		return projection.Normalize()
	}
	candidates := managedObjectiveReplannerCandidates(managed, input.AllowedStrategies)
	if len(candidates) == 0 {
		projection.Status = "blocked"
		projection.FailureClass = FailureConfigMissing
		projection.MissingInputs = AppendMissingInputs(projection.MissingInputs, "host:strategy_scope")
		projection.DecisionBasis = appendUniqueDisplaySafeRef(projection.DecisionBasis, "strategy_scope_missing")
		projection.NextHostAction = "provide_managed_objective_contract"
		return projection.Normalize()
	}
	switch status {
	case VerificationPartial, VerificationFailed, VerificationBlocked, VerificationReviewRequired:
		projection.Status = "candidate"
		projection.Candidates = candidates
		projection.DecisionBasis = appendUniqueDisplaySafeRef(projection.DecisionBasis, "managed_replanner_candidate")
		projection.NextHostAction = "request_host_replanner_decision"
		projection.Proposal = HostActionProposal{
			Kind:             "managed_objective_replan",
			Status:           HostActionRequiresApproval,
			RequiresApproval: true,
			ApprovalRefs:     cloneDisplaySafeRefs(managed.ApprovalRefs),
			ActionRefs:       managedObjectiveReplannerActionRefs(candidates),
			EvidenceRefs:     cloneEvidenceRefs(evidenceRefs),
			FailureClass:     firstFailureClass(verification.FailureClass, FailureVerificationFailed),
			FailureReason:    managedObjectiveReplannerSafeReason(projection.BlockedReason),
			Boundaries: []Boundary{
				"proposal_only",
				"host_must_approve_strategy",
				"runner_does_not_execute_replan",
			},
			NextHostAction: "request_host_replanner_decision",
		}.Normalize()
		return projection.Normalize()
	default:
		projection.Status = "not_candidate"
		projection.DecisionBasis = appendUniqueDisplaySafeRef(projection.DecisionBasis, "unsupported_verification_status")
		projection.NextHostAction = "return_current_result"
		return projection.Normalize()
	}
}

func BuildReplannerSourceProjection(input ReplannerSourceInput) ReplannerSourceProjection {
	sourceKind := NormalizeReplannerSourceKind(string(input.SourceKind))
	sourceRef, _ := NormalizeDisplaySafeRef(string(input.SourceRef))
	producer, _ := NormalizeDisplaySafeRef(string(input.Producer))
	candidateRef, hasCandidate := NormalizeDisplaySafeRef(string(input.CandidateStrategyRef))
	failure := NormalizeFailureClass(string(input.FailureClass))
	status := replannerSourceStatus(input.Status, failure)
	next := firstNextHostAction(input.NextHostAction, replannerSourceDefaultNextHostAction(status))
	evidenceRefs := normalizeEvidenceRefs(input.EvidenceRefs)
	sourceRefs := normalizeDisplaySafeRefs(append(append(cloneDisplaySafeRefs(input.SourceRefs), sourceRef), producer))
	proposalRefs := normalizeDisplaySafeRefs(input.ProposalRefs)
	capabilityRefs := normalizeDisplaySafeRefs(input.CapabilityRefs)
	displaySafeRefs := replannerSourceDisplaySafeRefs(sourceRefs, proposalRefs, capabilityRefs, []DisplaySafeRef{candidateRef})
	boundaries := MergeBoundaries(
		[]Boundary{
			"replanner_source_projection",
			"proposal_only",
			"display_safe_refs_only",
			"no_runner_dispatch",
			"no_apply_or_dispatch_implementation",
			"core_must_not_parse_scene_payload",
		},
		[]Boundary{replannerSourceBoundary(sourceKind)},
		input.Boundaries,
	)
	projection := ReplannerSourceProjection{
		ContractVersion: ContractVersion,
		Projected:       true,
		SourceKind:      sourceKind,
		SourceRef:       sourceRef,
		Producer:        producer,
		ControlMode:     replannerSourceControlMode(sourceKind),
		Status:          status,
		FailureClass:    failure,
		EvidenceRefs:    evidenceRefs,
		SourceRefs:      sourceRefs,
		ProposalRefs:    proposalRefs,
		MissingInputs:   normalizeMissingInputs(input.MissingInputs),
		Boundaries:      boundaries,
		NextHostAction:  next,
		RunnerEffect:    "none",
		PromptEffect:    "none",
		RawOutputLoaded: input.RawOutputLoaded,
	}
	if sourceKind == "" {
		projection.Status = VerificationReviewRequired
		projection.FailureClass = FailureInvalidInput
		projection.MissingInputs = AppendMissingInputs(projection.MissingInputs, "controlplane:replanner_source_kind")
		projection.NextHostAction = "provide_replanner_source_kind"
		projection.Observation = Observation{
			Kind:              "replanner_source",
			Source:            replannerSourcePrimaryRef(sourceRef, producer),
			Name:              "unknown_source_status",
			Value:             string(projection.Status),
			Strength:          EvidenceWeak,
			EvidenceRefs:      evidenceRefs,
			DisplaySafeRefs:   displaySafeRefs,
			DegradationReason: replannerSourceDegradationReason(projection.FailureClass),
			RawOutputLoaded:   input.RawOutputLoaded,
		}
		projection.Verification = VerificationResult{
			Status:          projection.Status,
			FailureClass:    projection.FailureClass,
			EvidenceRefs:    evidenceRefs,
			MissingInputs:   cloneMissingInputs(projection.MissingInputs),
			Boundaries:      cloneBoundaries(projection.Boundaries),
			NextHostAction:  projection.NextHostAction,
			RawOutputLoaded: input.RawOutputLoaded,
		}
		return replannerSourceApplyDisplaySafeReviewIfNeeded(projection, input).Normalize()
	}
	if sourceRef == "" && producer == "" {
		projection.Status = VerificationReviewRequired
		if projection.FailureClass == FailureNone {
			projection.FailureClass = FailureInsufficientInformation
		}
		projection.MissingInputs = AppendMissingInputs(projection.MissingInputs, "host:replanner_source_ref")
		projection.NextHostAction = "provide_replanner_source_ref"
	}
	projection.Observation = Observation{
		Kind:              "replanner_source",
		Source:            replannerSourcePrimaryRef(sourceRef, producer),
		Subject:           candidateRef,
		Name:              string(sourceKind) + "_status",
		Value:             string(projection.Status),
		Strength:          replannerSourceEvidenceStrength(projection.Status, projection.FailureClass),
		EvidenceRefs:      evidenceRefs,
		DisplaySafeRefs:   displaySafeRefs,
		DegradationReason: replannerSourceDegradationReason(projection.FailureClass),
		RawOutputLoaded:   input.RawOutputLoaded,
	}
	projection.Verification = VerificationResult{
		Status:          projection.Status,
		Satisfied:       projection.Status == VerificationSatisfied,
		FailureClass:    projection.FailureClass,
		FailureReason:   firstNonEmptyContractString(input.FailureReason, string(projection.FailureClass)),
		EvidenceRefs:    evidenceRefs,
		MissingInputs:   cloneMissingInputs(projection.MissingInputs),
		Boundaries:      cloneBoundaries(projection.Boundaries),
		NextHostAction:  projection.NextHostAction,
		RawOutputLoaded: input.RawOutputLoaded,
	}
	if replannerSourceNeedsAction(projection.Status) {
		if hasCandidate {
			projection.Candidate = StrategyCandidate{
				ID:               string(candidateRef),
				Kind:             "replanner_source_strategy",
				ControlMode:      projection.ControlMode,
				MinIntensity:     replannerSourceIntensity(input.MinIntensity, IntensityL3ManagedObjective),
				MaxIntensity:     replannerSourceIntensity(input.MaxIntensity, replannerSourceIntensity(input.MinIntensity, IntensityL3ManagedObjective)),
				CapabilityRefs:   capabilityRefs,
				ExpectedEvidence: evidenceRefs,
				Preconditions:    cloneMissingInputs(projection.MissingInputs),
				RequiresApproval: true,
				Owner:            "host",
				Boundaries: []Boundary{
					"proposal_only",
					"host_declared_strategy_scope",
					"host_must_dispatch_strategy",
				},
			}.Normalize()
		}
		if !hasCandidate && len(proposalRefs) == 0 {
			projection.MissingInputs = AppendMissingInputs(projection.MissingInputs, "host:replanner_strategy_ref")
			projection.NextHostAction = "provide_replanner_strategy_ref"
			projection.Verification.MissingInputs = cloneMissingInputs(projection.MissingInputs)
			projection.Verification.NextHostAction = projection.NextHostAction
		}
		projection.Proposal = HostActionProposal{
			Kind:             replannerSourceProposalKind(sourceKind),
			Status:           HostActionRequiresApproval,
			RequiresApproval: true,
			ApprovalRefs:     normalizeDisplaySafeRefs(input.ApprovalRefs),
			ActionRefs:       replannerSourceProposalActionRefs(candidateRef, proposalRefs, capabilityRefs),
			EvidenceRefs:     evidenceRefs,
			FailureClass:     firstFailureClass(projection.FailureClass, FailureVerificationFailed),
			FailureReason:    firstNonEmptyContractString(input.FailureReason, string(projection.FailureClass)),
			MissingInputs:    cloneMissingInputs(projection.MissingInputs),
			Boundaries: []Boundary{
				"proposal_only",
				"host_must_approve_strategy",
				"runner_does_not_execute_replan",
				"no_apply_or_dispatch_implementation",
			},
			NextHostAction: projection.NextHostAction,
		}.Normalize()
	}
	return replannerSourceApplyDisplaySafeReviewIfNeeded(projection, input).Normalize()
}

func managedObjectiveReplannerPolicyBlocked(failure FailureClass) bool {
	switch NormalizeFailureClass(string(failure)) {
	case FailureBudgetExhausted,
		FailureApprovalRequired,
		FailurePermissionDenied,
		FailurePolicyBlocked,
		FailureAuthorizationMissing,
		FailureCredentialMissing:
		return true
	default:
		return false
	}
}

func managedObjectiveReplannerCandidates(managed ManagedObjectiveProjection, configured []StrategyCandidate) []StrategyCandidate {
	allowedRefs := normalizeDisplaySafeRefs(managed.AllowedStrategyRefs)
	if len(allowedRefs) == 0 {
		return nil
	}
	if len(configured) > 0 {
		out := []StrategyCandidate{}
		for _, candidate := range normalizeStrategyCandidates(configured) {
			ref, ok := NormalizeDisplaySafeRef(candidate.ID)
			if !ok || !displaySafeRefSliceContains(allowedRefs, ref) {
				continue
			}
			out = append(out, candidate)
		}
		if len(out) > 0 {
			return out
		}
	}
	out := make([]StrategyCandidate, 0, len(allowedRefs))
	for _, ref := range allowedRefs {
		out = append(out, StrategyCandidate{
			ID:               string(ref),
			Kind:             "managed_objective_strategy",
			ControlMode:      ControlModeObjective,
			MinIntensity:     IntensityL3ManagedObjective,
			MaxIntensity:     IntensityL3ManagedObjective,
			CapabilityRefs:   []DisplaySafeRef{ref},
			RequiresApproval: true,
			Owner:            "host",
			Boundaries: []Boundary{
				"proposal_only",
				"host_declared_strategy_scope",
				"host_must_dispatch_strategy",
			},
		}.Normalize())
	}
	return out
}

func managedObjectiveReplannerActionRefs(candidates []StrategyCandidate) []DisplaySafeRef {
	out := []DisplaySafeRef{}
	for _, candidate := range normalizeStrategyCandidates(candidates) {
		ref, ok := NormalizeDisplaySafeRef(candidate.ID)
		if !ok {
			continue
		}
		out = append(out, ref)
	}
	return normalizeDisplaySafeRefs(out)
}

func replannerSourceStatus(status VerificationStatus, failure FailureClass) VerificationStatus {
	normalized := NormalizeVerificationStatus(string(status))
	if normalized != VerificationNotEvaluated || NormalizeFailureClass(string(failure)) == FailureNone {
		return normalized
	}
	switch NormalizeFailureClass(string(failure)) {
	case FailureEvidenceMissing, FailureEvidenceWeak, FailureVerificationFailed, FailureRepeatedNoProgress:
		return VerificationPartial
	case FailureToolUnavailable,
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
		FailureTargetUnavailable,
		FailureTargetNotFound,
		FailureExternalDependencyUnavailable,
		FailureUnsupportedOperation:
		return VerificationBlocked
	default:
		return VerificationReviewRequired
	}
}

func replannerSourceControlMode(sourceKind ReplannerSourceKind) ControlMode {
	switch NormalizeReplannerSourceKind(string(sourceKind)) {
	case ReplannerSourceOperations:
		return ControlModeOperations
	case ReplannerSourceCapability:
		return ControlModeCapabilityResolution
	case ReplannerSourceWorkflow:
		return ControlModeWorkflow
	case ReplannerSourceRecovery:
		return ControlModeObjective
	default:
		return ""
	}
}

func replannerSourceBoundary(sourceKind ReplannerSourceKind) Boundary {
	switch NormalizeReplannerSourceKind(string(sourceKind)) {
	case ReplannerSourceOperations:
		return "operations_source_projection"
	case ReplannerSourceCapability:
		return "capability_source_projection"
	case ReplannerSourceWorkflow:
		return "workflow_source_projection"
	case ReplannerSourceRecovery:
		return "recovery_source_projection"
	default:
		return "replanner_source_kind_missing"
	}
}

func replannerSourceProposalKind(sourceKind ReplannerSourceKind) string {
	switch NormalizeReplannerSourceKind(string(sourceKind)) {
	case ReplannerSourceOperations:
		return "operations_replanner_source_proposal"
	case ReplannerSourceCapability:
		return "capability_resolution_proposal"
	case ReplannerSourceWorkflow:
		return "workflow_strategy_proposal"
	case ReplannerSourceRecovery:
		return "objective_recovery_proposal"
	default:
		return "replanner_source_proposal"
	}
}

func replannerSourceDefaultNextHostAction(status VerificationStatus) NextHostAction {
	switch NormalizeVerificationStatus(string(status)) {
	case VerificationSatisfied, VerificationNotApplicable:
		return "none"
	case VerificationNotEvaluated:
		return "provide_verification_result"
	default:
		return "request_host_replanner_decision"
	}
}

func replannerSourceNeedsAction(status VerificationStatus) bool {
	switch NormalizeVerificationStatus(string(status)) {
	case VerificationPartial, VerificationBlocked, VerificationReviewRequired, VerificationFailed:
		return true
	default:
		return false
	}
}

func replannerSourceEvidenceStrength(status VerificationStatus, failure FailureClass) EvidenceStrength {
	if NormalizeFailureClass(string(failure)) == FailureEvidenceWeak {
		return EvidenceWeak
	}
	switch NormalizeVerificationStatus(string(status)) {
	case VerificationSatisfied:
		return EvidenceStrong
	case VerificationPartial, VerificationBlocked, VerificationFailed, VerificationReviewRequired:
		return EvidenceWeak
	case VerificationNotApplicable:
		return EvidenceAdequate
	default:
		return EvidenceMissing
	}
}

func replannerSourcePrimaryRef(sourceRef DisplaySafeRef, producer DisplaySafeRef) DisplaySafeRef {
	if sourceRef != "" {
		return sourceRef
	}
	return producer
}

func replannerSourceIntensity(value ExecutionIntensity, fallback ExecutionIntensity) ExecutionIntensity {
	normalized := NormalizeExecutionIntensity(string(value))
	if normalized != "" {
		return normalized
	}
	return fallback
}

func replannerSourceDisplaySafeRefs(groups ...[]DisplaySafeRef) []DisplaySafeRef {
	out := []DisplaySafeRef{}
	for _, group := range groups {
		out = append(out, group...)
	}
	return normalizeDisplaySafeRefs(out)
}

func replannerSourceProposalActionRefs(candidateRef DisplaySafeRef, proposalRefs []DisplaySafeRef, capabilityRefs []DisplaySafeRef) []DisplaySafeRef {
	refs := []DisplaySafeRef{}
	if candidateRef != "" {
		refs = append(refs, candidateRef)
	}
	refs = append(refs, proposalRefs...)
	refs = append(refs, capabilityRefs...)
	return normalizeDisplaySafeRefs(refs)
}

func replannerSourceApplyDisplaySafeReviewIfNeeded(projection ReplannerSourceProjection, input ReplannerSourceInput) ReplannerSourceProjection {
	if !replannerSourceUnsafeInput(input) {
		return projection
	}
	projection.Status = VerificationReviewRequired
	if projection.FailureClass == FailureNone {
		projection.FailureClass = FailureEvidenceWeak
	}
	projection.Boundaries = AppendBoundaries(projection.Boundaries, "raw_output_not_allowed")
	projection.MissingInputs = AppendMissingInputs(projection.MissingInputs, "host:display_safe_refs")
	projection.NextHostAction = "provide_display_safe_refs"
	projection.Observation.Value = string(projection.Status)
	projection.Observation.Strength = EvidenceWeak
	projection.Observation.DegradationReason = replannerSourceDegradationReason(projection.FailureClass)
	projection.Verification = RequireVerificationReview(projection.Verification, ReviewRequiredInput{
		FailureClass:   projection.FailureClass,
		Boundary:       "raw_output_not_allowed",
		MissingInput:   "host:display_safe_refs",
		NextHostAction: "provide_display_safe_refs",
	})
	if projection.Proposal.Kind != "" {
		projection.Proposal = RequireHostActionReview(projection.Proposal, ReviewRequiredInput{
			FailureClass:   projection.FailureClass,
			Boundary:       "raw_output_not_allowed",
			MissingInput:   "host:display_safe_refs",
			NextHostAction: "provide_display_safe_refs",
		})
	}
	return projection
}

func replannerSourceDegradationReason(failure FailureClass) string {
	normalized := NormalizeFailureClass(string(failure))
	if normalized == FailureNone {
		return ""
	}
	return string(normalized)
}

func replannerSourceUnsafeInput(input ReplannerSourceInput) bool {
	return displaySafeRefRejected(input.SourceRef) ||
		displaySafeRefRejected(input.Producer) ||
		displaySafeRefRejected(input.CandidateStrategyRef) ||
		displaySafeRefSliceRejected(input.CapabilityRefs) ||
		displaySafeRefSliceRejected(input.ApprovalRefs) ||
		displaySafeRefSliceRejected(input.SourceRefs) ||
		displaySafeRefSliceRejected(input.ProposalRefs) ||
		evidenceRefRejected(input.EvidenceRefs)
}

func displaySafeRefRejected(value DisplaySafeRef) bool {
	raw := strings.TrimSpace(string(value))
	if raw == "" {
		return false
	}
	_, ok := NormalizeDisplaySafeRef(raw)
	return !ok
}

func displaySafeRefSliceRejected(values []DisplaySafeRef) bool {
	for _, value := range values {
		if displaySafeRefRejected(value) {
			return true
		}
	}
	return false
}

func evidenceRefRejected(values []EvidenceRef) bool {
	for _, value := range values {
		if displaySafeRefRejected(value.Ref) || displaySafeRefRejected(value.Source) {
			return true
		}
	}
	return false
}

func firstFailureClass(values ...FailureClass) FailureClass {
	for _, value := range values {
		normalized := NormalizeFailureClass(string(value))
		if normalized != FailureNone {
			return normalized
		}
	}
	return FailureNone
}

func firstNonEmptyContractString(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" || ContainsUnsafeRawOutput(trimmed) {
			continue
		}
		return trimmed
	}
	return ""
}

func managedObjectiveReplannerSafeReason(value string) string {
	reason := firstNonEmptyContractString(value)
	if len(reason) > 160 {
		reason = strings.TrimSpace(reason[:160])
	}
	return reason
}

func appendUniqueDisplaySafeRef(in []DisplaySafeRef, raw DisplaySafeRef) []DisplaySafeRef {
	ref, ok := NormalizeDisplaySafeRef(string(raw))
	if !ok {
		return normalizeDisplaySafeRefs(in)
	}
	return normalizeDisplaySafeRefs(append(cloneDisplaySafeRefs(in), ref))
}

func appendManagedObjectiveStrategyMissingInputs(base []MissingInput, allowedStrategyRefs []DisplaySafeRef) []MissingInput {
	if len(allowedStrategyRefs) > 0 {
		return base
	}
	return AppendMissingInputs(base, "host:strategy_scope")
}

func normalizeManagedObjectiveFrame(frame ObjectiveFrame) ObjectiveFrame {
	out := frame.Normalize()
	if out.ControlMode == "" {
		out.ControlMode = ControlModeObjective
	}
	if out.Intensity == "" {
		out.Intensity = IntensityL3ManagedObjective
	}
	out.Boundaries = AppendBoundaries(out.Boundaries, "managed_objective_frame")
	return out.Normalize()
}

func appendManagedObjectiveFrameMissingInputs(base []MissingInput, frame ObjectiveFrame, ledger AttemptLedgerPatch) []MissingInput {
	out := base
	if strings.TrimSpace(frame.ID) == "" {
		out = AppendMissingInputs(out, "control_plane:objective_frame")
	}
	if strings.TrimSpace(frame.UserGoalDigest) == "" {
		out = AppendMissingInputs(out, "control_plane:user_goal_digest")
	}
	if len(frame.SuccessCriteria) == 0 {
		out = AppendMissingInputs(out, "host:success_criteria")
	}
	if ledger.LedgerRef == "" {
		out = AppendMissingInputs(out, "host:managed_objective_ledger_ref")
	}
	return out
}

func appendManagedObjectivePolicyMissingInputs(base []MissingInput, policyRefs []DisplaySafeRef) []MissingInput {
	out := base
	for _, required := range requiredManagedObjectivePolicyRefs {
		if !displaySafeRefSliceContains(policyRefs, required) {
			out = AppendMissingInputs(out, MissingInput(required))
		}
	}
	return out
}

func displaySafeRefSliceContains(values []DisplaySafeRef, needle DisplaySafeRef) bool {
	for _, value := range normalizeDisplaySafeRefs(values) {
		if value == needle {
			return true
		}
	}
	return false
}

type ApprovalGateInput struct {
	Kind             string           `json:"kind,omitempty"`
	RequiresApproval bool             `json:"requires_approval"`
	Approved         bool             `json:"approved"`
	ApprovalRefs     []DisplaySafeRef `json:"approval_refs,omitempty"`
	ActionRefs       []DisplaySafeRef `json:"action_refs,omitempty"`
	EvidenceRefs     []EvidenceRef    `json:"evidence_refs,omitempty"`
	MissingInput     MissingInput     `json:"missing_input,omitempty"`
	RequiredBoundary Boundary         `json:"required_boundary,omitempty"`
	GrantedBoundary  Boundary         `json:"granted_boundary,omitempty"`
	NextHostAction   NextHostAction   `json:"next_host_action,omitempty"`
}

func EvaluateHostApprovalGate(input ApprovalGateInput) HostActionProposal {
	kind := normalizeControlToken(input.Kind)
	if kind == "" {
		kind = defaultHostActionKind
	}
	missingInput := firstMissingInput(input.MissingInput, "host:approval")
	requiredBoundary := firstBoundary(input.RequiredBoundary, "host_approval_required")
	grantedBoundary := firstBoundary(input.GrantedBoundary, "host_approval_granted")
	next := firstNextHostAction(input.NextHostAction, "request_host_approval")
	proposal := HostActionProposal{
		ContractVersion:  ContractVersion,
		Kind:             kind,
		RequiresApproval: input.RequiresApproval,
		ApprovalRefs:     normalizeDisplaySafeRefs(input.ApprovalRefs),
		ActionRefs:       normalizeDisplaySafeRefs(input.ActionRefs),
		EvidenceRefs:     normalizeEvidenceRefs(input.EvidenceRefs),
		FailureClass:     FailureNone,
	}
	if !input.RequiresApproval {
		proposal.Status = HostActionReady
		proposal.Boundaries = []Boundary{"host_approval_not_required"}
		proposal.NextHostAction = "host_may_continue"
		return proposal.Normalize()
	}
	if !input.Approved {
		proposal.Status = HostActionRequiresApproval
		proposal.FailureClass = FailureApprovalRequired
		proposal.MissingInputs = []MissingInput{missingInput}
		proposal.Boundaries = []Boundary{requiredBoundary}
		proposal.NextHostAction = next
		return proposal.Normalize()
	}
	if len(proposal.ApprovalRefs) == 0 {
		proposal.Status = HostActionReviewRequired
		proposal.FailureClass = FailureEvidenceMissing
		proposal.MissingInputs = []MissingInput{"host:approval_ref"}
		proposal.Boundaries = []Boundary{"host_approval_ref_missing"}
		proposal.NextHostAction = "provide_host_approval_ref"
		return proposal.Normalize()
	}
	proposal.Status = HostActionReady
	proposal.Boundaries = []Boundary{grantedBoundary}
	proposal.NextHostAction = "host_may_continue"
	return proposal.Normalize()
}

type BudgetGateInput struct {
	Limit          int            `json:"limit,omitempty"`
	Used           int            `json:"used,omitempty"`
	Increment      int            `json:"increment,omitempty"`
	Scope          DisplaySafeRef `json:"scope,omitempty"`
	EvidenceRefs   []EvidenceRef  `json:"evidence_refs,omitempty"`
	NextHostAction NextHostAction `json:"next_host_action,omitempty"`
}

type BudgetGateResult struct {
	ContractVersion      string             `json:"contract_version,omitempty"`
	Allowed              bool               `json:"allowed"`
	Status               VerificationStatus `json:"status,omitempty"`
	Scope                DisplaySafeRef     `json:"scope,omitempty"`
	RetryBudgetLimit     int                `json:"retry_budget_limit,omitempty"`
	RetryBudgetUsed      int                `json:"retry_budget_used,omitempty"`
	RetryBudgetRemaining int                `json:"retry_budget_remaining,omitempty"`
	AttemptIncrement     int                `json:"attempt_increment,omitempty"`
	FailureClass         FailureClass       `json:"failure_class,omitempty"`
	EvidenceRefs         []EvidenceRef      `json:"evidence_refs,omitempty"`
	Boundaries           []Boundary         `json:"boundaries,omitempty"`
	MissingInputs        []MissingInput     `json:"missing_inputs,omitempty"`
	NextHostAction       NextHostAction     `json:"next_host_action,omitempty"`
}

func EvaluateRetryBudgetGate(input BudgetGateInput) BudgetGateResult {
	limit := input.Limit
	used := input.Used
	increment := input.Increment
	if used < 0 {
		used = 0
	}
	if increment <= 0 {
		increment = 1
	}
	scope, _ := NormalizeDisplaySafeRef(string(input.Scope))
	result := BudgetGateResult{
		ContractVersion:  ContractVersion,
		Status:           VerificationBlocked,
		Scope:            scope,
		RetryBudgetLimit: limit,
		RetryBudgetUsed:  used,
		AttemptIncrement: increment,
		EvidenceRefs:     normalizeEvidenceRefs(input.EvidenceRefs),
		FailureClass:     FailureNone,
	}
	if limit <= 0 {
		result.RetryBudgetLimit = 0
		result.RetryBudgetRemaining = 0
		result.FailureClass = FailureConfigMissing
		result.MissingInputs = []MissingInput{"contract:retry_budget"}
		result.Boundaries = []Boundary{"retry_budget_not_configured"}
		result.NextHostAction = firstNextHostAction(input.NextHostAction, "provide_retry_budget_policy")
		return result.Normalize()
	}
	remaining := limit - used
	if remaining < 0 {
		remaining = 0
	}
	result.RetryBudgetRemaining = remaining
	if used+increment > limit {
		result.FailureClass = FailureBudgetExhausted
		result.Boundaries = []Boundary{"retry_budget_exhausted"}
		result.NextHostAction = firstNextHostAction(input.NextHostAction, "return_partial_or_request_upgrade")
		return result.Normalize()
	}
	result.Allowed = true
	result.Status = VerificationSatisfied
	result.RetryBudgetRemaining = limit - used - increment
	result.Boundaries = []Boundary{"retry_budget_available"}
	result.NextHostAction = firstNextHostAction(input.NextHostAction, "host_may_continue")
	return result.Normalize()
}

func CloneBudgetGateResult(in BudgetGateResult) BudgetGateResult {
	out := in
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	return out
}

func (r BudgetGateResult) Normalize() BudgetGateResult {
	out := CloneBudgetGateResult(r)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Status = NormalizeVerificationStatus(string(out.Status))
	out.Scope, _ = NormalizeDisplaySafeRef(string(out.Scope))
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	if out.RetryBudgetLimit < 0 {
		out.RetryBudgetLimit = 0
	}
	if out.RetryBudgetUsed < 0 {
		out.RetryBudgetUsed = 0
	}
	if out.RetryBudgetRemaining < 0 {
		out.RetryBudgetRemaining = 0
	}
	if out.AttemptIncrement < 0 {
		out.AttemptIncrement = 0
	}
	out.Allowed = out.Status == VerificationSatisfied
	return out
}

type IdempotencyOperation string

const (
	IdempotencyCreate     IdempotencyOperation = "create"
	IdempotencyReplay     IdempotencyOperation = "idempotent_replay"
	IdempotencyConflict   IdempotencyOperation = "idempotency_conflict"
	IdempotencyKeyMissing IdempotencyOperation = "idempotency_key_missing"
)

type IdempotencyCheckInput struct {
	RequestedKey         DisplaySafeRef `json:"requested_key,omitempty"`
	ExistingKey          DisplaySafeRef `json:"existing_key,omitempty"`
	RequestedFingerprint string         `json:"requested_fingerprint,omitempty"`
	ExistingFingerprint  string         `json:"existing_fingerprint,omitempty"`
	ExistingRef          DisplaySafeRef `json:"existing_ref,omitempty"`
	ActionKind           string         `json:"action_kind,omitempty"`
	EvidenceRefs         []EvidenceRef  `json:"evidence_refs,omitempty"`
}

type IdempotencyCheckResult struct {
	ContractVersion string               `json:"contract_version,omitempty"`
	Operation       IdempotencyOperation `json:"operation,omitempty"`
	Status          HostActionStatus     `json:"status,omitempty"`
	ReadyForCreate  bool                 `json:"ready_for_create"`
	Replayed        bool                 `json:"replayed"`
	Conflict        bool                 `json:"conflict"`
	RequestedKey    DisplaySafeRef       `json:"requested_key,omitempty"`
	ExistingKey     DisplaySafeRef       `json:"existing_key,omitempty"`
	ExistingRef     DisplaySafeRef       `json:"existing_ref,omitempty"`
	ActionKind      string               `json:"action_kind,omitempty"`
	FailureClass    FailureClass         `json:"failure_class,omitempty"`
	EvidenceRefs    []EvidenceRef        `json:"evidence_refs,omitempty"`
	Boundaries      []Boundary           `json:"boundaries,omitempty"`
	MissingInputs   []MissingInput       `json:"missing_inputs,omitempty"`
	NextHostAction  NextHostAction       `json:"next_host_action,omitempty"`
}

func CheckIdempotency(input IdempotencyCheckInput) IdempotencyCheckResult {
	requestedKey, _ := NormalizeDisplaySafeRef(string(input.RequestedKey))
	existingKey, _ := NormalizeDisplaySafeRef(string(input.ExistingKey))
	existingRef, _ := NormalizeDisplaySafeRef(string(input.ExistingRef))
	actionKind := normalizeControlToken(input.ActionKind)
	if actionKind == "" {
		actionKind = defaultHostActionKind
	}
	result := IdempotencyCheckResult{
		ContractVersion: ContractVersion,
		RequestedKey:    requestedKey,
		ExistingKey:     existingKey,
		ExistingRef:     existingRef,
		ActionKind:      actionKind,
		FailureClass:    FailureNone,
		EvidenceRefs:    normalizeEvidenceRefs(input.EvidenceRefs),
	}
	if requestedKey == "" {
		result.Operation = IdempotencyKeyMissing
		result.Status = HostActionBlocked
		result.FailureClass = FailureInvalidInput
		result.MissingInputs = []MissingInput{"host:idempotency_key"}
		result.Boundaries = []Boundary{"idempotency_key_missing"}
		result.NextHostAction = "provide_idempotency_key"
		return result.Normalize()
	}
	if existingKey == "" || existingKey != requestedKey {
		result.Operation = IdempotencyCreate
		result.Status = HostActionReady
		result.ReadyForCreate = true
		result.Boundaries = []Boundary{"idempotency_create_ready"}
		result.NextHostAction = "host_may_apply_action"
		return result.Normalize()
	}
	requestedFingerprint := normalizeFingerprint(input.RequestedFingerprint)
	existingFingerprint := normalizeFingerprint(input.ExistingFingerprint)
	if requestedFingerprint != "" && existingFingerprint != "" && requestedFingerprint != existingFingerprint {
		result.Operation = IdempotencyConflict
		result.Status = HostActionReviewRequired
		result.Conflict = true
		result.FailureClass = FailureVerificationFailed
		result.MissingInputs = []MissingInput{"host:idempotency_conflict_review"}
		result.Boundaries = []Boundary{"idempotency_conflict"}
		result.NextHostAction = "review_idempotency_conflict"
		return result.Normalize()
	}
	result.Operation = IdempotencyReplay
	result.Status = HostActionRecorded
	result.Replayed = true
	result.Boundaries = []Boundary{"idempotent_replay"}
	result.NextHostAction = "host_may_report_existing_action"
	return result.Normalize()
}

func CloneIdempotencyCheckResult(in IdempotencyCheckResult) IdempotencyCheckResult {
	out := in
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	return out
}

func (r IdempotencyCheckResult) Normalize() IdempotencyCheckResult {
	out := CloneIdempotencyCheckResult(r)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.RequestedKey, _ = NormalizeDisplaySafeRef(string(out.RequestedKey))
	out.ExistingKey, _ = NormalizeDisplaySafeRef(string(out.ExistingKey))
	out.ExistingRef, _ = NormalizeDisplaySafeRef(string(out.ExistingRef))
	out.ActionKind = normalizeControlToken(out.ActionKind)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	out.ReadyForCreate = out.Operation == IdempotencyCreate && out.Status == HostActionReady
	out.Replayed = out.Operation == IdempotencyReplay && out.Status == HostActionRecorded
	out.Conflict = out.Operation == IdempotencyConflict
	return out
}

type EventRef struct {
	Ref         DisplaySafeRef `json:"ref,omitempty"`
	Kind        string         `json:"kind,omitempty"`
	SourceRef   DisplaySafeRef `json:"source_ref,omitempty"`
	CreatedAt   string         `json:"created_at,omitempty"`
	Sequence    int64          `json:"sequence,omitempty"`
	Priority    int            `json:"priority,omitempty"`
	Fingerprint string         `json:"fingerprint,omitempty"`
}

func (e EventRef) Normalize() EventRef {
	out := e
	out.Ref, _ = NormalizeDisplaySafeRef(string(out.Ref))
	out.Kind = normalizeControlToken(out.Kind)
	out.SourceRef, _ = NormalizeDisplaySafeRef(string(out.SourceRef))
	out.CreatedAt = strings.TrimSpace(out.CreatedAt)
	if ContainsUnsafeRawOutput(out.CreatedAt) {
		out.CreatedAt = ""
	}
	out.Fingerprint = normalizeFingerprint(out.Fingerprint)
	return out
}

func SelectLatestEvent(events []EventRef) (EventRef, bool) {
	normalized := make([]EventRef, 0, len(events))
	for _, event := range events {
		out := event.Normalize()
		if out.Ref == "" {
			continue
		}
		normalized = append(normalized, out)
	}
	if len(normalized) == 0 {
		return EventRef{}, false
	}
	sort.SliceStable(normalized, func(i, j int) bool {
		return EventIsNewer(normalized[i], normalized[j])
	})
	return normalized[0], true
}

func EventIsNewer(candidate EventRef, base EventRef) bool {
	left := candidate.Normalize()
	right := base.Normalize()
	leftTime, leftHasTime := parseEventTime(left.CreatedAt)
	rightTime, rightHasTime := parseEventTime(right.CreatedAt)
	if leftHasTime && rightHasTime && !leftTime.Equal(rightTime) {
		return leftTime.After(rightTime)
	}
	if leftHasTime != rightHasTime {
		return leftHasTime
	}
	if left.Sequence != right.Sequence {
		return left.Sequence > right.Sequence
	}
	if left.Priority != right.Priority {
		return left.Priority > right.Priority
	}
	if left.Kind != right.Kind {
		return left.Kind > right.Kind
	}
	return string(left.Ref) > string(right.Ref)
}

type LatestSourceCheckResult struct {
	ContractVersion string             `json:"contract_version,omitempty"`
	Status          VerificationStatus `json:"status,omitempty"`
	SourceRef       DisplaySafeRef     `json:"source_ref,omitempty"`
	LatestRef       DisplaySafeRef     `json:"latest_ref,omitempty"`
	Stale           bool               `json:"stale"`
	FailureClass    FailureClass       `json:"failure_class,omitempty"`
	Boundaries      []Boundary         `json:"boundaries,omitempty"`
	MissingInputs   []MissingInput     `json:"missing_inputs,omitempty"`
	NextHostAction  NextHostAction     `json:"next_host_action,omitempty"`
}

func CheckLatestSourceEvent(source EventRef, latest EventRef) LatestSourceCheckResult {
	source = source.Normalize()
	latest = latest.Normalize()
	result := LatestSourceCheckResult{
		ContractVersion: ContractVersion,
		Status:          VerificationBlocked,
		SourceRef:       source.Ref,
		LatestRef:       latest.Ref,
		FailureClass:    FailureNone,
	}
	if source.Ref == "" {
		result.FailureClass = FailureInvalidInput
		result.MissingInputs = []MissingInput{"host:source_event_ref"}
		result.Boundaries = []Boundary{"source_event_ref_missing"}
		result.NextHostAction = "provide_source_event_ref"
		return result.Normalize()
	}
	if latest.Ref == "" {
		result.FailureClass = FailureInsufficientInformation
		result.MissingInputs = []MissingInput{"host:latest_event_ref"}
		result.Boundaries = []Boundary{"latest_event_ref_missing"}
		result.NextHostAction = "provide_latest_event_ref"
		return result.Normalize()
	}
	if latest.Ref != source.Ref && EventIsNewer(latest, source) {
		result.Status = VerificationReviewRequired
		result.Stale = true
		result.FailureClass = FailureVerificationFailed
		result.MissingInputs = []MissingInput{"host:stale_source_review"}
		result.Boundaries = []Boundary{"source_event_stale"}
		result.NextHostAction = "review_stale_source_event"
		return result.Normalize()
	}
	result.Status = VerificationSatisfied
	result.Boundaries = []Boundary{"source_event_current"}
	result.NextHostAction = "host_may_continue"
	return result.Normalize()
}

func CloneLatestSourceCheckResult(in LatestSourceCheckResult) LatestSourceCheckResult {
	out := in
	out.Boundaries = cloneBoundaries(in.Boundaries)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	return out
}

func (r LatestSourceCheckResult) Normalize() LatestSourceCheckResult {
	out := CloneLatestSourceCheckResult(r)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Status = NormalizeVerificationStatus(string(out.Status))
	out.SourceRef, _ = NormalizeDisplaySafeRef(string(out.SourceRef))
	out.LatestRef, _ = NormalizeDisplaySafeRef(string(out.LatestRef))
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	out.Stale = out.Status == VerificationReviewRequired && containsBoundary(out.Boundaries, "source_event_stale")
	return out
}

type ReviewRequiredInput struct {
	FailureClass   FailureClass   `json:"failure_class,omitempty"`
	Boundary       Boundary       `json:"boundary,omitempty"`
	MissingInput   MissingInput   `json:"missing_input,omitempty"`
	Finding        string         `json:"finding,omitempty"`
	NextHostAction NextHostAction `json:"next_host_action,omitempty"`
}

func RequireVerificationReview(result VerificationResult, review ReviewRequiredInput) VerificationResult {
	out := result.Normalize()
	out.Status = VerificationReviewRequired
	out.Satisfied = false
	if failure := NormalizeFailureClass(string(review.FailureClass)); failure != FailureNone {
		out.FailureClass = failure
	} else if out.FailureClass == FailureNone {
		out.FailureClass = FailureVerificationFailed
	}
	out.Boundaries = appendUniqueBoundary(out.Boundaries, string(firstBoundary(review.Boundary, "review_required")))
	if missing := firstMissingInput(review.MissingInput, "host:review"); missing != "" {
		out.MissingInputs = appendUniqueMissingInput(out.MissingInputs, string(missing))
	}
	if finding := strings.TrimSpace(review.Finding); finding != "" && !ContainsUnsafeRawOutput(finding) {
		out.Findings = normalizeStringList(append(out.Findings, finding))
	}
	out.NextHostAction = firstNextHostAction(review.NextHostAction, "review_required")
	return out.Normalize()
}

func RequireHostActionReview(proposal HostActionProposal, review ReviewRequiredInput) HostActionProposal {
	out := proposal.Normalize()
	out.Status = HostActionReviewRequired
	if failure := NormalizeFailureClass(string(review.FailureClass)); failure != FailureNone {
		out.FailureClass = failure
	} else if out.FailureClass == FailureNone {
		out.FailureClass = FailureVerificationFailed
	}
	out.Boundaries = appendUniqueBoundary(out.Boundaries, string(firstBoundary(review.Boundary, "review_required")))
	if missing := firstMissingInput(review.MissingInput, "host:review"); missing != "" {
		out.MissingInputs = appendUniqueMissingInput(out.MissingInputs, string(missing))
	}
	out.NextHostAction = firstNextHostAction(review.NextHostAction, "review_required")
	return out.Normalize()
}

type LifecycleStage string

const (
	LifecycleStageNotReady LifecycleStage = "not_ready"
	LifecycleStageReady    LifecycleStage = "ready"
	LifecycleStageApplied  LifecycleStage = "applied"
	LifecycleStageReadback LifecycleStage = "readback"
	LifecycleStageHandoff  LifecycleStage = "handoff"
	LifecycleStageAudit    LifecycleStage = "audit"
	LifecycleStageClosed   LifecycleStage = "lifecycle_closed"
)

func NormalizeLifecycleStage(raw string) LifecycleStage {
	switch normalizeEnumToken(raw) {
	case "", "not_ready", "blocked", "pending":
		return LifecycleStageNotReady
	case "ready":
		return LifecycleStageReady
	case "applied", "apply":
		return LifecycleStageApplied
	case "readback", "read_back":
		return LifecycleStageReadback
	case "handoff", "hand_off":
		return LifecycleStageHandoff
	case "audit", "audited":
		return LifecycleStageAudit
	case "lifecycle_closed", "closed", "complete", "completed":
		return LifecycleStageClosed
	default:
		return ""
	}
}

type LifecycleTransitionResult struct {
	ContractVersion  string           `json:"contract_version,omitempty"`
	From             LifecycleStage   `json:"from,omitempty"`
	To               LifecycleStage   `json:"to,omitempty"`
	Allowed          bool             `json:"allowed"`
	IdempotentReplay bool             `json:"idempotent_replay"`
	Status           HostActionStatus `json:"status,omitempty"`
	FailureClass     FailureClass     `json:"failure_class,omitempty"`
	Boundaries       []Boundary       `json:"boundaries,omitempty"`
	MissingInputs    []MissingInput   `json:"missing_inputs,omitempty"`
	NextHostAction   NextHostAction   `json:"next_host_action,omitempty"`
}

// CheckLifecycleTransition reports whether a lifecycle transition is allowed.
//
// Deprecated: lifecycle transition reduction is owned by controlplane closeout
// builders. External hosts should consume their projected transition result
// instead of invoking this reducer directly.
func CheckLifecycleTransition(from LifecycleStage, to LifecycleStage) LifecycleTransitionResult {
	return checkLifecycleTransition(from, to)
}

func checkLifecycleTransition(from LifecycleStage, to LifecycleStage) LifecycleTransitionResult {
	from = NormalizeLifecycleStage(string(from))
	to = NormalizeLifecycleStage(string(to))
	result := LifecycleTransitionResult{
		ContractVersion: ContractVersion,
		From:            from,
		To:              to,
		Status:          HostActionBlocked,
		FailureClass:    FailureNone,
	}
	fromOrdinal, fromOK := lifecycleStageOrdinal(from)
	toOrdinal, toOK := lifecycleStageOrdinal(to)
	if !fromOK || !toOK {
		result.FailureClass = FailureInvalidInput
		result.MissingInputs = []MissingInput{"controlplane:lifecycle_stage"}
		result.Boundaries = []Boundary{"lifecycle_stage_invalid"}
		result.NextHostAction = "provide_lifecycle_stage"
		return result.Normalize()
	}
	switch {
	case toOrdinal == fromOrdinal:
		result.Allowed = true
		result.IdempotentReplay = true
		result.Status = HostActionRecorded
		result.Boundaries = []Boundary{"lifecycle_idempotent_replay"}
		result.NextHostAction = "host_may_report_existing_state"
	case toOrdinal == fromOrdinal+1:
		result.Allowed = true
		result.Status = HostActionReady
		result.Boundaries = []Boundary{"lifecycle_transition_ready"}
		result.NextHostAction = "host_may_apply_action"
	case toOrdinal < fromOrdinal:
		result.Status = HostActionReviewRequired
		result.FailureClass = FailureVerificationFailed
		result.MissingInputs = []MissingInput{"host:lifecycle_regression_review"}
		result.Boundaries = []Boundary{"lifecycle_regression_review_required"}
		result.NextHostAction = "review_lifecycle_regression"
	default:
		result.Status = HostActionReviewRequired
		result.FailureClass = FailureInsufficientInformation
		result.MissingInputs = []MissingInput{"host:lifecycle_intermediate_state"}
		result.Boundaries = []Boundary{"lifecycle_transition_gap"}
		result.NextHostAction = "provide_lifecycle_intermediate_state"
	}
	return result.Normalize()
}

func CloneLifecycleTransitionResult(in LifecycleTransitionResult) LifecycleTransitionResult {
	out := in
	out.Boundaries = cloneBoundaries(in.Boundaries)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	return out
}

func (r LifecycleTransitionResult) Normalize() LifecycleTransitionResult {
	out := CloneLifecycleTransitionResult(r)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.From = NormalizeLifecycleStage(string(out.From))
	out.To = NormalizeLifecycleStage(string(out.To))
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	out.Allowed = out.Status == HostActionReady || out.Status == HostActionRecorded
	out.IdempotentReplay = out.Status == HostActionRecorded && containsBoundary(out.Boundaries, "lifecycle_idempotent_replay")
	return out
}

func MergeBoundaries(groups ...[]Boundary) []Boundary {
	merged := []Boundary{}
	for _, group := range groups {
		merged = append(merged, group...)
	}
	return normalizeBoundaries(merged)
}

func AppendBoundaries(base []Boundary, values ...Boundary) []Boundary {
	return MergeBoundaries(base, values)
}

func MergeMissingInputs(groups ...[]MissingInput) []MissingInput {
	merged := []MissingInput{}
	for _, group := range groups {
		merged = append(merged, group...)
	}
	return normalizeMissingInputs(merged)
}

func AppendMissingInputs(base []MissingInput, values ...MissingInput) []MissingInput {
	return MergeMissingInputs(base, values)
}

func MergeEvidenceRefs(groups ...[]EvidenceRef) []EvidenceRef {
	merged := []EvidenceRef{}
	for _, group := range groups {
		merged = append(merged, group...)
	}
	return normalizeEvidenceRefs(merged)
}

func firstBoundary(value Boundary, fallback Boundary) Boundary {
	normalized := normalizeBoundaries([]Boundary{value})
	if len(normalized) > 0 {
		return normalized[0]
	}
	normalized = normalizeBoundaries([]Boundary{fallback})
	if len(normalized) > 0 {
		return normalized[0]
	}
	return ""
}

func firstMissingInput(value MissingInput, fallback MissingInput) MissingInput {
	normalized := normalizeMissingInputs([]MissingInput{value})
	if len(normalized) > 0 {
		return normalized[0]
	}
	normalized = normalizeMissingInputs([]MissingInput{fallback})
	if len(normalized) > 0 {
		return normalized[0]
	}
	return ""
}

func firstNextHostAction(value NextHostAction, fallback NextHostAction) NextHostAction {
	normalized := NormalizeNextHostAction(string(value))
	if normalized != "" {
		return normalized
	}
	return NormalizeNextHostAction(string(fallback))
}

func normalizeFingerprint(raw string) string {
	fingerprint := strings.TrimSpace(raw)
	if fingerprint == "" || ContainsUnsafeRawOutput(fingerprint) {
		return ""
	}
	return fingerprint
}

func parseEventTime(raw string) (time.Time, bool) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return time.Time{}, false
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339} {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed, true
		}
	}
	return time.Time{}, false
}

func lifecycleStageOrdinal(stage LifecycleStage) (int, bool) {
	switch NormalizeLifecycleStage(string(stage)) {
	case LifecycleStageNotReady:
		return 0, true
	case LifecycleStageReady:
		return 1, true
	case LifecycleStageApplied:
		return 2, true
	case LifecycleStageReadback:
		return 3, true
	case LifecycleStageHandoff:
		return 4, true
	case LifecycleStageAudit:
		return 5, true
	case LifecycleStageClosed:
		return 6, true
	default:
		return 0, false
	}
}

func containsBoundary(boundaries []Boundary, value Boundary) bool {
	needle := firstBoundary(value, "")
	if needle == "" {
		return false
	}
	for _, boundary := range normalizeBoundaries(boundaries) {
		if boundary == needle {
			return true
		}
	}
	return false
}
