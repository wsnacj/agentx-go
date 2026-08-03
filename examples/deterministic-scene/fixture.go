package main

import (
	astock "github.com/wsnacj/agentx-go/scenes/astock"
	astockcontracts "github.com/wsnacj/agentx-go/scenes/astock/contracts"
)

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
