package metrics

import "testing"

func TestBuildLatestMetricsCaseInputFromAshareShortQuery(t *testing.T) {
	input, ok := BuildLatestMetricsCaseInput("帮我去东方财富找下贵州茅台最新的财报，看看收入和利润情况，以及这两个指标的增长情况")
	if !ok {
		t.Fatalf("expected case input")
	}
	entity, _ := input["entity"].(map[string]any)
	if entity["name"] != "贵州茅台" {
		t.Fatalf("unexpected entity: %#v", entity)
	}
	if got := input["requested_metrics"]; !sameStringAnySlice(got, []string{"revenue", "revenue_growth", "net_profit", "net_profit_growth"}) {
		t.Fatalf("unexpected requested metrics: %#v", got)
	}
	if got := input["requested_outputs"]; !sameStringAnySlice(got, []string{"metrics"}) {
		t.Fatalf("unexpected requested outputs: %#v", got)
	}
	if assessment, _ := input["assessment"].(map[string]any); assessment["kind"] != AssessmentKindNone {
		t.Fatalf("unexpected assessment frame: %#v", assessment)
	}
	if input["period_policy"] != "latest_disclosed_report" || input["stop_condition"] != "guard_passed" {
		t.Fatalf("unexpected case policy fields: %#v", input)
	}
}

func TestBuildLatestMetricsCaseInputPreservesExplicitStockCode(t *testing.T) {
	input, ok := BuildLatestMetricsCaseInput("请去东方财富上找到瑞可达 (688800) 这只股票的财报，并提取它的营收、净利润。")
	if !ok {
		t.Fatalf("expected case input")
	}
	entity, _ := input["entity"].(map[string]any)
	identifiers, _ := entity["identifiers"].(map[string]any)
	if entity["name"] != "瑞可达" || identifiers["stock_code"] != "688800" {
		t.Fatalf("unexpected entity frame: %#v", entity)
	}
	if got := input["requested_metrics"]; !sameStringAnySlice(got, []string{"revenue", "net_profit"}) {
		t.Fatalf("unexpected requested metrics: %#v", got)
	}
}

func TestBuildLatestMetricsCaseInputKeepsGenericEntityWithoutSiteLogic(t *testing.T) {
	input, ok := BuildLatestMetricsCaseInput("去东方财富看下腾讯科技的最新财报，25年营收和利润分别是多少")
	if !ok {
		t.Fatalf("expected case input")
	}
	entity, _ := input["entity"].(map[string]any)
	if entity["name"] != "腾讯科技" {
		t.Fatalf("unexpected entity: %#v", entity)
	}
	if got := input["requested_metrics"]; !sameStringAnySlice(got, []string{"revenue", "net_profit"}) {
		t.Fatalf("unexpected requested metrics: %#v", got)
	}
}

func TestBuildLatestMetricsCaseInputFromSEC20FShortQuery(t *testing.T) {
	input, ok := BuildLatestMetricsCaseInput("去SEC官方20-F看下百度25年营收和利润分别是多少")
	if !ok {
		t.Fatalf("expected case input")
	}
	entity, _ := input["entity"].(map[string]any)
	if entity["name"] != "百度" {
		t.Fatalf("unexpected entity: %#v", entity)
	}
	if input["period_policy"] != "latest_disclosed_annual" {
		t.Fatalf("unexpected period policy: %#v", input)
	}
	if got := input["requested_metrics"]; !sameStringAnySlice(got, []string{"revenue", "net_profit"}) {
		t.Fatalf("unexpected requested metrics: %#v", got)
	}
}

func TestBuildLatestMetricsCaseInputSkipsPronounMetricClause(t *testing.T) {
	input, ok := BuildLatestMetricsCaseInput("请帮我查下中国银行最新的财报，看下他的收入和利润，以及近几年的收入和利润增长情况")
	if !ok {
		t.Fatalf("expected case input")
	}
	entity, _ := input["entity"].(map[string]any)
	if entity["name"] != "中国银行" {
		t.Fatalf("unexpected entity: %#v", entity)
	}
	if got := input["requested_metrics"]; !sameStringAnySlice(got, []string{"revenue", "revenue_growth", "net_profit", "net_profit_growth"}) {
		t.Fatalf("unexpected requested metrics: %#v", got)
	}
}

func TestBuildLatestMetricsCaseInputSupportsChaYiXia(t *testing.T) {
	input, ok := BuildLatestMetricsCaseInput("请查一下腾讯音乐最新的财报，看下他的收入和利润情况")
	if !ok {
		t.Fatalf("expected case input")
	}
	entity, _ := input["entity"].(map[string]any)
	if entity["name"] != "腾讯音乐" {
		t.Fatalf("unexpected entity: %#v", entity)
	}
	if got := input["requested_metrics"]; !sameStringAnySlice(got, []string{"revenue", "net_profit"}) {
		t.Fatalf("unexpected requested metrics: %#v", got)
	}
}

func TestBuildLatestMetricsCaseInputDefaultsReportMetricsForPerformanceQuestion(t *testing.T) {
	input, ok := BuildLatestMetricsCaseInput("帮我查下百望云的最新财报，看下他们的业绩情况，帮我评估这家公司是否靠谱值得投资")
	if !ok {
		t.Fatalf("expected case input")
	}
	entity, _ := input["entity"].(map[string]any)
	if entity["name"] != "百望云" {
		t.Fatalf("unexpected entity: %#v", entity)
	}
	if got := input["requested_metrics"]; !sameStringAnySlice(got, []string{"revenue", "revenue_growth", "net_profit", "net_profit_growth", "operating_cash_flow"}) {
		t.Fatalf("unexpected requested metrics: %#v", got)
	}
	if got := input["requested_outputs"]; !sameStringAnySlice(got, []string{"metrics", "performance_assessment", "investment_assessment"}) {
		t.Fatalf("unexpected requested outputs: %#v", got)
	}
	assessment, _ := input["assessment"].(map[string]any)
	if assessment["kind"] != AssessmentKindInvestmentRisk || assessment["requires_valuation"] != true {
		t.Fatalf("unexpected assessment frame: %#v", assessment)
	}
}

func TestBuildLatestMetricsCaseInputRequestsPerformanceAssessment(t *testing.T) {
	input, ok := BuildLatestMetricsCaseInput("查一下联想最新的财报，看下他的业绩情况，给个你的评估")
	if !ok {
		t.Fatalf("expected case input")
	}
	entity, _ := input["entity"].(map[string]any)
	if entity["name"] != "联想" {
		t.Fatalf("unexpected entity: %#v", entity)
	}
	if got := input["requested_metrics"]; !sameStringAnySlice(got, []string{"revenue", "revenue_growth", "net_profit", "net_profit_growth", "operating_cash_flow"}) {
		t.Fatalf("unexpected requested metrics: %#v", got)
	}
	if got := input["requested_outputs"]; !sameStringAnySlice(got, []string{"metrics", "performance_assessment"}) {
		t.Fatalf("unexpected requested outputs: %#v", got)
	}
	assessment, _ := input["assessment"].(map[string]any)
	if assessment["kind"] != AssessmentKindBusinessPerformance || assessment["requires_valuation"] != false {
		t.Fatalf("unexpected assessment frame: %#v", assessment)
	}
}

func TestBuildLatestMetricsCaseInputRequestsPerformanceAssessmentFromWeakChineseWording(t *testing.T) {
	input, ok := BuildLatestMetricsCaseInput("看看联想最新财报，表现怎么样")
	if !ok {
		t.Fatalf("expected case input")
	}
	entity, _ := input["entity"].(map[string]any)
	if entity["name"] != "联想" {
		t.Fatalf("unexpected entity: %#v", entity)
	}
	if got := input["requested_outputs"]; !sameStringAnySlice(got, []string{"metrics", "performance_assessment"}) {
		t.Fatalf("unexpected requested outputs: %#v", got)
	}
	assessment, _ := input["assessment"].(map[string]any)
	if assessment["kind"] != AssessmentKindBusinessPerformance {
		t.Fatalf("unexpected assessment frame: %#v", assessment)
	}
}

func TestBuildLatestMetricsCaseInputRequestsPerformanceAssessmentFromEnglishWording(t *testing.T) {
	input, ok := BuildLatestMetricsCaseInput("Check Lenovo's latest financial report and assess its business performance")
	if !ok {
		t.Fatalf("expected case input")
	}
	entity, _ := input["entity"].(map[string]any)
	if entity["name"] != "Lenovo" {
		t.Fatalf("unexpected entity: %#v", entity)
	}
	if got := input["requested_outputs"]; !sameStringAnySlice(got, []string{"metrics", "performance_assessment"}) {
		t.Fatalf("unexpected requested outputs: %#v", got)
	}
	assessment, _ := input["assessment"].(map[string]any)
	if assessment["kind"] != AssessmentKindBusinessPerformance {
		t.Fatalf("unexpected assessment frame: %#v", assessment)
	}
}

func TestBuildLatestMetricsCaseInputExpandsRichReportFields(t *testing.T) {
	input, ok := BuildLatestMetricsCaseInput("请去东方财富看下爱奇艺最新财报，尽可能提取丰富字段。")
	if !ok {
		t.Fatalf("expected case input")
	}
	entity, _ := input["entity"].(map[string]any)
	if entity["name"] != "爱奇艺" {
		t.Fatalf("unexpected entity: %#v", entity)
	}
	if got := input["requested_metrics"]; !sameStringAnySlice(got, []string{"revenue", "revenue_growth", "net_profit", "net_profit_growth", "operating_cash_flow"}) {
		t.Fatalf("unexpected requested metrics: %#v", got)
	}
}

func TestBuildReportMetricsCaseInputSelectsTrendCase(t *testing.T) {
	input, caseType, ok := BuildReportMetricsCaseInput("帮我查询腾讯音乐近几年的财报，看下他收入和利润近几年增长情况")
	if !ok {
		t.Fatalf("expected case input")
	}
	if caseType != CaseTypeTrend {
		t.Fatalf("expected trend case type, got %q input=%#v", caseType, input)
	}
	if input["period_policy"] != "recent_years" || input["stop_condition"] != "guard_passed_or_review_required" {
		t.Fatalf("unexpected trend policy fields: %#v", input)
	}
	entity, _ := input["entity"].(map[string]any)
	if entity["name"] != "腾讯音乐" {
		t.Fatalf("unexpected entity: %#v", entity)
	}
}

func TestBuildLatestMetricsCaseInputStripsRequestedYearFromEntity(t *testing.T) {
	for _, message := range []string{
		"请去东方财富看下东方雨虹的2025年年报，营收、净利润以及这两个指标的增长情况分别是多少",
		"请去东方财富看下东方雨虹25年年报，营收、净利润以及这两个指标的增长情况分别是多少",
	} {
		input, ok := BuildLatestMetricsCaseInput(message)
		if !ok {
			t.Fatalf("expected case input for %q", message)
		}
		entity, _ := input["entity"].(map[string]any)
		if entity["name"] != "东方雨虹" {
			t.Fatalf("unexpected entity for %q: %#v", message, entity)
		}
		if input["period_policy"] != "latest_disclosed_annual" {
			t.Fatalf("unexpected period policy for %q: %#v", message, input)
		}
		if got := input["requested_metrics"]; !sameStringAnySlice(got, []string{"revenue", "revenue_growth", "net_profit", "net_profit_growth"}) {
			t.Fatalf("unexpected requested metrics for %q: %#v", message, got)
		}
	}
}

func TestBuildLatestMetricsCaseInputRejectsMissingEntity(t *testing.T) {
	if input, ok := BuildLatestMetricsCaseInput("请基于这个官方页面 https://example.com/report 提取营收和净利润"); ok {
		t.Fatalf("did not expect case input without entity, got %#v", input)
	}
}

func sameStringAnySlice(got any, want []string) bool {
	items, ok := got.([]any)
	if !ok || len(items) != len(want) {
		return false
	}
	for i, item := range items {
		value, ok := item.(string)
		if !ok || value != want[i] {
			return false
		}
	}
	return true
}
