package hostkit

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/wsnacj/agentx-go/components/llm"
	agentxtools "github.com/wsnacj/agentx-go/tools"
)

type ToolPayloadHandler func(context.Context, map[string]any) (any, error)

// ArgumentDecoder decodes one model-emitted function argument object.
// Hosts may inject a compatibility decoder without moving repair policy into this package.
type ArgumentDecoder func(string) (map[string]any, error)

type ToolHandlers struct {
	OpenTarget                ToolPayloadHandler
	FillFields                ToolPayloadHandler
	CapturePageSnapshot       ToolPayloadHandler
	CaptureSubmissionEvidence ToolPayloadHandler
	DownloadFile              ToolPayloadHandler
}

func RegisterTools(reg *agentxtools.Registry, handlers ToolHandlers) {
	RegisterToolsWithDecoder(reg, handlers, DecodeToolArguments)
}

// RegisterToolsWithDecoder registers the semantic Browser Ops tool surface with an explicit decoder.
func RegisterToolsWithDecoder(reg *agentxtools.Registry, handlers ToolHandlers, decoder ArgumentDecoder) {
	if decoder == nil {
		decoder = DecodeToolArguments
	}
	registerPayloadTool(reg, BrowserOpenTargetTool(), handlers.OpenTarget, decoder)
	registerPayloadTool(reg, BrowserFillFieldsTool(), handlers.FillFields, decoder)
	registerPayloadTool(reg, BrowserCapturePageSnapshotTool(), handlers.CapturePageSnapshot, decoder)
	registerPayloadTool(reg, BrowserCaptureSubmissionEvidenceTool(), handlers.CaptureSubmissionEvidence, decoder)
	registerPayloadTool(reg, BrowserDownloadFileTool(), handlers.DownloadFile, decoder)
}

// DecodeToolArguments strictly decodes one JSON object. Compatibility repair belongs to the Host.
func DecodeToolArguments(raw string) (map[string]any, error) {
	if strings.TrimSpace(raw) == "" {
		return map[string]any{}, nil
	}
	var out map[string]any
	err := json.Unmarshal([]byte(raw), &out)
	if err != nil {
		return nil, fmt.Errorf("decode browser-ops tool arguments: %w", err)
	}
	if out == nil {
		return nil, fmt.Errorf("decode browser-ops tool arguments: top-level JSON object is required")
	}
	return out, nil
}

func registerPayloadTool(reg *agentxtools.Registry, tool llm.Tool, handler ToolPayloadHandler, decoder ArgumentDecoder) {
	if reg == nil || handler == nil {
		return
	}
	reg.Register(tool, func(ctx context.Context, call llm.FunctionCall) (string, error) {
		params, err := decoder(call.Arguments)
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

func BrowserOpenTargetTool() llm.Tool {
	return functionTool(
		ToolBrowserOpenTarget,
		"Open the host-approved browser target URL and keep the current browser context on that page. Host adapters own runtime target selection, browser profile, login state, target-site policy, approvals, and network/download policy.",
		map[string]any{
			"url":            map[string]any{"type": "string", "description": "Target URL to open."},
			"target_url":     map[string]any{"type": "string", "description": "Alias for url when the value comes from case input."},
			"runtime_target": map[string]any{"type": "string", "description": "Host runtime selector such as node or host when multiple browser runtimes are available."},
			"profile":        map[string]any{"type": "string", "description": "Host browser profile identifier; host owns profile creation, credentials, and login state."},
			"browser":        map[string]any{"type": "string", "description": "Browser backend hint when the host supports multiple browser engines or app targets."},
			"browser_app":    map[string]any{"type": "string", "description": "Browser application hint for hosts that route through a named local app."},
			"wait_ms":        map[string]any{"type": "integer", "minimum": 0, "description": "Maximum wait after opening the URL before returning runtime evidence."},
			"force":          map[string]any{"type": "boolean", "description": "Ask the host adapter to bypass cached state or reuse decisions when allowed by policy."},
		},
		[]string{"url"},
	)
}

func BrowserFillFieldsTool() llm.Tool {
	return functionTool(
		ToolBrowserFillFields,
		"Fill a host-approved form field list in the current browser context and optionally submit it. Host adapters own selector policy, actionability handling, approvals, browser profile, login state, and site-specific behavior.",
		map[string]any{
			"fields": map[string]any{
				"type":        "array",
				"description": "Form fields to fill. Each item should identify one target control by ref, element_ref, input_ref, or selector and provide value or values.",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"ref":         map[string]any{"type": "string", "description": "Stable element reference from a prior browser snapshot."},
						"element_ref": map[string]any{"type": "string", "description": "Alias for ref when the snapshot labels the target as an element reference."},
						"input_ref":   map[string]any{"type": "string", "description": "Alias for ref when the target is specifically an input control."},
						"selector":    map[string]any{"type": "string", "description": "Host-approved CSS or runtime selector fallback when no snapshot ref is available."},
						"type":        map[string]any{"type": "string", "description": "Optional control or input type hint such as text, select, checkbox, or textarea."},
						"value":       map[string]any{"type": "string", "description": "Single value to fill into the target control."},
						"values":      map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Multiple values for controls that support multi-select or repeated entries."},
					},
					"additionalProperties": true,
				},
			},
			"submit":         map[string]any{"type": "boolean", "description": "Whether to submit the form after filling fields; host may still block unsafe submissions."},
			"target":         map[string]any{"type": "string", "description": "Browser target selector such as current, tab:N, or a host runtime target id."},
			"tab_index":      map[string]any{"type": "integer", "minimum": 1, "description": "One-based browser tab index alias for target selection."},
			"index":          map[string]any{"type": "integer", "minimum": 1, "description": "One-based tab index alias kept for runtime compatibility."},
			"runtime_target": map[string]any{"type": "string", "description": "Host runtime selector such as node or host when multiple browser runtimes are available."},
			"profile":        map[string]any{"type": "string", "description": "Host browser profile identifier; host owns profile creation, credentials, and login state."},
			"browser":        map[string]any{"type": "string", "description": "Browser backend hint when the host supports multiple browser engines or app targets."},
			"browser_app":    map[string]any{"type": "string", "description": "Browser application hint for hosts that route through a named local app."},
			"wait_ms":        map[string]any{"type": "integer", "minimum": 0, "description": "Maximum wait before or during the fill operation."},
			"post_wait_ms":   map[string]any{"type": "integer", "minimum": 0, "description": "Additional wait after fill or submit so the host can observe navigation or validation state."},
			"force":          map[string]any{"type": "boolean", "description": "Ask the host adapter to bypass cached state or reuse decisions when allowed by policy."},
		},
		[]string{"fields"},
	)
}

func BrowserCapturePageSnapshotTool() llm.Tool {
	return functionTool(
		ToolBrowserCapturePageSnapshot,
		"Capture a compact structural page snapshot from the current browser context for evaluator grounding. Host adapters own the browser runtime, tab/session selection, and privacy/source policy.",
		map[string]any{
			"target":         map[string]any{"type": "string", "description": "Browser target selector such as current, tab:N, or a host runtime target id."},
			"tab_index":      map[string]any{"type": "integer", "minimum": 1, "description": "One-based browser tab index alias for target selection."},
			"index":          map[string]any{"type": "integer", "minimum": 1, "description": "One-based tab index alias kept for runtime compatibility."},
			"runtime_target": map[string]any{"type": "string", "description": "Host runtime selector such as node or host when multiple browser runtimes are available."},
			"profile":        map[string]any{"type": "string", "description": "Host browser profile identifier; host owns profile creation, credentials, and login state."},
			"browser":        map[string]any{"type": "string", "description": "Browser backend hint when the host supports multiple browser engines or app targets."},
			"browser_app":    map[string]any{"type": "string", "description": "Browser application hint for hosts that route through a named local app."},
			"max_chars":      map[string]any{"type": "integer", "minimum": 64, "description": "Maximum text characters to include in the compact structural snapshot."},
			"max_elements":   map[string]any{"type": "integer", "minimum": 1, "description": "Maximum interactive or structural elements to include in the compact snapshot."},
			"wait_ms":        map[string]any{"type": "integer", "minimum": 0, "description": "Maximum wait before capturing the page snapshot."},
			"force":          map[string]any{"type": "boolean", "description": "Ask the host adapter to refresh page state instead of reusing cached snapshot data when allowed."},
		},
		nil,
	)
}

func BrowserCaptureSubmissionEvidenceTool() llm.Tool {
	return functionTool(
		ToolBrowserCaptureSubmissionEvidence,
		"Capture final screenshot evidence from the current browser context for artifact and event-log grounding. Host adapters own output roots, retention policy, profile/session state, and approval policy.",
		map[string]any{
			"target":          map[string]any{"type": "string", "description": "Browser target selector such as current, tab:N, or a host runtime target id."},
			"tab_index":       map[string]any{"type": "integer", "minimum": 1, "description": "One-based browser tab index alias for target selection."},
			"index":           map[string]any{"type": "integer", "minimum": 1, "description": "One-based tab index alias kept for runtime compatibility."},
			"path":            map[string]any{"type": "string", "description": "Workspace-local or host-approved screenshot artifact path."},
			"output":          map[string]any{"type": "string", "description": "Alias for path used by callers that name the screenshot destination output."},
			"output_path":     map[string]any{"type": "string", "description": "Alias for path used by artifact-oriented callers."},
			"runtime_target":  map[string]any{"type": "string", "description": "Host runtime selector such as node or host when multiple browser runtimes are available."},
			"profile":         map[string]any{"type": "string", "description": "Host browser profile identifier; host owns profile creation, credentials, and login state."},
			"browser":         map[string]any{"type": "string", "description": "Browser backend hint when the host supports multiple browser engines or app targets."},
			"browser_app":     map[string]any{"type": "string", "description": "Browser application hint for hosts that route through a named local app."},
			"full_page":       map[string]any{"type": "boolean", "description": "Capture the full scrollable page when supported; otherwise capture the viewport."},
			"wait_ms":         map[string]any{"type": "integer", "minimum": 0, "description": "Maximum wait before capturing screenshot evidence."},
			"force":           map[string]any{"type": "boolean", "description": "Ask the host adapter to refresh page state before capture when allowed."},
			"remember_target": map[string]any{"type": "boolean", "description": "Whether the host should remember this browser target for later browser-ops steps."},
			"remember":        map[string]any{"type": "boolean", "description": "Alias for remember_target kept for runtime compatibility."},
		},
		nil,
	)
}

func BrowserDownloadFileTool() llm.Tool {
	return functionTool(
		ToolBrowserDownloadFile,
		"Download a host-approved browser file artifact through the current browser runtime. Host adapters own download roots, artifact retention, target-site policy, approval policy, browser profile, login state, and whether direct download or wait_download is allowed.",
		map[string]any{
			"mode":                        map[string]any{"type": "string", "description": "download for direct URL downloads, or wait_download to wait for a browser-triggered download."},
			"kind":                        map[string]any{"type": "string", "description": "Alias for mode when the caller uses browser runtime action terminology."},
			"url":                         map[string]any{"type": "string", "description": "Direct download URL when mode is download."},
			"href":                        map[string]any{"type": "string", "description": "Alias for url when the link was extracted from a page anchor."},
			"download_url":                map[string]any{"type": "string", "description": "Alias for url when the caller already resolved a download URL."},
			"target":                      map[string]any{"type": "string", "description": "Browser target selector used when waiting for a download from the current page."},
			"tab_index":                   map[string]any{"type": "integer", "minimum": 1, "description": "One-based browser tab index alias for target selection."},
			"index":                       map[string]any{"type": "integer", "minimum": 1, "description": "One-based tab index alias kept for runtime compatibility."},
			"path":                        map[string]any{"type": "string", "description": "Workspace-local or host-approved destination path for the downloaded artifact."},
			"output":                      map[string]any{"type": "string", "description": "Alias for path used by callers that name the destination output."},
			"output_path":                 map[string]any{"type": "string", "description": "Alias for path used by artifact-oriented callers."},
			"runtime_target":              map[string]any{"type": "string", "description": "Host runtime selector such as node or host when multiple browser runtimes are available."},
			"profile":                     map[string]any{"type": "string", "description": "Host browser profile identifier; host owns profile creation, credentials, and login state."},
			"browser":                     map[string]any{"type": "string", "description": "Browser backend hint when the host supports multiple browser engines or app targets."},
			"browser_app":                 map[string]any{"type": "string", "description": "Browser application hint for hosts that route through a named local app."},
			"wait_ms":                     map[string]any{"type": "integer", "minimum": 0, "description": "Maximum wait for a browser-triggered download event."},
			"timeout_ms":                  map[string]any{"type": "integer", "minimum": 0, "description": "Overall timeout for download or wait_download execution."},
			"force":                       map[string]any{"type": "boolean", "description": "Ask the host adapter to overwrite or refetch the artifact when allowed by policy."},
			"allow_recent_download_reuse": map[string]any{"type": "boolean", "description": "Allow the host adapter to reuse a recent matching download instead of waiting for a new one."},
		},
		nil,
	)
}

func functionTool(name string, description string, properties map[string]any, required []string) llm.Tool {
	params := map[string]any{
		"type":       "object",
		"properties": properties,
	}
	if len(required) != 0 {
		params["required"] = append([]string(nil), required...)
	}
	return llm.Tool{
		Type: "function",
		Function: llm.Function{
			Name:        name,
			Description: description,
			Parameters:  params,
		},
	}
}
