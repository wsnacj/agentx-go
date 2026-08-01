package packresearch

import (
	astockcontracts "github.com/wsnacj/agentx-go/extensions/astock/contracts"
	agentxpack "github.com/wsnacj/agentx-go/extensions/pack"
	agentxexecution "github.com/wsnacj/agentx-go/runtime/executionpolicy"
	agentxworkflow "github.com/wsnacj/agentx-go/runtime/workflow"
)

const (
	PackID           = "a-stock-research-pack"
	CaseTypeResearch = "a_stock.research_lookup"
	DefaultWorkflow  = "a_stock_research_lookup_v1"
)

func researchCaseSchema() map[string]any {
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
			"query": map[string]any{"type": "string"},
			"research_fields": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "string",
					"enum": []string{
						"title",
						"institution",
						"analyst",
						"published_at",
						"rating",
						"eps_forecast",
						"profit_forecast",
						"target_price",
						"pdf_url",
						"summary",
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
					"lookback_days":              map[string]any{"type": "integer"},
					"require_latest_trading_day": map[string]any{"type": "boolean"},
				},
			},
			"source_policy":  map[string]any{"type": "string"},
			"limit":          map[string]any{"type": "integer"},
			"stop_condition": map[string]any{"type": "string"},
		},
		"required": []string{"user_message", "entity", "research_fields", "requested_outputs"},
	}
}

func Definition() agentxpack.Definition {
	return agentxpack.Definition{
		Manifest: agentxpack.Manifest{
			ID:      PackID,
			Version: "0.1.0",
			Domain:  "a_stock_research",
			RouteHints: []string{
				"A股研报",
				"机构评级 盈利预测 目标价",
				"A-share research report",
				"rating eps forecast profit forecast",
			},
			SupportedCaseTypes: []string{CaseTypeResearch},
			DefaultWorkflow:    DefaultWorkflow,
			PolicyProfiles:     []string{"a_stock_research_readonly"},
			ArtifactTypes:      []string{"a_stock.research.evidence"},
			Evaluators:         []string{"a_stock_research_evidence_guard"},
			EvalSuites:         []string{"a_stock_research_success_suite"},
		},
		CaseSchemas: []agentxpack.CaseSchema{
			{
				CaseType:    CaseTypeResearch,
				Description: "查询 A 股公开研报、评级、机构观点和预测字段，并保留来源、发布时间、PDF 元数据和降级原因。",
				RouteHints:  []string{"A股研报", "机构评级", "盈利预测", "research report"},
				Schema:      researchCaseSchema(),
			},
		},
		Workflows: []agentxworkflow.Spec{
			{
				ID:              DefaultWorkflow,
				Title:           "A-Stock Research Lookup",
				Description:     "使用模型填充的结构化 task frame 查询 A 股研报列表和预测字段；adapter 验证主体、来源和发布时间。",
				Version:         "v1",
				Pack:            PackID,
				CaseTypes:       []string{CaseTypeResearch},
				RouteHints:      []string{"A股研报", "评级", "盈利预测", "research"},
				PlanningMode:    agentxworkflow.PlanningBounded,
				EntryNode:       "lookup_research",
				DefaultContract: "a_stock_research_readonly",
				StateSchema: []agentxworkflow.StateSlotSpec{
					{Name: "research.answer_ready", Type: "boolean", Required: true},
					{Name: "research.adapter_status", Type: "string", Required: true},
					{Name: "research.failure_code", Type: "string"},
					{Name: "research.degraded", Type: "boolean"},
					{Name: "research.degrade_reason", Type: "string"},
					{Name: "research.subject_name", Type: "string", Required: true},
					{Name: "research.subject_stock_code", Type: "string", Required: true},
					{Name: "research.query", Type: "string"},
					{Name: "research.as_of", Type: "string"},
					{Name: "research.source_url", Type: "string"},
					{Name: "research.reports", Type: "array"},
					{Name: "research.consensus", Type: "array"},
					{Name: "research.latest_title", Type: "string"},
					{Name: "research.latest_institution", Type: "string"},
					{Name: "research.latest_published_at", Type: "string"},
					{Name: "research.latest_rating", Type: "string"},
					{Name: "research.latest_pdf_url", Type: "string"},
					{Name: "research.warnings", Type: "array"},
					{Name: "research.summary", Type: "string", Required: true},
					{Name: "research.passed", Type: "boolean", Required: true},
					{Name: "research.subject_correct", Type: "boolean"},
					{Name: "research.freshness_accepted", Type: "boolean"},
					{Name: "research.fields_ready", Type: "boolean"},
					{Name: "research.source_accepted", Type: "boolean"},
					{Name: "research.advice_boundary_respected", Type: "boolean"},
					{Name: "research.failure_reason", Type: "string"},
				},
				EvaluatorSchema: []agentxworkflow.EvaluatorRef{
					{Name: "a_stock_research_evidence_guard"},
				},
				Nodes: []agentxworkflow.NodeSpec{
					{
						ID:          "lookup_research",
						Kind:        agentxworkflow.NodeTool,
						Title:       "Lookup A-stock research",
						Description: "查询主体公开研报和预测字段，并将 readiness/source/freshness 投影到 pack state。",
						Inputs: []agentxworkflow.BindingSpec{
							{From: "case.input.user_message", To: "args.user_message"},
							{From: "case.input.entity.name", To: "args.entity_name"},
							{From: "case.input.research_fields", To: "args.research_fields"},
						},
						Outputs: []agentxworkflow.BindingSpec{
							{From: "result.readiness.answer_ready", To: "state.research.answer_ready"},
							{From: "result.readiness.answer_ready", To: "state.research.passed"},
							{From: "result.adapter_status", To: "state.research.adapter_status"},
							{From: "result.subject.entity_name", To: "state.research.subject_name"},
							{From: "result.subject.stock_code", To: "state.research.subject_stock_code"},
							{From: "result.freshness.as_of", To: "state.research.as_of"},
							{From: "result.evidence.source_url", To: "state.research.source_url"},
							{From: "result.reports", To: "state.research.reports"},
							{From: "result.adapter_status", To: "state.research.summary"},
						},
						Config: map[string]any{
							"tool_name": "lookup_a_stock_research",
						},
					},
					{
						ID:          "format_answer",
						Kind:        agentxworkflow.NodeTool,
						Title:       "Format A-stock research answer",
						Description: "把已验证的研报 payload 转为用户可读的中文回答；不新增事实、不生成投资建议。",
						Inputs: []agentxworkflow.BindingSpec{
							{From: "case.input.user_message", To: "args.user_message"},
							{From: "node.lookup_research.result", To: "args.payload"},
						},
						Config: map[string]any{
							"tool_name":   "format_a_stock_answer",
							"answer_kind": "research",
						},
					},
				},
				Edges: []agentxworkflow.EdgeSpec{
					{From: "lookup_research", To: "format_answer", On: "success"},
				},
			},
		},
		Tools: []agentxpack.SemanticTool{
			{
				Name:        "lookup_a_stock_research",
				Description: "查询 A 股公开研报、评级和预测字段。host/live adapter 负责主体验证、来源、发布时间和降级原因。",
				RuntimeTool: astockcontracts.ToolAStockResearchLookup,
				Tags:        []string{"a-stock", "research", "rating", "forecast", "public-source", "adapter"},
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
				Name:        "a_stock_research_evidence_guard",
				Description: "检查 A 股研报查询的主体、发布时间、请求字段、来源和投资建议边界是否满足答复要求。",
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
				Name:        "a_stock_research_success_suite",
				Description: "要求 A 股研报工具返回 answer-ready 的主体化证据；评级/预测解释边界由 evaluator 或 host gate 消费。",
				Mode:        agentxpack.EvalSuiteModeGate,
				WorkflowIDs: []string{DefaultWorkflow},
				RequiredState: []string{
					"research.answer_ready",
					"research.adapter_status",
					"research.subject_name",
					"research.subject_stock_code",
					"research.passed",
				},
				PassPath:    "research.passed",
				SummaryPath: "research.summary",
				Default:     true,
			},
		},
		PolicyProfiles: []agentxpack.PolicyProfile{
			{
				Name: "a_stock_research_readonly",
				Contract: agentxexecution.Contract{
					ID:      "a-stock-research-readonly",
					Strict:  true,
					Version: 1,
					Visibility: agentxexecution.VisibilityPolicy{
						AllowTools: []string{
							astockcontracts.ToolAStockResearchLookup,
							astockcontracts.ToolAStockAnswerFormat,
						},
						DeclaredTools: []string{
							astockcontracts.ToolAStockResearchLookup,
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
				Name:        "a_stock_research_memory",
				Description: "沉淀 A 股研报查询中的主体、字段、来源、发布时间、readiness 和失败原因。",
				Default:     true,
				Schema: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"pack_id":             map[string]any{"type": "string"},
						"case_type":           map[string]any{"type": "string"},
						"workflow_id":         map[string]any{"type": "string"},
						"subject_name":        map[string]any{"type": "string"},
						"stock_code":          map[string]any{"type": "string"},
						"research_fields":     map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
						"source_url":          map[string]any{"type": "string"},
						"latest_published_at": map[string]any{"type": "string"},
						"adapter_status":      map[string]any{"type": "string"},
						"failure_code":        map[string]any{"type": "string"},
						"answer_ready":        map[string]any{"type": "boolean"},
						"summary":             map[string]any{"type": "string"},
					},
					"required": []string{"pack_id", "case_type", "workflow_id", "summary"},
				},
			},
		},
		MemoryRecallPolicy: &agentxpack.MemoryRecallPolicy{
			QueryHints: []string{
				"A股研报",
				"机构评级 盈利预测",
				"stock research report",
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
