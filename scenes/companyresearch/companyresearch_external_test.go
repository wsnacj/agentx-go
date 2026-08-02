package companyresearch_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/wsnacj/agentx-go/extensions/pack"
	"github.com/wsnacj/agentx-go/runtime/workflow"
	companyresearch "github.com/wsnacj/agentx-go/scenes/companyresearch"
	"github.com/wsnacj/agentx-go/scenes/companyresearch/hostkit"
)

type validator struct{}

func (validator) ValidateSpec(workflow.Spec) error { return nil }

type lowerer struct{}

func (lowerer) LowerToolArguments(node workflow.NodeSpec) (string, error) {
	arguments, _ := node.Config["args"].(map[string]any)
	encoded, err := json.Marshal(arguments)
	return string(encoded), err
}

func TestPortableCompanyResearchContract(t *testing.T) {
	coordinator, err := pack.NewCoordinator(validator{}, lowerer{})
	if err != nil {
		t.Fatalf("NewCoordinator(): %v", err)
	}
	registry, err := pack.NewMemoryRegistry(coordinator)
	if err != nil {
		t.Fatalf("NewMemoryRegistry(): %v", err)
	}
	if err := companyresearch.RegisterInto(registry); err != nil {
		t.Fatalf("RegisterInto(): %v", err)
	}
	for name, materialize := range map[string]func(*pack.Coordinator) (workflow.Spec, error){
		companyresearch.DefaultWorkflow: companyresearch.MaterializedDefaultWorkflow,
		companyresearch.CompareWorkflow: companyresearch.MaterializedCompareWorkflow,
	} {
		spec, err := materialize(coordinator)
		if err != nil || spec.ID != name {
			t.Fatalf("materialize %s = %#v, %v", name, spec, err)
		}
	}

	payload, err := hostkit.BuildCompanyResearchLookupPayload(context.Background(), hostkit.CompanyResearchConfig{
		Handlers: hostkit.CompanyResearchHandlers{
			Finance: func(context.Context, map[string]any) (any, error) {
				return map[string]any{"adapter_status": "ok", "source_url": "https://example.com/filing"}, nil
			},
		},
	}, map[string]any{
		"user_message":         "研究示例公司的最新财报",
		"entity_name":          "示例公司",
		"requested_dimensions": []any{"financials"},
	})
	if err != nil || payload.Tool != companyresearch.ToolCompanyResearchLookup {
		t.Fatalf("BuildCompanyResearchLookupPayload() = %#v, %v", payload, err)
	}
}
