package publictransport_test

import (
	"context"
	"testing"

	control "github.com/wsnacj/agentx-go/runtime/controlcontract"
	publictransport "github.com/wsnacj/agentx-go/scenes/publictransport"
)

type fixtureCollector struct{}

func (fixtureCollector) CollectPublicTransportTicketEvidence(context.Context, publictransport.Request) (publictransport.Report, error) {
	return publictransport.Report{
		Status: control.VerificationSatisfied, InventoryObserved: true, TicketResultClaimed: true,
		Evidence:      []publictransport.Evidence{{EvidenceRef: "evidence:fixture", InventoryRef: "inventory:fixture", Strength: control.EvidenceAdequate}},
		InventoryRows: []publictransport.InventoryRow{{TrainNo: "D100", DepartureTime: "09:00", ArrivalTime: "11:00", SeatSummary: "second_class=3"}},
	}, nil
}

func TestExternalHostCanComposePortableCoordinator(t *testing.T) {
	request := control.RuntimeAdapterExecutionRequest{
		Status: control.HostActionReady, ReadyForHostExecution: true,
		AdapterRef: publictransport.DefaultAdapterRef, StrategyRef: publictransport.DefaultStrategyRef,
		IdempotencyRef: "idempotency:external", InputRefs: []control.DisplaySafeRef{"query:external"},
	}
	execution := publictransport.NewCoordinator(fixtureCollector{}).Execute(context.Background(), request)
	if execution.Result.Status != control.VerificationSatisfied || len(execution.Report.InventoryRows) != 1 {
		t.Fatalf("execution = %#v", execution)
	}
}
