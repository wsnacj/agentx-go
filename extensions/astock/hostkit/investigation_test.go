package hostkit

import (
	"context"
	"errors"
	"strings"
	"testing"

	astockcontracts "github.com/wsnacj/agentx-go/extensions/astock/contracts"
)

func TestBuildAStockInvestigationPayloadProjectsHandlerErrors(t *testing.T) {
	raw := "secret=sk-agentx path=/private/a-stock.json query=300033 provider_response=denied"
	payload, err := BuildAStockInvestigationPayload(context.Background(), InvestigationConfig{
		Handlers: InvestigationHandlers{
			Quote: func(context.Context, map[string]any) (astockcontracts.QuotePayload, error) {
				return astockcontracts.QuotePayload{}, errors.New(raw)
			},
		},
	}, map[string]any{
		"task_kind":  "quote_snapshot",
		"stock_code": "300033",
	})
	if err != nil {
		t.Fatalf("build payload: %v", err)
	}
	warnings := strings.Join(payload.Warnings, "\n")
	if strings.Contains(warnings, raw) || strings.Contains(warnings, "sk-agentx") || strings.Contains(warnings, "/private/a-stock.json") {
		t.Fatalf("raw handler error reached model-facing warnings: %q", warnings)
	}
	if !strings.Contains(warnings, "class=a_stock_investigation code=quote_lookup_failed identity=") {
		t.Fatalf("expected safe error projection, got %q", warnings)
	}
}

func TestIntentFromParamsKeepsStructuredFields(t *testing.T) {
	intent := IntentFromParams(map[string]any{
		"user_message":      "查同花顺行情",
		"task_kind":         "quote_snapshot",
		"entity_name":       "同花顺",
		"entity_mentions":   []any{"同花顺", "300033"},
		"stock_code":        "300033",
		"market":            "auto",
		"requested_fields":  []any{"price", "pe_ttm"},
		"requested_outputs": "brief, evidence_table",
		"freshness": map[string]any{
			"mode":                       "realtime",
			"relative_date_hint":         "今天",
			"require_latest_trading_day": true,
		},
	})
	if intent.UserMessage != "查同花顺行情" || intent.TaskKind != astockcontracts.TaskKindQuoteSnapshot || intent.Market != "" {
		t.Fatalf("unexpected intent: %#v", intent)
	}
	if len(intent.EntityMentions) != 2 || len(intent.RequestedFields) != 2 || len(intent.RequestedOutputs) != 2 {
		t.Fatalf("unexpected list fields: %#v", intent)
	}
	if intent.Freshness.Mode != astockcontracts.FreshnessModeRealtime || !intent.Freshness.RequireLatestTradingDay {
		t.Fatalf("unexpected freshness: %#v", intent.Freshness)
	}
}

func TestBuildAStockInvestigationPayloadRunsPlannedHandler(t *testing.T) {
	payload, err := BuildAStockInvestigationPayload(context.Background(), InvestigationConfig{
		Source: "test_host",
		Handlers: InvestigationHandlers{
			Quote: func(context.Context, map[string]any) (astockcontracts.QuotePayload, error) {
				return astockcontracts.QuotePayload{
					Readiness: astockcontracts.BuildReadiness(astockcontracts.AdapterStatusOK, astockcontracts.FailureCodeNone, true, nil, nil),
					Subject:   astockcontracts.Subject{StockCode: "300033", Verified: true},
					IdentityResolution: &astockcontracts.IdentityResolution{
						InputTerm:      "同花顺",
						Strategy:       "cninfo_top_search",
						SelectedReason: "cninfo_top_search_selected",
						SelectedCandidate: &astockcontracts.IdentityResolutionCandidate{
							Code:     "300033",
							Market:   astockcontracts.MarketSZ,
							Selected: true,
						},
					},
				}, nil
			},
		},
	}, map[string]any{
		"user_message": "查同花顺行情",
		"task_kind":    "quote_snapshot",
		"stock_code":   "300033",
	})
	if err != nil {
		t.Fatalf("build payload: %v", err)
	}
	if !payload.Readiness.AnswerReady || payload.Quote == nil || payload.Source != "test_host" {
		t.Fatalf("unexpected payload: %#v", payload)
	}
	if payload.IdentityResolution == nil ||
		payload.IdentityResolution.SelectedCandidate == nil ||
		payload.IdentityResolution.SelectedCandidate.Code != "300033" {
		t.Fatalf("expected projected identity resolution, got %#v", payload.IdentityResolution)
	}
}

func TestBuildAStockInvestigationPayloadRunsMultiEntityQuoteComparison(t *testing.T) {
	calls := []string{}
	payload, err := BuildAStockInvestigationPayload(context.Background(), InvestigationConfig{
		Handlers: InvestigationHandlers{
			Quote: func(_ context.Context, params map[string]any) (astockcontracts.QuotePayload, error) {
				entity := StringArg(params["entity_name"])
				calls = append(calls, entity)
				return astockcontracts.QuotePayload{
					Readiness: astockcontracts.BuildReadiness(astockcontracts.AdapterStatusOK, astockcontracts.FailureCodeNone, true, nil, nil),
					Subject:   astockcontracts.Subject{EntityName: entity, Verified: true},
					IdentityResolution: &astockcontracts.IdentityResolution{
						InputTerm:      entity,
						Strategy:       "cninfo_top_search",
						SelectedReason: "cninfo_top_search_selected",
						SelectedCandidate: &astockcontracts.IdentityResolutionCandidate{
							Name:     entity,
							Selected: true,
						},
					},
				}, nil
			},
		},
	}, map[string]any{
		"user_message":      "比较同花顺、东方雨虹估值",
		"task_kind":         "comparison",
		"entity_mentions":   []any{"同花顺", "东方雨虹"},
		"requested_fields":  []any{"price", "market_cap", "pe_ttm", "pb"},
		"requested_outputs": []any{"quote", "valuation"},
	})
	if err != nil {
		t.Fatalf("build payload: %v", err)
	}
	if !payload.Readiness.AnswerReady || len(payload.Quotes) != 2 || len(calls) != 2 {
		t.Fatalf("expected two quote payloads, got payload=%#v calls=%#v", payload, calls)
	}
	if payload.Quotes[0].Subject.EntityName != "同花顺" || payload.Quotes[1].Subject.EntityName != "东方雨虹" {
		t.Fatalf("unexpected quote subjects: %#v", payload.Quotes)
	}
	if payload.IdentityResolution != nil {
		t.Fatalf("multi-entity comparison should keep identity diagnostics on each quote, got top-level %#v", payload.IdentityResolution)
	}
	for _, quote := range payload.Quotes {
		if quote.IdentityResolution == nil {
			t.Fatalf("expected quote-level identity diagnostics, got %#v", payload.Quotes)
		}
	}
}

func TestBuildAStockInvestigationPayloadRestrictsFullInvestigationToRequestedOutputs(t *testing.T) {
	payload, err := BuildAStockInvestigationPayload(context.Background(), InvestigationConfig{
		Handlers: InvestigationHandlers{
			Quote: func(context.Context, map[string]any) (astockcontracts.QuotePayload, error) {
				return astockcontracts.QuotePayload{
					Readiness: astockcontracts.BuildReadiness(astockcontracts.AdapterStatusOK, astockcontracts.FailureCodeNone, true, nil, nil),
					Subject:   astockcontracts.Subject{StockCode: "300750", Verified: true},
				}, nil
			},
			Announcement: func(context.Context, map[string]any) (astockcontracts.AnnouncementPayload, error) {
				return astockcontracts.AnnouncementPayload{
					Readiness: astockcontracts.BuildReadiness(astockcontracts.AdapterStatusOK, astockcontracts.FailureCodeNone, true, nil, nil),
					Subject:   astockcontracts.Subject{StockCode: "300750", Verified: true},
				}, nil
			},
			Profile: func(context.Context, map[string]any) (astockcontracts.ProfilePayload, error) {
				t.Fatalf("profile handler should not run for requested quote+announcements only")
				return astockcontracts.ProfilePayload{}, nil
			},
			Research: func(context.Context, map[string]any) (astockcontracts.ResearchPayload, error) {
				t.Fatalf("research handler should not run for requested quote+announcements only")
				return astockcontracts.ResearchPayload{}, nil
			},
		},
	}, map[string]any{
		"user_message":      "宁德时代最近公告，顺便看当前价格",
		"task_kind":         "full_investigation",
		"entity_name":       "宁德时代",
		"requested_outputs": []any{"quote", "announcements"},
	})
	if err != nil {
		t.Fatalf("build payload: %v", err)
	}
	if !payload.Readiness.AnswerReady || payload.Quote == nil || payload.Announcement == nil || payload.Profile != nil || payload.Research != nil {
		t.Fatalf("unexpected restricted investigation payload: %#v", payload)
	}
}

func TestBuildAStockInvestigationPayloadReturnsUnsupportedSignalWhenMissing(t *testing.T) {
	payload, err := BuildAStockInvestigationPayload(context.Background(), InvestigationConfig{}, map[string]any{
		"user_message": "查题材",
		"task_kind":    "signal_lookup",
	})
	if err != nil {
		t.Fatalf("build payload: %v", err)
	}
	if payload.Signal == nil || payload.FailureCode != astockcontracts.FailureCodeUnsupported || payload.Readiness.Degraded {
		t.Fatalf("expected unsupported signal payload, got %#v", payload)
	}
}

func TestBuildAStockInvestigationPayloadAllowsPartialMultiSectionAnswer(t *testing.T) {
	payload, err := BuildAStockInvestigationPayload(context.Background(), InvestigationConfig{
		Handlers: InvestigationHandlers{
			Quote: func(context.Context, map[string]any) (astockcontracts.QuotePayload, error) {
				return astockcontracts.QuotePayload{
					Readiness: astockcontracts.BuildReadiness(astockcontracts.AdapterStatusOK, astockcontracts.FailureCodeNone, true, nil, nil),
					Subject:   astockcontracts.Subject{StockCode: "300033", Verified: true},
				}, nil
			},
			Profile: func(context.Context, map[string]any) (astockcontracts.ProfilePayload, error) {
				return astockcontracts.ProfilePayload{
					Readiness: astockcontracts.BuildReadiness(astockcontracts.AdapterStatusEvidenceIncomplete, astockcontracts.FailureCodeMissingFields, false, []string{"listing_date"}, nil),
					Subject:   astockcontracts.Subject{StockCode: "300033", Verified: true},
				}, nil
			},
			Research: func(context.Context, map[string]any) (astockcontracts.ResearchPayload, error) {
				return astockcontracts.ResearchPayload{
					Readiness: astockcontracts.BuildReadiness(astockcontracts.AdapterStatusOK, astockcontracts.FailureCodeNone, true, nil, nil),
					Subject:   astockcontracts.Subject{StockCode: "300033", Verified: true},
				}, nil
			},
			Announcement: func(context.Context, map[string]any) (astockcontracts.AnnouncementPayload, error) {
				return astockcontracts.AnnouncementPayload{
					Readiness: astockcontracts.BuildReadiness(astockcontracts.AdapterStatusOK, astockcontracts.FailureCodeNone, true, nil, nil),
					Subject:   astockcontracts.Subject{StockCode: "300033", Verified: true},
				}, nil
			},
		},
	}, map[string]any{
		"user_message": "综合看下同花顺",
		"task_kind":    "full_investigation",
		"stock_code":   "300033",
	})
	if err != nil {
		t.Fatalf("build payload: %v", err)
	}
	if !payload.Readiness.AnswerReady ||
		!payload.Readiness.Degraded ||
		payload.Readiness.FailureCode != astockcontracts.FailureCodeMissingFields ||
		len(payload.Readiness.MissingFields) != 1 ||
		payload.Readiness.MissingFields[0] != "step_2.listing_date" {
		t.Fatalf("expected partial multi-section answer readiness, got %#v", payload.Readiness)
	}
}
