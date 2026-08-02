package publicnews

import "strings"

// LatestNewsEvaluatorReport is a read-only quality projection for
// latest_news_lookup results. It does not change lookup decisions and should not
// encode publisher, topic, or region-specific policy.
type LatestNewsEvaluatorReport struct {
	FreshnessMatch       string   `json:"freshness_match,omitempty"`
	SourceIndependence   string   `json:"source_independence,omitempty"`
	GroundedTextPresence string   `json:"grounded_text_presence,omitempty"`
	FreshnessConfirmed   bool     `json:"freshness_confirmed"`
	SourceAccepted       bool     `json:"source_accepted"`
	CrossCheckReady      bool     `json:"cross_check_ready"`
	PrimaryGrounded      bool     `json:"primary_grounded"`
	SupportingGrounded   int      `json:"supporting_grounded_count,omitempty"`
	ObservedSourceCount  int      `json:"observed_source_count,omitempty"`
	Degraded             bool     `json:"degraded"`
	StopReason           string   `json:"stop_reason,omitempty"`
	DegradeReason        string   `json:"degrade_reason,omitempty"`
	FailureClass         string   `json:"failure_class,omitempty"`
	MissingNewsFields    []string `json:"missing_news_fields,omitempty"`
	ReviewReasons        []string `json:"review_reasons,omitempty"`
	Warnings             []string `json:"warnings,omitempty"`
}

func AttachLatestNewsEvaluatorReport(payload LatestNewsLookupPayload) LatestNewsLookupPayload {
	report := BuildLatestNewsEvaluatorReport(payload)
	payload.EvaluatorReport = &report
	return payload
}

func BuildLatestNewsEvaluatorReport(payload LatestNewsLookupPayload) LatestNewsEvaluatorReport {
	primaryGrounded := latestNewsLookupPrimaryGrounded(payload)
	supportingGrounded := latestNewsLookupSupportingGroundedCount(payload)
	observedSources := latestNewsLookupObservedSourceCount(payload)
	stopReason := latestNewsLookupStopReason(payload)
	degradeReason := latestNewsLookupDegradeReason(payload, stopReason)
	return LatestNewsEvaluatorReport{
		FreshnessMatch:       latestNewsLookupFreshnessMatch(payload),
		SourceIndependence:   latestNewsLookupSourceIndependence(payload, observedSources),
		GroundedTextPresence: latestNewsLookupGroundedTextPresence(primaryGrounded, supportingGrounded),
		FreshnessConfirmed:   payload.FreshnessConfirmed,
		SourceAccepted:       payload.SourceAccepted,
		CrossCheckReady:      payload.CrossCheckReady,
		PrimaryGrounded:      primaryGrounded,
		SupportingGrounded:   supportingGrounded,
		ObservedSourceCount:  observedSources,
		Degraded:             degradeReason != "",
		StopReason:           stopReason,
		DegradeReason:        degradeReason,
		FailureClass:         strings.TrimSpace(payload.FailureClass),
		MissingNewsFields:    append([]string(nil), payload.MissingNewsFields...),
		ReviewReasons:        append([]string(nil), payload.ReviewReasons...),
		Warnings:             append([]string(nil), payload.Warnings...),
	}
}

func latestNewsLookupFreshnessMatch(payload LatestNewsLookupPayload) string {
	if payload.FreshnessConfirmed {
		return "confirmed"
	}
	publishedAt := strings.TrimSpace(payload.PublishedAt)
	if publishedAt == "" || strings.EqualFold(publishedAt, "unknown") {
		return "unknown"
	}
	return "not_confirmed"
}

func latestNewsLookupSourceIndependence(payload LatestNewsLookupPayload, observedSources int) string {
	if payload.CrossCheckReady {
		return "cross_checked"
	}
	switch {
	case observedSources >= 2:
		return "multiple_sources_not_cross_checked"
	case observedSources == 1:
		return "single_source"
	default:
		return "no_source"
	}
}

func latestNewsLookupGroundedTextPresence(primaryGrounded bool, supportingGrounded int) string {
	switch {
	case primaryGrounded && supportingGrounded > 0:
		return "primary_and_supporting"
	case primaryGrounded:
		return "primary_only"
	case supportingGrounded > 0:
		return "supporting_only"
	default:
		return "missing"
	}
}

func latestNewsLookupPrimaryGrounded(payload LatestNewsLookupPayload) bool {
	if payload.Guard != nil {
		return payload.Guard.Evidence.GroundedTextAvailable
	}
	if payload.Extract != nil {
		return payload.Extract.Evidence.GroundedTextAvailable
	}
	if payload.Sources != nil {
		return strings.TrimSpace(payload.Sources.PrimarySource.Text) != ""
	}
	return false
}

func latestNewsLookupSupportingGroundedCount(payload LatestNewsLookupPayload) int {
	if payload.Guard != nil {
		count := 0
		for _, source := range payload.Guard.SupportingSources {
			if source.GroundedTextAvailable {
				count++
			}
		}
		return count
	}
	if payload.Sources == nil {
		return 0
	}
	count := 0
	for _, source := range payload.Sources.SupportingSources {
		if strings.TrimSpace(source.Text) != "" {
			count++
		}
	}
	return count
}

func latestNewsLookupObservedSourceCount(payload LatestNewsLookupPayload) int {
	if payload.Guard != nil && payload.Guard.ObservedSourceCount > 0 {
		return payload.Guard.ObservedSourceCount
	}
	if payload.Sources == nil {
		return 0
	}
	count := 0
	if sourceHasEvidence(payload.Sources.PrimarySource) {
		count++
	}
	for _, source := range payload.Sources.SupportingSources {
		if sourceHasEvidence(source) {
			count++
		}
	}
	return count
}

func sourceHasEvidence(source LatestNewsLookupSource) bool {
	return strings.TrimSpace(source.SourceURL) != "" ||
		strings.TrimSpace(source.Text) != "" ||
		strings.TrimSpace(source.Title) != "" ||
		strings.TrimSpace(source.Headline) != ""
}

func latestNewsLookupStopReason(payload LatestNewsLookupPayload) string {
	if code := strings.TrimSpace(payload.FailureCode); code != "" {
		return code
	}
	guardStatus := strings.TrimSpace(payload.GuardStatus)
	if guardStatus != "" && !strings.EqualFold(guardStatus, "passed") {
		return guardStatus
	}
	adapterStatus := strings.TrimSpace(payload.AdapterStatus)
	if adapterStatus != "" && !strings.EqualFold(adapterStatus, "ok") {
		return adapterStatus
	}
	if payload.Passed {
		return "guard_passed"
	}
	if guardStatus != "" {
		return guardStatus
	}
	return "unknown"
}

func latestNewsLookupDegradeReason(payload LatestNewsLookupPayload, stopReason string) string {
	if payload.Passed && strings.EqualFold(stopReason, "guard_passed") {
		return ""
	}
	if code := strings.TrimSpace(payload.FailureCode); code != "" {
		return code
	}
	if len(payload.MissingNewsFields) > 0 {
		return "missing_news_fields"
	}
	if len(payload.ReviewReasons) > 0 {
		return "review_required"
	}
	if reason := strings.TrimSpace(stopReason); reason != "" && !strings.EqualFold(reason, "guard_passed") {
		return reason
	}
	return "not_passed"
}
