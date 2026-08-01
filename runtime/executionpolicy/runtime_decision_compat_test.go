package executionpolicy

import "testing"

func TestExecutionDecisionPacketReExport(t *testing.T) {
	packet := NewDecisionPacket(DecisionPacketInput{
		ContractID: "contract-1",
		Source:     "authorization",
		RuntimeDecisions: []RuntimeDecision{
			{
				Checked:       true,
				Tool:          "exec",
				Action:        "exec",
				Allowed:       true,
				Reason:        "exec_allowed",
				PolicySource:  RuntimePolicySourceToolCallPolicy,
				ControlSource: RuntimeControlSourceExecutionContract,
			},
		},
	})
	if packet.SchemaVersion != DecisionPacketSchemaV1 ||
		packet.FinalAction != DecisionActionAllow ||
		len(packet.Steps) != 1 ||
		packet.Steps[0].Kind != DecisionStepKindSandbox ||
		packet.Audit == nil ||
		packet.Audit.ContractID != "contract-1" {
		t.Fatalf("unexpected re-exported decision packet: %#v", packet)
	}
}

func TestExecutionLoopReportReExport(t *testing.T) {
	packet := NewDecisionPacket(DecisionPacketInput{
		ContractID: "contract-loop-reexport",
		Source:     "authorization",
		RuntimeDecisions: []RuntimeDecision{
			{
				Checked:       true,
				Tool:          "exec",
				Action:        "exec",
				Allowed:       true,
				Reason:        "exec_allowed",
				PolicySource:  RuntimePolicySourceToolCallPolicy,
				ControlSource: RuntimeControlSourceExecutionContract,
			},
		},
	})

	report := BuildExecutionLoopReport(ExecutionLoopReportInput{
		Enabled:        true,
		DecisionPacket: packet,
	})
	if report.Contract.Version != ExecutionLoopReportContractVersion ||
		report.ReportKind != ExecutionLoopReportKind ||
		report.Status != ExecutionLoopStatusReadyForHostExecution ||
		!report.ReadyForHostExecution ||
		report.Contract.Owner != "execution" {
		t.Fatalf("unexpected re-exported execution loop report: %#v", report)
	}
}
