package companyresearch

import (
	agentxpack "github.com/wsnacj/agentx-go/extensions/pack"
	agentxexecution "github.com/wsnacj/agentx-go/runtime/executionpolicy"
	agentxworkflow "github.com/wsnacj/agentx-go/runtime/workflow"
)

const (
	CaseTypeSingle      = "company_research.single_company"
	CaseTypeCompare     = "company_research.multi_company_compare"
	DefaultWorkflow     = WorkflowID
	CompareWorkflow     = "company_compare_lookup_v1"
	DefaultPolicy       = "company_research_readonly"
	DefaultEvaluator    = "company_research_evidence_guard"
	DefaultSuccessSuite = "company_research_success_suite"
)

func singleCompanyCaseSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"user_message": map[string]any{"type": "string"},
			"subject": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name":        map[string]any{"type": "string"},
					"market_hint": map[string]any{"type": "string"},
					"mentions":    stringArraySchema("Company aliases, display names, or tickers from the user request."),
				},
				"required": []string{"name"},
			},
			"requested_dimensions": stringArraySchema("Required evidence dimensions such as financials, market_data, news, risk, valuation, announcements, or research."),
			"requested_outputs":    stringArraySchema("Requested answer products such as brief, risk_summary, evidence_table, or investment_boundary."),
			"freshness":            map[string]any{"type": "object"},
			"risk_scope":           map[string]any{"type": "string"},
			"source_policy":        map[string]any{"type": "string"},
			"stop_condition":       map[string]any{"type": "string"},
		},
		"required": []string{"user_message", "subject", "requested_dimensions"},
	}
}

func compareCaseSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"user_message": map[string]any{"type": "string"},
			"comparison_subjects": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"entity_name":     map[string]any{"type": "string"},
						"market_hint":     map[string]any{"type": "string"},
						"entity_mentions": stringArraySchema("Aliases or tickers for this subject."),
					},
					"required": []string{"entity_name"},
				},
			},
			"requested_dimensions": stringArraySchema("Required evidence dimensions across all subjects."),
			"requested_outputs":    stringArraySchema("Requested answer products such as comparison, evidence_table, risk_summary, or investment_boundary."),
			"freshness":            map[string]any{"type": "object"},
			"risk_scope":           map[string]any{"type": "string"},
			"source_policy":        map[string]any{"type": "string"},
			"stop_condition":       map[string]any{"type": "string"},
		},
		"required": []string{"user_message", "comparison_subjects", "requested_dimensions"},
	}
}

func Definition() agentxpack.Definition {
	return agentxpack.Definition{
		Manifest: agentxpack.Manifest{
			ID:      PackID,
			Version: "0.1.0",
			Domain:  "company_research",
			RouteHints: []string{
				"company research",
				"investment risk review",
				"company comparison",
				"公司研究",
				"公司对比",
				"业绩 风险 估值 新闻",
				"财报 港股行情和相关新闻",
				"港股行情和相关新闻",
				"财报 当前股价估值和最近新闻",
				"财报 当前港股估值和最近新闻风险",
				"财报 港股估值 新闻风险",
				"财报 估值 新闻风险 边界评估",
				"股价估值和最近新闻",
				"财报 市场数据 新闻",
				"这家公司 最近新闻 估值判断",
				"新闻会不会改变前面的估值判断",
				"does recent news change the previous valuation",
			},
			SupportedCaseTypes: []string{CaseTypeSingle, CaseTypeCompare},
			DefaultWorkflow:    DefaultWorkflow,
			PolicyProfiles:     []string{DefaultPolicy},
			ArtifactTypes:      []string{"company_research.evidence"},
			Evaluators:         []string{DefaultEvaluator},
			EvalSuites:         []string{DefaultSuccessSuite},
		},
		CaseSchemas: []agentxpack.CaseSchema{
			{
				CaseType:    CaseTypeSingle,
				Description: "组合财报、市场数据、公开新闻和风险边界，完成单公司研究。",
				RouteHints:  []string{"公司研究", "投资风险", "业绩 新闻 风险", "财报 港股行情和相关新闻", "港股行情和相关新闻", "财报 当前股价估值和最近新闻", "财报 当前港股估值和最近新闻风险", "财报 港股估值 新闻风险", "财报 估值 新闻风险 边界评估", "股价估值和最近新闻", "这家公司 最近新闻 估值判断", "新闻会不会改变前面的估值判断", "does recent news change the previous valuation", "company research"},
				Schema:      singleCompanyCaseSchema(),
			},
			{
				CaseType:    CaseTypeCompare,
				Description: "组合多个主体的财报、市场数据、公开新闻和风险边界，完成公司横向比较。",
				RouteHints:  []string{"公司对比", "同行比较", "peer comparison", "company compare"},
				Schema:      compareCaseSchema(),
			},
		},
		Workflows: []agentxworkflow.Spec{
			singleCompanyWorkflow(),
			compareWorkflow(),
		},
		Tools: []agentxpack.SemanticTool{
			{
				Name:        "lookup_company_research",
				Description: "高层单公司研究入口。宿主负责把 finance、stock、public-news 等下游证据接入并输出 readiness/answer_contract。",
				RuntimeTool: ToolCompanyResearchLookup,
				Tags:        []string{"company-research", "lookup", "adapter"},
			},
			{
				Name:        "lookup_company_compare",
				Description: "高层多公司对比入口。宿主负责逐主体调用下游证据并输出 per-subject readiness。",
				RuntimeTool: ToolCompanyCompareLookup,
				Tags:        []string{"company-research", "compare", "adapter"},
			},
			{
				Name:        "guard_company_research",
				Description: "检查公司研究证据完整性、缺失维度和回答边界风险。",
				RuntimeTool: ToolCompanyResearchGuard,
				Tags:        []string{"company-research", "guard", "adapter"},
			},
		},
		Evaluators: []agentxpack.Evaluator{
			{
				Name:         DefaultEvaluator,
				Description:  "检查公司研究是否主体正确、证据完整、来源化、freshness 可解释，并避免过度投资结论。",
				OutputSchema: evaluationOutputSchema(),
			},
		},
		EvalSuites: []agentxpack.EvalSuite{
			{
				Name:        DefaultSuccessSuite,
				Description: "要求公司研究完成高层 lookup，输出 readiness/guard 状态，并以 bounded pass path 判断可答状态。",
				Mode:        agentxpack.EvalSuiteModeGate,
				WorkflowIDs: []string{DefaultWorkflow, CompareWorkflow},
				RequiredState: []string{
					"research.answer_ready",
					"research.guard_status",
					"research.passed",
				},
				PassPath:    "research.passed",
				SummaryPath: "research.summary",
				Default:     true,
			},
		},
		PolicyProfiles: []agentxpack.PolicyProfile{policyProfile()},
		MemorySchemas: []agentxpack.MemorySchema{
			{
				Name:        "company_research_memory",
				Description: "沉淀公司研究中的主体、维度、来源 readiness、边界状态和摘要。",
				Default:     true,
				Schema: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"pack_id":              map[string]any{"type": "string"},
						"case_type":            map[string]any{"type": "string"},
						"workflow_id":          map[string]any{"type": "string"},
						"entity_name":          map[string]any{"type": "string"},
						"subject_count":        map[string]any{"type": "integer"},
						"requested_dimensions": stringArraySchema("Requested evidence dimensions."),
						"ready_dimensions":     stringArraySchema("Ready evidence dimensions."),
						"missing_dimensions":   stringArraySchema("Missing evidence dimensions."),
						"guard_status":         map[string]any{"type": "string"},
						"summary":              map[string]any{"type": "string"},
						"failure_reason":       map[string]any{"type": "string"},
					},
					"required": []string{"pack_id", "case_type", "workflow_id", "summary"},
				},
			},
		},
		MemoryRecallPolicy: &agentxpack.MemoryRecallPolicy{
			QueryHints: []string{"公司研究", "投资风险", "财报 股价 新闻", "company research"},
			Limit:      4,
			MaxChars:   1600,
			ScopedOnly: true,
		},
	}
}

func singleCompanyWorkflow() agentxworkflow.Spec {
	return agentxworkflow.Spec{
		ID:              DefaultWorkflow,
		Title:           "Company Research Lookup",
		Description:     "调用高层 company_research_lookup，组合宿主注入的财报、市场数据和新闻证据；workflow 只表达结构化 case 到高层工具的 bounded path。",
		Version:         "v1",
		Pack:            PackID,
		CaseTypes:       []string{CaseTypeSingle},
		RouteHints:      []string{"公司研究", "业绩 风险 新闻", "财报 港股行情和相关新闻", "港股行情和相关新闻", "财报 当前股价估值和最近新闻", "财报 当前港股估值和最近新闻风险", "财报 港股估值 新闻风险", "财报 估值 新闻风险 边界评估", "股价估值和最近新闻", "这家公司 最近新闻 估值判断", "新闻会不会改变前面的估值判断", "does recent news change the previous valuation", "company research"},
		PlanningMode:    agentxworkflow.PlanningBounded,
		EntryNode:       "lookup_company",
		DefaultContract: DefaultPolicy,
		StateSchema:     workflowStateSchema(),
		EvaluatorSchema: []agentxworkflow.EvaluatorRef{{Name: DefaultEvaluator}},
		Nodes: []agentxworkflow.NodeSpec{
			{
				ID:          "lookup_company",
				Kind:        agentxworkflow.NodeTool,
				Title:       "Lookup company research evidence",
				Description: "通过 company_research_lookup 调用宿主接入的组合证据入口。",
				Inputs: []agentxworkflow.BindingSpec{
					{From: "case.input.user_message", To: "args.user_message"},
					{From: "case.input.subject.name", To: "args.entity_name"},
					{From: "case.input.requested_dimensions", To: "args.requested_dimensions"},
				},
				Outputs: companyLookupOutputs(),
				Config: map[string]any{
					"tool_name": "lookup_company_research",
					"args": map[string]any{
						"task_kind":      "company_research",
						"source_policy":  "public_company_research_sources",
						"stop_condition": "answer_ready_or_bounded_answer_contract",
					},
				},
			},
		},
	}
}

func compareWorkflow() agentxworkflow.Spec {
	return agentxworkflow.Spec{
		ID:              CompareWorkflow,
		Title:           "Company Compare Lookup",
		Description:     "调用高层 company_compare_lookup，按主体组合财报、市场数据和新闻证据并输出 per-subject readiness。",
		Version:         "v1",
		Pack:            PackID,
		CaseTypes:       []string{CaseTypeCompare},
		RouteHints:      []string{"公司对比", "同行比较", "company compare"},
		PlanningMode:    agentxworkflow.PlanningBounded,
		EntryNode:       "lookup_compare",
		DefaultContract: DefaultPolicy,
		StateSchema:     workflowStateSchema(),
		EvaluatorSchema: []agentxworkflow.EvaluatorRef{{Name: DefaultEvaluator}},
		Nodes: []agentxworkflow.NodeSpec{
			{
				ID:          "lookup_compare",
				Kind:        agentxworkflow.NodeTool,
				Title:       "Lookup company comparison evidence",
				Description: "通过 company_compare_lookup 调用宿主接入的多主体组合证据入口。",
				Inputs: []agentxworkflow.BindingSpec{
					{From: "case.input.user_message", To: "args.user_message"},
					{From: "case.input.comparison_subjects", To: "args.comparison_subjects"},
					{From: "case.input.requested_dimensions", To: "args.requested_dimensions"},
				},
				Outputs: companyLookupOutputs(),
				Config: map[string]any{
					"tool_name": "lookup_company_compare",
					"args": map[string]any{
						"task_kind":      "company_compare",
						"source_policy":  "public_company_research_sources",
						"stop_condition": "answer_ready_or_bounded_answer_contract",
					},
				},
			},
		},
	}
}

func workflowStateSchema() []agentxworkflow.StateSlotSpec {
	return []agentxworkflow.StateSlotSpec{
		{Name: "research.answer_ready", Type: "boolean", Required: true},
		{Name: "research.guard_status", Type: "string", Required: true},
		{Name: "research.adapter_status", Type: "string"},
		{Name: "research.failure_code", Type: "string"},
		{Name: "research.source_backed", Type: "boolean"},
		{Name: "research.boundary_ok", Type: "boolean"},
		{Name: "research.answer_contract_recommended", Type: "boolean"},
		{Name: "research.ready_dimensions", Type: "array"},
		{Name: "research.missing_dimensions", Type: "array"},
		{Name: "research.passed", Type: "boolean", Required: true},
		{Name: "research.summary", Type: "string"},
	}
}

func companyLookupOutputs() []agentxworkflow.BindingSpec {
	return []agentxworkflow.BindingSpec{
		{From: "result.answer_readiness.answer_ready", To: "state.research.answer_ready"},
		{From: "result.guard_status", To: "state.research.guard_status"},
		{From: "result.adapter_status", To: "state.research.adapter_status"},
		{From: "result.answer_readiness.safe_to_answer", To: "state.research.passed"},
		{From: "result.answer_readiness.allowed_summary_scope", To: "state.research.summary"},
	}
}

func evaluationOutputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"passed":                         map[string]any{"type": "boolean"},
			"subject_correct":                map[string]any{"type": "boolean"},
			"evidence_complete":              map[string]any{"type": "boolean"},
			"source_backed":                  map[string]any{"type": "boolean"},
			"freshness_confirmed":            map[string]any{"type": "boolean"},
			"answer_ready":                   map[string]any{"type": "boolean"},
			"guard_passed":                   map[string]any{"type": "boolean"},
			"boundary_ok":                    map[string]any{"type": "boolean"},
			"over_claim_detected":            map[string]any{"type": "boolean"},
			"answer_contract_recommended":    map[string]any{"type": "boolean"},
			"final_answer_boundary_observed": map[string]any{"type": "boolean"},
			"task_conflict_free":             map[string]any{"type": "boolean"},
			"subject_resolution_drift":       map[string]any{"type": "boolean"},
			"unguarded_synthesis_detected":   map[string]any{"type": "boolean"},
			"ready_dimensions":               stringArraySchema("Ready dimensions."),
			"missing_dimensions":             stringArraySchema("Missing dimensions."),
			"failure_reason":                 map[string]any{"type": "string"},
		},
		"required": []string{"passed", "subject_correct", "evidence_complete", "source_backed", "answer_ready", "boundary_ok"},
	}
}

func policyProfile() agentxpack.PolicyProfile {
	tools := []string{
		ToolCompanyResearchLookup,
		ToolCompanyCompareLookup,
		ToolCompanyResearchGuard,
		"finance_report_lookup",
		"a_stock_investigation",
		"global_stock_investigation",
		"latest_news_lookup",
	}
	return agentxpack.PolicyProfile{
		Name: DefaultPolicy,
		Contract: agentxexecution.Contract{
			ID:      "company-research-readonly",
			Strict:  true,
			Version: 1,
			Visibility: agentxexecution.VisibilityPolicy{
				AllowTools:      tools,
				DeclaredTools:   tools,
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
	}
}

func stringArraySchema(description string) map[string]any {
	return map[string]any{
		"type":        "array",
		"items":       map[string]any{"type": "string"},
		"description": description,
	}
}

func MaterializedDefaultWorkflow(coordinator *agentxpack.Coordinator) (agentxworkflow.Spec, error) {
	return coordinator.MaterializeWorkflow(Definition(), DefaultWorkflow)
}

func MaterializedCompareWorkflow(coordinator *agentxpack.Coordinator) (agentxworkflow.Spec, error) {
	return coordinator.MaterializeWorkflow(Definition(), CompareWorkflow)
}

func RegisterInto(reg agentxpack.Registry) error {
	if reg == nil {
		return nil
	}
	return reg.Register(Definition())
}
