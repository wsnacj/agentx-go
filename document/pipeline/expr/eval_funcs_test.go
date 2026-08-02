package expr_test

import (
	"github.com/wsnacj/agentx-go/document/pipeline/expr"
	"math"
	"testing"
)

func TestEvalNumericExpr_AbsApproxBetween(t *testing.T) {
	lookup := func(id string) (float64, bool) {
		switch id {
		case "a":
			return 1, true
		case "b":
			return 2, true
		default:
			return 0, false
		}
	}

	// abs: simple and with spaces/case
	for _, tc := range []struct {
		expr string
		want float64
	}{
		{"abs(-3)", 3},
		{"AbS ( -2 )", 2},
		{"abs( -(a+b) )", 3}, // uses unary minus and variables
	} {
		got, ok := expr.EvalNumericExpr(tc.expr, lookup)
		if !ok {
			t.Fatalf("abs eval failed for %q", tc.expr)
		}
		if math.Abs(got-tc.want) > 1e-9 {
			t.Fatalf("abs wrong for %q: got %v want %v", tc.expr, got, tc.want)
		}
	}

	// approx: boundary and case/space variants
	for _, tc := range []struct {
		expr string
		want float64
	}{
		{"approx(1.0, 1.01, 0.02)", 1},     // |1-1.01| = 0.01 <= 0.02
		{"ApPrOx ( 1 , 1.03 , 0.02 )", 0},  // 0.03 > 0.02
		{"approx(abs(-(a+b)), 3, 0.1)", 1}, // abs(-(1+2))=3 ~ 3
	} {
		got, ok := expr.EvalNumericExpr(tc.expr, lookup)
		if !ok {
			t.Fatalf("approx eval failed for %q", tc.expr)
		}
		if math.Abs(got-tc.want) > 1e-9 {
			t.Fatalf("approx wrong for %q: got %v want %v", tc.expr, got, tc.want)
		}
	}

	// between: inside/outside and case/space variants
	for _, tc := range []struct {
		expr string
		want float64
	}{
		{"between(5,1,10)", 1},
		{"BeTwEeN ( 11 , 1 , 10 )", 0},
		{"between(abs(-(a+b)), 3, 3)", 1}, // 3 in [3,3]
	} {
		got, ok := expr.EvalNumericExpr(tc.expr, lookup)
		if !ok {
			t.Fatalf("between eval failed for %q", tc.expr)
		}
		if math.Abs(got-tc.want) > 1e-9 {
			t.Fatalf("between wrong for %q: got %v want %v", tc.expr, got, tc.want)
		}
	}
}

func TestNormalizeNumber_EnglishCurrencyUnits(t *testing.T) {
	cases := []struct {
		raw  string
		want float64
	}{
		{"RMB364.9 billion", 364900000000},
		{"HK$ 1,234 million", 1234000000},
		{"US$ 12.5 thousand", 12500},
		{"(23.4) RMB million", -23400000},
	}
	for _, tc := range cases {
		got, ok := expr.NormalizeNumber(tc.raw)
		if !ok {
			t.Fatalf("NormalizeNumber(%q) failed", tc.raw)
		}
		f, ok := got.(float64)
		if !ok {
			t.Fatalf("NormalizeNumber(%q) returned %T", tc.raw, got)
		}
		if math.Abs(f-tc.want) > 1e-6 {
			t.Fatalf("NormalizeNumber(%q)=%v want %v", tc.raw, f, tc.want)
		}
	}
}
