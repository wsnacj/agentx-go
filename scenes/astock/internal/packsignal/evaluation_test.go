package packsignal

import (
	"strings"
	"testing"

	astockcontracts "github.com/wsnacj/agentx-go/scenes/astock/contracts"
)

func TestEvaluateSignalEvidencePassesSourceBackedSignal(t *testing.T) {
	eval := EvaluateSignalEvidence(SignalEvaluationInput{
		ExpectedEntityName:   "同花顺",
		ExpectedStockCode:    "300033",
		EvidenceEntityName:   "同花顺",
		EvidenceStockCode:    "SZ300033",
		AdapterStatus:        astockcontracts.AdapterStatusOK,
		FailureCode:          astockcontracts.FailureCodeNone,
		AnswerReady:          true,
		RequestedSignalTypes: []string{"fund_flow", "dragon_tiger_board"},
		ReturnedSignalTypes:  []string{"fund_flow", "dragon_tiger_board"},
		TradeDate:            "2026-05-15",
		SourceURLs: []string{
			"https://data.eastmoney.com/stock/lhb.html",
			"https://push2his.eastmoney.com/api/qt/stock/fflow/daykline/get",
		},
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

func TestEvaluateSignalEvidenceRejectsMissingSource(t *testing.T) {
	eval := EvaluateSignalEvidence(SignalEvaluationInput{
		AdapterStatus:        astockcontracts.AdapterStatusOK,
		FailureCode:          astockcontracts.FailureCodeNone,
		AnswerReady:          true,
		RequestedSignalTypes: []string{"hot_reason"},
		ReturnedSignalTypes:  []string{"hot_reason"},
		AsOf:                 "2026-05-15T15:00:00+08:00",
	})

	if eval.Passed || eval.SourceAccepted {
		t.Fatalf("expected source guard to fail, got %#v", eval)
	}
	if !strings.Contains(eval.FailureReason, "source_unaccepted") {
		t.Fatalf("expected source failure reason, got %q", eval.FailureReason)
	}
}

func TestEvaluateSignalEvidenceRejectsAdviceWithoutBoundary(t *testing.T) {
	eval := EvaluateSignalEvidence(SignalEvaluationInput{
		AdapterStatus:             astockcontracts.AdapterStatusOK,
		FailureCode:               astockcontracts.FailureCodeNone,
		AnswerReady:               true,
		RequestedSignalTypes:      []string{"hot_reason"},
		ReturnedSignalTypes:       []string{"hot_reason"},
		TradeDate:                 "2026-05-15",
		SourceURLs:                []string{"https://www.10jqka.com.cn/"},
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

func TestEvaluateSignalEvidenceRejectsRequestedSignalTypeGap(t *testing.T) {
	eval := EvaluateSignalEvidence(SignalEvaluationInput{
		AdapterStatus:        astockcontracts.AdapterStatusOK,
		FailureCode:          astockcontracts.FailureCodeNone,
		AnswerReady:          true,
		RequestedSignalTypes: []string{"fund_flow", "northbound_flow"},
		ReturnedSignalTypes:  []string{"fund_flow"},
		TradeDate:            "2026-05-15",
		SourceURLs:           []string{"https://push2his.eastmoney.com/api/qt/stock/fflow/daykline/get"},
	})

	if eval.Passed || eval.FieldsReady {
		t.Fatalf("expected requested signal type gap to fail, got %#v", eval)
	}
	if !strings.Contains(eval.FailureReason, "requested_signal_types_missing") {
		t.Fatalf("expected signal type gap failure reason, got %q", eval.FailureReason)
	}
}
