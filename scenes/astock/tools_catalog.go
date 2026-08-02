package astock

import (
	llm "github.com/wsnacj/agentx-go/components/llm"
	astockcontracts "github.com/wsnacj/agentx-go/scenes/astock/contracts"
)

// ToolDefinitions returns fresh, caller-owned LLM tool schemas. It does not
// register handlers or grant access to a provider.
func ToolDefinitions() []llm.Tool {
	return []llm.Tool{
		AStockInvestigationTool(),
		AStockQuoteLookupTool(),
		AStockResearchLookupTool(),
		AStockSignalLookupTool(),
		AStockAnnouncementLookupTool(),
		AStockProfileLookupTool(),
		AStockAnswerFormatTool(),
	}
}

func AStockInvestigationTool() llm.Tool {
	return functionTool(
		ToolAStockInvestigation,
		"High-level A-share investigation for market-data tasks that may combine quote, valuation, research, announcement-list, company profile, signals, screening, comparison, unsupported A-share identity checks, or investment-risk framing. It does not extract public financial-report metrics or report briefs; if the user asks for annual/quarterly report facts, revenue, net profit, growth, cash flow, or report brief, call finance_report_lookup first and use A-share tools only for the requested market/research/signal evidence. The model supplies structured intent; adapters verify identity, source, timestamp, market session, requested fields, and evidence readiness. For private, overseas, or non-A-share subjects in an A-share price/code request, use this tool or a_stock_quote_lookup so adapters can return a structured unsupported/identity_not_found boundary instead of guessing. An identity_not_found result proves only that no verified A-share identity was resolved; do not infer global listing, private-company, legal, or other-market status without separate evidence. For multi-entity valuation comparisons, provide entity_mentions and requested_fields; unsupported historical/peer valuation conclusions must be stated as unavailable unless a configured source provides that evidence.",
		map[string]any{
			"user_message": map[string]any{"type": "string", "description": "Original user request. Preserve company names, symbols, relative-date hints, and freshness requirements verbatim."},
			"task_kind":    taskKindSchema(),
			"entity_name":  map[string]any{"type": "string", "description": "Candidate A-share entity/company name inferred from the request; adapters must verify the subject."},
			"entity_mentions": stringArraySchema(
				"Original entity/company mentions from the request, including aliases or bilingual names when present.",
			),
			"stock_code":        stockCodeSchema(),
			"market":            marketSchema(),
			"requested_fields":  stringArraySchema("A-share data fields requested by the user, such as price, pe_ttm, research_rating, announcements, industry, or hot_reason."),
			"requested_outputs": stringArraySchema("Answer products requested by the user, such as brief, comparison, risk_summary, valuation_snapshot, or evidence_table."),
			"assessment":        assessmentSchema(),
			"freshness":         freshnessSchema(),
			"source_hint":       sourceHintSchema(),
			"source_policy":     sourcePolicySchema(),
		},
	)
}

func AStockQuoteLookupTool() llm.Tool {
	return functionTool(
		ToolAStockQuoteLookup,
		"A-share quote and valuation snapshot lookup. Host adapters verify stock code, market, quote timestamp, market session, PE/PB, market cap, turnover, and limit prices. Use this tool for A-share price/code requests even when the subject may be private, overseas, or unsupported; adapters return structured unsupported/identity_not_found boundaries instead of guessing. An identity_not_found result proves only that no verified A-share identity was resolved; do not infer global listing, private-company, legal, or other-market status without separate evidence. This is point-in-time valuation evidence only; do not claim historical percentile, cheap/expensive versus history, or peer-relative ranking unless another configured tool/source provides that evidence. Final answers based on this tool must include a short non-personalized investment-advice boundary.",
		map[string]any{
			"user_message": userMessageSchema("Original user request. Preserve company names, symbols, relative-date hints, and freshness requirements verbatim."),
			"entity_name":  entityNameSchema(),
			"stock_code":   stockCodeSchema(),
			"market":       marketSchema(),
			"quote_fields": stringEnumArraySchema("Quote or valuation fields requested by the user.", []string{
				"price", "change_pct", "turnover", "pe_ttm", "pe_static", "pb", "market_cap", "float_market_cap", "limit_up", "limit_down", "kline", "order_book", "transactions",
			}),
			"freshness":     freshnessSchema(),
			"source_policy": sourcePolicySchema(),
		},
	)
}

func AStockResearchLookupTool() llm.Tool {
	return functionTool(
		ToolAStockResearchLookup,
		"A-share research report lookup. Host adapters verify stock identity, report source, publication date, rating, forecast fields, and PDF metadata.",
		map[string]any{
			"user_message":      userMessageSchema("Original user request. Preserve company names, symbols, research fields, source constraints, and freshness hints verbatim."),
			"entity_name":       entityNameSchema(),
			"stock_code":        stockCodeSchema(),
			"market":            marketSchema(),
			"requested_fields":  stringEnumArraySchema("Research fields requested by the user.", []string{"rating", "rating_change", "eps_forecast", "profit_forecast", "target_price", "report_pdf", "institution"}),
			"period_scope":      periodScopeSchema(),
			"limit":             limitSchema(),
			"freshness":         freshnessSchema(),
			"source_policy":     sourcePolicySchema(),
			"require_pdf_links": map[string]any{"type": "boolean", "description": "Whether the caller explicitly needs research PDF links when available. Missing PDF links must be reported as incomplete rather than inferred."},
		},
	)
}

func AStockSignalLookupTool() llm.Tool {
	return functionTool(
		ToolAStockSignalLookup,
		"A-share signal lookup. The default livekit supports Tonghuashun hot-topic/reason-tag evidence for hot_reason, Eastmoney subject concept blocks for concept_blocks, Eastmoney stock fund-flow evidence for fund_flow, Eastmoney Dragon Tiger Board evidence for dragon_tiger_board/daily_dragon_tiger, Eastmoney lockup-expiry evidence for lockup_expiry, Eastmoney industry-board evidence for industry_comparison, and host-owned cache evidence for northbound_flow when NorthboundFlowCacheRoot or NorthboundFlowCacheFile is configured. Northbound flow must return an unsupported boundary when that host cache is not configured.",
		map[string]any{
			"user_message": userMessageSchema("Original user request. Preserve company names, symbols, signal types, trade dates, source constraints, and freshness hints verbatim."),
			"entity_name":  entityNameSchema(),
			"stock_code":   stockCodeSchema(),
			"market":       marketSchema(),
			"signal_types": stringEnumArraySchema("Signal types requested by the user.", []string{
				string(astockcontracts.SignalTypeHotReason),
				string(astockcontracts.SignalTypeConceptBlocks),
				string(astockcontracts.SignalTypeFundFlow),
				string(astockcontracts.SignalTypeNorthboundFlow),
				string(astockcontracts.SignalTypeDragonTigerBoard),
				string(astockcontracts.SignalTypeDailyBillboard),
				string(astockcontracts.SignalTypeLockupExpiry),
				string(astockcontracts.SignalTypeIndustryCompare),
			}),
			"trade_date":    tradeDateSchema(),
			"lookback_days": map[string]any{"type": "integer", "description": "Number of days to look back for signal evidence when the user asks for recent or rolling-window signals."},
			"forward_days":  map[string]any{"type": "integer", "description": "Number of future days to inspect for scheduled or upcoming signal events when supported by the host source."},
			"freshness":     freshnessSchema(),
			"source_policy": sourcePolicySchema(),
			"limit":         limitSchema(),
		},
	)
}

func AStockAnnouncementLookupTool() llm.Tool {
	return functionTool(
		ToolAStockAnnouncementLookup,
		"A-share announcement lookup. Host adapters verify issuer identity, announcement type, date, title, source URL, and full-text availability.",
		map[string]any{
			"user_message":       userMessageSchema("Original user request. Preserve company names, symbols, announcement categories, dates, source constraints, and freshness hints verbatim."),
			"entity_name":        entityNameSchema(),
			"stock_code":         stockCodeSchema(),
			"market":             marketSchema(),
			"announcement_types": stringArraySchema("Announcement categories or title keywords requested by the user."),
			"period_scope":       periodScopeSchema(),
			"limit":              limitSchema(),
			"freshness":          freshnessSchema(),
			"source_policy":      sourcePolicySchema(),
			"require_full_text":  map[string]any{"type": "boolean", "description": "Whether the caller needs full announcement text rather than metadata only. The host decides whether full text is available."},
		},
	)
}

func AStockProfileLookupTool() llm.Tool {
	return functionTool(
		ToolAStockProfileLookup,
		"A-share company profile lookup. Host adapters verify stock identity, industry, listing date, share capital, market cap, F10 profile, and basic quarterly snapshot fields.",
		map[string]any{
			"user_message": map[string]any{"type": "string", "description": "Original user request. Preserve company names, symbols, requested profile fields, and freshness hints verbatim."},
			"entity_name":  entityNameSchema(),
			"stock_code":   stockCodeSchema(),
			"market":       marketSchema(),
			"profile_fields": stringEnumArraySchema("Company profile fields requested by the user.", []string{
				"name", "industry", "listing_date", "share_capital", "market_cap", "f10_profile", "finance_snapshot", "eps", "roe", "net_profit", "income",
			}),
			"freshness":     freshnessSchema(),
			"source_policy": sourcePolicySchema(),
		},
	)
}

func AStockAnswerFormatTool() llm.Tool {
	return llm.Tool{
		Type: "function",
		Function: llm.Function{
			Name:        ToolAStockAnswerFormat,
			Description: "Format verified A-share tool payloads into a concise Chinese answer. Use only for evidence already returned by A-share lookup tools; do not fetch new facts or infer unsupported investment conclusions. Preserve source/time evidence and include a short non-personalized investment-advice boundary.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"user_message": map[string]any{"type": "string", "description": "Original user request for answer formatting and boundary preservation."},
					"answer_kind":  map[string]any{"type": "string", "description": "Expected answer format for the already verified A-share payload.", "enum": []string{"valuation", "research", "signal"}},
					"payload": map[string]any{
						"type":        "object",
						"description": "Exact payload returned by an A-share lookup tool. Preserve its top-level tool and readiness contract; do not reconstruct or add unsupported facts.",
						"properties": map[string]any{
							"tool": map[string]any{
								"type":        "string",
								"enum":        []string{ToolAStockQuoteLookup, ToolAStockResearchLookup, ToolAStockSignalLookup},
								"description": "Original A-share lookup tool identity from the verified payload.",
							},
							"readiness": map[string]any{
								"type":        "object",
								"description": "Original readiness contract from the verified payload. Do not omit it when relaying the payload.",
								"properties": map[string]any{
									"answer_ready": map[string]any{
										"type":        "boolean",
										"description": "Whether the source lookup marked the evidence ready for an answer.",
									},
								},
								"required": []string{"answer_ready"},
							},
						},
						"required": []string{"tool", "readiness"},
					},
				},
				"required": []string{"user_message", "payload"},
			},
		},
	}
}

func functionTool(name string, description string, properties map[string]any) llm.Tool {
	return llm.Tool{
		Type: "function",
		Function: llm.Function{
			Name:        name,
			Description: description,
			Parameters: map[string]any{
				"type":       "object",
				"properties": properties,
				"required":   []string{"user_message"},
			},
		},
	}
}

func taskKindSchema() map[string]any {
	return map[string]any{
		"type": "string",
		"enum": []string{
			string(astockcontracts.TaskKindQuoteSnapshot),
			string(astockcontracts.TaskKindValuationSnapshot),
			string(astockcontracts.TaskKindResearchLookup),
			string(astockcontracts.TaskKindSignalLookup),
			string(astockcontracts.TaskKindAnnouncement),
			string(astockcontracts.TaskKindProfileLookup),
			string(astockcontracts.TaskKindScreening),
			string(astockcontracts.TaskKindComparison),
			string(astockcontracts.TaskKindFullInvestigation),
		},
		"description": "Structured user intent. This is not a factual conclusion; adapters and guards verify selected sources.",
	}
}

func marketSchema() map[string]any {
	return map[string]any{
		"type":        "string",
		"enum":        []string{"auto", "sh", "sz", "bj"},
		"description": "Candidate market hint. Host adapters normalize and verify the final market.",
	}
}

func freshnessSchema() map[string]any {
	return map[string]any{
		"type":        "object",
		"description": "Freshness constraints from the user. Host adapters must verify source timestamps, trading-day status, and delayed/realtime availability.",
		"properties": map[string]any{
			"mode":                       map[string]any{"type": "string", "description": "Freshness mode such as latest_available, realtime, delayed_ok, explicit_trade_date, or recent_period."},
			"relative_date_hint":         map[string]any{"type": "string", "description": "Relative date phrase from the user, kept verbatim for freshness and trading-day guard."},
			"trade_date":                 tradeDateSchema(),
			"as_of":                      map[string]any{"type": "string", "description": "Timestamp or date at which the source evidence should be considered current."},
			"require_realtime":           map[string]any{"type": "boolean", "description": "Whether delayed market data is unacceptable for this request."},
			"require_latest_trading_day": map[string]any{"type": "boolean", "description": "Whether the response must use the latest trading day available for the verified market."},
		},
	}
}

func userMessageSchema(description string) map[string]any {
	return map[string]any{
		"type":        "string",
		"description": description,
	}
}

func entityNameSchema() map[string]any {
	return map[string]any{
		"type":        "string",
		"description": "Candidate A-share entity/company name inferred from the request; adapters must verify the subject.",
	}
}

func stockCodeSchema() map[string]any {
	return map[string]any{
		"type":        "string",
		"description": "A-share stock code or symbol hint from the user or upstream resolver. Candidate only; adapters verify the listed security.",
	}
}

func sourceHintSchema() map[string]any {
	return map[string]any{
		"type":        "string",
		"description": "User-requested source, provider, exchange, or source class hint. This is intent only; host adapters still choose and verify sources.",
	}
}

func sourcePolicySchema() map[string]any {
	return map[string]any{
		"type":        "string",
		"description": "Optional host/source policy hint such as official-only, delayed-data acceptable, or provider preference. This is not an authorization source.",
	}
}

func assessmentSchema() map[string]any {
	return map[string]any{
		"type":        "string",
		"enum":        []string{"none", "business_performance", "investment_risk"},
		"description": "Assessment intent requested by the user or upstream router. This is not a factual conclusion or investment advice.",
	}
}

func periodScopeSchema() map[string]any {
	return map[string]any{
		"type":        "string",
		"description": "Requested period or lookback scope, such as latest, recent_days, recent_months, explicit_date, latest_announcement, or latest_quarter.",
	}
}

func tradeDateSchema() map[string]any {
	return map[string]any{
		"type":        "string",
		"description": "Explicit A-share trade date requested by the user or host policy.",
	}
}

func limitSchema() map[string]any {
	return map[string]any{
		"type":        "integer",
		"description": "Maximum number of source records or evidence items requested. Host adapters may cap this to provider-safe limits.",
	}
}

func stringArraySchema(description string) map[string]any {
	return map[string]any{
		"type":        "array",
		"items":       map[string]any{"type": "string"},
		"description": description,
	}
}

func stringEnumArraySchema(description string, values []string) map[string]any {
	return map[string]any{
		"type":        "array",
		"items":       map[string]any{"type": "string", "enum": values},
		"description": description,
	}
}
