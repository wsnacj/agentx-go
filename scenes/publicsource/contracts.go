package publicsource

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"strconv"
	"strings"
	"time"

	control "github.com/wsnacj/agentx-go/runtime/controlcontract"
)

const (
	DefaultAdapterRef  control.DisplaySafeRef = "adapter:public_source_retrieval"
	DefaultStrategyRef control.DisplaySafeRef = "strategy:public_source_retrieval"
	DefaultRunRef      control.DisplaySafeRef = "adapter_run:public_source_retrieval"
	DefaultEvidenceRef control.DisplaySafeRef = "evidence:public_source_result"
)

// Collector is the only source-acquisition runtime port. Implementations may
// use HTTP, a browser, a fixture, a cache, or another host service.
type Collector interface {
	CollectPublicSourceEvidence(context.Context, Request) (Report, error)
}

// Request carries only display-safe identity and policy references.
type Request struct {
	RuntimeRequest control.RuntimeAdapterExecutionRequest `json:"runtime_request,omitempty"`
	QueryRefs      []control.DisplaySafeRef               `json:"query_refs,omitempty"`
	SourceRefs     []control.DisplaySafeRef               `json:"source_refs,omitempty"`
	PolicyRefs     []control.DisplaySafeRef               `json:"policy_refs,omitempty"`
	ObservedAt     string                                 `json:"observed_at,omitempty"`
	Boundaries     []control.Boundary                     `json:"boundaries,omitempty"`
}

// Evidence is one display-safe source observation.
type Evidence struct {
	SourceRef      control.DisplaySafeRef   `json:"source_ref,omitempty"`
	QueryRef       control.DisplaySafeRef   `json:"query_ref,omitempty"`
	EvidenceRef    control.DisplaySafeRef   `json:"evidence_ref,omitempty"`
	Kind           string                   `json:"kind,omitempty"`
	Strength       control.EvidenceStrength `json:"strength,omitempty"`
	ObservedAt     string                   `json:"observed_at,omitempty"`
	Confidence     string                   `json:"confidence,omitempty"`
	ConfidenceRef  control.DisplaySafeRef   `json:"confidence_ref,omitempty"`
	DisplaySafeRef control.DisplaySafeRef   `json:"display_safe_ref,omitempty"`
}

// DisplaySummary is Host-attested display-safe text. Raw URLs and page bodies
// must remain in the Host fetcher.
type DisplaySummary struct {
	SourceRef           control.DisplaySafeRef `json:"source_ref,omitempty"`
	QueryRef            control.DisplaySafeRef `json:"query_ref,omitempty"`
	EvidenceRef         control.DisplaySafeRef `json:"evidence_ref,omitempty"`
	SummaryRef          control.DisplaySafeRef `json:"summary_ref,omitempty"`
	AttestationRef      control.DisplaySafeRef `json:"attestation_ref,omitempty"`
	RedactionRef        control.DisplaySafeRef `json:"redaction_ref,omitempty"`
	DisplayPolicyRef    control.DisplaySafeRef `json:"display_policy_ref,omitempty"`
	DisplaySafeAttested bool                   `json:"display_safe_attested"`
	Title               string                 `json:"title,omitempty"`
	Summary             string                 `json:"summary,omitempty"`
	Snippet             string                 `json:"snippet,omitempty"`
	Published           string                 `json:"published,omitempty"`
	ObservedAt          string                 `json:"observed_at,omitempty"`
	Confidence          string                 `json:"confidence,omitempty"`
	Boundaries          []control.Boundary     `json:"boundaries,omitempty"`
}

// Report is the provider-neutral source-acquisition readback.
type Report struct {
	Status             control.VerificationStatus `json:"status,omitempty"`
	FailureClass       control.FailureClass       `json:"failure_class,omitempty"`
	FailureReason      string                     `json:"failure_reason,omitempty"`
	ObservedAt         string                     `json:"observed_at,omitempty"`
	QueryEvidenceRefs  []control.DisplaySafeRef   `json:"query_evidence_refs,omitempty"`
	SourceRefs         []control.DisplaySafeRef   `json:"source_refs,omitempty"`
	Evidence           []Evidence                 `json:"evidence,omitempty"`
	DisplaySummaries   []DisplaySummary           `json:"display_summaries,omitempty"`
	UnavailableReasons []string                   `json:"unavailable_reasons,omitempty"`
	Boundaries         []control.Boundary         `json:"boundaries,omitempty"`
	RawOutputLoaded    bool                       `json:"raw_output_loaded"`
}

// SearchResult is the portable subset of a Host search result.
type SearchResult struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Published   string `json:"published,omitempty"`
	SiteName    string `json:"site_name,omitempty"`
}

// SearchPayload is the portable subset consumed by deterministic projection.
type SearchPayload struct {
	Query    string         `json:"query"`
	Provider string         `json:"provider"`
	Count    int            `json:"count"`
	TookMs   int64          `json:"took_ms"`
	Results  []SearchResult `json:"results,omitempty"`
	Cached   bool           `json:"cached,omitempty"`
}

// SearchReportInput maps Host search output to canonical evidence.
type SearchReportInput struct {
	Payload          SearchPayload
	DisplaySummaries []DisplaySummary
	QueryRef         control.DisplaySafeRef
	SourcePolicy     SourcePolicy
	ObservedAt       string
	Strength         control.EvidenceStrength
}

// Normalize returns a detached fail-closed report.
func (report Report) Normalize() Report {
	out := report
	out.Status = control.NormalizeVerificationStatus(string(out.Status))
	if out.Status == control.VerificationNotEvaluated {
		out.Status = control.VerificationBlocked
	}
	out.FailureClass = control.NormalizeFailureClass(string(out.FailureClass))
	out.FailureReason = controlToken(out.FailureReason)
	out.ObservedAt = strings.TrimSpace(out.ObservedAt)
	out.QueryEvidenceRefs = refs(out.QueryEvidenceRefs)
	out.SourceRefs = refs(out.SourceRefs)
	out.Evidence = normalizeEvidence(out.Evidence, out.ObservedAt)
	for _, evidence := range out.Evidence {
		out.SourceRefs = appendRef(out.SourceRefs, evidence.SourceRef)
		out.QueryEvidenceRefs = appendRef(out.QueryEvidenceRefs, evidence.QueryRef)
	}
	out.DisplaySummaries = normalizeSummaries(out.DisplaySummaries, out.ObservedAt)
	for _, summary := range out.DisplaySummaries {
		out.SourceRefs = appendRef(out.SourceRefs, summary.SourceRef)
		out.QueryEvidenceRefs = appendRef(out.QueryEvidenceRefs, summary.QueryRef)
	}
	out.UnavailableReasons = controlTokens(out.UnavailableReasons)
	out.Boundaries = control.AppendBoundaries(nil, out.Boundaries...)
	if out.RawOutputLoaded {
		out.Status = control.VerificationReviewRequired
		if out.FailureClass == control.FailureNone {
			out.FailureClass = control.FailureEvidenceWeak
		}
		out.Boundaries = control.AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
	}
	if out.Status == control.VerificationSatisfied && len(out.Evidence) == 0 {
		out.Status = control.VerificationBlocked
		out.FailureClass = control.FailureEvidenceMissing
		out.FailureReason = "public_source_evidence_missing"
	}
	if out.Status != control.VerificationSatisfied && out.FailureClass == control.FailureNone {
		out.FailureClass = control.FailureTargetUnavailable
	}
	return out
}

// BuildReportFromSearch deterministically maps Host search metadata to a
// display-safe report. It never fetches URLs.
func BuildReportFromSearch(input SearchReportInput) Report {
	payload, summaries, decision := input.SourcePolicy.Filter(input.Payload, input.DisplaySummaries)
	observedAt := strings.TrimSpace(input.ObservedAt)
	if observedAt == "" {
		observedAt = time.Now().UTC().Format(time.RFC3339)
	}
	strength := control.NormalizeEvidenceStrength(string(input.Strength))
	if strength == control.EvidenceMissing {
		strength = control.EvidenceAdequate
	}
	queryRef := normalizeRef(input.QueryRef)
	if queryRef == "" {
		queryRef = makeRef("query", payload.Query)
	}
	evidenceRef := QueryResultEvidenceRef(queryRef)
	report := Report{
		Status: control.VerificationBlocked, FailureClass: control.FailureTargetUnavailable,
		FailureReason: "public_source_results_empty", ObservedAt: observedAt,
		QueryEvidenceRefs: []control.DisplaySafeRef{queryRef},
		Boundaries:        []control.Boundary{"public_source_search_payload_mapped", control.Boundary("public_source_evidence_strength_" + string(strength)), "public_source_evidence_strength_declared_by_host_collector", "raw_urls_retained_by_host"},
	}
	if decision.Applied {
		report.Boundaries = control.AppendBoundaries(report.Boundaries, decision.Boundaries...)
	}
	for index, item := range payload.Results {
		sourceRef := resultRef(item.URL, item.SiteName, index+1)
		if sourceRef == "" {
			continue
		}
		report.SourceRefs = appendRef(report.SourceRefs, sourceRef)
		report.Evidence = append(report.Evidence, Evidence{
			SourceRef: sourceRef, QueryRef: queryRef, EvidenceRef: evidenceRef,
			Kind: "public_source_result", Strength: strength, ObservedAt: observedAt,
			Confidence: confidence(index + 1), ConfidenceRef: makeRef("confidence:public_source", string(sourceRef)),
			DisplaySafeRef: makeRef("result:public_source", string(sourceRef)),
		})
	}
	for index, item := range summaries {
		var fallback control.DisplaySafeRef
		if index < len(report.SourceRefs) {
			fallback = report.SourceRefs[index]
		}
		summary := normalizeSummary(item, index+1, observedAt, queryRef, evidenceRef, fallback)
		if summary.SummaryRef == "" {
			report.Boundaries = control.AppendBoundaries(report.Boundaries, "public_source_display_summary_dropped")
			continue
		}
		report.DisplaySummaries = append(report.DisplaySummaries, summary)
	}
	if len(report.Evidence) > 0 {
		report.Status, report.FailureClass, report.FailureReason = control.VerificationSatisfied, control.FailureNone, ""
	}
	if decision.Applied && !decision.Allowed {
		report.Status, report.FailureClass, report.FailureReason = control.VerificationBlocked, decision.FailureClass, decision.FailureReason
		report.UnavailableReasons = []string{decision.FailureReason}
	}
	return report.Normalize()
}

// QueryResultEvidenceRef derives the stable evidence identity for a query.
func QueryResultEvidenceRef(queryRef control.DisplaySafeRef) control.DisplaySafeRef {
	if queryRef = normalizeRef(queryRef); queryRef != "" {
		return makeRef("evidence:public_source_result", string(queryRef))
	}
	return DefaultEvidenceRef
}

func normalizeEvidence(values []Evidence, fallbackObservedAt string) []Evidence {
	out := make([]Evidence, 0, len(values))
	for _, value := range values {
		item := value
		item.SourceRef, item.QueryRef, item.EvidenceRef = normalizeRef(item.SourceRef), normalizeRef(item.QueryRef), normalizeRef(item.EvidenceRef)
		item.Kind = first(controlToken(item.Kind), "public_source_result")
		item.Strength = control.NormalizeEvidenceStrength(string(item.Strength))
		if item.Strength == control.EvidenceMissing {
			item.Strength = control.EvidenceAdequate
		}
		item.ObservedAt = first(strings.TrimSpace(item.ObservedAt), fallbackObservedAt)
		item.Confidence = first(controlToken(item.Confidence), "confidence:medium")
		item.ConfidenceRef, item.DisplaySafeRef = normalizeRef(item.ConfidenceRef), normalizeRef(item.DisplaySafeRef)
		if item.SourceRef == "" || item.EvidenceRef == "" {
			continue
		}
		if item.DisplaySafeRef == "" {
			item.DisplaySafeRef = makeRef("result:public_source", string(item.EvidenceRef))
		}
		out = append(out, item)
	}
	return out
}

func normalizeSummaries(values []DisplaySummary, fallbackObservedAt string) []DisplaySummary {
	out := make([]DisplaySummary, 0, len(values))
	for index, value := range values {
		item := normalizeSummary(value, index+1, fallbackObservedAt, "", "", "")
		if item.SummaryRef != "" {
			out = append(out, item)
		}
	}
	return out
}

func normalizeSummary(value DisplaySummary, rank int, observedAt string, queryRef, evidenceRef, sourceRef control.DisplaySafeRef) DisplaySummary {
	item := value
	item.SourceRef = normalizeRef(firstRef(item.SourceRef, sourceRef))
	item.QueryRef = normalizeRef(firstRef(item.QueryRef, queryRef))
	item.EvidenceRef = normalizeRef(firstRef(item.EvidenceRef, evidenceRef))
	item.AttestationRef, item.RedactionRef, item.DisplayPolicyRef = normalizeRef(item.AttestationRef), normalizeRef(item.RedactionRef), normalizeRef(item.DisplayPolicyRef)
	item.DisplaySafeAttested = item.DisplaySafeAttested || (item.AttestationRef != "" && item.RedactionRef != "")
	if item.DisplaySafeAttested && (item.AttestationRef == "" || item.RedactionRef == "") {
		item.DisplaySafeAttested = false
	}
	if item.EvidenceRef == "" && item.QueryRef != "" {
		item.EvidenceRef = QueryResultEvidenceRef(item.QueryRef)
	}
	item.Title, item.Summary, item.Snippet = displayText(item.Title, 160), displayText(item.Summary, 640), displayText(item.Snippet, 320)
	item.Published = displayText(item.Published, 80)
	item.ObservedAt = first(strings.TrimSpace(item.ObservedAt), observedAt)
	item.Confidence = first(controlToken(item.Confidence), confidence(rank))
	item.Boundaries = control.AppendBoundaries(item.Boundaries, "public_source_display_summary", "display_safe_public_source_summary", "raw_url_title_description_not_required")
	if item.DisplaySafeAttested {
		item.Boundaries = control.AppendBoundaries(item.Boundaries, "host_attested_public_source_display_summary", "display_summary_redaction_ref_bound")
	} else {
		item.Boundaries = control.AppendBoundaries(item.Boundaries, "public_source_display_summary_unattested", "display_summary_requires_host_attestation_for_production")
	}
	item.SummaryRef = normalizeRef(item.SummaryRef)
	if item.SummaryRef == "" && (item.Title != "" || item.Summary != "" || item.Snippet != "") {
		item.SummaryRef = makeRef("summary:public_source", strings.Join([]string{string(item.SourceRef), item.Title, item.Summary, item.Snippet}, ":"))
	}
	if item.SourceRef == "" && item.SummaryRef != "" {
		item.SourceRef = makeRef("source:public_source_summary", string(item.SummaryRef))
	}
	if item.QueryRef == "" && item.SummaryRef != "" {
		item.QueryRef = makeRef("query:public_source_summary", string(item.SummaryRef))
	}
	if item.EvidenceRef == "" && item.SummaryRef != "" {
		item.EvidenceRef = makeRef("evidence:public_source_display_summary", string(item.SummaryRef))
	}
	if item.SummaryRef == "" || (item.Title == "" && item.Summary == "" && item.Snippet == "") {
		return DisplaySummary{}
	}
	return item
}

func summaryStrength(item DisplaySummary) control.EvidenceStrength {
	if item.DisplaySafeAttested && item.AttestationRef != "" && item.RedactionRef != "" {
		return control.EvidenceAdequate
	}
	return control.EvidenceWeak
}

func displayText(value string, limit int) string {
	text := strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
	lower := strings.ToLower(text)
	if text == "" || strings.Contains(lower, "://") || strings.Contains(lower, "file:") || strings.Contains(text, "/Users/") || strings.Contains(text, "\\") {
		return ""
	}
	runes := []rune(text)
	if limit > 0 && len(runes) > limit {
		return string(runes[:limit])
	}
	return text
}

func resultRef(rawURL, siteName string, rank int) control.DisplaySafeRef {
	host := strings.TrimSpace(siteName)
	if host == "" {
		if parsed, err := url.Parse(strings.TrimSpace(rawURL)); err == nil && parsed != nil {
			host = parsed.Hostname()
		}
	}
	host = controlToken(strings.ReplaceAll(host, ".", "_"))
	if host == "" {
		host = "unknown_source"
	}
	key := first(rawURL, host)
	return makeRef("source:public_web:"+host, shortHash(key)+":"+rankToken(rank))
}

func makeRef(prefix, raw string) control.DisplaySafeRef {
	prefix = strings.Trim(strings.TrimSpace(prefix), ":")
	token := controlToken(raw)
	if token == "" || len(token) > 48 {
		token = shortHash(raw)
	}
	ref, _ := control.NormalizeDisplaySafeRef(prefix + ":" + token)
	return ref
}

func shortHash(raw string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(raw)))
	return hex.EncodeToString(sum[:])[:12]
}
func rankToken(rank int) string {
	if rank <= 0 {
		return "rank0"
	}
	return "rank" + strconv.Itoa(rank)
}
func confidence(rank int) string {
	if rank <= 1 {
		return "confidence:high"
	}
	if rank <= 3 {
		return "confidence:medium"
	}
	return "confidence:low"
}

func controlToken(value string) string {
	token := strings.NewReplacer(" ", "_", "/", "_", "\\", "_", ".", "_", "-", "_").Replace(strings.ToLower(strings.TrimSpace(value)))
	var b strings.Builder
	for _, r := range token {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' || r == ':' {
			b.WriteRune(r)
		}
	}
	return strings.Trim(b.String(), "_:")
}

func controlTokens(values []string) []string {
	out, seen := []string{}, map[string]bool{}
	for _, value := range values {
		token := controlToken(value)
		if token != "" && !seen[token] {
			seen[token] = true
			out = append(out, token)
		}
	}
	return out
}

func normalizeRef(value control.DisplaySafeRef) control.DisplaySafeRef {
	ref, _ := control.NormalizeDisplaySafeRef(string(value))
	return ref
}
func refs(values []control.DisplaySafeRef) []control.DisplaySafeRef {
	out := []control.DisplaySafeRef{}
	for _, value := range values {
		out = appendRef(out, value)
	}
	return out
}
func appendRef(values []control.DisplaySafeRef, value control.DisplaySafeRef) []control.DisplaySafeRef {
	value = normalizeRef(value)
	if value == "" {
		return values
	}
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}
func containsRef(values []control.DisplaySafeRef, want control.DisplaySafeRef) bool {
	want = normalizeRef(want)
	for _, value := range refs(values) {
		if value == want {
			return true
		}
	}
	return false
}
func firstRef(values ...control.DisplaySafeRef) control.DisplaySafeRef {
	for _, value := range values {
		if value = normalizeRef(value); value != "" {
			return value
		}
	}
	return ""
}
func first(values ...string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}
