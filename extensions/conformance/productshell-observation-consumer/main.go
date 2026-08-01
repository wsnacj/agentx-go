package main

import (
	"fmt"
	"os"

	"github.com/wsnacj/agentx-go/extensions/productshell"
)

func run() (string, error) {
	session := productshell.BuildSessionObservation(productshell.SessionObservationInput{
		SessionID:     "session-001",
		CreatedUnixMs: 1000,
		Events: []productshell.SessionEventObservationInput{
			{Role: "user", Content: "run tests"},
			{Role: "assistant", Content: "running", ToolCallCount: 1},
			{Role: "tool", Content: "ok", ToolCallID: "call-001"},
		},
		Branches: []productshell.SessionBranchObservationInput{{
			BranchID: "main", NodeExecID: "node-exec-001", NodeID: "test", Status: "completed",
			StartedAt: 1010, FinishedAt: 1020,
		}},
		Compaction: productshell.SessionCompactionObservationInput{
			Passes: 1, CompactedToolOutputs: 1,
		},
	})
	if session == nil {
		return "", fmt.Errorf("session observation is missing")
	}

	progress := productshell.BuildHostProcessProgressObservation(productshell.HostProcessProgressObservationInput{
		Source:                         "external_host_process_view",
		Available:                      true,
		Enabled:                        true,
		Status:                         "completed",
		DisplayKind:                    "progress",
		SummaryCode:                    "host_process_completed",
		DisplayLine:                    "kind=progress;status=completed;process_count=1",
		SessionKey:                     session.SessionID,
		ProcessRef:                     "process-001",
		RunID:                          "run-001",
		ProcessStatus:                  "completed",
		LastKind:                       "exit",
		ExitCode:                       0,
		ExitCodeKnown:                  true,
		ReadyForReadback:               true,
		Started:                        true,
		Terminal:                       true,
		ProcessCount:                   1,
		TerminalCount:                  1,
		HostProcessEventCount:          2,
		ViewReady:                      true,
		ProgressReady:                  true,
		ConsumesHostProcessSessionView: true,
		Boundaries: []string{
			"typed_host_process_view",
			"no_raw_tool_output_decode",
		},
		NextHostAction: "render_operator_line",
	})
	if progress == nil {
		return "", fmt.Errorf("host process observation is missing")
	}

	operatorLine := productshell.BuildHostDiagnosticOperatorLineObservation(productshell.HostDiagnosticOperatorLineObservationInput{
		Source:              "external_host_adapter",
		Key:                 "progress.completed",
		Available:           progress.Available,
		Status:              progress.Status,
		OperatorDisplayLine: progress.DisplayLine,
		Boundaries: []string{
			"typed_session_observation_consumed",
			"typed_host_process_observation_consumed",
		},
		NextHostAction: "render_log_fields",
	})
	if operatorLine == nil {
		return "", fmt.Errorf("operator-line observation is missing")
	}

	envelope := productshell.BuildHostUIHandoffEnvelopeFromOperatorLines(
		[]productshell.HostDiagnosticOperatorLineObservation{*operatorLine},
		productshell.HostUIHandoffInput{
			Target: productshell.HostUIHandoffTargetLog,
			Source: "external_host",
		},
	)
	if envelope == nil {
		return "", fmt.Errorf("host UI handoff envelope is missing")
	}
	logFields, ok := productshell.RenderHostUIHandoffLogFields(envelope.Entries[0])
	if !ok {
		return "", fmt.Errorf("display-safe log fields are missing")
	}

	conformance := productshell.BuildHostUIHandoffConsumerConformanceReport(productshell.HostUIHandoffConsumerConformanceInput{
		Consumer:       "productshell_observation_consumer",
		ExpectedTarget: productshell.HostUIHandoffTargetLog,
		ExpectedSource: "external_host",
		Envelope:       envelope,
	})
	if !conformance.Passed {
		return "", fmt.Errorf("consumer conformance blocked: %v", conformance.BlockedReasons)
	}

	runtimeUse := productshell.BuildHostUIHandoffRuntimeUseReport(productshell.HostUIHandoffRuntimeUseInput{
		Consumer:           "productshell_observation_consumer",
		HostAdapter:        "deterministic_log_adapter",
		Target:             productshell.HostUIHandoffTargetLog,
		Source:             "external_host",
		HandoffSchema:      productshell.HostUIHandoffSchemaV1,
		ConformanceSchema:  productshell.HostUIHandoffConsumerConformanceSchemaV1,
		ConformanceRefs:    []string{"consumer:test"},
		Envelope:           envelope,
		ConsumedEntryCount: len(envelope.Entries),
		Boundaries: []string{
			"host_adapter_owns_log_delivery",
			"no_raw_host_diagnostics_decode",
		},
	})
	if !runtimeUse.Ready {
		return "", fmt.Errorf("runtime-use evidence blocked: %v", runtimeUse.BlockedReasons)
	}

	return fmt.Sprintf(
		"agentx-productshell-observation-ok:%s:%s:%d:%s:%s:%s",
		session.SessionID,
		progress.Status,
		envelope.EntryCount,
		conformance.Status,
		runtimeUse.Status,
		logFields,
	), nil
}

func main() {
	output, err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(output)
}
