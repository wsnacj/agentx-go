package astock_test

import (
	"encoding/json"
	"io/fs"
	"testing"

	astock "github.com/wsnacj/agentx-go/scenes/astock"
	astockcontracts "github.com/wsnacj/agentx-go/scenes/astock/contracts"
	"github.com/wsnacj/agentx-go/extensions/pack"
	"github.com/wsnacj/agentx-go/runtime/workflow"
)

type validator struct{}

func (validator) ValidateSpec(workflow.Spec) error { return nil }

type lowerer struct{}

func (lowerer) LowerToolArguments(node workflow.NodeSpec) (string, error) {
	arguments, _ := node.Config["args"].(map[string]any)
	encoded, err := json.Marshal(arguments)
	return string(encoded), err
}

func TestPortableAStockExtensionEndToEnd(t *testing.T) {
	manifest := astock.Manifest()
	if manifest.ID != astock.ModuleID || len(manifest.Packs) != 3 || len(manifest.Tools) != 7 {
		t.Fatalf("Manifest() = %#v", manifest)
	}

	provider := astock.Assets()
	if provider.IsZero() || provider.Fingerprint() == "" {
		t.Fatalf("Assets() = %#v", provider)
	}
	for _, path := range []string{
		"skills/a-stock-data/SKILL.md",
		"tools/a_stock_quote_lookup.tool.json",
		"tools/a_stock_signal_lookup.tool.json",
	} {
		if content, err := fs.ReadFile(astock.ExtensionFS(), path); err != nil || len(content) == 0 {
			t.Fatalf("read %s: bytes=%d err=%v", path, len(content), err)
		}
	}

	coordinator, err := pack.NewCoordinator(validator{}, lowerer{})
	if err != nil {
		t.Fatalf("NewCoordinator(): %v", err)
	}
	registry, err := pack.NewMemoryRegistry(coordinator)
	if err != nil {
		t.Fatalf("NewMemoryRegistry(): %v", err)
	}
	if err := astock.RegisterPacks(registry); err != nil {
		t.Fatalf("RegisterPacks(): %v", err)
	}
	if definitions := registry.List(); len(definitions) != 3 {
		t.Fatalf("registered definitions = %d", len(definitions))
	}
	selection, matched := pack.SelectBinding(registry, "请做 A股估值 行情快照 valuation snapshot", pack.SelectOptions{})
	if !matched || selection.Selected.PackID != astock.ValuationPackID {
		t.Fatalf("SelectBinding() = %#v, %v", selection, matched)
	}
	binding, ok, err := coordinator.ResolveBinding(
		registry,
		selection.Selected.PackID,
		selection.Selected.CaseType,
		selection.Selected.WorkflowID,
	)
	if err != nil || !ok || binding.Workflow.ID != astock.ValuationDefaultWorkflow {
		t.Fatalf("ResolveBinding() = %#v, %v, %v", binding, ok, err)
	}

	toolDefinitions := astock.ToolDefinitions()
	if len(toolDefinitions) != len(astock.ToolNames()) || toolDefinitions[0].Function.Name != astock.ToolAStockInvestigation {
		t.Fatalf("ToolDefinitions() = %#v", toolDefinitions)
	}

	evaluation := astock.EvaluateValuationEvidence(astock.ValuationEvaluationInput{
		ExpectedEntityName: "同花顺",
		ExpectedStockCode:  "300033",
		EvidenceEntityName: "同花顺",
		EvidenceStockCode:  "sz300033",
		AdapterStatus:      astockcontracts.AdapterStatusOK,
		FailureCode:        astockcontracts.FailureCodeNone,
		AnswerReady:        true,
		RequestedFields:    []string{"price", "pe_ttm"},
		FieldValues:        map[string]string{"price": "100.00", "pe_ttm": "31.76"},
		AsOf:               "2026-05-15T15:00:00+08:00",
		SourceURL:          "https://qt.gtimg.cn/q=sz300033",
	})
	if !evaluation.Passed {
		t.Fatalf("EvaluateValuationEvidence() = %#v", evaluation)
	}
}
