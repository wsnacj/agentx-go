package executionpolicy_test

import (
	"encoding/json"
	"testing"

	executionpolicy "github.com/wsnacj/agentx-go/runtime/executionpolicy"
)

func TestSnapshotMetadataRoundTripPreservesHostMetadata(t *testing.T) {
	merged, err := executionpolicy.MergeSnapshotMetaJSON(`{"host":"kept"}`, executionpolicy.Snapshot{
		Contract:  executionpolicy.Contract{ID: "contract:m5q"},
		CreatedAt: 42,
	})
	if err != nil {
		t.Fatalf("MergeSnapshotMetaJSON(): %v", err)
	}
	var envelope map[string]any
	if err := json.Unmarshal([]byte(merged), &envelope); err != nil {
		t.Fatalf("Unmarshal(): %v", err)
	}
	if envelope["host"] != "kept" {
		t.Fatalf("Host metadata was not preserved: %#v", envelope)
	}
	loaded, ok := executionpolicy.LoadSnapshotMetaJSON(merged)
	if !ok || loaded.Contract.ID != "contract:m5q" || loaded.CreatedAt != 42 {
		t.Fatalf("LoadSnapshotMetaJSON() = %#v, %v", loaded, ok)
	}
}

func TestRetryAndExecutionLoopRemainHostOwned(t *testing.T) {
	retry := executionpolicy.CompileRetryRuntimeCommand(executionpolicy.RetryRuntimeCommandInput{
		Enabled:        true,
		RuntimeEnabled: true,
		ExecutorMode:   executionpolicy.RetryRuntimeExecutorToolLoop,
		Verification:   "failed",
		MaxRunRetries:  1,
		Budget: executionpolicy.BudgetPolicy{
			ToolCallDedupe: executionpolicy.ToolCallDedupeEnabled,
		},
	})
	if retry.Allowed || retry.BlockedReason != "runtime_execution_preflight_missing_inputs" {
		t.Fatalf("Retry command must fail closed without Host confirmations: %#v", retry)
	}

	packet := executionpolicy.NewDecisionPacket(executionpolicy.DecisionPacketInput{
		ContractID: "contract:m5q",
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
		t.Fatalf("Execution-loop report crossed the Host boundary: %#v", report)
	}
}

func TestSoftRejectionSelectsTheStrictestDecision(t *testing.T) {
	decision, ok := executionpolicy.PrimarySoftRejectionDecision([]executionpolicy.SoftRejectionDecision{
		executionpolicy.NewSoftRejectionDecision("allow", "policy", "tool", "", ""),
		executionpolicy.NewSoftRejectionDecision("reject", "policy", "tool", "unsafe_content", ""),
		executionpolicy.NewSoftRejectionDecision("deny", "policy", "tool", "host_blocked", ""),
	})
	if !ok || decision.Action != executionpolicy.SoftRejectionActionHalt || decision.Reason != "host_blocked" {
		t.Fatalf("PrimarySoftRejectionDecision() = %#v, %v", decision, ok)
	}
}
