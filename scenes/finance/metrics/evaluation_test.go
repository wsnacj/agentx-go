package metrics

import "testing"

func TestEvaluateLatestMetricsEvidencePassesCompleteGuardedEvidence(t *testing.T) {
	result := EvaluateLatestMetricsEvidence(LatestMetricsEvaluationInput{
		ExpectedEntityName:   "贵州茅台",
		ExpectedStockCode:    "600519",
		EvidenceEntityName:   "贵州茅台",
		EvidenceStockCode:    "600519.SH",
		SourceURL:            "https://data.eastmoney.com/bbsj/600519.html?code=SH600519",
		ReportPeriod:         "2025-12-31",
		RequestedFields:      []string{"revenue", "revenue_growth", "net_profit", "net_profit_growth"},
		RequestedFieldsReady: true,
		GuardStatus:          "passed",
		StopAfterGuardPassed: true,
	})
	if !result.Passed ||
		!result.SubjectCorrect ||
		!result.PeriodLatest ||
		!result.RequestedFieldsReady ||
		!result.GrowthFieldsConsistent ||
		!result.SourceAccepted ||
		!result.StopAfterGuardPassed ||
		result.FailureReason != "" {
		t.Fatalf("expected complete evidence to pass, got %#v", result)
	}
}

func TestEvaluateLatestMetricsEvidenceNormalizesMarketCodeSuffixes(t *testing.T) {
	result := EvaluateLatestMetricsEvidence(LatestMetricsEvaluationInput{
		ExpectedEntityName:   "阿里巴巴",
		ExpectedStockCode:    "BABA.N",
		EvidenceEntityName:   "Alibaba Group Holding Ltd",
		EvidenceStockCode:    "BABA",
		SourceURL:            "https://www.sec.gov/Archives/edgar/data/1577552/example.htm",
		ReportPeriod:         "2026-03-31 20-F",
		RequestedFields:      []string{"revenue", "net_profit"},
		RequestedFieldsReady: true,
		GuardStatus:          "passed",
		StopAfterGuardPassed: true,
	})
	if !result.Passed || !result.SubjectCorrect {
		t.Fatalf("expected BABA.N request to match BABA evidence, got %#v", result)
	}
}

func TestEvaluateLatestMetricsEvidenceFailsMismatchAndReviewFields(t *testing.T) {
	result := EvaluateLatestMetricsEvidence(LatestMetricsEvaluationInput{
		ExpectedEntityName:     "贵州茅台",
		EvidenceEntityName:     "浦发银行",
		SourceURL:              "东方财富网",
		ReportPeriod:           "unknown",
		RequestedFields:        []string{"revenue", "revenue_growth", "net_profit"},
		MissingRequestedFields: []string{"net_profit"},
		ReviewRequiredFields:   []string{"revenue_growth"},
		RequestedFieldsReady:   false,
		GuardStatus:            "needs_review",
		StopAfterGuardPassed:   false,
	})
	if result.Passed ||
		result.SubjectCorrect ||
		result.PeriodLatest ||
		result.RequestedFieldsReady ||
		result.GrowthFieldsConsistent ||
		result.SourceAccepted ||
		result.StopAfterGuardPassed {
		t.Fatalf("expected incomplete evidence to fail, got %#v", result)
	}
	for _, reason := range []string{
		"guard_not_passed",
		"subject_mismatch",
		"report_period_missing",
		"requested_fields_missing",
		"review_required_fields",
		"growth_fields_incomplete",
		"source_unaccepted",
		"stop_after_guard_not_confirmed",
	} {
		if !containsSubstring(result.FailureReason, reason) {
			t.Fatalf("expected failure reason %q in %#v", reason, result)
		}
	}
}

func TestEvaluateLatestMetricsEvidenceRejectsSearchResultSource(t *testing.T) {
	result := EvaluateLatestMetricsEvidence(LatestMetricsEvaluationInput{
		ExpectedEntityName:   "贵州茅台",
		ExpectedStockCode:    "600519",
		EvidenceEntityName:   "贵州茅台",
		EvidenceStockCode:    "600519.SH",
		SourceURL:            "https://www.baidu.com/s?wd=%E8%B4%B5%E5%B7%9E%E8%8C%85%E5%8F%B0+%E5%B9%B4%E6%8A%A5",
		ReportPeriod:         "2025-12-31",
		RequestedFields:      []string{"revenue", "net_profit"},
		RequestedFieldsReady: true,
		GuardStatus:          "passed",
		StopAfterGuardPassed: true,
	})
	if result.Passed ||
		result.SourceAccepted ||
		!containsSubstring(result.FailureReason, "source_unaccepted") {
		t.Fatalf("expected search result source to fail source acceptance, got %#v", result)
	}
}

func TestEvaluateLatestMetricsEvidenceRequiresFieldSourcesBoundToFinalEvidence(t *testing.T) {
	result := EvaluateLatestMetricsEvidence(LatestMetricsEvaluationInput{
		ExpectedEntityName: "贵州茅台",
		ExpectedStockCode:  "600519",
		EvidenceEntityName: "贵州茅台",
		EvidenceStockCode:  "600519.SH",
		SourceURL:          "https://data.eastmoney.com/bbsj/600519.html?code=SH600519",
		ReportPeriod:       "2025-12-31",
		RequestedFields:    []string{"revenue", "net_profit"},
		FieldEvidence: map[string]ReportDocumentMetricFieldEvidence{
			"revenue": {
				Field:  "revenue",
				Value:  "1720.54亿元",
				Source: "https://data.eastmoney.com/bbsj/600519.html?code=SH600519",
				Period: "2025-12-31",
			},
			"net_profit": {
				Field:  "net_profit",
				Value:  "823.20亿元",
				Source: "https://finance.example.com/repost/maotai.html",
				Period: "2025-12-31",
			},
		},
		RequestedFieldsReady: true,
		GuardStatus:          "passed",
		StopAfterGuardPassed: true,
	})
	if result.Passed ||
		result.FieldSourcesAccepted ||
		!containsSubstring(result.FailureReason, "field_sources_unaccepted") {
		t.Fatalf("expected mismatched field source to fail field source acceptance, got %#v", result)
	}

	result = EvaluateLatestMetricsEvidence(LatestMetricsEvaluationInput{
		ExpectedEntityName: "贵州茅台",
		ExpectedStockCode:  "600519",
		EvidenceEntityName: "贵州茅台",
		EvidenceStockCode:  "600519.SH",
		SourceURL:          "https://data.eastmoney.com/bbsj/600519.html?code=SH600519",
		ReportPeriod:       "2025-12-31",
		RequestedFields:    []string{"revenue", "net_profit"},
		FieldEvidence: map[string]ReportDocumentMetricFieldEvidence{
			"revenue": {
				Field:  "revenue",
				Value:  "1720.54亿元",
				Source: "https://data.eastmoney.com/bbsj/600519.html?code=SH600519",
				Period: "2025-12-31",
			},
			"net_profit": {
				Field:    "net_profit",
				Value:    "823.20亿元",
				Source:   "table",
				Period:   "2025-12-31",
				PageRefs: []int{88},
			},
		},
		RequestedFieldsReady: true,
		GuardStatus:          "passed",
		StopAfterGuardPassed: true,
	})
	if !result.Passed || !result.FieldSourcesAccepted {
		t.Fatalf("expected same-source URL plus parser-local field evidence to pass, got %#v", result)
	}
}

func TestEvaluateLatestMetricsEvidenceRequiresTrendSeriesForTrendRequests(t *testing.T) {
	result := EvaluateLatestMetricsEvidence(LatestMetricsEvaluationInput{
		ExpectedEntityName:     "中国银行",
		EvidenceEntityName:     "中国银行",
		SourceURL:              "https://data.eastmoney.com/bbsj/601988.html?code=SH601988",
		ReportPeriod:           "2025-12-31",
		RequestedFields:        []string{"revenue", "revenue_growth", "net_profit", "net_profit_growth"},
		RequestedFieldsReady:   false,
		ReviewRequiredFields:   []string{"metrics_trend_series"},
		GuardStatus:            "needs_review",
		StopAfterGuardPassed:   false,
		TrendRequested:         true,
		TrendSeriesReady:       false,
		TrendSeriesPeriodCount: 1,
	})
	if result.Passed ||
		result.TrendSeriesReady ||
		!containsSubstring(result.FailureReason, "trend_series_incomplete") {
		t.Fatalf("expected trend request without series to fail, got %#v", result)
	}
}

func TestEvaluateLatestMetricsEvidencePassesReadyTrendSeries(t *testing.T) {
	result := EvaluateLatestMetricsEvidence(LatestMetricsEvaluationInput{
		ExpectedEntityName:     "中国银行",
		EvidenceEntityName:     "中国银行",
		SourceURL:              "https://data.eastmoney.com/bbsj/601988.html?code=SH601988",
		ReportPeriod:           "2025-12-31",
		RequestedFields:        []string{"revenue", "revenue_growth", "net_profit", "net_profit_growth"},
		RequestedFieldsReady:   true,
		GuardStatus:            "passed",
		StopAfterGuardPassed:   true,
		TrendRequested:         true,
		TrendSeriesReady:       true,
		TrendSeriesPeriodCount: 3,
	})
	if !result.Passed ||
		!result.TrendRequested ||
		!result.TrendSeriesReady ||
		result.TrendSeriesPeriodCount != 3 ||
		result.FailureReason != "" {
		t.Fatalf("expected ready trend series to pass, got %#v", result)
	}
}

func containsSubstring(value string, needle string) bool {
	for i := 0; i+len(needle) <= len(value); i++ {
		if value[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
