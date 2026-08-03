package finance

import (
	"strings"
	"testing"

	financialreportbrief "github.com/wsnacj/agentx-go/scenes/finance/brief"
)

func TestBuildBriefTextSkipsRawOrReviewRequiredSnippets(t *testing.T) {
	evidence := financialreportbrief.BriefEvidence{
		CompanyName:  "腾讯控股",
		ReportPeriod: "2025-12-31",
		KeyPoints: []financialreportbrief.BriefKeyPoint{
			{Category: "financial", Title: "核心财务表现", Text: "收入 7517.66亿元；利润 2248.42亿元"},
			{Category: "business", Title: "主营业务", Text: "We are a leading internet company that provides social platforms, digital content, fintech and cloud services to customers."},
			{Category: "risk", Title: "风险与不确定性", Text: "主要风险包括监管变化和市场竞争。", ReviewRequired: true},
			{Category: "outlook", Title: "未来展望", Text: "公司将继续投入 AI、云和内容生态。"},
		},
	}

	text := BuildBriefText(evidence)
	if !strings.Contains(text, "核心财务表现：收入 7517.66亿元；利润 2248.42亿元") ||
		!strings.Contains(text, "未来展望：公司将继续投入 AI、云和内容生态。") {
		t.Fatalf("expected usable key points in brief text, got %q", text)
	}
	if strings.Contains(text, "We are a leading internet company") ||
		strings.Contains(text, "主要风险包括监管变化和市场竞争") {
		t.Fatalf("expected raw/review-required snippets to stay out of generated brief text, got %q", text)
	}
	if !strings.Contains(text, "补充说明：业务概况/风险信息已在报告证据中定位") {
		t.Fatalf("expected synthesis boundary note, got %q", text)
	}
}

func TestBuildBriefTextLabelsQuarterlySECFiling(t *testing.T) {
	evidence := financialreportbrief.BriefEvidence{
		CompanyName:  "MICROSOFT CORP",
		ReportPeriod: "2026-03-31 10-Q",
		KeyPoints: []financialreportbrief.BriefKeyPoint{
			{Category: "financial", Title: "核心财务表现", Text: "收入 USD82.886 billion；利润 USD31.778 billion"},
		},
	}

	text := BuildBriefText(evidence)
	if !strings.Contains(text, "2026-03-31 10-Q 季报简报") || strings.Contains(text, "年报简报") {
		t.Fatalf("expected 10-Q brief to use a quarterly label, got %q", text)
	}
}

func TestBuildBriefTextKeepsAnnualLabelForAnnualEvidence(t *testing.T) {
	evidence := financialreportbrief.BriefEvidence{
		CompanyName:  "Baidu, Inc.",
		ReportPeriod: "2025-12-31",
		KeyPoints: []financialreportbrief.BriefKeyPoint{
			{Category: "financial", Title: "核心财务表现", Text: "收入 RMB129.079 billion"},
		},
	}

	if text := BuildBriefText(evidence); !strings.Contains(text, "2025-12-31 年报简报") {
		t.Fatalf("expected annual evidence to retain annual label, got %q", text)
	}
}

func TestBriefFinancialKeyPointIncludesOperatingCashFlow(t *testing.T) {
	point := BriefFinancialKeyPoint(financialreportbrief.BriefEvidence{
		Metrics: []financialreportbrief.BriefMetric{
			{Name: "revenue", Value: "4573.00亿元", Source: "official_report"},
			{Name: "net_profit", Value: "416.43亿元", Source: "official_report"},
			{Name: "operating_cash_flow", Value: "341.42亿元", Source: "official_report"},
		},
	})
	if !strings.Contains(point.Text, "收入 4573.00亿元") ||
		!strings.Contains(point.Text, "利润 416.43亿元") ||
		!strings.Contains(point.Text, "经营现金流 341.42亿元") {
		t.Fatalf("expected financial key point to include operating cash flow, got %q", point.Text)
	}
}

func TestBriefFinancialKeyPointIncludesFieldPeriods(t *testing.T) {
	point := BriefFinancialKeyPoint(financialreportbrief.BriefEvidence{
		Metrics: []financialreportbrief.BriefMetric{
			{Name: "revenue", Value: "USD111.184 billion", Period: "三个月", Source: "sec"},
			{Name: "operating_cash_flow", Value: "USD82.627 billion", Period: "六个月累计", Source: "sec"},
		},
	})
	if !strings.Contains(point.Text, "收入（三个月） USD111.184 billion") ||
		!strings.Contains(point.Text, "经营现金流（六个月累计） USD82.627 billion") {
		t.Fatalf("expected field-level periods in financial key point, got %q", point.Text)
	}
}
