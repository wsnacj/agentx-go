package tools

import "strings"

var browserInteractiveApprovalExemptActKinds = map[string]bool{
	"extract":       true,
	"snapshot":      true,
	"screenshot":    true,
	"console":       true,
	"requests":      true,
	"response_body": true,
	"errors":        true,
	"cookies":       true,
	"storage":       true,
	"list_tabs":     true,
}

// BrowserSkipsInteractiveApprovalPrompt centralizes browser-specific
// non-mutating prompt exemptions used by operator-facing approval UIs.
// It is intentionally narrower than execution/risk control owners: this helper
// only decides whether an interactive side-effect prompt can be skipped.
func BrowserSkipsInteractiveApprovalPrompt(toolName string, arguments string) bool {
	normalized := NormalizeToolName(toolName)
	if normalized == "" {
		return false
	}
	params, err := decodeArgs(arguments)
	if err != nil {
		return false
	}
	if browserCompatSkipsInteractiveApprovalPrompt(normalized, params) {
		return true
	}
	switch normalized {
	case "browser_act":
		kind := strings.ToLower(strings.TrimSpace(firstString(params, "kind")))
		return browserInteractiveApprovalExemptActKinds[kind]
	default:
		return false
	}
}
