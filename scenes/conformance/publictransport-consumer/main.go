package main

import (
	"context"
	"fmt"
	"os"

	control "github.com/wsnacj/agentx-go/runtime/controlcontract"
	publictransport "github.com/wsnacj/agentx-go/scenes/publictransport"
)

type fixtureCollector struct{ calls int }

func (c *fixtureCollector) CollectPublicTransportTicketEvidence(context.Context, publictransport.Request) (publictransport.Report, error) {
	c.calls++
	return publictransport.Report{
		Status: control.VerificationSatisfied, InventoryObserved: true, TicketResultClaimed: true,
		ObservedAt: "2026-08-02T12:00:00Z", SourceRefs: []control.DisplaySafeRef{"source:fixture"},
		Evidence: []publictransport.Evidence{{
			EvidenceRef: "evidence:fixture", QueryRef: "query:fixture", InventoryRef: "inventory:fixture",
			SourceRef: "source:fixture", Strength: control.EvidenceAdequate,
		}},
		InventoryRows: []publictransport.InventoryRow{{
			TrainNo: "G123", DepartureTime: "08:00", ArrivalTime: "10:00",
			SeatSummary: "second_class=5", FareSummary: "second_class=100",
		}},
	}, nil
}

func run() (string, error) {
	collector := &fixtureCollector{}
	request := control.RuntimeAdapterExecutionRequest{
		Status: control.HostActionReady, ReadyForHostExecution: true,
		AdapterRef: publictransport.DefaultAdapterRef, StrategyRef: publictransport.DefaultStrategyRef,
		IdempotencyRef: "idempotency:fixed_consumer", InputRefs: []control.DisplaySafeRef{"query:fixture", "route:fixture", "date:2026-08-03"},
		Frame: control.ObjectiveFrame{ID: "objective:fixed_consumer", SourceContext: []control.DisplaySafeRef{"source:fixture"}},
	}
	execution := publictransport.NewCoordinator(collector).Execute(context.Background(), request)
	evaluation := publictransport.EvaluateInventory(publictransport.InventoryEvaluationInput{
		Report: execution.Report, MinimumRows: 1, RequiredTrainPrefixes: []string{"G"},
		RequiredSeatTokens: []string{"second_class"}, RequireFareEvidence: true,
	})
	definition := publictransport.Definition()
	if collector.calls != 1 || execution.Result.Status != control.VerificationSatisfied || !evaluation.Passed {
		return "", fmt.Errorf("public transport fixture mismatch: calls=%d status=%s evaluation=%#v", collector.calls, execution.Result.Status, evaluation)
	}
	return fmt.Sprintf("agentx-publictransport-ok:%s:%s:%d:%t", definition.Manifest.ID, definition.Manifest.DefaultWorkflow, collector.calls, evaluation.Passed), nil
}

func main() {
	result, err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(result)
}
