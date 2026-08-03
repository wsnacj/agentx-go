package hostkit

import (
	"context"
	"encoding/json"
	"testing"

	globalcontracts "github.com/wsnacj/agentx-go/scenes/globalstock/contracts"
)

func TestBuildGlobalStockInvestigationPayloadReturnsFinanceReportHandoff(t *testing.T) {
	payload, err := BuildGlobalStockInvestigationPayload(context.Background(), InvestigationConfig{
		Source:              "test_global_stock",
		SourcePolicyDefault: "public_hk_us_stock_sources",
		Handlers: InvestigationHandlers{
			Quote: func(context.Context, map[string]any) (globalcontracts.QuotePayload, error) {
				t.Fatal("quote handler should not be called for finance report handoff")
				return globalcontracts.QuotePayload{}, nil
			},
		},
	}, map[string]any{
		"user_message":      "帮我看下腾讯控股最新财报的营收和净利润",
		"task_kind":         string(globalcontracts.TaskKindFinanceReportHandoff),
		"entity_name":       "腾讯控股",
		"entity_mentions":   []any{"腾讯控股"},
		"stock_code":        "00700",
		"market":            "hk",
		"requested_fields":  []any{"revenue", "net_profit"},
		"requested_outputs": []any{"report_metrics"},
		"freshness": map[string]any{
			"mode":               "latest_published",
			"relative_date_hint": "最新",
		},
	})
	if err != nil {
		t.Fatalf("BuildGlobalStockInvestigationPayload: %v", err)
	}
	if payload.Handoff == nil {
		t.Fatalf("expected handoff payload: %#v", payload)
	}
	if payload.Handoff.TargetPackage != "agentx_finance" || payload.Handoff.TargetTool != "finance_report_lookup" {
		t.Fatalf("unexpected handoff target: %#v", payload.Handoff)
	}
	if payload.Readiness.AnswerReady || payload.Readiness.NextRepairHint != "call_finance_report_lookup" {
		t.Fatalf("unexpected readiness: %#v", payload.Readiness)
	}
	args := payload.Handoff.Arguments
	if args["task_kind"] != "latest_report_metrics" || args["report_kind"] != "auto" || args["entity_name"] != "腾讯控股" || args["stock_code"] != "00700" {
		t.Fatalf("unexpected finance args: %#v", args)
	}
	metrics, ok := args["requested_metrics"].([]string)
	if !ok || len(metrics) != 2 || metrics[0] != "revenue" || metrics[1] != "net_profit" {
		t.Fatalf("unexpected requested_metrics: %#v", args["requested_metrics"])
	}
	outputs, ok := args["requested_outputs"].([]string)
	if !ok || len(outputs) != 1 || outputs[0] != "metrics" {
		t.Fatalf("unexpected requested_outputs: %#v", args["requested_outputs"])
	}
	freshness, ok := args["freshness"].(map[string]any)
	if !ok || freshness["mode"] != "latest_published" || freshness["relative_date_hint"] != "最新" {
		t.Fatalf("unexpected freshness: %#v", args["freshness"])
	}
}

func TestBuildGlobalStockInvestigationPayloadReturnsFinanceTrendHandoffFromStructuredOutput(t *testing.T) {
	payload, err := BuildGlobalStockInvestigationPayload(context.Background(), InvestigationConfig{}, map[string]any{
		"user_message":      "Compare recent report metrics for Tesla",
		"entity_name":       "Tesla",
		"stock_code":        "TSLA",
		"market":            "us",
		"requested_fields":  []any{"revenue", "net_profit_growth"},
		"requested_outputs": []any{"report_metrics_trend"},
	})
	if err != nil {
		t.Fatalf("BuildGlobalStockInvestigationPayload: %v", err)
	}
	if payload.Handoff == nil {
		t.Fatalf("expected handoff payload: %#v", payload)
	}
	args := payload.Handoff.Arguments
	if args["task_kind"] != "report_metrics_trend" || args["period_scope"] != "recent_years" || args["ticker"] != "TSLA" {
		t.Fatalf("unexpected trend finance args: %#v", args)
	}
}

func TestFinanceReportHandoffRemainsTypedAcrossJSON(t *testing.T) {
	payload := globalcontracts.InvestigationPayload{
		Tool:          "global_stock_investigation",
		AdapterStatus: globalcontracts.AdapterStatusUnsupported,
		FailureCode:   globalcontracts.FailureCodeUnsupported,
		Readiness: globalcontracts.BuildReadiness(
			globalcontracts.AdapterStatusUnsupported,
			globalcontracts.FailureCodeUnsupported,
			false,
			[]string{"finance_report_lookup"},
			nil,
		),
		Evidence: globalcontracts.SourceEvidence{
			Provider: "agentx_global_stock",
			Source:   "agentx_global_stock_hostkit",
		},
		Handoff: &globalcontracts.ToolHandoff{
			TargetPackage: "agentx_finance",
			TargetTool:    "finance_report_lookup",
			Reason:        "financial_report_metrics_are_owned_by_agentx_finance",
		},
	}
	blob, _ := json.Marshal(payload)
	var decoded globalcontracts.InvestigationPayload
	if err := json.Unmarshal(blob, &decoded); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if decoded.Handoff == nil || decoded.Handoff.TargetTool != "finance_report_lookup" {
		t.Fatalf("typed handoff lost after JSON round trip: %#v", decoded.Handoff)
	}
}
