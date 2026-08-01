package main

import (
	"context"
	"fmt"
	"os"
	"time"

	agentx "github.com/wsnacj/agentx-go"
	"github.com/wsnacj/agentx-go/runtime/execution"
	"github.com/wsnacj/agentx-go/runtime/executionpolicy"
)

func main() {
	kernelStatus, err := validateExecutionPolicyKernel()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	client, err := newClient()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	result, err := client.Run(context.Background(), agentx.RunRequest{
		Input:     "execution conformance",
		SessionID: "execution-conformance",
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := client.Shutdown(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("agentx-execution-ok:%s:%s:%s\n", result.Status, result.SessionID, kernelStatus)
}

func validateExecutionPolicyKernel() (string, error) {
	meta, err := executionpolicy.MergeSnapshotMetaJSON(`{"host":"kept"}`, executionpolicy.Snapshot{
		Contract: executionpolicy.Contract{ID: "contract:fixed-consumer"},
	})
	if err != nil {
		return "", fmt.Errorf("merge execution snapshot metadata: %w", err)
	}
	loaded, ok := executionpolicy.LoadSnapshotMetaJSON(meta)
	if !ok || loaded.Contract.ID != "contract:fixed-consumer" {
		return "", fmt.Errorf("execution snapshot metadata is not readable: %#v", loaded)
	}
	retry := executionpolicy.CompileRetryRuntimeCommand(executionpolicy.RetryRuntimeCommandInput{
		Enabled:         true,
		RuntimeEnabled:  true,
		ExecutorMode:    executionpolicy.RetryRuntimeExecutorToolLoop,
		Verification:    "failed",
		MaxRunRetries:   2,
		RetryBudgetUsed: 0,
		Budget: executionpolicy.BudgetPolicy{
			ToolCallDedupe: executionpolicy.ToolCallDedupeEnabled,
			MaxToolCalls:   3,
		},
		Usage: executionpolicy.BudgetUsage{ToolCalls: 1},
		ConfirmedInputs: []string{
			"host:immutable_run_input_snapshot",
			"host:attempt_ledger_ref",
			"host:retry_budget_used",
			"host:idempotency_or_dedupe_policy",
			"host:approval_or_policy_confirmation",
		},
	})
	if !retry.Allowed || retry.Status != executionpolicy.RetryRuntimeStatusPreflight || retry.RemainingBudget.MaxToolCalls != 2 {
		return "", fmt.Errorf("bounded retry command is not ready: %#v", retry)
	}
	packet := executionpolicy.NewDecisionPacket(executionpolicy.DecisionPacketInput{
		ContractID: "contract:fixed-consumer",
		RuntimeDecisions: []executionpolicy.RuntimeDecision{{
			Checked: true,
			Allowed: true,
			Tool:    "read",
		}},
	})
	report := executionpolicy.BuildExecutionLoopReport(executionpolicy.ExecutionLoopReportInput{
		Enabled:        true,
		DecisionPacket: packet,
	})
	if !report.ReadyForHostExecution || !report.NoCoreExecution || !report.NoToolInvocation || !report.NoHostAdapterInvocation {
		return "", fmt.Errorf("execution-loop report crossed the Host boundary: %#v", report)
	}
	return "execution_policy_kernel_ready", nil
}

func newClient() (*agentx.Client, error) {
	runtime, err := execution.New(conformanceHost{})
	if err != nil {
		return nil, err
	}
	return agentx.New(agentx.Config{Adapter: runtime})
}

type conformanceHost struct{}

func (conformanceHost) Run(_ context.Context, request execution.Request) (*execution.Result, error) {
	return &execution.Result{
		RunID:     "execution-conformance-run",
		SessionID: request.SessionID,
		Status:    "completed",
		Reply:     "agentx-execution-ok",
	}, nil
}

func (conformanceHost) Shutdown(context.Context) error {
	return nil
}

func (conformanceHost) ClassifyError(error) agentx.ErrorCode {
	return agentx.CodeExecutionFailed
}
