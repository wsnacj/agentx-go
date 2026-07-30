// Package transition 提供与执行 substrate 无关的 Workflow 遍历、状态归一化和
// edge routing 机制。
//
// 当前 package 处于 Experimental/private validation。它不拥有 validation、
// lowering、node executor、RunStore、durable lifecycle、retry、cancellation、
// provider 或 Scene policy。
package transition

import (
	"fmt"

	workflow "github.com/wsnacj/agentx-go/runtime/workflow"
)

// Trigger selects the outgoing edge class after a node execution.
type Trigger string

const (
	// TriggerSuccess selects success and always edges.
	TriggerSuccess Trigger = "success"
	// TriggerFailure selects failure and always edges.
	TriggerFailure Trigger = "failure"
)

// Plan is the substrate-neutral graph projection required by Machine.
type Plan struct {
	EntryNode string
	NodeIDs   []string
	Edges     []workflow.EdgeSpec
}

// Machine owns traversal state for one workflow execution.
//
// Machine is not safe for concurrent use.
type Machine struct {
	current string
	nodes   map[string]struct{}
	edges   []workflow.EdgeSpec
	visited map[string]bool
}

// New constructs an isolated traversal machine.
func New(plan Plan) *Machine {
	nodes := make(map[string]struct{}, len(plan.NodeIDs))
	for _, nodeID := range plan.NodeIDs {
		nodes[nodeID] = struct{}{}
	}
	return &Machine{
		current: plan.EntryNode,
		nodes:   nodes,
		edges:   append([]workflow.EdgeSpec(nil), plan.Edges...),
		visited: map[string]bool{},
	}
}

// Enter returns the current node after enforcing exact node identity and
// single-pass traversal.
func (m *Machine) Enter() (string, error) {
	if m == nil || m.current == "" {
		return "", nil
	}
	current := m.current
	if m.visited[current] {
		return "", fmt.Errorf("workflow: detected cycle at node %q during inline execution", current)
	}
	m.visited[current] = true
	if _, ok := m.nodes[current]; !ok {
		return "", fmt.Errorf("workflow: lowered node %q missing", current)
	}
	return current, nil
}

// Advance selects the unique edge matching trigger, updates the current node,
// and returns the selected node ID. An empty result means terminal execution.
func (m *Machine) Advance(trigger Trigger) (string, error) {
	if m == nil || m.current == "" {
		return "", nil
	}
	current := m.current
	matches := make([]workflow.EdgeSpec, 0, 1)
	for _, edge := range m.edges {
		if edge.From != current {
			continue
		}
		edgeTrigger := edge.On
		if edgeTrigger == "" {
			edgeTrigger = string(TriggerSuccess)
		}
		if edgeTrigger == "always" || edgeTrigger == string(trigger) {
			matches = append(matches, edge)
		}
	}
	if len(matches) > 1 {
		return "", fmt.Errorf(
			"workflow: node %q has multiple outgoing %s edges; inline executor only supports a single path",
			current,
			trigger,
		)
	}
	if len(matches) == 0 {
		m.current = ""
		return "", nil
	}
	m.current = matches[0].To
	return m.current, nil
}

// NormalizeFinalStatus preserves recognized exact status values and maps all
// other successful outcomes to completed. A dependency error always wins.
func NormalizeFinalStatus(finalStatus string, failed bool) string {
	switch finalStatus {
	case "completed", "failed", "incomplete":
		if failed {
			return "failed"
		}
		return finalStatus
	}
	if failed {
		return "failed"
	}
	return "completed"
}
