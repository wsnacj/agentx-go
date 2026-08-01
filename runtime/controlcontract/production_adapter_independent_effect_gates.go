package controlcontract

type ProductionAdapterEffectGateKind string

const (
	ProductionAdapterEffectGateSchedulerApply       ProductionAdapterEffectGateKind = "scheduler_apply"
	ProductionAdapterEffectGateInstallerApply       ProductionAdapterEffectGateKind = "installer_apply"
	ProductionAdapterEffectGateWorkflowRetryApply   ProductionAdapterEffectGateKind = "workflow_retry_apply"
	ProductionAdapterEffectGateRuntimeExecutor      ProductionAdapterEffectGateKind = "runtime_executor"
	ProductionAdapterEffectGateDelegationWorker     ProductionAdapterEffectGateKind = "delegation_worker_runtime"
	ProductionAdapterEffectGateMemoryApply          ProductionAdapterEffectGateKind = "memory_apply"
	ProductionAdapterEffectGateCompensationExecutor ProductionAdapterEffectGateKind = "compensation_executor"
)

func KnownProductionAdapterEffectGateKinds() []ProductionAdapterEffectGateKind {
	return []ProductionAdapterEffectGateKind{
		ProductionAdapterEffectGateSchedulerApply,
		ProductionAdapterEffectGateInstallerApply,
		ProductionAdapterEffectGateWorkflowRetryApply,
		ProductionAdapterEffectGateRuntimeExecutor,
		ProductionAdapterEffectGateDelegationWorker,
		ProductionAdapterEffectGateMemoryApply,
		ProductionAdapterEffectGateCompensationExecutor,
	}
}

func NormalizeProductionAdapterEffectGateKind(raw string) ProductionAdapterEffectGateKind {
	switch normalizeEnumToken(raw) {
	case "scheduler_apply", "schedule_apply", "operations_schedule_apply":
		return ProductionAdapterEffectGateSchedulerApply
	case "installer_apply", "install_apply", "capability_install_apply":
		return ProductionAdapterEffectGateInstallerApply
	case "workflow_retry_apply", "workflow_retry", "retry_workflow":
		return ProductionAdapterEffectGateWorkflowRetryApply
	case "runtime_executor", "runner_executor", "executor":
		return ProductionAdapterEffectGateRuntimeExecutor
	case "delegation_worker_runtime", "delegation_worker", "worker_runtime", "worker_dispatch":
		return ProductionAdapterEffectGateDelegationWorker
	case "memory_apply", "memory_proposal_apply", "skill_memory_apply", "proposal_memory_apply":
		return ProductionAdapterEffectGateMemoryApply
	case "compensation_executor", "compensation_apply", "rollback_executor":
		return ProductionAdapterEffectGateCompensationExecutor
	default:
		return ""
	}
}

type ProductionAdapterIndependentEffectGateSpec struct {
	Kind                  ProductionAdapterEffectGateKind `json:"kind,omitempty"`
	GateRef               DisplaySafeRef                  `json:"gate_ref,omitempty"`
	AdapterRef            DisplaySafeRef                  `json:"adapter_ref,omitempty"`
	ContractRef           DisplaySafeRef                  `json:"contract_ref,omitempty"`
	PolicyRef             DisplaySafeRef                  `json:"policy_ref,omitempty"`
	ApprovalRef           DisplaySafeRef                  `json:"approval_ref,omitempty"`
	BudgetRef             DisplaySafeRef                  `json:"budget_ref,omitempty"`
	IdempotencyRef        DisplaySafeRef                  `json:"idempotency_ref,omitempty"`
	ReadbackRef           DisplaySafeRef                  `json:"readback_ref,omitempty"`
	EvalRef               DisplaySafeRef                  `json:"eval_ref,omitempty"`
	FailureReviewRef      DisplaySafeRef                  `json:"failure_review_ref,omitempty"`
	CompensationReviewRef DisplaySafeRef                  `json:"compensation_review_ref,omitempty"`
	EvidenceRefs          []DisplaySafeRef                `json:"evidence_refs,omitempty"`
	Boundaries            []Boundary                      `json:"boundaries,omitempty"`
	RawOutputLoaded       bool                            `json:"raw_output_loaded"`
}

type ProductionAdapterIndependentEffectGatePlanInput struct {
	PlanRef                      DisplaySafeRef                               `json:"plan_ref,omitempty"`
	GateSpecs                    []ProductionAdapterIndependentEffectGateSpec `json:"gate_specs,omitempty"`
	AggregateExecutorRequested   bool                                         `json:"aggregate_executor_requested"`
	AggregateExecutorRef         DisplaySafeRef                               `json:"aggregate_executor_ref,omitempty"`
	AggregateExecutorPolicyRef   DisplaySafeRef                               `json:"aggregate_executor_policy_ref,omitempty"`
	AggregateExecutorApprovalRef DisplaySafeRef                               `json:"aggregate_executor_approval_ref,omitempty"`
	RawOutputLoaded              bool                                         `json:"raw_output_loaded"`
}

type ProductionAdapterIndependentEffectGatePlan struct {
	ContractVersion             string                                   `json:"contract_version,omitempty"`
	Projected                   bool                                     `json:"projected"`
	Status                      HostActionStatus                         `json:"status,omitempty"`
	ReadyForIndependentGatePlan bool                                     `json:"ready_for_independent_gate_plan"`
	AggregateExecutorBlocked    bool                                     `json:"aggregate_executor_blocked"`
	PlanRef                     DisplaySafeRef                           `json:"plan_ref,omitempty"`
	AggregateExecutorRef        DisplaySafeRef                           `json:"aggregate_executor_ref,omitempty"`
	GateKinds                   []ProductionAdapterEffectGateKind        `json:"gate_kinds,omitempty"`
	GateRefs                    []DisplaySafeRef                         `json:"gate_refs,omitempty"`
	AdapterRefs                 []DisplaySafeRef                         `json:"adapter_refs,omitempty"`
	ContractRefs                []DisplaySafeRef                         `json:"contract_refs,omitempty"`
	PolicyRefs                  []DisplaySafeRef                         `json:"policy_refs,omitempty"`
	ApprovalRefs                []DisplaySafeRef                         `json:"approval_refs,omitempty"`
	BudgetRefs                  []DisplaySafeRef                         `json:"budget_refs,omitempty"`
	IdempotencyRefs             []DisplaySafeRef                         `json:"idempotency_refs,omitempty"`
	ReadbackRefs                []DisplaySafeRef                         `json:"readback_refs,omitempty"`
	EvalRefs                    []DisplaySafeRef                         `json:"eval_refs,omitempty"`
	FailureReviewRefs           []DisplaySafeRef                         `json:"failure_review_refs,omitempty"`
	CompensationReviewRefs      []DisplaySafeRef                         `json:"compensation_review_refs,omitempty"`
	Gates                       []ProductionAdapterIndependentEffectGate `json:"gates,omitempty"`
	MissingInputs               []MissingInput                           `json:"missing_inputs,omitempty"`
	BlockedReasons              []string                                 `json:"blocked_reasons,omitempty"`
	FailureClass                FailureClass                             `json:"failure_class,omitempty"`
	Boundaries                  []Boundary                               `json:"boundaries,omitempty"`
	NextHostAction              NextHostAction                           `json:"next_host_action,omitempty"`
	RunnerEffect                string                                   `json:"runner_effect,omitempty"`
	PromptEffect                string                                   `json:"prompt_effect,omitempty"`
	RawOutputLoaded             bool                                     `json:"raw_output_loaded"`
}

type ProductionAdapterIndependentEffectGate struct {
	ContractVersion             string                          `json:"contract_version,omitempty"`
	Projected                   bool                            `json:"projected"`
	Status                      HostActionStatus                `json:"status,omitempty"`
	Kind                        ProductionAdapterEffectGateKind `json:"kind,omitempty"`
	ReadyForIndependentGatePlan bool                            `json:"ready_for_independent_gate_plan"`
	GateRef                     DisplaySafeRef                  `json:"gate_ref,omitempty"`
	AdapterRef                  DisplaySafeRef                  `json:"adapter_ref,omitempty"`
	ContractRef                 DisplaySafeRef                  `json:"contract_ref,omitempty"`
	PolicyRef                   DisplaySafeRef                  `json:"policy_ref,omitempty"`
	ApprovalRef                 DisplaySafeRef                  `json:"approval_ref,omitempty"`
	BudgetRef                   DisplaySafeRef                  `json:"budget_ref,omitempty"`
	IdempotencyRef              DisplaySafeRef                  `json:"idempotency_ref,omitempty"`
	ReadbackRef                 DisplaySafeRef                  `json:"readback_ref,omitempty"`
	EvalRef                     DisplaySafeRef                  `json:"eval_ref,omitempty"`
	FailureReviewRef            DisplaySafeRef                  `json:"failure_review_ref,omitempty"`
	CompensationReviewRef       DisplaySafeRef                  `json:"compensation_review_ref,omitempty"`
	EvidenceRefs                []DisplaySafeRef                `json:"evidence_refs,omitempty"`
	MissingInputs               []MissingInput                  `json:"missing_inputs,omitempty"`
	BlockedReasons              []string                        `json:"blocked_reasons,omitempty"`
	FailureClass                FailureClass                    `json:"failure_class,omitempty"`
	Boundaries                  []Boundary                      `json:"boundaries,omitempty"`
	NextHostAction              NextHostAction                  `json:"next_host_action,omitempty"`
	RunnerEffect                string                          `json:"runner_effect,omitempty"`
	PromptEffect                string                          `json:"prompt_effect,omitempty"`
	RawOutputLoaded             bool                            `json:"raw_output_loaded"`
}

func BuildProductionAdapterIndependentEffectGatePlan(input ProductionAdapterIndependentEffectGatePlanInput) ProductionAdapterIndependentEffectGatePlan {
	specsByKind := map[ProductionAdapterEffectGateKind]ProductionAdapterIndependentEffectGateSpec{}
	duplicateKinds := map[ProductionAdapterEffectGateKind]struct{}{}
	invalidSpecCount := 0
	for _, spec := range input.GateSpecs {
		kind := NormalizeProductionAdapterEffectGateKind(string(spec.Kind))
		if kind == "" {
			invalidSpecCount++
			continue
		}
		spec.Kind = kind
		if _, exists := specsByKind[kind]; exists {
			duplicateKinds[kind] = struct{}{}
			continue
		}
		specsByKind[kind] = spec
	}
	result := ProductionAdapterIndependentEffectGatePlan{
		ContractVersion:      ContractVersion,
		Projected:            true,
		Status:               HostActionBlocked,
		PlanRef:              normalizeOneDisplaySafeRef(input.PlanRef),
		AggregateExecutorRef: normalizeOneDisplaySafeRef(input.AggregateExecutorRef),
		FailureClass:         FailureNone,
		Boundaries: []Boundary{
			"production_adapter_independent_effect_gate_plan",
			"effect_gate_plan_projection_only",
			"host_owned_effect_gates",
			"display_safe_refs_only",
			"no_runner_dispatch",
			"no_unified_auto_executor",
			"no_scheduler_apply",
			"no_installer_apply",
			"no_workflow_retry_apply",
			"no_runtime_executor",
			"no_delegation_worker_runtime",
			"no_memory_apply",
			"no_compensation_executor",
		},
		NextHostAction:  "review_independent_effect_gate_plan",
		RunnerEffect:    "none",
		PromptEffect:    "none",
		RawOutputLoaded: input.RawOutputLoaded,
	}
	if productionAdapterIndependentEffectGatePlanUnsafe(input) {
		result = productionAdapterIndependentEffectGatePlanBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		result.Boundaries = AppendBoundaries(result.Boundaries, "raw_output_not_allowed")
	}
	if result.PlanRef == "" {
		result = productionAdapterIndependentEffectGatePlanBlock(result, FailureEvidenceMissing, "effect_gate_plan_ref_missing", "host:effect_gate_plan_ref", "provide_independent_effect_gate_plan")
	}
	if input.AggregateExecutorRequested || input.AggregateExecutorRef != "" || input.AggregateExecutorPolicyRef != "" || input.AggregateExecutorApprovalRef != "" {
		result.AggregateExecutorBlocked = true
		result = productionAdapterIndependentEffectGatePlanBlock(result, FailurePolicyBlocked, "aggregate_effect_executor_not_allowed", "host:independent_effect_gate", "split_effect_gate_by_kind")
		result.Boundaries = AppendBoundaries(result.Boundaries, "aggregate_effect_executor_blocked")
	}
	if invalidSpecCount > 0 {
		result = productionAdapterIndependentEffectGatePlanBlock(result, FailureInvalidInput, "effect_gate_kind_missing", "host:effect_gate_kind", "provide_independent_effect_gate_kind")
	}
	for _, kind := range KnownProductionAdapterEffectGateKinds() {
		if _, duplicate := duplicateKinds[kind]; duplicate {
			result = productionAdapterIndependentEffectGatePlanBlock(result, FailureInvalidInput, string(kind)+"_gate_duplicate", MissingInput("host:"+string(kind)+"_gate_ref"), "deduplicate_independent_effect_gate")
		}
		spec, exists := specsByKind[kind]
		if !exists {
			spec = ProductionAdapterIndependentEffectGateSpec{Kind: kind}
		}
		gate := BuildProductionAdapterIndependentEffectGate(spec)
		result.Gates = append(result.Gates, gate)
		result.GateKinds = append(result.GateKinds, gate.Kind)
		result.GateRefs = appendDisplaySafeRefIfPresent(result.GateRefs, gate.GateRef)
		result.AdapterRefs = appendDisplaySafeRefIfPresent(result.AdapterRefs, gate.AdapterRef)
		result.ContractRefs = appendDisplaySafeRefIfPresent(result.ContractRefs, gate.ContractRef)
		result.PolicyRefs = appendDisplaySafeRefIfPresent(result.PolicyRefs, gate.PolicyRef)
		result.ApprovalRefs = appendDisplaySafeRefIfPresent(result.ApprovalRefs, gate.ApprovalRef)
		result.BudgetRefs = appendDisplaySafeRefIfPresent(result.BudgetRefs, gate.BudgetRef)
		result.IdempotencyRefs = appendDisplaySafeRefIfPresent(result.IdempotencyRefs, gate.IdempotencyRef)
		result.ReadbackRefs = appendDisplaySafeRefIfPresent(result.ReadbackRefs, gate.ReadbackRef)
		result.EvalRefs = appendDisplaySafeRefIfPresent(result.EvalRefs, gate.EvalRef)
		result.FailureReviewRefs = appendDisplaySafeRefIfPresent(result.FailureReviewRefs, gate.FailureReviewRef)
		result.CompensationReviewRefs = appendDisplaySafeRefIfPresent(result.CompensationReviewRefs, gate.CompensationReviewRef)
		result.MissingInputs = AppendMissingInputs(result.MissingInputs, gate.MissingInputs...)
		result.BlockedReasons = appendUniqueControlTokens(result.BlockedReasons, gate.BlockedReasons)
		result.Boundaries = AppendBoundaries(result.Boundaries, gate.Boundaries...)
	}
	if len(result.MissingInputs) == 0 && len(result.BlockedReasons) == 0 {
		result.Status = HostActionReady
		result.ReadyForIndependentGatePlan = true
		result.NextHostAction = "host_may_implement_independent_effect_gates"
		result.Boundaries = AppendBoundaries(result.Boundaries, "independent_effect_gate_plan_ready")
	}
	return result.Normalize()
}

func BuildProductionAdapterIndependentEffectGate(spec ProductionAdapterIndependentEffectGateSpec) ProductionAdapterIndependentEffectGate {
	kind := NormalizeProductionAdapterEffectGateKind(string(spec.Kind))
	result := ProductionAdapterIndependentEffectGate{
		ContractVersion:       ContractVersion,
		Projected:             true,
		Status:                HostActionBlocked,
		Kind:                  kind,
		GateRef:               normalizeOneDisplaySafeRef(spec.GateRef),
		AdapterRef:            normalizeOneDisplaySafeRef(spec.AdapterRef),
		ContractRef:           normalizeOneDisplaySafeRef(spec.ContractRef),
		PolicyRef:             normalizeOneDisplaySafeRef(spec.PolicyRef),
		ApprovalRef:           normalizeOneDisplaySafeRef(spec.ApprovalRef),
		BudgetRef:             normalizeOneDisplaySafeRef(spec.BudgetRef),
		IdempotencyRef:        normalizeOneDisplaySafeRef(spec.IdempotencyRef),
		ReadbackRef:           normalizeOneDisplaySafeRef(spec.ReadbackRef),
		EvalRef:               normalizeOneDisplaySafeRef(spec.EvalRef),
		FailureReviewRef:      normalizeOneDisplaySafeRef(spec.FailureReviewRef),
		CompensationReviewRef: normalizeOneDisplaySafeRef(spec.CompensationReviewRef),
		EvidenceRefs:          normalizeDisplaySafeRefs(spec.EvidenceRefs),
		FailureClass:          FailureNone,
		Boundaries: []Boundary{
			"production_adapter_independent_effect_gate",
			"effect_gate_projection_only",
			"host_owned_effect_gate",
			"display_safe_refs_only",
			"no_runner_dispatch",
			"no_unified_auto_executor",
			"no_scheduler_apply",
			"no_installer_apply",
			"no_workflow_retry_apply",
			"no_runtime_executor",
			"no_delegation_worker_runtime",
			"no_memory_apply",
			"no_compensation_executor",
		},
		NextHostAction:  productionAdapterIndependentEffectGateNextHostAction(kind),
		RunnerEffect:    "none",
		PromptEffect:    "none",
		RawOutputLoaded: spec.RawOutputLoaded,
	}
	result.Boundaries = AppendBoundaries(result.Boundaries, spec.Boundaries...)
	if kind == "" {
		result = productionAdapterIndependentEffectGateBlock(result, FailureInvalidInput, "effect_gate_kind_missing", "host:effect_gate_kind", "provide_independent_effect_gate_kind")
		return result.Normalize()
	}
	if productionAdapterIndependentEffectGateUnsafe(spec) {
		result = productionAdapterIndependentEffectGateBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		result.Boundaries = AppendBoundaries(result.Boundaries, "raw_output_not_allowed")
		return result.Normalize()
	}
	for _, check := range []struct {
		ref     DisplaySafeRef
		suffix  string
		failure FailureClass
	}{
		{result.GateRef, "gate_ref", FailureEvidenceMissing},
		{result.AdapterRef, "adapter_ref", FailureHostAdapterMissing},
		{result.ContractRef, "contract_ref", FailureConfigMissing},
		{result.PolicyRef, "policy_ref", FailurePolicyBlocked},
		{result.ApprovalRef, "approval_ref", FailureApprovalRequired},
		{result.BudgetRef, "budget_ref", FailureConfigMissing},
		{result.IdempotencyRef, "idempotency_ref", FailureConfigMissing},
		{result.ReadbackRef, "readback_ref", FailureConfigMissing},
		{result.EvalRef, "eval_ref", FailureConfigMissing},
		{result.FailureReviewRef, "failure_review_ref", FailureConfigMissing},
		{result.CompensationReviewRef, "compensation_review_ref", FailureConfigMissing},
	} {
		if check.ref == "" {
			result = productionAdapterIndependentEffectGateBlock(result, check.failure, string(kind)+"_"+check.suffix+"_missing", MissingInput("host:"+string(kind)+"_"+check.suffix), NextHostAction("provide_"+string(kind)+"_"+check.suffix))
		}
	}
	if len(result.MissingInputs) == 0 && len(result.BlockedReasons) == 0 {
		result.Status = HostActionReady
		result.ReadyForIndependentGatePlan = true
		result.NextHostAction = NextHostAction("host_may_implement_" + string(kind) + "_gate")
		result.Boundaries = AppendBoundaries(result.Boundaries, Boundary(string(kind)+"_independent_gate_ready"))
	}
	return result.Normalize()
}

func CloneProductionAdapterIndependentEffectGatePlan(in ProductionAdapterIndependentEffectGatePlan) ProductionAdapterIndependentEffectGatePlan {
	out := in
	out.GateKinds = append([]ProductionAdapterEffectGateKind(nil), in.GateKinds...)
	out.GateRefs = cloneDisplaySafeRefs(in.GateRefs)
	out.AdapterRefs = cloneDisplaySafeRefs(in.AdapterRefs)
	out.ContractRefs = cloneDisplaySafeRefs(in.ContractRefs)
	out.PolicyRefs = cloneDisplaySafeRefs(in.PolicyRefs)
	out.ApprovalRefs = cloneDisplaySafeRefs(in.ApprovalRefs)
	out.BudgetRefs = cloneDisplaySafeRefs(in.BudgetRefs)
	out.IdempotencyRefs = cloneDisplaySafeRefs(in.IdempotencyRefs)
	out.ReadbackRefs = cloneDisplaySafeRefs(in.ReadbackRefs)
	out.EvalRefs = cloneDisplaySafeRefs(in.EvalRefs)
	out.FailureReviewRefs = cloneDisplaySafeRefs(in.FailureReviewRefs)
	out.CompensationReviewRefs = cloneDisplaySafeRefs(in.CompensationReviewRefs)
	out.Gates = cloneProductionAdapterIndependentEffectGates(in.Gates)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (p ProductionAdapterIndependentEffectGatePlan) Clone() ProductionAdapterIndependentEffectGatePlan {
	return CloneProductionAdapterIndependentEffectGatePlan(p)
}

func (p ProductionAdapterIndependentEffectGatePlan) Normalize() ProductionAdapterIndependentEffectGatePlan {
	out := CloneProductionAdapterIndependentEffectGatePlan(p)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.PlanRef = normalizeOneDisplaySafeRef(out.PlanRef)
	out.AggregateExecutorRef = normalizeOneDisplaySafeRef(out.AggregateExecutorRef)
	out.GateKinds = normalizeProductionAdapterEffectGateKinds(out.GateKinds)
	out.GateRefs = normalizeDisplaySafeRefs(out.GateRefs)
	out.AdapterRefs = normalizeDisplaySafeRefs(out.AdapterRefs)
	out.ContractRefs = normalizeDisplaySafeRefs(out.ContractRefs)
	out.PolicyRefs = normalizeDisplaySafeRefs(out.PolicyRefs)
	out.ApprovalRefs = normalizeDisplaySafeRefs(out.ApprovalRefs)
	out.BudgetRefs = normalizeDisplaySafeRefs(out.BudgetRefs)
	out.IdempotencyRefs = normalizeDisplaySafeRefs(out.IdempotencyRefs)
	out.ReadbackRefs = normalizeDisplaySafeRefs(out.ReadbackRefs)
	out.EvalRefs = normalizeDisplaySafeRefs(out.EvalRefs)
	out.FailureReviewRefs = normalizeDisplaySafeRefs(out.FailureReviewRefs)
	out.CompensationReviewRefs = normalizeDisplaySafeRefs(out.CompensationReviewRefs)
	for i := range out.Gates {
		out.Gates[i] = out.Gates[i].Normalize()
	}
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.BlockedReasons = normalizeControlTokenList(out.BlockedReasons)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.Boundaries = MergeBoundaries([]Boundary{
		"production_adapter_independent_effect_gate_plan",
		"effect_gate_plan_projection_only",
		"host_owned_effect_gates",
		"display_safe_refs_only",
		"no_runner_dispatch",
		"no_unified_auto_executor",
		"no_scheduler_apply",
		"no_installer_apply",
		"no_workflow_retry_apply",
		"no_runtime_executor",
		"no_delegation_worker_runtime",
		"no_memory_apply",
		"no_compensation_executor",
	}, out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	if out.RunnerEffect == "" {
		out.RunnerEffect = "none"
	}
	if out.PromptEffect == "" {
		out.PromptEffect = "none"
	}
	if out.RawOutputLoaded || productionAdapterIndependentEffectGatePlanOutputUnsafe(out) {
		out.RawOutputLoaded = true
		out.Status = HostActionBlocked
		out.ReadyForIndependentGatePlan = false
		out.FailureClass = firstFailureClass(out.FailureClass, FailureEvidenceWeak)
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		out.NextHostAction = "provide_display_safe_refs"
	}
	allRowsReady := len(out.Gates) == len(KnownProductionAdapterEffectGateKinds())
	for _, gate := range out.Gates {
		if !gate.ReadyForIndependentGatePlan {
			allRowsReady = false
			break
		}
	}
	out.ReadyForIndependentGatePlan = out.ReadyForIndependentGatePlan &&
		out.Status == HostActionReady &&
		out.PlanRef != "" &&
		allRowsReady &&
		!out.AggregateExecutorBlocked &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.RawOutputLoaded &&
		out.RunnerEffect == "none" &&
		out.PromptEffect == "none"
	if !out.ReadyForIndependentGatePlan && out.Status == HostActionReady {
		out.Status = HostActionBlocked
	}
	if out.NextHostAction == "" {
		out.NextHostAction = "review_independent_effect_gate_plan"
	}
	return out
}

func CloneProductionAdapterIndependentEffectGate(in ProductionAdapterIndependentEffectGate) ProductionAdapterIndependentEffectGate {
	out := in
	out.EvidenceRefs = cloneDisplaySafeRefs(in.EvidenceRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (g ProductionAdapterIndependentEffectGate) Clone() ProductionAdapterIndependentEffectGate {
	return CloneProductionAdapterIndependentEffectGate(g)
}

func (g ProductionAdapterIndependentEffectGate) Normalize() ProductionAdapterIndependentEffectGate {
	out := CloneProductionAdapterIndependentEffectGate(g)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.Kind = NormalizeProductionAdapterEffectGateKind(string(out.Kind))
	out.GateRef = normalizeOneDisplaySafeRef(out.GateRef)
	out.AdapterRef = normalizeOneDisplaySafeRef(out.AdapterRef)
	out.ContractRef = normalizeOneDisplaySafeRef(out.ContractRef)
	out.PolicyRef = normalizeOneDisplaySafeRef(out.PolicyRef)
	out.ApprovalRef = normalizeOneDisplaySafeRef(out.ApprovalRef)
	out.BudgetRef = normalizeOneDisplaySafeRef(out.BudgetRef)
	out.IdempotencyRef = normalizeOneDisplaySafeRef(out.IdempotencyRef)
	out.ReadbackRef = normalizeOneDisplaySafeRef(out.ReadbackRef)
	out.EvalRef = normalizeOneDisplaySafeRef(out.EvalRef)
	out.FailureReviewRef = normalizeOneDisplaySafeRef(out.FailureReviewRef)
	out.CompensationReviewRef = normalizeOneDisplaySafeRef(out.CompensationReviewRef)
	out.EvidenceRefs = normalizeDisplaySafeRefs(out.EvidenceRefs)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.BlockedReasons = normalizeControlTokenList(out.BlockedReasons)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.Boundaries = MergeBoundaries([]Boundary{
		"production_adapter_independent_effect_gate",
		"effect_gate_projection_only",
		"host_owned_effect_gate",
		"display_safe_refs_only",
		"no_runner_dispatch",
		"no_unified_auto_executor",
		"no_scheduler_apply",
		"no_installer_apply",
		"no_workflow_retry_apply",
		"no_runtime_executor",
		"no_delegation_worker_runtime",
		"no_memory_apply",
		"no_compensation_executor",
	}, out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	if out.RunnerEffect == "" {
		out.RunnerEffect = "none"
	}
	if out.PromptEffect == "" {
		out.PromptEffect = "none"
	}
	if out.RawOutputLoaded || productionAdapterIndependentEffectGateOutputUnsafe(out) {
		out.RawOutputLoaded = true
		out.Status = HostActionBlocked
		out.ReadyForIndependentGatePlan = false
		out.FailureClass = firstFailureClass(out.FailureClass, FailureEvidenceWeak)
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		out.NextHostAction = "provide_display_safe_refs"
	}
	out.ReadyForIndependentGatePlan = out.ReadyForIndependentGatePlan &&
		out.Status == HostActionReady &&
		out.Kind != "" &&
		out.GateRef != "" &&
		out.AdapterRef != "" &&
		out.ContractRef != "" &&
		out.PolicyRef != "" &&
		out.ApprovalRef != "" &&
		out.BudgetRef != "" &&
		out.IdempotencyRef != "" &&
		out.ReadbackRef != "" &&
		out.EvalRef != "" &&
		out.FailureReviewRef != "" &&
		out.CompensationReviewRef != "" &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.RawOutputLoaded &&
		out.RunnerEffect == "none" &&
		out.PromptEffect == "none"
	if !out.ReadyForIndependentGatePlan && out.Status == HostActionReady {
		out.Status = HostActionBlocked
	}
	if out.NextHostAction == "" {
		out.NextHostAction = "review_independent_effect_gate"
	}
	return out
}

func productionAdapterIndependentEffectGatePlanBlock(result ProductionAdapterIndependentEffectGatePlan, failure FailureClass, reason string, missing MissingInput, next NextHostAction) ProductionAdapterIndependentEffectGatePlan {
	result.Status = HostActionBlocked
	result.ReadyForIndependentGatePlan = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.Boundaries = AppendBoundaries(result.Boundaries, "independent_effect_gate_plan_blocked")
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	return result
}

func productionAdapterIndependentEffectGateBlock(result ProductionAdapterIndependentEffectGate, failure FailureClass, reason string, missing MissingInput, next NextHostAction) ProductionAdapterIndependentEffectGate {
	result.Status = HostActionBlocked
	result.ReadyForIndependentGatePlan = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.Boundaries = AppendBoundaries(result.Boundaries, "independent_effect_gate_blocked")
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	return result
}

func productionAdapterIndependentEffectGatePlanUnsafe(input ProductionAdapterIndependentEffectGatePlanInput) bool {
	if input.RawOutputLoaded ||
		displaySafeRefRejected(input.PlanRef) ||
		displaySafeRefRejected(input.AggregateExecutorRef) ||
		displaySafeRefRejected(input.AggregateExecutorPolicyRef) ||
		displaySafeRefRejected(input.AggregateExecutorApprovalRef) {
		return true
	}
	for _, spec := range input.GateSpecs {
		if productionAdapterIndependentEffectGateUnsafe(spec) {
			return true
		}
	}
	return false
}

func productionAdapterIndependentEffectGateUnsafe(input ProductionAdapterIndependentEffectGateSpec) bool {
	return input.RawOutputLoaded ||
		displaySafeRefRejected(input.GateRef) ||
		displaySafeRefRejected(input.AdapterRef) ||
		displaySafeRefRejected(input.ContractRef) ||
		displaySafeRefRejected(input.PolicyRef) ||
		displaySafeRefRejected(input.ApprovalRef) ||
		displaySafeRefRejected(input.BudgetRef) ||
		displaySafeRefRejected(input.IdempotencyRef) ||
		displaySafeRefRejected(input.ReadbackRef) ||
		displaySafeRefRejected(input.EvalRef) ||
		displaySafeRefRejected(input.FailureReviewRef) ||
		displaySafeRefRejected(input.CompensationReviewRef) ||
		displaySafeRefSliceRejected(input.EvidenceRefs)
}

func productionAdapterIndependentEffectGatePlanOutputUnsafe(input ProductionAdapterIndependentEffectGatePlan) bool {
	if displaySafeRefRejected(input.PlanRef) ||
		displaySafeRefRejected(input.AggregateExecutorRef) ||
		displaySafeRefSliceRejected(input.GateRefs) ||
		displaySafeRefSliceRejected(input.AdapterRefs) ||
		displaySafeRefSliceRejected(input.ContractRefs) ||
		displaySafeRefSliceRejected(input.PolicyRefs) ||
		displaySafeRefSliceRejected(input.ApprovalRefs) ||
		displaySafeRefSliceRejected(input.BudgetRefs) ||
		displaySafeRefSliceRejected(input.IdempotencyRefs) ||
		displaySafeRefSliceRejected(input.ReadbackRefs) ||
		displaySafeRefSliceRejected(input.EvalRefs) ||
		displaySafeRefSliceRejected(input.FailureReviewRefs) ||
		displaySafeRefSliceRejected(input.CompensationReviewRefs) ||
		input.RawOutputLoaded {
		return true
	}
	for _, gate := range input.Gates {
		if productionAdapterIndependentEffectGateOutputUnsafe(gate) {
			return true
		}
	}
	return false
}

func productionAdapterIndependentEffectGateOutputUnsafe(input ProductionAdapterIndependentEffectGate) bool {
	return displaySafeRefRejected(input.GateRef) ||
		displaySafeRefRejected(input.AdapterRef) ||
		displaySafeRefRejected(input.ContractRef) ||
		displaySafeRefRejected(input.PolicyRef) ||
		displaySafeRefRejected(input.ApprovalRef) ||
		displaySafeRefRejected(input.BudgetRef) ||
		displaySafeRefRejected(input.IdempotencyRef) ||
		displaySafeRefRejected(input.ReadbackRef) ||
		displaySafeRefRejected(input.EvalRef) ||
		displaySafeRefRejected(input.FailureReviewRef) ||
		displaySafeRefRejected(input.CompensationReviewRef) ||
		displaySafeRefSliceRejected(input.EvidenceRefs) ||
		input.RawOutputLoaded
}

func normalizeProductionAdapterEffectGateKinds(in []ProductionAdapterEffectGateKind) []ProductionAdapterEffectGateKind {
	out := make([]ProductionAdapterEffectGateKind, 0, len(in))
	seen := map[ProductionAdapterEffectGateKind]struct{}{}
	for _, value := range in {
		kind := NormalizeProductionAdapterEffectGateKind(string(value))
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

func cloneProductionAdapterIndependentEffectGates(in []ProductionAdapterIndependentEffectGate) []ProductionAdapterIndependentEffectGate {
	out := make([]ProductionAdapterIndependentEffectGate, 0, len(in))
	for _, gate := range in {
		out = append(out, gate.Clone())
	}
	return out
}

func appendDisplaySafeRefIfPresent(in []DisplaySafeRef, ref DisplaySafeRef) []DisplaySafeRef {
	ref = normalizeOneDisplaySafeRef(ref)
	if ref == "" {
		return in
	}
	for _, existing := range in {
		if existing == ref {
			return in
		}
	}
	return append(in, ref)
}

func appendUniqueControlTokens(in []string, values ...[]string) []string {
	out := cloneStringSlice(in)
	for _, group := range values {
		for _, value := range group {
			out = appendUniqueControlToken(out, value)
		}
	}
	return out
}

func productionAdapterIndependentEffectGateNextHostAction(kind ProductionAdapterEffectGateKind) NextHostAction {
	if kind == "" {
		return "review_independent_effect_gate"
	}
	return NextHostAction("review_" + string(kind) + "_independent_gate")
}
