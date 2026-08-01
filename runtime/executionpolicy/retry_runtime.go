package executionpolicy

import "strings"

const (
	RetryRuntimeExecutorToolLoop = "tool_loop"
	RetryRuntimeStatusBlocked    = "blocked"
	RetryRuntimeStatusPreflight  = "preflight_ready"
)

var retryRuntimeRequiredHostInputs = []string{
	"host:immutable_run_input_snapshot",
	"host:attempt_ledger_ref",
	"host:retry_budget_used",
	"host:idempotency_or_dedupe_policy",
	"host:approval_or_policy_confirmation",
}

// BudgetUsage is the aggregate usage already consumed by all attempts of one
// logical run. It is deliberately separate from diagnostics so execution can
// clamp the next attempt without treating a projection as policy.
type BudgetUsage struct {
	ToolCalls     int
	DurationMs    int64
	InputTokens   int64
	OutputTokens  int64
	CostMicrosUSD int64
}

// RetryRuntimeCommandInput contains execution-owned facts needed to authorize
// one bounded same-strategy retry. Callers provide observations; this package
// decides whether a command may be issued.
type RetryRuntimeCommandInput struct {
	Enabled         bool
	RuntimeEnabled  bool
	ExecutorMode    string
	Verification    string
	MaxRunRetries   int
	RetryBudgetUsed int
	Budget          BudgetPolicy
	Usage           BudgetUsage
	ConfirmedInputs []string
}

// RetryRuntimeCommand is the only executable result of retry preflight. It is
// intentionally smaller than the diagnostics projection and contains the
// remaining aggregate budget for the next dispatch.
type RetryRuntimeCommand struct {
	Allowed               bool
	Status                string
	BlockedReason         string
	ExecutorMode          string
	AttemptIncrement      int
	RetryBudgetUsedBefore int
	RetryBudgetUsedAfter  int
	RetryBudgetRemaining  int
	RemainingBudget       BudgetPolicy
	RequiredInputs        []string
	ConfirmedInputs       []string
	MissingInputs         []string
}

// CompileRetryRuntimeCommand applies the execution contract to a retry
// candidate. A command is never allowed when the aggregate budget is already
// exhausted or when dedupe is not enabled for the new attempt.
func CompileRetryRuntimeCommand(input RetryRuntimeCommandInput) RetryRuntimeCommand {
	input.ExecutorMode = normalizeRetryRuntimeExecutorMode(input.ExecutorMode)
	command := RetryRuntimeCommand{
		Status:                RetryRuntimeStatusBlocked,
		ExecutorMode:          strings.TrimSpace(input.ExecutorMode),
		AttemptIncrement:      1,
		RetryBudgetUsedBefore: retryMaxInt(input.RetryBudgetUsed, 0),
		RequiredInputs:        append([]string(nil), retryRuntimeRequiredHostInputs...),
	}
	command.ConfirmedInputs = confirmedRetryRuntimeInputs(command.RequiredInputs, input.ConfirmedInputs)
	for _, required := range command.RequiredInputs {
		if !containsString(command.ConfirmedInputs, required) {
			command.MissingInputs = append(command.MissingInputs, required)
		}
	}
	command.RetryBudgetUsedAfter = command.RetryBudgetUsedBefore + command.AttemptIncrement
	if input.MaxRunRetries > command.RetryBudgetUsedBefore {
		command.RetryBudgetRemaining = input.MaxRunRetries - command.RetryBudgetUsedAfter
		if command.RetryBudgetRemaining < 0 {
			command.RetryBudgetRemaining = 0
		}
	}

	switch {
	case !input.Enabled:
		command.BlockedReason = "runtime_retry_not_requested"
	case !input.RuntimeEnabled:
		command.BlockedReason = "runtime_retry_contract_gate_missing"
	case command.ExecutorMode != RetryRuntimeExecutorToolLoop:
		command.BlockedReason = "runtime_retry_executor_not_enabled"
	case input.Verification != "partial" && input.Verification != "failed":
		command.BlockedReason = "verification_not_retryable"
	case input.MaxRunRetries <= 0:
		command.BlockedReason = "run_retry_budget_not_configured"
	case command.RetryBudgetUsedBefore >= input.MaxRunRetries:
		command.BlockedReason = "run_retry_budget_exhausted"
	case input.Budget.ToolCallDedupe != ToolCallDedupeEnabled:
		command.BlockedReason = "retry_requires_idempotency_or_dedupe"
	case budgetUsageExhausted(input.Budget, input.Usage):
		command.BlockedReason = "aggregate_budget_exhausted"
	case len(command.MissingInputs) > 0:
		command.BlockedReason = "runtime_execution_preflight_missing_inputs"
	default:
		command.Allowed = true
		command.Status = RetryRuntimeStatusPreflight
		command.BlockedReason = ""
	}
	command.RemainingBudget = RemainingBudgetAfterUsage(input.Budget, input.Usage)
	return command
}

func normalizeRetryRuntimeExecutorMode(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case RetryRuntimeExecutorToolLoop, "tool_loop_once":
		return RetryRuntimeExecutorToolLoop
	default:
		return strings.ToLower(strings.TrimSpace(raw))
	}
}

// RemainingBudgetAfterUsage narrows each aggregate limit for a follow-up
// attempt. Zero limits remain unlimited, matching the existing budget contract.
func RemainingBudgetAfterUsage(policy BudgetPolicy, usage BudgetUsage) BudgetPolicy {
	policy.MaxToolCalls = remainingInt(policy.MaxToolCalls, usage.ToolCalls)
	policy.MaxDurationMs = remainingInt64(policy.MaxDurationMs, usage.DurationMs)
	policy.MaxInputTokens = remainingInt64(policy.MaxInputTokens, usage.InputTokens)
	policy.MaxOutputTokens = remainingInt64(policy.MaxOutputTokens, usage.OutputTokens)
	policy.MaxCostMicrosUSD = remainingInt64(policy.MaxCostMicrosUSD, usage.CostMicrosUSD)
	return policy
}

func budgetUsageExhausted(policy BudgetPolicy, usage BudgetUsage) bool {
	return (policy.MaxToolCalls > 0 && usage.ToolCalls >= policy.MaxToolCalls) ||
		(policy.MaxDurationMs > 0 && usage.DurationMs >= policy.MaxDurationMs) ||
		(policy.MaxInputTokens > 0 && usage.InputTokens >= policy.MaxInputTokens) ||
		(policy.MaxOutputTokens > 0 && usage.OutputTokens >= policy.MaxOutputTokens) ||
		(policy.MaxCostMicrosUSD > 0 && usage.CostMicrosUSD >= policy.MaxCostMicrosUSD)
}

func confirmedRetryRuntimeInputs(required, confirmed []string) []string {
	var out []string
	for _, value := range required {
		if containsString(confirmed, value) {
			out = append(out, value)
		}
	}
	return out
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if strings.TrimSpace(value) == want {
			return true
		}
	}
	return false
}

func retryMaxInt(value, fallback int) int {
	if value < fallback {
		return fallback
	}
	return value
}

func remainingInt(limit, used int) int {
	if limit <= 0 {
		return limit
	}
	remaining := limit - retryMaxInt(used, 0)
	if remaining < 0 {
		return 0
	}
	return remaining
}

func remainingInt64(limit, used int64) int64 {
	if limit <= 0 {
		return limit
	}
	remaining := limit - used
	if remaining < 0 {
		return 0
	}
	return remaining
}
