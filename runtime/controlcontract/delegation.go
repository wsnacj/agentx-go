package controlcontract

type DelegationRequestInput struct {
	Activation                        Activation              `json:"activation,omitempty"`
	Frame                             ObjectiveFrame          `json:"frame,omitempty"`
	RequestedIntensity                ExecutionIntensity      `json:"requested_intensity,omitempty"`
	SubgoalRef                        DisplaySafeRef          `json:"subgoal_ref,omitempty"`
	WorkerRef                         DisplaySafeRef          `json:"worker_ref,omitempty"`
	AllowedToolRefs                   []DisplaySafeRef        `json:"allowed_tool_refs,omitempty"`
	DeniedToolRefs                    []DisplaySafeRef        `json:"denied_tool_refs,omitempty"`
	Budget                            ObjectiveBudgetSnapshot `json:"budget,omitempty"`
	EvidenceRequirements              []EvidenceRef           `json:"evidence_requirements,omitempty"`
	StopConditionRefs                 []DisplaySafeRef        `json:"stop_condition_refs,omitempty"`
	RedactionPolicyRef                DisplaySafeRef          `json:"redaction_policy_ref,omitempty"`
	MergePolicyRef                    DisplaySafeRef          `json:"merge_policy_ref,omitempty"`
	ExecutionContractAllowsDelegation bool                    `json:"execution_contract_allows_delegation"`
	HostAllowsL4Delegation            bool                    `json:"host_allows_l4_delegation"`
	L5Enabled                         bool                    `json:"l5_enabled"`
	UserConfirmed                     bool                    `json:"user_confirmed"`
	HostApproved                      bool                    `json:"host_approved"`
	ApprovalRefs                      []DisplaySafeRef        `json:"approval_refs,omitempty"`
	PolicyRefs                        []DisplaySafeRef        `json:"policy_refs,omitempty"`
	DecisionBasis                     []DisplaySafeRef        `json:"decision_basis,omitempty"`
	Boundaries                        []Boundary              `json:"boundaries,omitempty"`
	RawOutputLoaded                   bool                    `json:"raw_output_loaded"`
}

type DelegationRequestProjection struct {
	ContractVersion                   string                  `json:"contract_version,omitempty"`
	Projected                         bool                    `json:"projected"`
	Status                            VerificationStatus      `json:"status,omitempty"`
	ReadyForWorkerDispatch            bool                    `json:"ready_for_worker_dispatch"`
	DelegationAllowed                 bool                    `json:"delegation_allowed"`
	Activation                        Activation              `json:"activation,omitempty"`
	Frame                             ObjectiveFrame          `json:"frame,omitempty"`
	RequestedIntensity                ExecutionIntensity      `json:"requested_intensity,omitempty"`
	SubgoalRef                        DisplaySafeRef          `json:"subgoal_ref,omitempty"`
	WorkerRef                         DisplaySafeRef          `json:"worker_ref,omitempty"`
	AllowedToolRefs                   []DisplaySafeRef        `json:"allowed_tool_refs,omitempty"`
	DeniedToolRefs                    []DisplaySafeRef        `json:"denied_tool_refs,omitempty"`
	Budget                            ObjectiveBudgetSnapshot `json:"budget,omitempty"`
	EvidenceRequirements              []EvidenceRef           `json:"evidence_requirements,omitempty"`
	StopConditionRefs                 []DisplaySafeRef        `json:"stop_condition_refs,omitempty"`
	RedactionPolicyRef                DisplaySafeRef          `json:"redaction_policy_ref,omitempty"`
	MergePolicyRef                    DisplaySafeRef          `json:"merge_policy_ref,omitempty"`
	ExecutionContractAllowsDelegation bool                    `json:"execution_contract_allows_delegation"`
	HostAllowsL4Delegation            bool                    `json:"host_allows_l4_delegation"`
	L5Enabled                         bool                    `json:"l5_enabled"`
	UserConfirmed                     bool                    `json:"user_confirmed"`
	HostApproved                      bool                    `json:"host_approved"`
	ApprovalRefs                      []DisplaySafeRef        `json:"approval_refs,omitempty"`
	PolicyRefs                        []DisplaySafeRef        `json:"policy_refs,omitempty"`
	MissingInputs                     []MissingInput          `json:"missing_inputs,omitempty"`
	BlockedReasons                    []string                `json:"blocked_reasons,omitempty"`
	FailureClass                      FailureClass            `json:"failure_class,omitempty"`
	DecisionBasis                     []DisplaySafeRef        `json:"decision_basis,omitempty"`
	Boundaries                        []Boundary              `json:"boundaries,omitempty"`
	NextHostAction                    NextHostAction          `json:"next_host_action,omitempty"`
	WorkerResultRequiresVerification  bool                    `json:"worker_result_requires_verification"`
	WorkerOutputAcceptedAsFact        bool                    `json:"worker_output_accepted_as_fact"`
	RunnerEffect                      string                  `json:"runner_effect,omitempty"`
	PromptEffect                      string                  `json:"prompt_effect,omitempty"`
	RawOutputLoaded                   bool                    `json:"raw_output_loaded"`
}

type DelegationWorkerResultReviewInput struct {
	Request         DelegationRequestProjection     `json:"request,omitempty"`
	WorkerRunRef    DisplaySafeRef                  `json:"worker_run_ref,omitempty"`
	WorkerResultRef DisplaySafeRef                  `json:"worker_result_ref,omitempty"`
	Verification    ObjectiveVerificationGateResult `json:"verification,omitempty"`
	EvidenceRefs    []EvidenceRef                   `json:"evidence_refs,omitempty"`
	Boundaries      []Boundary                      `json:"boundaries,omitempty"`
	RawOutputLoaded bool                            `json:"raw_output_loaded"`
}

type DelegationWorkerResultReview struct {
	ContractVersion                  string             `json:"contract_version,omitempty"`
	Projected                        bool               `json:"projected"`
	Status                           VerificationStatus `json:"status,omitempty"`
	ReadyForParentMerge              bool               `json:"ready_for_parent_merge"`
	WorkerResultRequiresVerification bool               `json:"worker_result_requires_verification"`
	WorkerOutputAcceptedAsFact       bool               `json:"worker_output_accepted_as_fact"`
	WorkerRunRef                     DisplaySafeRef     `json:"worker_run_ref,omitempty"`
	WorkerResultRef                  DisplaySafeRef     `json:"worker_result_ref,omitempty"`
	ParentVerificationStatus         VerificationStatus `json:"parent_verification_status,omitempty"`
	EvidenceRefs                     []EvidenceRef      `json:"evidence_refs,omitempty"`
	MissingInputs                    []MissingInput     `json:"missing_inputs,omitempty"`
	BlockedReasons                   []string           `json:"blocked_reasons,omitempty"`
	FailureClass                     FailureClass       `json:"failure_class,omitempty"`
	Boundaries                       []Boundary         `json:"boundaries,omitempty"`
	NextHostAction                   NextHostAction     `json:"next_host_action,omitempty"`
	RunnerEffect                     string             `json:"runner_effect,omitempty"`
	PromptEffect                     string             `json:"prompt_effect,omitempty"`
	RawOutputLoaded                  bool               `json:"raw_output_loaded"`
}

type DelegationWorkerParentMergeInput struct {
	Invocation               HostOwnedDelegationWorkerRuntimeInvocation `json:"invocation,omitempty"`
	ParentFrame              ObjectiveFrame                             `json:"parent_frame,omitempty"`
	ParentLedgerRef          DisplaySafeRef                             `json:"parent_ledger_ref,omitempty"`
	WorkerAttemptRef         AttemptRef                                 `json:"worker_attempt_ref,omitempty"`
	MergeRef                 DisplaySafeRef                             `json:"merge_ref,omitempty"`
	MergePolicyRef           DisplaySafeRef                             `json:"merge_policy_ref,omitempty"`
	WorkerObservations       []Observation                              `json:"worker_observations,omitempty"`
	EvidenceRefs             []EvidenceRef                              `json:"evidence_refs,omitempty"`
	RequiredEvidence         []EvidenceRef                              `json:"required_evidence,omitempty"`
	ExpectedObservationKinds []string                                   `json:"expected_observation_kinds,omitempty"`
	Boundaries               []Boundary                                 `json:"boundaries,omitempty"`
	RawOutputLoaded          bool                                       `json:"raw_output_loaded"`
}

type DelegationWorkerParentMerge struct {
	ContractVersion                  string                                     `json:"contract_version,omitempty"`
	Projected                        bool                                       `json:"projected"`
	Status                           VerificationStatus                         `json:"status,omitempty"`
	ReadyForParentMerge              bool                                       `json:"ready_for_parent_merge"`
	WorkerResultRequiresVerification bool                                       `json:"worker_result_requires_verification"`
	WorkerOutputAcceptedAsFact       bool                                       `json:"worker_output_accepted_as_fact"`
	Invocation                       HostOwnedDelegationWorkerRuntimeInvocation `json:"invocation,omitempty"`
	Request                          DelegationRequestProjection                `json:"request,omitempty"`
	ParentFrame                      ObjectiveFrame                             `json:"parent_frame,omitempty"`
	ParentLedgerRef                  DisplaySafeRef                             `json:"parent_ledger_ref,omitempty"`
	WorkerAttemptRef                 AttemptRef                                 `json:"worker_attempt_ref,omitempty"`
	MergeRef                         DisplaySafeRef                             `json:"merge_ref,omitempty"`
	MergePolicyRef                   DisplaySafeRef                             `json:"merge_policy_ref,omitempty"`
	WorkerRunRef                     DisplaySafeRef                             `json:"worker_run_ref,omitempty"`
	WorkerResultRef                  DisplaySafeRef                             `json:"worker_result_ref,omitempty"`
	Normalization                    ObservationNormalizationResult             `json:"normalization,omitempty"`
	ParentVerification               ObjectiveVerificationGateResult            `json:"parent_verification,omitempty"`
	ResultReview                     DelegationWorkerResultReview               `json:"result_review,omitempty"`
	ParentLedgerPatch                AttemptLedgerPatch                         `json:"parent_ledger_patch,omitempty"`
	WorkerAttempt                    AttemptSummary                             `json:"worker_attempt,omitempty"`
	Observations                     []Observation                              `json:"observations,omitempty"`
	EvidenceRefs                     []EvidenceRef                              `json:"evidence_refs,omitempty"`
	MissingInputs                    []MissingInput                             `json:"missing_inputs,omitempty"`
	BlockedReasons                   []string                                   `json:"blocked_reasons,omitempty"`
	FailureClass                     FailureClass                               `json:"failure_class,omitempty"`
	Boundaries                       []Boundary                                 `json:"boundaries,omitempty"`
	NextHostAction                   NextHostAction                             `json:"next_host_action,omitempty"`
	RunnerEffect                     string                                     `json:"runner_effect,omitempty"`
	PromptEffect                     string                                     `json:"prompt_effect,omitempty"`
	RawOutputLoaded                  bool                                       `json:"raw_output_loaded"`
}

type DelegationWorkerFailureReviewInput struct {
	Request             DelegationRequestProjection                  `json:"request,omitempty"`
	Invocations         []HostOwnedDelegationWorkerRuntimeInvocation `json:"invocations,omitempty"`
	ParentMerges        []DelegationWorkerParentMerge                `json:"parent_merges,omitempty"`
	WorkerAttempts      []AttemptSummary                             `json:"worker_attempts,omitempty"`
	FailureReviewRef    DisplaySafeRef                               `json:"failure_review_ref,omitempty"`
	FailureRef          DisplaySafeRef                               `json:"failure_ref,omitempty"`
	CompensationRef     DisplaySafeRef                               `json:"compensation_ref,omitempty"`
	ConflictRefs        []DisplaySafeRef                             `json:"conflict_refs,omitempty"`
	NoProgressThreshold int                                          `json:"no_progress_threshold,omitempty"`
	EvidenceRefs        []EvidenceRef                                `json:"evidence_refs,omitempty"`
	Boundaries          []Boundary                                   `json:"boundaries,omitempty"`
	RawOutputLoaded     bool                                         `json:"raw_output_loaded"`
}

type DelegationWorkerFailureReview struct {
	ContractVersion                  string                                       `json:"contract_version,omitempty"`
	Projected                        bool                                         `json:"projected"`
	Status                           string                                       `json:"status,omitempty"`
	ReadyForFailureReview            bool                                         `json:"ready_for_failure_review"`
	ReadyForCompensationReview       bool                                         `json:"ready_for_compensation_review"`
	WorkerResultRequiresVerification bool                                         `json:"worker_result_requires_verification"`
	WorkerOutputAcceptedAsFact       bool                                         `json:"worker_output_accepted_as_fact"`
	NoProgressDetected               bool                                         `json:"no_progress_detected"`
	ConflictingResultsDetected       bool                                         `json:"conflicting_results_detected"`
	WorkerFailureReported            bool                                         `json:"worker_failure_reported"`
	NoProgressAttemptCount           int                                          `json:"no_progress_attempt_count,omitempty"`
	NoProgressThreshold              int                                          `json:"no_progress_threshold,omitempty"`
	Request                          DelegationRequestProjection                  `json:"request,omitempty"`
	Invocations                      []HostOwnedDelegationWorkerRuntimeInvocation `json:"invocations,omitempty"`
	ParentMerges                     []DelegationWorkerParentMerge                `json:"parent_merges,omitempty"`
	WorkerAttempts                   []AttemptSummary                             `json:"worker_attempts,omitempty"`
	FailureReviewRef                 DisplaySafeRef                               `json:"failure_review_ref,omitempty"`
	FailureRef                       DisplaySafeRef                               `json:"failure_ref,omitempty"`
	CompensationRef                  DisplaySafeRef                               `json:"compensation_ref,omitempty"`
	ConflictRefs                     []DisplaySafeRef                             `json:"conflict_refs,omitempty"`
	NoProgressAttemptRefs            []AttemptRef                                 `json:"no_progress_attempt_refs,omitempty"`
	FailedWorkerRunRefs              []DisplaySafeRef                             `json:"failed_worker_run_refs,omitempty"`
	EvidenceRefs                     []EvidenceRef                                `json:"evidence_refs,omitempty"`
	MissingInputs                    []MissingInput                               `json:"missing_inputs,omitempty"`
	BlockedReasons                   []string                                     `json:"blocked_reasons,omitempty"`
	FailureClass                     FailureClass                                 `json:"failure_class,omitempty"`
	Boundaries                       []Boundary                                   `json:"boundaries,omitempty"`
	NextHostAction                   NextHostAction                               `json:"next_host_action,omitempty"`
	RunnerEffect                     string                                       `json:"runner_effect,omitempty"`
	PromptEffect                     string                                       `json:"prompt_effect,omitempty"`
	RawOutputLoaded                  bool                                         `json:"raw_output_loaded"`
}

func BuildDelegationRequestProjection(input DelegationRequestInput) DelegationRequestProjection {
	frame := input.Frame.Normalize()
	activation := NormalizeActivation(string(input.Activation))
	intensity := firstIntensity(input.RequestedIntensity, frame.Intensity)
	budget := input.Budget.Normalize()
	result := DelegationRequestProjection{
		ContractVersion:                   ContractVersion,
		Projected:                         true,
		Status:                            VerificationBlocked,
		Activation:                        activation,
		Frame:                             frame,
		RequestedIntensity:                intensity,
		SubgoalRef:                        normalizeOneDisplaySafeRef(input.SubgoalRef),
		WorkerRef:                         normalizeOneDisplaySafeRef(input.WorkerRef),
		AllowedToolRefs:                   normalizeDisplaySafeRefs(input.AllowedToolRefs),
		DeniedToolRefs:                    normalizeDisplaySafeRefs(input.DeniedToolRefs),
		Budget:                            budget,
		EvidenceRequirements:              normalizeEvidenceRefs(input.EvidenceRequirements),
		StopConditionRefs:                 normalizeDisplaySafeRefs(input.StopConditionRefs),
		RedactionPolicyRef:                normalizeOneDisplaySafeRef(input.RedactionPolicyRef),
		MergePolicyRef:                    normalizeOneDisplaySafeRef(input.MergePolicyRef),
		ExecutionContractAllowsDelegation: input.ExecutionContractAllowsDelegation,
		HostAllowsL4Delegation:            input.HostAllowsL4Delegation,
		L5Enabled:                         input.L5Enabled,
		UserConfirmed:                     input.UserConfirmed,
		HostApproved:                      input.HostApproved,
		ApprovalRefs:                      normalizeDisplaySafeRefs(input.ApprovalRefs),
		PolicyRefs:                        normalizeDisplaySafeRefs(input.PolicyRefs),
		FailureClass:                      FailureNone,
		DecisionBasis: normalizeDisplaySafeRefs(append(
			[]DisplaySafeRef{
				"delegation:request_projection",
				DisplaySafeRef("requested_intensity:" + string(intensity)),
			},
			input.DecisionBasis...,
		)),
		Boundaries: MergeBoundaries(
			[]Boundary{
				"delegation_request_projection",
				"delegation_worker_boundary",
				"projection_only",
				"display_safe_refs_only",
				"no_runner_dispatch",
				"no_worker_dispatch",
				"model_route_is_not_authorization",
				"worker_output_not_fact",
				"worker_result_requires_verification",
			},
			input.Boundaries,
		),
		NextHostAction:                   "review_delegation_request",
		WorkerResultRequiresVerification: true,
		WorkerOutputAcceptedAsFact:       false,
		RunnerEffect:                     "none",
		PromptEffect:                     "none",
		RawOutputLoaded:                  input.RawOutputLoaded,
	}
	if input.RawOutputLoaded || delegationRequestUnsafeInput(input) {
		return delegationRequestBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs", "raw_output_not_allowed")
	}
	if activation != ActivationManaged {
		return delegationRequestBlock(result, FailurePolicyBlocked, "managed_activation_required", "control_plane:managed_activation", "enable_managed_objective", "delegation_requires_managed_activation")
	}
	if frame.ID == "" {
		return delegationRequestBlock(result, FailureConfigMissing, "parent_objective_frame_missing", "host:objective_frame", "provide_objective_frame", "delegation_parent_frame_missing")
	}
	if executionIntensityRank(intensity) < executionIntensityRank(IntensityL4DurableLongRun) {
		return delegationRequestBlock(result, FailurePolicyBlocked, "delegation_requires_l4_or_l5", "contract:intensity_l4_or_l5", "request_intensity_upgrade_confirmation", "delegation_below_l4_blocked")
	}
	if !input.ExecutionContractAllowsDelegation {
		return delegationRequestBlock(result, FailurePolicyBlocked, "execution_contract_disallows_delegation", "contract:delegation", "provide_execution_contract", "delegation_contract_missing")
	}
	switch intensity {
	case IntensityL4DurableLongRun:
		if !input.HostAllowsL4Delegation {
			return delegationRequestBlock(result, FailureApprovalRequired, "l4_delegation_requires_explicit_host_policy", "host:l4_delegation_policy", "request_host_approval", "l4_delegation_not_explicitly_allowed")
		}
	case IntensityL5Autonomous:
		if !input.L5Enabled {
			return delegationRequestBlock(result, FailureApprovalRequired, "l5_delegation_disabled_by_default", "host:l5_delegation_policy", "request_host_approval", "l5_delegation_default_off")
		}
	default:
		return delegationRequestBlock(result, FailurePolicyBlocked, "delegation_intensity_not_supported", MissingInput("contract:intensity:"+string(intensity)), "request_intensity_upgrade_confirmation", "delegation_intensity_not_supported")
	}
	if !input.UserConfirmed {
		return delegationRequestBlock(result, FailureApprovalRequired, "user_confirmation_required", "user:delegation_confirmation", "request_user_confirmation", "delegation_user_confirmation_required")
	}
	if !input.HostApproved {
		return delegationRequestBlock(result, FailureApprovalRequired, "host_approval_required", "host:delegation_approval", "request_host_approval", "delegation_host_approval_required")
	}
	if len(result.ApprovalRefs) == 0 {
		return delegationRequestBlock(result, FailureEvidenceMissing, "delegation_approval_ref_missing", "host:approval_ref", "provide_host_approval_ref", "delegation_approval_ref_missing")
	}
	if result.SubgoalRef == "" {
		result = delegationRequestAccumulate(result, FailureEvidenceMissing, "delegation_subgoal_ref_missing", "host:delegation_subgoal_ref", "provide_delegation_subgoal_ref", "delegation_subgoal_ref_missing")
	}
	if result.WorkerRef == "" {
		result = delegationRequestAccumulate(result, FailureHostAdapterMissing, "delegation_worker_ref_missing", "host:delegation_worker_ref", "provide_delegation_worker_ref", "delegation_worker_ref_missing")
	}
	if len(result.AllowedToolRefs) == 0 {
		result = delegationRequestAccumulate(result, FailurePolicyBlocked, "worker_visible_tools_missing", "host:worker_allowed_tools", "provide_worker_tool_boundary", "worker_tool_boundary_missing")
	}
	if budget.BudgetRef == "" || budget.Limit <= 0 || budget.Exhausted {
		result = delegationRequestAccumulate(result, FailureBudgetExhausted, "delegation_budget_missing_or_exhausted", "host:delegation_budget", "provide_delegation_budget", "delegation_budget_missing_or_exhausted")
	}
	if len(result.EvidenceRequirements) == 0 {
		result = delegationRequestAccumulate(result, FailureEvidenceMissing, "delegation_evidence_requirements_missing", "host:delegation_evidence_requirements", "provide_delegation_evidence_requirements", "delegation_evidence_requirements_missing")
	}
	if len(result.StopConditionRefs) == 0 {
		result = delegationRequestAccumulate(result, FailureConfigMissing, "delegation_stop_condition_missing", "host:delegation_stop_condition", "provide_delegation_stop_condition", "delegation_stop_condition_missing")
	}
	if result.RedactionPolicyRef == "" {
		result = delegationRequestAccumulate(result, FailurePolicyBlocked, "delegation_redaction_policy_missing", "host:redaction_policy", "provide_redaction_policy", "delegation_redaction_policy_missing")
	}
	if result.MergePolicyRef == "" {
		result = delegationRequestAccumulate(result, FailurePolicyBlocked, "delegation_merge_policy_missing", "host:merge_policy", "provide_merge_policy", "delegation_merge_policy_missing")
	}
	if len(result.MissingInputs) > 0 || len(result.BlockedReasons) > 0 {
		return result.Normalize()
	}
	result.Status = VerificationSatisfied
	result.ReadyForWorkerDispatch = true
	result.DelegationAllowed = true
	result.NextHostAction = "host_may_dispatch_delegated_worker"
	result.Boundaries = AppendBoundaries(result.Boundaries, "delegation_request_ready", "host_may_dispatch_worker")
	return result.Normalize()
}

func CloneDelegationRequestProjection(in DelegationRequestProjection) DelegationRequestProjection {
	out := in
	out.Frame = in.Frame.Clone()
	out.AllowedToolRefs = cloneDisplaySafeRefs(in.AllowedToolRefs)
	out.DeniedToolRefs = cloneDisplaySafeRefs(in.DeniedToolRefs)
	out.Budget = in.Budget.Clone()
	out.EvidenceRequirements = cloneEvidenceRefs(in.EvidenceRequirements)
	out.StopConditionRefs = cloneDisplaySafeRefs(in.StopConditionRefs)
	out.ApprovalRefs = cloneDisplaySafeRefs(in.ApprovalRefs)
	out.PolicyRefs = cloneDisplaySafeRefs(in.PolicyRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.DecisionBasis = cloneDisplaySafeRefs(in.DecisionBasis)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (p DelegationRequestProjection) Clone() DelegationRequestProjection {
	return CloneDelegationRequestProjection(p)
}

func (p DelegationRequestProjection) Normalize() DelegationRequestProjection {
	out := CloneDelegationRequestProjection(p)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeVerificationStatus(string(out.Status))
	out.Activation = NormalizeActivation(string(out.Activation))
	out.Frame = out.Frame.Normalize()
	out.RequestedIntensity = NormalizeExecutionIntensity(string(out.RequestedIntensity))
	out.SubgoalRef = normalizeOneDisplaySafeRef(out.SubgoalRef)
	out.WorkerRef = normalizeOneDisplaySafeRef(out.WorkerRef)
	out.AllowedToolRefs = normalizeDisplaySafeRefs(out.AllowedToolRefs)
	out.DeniedToolRefs = normalizeDisplaySafeRefs(out.DeniedToolRefs)
	out.Budget = out.Budget.Normalize()
	out.EvidenceRequirements = normalizeEvidenceRefs(out.EvidenceRequirements)
	out.StopConditionRefs = normalizeDisplaySafeRefs(out.StopConditionRefs)
	out.RedactionPolicyRef = normalizeOneDisplaySafeRef(out.RedactionPolicyRef)
	out.MergePolicyRef = normalizeOneDisplaySafeRef(out.MergePolicyRef)
	out.ApprovalRefs = normalizeDisplaySafeRefs(out.ApprovalRefs)
	out.PolicyRefs = normalizeDisplaySafeRefs(out.PolicyRefs)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.BlockedReasons = normalizeControlTokenList(out.BlockedReasons)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.DecisionBasis = normalizeDisplaySafeRefs(out.DecisionBasis)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	if out.Status == VerificationNotEvaluated {
		out.Status = VerificationBlocked
	}
	if out.RunnerEffect == "" {
		out.RunnerEffect = "none"
	}
	if out.PromptEffect == "" {
		out.PromptEffect = "none"
	}
	if out.RawOutputLoaded {
		out.Status = VerificationReviewRequired
		out.ReadyForWorkerDispatch = false
		out.DelegationAllowed = false
		if out.FailureClass == FailureNone {
			out.FailureClass = FailureEvidenceWeak
		}
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		if out.NextHostAction == "" {
			out.NextHostAction = "provide_display_safe_refs"
		}
	}
	out.WorkerResultRequiresVerification = true
	out.WorkerOutputAcceptedAsFact = false
	out.ReadyForWorkerDispatch = out.Status == VerificationSatisfied &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.RawOutputLoaded &&
		out.DelegationAllowed
	return out
}

func BuildDelegationWorkerResultReview(input DelegationWorkerResultReviewInput) DelegationWorkerResultReview {
	request := input.Request.Normalize()
	verificationProvided := delegationParentVerificationProvided(input.Verification)
	verification := input.Verification.Normalize()
	result := DelegationWorkerResultReview{
		ContractVersion:                  ContractVersion,
		Projected:                        true,
		Status:                           VerificationBlocked,
		WorkerResultRequiresVerification: true,
		WorkerOutputAcceptedAsFact:       false,
		WorkerRunRef:                     normalizeOneDisplaySafeRef(input.WorkerRunRef),
		WorkerResultRef:                  normalizeOneDisplaySafeRef(input.WorkerResultRef),
		ParentVerificationStatus:         verification.Status,
		EvidenceRefs:                     MergeEvidenceRefs(input.EvidenceRefs, verification.EvidenceRefs),
		FailureClass:                     FailureNone,
		Boundaries: MergeBoundaries(
			[]Boundary{
				"delegation_worker_result_review",
				"worker_output_not_fact",
				"worker_result_requires_verification",
				"parent_verification_required",
				"projection_only",
				"display_safe_refs_only",
				"no_runner_dispatch",
				"no_worker_dispatch",
			},
			input.Boundaries,
			verification.Boundaries,
		),
		NextHostAction:  "run_parent_verification_gate",
		RunnerEffect:    "none",
		PromptEffect:    "none",
		RawOutputLoaded: input.RawOutputLoaded || request.RawOutputLoaded || verification.RawOutputLoaded,
	}
	if result.RawOutputLoaded || delegationWorkerResultUnsafeInput(input) {
		return delegationWorkerResultReviewBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs", "raw_output_not_allowed")
	}
	if !request.ReadyForWorkerDispatch {
		result.MissingInputs = AppendMissingInputs(result.MissingInputs, request.MissingInputs...)
		result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, "delegation_request_not_ready")
		return delegationWorkerResultReviewBlock(result, firstFailureClass(request.FailureClass, FailurePolicyBlocked), "delegation_request_not_ready", "host:delegation_request", "provide_delegation_request", "delegation_request_not_ready")
	}
	if result.WorkerRunRef == "" {
		result = delegationWorkerResultReviewAccumulate(result, FailureEvidenceMissing, "worker_run_ref_missing", "host:worker_run_ref", "provide_worker_result_refs", "worker_run_ref_missing")
	}
	if result.WorkerResultRef == "" {
		result = delegationWorkerResultReviewAccumulate(result, FailureEvidenceMissing, "worker_result_ref_missing", "host:worker_result_ref", "provide_worker_result_refs", "worker_result_ref_missing")
	}
	if !verificationProvided {
		result = delegationWorkerResultReviewAccumulate(result, FailureEvidenceMissing, "parent_verification_missing", "host:parent_verification", "run_parent_verification_gate", "parent_verification_missing")
	} else if !verification.Satisfied {
		result = delegationWorkerResultReviewAccumulate(result, firstFailureClass(verification.FailureClass, FailureVerificationFailed), "parent_verification_not_satisfied", "host:parent_verification", firstNextHostAction(verification.NextHostAction, "request_replan_or_return_partial"), "parent_verification_not_satisfied")
	}
	if len(result.MissingInputs) > 0 || len(result.BlockedReasons) > 0 {
		return result.Normalize()
	}
	result.Status = VerificationSatisfied
	result.ReadyForParentMerge = true
	result.NextHostAction = "merge_verified_worker_result"
	result.Boundaries = AppendBoundaries(result.Boundaries, "verified_worker_result_ready_for_parent_merge")
	return result.Normalize()
}

func BuildDelegationWorkerParentMerge(input DelegationWorkerParentMergeInput) DelegationWorkerParentMerge {
	invocation := input.Invocation.Normalize()
	request := invocation.Readiness.Request.Normalize()
	frame := input.ParentFrame.Normalize()
	if frame.ID == "" {
		frame = request.Frame.Normalize()
	}
	mergePolicyRef := normalizeOneDisplaySafeRef(input.MergePolicyRef)
	if mergePolicyRef == "" {
		mergePolicyRef = request.MergePolicyRef
	}
	sourceRef := firstDisplaySafeRef(input.MergeRef, invocation.WorkerResultRef, invocation.WorkerRunRef)
	normalization := BuildStructuredObservationNormalization(StructuredObservationNormalizationInput{
		Frame:                    frame,
		SourceKind:               "delegation_worker_result",
		SourceRef:                sourceRef,
		Observations:             input.WorkerObservations,
		EvidenceRefs:             input.EvidenceRefs,
		ExpectedObservationKinds: input.ExpectedObservationKinds,
		Boundaries: []Boundary{
			"delegation_worker_parent_merge_observation_normalization",
			"worker_observation_must_be_structured",
		},
		RawOutputLoaded: input.RawOutputLoaded || invocation.RawOutputLoaded,
	})
	verification := BuildObjectiveVerificationGate(ObjectiveVerificationGateInput{
		Frame:            frame,
		Normalization:    normalization,
		RequiredEvidence: input.RequiredEvidence,
		Boundaries: []Boundary{
			"delegation_worker_parent_merge_verification",
			"parent_objective_reverification_required",
		},
		RawOutputLoaded: input.RawOutputLoaded || normalization.RawOutputLoaded,
	})
	resultReview := BuildDelegationWorkerResultReview(DelegationWorkerResultReviewInput{
		Request:         request,
		WorkerRunRef:    invocation.WorkerRunRef,
		WorkerResultRef: invocation.WorkerResultRef,
		Verification:    verification,
		EvidenceRefs:    MergeEvidenceRefs(input.EvidenceRefs, verification.EvidenceRefs),
		Boundaries: []Boundary{
			"delegation_worker_parent_merge_result_review",
			"parent_verification_before_merge",
		},
		RawOutputLoaded: input.RawOutputLoaded || verification.RawOutputLoaded,
	})
	result := DelegationWorkerParentMerge{
		ContractVersion:                  ContractVersion,
		Projected:                        true,
		Status:                           VerificationBlocked,
		WorkerResultRequiresVerification: true,
		WorkerOutputAcceptedAsFact:       false,
		Invocation:                       invocation,
		Request:                          request,
		ParentFrame:                      frame,
		ParentLedgerRef:                  normalizeOneDisplaySafeRef(input.ParentLedgerRef),
		WorkerAttemptRef:                 normalizeOneAttemptRef(input.WorkerAttemptRef),
		MergeRef:                         normalizeOneDisplaySafeRef(input.MergeRef),
		MergePolicyRef:                   mergePolicyRef,
		WorkerRunRef:                     invocation.WorkerRunRef,
		WorkerResultRef:                  invocation.WorkerResultRef,
		Normalization:                    normalization,
		ParentVerification:               verification,
		ResultReview:                     resultReview,
		Observations:                     cloneObservations(normalization.Observations),
		EvidenceRefs:                     MergeEvidenceRefs(input.EvidenceRefs, verification.EvidenceRefs, resultReview.EvidenceRefs),
		FailureClass:                     FailureNone,
		Boundaries: MergeBoundaries(
			[]Boundary{
				"delegation_worker_parent_merge",
				"parent_objective_reverified_before_merge",
				"worker_output_not_fact",
				"worker_result_requires_verification",
				"projection_only",
				"display_safe_refs_only",
				"ledger_patch_projection_only",
				"no_worker_dispatch",
				"no_runner_dispatch",
				"no_store_mutation_by_core",
			},
			input.Boundaries,
			normalization.Boundaries,
			verification.Boundaries,
			resultReview.Boundaries,
		),
		NextHostAction:  "run_parent_verification_gate_for_delegation_worker_result",
		RunnerEffect:    "none",
		PromptEffect:    "none",
		RawOutputLoaded: input.RawOutputLoaded || invocation.RawOutputLoaded || normalization.RawOutputLoaded || verification.RawOutputLoaded || resultReview.RawOutputLoaded,
	}
	if delegationWorkerParentMergeUnsafe(input) {
		result.RawOutputLoaded = true
		result.FailureClass = FailureNone
		return delegationWorkerParentMergeBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs", "raw_output_not_allowed")
	}
	if !invocation.ReadyForWorkerResultReview || !invocation.HostInvocationCompleted || invocation.HostInvocationFailed {
		result.MissingInputs = AppendMissingInputs(result.MissingInputs, invocation.MissingInputs...)
		result.BlockedReasons = appendUniqueControlTokens(result.BlockedReasons, invocation.BlockedReasons)
		return delegationWorkerParentMergeBlock(result, firstFailureClass(invocation.FailureClass, FailureConfigMissing), "delegation_worker_runtime_invocation_not_ready_for_parent_merge", "host:delegation_worker_runtime_invocation", "review_delegation_worker_runtime_invocation", "delegation_worker_runtime_invocation_not_ready_for_parent_merge")
	}
	for _, check := range []struct {
		ok      bool
		reason  string
		missing MissingInput
		next    NextHostAction
	}{
		{frame.ID != "", "parent_objective_frame_missing", "host:parent_objective_frame", "provide_parent_objective_frame"},
		{result.ParentLedgerRef != "", "parent_ledger_ref_missing", "host:parent_attempt_ledger", "provide_parent_attempt_ledger"},
		{result.WorkerAttemptRef != "", "worker_attempt_ref_missing", "host:worker_attempt_ref", "provide_worker_attempt_ref"},
		{result.MergeRef != "", "delegation_worker_merge_ref_missing", "host:delegation_worker_merge_ref", "provide_delegation_worker_merge_ref"},
		{result.MergePolicyRef != "", "delegation_merge_policy_missing", "host:merge_policy", "provide_merge_policy"},
		{len(result.Observations) > 0, "delegation_worker_observations_missing", "host:delegation_worker_observations", "provide_delegation_worker_observations"},
	} {
		if !check.ok {
			result = delegationWorkerParentMergeAccumulate(result, FailureConfigMissing, check.reason, check.missing, check.next, Boundary(check.reason))
		}
	}
	if !normalization.ReadyForVerification {
		result.MissingInputs = AppendMissingInputs(result.MissingInputs, normalization.MissingInputs...)
		result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, "delegation_worker_observation_normalization_not_ready")
		result = delegationWorkerParentMergeAccumulate(result, firstFailureClass(normalization.FailureClass, FailureEvidenceMissing), "delegation_worker_observation_normalization_not_ready", "control_plane:normalized_observations", firstNextHostAction(normalization.NextHostAction, "normalize_observations"), "delegation_worker_observation_normalization_not_ready")
	}
	if !verification.Satisfied {
		result.MissingInputs = AppendMissingInputs(result.MissingInputs, verification.MissingInputs...)
		result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, "parent_verification_not_satisfied")
		result = delegationWorkerParentMergeAccumulate(result, firstFailureClass(verification.FailureClass, FailureVerificationFailed), "parent_verification_not_satisfied", "host:parent_verification", firstNextHostAction(verification.NextHostAction, "request_replan_or_return_partial"), "parent_verification_not_satisfied")
	}
	if !resultReview.ReadyForParentMerge {
		result.MissingInputs = AppendMissingInputs(result.MissingInputs, resultReview.MissingInputs...)
		result.BlockedReasons = appendUniqueControlTokens(result.BlockedReasons, resultReview.BlockedReasons)
		result = delegationWorkerParentMergeAccumulate(result, firstFailureClass(resultReview.FailureClass, FailureVerificationFailed), "delegation_worker_result_review_not_ready_for_parent_merge", "host:delegation_worker_result_review", firstNextHostAction(resultReview.NextHostAction, "run_parent_verification_gate"), "delegation_worker_result_review_not_ready_for_parent_merge")
	}
	if len(result.MissingInputs) > 0 || len(result.BlockedReasons) > 0 {
		return result.Normalize()
	}
	result.Status = VerificationSatisfied
	result.ReadyForParentMerge = true
	result.FailureClass = FailureNone
	result.NextHostAction = "update_objective_controller"
	result.WorkerAttempt = AttemptSummary{
		Ref:              result.WorkerAttemptRef,
		ObjectiveID:      frame.ID,
		StrategyID:       string(request.WorkerRef),
		Index:            1,
		ControlMode:      ControlModeDelegated,
		Intensity:        request.RequestedIntensity,
		Status:           VerificationSatisfied,
		EvidenceRefs:     cloneEvidenceRefs(result.EvidenceRefs),
		ObservationCount: len(result.Observations),
		FailureClass:     FailureNone,
		NextHostAction:   "update_objective_controller",
		Boundaries: []Boundary{
			"delegation_worker_attempt_verified",
			"worker_result_ready_for_parent_ledger_patch",
		},
	}
	result.ParentLedgerPatch = AttemptLedgerPatch{
		ObjectiveID:          frame.ID,
		LedgerRef:            result.ParentLedgerRef,
		Attempts:             []AttemptSummary{result.WorkerAttempt},
		RetryBudgetRemaining: resultRequestBudgetRemaining(request),
		EvidenceRefs:         cloneEvidenceRefs(result.EvidenceRefs),
		Boundaries: []Boundary{
			"delegation_worker_parent_ledger_patch",
			"verified_worker_result_merged_into_parent_ledger",
			"ledger_patch_projection_only",
		},
		NextHostAction: "update_objective_controller",
	}
	result.Boundaries = AppendBoundaries(result.Boundaries, "verified_worker_result_ready_for_parent_merge", "verified_worker_result_merged_into_parent_ledger")
	return result.Normalize()
}

func BuildDelegationWorkerFailureReview(input DelegationWorkerFailureReviewInput) DelegationWorkerFailureReview {
	request := input.Request.Normalize()
	invocations := normalizeDelegationWorkerRuntimeInvocations(input.Invocations)
	parentMerges := normalizeDelegationWorkerParentMerges(input.ParentMerges)
	workerAttempts := normalizeAttemptSummaries(input.WorkerAttempts)
	if !request.ReadyForWorkerDispatch {
		for _, invocation := range invocations {
			if invocation.Readiness.Request.ReadyForWorkerDispatch {
				request = invocation.Readiness.Request.Normalize()
				break
			}
		}
	}
	if !request.ReadyForWorkerDispatch {
		for _, merge := range parentMerges {
			if merge.Request.ReadyForWorkerDispatch {
				request = merge.Request.Normalize()
				break
			}
		}
	}
	noProgress, noProgressCount, noProgressRefs := delegationWorkerNoProgress(workerAttempts, parentMerges, input.NoProgressThreshold)
	conflict, conflictRefs := delegationWorkerConflictRefs(parentMerges, input.ConflictRefs)
	workerFailureReported, failedWorkerRunRefs, failureClass := delegationWorkerFailureReported(invocations)
	result := DelegationWorkerFailureReview{
		ContractVersion:                  ContractVersion,
		Projected:                        true,
		Status:                           "blocked",
		WorkerResultRequiresVerification: true,
		WorkerOutputAcceptedAsFact:       false,
		NoProgressDetected:               noProgress,
		ConflictingResultsDetected:       conflict,
		WorkerFailureReported:            workerFailureReported,
		NoProgressAttemptCount:           noProgressCount,
		NoProgressThreshold:              delegationWorkerNoProgressThreshold(input.NoProgressThreshold),
		Request:                          request,
		Invocations:                      invocations,
		ParentMerges:                     parentMerges,
		WorkerAttempts:                   workerAttempts,
		FailureReviewRef:                 normalizeOneDisplaySafeRef(input.FailureReviewRef),
		FailureRef:                       normalizeOneDisplaySafeRef(input.FailureRef),
		CompensationRef:                  normalizeOneDisplaySafeRef(input.CompensationRef),
		ConflictRefs:                     conflictRefs,
		NoProgressAttemptRefs:            noProgressRefs,
		FailedWorkerRunRefs:              failedWorkerRunRefs,
		EvidenceRefs:                     MergeEvidenceRefs(input.EvidenceRefs, delegationWorkerFailureReviewEvidenceRefs(invocations, parentMerges, workerAttempts)),
		FailureClass:                     firstFailureClass(failureClass, delegationWorkerFailureReviewClass(noProgress, conflict), FailureNone),
		Boundaries: MergeBoundaries(
			[]Boundary{
				"delegation_worker_failure_review",
				"worker_no_progress_and_conflict_review",
				"worker_output_not_fact",
				"worker_result_requires_verification",
				"projection_only",
				"display_safe_refs_only",
				"compensation_review_only",
				"no_worker_dispatch",
				"no_runner_dispatch",
				"no_store_mutation_by_core",
				"compensation_not_executed",
			},
			input.Boundaries,
		),
		NextHostAction:  "review_delegation_worker_failure",
		RunnerEffect:    "none",
		PromptEffect:    "none",
		RawOutputLoaded: input.RawOutputLoaded,
	}
	if delegationWorkerFailureReviewUnsafe(input, invocations, parentMerges, workerAttempts) {
		result.RawOutputLoaded = true
		result.FailureClass = FailureNone
		return delegationWorkerFailureReviewBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs", "raw_output_not_allowed")
	}
	if result.FailureReviewRef == "" {
		result = delegationWorkerFailureReviewAccumulate(result, FailureEvidenceMissing, "delegation_worker_failure_review_ref_missing", "host:delegation_worker_failure_review_ref", "provide_delegation_worker_failure_review_ref", "delegation_worker_failure_review_ref_missing")
	}
	if result.FailureRef == "" {
		result = delegationWorkerFailureReviewAccumulate(result, FailureEvidenceMissing, "delegation_worker_failure_ref_missing", "host:delegation_worker_failure_ref", "provide_delegation_worker_failure_ref", "delegation_worker_failure_ref_missing")
	}
	if !result.NoProgressDetected && !result.ConflictingResultsDetected && !result.WorkerFailureReported {
		result = delegationWorkerFailureReviewAccumulate(result, FailureEvidenceMissing, "delegation_worker_failure_signal_missing", "host:delegation_worker_failure_signal", "provide_delegation_worker_failure_signal", "delegation_worker_failure_signal_missing")
	}
	if len(result.MissingInputs) > 0 || len(result.BlockedReasons) > 0 {
		return result.Normalize()
	}
	result.Status = "ready_for_delegation_worker_failure_review"
	result.ReadyForFailureReview = true
	result.ReadyForCompensationReview = result.CompensationRef != ""
	result.NextHostAction = "review_delegation_worker_failure"
	result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_delegation_worker_failure_review")
	if result.NoProgressDetected {
		result.Boundaries = AppendBoundaries(result.Boundaries, "delegation_worker_no_progress_detected")
	}
	if result.ConflictingResultsDetected {
		result.Boundaries = AppendBoundaries(result.Boundaries, "delegation_worker_conflict_detected")
	}
	if result.WorkerFailureReported {
		result.Boundaries = AppendBoundaries(result.Boundaries, "delegation_worker_failure_reported")
	}
	if result.ReadyForCompensationReview {
		result.Boundaries = AppendBoundaries(result.Boundaries, "delegation_worker_compensation_review_ready")
	}
	return result.Normalize()
}

func CloneDelegationWorkerResultReview(in DelegationWorkerResultReview) DelegationWorkerResultReview {
	out := in
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (r DelegationWorkerResultReview) Clone() DelegationWorkerResultReview {
	return CloneDelegationWorkerResultReview(r)
}

func (r DelegationWorkerResultReview) Normalize() DelegationWorkerResultReview {
	out := CloneDelegationWorkerResultReview(r)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeVerificationStatus(string(out.Status))
	out.WorkerRunRef = normalizeOneDisplaySafeRef(out.WorkerRunRef)
	out.WorkerResultRef = normalizeOneDisplaySafeRef(out.WorkerResultRef)
	out.ParentVerificationStatus = NormalizeVerificationStatus(string(out.ParentVerificationStatus))
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.BlockedReasons = normalizeControlTokenList(out.BlockedReasons)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	if out.Status == VerificationNotEvaluated {
		out.Status = VerificationBlocked
	}
	if out.RunnerEffect == "" {
		out.RunnerEffect = "none"
	}
	if out.PromptEffect == "" {
		out.PromptEffect = "none"
	}
	if out.RawOutputLoaded {
		out.Status = VerificationReviewRequired
		out.ReadyForParentMerge = false
		if out.FailureClass == FailureNone {
			out.FailureClass = FailureEvidenceWeak
		}
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		if out.NextHostAction == "" {
			out.NextHostAction = "provide_display_safe_refs"
		}
	}
	out.WorkerResultRequiresVerification = true
	out.WorkerOutputAcceptedAsFact = false
	out.ReadyForParentMerge = out.Status == VerificationSatisfied &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.RawOutputLoaded
	return out
}

func CloneDelegationWorkerParentMerge(in DelegationWorkerParentMerge) DelegationWorkerParentMerge {
	out := in
	out.Invocation = in.Invocation.Clone()
	out.Request = in.Request.Clone()
	out.ParentFrame = in.ParentFrame.Clone()
	out.Normalization = in.Normalization.Clone()
	out.ParentVerification = in.ParentVerification.Clone()
	out.ResultReview = in.ResultReview.Clone()
	out.ParentLedgerPatch = in.ParentLedgerPatch.Clone()
	out.WorkerAttempt = in.WorkerAttempt.Clone()
	out.Observations = cloneObservations(in.Observations)
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (m DelegationWorkerParentMerge) Clone() DelegationWorkerParentMerge {
	return CloneDelegationWorkerParentMerge(m)
}

func (m DelegationWorkerParentMerge) Normalize() DelegationWorkerParentMerge {
	out := CloneDelegationWorkerParentMerge(m)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeVerificationStatus(string(out.Status))
	out.Invocation = out.Invocation.Normalize()
	out.Request = out.Request.Normalize()
	out.ParentFrame = out.ParentFrame.Normalize()
	out.ParentLedgerRef = normalizeOneDisplaySafeRef(out.ParentLedgerRef)
	out.WorkerAttemptRef = normalizeOneAttemptRef(out.WorkerAttemptRef)
	out.MergeRef = normalizeOneDisplaySafeRef(out.MergeRef)
	out.MergePolicyRef = normalizeOneDisplaySafeRef(out.MergePolicyRef)
	out.WorkerRunRef = normalizeOneDisplaySafeRef(out.WorkerRunRef)
	out.WorkerResultRef = normalizeOneDisplaySafeRef(out.WorkerResultRef)
	out.Normalization = out.Normalization.Normalize()
	out.ParentVerification = out.ParentVerification.Normalize()
	out.ResultReview = out.ResultReview.Normalize()
	out.ParentLedgerPatch = out.ParentLedgerPatch.Normalize()
	out.WorkerAttempt = out.WorkerAttempt.Normalize()
	out.Observations = normalizeObservations(out.Observations)
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.BlockedReasons = normalizeControlTokenList(out.BlockedReasons)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	if out.Status == VerificationNotEvaluated {
		out.Status = VerificationBlocked
	}
	if out.RunnerEffect == "" {
		out.RunnerEffect = "none"
	}
	if out.PromptEffect == "" {
		out.PromptEffect = "none"
	}
	out.WorkerResultRequiresVerification = true
	out.WorkerOutputAcceptedAsFact = false
	if out.RawOutputLoaded {
		out.Status = VerificationReviewRequired
		out.ReadyForParentMerge = false
		if out.FailureClass == FailureNone {
			out.FailureClass = FailureEvidenceWeak
		}
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		if out.NextHostAction == "" {
			out.NextHostAction = "provide_display_safe_refs"
		}
	}
	if len(out.MissingInputs) > 0 || len(out.BlockedReasons) > 0 || out.RawOutputLoaded {
		out.ReadyForParentMerge = false
		if out.Status == VerificationSatisfied {
			out.Status = VerificationBlocked
		}
	}
	out.ReadyForParentMerge = out.ReadyForParentMerge &&
		out.Status == VerificationSatisfied &&
		out.ResultReview.ReadyForParentMerge &&
		out.ParentVerification.Satisfied &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.RawOutputLoaded
	return out
}

func CloneDelegationWorkerFailureReview(in DelegationWorkerFailureReview) DelegationWorkerFailureReview {
	out := in
	out.Request = in.Request.Clone()
	out.Invocations = cloneDelegationWorkerRuntimeInvocations(in.Invocations)
	out.ParentMerges = cloneDelegationWorkerParentMerges(in.ParentMerges)
	out.WorkerAttempts = cloneAttemptSummaries(in.WorkerAttempts)
	out.ConflictRefs = cloneDisplaySafeRefs(in.ConflictRefs)
	out.NoProgressAttemptRefs = cloneAttemptRefs(in.NoProgressAttemptRefs)
	out.FailedWorkerRunRefs = cloneDisplaySafeRefs(in.FailedWorkerRunRefs)
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (r DelegationWorkerFailureReview) Clone() DelegationWorkerFailureReview {
	return CloneDelegationWorkerFailureReview(r)
}

func (r DelegationWorkerFailureReview) Normalize() DelegationWorkerFailureReview {
	out := CloneDelegationWorkerFailureReview(r)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = normalizeControlToken(out.Status)
	out.Request = out.Request.Normalize()
	out.Invocations = normalizeDelegationWorkerRuntimeInvocations(out.Invocations)
	out.ParentMerges = normalizeDelegationWorkerParentMerges(out.ParentMerges)
	out.WorkerAttempts = normalizeAttemptSummaries(out.WorkerAttempts)
	out.FailureReviewRef = normalizeOneDisplaySafeRef(out.FailureReviewRef)
	out.FailureRef = normalizeOneDisplaySafeRef(out.FailureRef)
	out.CompensationRef = normalizeOneDisplaySafeRef(out.CompensationRef)
	out.ConflictRefs = normalizeDisplaySafeRefs(out.ConflictRefs)
	out.NoProgressAttemptRefs = normalizeAttemptRefs(out.NoProgressAttemptRefs)
	out.FailedWorkerRunRefs = normalizeDisplaySafeRefs(out.FailedWorkerRunRefs)
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.BlockedReasons = normalizeControlTokenList(out.BlockedReasons)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	if out.Status == "" {
		out.Status = "blocked"
	}
	if out.RunnerEffect == "" {
		out.RunnerEffect = "none"
	}
	if out.PromptEffect == "" {
		out.PromptEffect = "none"
	}
	if out.NoProgressThreshold <= 0 {
		out.NoProgressThreshold = delegationWorkerNoProgressThreshold(out.NoProgressThreshold)
	}
	out.WorkerResultRequiresVerification = true
	out.WorkerOutputAcceptedAsFact = false
	if out.RawOutputLoaded {
		out.Status = "review_required"
		out.ReadyForFailureReview = false
		out.ReadyForCompensationReview = false
		if out.FailureClass == FailureNone {
			out.FailureClass = FailureEvidenceWeak
		}
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		if out.NextHostAction == "" {
			out.NextHostAction = "provide_display_safe_refs"
		}
	}
	if len(out.MissingInputs) > 0 || len(out.BlockedReasons) > 0 || out.RawOutputLoaded {
		out.ReadyForFailureReview = false
		out.ReadyForCompensationReview = false
	}
	out.ReadyForFailureReview = out.ReadyForFailureReview &&
		out.Status == "ready_for_delegation_worker_failure_review" &&
		out.FailureReviewRef != "" &&
		out.FailureRef != "" &&
		(out.NoProgressDetected || out.ConflictingResultsDetected || out.WorkerFailureReported) &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.RawOutputLoaded
	out.ReadyForCompensationReview = out.ReadyForCompensationReview &&
		out.ReadyForFailureReview &&
		out.CompensationRef != ""
	return out
}

func delegationRequestBlock(result DelegationRequestProjection, failure FailureClass, reason string, missing MissingInput, next NextHostAction, boundary Boundary) DelegationRequestProjection {
	result = delegationRequestAccumulate(result, failure, reason, missing, next, boundary)
	return result.Normalize()
}

func delegationRequestAccumulate(result DelegationRequestProjection, failure FailureClass, reason string, missing MissingInput, next NextHostAction, boundary Boundary) DelegationRequestProjection {
	result.Status = VerificationBlocked
	result.ReadyForWorkerDispatch = false
	result.DelegationAllowed = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.Boundaries = AppendBoundaries(result.Boundaries, boundary)
	result.NextHostAction = firstNextHostAction(next, result.NextHostAction)
	return result
}

func delegationWorkerResultReviewBlock(result DelegationWorkerResultReview, failure FailureClass, reason string, missing MissingInput, next NextHostAction, boundary Boundary) DelegationWorkerResultReview {
	result = delegationWorkerResultReviewAccumulate(result, failure, reason, missing, next, boundary)
	return result.Normalize()
}

func delegationWorkerResultReviewAccumulate(result DelegationWorkerResultReview, failure FailureClass, reason string, missing MissingInput, next NextHostAction, boundary Boundary) DelegationWorkerResultReview {
	result.Status = VerificationBlocked
	result.ReadyForParentMerge = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.Boundaries = AppendBoundaries(result.Boundaries, boundary)
	result.NextHostAction = firstNextHostAction(next, result.NextHostAction)
	return result
}

func delegationWorkerParentMergeBlock(result DelegationWorkerParentMerge, failure FailureClass, reason string, missing MissingInput, next NextHostAction, boundary Boundary) DelegationWorkerParentMerge {
	result = delegationWorkerParentMergeAccumulate(result, failure, reason, missing, next, boundary)
	return result.Normalize()
}

func delegationWorkerParentMergeAccumulate(result DelegationWorkerParentMerge, failure FailureClass, reason string, missing MissingInput, next NextHostAction, boundary Boundary) DelegationWorkerParentMerge {
	result.Status = VerificationBlocked
	result.ReadyForParentMerge = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.Boundaries = AppendBoundaries(result.Boundaries, boundary)
	result.NextHostAction = firstNextHostAction(next, result.NextHostAction)
	if result.NextHostAction == "" {
		result.NextHostAction = "review_delegation_worker_parent_merge"
	}
	return result
}

func delegationWorkerFailureReviewBlock(result DelegationWorkerFailureReview, failure FailureClass, reason string, missing MissingInput, next NextHostAction, boundary Boundary) DelegationWorkerFailureReview {
	result = delegationWorkerFailureReviewAccumulate(result, failure, reason, missing, next, boundary)
	return result.Normalize()
}

func delegationWorkerFailureReviewAccumulate(result DelegationWorkerFailureReview, failure FailureClass, reason string, missing MissingInput, next NextHostAction, boundary Boundary) DelegationWorkerFailureReview {
	result.Status = "blocked"
	result.ReadyForFailureReview = false
	result.ReadyForCompensationReview = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.Boundaries = AppendBoundaries(result.Boundaries, boundary)
	result.NextHostAction = firstNextHostAction(next, result.NextHostAction)
	if result.NextHostAction == "" {
		result.NextHostAction = "review_delegation_worker_failure"
	}
	return result
}

func delegationRequestUnsafeInput(input DelegationRequestInput) bool {
	return displaySafeRefRejected(input.SubgoalRef) ||
		displaySafeRefRejected(input.WorkerRef) ||
		displaySafeRefSliceRejected(input.AllowedToolRefs) ||
		displaySafeRefSliceRejected(input.DeniedToolRefs) ||
		evidenceRefRejected(input.EvidenceRequirements) ||
		displaySafeRefSliceRejected(input.StopConditionRefs) ||
		displaySafeRefRejected(input.RedactionPolicyRef) ||
		displaySafeRefRejected(input.MergePolicyRef) ||
		displaySafeRefSliceRejected(input.ApprovalRefs) ||
		displaySafeRefSliceRejected(input.PolicyRefs) ||
		displaySafeRefSliceRejected(input.DecisionBasis) ||
		input.RawOutputLoaded
}

func delegationWorkerResultUnsafeInput(input DelegationWorkerResultReviewInput) bool {
	return displaySafeRefRejected(input.WorkerRunRef) ||
		displaySafeRefRejected(input.WorkerResultRef) ||
		evidenceRefRejected(input.EvidenceRefs) ||
		input.RawOutputLoaded
}

func delegationWorkerParentMergeUnsafe(input DelegationWorkerParentMergeInput) bool {
	return input.RawOutputLoaded ||
		input.Invocation.RawOutputLoaded ||
		displaySafeRefRejected(input.ParentLedgerRef) ||
		attemptRefRejected(input.WorkerAttemptRef) ||
		displaySafeRefRejected(input.MergeRef) ||
		displaySafeRefRejected(input.MergePolicyRef) ||
		observationSliceUnsafePayload(input.WorkerObservations) ||
		evidenceRefRejected(input.EvidenceRefs) ||
		evidenceRefRejected(input.RequiredEvidence)
}

func delegationWorkerFailureReviewUnsafe(input DelegationWorkerFailureReviewInput, invocations []HostOwnedDelegationWorkerRuntimeInvocation, parentMerges []DelegationWorkerParentMerge, attempts []AttemptSummary) bool {
	if input.RawOutputLoaded ||
		input.Request.RawOutputLoaded ||
		displaySafeRefRejected(input.FailureReviewRef) ||
		displaySafeRefRejected(input.FailureRef) ||
		displaySafeRefRejected(input.CompensationRef) ||
		displaySafeRefSliceRejected(input.ConflictRefs) ||
		evidenceRefRejected(input.EvidenceRefs) {
		return true
	}
	for _, invocation := range invocations {
		if invocation.RawOutputLoaded {
			return true
		}
	}
	for _, merge := range parentMerges {
		if merge.RawOutputLoaded || delegationWorkerParentMergeOutputUnsafe(merge) {
			return true
		}
	}
	for _, attempt := range attempts {
		if attempt.RawOutputLoaded ||
			evidenceRefRejected(attempt.EvidenceRefs) ||
			attemptRefRejected(attempt.Ref) ||
			ContainsUnsafeRawOutput(attempt.ObjectiveID, attempt.StrategyID, attempt.FailureReason) {
			return true
		}
	}
	return false
}

func delegationParentVerificationProvided(verification ObjectiveVerificationGateResult) bool {
	return verification.Status != "" ||
		verification.Satisfied ||
		verification.Frame.ID != "" ||
		len(verification.Requirements) > 0 ||
		len(verification.Observations) > 0 ||
		len(verification.EvidenceRefs) > 0 ||
		verification.FailureClass != "" ||
		verification.FailureReason != "" ||
		len(verification.MissingInputs) > 0 ||
		len(verification.Boundaries) > 0 ||
		verification.NextHostAction != "" ||
		verification.RawOutputLoaded
}

func attemptRefRejected(value AttemptRef) bool {
	raw := string(value)
	if raw == "" {
		return false
	}
	_, ok := NormalizeAttemptRef(raw)
	return !ok
}

func resultRequestBudgetRemaining(request DelegationRequestProjection) int {
	budget := request.Budget.Normalize()
	if budget.Remaining > 0 {
		return budget.Remaining
	}
	return 0
}

func normalizeDelegationWorkerRuntimeInvocations(in []HostOwnedDelegationWorkerRuntimeInvocation) []HostOwnedDelegationWorkerRuntimeInvocation {
	out := make([]HostOwnedDelegationWorkerRuntimeInvocation, 0, len(in))
	for _, value := range in {
		normalized := value.Normalize()
		if !delegationWorkerRuntimeInvocationEmpty(normalized) {
			out = append(out, normalized)
		}
	}
	return out
}

func cloneDelegationWorkerRuntimeInvocations(in []HostOwnedDelegationWorkerRuntimeInvocation) []HostOwnedDelegationWorkerRuntimeInvocation {
	out := make([]HostOwnedDelegationWorkerRuntimeInvocation, len(in))
	for i := range in {
		out[i] = in[i].Clone()
	}
	return out
}

func delegationWorkerRuntimeInvocationEmpty(in HostOwnedDelegationWorkerRuntimeInvocation) bool {
	return !in.Projected &&
		in.Status == "" &&
		in.WorkerRunRef == "" &&
		in.WorkerResultRef == "" &&
		!in.HostInvocationReported &&
		!in.HostInvocationCompleted &&
		!in.HostInvocationFailed &&
		len(in.EvidenceRefs) == 0 &&
		len(in.MissingInputs) == 0 &&
		len(in.BlockedReasons) == 0 &&
		!in.RawOutputLoaded
}

func normalizeDelegationWorkerParentMerges(in []DelegationWorkerParentMerge) []DelegationWorkerParentMerge {
	out := make([]DelegationWorkerParentMerge, 0, len(in))
	for _, value := range in {
		normalized := value.Normalize()
		if !delegationWorkerParentMergeEmpty(normalized) {
			out = append(out, normalized)
		}
	}
	return out
}

func cloneDelegationWorkerParentMerges(in []DelegationWorkerParentMerge) []DelegationWorkerParentMerge {
	out := make([]DelegationWorkerParentMerge, len(in))
	for i := range in {
		out[i] = in[i].Clone()
	}
	return out
}

func delegationWorkerParentMergeEmpty(in DelegationWorkerParentMerge) bool {
	return !in.Projected &&
		in.Status == "" &&
		in.WorkerRunRef == "" &&
		in.WorkerResultRef == "" &&
		len(in.Observations) == 0 &&
		len(in.EvidenceRefs) == 0 &&
		len(in.MissingInputs) == 0 &&
		len(in.BlockedReasons) == 0 &&
		!in.RawOutputLoaded
}

func delegationWorkerParentMergeOutputUnsafe(merge DelegationWorkerParentMerge) bool {
	return observationSliceUnsafePayload(merge.Observations) ||
		evidenceRefRejected(merge.EvidenceRefs)
}

func delegationWorkerNoProgress(attempts []AttemptSummary, parentMerges []DelegationWorkerParentMerge, threshold int) (bool, int, []AttemptRef) {
	threshold = delegationWorkerNoProgressThreshold(threshold)
	count := 0
	refs := []AttemptRef{}
	for _, attempt := range normalizeAttemptSummaries(attempts) {
		if attempt.Status == VerificationSatisfied {
			continue
		}
		if len(attempt.EvidenceRefs) > 0 || attempt.ObservationCount > 0 {
			continue
		}
		count++
		if attempt.Ref != "" {
			refs = append(refs, attempt.Ref)
		}
		if attempt.FailureClass == FailureRepeatedNoProgress || count >= threshold {
			return true, count, normalizeAttemptRefs(refs)
		}
	}
	for _, merge := range normalizeDelegationWorkerParentMerges(parentMerges) {
		if merge.ReadyForParentMerge || merge.Status == VerificationSatisfied {
			continue
		}
		if len(merge.Observations) > 0 || len(merge.EvidenceRefs) > 0 {
			continue
		}
		count++
		if merge.WorkerAttemptRef != "" {
			refs = append(refs, merge.WorkerAttemptRef)
		}
		if merge.FailureClass == FailureRepeatedNoProgress || count >= threshold {
			return true, count, normalizeAttemptRefs(refs)
		}
	}
	return false, count, normalizeAttemptRefs(refs)
}

func delegationWorkerNoProgressThreshold(threshold int) int {
	if threshold <= 0 {
		return 2
	}
	return threshold
}

func delegationWorkerConflictRefs(parentMerges []DelegationWorkerParentMerge, explicit []DisplaySafeRef) (bool, []DisplaySafeRef) {
	refs := normalizeDisplaySafeRefs(explicit)
	type observedValue struct {
		value string
		ref   DisplaySafeRef
	}
	seen := map[string]observedValue{}
	for _, merge := range normalizeDelegationWorkerParentMerges(parentMerges) {
		fallbackRef := firstDisplaySafeRef(merge.MergeRef, merge.WorkerResultRef, merge.WorkerRunRef)
		for _, observation := range normalizeObservations(merge.Observations) {
			key := delegationWorkerObservationConflictKey(observation)
			if key == "" || observation.Value == "" {
				continue
			}
			ref := firstDisplaySafeRef(delegationWorkerObservationRef(observation), fallbackRef)
			if prior, exists := seen[key]; exists {
				if prior.value != observation.Value {
					refs = normalizeDisplaySafeRefs(append(refs, prior.ref, ref))
				}
				continue
			}
			seen[key] = observedValue{value: observation.Value, ref: ref}
		}
	}
	refs = normalizeDisplaySafeRefs(refs)
	return len(refs) > 0, refs
}

func delegationWorkerObservationConflictKey(observation Observation) string {
	observation = observation.Normalize()
	if observation.Kind == "" || observation.Name == "" {
		return ""
	}
	return string(observation.Subject) + "|" + observation.Kind + "|" + observation.Name + "|" + observation.Unit
}

func delegationWorkerObservationRef(observation Observation) DisplaySafeRef {
	observation = observation.Normalize()
	for _, ref := range observation.DisplaySafeRefs {
		if normalized := normalizeOneDisplaySafeRef(ref); normalized != "" {
			return normalized
		}
	}
	for _, evidence := range normalizeEvidenceRefs(observation.EvidenceRefs) {
		if evidence.Ref != "" {
			return evidence.Ref
		}
	}
	return ""
}

func delegationWorkerFailureReported(invocations []HostOwnedDelegationWorkerRuntimeInvocation) (bool, []DisplaySafeRef, FailureClass) {
	failed := false
	refs := []DisplaySafeRef{}
	failure := FailureNone
	for _, invocation := range normalizeDelegationWorkerRuntimeInvocations(invocations) {
		if !invocation.HostInvocationFailed && !invocation.ReadyForFailureReview {
			continue
		}
		failed = true
		failure = firstFailureClass(failure, invocation.FailureClass, FailureVerificationFailed)
		if invocation.WorkerRunRef != "" {
			refs = append(refs, invocation.WorkerRunRef)
		}
	}
	return failed, normalizeDisplaySafeRefs(refs), failure
}

func delegationWorkerFailureReviewClass(noProgress bool, conflict bool) FailureClass {
	switch {
	case noProgress:
		return FailureRepeatedNoProgress
	case conflict:
		return FailureVerificationFailed
	default:
		return FailureNone
	}
}

func delegationWorkerFailureReviewEvidenceRefs(invocations []HostOwnedDelegationWorkerRuntimeInvocation, parentMerges []DelegationWorkerParentMerge, attempts []AttemptSummary) []EvidenceRef {
	evidence := []EvidenceRef{}
	for _, invocation := range normalizeDelegationWorkerRuntimeInvocations(invocations) {
		if invocation.ReadyForFailureReview || invocation.HostInvocationFailed {
			evidence = MergeEvidenceRefs(evidence, invocation.EvidenceRefs)
		}
	}
	for _, merge := range normalizeDelegationWorkerParentMerges(parentMerges) {
		if !merge.ReadyForParentMerge {
			evidence = MergeEvidenceRefs(evidence, merge.EvidenceRefs)
		}
	}
	for _, attempt := range normalizeAttemptSummaries(attempts) {
		if attempt.Status != VerificationSatisfied {
			evidence = MergeEvidenceRefs(evidence, attempt.EvidenceRefs)
		}
	}
	return evidence
}
