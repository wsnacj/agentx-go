package budget

import "testing"

func TestExceedsHelperCoversRemainingReasonsAndBoundaries(t *testing.T) {
	tests := []struct {
		name   string
		limit  Limit
		snap   Snapshot
		reason string
		hard   bool
	}{
		{
			name:   "input tokens exceeded",
			limit:  Limit{MaxInputTokens: 100},
			snap:   Snapshot{InputTokens: 101},
			reason: ReasonMaxInputTokens,
			hard:   true,
		},
		{
			name:   "output tokens exceeded",
			limit:  Limit{MaxOutputTokens: 200},
			snap:   Snapshot{OutputTokens: 201},
			reason: ReasonMaxOutputTokens,
			hard:   true,
		},
		{
			name:   "cost exceeded",
			limit:  Limit{MaxCostMicrosUSD: 300},
			snap:   Snapshot{CostMicrosUSD: 301},
			reason: ReasonMaxCostMicros,
			hard:   true,
		},
		{
			name:   "equal to limit does not exceed",
			limit:  Limit{MaxToolCalls: 10, MaxDurationMs: 1000},
			snap:   Snapshot{ToolCalls: 10, DurationMs: 1000},
			reason: "",
			hard:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reason, hard := exceeds(tc.limit, tc.snap)
			if reason != tc.reason || hard != tc.hard {
				t.Fatalf("exceeds(%#v, %#v) = (%q, %v), want (%q, %v)", tc.limit, tc.snap, reason, hard, tc.reason, tc.hard)
			}
		})
	}
}

func TestNearLimitWarningsAndWarnRatioFallback(t *testing.T) {
	limit := Limit{
		MaxToolCalls:  10,
		MaxDurationMs: 1000,
	}
	snap := Snapshot{
		ToolCalls:  8,
		DurationMs: 900,
	}

	warnings := nearLimitWarnings(limit, snap, 0.8)
	if len(warnings) != 2 {
		t.Fatalf("expected both tool and duration warnings, got %#v", warnings)
	}

	if warnings := nearLimitWarnings(limit, snap, 0.95); len(warnings) != 0 {
		t.Fatalf("expected stricter ratio to suppress warnings, got %#v", warnings)
	}

	ctrl := Controller{warnRatio: 0}
	verdict := ctrl.Check(limit, Snapshot{ToolCalls: 8})
	if !verdict.Allowed || verdict.Stage != StageWarn || len(verdict.Warnings) == 0 {
		t.Fatalf("expected invalid warnRatio to fall back to default warn behavior, got %#v", verdict)
	}

	ctrl = Controller{warnRatio: 2}
	verdict = ctrl.Check(limit, Snapshot{ToolCalls: 8})
	if !verdict.Allowed || verdict.Stage != StageWarn || len(verdict.Warnings) == 0 {
		t.Fatalf("expected warnRatio >1 to fall back to default warn behavior, got %#v", verdict)
	}
}
