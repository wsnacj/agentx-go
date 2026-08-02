package publicnews

import "strings"

// LatestNewsLookupAnswerReadiness projects full and bounded news outcomes into
// the common dimensional-readiness contract consumed by hosts and evaluators.
func LatestNewsLookupAnswerReadiness(payload LatestNewsLookupPayload) LatestNewsAnswerReadiness {
	ready := []string{}
	missing := []string{}
	if latestNewsKnownValue(firstNonEmpty(payload.SourceURL, payloadEvidence(payload).SourceURL, latestNewsPrimarySourceURL(payload))) {
		ready = appendUniqueString(ready, "candidate_source")
	}
	if latestNewsKnownValue(firstNonEmpty(payload.PublishedAt, payloadEvidence(payload).PublishedAt, latestNewsPrimaryPublishedAt(payload))) {
		ready = appendUniqueString(ready, "candidate_published_at")
	}
	if latestNewsKnownValue(firstNonEmpty(payload.Summary, payloadTitle(payload))) {
		ready = appendUniqueString(ready, "candidate_summary")
	}
	if payload.NewsFieldsReady {
		ready = appendUniqueString(ready, "news_fields")
	} else {
		missing = appendUniqueString(missing, "news_fields")
	}
	if payload.CrossCheckReady {
		ready = appendUniqueString(ready, "independent_cross_check")
	} else {
		missing = appendUniqueString(missing, "independent_cross_check")
	}
	if !latestNewsLookupGuardPrimaryGrounded(payload) {
		missing = appendUniqueString(missing, "grounded_primary_source")
	}

	if payload.Passed && strings.EqualFold(strings.TrimSpace(payload.GuardStatus), "passed") {
		return LatestNewsAnswerReadiness{
			AnswerReady:     true,
			SafeToAnswer:    true,
			AllowedScope:    LatestNewsAnswerScopeGuardedBrief,
			ReadyDimensions: ready,
		}
	}
	if len(ready) == 0 {
		ready = append(ready, "provider_diagnostics")
	}
	if len(missing) == 0 {
		missing = append(missing, "guarded_news_brief")
	}
	contract := payload.AnswerContract
	if contract == nil {
		contract = LatestNewsLookupAnswerContract(payload)
	}
	safeToAnswer := contract != nil && contract.FinalAnswerRecommended
	allowedScope := ""
	if contract != nil {
		allowedScope = contract.AllowedSummaryScope
	}
	return LatestNewsAnswerReadiness{
		AnswerReady:       false,
		SafeToAnswer:      safeToAnswer,
		Degraded:          safeToAnswer,
		DegradeReason:     firstNonEmpty(payload.FailureCode, payload.GuardStatus, payload.AdapterStatus),
		AllowedScope:      allowedScope,
		MissingDimensions: missing,
		ReadyDimensions:   ready,
		FailureCode:       payload.FailureCode,
		FailureClass:      payload.FailureClass,
	}
}
