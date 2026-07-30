package main

import (
	"encoding/json"
	"errors"
	"fmt"

	agentxmedia "github.com/wsnacj/agentx-go/runtime/mediaartifact"
	agentxprotocol "github.com/wsnacj/agentx-go/runtime/protocol"
	agentxtelemetry "github.com/wsnacj/agentx-go/runtime/telemetry"
	agentxsafeerror "github.com/wsnacj/agentx-go/runtime/telemetry/safeerror"
	agentxtoolerrors "github.com/wsnacj/agentx-go/runtime/toolerrors"
	agentxworkflow "github.com/wsnacj/agentx-go/runtime/workflow"
)

func canonicalEventJSON() ([]byte, error) {
	event := agentxprotocol.NormalizeRunEvent(agentxprotocol.RunEvent{
		Envelope: agentxprotocol.Envelope{
			Kind:  " tool.call.completed ",
			RunID: " run-consumer-1 ",
		},
		Status:   " Completed ",
		ToolName: " Browser_Open ",
	})
	if err := agentxprotocol.ValidateRunEvent(event); err != nil {
		return nil, err
	}
	return json.Marshal(event)
}

func canonicalSafeErrorJSON() ([]byte, error) {
	cause := errors.New("private consumer sentinel")
	wrapped := agentxsafeerror.WrapWithIdentity(
		cause,
		"operation failed",
		"consumer-error-1",
	)
	if !errors.Is(wrapped, cause) {
		return nil, fmt.Errorf("safeerror wrapper lost cause")
	}
	projection := agentxsafeerror.Project(
		wrapped,
		" Runtime Error ",
		" UPSTREAM/FAILED ",
	)
	return json.Marshal(projection)
}

func canonicalMediaArtifactJSON() ([]byte, error) {
	hasAudio := false
	return json.Marshal(agentxmedia.Descriptor{
		Source:     "nodes",
		Kind:       "video",
		Path:       ".agentx/nodes/capture.mp4",
		MIMEType:   "video/mp4",
		Format:     "mp4",
		Bytes:      4096,
		DurationMs: 2500,
		FPS:        30,
		HasAudio:   &hasAudio,
	})
}

func canonicalToolArgumentError() (*agentxtoolerrors.ToolArgumentError, error) {
	cause := errors.New("decode: top-level JSON object is required")
	err := agentxtoolerrors.NewInvalidJSONToolArgumentError(" browser ", cause)
	wrapped := fmt.Errorf("consumer wrapper: %w", err)
	if !errors.Is(wrapped, cause) {
		return nil, fmt.Errorf("tool argument error lost cause")
	}
	typed, ok := agentxtoolerrors.AsToolArgumentError(wrapped)
	if !ok {
		return nil, fmt.Errorf("tool argument error lost typed identity")
	}
	return typed, nil
}

func canonicalTelemetryJSON() ([]byte, error) {
	events := agentxtelemetry.ProjectToolEvents(agentxtelemetry.Event{
		Component: "tool",
		Name:      "tool.finish",
		Tool:      " Browser_Open ",
		Status:    "ok",
		Attrs: map[string]any{
			"duration_ms": 42,
		},
	})
	if len(events) != 1 {
		return nil, fmt.Errorf("telemetry projection count = %d, want 1", len(events))
	}
	return json.Marshal(events[0])
}

func canonicalWorkflowJSON() ([]byte, error) {
	return json.Marshal(agentxworkflow.Spec{
		ID:           "consumer-workflow",
		Version:      "1",
		PlanningMode: agentxworkflow.PlanningBounded,
		EntryNode:    "collect",
		Nodes: []agentxworkflow.NodeSpec{{
			ID:            "collect",
			Kind:          agentxworkflow.NodeCollect,
			ExecutionMode: agentxworkflow.ExecInline,
			Inputs: []agentxworkflow.BindingSpec{{
				From: "request.query",
				To:   "query",
			}},
			Outputs: []agentxworkflow.BindingSpec{{
				From: "result",
				To:   "state.report",
			}},
			Retry: agentxworkflow.RetryPolicy{
				MaxAttempts: 2,
				BackoffMs:   []int{100},
			},
			Config: map[string]any{"format": "markdown"},
		}},
		Edges: []agentxworkflow.EdgeSpec{{
			From: "collect",
			To:   "collect",
			On:   "retry",
		}},
		StateSchema: []agentxworkflow.StateSlotSpec{{
			Name:     "report",
			Type:     "string",
			Required: true,
		}},
		ArtifactSchema: []agentxworkflow.ArtifactTypeRef{{
			Type: "report",
		}},
		EvaluatorSchema: []agentxworkflow.EvaluatorRef{{
			Name: "quality",
		}},
	})
}

func main() {
	payload, err := canonicalWorkflowJSON()
	if err != nil {
		panic(err)
	}
	fmt.Println(string(payload))
}
