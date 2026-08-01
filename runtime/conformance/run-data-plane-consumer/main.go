package main

import (
	"context"
	"errors"
	"fmt"

	artifact "github.com/wsnacj/agentx-go/runtime/artifact"
	runstore "github.com/wsnacj/agentx-go/runtime/runstore"
)

func runDataPlaneConsumer(ctx context.Context) (string, error) {
	runs := runstore.NewMemoryStore()
	artifacts := artifact.NewMemoryRegistry()

	if err := runs.CreateRun(ctx, runstore.Run{
		RunID:     " run-consumer-1 ",
		Status:    "running",
		StartedAt: 1,
	}); err != nil {
		return "", err
	}
	if err := runs.CreateRun(ctx, runstore.Run{RunID: "run-consumer-1"}); !errors.Is(err, runstore.ErrAlreadyExists) {
		return "", fmt.Errorf("duplicate run error = %v", err)
	}
	if err := runs.UpsertNodeExecution(ctx, runstore.NodeExecution{
		NodeExecID:      "node-consumer-1",
		RunID:           "run-consumer-1",
		NodeID:          "collect",
		Status:          "completed",
		TerminationJSON: `{"kind":"completed"}`,
		StartedAt:       2,
		FinishedAt:      3,
	}); err != nil {
		return "", err
	}
	for _, event := range []runstore.Event{
		{EventID: "event-2", RunID: "run-consumer-1", NodeExecID: "node-consumer-1", Name: "node.completed", CreatedAt: 3},
		{EventID: "event-1", RunID: "run-consumer-1", Name: "run.started", CreatedAt: 1},
	} {
		if err := runs.AppendEvent(ctx, event); err != nil {
			return "", err
		}
	}
	if err := artifacts.Register(ctx, artifact.Record{
		ArtifactID: "artifact-consumer-1",
		RunID:      "run-consumer-1",
		NodeExecID: "node-consumer-1",
		Kind:       "report",
		StorageRef: "host://reports/consumer-1",
		CreatedAt:  4,
	}); err != nil {
		return "", err
	}
	if err := artifacts.Link(ctx, artifact.Link{
		SourceArtifactID: "manifest-consumer-1",
		TargetArtifactID: "artifact-consumer-1",
		Relation:         "contains",
		CreatedAt:        5,
	}); err != nil {
		return "", err
	}

	nodes, err := runs.ListNodeExecutions(ctx, "run-consumer-1")
	if err != nil || len(nodes) != 1 {
		return "", fmt.Errorf("node executions = %d, %v", len(nodes), err)
	}
	events, err := runs.ListEvents(ctx, "run-consumer-1", 0)
	if err != nil || len(events) != 2 || events[0].EventID != "event-1" {
		return "", fmt.Errorf("events = %#v, %v", events, err)
	}
	records, err := artifacts.ListByRun(ctx, "run-consumer-1")
	if err != nil || len(records) != 1 {
		return "", fmt.Errorf("artifacts = %#v, %v", records, err)
	}
	links, err := artifacts.ListLinks(ctx, artifact.LinkFilter{
		ArtifactID: "manifest-consumer-1",
		Direction:  "outbound",
	})
	if err != nil || len(links) != 1 {
		return "", fmt.Errorf("artifact links = %#v, %v", links, err)
	}
	projection := nodes[0].Projection()
	if projection == nil || projection.Termination == nil || projection.Termination.Kind != "completed" {
		return "", fmt.Errorf("node projection = %#v", projection)
	}
	return fmt.Sprintf(
		"agentx-run-data-plane-ok:%s:%s:%d:%d",
		events[0].EventID,
		records[0].ArtifactID,
		len(nodes),
		len(links),
	), nil
}

func main() {
	result, err := runDataPlaneConsumer(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Println(result)
}
