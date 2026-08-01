package controlcontract

type ProductionAdapterObjectiveCloseoutWriterRunnerGateInput struct {
	RunnerGateRef         DisplaySafeRef                                          `json:"runner_gate_ref,omitempty"`
	ExecutionBridge       ProductionAdapterObjectiveCloseoutWriterExecutionBridge `json:"execution_bridge,omitempty"`
	HostRunnerRef         DisplaySafeRef                                          `json:"host_runner_ref,omitempty"`
	HostRunnerVersionRef  DisplaySafeRef                                          `json:"host_runner_version_ref,omitempty"`
	RunnerInvocationRef   DisplaySafeRef                                          `json:"runner_invocation_ref,omitempty"`
	HostConfirmationRef   DisplaySafeRef                                          `json:"host_confirmation_ref,omitempty"`
	PolicyRefs            []DisplaySafeRef                                        `json:"policy_refs,omitempty"`
	ApprovalBindingRefs   []DisplaySafeRef                                        `json:"approval_binding_refs,omitempty"`
	BudgetRef             DisplaySafeRef                                          `json:"budget_ref,omitempty"`
	IdempotencyRef        DisplaySafeRef                                          `json:"idempotency_ref,omitempty"`
	TimeoutPolicyRef      DisplaySafeRef                                          `json:"timeout_policy_ref,omitempty"`
	CancellationPolicyRef DisplaySafeRef                                          `json:"cancellation_policy_ref,omitempty"`
	RawOutputLoaded       bool                                                    `json:"raw_output_loaded"`
}

type ProductionAdapterObjectiveCloseoutWriterRunnerGate struct {
	ContractVersion                 string           `json:"contract_version,omitempty"`
	Projected                       bool             `json:"projected"`
	Available                       bool             `json:"available"`
	Status                          string           `json:"status,omitempty"`
	Mode                            string           `json:"mode,omitempty"`
	ReadyForHostDisplay             bool             `json:"ready_for_host_display"`
	ReadyForHostRunner              bool             `json:"ready_for_host_runner"`
	ReadyForHostAdapterExecution    bool             `json:"ready_for_host_adapter_execution"`
	HostRunnerAuthorized            bool             `json:"host_runner_authorized"`
	HostConfirmationBound           bool             `json:"host_confirmation_bound"`
	DryRunFirstSatisfied            bool             `json:"dry_run_first_satisfied"`
	PolicyBound                     bool             `json:"policy_bound"`
	ApprovalBound                   bool             `json:"approval_bound"`
	BudgetBound                     bool             `json:"budget_bound"`
	IdempotencyBound                bool             `json:"idempotency_bound"`
	TimeoutBound                    bool             `json:"timeout_bound"`
	HostMayInvokeWriterAdapter      bool             `json:"host_may_invoke_writer_adapter"`
	HostMayExecuteDurableWrite      bool             `json:"host_may_execute_durable_write"`
	CoreInvocationExecuted          bool             `json:"core_invocation_executed"`
	DryRunByCore                    bool             `json:"dry_run_by_core"`
	DurableWriteByCore              bool             `json:"durable_write_by_core"`
	ObjectiveStoreWriteByCore       bool             `json:"objective_store_write_by_core"`
	RunstoreWriteByCore             bool             `json:"runstore_write_by_core"`
	RunnerGateRef                   DisplaySafeRef   `json:"runner_gate_ref,omitempty"`
	BridgeRef                       DisplaySafeRef   `json:"bridge_ref,omitempty"`
	HostRunnerRef                   DisplaySafeRef   `json:"host_runner_ref,omitempty"`
	HostRunnerVersionRef            DisplaySafeRef   `json:"host_runner_version_ref,omitempty"`
	RunnerInvocationRef             DisplaySafeRef   `json:"runner_invocation_ref,omitempty"`
	HostConfirmationRef             DisplaySafeRef   `json:"host_confirmation_ref,omitempty"`
	HostUIHandoffRef                DisplaySafeRef   `json:"host_ui_handoff_ref,omitempty"`
	PrimaryDisplayRef               DisplaySafeRef   `json:"primary_display_ref,omitempty"`
	ReviewRef                       DisplaySafeRef   `json:"review_ref,omitempty"`
	FixtureRef                      DisplaySafeRef   `json:"fixture_ref,omitempty"`
	InvocationEnvelopeRef           DisplaySafeRef   `json:"invocation_envelope_ref,omitempty"`
	ResultEnvelopeRef               DisplaySafeRef   `json:"result_envelope_ref,omitempty"`
	ReviewPacketRef                 DisplaySafeRef   `json:"review_packet_ref,omitempty"`
	DurableRequestRef               DisplaySafeRef   `json:"durable_request_ref,omitempty"`
	ExpectedDurableResultRef        DisplaySafeRef   `json:"expected_durable_result_ref,omitempty"`
	WriterInvocationRef             DisplaySafeRef   `json:"writer_invocation_ref,omitempty"`
	WriterOptInRef                  DisplaySafeRef   `json:"writer_opt_in_ref,omitempty"`
	WriterRef                       DisplaySafeRef   `json:"writer_ref,omitempty"`
	HostWriterBindingRef            DisplaySafeRef   `json:"host_writer_binding_ref,omitempty"`
	HostAdapterVersionRef           DisplaySafeRef   `json:"host_adapter_version_ref,omitempty"`
	ExpectedHostAdapterRunRef       DisplaySafeRef   `json:"expected_host_adapter_run_ref,omitempty"`
	DryRunSmokeRef                  DisplaySafeRef   `json:"dry_run_smoke_ref,omitempty"`
	DryRunResultRef                 DisplaySafeRef   `json:"dry_run_result_ref,omitempty"`
	ExpectedReadbackRef             DisplaySafeRef   `json:"expected_readback_ref,omitempty"`
	ExpectedFailureRef              DisplaySafeRef   `json:"expected_failure_ref,omitempty"`
	ExpectedCompensationRef         DisplaySafeRef   `json:"expected_compensation_ref,omitempty"`
	ObjectiveCloseoutHandoffRef     DisplaySafeRef   `json:"objective_closeout_handoff_ref,omitempty"`
	ObjectiveCloseoutPacketRef      DisplaySafeRef   `json:"objective_closeout_packet_ref,omitempty"`
	ObjectiveRef                    DisplaySafeRef   `json:"objective_ref,omitempty"`
	HostRunstoreRef                 DisplaySafeRef   `json:"host_runstore_ref,omitempty"`
	ExpectedDurableEventRef         DisplaySafeRef   `json:"expected_durable_event_ref,omitempty"`
	ExpectedObjectiveStateRef       DisplaySafeRef   `json:"expected_objective_state_ref,omitempty"`
	HostDurableWriteConfirmationRef DisplaySafeRef   `json:"host_durable_write_confirmation_ref,omitempty"`
	CapabilityProofRefs             []DisplaySafeRef `json:"capability_proof_refs,omitempty"`
	PolicyRefs                      []DisplaySafeRef `json:"policy_refs,omitempty"`
	RequiredPolicyRefs              []DisplaySafeRef `json:"required_policy_refs,omitempty"`
	ApprovalBindingRefs             []DisplaySafeRef `json:"approval_binding_refs,omitempty"`
	RequiredApprovalBindingRefs     []DisplaySafeRef `json:"required_approval_binding_refs,omitempty"`
	ApprovalRefs                    []DisplaySafeRef `json:"approval_refs,omitempty"`
	RequiredApprovalRefs            []DisplaySafeRef `json:"required_approval_refs,omitempty"`
	BudgetRef                       DisplaySafeRef   `json:"budget_ref,omitempty"`
	RequiredBudgetRef               DisplaySafeRef   `json:"required_budget_ref,omitempty"`
	IdempotencyRef                  DisplaySafeRef   `json:"idempotency_ref,omitempty"`
	IdempotencyContractRef          DisplaySafeRef   `json:"idempotency_contract_ref,omitempty"`
	TimeoutPolicyRef                DisplaySafeRef   `json:"timeout_policy_ref,omitempty"`
	CancellationPolicyRef           DisplaySafeRef   `json:"cancellation_policy_ref,omitempty"`
	RequiredInputs                  []MissingInput   `json:"required_inputs,omitempty"`
	MissingInputs                   []MissingInput   `json:"missing_inputs,omitempty"`
	BlockedReasons                  []string         `json:"blocked_reasons,omitempty"`
	FailureClass                    FailureClass     `json:"failure_class,omitempty"`
	Boundaries                      []Boundary       `json:"boundaries,omitempty"`
	NextHostAction                  NextHostAction   `json:"next_host_action,omitempty"`
	RunnerEffect                    string           `json:"runner_effect,omitempty"`
	PromptEffect                    string           `json:"prompt_effect,omitempty"`
	RawOutputLoaded                 bool             `json:"raw_output_loaded"`
}

func BuildProductionAdapterObjectiveCloseoutWriterRunnerGate(input ProductionAdapterObjectiveCloseoutWriterRunnerGateInput) ProductionAdapterObjectiveCloseoutWriterRunnerGate {
	if productionAdapterObjectiveCloseoutWriterExecutionBridgeEmpty(input.ExecutionBridge) {
		return unavailableProductionAdapterObjectiveCloseoutWriterRunnerGate()
	}
	bridge := input.ExecutionBridge.Normalize()
	result := ProductionAdapterObjectiveCloseoutWriterRunnerGate{
		ContractVersion:                 ContractVersion,
		Projected:                       true,
		Available:                       bridge.Available,
		Status:                          "blocked",
		Mode:                            "production_adapter_objective_closeout_writer_runner_gate",
		RunnerGateRef:                   normalizeOneDisplaySafeRef(input.RunnerGateRef),
		BridgeRef:                       bridge.BridgeRef,
		HostRunnerRef:                   normalizeOneDisplaySafeRef(input.HostRunnerRef),
		HostRunnerVersionRef:            normalizeOneDisplaySafeRef(input.HostRunnerVersionRef),
		RunnerInvocationRef:             normalizeOneDisplaySafeRef(input.RunnerInvocationRef),
		HostConfirmationRef:             normalizeOneDisplaySafeRef(input.HostConfirmationRef),
		HostUIHandoffRef:                bridge.HostUIHandoffRef,
		PrimaryDisplayRef:               bridge.PrimaryDisplayRef,
		ReviewRef:                       bridge.ReviewRef,
		FixtureRef:                      bridge.FixtureRef,
		InvocationEnvelopeRef:           bridge.InvocationEnvelopeRef,
		ResultEnvelopeRef:               bridge.ResultEnvelopeRef,
		ReviewPacketRef:                 bridge.ReviewPacketRef,
		DurableRequestRef:               bridge.DurableRequestRef,
		ExpectedDurableResultRef:        bridge.ExpectedDurableResultRef,
		WriterInvocationRef:             bridge.WriterInvocationRef,
		WriterOptInRef:                  bridge.WriterOptInRef,
		WriterRef:                       bridge.WriterRef,
		HostWriterBindingRef:            bridge.HostWriterBindingRef,
		HostAdapterVersionRef:           bridge.HostAdapterVersionRef,
		ExpectedHostAdapterRunRef:       bridge.ExpectedHostAdapterRunRef,
		DryRunSmokeRef:                  bridge.DryRunSmokeRef,
		DryRunResultRef:                 bridge.DryRunResultRef,
		ExpectedReadbackRef:             bridge.ExpectedReadbackRef,
		ExpectedFailureRef:              bridge.ExpectedFailureRef,
		ExpectedCompensationRef:         bridge.ExpectedCompensationRef,
		ObjectiveCloseoutHandoffRef:     bridge.ObjectiveCloseoutHandoffRef,
		ObjectiveCloseoutPacketRef:      bridge.ObjectiveCloseoutPacketRef,
		ObjectiveRef:                    bridge.ObjectiveRef,
		HostRunstoreRef:                 bridge.HostRunstoreRef,
		ExpectedDurableEventRef:         bridge.ExpectedDurableEventRef,
		ExpectedObjectiveStateRef:       bridge.ExpectedObjectiveStateRef,
		HostDurableWriteConfirmationRef: bridge.HostDurableWriteConfirmationRef,
		CapabilityProofRefs:             cloneDisplaySafeRefs(bridge.CapabilityProofRefs),
		PolicyRefs:                      normalizeDisplaySafeRefs(input.PolicyRefs),
		RequiredPolicyRefs:              cloneDisplaySafeRefs(bridge.RequiredPolicyRefs),
		ApprovalBindingRefs:             normalizeDisplaySafeRefs(input.ApprovalBindingRefs),
		RequiredApprovalBindingRefs:     cloneDisplaySafeRefs(bridge.ApprovalBindingRefs),
		ApprovalRefs:                    cloneDisplaySafeRefs(bridge.ApprovalRefs),
		RequiredApprovalRefs:            cloneDisplaySafeRefs(bridge.RequiredApprovalRefs),
		BudgetRef:                       normalizeOneDisplaySafeRef(input.BudgetRef),
		RequiredBudgetRef:               firstDisplaySafeRef(bridge.RequiredBudgetRef, bridge.BudgetRef),
		IdempotencyRef:                  normalizeOneDisplaySafeRef(input.IdempotencyRef),
		IdempotencyContractRef:          bridge.IdempotencyContractRef,
		TimeoutPolicyRef:                normalizeOneDisplaySafeRef(input.TimeoutPolicyRef),
		CancellationPolicyRef:           normalizeOneDisplaySafeRef(input.CancellationPolicyRef),
		RequiredInputs:                  productionAdapterObjectiveCloseoutWriterRunnerGateRequiredInputs(),
		FailureClass:                    FailureNone,
		Boundaries:                      productionAdapterObjectiveCloseoutWriterRunnerGateBoundaries(bridge.Boundaries),
		RunnerEffect:                    "none",
		PromptEffect:                    "none",
		RawOutputLoaded:                 input.RawOutputLoaded || bridge.RawOutputLoaded,
	}
	if productionAdapterObjectiveCloseoutWriterRunnerGateUnsafe(input, bridge) {
		result = productionAdapterObjectiveCloseoutWriterRunnerGateBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if !bridge.ReadyForHostAdapterExecution || !bridge.HostAdapterExecutionAuthorized || !bridge.HostMayInvokeWriterAdapter || !bridge.HostMayExecuteDurableWrite {
		result = productionAdapterObjectiveCloseoutWriterRunnerGateBlock(result, firstFailureClass(bridge.FailureClass, FailureAuthorizationMissing), "writer_execution_bridge_not_ready", "host:objective_closeout_writer_execution_bridge", firstNextHostAction(bridge.NextHostAction, "review_objective_closeout_writer_execution_bridge"))
	}
	if result.RunnerGateRef == "" {
		result = productionAdapterObjectiveCloseoutWriterRunnerGateBlock(result, FailureEvidenceMissing, "writer_runner_gate_ref_missing", "host:objective_closeout_writer_runner_gate_ref", "provide_objective_closeout_writer_runner_gate_ref")
	}
	if result.HostRunnerRef == "" {
		result = productionAdapterObjectiveCloseoutWriterRunnerGateBlock(result, FailureConfigMissing, "host_runner_ref_missing", "host:objective_closeout_writer_runner_ref", "provide_objective_closeout_writer_runner")
	}
	if result.HostRunnerVersionRef == "" {
		result = productionAdapterObjectiveCloseoutWriterRunnerGateBlock(result, FailureConfigMissing, "host_runner_version_ref_missing", "host:objective_closeout_writer_runner_version_ref", "provide_objective_closeout_writer_runner_version")
	}
	if result.RunnerInvocationRef == "" {
		result = productionAdapterObjectiveCloseoutWriterRunnerGateBlock(result, FailureEvidenceMissing, "runner_invocation_ref_missing", "host:objective_closeout_writer_runner_invocation_ref", "provide_objective_closeout_writer_runner_invocation_ref")
	}
	if result.HostConfirmationRef == "" {
		result = productionAdapterObjectiveCloseoutWriterRunnerGateBlock(result, FailureAuthorizationMissing, "host_confirmation_ref_missing", "host:objective_closeout_writer_host_confirmation_ref", "request_objective_closeout_writer_host_confirmation")
	}
	if !displaySafeRefsContainAll(result.PolicyRefs, result.RequiredPolicyRefs) {
		result = productionAdapterObjectiveCloseoutWriterRunnerGateBlock(result, FailurePolicyBlocked, "policy_ref_missing", "host:policy_ref", "request_host_policy_or_budget_review")
	}
	if !displaySafeRefsContainAll(result.ApprovalBindingRefs, result.RequiredApprovalBindingRefs) {
		result = productionAdapterObjectiveCloseoutWriterRunnerGateBlock(result, FailureApprovalRequired, "approval_binding_ref_missing", "host:objective_closeout_writer_approval_binding_ref", "request_objective_closeout_writer_approval")
	}
	if result.BudgetRef == "" {
		result = productionAdapterObjectiveCloseoutWriterRunnerGateBlock(result, FailureBudgetExhausted, "budget_ref_missing", "host:budget_ref", "request_host_policy_or_budget_review")
	} else if result.RequiredBudgetRef != "" && result.BudgetRef != result.RequiredBudgetRef {
		result = productionAdapterObjectiveCloseoutWriterRunnerGateBlock(result, FailurePolicyBlocked, "budget_ref_mismatch", "host:budget_ref", "request_host_policy_or_budget_review")
	}
	if result.IdempotencyRef == "" {
		result = productionAdapterObjectiveCloseoutWriterRunnerGateBlock(result, FailureInvalidInput, "idempotency_ref_missing", "host:idempotency_ref", "provide_idempotency_ref")
	} else if bridge.IdempotencyRef != "" && result.IdempotencyRef != bridge.IdempotencyRef {
		result = productionAdapterObjectiveCloseoutWriterRunnerGateBlock(result, FailureInvalidInput, "idempotency_ref_mismatch", "host:idempotency_ref", "review_objective_closeout_writer_runner_gate")
	}
	if result.TimeoutPolicyRef == "" {
		result = productionAdapterObjectiveCloseoutWriterRunnerGateBlock(result, FailureTimeout, "timeout_policy_ref_missing", "host:timeout_policy_ref", "provide_timeout_policy_ref")
	} else if bridge.TimeoutPolicyRef != "" && result.TimeoutPolicyRef != bridge.TimeoutPolicyRef {
		result = productionAdapterObjectiveCloseoutWriterRunnerGateBlock(result, FailureTimeout, "timeout_policy_ref_mismatch", "host:timeout_policy_ref", "review_objective_closeout_writer_runner_gate")
	}
	if result.CancellationPolicyRef == "" {
		result = productionAdapterObjectiveCloseoutWriterRunnerGateBlock(result, FailureConfigMissing, "cancellation_policy_ref_missing", "host:cancellation_policy_ref", "provide_cancellation_policy_ref")
	}
	if result.DryRunSmokeRef == "" || result.DryRunResultRef == "" {
		result = productionAdapterObjectiveCloseoutWriterRunnerGateBlock(result, FailureVerificationFailed, "dry_run_first_evidence_missing", "host:objective_closeout_writer_dry_run_smoke", "review_objective_closeout_writer_dry_run_smoke")
	}
	if len(result.BlockedReasons) == 0 && len(result.MissingInputs) == 0 {
		result.Status = "ready_for_objective_closeout_writer_host_adapter_runner_gate"
		result.ReadyForHostDisplay = true
		result.ReadyForHostRunner = true
		result.ReadyForHostAdapterExecution = true
		result.HostRunnerAuthorized = true
		result.HostConfirmationBound = true
		result.DryRunFirstSatisfied = true
		result.PolicyBound = true
		result.ApprovalBound = true
		result.BudgetBound = true
		result.IdempotencyBound = true
		result.TimeoutBound = true
		result.HostMayInvokeWriterAdapter = true
		result.HostMayExecuteDurableWrite = true
		result.NextHostAction = "host_may_run_objective_closeout_writer_adapter"
		result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_objective_closeout_writer_host_adapter_runner_gate", "host_may_run_writer_adapter")
	}
	return result.Normalize()
}

func CloneProductionAdapterObjectiveCloseoutWriterRunnerGate(in ProductionAdapterObjectiveCloseoutWriterRunnerGate) ProductionAdapterObjectiveCloseoutWriterRunnerGate {
	out := in
	out.CapabilityProofRefs = cloneDisplaySafeRefs(in.CapabilityProofRefs)
	out.PolicyRefs = cloneDisplaySafeRefs(in.PolicyRefs)
	out.RequiredPolicyRefs = cloneDisplaySafeRefs(in.RequiredPolicyRefs)
	out.ApprovalBindingRefs = cloneDisplaySafeRefs(in.ApprovalBindingRefs)
	out.RequiredApprovalBindingRefs = cloneDisplaySafeRefs(in.RequiredApprovalBindingRefs)
	out.ApprovalRefs = cloneDisplaySafeRefs(in.ApprovalRefs)
	out.RequiredApprovalRefs = cloneDisplaySafeRefs(in.RequiredApprovalRefs)
	out.RequiredInputs = cloneMissingInputs(in.RequiredInputs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (g ProductionAdapterObjectiveCloseoutWriterRunnerGate) Clone() ProductionAdapterObjectiveCloseoutWriterRunnerGate {
	return CloneProductionAdapterObjectiveCloseoutWriterRunnerGate(g)
}

func (g ProductionAdapterObjectiveCloseoutWriterRunnerGate) Normalize() ProductionAdapterObjectiveCloseoutWriterRunnerGate {
	out := CloneProductionAdapterObjectiveCloseoutWriterRunnerGate(g)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = normalizeControlToken(out.Status)
	out.Mode = normalizeControlToken(out.Mode)
	if out.Mode == "" {
		out.Mode = "production_adapter_objective_closeout_writer_runner_gate"
	}
	out.RunnerGateRef = normalizeOneDisplaySafeRef(out.RunnerGateRef)
	out.BridgeRef = normalizeOneDisplaySafeRef(out.BridgeRef)
	out.HostRunnerRef = normalizeOneDisplaySafeRef(out.HostRunnerRef)
	out.HostRunnerVersionRef = normalizeOneDisplaySafeRef(out.HostRunnerVersionRef)
	out.RunnerInvocationRef = normalizeOneDisplaySafeRef(out.RunnerInvocationRef)
	out.HostConfirmationRef = normalizeOneDisplaySafeRef(out.HostConfirmationRef)
	out.HostUIHandoffRef = normalizeOneDisplaySafeRef(out.HostUIHandoffRef)
	out.PrimaryDisplayRef = normalizeOneDisplaySafeRef(out.PrimaryDisplayRef)
	out.ReviewRef = normalizeOneDisplaySafeRef(out.ReviewRef)
	out.FixtureRef = normalizeOneDisplaySafeRef(out.FixtureRef)
	out.InvocationEnvelopeRef = normalizeOneDisplaySafeRef(out.InvocationEnvelopeRef)
	out.ResultEnvelopeRef = normalizeOneDisplaySafeRef(out.ResultEnvelopeRef)
	out.ReviewPacketRef = normalizeOneDisplaySafeRef(out.ReviewPacketRef)
	out.DurableRequestRef = normalizeOneDisplaySafeRef(out.DurableRequestRef)
	out.ExpectedDurableResultRef = normalizeOneDisplaySafeRef(out.ExpectedDurableResultRef)
	out.WriterInvocationRef = normalizeOneDisplaySafeRef(out.WriterInvocationRef)
	out.WriterOptInRef = normalizeOneDisplaySafeRef(out.WriterOptInRef)
	out.WriterRef = normalizeOneDisplaySafeRef(out.WriterRef)
	out.HostWriterBindingRef = normalizeOneDisplaySafeRef(out.HostWriterBindingRef)
	out.HostAdapterVersionRef = normalizeOneDisplaySafeRef(out.HostAdapterVersionRef)
	out.ExpectedHostAdapterRunRef = normalizeOneDisplaySafeRef(out.ExpectedHostAdapterRunRef)
	out.DryRunSmokeRef = normalizeOneDisplaySafeRef(out.DryRunSmokeRef)
	out.DryRunResultRef = normalizeOneDisplaySafeRef(out.DryRunResultRef)
	out.ExpectedReadbackRef = normalizeOneDisplaySafeRef(out.ExpectedReadbackRef)
	out.ExpectedFailureRef = normalizeOneDisplaySafeRef(out.ExpectedFailureRef)
	out.ExpectedCompensationRef = normalizeOneDisplaySafeRef(out.ExpectedCompensationRef)
	out.ObjectiveCloseoutHandoffRef = normalizeOneDisplaySafeRef(out.ObjectiveCloseoutHandoffRef)
	out.ObjectiveCloseoutPacketRef = normalizeOneDisplaySafeRef(out.ObjectiveCloseoutPacketRef)
	out.ObjectiveRef = normalizeOneDisplaySafeRef(out.ObjectiveRef)
	out.HostRunstoreRef = normalizeOneDisplaySafeRef(out.HostRunstoreRef)
	out.ExpectedDurableEventRef = normalizeOneDisplaySafeRef(out.ExpectedDurableEventRef)
	out.ExpectedObjectiveStateRef = normalizeOneDisplaySafeRef(out.ExpectedObjectiveStateRef)
	out.HostDurableWriteConfirmationRef = normalizeOneDisplaySafeRef(out.HostDurableWriteConfirmationRef)
	out.CapabilityProofRefs = normalizeDisplaySafeRefs(out.CapabilityProofRefs)
	out.PolicyRefs = normalizeDisplaySafeRefs(out.PolicyRefs)
	out.RequiredPolicyRefs = normalizeDisplaySafeRefs(out.RequiredPolicyRefs)
	out.ApprovalBindingRefs = normalizeDisplaySafeRefs(out.ApprovalBindingRefs)
	out.RequiredApprovalBindingRefs = normalizeDisplaySafeRefs(out.RequiredApprovalBindingRefs)
	out.ApprovalRefs = normalizeDisplaySafeRefs(out.ApprovalRefs)
	out.RequiredApprovalRefs = normalizeDisplaySafeRefs(out.RequiredApprovalRefs)
	out.BudgetRef = normalizeOneDisplaySafeRef(out.BudgetRef)
	out.RequiredBudgetRef = normalizeOneDisplaySafeRef(out.RequiredBudgetRef)
	out.IdempotencyRef = normalizeOneDisplaySafeRef(out.IdempotencyRef)
	out.IdempotencyContractRef = normalizeOneDisplaySafeRef(out.IdempotencyContractRef)
	out.TimeoutPolicyRef = normalizeOneDisplaySafeRef(out.TimeoutPolicyRef)
	out.CancellationPolicyRef = normalizeOneDisplaySafeRef(out.CancellationPolicyRef)
	out.RequiredInputs = normalizeMissingInputs(out.RequiredInputs)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.BlockedReasons = normalizeControlTokenList(out.BlockedReasons)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
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
	out.CoreInvocationExecuted = false
	out.DryRunByCore = false
	out.DurableWriteByCore = false
	out.ObjectiveStoreWriteByCore = false
	out.RunstoreWriteByCore = false
	if !out.Available {
		out.Status = "unavailable"
		out.ReadyForHostDisplay = false
		out.ReadyForHostRunner = false
		out.ReadyForHostAdapterExecution = false
		out.HostRunnerAuthorized = false
		out.HostMayInvokeWriterAdapter = false
		out.HostMayExecuteDurableWrite = false
	}
	if out.Status == "" {
		out.Status = "blocked"
	}
	if out.RawOutputLoaded || productionAdapterObjectiveCloseoutWriterRunnerGateUnsafeOutput(out) {
		out.RawOutputLoaded = true
		out = productionAdapterObjectiveCloseoutWriterRunnerGateBlock(out, firstFailureClass(out.FailureClass, FailureEvidenceWeak), "unsafe_input_ref", "host:display_safe_refs", firstNextHostAction(out.NextHostAction, "provide_display_safe_refs"))
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
	}
	out.ReadyForHostDisplay = out.ReadyForHostDisplay &&
		out.Available &&
		out.RunnerGateRef != "" &&
		out.BridgeRef != "" &&
		out.HostRunnerRef != "" &&
		out.RunnerInvocationRef != "" &&
		out.WriterRef != "" &&
		!out.RawOutputLoaded
	ready := out.Status == "ready_for_objective_closeout_writer_host_adapter_runner_gate" &&
		out.ReadyForHostDisplay &&
		out.RunnerGateRef != "" &&
		out.BridgeRef != "" &&
		out.HostRunnerRef != "" &&
		out.HostRunnerVersionRef != "" &&
		out.RunnerInvocationRef != "" &&
		out.HostConfirmationRef != "" &&
		out.InvocationEnvelopeRef != "" &&
		out.DurableRequestRef != "" &&
		out.WriterInvocationRef != "" &&
		out.WriterRef != "" &&
		out.HostWriterBindingRef != "" &&
		out.HostAdapterVersionRef != "" &&
		out.ExpectedHostAdapterRunRef != "" &&
		out.ExpectedDurableResultRef != "" &&
		out.DryRunSmokeRef != "" &&
		out.DryRunResultRef != "" &&
		out.ExpectedReadbackRef != "" &&
		out.ExpectedFailureRef != "" &&
		out.ExpectedCompensationRef != "" &&
		out.HostDurableWriteConfirmationRef != "" &&
		len(out.CapabilityProofRefs) > 0 &&
		displaySafeRefsContainAll(out.PolicyRefs, out.RequiredPolicyRefs) &&
		displaySafeRefsContainAll(out.ApprovalBindingRefs, out.RequiredApprovalBindingRefs) &&
		out.BudgetRef != "" &&
		(out.RequiredBudgetRef == "" || out.BudgetRef == out.RequiredBudgetRef) &&
		out.IdempotencyRef != "" &&
		out.TimeoutPolicyRef != "" &&
		out.CancellationPolicyRef != "" &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.RawOutputLoaded &&
		!out.CoreInvocationExecuted &&
		!out.DryRunByCore &&
		!out.DurableWriteByCore &&
		!out.ObjectiveStoreWriteByCore &&
		!out.RunstoreWriteByCore
	out.ReadyForHostRunner = out.ReadyForHostRunner && ready
	out.ReadyForHostAdapterExecution = out.ReadyForHostAdapterExecution && ready
	out.HostRunnerAuthorized = out.HostRunnerAuthorized && ready
	out.HostConfirmationBound = out.HostConfirmationBound && ready
	out.DryRunFirstSatisfied = out.DryRunFirstSatisfied && ready
	out.PolicyBound = out.PolicyBound && ready
	out.ApprovalBound = out.ApprovalBound && ready
	out.BudgetBound = out.BudgetBound && ready
	out.IdempotencyBound = out.IdempotencyBound && ready
	out.TimeoutBound = out.TimeoutBound && ready
	out.HostMayInvokeWriterAdapter = out.HostMayInvokeWriterAdapter && ready
	out.HostMayExecuteDurableWrite = out.HostMayExecuteDurableWrite && ready
	return out
}

func productionAdapterObjectiveCloseoutWriterRunnerGateBlock(result ProductionAdapterObjectiveCloseoutWriterRunnerGate, failure FailureClass, reason string, missing MissingInput, next NextHostAction) ProductionAdapterObjectiveCloseoutWriterRunnerGate {
	result.Status = "blocked"
	result.ReadyForHostRunner = false
	result.ReadyForHostAdapterExecution = false
	result.HostRunnerAuthorized = false
	result.HostConfirmationBound = false
	result.DryRunFirstSatisfied = false
	result.PolicyBound = false
	result.ApprovalBound = false
	result.BudgetBound = false
	result.IdempotencyBound = false
	result.TimeoutBound = false
	result.HostMayInvokeWriterAdapter = false
	result.HostMayExecuteDurableWrite = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	result.Boundaries = AppendBoundaries(result.Boundaries, "objective_closeout_writer_runner_gate_blocked")
	return result
}

func productionAdapterObjectiveCloseoutWriterRunnerGateRequiredInputs() []MissingInput {
	return []MissingInput{
		"host:objective_closeout_writer_runner_gate_ref",
		"host:objective_closeout_writer_execution_bridge",
		"host:objective_closeout_writer_runner_ref",
		"host:objective_closeout_writer_runner_version_ref",
		"host:objective_closeout_writer_runner_invocation_ref",
		"host:objective_closeout_writer_host_confirmation_ref",
		"host:policy_ref",
		"host:objective_closeout_writer_approval_binding_ref",
		"host:budget_ref",
		"host:idempotency_ref",
		"host:timeout_policy_ref",
		"host:cancellation_policy_ref",
	}
}

func productionAdapterObjectiveCloseoutWriterRunnerGateBoundaries(groups ...[]Boundary) []Boundary {
	return MergeBoundaries(
		[]Boundary{
			"production_adapter_objective_closeout_writer_runner_gate",
			"objective_closeout_writer_runner_gate_projection_only",
			"host_owned_objective_closeout_writer_adapter",
			"host_owned_writer_adapter_runner_facade",
			"dry_run_first_runtime_gate",
			"explicit_host_confirmation_required",
			"display_safe_refs_only",
			"display_safe_result_refs_only",
			"no_runner_dispatch",
			"no_adapter_invocation_by_core",
			"no_dry_run_by_core",
			"no_durable_write_by_core",
			"no_objective_store_write_by_core",
			"no_runstore_write_by_core",
			"no_compensation_execution_by_core",
		},
		MergeBoundaries(groups...),
	)
}

func productionAdapterObjectiveCloseoutWriterRunnerGateUnsafe(input ProductionAdapterObjectiveCloseoutWriterRunnerGateInput, bridge ProductionAdapterObjectiveCloseoutWriterExecutionBridge) bool {
	return input.RawOutputLoaded ||
		displaySafeRefRejected(input.RunnerGateRef) ||
		displaySafeRefRejected(input.HostRunnerRef) ||
		displaySafeRefRejected(input.HostRunnerVersionRef) ||
		displaySafeRefRejected(input.RunnerInvocationRef) ||
		displaySafeRefRejected(input.HostConfirmationRef) ||
		displaySafeRefSliceRejected(input.PolicyRefs) ||
		displaySafeRefSliceRejected(input.ApprovalBindingRefs) ||
		displaySafeRefRejected(input.BudgetRef) ||
		displaySafeRefRejected(input.IdempotencyRef) ||
		displaySafeRefRejected(input.TimeoutPolicyRef) ||
		displaySafeRefRejected(input.CancellationPolicyRef) ||
		productionAdapterObjectiveCloseoutWriterExecutionBridgeUnsafeOutput(bridge)
}

func productionAdapterObjectiveCloseoutWriterRunnerGateUnsafeOutput(input ProductionAdapterObjectiveCloseoutWriterRunnerGate) bool {
	return displaySafeRefRejected(input.RunnerGateRef) ||
		displaySafeRefRejected(input.BridgeRef) ||
		displaySafeRefRejected(input.HostRunnerRef) ||
		displaySafeRefRejected(input.HostRunnerVersionRef) ||
		displaySafeRefRejected(input.RunnerInvocationRef) ||
		displaySafeRefRejected(input.HostConfirmationRef) ||
		displaySafeRefRejected(input.HostUIHandoffRef) ||
		displaySafeRefRejected(input.PrimaryDisplayRef) ||
		displaySafeRefRejected(input.ReviewRef) ||
		displaySafeRefRejected(input.FixtureRef) ||
		displaySafeRefRejected(input.InvocationEnvelopeRef) ||
		displaySafeRefRejected(input.ResultEnvelopeRef) ||
		displaySafeRefRejected(input.ReviewPacketRef) ||
		displaySafeRefRejected(input.DurableRequestRef) ||
		displaySafeRefRejected(input.ExpectedDurableResultRef) ||
		displaySafeRefRejected(input.WriterInvocationRef) ||
		displaySafeRefRejected(input.WriterOptInRef) ||
		displaySafeRefRejected(input.WriterRef) ||
		displaySafeRefRejected(input.HostWriterBindingRef) ||
		displaySafeRefRejected(input.HostAdapterVersionRef) ||
		displaySafeRefRejected(input.ExpectedHostAdapterRunRef) ||
		displaySafeRefRejected(input.DryRunSmokeRef) ||
		displaySafeRefRejected(input.DryRunResultRef) ||
		displaySafeRefRejected(input.ExpectedReadbackRef) ||
		displaySafeRefRejected(input.ExpectedFailureRef) ||
		displaySafeRefRejected(input.ExpectedCompensationRef) ||
		displaySafeRefRejected(input.ObjectiveCloseoutHandoffRef) ||
		displaySafeRefRejected(input.ObjectiveCloseoutPacketRef) ||
		displaySafeRefRejected(input.ObjectiveRef) ||
		displaySafeRefRejected(input.HostRunstoreRef) ||
		displaySafeRefRejected(input.ExpectedDurableEventRef) ||
		displaySafeRefRejected(input.ExpectedObjectiveStateRef) ||
		displaySafeRefRejected(input.HostDurableWriteConfirmationRef) ||
		displaySafeRefSliceRejected(input.CapabilityProofRefs) ||
		displaySafeRefSliceRejected(input.PolicyRefs) ||
		displaySafeRefSliceRejected(input.RequiredPolicyRefs) ||
		displaySafeRefSliceRejected(input.ApprovalBindingRefs) ||
		displaySafeRefSliceRejected(input.RequiredApprovalBindingRefs) ||
		displaySafeRefSliceRejected(input.ApprovalRefs) ||
		displaySafeRefSliceRejected(input.RequiredApprovalRefs) ||
		displaySafeRefRejected(input.BudgetRef) ||
		displaySafeRefRejected(input.RequiredBudgetRef) ||
		displaySafeRefRejected(input.IdempotencyRef) ||
		displaySafeRefRejected(input.IdempotencyContractRef) ||
		displaySafeRefRejected(input.TimeoutPolicyRef) ||
		displaySafeRefRejected(input.CancellationPolicyRef) ||
		input.RawOutputLoaded
}

func productionAdapterObjectiveCloseoutWriterExecutionBridgeEmpty(bridge ProductionAdapterObjectiveCloseoutWriterExecutionBridge) bool {
	return !bridge.Projected &&
		!bridge.Available &&
		bridge.Status == "" &&
		bridge.Mode == "" &&
		bridge.BridgeRef == "" &&
		bridge.HostUIHandoffRef == "" &&
		bridge.InvocationEnvelopeRef == "" &&
		bridge.DurableRequestRef == "" &&
		bridge.WriterRef == "" &&
		len(bridge.MissingInputs) == 0 &&
		len(bridge.BlockedReasons) == 0 &&
		len(bridge.Boundaries) == 0 &&
		bridge.NextHostAction == "" &&
		!bridge.RawOutputLoaded
}

func unavailableProductionAdapterObjectiveCloseoutWriterRunnerGate() ProductionAdapterObjectiveCloseoutWriterRunnerGate {
	return ProductionAdapterObjectiveCloseoutWriterRunnerGate{
		ContractVersion: ContractVersion,
		Projected:       true,
		Available:       false,
		Status:          "unavailable",
		Mode:            "production_adapter_objective_closeout_writer_runner_gate",
		RequiredInputs:  productionAdapterObjectiveCloseoutWriterRunnerGateRequiredInputs(),
		Boundaries: []Boundary{
			"production_adapter_objective_closeout_writer_runner_gate",
			"objective_closeout_writer_runner_gate_projection_only",
			"host_owned_objective_closeout_writer_adapter",
			"display_safe_refs_only",
			"no_adapter_invocation_by_core",
			"no_runner_dispatch",
			"no_durable_write_by_core",
		},
		NextHostAction: "provide_objective_closeout_writer_execution_bridge",
		RunnerEffect:   "none",
		PromptEffect:   "none",
	}
}
