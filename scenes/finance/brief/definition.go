package brief

import (
	agentxpack "github.com/wsnacj/agentx-go/extensions/pack"
	agentxexecution "github.com/wsnacj/agentx-go/runtime/executionpolicy"
	agentxworkflow "github.com/wsnacj/agentx-go/runtime/workflow"
)

const (
	PackID          = "financial-report-brief-pack"
	CaseTypeBrief   = "financial_report.brief"
	DefaultWorkflow = "financial_report_brief_v1"
)

func briefCaseSchema() map[string]any {
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
			"brief_focus": map[string]any{
				"type":  "array",
				"items": map[string]any{"type": "string"},
			},
			"period_policy":  map[string]any{"type": "string"},
			"source_policy":  map[string]any{"type": "string"},
			"freshness":      map[string]any{"type": "string"},
			"output_style":   map[string]any{"type": "string"},
			"stop_condition": map[string]any{"type": "string"},
		},
		"required": []string{"user_message", "entity", "brief_focus"},
	}
}

func Definition() agentxpack.Definition {
	return agentxpack.Definition{
		Manifest: agentxpack.Manifest{
			ID:      PackID,
			Version: "0.1.0",
			Domain:  "financial_report_brief",
			RouteHints: []string{
				"financial report brief",
				"annual report summary",
				"key information from annual report",
				"财报简报",
				"年报摘要",
				"财报关键信息",
			},
			SupportedCaseTypes: []string{CaseTypeBrief},
			DefaultWorkflow:    DefaultWorkflow,
			PolicyProfiles:     []string{"public_web_report_brief_readonly"},
			ArtifactTypes:      []string{"public_web.report_brief.evidence"},
			Evaluators:         []string{"financial_report_brief_guard"},
			EvalSuites:         []string{"financial_report_brief_success_suite"},
		},
		CaseSchemas: []agentxpack.CaseSchema{
			{
				CaseType:    CaseTypeBrief,
				Description: "从官方或可信公开财报 artifact 中提炼主体最新年报的关键财务、经营、风险/展望和股东回报信息，并生成简报。",
				RouteHints: []string{
					"财报简报",
					"年报关键信息",
					"annual report brief",
					"summarize annual report key points",
				},
				Schema: briefCaseSchema(),
			},
		},
		Workflows: []agentxworkflow.Spec{
			{
				ID:              DefaultWorkflow,
				Title:           "Financial Report Brief",
				Description:     "解析财报简报请求，获取官方年报 artifact，提取关键点，并用 guard 判断是否可交付简报。",
				Version:         "v1",
				Pack:            PackID,
				CaseTypes:       []string{CaseTypeBrief},
				RouteHints:      []string{"财报简报", "annual report summary", "key points"},
				PlanningMode:    agentxworkflow.PlanningBounded,
				EntryNode:       "extract_brief",
				DefaultContract: "public_web_report_brief_readonly",
				StateSchema: []agentxworkflow.StateSlotSpec{
					{Name: "brief.guard_status", Type: "string", Required: true},
					{Name: "brief.passed", Type: "boolean", Required: true},
					{Name: "brief.subject_correct", Type: "boolean", Required: true},
					{Name: "brief.period_latest", Type: "boolean", Required: true},
					{Name: "brief.source_accepted", Type: "boolean", Required: true},
					{Name: "brief.brief_ready", Type: "boolean", Required: true},
					{Name: "brief.key_points_ready", Type: "boolean", Required: true},
					{Name: "brief.financials_ready", Type: "boolean", Required: true},
					{Name: "brief.risk_or_outlook_ready", Type: "boolean", Required: true},
					{Name: "brief.stop_after_guard_passed", Type: "boolean", Required: true},
					{Name: "brief.failure_reason", Type: "string"},
					{Name: "brief.report_period", Type: "string", Required: true},
					{Name: "brief.source_url", Type: "string", Required: true},
					{Name: "brief.summary", Type: "string", Required: true},
				},
				EvaluatorSchema: []agentxworkflow.EvaluatorRef{
					{Name: "financial_report_brief_guard"},
				},
				Nodes: []agentxworkflow.NodeSpec{
					{
						ID:          "extract_brief",
						Kind:        agentxworkflow.NodeTool,
						Title:       "Extract report brief evidence",
						Description: "从官方年报 PDF、docparse 结果或项目 adapter 中提取财务、经营、风险/展望等简报证据。",
						Inputs: []agentxworkflow.BindingSpec{
							{From: "case.input.user_message", To: "args.user_message"},
							{From: "case.input.entity.name", To: "args.entity_name"},
							{From: "case.input.entity.identifiers.stock_code", To: "args.stock_code"},
							{From: "case.input.entity.identifiers.ticker", To: "args.ticker"},
							{From: "case.input.period_policy", To: "args.period_scope"},
							{From: "case.input.source_policy", To: "args.source_policy"},
							{From: "case.input.output_style", To: "args.output_style"},
						},
						Config: map[string]any{
							"tool_name": "extract_report_brief",
						},
					},
					{
						ID:          "guard_brief",
						Kind:        agentxworkflow.NodeTool,
						Title:       "Guard report brief",
						Description: "确认简报主体、报告期、来源、财务信息、经营/风险信息和 stop condition 是否足以支撑最终交付。",
						Inputs: []agentxworkflow.BindingSpec{
							{From: "case.input.user_message", To: "args.user_message"},
							{From: "case.input.entity.name", To: "args.entity_name"},
							{From: "case.input.entity.identifiers.stock_code", To: "args.stock_code"},
							{From: "case.input.output_style", To: "args.output_style"},
							{From: "node.extract_brief.result.evidence", To: "args.evidence"},
						},
						Outputs: []agentxworkflow.BindingSpec{
							{From: "result.guard_status", To: "state.brief.guard_status"},
							{From: "result.evaluation.passed", To: "state.brief.passed"},
							{From: "result.evaluation.subject_correct", To: "state.brief.subject_correct"},
							{From: "result.evaluation.period_latest", To: "state.brief.period_latest"},
							{From: "result.evaluation.source_accepted", To: "state.brief.source_accepted"},
							{From: "result.evaluation.brief_ready", To: "state.brief.brief_ready"},
							{From: "result.evaluation.key_points_ready", To: "state.brief.key_points_ready"},
							{From: "result.evaluation.financials_ready", To: "state.brief.financials_ready"},
							{From: "result.evaluation.risk_or_outlook_ready", To: "state.brief.risk_or_outlook_ready"},
							{From: "result.evaluation.stop_after_guard_passed", To: "state.brief.stop_after_guard_passed"},
							{From: "result.evaluation.failure_reason", To: "state.brief.failure_reason"},
							{From: "result.evidence.report_period", To: "state.brief.report_period"},
							{From: "result.evidence.source_url", To: "state.brief.source_url"},
							{From: "result.evidence.brief", To: "state.brief.summary"},
						},
						Config: map[string]any{
							"tool_name": "guard_report_brief",
						},
					},
				},
				Edges: []agentxworkflow.EdgeSpec{
					{From: "extract_brief", To: "guard_brief", On: "success"},
				},
			},
		},
		Tools: []agentxpack.SemanticTool{
			{
				Name:        "extract_report_brief",
				Description: "从官方年报 artifact 或 docparse 结果中提取财报简报证据。",
				RuntimeTool: "report_brief_extract",
				Tags:        []string{"public-web", "document-parse", "schema-extraction", "adapter"},
			},
			{
				Name:        "guard_report_brief",
				Description: "确认财报简报证据是否足够并输出可交付简报。",
				RuntimeTool: "report_brief_guard",
				Tags:        []string{"public-web", "guard", "adapter"},
			},
		},
		Evaluators: []agentxpack.Evaluator{
			{
				Name:        "financial_report_brief_guard",
				Description: "检查财报简报是否包含主体、报告期、可信来源、核心财务信息、经营/风险或展望信息和 stop condition。",
				OutputSchema: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"passed":                  map[string]any{"type": "boolean"},
						"subject_correct":         map[string]any{"type": "boolean"},
						"period_latest":           map[string]any{"type": "boolean"},
						"source_accepted":         map[string]any{"type": "boolean"},
						"brief_ready":             map[string]any{"type": "boolean"},
						"key_points_ready":        map[string]any{"type": "boolean"},
						"financials_ready":        map[string]any{"type": "boolean"},
						"risk_or_outlook_ready":   map[string]any{"type": "boolean"},
						"stop_after_guard_passed": map[string]any{"type": "boolean"},
						"failure_reason":          map[string]any{"type": "string"},
						"review_reasons":          map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
						"source_url":              map[string]any{"type": "string"},
						"report_period":           map[string]any{"type": "string"},
					},
					"required": []string{"passed", "brief_ready", "source_accepted", "stop_after_guard_passed"},
				},
			},
		},
		EvalSuites: []agentxpack.EvalSuite{
			{
				Name:        "financial_report_brief_success_suite",
				Description: "要求财报简报字段完整、来源可信、报告期正确，并在 guard 通过后停止。",
				Mode:        agentxpack.EvalSuiteModeGate,
				WorkflowIDs: []string{DefaultWorkflow},
				RequiredState: []string{
					"brief.passed",
					"brief.subject_correct",
					"brief.period_latest",
					"brief.source_accepted",
					"brief.brief_ready",
					"brief.key_points_ready",
					"brief.financials_ready",
					"brief.risk_or_outlook_ready",
					"brief.stop_after_guard_passed",
					"brief.guard_status",
					"brief.report_period",
					"brief.source_url",
				},
				PassPath:    "brief.passed",
				SummaryPath: "brief.summary",
				Default:     true,
			},
		},
		PolicyProfiles: []agentxpack.PolicyProfile{
			{
				Name: "public_web_report_brief_readonly",
				Contract: agentxexecution.Contract{
					ID:      "public-web-report-brief-readonly",
					Strict:  true,
					Version: 1,
					Visibility: agentxexecution.VisibilityPolicy{
						AllowTools: []string{
							"search",
							"open_page",
							"web_fetch",
							"browser",
							"pdf",
							"document_parse",
							"report_brief_extract",
							"report_brief_guard",
						},
						DeclaredTools: []string{
							"search",
							"open_page",
							"web_fetch",
							"browser",
							"pdf",
							"document_parse",
							"report_brief_extract",
							"report_brief_guard",
						},
						RequireDeclared: true,
						MaxRisk:         "medium",
					},
					Budget: agentxexecution.BudgetPolicy{MaxToolCalls: 8},
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
				Name:        "financial_report_brief_memory",
				Description: "沉淀公开财报简报任务中的主体、官方来源、报告期、关键点、guard 结果和失败原因。",
				Default:     true,
				Schema: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"pack_id":        map[string]any{"type": "string"},
						"case_type":      map[string]any{"type": "string"},
						"workflow_id":    map[string]any{"type": "string"},
						"entity_name":    map[string]any{"type": "string"},
						"source_url":     map[string]any{"type": "string"},
						"report_period":  map[string]any{"type": "string"},
						"guard_status":   map[string]any{"type": "string"},
						"brief":          map[string]any{"type": "string"},
						"failure_reason": map[string]any{"type": "string"},
					},
					"required": []string{"pack_id", "case_type", "workflow_id", "brief"},
				},
			},
		},
		MemoryRecallPolicy: &agentxpack.MemoryRecallPolicy{
			QueryHints: []string{"财报简报", "年报摘要", "annual report brief", "financial report key points"},
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
