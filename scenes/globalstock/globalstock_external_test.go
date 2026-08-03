package globalstock_test

import (
	"context"
	"testing"

	"github.com/wsnacj/agentx-go/scenes/globalstock"
	globalcontracts "github.com/wsnacj/agentx-go/scenes/globalstock/contracts"
	globalhostkit "github.com/wsnacj/agentx-go/scenes/globalstock/hostkit"
)

func TestExternalConsumerCoordinatesQuoteExactlyOnce(t *testing.T) {
	calls := 0
	payload, err := globalhostkit.BuildGlobalStockInvestigationPayload(context.Background(), globalhostkit.InvestigationConfig{
		Handlers: globalhostkit.InvestigationHandlers{
			Quote: func(context.Context, map[string]any) (globalcontracts.QuotePayload, error) {
				calls++
				return globalcontracts.QuotePayload{
					AdapterStatus: globalcontracts.AdapterStatusOK,
					Readiness:     globalcontracts.BuildReadiness(globalcontracts.AdapterStatusOK, globalcontracts.FailureCodeNone, true, nil, nil),
				}, nil
			},
		},
	}, map[string]any{"entity_name": "Example", "market": "us"})
	if err != nil || calls != 1 || payload.Tool != globalstock.ToolGlobalStockInvestigation || !payload.Readiness.AnswerReady {
		t.Fatalf("payload=%#v calls=%d err=%v", payload, calls, err)
	}
}
