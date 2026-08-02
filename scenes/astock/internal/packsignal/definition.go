package packsignal

import (
	astockcontracts "github.com/wsnacj/agentx-go/scenes/astock/contracts"
	agentxpack "github.com/wsnacj/agentx-go/extensions/pack"
	agentxexecution "github.com/wsnacj/agentx-go/runtime/executionpolicy"
	agentxworkflow "github.com/wsnacj/agentx-go/runtime/workflow"
)

const (
	PackID          = "a-stock-signal-pack"
	CaseTypeSignal  = "a_stock.signal_lookup"
	DefaultWorkflow = "a_stock_signal_lookup_v1"
)

func signalCaseSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"user_message": map[string]any{"type": "string"},
			"entity_name": map[string]any{
				"type":        "string",
				"description": "主体名称；市场级信号可以填空字符串，个股资金流、概念、龙虎榜、解禁、行业对比等必须填公司简称或证券简称，由 adapter 验证代码。",
			},
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
			},
			"signal_types": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "string",
					"enum": []string{
						"hot_reason",
						"concept_blocks",
						"fund_flow",
						"northbound_flow",
						"dragon_tiger_board",
						"daily_dragon_tiger",
						"lockup_expiry",
						"industry_comparison",
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
					"trade_date":                 map[string]any{"type": "string"},
					"lookback_days":              map[string]any{"type": "integer"},
					"forward_days":               map[string]any{"type": "integer"},
					"require_latest_trading_day": map[string]any{"type": "boolean"},
				},
			},
			"source_policy":  map[string]any{"type": "string"},
			"limit":          map[string]any{"type": "integer"},
			"stop_condition": map[string]any{"type": "string"},
		},
		"required": []string{"user_message", "entity_name", "signal_types", "requested_outputs"},
	}
}

func Definition() agentxpack.Definition {
	return agentxpack.Definition{
		Manifest: agentxpack.Manifest{
			ID:      PackID,
			Version: "0.1.0",
			Domain:  "a_stock_signal",
			RouteHints: []string{
				"A股市场信号",
				"龙虎榜 资金流 热点原因 限售解禁",
				"A-share signal evidence",
				"fund flow dragon tiger board hot reason",
			},
			SupportedCaseTypes: []string{CaseTypeSignal},
			DefaultWorkflow:    DefaultWorkflow,
			PolicyProfiles:     []string{"a_stock_signal_readonly"},
			ArtifactTypes:      []string{"a_stock.signal.evidence"},
			Evaluators:         []string{"a_stock_signal_evidence_guard"},
			EvalSuites:         []string{"a_stock_signal_success_suite"},
		},
		CaseSchemas: []agentxpack.CaseSchema{
			{
				CaseType:    CaseTypeSignal,
				Description: "查询 A 股热点、资金流、龙虎榜、限售解禁、行业对比等公开市场信号，并保留来源、时点和降级原因。",
				RouteHints: []string{
					"A股信号",
					"资金流 龙虎榜 热点题材",
					"stock signal lookup",
				},
				Schema: signalCaseSchema(),
			},
		},
		Workflows: []agentxworkflow.Spec{
			{
				ID:              DefaultWorkflow,
				Title:           "A-Stock Signal Lookup",
				Description:     "使用模型填充的结构化 task frame 查询公开 A 股信号；来源选择和数据验证由 host/live adapter 负责。",
				Version:         "v1",
				Pack:            PackID,
				CaseTypes:       []string{CaseTypeSignal},
				RouteHints:      []string{"A股信号", "龙虎榜", "hot reason", "fund flow"},
				PlanningMode:    agentxworkflow.PlanningBounded,
				EntryNode:       "lookup_signal",
				DefaultContract: "a_stock_signal_readonly",
				StateSchema: []agentxworkflow.StateSlotSpec{
					{Name: "signal.answer_ready", Type: "boolean", Required: true},
					{Name: "signal.adapter_status", Type: "string", Required: true},
					{Name: "signal.failure_code", Type: "string"},
					{Name: "signal.degraded", Type: "boolean"},
					{Name: "signal.degrade_reason", Type: "string"},
					{Name: "signal.subject_name", Type: "string"},
					{Name: "signal.subject_stock_code", Type: "string"},
					{Name: "signal.trade_date", Type: "string"},
					{Name: "signal.as_of", Type: "string"},
					{Name: "signal.source_url", Type: "string"},
					{Name: "signal.events", Type: "array"},
					{Name: "signal.warnings", Type: "array"},
					{Name: "signal.summary", Type: "string", Required: true},
					{Name: "signal.passed", Type: "boolean", Required: true},
					{Name: "signal.subject_correct", Type: "boolean"},
					{Name: "signal.freshness_accepted", Type: "boolean"},
					{Name: "signal.fields_ready", Type: "boolean"},
					{Name: "signal.source_accepted", Type: "boolean"},
					{Name: "signal.advice_boundary_respected", Type: "boolean"},
					{Name: "signal.failure_reason", Type: "string"},
				},
				EvaluatorSchema: []agentxworkflow.EvaluatorRef{
					{Name: "a_stock_signal_evidence_guard"},
				},
				Nodes: []agentxworkflow.NodeSpec{
					{
						ID:          "lookup_signal",
						Kind:        agentxworkflow.NodeTool,
						Title:       "Lookup A-stock signal",
						Description: "查询由 task frame 指定的 A 股公开信号，并将 readiness/source/freshness 投影到 pack state。",
						Inputs: []agentxworkflow.BindingSpec{
							{From: "case.input.user_message", To: "args.user_message"},
							{From: "case.input.entity_name", To: "args.entity_name"},
							{From: "case.input.signal_types", To: "args.signal_types"},
							{From: "case.input.requested_outputs", To: "args.requested_outputs"},
						},
						Outputs: []agentxworkflow.BindingSpec{
							{From: "result.readiness.answer_ready", To: "state.signal.answer_ready"},
							{From: "result.readiness.answer_ready", To: "state.signal.passed"},
							{From: "result.adapter_status", To: "state.signal.adapter_status"},
							{From: "result.adapter_status", To: "state.signal.summary"},
						},
						Config: map[string]any{
							"tool_name": "lookup_a_stock_signal",
						},
					},
					{
						ID:          "format_answer",
						Kind:        agentxworkflow.NodeTool,
						Title:       "Format A-stock signal answer",
						Description: "把已验证的市场信号 payload 转为用户可读的中文回答；不新增事实、不生成投资建议。",
						Inputs: []agentxworkflow.BindingSpec{
							{From: "case.input.user_message", To: "args.user_message"},
							{From: "node.lookup_signal.result", To: "args.payload"},
						},
						Config: map[string]any{
							"tool_name":   "format_a_stock_answer",
							"answer_kind": "signal",
						},
					},
				},
				Edges: []agentxworkflow.EdgeSpec{
					{From: "lookup_signal", To: "format_answer", On: "success"},
				},
			},
		},
		Tools: []agentxpack.SemanticTool{
			{
				Name:        "lookup_a_stock_signal",
				Description: "查询 A 股公开市场信号。host/live adapter 负责来源策略、身份验证、时点和降级原因。",
				RuntimeTool: astockcontracts.ToolAStockSignalLookup,
				Tags:        []string{"a-stock", "signal", "public-source", "adapter"},
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
				Name:        "a_stock_signal_evidence_guard",
				Description: "检查 A 股信号查询的主体、时点、来源、请求字段和投资建议边界是否满足答复要求。",
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
						"source_urls":               map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
						"returned_signal_types":     map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
						"requested_signal_types":    map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
						"adapter_status":            map[string]any{"type": "string"},
						"failure_code":              map[string]any{"type": "string"},
					},
					"required": []string{"passed", "subject_correct", "freshness_accepted", "fields_ready", "source_accepted", "advice_boundary_respected"},
				},
			},
		},
		EvalSuites: []agentxpack.EvalSuite{
			{
				Name:        "a_stock_signal_success_suite",
				Description: "要求 A 股信号工具返回 answer-ready 的来源化证据；更细的主体/边界评估由 pack evaluator 或 host gate 消费。",
				Mode:        agentxpack.EvalSuiteModeGate,
				WorkflowIDs: []string{DefaultWorkflow},
				RequiredState: []string{
					"signal.answer_ready",
					"signal.adapter_status",
					"signal.passed",
				},
				PassPath:    "signal.passed",
				SummaryPath: "signal.summary",
				Default:     true,
			},
		},
		PolicyProfiles: []agentxpack.PolicyProfile{
			{
				Name: "a_stock_signal_readonly",
				Contract: agentxexecution.Contract{
					ID:      "a-stock-signal-readonly",
					Strict:  true,
					Version: 1,
					Visibility: agentxexecution.VisibilityPolicy{
						AllowTools: []string{
							astockcontracts.ToolAStockSignalLookup,
							astockcontracts.ToolAStockAnswerFormat,
						},
						DeclaredTools: []string{
							astockcontracts.ToolAStockSignalLookup,
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
				Name:        "a_stock_signal_memory",
				Description: "沉淀 A 股信号查询中的主体、信号类型、时点、来源、readiness 和失败原因。",
				Default:     true,
				Schema: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"pack_id":        map[string]any{"type": "string"},
						"case_type":      map[string]any{"type": "string"},
						"workflow_id":    map[string]any{"type": "string"},
						"subject_name":   map[string]any{"type": "string"},
						"stock_code":     map[string]any{"type": "string"},
						"signal_types":   map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
						"source_url":     map[string]any{"type": "string"},
						"trade_date":     map[string]any{"type": "string"},
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
				"A股信号",
				"资金流 龙虎榜 热点原因",
				"stock signal evidence",
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
