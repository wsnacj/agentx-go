package budget

import "fmt"

const defaultWarnRatio = 0.8

type Controller struct {
	warnRatio float64
}

func NewController() Controller {
	return Controller{warnRatio: defaultWarnRatio}
}

func (c Controller) Check(limit Limit, snap Snapshot) Verdict {
	if c.warnRatio <= 0 || c.warnRatio >= 1 {
		c.warnRatio = defaultWarnRatio
	}
	if reason, hard := exceeds(limit, snap); hard {
		stage := StageHardStop
		if reason == ReasonMaxToolCalls {
			stage = StageSoftStop
		}
		return Verdict{
			Allowed: false,
			Stage:   stage,
			Reason:  reason,
		}
	}
	warnings := nearLimitWarnings(limit, snap, c.warnRatio)
	if len(warnings) > 0 {
		return Verdict{
			Allowed:  true,
			Stage:    StageWarn,
			Warnings: warnings,
		}
	}
	return Verdict{
		Allowed: true,
		Stage:   StageOK,
	}
}

func exceeds(limit Limit, snap Snapshot) (string, bool) {
	if limit.MaxToolCalls > 0 && snap.ToolCalls > limit.MaxToolCalls {
		return ReasonMaxToolCalls, true
	}
	if limit.MaxDurationMs > 0 && snap.DurationMs > limit.MaxDurationMs {
		return ReasonMaxDurationMs, true
	}
	if limit.MaxInputTokens > 0 && snap.InputTokens > limit.MaxInputTokens {
		return ReasonMaxInputTokens, true
	}
	if limit.MaxOutputTokens > 0 && snap.OutputTokens > limit.MaxOutputTokens {
		return ReasonMaxOutputTokens, true
	}
	if limit.MaxCostMicrosUSD > 0 && snap.CostMicrosUSD > limit.MaxCostMicrosUSD {
		return ReasonMaxCostMicros, true
	}
	return "", false
}

func nearLimitWarnings(limit Limit, snap Snapshot, ratio float64) []string {
	warnings := make([]string, 0, 2)
	if limit.MaxToolCalls > 0 && float64(snap.ToolCalls) >= float64(limit.MaxToolCalls)*ratio {
		warnings = append(warnings, fmt.Sprintf("budget near limit (%s): %d/%d", ReasonMaxToolCalls, snap.ToolCalls, limit.MaxToolCalls))
	}
	if limit.MaxDurationMs > 0 && float64(snap.DurationMs) >= float64(limit.MaxDurationMs)*ratio {
		warnings = append(warnings, fmt.Sprintf("budget near limit (%s): %d/%dms", ReasonMaxDurationMs, snap.DurationMs, limit.MaxDurationMs))
	}
	return warnings
}
