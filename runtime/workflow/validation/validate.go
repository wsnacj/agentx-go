package validation

import (
	"fmt"
	"strings"

	workflow "github.com/wsnacj/agentx-go/runtime/workflow"
)

// ValidateSpec 按稳定顺序执行 Workflow Spec structural validation，并在既有
// policy 点调用 host Policy。nil Policy 会 fail closed。
func ValidateSpec(spec workflow.Spec, policy Policy) error {
	if policy == nil {
		return fmt.Errorf("workflow validation: policy is required")
	}
	return validateSpecWithPolicy(spec, policy)
}

func validateSpecWithPolicy(spec workflow.Spec, policy Policy) error {
	if err := ValidateTrimmedField(spec.ID, "spec id"); err != nil {
		return err
	}
	if err := ValidateOptionalField(spec.Title, "title"); err != nil {
		return err
	}
	if err := ValidateOptionalField(spec.Description, "description"); err != nil {
		return err
	}
	if err := ValidateOptionalField(spec.Version, "version"); err != nil {
		return err
	}
	if err := ValidateOptionalField(string(spec.PlanningMode), "planning_mode"); err != nil {
		return err
	}
	if strings.TrimSpace(spec.ID) == "" {
		return fmt.Errorf("workflow: spec id is required")
	}
	if err := policy.ValidatePackScopedContractUsage(spec); err != nil {
		return err
	}
	if err := policy.ValidatePackScopedWorkflowMetadataUsage(spec); err != nil {
		return err
	}
	if planningMode := strings.TrimSpace(string(spec.PlanningMode)); planningMode != "" {
		switch spec.PlanningMode {
		case workflow.PlanningOpen, workflow.PlanningBounded, workflow.PlanningPlanless:
		default:
			return fmt.Errorf("workflow: unsupported planning mode %q", planningMode)
		}
	}
	if err := ValidateTrimmedField(spec.EntryNode, "entry node"); err != nil {
		return err
	}
	if strings.TrimSpace(spec.EntryNode) == "" {
		return fmt.Errorf("workflow: entry node is required")
	}
	if len(spec.Nodes) == 0 {
		return fmt.Errorf("workflow: at least one node is required")
	}
	stateSlots := map[string]bool{}
	for _, slot := range spec.StateSchema {
		if err := validateStateSlotName(slot.Name); err != nil {
			return err
		}
		name := strings.TrimSpace(slot.Name)
		if stateSlots[name] {
			return fmt.Errorf("workflow: duplicate state slot %q", name)
		}
		stateSlots[name] = true
	}
	seen := map[string]bool{}
	hasEntry := false
	for _, node := range spec.Nodes {
		if err := ValidateTrimmedField(node.ID, "node id"); err != nil {
			return err
		}
		id := strings.TrimSpace(node.ID)
		if id == "" {
			return fmt.Errorf("workflow: node id is required")
		}
		if strings.Contains(id, ".") {
			return fmt.Errorf("workflow: node id %q cannot contain \".\" because binding sources use dot-delimited node.<id> paths", id)
		}
		if seen[id] {
			return fmt.Errorf("workflow: duplicate node id %q", id)
		}
		seen[id] = true
		if err := validateNodeSpecWithPolicy(node, policy); err != nil {
			return fmt.Errorf("workflow: node %q invalid: %w", id, err)
		}
		if id == strings.TrimSpace(spec.EntryNode) {
			hasEntry = true
		}
	}
	if !hasEntry {
		return fmt.Errorf("workflow: entry node %q not found", strings.TrimSpace(spec.EntryNode))
	}
	graph := newBindingValidationGraph(strings.TrimSpace(spec.EntryNode), seen)
	for _, edge := range spec.Edges {
		if err := ValidateTrimmedField(edge.From, "edge from"); err != nil {
			return err
		}
		if err := ValidateTrimmedField(edge.To, "edge to"); err != nil {
			return err
		}
		if err := ValidateOptionalField(edge.On, "edge trigger"); err != nil {
			return err
		}
		if err := ValidateOptionalField(edge.Condition, "edge condition"); err != nil {
			return err
		}
		from := strings.TrimSpace(edge.From)
		to := strings.TrimSpace(edge.To)
		if from == "" || to == "" {
			return fmt.Errorf("workflow: edge from/to is required")
		}
		if !seen[from] {
			return fmt.Errorf("workflow: edge from node %q not found", from)
		}
		if !seen[to] {
			return fmt.Errorf("workflow: edge to node %q not found", to)
		}
		if err := policy.ValidateEdgeRuntimeCapabilities(edge, from, to); err != nil {
			return err
		}
		graph.addEdge(from, to)
	}
	graph.resolve()
	if err := policy.ValidateLinearRuntimeEdgeDeterminism(spec.Edges); err != nil {
		return err
	}
	for _, node := range spec.Nodes {
		nodeID := strings.TrimSpace(node.ID)
		if nodeID == "" {
			continue
		}
		if graph.reachable[nodeID] {
			continue
		}
		return fmt.Errorf("workflow: node %q is unreachable from entry node %q", nodeID, strings.TrimSpace(spec.EntryNode))
	}
	if cycleNode, ok := graph.firstReachableCycle(); ok {
		if err := policy.ValidateReachableCycleRuntimeCapability(cycleNode); err != nil {
			return err
		}
	}
	for _, node := range spec.Nodes {
		if err := validateNodeBindingReferences(node, seen, graph, policy); err != nil {
			return fmt.Errorf("workflow: node %q invalid: %w", strings.TrimSpace(node.ID), err)
		}
	}
	return nil
}

// ValidateNodeSpec 验证单个 NodeSpec 的 structural shape，并调用 node、
// binding 与 config policy。nil Policy 会 fail closed。
func ValidateNodeSpec(node workflow.NodeSpec, policy Policy) error {
	if policy == nil {
		return fmt.Errorf("workflow validation: policy is required")
	}
	return validateNodeSpecWithPolicy(node, policy)
}

func validateNodeSpecWithPolicy(node workflow.NodeSpec, policy Policy) error {
	if err := ValidateOptionalField(string(node.ExecutionMode), "execution_mode"); err != nil {
		return err
	}
	if err := ValidateOptionalField(node.Title, "node title"); err != nil {
		return err
	}
	if err := ValidateOptionalField(node.Description, "node description"); err != nil {
		return err
	}
	if !isKnownNodeKind(node.Kind) {
		return fmt.Errorf("kind is required and must be supported")
	}
	if err := policy.ValidateNodeRuntimeCapabilities(node); err != nil {
		return err
	}
	if err := validateBindings(node.Inputs, "input", policy); err != nil {
		return err
	}
	if err := validateBindings(node.Outputs, "output", policy); err != nil {
		return err
	}
	return policy.ValidateNodeConfig(node)
}
