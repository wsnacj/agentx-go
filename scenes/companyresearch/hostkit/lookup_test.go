package hostkit

import (
	"context"
	"errors"
	"strings"
	"testing"

	research "github.com/wsnacj/agentx-go/scenes/companyresearch"
)

func TestBuildCompanyResearchLookupPayloadCallsStructuredDimensions(t *testing.T) {
	calls := []string{}
	cfg := CompanyResearchConfig{
		Source: "test_hostkit",
		Handlers: CompanyResearchHandlers{
			Finance: func(context.Context, map[string]any) (any, error) {
				calls = append(calls, "finance")
				return map[string]any{"tool": "finance_report_lookup", "adapter_status": "ok"}, nil
			},
			AStock: func(context.Context, map[string]any) (any, error) {
				calls = append(calls, "a_stock")
				return map[string]any{"tool": "a_stock_investigation", "adapter_status": "ok"}, nil
			},
			GlobalStock: func(context.Context, map[string]any) (any, error) {
				calls = append(calls, "global_stock")
				return map[string]any{"tool": "global_stock_investigation", "adapter_status": "ok"}, nil
			},
			News: func(context.Context, map[string]any) (any, error) {
				calls = append(calls, "news")
				return map[string]any{"tool": "latest_news_lookup", "adapter_status": "ok"}, nil
			},
		},
	}
	payload, err := BuildCompanyResearchLookupPayload(context.Background(), cfg, map[string]any{
		"user_message":         "看下腾讯的财报、股价和新闻",
		"entity_name":          "腾讯",
		"requested_dimensions": []any{"financials", "market_data", "news"},
		"market_hint":          "HK",
	})
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if payload.Tool != research.ToolCompanyResearchLookup || !payload.AnswerReadiness.AnswerReady {
		t.Fatalf("unexpected payload: %#v", payload)
	}
	if payload.TaskPlan == nil {
		t.Fatalf("expected task plan in payload")
	}
	for _, role := range []research.CompanyResearchTaskRole{
		research.CompanyResearchRoleSubjectResolver,
		research.CompanyResearchRoleFinanceAnalyst,
		research.CompanyResearchRoleMarketAnalyst,
		research.CompanyResearchRoleNewsAnalyst,
		research.CompanyResearchRoleEvidenceGuard,
		research.CompanyResearchRoleSynthesisEditor,
	} {
		if !payloadHasTaskResult(payload, role) {
			t.Fatalf("expected task result for role %s in %#v", role, payload.TaskResults)
		}
	}
	if !payload.AnswerReadiness.SafeToAnswer {
		t.Fatalf("expected ready lookup to be safe to answer, got %#v", payload.AnswerReadiness)
	}
	if payload.AnswerContract == nil || !payload.AnswerContract.FinalAnswerRecommended {
		t.Fatalf("expected ready lookup to produce final answer contract, got %#v", payload.AnswerContract)
	}
	if draft := payload.AnswerContract.FinalAnswerDraft; !strings.Contains(draft, "公司研究已完成") ||
		!strings.Contains(draft, "已通过证据维度") ||
		!strings.Contains(draft, "缺失或未通过 guard 的维度：无") ||
		!strings.Contains(draft, "可能影响：") ||
		!strings.Contains(draft, "风险边界：") ||
		!strings.Contains(draft, "不构成投资建议") {
		t.Fatalf("expected ready answer contract draft, got %q", draft)
	}
	if len(calls) != 3 || calls[0] != "finance" || calls[1] != "global_stock" || calls[2] != "news" {
		t.Fatalf("unexpected calls: %#v", calls)
	}
}

func payloadHasTaskResult(payload research.CompanyResearchPayload, role research.CompanyResearchTaskRole) bool {
	for _, result := range payload.TaskResults {
		if result.Role == role {
			return true
		}
	}
	return false
}

func TestReadyCompanyResearchAnswerContractCarriesNewsAssessmentBoundary(t *testing.T) {
	payload := research.CompanyResearchPayload{
		Intent: research.CompanyResearchIntent{EntityName: "Example Tech"},
		AnswerReadiness: research.CompanyResearchAnswerReadiness{
			AnswerReady:     true,
			SafeToAnswer:    true,
			ReadyDimensions: []string{"market_data", "news"},
		},
		Evidence: research.CompanyResearchEvidence{
			GlobalStock: map[string]any{
				"adapter_status": "ok",
				"quote": map[string]any{
					"subject": map[string]any{
						"entity_name": "Example Tech",
						"stock_code":  "06657",
						"market":      "hk",
					},
					"quote": map[string]any{
						"price": map[string]any{"value": "13.50", "unit": "HKD"},
					},
				},
			},
			News: map[string]any{
				"adapter_status": "ok",
				"guard_status":   "passed",
				"passed":         true,
				"summary":        "Example Tech released a source-backed update.",
				"answer_contract": map[string]any{
					"possible_impact": "The evidence confirms the update but does not quantify business impact.",
					"risk_boundary":   "The assessment is limited to opened and guarded public sources.",
				},
			},
		},
	}
	contract := readyCompanyResearchAnswerContract(payload)
	for _, expected := range []string{
		"主体：Example Tech（06657.HK）。",
		"行情/估值：Example Tech (06657.HK)，价格 13.50HKD",
		"新闻/风险：Example Tech released a source-backed update.",
		"新闻可能影响：The evidence confirms the update but does not quantify business impact.",
		"新闻风险边界：The assessment is limited to opened and guarded public sources.",
	} {
		if !strings.Contains(contract.FinalAnswerDraft, expected) {
			t.Fatalf("expected ready company-research draft to contain %q, got %q", expected, contract.FinalAnswerDraft)
		}
	}
	payload.AnswerReadiness = research.CompanyResearchAnswerReadiness{
		Degraded:          true,
		SafeToAnswer:      true,
		ReadyDimensions:   []string{"market_data"},
		MissingDimensions: []string{"news"},
		FailureClass:      "evidence_weak",
	}
	degraded := CompanyResearchAnswerContract(payload)
	if degraded == nil || !strings.Contains(degraded.FinalAnswerDraft, "主体：Example Tech（06657.HK）。") {
		t.Fatalf("expected degraded contract to preserve verified security identity, got %#v", degraded)
	}
	if degraded.PossibleImpact == "" || degraded.RiskBoundary == "" ||
		!strings.Contains(degraded.FinalAnswerDraft, "可能影响：") || !strings.Contains(degraded.FinalAnswerDraft, "风险边界：") {
		t.Fatalf("expected degraded impact and risk boundary, got %#v", degraded)
	}
}

func TestReadyCompanyResearchAnswerContractDirectlyAnswersNewsValuationRelationship(t *testing.T) {
	payload := research.CompanyResearchPayload{
		Intent: research.CompanyResearchIntent{
			UserMessage:    "那这家公司最近新闻会不会改变前面的估值判断？",
			OriginalIntent: "那这家公司最近新闻会不会改变前面的估值判断？",
			EntityName:     "Example Tech",
		},
		AnswerReadiness: research.CompanyResearchAnswerReadiness{
			AnswerReady:     true,
			SafeToAnswer:    true,
			ReadyDimensions: []string{"market_data", "news"},
		},
	}

	contract := readyCompanyResearchAnswerContract(payload)
	if !strings.Contains(contract.FinalAnswerDraft, "估值影响判断：现有证据不足以认定该新闻已经改变前述估值判断。") {
		t.Fatalf("expected direct news-to-valuation judgment, got %q", contract.FinalAnswerDraft)
	}
	if !strings.Contains(contract.PossibleImpact, "缺少经验证的盈利预测、现金流修订或估值倍数传导证据") {
		t.Fatalf("expected valuation transmission boundary, got %#v", contract)
	}
}

func TestReadyCompanyResearchAnswerContractDoesNotInferNewsValuationRelationship(t *testing.T) {
	payload := research.CompanyResearchPayload{
		Intent: research.CompanyResearchIntent{
			UserMessage: "查询最新新闻和当前估值，并说明各维度 readiness。",
			EntityName:  "Example Tech",
		},
		AnswerReadiness: research.CompanyResearchAnswerReadiness{
			AnswerReady:     true,
			SafeToAnswer:    true,
			ReadyDimensions: []string{"market_data", "news"},
		},
	}

	contract := readyCompanyResearchAnswerContract(payload)
	if strings.Contains(contract.FinalAnswerDraft, "估值影响判断：") {
		t.Fatalf("expected independent news and valuation request to keep generic contract, got %q", contract.FinalAnswerDraft)
	}
	if contract.PossibleImpact != "已通过证据可用于描述当前经营、市场与公开新闻状态，但不能单独证明未来影响或投资结果。" {
		t.Fatalf("expected generic impact boundary, got %#v", contract)
	}
}

func payloadTaskResult(payload research.CompanyResearchPayload, role research.CompanyResearchTaskRole) (research.CompanyResearchTaskResult, bool) {
	for _, result := range payload.TaskResults {
		if result.Role == role {
			return result, true
		}
	}
	return research.CompanyResearchTaskResult{}, false
}

func TestBuildCompanyResearchLookupPayloadTaskExecutorCanHandleFinanceTask(t *testing.T) {
	financeHandlerCalls := 0
	cfg := CompanyResearchConfig{
		TaskExecutor: func(_ context.Context, request CompanyResearchTaskExecutionRequest) (CompanyResearchTaskExecutionResult, error) {
			if request.Task.Role != research.CompanyResearchRoleFinanceAnalyst {
				return CompanyResearchTaskExecutionResult{}, nil
			}
			return CompanyResearchTaskExecutionResult{
				Handled: true,
				Evidence: map[string]any{
					"tool":           "company_research_finance_task",
					"adapter_status": "ok",
					"summary":        "task executor finance evidence",
				},
				TaskResult: &research.CompanyResearchTaskResult{
					Status:        research.CompanyResearchTaskStatusReady,
					EvidenceReady: true,
					ExecutorID:    "test_task_executor",
					Summary:       "finance task handled by host executor",
					Diagnostics: map[string]string{
						"duration_ms": "12",
						"model":       "must_not_leak",
						"provider":    "must_not_leak",
					},
				},
			}, nil
		},
		Handlers: CompanyResearchHandlers{
			Finance: func(context.Context, map[string]any) (any, error) {
				financeHandlerCalls++
				return map[string]any{"tool": "finance_report_lookup", "adapter_status": "ok"}, nil
			},
		},
	}
	payload, err := BuildCompanyResearchLookupPayload(context.Background(), cfg, map[string]any{
		"user_message":         "帮我看下富途牛牛最新财报。",
		"entity_name":          "富途牛牛",
		"requested_dimensions": []any{"financials"},
	})
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if financeHandlerCalls != 0 {
		t.Fatalf("expected host task executor to handle finance without fallback handler, calls=%d", financeHandlerCalls)
	}
	if !payload.AnswerReadiness.AnswerReady {
		t.Fatalf("expected finance-only task executor evidence to be ready, got %#v", payload.AnswerReadiness)
	}
	if payload.Evidence.Finance["summary"] != "task executor finance evidence" {
		t.Fatalf("expected task executor evidence, got %#v", payload.Evidence.Finance)
	}
	result, ok := payloadTaskResult(payload, research.CompanyResearchRoleFinanceAnalyst)
	if !ok || result.ExecutorID != "test_task_executor" || result.Status != research.CompanyResearchTaskStatusReady {
		t.Fatalf("expected task executor result, got %#v ok=%v", result, ok)
	}
	if result.Diagnostics["duration_ms"] != "12" || result.Diagnostics["model"] != "" || result.Diagnostics["provider"] != "" {
		t.Fatalf("expected sanitized task executor diagnostics, got %#v", result.Diagnostics)
	}
}

func TestBuildCompanyResearchLookupPayloadTaskExecutorErrorFallsBackToHandler(t *testing.T) {
	newsHandlerCalls := 0
	cfg := CompanyResearchConfig{
		TaskExecutor: func(_ context.Context, request CompanyResearchTaskExecutionRequest) (CompanyResearchTaskExecutionResult, error) {
			if request.Task.Role != research.CompanyResearchRoleNewsAnalyst {
				return CompanyResearchTaskExecutionResult{}, nil
			}
			return CompanyResearchTaskExecutionResult{}, errors.New("temporary task runtime failure")
		},
		Handlers: CompanyResearchHandlers{
			News: func(context.Context, map[string]any) (any, error) {
				newsHandlerCalls++
				return map[string]any{"tool": "latest_news_lookup", "adapter_status": "ok"}, nil
			},
		},
	}
	payload, err := BuildCompanyResearchLookupPayload(context.Background(), cfg, map[string]any{
		"user_message":         "帮我看下富途牛牛最近新闻风险。",
		"entity_name":          "富途牛牛",
		"requested_dimensions": []any{"news"},
	})
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if newsHandlerCalls != 1 {
		t.Fatalf("expected news fallback handler to run once, got %d", newsHandlerCalls)
	}
	if !payload.AnswerReadiness.AnswerReady {
		t.Fatalf("expected fallback news evidence to be ready, got %#v", payload.AnswerReadiness)
	}
	if !containsWarning(payload.Warnings, "task_executor_error:news_analyst") {
		t.Fatalf("expected task executor warning, got %#v", payload.Warnings)
	}
	result, ok := payloadTaskResult(payload, research.CompanyResearchRoleNewsAnalyst)
	if !ok || result.Status != research.CompanyResearchTaskStatusReady || result.Diagnostics["task_executor_failure_code"] != "task_executor_error" {
		t.Fatalf("expected ready news result with task executor diagnostics, got %#v ok=%v", result, ok)
	}
}

func TestBuildCompanyResearchLookupPayloadDegradesWhenNewsGuardNeedsReview(t *testing.T) {
	cfg := CompanyResearchConfig{
		Handlers: CompanyResearchHandlers{
			News: func(context.Context, map[string]any) (any, error) {
				return map[string]any{
					"tool":           "latest_news_lookup",
					"adapter_status": "ok",
					"guard_status":   "needs_cross_check",
					"passed":         false,
					"source_url":     "https://news.example.com/noisy",
					"answer_contract": map[string]any{
						"final_answer_recommended": true,
						"reason":                   "source_quality_needs_review",
					},
				}, nil
			},
		},
	}
	payload, err := BuildCompanyResearchLookupPayload(context.Background(), cfg, map[string]any{
		"user_message":         "查下样例公司最近新闻风险",
		"entity_name":          "样例公司",
		"requested_dimensions": []any{"news"},
	})
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if payload.AnswerReadiness.AnswerReady || !containsString(payload.AnswerReadiness.MissingDimensions, "news") {
		t.Fatalf("expected source-quality news guard to degrade news readiness, got %#v", payload.AnswerReadiness)
	}
	result, ok := payloadTaskResult(payload, research.CompanyResearchRoleNewsAnalyst)
	if !ok || result.Status != research.CompanyResearchTaskStatusDegraded || result.EvidenceReady {
		t.Fatalf("expected degraded news task result, got %#v ok=%v", result, ok)
	}
	if result.Diagnostics["news_quality_code"] == "" {
		t.Fatalf("expected news quality diagnostic, got %#v", result)
	}
	if payload.TaskSummary == nil || !containsRole(payload.TaskSummary.DegradedRoles, research.CompanyResearchRoleNewsAnalyst) {
		t.Fatalf("expected task summary to show degraded news role, got %#v", payload.TaskSummary)
	}
	if !strings.Contains(payload.AnswerContract.FinalAnswerDraft, "未完全就绪的任务角色：news_analyst、evidence_guard") ||
		strings.Contains(payload.AnswerContract.FinalAnswerDraft, "任务缺口摘要") {
		t.Fatalf("expected degraded roles to be described without implying missing execution, got %q", payload.AnswerContract.FinalAnswerDraft)
	}
}

func TestTaskSummaryContractLinesDistinguishesFailedRoles(t *testing.T) {
	lines := taskSummaryContractLines(research.CompanyResearchPayload{
		TaskSummary: &research.CompanyResearchTaskSummary{
			DegradedRoles: []research.CompanyResearchTaskRole{research.CompanyResearchRoleNewsAnalyst},
			FailedRoles:   []research.CompanyResearchTaskRole{research.CompanyResearchRoleFinanceAnalyst},
		},
	})
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "未完全就绪的任务角色：news_analyst") ||
		!strings.Contains(joined, "执行失败的任务角色：finance_analyst") {
		t.Fatalf("expected distinct degraded and failed role labels, got %q", joined)
	}
}

func TestBuildCompanyResearchLookupPayloadPropagatesNewsFailureClass(t *testing.T) {
	cfg := CompanyResearchConfig{
		Handlers: CompanyResearchHandlers{
			Finance: func(context.Context, map[string]any) (any, error) {
				return map[string]any{"tool": "finance_report_lookup", "adapter_status": "ok"}, nil
			},
			GlobalStock: func(context.Context, map[string]any) (any, error) {
				return map[string]any{"tool": "global_stock_investigation", "adapter_status": "ok"}, nil
			},
			News: func(context.Context, map[string]any) (any, error) {
				return map[string]any{
					"tool":           "latest_news_lookup",
					"adapter_status": "provider_execution_failed",
					"failure_code":   "search_provider_failure_subscription_token_invalid",
					"failure_class":  "config_invalid",
					"sources": map[string]any{
						"provider":            "brave",
						"provider_status":     "SUBSCRIPTION_TOKEN_INVALID",
						"failure_class":       "config_invalid",
						"retryable":           false,
						"retry_attempt_count": 1,
						"retry_exhausted":     false,
						"failure_code":        "search_provider_failure_subscription_token_invalid",
						"effective_provider":  "brave",
						"fallback_hint":       "replace BRAVE_API_KEY",
					},
				}, nil
			},
		},
	}
	payload, err := BuildCompanyResearchLookupPayload(context.Background(), cfg, map[string]any{
		"user_message":         "看下样例公司的财报、股价和新闻",
		"entity_name":          "样例公司",
		"market_hint":          "HK",
		"requested_dimensions": []any{"financials", "market_data", "news"},
	})
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if payload.AnswerReadiness.AnswerReady ||
		payload.AnswerReadiness.FailureClass != "config_invalid" ||
		payload.FailureClass != "config_invalid" ||
		!containsString(payload.AnswerReadiness.MissingDimensions, "news") {
		t.Fatalf("expected news config failure class to propagate, got readiness=%#v payload_failure=%q", payload.AnswerReadiness, payload.FailureClass)
	}
	result, ok := payloadTaskResult(payload, research.CompanyResearchRoleNewsAnalyst)
	if !ok ||
		result.FailureClass != "config_invalid" ||
		result.FailureCode != "search_provider_failure_subscription_token_invalid" {
		t.Fatalf("expected news task result failure class/code, got %#v ok=%v", result, ok)
	}
	if payload.AnswerContract == nil ||
		!strings.Contains(payload.AnswerContract.FinalAnswerDraft, "搜索 provider 不可用") ||
		!strings.Contains(payload.AnswerContract.FinalAnswerDraft, "search_provider_failure_subscription_token_invalid") {
		t.Fatalf("expected bounded answer contract to retain provider root cause, got %#v", payload.AnswerContract)
	}
}

func TestCompanyResearchTreatsQuotaExhaustionAsTerminalProviderFailure(t *testing.T) {
	if got := normalizeCompanyFailureClass("quota_limited"); got != "quota_limited" {
		t.Fatalf("expected quota failure class to remain stable, got %q", got)
	}
	if got := companyFailureClassFromCode("search provider quota exhausted"); got != "quota_limited" {
		t.Fatalf("expected quota failure code classification, got %q", got)
	}
	if companyResearchFailureClassRecoverable("quota_limited") {
		t.Fatal("quota exhaustion must not recommend immediate same-run recovery")
	}
}

func TestBuildCompanyComparePayloadExposesTaskPlanForPromptShapedComparison(t *testing.T) {
	cfg := CompanyResearchConfig{
		Handlers: CompanyResearchHandlers{
			Finance: func(context.Context, map[string]any) (any, error) {
				return map[string]any{"tool": "finance_report_lookup", "adapter_status": "ok"}, nil
			},
			GlobalStock: func(context.Context, map[string]any) (any, error) {
				return map[string]any{"tool": "global_stock_investigation", "adapter_status": "ok"}, nil
			},
			News: func(context.Context, map[string]any) (any, error) {
				return map[string]any{"tool": "latest_news_lookup", "adapter_status": "ok"}, nil
			},
		},
	}
	payload, err := BuildCompanyComparePayload(context.Background(), cfg, map[string]any{
		"user_message": "帮我比较腾讯控股和 Microsoft 最近财报表现、股价和新闻风险。",
		"comparison_subjects": []any{
			map[string]any{"entity_name": "腾讯控股", "market_hint": "HK"},
			map[string]any{"entity_name": "Microsoft", "market_hint": "US"},
		},
		"requested_dimensions": []any{"financials", "market_data", "news"},
	})
	if err != nil {
		t.Fatalf("compare: %v", err)
	}
	if payload.TaskPlan == nil || len(payload.TaskPlan.Tasks) == 0 {
		t.Fatalf("expected top-level task plan, got %#v", payload.TaskPlan)
	}
	if !payloadHasTaskResult(payload, research.CompanyResearchRoleEvidenceGuard) ||
		!payloadHasTaskResult(payload, research.CompanyResearchRoleSynthesisEditor) {
		t.Fatalf("expected top-level guard/synthesis task results, got %#v", payload.TaskResults)
	}
	if len(payload.Subjects) != 2 {
		t.Fatalf("expected two subject payloads, got %#v", payload.Subjects)
	}
	for _, subject := range payload.Subjects {
		if subject.TaskPlan == nil || !payloadHasTaskResult(subject, research.CompanyResearchRoleFinanceAnalyst) ||
			!payloadHasTaskResult(subject, research.CompanyResearchRoleMarketAnalyst) ||
			!payloadHasTaskResult(subject, research.CompanyResearchRoleNewsAnalyst) {
			t.Fatalf("expected per-subject task plan/results, got subject=%#v", subject)
		}
	}
}

func containsWarning(warnings []string, want string) bool {
	for _, warning := range warnings {
		if warning == want {
			return true
		}
	}
	return false
}

func TestBuildCompanyResearchLookupPayloadDegradesWhenDimensionsMissing(t *testing.T) {
	payload, err := BuildCompanyResearchLookupPayload(context.Background(), CompanyResearchConfig{}, map[string]any{
		"user_message":         "看下某公司的财报和新闻",
		"entity_name":          "某公司",
		"requested_dimensions": []any{"financials", "news"},
	})
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if payload.AnswerReadiness.AnswerReady || payload.AnswerReadiness.FailureCode != "company_research_missing_required_dimensions" {
		t.Fatalf("expected degraded readiness, got %#v", payload.AnswerReadiness)
	}
	if !payload.AnswerReadiness.SafeToAnswer {
		t.Fatalf("expected degraded lookup to be safe for bounded answer, got %#v", payload.AnswerReadiness)
	}
	if payload.AnswerContract == nil || !payload.AnswerContract.FinalAnswerRecommended {
		t.Fatalf("expected final answer contract for degraded lookup, got %#v", payload.AnswerContract)
	}
	if !strings.Contains(payload.AnswerContract.FinalAnswerDraft, "缺失或未通过 guard 的维度") ||
		!strings.Contains(payload.AnswerContract.FinalAnswerDraft, "financials") ||
		!strings.Contains(payload.AnswerContract.FinalAnswerDraft, "news") ||
		!strings.Contains(payload.AnswerContract.FinalAnswerDraft, "可能影响：") ||
		!strings.Contains(payload.AnswerContract.FinalAnswerDraft, "风险边界：") {
		t.Fatalf("expected bounded missing-dimension draft, got %q", payload.AnswerContract.FinalAnswerDraft)
	}
}

func TestBuildCompanyResearchLookupPayloadAnswerContractSummarizesReadyEvidence(t *testing.T) {
	cfg := CompanyResearchConfig{
		Handlers: CompanyResearchHandlers{
			Finance: func(context.Context, map[string]any) (any, error) {
				return map[string]any{
					"tool":           "finance_report_lookup",
					"adapter_status": "ok",
					"brief": map[string]any{
						"evidence": map[string]any{
							"brief": "样例公司最新报告显示收入 100 亿元，净利润 20 亿元。",
						},
					},
				}, nil
			},
			GlobalStock: func(context.Context, map[string]any) (any, error) {
				return map[string]any{
					"tool":           "global_stock_investigation",
					"adapter_status": "ok",
					"quote": map[string]any{
						"subject": map[string]any{
							"entity_name": "样例公司",
						},
						"quote": map[string]any{
							"price":  map[string]any{"value": "12.30", "unit": "HKD"},
							"pe_ttm": map[string]any{"value": "18.5"},
							"pb":     map[string]any{"value": "2.1"},
						},
						"freshness": map[string]any{"as_of": "2026-05-27 10:00:00"},
					},
				}, nil
			},
			News: func(context.Context, map[string]any) (any, error) {
				return map[string]any{"tool": "latest_news_lookup", "adapter_status": "unsupported", "failure_code": "missing_credentials"}, nil
			},
		},
	}
	payload, err := BuildCompanyResearchLookupPayload(context.Background(), cfg, map[string]any{
		"user_message":         "看下样例公司的财报、股价和新闻",
		"entity_name":          "样例公司",
		"market_hint":          "HK",
		"requested_dimensions": []any{"financials", "market_data", "news"},
	})
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if payload.AnswerContract == nil {
		t.Fatalf("expected final answer contract for degraded lookup")
	}
	draft := payload.AnswerContract.FinalAnswerDraft
	for _, want := range []string{
		"已通过证据摘要",
		"财务：样例公司最新报告显示收入 100 亿元，净利润 20 亿元。",
		"行情/估值：样例公司，价格 12.30HKD，PE(TTM) 18.5，PB 2.1",
		"缺失或未通过 guard 的维度：news",
		"不能把未 ready 的财报、行情、新闻或风险证据写成已核实事实",
	} {
		if !strings.Contains(draft, want) {
			t.Fatalf("expected draft to contain %q, got %q", want, draft)
		}
	}
}

func TestBuildCompanyResearchLookupPayloadAnswerContractExplainsIdentityFailure(t *testing.T) {
	cfg := CompanyResearchConfig{
		Handlers: CompanyResearchHandlers{
			Finance: func(context.Context, map[string]any) (any, error) {
				return map[string]any{
					"tool":           "finance_report_lookup",
					"adapter_status": "evidence_incomplete",
					"failure_code":   "requested_fields_missing",
					"candidates": map[string]any{
						"adapter_status": "no_candidate",
						"failure_code":   "eastmoney_identity_not_found",
					},
				}, nil
			},
			AStock: func(context.Context, map[string]any) (any, error) {
				return map[string]any{"tool": "a_stock_investigation", "adapter_status": "unavailable", "failure_code": "identity_not_found"}, nil
			},
			GlobalStock: func(context.Context, map[string]any) (any, error) {
				return map[string]any{"tool": "global_stock_investigation", "adapter_status": "unavailable", "failure_code": "identity_not_found"}, nil
			},
			News: func(context.Context, map[string]any) (any, error) {
				return map[string]any{"tool": "latest_news_lookup", "adapter_status": "provider_unavailable", "failure_code": "search_provider_failure_missing"}, nil
			},
		},
	}
	payload, err := BuildCompanyResearchLookupPayload(context.Background(), cfg, map[string]any{
		"user_message":         "帮我查下富途牛牛最新财报、当前股价估值和最近新闻风险。",
		"entity_name":          "富途牛牛",
		"requested_dimensions": []any{"financials", "market_data", "news"},
	})
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if payload.AnswerContract == nil {
		t.Fatalf("expected bounded answer contract")
	}
	draft := payload.AnswerContract.FinalAnswerDraft
	for _, want := range []string{
		"未通过原因摘要",
		"财务：下游财报解析器未确认主体身份（eastmoney_identity_not_found）",
		"不能把输入名称直接等同为上市主体",
		"行情/估值：A股和港美股下游解析器都未确认主体身份",
		"新闻/风险：搜索 provider 不可用（search_provider_failure_missing）",
	} {
		if !strings.Contains(draft, want) {
			t.Fatalf("expected draft to contain %q, got %q", want, draft)
		}
	}
}

func TestBuildCompanyResearchLookupPayloadHandsOffFinanceResolvedMarketIdentity(t *testing.T) {
	aStockCalls := 0
	globalParams := map[string]any{}
	cfg := CompanyResearchConfig{
		Handlers: CompanyResearchHandlers{
			Finance: func(context.Context, map[string]any) (any, error) {
				return map[string]any{
					"tool":           "finance_report_lookup",
					"adapter_status": "ok",
					"candidates": map[string]any{
						"resolved_code":    "06657.HK",
						"resolved_company": "百望股份",
						"resolved_market":  "HK",
					},
				}, nil
			},
			AStock: func(context.Context, map[string]any) (any, error) {
				aStockCalls++
				return map[string]any{"tool": "a_stock_investigation", "adapter_status": "ok"}, nil
			},
			GlobalStock: func(_ context.Context, params map[string]any) (any, error) {
				globalParams = params
				return map[string]any{"tool": "global_stock_investigation", "adapter_status": "ok"}, nil
			},
		},
	}
	payload, err := BuildCompanyResearchLookupPayload(context.Background(), cfg, map[string]any{
		"user_message":         "看下百望云的财报和估值",
		"entity_name":          "百望云",
		"market_hint":          "global",
		"requested_dimensions": []any{"financials", "market_data"},
	})
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if !payload.AnswerReadiness.AnswerReady {
		t.Fatalf("expected ready after finance identity handoff, got %#v", payload.AnswerReadiness)
	}
	if aStockCalls != 0 {
		t.Fatalf("expected HK finance identity to avoid A-share handler, got calls=%d", aStockCalls)
	}
	if globalParams["stock_code"] != "06657" || globalParams["market_hint"] != "hk" {
		t.Fatalf("expected global params enriched from finance identity, got %#v", globalParams)
	}
	if mentions := research.StringListArg(globalParams["entity_mentions"]); containsString(mentions, "百望股份") || containsString(mentions, "06657") {
		t.Fatalf("resolved identity handoff should not turn aliases into multi-entity mentions, got %#v params=%#v", mentions, globalParams)
	}
}

func TestBuildCompanyResearchLookupPayloadHandsOffFinanceResolvedIdentityToNews(t *testing.T) {
	newsParams := map[string]any{}
	cfg := CompanyResearchConfig{
		Handlers: CompanyResearchHandlers{
			Finance: func(context.Context, map[string]any) (any, error) {
				return map[string]any{
					"tool":           "finance_report_lookup",
					"adapter_status": "ok",
					"candidates": map[string]any{
						"resolved_code":    "00700.HK",
						"resolved_company": "腾讯控股",
						"resolved_market":  "HK",
					},
				}, nil
			},
			News: func(_ context.Context, params map[string]any) (any, error) {
				newsParams = params
				return map[string]any{"tool": "latest_news_lookup", "adapter_status": "ok"}, nil
			},
		},
	}
	_, err := BuildCompanyResearchLookupPayload(context.Background(), cfg, map[string]any{
		"user_message":         "帮我查下腾讯科技最新财报和新闻",
		"entity_name":          "腾讯科技",
		"entity_mentions":      []any{"腾讯科技"},
		"requested_dimensions": []any{"financials", "news"},
	})
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if newsParams["entity_name"] != "腾讯控股" || newsParams["stock_code"] != "00700" || newsParams["market_hint"] != "hk" {
		t.Fatalf("expected news params enriched from verified finance identity, got %#v", newsParams)
	}
	if mentions := research.StringListArg(newsParams["entity_mentions"]); !containsString(mentions, "腾讯控股") || containsString(mentions, "腾讯科技") {
		t.Fatalf("expected news mentions to use verified company identity only, got %#v", mentions)
	}
}

func TestBuildCompanyResearchLookupPayloadPrefersGuardedFinanceIssuerName(t *testing.T) {
	newsParams := map[string]any{}
	cfg := CompanyResearchConfig{
		Handlers: CompanyResearchHandlers{
			Finance: func(context.Context, map[string]any) (any, error) {
				return map[string]any{
					"tool":           "finance_report_lookup",
					"adapter_status": "ok",
					"candidates": map[string]any{
						"resolved_code":    "06657.HK",
						"resolved_company": "帮我查百望云",
						"resolved_market":  "HK",
					},
					"metrics": map[string]any{
						"adapter_status": "ok",
						"guard_status":   "passed",
						"evidence": map[string]any{
							"company_name":    "百望股份",
							"stock_code":      "06657.HK",
							"report_period":   "2025年年报",
							"official_source": "https://example.com/06657-report",
							"revenue":         "7.29亿元",
							"net_profit":      "-0.10亿元",
						},
					},
				}, nil
			},
			News: func(_ context.Context, params map[string]any) (any, error) {
				newsParams = params
				return map[string]any{"tool": "latest_news_lookup", "adapter_status": "evidence_incomplete"}, nil
			},
		},
	}
	payload, err := BuildCompanyResearchLookupPayload(context.Background(), cfg, map[string]any{
		"user_message":         "帮我查百望云（06657.HK）的财报和新闻",
		"entity_name":          "百望云",
		"entity_mentions":      []any{"百望云", "06657.HK"},
		"requested_dimensions": []any{"financials", "news"},
	})
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if newsParams["entity_name"] != "百望股份" {
		t.Fatalf("expected guarded metric issuer to replace candidate request text, got %#v", newsParams)
	}
	if payload.TaskPlan == nil || payload.TaskPlan.Subject.CanonicalName != "百望股份" {
		t.Fatalf("expected canonical task subject from guarded finance evidence, got %#v", payload.TaskPlan)
	}
}

func TestMarketEvidenceContractSummaryAnnotatesNegativePE(t *testing.T) {
	summary := marketEvidenceContractSummary(map[string]any{
		"adapter_status": "ok",
		"answer_ready":   true,
		"quote": map[string]any{
			"subject": map[string]any{"entity_name": "百望股份", "stock_code": "06657", "market": "hk"},
			"quote": map[string]any{
				"price":  map[string]any{"value": "12.750", "unit": "HKD"},
				"pe_ttm": map[string]any{"value": "-270.21"},
			},
		},
	})
	if !strings.Contains(summary, "PE(TTM) -270.21（TTM 盈利为负，PE 不具常规可比意义）") {
		t.Fatalf("expected negative PE boundary, got %q", summary)
	}
}

func TestBuildCompanyResearchLookupPayloadConfirmsSubjectFromDownstreamEvidence(t *testing.T) {
	cfg := CompanyResearchConfig{
		Handlers: CompanyResearchHandlers{
			Finance: func(context.Context, map[string]any) (any, error) {
				return map[string]any{
					"tool":           "finance_report_lookup",
					"adapter_status": "ok",
					"candidates": map[string]any{
						"resolved_code":    "00700.HK",
						"resolved_company": "腾讯控股",
						"resolved_market":  "HK",
					},
				}, nil
			},
			GlobalStock: func(context.Context, map[string]any) (any, error) {
				return map[string]any{
					"tool":           "global_stock_investigation",
					"adapter_status": "ok",
					"quote": map[string]any{
						"subject": map[string]any{
							"entity_name": "腾讯控股",
							"ticker":      "00700.HK",
							"market":      "HK",
						},
					},
				}, nil
			},
		},
	}
	payload, err := BuildCompanyResearchLookupPayload(context.Background(), cfg, map[string]any{
		"user_message":         "看下腾讯科技的财报和股价",
		"entity_name":          "腾讯科技",
		"entity_mentions":      []any{"腾讯科技"},
		"market_hint":          "HK",
		"requested_dimensions": []any{"financials", "market_data"},
	})
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if payload.TaskPlan == nil || !payload.TaskPlan.Subject.Verified || payload.TaskPlan.Subject.CanonicalName != "腾讯控股" {
		t.Fatalf("expected downstream identity to confirm task-plan subject, got %#v", payload.TaskPlan)
	}
	result, ok := payloadTaskResult(payload, research.CompanyResearchRoleSubjectResolver)
	if !ok || result.Status != research.CompanyResearchTaskStatusReady || result.Diagnostics["identity_source"] != "downstream_evidence" {
		t.Fatalf("expected ready subject task result from downstream evidence, got %#v ok=%v", result, ok)
	}
	if payload.TaskSummary == nil || !containsRole(payload.TaskSummary.ReadyRoles, research.CompanyResearchRoleSubjectResolver) {
		t.Fatalf("expected task summary to include ready subject resolver, got %#v", payload.TaskSummary)
	}
}

func TestBuildCompanyResearchLookupPayloadDoesNotConfirmSubjectFromEchoedIntent(t *testing.T) {
	cfg := CompanyResearchConfig{
		Handlers: CompanyResearchHandlers{
			Finance: func(context.Context, map[string]any) (any, error) {
				return map[string]any{
					"tool":           "finance_report_lookup",
					"adapter_status": "ok",
					"intent": map[string]any{
						"entity_name": "富途牛牛",
					},
				}, nil
			},
		},
	}
	payload, err := BuildCompanyResearchLookupPayload(context.Background(), cfg, map[string]any{
		"user_message":         "帮我查下富途牛牛最新财报。",
		"entity_name":          "富途牛牛",
		"requested_dimensions": []any{"financials"},
	})
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if payload.TaskPlan == nil {
		t.Fatalf("expected task plan")
	}
	if payload.TaskPlan.Subject.Verified {
		t.Fatalf("echoed input intent must not verify the subject, got %#v", payload.TaskPlan.Subject)
	}
	result, ok := payloadTaskResult(payload, research.CompanyResearchRoleSubjectResolver)
	if !ok || result.Status == research.CompanyResearchTaskStatusReady {
		t.Fatalf("expected unresolved subject task result, got %#v ok=%v", result, ok)
	}
}

func TestBuildCompanyResearchLookupPayloadUsesVerifiedSubjectResolverHandoff(t *testing.T) {
	financeParams := map[string]any{}
	globalParams := map[string]any{}
	aStockCalls := 0
	cfg := CompanyResearchConfig{
		SubjectResolver: func(_ context.Context, request research.SubjectResolutionRequest) (research.SubjectResolution, error) {
			if request.EntityName != "富途牛牛" {
				t.Fatalf("unexpected resolver request: %#v", request)
			}
			return research.SubjectResolution{
				AdapterID:       "test_subject_resolver",
				AdapterStatus:   "ok",
				InputTerm:       request.EntityName,
				PreferredMarket: "US",
				Strategy:        "host_owned_source_backed_lookup",
				SelectedCandidate: &research.SubjectResolutionCandidate{
					EntityName:  "Futu Holdings",
					DisplayName: "Futu Holdings",
					StockCode:   "FUTU",
					Ticker:      "FUTU",
					Market:      "US",
					Source:      "test_source",
					EvidenceURL: "https://example.test/futu",
					Confidence:  0.98,
					Verified:    true,
					MatchReason: "source-backed product alias",
				},
			}, nil
		},
		Handlers: CompanyResearchHandlers{
			Finance: func(_ context.Context, params map[string]any) (any, error) {
				financeParams = params
				return map[string]any{
					"tool":           "finance_report_lookup",
					"adapter_status": "ok",
					"candidates": map[string]any{
						"resolved_code":    "FUTU",
						"resolved_company": "Futu Holdings",
						"resolved_market":  "US",
					},
				}, nil
			},
			AStock: func(context.Context, map[string]any) (any, error) {
				aStockCalls++
				return map[string]any{"tool": "a_stock_investigation", "adapter_status": "ok"}, nil
			},
			GlobalStock: func(_ context.Context, params map[string]any) (any, error) {
				globalParams = params
				return map[string]any{"tool": "global_stock_investigation", "adapter_status": "ok"}, nil
			},
		},
	}
	payload, err := BuildCompanyResearchLookupPayload(context.Background(), cfg, map[string]any{
		"user_message":         "帮我查下富途牛牛最新财报、当前股价估值和最近新闻风险。",
		"entity_name":          "富途牛牛",
		"requested_dimensions": []any{"financials", "market_data"},
	})
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if payload.SubjectResolution == nil || payload.SubjectResolution.SelectedCandidate == nil {
		t.Fatalf("expected subject resolution in payload, got %#v", payload.SubjectResolution)
	}
	if payload.Intent.EntityName != "富途牛牛" {
		t.Fatalf("top-level intent should preserve user input, got %#v", payload.Intent)
	}
	if !payload.AnswerReadiness.AnswerReady {
		t.Fatalf("expected verified subject resolver handoff to make evidence ready, got %#v", payload.AnswerReadiness)
	}
	if aStockCalls != 0 {
		t.Fatalf("expected US subject resolution to avoid A-share handler, got %d calls", aStockCalls)
	}
	for name, params := range map[string]map[string]any{"finance": financeParams, "global": globalParams} {
		if params["entity_name"] != "Futu Holdings" || params["stock_code"] != "FUTU" || params["ticker"] != "FUTU" || params["market_hint"] != "us" {
			t.Fatalf("expected %s params enriched from verified subject resolution, got %#v", name, params)
		}
		mentions := research.StringListArg(params["entity_mentions"])
		if len(mentions) != 1 || mentions[0] != "Futu Holdings" {
			t.Fatalf("expected %s mentions to stay single canonical subject, got %#v params=%#v", name, mentions, params)
		}
	}
}

func TestBuildCompanyResearchLookupPayloadUsesOriginalAliasForIdentityCheckOnly(t *testing.T) {
	financeParams := map[string]any{}
	cfg := CompanyResearchConfig{
		SubjectResolver: func(_ context.Context, request research.SubjectResolutionRequest) (research.SubjectResolution, error) {
			return research.SubjectResolution{
				AdapterID:       "test_subject_resolver",
				AdapterStatus:   "ok",
				InputTerm:       request.EntityName,
				PreferredMarket: "HK",
				Strategy:        "host_owned_source_backed_lookup",
				SelectedCandidate: &research.SubjectResolutionCandidate{
					EntityName:  "Tencent Holdings Limited",
					DisplayName: "Tencent Holdings Limited",
					StockCode:   "00700",
					Ticker:      "0700.HK",
					Market:      "HK",
					Source:      "test_source",
					EvidenceURL: "https://example.test/tencent",
					Confidence:  0.96,
					Verified:    true,
					MatchReason: "source-backed listed subject",
				},
			}, nil
		},
		Handlers: CompanyResearchHandlers{
			Finance: func(_ context.Context, params map[string]any) (any, error) {
				financeParams = params
				return map[string]any{
					"tool":           "finance_report_lookup",
					"adapter_status": "ok",
					"candidates": map[string]any{
						"resolved_code":    "00700.HK",
						"resolved_company": "腾讯控股",
						"resolved_market":  "HK",
					},
				}, nil
			},
		},
	}
	payload, err := BuildCompanyResearchLookupPayload(context.Background(), cfg, map[string]any{
		"user_message":         "帮我查下腾讯最新财报。",
		"entity_name":          "腾讯",
		"entity_mentions":      []any{"腾讯"},
		"requested_dimensions": []any{"financials"},
	})
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if !payload.AnswerReadiness.AnswerReady || containsString(payload.Warnings, "finance_subject_mismatch") {
		t.Fatalf("expected original alias to participate in identity check without mismatch, readiness=%#v warnings=%#v", payload.AnswerReadiness, payload.Warnings)
	}
	if financeParams["entity_name"] != "Tencent Holdings Limited" || financeParams["stock_code"] != "00700" || financeParams["market_hint"] != "hk" {
		t.Fatalf("expected downstream params to keep verified canonical handoff, got %#v", financeParams)
	}
	mentions := research.StringListArg(financeParams["entity_mentions"])
	if len(mentions) != 1 || mentions[0] != "Tencent Holdings Limited" {
		t.Fatalf("downstream mentions must stay single canonical subject, got %#v", mentions)
	}
}

func TestBuildCompanyResearchLookupPayloadAcceptsSameSecurityAcrossLanguageNames(t *testing.T) {
	cfg := CompanyResearchConfig{
		SubjectResolver: func(_ context.Context, request research.SubjectResolutionRequest) (research.SubjectResolution, error) {
			return research.SubjectResolution{
				AdapterID:       "test_subject_resolver",
				AdapterStatus:   "ok",
				InputTerm:       request.EntityName,
				PreferredMarket: "US",
				SelectedCandidate: &research.SubjectResolutionCandidate{
					EntityName:  "全球最大中资港美股券商：富途证券 Futu Holdings Limited",
					DisplayName: "全球最大中资港美股券商：富途证券 Futu Holdings Limited",
					StockCode:   "FUTU",
					Ticker:      "FUTU",
					Market:      "US",
					Source:      "test_source",
					EvidenceURL: "https://example.test/futu",
					Confidence:  0.94,
					Verified:    true,
				},
			}, nil
		},
		Handlers: CompanyResearchHandlers{
			Finance: func(context.Context, map[string]any) (any, error) {
				return map[string]any{
					"tool":           "finance_report_lookup",
					"adapter_status": "ok",
					"candidates": map[string]any{
						"resolved_code":    "FUTU",
						"resolved_company": "Futu Holdings Ltd",
						"resolved_market":  "NASDAQ",
					},
				}, nil
			},
			GlobalStock: func(context.Context, map[string]any) (any, error) {
				return map[string]any{
					"tool":           "global_stock_investigation",
					"adapter_status": "ok",
					"quote": map[string]any{
						"subject": map[string]any{
							"entity_name": "富途控股",
							"ticker":      "FUTU",
							"market":      "US",
						},
					},
				}, nil
			},
		},
	}
	payload, err := BuildCompanyResearchLookupPayload(context.Background(), cfg, map[string]any{
		"user_message":         "帮我查下富途牛牛最新财报、当前股价估值。",
		"entity_name":          "富途牛牛",
		"requested_dimensions": []any{"financials", "market_data"},
	})
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if !payload.AnswerReadiness.AnswerReady {
		t.Fatalf("expected same-code same-market cross-language evidence to stay ready, got readiness=%#v summary=%#v", payload.AnswerReadiness, payload.TaskSummary)
	}
	if payload.TaskSummary != nil && len(payload.TaskSummary.Conflicts) > 0 {
		t.Fatalf("expected no task conflicts for same security evidence, got %#v", payload.TaskSummary.Conflicts)
	}
}

func TestBuildCompanyResearchLookupPayloadIgnoresUnverifiedSubjectResolverHandoff(t *testing.T) {
	globalParams := map[string]any{}
	cfg := CompanyResearchConfig{
		SubjectResolver: func(context.Context, research.SubjectResolutionRequest) (research.SubjectResolution, error) {
			return research.SubjectResolution{
				AdapterID:     "test_subject_resolver",
				AdapterStatus: "ambiguous",
				SelectedCandidate: &research.SubjectResolutionCandidate{
					EntityName: "Futu Holdings",
					StockCode:  "FUTU",
					Ticker:     "FUTU",
					Market:     "US",
					Verified:   false,
				},
				Warnings: []string{"ambiguous_alias"},
			}, nil
		},
		Handlers: CompanyResearchHandlers{
			AStock: func(context.Context, map[string]any) (any, error) {
				return map[string]any{"tool": "a_stock_investigation", "adapter_status": "ok"}, nil
			},
			GlobalStock: func(_ context.Context, params map[string]any) (any, error) {
				globalParams = params
				return map[string]any{"tool": "global_stock_investigation", "adapter_status": "ok"}, nil
			},
		},
	}
	_, err := BuildCompanyResearchLookupPayload(context.Background(), cfg, map[string]any{
		"user_message":         "帮我查下富途牛牛当前股价估值。",
		"entity_name":          "富途牛牛",
		"requested_dimensions": []any{"market_data"},
	})
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if globalParams["entity_name"] != "富途牛牛" {
		t.Fatalf("expected unverified resolution to keep original entity, got %#v", globalParams)
	}
	if globalParams["stock_code"] == "FUTU" || globalParams["ticker"] == "FUTU" || globalParams["market_hint"] == "us" {
		t.Fatalf("expected unverified resolution not to enrich downstream params, got %#v", globalParams)
	}
}

func TestCompanyResearchAnswerContractPrefersVerifiedDownstreamIdentityOverUnverifiedResolverLabel(t *testing.T) {
	cfg := CompanyResearchConfig{
		SubjectResolver: func(context.Context, research.SubjectResolutionRequest) (research.SubjectResolution, error) {
			return research.SubjectResolution{
				AdapterID:     "test_subject_resolver",
				AdapterStatus: "ambiguous",
				InputTerm:     "联想",
				SelectedCandidate: &research.SubjectResolutionCandidate{
					EntityName: "常见问题",
					StockCode:  "00992",
					Market:     "hk",
					Verified:   false,
				},
				Warnings: []string{"subject_resolution_ambiguous"},
			}, nil
		},
		Handlers: CompanyResearchHandlers{
			Finance: func(context.Context, map[string]any) (any, error) {
				return map[string]any{
					"tool":           "finance_report_lookup",
					"adapter_status": "ok",
					"candidates": map[string]any{
						"resolved_code":    "00992.HK",
						"resolved_company": "联想集团",
						"resolved_market":  "HK",
					},
				}, nil
			},
		},
	}
	payload, err := BuildCompanyResearchLookupPayload(context.Background(), cfg, map[string]any{
		"user_message":         "查一下联想最新财报。",
		"entity_name":          "联想",
		"requested_dimensions": []any{"financials"},
	})
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if !payload.AnswerReadiness.AnswerReady {
		t.Fatalf("expected downstream finance evidence to make answer ready, got %#v", payload.AnswerReadiness)
	}
	reply := payload.AnswerContract.FinalAnswerDraft
	if !strings.Contains(reply, "主体：联想集团") || strings.Contains(reply, "主体：常见问题") {
		t.Fatalf("expected answer contract to display verified downstream company identity, got %q", reply)
	}
}

func TestResolveSubjectKeepsPrefixedFailureWarningStable(t *testing.T) {
	resolution, warnings := resolveSubject(context.Background(), func(context.Context, research.SubjectResolutionRequest) (research.SubjectResolution, error) {
		return research.SubjectResolution{
			AdapterStatus: "source_unavailable",
			FailureCode:   "subject_resolution_source_unavailable",
			InputTerm:     "富途牛牛",
		}, nil
	}, research.CompanyResearchIntent{EntityName: "富途牛牛"})
	if resolution == nil || resolution.FailureCode != "subject_resolution_source_unavailable" {
		t.Fatalf("unexpected subject resolution: %#v", resolution)
	}
	if len(warnings) != 1 || warnings[0] != "subject_resolution_source_unavailable" {
		t.Fatalf("expected stable warning code without duplicate prefix, got %#v", warnings)
	}
}

func TestBuildCompanyResearchLookupPayloadBlocksMismatchedFinanceIdentityHandoff(t *testing.T) {
	globalParams := map[string]any{}
	newsParams := map[string]any{}
	cfg := CompanyResearchConfig{
		Handlers: CompanyResearchHandlers{
			Finance: func(context.Context, map[string]any) (any, error) {
				return map[string]any{
					"tool":           "finance_report_lookup",
					"adapter_status": "ok",
					"candidates": map[string]any{
						"resolved_code":    "MSFT",
						"resolved_company": "MICROSOFT CORP",
						"resolved_market":  "US",
					},
				}, nil
			},
			GlobalStock: func(_ context.Context, params map[string]any) (any, error) {
				globalParams = params
				return map[string]any{"tool": "global_stock_investigation", "adapter_status": "ok"}, nil
			},
			News: func(_ context.Context, params map[string]any) (any, error) {
				newsParams = params
				return map[string]any{"tool": "latest_news_lookup", "adapter_status": "ok"}, nil
			},
		},
	}
	payload, err := BuildCompanyResearchLookupPayload(context.Background(), cfg, map[string]any{
		"user_message":         "比较阿里巴巴、Microsoft",
		"entity_name":          "阿里巴巴",
		"entity_mentions":      []any{"阿里巴巴"},
		"market_hint":          "us/hk",
		"requested_dimensions": []any{"financials", "market_data", "news"},
	})
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if payload.AnswerReadiness.AnswerReady || !containsString(payload.AnswerReadiness.MissingDimensions, "financials") {
		t.Fatalf("expected mismatched finance subject to degrade financials, got %#v", payload.AnswerReadiness)
	}
	if !containsString(payload.Warnings, "finance_subject_mismatch") {
		t.Fatalf("expected finance_subject_mismatch warning, got %#v", payload.Warnings)
	}
	if payload.TaskSummary == nil || len(payload.TaskSummary.Conflicts) == 0 {
		t.Fatalf("expected task summary conflicts for mismatched finance identity, got %#v", payload.TaskSummary)
	}
	if !strings.Contains(payload.AnswerContract.FinalAnswerDraft, "任务冲突摘要") {
		t.Fatalf("expected answer contract to include task conflict summary, got %q", payload.AnswerContract.FinalAnswerDraft)
	}
	if globalParams["stock_code"] == "MSFT" || globalParams["ticker"] == "MSFT" || globalParams["entity_name"] != "阿里巴巴" {
		t.Fatalf("expected mismatched finance identity not to poison global-stock handoff, got %#v", globalParams)
	}
	if newsParams["stock_code"] == "MSFT" || newsParams["ticker"] == "MSFT" || newsParams["entity_name"] != "阿里巴巴" {
		t.Fatalf("expected mismatched finance identity not to poison news handoff, got %#v", newsParams)
	}
}

func TestCompanyNameMatchesIntentSuppressesShortCJKPrefixWhenSpecificAliasExists(t *testing.T) {
	intent := research.CompanyResearchIntent{
		EntityName:     "阿里",
		EntityMentions: []string{"阿里", "阿里巴巴"},
	}
	if companyNameMatchesIntent(intent, "阿里科") {
		t.Fatal("expected short CJK prefix not to match a different observed company when a specific alias exists")
	}
	if !companyNameMatchesIntent(intent, "阿里巴巴-W") {
		t.Fatal("expected specific CJK alias to match observed listed company name")
	}
}

func TestCompanyNameMatchesIntentAllowsShortCJKCorporateSuffix(t *testing.T) {
	intent := research.CompanyResearchIntent{EntityName: "腾讯"}
	if !companyNameMatchesIntent(intent, "腾讯控股") {
		t.Fatal("expected short CJK company alias to match a generic corporate suffix")
	}
	if companyNameMatchesIntent(intent, "腾讯音乐") {
		t.Fatal("expected short CJK company alias not to match a different product/subsidiary-style suffix")
	}
}

func TestBuildCompanyResearchLookupPayloadDoesNotTrustEchoedFinanceIntentForIdentity(t *testing.T) {
	cfg := CompanyResearchConfig{
		SubjectResolver: func(_ context.Context, request research.SubjectResolutionRequest) (research.SubjectResolution, error) {
			return research.SubjectResolution{
				AdapterID:       "test_subject_resolver",
				AdapterStatus:   "ok",
				InputTerm:       request.EntityName,
				PreferredMarket: "US",
				SelectedCandidate: &research.SubjectResolutionCandidate{
					EntityName:  "Microsoft Corporation",
					DisplayName: "Microsoft Corporation",
					StockCode:   "MSFT",
					Ticker:      "MSFT",
					Market:      "US",
					Source:      "test_source",
					EvidenceURL: "https://example.test/msft",
					Confidence:  0.96,
					Verified:    true,
				},
			}, nil
		},
		Handlers: CompanyResearchHandlers{
			Finance: func(context.Context, map[string]any) (any, error) {
				return map[string]any{
					"tool":           "finance_report_lookup",
					"adapter_status": "ok",
					"intent": map[string]any{
						"entity_name": "Microsoft Corporation",
						"stock_code":  "MSFT",
						"ticker":      "MSFT",
						"market":      "US",
					},
					"candidates": map[string]any{
						"resolved_code":    "AUBN.O",
						"resolved_company": "Auburn National Bancorporation",
						"resolved_market":  "US",
					},
				}, nil
			},
			News: func(context.Context, map[string]any) (any, error) {
				return map[string]any{"tool": "latest_news_lookup", "adapter_status": "ok", "guard_status": "passed", "passed": true}, nil
			},
		},
	}
	payload, err := BuildCompanyResearchLookupPayload(context.Background(), cfg, map[string]any{
		"user_message":         "查一下 Microsoft 最近重要新闻，再结合最新财报表现说明风险边界。",
		"entity_name":          "Microsoft",
		"requested_dimensions": []any{"financials", "news"},
	})
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if payload.AnswerReadiness.AnswerReady || payload.AnswerReadiness.FailureCode != "company_research_missing_required_dimensions" ||
		!containsString(payload.AnswerReadiness.MissingDimensions, "financials") {
		t.Fatalf("expected echoed-intent mismatch to remove financials readiness, got %#v", payload.AnswerReadiness)
	}
	if !containsString(payload.Warnings, "finance_subject_mismatch") {
		t.Fatalf("expected finance_subject_mismatch warning, got %#v", payload.Warnings)
	}
	if !payloadHasConflict(payload, "subject_identity_mismatch") {
		t.Fatalf("expected subject identity conflict, got %#v", payload.TaskSummary)
	}
}

func TestBuildCompanyResearchLookupPayloadTreatsStructuredFinanceMetricsAsReady(t *testing.T) {
	cfg := CompanyResearchConfig{
		Handlers: CompanyResearchHandlers{
			Finance: func(context.Context, map[string]any) (any, error) {
				return map[string]any{
					"tool":           "finance_report_lookup",
					"adapter_status": "ok",
					"failure_code":   "guard_not_passed,brief_missing,stop_after_guard_not_confirmed",
					"answer_readiness": map[string]any{
						"answer_ready":           false,
						"requested_fields_ready": true,
						"degrade_reason":         "brief_not_ready",
					},
					"candidates": map[string]any{
						"resolved_company": "Example Tech Inc",
						"resolved_code":    "EXM",
						"resolved_market":  "US",
					},
					"metrics": map[string]any{
						"evidence": map[string]any{
							"company_name":        "Example Tech Inc",
							"stock_code":          "EXM",
							"report_period":       "2026-03-31 10-K",
							"official_source":     "https://www.sec.gov/example/exm-20260331.htm",
							"revenue":             "USD10.0 billion",
							"net_profit":          "USD2.0 billion",
							"operating_cash_flow": "USD3.0 billion",
							"metric_evidence": map[string]any{
								"revenue": map[string]any{
									"value":  "USD10.0 billion",
									"source": "https://www.sec.gov/example/exm-20260331.htm",
								},
								"net_profit": map[string]any{
									"value":  "USD2.0 billion",
									"source": "https://www.sec.gov/example/exm-20260331.htm",
								},
							},
						},
					},
				}, nil
			},
			GlobalStock: func(context.Context, map[string]any) (any, error) {
				return map[string]any{
					"tool":           "global_stock_investigation",
					"adapter_status": "ok",
					"quote": map[string]any{
						"subject": map[string]any{
							"entity_name": "Example Tech Inc",
							"stock_code":  "EXM",
							"market":      "us",
						},
						"quote": map[string]any{
							"price": map[string]any{"value": "100.00", "unit": "USD"},
						},
					},
				}, nil
			},
			News: func(context.Context, map[string]any) (any, error) {
				return map[string]any{
					"tool":           "latest_news_lookup",
					"adapter_status": "ok",
					"guard_status":   "passed",
					"passed":         true,
					"summary":        "Example Tech Inc released a source-backed operating update.",
					"source_url":     "https://news.example.com/example-tech-update",
				}, nil
			},
		},
	}
	payload, err := BuildCompanyResearchLookupPayload(context.Background(), cfg, map[string]any{
		"user_message":         "比较 Example Tech 的财务、行情和新闻证据。",
		"entity_name":          "Example Tech",
		"market_hint":          "US",
		"requested_dimensions": []any{"financials", "market_data", "news"},
	})
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if !payload.AnswerReadiness.AnswerReady || !containsString(payload.AnswerReadiness.ReadyDimensions, "financials") {
		t.Fatalf("expected structured finance metrics to make financials ready, got %#v", payload.AnswerReadiness)
	}
	result, ok := payloadTaskResult(payload, research.CompanyResearchRoleFinanceAnalyst)
	if !ok || !result.EvidenceReady || result.Status != research.CompanyResearchTaskStatusReady {
		t.Fatalf("expected finance task result ready, got %#v ok=%v", result, ok)
	}
	if result.Diagnostics["finance_brief_failure_code"] == "" {
		t.Fatalf("expected original brief failure in diagnostics, got %#v", result.Diagnostics)
	}
	if !strings.Contains(payload.AnswerContract.FinalAnswerDraft, "财务：Example Tech Inc") ||
		!strings.Contains(payload.AnswerContract.FinalAnswerDraft, "收入 USD10.0 billion") {
		t.Fatalf("expected final answer draft to include structured finance metrics, got %q", payload.AnswerContract.FinalAnswerDraft)
	}
}

func TestBuildCompanyResearchLookupPayloadDegradesOnCrossTaskCodeConflict(t *testing.T) {
	cfg := CompanyResearchConfig{
		Handlers: CompanyResearchHandlers{
			Finance: func(context.Context, map[string]any) (any, error) {
				return map[string]any{
					"tool":           "finance_report_lookup",
					"adapter_status": "ok",
					"candidates": map[string]any{
						"resolved_code":    "00700.HK",
						"resolved_company": "腾讯控股",
						"resolved_market":  "HK",
					},
				}, nil
			},
			GlobalStock: func(context.Context, map[string]any) (any, error) {
				return map[string]any{
					"tool":           "global_stock_investigation",
					"adapter_status": "ok",
					"quote": map[string]any{
						"subject": map[string]any{
							"entity_name": "腾讯控股",
							"ticker":      "09988.HK",
							"market":      "HK",
						},
					},
				}, nil
			},
		},
	}
	payload, err := BuildCompanyResearchLookupPayload(context.Background(), cfg, map[string]any{
		"user_message":         "看下腾讯控股的财报和估值",
		"entity_name":          "腾讯控股",
		"market_hint":          "HK",
		"requested_dimensions": []any{"financials", "market_data"},
	})
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if payload.AnswerReadiness.AnswerReady || payload.AnswerReadiness.FailureCode != "company_research_task_conflicts" {
		t.Fatalf("expected cross-task code conflict to degrade readiness, got %#v", payload.AnswerReadiness)
	}
	if payload.TaskSummary == nil || len(payload.TaskSummary.Conflicts) == 0 {
		t.Fatalf("expected task summary conflicts, got %#v", payload.TaskSummary)
	}
	if got := payload.TaskSummary.Conflicts[0].Code; got != "cross_task_code_mismatch" {
		t.Fatalf("expected code conflict first, got %#v", payload.TaskSummary.Conflicts)
	}
}

func TestBuildCompanyResearchLookupPayloadDoesNotConflictOnHKCodePadding(t *testing.T) {
	cfg := CompanyResearchConfig{
		Handlers: CompanyResearchHandlers{
			Finance: func(context.Context, map[string]any) (any, error) {
				return map[string]any{
					"tool":           "finance_report_lookup",
					"adapter_status": "ok",
					"candidates": map[string]any{
						"resolved_code":    "00700.HK",
						"resolved_company": "腾讯控股",
						"resolved_market":  "HK",
					},
				}, nil
			},
			GlobalStock: func(context.Context, map[string]any) (any, error) {
				return map[string]any{
					"tool":           "global_stock_investigation",
					"adapter_status": "ok",
					"quote": map[string]any{
						"subject": map[string]any{
							"entity_name": "腾讯控股",
							"ticker":      "0700.HK",
							"market":      "HK",
						},
					},
				}, nil
			},
		},
	}
	payload, err := BuildCompanyResearchLookupPayload(context.Background(), cfg, map[string]any{
		"user_message":         "看下腾讯控股的财报和估值",
		"entity_name":          "腾讯控股",
		"market_hint":          "HK",
		"requested_dimensions": []any{"financials", "market_data"},
	})
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if !payload.AnswerReadiness.AnswerReady {
		t.Fatalf("expected HK code padding variants to stay ready, got %#v summary=%#v", payload.AnswerReadiness, payload.TaskSummary)
	}
	if payload.TaskSummary != nil && len(payload.TaskSummary.Conflicts) > 0 {
		t.Fatalf("expected no conflict for HK code padding variants, got %#v", payload.TaskSummary.Conflicts)
	}
}

func TestBuildCompanyResearchLookupPayloadSummarizesGlobalStockQuotesArray(t *testing.T) {
	cfg := CompanyResearchConfig{
		Handlers: CompanyResearchHandlers{
			Finance: func(context.Context, map[string]any) (any, error) {
				return map[string]any{
					"tool":           "finance_report_lookup",
					"adapter_status": "ok",
					"candidates": map[string]any{
						"resolved_code":    "00700.HK",
						"resolved_company": "腾讯控股",
						"resolved_market":  "HK",
					},
				}, nil
			},
			GlobalStock: func(context.Context, map[string]any) (any, error) {
				return map[string]any{
					"tool":           "global_stock_investigation",
					"adapter_status": "ok",
					"quotes": []any{
						map[string]any{
							"adapter_status": "ok",
							"freshness":      map[string]any{"as_of": "2026/05/28 16:08:31"},
							"subject": map[string]any{
								"entity_name": "腾讯控股",
								"stock_code":  "00700",
								"market":      "hk",
							},
							"quote": map[string]any{
								"price":      map[string]any{"value": "425.000", "unit": "HKD"},
								"pe_ttm":     map[string]any{"value": "15.57"},
								"pb":         map[string]any{"value": "3.05"},
								"market_cap": map[string]any{"value": "38751.7609", "unit": "100_million", "currency": "HKD"},
							},
						},
					},
				}, nil
			},
		},
	}
	payload, err := BuildCompanyResearchLookupPayload(context.Background(), cfg, map[string]any{
		"user_message":         "帮我交叉核验腾讯控股最新财报、当前股价市值。",
		"entity_name":          "腾讯控股",
		"market_hint":          "HK",
		"requested_dimensions": []any{"financials", "market_data"},
	})
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if !payload.AnswerReadiness.AnswerReady {
		t.Fatalf("expected ready payload, got %#v", payload.AnswerReadiness)
	}
	result, ok := payloadTaskResult(payload, research.CompanyResearchRoleSubjectResolver)
	if !ok || !strings.Contains(result.Diagnostics["identity_source_roles"], "market_analyst") {
		t.Fatalf("expected market quotes array to participate in subject confirmation, got %#v ok=%v", result, ok)
	}
	draft := payload.AnswerContract.FinalAnswerDraft
	for _, want := range []string{"行情/估值：腾讯控股", "价格 425.000HKD", "PE(TTM) 15.57", "市值 38751.7609亿 HKD"} {
		if !strings.Contains(draft, want) {
			t.Fatalf("expected draft to contain %q, got %q", want, draft)
		}
	}
}

func TestBuildCompanyResearchLookupPayloadDegradesOnFreshnessConflict(t *testing.T) {
	cfg := CompanyResearchConfig{
		Handlers: CompanyResearchHandlers{
			Finance: func(context.Context, map[string]any) (any, error) {
				return map[string]any{
					"tool":           "finance_report_lookup",
					"adapter_status": "ok",
					"metrics": map[string]any{
						"evaluation": map[string]any{
							"period_latest": false,
						},
						"evidence": map[string]any{
							"company_name":  "样例公司",
							"report_period": "2024年年报",
						},
					},
				}, nil
			},
		},
	}
	payload, err := BuildCompanyResearchLookupPayload(context.Background(), cfg, map[string]any{
		"user_message":         "查下样例公司的最新财报",
		"entity_name":          "样例公司",
		"requested_dimensions": []any{"financials"},
		"freshness":            map[string]any{"require_latest": true},
	})
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if payload.AnswerReadiness.AnswerReady || payload.AnswerReadiness.FailureCode != "company_research_task_conflicts" {
		t.Fatalf("expected freshness conflict to degrade readiness, got %#v", payload.AnswerReadiness)
	}
	if !payloadHasConflict(payload, "freshness_not_confirmed") {
		t.Fatalf("expected freshness conflict in task summary, got %#v", payload.TaskSummary)
	}
}

func TestBuildCompanyResearchLookupPayloadDegradesOnSourceFactMismatch(t *testing.T) {
	cfg := CompanyResearchConfig{
		Handlers: CompanyResearchHandlers{
			Finance: func(context.Context, map[string]any) (any, error) {
				return map[string]any{
					"tool":           "finance_report_lookup",
					"adapter_status": "ok",
					"metrics": map[string]any{
						"evaluation": map[string]any{
							"source_accepted": true,
						},
						"evidence": map[string]any{
							"company_name": "样例公司",
							"metric_evidence": map[string]any{
								"revenue": map[string]any{
									"value":  "101亿元",
									"period": "2025年年报",
									"source": "https://example.test/metrics",
								},
							},
						},
					},
					"assessment_projection": map[string]any{
						"verified_facts": []any{
							map[string]any{
								"field":  "revenue",
								"value":  "100亿元",
								"period": "2025年年报",
								"source": "https://example.test/brief",
							},
						},
					},
				}, nil
			},
		},
	}
	payload, err := BuildCompanyResearchLookupPayload(context.Background(), cfg, map[string]any{
		"user_message":         "查下样例公司的财报",
		"entity_name":          "样例公司",
		"requested_dimensions": []any{"financials"},
	})
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if payload.AnswerReadiness.AnswerReady || payload.AnswerReadiness.FailureCode != "company_research_task_conflicts" {
		t.Fatalf("expected source fact mismatch to degrade readiness, got %#v", payload.AnswerReadiness)
	}
	if !payloadHasConflict(payload, "source_fact_mismatch") {
		t.Fatalf("expected source fact mismatch in task summary, got %#v", payload.TaskSummary)
	}
	if !strings.Contains(payload.AnswerContract.FinalAnswerDraft, "任务冲突摘要") {
		t.Fatalf("expected answer contract to expose task conflict, got %q", payload.AnswerContract.FinalAnswerDraft)
	}
}

func TestBuildCompanyComparePayloadUsesSubjects(t *testing.T) {
	cfg := CompanyResearchConfig{
		Handlers: CompanyResearchHandlers{
			Finance: func(context.Context, map[string]any) (any, error) {
				return map[string]any{"tool": "finance_report_lookup", "adapter_status": "ok"}, nil
			},
			News: func(context.Context, map[string]any) (any, error) {
				return map[string]any{"tool": "latest_news_lookup", "adapter_status": "ok"}, nil
			},
		},
	}
	payload, err := BuildCompanyComparePayload(context.Background(), cfg, map[string]any{
		"user_message":         "对比小鹏和理想",
		"requested_dimensions": []any{"financials", "news"},
		"comparison_subjects": []any{
			map[string]any{"entity_name": "小鹏汽车", "market_hint": "HK"},
			map[string]any{"entity_name": "理想汽车", "market_hint": "HK"},
		},
	})
	if err != nil {
		t.Fatalf("compare: %v", err)
	}
	if len(payload.Subjects) != 2 || !payload.AnswerReadiness.AnswerReady {
		t.Fatalf("unexpected compare payload: %#v", payload)
	}
}

func TestBuildCompanyComparePayloadRoutesCompositeMarketHints(t *testing.T) {
	globalCalls := 0
	payload, err := BuildCompanyComparePayload(context.Background(), CompanyResearchConfig{
		Handlers: CompanyResearchHandlers{
			Finance: func(context.Context, map[string]any) (any, error) {
				return map[string]any{"tool": "finance_report_lookup", "adapter_status": "ok"}, nil
			},
			GlobalStock: func(context.Context, map[string]any) (any, error) {
				globalCalls++
				return map[string]any{"tool": "global_stock_investigation", "adapter_status": "ok"}, nil
			},
			News: func(context.Context, map[string]any) (any, error) {
				return map[string]any{"tool": "latest_news_lookup", "adapter_status": "ok"}, nil
			},
		},
	}, map[string]any{
		"user_message":         "对比小鹏、理想和蔚来",
		"requested_dimensions": []any{"financials", "market_data", "news"},
		"comparison_subjects": []any{
			map[string]any{"entity_name": "小鹏汽车", "market_hint": "HK/US"},
			map[string]any{"entity_name": "理想汽车", "market_hint": "HK, US"},
			map[string]any{"entity_name": "蔚来汽车", "market_hint": "港股/美股"},
		},
	})
	if err != nil {
		t.Fatalf("compare: %v", err)
	}
	if globalCalls != 3 {
		t.Fatalf("expected composite market hints to route to global-stock for each subject, got %d calls payload=%#v", globalCalls, payload)
	}
	if !payload.AnswerReadiness.AnswerReady {
		t.Fatalf("expected composite market hints to satisfy market_data, got %#v", payload)
	}
}

func TestBuildCompanyComparePayloadRetriesCompositeMarketHintsWhenSubjectPartial(t *testing.T) {
	financeMarkets := []string{}
	payload, err := BuildCompanyComparePayload(context.Background(), CompanyResearchConfig{
		Handlers: CompanyResearchHandlers{
			Finance: func(_ context.Context, params map[string]any) (any, error) {
				market := research.StringArg(params["market_hint"])
				financeMarkets = append(financeMarkets, market)
				if strings.EqualFold(market, "hk") {
					return map[string]any{
						"tool":           "finance_report_lookup",
						"adapter_status": "ok",
						"candidates": map[string]any{
							"resolved_company": "阿里巴巴",
							"resolved_code":    "09988.HK",
							"resolved_market":  "HK",
						},
					}, nil
				}
				return map[string]any{
					"tool":           "finance_report_lookup",
					"adapter_status": "evidence_incomplete",
					"failure_code":   "requested_fields_missing",
					"failure_class":  "evidence_missing",
				}, nil
			},
			GlobalStock: func(_ context.Context, params map[string]any) (any, error) {
				return map[string]any{
					"tool":           "global_stock_investigation",
					"adapter_status": "ok",
					"quote": map[string]any{
						"subject": map[string]any{
							"entity_name": params["entity_name"],
							"ticker":      "09988",
							"market":      "HK",
						},
					},
				}, nil
			},
			News: func(context.Context, map[string]any) (any, error) {
				return map[string]any{"tool": "latest_news_lookup", "adapter_status": "ok"}, nil
			},
		},
	}, map[string]any{
		"user_message":         "比较阿里巴巴的财报、市场数据和新闻",
		"requested_dimensions": []any{"financials", "market_data", "news"},
		"comparison_subjects": []any{
			map[string]any{"entity_name": "阿里巴巴", "market_hint": "HK/US"},
		},
	})
	if err != nil {
		t.Fatalf("compare: %v", err)
	}
	if !payload.AnswerReadiness.AnswerReady || len(payload.Subjects) != 1 || !payload.Subjects[0].AnswerReadiness.AnswerReady {
		t.Fatalf("expected composite-market fallback to choose ready subject payload, got readiness=%#v subjects=%#v", payload.AnswerReadiness, payload.Subjects)
	}
	if !containsString(financeMarkets, "HK/US") || !containsString(financeMarkets, "hk") {
		t.Fatalf("expected initial composite market and narrowed HK retry, got %#v", financeMarkets)
	}
	if got := payload.Subjects[0].Intent.MarketHint; got != "hk" {
		t.Fatalf("expected selected subject payload to expose narrowed market hint, got %q", got)
	}
}

func TestNormalizeIdentityCodeForMarketTreatsUSRICAsSameSecurity(t *testing.T) {
	if left, right := normalizeIdentityCodeForMarket("MSFT.O", "US"), normalizeIdentityCodeForMarket("MSFT", "US"); left != right {
		t.Fatalf("expected RIC and plain ticker to normalize equally, got %q vs %q", left, right)
	}
	if got := marketHintFromCode("BABA.N"); got != "us" {
		t.Fatalf("expected RIC .N suffix to imply US market, got %q", got)
	}
	if !identityHintsSameSecurity(
		downstreamIdentityHint{CompanyName: "微软", StockCode: "MSFT.O", MarketHint: "us"},
		downstreamIdentityHint{CompanyName: "微软", StockCode: "MSFT", MarketHint: "us"},
	) {
		t.Fatalf("expected RIC and plain ticker hints to be treated as the same security")
	}
}

func TestMarketHintRoutingUsesStructuredCategories(t *testing.T) {
	tests := []struct {
		name       string
		marketHint string
		wantA      bool
		wantGlobal bool
	}{
		{name: "empty probes both", marketHint: "", wantA: true, wantGlobal: true},
		{name: "a share zh", marketHint: "A股", wantA: true, wantGlobal: false},
		{name: "a share en", marketHint: "A-share", wantA: true, wantGlobal: false},
		{name: "hk us slash", marketHint: "HK/US", wantA: false, wantGlobal: true},
		{name: "hk us zh", marketHint: "港股/美股", wantA: false, wantGlobal: true},
		{name: "hk us compact", marketHint: "港美股", wantA: false, wantGlobal: true},
		{name: "global structured", marketHint: "global", wantA: false, wantGlobal: true},
		{name: "overseas structured", marketHint: "overseas", wantA: false, wantGlobal: true},
		{name: "avoid single char company-like false positive", marketHint: "美团", wantA: false, wantGlobal: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			intent := research.CompanyResearchIntent{MarketHint: tt.marketHint}
			if got := shouldCallAStock(intent); got != tt.wantA {
				t.Fatalf("shouldCallAStock(%q)=%v want %v", tt.marketHint, got, tt.wantA)
			}
			if got := shouldCallGlobalStock(intent); got != tt.wantGlobal {
				t.Fatalf("shouldCallGlobalStock(%q)=%v want %v", tt.marketHint, got, tt.wantGlobal)
			}
		})
	}
}

func TestBuildCompanyComparePayloadAddsAnswerContractWhenPartial(t *testing.T) {
	cfg := CompanyResearchConfig{
		Handlers: CompanyResearchHandlers{
			Finance: func(context.Context, map[string]any) (any, error) {
				return map[string]any{"tool": "finance_report_lookup", "adapter_status": "ok"}, nil
			},
			News: func(_ context.Context, params map[string]any) (any, error) {
				if params["entity_name"] == "乙公司" {
					return map[string]any{
						"tool":           "latest_news_lookup",
						"adapter_status": "evidence_incomplete",
						"failure_code":   "latest_news_search_open_page_no_grounded_sources",
						"failure_class":  "evidence_missing",
						"answer_contract": map[string]any{
							"final_answer_recommended": true,
							"do_not_retry_tools":       []any{"latest_news_lookup"},
						},
					}, nil
				}
				return map[string]any{"tool": "latest_news_lookup", "adapter_status": "ok"}, nil
			},
		},
	}
	payload, err := BuildCompanyComparePayload(context.Background(), cfg, map[string]any{
		"user_message":         "比较甲公司和乙公司",
		"requested_dimensions": []any{"financials", "news"},
		"comparison_subjects": []any{
			map[string]any{"entity_name": "甲公司"},
			map[string]any{"entity_name": "乙公司"},
		},
	})
	if err != nil {
		t.Fatalf("compare: %v", err)
	}
	if payload.AnswerReadiness.AnswerReady || payload.AnswerReadiness.FailureCode != "company_compare_partial" {
		t.Fatalf("expected partial compare readiness, got %#v", payload.AnswerReadiness)
	}
	if !payload.AnswerReadiness.SafeToAnswer {
		t.Fatalf("expected partial compare to be safe for bounded answer, got %#v", payload.AnswerReadiness)
	}
	if payload.AnswerContract == nil || !payload.AnswerContract.FinalAnswerRecommended {
		t.Fatalf("expected final answer contract for partial compare, got %#v", payload.AnswerContract)
	}
	if payload.AnswerContract.RecoveryRecommended ||
		!containsString(payload.AnswerReadiness.ReadyDimensions, "financials") {
		t.Fatalf("expected terminal downstream news contract and common financial readiness, got readiness=%#v contract=%#v", payload.AnswerReadiness, payload.AnswerContract)
	}
	if !strings.Contains(payload.AnswerContract.FinalAnswerDraft, "已 ready 的主体：甲公司") ||
		!strings.Contains(payload.AnswerContract.FinalAnswerDraft, "乙公司(news)") {
		t.Fatalf("expected compare draft to name missing subjects, got %q", payload.AnswerContract.FinalAnswerDraft)
	}
}

func TestBuildCompanyComparePayloadKeepsRecoverablePartialOpen(t *testing.T) {
	cfg := CompanyResearchConfig{
		Handlers: CompanyResearchHandlers{
			Finance: func(_ context.Context, params map[string]any) (any, error) {
				if params["entity_name"] == "乙公司" {
					return map[string]any{
						"tool":           "finance_report_lookup",
						"adapter_status": "evidence_incomplete",
						"failure_code":   "requested_fields_missing",
					}, nil
				}
				return map[string]any{"tool": "finance_report_lookup", "adapter_status": "ok"}, nil
			},
			News: func(context.Context, map[string]any) (any, error) {
				return map[string]any{"tool": "latest_news_lookup", "adapter_status": "ok"}, nil
			},
		},
	}
	payload, err := BuildCompanyComparePayload(context.Background(), cfg, map[string]any{
		"user_message":         "比较甲公司和乙公司",
		"requested_dimensions": []any{"financials", "news"},
		"comparison_subjects": []any{
			map[string]any{"entity_name": "甲公司"},
			map[string]any{"entity_name": "乙公司"},
		},
	})
	if err != nil {
		t.Fatalf("compare: %v", err)
	}
	if payload.AnswerReadiness.AnswerReady || payload.AnswerReadiness.FailureClass != "evidence_missing" {
		t.Fatalf("expected recoverable compare partial, got %#v", payload.AnswerReadiness)
	}
	if payload.AnswerContract == nil || payload.AnswerContract.FinalAnswerRecommended {
		t.Fatalf("expected recoverable partial to stay open for follow-up tools, got %#v", payload.AnswerContract)
	}
	if !payload.AnswerContract.RecoveryRecommended ||
		!containsString(payload.AnswerContract.SuggestedRecoveryTools, "finance_report_lookup") ||
		!containsString(payload.AnswerContract.DoNotRetryTools, research.ToolCompanyCompareLookup) ||
		containsString(payload.AnswerContract.DoNotRetryTools, research.ToolCompanyResearchLookup) {
		t.Fatalf("unexpected recovery contract: %#v", payload.AnswerContract)
	}
	if len(payload.AnswerContract.RecoveryTargets) != 1 ||
		payload.AnswerContract.RecoveryTargets[0].EntityName != "乙公司" ||
		payload.AnswerContract.RecoveryTargets[0].MissingDimension != "financials" {
		t.Fatalf("unexpected recovery targets: %#v", payload.AnswerContract.RecoveryTargets)
	}
	if !strings.Contains(payload.AnswerContract.FinalAnswerDraft, "恢复建议") ||
		!strings.Contains(payload.AnswerContract.FinalAnswerDraft, "finance_report_lookup") {
		t.Fatalf("expected recovery guidance in draft, got %q", payload.AnswerContract.FinalAnswerDraft)
	}
}

func TestBuildCompanyComparePayloadAnswerContractShowsPartialSubjectDimensions(t *testing.T) {
	cfg := CompanyResearchConfig{
		Handlers: CompanyResearchHandlers{
			Finance: func(_ context.Context, params map[string]any) (any, error) {
				if params["entity_name"] == "乙公司" {
					return map[string]any{"tool": "finance_report_lookup", "adapter_status": "unsupported", "failure_code": "unsupported"}, nil
				}
				return map[string]any{"tool": "finance_report_lookup", "adapter_status": "ok"}, nil
			},
			News: func(_ context.Context, params map[string]any) (any, error) {
				if params["entity_name"] == "甲公司" {
					return map[string]any{"tool": "latest_news_lookup", "adapter_status": "unsupported", "failure_code": "unsupported"}, nil
				}
				return map[string]any{"tool": "latest_news_lookup", "adapter_status": "ok"}, nil
			},
		},
	}
	payload, err := BuildCompanyComparePayload(context.Background(), cfg, map[string]any{
		"user_message":         "比较甲公司和乙公司",
		"requested_dimensions": []any{"financials", "news"},
		"comparison_subjects": []any{
			map[string]any{"entity_name": "甲公司"},
			map[string]any{"entity_name": "乙公司"},
		},
	})
	if err != nil {
		t.Fatalf("compare: %v", err)
	}
	if payload.AnswerContract == nil {
		t.Fatalf("expected final answer contract for partial compare")
	}
	draft := payload.AnswerContract.FinalAnswerDraft
	for _, want := range []string{
		"没有所有主体都完整 ready 的维度",
		"部分 ready 的主体维度：甲公司(financials)、乙公司(news)",
		"未 ready 的主体及缺失维度：甲公司(news)、乙公司(financials)",
	} {
		if !strings.Contains(draft, want) {
			t.Fatalf("expected draft to contain %q, got %q", want, draft)
		}
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func containsRole(values []research.CompanyResearchTaskRole, want research.CompanyResearchTaskRole) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func payloadHasConflict(payload research.CompanyResearchPayload, code string) bool {
	if payload.TaskSummary == nil {
		return false
	}
	for _, conflict := range payload.TaskSummary.Conflicts {
		if conflict.Code == code {
			return true
		}
	}
	return false
}
