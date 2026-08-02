package tools

import (
	"strings"

	types "github.com/wsnacj/agentx-go/components/llm"
)

func browserLegacyCompatibilityDescription(base string) string {
	base = strings.TrimSpace(base)
	if base == "" {
		return "Deprecated compatibility wrapper for the unified `browser` workbench; treat this as a migration-only fallback and prefer `browser` for new work."
	}
	return base + " Deprecated compatibility wrapper for the unified `browser` workbench; treat this as a migration-only fallback and prefer `browser` for new work unless you are explicitly targeting this legacy tool."
}

func browserRuntimeProfileSchema() map[string]any {
	return map[string]any{
		"type":        "string",
		"enum":        []string{"default", "isolated", "relay"},
		"description": "Browser runtime profile. Use isolated for managed browserd work unless the user or diagnostics require another profile.",
	}
}

func browserRuntimeTargetSchema() map[string]any {
	return map[string]any{
		"type":        "string",
		"enum":        []string{"host", "sandbox", "node"},
		"description": "Browser runtime substrate. Use node for managed browserd. Use sandbox only when diagnostics explicitly report an available sandbox runtime. Use host only when explicitly targeting the legacy system browser or when a diagnostic asks for runtime_target=host.",
	}
}

func browserCompatDefinition(kind string, description string, parameters map[string]any) types.Tool {
	return types.Tool{
		Type: "function",
		Function: types.Function{
			Name:        browserCompatToolForManagedOptInActKind(kind),
			Description: browserLegacyCompatibilityDescription(description),
			Parameters:  browserDescribedInputSchema(parameters),
		},
	}
}

func browserOpenDefinition() types.Tool {
	return browserCompatDefinition(
		"open",
		"Open a web page in a local browser window or tab.",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"url":            map[string]any{"type": "string"},
				"browser":        map[string]any{"type": "string"},
				"browser_app":    map[string]any{"type": "string"},
				"profile":        browserRuntimeProfileSchema(),
				"runtime_target": browserRuntimeTargetSchema(),
				"wait_ms":        map[string]any{"type": "integer", "minimum": 0},
			},
			"required": []string{"url"},
		},
	)
}

func browserNavigateDefinition() types.Tool {
	return browserCompatDefinition(
		"navigate",
		"Navigate the current local browser tab, or a specific tab selected by target/tab_index, to a new URL and wait for it to load.",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"url":            map[string]any{"type": "string"},
				"target":         map[string]any{"type": "string"},
				"tab_index":      map[string]any{"type": "integer", "minimum": 1},
				"index":          map[string]any{"type": "integer", "minimum": 1},
				"browser":        map[string]any{"type": "string"},
				"browser_app":    map[string]any{"type": "string"},
				"profile":        browserRuntimeProfileSchema(),
				"runtime_target": browserRuntimeTargetSchema(),
				"wait_ms":        map[string]any{"type": "integer", "minimum": 0},
				"force":          map[string]any{"type": "boolean"},
			},
			"required": []string{"url"},
		},
	)
}

func browserTabsDefinition() types.Tool {
	return browserCompatDefinition(
		"list_tabs",
		"List browser tabs, or focus/close a specific tab selected by target/tab_index in a local browser window.",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"action":          map[string]any{"type": "string", "enum": []string{"list", "focus", "close"}},
				"target":          map[string]any{"type": "string"},
				"tab_index":       map[string]any{"type": "integer", "minimum": 1},
				"index":           map[string]any{"type": "integer", "minimum": 1},
				"browser":         map[string]any{"type": "string"},
				"browser_app":     map[string]any{"type": "string"},
				"profile":         browserRuntimeProfileSchema(),
				"runtime_target":  browserRuntimeTargetSchema(),
				"wait_ms":         map[string]any{"type": "integer", "minimum": 0},
				"force":           map[string]any{"type": "boolean"},
				"remember_target": map[string]any{"type": "boolean"},
				"remember":        map[string]any{"type": "boolean"},
			},
		},
	)
}

func browserExtractDefinition() types.Tool {
	return browserCompatDefinition(
		"extract",
		"Extract normalized page text from a provided URL or from the current/targeted browser tab, preferring a local browser backend when available.",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"url":            map[string]any{"type": "string"},
				"target":         map[string]any{"type": "string"},
				"tab_index":      map[string]any{"type": "integer", "minimum": 1},
				"index":          map[string]any{"type": "integer", "minimum": 1},
				"browser":        map[string]any{"type": "string"},
				"browser_app":    map[string]any{"type": "string"},
				"profile":        browserRuntimeProfileSchema(),
				"runtime_target": browserRuntimeTargetSchema(),
				"wait_ms":        map[string]any{"type": "integer", "minimum": 0},
				"max_chars":      map[string]any{"type": "integer", "minimum": 256},
				"force":          map[string]any{"type": "boolean"},
			},
		},
	)
}

func browserScreenshotDefinition() types.Tool {
	return browserCompatDefinition(
		"screenshot",
		"Capture a screenshot from a provided URL or from the current/targeted browser tab, preferring the visible browser page or window region over a full-screen capture; Safari also supports full-page and element(ref/selector) capture.",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"url":             map[string]any{"type": "string"},
				"target":          map[string]any{"type": "string"},
				"tab_index":       map[string]any{"type": "integer", "minimum": 1},
				"index":           map[string]any{"type": "integer", "minimum": 1},
				"path":            map[string]any{"type": "string"},
				"output":          map[string]any{"type": "string"},
				"output_path":     map[string]any{"type": "string"},
				"ref":             map[string]any{"type": "string"},
				"element_ref":     map[string]any{"type": "string"},
				"selector":        map[string]any{"type": "string"},
				"full_page":       map[string]any{"type": "boolean"},
				"browser":         map[string]any{"type": "string"},
				"browser_app":     map[string]any{"type": "string"},
				"profile":         browserRuntimeProfileSchema(),
				"runtime_target":  browserRuntimeTargetSchema(),
				"wait_ms":         map[string]any{"type": "integer", "minimum": 0},
				"force":           map[string]any{"type": "boolean"},
				"remember_target": map[string]any{"type": "boolean"},
				"remember":        map[string]any{"type": "boolean"},
			},
		},
	)
}

func browserClickDefinition() types.Tool {
	return browserCompatDefinition(
		"click",
		"Click a DOM element in the current browser tab or a specific tab selected by target/tab_index, using a snapshot ref when available and optionally navigating to a URL first.",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"url":            map[string]any{"type": "string"},
				"target":         map[string]any{"type": "string"},
				"tab_index":      map[string]any{"type": "integer", "minimum": 1},
				"index":          map[string]any{"type": "integer", "minimum": 1},
				"ref":            map[string]any{"type": "string"},
				"element_ref":    map[string]any{"type": "string"},
				"selector":       map[string]any{"type": "string"},
				"element":        map[string]any{"type": "string"},
				"label":          map[string]any{"type": "string"},
				"text":           map[string]any{"type": "string"},
				"browser":        map[string]any{"type": "string"},
				"browser_app":    map[string]any{"type": "string"},
				"profile":        browserRuntimeProfileSchema(),
				"runtime_target": browserRuntimeTargetSchema(),
				"wait_ms":        map[string]any{"type": "integer", "minimum": 0},
				"post_wait_ms":   map[string]any{"type": "integer", "minimum": 0},
				"force":          map[string]any{"type": "boolean"},
			},
		},
	)
}

func browserTypeDefinition() types.Tool {
	return browserCompatDefinition(
		"type",
		"Type text into an input, textarea, or contenteditable element in the current browser tab or a specific tab selected by target/tab_index, using a snapshot ref when available.",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"url":            map[string]any{"type": "string"},
				"target":         map[string]any{"type": "string"},
				"tab_index":      map[string]any{"type": "integer", "minimum": 1},
				"index":          map[string]any{"type": "integer", "minimum": 1},
				"ref":            map[string]any{"type": "string"},
				"element_ref":    map[string]any{"type": "string"},
				"selector":       map[string]any{"type": "string"},
				"element":        map[string]any{"type": "string"},
				"text":           map[string]any{"type": "string"},
				"value":          map[string]any{"type": "string"},
				"submit":         map[string]any{"type": "boolean"},
				"browser":        map[string]any{"type": "string"},
				"browser_app":    map[string]any{"type": "string"},
				"profile":        browserRuntimeProfileSchema(),
				"runtime_target": browserRuntimeTargetSchema(),
				"wait_ms":        map[string]any{"type": "integer", "minimum": 0},
				"post_wait_ms":   map[string]any{"type": "integer", "minimum": 0},
				"force":          map[string]any{"type": "boolean"},
			},
		},
	)
}

func browserEvalDefinition() types.Tool {
	return browserCompatDefinition(
		"evaluate",
		"Evaluate JavaScript in the current browser tab or a specific tab selected by target/tab_index, optionally navigating to a URL first.",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"url":            map[string]any{"type": "string"},
				"target":         map[string]any{"type": "string"},
				"tab_index":      map[string]any{"type": "integer", "minimum": 1},
				"index":          map[string]any{"type": "integer", "minimum": 1},
				"script":         map[string]any{"type": "string"},
				"javascript":     map[string]any{"type": "string"},
				"js":             map[string]any{"type": "string"},
				"browser":        map[string]any{"type": "string"},
				"browser_app":    map[string]any{"type": "string"},
				"profile":        browserRuntimeProfileSchema(),
				"runtime_target": browserRuntimeTargetSchema(),
				"wait_ms":        map[string]any{"type": "integer", "minimum": 0},
				"max_chars":      map[string]any{"type": "integer", "minimum": 64},
				"force":          map[string]any{"type": "boolean"},
			},
			"required": []string{"script"},
		},
	)
}

func browserActKindsSummary(kinds []string) string {
	if len(kinds) == 0 {
		return ""
	}
	return strings.Join(kinds, ", ")
}
