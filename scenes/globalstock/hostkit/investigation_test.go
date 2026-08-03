package hostkit

import (
	"context"
	"errors"
	"strings"
	"testing"

	globalcontracts "github.com/wsnacj/agentx-go/scenes/globalstock/contracts"
)

func TestBuildGlobalStockInvestigationPayloadProjectsHandlerErrors(t *testing.T) {
	raw := "secret=sk-agentx path=/private/global-stock.json query=AAPL provider_response=denied"
	payload, err := BuildGlobalStockInvestigationPayload(context.Background(), InvestigationConfig{
		Handlers: InvestigationHandlers{
			Quote: func(context.Context, map[string]any) (globalcontracts.QuotePayload, error) {
				return globalcontracts.QuotePayload{}, errors.New(raw)
			},
		},
	}, map[string]any{
		"task_kind":  "quote_snapshot",
		"stock_code": "AAPL",
		"market":     "us",
	})
	if err != nil {
		t.Fatalf("build payload: %v", err)
	}
	warnings := strings.Join(payload.Warnings, "\n")
	if strings.Contains(warnings, raw) || strings.Contains(warnings, "sk-agentx") || strings.Contains(warnings, "/private/global-stock.json") {
		t.Fatalf("raw handler error reached model-facing warnings: %q", warnings)
	}
	if !strings.Contains(warnings, "class=global_stock_investigation code=quote_lookup_failed identity=") {
		t.Fatalf("expected safe error projection, got %q", warnings)
	}
}

func TestParamsForEntityMentionKeepsTitleCaseNamesAsEntities(t *testing.T) {
	base := map[string]any{
		"entity_name": "placeholder",
		"market":      "auto",
		"stock_code":  "PLACEHOLDER",
	}
	for _, mention := range []string{"Apple", "Tesla", "Microsoft", "NVIDIA"} {
		got := paramsForEntityMention(base, mention)
		if got["entity_name"] != mention {
			t.Fatalf("mention %q should remain entity_name, got %#v", mention, got)
		}
		if _, ok := got["stock_code"]; ok {
			t.Fatalf("mention %q should not be treated as stock_code: %#v", mention, got)
		}
		if _, ok := got["market"]; ok {
			t.Fatalf("mention %q should not carry stale market: %#v", mention, got)
		}
	}
}

func TestParamsForEntityMentionNormalizesExplicitCodes(t *testing.T) {
	tests := []struct {
		mention string
		code    string
		market  string
	}{
		{mention: "AAPL", code: "AAPL", market: "us"},
		{mention: "TSLA", code: "TSLA", market: "us"},
		{mention: "00700.HK", code: "00700", market: "hk"},
		{mention: "hk:9988", code: "09988", market: "hk"},
	}
	for _, tt := range tests {
		got := paramsForEntityMention(map[string]any{"entity_name": "placeholder"}, tt.mention)
		if got["stock_code"] != tt.code || got["market"] != tt.market {
			t.Fatalf("mention %q normalized to %#v, want code=%s market=%s", tt.mention, got, tt.code, tt.market)
		}
		if _, ok := got["entity_name"]; ok {
			t.Fatalf("mention %q should remove entity_name when explicit code is present: %#v", tt.mention, got)
		}
	}
}

func TestParamsForEntityMentionSplitsMarketQualifiedNames(t *testing.T) {
	tests := []struct {
		mention string
		entity  string
		market  string
	}{
		{mention: "阿里巴巴港股", entity: "阿里巴巴", market: "hk"},
		{mention: "腾讯音乐美股", entity: "腾讯音乐", market: "us"},
		{mention: "小米集团 HK", entity: "小米集团", market: "hk"},
		{mention: "NASDAQ NVIDIA", entity: "NVIDIA", market: "us"},
	}
	for _, tt := range tests {
		got := paramsForEntityMention(map[string]any{
			"entity_name": "placeholder",
			"market":      "auto",
			"stock_code":  "PLACEHOLDER",
		}, tt.mention)
		if got["entity_name"] != tt.entity || got["market"] != tt.market {
			t.Fatalf("mention %q split to %#v, want entity=%q market=%q", tt.mention, got, tt.entity, tt.market)
		}
		if _, ok := got["stock_code"]; ok {
			t.Fatalf("mention %q should not carry stale stock_code: %#v", tt.mention, got)
		}
	}
}

func TestBuildGlobalStockInvestigationPayloadTreatsExplicitStockCodeAsSingleSubject(t *testing.T) {
	calls := []map[string]any{}
	payload, err := BuildGlobalStockInvestigationPayload(context.Background(), InvestigationConfig{
		Handlers: InvestigationHandlers{
			Quote: func(_ context.Context, params map[string]any) (globalcontracts.QuotePayload, error) {
				calls = append(calls, params)
				return globalcontracts.QuotePayload{
					AdapterStatus: globalcontracts.AdapterStatusOK,
					Readiness:     globalcontracts.BuildReadiness(globalcontracts.AdapterStatusOK, globalcontracts.FailureCodeNone, true, nil, nil),
					Subject: globalcontracts.Subject{
						EntityName: "Alibaba Group Holding Ltd",
						StockCode:  "BABA",
						Market:     globalcontracts.MarketUS,
					},
				}, nil
			},
		},
	}, map[string]any{
		"user_message":      "看下阿里巴巴行情",
		"task_kind":         "full_investigation",
		"entity_name":       "Alibaba Group Holding Ltd",
		"entity_mentions":   []any{"阿里", "阿里巴巴", "Alibaba"},
		"stock_code":        "BABA",
		"market":            "us",
		"requested_outputs": []any{"valuation_snapshot"},
		"requested_fields":  []any{"price", "pe_ttm", "market_cap"},
	})
	if err != nil {
		t.Fatalf("BuildGlobalStockInvestigationPayload: %v", err)
	}
	if len(calls) != 1 {
		t.Fatalf("expected explicit stock_code to avoid multi-mention quote fanout, got calls=%#v", calls)
	}
	if calls[0]["stock_code"] != "BABA" || calls[0]["market"] != "us" {
		t.Fatalf("expected quote call to preserve explicit security, got %#v", calls[0])
	}
	if payload.Quote == nil || len(payload.Quotes) != 0 || !payload.Readiness.AnswerReady {
		t.Fatalf("expected single quote payload, got %#v", payload)
	}
}

func TestBuildGlobalStockInvestigationPayloadDeduplicatesResolvedAliases(t *testing.T) {
	calls := 0
	payload, err := BuildGlobalStockInvestigationPayload(context.Background(), InvestigationConfig{
		Handlers: InvestigationHandlers{
			Quote: func(_ context.Context, params map[string]any) (globalcontracts.QuotePayload, error) {
				calls++
				term := strings.ToUpper(strings.TrimSpace(firstNonEmptyString(params["stock_code"])))
				if term == "" {
					term = strings.ToUpper(strings.TrimSpace(firstNonEmptyString(params["entity_name"])))
				}
				code := map[string]string{
					"CISCO":  "CSCO",
					"CSCO":   "CSCO",
					"IBM":    "IBM",
					"ORACLE": "ORCL",
					"ORCL":   "ORCL",
				}[term]
				return globalcontracts.QuotePayload{
					AdapterStatus: globalcontracts.AdapterStatusOK,
					Readiness:     globalcontracts.BuildReadiness(globalcontracts.AdapterStatusOK, globalcontracts.FailureCodeNone, true, nil, nil),
					Subject: globalcontracts.Subject{
						EntityName: term,
						StockCode:  code,
						Market:     globalcontracts.MarketUS,
						Verified:   true,
					},
				}, nil
			},
		},
	}, map[string]any{
		"task_kind":         "comparison",
		"entity_mentions":   []any{"Cisco", "CSCO", "IBM", "Oracle", "ORCL"},
		"requested_outputs": []any{"valuation_snapshot"},
		"requested_fields":  []any{"price", "pe_ttm", "pb", "market_cap"},
	})
	if err != nil {
		t.Fatalf("BuildGlobalStockInvestigationPayload: %v", err)
	}
	if calls != 3 {
		t.Fatalf("quote calls = %d, want one per resolved security", calls)
	}
	if len(payload.Quotes) != 3 {
		t.Fatalf("quotes = %#v, want three unique securities", payload.Quotes)
	}
	gotCodes := []string{payload.Quotes[0].Subject.StockCode, payload.Quotes[1].Subject.StockCode, payload.Quotes[2].Subject.StockCode}
	if strings.Join(gotCodes, ",") != "CSCO,IBM,ORCL" {
		t.Fatalf("quote codes = %v, want stable first-seen order", gotCodes)
	}
	if !payload.Readiness.AnswerReady || payload.AdapterStatus != globalcontracts.AdapterStatusOK {
		t.Fatalf("unexpected aggregate readiness: %#v", payload.Readiness)
	}
}

func TestBuildGlobalStockInvestigationPayloadDelegatesComparisonAnswerContract(t *testing.T) {
	quotes := map[string]struct {
		name      string
		code      string
		marketCap string
	}{
		"Adobe":      {name: "Adobe Inc.", code: "ADBE", marketCap: "943.06875"},
		"Salesforce": {name: "Salesforce, Inc.", code: "CRM", marketCap: "1398.60630"},
	}
	payload, err := BuildGlobalStockInvestigationPayload(context.Background(), InvestigationConfig{
		AnswerContract: func(payload globalcontracts.InvestigationPayload) *globalcontracts.InvestigationAnswerContract {
			return &globalcontracts.InvestigationAnswerContract{
				FinalAnswerRecommended:     payload.Readiness.AnswerReady,
				Reason:                     "host_owned_comparison_draft",
				NumericConsistencyRequired: true,
				FinalAnswerDraft:           "host-owned",
			}
		},
		Handlers: InvestigationHandlers{
			Quote: func(_ context.Context, params map[string]any) (globalcontracts.QuotePayload, error) {
				entry := quotes[firstNonEmptyString(params["entity_name"])]
				return globalcontracts.QuotePayload{
					AdapterStatus: globalcontracts.AdapterStatusOK,
					Readiness:     globalcontracts.BuildReadiness(globalcontracts.AdapterStatusOK, globalcontracts.FailureCodeNone, true, nil, nil),
					Subject: globalcontracts.Subject{
						EntityName: entry.name,
						StockCode:  entry.code,
						Market:     globalcontracts.MarketUS,
						Exchange:   "NASDAQ",
						Currency:   "USD",
						Verified:   true,
					},
					Quote: globalcontracts.QuoteSnapshot{
						Price:     globalcontracts.MetricValue{Value: "100.00", Currency: "USD"},
						MarketCap: globalcontracts.MetricValue{Value: entry.marketCap, Unit: globalcontracts.MetricUnitHundredMillion, Currency: "USD"},
					},
				}, nil
			},
		},
	}, map[string]any{
		"task_kind":        "comparison",
		"entity_mentions":  []any{"Adobe", "Salesforce"},
		"requested_fields": []any{"price", "market_cap"},
	})
	if err != nil {
		t.Fatalf("BuildGlobalStockInvestigationPayload: %v", err)
	}
	contract := payload.AnswerContract
	if contract == nil || !contract.FinalAnswerRecommended || !contract.NumericConsistencyRequired {
		t.Fatalf("missing deterministic answer contract: %#v", contract)
	}
	if contract.Reason != "host_owned_comparison_draft" || contract.FinalAnswerDraft != "host-owned" {
		t.Fatalf("answer contract did not come from host callback: %#v", contract)
	}
}

func TestBuildGlobalStockInvestigationPayloadRoutesSignalLookup(t *testing.T) {
	called := false
	payload, err := BuildGlobalStockInvestigationPayload(context.Background(), InvestigationConfig{
		Handlers: InvestigationHandlers{
			Signal: func(_ context.Context, params map[string]any) (globalcontracts.SignalPayload, error) {
				called = true
				if params["entity_name"] != "Apple" {
					t.Fatalf("unexpected signal params: %#v", params)
				}
				return globalcontracts.SignalPayload{
					AdapterStatus: globalcontracts.AdapterStatusOK,
					Readiness:     globalcontracts.BuildReadiness(globalcontracts.AdapterStatusOK, globalcontracts.FailureCodeNone, true, nil, nil),
					Subject: globalcontracts.Subject{
						EntityName: "Apple",
						StockCode:  "AAPL",
						Market:     globalcontracts.MarketUS,
					},
					IdentityResolution: &globalcontracts.IdentityResolution{
						InputTerm:      "Apple",
						Strategy:       "eastmoney_suggest",
						SelectedReason: "candidate_score",
						SelectedCandidate: &globalcontracts.IdentityResolutionCandidate{
							Code:     "AAPL",
							Market:   globalcontracts.MarketUS,
							Selected: true,
						},
					},
					Signals: []globalcontracts.SignalEvent{{Type: globalcontracts.SignalTypeUSForm4, Form: "4"}},
				}, nil
			},
		},
	}, map[string]any{
		"user_message": "Apple 最近 insider Form 4",
		"task_kind":    "signal_lookup",
		"entity_name":  "Apple",
		"market":       "us",
		"signal_types": []any{"us_form_4"},
	})
	if err != nil {
		t.Fatalf("BuildGlobalStockInvestigationPayload: %v", err)
	}
	if !called || payload.Signal == nil || payload.Signal.Subject.StockCode != "AAPL" || !payload.Readiness.AnswerReady {
		t.Fatalf("expected signal lookup payload, got called=%v payload=%#v", called, payload)
	}
	if payload.IdentityResolution == nil ||
		payload.IdentityResolution.SelectedCandidate == nil ||
		payload.IdentityResolution.SelectedCandidate.Code != "AAPL" {
		t.Fatalf("expected investigation to project identity resolution diagnostics, got %#v", payload.IdentityResolution)
	}
}
