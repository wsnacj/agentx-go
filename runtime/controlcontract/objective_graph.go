package controlcontract

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
)

type ObjectiveNodeState string

const (
	ObjectiveNodeStateUnspecified     ObjectiveNodeState = "unspecified"
	ObjectiveNodeStatePending         ObjectiveNodeState = "pending"
	ObjectiveNodeStateReady           ObjectiveNodeState = "ready"
	ObjectiveNodeStateRunning         ObjectiveNodeState = "running"
	ObjectiveNodeStateSatisfied       ObjectiveNodeState = "satisfied"
	ObjectiveNodeStatePartial         ObjectiveNodeState = "partial"
	ObjectiveNodeStateBlocked         ObjectiveNodeState = "blocked"
	ObjectiveNodeStateFailedRetryable ObjectiveNodeState = "failed_retryable"
	ObjectiveNodeStateSkipped         ObjectiveNodeState = "skipped"
)

func NormalizeObjectiveNodeState(raw string) ObjectiveNodeState {
	switch normalizeEnumToken(raw) {
	case "pending", "waiting":
		return ObjectiveNodeStatePending
	case "ready", "ready_to_run":
		return ObjectiveNodeStateReady
	case "running", "in_progress":
		return ObjectiveNodeStateRunning
	case "satisfied", "success", "completed":
		return ObjectiveNodeStateSatisfied
	case "partial", "partially_satisfied":
		return ObjectiveNodeStatePartial
	case "blocked", "cannot_continue":
		return ObjectiveNodeStateBlocked
	case "failed_retryable", "retryable_failed", "retryable_failure":
		return ObjectiveNodeStateFailedRetryable
	case "skipped", "optional_skipped":
		return ObjectiveNodeStateSkipped
	default:
		return ObjectiveNodeStateUnspecified
	}
}

type ObjectiveNodeDependency struct {
	DependencyRef DisplaySafeRef `json:"dependency_ref,omitempty"`
	NodeRef       DisplaySafeRef `json:"node_ref,omitempty"`
	EvidenceRefs  []EvidenceRef  `json:"evidence_refs,omitempty"`
	Optional      bool           `json:"optional"`
}

func CloneObjectiveNodeDependency(in ObjectiveNodeDependency) ObjectiveNodeDependency {
	out := in
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	return out
}

func (d ObjectiveNodeDependency) Clone() ObjectiveNodeDependency {
	return CloneObjectiveNodeDependency(d)
}

func (d ObjectiveNodeDependency) Normalize() ObjectiveNodeDependency {
	out := CloneObjectiveNodeDependency(d)
	out.DependencyRef = normalizeOneDisplaySafeRef(out.DependencyRef)
	out.NodeRef = normalizeOneDisplaySafeRef(out.NodeRef)
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
	return out
}

type ObjectiveNodeAttemptPolicy struct {
	MaxAttempts             int              `json:"max_attempts,omitempty"`
	TimeoutSeconds          int              `json:"timeout_seconds,omitempty"`
	RetryableFailureClasses []FailureClass   `json:"retryable_failure_classes,omitempty"`
	NoProgressGate          bool             `json:"no_progress_gate"`
	PolicyRefs              []DisplaySafeRef `json:"policy_refs,omitempty"`
}

func CloneObjectiveNodeAttemptPolicy(in ObjectiveNodeAttemptPolicy) ObjectiveNodeAttemptPolicy {
	out := in
	out.RetryableFailureClasses = cloneFailureClasses(in.RetryableFailureClasses)
	out.PolicyRefs = cloneDisplaySafeRefs(in.PolicyRefs)
	return out
}

func (p ObjectiveNodeAttemptPolicy) Clone() ObjectiveNodeAttemptPolicy {
	return CloneObjectiveNodeAttemptPolicy(p)
}

func (p ObjectiveNodeAttemptPolicy) Normalize() ObjectiveNodeAttemptPolicy {
	out := CloneObjectiveNodeAttemptPolicy(p)
	if out.MaxAttempts <= 0 {
		out.MaxAttempts = 1
	}
	if out.TimeoutSeconds < 0 {
		out.TimeoutSeconds = 0
	}
	out.RetryableFailureClasses = normalizeFailureClasses(out.RetryableFailureClasses)
	out.PolicyRefs = normalizeDisplaySafeRefs(out.PolicyRefs)
	return out
}

type ObjectiveNode struct {
	NodeRef             DisplaySafeRef                     `json:"node_ref,omitempty"`
	Kind                string                             `json:"kind,omitempty"`
	State               ObjectiveNodeState                 `json:"state,omitempty"`
	Optional            bool                               `json:"optional"`
	CapabilityRef       DisplaySafeRef                     `json:"capability_ref,omitempty"`
	StrategyRef         DisplaySafeRef                     `json:"strategy_ref,omitempty"`
	DescriptorRef       DisplaySafeRef                     `json:"descriptor_ref,omitempty"`
	SourceRef           DisplaySafeRef                     `json:"source_ref,omitempty"`
	InputSchemaRef      DisplaySafeRef                     `json:"input_schema_ref,omitempty"`
	OutputSchemaRef     DisplaySafeRef                     `json:"output_schema_ref,omitempty"`
	EvidenceContractRef DisplaySafeRef                     `json:"evidence_contract_ref,omitempty"`
	RequiredEvidence    []EvidenceRef                      `json:"required_evidence,omitempty"`
	Dependencies        []ObjectiveNodeDependency          `json:"dependencies,omitempty"`
	AttemptPolicy       ObjectiveNodeAttemptPolicy         `json:"attempt_policy,omitempty"`
	SideEffectClass     ObjectiveCapabilitySideEffectClass `json:"side_effect_class,omitempty"`
	RequiresApproval    bool                               `json:"requires_approval"`
	ApprovalRefs        []DisplaySafeRef                   `json:"approval_refs,omitempty"`
	PolicyRefs          []DisplaySafeRef                   `json:"policy_refs,omitempty"`
	Boundaries          []Boundary                         `json:"boundaries,omitempty"`
	MissingInputs       []MissingInput                     `json:"missing_inputs,omitempty"`
	RawOutputLoaded     bool                               `json:"raw_output_loaded"`
}

func CloneObjectiveNode(in ObjectiveNode) ObjectiveNode {
	out := in
	out.RequiredEvidence = cloneEvidenceRefs(in.RequiredEvidence)
	out.Dependencies = cloneObjectiveNodeDependencies(in.Dependencies)
	out.AttemptPolicy = in.AttemptPolicy.Clone()
	out.ApprovalRefs = cloneDisplaySafeRefs(in.ApprovalRefs)
	out.PolicyRefs = cloneDisplaySafeRefs(in.PolicyRefs)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	return out
}

func (n ObjectiveNode) Clone() ObjectiveNode {
	return CloneObjectiveNode(n)
}

func (n ObjectiveNode) Normalize() ObjectiveNode {
	out := CloneObjectiveNode(n)
	out.NodeRef = normalizeOneDisplaySafeRef(out.NodeRef)
	out.Kind = normalizeControlToken(out.Kind)
	out.State = NormalizeObjectiveNodeState(string(out.State))
	if out.State == ObjectiveNodeStateUnspecified {
		out.State = ObjectiveNodeStatePending
	}
	out.CapabilityRef = normalizeOneDisplaySafeRef(out.CapabilityRef)
	out.StrategyRef = normalizeOneDisplaySafeRef(out.StrategyRef)
	out.DescriptorRef = normalizeOneDisplaySafeRef(out.DescriptorRef)
	out.SourceRef = normalizeOneDisplaySafeRef(out.SourceRef)
	out.InputSchemaRef = normalizeOneDisplaySafeRef(out.InputSchemaRef)
	out.OutputSchemaRef = normalizeOneDisplaySafeRef(out.OutputSchemaRef)
	out.EvidenceContractRef = normalizeOneDisplaySafeRef(out.EvidenceContractRef)
	out.RequiredEvidence = normalizeEvidenceRefs(out.RequiredEvidence)
	out.Dependencies = normalizeObjectiveNodeDependencies(out.Dependencies)
	out.AttemptPolicy = out.AttemptPolicy.Normalize()
	out.SideEffectClass = NormalizeObjectiveCapabilitySideEffectClass(string(out.SideEffectClass))
	out.ApprovalRefs = normalizeDisplaySafeRefs(out.ApprovalRefs)
	out.PolicyRefs = normalizeDisplaySafeRefs(out.PolicyRefs)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	return out
}

type ObjectiveGraph struct {
	ContractVersion  string              `json:"contract_version,omitempty"`
	GraphRef         DisplaySafeRef      `json:"graph_ref,omitempty"`
	SpecRef          DisplaySafeRef      `json:"spec_ref,omitempty"`
	ObjectiveID      string              `json:"objective_id,omitempty"`
	CatalogRef       DisplaySafeRef      `json:"catalog_ref,omitempty"`
	Nodes            []ObjectiveNode     `json:"nodes,omitempty"`
	RequiredEvidence []EvidenceRef       `json:"required_evidence,omitempty"`
	Budget           ObjectiveSpecBudget `json:"budget,omitempty"`
	PolicyRefs       []DisplaySafeRef    `json:"policy_refs,omitempty"`
	Boundaries       []Boundary          `json:"boundaries,omitempty"`
	MissingInputs    []MissingInput      `json:"missing_inputs,omitempty"`
	RawOutputLoaded  bool                `json:"raw_output_loaded"`
}

func CloneObjectiveGraph(in ObjectiveGraph) ObjectiveGraph {
	out := in
	out.Nodes = cloneObjectiveNodes(in.Nodes)
	out.RequiredEvidence = cloneEvidenceRefs(in.RequiredEvidence)
	out.Budget = in.Budget.Clone()
	out.PolicyRefs = cloneDisplaySafeRefs(in.PolicyRefs)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	return out
}

func (g ObjectiveGraph) Clone() ObjectiveGraph {
	return CloneObjectiveGraph(g)
}

func (g ObjectiveGraph) Normalize() ObjectiveGraph {
	out := CloneObjectiveGraph(g)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.GraphRef = normalizeOneDisplaySafeRef(out.GraphRef)
	out.SpecRef = normalizeOneDisplaySafeRef(out.SpecRef)
	out.ObjectiveID = objectiveGraphSafeID(out.ObjectiveID)
	out.CatalogRef = normalizeOneDisplaySafeRef(out.CatalogRef)
	out.Nodes = normalizeObjectiveNodes(out.Nodes)
	out.RequiredEvidence = normalizeEvidenceRefs(out.RequiredEvidence)
	out.Budget = out.Budget.Normalize()
	out.PolicyRefs = normalizeDisplaySafeRefs(out.PolicyRefs)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	return out
}

type ObjectiveGraphNodeValidation struct {
	ContractVersion string             `json:"contract_version,omitempty"`
	NodeRef         DisplaySafeRef     `json:"node_ref,omitempty"`
	Status          VerificationStatus `json:"status,omitempty"`
	State           ObjectiveNodeState `json:"state,omitempty"`
	FailureClass    FailureClass       `json:"failure_class,omitempty"`
	MissingInputs   []MissingInput     `json:"missing_inputs,omitempty"`
	BlockedReasons  []string           `json:"blocked_reasons,omitempty"`
	Boundaries      []Boundary         `json:"boundaries,omitempty"`
	NextHostAction  NextHostAction     `json:"next_host_action,omitempty"`
}

func CloneObjectiveGraphNodeValidation(in ObjectiveGraphNodeValidation) ObjectiveGraphNodeValidation {
	out := in
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (v ObjectiveGraphNodeValidation) Clone() ObjectiveGraphNodeValidation {
	return CloneObjectiveGraphNodeValidation(v)
}

func (v ObjectiveGraphNodeValidation) Normalize() ObjectiveGraphNodeValidation {
	out := v.Clone()
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.NodeRef = normalizeOneDisplaySafeRef(out.NodeRef)
	out.Status = NormalizeVerificationStatus(string(out.Status))
	if out.Status == VerificationNotEvaluated {
		out.Status = VerificationBlocked
	}
	out.State = NormalizeObjectiveNodeState(string(out.State))
	if out.State == ObjectiveNodeStateUnspecified {
		out.State = ObjectiveNodeStateBlocked
	}
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.BlockedReasons = normalizeControlTokenList(out.BlockedReasons)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	return out
}

type ObjectiveGraphValidationInput struct {
	Graph                 ObjectiveGraph                  `json:"graph,omitempty"`
	Spec                  ObjectiveSpec                   `json:"spec,omitempty"`
	Catalog               StrategyCatalogSnapshot         `json:"catalog,omitempty"`
	CapabilityDescriptors []ObjectiveCapabilityDescriptor `json:"capability_descriptors,omitempty"`
	Policy                ExecutionIntensityPolicy        `json:"policy,omitempty"`
	SourceRef             DisplaySafeRef                  `json:"source_ref,omitempty"`
	Boundaries            []Boundary                      `json:"boundaries,omitempty"`
	RawOutputLoaded       bool                            `json:"raw_output_loaded"`
}

type ObjectiveGraphValidationReport struct {
	ContractVersion     string                         `json:"contract_version,omitempty"`
	Validated           bool                           `json:"validated"`
	Available           bool                           `json:"available"`
	Status              VerificationStatus             `json:"status,omitempty"`
	Mode                string                         `json:"mode,omitempty"`
	RunnerEffect        string                         `json:"runner_effect,omitempty"`
	PromptEffect        string                         `json:"prompt_effect,omitempty"`
	SourceRef           DisplaySafeRef                 `json:"source_ref,omitempty"`
	Graph               ObjectiveGraph                 `json:"graph,omitempty"`
	NodeValidations     []ObjectiveGraphNodeValidation `json:"node_validations,omitempty"`
	ReadyForRuntimeLoop bool                           `json:"ready_for_runtime_loop"`
	ReadyNodeCount      int                            `json:"ready_node_count,omitempty"`
	SkippedNodeCount    int                            `json:"skipped_node_count,omitempty"`
	FailureClass        FailureClass                   `json:"failure_class,omitempty"`
	MissingInputs       []MissingInput                 `json:"missing_inputs,omitempty"`
	BlockedReasons      []string                       `json:"blocked_reasons,omitempty"`
	Boundaries          []Boundary                     `json:"boundaries,omitempty"`
	NextHostAction      NextHostAction                 `json:"next_host_action,omitempty"`
	RawOutputLoaded     bool                           `json:"raw_output_loaded"`
}

func BuildObjectiveGraphValidation(input ObjectiveGraphValidationInput) ObjectiveGraphValidationReport {
	result := baseObjectiveGraphValidationReport(input)
	if objectiveGraphValidationInputUnsafe(input) {
		return objectiveGraphValidationBlock(result, VerificationReviewRequired, FailureEvidenceWeak, "host:display_safe_refs", "provide_display_safe_refs", "raw_output_not_allowed")
	}

	specProjection := BuildObjectiveSpecFrameProjection(ObjectiveSpecFrameProjectionInput{
		Spec:          input.Spec,
		ProjectionRef: firstDisplaySafeRef(input.Graph.GraphRef, "projection:objective_graph_spec"),
		SourceRef:     firstDisplaySafeRef(input.SourceRef, "host:objective_graph_validator"),
		Boundaries:    []Boundary{"objective_graph_spec_projection"},
	})
	if specProjection.Status != VerificationSatisfied {
		result.MissingInputs = MergeMissingInputs(result.MissingInputs, specProjection.MissingInputs)
		result.Boundaries = MergeBoundaries(result.Boundaries, specProjection.Boundaries, []Boundary{"objective_graph_requires_ready_objective_spec"})
		result.FailureClass = firstFailureClass(specProjection.FailureClass, FailureInvalidInput)
		result.NextHostAction = firstNextHostAction(specProjection.NextHostAction, "provide_objective_spec")
		result.Status = specProjection.Status
		return result.Normalize()
	}

	graph := objectiveGraphApplyDefaults(result.Graph, specProjection.Spec, input.Catalog).Normalize()
	result.Graph = graph
	result.MissingInputs = MergeMissingInputs(result.MissingInputs, objectiveGraphMissingInputs(graph, input.Catalog))
	if len(result.MissingInputs) > 0 {
		result.FailureClass = objectiveGraphFailure(result.MissingInputs)
		result.NextHostAction = objectiveGraphNextAction(result.MissingInputs)
		result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, "objective_graph_incomplete")
		result.Boundaries = AppendBoundaries(result.Boundaries, "objective_graph_incomplete")
		return result.Normalize()
	}
	if graph.Budget.MaxNodes > 0 && len(graph.Nodes) > graph.Budget.MaxNodes {
		return objectiveGraphValidationBlock(result, VerificationBlocked, FailureBudgetExhausted, "contract:objective_graph_max_nodes", "reduce_objective_graph_scope", "objective_graph_node_budget_exceeded")
	}
	if graph.Budget.MaxAttempts > 0 && objectiveGraphAttemptBudget(graph) > graph.Budget.MaxAttempts {
		return objectiveGraphValidationBlock(result, VerificationBlocked, FailureBudgetExhausted, "contract:objective_graph_max_attempts", "reduce_objective_graph_attempts", "objective_graph_attempt_budget_exceeded")
	}
	if missing := objectiveGraphMissingSpecEvidence(specProjection.Spec, graph); len(missing) > 0 {
		result.MissingInputs = MergeMissingInputs(result.MissingInputs, missing)
		result.FailureClass = FailureEvidenceMissing
		result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, "objective_graph_required_evidence_missing")
		result.Boundaries = AppendBoundaries(result.Boundaries, "objective_graph_required_evidence_missing")
		result.NextHostAction = "revise_objective_graph"
		return result.Normalize()
	}
	if cycle := objectiveGraphCycleNodeRef(graph); cycle != "" {
		result.MissingInputs = AppendMissingInputs(result.MissingInputs, "host:objective_graph_acyclic")
		result.FailureClass = FailureInvalidInput
		result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, "objective_graph_cycle_detected")
		result.Boundaries = AppendBoundaries(result.Boundaries, "objective_graph_cycle_detected", Boundary("cycle_node:"+string(cycle)))
		result.NextHostAction = "revise_objective_graph"
		return result.Normalize()
	}

	indices := objectiveGraphValidationIndices{
		Catalog:     input.Catalog.Normalize(),
		Descriptors: normalizeObjectiveCapabilityDescriptors(input.CapabilityDescriptors),
		Policy:      input.Policy.Normalize(),
		Spec:        specProjection.Spec.Normalize(),
	}
	result.FailureClass = FailureNone
	nodeValidations := make([]ObjectiveGraphNodeValidation, 0, len(graph.Nodes))
	validatedNodes := make([]ObjectiveNode, 0, len(graph.Nodes))
	for _, node := range graph.Nodes {
		validation, updated := validateObjectiveGraphNode(node, graph, indices)
		nodeValidations = append(nodeValidations, validation)
		validatedNodes = append(validatedNodes, updated)
		if validation.Status == VerificationSatisfied {
			if validation.State == ObjectiveNodeStateReady {
				result.ReadyNodeCount++
			}
			continue
		}
		if validation.Status == VerificationNotApplicable && validation.State == ObjectiveNodeStateSkipped {
			result.SkippedNodeCount++
			continue
		}
		result.MissingInputs = MergeMissingInputs(result.MissingInputs, validation.MissingInputs)
		result.BlockedReasons = normalizeControlTokenList(append(result.BlockedReasons, validation.BlockedReasons...))
		result.Boundaries = MergeBoundaries(result.Boundaries, validation.Boundaries)
		result.FailureClass = firstFailureClass(result.FailureClass, validation.FailureClass)
	}
	result.Graph.Nodes = validatedNodes
	result.NodeValidations = nodeValidations
	if len(result.MissingInputs) > 0 || result.FailureClass != FailureNone {
		result.Status = VerificationBlocked
		result.FailureClass = firstFailureClass(result.FailureClass, FailureInvalidInput)
		result.NextHostAction = objectiveGraphNextAction(result.MissingInputs)
		return result.Normalize()
	}
	result.Status = VerificationSatisfied
	result.Validated = true
	result.ReadyForRuntimeLoop = true
	result.FailureClass = FailureNone
	result.NextHostAction = "run_bounded_objective_runtime_loop"
	result.Boundaries = AppendBoundaries(result.Boundaries, "objective_graph_validated")
	return result.Normalize()
}

func (r ObjectiveGraphValidationReport) Clone() ObjectiveGraphValidationReport {
	out := r
	out.Graph = r.Graph.Clone()
	out.NodeValidations = cloneObjectiveGraphNodeValidations(r.NodeValidations)
	out.MissingInputs = cloneMissingInputs(r.MissingInputs)
	out.BlockedReasons = cloneStringSlice(r.BlockedReasons)
	out.Boundaries = cloneBoundaries(r.Boundaries)
	return out
}

func (r ObjectiveGraphValidationReport) Normalize() ObjectiveGraphValidationReport {
	out := r.Clone()
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Status = NormalizeVerificationStatus(string(out.Status))
	if out.Status == VerificationNotEvaluated {
		out.Status = VerificationBlocked
	}
	out.Mode = normalizeControlToken(out.Mode)
	if out.Mode == "" {
		out.Mode = "objective_graph_validation"
	}
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	if out.RunnerEffect == "" {
		out.RunnerEffect = "none"
	}
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	if out.PromptEffect == "" {
		out.PromptEffect = "none"
	}
	out.SourceRef = normalizeOneDisplaySafeRef(out.SourceRef)
	out.Graph = out.Graph.Normalize()
	out.NodeValidations = normalizeObjectiveGraphNodeValidations(out.NodeValidations)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.BlockedReasons = normalizeControlTokenList(out.BlockedReasons)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	if out.RawOutputLoaded || out.Graph.RawOutputLoaded {
		out.Status = VerificationReviewRequired
		out.Validated = false
		out.ReadyForRuntimeLoop = false
		out.FailureClass = firstFailureClass(out.FailureClass, FailureEvidenceWeak)
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "objective_graph_unsafe_input")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		out.NextHostAction = "provide_display_safe_refs"
	}
	if out.Status != VerificationSatisfied {
		out.Validated = false
		out.ReadyForRuntimeLoop = false
	}
	return out
}

type ObjectiveGraphJSONDecodeInput struct {
	RawJSON               []byte                          `json:"-"`
	GraphRef              DisplaySafeRef                  `json:"graph_ref,omitempty"`
	Spec                  ObjectiveSpec                   `json:"spec,omitempty"`
	Catalog               StrategyCatalogSnapshot         `json:"catalog,omitempty"`
	CapabilityDescriptors []ObjectiveCapabilityDescriptor `json:"capability_descriptors,omitempty"`
	Policy                ExecutionIntensityPolicy        `json:"policy,omitempty"`
	AllowedCapabilityRefs []DisplaySafeRef                `json:"allowed_capability_refs,omitempty"`
	SourceRef             DisplaySafeRef                  `json:"source_ref,omitempty"`
	Boundaries            []Boundary                      `json:"boundaries,omitempty"`
	RawOutputLoaded       bool                            `json:"raw_output_loaded"`
}

type ObjectiveGraphJSONDecodeReport struct {
	ContractVersion     string                         `json:"contract_version,omitempty"`
	Decoded             bool                           `json:"decoded"`
	Available           bool                           `json:"available"`
	Status              VerificationStatus             `json:"status,omitempty"`
	Mode                string                         `json:"mode,omitempty"`
	RunnerEffect        string                         `json:"runner_effect,omitempty"`
	PromptEffect        string                         `json:"prompt_effect,omitempty"`
	GraphRef            DisplaySafeRef                 `json:"graph_ref,omitempty"`
	SourceRef           DisplaySafeRef                 `json:"source_ref,omitempty"`
	Graph               ObjectiveGraph                 `json:"graph,omitempty"`
	Validation          ObjectiveGraphValidationReport `json:"validation,omitempty"`
	ReadyForRuntimeLoop bool                           `json:"ready_for_runtime_loop"`
	FailureClass        FailureClass                   `json:"failure_class,omitempty"`
	MissingInputs       []MissingInput                 `json:"missing_inputs,omitempty"`
	Boundaries          []Boundary                     `json:"boundaries,omitempty"`
	NextHostAction      NextHostAction                 `json:"next_host_action,omitempty"`
	RawOutputLoaded     bool                           `json:"raw_output_loaded"`
}

func BuildObjectiveGraphFromJSON(input ObjectiveGraphJSONDecodeInput) ObjectiveGraphJSONDecodeReport {
	result := baseObjectiveGraphJSONDecodeReport(input)
	if objectiveGraphJSONDecodeInputUnsafe(input) {
		return objectiveGraphJSONDecodeReportBlock(result, VerificationReviewRequired, FailureEvidenceWeak, "host:display_safe_refs", "provide_display_safe_refs", "raw_output_not_allowed")
	}
	if len(bytes.TrimSpace(input.RawJSON)) == 0 {
		return objectiveGraphJSONDecodeReportBlock(result, VerificationBlocked, FailureInvalidInput, "host:objective_graph_json", "provide_objective_graph_json", "objective_graph_json_missing")
	}
	graph, ok, boundary := decodeObjectiveGraphJSON(input.RawJSON)
	if !ok {
		result = objectiveGraphJSONDecodeReportBlock(result, VerificationBlocked, FailureInvalidInput, "host:objective_graph_json", "provide_objective_graph_json", boundary)
		result.Boundaries = AppendBoundaries(result.Boundaries, "deterministic_blocked_fallback", "no_prompt_heuristic_fallback")
		return result.Normalize()
	}
	result.Decoded = true
	graph = objectiveGraphApplyDecodeDefaults(graph, input).Normalize()
	if !objectiveGraphCapabilitiesAllowed(graph, input.AllowedCapabilityRefs) {
		result.Graph = graph
		return objectiveGraphJSONDecodeReportBlock(result, VerificationBlocked, FailureCapabilityMissing, "host:allowed_capability_ref", "revise_objective_graph", "objective_graph_capability_not_allowed")
	}
	validation := BuildObjectiveGraphValidation(ObjectiveGraphValidationInput{
		Graph:                 graph,
		Spec:                  input.Spec,
		Catalog:               input.Catalog,
		CapabilityDescriptors: input.CapabilityDescriptors,
		Policy:                input.Policy,
		SourceRef:             input.SourceRef,
		Boundaries: MergeBoundaries(
			input.Boundaries,
			[]Boundary{"objective_graph_json_validated"},
		),
		RawOutputLoaded: input.RawOutputLoaded,
	})
	result.Graph = validation.Graph
	result.Validation = validation
	result.Status = validation.Status
	result.FailureClass = validation.FailureClass
	result.MissingInputs = MergeMissingInputs(result.MissingInputs, validation.MissingInputs)
	result.Boundaries = MergeBoundaries(result.Boundaries, validation.Boundaries)
	result.NextHostAction = validation.NextHostAction
	result.ReadyForRuntimeLoop = validation.ReadyForRuntimeLoop
	if result.ReadyForRuntimeLoop {
		result.FailureClass = FailureNone
	}
	return result.Normalize()
}

func (r ObjectiveGraphJSONDecodeReport) Clone() ObjectiveGraphJSONDecodeReport {
	out := r
	out.Graph = r.Graph.Clone()
	out.Validation = r.Validation.Clone()
	out.MissingInputs = cloneMissingInputs(r.MissingInputs)
	out.Boundaries = cloneBoundaries(r.Boundaries)
	return out
}

func (r ObjectiveGraphJSONDecodeReport) Normalize() ObjectiveGraphJSONDecodeReport {
	out := r.Clone()
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Status = NormalizeVerificationStatus(string(out.Status))
	if out.Status == VerificationNotEvaluated {
		out.Status = VerificationBlocked
	}
	out.Mode = normalizeControlToken(out.Mode)
	if out.Mode == "" {
		out.Mode = "objective_graph_json_decode"
	}
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	if out.RunnerEffect == "" {
		out.RunnerEffect = "none"
	}
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	if out.PromptEffect == "" {
		out.PromptEffect = "none"
	}
	out.GraphRef = normalizeOneDisplaySafeRef(out.GraphRef)
	out.SourceRef = normalizeOneDisplaySafeRef(out.SourceRef)
	out.Graph = out.Graph.Normalize()
	out.Validation = out.Validation.Normalize()
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	if out.RawOutputLoaded || out.Graph.RawOutputLoaded || out.Validation.RawOutputLoaded {
		out.Status = VerificationReviewRequired
		out.ReadyForRuntimeLoop = false
		out.FailureClass = firstFailureClass(out.FailureClass, FailureEvidenceWeak)
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		out.NextHostAction = "provide_display_safe_refs"
	}
	if out.Status != VerificationSatisfied {
		out.ReadyForRuntimeLoop = false
	}
	return out
}

type ObjectiveGraphPlanner interface {
	BuildObjectiveGraph(context.Context, ObjectiveGraphPlannerRequest) (ObjectiveGraphPlannerResponse, error)
}

type ObjectiveGraphPlannerFunc func(context.Context, ObjectiveGraphPlannerRequest) (ObjectiveGraphPlannerResponse, error)

func (f ObjectiveGraphPlannerFunc) BuildObjectiveGraph(ctx context.Context, request ObjectiveGraphPlannerRequest) (ObjectiveGraphPlannerResponse, error) {
	return f(ctx, request)
}

type ObjectiveGraphPlannerRequest struct {
	RequestRef            DisplaySafeRef                  `json:"request_ref,omitempty"`
	Spec                  ObjectiveSpec                   `json:"spec,omitempty"`
	Catalog               StrategyCatalogSnapshot         `json:"catalog,omitempty"`
	CapabilityDescriptors []ObjectiveCapabilityDescriptor `json:"capability_descriptors,omitempty"`
	Policy                ExecutionIntensityPolicy        `json:"policy,omitempty"`
	AllowedCapabilityRefs []DisplaySafeRef                `json:"allowed_capability_refs,omitempty"`
	PolicyRefs            []DisplaySafeRef                `json:"policy_refs,omitempty"`
	Boundaries            []Boundary                      `json:"boundaries,omitempty"`
	RawOutputLoaded       bool                            `json:"raw_output_loaded"`
}

func CloneObjectiveGraphPlannerRequest(in ObjectiveGraphPlannerRequest) ObjectiveGraphPlannerRequest {
	out := in
	out.Spec = in.Spec.Clone()
	out.Catalog = in.Catalog.Clone()
	out.CapabilityDescriptors = cloneObjectiveCapabilityDescriptors(in.CapabilityDescriptors)
	out.Policy = in.Policy.Clone()
	out.AllowedCapabilityRefs = cloneDisplaySafeRefs(in.AllowedCapabilityRefs)
	out.PolicyRefs = cloneDisplaySafeRefs(in.PolicyRefs)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (r ObjectiveGraphPlannerRequest) Clone() ObjectiveGraphPlannerRequest {
	return CloneObjectiveGraphPlannerRequest(r)
}

func (r ObjectiveGraphPlannerRequest) Normalize() ObjectiveGraphPlannerRequest {
	out := CloneObjectiveGraphPlannerRequest(r)
	out.RequestRef = normalizeOneDisplaySafeRef(out.RequestRef)
	out.Spec = out.Spec.Normalize()
	out.Catalog = out.Catalog.Normalize()
	out.CapabilityDescriptors = normalizeObjectiveCapabilityDescriptors(out.CapabilityDescriptors)
	out.Policy = out.Policy.Normalize()
	out.AllowedCapabilityRefs = normalizeDisplaySafeRefs(out.AllowedCapabilityRefs)
	out.PolicyRefs = normalizeDisplaySafeRefs(out.PolicyRefs)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	return out
}

type ObjectiveGraphPlannerResponse struct {
	ResponseRef     DisplaySafeRef `json:"response_ref,omitempty"`
	Graph           ObjectiveGraph `json:"graph,omitempty"`
	GraphJSON       []byte         `json:"-"`
	Boundaries      []Boundary     `json:"boundaries,omitempty"`
	RawOutputLoaded bool           `json:"raw_output_loaded"`
}

func CloneObjectiveGraphPlannerResponse(in ObjectiveGraphPlannerResponse) ObjectiveGraphPlannerResponse {
	out := in
	out.Graph = in.Graph.Clone()
	out.GraphJSON = append([]byte(nil), in.GraphJSON...)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (r ObjectiveGraphPlannerResponse) Clone() ObjectiveGraphPlannerResponse {
	return CloneObjectiveGraphPlannerResponse(r)
}

func (r ObjectiveGraphPlannerResponse) Normalize() ObjectiveGraphPlannerResponse {
	out := CloneObjectiveGraphPlannerResponse(r)
	out.ResponseRef = normalizeOneDisplaySafeRef(out.ResponseRef)
	out.Graph = out.Graph.Normalize()
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	return out
}

type ObjectiveGraphBuildInput struct {
	Enabled    bool                         `json:"enabled"`
	Planner    ObjectiveGraphPlanner        `json:"-"`
	Request    ObjectiveGraphPlannerRequest `json:"request,omitempty"`
	GraphRef   DisplaySafeRef               `json:"graph_ref,omitempty"`
	SourceRef  DisplaySafeRef               `json:"source_ref,omitempty"`
	Boundaries []Boundary                   `json:"boundaries,omitempty"`
}

type ObjectiveGraphBuildReport struct {
	ContractVersion     string                         `json:"contract_version,omitempty"`
	Built               bool                           `json:"built"`
	Available           bool                           `json:"available"`
	Status              VerificationStatus             `json:"status,omitempty"`
	Mode                string                         `json:"mode,omitempty"`
	RunnerEffect        string                         `json:"runner_effect,omitempty"`
	PromptEffect        string                         `json:"prompt_effect,omitempty"`
	RequestRef          DisplaySafeRef                 `json:"request_ref,omitempty"`
	SourceRef           DisplaySafeRef                 `json:"source_ref,omitempty"`
	PlannerCalled       bool                           `json:"planner_called"`
	DecodeAttempted     bool                           `json:"decode_attempted"`
	Graph               ObjectiveGraph                 `json:"graph,omitempty"`
	JSONDecode          ObjectiveGraphJSONDecodeReport `json:"json_decode,omitempty"`
	Validation          ObjectiveGraphValidationReport `json:"validation,omitempty"`
	ReadyForRuntimeLoop bool                           `json:"ready_for_runtime_loop"`
	FailureClass        FailureClass                   `json:"failure_class,omitempty"`
	MissingInputs       []MissingInput                 `json:"missing_inputs,omitempty"`
	Boundaries          []Boundary                     `json:"boundaries,omitempty"`
	NextHostAction      NextHostAction                 `json:"next_host_action,omitempty"`
	RawOutputLoaded     bool                           `json:"raw_output_loaded"`
}

func BuildObjectiveGraphWithPlanner(ctx context.Context, input ObjectiveGraphBuildInput) ObjectiveGraphBuildReport {
	result := baseObjectiveGraphBuildReport(input)
	if objectiveGraphBuildInputUnsafe(input) {
		return objectiveGraphBuildReportBlock(result, VerificationReviewRequired, FailureEvidenceWeak, "host:display_safe_refs", "provide_display_safe_refs", "raw_output_not_allowed")
	}
	if !input.Enabled {
		return objectiveGraphBuildReportBlock(result, VerificationBlocked, FailureInsufficientInformation, "host:objective_graph_planner_enabled", "enable_objective_closure", "objective_graph_planner_disabled")
	}
	if input.Planner == nil {
		return objectiveGraphBuildReportBlock(result, VerificationBlocked, FailureHostAdapterMissing, "host:objective_graph_planner", "provide_objective_graph_planner", "objective_graph_planner_missing")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	request := input.Request.Normalize()
	result.PlannerCalled = true
	response, err := input.Planner.BuildObjectiveGraph(ctx, request)
	if err != nil {
		result = objectiveGraphBuildReportBlock(result, VerificationBlocked, FailureExternalDependencyUnavailable, "host:objective_graph_planner_result", "provide_objective_graph", "objective_graph_planner_failed")
		result.Boundaries = AppendBoundaries(result.Boundaries, "deterministic_blocked_fallback", "no_prompt_heuristic_fallback")
		return result.Normalize()
	}
	response = response.Normalize()
	if response.RawOutputLoaded || displaySafeRefRejected(response.ResponseRef) {
		return objectiveGraphBuildReportBlock(result, VerificationReviewRequired, FailureEvidenceWeak, "host:display_safe_refs", "provide_display_safe_refs", "raw_output_not_allowed")
	}
	if len(bytes.TrimSpace(response.GraphJSON)) > 0 {
		result.DecodeAttempted = true
		decode := BuildObjectiveGraphFromJSON(ObjectiveGraphJSONDecodeInput{
			RawJSON:               response.GraphJSON,
			GraphRef:              firstDisplaySafeRef(input.GraphRef, response.ResponseRef),
			Spec:                  request.Spec,
			Catalog:               request.Catalog,
			CapabilityDescriptors: request.CapabilityDescriptors,
			Policy:                request.Policy,
			AllowedCapabilityRefs: request.AllowedCapabilityRefs,
			SourceRef:             firstDisplaySafeRef(input.SourceRef, response.ResponseRef),
			Boundaries: MergeBoundaries(
				input.Boundaries,
				request.Boundaries,
				response.Boundaries,
				[]Boundary{"objective_graph_planner_json_response"},
			),
			RawOutputLoaded: response.RawOutputLoaded,
		})
		return objectiveGraphBuildReportFromDecode(result, decode)
	}
	if objectiveGraphIsEmpty(response.Graph) {
		return objectiveGraphBuildReportBlock(result, VerificationBlocked, FailureInvalidInput, "host:objective_graph", "provide_objective_graph", "objective_graph_planner_empty_response")
	}
	graph := objectiveGraphApplyPlannerDefaults(response.Graph, request, response, input).Normalize()
	if !objectiveGraphCapabilitiesAllowed(graph, request.AllowedCapabilityRefs) {
		result.Graph = graph
		return objectiveGraphBuildReportBlock(result, VerificationBlocked, FailureCapabilityMissing, "host:allowed_capability_ref", "revise_objective_graph", "objective_graph_capability_not_allowed")
	}
	validation := BuildObjectiveGraphValidation(ObjectiveGraphValidationInput{
		Graph:                 graph,
		Spec:                  request.Spec,
		Catalog:               request.Catalog,
		CapabilityDescriptors: request.CapabilityDescriptors,
		Policy:                request.Policy,
		SourceRef:             firstDisplaySafeRef(input.SourceRef, response.ResponseRef),
		Boundaries: MergeBoundaries(
			input.Boundaries,
			request.Boundaries,
			response.Boundaries,
			[]Boundary{"objective_graph_planner_direct_response"},
		),
	})
	result.Graph = validation.Graph
	result.Validation = validation
	result.Status = validation.Status
	result.FailureClass = validation.FailureClass
	result.MissingInputs = MergeMissingInputs(result.MissingInputs, validation.MissingInputs)
	result.Boundaries = MergeBoundaries(result.Boundaries, validation.Boundaries)
	result.NextHostAction = validation.NextHostAction
	result.ReadyForRuntimeLoop = validation.ReadyForRuntimeLoop
	result.Built = validation.ReadyForRuntimeLoop
	if result.ReadyForRuntimeLoop {
		result.FailureClass = FailureNone
	}
	return result.Normalize()
}

func (r ObjectiveGraphBuildReport) Clone() ObjectiveGraphBuildReport {
	out := r
	out.Graph = r.Graph.Clone()
	out.JSONDecode = r.JSONDecode.Clone()
	out.Validation = r.Validation.Clone()
	out.MissingInputs = cloneMissingInputs(r.MissingInputs)
	out.Boundaries = cloneBoundaries(r.Boundaries)
	return out
}

func (r ObjectiveGraphBuildReport) Normalize() ObjectiveGraphBuildReport {
	out := r.Clone()
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Status = NormalizeVerificationStatus(string(out.Status))
	if out.Status == VerificationNotEvaluated {
		out.Status = VerificationBlocked
	}
	out.Mode = normalizeControlToken(out.Mode)
	if out.Mode == "" {
		out.Mode = "objective_graph_planner"
	}
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	if out.RunnerEffect == "" {
		out.RunnerEffect = "none"
	}
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	if out.PromptEffect == "" {
		out.PromptEffect = "host_planner_interface_only"
	}
	out.RequestRef = normalizeOneDisplaySafeRef(out.RequestRef)
	out.SourceRef = normalizeOneDisplaySafeRef(out.SourceRef)
	out.Graph = out.Graph.Normalize()
	out.JSONDecode = out.JSONDecode.Normalize()
	out.Validation = out.Validation.Normalize()
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	if out.RawOutputLoaded || out.Graph.RawOutputLoaded || out.JSONDecode.RawOutputLoaded || out.Validation.RawOutputLoaded {
		out.Status = VerificationReviewRequired
		out.Built = false
		out.ReadyForRuntimeLoop = false
		out.FailureClass = firstFailureClass(out.FailureClass, FailureEvidenceWeak)
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		out.NextHostAction = "provide_display_safe_refs"
	}
	if out.Status != VerificationSatisfied {
		out.Built = false
		out.ReadyForRuntimeLoop = false
	}
	return out
}

func baseObjectiveGraphValidationReport(input ObjectiveGraphValidationInput) ObjectiveGraphValidationReport {
	graph := input.Graph.Normalize()
	return ObjectiveGraphValidationReport{
		ContractVersion: ContractVersion,
		Available:       true,
		Status:          VerificationBlocked,
		Mode:            "objective_graph_validation",
		RunnerEffect:    "none",
		PromptEffect:    "none",
		SourceRef:       normalizeOneDisplaySafeRef(input.SourceRef),
		Graph:           graph,
		FailureClass:    FailureInsufficientInformation,
		Boundaries: MergeBoundaries(
			[]Boundary{
				"objective_graph_validator",
				"deterministic_graph_validator",
				"catalog_ref_validation",
				"capability_ref_validation",
				"acyclic_graph_validation",
				"budget_validation",
				"side_effect_gate",
				"no_llm_call",
				"no_runner_dispatch",
				"no_backend_execution",
			},
			input.Boundaries,
			graph.Boundaries,
		),
		NextHostAction:  "provide_objective_graph",
		RawOutputLoaded: input.RawOutputLoaded || graph.RawOutputLoaded,
	}
}

func baseObjectiveGraphJSONDecodeReport(input ObjectiveGraphJSONDecodeInput) ObjectiveGraphJSONDecodeReport {
	return ObjectiveGraphJSONDecodeReport{
		ContractVersion: ContractVersion,
		Available:       true,
		Status:          VerificationBlocked,
		Mode:            "objective_graph_json_decode",
		RunnerEffect:    "none",
		PromptEffect:    "none",
		GraphRef:        normalizeOneDisplaySafeRef(input.GraphRef),
		SourceRef:       normalizeOneDisplaySafeRef(input.SourceRef),
		FailureClass:    FailureInsufficientInformation,
		Boundaries: MergeBoundaries(
			[]Boundary{
				"objective_graph_json_decoder",
				"strict_json_decoder",
				"disallow_unknown_fields",
				"no_code_fence_stripping",
				"no_llm_call",
				"no_runner_dispatch",
				"no_backend_execution",
			},
			input.Boundaries,
		),
		NextHostAction:  "provide_objective_graph_json",
		RawOutputLoaded: input.RawOutputLoaded,
	}
}

func baseObjectiveGraphBuildReport(input ObjectiveGraphBuildInput) ObjectiveGraphBuildReport {
	request := input.Request.Normalize()
	return ObjectiveGraphBuildReport{
		ContractVersion: ContractVersion,
		Available:       true,
		Status:          VerificationBlocked,
		Mode:            "objective_graph_planner",
		RunnerEffect:    "none",
		PromptEffect:    "host_planner_interface_only",
		RequestRef:      request.RequestRef,
		SourceRef:       normalizeOneDisplaySafeRef(input.SourceRef),
		FailureClass:    FailureInsufficientInformation,
		Boundaries: MergeBoundaries(
			[]Boundary{
				"objective_graph_planner_interface",
				"host_owned_llm_planner",
				"core_validator_only",
				"no_concrete_model_binding",
				"no_prompt_heuristic_fallback",
				"no_runner_dispatch",
				"no_backend_execution",
			},
			input.Boundaries,
			request.Boundaries,
		),
		NextHostAction:  "provide_objective_graph",
		RawOutputLoaded: request.RawOutputLoaded,
	}
}

func objectiveGraphValidationBlock(result ObjectiveGraphValidationReport, status VerificationStatus, failure FailureClass, missing MissingInput, next NextHostAction, boundary Boundary) ObjectiveGraphValidationReport {
	result.Status = status
	result.FailureClass = failure
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, string(boundary))
	result.Boundaries = AppendBoundaries(result.Boundaries, boundary)
	result.NextHostAction = next
	return result.Normalize()
}

func objectiveGraphJSONDecodeReportBlock(result ObjectiveGraphJSONDecodeReport, status VerificationStatus, failure FailureClass, missing MissingInput, next NextHostAction, boundary Boundary) ObjectiveGraphJSONDecodeReport {
	result.Status = status
	result.FailureClass = failure
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.Boundaries = AppendBoundaries(result.Boundaries, boundary)
	result.NextHostAction = next
	return result.Normalize()
}

func objectiveGraphBuildReportBlock(result ObjectiveGraphBuildReport, status VerificationStatus, failure FailureClass, missing MissingInput, next NextHostAction, boundary Boundary) ObjectiveGraphBuildReport {
	result.Status = status
	result.FailureClass = failure
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.Boundaries = AppendBoundaries(result.Boundaries, boundary)
	result.NextHostAction = next
	return result.Normalize()
}

func objectiveGraphBuildReportFromDecode(result ObjectiveGraphBuildReport, decode ObjectiveGraphJSONDecodeReport) ObjectiveGraphBuildReport {
	result.JSONDecode = decode
	result.Graph = decode.Graph
	result.Validation = decode.Validation
	result.Status = decode.Status
	result.FailureClass = decode.FailureClass
	result.MissingInputs = MergeMissingInputs(result.MissingInputs, decode.MissingInputs)
	result.Boundaries = MergeBoundaries(result.Boundaries, decode.Boundaries)
	result.NextHostAction = decode.NextHostAction
	result.ReadyForRuntimeLoop = decode.ReadyForRuntimeLoop
	result.Built = decode.ReadyForRuntimeLoop
	result.RawOutputLoaded = result.RawOutputLoaded || decode.RawOutputLoaded
	if result.ReadyForRuntimeLoop {
		result.FailureClass = FailureNone
	}
	return result.Normalize()
}

func decodeObjectiveGraphJSON(raw []byte) (ObjectiveGraph, bool, Boundary) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var graph ObjectiveGraph
	if err := decoder.Decode(&graph); err != nil {
		return ObjectiveGraph{}, false, "objective_graph_json_decode_failed"
	}
	var trailing struct{}
	if err := decoder.Decode(&trailing); err != io.EOF {
		return ObjectiveGraph{}, false, "objective_graph_json_trailing_tokens"
	}
	return graph, true, ""
}

type objectiveGraphValidationIndices struct {
	Catalog     StrategyCatalogSnapshot
	Descriptors []ObjectiveCapabilityDescriptor
	Policy      ExecutionIntensityPolicy
	Spec        ObjectiveSpec
}

func validateObjectiveGraphNode(node ObjectiveNode, graph ObjectiveGraph, indices objectiveGraphValidationIndices) (ObjectiveGraphNodeValidation, ObjectiveNode) {
	normalized := node.Normalize()
	state := objectiveGraphInitialNodeState(normalized)
	result := ObjectiveGraphNodeValidation{
		ContractVersion: ContractVersion,
		NodeRef:         normalized.NodeRef,
		Status:          VerificationSatisfied,
		State:           state,
		FailureClass:    FailureNone,
		Boundaries: MergeBoundaries(
			[]Boundary{"objective_graph_node_validated"},
			normalized.Boundaries,
		),
		NextHostAction: "run_bounded_objective_runtime_loop",
	}
	result.MissingInputs = MergeMissingInputs(result.MissingInputs, objectiveGraphNodeMissingInputs(normalized))
	if len(result.MissingInputs) > 0 {
		return objectiveGraphNodeBlock(result, normalized, FailureInvalidInput, "objective_graph_node_incomplete", "revise_objective_graph")
	}
	if !objectiveGraphCapabilityKnown(normalized.CapabilityRef, indices.Catalog, indices.Descriptors) {
		result.MissingInputs = AppendMissingInputs(result.MissingInputs, MissingInput("host:available_"+string(normalized.CapabilityRef)))
		if normalized.Optional {
			result.Status = VerificationNotApplicable
			result.State = ObjectiveNodeStateSkipped
			result.FailureClass = FailureNone
			result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, "optional_node_capability_unavailable")
			result.Boundaries = AppendBoundaries(result.Boundaries, "optional_node_skipped")
			result.NextHostAction = "continue_with_available_objective_graph"
			normalized.State = ObjectiveNodeStateSkipped
			return result.Normalize(), normalized.Normalize()
		}
		return objectiveGraphNodeBlock(result, normalized, FailureCapabilityMissing, "objective_graph_node_capability_missing", "resolve_objective_graph_capability")
	}
	if normalized.StrategyRef != "" && !objectiveGraphStrategyCapabilityMatches(normalized.StrategyRef, normalized.CapabilityRef, indices.Catalog) {
		result.MissingInputs = AppendMissingInputs(result.MissingInputs, "host:objective_graph_strategy_ref")
		return objectiveGraphNodeBlock(result, normalized, FailureCapabilityMissing, "objective_graph_node_strategy_missing", "revise_objective_graph")
	}
	if normalized.DescriptorRef != "" && len(indices.Descriptors) > 0 && !objectiveGraphDescriptorCapabilityMatches(normalized.DescriptorRef, normalized.CapabilityRef, indices.Descriptors) {
		result.MissingInputs = AppendMissingInputs(result.MissingInputs, "host:objective_graph_descriptor_ref")
		return objectiveGraphNodeBlock(result, normalized, FailureCapabilityMissing, "objective_graph_node_descriptor_missing", "revise_objective_graph")
	}
	if !objectiveGraphDependenciesExist(normalized, graph) {
		result.MissingInputs = AppendMissingInputs(result.MissingInputs, "host:objective_graph_dependency")
		return objectiveGraphNodeBlock(result, normalized, FailureInvalidInput, "objective_graph_node_dependency_missing", "revise_objective_graph")
	}
	if !objectiveGraphNodeSideEffectAllowed(normalized, indices.Spec, indices.Policy) {
		result.MissingInputs = AppendMissingInputs(result.MissingInputs, "contract:objective_graph_side_effect")
		return objectiveGraphNodeBlock(result, normalized, FailurePolicyBlocked, "objective_graph_node_side_effect_denied", "request_host_approval")
	}
	if normalized.RequiresApproval && len(normalized.ApprovalRefs) == 0 && len(indices.Spec.ApprovalRefs) == 0 {
		result.MissingInputs = AppendMissingInputs(result.MissingInputs, "host:objective_graph_node_approval_ref")
		return objectiveGraphNodeBlock(result, normalized, FailureApprovalRequired, "objective_graph_node_approval_missing", "request_host_approval")
	}
	normalized.State = state
	return result.Normalize(), normalized.Normalize()
}

func objectiveGraphNodeBlock(result ObjectiveGraphNodeValidation, node ObjectiveNode, failure FailureClass, reason string, next NextHostAction) (ObjectiveGraphNodeValidation, ObjectiveNode) {
	result.Status = VerificationBlocked
	result.State = ObjectiveNodeStateBlocked
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.Boundaries = AppendBoundaries(result.Boundaries, Boundary(reason))
	result.NextHostAction = next
	node.State = ObjectiveNodeStateBlocked
	node.MissingInputs = MergeMissingInputs(node.MissingInputs, result.MissingInputs)
	node.Boundaries = MergeBoundaries(node.Boundaries, result.Boundaries)
	return result.Normalize(), node.Normalize()
}

func objectiveGraphInitialNodeState(node ObjectiveNode) ObjectiveNodeState {
	if node.State == ObjectiveNodeStateSkipped ||
		node.State == ObjectiveNodeStateSatisfied ||
		node.State == ObjectiveNodeStatePartial ||
		node.State == ObjectiveNodeStateBlocked {
		return node.State
	}
	if len(node.Dependencies) == 0 {
		return ObjectiveNodeStateReady
	}
	return ObjectiveNodeStatePending
}

func objectiveGraphApplyDefaults(graph ObjectiveGraph, spec ObjectiveSpec, catalog StrategyCatalogSnapshot) ObjectiveGraph {
	out := graph.Clone()
	if out.SpecRef == "" {
		out.SpecRef = spec.SpecRef
	}
	if out.ObjectiveID == "" {
		out.ObjectiveID = spec.ObjectiveID
	}
	if out.CatalogRef == "" {
		out.CatalogRef = catalog.CatalogRef
	}
	if len(out.RequiredEvidence) == 0 {
		out.RequiredEvidence = cloneEvidenceRefs(spec.RequiredEvidence)
	}
	if out.Budget.BudgetRef == "" && out.Budget.MaxNodes == 0 && out.Budget.MaxAttempts == 0 && out.Budget.MaxDurationSeconds == 0 && out.Budget.MaxCostUnits == 0 {
		out.Budget = spec.Budget.Clone()
	}
	out.PolicyRefs = normalizeDisplaySafeRefs(append(out.PolicyRefs, spec.PolicyRefs...))
	return out
}

func objectiveGraphApplyDecodeDefaults(graph ObjectiveGraph, input ObjectiveGraphJSONDecodeInput) ObjectiveGraph {
	out := objectiveGraphApplyDefaults(graph, input.Spec, input.Catalog)
	if out.GraphRef == "" {
		out.GraphRef = input.GraphRef
	}
	out.Boundaries = MergeBoundaries(out.Boundaries, input.Boundaries)
	return out
}

func objectiveGraphApplyPlannerDefaults(graph ObjectiveGraph, request ObjectiveGraphPlannerRequest, response ObjectiveGraphPlannerResponse, input ObjectiveGraphBuildInput) ObjectiveGraph {
	out := objectiveGraphApplyDefaults(graph, request.Spec, request.Catalog)
	if out.GraphRef == "" {
		out.GraphRef = firstDisplaySafeRef(input.GraphRef, response.ResponseRef)
	}
	out.PolicyRefs = normalizeDisplaySafeRefs(append(out.PolicyRefs, request.PolicyRefs...))
	out.Boundaries = MergeBoundaries(out.Boundaries, request.Boundaries, response.Boundaries, input.Boundaries)
	return out
}

func objectiveGraphMissingInputs(graph ObjectiveGraph, catalog StrategyCatalogSnapshot) []MissingInput {
	var missing []MissingInput
	if graph.GraphRef == "" {
		missing = AppendMissingInputs(missing, "host:objective_graph_ref")
	}
	if graph.SpecRef == "" && graph.ObjectiveID == "" {
		missing = AppendMissingInputs(missing, "host:objective_spec_ref")
	}
	if graph.CatalogRef == "" && catalog.CatalogRef == "" {
		missing = AppendMissingInputs(missing, "host:strategy_catalog_ref")
	}
	if len(graph.Nodes) == 0 {
		missing = AppendMissingInputs(missing, "host:objective_graph_nodes")
	}
	missing = AppendMissingInputs(missing, graph.MissingInputs...)
	return missing
}

func objectiveGraphNodeMissingInputs(node ObjectiveNode) []MissingInput {
	var missing []MissingInput
	if node.NodeRef == "" {
		missing = AppendMissingInputs(missing, "host:objective_graph_node_ref")
	}
	if node.CapabilityRef == "" {
		missing = AppendMissingInputs(missing, "host:objective_graph_node_capability_ref")
	}
	if node.InputSchemaRef == "" {
		missing = AppendMissingInputs(missing, "host:objective_graph_node_input_schema_ref")
	}
	if node.OutputSchemaRef == "" {
		missing = AppendMissingInputs(missing, "host:objective_graph_node_output_schema_ref")
	}
	if node.EvidenceContractRef == "" {
		missing = AppendMissingInputs(missing, "host:objective_graph_node_evidence_contract_ref")
	}
	if len(node.RequiredEvidence) == 0 {
		missing = AppendMissingInputs(missing, "host:objective_graph_node_required_evidence")
	}
	if node.SideEffectClass == ObjectiveCapabilitySideEffectUnspecified {
		missing = AppendMissingInputs(missing, "host:objective_graph_node_side_effect_class")
	}
	missing = AppendMissingInputs(missing, node.MissingInputs...)
	return missing
}

func objectiveGraphFailure(missing []MissingInput) FailureClass {
	for _, value := range missing {
		switch value {
		case "host:objective_graph_nodes",
			"host:objective_graph_node_ref",
			"host:objective_graph_node_input_schema_ref",
			"host:objective_graph_node_output_schema_ref",
			"host:objective_graph_node_evidence_contract_ref",
			"host:objective_graph_node_required_evidence":
			return FailureInvalidInput
		case "host:strategy_catalog_ref":
			return FailureConfigMissing
		}
	}
	return FailureInsufficientInformation
}

func objectiveGraphNextAction(missing []MissingInput) NextHostAction {
	for _, value := range missing {
		switch value {
		case "host:strategy_catalog_ref":
			return "provide_strategy_catalog"
		case "host:objective_graph_nodes",
			"host:objective_graph_node_ref",
			"host:objective_graph_node_input_schema_ref",
			"host:objective_graph_node_output_schema_ref",
			"host:objective_graph_node_evidence_contract_ref",
			"host:objective_graph_node_required_evidence":
			return "revise_objective_graph"
		case "host:objective_graph_acyclic":
			return "revise_objective_graph"
		case "contract:objective_graph_side_effect", "host:objective_graph_node_approval_ref":
			return "request_host_approval"
		}
	}
	return "revise_objective_graph"
}

func objectiveGraphAttemptBudget(graph ObjectiveGraph) int {
	total := 0
	for _, node := range graph.Nodes {
		if node.Optional {
			continue
		}
		total += node.AttemptPolicy.Normalize().MaxAttempts
	}
	return total
}

func objectiveGraphMissingSpecEvidence(spec ObjectiveSpec, graph ObjectiveGraph) []MissingInput {
	required := MergeEvidenceRefs(spec.RequiredEvidence, graph.RequiredEvidence)
	if len(required) == 0 {
		return nil
	}
	var available []EvidenceRef
	for _, node := range graph.Nodes {
		available = MergeEvidenceRefs(available, node.RequiredEvidence)
	}
	var missing []MissingInput
	for _, want := range required {
		found := false
		for _, have := range available {
			if candidateEvidenceMatchesRequired(have, want) {
				found = true
				break
			}
		}
		if !found {
			key := string(want.Ref)
			if key == "" {
				key = want.Kind
			}
			if key == "" {
				key = "required_evidence"
			}
			missing = AppendMissingInputs(missing, MissingInput("host:objective_graph_evidence_"+key))
		}
	}
	return missing
}

func objectiveGraphCycleNodeRef(graph ObjectiveGraph) DisplaySafeRef {
	nodes := map[DisplaySafeRef]ObjectiveNode{}
	for _, node := range graph.Nodes {
		normalized := node.Normalize()
		nodes[normalized.NodeRef] = normalized
	}
	visiting := map[DisplaySafeRef]bool{}
	visited := map[DisplaySafeRef]bool{}
	var visit func(DisplaySafeRef) DisplaySafeRef
	visit = func(ref DisplaySafeRef) DisplaySafeRef {
		if ref == "" {
			return ""
		}
		if visiting[ref] {
			return ref
		}
		if visited[ref] {
			return ""
		}
		visiting[ref] = true
		node := nodes[ref]
		for _, dependency := range node.Dependencies {
			if cycle := visit(dependency.NodeRef); cycle != "" {
				return cycle
			}
		}
		visiting[ref] = false
		visited[ref] = true
		return ""
	}
	for ref := range nodes {
		if cycle := visit(ref); cycle != "" {
			return cycle
		}
	}
	return ""
}

func objectiveGraphCapabilityKnown(ref DisplaySafeRef, catalog StrategyCatalogSnapshot, descriptors []ObjectiveCapabilityDescriptor) bool {
	ref = normalizeOneDisplaySafeRef(ref)
	if ref == "" {
		return false
	}
	for _, descriptor := range normalizeObjectiveCapabilityDescriptors(descriptors) {
		if descriptor.CapabilityRef == ref {
			return true
		}
	}
	for _, entry := range catalog.Normalize().Entries {
		if displaySafeRefSliceContains(entry.Candidate.CapabilityRefs, ref) {
			return true
		}
	}
	return false
}

func objectiveGraphStrategyCapabilityMatches(strategyRef DisplaySafeRef, capabilityRef DisplaySafeRef, catalog StrategyCatalogSnapshot) bool {
	strategyRef = normalizeOneDisplaySafeRef(strategyRef)
	capabilityRef = normalizeOneDisplaySafeRef(capabilityRef)
	if strategyRef == "" {
		return true
	}
	for _, entry := range catalog.Normalize().Entries {
		ref, ok := NormalizeDisplaySafeRef(entry.Candidate.ID)
		if !ok || ref != strategyRef {
			continue
		}
		if capabilityRef == "" || displaySafeRefSliceContains(entry.Candidate.CapabilityRefs, capabilityRef) {
			return true
		}
	}
	return false
}

func objectiveGraphDescriptorCapabilityMatches(descriptorRef DisplaySafeRef, capabilityRef DisplaySafeRef, descriptors []ObjectiveCapabilityDescriptor) bool {
	descriptorRef = normalizeOneDisplaySafeRef(descriptorRef)
	capabilityRef = normalizeOneDisplaySafeRef(capabilityRef)
	if descriptorRef == "" {
		return true
	}
	for _, descriptor := range normalizeObjectiveCapabilityDescriptors(descriptors) {
		if descriptor.DescriptorRef != descriptorRef {
			continue
		}
		if capabilityRef == "" || descriptor.CapabilityRef == capabilityRef {
			return true
		}
	}
	return false
}

func objectiveGraphDependenciesExist(node ObjectiveNode, graph ObjectiveGraph) bool {
	nodeRefs := map[DisplaySafeRef]struct{}{}
	for _, graphNode := range graph.Nodes {
		normalized := graphNode.Normalize()
		nodeRefs[normalized.NodeRef] = struct{}{}
	}
	for _, dependency := range node.Dependencies {
		if dependency.NodeRef == "" || dependency.NodeRef == node.NodeRef {
			return false
		}
		if _, ok := nodeRefs[dependency.NodeRef]; !ok && !dependency.Optional {
			return false
		}
	}
	return true
}

func objectiveGraphNodeSideEffectAllowed(node ObjectiveNode, spec ObjectiveSpec, policy ExecutionIntensityPolicy) bool {
	sideEffect := NormalizeObjectiveCapabilitySideEffectClass(string(node.SideEffectClass))
	if sideEffect == ObjectiveCapabilitySideEffectUnspecified {
		return false
	}
	if strategySideEffectDenied(policy.Normalize(), spec.Intensity, string(sideEffect)) {
		return false
	}
	if sideEffect == ObjectiveCapabilitySideEffectReadOnly {
		return true
	}
	switch spec.SideEffectPolicy {
	case ObjectiveSpecSideEffectAllowed:
		return true
	case ObjectiveSpecSideEffectRequiresApproval:
		return node.RequiresApproval || len(node.ApprovalRefs) > 0 || len(spec.ApprovalRefs) > 0
	default:
		return false
	}
}

func objectiveGraphCapabilitiesAllowed(graph ObjectiveGraph, allowed []DisplaySafeRef) bool {
	allowed = normalizeDisplaySafeRefs(allowed)
	if len(allowed) == 0 {
		return true
	}
	for _, node := range graph.Normalize().Nodes {
		if node.CapabilityRef == "" {
			continue
		}
		if !displaySafeRefSliceContains(allowed, node.CapabilityRef) {
			return false
		}
	}
	return true
}

func objectiveGraphValidationInputUnsafe(input ObjectiveGraphValidationInput) bool {
	return input.RawOutputLoaded ||
		objectiveGraphUnsafe(input.Graph) ||
		objectiveSpecUnsafe(input.Spec) ||
		input.Catalog.RawOutputLoaded ||
		displaySafeRefRejected(input.SourceRef) ||
		objectiveCapabilityDescriptorSliceUnsafe(input.CapabilityDescriptors)
}

func objectiveGraphJSONDecodeInputUnsafe(input ObjectiveGraphJSONDecodeInput) bool {
	return input.RawOutputLoaded ||
		displaySafeRefRejected(input.GraphRef) ||
		displaySafeRefRejected(input.SourceRef) ||
		displaySafeRefSliceRejected(input.AllowedCapabilityRefs) ||
		objectiveSpecUnsafe(input.Spec) ||
		input.Catalog.RawOutputLoaded ||
		objectiveCapabilityDescriptorSliceUnsafe(input.CapabilityDescriptors)
}

func objectiveGraphBuildInputUnsafe(input ObjectiveGraphBuildInput) bool {
	request := input.Request
	return request.RawOutputLoaded ||
		displaySafeRefRejected(input.GraphRef) ||
		displaySafeRefRejected(input.SourceRef) ||
		displaySafeRefRejected(request.RequestRef) ||
		displaySafeRefSliceRejected(request.AllowedCapabilityRefs) ||
		displaySafeRefSliceRejected(request.PolicyRefs) ||
		objectiveSpecUnsafe(request.Spec) ||
		request.Catalog.RawOutputLoaded ||
		objectiveCapabilityDescriptorSliceUnsafe(request.CapabilityDescriptors)
}

func objectiveGraphUnsafe(graph ObjectiveGraph) bool {
	if graph.RawOutputLoaded ||
		displaySafeRefRejected(graph.GraphRef) ||
		displaySafeRefRejected(graph.SpecRef) ||
		displaySafeRefRejected(graph.CatalogRef) ||
		displaySafeRefSliceRejected(graph.PolicyRefs) ||
		evidenceRefRejected(graph.RequiredEvidence) ||
		ContainsUnsafeRawOutput(graph.ObjectiveID) {
		return true
	}
	if displaySafeRefRejected(graph.Budget.BudgetRef) || displaySafeRefSliceRejected(graph.Budget.PolicyRefs) {
		return true
	}
	for _, node := range graph.Nodes {
		if objectiveNodeUnsafe(node) {
			return true
		}
	}
	return false
}

func objectiveNodeUnsafe(node ObjectiveNode) bool {
	if node.RawOutputLoaded ||
		displaySafeRefRejected(node.NodeRef) ||
		displaySafeRefRejected(node.CapabilityRef) ||
		displaySafeRefRejected(node.StrategyRef) ||
		displaySafeRefRejected(node.DescriptorRef) ||
		displaySafeRefRejected(node.SourceRef) ||
		displaySafeRefRejected(node.InputSchemaRef) ||
		displaySafeRefRejected(node.OutputSchemaRef) ||
		displaySafeRefRejected(node.EvidenceContractRef) ||
		displaySafeRefSliceRejected(node.ApprovalRefs) ||
		displaySafeRefSliceRejected(node.PolicyRefs) ||
		evidenceRefRejected(node.RequiredEvidence) {
		return true
	}
	for _, dependency := range node.Dependencies {
		if displaySafeRefRejected(dependency.DependencyRef) ||
			displaySafeRefRejected(dependency.NodeRef) ||
			evidenceRefRejected(dependency.EvidenceRefs) {
			return true
		}
	}
	return false
}

func objectiveCapabilityDescriptorSliceUnsafe(descriptors []ObjectiveCapabilityDescriptor) bool {
	for _, descriptor := range descriptors {
		if objectiveCapabilityDescriptorUnsafe(descriptor) {
			return true
		}
	}
	return false
}

func objectiveGraphIsEmpty(graph ObjectiveGraph) bool {
	normalized := graph.Normalize()
	return normalized.GraphRef == "" &&
		normalized.SpecRef == "" &&
		normalized.ObjectiveID == "" &&
		normalized.CatalogRef == "" &&
		len(normalized.Nodes) == 0 &&
		len(normalized.RequiredEvidence) == 0
}

func objectiveGraphSafeID(value string) string {
	trimmed := normalizeControlToken(value)
	if trimmed != "" {
		return trimmed
	}
	return ""
}

func cloneObjectiveNodeDependencies(in []ObjectiveNodeDependency) []ObjectiveNodeDependency {
	if len(in) == 0 {
		return nil
	}
	out := make([]ObjectiveNodeDependency, len(in))
	for i, value := range in {
		out[i] = value.Clone()
	}
	return out
}

func normalizeObjectiveNodeDependencies(in []ObjectiveNodeDependency) []ObjectiveNodeDependency {
	out := make([]ObjectiveNodeDependency, 0, len(in))
	seen := map[string]struct{}{}
	for _, value := range in {
		normalized := value.Normalize()
		if normalized.NodeRef == "" {
			continue
		}
		key := string(normalized.DependencyRef) + "|" + string(normalized.NodeRef)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, normalized)
	}
	return out
}

func cloneObjectiveNodes(in []ObjectiveNode) []ObjectiveNode {
	if len(in) == 0 {
		return nil
	}
	out := make([]ObjectiveNode, len(in))
	for i, value := range in {
		out[i] = value.Clone()
	}
	return out
}

func normalizeObjectiveNodes(in []ObjectiveNode) []ObjectiveNode {
	out := make([]ObjectiveNode, 0, len(in))
	seen := map[DisplaySafeRef]struct{}{}
	for _, value := range in {
		normalized := value.Normalize()
		if normalized.NodeRef == "" {
			continue
		}
		if _, exists := seen[normalized.NodeRef]; exists {
			continue
		}
		seen[normalized.NodeRef] = struct{}{}
		out = append(out, normalized)
	}
	return out
}

func cloneObjectiveGraphNodeValidations(in []ObjectiveGraphNodeValidation) []ObjectiveGraphNodeValidation {
	if len(in) == 0 {
		return nil
	}
	out := make([]ObjectiveGraphNodeValidation, len(in))
	for i, value := range in {
		out[i] = value.Clone()
	}
	return out
}

func normalizeObjectiveGraphNodeValidations(in []ObjectiveGraphNodeValidation) []ObjectiveGraphNodeValidation {
	out := make([]ObjectiveGraphNodeValidation, 0, len(in))
	seen := map[DisplaySafeRef]struct{}{}
	for _, value := range in {
		normalized := value.Normalize()
		if normalized.NodeRef == "" {
			continue
		}
		if _, exists := seen[normalized.NodeRef]; exists {
			continue
		}
		seen[normalized.NodeRef] = struct{}{}
		out = append(out, normalized)
	}
	return out
}

func cloneObjectiveCapabilityDescriptors(in []ObjectiveCapabilityDescriptor) []ObjectiveCapabilityDescriptor {
	if len(in) == 0 {
		return nil
	}
	out := make([]ObjectiveCapabilityDescriptor, len(in))
	for i, value := range in {
		out[i] = value.Clone()
	}
	return out
}

func normalizeObjectiveCapabilityDescriptors(in []ObjectiveCapabilityDescriptor) []ObjectiveCapabilityDescriptor {
	out := make([]ObjectiveCapabilityDescriptor, 0, len(in))
	seen := map[DisplaySafeRef]struct{}{}
	for _, value := range in {
		normalized := value.Normalize()
		if normalized.DescriptorRef == "" && normalized.CapabilityRef == "" {
			continue
		}
		key := normalized.DescriptorRef
		if key == "" {
			key = normalized.CapabilityRef
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, normalized)
	}
	return out
}
