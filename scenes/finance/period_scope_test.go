package finance

import "testing"

func TestMetricsPeriodScopeReviewFieldsRejectsAnnualForRecentQuarters(t *testing.T) {
	spec := MetricSpec{Revenue: true, NetProfit: true}
	evidence := MetricsEvidence{
		ReportPeriod: "2025-12-31",
		Revenue:      "100亿元",
		NetProfit:    "10亿元",
		TrendSeries: []MetricsTrendSeriesPoint{
			{Period: "2025-12-31", Revenue: "100亿元", NetProfit: "10亿元", Source: "https://example.com/report"},
			{Period: "2024-12-31", Revenue: "90亿元", NetProfit: "9亿元", Source: "https://example.com/report"},
		},
	}
	fields := MetricsPeriodScopeReviewFields(MetricsPeriodScopeRequest{
		PeriodScope:    "recent_quarters",
		TrendRequested: true,
		Spec:           spec,
	}, evidence)
	if len(fields) != 1 || fields[0] != MetricPeriodScopeReviewField {
		t.Fatalf("expected period_scope review, got %#v", fields)
	}
}

func TestMetricsPeriodScopeReviewFieldsAcceptsQuarterSeries(t *testing.T) {
	spec := MetricSpec{Revenue: true, NetProfit: true}
	evidence := MetricsEvidence{
		ReportPeriod: "2026-03-31",
		Revenue:      "100亿元",
		NetProfit:    "10亿元",
		TrendSeries: []MetricsTrendSeriesPoint{
			{Period: "2026-03-31", Revenue: "100亿元", NetProfit: "10亿元", Source: "https://example.com/report"},
			{Period: "2025-09-30", Revenue: "90亿元", NetProfit: "9亿元", Source: "https://example.com/report"},
		},
	}
	fields := MetricsPeriodScopeReviewFields(MetricsPeriodScopeRequest{
		PeriodScope:    "recent_quarters",
		TrendRequested: true,
		Spec:           spec,
	}, evidence)
	if len(fields) != 0 {
		t.Fatalf("expected no period scope review, got %#v", fields)
	}
}

func TestQuarterTrendSeriesPeriodCountCountsSameYearQuarters(t *testing.T) {
	spec := MetricSpec{Revenue: true, NetProfit: true}
	series := []MetricsTrendSeriesPoint{
		{Period: "2026-06-30 6-K", Revenue: "120亿元", NetProfit: "12亿元", Source: "https://example.com/q2"},
		{Period: "2026-03-31 6-K", Revenue: "100亿元", NetProfit: "10亿元", Source: "https://example.com/q1"},
	}
	if got := QuarterTrendSeriesPeriodCount(spec, series); got != 2 {
		t.Fatalf("quarter trend period count = %d, want 2", got)
	}
	if !MetricQuarterTrendSeriesReady(spec, series) {
		t.Fatalf("expected same-year 6-K quarter series to be ready")
	}
	if got := TrendSeriesPeriodCount(spec, series); got != 1 {
		t.Fatalf("annual trend period count should still de-duplicate by year, got %d", got)
	}
}

func TestMetricReportPeriodTreatsSingleQuarterQ4AsQuarter(t *testing.T) {
	if MetricReportPeriodLooksAnnual("2025-12-31 单季报") {
		t.Fatalf("single-quarter Q4 should not be treated as annual")
	}
	if !MetricReportPeriodLooksQuarter("2025-12-31 单季报") {
		t.Fatalf("single-quarter Q4 should be treated as quarter evidence")
	}
	if MetricReportPeriodLooksAnnual("2025-12-31 Q4") {
		t.Fatalf("Q4 period label should not be treated as annual")
	}
	if !MetricReportPeriodLooksQuarter("2025-12-31 Q4") {
		t.Fatalf("Q4 period label should be treated as quarter evidence")
	}
}

func TestMetricsPeriodScopeReviewFieldsRejectsQuarterForAnnual(t *testing.T) {
	fields := MetricsPeriodScopeReviewFields(MetricsPeriodScopeRequest{
		PeriodScope: "latest_annual",
		Spec:        MetricSpec{Revenue: true},
	}, MetricsEvidence{ReportPeriod: "2026-03-31", Revenue: "100亿元"})
	if len(fields) != 1 || fields[0] != MetricPeriodScopeReviewField {
		t.Fatalf("expected period_scope review, got %#v", fields)
	}
}

func TestMetricsPeriodScopeReviewFieldsRecognizesDisclosedPeriodAliases(t *testing.T) {
	spec := MetricSpec{Revenue: true, NetProfit: true}
	quarterFields := MetricsPeriodScopeReviewFields(MetricsPeriodScopeRequest{
		PeriodScope:    "latest_disclosed_quarters",
		TrendRequested: true,
		Spec:           spec,
	}, MetricsEvidence{
		ReportPeriod: "2026-03-31",
		Revenue:      "100亿元",
		NetProfit:    "10亿元",
		TrendSeries: []MetricsTrendSeriesPoint{
			{Period: "2026-03-31", Revenue: "100亿元", NetProfit: "10亿元", Source: "https://example.com/report"},
		},
	})
	if len(quarterFields) != 1 || quarterFields[0] != MetricPeriodScopeReviewField {
		t.Fatalf("expected latest_disclosed_quarters to require a quarter series, got %#v", quarterFields)
	}

	annualFields := MetricsPeriodScopeReviewFields(MetricsPeriodScopeRequest{
		PeriodScope: "latest_disclosed_annual_three_years",
		Spec:        spec,
	}, MetricsEvidence{
		ReportPeriod: "2026-03-31",
		Revenue:      "100亿元",
		NetProfit:    "10亿元",
	})
	if len(annualFields) != 1 || annualFields[0] != MetricPeriodScopeReviewField {
		t.Fatalf("expected annual-three-years alias to reject quarter evidence, got %#v", annualFields)
	}
}

func TestMetricsPeriodScopeReviewFieldsRejectsSingleAnnualForRecentYears(t *testing.T) {
	spec := MetricSpec{Revenue: true, NetProfit: true}
	fields := MetricsPeriodScopeReviewFields(MetricsPeriodScopeRequest{
		PeriodScope:    "recent_years",
		TrendRequested: true,
		Spec:           spec,
	}, MetricsEvidence{
		ReportPeriod: "2025-12-31",
		Revenue:      "100亿元",
		NetProfit:    "10亿元",
	})
	if len(fields) != 1 || fields[0] != MetricPeriodScopeReviewField {
		t.Fatalf("expected recent_years to require annual trend coverage, got %#v", fields)
	}
}

func TestMetricsPeriodScopeReviewFieldsAcceptsRecentYearsSeries(t *testing.T) {
	spec := MetricSpec{Revenue: true, NetProfit: true}
	fields := MetricsPeriodScopeReviewFields(MetricsPeriodScopeRequest{
		PeriodScope:    "recent_years",
		TrendRequested: true,
		Spec:           spec,
	}, MetricsEvidence{
		ReportPeriod: "2025-12-31",
		Revenue:      "100亿元",
		NetProfit:    "10亿元",
		TrendSeries: []MetricsTrendSeriesPoint{
			{Period: "2025-12-31", Revenue: "100亿元", NetProfit: "10亿元", Source: "https://example.com/report"},
			{Period: "2024-12-31", Revenue: "90亿元", NetProfit: "9亿元", Source: "https://example.com/report"},
		},
	})
	if len(fields) != 0 {
		t.Fatalf("expected recent_years trend coverage to pass, got %#v", fields)
	}
}

func TestMetricsPeriodScopeReviewFieldsLatestDisclosedReportAcceptsSingleKnownPeriod(t *testing.T) {
	spec := MetricSpec{Revenue: true, NetProfit: true}
	annualFields := MetricsPeriodScopeReviewFields(MetricsPeriodScopeRequest{
		PeriodScope: "latest_disclosed_report",
		Spec:        spec,
	}, MetricsEvidence{
		ReportPeriod: "2025-12-31",
		Revenue:      "100亿元",
		NetProfit:    "10亿元",
	})
	if len(annualFields) != 0 {
		t.Fatalf("expected latest_disclosed_report to accept a verified annual period, got %#v", annualFields)
	}
	quarterFields := MetricsPeriodScopeReviewFields(MetricsPeriodScopeRequest{
		PeriodScope: "latest_disclosed_report",
		Spec:        spec,
	}, MetricsEvidence{
		ReportPeriod: "2026-03-31 一季报",
		Revenue:      "100亿元",
		NetProfit:    "10亿元",
	})
	if len(quarterFields) != 0 {
		t.Fatalf("expected latest_disclosed_report to accept a verified quarter period, got %#v", quarterFields)
	}
}

func TestMetricsPeriodScopeReviewFieldsLatestDisclosedReportRejectsUnknownPeriod(t *testing.T) {
	fields := MetricsPeriodScopeReviewFields(MetricsPeriodScopeRequest{
		PeriodScope: "latest_disclosed_report",
		Spec:        MetricSpec{Revenue: true},
	}, MetricsEvidence{
		ReportPeriod: "unknown",
		Revenue:      "100亿元",
	})
	if len(fields) != 1 || fields[0] != MetricPeriodScopeReviewField {
		t.Fatalf("expected latest_disclosed_report to require a known disclosure period, got %#v", fields)
	}
}
