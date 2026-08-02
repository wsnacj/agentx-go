package hostkit

import (
	"context"
	"strings"
	"testing"

	newsbrief "github.com/wsnacj/agentx-go/scenes/publicnews"
)

func TestParamsWithExtractReplacesOnlyInvalidSourceKeyUpdate(t *testing.T) {
	headline := "骂了大半年，马斯克突然改口"
	extracted := "马斯克表示自己此前看错Anthropic，并称其当前模型处于AI行业领先位置。"
	params := paramsWithExtract(map[string]any{
		"headline":   headline,
		"key_update": "换句话说，马斯克要是想搞垮Anthropic，随时能做到。",
	}, newsbrief.Payload{Evidence: newsbrief.Evidence{
		Headline:  headline,
		KeyUpdate: extracted,
	}})
	if got := newsbrief.StringArg(params["key_update"]); got != extracted {
		t.Fatalf("key_update = %q, want extracted factual update %q", got, extracted)
	}

	current := "马斯克确认自己此前看错Anthropic，并公开称赞其模型能力。"
	params = paramsWithExtract(map[string]any{
		"headline":   headline,
		"key_update": current,
	}, newsbrief.Payload{Evidence: newsbrief.Evidence{
		Headline:  headline,
		KeyUpdate: extracted,
	}})
	if got := newsbrief.StringArg(params["key_update"]); got != current {
		t.Fatalf("key_update = %q, want existing valid update %q", got, current)
	}
}

func TestLatestNewsLookupUsesHostSourcesAndGuard(t *testing.T) {
	payload, err := BuildLatestNewsLookupPayload(context.Background(), LatestNewsLookupConfig{
		Source:              "host_test_latest_news_lookup",
		SourcePolicyDefault: "host_public_news_policy",
		AnswerContract:      true,
		Handlers: LatestNewsLookupHandlers{
			Sources: func(_ context.Context, _ map[string]any, intent newsbrief.LatestNewsLookupIntent) (newsbrief.LatestNewsSourcesPayload, error) {
				return newsbrief.LatestNewsSourcesPayload{
					Source:        "host_test_latest_news_sources",
					AdapterID:     "host_news_sources",
					AdapterStatus: "ok",
					UserMessage:   intent.UserMessage,
					Intent:        intent,
					PrimarySource: newsbrief.LatestNewsLookupSource{
						Title:     "示例事件最新进展",
						SourceURL: "https://news.example.com/a",
						Text:      "发布时间: 2026-05-15 09:00 UTC\n示例事件出现最新进展。",
					},
					SupportingSources: []newsbrief.LatestNewsLookupSource{{
						Title:     "第二来源确认示例事件",
						SourceURL: "https://wire.example.net/b",
						Text:      "发布时间: 2026-05-15 09:10 UTC\n第二来源确认示例事件出现最新进展。",
					}},
				}, nil
			},
		},
	}, map[string]any{
		"user_message":      "帮我看下示例事件最新新闻",
		"task_kind":         "latest_news_brief",
		"topic":             "示例事件",
		"requested_outputs": []any{"brief", "source_verification"},
	})
	if err != nil {
		t.Fatalf("build lookup payload: %v", err)
	}
	if payload.Tool != newsbrief.ToolLatestNewsLookup ||
		payload.Source != "host_test_latest_news_lookup" ||
		payload.AdapterID != "host_news_sources" ||
		payload.GuardStatus != "passed" ||
		!payload.Passed ||
		!payload.NewsFieldsReady ||
		!payload.CrossCheckReady ||
		payload.SourceURL != "https://news.example.com/a" ||
		payload.Guard == nil ||
		!payload.Guard.CrossCheckReady ||
		payload.Intent.Topic != "示例事件" ||
		payload.Intent.SourcePolicy != "host_public_news_policy" {
		t.Fatalf("unexpected lookup payload: %#v", payload)
	}
	if payload.MissingNewsFields == nil || payload.ReviewReasons == nil || len(payload.MissingNewsFields) != 0 || len(payload.ReviewReasons) != 0 {
		t.Fatalf("expected stable empty top-level review arrays on success, got missing=%#v review=%#v", payload.MissingNewsFields, payload.ReviewReasons)
	}
	if payload.EvaluatorReport == nil ||
		payload.EvaluatorReport.Degraded ||
		payload.EvaluatorReport.FreshnessMatch != "confirmed" ||
		payload.EvaluatorReport.SourceIndependence != "cross_checked" ||
		payload.EvaluatorReport.GroundedTextPresence != "primary_and_supporting" ||
		payload.EvaluatorReport.StopReason != "guard_passed" {
		t.Fatalf("expected passing evaluator report, got %#v", payload.EvaluatorReport)
	}
	if payload.AnswerContract == nil ||
		!payload.AnswerContract.FinalAnswerRecommended ||
		payload.AnswerContract.Reason != "guard_passed" {
		t.Fatalf("expected guarded answer contract when enabled, got %#v", payload.AnswerContract)
	}
	if !payload.AnswerReadiness.AnswerReady || !payload.AnswerReadiness.SafeToAnswer || payload.AnswerReadiness.Degraded {
		t.Fatalf("expected full answer readiness on guarded success, got %#v", payload.AnswerReadiness)
	}
}

func TestLatestNewsLookupProjectsEvidenceReview(t *testing.T) {
	payload, err := BuildLatestNewsLookupPayload(context.Background(), LatestNewsLookupConfig{
		AnswerContract: true,
		EvidenceReviewer: newsbrief.EvidenceReviewerFunc(func(context.Context, newsbrief.EvidenceReviewInput) (newsbrief.EvidenceReviewResult, error) {
			return newsbrief.EvidenceReviewResult{
				Accepted:       false,
				RequiresReview: true,
				ReasonCodes:    []string{"page_noise"},
				Confidence:     0.8,
				Source:         "test_reviewer",
			}, nil
		}),
		Handlers: LatestNewsLookupHandlers{
			Sources: func(_ context.Context, _ map[string]any, intent newsbrief.LatestNewsLookupIntent) (newsbrief.LatestNewsSourcesPayload, error) {
				return newsbrief.LatestNewsSourcesPayload{
					AdapterStatus: "ok",
					UserMessage:   intent.UserMessage,
					Intent:        intent,
					PrimarySource: newsbrief.LatestNewsLookupSource{
						Title:       "示例事件最新进展",
						SourceURL:   "https://news.example.com/a",
						PublishedAt: "2026-05-15T09:00:00",
						KeyUpdate:   "示例事件出现最新进展。",
						Text:        "发布时间: 2026-05-15 09:00 UTC\n示例事件出现最新进展。",
					},
					SupportingSources: []newsbrief.LatestNewsLookupSource{{
						Title:       "第二来源确认示例事件",
						SourceURL:   "https://wire.example.net/b",
						PublishedAt: "2026-05-15T09:10:00",
						KeyUpdate:   "第二来源确认示例事件出现最新进展。",
						Text:        "发布时间: 2026-05-15 09:10 UTC\n第二来源确认示例事件出现最新进展。",
					}},
				}, nil
			},
		},
	}, map[string]any{
		"user_message": "帮我看下示例事件最新新闻",
	})
	if err != nil {
		t.Fatalf("build lookup payload: %v", err)
	}
	if payload.EvidenceReview == nil || payload.Guard == nil || payload.Guard.EvidenceReview == nil {
		t.Fatalf("expected evidence review on lookup and guard payload, got %#v", payload)
	}
	if payload.GuardStatus != "needs_cross_check" || payload.Passed || payload.AnswerContract == nil ||
		payload.AnswerContract.Reason != "source_quality_needs_review" {
		t.Fatalf("expected rejected evidence review to degrade lookup, got %#v", payload)
	}
	if payload.AnswerReadiness.AnswerReady || !payload.AnswerReadiness.SafeToAnswer || !payload.AnswerReadiness.Degraded ||
		len(payload.AnswerReadiness.ReadyDimensions) == 0 || len(payload.AnswerReadiness.MissingDimensions) == 0 {
		t.Fatalf("expected bounded answer readiness on rejected evidence, got %#v", payload.AnswerReadiness)
	}
	if len(payload.ReviewReasons) == 0 || payload.ReviewReasons[0] != "evidence_review:page_noise" {
		t.Fatalf("expected projected evidence review reason, got %#v", payload.ReviewReasons)
	}
}

func TestLatestNewsLookupCarriesSourceMetadataIntoGuard(t *testing.T) {
	payload, err := BuildLatestNewsLookupPayload(context.Background(), LatestNewsLookupConfig{
		AnswerContract: true,
		Handlers: LatestNewsLookupHandlers{
			Sources: func(_ context.Context, _ map[string]any, intent newsbrief.LatestNewsLookupIntent) (newsbrief.LatestNewsSourcesPayload, error) {
				return newsbrief.LatestNewsSourcesPayload{
					AdapterID:     "host_news_sources",
					AdapterStatus: "ok",
					UserMessage:   intent.UserMessage,
					Intent:        intent,
					PrimarySource: newsbrief.LatestNewsLookupSource{
						Title:       "示例事件最新进展",
						SourceURL:   "https://news.example.com/a",
						PublishedAt: "2026-05-15T09:00:00",
						Text:        "示例事件出现最新进展。",
					},
					SupportingSources: []newsbrief.LatestNewsLookupSource{{
						Title:       "第二来源确认示例事件",
						SourceURL:   "https://wire.example.net/b",
						PublishedAt: "2026-05-15T09:10:00",
						Text:        "第二来源确认示例事件出现最新进展。",
					}},
				}, nil
			},
		},
	}, map[string]any{
		"user_message": "帮我看下示例事件最新新闻",
	})
	if err != nil {
		t.Fatalf("build lookup payload: %v", err)
	}
	if payload.Guard == nil || payload.Guard.GuardStatus != "passed" {
		t.Fatalf("expected metadata-backed guard pass, got %#v", payload)
	}
	if payload.Guard.Evidence.PublishedAt != "2026-05-15T09:00:00" {
		t.Fatalf("expected primary published metadata to reach guard, got %#v", payload.Guard.Evidence)
	}
	if len(payload.Guard.SupportingSources) != 1 || payload.Guard.SupportingSources[0].PublishedAt != "2026-05-15T09:10:00" {
		t.Fatalf("expected supporting published metadata to reach guard, got %#v", payload.Guard.SupportingSources)
	}
	if latestNewsLookupWarningContains(payload.Warnings, "published_at_missing") {
		t.Fatalf("did not expect stale published_at_missing warning after metadata-backed guard pass, got %#v", payload.Warnings)
	}
	if payload.EvaluatorReport != nil && latestNewsLookupWarningContains(payload.EvaluatorReport.Warnings, "published_at_missing") {
		t.Fatalf("did not expect stale evaluator warning after metadata-backed guard pass, got %#v", payload.EvaluatorReport.Warnings)
	}
}

func TestLatestNewsLookupDoesNotTreatSearchSnippetPrimaryAsGrounded(t *testing.T) {
	payload, err := BuildLatestNewsLookupPayload(context.Background(), LatestNewsLookupConfig{
		AnswerContract: true,
		Handlers: LatestNewsLookupHandlers{
			Sources: func(_ context.Context, _ map[string]any, intent newsbrief.LatestNewsLookupIntent) (newsbrief.LatestNewsSourcesPayload, error) {
				return newsbrief.LatestNewsSourcesPayload{
					AdapterID:     "host_news_sources",
					AdapterStatus: "ok",
					UserMessage:   intent.UserMessage,
					Intent:        intent,
					Warnings:      []string{"latest_news_search_snippet_primary_used"},
					PrimarySource: newsbrief.LatestNewsLookupSource{
						Title:       "示例事件最新进展",
						SourceURL:   "https://news.example.com/a",
						PublishedAt: "2026-05-15T09:00:00",
						KeyUpdate:   "示例事件出现最新进展。",
						Text:        "搜索摘要不应作为主来源 grounded page text。",
					},
					SupportingSources: []newsbrief.LatestNewsLookupSource{{
						Title:       "第二来源确认示例事件",
						SourceURL:   "https://wire.example.net/b",
						PublishedAt: "2026-05-15T09:10:00",
						Text:        "第二来源确认示例事件出现最新进展。",
					}},
				}, nil
			},
		},
	}, map[string]any{
		"user_message": "帮我看下示例事件最新新闻",
		"page_id":      "stale_page_from_original_args",
		"text":         "原始入参正文也不能覆盖 snippet-primary grounding guard。",
	})
	if err != nil {
		t.Fatalf("build lookup payload: %v", err)
	}
	if payload.Guard == nil || payload.Guard.GuardStatus == "passed" || payload.CrossCheckReady {
		t.Fatalf("expected snippet-primary evidence not to pass guard, got %#v", payload)
	}
	if payload.AnswerContract == nil ||
		payload.AnswerContract.Reason != "source_quality_needs_review" ||
		payload.AnswerContract.AllowedSummaryScope != newsbrief.LatestNewsAnswerScopeSourceDiagnostic {
		t.Fatalf("expected source-quality diagnostic contract for snippet-primary evidence, got %#v", payload.AnswerContract)
	}
	if len(payload.Guard.MissingNewsFields) == 0 || payload.Guard.MissingNewsFields[0] != "grounded_page_text" {
		t.Fatalf("expected grounded_page_text gap, got %#v", payload.Guard.MissingNewsFields)
	}
	if payload.EvaluatorReport == nil ||
		!payload.EvaluatorReport.Degraded ||
		payload.EvaluatorReport.GroundedTextPresence != "supporting_only" ||
		payload.EvaluatorReport.DegradeReason != "latest_news_missing_fields" {
		t.Fatalf("expected evaluator report to expose snippet-primary grounding gap, got %#v", payload.EvaluatorReport)
	}
}

func TestLatestNewsLookupReportsMissingHostSources(t *testing.T) {
	payload, err := BuildLatestNewsLookupPayload(context.Background(), LatestNewsLookupConfig{}, map[string]any{
		"user_message": "帮我看下示例事件最新新闻",
	})
	if err != nil {
		t.Fatalf("build lookup payload: %v", err)
	}
	if payload.AdapterStatus != "unsupported" ||
		payload.FailureCode != "latest_news_source_adapter_not_configured" ||
		payload.GuardStatus != "missing_news_fields" ||
		!payload.Intent.NeedsSourceVerify {
		t.Fatalf("expected unsupported source diagnostics, got %#v", payload)
	}
	if payload.EvaluatorReport == nil ||
		!payload.EvaluatorReport.Degraded ||
		payload.EvaluatorReport.StopReason != "latest_news_source_adapter_not_configured" ||
		payload.EvaluatorReport.SourceIndependence != "no_source" ||
		payload.EvaluatorReport.GroundedTextPresence != "missing" {
		t.Fatalf("expected unsupported evaluator report, got %#v", payload.EvaluatorReport)
	}
}

func TestLatestNewsLookupStopsAfterProviderUnavailableSources(t *testing.T) {
	payload, err := BuildLatestNewsLookupPayload(context.Background(), LatestNewsLookupConfig{
		AnswerContract: true,
		Handlers: LatestNewsLookupHandlers{
			Sources: func(_ context.Context, _ map[string]any, intent newsbrief.LatestNewsLookupIntent) (newsbrief.LatestNewsSourcesPayload, error) {
				return newsbrief.LatestNewsSourcesPayload{
					Source:            "host_test_latest_news_sources",
					AdapterID:         "host_search_provider",
					AdapterStatus:     "provider_unavailable",
					FailureCode:       "search_provider_failure_missing",
					ErrorClass:        "provider_unavailable",
					Provider:          "brave",
					EffectiveProvider: "brave",
					ProviderStatus:    "missing_credentials",
					FallbackHint:      "configure BRAVE_API_KEY or provide a direct URL",
					Retryable:         true,
					UserMessage:       intent.UserMessage,
					Intent:            intent,
				}, nil
			},
		},
	}, map[string]any{
		"user_message": "帮我看下示例事件最新新闻",
	})
	if err != nil {
		t.Fatalf("build lookup payload: %v", err)
	}
	if payload.AdapterStatus != "provider_unavailable" ||
		payload.FailureCode != "search_provider_failure_missing" ||
		payload.FailureClass != "auth_missing" ||
		payload.GuardStatus != "provider_unavailable" ||
		payload.NewsFieldsReady ||
		payload.CrossCheckReady ||
		payload.Passed ||
		payload.Sources == nil ||
		payload.Sources.ProviderStatus != "missing_credentials" ||
		payload.Sources.Retryable ||
		payload.Sources.RetrySuppressedReason != "terminal_failure_class:auth_missing" ||
		payload.RetrySuppressedReason != "terminal_failure_class:auth_missing" ||
		payload.Sources.RetryAttemptCount != 1 ||
		payload.Extract != nil ||
		payload.Guard != nil {
		t.Fatalf("expected provider-unavailable lookup to stop before extract/guard, got %#v", payload)
	}
	if payload.AnswerContract == nil ||
		!payload.AnswerContract.FinalAnswerRecommended ||
		payload.AnswerContract.Reason != "search_provider_config_invalid" ||
		!containsString(payload.AnswerContract.DoNotRetryTools, newsbrief.ToolLatestNewsLookup) ||
		!containsString(payload.AnswerContract.DoNotRetryTools, "search") {
		t.Fatalf("expected provider-unavailable answer contract when enabled, got %#v", payload.AnswerContract)
	}
	if payload.EvaluatorReport == nil ||
		!payload.EvaluatorReport.Degraded ||
		payload.EvaluatorReport.StopReason != "search_provider_failure_missing" ||
		payload.EvaluatorReport.DegradeReason != "search_provider_failure_missing" ||
		payload.EvaluatorReport.SourceIndependence != "no_source" {
		t.Fatalf("expected provider-unavailable evaluator report, got %#v", payload.EvaluatorReport)
	}
}

func TestLatestNewsLookupRetriesTransientProviderFailure(t *testing.T) {
	calls := 0
	payload, err := BuildLatestNewsLookupPayload(context.Background(), LatestNewsLookupConfig{
		AnswerContract: true,
		Handlers: LatestNewsLookupHandlers{
			Sources: func(_ context.Context, _ map[string]any, intent newsbrief.LatestNewsLookupIntent) (newsbrief.LatestNewsSourcesPayload, error) {
				calls++
				if calls == 1 {
					return newsbrief.LatestNewsSourcesPayload{
						AdapterStatus:  "provider_execution_failed",
						FailureCode:    "search_provider_status_503",
						ErrorClass:     "provider_execution_failed",
						Provider:       "brave",
						ProviderStatus: "status_503",
						Retryable:      true,
						UserMessage:    intent.UserMessage,
						Intent:         intent,
					}, nil
				}
				return newsbrief.LatestNewsSourcesPayload{
					AdapterStatus: "ok",
					UserMessage:   intent.UserMessage,
					Intent:        intent,
					PrimarySource: newsbrief.LatestNewsLookupSource{
						Title:       "示例事件最新进展",
						SourceURL:   "https://news.example.com/a",
						PublishedAt: "2026-05-15T09:00:00",
						Text:        "发布时间: 2026-05-15 09:00 UTC\n示例事件出现最新进展。",
					},
					SupportingSources: []newsbrief.LatestNewsLookupSource{{
						Title:       "第二来源确认示例事件",
						SourceURL:   "https://wire.example.net/b",
						PublishedAt: "2026-05-15T09:10:00",
						Text:        "发布时间: 2026-05-15 09:10 UTC\n第二来源确认示例事件出现最新进展。",
					}},
				}, nil
			},
		},
	}, map[string]any{"user_message": "帮我看下示例事件最新新闻"})
	if err != nil {
		t.Fatalf("build lookup payload: %v", err)
	}
	if calls != 2 {
		t.Fatalf("expected one transient retry, got calls=%d", calls)
	}
	if !payload.Passed || payload.GuardStatus != "passed" || payload.FailureClass != "" {
		t.Fatalf("expected retry to recover to passed lookup, got %#v", payload)
	}
	if payload.RetryAttemptCount != 1 ||
		len(payload.RetryAttempts) != 1 ||
		payload.RetryAttempts[0].FailureClass != "transient_network" ||
		!payload.RetryAttempts[0].Retryable ||
		payload.RetryExhausted {
		t.Fatalf("expected retry attempt diagnostics, got attempts=%#v count=%d exhausted=%t", payload.RetryAttempts, payload.RetryAttemptCount, payload.RetryExhausted)
	}
}

func TestLatestNewsLookupRetriesProviderFailureEvenWhenAdapterStatusOK(t *testing.T) {
	calls := 0
	payload, err := BuildLatestNewsLookupPayload(context.Background(), LatestNewsLookupConfig{
		Handlers: LatestNewsLookupHandlers{
			Sources: func(_ context.Context, _ map[string]any, intent newsbrief.LatestNewsLookupIntent) (newsbrief.LatestNewsSourcesPayload, error) {
				calls++
				if calls == 1 {
					return newsbrief.LatestNewsSourcesPayload{
						AdapterStatus:  "ok",
						FailureCode:    "search_provider_status_503",
						Provider:       "brave",
						ProviderStatus: "status_503",
						Retryable:      true,
						UserMessage:    intent.UserMessage,
						Intent:         intent,
					}, nil
				}
				return newsbrief.LatestNewsSourcesPayload{
					AdapterStatus: "ok",
					UserMessage:   intent.UserMessage,
					Intent:        intent,
					PrimarySource: newsbrief.LatestNewsLookupSource{
						Title:       "示例事件最新进展",
						SourceURL:   "https://news.example.com/a",
						PublishedAt: "2026-05-15T09:00:00",
						Text:        "发布时间: 2026-05-15 09:00 UTC\n示例事件出现最新进展。",
					},
					SupportingSources: []newsbrief.LatestNewsLookupSource{{
						Title:       "第二来源确认示例事件",
						SourceURL:   "https://wire.example.net/b",
						PublishedAt: "2026-05-15T09:10:00",
						Text:        "发布时间: 2026-05-15 09:10 UTC\n第二来源确认示例事件出现最新进展。",
					}},
				}, nil
			},
		},
	}, map[string]any{"user_message": "帮我看下示例事件最新新闻"})
	if err != nil {
		t.Fatalf("build lookup payload: %v", err)
	}
	if calls != 2 || !payload.Passed || payload.RetryAttemptCount != 1 ||
		payload.RetryAttempts[0].FailureClass != "transient_network" {
		t.Fatalf("expected provider-status failure to retry before success, calls=%d payload=%#v", calls, payload)
	}
}

func TestLatestNewsLookupIgnoresUnavailableProviderDiagnosticsAfterRecovery(t *testing.T) {
	payload, err := BuildLatestNewsLookupPayload(context.Background(), LatestNewsLookupConfig{
		AnswerContract: true,
		Handlers: LatestNewsLookupHandlers{
			Sources: func(_ context.Context, _ map[string]any, intent newsbrief.LatestNewsLookupIntent) (newsbrief.LatestNewsSourcesPayload, error) {
				return newsbrief.LatestNewsSourcesPayload{
					AdapterStatus:     "ok",
					Provider:          "ark",
					EffectiveProvider: "ark",
					ProviderStatus:    "available",
					UnavailableProviders: []newsbrief.ProviderStatus{
						{Provider: "baidu", Reason: "rate_limited"},
					},
					UserMessage: intent.UserMessage,
					Intent:      intent,
					PrimarySource: newsbrief.LatestNewsLookupSource{
						Title:       "示例事件最新进展",
						SourceURL:   "https://news.example.com/a",
						SourceSite:  "news.example.com",
						PublishedAt: "2026-05-15T09:00:00",
						KeyUpdate:   "示例事件出现最新进展。",
						Text:        "发布时间: 2026-05-15 09:00 UTC\n示例事件出现最新进展。",
					},
					SupportingSources: []newsbrief.LatestNewsLookupSource{{
						Title:       "独立来源确认示例事件",
						SourceURL:   "https://wire.example.net/b",
						SourceSite:  "wire.example.net",
						PublishedAt: "2026-05-15T09:10:00",
						KeyUpdate:   "独立来源确认示例事件出现最新进展。",
						Text:        "发布时间: 2026-05-15 09:10 UTC\n独立来源确认示例事件出现最新进展。",
					}},
				}, nil
			},
		},
	}, map[string]any{"user_message": "帮我看下示例事件最新新闻", "topic": "示例事件"})
	if err != nil {
		t.Fatalf("build lookup payload: %v", err)
	}
	if !payload.Passed ||
		payload.FailureClass != "" ||
		payload.RetryAttemptCount != 0 ||
		payload.AnswerContract == nil ||
		!payload.AnswerContract.FinalAnswerRecommended ||
		payload.AnswerContract.Reason != "guard_passed" ||
		payload.Sources == nil ||
		payload.Sources.FailureClass != "" ||
		len(payload.Sources.UnavailableProviders) != 1 ||
		payload.Sources.UnavailableProviders[0].Reason != "rate_limited" {
		t.Fatalf("expected recovered provider diagnostics to stay non-failing, got %#v", payload)
	}
}

func TestLatestNewsLookupIgnoresFallbackHintWhenEffectiveProviderHasGroundedSources(t *testing.T) {
	payload, err := BuildLatestNewsLookupPayload(context.Background(), LatestNewsLookupConfig{
		AnswerContract: true,
		Handlers: LatestNewsLookupHandlers{
			Sources: func(_ context.Context, _ map[string]any, intent newsbrief.LatestNewsLookupIntent) (newsbrief.LatestNewsSourcesPayload, error) {
				return newsbrief.LatestNewsSourcesPayload{
					AdapterStatus:     "ok",
					Provider:          "ark",
					EffectiveProvider: "ark",
					ProviderStatus:    "available",
					FallbackProvider:  "brave",
					FallbackHint:      "credential_invalid",
					UserMessage:       intent.UserMessage,
					Intent:            intent,
					PrimarySource: newsbrief.LatestNewsLookupSource{
						Title:       "示例事件最新进展",
						SourceURL:   "https://news.example.com/a",
						SourceSite:  "news.example.com",
						PublishedAt: "2026-05-15T09:00:00",
						KeyUpdate:   "示例事件出现最新进展。",
						Text:        "发布时间: 2026-05-15 09:00 UTC\n示例事件出现最新进展。",
					},
					SupportingSources: []newsbrief.LatestNewsLookupSource{{
						Title:       "独立来源确认示例事件",
						SourceURL:   "https://wire.example.net/b",
						SourceSite:  "wire.example.net",
						PublishedAt: "2026-05-15T09:10:00",
						KeyUpdate:   "独立来源确认示例事件出现最新进展。",
						Text:        "发布时间: 2026-05-15 09:10 UTC\n独立来源确认示例事件出现最新进展。",
					}},
				}, nil
			},
		},
	}, map[string]any{"user_message": "帮我看下示例事件最新新闻", "topic": "示例事件"})
	if err != nil {
		t.Fatalf("build lookup payload: %v", err)
	}
	if !payload.Passed ||
		payload.FailureClass != "" ||
		payload.FailureCode != "" ||
		payload.GuardStatus != "passed" ||
		payload.Sources == nil ||
		payload.Sources.FailureClass != "" ||
		payload.Sources.FallbackHint != "credential_invalid" ||
		payload.AnswerContract == nil ||
		payload.AnswerContract.Reason != "guard_passed" {
		t.Fatalf("expected fallback hint to remain diagnostic after grounded provider success, got %#v", payload)
	}
}

func TestLatestNewsLookupDoesNotRetryInvalidProviderConfig(t *testing.T) {
	calls := 0
	payload, err := BuildLatestNewsLookupPayload(context.Background(), LatestNewsLookupConfig{
		AnswerContract: true,
		Handlers: LatestNewsLookupHandlers{
			Sources: func(_ context.Context, _ map[string]any, intent newsbrief.LatestNewsLookupIntent) (newsbrief.LatestNewsSourcesPayload, error) {
				calls++
				return newsbrief.LatestNewsSourcesPayload{
					AdapterStatus:  "provider_execution_failed",
					FailureCode:    "search_provider_failure_subscription_token_invalid",
					ErrorClass:     "provider_execution_failed",
					Provider:       "brave",
					ProviderStatus: "SUBSCRIPTION_TOKEN_INVALID",
					FallbackHint:   "replace BRAVE_API_KEY",
					Retryable:      true,
					UserMessage:    intent.UserMessage,
					Intent:         intent,
				}, nil
			},
		},
	}, map[string]any{"user_message": "帮我看下示例事件最新新闻"})
	if err != nil {
		t.Fatalf("build lookup payload: %v", err)
	}
	if calls != 1 {
		t.Fatalf("invalid provider config must not retry, calls=%d", calls)
	}
	if payload.FailureClass != "config_invalid" ||
		payload.Sources == nil ||
		payload.Sources.Retryable ||
		payload.Sources.RetrySuppressedReason != "terminal_failure_class:config_invalid" ||
		payload.RetrySuppressedReason != "terminal_failure_class:config_invalid" ||
		payload.RetryAttemptCount != 1 ||
		payload.AnswerContract == nil ||
		payload.AnswerContract.Reason != "search_provider_config_invalid" {
		t.Fatalf("expected config-invalid diagnostics, got %#v", payload)
	}
}

func TestLatestNewsLookupRetriesRateLimitUntilBudgetExhausted(t *testing.T) {
	calls := 0
	payload, err := BuildLatestNewsLookupPayload(context.Background(), LatestNewsLookupConfig{
		AnswerContract: true,
		Retry:          LatestNewsRetryPolicy{MaxAttempts: 2},
		Handlers: LatestNewsLookupHandlers{
			Sources: func(_ context.Context, _ map[string]any, intent newsbrief.LatestNewsLookupIntent) (newsbrief.LatestNewsSourcesPayload, error) {
				calls++
				return newsbrief.LatestNewsSourcesPayload{
					AdapterStatus:               "rate_limited",
					FailureCode:                 "search_provider_status_429",
					ErrorClass:                  "rate_limited",
					Provider:                    "brave",
					ProviderStatus:              "status_429",
					SearchAttemptCount:          2,
					PrimarySearchAttemptCount:   1,
					AlternateSearchAttemptCount: 1,
					UnavailableProviders: []newsbrief.ProviderStatus{
						{Provider: "perplexity", Reason: "missing_credentials", RequiresEnv: []string{"PERPLEXITY_API_KEY"}},
					},
					Retryable:   true,
					UserMessage: intent.UserMessage,
					Intent:      intent,
				}, nil
			},
		},
	}, map[string]any{"user_message": "帮我看下示例事件最新新闻"})
	if err != nil {
		t.Fatalf("build lookup payload: %v", err)
	}
	if calls != 2 ||
		payload.FailureClass != "rate_limited" ||
		payload.RetryAttemptCount != 2 ||
		!payload.RetryExhausted ||
		payload.Sources == nil ||
		payload.Sources.SearchAttemptCount != 4 ||
		payload.Sources.PrimarySearchAttemptCount != 2 ||
		payload.Sources.AlternateSearchAttemptCount != 2 ||
		payload.AnswerContract == nil ||
		payload.AnswerContract.Reason != "search_provider_unavailable" {
		t.Fatalf("expected rate-limit retry exhaustion, calls=%d payload=%#v", calls, payload)
	}
}

func TestLatestNewsLookupDoesNotRetryExhaustedQuota(t *testing.T) {
	calls := 0
	payload, err := BuildLatestNewsLookupPayload(context.Background(), LatestNewsLookupConfig{
		AnswerContract: true,
		Retry:          LatestNewsRetryPolicy{MaxAttempts: 2},
		Handlers: LatestNewsLookupHandlers{
			Sources: func(_ context.Context, _ map[string]any, intent newsbrief.LatestNewsLookupIntent) (newsbrief.LatestNewsSourcesPayload, error) {
				calls++
				return newsbrief.LatestNewsSourcesPayload{
					AdapterStatus:               "provider_execution_failed",
					FailureCode:                 "search_provider_failure_missing",
					ErrorClass:                  "provider_execution_failed",
					Provider:                    "baidu",
					ProviderStatus:              "quota_limited",
					RetrySuppressedReason:       "terminal_provider_status:quota_limited",
					SearchAttemptCount:          3,
					PrimarySearchAttemptCount:   1,
					AlternateSearchAttemptCount: 2,
					UserMessage:                 intent.UserMessage,
					Intent:                      intent,
				}, nil
			},
		},
	}, map[string]any{"user_message": "帮我看下示例事件最新新闻"})
	if err != nil {
		t.Fatalf("build lookup payload: %v", err)
	}
	if calls != 1 ||
		payload.FailureClass != "quota_limited" ||
		payload.RetryAttemptCount != 1 ||
		payload.RetryExhausted ||
		payload.RetrySuppressedReason != "terminal_provider_status:quota_limited" ||
		payload.Sources == nil ||
		payload.Sources.SearchAttemptCount != 3 ||
		payload.AnswerContract == nil ||
		payload.AnswerContract.Reason != "search_provider_quota_limited" ||
		!containsString(payload.AnswerContract.DoNotRetryTools, newsbrief.ToolLatestNewsLookup) ||
		!containsString(payload.AnswerContract.DoNotRetryTools, "search") {
		t.Fatalf("expected exhausted quota to stop immediate retries, calls=%d payload=%#v", calls, payload)
	}
}

func TestLatestNewsLookupRetriesEvidenceWeakWithSeedSources(t *testing.T) {
	calls := 0
	payload, err := BuildLatestNewsLookupPayload(context.Background(), LatestNewsLookupConfig{
		AnswerContract: true,
		Retry:          LatestNewsRetryPolicy{MaxAttempts: 2},
		Handlers: LatestNewsLookupHandlers{
			Sources: func(_ context.Context, params map[string]any, intent newsbrief.LatestNewsLookupIntent) (newsbrief.LatestNewsSourcesPayload, error) {
				calls++
				if calls == 1 {
					return newsbrief.LatestNewsSourcesPayload{
						AdapterStatus: "ok",
						UserMessage:   intent.UserMessage,
						Intent:        intent,
						PrimarySource: newsbrief.LatestNewsLookupSource{
							Title:       "示例事件最新进展",
							SourceURL:   "https://news.example.com/a",
							PublishedAt: "2026-05-15T09:00:00",
							KeyUpdate:   "示例事件出现最新进展。",
							Text:        "发布时间: 2026-05-15 09:00 UTC\n示例事件出现最新进展。",
						},
					}, nil
				}
				if got := newsbrief.IntArg(params["latest_news_retry_attempt"]); got != 1 {
					t.Fatalf("expected retry attempt marker, got %d params=%#v", got, params)
				}
				seeds, ok := params["latest_news_seed_sources"].([]map[string]any)
				if !ok || len(seeds) != 1 || newsbrief.StringArg(seeds[0]["source_url"]) != "https://news.example.com/a" {
					t.Fatalf("expected previous primary source seed, got %#v", params["latest_news_seed_sources"])
				}
				return newsbrief.LatestNewsSourcesPayload{
					AdapterStatus: "ok",
					UserMessage:   intent.UserMessage,
					Intent:        intent,
					PrimarySource: newsbrief.LatestNewsLookupSource{
						Title:       "示例事件最新进展",
						SourceURL:   "https://news.example.com/a",
						PublishedAt: "2026-05-15T09:00:00",
						KeyUpdate:   "示例事件出现最新进展。",
						Text:        "发布时间: 2026-05-15 09:00 UTC\n示例事件出现最新进展。",
					},
					SupportingSources: []newsbrief.LatestNewsLookupSource{{
						Title:       "独立来源确认示例事件",
						SourceURL:   "https://wire.example.net/b",
						PublishedAt: "2026-05-15T09:10:00",
						KeyUpdate:   "独立来源确认示例事件出现最新进展。",
						Text:        "发布时间: 2026-05-15 09:10 UTC\n独立来源确认示例事件出现最新进展。",
					}},
				}, nil
			},
		},
	}, map[string]any{"user_message": "帮我看下示例事件最新新闻", "topic": "示例事件"})
	if err != nil {
		t.Fatalf("build lookup payload: %v", err)
	}
	if calls != 2 {
		t.Fatalf("expected evidence retry, got calls=%d", calls)
	}
	if !payload.Passed || payload.GuardStatus != "passed" || payload.RetryAttemptCount != 1 || payload.RetryExhausted {
		t.Fatalf("expected retry to recover to guarded pass, got %#v", payload)
	}
	if len(payload.RetryAttempts) != 1 ||
		payload.RetryAttempts[0].FailureClass != "evidence_weak" ||
		payload.RetryAttempts[0].FailureCode != "single_source_only" ||
		!payload.RetryAttempts[0].Retryable {
		t.Fatalf("expected evidence weak retry attempt diagnostics, got %#v", payload.RetryAttempts)
	}
}

func TestLatestNewsLookupDoesNotRepeatAfterSourceEvidenceRetryExhausted(t *testing.T) {
	calls := 0
	payload, err := BuildLatestNewsLookupPayload(context.Background(), LatestNewsLookupConfig{
		AnswerContract: true,
		Retry:          LatestNewsRetryPolicy{MaxAttempts: 2},
		Handlers: LatestNewsLookupHandlers{
			Sources: func(_ context.Context, _ map[string]any, intent newsbrief.LatestNewsLookupIntent) (newsbrief.LatestNewsSourcesPayload, error) {
				calls++
				return newsbrief.LatestNewsSourcesPayload{
					AdapterStatus: "evidence_incomplete",
					FailureCode:   "latest_news_search_open_page_no_grounded_sources",
					FailureClass:  "evidence_missing",
					Retryable:     true,
					UserMessage:   intent.UserMessage,
					Intent:        intent,
				}, nil
			},
		},
	}, map[string]any{"user_message": "帮我看下示例事件最新新闻"})
	if err != nil {
		t.Fatalf("build lookup payload: %v", err)
	}
	if calls != 2 || payload.FailureClass != "evidence_missing" || !payload.RetryExhausted {
		t.Fatalf("expected one exhausted source retry budget without an outer duplicate retry, calls=%d payload=%#v", calls, payload)
	}
	if payload.RetryAttemptCount != 2 || payload.Sources == nil || !payload.Sources.RetryExhausted {
		t.Fatalf("expected source retry exhaustion diagnostics to propagate, got %#v", payload)
	}
	if payload.Summary != "" || containsString(payload.AnswerReadiness.ReadyDimensions, "candidate_summary") {
		t.Fatalf("expected internal failure diagnostics to stay out of candidate content, got %#v", payload)
	}
	if payload.AnswerContract == nil || strings.Contains(payload.AnswerContract.FinalAnswerDraft, "候选摘要：") || strings.Contains(payload.AnswerContract.FinalAnswerDraft, payload.FailureCode) {
		t.Fatalf("expected bounded draft without diagnostic-code candidate summary, got %#v", payload.AnswerContract)
	}
}

func TestLatestNewsFailureClassPrefersProviderFailureOverEvidenceBoundary(t *testing.T) {
	if got := latestNewsFailureClassFromValues("evidence_missing", "search_provider_failure_missing", "quota_limited"); got != "quota_limited" {
		t.Fatalf("expected provider root cause to outrank evidence boundary, got %q", got)
	}
	if got := latestNewsFailureClassFromValues("evidence_missing"); got != "evidence_missing" {
		t.Fatalf("expected standalone evidence boundary to remain classified, got %q", got)
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
