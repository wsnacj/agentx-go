package browserd

import (
	"testing"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

func TestCapabilitiesForNodeBackendPlanBundledFallback(t *testing.T) {
	plan := NodeBackendPlan{
		Source: NodeBackendSourceManagedBrowser,
		Managed: Plan{
			Enabled: true,
			Command: bundledBrowserdCommandAlias,
		},
	}
	got := CapabilitiesForNodeBackendPlan(plan, agentxbrowserruntime.BrowserCapabilities{})
	if !got.Open || !got.Navigate || !got.Tabs || !got.Extract || !got.Snapshot || !got.Screenshot || !got.Errors || !got.Download || !got.WaitDownload || !got.Dialog || !got.Upload || !got.Fill || !got.Select || !got.Hover || !got.Drag || !got.Click || !got.TypeText || !got.Evaluate || !got.Wait {
		t.Fatalf("expected bundled browserd fallback page actions, got %#v", got)
	}
	if !got.RuntimeStatus || !got.RuntimeStart || !got.RuntimeCreate || !got.RuntimeDelete || !got.RuntimeStop || !got.RuntimeList {
		t.Fatalf("expected bundled browserd fallback runtime lifecycle actions, got %#v", got)
	}
	if got.Resize || got.Console || got.Requests {
		t.Fatalf("expected bundled browserd fallback to stay narrow, got %#v", got)
	}
}

func TestCapabilitiesForNodeBackendPlanPreservesExplicitActKinds(t *testing.T) {
	plan := NodeBackendPlan{
		Source: NodeBackendSourceManagedBrowser,
		Managed: Plan{
			Enabled: true,
			Command: "node",
			Args:    []string{"core/agentx/browserdaemon/node/agentx-browserd.mjs"},
		},
	}
	explicit := agentxbrowserruntime.BrowserCapabilities{Screenshot: true}
	got := CapabilitiesForNodeBackendPlan(plan, explicit)
	if !got.Screenshot || got.Open || got.Navigate || got.Click {
		t.Fatalf("expected explicit act kinds to win over bundled fallback, got %#v", got)
	}
}
