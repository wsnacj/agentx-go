package telemetry

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"
)

type Level string

const (
	LevelDebug    Level = "debug"
	LevelInfo     Level = "info"
	LevelWarn     Level = "warn"
	LevelError    Level = "error"
	EventSchemaV1       = "v1"
)

type Event struct {
	SchemaVersion string         `json:"schema_version"`
	Timestamp     time.Time      `json:"timestamp"`
	Component     string         `json:"component"`
	Name          string         `json:"name"`
	Level         Level          `json:"level"`
	SessionID     string         `json:"session_id,omitempty"`
	Round         int            `json:"round,omitempty"`
	Tool          string         `json:"tool,omitempty"`
	Model         string         `json:"model,omitempty"`
	Status        string         `json:"status,omitempty"`
	Attrs         map[string]any `json:"attrs,omitempty"`
}

type Sink interface {
	Emit(ctx context.Context, event Event) error
}

type MultiSink struct {
	Sinks []Sink
}

func (m MultiSink) Emit(ctx context.Context, event Event) error {
	if len(m.Sinks) == 0 {
		return nil
	}
	var errs []string
	for _, sink := range m.Sinks {
		if sink == nil {
			continue
		}
		if err := sink.Emit(ctx, event); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errors.New(strings.Join(errs, "; "))
}

type JSONLSink struct {
	mu   sync.Mutex
	path string
}

func NewJSONLSink(path string) (*JSONLSink, error) {
	absPath, err := preparePrivateJSONLPath(path, "jsonl")
	if err != nil {
		return nil, err
	}
	return &JSONLSink{path: absPath}, nil
}

func (s *JSONLSink) Emit(_ context.Context, event Event) error {
	if s == nil {
		return nil
	}
	normalized := normalizeEvent(event)
	payload, err := json.Marshal(normalized)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	file, err := openPrivateJSONLAppend(s.path)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := file.Write(append(payload, '\n')); err != nil {
		return err
	}
	return nil
}

func normalizeEvent(event Event) Event {
	out := event
	out.SchemaVersion = normalizeEventSchemaVersion(out.SchemaVersion)
	if out.Timestamp.IsZero() {
		out.Timestamp = time.Now().UTC()
	} else {
		out.Timestamp = out.Timestamp.UTC()
	}
	out.Component = strings.TrimSpace(out.Component)
	out.Name = strings.TrimSpace(out.Name)
	out.SessionID = strings.TrimSpace(out.SessionID)
	out.Tool = strings.TrimSpace(out.Tool)
	out.Model = strings.TrimSpace(out.Model)
	out.Status = strings.TrimSpace(out.Status)
	out.Attrs = normalizeEventAttrs(out.Name, out.Attrs)
	if out.Level == "" {
		out.Level = LevelInfo
	}
	if len(out.Attrs) == 0 {
		out.Attrs = nil
	}
	return out
}

func normalizeEventAttrs(name string, attrs map[string]any) map[string]any {
	if len(attrs) == 0 || strings.TrimSpace(name) != "tool.start" {
		return attrs
	}
	out := make(map[string]any, len(attrs))
	for key, value := range attrs {
		out[key] = value
	}
	for _, key := range []string{"arguments", "args", "arguments_json", "effective_arguments", "requested_arguments"} {
		delete(out, key)
	}
	return out
}

func normalizeEventSchemaVersion(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "1", EventSchemaV1:
		return EventSchemaV1
	default:
		return EventSchemaV1
	}
}
