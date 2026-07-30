package telemetry

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestJSONLSinkEmit(t *testing.T) {
	path := filepath.Join(t.TempDir(), "events.jsonl")
	sink, err := NewJSONLSink(path)
	if err != nil {
		t.Fatalf("new jsonl sink: %v", err)
	}
	if err := sink.Emit(context.Background(), Event{
		Component: "runner",
		Name:      "run.start",
		SessionID: "s-1",
	}); err != nil {
		t.Fatalf("emit event: %v", err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read jsonl: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
	if len(lines) != 1 {
		t.Fatalf("unexpected line count: %d", len(lines))
	}
	var event Event
	if err := json.Unmarshal([]byte(lines[0]), &event); err != nil {
		t.Fatalf("unmarshal event: %v", err)
	}
	if event.Name != "run.start" || event.Component != "runner" || event.SessionID != "s-1" {
		t.Fatalf("unexpected event: %#v", event)
	}
	if event.SchemaVersion != EventSchemaV1 {
		t.Fatalf("unexpected schema version: %s", event.SchemaVersion)
	}
	if event.Timestamp.IsZero() {
		t.Fatalf("expected timestamp")
	}
}

func TestNewJSONLSinkValidationAndNilEmit(t *testing.T) {
	if _, err := NewJSONLSink("   "); err == nil {
		t.Fatalf("expected blank path to be rejected")
	}
	var sink *JSONLSink
	if err := sink.Emit(context.Background(), Event{Name: "noop"}); err != nil {
		t.Fatalf("nil jsonl sink emit should be noop, got %v", err)
	}
}

func TestMultiSinkAggregatesErrors(t *testing.T) {
	sink := MultiSink{Sinks: []Sink{errorSink("a"), errorSink("b")}}
	err := sink.Emit(context.Background(), Event{Name: "x"})
	if err == nil {
		t.Fatalf("expected aggregated error")
	}
	text := err.Error()
	if !strings.Contains(text, "a") || !strings.Contains(text, "b") {
		t.Fatalf("unexpected error text: %s", text)
	}
}

func TestMultiSinkNoopAndNilSink(t *testing.T) {
	if err := (MultiSink{}).Emit(context.Background(), Event{Name: "x"}); err != nil {
		t.Fatalf("empty multi sink should be noop, got %v", err)
	}
	sink := MultiSink{Sinks: []Sink{nil, captureSink{}}}
	if err := sink.Emit(context.Background(), Event{Name: "ok"}); err != nil {
		t.Fatalf("multi sink with nil entries should skip nil sinks, got %v", err)
	}
}

func TestNormalizeEvent(t *testing.T) {
	now := time.Date(2026, time.March, 15, 12, 0, 0, 0, time.FixedZone("CST", 8*3600))
	event := normalizeEvent(Event{
		SchemaVersion: " 1 ",
		Timestamp:     now,
		Component:     " runner ",
		Name:          " run.start ",
		SessionID:     " s-1 ",
		Tool:          " browser ",
		Model:         " gpt ",
		Status:        " completed ",
		Attrs:         map[string]any{},
	})
	if event.SchemaVersion != EventSchemaV1 {
		t.Fatalf("unexpected schema version: %s", event.SchemaVersion)
	}
	if event.Timestamp.Location() != time.UTC {
		t.Fatalf("expected timestamp normalized to UTC, got %v", event.Timestamp.Location())
	}
	if event.Component != "runner" || event.Name != "run.start" || event.SessionID != "s-1" {
		t.Fatalf("unexpected normalized fields: %#v", event)
	}
	if event.Tool != "browser" || event.Model != "gpt" || event.Status != "completed" {
		t.Fatalf("unexpected normalized tool/model/status: %#v", event)
	}
	if event.Level != LevelInfo {
		t.Fatalf("expected blank level to default to info, got %s", event.Level)
	}
	if event.Attrs != nil {
		t.Fatalf("expected empty attrs to normalize to nil, got %#v", event.Attrs)
	}
}

func TestNormalizeEventDropsRawToolStartArguments(t *testing.T) {
	const secret = "tool-argument-secret-sentinel"
	event := normalizeEvent(Event{
		Component: "tool",
		Name:      "tool.start",
		Attrs: map[string]any{
			"arguments":           `{"token":"` + secret + `"}`,
			"args":                secret,
			"arguments_json":      secret,
			"effective_arguments": secret,
			"requested_arguments": secret,
			"arguments_hash":      "sha256:safe",
		},
	})
	for _, key := range []string{"arguments", "args", "arguments_json", "effective_arguments", "requested_arguments"} {
		if _, ok := event.Attrs[key]; ok {
			t.Fatalf("raw tool argument attr %q survived normalization: %#v", key, event.Attrs)
		}
	}
	if event.Attrs["arguments_hash"] != "sha256:safe" {
		t.Fatalf("expected controlled projection attrs to survive: %#v", event.Attrs)
	}
	encoded, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("marshal normalized event: %v", err)
	}
	if strings.Contains(string(encoded), secret) {
		t.Fatalf("normalized event exposed raw tool argument: %s", encoded)
	}
}

func TestNormalizeEventSchemaVersion(t *testing.T) {
	if got, want := normalizeEventSchemaVersion(""), EventSchemaV1; got != want {
		t.Fatalf("unexpected normalized schema version for blank: got=%q want=%q", got, want)
	}
	if got, want := normalizeEventSchemaVersion("1"), EventSchemaV1; got != want {
		t.Fatalf("unexpected normalized schema version for numeric: got=%q want=%q", got, want)
	}
	if got, want := normalizeEventSchemaVersion("future"), EventSchemaV1; got != want {
		t.Fatalf("unexpected normalized schema version for unknown: got=%q want=%q", got, want)
	}
}

func TestJSONLSinkEmitNormalizesPayload(t *testing.T) {
	path := filepath.Join(t.TempDir(), "events.jsonl")
	sink, err := NewJSONLSink(path)
	if err != nil {
		t.Fatalf("new jsonl sink: %v", err)
	}
	if err := sink.Emit(context.Background(), Event{
		SchemaVersion: "1",
		Component:     " runner ",
		Name:          " run.finish ",
		Level:         "",
		Attrs:         map[string]any{},
	}); err != nil {
		t.Fatalf("emit event: %v", err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read jsonl: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
	if len(lines) != 1 {
		t.Fatalf("unexpected line count: %d", len(lines))
	}
	var event map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &event); err != nil {
		t.Fatalf("unmarshal event: %v", err)
	}
	if event["schema_version"] != EventSchemaV1 || event["component"] != "runner" || event["name"] != "run.finish" || event["level"] != string(LevelInfo) {
		t.Fatalf("unexpected normalized json payload: %#v", event)
	}
	if _, ok := event["attrs"]; ok {
		t.Fatalf("expected empty attrs to be omitted from json payload, got %#v", event)
	}
}

func TestErrorSinkImplementsError(t *testing.T) {
	if got, want := errText("boom").Error(), "boom"; got != want {
		t.Fatalf("unexpected errText Error output: got=%q want=%q", got, want)
	}
	if got, want := reflect.TypeOf(errorSink("x")).String(), "telemetry.errorSink"; got != want {
		t.Fatalf("unexpected helper type string: got=%q want=%q", got, want)
	}
}

type errorSink string

func (e errorSink) Emit(context.Context, Event) error {
	return errText(string(e))
}

type errText string

func (e errText) Error() string {
	return string(e)
}

type captureSink struct{}

func (captureSink) Emit(context.Context, Event) error { return nil }
