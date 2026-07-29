package budget

const (
	StageOK       = "ok"
	StageWarn     = "warn"
	StageSoftStop = "soft_stop"
	StageHardStop = "hard_stop"

	ReasonMaxToolCalls    = "max_tool_calls"
	ReasonMaxDurationMs   = "max_duration_ms"
	ReasonMaxInputTokens  = "max_input_tokens"
	ReasonMaxOutputTokens = "max_output_tokens"
	ReasonMaxCostMicros   = "max_cost_micros_usd"
)

type Limit struct {
	MaxToolCalls     int
	MaxDurationMs    int64
	MaxInputTokens   int64
	MaxOutputTokens  int64
	MaxCostMicrosUSD int64
}

type Snapshot struct {
	ToolCalls     int
	DurationMs    int64
	InputTokens   int64
	OutputTokens  int64
	CostMicrosUSD int64
}

type Verdict struct {
	Allowed  bool
	Stage    string
	Reason   string
	Warnings []string
}
