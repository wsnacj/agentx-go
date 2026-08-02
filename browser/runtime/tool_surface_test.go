package browserruntime

import (
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func TestBrowserToolSurfaceSharedInventory(t *testing.T) {
	if !reflect.DeepEqual(BrowserUnifiedToolNames(), []string{"browser"}) {
		t.Fatalf("unexpected unified tool names: %#v", BrowserUnifiedToolNames())
	}
	if !reflect.DeepEqual(BrowserSpecialistToolNames(), []string{"browser_runtime", "browser_act"}) {
		t.Fatalf("unexpected specialist tool names: %#v", BrowserSpecialistToolNames())
	}
	if !reflect.DeepEqual(BrowserCompatToolNames(), []string{
		"browser_open",
		"browser_navigate",
		"browser_tabs",
		"browser_extract",
		"browser_screenshot",
		"browser_click",
		"browser_type",
		"browser_eval",
	}) {
		t.Fatalf("unexpected compat tool names: %#v", BrowserCompatToolNames())
	}
	if !reflect.DeepEqual(BrowserAllToolNames(), []string{
		"browser",
		"browser_runtime",
		"browser_act",
		"browser_open",
		"browser_navigate",
		"browser_tabs",
		"browser_extract",
		"browser_screenshot",
		"browser_click",
		"browser_type",
		"browser_eval",
	}) {
		t.Fatalf("unexpected all tool names: %#v", BrowserAllToolNames())
	}
	for _, name := range BrowserUnifiedToolNames() {
		if !IsBrowserUnifiedToolName(name) || IsBrowserSpecialistToolName(name) || IsBrowserCompatToolName(name) {
			t.Fatalf("unexpected unified membership for %q", name)
		}
	}
	for _, name := range BrowserSpecialistToolNames() {
		if !IsBrowserSpecialistToolName(name) || IsBrowserUnifiedToolName(name) || IsBrowserCompatToolName(name) {
			t.Fatalf("unexpected specialist membership for %q", name)
		}
	}
	for _, name := range BrowserCompatToolNames() {
		if !IsBrowserCompatToolName(name) || IsBrowserUnifiedToolName(name) || IsBrowserSpecialistToolName(name) {
			t.Fatalf("unexpected compat membership for %q", name)
		}
	}
}

func TestBrowserToolSurfaceSharedMappings(t *testing.T) {
	cases := map[string]struct {
		actKind       string
		runtimeAction string
	}{
		"browser_open":       {actKind: "open", runtimeAction: "open"},
		"browser_navigate":   {actKind: "navigate", runtimeAction: "navigate"},
		"browser_tabs":       {actKind: "list_tabs", runtimeAction: "tabs"},
		"browser_extract":    {actKind: "extract", runtimeAction: "extract"},
		"browser_screenshot": {actKind: "screenshot", runtimeAction: "screenshot"},
		"browser_click":      {actKind: "click", runtimeAction: "click"},
		"browser_type":       {actKind: "type", runtimeAction: "type"},
		"browser_eval":       {actKind: "evaluate", runtimeAction: "evaluate"},
	}
	for toolName, tc := range cases {
		if !IsBrowserToolName(toolName) {
			t.Fatalf("expected %q to resolve as browser tool", toolName)
		}
		if got := BrowserCompatActKindForToolName(toolName); got != tc.actKind {
			t.Fatalf("BrowserCompatActKindForToolName(%q) = %q, want %q", toolName, got, tc.actKind)
		}
		if got := BrowserCompatToolNameForActKind(tc.actKind); got != toolName {
			t.Fatalf("BrowserCompatToolNameForActKind(%q) = %q, want %q", tc.actKind, got, toolName)
		}
		if got := BrowserRuntimeActionForToolCall(toolName, nil); got != tc.runtimeAction {
			t.Fatalf("BrowserRuntimeActionForToolCall(%q) = %q, want %q", toolName, got, tc.runtimeAction)
		}
	}
	if IsBrowserToolName("read") {
		t.Fatal("expected non-browser tool to stay out of browser inventory")
	}
	if got := BrowserCompatActKindForToolName("missing"); got != "" {
		t.Fatalf("expected missing compat tool to have empty act kind, got %q", got)
	}
	if got := BrowserCompatToolNameForActKind("missing"); got != "" {
		t.Fatalf("expected missing act kind to resolve empty tool, got %q", got)
	}
	if got := BrowserRuntimeActionForToolCall("browser_act", map[string]any{"kind": " click "}); got != "click" {
		t.Fatalf("expected browser_act kind to normalize to click, got %q", got)
	}
	if got := BrowserRuntimeActionForToolCall("browser", map[string]any{"action": " navigate "}); got != "navigate" {
		t.Fatalf("expected browser action to normalize to navigate, got %q", got)
	}
	if got := BrowserRuntimeActionForToolCall("browser", map[string]any{"kind": " extract "}); got != "extract" {
		t.Fatalf("expected browser kind fallback to normalize to extract, got %q", got)
	}
}

func TestBrowserCompatToolSurfaceImplementationUsesSingleInventory(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test file path")
	}
	raw, err := os.ReadFile(filepath.Join(filepath.Dir(file), "tool_surface.go"))
	if err != nil {
		t.Fatalf("read tool_surface.go: %v", err)
	}
	source := string(raw)
	for _, name := range BrowserCompatToolNames() {
		if got := strings.Count(source, `"`+name+`"`); got != 1 {
			t.Fatalf("expected compat tool %q to appear once in shared runtime inventory, got %d", name, got)
		}
	}
}

func TestBrowserCompatToolNamesStayCentralizedInProductionGo(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test file path")
	}
	moduleRoot := filepath.Dir(filepath.Dir(file))
	allowedFile := filepath.Join(moduleRoot, "runtime", "tool_surface.go")
	var leaks []string
	err := filepath.WalkDir(moduleRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == "docs" || d.Name() == "examples" {
				return filepath.SkipDir
			}
			return nil
		}
		if path == allowedFile || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		source := string(raw)
		for _, name := range BrowserCompatToolNames() {
			if strings.Contains(source, `"`+name+`"`) {
				rel, _ := filepath.Rel(moduleRoot, path)
				leaks = append(leaks, rel+":"+name)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("scan production Go files: %v", err)
	}
	if len(leaks) > 0 {
		t.Fatalf("compat wrapper names must stay centralized in runtime/tool_surface.go, found %s", strings.Join(leaks, ", "))
	}
}
