package budget

import "testing"

func TestControllerCheckWarnAndStop(t *testing.T) {
	ctrl := NewController()
	limit := Limit{
		MaxToolCalls:   10,
		MaxDurationMs:  1000,
		MaxInputTokens: 2000,
	}
	warn := ctrl.Check(limit, Snapshot{
		ToolCalls: 8,
	})
	if !warn.Allowed || warn.Stage != StageWarn {
		t.Fatalf("expected warn verdict, got %#v", warn)
	}
	stop := ctrl.Check(limit, Snapshot{
		ToolCalls: 11,
	})
	if stop.Allowed || stop.Stage != StageSoftStop || stop.Reason != ReasonMaxToolCalls {
		t.Fatalf("expected soft stop on tool calls, got %#v", stop)
	}
}

func TestControllerCheckDurationHardStop(t *testing.T) {
	ctrl := NewController()
	limit := Limit{
		MaxDurationMs: 1000,
	}
	verdict := ctrl.Check(limit, Snapshot{
		DurationMs: 1200,
	})
	if verdict.Allowed || verdict.Stage != StageHardStop || verdict.Reason != ReasonMaxDurationMs {
		t.Fatalf("expected hard stop on duration, got %#v", verdict)
	}
}

func TestControllerCheckNoLimitsAlwaysOK(t *testing.T) {
	ctrl := NewController()
	verdict := ctrl.Check(Limit{}, Snapshot{
		ToolCalls:  999,
		DurationMs: 999999,
	})
	if !verdict.Allowed || verdict.Stage != StageOK {
		t.Fatalf("expected ok without limits, got %#v", verdict)
	}
}
