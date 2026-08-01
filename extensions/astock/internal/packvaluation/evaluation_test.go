package packvaluation

import (
	"strings"
	"testing"

	astockcontracts "github.com/wsnacj/agentx-go/extensions/astock/contracts"
)

func TestEvaluateValuationEvidencePassesSourceBackedSnapshot(t *testing.T) {
	eval := EvaluateValuationEvidence(ValuationEvaluationInput{
		ExpectedEntityName: "同花顺",
		ExpectedStockCode:  "300033",
		EvidenceEntityName: "同花顺",
		EvidenceStockCode:  "sz300033",
		AdapterStatus:      astockcontracts.AdapterStatusOK,
		FailureCode:        astockcontracts.FailureCodeNone,
		AnswerReady:        true,
		RequestedFields:    []string{"price", "change_pct", "turnover", "pe_ttm", "pb", "market_cap"},
		FieldValues: map[string]string{
			"price":            "100.00",
			"change_percent":   "2.04",
			"turnover_percent": "4.55",
			"pe_ttm":           "31.76",
			"pb":               "7.22",
			"market_cap":       "500.88",
		},
		AsOf:                      "2026-05-15T15:00:00+08:00",
		SourceURL:                 "https://qt.gtimg.cn/q=sz300033",
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

func TestEvaluateValuationEvidenceRejectsFieldGap(t *testing.T) {
	eval := EvaluateValuationEvidence(ValuationEvaluationInput{
		ExpectedEntityName: "同花顺",
		EvidenceEntityName: "同花顺",
		AdapterStatus:      astockcontracts.AdapterStatusOK,
		FailureCode:        astockcontracts.FailureCodeNone,
		AnswerReady:        true,
		RequestedFields:    []string{"price", "pe_ttm", "pb"},
		FieldValues: map[string]string{
			"price":  "100.00",
			"pe_ttm": "31.76",
		},
		AsOf:      "2026-05-15T15:00:00+08:00",
		SourceURL: "https://qt.gtimg.cn/q=sz300033",
	})

	if eval.Passed || eval.FieldsReady {
		t.Fatalf("expected missing requested field to fail, got %#v", eval)
	}
	if !strings.Contains(eval.FailureReason, "requested_quote_fields_missing") {
		t.Fatalf("expected requested field failure reason, got %q", eval.FailureReason)
	}
}

func TestEvaluateValuationEvidenceRejectsInvalidSource(t *testing.T) {
	eval := EvaluateValuationEvidence(ValuationEvaluationInput{
		ExpectedEntityName: "同花顺",
		EvidenceEntityName: "同花顺",
		AdapterStatus:      astockcontracts.AdapterStatusOK,
		FailureCode:        astockcontracts.FailureCodeNone,
		AnswerReady:        true,
		RequestedFields:    []string{"price"},
		FieldValues:        map[string]string{"price": "100.00"},
		AsOf:               "2026-05-15T15:00:00+08:00",
		SourceURL:          "local-cache",
	})

	if eval.Passed || eval.SourceAccepted {
		t.Fatalf("expected source guard to fail, got %#v", eval)
	}
	if !strings.Contains(eval.FailureReason, "source_unaccepted") {
		t.Fatalf("expected source failure reason, got %q", eval.FailureReason)
	}
}

func TestEvaluateValuationEvidenceRejectsAdviceWithoutBoundary(t *testing.T) {
	eval := EvaluateValuationEvidence(ValuationEvaluationInput{
		ExpectedEntityName:        "同花顺",
		EvidenceEntityName:        "同花顺",
		AdapterStatus:             astockcontracts.AdapterStatusOK,
		FailureCode:               astockcontracts.FailureCodeNone,
		AnswerReady:               true,
		RequestedFields:           []string{"price"},
		FieldValues:               map[string]string{"price": "100.00"},
		AsOf:                      "2026-05-15T15:00:00+08:00",
		SourceURL:                 "https://qt.gtimg.cn/q=sz300033",
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
