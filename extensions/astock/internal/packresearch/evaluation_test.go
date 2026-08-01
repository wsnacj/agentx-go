package packresearch

import (
	"strings"
	"testing"

	astockcontracts "github.com/wsnacj/agentx-go/extensions/astock/contracts"
)

func TestEvaluateResearchEvidencePassesSourceBackedReport(t *testing.T) {
	eval := EvaluateResearchEvidence(ResearchEvaluationInput{
		ExpectedEntityName: "贵州茅台",
		ExpectedStockCode:  "600519",
		EvidenceEntityName: "贵州茅台",
		EvidenceStockCode:  "SH600519",
		AdapterStatus:      astockcontracts.AdapterStatusOK,
		FailureCode:        astockcontracts.FailureCodeNone,
		AnswerReady:        true,
		RequestedFields:    []string{"title", "institution", "published_at", "rating", "eps_forecast", "pdf_url"},
		FieldValues: map[string]string{
			"title":        "业绩稳健增长",
			"institution":  "测试证券",
			"published_at": "2026-05-15",
			"rating":       "买入",
			"pdf_url":      "https://example.test/report.pdf",
		},
		ConsensusFields:           []string{"eps_forecast"},
		ReportCount:               2,
		LatestPublishedAt:         "2026-05-15",
		SourceURLs:                []string{"https://data.eastmoney.com/report/stock.jshtml"},
		InvestmentAdviceRequested: true,
		AdviceBoundaryStated:      true,
	})

	if !eval.Passed {
		t.Fatalf("expected evaluation to pass, got %#v", eval)
	}
	if !eval.SubjectCorrect || !eval.FreshnessAccepted || !eval.FieldsReady || !eval.SourceAccepted || !eval.AdviceBoundaryRespected {
		t.Fatalf("expected all guard dimensions to pass, got %#v", eval)
	}
}

func TestEvaluateResearchEvidenceRejectsMissingReports(t *testing.T) {
	eval := EvaluateResearchEvidence(ResearchEvaluationInput{
		ExpectedEntityName: "贵州茅台",
		EvidenceEntityName: "贵州茅台",
		AdapterStatus:      astockcontracts.AdapterStatusOK,
		FailureCode:        astockcontracts.FailureCodeNone,
		AnswerReady:        true,
		RequestedFields:    []string{"title"},
		FieldValues:        map[string]string{"title": "业绩稳健增长"},
		LatestPublishedAt:  "2026-05-15",
		SourceURLs:         []string{"https://data.eastmoney.com/report/stock.jshtml"},
	})

	if eval.Passed || eval.FieldsReady {
		t.Fatalf("expected missing reports to fail, got %#v", eval)
	}
	if !strings.Contains(eval.FailureReason, "research_reports_missing") {
		t.Fatalf("expected missing reports failure reason, got %q", eval.FailureReason)
	}
}

func TestEvaluateResearchEvidenceRejectsRequestedFieldGap(t *testing.T) {
	eval := EvaluateResearchEvidence(ResearchEvaluationInput{
		ExpectedEntityName: "贵州茅台",
		EvidenceEntityName: "贵州茅台",
		AdapterStatus:      astockcontracts.AdapterStatusOK,
		FailureCode:        astockcontracts.FailureCodeNone,
		AnswerReady:        true,
		RequestedFields:    []string{"title", "profit_forecast"},
		FieldValues:        map[string]string{"title": "业绩稳健增长"},
		ReportCount:        1,
		LatestPublishedAt:  "2026-05-15",
		SourceURLs:         []string{"https://data.eastmoney.com/report/stock.jshtml"},
	})

	if eval.Passed || eval.FieldsReady {
		t.Fatalf("expected requested field gap to fail, got %#v", eval)
	}
	if !strings.Contains(eval.FailureReason, "requested_research_fields_missing") {
		t.Fatalf("expected requested field failure reason, got %q", eval.FailureReason)
	}
}

func TestEvaluateResearchEvidenceRejectsAdviceWithoutBoundary(t *testing.T) {
	eval := EvaluateResearchEvidence(ResearchEvaluationInput{
		ExpectedEntityName:        "贵州茅台",
		EvidenceEntityName:        "贵州茅台",
		AdapterStatus:             astockcontracts.AdapterStatusOK,
		FailureCode:               astockcontracts.FailureCodeNone,
		AnswerReady:               true,
		RequestedFields:           []string{"title"},
		FieldValues:               map[string]string{"title": "业绩稳健增长"},
		ReportCount:               1,
		LatestPublishedAt:         "2026-05-15",
		SourceURLs:                []string{"https://data.eastmoney.com/report/stock.jshtml"},
		InvestmentAdviceRequested: true,
		AdviceBoundaryStated:      false,
	})

	if eval.Passed || eval.AdviceBoundaryRespected {
		t.Fatalf("expected advice boundary guard to fail, got %#v", eval)
	}
	if !strings.Contains(eval.FailureReason, "advice_boundary_missing") {
		t.Fatalf("expected advice boundary failure reason, got %q", eval.FailureReason)
	}
}
