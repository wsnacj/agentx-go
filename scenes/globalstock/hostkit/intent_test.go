package hostkit

import (
	"strings"
	"testing"
)

func TestIntentFromParamsMergesWorkflowDefaultsWithExtractedFields(t *testing.T) {
	intent := IntentFromParams(map[string]any{
		"default_requested_fields":  []any{"price", "pe_ttm", "pb", "market_cap"},
		"requested_fields":          []any{"pe_ttm", "pb", "market_cap", "change_pct"},
		"default_requested_outputs": []any{"comparison", "valuation_snapshot"},
		"requested_outputs":         []any{"comparison"},
	})
	if got := strings.Join(intent.RequestedFields, ","); got != "price,pe_ttm,pb,market_cap,change_pct" {
		t.Fatalf("merged requested fields = %q", got)
	}
	if got := strings.Join(intent.RequestedOutputs, ","); got != "comparison,valuation_snapshot" {
		t.Fatalf("merged requested outputs = %q", got)
	}
}

func TestIntentFromParamsKeepsOpenToolFieldScopeWithoutWorkflowDefaults(t *testing.T) {
	intent := IntentFromParams(map[string]any{
		"requested_fields":  []any{"pe_ttm"},
		"requested_outputs": []any{"valuation_snapshot"},
	})
	if got := strings.Join(intent.RequestedFields, ","); got != "pe_ttm" {
		t.Fatalf("open tool requested fields = %q", got)
	}
}
