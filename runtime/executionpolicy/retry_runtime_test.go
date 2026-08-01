package executionpolicy

import "testing"

func TestCompileRetryRuntimeCommandUsesIndependentRunBudget(t *testing.T) {
	command := CompileRetryRuntimeCommand(RetryRuntimeCommandInput{
		Enabled:         true,
		RuntimeEnabled:  true,
		ExecutorMode:    RetryRuntimeExecutorToolLoop,
		Verification:    "partial",
		MaxRunRetries:   1,
		Budget:          BudgetPolicy{MaxToolCalls: 4, MaxRunRetries: 1, ToolCallDedupe: ToolCallDedupeEnabled},
		Usage:           BudgetUsage{ToolCalls: 1},
		ConfirmedInputs: retryRuntimeRequiredHostInputs,
	})
	if !command.Allowed || command.Status != RetryRuntimeStatusPreflight {
		t.Fatalf("expected retry command to be ready, got %#v", command)
	}
	if command.RetryBudgetUsedAfter != 1 || command.RetryBudgetRemaining != 0 {
		t.Fatalf("unexpected run retry budget accounting: %#v", command)
	}
	if command.RemainingBudget.MaxToolCalls != 3 {
		t.Fatalf("expected aggregate tool-call budget to be narrowed, got %#v", command.RemainingBudget)
	}
}

func TestCompileRetryRuntimeCommandBlocksToolRetryBudgetReuse(t *testing.T) {
	command := CompileRetryRuntimeCommand(RetryRuntimeCommandInput{
		Enabled:         true,
		RuntimeEnabled:  true,
		ExecutorMode:    RetryRuntimeExecutorToolLoop,
		Verification:    "partial",
		MaxRunRetries:   0,
		Budget:          BudgetPolicy{MaxRetries: 3, ToolCallDedupe: ToolCallDedupeEnabled},
		ConfirmedInputs: retryRuntimeRequiredHostInputs,
	})
	if command.Allowed || command.BlockedReason != "run_retry_budget_not_configured" {
		t.Fatalf("expected independent run retry budget to be required, got %#v", command)
	}
}

func TestCompileRetryRuntimeCommandBlocksAggregateBudgetAndMissingConfirmation(t *testing.T) {
	command := CompileRetryRuntimeCommand(RetryRuntimeCommandInput{
		Enabled:         true,
		RuntimeEnabled:  true,
		ExecutorMode:    RetryRuntimeExecutorToolLoop,
		Verification:    "failed",
		MaxRunRetries:   1,
		Budget:          BudgetPolicy{MaxToolCalls: 2, ToolCallDedupe: ToolCallDedupeEnabled},
		Usage:           BudgetUsage{ToolCalls: 2},
		ConfirmedInputs: nil,
	})
	if command.Allowed || command.BlockedReason != "aggregate_budget_exhausted" {
		t.Fatalf("expected aggregate budget to win before confirmation, got %#v", command)
	}

	command = CompileRetryRuntimeCommand(RetryRuntimeCommandInput{
		Enabled:        true,
		RuntimeEnabled: true,
		ExecutorMode:   RetryRuntimeExecutorToolLoop,
		Verification:   "failed",
		MaxRunRetries:  1,
		Budget:         BudgetPolicy{MaxToolCalls: 2, ToolCallDedupe: ToolCallDedupeEnabled},
		Usage:          BudgetUsage{ToolCalls: 1},
	})
	if command.Allowed || command.BlockedReason != "runtime_execution_preflight_missing_inputs" || len(command.MissingInputs) == 0 {
		t.Fatalf("expected host confirmation to be required, got %#v", command)
	}
}
