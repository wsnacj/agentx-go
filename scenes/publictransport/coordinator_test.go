package publictransport

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
	err     error
}

func (c *recordingCollector) CollectPublicTransportTicketEvidence(_ context.Context, request Request) (Report, error) {
	c.calls++
	c.request = request
	return c.report, c.err
}

func readyRequest() control.RuntimeAdapterExecutionRequest {
	return control.RuntimeAdapterExecutionRequest{
		Status:                control.HostActionReady,
		ReadyForHostExecution: true,
		AdapterRef:            DefaultAdapterRef,
		StrategyRef:           DefaultStrategyRef,
		IdempotencyRef:        "idempotency:public_transport_test",
		InputRefs:             []control.DisplaySafeRef{"query:test", "route:test", "date:2026-08-03"},
		Frame:                 control.ObjectiveFrame{ID: "objective-test", SourceContext: []control.DisplaySafeRef{"source:test"}},
	}
}

func satisfiedReport() Report {
	return Report{
		Status: control.VerificationSatisfied, InventoryObserved: true, TicketResultClaimed: true,
		ObservedAt: "2026-08-02T12:00:00Z", SourceRefs: []control.DisplaySafeRef{"source:test"},
		Evidence:      []Evidence{{EvidenceRef: "evidence:test", InventoryRef: "inventory:test", SourceRef: "source:test", Strength: control.EvidenceAdequate}},
		InventoryRows: []InventoryRow{{TrainNo: "G123", DepartureTime: "08:00", ArrivalTime: "10:00", SeatSummary: "second_class=5", FareSummary: "second_class=100"}},
	}
}

func TestCoordinatorInvokesCollectorExactlyOnce(t *testing.T) {
	collector := &recordingCollector{report: satisfiedReport()}
	coordinator := NewCoordinator(collector)
	coordinator.Now = func() time.Time { return time.Date(2026, 8, 2, 12, 0, 0, 0, time.UTC) }
	execution := coordinator.Execute(context.Background(), readyRequest())
	if collector.calls != 1 {
		t.Fatalf("collector calls = %d, want 1", collector.calls)
	}
	if execution.Result.Status != control.VerificationSatisfied || !execution.Report.InventoryObserved {
		t.Fatalf("execution = %#v", execution)
	}
	if len(collector.request.RouteRefs) != 1 || collector.request.ObservedAt != "2026-08-02T12:00:00Z" {
		t.Fatalf("collector request = %#v", collector.request)
	}
}

func TestCoordinatorFailsClosedWithoutCollector(t *testing.T) {
	execution := (Coordinator{}).Execute(context.Background(), readyRequest())
	if execution.Result.FailureClass != control.FailureHostAdapterMissing || execution.Result.Status != control.VerificationBlocked {
		t.Fatalf("execution result = %#v", execution.Result)
	}
}

func TestEvaluateInventory(t *testing.T) {
	evaluation := EvaluateInventory(InventoryEvaluationInput{
		Report: satisfiedReport(), MinimumRows: 1, RequiredTrainPrefixes: []string{"G"},
		RequiredSeatTokens: []string{"second_class"}, SeatEvidenceMode: SeatEvidenceModeAvailable, RequireFareEvidence: true,
	})
	if !evaluation.Passed || len(evaluation.MatchingRows) != 1 {
		t.Fatalf("evaluation = %#v", evaluation)
	}
}

func TestDefinitionIsValid(t *testing.T) {
	definition := Definition()
	if definition.Manifest.ID != PackID || definition.Manifest.DefaultWorkflow != DefaultWorkflow {
		t.Fatalf("definition manifest = %#v", definition.Manifest)
	}
}
