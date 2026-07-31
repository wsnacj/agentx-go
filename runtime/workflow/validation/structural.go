package validation

import (
	"fmt"
	"strings"

	workflow "github.com/wsnacj/agentx-go/runtime/workflow"
	workflowbindingstate "github.com/wsnacj/agentx-go/runtime/workflow/bindingstate"
)

// ValidateTrimmedField 接受空值或没有外围空白的字段。
func ValidateTrimmedField(value string, label string) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	if value != trimmed {
		return fmt.Errorf("workflow: %s %q must not include surrounding whitespace", label, value)
	}
	return nil
}

// ValidateOptionalField 接受空值，拒绝 whitespace-only 和外围空白字段。
func ValidateOptionalField(value string, label string) error {
	if value == "" {
		return nil
	}
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("workflow: %s %q must not be whitespace-only", label, value)
	}
	return ValidateTrimmedField(value, label)
}

type bindingValidationGraph struct {
	entryNode    string
	successors   map[string][]string
	predecessors map[string][]string
	reachable    map[string]bool
	dominators   map[string]map[string]bool
}

func newBindingValidationGraph(entryNode string, seen map[string]bool) bindingValidationGraph {
	graph := bindingValidationGraph{
		entryNode:    strings.TrimSpace(entryNode),
		successors:   map[string][]string{},
		predecessors: map[string][]string{},
		reachable:    map[string]bool{},
		dominators:   map[string]map[string]bool{},
	}
	for nodeID := range seen {
		graph.successors[nodeID] = nil
		graph.predecessors[nodeID] = nil
	}
	return graph
}

func (g *bindingValidationGraph) addEdge(from string, to string) {
	if g == nil {
		return
	}
	g.successors[from] = append(g.successors[from], to)
	g.predecessors[to] = append(g.predecessors[to], from)
}

func (g *bindingValidationGraph) resolve() {
	if g == nil {
		return
	}
	g.reachable = reachableNodeSet(g.entryNode, g.successors)
	if len(g.reachable) == 0 {
		g.dominators = map[string]map[string]bool{}
		return
	}
	g.dominators = map[string]map[string]bool{}
	for nodeID := range g.reachable {
		if nodeID == g.entryNode {
			g.dominators[nodeID] = map[string]bool{nodeID: true}
			continue
		}
		all := make(map[string]bool, len(g.reachable))
		for candidate := range g.reachable {
			all[candidate] = true
		}
		g.dominators[nodeID] = all
	}
	changed := true
	for changed {
		changed = false
		for nodeID := range g.reachable {
			if nodeID == g.entryNode {
				continue
			}
			next := map[string]bool{nodeID: true}
			predSet := reachablePredecessorDominators(g.predecessors[nodeID], g.reachable, g.dominators)
			for dominatedNode := range predSet {
				next[dominatedNode] = true
			}
			if !sameNodeSet(g.dominators[nodeID], next) {
				g.dominators[nodeID] = next
				changed = true
			}
		}
	}
}

func reachableNodeSet(entryNode string, successors map[string][]string) map[string]bool {
	entryNode = strings.TrimSpace(entryNode)
	if entryNode == "" {
		return map[string]bool{}
	}
	visited := map[string]bool{}
	queue := []string{entryNode}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if visited[current] {
			continue
		}
		visited[current] = true
		for _, next := range successors[current] {
			next = strings.TrimSpace(next)
			if next == "" || visited[next] {
				continue
			}
			queue = append(queue, next)
		}
	}
	return visited
}

func reachablePredecessorDominators(predecessors []string, reachable map[string]bool, dominators map[string]map[string]bool) map[string]bool {
	var intersection map[string]bool
	for _, predecessor := range predecessors {
		predecessor = strings.TrimSpace(predecessor)
		if predecessor == "" || !reachable[predecessor] {
			continue
		}
		current := dominators[predecessor]
		if len(current) == 0 {
			continue
		}
		if intersection == nil {
			intersection = cloneNodeSet(current)
			continue
		}
		for nodeID := range intersection {
			if !current[nodeID] {
				delete(intersection, nodeID)
			}
		}
	}
	if intersection == nil {
		return map[string]bool{}
	}
	return intersection
}

func cloneNodeSet(value map[string]bool) map[string]bool {
	if len(value) == 0 {
		return map[string]bool{}
	}
	out := make(map[string]bool, len(value))
	for key, present := range value {
		if present {
			out[key] = true
		}
	}
	return out
}

func sameNodeSet(left map[string]bool, right map[string]bool) bool {
	if len(left) != len(right) {
		return false
	}
	for key := range left {
		if !right[key] {
			return false
		}
	}
	return true
}

func (g *bindingValidationGraph) firstReachableCycle() (string, bool) {
	if g == nil || len(g.reachable) == 0 {
		return "", false
	}
	visiting := map[string]bool{}
	visited := map[string]bool{}
	var walk func(string) (string, bool)
	walk = func(nodeID string) (string, bool) {
		nodeID = strings.TrimSpace(nodeID)
		if nodeID == "" || !g.reachable[nodeID] {
			return "", false
		}
		if visiting[nodeID] {
			return nodeID, true
		}
		if visited[nodeID] {
			return "", false
		}
		visiting[nodeID] = true
		for _, next := range g.successors[nodeID] {
			if cycleNode, ok := walk(next); ok {
				return cycleNode, true
			}
		}
		delete(visiting, nodeID)
		visited[nodeID] = true
		return "", false
	}
	return walk(g.entryNode)
}

func isKnownNodeKind(kind workflow.NodeKind) bool {
	switch kind {
	case workflow.NodeTool, workflow.NodeLLM, workflow.NodeAgent, workflow.NodeParallel, workflow.NodeCollect, workflow.NodeWait, workflow.NodeEvaluate, workflow.NodeApprove, workflow.NodeSubflow, workflow.NodeHumanInput:
		return true
	default:
		return false
	}
}

func validateBindings(bindings []workflow.BindingSpec, label string, policy Policy) error {
	for _, binding := range bindings {
		if strings.TrimSpace(binding.From) == "" {
			return fmt.Errorf("%s binding from is required", label)
		}
		if strings.TrimSpace(binding.To) == "" {
			return fmt.Errorf("%s binding to is required", label)
		}
		if err := validateBindingPathSegments(binding.From, label, "source"); err != nil {
			return err
		}
		if err := validateBindingPathSegments(binding.To, label, "target"); err != nil {
			return err
		}
		if err := policy.ValidateBindingTargetShape(binding.To, label); err != nil {
			return err
		}
	}
	return nil
}

func validateBindingPathSegments(path string, label string, field string) error {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return nil
	}
	if path != trimmed {
		return fmt.Errorf("%s binding %s %q must not include surrounding whitespace", label, field, path)
	}
	for _, segment := range strings.Split(trimmed, ".") {
		segmentTrimmed := strings.TrimSpace(segment)
		if segmentTrimmed == "" {
			return fmt.Errorf("%s binding %s %q contains empty path segment", label, field, trimmed)
		}
		if segment != segmentTrimmed {
			return fmt.Errorf("%s binding %s %q contains path segment with surrounding whitespace", label, field, trimmed)
		}
	}
	return nil
}

func validateStateSlotName(name string) error {
	return workflowbindingstate.ValidateSlotName(name)
}

func splitBindingPath(path string) []string {
	if path == "" {
		return nil
	}
	return strings.Split(path, ".")
}

func validateNodeBindingReferences(node workflow.NodeSpec, seen map[string]bool, graph bindingValidationGraph, policy Policy) error {
	nodeID := strings.TrimSpace(node.ID)
	for _, binding := range node.Inputs {
		if err := validateBindingSourceReference(binding.From, nodeID, seen, graph, "input", policy); err != nil {
			return err
		}
	}
	for _, binding := range node.Outputs {
		if err := validateBindingSourceReference(binding.From, nodeID, seen, graph, "output", policy); err != nil {
			return err
		}
	}
	return nil
}

func validateBindingSourceReference(source string, currentNodeID string, seen map[string]bool, graph bindingValidationGraph, label string, policy Policy) error {
	source = strings.TrimSpace(source)
	if err := policy.ValidateBindingSourceShape(source, label); err != nil {
		return err
	}
	if !strings.HasPrefix(source, "node.") {
		return nil
	}
	parts := splitBindingPath(source)
	if len(parts) < 3 {
		return fmt.Errorf("%s binding source %q must include node id and field", label, source)
	}
	nodeID := strings.TrimSpace(parts[1])
	if nodeID == "" {
		return fmt.Errorf("%s binding source %q must include node id", label, source)
	}
	if !seen[nodeID] {
		return fmt.Errorf("%s binding source references unknown node %q", label, nodeID)
	}
	currentNodeID = strings.TrimSpace(currentNodeID)
	if currentNodeID == "" {
		return nil
	}
	if nodeID == currentNodeID {
		return fmt.Errorf("%s binding source references current node %q before it has recorded output", label, nodeID)
	}
	if !graph.reachable[currentNodeID] {
		return nil
	}
	if !graph.reachable[nodeID] {
		return fmt.Errorf("%s binding source references unreachable node %q", label, nodeID)
	}
	if !graph.dominators[currentNodeID][nodeID] {
		return fmt.Errorf("%s binding source references node %q which cannot execute before node %q", label, nodeID, currentNodeID)
	}
	return nil
}
