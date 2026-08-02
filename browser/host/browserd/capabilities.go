package browserd

import (
	"path/filepath"
	"strings"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

const bundledBrowserdEntry = "agentx-browserd.mjs"

func CapabilitiesForNodeBackendPlan(plan NodeBackendPlan, fallback agentxbrowserruntime.BrowserCapabilities) agentxbrowserruntime.BrowserCapabilities {
	if fallback.SupportsAnyActKind() || !isBundledNodeBrowserdPlan(plan) {
		return fallback
	}
	return agentxbrowserruntime.BrowserCapabilities{
		RuntimeStatus: true,
		RuntimeStart:  true,
		RuntimeCreate: true,
		RuntimeDelete: true,
		RuntimeStop:   true,
		RuntimeList:   true,
		Open:          true,
		Navigate:      true,
		Tabs:          true,
		Extract:       true,
		Snapshot:      true,
		Screenshot:    true,
		Errors:        true,
		Download:      true,
		WaitDownload:  true,
		Dialog:        true,
		Upload:        true,
		Fill:          true,
		Select:        true,
		Hover:         true,
		Drag:          true,
		Click:         true,
		TypeText:      true,
		Evaluate:      true,
		Wait:          true,
	}
}

func isBundledNodeBrowserdPlan(plan NodeBackendPlan) bool {
	if !plan.UsesManagedBrowserd() {
		return false
	}
	if bundledBrowserdPathMatch(plan.Managed.Command) {
		return true
	}
	for _, arg := range plan.Managed.Args {
		if bundledBrowserdPathMatch(arg) {
			return true
		}
	}
	return false
}

func bundledBrowserdPathMatch(value string) bool {
	trimmed := strings.TrimSpace(value)
	if isBundledBrowserdCommand(trimmed) {
		return true
	}
	base := strings.TrimSpace(filepath.Base(trimmed))
	return strings.EqualFold(base, bundledBrowserdEntry)
}
