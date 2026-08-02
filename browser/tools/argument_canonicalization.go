package tools

import (
	"encoding/json"
	"strings"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

// DecodeToolArguments exposes the tolerant Browser argument decoder.
func DecodeToolArguments(raw string) (map[string]any, error) {
	return decodeArgs(raw)
}

// CanonicalizeToolArguments applies deterministic cleanup to Browser tool
// arguments that already support tolerant aliases and noisy model tokens.
func CanonicalizeToolArguments(toolName, raw string) (string, bool, error) {
	if strings.TrimSpace(raw) == "" {
		return "", false, nil
	}
	params, err := decodeArgs(raw)
	if err != nil {
		return "", false, err
	}
	mutated := false
	switch normalized := NormalizeToolName(toolName); {
	case normalized == "browser":
		mutated = canonicalizeBrowserUnifiedArguments(params)
	case browserToolUsesActionArgumentCanonicalization(normalized, params):
		mutated = canonicalizeBrowserActionArguments(params)
	default:
		return "", false, nil
	}
	if !mutated {
		return "", false, nil
	}
	blob, err := json.Marshal(params)
	if err != nil {
		return "", false, err
	}
	canonical := string(blob)
	return canonical, strings.TrimSpace(canonical) != strings.TrimSpace(raw), nil
}

func browserToolUsesActionArgumentCanonicalization(toolName string, params map[string]any) bool {
	if !agentxbrowserruntime.IsBrowserToolName(toolName) || toolName == "browser" || toolName == "browser_runtime" {
		return false
	}
	return browserToolSharedActionName(toolName, params) != ""
}

func browserToolSharedActionName(toolName string, params map[string]any) string {
	normalized := NormalizeToolName(toolName)
	if !agentxbrowserruntime.IsBrowserToolName(normalized) {
		return ""
	}
	switch normalized {
	case "browser_runtime":
		return ""
	case "browser":
		action := browserNormalizeToolToken(agentxbrowserruntime.BrowserRuntimeActionForToolCall(normalized, params))
		if action == "" || action == "browser" {
			return ""
		}
		return browserNormalizeToolToken(browserUnifiedActKind(params, action))
	case "browser_act":
		action := browserNormalizeToolToken(agentxbrowserruntime.BrowserRuntimeActionForToolCall(normalized, params))
		if action == "" || action == "act" {
			return ""
		}
		return action
	default:
		return browserNormalizeToolToken(agentxbrowserruntime.BrowserCompatActKindForToolName(normalized))
	}
}

func canonicalizeBrowserUnifiedArguments(params map[string]any) bool {
	if params == nil {
		return false
	}
	mutated := canonicalizeBrowserTokenField(params, "action")
	mutated = canonicalizeBrowserTokenField(params, "operation") || mutated
	action := browserNormalizeToolToken(firstString(params, "action", "operation", "mode"))
	actKind := browserNormalizeToolToken(browserUnifiedActKind(params, action))
	if action == "snapshot" || actKind == "snapshot" {
		mutated = canonicalizeBrowserSnapshotFields(params) || mutated
	}
	for _, key := range []string{"selector", "ref", "element", "target", "url"} {
		mutated = canonicalizeBrowserStringField(params, key) || mutated
	}
	return mutated
}

func canonicalizeBrowserActionArguments(params map[string]any) bool {
	if params == nil {
		return false
	}
	mutated := canonicalizeBrowserTokenField(params, "kind")
	if browserNormalizeToolToken(firstString(params, "kind")) == "snapshot" {
		mutated = canonicalizeBrowserSnapshotFields(params) || mutated
	}
	for _, key := range []string{
		"selector", "ref", "element", "text", "value", "target", "url",
		"start_selector", "start_ref", "start_element", "start_label",
		"end_selector", "end_ref", "end_element", "end_label",
	} {
		mutated = canonicalizeBrowserStringField(params, key) || mutated
	}
	mutated = canonicalizeBrowserEndpointField(params, "from") || mutated
	mutated = canonicalizeBrowserEndpointField(params, "to") || mutated
	return mutated
}

func canonicalizeBrowserSnapshotFields(params map[string]any) bool {
	if params == nil {
		return false
	}
	mutated := false
	for _, key := range []string{"format", "snapshot_format"} {
		if value, changed := canonicalizeBrowserSnapshotFormatField(params, key); changed {
			params[key] = value
			mutated = true
		}
	}
	mutated = canonicalizeBrowserTokenField(params, "mode") || mutated
	if value, changed := canonicalizeBrowserSnapshotRefsField(params, "refs"); changed {
		params["refs"] = value
		mutated = true
	}
	return mutated
}

func canonicalizeBrowserSnapshotFormatField(params map[string]any, key string) (string, bool) {
	current, ok := params[key].(string)
	if !ok {
		return "", false
	}
	canonical := browserCanonicalSnapshotFormat(current)
	return canonical, canonical != "" && canonical != strings.TrimSpace(current)
}

func canonicalizeBrowserSnapshotRefsField(params map[string]any, key string) (string, bool) {
	current, ok := params[key].(string)
	if !ok {
		return "", false
	}
	normalized := browserNormalizeToolToken(current)
	return normalized, normalized != "" && normalized != strings.TrimSpace(current)
}

func canonicalizeBrowserTokenField(params map[string]any, key string) bool {
	current, ok := params[key].(string)
	if !ok {
		return false
	}
	normalized := browserNormalizeToolToken(current)
	if normalized == "" || normalized == strings.TrimSpace(current) {
		return false
	}
	params[key] = normalized
	return true
}

func canonicalizeBrowserStringField(params map[string]any, key string) bool {
	current, ok := params[key].(string)
	if !ok {
		return false
	}
	sanitized := browserSanitizeLooseArgumentString(current)
	if sanitized == "" || sanitized == strings.TrimSpace(current) {
		return false
	}
	params[key] = sanitized
	return true
}

func canonicalizeBrowserEndpointField(params map[string]any, key string) bool {
	raw, ok := params[key]
	if !ok || raw == nil {
		return false
	}
	switch current := raw.(type) {
	case string:
		sanitized := browserSanitizeLooseArgumentString(current)
		if sanitized == "" || sanitized == strings.TrimSpace(current) {
			return false
		}
		params[key] = sanitized
		return true
	case map[string]any:
		mutated := false
		for _, field := range []string{"selector", "ref", "element", "label"} {
			mutated = canonicalizeBrowserStringField(current, field) || mutated
		}
		return mutated
	default:
		return false
	}
}
