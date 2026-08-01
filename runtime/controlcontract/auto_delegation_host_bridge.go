package controlcontract

type AutoDelegationHostBridgeInput struct {
	PlanReview                        AutoDelegationPlanReview               `json:"plan_review,omitempty"`
	Activation                        Activation                             `json:"activation,omitempty"`
	ParentFrame                       ObjectiveFrame                         `json:"parent_frame,omitempty"`
	Budget                            ObjectiveBudgetSnapshot                `json:"budget,omitempty"`
	WorkerRuntimeGate                 ProductionAdapterIndependentEffectGate `json:"worker_runtime_gate,omitempty"`
	RequestedIntensity                ExecutionIntensity                     `json:"requested_intensity,omitempty"`
	ExecutionContractAllowsDelegation bool                                   `json:"execution_contract_allows_delegation"`
	HostAllowsL4Delegation            bool                                   `json:"host_allows_l4_delegation"`
	L5Enabled                         bool                                   `json:"l5_enabled"`
	UserConfirmed                     bool                                   `json:"user_confirmed"`
	HostApproved                      bool                                   `json:"host_approved"`
	MaxChildren                       int                                    `json:"max_children,omitempty"`
	MaxParallelism                    int                                    `json:"max_parallelism,omitempty"`
	MaxAttemptsPerChild               int                                    `json:"max_attempts_per_child,omitempty"`
	MaxDurationSeconds                int                                    `json:"max_duration_seconds,omitempty"`
	MaxDepth                          int                                    `json:"max_depth,omitempty"`
	AllowedToolRefs                   []DisplaySafeRef                       `json:"allowed_tool_refs,omitempty"`
	DeniedToolRefs                    []DisplaySafeRef                       `json:"denied_tool_refs,omitempty"`
	ApprovalRefs                      []DisplaySafeRef                       `json:"approval_refs,omitempty"`
	PolicyRefs                        []DisplaySafeRef                       `json:"policy_refs,omitempty"`
	StopConditionRefs                 []DisplaySafeRef                       `json:"stop_condition_refs,omitempty"`
	RedactionPolicyRef                DisplaySafeRef                         `json:"redaction_policy_ref,omitempty"`
	MergePolicyRef                    DisplaySafeRef                         `json:"merge_policy_ref,omitempty"`
	AdapterRef                        DisplaySafeRef                         `json:"adapter_ref,omitempty"`
	AdapterVersionRef                 DisplaySafeRef                         `json:"adapter_version_ref,omitempty"`
	AdapterCapabilityRef              DisplaySafeRef                         `json:"adapter_capability_ref,omitempty"`
	AdapterContractRef                DisplaySafeRef                         `json:"adapter_contract_ref,omitempty"`
	HostConfirmationRef               DisplaySafeRef                         `json:"host_confirmation_ref,omitempty"`
	WorkerRuntimeRef                  DisplaySafeRef                         `json:"worker_runtime_ref,omitempty"`
	IdempotencyRef                    DisplaySafeRef                         `json:"idempotency_ref,omitempty"`
	BudgetRef                         DisplaySafeRef                         `json:"budget_ref,omitempty"`
	VerificationRef                   DisplaySafeRef                         `json:"verification_ref,omitempty"`
	FailureReviewRef                  DisplaySafeRef                         `json:"failure_review_ref,omitempty"`
	FailureBindingRef                 DisplaySafeRef                         `json:"failure_binding_ref,omitempty"`
	CompensationRef                   DisplaySafeRef                         `json:"compensation_ref,omitempty"`
	CancellationRef                   DisplaySafeRef                         `json:"cancellation_ref,omitempty"`
	ChildBindings                     []AutoDelegationHostChildBinding       `json:"child_bindings,omitempty"`
	DecisionBasis                     []DisplaySafeRef                       `json:"decision_basis,omitempty"`
	Boundaries                        []Boundary                             `json:"boundaries,omitempty"`
	ProgressSummary                   string                                 `json:"progress_summary,omitempty"`
	CompletionSummary                 string                                 `json:"completion_summary,omitempty"`
	RawOutputLoaded                   bool                                   `json:"raw_output_loaded"`
}

type AutoDelegationHostChildBinding struct {
	ChildRef             DisplaySafeRef   `json:"child_ref,omitempty"`
	SubgoalRef           DisplaySafeRef   `json:"subgoal_ref,omitempty"`
	WorkerRef            DisplaySafeRef   `json:"worker_ref,omitempty"`
	AllowedToolRefs      []DisplaySafeRef `json:"allowed_tool_refs,omitempty"`
	BoundAllowedToolRefs []DisplaySafeRef `json:"bound_allowed_tool_refs,omitempty"`
	BoundCapabilityRefs  []DisplaySafeRef `json:"bound_capability_refs,omitempty"`
	DeniedToolRefs       []DisplaySafeRef `json:"denied_tool_refs,omitempty"`
	ApprovalRefs         []DisplaySafeRef `json:"approval_refs,omitempty"`
	PolicyRefs           []DisplaySafeRef `json:"policy_refs,omitempty"`
	StopConditionRefs    []DisplaySafeRef `json:"stop_condition_refs,omitempty"`
	RedactionPolicyRef   DisplaySafeRef   `json:"redaction_policy_ref,omitempty"`
	MergePolicyRef       DisplaySafeRef   `json:"merge_policy_ref,omitempty"`
	AdapterRef           DisplaySafeRef   `json:"adapter_ref,omitempty"`
	AdapterVersionRef    DisplaySafeRef   `json:"adapter_version_ref,omitempty"`
	AdapterCapabilityRef DisplaySafeRef   `json:"adapter_capability_ref,omitempty"`
	AdapterContractRef   DisplaySafeRef   `json:"adapter_contract_ref,omitempty"`
	HostConfirmationRef  DisplaySafeRef   `json:"host_confirmation_ref,omitempty"`
	WorkerRunRef         DisplaySafeRef   `json:"worker_run_ref,omitempty"`
	WorkerRequestRef     DisplaySafeRef   `json:"worker_request_ref,omitempty"`
	InvocationRef        DisplaySafeRef   `json:"invocation_ref,omitempty"`
	ResultBindingRef     DisplaySafeRef   `json:"result_binding_ref,omitempty"`
	ReadbackBindingRef   DisplaySafeRef   `json:"readback_binding_ref,omitempty"`
	IdempotencyRef       DisplaySafeRef   `json:"idempotency_ref,omitempty"`
	BudgetRef            DisplaySafeRef   `json:"budget_ref,omitempty"`
	VerificationRef      DisplaySafeRef   `json:"verification_ref,omitempty"`
	FailureBindingRef    DisplaySafeRef   `json:"failure_binding_ref,omitempty"`
	CompensationRef      DisplaySafeRef   `json:"compensation_ref,omitempty"`
	CancellationRef      DisplaySafeRef   `json:"cancellation_ref,omitempty"`
	LineageRefs          []DisplaySafeRef `json:"lineage_refs,omitempty"`
	ContextRefs          []DisplaySafeRef `json:"context_refs,omitempty"`
	DecisionBasis        []DisplaySafeRef `json:"decision_basis,omitempty"`
	Active               bool             `json:"active"`
	Completed            bool             `json:"completed"`
	Failed               bool             `json:"failed"`
	Cancelled            bool             `json:"cancelled"`
	ProgressSummary      string           `json:"progress_summary,omitempty"`
	CompletionSummary    string           `json:"completion_summary,omitempty"`
	FailureClass         FailureClass     `json:"failure_class,omitempty"`
	Boundaries           []Boundary       `json:"boundaries,omitempty"`
	MissingInputs        []MissingInput   `json:"missing_inputs,omitempty"`
	RawOutputLoaded      bool             `json:"raw_output_loaded"`
}

type AutoDelegationHostChildRuntime struct {
	ContractVersion                  string                                    `json:"contract_version,omitempty"`
	Projected                        bool                                      `json:"projected"`
	Status                           HostActionStatus                          `json:"status,omitempty"`
	Ready                            bool                                      `json:"ready"`
	Queued                           bool                                      `json:"queued"`
	Active                           bool                                      `json:"active"`
	Completed                        bool                                      `json:"completed"`
	Failed                           bool                                      `json:"failed"`
	Cancelled                        bool                                      `json:"cancelled"`
	LeafDelegationDisabled           bool                                      `json:"leaf_delegation_disabled"`
	WorkerResultRequiresVerification bool                                      `json:"worker_result_requires_verification"`
	WorkerOutputAcceptedAsFact       bool                                      `json:"worker_output_accepted_as_fact"`
	Child                            AutoDelegationChildTask                   `json:"child,omitempty"`
	Binding                          AutoDelegationHostChildBinding            `json:"binding,omitempty"`
	Request                          DelegationRequestProjection               `json:"request,omitempty"`
	Readiness                        HostOwnedDelegationWorkerRuntimeReadiness `json:"readiness,omitempty"`
	EffectiveAllowedToolRefs         []DisplaySafeRef                          `json:"effective_allowed_tool_refs,omitempty"`
	EffectiveDeniedToolRefs          []DisplaySafeRef                          `json:"effective_denied_tool_refs,omitempty"`
	LineageRefs                      []DisplaySafeRef                          `json:"lineage_refs,omitempty"`
	ContextRefs                      []DisplaySafeRef                          `json:"context_refs,omitempty"`
	CancellationRef                  DisplaySafeRef                            `json:"cancellation_ref,omitempty"`
	ProgressSummary                  string                                    `json:"progress_summary,omitempty"`
	CompletionSummary                string                                    `json:"completion_summary,omitempty"`
	MissingInputs                    []MissingInput                            `json:"missing_inputs,omitempty"`
	BlockedReasons                   []string                                  `json:"blocked_reasons,omitempty"`
	FailureClass                     FailureClass                              `json:"failure_class,omitempty"`
	DecisionBasis                    []DisplaySafeRef                          `json:"decision_basis,omitempty"`
	Boundaries                       []Boundary                                `json:"boundaries,omitempty"`
	NextHostAction                   NextHostAction                            `json:"next_host_action,omitempty"`
	RunnerEffect                     string                                    `json:"runner_effect,omitempty"`
	PromptEffect                     string                                    `json:"prompt_effect,omitempty"`
	RuntimeEffect                    string                                    `json:"runtime_effect,omitempty"`
	RawOutputLoaded                  bool                                      `json:"raw_output_loaded"`
}

type AutoDelegationHostBridge struct {
	ContractVersion                  string                                      `json:"contract_version,omitempty"`
	Projected                        bool                                        `json:"projected"`
	Status                           HostActionStatus                            `json:"status,omitempty"`
	Ready                            bool                                        `json:"ready"`
	HostMayInvokeWorkerRuntime       bool                                        `json:"host_may_invoke_worker_runtime"`
	PlanReview                       AutoDelegationPlanReview                    `json:"plan_review,omitempty"`
	WorkerRuntimeGate                ProductionAdapterIndependentEffectGate      `json:"worker_runtime_gate,omitempty"`
	Children                         []AutoDelegationHostChildRuntime            `json:"children,omitempty"`
	DelegationRequests               []DelegationRequestProjection               `json:"delegation_requests,omitempty"`
	WorkerReadiness                  []HostOwnedDelegationWorkerRuntimeReadiness `json:"worker_readiness,omitempty"`
	MaxChildren                      int                                         `json:"max_children,omitempty"`
	MaxParallelism                   int                                         `json:"max_parallelism,omitempty"`
	MaxAttemptsPerChild              int                                         `json:"max_attempts_per_child,omitempty"`
	MaxDurationSeconds               int                                         `json:"max_duration_seconds,omitempty"`
	MaxDepth                         int                                         `json:"max_depth,omitempty"`
	AcceptedChildRefs                []DisplaySafeRef                            `json:"accepted_child_refs,omitempty"`
	InvokableChildRefs               []DisplaySafeRef                            `json:"invokable_child_refs,omitempty"`
	QueuedChildRefs                  []DisplaySafeRef                            `json:"queued_child_refs,omitempty"`
	ActiveChildRefs                  []DisplaySafeRef                            `json:"active_child_refs,omitempty"`
	CompletedChildRefs               []DisplaySafeRef                            `json:"completed_child_refs,omitempty"`
	FailedChildRefs                  []DisplaySafeRef                            `json:"failed_child_refs,omitempty"`
	CancelledChildRefs               []DisplaySafeRef                            `json:"cancelled_child_refs,omitempty"`
	LineageRefs                      []DisplaySafeRef                            `json:"lineage_refs,omitempty"`
	CancellationRef                  DisplaySafeRef                              `json:"cancellation_ref,omitempty"`
	FailureReviewRef                 DisplaySafeRef                              `json:"failure_review_ref,omitempty"`
	ProgressSummary                  string                                      `json:"progress_summary,omitempty"`
	CompletionSummary                string                                      `json:"completion_summary,omitempty"`
	MissingInputs                    []MissingInput                              `json:"missing_inputs,omitempty"`
	BlockedReasons                   []string                                    `json:"blocked_reasons,omitempty"`
	FailureClass                     FailureClass                                `json:"failure_class,omitempty"`
	DecisionBasis                    []DisplaySafeRef                            `json:"decision_basis,omitempty"`
	Boundaries                       []Boundary                                  `json:"boundaries,omitempty"`
	NextHostAction                   NextHostAction                              `json:"next_host_action,omitempty"`
	RunnerEffect                     string                                      `json:"runner_effect,omitempty"`
	PromptEffect                     string                                      `json:"prompt_effect,omitempty"`
	RuntimeEffect                    string                                      `json:"runtime_effect,omitempty"`
	WorkerResultRequiresVerification bool                                        `json:"worker_result_requires_verification"`
	WorkerOutputAcceptedAsFact       bool                                        `json:"worker_output_accepted_as_fact"`
	RawOutputLoaded                  bool                                        `json:"raw_output_loaded"`
}

func BuildAutoDelegationHostBridge(input AutoDelegationHostBridgeInput) AutoDelegationHostBridge {
	planReview := input.PlanReview.Normalize()
	plan := planReview.Plan.Normalize()
	policy := plan.Policy.Normalize()
	gate := input.WorkerRuntimeGate.Normalize()
	limits := autoDelegationHostBridgeLimits(input, plan, policy)
	result := AutoDelegationHostBridge{
		ContractVersion:     ContractVersion,
		Projected:           true,
		Status:              HostActionBlocked,
		PlanReview:          planReview,
		WorkerRuntimeGate:   gate,
		MaxChildren:         limits.maxChildren,
		MaxParallelism:      limits.maxParallelism,
		MaxAttemptsPerChild: limits.maxAttemptsPerChild,
		MaxDurationSeconds:  limits.maxDurationSeconds,
		MaxDepth:            limits.maxDepth,
		AcceptedChildRefs:   cloneDisplaySafeRefs(planReview.AcceptedChildRefs),
		CancellationRef:     normalizeOneDisplaySafeRef(input.CancellationRef),
		FailureReviewRef:    normalizeOneDisplaySafeRef(input.FailureReviewRef),
		ProgressSummary:     autoDelegationPlanSafeText(input.ProgressSummary),
		CompletionSummary:   autoDelegationPlanSafeText(input.CompletionSummary),
		FailureClass:        FailureNone,
		DecisionBasis: normalizeDisplaySafeRefs(append(
			[]DisplaySafeRef{
				"auto_delegation:host_bridge",
				"auto_delegation:host_owned_worker_runtime",
			},
			input.DecisionBasis...,
		)),
		Boundaries: MergeBoundaries(
			[]Boundary{
				"auto_delegation_host_bridge",
				"host_owned_task_subagent_bridge",
				"projection_only",
				"display_safe_refs_only",
				"no_child_task_spawn_by_core",
				"no_subagent_dispatch_by_core",
				"no_runner_dispatch",
				"worker_as_tool_default",
				"parent_verification_required",
				"child_output_not_fact",
			},
			input.Boundaries,
		),
		NextHostAction:                   "review_auto_delegation_host_bridge",
		RunnerEffect:                     "none",
		PromptEffect:                     "none",
		RuntimeEffect:                    "none",
		WorkerResultRequiresVerification: true,
		WorkerOutputAcceptedAsFact:       false,
		RawOutputLoaded:                  input.RawOutputLoaded || planReview.RawOutputLoaded || gate.RawOutputLoaded,
	}
	if autoDelegationHostBridgeUnsafe(input) {
		result.RawOutputLoaded = true
		return autoDelegationHostBridgeBlock(result, HostActionReviewRequired, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs", "raw_output_not_allowed").Normalize()
	}
	if !planReview.Ready || !planReview.HostMayDispatch {
		result.MissingInputs = AppendMissingInputs(result.MissingInputs, planReview.MissingInputs...)
		result.BlockedReasons = appendUniqueControlTokens(result.BlockedReasons, planReview.BlockedReasons)
		return autoDelegationHostBridgeBlock(result, HostActionBlocked, firstFailureClass(planReview.FailureClass, FailurePolicyBlocked), "auto_delegation_plan_not_dispatchable", "host:auto_delegation_plan_review", firstNextHostAction(planReview.NextHostAction, "review_auto_delegation_plan"), "auto_delegation_plan_not_dispatchable").Normalize()
	}
	if len(result.AcceptedChildRefs) == 0 {
		result = autoDelegationHostBridgeBlock(result, HostActionBlocked, FailureInsufficientInformation, "auto_delegation_accepted_children_missing", "host:auto_delegation_accepted_children", "provide_auto_delegation_plan", "auto_delegation_accepted_children_missing")
	}
	if len(result.AcceptedChildRefs) > limits.maxChildren {
		result = autoDelegationHostBridgeBlock(result, HostActionBlocked, FailurePolicyBlocked, "auto_delegation_child_budget_exceeded", "host:auto_delegation_child_budget", "reduce_auto_delegation_children", "auto_delegation_child_budget_exceeded")
	}

	bindingsByChild := autoDelegationHostBridgeBindingsByChild(input.ChildBindings)
	activeChildCount := autoDelegationHostBridgeActiveChildCount(plan.Children, planReview.AcceptedChildRefs, bindingsByChild)
	if activeChildCount > limits.maxParallelism {
		result = autoDelegationHostBridgeBlock(result, HostActionBlocked, FailurePolicyBlocked, "auto_delegation_parallelism_exceeded", "host:auto_delegation_parallelism_slot", "wait_for_child_completion", "auto_delegation_parallelism_exceeded")
	}
	availableSlots := limits.maxParallelism - activeChildCount
	if availableSlots < 0 {
		availableSlots = 0
	}

	accepted := autoDelegationHostBridgeAcceptedSet(planReview.AcceptedChildRefs)
	for _, child := range plan.Children {
		child = child.Normalize()
		if _, ok := accepted[child.ChildRef]; !ok {
			continue
		}
		binding := autoDelegationHostBridgeResolvedBinding(input, child, bindingsByChild[child.ChildRef])
		runtime := buildAutoDelegationHostChildRuntime(input, plan, policy, child, binding, gate, limits, availableSlots > 0)
		if runtime.Ready && !runtime.Active && !runtime.Completed && !runtime.Failed && !runtime.Cancelled {
			availableSlots--
		}
		result.Children = append(result.Children, runtime)
		result.DelegationRequests = append(result.DelegationRequests, runtime.Request)
		result.WorkerReadiness = append(result.WorkerReadiness, runtime.Readiness)
		result.LineageRefs = append(result.LineageRefs, runtime.LineageRefs...)
		result.MissingInputs = AppendMissingInputs(result.MissingInputs, autoDelegationHostBridgeAggregateChildMissingInputs(runtime)...)
		result.BlockedReasons = appendUniqueControlTokens(result.BlockedReasons, autoDelegationHostBridgeAggregateChildReasons(runtime))
		result.Boundaries = AppendBoundaries(result.Boundaries, runtime.Boundaries...)
		result.FailureClass = firstFailureClass(result.FailureClass, runtime.FailureClass)
		switch {
		case runtime.Ready:
			result.InvokableChildRefs = appendDisplaySafeRefIfPresent(result.InvokableChildRefs, runtime.Child.ChildRef)
		case runtime.Queued:
			result.QueuedChildRefs = appendDisplaySafeRefIfPresent(result.QueuedChildRefs, runtime.Child.ChildRef)
		}
		if runtime.Active {
			result.ActiveChildRefs = appendDisplaySafeRefIfPresent(result.ActiveChildRefs, runtime.Child.ChildRef)
		}
		if runtime.Completed {
			result.CompletedChildRefs = appendDisplaySafeRefIfPresent(result.CompletedChildRefs, runtime.Child.ChildRef)
		}
		if runtime.Failed {
			result.FailedChildRefs = appendDisplaySafeRefIfPresent(result.FailedChildRefs, runtime.Child.ChildRef)
		}
		if runtime.Cancelled {
			result.CancelledChildRefs = appendDisplaySafeRefIfPresent(result.CancelledChildRefs, runtime.Child.ChildRef)
		}
	}
	result.LineageRefs = normalizeDisplaySafeRefs(result.LineageRefs)
	if len(result.Children) == 0 {
		result = autoDelegationHostBridgeBlock(result, HostActionBlocked, FailureInsufficientInformation, "auto_delegation_child_runtime_missing", "host:auto_delegation_child_runtime", "provide_auto_delegation_host_bindings", "auto_delegation_child_runtime_missing")
	}
	if len(result.InvokableChildRefs) == 0 && len(result.ActiveChildRefs) == 0 && len(result.CompletedChildRefs) == 0 && len(result.FailedChildRefs) == 0 && len(result.CancelledChildRefs) == 0 && len(result.MissingInputs) == 0 {
		result = autoDelegationHostBridgeBlock(result, HostActionBlocked, FailurePolicyBlocked, "auto_delegation_parallelism_slot_missing", "host:auto_delegation_parallelism_slot", "wait_for_child_completion", "auto_delegation_parallelism_window_queued")
	}
	if len(result.MissingInputs) == 0 && len(result.BlockedReasons) == 0 && len(result.InvokableChildRefs) > 0 {
		result.Status = HostActionReady
		result.Ready = true
		result.HostMayInvokeWorkerRuntime = true
		result.NextHostAction = "host_may_invoke_auto_delegation_children"
		result.Boundaries = AppendBoundaries(result.Boundaries, "auto_delegation_host_bridge_ready", "host_may_invoke_auto_delegation_children")
	}
	return result.Normalize()
}

func buildAutoDelegationHostChildRuntime(input AutoDelegationHostBridgeInput, plan AutoDelegationPlan, policy AutoDelegationPolicy, child AutoDelegationChildTask, binding AutoDelegationHostChildBinding, gate ProductionAdapterIndependentEffectGate, limits autoDelegationHostBridgeLimitSet, slotAvailable bool) AutoDelegationHostChildRuntime {
	effectiveAllowed, effectiveDenied := autoDelegationHostBridgeEffectiveToolRefs(input, policy, child, binding)
	budget := autoDelegationHostBridgeBudget(input.Budget, policy, firstDisplaySafeRef(binding.BudgetRef, input.BudgetRef))
	frame := autoDelegationHostBridgeFrame(input.ParentFrame, plan, child)
	request := BuildDelegationRequestProjection(DelegationRequestInput{
		Activation:                        input.Activation,
		Frame:                             frame,
		RequestedIntensity:                firstIntensity(input.RequestedIntensity, IntensityL4DurableLongRun),
		SubgoalRef:                        firstDisplaySafeRef(binding.SubgoalRef, child.ChildRef),
		WorkerRef:                         binding.WorkerRef,
		AllowedToolRefs:                   effectiveAllowed,
		DeniedToolRefs:                    effectiveDenied,
		Budget:                            budget,
		EvidenceRequirements:              child.ExpectedEvidence,
		StopConditionRefs:                 binding.StopConditionRefs,
		RedactionPolicyRef:                binding.RedactionPolicyRef,
		MergePolicyRef:                    binding.MergePolicyRef,
		ExecutionContractAllowsDelegation: input.ExecutionContractAllowsDelegation,
		HostAllowsL4Delegation:            input.HostAllowsL4Delegation,
		L5Enabled:                         input.L5Enabled,
		UserConfirmed:                     input.UserConfirmed,
		HostApproved:                      input.HostApproved,
		ApprovalRefs:                      binding.ApprovalRefs,
		PolicyRefs:                        binding.PolicyRefs,
		DecisionBasis: mergeDisplaySafeRefs(
			[]DisplaySafeRef{"auto_delegation:child_delegation_request", child.ChildRef},
			binding.DecisionBasis,
			input.DecisionBasis,
		),
		Boundaries: MergeBoundaries(
			[]Boundary{"auto_delegation_child_delegation_request"},
			child.Boundaries,
			binding.Boundaries,
		),
		RawOutputLoaded: input.RawOutputLoaded || child.RawOutputLoaded || binding.RawOutputLoaded,
	})
	readiness := BuildHostOwnedDelegationWorkerRuntimeReadiness(HostOwnedDelegationWorkerRuntimeReadinessInput{
		Request:              request,
		WorkerRuntimeGate:    gate,
		AdapterRef:           binding.AdapterRef,
		AdapterVersionRef:    binding.AdapterVersionRef,
		AdapterCapabilityRef: binding.AdapterCapabilityRef,
		AdapterContractRef:   binding.AdapterContractRef,
		HostConfirmationRef:  binding.HostConfirmationRef,
		WorkerRunRef:         binding.WorkerRunRef,
		WorkerRequestRef:     binding.WorkerRequestRef,
		InvocationRef:        binding.InvocationRef,
		ResultBindingRef:     binding.ResultBindingRef,
		ReadbackBindingRef:   binding.ReadbackBindingRef,
		IdempotencyRef:       binding.IdempotencyRef,
		BudgetRef:            binding.BudgetRef,
		VerificationRef:      binding.VerificationRef,
		FailureBindingRef:    binding.FailureBindingRef,
		CompensationRef:      binding.CompensationRef,
		EvidenceRefs:         child.ExpectedEvidence,
		DecisionBasis:        binding.DecisionBasis,
		Boundaries:           binding.Boundaries,
		RawOutputLoaded:      input.RawOutputLoaded || child.RawOutputLoaded || binding.RawOutputLoaded,
	})
	runtime := AutoDelegationHostChildRuntime{
		ContractVersion:          ContractVersion,
		Projected:                true,
		Status:                   HostActionBlocked,
		Child:                    child,
		Binding:                  binding,
		Request:                  request,
		Readiness:                readiness,
		EffectiveAllowedToolRefs: effectiveAllowed,
		EffectiveDeniedToolRefs:  effectiveDenied,
		LineageRefs:              normalizeDisplaySafeRefs(append([]DisplaySafeRef{plan.ParentObjectiveRef, child.ParentObjectiveRef, child.ChildRef}, binding.LineageRefs...)),
		ContextRefs:              mergeDisplaySafeRefs(child.ContextRefs, binding.ContextRefs),
		CancellationRef:          binding.CancellationRef,
		Active:                   binding.Active,
		Completed:                binding.Completed,
		Failed:                   binding.Failed,
		Cancelled:                binding.Cancelled,
		LeafDelegationDisabled:   child.Role == AutoDelegationChildRoleLeaf,
		ProgressSummary:          autoDelegationPlanSafeText(binding.ProgressSummary),
		CompletionSummary:        autoDelegationPlanSafeText(binding.CompletionSummary),
		MissingInputs:            MergeMissingInputs(binding.MissingInputs, request.MissingInputs, readiness.MissingInputs),
		BlockedReasons:           appendUniqueControlTokens(request.BlockedReasons, readiness.BlockedReasons),
		FailureClass:             firstFailureClass(binding.FailureClass, request.FailureClass, readiness.FailureClass, FailureNone),
		DecisionBasis: mergeDisplaySafeRefs(
			[]DisplaySafeRef{"auto_delegation:host_child_runtime", child.ChildRef},
			binding.DecisionBasis,
		),
		Boundaries: MergeBoundaries(
			[]Boundary{
				"auto_delegation_host_child_runtime",
				"host_owned_child_runtime",
				"worker_as_tool_child",
				"parent_verification_required",
				"child_output_not_fact",
			},
			child.Boundaries,
			binding.Boundaries,
			request.Boundaries,
			readiness.Boundaries,
		),
		NextHostAction:                   firstNextHostAction(firstNextHostAction(readiness.NextHostAction, request.NextHostAction), "provide_auto_delegation_host_binding"),
		RunnerEffect:                     "none",
		PromptEffect:                     "none",
		RuntimeEffect:                    "none",
		WorkerResultRequiresVerification: true,
		WorkerOutputAcceptedAsFact:       false,
		RawOutputLoaded:                  input.RawOutputLoaded || child.RawOutputLoaded || binding.RawOutputLoaded || request.RawOutputLoaded || readiness.RawOutputLoaded,
	}
	if runtime.LeafDelegationDisabled {
		runtime.Boundaries = AppendBoundaries(runtime.Boundaries, "child_leaf_no_recursive_delegation")
	}
	if child.Role == AutoDelegationChildRoleOrchestrator {
		runtime.Boundaries = AppendBoundaries(runtime.Boundaries, "child_orchestrator_depth_policy_required")
	}
	if child.MaxAttempts > limits.maxAttemptsPerChild && child.MaxAttempts > 0 {
		runtime = autoDelegationHostChildRuntimeBlock(runtime, FailurePolicyBlocked, "auto_delegation_child_attempt_limit_exceeded", "host:auto_delegation_attempt_budget", "reduce_auto_delegation_child_attempts", "auto_delegation_child_attempt_limit_exceeded")
	}
	if child.MaxDurationSeconds > limits.maxDurationSeconds && child.MaxDurationSeconds > 0 {
		runtime = autoDelegationHostChildRuntimeBlock(runtime, FailurePolicyBlocked, "auto_delegation_child_timeout_exceeded", "host:auto_delegation_timeout", "reduce_auto_delegation_child_timeout", "auto_delegation_child_timeout_exceeded")
	}
	if child.Depth >= limits.maxDepth && child.Role == AutoDelegationChildRoleOrchestrator {
		runtime = autoDelegationHostChildRuntimeBlock(runtime, FailurePolicyBlocked, "auto_delegation_child_depth_exceeded", "host:auto_delegation_depth_policy", "review_auto_delegation_depth_policy", "auto_delegation_child_depth_exceeded")
	}
	if binding.Cancelled && binding.CancellationRef == "" {
		runtime = autoDelegationHostChildRuntimeBlock(runtime, FailureConfigMissing, "auto_delegation_child_cancellation_ref_missing", "host:auto_delegation_child_cancellation_ref", "provide_child_cancellation_ref", "auto_delegation_child_cancellation_ref_missing")
	}
	if binding.Failed && binding.FailureBindingRef == "" {
		runtime = autoDelegationHostChildRuntimeBlock(runtime, FailureConfigMissing, "auto_delegation_child_failure_binding_missing", "host:auto_delegation_child_failure_binding", "provide_child_failure_binding", "auto_delegation_child_failure_binding_missing")
	}
	if autoDelegationHostChildBindingDispatched(binding) {
		runtime = autoDelegationHostChildRuntimeValidateImmutableBinding(runtime, input, policy, child, binding)
		if runtime.Status == HostActionBlocked && len(runtime.BlockedReasons) > 0 {
			return runtime.Normalize()
		}
	}
	if binding.Completed {
		runtime.Status = HostActionRecorded
		runtime.Ready = false
		runtime.NextHostAction = "review_auto_delegation_child_completion"
		runtime.Boundaries = AppendBoundaries(runtime.Boundaries, "auto_delegation_child_completed")
		return runtime.Normalize()
	}
	if binding.Active {
		runtime.Status = HostActionRecorded
		runtime.Ready = false
		runtime.NextHostAction = "monitor_auto_delegation_child"
		runtime.Boundaries = AppendBoundaries(runtime.Boundaries, "auto_delegation_child_active")
		return runtime.Normalize()
	}
	if binding.Failed {
		runtime.Status = HostActionReviewRequired
		runtime.Ready = false
		runtime.NextHostAction = "review_auto_delegation_child_failure"
		runtime.Boundaries = AppendBoundaries(runtime.Boundaries, "auto_delegation_child_failed")
		return runtime.Normalize()
	}
	if binding.Cancelled {
		runtime.Status = HostActionRecorded
		runtime.Ready = false
		runtime.NextHostAction = "review_auto_delegation_child_cancellation"
		runtime.Boundaries = AppendBoundaries(runtime.Boundaries, "auto_delegation_child_cancelled")
		return runtime.Normalize()
	}
	if !slotAvailable {
		runtime.Status = HostActionNotReady
		runtime.Queued = true
		runtime.Ready = false
		runtime.BlockedReasons = appendUniqueControlToken(runtime.BlockedReasons, "auto_delegation_parallelism_window_queued")
		runtime.MissingInputs = AppendMissingInputs(runtime.MissingInputs, "host:auto_delegation_parallelism_slot")
		runtime.NextHostAction = "wait_for_child_completion"
		runtime.Boundaries = AppendBoundaries(runtime.Boundaries, "auto_delegation_parallelism_window_queued")
		return runtime.Normalize()
	}
	if len(runtime.MissingInputs) == 0 && len(runtime.BlockedReasons) == 0 && readiness.HostMayInvokeWorkerRuntime {
		runtime.Status = HostActionReady
		runtime.Ready = true
		runtime.NextHostAction = "host_may_invoke_auto_delegation_child"
		runtime.Boundaries = AppendBoundaries(runtime.Boundaries, "auto_delegation_child_ready_for_host_runtime", "host_may_invoke_auto_delegation_child")
	}
	return runtime.Normalize()
}

func (b AutoDelegationHostBridge) Normalize() AutoDelegationHostBridge {
	out := b
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.PlanReview = out.PlanReview.Normalize()
	out.WorkerRuntimeGate = out.WorkerRuntimeGate.Normalize()
	for i := range out.Children {
		out.Children[i] = out.Children[i].Normalize()
	}
	for i := range out.DelegationRequests {
		out.DelegationRequests[i] = out.DelegationRequests[i].Normalize()
	}
	for i := range out.WorkerReadiness {
		out.WorkerReadiness[i] = out.WorkerReadiness[i].Normalize()
	}
	out.AcceptedChildRefs = normalizeDisplaySafeRefs(out.AcceptedChildRefs)
	out.InvokableChildRefs = normalizeDisplaySafeRefs(out.InvokableChildRefs)
	out.QueuedChildRefs = normalizeDisplaySafeRefs(out.QueuedChildRefs)
	out.ActiveChildRefs = normalizeDisplaySafeRefs(out.ActiveChildRefs)
	out.CompletedChildRefs = normalizeDisplaySafeRefs(out.CompletedChildRefs)
	out.FailedChildRefs = normalizeDisplaySafeRefs(out.FailedChildRefs)
	out.CancelledChildRefs = normalizeDisplaySafeRefs(out.CancelledChildRefs)
	out.LineageRefs = normalizeDisplaySafeRefs(out.LineageRefs)
	out.CancellationRef = normalizeOneDisplaySafeRef(out.CancellationRef)
	out.FailureReviewRef = normalizeOneDisplaySafeRef(out.FailureReviewRef)
	out.ProgressSummary = autoDelegationPlanSafeText(out.ProgressSummary)
	out.CompletionSummary = autoDelegationPlanSafeText(out.CompletionSummary)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.BlockedReasons = normalizeControlTokenList(out.BlockedReasons)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.DecisionBasis = normalizeDisplaySafeRefs(out.DecisionBasis)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	out.RuntimeEffect = normalizeControlToken(out.RuntimeEffect)
	if out.RunnerEffect == "" {
		out.RunnerEffect = "none"
	}
	if out.PromptEffect == "" {
		out.PromptEffect = "none"
	}
	if out.RuntimeEffect == "" {
		out.RuntimeEffect = "none"
	}
	if out.Status == HostActionNotReady && (len(out.MissingInputs) > 0 || len(out.BlockedReasons) > 0) {
		out.Status = HostActionBlocked
	}
	if out.RawOutputLoaded {
		out.Status = HostActionReviewRequired
		out.Ready = false
		out.HostMayInvokeWorkerRuntime = false
		out.FailureClass = firstFailureClass(out.FailureClass, FailureEvidenceWeak)
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		out.NextHostAction = "provide_display_safe_refs"
	}
	return out
}

func (r AutoDelegationHostChildRuntime) Normalize() AutoDelegationHostChildRuntime {
	out := r
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.Child = out.Child.Normalize()
	out.Binding = out.Binding.Normalize()
	out.Request = out.Request.Normalize()
	out.Readiness = out.Readiness.Normalize()
	out.EffectiveAllowedToolRefs = normalizeDisplaySafeRefs(out.EffectiveAllowedToolRefs)
	out.EffectiveDeniedToolRefs = normalizeDisplaySafeRefs(out.EffectiveDeniedToolRefs)
	out.LineageRefs = normalizeDisplaySafeRefs(out.LineageRefs)
	out.ContextRefs = normalizeDisplaySafeRefs(out.ContextRefs)
	out.CancellationRef = normalizeOneDisplaySafeRef(out.CancellationRef)
	out.ProgressSummary = autoDelegationPlanSafeText(out.ProgressSummary)
	out.CompletionSummary = autoDelegationPlanSafeText(out.CompletionSummary)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.BlockedReasons = normalizeControlTokenList(out.BlockedReasons)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.DecisionBasis = normalizeDisplaySafeRefs(out.DecisionBasis)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	out.RuntimeEffect = normalizeControlToken(out.RuntimeEffect)
	if out.RunnerEffect == "" {
		out.RunnerEffect = "none"
	}
	if out.PromptEffect == "" {
		out.PromptEffect = "none"
	}
	if out.RuntimeEffect == "" {
		out.RuntimeEffect = "none"
	}
	if out.RawOutputLoaded {
		out.Status = HostActionReviewRequired
		out.Ready = false
		out.FailureClass = firstFailureClass(out.FailureClass, FailureEvidenceWeak)
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		out.NextHostAction = "provide_display_safe_refs"
	}
	return out
}

func (b AutoDelegationHostChildBinding) Normalize() AutoDelegationHostChildBinding {
	out := b
	out.ChildRef = normalizeOneDisplaySafeRef(out.ChildRef)
	out.SubgoalRef = normalizeOneDisplaySafeRef(out.SubgoalRef)
	out.WorkerRef = normalizeOneDisplaySafeRef(out.WorkerRef)
	out.AllowedToolRefs = normalizeDisplaySafeRefs(out.AllowedToolRefs)
	out.BoundAllowedToolRefs = normalizeDisplaySafeRefs(out.BoundAllowedToolRefs)
	out.BoundCapabilityRefs = normalizeDisplaySafeRefs(out.BoundCapabilityRefs)
	out.DeniedToolRefs = normalizeDisplaySafeRefs(out.DeniedToolRefs)
	out.ApprovalRefs = normalizeDisplaySafeRefs(out.ApprovalRefs)
	out.PolicyRefs = normalizeDisplaySafeRefs(out.PolicyRefs)
	out.StopConditionRefs = normalizeDisplaySafeRefs(out.StopConditionRefs)
	out.RedactionPolicyRef = normalizeOneDisplaySafeRef(out.RedactionPolicyRef)
	out.MergePolicyRef = normalizeOneDisplaySafeRef(out.MergePolicyRef)
	out.AdapterRef = normalizeOneDisplaySafeRef(out.AdapterRef)
	out.AdapterVersionRef = normalizeOneDisplaySafeRef(out.AdapterVersionRef)
	out.AdapterCapabilityRef = normalizeOneDisplaySafeRef(out.AdapterCapabilityRef)
	out.AdapterContractRef = normalizeOneDisplaySafeRef(out.AdapterContractRef)
	out.HostConfirmationRef = normalizeOneDisplaySafeRef(out.HostConfirmationRef)
	out.WorkerRunRef = normalizeOneDisplaySafeRef(out.WorkerRunRef)
	out.WorkerRequestRef = normalizeOneDisplaySafeRef(out.WorkerRequestRef)
	out.InvocationRef = normalizeOneDisplaySafeRef(out.InvocationRef)
	out.ResultBindingRef = normalizeOneDisplaySafeRef(out.ResultBindingRef)
	out.ReadbackBindingRef = normalizeOneDisplaySafeRef(out.ReadbackBindingRef)
	out.IdempotencyRef = normalizeOneDisplaySafeRef(out.IdempotencyRef)
	out.BudgetRef = normalizeOneDisplaySafeRef(out.BudgetRef)
	out.VerificationRef = normalizeOneDisplaySafeRef(out.VerificationRef)
	out.FailureBindingRef = normalizeOneDisplaySafeRef(out.FailureBindingRef)
	out.CompensationRef = normalizeOneDisplaySafeRef(out.CompensationRef)
	out.CancellationRef = normalizeOneDisplaySafeRef(out.CancellationRef)
	out.LineageRefs = normalizeDisplaySafeRefs(out.LineageRefs)
	out.ContextRefs = normalizeDisplaySafeRefs(out.ContextRefs)
	out.DecisionBasis = normalizeDisplaySafeRefs(out.DecisionBasis)
	out.ProgressSummary = autoDelegationPlanSafeText(out.ProgressSummary)
	out.CompletionSummary = autoDelegationPlanSafeText(out.CompletionSummary)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	return out
}

type autoDelegationHostBridgeLimitSet struct {
	maxChildren         int
	maxParallelism      int
	maxAttemptsPerChild int
	maxDurationSeconds  int
	maxDepth            int
}

func autoDelegationHostBridgeLimits(input AutoDelegationHostBridgeInput, plan AutoDelegationPlan, policy AutoDelegationPolicy) autoDelegationHostBridgeLimitSet {
	limits := autoDelegationHostBridgeLimitSet{
		maxChildren:         firstPositiveInt(input.MaxChildren, plan.MaxChildren, policy.MaxChildren, DefaultAutoDelegationMaxChildren),
		maxParallelism:      firstPositiveInt(input.MaxParallelism, plan.MaxParallelism, policy.MaxParallelism, DefaultAutoDelegationMaxParallelism),
		maxAttemptsPerChild: firstPositiveInt(input.MaxAttemptsPerChild, policy.MaxAttemptsPerChild, DefaultAutoDelegationMaxAttemptsPerChild),
		maxDurationSeconds:  firstPositiveInt(input.MaxDurationSeconds, policy.MaxDurationSeconds, DefaultAutoDelegationMaxDurationSeconds),
		maxDepth:            firstPositiveInt(input.MaxDepth, plan.MaxDepth, DefaultAutoDelegationMaxDepth),
	}
	if limits.maxParallelism > limits.maxChildren {
		limits.maxParallelism = limits.maxChildren
	}
	return limits
}

func autoDelegationHostBridgeBindingsByChild(bindings []AutoDelegationHostChildBinding) map[DisplaySafeRef]AutoDelegationHostChildBinding {
	out := map[DisplaySafeRef]AutoDelegationHostChildBinding{}
	for _, binding := range bindings {
		normalized := binding.Normalize()
		if normalized.ChildRef == "" {
			continue
		}
		if _, exists := out[normalized.ChildRef]; exists {
			continue
		}
		out[normalized.ChildRef] = normalized
	}
	return out
}

func autoDelegationHostBridgeAcceptedSet(refs []DisplaySafeRef) map[DisplaySafeRef]struct{} {
	out := map[DisplaySafeRef]struct{}{}
	for _, ref := range normalizeDisplaySafeRefs(refs) {
		out[ref] = struct{}{}
	}
	return out
}

func autoDelegationHostBridgeActiveChildCount(children []AutoDelegationChildTask, acceptedRefs []DisplaySafeRef, bindings map[DisplaySafeRef]AutoDelegationHostChildBinding) int {
	accepted := autoDelegationHostBridgeAcceptedSet(acceptedRefs)
	count := 0
	for _, child := range children {
		child = child.Normalize()
		if _, ok := accepted[child.ChildRef]; !ok {
			continue
		}
		if bindings[child.ChildRef].Active {
			count++
		}
	}
	return count
}

func autoDelegationHostBridgeResolvedBinding(input AutoDelegationHostBridgeInput, child AutoDelegationChildTask, binding AutoDelegationHostChildBinding) AutoDelegationHostChildBinding {
	out := binding.Normalize()
	if out.ChildRef == "" {
		out.ChildRef = child.ChildRef
	}
	out.SubgoalRef = firstDisplaySafeRef(out.SubgoalRef, child.ChildRef)
	out.WorkerRef = firstDisplaySafeRef(out.WorkerRef, firstDisplaySafeRef(input.WorkerRuntimeRef, firstDisplaySafeRef(firstDisplaySafeRefFromSlice(child.CapabilityRefs), out.ChildRef)))
	out.AllowedToolRefs = firstDisplaySafeRefs(out.AllowedToolRefs, input.AllowedToolRefs)
	out.DeniedToolRefs = mergeDisplaySafeRefs(out.DeniedToolRefs, input.DeniedToolRefs, child.DeniedToolRefs)
	out.ApprovalRefs = firstDisplaySafeRefs(out.ApprovalRefs, input.ApprovalRefs)
	out.PolicyRefs = firstDisplaySafeRefs(out.PolicyRefs, input.PolicyRefs)
	out.StopConditionRefs = firstDisplaySafeRefs(out.StopConditionRefs, input.StopConditionRefs)
	out.RedactionPolicyRef = firstDisplaySafeRef(out.RedactionPolicyRef, input.RedactionPolicyRef)
	out.MergePolicyRef = firstDisplaySafeRef(out.MergePolicyRef, input.MergePolicyRef)
	out.AdapterRef = firstDisplaySafeRef(out.AdapterRef, input.AdapterRef)
	out.AdapterVersionRef = firstDisplaySafeRef(out.AdapterVersionRef, input.AdapterVersionRef)
	out.AdapterCapabilityRef = firstDisplaySafeRef(out.AdapterCapabilityRef, input.AdapterCapabilityRef)
	out.AdapterContractRef = firstDisplaySafeRef(out.AdapterContractRef, input.AdapterContractRef)
	out.HostConfirmationRef = firstDisplaySafeRef(out.HostConfirmationRef, input.HostConfirmationRef)
	out.IdempotencyRef = firstDisplaySafeRef(out.IdempotencyRef, input.IdempotencyRef)
	out.BudgetRef = firstDisplaySafeRef(out.BudgetRef, input.BudgetRef)
	out.VerificationRef = firstDisplaySafeRef(out.VerificationRef, input.VerificationRef)
	out.FailureBindingRef = firstDisplaySafeRef(out.FailureBindingRef, input.FailureBindingRef)
	out.CompensationRef = firstDisplaySafeRef(out.CompensationRef, input.CompensationRef)
	out.CancellationRef = firstDisplaySafeRef(out.CancellationRef, input.CancellationRef)
	out.LineageRefs = mergeDisplaySafeRefs(out.LineageRefs, []DisplaySafeRef{child.ParentObjectiveRef, child.ChildRef})
	out.ContextRefs = mergeDisplaySafeRefs(out.ContextRefs, child.ContextRefs)
	return out.Normalize()
}

func autoDelegationHostBridgeEffectiveToolRefs(input AutoDelegationHostBridgeInput, policy AutoDelegationPolicy, child AutoDelegationChildTask, binding AutoDelegationHostChildBinding) ([]DisplaySafeRef, []DisplaySafeRef) {
	allowed := normalizeDisplaySafeRefs(child.AllowedToolRefs)
	if len(allowed) == 0 {
		allowed = normalizeDisplaySafeRefs(binding.AllowedToolRefs)
	}
	if len(allowed) == 0 {
		allowed = normalizeDisplaySafeRefs(input.AllowedToolRefs)
	}
	if len(policy.AllowedToolRefs) > 0 {
		allowed = intersectDisplaySafeRefs(allowed, policy.AllowedToolRefs)
	}
	if len(input.AllowedToolRefs) > 0 {
		allowed = intersectDisplaySafeRefs(allowed, input.AllowedToolRefs)
	}
	if len(binding.AllowedToolRefs) > 0 && len(child.AllowedToolRefs) > 0 {
		allowed = intersectDisplaySafeRefs(allowed, binding.AllowedToolRefs)
	}
	denied := mergeDisplaySafeRefs(policy.DeniedToolRefs, input.DeniedToolRefs, child.DeniedToolRefs, binding.DeniedToolRefs)
	allowed = subtractDisplaySafeRefs(allowed, denied)
	return allowed, denied
}

func autoDelegationHostBridgeBudget(in ObjectiveBudgetSnapshot, policy AutoDelegationPolicy, budgetRef DisplaySafeRef) ObjectiveBudgetSnapshot {
	out := in.Normalize()
	if out.BudgetRef == "" {
		out.BudgetRef = normalizeOneDisplaySafeRef(budgetRef)
	}
	if out.Limit == 0 && policy.MaxBudgetTokens > 0 {
		out.Limit = policy.MaxBudgetTokens
	}
	return out.Normalize()
}

func autoDelegationHostBridgeFrame(in ObjectiveFrame, plan AutoDelegationPlan, child AutoDelegationChildTask) ObjectiveFrame {
	out := in.Normalize()
	if out.ID == "" {
		out.ID = string(firstDisplaySafeRef(child.ParentObjectiveRef, plan.ParentObjectiveRef))
	}
	if out.Intensity == "" {
		out.Intensity = IntensityL4DurableLongRun
	}
	if len(out.RequiredEvidence) == 0 {
		out.RequiredEvidence = MergeEvidenceRefs(plan.RequiredEvidence, child.ExpectedEvidence)
	}
	if len(out.CandidateCapabilities) == 0 {
		out.CandidateCapabilities = child.CapabilityRefs
	}
	return out.Normalize()
}

func autoDelegationHostChildBindingDispatched(binding AutoDelegationHostChildBinding) bool {
	return binding.Active || binding.Completed || binding.Failed || binding.Cancelled
}

func autoDelegationHostChildRuntimeValidateImmutableBinding(runtime AutoDelegationHostChildRuntime, input AutoDelegationHostBridgeInput, policy AutoDelegationPolicy, child AutoDelegationChildTask, binding AutoDelegationHostChildBinding) AutoDelegationHostChildRuntime {
	boundCapabilities := normalizeDisplaySafeRefs(binding.BoundCapabilityRefs)
	if len(child.CapabilityRefs) > 0 {
		if len(boundCapabilities) == 0 {
			runtime = autoDelegationHostChildRuntimeBlock(runtime, FailureConfigMissing, "auto_delegation_child_bound_capability_refs_missing", "host:auto_delegation_bound_capability_refs", "provide_auto_delegation_child_bound_refs", "auto_delegation_child_capability_refs_immutable_after_dispatch")
		} else if extras := autoDelegationHostBridgeRefsOutside(boundCapabilities, child.CapabilityRefs); len(extras) > 0 {
			runtime = autoDelegationHostChildRuntimeBlock(runtime, FailurePolicyBlocked, "auto_delegation_child_capability_refs_widened_after_dispatch", "host:auto_delegation_bound_capability_refs", "review_auto_delegation_child_bound_refs", "auto_delegation_child_capability_refs_immutable_after_dispatch")
		}
	}

	boundTools := normalizeDisplaySafeRefs(binding.BoundAllowedToolRefs)
	immutableAllowedTools := autoDelegationHostBridgeImmutableAllowedToolRefs(input, policy, child)
	if len(immutableAllowedTools) > 0 {
		if len(boundTools) == 0 {
			runtime = autoDelegationHostChildRuntimeBlock(runtime, FailureConfigMissing, "auto_delegation_child_bound_tool_refs_missing", "host:auto_delegation_bound_allowed_tool_refs", "provide_auto_delegation_child_bound_refs", "auto_delegation_child_tool_refs_immutable_after_dispatch")
		} else if extras := autoDelegationHostBridgeRefsOutside(boundTools, immutableAllowedTools); len(extras) > 0 {
			runtime = autoDelegationHostChildRuntimeBlock(runtime, FailurePolicyBlocked, "auto_delegation_child_tool_refs_widened_after_dispatch", "host:auto_delegation_bound_allowed_tool_refs", "review_auto_delegation_child_bound_refs", "auto_delegation_child_tool_refs_immutable_after_dispatch")
		}
	} else if len(boundTools) > 0 {
		runtime = autoDelegationHostChildRuntimeBlock(runtime, FailurePolicyBlocked, "auto_delegation_child_tool_refs_widened_after_dispatch", "host:auto_delegation_bound_allowed_tool_refs", "review_auto_delegation_child_bound_refs", "auto_delegation_child_tool_refs_immutable_after_dispatch")
	}
	if runtime.Status != HostActionBlocked {
		runtime.Boundaries = AppendBoundaries(runtime.Boundaries, "auto_delegation_child_bound_refs_verified", "auto_delegation_child_refs_immutable_after_dispatch")
	}
	return runtime
}

func autoDelegationHostBridgeImmutableAllowedToolRefs(input AutoDelegationHostBridgeInput, policy AutoDelegationPolicy, child AutoDelegationChildTask) []DisplaySafeRef {
	allowed := normalizeDisplaySafeRefs(child.AllowedToolRefs)
	if len(allowed) == 0 {
		allowed = normalizeDisplaySafeRefs(input.AllowedToolRefs)
	}
	if len(policy.AllowedToolRefs) > 0 {
		allowed = intersectDisplaySafeRefs(allowed, policy.AllowedToolRefs)
	}
	if len(input.AllowedToolRefs) > 0 {
		allowed = intersectDisplaySafeRefs(allowed, input.AllowedToolRefs)
	}
	denied := mergeDisplaySafeRefs(policy.DeniedToolRefs, input.DeniedToolRefs, child.DeniedToolRefs)
	return subtractDisplaySafeRefs(allowed, denied)
}

func autoDelegationHostBridgeRefsOutside(values []DisplaySafeRef, allowed []DisplaySafeRef) []DisplaySafeRef {
	normalizedValues := normalizeDisplaySafeRefs(values)
	normalizedAllowed := normalizeDisplaySafeRefs(allowed)
	if len(normalizedValues) == 0 {
		return nil
	}
	allowedSet := map[DisplaySafeRef]struct{}{}
	for _, value := range normalizedAllowed {
		allowedSet[value] = struct{}{}
	}
	out := make([]DisplaySafeRef, 0, len(normalizedValues))
	for _, value := range normalizedValues {
		if _, ok := allowedSet[value]; ok {
			continue
		}
		out = append(out, value)
	}
	return normalizeDisplaySafeRefs(out)
}

func autoDelegationHostBridgeAggregateChildReasons(runtime AutoDelegationHostChildRuntime) []string {
	if runtime.Status == HostActionBlocked {
		return runtime.BlockedReasons
	}
	if runtime.Queued || runtime.Active || runtime.Completed || runtime.Failed || runtime.Cancelled {
		return nil
	}
	return runtime.BlockedReasons
}

func autoDelegationHostBridgeAggregateChildMissingInputs(runtime AutoDelegationHostChildRuntime) []MissingInput {
	if runtime.Status == HostActionBlocked {
		return runtime.MissingInputs
	}
	if runtime.Queued || runtime.Active || runtime.Completed || runtime.Cancelled {
		return nil
	}
	return runtime.MissingInputs
}

func autoDelegationHostBridgeBlock(result AutoDelegationHostBridge, status HostActionStatus, failure FailureClass, reason string, missing MissingInput, next NextHostAction, boundary Boundary) AutoDelegationHostBridge {
	result.Status = status
	result.Ready = false
	result.HostMayInvokeWorkerRuntime = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = next
	result.Boundaries = AppendBoundaries(result.Boundaries, boundary)
	return result
}

func autoDelegationHostChildRuntimeBlock(runtime AutoDelegationHostChildRuntime, failure FailureClass, reason string, missing MissingInput, next NextHostAction, boundary Boundary) AutoDelegationHostChildRuntime {
	runtime.Status = HostActionBlocked
	runtime.Ready = false
	runtime.FailureClass = firstFailureClass(runtime.FailureClass, failure)
	runtime.BlockedReasons = appendUniqueControlToken(runtime.BlockedReasons, reason)
	runtime.MissingInputs = AppendMissingInputs(runtime.MissingInputs, missing)
	runtime.NextHostAction = next
	runtime.Boundaries = AppendBoundaries(runtime.Boundaries, boundary)
	return runtime
}

func autoDelegationHostBridgeUnsafe(input AutoDelegationHostBridgeInput) bool {
	if input.RawOutputLoaded ||
		displaySafeRefSliceRejected(input.AllowedToolRefs) ||
		displaySafeRefSliceRejected(input.DeniedToolRefs) ||
		displaySafeRefSliceRejected(input.ApprovalRefs) ||
		displaySafeRefSliceRejected(input.PolicyRefs) ||
		displaySafeRefSliceRejected(input.StopConditionRefs) ||
		displaySafeRefRejected(input.RedactionPolicyRef) ||
		displaySafeRefRejected(input.MergePolicyRef) ||
		displaySafeRefRejected(input.AdapterRef) ||
		displaySafeRefRejected(input.AdapterVersionRef) ||
		displaySafeRefRejected(input.AdapterCapabilityRef) ||
		displaySafeRefRejected(input.AdapterContractRef) ||
		displaySafeRefRejected(input.HostConfirmationRef) ||
		displaySafeRefRejected(input.WorkerRuntimeRef) ||
		displaySafeRefRejected(input.IdempotencyRef) ||
		displaySafeRefRejected(input.BudgetRef) ||
		displaySafeRefRejected(input.VerificationRef) ||
		displaySafeRefRejected(input.FailureReviewRef) ||
		displaySafeRefRejected(input.FailureBindingRef) ||
		displaySafeRefRejected(input.CompensationRef) ||
		displaySafeRefRejected(input.CancellationRef) {
		return true
	}
	for _, binding := range input.ChildBindings {
		if autoDelegationHostBridgeBindingUnsafe(binding) {
			return true
		}
	}
	return false
}

func autoDelegationHostBridgeBindingUnsafe(binding AutoDelegationHostChildBinding) bool {
	return binding.RawOutputLoaded ||
		displaySafeRefRejected(binding.ChildRef) ||
		displaySafeRefRejected(binding.SubgoalRef) ||
		displaySafeRefRejected(binding.WorkerRef) ||
		displaySafeRefSliceRejected(binding.AllowedToolRefs) ||
		displaySafeRefSliceRejected(binding.BoundAllowedToolRefs) ||
		displaySafeRefSliceRejected(binding.BoundCapabilityRefs) ||
		displaySafeRefSliceRejected(binding.DeniedToolRefs) ||
		displaySafeRefSliceRejected(binding.ApprovalRefs) ||
		displaySafeRefSliceRejected(binding.PolicyRefs) ||
		displaySafeRefSliceRejected(binding.StopConditionRefs) ||
		displaySafeRefRejected(binding.RedactionPolicyRef) ||
		displaySafeRefRejected(binding.MergePolicyRef) ||
		displaySafeRefRejected(binding.AdapterRef) ||
		displaySafeRefRejected(binding.AdapterVersionRef) ||
		displaySafeRefRejected(binding.AdapterCapabilityRef) ||
		displaySafeRefRejected(binding.AdapterContractRef) ||
		displaySafeRefRejected(binding.HostConfirmationRef) ||
		displaySafeRefRejected(binding.WorkerRunRef) ||
		displaySafeRefRejected(binding.WorkerRequestRef) ||
		displaySafeRefRejected(binding.InvocationRef) ||
		displaySafeRefRejected(binding.ResultBindingRef) ||
		displaySafeRefRejected(binding.ReadbackBindingRef) ||
		displaySafeRefRejected(binding.IdempotencyRef) ||
		displaySafeRefRejected(binding.BudgetRef) ||
		displaySafeRefRejected(binding.VerificationRef) ||
		displaySafeRefRejected(binding.FailureBindingRef) ||
		displaySafeRefRejected(binding.CompensationRef) ||
		displaySafeRefRejected(binding.CancellationRef) ||
		displaySafeRefSliceRejected(binding.LineageRefs) ||
		displaySafeRefSliceRejected(binding.ContextRefs) ||
		displaySafeRefSliceRejected(binding.DecisionBasis)
}

func firstPositiveInt(values ...int) int {
	for _, value := range values {
		if value > 0 {
			return value
		}
	}
	return 0
}

func firstDisplaySafeRefFromSlice(values []DisplaySafeRef) DisplaySafeRef {
	for _, value := range normalizeDisplaySafeRefs(values) {
		return value
	}
	return ""
}

func firstDisplaySafeRefs(values ...[]DisplaySafeRef) []DisplaySafeRef {
	for _, value := range values {
		normalized := normalizeDisplaySafeRefs(value)
		if len(normalized) > 0 {
			return normalized
		}
	}
	return nil
}

func mergeDisplaySafeRefs(groups ...[]DisplaySafeRef) []DisplaySafeRef {
	merged := []DisplaySafeRef{}
	for _, group := range groups {
		merged = append(merged, group...)
	}
	return normalizeDisplaySafeRefs(merged)
}

func intersectDisplaySafeRefs(left []DisplaySafeRef, right []DisplaySafeRef) []DisplaySafeRef {
	normalizedLeft := normalizeDisplaySafeRefs(left)
	normalizedRight := normalizeDisplaySafeRefs(right)
	if len(normalizedLeft) == 0 || len(normalizedRight) == 0 {
		return nil
	}
	rightSet := map[DisplaySafeRef]struct{}{}
	for _, value := range normalizedRight {
		rightSet[value] = struct{}{}
	}
	out := make([]DisplaySafeRef, 0, len(normalizedLeft))
	for _, value := range normalizedLeft {
		if _, ok := rightSet[value]; ok {
			out = append(out, value)
		}
	}
	return normalizeDisplaySafeRefs(out)
}

func subtractDisplaySafeRefs(left []DisplaySafeRef, denied []DisplaySafeRef) []DisplaySafeRef {
	normalizedLeft := normalizeDisplaySafeRefs(left)
	normalizedDenied := normalizeDisplaySafeRefs(denied)
	if len(normalizedLeft) == 0 || len(normalizedDenied) == 0 {
		return normalizedLeft
	}
	deniedSet := map[DisplaySafeRef]struct{}{}
	for _, value := range normalizedDenied {
		deniedSet[value] = struct{}{}
	}
	out := make([]DisplaySafeRef, 0, len(normalizedLeft))
	for _, value := range normalizedLeft {
		if _, denied := deniedSet[value]; denied {
			continue
		}
		out = append(out, value)
	}
	return normalizeDisplaySafeRefs(out)
}
