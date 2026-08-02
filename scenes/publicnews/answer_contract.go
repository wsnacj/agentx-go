package publicnews

import (
	"fmt"
	"strings"
)

const (
	LatestNewsAnswerScopeGuardedBrief       = "guarded_latest_news_brief"
	LatestNewsAnswerScopeProviderDiagnostic = "provider_diagnostic_only"
	LatestNewsAnswerScopeSourceDiagnostic   = "source_quality_diagnostic_only"
)

// LatestNewsLookupAnswerContract projects a high-level latest-news lookup
// payload into a user-facing final-answer handoff when the evidence boundary is
// already clear. Hosts can opt in to this projection to avoid returning raw
// JSON from one-tool pack workflows.
func LatestNewsLookupAnswerContract(payload LatestNewsLookupPayload) *LatestNewsAnswerContract {
	if payload.Passed &&
		strings.EqualFold(strings.TrimSpace(payload.GuardStatus), "passed") &&
		payload.NewsFieldsReady &&
		payload.CrossCheckReady &&
		!latestNewsLookupUsesUngroundedSearchSnippetEvidence(payload) {
		return latestNewsGuardPassedAnswerContract(payload)
	}
	if latestNewsLookupProviderUnavailable(payload) {
		return latestNewsProviderUnavailableAnswerContract(payload)
	}
	if latestNewsLookupSourceQualityTerminal(payload) {
		return latestNewsSourceQualityAnswerContract(payload)
	}
	return nil
}

func latestNewsLookupUsesUngroundedSearchSnippetEvidence(payload LatestNewsLookupPayload) bool {
	primarySnippet := false
	supportingSnippet := false
	for _, warning := range payload.Warnings {
		primarySnippet = primarySnippet || latestNewsLookupPrimarySnippetWarning(warning)
		supportingSnippet = supportingSnippet || latestNewsLookupSupportingSnippetWarning(warning)
	}
	if payload.Sources != nil {
		for _, warning := range payload.Sources.Warnings {
			primarySnippet = primarySnippet || latestNewsLookupPrimarySnippetWarning(warning)
			supportingSnippet = supportingSnippet || latestNewsLookupSupportingSnippetWarning(warning)
		}
	}
	if primarySnippet && !latestNewsLookupGuardPrimaryGrounded(payload) {
		return true
	}
	if supportingSnippet && !latestNewsLookupGuardSupportingGrounded(payload) {
		return true
	}
	return false
}

func latestNewsLookupPrimarySnippetWarning(warning string) bool {
	normalized := strings.ToLower(strings.TrimSpace(warning))
	return normalized == "latest_news_search_snippet_primary_used" ||
		strings.Contains(normalized, "snippet_primary")
}

func latestNewsLookupSupportingSnippetWarning(warning string) bool {
	normalized := strings.ToLower(strings.TrimSpace(warning))
	return normalized == "latest_news_search_snippet_supporting_source_used" ||
		strings.Contains(normalized, "snippet_supporting")
}

func latestNewsLookupGuardPrimaryGrounded(payload LatestNewsLookupPayload) bool {
	return payload.Guard != nil && payload.Guard.Evidence.GroundedTextAvailable
}

func latestNewsLookupGuardSupportingGrounded(payload LatestNewsLookupPayload) bool {
	if payload.Guard == nil {
		return false
	}
	for _, source := range payload.Guard.SupportingSources {
		if EvidenceUsableForCrossCheck(source) {
			return true
		}
	}
	return false
}

func latestNewsGuardPassedAnswerContract(payload LatestNewsLookupPayload) *LatestNewsAnswerContract {
	summary := firstNonEmpty(payload.Summary, payloadTitle(payload), "已核验到一条最新新闻更新")
	publishedAt := firstNonEmpty(payload.PublishedAt, payloadEvidence(payload).PublishedAt, "unknown")
	sourceURL := firstNonEmpty(payload.SourceURL, payloadEvidence(payload).SourceURL, "unknown")
	lines := []string{}
	if target := latestNewsAnswerTarget(payload.Intent); target != "" {
		lines = append(lines, "查询目标："+target)
	}
	lines = append(lines,
		"一句话摘要："+ensureTerminalPeriod(summary),
		"发布时间："+publishedAt,
		"来源："+sourceURL,
	)
	if supporting := latestNewsSupportingSourceURLs(payload, 3); len(supporting) > 0 {
		lines = append(lines, "交叉核对："+strings.Join(supporting, "；"))
	}
	possibleImpact := ""
	if latestNewsIntentRequestsImpact(payload.Intent) {
		possibleImpact = "当前证据足以确认上述进展，但不足以独立量化其对产品、市场或政策的实际影响；具体影响需要结合后续正式信息继续判断。"
		lines = append(lines, "可能影响："+possibleImpact)
	}
	riskBoundary := "以上仅基于已打开并通过 guard 的公开来源；不要把可能影响当作已发生事实，事件继续变化时应重新检索。"
	lines = append(lines, "风险边界："+riskBoundary)
	return &LatestNewsAnswerContract{
		FinalAnswerRecommended: true,
		Reason:                 "guard_passed",
		AllowedSummaryScope:    LatestNewsAnswerScopeGuardedBrief,
		DoNotRetryTools:        []string{ToolLatestNewsLookup, ToolLatestNewsExtract, ToolLatestNewsGuard, "search", "open_page", "web_fetch", "browser"},
		PossibleImpact:         possibleImpact,
		RiskBoundary:           riskBoundary,
		FinalAnswerDraft:       strings.Join(lines, "\n"),
	}
}

func latestNewsIntentRequestsImpact(intent LatestNewsLookupIntent) bool {
	for _, value := range append(append([]string{}, intent.RequestedFields...), intent.RequestedOutputs...) {
		normalized := strings.ToLower(strings.TrimSpace(value))
		if strings.Contains(normalized, "impact") || strings.Contains(normalized, "影响") {
			return true
		}
	}
	message := strings.ToLower(strings.TrimSpace(firstNonEmpty(intent.UserMessage, intent.OriginalIntent)))
	return strings.Contains(message, "impact") || strings.Contains(message, "影响")
}

func latestNewsAnswerTarget(intent LatestNewsLookupIntent) string {
	if len(intent.EntityMentions) > 0 {
		return strings.Join(intent.EntityMentions, "、")
	}
	return strings.TrimSpace(intent.Topic)
}

func latestNewsProviderUnavailableAnswerContract(payload LatestNewsLookupPayload) *LatestNewsAnswerContract {
	sources := payload.Sources
	provider := firstNonEmpty(payloadProvider(payload), "unknown")
	status := firstNonEmpty(payloadProviderStatus(payload), payload.AdapterStatus, payload.FailureCode, "unavailable")
	failureClass := firstNonEmpty(strings.TrimSpace(payload.FailureClass), payloadProviderFailureClass(payload))
	hint := ""
	if sources != nil {
		hint = firstNonEmpty(sources.FallbackHint, sources.FailureCode)
	}
	riskBoundary := "不能把模型猜测或空候选页包装成最新新闻结论。"
	lines := []string{
		"当前不能完成这次最新新闻查询。",
	}
	if target := latestNewsAnswerTarget(payload.Intent); target != "" {
		lines = append(lines, "查询目标："+target)
	}
	lines = append(lines, fmt.Sprintf("原因：搜索 provider=%s 当前状态为 %s，未拿到可打开核验的公开新闻来源。", provider, status))
	possibleImpact := ""
	if latestNewsIntentRequestsImpact(payload.Intent) {
		possibleImpact = "未获得可核验的事件证据，因此不能评估对公司、市场或政策的实际影响。"
		lines = append(lines, "可能影响："+possibleImpact)
	}
	lines = append(lines, "风险边界："+riskBoundary)
	if failureClass != "" {
		lines = append(lines, "失败分类："+failureClass+"。")
	}
	if payload.RetryAttemptCount > 0 {
		lines = append(lines, fmt.Sprintf("重试记录：已记录 %d 次失败尝试，retry_exhausted=%t。", payload.RetryAttemptCount, payload.RetryExhausted))
	}
	if strings.TrimSpace(payload.RetrySuppressedReason) != "" {
		lines = append(lines, "重试抑制："+strings.TrimSpace(payload.RetrySuppressedReason)+"。")
	}
	if strings.TrimSpace(hint) != "" {
		lines = append(lines, "下一步："+strings.TrimSpace(hint))
	} else {
		lines = append(lines, "下一步：配置可用搜索 provider，或直接提供明确新闻 URL 后再继续抽取和 guard。")
	}
	reason := "search_provider_unavailable"
	doNotRetry := []string{ToolLatestNewsExtract, ToolLatestNewsGuard, "open_page", "web_fetch", "browser"}
	if strings.EqualFold(strings.TrimSpace(failureClass), "quota_limited") {
		reason = "search_provider_quota_limited"
		doNotRetry = append([]string{ToolLatestNewsLookup, "search"}, doNotRetry...)
	} else if latestNewsProviderFailureClassTerminal(failureClass) {
		reason = "search_provider_config_invalid"
		doNotRetry = append([]string{ToolLatestNewsLookup, "search"}, doNotRetry...)
	}
	return &LatestNewsAnswerContract{
		FinalAnswerRecommended: true,
		Reason:                 reason,
		AllowedSummaryScope:    LatestNewsAnswerScopeProviderDiagnostic,
		DoNotRetryTools:        doNotRetry,
		PossibleImpact:         possibleImpact,
		RiskBoundary:           riskBoundary,
		FinalAnswerDraft:       strings.Join(lines, "\n"),
	}
}

func latestNewsSourceQualityAnswerContract(payload LatestNewsLookupPayload) *LatestNewsAnswerContract {
	sourceURL := firstNonEmpty(payload.SourceURL, payloadEvidence(payload).SourceURL, latestNewsPrimarySourceURL(payload), "unknown")
	publishedAt := firstNonEmpty(payload.PublishedAt, payloadEvidence(payload).PublishedAt, latestNewsPrimaryPublishedAt(payload), "unknown")
	summary := firstNonEmpty(payload.Summary, payloadTitle(payload))
	hasCandidateSource := latestNewsKnownValue(sourceURL)
	reasons := latestNewsSourceQualityReasons(payload)
	if len(reasons) == 0 {
		reasons = append(reasons, firstNonEmpty(payload.FailureCode, payload.GuardStatus, payload.AdapterStatus, "source_quality_needs_review"))
	}
	lines := []string{
		"当前不能把这次查询包装成已完整核验的最新新闻结论。",
	}
	if target := latestNewsAnswerTarget(payload.Intent); target != "" {
		lines = append(lines, "查询目标："+target)
	}
	if hasCandidateSource {
		lines = append(lines, "已发现候选来源："+sourceURL)
	} else {
		lines = append(lines, "未取得可打开且可 grounding 的候选来源。")
	}
	if latestNewsKnownValue(publishedAt) {
		lines = append(lines, "候选发布时间："+publishedAt)
	}
	if latestNewsKnownValue(summary) {
		lines = append(lines, "候选摘要："+ensureTerminalPeriod(summary))
	}
	riskBoundary := "只能说明已发现候选来源但正文 grounding、独立交叉核对或必要字段未完成；不要把搜索摘要、列表页或单源信息包装成已核实事实。"
	if !hasCandidateSource {
		riskBoundary = "未取得可打开且可 grounding 的候选来源；只能报告检索和来源质量诊断，不能据此推导任何事件事实或影响。"
	}
	possibleImpact := ""
	if latestNewsIntentRequestsImpact(payload.Intent) {
		if hasCandidateSource {
			possibleImpact = "独立交叉核对尚未完成，不能可靠判断候选事件的实际影响；需等待可打开且事件一致的独立来源。"
		} else {
			possibleImpact = "未取得可核验的事件证据，因此不能判断对公司、市场或政策的实际影响。"
		}
		lines = append(lines, "可能影响："+possibleImpact)
	}
	lines = append(lines,
		"未通过原因："+strings.Join(reasons, "、"),
		"风险边界："+riskBoundary,
	)
	return &LatestNewsAnswerContract{
		FinalAnswerRecommended: true,
		Reason:                 "source_quality_needs_review",
		AllowedSummaryScope:    LatestNewsAnswerScopeSourceDiagnostic,
		DoNotRetryTools:        []string{ToolLatestNewsLookup, ToolLatestNewsExtract, ToolLatestNewsGuard, "search", "open_page", "web_fetch", "browser"},
		PossibleImpact:         possibleImpact,
		RiskBoundary:           riskBoundary,
		FinalAnswerDraft:       strings.Join(lines, "\n"),
	}
}

func latestNewsLookupSourceQualityTerminal(payload LatestNewsLookupPayload) bool {
	if payload.Passed || strings.EqualFold(strings.TrimSpace(payload.GuardStatus), "passed") {
		return false
	}
	if latestNewsLookupProviderUnavailable(payload) {
		return false
	}
	reasons := latestNewsSourceQualityReasons(payload)
	if len(reasons) == 0 {
		return false
	}
	if latestNewsKnownValue(firstNonEmpty(payload.SourceURL, payloadEvidence(payload).SourceURL, latestNewsPrimarySourceURL(payload))) {
		return true
	}
	code := strings.ToLower(strings.TrimSpace(firstNonEmpty(payload.FailureCode, payload.GuardStatus, payload.AdapterStatus)))
	return strings.Contains(code, "no_grounded_sources") || strings.Contains(code, "missing_news_fields")
}

func latestNewsSourceQualityReasons(payload LatestNewsLookupPayload) []string {
	values := []string{}
	values = append(values, payload.MissingNewsFields...)
	values = append(values, payload.ReviewReasons...)
	if payload.Guard != nil {
		values = append(values, payload.Guard.MissingNewsFields...)
		values = append(values, payload.Guard.ReviewReasons...)
	}
	values = append(values, payload.Warnings...)
	if payload.Sources != nil {
		values = append(values, payload.Sources.Warnings...)
	}
	out := []string{}
	for _, value := range values {
		normalized := strings.ToLower(strings.TrimSpace(value))
		switch {
		case normalized == "grounded_page_text" ||
			normalized == "ungrounded_news_fields" ||
			normalized == "no_usable_source" ||
			normalized == "single_source_only" ||
			normalized == "cross_check_evidence_missing" ||
			normalized == "supporting_source_ungrounded" ||
			normalized == "latest_news_independent_supporting_source_missing" ||
			normalized == "latest_news_open_page_attempt_budget_exhausted" ||
			strings.HasPrefix(normalized, "evidence_review") ||
			strings.Contains(normalized, "missing_news_fields") ||
			strings.Contains(normalized, "cross_check_not_ready") ||
			strings.Contains(normalized, "ungrounded"):
			out = appendUniqueString(out, strings.TrimSpace(value))
		}
	}
	return out
}

func latestNewsLookupProviderUnavailable(payload LatestNewsLookupPayload) bool {
	status := strings.ToLower(strings.TrimSpace(payload.AdapterStatus))
	code := strings.ToLower(strings.TrimSpace(payload.FailureCode))
	failureClass := strings.ToLower(strings.TrimSpace(payload.FailureClass))
	if status == "provider_unavailable" || status == "provider_execution_failed" || status == "quota_limited" || status == "rate_limited" || status == "timeout" ||
		failureClass == "config_invalid" || failureClass == "auth_missing" || failureClass == "quota_limited" || failureClass == "rate_limited" || failureClass == "transient_network" || failureClass == "temporary_provider_error" {
		return true
	}
	if strings.Contains(code, "search_provider_failure") || strings.Contains(code, "provider_unavailable") || strings.Contains(code, "missing_credentials") || strings.Contains(code, "subscription_token_invalid") {
		return true
	}
	if payload.Sources == nil {
		return false
	}
	sourceStatus := strings.ToLower(strings.TrimSpace(firstNonEmpty(payload.Sources.FailureClass, payload.Sources.ErrorClass, payload.Sources.AdapterStatus, payload.Sources.ProviderStatus)))
	sourceCode := strings.ToLower(strings.TrimSpace(payload.Sources.FailureCode))
	return sourceStatus == "provider_unavailable" ||
		sourceStatus == "provider_execution_failed" ||
		sourceStatus == "missing_credentials" ||
		sourceStatus == "config_invalid" ||
		sourceStatus == "auth_missing" ||
		sourceStatus == "quota_limited" ||
		sourceStatus == "rate_limited" ||
		sourceStatus == "transient_network" ||
		strings.Contains(sourceCode, "search_provider_failure") ||
		strings.Contains(sourceCode, "missing_credentials") ||
		strings.Contains(sourceCode, "subscription_token_invalid")
}

func latestNewsSupportingSourceURLs(payload LatestNewsLookupPayload, limit int) []string {
	if limit <= 0 {
		return nil
	}
	values := []string{}
	if payload.Guard != nil {
		for _, source := range payload.Guard.SupportingSources {
			if value := strings.TrimSpace(source.SourceURL); value != "" && value != "unknown" {
				values = appendUniqueString(values, value)
			}
		}
	}
	if len(values) < limit && payload.Sources != nil {
		for _, source := range payload.Sources.SupportingSources {
			if value := strings.TrimSpace(source.SourceURL); value != "" && value != "unknown" {
				values = appendUniqueString(values, value)
			}
		}
	}
	if len(values) > limit {
		values = values[:limit]
	}
	return values
}

func payloadEvidence(payload LatestNewsLookupPayload) Evidence {
	if payload.Guard != nil {
		return payload.Guard.Evidence
	}
	if payload.Extract != nil {
		return payload.Extract.Evidence
	}
	return Evidence{}
}

func payloadTitle(payload LatestNewsLookupPayload) string {
	if payload.Guard != nil {
		return firstNonEmpty(payload.Guard.Title, payload.Guard.Evidence.Headline)
	}
	if payload.Extract != nil {
		return firstNonEmpty(payload.Extract.Title, payload.Extract.Evidence.Headline)
	}
	if payload.Sources != nil {
		return firstNonEmpty(payload.Sources.PrimarySource.Title, payload.Sources.PrimarySource.Headline)
	}
	return ""
}

func payloadProvider(payload LatestNewsLookupPayload) string {
	if payload.Sources == nil {
		return ""
	}
	return firstNonEmpty(payload.Sources.EffectiveProvider, payload.Sources.Provider, payload.Sources.RequestedProvider)
}

func payloadProviderStatus(payload LatestNewsLookupPayload) string {
	if payload.Sources == nil {
		return ""
	}
	return firstNonEmpty(payload.Sources.ProviderStatus, payload.Sources.FailureClass, payload.Sources.ErrorClass, payload.Sources.AdapterStatus)
}

func payloadProviderFailureClass(payload LatestNewsLookupPayload) string {
	if payload.Sources == nil {
		return ""
	}
	return strings.TrimSpace(payload.Sources.FailureClass)
}

func latestNewsProviderFailureClassTerminal(failureClass string) bool {
	switch strings.ToLower(strings.TrimSpace(failureClass)) {
	case "config_invalid", "auth_missing", "unsupported":
		return true
	default:
		return false
	}
}

func latestNewsPrimarySourceURL(payload LatestNewsLookupPayload) string {
	if payload.Sources == nil {
		return ""
	}
	return strings.TrimSpace(payload.Sources.PrimarySource.SourceURL)
}

func latestNewsPrimaryPublishedAt(payload LatestNewsLookupPayload) string {
	if payload.Sources == nil {
		return ""
	}
	return strings.TrimSpace(payload.Sources.PrimarySource.PublishedAt)
}

func latestNewsKnownValue(value string) bool {
	value = strings.TrimSpace(value)
	return value != "" && !strings.EqualFold(value, "unknown") && value != "-"
}

func ensureTerminalPeriod(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return value
	}
	last := []rune(value)[len([]rune(value))-1]
	switch last {
	case '.', '。', '！', '!', '？', '?':
		return value
	default:
		return value + "。"
	}
}
