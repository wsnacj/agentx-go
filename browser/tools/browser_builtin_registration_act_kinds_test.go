package tools

import (
	"context"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func TestBrowserActKindsForRegistration_DefaultSystemBackend(t *testing.T) {
	got := browserActKindsForRegistration(BrowserToolOptions{})
	want := []string{"open", "extract", "snapshot", "wait"}
	if runtime.GOOS == "darwin" {
		want = []string{"open", "navigate", "extract", "snapshot", "screenshot", "click", "type", "evaluate", "wait", "list_tabs", "focus_tab", "close_tab"}
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected default browser_act kinds: got=%#v want=%#v", got, want)
	}
}

func TestBrowserActKindsForRegistration_CustomBackendKeepsFullSurface(t *testing.T) {
	got := browserActKindsForRegistration(BrowserToolOptions{Backend: &fakeBrowserBackend{}})
	want := []string{"open", "navigate", "extract", "snapshot", "screenshot", "click", "type", "evaluate", "wait", "list_tabs", "focus_tab", "close_tab"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected custom backend to keep full browser_act surface, got %#v", got)
	}
}

func TestBrowserActKindsForRegistration_CustomCapabilityBackendNarrowsSurface(t *testing.T) {
	backend := &capabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		capabilities: BrowserCapabilities{
			Open:     true,
			Extract:  true,
			Snapshot: true,
			Wait:     true,
		},
	}
	got := browserActKindsForRegistration(BrowserToolOptions{Backend: backend})
	want := []string{"open", "extract", "snapshot", "wait"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected capability-constrained browser_act kinds: got=%#v want=%#v", got, want)
	}
}

func TestBrowserActKindsForRegistration_NodeBackendExpandsSurface(t *testing.T) {
	host := &capabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		capabilities: BrowserCapabilities{
			Open:     true,
			Extract:  true,
			Snapshot: true,
			Wait:     true,
		},
	}
	node := &capabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		capabilities: BrowserCapabilities{
			Click:    true,
			TypeText: true,
			Evaluate: true,
		},
	}
	got := browserActKindsForRegistration(BrowserToolOptions{
		Backend:     host,
		NodeBackend: node,
	})
	want := []string{"open", "extract", "snapshot", "click", "type", "evaluate", "wait"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected node backend to expand browser_act surface, got=%#v want=%#v", got, want)
	}
}

func TestBrowserActKindsForRegistration_NodeBackendRouteResolutionFailureDoesNotExpandSurface(t *testing.T) {
	host := &capabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		capabilities: BrowserCapabilities{
			Open:     true,
			Extract:  true,
			Snapshot: true,
			Wait:     true,
		},
	}
	node := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{},
			runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
			capabilities: BrowserCapabilities{
				Click:    true,
				TypeText: true,
				Evaluate: true,
			},
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			return BrowserRuntimeInfo{}, context.DeadlineExceeded
		},
	}
	got := browserActKindsForRegistration(BrowserToolOptions{
		Backend:     host,
		NodeBackend: node,
	})
	want := []string{"open", "extract", "snapshot", "wait"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected unresolved node route to stay out of browser_act surface, got=%#v want=%#v", got, want)
	}
}

func TestBrowserActKindsForRegistration_NodeManagedDefaultExpandsManagedOnlySurfaceBeforeHiddenImplicitHostFallback(t *testing.T) {
	node := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{},
			runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
			capabilities: BrowserCapabilities{
				Console:  true,
				Requests: true,
			},
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			requested = normalizeBrowserRuntimeInfo(requested)
			if strings.TrimSpace(requested.Target) != "node" {
				return BrowserRuntimeInfo{}, context.DeadlineExceeded
			}
			if strings.EqualFold(strings.TrimSpace(requested.Profile), "isolated") {
				return BrowserRuntimeInfo{}, context.DeadlineExceeded
			}
			return BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}, nil
		},
	}
	got := browserActKindsForRegistration(BrowserToolOptions{
		NodeBackend: node,
	})
	if !sliceContainsString(got, "console") || !sliceContainsString(got, "requests") {
		t.Fatalf("expected hidden implicit-host registration surface to pick up managed-only act kinds from dynamic managed-default, got %#v", got)
	}
}
