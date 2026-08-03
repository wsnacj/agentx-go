package hostkit

import (
	"context"
	"testing"

	financereports "github.com/wsnacj/agentx-go/scenes/finance"
	financialreportmetrics "github.com/wsnacj/agentx-go/scenes/finance/metrics"
)

func TestFinanceReportLookupHandlerOrchestratesCandidatesMetricsAndBrief(t *testing.T) {
	var metricsSawSourceURL bool
	var briefCalled bool
	handler := BuildFinanceReportLookupHandler(FinanceReportLookupConfig{
		Source:              "host_test_lookup",
		SourcePolicyDefault: "host_policy",
		Handlers: FinanceReportLookupHandlers{
			Candidates: func(_ context.Context, params map[string]any) (financereports.MetricsCandidatesPayload, error) {
				if params["source_policy"] != "host_policy" {
					t.Fatalf("expected source policy default before candidates, got %#v", params["source_policy"])
				}
				return financereports.MetricsCandidatesPayload{
					Tool:            financereports.ToolReportMetricsCandidates,
					AdapterID:       "candidate_adapter",
					AdapterStatus:   "ok",
					ResolvedCompany: "Tencent Holdings Ltd.",
					ResolvedCode:    "00700.HK",
					ResolvedMarket:  "HK",
					PrimaryURL:      "https://www.tencent.com/en-us/investors/financial-news.html",
					ResolvedEntities: []financereports.ResolvedEntityCandidate{{
						EntityName:   "Tencent Holdings Ltd.",
						CodeOrTicker: "00700.HK",
						Market:       "HK",
						Source:       "official_ir",
						EvidenceURL:  "https://www.tencent.com/en-us/investors/financial-news.html",
						Confidence:   0.93,
						MatchReason:  "matched_official_ir_profile",
					}},
					Candidates: []financereports.MetricsCandidate{{
						URL:                 "https://www.tencent.com/en-us/investors/financial-news.html",
						SourceKind:          "official_ir",
						PreferredNextAction: "open_page",
					}},
				}, nil
			},
			MetricsGuard: func(_ context.Context, params map[string]any) (financereports.MetricsToolPayload, error) {
				metricsSawSourceURL = params["source_url"] == "https://www.tencent.com/en-us/investors/financial-news.html"
				return financereports.MetricsToolPayload{
					Tool:                 financereports.ToolReportMetricsGuard,
					AdapterID:            "metrics_guard",
					AdapterStatus:        "ok",
					GuardStatus:          "passed",
					RequestedFieldsReady: true,
					Evidence: financereports.MetricsEvidence{
						CompanyName:  "Tencent Holdings Ltd.",
						StockCode:    "00700.HK",
						ReportPeriod: "2026-03-31",
						Revenue:      "1964.58亿元",
						NetProfit:    "580.93亿元",
					},
				}, nil
			},
			BriefGuard: func(_ context.Context, params map[string]any) (financereports.BriefToolPayload, error) {
				briefCalled = true
				if params["ticker"] != "00700.HK" {
					t.Fatalf("expected candidate ticker enrichment before brief, got %#v", params["ticker"])
				}
				return financereports.BriefToolPayload{
					Tool:          financereports.ToolReportBriefGuard,
					AdapterID:     "brief_guard",
					AdapterStatus: "ok",
					GuardStatus:   "passed",
					BriefReady:    true,
				}, nil
			},
		},
	})

	raw, err := handler(context.Background(), map[string]any{
		"user_message":      "tell me Tencent latest report",
		"entity_name":       "Tencent",
		"task_kind":         "latest_published_report_brief",
		"report_kind":       "earnings_release",
		"requested_outputs": []any{"metrics", "brief"},
		"freshness": map[string]any{
			"mode":               "latest_published",
			"relative_date_hint": "yesterday",
		},
	})
	if err != nil {
		t.Fatalf("lookup handler returned error: %v", err)
	}
	payload, ok := raw.(financereports.FinanceReportLookupPayload)
	if !ok {
		t.Fatalf("expected FinanceReportLookupPayload, got %T", raw)
	}
	if payload.Source != "host_test_lookup" ||
		payload.AdapterID != "metrics_guard" ||
		payload.GuardStatus != "passed" ||
		payload.Intent.EntityName != "Tencent" ||
		payload.Intent.StockCode != "00700.HK" ||
		payload.Intent.Freshness["relative_date_hint"] != "yesterday" {
		t.Fatalf("unexpected lookup payload: %#v", payload)
	}
	if payload.Candidates == nil || payload.Candidates.PrimaryURL == "" || payload.Metrics == nil || payload.Brief == nil {
		t.Fatalf("expected candidates, metrics, and brief payloads: %#v", payload)
	}
	if payload.IdentityResolution == nil ||
		payload.IdentityResolution.SelectedCandidate == nil ||
		payload.IdentityResolution.SelectedCandidate.CodeOrTicker != "00700.HK" ||
		payload.Candidates.IdentityResolution == nil {
		t.Fatalf("expected identity resolution projection, got lookup=%#v candidates=%#v", payload.IdentityResolution, payload.Candidates)
	}
	if payload.AnswerReady == nil || !payload.AnswerReady.AnswerReady || payload.AnswerReady.AllowedSummaryScope != financereports.AnswerScopeRequested {
		t.Fatalf("expected passed answer readiness projection, got %#v", payload.AnswerReady)
	}
	if !metricsSawSourceURL || !briefCalled {
		t.Fatalf("expected enriched metrics params and brief execution, metricsSawSourceURL=%v briefCalled=%v", metricsSawSourceURL, briefCalled)
	}
}

func TestFinanceReportLookupHandlerUsesStructuredAssessmentWithoutBrief(t *testing.T) {
	briefCalled := false
	handler := BuildFinanceReportLookupHandler(FinanceReportLookupConfig{
		Handlers: FinanceReportLookupHandlers{
			Candidates: func(context.Context, map[string]any) (financereports.MetricsCandidatesPayload, error) {
				return financereports.MetricsCandidatesPayload{
					Tool:          financereports.ToolReportMetricsCandidates,
					AdapterID:     "candidate_adapter",
					AdapterStatus: "ok",
				}, nil
			},
			MetricsGuard: func(_ context.Context, params map[string]any) (financereports.MetricsToolPayload, error) {
				assessment, _ := params["assessment"].(map[string]any)
				if assessment["kind"] != financialreportmetrics.AssessmentKindBusinessPerformance {
					t.Fatalf("expected structured business assessment params, got %#v", assessment)
				}
				return financereports.MetricsToolPayload{
					Tool:          financereports.ToolReportMetricsGuard,
					AdapterID:     "metrics_guard",
					AdapterStatus: "ok",
					GuardStatus:   "passed",
				}, nil
			},
			BriefGuard: func(context.Context, map[string]any) (financereports.BriefToolPayload, error) {
				briefCalled = true
				return financereports.BriefToolPayload{}, nil
			},
		},
	})

	raw, err := handler(context.Background(), map[string]any{
		"user_message": "assess performance",
		"task_kind":    "business_performance_assessment",
	})
	if err != nil {
		t.Fatalf("lookup handler returned error: %v", err)
	}
	payload := raw.(financereports.FinanceReportLookupPayload)
	if payload.Intent.Assessment["kind"] != financialreportmetrics.AssessmentKindBusinessPerformance ||
		len(payload.Intent.RequestedOutputs) != 2 ||
		payload.Brief != nil ||
		briefCalled {
		t.Fatalf("unexpected assessment lookup payload: %#v briefCalled=%v", payload, briefCalled)
	}
}

func TestFinanceReportLookupHandlerCanRunExtractBeforeGuard(t *testing.T) {
	handler := BuildFinanceReportLookupHandler(FinanceReportLookupConfig{
		Handlers: FinanceReportLookupHandlers{
			MetricsExtract: func(context.Context, map[string]any) (financereports.MetricsToolPayload, error) {
				return financereports.MetricsToolPayload{
					Tool:          financereports.ToolReportMetricsExtract,
					AdapterID:     "extract",
					AdapterStatus: "ok",
					Evidence: financereports.MetricsEvidence{
						CompanyName: "Example Inc.",
						Revenue:     "100",
						NetProfit:   "20",
					},
				}, nil
			},
			MetricsGuard: func(_ context.Context, params map[string]any) (financereports.MetricsToolPayload, error) {
				if params["company_name"] != "Example Inc." || params["revenue"] != "100" {
					t.Fatalf("guard did not receive extracted evidence params: %#v", params)
				}
				return financereports.MetricsToolPayload{
					Tool:          financereports.ToolReportMetricsGuard,
					AdapterID:     "guard",
					AdapterStatus: "ok",
					GuardStatus:   "passed",
				}, nil
			},
		},
	})
	raw, err := handler(context.Background(), map[string]any{"user_message": "example metrics"})
	if err != nil {
		t.Fatalf("lookup handler returned error: %v", err)
	}
	payload := raw.(financereports.FinanceReportLookupPayload)
	if payload.AdapterID != "guard" || payload.GuardStatus != "passed" {
		t.Fatalf("unexpected lookup payload: %#v", payload)
	}
}

func TestFinanceReportLookupHandlerProjectsTerminalAnswerContract(t *testing.T) {
	handler := BuildFinanceReportLookupHandler(FinanceReportLookupConfig{
		Handlers: FinanceReportLookupHandlers{
			Candidates: func(context.Context, map[string]any) (financereports.MetricsCandidatesPayload, error) {
				return financereports.MetricsCandidatesPayload{
					Tool:            financereports.ToolReportMetricsCandidates,
					AdapterID:       "exchange_a_share_official_disclosure",
					AdapterStatus:   "needs_review",
					ResolvedCompany: "瑞可达",
					ResolvedCode:    "688800.SH",
					PrimaryURL:      "https://www.sse.com.cn/disclosure/listedinfo/announcement/c/new/2026-04-21/688800_20260421_11VH.pdf",
				}, nil
			},
			MetricsGuard: func(context.Context, map[string]any) (financereports.MetricsToolPayload, error) {
				return financereports.MetricsToolPayload{
					Tool:                        financereports.ToolReportMetricsGuard,
					AdapterID:                   "exchange_a_share_official_disclosure",
					AdapterStatus:               "needs_review",
					GuardStatus:                 "missing_requested_fields",
					RequestedFieldsReady:        false,
					MissingRequestedFields:      []string{"revenue", "net_profit"},
					ReviewRequiredFields:        []string{"revenue", "net_profit"},
					RequestedFields:             []string{"revenue", "net_profit"},
					FailureCode:                 "review_required_fields",
					Warnings:                    []string{"exchange_a_share_pdf_download_returned_non_pdf:official ir artifact is not a valid pdf after curl fallback (content_type=text/html artifact_kind=html)"},
					SourcePolicy:                "official_source_only",
					AssessmentRequiresValuation: false,
					Evidence: financereports.MetricsEvidence{
						CompanyName:    "瑞可达",
						StockCode:      "688800.SH",
						OfficialSource: "https://www.sse.com.cn/disclosure/listedinfo/announcement/c/new/2026-04-21/688800_20260421_11VH.pdf",
					},
				}, nil
			},
		},
	})
	raw, err := handler(context.Background(), map[string]any{
		"user_message":      "请基于上交所官方披露文件查一下瑞可达688800最新年报的营收和净利润。",
		"entity_name":       "瑞可达",
		"stock_code":        "688800",
		"requested_metrics": []any{"revenue", "net_profit"},
	})
	if err != nil {
		t.Fatalf("lookup handler returned error: %v", err)
	}
	payload := raw.(financereports.FinanceReportLookupPayload)
	if payload.AnswerReady == nil || !payload.AnswerReady.StopRecommended {
		t.Fatalf("expected stop recommendation, got %#v", payload.AnswerReady)
	}
	if payload.AnswerContract == nil ||
		!payload.AnswerContract.FinalAnswerRecommended ||
		payload.AnswerContract.Reason != financereports.AnswerDegradeSourceReturnedNonPDF {
		t.Fatalf("expected terminal answer contract, got %#v", payload.AnswerContract)
	}
}
