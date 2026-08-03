package publicsource

import (
	"context"
	"strings"
	"time"

	control "github.com/wsnacj/agentx-go/runtime/controlcontract"
)

type SearchRequest struct {
	QueryRef   control.DisplaySafeRef   `json:"query_ref,omitempty"`
	QueryRefs  []control.DisplaySafeRef `json:"query_refs,omitempty"`
	SourceRefs []control.DisplaySafeRef `json:"source_refs,omitempty"`
	PolicyRefs []control.DisplaySafeRef `json:"policy_refs,omitempty"`
	ObservedAt string                   `json:"observed_at,omitempty"`
	Boundaries []control.Boundary       `json:"boundaries,omitempty"`
}

type SearchExecution struct {
	Payload          SearchPayload      `json:"payload"`
	DisplaySummaries []DisplaySummary   `json:"display_summaries,omitempty"`
	Boundaries       []control.Boundary `json:"boundaries,omitempty"`
}

type SearchExecutor interface {
	SearchPublicSource(context.Context, SearchRequest) (SearchExecution, error)
}
type SearchExecutorFunc func(context.Context, SearchRequest) (SearchExecution, error)

func (fn SearchExecutorFunc) SearchPublicSource(ctx context.Context, request SearchRequest) (SearchExecution, error) {
	return fn(ctx, request)
}

type SearchCollector struct {
	Executor     SearchExecutor
	QueryRef     control.DisplaySafeRef
	SourcePolicy SourcePolicy
	Now          func() time.Time
	Strength     control.EvidenceStrength
	Boundaries   []control.Boundary
}

func (collector SearchCollector) CollectPublicSourceEvidence(ctx context.Context, request Request) (Report, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	now := collector.Now
	if now == nil {
		now = time.Now
	}
	observedAt := first(strings.TrimSpace(request.ObservedAt), now().UTC().Format(time.RFC3339))
	queryRef := firstRef(collector.QueryRef, firstRef(request.QueryRefs...))
	base := Report{Status: control.VerificationBlocked, FailureClass: control.FailureTargetUnavailable, FailureReason: "public_source_search_unavailable", ObservedAt: observedAt, QueryEvidenceRefs: refs(append(request.QueryRefs, queryRef)), SourceRefs: refs(request.SourceRefs), UnavailableReasons: []string{"public_source_search_unavailable"}, Boundaries: control.AppendBoundaries(request.Boundaries, append(collector.Boundaries, "runtime_public_source_search_collector", "host_owned_public_source_search_executor", "public_source_search_read_only")...)}
	if collector.Executor == nil {
		base.FailureClass, base.FailureReason, base.UnavailableReasons = control.FailureHostAdapterMissing, "public_source_search_executor_missing", []string{"public_source_search_executor_missing"}
		return base.Normalize(), nil
	}
	if queryRef == "" {
		base.FailureClass, base.FailureReason, base.UnavailableReasons = control.FailureConfigMissing, "public_source_query_ref_missing", []string{"public_source_query_ref_missing"}
		return base.Normalize(), nil
	}
	result, err := collector.Executor.SearchPublicSource(ctx, SearchRequest{QueryRef: queryRef, QueryRefs: refs(append(request.QueryRefs, queryRef)), SourceRefs: refs(request.SourceRefs), PolicyRefs: refs(request.PolicyRefs), ObservedAt: observedAt, Boundaries: base.Boundaries})
	if err != nil {
		base.FailureClass, base.FailureReason, base.UnavailableReasons = control.FailureExternalDependencyUnavailable, "public_source_search_executor_error", []string{"public_source_search_executor_error"}
		base.Boundaries = control.AppendBoundaries(base.Boundaries, "public_source_search_executor_error")
		return base.Normalize(), nil
	}
	report := BuildReportFromSearch(SearchReportInput{Payload: result.Payload, DisplaySummaries: result.DisplaySummaries, QueryRef: queryRef, SourcePolicy: collector.SourcePolicy, ObservedAt: observedAt, Strength: collector.Strength})
	report.SourceRefs = refs(append(report.SourceRefs, request.SourceRefs...))
	report.Boundaries = control.AppendBoundaries(base.Boundaries, append(result.Boundaries, report.Boundaries...)...)
	return report.Normalize(), nil
}

// Document is the portable display and evidence subset produced by a Host
// fetcher. It deliberately excludes cookies, headers, provider diagnostics,
// local paths, and fetch implementation details.
type Document struct {
	PageRef    control.DisplaySafeRef `json:"page_ref,omitempty"`
	RequestURL string                 `json:"-"`
	FinalURL   string                 `json:"-"`
	Status     int                    `json:"status,omitempty"`
	SiteName   string                 `json:"site_name,omitempty"`
	Title      string                 `json:"title,omitempty"`
	Excerpt    string                 `json:"excerpt,omitempty"`
	Text       string                 `json:"-"`
	Published  string                 `json:"published,omitempty"`
	TextReady  bool                   `json:"text_ready"`
}

type DocumentRequest struct {
	URL        string                   `json:"-"`
	URLRef     control.DisplaySafeRef   `json:"url_ref,omitempty"`
	QueryRef   control.DisplaySafeRef   `json:"query_ref,omitempty"`
	QueryRefs  []control.DisplaySafeRef `json:"query_refs,omitempty"`
	SourceRefs []control.DisplaySafeRef `json:"source_refs,omitempty"`
	PolicyRefs []control.DisplaySafeRef `json:"policy_refs,omitempty"`
	ObservedAt string                   `json:"observed_at,omitempty"`
	Boundaries []control.Boundary       `json:"boundaries,omitempty"`
}

type DocumentExecution struct {
	Document           Document             `json:"document"`
	DisplaySummaries   []DisplaySummary     `json:"display_summaries,omitempty"`
	UnavailableReasons []string             `json:"unavailable_reasons,omitempty"`
	FailureClass       control.FailureClass `json:"failure_class,omitempty"`
	FailureReason      string               `json:"failure_reason,omitempty"`
	Boundaries         []control.Boundary   `json:"boundaries,omitempty"`
}

type DocumentFetcher interface {
	FetchPublicSourceDocument(context.Context, DocumentRequest) (DocumentExecution, error)
}
type DocumentFetcherFunc func(context.Context, DocumentRequest) (DocumentExecution, error)

func (fn DocumentFetcherFunc) FetchPublicSourceDocument(ctx context.Context, request DocumentRequest) (DocumentExecution, error) {
	return fn(ctx, request)
}

type DocumentCollector struct {
	Fetcher          DocumentFetcher
	URL              string
	URLRef           control.DisplaySafeRef
	QueryRef         control.DisplaySafeRef
	SourceRef        control.DisplaySafeRef
	FetcherRef       control.DisplaySafeRef
	AttestationRef   control.DisplaySafeRef
	RedactionRef     control.DisplaySafeRef
	DisplayPolicyRef control.DisplaySafeRef
	SourcePolicy     SourcePolicy
	Now              func() time.Time
	Strength         control.EvidenceStrength
	Boundaries       []control.Boundary
}

func (collector DocumentCollector) CollectPublicSourceEvidence(ctx context.Context, request Request) (Report, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	now := collector.Now
	if now == nil {
		now = time.Now
	}
	observedAt := first(strings.TrimSpace(request.ObservedAt), now().UTC().Format(time.RFC3339))
	rawURL := strings.TrimSpace(collector.URL)
	urlRef := normalizeRef(collector.URLRef)
	if urlRef == "" && rawURL != "" {
		urlRef = makeRef("url:public_source_document", rawURL)
	}
	queryRef := firstRef(collector.QueryRef, firstRef(request.QueryRefs...))
	sourceRef := normalizeRef(collector.SourceRef)
	if sourceRef == "" && rawURL != "" {
		sourceRef = makeRef("source:public_source_document", rawURL)
	}
	fetcherRef := firstRef(collector.FetcherRef, "fetcher:host_public_source_document")
	report := Report{Status: control.VerificationBlocked, FailureClass: control.FailureTargetUnavailable, FailureReason: "public_source_document_unavailable", ObservedAt: observedAt, QueryEvidenceRefs: refs(append(request.QueryRefs, queryRef)), SourceRefs: refs([]control.DisplaySafeRef{sourceRef, fetcherRef}), UnavailableReasons: []string{"public_source_document_unavailable"}, Boundaries: control.AppendBoundaries(request.Boundaries, append(collector.Boundaries, "runtime_public_source_document_collector", "host_owned_public_source_document_fetcher", "url_supplied_by_host", "raw_document_payload_not_reported", "public_source_document_read_only")...)}
	if collector.Fetcher == nil {
		report.FailureClass, report.FailureReason, report.UnavailableReasons = control.FailureHostAdapterMissing, "public_source_document_fetcher_missing", []string{"public_source_document_fetcher_missing"}
		return report.Normalize(), nil
	}
	if rawURL == "" {
		report.FailureClass, report.FailureReason, report.UnavailableReasons = control.FailureConfigMissing, "public_source_document_url_missing", []string{"public_source_document_url_missing"}
		return report.Normalize(), nil
	}
	if decision := collector.SourcePolicy.CheckURL(rawURL); decision.Applied {
		report.Boundaries = control.AppendBoundaries(report.Boundaries, decision.Boundaries...)
		if !decision.Allowed {
			report.FailureClass, report.FailureReason, report.UnavailableReasons = decision.FailureClass, decision.FailureReason, []string{decision.FailureReason}
			return report.Normalize(), nil
		}
	}
	result, err := collector.Fetcher.FetchPublicSourceDocument(ctx, DocumentRequest{URL: rawURL, URLRef: urlRef, QueryRef: queryRef, QueryRefs: refs(append(request.QueryRefs, queryRef)), SourceRefs: refs([]control.DisplaySafeRef{sourceRef, fetcherRef}), PolicyRefs: refs(request.PolicyRefs), ObservedAt: observedAt, Boundaries: report.Boundaries})
	if err != nil {
		report.FailureClass, report.FailureReason, report.UnavailableReasons = control.FailureExternalDependencyUnavailable, "public_source_document_fetcher_error", []string{"public_source_document_fetcher_error"}
		report.Boundaries = control.AppendBoundaries(report.Boundaries, "public_source_document_fetcher_error")
		return report.Normalize(), nil
	}
	if result.Document.Status != 0 && (result.Document.Status < 200 || result.Document.Status >= 400) {
		return collector.blockedDocument(report, result, control.FailureTargetUnavailable, "public_source_document_http_not_ok"), nil
	}
	if !result.Document.TextReady || strings.TrimSpace(first(result.Document.Text, result.Document.Excerpt, result.Document.Title)) == "" {
		return collector.blockedDocument(report, result, control.FailureEvidenceMissing, "public_source_document_text_unavailable"), nil
	}
	finalURL := first(result.Document.FinalURL, result.Document.RequestURL, rawURL)
	summaries := result.DisplaySummaries
	if len(summaries) == 0 {
		summaries = collector.summaries(result.Document)
	}
	sourceReport := BuildReportFromSearch(SearchReportInput{Payload: SearchPayload{Query: string(queryRef), Provider: "host_public_source_document", Count: 1, Results: []SearchResult{{Title: result.Document.Title, URL: finalURL, Description: result.Document.Excerpt, Published: result.Document.Published, SiteName: result.Document.SiteName}}}, DisplaySummaries: summaries, QueryRef: queryRef, SourcePolicy: collector.SourcePolicy, ObservedAt: observedAt, Strength: collector.Strength})
	sourceReport.QueryEvidenceRefs = refs(append(sourceReport.QueryEvidenceRefs, request.QueryRefs...))
	sourceReport.SourceRefs = refs(append(sourceReport.SourceRefs, sourceRef, fetcherRef))
	sourceReport.Boundaries = control.AppendBoundaries(report.Boundaries, append(result.Boundaries, append(sourceReport.Boundaries, "public_source_document_fetcher_invoked", "public_source_document_payload_loaded")...)...)
	sourceReport.UnavailableReasons = controlTokens(append(sourceReport.UnavailableReasons, result.UnavailableReasons...))
	return sourceReport.Normalize(), nil
}

func (collector DocumentCollector) blockedDocument(report Report, result DocumentExecution, failure control.FailureClass, reason string) Report {
	if result.FailureClass != "" && result.FailureClass != control.FailureNone {
		failure = result.FailureClass
	}
	if result.FailureReason != "" {
		reason = result.FailureReason
	}
	report.FailureClass, report.FailureReason = failure, reason
	report.UnavailableReasons = controlTokens(append([]string{reason}, result.UnavailableReasons...))
	report.Boundaries = control.AppendBoundaries(report.Boundaries, result.Boundaries...)
	return report.Normalize()
}

func (collector DocumentCollector) summaries(document Document) []DisplaySummary {
	title, summary := strings.TrimSpace(document.Title), strings.TrimSpace(document.Excerpt)
	if summary == "" {
		summary = strings.TrimSpace(document.Text)
		runes := []rune(summary)
		if len(runes) > 240 {
			summary = string(runes[:240])
		}
	}
	if title == "" && summary == "" {
		return nil
	}
	attested := normalizeRef(collector.AttestationRef) != "" && normalizeRef(collector.RedactionRef) != ""
	boundaries := []control.Boundary{"host_generated_public_source_document_display_summary"}
	if attested {
		boundaries = append(boundaries, "host_attested_public_source_document_display_summary")
	} else {
		boundaries = append(boundaries, "host_unattested_public_source_document_display_summary", "display_summary_requires_host_attestation_for_production")
	}
	return []DisplaySummary{{Title: title, Summary: summary, DisplaySafeAttested: attested, AttestationRef: normalizeRef(collector.AttestationRef), RedactionRef: normalizeRef(collector.RedactionRef), DisplayPolicyRef: normalizeRef(collector.DisplayPolicyRef), Boundaries: boundaries}}
}

var _ Collector = SearchCollector{}
var _ Collector = DocumentCollector{}
