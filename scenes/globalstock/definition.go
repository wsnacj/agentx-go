package globalstock

import (
	agentxpack "github.com/wsnacj/agentx-go/extensions/pack"
	agentxexecution "github.com/wsnacj/agentx-go/runtime/executionpolicy"
	agentxworkflow "github.com/wsnacj/agentx-go/runtime/workflow"
	globalcontracts "github.com/wsnacj/agentx-go/scenes/globalstock/contracts"
)

const (
	PackID             = "global-stock-quote-pack"
	CaseTypeQuote      = "global_stock.quote_snapshot"
	CaseTypeValuation  = "global_stock.valuation_snapshot"
	CaseTypeComparison = "global_stock.valuation_comparison"
	DefaultWorkflow    = "global_stock_quote_lookup_v1"
	ComparisonWorkflow = "global_stock_comparison_lookup_v1"
)

func quoteCaseSchema() map[string]any {
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
					"enum": []string{"price", "change_pct", "open", "high", "low", "last_close", "volume", "amount", "pe_ttm", "pb", "market_cap"},
				},
			},
			"requested_outputs": map[string]any{
				"type":  "array",
				"items": map[string]any{"type": "string"},
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
		"required": []string{"user_message", "entity", "quote_fields"},
	}
}

func comparisonCaseSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"user_message": map[string]any{"type": "string"},
			"entities": map[string]any{
				"type":  "array",
				"items": map[string]any{"type": "string"},
			},
			"requested_fields": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "string",
					"enum": []string{"price", "change_pct", "open", "high", "low", "last_close", "volume", "amount", "pe_ttm", "pb", "market_cap"},
				},
			},
			"requested_outputs": map[string]any{
				"type":  "array",
				"items": map[string]any{"type": "string"},
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
		"required": []string{"user_message", "entities"},
	}
}

func Definition() agentxpack.Definition {
	return agentxpack.Definition{
		Manifest: agentxpack.Manifest{
			ID:      PackID,
			Version: "0.1.0",
			Domain:  "global_stock_quote",
			RouteHints: []string{
				"港股行情",
				"美股行情",
				"HK/US stock quote",
				"global stock quote",
			},
			SupportedCaseTypes: []string{CaseTypeQuote, CaseTypeValuation, CaseTypeComparison},
			DefaultWorkflow:    DefaultWorkflow,
			PolicyProfiles:     []string{"global_stock_quote_readonly", "global_stock_comparison_readonly"},
			ArtifactTypes:      []string{"global_stock.quote.evidence"},
			Evaluators:         []string{"global_stock_quote_evidence_guard"},
			EvalSuites:         []string{"global_stock_quote_success_suite"},
		},
		CaseSchemas: []agentxpack.CaseSchema{
			{
				CaseType:    CaseTypeQuote,
				Description: "查询港股/美股公开行情快照，包括价格、涨跌幅、成交、市值等点时证据。",
				RouteHints:  []string{"港股行情", "美股行情", "quote snapshot"},
				Schema:      quoteCaseSchema(),
			},
			{
				CaseType:    CaseTypeValuation,
				Description: "查询港股/美股点时估值快照，包括估值倍数、市值等来源化证据；不自动宣称历史分位或投资结论。",
				RouteHints:  []string{"港股估值", "美股估值", "估值快照", "valuation snapshot"},
				Schema:      quoteCaseSchema(),
			},
			{
				CaseType:    CaseTypeComparison,
				Description: "比较多个港股/美股主体的点时估值快照；逐主体保留身份、字段、时点、来源和缺失边界。",
				RouteHints:  []string{"valuation snapshot"},
				Schema:      comparisonCaseSchema(),
			},
		},
		Workflows: []agentxworkflow.Spec{
			{
				ID:              DefaultWorkflow,
				Title:           "Global Stock Quote Lookup",
				Description:     "使用模型填充的结构化 task frame 查询港股/美股行情/估值快照；adapter 验证主体、市场、币种、时点和字段。",
				Version:         "v1",
				Pack:            PackID,
				CaseTypes:       []string{CaseTypeQuote, CaseTypeValuation},
				RouteHints:      []string{"港股行情", "美股行情", "行情快照"},
				PlanningMode:    agentxworkflow.PlanningBounded,
				EntryNode:       "lookup_quote",
				DefaultContract: "global_stock_quote_readonly",
				StateSchema: []agentxworkflow.StateSlotSpec{
					{Name: "quote.answer_ready", Type: "boolean", Required: true},
					{Name: "quote.adapter_status", Type: "string", Required: true},
					{Name: "quote.failure_code", Type: "string"},
					{Name: "quote.subject_name", Type: "string"},
					{Name: "quote.subject_stock_code", Type: "string"},
					{Name: "quote.market", Type: "string"},
					{Name: "quote.exchange", Type: "string"},
					{Name: "quote.currency", Type: "string"},
					{Name: "quote.as_of", Type: "string"},
					{Name: "quote.source_url", Type: "string"},
					{Name: "quote.price", Type: "string"},
					{Name: "quote.change_percent", Type: "string"},
					{Name: "quote.market_cap", Type: "string"},
					{Name: "quote.summary", Type: "string", Required: true},
					{Name: "quote.passed", Type: "boolean", Required: true},
				},
				EvaluatorSchema: []agentxworkflow.EvaluatorRef{
					{Name: "global_stock_quote_evidence_guard"},
				},
				Nodes: []agentxworkflow.NodeSpec{
					{
						ID:          "lookup_quote",
						Kind:        agentxworkflow.NodeTool,
						Title:       "Lookup global stock quote",
						Description: "查询主体的公开行情/估值快照，并将 readiness/source/freshness 投影到 pack state。",
						Inputs: []agentxworkflow.BindingSpec{
							{From: "case.input.user_message", To: "args.user_message"},
							{From: "case.input.entity.name", To: "args.entity_name"},
							{From: "case.input.quote_fields", To: "args.quote_fields"},
						},
						Outputs: []agentxworkflow.BindingSpec{
							{From: "result.readiness.answer_ready", To: "state.quote.answer_ready"},
							{From: "result.readiness.answer_ready", To: "state.quote.passed"},
							{From: "result.adapter_status", To: "state.quote.adapter_status"},
							{From: "result.adapter_status", To: "state.quote.summary"},
						},
						Config: map[string]any{
							"tool_name": "lookup_global_stock_quote",
						},
					},
					{
						ID:          "format_answer",
						Kind:        agentxworkflow.NodeTool,
						Title:       "Format global stock quote answer",
						Description: "把已验证的行情/估值 payload 转为用户可读的中文回答；不新增事实、不生成投资建议。",
						Inputs: []agentxworkflow.BindingSpec{
							{From: "case.input.user_message", To: "args.user_message"},
							{From: "case.input.quote_fields", To: "args.requested_fields"},
							{From: "node.lookup_quote.result", To: "args.payload"},
						},
						Config: map[string]any{
							"tool_name":   "format_global_stock_answer",
							"answer_kind": "quote",
						},
					},
				},
				Edges: []agentxworkflow.EdgeSpec{
					{From: "lookup_quote", To: "format_answer", On: "success"},
				},
			},
			{
				ID:              ComparisonWorkflow,
				Title:           "Compare Global Stocks",
				Description:     "比较多个港股/美股主体的行情与估值字段，由高层 investigation 批量取证并逐主体保留缺失边界。",
				Version:         "v1",
				Pack:            PackID,
				CaseTypes:       []string{CaseTypeComparison},
				PlanningMode:    agentxworkflow.PlanningBounded,
				EntryNode:       "lookup_comparison",
				DefaultContract: "global_stock_comparison_readonly",
				StateSchema: []agentxworkflow.StateSlotSpec{
					{Name: "quote.answer_ready", Type: "boolean", Required: true},
					{Name: "quote.adapter_status", Type: "string", Required: true},
					{Name: "quote.failure_code", Type: "string"},
					{Name: "quote.summary", Type: "string", Required: true},
					{Name: "quote.passed", Type: "boolean", Required: true},
				},
				EvaluatorSchema: []agentxworkflow.EvaluatorRef{
					{Name: "global_stock_quote_evidence_guard"},
				},
				Nodes: []agentxworkflow.NodeSpec{
					{
						ID:          "lookup_comparison",
						Kind:        agentxworkflow.NodeTool,
						Title:       "Lookup global stock comparison",
						Description: "批量查询多个主体的公开行情/估值快照，并聚合 readiness。",
						Inputs: []agentxworkflow.BindingSpec{
							{From: "case.input.user_message", To: "args.user_message"},
							{From: "case.input.entities", To: "args.entity_mentions"},
							{From: "case.input.requested_fields", To: "args.requested_fields", Optional: true},
							{From: "case.input.requested_outputs", To: "args.requested_outputs", Optional: true},
						},
						Outputs: []agentxworkflow.BindingSpec{
							{From: "result.readiness.answer_ready", To: "state.quote.answer_ready"},
							{From: "result.readiness.answer_ready", To: "state.quote.passed"},
							{From: "result.adapter_status", To: "state.quote.adapter_status"},
							{From: "result.adapter_status", To: "state.quote.summary"},
						},
						Config: map[string]any{
							"tool_name": "investigate_global_stocks",
							"task_kind": string(globalcontracts.TaskKindComparison),
						},
					},
					{
						ID:          "format_answer",
						Kind:        agentxworkflow.NodeTool,
						Title:       "Format global stock comparison answer",
						Description: "把逐主体核验后的行情/估值 payload 转为用户可读回答，不新增事实。",
						Inputs: []agentxworkflow.BindingSpec{
							{From: "case.input.user_message", To: "args.user_message"},
							{From: "node.lookup_comparison.result", To: "args.payload"},
						},
						Config: map[string]any{
							"tool_name":   "format_global_stock_answer",
							"answer_kind": "comparison",
						},
					},
				},
				Edges: []agentxworkflow.EdgeSpec{
					{From: "lookup_comparison", To: "format_answer", On: "success"},
				},
			},
		},
		Tools: []agentxpack.SemanticTool{
			{
				Name:        "lookup_global_stock_quote",
				Description: "查询港股/美股行情和点时估值快照。host/live adapter 负责主体验证、市场、币种、时点、来源和降级原因。",
				RuntimeTool: ToolGlobalStockQuoteLookup,
				Tags:        []string{"global-stock", "hk-stock", "us-stock", "quote", "valuation", "public-source", "adapter"},
			},
			{
				Name:        "investigate_global_stocks",
				Description: "批量查询多个港股/美股主体的行情和点时估值，逐主体保留身份、来源、时点和字段 readiness。",
				RuntimeTool: ToolGlobalStockInvestigation,
				RuntimeArgs: map[string]any{
					"default_requested_fields":  []any{"price", "pe_ttm", "pb", "market_cap"},
					"default_requested_outputs": []any{"comparison", "valuation_snapshot"},
				},
				Tags: []string{"global-stock", "multi-stock", "comparison", "valuation", "public-source", "adapter"},
			},
			{
				Name:        "format_global_stock_answer",
				Description: "将已验证的港股/美股工具 payload 格式化为中文最终回答，不新增事实。",
				RuntimeTool: ToolGlobalStockAnswerFormat,
				Tags:        []string{"global-stock", "answer-format", "public-source"},
			},
		},
		Evaluators: []agentxpack.Evaluator{
			{
				Name:        "global_stock_quote_evidence_guard",
				Description: "检查港股/美股行情查询的主体、市场、币种、时点、来源和请求字段是否满足答复要求。",
				OutputSchema: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"passed":             map[string]any{"type": "boolean"},
						"subject_correct":    map[string]any{"type": "boolean"},
						"freshness_accepted": map[string]any{"type": "boolean"},
						"fields_ready":       map[string]any{"type": "boolean"},
						"source_accepted":    map[string]any{"type": "boolean"},
						"failure_reason":     map[string]any{"type": "string"},
						"missing_fields":     map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
					},
				},
			},
		},
		EvalSuites: []agentxpack.EvalSuite{
			{
				Name:        "global_stock_quote_success_suite",
				Description: "要求港股/美股行情工具完成 lookup 并记录 readiness；answer_ready=false 时保留降级说明而不让运行失败。",
				Mode:        agentxpack.EvalSuiteModeGate,
				WorkflowIDs: []string{DefaultWorkflow, ComparisonWorkflow},
				RequiredState: []string{
					"quote.answer_ready",
					"quote.adapter_status",
					"quote.passed",
				},
				SummaryPath: "quote.summary",
				Default:     true,
			},
		},
		PolicyProfiles: []agentxpack.PolicyProfile{
			{
				Name: "global_stock_quote_readonly",
				Contract: agentxexecution.Contract{
					ID:      "global-stock-quote-readonly",
					Strict:  true,
					Version: 1,
					Visibility: agentxexecution.VisibilityPolicy{
						AllowTools: []string{
							ToolGlobalStockQuoteLookup,
							ToolGlobalStockAnswerFormat,
						},
						DeclaredTools: []string{
							ToolGlobalStockQuoteLookup,
							ToolGlobalStockAnswerFormat,
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
			{
				Name: "global_stock_comparison_readonly",
				Contract: agentxexecution.Contract{
					ID:      "global-stock-comparison-readonly",
					Strict:  true,
					Version: 1,
					Visibility: agentxexecution.VisibilityPolicy{
						AllowTools: []string{
							ToolGlobalStockInvestigation,
							ToolGlobalStockAnswerFormat,
						},
						DeclaredTools: []string{
							ToolGlobalStockInvestigation,
							ToolGlobalStockAnswerFormat,
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
			},
		},
	}
}

func MaterializedDefaultWorkflow(coordinator *agentxpack.Coordinator) (agentxworkflow.Spec, error) {
	return coordinator.MaterializeWorkflow(Definition(), DefaultWorkflow)
}

func MaterializedComparisonWorkflow(coordinator *agentxpack.Coordinator) (agentxworkflow.Spec, error) {
	return coordinator.MaterializeWorkflow(Definition(), ComparisonWorkflow)
}

func RegisterInto(reg agentxpack.Registry) error {
	if reg == nil {
		return nil
	}
	return reg.Register(Definition())
}
