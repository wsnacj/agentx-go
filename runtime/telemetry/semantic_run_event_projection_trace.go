package telemetry

import (
	"encoding/json"
	"strings"
	"time"
)

const (
	SemanticRunEventProjectionTraceSchemaV1        = "semantic_run_event_projection_trace_v1"
	SemanticRunEventProjectionSourceRunstore       = "runstore_events"
	SemanticRunEventProjectionSourceTelemetryJSONL = "telemetry_jsonl"
)

type SemanticRunEventProjectionTrace struct {
	SchemaVersion       string                                  `json:"schema_version"`
	Source              string                                  `json:"source"`
	SourceEventCount    int                                     `json:"source_event_count"`
	DecodedEventCount   int                                     `json:"decoded_event_count"`
	ProjectedEventCount int                                     `json:"projected_event_count"`
	InvalidEventCount   int                                     `json:"invalid_event_count,omitempty"`
	Sources             []SemanticRunEventProjectionSourceTrace `json:"sources,omitempty"`
	Summary             SemanticRunEventSummary                 `json:"summary"`
}

type SemanticRunEventProjectionSourceTrace struct {
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

func ReplaySemanticRunEventsFromStoredRecords(records []StoredRawEventRecord) ([]SemanticRunEvent, SemanticRunEventProjectionTrace) {
	trace := SemanticRunEventProjectionTrace{
		SchemaVersion:    SemanticRunEventProjectionTraceSchemaV1,
		Source:           SemanticRunEventProjectionSourceRunstore,
		SourceEventCount: len(records),
		Sources:          make([]SemanticRunEventProjectionSourceTrace, 0, len(records)),
	}
	out := make([]SemanticRunEvent, 0, len(records))
	for _, record := range records {
		sourceTrace := SemanticRunEventProjectionSourceTrace{
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
		projected := ProjectSemanticRunEvents(event)
		for i := range projected {
			projected[i].SourceEventID = sourceTrace.EventID
			if projected[i].RunID == "" {
				projected[i].RunID = sourceTrace.RunID
			}
			if projected[i].BranchID == "" {
				projected[i].BranchID = sourceTrace.BranchID
			}
			if projected[i].NodeExecID == "" {
				projected[i].NodeExecID = sourceTrace.NodeExecID
			}
			projected[i] = normalizeSemanticRunEvent(projected[i])
		}
		sourceTrace.SourceEvent = strings.TrimSpace(event.Name)
		sourceTrace.ProjectedCount = len(projected)
		sourceTrace.ProjectedKinds = semanticRunEventKinds(projected)
		trace.DecodedEventCount++
		out = append(out, projected...)
		trace.Sources = append(trace.Sources, sourceTrace)
	}
	trace.ProjectedEventCount = len(out)
	trace.Summary = SummarizeSemanticRunEvents(out)
	if len(trace.Sources) == 0 {
		trace.Sources = nil
	}
	return out, trace
}

func semanticRunEventKinds(events []SemanticRunEvent) []string {
	if len(events) == 0 {
		return nil
	}
	out := make([]string, 0, len(events))
	for _, event := range events {
		kind := NormalizeSemanticRunEventKind(event.Kind)
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
