package tools

import (
	"testing"

	llmxtools "github.com/wsnacj/agentx-go/tools"
)

func TestInferBrowserToolMetadataMissingSkipsCompleteProvidedEntries(t *testing.T) {
	reg := llmxtools.NewRegistry()
	RegisterBrowserTools(reg, BrowserToolOptions{Root: t.TempDir(), EnabledTools: []string{"browser_runtime", "browser_open", "browser_act"}})

	provided := BrowserToolMetadata(reg.Definitions(), BrowserRuntimeInfo{
		Backend: "proxy",
		Profile: "isolated",
		Target:  "node",
	})
	missing := inferBrowserToolMetadataMissing(reg.Definitions(), map[string]ToolMetadata{
		"browser_open": provided["browser_open"],
	})
	if _, ok := missing["browser_open"]; ok {
		t.Fatalf("expected complete provided browser_open metadata to skip redundant inference, got %#v", missing["browser_open"])
	}
	if missing["browser_runtime"].Type != "browser" || len(missing["browser_runtime"].Capabilities) == 0 {
		t.Fatalf("expected missing browser_runtime metadata to still be inferred, got %#v", missing["browser_runtime"])
	}
}

func TestBrowserToolMetadataCarriesBuiltinRiskProfile(t *testing.T) {
	reg := llmxtools.NewRegistry()
	RegisterBrowserTools(reg, BrowserToolOptions{Root: t.TempDir(), EnabledTools: []string{"browser", "browser_runtime", "browser_act", "browser_extract"}})

	metadata := BrowserToolMetadataForOptions(reg.Definitions(), BrowserToolOptions{Root: t.TempDir(), EnabledTools: []string{"browser", "browser_runtime", "browser_act", "browser_extract"}})
	if got := metadata["browser"].RiskProfile; got != "low" {
		t.Fatalf("expected browser risk profile low, got %#v", metadata["browser"])
	}
	if got := metadata["browser_runtime"].RiskProfile; got != "low" {
		t.Fatalf("expected browser_runtime risk profile low, got %#v", metadata["browser_runtime"])
	}
	if got := metadata["browser_act"].RiskProfile; got != "high" {
		t.Fatalf("expected browser_act risk profile high, got %#v", metadata["browser_act"])
	}
	if got := metadata["browser_extract"].RiskProfile; got != "medium" {
		t.Fatalf("expected browser_extract risk profile medium, got %#v", metadata["browser_extract"])
	}
	if got := metadata["browser_extract"].Source; got != ToolSourceBuiltin {
		t.Fatalf("expected browser_extract metadata source builtin, got %#v", metadata["browser_extract"])
	}
}
