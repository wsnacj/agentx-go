package publicsource

import (
	"context"
	"testing"
	"time"

	control "github.com/wsnacj/agentx-go/runtime/controlcontract"
)

type recordingCollector struct {
	calls   int
	request Request
	report  Report
}

func (collector *recordingCollector) CollectPublicSourceEvidence(_ context.Context, request Request) (Report, error) {
	collector.calls++
	collector.request = request
	return collector.report, nil
}

func readyRequest() control.RuntimeAdapterExecutionRequest {
	return control.RuntimeAdapterExecutionRequest{Status: control.HostActionReady, ReadyForHostExecution: true, AdapterRef: DefaultAdapterRef, StrategyRef: DefaultStrategyRef, IdempotencyRef: "idempotency:public_source_test", InputRefs: []control.DisplaySafeRef{"query:test"}, Frame: control.ObjectiveFrame{ID: "objective-test", SourceContext: []control.DisplaySafeRef{"source:test"}}}
}

func satisfiedReport() Report {
	return Report{Status: control.VerificationSatisfied, ObservedAt: "2026-08-03T00:00:00Z", SourceRefs: []control.DisplaySafeRef{"source:test"}, Evidence: []Evidence{{SourceRef: "source:test", QueryRef: "query:test", EvidenceRef: "evidence:test", Strength: control.EvidenceAdequate}}, DisplaySummaries: []DisplaySummary{{SourceRef: "source:test", QueryRef: "query:test", EvidenceRef: "evidence:test", Title: "Example", Summary: "Grounded summary", AttestationRef: "attestation:test", RedactionRef: "redaction:test"}}}
}

func TestCoordinatorInvokesCollectorExactlyOnce(t *testing.T) {
	collector := &recordingCollector{report: satisfiedReport()}
	coordinator := NewCoordinator(collector)
	coordinator.Now = func() time.Time { return time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC) }
	execution := coordinator.Execute(context.Background(), readyRequest())
	if collector.calls != 1 || execution.Result.Status != control.VerificationSatisfied {
		t.Fatalf("execution=%#v calls=%d", execution, collector.calls)
	}
	if collector.request.ObservedAt != "2026-08-03T00:00:00Z" || len(collector.request.QueryRefs) != 1 {
		t.Fatalf("request=%#v", collector.request)
	}
}

func TestBuildReportFromSearchAndPolicy(t *testing.T) {
	report := BuildReportFromSearch(SearchReportInput{Payload: SearchPayload{Query: "agentx", Results: []SearchResult{{Title: "Allowed", URL: "https://docs.example.com/a", Description: "summary"}, {Title: "Blocked", URL: "https://blocked.invalid/a"}}}, DisplaySummaries: []DisplaySummary{{Title: "Allowed", Summary: "verified", AttestationRef: "attestation:test", RedactionRef: "redaction:test"}, {Title: "Blocked", Summary: "drop"}}, SourcePolicy: SourcePolicy{AllowedHosts: []string{"docs.example.com"}, RequireHTTPS: true}, ObservedAt: "2026-08-03T00:00:00Z"})
	if report.Status != control.VerificationSatisfied || len(report.Evidence) != 1 || len(report.DisplaySummaries) != 1 {
		t.Fatalf("report=%#v", report)
	}
	if !report.DisplaySummaries[0].DisplaySafeAttested {
		t.Fatalf("summary=%#v", report.DisplaySummaries[0])
	}
}

func TestEvaluateRequiresAttestation(t *testing.T) {
	evaluation := Evaluate(satisfiedReport(), true)
	if !evaluation.Passed || !evaluation.AttestedSummaryObserved {
		t.Fatalf("evaluation=%#v", evaluation)
	}
}

func TestDefinitionIsReadOnlyAndBounded(t *testing.T) {
	definition := Definition()
	if definition.Manifest.ID != PackID || len(definition.Tools) != 1 || definition.PolicyProfiles[0].Contract.Budget.MaxToolCalls != 1 {
		t.Fatalf("definition=%#v", definition)
	}
}
