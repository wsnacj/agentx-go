package productshell_test

import (
	"context"
	"testing"

	"github.com/wsnacj/agentx-go/extensions/productshell"
)

func TestExperimentalPreparationConsumer(t *testing.T) {
	pipeline := productshell.NewPreparationPipeline(productshell.PreparationRuntimeFuncs{})
	result, err := pipeline.Prepare(context.Background(), "session-a", productshell.Input{
		UserMessage: "[skill:review] check this",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.UserMessage != "check this" || len(result.RequestedSkills) != 1 || result.RequestedSkills[0] != "review" {
		t.Fatalf("unexpected preparation result: %#v", result)
	}
}

func TestExternalTypedObservationHandoff(t *testing.T) {
	line := productshell.BuildHostDiagnosticOperatorLineObservation(productshell.HostDiagnosticOperatorLineObservationInput{
		Source:              "host_reader",
		Key:                 "progress.ready",
		Available:           true,
		Status:              "ready",
		OperatorDisplayLine: "kind=progress;status=ready",
	})
	if line == nil {
		t.Fatal("expected normalized operator line")
	}
	envelope := productshell.BuildHostUIHandoffEnvelopeFromOperatorLines(
		[]productshell.HostDiagnosticOperatorLineObservation{*line},
		productshell.HostUIHandoffInput{Target: productshell.HostUIHandoffTargetLog, Source: "host_agent"},
	)
	report := productshell.BuildHostUIHandoffRuntimeUseReport(productshell.HostUIHandoffRuntimeUseInput{
		Consumer:           "host_log_renderer",
		HostAdapter:        "host_agent",
		Target:             productshell.HostUIHandoffTargetLog,
		Source:             "host_agent",
		Envelope:           envelope,
		ConsumedEntryCount: len(envelope.Entries),
	})
	if !report.Ready {
		t.Fatalf("runtime use report = %#v, want ready", report)
	}
}
