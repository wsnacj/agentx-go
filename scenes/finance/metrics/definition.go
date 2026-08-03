package metrics

import (
	agentxpack "github.com/wsnacj/agentx-go/extensions/pack"
	agentxexecution "github.com/wsnacj/agentx-go/runtime/executionpolicy"
	agentxworkflow "github.com/wsnacj/agentx-go/runtime/workflow"
)

const (
	PackID          = "financial-report-metrics-pack"
	CaseTypeLatest  = "financial_report.latest_metrics"
	CaseTypeTrend   = "financial_report.metrics_trend"
	DefaultWorkflow = "financial_report_metrics_lookup_v1"
)

func latestReportMetricsCaseSchema() map[string]any {
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
							"ticker":     map[string]any{"type": "string"},
							"exchange":   map[string]any{"type": "string"},
							"market":     map[string]any{"type": "string"},
						},
					},
				},
				"required": []string{"name"},
			},
			"requested_metrics": map[string]any{
				"type":  "array",
				"items": map[string]any{"type": "string"},
			},
			"requested_outputs": map[string]any{
				"type":  "array",
				"items": map[string]any{"type": "string"},
			},
			"assessment": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"kind":               map[string]any{"type": "string"},
					"scope":              map[string]any{"type": "string"},
					"requires_valuation": map[string]any{"type": "boolean"},
				},
				"required": []string{"kind", "scope", "requires_valuation"},
			},
			"period_policy": map[string]any{"type": "string"},
			"source_policy": map[string]any{"type": "string"},
			"freshness":     map[string]any{"type": "string"},
			"stop_condition": map[string]any{
				"type": "string",
			},
		},
		"required": []string{"user_message", "entity", "requested_metrics", "requested_outputs", "assessment"},
	}
}

func Definition() agentxpack.Definition {
	return agentxpack.Definition{
		Manifest: agentxpack.Manifest{
			ID:      PackID,
			Version: "0.1.0",
			Domain:  "financial_report_metrics",
			RouteHints: []string{
				"latest financial report metrics",
				"public company revenue profit growth",
				"multi-year financial report metrics trend",
				"财报指标",
				"营收 净利润 增长率",
				"近几年 财报 趋势",
				"最新年报 指标",
			},
			SupportedCaseTypes: []string{CaseTypeLatest, CaseTypeTrend},
			DefaultWorkflow:    DefaultWorkflow,
			PolicyProfiles:     []string{"public_web_report_metrics_readonly"},
			ArtifactTypes:      []string{"public_web.report_metrics.evidence"},
			Evaluators:         []string{"financial_report_metrics_guard"},
			EvalSuites:         []string{"financial_report_metrics_success_suite"},
		},
		CaseSchemas: []agentxpack.CaseSchema{
			{
				CaseType:    CaseTypeLatest,
				Description: "从公开来源提取主体最新已披露报告中的收入、利润、增长率、现金流等指标，并在 guard 通过后停止。",
				RouteHints: []string{
					"查财报指标",
					"最新年报营收利润",
					"revenue net profit growth",
					"financial report metrics",
				},
				Schema: latestReportMetricsCaseSchema(),
			},
			{
				CaseType:    CaseTypeTrend,
				Description: "从公开来源提取主体近几年或多期报告中的收入、利润等指标序列，并校验连续期间、单位币种和增长计算依据。",
				RouteHints: []string{
					"近几年财报增长",
					"收入利润趋势",
					"multi-year revenue profit trend",
					"recent years financial metrics",
				},
				Schema: latestReportMetricsCaseSchema(),
			},
		},
		Workflows: []agentxworkflow.Spec{
			{
				ID:              DefaultWorkflow,
				Title:           "Financial Report Metrics Lookup",
				Description:     "解析公开公司/主体，生成低噪候选页，读取页面，抽取指标，并用 guard 判断是否可以停止。",
				Version:         "v1",
				Pack:            PackID,
				CaseTypes:       []string{CaseTypeLatest, CaseTypeTrend},
				RouteHints:      []string{"财报指标", "报告指标", "revenue profit growth"},
				PlanningMode:    agentxworkflow.PlanningBounded,
				EntryNode:       "generate_candidates",
				DefaultContract: "public_web_report_metrics_readonly",
				StateSchema: []agentxworkflow.StateSlotSpec{
					{Name: "candidate.primary_url", Type: "string", Required: true},
					{Name: "candidate.source_kind", Type: "string"},
					{Name: "candidate.adapter_id", Type: "string"},
					{Name: "candidate.adapter_status", Type: "string"},
					{Name: "candidate.source_policy", Type: "string"},
					{Name: "candidate.failure_code", Type: "string"},
					{Name: "page.final_url", Type: "string", Required: true},
					{Name: "metrics.requested_fields_ready", Type: "boolean", Required: true},
					{Name: "metrics.guard_status", Type: "string", Required: true},
					{Name: "metrics.adapter_id", Type: "string"},
					{Name: "metrics.adapter_status", Type: "string"},
					{Name: "metrics.source_policy", Type: "string"},
					{Name: "metrics.failure_code", Type: "string"},
					{Name: "metrics.requested_outputs", Type: "array"},
					{Name: "metrics.assessment_kind", Type: "string"},
					{Name: "metrics.assessment_scope", Type: "string"},
					{Name: "metrics.passed", Type: "boolean", Required: true},
					{Name: "metrics.subject_correct", Type: "boolean", Required: true},
					{Name: "metrics.period_latest", Type: "boolean", Required: true},
					{Name: "metrics.growth_fields_consistent", Type: "boolean", Required: true},
					{Name: "metrics.source_accepted", Type: "boolean", Required: true},
					{Name: "metrics.field_sources_accepted", Type: "boolean", Required: true},
					{Name: "metrics.stop_after_guard_passed", Type: "boolean", Required: true},
					{Name: "metrics.trend_series_ready", Type: "boolean"},
					{Name: "metrics.trend_series_period_count", Type: "number"},
					{Name: "metrics.missing_requested_fields", Type: "array"},
					{Name: "metrics.review_required_fields", Type: "array"},
					{Name: "metrics.failure_reason", Type: "string"},
					{Name: "metrics.report_period", Type: "string", Required: true},
					{Name: "metrics.source_url", Type: "string", Required: true},
					{Name: "metrics.summary", Type: "string", Required: true},
				},
				EvaluatorSchema: []agentxworkflow.EvaluatorRef{
					{Name: "financial_report_metrics_guard"},
				},
				Nodes: []agentxworkflow.NodeSpec{
					{
						ID:          "generate_candidates",
						Kind:        agentxworkflow.NodeTool,
						Title:       "Generate report candidates",
						Description: "从主体、证券代码、用户原始请求和来源策略生成低噪报告候选页。",
						Inputs: []agentxworkflow.BindingSpec{
							{From: "case.input.user_message", To: "args.user_message"},
							{From: "case.input.entity.name", To: "args.entity_name"},
							{From: "case.input.entity.identifiers.stock_code", To: "args.stock_code"},
							{From: "case.input.entity.identifiers.ticker", To: "args.ticker"},
							{From: "case.input.requested_metrics", To: "args.requested_metrics"},
							{From: "case.input.requested_outputs", To: "args.requested_outputs"},
							{From: "case.input.assessment.kind", To: "args.assessment_kind"},
							{From: "case.input.assessment.scope", To: "args.assessment_scope"},
							{From: "case.input.assessment.requires_valuation", To: "args.assessment_requires_valuation"},
							{From: "case.input.period_policy", To: "args.period_scope"},
							{From: "case.input.source_policy", To: "args.source_policy"},
						},
						Outputs: []agentxworkflow.BindingSpec{
							{From: "result.primary_url", To: "state.candidate.primary_url"},
							{From: "result.primary_source_kind", To: "state.candidate.source_kind"},
							{From: "result.adapter_id", To: "state.candidate.adapter_id"},
							{From: "result.adapter_status", To: "state.candidate.adapter_status"},
							{From: "result.source_policy", To: "state.candidate.source_policy"},
							{From: "result.failure_code", To: "state.candidate.failure_code"},
						},
						Config: map[string]any{
							"tool_name": "generate_report_candidates",
						},
					},
					{
						ID:          "open_candidate",
						Kind:        agentxworkflow.NodeTool,
						Title:       "Open primary candidate",
						Description: "打开最高优先级候选页，保留 final_url 作为后续抽取证据；如果静态页面读取失败，后续 project adapter 仍可基于候选 URL 做专用获取。",
						Inputs: []agentxworkflow.BindingSpec{
							{From: "state.candidate.primary_url", To: "args.url"},
						},
						Outputs: []agentxworkflow.BindingSpec{
							{From: "result.final_url", To: "state.page.final_url"},
						},
						Config: map[string]any{
							"tool_name": "open_report_candidate",
						},
					},
					{
						ID:          "extract_metrics",
						Kind:        agentxworkflow.NodeTool,
						Title:       "Extract requested metrics",
						Description: "从已打开页面、页面缓存或 adapter fallback 中抽取请求字段与候选证据。",
						Inputs: []agentxworkflow.BindingSpec{
							{From: "case.input.user_message", To: "args.user_message"},
							{From: "case.input.entity.name", To: "args.entity_name"},
							{From: "case.input.entity.identifiers.stock_code", To: "args.stock_code"},
							{From: "case.input.entity.identifiers.ticker", To: "args.ticker"},
							{From: "case.input.requested_metrics", To: "args.requested_metrics"},
							{From: "case.input.requested_outputs", To: "args.requested_outputs"},
							{From: "case.input.assessment.kind", To: "args.assessment_kind"},
							{From: "case.input.assessment.scope", To: "args.assessment_scope"},
							{From: "case.input.assessment.requires_valuation", To: "args.assessment_requires_valuation"},
							{From: "case.input.period_policy", To: "args.period_scope"},
							{From: "case.input.source_policy", To: "args.source_policy"},
							{From: "state.page.final_url", To: "args.source_url"},
						},
						Config: map[string]any{
							"tool_name": "extract_report_metrics",
						},
					},
					{
						ID:          "guard_metrics",
						Kind:        agentxworkflow.NodeTool,
						Title:       "Guard requested metrics",
						Description: "确认请求字段、来源、报告期和增长率是否足以支撑最终答案；通过后应停止。",
						Inputs: []agentxworkflow.BindingSpec{
							{From: "case.input.user_message", To: "args.user_message"},
							{From: "case.input.entity.name", To: "args.entity_name"},
							{From: "case.input.entity.identifiers.ticker", To: "args.ticker"},
							{From: "case.input.requested_metrics", To: "args.requested_metrics"},
							{From: "case.input.requested_outputs", To: "args.requested_outputs"},
							{From: "case.input.assessment.kind", To: "args.assessment_kind"},
							{From: "case.input.assessment.scope", To: "args.assessment_scope"},
							{From: "case.input.assessment.requires_valuation", To: "args.assessment_requires_valuation"},
							{From: "case.input.period_policy", To: "args.period_scope"},
							{From: "case.input.source_policy", To: "args.source_policy"},
							{From: "state.page.final_url", To: "args.source_url"},
							{From: "node.extract_metrics.result.evidence.company_name", To: "args.company_name"},
							{From: "node.extract_metrics.result.evidence.stock_code", To: "args.stock_code"},
							{From: "node.extract_metrics.result.evidence.selection_reason", To: "args.selection_reason"},
							{From: "node.extract_metrics.result.evidence.official_source", To: "args.official_source"},
							{From: "node.extract_metrics.result.evidence.report_period", To: "args.report_period"},
							{From: "node.extract_metrics.result.evidence.revenue", To: "args.revenue"},
							{From: "node.extract_metrics.result.evidence.revenue_growth", To: "args.revenue_growth"},
							{From: "node.extract_metrics.result.evidence.net_profit", To: "args.net_profit"},
							{From: "node.extract_metrics.result.evidence.net_profit_growth", To: "args.net_profit_growth"},
							{From: "node.extract_metrics.result.evidence.operating_cash_flow", To: "args.operating_cash_flow"},
							{From: "node.extract_metrics.result.evidence.page_title", To: "args.page_title"},
							{From: "node.extract_metrics.result.evidence.trend_series", To: "args.trend_series"},
						},
						Outputs: []agentxworkflow.BindingSpec{
							{From: "result.requested_fields_ready", To: "state.metrics.requested_fields_ready"},
							{From: "result.guard_status", To: "state.metrics.guard_status"},
							{From: "result.adapter_id", To: "state.metrics.adapter_id"},
							{From: "result.adapter_status", To: "state.metrics.adapter_status"},
							{From: "result.source_policy", To: "state.metrics.source_policy"},
							{From: "result.failure_code", To: "state.metrics.failure_code"},
							{From: "result.requested_outputs", To: "state.metrics.requested_outputs"},
							{From: "result.assessment_kind", To: "state.metrics.assessment_kind"},
							{From: "result.assessment_scope", To: "state.metrics.assessment_scope"},
							{From: "result.evaluation.passed", To: "state.metrics.passed"},
							{From: "result.evaluation.subject_correct", To: "state.metrics.subject_correct"},
							{From: "result.evaluation.period_latest", To: "state.metrics.period_latest"},
							{From: "result.evaluation.growth_fields_consistent", To: "state.metrics.growth_fields_consistent"},
							{From: "result.evaluation.source_accepted", To: "state.metrics.source_accepted"},
							{From: "result.evaluation.field_sources_accepted", To: "state.metrics.field_sources_accepted"},
							{From: "result.evaluation.stop_after_guard_passed", To: "state.metrics.stop_after_guard_passed"},
							{From: "result.evaluation.trend_series_ready", To: "state.metrics.trend_series_ready"},
							{From: "result.evaluation.trend_series_period_count", To: "state.metrics.trend_series_period_count"},
							{From: "result.missing_requested_fields", To: "state.metrics.missing_requested_fields"},
							{From: "result.review_required_fields", To: "state.metrics.review_required_fields"},
							{From: "result.evaluation.failure_reason", To: "state.metrics.failure_reason"},
							{From: "result.evidence.report_period", To: "state.metrics.report_period"},
							{From: "result.final_url", To: "state.metrics.source_url"},
							{From: "result.guard_status", To: "state.metrics.summary"},
						},
						Config: map[string]any{
							"tool_name": "guard_report_metrics",
						},
					},
				},
				Edges: []agentxworkflow.EdgeSpec{
					{From: "generate_candidates", To: "open_candidate", On: "success"},
					{From: "open_candidate", To: "extract_metrics", On: "success"},
					{From: "extract_metrics", To: "guard_metrics", On: "success"},
				},
			},
		},
		Tools: []agentxpack.SemanticTool{
			{
				Name:        "generate_report_candidates",
				Description: "生成公开报告指标候选页。当前 runtime adapter 原型由 project/plugin/extension 提供，不进入 core retrieval。",
				RuntimeTool: "report_metrics_candidates",
				Tags:        []string{"public-web", "candidate-generation", "adapter"},
			},
			{
				Name:        "open_report_candidate",
				Description: "打开候选页并把 page_id/final_url 留给后续抽取。",
				RuntimeTool: "open_page",
				Tags:        []string{"public-web", "page-read"},
			},
			{
				Name:        "extract_report_metrics",
				Description: "从页面正文、页面缓存或 adapter fallback 中抽取财报指标候选值。",
				RuntimeTool: "report_metrics_extract",
				Tags:        []string{"public-web", "schema-extraction", "adapter"},
			},
			{
				Name:        "guard_report_metrics",
				Description: "确认请求字段是否完整、是否需要复核、是否可以停止并输出最终答案。",
				RuntimeTool: "report_metrics_guard",
				Tags:        []string{"public-web", "guard", "adapter"},
			},
		},
		Evaluators: []agentxpack.Evaluator{
			{
				Name:        "financial_report_metrics_guard",
				Description: "检查主体、报告期、字段完整度、增长率、来源和 stop condition 是否满足公开财报指标答复要求。",
				OutputSchema: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"passed":                   map[string]any{"type": "boolean"},
						"subject_correct":          map[string]any{"type": "boolean"},
						"period_latest":            map[string]any{"type": "boolean"},
						"requested_fields_ready":   map[string]any{"type": "boolean"},
						"growth_fields_consistent": map[string]any{"type": "boolean"},
						"source_accepted":          map[string]any{"type": "boolean"},
						"field_sources_accepted":   map[string]any{"type": "boolean"},
						"stop_after_guard_passed":  map[string]any{"type": "boolean"},
						"missing_requested_fields": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
						"review_required_fields":   map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
						"failure_reason":           map[string]any{"type": "string"},
						"source_url":               map[string]any{"type": "string"},
						"report_period":            map[string]any{"type": "string"},
					},
					"required": []string{"passed", "requested_fields_ready", "source_accepted", "field_sources_accepted", "stop_after_guard_passed"},
				},
			},
		},
		EvalSuites: []agentxpack.EvalSuite{
			{
				Name:        "financial_report_metrics_success_suite",
				Description: "要求财报指标字段完整、来源可信、报告期正确，并在 guard 通过后停止。",
				Mode:        agentxpack.EvalSuiteModeGate,
				WorkflowIDs: []string{DefaultWorkflow},
				RequiredState: []string{
					"metrics.passed",
					"metrics.subject_correct",
					"metrics.period_latest",
					"metrics.requested_fields_ready",
					"metrics.growth_fields_consistent",
					"metrics.source_accepted",
					"metrics.field_sources_accepted",
					"metrics.stop_after_guard_passed",
					"metrics.guard_status",
					"metrics.report_period",
					"metrics.source_url",
				},
				PassPath:    "metrics.passed",
				SummaryPath: "metrics.summary",
				Default:     true,
			},
		},
		PolicyProfiles: []agentxpack.PolicyProfile{
			{
				Name: "public_web_report_metrics_readonly",
				Contract: agentxexecution.Contract{
					ID:      "public-web-report-metrics-readonly",
					Strict:  true,
					Version: 1,
					Visibility: agentxexecution.VisibilityPolicy{
						AllowTools: []string{
							"search",
							"open_page",
							"find_in_page",
							"web_fetch",
							"browser",
							"pdf",
							"report_metrics_candidates",
							"report_metrics_extract",
							"report_metrics_guard",
						},
						DeclaredTools: []string{
							"search",
							"open_page",
							"find_in_page",
							"web_fetch",
							"browser",
							"pdf",
							"report_metrics_candidates",
							"report_metrics_extract",
							"report_metrics_guard",
						},
						RequireDeclared: true,
						MaxRisk:         "medium",
					},
					Budget: agentxexecution.BudgetPolicy{
						MaxToolCalls: 8,
					},
					Loop: agentxexecution.LoopPolicy{
						MaxRounds:                8,
						LoopDetectionEnabled:     true,
						ToolFailureFuseEnabled:   true,
						ToolFailureFuseThreshold: 3,
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
				Name:        "financial_report_metrics_memory",
				Description: "沉淀公开财报指标任务中的主体、候选来源、字段、报告期、guard 结果和失败原因。",
				Default:     true,
				Schema: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"pack_id":          map[string]any{"type": "string"},
						"case_type":        map[string]any{"type": "string"},
						"workflow_id":      map[string]any{"type": "string"},
						"entity_name":      map[string]any{"type": "string"},
						"source_url":       map[string]any{"type": "string"},
						"adapter_id":       map[string]any{"type": "string"},
						"adapter_status":   map[string]any{"type": "string"},
						"source_policy":    map[string]any{"type": "string"},
						"report_period":    map[string]any{"type": "string"},
						"requested_fields": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
						"guard_status":     map[string]any{"type": "string"},
						"summary":          map[string]any{"type": "string"},
						"failure_code":     map[string]any{"type": "string"},
						"failure_reason":   map[string]any{"type": "string"},
					},
					"required": []string{"pack_id", "case_type", "workflow_id", "summary"},
				},
			},
		},
		MemoryRecallPolicy: &agentxpack.MemoryRecallPolicy{
			QueryHints: []string{
				"财报指标",
				"营收 净利润 增长率",
				"公开报告来源",
				"financial report metrics",
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
