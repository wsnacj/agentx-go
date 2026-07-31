// Package lowering 提供 Workflow validation 后的 portable lowering 与
// orchestration plan projection mechanism。
//
// 当前 package 处于 Experimental/private validation。具体 tool/model/task
// mapping、默认值、provider、credential 和 backend policy 必须由 host 注入。
package lowering

import (
	"encoding/json"
	"fmt"

	workflow "github.com/wsnacj/agentx-go/runtime/workflow"
	workflownodeexec "github.com/wsnacj/agentx-go/runtime/workflow/nodeexec"
	workfloworchestration "github.com/wsnacj/agentx-go/runtime/workflow/orchestration"
)

// Validator owns host-selected Workflow and Node admission.
type Validator interface {
	ValidateSpec(workflow.Spec) error
	ValidateNode(workflow.NodeSpec) error
}

// Mapper maps one validated node to a host-selected call and argument object.
type Mapper interface {
	MapNode(workflow.NodeSpec, workflow.ExecutionMode) (MappedCall, error)
}

// MappedCall is the pre-JSON result returned by a host Mapper.
type MappedCall struct {
	Name      string
	Arguments map[string]any
}

// Dependencies contains the two synchronous host capabilities required by
// lowering.
type Dependencies struct {
	Validator Validator
	Mapper    Mapper
}

// Node is one portable lowered node.
type Node struct {
	NodeID         string
	Spec           workflow.NodeSpec
	Kind           workflow.NodeKind
	ExecutionMode  workflow.ExecutionMode
	Call           workflownodeexec.Call
	OriginalConfig map[string]any
}

// Plan is the portable lowering result for one Workflow Spec.
type Plan struct {
	SpecID      string
	Version     string
	EntryNode   string
	Nodes       map[string]Node
	Edges       []workflow.EdgeSpec
	StateSchema []workflow.StateSlotSpec
}

// LowerSpec validates and lowers every node in declaration order.
func LowerSpec(spec workflow.Spec, dependencies Dependencies) (Plan, error) {
	if dependencies.Validator == nil {
		return Plan{}, fmt.Errorf("workflow lowering: validator is required")
	}
	if dependencies.Mapper == nil {
		return Plan{}, fmt.Errorf("workflow lowering: mapper is required")
	}
	if err := dependencies.Validator.ValidateSpec(spec); err != nil {
		return Plan{}, err
	}
	out := Plan{
		SpecID:      spec.ID,
		Version:     spec.Version,
		EntryNode:   spec.EntryNode,
		Nodes:       make(map[string]Node, len(spec.Nodes)),
		Edges:       append([]workflow.EdgeSpec(nil), spec.Edges...),
		StateSchema: spec.StateSchema,
	}
	for _, node := range spec.Nodes {
		lowered, err := LowerNode(node, dependencies)
		if err != nil {
			return Plan{}, fmt.Errorf("workflow: lower node %q: %w", node.ID, err)
		}
		out.Nodes[lowered.NodeID] = lowered
	}
	return out, nil
}

// LowerNode validates and lowers one node.
func LowerNode(node workflow.NodeSpec, dependencies Dependencies) (Node, error) {
	if dependencies.Validator == nil {
		return Node{}, fmt.Errorf("workflow lowering: validator is required")
	}
	if dependencies.Mapper == nil {
		return Node{}, fmt.Errorf("workflow lowering: mapper is required")
	}
	if err := dependencies.Validator.ValidateNode(node); err != nil {
		return Node{}, err
	}
	mode := effectiveExecutionMode(node.ExecutionMode)
	mapped, err := dependencies.Mapper.MapNode(node, mode)
	if err != nil {
		return Node{}, err
	}
	argumentsJSON, err := marshalArguments(mapped.Arguments)
	if err != nil {
		return Node{}, err
	}
	return Node{
		NodeID:         node.ID,
		Spec:           node,
		Kind:           node.Kind,
		ExecutionMode:  mode,
		Call:           workflownodeexec.Call{Name: mapped.Name, Arguments: argumentsJSON},
		OriginalConfig: cloneConfigMap(node.Config),
	}, nil
}

// OrchestrationPlan projects a lowering Plan into the canonical run-loop
// input without selecting an executor or durable backend.
func (p Plan) OrchestrationPlan(workflowID string) workfloworchestration.Plan {
	nodes := make(map[string]workfloworchestration.PlannedNode, len(p.Nodes))
	nodeIDs := make([]string, 0, len(p.Nodes))
	for nodeID, node := range p.Nodes {
		nodeIDs = append(nodeIDs, nodeID)
		nodes[nodeID] = workfloworchestration.PlannedNode{
			Spec:           node.Spec,
			Call:           node.Call,
			Kind:           node.Kind,
			ExecutionMode:  node.ExecutionMode,
			OriginalConfig: node.OriginalConfig,
		}
	}
	return workfloworchestration.Plan{
		WorkflowID:  workflowID,
		Version:     p.Version,
		EntryNode:   p.EntryNode,
		NodeIDs:     nodeIDs,
		Nodes:       nodes,
		Edges:       p.Edges,
		StateSchema: p.StateSchema,
	}
}

func marshalArguments(arguments map[string]any) (string, error) {
	if len(arguments) == 0 {
		return "{}", nil
	}
	raw, err := json.Marshal(arguments)
	if err != nil {
		return "", fmt.Errorf("marshal tool arguments: %w", err)
	}
	return string(raw), nil
}

func effectiveExecutionMode(mode workflow.ExecutionMode) workflow.ExecutionMode {
	if string(mode) == "" {
		return workflow.ExecInline
	}
	return mode
}

func cloneConfigMap(config map[string]any) map[string]any {
	if len(config) == 0 {
		return nil
	}
	out := make(map[string]any, len(config))
	for key, value := range config {
		out[key] = value
	}
	return out
}
