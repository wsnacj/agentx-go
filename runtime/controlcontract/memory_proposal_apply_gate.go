package controlcontract

import "strings"

func displaySafeRefsContainAll(haystack []DisplaySafeRef, needles []DisplaySafeRef) bool {
	for _, needle := range normalizeDisplaySafeRefs(needles) {
		if !displaySafeRefSliceContains(haystack, needle) {
			return false
		}
	}
	return true
}

type MemoryProposalReviewPacketInput struct {
	ProposalSet                  RepeatedSuccessMemoryProposalSet `json:"proposal_set,omitempty"`
	ReviewPacketRef              DisplaySafeRef                   `json:"review_packet_ref,omitempty"`
	ReviewerRef                  DisplaySafeRef                   `json:"reviewer_ref,omitempty"`
	HostReviewRef                DisplaySafeRef                   `json:"host_review_ref,omitempty"`
	MemoryApplyPolicyRef         DisplaySafeRef                   `json:"memory_apply_policy_ref,omitempty"`
	MemoryApplyAdapterRef        DisplaySafeRef                   `json:"memory_apply_adapter_ref,omitempty"`
	SelectedProposalRefs         []DisplaySafeRef                 `json:"selected_proposal_refs,omitempty"`
	RejectedProposalRefs         []DisplaySafeRef                 `json:"rejected_proposal_refs,omitempty"`
	ExpectedMemoryArtifactRefs   []DisplaySafeRef                 `json:"expected_memory_artifact_refs,omitempty"`
	ExpectedMemoryApplyResultRef DisplaySafeRef                   `json:"expected_memory_apply_result_ref,omitempty"`
	ExpectedReadbackRef          DisplaySafeRef                   `json:"expected_readback_ref,omitempty"`
	IdempotencyRef               DisplaySafeRef                   `json:"idempotency_ref,omitempty"`
	RollbackPathRef              DisplaySafeRef                   `json:"rollback_path_ref,omitempty"`
	ApprovalRefs                 []DisplaySafeRef                 `json:"approval_refs,omitempty"`
	EvidenceRefs                 []EvidenceRef                    `json:"evidence_refs,omitempty"`
	DecisionBasis                []DisplaySafeRef                 `json:"decision_basis,omitempty"`
	Boundaries                   []Boundary                       `json:"boundaries,omitempty"`
	HostReviewCompleted          bool                             `json:"host_review_completed"`
	HostReviewApproved           bool                             `json:"host_review_approved"`
	HostRejectedUnsafeProposal   bool                             `json:"host_rejected_unsafe_proposal"`
	RawOutputLoaded              bool                             `json:"raw_output_loaded"`
}

type MemoryProposalReviewPacket struct {
	ContractVersion              string                           `json:"contract_version,omitempty"`
	Projected                    bool                             `json:"projected"`
	Status                       HostActionStatus                 `json:"status,omitempty"`
	ReadyForHostMemoryApply      bool                             `json:"ready_for_host_memory_apply"`
	HostMemoryApplyAuthorized    bool                             `json:"host_memory_apply_authorized"`
	HostMayApplyMemoryMutation   bool                             `json:"host_may_apply_memory_mutation"`
	ProposalOnly                 bool                             `json:"proposal_only"`
	ProposalSet                  RepeatedSuccessMemoryProposalSet `json:"proposal_set,omitempty"`
	ReviewPacketRef              DisplaySafeRef                   `json:"review_packet_ref,omitempty"`
	ReviewerRef                  DisplaySafeRef                   `json:"reviewer_ref,omitempty"`
	HostReviewRef                DisplaySafeRef                   `json:"host_review_ref,omitempty"`
	MemoryApplyPolicyRef         DisplaySafeRef                   `json:"memory_apply_policy_ref,omitempty"`
	MemoryApplyAdapterRef        DisplaySafeRef                   `json:"memory_apply_adapter_ref,omitempty"`
	SelectedProposalRefs         []DisplaySafeRef                 `json:"selected_proposal_refs,omitempty"`
	RejectedProposalRefs         []DisplaySafeRef                 `json:"rejected_proposal_refs,omitempty"`
	ExpectedMemoryArtifactRefs   []DisplaySafeRef                 `json:"expected_memory_artifact_refs,omitempty"`
	ExpectedMemoryApplyResultRef DisplaySafeRef                   `json:"expected_memory_apply_result_ref,omitempty"`
	ExpectedReadbackRef          DisplaySafeRef                   `json:"expected_readback_ref,omitempty"`
	IdempotencyRef               DisplaySafeRef                   `json:"idempotency_ref,omitempty"`
	RollbackPathRef              DisplaySafeRef                   `json:"rollback_path_ref,omitempty"`
	ApprovalRefs                 []DisplaySafeRef                 `json:"approval_refs,omitempty"`
	EvidenceRefs                 []EvidenceRef                    `json:"evidence_refs,omitempty"`
	FailureClass                 FailureClass                     `json:"failure_class,omitempty"`
	BlockedReasons               []string                         `json:"blocked_reasons,omitempty"`
	MissingInputs                []MissingInput                   `json:"missing_inputs,omitempty"`
	DecisionBasis                []DisplaySafeRef                 `json:"decision_basis,omitempty"`
	Boundaries                   []Boundary                       `json:"boundaries,omitempty"`
	NextHostAction               NextHostAction                   `json:"next_host_action,omitempty"`
	SkillWriteByCore             bool                             `json:"skill_write_by_core"`
	WorkflowWriteByCore          bool                             `json:"workflow_write_by_core"`
	TemplateWriteByCore          bool                             `json:"template_write_by_core"`
	InstallExecutedByCore        bool                             `json:"install_executed_by_core"`
	RuntimeReloadByCore          bool                             `json:"runtime_reload_by_core"`
	CoreMutationExecuted         bool                             `json:"core_mutation_executed"`
	RunnerDispatched             bool                             `json:"runner_dispatched"`
	RuntimeAdapterExecuted       bool                             `json:"runtime_adapter_executed"`
	WorkflowDispatched           bool                             `json:"workflow_dispatched"`
	WorkerDispatched             bool                             `json:"worker_dispatched"`
	StoreMutationExecuted        bool                             `json:"store_mutation_executed"`
	CompensationExecuted         bool                             `json:"compensation_executed"`
	RunnerEffect                 string                           `json:"runner_effect,omitempty"`
	PromptEffect                 string                           `json:"prompt_effect,omitempty"`
	RuntimeEffect                string                           `json:"runtime_effect,omitempty"`
	RawOutputLoaded              bool                             `json:"raw_output_loaded"`
}

type HostOwnedMemoryProposalApplyReadinessInput struct {
	ReviewPacket         MemoryProposalReviewPacket             `json:"review_packet,omitempty"`
	MemoryApplyGate      ProductionAdapterIndependentEffectGate `json:"memory_apply_gate,omitempty"`
	AdapterRef           DisplaySafeRef                         `json:"adapter_ref,omitempty"`
	AdapterVersionRef    DisplaySafeRef                         `json:"adapter_version_ref,omitempty"`
	AdapterCapabilityRef DisplaySafeRef                         `json:"adapter_capability_ref,omitempty"`
	AdapterContractRef   DisplaySafeRef                         `json:"adapter_contract_ref,omitempty"`
	HostConfirmationRef  DisplaySafeRef                         `json:"host_confirmation_ref,omitempty"`
	AdapterDryRunRef     DisplaySafeRef                         `json:"adapter_dry_run_ref,omitempty"`
	InvocationRef        DisplaySafeRef                         `json:"invocation_ref,omitempty"`
	ResultBindingRef     DisplaySafeRef                         `json:"result_binding_ref,omitempty"`
	ReadbackBindingRef   DisplaySafeRef                         `json:"readback_binding_ref,omitempty"`
	IdempotencyRef       DisplaySafeRef                         `json:"idempotency_ref,omitempty"`
	RollbackRef          DisplaySafeRef                         `json:"rollback_ref,omitempty"`
	FailureBindingRef    DisplaySafeRef                         `json:"failure_binding_ref,omitempty"`
	CompensationRef      DisplaySafeRef                         `json:"compensation_ref,omitempty"`
	ApprovalRefs         []DisplaySafeRef                       `json:"approval_refs,omitempty"`
	PolicyRefs           []DisplaySafeRef                       `json:"policy_refs,omitempty"`
	EvidenceRefs         []EvidenceRef                          `json:"evidence_refs,omitempty"`
	DecisionBasis        []DisplaySafeRef                       `json:"decision_basis,omitempty"`
	Boundaries           []Boundary                             `json:"boundaries,omitempty"`
	RawOutputLoaded      bool                                   `json:"raw_output_loaded"`
}

type HostOwnedMemoryProposalApplyReadiness struct {
	ContractVersion                       string                                 `json:"contract_version,omitempty"`
	Projected                             bool                                   `json:"projected"`
	Status                                HostActionStatus                       `json:"status,omitempty"`
	ReadyForHostMemoryAdapterInvocation   bool                                   `json:"ready_for_host_memory_adapter_invocation"`
	HostMemoryAdapterInvocationAuthorized bool                                   `json:"host_memory_adapter_invocation_authorized"`
	HostMayInvokeMemoryAdapter            bool                                   `json:"host_may_invoke_memory_adapter"`
	ReviewPacket                          MemoryProposalReviewPacket             `json:"review_packet,omitempty"`
	MemoryApplyGate                       ProductionAdapterIndependentEffectGate `json:"memory_apply_gate,omitempty"`
	ReviewPacketRef                       DisplaySafeRef                         `json:"review_packet_ref,omitempty"`
	MemoryApplyPolicyRef                  DisplaySafeRef                         `json:"memory_apply_policy_ref,omitempty"`
	MemoryApplyAdapterRef                 DisplaySafeRef                         `json:"memory_apply_adapter_ref,omitempty"`
	SelectedProposalRefs                  []DisplaySafeRef                       `json:"selected_proposal_refs,omitempty"`
	ExpectedMemoryArtifactRefs            []DisplaySafeRef                       `json:"expected_memory_artifact_refs,omitempty"`
	ExpectedMemoryApplyResultRef          DisplaySafeRef                         `json:"expected_memory_apply_result_ref,omitempty"`
	ExpectedReadbackRef                   DisplaySafeRef                         `json:"expected_readback_ref,omitempty"`
	RollbackPathRef                       DisplaySafeRef                         `json:"rollback_path_ref,omitempty"`
	AdapterRef                            DisplaySafeRef                         `json:"adapter_ref,omitempty"`
	AdapterVersionRef                     DisplaySafeRef                         `json:"adapter_version_ref,omitempty"`
	AdapterCapabilityRef                  DisplaySafeRef                         `json:"adapter_capability_ref,omitempty"`
	AdapterContractRef                    DisplaySafeRef                         `json:"adapter_contract_ref,omitempty"`
	HostConfirmationRef                   DisplaySafeRef                         `json:"host_confirmation_ref,omitempty"`
	AdapterDryRunRef                      DisplaySafeRef                         `json:"adapter_dry_run_ref,omitempty"`
	InvocationRef                         DisplaySafeRef                         `json:"invocation_ref,omitempty"`
	ResultBindingRef                      DisplaySafeRef                         `json:"result_binding_ref,omitempty"`
	ReadbackBindingRef                    DisplaySafeRef                         `json:"readback_binding_ref,omitempty"`
	IdempotencyRef                        DisplaySafeRef                         `json:"idempotency_ref,omitempty"`
	RollbackRef                           DisplaySafeRef                         `json:"rollback_ref,omitempty"`
	FailureBindingRef                     DisplaySafeRef                         `json:"failure_binding_ref,omitempty"`
	CompensationRef                       DisplaySafeRef                         `json:"compensation_ref,omitempty"`
	ApprovalRefs                          []DisplaySafeRef                       `json:"approval_refs,omitempty"`
	PolicyRefs                            []DisplaySafeRef                       `json:"policy_refs,omitempty"`
	EvidenceRefs                          []EvidenceRef                          `json:"evidence_refs,omitempty"`
	FailureClass                          FailureClass                           `json:"failure_class,omitempty"`
	BlockedReasons                        []string                               `json:"blocked_reasons,omitempty"`
	MissingInputs                         []MissingInput                         `json:"missing_inputs,omitempty"`
	DecisionBasis                         []DisplaySafeRef                       `json:"decision_basis,omitempty"`
	Boundaries                            []Boundary                             `json:"boundaries,omitempty"`
	NextHostAction                        NextHostAction                         `json:"next_host_action,omitempty"`
	SkillWriteByCore                      bool                                   `json:"skill_write_by_core"`
	WorkflowWriteByCore                   bool                                   `json:"workflow_write_by_core"`
	TemplateWriteByCore                   bool                                   `json:"template_write_by_core"`
	InstallExecutedByCore                 bool                                   `json:"install_executed_by_core"`
	RuntimeReloadByCore                   bool                                   `json:"runtime_reload_by_core"`
	CoreMutationExecuted                  bool                                   `json:"core_mutation_executed"`
	RunnerDispatched                      bool                                   `json:"runner_dispatched"`
	RuntimeAdapterExecuted                bool                                   `json:"runtime_adapter_executed"`
	WorkflowDispatched                    bool                                   `json:"workflow_dispatched"`
	WorkerDispatched                      bool                                   `json:"worker_dispatched"`
	StoreMutationExecuted                 bool                                   `json:"store_mutation_executed"`
	CompensationExecuted                  bool                                   `json:"compensation_executed"`
	RunnerEffect                          string                                 `json:"runner_effect,omitempty"`
	PromptEffect                          string                                 `json:"prompt_effect,omitempty"`
	RuntimeEffect                         string                                 `json:"runtime_effect,omitempty"`
	RawOutputLoaded                       bool                                   `json:"raw_output_loaded"`
}

type HostOwnedMemoryProposalApplyInvocationInput struct {
	Readiness                      HostOwnedMemoryProposalApplyReadiness `json:"readiness,omitempty"`
	InvocationReportRef            DisplaySafeRef                        `json:"invocation_report_ref,omitempty"`
	ObservedInvocationRef          DisplaySafeRef                        `json:"observed_invocation_ref,omitempty"`
	HostMemoryAdapterRunRef        DisplaySafeRef                        `json:"host_memory_adapter_run_ref,omitempty"`
	MemoryApplyResultRef           DisplaySafeRef                        `json:"memory_apply_result_ref,omitempty"`
	MemoryReadbackRef              DisplaySafeRef                        `json:"memory_readback_ref,omitempty"`
	AppliedMemoryArtifactRefs      []DisplaySafeRef                      `json:"applied_memory_artifact_refs,omitempty"`
	ReadbackMemoryArtifactRefs     []DisplaySafeRef                      `json:"readback_memory_artifact_refs,omitempty"`
	FailureRef                     DisplaySafeRef                        `json:"failure_ref,omitempty"`
	CompensationRef                DisplaySafeRef                        `json:"compensation_ref,omitempty"`
	HostAdapterInvocationReported  bool                                  `json:"host_adapter_invocation_reported"`
	HostAdapterInvocationCompleted bool                                  `json:"host_adapter_invocation_completed"`
	HostAdapterInvocationFailed    bool                                  `json:"host_adapter_invocation_failed"`
	MemoryEvidenceRefs             []DisplaySafeRef                      `json:"memory_evidence_refs,omitempty"`
	Boundaries                     []Boundary                            `json:"boundaries,omitempty"`
	RawOutputLoaded                bool                                  `json:"raw_output_loaded"`
}

type HostOwnedMemoryProposalApplyInvocation struct {
	ContractVersion                string                                `json:"contract_version,omitempty"`
	Projected                      bool                                  `json:"projected"`
	Status                         HostActionStatus                      `json:"status,omitempty"`
	ReadyForMemoryApplyReadback    bool                                  `json:"ready_for_memory_apply_readback"`
	ReadyForFailureReview          bool                                  `json:"ready_for_failure_review"`
	HostAdapterInvocationReported  bool                                  `json:"host_adapter_invocation_reported"`
	HostAdapterInvocationCompleted bool                                  `json:"host_adapter_invocation_completed"`
	HostAdapterInvocationFailed    bool                                  `json:"host_adapter_invocation_failed"`
	Readiness                      HostOwnedMemoryProposalApplyReadiness `json:"readiness,omitempty"`
	ReviewPacketRef                DisplaySafeRef                        `json:"review_packet_ref,omitempty"`
	MemoryApplyAdapterRef          DisplaySafeRef                        `json:"memory_apply_adapter_ref,omitempty"`
	SelectedProposalRefs           []DisplaySafeRef                      `json:"selected_proposal_refs,omitempty"`
	ExpectedMemoryArtifactRefs     []DisplaySafeRef                      `json:"expected_memory_artifact_refs,omitempty"`
	ExpectedMemoryApplyResultRef   DisplaySafeRef                        `json:"expected_memory_apply_result_ref,omitempty"`
	ExpectedReadbackRef            DisplaySafeRef                        `json:"expected_readback_ref,omitempty"`
	RollbackPathRef                DisplaySafeRef                        `json:"rollback_path_ref,omitempty"`
	AdapterRef                     DisplaySafeRef                        `json:"adapter_ref,omitempty"`
	AdapterVersionRef              DisplaySafeRef                        `json:"adapter_version_ref,omitempty"`
	InvocationRef                  DisplaySafeRef                        `json:"invocation_ref,omitempty"`
	InvocationReportRef            DisplaySafeRef                        `json:"invocation_report_ref,omitempty"`
	ObservedInvocationRef          DisplaySafeRef                        `json:"observed_invocation_ref,omitempty"`
	HostMemoryAdapterRunRef        DisplaySafeRef                        `json:"host_memory_adapter_run_ref,omitempty"`
	MemoryApplyResultRef           DisplaySafeRef                        `json:"memory_apply_result_ref,omitempty"`
	MemoryReadbackRef              DisplaySafeRef                        `json:"memory_readback_ref,omitempty"`
	AppliedMemoryArtifactRefs      []DisplaySafeRef                      `json:"applied_memory_artifact_refs,omitempty"`
	ReadbackMemoryArtifactRefs     []DisplaySafeRef                      `json:"readback_memory_artifact_refs,omitempty"`
	FailureRef                     DisplaySafeRef                        `json:"failure_ref,omitempty"`
	CompensationRef                DisplaySafeRef                        `json:"compensation_ref,omitempty"`
	MemoryEvidenceRefs             []DisplaySafeRef                      `json:"memory_evidence_refs,omitempty"`
	FailureClass                   FailureClass                          `json:"failure_class,omitempty"`
	BlockedReasons                 []string                              `json:"blocked_reasons,omitempty"`
	MissingInputs                  []MissingInput                        `json:"missing_inputs,omitempty"`
	Boundaries                     []Boundary                            `json:"boundaries,omitempty"`
	NextHostAction                 NextHostAction                        `json:"next_host_action,omitempty"`
	SkillWriteByCore               bool                                  `json:"skill_write_by_core"`
	WorkflowWriteByCore            bool                                  `json:"workflow_write_by_core"`
	TemplateWriteByCore            bool                                  `json:"template_write_by_core"`
	InstallExecutedByCore          bool                                  `json:"install_executed_by_core"`
	RuntimeReloadByCore            bool                                  `json:"runtime_reload_by_core"`
	CoreMutationExecuted           bool                                  `json:"core_mutation_executed"`
	RunnerDispatched               bool                                  `json:"runner_dispatched"`
	RuntimeAdapterExecuted         bool                                  `json:"runtime_adapter_executed"`
	WorkflowDispatched             bool                                  `json:"workflow_dispatched"`
	WorkerDispatched               bool                                  `json:"worker_dispatched"`
	StoreMutationExecuted          bool                                  `json:"store_mutation_executed"`
	CompensationExecuted           bool                                  `json:"compensation_executed"`
	RunnerEffect                   string                                `json:"runner_effect,omitempty"`
	PromptEffect                   string                                `json:"prompt_effect,omitempty"`
	RuntimeEffect                  string                                `json:"runtime_effect,omitempty"`
	RawOutputLoaded                bool                                  `json:"raw_output_loaded"`
}

func BuildMemoryProposalReviewPacket(input MemoryProposalReviewPacketInput) MemoryProposalReviewPacket {
	proposalSet := input.ProposalSet.Normalize()
	result := MemoryProposalReviewPacket{
		ContractVersion:              ContractVersion,
		Projected:                    true,
		Status:                       HostActionBlocked,
		ProposalOnly:                 true,
		ProposalSet:                  proposalSet,
		ReviewPacketRef:              normalizeOneDisplaySafeRef(input.ReviewPacketRef),
		ReviewerRef:                  normalizeOneDisplaySafeRef(input.ReviewerRef),
		HostReviewRef:                normalizeOneDisplaySafeRef(input.HostReviewRef),
		MemoryApplyPolicyRef:         normalizeOneDisplaySafeRef(input.MemoryApplyPolicyRef),
		MemoryApplyAdapterRef:        normalizeOneDisplaySafeRef(input.MemoryApplyAdapterRef),
		SelectedProposalRefs:         normalizeDisplaySafeRefs(input.SelectedProposalRefs),
		RejectedProposalRefs:         normalizeDisplaySafeRefs(input.RejectedProposalRefs),
		ExpectedMemoryArtifactRefs:   normalizeDisplaySafeRefs(input.ExpectedMemoryArtifactRefs),
		ExpectedMemoryApplyResultRef: normalizeOneDisplaySafeRef(input.ExpectedMemoryApplyResultRef),
		ExpectedReadbackRef:          normalizeOneDisplaySafeRef(input.ExpectedReadbackRef),
		IdempotencyRef:               normalizeOneDisplaySafeRef(input.IdempotencyRef),
		RollbackPathRef:              normalizeOneDisplaySafeRef(input.RollbackPathRef),
		ApprovalRefs:                 normalizeDisplaySafeRefs(input.ApprovalRefs),
		EvidenceRefs:                 MergeEvidenceRefs(input.EvidenceRefs, proposalSet.EvidenceRefs),
		FailureClass:                 FailureNone,
		DecisionBasis: normalizeDisplaySafeRefs(append([]DisplaySafeRef{
			"memory_proposal_review:host_owned",
			"memory_proposal_apply:proposal_not_apply",
		}, input.DecisionBasis...)),
		Boundaries:      memoryProposalReviewPacketBoundaries(proposalSet.Boundaries, input.Boundaries),
		NextHostAction:  "review_memory_proposal_packet",
		RunnerEffect:    "none",
		PromptEffect:    "none",
		RuntimeEffect:   "none",
		RawOutputLoaded: input.RawOutputLoaded || proposalSet.RawOutputLoaded,
	}
	if memoryProposalReviewPacketUnsafe(input, proposalSet) {
		result.RawOutputLoaded = true
		result = memoryProposalReviewPacketBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if !proposalSet.ReadyForHostReview {
		result = memoryProposalReviewPacketBlock(result, firstFailureClass(proposalSet.FailureClass, FailureConfigMissing), "memory_proposal_set_not_ready", "host:memory_proposal_set", firstNextHostAction(proposalSet.NextHostAction, "review_memory_proposal_set"))
	}
	for _, check := range []struct {
		ok      bool
		reason  string
		missing MissingInput
		next    NextHostAction
	}{
		{result.ReviewPacketRef != "", "review_packet_ref_missing", "host:memory_review_packet_ref", "provide_memory_review_packet_ref"},
		{result.ReviewerRef != "", "reviewer_ref_missing", "host:memory_reviewer_ref", "provide_memory_reviewer"},
		{result.HostReviewRef != "", "host_review_ref_missing", "host:memory_review_ref", "provide_memory_review_ref"},
		{result.MemoryApplyPolicyRef != "", "memory_apply_policy_ref_missing", "host:memory_apply_policy_ref", "provide_memory_apply_policy"},
		{result.MemoryApplyAdapterRef != "", "memory_apply_adapter_ref_missing", "host:memory_apply_adapter_ref", "provide_memory_apply_adapter_binding"},
		{input.HostReviewCompleted, "host_review_not_completed", "host:memory_review_completed", "complete_memory_proposal_review"},
		{input.HostReviewApproved, "host_review_not_approved", "host:memory_review_approval", "approve_memory_proposal_review"},
		{!input.HostRejectedUnsafeProposal, "unsafe_proposal_rejected_by_host", "host:memory_proposal_review", "select_safe_memory_proposals"},
		{len(result.SelectedProposalRefs) > 0, "selected_proposal_refs_missing", "host:selected_memory_proposal_refs", "select_memory_proposals"},
		{len(result.ExpectedMemoryArtifactRefs) > 0, "expected_memory_artifact_refs_missing", "host:expected_memory_artifact_refs", "provide_expected_memory_artifact_refs"},
		{result.ExpectedMemoryApplyResultRef != "", "expected_memory_apply_result_ref_missing", "host:expected_memory_apply_result_ref", "provide_memory_apply_result_binding"},
		{result.ExpectedReadbackRef != "", "expected_readback_ref_missing", "host:memory_apply_readback_ref", "provide_memory_apply_readback_binding"},
		{result.IdempotencyRef != "", "idempotency_ref_missing", "host:memory_apply_idempotency_ref", "provide_memory_apply_idempotency_ref"},
		{result.RollbackPathRef != "", "rollback_path_ref_missing", "host:memory_apply_rollback_path_ref", "provide_memory_apply_rollback_path"},
		{len(result.ApprovalRefs) > 0, "approval_refs_missing", "host:memory_apply_approval_ref", "provide_memory_apply_approval"},
	} {
		if !check.ok {
			result = memoryProposalReviewPacketBlock(result, FailureConfigMissing, check.reason, check.missing, check.next)
		}
	}
	if len(result.SelectedProposalRefs) > 0 && !memoryProposalRefsInSet(result.SelectedProposalRefs, proposalSet.Proposals) {
		result = memoryProposalReviewPacketBlock(result, FailureVerificationFailed, "selected_proposal_ref_not_in_set", "host:selected_memory_proposal_refs", "review_memory_proposal_selection")
	}
	if len(result.ExpectedMemoryArtifactRefs) > 0 && len(result.SelectedProposalRefs) > 0 && len(result.ExpectedMemoryArtifactRefs) != len(result.SelectedProposalRefs) {
		result = memoryProposalReviewPacketBlock(result, FailureInvalidInput, "expected_artifact_count_mismatch", "host:expected_memory_artifact_refs", "review_memory_artifact_bindings")
	}
	if len(result.MissingInputs) == 0 && len(result.BlockedReasons) == 0 {
		result.Status = HostActionReady
		result.ReadyForHostMemoryApply = true
		result.HostMemoryApplyAuthorized = true
		result.HostMayApplyMemoryMutation = true
		result.NextHostAction = "host_may_apply_reviewed_memory_proposals"
		result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_host_memory_apply", "host_may_apply_reviewed_memory_proposals")
	}
	return result.Normalize()
}

func BuildHostOwnedMemoryProposalApplyReadiness(input HostOwnedMemoryProposalApplyReadinessInput) HostOwnedMemoryProposalApplyReadiness {
	review := input.ReviewPacket.Normalize()
	gate := input.MemoryApplyGate.Normalize()
	result := HostOwnedMemoryProposalApplyReadiness{
		ContractVersion:              ContractVersion,
		Projected:                    true,
		Status:                       HostActionBlocked,
		ReviewPacket:                 review,
		MemoryApplyGate:              gate,
		ReviewPacketRef:              review.ReviewPacketRef,
		MemoryApplyPolicyRef:         review.MemoryApplyPolicyRef,
		MemoryApplyAdapterRef:        review.MemoryApplyAdapterRef,
		SelectedProposalRefs:         cloneDisplaySafeRefs(review.SelectedProposalRefs),
		ExpectedMemoryArtifactRefs:   cloneDisplaySafeRefs(review.ExpectedMemoryArtifactRefs),
		ExpectedMemoryApplyResultRef: review.ExpectedMemoryApplyResultRef,
		ExpectedReadbackRef:          review.ExpectedReadbackRef,
		RollbackPathRef:              review.RollbackPathRef,
		AdapterRef:                   normalizeOneDisplaySafeRef(input.AdapterRef),
		AdapterVersionRef:            normalizeOneDisplaySafeRef(input.AdapterVersionRef),
		AdapterCapabilityRef:         normalizeOneDisplaySafeRef(input.AdapterCapabilityRef),
		AdapterContractRef:           normalizeOneDisplaySafeRef(input.AdapterContractRef),
		HostConfirmationRef:          normalizeOneDisplaySafeRef(input.HostConfirmationRef),
		AdapterDryRunRef:             normalizeOneDisplaySafeRef(input.AdapterDryRunRef),
		InvocationRef:                normalizeOneDisplaySafeRef(input.InvocationRef),
		ResultBindingRef:             normalizeOneDisplaySafeRef(input.ResultBindingRef),
		ReadbackBindingRef:           normalizeOneDisplaySafeRef(input.ReadbackBindingRef),
		IdempotencyRef:               normalizeOneDisplaySafeRef(input.IdempotencyRef),
		RollbackRef:                  normalizeOneDisplaySafeRef(input.RollbackRef),
		FailureBindingRef:            normalizeOneDisplaySafeRef(input.FailureBindingRef),
		CompensationRef:              normalizeOneDisplaySafeRef(input.CompensationRef),
		ApprovalRefs:                 normalizeDisplaySafeRefs(append(cloneDisplaySafeRefs(review.ApprovalRefs), input.ApprovalRefs...)),
		PolicyRefs:                   normalizeDisplaySafeRefs(append([]DisplaySafeRef{review.MemoryApplyPolicyRef, gate.PolicyRef}, input.PolicyRefs...)),
		EvidenceRefs:                 MergeEvidenceRefs(input.EvidenceRefs, review.EvidenceRefs),
		FailureClass:                 FailureNone,
		DecisionBasis: normalizeDisplaySafeRefs(append([]DisplaySafeRef{
			"memory_proposal_apply:host_owned",
			"memory_proposal_apply:adapter_readiness",
		}, input.DecisionBasis...)),
		Boundaries:      hostOwnedMemoryProposalApplyReadinessBoundaries(review.Boundaries, gate.Boundaries, input.Boundaries),
		NextHostAction:  "provide_memory_apply_adapter_binding",
		RunnerEffect:    "none",
		PromptEffect:    "none",
		RuntimeEffect:   "none",
		RawOutputLoaded: input.RawOutputLoaded || review.RawOutputLoaded || gate.RawOutputLoaded,
	}
	if hostOwnedMemoryProposalApplyReadinessUnsafe(input, review, gate) {
		result.RawOutputLoaded = true
		result = hostOwnedMemoryProposalApplyReadinessBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if !review.ReadyForHostMemoryApply {
		result = hostOwnedMemoryProposalApplyReadinessBlock(result, firstFailureClass(review.FailureClass, FailureConfigMissing), "memory_review_packet_not_ready", "host:memory_review_packet", firstNextHostAction(review.NextHostAction, "review_memory_proposal_packet"))
	}
	if !gate.ReadyForIndependentGatePlan || gate.Kind != ProductionAdapterEffectGateMemoryApply {
		result = hostOwnedMemoryProposalApplyReadinessBlock(result, firstFailureClass(gate.FailureClass, FailureHostAdapterMissing), "memory_apply_independent_gate_not_ready", "host:memory_apply_gate_ref", firstNextHostAction(gate.NextHostAction, "provide_memory_apply_independent_gate"))
	}
	for _, check := range []struct {
		ok      bool
		reason  string
		missing MissingInput
		next    NextHostAction
	}{
		{result.AdapterRef != "", "adapter_ref_missing", "host:memory_apply_adapter_ref", "provide_memory_apply_adapter_binding"},
		{result.AdapterVersionRef != "", "adapter_version_ref_missing", "host:memory_apply_adapter_version_ref", "provide_memory_apply_adapter_binding"},
		{result.AdapterCapabilityRef != "", "adapter_capability_ref_missing", "host:memory_apply_adapter_capability_ref", "provide_memory_apply_adapter_capability"},
		{result.AdapterContractRef != "", "adapter_contract_ref_missing", "contract:memory_apply_adapter", "provide_memory_apply_adapter_contract"},
		{result.HostConfirmationRef != "", "host_confirmation_ref_missing", "host:memory_apply_adapter_confirmation", "request_memory_apply_adapter_confirmation"},
		{result.AdapterDryRunRef != "", "adapter_dry_run_ref_missing", "host:memory_apply_adapter_dry_run_ref", "provide_memory_apply_adapter_dry_run"},
		{result.InvocationRef != "", "invocation_ref_missing", "host:memory_apply_adapter_invocation_ref", "provide_memory_apply_adapter_invocation_ref"},
		{result.ResultBindingRef != "", "result_binding_ref_missing", "host:memory_apply_result_binding", "provide_memory_apply_result_binding"},
		{result.ReadbackBindingRef != "", "readback_binding_ref_missing", "host:memory_apply_readback_binding", "provide_memory_apply_readback_binding"},
		{result.IdempotencyRef != "", "idempotency_ref_missing", "host:memory_apply_idempotency_ref", "provide_memory_apply_idempotency_ref"},
		{result.RollbackRef != "", "rollback_ref_missing", "host:memory_apply_rollback_path_ref", "provide_memory_apply_rollback_path"},
		{result.FailureBindingRef != "", "failure_binding_ref_missing", "host:memory_apply_failure_ref", "provide_memory_apply_failure_binding"},
		{result.CompensationRef != "", "compensation_ref_missing", "host:memory_apply_compensation_ref", "provide_memory_apply_compensation_binding"},
	} {
		if !check.ok {
			result = hostOwnedMemoryProposalApplyReadinessBlock(result, FailureConfigMissing, check.reason, check.missing, check.next)
		}
	}
	for _, check := range []struct {
		got     DisplaySafeRef
		want    DisplaySafeRef
		reason  string
		missing MissingInput
		next    NextHostAction
	}{
		{result.AdapterRef, review.MemoryApplyAdapterRef, "adapter_ref_mismatch", "host:memory_apply_adapter_ref", "review_memory_apply_adapter_binding"},
		{result.HostConfirmationRef, firstDisplaySafeRef(review.ApprovalRefs...), "host_confirmation_ref_mismatch", "host:memory_apply_adapter_confirmation", "review_memory_apply_adapter_binding"},
		{result.ResultBindingRef, review.ExpectedMemoryApplyResultRef, "result_binding_ref_mismatch", "host:memory_apply_result_binding", "review_memory_apply_adapter_binding"},
		{result.ReadbackBindingRef, review.ExpectedReadbackRef, "readback_binding_ref_mismatch", "host:memory_apply_readback_binding", "review_memory_apply_adapter_binding"},
		{result.IdempotencyRef, review.IdempotencyRef, "idempotency_ref_mismatch", "host:memory_apply_idempotency_ref", "review_memory_apply_adapter_binding"},
		{result.RollbackRef, review.RollbackPathRef, "rollback_ref_mismatch", "host:memory_apply_rollback_path_ref", "review_memory_apply_adapter_binding"},
	} {
		if check.want != "" && check.got != "" && check.got != check.want {
			result = hostOwnedMemoryProposalApplyReadinessBlock(result, FailureVerificationFailed, check.reason, check.missing, check.next)
		}
	}
	if len(result.MissingInputs) == 0 && len(result.BlockedReasons) == 0 {
		result.Status = HostActionReady
		result.ReadyForHostMemoryAdapterInvocation = true
		result.HostMemoryAdapterInvocationAuthorized = true
		result.HostMayInvokeMemoryAdapter = true
		result.NextHostAction = "host_may_invoke_memory_apply_adapter"
		result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_host_owned_memory_apply_adapter_invocation", "host_may_invoke_memory_apply_adapter")
	}
	return result.Normalize()
}

func BuildHostOwnedMemoryProposalApplyInvocation(input HostOwnedMemoryProposalApplyInvocationInput) HostOwnedMemoryProposalApplyInvocation {
	readiness := input.Readiness.Normalize()
	result := HostOwnedMemoryProposalApplyInvocation{
		ContractVersion:                ContractVersion,
		Projected:                      true,
		Status:                         HostActionBlocked,
		Readiness:                      readiness,
		ReviewPacketRef:                readiness.ReviewPacketRef,
		MemoryApplyAdapterRef:          readiness.MemoryApplyAdapterRef,
		SelectedProposalRefs:           cloneDisplaySafeRefs(readiness.SelectedProposalRefs),
		ExpectedMemoryArtifactRefs:     cloneDisplaySafeRefs(readiness.ExpectedMemoryArtifactRefs),
		ExpectedMemoryApplyResultRef:   readiness.ExpectedMemoryApplyResultRef,
		ExpectedReadbackRef:            readiness.ExpectedReadbackRef,
		RollbackPathRef:                readiness.RollbackPathRef,
		AdapterRef:                     readiness.AdapterRef,
		AdapterVersionRef:              readiness.AdapterVersionRef,
		InvocationRef:                  readiness.InvocationRef,
		InvocationReportRef:            normalizeOneDisplaySafeRef(input.InvocationReportRef),
		ObservedInvocationRef:          normalizeOneDisplaySafeRef(input.ObservedInvocationRef),
		HostMemoryAdapterRunRef:        normalizeOneDisplaySafeRef(input.HostMemoryAdapterRunRef),
		MemoryApplyResultRef:           normalizeOneDisplaySafeRef(input.MemoryApplyResultRef),
		MemoryReadbackRef:              normalizeOneDisplaySafeRef(input.MemoryReadbackRef),
		AppliedMemoryArtifactRefs:      normalizeDisplaySafeRefs(input.AppliedMemoryArtifactRefs),
		ReadbackMemoryArtifactRefs:     normalizeDisplaySafeRefs(input.ReadbackMemoryArtifactRefs),
		FailureRef:                     normalizeOneDisplaySafeRef(input.FailureRef),
		CompensationRef:                normalizeOneDisplaySafeRef(input.CompensationRef),
		MemoryEvidenceRefs:             normalizeDisplaySafeRefs(input.MemoryEvidenceRefs),
		FailureClass:                   FailureNone,
		HostAdapterInvocationReported:  input.HostAdapterInvocationReported,
		HostAdapterInvocationCompleted: input.HostAdapterInvocationCompleted,
		HostAdapterInvocationFailed:    input.HostAdapterInvocationFailed,
		Boundaries:                     hostOwnedMemoryProposalApplyInvocationBoundaries(readiness.Boundaries, input.Boundaries),
		NextHostAction:                 "provide_memory_apply_adapter_invocation_report",
		RunnerEffect:                   "none",
		PromptEffect:                   "none",
		RuntimeEffect:                  "none",
		RawOutputLoaded:                input.RawOutputLoaded || readiness.RawOutputLoaded,
	}
	if hostOwnedMemoryProposalApplyInvocationUnsafe(input, readiness) {
		result.RawOutputLoaded = true
		result = hostOwnedMemoryProposalApplyInvocationBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if !readiness.ReadyForHostMemoryAdapterInvocation {
		result = hostOwnedMemoryProposalApplyInvocationBlock(result, firstFailureClass(readiness.FailureClass, FailureConfigMissing), "adapter_readiness_not_ready", "host:memory_apply_adapter_readiness", firstNextHostAction(readiness.NextHostAction, "review_memory_apply_adapter_readiness"))
	}
	if result.InvocationReportRef == "" {
		result = hostOwnedMemoryProposalApplyInvocationBlock(result, FailureEvidenceMissing, "invocation_report_ref_missing", "host:memory_apply_adapter_invocation_report_ref", "provide_memory_apply_adapter_invocation_report")
	}
	if result.ObservedInvocationRef == "" {
		result = hostOwnedMemoryProposalApplyInvocationBlock(result, FailureEvidenceMissing, "observed_invocation_ref_missing", "host:memory_apply_adapter_invocation_ref", "provide_memory_apply_adapter_invocation_report")
	} else if result.InvocationRef != "" && result.ObservedInvocationRef != result.InvocationRef {
		result = hostOwnedMemoryProposalApplyInvocationBlock(result, FailureVerificationFailed, "observed_invocation_ref_mismatch", "host:memory_apply_adapter_invocation_ref", "review_memory_apply_adapter_invocation_report")
	}
	if result.HostMemoryAdapterRunRef == "" {
		result = hostOwnedMemoryProposalApplyInvocationBlock(result, FailureEvidenceMissing, "host_memory_adapter_run_ref_missing", "host:memory_apply_adapter_run_ref", "provide_memory_apply_adapter_invocation_report")
	}
	if !result.HostAdapterInvocationReported {
		result = hostOwnedMemoryProposalApplyInvocationBlock(result, FailureEvidenceMissing, "host_adapter_invocation_not_reported", "host:memory_apply_adapter_invocation_report", "provide_memory_apply_adapter_invocation_report")
	}
	if !result.HostAdapterInvocationCompleted && !result.HostAdapterInvocationFailed {
		result = hostOwnedMemoryProposalApplyInvocationBlock(result, FailureEvidenceMissing, "host_adapter_invocation_status_missing", "host:memory_apply_adapter_invocation_status", "provide_memory_apply_adapter_invocation_report")
	}
	if result.HostAdapterInvocationCompleted && result.HostAdapterInvocationFailed {
		result = hostOwnedMemoryProposalApplyInvocationBlock(result, FailureInvalidInput, "host_adapter_invocation_status_conflict", "host:memory_apply_adapter_invocation_status", "review_memory_apply_adapter_invocation_report")
	}
	if result.HostAdapterInvocationFailed {
		if result.FailureRef == "" {
			result = hostOwnedMemoryProposalApplyInvocationBlock(result, FailureEvidenceMissing, "failure_ref_missing", "host:memory_apply_adapter_failure_ref", "provide_memory_apply_adapter_failure_ref")
		}
		if result.CompensationRef == "" {
			result = hostOwnedMemoryProposalApplyInvocationBlock(result, FailureEvidenceMissing, "compensation_ref_missing", "host:memory_apply_adapter_compensation_ref", "provide_memory_apply_adapter_compensation_ref")
		}
	} else {
		for _, check := range []struct {
			got     DisplaySafeRef
			want    DisplaySafeRef
			reason  string
			missing MissingInput
			next    NextHostAction
		}{
			{result.MemoryApplyResultRef, readiness.ExpectedMemoryApplyResultRef, "memory_apply_result_ref_mismatch", "host:memory_apply_result_ref", "review_memory_apply_adapter_invocation_report"},
			{result.MemoryReadbackRef, readiness.ExpectedReadbackRef, "memory_readback_ref_mismatch", "host:memory_apply_readback_ref", "review_memory_apply_adapter_invocation_report"},
		} {
			if check.got == "" {
				result = hostOwnedMemoryProposalApplyInvocationBlock(result, FailureEvidenceMissing, strings.TrimSuffix(check.reason, "_mismatch")+"_missing", check.missing, "provide_memory_apply_adapter_invocation_report")
			} else if check.want != "" && check.got != check.want {
				result = hostOwnedMemoryProposalApplyInvocationBlock(result, FailureVerificationFailed, check.reason, check.missing, check.next)
			}
		}
		if !displaySafeRefsContainAll(result.AppliedMemoryArtifactRefs, readiness.ExpectedMemoryArtifactRefs) {
			result = hostOwnedMemoryProposalApplyInvocationBlock(result, FailureVerificationFailed, "applied_memory_artifact_refs_mismatch", "host:applied_memory_artifact_refs", "review_memory_apply_adapter_invocation_report")
		}
		if !displaySafeRefsContainAll(result.ReadbackMemoryArtifactRefs, readiness.ExpectedMemoryArtifactRefs) {
			result = hostOwnedMemoryProposalApplyInvocationBlock(result, FailureVerificationFailed, "readback_memory_artifact_refs_mismatch", "host:memory_apply_readback_artifact_refs", "review_memory_apply_adapter_invocation_report")
		}
	}
	if len(result.MissingInputs) == 0 && len(result.BlockedReasons) == 0 {
		result.Status = HostActionRecorded
		if result.HostAdapterInvocationFailed {
			result.ReadyForFailureReview = true
			result.FailureClass = firstFailureClass(result.FailureClass, FailureVerificationFailed)
			result.NextHostAction = "review_memory_apply_adapter_failure"
			result.Boundaries = AppendBoundaries(result.Boundaries, "host_owned_memory_apply_adapter_failure_recorded")
		} else {
			result.ReadyForMemoryApplyReadback = true
			result.NextHostAction = "continue_after_memory_apply_readback"
			result.Boundaries = AppendBoundaries(result.Boundaries, "host_owned_memory_apply_adapter_invocation_recorded", "memory_apply_readback_bound")
		}
	}
	return result.Normalize()
}

func CloneMemoryProposalReviewPacket(in MemoryProposalReviewPacket) MemoryProposalReviewPacket {
	out := in
	out.ProposalSet = in.ProposalSet.Clone()
	out.SelectedProposalRefs = cloneDisplaySafeRefs(in.SelectedProposalRefs)
	out.RejectedProposalRefs = cloneDisplaySafeRefs(in.RejectedProposalRefs)
	out.ExpectedMemoryArtifactRefs = cloneDisplaySafeRefs(in.ExpectedMemoryArtifactRefs)
	out.ApprovalRefs = cloneDisplaySafeRefs(in.ApprovalRefs)
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.DecisionBasis = cloneDisplaySafeRefs(in.DecisionBasis)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (p MemoryProposalReviewPacket) Clone() MemoryProposalReviewPacket {
	return CloneMemoryProposalReviewPacket(p)
}

func (p MemoryProposalReviewPacket) Normalize() MemoryProposalReviewPacket {
	out := CloneMemoryProposalReviewPacket(p)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.ProposalSet = out.ProposalSet.Normalize()
	out.RawOutputLoaded = out.RawOutputLoaded || out.ProposalSet.RawOutputLoaded
	out.ReviewPacketRef = normalizeOneDisplaySafeRef(out.ReviewPacketRef)
	out.ReviewerRef = normalizeOneDisplaySafeRef(out.ReviewerRef)
	out.HostReviewRef = normalizeOneDisplaySafeRef(out.HostReviewRef)
	out.MemoryApplyPolicyRef = normalizeOneDisplaySafeRef(out.MemoryApplyPolicyRef)
	out.MemoryApplyAdapterRef = normalizeOneDisplaySafeRef(out.MemoryApplyAdapterRef)
	out.SelectedProposalRefs = normalizeDisplaySafeRefs(out.SelectedProposalRefs)
	out.RejectedProposalRefs = normalizeDisplaySafeRefs(out.RejectedProposalRefs)
	out.ExpectedMemoryArtifactRefs = normalizeDisplaySafeRefs(out.ExpectedMemoryArtifactRefs)
	out.ExpectedMemoryApplyResultRef = normalizeOneDisplaySafeRef(out.ExpectedMemoryApplyResultRef)
	out.ExpectedReadbackRef = normalizeOneDisplaySafeRef(out.ExpectedReadbackRef)
	out.IdempotencyRef = normalizeOneDisplaySafeRef(out.IdempotencyRef)
	out.RollbackPathRef = normalizeOneDisplaySafeRef(out.RollbackPathRef)
	out.ApprovalRefs = normalizeDisplaySafeRefs(out.ApprovalRefs)
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.BlockedReasons = normalizeControlTokenList(out.BlockedReasons)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.DecisionBasis = normalizeDisplaySafeRefs(out.DecisionBasis)
	out.Boundaries = memoryProposalReviewPacketBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	out.RuntimeEffect = normalizeControlToken(out.RuntimeEffect)
	out = normalizeMemoryProposalApplyReviewEffects(out)
	out.ReadyForHostMemoryApply = out.Status == HostActionReady && len(out.SelectedProposalRefs) > 0 && len(out.ExpectedMemoryArtifactRefs) > 0 && len(out.MissingInputs) == 0 && len(out.BlockedReasons) == 0 && !out.RawOutputLoaded
	out.HostMemoryApplyAuthorized = out.ReadyForHostMemoryApply
	out.HostMayApplyMemoryMutation = out.ReadyForHostMemoryApply
	if out.RawOutputLoaded {
		out.Status = HostActionReviewRequired
		out.ReadyForHostMemoryApply = false
		out.HostMemoryApplyAuthorized = false
		out.HostMayApplyMemoryMutation = false
		if out.FailureClass == FailureNone {
			out.FailureClass = FailureEvidenceWeak
		}
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		out.NextHostAction = "provide_display_safe_refs"
	}
	return out
}

func CloneHostOwnedMemoryProposalApplyReadiness(in HostOwnedMemoryProposalApplyReadiness) HostOwnedMemoryProposalApplyReadiness {
	out := in
	out.ReviewPacket = in.ReviewPacket.Clone()
	out.MemoryApplyGate = in.MemoryApplyGate.Clone()
	out.SelectedProposalRefs = cloneDisplaySafeRefs(in.SelectedProposalRefs)
	out.ExpectedMemoryArtifactRefs = cloneDisplaySafeRefs(in.ExpectedMemoryArtifactRefs)
	out.ApprovalRefs = cloneDisplaySafeRefs(in.ApprovalRefs)
	out.PolicyRefs = cloneDisplaySafeRefs(in.PolicyRefs)
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.DecisionBasis = cloneDisplaySafeRefs(in.DecisionBasis)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (r HostOwnedMemoryProposalApplyReadiness) Clone() HostOwnedMemoryProposalApplyReadiness {
	return CloneHostOwnedMemoryProposalApplyReadiness(r)
}

func (r HostOwnedMemoryProposalApplyReadiness) Normalize() HostOwnedMemoryProposalApplyReadiness {
	out := CloneHostOwnedMemoryProposalApplyReadiness(r)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.ReviewPacket = out.ReviewPacket.Normalize()
	out.MemoryApplyGate = out.MemoryApplyGate.Normalize()
	out.RawOutputLoaded = out.RawOutputLoaded || out.ReviewPacket.RawOutputLoaded || out.MemoryApplyGate.RawOutputLoaded
	out.ReviewPacketRef = normalizeOneDisplaySafeRef(out.ReviewPacketRef)
	out.MemoryApplyPolicyRef = normalizeOneDisplaySafeRef(out.MemoryApplyPolicyRef)
	out.MemoryApplyAdapterRef = normalizeOneDisplaySafeRef(out.MemoryApplyAdapterRef)
	out.SelectedProposalRefs = normalizeDisplaySafeRefs(out.SelectedProposalRefs)
	out.ExpectedMemoryArtifactRefs = normalizeDisplaySafeRefs(out.ExpectedMemoryArtifactRefs)
	out.ExpectedMemoryApplyResultRef = normalizeOneDisplaySafeRef(out.ExpectedMemoryApplyResultRef)
	out.ExpectedReadbackRef = normalizeOneDisplaySafeRef(out.ExpectedReadbackRef)
	out.RollbackPathRef = normalizeOneDisplaySafeRef(out.RollbackPathRef)
	out.AdapterRef = normalizeOneDisplaySafeRef(out.AdapterRef)
	out.AdapterVersionRef = normalizeOneDisplaySafeRef(out.AdapterVersionRef)
	out.AdapterCapabilityRef = normalizeOneDisplaySafeRef(out.AdapterCapabilityRef)
	out.AdapterContractRef = normalizeOneDisplaySafeRef(out.AdapterContractRef)
	out.HostConfirmationRef = normalizeOneDisplaySafeRef(out.HostConfirmationRef)
	out.AdapterDryRunRef = normalizeOneDisplaySafeRef(out.AdapterDryRunRef)
	out.InvocationRef = normalizeOneDisplaySafeRef(out.InvocationRef)
	out.ResultBindingRef = normalizeOneDisplaySafeRef(out.ResultBindingRef)
	out.ReadbackBindingRef = normalizeOneDisplaySafeRef(out.ReadbackBindingRef)
	out.IdempotencyRef = normalizeOneDisplaySafeRef(out.IdempotencyRef)
	out.RollbackRef = normalizeOneDisplaySafeRef(out.RollbackRef)
	out.FailureBindingRef = normalizeOneDisplaySafeRef(out.FailureBindingRef)
	out.CompensationRef = normalizeOneDisplaySafeRef(out.CompensationRef)
	out.ApprovalRefs = normalizeDisplaySafeRefs(out.ApprovalRefs)
	out.PolicyRefs = normalizeDisplaySafeRefs(out.PolicyRefs)
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.BlockedReasons = normalizeControlTokenList(out.BlockedReasons)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.DecisionBasis = normalizeDisplaySafeRefs(out.DecisionBasis)
	out.Boundaries = hostOwnedMemoryProposalApplyReadinessBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	out.RuntimeEffect = normalizeControlToken(out.RuntimeEffect)
	out = normalizeMemoryProposalApplyReadinessEffects(out)
	out.ReadyForHostMemoryAdapterInvocation = out.Status == HostActionReady && len(out.MissingInputs) == 0 && len(out.BlockedReasons) == 0 && !out.RawOutputLoaded
	out.HostMemoryAdapterInvocationAuthorized = out.ReadyForHostMemoryAdapterInvocation
	out.HostMayInvokeMemoryAdapter = out.ReadyForHostMemoryAdapterInvocation
	if out.RawOutputLoaded {
		out.Status = HostActionReviewRequired
		out.ReadyForHostMemoryAdapterInvocation = false
		out.HostMemoryAdapterInvocationAuthorized = false
		out.HostMayInvokeMemoryAdapter = false
		if out.FailureClass == FailureNone {
			out.FailureClass = FailureEvidenceWeak
		}
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		out.NextHostAction = "provide_display_safe_refs"
	}
	return out
}

func CloneHostOwnedMemoryProposalApplyInvocation(in HostOwnedMemoryProposalApplyInvocation) HostOwnedMemoryProposalApplyInvocation {
	out := in
	out.Readiness = in.Readiness.Clone()
	out.SelectedProposalRefs = cloneDisplaySafeRefs(in.SelectedProposalRefs)
	out.ExpectedMemoryArtifactRefs = cloneDisplaySafeRefs(in.ExpectedMemoryArtifactRefs)
	out.AppliedMemoryArtifactRefs = cloneDisplaySafeRefs(in.AppliedMemoryArtifactRefs)
	out.ReadbackMemoryArtifactRefs = cloneDisplaySafeRefs(in.ReadbackMemoryArtifactRefs)
	out.MemoryEvidenceRefs = cloneDisplaySafeRefs(in.MemoryEvidenceRefs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (r HostOwnedMemoryProposalApplyInvocation) Clone() HostOwnedMemoryProposalApplyInvocation {
	return CloneHostOwnedMemoryProposalApplyInvocation(r)
}

func (r HostOwnedMemoryProposalApplyInvocation) Normalize() HostOwnedMemoryProposalApplyInvocation {
	out := CloneHostOwnedMemoryProposalApplyInvocation(r)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.Readiness = out.Readiness.Normalize()
	out.RawOutputLoaded = out.RawOutputLoaded || out.Readiness.RawOutputLoaded
	out.ReviewPacketRef = normalizeOneDisplaySafeRef(out.ReviewPacketRef)
	out.MemoryApplyAdapterRef = normalizeOneDisplaySafeRef(out.MemoryApplyAdapterRef)
	out.SelectedProposalRefs = normalizeDisplaySafeRefs(out.SelectedProposalRefs)
	out.ExpectedMemoryArtifactRefs = normalizeDisplaySafeRefs(out.ExpectedMemoryArtifactRefs)
	out.ExpectedMemoryApplyResultRef = normalizeOneDisplaySafeRef(out.ExpectedMemoryApplyResultRef)
	out.ExpectedReadbackRef = normalizeOneDisplaySafeRef(out.ExpectedReadbackRef)
	out.RollbackPathRef = normalizeOneDisplaySafeRef(out.RollbackPathRef)
	out.AdapterRef = normalizeOneDisplaySafeRef(out.AdapterRef)
	out.AdapterVersionRef = normalizeOneDisplaySafeRef(out.AdapterVersionRef)
	out.InvocationRef = normalizeOneDisplaySafeRef(out.InvocationRef)
	out.InvocationReportRef = normalizeOneDisplaySafeRef(out.InvocationReportRef)
	out.ObservedInvocationRef = normalizeOneDisplaySafeRef(out.ObservedInvocationRef)
	out.HostMemoryAdapterRunRef = normalizeOneDisplaySafeRef(out.HostMemoryAdapterRunRef)
	out.MemoryApplyResultRef = normalizeOneDisplaySafeRef(out.MemoryApplyResultRef)
	out.MemoryReadbackRef = normalizeOneDisplaySafeRef(out.MemoryReadbackRef)
	out.AppliedMemoryArtifactRefs = normalizeDisplaySafeRefs(out.AppliedMemoryArtifactRefs)
	out.ReadbackMemoryArtifactRefs = normalizeDisplaySafeRefs(out.ReadbackMemoryArtifactRefs)
	out.FailureRef = normalizeOneDisplaySafeRef(out.FailureRef)
	out.CompensationRef = normalizeOneDisplaySafeRef(out.CompensationRef)
	out.MemoryEvidenceRefs = normalizeDisplaySafeRefs(out.MemoryEvidenceRefs)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.BlockedReasons = normalizeControlTokenList(out.BlockedReasons)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.Boundaries = hostOwnedMemoryProposalApplyInvocationBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	out.RuntimeEffect = normalizeControlToken(out.RuntimeEffect)
	out = normalizeMemoryProposalApplyInvocationEffects(out)
	out.ReadyForMemoryApplyReadback = out.Status == HostActionRecorded && out.HostAdapterInvocationCompleted && !out.HostAdapterInvocationFailed && len(out.MissingInputs) == 0 && len(out.BlockedReasons) == 0 && !out.RawOutputLoaded
	out.ReadyForFailureReview = out.Status == HostActionRecorded && out.HostAdapterInvocationFailed && len(out.MissingInputs) == 0 && len(out.BlockedReasons) == 0 && !out.RawOutputLoaded
	if out.RawOutputLoaded {
		out.Status = HostActionReviewRequired
		out.ReadyForMemoryApplyReadback = false
		out.ReadyForFailureReview = false
		if out.FailureClass == FailureNone {
			out.FailureClass = FailureEvidenceWeak
		}
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		out.NextHostAction = "provide_display_safe_refs"
	}
	return out
}

func memoryProposalReviewPacketBlock(result MemoryProposalReviewPacket, failure FailureClass, reason string, missing MissingInput, next NextHostAction) MemoryProposalReviewPacket {
	result.Status = HostActionBlocked
	result.ReadyForHostMemoryApply = false
	result.HostMemoryApplyAuthorized = false
	result.HostMayApplyMemoryMutation = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	return result
}

func hostOwnedMemoryProposalApplyReadinessBlock(result HostOwnedMemoryProposalApplyReadiness, failure FailureClass, reason string, missing MissingInput, next NextHostAction) HostOwnedMemoryProposalApplyReadiness {
	result.Status = HostActionBlocked
	result.ReadyForHostMemoryAdapterInvocation = false
	result.HostMemoryAdapterInvocationAuthorized = false
	result.HostMayInvokeMemoryAdapter = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	return result
}

func hostOwnedMemoryProposalApplyInvocationBlock(result HostOwnedMemoryProposalApplyInvocation, failure FailureClass, reason string, missing MissingInput, next NextHostAction) HostOwnedMemoryProposalApplyInvocation {
	result.Status = HostActionBlocked
	result.ReadyForMemoryApplyReadback = false
	result.ReadyForFailureReview = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	return result
}

func memoryProposalReviewPacketUnsafe(input MemoryProposalReviewPacketInput, proposalSet RepeatedSuccessMemoryProposalSet) bool {
	return input.RawOutputLoaded ||
		proposalSet.RawOutputLoaded ||
		displaySafeRefRejected(input.ReviewPacketRef) ||
		displaySafeRefRejected(input.ReviewerRef) ||
		displaySafeRefRejected(input.HostReviewRef) ||
		displaySafeRefRejected(input.MemoryApplyPolicyRef) ||
		displaySafeRefRejected(input.MemoryApplyAdapterRef) ||
		displaySafeRefRejected(input.ExpectedMemoryApplyResultRef) ||
		displaySafeRefRejected(input.ExpectedReadbackRef) ||
		displaySafeRefRejected(input.IdempotencyRef) ||
		displaySafeRefRejected(input.RollbackPathRef) ||
		displaySafeRefSliceRejected(input.SelectedProposalRefs) ||
		displaySafeRefSliceRejected(input.RejectedProposalRefs) ||
		displaySafeRefSliceRejected(input.ExpectedMemoryArtifactRefs) ||
		displaySafeRefSliceRejected(input.ApprovalRefs) ||
		displaySafeRefSliceRejected(input.DecisionBasis) ||
		evidenceRefRejected(input.EvidenceRefs) ||
		repeatedSuccessMemoryProposalSetOutputUnsafe(proposalSet)
}

func hostOwnedMemoryProposalApplyReadinessUnsafe(input HostOwnedMemoryProposalApplyReadinessInput, review MemoryProposalReviewPacket, gate ProductionAdapterIndependentEffectGate) bool {
	return input.RawOutputLoaded ||
		review.RawOutputLoaded ||
		gate.RawOutputLoaded ||
		displaySafeRefRejected(input.AdapterRef) ||
		displaySafeRefRejected(input.AdapterVersionRef) ||
		displaySafeRefRejected(input.AdapterCapabilityRef) ||
		displaySafeRefRejected(input.AdapterContractRef) ||
		displaySafeRefRejected(input.HostConfirmationRef) ||
		displaySafeRefRejected(input.AdapterDryRunRef) ||
		displaySafeRefRejected(input.InvocationRef) ||
		displaySafeRefRejected(input.ResultBindingRef) ||
		displaySafeRefRejected(input.ReadbackBindingRef) ||
		displaySafeRefRejected(input.IdempotencyRef) ||
		displaySafeRefRejected(input.RollbackRef) ||
		displaySafeRefRejected(input.FailureBindingRef) ||
		displaySafeRefRejected(input.CompensationRef) ||
		displaySafeRefSliceRejected(input.ApprovalRefs) ||
		displaySafeRefSliceRejected(input.PolicyRefs) ||
		displaySafeRefSliceRejected(input.DecisionBasis) ||
		evidenceRefRejected(input.EvidenceRefs) ||
		memoryProposalReviewPacketOutputUnsafe(review) ||
		productionAdapterIndependentEffectGateOutputUnsafe(gate)
}

func hostOwnedMemoryProposalApplyInvocationUnsafe(input HostOwnedMemoryProposalApplyInvocationInput, readiness HostOwnedMemoryProposalApplyReadiness) bool {
	return input.RawOutputLoaded ||
		readiness.RawOutputLoaded ||
		displaySafeRefRejected(input.InvocationReportRef) ||
		displaySafeRefRejected(input.ObservedInvocationRef) ||
		displaySafeRefRejected(input.HostMemoryAdapterRunRef) ||
		displaySafeRefRejected(input.MemoryApplyResultRef) ||
		displaySafeRefRejected(input.MemoryReadbackRef) ||
		displaySafeRefRejected(input.FailureRef) ||
		displaySafeRefRejected(input.CompensationRef) ||
		displaySafeRefSliceRejected(input.AppliedMemoryArtifactRefs) ||
		displaySafeRefSliceRejected(input.ReadbackMemoryArtifactRefs) ||
		displaySafeRefSliceRejected(input.MemoryEvidenceRefs) ||
		hostOwnedMemoryProposalApplyReadinessOutputUnsafe(readiness)
}

func memoryProposalReviewPacketBoundaries(groups ...[]Boundary) []Boundary {
	all := append([][]Boundary{{
		"memory_proposal_review_packet",
		"memory_proposal_review_required",
		"memory_proposal_apply_request",
		"proposal_not_apply",
		"host_owned_memory_apply",
		"explicit_host_review_required",
		"explicit_host_approval_required",
		"display_safe_refs_only",
		"no_skill_write_by_core",
		"no_workflow_write_by_core",
		"no_template_write_by_core",
		"no_install_apply_by_core",
		"no_runtime_reload_by_core",
		"no_core_mutation",
		"no_runner_dispatch",
		"no_runtime_adapter_execution",
		"no_workflow_dispatch",
		"no_worker_dispatch",
		"no_store_mutation_by_core",
		"no_compensation_execution",
		"projection_only",
	}}, groups...)
	return MergeBoundaries(all...)
}

func hostOwnedMemoryProposalApplyReadinessBoundaries(groups ...[]Boundary) []Boundary {
	all := append([][]Boundary{{
		"host_owned_memory_apply_adapter_gate",
		"host_owned_memory_apply_adapter_readiness",
		"memory_apply_adapter_invocation_gate",
		"memory_apply_independent_gate_required",
		"explicit_host_confirmation_required",
		"memory_apply_adapter_dry_run_required",
		"display_safe_refs_only",
		"no_skill_write_by_core",
		"no_workflow_write_by_core",
		"no_template_write_by_core",
		"no_install_apply_by_core",
		"no_runtime_reload_by_core",
		"no_core_mutation",
		"no_runner_dispatch",
		"no_runtime_adapter_execution",
		"no_workflow_dispatch",
		"no_worker_dispatch",
		"no_store_mutation_by_core",
		"no_compensation_execution",
		"projection_only",
	}}, groups...)
	return MergeBoundaries(all...)
}

func hostOwnedMemoryProposalApplyInvocationBoundaries(groups ...[]Boundary) []Boundary {
	all := append([][]Boundary{{
		"host_owned_memory_apply_adapter_gate",
		"host_owned_memory_apply_adapter_invocation_report",
		"host_memory_adapter_invocation_report_only",
		"host_adapter_memory_mutation_reported_only",
		"memory_apply_result_requires_readback",
		"proposal_not_apply",
		"display_safe_refs_only",
		"no_skill_write_by_core",
		"no_workflow_write_by_core",
		"no_template_write_by_core",
		"no_install_apply_by_core",
		"no_runtime_reload_by_core",
		"no_core_mutation",
		"no_runner_dispatch",
		"no_runtime_adapter_execution",
		"no_workflow_dispatch",
		"no_worker_dispatch",
		"no_store_mutation_by_core",
		"no_compensation_execution",
		"projection_only",
	}}, groups...)
	return MergeBoundaries(all...)
}

func normalizeMemoryProposalApplyReviewEffects(value MemoryProposalReviewPacket) MemoryProposalReviewPacket {
	if value.RunnerEffect == "" {
		value.RunnerEffect = "none"
	}
	if value.PromptEffect == "" {
		value.PromptEffect = "none"
	}
	if value.RuntimeEffect == "" {
		value.RuntimeEffect = "none"
	}
	value.ProposalOnly = true
	value.SkillWriteByCore = false
	value.WorkflowWriteByCore = false
	value.TemplateWriteByCore = false
	value.InstallExecutedByCore = false
	value.RuntimeReloadByCore = false
	value.CoreMutationExecuted = false
	value.RunnerDispatched = false
	value.RuntimeAdapterExecuted = false
	value.WorkflowDispatched = false
	value.WorkerDispatched = false
	value.StoreMutationExecuted = false
	value.CompensationExecuted = false
	return value
}

func normalizeMemoryProposalApplyReadinessEffects(value HostOwnedMemoryProposalApplyReadiness) HostOwnedMemoryProposalApplyReadiness {
	if value.RunnerEffect == "" {
		value.RunnerEffect = "none"
	}
	if value.PromptEffect == "" {
		value.PromptEffect = "none"
	}
	if value.RuntimeEffect == "" {
		value.RuntimeEffect = "none"
	}
	value.SkillWriteByCore = false
	value.WorkflowWriteByCore = false
	value.TemplateWriteByCore = false
	value.InstallExecutedByCore = false
	value.RuntimeReloadByCore = false
	value.CoreMutationExecuted = false
	value.RunnerDispatched = false
	value.RuntimeAdapterExecuted = false
	value.WorkflowDispatched = false
	value.WorkerDispatched = false
	value.StoreMutationExecuted = false
	value.CompensationExecuted = false
	return value
}

func normalizeMemoryProposalApplyInvocationEffects(value HostOwnedMemoryProposalApplyInvocation) HostOwnedMemoryProposalApplyInvocation {
	if value.RunnerEffect == "" {
		value.RunnerEffect = "none"
	}
	if value.PromptEffect == "" {
		value.PromptEffect = "none"
	}
	if value.RuntimeEffect == "" {
		value.RuntimeEffect = "none"
	}
	value.SkillWriteByCore = false
	value.WorkflowWriteByCore = false
	value.TemplateWriteByCore = false
	value.InstallExecutedByCore = false
	value.RuntimeReloadByCore = false
	value.CoreMutationExecuted = false
	value.RunnerDispatched = false
	value.RuntimeAdapterExecuted = false
	value.WorkflowDispatched = false
	value.WorkerDispatched = false
	value.StoreMutationExecuted = false
	value.CompensationExecuted = false
	return value
}

func memoryProposalReviewPacketOutputUnsafe(input MemoryProposalReviewPacket) bool {
	return displaySafeRefRejected(input.ReviewPacketRef) ||
		displaySafeRefRejected(input.ReviewerRef) ||
		displaySafeRefRejected(input.HostReviewRef) ||
		displaySafeRefRejected(input.MemoryApplyPolicyRef) ||
		displaySafeRefRejected(input.MemoryApplyAdapterRef) ||
		displaySafeRefRejected(input.ExpectedMemoryApplyResultRef) ||
		displaySafeRefRejected(input.ExpectedReadbackRef) ||
		displaySafeRefRejected(input.IdempotencyRef) ||
		displaySafeRefRejected(input.RollbackPathRef) ||
		displaySafeRefSliceRejected(input.SelectedProposalRefs) ||
		displaySafeRefSliceRejected(input.RejectedProposalRefs) ||
		displaySafeRefSliceRejected(input.ExpectedMemoryArtifactRefs) ||
		displaySafeRefSliceRejected(input.ApprovalRefs) ||
		displaySafeRefSliceRejected(input.DecisionBasis) ||
		evidenceRefRejected(input.EvidenceRefs) ||
		repeatedSuccessMemoryProposalSetOutputUnsafe(input.ProposalSet)
}

func hostOwnedMemoryProposalApplyReadinessOutputUnsafe(input HostOwnedMemoryProposalApplyReadiness) bool {
	return displaySafeRefRejected(input.ReviewPacketRef) ||
		displaySafeRefRejected(input.MemoryApplyPolicyRef) ||
		displaySafeRefRejected(input.MemoryApplyAdapterRef) ||
		displaySafeRefSliceRejected(input.SelectedProposalRefs) ||
		displaySafeRefSliceRejected(input.ExpectedMemoryArtifactRefs) ||
		displaySafeRefRejected(input.ExpectedMemoryApplyResultRef) ||
		displaySafeRefRejected(input.ExpectedReadbackRef) ||
		displaySafeRefRejected(input.RollbackPathRef) ||
		displaySafeRefRejected(input.AdapterRef) ||
		displaySafeRefRejected(input.AdapterVersionRef) ||
		displaySafeRefRejected(input.AdapterCapabilityRef) ||
		displaySafeRefRejected(input.AdapterContractRef) ||
		displaySafeRefRejected(input.HostConfirmationRef) ||
		displaySafeRefRejected(input.AdapterDryRunRef) ||
		displaySafeRefRejected(input.InvocationRef) ||
		displaySafeRefRejected(input.ResultBindingRef) ||
		displaySafeRefRejected(input.ReadbackBindingRef) ||
		displaySafeRefRejected(input.IdempotencyRef) ||
		displaySafeRefRejected(input.RollbackRef) ||
		displaySafeRefRejected(input.FailureBindingRef) ||
		displaySafeRefRejected(input.CompensationRef) ||
		displaySafeRefSliceRejected(input.ApprovalRefs) ||
		displaySafeRefSliceRejected(input.PolicyRefs) ||
		displaySafeRefSliceRejected(input.DecisionBasis) ||
		evidenceRefRejected(input.EvidenceRefs) ||
		memoryProposalReviewPacketOutputUnsafe(input.ReviewPacket) ||
		productionAdapterIndependentEffectGateOutputUnsafe(input.MemoryApplyGate)
}

func repeatedSuccessMemoryProposalSetOutputUnsafe(input RepeatedSuccessMemoryProposalSet) bool {
	if displaySafeRefRejected(input.ProposalSetRef) ||
		displaySafeRefRejected(input.ProposalOwnerRef) ||
		displaySafeRefRejected(input.MemoryPolicyRef) ||
		displaySafeRefRejected(input.StrategyCatalogRef) ||
		displaySafeRefSliceRejected(input.RepeatedStrategyRefs) ||
		displaySafeRefSliceRejected(input.ProvenanceRefs) ||
		displaySafeRefSliceRejected(input.DecisionBasis) ||
		evidenceRefRejected(input.EvidenceRefs) {
		return true
	}
	for _, proposal := range input.Proposals {
		if memoryProposalOutputUnsafe(proposal) {
			return true
		}
	}
	return false
}

func memoryProposalOutputUnsafe(input MemoryProposal) bool {
	return displaySafeRefRejected(input.ProposalRef) ||
		displaySafeRefRejected(input.SourceStrategyRef) ||
		displaySafeRefRejected(input.SourceRef) ||
		displaySafeRefSliceRejected(input.ProvenanceRefs) ||
		evidenceRefRejected(input.EvidenceRefs)
}

func memoryProposalRefsInSet(selected []DisplaySafeRef, proposals []MemoryProposal) bool {
	allowed := map[DisplaySafeRef]struct{}{}
	for _, proposal := range normalizeMemoryProposals(proposals) {
		if proposal.ProposalRef != "" {
			allowed[proposal.ProposalRef] = struct{}{}
		}
	}
	for _, ref := range normalizeDisplaySafeRefs(selected) {
		if _, ok := allowed[ref]; !ok {
			return false
		}
	}
	return len(selected) > 0
}
