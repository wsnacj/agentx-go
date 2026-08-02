package hostkit

import (
	"context"
	"encoding/json"
	"fmt"

	types "github.com/wsnacj/agentx-go/components/llm"
	agentxtools "github.com/wsnacj/agentx-go/tools"
)

const (
	ToolDocparseSpecSelect    = "docparse_spec_select"
	ToolDocparseProfileProbe  = "docparse_profile_probe"
	ToolDocparseExtractFields = "docparse_extract_fields"
	ToolDocparseExtractTable  = "docparse_extract_table"
	ToolDocparseTraceEvidence = "docparse_trace_evidence"
	ToolDocparseValidate      = "docparse_validate"
	ToolDocparseGuard         = "docparse_guard"

	RuntimeToolDocumentParse         = "document_parse"
	RuntimeToolDocumentSpecRecommend = "document_spec_recommend"
)

type ToolPayloadHandler func(context.Context, map[string]any) (any, error)

type ToolHandlers struct {
	SpecSelect    ToolPayloadHandler
	ProfileProbe  ToolPayloadHandler
	ExtractFields ToolPayloadHandler
	ExtractTable  ToolPayloadHandler
	TraceEvidence ToolPayloadHandler
	Validate      ToolPayloadHandler
	Guard         ToolPayloadHandler
}

func ToolNames() []string {
	return []string{
		ToolDocparseSpecSelect,
		ToolDocparseProfileProbe,
		ToolDocparseExtractFields,
		ToolDocparseExtractTable,
		ToolDocparseTraceEvidence,
		ToolDocparseValidate,
		ToolDocparseGuard,
	}
}

func RegisterTools(reg *agentxtools.Registry, handlers ToolHandlers) {
	registerPayloadTool(reg, DocparseSpecSelectTool(), handlers.SpecSelect)
	registerPayloadTool(reg, DocparseProfileProbeTool(), handlers.ProfileProbe)
	registerPayloadTool(reg, DocparseExtractFieldsTool(), handlers.ExtractFields)
	registerPayloadTool(reg, DocparseExtractTableTool(), handlers.ExtractTable)
	registerPayloadTool(reg, DocparseTraceEvidenceTool(), handlers.TraceEvidence)
	registerPayloadTool(reg, DocparseValidateTool(), handlers.Validate)
	registerPayloadTool(reg, DocparseGuardTool(), handlers.Guard)
}

func DecodeToolArguments(raw string) (map[string]any, error) {
	out := map[string]any{}
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, fmt.Errorf("decode docparse tool arguments: %w", err)
	}
	return out, nil
}

func registerPayloadTool(reg *agentxtools.Registry, tool types.Tool, handler ToolPayloadHandler) {
	if reg == nil || handler == nil {
		return
	}
	reg.Register(tool, func(ctx context.Context, call types.FunctionCall) (string, error) {
		params, err := DecodeToolArguments(call.Arguments)
		if err != nil {
			return "", err
		}
		payload, err := handler(ctx, params)
		if err != nil {
			return "", err
		}
		switch typed := payload.(type) {
		case string:
			return typed, nil
		case []byte:
			return string(typed), nil
		default:
			blob, err := json.Marshal(payload)
			if err != nil {
				return "", err
			}
			return string(blob), nil
		}
	})
}

func DocparseSpecSelectTool() types.Tool {
	return functionTool(
		ToolDocparseSpecSelect,
		"Rank caller-provided docparse spec candidates for a host-approved local document. Host adapters own private spec registries and document roots.",
		map[string]any{
			"user_message": map[string]any{"type": "string", "description": "Original user request or short task note used only for ranking context."},
			"document_path": map[string]any{
				"type":        "string",
				"description": "Workspace-local or host-approved document artifact reference.",
			},
			"spec_paths": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": "Caller-provided workspace-local candidate specs.",
			},
			"spec_path":       map[string]any{"type": "string", "description": "Single candidate spec path when the caller wants to rank or validate one spec."},
			"page_limit":      map[string]any{"type": "integer", "minimum": 1, "description": "Maximum number of leading pages to inspect for spec ranking."},
			"artifact_policy": map[string]any{"type": "string", "description": "Host artifact policy hint such as summary, full, or none."},
		},
		[]string{"document_path"},
	)
}

func DocparseExtractFieldsTool() types.Tool {
	return functionTool(
		ToolDocparseExtractFields,
		"Extract structured fields from a host-approved local document or parse result fixture using caller-provided specs. The scene does not own private schemas.",
		docparseCommonProperties(),
		nil,
	)
}

func DocparseExtractTableTool() types.Tool {
	return functionTool(
		ToolDocparseExtractTable,
		"Extract table evidence from a host-approved local document or parse result fixture using caller-provided specs.",
		docparseCommonProperties(),
		nil,
	)
}

func DocparseProfileProbeTool() types.Tool {
	return functionTool(
		ToolDocparseProfileProbe,
		"Classify a host-approved local document or parse result into a review-required profile proposal without extracting final fields or table rows.",
		docparseProfileProbeProperties(),
		nil,
	)
}

func DocparseTraceEvidenceTool() types.Tool {
	return functionTool(
		ToolDocparseTraceEvidence,
		"Project parsed field/table output into auditable evidence trace metadata such as page refs, bbox refs, table cells, and source snippets.",
		docparseResultProperties(),
		nil,
	)
}

func DocparseValidateTool() types.Tool {
	return functionTool(
		ToolDocparseValidate,
		"Validate extracted document fields and tables against caller-provided expected values, unit/period requirements, and evidence completeness rules.",
		docparseResultProperties(),
		nil,
	)
}

func DocparseGuardTool() types.Tool {
	return functionTool(
		ToolDocparseGuard,
		"Convert document extraction diagnostics into readiness, review_required, failure_class, and answer-boundary status.",
		docparseResultProperties(),
		nil,
	)
}

func docparseCommonProperties() map[string]any {
	props := docparseResultProperties()
	props["document_path"] = map[string]any{"type": "string", "description": "Workspace-local or host-approved document artifact reference to parse."}
	props["spec_path"] = map[string]any{"type": "string", "description": "Workspace-local docparse spec directory or main.yaml path selected by caller or spec selection."}
	props["spec_paths"] = map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Candidate spec paths for host-side selection before extraction."}
	props["output_dir"] = map[string]any{"type": "string", "description": "Optional workspace-local artifact directory for parser outputs."}
	props["page_limit"] = map[string]any{"type": "integer", "minimum": 1, "description": "Maximum number of pages to parse or inspect for this request."}
	props["page_range"] = map[string]any{"type": "string", "description": "Optional 1-based contiguous page range such as 2 or 2-4 when a host-approved bundled document should be parsed in slices."}
	props["max_chunk_chars"] = map[string]any{"type": "integer", "minimum": 200, "description": "Optional per-section text budget override for parser/model extraction."}
	props["pdf_parse_mode"] = map[string]any{"type": "string", "description": "PDF extraction mode hint, for example default, simple, fast, normal, ocr, or force_ocr."}
	props["extraction_mode"] = map[string]any{"type": "string", "description": "Document representation mode hint, for example default, legacy, table_first, text_layer_first, ocr_first, or auto."}
	props["model"] = map[string]any{"type": "string", "description": "Optional model name for specs or adapters that perform model-assisted extraction."}
	props["artifact_policy"] = map[string]any{"type": "string", "description": "Artifact retention hint such as summary, full, or none; host remains the policy owner."}
	props["include_full_result"] = map[string]any{"type": "boolean", "description": "Whether the parser response should include the full raw parse result when supported."}
	return props
}

func docparseProfileProbeProperties() map[string]any {
	return map[string]any{
		"user_message":                    map[string]any{"type": "string", "description": "Original user request or short task note used for profile-probe context."},
		"document_path":                   map[string]any{"type": "string", "description": "Workspace-local or host-approved document artifact reference to inspect."},
		"result_path":                     map[string]any{"type": "string", "description": "Workspace-local JSON profile-probe or parser result produced by document_parse or a host parser adapter."},
		"parse_result":                    map[string]any{"type": "object", "additionalProperties": true, "description": "Inline parser or profile-probe result when the caller already has host-produced output."},
		"document_type":                   map[string]any{"type": "string", "description": "Caller-provided document type hint; unknown values keep proposal-only routing."},
		"profile_id":                      map[string]any{"type": "string", "description": "Optional host profile identifier; when present the host may skip public discovery."},
		"page_limit":                      map[string]any{"type": "integer", "minimum": 1, "description": "Maximum number of pages to inspect for profile probing."},
		"page_range":                      map[string]any{"type": "string", "description": "Optional 1-based contiguous page range such as 2 or 2-4 when probing part of a bundled document."},
		"profile_probe_only":              map[string]any{"type": "boolean", "description": "Force classify-only profile probing; do not extract final fields or table rows."},
		"classify_only":                   map[string]any{"type": "boolean", "description": "Alias for profile_probe_only when callers ask only for document type classification."},
		"profile_discovery_search":        map[string]any{"type": "boolean", "description": "Allow host-owned public discovery search when the host explicitly configured a search provider."},
		"enable_profile_discovery_search": map[string]any{"type": "boolean", "description": "Alias for profile_discovery_search used by host playbooks."},
		"disable_profile_discovery_search": map[string]any{
			"type":        "boolean",
			"description": "Disable host-owned public discovery search for this profile probe request.",
		},
		"artifact_policy": map[string]any{"type": "string", "description": "Artifact retention hint such as summary, full, or none; host remains the policy owner."},
	}
}

func docparseResultProperties() map[string]any {
	return map[string]any{
		"user_message":                map[string]any{"type": "string", "description": "Original user request or short task note used for evidence projection context."},
		"result_path":                 map[string]any{"type": "string", "description": "Workspace-local JSON result produced by document_parse or a host parser adapter."},
		"parse_result":                map[string]any{"type": "object", "additionalProperties": true, "description": "Inline parse result object when the caller already has parser output and wants projection without re-parsing."},
		"document_type":               map[string]any{"type": "string", "description": "Caller-provided or host-detected document type label."},
		"expected_document_type":      map[string]any{"type": "string", "description": "Expected document type used to validate classification or profile routing."},
		"profile_id":                  map[string]any{"type": "string", "description": "Optional host/domain profile identifier used for extraction and validation rules."},
		"expected_fields":             map[string]any{"type": "array", "items": map[string]any{"type": "object", "additionalProperties": true}, "description": "Expected field assertions, accepted values, units, periods, or evidence requirements."},
		"expected_tables":             map[string]any{"type": "array", "items": map[string]any{"type": "object", "additionalProperties": true}, "description": "Expected table assertions such as table names, columns, rows, or required cells."},
		"required_fields":             map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Field names that must be present in the projected extraction result."},
		"require_page_refs":           map[string]any{"type": "boolean", "description": "Require page references for extracted fields and tables when evidence is available."},
		"require_bounding_boxes":      map[string]any{"type": "boolean", "description": "Require bounding-box evidence for extracted fields and table cells when available."},
		"require_bbox":                map[string]any{"type": "boolean", "description": "Alias for require_bounding_boxes kept for caller compatibility."},
		"require_table_cells":         map[string]any{"type": "boolean", "description": "Require table cell provenance for table extraction results when available."},
		"require_complete_table_rows": map[string]any{"type": "boolean", "description": "Opt in to projecting every parsed table row and require rows to match row_count and declared columns."},
		"allow_review_required":       map[string]any{"type": "boolean", "description": "Allow a review_required result instead of treating incomplete evidence as a hard failure."},
		"required_evidence": map[string]any{
			"type":                 "object",
			"additionalProperties": true,
			"description":          "Structured evidence requirements such as page refs, bbox refs, snippets, table cells, or field-level provenance.",
		},
		"output_schema": map[string]any{"type": "object", "additionalProperties": true, "description": "Optional caller-provided schema for shaping the final extraction payload."},
	}
}

func functionTool(name string, description string, properties map[string]any, required []string) types.Tool {
	params := map[string]any{
		"type":       "object",
		"properties": properties,
	}
	if len(required) != 0 {
		params["required"] = append([]string(nil), required...)
	}
	return types.Tool{
		Type: "function",
		Function: types.Function{
			Name:        name,
			Description: description,
			Parameters:  params,
		},
	}
}
