package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	types "github.com/wsnacj/agentx-go/components/llm"
	"github.com/wsnacj/agentx-go/scenes/docparse/planner"
	llmxtools "github.com/wsnacj/agentx-go/tools"
)

const (
	SpecDocparseAdapterID       = "spec_docparse"
	DefaultRuntimeDocumentParse = "document_parse"
)

// SpecDocparseAdapter delegates an explicit spec route to the host/runtime
// document_parse executor and projects the JSON response for fusion.
type SpecDocparseAdapter struct {
	executor    llmxtools.Executor
	runtimeTool string
}

// NewSpecDocparseAdapter creates a spec-based adapter.
func NewSpecDocparseAdapter(executor llmxtools.Executor, runtimeTool string) *SpecDocparseAdapter {
	runtimeTool = strings.TrimSpace(runtimeTool)
	if runtimeTool == "" {
		runtimeTool = DefaultRuntimeDocumentParse
	}
	return &SpecDocparseAdapter{executor: executor, runtimeTool: runtimeTool}
}

// ID implements Adapter.
func (a *SpecDocparseAdapter) ID() string { return SpecDocparseAdapterID }

// Supports implements Adapter.
func (a *SpecDocparseAdapter) Supports(route planner.Route) bool {
	return strings.TrimSpace(route.Kind) == planner.RouteSpecDocparse
}

// Execute implements Adapter.
func (a *SpecDocparseAdapter) Execute(ctx context.Context, input Input) (Output, error) {
	if a == nil || a.executor == nil {
		return Output{}, fmt.Errorf("document_parse executor is not configured")
	}
	args := cloneParams(input.Params)
	if readString(args, "spec_path") == "" && strings.TrimSpace(input.Route.SpecPath) != "" {
		args["spec_path"] = input.Route.SpecPath
	}
	blob, err := json.Marshal(args)
	if err != nil {
		return Output{}, fmt.Errorf("encode document_parse arguments: %w", err)
	}
	raw, err := a.executor.Execute(ctx, types.FunctionCall{Name: a.runtimeTool, Arguments: string(blob)})
	if err != nil {
		return Output{}, err
	}
	payload, err := decodeJSONMap(raw)
	if err != nil {
		return Output{}, err
	}
	return Output{
		AdapterID:      a.ID(),
		RouteKind:      planner.RouteSpecDocparse,
		Status:         readString(payload, "status"),
		Payload:        payload,
		Fields:         readObjectArray(payload["fields"]),
		Tables:         readObjectArray(payload["tables"]),
		ReviewRequired: readBool(payload, "review_required") || readString(payload, "status") != "success",
		Warnings:       readStringArray(payload["warnings"]),
		Diagnostics:    adapterDiagnosticsFromPayload(payload),
	}, nil
}

func decodeJSONMap(raw string) (map[string]any, error) {
	decoder := json.NewDecoder(bytes.NewBufferString(raw))
	decoder.UseNumber()
	payload := map[string]any{}
	if err := decoder.Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode document_parse output: %w", err)
	}
	return payload, nil
}

func adapterDiagnosticsFromPayload(payload map[string]any) map[string]any {
	out := map[string]any{
		"payload_status": readString(payload, "status"),
	}
	for _, key := range []string{
		"tool",
		"document_path",
		"spec_path",
		"output_dir",
		"artifact_policy",
		"extraction_mode",
		"page_count",
		"text_quality",
		"text_source",
		"field_count",
		"chapter_count",
		"error_class",
		"error",
	} {
		if value, ok := payload[key]; ok && value != nil {
			out[key] = value
		}
	}
	if diagnostics, ok := payload["diagnostics"].(map[string]any); ok && len(diagnostics) > 0 {
		out["document_diagnostics"] = diagnostics
	}
	return out
}

func cloneParams(params map[string]any) map[string]any {
	out := make(map[string]any, len(params))
	for key, value := range params {
		if strings.TrimSpace(key) == "" {
			continue
		}
		out[key] = value
	}
	return out
}

func readObjectArray(value any) []map[string]any {
	items, ok := value.([]any)
	if !ok {
		return nil
	}
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		if object, ok := item.(map[string]any); ok {
			out = append(out, cloneParams(object))
		}
	}
	return out
}

func readStringArray(value any) []string {
	switch typed := value.(type) {
	case []any:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			text := strings.TrimSpace(fmt.Sprint(item))
			if text != "" && text != "<nil>" {
				out = append(out, text)
			}
		}
		return out
	case []string:
		return append([]string(nil), typed...)
	default:
		return nil
	}
}

func readString(payload map[string]any, key string) string {
	value, ok := payload[key]
	if !ok || value == nil {
		return ""
	}
	text := strings.TrimSpace(fmt.Sprint(value))
	if text == "<nil>" {
		return ""
	}
	return text
}

func readBool(payload map[string]any, key string) bool {
	value, ok := payload[key]
	if !ok {
		return false
	}
	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		return strings.EqualFold(strings.TrimSpace(typed), "true")
	default:
		return false
	}
}
