package publicnews

import (
	"encoding/json"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

const (
	DefaultExtractSource = "host_latest_news_extract_tool"
	DefaultGuardSource   = "host_latest_news_guard_tool"
)

type PayloadOptions struct {
	ExtractSource string
	GuardSource   string
}

type Context struct {
	UserMessage string
	PageID      string
	Title       string
	SourceURL   string
	Text        string
}

type ContextResolver func(map[string]any) Context

type Evidence struct {
	PageID                string `json:"page_id,omitempty"`
	Headline              string `json:"headline"`
	SourceURL             string `json:"source_url"`
	SourceSite            string `json:"source_site"`
	PublishedAt           string `json:"published_at"`
	KeyUpdate             string `json:"key_update"`
	GroundedTextAvailable bool   `json:"grounded_text_available,omitempty"`
	groundedText          string
}

type Payload struct {
	Tool                string                     `json:"tool"`
	Source              string                     `json:"source"`
	PackID              string                     `json:"pack_id,omitempty"`
	CaseType            string                     `json:"case_type,omitempty"`
	WorkflowID          string                     `json:"workflow_id,omitempty"`
	PageID              string                     `json:"page_id,omitempty"`
	FinalURL            string                     `json:"final_url,omitempty"`
	Title               string                     `json:"title,omitempty"`
	NewsFields          []string                   `json:"news_fields,omitempty"`
	NewsFieldsReady     bool                       `json:"news_fields_ready"`
	MissingNewsFields   []string                   `json:"missing_news_fields,omitempty"`
	ReviewReasons       []string                   `json:"review_reasons,omitempty"`
	Warnings            []string                   `json:"warnings,omitempty"`
	GuardStatus         string                     `json:"guard_status,omitempty"`
	CrossCheckReady     bool                       `json:"cross_check_ready"`
	Evidence            Evidence                   `json:"evidence"`
	SupportingSources   []Evidence                 `json:"supporting_sources,omitempty"`
	ObservedSourceCount int                        `json:"observed_source_count,omitempty"`
	PublishedAfter      string                     `json:"published_after,omitempty"`
	Evaluation          *LatestNewsBriefEvaluation `json:"evaluation,omitempty"`
	EvidenceReview      *EvidenceReviewResult      `json:"evidence_review,omitempty"`
}

var publishedAtPattern = regexp.MustCompile(`(?im)(?:发布时间|published at)[:：]?\s*([0-9]{4}-[0-9]{2}-[0-9]{2}(?:\s+[0-9]{2}:[0-9]{2}(?::[0-9]{2})?(?:\s*[A-Z]{2,5})?)?)`)

func BuildExtractPayload(ctx Context, opts PayloadOptions) Payload {
	newsFields := []string{"published_at", "key_update", "source_url"}
	evidence := ExtractEvidence(ctx.Text, ctx.Title, ctx.SourceURL)
	evidence.PageID = strings.TrimSpace(ctx.PageID)
	evidence.GroundedTextAvailable = GroundedTextAvailable(ctx)
	payload := Payload{
		Tool:                ToolLatestNewsExtract,
		Source:              firstNonEmpty(opts.ExtractSource, DefaultExtractSource),
		PackID:              PackID,
		CaseType:            CaseTypeLatestBrief,
		WorkflowID:          DefaultWorkflow,
		PageID:              firstNonEmpty(ctx.PageID, "unknown"),
		FinalURL:            firstNonEmpty(ctx.SourceURL, "unknown"),
		Title:               firstNonEmpty(ctx.Title, evidence.Headline, "unknown"),
		NewsFields:          newsFields,
		Evidence:            evidence,
		MissingNewsFields:   MissingFields(evidence),
		Warnings:            ToolWarnings(ctx, evidence),
		ObservedSourceCount: 1,
	}
	payload.NewsFieldsReady = len(payload.MissingNewsFields) == 0
	return payload
}

func BuildGuardPayload(ctx Context, params map[string]any, resolver ContextResolver, opts PayloadOptions) Payload {
	extractPayload := BuildExtractPayload(ctx, opts)
	intent := LatestNewsLookupIntentFromParams(params)
	publishedAfter := LatestNewsPublishedAfter(intent)
	sourceURL := firstNonEmpty(StringArg(params["source_url"]), extractPayload.Evidence.SourceURL)
	evidence := Evidence{
		PageID:                strings.TrimSpace(ctx.PageID),
		Headline:              firstNonEmpty(StringArg(params["headline"]), extractPayload.Evidence.Headline),
		SourceURL:             sourceURL,
		SourceSite:            PreferredSourceSite(sourceURL, firstNonEmpty(StringArg(params["source_site"]), extractPayload.Evidence.SourceSite)),
		PublishedAt:           firstNonEmpty(StringArg(params["published_at"]), extractPayload.Evidence.PublishedAt),
		KeyUpdate:             firstNonEmpty(StringArg(params["key_update"]), extractPayload.Evidence.KeyUpdate),
		GroundedTextAvailable: GroundedTextAvailable(ctx),
		groundedText:          strings.TrimSpace(ctx.Text),
	}
	supportingSources, relevanceStats := supportingSourcesForPrimaryEvent(
		evidence,
		SupportingSourcesFromParams(params, resolver),
		intent,
		publishedAfter,
	)
	crossCheckStats := evaluateCrossCheckSources(evidence, supportingSources)
	sourceCount := crossCheckStats.uniqueCount
	claimedSourceCount := IntArg(params["source_count"])
	reviewReasons := []string{}
	missingFields := MissingFields(evidence)
	primaryCollection := SourceLooksLikeNewsCollection(evidence.Headline, evidence.groundedText)
	supportingCollectionCount := newsCollectionSupportingSourceCount(supportingSources)
	primaryAIGenerated := AIGeneratedSourceText(evidence.groundedText)
	supportingAIGeneratedCount := aiGeneratedSupportingSourceCount(supportingSources)
	if !GroundedTextAvailable(ctx) {
		missingFields = appendUniqueString(missingFields, "grounded_page_text")
		reviewReasons = appendUniqueString(reviewReasons, "ungrounded_news_fields")
	}
	switch sourceCount {
	case 0:
		reviewReasons = appendUniqueString(reviewReasons, "no_usable_source")
	case 1:
		reviewReasons = appendUniqueString(reviewReasons, "single_source_only")
	}
	if publishedAfter != "" &&
		strings.TrimSpace(evidence.PublishedAt) != "" &&
		!strings.EqualFold(strings.TrimSpace(evidence.PublishedAt), "unknown") &&
		!LatestNewsFreshnessConfirmed(evidence.PublishedAt, publishedAfter) {
		reviewReasons = appendUniqueString(reviewReasons, "primary_source_outside_freshness_window")
	}
	if evidenceUsesCommunitySurface(evidence) {
		reviewReasons = appendUniqueString(reviewReasons, "primary_source_community_surface")
	}
	if primaryCollection {
		reviewReasons = appendUniqueString(reviewReasons, "primary_source_collection_surface")
	}
	if primaryAIGenerated {
		reviewReasons = appendUniqueString(reviewReasons, "primary_source_ai_generated")
	}
	if relevanceStats.eventMismatchCount > 0 {
		reviewReasons = appendUniqueString(reviewReasons, "supporting_source_event_mismatch")
	}
	if relevanceStats.topicMismatchCount > 0 {
		reviewReasons = appendUniqueString(reviewReasons, "supporting_source_topic_mismatch")
	}
	if relevanceStats.freshnessMismatchCount > 0 {
		reviewReasons = appendUniqueString(reviewReasons, "supporting_source_outside_freshness_window")
	}
	if relevanceStats.communitySourceCount > 0 {
		reviewReasons = appendUniqueString(reviewReasons, "supporting_source_community_surface")
	}
	if claimedSourceCount >= 2 && sourceCount < 2 {
		reviewReasons = appendUniqueString(reviewReasons, "cross_check_evidence_missing")
	}
	if UngroundedSupportingSourceCount(supportingSources) > 0 {
		reviewReasons = appendUniqueString(reviewReasons, "supporting_source_ungrounded")
	}
	if NoisySupportingSourceCount(supportingSources) > 0 {
		reviewReasons = appendUniqueString(reviewReasons, "supporting_source_low_quality")
	}
	if supportingCollectionCount > 0 {
		reviewReasons = appendUniqueString(reviewReasons, "supporting_source_collection_surface")
	}
	if supportingAIGeneratedCount > 0 {
		reviewReasons = appendUniqueString(reviewReasons, "supporting_source_ai_generated")
	}
	if (crossCheckStats.syndicatedCopyCount > 0 || crossCheckStats.duplicatePublisherCount > 0 || crossCheckStats.attributedRepublicationCount > 0 || crossCheckStats.sharedUpstreamPublisherCount > 0) && sourceCount < 2 {
		reviewReasons = appendUniqueString(reviewReasons, "supporting_source_not_independent")
	}
	warnings := append([]string{}, extractPayload.Warnings...)
	if !GroundedTextAvailable(ctx) {
		warnings = appendUniqueString(warnings, "ungrounded_news_fields")
	}
	if EvidenceTextLooksEncodingCorrupt(evidence.KeyUpdate) {
		warnings = appendUniqueString(warnings, "key_update_encoding_noise")
	}
	if crossCheckStats.syndicatedCopyCount > 0 {
		warnings = appendUniqueString(warnings, "supporting_source_syndicated_copy_ignored")
	}
	if crossCheckStats.duplicatePublisherCount > 0 {
		warnings = appendUniqueString(warnings, "supporting_source_same_publisher_ignored")
	}
	if crossCheckStats.attributedRepublicationCount > 0 {
		warnings = appendUniqueString(warnings, "supporting_source_attributed_republication_ignored")
	}
	if crossCheckStats.sharedUpstreamPublisherCount > 0 {
		warnings = appendUniqueString(warnings, "supporting_source_shared_upstream_publisher_ignored")
	}
	if relevanceStats.eventMismatchCount > 0 {
		warnings = appendUniqueString(warnings, "supporting_source_event_mismatch_ignored")
	}
	if relevanceStats.topicMismatchCount > 0 {
		warnings = appendUniqueString(warnings, "supporting_source_topic_mismatch_ignored")
	}
	if relevanceStats.freshnessMismatchCount > 0 {
		warnings = appendUniqueString(warnings, "supporting_source_outside_freshness_window_ignored")
	}
	if relevanceStats.communitySourceCount > 0 {
		warnings = appendUniqueString(warnings, "supporting_source_community_surface_ignored")
	}
	if primaryCollection {
		warnings = appendUniqueString(warnings, "primary_source_collection_surface_ignored")
	}
	if supportingCollectionCount > 0 {
		warnings = appendUniqueString(warnings, "supporting_source_collection_surface_ignored")
	}
	if primaryAIGenerated {
		warnings = appendUniqueString(warnings, "primary_source_ai_generated_ignored")
	}
	if supportingAIGeneratedCount > 0 {
		warnings = appendUniqueString(warnings, "supporting_source_ai_generated_ignored")
	}
	payload := Payload{
		Tool:                ToolLatestNewsGuard,
		Source:              firstNonEmpty(opts.GuardSource, DefaultGuardSource),
		PackID:              PackID,
		CaseType:            CaseTypeLatestBrief,
		WorkflowID:          DefaultWorkflow,
		PageID:              extractPayload.PageID,
		FinalURL:            firstNonEmpty(evidence.SourceURL, extractPayload.FinalURL, "unknown"),
		Title:               firstNonEmpty(evidence.Headline, extractPayload.Title, "unknown"),
		NewsFields:          extractPayload.NewsFields,
		MissingNewsFields:   missingFields,
		ReviewReasons:       reviewReasons,
		Warnings:            warnings,
		Evidence:            evidence,
		SupportingSources:   supportingSources,
		ObservedSourceCount: sourceCount,
		CrossCheckReady:     sourceCount >= 2,
		PublishedAfter:      publishedAfter,
	}
	switch {
	case len(missingFields) > 0:
		payload.GuardStatus = "missing_news_fields"
	case len(reviewReasons) > 0:
		payload.GuardStatus = "needs_cross_check"
	default:
		payload.GuardStatus = "passed"
	}
	payload.NewsFieldsReady = len(missingFields) == 0
	evaluation := BuildEvaluation(payload)
	payload.Evaluation = &evaluation
	return payload
}

func ExtractEvidence(text string, title string, sourceURL string) Evidence {
	return ExtractEvidenceWithPolicy(text, title, sourceURL, DefaultEvidenceQualityPolicy())
}

func ExtractEvidenceWithPolicy(text string, title string, sourceURL string, policy EvidenceQualityPolicy) Evidence {
	return ExtractEvidenceWithPolicyAndFilter(text, title, sourceURL, policy, nil)
}

// ExtractEvidenceWithPolicyAndFilter returns the first quality-approved line
// that also satisfies the caller-owned relevance filter.
func ExtractEvidenceWithPolicyAndFilter(text string, title string, sourceURL string, policy EvidenceQualityPolicy, lineFilter func(string) bool) Evidence {
	return extractEvidenceWithPolicy(text, title, sourceURL, policy, lineFilter, nil)
}

// ExtractEvidenceWithPolicyAndScorer selects the highest-scoring approved
// sentence. A non-positive score excludes the sentence.
func ExtractEvidenceWithPolicyAndScorer(text string, title string, sourceURL string, policy EvidenceQualityPolicy, lineScore func(string) int) Evidence {
	return extractEvidenceWithPolicy(text, title, sourceURL, policy, nil, lineScore)
}

func extractEvidenceWithPolicy(text string, title string, sourceURL string, policy EvidenceQualityPolicy, lineFilter func(string) bool, lineScore func(string) int) Evidence {
	if policy == nil {
		policy = DefaultEvidenceQualityPolicy()
	}
	evidence := Evidence{
		Headline:   firstNonEmpty(strings.TrimSpace(title), "unknown"),
		SourceURL:  firstNonEmpty(strings.TrimSpace(sourceURL), "unknown"),
		SourceSite: SourceSite(sourceURL),
	}
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return evidence
	}
	if match := publishedAtPattern.FindStringSubmatch(trimmed); len(match) > 1 {
		evidence.PublishedAt = strings.TrimSpace(match[1])
	}
	bestScore := 0
	for _, line := range strings.Split(trimmed, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		for _, segment := range evidenceLineSegments(line) {
			if !policy.EvaluateLine(EvidenceQualityInput{Headline: title, Line: segment, KeyUpdate: segment}).Accepted ||
				(lineFilter != nil && !lineFilter(segment)) {
				continue
			}
			if lineScore != nil {
				score := lineScore(segment)
				if score > bestScore {
					bestScore = score
					evidence.KeyUpdate = segment
				}
				continue
			}
			evidence.KeyUpdate = segment
			break
		}
		if lineScore == nil && evidence.KeyUpdate != "" {
			break
		}
	}
	if evidence.PublishedAt == "" {
		evidence.PublishedAt = "unknown"
	}
	if evidence.KeyUpdate == "" {
		evidence.KeyUpdate = "unknown"
	}
	return evidence
}

func evidenceLineSegments(line string) []string {
	runes := []rune(strings.TrimSpace(line))
	if len(runes) == 0 {
		return nil
	}
	out := []string{}
	start := 0
	appendSegment := func(end int) {
		if segment := strings.TrimSpace(string(runes[start:end])); segment != "" {
			out = append(out, segment)
		}
		start = end
	}
	for idx, r := range runes {
		split := strings.ContainsRune("。！？；", r)
		if !split && (r == '.' || r == '!' || r == '?' || r == ';') {
			nextBoundary := idx+1 == len(runes) || unicode.IsSpace(runes[idx+1])
			decimalPoint := r == '.' && idx > 0 && idx+1 < len(runes) && unicode.IsDigit(runes[idx-1]) && unicode.IsDigit(runes[idx+1])
			split = nextBoundary && !decimalPoint && idx-start >= 12
		}
		if split {
			appendSegment(idx + 1)
		}
	}
	appendSegment(len(runes))
	return out
}

func BuildEvaluation(payload Payload) LatestNewsBriefEvaluation {
	return EvaluateLatestNewsBriefEvidence(LatestNewsBriefEvaluationInput{
		SourceURL:              firstNonEmpty(payload.Evidence.SourceURL, payload.FinalURL),
		PublishedAt:            payload.Evidence.PublishedAt,
		PublishedAfter:         payload.PublishedAfter,
		NewsFields:             payload.NewsFields,
		MissingNewsFields:      payload.MissingNewsFields,
		ReviewReasons:          payload.ReviewReasons,
		NewsFieldsReady:        payload.NewsFieldsReady,
		GuardStatus:            payload.GuardStatus,
		CrossCheckReady:        payload.CrossCheckReady,
		ObservedSourceCount:    payload.ObservedSourceCount,
		SourceEvidenceAccepted: EvidenceUsableForCrossCheck(payload.Evidence),
		StopAfterGuardPassed:   strings.EqualFold(strings.TrimSpace(payload.GuardStatus), "passed"),
	})
}

type supportingSourceRelevanceStats struct {
	eventMismatchCount     int
	topicMismatchCount     int
	freshnessMismatchCount int
	communitySourceCount   int
}

func supportingSourcesForPrimaryEvent(primary Evidence, supporting []Evidence, intent LatestNewsLookupIntent, publishedAfter string) ([]Evidence, supportingSourceRelevanceStats) {
	stats := supportingSourceRelevanceStats{}
	primaryEventReady := KeyUpdateSufficientForHeadline(primary.Headline, primary.KeyUpdate)
	primarySource := lookupSourceFromEvidence(primary)
	out := make([]Evidence, 0, len(supporting))
	for _, evidence := range supporting {
		if evidenceUsesCommunitySurface(evidence) {
			stats.communitySourceCount++
			continue
		}
		if publishedAfter != "" && !LatestNewsFreshnessConfirmed(evidence.PublishedAt, publishedAfter) {
			stats.freshnessMismatchCount++
			continue
		}
		if !primaryEventReady || !KeyUpdateSufficientForHeadline(evidence.Headline, evidence.KeyUpdate) {
			out = append(out, evidence)
			continue
		}
		decision := DefaultSourceRelevancePolicy().EvaluateSupportingSource(SourceRelevanceInput{
			Primary:   primarySource,
			Candidate: lookupSourceFromEvidence(evidence),
			Intent:    intent,
		})
		switch decision.RuleID {
		case SourceRelevanceRuleEventCoherenceNeeded:
			stats.eventMismatchCount++
			continue
		case SourceRelevanceRuleTopicSpecificSupportNeeded:
			stats.topicMismatchCount++
			continue
		}
		out = append(out, evidence)
	}
	return out, stats
}

func lookupSourceFromEvidence(evidence Evidence) LatestNewsLookupSource {
	return LatestNewsLookupSource{
		PageID:      evidence.PageID,
		Title:       evidence.Headline,
		Headline:    evidence.Headline,
		SourceURL:   evidence.SourceURL,
		SourceSite:  evidence.SourceSite,
		PublishedAt: evidence.PublishedAt,
		KeyUpdate:   evidence.KeyUpdate,
		Text:        evidence.groundedText,
	}
}

func DecodePayload(raw string) (Payload, bool) {
	var payload Payload
	if err := json.Unmarshal([]byte(strings.TrimSpace(raw)), &payload); err != nil {
		return Payload{}, false
	}
	return payload, true
}

func GroundedTextAvailable(ctx Context) bool {
	text := strings.TrimSpace(ctx.Text)
	return text != "" && !strings.EqualFold(text, "unknown")
}

func SupportingSourcesFromParams(params map[string]any, resolver ContextResolver) []Evidence {
	raw, ok := params["supporting_sources"]
	if !ok || raw == nil {
		raw = params["sources"]
	}
	values, ok := anySlice(raw)
	if !ok {
		return nil
	}
	if resolver == nil {
		resolver = contextFromParams
	}
	out := []Evidence{}
	for _, value := range values {
		item, ok := value.(map[string]any)
		if !ok {
			continue
		}
		ctx := resolver(map[string]any{
			"page_id":    item["page_id"],
			"text":       item["text"],
			"title":      firstNonEmpty(StringArg(item["title"]), StringArg(item["headline"])),
			"source_url": item["source_url"],
		})
		extracted := ExtractEvidence(ctx.Text, ctx.Title, ctx.SourceURL)
		sourceURL := firstNonEmpty(StringArg(item["source_url"]), extracted.SourceURL)
		evidence := Evidence{
			PageID:                strings.TrimSpace(ctx.PageID),
			Headline:              firstNonEmpty(StringArg(item["headline"]), extracted.Headline),
			SourceURL:             sourceURL,
			SourceSite:            strings.TrimSpace(StringArg(item["source_site"])),
			PublishedAt:           firstNonEmpty(StringArg(item["published_at"]), extracted.PublishedAt),
			KeyUpdate:             firstNonEmpty(StringArg(item["key_update"]), extracted.KeyUpdate),
			GroundedTextAvailable: GroundedTextAvailable(ctx),
			groundedText:          strings.TrimSpace(ctx.Text),
		}
		evidence.SourceSite = PreferredSourceSite(evidence.SourceURL, evidence.SourceSite)
		if CrossCheckSourceKey(evidence) == "" {
			continue
		}
		out = append(out, evidence)
	}
	return out
}

func CrossCheckSourceCount(primary Evidence, supporting []Evidence) int {
	return evaluateCrossCheckSources(primary, supporting).uniqueCount
}

func EvidenceUsableForCrossCheck(e Evidence) bool {
	if !e.GroundedTextAvailable {
		return false
	}
	if evidenceUsesCommunitySurface(e) {
		return false
	}
	if SourceLooksLikeNewsCollection(e.Headline, e.groundedText) {
		return false
	}
	if AIGeneratedSourceText(e.groundedText) {
		return false
	}
	if CrossCheckSourceKey(e) == "" {
		return false
	}
	if EvidenceLooksNoisy(e) {
		return false
	}
	return firstNonEmpty(strings.TrimSpace(e.PublishedAt), strings.TrimSpace(e.KeyUpdate), strings.TrimSpace(e.Headline), "unknown") != "unknown"
}

func newsCollectionSupportingSourceCount(supporting []Evidence) int {
	count := 0
	for _, item := range supporting {
		if CrossCheckSourceKey(item) != "" && SourceLooksLikeNewsCollection(item.Headline, item.groundedText) {
			count++
		}
	}
	return count
}

func aiGeneratedSupportingSourceCount(supporting []Evidence) int {
	count := 0
	for _, item := range supporting {
		if CrossCheckSourceKey(item) != "" && AIGeneratedSourceText(item.groundedText) {
			count++
		}
	}
	return count
}

func UngroundedSupportingSourceCount(supporting []Evidence) int {
	count := 0
	for _, item := range supporting {
		if CrossCheckSourceKey(item) != "" && !item.GroundedTextAvailable {
			count++
		}
	}
	return count
}

func NoisySupportingSourceCount(supporting []Evidence) int {
	count := 0
	for _, item := range supporting {
		if CrossCheckSourceKey(item) != "" && EvidenceLooksNoisy(item) {
			count++
		}
	}
	return count
}

func EvidenceLooksNoisy(e Evidence) bool {
	for _, value := range []string{e.Headline, e.KeyUpdate} {
		if EvidenceTextLooksEncodingCorrupt(value) {
			return true
		}
	}
	if known := strings.TrimSpace(e.KeyUpdate); known != "" && !strings.EqualFold(known, "unknown") {
		return !KeyUpdateSufficientForHeadline(e.Headline, known)
	}
	return false
}

func CrossCheckSourceKey(e Evidence) string {
	if value := SourcePublisherFamily(e.SourceURL); value != "" && value != "unknown" {
		return "publisher:" + value
	}
	if value := PublisherHostFamily(e.SourceSite); value != "" && value != "unknown" {
		return "publisher:" + value
	}
	return ""
}

func PreferredSourceSite(sourceURL string, claimedSite string) string {
	if value := SourceSite(sourceURL); value != "" && value != "unknown" {
		return value
	}
	if value := NormalizeSourceHost(claimedSite); value != "" && value != "unknown" {
		return value
	}
	return "unknown"
}

func MissingFields(e Evidence) []string {
	missing := []string{}
	if firstNonEmpty(strings.TrimSpace(e.PublishedAt), "unknown") == "unknown" {
		missing = append(missing, "published_at")
	}
	if !KeyUpdateSufficientForHeadline(e.Headline, e.KeyUpdate) {
		missing = append(missing, "key_update")
	}
	if firstNonEmpty(strings.TrimSpace(e.SourceURL), "unknown") == "unknown" {
		missing = append(missing, "source_url")
	}
	return missing
}

func ToolWarnings(ctx Context, e Evidence) []string {
	var out []string
	if strings.TrimSpace(ctx.Text) == "" {
		out = append(out, "page_text_missing")
	}
	if e.PublishedAt == "unknown" {
		out = append(out, "published_at_missing")
	}
	if e.KeyUpdate == "unknown" {
		out = append(out, "key_update_missing")
	} else if EvidenceTextLooksEncodingCorrupt(e.KeyUpdate) {
		out = append(out, "key_update_encoding_noise")
	} else if !KeyUpdateSufficientForHeadline(e.Headline, e.KeyUpdate) {
		out = append(out, "key_update_low_quality")
	}
	return out
}

func SourceSite(sourceURL string) string {
	trimmed := strings.TrimSpace(sourceURL)
	if trimmed == "" {
		return "unknown"
	}
	parsed, err := url.Parse(trimmed)
	if err != nil || strings.TrimSpace(parsed.Hostname()) == "" {
		return "unknown"
	}
	return NormalizeSourceHost(parsed.Hostname())
}

func NormalizeSourceHost(host string) string {
	value := strings.ToLower(strings.TrimSpace(host))
	value = strings.TrimSuffix(value, ".")
	value = strings.TrimPrefix(value, "www.")
	if value == "" {
		return "unknown"
	}
	return value
}

func StringArg(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case fmtStringer:
		return typed.String()
	case nil:
		return ""
	default:
		return strings.TrimSpace(strings.Trim(strings.ReplaceAll(strings.TrimSpace(toJSON(typed)), "\n", " "), `"`))
	}
}

func IntArg(v any) int {
	switch value := v.(type) {
	case int:
		return value
	case int64:
		return int(value)
	case float64:
		return int(value)
	case json.Number:
		if parsed, err := value.Int64(); err == nil {
			return int(parsed)
		}
	case string:
		value = strings.TrimSpace(value)
		if value == "" {
			return 0
		}
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return 0
}

func contextFromParams(params map[string]any) Context {
	return Context{
		UserMessage: strings.TrimSpace(StringArg(params["user_message"])),
		PageID:      strings.TrimSpace(StringArg(params["page_id"])),
		Title:       strings.TrimSpace(StringArg(params["title"])),
		SourceURL:   strings.TrimSpace(StringArg(params["source_url"])),
		Text:        strings.TrimSpace(StringArg(params["text"])),
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
	case string:
		value = strings.TrimSpace(value)
		if value == "" {
			return nil, false
		}
		var decoded []any
		if err := json.Unmarshal([]byte(value), &decoded); err != nil {
			return nil, false
		}
		return decoded, true
	default:
		return nil, false
	}
}

func appendUniqueString(values []string, value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return values
	}
	for _, item := range values {
		if strings.TrimSpace(item) == value {
			return values
		}
	}
	return append(values, value)
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

func containsAnyFold(text string, needles ...string) bool {
	lower := strings.ToLower(text)
	for _, needle := range needles {
		if strings.Contains(lower, strings.ToLower(needle)) {
			return true
		}
	}
	return false
}

type fmtStringer interface {
	String() string
}

func toJSON(value any) string {
	blob, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(blob)
}
