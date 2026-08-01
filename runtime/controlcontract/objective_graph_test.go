package controlcontract

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func TestBuildObjectiveGraphValidationAcceptsMultiNodeChain(t *testing.T) {
	spec := objectiveGraphTestSpec(
		[]DisplaySafeRef{"capability:docker_inventory", "capability:postgres_service_discovery", "capability:postgres_schema"},
		ObjectiveSpecBudget{BudgetRef: "budget:graph", MaxNodes: 4, MaxAttempts: 5},
	)
	catalog := objectiveGraphTestCatalog(
		objectiveGraphTestCatalogEntry("strategy:docker_inventory", "capability:docker_inventory", "evidence:docker_inventory", "docker_inventory"),
		objectiveGraphTestCatalogEntry("strategy:postgres_service_discovery", "capability:postgres_service_discovery", "evidence:postgres_service", "service_discovery"),
		objectiveGraphTestCatalogEntry("strategy:postgres_schema", "capability:postgres_schema", "evidence:postgres_schema", "database_schema"),
	)
	graph := ObjectiveGraph{
		GraphRef:   "graph:docker_postgres",
		CatalogRef: "catalog:objective_graph",
		Nodes: []ObjectiveNode{
			objectiveGraphTestNode("node:inventory", "strategy:docker_inventory", "capability:docker_inventory", "evidence:docker_inventory", "docker_inventory"),
			objectiveGraphTestNodeWithDeps("node:discovery", "strategy:postgres_service_discovery", "capability:postgres_service_discovery", "evidence:postgres_service", "service_discovery", []ObjectiveNodeDependency{{
				DependencyRef: "dependency:inventory_to_discovery",
				NodeRef:       "node:inventory",
			}}),
			objectiveGraphTestNodeWithDeps("node:schema", "strategy:postgres_schema", "capability:postgres_schema", "evidence:postgres_schema", "database_schema", []ObjectiveNodeDependency{{
				DependencyRef: "dependency:discovery_to_schema",
				NodeRef:       "node:discovery",
			}}),
		},
	}

	got := BuildObjectiveGraphValidation(ObjectiveGraphValidationInput{
		Graph:                 graph,
		Spec:                  spec,
		Catalog:               catalog,
		CapabilityDescriptors: objectiveGraphTestDescriptors("capability:docker_inventory", "capability:postgres_service_discovery", "capability:postgres_schema"),
		Policy:                objectiveGraphTestPolicy(),
		SourceRef:             "host:objective_graph_test",
	})
	if got.Status != VerificationSatisfied ||
		!got.ReadyForRuntimeLoop ||
		got.NextHostAction != "run_bounded_objective_runtime_loop" ||
		got.ReadyNodeCount != 1 ||
		len(got.Graph.Nodes) != 3 {
		t.Fatalf("unexpected graph validation = %#v", got)
	}
	if got.Graph.Nodes[0].State != ObjectiveNodeStateReady ||
		got.Graph.Nodes[1].State != ObjectiveNodeStatePending ||
		got.Graph.Nodes[2].State != ObjectiveNodeStatePending {
		t.Fatalf("unexpected node states = %#v", got.Graph.Nodes)
	}
	if !objectiveGraphTestBoundaryContains(got.Boundaries, "objective_graph_validated") ||
		!objectiveGraphTestBoundaryContains(got.Boundaries, "no_runner_dispatch") ||
		!objectiveGraphTestBoundaryContains(got.Boundaries, "deterministic_graph_validator") {
		t.Fatalf("boundaries = %#v", got.Boundaries)
	}
}

func TestBuildObjectiveGraphWithPlannerJSONAllowsOptionalPathSkip(t *testing.T) {
	spec := objectiveGraphTestSpec(
		[]DisplaySafeRef{"capability:public_source_retrieval"},
		ObjectiveSpecBudget{BudgetRef: "budget:optional", MaxNodes: 3, MaxAttempts: 3},
	)
	spec.SuccessCriteria[0].RequiredEvidence = []EvidenceRef{{
		Ref:      "evidence:public_source_summary",
		Kind:     "summary",
		Strength: EvidenceAdequate,
		Source:   "source:objective_graph",
	}}
	spec.RequiredEvidence = cloneEvidenceRefs(spec.SuccessCriteria[0].RequiredEvidence)
	catalog := objectiveGraphTestCatalog(
		objectiveGraphTestCatalogEntry("strategy:public_source", "capability:public_source_retrieval", "evidence:public_source_summary", "summary"),
	)
	graph := ObjectiveGraph{
		GraphRef:   "graph:public_source",
		CatalogRef: "catalog:objective_graph",
		Nodes: []ObjectiveNode{
			objectiveGraphTestNode("node:fetch", "strategy:public_source", "capability:public_source_retrieval", "evidence:public_source_summary", "summary"),
			func() ObjectiveNode {
				node := objectiveGraphTestNode("node:optional_alt", "strategy:missing_optional", "capability:missing_optional", "evidence:optional_alt", "summary")
				node.Optional = true
				return node
			}(),
		},
	}
	raw, err := json.Marshal(graph)
	if err != nil {
		t.Fatal(err)
	}

	got := BuildObjectiveGraphWithPlanner(context.Background(), ObjectiveGraphBuildInput{
		Enabled: true,
		Planner: ObjectiveGraphPlannerFunc(func(context.Context, ObjectiveGraphPlannerRequest) (ObjectiveGraphPlannerResponse, error) {
			return ObjectiveGraphPlannerResponse{
				ResponseRef: "response:graph_json",
				GraphJSON:   raw,
			}, nil
		}),
		Request: ObjectiveGraphPlannerRequest{
			RequestRef:            "request:graph",
			Spec:                  spec,
			Catalog:               catalog,
			CapabilityDescriptors: objectiveGraphTestDescriptors("capability:public_source_retrieval"),
			AllowedCapabilityRefs: []DisplaySafeRef{"capability:public_source_retrieval", "capability:missing_optional"},
			Policy:                objectiveGraphTestPolicy(),
			RawOutputLoaded:       false,
			Boundaries:            []Boundary{"test_planner_request"},
			PolicyRefs:            []DisplaySafeRef{"policy:objective_graph_test"},
		},
		GraphRef:  "graph:planner_default",
		SourceRef: "host:objective_graph_planner_test",
	})
	if got.Status != VerificationSatisfied ||
		!got.Built ||
		!got.DecodeAttempted ||
		got.Validation.SkippedNodeCount != 1 ||
		got.Validation.ReadyNodeCount != 1 ||
		got.Graph.Nodes[1].State != ObjectiveNodeStateSkipped {
		t.Fatalf("unexpected planner graph = %#v", got)
	}
	if got.PlannerCalled != true ||
		got.RunnerEffect != "none" ||
		got.PromptEffect != "host_planner_interface_only" {
		t.Fatalf("planner boundary/effects = %#v", got)
	}
}

func TestBuildObjectiveGraphValidationBlocksMissingCapability(t *testing.T) {
	spec := objectiveGraphTestSpec([]DisplaySafeRef{"capability:known"}, ObjectiveSpecBudget{BudgetRef: "budget:missing_cap", MaxNodes: 2, MaxAttempts: 2})
	graph := ObjectiveGraph{
		GraphRef:   "graph:missing_capability",
		CatalogRef: "catalog:objective_graph",
		Nodes: []ObjectiveNode{
			objectiveGraphTestNode("node:missing", "strategy:missing", "capability:missing", "evidence:postgres_schema", "database_schema"),
		},
	}
	got := BuildObjectiveGraphValidation(ObjectiveGraphValidationInput{
		Graph:   graph,
		Spec:    spec,
		Catalog: objectiveGraphTestCatalog(objectiveGraphTestCatalogEntry("strategy:known", "capability:known", "evidence:postgres_schema", "database_schema")),
		Policy:  objectiveGraphTestPolicy(),
	})
	if got.Status != VerificationBlocked ||
		got.ReadyForRuntimeLoop ||
		got.FailureClass != FailureCapabilityMissing ||
		!objectiveGraphTestMissingContains(got.MissingInputs, "host:available_capability:missing") ||
		!objectiveGraphTestStringContains(got.BlockedReasons, "objective_graph_node_capability_missing") {
		t.Fatalf("missing capability report = %#v", got)
	}
}

func TestBuildObjectiveGraphValidationBlocksMissingNodeContracts(t *testing.T) {
	spec := objectiveGraphTestSpec([]DisplaySafeRef{"capability:known"}, ObjectiveSpecBudget{BudgetRef: "budget:node_contracts", MaxNodes: 1, MaxAttempts: 1})
	node := objectiveGraphTestNode("node:contractless", "strategy:known", "capability:known", "evidence:known", "known_result")
	node.InputSchemaRef = ""
	node.OutputSchemaRef = ""
	node.EvidenceContractRef = ""
	graph := ObjectiveGraph{
		GraphRef:   "graph:node_contracts",
		CatalogRef: "catalog:objective_graph",
		Nodes:      []ObjectiveNode{node},
	}
	got := BuildObjectiveGraphValidation(ObjectiveGraphValidationInput{
		Graph:   graph,
		Spec:    spec,
		Catalog: objectiveGraphTestCatalog(objectiveGraphTestCatalogEntry("strategy:known", "capability:known", "evidence:known", "known_result")),
		Policy:  objectiveGraphTestPolicy(),
	})
	if got.Status != VerificationBlocked ||
		got.ReadyForRuntimeLoop ||
		got.FailureClass != FailureInvalidInput ||
		got.NextHostAction != "revise_objective_graph" ||
		!objectiveGraphTestMissingContains(got.MissingInputs, "host:objective_graph_node_input_schema_ref") ||
		!objectiveGraphTestMissingContains(got.MissingInputs, "host:objective_graph_node_output_schema_ref") ||
		!objectiveGraphTestMissingContains(got.MissingInputs, "host:objective_graph_node_evidence_contract_ref") ||
		!objectiveGraphTestStringContains(got.BlockedReasons, "objective_graph_node_incomplete") {
		t.Fatalf("missing node contract report = %#v", got)
	}
}

func TestBuildObjectiveGraphValidationBlocksSideEffectWithoutApproval(t *testing.T) {
	spec := objectiveGraphTestSpec([]DisplaySafeRef{"capability:purchase"}, ObjectiveSpecBudget{BudgetRef: "budget:side_effect", MaxNodes: 1, MaxAttempts: 1})
	spec.SideEffectPolicy = ObjectiveSpecSideEffectReadOnly
	node := objectiveGraphTestNode("node:purchase", "strategy:purchase", "capability:purchase", "evidence:postgres_schema", "database_schema")
	node.SideEffectClass = ObjectiveCapabilitySideEffectExternalWrite
	graph := ObjectiveGraph{
		GraphRef:   "graph:side_effect",
		CatalogRef: "catalog:objective_graph",
		Nodes:      []ObjectiveNode{node},
	}
	got := BuildObjectiveGraphValidation(ObjectiveGraphValidationInput{
		Graph:   graph,
		Spec:    spec,
		Catalog: objectiveGraphTestCatalog(objectiveGraphTestCatalogEntry("strategy:purchase", "capability:purchase", "evidence:postgres_schema", "database_schema")),
		Policy:  objectiveGraphTestPolicy(),
	})
	if got.Status != VerificationBlocked ||
		got.FailureClass != FailurePolicyBlocked ||
		got.NextHostAction != "request_host_approval" ||
		!objectiveGraphTestMissingContains(got.MissingInputs, "contract:objective_graph_side_effect") {
		t.Fatalf("side effect report = %#v", got)
	}
}

func TestBuildObjectiveGraphValidationBlocksCycleAndBudget(t *testing.T) {
	spec := objectiveGraphTestSpec([]DisplaySafeRef{"capability:a", "capability:b"}, ObjectiveSpecBudget{BudgetRef: "budget:cycle", MaxNodes: 3, MaxAttempts: 4})
	catalog := objectiveGraphTestCatalog(
		objectiveGraphTestCatalogEntry("strategy:a", "capability:a", "evidence:postgres_schema", "database_schema"),
		objectiveGraphTestCatalogEntry("strategy:b", "capability:b", "evidence:postgres_schema", "database_schema"),
	)
	cyclic := ObjectiveGraph{
		GraphRef:   "graph:cycle",
		CatalogRef: "catalog:objective_graph",
		Nodes: []ObjectiveNode{
			objectiveGraphTestNodeWithDeps("node:a", "strategy:a", "capability:a", "evidence:postgres_schema", "database_schema", []ObjectiveNodeDependency{{DependencyRef: "dependency:a_b", NodeRef: "node:b"}}),
			objectiveGraphTestNodeWithDeps("node:b", "strategy:b", "capability:b", "evidence:postgres_schema", "database_schema", []ObjectiveNodeDependency{{DependencyRef: "dependency:b_a", NodeRef: "node:a"}}),
		},
	}
	got := BuildObjectiveGraphValidation(ObjectiveGraphValidationInput{
		Graph:   cyclic,
		Spec:    spec,
		Catalog: catalog,
		Policy:  objectiveGraphTestPolicy(),
	})
	if got.Status != VerificationBlocked ||
		got.FailureClass != FailureInvalidInput ||
		!objectiveGraphTestMissingContains(got.MissingInputs, "host:objective_graph_acyclic") ||
		!objectiveGraphTestStringContains(got.BlockedReasons, "objective_graph_cycle_detected") {
		t.Fatalf("cycle report = %#v", got)
	}

	overBudgetSpec := objectiveGraphTestSpec([]DisplaySafeRef{"capability:a", "capability:b"}, ObjectiveSpecBudget{BudgetRef: "budget:small", MaxNodes: 1, MaxAttempts: 4})
	overBudget := cyclic
	overBudget.Nodes[0].Dependencies = nil
	overBudget.Nodes[1].Dependencies = nil
	got = BuildObjectiveGraphValidation(ObjectiveGraphValidationInput{
		Graph:   overBudget,
		Spec:    overBudgetSpec,
		Catalog: catalog,
		Policy:  objectiveGraphTestPolicy(),
	})
	if got.Status != VerificationBlocked ||
		got.FailureClass != FailureBudgetExhausted ||
		!objectiveGraphTestMissingContains(got.MissingInputs, "contract:objective_graph_max_nodes") {
		t.Fatalf("budget report = %#v", got)
	}
}

func TestBuildObjectiveGraphFromJSONRejectsUnknownFieldsAndAllowedCapabilityMiss(t *testing.T) {
	spec := objectiveGraphTestSpec([]DisplaySafeRef{"capability:public_source_retrieval"}, ObjectiveSpecBudget{BudgetRef: "budget:json", MaxNodes: 2, MaxAttempts: 2})
	unknown := BuildObjectiveGraphFromJSON(ObjectiveGraphJSONDecodeInput{
		RawJSON: []byte(`{"graph_ref":"graph:bad","unexpected":true}`),
		Spec:    spec,
		Catalog: objectiveGraphTestCatalog(),
	})
	if unknown.Status != VerificationBlocked ||
		unknown.FailureClass != FailureInvalidInput ||
		!objectiveGraphTestBoundaryContains(unknown.Boundaries, "objective_graph_json_decode_failed") {
		t.Fatalf("unknown field report = %#v", unknown)
	}

	graph := ObjectiveGraph{
		GraphRef:   "graph:disallowed",
		CatalogRef: "catalog:objective_graph",
		Nodes: []ObjectiveNode{
			objectiveGraphTestNode("node:fetch", "strategy:public_source", "capability:public_source_retrieval", "evidence:public_source_summary", "summary"),
		},
	}
	raw, err := json.Marshal(graph)
	if err != nil {
		t.Fatal(err)
	}
	disallowed := BuildObjectiveGraphFromJSON(ObjectiveGraphJSONDecodeInput{
		RawJSON:               raw,
		Spec:                  spec,
		Catalog:               objectiveGraphTestCatalog(objectiveGraphTestCatalogEntry("strategy:public_source", "capability:public_source_retrieval", "evidence:public_source_summary", "summary")),
		AllowedCapabilityRefs: []DisplaySafeRef{"capability:other"},
		Policy:                objectiveGraphTestPolicy(),
	})
	if disallowed.Status != VerificationBlocked ||
		disallowed.FailureClass != FailureCapabilityMissing ||
		!objectiveGraphTestMissingContains(disallowed.MissingInputs, "host:allowed_capability_ref") {
		t.Fatalf("allowed capability report = %#v", disallowed)
	}
}

func TestBuildObjectiveGraphWithPlannerFallbacks(t *testing.T) {
	spec := objectiveGraphTestSpec([]DisplaySafeRef{"capability:public_source_retrieval"}, ObjectiveSpecBudget{BudgetRef: "budget:fallback", MaxNodes: 1, MaxAttempts: 1})
	called := false
	disabled := BuildObjectiveGraphWithPlanner(context.Background(), ObjectiveGraphBuildInput{
		Enabled: false,
		Planner: ObjectiveGraphPlannerFunc(func(context.Context, ObjectiveGraphPlannerRequest) (ObjectiveGraphPlannerResponse, error) {
			called = true
			return ObjectiveGraphPlannerResponse{}, nil
		}),
		Request: ObjectiveGraphPlannerRequest{RequestRef: "request:disabled", Spec: spec},
	})
	if disabled.Status != VerificationBlocked ||
		disabled.FailureClass != FailureInsufficientInformation ||
		called {
		t.Fatalf("disabled report = %#v called=%v", disabled, called)
	}

	failed := BuildObjectiveGraphWithPlanner(context.Background(), ObjectiveGraphBuildInput{
		Enabled: true,
		Planner: ObjectiveGraphPlannerFunc(func(context.Context, ObjectiveGraphPlannerRequest) (ObjectiveGraphPlannerResponse, error) {
			return ObjectiveGraphPlannerResponse{}, errors.New("planner failed with raw backend details")
		}),
		Request: ObjectiveGraphPlannerRequest{RequestRef: "request:failed", Spec: spec},
	})
	if failed.Status != VerificationBlocked ||
		failed.FailureClass != FailureExternalDependencyUnavailable ||
		!objectiveGraphTestBoundaryContains(failed.Boundaries, "deterministic_blocked_fallback") ||
		objectiveGraphTestBoundaryContains(failed.Boundaries, "raw_backend_details") {
		t.Fatalf("failed report = %#v", failed)
	}
}

func objectiveGraphTestSpec(capabilities []DisplaySafeRef, budget ObjectiveSpecBudget) ObjectiveSpec {
	return ObjectiveSpec{
		SpecRef:               "spec:objective_graph",
		ObjectiveID:           "objective:objective_graph",
		UserGoalDigest:        "sha256:objective_graph",
		RawGoalRef:            "goal:objective_graph",
		GoalSummary:           "inspect objective graph capability result",
		ControlMode:           ControlModeObjective,
		Intensity:             IntensityL3ManagedObjective,
		CandidateCapabilities: normalizeDisplaySafeRefs(capabilities),
		SuccessCriteria: []ObjectiveSuccessCriterion{{
			CriteriaRef: "criteria:objective_graph",
			Text:        "collect required graph evidence",
			RequiredEvidence: []EvidenceRef{{
				Ref:      "evidence:postgres_schema",
				Kind:     "database_schema",
				Strength: EvidenceAdequate,
				Source:   "source:objective_graph",
			}},
		}},
		RequiredEvidence: []EvidenceRef{{
			Ref:      "evidence:postgres_schema",
			Kind:     "database_schema",
			Strength: EvidenceAdequate,
			Source:   "source:objective_graph",
		}},
		SideEffectPolicy:  ObjectiveSpecSideEffectReadOnly,
		MissingInfoPolicy: ObjectiveSpecMissingInfoAskUser,
		Budget:            budget,
		PolicyRefs:        []DisplaySafeRef{"policy:display_safe"},
	}
}

func objectiveGraphTestCatalog(entries ...StrategyCatalogEntry) StrategyCatalogSnapshot {
	return StrategyCatalogSnapshot{
		CatalogRef: "catalog:objective_graph",
		Entries:    entries,
		PolicyRefs: []DisplaySafeRef{"policy:objective_graph_catalog"},
	}
}

func objectiveGraphTestCatalogEntry(strategyRef string, capabilityRef DisplaySafeRef, evidenceRef DisplaySafeRef, evidenceKind string) StrategyCatalogEntry {
	return StrategyCatalogEntry{
		SourceKind: StrategyCatalogSourceHostAdapter,
		SourceRef:  DisplaySafeRef("source:" + normalizeControlToken(strategyRef)),
		Status:     VerificationSatisfied,
		Candidate: StrategyCandidate{
			ID:              strategyRef,
			Kind:            "host_adapter",
			ControlMode:     ControlModeObjective,
			MinIntensity:    IntensityL3ManagedObjective,
			CapabilityRefs:  []DisplaySafeRef{capabilityRef},
			SideEffectClass: string(ObjectiveCapabilitySideEffectReadOnly),
			ExpectedEvidence: []EvidenceRef{{
				Ref:      evidenceRef,
				Kind:     evidenceKind,
				Strength: EvidenceAdequate,
				Source:   "source:objective_graph",
			}},
			Owner: "host",
		},
	}
}

func objectiveGraphTestNode(nodeRef DisplaySafeRef, strategyRef DisplaySafeRef, capabilityRef DisplaySafeRef, evidenceRef DisplaySafeRef, evidenceKind string) ObjectiveNode {
	return objectiveGraphTestNodeWithDeps(nodeRef, strategyRef, capabilityRef, evidenceRef, evidenceKind, nil)
}

func objectiveGraphTestNodeWithDeps(nodeRef DisplaySafeRef, strategyRef DisplaySafeRef, capabilityRef DisplaySafeRef, evidenceRef DisplaySafeRef, evidenceKind string, deps []ObjectiveNodeDependency) ObjectiveNode {
	return ObjectiveNode{
		NodeRef:             nodeRef,
		Kind:                "host_adapter",
		CapabilityRef:       capabilityRef,
		StrategyRef:         strategyRef,
		DescriptorRef:       DisplaySafeRef("descriptor:" + normalizeControlToken(string(capabilityRef))),
		SourceRef:           "source:objective_graph",
		InputSchemaRef:      "schema:objective_graph.input.v1",
		OutputSchemaRef:     "schema:objective_graph.output.v1",
		EvidenceContractRef: "evidence:objective_graph.contract.v1",
		RequiredEvidence: []EvidenceRef{{
			Ref:      evidenceRef,
			Kind:     evidenceKind,
			Strength: EvidenceAdequate,
			Source:   "source:objective_graph",
		}},
		Dependencies:    deps,
		AttemptPolicy:   ObjectiveNodeAttemptPolicy{MaxAttempts: 1, TimeoutSeconds: 30, NoProgressGate: true},
		SideEffectClass: ObjectiveCapabilitySideEffectReadOnly,
		PolicyRefs:      []DisplaySafeRef{"policy:objective_graph_node"},
	}
}

func objectiveGraphTestDescriptors(capabilities ...DisplaySafeRef) []ObjectiveCapabilityDescriptor {
	out := make([]ObjectiveCapabilityDescriptor, 0, len(capabilities))
	for _, capability := range capabilities {
		capability = normalizeOneDisplaySafeRef(capability)
		if capability == "" {
			continue
		}
		token := normalizeControlToken(string(capability))
		out = append(out, ObjectiveCapabilityDescriptor{
			DescriptorRef:             DisplaySafeRef("descriptor:" + token),
			CapabilityRef:             capability,
			StrategyRef:               DisplaySafeRef("strategy:" + token),
			SourceKind:                StrategyCatalogSourceHostAdapter,
			SourceRef:                 "source:objective_graph",
			OwnerRef:                  "owner:objective_graph",
			ProviderRef:               "provider:objective_graph",
			StrategyKind:              "host_adapter",
			ControlMode:               ControlModeObjective,
			MinIntensity:              IntensityL3ManagedObjective,
			MaxIntensity:              IntensityL3ManagedObjective,
			InputSchemaRef:            "schema:objective_graph.input.v1",
			OutputSchemaRef:           "schema:objective_graph.output.v1",
			EvidenceContractRef:       "evidence:objective_graph.contract.v1",
			RequiredEvidence:          []EvidenceRef{{Ref: "evidence:postgres_schema", Kind: "database_schema", Strength: EvidenceAdequate, Source: "source:objective_graph"}},
			SideEffectClass:           ObjectiveCapabilitySideEffectReadOnly,
			CredentialRequirementRefs: []DisplaySafeRef{"auth:none"},
			ConfigRequirementRefs:     []DisplaySafeRef{"config:objective_graph"},
			FailureClasses:            []FailureClass{FailureTargetUnavailable},
			ExampleRefs:               []DisplaySafeRef{"example:objective_graph"},
			VerificationHintRefs:      []DisplaySafeRef{"verification:objective_graph"},
		})
	}
	return out
}

func objectiveGraphTestPolicy() ExecutionIntensityPolicy {
	return ExecutionIntensityPolicy{
		PolicyRef:            "policy:objective_graph",
		ExecutionContractRef: "contract:objective_graph",
		Activation:           ActivationManaged,
		DefaultIntensity:     IntensityL3ManagedObjective,
		MaxDefaultIntensity:  IntensityL3ManagedObjective,
		MaxAllowedIntensity:  IntensityL3ManagedObjective,
		AllowedControlModesByIntensity: map[ExecutionIntensity][]ControlMode{
			IntensityL3ManagedObjective: {ControlModeObjective},
		},
		PolicyRefs: []DisplaySafeRef{"policy:objective_graph"},
	}
}

func objectiveGraphTestBoundaryContains(values []Boundary, want Boundary) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func objectiveGraphTestMissingContains(values []MissingInput, want MissingInput) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func objectiveGraphTestStringContains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
