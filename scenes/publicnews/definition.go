package publicnews

import (
	agentxpack "github.com/wsnacj/agentx-go/extensions/pack"
	agentxexecution "github.com/wsnacj/agentx-go/runtime/executionpolicy"
	agentxworkflow "github.com/wsnacj/agentx-go/runtime/workflow"
)

const (
	PackID              = "public-news-brief-pack"
	CaseTypeLatestBrief = "public_news.latest_brief"
	DefaultWorkflow     = "public_news_brief_lookup_v1"
)

func latestNewsBriefCaseSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"user_message": map[string]any{"type": "string"},
			"topic": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{"type": "string"},
					"entities": map[string]any{
						"type":  "array",
						"items": map[string]any{"type": "string"},
					},
				},
				"required": []string{"name"},
			},
			"requested_fields": map[string]any{
				"type":  "array",
				"items": map[string]any{"type": "string"},
			},
			"source_policy":      map[string]any{"type": "string"},
			"freshness":          map[string]any{"type": "string"},
			"cross_check_policy": map[string]any{"type": "string"},
			"stop_condition":     map[string]any{"type": "string"},
		},
		"required": []string{"user_message", "topic", "requested_fields"},
	}
}

func Definition() agentxpack.Definition {
	return agentxpack.Definition{
		Manifest: agentxpack.Manifest{
			ID:      PackID,
			Version: "0.1.0",
			Domain:  "public_news_brief",
			RouteHints: []string{
				"latest public news",
				"latest event update",
				"最新新闻",
				"最新进展",
				"要闻 简报",
			},
			SupportedCaseTypes: []string{CaseTypeLatestBrief},
			DefaultWorkflow:    DefaultWorkflow,
			PolicyProfiles:     []string{"public_web_news_brief_readonly"},
			ArtifactTypes:      []string{"public_web.news_brief.evidence"},
			Evaluators:         []string{"public_news_brief_guard"},
			EvalSuites:         []string{"public_news_brief_success_suite"},
		},
		CaseSchemas: []agentxpack.CaseSchema{
			{
				CaseType:    CaseTypeLatestBrief,
				Description: "从公开网页检索某个主题、事件或主体的最新新闻/最新进展，保留发布时间、关键进展和来源证据。",
				RouteHints: []string{
					"最新新闻",
					"最新进展",
					"latest news brief",
					"breaking update",
				},
				Schema: latestNewsBriefCaseSchema(),
			},
		},
		Workflows: []agentxworkflow.Spec{
			{
				ID:              DefaultWorkflow,
				Title:           "Public News Brief Lookup",
				Description:     "调用高层 latest_news_lookup 入口，由宿主注入搜索、打开页面、来源筛选、抽取和 guard；pack 只保留结构化 case/workflow/evaluator 合约。",
				Version:         "v1",
				Pack:            PackID,
				CaseTypes:       []string{CaseTypeLatestBrief},
				RouteHints:      []string{"最新新闻", "最新进展", "latest news"},
				PlanningMode:    agentxworkflow.PlanningBounded,
				EntryNode:       "lookup_news",
				DefaultContract: "public_web_news_brief_readonly",
				StateSchema: []agentxworkflow.StateSlotSpec{
					{Name: "news.news_fields_ready", Type: "boolean", Required: true},
					{Name: "news.guard_status", Type: "string", Required: true},
					{Name: "news.adapter_status", Type: "string"},
					{Name: "news.cross_check_ready", Type: "boolean", Required: true},
					{Name: "news.passed", Type: "boolean", Required: true},
					{Name: "news.freshness_confirmed", Type: "boolean", Required: true},
					{Name: "news.source_accepted", Type: "boolean", Required: true},
					{Name: "news.stop_after_guard_passed", Type: "boolean", Required: true},
					{Name: "news.missing_news_fields", Type: "array"},
					{Name: "news.review_reasons", Type: "array"},
					{Name: "news.failure_reason", Type: "string"},
					{Name: "news.source_url", Type: "string", Required: true},
					{Name: "news.published_at", Type: "string", Required: true},
					{Name: "news.summary", Type: "string", Required: true},
				},
				EvaluatorSchema: []agentxworkflow.EvaluatorRef{
					{Name: "public_news_brief_guard"},
				},
				Nodes: []agentxworkflow.NodeSpec{
					{
						ID:          "lookup_news",
						Kind:        agentxworkflow.NodeTool,
						Title:       "Lookup latest news brief",
						Description: "通过高层 latest_news_lookup 执行最新新闻检索、来源验证、正文抽取和 guard；具体 provider/source policy 由宿主实现。",
						Inputs: []agentxworkflow.BindingSpec{
							{From: "case.input.user_message", To: "args.user_message"},
							{From: "case.input.topic.name", To: "args.topic"},
							{From: "case.input.requested_fields", To: "args.requested_fields"},
						},
						Outputs: []agentxworkflow.BindingSpec{
							{From: "result.guard_status", To: "state.news.guard_status"},
							{From: "result.adapter_status", To: "state.news.adapter_status"},
							{From: "result.news_fields_ready", To: "state.news.news_fields_ready"},
							{From: "result.cross_check_ready", To: "state.news.cross_check_ready"},
							{From: "result.passed", To: "state.news.passed"},
							{From: "result.freshness_confirmed", To: "state.news.freshness_confirmed"},
							{From: "result.source_accepted", To: "state.news.source_accepted"},
							{From: "result.stop_after_guard_passed", To: "state.news.stop_after_guard_passed"},
							{From: "result.missing_news_fields", To: "state.news.missing_news_fields"},
							{From: "result.review_reasons", To: "state.news.review_reasons"},
							{From: "result.failure_code", To: "state.news.failure_reason"},
							{From: "result.source_url", To: "state.news.source_url"},
							{From: "result.published_at", To: "state.news.published_at"},
							{From: "result.summary", To: "state.news.summary"},
						},
						Config: map[string]any{
							"tool_name": "lookup_latest_news",
							"args": map[string]any{
								"task_kind":          "latest_news_brief",
								"requested_outputs":  []string{"brief", "source_verification"},
								"source_policy":      "public_web_prefer_official_or_authoritative_news_source",
								"cross_check_policy": "at_least_two_independent_source_sites_for_key_facts",
								"stop_condition":     "guard_passed",
							},
						},
					},
				},
			},
		},
		Tools: []agentxpack.SemanticTool{
			{
				Name:        "lookup_latest_news",
				Description: "高层最新新闻查询入口。模型提供结构化意图，宿主负责搜索、打开页面、来源筛选、抽取、交叉核对和 guard。",
				RuntimeTool: "latest_news_lookup",
				Tags:        []string{"public-web", "latest-news", "lookup", "adapter"},
			},
			{
				Name:        "search_news_candidates",
				Description: "检索公开新闻候选。搜索 provider、站点策略和排序仍由 runtime/project adapter 负责。",
				RuntimeTool: "search",
				Tags:        []string{"public-web", "candidate-generation", "latest-news"},
			},
			{
				Name:        "open_news_candidate",
				Description: "打开候选新闻页并保留 page_id/final_url/title。",
				RuntimeTool: "open_page",
				Tags:        []string{"public-web", "page-read"},
			},
			{
				Name:        "extract_latest_news",
				Description: "从页面正文或页面缓存中抽取最新新闻短简报候选字段。",
				RuntimeTool: "latest_news_extract",
				Tags:        []string{"public-web", "schema-extraction", "adapter"},
			},
			{
				Name:        "guard_latest_news",
				Description: "确认新闻字段、来源和交叉核对是否足够支撑短简报，并输出可停止信号。",
				RuntimeTool: "latest_news_guard",
				Tags:        []string{"public-web", "guard", "adapter"},
			},
		},
		Evaluators: []agentxpack.Evaluator{
			{
				Name:        "public_news_brief_guard",
				Description: "检查最新新闻短简报是否包含发布时间、关键进展、可信来源、交叉核对和 stop condition。",
				OutputSchema: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"passed":                  map[string]any{"type": "boolean"},
						"freshness_confirmed":     map[string]any{"type": "boolean"},
						"news_fields_ready":       map[string]any{"type": "boolean"},
						"source_accepted":         map[string]any{"type": "boolean"},
						"cross_check_ready":       map[string]any{"type": "boolean"},
						"stop_after_guard_passed": map[string]any{"type": "boolean"},
						"missing_news_fields":     map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
						"review_reasons":          map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
						"failure_reason":          map[string]any{"type": "string"},
						"source_url":              map[string]any{"type": "string"},
						"published_at":            map[string]any{"type": "string"},
					},
					"required": []string{"passed", "news_fields_ready", "source_accepted", "cross_check_ready", "stop_after_guard_passed"},
				},
			},
		},
		EvalSuites: []agentxpack.EvalSuite{
			{
				Name:        "public_news_brief_success_suite",
				Description: "要求最新新闻短简报字段完整、来源可信、完成交叉核对，并在 guard 通过后停止。",
				Mode:        agentxpack.EvalSuiteModeGate,
				WorkflowIDs: []string{DefaultWorkflow},
				RequiredState: []string{
					"news.passed",
					"news.freshness_confirmed",
					"news.news_fields_ready",
					"news.source_accepted",
					"news.cross_check_ready",
					"news.stop_after_guard_passed",
					"news.guard_status",
					"news.source_url",
					"news.published_at",
				},
				PassPath:    "news.passed",
				SummaryPath: "news.summary",
				Default:     true,
			},
		},
		PolicyProfiles: []agentxpack.PolicyProfile{
			{
				Name: "public_web_news_brief_readonly",
				Contract: agentxexecution.Contract{
					ID:      "public-web-news-brief-readonly",
					Strict:  true,
					Version: 1,
					Visibility: agentxexecution.VisibilityPolicy{
						AllowTools: []string{
							"latest_news_lookup",
							"search",
							"open_page",
							"find_in_page",
							"web_fetch",
							"browser",
							"latest_news_extract",
							"latest_news_guard",
						},
						DeclaredTools: []string{
							"latest_news_lookup",
							"search",
							"open_page",
							"find_in_page",
							"web_fetch",
							"browser",
							"latest_news_extract",
							"latest_news_guard",
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
				Name:        "public_news_brief_memory",
				Description: "沉淀公开新闻短简报中的主题、来源、发布时间、关键进展、guard 结果和失败原因。",
				Default:     true,
				Schema: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"pack_id":          map[string]any{"type": "string"},
						"case_type":        map[string]any{"type": "string"},
						"workflow_id":      map[string]any{"type": "string"},
						"topic_name":       map[string]any{"type": "string"},
						"source_url":       map[string]any{"type": "string"},
						"published_at":     map[string]any{"type": "string"},
						"source_policy":    map[string]any{"type": "string"},
						"requested_fields": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
						"guard_status":     map[string]any{"type": "string"},
						"summary":          map[string]any{"type": "string"},
						"failure_reason":   map[string]any{"type": "string"},
					},
					"required": []string{"pack_id", "case_type", "workflow_id", "summary"},
				},
			},
		},
		MemoryRecallPolicy: &agentxpack.MemoryRecallPolicy{
			QueryHints: []string{
				"最新新闻",
				"最新进展",
				"公开来源",
				"latest news brief",
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
