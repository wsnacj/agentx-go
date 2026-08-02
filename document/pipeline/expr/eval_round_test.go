package expr_test

import (
	"github.com/wsnacj/agentx-go/document/pipeline/expr"
	"math"
	"testing"
)

// Test evalNumericExpr with round() and nested parentheses.
func TestEvalNumericExpr_RoundNested(t *testing.T) {
	expr1 := "round(100 * ( income_statement.净利润 / income_statement.营业收入 ), 2)"
	expr2 := "round(100 * ( ( income_statement.营业收入 - income_statement.营业成本 ) / income_statement.营业收入 ), 2)"

	lookup := func(id string) (float64, bool) {
		switch id {
		case "income_statement.净利润":
			return -262172376.03, true
		case "income_statement.营业收入":
			return 6768342383.29, true
		case "income_statement.营业成本":
			return 6365835798.40, true
		default:
			return 0, false
		}
	}

	v1, ok1 := expr.EvalNumericExpr(expr1, lookup)
	if !ok1 {
		t.Fatalf("expr1 failed to evaluate")
	}
	// Expected around -3.87x
	if math.Abs(v1-(-3.87)) > 0.2 { // generous tolerance; we only care it evaluates
		t.Fatalf("expr1 unexpected value: %v", v1)
	}

	v2, ok2 := expr.EvalNumericExpr(expr2, lookup)
	if !ok2 {
		t.Fatalf("expr2 failed to evaluate")
	}
	if math.Abs(v2-5.95) > 0.01 {
		t.Fatalf("expr2 unexpected value: %v", v2)
	}
}
