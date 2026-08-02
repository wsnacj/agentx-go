package hostkit

import (
	"context"
	"strings"
	"time"

	newsbrief "github.com/wsnacj/agentx-go/scenes/publicnews"
)

type LatestNewsSourcesHandler func(context.Context, map[string]any, newsbrief.LatestNewsLookupIntent) (newsbrief.LatestNewsSourcesPayload, error)
type LatestNewsPayloadHandler func(context.Context, map[string]any) (newsbrief.Payload, error)

type LatestNewsLookupHandlers struct {
	Sources LatestNewsSourcesHandler
	Extract LatestNewsPayloadHandler
	Guard   LatestNewsPayloadHandler
}

type LatestNewsLookupConfig struct {
	Source              string
	SourcePolicyDefault string
	AnswerContract      bool
	EvidenceReviewer    newsbrief.EvidenceReviewer
	Retry               LatestNewsRetryPolicy
	Handlers            LatestNewsLookupHandlers
}

type LatestNewsRetryPolicy struct {
	MaxAttempts int
	Backoff     time.Duration
}

func BuildLatestNewsLookupHandler(cfg LatestNewsLookupConfig) newsbrief.ToolPayloadHandler {
	return func(ctx context.Context, params map[string]any) (any, error) {
		return BuildLatestNewsLookupPayload(ctx, cfg, params)
	}
}

func BuildLatestNewsLookupPayload(ctx context.Context, cfg LatestNewsLookupConfig, params map[string]any) (newsbrief.LatestNewsLookupPayload, error) {
	working := copyParams(params)
	if strings.TrimSpace(newsbrief.StringArg(working["source_policy"])) == "" && strings.TrimSpace(cfg.SourcePolicyDefault) != "" {
		working["source_policy"] = strings.TrimSpace(cfg.SourcePolicyDefault)
	}
	intent := newsbrief.LatestNewsLookupIntentFromParams(working)
	maxAttempts := latestNewsLookupMaxAttempts(cfg.Retry)
	attempts := []newsbrief.RetryAttempt{}
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		attemptParams := copyParams(working)
		if attempt > 1 {
			attemptParams["latest_news_retry_attempt"] = attempt - 1
		}
		payload, err := buildLatestNewsLookupPayloadOnce(ctx, cfg, attemptParams, intent)
		if err != nil {
			return newsbrief.LatestNewsLookupPayload{}, err
		}
		attempts = append(attempts, payload.RetryAttempts...)
		record, recordable := latestNewsRetryAttemptFromLookupPayload(attempt, payload)
		retryable := recordable && record.Retryable
		if payload.Passed || !retryable || attempt >= maxAttempts {
			if recordable && !payload.Passed {
				attempts = append(attempts, record)
			}
			payload.RetryAttempts = append([]newsbrief.RetryAttempt(nil), attempts...)
			payload.RetryAttemptCount = len(payload.RetryAttempts)
			payload.RetryExhausted = payload.RetryExhausted || (retryable && attempt >= maxAttempts && !payload.Passed)
			if !retryable && payload.RetrySuppressedReason == "" && record.FailureClass != "" {
				payload.RetrySuppressedReason = latestNewsRetrySuppressedReason(record.FailureClass)
			}
			return finalizeLatestNewsLookupPayload(cfg, payload, payloadSources(payload), payloadExtract(payload), payloadGuard(payload)), nil
		}
		attempts = append(attempts, record)
		if !sleepLatestNewsRetry(ctx, cfg.Retry.Backoff) {
			return newsbrief.LatestNewsLookupPayload{}, ctx.Err()
		}
		working = latestNewsRetryParams(working, payload)
	}
	return newsbrief.LatestNewsLookupPayload{}, nil
}

func buildLatestNewsLookupPayloadOnce(ctx context.Context, cfg LatestNewsLookupConfig, working map[string]any, intent newsbrief.LatestNewsLookupIntent) (newsbrief.LatestNewsLookupPayload, error) {
	sources, err := runLatestNewsSources(ctx, cfg.Handlers.Sources, working, intent, cfg.Retry)
	if err != nil {
		return newsbrief.LatestNewsLookupPayload{}, err
	}
	intent.NeedsSourceVerify = !strings.EqualFold(strings.TrimSpace(sources.AdapterStatus), "ok")
	if latestNewsLookupSourcesTerminal(sources) {
		payload := newsbrief.LatestNewsLookupPayload{
			Tool:                  newsbrief.ToolLatestNewsLookup,
			Source:                firstNonEmpty(cfg.Source, "agentx_public_news_hostkit"),
			PackID:                newsbrief.PackID,
			CaseType:              newsbrief.CaseTypeLatestBrief,
			WorkflowID:            newsbrief.DefaultWorkflow,
			AdapterID:             sources.AdapterID,
			AdapterStatus:         latestNewsLookupAdapterStatus(sources, newsbrief.Payload{}),
			FailureCode:           latestNewsLookupFailureCode(sources, newsbrief.Payload{}),
			FailureClass:          latestNewsLookupFailureClass(sources, newsbrief.Payload{}),
			Intent:                intent,
			Sources:               &sources,
			Warnings:              latestNewsLookupWarnings(sources, newsbrief.Payload{}, newsbrief.Payload{}),
			RetryAttempts:         append([]newsbrief.RetryAttempt(nil), sources.RetryAttempts...),
			RetryAttemptCount:     sources.RetryAttemptCount,
			RetryExhausted:        sources.RetryExhausted,
			RetrySuppressedReason: sources.RetrySuppressedReason,
		}
		return finalizeLatestNewsLookupPayload(cfg, payload, sources, newsbrief.Payload{}, newsbrief.Payload{}), nil
	}
	sources = latestNewsLookupSanitizeSnippetGrounding(sources)
	working = paramsWithSources(working, sources, intent)

	extract, err := runLatestNewsExtract(ctx, cfg.Handlers.Extract, working, sources, intent)
	if err != nil {
		return newsbrief.LatestNewsLookupPayload{}, err
	}
	working = paramsWithExtract(working, extract)

	guard, err := runLatestNewsGuard(ctx, cfg.Handlers.Guard, working, sources, intent)
	if err != nil {
		return newsbrief.LatestNewsLookupPayload{}, err
	}
	guard = newsbrief.ApplyEvidenceReview(ctx, guard, newsbrief.ContextFromLookupSource(sources.PrimarySource, intent.UserMessage), intent, cfg.EvidenceReviewer)
	if !strings.EqualFold(strings.TrimSpace(guard.GuardStatus), "passed") {
		intent.NeedsSourceVerify = true
	}

	payload := newsbrief.LatestNewsLookupPayload{
		Tool:                  newsbrief.ToolLatestNewsLookup,
		Source:                firstNonEmpty(cfg.Source, "agentx_public_news_hostkit"),
		PackID:                newsbrief.PackID,
		CaseType:              newsbrief.CaseTypeLatestBrief,
		WorkflowID:            newsbrief.DefaultWorkflow,
		AdapterID:             firstNonEmpty(sources.AdapterID, guard.Source, extract.Source),
		AdapterStatus:         latestNewsLookupAdapterStatus(sources, guard),
		FailureCode:           latestNewsLookupFailureCode(sources, guard),
		FailureClass:          latestNewsLookupFailureClass(sources, guard),
		GuardStatus:           guard.GuardStatus,
		Intent:                intent,
		Sources:               &sources,
		Extract:               &extract,
		Guard:                 &guard,
		Warnings:              latestNewsLookupWarnings(sources, extract, guard),
		RetryAttempts:         append([]newsbrief.RetryAttempt(nil), sources.RetryAttempts...),
		RetryAttemptCount:     sources.RetryAttemptCount,
		RetryExhausted:        sources.RetryExhausted,
		RetrySuppressedReason: sources.RetrySuppressedReason,
	}
	return finalizeLatestNewsLookupPayload(cfg, payload, sources, extract, guard), nil
}

func latestNewsLookupMaxAttempts(retry LatestNewsRetryPolicy) int {
	if retry.MaxAttempts <= 0 {
		return 2
	}
	return retry.MaxAttempts
}

func latestNewsRetryAttemptFromLookupPayload(attempt int, payload newsbrief.LatestNewsLookupPayload) (newsbrief.RetryAttempt, bool) {
	if payload.Passed || !latestNewsLookupEvidenceRetryable(payload) {
		return newsbrief.RetryAttempt{}, false
	}
	failureClass := firstNonEmpty(payload.FailureClass, latestNewsLookupFailureClass(payloadSources(payload), payloadGuard(payload)))
	if failureClass == "" {
		return newsbrief.RetryAttempt{}, false
	}
	return newsbrief.RetryAttempt{
		Attempt:        attempt,
		AdapterStatus:  firstNonEmpty(payload.GuardStatus, payload.AdapterStatus),
		FailureCode:    firstNonEmpty(payload.FailureCode, strings.Join(payload.ReviewReasons, ","), strings.Join(payload.MissingNewsFields, ",")),
		FailureClass:   failureClass,
		Provider:       latestNewsPayloadProvider(payload),
		ProviderStatus: latestNewsPayloadProviderStatus(payload),
		Retryable:      true,
	}, true
}

func latestNewsLookupEvidenceRetryable(payload newsbrief.LatestNewsLookupPayload) bool {
	if payload.Passed || strings.EqualFold(strings.TrimSpace(payload.GuardStatus), "passed") {
		return false
	}
	if payload.Sources != nil && payload.Sources.RetryExhausted {
		return false
	}
	if latestNewsLookupProviderFailure(payload) || latestNewsLookupSourcesTerminal(payloadSources(payload)) {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(payload.FailureClass)) {
	case "evidence_missing", "evidence_weak":
		return true
	}
	for _, value := range latestNewsLookupEvidenceRetrySignals(payload) {
		normalized := strings.ToLower(strings.TrimSpace(value))
		switch {
		case normalized == "grounded_page_text" ||
			normalized == "no_usable_source" ||
			normalized == "single_source_only" ||
			normalized == "cross_check_evidence_missing" ||
			normalized == "supporting_source_ungrounded" ||
			normalized == "latest_news_independent_supporting_source_missing" ||
			normalized == "latest_news_open_page_attempt_budget_exhausted" ||
			normalized == "latest_news_missing_fields" ||
			strings.HasPrefix(normalized, "evidence_review") ||
			strings.Contains(normalized, "missing_news_fields") ||
			strings.Contains(normalized, "cross_check") ||
			strings.Contains(normalized, "ungrounded"):
			return true
		}
	}
	return false
}

func latestNewsLookupProviderFailure(payload newsbrief.LatestNewsLookupPayload) bool {
	status := strings.ToLower(strings.TrimSpace(payload.AdapterStatus))
	code := strings.ToLower(strings.TrimSpace(payload.FailureCode))
	failureClass := strings.ToLower(strings.TrimSpace(payload.FailureClass))
	switch failureClass {
	case "config_invalid", "auth_missing", "quota_limited", "rate_limited", "transient_network", "temporary_provider_error", "provider_unavailable":
		return true
	}
	switch status {
	case "provider_unavailable", "provider_execution_failed", "rate_limited", "timeout":
		return true
	}
	if strings.Contains(code, "search_provider_failure") ||
		strings.Contains(code, "provider_unavailable") ||
		strings.Contains(code, "missing_credentials") ||
		strings.Contains(code, "subscription_token_invalid") {
		return true
	}
	if payload.Sources == nil {
		return false
	}
	sourceStatus := strings.ToLower(strings.TrimSpace(firstNonEmpty(payload.Sources.FailureClass, payload.Sources.ErrorClass, payload.Sources.AdapterStatus, payload.Sources.ProviderStatus)))
	sourceCode := strings.ToLower(strings.TrimSpace(payload.Sources.FailureCode))
	switch sourceStatus {
	case "provider_unavailable", "provider_execution_failed", "missing_credentials", "config_invalid", "auth_missing", "quota_limited", "rate_limited", "transient_network":
		return true
	}
	return strings.Contains(sourceCode, "search_provider_failure") ||
		strings.Contains(sourceCode, "missing_credentials") ||
		strings.Contains(sourceCode, "subscription_token_invalid")
}

func latestNewsLookupEvidenceRetrySignals(payload newsbrief.LatestNewsLookupPayload) []string {
	values := []string{payload.FailureCode, payload.GuardStatus}
	values = append(values, payload.MissingNewsFields...)
	values = append(values, payload.ReviewReasons...)
	values = append(values, payload.Warnings...)
	if payload.Guard != nil {
		values = append(values, payload.Guard.GuardStatus)
		values = append(values, payload.Guard.MissingNewsFields...)
		values = append(values, payload.Guard.ReviewReasons...)
		values = append(values, payload.Guard.Warnings...)
	}
	if payload.Sources != nil {
		values = append(values, payload.Sources.FailureCode)
		values = append(values, payload.Sources.Warnings...)
	}
	return values
}

func latestNewsRetryParams(params map[string]any, payload newsbrief.LatestNewsLookupPayload) map[string]any {
	out := copyParams(params)
	for _, key := range []string{"page_id", "title", "headline", "source_url", "source_site", "published_at", "key_update", "text", "supporting_sources", "sources", "source_count"} {
		delete(out, key)
	}
	seeds := latestNewsRetrySeedSources(payload)
	if len(seeds) > 0 {
		out["latest_news_seed_sources"] = lookupSourceParamMaps(seeds, false)
	}
	out["latest_news_retry_previous_failure_code"] = payload.FailureCode
	out["latest_news_retry_previous_guard_status"] = payload.GuardStatus
	return out
}

func latestNewsRetrySeedSources(payload newsbrief.LatestNewsLookupPayload) []newsbrief.LatestNewsLookupSource {
	if payload.Sources == nil {
		return nil
	}
	out := []newsbrief.LatestNewsLookupSource{}
	if !latestNewsLookupSourceEmpty(payload.Sources.PrimarySource) {
		out = append(out, payload.Sources.PrimarySource)
	}
	for _, source := range payload.Sources.SupportingSources {
		if latestNewsLookupSourceEmpty(source) {
			continue
		}
		out = append(out, source)
	}
	return out
}

func latestNewsLookupSourceEmpty(source newsbrief.LatestNewsLookupSource) bool {
	return strings.TrimSpace(source.SourceURL) == "" &&
		strings.TrimSpace(source.Text) == "" &&
		strings.TrimSpace(source.Title) == "" &&
		strings.TrimSpace(source.Headline) == ""
}

func payloadSources(payload newsbrief.LatestNewsLookupPayload) newsbrief.LatestNewsSourcesPayload {
	if payload.Sources == nil {
		return newsbrief.LatestNewsSourcesPayload{}
	}
	return *payload.Sources
}

func payloadExtract(payload newsbrief.LatestNewsLookupPayload) newsbrief.Payload {
	if payload.Extract == nil {
		return newsbrief.Payload{}
	}
	return *payload.Extract
}

func payloadGuard(payload newsbrief.LatestNewsLookupPayload) newsbrief.Payload {
	if payload.Guard == nil {
		return newsbrief.Payload{}
	}
	return *payload.Guard
}

func latestNewsPayloadProvider(payload newsbrief.LatestNewsLookupPayload) string {
	if payload.Sources == nil {
		return ""
	}
	return firstNonEmpty(payload.Sources.EffectiveProvider, payload.Sources.Provider, payload.Sources.RequestedProvider)
}

func latestNewsPayloadProviderStatus(payload newsbrief.LatestNewsLookupPayload) string {
	if payload.Sources == nil {
		return ""
	}
	return firstNonEmpty(payload.Sources.ProviderStatus, payload.Sources.FailureClass, payload.Sources.ErrorClass, payload.Sources.AdapterStatus)
}

func finalizeLatestNewsLookupPayload(cfg LatestNewsLookupConfig, payload newsbrief.LatestNewsLookupPayload, sources newsbrief.LatestNewsSourcesPayload, extract newsbrief.Payload, guard newsbrief.Payload) newsbrief.LatestNewsLookupPayload {
	payload = projectLatestNewsLookupPayload(payload, sources, extract, guard)
	payload = newsbrief.AttachLatestNewsEvaluatorReport(payload)
	if cfg.AnswerContract {
		payload.AnswerContract = newsbrief.LatestNewsLookupAnswerContract(payload)
	}
	payload.AnswerReadiness = newsbrief.LatestNewsLookupAnswerReadiness(payload)
	return payload
}

func latestNewsLookupSanitizeSnippetGrounding(sources newsbrief.LatestNewsSourcesPayload) newsbrief.LatestNewsSourcesPayload {
	if latestNewsLookupWarningContains(sources.Warnings, "latest_news_search_snippet_primary_used") {
		sources.PrimarySource.Text = ""
		sources.PrimarySource.PageID = ""
	}
	if latestNewsLookupWarningContains(sources.Warnings, "latest_news_search_snippet_supporting_source_used") {
		for idx := range sources.SupportingSources {
			sources.SupportingSources[idx].Text = ""
			sources.SupportingSources[idx].PageID = ""
		}
	}
	return sources
}

func runLatestNewsSources(ctx context.Context, handler LatestNewsSourcesHandler, params map[string]any, intent newsbrief.LatestNewsLookupIntent, retry LatestNewsRetryPolicy) (newsbrief.LatestNewsSourcesPayload, error) {
	if handler != nil {
		return runLatestNewsSourcesWithRetry(ctx, handler, params, intent, retry)
	}
	primary := newsbrief.LookupSourceFromParams(params)
	status := "ok"
	failureCode := ""
	if strings.TrimSpace(primary.PageID) == "" && strings.TrimSpace(primary.Text) == "" && strings.TrimSpace(primary.SourceURL) == "" {
		status = "unsupported"
		failureCode = "latest_news_source_adapter_not_configured"
	}
	payload := newsbrief.LatestNewsSourcesPayload{
		Tool:              newsbrief.ToolLatestNewsLookup,
		Source:            "agentx_public_news_hostkit_params_source",
		PackID:            newsbrief.PackID,
		CaseType:          newsbrief.CaseTypeLatestBrief,
		WorkflowID:        newsbrief.DefaultWorkflow,
		AdapterStatus:     status,
		FailureCode:       failureCode,
		UserMessage:       intent.UserMessage,
		Intent:            intent,
		PrimarySource:     primary,
		SupportingSources: supportingSourcesFromParams(params),
	}
	return normalizeLatestNewsSourcesFailure(payload), nil
}

func runLatestNewsSourcesWithRetry(ctx context.Context, handler LatestNewsSourcesHandler, params map[string]any, intent newsbrief.LatestNewsLookupIntent, retry LatestNewsRetryPolicy) (newsbrief.LatestNewsSourcesPayload, error) {
	maxAttempts := retry.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 2
	}
	attempts := []newsbrief.RetryAttempt{}
	searchAttemptCount := 0
	primarySearchAttemptCount := 0
	alternateSearchAttemptCount := 0
	applySearchAttemptTotals := func(payload newsbrief.LatestNewsSourcesPayload) newsbrief.LatestNewsSourcesPayload {
		payload.SearchAttemptCount = searchAttemptCount
		payload.PrimarySearchAttemptCount = primarySearchAttemptCount
		payload.AlternateSearchAttemptCount = alternateSearchAttemptCount
		return payload
	}
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		payload, err := handler(ctx, params, intent)
		if err != nil {
			record := latestNewsRetryAttemptFromError(attempt, err)
			attempts = append(attempts, record)
			if attempt >= maxAttempts || !record.Retryable {
				return newsbrief.LatestNewsSourcesPayload{}, err
			}
			if !sleepLatestNewsRetry(ctx, retry.Backoff) {
				return newsbrief.LatestNewsSourcesPayload{}, ctx.Err()
			}
			continue
		}
		searchAttemptCount += payload.SearchAttemptCount
		primarySearchAttemptCount += payload.PrimarySearchAttemptCount
		alternateSearchAttemptCount += payload.AlternateSearchAttemptCount
		payload = applySearchAttemptTotals(normalizeLatestNewsSourcesFailure(payload))
		payload.RetryAttempts = append([]newsbrief.RetryAttempt(nil), attempts...)
		payload.RetryAttemptCount = len(attempts)
		if latestNewsSourcesPayloadSucceeded(payload) {
			return payload, nil
		}
		record := latestNewsRetryAttemptFromPayload(attempt, payload)
		if record.FailureClass != "" {
			attempts = append(attempts, record)
		}
		if attempt >= maxAttempts || !record.Retryable {
			payload.RetryAttempts = append([]newsbrief.RetryAttempt(nil), attempts...)
			payload.RetryAttemptCount = len(attempts)
			payload.RetryExhausted = record.Retryable && attempt >= maxAttempts
			if !record.Retryable && payload.RetrySuppressedReason == "" {
				payload.RetrySuppressedReason = latestNewsRetrySuppressedReason(record.FailureClass)
			}
			return payload, nil
		}
		if !sleepLatestNewsRetry(ctx, retry.Backoff) {
			return newsbrief.LatestNewsSourcesPayload{}, ctx.Err()
		}
	}
	return newsbrief.LatestNewsSourcesPayload{}, nil
}

func sleepLatestNewsRetry(ctx context.Context, delay time.Duration) bool {
	if delay <= 0 {
		return true
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func normalizeLatestNewsSourcesFailure(payload newsbrief.LatestNewsSourcesPayload) newsbrief.LatestNewsSourcesPayload {
	payload.FailureClass = latestNewsSourcesFailureClass(payload)
	if payload.FailureClass != "" {
		payload.Retryable = latestNewsFailureClassRetryable(payload.FailureClass)
	}
	if payload.RetryAttemptCount == 0 && len(payload.RetryAttempts) > 0 {
		payload.RetryAttemptCount = len(payload.RetryAttempts)
	}
	if payload.FailureClass != "" && !payload.Retryable && payload.RetrySuppressedReason == "" {
		payload.RetrySuppressedReason = latestNewsRetrySuppressedReason(payload.FailureClass)
	}
	return payload
}

func latestNewsSourcesPayloadSucceeded(payload newsbrief.LatestNewsSourcesPayload) bool {
	if latestNewsSourcesFailureClass(payload) != "" {
		return false
	}
	for _, raw := range []string{payload.ErrorClass, payload.ProviderStatus, payload.AdapterStatus} {
		status := strings.ToLower(strings.TrimSpace(raw))
		if status == "" || status == "ok" || status == "available" {
			continue
		}
		return false
	}
	return true
}

func latestNewsRetryAttemptFromPayload(attempt int, payload newsbrief.LatestNewsSourcesPayload) newsbrief.RetryAttempt {
	failureClass := firstNonEmpty(payload.FailureClass, latestNewsSourcesFailureClass(payload))
	return newsbrief.RetryAttempt{
		Attempt:        attempt,
		AdapterStatus:  firstNonEmpty(payload.ErrorClass, payload.AdapterStatus),
		FailureCode:    payload.FailureCode,
		FailureClass:   failureClass,
		Provider:       firstNonEmpty(payload.EffectiveProvider, payload.Provider, payload.RequestedProvider),
		ProviderStatus: payload.ProviderStatus,
		Retryable:      latestNewsFailureClassRetryable(failureClass),
	}
}

func latestNewsRetryAttemptFromError(attempt int, err error) newsbrief.RetryAttempt {
	failureClass := latestNewsFailureClassFromValues(err.Error())
	return newsbrief.RetryAttempt{
		Attempt:      attempt,
		FailureCode:  "latest_news_source_handler_error",
		FailureClass: failureClass,
		Retryable:    latestNewsFailureClassRetryable(failureClass),
	}
}

func latestNewsSourcesFailureClass(payload newsbrief.LatestNewsSourcesPayload) string {
	return latestNewsFailureClassFromValues(
		payload.FailureClass,
		payload.FailureCode,
		payload.ErrorClass,
		payload.AdapterStatus,
		payload.ProviderStatus,
	)
}

func latestNewsFailureClassFromValues(values ...string) string {
	joined := strings.ToLower(strings.Join(cleanLatestNewsClassTokens(values), " "))
	if joined == "" {
		return ""
	}
	switch {
	case strings.Contains(joined, "latest_news_source_adapter_not_configured") ||
		strings.Contains(joined, "source_adapter_not_configured") ||
		strings.Contains(joined, "unsupported"):
		return "unsupported"
	case strings.Contains(joined, "subscription_token_invalid") ||
		strings.Contains(joined, "credential_invalid") ||
		strings.Contains(joined, "credential invalid") ||
		strings.Contains(joined, "invalid api key") ||
		strings.Contains(joined, "invalid_api_key") ||
		strings.Contains(joined, "invalid token") ||
		strings.Contains(joined, "invalid_token") ||
		strings.Contains(joined, "unauthorized") ||
		strings.Contains(joined, "forbidden") ||
		strings.Contains(joined, "http_401") ||
		strings.Contains(joined, "status_401") ||
		strings.Contains(joined, "http_403") ||
		strings.Contains(joined, "status_403"):
		return "config_invalid"
	case strings.Contains(joined, "missing_credentials") ||
		strings.Contains(joined, "missing credential") ||
		strings.Contains(joined, "api key missing") ||
		strings.Contains(joined, "not_configured") ||
		strings.Contains(joined, "not configured") ||
		strings.Contains(joined, "unconfigured"):
		return "auth_missing"
	case strings.Contains(joined, "quota_limited") ||
		strings.Contains(joined, "quota limited") ||
		strings.Contains(joined, "quota_exhausted") ||
		strings.Contains(joined, "quota exhausted"):
		return "quota_limited"
	case strings.Contains(joined, "rate_limited") ||
		strings.Contains(joined, "rate limit") ||
		strings.Contains(joined, "retry-after") ||
		strings.Contains(joined, "http_429") ||
		strings.Contains(joined, "status_429"):
		return "rate_limited"
	case strings.Contains(joined, "timeout") ||
		strings.Contains(joined, "deadline exceeded") ||
		strings.Contains(joined, "connection reset") ||
		strings.Contains(joined, "temporary") ||
		strings.Contains(joined, "eof") ||
		strings.Contains(joined, "http_502") ||
		strings.Contains(joined, "status_502") ||
		strings.Contains(joined, "http_503") ||
		strings.Contains(joined, "status_503") ||
		strings.Contains(joined, "http_504") ||
		strings.Contains(joined, "status_504"):
		return "transient_network"
	case strings.Contains(joined, "provider_execution_failed"):
		return "temporary_provider_error"
	case strings.Contains(joined, "provider_unavailable") ||
		strings.Contains(joined, "search_provider_failure"):
		return "provider_unavailable"
	case strings.Contains(joined, "evidence_missing") ||
		strings.Contains(joined, "evidence missing"):
		return "evidence_missing"
	case strings.Contains(joined, "evidence_weak") ||
		strings.Contains(joined, "evidence weak"):
		return "evidence_weak"
	default:
		return ""
	}
}

func cleanLatestNewsClassTokens(values []string) []string {
	out := []string{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			out = append(out, value)
		}
	}
	return out
}

func latestNewsFailureClassRetryable(failureClass string) bool {
	switch strings.ToLower(strings.TrimSpace(failureClass)) {
	case "transient_network", "rate_limited", "temporary_provider_error", "evidence_missing", "evidence_weak":
		return true
	default:
		return false
	}
}

func latestNewsRetrySuppressedReason(failureClass string) string {
	failureClass = strings.ToLower(strings.TrimSpace(failureClass))
	if failureClass == "" || latestNewsFailureClassRetryable(failureClass) {
		return ""
	}
	return "terminal_failure_class:" + failureClass
}

func runLatestNewsExtract(ctx context.Context, handler LatestNewsPayloadHandler, params map[string]any, sources newsbrief.LatestNewsSourcesPayload, intent newsbrief.LatestNewsLookupIntent) (newsbrief.Payload, error) {
	if handler != nil {
		return handler(ctx, params)
	}
	return newsbrief.BuildExtractPayload(newsbrief.ContextFromLookupSource(sources.PrimarySource, intent.UserMessage), newsbrief.PayloadOptions{
		ExtractSource: "agentx_public_news_hostkit_extract",
	}), nil
}

func runLatestNewsGuard(ctx context.Context, handler LatestNewsPayloadHandler, params map[string]any, sources newsbrief.LatestNewsSourcesPayload, intent newsbrief.LatestNewsLookupIntent) (newsbrief.Payload, error) {
	if handler != nil {
		return handler(ctx, params)
	}
	return newsbrief.BuildGuardPayload(newsbrief.ContextFromLookupSource(sources.PrimarySource, intent.UserMessage), params, nil, newsbrief.PayloadOptions{
		GuardSource: "agentx_public_news_hostkit_guard",
	}), nil
}

func paramsWithSources(params map[string]any, sources newsbrief.LatestNewsSourcesPayload, intent newsbrief.LatestNewsLookupIntent) map[string]any {
	out := copyParams(params)
	primary := sources.PrimarySource
	primaryPageID := primary.PageID
	primaryText := primary.Text
	if latestNewsLookupWarningContains(sources.Warnings, "latest_news_search_snippet_primary_used") {
		delete(out, "page_id")
		delete(out, "text")
		primaryPageID = ""
		primaryText = ""
	}
	setString(out, "user_message", firstNonEmpty(newsbrief.StringArg(out["user_message"]), intent.UserMessage))
	setString(out, "page_id", firstNonEmpty(newsbrief.StringArg(out["page_id"]), primaryPageID))
	setString(out, "title", firstNonEmpty(newsbrief.StringArg(out["title"]), primary.Title, primary.Headline))
	setString(out, "headline", firstNonEmpty(newsbrief.StringArg(out["headline"]), primary.Headline, primary.Title))
	setString(out, "source_url", firstNonEmpty(newsbrief.StringArg(out["source_url"]), primary.SourceURL))
	setString(out, "source_site", firstNonEmpty(newsbrief.StringArg(out["source_site"]), primary.SourceSite))
	setString(out, "published_at", firstNonEmpty(newsbrief.StringArg(out["published_at"]), primary.PublishedAt))
	setString(out, "key_update", firstNonEmpty(newsbrief.StringArg(out["key_update"]), primary.KeyUpdate))
	setString(out, "text", firstNonEmpty(newsbrief.StringArg(out["text"]), primaryText))
	if len(sources.SupportingSources) > 0 {
		out["supporting_sources"] = lookupSourceParamMaps(
			sources.SupportingSources,
			latestNewsLookupWarningContains(sources.Warnings, "latest_news_search_snippet_supporting_source_used"),
		)
		out["source_count"] = len(sources.SupportingSources) + 1
	}
	return out
}

func paramsWithExtract(params map[string]any, extract newsbrief.Payload) map[string]any {
	out := copyParams(params)
	setString(out, "headline", firstNonEmpty(newsbrief.StringArg(out["headline"]), extract.Evidence.Headline))
	setString(out, "source_url", firstNonEmpty(newsbrief.StringArg(out["source_url"]), extract.Evidence.SourceURL, extract.FinalURL))
	setString(out, "source_site", firstNonEmpty(newsbrief.StringArg(out["source_site"]), extract.Evidence.SourceSite))
	setString(out, "published_at", firstNonEmpty(newsbrief.StringArg(out["published_at"]), extract.Evidence.PublishedAt))
	setString(out, "key_update", preferredExtractedKeyUpdate(
		newsbrief.StringArg(out["headline"]),
		newsbrief.StringArg(out["key_update"]),
		extract.Evidence.KeyUpdate,
	))
	return out
}

func preferredExtractedKeyUpdate(headline string, current string, extracted string) string {
	current = strings.TrimSpace(current)
	extracted = strings.TrimSpace(extracted)
	if !newsbrief.KeyUpdateSufficientForHeadline(headline, current) &&
		newsbrief.KeyUpdateSufficientForHeadline(headline, extracted) {
		return extracted
	}
	return firstNonEmpty(current, extracted)
}

func lookupSourceParamMaps(sources []newsbrief.LatestNewsLookupSource, stripText bool) []map[string]any {
	out := make([]map[string]any, 0, len(sources))
	for _, source := range sources {
		text := source.Text
		if stripText {
			text = ""
		}
		out = append(out, map[string]any{
			"page_id":      source.PageID,
			"title":        firstNonEmpty(source.Title, source.Headline),
			"headline":     source.Headline,
			"source_url":   source.SourceURL,
			"source_site":  source.SourceSite,
			"published_at": source.PublishedAt,
			"key_update":   source.KeyUpdate,
			"text":         text,
		})
	}
	return out
}

func supportingSourcesFromParams(params map[string]any) []newsbrief.LatestNewsLookupSource {
	raw, ok := params["supporting_sources"]
	if !ok || raw == nil {
		raw = params["sources"]
	}
	if typed, ok := raw.([]newsbrief.LatestNewsLookupSource); ok {
		return typed
	}
	values, ok := anySlice(raw)
	if !ok {
		return nil
	}
	out := []newsbrief.LatestNewsLookupSource{}
	for _, value := range values {
		item, ok := value.(map[string]any)
		if !ok {
			continue
		}
		out = append(out, newsbrief.LookupSourceFromParams(item))
	}
	return out
}

func latestNewsLookupAdapterStatus(sources newsbrief.LatestNewsSourcesPayload, guard newsbrief.Payload) string {
	if status := strings.TrimSpace(sources.ErrorClass); status != "" && !strings.EqualFold(status, "ok") {
		return status
	}
	if status := strings.TrimSpace(sources.AdapterStatus); status != "" && !strings.EqualFold(status, "ok") {
		return status
	}
	if strings.EqualFold(strings.TrimSpace(guard.GuardStatus), "passed") {
		return "ok"
	}
	if strings.TrimSpace(guard.GuardStatus) != "" {
		return "needs_review"
	}
	return firstNonEmpty(sources.AdapterStatus, "needs_review")
}

func latestNewsLookupFailureCode(sources newsbrief.LatestNewsSourcesPayload, guard newsbrief.Payload) string {
	if code := strings.TrimSpace(sources.FailureCode); code != "" {
		return code
	}
	if strings.EqualFold(strings.TrimSpace(guard.GuardStatus), "passed") {
		return ""
	}
	if len(guard.MissingNewsFields) > 0 {
		return "latest_news_missing_fields"
	}
	if len(guard.ReviewReasons) > 0 {
		return strings.Join(guard.ReviewReasons, ",")
	}
	return ""
}

func latestNewsLookupFailureClass(sources newsbrief.LatestNewsSourcesPayload, guard newsbrief.Payload) string {
	if failureClass := strings.TrimSpace(sources.FailureClass); failureClass != "" {
		return failureClass
	}
	if strings.EqualFold(strings.TrimSpace(guard.GuardStatus), "passed") {
		return ""
	}
	if len(guard.MissingNewsFields) > 0 {
		return "evidence_missing"
	}
	if len(guard.ReviewReasons) > 0 {
		return "evidence_weak"
	}
	return latestNewsSourcesFailureClass(sources)
}

func latestNewsLookupWarnings(sources newsbrief.LatestNewsSourcesPayload, extract newsbrief.Payload, guard newsbrief.Payload) []string {
	out := append([]string{}, sources.Warnings...)
	out = appendUniqueStrings(out, extract.Warnings...)
	out = appendUniqueStrings(out, guard.Warnings...)
	return latestNewsLookupDropResolvedWarnings(out, guard)
}

func latestNewsLookupDropResolvedWarnings(warnings []string, guard newsbrief.Payload) []string {
	if len(warnings) == 0 {
		return nil
	}
	out := make([]string, 0, len(warnings))
	for _, warning := range warnings {
		switch strings.TrimSpace(warning) {
		case "published_at_missing":
			if !latestNewsLookupUnknown(guard.Evidence.PublishedAt) {
				continue
			}
		case "key_update_missing", "key_update_low_quality":
			if newsbrief.KeyUpdateSufficientForHeadline(guard.Evidence.Headline, guard.Evidence.KeyUpdate) {
				continue
			}
		}
		out = append(out, warning)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func latestNewsLookupUnknown(value string) bool {
	value = strings.TrimSpace(value)
	return value == "" || strings.EqualFold(value, "unknown") || strings.EqualFold(value, "not_confirmed") || value == "-"
}

func latestNewsLookupWarningContains(warnings []string, marker string) bool {
	marker = strings.ToLower(strings.TrimSpace(marker))
	if marker == "" {
		return false
	}
	for _, warning := range warnings {
		normalized := strings.ToLower(strings.TrimSpace(warning))
		if normalized == marker {
			return true
		}
	}
	return false
}

func projectLatestNewsLookupPayload(payload newsbrief.LatestNewsLookupPayload, sources newsbrief.LatestNewsSourcesPayload, extract newsbrief.Payload, guard newsbrief.Payload) newsbrief.LatestNewsLookupPayload {
	payload.GuardStatus = firstNonEmpty(payload.GuardStatus, guard.GuardStatus, latestNewsLookupTerminalGuardStatus(sources))
	payload.FailureClass = firstNonEmpty(payload.FailureClass, latestNewsLookupFailureClass(sources, guard))
	if payload.RetryAttemptCount == 0 && len(payload.RetryAttempts) > 0 {
		payload.RetryAttemptCount = len(payload.RetryAttempts)
	}
	if payload.RetrySuppressedReason == "" {
		payload.RetrySuppressedReason = sources.RetrySuppressedReason
	}
	if payload.RetrySuppressedReason == "" && payload.FailureClass != "" && !latestNewsFailureClassRetryable(payload.FailureClass) {
		payload.RetrySuppressedReason = latestNewsRetrySuppressedReason(payload.FailureClass)
	}
	payload.NewsFieldsReady = guard.NewsFieldsReady
	payload.CrossCheckReady = guard.CrossCheckReady
	if guard.Evaluation != nil {
		payload.Passed = guard.Evaluation.Passed
		payload.FreshnessConfirmed = guard.Evaluation.FreshnessConfirmed
		payload.SourceAccepted = guard.Evaluation.SourceAccepted
		payload.StopAfterGuardPassed = guard.Evaluation.StopAfterGuardPassed
	}
	payload.MissingNewsFields = append([]string(nil), guard.MissingNewsFields...)
	if len(payload.MissingNewsFields) == 0 && len(guard.ReviewReasons) == 0 && !strings.EqualFold(payload.GuardStatus, "passed") {
		payload.MissingNewsFields = latestNewsLookupMissingFields(sources, guard)
	}
	if payload.MissingNewsFields == nil {
		payload.MissingNewsFields = []string{}
	}
	payload.ReviewReasons = append([]string(nil), guard.ReviewReasons...)
	payload.EvidenceReview = guard.EvidenceReview
	if reason := latestNewsLookupReviewReason(sources); reason != "" {
		payload.ReviewReasons = appendUniqueStrings(payload.ReviewReasons, reason)
	}
	if payload.ReviewReasons == nil {
		payload.ReviewReasons = []string{}
	}
	payload.SourceURL = firstNonEmpty(guard.Evidence.SourceURL, guard.FinalURL, extract.Evidence.SourceURL, extract.FinalURL, sources.PrimarySource.SourceURL)
	payload.PublishedAt = firstNonEmpty(guard.Evidence.PublishedAt, extract.Evidence.PublishedAt, sources.PrimarySource.PublishedAt)
	payload.Summary = firstNonEmpty(guard.Evidence.KeyUpdate, extract.Evidence.KeyUpdate, sources.PrimarySource.KeyUpdate)
	return payload
}

func latestNewsLookupTerminalGuardStatus(sources newsbrief.LatestNewsSourcesPayload) string {
	status := strings.TrimSpace(firstNonEmpty(sources.ErrorClass, sources.AdapterStatus))
	if status == "" || strings.EqualFold(status, "ok") {
		return "missing_news_fields"
	}
	return status
}

func latestNewsLookupMissingFields(sources newsbrief.LatestNewsSourcesPayload, guard newsbrief.Payload) []string {
	out := []string{}
	if strings.TrimSpace(firstNonEmpty(guard.Evidence.SourceURL, sources.PrimarySource.SourceURL)) == "" {
		out = append(out, "source_url")
	}
	if strings.TrimSpace(firstNonEmpty(guard.Evidence.PublishedAt, sources.PrimarySource.PublishedAt)) == "" {
		out = append(out, "published_at")
	}
	if strings.TrimSpace(firstNonEmpty(guard.Evidence.KeyUpdate, sources.PrimarySource.KeyUpdate)) == "" {
		out = append(out, "key_update")
	}
	if len(out) == 0 {
		out = append(out, "guard_passed")
	}
	return out
}

func latestNewsLookupReviewReason(sources newsbrief.LatestNewsSourcesPayload) string {
	if code := strings.TrimSpace(sources.FailureCode); code != "" {
		return code
	}
	if failureClass := strings.TrimSpace(sources.FailureClass); failureClass != "" {
		return failureClass
	}
	if status := strings.TrimSpace(firstNonEmpty(sources.ErrorClass, sources.AdapterStatus)); status != "" && !strings.EqualFold(status, "ok") {
		return status
	}
	return ""
}

func latestNewsLookupSourcesTerminal(sources newsbrief.LatestNewsSourcesPayload) bool {
	status := strings.ToLower(strings.TrimSpace(firstNonEmpty(sources.FailureClass, sources.ErrorClass, sources.AdapterStatus)))
	code := strings.ToLower(strings.TrimSpace(sources.FailureCode))
	switch status {
	case "provider_unavailable", "provider_execution_failed", "quota_limited", "rate_limited", "timeout", "config_invalid", "auth_missing", "transient_network", "temporary_provider_error":
		return true
	}
	return strings.Contains(code, "search_provider_failure") ||
		strings.Contains(code, "provider_unavailable") ||
		strings.Contains(code, "missing_credentials") ||
		strings.Contains(code, "subscription_token_invalid")
}

func copyParams(params map[string]any) map[string]any {
	out := map[string]any{}
	for key, value := range params {
		out[key] = value
	}
	return out
}

func setString(values map[string]any, key string, value string) {
	if strings.TrimSpace(value) != "" {
		values[key] = strings.TrimSpace(value)
	}
}

func anySlice(raw any) ([]any, bool) {
	switch value := raw.(type) {
	case []any:
		return value, true
	case []map[string]any:
		out := make([]any, 0, len(value))
		for _, item := range value {
			out = append(out, item)
		}
		return out, true
	default:
		return nil, false
	}
}

func appendUniqueStrings(base []string, values ...string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(base)+len(values))
	for _, value := range base {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}
