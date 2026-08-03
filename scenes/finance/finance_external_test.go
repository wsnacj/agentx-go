package finance_test

import (
	"context"
	"testing"

	finance "github.com/wsnacj/agentx-go/scenes/finance"
	financehostkit "github.com/wsnacj/agentx-go/scenes/finance/hostkit"
)

func TestExternalConsumerCoordinatesFinancialReportExactlyOnce(t *testing.T) {
	candidateCalls := 0
	extractCalls := 0
	guardCalls := 0
	payload, err := financehostkit.BuildFinanceReportLookupPayload(context.Background(), financehostkit.FinanceReportLookupConfig{
		Handlers: financehostkit.FinanceReportLookupHandlers{
			Candidates: func(context.Context, map[string]any) (finance.MetricsCandidatesPayload, error) {
				candidateCalls++
				return finance.MetricsCandidatesPayload{AdapterStatus: "ok", PrimaryURL: "https://example.com/report.pdf", ResolvedCompany: "Example"}, nil
			},
			MetricsExtract: func(context.Context, map[string]any) (finance.MetricsToolPayload, error) {
				extractCalls++
				return finance.MetricsToolPayload{AdapterStatus: "ok", Evidence: finance.MetricsEvidence{CompanyName: "Example", OfficialSource: "https://example.com/report.pdf", ReportPeriod: "2025"}}, nil
			},
			MetricsGuard: func(context.Context, map[string]any) (finance.MetricsToolPayload, error) {
				guardCalls++
				return finance.MetricsToolPayload{AdapterStatus: "ok", GuardStatus: "passed", RequestedFieldsReady: true, Evidence: finance.MetricsEvidence{CompanyName: "Example", OfficialSource: "https://example.com/report.pdf", ReportPeriod: "2025"}}, nil
			},
		},
	}, map[string]any{"entity_name": "Example", "requested_metrics": []any{"revenue"}})
	if err != nil || candidateCalls != 1 || extractCalls != 1 || guardCalls != 1 || payload.Tool != finance.ToolFinanceReportLookup {
		t.Fatalf("payload=%#v calls=%d/%d/%d err=%v", payload, candidateCalls, extractCalls, guardCalls, err)
	}
}
