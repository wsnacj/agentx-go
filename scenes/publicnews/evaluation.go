package publicnews

import (
	"net/url"
	"strings"
	"time"
)

// LatestNewsBriefEvaluationInput is the pack-level evidence shape consumed by
// host/plugin adapters and live evaluators. It is source-neutral: adapters
// own provider choice, page handling, and any news-site-specific extraction.
type LatestNewsBriefEvaluationInput struct {
	UserMessage            string
	TopicName              string
	SourceURL              string
	PublishedAt            string
	PublishedAfter         string
	NewsFields             []string
	MissingNewsFields      []string
	ReviewReasons          []string
	NewsFieldsReady        bool
	GuardStatus            string
	CrossCheckReady        bool
	ObservedSourceCount    int
	SourceEvidenceAccepted bool
	StopAfterGuardPassed   bool
}

// LatestNewsBriefEvaluation is the stable pack-level evaluator contract for
// public latest-news briefs. It intentionally does not know about specific
// publishers, search providers, or regional news heuristics.
type LatestNewsBriefEvaluation struct {
	Passed               bool     `json:"passed"`
	FreshnessConfirmed   bool     `json:"freshness_confirmed"`
	NewsFieldsReady      bool     `json:"news_fields_ready"`
	SourceAccepted       bool     `json:"source_accepted"`
	CrossCheckReady      bool     `json:"cross_check_ready"`
	StopAfterGuardPassed bool     `json:"stop_after_guard_passed"`
	MissingNewsFields    []string `json:"missing_news_fields"`
	ReviewReasons        []string `json:"review_reasons"`
	FailureReason        string   `json:"failure_reason"`
	SourceURL            string   `json:"source_url"`
	PublishedAt          string   `json:"published_at"`
	PublishedAfter       string   `json:"published_after,omitempty"`
}

func EvaluateLatestNewsBriefEvidence(input LatestNewsBriefEvaluationInput) LatestNewsBriefEvaluation {
	missing := normalizeLatestNewsStringList(input.MissingNewsFields)
	review := normalizeLatestNewsStringList(input.ReviewReasons)
	guardPassed := strings.EqualFold(strings.TrimSpace(input.GuardStatus), "passed")
	out := LatestNewsBriefEvaluation{
		FreshnessConfirmed:   LatestNewsFreshnessConfirmed(input.PublishedAt, input.PublishedAfter),
		NewsFieldsReady:      input.NewsFieldsReady && len(missing) == 0,
		SourceAccepted:       latestNewsSourceAccepted(input.SourceURL) && input.SourceEvidenceAccepted,
		CrossCheckReady:      input.CrossCheckReady,
		StopAfterGuardPassed: guardPassed && input.StopAfterGuardPassed,
		MissingNewsFields:    missing,
		ReviewReasons:        review,
		SourceURL:            strings.TrimSpace(input.SourceURL),
		PublishedAt:          strings.TrimSpace(input.PublishedAt),
		PublishedAfter:       strings.TrimSpace(input.PublishedAfter),
	}
	reasons := []string{}
	if !guardPassed {
		reasons = append(reasons, "guard_not_passed")
	}
	if !out.FreshnessConfirmed {
		if strings.TrimSpace(input.PublishedAt) == "" || strings.EqualFold(strings.TrimSpace(input.PublishedAt), "unknown") {
			reasons = append(reasons, "published_at_missing")
		} else {
			reasons = append(reasons, "published_at_outside_freshness_window")
		}
	}
	if !out.NewsFieldsReady {
		if len(missing) > 0 {
			reasons = append(reasons, "news_fields_missing")
		} else {
			reasons = append(reasons, "news_fields_not_ready")
		}
	}
	if !out.SourceAccepted {
		reasons = append(reasons, "source_unaccepted")
	}
	if !out.CrossCheckReady {
		reasons = append(reasons, "cross_check_not_ready")
	}
	if !out.StopAfterGuardPassed {
		reasons = append(reasons, "stop_after_guard_not_confirmed")
	}
	if len(review) > 0 {
		reasons = append(reasons, "review_required")
	}
	out.Passed = len(reasons) == 0
	out.FailureReason = strings.Join(reasons, ",")
	return out
}

// LatestNewsFreshnessConfirmed validates that primary evidence has a usable
// publication time and, when requested, falls on or after the cutoff.
func LatestNewsFreshnessConfirmed(publishedAt string, publishedAfter string) bool {
	publishedAt = strings.TrimSpace(publishedAt)
	if publishedAt == "" || strings.EqualFold(publishedAt, "unknown") {
		return false
	}
	publishedAfter = strings.TrimSpace(publishedAfter)
	if publishedAfter == "" {
		return true
	}
	cutoff, ok := parseLatestNewsTime(publishedAfter, time.UTC)
	if !ok {
		return false
	}
	published, ok := parseLatestNewsTime(publishedAt, cutoff.Location())
	return ok && !published.Before(cutoff)
}

func parseLatestNewsTime(value string, location *time.Location) (time.Time, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, false
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05 MST", "2006-01-02 15:04 MST"} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed, true
		}
	}
	if location == nil {
		location = time.UTC
	}
	for _, layout := range []string{"2006-01-02T15:04:05", "2006-01-02 15:04:05", "2006-01-02 15:04", "2006-01-02"} {
		if parsed, err := time.ParseInLocation(layout, value, location); err == nil {
			return parsed, true
		}
	}
	return time.Time{}, false
}

func latestNewsSourceAccepted(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" || strings.EqualFold(value, "unknown") {
		return false
	}
	parsed, err := url.Parse(value)
	return err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != ""
}

func normalizeLatestNewsStringList(values []string) []string {
	out := []string{}
	seen := map[string]bool{}
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
