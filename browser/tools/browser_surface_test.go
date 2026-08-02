package tools

import (
	"errors"
	"reflect"
	"testing"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
	types "github.com/wsnacj/agentx-go/components/llm"
)

func TestBrowserToolNameLists(t *testing.T) {
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
}

func TestBrowserCompatToolErrorfUsesSharedActKindResolver(t *testing.T) {
	err := browserCompatToolErrorf("screenshot", "failed: %w", errors.New("boom"))
	if got := err.Error(); got != "browser_screenshot: failed: boom" {
		t.Fatalf("unexpected compat error prefix: %q", got)
	}
	if got := browserCompatToolErrorf("missing", "failed").Error(); got != "browser: failed" {
		t.Fatalf("unexpected missing compat error prefix fallback: %q", got)
	}
}

func TestBrowserCompatDescriptorsStayInSync(t *testing.T) {
	compatNames := BrowserCompatToolNames()
	if len(browserCompatDescriptorsByName) != len(compatNames) {
		t.Fatalf("expected compat descriptor count to match shared runtime inventory: descriptors=%d inventory=%d", len(browserCompatDescriptorsByName), len(compatNames))
	}
	for name := range browserCompatDescriptorsByName {
		if !agentxbrowserruntime.IsBrowserCompatToolName(name) {
			t.Fatalf("expected compat descriptor %q to exist in shared runtime inventory", name)
		}
	}
	for _, name := range compatNames {
		descriptor, ok := browserCompatDescriptorForTool(name)
		if !ok {
			t.Fatalf("expected compat descriptor for %q", name)
		}
		if descriptor.Name != name {
			t.Fatalf("expected descriptor name %q, got %#v", name, descriptor)
		}
		wantActKind := agentxbrowserruntime.BrowserCompatActKindForToolName(name)
		if wantActKind == "" {
			t.Fatalf("expected shared managed opt-in act kind for %q", name)
		}
		if descriptor.ActKind != wantActKind {
			t.Fatalf("expected descriptor act kind %q for %q, got %#v", wantActKind, name, descriptor)
		}
		meta, ok := browserCompatMetadataForTool(name)
		if !ok {
			t.Fatalf("expected compat metadata for %q", name)
		}
		if len(meta.Groups) != 1 || meta.Groups[0] != BrowserSurfaceCompat {
			t.Fatalf("expected compat metadata groups for %q, got %#v", name, meta)
		}
		if len(meta.Capabilities) == 0 || len(meta.AuditTags) == 0 {
			t.Fatalf("expected compat metadata payload for %q, got %#v", name, meta)
		}
		if browserCompatRegistrationHandler(name) == nil {
			t.Fatalf("expected compat registration owner for %q", name)
		}
		if got := browserCompatManagedOptInActKind(name); got != wantActKind {
			t.Fatalf("expected compat act kind %q for %q, got %q", wantActKind, name, got)
		}
		if got := browserCompatToolForManagedOptInActKind(wantActKind); got != name {
			t.Fatalf("expected compat tool lookup by act kind %q to resolve %q, got %q", wantActKind, name, got)
		}
		if got := browserCompatIsReadOnly(name); got != descriptor.ReadOnly {
			t.Fatalf("expected compat read-only=%v for %q, got %v", descriptor.ReadOnly, name, got)
		}
		if got := browserCompatIsLikelySideEffect(name); got != descriptor.LikelySideEffect {
			t.Fatalf("expected compat likely-side-effect=%v for %q, got %v", descriptor.LikelySideEffect, name, got)
		}
		gotRisk, gotKnown := browserCompatBuiltinRiskLevel(name)
		wantKnown := descriptor.BuiltinRisk != RiskUnknown
		if gotKnown != wantKnown || gotRisk != descriptor.BuiltinRisk {
			t.Fatalf("expected compat builtin risk (%v, %v) for %q, got (%v, %v)", descriptor.BuiltinRisk, wantKnown, name, gotRisk, gotKnown)
		}
	}
	if _, ok := browserCompatDescriptorForTool("browser"); ok {
		t.Fatal("expected unified browser not to resolve as compat descriptor")
	}
	if got := browserCompatToolForManagedOptInActKind("missing"); got != "" {
		t.Fatalf("expected unknown compat act kind to resolve empty tool name, got %q", got)
	}
}

func TestBrowserToolRegistrationsStayInSync(t *testing.T) {
	want := append([]string{"browser_runtime"}, BrowserCompatToolNames()...)
	want = append(want, "browser_act", "browser")
	got := make([]string, 0, len(browserToolRegistrations))
	for _, item := range browserToolRegistrations {
		got = append(got, item.name)
		if item.register == nil {
			t.Fatalf("expected registration handler for %q", item.name)
		}
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected browser tool registrations: got %#v want %#v", got, want)
	}
}

func TestBrowserCompatDefinitionsUseSharedActionKindResolver(t *testing.T) {
	cases := []struct {
		kind       string
		definition func() types.Tool
	}{
		{kind: "open", definition: browserOpenDefinition},
		{kind: "navigate", definition: browserNavigateDefinition},
		{kind: "list_tabs", definition: browserTabsDefinition},
		{kind: "extract", definition: browserExtractDefinition},
		{kind: "screenshot", definition: browserScreenshotDefinition},
		{kind: "click", definition: browserClickDefinition},
		{kind: "type", definition: browserTypeDefinition},
		{kind: "evaluate", definition: browserEvalDefinition},
	}
	for _, tc := range cases {
		t.Run(tc.kind, func(t *testing.T) {
			want := browserCompatToolForManagedOptInActKind(tc.kind)
			if want == "" {
				t.Fatalf("missing compat tool for act kind %q", tc.kind)
			}
			if got := tc.definition().Function.Name; got != want {
				t.Fatalf("expected %q definition to use shared compat tool %q, got %q", tc.kind, want, got)
			}
		})
	}
}

func TestBrowserToolNameSemantics(t *testing.T) {
	tests := []struct {
		name            string
		toolName        string
		wantReadOnly    bool
		wantSideEffect  bool
		wantBuiltinRisk RiskLevel
		wantKnownRisk   bool
	}{
		{name: "unified browser", toolName: "browser", wantBuiltinRisk: RiskLow, wantKnownRisk: true},
		{name: "runtime specialist", toolName: "browser_runtime", wantSideEffect: true, wantBuiltinRisk: RiskLow, wantKnownRisk: true},
		{name: "act specialist", toolName: "browser_act", wantSideEffect: true, wantBuiltinRisk: RiskHigh, wantKnownRisk: true},
		{name: "extract compat", toolName: "browser_extract", wantReadOnly: true, wantBuiltinRisk: RiskMedium, wantKnownRisk: true},
		{name: "open compat", toolName: "browser_open", wantSideEffect: true, wantBuiltinRisk: RiskMedium, wantKnownRisk: true},
		{name: "screenshot compat", toolName: "browser_screenshot", wantSideEffect: true, wantBuiltinRisk: RiskHigh, wantKnownRisk: true},
		{name: "non-browser", toolName: "read"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsBrowserReadOnlyToolName(tc.toolName); got != tc.wantReadOnly {
				t.Fatalf("IsBrowserReadOnlyToolName(%q) = %v, want %v", tc.toolName, got, tc.wantReadOnly)
			}
			if got := IsBrowserLikelySideEffectToolName(tc.toolName); got != tc.wantSideEffect {
				t.Fatalf("IsBrowserLikelySideEffectToolName(%q) = %v, want %v", tc.toolName, got, tc.wantSideEffect)
			}
			gotRisk, gotKnown := BrowserBuiltinRiskLevel(tc.toolName)
			if gotRisk != tc.wantBuiltinRisk || gotKnown != tc.wantKnownRisk {
				t.Fatalf("BrowserBuiltinRiskLevel(%q) = (%v, %v), want (%v, %v)", tc.toolName, gotRisk, gotKnown, tc.wantBuiltinRisk, tc.wantKnownRisk)
			}
		})
	}
}

func TestBrowserSurfaceStatus(t *testing.T) {
	tests := []struct {
		name       string
		unified    bool
		specialist []string
		compat     []string
		want       string
	}{
		{name: "unified wins", unified: true, specialist: []string{"browser_runtime"}, compat: []string{"browser_open"}, want: "ok"},
		{name: "specialist fallback warns", specialist: []string{"browser_runtime"}, want: "warn"},
		{name: "compat fallback warns", compat: []string{"browser_open"}, want: "warn"},
		{name: "none errors", want: "error"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := BrowserSurfaceStatus(tc.unified, tc.specialist, tc.compat); got != tc.want {
				t.Fatalf("BrowserSurfaceStatus() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestBrowserDefaultEntrypointAndSurface(t *testing.T) {
	tests := []struct {
		name           string
		unified        bool
		specialist     []string
		compat         []string
		wantEntrypoint string
		wantSurface    string
	}{
		{
			name:           "unified default",
			unified:        true,
			specialist:     []string{"browser_runtime"},
			compat:         []string{"browser_open"},
			wantEntrypoint: "browser",
			wantSurface:    "browser_unified",
		},
		{
			name:           "specialist fallback",
			specialist:     []string{"browser_runtime", "browser_act"},
			compat:         []string{"browser_open"},
			wantEntrypoint: "browser_runtime",
			wantSurface:    "browser_specialist",
		},
		{
			name:           "compat fallback",
			compat:         []string{"browser_open", "browser_tabs"},
			wantEntrypoint: "browser_open",
			wantSurface:    "browser_compat",
		},
		{
			name:           "none",
			wantEntrypoint: "",
			wantSurface:    "",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := BrowserDefaultEntrypoint(tc.unified, tc.specialist, tc.compat); got != tc.wantEntrypoint {
				t.Fatalf("BrowserDefaultEntrypoint() = %q, want %q", got, tc.wantEntrypoint)
			}
			if got := BrowserDefaultSurface(tc.unified, tc.specialist, tc.compat); got != tc.wantSurface {
				t.Fatalf("BrowserDefaultSurface() = %q, want %q", got, tc.wantSurface)
			}
		})
	}
}

func TestNormalizeAndResolveBrowserSurface(t *testing.T) {
	if got := NormalizeBrowserSurface(" browser_specialist "); got != BrowserSurfaceSpecialist {
		t.Fatalf("NormalizeBrowserSurface() = %q, want %q", got, BrowserSurfaceSpecialist)
	}
	if got := BrowserSurfaceFallbackEntrypoint(BrowserSurfaceCompat); got != "browser_open" {
		t.Fatalf("BrowserSurfaceFallbackEntrypoint() = %q, want %q", got, "browser_open")
	}

	tests := []struct {
		name           string
		surface        string
		entrypoint     string
		wantSurface    string
		wantEntrypoint string
	}{
		{
			name:           "surface implies fallback entrypoint",
			surface:        BrowserSurfaceSpecialist,
			wantSurface:    BrowserSurfaceSpecialist,
			wantEntrypoint: "browser_runtime",
		},
		{
			name:           "entrypoint infers unified",
			entrypoint:     "browser",
			wantSurface:    BrowserSurfaceUnified,
			wantEntrypoint: "browser",
		},
		{
			name:           "entrypoint infers compat",
			entrypoint:     "browser_tabs",
			wantSurface:    BrowserSurfaceCompat,
			wantEntrypoint: "browser_tabs",
		},
		{
			name:           "unknown preserved",
			surface:        "weird",
			entrypoint:     "custom_browser",
			wantSurface:    "",
			wantEntrypoint: "custom_browser",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotSurface, gotEntrypoint := ResolveBrowserSurface(tc.surface, tc.entrypoint)
			if gotSurface != tc.wantSurface || gotEntrypoint != tc.wantEntrypoint {
				t.Fatalf("ResolveBrowserSurface() = (%q, %q), want (%q, %q)", gotSurface, gotEntrypoint, tc.wantSurface, tc.wantEntrypoint)
			}
		})
	}
}

func TestBrowserSurfaceForToolName(t *testing.T) {
	tests := []struct {
		name        string
		toolName    string
		wantSurface string
	}{
		{name: "unified", toolName: "browser", wantSurface: BrowserSurfaceUnified},
		{name: "specialist runtime", toolName: "browser_runtime", wantSurface: BrowserSurfaceSpecialist},
		{name: "specialist act", toolName: "browser_act", wantSurface: BrowserSurfaceSpecialist},
		{name: "compat", toolName: "browser_open", wantSurface: BrowserSurfaceCompat},
		{name: "unknown", toolName: "read", wantSurface: ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := BrowserSurfaceForToolName(tc.toolName); got != tc.wantSurface {
				t.Fatalf("BrowserSurfaceForToolName() = %q, want %q", got, tc.wantSurface)
			}
		})
	}
}

func TestBrowserArtifactKindForToolName(t *testing.T) {
	if got := BrowserArtifactKindForToolName(" browser_screenshot "); got != "screenshot" {
		t.Fatalf("expected browser_screenshot artifact kind, got %q", got)
	}
	if got := BrowserArtifactKindForToolName("browser_extract"); got != "" {
		t.Fatalf("expected browser_extract to have no stable artifact kind, got %q", got)
	}
	if got := BrowserArtifactKindForToolName("read"); got != "" {
		t.Fatalf("expected non-browser tool to have no browser artifact kind, got %q", got)
	}
}

func TestBrowserVisibleSurfaceLabels(t *testing.T) {
	if got := BrowserVisibleSurfaceLabels(nil, nil); got != nil {
		t.Fatalf("expected nil labels, got %#v", got)
	}
	got := BrowserVisibleSurfaceLabels([]string{"browser_runtime", "browser_act"}, []string{"browser_open"})
	want := []string{
		"specialist=browser_runtime/browser_act",
		"deprecated_compat=browser_open",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("BrowserVisibleSurfaceLabels() = %#v, want %#v", got, want)
	}
}
