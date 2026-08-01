package pack_test

import (
	"encoding/json"
	"testing"

	pack "github.com/wsnacj/agentx-go/extensions/pack"
	workflow "github.com/wsnacj/agentx-go/runtime/workflow"
)

type externalValidator struct{}

func (externalValidator) ValidateSpec(workflow.Spec) error { return nil }

type externalLowerer struct{}

func (externalLowerer) LowerToolArguments(node workflow.NodeSpec) (string, error) {
	args, _ := node.Config["args"].(map[string]any)
	encoded, err := json.Marshal(args)
	return string(encoded), err
}

func TestExternalConstructionIsExplicit(t *testing.T) {
	coordinator, err := pack.NewCoordinator(externalValidator{}, externalLowerer{})
	if err != nil {
		t.Fatalf("NewCoordinator(): %v", err)
	}
	registry, err := pack.NewMemoryRegistry(coordinator)
	if err != nil {
		t.Fatalf("NewMemoryRegistry(): %v", err)
	}
	if got := registry.List(); len(got) != 0 {
		t.Fatalf("empty List() = %#v", got)
	}
}
