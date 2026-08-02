package tools

import (
	"strings"
	"testing"
)

func TestBuildToolResultMiddlewareEventSmallResultObserveOnly(t *testing.T) {
	event := BuildToolResultMiddlewareEvent(ToolResultMiddlewareInput{
		ToolName:  "echo",
		SessionID: "session-1",
		RunID:     "run-1",
		Arguments: `{"message":"hello"}`,
		Output:    `{"ok":true,"message":"hello"}`,
	})
	if !event.ObservationOnly || event.Mode != ToolResultMiddlewareModeObserveOnly {
		t.Fatalf("expected observation-only event, got %#v", event)
	}
	if event.WouldTransform {
		t.Fatalf("small result should not request transform: %#v", event)
	}
	if !event.Content.JSONValid || strings.Join(event.Content.JSONTopLevelKeys, ",") != "message,ok" {
		t.Fatalf("expected JSON summary without raw values, got %#v", event.Content)
	}
	if event.Arguments.Fingerprint == "" || event.Arguments.Fingerprint == `{"message":"hello"}` {
		t.Fatalf("expected argument fingerprint, got %#v", event.Arguments)
	}
}

func TestBuildToolResultMiddlewareEventLargeExternalResult(t *testing.T) {
	event := BuildToolResultMiddlewareEvent(ToolResultMiddlewareInput{
		ToolName:         "open_page",
		Output:           strings.Repeat("line\n", 250),
		LargeResultBytes: 1024 * 1024,
		LargeResultLines: 100,
	})
	if !event.ExternalContent || !event.UntrustedContent {
		t.Fatalf("expected open_page result to preserve external/untrusted boundary, got %#v", event)
	}
	if !event.WouldTransform {
		t.Fatalf("expected large output transform observation, got %#v", event)
	}
	if !containsMiddlewareString(event.WouldTransformReasons, ToolResultMiddlewareReasonLargeResult) {
		t.Fatalf("expected large_result reason, got %#v", event.WouldTransformReasons)
	}
	if !containsMiddlewareString(event.SuggestedStrategies, ToolResultMiddlewareStrategySummarizeWithRawRef) {
		t.Fatalf("expected summarize strategy, got %#v", event.SuggestedStrategies)
	}
}

func TestBuildToolResultMiddlewareEventDetectsSensitiveAndBinaryKeys(t *testing.T) {
	event := BuildToolResultMiddlewareEvent(ToolResultMiddlewareInput{
		ToolName: "http_request",
		Output:   `{"headers":{"Authorization":"Bearer redacted"},"content_base64":"` + strings.Repeat("a", 600) + `"}`,
	})
	if !event.WouldTransform {
		t.Fatalf("expected sensitive/binary observation, got %#v", event)
	}
	if !containsMiddlewareString(event.WouldTransformReasons, ToolResultMiddlewareReasonSensitiveKeyPresent) {
		t.Fatalf("expected sensitive key reason, got %#v", event.WouldTransformReasons)
	}
	if !containsMiddlewareString(event.WouldTransformReasons, ToolResultMiddlewareReasonEmbeddedBinaryPayload) {
		t.Fatalf("expected embedded binary reason, got %#v", event.WouldTransformReasons)
	}
	attrs := AppendToolResultMiddlewareTelemetryAttrs(nil, event)
	if got := attrs["result_middleware_external_content"]; got != true {
		t.Fatalf("expected external content telemetry, got %#v", attrs)
	}
	if got := attrs["result_middleware_reasons"]; !strings.Contains(asStringForTest(got), ToolResultMiddlewareReasonSensitiveKeyPresent) {
		t.Fatalf("expected reasons telemetry, got %#v", attrs)
	}
}

func TestBuildToolResultMiddlewareEventSummarizesOutputSchema(t *testing.T) {
	event := BuildToolResultMiddlewareEvent(ToolResultMiddlewareInput{
		ToolName: "search",
		Output:   `{"query":"agentx","results":[],"extra":"diagnostic"}`,
		OutputSchema: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"query":   map[string]any{"type": "string"},
				"count":   map[string]any{"type": "integer"},
				"results": map[string]any{"type": "array"},
			},
			"required": []string{"query", "count"},
		},
	})
	if event.OutputSchema == nil || !event.OutputSchema.Present || !event.OutputSchema.Closed {
		t.Fatalf("expected output schema summary, got %#v", event.OutputSchema)
	}
	if event.OutputSchema.PropertyCount != 3 {
		t.Fatalf("expected property count, got %#v", event.OutputSchema)
	}
	if !containsMiddlewareString(event.OutputSchema.MissingRequired, "count") {
		t.Fatalf("expected missing required key, got %#v", event.OutputSchema)
	}
	if !containsMiddlewareString(event.OutputSchema.UnexpectedTopLevelKeys, "extra") {
		t.Fatalf("expected unexpected top-level key, got %#v", event.OutputSchema)
	}
	if !event.OutputSchema.Drift {
		t.Fatalf("expected schema drift diagnostic, got %#v", event.OutputSchema)
	}
	attrs := AppendToolResultMiddlewareTelemetryAttrs(nil, event)
	if attrs["result_middleware_output_schema_present"] != true ||
		attrs["result_middleware_output_schema_closed"] != true ||
		attrs["result_middleware_output_schema_property_count"] != 3 ||
		attrs["result_middleware_output_schema_drift"] != true {
		t.Fatalf("expected output schema telemetry attrs, got %#v", attrs)
	}
	if got := asStringForTest(attrs["result_middleware_output_schema_missing_required"]); !strings.Contains(got, "count") {
		t.Fatalf("expected missing required attr, got %#v", attrs)
	}
}

func TestBuildToolResultMiddlewareEventProjectsTerminalObservation(t *testing.T) {
	event := BuildToolResultMiddlewareEvent(ToolResultMiddlewareInput{
		ToolName: "exec",
		Output: `{
			"command":"go test ./core/agentx/tools",
			"workdir":"/repo",
			"exit_code":0,
			"stdout":"ok",
			"stderr":"",
			"duration_ms":12,
			"truncated":false,
			"terminal_observation":{
				"kind":"terminal_command",
				"visibility":"visible_command",
				"surface":"runtime_exec",
				"action":"exec",
				"display_command":"go test ./core/agentx/tools",
				"user_visible":true,
				"reason":"explicit_exec_command"
			}
		}`,
	})
	if event.TerminalObservation == nil ||
		!event.TerminalObservation.Present ||
		event.TerminalObservation.Visibility != "visible_command" ||
		!event.TerminalObservation.UserVisible {
		t.Fatalf("expected terminal observation summary, got %#v", event.TerminalObservation)
	}
	attrs := AppendToolResultMiddlewareTelemetryAttrs(nil, event)
	if attrs["terminal_observation_present"] != true ||
		attrs["terminal_observation_kind"] != "terminal_command" ||
		attrs["terminal_observation_visibility"] != "visible_command" ||
		attrs["terminal_observation_surface"] != "runtime_exec" ||
		attrs["terminal_observation_user_visible"] != true {
		t.Fatalf("expected terminal observation telemetry attrs, got %#v", attrs)
	}
}

func TestBuildToolResultMiddlewareEventPreservesErrors(t *testing.T) {
	event := BuildToolResultMiddlewareEvent(ToolResultMiddlewareInput{
		ToolName: "exec",
		Output:   strings.Repeat("stderr\n", 500),
		IsError:  true,
	})
	if !event.IsError || !event.UntrustedContent {
		t.Fatalf("expected error/untrusted flags, got %#v", event)
	}
	if event.WouldTransform {
		t.Fatalf("error result should not be transformed in observer phase, got %#v", event)
	}
	if !containsMiddlewareString(event.Details, "error_result_preserved") {
		t.Fatalf("expected error preserved detail, got %#v", event.Details)
	}
	attrs := AppendToolResultMiddlewareTelemetryAttrs(nil, event)
	if got := attrs["result_middleware_error_preserved"]; got != true {
		t.Fatalf("expected error preserved telemetry, got %#v", attrs)
	}
}

func TestBuildToolResultMiddlewareEventRuntimeGuardTruncation(t *testing.T) {
	event := BuildToolResultMiddlewareEvent(ToolResultMiddlewareInput{
		ToolName:            "pdf_extract",
		Output:              "kept",
		OutputOriginalBytes: 4096,
		OutputKeptBytes:     4,
		OutputTruncated:     true,
		OutputGuardStrategy: "head_tail",
	})
	if !event.WouldTransform {
		t.Fatalf("expected guard truncation observation, got %#v", event)
	}
	if !containsMiddlewareString(event.WouldTransformReasons, ToolResultMiddlewareReasonRuntimeGuardTruncated) {
		t.Fatalf("expected runtime guard reason, got %#v", event.WouldTransformReasons)
	}
	if event.Content.OriginalBytes != 4096 || event.Content.KeptBytes != 4 || event.Content.GuardStrategy != "head_tail" {
		t.Fatalf("expected guard metadata projected, got %#v", event.Content)
	}
}

func TestBuildControlledToolResultTransformRequiresRawRefForLargeResult(t *testing.T) {
	event := BuildToolResultMiddlewareEvent(ToolResultMiddlewareInput{
		ToolName:         "open_page",
		Output:           strings.Repeat("line\n", 250),
		LargeResultBytes: 1024 * 1024,
		LargeResultLines: 100,
	})
	result := BuildControlledToolResultTransform(ToolResultTransformInput{
		Event:  event,
		Output: strings.Repeat("line\n", 250),
	})
	if result.Applied {
		t.Fatalf("large result transform should require raw ref or artifact ref, got %#v", result)
	}
	if !containsMiddlewareString(result.Details, ToolResultTransformReasonMissingRawResultRef) {
		t.Fatalf("expected missing raw ref detail, got %#v", result.Details)
	}
}

func TestBuildControlledToolResultTransformSummarizesWithRawRef(t *testing.T) {
	event := BuildToolResultMiddlewareEvent(ToolResultMiddlewareInput{
		ToolName:         "open_page",
		Output:           strings.Repeat("line\n", 250),
		RawResultRef:     "artifact://tool/raw/1",
		LargeResultBytes: 1024 * 1024,
		LargeResultLines: 100,
		OutputSchema: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"text": map[string]any{"type": "string"},
			},
		},
	})
	result := BuildControlledToolResultTransform(ToolResultTransformInput{
		Event:  event,
		Output: strings.Repeat("line\n", 250),
	})
	if !result.Applied || result.RawResultRef != "artifact://tool/raw/1" {
		t.Fatalf("expected controlled transform with raw ref, got %#v", result)
	}
	if !strings.Contains(result.Output, "tool_result_summary") ||
		strings.Contains(result.Output, strings.Repeat("line\n", 20)) {
		t.Fatalf("expected summary envelope without raw content, got %s", result.Output)
	}
	if result.Summary == nil || result.Summary.OutputSchema == nil || !result.Summary.OutputSchema.Present {
		t.Fatalf("expected summary envelope to carry output schema diagnostics, got %#v", result.Summary)
	}
	attrs := AppendToolResultTransformTelemetryAttrs(nil, result)
	if attrs["result_middleware_applied"] != true ||
		attrs["result_middleware_mode"] != ToolResultMiddlewareModeControlledTransform {
		t.Fatalf("expected transform telemetry attrs, got %#v", attrs)
	}
}

func TestBuildControlledToolResultTransformPreservesErrors(t *testing.T) {
	event := BuildToolResultMiddlewareEvent(ToolResultMiddlewareInput{
		ToolName: "exec",
		Output:   strings.Repeat("stderr\n", 500),
		IsError:  true,
	})
	result := BuildControlledToolResultTransform(ToolResultTransformInput{
		Event:  event,
		Output: event.Content.GuardStrategy + strings.Repeat("stderr\n", 500),
	})
	if result.Applied || !result.ErrorPreserved {
		t.Fatalf("error transform should preserve output, got %#v", result)
	}
	if !containsMiddlewareString(result.Details, "error_result_preserved") {
		t.Fatalf("expected error preserved detail, got %#v", result.Details)
	}
}

func containsMiddlewareString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func asStringForTest(value any) string {
	if s, ok := value.(string); ok {
		return s
	}
	return ""
}
