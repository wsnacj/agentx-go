package main

import (
	"encoding/json"
	"fmt"
	"os"

	pack "github.com/wsnacj/agentx-go/extensions/pack"
	workflow "github.com/wsnacj/agentx-go/runtime/workflow"
)

type validator struct{}

func (validator) ValidateSpec(spec workflow.Spec) error {
	if spec.ID == "" {
		return fmt.Errorf("workflow id is required")
	}
	return nil
}

type lowerer struct{}

func (lowerer) LowerToolArguments(node workflow.NodeSpec) (string, error) {
	args, _ := node.Config["args"].(map[string]any)
	encoded, err := json.Marshal(args)
	return string(encoded), err
}

func run() (string, error) {
	coordinator, err := pack.NewCoordinator(validator{}, lowerer{})
	if err != nil {
		return "", err
	}
	registry, err := pack.NewMemoryRegistry(coordinator)
	if err != nil {
		return "", err
	}
	definition := pack.Definition{
		Manifest: pack.Manifest{
			ID:                 "portable-research",
			Version:            "1.0.0",
			Domain:             "research",
			RouteHints:         []string{"研究资料"},
			SupportedCaseTypes: []string{"research.lookup"},
			DefaultWorkflow:    "collect-v1",
		},
		Workflows: []workflow.Spec{{
			ID:         "collect-v1",
			Pack:       "portable-research",
			CaseTypes:  []string{"research.lookup"},
			RouteHints: []string{"收集资料"},
			EntryNode:  "collect",
			Nodes: []workflow.NodeSpec{{
				ID:     "collect",
				Kind:   workflow.NodeTool,
				Config: map[string]any{"tool_name": "collect"},
			}},
		}},
		Tools: []pack.SemanticTool{{Name: "collect", RuntimeTool: "host_collect"}},
	}
	if err := registry.Register(definition); err != nil {
		return "", err
	}
	selection, matched := pack.SelectBinding(registry, "请研究资料并收集资料", pack.SelectOptions{})
	if !matched {
		return "", fmt.Errorf("pack selection did not match: %#v", selection)
	}
	binding, ok, err := coordinator.ResolveBinding(registry, selection.Selected.PackID, selection.Selected.CaseType, selection.Selected.WorkflowID)
	if err != nil {
		return "", err
	}
	if !ok || len(binding.Workflow.Nodes) != 1 {
		return "", fmt.Errorf("pack binding is unavailable")
	}
	tool, _ := binding.Workflow.Nodes[0].Config["tool"].(string)
	return fmt.Sprintf("agentx-pack-ok:%s:%s:%s", binding.PackID, binding.WorkflowID, tool), nil
}

func main() {
	result, err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(result)
}
