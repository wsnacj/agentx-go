package pack

import (
	"encoding/json"
	"errors"
	"reflect"
	"sync"
	"testing"

	executionpolicy "github.com/wsnacj/agentx-go/runtime/executionpolicy"
	workflow "github.com/wsnacj/agentx-go/runtime/workflow"
)

type validatorFunc func(workflow.Spec) error

func (f validatorFunc) ValidateSpec(spec workflow.Spec) error { return f(spec) }

type lowererFunc func(workflow.NodeSpec) (string, error)

func (f lowererFunc) LowerToolArguments(node workflow.NodeSpec) (string, error) { return f(node) }

func testCoordinator(t *testing.T) *Coordinator {
	t.Helper()
	coordinator, err := NewCoordinator(
		validatorFunc(func(workflow.Spec) error { return nil }),
		lowererFunc(func(node workflow.NodeSpec) (string, error) {
			args, _ := node.Config["args"].(map[string]any)
			encoded, err := json.Marshal(args)
			return string(encoded), err
		}),
	)
	if err != nil {
		t.Fatalf("NewCoordinator(): %v", err)
	}
	return coordinator
}

func testDefinition() Definition {
	return Definition{
		Manifest: Manifest{
			ID:                 "research-pack",
			Version:            "1.0.0",
			Domain:             "research",
			RouteHints:         []string{"研究 agentx"},
			SupportedCaseTypes: []string{"research.lookup"},
			DefaultWorkflow:    "collect-v1",
		},
		CaseSchemas: []CaseSchema{{
			CaseType:   "research.lookup",
			RouteHints: []string{"查询资料"},
			Schema: map[string]any{
				"type":       "object",
				"properties": map[string]any{"query": map[string]any{"type": "string"}},
				"required":   []string{"query"},
			},
		}},
		Workflows: []workflow.Spec{{
			ID:              "collect-v1",
			Title:           "收集研究资料",
			Pack:            "research-pack",
			CaseTypes:       []string{"research.lookup"},
			RouteHints:      []string{"收集资料"},
			EntryNode:       "collect",
			DefaultContract: "readonly",
			Nodes: []workflow.NodeSpec{{
				ID:   "collect",
				Kind: workflow.NodeTool,
				Config: map[string]any{
					"tool_name": "collect",
					"args":      map[string]any{"query": "agentx"},
				},
			}},
		}},
		Tools: []SemanticTool{{
			Name:        "collect",
			RuntimeTool: "host_collect",
			RuntimeArgs: map[string]any{"limit": float64(2)},
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"query": map[string]any{"type": "string"},
					"limit": map[string]any{"type": "number"},
				},
				"required": []string{"query", "limit"},
			},
		}},
		PolicyProfiles: []PolicyProfile{{
			Name:    "readonly",
			Default: true,
			Contract: executionpolicy.Contract{
				ID:         "readonly",
				Visibility: executionpolicy.VisibilityPolicy{AllowTools: []string{"host_collect"}},
			},
		}},
	}
}

func TestCoordinatorRegistryMaterializationBindingAndSelection(t *testing.T) {
	coordinator := testCoordinator(t)
	registry, err := NewMemoryRegistry(coordinator)
	if err != nil {
		t.Fatalf("NewMemoryRegistry(): %v", err)
	}
	def := testDefinition()
	if err := registry.Register(def); err != nil {
		t.Fatalf("Register(): %v", err)
	}

	got, ok := registry.Get(def.Manifest.ID)
	if !ok || !reflect.DeepEqual(got, def) {
		t.Fatalf("Get() = %#v, %v", got, ok)
	}
	got.Manifest.Domain = "mutated"
	again, _ := registry.Get(def.Manifest.ID)
	if again.Manifest.Domain != "research" {
		t.Fatal("registry returned mutable source authority")
	}

	materialized, ok, err := registry.ResolveMaterializedWorkflow(def.Manifest.ID, "research.lookup")
	if err != nil || !ok {
		t.Fatalf("ResolveMaterializedWorkflow() = %#v, %v, %v", materialized, ok, err)
	}
	config := materialized.Nodes[0].Config
	if config["tool"] != "host_collect" {
		t.Fatalf("materialized config = %#v", config)
	}
	args, _ := config["args"].(map[string]any)
	if args["query"] != "agentx" || args["limit"] != float64(2) {
		t.Fatalf("materialized args = %#v", args)
	}

	binding, ok, err := coordinator.ResolveBinding(registry, def.Manifest.ID, "research.lookup", "")
	if err != nil || !ok {
		t.Fatalf("ResolveBinding() = %#v, %v, %v", binding, ok, err)
	}
	if binding.PolicyProfile == nil || binding.PolicyProfile.Contract.ID != "readonly" {
		t.Fatalf("binding policy = %#v", binding.PolicyProfile)
	}
	if err := binding.ValidateCaseInput(map[string]any{"query": "agentx"}); err != nil {
		t.Fatalf("ValidateCaseInput(): %v", err)
	}
	if err := binding.ValidateCaseInput(map[string]any{}); err == nil {
		t.Fatal("ValidateCaseInput(empty) error = nil")
	}

	selection, matched := SelectBinding(registry, "请研究 agentx 并收集资料", SelectOptions{})
	if !matched || !selection.Matched || selection.Selected.PackID != def.Manifest.ID {
		t.Fatalf("SelectBinding() = %#v, %v", selection, matched)
	}
}

func TestCoordinatorPreservesValidatorAndLowererErrors(t *testing.T) {
	validationErr := errors.New("host validation sentinel")
	coordinator, err := NewCoordinator(
		validatorFunc(func(workflow.Spec) error { return validationErr }),
		lowererFunc(func(workflow.NodeSpec) (string, error) { return `{}`, nil }),
	)
	if err != nil {
		t.Fatalf("NewCoordinator(): %v", err)
	}
	err = coordinator.ValidateDefinition(testDefinition())
	if !errors.Is(err, validationErr) || err.Error() != `pack: workflow "collect-v1" invalid: host validation sentinel` {
		t.Fatalf("ValidateDefinition() error = %v", err)
	}

	loweringErr := errors.New("host lowering sentinel")
	coordinator, err = NewCoordinator(
		validatorFunc(func(workflow.Spec) error { return nil }),
		lowererFunc(func(workflow.NodeSpec) (string, error) { return "", loweringErr }),
	)
	if err != nil {
		t.Fatalf("NewCoordinator(): %v", err)
	}
	_, err = coordinator.MaterializeWorkflow(testDefinition(), "")
	if !errors.Is(err, loweringErr) || err.Error() != "pack: materialize semantic tool \"collect\" for node \"collect\": lower semantic tool payload: host lowering sentinel" {
		t.Fatalf("MaterializeWorkflow() error = %v", err)
	}
}

func TestCoordinatorFailsClosedAndRegistryIsConcurrent(t *testing.T) {
	if _, err := NewCoordinator(nil, lowererFunc(func(workflow.NodeSpec) (string, error) { return `{}`, nil })); err == nil {
		t.Fatal("NewCoordinator(nil validator) error = nil")
	}
	if _, err := NewCoordinator(validatorFunc(func(workflow.Spec) error { return nil }), nil); err == nil {
		t.Fatal("NewCoordinator(nil lowerer) error = nil")
	}

	registry, err := NewMemoryRegistry(testCoordinator(t))
	if err != nil {
		t.Fatalf("NewMemoryRegistry(): %v", err)
	}
	if err := registry.Register(testDefinition()); err != nil {
		t.Fatalf("Register(): %v", err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, ok := registry.Get("research-pack"); !ok {
				t.Error("concurrent Get() missed definition")
			}
			if len(registry.List()) != 1 {
				t.Error("concurrent List() changed size")
			}
		}()
	}
	wg.Wait()
}
