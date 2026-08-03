package metrics

import (
	"net/url"
	"strings"
)

// LatestMetricsEvaluationInput is the pack-level evidence shape consumed by
// adapters and live evaluators. It stays source-neutral: project/plugin
// adapters own site-specific extraction and only project their evidence here.
type LatestMetricsEvaluationInput struct {
	UserMessage            string
	ExpectedEntityName     string
	ExpectedStockCode      string
	EvidenceEntityName     string
	EvidenceStockCode      string
	SourceURL              string
	ReportPeriod           string
	RequestedFields        []string
	FieldEvidence          map[string]ReportDocumentMetricFieldEvidence
	MissingRequestedFields []string
	ReviewRequiredFields   []string
	RequestedFieldsReady   bool
	GuardStatus            string
	StopAfterGuardPassed   bool
	TrendRequested         bool
	TrendSeriesReady       bool
	TrendSeriesPeriodCount int
}

// LatestMetricsEvaluation is the stable pack-level evaluator contract for
// public financial report metrics. It is intentionally generic; it does not
// know about Eastmoney, official IR pages, or PDF-specific playbooks.
type LatestMetricsEvaluation struct {
	Passed                 bool     `json:"passed"`
	SubjectCorrect         bool     `json:"subject_correct"`
	PeriodLatest           bool     `json:"period_latest"`
	RequestedFieldsReady   bool     `json:"requested_fields_ready"`
	GrowthFieldsConsistent bool     `json:"growth_fields_consistent"`
	SourceAccepted         bool     `json:"source_accepted"`
	FieldSourcesAccepted   bool     `json:"field_sources_accepted"`
	StopAfterGuardPassed   bool     `json:"stop_after_guard_passed"`
	TrendRequested         bool     `json:"trend_requested,omitempty"`
	TrendSeriesReady       bool     `json:"trend_series_ready"`
	TrendSeriesPeriodCount int      `json:"trend_series_period_count"`
	MissingRequestedFields []string `json:"missing_requested_fields"`
	ReviewRequiredFields   []string `json:"review_required_fields"`
	FailureReason          string   `json:"failure_reason"`
	SourceURL              string   `json:"source_url"`
	ReportPeriod           string   `json:"report_period"`
}

func EvaluateLatestMetricsEvidence(input LatestMetricsEvaluationInput) LatestMetricsEvaluation {
	missing := normalizeStringList(input.MissingRequestedFields)
	review := normalizeStringList(input.ReviewRequiredFields)
	requested := normalizeStringList(input.RequestedFields)
	guardPassed := strings.EqualFold(strings.TrimSpace(input.GuardStatus), "passed")
	out := LatestMetricsEvaluation{
		SubjectCorrect:         latestMetricsSubjectCorrect(input),
		PeriodLatest:           latestMetricsPeriodAccepted(input.ReportPeriod),
		RequestedFieldsReady:   input.RequestedFieldsReady && len(missing) == 0 && len(review) == 0,
		GrowthFieldsConsistent: latestMetricsGrowthFieldsConsistent(requested, missing, review),
		SourceAccepted:         latestMetricsSourceAccepted(input.SourceURL),
		FieldSourcesAccepted:   latestMetricsFieldSourcesAccepted(input.SourceURL, requested, input.FieldEvidence),
		StopAfterGuardPassed:   guardPassed && input.StopAfterGuardPassed,
		TrendRequested:         input.TrendRequested,
		TrendSeriesReady:       !input.TrendRequested || (input.TrendSeriesReady && input.TrendSeriesPeriodCount >= 2),
		TrendSeriesPeriodCount: input.TrendSeriesPeriodCount,
		MissingRequestedFields: missing,
		ReviewRequiredFields:   review,
		SourceURL:              strings.TrimSpace(input.SourceURL),
		ReportPeriod:           strings.TrimSpace(input.ReportPeriod),
	}
	reasons := []string{}
	if !guardPassed {
		reasons = append(reasons, "guard_not_passed")
	}
	if !out.SubjectCorrect {
		reasons = append(reasons, "subject_mismatch")
	}
	if !out.PeriodLatest {
		reasons = append(reasons, "report_period_missing")
	}
	if !out.RequestedFieldsReady {
		if len(missing) > 0 {
			reasons = append(reasons, "requested_fields_missing")
		}
		if len(review) > 0 {
			reasons = append(reasons, "review_required_fields")
		}
		if len(missing) == 0 && len(review) == 0 {
			reasons = append(reasons, "requested_fields_not_ready")
		}
	}
	if !out.GrowthFieldsConsistent {
		reasons = append(reasons, "growth_fields_incomplete")
	}
	if !out.SourceAccepted {
		reasons = append(reasons, "source_unaccepted")
	}
	if !out.FieldSourcesAccepted {
		reasons = append(reasons, "field_sources_unaccepted")
	}
	if !out.StopAfterGuardPassed {
		reasons = append(reasons, "stop_after_guard_not_confirmed")
	}
	if input.TrendRequested && !out.TrendSeriesReady {
		reasons = append(reasons, "trend_series_incomplete")
	}
	out.Passed = len(reasons) == 0
	out.FailureReason = strings.Join(reasons, ",")
	return out
}

func latestMetricsSubjectCorrect(input LatestMetricsEvaluationInput) bool {
	expectedCode := latestMetricsNormalizeCode(input.ExpectedStockCode)
	evidenceCode := latestMetricsNormalizeCode(input.EvidenceStockCode)
	if expectedCode != "" {
		return evidenceCode == expectedCode || strings.Contains(evidenceCode, expectedCode)
	}
	expectedName := latestMetricsNormalizeEntity(input.ExpectedEntityName)
	evidenceName := latestMetricsNormalizeEntity(input.EvidenceEntityName)
	if expectedName != "" {
		return evidenceName != "" && (strings.Contains(evidenceName, expectedName) || strings.Contains(expectedName, evidenceName))
	}
	return evidenceName != "" || evidenceCode != ""
}

func latestMetricsNormalizeCode(value string) string {
	value = strings.ToUpper(strings.TrimSpace(value))
	if value == "" || strings.EqualFold(value, "unknown") {
		return ""
	}
	for _, suffix := range []string{".US", ".N", ".O", ".HK", ".SH", ".SZ", ".SS", ".BJ"} {
		value = strings.TrimSuffix(value, suffix)
	}
	replacer := strings.NewReplacer(".", "", "-", "", "_", "", " ", "")
	value = replacer.Replace(value)
	for _, prefix := range []string{"SH", "SZ", "BJ", "HK"} {
		value = strings.TrimPrefix(value, prefix)
	}
	return value
}

func latestMetricsNormalizeEntity(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" || value == "unknown" {
		return ""
	}
	replacer := strings.NewReplacer(" ", "", "\t", "", "\n", "", "有限公司", "", "股份", "", "公司", "")
	return replacer.Replace(value)
}

func latestMetricsPeriodAccepted(value string) bool {
	value = strings.TrimSpace(value)
	return value != "" && !strings.EqualFold(value, "unknown")
}

func latestMetricsSourceAccepted(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" || strings.EqualFold(value, "unknown") {
		return false
	}
	parsed, err := url.Parse(value)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return false
	}
	return !latestMetricsSearchResultSource(parsed)
}

// LatestMetricsFieldSourcesAccepted verifies that requested metric fields are
// bound to the same accepted final evidence URL. Parser-local labels such as
// table/docparse are accepted only when the enclosing final evidence URL is
// accepted; URL-looking field sources must match that final evidence URL.
func LatestMetricsFieldSourcesAccepted(sourceURL string, requestedFields []string, fieldEvidence map[string]ReportDocumentMetricFieldEvidence) bool {
	return latestMetricsFieldSourcesAccepted(sourceURL, normalizeStringList(requestedFields), fieldEvidence)
}

func latestMetricsFieldSourcesAccepted(sourceURL string, requestedFields []string, fieldEvidence map[string]ReportDocumentMetricFieldEvidence) bool {
	if len(requestedFields) == 0 || len(fieldEvidence) == 0 {
		return true
	}
	sourceURL = strings.TrimSpace(sourceURL)
	if !latestMetricsSourceAccepted(sourceURL) {
		return false
	}
	expected := latestMetricsCanonicalEvidenceURL(sourceURL)
	if expected == "" {
		return false
	}
	for _, field := range requestedFields {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}
		evidence, ok := fieldEvidence[field]
		if !ok {
			return false
		}
		if evidence.ReviewRequired || strings.TrimSpace(evidence.Value) == "" || strings.EqualFold(strings.TrimSpace(evidence.Value), "unknown") {
			return false
		}
		source := strings.TrimSpace(evidence.Source)
		if source == "" || strings.EqualFold(source, "unknown") {
			return false
		}
		if latestMetricsSourceAccepted(source) {
			if latestMetricsCanonicalEvidenceURL(source) != expected {
				return false
			}
			continue
		}
		if latestMetricsLooksLikeURL(source) {
			return false
		}
		if !latestMetricsLocalFieldSourceAccepted(evidence) {
			return false
		}
	}
	return true
}

func latestMetricsCanonicalEvidenceURL(value string) string {
	parsed, err := url.Parse(strings.TrimSpace(value))
	if err != nil || parsed.Host == "" {
		return ""
	}
	scheme := strings.ToLower(strings.TrimSpace(parsed.Scheme))
	host := strings.ToLower(strings.TrimSpace(parsed.Host))
	path := strings.TrimRight(parsed.EscapedPath(), "/")
	if path == "" {
		path = "/"
	}
	if parsed.RawQuery != "" {
		return scheme + "://" + host + path + "?" + parsed.RawQuery
	}
	return scheme + "://" + host + path
}

func latestMetricsLooksLikeURL(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	return strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://")
}

func latestMetricsLocalFieldSourceAccepted(evidence ReportDocumentMetricFieldEvidence) bool {
	source := strings.ToLower(strings.TrimSpace(evidence.Source))
	if source == "" || source == "unknown" {
		return false
	}
	if latestMetricsContainsAny(source, "table", "docparse", "pdf", "ocr", "xbrl", "filing", "annual_report", "structured_text", "regex") {
		return true
	}
	if strings.TrimSpace(evidence.Evidence) != "" || strings.TrimSpace(evidence.Chapter) != "" || len(evidence.PageRefs) > 0 || len(evidence.Candidates) > 0 {
		return true
	}
	return false
}

func latestMetricsContainsAny(text string, needles ...string) bool {
	text = strings.ToLower(text)
	for _, needle := range needles {
		if strings.Contains(text, strings.ToLower(needle)) {
			return true
		}
	}
	return false
}

func latestMetricsSearchResultSource(parsed *url.URL) bool {
	if parsed == nil {
		return false
	}
	host := strings.ToLower(strings.TrimSpace(parsed.Hostname()))
	path := strings.ToLower(strings.Trim(strings.TrimSpace(parsed.EscapedPath()), "/"))
	query := parsed.Query()
	switch host {
	case "google.com", "www.google.com":
		return path == "search" || path == "url" || query.Get("q") != ""
	case "bing.com", "www.bing.com":
		return path == "search" || path == "ck/a" || query.Get("q") != ""
	case "baidu.com", "www.baidu.com", "m.baidu.com":
		return path == "s" || path == "link" || query.Get("wd") != "" || query.Get("word") != ""
	case "sogou.com", "www.sogou.com", "m.sogou.com":
		return path == "web" || query.Get("query") != "" || query.Get("keyword") != ""
	case "so.com", "www.so.com":
		return path == "s" || query.Get("q") != ""
	case "duckduckgo.com", "www.duckduckgo.com":
		return (path == "" && query.Get("q") != "") || path == "html" || path == "lite"
	default:
		return false
	}
}

func latestMetricsGrowthFieldsConsistent(requested []string, missing []string, review []string) bool {
	for _, field := range []string{"revenue_growth", "net_profit_growth"} {
		if stringListContains(requested, field) && (stringListContains(missing, field) || stringListContains(review, field)) {
			return false
		}
	}
	return true
}

func normalizeStringList(values []string) []string {
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

func stringListContains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
