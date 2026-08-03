package brief

import "testing"

func TestEvaluateBriefEvidencePassesCompleteGuardedBrief(t *testing.T) {
	result := EvaluateBriefEvidence(BriefEvaluationInput{
		ExpectedEntityName:   "浦发银行",
		ExpectedStockCode:    "600000",
		GuardStatus:          "passed",
		StopAfterGuardPassed: true,
		Evidence: BriefEvidence{
			CompanyName:  "浦发银行",
			StockCode:    "600000.SH",
			ReportPeriod: "2025-12-31",
			SourceURL:    "https://static.cninfo.com.cn/finalpage/2026-03-30/report.pdf",
			Brief:        "浦发银行2025年报显示，收入和利润已披露，管理层讨论了经营修复和主要风险。",
			KeyPoints: []BriefKeyPoint{
				{Category: "financial", Text: "营业收入和净利润已披露。"},
				{Category: "business", Text: "管理层讨论了经营表现。"},
				{Category: "risk", Text: "年报披露了信用风险和市场风险。"},
			},
			Metrics: []BriefMetric{
				{Name: "revenue", Value: "100亿元"},
				{Name: "net_profit", Value: "10亿元"},
			},
		},
	})
	if !result.Passed || result.FailureReason != "" {
		t.Fatalf("expected complete brief evidence to pass, got %#v", result)
	}
}

func TestEvaluateBriefEvidenceNormalizesMarketCodeSuffixes(t *testing.T) {
	result := EvaluateBriefEvidence(BriefEvaluationInput{
		ExpectedEntityName:   "阿里巴巴",
		ExpectedStockCode:    "BABA.N",
		GuardStatus:          "passed",
		StopAfterGuardPassed: true,
		Evidence: BriefEvidence{
			CompanyName:  "Alibaba Group Holding Ltd",
			StockCode:    "BABA",
			ReportPeriod: "2026-03-31 20-F",
			SourceURL:    "https://www.sec.gov/Archives/edgar/data/1577552/example.htm",
			Brief:        "Alibaba Group Holding Ltd 20-F brief with financials and risks.",
			KeyPoints: []BriefKeyPoint{
				{Category: "financial", Text: "Revenue and net profit are disclosed."},
				{Category: "business", Text: "Business operations are discussed."},
				{Category: "risk", Text: "Risk factors are disclosed."},
			},
			Metrics: []BriefMetric{
				{Name: "revenue", Value: "RMB1023.67 billion"},
				{Name: "net_profit", Value: "RMB103.592 billion"},
			},
		},
	})
	if !result.Passed || !result.SubjectCorrect {
		t.Fatalf("expected BABA.N request to match BABA evidence, got %#v", result)
	}
}

func TestEvaluateBriefEvidenceFailsMissingRiskOrOutlook(t *testing.T) {
	result := EvaluateBriefEvidence(BriefEvaluationInput{
		ExpectedEntityName:   "浦发银行",
		GuardStatus:          "passed",
		StopAfterGuardPassed: true,
		Evidence: BriefEvidence{
			CompanyName:  "浦发银行",
			ReportPeriod: "2025-12-31",
			SourceURL:    "https://static.cninfo.com.cn/finalpage/2026-03-30/report.pdf",
			Brief:        "浦发银行年报简报。",
			KeyPoints: []BriefKeyPoint{
				{Category: "financial", Text: "营业收入和净利润已披露。"},
				{Category: "business", Text: "管理层讨论了经营表现。"},
				{Category: "dividend", Text: "披露分红方案。"},
			},
			Metrics: []BriefMetric{{Name: "revenue", Value: "100亿元"}},
		},
	})
	if result.Passed || !containsBriefSubstring(result.FailureReason, "risk_or_outlook_missing") {
		t.Fatalf("expected missing risk/outlook failure, got %#v", result)
	}
}

func containsBriefSubstring(value string, needle string) bool {
	for i := 0; i+len(needle) <= len(value); i++ {
		if value[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
