package packvaluation

import (
	astockcontracts "github.com/wsnacj/agentx-go/scenes/astock/contracts"
	agentxpack "github.com/wsnacj/agentx-go/extensions/pack"
	agentxexecution "github.com/wsnacj/agentx-go/runtime/executionpolicy"
	agentxworkflow "github.com/wsnacj/agentx-go/runtime/workflow"
)

const (
	PackID            = "a-stock-valuation-pack"
	CaseTypeQuote     = "a_stock.quote_snapshot"
	CaseTypeValuation = "a_stock.valuation_snapshot"
	DefaultWorkflow   = "a_stock_valuation_lookup_v1"
)

func valuationCaseSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"user_message": map[string]any{"type": "string"},
			"entity": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{"type": "string"},
					"identifiers": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"stock_code": map[string]any{"type": "string"},
							"market":     map[string]any{"type": "string"},
						},
					},
				},
				"required": []string{"name"},
			},
			"quote_fields": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "string",
					"enum": []string{
						"price",
						"change_pct",
						"turnover",
						"pe_ttm",
						"pe_static",
						"pb",
						"market_cap",
						"float_market_cap",
						"limit_up",
						"limit_down",
						"kline",
						"order_book",
						"transactions",
					},
				},
			},
			"requested_outputs": map[string]any{
				"type":  "array",
				"items": map[string]any{"type": "string"},
			},
			"assessment": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"kind":                         map[string]any{"type": "string"},
					"scope":                        map[string]any{"type": "string"},
					"requires_investment_boundary": map[string]any{"type": "boolean"},
				},
			},
			"freshness": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"mode":                       map[string]any{"type": "string"},
					"relative_date_hint":         map[string]any{"type": "string"},
					"require_realtime":           map[string]any{"type": "boolean"},
					"require_latest_trading_day": map[string]any{"type": "boolean"},
				},
			},
			"source_policy":  map[string]any{"type": "string"},
			"stop_condition": map[string]any{"type": "string"},
		},
		"required": []string{"user_message", "entity", "quote_fields", "requested_outputs"},
	}
}

func Definition() agentxpack.Definition {
	return agentxpack.Definition{
		Manifest: agentxpack.Manifest{
			ID:      PackID,
			Version: "0.1.0",
			Domain:  "a_stock_valuation",
			RouteHints: []string{
				"A股行情估值",
				"A-share valuation snapshot",
			},
			SupportedCaseTypes: []string{CaseTypeQuote, CaseTypeValuation},
			DefaultWorkflow:    DefaultWorkflow,
			PolicyProfiles:     []string{"a_stock_valuation_readonly"},
			ArtifactTypes:      []string{"a_stock.valuation.evidence"},
			Evaluators:         []string{"a_stock_valuation_evidence_guard"},
			EvalSuites:         []string{"a_stock_valuation_success_suite"},
		},
		CaseSchemas: []agentxpack.CaseSchema{
			{
				CaseType:    CaseTypeQuote,
				Description: "查询 A 股公开行情快照，包括价格、涨跌幅、成交、换手率和涨跌停等点时证据。",
				RouteHints:  []string{"A股行情", "quote snapshot"},
				Schema:      valuationCaseSchema(),
			},
			{
				CaseType:    CaseTypeValuation,
				Description: "查询 A 股点时估值快照，包括 PE、PB、市值、换手率等来源化证据；不自动宣称历史分位或投资结论。",
				RouteHints:  []string{"A股估值", "valuation snapshot"},
				Schema:      valuationCaseSchema(),
			},
		},
		Workflows: []agentxworkflow.Spec{
			{
				ID:              DefaultWorkflow,
				Title:           "A-Stock Valuation Lookup",
				Description:     "使用模型填充的结构化 task frame 查询 A 股行情/估值快照；adapter 验证主体、时点和字段。",
				Version:         "v1",
				Pack:            PackID,
				CaseTypes:       []string{CaseTypeQuote, CaseTypeValuation},
				RouteHints:      []string{"A股估值", "行情快照"},
				PlanningMode:    agentxworkflow.PlanningBounded,
				EntryNode:       "lookup_quote",
				DefaultContract: "a_stock_valuation_readonly",
				StateSchema: []agentxworkflow.StateSlotSpec{
					{Name: "valuation.answer_ready", Type: "boolean", Required: true},
					{Name: "valuation.adapter_status", Type: "string", Required: true},
					{Name: "valuation.failure_code", Type: "string"},
					{Name: "valuation.degraded", Type: "boolean"},
					{Name: "valuation.degrade_reason", Type: "string"},
					{Name: "valuation.subject_name", Type: "string", Required: true},
					{Name: "valuation.subject_stock_code", Type: "string", Required: true},
					{Name: "valuation.market", Type: "string"},
					{Name: "valuation.as_of", Type: "string"},
					{Name: "valuation.source_url", Type: "string"},
					{Name: "valuation.price", Type: "string"},
					{Name: "valuation.change_percent", Type: "string"},
					{Name: "valuation.turnover_percent", Type: "string"},
					{Name: "valuation.pe_ttm", Type: "string"},
					{Name: "valuation.pe_static", Type: "string"},
					{Name: "valuation.pb", Type: "string"},
					{Name: "valuation.market_cap", Type: "string"},
					{Name: "valuation.float_market_cap", Type: "string"},
					{Name: "valuation.limit_up", Type: "string"},
					{Name: "valuation.limit_down", Type: "string"},
					{Name: "valuation.summary", Type: "string", Required: true},
					{Name: "valuation.passed", Type: "boolean", Required: true},
					{Name: "valuation.subject_correct", Type: "boolean"},
					{Name: "valuation.freshness_accepted", Type: "boolean"},
					{Name: "valuation.fields_ready", Type: "boolean"},
					{Name: "valuation.source_accepted", Type: "boolean"},
					{Name: "valuation.advice_boundary_respected", Type: "boolean"},
					{Name: "valuation.failure_reason", Type: "string"},
				},
				EvaluatorSchema: []agentxworkflow.EvaluatorRef{
					{Name: "a_stock_valuation_evidence_guard"},
				},
				Nodes: []agentxworkflow.NodeSpec{
					{
						ID:          "lookup_quote",
						Kind:        agentxworkflow.NodeTool,
						Title:       "Lookup A-stock quote and valuation",
						Description: "查询主体的公开行情/估值快照，并将 readiness/source/freshness 投影到 pack state。",
						Inputs: []agentxworkflow.BindingSpec{
							{From: "case.input.user_message", To: "args.user_message"},
							{From: "case.input.entity.name", To: "args.entity_name"},
							{From: "case.input.quote_fields", To: "args.quote_fields"},
						},
						Outputs: []agentxworkflow.BindingSpec{
							{From: "result.readiness.answer_ready", To: "state.valuation.answer_ready"},
							{From: "result.readiness.answer_ready", To: "state.valuation.passed"},
							{From: "result.adapter_status", To: "state.valuation.adapter_status"},
							{From: "result.subject.entity_name", To: "state.valuation.subject_name"},
							{From: "result.subject.stock_code", To: "state.valuation.subject_stock_code"},
							{From: "result.subject.market", To: "state.valuation.market"},
							{From: "result.freshness.as_of", To: "state.valuation.as_of"},
							{From: "result.evidence.source_url", To: "state.valuation.source_url"},
							{From: "result.quote.price.value", To: "state.valuation.price"},
							{From: "result.quote.change_percent.value", To: "state.valuation.change_percent"},
							{From: "result.quote.turnover_percent.value", To: "state.valuation.turnover_percent"},
							{From: "result.quote.pe_ttm.value", To: "state.valuation.pe_ttm"},
							{From: "result.quote.pe_static.value", To: "state.valuation.pe_static"},
							{From: "result.quote.pb.value", To: "state.valuation.pb"},
							{From: "result.quote.market_cap.value", To: "state.valuation.market_cap"},
							{From: "result.quote.float_market_cap.value", To: "state.valuation.float_market_cap"},
							{From: "result.adapter_status", To: "state.valuation.summary"},
						},
						Config: map[string]any{
							"tool_name": "lookup_a_stock_valuation",
						},
					},
					{
						ID:          "format_answer",
						Kind:        agentxworkflow.NodeTool,
						Title:       "Format A-stock valuation answer",
						Description: "把已验证的行情/估值 payload 转为用户可读的中文回答；不新增事实、不生成投资建议。",
						Inputs: []agentxworkflow.BindingSpec{
							{From: "case.input.user_message", To: "args.user_message"},
							{From: "node.lookup_quote.result", To: "args.payload"},
						},
						Config: map[string]any{
							"tool_name":   "format_a_stock_answer",
							"answer_kind": "valuation",
						},
					},
				},
				Edges: []agentxworkflow.EdgeSpec{
					{From: "lookup_quote", To: "format_answer", On: "success"},
				},
			},
		},
		Tools: []agentxpack.SemanticTool{
			{
				Name:        "lookup_a_stock_valuation",
				Description: "查询 A 股行情和点时估值快照。host/live adapter 负责主体验证、时点、来源和降级原因。",
				RuntimeTool: astockcontracts.ToolAStockQuoteLookup,
				Tags:        []string{"a-stock", "quote", "valuation", "public-source", "adapter"},
			},
			{
				Name:        "format_a_stock_answer",
				Description: "将已验证的 A 股工具 payload 格式化为中文最终回答，不新增事实。",
				RuntimeTool: astockcontracts.ToolAStockAnswerFormat,
				Tags:        []string{"a-stock", "answer-format", "public-source"},
			},
		},
		Evaluators: []agentxpack.Evaluator{
			{
				Name:        "a_stock_valuation_evidence_guard",
				Description: "检查 A 股行情/估值快照的主体、时点、字段完整度、来源和投资建议边界是否满足答复要求。",
				OutputSchema: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"passed":                    map[string]any{"type": "boolean"},
						"subject_correct":           map[string]any{"type": "boolean"},
						"freshness_accepted":        map[string]any{"type": "boolean"},
						"fields_ready":              map[string]any{"type": "boolean"},
						"source_accepted":           map[string]any{"type": "boolean"},
						"advice_boundary_respected": map[string]any{"type": "boolean"},
						"missing_requested_fields":  map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
						"review_required_fields":    map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
						"failure_reason":            map[string]any{"type": "string"},
						"source_url":                map[string]any{"type": "string"},
						"requested_fields":          map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
						"adapter_status":            map[string]any{"type": "string"},
						"failure_code":              map[string]any{"type": "string"},
					},
					"required": []string{"passed", "subject_correct", "freshness_accepted", "fields_ready", "source_accepted", "advice_boundary_respected"},
				},
			},
		},
		EvalSuites: []agentxpack.EvalSuite{
			{
				Name:        "a_stock_valuation_success_suite",
				Description: "要求 A 股行情/估值工具返回 answer-ready 的主体化证据；估值解释边界由 evaluator 或 host gate 消费。",
				Mode:        agentxpack.EvalSuiteModeGate,
				WorkflowIDs: []string{DefaultWorkflow},
				RequiredState: []string{
					"valuation.answer_ready",
					"valuation.adapter_status",
					"valuation.subject_name",
					"valuation.subject_stock_code",
					"valuation.passed",
				},
				PassPath:    "valuation.passed",
				SummaryPath: "valuation.summary",
				Default:     true,
			},
		},
		PolicyProfiles: []agentxpack.PolicyProfile{
			{
				Name: "a_stock_valuation_readonly",
				Contract: agentxexecution.Contract{
					ID:      "a-stock-valuation-readonly",
					Strict:  true,
					Version: 1,
					Visibility: agentxexecution.VisibilityPolicy{
						AllowTools: []string{
							astockcontracts.ToolAStockQuoteLookup,
							astockcontracts.ToolAStockAnswerFormat,
						},
						DeclaredTools: []string{
							astockcontracts.ToolAStockQuoteLookup,
							astockcontracts.ToolAStockAnswerFormat,
						},
						RequireDeclared: true,
						MaxRisk:         "low",
					},
					Budget: agentxexecution.BudgetPolicy{MaxToolCalls: 3},
					Loop: agentxexecution.LoopPolicy{
						MaxRounds:              3,
						LoopDetectionEnabled:   true,
						ToolFailureFuseEnabled: true,
					},
					SideEffects: agentxexecution.SideEffectPolicy{
						MaxClass:       agentxexecution.SideEffectReadOnly,
						StrictRecovery: true,
					},
					Audit: agentxexecution.AuditPolicy{PersistSnapshot: true},
				},
				Default: true,
			},
		},
		MemorySchemas: []agentxpack.MemorySchema{
			{
				Name:        "a_stock_valuation_memory",
				Description: "沉淀 A 股行情/估值查询中的主体、字段、时点、来源、readiness 和失败原因。",
				Default:     true,
				Schema: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"pack_id":        map[string]any{"type": "string"},
						"case_type":      map[string]any{"type": "string"},
						"workflow_id":    map[string]any{"type": "string"},
						"subject_name":   map[string]any{"type": "string"},
						"stock_code":     map[string]any{"type": "string"},
						"quote_fields":   map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
						"source_url":     map[string]any{"type": "string"},
						"as_of":          map[string]any{"type": "string"},
						"adapter_status": map[string]any{"type": "string"},
						"failure_code":   map[string]any{"type": "string"},
						"answer_ready":   map[string]any{"type": "boolean"},
						"summary":        map[string]any{"type": "string"},
					},
					"required": []string{"pack_id", "case_type", "workflow_id", "summary"},
				},
			},
		},
		MemoryRecallPolicy: &agentxpack.MemoryRecallPolicy{
			QueryHints: []string{
				"A股估值",
				"行情 PE PB 市值",
				"stock valuation snapshot",
			},
			Limit:      4,
			MaxChars:   1200,
			ScopedOnly: true,
		},
	}
}

func MaterializedDefaultWorkflow(coordinator *agentxpack.Coordinator) (agentxworkflow.Spec, error) {
	return coordinator.MaterializeWorkflow(Definition(), DefaultWorkflow)
}

func RegisterInto(reg agentxpack.Registry) error {
	if reg == nil {
		return nil
	}
	return reg.Register(Definition())
}
