package telemetry

import (
	"encoding/json"
	"strings"
	"time"
)

const (
	ToolEventProjectionTraceSchemaV1        = "tool_event_projection_trace_v1"
	ToolEventProjectionSourceRunstore       = "runstore_events"
	ToolEventProjectionSourceTelemetryJSONL = "telemetry_jsonl"
)

type StoredRawEventRecord struct {
	EventID     string `json:"event_id,omitempty"`
	RunID       string `json:"run_id,omitempty"`
	BranchID    string `json:"branch_id,omitempty"`
	NodeExecID  string `json:"node_exec_id,omitempty"`
	Name        string `json:"name,omitempty"`
	PayloadJSON string `json:"payload_json,omitempty"`
	CreatedAt   int64  `json:"created_at,omitempty"`
}

type ToolEventProjectionTrace struct {
	SchemaVersion       string                           `json:"schema_version"`
	Source              string                           `json:"source"`
	SourceEventCount    int                              `json:"source_event_count"`
	DecodedEventCount   int                              `json:"decoded_event_count"`
	ProjectedEventCount int                              `json:"projected_event_count"`
	InvalidEventCount   int                              `json:"invalid_event_count,omitempty"`
	Sources             []ToolEventProjectionSourceTrace `json:"sources,omitempty"`
	Summary             ToolEventSummary                 `json:"summary"`
}

type ToolEventProjectionSourceTrace struct {
	EventID        string   `json:"event_id,omitempty"`
	RunID          string   `json:"run_id,omitempty"`
	BranchID       string   `json:"branch_id,omitempty"`
	NodeExecID     string   `json:"node_exec_id,omitempty"`
	StoredName     string   `json:"stored_name,omitempty"`
	SourceEvent    string   `json:"source_event,omitempty"`
	CreatedAt      int64    `json:"created_at,omitempty"`
	ProjectedCount int      `json:"projected_count"`
	ProjectedKinds []string `json:"projected_kinds,omitempty"`
	Error          string   `json:"error,omitempty"`
}

func ReplayToolEventsFromStoredRecords(records []StoredRawEventRecord) ([]ToolEvent, ToolEventProjectionTrace) {
	trace := ToolEventProjectionTrace{
		SchemaVersion:    ToolEventProjectionTraceSchemaV1,
		Source:           ToolEventProjectionSourceRunstore,
		SourceEventCount: len(records),
		Sources:          make([]ToolEventProjectionSourceTrace, 0, len(records)),
	}
	out := make([]ToolEvent, 0, len(records))
	for _, record := range records {
		sourceTrace := ToolEventProjectionSourceTrace{
			EventID:    strings.TrimSpace(record.EventID),
			RunID:      strings.TrimSpace(record.RunID),
			BranchID:   strings.TrimSpace(record.BranchID),
			NodeExecID: strings.TrimSpace(record.NodeExecID),
			StoredName: strings.TrimSpace(record.Name),
			CreatedAt:  record.CreatedAt,
		}
		raw := strings.TrimSpace(record.PayloadJSON)
		if raw == "" {
			sourceTrace.Error = "missing_payload"
			trace.InvalidEventCount++
			trace.Sources = append(trace.Sources, sourceTrace)
			continue
		}
		var event Event
		if err := json.Unmarshal([]byte(raw), &event); err != nil {
			sourceTrace.Error = "invalid_payload_json"
			trace.InvalidEventCount++
			trace.Sources = append(trace.Sources, sourceTrace)
			continue
		}
		if event.Timestamp.IsZero() && record.CreatedAt > 0 {
			event.Timestamp = time.UnixMilli(record.CreatedAt).UTC()
		}
		projected := ProjectToolEvents(event)
		for i := range projected {
			projected[i].SourceEventID = sourceTrace.EventID
			projected[i] = normalizeToolEvent(projected[i])
		}
		sourceTrace.SourceEvent = strings.TrimSpace(event.Name)
		sourceTrace.ProjectedCount = len(projected)
		sourceTrace.ProjectedKinds = toolEventKinds(projected)
		trace.DecodedEventCount++
		out = append(out, projected...)
		trace.Sources = append(trace.Sources, sourceTrace)
	}
	trace.ProjectedEventCount = len(out)
	trace.Summary = SummarizeToolEvents(out)
	if len(trace.Sources) == 0 {
		trace.Sources = nil
	}
	return out, trace
}

func toolEventKinds(events []ToolEvent) []string {
	if len(events) == 0 {
		return nil
	}
	out := make([]string, 0, len(events))
	for _, event := range events {
		kind := NormalizeToolEventKind(event.Kind)
		if kind == "" {
			continue
		}
		out = append(out, kind)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
