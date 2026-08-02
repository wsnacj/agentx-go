package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"strings"

	astock "github.com/wsnacj/agentx-go/scenes/astock"
	astockcontracts "github.com/wsnacj/agentx-go/scenes/astock/contracts"
	astockhostkit "github.com/wsnacj/agentx-go/scenes/astock/hostkit"
	pack "github.com/wsnacj/agentx-go/extensions/pack"
	workflow "github.com/wsnacj/agentx-go/runtime/workflow"
)

type validator struct{}

func (validator) ValidateSpec(spec workflow.Spec) error {
	if strings.TrimSpace(spec.ID) == "" {
		return fmt.Errorf("workflow id is required")
	}
	return nil
}

type lowerer struct{}

func (lowerer) LowerToolArguments(node workflow.NodeSpec) (string, error) {
	arguments, _ := node.Config["args"].(map[string]any)
	encoded, err := json.Marshal(arguments)
	return string(encoded), err
}

func run() (string, error) {
	manifest := astock.Manifest()
	if _, err := fs.ReadFile(astock.ExtensionFS(), "skills/a-stock-data/SKILL.md"); err != nil {
		return "", fmt.Errorf("read embedded A-stock skill: %w", err)
	}

	coordinator, err := pack.NewCoordinator(validator{}, lowerer{})
	if err != nil {
		return "", err
	}
	registry, err := pack.NewMemoryRegistry(coordinator)
	if err != nil {
		return "", err
	}
	if err := astock.RegisterPacks(registry); err != nil {
		return "", err
	}
	selection, matched := pack.SelectBinding(registry, "请查询平安银行 A股估值和行情快照", pack.SelectOptions{})
	if !matched {
		return "", fmt.Errorf("A-stock pack selection did not match")
	}
	binding, ok, err := coordinator.ResolveBinding(
		registry,
		selection.Selected.PackID,
		selection.Selected.CaseType,
		selection.Selected.WorkflowID,
	)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", fmt.Errorf("A-stock binding is unavailable")
	}

	quote := fixtureQuote()
	investigation, err := astockhostkit.BuildAStockInvestigationPayload(
		context.Background(),
		astockhostkit.InvestigationConfig{
			Source: "fixture",
			Handlers: astockhostkit.InvestigationHandlers{
				Quote: func(context.Context, map[string]any) (astockcontracts.QuotePayload, error) {
					return quote, nil
				},
			},
		},
		map[string]any{
			"user_message":      "请查询平安银行 A股估值和行情快照",
			"task_kind":         string(astockcontracts.TaskKindValuationSnapshot),
			"entity_name":       "平安银行",
			"stock_code":        "000001",
			"market":            "sz",
			"requested_fields":  []string{"price", "pe_ttm"},
			"requested_outputs": []string{"valuation_snapshot"},
		},
	)
	if err != nil {
		return "", err
	}
	if !investigation.Readiness.AnswerReady || investigation.Quote == nil {
		return "", fmt.Errorf("fixture investigation is not answer ready: %#v", investigation.Readiness)
	}

	evaluation := astock.EvaluateValuationEvidence(astock.ValuationEvaluationInput{
		ExpectedEntityName: "平安银行",
		ExpectedStockCode:  "000001",
		EvidenceEntityName: quote.Subject.EntityName,
		EvidenceStockCode:  quote.Subject.StockCode,
		AdapterStatus:      quote.AdapterStatus,
		FailureCode:        quote.FailureCode,
		AnswerReady:        quote.Readiness.AnswerReady,
		RequestedFields:    []string{"price", "pe_ttm"},
		FieldValues:        map[string]string{"price": quote.Quote.Price.Value, "pe_ttm": quote.Quote.PETTM.Value},
		AsOf:               quote.Evidence.AsOf,
		SourceURL:          quote.Evidence.SourceURL,
	})
	if !evaluation.Passed {
		return "", fmt.Errorf("fixture valuation evidence did not pass: %#v", evaluation)
	}

	return fmt.Sprintf(
		"agentx-astock-ok:%s:%s:%s:%d:%t:%t",
		manifest.ID,
		binding.PackID,
		binding.WorkflowID,
		len(astock.ToolDefinitions()),
		investigation.Readiness.AnswerReady,
		evaluation.Passed,
	), nil
}

func fixtureQuote() astockcontracts.QuotePayload {
	return astockcontracts.QuotePayload{
		Tool:          astock.ToolAStockQuoteLookup,
		Source:        "fixture",
		AdapterID:     "fixture_quote",
		AdapterStatus: astockcontracts.AdapterStatusOK,
		FailureCode:   astockcontracts.FailureCodeNone,
		Subject: astockcontracts.Subject{
			EntityName: "平安银行",
			StockCode:  "sz000001",
			Market:     astockcontracts.MarketSZ,
			Verified:   true,
		},
		Evidence: astockcontracts.SourceEvidence{
			Source:    "fixture",
			SourceURL: "https://example.invalid/quote/sz000001",
			AsOf:      "2026-08-01T15:00:00+08:00",
		},
		Readiness: astockcontracts.BuildReadiness(
			astockcontracts.AdapterStatusOK,
			astockcontracts.FailureCodeNone,
			true,
			nil,
			nil,
		),
		Quote: astockcontracts.QuoteSnapshot{
			Price: astockcontracts.MetricValue{Field: "price", Value: "11.38", Currency: "CNY"},
			PETTM: astockcontracts.MetricValue{Field: "pe_ttm", Value: "5.72"},
		},
	}
}

func main() {
	result, err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(result)
}
