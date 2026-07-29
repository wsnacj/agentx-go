package budget

import (
	"reflect"
	"testing"
)

func TestConstantContract(t *testing.T) {
	got := []string{
		StageOK,
		StageWarn,
		StageSoftStop,
		StageHardStop,
		ReasonMaxToolCalls,
		ReasonMaxDurationMs,
		ReasonMaxInputTokens,
		ReasonMaxOutputTokens,
		ReasonMaxCostMicros,
	}
	want := []string{
		"ok",
		"warn",
		"soft_stop",
		"hard_stop",
		"max_tool_calls",
		"max_duration_ms",
		"max_input_tokens",
		"max_output_tokens",
		"max_cost_micros_usd",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("constants = %#v, want %#v", got, want)
	}
}

func TestStructReflectContract(t *testing.T) {
	assertStructFields(t, reflect.TypeFor[Limit](), []fieldContract{
		{name: "MaxToolCalls", typ: reflect.TypeFor[int]()},
		{name: "MaxDurationMs", typ: reflect.TypeFor[int64]()},
		{name: "MaxInputTokens", typ: reflect.TypeFor[int64]()},
		{name: "MaxOutputTokens", typ: reflect.TypeFor[int64]()},
		{name: "MaxCostMicrosUSD", typ: reflect.TypeFor[int64]()},
	})
	assertStructFields(t, reflect.TypeFor[Snapshot](), []fieldContract{
		{name: "ToolCalls", typ: reflect.TypeFor[int]()},
		{name: "DurationMs", typ: reflect.TypeFor[int64]()},
		{name: "InputTokens", typ: reflect.TypeFor[int64]()},
		{name: "OutputTokens", typ: reflect.TypeFor[int64]()},
		{name: "CostMicrosUSD", typ: reflect.TypeFor[int64]()},
	})
	assertStructFields(t, reflect.TypeFor[Verdict](), []fieldContract{
		{name: "Allowed", typ: reflect.TypeFor[bool]()},
		{name: "Stage", typ: reflect.TypeFor[string]()},
		{name: "Reason", typ: reflect.TypeFor[string]()},
		{name: "Warnings", typ: reflect.TypeFor[[]string]()},
	})
}

func TestReasonPrecedenceAndStopStageContract(t *testing.T) {
	limit := Limit{
		MaxToolCalls:     1,
		MaxDurationMs:    1,
		MaxInputTokens:   1,
		MaxOutputTokens:  1,
		MaxCostMicrosUSD: 1,
	}
	allExceeded := Snapshot{
		ToolCalls:     2,
		DurationMs:    2,
		InputTokens:   2,
		OutputTokens:  2,
		CostMicrosUSD: 2,
	}
	got := NewController().Check(limit, allExceeded)
	want := Verdict{Allowed: false, Stage: StageSoftStop, Reason: ReasonMaxToolCalls}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("all-exceeded verdict = %#v, want %#v", got, want)
	}

	got = NewController().Check(Limit{MaxCostMicrosUSD: 1}, Snapshot{CostMicrosUSD: 2})
	want = Verdict{Allowed: false, Stage: StageHardStop, Reason: ReasonMaxCostMicros}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("cost verdict = %#v, want %#v", got, want)
	}
}

func TestWarningBoundaryOrderAndTextContract(t *testing.T) {
	got := (Controller{}).Check(
		Limit{MaxToolCalls: 10, MaxDurationMs: 1000},
		Snapshot{ToolCalls: 8, DurationMs: 800},
	)
	want := Verdict{
		Allowed: true,
		Stage:   StageWarn,
		Warnings: []string{
			"budget near limit (max_tool_calls): 8/10",
			"budget near limit (max_duration_ms): 800/1000ms",
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("warning verdict = %#v, want %#v", got, want)
	}

	got = NewController().Check(
		Limit{MaxToolCalls: 10, MaxDurationMs: 1000},
		Snapshot{ToolCalls: 10, DurationMs: 1000},
	)
	if !reflect.DeepEqual(got, Verdict{
		Allowed: true,
		Stage:   StageWarn,
		Warnings: []string{
			"budget near limit (max_tool_calls): 10/10",
			"budget near limit (max_duration_ms): 1000/1000ms",
		},
	}) {
		t.Fatalf("equal-limit verdict = %#v", got)
	}
}

func TestNonPositiveLimitsAndZeroVerdictShape(t *testing.T) {
	got := NewController().Check(
		Limit{
			MaxToolCalls:     -1,
			MaxDurationMs:    -1,
			MaxInputTokens:   -1,
			MaxOutputTokens:  -1,
			MaxCostMicrosUSD: -1,
		},
		Snapshot{
			ToolCalls:     1,
			DurationMs:    1,
			InputTokens:   1,
			OutputTokens:  1,
			CostMicrosUSD: 1,
		},
	)
	if !reflect.DeepEqual(got, Verdict{Allowed: true, Stage: StageOK}) {
		t.Fatalf("unlimited verdict = %#v", got)
	}
	if got.Warnings != nil {
		t.Fatalf("OK warnings = %#v, want nil", got.Warnings)
	}
}

type fieldContract struct {
	name string
	typ  reflect.Type
}

func assertStructFields(t *testing.T, typ reflect.Type, want []fieldContract) {
	t.Helper()
	if typ.NumField() != len(want) {
		t.Fatalf("%s field count = %d, want %d", typ, typ.NumField(), len(want))
	}
	for i, expected := range want {
		field := typ.Field(i)
		if field.Name != expected.name || field.Type != expected.typ {
			t.Fatalf(
				"%s field[%d] = %s %s, want %s %s",
				typ,
				i,
				field.Name,
				field.Type,
				expected.name,
				expected.typ,
			)
		}
	}
}
