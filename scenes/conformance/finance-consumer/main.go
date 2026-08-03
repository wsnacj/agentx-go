package main

import (
	"context"
	"fmt"

	finance "github.com/wsnacj/agentx-go/scenes/finance"
	financehostkit "github.com/wsnacj/agentx-go/scenes/finance/hostkit"
	financialmetrics "github.com/wsnacj/agentx-go/scenes/finance/metrics"
	globalstock "github.com/wsnacj/agentx-go/scenes/globalstock"
	globalcontracts "github.com/wsnacj/agentx-go/scenes/globalstock/contracts"
	globalhostkit "github.com/wsnacj/agentx-go/scenes/globalstock/hostkit"
)

func run(ctx context.Context) (string, error) {
	quoteCalls := 0
	stock, err := globalhostkit.BuildGlobalStockInvestigationPayload(ctx, globalhostkit.InvestigationConfig{
		Handlers: globalhostkit.InvestigationHandlers{
			Quote: func(context.Context, map[string]any) (globalcontracts.QuotePayload, error) {
				quoteCalls++
				return globalcontracts.QuotePayload{
					AdapterStatus: globalcontracts.AdapterStatusOK,
					Readiness:     globalcontracts.BuildReadiness(globalcontracts.AdapterStatusOK, globalcontracts.FailureCodeNone, true, nil, nil),
				}, nil
			},
		},
	}, map[string]any{"entity_name": "Fixture", "stock_code": "TEST", "market": "us"})
	if err != nil || quoteCalls != 1 || !stock.Readiness.AnswerReady {
		return "", fmt.Errorf("globalstock coordination failed: calls=%d err=%v", quoteCalls, err)
	}

	financeCalls := 0
	report, err := financehostkit.BuildFinanceReportLookupPayload(ctx, financehostkit.FinanceReportLookupConfig{
		Handlers: financehostkit.FinanceReportLookupHandlers{
			Candidates: func(context.Context, map[string]any) (finance.MetricsCandidatesPayload, error) {
				financeCalls++
				return finance.MetricsCandidatesPayload{AdapterStatus: "ok", PrimaryURL: "https://example.com/report.pdf", ResolvedCompany: "Fixture"}, nil
			},
			MetricsGuard: func(context.Context, map[string]any) (finance.MetricsToolPayload, error) {
				financeCalls++
				return finance.MetricsToolPayload{AdapterStatus: "ok", GuardStatus: "passed", RequestedFieldsReady: true, Evidence: finance.MetricsEvidence{CompanyName: "Fixture", OfficialSource: "https://example.com/report.pdf", ReportPeriod: "2025"}}, nil
			},
		},
	}, map[string]any{"entity_name": "Fixture", "requested_metrics": []any{"revenue"}})
	if err != nil || financeCalls != 2 || report.Tool != finance.ToolFinanceReportLookup {
		return "", fmt.Errorf("finance coordination failed: calls=%d err=%v", financeCalls, err)
	}

	return fmt.Sprintf("agentx-finance-ok:%s:%s:%d:%d", globalstock.PackID, financialmetrics.PackID, quoteCalls, financeCalls), nil
}

func main() {
	output, err := run(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Println(output)
}
