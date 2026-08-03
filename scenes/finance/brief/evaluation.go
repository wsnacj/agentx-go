package brief

import (
	"net/url"
	"strings"
)

type BriefKeyPoint struct {
	Category       string   `json:"category,omitempty"`
	Title          string   `json:"title,omitempty"`
	Text           string   `json:"text,omitempty"`
	Source         string   `json:"source,omitempty"`
	PageRefs       []int    `json:"page_refs,omitempty"`
	Confidence     float64  `json:"confidence,omitempty"`
	ReviewRequired bool     `json:"review_required,omitempty"`
	Warnings       []string `json:"warnings,omitempty"`
}

type BriefMetric struct {
	Name           string  `json:"name,omitempty"`
	Value          string  `json:"value,omitempty"`
	Unit           string  `json:"unit,omitempty"`
	Period         string  `json:"period,omitempty"`
	Source         string  `json:"source,omitempty"`
	PageRefs       []int   `json:"page_refs,omitempty"`
	Confidence     float64 `json:"confidence,omitempty"`
	ReviewRequired bool    `json:"review_required,omitempty"`
}

type BriefEvidence struct {
	CompanyName     string          `json:"company_name,omitempty"`
	StockCode       string          `json:"stock_code,omitempty"`
	ReportPeriod    string          `json:"report_period,omitempty"`
	SourceURL       string          `json:"source_url,omitempty"`
	ReportPath      string          `json:"report_path,omitempty"`
	ParserID        string          `json:"parser_id,omitempty"`
	Brief           string          `json:"brief,omitempty"`
	KeyPoints       []BriefKeyPoint `json:"key_points,omitempty"`
	Metrics         []BriefMetric   `json:"metrics,omitempty"`
	SourceChapters  []string        `json:"source_chapters,omitempty"`
	ReviewRequired  bool            `json:"review_required,omitempty"`
	ReviewReasons   []string        `json:"review_reasons,omitempty"`
	ExtractWarnings []string        `json:"extract_warnings,omitempty"`
}

type BriefEvaluationInput struct {
	ExpectedEntityName   string
	ExpectedStockCode    string
	Evidence             BriefEvidence
	GuardStatus          string
	StopAfterGuardPassed bool
}

type BriefEvaluation struct {
	Passed               bool     `json:"passed"`
	SubjectCorrect       bool     `json:"subject_correct"`
	PeriodLatest         bool     `json:"period_latest"`
	SourceAccepted       bool     `json:"source_accepted"`
	BriefReady           bool     `json:"brief_ready"`
	KeyPointsReady       bool     `json:"key_points_ready"`
	FinancialsReady      bool     `json:"financials_ready"`
	RiskOrOutlookReady   bool     `json:"risk_or_outlook_ready"`
	StopAfterGuardPassed bool     `json:"stop_after_guard_passed"`
	FailureReason        string   `json:"failure_reason"`
	ReviewReasons        []string `json:"review_reasons"`
	SourceURL            string   `json:"source_url"`
	ReportPeriod         string   `json:"report_period"`
}

func EvaluateBriefEvidence(input BriefEvaluationInput) BriefEvaluation {
	evidence := input.Evidence
	reviewReasons := normalizeBriefStringList(evidence.ReviewReasons)
	if evidence.ReviewRequired && !briefStringListContains(reviewReasons, "review_required") {
		reviewReasons = append(reviewReasons, "review_required")
	}
	guardPassed := strings.EqualFold(strings.TrimSpace(input.GuardStatus), "passed")
	out := BriefEvaluation{
		SubjectCorrect:       briefSubjectCorrect(input.ExpectedEntityName, input.ExpectedStockCode, evidence.CompanyName, evidence.StockCode),
		PeriodLatest:         strings.TrimSpace(evidence.ReportPeriod) != "" && !strings.EqualFold(strings.TrimSpace(evidence.ReportPeriod), "unknown"),
		SourceAccepted:       briefSourceAccepted(evidence.SourceURL),
		BriefReady:           strings.TrimSpace(evidence.Brief) != "",
		KeyPointsReady:       len(evidence.KeyPoints) >= 3,
		FinancialsReady:      briefHasCategory(evidence.KeyPoints, "financial") || len(evidence.Metrics) >= 2,
		RiskOrOutlookReady:   briefHasCategory(evidence.KeyPoints, "risk") || briefHasCategory(evidence.KeyPoints, "outlook"),
		StopAfterGuardPassed: guardPassed && input.StopAfterGuardPassed,
		ReviewReasons:        reviewReasons,
		SourceURL:            strings.TrimSpace(evidence.SourceURL),
		ReportPeriod:         strings.TrimSpace(evidence.ReportPeriod),
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
	if !out.SourceAccepted {
		reasons = append(reasons, "source_unaccepted")
	}
	if !out.BriefReady {
		reasons = append(reasons, "brief_missing")
	}
	if !out.KeyPointsReady {
		reasons = append(reasons, "key_points_incomplete")
	}
	if !out.FinancialsReady {
		reasons = append(reasons, "financials_missing")
	}
	if !out.RiskOrOutlookReady {
		reasons = append(reasons, "risk_or_outlook_missing")
	}
	if len(reviewReasons) > 0 {
		reasons = append(reasons, "review_required")
	}
	if !out.StopAfterGuardPassed {
		reasons = append(reasons, "stop_after_guard_not_confirmed")
	}
	out.Passed = len(reasons) == 0
	out.FailureReason = strings.Join(reasons, ",")
	return out
}

func briefSubjectCorrect(expectedName, expectedCode, evidenceName, evidenceCode string) bool {
	expectedCode = normalizeBriefCode(expectedCode)
	evidenceCode = normalizeBriefCode(evidenceCode)
	if expectedCode != "" {
		return evidenceCode == expectedCode || strings.Contains(evidenceCode, expectedCode)
	}
	expectedName = normalizeBriefEntity(expectedName)
	evidenceName = normalizeBriefEntity(evidenceName)
	if expectedName != "" {
		return evidenceName != "" && (strings.Contains(evidenceName, expectedName) || strings.Contains(expectedName, evidenceName))
	}
	return evidenceName != "" || evidenceCode != ""
}

func normalizeBriefCode(value string) string {
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

func normalizeBriefEntity(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" || value == "unknown" {
		return ""
	}
	replacer := strings.NewReplacer(" ", "", "\t", "", "\n", "", "有限公司", "", "股份", "", "公司", "")
	return replacer.Replace(value)
}

func briefSourceAccepted(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" || strings.EqualFold(value, "unknown") {
		return false
	}
	parsed, err := url.Parse(value)
	return err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != ""
}

func briefHasCategory(points []BriefKeyPoint, category string) bool {
	category = strings.ToLower(strings.TrimSpace(category))
	for _, point := range points {
		if strings.ToLower(strings.TrimSpace(point.Category)) == category && strings.TrimSpace(point.Text) != "" {
			return true
		}
	}
	return false
}

func normalizeBriefStringList(values []string) []string {
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

func briefStringListContains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
