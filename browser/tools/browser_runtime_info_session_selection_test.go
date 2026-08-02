package tools

import (
	"context"
	"strings"
	"testing"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

func TestBrowserRuntimeApplySessionProfileSelectionIgnoresImplicitHostTargetlessHostSelection(t *testing.T) {
	registry := NewBrowserSessionStateRegistry()
	registry.SelectBrowserProfile("s1", agentxbrowserruntime.SharedSessionBrowserProfileSelection{
		Backend:       "system",
		Profile:       "relay",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Source:        "remember_profile",
	})

	got := browserRuntimeApplySessionProfileSelection(
		WithToolSessionID(context.Background(), "s1"),
		registry,
		map[string]any{},
		true,
	)
	if profile := strings.TrimSpace(firstString(got, "profile", "browser_profile", "runtime_profile")); profile != "" {
		t.Fatalf("expected hidden implicit-host default base to ignore targetless host selection, got profile=%q", profile)
	}
	if target := strings.TrimSpace(firstString(got, "runtime_target", "browser_target", "placement")); target != "" {
		t.Fatalf("expected hidden implicit-host default base to ignore targetless host selection, got runtime_target=%q", target)
	}
}

func TestBrowserRuntimeApplySessionProfileSelectionKeepsManagedSelectionForHiddenDefaultBase(t *testing.T) {
	registry := NewBrowserSessionStateRegistry()
	registry.SelectBrowserProfile("s1", agentxbrowserruntime.SharedSessionBrowserProfileSelection{
		Backend:       "proxy",
		Profile:       "workbench",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Source:        "remember_profile",
	})

	got := browserRuntimeApplySessionProfileSelection(
		WithToolSessionID(context.Background(), "s1"),
		registry,
		map[string]any{},
		true,
	)
	if profile := strings.TrimSpace(firstString(got, "profile", "browser_profile", "runtime_profile")); profile != "workbench" {
		t.Fatalf("expected hidden implicit-host default base to keep managed remembered profile, got %q", profile)
	}
	if target := strings.TrimSpace(firstString(got, "runtime_target", "browser_target", "placement")); target != "node" {
		t.Fatalf("expected hidden implicit-host default base to keep managed remembered target, got %q", target)
	}
}

func TestBrowserRuntimeApplySessionTargetRouteSelectionIgnoresImplicitHostCurrentTarget(t *testing.T) {
	registry := NewBrowserSessionRegistry()
	registry.TrackCurrentTarget("s1", BrowserSessionTarget{
		ID:         "host-current",
		TabIndex:   1,
		URL:        "https://host.example/current",
		Title:      "Host Current",
		BrowserApp: "Safari",
		Backend:    "system",
		Profile:    "relay",
		Target:     "host",
	}, "tracked_active_tab")

	got := browserRuntimeApplySessionTargetRouteSelection(
		WithToolSessionID(context.Background(), "s1"),
		registry,
		map[string]any{},
		true,
	)
	if profile := strings.TrimSpace(firstString(got, "profile", "browser_profile", "runtime_profile")); profile != "" {
		t.Fatalf("expected hidden implicit-host default base to ignore current host target profile injection, got %q", profile)
	}
	if target := strings.TrimSpace(firstString(got, "runtime_target", "browser_target", "placement")); target != "" {
		t.Fatalf("expected hidden implicit-host default base to ignore current host target runtime_target injection, got %q", target)
	}
}

func TestBrowserRuntimeApplySessionTargetRouteSelectionKeepsManagedCurrentTargetForHiddenDefaultBase(t *testing.T) {
	registry := NewBrowserSessionRegistry()
	registry.TrackCurrentTarget("s1", BrowserSessionTarget{
		ID:         "node-current",
		TabIndex:   2,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "tracked_active_tab")

	got := browserRuntimeApplySessionTargetRouteSelection(
		WithToolSessionID(context.Background(), "s1"),
		registry,
		map[string]any{},
		true,
	)
	if profile := strings.TrimSpace(firstString(got, "profile", "browser_profile", "runtime_profile")); profile != "workbench" {
		t.Fatalf("expected hidden implicit-host default base to keep managed current target profile, got %q", profile)
	}
	if target := strings.TrimSpace(firstString(got, "runtime_target", "browser_target", "placement")); target != "node" {
		t.Fatalf("expected hidden implicit-host default base to keep managed current target runtime_target, got %q", target)
	}
}

func TestBrowserRuntimeApplyTargetHandleRouteSelectionIgnoresImplicitHostTargetHandleSelection(t *testing.T) {
	registry := NewBrowserSessionRegistry()
	tracked := registry.TrackCurrentTarget("s1", BrowserSessionTarget{
		ID:         "host-target",
		TabIndex:   1,
		URL:        "https://host.example/current",
		Title:      "Host Current",
		BrowserApp: "Safari",
		Backend:    "system",
		Profile:    "relay",
		Target:     "host",
	}, "tracked_active_tab")

	got := browserRuntimeApplyTargetHandleRouteSelection(
		WithToolSessionID(context.Background(), "s1"),
		registry,
		map[string]any{"target": "target:" + tracked.ID},
		BrowserRuntimeInfo{},
		true,
	)
	if profile := strings.TrimSpace(firstString(got, "profile", "browser_profile", "runtime_profile")); profile != "" {
		t.Fatalf("expected hidden implicit-host default base to ignore host target-handle profile injection, got %q", profile)
	}
	if target := strings.TrimSpace(firstString(got, "runtime_target", "browser_target", "placement")); target != "" {
		t.Fatalf("expected hidden implicit-host default base to ignore host target-handle runtime_target injection, got %q", target)
	}
	if browserApp := strings.TrimSpace(firstString(got, "browser", "browser_app", "app")); browserApp != "" {
		t.Fatalf("expected hidden implicit-host default base to ignore host target-handle browser_app injection, got %q", browserApp)
	}
}

func TestBrowserRuntimeApplyTargetHandleRouteSelectionKeepsManagedTargetHandleSelectionForHiddenDefaultBase(t *testing.T) {
	registry := NewBrowserSessionRegistry()
	tracked := registry.TrackCurrentTarget("s1", BrowserSessionTarget{
		ID:         "node-target",
		TabIndex:   2,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "tracked_active_tab")

	got := browserRuntimeApplyTargetHandleRouteSelection(
		WithToolSessionID(context.Background(), "s1"),
		registry,
		map[string]any{"target": "target:" + tracked.ID},
		BrowserRuntimeInfo{},
		true,
	)
	if profile := strings.TrimSpace(firstString(got, "profile", "browser_profile", "runtime_profile")); profile != "workbench" {
		t.Fatalf("expected hidden implicit-host default base to keep managed target-handle profile injection, got %q", profile)
	}
	if target := strings.TrimSpace(firstString(got, "runtime_target", "browser_target", "placement")); target != "node" {
		t.Fatalf("expected hidden implicit-host default base to keep managed target-handle runtime_target injection, got %q", target)
	}
	if browserApp := strings.TrimSpace(firstString(got, "browser", "browser_app", "app")); browserApp != "Chromium" {
		t.Fatalf("expected hidden implicit-host default base to keep managed target-handle browser_app injection, got %q", browserApp)
	}
}

func TestBrowserRuntimeApplyTargetHandleRouteSelectionKeepsManagedExplicitCurrentSelectionForHiddenDefaultBase(t *testing.T) {
	registry := NewBrowserSessionRegistry()
	registry.TrackCurrentTarget("s1", BrowserSessionTarget{
		ID:         "node-current",
		TabIndex:   2,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "tracked_active_tab")

	got := browserRuntimeApplyTargetHandleRouteSelection(
		WithToolSessionID(context.Background(), "s1"),
		registry,
		map[string]any{"target": "current"},
		BrowserRuntimeInfo{},
		true,
	)
	if profile := strings.TrimSpace(firstString(got, "profile", "browser_profile", "runtime_profile")); profile != "workbench" {
		t.Fatalf("expected hidden implicit-host default base to keep managed explicit current profile injection, got %q", profile)
	}
	if target := strings.TrimSpace(firstString(got, "runtime_target", "browser_target", "placement")); target != "node" {
		t.Fatalf("expected hidden implicit-host default base to keep managed explicit current runtime_target injection, got %q", target)
	}
	if browserApp := strings.TrimSpace(firstString(got, "browser", "browser_app", "app")); browserApp != "Chromium" {
		t.Fatalf("expected hidden implicit-host default base to keep managed explicit current browser_app injection, got %q", browserApp)
	}
}

func TestBrowserRuntimeApplyTargetHandleRouteSelectionKeepsManagedExplicitTabSelectionForHiddenDefaultBase(t *testing.T) {
	registry := NewBrowserSessionRegistry()
	registry.TrackCurrentTarget("s1", BrowserSessionTarget{
		ID:         "node-current",
		TabIndex:   4,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "tracked_active_tab")

	got := browserRuntimeApplyTargetHandleRouteSelection(
		WithToolSessionID(context.Background(), "s1"),
		registry,
		map[string]any{"tab_index": 2},
		BrowserRuntimeInfo{},
		true,
	)
	if profile := strings.TrimSpace(firstString(got, "profile", "browser_profile", "runtime_profile")); profile != "workbench" {
		t.Fatalf("expected hidden implicit-host default base to keep managed explicit tab profile injection, got %q", profile)
	}
	if target := strings.TrimSpace(firstString(got, "runtime_target", "browser_target", "placement")); target != "node" {
		t.Fatalf("expected hidden implicit-host default base to keep managed explicit tab runtime_target injection, got %q", target)
	}
	if browserApp := strings.TrimSpace(firstString(got, "browser", "browser_app", "app")); browserApp != "Chromium" {
		t.Fatalf("expected hidden implicit-host default base to keep managed explicit tab browser_app injection, got %q", browserApp)
	}
}

func TestBrowserRuntimeApplyPageBoundElementRouteSelectionKeepsManagedCurrentRouteForHiddenDefaultBase(t *testing.T) {
	registry := NewBrowserSessionRegistry()
	registry.TrackCurrentTarget("s1", BrowserSessionTarget{
		ID:         "node-current",
		TabIndex:   2,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "tracked_active_tab")
	ref := browserElementRefForSnapshotElement(BrowserSnapshotElement{
		Selector: `button[aria-label="Run"]`,
		Label:    "Run",
		Role:     "button",
	}, "https://node.example/workbench", "Workbench")

	got := browserRuntimeApplyPageBoundElementRouteSelection(
		WithToolSessionID(context.Background(), "s1"),
		registry,
		map[string]any{"ref": ref},
		BrowserRuntimeInfo{},
		true,
	)
	if profile := strings.TrimSpace(firstString(got, "profile", "browser_profile", "runtime_profile")); profile != "workbench" {
		t.Fatalf("expected hidden implicit-host default base to keep managed page-bound ref profile injection, got %q", profile)
	}
	if target := strings.TrimSpace(firstString(got, "runtime_target", "browser_target", "placement")); target != "node" {
		t.Fatalf("expected hidden implicit-host default base to keep managed page-bound ref runtime_target injection, got %q", target)
	}
	if browserApp := strings.TrimSpace(firstString(got, "browser", "browser_app", "app")); browserApp != "Chromium" {
		t.Fatalf("expected hidden implicit-host default base to keep managed page-bound ref browser_app injection, got %q", browserApp)
	}
}

func TestResolveBrowserToolTargetIgnoresImplicitHostCurrentTargetFallback(t *testing.T) {
	registry := NewBrowserSessionRegistry()
	registry.TrackCurrentTarget("s1", BrowserSessionTarget{
		ID:         "host-current",
		TabIndex:   1,
		URL:        "https://host.example/current",
		Title:      "Host Current",
		BrowserApp: "Safari",
		Backend:    "system",
		Profile:    "default",
		Target:     "host",
	}, "tracked_active_tab")

	got, err := resolveBrowserToolTarget(
		WithToolSessionID(context.Background(), "s1"),
		registry,
		BrowserRuntimeInfo{},
		true,
		"",
		map[string]any{},
	)
	if err != nil {
		t.Fatalf("resolveBrowserToolTarget(default hidden base): %v", err)
	}
	if got != (browserToolTarget{}) {
		t.Fatalf("expected hidden implicit-host default base to ignore current-target fallback, got %#v", got)
	}

	got, err = resolveBrowserToolTarget(
		WithToolSessionID(context.Background(), "s1"),
		registry,
		BrowserRuntimeInfo{},
		true,
		"",
		map[string]any{"target": "current"},
	)
	if err == nil || !strings.Contains(err.Error(), `requires explicit runtime_target=host`) {
		t.Fatalf("expected explicit current target on hidden implicit-host base to require explicit host runtime_target, got target=%#v err=%v", got, err)
	}
}

func TestResolveBrowserToolTargetRequiresExplicitHostTabTargetForHiddenDefaultBase(t *testing.T) {
	got, err := resolveBrowserToolTarget(
		WithToolSessionID(context.Background(), "s1"),
		NewBrowserSessionRegistry(),
		BrowserRuntimeInfo{},
		true,
		"",
		map[string]any{"target": "tab:2"},
	)
	if err == nil || !strings.Contains(err.Error(), `requires explicit runtime_target=host`) {
		t.Fatalf("expected hidden implicit-host default base to require explicit host runtime_target for tab target, got target=%#v err=%v", got, err)
	}
}

func TestResolveBrowserToolTargetRequiresExplicitHostTargetHandleForHiddenDefaultBase(t *testing.T) {
	registry := NewBrowserSessionRegistry()
	tracked := registry.TrackCurrentTarget("s1", BrowserSessionTarget{
		ID:         "host-target",
		TabIndex:   2,
		URL:        "https://host.example/target",
		Title:      "Host Target",
		BrowserApp: "Safari",
		Backend:    "system",
		Profile:    "default",
		Target:     "host",
	}, "tracked_active_tab")

	got, err := resolveBrowserToolTarget(
		WithToolSessionID(context.Background(), "s1"),
		registry,
		BrowserRuntimeInfo{},
		true,
		"",
		map[string]any{"target": "target:" + tracked.ID},
	)
	if err == nil || !strings.Contains(err.Error(), `requires explicit runtime_target=host`) {
		t.Fatalf("expected hidden implicit-host default base to require explicit host runtime_target for host target handle, got target=%#v err=%v", got, err)
	}
}

func TestResolveBrowserToolTargetKeepsManagedCurrentTargetForManagedBase(t *testing.T) {
	registry := NewBrowserSessionRegistry()
	tracked := registry.TrackCurrentTarget("s1", BrowserSessionTarget{
		ID:         "node-current",
		TabIndex:   2,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "tracked_active_tab")

	got, err := resolveBrowserToolTarget(
		WithToolSessionID(context.Background(), "s1"),
		registry,
		BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
		false,
		"Chromium",
		map[string]any{"target": "current"},
	)
	if err != nil {
		t.Fatalf("resolveBrowserToolTarget(current managed base): %v", err)
	}
	if got.TargetID != tracked.ID || got.TabIndex != tracked.TabIndex || got.Value != "target:"+tracked.ID || !got.Explicit {
		t.Fatalf("expected managed current target to resolve for managed base, got %#v", got)
	}
}

func TestBrowserEffectiveBrowserAppIgnoresImplicitHostCurrentTargetForHiddenDefaultBase(t *testing.T) {
	registry := NewBrowserSessionRegistry()
	registry.TrackCurrentTarget("s1", BrowserSessionTarget{
		ID:         "host-current",
		TabIndex:   1,
		URL:        "https://host.example/current",
		Title:      "Host Current",
		BrowserApp: "Safari",
		Backend:    "system",
		Profile:    "default",
		Target:     "host",
	}, "tracked_active_tab")

	got := browserEffectiveBrowserApp(
		WithToolSessionID(context.Background(), "s1"),
		registry,
		BrowserRuntimeInfo{},
		true,
		"",
		browserToolTarget{Value: "current", Explicit: true},
	)
	if got != "" {
		t.Fatalf("expected hidden implicit-host default base to ignore current-target browser app fallback, got %q", got)
	}
}

func TestBrowserResolvedTargetURLIgnoresImplicitHostCurrentTargetForHiddenDefaultBase(t *testing.T) {
	registry := NewBrowserSessionRegistry()
	registry.TrackCurrentTarget("s1", BrowserSessionTarget{
		ID:         "host-current",
		TabIndex:   1,
		URL:        "https://host.example/current",
		Title:      "Host Current",
		BrowserApp: "Safari",
		Backend:    "system",
		Profile:    "default",
		Target:     "host",
	}, "tracked_active_tab")

	got := browserResolvedTargetURL(
		WithToolSessionID(context.Background(), "s1"),
		registry,
		BrowserRuntimeInfo{},
		true,
		"",
		browserToolTarget{Value: "current", Explicit: true},
	)
	if got != "" {
		t.Fatalf("expected hidden implicit-host default base to ignore current-target URL fallback, got %q", got)
	}
}

func TestBrowserResolvedTargetURLKeepsManagedCurrentTargetForManagedBase(t *testing.T) {
	registry := NewBrowserSessionRegistry()
	registry.TrackCurrentTarget("s1", BrowserSessionTarget{
		ID:         "node-current",
		TabIndex:   2,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "tracked_active_tab")

	got := browserResolvedTargetURL(
		WithToolSessionID(context.Background(), "s1"),
		registry,
		BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
		false,
		"Chromium",
		browserToolTarget{Value: "current", Explicit: true},
	)
	if got != "https://node.example/workbench" {
		t.Fatalf("expected managed current target URL to remain available for managed base, got %q", got)
	}
}

func TestBrowserResolveTrackedTargetForElementBindingIgnoresImplicitHostCurrentTargetFallback(t *testing.T) {
	registry := NewBrowserSessionRegistry()
	registry.TrackCurrentTarget("s1", BrowserSessionTarget{
		ID:         "host-current",
		TabIndex:   1,
		URL:        "https://host.example/current",
		Title:      "Host Current",
		BrowserApp: "Safari",
		Backend:    "system",
		Profile:    "default",
		Target:     "host",
	}, "tracked_active_tab")

	got, ok := browserResolveTrackedTargetForElementBinding(
		WithToolSessionID(context.Background(), "s1"),
		registry,
		BrowserRuntimeInfo{},
		true,
		"",
		browserToolTarget{},
	)
	if ok {
		t.Fatalf("expected hidden implicit-host default base to ignore element-binding current-target fallback, got %#v", got)
	}
}

func TestBrowserResolveTrackedTargetForElementBindingKeepsManagedCurrentTargetForManagedBase(t *testing.T) {
	registry := NewBrowserSessionRegistry()
	registry.TrackCurrentTarget("s1", BrowserSessionTarget{
		ID:         "node-current",
		TabIndex:   2,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "tracked_active_tab")

	got, ok := browserResolveTrackedTargetForElementBinding(
		WithToolSessionID(context.Background(), "s1"),
		registry,
		BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
		false,
		"Chromium",
		browserToolTarget{},
	)
	if !ok {
		t.Fatalf("expected managed current target to remain available for element-binding route lookup")
	}
	if got.URL != "https://node.example/workbench" || got.Target != "node" || got.Profile != "workbench" {
		t.Fatalf("unexpected managed tracked target for element-binding route lookup: %#v", got)
	}
}

func TestBrowserAutoFollowPendingTargetReviewStateIgnoresImplicitHostCurrentTargetFallback(t *testing.T) {
	registry := NewBrowserSessionRegistry()
	popup := registry.TrackTab("s1", BrowserSessionTarget{
		TabIndex:   2,
		URL:        "https://host.example/popup",
		Title:      "Popup",
		BrowserApp: "Safari",
		Backend:    "system",
		Profile:    "default",
		Target:     "host",
	}, false)
	registry.RecordPendingTargetReviewForRoute("s1", BrowserSessionRoute{
		Backend:    "system",
		Profile:    "default",
		Target:     "host",
		BrowserApp: "Safari",
	}, agentxbrowserruntime.BrowserSessionTargetReview{
		ID:         popup.ID,
		TabIndex:   popup.TabIndex,
		URL:        popup.URL,
		Title:      popup.Title,
		BrowserApp: popup.BrowserApp,
		Backend:    popup.Backend,
		Profile:    popup.Profile,
		Target:     popup.Target,
		Decision:   "session_target_popup_review_required",
		Reason:     "popup review",
	})

	got := browserAutoFollowPendingTargetReviewStateForRuntimeTarget(
		WithToolSessionID(context.Background(), "s1"),
		registry,
		BrowserRuntimeInfo{},
		true,
		"",
		browserToolTarget{Value: "current", Explicit: true},
	)
	if got.Review != nil || got.Count != 0 {
		t.Fatalf("expected hidden implicit-host default base to ignore current-target pending review fallback, got %#v", got)
	}
}

func TestBrowserAutoFollowPendingTargetReviewStateKeepsManagedRouteReview(t *testing.T) {
	registry := NewBrowserSessionRegistry()
	popup := registry.TrackTab("s1", BrowserSessionTarget{
		TabIndex:   3,
		URL:        "https://node.example/popup",
		Title:      "Popup",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, false)
	registry.RecordPendingTargetReviewForRoute("s1", BrowserSessionRoute{
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
		BrowserApp: "Chromium",
	}, agentxbrowserruntime.BrowserSessionTargetReview{
		ID:         popup.ID,
		TabIndex:   popup.TabIndex,
		URL:        popup.URL,
		Title:      popup.Title,
		BrowserApp: popup.BrowserApp,
		Backend:    popup.Backend,
		Profile:    popup.Profile,
		Target:     popup.Target,
		Decision:   "session_target_popup_review_required",
		Reason:     "popup review",
	})

	got := browserAutoFollowPendingTargetReviewStateForRuntimeTarget(
		WithToolSessionID(context.Background(), "s1"),
		registry,
		BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
		false,
		"Chromium",
		browserToolTarget{Value: "current", Explicit: true},
	)
	if got.Review == nil || got.Review.ID != popup.ID || got.Count != 1 {
		t.Fatalf("expected managed route pending review to remain available, got %#v", got)
	}
}
