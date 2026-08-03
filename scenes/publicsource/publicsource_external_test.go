package publicsource_test

import (
	"context"
	"testing"

	control "github.com/wsnacj/agentx-go/runtime/controlcontract"
	publicsource "github.com/wsnacj/agentx-go/scenes/publicsource"
)

type fixtureCollector struct{}

func (fixtureCollector) CollectPublicSourceEvidence(context.Context, publicsource.Request) (publicsource.Report, error) {
	return publicsource.Report{Status: control.VerificationSatisfied, Evidence: []publicsource.Evidence{{SourceRef: "source:fixture", EvidenceRef: "evidence:fixture", Strength: control.EvidenceAdequate}}, DisplaySummaries: []publicsource.DisplaySummary{{Title: "Fixture", Summary: "Display safe", AttestationRef: "attestation:fixture", RedactionRef: "redaction:fixture"}}}, nil
}

func TestExternalHostCanComposeCoordinator(t *testing.T) {
	request := control.RuntimeAdapterExecutionRequest{Status: control.HostActionReady, ReadyForHostExecution: true, AdapterRef: publicsource.DefaultAdapterRef, StrategyRef: publicsource.DefaultStrategyRef, IdempotencyRef: "idempotency:external", InputRefs: []control.DisplaySafeRef{"query:external"}}
	execution := publicsource.NewCoordinator(fixtureCollector{}).Execute(context.Background(), request)
	if execution.Result.Status != control.VerificationSatisfied || len(execution.Report.Evidence) != 1 {
		t.Fatalf("execution=%#v", execution)
	}
}
