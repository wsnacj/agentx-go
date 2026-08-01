package controlcontract

import "testing"

func boundaryContains(values []Boundary, want Boundary) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func TestBuildManagedObjectiveIngressDefaultOff(t *testing.T) {
	got := BuildManagedObjectiveIngress(ManagedObjectiveIngressInput{
		Activation:     ActivationOff,
		ObjectiveID:    "objective:test_ingress",
		UserGoalDigest: "sha256:test",
		SourceRef:      "host:test_ingress",
	})
	if got.ReadyForObjectiveLoop ||
		got.ReadyForStrategyPlanning ||
		got.Status != VerificationBlocked ||
		got.FailureClass != FailurePolicyBlocked ||
		got.RunnerEffect != "none" ||
		got.PromptEffect != "none" {
		t.Fatalf("unexpected default-off ingress = %#v", got)
	}
	if !missingInputContains(got.MissingInputs, "control_plane:managed_activation") ||
		!boundaryContains(got.Boundaries, "no_runner_dispatch") ||
		!boundaryContains(got.Boundaries, "core_does_not_parse_goal_text") {
		t.Fatalf("expected activation/boundary guard, missing=%#v boundaries=%#v", got.MissingInputs, got.Boundaries)
	}
}

func TestBuildManagedObjectiveIngressBlocksWithoutStrategyCatalog(t *testing.T) {
	got := BuildManagedObjectiveIngress(managedObjectiveIngressReadyInput(StrategyCatalogSnapshot{}))
	if !got.ReadyForObjectiveLoop ||
		!got.ReadyForStrategyPlanning ||
		got.ReadyForStrategyFinalGate ||
		got.Status != VerificationBlocked ||
		got.FailureClass != FailureConfigMissing ||
		got.NextHostAction != "provide_strategy_catalog" {
		t.Fatalf("unexpected no-catalog ingress = %#v", got)
	}
	if !missingInputContains(got.MissingInputs, "host:strategy_catalog") ||
		!boundaryContains(got.StrategyPlan.Boundaries, "strategy_catalog_empty") ||
		got.ManagedObjective.Status != HostActionReady ||
		!got.ManagedObjective.Ready {
		t.Fatalf("expected managed ready but strategy catalog blocked, got managed=%#v plan=%#v missing=%#v", got.ManagedObjective, got.StrategyPlan, got.MissingInputs)
	}
}

func TestBuildManagedObjectiveIngressPlansHostCatalogCandidateButBlocksWithoutAdapterRegistry(t *testing.T) {
	input := managedObjectiveIngressReadyInput(managedObjectiveIngressTestCatalog())
	got := BuildManagedObjectiveIngress(input)
	if !got.ReadyForObjectiveLoop ||
		!got.ReadyForStrategyPlanning ||
		!got.ReadyForStrategyFinalGate ||
		got.ReadyForRuntimeAdapter ||
		got.Status != VerificationBlocked ||
		got.FailureClass != FailureHostAdapterMissing ||
		got.NextHostAction != "provide_adapter_registry" ||
		got.SelectedStrategyRef != "strategy:test_host_adapter" ||
		got.StrategyPlan.Selected.Candidate.ID != "strategy:test_host_adapter" {
		t.Fatalf("unexpected planned ingress = %#v", got)
	}
	if !boundaryContains(got.Boundaries, "no_runtime_adapter_execution") ||
		!boundaryContains(got.StrategyPlan.Boundaries, "strategy_candidate_ranked") ||
		!boundaryContains(got.RuntimeAdapterRequest.Boundaries, "runtime_adapter_registry_not_ready") {
		t.Fatalf("expected projection-only planner/request boundaries, got ingress=%#v plan=%#v request=%#v", got.Boundaries, got.StrategyPlan.Boundaries, got.RuntimeAdapterRequest.Boundaries)
	}
}

func TestBuildManagedObjectiveIngressBuildsReadyRuntimeAdapterRequest(t *testing.T) {
	descriptor := managedObjectiveIngressTestAdapterDescriptor()
	input := managedObjectiveIngressReadyInput(managedObjectiveIngressTestCatalog())
	input.AdapterRegistry = BuildHostAdapterRegistry(HostAdapterRegistryInput{
		RegistryRef: "registry:test_ingress",
		Descriptors: []ProductionAdapterDescriptor{
			descriptor,
		},
		PolicyRefs: []DisplaySafeRef{"policy:test_ingress", "contract:test_ingress"},
	})
	input.RequestedAdapterRef = descriptor.AdapterRef
	input.AvailableCapabilityRefs = descriptor.RequiresCapabilityRefs
	input.IdempotencyRef = "idempotency:test_ingress"
	input.RuntimeInputRefs = descriptor.DisplaySafeInputRefs
	input.ExpectedObservationKinds = []string{"objective_answer"}

	got := BuildManagedObjectiveIngress(input)
	if !got.ReadyForObjectiveLoop ||
		!got.ReadyForStrategyPlanning ||
		!got.ReadyForStrategyFinalGate ||
		!got.ReadyForRuntimeAdapter ||
		got.Status != VerificationSatisfied ||
		got.NextHostAction != "host_may_execute_runtime_adapter" ||
		got.RuntimeAdapterRequest.AdapterRef != "adapter:test_host_readonly" ||
		!got.RuntimeAdapterRequest.ReadyForHostExecution {
		t.Fatalf("unexpected ready runtime adapter ingress = %#v", got)
	}
	if !boundaryContains(got.Boundaries, "adapter_request_ready_not_objective_satisfied") ||
		!boundaryContains(got.RuntimeAdapterRequest.Boundaries, "core_does_not_execute_adapter") ||
		!boundaryContains(got.RuntimeAdapterRequest.Boundaries, "ready_for_host_runtime_adapter_execution") {
		t.Fatalf("expected ready request without execution, got ingress=%#v request=%#v", got.Boundaries, got.RuntimeAdapterRequest.Boundaries)
	}

	clone := got.Clone()
	clone.StrategyPlan.Selected.Candidate.ID = "strategy:mutated"
	clone.ManagedObjective.AllowedStrategyRefs[0] = "strategy:mutated"
	clone.RuntimeAdapterRequest.AdapterRef = "adapter:mutated"
	if got.StrategyPlan.Selected.Candidate.ID != "strategy:test_host_adapter" ||
		got.ManagedObjective.AllowedStrategyRefs[0] == "strategy:mutated" ||
		got.RuntimeAdapterRequest.AdapterRef == "adapter:mutated" {
		t.Fatalf("clone mutated original = %#v", got)
	}
}

func TestBuildManagedObjectiveIngressPassesHostSideEffectAdapterApproval(t *testing.T) {
	descriptor := managedObjectiveIngressTestWorkflowAdapterDescriptor()
	input := managedObjectiveIngressReadyInput(managedObjectiveIngressWorkflowTestCatalog())
	input.Policy = managedObjectiveIngressWorkflowTestPolicy()
	input.AllowedStrategyRefs = []DisplaySafeRef{"strategy:test_workflow_runtime"}
	input.AdapterRegistry = BuildHostAdapterRegistry(HostAdapterRegistryInput{
		RegistryRef: "registry:test_workflow_runtime",
		Descriptors: []ProductionAdapterDescriptor{
			descriptor,
		},
		PolicyRefs: []DisplaySafeRef{"policy:test_workflow_runtime", "contract:test_workflow_runtime"},
	})
	input.RequestedAdapterRef = descriptor.AdapterRef
	input.AvailableCapabilityRefs = descriptor.RequiresCapabilityRefs
	input.ApprovalRefs = []DisplaySafeRef{"approval:test_workflow_runtime", "approval:test_workflow_runtime_side_effect"}
	input.AllowHostSideEffectAdapter = true
	input.HostSideEffectApprovalRefs = []DisplaySafeRef{"approval:test_workflow_runtime_side_effect"}
	input.Budget = ObjectiveBudgetSnapshot{BudgetRef: "budget:test_workflow_runtime", Limit: 3, Remaining: 3}
	input.IdempotencyRef = "idempotency:test_workflow_runtime"
	input.RuntimeInputRefs = descriptor.DisplaySafeInputRefs
	input.ExpectedObservationKinds = []string{"workflow_node_result"}
	input.RequiredEvidence = []EvidenceRef{{
		Ref:      "evidence:test_workflow_runtime_node",
		Kind:     "workflow_node_result",
		Strength: EvidenceAdequate,
		Source:   "workflow_spec:test_workflow_runtime",
	}}

	got := BuildManagedObjectiveIngress(input)
	if !got.ReadyForRuntimeAdapter ||
		got.Status != VerificationSatisfied ||
		got.RuntimeAdapterRequest.Descriptor.Kind != ProductionAdapterWorkflowDispatch ||
		!got.RuntimeAdapterRequest.ReadyForHostExecution ||
		!got.RuntimeAdapterRequest.HostSideEffectAdapterAllowed ||
		!displaySafeRefSliceContains(got.RuntimeAdapterRequest.HostSideEffectApprovalRefs, "approval:test_workflow_runtime_side_effect") {
		t.Fatalf("expected side-effect workflow adapter request to be ready: %#v", got)
	}
	if !boundaryContains(got.RuntimeAdapterRequest.Boundaries, "host_side_effect_adapter_explicitly_allowed") ||
		!boundaryContains(got.RuntimeAdapterRequest.Boundaries, "ready_for_host_runtime_adapter_execution") ||
		got.RuntimeAdapterRequest.RunnerEffect != "none" ||
		got.RuntimeAdapterRequest.PromptEffect != "none" {
		t.Fatalf("expected projection-only side-effect request boundaries: %#v", got.RuntimeAdapterRequest)
	}

	input.AllowHostSideEffectAdapter = false
	input.HostSideEffectApprovalRefs = nil
	blocked := BuildManagedObjectiveIngress(input)
	if blocked.ReadyForRuntimeAdapter ||
		blocked.RuntimeAdapterRequest.ReadyForHostExecution ||
		blocked.RuntimeAdapterRequest.FailureClass != FailurePolicyBlocked ||
		!missingInputContains(blocked.RuntimeAdapterRequest.MissingInputs, "contract:read_only_runtime_adapter") ||
		blocked.RuntimeAdapterRequest.NextHostAction != "request_host_approval" {
		t.Fatalf("expected workflow adapter to block without explicit side-effect opt-in: %#v", blocked.RuntimeAdapterRequest)
	}
}

func managedObjectiveIngressTestCatalog() StrategyCatalogSnapshot {
	return StrategyCatalogSnapshot{
		CatalogRef: "catalog:test_ingress",
		Entries: []StrategyCatalogEntry{{
			SourceKind: StrategyCatalogSourceHostAdapter,
			SourceRef:  "host:test_catalog",
			Status:     VerificationSatisfied,
			Candidate: StrategyCandidate{
				ID:               "strategy:test_host_adapter",
				Kind:             "host_declared_strategy",
				ControlMode:      ControlModeObjective,
				MinIntensity:     IntensityL3ManagedObjective,
				MaxIntensity:     IntensityL3ManagedObjective,
				ExpectedEvidence: []EvidenceRef{{Ref: "evidence:test_answer", Kind: "objective_answer", Strength: EvidenceAdequate, Source: "host:test_ingress"}},
				Boundaries:       []Boundary{"host_catalog_candidate"},
				RequiresApproval: true,
				Owner:            "host",
			},
		}},
	}
}

func managedObjectiveIngressWorkflowTestCatalog() StrategyCatalogSnapshot {
	evidence := EvidenceRef{
		Ref:      "evidence:test_workflow_runtime_node",
		Kind:     "workflow_node_result",
		Strength: EvidenceStrong,
		Source:   "workflow_spec:test_workflow_runtime",
	}
	return StrategyCatalogSnapshot{
		CatalogRef: "catalog:test_workflow_runtime",
		Entries: []StrategyCatalogEntry{{
			SourceKind: StrategyCatalogSourceWorkflow,
			SourceRef:  "workflow_spec:test_workflow_runtime",
			Status:     VerificationSatisfied,
			Candidate: StrategyCandidate{
				ID:               "strategy:test_workflow_runtime",
				Kind:             "workflow_runtime_strategy",
				ControlMode:      ControlModeWorkflow,
				MinIntensity:     IntensityL3ManagedObjective,
				MaxIntensity:     IntensityL3ManagedObjective,
				CapabilityRefs:   []DisplaySafeRef{"capability:test_workflow_runtime_backend"},
				ExpectedEvidence: []EvidenceRef{evidence},
				Boundaries: []Boundary{
					"workflow_strategy_candidate_metadata",
					"workflow_strategy_requires_host_runtime_backend",
					"controlplane_does_not_execute_workflow",
				},
				Risk:             "requires_review",
				SideEffectClass:  "host_workflow_runtime",
				RequiresApproval: true,
				Owner:            "host",
			},
			EvidenceRefs: []EvidenceRef{evidence},
			Boundaries: []Boundary{
				"workflow_strategy_catalog_entry",
				"workflow_strategy_metadata_only",
				"workflow_strategy_requires_host_runtime_backend",
				"workflow_strategy_candidate_not_executed",
				"controlplane_does_not_execute_workflow",
				"no_workflow_dispatch",
				"no_runner_dispatch",
			},
		}},
		PolicyRefs: []DisplaySafeRef{"policy:test_workflow_runtime"},
		Boundaries: []Boundary{
			"workflow_strategy_catalog_snapshot",
			"workflow_strategy_catalog_projection_only",
		},
	}.Normalize()
}

func managedObjectiveIngressTestAdapterDescriptor() ProductionAdapterDescriptor {
	return ProductionAdapterDescriptor{
		AdapterRef:             "adapter:test_host_readonly",
		Owner:                  "host",
		OwnerRef:               "host:test_ingress",
		Version:                "v1",
		Kind:                   ProductionAdapterSourceReadback,
		SupportedCandidateRefs: []DisplaySafeRef{"strategy:test_host_adapter"},
		RequiresCapabilityRefs: []DisplaySafeRef{"capability:test_host_readonly"},
		InputContractRef:       "contract:test_input",
		OutputContractRef:      "contract:test_output",
		ReadbackContractRef:    "contract:test_readback",
		RequiredPolicyRefs:     []DisplaySafeRef{"policy:test_ingress"},
		RequiredApprovalRefs:   []DisplaySafeRef{"approval:test_ingress"},
		RequiredBudgetRef:      "budget:test_ingress",
		IdempotencyContractRef: "contract:test_idempotency",
		SideEffectClass:        "read_only",
		DisplaySafeInputRefs:   []DisplaySafeRef{"input:test_ingress"},
		DisplaySafeOutputRefs:  []DisplaySafeRef{"output:test_ingress"},
	}
}

func managedObjectiveIngressTestWorkflowAdapterDescriptor() ProductionAdapterDescriptor {
	return ProductionAdapterDescriptor{
		AdapterRef:             "adapter:test_workflow_runtime",
		Owner:                  "host",
		OwnerRef:               "host:test_workflow_runtime",
		Version:                "v1",
		Kind:                   ProductionAdapterWorkflowDispatch,
		SupportedCandidateRefs: []DisplaySafeRef{"strategy:test_workflow_runtime"},
		ProvidesCapabilityRefs: []DisplaySafeRef{"capability:test_workflow_runtime_backend"},
		RequiresCapabilityRefs: []DisplaySafeRef{"capability:test_workflow_runtime_backend"},
		InputContractRef:       "contract:test_workflow_runtime_input",
		OutputContractRef:      "contract:test_workflow_runtime_output",
		ReadbackContractRef:    "contract:test_workflow_runtime_readback",
		RequiredPolicyRefs:     []DisplaySafeRef{"policy:test_workflow_runtime"},
		RequiredApprovalRefs:   []DisplaySafeRef{"approval:test_workflow_runtime", "approval:test_workflow_runtime_side_effect"},
		RequiredBudgetRef:      "budget:test_workflow_runtime",
		IdempotencyContractRef: "contract:test_workflow_runtime_idempotency",
		SideEffectClass:        "workflow_dispatch",
		DisplaySafeInputRefs:   []DisplaySafeRef{"input:test_workflow_runtime"},
		DisplaySafeOutputRefs:  []DisplaySafeRef{"output:test_workflow_runtime"},
		CompensationHandoffRef: "compensation:test_workflow_runtime",
	}
}

func managedObjectiveIngressReadyInput(catalog StrategyCatalogSnapshot) ManagedObjectiveIngressInput {
	return ManagedObjectiveIngressInput{
		Activation:      ActivationManaged,
		ObjectiveID:     "objective:test_ingress",
		UserGoalDigest:  "sha256:test",
		SourceRef:       "host:test_ingress",
		SuccessCriteria: []string{"produce a verified answer or an explicit blocker"},
		RequiredEvidence: []EvidenceRef{{
			Ref:      "evidence:test_answer",
			Kind:     "objective_answer",
			Strength: EvidenceAdequate,
			Source:   "host:test_ingress",
		}},
		LedgerRef:     "ledger:test_ingress",
		Policy:        managedObjectiveIngressTestPolicy(),
		Budget:        ObjectiveBudgetSnapshot{BudgetRef: "budget:test_ingress", Limit: 3, Remaining: 3},
		UserConfirmed: true,
		HostApproved:  true,
		ApprovalRefs:  []DisplaySafeRef{"approval:test_ingress"},
		AllowedStrategyRefs: []DisplaySafeRef{
			"strategy:test_host_adapter",
		},
		StrategyCatalog: catalog,
		PolicyRefs: []DisplaySafeRef{
			"contract:intensity_gate",
			"contract:budget",
			"contract:approval_policy",
			"contract:strategy_scope",
			"contract:redaction_policy",
		},
		DecisionBasis: []DisplaySafeRef{"host:test_ingress_decision"},
	}
}

func managedObjectiveIngressWorkflowTestPolicy() ExecutionIntensityPolicy {
	return ExecutionIntensityPolicy{
		PolicyRef:            "policy:test_workflow_runtime",
		ExecutionContractRef: "contract:test_workflow_runtime",
		Activation:           ActivationManaged,
		DefaultIntensity:     IntensityL3ManagedObjective,
		MaxDefaultIntensity:  IntensityL3ManagedObjective,
		MaxAllowedIntensity:  IntensityL3ManagedObjective,
		AllowedControlModesByIntensity: map[ExecutionIntensity][]ControlMode{
			IntensityL3ManagedObjective: {ControlModeObjective, ControlModeWorkflow},
		},
		DeniedSideEffectsByIntensity: map[ExecutionIntensity][]string{
			IntensityL3ManagedObjective: {"external_write", "schedule_apply", "install"},
		},
		PolicyRefs: []DisplaySafeRef{
			"contract:intensity_gate",
			"contract:budget",
			"contract:approval_policy",
			"contract:strategy_scope",
			"contract:redaction_policy",
			"policy:test_workflow_runtime",
			"contract:test_workflow_runtime",
		},
	}
}

func managedObjectiveIngressTestPolicy() ExecutionIntensityPolicy {
	return ExecutionIntensityPolicy{
		PolicyRef:            "policy:test_ingress",
		ExecutionContractRef: "contract:test_ingress",
		Activation:           ActivationManaged,
		DefaultIntensity:     IntensityL3ManagedObjective,
		MaxDefaultIntensity:  IntensityL3ManagedObjective,
		MaxAllowedIntensity:  IntensityL3ManagedObjective,
		AllowedControlModesByIntensity: map[ExecutionIntensity][]ControlMode{
			IntensityL3ManagedObjective: {ControlModeObjective},
		},
		DeniedSideEffectsByIntensity: map[ExecutionIntensity][]string{
			IntensityL3ManagedObjective: {"external_write", "schedule_apply", "install"},
		},
		PolicyRefs: []DisplaySafeRef{
			"contract:intensity_gate",
			"contract:budget",
			"contract:approval_policy",
			"contract:strategy_scope",
			"contract:redaction_policy",
		},
	}
}
