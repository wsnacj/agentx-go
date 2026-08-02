package companyresearch

import "testing"

func TestBuildCompanyResearchTaskPlanDoesNotRequireTickerForAliasInput(t *testing.T) {
	plan := BuildCompanyResearchTaskPlan(CompanyResearchTaskPlanInput{
		CaseType:   "company_research.single_company",
		WorkflowID: WorkflowID,
		Intent: CompanyResearchIntent{
			UserMessage:         "帮我查下富途牛牛最新财报、当前股价估值和最近新闻风险。",
			EntityName:          "富途牛牛",
			RequestedDimensions: []string{"financials", "market_data", "news", "risk"},
		},
	})
	if plan.Subject.InputTerm != "富途牛牛" {
		t.Fatalf("unexpected subject input: %#v", plan.Subject)
	}
	if plan.Subject.StockCode != "" || plan.Subject.Ticker != "" {
		t.Fatalf("task plan should not require ticker/code at entry, got %#v", plan.Subject)
	}
	if len(plan.Tasks) == 0 || plan.Tasks[0].Role != CompanyResearchRoleSubjectResolver {
		t.Fatalf("expected subject resolver as first task, got %#v", plan.Tasks)
	}
	for _, role := range []CompanyResearchTaskRole{
		CompanyResearchRoleFinanceAnalyst,
		CompanyResearchRoleMarketAnalyst,
		CompanyResearchRoleNewsAnalyst,
		CompanyResearchRoleRiskReviewer,
		CompanyResearchRoleEvidenceGuard,
		CompanyResearchRoleSynthesisEditor,
	} {
		if _, ok := plan.TaskByRole(role); !ok {
			t.Fatalf("missing task role %s in %#v", role, plan.Tasks)
		}
	}
	finance, _ := plan.TaskByRole(CompanyResearchRoleFinanceAnalyst)
	if len(finance.DependsOn) != 1 || finance.DependsOn[0] != "subject_resolution" {
		t.Fatalf("expected finance task to depend on subject resolution, got %#v", finance.DependsOn)
	}
}

func TestBuildCompanyResearchTaskPlanUsesVerifiedSubjectResolution(t *testing.T) {
	plan := BuildCompanyResearchTaskPlan(CompanyResearchTaskPlanInput{
		Intent: CompanyResearchIntent{
			UserMessage: "看下这个公司",
			EntityName:  "富途牛牛",
		},
		SubjectResolution: &SubjectResolution{
			AdapterStatus:   "ok",
			InputTerm:       "富途牛牛",
			PreferredMarket: "us",
			SelectedCandidate: &SubjectResolutionCandidate{
				EntityName:  "Futu Holdings Ltd",
				DisplayName: "Futu",
				Ticker:      "FUTU",
				Market:      "us",
				Source:      "public_search",
				EvidenceURL: "https://example.com/futu",
				Verified:    true,
				Confidence:  0.92,
			},
		},
	})
	if !plan.Subject.Verified {
		t.Fatalf("expected verified subject, got %#v", plan.Subject)
	}
	if plan.Subject.CanonicalName != "Futu Holdings Ltd" || plan.Subject.Ticker != "FUTU" || plan.Subject.MarketHint != "us" {
		t.Fatalf("unexpected resolved subject: %#v", plan.Subject)
	}
	result, ok := SubjectResolutionTaskResult(plan)
	if !ok || result.Status != CompanyResearchTaskStatusReady || !result.EvidenceReady {
		t.Fatalf("unexpected subject task result: %#v ok=%v", result, ok)
	}
}

func TestTaskResultFromEvidenceAggregatesMarketAdapters(t *testing.T) {
	plan := BuildCompanyResearchTaskPlan(CompanyResearchTaskPlanInput{
		Intent: CompanyResearchIntent{
			EntityName:          "样例公司",
			RequestedDimensions: []string{"market_data"},
		},
	})
	result, ok := TaskResultFromEvidence(
		plan,
		CompanyResearchRoleMarketAnalyst,
		map[string]any{"adapter_status": "unavailable", "failure_code": "identity_not_found"},
		map[string]any{"adapter_status": "ok", "quote": map[string]any{"price": "12.3"}},
	)
	if !ok {
		t.Fatal("expected market task result")
	}
	if result.Status != CompanyResearchTaskStatusReady || !result.EvidenceReady {
		t.Fatalf("expected one ready adapter to make market task ready, got %#v", result)
	}
}

func TestTaskResultFromReadinessAndAnswerContract(t *testing.T) {
	plan := BuildCompanyResearchTaskPlan(CompanyResearchTaskPlanInput{
		Intent: CompanyResearchIntent{
			EntityName:          "样例公司",
			RequestedDimensions: []string{"financials", "news", "risk"},
		},
	})
	guard, ok := TaskResultFromReadiness(plan, CompanyResearchRoleEvidenceGuard, CompanyResearchAnswerReadiness{
		AnswerReady:   false,
		SafeToAnswer:  true,
		Degraded:      true,
		DegradeReason: "missing_required_dimensions",
		FailureCode:   "company_research_missing_required_dimensions",
	})
	if !ok || guard.Status != CompanyResearchTaskStatusDegraded || guard.FailureCode == "" {
		t.Fatalf("unexpected guard result: %#v ok=%v", guard, ok)
	}
	synthesis, ok := TaskResultFromAnswerContract(plan, &CompanyResearchAnswerContract{
		FinalAnswerRecommended: true,
		AllowedSummaryScope:    "partial_company_research",
	})
	if !ok || synthesis.Status != CompanyResearchTaskStatusReady || !synthesis.EvidenceReady {
		t.Fatalf("unexpected synthesis result: %#v ok=%v", synthesis, ok)
	}
}
