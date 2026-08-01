package controlcontract

type ProductionAdapterCompensableEffectKind string

const (
	ProductionAdapterCompensableEffectSchedulerApply   ProductionAdapterCompensableEffectKind = "scheduler_apply"
	ProductionAdapterCompensableEffectInstallerApply   ProductionAdapterCompensableEffectKind = "installer_apply"
	ProductionAdapterCompensableEffectStoreMutation    ProductionAdapterCompensableEffectKind = "store_mutation"
	ProductionAdapterCompensableEffectWorkflowRuntime  ProductionAdapterCompensableEffectKind = "workflow_runtime"
	ProductionAdapterCompensableEffectDelegationWorker ProductionAdapterCompensableEffectKind = "delegation_worker_runtime"
)

func KnownProductionAdapterCompensableEffectKinds() []ProductionAdapterCompensableEffectKind {
	return []ProductionAdapterCompensableEffectKind{
		ProductionAdapterCompensableEffectSchedulerApply,
		ProductionAdapterCompensableEffectInstallerApply,
		ProductionAdapterCompensableEffectStoreMutation,
		ProductionAdapterCompensableEffectWorkflowRuntime,
		ProductionAdapterCompensableEffectDelegationWorker,
	}
}

func NormalizeProductionAdapterCompensableEffectKind(raw string) ProductionAdapterCompensableEffectKind {
	switch normalizeEnumToken(raw) {
	case "scheduler_apply", "schedule_apply", "operations_schedule_apply":
		return ProductionAdapterCompensableEffectSchedulerApply
	case "installer_apply", "install_apply", "capability_install_apply", "capability_apply":
		return ProductionAdapterCompensableEffectInstallerApply
	case "store_mutation", "objective_store_mutation", "run_store_mutation", "runstore_mutation", "objective_run_store":
		return ProductionAdapterCompensableEffectStoreMutation
	case "workflow_runtime", "workflow_runtime_executor", "runtime_executor", "workflow_retry_apply", "workflow_retry":
		return ProductionAdapterCompensableEffectWorkflowRuntime
	case "delegation_worker_runtime", "delegation_worker", "worker_runtime", "worker_dispatch":
		return ProductionAdapterCompensableEffectDelegationWorker
	default:
		return ""
	}
}

type ProductionAdapterCompensationCapabilitySpec struct {
	EffectKind                ProductionAdapterCompensableEffectKind `json:"effect_kind,omitempty"`
	EffectRef                 DisplaySafeRef                         `json:"effect_ref,omitempty"`
	EffectGateRef             DisplaySafeRef                         `json:"effect_gate_ref,omitempty"`
	AdapterRef                DisplaySafeRef                         `json:"adapter_ref,omitempty"`
	CompensationCapabilityRef DisplaySafeRef                         `json:"compensation_capability_ref,omitempty"`
	CompensationRef           DisplaySafeRef                         `json:"compensation_ref,omitempty"`
	CompensationExecutorRef   DisplaySafeRef                         `json:"compensation_executor_ref,omitempty"`
	CompensationPolicyRef     DisplaySafeRef                         `json:"compensation_policy_ref,omitempty"`
	CompensationApprovalRef   DisplaySafeRef                         `json:"compensation_approval_ref,omitempty"`
	IdempotencyContractRef    DisplaySafeRef                         `json:"idempotency_contract_ref,omitempty"`
	ReadbackContractRef       DisplaySafeRef                         `json:"readback_contract_ref,omitempty"`
	RollbackContractRef       DisplaySafeRef                         `json:"rollback_contract_ref,omitempty"`
	FailureReviewRef          DisplaySafeRef                         `json:"failure_review_ref,omitempty"`
	ResidualRiskRef           DisplaySafeRef                         `json:"residual_risk_ref,omitempty"`
	EvidenceRefs              []DisplaySafeRef                       `json:"evidence_refs,omitempty"`
	Boundaries                []Boundary                             `json:"boundaries,omitempty"`
	RawOutputLoaded           bool                                   `json:"raw_output_loaded"`
}

type ProductionAdapterCompensationCapabilityDeclaration struct {
	ContractVersion                string                                 `json:"contract_version,omitempty"`
	Projected                      bool                                   `json:"projected"`
	Status                         HostActionStatus                       `json:"status,omitempty"`
	EffectKind                     ProductionAdapterCompensableEffectKind `json:"effect_kind,omitempty"`
	ReadyForCompensationPlan       bool                                   `json:"ready_for_compensation_plan"`
	CompensationCapabilityDeclared bool                                   `json:"compensation_capability_declared"`
	ResidualRiskDeclared           bool                                   `json:"residual_risk_declared"`
	EffectRef                      DisplaySafeRef                         `json:"effect_ref,omitempty"`
	EffectGateRef                  DisplaySafeRef                         `json:"effect_gate_ref,omitempty"`
	AdapterRef                     DisplaySafeRef                         `json:"adapter_ref,omitempty"`
	CompensationCapabilityRef      DisplaySafeRef                         `json:"compensation_capability_ref,omitempty"`
	CompensationRef                DisplaySafeRef                         `json:"compensation_ref,omitempty"`
	CompensationExecutorRef        DisplaySafeRef                         `json:"compensation_executor_ref,omitempty"`
	CompensationPolicyRef          DisplaySafeRef                         `json:"compensation_policy_ref,omitempty"`
	CompensationApprovalRef        DisplaySafeRef                         `json:"compensation_approval_ref,omitempty"`
	IdempotencyContractRef         DisplaySafeRef                         `json:"idempotency_contract_ref,omitempty"`
	ReadbackContractRef            DisplaySafeRef                         `json:"readback_contract_ref,omitempty"`
	RollbackContractRef            DisplaySafeRef                         `json:"rollback_contract_ref,omitempty"`
	FailureReviewRef               DisplaySafeRef                         `json:"failure_review_ref,omitempty"`
	ResidualRiskRef                DisplaySafeRef                         `json:"residual_risk_ref,omitempty"`
	EvidenceRefs                   []DisplaySafeRef                       `json:"evidence_refs,omitempty"`
	MissingInputs                  []MissingInput                         `json:"missing_inputs,omitempty"`
	BlockedReasons                 []string                               `json:"blocked_reasons,omitempty"`
	FailureClass                   FailureClass                           `json:"failure_class,omitempty"`
	Boundaries                     []Boundary                             `json:"boundaries,omitempty"`
	NextHostAction                 NextHostAction                         `json:"next_host_action,omitempty"`
	RunnerEffect                   string                                 `json:"runner_effect,omitempty"`
	PromptEffect                   string                                 `json:"prompt_effect,omitempty"`
	CoreExecutionExecuted          bool                                   `json:"core_execution_executed"`
	RunnerDispatched               bool                                   `json:"runner_dispatched"`
	ToolExecuted                   bool                                   `json:"tool_executed"`
	WorkflowDispatched             bool                                   `json:"workflow_dispatched"`
	SchedulerApplied               bool                                   `json:"scheduler_applied"`
	InstallerExecuted              bool                                   `json:"installer_executed"`
	StoreMutationExecuted          bool                                   `json:"store_mutation_executed"`
	CompensationExecutedByCore     bool                                   `json:"compensation_executed_by_core"`
	RawOutputLoaded                bool                                   `json:"raw_output_loaded"`
}

type ProductionAdapterCompensationCapabilityPlanInput struct {
	PlanRef                DisplaySafeRef                                `json:"plan_ref,omitempty"`
	ExecutorDescriptorRef  DisplaySafeRef                                `json:"executor_descriptor_ref,omitempty"`
	ExecutorRef            DisplaySafeRef                                `json:"executor_ref,omitempty"`
	OwnerRef               DisplaySafeRef                                `json:"owner_ref,omitempty"`
	IdempotencyContractRef DisplaySafeRef                                `json:"idempotency_contract_ref,omitempty"`
	ReadbackContractRef    DisplaySafeRef                                `json:"readback_contract_ref,omitempty"`
	RollbackContractRef    DisplaySafeRef                                `json:"rollback_contract_ref,omitempty"`
	TimeoutPolicyRef       DisplaySafeRef                                `json:"timeout_policy_ref,omitempty"`
	RequiredPolicyRefs     []DisplaySafeRef                              `json:"required_policy_refs,omitempty"`
	RequiredApprovalRefs   []DisplaySafeRef                              `json:"required_approval_refs,omitempty"`
	CapabilitySpecs        []ProductionAdapterCompensationCapabilitySpec `json:"capability_specs,omitempty"`
	Boundaries             []Boundary                                    `json:"boundaries,omitempty"`
	RawOutputLoaded        bool                                          `json:"raw_output_loaded"`
}

type ProductionAdapterCompensationCapabilityPlan struct {
	ContractVersion                        string                                               `json:"contract_version,omitempty"`
	Projected                              bool                                                 `json:"projected"`
	Status                                 HostActionStatus                                     `json:"status,omitempty"`
	ReadyForCompensationCapabilities       bool                                                 `json:"ready_for_compensation_capabilities"`
	ReadyForCompensationExecutorDescriptor bool                                                 `json:"ready_for_compensation_executor_descriptor"`
	AllRequiredCapabilitiesDeclared        bool                                                 `json:"all_required_capabilities_declared"`
	ResidualRiskDeclared                   bool                                                 `json:"residual_risk_declared"`
	PlanRef                                DisplaySafeRef                                       `json:"plan_ref,omitempty"`
	ExecutorDescriptorRef                  DisplaySafeRef                                       `json:"executor_descriptor_ref,omitempty"`
	ExecutorRef                            DisplaySafeRef                                       `json:"executor_ref,omitempty"`
	OwnerRef                               DisplaySafeRef                                       `json:"owner_ref,omitempty"`
	IdempotencyContractRef                 DisplaySafeRef                                       `json:"idempotency_contract_ref,omitempty"`
	ReadbackContractRef                    DisplaySafeRef                                       `json:"readback_contract_ref,omitempty"`
	RollbackContractRef                    DisplaySafeRef                                       `json:"rollback_contract_ref,omitempty"`
	TimeoutPolicyRef                       DisplaySafeRef                                       `json:"timeout_policy_ref,omitempty"`
	EffectKinds                            []ProductionAdapterCompensableEffectKind             `json:"effect_kinds,omitempty"`
	EffectRefs                             []DisplaySafeRef                                     `json:"effect_refs,omitempty"`
	EffectGateRefs                         []DisplaySafeRef                                     `json:"effect_gate_refs,omitempty"`
	AdapterRefs                            []DisplaySafeRef                                     `json:"adapter_refs,omitempty"`
	CompensationCapabilityRefs             []DisplaySafeRef                                     `json:"compensation_capability_refs,omitempty"`
	CompensationRefs                       []DisplaySafeRef                                     `json:"compensation_refs,omitempty"`
	CompensationExecutorRefs               []DisplaySafeRef                                     `json:"compensation_executor_refs,omitempty"`
	CompensationPolicyRefs                 []DisplaySafeRef                                     `json:"compensation_policy_refs,omitempty"`
	CompensationApprovalRefs               []DisplaySafeRef                                     `json:"compensation_approval_refs,omitempty"`
	FailureReviewRefs                      []DisplaySafeRef                                     `json:"failure_review_refs,omitempty"`
	ResidualRiskRefs                       []DisplaySafeRef                                     `json:"residual_risk_refs,omitempty"`
	RequiredPolicyRefs                     []DisplaySafeRef                                     `json:"required_policy_refs,omitempty"`
	RequiredApprovalRefs                   []DisplaySafeRef                                     `json:"required_approval_refs,omitempty"`
	Declarations                           []ProductionAdapterCompensationCapabilityDeclaration `json:"declarations,omitempty"`
	ExecutorDescriptor                     ObjectiveCompensationExecutorDescriptor              `json:"executor_descriptor,omitempty"`
	MissingInputs                          []MissingInput                                       `json:"missing_inputs,omitempty"`
	BlockedReasons                         []string                                             `json:"blocked_reasons,omitempty"`
	FailureClass                           FailureClass                                         `json:"failure_class,omitempty"`
	Boundaries                             []Boundary                                           `json:"boundaries,omitempty"`
	NextHostAction                         NextHostAction                                       `json:"next_host_action,omitempty"`
	RunnerEffect                           string                                               `json:"runner_effect,omitempty"`
	PromptEffect                           string                                               `json:"prompt_effect,omitempty"`
	CoreExecutionExecuted                  bool                                                 `json:"core_execution_executed"`
	RunnerDispatched                       bool                                                 `json:"runner_dispatched"`
	ToolExecuted                           bool                                                 `json:"tool_executed"`
	WorkflowDispatched                     bool                                                 `json:"workflow_dispatched"`
	SchedulerApplied                       bool                                                 `json:"scheduler_applied"`
	InstallerExecuted                      bool                                                 `json:"installer_executed"`
	StoreMutationExecuted                  bool                                                 `json:"store_mutation_executed"`
	CompensationExecutedByCore             bool                                                 `json:"compensation_executed_by_core"`
	RawOutputLoaded                        bool                                                 `json:"raw_output_loaded"`
}

func BuildProductionAdapterCompensationCapabilityDeclaration(input ProductionAdapterCompensationCapabilitySpec) ProductionAdapterCompensationCapabilityDeclaration {
	kind := NormalizeProductionAdapterCompensableEffectKind(string(input.EffectKind))
	result := ProductionAdapterCompensationCapabilityDeclaration{
		ContractVersion:           ContractVersion,
		Projected:                 true,
		Status:                    HostActionBlocked,
		EffectKind:                kind,
		EffectRef:                 normalizeOneDisplaySafeRef(input.EffectRef),
		EffectGateRef:             normalizeOneDisplaySafeRef(input.EffectGateRef),
		AdapterRef:                normalizeOneDisplaySafeRef(input.AdapterRef),
		CompensationCapabilityRef: normalizeOneDisplaySafeRef(input.CompensationCapabilityRef),
		CompensationRef:           normalizeOneDisplaySafeRef(input.CompensationRef),
		CompensationExecutorRef:   normalizeOneDisplaySafeRef(input.CompensationExecutorRef),
		CompensationPolicyRef:     normalizeOneDisplaySafeRef(input.CompensationPolicyRef),
		CompensationApprovalRef:   normalizeOneDisplaySafeRef(input.CompensationApprovalRef),
		IdempotencyContractRef:    normalizeOneDisplaySafeRef(input.IdempotencyContractRef),
		ReadbackContractRef:       normalizeOneDisplaySafeRef(input.ReadbackContractRef),
		RollbackContractRef:       normalizeOneDisplaySafeRef(input.RollbackContractRef),
		FailureReviewRef:          normalizeOneDisplaySafeRef(input.FailureReviewRef),
		ResidualRiskRef:           normalizeOneDisplaySafeRef(input.ResidualRiskRef),
		EvidenceRefs:              normalizeDisplaySafeRefs(input.EvidenceRefs),
		FailureClass:              FailureNone,
		Boundaries: MergeBoundaries(
			[]Boundary{
				"production_adapter_compensation_capability_declaration",
				"host_owned_compensation_capability_declaration",
				"compensation_capability_declaration_only",
				"display_safe_refs_only",
				"no_compensation_execution_by_core",
				"no_runner_dispatch",
				"no_tool_execution_by_core",
				"no_workflow_dispatch_by_core",
				"no_scheduler_apply_by_core",
				"no_install_apply_by_core",
				"no_store_mutation_by_core",
			},
			input.Boundaries,
		),
		NextHostAction:  "declare_compensation_capability",
		RunnerEffect:    "none",
		PromptEffect:    "none",
		RawOutputLoaded: input.RawOutputLoaded,
	}
	if kind != "" {
		result.Boundaries = AppendBoundaries(result.Boundaries, Boundary(string(kind)+"_compensation_capability_declaration"))
	}
	if productionAdapterCompensationCapabilitySpecUnsafe(input) {
		result = productionAdapterCompensationCapabilityDeclarationBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		result.RawOutputLoaded = true
		return result.Normalize()
	}
	if kind == "" {
		result = productionAdapterCompensationCapabilityDeclarationBlock(result, FailureInvalidInput, "compensable_effect_kind_missing", "host:compensable_effect_kind", "provide_compensable_effect_kind")
		return result.Normalize()
	}
	if !productionAdapterCompensationCapabilitySpecHasCapability(input) {
		if result.ResidualRiskRef != "" {
			result.Status = HostActionRecorded
			result.ResidualRiskDeclared = true
			result.NextHostAction = NextHostAction("review_" + string(kind) + "_compensation_residual_risk")
			result.Boundaries = AppendBoundaries(result.Boundaries, "compensation_capability_absent_residual_risk_declared")
			return result.Normalize()
		}
		result = productionAdapterCompensationCapabilityDeclarationBlock(result, FailureHostAdapterMissing, string(kind)+"_compensation_capability_missing", MissingInput("host:"+string(kind)+"_compensation_capability_ref"), "declare_compensation_capability")
		result = productionAdapterCompensationCapabilityDeclarationBlock(result, FailureEvidenceMissing, string(kind)+"_residual_risk_ref_missing", MissingInput("host:"+string(kind)+"_residual_risk_ref"), "record_residual_risk")
		return result.Normalize()
	}
	for _, check := range []struct {
		ref     DisplaySafeRef
		suffix  string
		failure FailureClass
		next    NextHostAction
	}{
		{result.EffectRef, "effect_ref", FailureEvidenceMissing, "provide_compensable_effect_ref"},
		{result.EffectGateRef, "effect_gate_ref", FailureEvidenceMissing, "provide_compensable_effect_gate_ref"},
		{result.AdapterRef, "adapter_ref", FailureHostAdapterMissing, "provide_compensable_effect_adapter_ref"},
		{result.CompensationCapabilityRef, "compensation_capability_ref", FailureHostAdapterMissing, "declare_compensation_capability"},
		{result.CompensationRef, "compensation_ref", FailureEvidenceMissing, "provide_compensation_ref"},
		{result.CompensationExecutorRef, "compensation_executor_ref", FailureHostAdapterMissing, "configure_compensation_executor"},
		{result.CompensationPolicyRef, "compensation_policy_ref", FailurePolicyBlocked, "provide_compensation_policy"},
		{result.CompensationApprovalRef, "compensation_approval_ref", FailureApprovalRequired, "request_host_compensation_approval"},
		{result.IdempotencyContractRef, "idempotency_contract_ref", FailureConfigMissing, "provide_compensation_idempotency_contract"},
		{result.ReadbackContractRef, "readback_contract_ref", FailureConfigMissing, "provide_compensation_readback_contract"},
		{result.RollbackContractRef, "rollback_contract_ref", FailureConfigMissing, "provide_compensation_rollback_contract"},
		{result.FailureReviewRef, "failure_review_ref", FailureConfigMissing, "provide_compensation_failure_review_ref"},
	} {
		if check.ref == "" {
			result = productionAdapterCompensationCapabilityDeclarationBlock(result, check.failure, string(kind)+"_"+check.suffix+"_missing", MissingInput("host:"+string(kind)+"_"+check.suffix), check.next)
		}
	}
	if len(result.MissingInputs) == 0 && len(result.BlockedReasons) == 0 {
		result.Status = HostActionReady
		result.ReadyForCompensationPlan = true
		result.CompensationCapabilityDeclared = true
		result.NextHostAction = NextHostAction("host_may_bind_" + string(kind) + "_compensation_capability")
		result.Boundaries = AppendBoundaries(result.Boundaries, Boundary(string(kind)+"_compensation_capability_ready"))
	}
	return result.Normalize()
}

func BuildProductionAdapterCompensationCapabilityPlan(input ProductionAdapterCompensationCapabilityPlanInput) ProductionAdapterCompensationCapabilityPlan {
	specsByKind := map[ProductionAdapterCompensableEffectKind]ProductionAdapterCompensationCapabilitySpec{}
	duplicateKinds := map[ProductionAdapterCompensableEffectKind]struct{}{}
	invalidSpecCount := 0
	for _, spec := range input.CapabilitySpecs {
		kind := NormalizeProductionAdapterCompensableEffectKind(string(spec.EffectKind))
		if kind == "" {
			invalidSpecCount++
			continue
		}
		spec.EffectKind = kind
		if _, exists := specsByKind[kind]; exists {
			duplicateKinds[kind] = struct{}{}
			continue
		}
		specsByKind[kind] = spec
	}
	result := ProductionAdapterCompensationCapabilityPlan{
		ContractVersion:        ContractVersion,
		Projected:              true,
		Status:                 HostActionBlocked,
		PlanRef:                normalizeOneDisplaySafeRef(input.PlanRef),
		ExecutorDescriptorRef:  normalizeOneDisplaySafeRef(input.ExecutorDescriptorRef),
		ExecutorRef:            normalizeOneDisplaySafeRef(input.ExecutorRef),
		OwnerRef:               normalizeOneDisplaySafeRef(input.OwnerRef),
		IdempotencyContractRef: normalizeOneDisplaySafeRef(input.IdempotencyContractRef),
		ReadbackContractRef:    normalizeOneDisplaySafeRef(input.ReadbackContractRef),
		RollbackContractRef:    normalizeOneDisplaySafeRef(input.RollbackContractRef),
		TimeoutPolicyRef:       normalizeOneDisplaySafeRef(input.TimeoutPolicyRef),
		RequiredPolicyRefs:     normalizeDisplaySafeRefs(input.RequiredPolicyRefs),
		RequiredApprovalRefs:   normalizeDisplaySafeRefs(input.RequiredApprovalRefs),
		FailureClass:           FailureNone,
		Boundaries: MergeBoundaries(
			[]Boundary{
				"production_adapter_compensation_capability_plan",
				"host_owned_compensation_capability_plan",
				"compensation_capability_declaration_only",
				"display_safe_refs_only",
				"no_compensation_execution_by_core",
				"no_runner_dispatch",
				"no_tool_execution_by_core",
				"no_workflow_dispatch_by_core",
				"no_scheduler_apply_by_core",
				"no_install_apply_by_core",
				"no_store_mutation_by_core",
			},
			input.Boundaries,
		),
		NextHostAction:  "review_compensation_capability_plan",
		RunnerEffect:    "none",
		PromptEffect:    "none",
		RawOutputLoaded: input.RawOutputLoaded,
	}
	if productionAdapterCompensationCapabilityPlanInputUnsafe(input) {
		result = productionAdapterCompensationCapabilityPlanBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		result.RawOutputLoaded = true
		return result.Normalize()
	}
	for _, check := range []struct {
		ref     DisplaySafeRef
		reason  string
		missing MissingInput
		next    NextHostAction
	}{
		{result.PlanRef, "compensation_capability_plan_ref_missing", "host:compensation_capability_plan_ref", "provide_compensation_capability_plan"},
		{result.ExecutorDescriptorRef, "compensation_executor_descriptor_ref_missing", "host:compensation_executor_descriptor_ref", "provide_compensation_executor_descriptor_ref"},
		{result.ExecutorRef, "compensation_executor_ref_missing", "host:compensation_executor_ref", "configure_compensation_executor"},
		{result.OwnerRef, "compensation_executor_owner_ref_missing", "host:compensation_executor_owner_ref", "provide_compensation_executor_owner_ref"},
		{result.IdempotencyContractRef, "compensation_idempotency_contract_missing", "contract:compensation_idempotency", "provide_compensation_idempotency_contract"},
		{result.ReadbackContractRef, "compensation_readback_contract_missing", "contract:compensation_readback", "provide_compensation_readback_contract"},
		{result.RollbackContractRef, "compensation_rollback_contract_missing", "contract:compensation_rollback", "provide_compensation_rollback_contract"},
	} {
		if check.ref == "" {
			result = productionAdapterCompensationCapabilityPlanBlock(result, FailureConfigMissing, check.reason, check.missing, check.next)
		}
	}
	if invalidSpecCount > 0 {
		result = productionAdapterCompensationCapabilityPlanBlock(result, FailureInvalidInput, "compensable_effect_kind_missing", "host:compensable_effect_kind", "provide_compensable_effect_kind")
	}
	for _, kind := range KnownProductionAdapterCompensableEffectKinds() {
		if _, duplicate := duplicateKinds[kind]; duplicate {
			result = productionAdapterCompensationCapabilityPlanBlock(result, FailureInvalidInput, string(kind)+"_compensation_capability_duplicate", MissingInput("host:"+string(kind)+"_compensation_capability_ref"), "deduplicate_compensation_capability")
		}
		spec, exists := specsByKind[kind]
		if !exists {
			spec = ProductionAdapterCompensationCapabilitySpec{EffectKind: kind}
		}
		declaration := BuildProductionAdapterCompensationCapabilityDeclaration(spec)
		result.Declarations = append(result.Declarations, declaration)
		result.EffectKinds = append(result.EffectKinds, declaration.EffectKind)
		result.EffectRefs = appendDisplaySafeRefIfPresent(result.EffectRefs, declaration.EffectRef)
		result.EffectGateRefs = appendDisplaySafeRefIfPresent(result.EffectGateRefs, declaration.EffectGateRef)
		result.AdapterRefs = appendDisplaySafeRefIfPresent(result.AdapterRefs, declaration.AdapterRef)
		result.CompensationCapabilityRefs = appendDisplaySafeRefIfPresent(result.CompensationCapabilityRefs, declaration.CompensationCapabilityRef)
		result.CompensationRefs = appendDisplaySafeRefIfPresent(result.CompensationRefs, declaration.CompensationRef)
		result.CompensationExecutorRefs = appendDisplaySafeRefIfPresent(result.CompensationExecutorRefs, declaration.CompensationExecutorRef)
		result.CompensationPolicyRefs = appendDisplaySafeRefIfPresent(result.CompensationPolicyRefs, declaration.CompensationPolicyRef)
		result.CompensationApprovalRefs = appendDisplaySafeRefIfPresent(result.CompensationApprovalRefs, declaration.CompensationApprovalRef)
		result.FailureReviewRefs = appendDisplaySafeRefIfPresent(result.FailureReviewRefs, declaration.FailureReviewRef)
		result.ResidualRiskRefs = appendDisplaySafeRefIfPresent(result.ResidualRiskRefs, declaration.ResidualRiskRef)
		result.MissingInputs = AppendMissingInputs(result.MissingInputs, declaration.MissingInputs...)
		result.BlockedReasons = appendUniqueControlTokens(result.BlockedReasons, declaration.BlockedReasons)
		result.Boundaries = AppendBoundaries(result.Boundaries, declaration.Boundaries...)
	}
	allDeclared := len(result.Declarations) == len(KnownProductionAdapterCompensableEffectKinds())
	for _, declaration := range result.Declarations {
		if !declaration.ReadyForCompensationPlan || !declaration.CompensationCapabilityDeclared {
			allDeclared = false
			break
		}
	}
	result.AllRequiredCapabilitiesDeclared = allDeclared
	result.ResidualRiskDeclared = len(result.ResidualRiskRefs) > 0
	if allDeclared {
		policyRefs := normalizeDisplaySafeRefs(append(cloneDisplaySafeRefs(result.RequiredPolicyRefs), result.CompensationPolicyRefs...))
		approvalRefs := normalizeDisplaySafeRefs(append(cloneDisplaySafeRefs(result.RequiredApprovalRefs), result.CompensationApprovalRefs...))
		result.ExecutorDescriptor = BuildObjectiveCompensationExecutorDescriptor(ObjectiveCompensationExecutorDescriptor{
			Available:                 true,
			DescriptorRef:             result.ExecutorDescriptorRef,
			ExecutorRef:               result.ExecutorRef,
			OwnerRef:                  result.OwnerRef,
			SupportedCompensationRefs: result.CompensationRefs,
			IdempotencyContractRef:    result.IdempotencyContractRef,
			ReadbackContractRef:       result.ReadbackContractRef,
			RollbackContractRef:       result.RollbackContractRef,
			TimeoutPolicyRef:          result.TimeoutPolicyRef,
			PolicyRefs:                policyRefs,
			RequiredPolicyRefs:        result.RequiredPolicyRefs,
			ApprovalRefs:              approvalRefs,
			RequiredApprovalRefs:      result.RequiredApprovalRefs,
			Boundaries:                []Boundary{"compensation_capability_plan_executor_descriptor"},
			RunnerEffect:              "none",
			PromptEffect:              "none",
		})
		result.MissingInputs = AppendMissingInputs(result.MissingInputs, result.ExecutorDescriptor.MissingInputs...)
		result.BlockedReasons = appendUniqueControlTokens(result.BlockedReasons, result.ExecutorDescriptor.BlockedReasons)
		result.Boundaries = AppendBoundaries(result.Boundaries, result.ExecutorDescriptor.Boundaries...)
		if result.ExecutorDescriptor.Status == HostActionReady {
			result.ReadyForCompensationExecutorDescriptor = true
		}
	}
	if len(result.MissingInputs) == 0 && len(result.BlockedReasons) == 0 {
		if allDeclared && result.ReadyForCompensationExecutorDescriptor {
			result.Status = HostActionReady
			result.ReadyForCompensationCapabilities = true
			result.NextHostAction = "host_may_bind_compensation_executor_descriptor"
			result.Boundaries = AppendBoundaries(result.Boundaries, "compensation_capability_plan_ready")
		} else if result.ResidualRiskDeclared {
			result.Status = HostActionReviewRequired
			result.NextHostAction = "review_compensation_residual_risk"
			result.Boundaries = AppendBoundaries(result.Boundaries, "compensation_capability_plan_has_residual_risk")
		}
	}
	return result.Normalize()
}

func CloneProductionAdapterCompensationCapabilityDeclaration(in ProductionAdapterCompensationCapabilityDeclaration) ProductionAdapterCompensationCapabilityDeclaration {
	out := in
	out.EvidenceRefs = cloneDisplaySafeRefs(in.EvidenceRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (d ProductionAdapterCompensationCapabilityDeclaration) Clone() ProductionAdapterCompensationCapabilityDeclaration {
	return CloneProductionAdapterCompensationCapabilityDeclaration(d)
}

func (d ProductionAdapterCompensationCapabilityDeclaration) Normalize() ProductionAdapterCompensationCapabilityDeclaration {
	unsafeInput := productionAdapterCompensationCapabilityDeclarationOutputUnsafe(d)
	out := CloneProductionAdapterCompensationCapabilityDeclaration(d)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.EffectKind = NormalizeProductionAdapterCompensableEffectKind(string(out.EffectKind))
	out.EffectRef = normalizeOneDisplaySafeRef(out.EffectRef)
	out.EffectGateRef = normalizeOneDisplaySafeRef(out.EffectGateRef)
	out.AdapterRef = normalizeOneDisplaySafeRef(out.AdapterRef)
	out.CompensationCapabilityRef = normalizeOneDisplaySafeRef(out.CompensationCapabilityRef)
	out.CompensationRef = normalizeOneDisplaySafeRef(out.CompensationRef)
	out.CompensationExecutorRef = normalizeOneDisplaySafeRef(out.CompensationExecutorRef)
	out.CompensationPolicyRef = normalizeOneDisplaySafeRef(out.CompensationPolicyRef)
	out.CompensationApprovalRef = normalizeOneDisplaySafeRef(out.CompensationApprovalRef)
	out.IdempotencyContractRef = normalizeOneDisplaySafeRef(out.IdempotencyContractRef)
	out.ReadbackContractRef = normalizeOneDisplaySafeRef(out.ReadbackContractRef)
	out.RollbackContractRef = normalizeOneDisplaySafeRef(out.RollbackContractRef)
	out.FailureReviewRef = normalizeOneDisplaySafeRef(out.FailureReviewRef)
	out.ResidualRiskRef = normalizeOneDisplaySafeRef(out.ResidualRiskRef)
	out.EvidenceRefs = normalizeDisplaySafeRefs(out.EvidenceRefs)
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
	objectiveCompensationClearCoreEffects(&out.CoreExecutionExecuted, &out.RunnerDispatched, &out.ToolExecuted, &out.WorkflowDispatched, &out.SchedulerApplied, &out.InstallerExecuted, &out.StoreMutationExecuted, &out.CompensationExecutedByCore)
	if unsafeInput || out.RawOutputLoaded {
		out.RawOutputLoaded = true
		out.Status = HostActionBlocked
		out.ReadyForCompensationPlan = false
		out.CompensationCapabilityDeclared = false
		out.ResidualRiskDeclared = false
		if out.FailureClass == FailureNone {
			out.FailureClass = FailureEvidenceWeak
		}
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		out.NextHostAction = "provide_display_safe_refs"
	}
	out.CompensationCapabilityDeclared = out.CompensationCapabilityDeclared &&
		out.Status == HostActionReady &&
		out.EffectKind != "" &&
		out.EffectRef != "" &&
		out.EffectGateRef != "" &&
		out.AdapterRef != "" &&
		out.CompensationCapabilityRef != "" &&
		out.CompensationRef != "" &&
		out.CompensationExecutorRef != "" &&
		out.CompensationPolicyRef != "" &&
		out.CompensationApprovalRef != "" &&
		out.IdempotencyContractRef != "" &&
		out.ReadbackContractRef != "" &&
		out.RollbackContractRef != "" &&
		out.FailureReviewRef != "" &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.RawOutputLoaded
	out.ReadyForCompensationPlan = out.ReadyForCompensationPlan &&
		out.CompensationCapabilityDeclared &&
		out.RunnerEffect == "none" &&
		out.PromptEffect == "none"
	out.ResidualRiskDeclared = out.ResidualRiskDeclared &&
		out.Status == HostActionRecorded &&
		out.ResidualRiskRef != "" &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.RawOutputLoaded
	return out
}

func CloneProductionAdapterCompensationCapabilityPlan(in ProductionAdapterCompensationCapabilityPlan) ProductionAdapterCompensationCapabilityPlan {
	out := in
	out.EffectKinds = cloneProductionAdapterCompensableEffectKinds(in.EffectKinds)
	out.EffectRefs = cloneDisplaySafeRefs(in.EffectRefs)
	out.EffectGateRefs = cloneDisplaySafeRefs(in.EffectGateRefs)
	out.AdapterRefs = cloneDisplaySafeRefs(in.AdapterRefs)
	out.CompensationCapabilityRefs = cloneDisplaySafeRefs(in.CompensationCapabilityRefs)
	out.CompensationRefs = cloneDisplaySafeRefs(in.CompensationRefs)
	out.CompensationExecutorRefs = cloneDisplaySafeRefs(in.CompensationExecutorRefs)
	out.CompensationPolicyRefs = cloneDisplaySafeRefs(in.CompensationPolicyRefs)
	out.CompensationApprovalRefs = cloneDisplaySafeRefs(in.CompensationApprovalRefs)
	out.FailureReviewRefs = cloneDisplaySafeRefs(in.FailureReviewRefs)
	out.ResidualRiskRefs = cloneDisplaySafeRefs(in.ResidualRiskRefs)
	out.RequiredPolicyRefs = cloneDisplaySafeRefs(in.RequiredPolicyRefs)
	out.RequiredApprovalRefs = cloneDisplaySafeRefs(in.RequiredApprovalRefs)
	out.Declarations = cloneProductionAdapterCompensationCapabilityDeclarations(in.Declarations)
	out.ExecutorDescriptor = in.ExecutorDescriptor.Clone()
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (p ProductionAdapterCompensationCapabilityPlan) Clone() ProductionAdapterCompensationCapabilityPlan {
	return CloneProductionAdapterCompensationCapabilityPlan(p)
}

func (p ProductionAdapterCompensationCapabilityPlan) Normalize() ProductionAdapterCompensationCapabilityPlan {
	unsafeInput := productionAdapterCompensationCapabilityPlanOutputUnsafe(p)
	out := CloneProductionAdapterCompensationCapabilityPlan(p)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.PlanRef = normalizeOneDisplaySafeRef(out.PlanRef)
	out.ExecutorDescriptorRef = normalizeOneDisplaySafeRef(out.ExecutorDescriptorRef)
	out.ExecutorRef = normalizeOneDisplaySafeRef(out.ExecutorRef)
	out.OwnerRef = normalizeOneDisplaySafeRef(out.OwnerRef)
	out.IdempotencyContractRef = normalizeOneDisplaySafeRef(out.IdempotencyContractRef)
	out.ReadbackContractRef = normalizeOneDisplaySafeRef(out.ReadbackContractRef)
	out.RollbackContractRef = normalizeOneDisplaySafeRef(out.RollbackContractRef)
	out.TimeoutPolicyRef = normalizeOneDisplaySafeRef(out.TimeoutPolicyRef)
	out.EffectKinds = normalizeProductionAdapterCompensableEffectKinds(out.EffectKinds)
	out.EffectRefs = normalizeDisplaySafeRefs(out.EffectRefs)
	out.EffectGateRefs = normalizeDisplaySafeRefs(out.EffectGateRefs)
	out.AdapterRefs = normalizeDisplaySafeRefs(out.AdapterRefs)
	out.CompensationCapabilityRefs = normalizeDisplaySafeRefs(out.CompensationCapabilityRefs)
	out.CompensationRefs = normalizeDisplaySafeRefs(out.CompensationRefs)
	out.CompensationExecutorRefs = normalizeDisplaySafeRefs(out.CompensationExecutorRefs)
	out.CompensationPolicyRefs = normalizeDisplaySafeRefs(out.CompensationPolicyRefs)
	out.CompensationApprovalRefs = normalizeDisplaySafeRefs(out.CompensationApprovalRefs)
	out.FailureReviewRefs = normalizeDisplaySafeRefs(out.FailureReviewRefs)
	out.ResidualRiskRefs = normalizeDisplaySafeRefs(out.ResidualRiskRefs)
	out.RequiredPolicyRefs = normalizeDisplaySafeRefs(out.RequiredPolicyRefs)
	out.RequiredApprovalRefs = normalizeDisplaySafeRefs(out.RequiredApprovalRefs)
	for i := range out.Declarations {
		out.Declarations[i] = out.Declarations[i].Normalize()
	}
	out.ExecutorDescriptor = out.ExecutorDescriptor.Normalize()
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
	objectiveCompensationClearCoreEffects(&out.CoreExecutionExecuted, &out.RunnerDispatched, &out.ToolExecuted, &out.WorkflowDispatched, &out.SchedulerApplied, &out.InstallerExecuted, &out.StoreMutationExecuted, &out.CompensationExecutedByCore)
	if unsafeInput || out.RawOutputLoaded {
		out.RawOutputLoaded = true
		out.Status = HostActionBlocked
		out.ReadyForCompensationCapabilities = false
		out.ReadyForCompensationExecutorDescriptor = false
		out.AllRequiredCapabilitiesDeclared = false
		out.ResidualRiskDeclared = false
		if out.FailureClass == FailureNone {
			out.FailureClass = FailureEvidenceWeak
		}
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		out.NextHostAction = "provide_display_safe_refs"
	}
	allDeclared := len(out.Declarations) == len(KnownProductionAdapterCompensableEffectKinds())
	for _, declaration := range out.Declarations {
		if !declaration.ReadyForCompensationPlan || !declaration.CompensationCapabilityDeclared {
			allDeclared = false
			break
		}
	}
	out.AllRequiredCapabilitiesDeclared = out.AllRequiredCapabilitiesDeclared && allDeclared
	out.ReadyForCompensationExecutorDescriptor = out.ReadyForCompensationExecutorDescriptor &&
		out.ExecutorDescriptor.Status == HostActionReady &&
		len(out.CompensationRefs) == len(KnownProductionAdapterCompensableEffectKinds()) &&
		!out.RawOutputLoaded
	out.ReadyForCompensationCapabilities = out.ReadyForCompensationCapabilities &&
		out.Status == HostActionReady &&
		out.PlanRef != "" &&
		out.ExecutorDescriptorRef != "" &&
		out.ExecutorRef != "" &&
		out.OwnerRef != "" &&
		out.IdempotencyContractRef != "" &&
		out.ReadbackContractRef != "" &&
		out.RollbackContractRef != "" &&
		out.AllRequiredCapabilitiesDeclared &&
		out.ReadyForCompensationExecutorDescriptor &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.RawOutputLoaded &&
		out.RunnerEffect == "none" &&
		out.PromptEffect == "none"
	out.ResidualRiskDeclared = out.ResidualRiskDeclared && len(out.ResidualRiskRefs) > 0 && !out.RawOutputLoaded
	if !out.ReadyForCompensationCapabilities && out.Status == HostActionReady {
		out.Status = HostActionBlocked
	}
	return out
}

func productionAdapterCompensationCapabilityDeclarationBlock(result ProductionAdapterCompensationCapabilityDeclaration, failure FailureClass, reason string, missing MissingInput, next NextHostAction) ProductionAdapterCompensationCapabilityDeclaration {
	result.Status = HostActionBlocked
	result.ReadyForCompensationPlan = false
	result.CompensationCapabilityDeclared = false
	result.ResidualRiskDeclared = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = firstNextHostAction(next, result.NextHostAction)
	result.Boundaries = AppendBoundaries(result.Boundaries, "compensation_capability_declaration_blocked")
	return result
}

func productionAdapterCompensationCapabilityPlanBlock(result ProductionAdapterCompensationCapabilityPlan, failure FailureClass, reason string, missing MissingInput, next NextHostAction) ProductionAdapterCompensationCapabilityPlan {
	result.Status = HostActionBlocked
	result.ReadyForCompensationCapabilities = false
	result.ReadyForCompensationExecutorDescriptor = false
	result.AllRequiredCapabilitiesDeclared = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = firstNextHostAction(next, result.NextHostAction)
	result.Boundaries = AppendBoundaries(result.Boundaries, "compensation_capability_plan_blocked")
	return result
}

func productionAdapterCompensationCapabilitySpecHasCapability(input ProductionAdapterCompensationCapabilitySpec) bool {
	return input.CompensationCapabilityRef != "" ||
		input.CompensationRef != "" ||
		input.CompensationExecutorRef != "" ||
		input.CompensationPolicyRef != "" ||
		input.CompensationApprovalRef != "" ||
		input.IdempotencyContractRef != "" ||
		input.ReadbackContractRef != "" ||
		input.RollbackContractRef != ""
}

func productionAdapterCompensationCapabilitySpecUnsafe(input ProductionAdapterCompensationCapabilitySpec) bool {
	return input.RawOutputLoaded ||
		displaySafeRefRejected(input.EffectRef) ||
		displaySafeRefRejected(input.EffectGateRef) ||
		displaySafeRefRejected(input.AdapterRef) ||
		displaySafeRefRejected(input.CompensationCapabilityRef) ||
		displaySafeRefRejected(input.CompensationRef) ||
		displaySafeRefRejected(input.CompensationExecutorRef) ||
		displaySafeRefRejected(input.CompensationPolicyRef) ||
		displaySafeRefRejected(input.CompensationApprovalRef) ||
		displaySafeRefRejected(input.IdempotencyContractRef) ||
		displaySafeRefRejected(input.ReadbackContractRef) ||
		displaySafeRefRejected(input.RollbackContractRef) ||
		displaySafeRefRejected(input.FailureReviewRef) ||
		displaySafeRefRejected(input.ResidualRiskRef) ||
		displaySafeRefSliceRejected(input.EvidenceRefs)
}

func productionAdapterCompensationCapabilityPlanInputUnsafe(input ProductionAdapterCompensationCapabilityPlanInput) bool {
	if input.RawOutputLoaded ||
		displaySafeRefRejected(input.PlanRef) ||
		displaySafeRefRejected(input.ExecutorDescriptorRef) ||
		displaySafeRefRejected(input.ExecutorRef) ||
		displaySafeRefRejected(input.OwnerRef) ||
		displaySafeRefRejected(input.IdempotencyContractRef) ||
		displaySafeRefRejected(input.ReadbackContractRef) ||
		displaySafeRefRejected(input.RollbackContractRef) ||
		displaySafeRefRejected(input.TimeoutPolicyRef) ||
		displaySafeRefSliceRejected(input.RequiredPolicyRefs) ||
		displaySafeRefSliceRejected(input.RequiredApprovalRefs) {
		return true
	}
	for _, spec := range input.CapabilitySpecs {
		if productionAdapterCompensationCapabilitySpecUnsafe(spec) {
			return true
		}
	}
	return false
}

func productionAdapterCompensationCapabilityDeclarationOutputUnsafe(input ProductionAdapterCompensationCapabilityDeclaration) bool {
	return displaySafeRefRejected(input.EffectRef) ||
		displaySafeRefRejected(input.EffectGateRef) ||
		displaySafeRefRejected(input.AdapterRef) ||
		displaySafeRefRejected(input.CompensationCapabilityRef) ||
		displaySafeRefRejected(input.CompensationRef) ||
		displaySafeRefRejected(input.CompensationExecutorRef) ||
		displaySafeRefRejected(input.CompensationPolicyRef) ||
		displaySafeRefRejected(input.CompensationApprovalRef) ||
		displaySafeRefRejected(input.IdempotencyContractRef) ||
		displaySafeRefRejected(input.ReadbackContractRef) ||
		displaySafeRefRejected(input.RollbackContractRef) ||
		displaySafeRefRejected(input.FailureReviewRef) ||
		displaySafeRefRejected(input.ResidualRiskRef) ||
		displaySafeRefSliceRejected(input.EvidenceRefs) ||
		input.RawOutputLoaded
}

func productionAdapterCompensationCapabilityPlanOutputUnsafe(input ProductionAdapterCompensationCapabilityPlan) bool {
	if displaySafeRefRejected(input.PlanRef) ||
		displaySafeRefRejected(input.ExecutorDescriptorRef) ||
		displaySafeRefRejected(input.ExecutorRef) ||
		displaySafeRefRejected(input.OwnerRef) ||
		displaySafeRefRejected(input.IdempotencyContractRef) ||
		displaySafeRefRejected(input.ReadbackContractRef) ||
		displaySafeRefRejected(input.RollbackContractRef) ||
		displaySafeRefRejected(input.TimeoutPolicyRef) ||
		displaySafeRefSliceRejected(input.EffectRefs) ||
		displaySafeRefSliceRejected(input.EffectGateRefs) ||
		displaySafeRefSliceRejected(input.AdapterRefs) ||
		displaySafeRefSliceRejected(input.CompensationCapabilityRefs) ||
		displaySafeRefSliceRejected(input.CompensationRefs) ||
		displaySafeRefSliceRejected(input.CompensationExecutorRefs) ||
		displaySafeRefSliceRejected(input.CompensationPolicyRefs) ||
		displaySafeRefSliceRejected(input.CompensationApprovalRefs) ||
		displaySafeRefSliceRejected(input.FailureReviewRefs) ||
		displaySafeRefSliceRejected(input.ResidualRiskRefs) ||
		displaySafeRefSliceRejected(input.RequiredPolicyRefs) ||
		displaySafeRefSliceRejected(input.RequiredApprovalRefs) ||
		objectiveCompensationExecutorDescriptorUnsafe(input.ExecutorDescriptor) ||
		input.RawOutputLoaded {
		return true
	}
	for _, declaration := range input.Declarations {
		if productionAdapterCompensationCapabilityDeclarationOutputUnsafe(declaration) {
			return true
		}
	}
	return false
}

func normalizeProductionAdapterCompensableEffectKinds(in []ProductionAdapterCompensableEffectKind) []ProductionAdapterCompensableEffectKind {
	out := make([]ProductionAdapterCompensableEffectKind, 0, len(in))
	seen := map[ProductionAdapterCompensableEffectKind]struct{}{}
	for _, value := range in {
		kind := NormalizeProductionAdapterCompensableEffectKind(string(value))
		if kind == "" {
			continue
		}
		if _, exists := seen[kind]; exists {
			continue
		}
		seen[kind] = struct{}{}
		out = append(out, kind)
	}
	return out
}

func cloneProductionAdapterCompensableEffectKinds(in []ProductionAdapterCompensableEffectKind) []ProductionAdapterCompensableEffectKind {
	out := make([]ProductionAdapterCompensableEffectKind, 0, len(in))
	for _, value := range in {
		out = append(out, value)
	}
	return out
}

func cloneProductionAdapterCompensationCapabilityDeclarations(in []ProductionAdapterCompensationCapabilityDeclaration) []ProductionAdapterCompensationCapabilityDeclaration {
	out := make([]ProductionAdapterCompensationCapabilityDeclaration, 0, len(in))
	for _, declaration := range in {
		out = append(out, declaration.Clone())
	}
	return out
}
