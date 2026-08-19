package publicnews

import (
	"context"
	"strings"
)

// EvidenceReviewer is an optional host-owned review seam for grounded news
// evidence. Implementations may use an LLM, human review, cached classifiers, or
// another project-specific reviewer; newsbrief only owns the structured contract
// and deterministic guard merge behavior.
type EvidenceReviewer interface {
	ReviewEvidence(context.Context, EvidenceReviewInput) (EvidenceReviewResult, error)
}

type EvidenceReviewerFunc func(context.Context, EvidenceReviewInput) (EvidenceReviewResult, error)

func (fn EvidenceReviewerFunc) ReviewEvidence(ctx context.Context, input EvidenceReviewInput) (EvidenceReviewResult, error) {
	if fn == nil {
		return EvidenceReviewResult{}, nil
	}
	return fn(ctx, input)
}

type EvidenceReviewInput struct {
	Intent            LatestNewsLookupIntent `json:"intent,omitempty"`
	Headline          string                 `json:"headline,omitempty"`
	SourceURL         string                 `json:"source_url,omitempty"`
	SourceSite        string                 `json:"source_site,omitempty"`
	PublishedAt       string                 `json:"published_at,omitempty"`
	KeyUpdate         string                 `json:"key_update,omitempty"`
	Text              string                 `json:"text,omitempty"`
	SupportingSources []Evidence             `json:"supporting_sources,omitempty"`
}

type EvidenceReviewResult struct {
	Reviewed       bool     `json:"reviewed"`
	Accepted       bool     `json:"accepted"`
	CleanSummary   string   `json:"clean_summary,omitempty"`
	ReasonCodes    []string `json:"reason_codes,omitempty"`
	Confidence     float64  `json:"confidence,omitempty"`
	EvidenceSpans  []string `json:"evidence_spans,omitempty"`
	RequiresReview bool     `json:"requires_review,omitempty"`
	Source         string   `json:"source,omitempty"`
}

func ApplyEvidenceReview(ctx context.Context, payload Payload, source Context, intent LatestNewsLookupIntent, reviewer EvidenceReviewer) Payload {
	if reviewer == nil || payload.EvidenceReview != nil {
		return payload
	}
	if !payload.Evidence.GroundedTextAvailable || strings.TrimSpace(source.Text) == "" {
		return payload
	}
	result, err := reviewer.ReviewEvidence(ctx, EvidenceReviewInputFromPayload(source, payload, intent))
	if err != nil {
		payload.Warnings = appendUniqueString(payload.Warnings, "evidence_review_failed")
		return payload
	}
	result = normalizeEvidenceReviewResult(result)
	payload.EvidenceReview = &result
	if clean := strings.TrimSpace(result.CleanSummary); result.Accepted && clean != "" {
		if KeyUpdateSufficientForHeadline(payload.Evidence.Headline, clean) {
			payload.Evidence.KeyUpdate = clean
			payload.MissingNewsFields = removeStringValue(payload.MissingNewsFields, "key_update")
		} else {
			payload.ReviewReasons = appendUniqueString(payload.ReviewReasons, "evidence_review_clean_summary_low_quality")
		}
	}
	if !result.Accepted {
		payload.ReviewReasons = appendEvidenceReviewReasons(payload.ReviewReasons, result.ReasonCodes, "evidence_review_rejected")
	}
	if result.RequiresReview {
		payload.ReviewReasons = appendEvidenceReviewReasons(payload.ReviewReasons, result.ReasonCodes, "evidence_review_requires_review")
	}
	payload = refreshEvidenceReviewCrossCheckProjection(payload)
	payload.NewsFieldsReady = len(payload.MissingNewsFields) == 0
	switch {
	case len(payload.MissingNewsFields) > 0:
		payload.GuardStatus = "missing_news_fields"
	case len(payload.ReviewReasons) > 0:
		payload.GuardStatus = "needs_cross_check"
	default:
		payload.GuardStatus = "passed"
	}
	evaluation := BuildEvaluation(payload)
	payload.Evaluation = &evaluation
	return payload
}

func refreshEvidenceReviewCrossCheckProjection(payload Payload) Payload {
	sourceCount := CrossCheckSourceCount(payload.Evidence, payload.SupportingSources)
	payload.ObservedSourceCount = sourceCount
	payload.CrossCheckReady = sourceCount >= 2
	switch sourceCount {
	case 0:
		payload.ReviewReasons = removeStringValues(payload.ReviewReasons, "single_source_only")
		payload.ReviewReasons = appendUniqueString(payload.ReviewReasons, "no_usable_source")
		if len(payload.SupportingSources) > 0 {
			payload.ReviewReasons = appendUniqueString(payload.ReviewReasons, "cross_check_evidence_missing")
		}
	case 1:
		payload.ReviewReasons = removeStringValues(payload.ReviewReasons, "no_usable_source")
		payload.ReviewReasons = appendUniqueString(payload.ReviewReasons, "single_source_only")
		if len(payload.SupportingSources) > 0 {
			payload.ReviewReasons = appendUniqueString(payload.ReviewReasons, "cross_check_evidence_missing")
		}
	default:
		payload.ReviewReasons = removeStringValues(payload.ReviewReasons, "no_usable_source", "single_source_only", "cross_check_evidence_missing")
	}
	if UngroundedSupportingSourceCount(payload.SupportingSources) > 0 {
		payload.Warnings = appendUniqueString(payload.Warnings, "supporting_source_ungrounded_ignored")
		if sourceCount < 2 {
			payload.ReviewReasons = appendUniqueString(payload.ReviewReasons, "supporting_source_ungrounded")
		} else {
			payload.ReviewReasons = removeStringValues(payload.ReviewReasons, "supporting_source_ungrounded")
		}
	} else {
		payload.ReviewReasons = removeStringValues(payload.ReviewReasons, "supporting_source_ungrounded")
	}
	if NoisySupportingSourceCount(payload.SupportingSources) > 0 {
		payload.Warnings = appendUniqueString(payload.Warnings, "supporting_source_low_quality_ignored")
		if sourceCount < 2 {
			payload.ReviewReasons = appendUniqueString(payload.ReviewReasons, "supporting_source_low_quality")
		} else {
			payload.ReviewReasons = removeStringValues(payload.ReviewReasons, "supporting_source_low_quality")
		}
	} else {
		payload.ReviewReasons = removeStringValues(payload.ReviewReasons, "supporting_source_low_quality")
	}
	return payload
}

func EvidenceReviewInputFromPayload(source Context, payload Payload, intent LatestNewsLookupIntent) EvidenceReviewInput {
	return EvidenceReviewInput{
		Intent:            intent,
		Headline:          payload.Evidence.Headline,
		SourceURL:         firstNonEmpty(payload.Evidence.SourceURL, payload.FinalURL, source.SourceURL),
		SourceSite:        PreferredSourceSite(firstNonEmpty(payload.Evidence.SourceURL, payload.FinalURL, source.SourceURL), payload.Evidence.SourceSite),
		PublishedAt:       payload.Evidence.PublishedAt,
		KeyUpdate:         payload.Evidence.KeyUpdate,
		Text:              strings.TrimSpace(source.Text),
		SupportingSources: append([]Evidence(nil), payload.SupportingSources...),
	}
}

func normalizeEvidenceReviewResult(result EvidenceReviewResult) EvidenceReviewResult {
	result.Reviewed = true
	result.CleanSummary = strings.TrimSpace(result.CleanSummary)
	result.Source = strings.TrimSpace(result.Source)
	result.ReasonCodes = normalizeStringList(result.ReasonCodes)
	result.EvidenceSpans = normalizeStringList(result.EvidenceSpans)
	return result
}

func appendEvidenceReviewReasons(out []string, codes []string, fallback string) []string {
	if len(codes) == 0 {
		return appendUniqueString(out, fallback)
	}
	for _, code := range codes {
		code = strings.TrimSpace(code)
		if code == "" {
			continue
		}
		out = appendUniqueString(out, "evidence_review:"+code)
	}
	return out
}

func removeStringValues(values []string, targets ...string) []string {
	if len(values) == 0 || len(targets) == 0 {
		return values
	}
	targetSet := map[string]bool{}
	for _, target := range targets {
		target = strings.ToLower(strings.TrimSpace(target))
		if target != "" {
			targetSet[target] = true
		}
	}
	if len(targetSet) == 0 {
		return values
	}
	out := values[:0]
	for _, value := range values {
		if targetSet[strings.ToLower(strings.TrimSpace(value))] {
			continue
		}
		out = append(out, value)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func removeStringValue(values []string, target string) []string {
	target = strings.ToLower(strings.TrimSpace(target))
	if target == "" || len(values) == 0 {
		return values
	}
	out := values[:0]
	for _, value := range values {
		if strings.ToLower(strings.TrimSpace(value)) == target {
			continue
		}
		out = append(out, value)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
